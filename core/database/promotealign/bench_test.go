package promotealign_test

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"
	"sort"
	"testing"

	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// catalogShape is the catalog bulkCanon writes. City event i is the birth of
// person i when i < Persons, else the death of person i % Persons; city events
// go round-robin over the cities. Town event j is a death of a random person at a
// random town. With Dates, person i is born in 1780 + Rand.IntN(120) and dies
// 20 + Rand.IntN(60) years later, and each event's date comes from fixtureDate.
type catalogShape struct {
	Persons, Cities, CityEvents, Towns, TownEvents int
	Name                                           func(i int) string
	Rand                                           *rand.Rand
	Dates                                          bool
	// TreeGenerations, when above 0, adds a parent tree of 2^TreeGenerations - 1
	// persons whose root is named TreeRoot: person n has parents 2n and 2n+1,
	// each as a filed parent relationship, and each couple has a marriage with
	// two spouse participations. Generations 12 is the tree BenchmarkDerive
	// reads: 8,190 parent links.
	TreeGenerations int
	TreeRoot        string
}

// benchShape is the catalog the reported benchmarks run on: 20k persons and 30k
// events over 5 cities, undated (events 20,000 and up are second births of
// persons 0-9,999, written as deaths).
func benchShape() catalogShape {
	return catalogShape{
		Persons: 20000, Cities: 5, CityEvents: 30000, Name: benchName,
		Rand: rand.New(rand.NewPCG(1, 2)),
	}
}

// largeShape is the catalog TestLargeSourceBudgets runs on: 4,000 persons, one
// city of 6,000 events, 50 towns of 2,000 events, Zipf names, dated, and a
// parent tree of 12 generations whose root is named treeRoot, so a proposal
// that reaches it runs the kinship patterns over about 8,000 parent links.
func largeShape(r *rand.Rand, names *zipfNames, treeRoot string) catalogShape {
	return catalogShape{
		Persons: 4000, Cities: 1, CityEvents: 6000, Towns: 50, TownEvents: 2000,
		Name: func(int) string { return names.next() }, Rand: r, Dates: true,
		TreeGenerations: 12, TreeRoot: treeRoot,
	}
}

// benchRef is a canonical entity ref the fixture writes. The "U" cannot appear
// in a ref core/ref mints (its alphabet leaves out I, L, O and U), so a fixture
// ref never equals a ref the test mints at random.
func benchRef(prefix string, n int) string {
	return prefix + "-U" + fmt.Sprintf("%05d", n)
}

// letters is the lowercase base 26 of k: a for 0, ba for 26.
func letters(k int) string {
	if k == 0 {
		return "a"
	}
	var b []byte
	for ; k > 0; k /= 26 {
		b = append(b, byte('a'+k%26))
	}
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

// fixtureDate is a date in or near year: one in twenty is BEF, one AFT, one
// absent (nil), the rest a day in the year.
func fixtureDate(r *rand.Rand, year int) *datevalue.Value {
	switch r.IntN(20) {
	case 0:
		return datevaluetest.Before(year, 0, 0)
	case 1:
		return datevaluetest.After(year, 0, 0)
	case 2:
		return nil
	}
	return datevaluetest.YMD(year, 1+r.IntN(12), 1+r.IntN(28))
}

// zipfNames draws "Giv<letters> Sur<letters>" names: a given name from 200
// ranks (exponent 0.9) and a surname from 3,000 (exponent 0.7), so the top
// given name keys about 13% of the catalog and the top surname about 3%.
type zipfNames struct {
	r       *rand.Rand
	given   []float64
	surname []float64
}

func newZipfNames(r *rand.Rand) *zipfNames {
	return &zipfNames{r: r, given: cumulativeZipf(200, 0.9), surname: cumulativeZipf(3000, 0.7)}
}

// cumulativeZipf is the cumulative normalized weights 1/k^s for k = 1..n.
func cumulativeZipf(n int, s float64) []float64 {
	out := make([]float64, n)
	sum := 0.0
	for k := 1; k <= n; k++ {
		sum += 1 / math.Pow(float64(k), s)
		out[k-1] = sum
	}
	for i := range out {
		out[i] /= sum
	}
	return out
}

func (z *zipfNames) draw(cum []float64) int {
	return min(sort.SearchFloat64s(cum, z.r.Float64()), len(cum)-1)
}

func (z *zipfNames) next() string {
	g := z.draw(z.given)
	sn := z.draw(z.surname)
	return "Giv" + letters(g) + " " + "Sur" + letters(sn)
}

// benchName is the name of person i: a given name from a pool of 200 and a
// surname from a pool of 2000, so a probe shares a word with a few dozen.
func benchName(i int) string {
	return fmt.Sprintf("Given%d Surname%d", i%200, i%2000)
}

// BenchmarkProposeColdCity times a cold proposal against a catalog of about
// 20k persons, 30k events, and a few cities, and reports the heap afterward.
func BenchmarkProposeColdCity(b *testing.B) {
	f := newFixture(b)
	city := f.place(f.source, f.artifact, "City")
	cityID := f.promote(city).Entity.ID
	f.person(f.source, f.artifact, benchName(7))
	db, err := f.c.DB()
	must(b, err)
	must(b, bulkCanon(db, cityID, benchShape()))

	b.ResetTimer()
	var heap uint64
	for i := 0; i < b.N; i++ {
		f.c.Graph().Drop()
		runtime.GC()
		_, _, err := promotealign.Propose(f.c, f.graph(f.source), nil)
		must(b, err)
		runtime.GC()
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		heap = ms.HeapAlloc
	}
	b.ReportMetric(float64(heap), "heap-B")
}

// BenchmarkProposeWarmCity times a repeat proposal on BenchmarkProposeColdCity's
// catalog, with no drop between, and reports the heap afterward. The Source
// holds one unfixed named person, so each proposal looks up the shortlist index
// and the first one builds it. Reported, not gated.
func BenchmarkProposeWarmCity(b *testing.B) {
	f := newFixture(b)
	city := f.place(f.source, f.artifact, "City")
	cityID := f.promote(city).Entity.ID
	f.person(f.source, f.artifact, benchName(7))
	db, err := f.c.DB()
	must(b, err)
	must(b, bulkCanon(db, cityID, benchShape()))
	_, _, err = promotealign.Propose(f.c, f.graph(f.source), nil)
	must(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := promotealign.Propose(f.c, f.graph(f.source), nil)
		must(b, err)
	}
	b.StopTimer()
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	b.ReportMetric(float64(ms.HeapAlloc), "heap-B")
}

// BenchmarkProposeAfterCommit times a proposal that follows a handle commit,
// so each one rebuilds the shortlist index. Reported, not gated.
func BenchmarkProposeAfterCommit(b *testing.B) {
	f := newFixture(b)
	city := f.place(f.source, f.artifact, "City")
	cityID := f.promote(city).Entity.ID
	f.person(f.source, f.artifact, benchName(7))
	db, err := f.c.DB()
	must(b, err)
	must(b, bulkCanon(db, cityID, benchShape()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		f.promote(f.person(f.source, f.artifact, fmt.Sprintf("Added%d Person", i)))
		b.StartTimer()
		_, _, err := promotealign.Propose(f.c, f.graph(f.source), nil)
		must(b, err)
	}
}

func benchID(n int) []byte {
	var id [16]byte
	id[0] = 0xB0
	binary.BigEndian.PutUint64(id[8:], uint64(n))
	return id[:]
}

func bulkCanon(db *sql.DB, city []byte, shape catalogShape) error {
	idOf := func(table, key string) ([]byte, error) {
		var id []byte
		err := db.QueryRow(`SELECT id FROM `+table+` WHERE key = ? AND origin = 'provenencia'`, key).Scan(&id)
		return append([]byte(nil), id...), err
	}
	personT, err := idOf("subject_types", "person")
	if err != nil {
		return err
	}
	eventT, err := idOf("subject_types", "event")
	if err != nil {
		return err
	}
	placeT, err := idOf("subject_types", "place")
	if err != nil {
		return err
	}
	partT, err := idOf("subject_types", "participation")
	if err != nil {
		return err
	}
	locT, err := idOf("subject_types", "location")
	if err != nil {
		return err
	}
	personP, err := idOf("properties", "person")
	if err != nil {
		return err
	}
	eventP, err := idOf("properties", "event")
	if err != nil {
		return err
	}
	placeP, err := idOf("properties", "place")
	if err != nil {
		return err
	}
	roleP, err := idOf("properties", "role")
	if err != nil {
		return err
	}
	typeP, err := idOf("properties", "event_type")
	if err != nil {
		return err
	}
	nameP, err := idOf("properties", "name")
	if err != nil {
		return err
	}
	toponymP, err := idOf("properties", "toponym")
	if err != nil {
		return err
	}
	dateP, err := idOf("properties", "date")
	if err != nil {
		return err
	}
	var roleTerm, birthTerm, deathTerm []byte
	if err := db.QueryRow(`SELECT t.id FROM property_terms t JOIN properties p ON p.id = t.property_id
		WHERE p.key = 'role' AND p.origin = 'provenencia' AND t.key = 'subject'`).Scan(&roleTerm); err != nil {
		return err
	}
	if err := db.QueryRow(`SELECT t.id FROM property_terms t JOIN properties p ON p.id = t.property_id
		WHERE p.key = 'event_type' AND p.origin = 'provenencia' AND t.key = 'birth'`).Scan(&birthTerm); err != nil {
		return err
	}
	if err := db.QueryRow(`SELECT t.id FROM property_terms t JOIN properties p ON p.id = t.property_id
		WHERE p.key = 'event_type' AND p.origin = 'provenencia' AND t.key = 'death'`).Scan(&deathTerm); err != nil {
		return err
	}
	deathTerm = append([]byte(nil), deathTerm...)
	roleTerm = append([]byte(nil), roleTerm...)
	birthTerm = append([]byte(nil), birthTerm...)

	var tree treeIDs
	if shape.TreeGenerations > 0 {
		if tree, err = lookupTree(db); err != nil {
			return err
		}
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	ent, err := tx.Prepare(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer ent.Close()
	val, err := tx.Prepare(`INSERT INTO auto_reconciler_values
		(entity_id, property_id, rank, value_term_id, value_entity_id, support, against, reason)
		VALUES (?, ?, 1, ?, ?, 1, 0, 'kept')`)
	if err != nil {
		return err
	}
	defer val.Close()

	nameVal, err := tx.Prepare(`INSERT INTO auto_reconciler_values
		(entity_id, property_id, rank, value_name, support, against, reason)
		VALUES (?, ?, 1, ?, 1, 0, 'kept')`)
	if err != nil {
		return err
	}
	defer nameVal.Close()

	dateVal, err := tx.Prepare(`INSERT INTO auto_reconciler_values
		(entity_id, property_id, rank, value_date, support, against, reason)
		VALUES (?, ?, 1, ?, 1, 0, 'kept')`)
	if err != nil {
		return err
	}
	defer dateVal.Close()

	textVal, err := tx.Prepare(`INSERT INTO auto_reconciler_values
		(entity_id, property_id, rank, value_text, support, against, reason)
		VALUES (?, ?, 1, ?, 1, 0, 'kept')`)
	if err != nil {
		return err
	}
	defer textVal.Close()

	r := shape.Rand
	cities := make([][]byte, shape.Cities)
	cities[0] = city
	for i := 1; i < shape.Cities; i++ {
		id := benchID(i)
		if _, err := ent.Exec(id, placeT, benchRef("PLC", i)); err != nil {
			return err
		}
		cities[i] = id
	}
	towns := make([][]byte, shape.Towns)
	for k := range towns {
		id := benchID(900000 + k)
		if _, err := ent.Exec(id, placeT, benchRef("PLC", 90000+k)); err != nil {
			return err
		}
		if _, err := textVal.Exec(id, toponymP, "Town"+letters(k)); err != nil {
			return err
		}
		towns[k] = id
	}
	persons := make([][]byte, shape.Persons)
	birthYear := make([]int, shape.Persons)
	for i := 0; i < shape.Persons; i++ {
		id := benchID(100 + i)
		if _, err := ent.Exec(id, personT, benchRef("PER", i)); err != nil {
			return err
		}
		persons[i] = id
		name, err := valuecodec.MarshalName(*namevaluetest.Western(shape.Name(i)))
		if err != nil {
			return err
		}
		if _, err := nameVal.Exec(id, nameP, name); err != nil {
			return err
		}
		if shape.Dates {
			birthYear[i] = 1780 + r.IntN(120)
		}
	}
	deathYear := make([]int, shape.Persons)
	if shape.Dates {
		for i := range deathYear {
			deathYear[i] = birthYear[i] + 20 + r.IntN(60)
		}
	}
	// event writes one event with its subject participation, its place and,
	// when it has one, its date.
	event := func(eventID, partID, locID, n, refN int, person, place, typeTerm []byte, d *datevalue.Value) error {
		ev := benchID(eventID + n)
		if _, err := ent.Exec(ev, eventT, benchRef("EVT", refN)); err != nil {
			return err
		}
		if _, err := val.Exec(ev, typeP, typeTerm, nil); err != nil {
			return err
		}
		if d != nil {
			b, err := valuecodec.MarshalDate(*d)
			if err != nil {
				return err
			}
			if _, err := dateVal.Exec(ev, dateP, b); err != nil {
				return err
			}
		}
		part := benchID(partID + n)
		if _, err := ent.Exec(part, partT, benchRef("PAR", refN)); err != nil {
			return err
		}
		if _, err := val.Exec(part, personP, nil, person); err != nil {
			return err
		}
		if _, err := val.Exec(part, eventP, nil, ev); err != nil {
			return err
		}
		if _, err := val.Exec(part, roleP, roleTerm, nil); err != nil {
			return err
		}
		loc := benchID(locID + n)
		if _, err := ent.Exec(loc, locT, benchRef("LOC", refN)); err != nil {
			return err
		}
		if _, err := val.Exec(loc, eventP, nil, ev); err != nil {
			return err
		}
		_, err := val.Exec(loc, placeP, nil, place)
		return err
	}
	dateFor := func(year int) *datevalue.Value {
		if !shape.Dates {
			return nil
		}
		return fixtureDate(r, year)
	}
	for i := 0; i < shape.CityEvents; i++ {
		p, term, year := i, birthTerm, 0
		if i < shape.Persons {
			year = birthYear[p]
		} else {
			p, term = i%shape.Persons, deathTerm
			year = deathYear[p]
		}
		if err := event(100000, 200000, 400000, i, i, persons[p], cities[i%shape.Cities], term, dateFor(year)); err != nil {
			return err
		}
	}
	for j := 0; j < shape.TownEvents; j++ {
		p := r.IntN(shape.Persons)
		k := r.IntN(shape.Towns)
		if err := event(600000, 700000, 800000, j, 500000+j, persons[p], towns[k], deathTerm, dateFor(deathYear[p])); err != nil {
			return err
		}
	}
	if shape.TreeGenerations > 0 {
		if err := bulkTree(tree, ent, val, nameVal, shape); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// treeIDs are the vocabulary ids bulkTree writes with.
type treeIDs struct {
	person, event, participation, relationship []byte
	personP, eventP, roleP, relatedP, relTypeP []byte
	typeP, nameP                               []byte
	subject, marriage, parent                  []byte
}

// lookupTree reads treeIDs. It runs before bulkCanon's transaction, because
// the catalog has one connection.
func lookupTree(db *sql.DB) (treeIDs, error) {
	var ids treeIDs
	idOf := func(table, key string) []byte {
		var id []byte
		err := db.QueryRow(`SELECT id FROM `+table+` WHERE key = ? AND origin = 'provenencia'`, key).Scan(&id)
		if err != nil {
			return nil
		}
		return append([]byte(nil), id...)
	}
	termOf := func(property, key string) []byte {
		var id []byte
		err := db.QueryRow(`SELECT t.id FROM property_terms t JOIN properties p ON p.id = t.property_id
			WHERE p.key = ? AND p.origin = 'provenencia' AND t.key = ?`, property, key).Scan(&id)
		if err != nil {
			return nil
		}
		return append([]byte(nil), id...)
	}
	ids = treeIDs{
		person: idOf("subject_types", "person"), event: idOf("subject_types", "event"),
		participation: idOf("subject_types", "participation"), relationship: idOf("subject_types", "relationship"),
		personP: idOf("properties", "person"), eventP: idOf("properties", "event"), roleP: idOf("properties", "role"),
		relatedP: idOf("properties", "related_to"), relTypeP: idOf("properties", "relationship_type"),
		typeP: idOf("properties", "event_type"), nameP: idOf("properties", "name"),
		subject: termOf("role", "subject"), marriage: termOf("event_type", "marriage"), parent: termOf("relationship_type", "parent"),
	}
	for _, id := range [][]byte{ids.person, ids.event, ids.participation, ids.relationship, ids.personP, ids.eventP,
		ids.roleP, ids.relatedP, ids.relTypeP, ids.typeP, ids.nameP, ids.subject, ids.marriage, ids.parent} {
		if id == nil {
			return ids, fmt.Errorf("the catalog lacks a vocabulary row the tree needs")
		}
	}
	return ids, nil
}

// bulkTree writes shape's parent tree (see catalogShape.TreeGenerations) in a
// transaction. A person with parents also gets a marriage of the two, so a walk
// through the tree has the spouse and chain patterns to read as well as the
// filed links.
func bulkTree(ids treeIDs, ent, val, nameVal *sql.Stmt, shape catalogShape) error {
	n := 1<<shape.TreeGenerations - 1 // persons 1..n
	person := func(i int) []byte { return benchID(1000000 + i) }
	for i := 1; i <= n; i++ {
		if _, err := ent.Exec(person(i), ids.person, benchRef("PER", 100000+i)); err != nil {
			return err
		}
		name := fmt.Sprintf("Tree%s Line%s", letters(i), letters(i%97))
		if i == 1 {
			name = shape.TreeRoot
		}
		v, err := valuecodec.MarshalName(*namevaluetest.Western(name))
		if err != nil {
			return err
		}
		if _, err := nameVal.Exec(person(i), ids.nameP, v); err != nil {
			return err
		}
	}
	link := 0
	for i := 1; 2*i+1 <= n; i++ {
		ev := benchID(2000000 + i)
		if _, err := ent.Exec(ev, ids.event, benchRef("EVT", 100000+i)); err != nil {
			return err
		}
		if _, err := val.Exec(ev, ids.typeP, ids.marriage, nil); err != nil {
			return err
		}
		for k, mate := range []int{2 * i, 2*i + 1} {
			part := benchID(3000000 + 2*i + k)
			if _, err := ent.Exec(part, ids.participation, benchRef("PAR", 100000+2*i+k)); err != nil {
				return err
			}
			if _, err := val.Exec(part, ids.personP, nil, person(mate)); err != nil {
				return err
			}
			if _, err := val.Exec(part, ids.eventP, nil, ev); err != nil {
				return err
			}
			if _, err := val.Exec(part, ids.roleP, ids.subject, nil); err != nil {
				return err
			}
			link++
			rel := benchID(4000000 + link)
			if _, err := ent.Exec(rel, ids.relationship, benchRef("REL", 100000+link)); err != nil {
				return err
			}
			if _, err := val.Exec(rel, ids.personP, nil, person(mate)); err != nil {
				return err
			}
			if _, err := val.Exec(rel, ids.relatedP, nil, person(i)); err != nil {
				return err
			}
			if _, err := val.Exec(rel, ids.relTypeP, ids.parent, nil); err != nil {
				return err
			}
		}
	}
	return nil
}
