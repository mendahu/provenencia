package handlers

import (
	"errors"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/locator"
	"google.golang.org/protobuf/proto"
)

const validLocatorJSON = `{"version":1,"selectors":[{"type":"page","artifact_page":12,"page_label":"10"}]}`

func citationFixture(t *testing.T) (projectDir, userID, sourceID, artifactID, placeSubjectID, toponymPropertyID string) {
	t.Helper()
	dir, userID, sourceID, _ := subjectFixture(t)
	aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
		ProjectDir: dir,
		UserId:     userID,
		SourceId:   sourceID,
		Label:      "Scan",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var art engine.CreateArtifactResponse
	if err := proto.Unmarshal(aout, &art); err != nil {
		t.Fatal(err)
	}
	typesOut, err := ListSubjectTypes(marshalProto(t, &engine.ListSubjectTypesRequest{ProjectDir: dir}))
	if err != nil {
		t.Fatal(err)
	}
	var types engine.ListSubjectTypesResponse
	if err := proto.Unmarshal(typesOut, &types); err != nil {
		t.Fatal(err)
	}
	placeTypeID := ""
	for _, typ := range types.Types {
		if typ.GetKey() == "place" {
			placeTypeID = typ.GetId()
			break
		}
	}
	if placeTypeID == "" {
		t.Fatal("place subject type missing")
	}
	sout, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
		ProjectDir:    dir,
		UserId:        userID,
		GraphId:       soleGraphID(t, dir, sourceID),
		SubjectTypeId: placeTypeID,
		Label:         "Boston",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var subj engine.CreateSubjectResponse
	if err := proto.Unmarshal(sout, &subj); err != nil {
		t.Fatal(err)
	}
	propsOut, err := ListProperties(marshalProto(t, &engine.ListPropertiesRequest{ProjectDir: dir}))
	if err != nil {
		t.Fatal(err)
	}
	var props engine.ListPropertiesResponse
	if err := proto.Unmarshal(propsOut, &props); err != nil {
		t.Fatal(err)
	}
	toponymID := ""
	for _, p := range props.Properties {
		if p.GetKey() == "toponym" {
			toponymID = p.GetId()
			break
		}
	}
	if toponymID == "" {
		t.Fatal("toponym property missing")
	}
	return dir, userID, sourceID, art.Artifact.GetId(), subj.Subject.GetId(), toponymID
}

const (
	likenessLocatorJSON = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`
	likenessRegionJSON  = `{"version":1,"selectors":[{"type":"page","artifact_page":1},{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.5}]}]}`
	likenessRegion2JSON = `{"version":1,"selectors":[{"type":"page","artifact_page":1},{"type":"region","unit":"normalized","points":[{"x":0.5,"y":0.5},{"x":0.9,"y":0.5},{"x":0.9,"y":0.9}]}]}`
)

var objectPathPattern = regexp.MustCompile(`^objects/[0-9a-f]{2}/[0-9a-f]{2}/[0-9a-f]{64}\.png$`)

// likenessFixture is depictionFixture for a Person and the likeness Property.
func likenessFixture(t *testing.T) (dir, userID, sourceID, artifactID, personID, likenessID string) {
	t.Helper()
	return depictionFixture(t, "person", "likeness")
}

// depictionFixture is a project whose Artifact has an 8 x 8 PNG File and whose
// Source has a Subject of kind with the given depiction Property key.
func depictionFixture(t *testing.T, kind, key string) (dir, userID, sourceID, artifactID, subjectID, propertyID string) {
	t.Helper()
	dir, userID, sourceID, _ = subjectFixture(t)
	typesOut, err := ListSubjectTypes(marshalProto(t, &engine.ListSubjectTypesRequest{ProjectDir: dir}))
	if err != nil {
		t.Fatal(err)
	}
	var types engine.ListSubjectTypesResponse
	if err := proto.Unmarshal(typesOut, &types); err != nil {
		t.Fatal(err)
	}
	typeID := ""
	for _, typ := range types.Types {
		if typ.GetKey() == kind {
			typeID = typ.GetId()
			break
		}
	}
	if typeID == "" {
		t.Fatalf("no subject type %q", kind)
	}
	aout, err := CreateArtifact(marshalProto(t, &engine.CreateArtifactRequest{
		ProjectDir: dir, UserId: userID, SourceId: sourceID, Label: "Photo",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var art engine.CreateArtifactResponse
	if err := proto.Unmarshal(aout, &art); err != nil {
		t.Fatal(err)
	}
	artifactID = art.Artifact.GetId()
	photo := filepath.Join(t.TempDir(), "photo.png")
	writePNG(t, photo)
	if _, err := IngestArtifactFile(marshalProto(t, &engine.IngestArtifactFileRequest{
		ProjectDir: dir, UserId: userID, ArtifactId: artifactID, Path: photo,
	})); err != nil {
		t.Fatal(err)
	}
	sout, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
		ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: typeID, Label: "Ada",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var subj engine.CreateSubjectResponse
	if err := proto.Unmarshal(sout, &subj); err != nil {
		t.Fatal(err)
	}
	return dir, userID, sourceID, artifactID, subj.Subject.GetId(), propertyIDByKey(t, dir, key)
}

// objectRelPaths lists every regular file under dir/objects, relative to dir, sorted.
func objectRelPaths(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	root := filepath.Join(dir, "objects")
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.Type().IsRegular() {
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(out)
	return out
}

func likenessDraft(personID, likenessID, region, imagePath string) *engine.ObservationDraft {
	return &engine.ObservationDraft{
		SubjectId:  personID,
		PropertyId: likenessID,
		Depiction:  &engine.DepictionValueInput{RegionJson: region, ImagePath: imagePath},
	}
}

// saveLikeness saves one likeness on a new Citation and returns the response.
func saveLikeness(t *testing.T, dir, userID, artifactID, personID, likenessID, crop string) *engine.CreateCitationWithObservationsResponse {
	t.Helper()
	out, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
		ProjectDir: dir, UserId: userID, ArtifactId: artifactID, LocatorJson: likenessLocatorJSON,
		Observations: []*engine.ObservationDraft{likenessDraft(personID, likenessID, likenessRegionJSON, crop)},
	}))
	if err != nil {
		t.Fatal(err)
	}
	var created engine.CreateCitationWithObservationsResponse
	if err := proto.Unmarshal(out, &created); err != nil {
		t.Fatal(err)
	}
	return &created
}

// sameLikeness asserts a read row carries the depiction id and value the save returned.
func sameLikeness(t *testing.T, got, saved *engine.Observation) {
	t.Helper()
	if got.GetValueDepictionId() != saved.GetValueDepictionId() || !proto.Equal(got.GetDepiction(), saved.GetDepiction()) {
		t.Fatalf("read %+v, saved %+v", got, saved)
	}
}

// checkSavedLikeness asserts the AC-1.1 shape of one saved likeness row.
func checkSavedLikeness(t *testing.T, dir string, row *engine.Observation) {
	t.Helper()
	if row.GetPropertyValueType() != "depiction" || row.GetValueDepictionId() == "" {
		t.Fatalf("row %+v", row)
	}
	if row.GetDepiction().GetRegionJson() != likenessRegionJSON {
		t.Fatalf("region %q", row.GetDepiction().GetRegionJson())
	}
	rel := row.GetDepiction().GetImageRelPath()
	if !objectPathPattern.MatchString(rel) {
		t.Fatalf("image path %q", rel)
	}
	f, err := os.Open(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cfg, err := png.DecodeConfig(f)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != 300 || cfg.Height != 200 {
		t.Fatalf("stored image %d x %d", cfg.Width, cfg.Height)
	}
}

func TestCreateCitationRefusesSymlinkedCrop(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	dir, userID, sourceID, artifactID, personID, likenessID := likenessFixture(t)
	before := objectRelPaths(t, dir)
	link := filepath.Join(t.TempDir(), "link.jpg")
	if err := os.Symlink(writeCrop(t, 300, 200, 128), link); err != nil {
		t.Fatal(err)
	}
	_, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
		ProjectDir: dir, UserId: userID, ArtifactId: artifactID, LocatorJson: likenessLocatorJSON,
		Observations: []*engine.ObservationDraft{likenessDraft(personID, likenessID, likenessRegionJSON, link)},
	}))
	if !errors.Is(err, ingest.ErrSymlink) {
		t.Fatalf("got %v", err)
	}
	if after := objectRelPaths(t, dir); strings.Join(after, ",") != strings.Join(before, ",") {
		t.Fatalf("objects changed: %v -> %v", before, after)
	}
	lout, err := ListCitationsByArtifact(marshalProto(t, &engine.ListCitationsByArtifactRequest{ProjectDir: dir, ArtifactId: artifactID, GraphId: soleGraphID(t, dir, sourceID)}))
	if err != nil {
		t.Fatal(err)
	}
	var list engine.ListCitationsByArtifactResponse
	if err := proto.Unmarshal(lout, &list); err != nil {
		t.Fatal(err)
	}
	if len(list.GetCitations()) != 0 {
		t.Fatalf("citations %+v", list.GetCitations())
	}
}

func TestLikenessReadHidesDamagedValue(t *testing.T) {
	offPage := strings.Replace(likenessRegionJSON, `"artifact_page":1`, `"artifact_page":9`, 1)
	tests := []struct {
		name       string
		damage     string
		args       []any
		wantRegion string
		wantImage  bool
	}{
		{name: "region off the page", damage: `UPDATE depiction_values SET region_json = ?`, args: []any{offPage}, wantRegion: "", wantImage: true},
		{name: "image not a PNG", damage: `UPDATE files SET media_type = 'application/pdf' WHERE id IN (SELECT image_file_id FROM depiction_values)`, wantRegion: likenessRegionJSON, wantImage: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() { _ = catalogsession.CloseAll() })
			dir, userID, sourceID, artifactID, personID, likenessID := likenessFixture(t)
			saved := saveLikeness(t, dir, userID, artifactID, personID, likenessID, writeCrop(t, 300, 200, 128))
			row := saved.Observations[0]
			damageCatalog(t, dir, tt.damage, tt.args...)
			check := func(t *testing.T, rows []*engine.Observation) {
				t.Helper()
				if len(rows) != 1 || rows[0].GetValueDepictionId() != row.GetValueDepictionId() {
					t.Fatalf("rows %+v", rows)
				}
				d := rows[0].GetDepiction()
				if d.GetRegionJson() != tt.wantRegion {
					t.Fatalf("region %q", d.GetRegionJson())
				}
				if (d.GetImageRelPath() != "") != tt.wantImage {
					t.Fatalf("image path %q", d.GetImageRelPath())
				}
			}
			gout, err := GetCitation(marshalProto(t, &engine.GetCitationRequest{ProjectDir: dir, CitationId: saved.Citation.GetId(), GraphId: graphLinking(t, dir, artifactID)}))
			if err != nil {
				t.Fatal(err)
			}
			var got engine.GetCitationResponse
			if err := proto.Unmarshal(gout, &got); err != nil {
				t.Fatal(err)
			}
			check(t, got.Observations)
			lout, err := ListObservationsByEvidenceGraph(marshalProto(t, &engine.ListObservationsByEvidenceGraphRequest{ProjectDir: dir, GraphId: soleGraphID(t, dir, sourceID)}))
			if err != nil {
				t.Fatal(err)
			}
			var list engine.ListObservationsByEvidenceGraphResponse
			if err := proto.Unmarshal(lout, &list); err != nil {
				t.Fatal(err)
			}
			check(t, list.Observations)
			if err := catalogsession.Do(dir, func(c *database.Catalog) error { return c.Graph().Verify() }); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCreateCitationWithObservations(t *testing.T) {
	var objectsBefore []string
	runRPC(t, CreateCitationWithObservations, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "creates citation with zero observations",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, _, _ := citationFixture(t)
				return &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "transcribe first",
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(created.Citation.GetRef(), "CIT-") {
					t.Fatalf("citation ref %q", created.Citation.GetRef())
				}
				if len(created.Observations) != 0 {
					t.Fatalf("observations %+v", created.Observations)
				}
				if created.Citation.GetTranscription() != "transcribe first" {
					t.Fatalf("transcription %q", created.Citation.GetTranscription())
				}
				if created.GetEffects() == nil || len(created.GetEffects().GetEvidenceGraphs()) == 0 {
					t.Fatalf("effects %+v", created.GetEffects())
				}
				cr := req.(*engine.CreateCitationWithObservationsRequest)
				assertLatestAuditAction(t, cr.ProjectDir, "create_citation_with_observations")
			},
		},
		{
			name: "bad locator",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				return &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: `{}`,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}
			},
			wantErr:   true,
			wantErrIs: locator.ErrInvalid,
		},
		{
			name: "saves a likeness and returns its image path",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, personID, likenessID := likenessFixture(t)
				objectsBefore = objectRelPaths(t, dir)
				return &engine.CreateCitationWithObservationsRequest{
					ProjectDir: dir, UserId: userID, ArtifactId: artifactID, LocatorJson: likenessLocatorJSON,
					Observations: []*engine.ObservationDraft{likenessDraft(personID, likenessID, likenessRegionJSON, writeCrop(t, 300, 200, 128))},
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				if len(created.Observations) != 1 {
					t.Fatalf("observations %+v", created.Observations)
				}
				dir := req.(*engine.CreateCitationWithObservationsRequest).ProjectDir
				checkSavedLikeness(t, dir, created.Observations[0])
				want := append(append([]string{}, objectsBefore...), created.Observations[0].GetDepiction().GetImageRelPath())
				sort.Strings(want)
				if got := objectRelPaths(t, dir); !slices.Equal(got, want) {
					t.Fatalf("objects %v, want %v", got, want)
				}
			},
		},
		{
			name: "creates citation with text observation",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				return &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "Boston",
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(out, &created); err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(created.Citation.GetRef(), "CIT-") {
					t.Fatalf("citation ref %q", created.Citation.GetRef())
				}
				if len(created.Observations) != 1 {
					t.Fatalf("observations %+v", created.Observations)
				}
				if !strings.HasPrefix(created.Observations[0].GetRef(), "OBS-") {
					t.Fatalf("obs ref %q", created.Observations[0].GetRef())
				}
				if created.Observations[0].GetValueText() != "Boston" {
					t.Fatalf("value %+v", created.Observations[0])
				}
				cr := req.(*engine.CreateCitationWithObservationsRequest)
				assertLatestAuditAction(t, cr.ProjectDir, "create_citation_with_observations")
			},
		},
	})
}

func TestAddObservationsToCitation(t *testing.T) {
	runRPC(t, AddObservationsToCitation, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "adds a likeness",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, personID, likenessID := likenessFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir: dir, UserId: userID, ArtifactId: artifactID, LocatorJson: likenessLocatorJSON,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.AddObservationsToCitationRequest{
					ProjectDir: dir, UserId: userID, CitationId: created.Citation.GetId(),
					Observations: []*engine.ObservationDraft{likenessDraft(personID, likenessID, likenessRegionJSON, writeCrop(t, 300, 200, 128))},
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var added engine.AddObservationsToCitationResponse
				if err := proto.Unmarshal(out, &added); err != nil {
					t.Fatal(err)
				}
				if len(added.Observations) != 1 {
					t.Fatalf("%+v", added.Observations)
				}
				checkSavedLikeness(t, req.(*engine.AddObservationsToCitationRequest).ProjectDir, added.Observations[0])
			},
		},
		{
			name: "appends observation",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Cambridge",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.AddObservationsToCitationRequest{
					ProjectDir: dir,
					UserId:     userID,
					CitationId: created.Citation.GetId(),
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "MA",
					}},
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var added engine.AddObservationsToCitationResponse
				if err := proto.Unmarshal(out, &added); err != nil {
					t.Fatal(err)
				}
				if len(added.Observations) != 1 || added.Observations[0].GetValueText() != "MA" {
					t.Fatalf("%+v", added.Observations)
				}
				ar := req.(*engine.AddObservationsToCitationRequest)
				assertLatestAuditAction(t, ar.ProjectDir, "add_observations")
			},
		},
	})
}

func TestListObservationsByEvidenceGraph(t *testing.T) {
	var saved *engine.CreateCitationWithObservationsResponse
	runRPC(t, ListObservationsByEvidenceGraph, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "returns a likeness",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, artifactID, personID, likenessID := likenessFixture(t)
				saved = saveLikeness(t, dir, userID, artifactID, personID, likenessID, writeCrop(t, 300, 200, 128))
				return &engine.ListObservationsByEvidenceGraphRequest{ProjectDir: dir, GraphId: soleGraphID(t, dir, sourceID)}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var list engine.ListObservationsByEvidenceGraphResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Observations) != 1 {
					t.Fatalf("%+v", list.Observations)
				}
				dir := req.(*engine.ListObservationsByEvidenceGraphRequest).ProjectDir
				checkSavedLikeness(t, dir, list.Observations[0])
				sameLikeness(t, list.Observations[0], saved.Observations[0])
				if err := catalogsession.Do(dir, func(c *database.Catalog) error { return c.Graph().Verify() }); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "lists by source",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, artifactID, placeID, propID := citationFixture(t)
				if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.ListObservationsByEvidenceGraphRequest{
					ProjectDir: dir,
					GraphId:    soleGraphID(t, dir, sourceID),
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListObservationsByEvidenceGraphResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Observations) != 1 {
					t.Fatalf("%+v", list.Observations)
				}
				if list.Observations[0].GetPropertyKey() != "toponym" {
					t.Fatalf("property summary %+v", list.Observations[0])
				}
				if list.Observations[0].GetValueText() != "Boston" {
					t.Fatalf("value %+v", list.Observations[0])
				}
			},
		},
		{
			name: "term label is the displayed value",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, artifactID, _, _ := citationFixture(t)
				typesOut, err := ListSubjectTypes(marshalProto(t, &engine.ListSubjectTypesRequest{ProjectDir: dir}))
				if err != nil {
					t.Fatal(err)
				}
				var types engine.ListSubjectTypesResponse
				if err := proto.Unmarshal(typesOut, &types); err != nil {
					t.Fatal(err)
				}
				eventTypeID := ""
				for _, typ := range types.Types {
					if typ.GetKey() == "event" {
						eventTypeID = typ.GetId()
						break
					}
				}
				sout, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
					ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: eventTypeID, Label: "Birth of Ada",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var event engine.CreateSubjectResponse
				if err := proto.Unmarshal(sout, &event); err != nil {
					t.Fatal(err)
				}
				propID := propertyIDByKey(t, dir, "event_type")
				termsOut, err := ListPropertyTerms(marshalProto(t, &engine.ListPropertyTermsRequest{
					ProjectDir: dir, PropertyId: propID,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var terms engine.ListPropertyTermsResponse
				if err := proto.Unmarshal(termsOut, &terms); err != nil {
					t.Fatal(err)
				}
				birthID := ""
				for _, term := range terms.Terms {
					if term.GetKey() == "birth" {
						birthID = term.GetId()
						break
					}
				}
				if birthID == "" {
					t.Fatal("birth term missing")
				}
				if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:   event.Subject.GetId(),
						PropertyId:  propID,
						ValueTermId: birthID,
					}},
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.ListObservationsByEvidenceGraphRequest{ProjectDir: dir, GraphId: soleGraphID(t, dir, sourceID)}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListObservationsByEvidenceGraphResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				var got *engine.Observation
				for _, row := range list.Observations {
					if row.GetPropertyKey() == "event_type" {
						got = row
						break
					}
				}
				if got == nil || got.GetValueText() != "Birth" {
					t.Fatalf("event type %+v", list.Observations)
				}
				if got.GetValueTermKey() != "birth" {
					t.Fatalf("value_term_key %q, want birth", got.GetValueTermKey())
				}
				if got.GetValueTermOrigin() != "provenencia" {
					t.Fatalf("value_term_origin %q, want provenencia", got.GetValueTermOrigin())
				}
			},
		},
	})
}

func TestCitationCountsByEvidenceGraph(t *testing.T) {
	runRPC(t, CitationCountsByEvidenceGraph, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{name: "missing graph_id", req: &engine.CitationCountsByEvidenceGraphRequest{ProjectDir: "x"}, wantErr: true},
		{name: "bad graph_id", req: &engine.CitationCountsByEvidenceGraphRequest{ProjectDir: "x", GraphId: "nope"}, wantErr: true},
		{
			name: "counts citations per artifact",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, artifactID, placeID, propID := citationFixture(t)
				if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				})); err != nil {
					t.Fatal(err)
				}
				if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston again",
					}},
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.CitationCountsByEvidenceGraphRequest{
					ProjectDir: dir,
					GraphId:    soleGraphID(t, dir, sourceID),
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.CitationCountsByEvidenceGraphResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Counts) != 1 || list.Counts[0].GetCount() != 2 {
					t.Fatalf("%+v", list.Counts)
				}
			},
		},
	})
}

func TestListCitationsByArtifact(t *testing.T) {
	runRPC(t, ListCitationsByArtifact, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "missing graph_id",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, artifactID, _, _ := citationFixture(t)
				return &engine.ListCitationsByArtifactRequest{ProjectDir: dir, ArtifactId: artifactID}
			},
			wantErr: true,
		},
		{
			name: "bad graph_id",
			reqFn: func(t *testing.T) proto.Message {
				dir, _, _, artifactID, _, _ := citationFixture(t)
				return &engine.ListCitationsByArtifactRequest{ProjectDir: dir, ArtifactId: artifactID, GraphId: "nope"}
			},
			wantErr: true,
		},
		{
			name: "lists citations with observation counts",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, artifactID, placeID, propID := citationFixture(t)
				if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "empty reading",
				})); err != nil {
					t.Fatal(err)
				}
				if _, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "Boston",
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				})); err != nil {
					t.Fatal(err)
				}
				return &engine.ListCitationsByArtifactRequest{
					ProjectDir: dir,
					ArtifactId: artifactID,
					GraphId:    soleGraphID(t, dir, sourceID),
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var list engine.ListCitationsByArtifactResponse
				if err := proto.Unmarshal(out, &list); err != nil {
					t.Fatal(err)
				}
				if len(list.Citations) != 2 {
					t.Fatalf("%+v", list.Citations)
				}
				counts := map[int32]int{}
				for _, row := range list.Citations {
					counts[row.GetObservationCount()]++
					if !strings.HasPrefix(row.GetCitation().GetRef(), "CIT-") {
						t.Fatalf("ref %q", row.GetCitation().GetRef())
					}
				}
				if counts[0] != 1 || counts[1] != 1 {
					t.Fatalf("counts %+v", counts)
				}
			},
		},
	})
}

func TestGetCitation(t *testing.T) {
	var saved *engine.CreateCitationWithObservationsResponse
	runRPC(t, GetCitation, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "returns a likeness",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, personID, likenessID := likenessFixture(t)
				saved = saveLikeness(t, dir, userID, artifactID, personID, likenessID, writeCrop(t, 300, 200, 128))
				return &engine.GetCitationRequest{ProjectDir: dir, CitationId: saved.Citation.GetId(), GraphId: graphLinking(t, dir, artifactID)}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var got engine.GetCitationResponse
				if err := proto.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if len(got.Observations) != 1 {
					t.Fatalf("%+v", got.Observations)
				}
				checkSavedLikeness(t, req.(*engine.GetCitationRequest).ProjectDir, got.Observations[0])
				sameLikeness(t, got.Observations[0], saved.Observations[0])
			},
		},
		{
			name: "loads citation with observations",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "Boston",
					CitationNotes: []string{"note A"},
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.GetCitationRequest{
					ProjectDir: dir,
					CitationId: created.Citation.GetId(),
					GraphId:    graphLinking(t, dir, artifactID),
				}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var got engine.GetCitationResponse
				if err := proto.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if got.Citation.GetTranscription() != "Boston" {
					t.Fatalf("citation %+v", got.Citation)
				}
				if len(got.Notes) != 1 || got.Notes[0] != "note A" {
					t.Fatalf("notes %+v", got.Notes)
				}
				if len(got.Observations) != 1 || got.Observations[0].GetValueText() != "Boston" {
					t.Fatalf("observations %+v", got.Observations)
				}
			},
		},
		{
			name: "term label comes from the loaded source",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, sourceID, artifactID, _, _ := citationFixture(t)
				typesOut, err := ListSubjectTypes(marshalProto(t, &engine.ListSubjectTypesRequest{ProjectDir: dir}))
				if err != nil {
					t.Fatal(err)
				}
				var types engine.ListSubjectTypesResponse
				if err := proto.Unmarshal(typesOut, &types); err != nil {
					t.Fatal(err)
				}
				eventTypeID := ""
				for _, typ := range types.Types {
					if typ.GetKey() == "event" {
						eventTypeID = typ.GetId()
						break
					}
				}
				sout, err := CreateSubject(marshalProto(t, &engine.CreateSubjectRequest{
					ProjectDir: dir, UserId: userID, GraphId: soleGraphID(t, dir, sourceID), SubjectTypeId: eventTypeID, Label: "Birth of Ada",
				}))
				if err != nil {
					t.Fatal(err)
				}
				var event engine.CreateSubjectResponse
				if err := proto.Unmarshal(sout, &event); err != nil {
					t.Fatal(err)
				}
				propID := propertyIDByKey(t, dir, "event_type")
				termsOut, err := ListPropertyTerms(marshalProto(t, &engine.ListPropertyTermsRequest{
					ProjectDir: dir, PropertyId: propID,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var terms engine.ListPropertyTermsResponse
				if err := proto.Unmarshal(termsOut, &terms); err != nil {
					t.Fatal(err)
				}
				birthID := ""
				for _, term := range terms.Terms {
					if term.GetKey() == "birth" {
						birthID = term.GetId()
						break
					}
				}
				if birthID == "" {
					t.Fatal("birth term missing")
				}
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:   event.Subject.GetId(),
						PropertyId:  propID,
						ValueTermId: birthID,
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.GetCitationRequest{ProjectDir: dir, CitationId: created.Citation.GetId(), GraphId: graphLinking(t, dir, artifactID)}
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				var got engine.GetCitationResponse
				if err := proto.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if len(got.Observations) != 1 || got.Observations[0].GetPropertyKey() != "event_type" || got.Observations[0].GetValueText() != "Birth" {
					t.Fatalf("observations %+v", got.Observations)
				}
			},
		},
	})
}

func TestGetCitationSecondReadUsesTheLoadedSource(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	dir, userID, _, artifactID, placeID, propID := citationFixture(t)
	createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
		ProjectDir:  dir,
		UserId:      userID,
		ArtifactId:  artifactID,
		LocatorJson: validLocatorJSON,
		Observations: []*engine.ObservationDraft{
			{SubjectId: placeID, PropertyId: propID, ValueText: "Boston"},
			{SubjectId: placeID, PropertyId: propID, ValueText: "Cambridge"},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	var created engine.CreateCitationWithObservationsResponse
	if err := proto.Unmarshal(createOut, &created); err != nil {
		t.Fatal(err)
	}
	otherOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
		ProjectDir:  dir,
		UserId:      userID,
		ArtifactId:  artifactID,
		LocatorJson: validLocatorJSON,
		Observations: []*engine.ObservationDraft{
			{SubjectId: placeID, PropertyId: propID, ValueText: "Elsewhere"},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	var other engine.CreateCitationWithObservationsResponse
	if err := proto.Unmarshal(otherOut, &other); err != nil {
		t.Fatal(err)
	}
	req := marshalProto(t, &engine.GetCitationRequest{ProjectDir: dir, CitationId: created.Citation.GetId(), GraphId: graphLinking(t, dir, artifactID)})
	first, err := GetCitation(req)
	if err != nil {
		t.Fatal(err)
	}
	var got engine.GetCitationResponse
	if err := proto.Unmarshal(first, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Observations) != 2 {
		t.Fatalf("observations %+v", got.Observations)
	}
	var n int
	if err := catalogsession.Do(dir, func(c *database.Catalog) error {
		c.Graph().SetQueryCounterForTest(func() { n++ })
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	second, err := GetCitation(req)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("second read issued %d graph queries", n)
	}
	if err := proto.Unmarshal(second, &got); err != nil {
		t.Fatal(err)
	}
	texts := map[string]int{}
	for _, row := range got.Observations {
		texts[row.GetValueText()]++
	}
	if texts["Boston"] != 1 || texts["Cambridge"] != 1 {
		t.Fatalf("observations %+v", got.Observations)
	}
	otherReq := marshalProto(t, &engine.GetCitationRequest{ProjectDir: dir, CitationId: other.Citation.GetId(), GraphId: graphLinking(t, dir, artifactID)})
	otherGot, err := GetCitation(otherReq)
	if err != nil {
		t.Fatal(err)
	}
	var listed engine.GetCitationResponse
	if err := proto.Unmarshal(otherGot, &listed); err != nil {
		t.Fatal(err)
	}
	if n != 0 || len(listed.Observations) != 1 || listed.Observations[0].GetValueText() != "Elsewhere" {
		t.Fatalf("queries %d observations %+v", n, listed.Observations)
	}
}

func TestUpdateCitation(t *testing.T) {
	runRPC(t, UpdateCitation, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "updates citation fields only",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ArtifactId:    artifactID,
					LocatorJson:   validLocatorJSON,
					Transcription: "was",
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.UpdateCitationRequest{
					ProjectDir:    dir,
					UserId:        userID,
					CitationId:    created.Citation.GetId(),
					LocatorJson:   validLocatorJSON,
					Transcription: "now",
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var updated engine.UpdateCitationResponse
				if err := proto.Unmarshal(out, &updated); err != nil {
					t.Fatal(err)
				}
				if updated.Citation.GetTranscription() != "now" {
					t.Fatalf("transcription %q", updated.Citation.GetTranscription())
				}
				ur := req.(*engine.UpdateCitationRequest)
				assertLatestAuditAction(t, ur.ProjectDir, "update_citation")
			},
		},
	})
}

func TestUpdateObservation(t *testing.T) {
	// likenessUpdate saves a likeness and returns the request that edits it.
	likenessUpdate := func(t *testing.T) (*engine.UpdateObservationRequest, *engine.Observation) {
		dir, userID, _, artifactID, personID, likenessID := likenessFixture(t)
		saved := saveLikeness(t, dir, userID, artifactID, personID, likenessID, writeCrop(t, 300, 200, 128))
		row := saved.Observations[0]
		return &engine.UpdateObservationRequest{ProjectDir: dir, UserId: userID, Observation: proto.Clone(row).(*engine.Observation)}, row
	}
	updated := func(t *testing.T, out []byte) *engine.Observation {
		t.Helper()
		var resp engine.UpdateObservationResponse
		if err := proto.Unmarshal(out, &resp); err != nil {
			t.Fatal(err)
		}
		return resp.Observation
	}
	// saved is the row before the update; the request carries the same row.
	var saved *engine.Observation
	runRPC(t, UpdateObservation, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "replaces a likeness image",
			reqFn: func(t *testing.T) proto.Message {
				req, row := likenessUpdate(t)
				saved = row
				req.DepictionInput = &engine.DepictionValueInput{RegionJson: likenessRegion2JSON, ImagePath: writeCrop(t, 200, 200, 64)}
				return req
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				got := updated(t, out)
				dir := req.(*engine.UpdateObservationRequest).ProjectDir
				if got.GetValueDepictionId() == "" || got.GetValueDepictionId() == saved.GetValueDepictionId() {
					t.Fatalf("value id %q was %q", got.GetValueDepictionId(), saved.GetValueDepictionId())
				}
				if got.GetDepiction().GetRegionJson() != likenessRegion2JSON {
					t.Fatalf("region %q", got.GetDepiction().GetRegionJson())
				}
				oldPath, newPath := saved.GetDepiction().GetImageRelPath(), got.GetDepiction().GetImageRelPath()
				if newPath == oldPath || !objectPathPattern.MatchString(newPath) {
					t.Fatalf("path %q was %q", newPath, oldPath)
				}
				onDisk := strings.Join(objectRelPaths(t, dir), ",")
				if strings.Contains(onDisk, oldPath) || !strings.Contains(onDisk, newPath) {
					t.Fatalf("objects %s", onDisk)
				}
			},
		},
		{
			name: "keeps a likeness on a polarity change",
			reqFn: func(t *testing.T) proto.Message {
				req, row := likenessUpdate(t)
				saved = row
				req.Observation.Polarity = "negative"
				return req
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				got := updated(t, out)
				dir := req.(*engine.UpdateObservationRequest).ProjectDir
				if got.GetPolarity() != "negative" || got.GetValueDepictionId() != saved.GetValueDepictionId() ||
					!proto.Equal(got.GetDepiction(), saved.GetDepiction()) {
					t.Fatalf("got %+v want %+v", got, saved)
				}
				if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(got.GetDepiction().GetImageRelPath()))); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "ignores a sent image path",
			reqFn: func(t *testing.T) proto.Message {
				req, row := likenessUpdate(t)
				saved = row
				other := ""
				for _, p := range objectRelPaths(t, req.ProjectDir) {
					if p != row.GetDepiction().GetImageRelPath() {
						other = p
					}
				}
				if other == "" {
					t.Fatal("artifact object missing")
				}
				req.Observation.Depiction.ImageRelPath = other
				return req
			},
			after: func(t *testing.T, out []byte, _ proto.Message) {
				got := updated(t, out)
				if !proto.Equal(got.GetDepiction(), saved.GetDepiction()) || got.GetValueDepictionId() != saved.GetValueDepictionId() {
					t.Fatalf("got %+v want %+v", got, saved)
				}
			},
		},
		{
			name: "updates one observation",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				obs := created.Observations[0]
				obs.ValueText = "Salem"
				return &engine.UpdateObservationRequest{
					ProjectDir:  dir,
					UserId:      userID,
					Observation: obs,
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var updated engine.UpdateObservationResponse
				if err := proto.Unmarshal(out, &updated); err != nil {
					t.Fatal(err)
				}
				if updated.Observation.GetValueText() != "Salem" || updated.Observation.GetPropertyKey() == "" {
					t.Fatalf("%+v", updated.Observation)
				}
				ur := req.(*engine.UpdateObservationRequest)
				assertLatestAuditAction(t, ur.ProjectDir, "update_observation")
			},
		},
	})
}

func TestDeleteCitation(t *testing.T) {
	runRPC(t, DeleteCitation, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "empty citation erases",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, _, _ := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteCitationRequest{
					ProjectDir: dir,
					UserId:     userID,
					CitationId: created.Citation.GetId(),
				}
			},
			want: &engine.DeleteCitationResponse{},
			after: func(t *testing.T, _ []byte, req proto.Message) {
				dr := req.(*engine.DeleteCitationRequest)
				assertLatestAuditAction(t, dr.ProjectDir, "delete_citation")
			},
		},
		{
			name: "inbound observation refuses",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteCitationRequest{
					ProjectDir: dir,
					UserId:     userID,
					CitationId: created.Citation.GetId(),
				}
			},
			wantErr:   true,
			wantErrIs: citations.ErrInUse,
		},
	})
}

func TestDeleteObservation(t *testing.T) {
	runRPC(t, DeleteObservation, []rpcTest{
		{name: "bad proto", raw: []byte{0xff}, wantErr: true},
		{
			name: "deletes observation and keeps citation",
			reqFn: func(t *testing.T) proto.Message {
				dir, userID, _, artifactID, placeID, propID := citationFixture(t)
				createOut, err := CreateCitationWithObservations(marshalProto(t, &engine.CreateCitationWithObservationsRequest{
					ProjectDir:  dir,
					UserId:      userID,
					ArtifactId:  artifactID,
					LocatorJson: validLocatorJSON,
					Observations: []*engine.ObservationDraft{{
						SubjectId:  placeID,
						PropertyId: propID,
						ValueText:  "Boston",
					}},
				}))
				if err != nil {
					t.Fatal(err)
				}
				var created engine.CreateCitationWithObservationsResponse
				if err := proto.Unmarshal(createOut, &created); err != nil {
					t.Fatal(err)
				}
				return &engine.DeleteObservationRequest{
					ProjectDir:    dir,
					UserId:        userID,
					ObservationId: created.Observations[0].GetId(),
				}
			},
			after: func(t *testing.T, out []byte, req proto.Message) {
				var deleted engine.DeleteObservationResponse
				if err := proto.Unmarshal(out, &deleted); err != nil {
					t.Fatal(err)
				}
				dr := req.(*engine.DeleteObservationRequest)
				assertLatestAuditAction(t, dr.ProjectDir, "delete_observation")
			},
		},
	})
}
