import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct SourceTypesModelTests {
    private let projectDir = "/tmp/types.provenencia"
    private let userID = "00000000-0000-7000-8000-000000000001"

    private func seededType(id: String = "t1", label: String = "Book", usedBy: Int = 0) -> CatalogSourceType {
        CatalogSourceType(
            id: id, key: "book", origin: "provenencia", label: label,
            description: "A published monograph.", usedBy: usedBy
        )
    }

    private func userType(id: String = "t2", label: String = "Family scrapbook", usedBy: Int = 0) -> CatalogSourceType {
        CatalogSourceType(
            id: id, key: FieldSlug.kebab(label), origin: "user", label: label,
            description: "Nan's albums.", usedBy: usedBy
        )
    }

    private func pluginType(id: String = "t3", label: String = "Grave memorial") -> CatalogSourceType {
        CatalogSourceType(
            id: id, key: "grave-memorial", origin: "plugin:findagrave", label: label,
            description: "From the plugin.", usedBy: 5
        )
    }

    private func field(id: String, label: String, dataType: String = "text", usedBy: Int = 0) -> CatalogMetadataField {
        CatalogMetadataField(
            id: id, key: FieldSlug.kebab(label), origin: "provenencia", label: label,
            dataType: dataType, description: "", usedBy: usedBy
        )
    }

    private func makeModel(
        store: FakeStore = FakeStore(),
        types: [CatalogSourceType] = [],
        fields: [CatalogMetadataField] = [],
        suggestions: [String: [CatalogTypeSuggestion]] = [:],
        catalogCounts: CatalogCounts? = nil
    ) -> SourceTypesModel {
        store.sourceTypesByProject[projectDir] = types
        store.fieldsByProject[projectDir] = fields
        store.suggestionsByType = suggestions
        return SourceTypesModel(
            projectDir: projectDir,
            userID: userID,
            store: store,
            catalogCounts: catalogCounts
        )
    }

    // MARK: List

    @Test func loadPopulatesTypesPoolAndCounts() async {
        let model = makeModel(
            types: [seededType(), userType(), pluginType()],
            fields: [field(id: "f1", label: "Author")]
        )
        await model.load()
        #expect(model.types.count == 3)
        #expect(model.fields.count == 1)
        #expect(model.types.seededCount == 1)
        #expect(model.types.userCount == 1)
        #expect(model.types.pluginCount == 1)
    }

    @Test func searchFiltersByLabelKeyAndDescription() async {
        let model = makeModel(types: [seededType(), userType()])
        await model.load()

        model.query = "scrapbook"
        #expect(model.visibleTypes.map(\.id) == ["t2"])

        model.query = "book"
        #expect(model.visibleTypes.map(\.id).sorted() == ["t1", "t2"])

        model.query = "monograph"
        #expect(model.visibleTypes.map(\.id) == ["t1"])

        model.query = "nomatch"
        #expect(model.visibleTypes.isEmpty)
    }

    @Test func sortingTogglesWithinAColumnAndResetsAcrossColumns() async {
        let store = FakeStore()
        let model = makeModel(
            store: store,
            types: [seededType(), userType()],
            fields: [field(id: "f1", label: "Author")],
            suggestions: ["t2": [CatalogTypeSuggestion(field: field(id: "f1", label: "Author"), sortOrder: 0)]]
        )
        await model.load()
        #expect(model.visibleTypes.map(\.id) == ["t1", "t2"])

        model.sortBy("label")
        #expect(!model.sortAscending)
        #expect(model.visibleTypes.map(\.id) == ["t2", "t1"])

        // Moving to a new column starts ascending again, not carrying the
        // previous column's direction.
        model.sortBy("fields")
        #expect(model.sortColumn == .fields)
        #expect(model.sortAscending)
        #expect(model.visibleTypes.map(\.id) == ["t1", "t2"])
    }

    // MARK: Selection and locking

    @Test func selectingSeededOrUserTypeEditsAndPluginTypeLocks() async {
        let model = makeModel(types: [seededType(), userType(), pluginType()])
        await model.load()

        model.select("t1")
        #expect(!model.isSelectedTypeLocked)
        #expect(model.canEditAssociations)

        model.select("t3")
        #expect(model.isSelectedTypeLocked)
        #expect(!model.canEditAssociations)
        #expect(!model.canDeleteSelectedType)
    }

    @Test func selectingLoadsThatTypesSuggestionsInStoredOrder() async {
        let author = field(id: "f1", label: "Author")
        let publisher = field(id: "f2", label: "Publisher")
        let model = makeModel(
            types: [seededType()],
            fields: [author, publisher],
            // Publisher was assigned first, so stored order and label order disagree.
            suggestions: ["t1": [
                CatalogTypeSuggestion(field: publisher, sortOrder: 0),
                CatalogTypeSuggestion(field: author, sortOrder: 1),
            ]]
        )
        await model.load()
        model.select("t1")
        await model.loadSuggestions(for: "t1")
        #expect(model.suggestions.map(\.field.label) == ["Publisher", "Author"])
    }

    @Test func assignPoolExcludesFieldsAlreadySuggested() async {
        let author = field(id: "f1", label: "Author")
        let model = makeModel(
            types: [seededType()],
            fields: [author, field(id: "f2", label: "Publisher")],
            suggestions: ["t1": [CatalogTypeSuggestion(field: author, sortOrder: 0)]]
        )
        await model.load()
        model.select("t1")
        await model.loadSuggestions(for: "t1")
        #expect(model.assignPool.map(\.id) == ["f2"])
    }

    // MARK: Associations

    @Test func assigningAppendsTheFieldAndUpdatesTheListCount() async {
        let model = makeModel(
            types: [seededType()],
            fields: [field(id: "f1", label: "Author"), field(id: "f2", label: "Publisher")]
        )
        await model.load()
        model.select("t1")
        await model.loadSuggestions(for: "t1")

        model.assignPick = "f2"
        await model.assignPickedField()
        model.assignPick = "f1"
        await model.assignPickedField()

        #expect(model.suggestions.map(\.field.id) == ["f2", "f1"])
        #expect(model.assignPick.isEmpty)
        #expect(model.types.first { $0.id == "t1" }?.suggestedFieldCount == 2)
        #expect(model.toast?.title == String(localized: L10n.SourceTypes.toastAssignedTitle))
    }

    @Test func removingASuggestionKeepsTheFieldInThePool() async {
        let author = field(id: "f1", label: "Author", usedBy: 3)
        let model = makeModel(
            types: [seededType()],
            fields: [author],
            suggestions: ["t1": [CatalogTypeSuggestion(field: author, sortOrder: 0)]]
        )
        await model.load()
        model.select("t1")
        await model.loadSuggestions(for: "t1")

        await model.removeSuggestion(fieldID: "f1")

        #expect(model.suggestions.isEmpty)
        #expect(model.types.first { $0.id == "t1" }?.suggestedFieldCount == 0)
        // The vocabulary row survives, so the field is assignable again.
        #expect(model.assignPool.map(\.id) == ["f1"])
        #expect(model.toast?.title == String(localized: L10n.SourceTypes.toastRemovedTitle))
        // T-20: the copy must not imply the existing values went with it.
        #expect(model.toast?.body.contains("3") == true)
    }

    @Test func removingASuggestionNoSourceUsesOmitsTheValueCount() async {
        let author = field(id: "f1", label: "Author")
        let model = makeModel(
            types: [seededType()],
            fields: [author],
            suggestions: ["t1": [CatalogTypeSuggestion(field: author, sortOrder: 0)]]
        )
        await model.load()
        model.select("t1")
        await model.loadSuggestions(for: "t1")

        await model.removeSuggestion(fieldID: "f1")
        #expect(model.toast?.body == L10n.SourceTypes.toastRemovedBody(field: "Author", type: "Book", valueCount: 0))
    }

    // MARK: Add

    @Test func addingMintsTheKeyFromTheLabelAndLandsOnTheNewType() async {
        let counts = CatalogCounts(projectDir: projectDir, store: FakeStore())
        let model = makeModel(types: [seededType()], catalogCounts: counts)
        await model.load()

        model.openAdd()
        #expect(model.isAdding)
        model.draft?.label = "Parish register"
        #expect(model.draftKey == "parish-register")

        await model.submit()

        #expect(!model.isAdding)
        #expect(model.selectedType?.key == "parish-register")
        #expect(model.selectedType?.origin == "user")
        #expect(model.selectedType?.iconKey == PVEvidenceIconKey.defaultTypeIcon.rawValue)
        #expect(model.suggestions.isEmpty)
        #expect(model.toast?.title == String(localized: L10n.SourceTypes.toastAddedTitle))
        #expect(counts.sourceTypes?.total == 2)
        #expect(counts.sourceTypes?.user == 1)
    }

    @Test func addingRefusesABlankLabelWithoutCallingTheStore() async {
        let model = makeModel()
        await model.load()
        model.openAdd()
        model.draft?.label = "   "

        #expect(!model.canSubmit)
        await model.submit()

        #expect(model.formError == String(localized: L10n.SourceTypes.errorLabelRequired))
        #expect(model.types.isEmpty)
    }

    @Test func addingRefusesALabelThatCannotBeSlugged() async {
        let model = makeModel()
        await model.load()
        model.openAdd()
        model.draft?.label = "—"

        await model.submit()

        #expect(model.formError == String(localized: L10n.SourceTypes.errorUnslugifiable))
        #expect(model.types.isEmpty)
    }

    @Test func addingADuplicateUserKeySurfacesTheErrorAndKeepsTheForm() async {
        let model = makeModel(types: [userType(id: "t2", label: "Family scrapbook")])
        await model.load()
        model.openAdd()
        model.draft?.label = "Family scrapbook"

        await model.submit()

        #expect(model.isAdding)
        #expect(model.formError == L10n.Errors.sourceTypesDuplicateKey(key: "family-scrapbook"))
        #expect(model.types.count == 1)
    }

    @Test func cancellingAddRestoresThePreviouslySelectedType() async {
        let model = makeModel(types: [seededType(), userType()])
        await model.load()
        model.select("t2")
        model.openAdd()
        model.cancelAdd()

        #expect(!model.isAdding)
        #expect(model.selectedType?.id == "t2")
    }

    // MARK: Edit

    @Test func editingASeededTypeIsAllowedAndTheKeyStays() async {
        let model = makeModel(types: [seededType()])
        await model.load()
        model.select("t1")
        #expect(!model.isDirty)

        model.draft?.label = "Printed book"
        #expect(model.isDirty)
        await model.submit()

        #expect(model.selectedType?.label == "Printed book")
        #expect(model.selectedType?.key == "book")
        #expect(model.toast?.title == String(localized: L10n.SourceTypes.toastUpdatedTitle))
    }

    @Test func changingIconMarksDirtyAndPersists() async {
        let model = makeModel(types: [userType()])
        await model.load()
        model.select("t2")
        #expect(model.draft?.iconKey == PVEvidenceIconKey.defaultTypeIcon.rawValue)

        model.draft?.iconKey = "type_scrapbook"
        #expect(model.isDirty)
        await model.submit()

        #expect(model.selectedType?.iconKey == "type_scrapbook")
        #expect(!model.isDirty)
    }

    @Test func revertingDiscardsUnsavedEdits() async {
        let model = makeModel(types: [userType()])
        await model.load()
        model.select("t2")
        model.draft?.description = "Changed"
        model.revertEdit()

        #expect(!model.isDirty)
        #expect(model.draft?.description == "Nan's albums.")
    }

    // MARK: Delete

    @Test func deleteIsGatedOnUseAndPluginOwnership() async {
        let model = makeModel(types: [
            userType(id: "t2", label: "Family scrapbook", usedBy: 0),
            seededType(id: "t1", label: "Book", usedBy: 26),
            pluginType(),
        ])
        await model.load()

        model.select("t2")
        #expect(model.canDeleteSelectedType)
        #expect(model.deleteTooltip == L10n.SourceTypes.deleteType)

        model.select("t1")
        #expect(!model.canDeleteSelectedType)
        #expect(model.deleteTooltip == L10n.SourceTypes.deleteInUse(count: 26))

        model.select("t3")
        #expect(!model.canDeleteSelectedType)
        #expect(model.deleteTooltip == L10n.SourceTypes.deleteOwnedByPlugin)
    }

    @Test func askDeleteIsIgnoredWhenTheTypeIsInUse() async {
        let model = makeModel(types: [seededType(usedBy: 26)])
        await model.load()
        model.select("t1")
        model.askDelete()

        #expect(model.pendingDeleteType == nil)
    }

    @Test func confirmingDeleteDropsTheTypeAndItsSuggestions() async {
        let author = field(id: "f1", label: "Author")
        let store = FakeStore()
        let counts = CatalogCounts(projectDir: projectDir, store: store)
        let model = makeModel(
            store: store,
            types: [userType()],
            fields: [author],
            suggestions: ["t2": [CatalogTypeSuggestion(field: author, sortOrder: 0)]],
            catalogCounts: counts
        )
        await model.load()
        model.select("t2")
        await model.loadSuggestions(for: "t2")
        model.askDelete()
        #expect(model.pendingDeleteType?.id == "t2")

        await model.confirmDelete()

        #expect(model.types.isEmpty)
        #expect(model.selectedType == nil)
        #expect(model.suggestions.isEmpty)
        #expect(model.pendingDeleteType == nil)
        #expect(model.toast?.title == String(localized: L10n.SourceTypes.toastDeletedTitle))
        // The suggestion join cascades; the field vocabulary row does not.
        #expect(store.fieldsByProject[projectDir]?.map(\.id) == ["f1"])
        #expect(counts.sourceTypes?.total == 0)
        #expect(counts.sourceFields?.total == 1)
    }
}
