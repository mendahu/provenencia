package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
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
		s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: uid[:]}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, uid[:], subjects.CreateInput{SourceID: src.ID, SubjectTypeID: tid[:]}, nil)
		})
		if err != nil {
			return err
		}
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
		art, err := runArtifactCreate(c, uid[:], artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
		if err != nil {
			return err
		}
		_, err = writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: uid[:]}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, uid[:], citations.CreateInput{
				ArtifactID: art.ID, LocatorJSON: `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`,
			}, []observations.Input{
				{SubjectID: s.ID, PropertyID: name.ID, Name: namevaluetest.Western("Mary Smith")},
				{SubjectID: s.ID, PropertyID: sex.ID, ValueTermID: term.ID, Polarity: polarity},
			})
		})
		sourceID, subjectID = database.UUIDString(src.ID), database.UUIDString(s.ID)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return sourceID, subjectID
}

// A row-level conflict from a negative Observation reaches the response as a
// comparison line that names the incoming value as negative and cannot be
// pinned.
func TestProposeNegativeConflictLine(t *testing.T) {
	dir, userID, _, typeID := subjectFixture(t)
	_, handleSubject := personOnNewSource(t, dir, userID, typeID, "male", observations.PolarityPositive)
	handle := promoteOK(t, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: handleSubject})
	sourceID, incoming := personOnNewSource(t, dir, userID, typeID, "male", observations.PolarityNegative)

	out, err := ProposePromoteGraphAlignment(marshalProto(t, &engine.ProposePromoteGraphAlignmentRequest{
		ProjectDir: dir, SourceId: sourceID,
		Fixed: []*engine.PromoteGraphAlignmentFixed{{SubjectId: incoming, HandleId: handle.Entity.GetId()}},
	}))
	if err != nil {
		t.Fatal(err)
	}
	var resp engine.ProposePromoteGraphAlignmentResponse
	if err := proto.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	for _, row := range resp.GetRows() {
		for _, c := range row.GetComparisons() {
			if c.GetPropertyKey() == "sex_at_birth" && c.GetOutcome() == "conflict" {
				if !c.GetIncomingNegative() || c.GetPinned() || c.GetIncomingObservationId() != "" || c.GetIncomingDisplay() != "Male" {
					t.Fatalf("conflict line %+v", c)
				}
				return
			}
		}
	}
	t.Fatalf("no sex_at_birth conflict line in %+v", resp.GetRows())
}

// The contested flag reaches the response at any band: a negative Observation
// against a filed positive one conflicts whether the row is skipped, fixed or
// (with a positive Observation) not contested at all.
func TestProposeSendsContested(t *testing.T) {
	type fixture struct {
		polarity string
		fix      bool
	}
	cases := []struct {
		name       string
		fixture    fixture
		wantTarget string
		wantReason string
		wantFlag   bool
	}{
		{"negative, unfixed", fixture{observations.PolarityNegative, false}, "skip", "weak", true},
		{"negative, fixed", fixture{observations.PolarityNegative, true}, "handle", "decided", true},
		{"positive", fixture{observations.PolarityPositive, false}, "handle", "agrees", false},
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
				sourceID, incoming = personOnNewSource(t, dir, userID, typeID, "male", tc.fixture.polarity)
				req := &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: dir, SourceId: sourceID}
				if tc.fixture.fix {
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
					if row.GetTarget() != tc.wantTarget || row.GetReason() != tc.wantReason || row.GetContested() != tc.wantFlag {
						t.Fatalf("row target %q reason %q contested %v; want %q %q %v",
							row.GetTarget(), row.GetReason(), row.GetContested(), tc.wantTarget, tc.wantReason, tc.wantFlag)
					}
					return
				}
				t.Fatalf("no row for incoming subject %s in %+v", incoming, resp.GetRows())
			},
		})
	}
	runRPC(t, ProposePromoteGraphAlignment, tests)
}

// A row whose earlier pair was released carries the move's reason and the
// neighbor that caused it; a row without a move carries neither.
func TestProposalProtoSendsMove(t *testing.T) {
	dir, _, _, _ := subjectFixture(t)
	a, b, n := uuid.New(), uuid.New(), uuid.New()
	prop := graphalign.Proposal{Rows: []graphalign.Row{
		{SubjectID: a[:], Target: graphalign.TargetSkip, Move: &graphalign.Move{Reason: graphalign.ReasonNeighborPlaced, NeighborSubjectID: n[:]}},
		{SubjectID: b[:], Target: graphalign.TargetSkip},
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

func TestProposePromoteGraphAlignment(t *testing.T) {
	runRPC(t, ProposePromoteGraphAlignment, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "bad source id",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, _ := subjectFixture(t)
				return &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: dir, SourceId: "nope"}
			},
			wantErr: true,
		},
		{
			name: "empty source returns no rows",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, sourceID, _ := subjectFixture(t)
				return &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: dir, SourceId: sourceID}
			},
			want: &engine.ProposePromoteGraphAlignmentResponse{},
		},
		{
			// Two look-alike handles, so the row has a reason property and so
			// does its alternative; each goes out as a product property.
			name: "reason properties go out as product properties",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, typeID := subjectFixture(t)
				for range 2 {
					_, handleSubject := personOnNewSource(t, dir, userID, typeID, "male", observations.PolarityPositive)
					promoteOK(t, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: handleSubject})
				}
				sourceID, _ := personOnNewSource(t, dir, userID, typeID, "male", observations.PolarityPositive)
				return &engine.ProposePromoteGraphAlignmentRequest{ProjectDir: dir, SourceId: sourceID}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var resp engine.ProposePromoteGraphAlignmentResponse
				if err := proto.Unmarshal(out, &resp); err != nil {
					t.Fatal(err)
				}
				var named int
				check := func(what, key, origin string) {
					switch {
					case key == "" && origin != "":
						t.Fatalf("%s: origin %q without a key", what, origin)
					case key != "" && origin != vocabkey.OriginProvenencia:
						t.Fatalf("%s: key %q has origin %q, want %q", what, key, origin, vocabkey.OriginProvenencia)
					case key != "":
						named++
					}
				}
				for _, row := range resp.GetRows() {
					check("row", row.GetReasonPropertyKey(), row.GetReasonPropertyOrigin())
					for _, alt := range row.GetAlternatives() {
						check("alternative", alt.GetReasonPropertyKey(), alt.GetReasonPropertyOrigin())
					}
				}
				if named < 2 {
					t.Fatalf("%d reason properties in %+v, want the row's and an alternative's", named, resp.GetRows())
				}
			},
		},
	})
}

// A row whose handle has no accepted member claim has no exhibits, yet every
// comparison line reaches the response with the two values it compared.
func TestProposeRowWithoutExhibitsSendsValues(t *testing.T) {
	dir, userID, _, typeID := subjectFixture(t)
	_, handleSubject := personOnNewSource(t, dir, userID, typeID, "female", observations.PolarityPositive)
	handle := promoteOK(t, &engine.PromoteSubjectRequest{ProjectDir: dir, UserId: userID, SubjectId: handleSubject})
	entityID := uuid.MustParse(handle.Entity.GetId())
	if err := withProjectCatalog(dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		if _, err := db.Exec(`UPDATE identity_claims SET status = 'provisional' WHERE entity_id = ?`, entityID[:]); err != nil {
			return err
		}
		c.Graph().Drop()
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sourceID, incoming := personOnNewSource(t, dir, userID, typeID, "female", observations.PolarityPositive)

	out, err := ProposePromoteGraphAlignment(marshalProto(t, &engine.ProposePromoteGraphAlignmentRequest{
		ProjectDir: dir, SourceId: sourceID,
		Fixed: []*engine.PromoteGraphAlignmentFixed{{SubjectId: incoming, HandleId: handle.Entity.GetId()}},
	}))
	if err != nil {
		t.Fatal(err)
	}
	var resp engine.ProposePromoteGraphAlignmentResponse
	if err := proto.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	shown := map[string][2]string{}
	for _, row := range resp.GetRows() {
		for _, c := range row.GetComparisons() {
			if c.GetIncomingDisplay() == "" || c.GetMemberDisplay() == "" {
				t.Fatalf("a line without both values: %+v", c)
			}
			shown[c.GetPropertyKey()] = [2]string{c.GetIncomingDisplay(), c.GetMemberDisplay()}
		}
	}
	if shown["name"] != [2]string{"Mary Smith", "Mary Smith"} || shown["sex_at_birth"] != [2]string{"Female", "Female"} {
		t.Fatalf("lines %v", shown)
	}
}
