package deepfixture_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/deepfixture"
)

func TestGenerateReachesAFewHundredHandles(t *testing.T) {
	cat, err := deepfixture.Generate(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cat.Cat.Close() })
	db, err := cat.Cat.DB()
	if err != nil {
		t.Fatal(err)
	}
	var handles int
	if err := db.QueryRow(`SELECT COUNT(*) FROM canonical_entities WHERE merged_into_id IS NULL`).Scan(&handles); err != nil {
		t.Fatal(err)
	}
	if handles < 200 {
		t.Fatalf("handles %d, want a few hundred", handles)
	}
	var members int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM identity_claims
		WHERE entity_id = ? AND status = 'accepted'`, cat.Person).Scan(&members); err != nil {
		t.Fatal(err)
	}
	if members < 10 {
		t.Fatalf("focal person members %d, want about ten Sources", members)
	}
	if len(cat.Place) != 16 || len(cat.NameObservation.ID) != 16 || len(cat.ObitSubjects) < 8 {
		t.Fatalf("focus incomplete place=%d obs=%d obit=%d", len(cat.Place), len(cat.NameObservation.ID), len(cat.ObitSubjects))
	}
}
