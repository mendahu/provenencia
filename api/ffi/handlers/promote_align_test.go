package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
	"google.golang.org/protobuf/proto"
)

// personOnNewSource creates a person named "Mary Smith" with a sex-at-birth
// Observation of the given polarity on a fresh Source, and returns the
// Source and Subject ids.
func personOnNewSource(t *testing.T, dir, userID, typeID, sexKey, polarity string) (sourceID, subjectID string) {
	t.Helper()
	return namedPersonOnNewSource(t, dir, userID, typeID, "Mary Smith", sexKey, polarity)
}

// namedPersonOnNewSource is personOnNewSource for a person with the given name.
func namedPersonOnNewSource(t *testing.T, dir, userID, typeID, personName, sexKey, polarity string) (sourceID, subjectID string) {
	t.Helper()
	uid := uuid.MustParse(userID)
	tid := uuid.MustParse(typeID)
	if err := withProjectCatalog(dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		var sourceTypeID []byte
		if err := db.QueryRow(`SELECT source_type_id FROM sources LIMIT 1`).Scan(&sourceTypeID); err != nil {
			return err
		}
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: uid[:]}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, uid[:], sources.CreateInput{SourceTypeID: sourceTypeID, Title: "Register " + polarity})
		})
		if err != nil {
			return err
		}
		graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
		s := subjectstest.OfType(t, c, uid[:], graph, tid[:], "")
		name, err := properties.Lookup(c, "name", vocabkey.OriginProvenencia)
		if err != nil {
			return err
		}
		sex, err := properties.Lookup(c, "sex_at_birth", vocabkey.OriginProvenencia)
		if err != nil {
			return err
		}
		term, err := propertyterms.Lookup(c, sex.ID, sexKey, vocabkey.OriginProvenencia)
		if err != nil {
			return err
		}
		art, err := artifactstest.Create(c, uid[:], artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
		if err != nil {
			return err
		}
		_, _, err = observationstest.Cite(c, uid[:], s.GraphID, art.ID, `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`, []observations.Input{
			{SubjectID: s.ID, PropertyID: name.ID, Name: namevaluetest.Western(personName)},
			{SubjectID: s.ID, PropertyID: sex.ID, ValueTermID: term.ID, Polarity: polarity},
		}...)
		sourceID, subjectID = database.UUIDString(src.ID), database.UUIDString(s.ID)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return sourceID, subjectID
}

// The proposal sends where each row starts: unset when nothing is suggested or
// the pair has no band, else the record the row sits on.
func TestProposeSendsWhereARowStarts(t *testing.T) {
	cases := []struct {
		name       string
		polarity   string
		fix        bool
		wantTarget string
	}{
		{"negative, unfixed", observations.PolarityNegative, false, ""},
		// The conflict leaves the fixed pair under the weak band: no band, no start.
		{"negative, fixed", observations.PolarityNegative, true, ""},
		{"positive", observations.PolarityPositive, false, "handle"},
	}
	var tests []rpcTest
	for _, tc := range cases {
		var incoming string
		tests = append(tests, rpcTest{
			name: tc.name,
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, typeID := subjectFixture(t)
				_, handleSubject := personOnNewSource(t, dir, userID, typeID, "male", observations.PolarityPositive)
				handle := promoteOK(t, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: handleSubject})
				var sourceID string
				sourceID, incoming = personOnNewSource(t, dir, userID, typeID, "male", tc.polarity)
				req := &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: dir, GraphId: soleGraphID(t, dir, sourceID)}
				if tc.fix {
					req.Fixed = []*engine.PromoteGraphAlignmentFixed{{SubjectId: incoming, HandleId: handle.Entity.GetId()}}
				}
				return req
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ProposePromoteGraphAlignmentResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				for _, row := range resp.GetRows() {
					if row.GetSubjectId() != incoming {
						continue
					}
					if row.GetTarget() != tc.wantTarget {
						t.Fatalf("row starts on %q, want %q", row.GetTarget(), tc.wantTarget)
					}
					return
				}
				t.Fatalf("no row for incoming subject %s in %+v", incoming, resp.GetRows())
			},
		})
	}
	runRPC(t, ProposePromoteGraphAlignment, tests)
}

// foundRecordFixture is a handle and, on a Source of its own, a person the shortlist does not reach from it.
type foundRecordFixture struct {
	dir, userID, typeID string
	handle              string
	members, claims     []string // Subject ids and their accepted claims' ids
	sourceID, incoming  string
	graphID             string // the Evidence graph of sourceID
}

// newFoundRecordFixture is a handle of one Mary Smith (male) and, on a Source
// of its own, a Polly Jones (male): a first name and surname the shortlist does
// not reach from the handle, so only a fixed pair joins them.
func newFoundRecordFixture(t *testing.T) foundRecordFixture {
	t.Helper()
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	dir, userID, _, typeID := subjectFixture(t)
	f := foundRecordFixture{dir: dir, userID: userID, typeID: typeID}
	_, subject := personOnNewSource(t, dir, userID, typeID, "male", observations.PolarityPositive)
	res := promoteOK(t, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: subject})
	f.handle = res.Entity.GetId()
	f.members = []string{subject}
	f.claims = []string{res.Claim.GetId()}
	f.sourceID, f.incoming = namedPersonOnNewSource(t, dir, userID, typeID, "Polly Jones", "male", observations.PolarityPositive)
	f.graphID = soleGraphID(t, dir, f.sourceID)
	return f
}

// A fixed pair on a record the shortlist does not reach: unfixed the row does
// not offer it; fixed the row holds it, decided and contested on a conflict.
func TestProposeHoldsAFoundRecord(t *testing.T) {
	f := newFoundRecordFixture(t)
	propose := func(fixed ...*engine.PromoteGraphAlignmentFixed) *engine.PromoteGraphAlignmentRow {
		t.Helper()
		var resp engine.ProposePromoteGraphAlignmentResponse
		mustRPC(t, ProposePromoteGraphAlignment, &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: f.dir, GraphId: f.graphID, Fixed: fixed}, &resp)
		for _, row := range resp.GetRows() {
			if row.GetSubjectId() == f.incoming {
				return row
			}
		}
		t.Fatalf("no row for %s in %+v", f.incoming, resp.GetRows())
		return nil
	}

	unfixed := propose()
	if unfixed.GetHandleId() == f.handle {
		t.Fatalf("unfixed row %+v targets the found record", unfixed)
	}
	for _, alt := range unfixed.GetAlternatives() {
		if alt.GetHandleId() == f.handle {
			t.Fatalf("unfixed row lists the found record: %+v", unfixed)
		}
	}

	row := propose(&engine.PromoteGraphAlignmentFixed{SubjectId: f.incoming, HandleId: f.handle})
	// The pair has no band, so the page starts the row unset; it keeps the
	// researcher's own decision.
	if row.GetTarget() != "" || row.GetHandleId() != "" {
		t.Fatalf("fixed row %+v, want it to start unset: the pair has no band", row)
	}
}

// A row whose earlier pair was released carries the move's reason and the
// neighbor that caused it; a row without a move carries neither.
func TestProposalProtoSendsMove(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	dir, _, _, _ := subjectFixture(t)
	a, b, n := uuid.New(), uuid.New(), uuid.New()
	prop := graphalign.Proposal{Rows: []graphalign.Row{
		{SubjectID: a[:], Move: &graphalign.Move{Reason: graphalign.ReasonNeighborPlaced, NeighborSubjectID: n[:]}},
		{SubjectID: b[:]},
	}}
	var resp *engine.ProposePromoteGraphAlignmentResponse
	if err := withProjectCatalog(dir, func(c *database.Catalog) error {
		var err error
		resp, err = proposalProto(c, prop)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	rows := resp.GetRows()
	if len(rows) != 2 {
		t.Fatalf("rows %d", len(rows))
	}
	if rows[0].GetMoveReason() != "neighbor_placed" || rows[0].GetMoveNeighborSubjectId() != n.String() {
		t.Fatalf("moved row %+v", rows[0])
	}
	if rows[1].GetMoveReason() != "" || rows[1].GetMoveNeighborSubjectId() != "" {
		t.Fatalf("plain row %+v", rows[1])
	}
}

// An unplaced row whose best record is in the weak band starts on that record,
// with its summary; one whose record another row took starts unset.
func TestProposalProtoStartsAWeakRow(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	dir, userID, _, typeID := subjectFixture(t)
	_, handleSubject := personOnNewSource(t, dir, userID, typeID, "male", observations.PolarityPositive)
	promoted := promoteOK(t, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: handleSubject})
	h, r := promoted.Entity.GetId(), promoted.Entity.GetRef()
	hid := uuid.MustParse(h)
	s := uuid.New()
	row := func(reason graphalign.Reason) graphalign.Row {
		return graphalign.Row{
			SubjectID: s[:], Kind: vocabkey.Person, Assessment: graphalign.AssessWeak, Reason: reason,
			Alternatives: []graphalign.Alternative{{HandleID: hid[:], Ref: r}},
		}
	}
	proposed := func(rw graphalign.Row) *engine.PromoteGraphAlignmentRow {
		t.Helper()
		var resp *engine.ProposePromoteGraphAlignmentResponse
		if err := withProjectCatalog(dir, func(c *database.Catalog) error {
			var err error
			resp, err = proposalProto(c, graphalign.Proposal{Rows: []graphalign.Row{rw}})
			return err
		}); err != nil {
			t.Fatal(err)
		}
		return resp.GetRows()[0]
	}

	started := proposed(row(graphalign.ReasonUnplaced))
	if started.GetTarget() != "handle" || started.GetHandleId() != h || started.GetHandleRef() != r ||
		started.GetSummary().GetPerson().GetEntity().GetId() != h {
		t.Fatalf("weak row %+v, want it started on %s (%s) with a person summary", started, h, r)
	}
	taken := proposed(row(graphalign.ReasonTaken))
	if taken.GetTarget() != "" || taken.GetHandleId() != "" || taken.Summary != nil {
		t.Fatalf("taken row %+v, want it unset with no summary", taken)
	}
}

func TestProposePromoteGraphAlignment(t *testing.T) {
	runRPC(t, ProposePromoteGraphAlignment, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "bad graph id",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: dir, GraphId: "nope"}
			},
			wantErr: true,
		},
		{
			name: "unknown graph",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: dir, GraphId: uuid.NewString()}
			},
			wantErr: true,
		},
		{
			name: "empty graph returns no rows",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, sourceID, _ := subjectFixture(t)
				return &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: dir, GraphId: soleGraphID(t, dir, sourceID)}
			},
			want: &engine.ProposePromoteGraphAlignmentResponse{},
		},
	})
}
