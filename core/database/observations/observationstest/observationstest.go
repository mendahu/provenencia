// Package observationstest saves likenesses through the real write path for
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
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/writes"
)

// A depiction's region sits on the page its Citation cites.
const (
	LikenessLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`
	LikenessRegion  = `{"version":1,"selectors":[{"type":"page","artifact_page":1},{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.5}]}]}`
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
// Observations in one write, discarding any stage left over afterwards.
func SaveCitation(c *database.Catalog, userID []byte, art artifacts.Artifact, inputs ...observations.Input) (citations.CreateResult, []rowchange.Change, error) {
	discard, err := ingest.StageDepictions(c, inputs)
	if err != nil {
		return citations.CreateResult{}, nil, err
	}
	defer discard()
	var changes []rowchange.Change
	res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		r, ch, err := citations.CreateWithObservations(tx, userID, citations.CreateInput{ArtifactID: art.ID, LocatorJSON: LikenessLocator}, inputs)
		changes = ch
		return r, ch, err
	})
	return res, changes, err
}

// UpdateObservation stages the input's depiction, then updates the Observation.
func UpdateObservation(c *database.Catalog, userID []byte, in observations.Input) ([]rowchange.Change, error) {
	discard, err := ingest.StageDepictions(c, []observations.Input{in})
	if err != nil {
		return nil, err
	}
	defer discard()
	var changes []rowchange.Change
	_, _, err = writes.Run(c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		_, ch, err := observations.Update(tx, userID, in)
		changes = ch
		return struct{}{}, ch, err
	})
	return changes, err
}
