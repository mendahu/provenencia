package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mendahu/provenencia/core/apperr"
)

func TestIsIgnoresParams(t *testing.T) {
	base := apperr.New(apperr.CodeCatalogUnsupportedVersion, apperr.KindUser)
	wrapped := base.WithParams("5")
	if !errors.Is(wrapped, base) {
		t.Fatal("expected errors.Is to match by code")
	}
	if wrapped.Code() != apperr.CodeCatalogUnsupportedVersion {
		t.Fatalf("code = %q", wrapped.Code())
	}
	if got := wrapped.Params(); len(got) != 1 || got[0] != "5" {
		t.Fatalf("params = %v", got)
	}
}

func TestFromUnwraps(t *testing.T) {
	inner := apperr.New(apperr.CodeOnboardingBlankName, apperr.KindUser)
	err := fmt.Errorf("complete_onboarding: %w", inner)
	got := apperr.From(err)
	if got.Code() != apperr.CodeOnboardingBlankName {
		t.Fatalf("got %q", got.Code())
	}
}

func TestFromUnknown(t *testing.T) {
	got := apperr.From(errors.New("sqlite boom"))
	if got.Code() != apperr.CodeInternalUnknown {
		t.Fatalf("got %q", got.Code())
	}
	if got.Kind() != apperr.KindInternal {
		t.Fatalf("kind = %v", got.Kind())
	}
}
