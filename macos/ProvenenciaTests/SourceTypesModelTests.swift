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

    private func pluginType(id: String = "t3", label: String = "Grave memorial", usedBy: Int = 5) -> CatalogSourceType {
        CatalogSourceType(
            id: id, key: "grave-memorial", origin: "plugin:findagrave", label: label,
            description: "From the plugin.", usedBy: usedBy
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
        suggestions: [String: [CatalogTypeSuggestion]] = [:]
    ) -> (SourceTypesModel, WorkspaceSession) {
        store.sourceTypesByProject[projectDir] = types
        store.fieldsByProject[projectDir] = fields
        store.suggestionsByType = suggestions
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = SourceTypesModel(
            session: session,
            userID: userID,
            store: store
        )
        return (model, session)
    }

    private func warm(_ model: SourceTypesModel, session: WorkspaceSession) async {
        model.warmListQueries()
        if let typesHandle: QueryHandle<[CatalogSourceType]> = session.queryHandle(
            SourceTypesModel.typesListKey(for: session)
        ) {
            await waitForQuery(typesHandle)
        }
        if let fieldsHandle: QueryHandle<[CatalogMetadataField]> = session.queryHandle(
            SourceTypesModel.fieldsListKey(for: session)
        ) {
            await waitForQuery(fieldsHandle)
        }
    }

    private func warmSuggestions(_ model: SourceTypesModel, session: WorkspaceSession, typeID: String) async {
        model.warmSuggestions(for: typeID)
        if let handle: QueryHandle<[CatalogTypeSuggestion]> = session.queryHandle(
            SourceTypesModel.suggestionsKey(for: session, typeID: typeID)
        ) {
            await waitForQuery(handle)
        }
    }

    // MARK: List

    @Test func loadPopulatesTypesPool() async {
        let (model, session) = makeModel(
            types: [seededType(), userType(), pluginType()],
            fields: [field(id: "f1", label: "Author")]
        )
        await warm(model, session: session)
        #expect(model.types.count == 3)
        #expect(model.fields.count == 1)
        #expect(model.types.seededCount == 1)
        #expect(model.types.userCount == 1)
        #expect(model.types.pluginCount == 1)
    }

    @Test func sortingTogglesWithinAColumnAndResetsAcrossColumns() async {
        let store = FakeStore()
        let (model, session) = makeModel(
            store: store,
            types: [seededType(), userType()],
            fields: [field(id: "f1", label: "Author")],
            suggestions: ["t2": [CatalogTypeSuggestion(field: field(id: "f1", label: "Author"), sortOrder: 0)]]
        )
        await warm(model, session: session)
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
        let (model, session) = makeModel(types: [seededType(), userType(), pluginType()])
        await warm(model, session: session)

        model.select("t1")
        #expect(!model.isSelectedTypeLocked)
        #expect(model.canEditAssociations)

        model.select("t3")
        #expect(model.isSelectedTypeLocked)
        #expect(!model.canEditAssociations)
        #expect(model.canDeleteSelectedType)
    }

    @Test func selectingLoadsThatTypesSuggestionsInStoredOrder() async {
        let author = field(id: "f1", label: "Author")
        let publisher = field(id: "f2", label: "Publisher")
        let (model, session) = makeModel(
            types: [seededType()],
            fields: [author, publisher],
            // Publisher was assigned first, so stored order and label order disagree.
            suggestions: ["t1": [
                CatalogTypeSuggestion(field: publisher, sortOrder: 0),
                CatalogTypeSuggestion(field: author, sortOrder: 1),
            ]]
        )
        await warm(model, session: session)
        model.select("t1")
        await warmSuggestions(model, session: session, typeID: "t1")
        #expect(model.suggestions.map(\.field.label) == ["Publisher", "Author"])
    }

    @Test func assignPoolExcludesFieldsAlreadySuggested() async {
        let author = field(id: "f1", label: "Author")
        let (model, session) = makeModel(
            types: [seededType()],
            fields: [author, field(id: "f2", label: "Publisher")],
            suggestions: ["t1": [CatalogTypeSuggestion(field: author, sortOrder: 0)]]
        )
        await warm(model, session: session)
        model.select("t1")
        await warmSuggestions(model, session: session, typeID: "t1")
        #expect(model.assignPool.map(\.id) == ["f2"])
    }

    // MARK: Associations

    @Test func assigningAppendsTheFieldAndUpdatesTheListCount() async {
        let (model, session) = makeModel(
            types: [seededType()],
            fields: [field(id: "f1", label: "Author"), field(id: "f2", label: "Publisher")]
        )
        await warm(model, session: session)
        model.select("t1")
        await warmSuggestions(model, session: session, typeID: "t1")

        model.assignPick = "f2"
        await model.assignPickedField()
        model.assignPick = "f1"
        await model.assignPickedField()

        #expect(model.suggestions.map(\.field.id) == ["f2", "f1"])
        #expect(model.assignPick.isEmpty)
        #expect(model.types.first { $0.id == "t1" }?.suggestedFieldCount == 2)
        #expect(model.toast?.title == L10n.string(L10n.SourceTypes.toastAssignedTitle))
    }

    @Test func selectingShowsOnlySourceLevelSuggestions() async {
        let author = field(id: "f1", label: "Author")
        let publisher = field(id: "f2", label: "Publisher")
        let store = FakeStore()
        let (model, session) = makeModel(
            store: store,
            types: [seededType()],
            fields: [author, publisher],
            suggestions: ["t1": [CatalogTypeSuggestion(field: author, sortOrder: 0)]]
        )
        store.suggestionsByLevel[.graph] = ["t1": [CatalogTypeSuggestion(field: publisher, sortOrder: 0)]]
        await warm(model, session: session)
        model.select("t1")
        await warmSuggestions(model, session: session, typeID: "t1")
        #expect(model.suggestions.map(\.field.id) == ["f1"])
    }

    @Test func assigningWritesAtSourceLevelOnly() async {
        let author = field(id: "f1", label: "Author")
        let publisher = field(id: "f2", label: "Publisher")
        let store = FakeStore()
        let (model, session) = makeModel(
            store: store,
            types: [seededType()],
            fields: [author, publisher]
        )
        let graphRow = CatalogTypeSuggestion(field: publisher, sortOrder: 0)
        store.suggestionsByLevel[.graph] = ["t1": [graphRow]]
        await warm(model, session: session)
        model.select("t1")
        await warmSuggestions(model, session: session, typeID: "t1")

        model.assignPick = "f1"
        await model.assignPickedField()

        #expect(store.suggestionsByLevel[.source]?["t1"]?.map(\.field.id) == ["f1"])
        #expect(store.suggestionsByLevel[.graph]?["t1"] == [graphRow])
        #expect(model.types.first { $0.id == "t1" }?.suggestedFieldCount == 1)
    }

    @Test func fieldsColumnShowsTheEnginesCountNotTheReplyLength() async {
        let author = field(id: "f1", label: "Author")
        let store = FakeStore()
        store.suggestedFieldCountOverride["t1"] = 7
        let (model, session) = makeModel(
            store: store,
            types: [seededType()],
            fields: [author]
        )
        await warm(model, session: session)
        model.select("t1")
        await warmSuggestions(model, session: session, typeID: "t1")

        model.assignPick = "f1"
        await model.assignPickedField()

        #expect(model.suggestions.count == 1)
        #expect(model.types.first { $0.id == "t1" }?.suggestedFieldCount == 7)
    }

    @Test func removingASuggestionKeepsTheFieldInThePool() async {
        let author = field(id: "f1", label: "Author", usedBy: 3)
        let (model, session) = makeModel(
            types: [seededType()],
            fields: [author],
            suggestions: ["t1": [CatalogTypeSuggestion(field: author, sortOrder: 0)]]
        )
        await warm(model, session: session)
        model.select("t1")
        await warmSuggestions(model, session: session, typeID: "t1")

        await model.removeSuggestion(fieldID: "f1")

        #expect(model.suggestions.isEmpty)
        #expect(model.types.first { $0.id == "t1" }?.suggestedFieldCount == 0)
        // The vocabulary row survives, so the field is assignable again.
        #expect(model.assignPool.map(\.id) == ["f1"])
        #expect(model.toast?.title == L10n.string(L10n.SourceTypes.toastRemovedTitle))
        // T-20: the copy must not imply the existing values went with it.
        #expect(model.toast?.body.contains("3") == true)
    }

    @Test func removingASuggestionNoSourceUsesOmitsTheValueCount() async {
        let author = field(id: "f1", label: "Author")
        let (model, session) = makeModel(
            types: [seededType()],
            fields: [author],
            suggestions: ["t1": [CatalogTypeSuggestion(field: author, sortOrder: 0)]]
        )
        await warm(model, session: session)
        model.select("t1")
        await warmSuggestions(model, session: session, typeID: "t1")

        await model.removeSuggestion(fieldID: "f1")
        #expect(model.toast?.body == L10n.SourceTypes.toastRemovedBody(field: "Author", type: "Book", valueCount: 0))
    }

    /// T-12 (AC-4.1): the header counts the rows the page shows.
    @Test func countLineCountsTheRowsShown() async {
        let (model, session) = makeModel(types: [
            seededType(id: "t1", label: "A"), seededType(id: "t3", label: "B"), seededType(id: "t4", label: "C"),
            userType(id: "t2", label: "D"), userType(id: "t5", label: "E"),
        ])
        await warm(model, session: session)
        #expect(model.countLine == L10n.SourceTypes.countLine(total: 5, seeded: 3, user: 2))
    }

    /// F-6: no count line until the list has a value.
    @Test func countLineIsAbsentUntilTheListLoads() async {
        let (model, session) = makeModel(types: [seededType(id: "t1", label: "A")])
        #expect(model.countLine == nil)
        await warm(model, session: session)
        #expect(model.countLine != nil)
    }

    // MARK: Add

    @Test func addingMintsTheKeyFromTheLabelAndLandsOnTheNewType() async {
        let (model, session) = makeModel(types: [seededType()])
        await warm(model, session: session)

        model.openAdd()
        #expect(model.isAdding)
        model.draft?.label = "Parish register"
        #expect(model.draftKey == "parish-register")

        await model.submit()

        #expect(!model.isAdding)
        #expect(model.selectedType?.key == "parish-register")
        #expect(model.selectedType?.origin == "user")
        #expect(model.selectedType?.iconKey == PVMarkKey.defaultTypeMark.rawValue)
        #expect(model.suggestions.isEmpty)
        #expect(model.toast?.title == L10n.string(L10n.SourceTypes.toastAddedTitle))
    }

    /// T-7 (AC-1.3): a type create reloads the Source vocabulary lists from its
    /// effects and leaves a visited Source page alone; an update that changed
    /// nothing reports no effects and reloads nothing. The store holds a row the
    /// cache has not seen, so a reload shows as that row arriving.
    @Test func addingATypeReloadsFromItsEffects() async throws {
        let store = FakeStore()
        let (model, session) = makeModel(store: store, types: [seededType()])
        await warm(model, session: session)
        let project = session.projectKey
        let pageKey = CatalogQueryKey.sourceWorkspace(project: project, sourceId: "s1")
        session.setQueryValue(pageKey, value: CatalogSourceWorkspace(
            source: CatalogSource(id: "s1", ref: "SRC-1", sourceTypeID: "t1", title: "Alpha", description: ""),
            notes: [], metadata: [], artifacts: [], credibility: nil
        ))
        let fieldsHandle: QueryHandle<[CatalogMetadataField]> = session.query(
            CatalogQueryKey.metadataFieldsList(project: project)
        )
        await waitForQuery(fieldsHandle)
        let typesHandle: QueryHandle<[CatalogSourceType]> = try #require(
            session.queryHandle(SourceTypesModel.typesListKey(for: session))
        )
        store.fieldsByProject[projectDir] = [field(id: "f9", label: "Unseen")]

        store.nextWriteEffects = CatalogWriteEffects(lists: [.sourceVocabulary])
        model.openAdd()
        model.draft?.label = "Diary"
        await model.submit()
        await session.settleWrites()

        #expect(model.selectedType?.key == "diary")
        #expect(typesHandle.value?.contains { $0.key == "diary" } == true)
        #expect(fieldsHandle.value?.map(\.id) == ["f9"])
        let page: QueryHandle<CatalogSourceWorkspace> = try #require(session.queryHandle(pageKey))
        #expect(!page.isStale)

        store.fieldsByProject[projectDir] = [field(id: "f10", label: "Later")]
        store.nextWriteEffects = .empty
        await model.submit()
        await session.settleWrites()
        #expect(fieldsHandle.value?.map(\.id) == ["f9"])
    }

    @Test func addingReturnsLocationForWorkspaceNavigation() async {
        let (model, session) = makeModel(types: [seededType()])
        await warm(model, session: session)
        model.openAdd()
        model.draft?.label = "Parish register"
        let location = await model.submit()
        #expect(location?.section == .sourceTypes)
        #expect(location?.typeId == model.selectedType?.id)
        #expect(location?.title == "Parish register")
    }

    @Test func addingRefusesABlankLabelWithoutCallingTheStore() async {
        let (model, session) = makeModel()
        await warm(model, session: session)
        model.openAdd()
        model.draft?.label = "   "

        #expect(!model.canSubmit)
        await model.submit()

        #expect(model.formError == L10n.string(L10n.SourceTypes.errorLabelRequired))
        #expect(model.types.isEmpty)
    }

    @Test func addingRefusesALabelThatCannotBeSlugged() async {
        let (model, session) = makeModel()
        await warm(model, session: session)
        model.openAdd()
        model.draft?.label = "—"

        await model.submit()

        #expect(model.formError == L10n.string(L10n.SourceTypes.errorUnslugifiable))
        #expect(model.types.isEmpty)
    }

    @Test func addingADuplicateUserKeySurfacesTheErrorAndKeepsTheForm() async {
        let (model, session) = makeModel(types: [userType(id: "t2", label: "Family scrapbook")])
        await warm(model, session: session)
        model.openAdd()
        model.draft?.label = "Family scrapbook"

        await model.submit()

        #expect(model.isAdding)
        #expect(model.formError == L10n.Errors.sourceTypesDuplicateKey(key: "family-scrapbook"))
        #expect(model.types.count == 1)
    }

    @Test func cancellingAddRestoresThePreviouslySelectedType() async {
        let (model, session) = makeModel(types: [seededType(), userType()])
        await warm(model, session: session)
        model.select("t2")
        model.openAdd()
        model.cancelAdd()

        #expect(!model.isAdding)
        #expect(model.selectedType?.id == "t2")
    }

    // MARK: Edit

    @Test func editingASeededTypeIsAllowedAndTheKeyStays() async {
        let (model, session) = makeModel(types: [seededType()])
        await warm(model, session: session)
        model.select("t1")
        #expect(!model.isDirty)

        model.draft?.label = "Printed book"
        #expect(model.isDirty)
        await model.submit()

        #expect(model.selectedType?.label == "Printed book")
        #expect(model.selectedType?.key == "book")
        #expect(model.toast?.title == L10n.string(L10n.SourceTypes.toastUpdatedTitle))
    }

    @Test func changingIconMarksDirtyAndPersists() async {
        let (model, session) = makeModel(types: [userType()])
        await warm(model, session: session)
        model.select("t2")
        #expect(model.draft?.iconKey == PVMarkKey.defaultTypeMark.rawValue)

        model.draft?.iconKey = "type_scrapbook"
        #expect(model.isDirty)
        await model.submit()

        #expect(model.selectedType?.iconKey == "type_scrapbook")
        #expect(!model.isDirty)
    }

    @Test func applyDraftIconKeyUpdatesPendingPickerSelection() async {
        let (model, session) = makeModel(types: [userType()])
        await warm(model, session: session)
        model.select("t2")
        let before = model.draft?.iconKey
        #expect(before == PVMarkKey.defaultTypeMark.rawValue)

        model.applyDraftIconKey("type_photograph")
        #expect(model.draft?.iconKey == "type_photograph")
        #expect(model.isDirty)

        // Cancel path: caller simply discards pending UI state; draft unchanged
        // until applyDraftIconKey — reverting restores the saved icon.
        model.revertEdit()
        #expect(model.draft?.iconKey == before)
        #expect(!model.isDirty)
    }

    @Test func revertingDiscardsUnsavedEdits() async {
        let (model, session) = makeModel(types: [userType()])
        await warm(model, session: session)
        model.select("t2")
        model.draft?.description = "Changed"
        model.revertEdit()

        #expect(!model.isDirty)
        #expect(model.draft?.description == "Nan's albums.")
    }

    // MARK: Delete

    @Test func trashIsOfferedWhenUsedByOrPlugin() async {
        let (model, session) = makeModel(types: [
            userType(id: "t2", label: "Family scrapbook", usedBy: 0),
            seededType(id: "t1", label: "Book", usedBy: 26),
            pluginType(),
        ])
        await warm(model, session: session)

        model.select("t2")
        #expect(model.canDeleteSelectedType)
        #expect(model.deleteTooltip == L10n.SourceTypes.deleteType)
        #expect(model.deleteAccessibilityLabel == L10n.SourceTypes.deleteTypeAccessibility(label: "Family scrapbook"))

        model.select("t1")
        #expect(model.canDeleteSelectedType)
        #expect(model.deleteTooltip == L10n.SourceTypes.deleteType)

        model.select("t3")
        #expect(model.canDeleteSelectedType)
        #expect(model.deleteTooltip == L10n.SourceTypes.deleteType)
    }

    @Test func unusedUserTypeConfirmDeletes() async {
        let store = FakeStore()
        let (model, session) = makeModel(store: store, types: [userType()])
        await warm(model, session: session)
        model.select("t2")
        await model.askDelete()

        #expect(model.pendingImpact?.report.allowed == true)
        #expect(model.pendingImpact?.target.kind == "source_type")
        #expect(await model.confirmPendingImpact())
        #expect(model.types.isEmpty)
        #expect(model.mode == .empty)
        #expect(store.sourceTypesByProject[projectDir]?.isEmpty == true)
    }

    @Test func unusedSeededTypeConfirmDeletes() async {
        let store = FakeStore()
        let (model, session) = makeModel(store: store, types: [seededType()])
        await warm(model, session: session)
        model.select("t1")
        await model.askDelete()

        #expect(model.pendingImpact?.report.allowed == true)
        #expect(await model.confirmPendingImpact())
        #expect(model.types.isEmpty)
        #expect(store.sourceTypesByProject[projectDir]?.isEmpty == true)
    }

    @Test func typeUsedByTwoSourcesShowsInboundNotice() async {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-AAAAA", sourceTypeID: "t1", title: "Parish", description: ""),
            CatalogSource(id: "s2", ref: "SRC-BBBBB", sourceTypeID: "t1", title: "Census", description: ""),
        ]
        let (model, session) = makeModel(store: store, types: [seededType(usedBy: 2)])
        await warm(model, session: session)
        model.select("t1")
        #expect(model.canDeleteSelectedType)
        await model.askDelete()

        let report = model.pendingImpact?.report
        #expect(report?.allowed == false)
        #expect(report?.gate == .inbound)
        #expect(report?.groups.first?.via == "sources.source_type_id")
        #expect(report?.groups.first?.total == 2)
        #expect(report?.groups.first?.total == model.selectedType?.usedBy)
        #expect(report?.groups.first?.listed.map(\.ref) == ["SRC-AAAAA", "SRC-BBBBB"])
        #expect(await model.confirmPendingImpact() == false)
        #expect(store.sourceTypesByProject[projectDir]?.map(\.id) == ["t1"])
        #expect(model.types.map(\.id) == ["t1"])
    }

    @Test func pluginTypeOriginLockedTrashStillOffered() async {
        let store = FakeStore()
        let (model, session) = makeModel(store: store, types: [pluginType(usedBy: 0)])
        await warm(model, session: session)
        model.select("t3")
        #expect(model.canDeleteSelectedType)
        await model.askDelete()

        #expect(model.pendingImpact?.report.allowed == false)
        #expect(model.pendingImpact?.report.gate == .originLocked)
        #expect(await model.confirmPendingImpact() == false)
        #expect(store.sourceTypesByProject[projectDir]?.map(\.id) == ["t3"])
    }

    @Test func confirmingDeleteDropsTheTypeAndItsSuggestions() async {
        let author = field(id: "f1", label: "Author")
        let store = FakeStore()
        let (model, session) = makeModel(
            store: store,
            types: [userType()],
            fields: [author],
            suggestions: ["t2": [CatalogTypeSuggestion(field: author, sortOrder: 0)]]
        )
        await warm(model, session: session)
        model.select("t2")
        await warmSuggestions(model, session: session, typeID: "t2")
        await model.askDelete()
        #expect(model.pendingDeleteType?.id == "t2")
        #expect(model.pendingImpact?.report.allowed == true)

        await model.confirmPendingImpact()

        #expect(model.types.isEmpty)
        #expect(model.selectedType == nil)
        #expect(model.suggestions.isEmpty)
        #expect(model.pendingDeleteType == nil)
        #expect(model.toast?.title == L10n.string(L10n.SourceTypes.toastDeletedTitle))
        // The suggestion join cascades; the field vocabulary row does not.
        #expect(store.fieldsByProject[projectDir]?.map(\.id) == ["f1"])
    }

    @Test func syncSelectionAppliesTypeAfterWarm() async {
        let (model, session) = makeModel(types: [
            CatalogSourceType(id: "t1", key: "book", origin: "provenencia", label: "Book", description: ""),
        ])
        await warm(model, session: session)
        let outcome = model.syncSelection(
            from: WorkspaceLocation(section: .sourceTypes, typeId: "t1", title: "Book")
        )
        #expect(outcome == .applied)
        #expect(model.selectedType?.id == "t1")
    }

    @Test func syncSelectionMissingDeepId() async {
        let (model, session) = makeModel(types: [
            CatalogSourceType(id: "t1", key: "book", origin: "provenencia", label: "Book", description: ""),
        ])
        await warm(model, session: session)
        let outcome = model.syncSelection(
            from: WorkspaceLocation(section: .sourceTypes, typeId: "gone", title: "Missing")
        )
        #expect(outcome == .missingDeepId)
        #expect(model.selectedType == nil)
    }

    @Test func syncSelectionFromSectionRootClearsSelection() async {
        let (model, session) = makeModel(types: [
            CatalogSourceType(id: "t1", key: "book", origin: "provenencia", label: "Book", description: ""),
        ])
        await warm(model, session: session)
        model.select("t1")
        await warmSuggestions(model, session: session, typeID: "t1")
        let outcome = model.syncSelection(from: .sectionRoot(.sourceTypes))
        #expect(outcome == .applied)
        #expect(model.selectedType == nil)
    }

    @Test func syncSelectionIgnoredWhileAdding() async {
        let (model, session) = makeModel(types: [
            CatalogSourceType(id: "t1", key: "book", origin: "provenencia", label: "Book", description: ""),
        ])
        await warm(model, session: session)
        model.openAdd()
        let outcome = model.syncSelection(
            from: WorkspaceLocation(section: .sourceTypes, typeId: "t1", title: "Book")
        )
        #expect(outcome == .ignored)
        #expect(model.isAdding)
    }

    @Test func createAndDeleteCommitThroughWorkspaceNavigation() async throws {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("types-nav-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        let url = dir.appendingPathComponent("history.json")
        let navigation = WorkspaceNavigation()
        navigation.attachProject(uuid: "00000000-0000-7000-8000-0000000000cc", fileURL: url)
        navigation.go(to: .sectionRoot(.sourceTypes))

        let (model, session) = makeModel()
        await warm(model, session: session)
        model.openAdd()
        model.draft?.label = "Parish register"
        if let location = await model.submit() {
            navigation.go(to: location)
        }
        #expect(navigation.currentLocation.typeId == model.selectedType?.id)

        await model.askDelete()
        #expect(await model.confirmPendingImpact())
        navigation.fallbackToSectionRoot()
        #expect(navigation.currentLocation == .sectionRoot(.sourceTypes))
        #expect(navigation.currentLocation.typeId == nil)
    }
}
