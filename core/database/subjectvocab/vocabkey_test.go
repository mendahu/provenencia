package subjectvocab

import (
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// vocabkey names the seed's primary kinds and the Property keys code reads by
// name; the seed stays the source of truth for what exists.
func TestVocabKeysMatchSeed(t *testing.T) {
	var roots []string
	for _, st := range seedTypes {
		if st.Role == RoleRoot {
			roots = append(roots, st.Key)
		}
	}
	want := vocabkey.PrimaryKinds()
	slices.Sort(roots)
	slices.Sort(want)
	if !slices.Equal(roots, want) {
		t.Errorf("seed root types %v, vocabkey.PrimaryKinds %v", roots, want)
	}

	seeded := map[string]bool{}
	for _, p := range seedProperties {
		seeded[p.Key] = true
	}
	seededTypes := map[string]bool{}
	for _, st := range seedTypes {
		seededTypes[st.Key] = true
	}
	for _, key := range []string{
		vocabkey.BridgeParticipation, vocabkey.BridgeRelationship,
		vocabkey.BridgeLocation, vocabkey.BridgePlaceRelationship,
	} {
		if !seededTypes[key] {
			t.Errorf("vocabkey bridge type %q is not a seeded type", key)
		}
	}
	for _, key := range []string{
		vocabkey.PropertyPerson, vocabkey.PropertyEvent, vocabkey.PropertyPlace,
		vocabkey.PropertyRelatedTo, vocabkey.PropertyFrom, vocabkey.PropertyTo,
	} {
		if !seeded[key] {
			t.Errorf("vocabkey Property %q is not a seeded Property", key)
		}
	}
	for _, key := range []string{
		vocabkey.PropertyName, vocabkey.PropertySexAtBirth, vocabkey.PropertyGender,
		vocabkey.PropertyEventType, vocabkey.PropertyDate, vocabkey.PropertyStartDate,
		vocabkey.PropertyEndDate, vocabkey.PropertyToponym,
		vocabkey.PropertyLikeness, vocabkey.PropertyEventPicture, vocabkey.PropertyPlacePicture,
	} {
		if !seeded[key] {
			t.Errorf("vocabkey Property %q is not a seeded Property", key)
		}
	}
	byKey := map[string]seedProperty{}
	for _, p := range seedProperties {
		byKey[p.Key] = p
	}
	for _, key := range vocabkey.GalleryKeys() {
		p, ok := byKey[key]
		if !ok {
			t.Errorf("GalleryKeys entry %q is not a seeded Property", key)
			continue
		}
		if p.ValueType != propertymeta.ValueTypeDepiction {
			t.Errorf("GalleryKeys entry %q ValueType %q, want depiction", key, p.ValueType)
		}
	}

	terms := map[string]bool{}
	for _, tm := range seedTerms {
		terms[tm.PropertyKey+"/"+tm.Key] = true
	}
	for _, tc := range []struct{ property, key string }{
		{vocabkey.PropertyEventType, string(vocabkey.EventBirth)},
		{vocabkey.PropertyEventType, string(vocabkey.EventDeath)},
		{vocabkey.PropertyEventType, string(vocabkey.EventBurial)},
		{vocabkey.PropertyEventType, string(vocabkey.EventMarriage)},
		{"role", string(vocabkey.RoleSubject)},
		{"role", string(vocabkey.RoleFather)},
		{"role", string(vocabkey.RoleMother)},
		{"relationship_type", string(vocabkey.TermSpouse)},
		{"relationship_type", string(vocabkey.TermParent)},
		{"relationship_type", string(vocabkey.TermGrandparent)},
		{"place_relationship_type", string(vocabkey.TermPartOf)},
		{vocabkey.PropertySexAtBirth, vocabkey.SexMale},
		{vocabkey.PropertySexAtBirth, vocabkey.SexFemale},
		{vocabkey.PropertyGender, vocabkey.GenderMan},
		{vocabkey.PropertyGender, vocabkey.GenderWoman},
	} {
		if !terms[tc.property+"/"+tc.key] {
			t.Errorf("vocabkey term %q is not a seeded %s term", tc.key, tc.property)
		}
	}
}
