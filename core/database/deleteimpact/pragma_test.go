package deleteimpact

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
)

// TestDeletableKindsResolve checks that every kind Impact looks up names
// exactly one catalogmodel table with a primary key. The existence query is
// built from that row, so a kind with no table fails here instead of at delete.
func TestDeletableKindsResolve(t *testing.T) {
	seen := map[catalogmodel.Kind]int{}
	for _, spec := range catalogmodel.Tables {
		if spec.Kind == "" || !deletableKinds[spec.Kind] {
			continue
		}
		seen[spec.Kind]++
		if spec.Name == "" || spec.PK == "" {
			t.Errorf("%s is deletable but %q has no name or primary key", spec.Kind, spec.Name)
		}
	}
	for kind := range deletableKinds {
		if seen[kind] != 1 {
			t.Errorf("%s is deletable on %d catalogmodel tables, want 1", kind, seen[kind])
		}
		if _, ok := tableByKind(kind); !ok {
			t.Errorf("%s does not resolve to a lookup", kind)
		}
	}
}

// TestResourceAndOwnedHonesty checks delete policy against the catalog model.
// A resource foreign key must have a list probe, and an owned-outbound
// foreign key must be on its parent's release list. Whether the model matches
// SQLite is catalogmodel.TestPragmaHonesty.
func TestResourceAndOwnedHonesty(t *testing.T) {
	resourceVias := resourceInboundVias()
	owned := ownedColumns()
	for _, fk := range catalogmodel.FKs {
		key := fk.From + "." + fk.Column
		switch fk.Bucket {
		case catalogmodel.BucketResource:
			if !resourceVias[key] {
				t.Errorf("resource FK %s has no list probe", key)
			}
		case catalogmodel.BucketOwnedOutbound:
			if !owned[key] {
				t.Errorf("owned-outbound FK %s missing from release list", key)
			}
		}
	}
}

// TestPragmaHonestyProjectors requires a title and location projector, with a
// Section, for every deletable kind outside the infra, skip, and pool buckets.
// Impact names those rows in the confirm UI; a missing projector would leave
// that place blank.
func TestPragmaHonestyProjectors(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	id := make([]byte, 16)
	id[15] = 1
	for _, spec := range catalogmodel.Tables {
		if !isDeletableTable(spec) {
			continue
		}
		switch spec.Bucket {
		case catalogmodel.BucketInfra, catalogmodel.BucketSkip, catalogmodel.BucketPool:
			continue
		}
		_, loc, err := projectKind(tx, spec.Kind, probeRow{ID: id, Ref: "REF-TEST"})
		if err != nil {
			t.Errorf("%s project %v", spec.Kind, err)
			continue
		}
		if loc.Section == "" {
			t.Errorf("%s projector missing Section", spec.Kind)
		}
	}
}

func normalizeOnDelete(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" || s == "RESTRICT" {
		return "NO ACTION"
	}
	return s
}

// TestFacetReleaseHonesty: every CASCADE FK says whether its rows are audited
// research; audited ones have exactly one facetRelease on the right parent
// (unless that parent has no official delete yet), and silent ones have none.
func TestFacetReleaseHonesty(t *testing.T) {
	releases := map[string]facetRelease{}
	for _, f := range facetReleases() {
		if _, dup := releases[f.Via]; dup {
			t.Errorf("duplicate facet release %s", f.Via)
		}
		releases[f.Via] = f
	}
	fks := map[string]catalogmodel.FK{}
	for _, fk := range catalogmodel.FKs {
		fks[fk.From+"."+fk.Column] = fk
	}
	tableKind := map[string]catalogmodel.Kind{}
	deletable := map[string]bool{}
	for _, spec := range catalogmodel.Tables {
		if spec.Kind != "" {
			tableKind[spec.Name] = spec.Kind
		}
		deletable[spec.Name] = isDeletableTable(spec)
	}

	for key, fk := range fks {
		_, has := releases[key]
		if normalizeOnDelete(fk.OnDelete) != "CASCADE" {
			if fk.Audited {
				t.Errorf("%s is Audited but not CASCADE", key)
			}
			continue
		}
		switch {
		case fk.Audited && !has && deletable[fk.To]:
			t.Errorf("audited CASCADE %s has no facet release (parent %s is deletable)", key, fk.To)
		case !fk.Audited && has:
			t.Errorf("silent CASCADE %s has a facet release; mark it Audited", key)
		}
	}
	for via, f := range releases {
		fk, ok := fks[via]
		if !ok {
			t.Errorf("facet release %s is not a registered FK", via)
			continue
		}
		if tableKind[fk.To] != f.Parent {
			t.Errorf("facet release %s parent %s, FK points at %s", via, f.Parent, fk.To)
		}
		if normalizeOnDelete(fk.OnDelete) == "CASCADE" {
			if !fk.Audited {
				t.Errorf("facet release %s covers a CASCADE not marked Audited", via)
			}
			if f.Remaining == "" {
				t.Errorf("facet release %s needs a Remaining check (it has a CASCADE backstop)", via)
			}
		}
		if f.Release == nil {
			t.Errorf("facet release %s has no Release", via)
		}
		if f.Named && (f.Count == nil || f.List == nil || f.Child == "") {
			t.Errorf("named facet release %s needs Count, List, and Child", via)
		}
	}
}
