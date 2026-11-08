package searchindex_test

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"slices"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/writes"
)

func TestReprojectHandlesBuildsOneGraph(t *testing.T) {
	r := evidencelayertest.New(t)
	db, err := r.C.DB()
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	restore := searchindex.SetNewGraphForTest(func(q graphcache.Querier) *graphcache.Graph {
		calls++
		return graphcache.Over(q)
	})
	defer restore()

	ids := make([][]byte, 1200)
	for i := range ids {
		ids[i] = make([]byte, 16)
		if _, err := rand.Read(ids[i]); err != nil {
			t.Fatal(err)
		}
	}
	if err := searchindex.ReprojectHandles(db, ids); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("1,200 ids built %d graphs, want 1", calls)
	}
	calls = 0
	if err := searchindex.ReprojectHandles(db, nil); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("no ids built %d graphs, want 0", calls)
	}
	if _, err := searchindex.ReprojectChanged(db, ids, ids[:10]); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("one ReprojectChanged call built %d graphs, want 1", calls)
	}
}

type inputRow struct {
	input, bridge, endpoint string
	association             bool
}

func inputRowsOf(t *testing.T, db *sql.DB, dependent []byte) []inputRow {
	t.Helper()
	rows, err := db.Query(`SELECT input_id, bridge, endpoint, dependent_is_association FROM catalog_search_inputs WHERE dependent_id = ?`, dependent)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []inputRow
	for rows.Next() {
		var id []byte
		var r inputRow
		if err := rows.Scan(&id, &r.bridge, &r.endpoint, &r.association); err != nil {
			t.Fatal(err)
		}
		r.input = string(id)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func hasRow(rows []inputRow, id []byte, bridge, endpoint string) bool {
	want := graphcache.InputKey{ID: id, Bridge: bridge, Endpoint: endpoint}
	return slices.ContainsFunc(rows, func(r inputRow) bool {
		return graphcache.InputKey{ID: []byte(r.input), Bridge: r.bridge, Endpoint: r.endpoint}.Equal(want)
	})
}

func TestReprojectWritesInputs(t *testing.T) {
	r := evidencelayertest.New(t)
	db, err := r.C.DB()
	if err != nil {
		t.Fatal(err)
	}
	user := evidencelayertest.UserID
	person := func(name string) subjects.Subject {
		p := r.Bare(r.Source, "person")
		r.Cite(r.Artifact, p, observations.Input{PropertyID: r.Prop("name").ID, Name: namevaluetest.Western(name)})
		return p
	}
	james, mary := person("James Robins"), person("Mary Smith")
	birth := r.Bare(r.Source, "event")
	r.Cite(r.Artifact, birth, observations.Input{PropertyID: r.Prop("event_type").ID, ValueTermID: r.Term("event_type", "birth").ID})
	r.Participation(r.Source, r.Artifact, james, birth, "subject")
	save := func(s subjects.Subject) []byte {
		return promotetest.Save(t, r.C, user, promote.Input{SubjectID: s.ID}).Entity.ID
	}
	jamesID, maryID, birthID := save(james), save(mary), save(birth)
	var partID []byte
	if err := db.QueryRow(`SELECT e.id FROM canonical_entities e JOIN subject_types st ON st.id = e.subject_type_id
		WHERE st.key = 'participation'`).Scan(&partID); err != nil {
		t.Fatal(err)
	}

	// The upkeep that already ran is cleared, so the rows below come from this call.
	if err := searchindex.ClearAll(db); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.ReprojectHandles(db, [][]byte{birthID, partID}); err != nil {
		t.Fatal(err)
	}
	birthRows := inputRowsOf(t, db, birthID)
	if !hasRow(birthRows, jamesID, "", "") {
		t.Fatalf("the birth's inputs %+v miss James", birthRows)
	}
	for _, r := range birthRows {
		if r.association {
			t.Fatalf("the birth's row %+v is flagged as an association's end", r)
		}
	}
	ends := inputRowsOf(t, db, partID)
	if len(ends) != 2 || !hasRow(ends, jamesID, "participation", "person") || !hasRow(ends, birthID, "participation", "event") {
		t.Fatalf("the participation's inputs %+v, want its two ends", ends)
	}
	for _, r := range ends {
		if !r.association {
			t.Fatalf("the participation's row %+v is not flagged as an association's end", r)
		}
	}

	// A handle that is merged keeps no rows as a dependent.
	if err := searchindex.ReprojectHandles(db, [][]byte{jamesID, maryID}); err != nil {
		t.Fatal(err)
	}
	if len(inputRowsOf(t, db, jamesID)) == 0 {
		t.Fatal("James has no inputs before he is merged")
	}
	if _, err := db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, maryID, jamesID); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.ReprojectHandles(db, [][]byte{jamesID}); err != nil {
		t.Fatal(err)
	}
	if rows := inputRowsOf(t, db, jamesID); len(rows) != 0 {
		t.Fatalf("a merged handle kept inputs %+v", rows)
	}

	// A handle whose document is dropped keeps no rows as a dependent.
	if _, err := writes.Call(r.C, writes.Op{Action: "unpromote_subject", UserID: user}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := promote.Unpromote(tx, user, birth.ID)
		return struct{}{}, ch, err
	}); err != nil {
		t.Fatal(err)
	}
	if rows := inputRowsOf(t, db, birthID); len(rows) != 0 {
		t.Fatalf("an unpromoted birth kept inputs %+v", rows)
	}

	if err := searchindex.ClearAll(db); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM catalog_search_inputs`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("ClearAll left %d input rows, err %v", n, err)
	}
}

// filedWorld is James Robins and a birth, both promoted, and a participation
// filed between them (role subject), with the index current.
type filedWorld struct {
	r                        *evidencelayertest.Records
	db                       *sql.DB
	james, birth             subjects.Subject
	jamesID, birthID, partID []byte
}

func newFiledWorld(t *testing.T) filedWorld {
	t.Helper()
	r := evidencelayertest.New(t)
	db, err := r.C.DB()
	if err != nil {
		t.Fatal(err)
	}
	james := r.Bare(r.Source, "person")
	r.Cite(r.Artifact, james, observations.Input{PropertyID: r.Prop("name").ID, Name: namevaluetest.Western("James Robins")})
	birth := r.Bare(r.Source, "event")
	r.Cite(r.Artifact, birth, observations.Input{PropertyID: r.Prop("event_type").ID, ValueTermID: r.Term("event_type", "birth").ID})
	r.Participation(r.Source, r.Artifact, james, birth, "subject")
	save := func(s subjects.Subject) []byte {
		return promotetest.Save(t, r.C, evidencelayertest.UserID, promote.Input{SubjectID: s.ID}).Entity.ID
	}
	w := filedWorld{r: r, db: db, james: james, birth: birth, jamesID: save(james), birthID: save(birth)}
	if err := db.QueryRow(`SELECT e.id FROM canonical_entities e JOIN subject_types st ON st.id = e.subject_type_id
		WHERE st.key = 'participation'`).Scan(&w.partID); err != nil {
		t.Fatal(err)
	}
	return w
}

func holds(ids [][]byte, id []byte) bool {
	return slices.ContainsFunc(ids, func(x []byte) bool { return bytes.Equal(x, id) })
}

func TestIndexReads(t *testing.T) {
	w := newFiledWorld(t)
	byValues, err := searchindex.DependentsOfForTest(w.db, []graphcache.InputKey{{ID: w.jamesID}})
	if err != nil || !holds(byValues, w.birthID) {
		t.Fatalf("rows composed from James's values %x, err %v: want the birth", byValues, err)
	}
	byLink, err := searchindex.DependentsOfForTest(w.db, []graphcache.InputKey{{ID: w.birthID, Bridge: "participation", Endpoint: "event"}})
	if err != nil || !holds(byLink, w.birthID) {
		t.Fatalf("rows composed from the birth's participation link %x, err %v: want the birth", byLink, err)
	}
	byHandle, err := searchindex.DependentsOfHandlesForTest(w.db, [][]byte{w.partID})
	if err != nil || !holds(byHandle, w.birthID) || !holds(byHandle, w.jamesID) {
		t.Fatalf("rows composed from the participation %x, err %v: want the birth and James", byHandle, err)
	}
	ends, err := searchindex.OldEndsForTest(w.db, [][]byte{w.partID, w.birthID, make([]byte, 16)})
	if err != nil {
		t.Fatal(err)
	}
	if got := ends[string(w.partID)]; len(got) != 2 || !hasKey(got, w.jamesID, "participation", "person") || !hasKey(got, w.birthID, "participation", "event") {
		t.Fatalf("the participation's old ends %+v, want James and the birth", got)
	}
	if len(ends) != 1 {
		t.Fatalf("old ends of a summary row and an unknown id: %+v, want none", ends)
	}
}

func hasKey(keys []graphcache.InputKey, id []byte, bridge, endpoint string) bool {
	want := graphcache.InputKey{ID: id, Bridge: bridge, Endpoint: endpoint}
	return slices.ContainsFunc(keys, func(k graphcache.InputKey) bool { return k.Equal(want) })
}

func TestUnpromoteReportsTheDeletedLinksEnds(t *testing.T) {
	w := newFiledWorld(t)
	user := evidencelayertest.UserID
	_, res, err := writes.Run(w.r.C, writes.Op{Action: "unpromote_subject", UserID: user}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := promote.Unpromote(tx, user, w.james.ID)
		return struct{}{}, ch, err
	})
	if err != nil {
		t.Fatal(err)
	}
	if deps := res.Effects.SummaryDeps; !holds(deps, w.birthID) || !holds(deps, w.jamesID) {
		t.Fatalf("SummaryDeps %x, want the birth and James", deps)
	}
	for _, id := range [][]byte{w.jamesID, w.partID} {
		if rows := inputRowsOf(t, w.db, id); len(rows) != 0 {
			t.Fatalf("a gone handle kept inputs %+v", rows)
		}
	}
	var title string
	if err := w.db.QueryRow(`SELECT title FROM catalog_search_docs WHERE entity_id = ?`, database.UUIDString(w.birthID)).Scan(&title); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(title, "James") {
		t.Fatalf("the birth's search title %q still names James", title)
	}
}
