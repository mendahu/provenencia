import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct MetadataModelTests {
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
        fields: [CatalogMetadataField] = []
    ) -> (MetadataModel, WorkspaceSession) {
        store.fieldsByProject[projectDir] = fields
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = MetadataModel(
            session: session,
            userID: userID,
            store: store
        )
        return (model, session)
    }

    private func warm(_ model: MetadataModel, session: WorkspaceSession) async {
        model.warmFieldsQuery()
        await settle(session)
    }

    /// Waits out the fields list's reload.
    private func settle(_ session: WorkspaceSession) async {
        guard let fieldsHandle: QueryHandle<[CatalogMetadataField]> = session.queryHandle(
            MetadataModel.fieldsListKey(for: session)
        ) else { return }
        await waitForQuery(fieldsHandle)
        #expect(fieldsHandle.isSettled)
    }

    @Test func loadPopulatesFields() async {
        let (model, session) = makeModel(fields: [seededField(), userField()])
        await warm(model, session: session)
        #expect(model.fields.count == 2)
        #expect(model.fields.seededCount == 1)
        #expect(model.fields.userCount == 1)
        #expect(model.fields.pluginCount == 0)
    }

    @Test func sortTogglesLabelDirection() async {
        let (model, session) = makeModel(fields: [
            seededField(id: "1", label: "Zebra"),
            userField(id: "2", label: "Album"),
        ])
        await warm(model, session: session)
        #expect(model.visibleFields.map(\.id) == ["2", "1"])
        model.toggleLabelSort()
        #expect(model.visibleFields.map(\.id) == ["1", "2"])
    }

    @Test func selectedFieldLockedForPluginOriginOnly() async {
        let plugin = CatalogMetadataField(
            id: "3", key: "memorial-id", origin: "plugin:findagrave",
            label: "Memorial id", dataType: "text", description: ""
        )
        let (model, session) = makeModel(fields: [seededField(), userField(), plugin])
        await warm(model, session: session)
        model.select("1")
        #expect(!model.isSelectedFieldLocked)
        model.select("2")
        #expect(!model.isSelectedFieldLocked)
        model.select("3")
        #expect(model.isSelectedFieldLocked)
    }

    @Test func submitEditUpdatesSeededFieldButKeepsKeyAndOrigin() async {
        let (model, session) = makeModel(fields: [seededField()])
        await warm(model, session: session)
        model.select("1")
        #expect(!model.isSelectedFieldLocked)
        model.draft?.label = "Author renamed"
        model.draft?.description = "Updated starter."
        await model.submit()
        #expect(model.formError == nil)
        #expect(model.fields.first?.label == "Author renamed")
        #expect(model.fields.first?.key == "author")
        #expect(model.fields.first?.origin == "provenencia")
        #expect(model.toast?.title == L10n.string(L10n.Metadata.toastUpdatedTitle))
    }

    @Test func seededAndUserFieldsAreEditedTheSameWay() async {
        let (model, session) = makeModel(fields: [seededField(), userField()])
        await warm(model, session: session)

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
        let (model, session) = makeModel(fields: [pluginField()])
        await warm(model, session: session)
        model.select("3")

        #expect(model.mode == .viewing(id: "3"))
        #expect(model.isSelectedFieldLocked)
        #expect(!model.canSubmit)
    }

    @Test func openAddSeedsBlankDraftAndClearsSelection() async {
        let (model, session) = makeModel(fields: [userField()])
        await warm(model, session: session)
        model.select("2")
        model.openAdd()
        #expect(model.selectedField == nil)
        #expect(model.draft == MetadataModel.Draft(label: "", dataType: "text", description: ""))
    }

    @Test func cancelAddResumesPriorSelection() async {
        let (model, session) = makeModel(fields: [userField()])
        await warm(model, session: session)
        model.select("2")
        model.openAdd()
        model.cancelAdd()
        #expect(model.selectedField?.id == "2")
    }

    @Test func submitAddCreatesFieldAndMintsKeyFromLabel() async {
        let (model, session) = makeModel(fields: [])
        await warm(model, session: session)
        model.openAdd()
        model.draft?.label = "Grandma's album code"
        let location = await model.submit()
        #expect(model.formError == nil)
        #expect(model.fields.count == 1)
        #expect(model.fields.first?.key == "grandmas-album-code")
        #expect(model.fields.first?.origin == "user")
        #expect(model.toast?.title == L10n.string(L10n.Metadata.toastAddedTitle))
        #expect(model.selectedField?.key == "grandmas-album-code")
        #expect(location?.section == .metadata)
        #expect(location?.fieldId == model.fields.first?.id)
        #expect(location?.title == "Grandma's album code")
    }

    @Test func submitAddWithBlankLabelSetsFormError() async {
        let (model, session) = makeModel(fields: [])
        await warm(model, session: session)
        model.openAdd()
        await model.submit()
        #expect(model.formError != nil)
        #expect(model.fields.isEmpty)
    }

    @Test func submitAddWithUnslugifiableLabelSetsFormError() async {
        let (model, session) = makeModel(fields: [])
        await warm(model, session: session)
        model.openAdd()
        model.draft?.label = "..."
        await model.submit()
        #expect(model.formError != nil)
        #expect(model.fields.isEmpty)
    }

    @Test func submitAddWithCollidingSlugSetsFormError() async {
        // "Album code" and "Album  Code!" both slug to "album-code".
        let (model, session) = makeModel(fields: [userField(id: "2", label: "Album code")])
        await warm(model, session: session)
        model.openAdd()
        model.draft?.label = "Album  Code!"
        await model.submit()
        #expect(model.formError == L10n.Errors.metadataDuplicateKey(key: "album-code"))
        #expect(model.fields.count == 1)
    }

    @Test func submitEditUpdatesFieldButKeepsKey() async {
        let (model, session) = makeModel(fields: [userField()])
        await warm(model, session: session)
        model.select("2")
        model.draft?.label = "Grandma's photo album code"
        model.draft?.description = "Updated."
        await model.submit()
        #expect(model.formError == nil)
        #expect(model.fields.first?.label == "Grandma's photo album code")
        #expect(model.fields.first?.key == "grandmas-album-code")
        #expect(model.toast?.title == L10n.string(L10n.Metadata.toastUpdatedTitle))
    }

    /// T-8 (AC-1.2): a field relabel reports `source_vocabulary`; the Metadata
    /// list and every type's suggestions reload, and the visited Source page is
    /// marked stale through `.updatedMetadataField`.
    @Test func renamingAFieldReloadsItsDependents() async throws {
        let store = FakeStore()
        let (model, session) = makeModel(store: store, fields: [userField()])
        store.suggestionsByType["t1"] = [CatalogTypeSuggestion(field: userField(), sortOrder: 0)]
        let project = session.projectKey
        let suggestionsKey = CatalogQueryKey.typeSuggestions(project: project, typeId: "t1")
        let suggestions: QueryHandle<[CatalogTypeSuggestion]> = session.query(suggestionsKey)
        await waitForQuery(suggestions)
        let workspaceKey = CatalogQueryKey.sourceWorkspace(project: project, sourceId: "s1")
        session.setQueryValue(workspaceKey, value: CatalogSourceWorkspace(
            source: CatalogSource(id: "s1", ref: "SRC-1", sourceTypeID: "t1", title: "Alpha", description: ""),
            notes: [], metadata: [], artifacts: [], credibility: nil
        ))
        await warm(model, session: session)

        // The engine restates every suggestion that embeds the field.
        store.suggestionsByType["t1"] = [
            CatalogTypeSuggestion(field: userField(label: "Book"), sortOrder: 0),
        ]
        store.nextWriteEffects = CatalogWriteEffects(lists: [.sourceVocabulary])
        model.select("2")
        model.draft?.label = "Book"
        await model.submit()
        await settle(session)
        await session.settleWrites()

        #expect(model.fields.first?.label == "Book")
        let reloaded: QueryHandle<[CatalogTypeSuggestion]> = try #require(session.queryHandle(suggestionsKey))
        #expect(reloaded.value?.first?.field.label == "Book")
        let page: QueryHandle<CatalogSourceWorkspace> = try #require(session.queryHandle(workspaceKey))
        #expect(page.isStale)
    }

    /// T-12 (AC-4.2): the header counts the rows the page shows.
    @Test func countLineCountsTheRowsShown() async {
        let (model, session) = makeModel(fields: [
            seededField(id: "1", label: "A"), seededField(id: "2", label: "B"),
            seededField(id: "3", label: "C"), seededField(id: "4", label: "D"),
            pluginField(id: "5"),
        ])
        await warm(model, session: session)
        #expect(model.countLine == L10n.Metadata.countLineWithPlugin(total: 5, seeded: 4, user: 0, plugin: 1))
    }

    /// F-6: no count line until the list has a value.
    @Test func countLineIsAbsentUntilTheListLoads() async {
        let (model, session) = makeModel(fields: [seededField(id: "1", label: "A")])
        #expect(model.countLine == nil)
        await warm(model, session: session)
        #expect(model.countLine != nil)
    }

    @Test func isDirtyTracksUnsavedEditsAndRevertClearsThem() async {
        let (model, session) = makeModel(fields: [userField()])
        await warm(model, session: session)
        model.select("2")
        #expect(!model.isDirty)
        model.draft?.label = "Changed"
        #expect(model.isDirty)
        model.revertEdit()
        #expect(!model.isDirty)
        #expect(model.draft?.label == "Grandma's album code")
    }

    @Test func canSubmitRequiresDirtyOnEditAndNonEmptyLabelOnAdd() async {
        let (model, session) = makeModel(fields: [userField()])
        await warm(model, session: session)
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
        let (model, session) = makeModel(fields: [userField()])
        await warm(model, session: session)
        #expect(!model.showsDelete)

        model.select("2")
        #expect(model.showsDelete)

        model.openAdd()
        #expect(!model.showsDelete)
    }

    @Test func trashIsOfferedWhenUsedByOrPlugin() async {
        let (model, session) = makeModel(fields: [
            userField(),
            seededField(),
            pluginField(),
            userField(id: "4", label: "Notes", usedBy: 3),
        ])
        await warm(model, session: session)

        model.select("2")
        #expect(model.canDeleteSelectedField)
        #expect(model.deleteTooltip == L10n.Metadata.deleteField)
        #expect(model.deleteAccessibilityLabel == L10n.Metadata.deleteFieldAccessibility(label: "Grandma's album code"))

        model.select("1")
        #expect(model.canDeleteSelectedField)
        #expect(model.deleteTooltip == L10n.Metadata.deleteField)

        model.select("3")
        #expect(model.canDeleteSelectedField)
        #expect(model.deleteTooltip == L10n.Metadata.deleteField)

        model.select("4")
        #expect(model.canDeleteSelectedField)
        #expect(model.deleteTooltip == L10n.Metadata.deleteField)
    }

    @Test func unusedUserFieldConfirmDeletes() async {
        let store = FakeStore()
        let (model, session) = makeModel(store: store, fields: [userField()])
        await warm(model, session: session)
        model.select("2")
        await model.askDelete()

        #expect(model.pendingImpact?.report.allowed == true)
        #expect(model.pendingImpact?.target.kind == "metadata_field")
        #expect(await model.confirmPendingImpact())
        #expect(model.fields.isEmpty)
        #expect(model.mode == .empty)
        #expect(store.fieldsByProject[projectDir]?.isEmpty == true)
    }

    @Test func unusedSeededFieldConfirmDeletes() async {
        let store = FakeStore()
        let (model, session) = makeModel(store: store, fields: [seededField()])
        await warm(model, session: session)
        model.select("1")
        await model.askDelete()

        #expect(model.pendingImpact?.report.allowed == true)
        #expect(await model.confirmPendingImpact())
        #expect(model.fields.isEmpty)
        #expect(store.fieldsByProject[projectDir]?.isEmpty == true)
    }

    @Test func openURLUsesOnlyOpenHref() {
        var entry = CatalogMetadataEntry(
            field: seededField(), valueText: "www.example.org", hasValue: true, suggested: false, sortOrder: 0
        )
        #expect(entry.openURL == nil)

        entry.openHref = "https://example.org"
        #expect(entry.openURL == URL(string: "https://example.org"))
    }

    @Test func fieldUsedByTwoSourcesShowsInboundNotice() async {
        let store = FakeStore()
        let field = userField(usedBy: 2)
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-AAAAA", sourceTypeID: "t1", title: "Parish", description: ""),
            CatalogSource(id: "s2", ref: "SRC-BBBBB", sourceTypeID: "t1", title: "Census", description: ""),
        ]
        store.metadataByOwner[MetadataOwner(level: .source, id: "s1")] = [
            CatalogMetadataEntry(field: field, valueText: "a", hasValue: true, suggested: false, sortOrder: 0),
        ]
        store.metadataByOwner[MetadataOwner(level: .source, id: "s2")] = [
            CatalogMetadataEntry(field: field, valueText: "b", hasValue: true, suggested: false, sortOrder: 0),
        ]
        let (model, session) = makeModel(store: store, fields: [field])
        await warm(model, session: session)
        model.select("2")
        #expect(model.canDeleteSelectedField)
        await model.askDelete()

        let report = model.pendingImpact?.report
        #expect(report?.allowed == false)
        #expect(report?.gate == .inbound)
        #expect(report?.groups.first?.via == "source_metadata.field_id")
        #expect(report?.groups.first?.total == 2)
        #expect(report?.groups.first?.total == model.selectedField?.usedBy)
        #expect(report?.groups.first?.listed.map(\.ref) == ["SRC-AAAAA", "SRC-BBBBB"])
        #expect(await model.confirmPendingImpact() == false)
        #expect(store.fieldsByProject[projectDir]?.map(\.id) == ["2"])
        #expect(model.fields.map(\.id) == ["2"])
    }

    @Test func pluginFieldOriginLockedTrashStillOffered() async {
        let store = FakeStore()
        let (model, session) = makeModel(store: store, fields: [pluginField()])
        await warm(model, session: session)
        model.select("3")
        #expect(model.canDeleteSelectedField)
        await model.askDelete()

        #expect(model.pendingImpact?.report.allowed == false)
        #expect(model.pendingImpact?.report.gate == .originLocked)
        #expect(await model.confirmPendingImpact() == false)
        #expect(store.fieldsByProject[projectDir]?.map(\.id) == ["3"])
    }

    @Test func cancelDeleteClosesTheConfirmationAndKeepsTheField() async {
        let (model, session) = makeModel(fields: [userField()])
        await warm(model, session: session)
        model.select("2")
        await model.askDelete()
        model.cancelDelete()

        #expect(model.pendingDeleteField == nil)
        #expect(model.pendingImpact == nil)
        #expect(model.fields.count == 1)
        #expect(model.selectedField?.id == "2")
    }

    @Test func confirmPendingImpactRemovesTheFieldClearsSelectionAndToasts() async {
        let store = FakeStore()
        let (model, session) = makeModel(store: store, fields: [seededField(), userField()])
        await warm(model, session: session)
        model.select("2")
        await model.askDelete()
        #expect(model.pendingImpact?.report.allowed == true)
        let deleted = await model.confirmPendingImpact()

        #expect(deleted)
        #expect(model.fields.map(\.id) == ["1"])
        #expect(store.fieldsByProject[projectDir]?.map(\.id) == ["1"])
        #expect(model.selectedField == nil)
        #expect(model.pendingDeleteField == nil)
        #expect(!model.isDeleting)
        #expect(model.toast?.title == L10n.string(L10n.Metadata.toastDeletedTitle))
        await settle(session)
    }

    @Test func confirmPendingImpactKeepsTheSheetOpenWhenTheStoreRefuses() async {
        let store = FakeStore()
        let (model, session) = makeModel(store: store, fields: [userField()])
        await warm(model, session: session)
        model.select("2")
        await model.askDelete()
        store.fieldsByProject[projectDir] = [userField(usedBy: 4)]
        let deleted = await model.confirmPendingImpact()

        #expect(!deleted)
        #expect(model.fields.count == 1)
        #expect(model.pendingDeleteField?.id == "2")
        #expect(model.pendingImpact?.report.allowed == false)
        #expect(model.pendingImpact?.report.gate == .inbound)
        #expect(model.deleteError == nil)
    }

    @Test func syncSelectionAppliesFieldAfterWarm() async {
        let (model, session) = makeModel(fields: [seededField(id: "1", label: "Author")])
        await warm(model, session: session)
        let outcome = model.syncSelection(
            from: WorkspaceLocation(section: .metadata, fieldId: "1", title: "Author")
        )
        #expect(outcome == .applied)
        #expect(model.selectedField?.id == "1")
    }

    @Test func syncSelectionMissingDeepId() async {
        let (model, session) = makeModel(fields: [seededField(id: "1")])
        await warm(model, session: session)
        let outcome = model.syncSelection(
            from: WorkspaceLocation(section: .metadata, fieldId: "gone", title: "Missing")
        )
        #expect(outcome == .missingDeepId)
        #expect(model.selectedField == nil)
    }

    @Test func syncSelectionFromSectionRootClearsSelection() async {
        let (model, session) = makeModel(fields: [seededField(id: "1")])
        await warm(model, session: session)
        model.select("1")
        let outcome = model.syncSelection(from: .sectionRoot(.metadata))
        #expect(outcome == .applied)
        #expect(model.selectedField == nil)
    }

    @Test func syncSelectionIgnoredWhileAdding() async {
        let (model, session) = makeModel(fields: [seededField(id: "1")])
        await warm(model, session: session)
        model.openAdd()
        let outcome = model.syncSelection(
            from: WorkspaceLocation(section: .metadata, fieldId: "1", title: "Author")
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
        navigation.go(to: .sectionRoot(.metadata))

        let (model, session) = makeModel()
        await warm(model, session: session)
        model.openAdd()
        model.draft?.label = "Album code"
        if let location = await model.submit() {
            navigation.go(to: location)
        }
        #expect(navigation.currentLocation.fieldId == model.fields.first?.id)
        #expect(navigation.canGoBack)

        await model.askDelete()
        #expect(await model.confirmPendingImpact())
        navigation.fallbackToSectionRoot()
        #expect(navigation.currentLocation == .sectionRoot(.metadata))
        #expect(navigation.currentLocation.fieldId == nil)
    }

}
