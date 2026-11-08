package searchindex_test

import (
	"crypto/rand"
	"database/sql"
	"slices"
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
}

type inputRow struct{ input, bridge, endpoint string }

func inputRowsOf(t *testing.T, db *sql.DB, dependent []byte) []inputRow {
	t.Helper()
	rows, err := db.Query(`SELECT input_id, bridge, endpoint FROM catalog_search_inputs WHERE dependent_id = ?`, dependent)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []inputRow
	for rows.Next() {
		var id []byte
		var r inputRow
		if err := rows.Scan(&id, &r.bridge, &r.endpoint); err != nil {
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
	if rows := inputRowsOf(t, db, birthID); !hasRow(rows, jamesID, "", "") {
		t.Fatalf("the birth's inputs %+v miss James", rows)
	}
	ends := inputRowsOf(t, db, partID)
	if len(ends) != 2 || !hasRow(ends, jamesID, "participation", "person") || !hasRow(ends, birthID, "participation", "event") {
		t.Fatalf("the participation's inputs %+v, want its two ends", ends)
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
