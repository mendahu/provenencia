package promotealign_test

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"runtime"
	"testing"

	"github.com/mendahu/provenencia/core/database/promotealign"
)

const (
	benchPersons = 20000
	benchEvents  = 30000
	benchHubs    = 5
)

// BenchmarkProposeColdHub times a cold proposal against a catalog of about
// 20k persons, 30k events, and a few hub places, and reports the heap afterward.
func BenchmarkProposeColdHub(b *testing.B) {
	f := newFixture(b)
	city := f.place(f.source, f.artifact, "Hub")
	hub := f.promote(city).Entity.ID
	db, err := f.c.DB()
	must(b, err)
	must(b, bulkCanon(db, hub))

	b.ResetTimer()
	var heap uint64
	for i := 0; i < b.N; i++ {
		f.c.Graph().Drop()
		runtime.GC()
		_, _, err := promotealign.Propose(f.c, f.source.ID, nil)
		must(b, err)
		runtime.GC()
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		heap = ms.HeapAlloc
	}
	b.ReportMetric(float64(heap), "heap-B")
}

func benchID(n int) []byte {
	var id [16]byte
	id[0] = 0xB0
	binary.BigEndian.PutUint64(id[8:], uint64(n))
	return id[:]
}

func bulkCanon(db *sql.DB, hub []byte) error {
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
	var roleTerm, birthTerm []byte
	if err := db.QueryRow(`SELECT t.id FROM property_terms t JOIN properties p ON p.id = t.property_id
		WHERE p.key = 'role' AND p.origin = 'provenencia' AND t.key = 'subject'`).Scan(&roleTerm); err != nil {
		return err
	}
	if err := db.QueryRow(`SELECT t.id FROM property_terms t JOIN properties p ON p.id = t.property_id
		WHERE p.key = 'event_type' AND p.origin = 'provenencia' AND t.key = 'birth'`).Scan(&birthTerm); err != nil {
		return err
	}
	roleTerm = append([]byte(nil), roleTerm...)
	birthTerm = append([]byte(nil), birthTerm...)

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

	hubs := make([][]byte, benchHubs)
	hubs[0] = hub
	for i := 1; i < benchHubs; i++ {
		id := benchID(i)
		if _, err := ent.Exec(id, placeT, fmt.Sprintf("PLC-%05d", i)); err != nil {
			return err
		}
		hubs[i] = id
	}
	persons := make([][]byte, benchPersons)
	for i := 0; i < benchPersons; i++ {
		id := benchID(100 + i)
		if _, err := ent.Exec(id, personT, fmt.Sprintf("PER-%05d", i)); err != nil {
			return err
		}
		persons[i] = id
	}
	for i := 0; i < benchEvents; i++ {
		event := benchID(100000 + i)
		if _, err := ent.Exec(event, eventT, fmt.Sprintf("EVT-%05d", i)); err != nil {
			return err
		}
		if _, err := val.Exec(event, typeP, birthTerm, nil); err != nil {
			return err
		}
		part := benchID(200000 + i)
		if _, err := ent.Exec(part, partT, fmt.Sprintf("PAR-%05d", i)); err != nil {
			return err
		}
		person := persons[i%benchPersons]
		if _, err := val.Exec(part, personP, nil, person); err != nil {
			return err
		}
		if _, err := val.Exec(part, eventP, nil, event); err != nil {
			return err
		}
		if _, err := val.Exec(part, roleP, roleTerm, nil); err != nil {
			return err
		}
		loc := benchID(400000 + i)
		if _, err := ent.Exec(loc, locT, fmt.Sprintf("LOC-%05d", i)); err != nil {
			return err
		}
		if _, err := val.Exec(loc, eventP, nil, event); err != nil {
			return err
		}
		if _, err := val.Exec(loc, placeP, nil, hubs[i%benchHubs]); err != nil {
			return err
		}
	}
	return tx.Commit()
}
