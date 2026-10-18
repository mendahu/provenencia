package depictionvalue

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/depictionvalue/depictionvaluetest"
)

func TestReadable(t *testing.T) {
	id := []byte("fedcba9876543210")
	offPage := strings.Replace(depictionvaluetest.Region, `"artifact_page":1`, `"artifact_page":9`, 1)
	tests := []struct {
		name       string
		image      []byte
		region     string
		media      string
		wantRegion string
		wantImage  bool
	}{
		{name: "both halves", image: id, region: depictionvaluetest.Region, media: "image/png", wantRegion: depictionvaluetest.Region, wantImage: true},
		{name: "off-page region", image: id, region: offPage, media: "image/png", wantRegion: "", wantImage: true},
		{name: "not a png", image: id, region: depictionvaluetest.Region, media: "image/jpeg", wantRegion: depictionvaluetest.Region, wantImage: false},
		{name: "empty image id", image: nil, region: depictionvaluetest.Region, media: "image/png", wantRegion: depictionvaluetest.Region, wantImage: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Readable(Value{RegionJSON: tt.region, ImageFileID: tt.image}, depictionvaluetest.Locator, tt.media)
			if got.RegionJSON != tt.wantRegion {
				t.Fatalf("region %q", got.RegionJSON)
			}
			if (len(got.ImageFileID) > 0) != tt.wantImage {
				t.Fatalf("image %x", got.ImageFileID)
			}
		})
	}
	t.Run("ReadableImage", func(t *testing.T) {
		if got := ReadableImage(id, "image/png"); !bytes.Equal(got, id) {
			t.Fatalf("got %x", got)
		}
		if got := ReadableImage(nil, "image/png"); got != nil {
			t.Fatalf("got %x", got)
		}
	})
}
