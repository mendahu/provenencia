import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct SourceFieldsModelTests {
    private let projectDir = "/tmp/fields.provenencia"
    private let userID = "00000000-0000-7000-8000-000000000001"

    private func seededField(id: String = "1", label: String = "Author") -> CatalogMetadataField {
        CatalogMetadataField(id: id, key: "author", origin: "provenencia", label: label, dataType: "text", description: "Seeded.")
    }

    private func userField(id: String = "2", label: String = "Grandma's album code", usedBy: Int = 0) -> CatalogMetadataField {
        CatalogMetadataField(id: id, key: FieldSlug.kebab(label), origin: "user", label: label, dataType: "text", description: "Pencil code.", usedBy: usedBy)
    }

    private func pluginField(id: String = "3", label: String = "Memorial id") -> CatalogMetadataField {
        CatalogMetadataField(id: id, key: "memorial-id", origin: "plugin:findagrave", label: label, dataType: "text", description: "From the plugin.")
    }

    private func makeModel(
        store: FakeStore = FakeStore(),
        fields: [CatalogMetadataField] = [],
        catalogCounts: CatalogCounts? = nil
    ) -> SourceFieldsModel {
        store.fieldsByProject[projectDir] = fields
        return SourceFieldsModel(
            projectDir: projectDir,
            userID: userID,
            store: store,
            catalogCounts: catalogCounts
        )
    }

    @Test func loadPopulatesFieldsAndCounts() async {
        let model = makeModel(fields: [seededField(), userField()])
        await model.load()
        #expect(model.fields.count == 2)
        #expect(model.fields.seededCount == 1)
        #expect(model.fields.userCount == 1)
        #expect(model.fields.pluginCount == 0)
    }

    @Test func sortTogglesLabelDirection() async {
        let model = makeModel(fields: [
            seededField(id: "1", label: "Zebra"),
            userField(id: "2", label: "Album"),
        ])
        await model.load()
        #expect(model.visibleFields.map(\.id) == ["2", "1"])
        model.toggleLabelSort()
        #expect(model.visibleFields.map(\.id) == ["1", "2"])
    }

    @Test func selectedFieldLockedForPluginOriginOnly() async {
        let plugin = CatalogMetadataField(
            id: "3", key: "memorial-id", origin: "plugin:findagrave",
            label: "Memorial id", dataType: "text", description: ""
        )
        let model = makeModel(fields: [seededField(), userField(), plugin])
        await model.load()
        model.select("1")
        #expect(!model.isSelectedFieldLocked)
        model.select("2")
        #expect(!model.isSelectedFieldLocked)
        model.select("3")
        #expect(model.isSelectedFieldLocked)
    }

    @Test func submitEditUpdatesSeededFieldButKeepsKeyAndOrigin() async {
        let model = makeModel(fields: [seededField()])
        await model.load()
        model.select("1")
        #expect(!model.isSelectedFieldLocked)
        model.draft?.label = "Author renamed"
        model.draft?.description = "Updated starter."
        await model.submit()
        #expect(model.formError == nil)
        #expect(model.fields.first?.label == "Author renamed")
        #expect(model.fields.first?.key == "author")
        #expect(model.fields.first?.origin == "provenencia")
        #expect(model.toast?.title == String(localized: L10n.SourceFields.toastUpdatedTitle))
    }

    @Test func seededAndUserFieldsAreEditedTheSameWay() async {
        let model = makeModel(fields: [seededField(), userField()])
        await model.load()

        for id in ["1", "2"] {
            model.select(id)
            #expect(model.mode == .editing(id: id))
            #expect(!model.isSelectedFieldLocked)
            #expect(!model.isDirty)

            model.draft?.label = "Renamed \(id)"
            #expect(model.isDirty)
            #expect(model.canSubmit)
        }
    }

    @Test func pluginFieldsAreTheOnlyReadOnlyOnes() async {
        let model = makeModel(fields: [pluginField()])
        await model.load()
        model.select("3")

        #expect(model.mode == .viewing(id: "3"))
        #expect(model.isSelectedFieldLocked)
        #expect(!model.canSubmit)
    }

    @Test func openAddSeedsBlankDraftAndClearsSelection() async {
        let model = makeModel(fields: [userField()])
        await model.load()
        model.select("2")
        model.openAdd()
        #expect(model.selectedField == nil)
        #expect(model.draft == SourceFieldsModel.Draft(label: "", dataType: "text", description: ""))
    }

    @Test func cancelAddResumesPriorSelection() async {
        let model = makeModel(fields: [userField()])
        await model.load()
        model.select("2")
        model.openAdd()
        model.cancelAdd()
        #expect(model.selectedField?.id == "2")
    }

    @Test func submitAddCreatesFieldAndMintsKeyFromLabel() async {
        let counts = CatalogCounts(projectDir: projectDir, store: FakeStore())
        let model = makeModel(fields: [], catalogCounts: counts)
        await model.load()
        model.openAdd()
        model.draft?.label = "Grandma's album code"
        let location = await model.submit()
        #expect(model.formError == nil)
        #expect(model.fields.count == 1)
        #expect(model.fields.first?.key == "grandmas-album-code")
        #expect(model.fields.first?.origin == "user")
        #expect(model.toast?.title == String(localized: L10n.SourceFields.toastAddedTitle))
        #expect(model.selectedField?.key == "grandmas-album-code")
        #expect(counts.sourceFields?.total == 1)
        #expect(counts.sourceFields?.user == 1)
        #expect(location?.section == .sourceFields)
        #expect(location?.fieldId == model.fields.first?.id)
        #expect(location?.title == "Grandma's album code")
    }

    @Test func submitAddWithBlankLabelSetsFormError() async {
        let model = makeModel(fields: [])
        await model.load()
        model.openAdd()
        await model.submit()
        #expect(model.formError != nil)
        #expect(model.fields.isEmpty)
    }

    @Test func submitAddWithUnslugifiableLabelSetsFormError() async {
        let model = makeModel(fields: [])
        await model.load()
        model.openAdd()
        model.draft?.label = "..."
        await model.submit()
        #expect(model.formError != nil)
        #expect(model.fields.isEmpty)
    }

    @Test func submitAddWithCollidingSlugSetsFormError() async {
        // "Album code" and "Album  Code!" both slug to "album-code".
        let model = makeModel(fields: [userField(id: "2", label: "Album code")])
        await model.load()
        model.openAdd()
        model.draft?.label = "Album  Code!"
        await model.submit()
        #expect(model.formError == L10n.Errors.sourceFieldsDuplicateKey(key: "album-code"))
        #expect(model.fields.count == 1)
    }

    @Test func submitEditUpdatesFieldButKeepsKey() async {
        let counts = CatalogCounts(projectDir: projectDir, store: FakeStore())
        let model = makeModel(fields: [userField()], catalogCounts: counts)
        await model.load()
        let totalBefore = counts.sourceFields?.total
        model.select("2")
        model.draft?.label = "Grandma's photo album code"
        model.draft?.description = "Updated."
        await model.submit()
        #expect(model.formError == nil)
        #expect(model.fields.first?.label == "Grandma's photo album code")
        #expect(model.fields.first?.key == "grandmas-album-code")
        #expect(model.toast?.title == String(localized: L10n.SourceFields.toastUpdatedTitle))
        #expect(counts.sourceFields?.total == totalBefore)
    }

    @Test func isDirtyTracksUnsavedEditsAndRevertClearsThem() async {
        let model = makeModel(fields: [userField()])
        await model.load()
        model.select("2")
        #expect(!model.isDirty)
        model.draft?.label = "Changed"
        #expect(model.isDirty)
        model.revertEdit()
        #expect(!model.isDirty)
        #expect(model.draft?.label == "Grandma's album code")
    }

    @Test func canSubmitRequiresDirtyOnEditAndNonEmptyLabelOnAdd() async {
        let model = makeModel(fields: [userField()])
        await model.load()
        model.select("2")
        #expect(!model.canSubmit)
        model.draft?.label = "Changed"
        #expect(model.canSubmit)

        model.openAdd()
        #expect(!model.canSubmit)
        model.draft?.label = "New field"
        #expect(model.canSubmit)
    }
    // MARK: Delete

    @Test func deleteIsOfferedForSavedFieldsOnly() async {
        let model = makeModel(fields: [userField()])
        await model.load()
        #expect(!model.showsDelete)

        model.select("2")
        #expect(model.showsDelete)

        model.openAdd()
        #expect(!model.showsDelete)
    }

    @Test func unusedProjectOwnedFieldsCanBeDeleted() async {
        let model = makeModel(fields: [seededField(), userField()])
        await model.load()

        model.select("1")
        #expect(model.canDeleteSelectedField)
        model.select("2")
        #expect(model.canDeleteSelectedField)
    }

    @Test func pluginFieldsCannotBeDeletedAndSayWhy() async {
        let model = makeModel(fields: [pluginField()])
        await model.load()
        model.select("3")

        #expect(!model.canDeleteSelectedField)
        #expect(String(localized: model.deleteTooltip) == String(localized: L10n.SourceFields.deleteOwnedByPlugin))
    }

    @Test func fieldsInUseCannotBeDeletedAndCountTheSources() async {
        let model = makeModel(fields: [userField(usedBy: 3)])
        await model.load()
        model.select("2")

        #expect(!model.canDeleteSelectedField)
        #expect(String(localized: model.deleteTooltip) == String(localized: L10n.SourceFields.deleteInUse(count: 3)))
    }

    @Test func aSingleUseReadsAsOneSource() async {
        let model = makeModel(fields: [userField(usedBy: 1)])
        await model.load()
        model.select("2")

        #expect(String(localized: model.deleteTooltip) == String(localized: L10n.SourceFields.deleteInUse(count: 1)))
        #expect(String(localized: model.deleteTooltip) != String(localized: L10n.SourceFields.deleteInUse(count: 2)))
    }

    @Test func deletableFieldTooltipNamesTheAction() async {
        let model = makeModel(fields: [userField()])
        await model.load()
        model.select("2")

        #expect(String(localized: model.deleteTooltip) == String(localized: L10n.SourceFields.deleteField))
    }

    @Test func askDeleteOpensConfirmationForTheSelectedField() async {
        let model = makeModel(fields: [userField()])
        await model.load()
        model.select("2")
        model.askDelete()

        #expect(model.pendingDeleteField?.id == "2")
    }

    @Test func askDeleteIsIgnoredWhenTheFieldCannotBeDeleted() async {
        let model = makeModel(fields: [userField(usedBy: 2)])
        await model.load()
        model.select("2")
        model.askDelete()

        #expect(model.pendingDeleteField == nil)
    }

    @Test func cancelDeleteClosesTheConfirmationAndKeepsTheField() async {
        let model = makeModel(fields: [userField()])
        await model.load()
        model.select("2")
        model.askDelete()
        model.cancelDelete()

        #expect(model.pendingDeleteField == nil)
        #expect(model.fields.count == 1)
        #expect(model.selectedField?.id == "2")
    }

    @Test func confirmDeleteRemovesTheFieldClearsSelectionAndToasts() async {
        let store = FakeStore()
        let counts = CatalogCounts(projectDir: projectDir, store: store)
        let model = makeModel(store: store, fields: [seededField(), userField()], catalogCounts: counts)
        await model.load()
        #expect(counts.sourceFields?.total == 2)
        model.select("2")
        model.askDelete()
        let deleted = await model.confirmDelete()

        #expect(deleted)
        #expect(model.fields.map(\.id) == ["1"])
        #expect(store.fieldsByProject[projectDir]?.map(\.id) == ["1"])
        #expect(model.selectedField == nil)
        #expect(model.pendingDeleteField == nil)
        #expect(!model.isDeleting)
        #expect(model.toast?.title == String(localized: L10n.SourceFields.toastDeletedTitle))
        #expect(counts.sourceFields?.total == 1)
        #expect(counts.sourceFields?.seeded == 1)
        #expect(counts.sourceFields?.user == 0)
    }

    @Test func confirmDeleteKeepsTheConfirmationOpenWhenTheStoreRefuses() async {
        let store = FakeStore()
        let model = makeModel(store: store, fields: [userField()])
        await model.load()
        model.select("2")
        model.askDelete()
        // A source picks the field up after the model last listed, so the
        // count the button was enabled from is stale and the engine refuses.
        // That race is why the dialog has an error path at all.
        store.fieldsByProject[projectDir] = [userField(usedBy: 4)]
        let deleted = await model.confirmDelete()

        #expect(!deleted)
        #expect(model.fields.count == 1)
        #expect(model.pendingDeleteField?.id == "2")
        #expect(model.deleteError != nil)
    }

    @Test func loadFromLocationAppliesFieldInSameCompletion() async {
        let model = makeModel(fields: [seededField(id: "1", label: "Author")])
        let outcome = await model.load(
            from: WorkspaceLocation(section: .sourceFields, fieldId: "1", title: "Author")
        )
        #expect(outcome == .applied)
        #expect(model.selectedField?.id == "1")
        #expect(model.hasCompletedInitialLoad)
    }

    @Test func loadFromLocationMissingDeepId() async {
        let model = makeModel(fields: [seededField(id: "1")])
        let outcome = await model.load(
            from: WorkspaceLocation(section: .sourceFields, fieldId: "gone", title: "Missing")
        )
        #expect(outcome == .missingDeepId)
        #expect(model.selectedField == nil)
    }

    @Test func applyFromLocationSelectsAfterLoad() async {
        let model = makeModel(fields: [seededField(id: "1", label: "Author")])
        await model.load()
        let outcome = model.apply(
            from: WorkspaceLocation(section: .sourceFields, fieldId: "1", title: "Author")
        )
        #expect(outcome == .applied)
        #expect(model.selectedField?.id == "1")
    }

    @Test func applyFromLocationMissingDeepId() async {
        let model = makeModel(fields: [seededField(id: "1")])
        await model.load()
        let outcome = model.apply(
            from: WorkspaceLocation(section: .sourceFields, fieldId: "gone", title: "Missing")
        )
        #expect(outcome == .missingDeepId)
        #expect(model.selectedField == nil)
    }

    @Test func applyFromSectionRootClearsSelection() async {
        let model = makeModel(fields: [seededField(id: "1")])
        await model.load()
        model.select("1")
        let outcome = model.apply(from: .sectionRoot(.sourceFields))
        #expect(outcome == .applied)
        #expect(model.selectedField == nil)
    }

    @Test func applyIgnoredWhileAdding() async {
        let model = makeModel(fields: [seededField(id: "1")])
        await model.load()
        model.openAdd()
        let outcome = model.apply(
            from: WorkspaceLocation(section: .sourceFields, fieldId: "1", title: "Author")
        )
        #expect(outcome == .ignored)
        #expect(model.isAdding)
    }

    @Test func createAndDeleteCommitThroughWorkspaceNavigation() async throws {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("fields-nav-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        let url = dir.appendingPathComponent("history.json")
        let navigation = WorkspaceNavigation()
        navigation.attachProject(uuid: "00000000-0000-7000-8000-0000000000bb", fileURL: url)
        navigation.go(to: .sectionRoot(.sourceFields))

        let model = makeModel()
        await model.load()
        model.openAdd()
        model.draft?.label = "Album code"
        if let location = await model.submit() {
            navigation.go(to: location)
        }
        #expect(navigation.currentLocation.fieldId == model.fields.first?.id)
        #expect(navigation.canGoBack)

        model.askDelete()
        #expect(await model.confirmDelete())
        navigation.fallbackToSectionRoot()
        #expect(navigation.currentLocation == .sectionRoot(.sourceFields))
        #expect(navigation.currentLocation.fieldId == nil)
    }

}
