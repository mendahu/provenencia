package handlers

import (
	"slices"
	"testing"

	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
)

// TestVocabularyWritesReturnEffects: every vocabulary write reports which
// vocabulary family it moved, so the app reloads the keys that show those
// labels and no others (S14-17 D-1, D-5, D-14). Each case works on rows it
// creates: seeded types, fields, terms and locked bindings refuse edits.
func TestVocabularyWritesReturnEffects(t *testing.T) {
	const source, property = "source_vocabulary", "property_vocabulary"
	want := func(t *testing.T, e *engine.WriteEffects, list string) {
		t.Helper()
		if !slices.Equal(e.GetLists(), []string{list}) || e.GetRevision() <= 0 {
			t.Fatalf("lists %v revision %d, want [%s] and a revision", e.GetLists(), e.GetRevision(), list)
		}
	}

	t.Run("source types and metadata fields report source_vocabulary", func(t *testing.T) {
		dir, userID, _ := sourceFixture(t)
		var created engine.CreateSourceTypeResponse
		mustRPC(t, CreateSourceType, &engine.CreateSourceTypeRequest{ProjectDir: dir, UserId: userID, Label: "Diary"}, &created)
		want(t, created.GetEffects(), source)
		typeID := created.Type.GetId()
		var updated engine.UpdateSourceTypeResponse
		mustRPC(t, UpdateSourceType, &engine.UpdateSourceTypeRequest{ProjectDir: dir, UserId: userID, TypeId: typeID, Label: "Journal"}, &updated)
		want(t, updated.GetEffects(), source)
		var deleted engine.DeleteSourceTypeResponse
		mustRPC(t, DeleteSourceType, &engine.DeleteSourceTypeRequest{ProjectDir: dir, UserId: userID, TypeId: typeID}, &deleted)
		want(t, deleted.GetEffects(), source)

		var field engine.CreateMetadataFieldResponse
		mustRPC(t, CreateMetadataField, &engine.CreateMetadataFieldRequest{ProjectDir: dir, UserId: userID, Label: "Volume", DataType: "text"}, &field)
		want(t, field.GetEffects(), source)
		fieldID := field.Field.GetId()
		var renamed engine.UpdateMetadataFieldResponse
		mustRPC(t, UpdateMetadataField, &engine.UpdateMetadataFieldRequest{ProjectDir: dir, UserId: userID, FieldId: fieldID, Label: "Book", DataType: "text"}, &renamed)
		want(t, renamed.GetEffects(), source)
		var gone engine.DeleteMetadataFieldResponse
		mustRPC(t, DeleteMetadataField, &engine.DeleteMetadataFieldRequest{ProjectDir: dir, UserId: userID, FieldId: fieldID}, &gone)
		want(t, gone.GetEffects(), source)
	})

	t.Run("properties, terms and bindings report property_vocabulary", func(t *testing.T) {
		dir, userID, _ := sourceFixture(t)
		var text engine.CreatePropertyResponse
		mustRPC(t, CreateProperty, &engine.CreatePropertyRequest{ProjectDir: dir, UserId: userID, Label: "Occupation", ValueType: "text"}, &text)
		want(t, text.GetEffects(), property)
		textID := text.Property.GetId()
		var renamed engine.UpdatePropertyResponse
		mustRPC(t, UpdateProperty, &engine.UpdatePropertyRequest{ProjectDir: dir, UserId: userID, PropertyId: textID, Label: "Trade", ValueType: "text"}, &renamed)
		want(t, renamed.GetEffects(), property)

		// A researcher cannot create a term-valued Property, so the terms go on a
		// seeded one that accepts them; each term row is the user's own.
		termedID := propertyIDByKey(t, dir, "event_type")
		wantTermRead := func(t *testing.T, e *engine.WriteEffects) {
			t.Helper()
			want(t, e, property)
			reads := e.GetReads()
			if len(reads) != 1 || reads[0].GetRead() != "property_terms" || !slices.Equal(reads[0].GetIds(), []string{termedID}) {
				t.Fatalf("reads %v, want property_terms [%s]", reads, termedID)
			}
		}
		var term engine.CreatePropertyTermResponse
		mustRPC(t, CreatePropertyTerm, &engine.CreatePropertyTermRequest{ProjectDir: dir, UserId: userID, PropertyId: termedID, Label: "Captain"}, &term)
		wantTermRead(t, term.GetEffects())
		var retermed engine.UpdatePropertyTermResponse
		mustRPC(t, UpdatePropertyTerm, &engine.UpdatePropertyTermRequest{ProjectDir: dir, UserId: userID, TermId: term.Term.GetId(), Label: "Major"}, &retermed)
		wantTermRead(t, retermed.GetEffects())
		var termGone engine.DeletePropertyTermResponse
		mustRPC(t, DeletePropertyTerm, &engine.DeletePropertyTermRequest{ProjectDir: dir, UserId: userID, TermId: term.Term.GetId()}, &termGone)
		wantTermRead(t, termGone.GetEffects())

		personID := subjectTypeIDByKey(t, dir, "person")
		var assigned engine.AssignSubjectTypePropertyResponse
		mustRPC(t, AssignSubjectTypeProperty, &engine.AssignSubjectTypePropertyRequest{ProjectDir: dir, UserId: userID, SubjectTypeId: personID, PropertyId: textID}, &assigned)
		want(t, assigned.GetEffects(), property)

		ids := bindingIDsForType(t, dir, personID)
		slices.Reverse(ids)
		var reordered engine.ReorderSubjectTypePropertiesResponse
		mustRPC(t, ReorderSubjectTypeProperties, &engine.ReorderSubjectTypePropertiesRequest{ProjectDir: dir, UserId: userID, SubjectTypeId: personID, PropertyIds: ids}, &reordered)
		want(t, reordered.GetEffects(), property)

		var removed engine.RemoveSubjectTypePropertyResponse
		mustRPC(t, RemoveSubjectTypeProperty, &engine.RemoveSubjectTypePropertyRequest{ProjectDir: dir, UserId: userID, SubjectTypeId: personID, PropertyId: textID}, &removed)
		want(t, removed.GetEffects(), property)
		var deleted engine.DeletePropertyResponse
		mustRPC(t, DeleteProperty, &engine.DeletePropertyRequest{ProjectDir: dir, UserId: userID, PropertyId: textID}, &deleted)
		want(t, deleted.GetEffects(), property)
	})

	t.Run("a cardinality change names the handles that observe it", func(t *testing.T) {
		dir, userID, _, artifactID, subjectID, _ := citationFixture(t)
		var created engine.CreatePropertyResponse
		mustRPC(t, CreateProperty, &engine.CreatePropertyRequest{ProjectDir: dir, UserId: userID, Label: "Parish", ValueType: "text"}, &created)
		propID := created.Property.GetId()
		mustRPC(t, AssignSubjectTypeProperty, &engine.AssignSubjectTypePropertyRequest{
			ProjectDir: dir, UserId: userID, SubjectTypeId: subjectTypeIDByKey(t, dir, "place"), PropertyId: propID,
		}, &engine.AssignSubjectTypePropertyResponse{})
		mustRPC(t, CreateCitationWithObservations, &engine.CreateCitationWithObservationsRequest{
			ProjectDir: dir, UserId: userID, ArtifactId: artifactID, GraphId: graphLinking(t, dir, artifactID), LocatorJson: validLocatorJSON,
			Observations: []*engine.ObservationDraft{{SubjectId: subjectID, PropertyId: propID, ValueText: "x"}},
		}, &engine.CreateCitationWithObservationsResponse{})
		promoted := promoteInDir(t, dir, userID, subjectID, "")
		handleID := database.UUIDString(promoted.Entity.ID)

		var updated engine.UpdatePropertyResponse
		mustRPC(t, UpdateProperty, &engine.UpdatePropertyRequest{
			ProjectDir: dir, UserId: userID, PropertyId: propID, Label: "Parish", ValueType: "text", Cardinality: "multiple",
		}, &updated)
		if !slices.Contains(updated.GetEffects().GetHandles(), handleID) {
			t.Fatalf("handles %v lack %s", updated.GetEffects().GetHandles(), handleID)
		}
		want(t, updated.GetEffects(), property)
	})

	t.Run("an unchanged update reports nothing", func(t *testing.T) {
		dir, userID, _ := sourceFixture(t)
		var created engine.CreateSourceTypeResponse
		mustRPC(t, CreateSourceType, &engine.CreateSourceTypeRequest{ProjectDir: dir, UserId: userID, Label: "Diary", Description: "kept"}, &created)
		var same engine.UpdateSourceTypeResponse
		mustRPC(t, UpdateSourceType, &engine.UpdateSourceTypeRequest{
			ProjectDir: dir, UserId: userID, TypeId: created.Type.GetId(), Label: "Diary", Description: "kept",
		}, &same)
		if len(same.GetEffects().GetLists()) != 0 || same.GetEffects().GetRevision() != 0 {
			t.Fatalf("unchanged update reported %v at revision %d", same.GetEffects().GetLists(), same.GetEffects().GetRevision())
		}
	})
}
