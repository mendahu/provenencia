// Package observationstest saves depictions through the real write path for
// test fixtures: a photo Artifact, a staged crop, and the Citation with its
// Observations, so the same pixels share one File and upkeep runs as in the
// product.
package observationstest

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/depictionvalue/depictionvaluetest"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// A depiction's region sits on the page its Citation cites.
const (
	LikenessLocator = depictionvaluetest.Locator
	LikenessRegion  = depictionvaluetest.Region
)

func writePNG(t testing.TB, path string, w, h int, c color.Color) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

// Crop writes a w x h PNG filled with gray shade and returns its path.
func Crop(t testing.TB, w, h int, shade uint8) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "crop.png")
	writePNG(t, path, w, h, color.Gray{Y: shade})
	return path
}

// FileArtifact imports a 300 x 200 red PNG and creates an Artifact on its File.
func FileArtifact(t testing.TB, c *database.Catalog, userID, sourceID []byte) artifacts.Artifact {
	t.Helper()
	path := filepath.Join(t.TempDir(), "page.png")
	writePNG(t, path, 300, 200, color.RGBA{R: 255, A: 255})
	res, err := ingest.File(c, path, userID)
	if err != nil {
		t.Fatal(err)
	}
	art, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: sourceID, FileID: res.File.ID, Label: "Photograph"})
		})
	if err != nil {
		t.Fatal(err)
	}
	return art
}

// SaveCitation stages every depiction, then files the Citation and its
// Observations through Cite, discarding any stage left over afterwards.
func SaveCitation(c *database.Catalog, userID, graphID []byte, art artifacts.Artifact, inputs ...observations.Input) (citations.CreateResult, []rowchange.Change, error) {
	stages, err := ingest.StageDepictions(c.Dir(), inputs)
	if err != nil {
		return citations.CreateResult{}, nil, err
	}
	defer stages.Settle(c)
	return Cite(c, userID, graphID, art.ID, LikenessLocator, inputs...)
}

// Cite files a Citation of the Artifact under the Evidence graph with its
// Observations in one write, and returns the changes the write made.
func Cite(c *database.Catalog, userID, graphID, artifactID []byte, locator string, inputs ...observations.Input) (citations.CreateResult, []rowchange.Change, error) {
	var changes []rowchange.Change
	res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		r, ch, err := citations.CreateWithObservations(tx, userID, citations.CreateInput{ArtifactID: artifactID, GraphID: graphID, LocatorJSON: locator}, inputs)
		changes = ch
		return r, ch, err
	})
	return res, changes, err
}

// UpdateObservation stages the input's depiction, then updates the Observation.
func UpdateObservation(c *database.Catalog, userID []byte, in observations.Input) ([]rowchange.Change, error) {
	stages, err := ingest.StageDepictions(c.Dir(), []observations.Input{in})
	if err != nil {
		return nil, err
	}
	defer stages.Settle(c)
	var changes []rowchange.Change
	_, _, err = writes.Run(c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		_, ch, err := observations.Update(tx, userID, in)
		changes = ch
		return struct{}{}, ch, err
	})
	return changes, err
}

// SaveLikeness files a likeness of the Subject on a fresh photo under the
// Source, through the real save path: shade picks its 40 x 30 crop.
func SaveLikeness(t testing.TB, c *database.Catalog, userID, sourceID, subjectID []byte, shade uint8) observations.Observation {
	t.Helper()
	return SaveDepiction(t, c, userID, sourceID, subjectID, vocabkey.PropertyLikeness, shade)
}

// SaveNegativeLikeness is SaveLikeness for a negative: the photo is not the Subject.
func SaveNegativeLikeness(t testing.TB, c *database.Catalog, userID, sourceID, subjectID []byte, shade uint8) observations.Observation {
	t.Helper()
	return saveDepiction(t, c, userID, sourceID, subjectID, vocabkey.PropertyLikeness, shade, observations.PolarityNegative)
}

// SaveDepiction files a depiction Observation under Property key (proveniencia
// origin) on a fresh photo under the Source: shade picks its 40 x 30 crop.
func SaveDepiction(t testing.TB, c *database.Catalog, userID, sourceID, subjectID []byte, key string, shade uint8) observations.Observation {
	t.Helper()
	return saveDepiction(t, c, userID, sourceID, subjectID, key, shade, observations.PolarityPositive)
}

func saveDepiction(t testing.TB, c *database.Catalog, userID, sourceID, subjectID []byte, key string, shade uint8, polarity string) observations.Observation {
	t.Helper()
	prop, err := properties.Lookup(c, key, vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	art := FileArtifact(t, c, userID, sourceID)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var graphID []byte
	if err := db.QueryRow(`SELECT graph_id FROM subjects WHERE id = ?`, subjectID).Scan(&graphID); err != nil {
		t.Fatal(err)
	}
	res, _, err := SaveCitation(c, userID, graphID, art, observations.Input{
		SubjectID: subjectID, PropertyID: prop.ID, Polarity: polarity,
		Depiction: &observations.DepictionInput{RegionJSON: LikenessRegion, ImagePath: Crop(t, 40, 30, shade)},
	})
	if err != nil {
		t.Fatal(err)
	}
	return res.Observations[0]
}

// ImageOf is the Observation's current image File id.
func ImageOf(t testing.TB, c *database.Catalog, observationID []byte) []byte {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var id []byte
	if err := db.QueryRow(`SELECT dv.image_file_id FROM observations o
		JOIN depiction_values dv ON dv.id = o.value_depiction_id WHERE o.id = ?`, observationID).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}
