package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/namevalues/namevaluestest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
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
		name, err := properties.Lookup(c, "name", properties.OriginProvenencia)
		if err != nil {
			return err
		}
		sex, err := properties.Lookup(c, "sex_at_birth", properties.OriginProvenencia)
		if err != nil {
			return err
		}
		term, err := propertyterms.Lookup(c, sex.ID, sexKey, propertyterms.OriginProvenencia)
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
				{SubjectID: s.ID, PropertyID: name.ID, Name: namevaluestest.Western("Mary Smith")},
				{SubjectID: s.ID, PropertyID: sex.ID, ValueTermID: term.ID, Polarity: polarity},
			})
		})
		sourceID, subjectID = uuidString(src.ID), uuidString(s.ID)
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
				if !c.GetIncomingNegative() || c.GetPinned() || c.GetIncomingObservationId() != "" || c.GetIncomingDisplay() != "male" {
					t.Fatalf("conflict line %+v", c)
				}
				return
			}
		}
	}
	t.Fatalf("no sex_at_birth conflict line in %+v", resp.GetRows())
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
	})
}
