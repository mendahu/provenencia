import Foundation
import Observation

/// State for the **Source types** workspace destination (S2-03 board /
/// S2-16 PR): browse/search/sort the project's `source_types` vocabulary,
/// create or edit project rows, delete unused ones, and assign or remove
/// `source_type_metadata_fields` suggestions each type carries.
@MainActor
@Observable
final class SourceTypesModel {
    struct Draft: Equatable {
        var label: String
        var description: String
        var iconKey: String = PVEvidenceIconKey.defaultTypeIcon.rawValue
    }

    enum SortColumn: String {
        case label
        case key
        case fields
    }

    private(set) var sortColumn: SortColumn = .label
    private(set) var sortAscending = true

    private(set) var mode: VocabularyPaneMode = .empty
    var draft: Draft?
    private(set) var isSaving = false
    var formError: String?
    var toast: VocabularyToast?

    var suggestionError: String?
    var assignPick = ""
    private(set) var isAssigning = false
    private(set) var removingFieldID: String?

    private(set) var pendingDeleteID: String?
    private(set) var isDeleting = false
    private(set) var deleteError: String?

    private let userID: String
    private let store: any GenealogyStore
    private let session: WorkspaceSession
    private let catalogCounts: CatalogCounts?

    init(
        session: WorkspaceSession,
        userID: String,
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil
    ) {
        self.session = session
        self.userID = userID
        self.store = store
        self.catalogCounts = catalogCounts
    }

    // MARK: Session handles

    static func typesListKey(for session: WorkspaceSession) -> CatalogQueryKey {
        CatalogQueryKey.sourceTypesList(project: session.projectKey)
    }

    static func fieldsListKey(for session: WorkspaceSession) -> CatalogQueryKey {
        CatalogQueryKey.metadataFieldsList(project: session.projectKey)
    }

    static func suggestionsKey(for session: WorkspaceSession, typeID: String) -> CatalogQueryKey {
        CatalogQueryKey.typeSuggestions(project: session.projectKey, typeId: typeID)
    }

    func warmListQueries() {
        let _: QueryHandle<[CatalogSourceType]> = session.query(Self.typesListKey(for: session))
        let _: QueryHandle<[CatalogMetadataField]> = session.query(Self.fieldsListKey(for: session))
    }

    func warmSuggestions(for typeID: String) {
        let _: QueryHandle<[CatalogTypeSuggestion]> = session.query(Self.suggestionsKey(for: session, typeID: typeID))
    }

    var types: [CatalogSourceType] {
        session.queryHandle(Self.typesListKey(for: session))?.value ?? []
    }

    var fields: [CatalogMetadataField] {
        session.queryHandle(Self.fieldsListKey(for: session))?.value ?? []
    }

    var isLoading: Bool {
        guard let handle: QueryHandle<[CatalogSourceType]> = session.queryHandle(Self.typesListKey(for: session))
        else { return true }
        return handle.status == .loading && types.isEmpty
    }

    var loadError: Error? {
        guard types.isEmpty else { return nil }
        let typesHandle: QueryHandle<[CatalogSourceType]>? = session.queryHandle(Self.typesListKey(for: session))
        let fieldsHandle: QueryHandle<[CatalogMetadataField]>? = session.queryHandle(Self.fieldsListKey(for: session))
        return typesHandle?.error ?? fieldsHandle?.error
    }

    var selectedTypeID: String? {
        switch mode {
        case .viewing(let id), .editing(let id): id
        case .empty, .adding: nil
        }
    }

    var suggestions: [CatalogTypeSuggestion] {
        guard let typeID = selectedTypeID else { return [] }
        return session.queryHandle(Self.suggestionsKey(for: session, typeID: typeID))?.value ?? []
    }

    var isLoadingSuggestions: Bool {
        guard let typeID = selectedTypeID else { return false }
        guard let handle: QueryHandle<[CatalogTypeSuggestion]> = session.queryHandle(
            Self.suggestionsKey(for: session, typeID: typeID)
        ) else { return false }
        return handle.status == .loading && handle.value == nil
    }

    // MARK: Derived

    var visibleTypes: [CatalogSourceType] {
        types.sorted { a, b in
            let order = compare(a, b)
            return sortAscending ? order == .orderedAscending : order == .orderedDescending
        }
    }

    var isAdding: Bool {
        if case .adding = mode { return true }
        return false
    }

    var selectedType: CatalogSourceType? {
        switch mode {
        case .viewing(let id), .editing(let id):
            return types.first { $0.id == id }
        case .empty, .adding:
            return nil
        }
    }

    var isSelectedTypeLocked: Bool {
        if case .viewing = mode { return true }
        return false
    }

    var draftKey: String {
        FieldSlug.kebab(draft?.label ?? "")
    }

    var isDirty: Bool {
        guard case .editing = mode, let type = selectedType, let draft else { return false }
        return draft.label != type.label
            || draft.description != type.description
            || draft.iconKey != type.iconKey
    }

    var canSubmit: Bool {
        guard let draft, !isSaving, !draft.label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            return false
        }
        return isAdding || isDirty
    }

    var canEditAssociations: Bool {
        guard let type = selectedType else { return false }
        return !CatalogOrigin.isPlugin(type.origin)
    }

    var assignPool: [CatalogMetadataField] {
        let assigned = Set(suggestions.map(\.field.id))
        return fields
            .filter { !assigned.contains($0.id) }
            .sorted { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }
    }

    var canAssign: Bool {
        !assignPick.isEmpty && !isAssigning && canEditAssociations
    }

    var pickedField: CatalogMetadataField? {
        assignPool.first { $0.id == assignPick }
    }

    var assignTooltip: LocalizedStringResource {
        if assignPool.isEmpty { return L10n.SourceTypes.assignTipPoolEmpty }
        guard let field = pickedField else { return L10n.SourceTypes.assignTipChoose }
        return L10n.SourceTypes.assignTipField(label: field.label)
    }

    var assignAccessibilityLabel: LocalizedStringResource {
        guard let field = pickedField, let type = selectedType else {
            return L10n.SourceTypes.assignField
        }
        return L10n.SourceTypes.assignFieldNamed(field: field.label, type: type.label)
    }

    var showsDelete: Bool { !isAdding && selectedType != nil }

    var canDeleteSelectedType: Bool {
        guard let type = selectedType else { return false }
        return !CatalogOrigin.isPlugin(type.origin) && type.usedBy == 0
    }

    var deleteTooltip: LocalizedStringResource {
        guard let type = selectedType else { return L10n.SourceTypes.deleteType }
        if CatalogOrigin.isPlugin(type.origin) { return L10n.SourceTypes.deleteOwnedByPlugin }
        if type.usedBy > 0 { return L10n.SourceTypes.deleteInUse(count: type.usedBy) }
        return L10n.SourceTypes.deleteType
    }

    var pendingDeleteType: CatalogSourceType? {
        guard let pendingDeleteID else { return nil }
        return types.first { $0.id == pendingDeleteID }
    }

    var countLine: String {
        let summary = catalogCounts?.sourceTypes ?? .from(types)
        if summary.plugin > 0 {
            return L10n.SourceTypes.countLineWithPlugin(
                total: summary.total, seeded: summary.seeded, user: summary.user, plugin: summary.plugin
            )
        }
        return L10n.SourceTypes.countLine(total: summary.total, seeded: summary.seeded, user: summary.user)
    }

    // MARK: Selection

    @discardableResult
    func syncSelection(from location: WorkspaceLocation) -> WorkspaceLocationReconcile {
        guard location.section == .sourceTypes else { return .ignored }
        guard let typesHandle: QueryHandle<[CatalogSourceType]> = session.queryHandle(Self.typesListKey(for: session)),
              typesHandle.status == .ready || !types.isEmpty else { return .ignored }
        if isAdding { return .ignored }
        if let typeId = location.typeId {
            guard applySelection(typeId) else {
                clearHistorySelection()
                return .missingDeepId
            }
            warmSuggestions(for: typeId)
            return .applied
        }
        clearHistorySelection()
        return .applied
    }

    func sortBy(_ columnID: String) {
        guard let column = SortColumn(rawValue: columnID) else { return }
        if column == sortColumn {
            sortAscending.toggle()
        } else {
            sortColumn = column
            sortAscending = true
        }
    }

    func select(_ id: String) {
        guard applySelection(id) else { return }
        warmSuggestions(for: id)
    }

    @discardableResult
    private func applySelection(_ id: String) -> Bool {
        guard let type = types.first(where: { $0.id == id }) else { return false }
        formError = nil
        suggestionError = nil
        assignPick = ""
        draft = Draft(label: type.label, description: type.description, iconKey: type.iconKey)
        mode = CatalogOrigin.isPlugin(type.origin) ? .viewing(id: id) : .editing(id: id)
        return true
    }

    func clearHistorySelection() {
        guard !isAdding else { return }
        mode = .empty
        suggestionError = nil
        assignPick = ""
    }

    func openAdd() {
        let resumeID: String? = switch mode {
        case .viewing(let id), .editing(let id): id
        case .empty, .adding: nil
        }
        formError = nil
        suggestionError = nil
        assignPick = ""
        mode = .adding(resumeID: resumeID)
        draft = Draft(label: "", description: "", iconKey: PVEvidenceIconKey.defaultTypeIcon.rawValue)
    }

    func cancelAdd() {
        guard case .adding(let resumeID) = mode else { return }
        formError = nil
        if let resumeID {
            select(resumeID)
        } else {
            mode = .empty
        }
    }

    func revertEdit() {
        guard case .editing(let id) = mode, let type = types.first(where: { $0.id == id }) else { return }
        draft = Draft(label: type.label, description: type.description, iconKey: type.iconKey)
        formError = nil
    }

    @discardableResult
    func submit() async -> WorkspaceLocation? {
        guard let draft else { return nil }
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !label.isEmpty else {
            formError = String(localized: L10n.SourceTypes.errorLabelRequired)
            return nil
        }
        if isAdding && FieldSlug.kebab(label).isEmpty {
            formError = String(localized: L10n.SourceTypes.errorUnslugifiable)
            return nil
        }
        isSaving = true
        formError = nil
        defer { isSaving = false }
        do {
            switch mode {
            case .adding:
                let created = try await store.createSourceType(
                    projectDir: session.projectKey.projectDir, userID: userID,
                    label: label, description: draft.description, iconKey: draft.iconKey
                )
                session.apply(.createdSourceType)
                patchTypesList { rows in
                    rows.append(created)
                }
                mode = .editing(id: created.id)
                self.draft = Draft(label: created.label, description: created.description, iconKey: created.iconKey)
                toast = VocabularyToast(
                    title: String(localized: L10n.SourceTypes.toastAddedTitle),
                    body: L10n.SourceTypes.toastAddedBody(label: created.label, key: created.key),
                    tone: .success
                )
                publishCounts()
                return WorkspaceLocation(
                    section: .sourceTypes,
                    typeId: created.id,
                    title: created.label
                )
            case .editing(let id):
                let updated = try await store.updateSourceType(
                    projectDir: session.projectKey.projectDir, userID: userID, typeID: id,
                    label: label, description: draft.description, iconKey: draft.iconKey
                )
                session.apply(.updatedSourceType(id: updated.id))
                patchTypesList { rows in
                    if let idx = rows.firstIndex(where: { $0.id == updated.id }) {
                        rows[idx] = updated
                    }
                }
                mode = .editing(id: updated.id)
                self.draft = Draft(label: updated.label, description: updated.description, iconKey: updated.iconKey)
                toast = VocabularyToast(
                    title: String(localized: L10n.SourceTypes.toastUpdatedTitle),
                    body: L10n.SourceTypes.toastUpdatedBody(label: updated.label, key: updated.key),
                    tone: .success
                )
                return nil
            case .empty, .viewing:
                return nil
            }
        } catch {
            formError = L10n.Errors.message(for: error)
            return nil
        }
    }

    func assignPickedField() async {
        guard canAssign, let type = selectedType else { return }
        let fieldID = assignPick
        guard let field = fields.first(where: { $0.id == fieldID }) else { return }
        isAssigning = true
        suggestionError = nil
        defer { isAssigning = false }
        do {
            let updated = try await store.assignTypeField(
                projectDir: session.projectKey.projectDir, userID: userID, typeID: type.id, fieldID: fieldID
            )
            session.apply(.assignedTypeSuggestion(typeId: type.id))
            patchSuggestions(updated, typeID: type.id)
            patchTypesList { rows in
                if let idx = rows.firstIndex(where: { $0.id == type.id }) {
                    rows[idx].suggestedFieldCount = updated.count
                }
            }
            assignPick = ""
            toast = VocabularyToast(
                title: String(localized: L10n.SourceTypes.toastAssignedTitle),
                body: L10n.SourceTypes.toastAssignedBody(field: field.label, type: type.label),
                tone: .success
            )
        } catch {
            suggestionError = L10n.Errors.message(for: error)
        }
    }

    func removeSuggestion(fieldID: String) async {
        guard canEditAssociations, removingFieldID == nil, let type = selectedType else { return }
        guard let field = suggestions.first(where: { $0.field.id == fieldID })?.field else { return }
        removingFieldID = fieldID
        suggestionError = nil
        defer { removingFieldID = nil }
        do {
            let updated = try await store.removeTypeField(
                projectDir: session.projectKey.projectDir, userID: userID, typeID: type.id, fieldID: fieldID
            )
            session.apply(.removedTypeSuggestion(typeId: type.id))
            patchSuggestions(updated, typeID: type.id)
            patchTypesList { rows in
                if let idx = rows.firstIndex(where: { $0.id == type.id }) {
                    rows[idx].suggestedFieldCount = updated.count
                }
            }
            toast = VocabularyToast(
                title: String(localized: L10n.SourceTypes.toastRemovedTitle),
                body: L10n.SourceTypes.toastRemovedBody(
                    field: field.label, type: type.label, valueCount: field.usedBy
                ),
                tone: .info
            )
        } catch {
            suggestionError = L10n.Errors.message(for: error)
        }
    }

    func askDelete() {
        guard canDeleteSelectedType, let type = selectedType else { return }
        deleteError = nil
        pendingDeleteID = type.id
    }

    func cancelDelete() {
        guard !isDeleting else { return }
        pendingDeleteID = nil
    }

    @discardableResult
    func confirmDelete() async -> Bool {
        guard let type = pendingDeleteType, !isDeleting else { return false }
        isDeleting = true
        deleteError = nil
        defer { isDeleting = false }
        do {
            try await store.deleteSourceType(
                projectDir: session.projectKey.projectDir, userID: userID, typeID: type.id
            )
            session.apply(.deletedSourceType(id: type.id))
            patchTypesList { rows in rows.removeAll { $0.id == type.id } }
            pendingDeleteID = nil
            formError = nil
            suggestionError = nil
            mode = .empty
            toast = VocabularyToast(
                title: String(localized: L10n.SourceTypes.toastDeletedTitle),
                body: L10n.SourceTypes.toastDeletedBody(label: type.label),
                tone: .success
            )
            publishCounts()
            return true
        } catch {
            deleteError = L10n.Errors.message(for: error)
            return false
        }
    }

    // MARK: Helpers

    private func publishCounts() {
        catalogCounts?.publishSourceTypes(.from(types))
        catalogCounts?.publishSourceFields(.from(fields))
    }

    private func patchTypesList(_ mutate: (inout [CatalogSourceType]) -> Void) {
        let key = CatalogQueryKey.sourceTypesList(project: session.projectKey)
        guard let handle: QueryHandle<[CatalogSourceType]> = session.queryHandle(key) else { return }
        var rows = handle.value ?? []
        mutate(&rows)
        session.setQueryValue(key, value: rows)
    }

    private func patchSuggestions(_ suggestions: [CatalogTypeSuggestion], typeID: String) {
        let key = CatalogQueryKey.typeSuggestions(project: session.projectKey, typeId: typeID)
        session.setQueryValue(key, value: suggestions)
    }

    private func compare(_ a: CatalogSourceType, _ b: CatalogSourceType) -> ComparisonResult {
        switch sortColumn {
        case .label:
            return a.label.localizedCaseInsensitiveCompare(b.label)
        case .key:
            return a.key.localizedCaseInsensitiveCompare(b.key)
        case .fields:
            if a.suggestedFieldCount == b.suggestedFieldCount {
                return a.label.localizedCaseInsensitiveCompare(b.label)
            }
            return a.suggestedFieldCount < b.suggestedFieldCount ? .orderedAscending : .orderedDescending
        }
    }
}
