import Foundation
import Observation

/// State for the **Source fields** workspace destination (S2-02 board /
/// S2-15 PR): browse/search/sort the project's `source_metadata_fields`
/// vocabulary, and create or edit project rows (`user` and create-time
/// `provenencia` starters). Plugin-origin rows stay view-only — see
/// `isSelectedFieldLocked`.
@MainActor
@Observable
final class SourceFieldsModel {
    /// The add/edit form's in-progress values. `key` is never part of this —
    /// it's minted server-side from `label` (`FieldSlug.kebab` mirrors the
    /// preview client-side; the engine is the source of truth).
    struct Draft: Equatable {
        var label: String
        var dataType: String
        var description: String
    }

    private(set) var sortAscending = true

    /// Detail-pane mode — see `VocabularyPaneMode` for the invariants.
    private(set) var mode: VocabularyPaneMode = .empty
    /// Non-nil whenever a field is selected or the add form is open.
    var draft: Draft?
    private(set) var isSaving = false
    var formError: String?
    var toast: VocabularyToast?

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

    static func fieldsListKey(for session: WorkspaceSession) -> CatalogQueryKey {
        CatalogQueryKey.metadataFieldsList(project: session.projectKey)
    }

    func warmFieldsQuery() {
        let _: QueryHandle<[CatalogMetadataField]> = session.query(Self.fieldsListKey(for: session))
    }

    var fields: [CatalogMetadataField] {
        session.queryHandle(Self.fieldsListKey(for: session))?.value ?? []
    }

    var isLoading: Bool {
        guard let handle: QueryHandle<[CatalogMetadataField]> = session.queryHandle(Self.fieldsListKey(for: session))
        else { return true }
        return handle.status == .loading && fields.isEmpty
    }

    var loadError: Error? {
        guard fields.isEmpty else { return nil }
        let handle: QueryHandle<[CatalogMetadataField]>? = session.queryHandle(Self.fieldsListKey(for: session))
        return handle?.error
    }

    // MARK: Derived

    var visibleFields: [CatalogMetadataField] {
        fields.sorted { a, b in
            let order = a.label.localizedCaseInsensitiveCompare(b.label)
            return sortAscending ? order == .orderedAscending : order == .orderedDescending
        }
    }

    var isAdding: Bool {
        if case .adding = mode { return true }
        return false
    }

    var selectedField: CatalogMetadataField? {
        switch mode {
        case .viewing(let id), .editing(let id):
            return fields.first { $0.id == id }
        case .empty, .adding:
            return nil
        }
    }

    var isSelectedFieldLocked: Bool {
        if case .viewing = mode { return true }
        return false
    }

    var draftKey: String {
        FieldSlug.kebab(draft?.label ?? "")
    }

    var isDirty: Bool {
        guard case .editing = mode, let field = selectedField, let draft else { return false }
        return draft.label != field.label || draft.description != field.description
    }

    var canSubmit: Bool {
        guard let draft, !isSaving, !draft.label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            return false
        }
        return isAdding || isDirty
    }

    var showsDelete: Bool { !isAdding && selectedField != nil }

    var canDeleteSelectedField: Bool {
        guard let field = selectedField else { return false }
        return !CatalogOrigin.isPlugin(field.origin) && field.usedBy == 0
    }

    var deleteTooltip: LocalizedStringResource {
        guard let field = selectedField else { return L10n.SourceFields.deleteField }
        if CatalogOrigin.isPlugin(field.origin) { return L10n.SourceFields.deleteOwnedByPlugin }
        if field.usedBy > 0 { return L10n.SourceFields.deleteInUse(count: field.usedBy) }
        return L10n.SourceFields.deleteField
    }

    var pendingDeleteField: CatalogMetadataField? {
        guard let pendingDeleteID else { return nil }
        return fields.first { $0.id == pendingDeleteID }
    }

    var countLine: String {
        let summary = catalogCounts?.sourceFields ?? .from(fields)
        if summary.plugin > 0 {
            return L10n.SourceFields.countLineWithPlugin(
                total: summary.total, seeded: summary.seeded, user: summary.user, plugin: summary.plugin
            )
        }
        return L10n.SourceFields.countLine(total: summary.total, seeded: summary.seeded, user: summary.user)
    }

    // MARK: Selection

    /// Reconciles list/detail state to `location` when rows are already cached.
    @discardableResult
    func syncSelection(from location: WorkspaceLocation) -> WorkspaceLocationReconcile {
        guard location.section == .sourceFields else { return .ignored }
        guard let handle: QueryHandle<[CatalogMetadataField]> = session.queryHandle(Self.fieldsListKey(for: session)),
              handle.status == .ready || !fields.isEmpty else { return .ignored }
        if isAdding { return .ignored }
        if let fieldId = location.fieldId {
            guard applySelection(fieldId) else {
                mode = .empty
                return .missingDeepId
            }
            return .applied
        }
        clearHistorySelection()
        return .applied
    }

    func toggleLabelSort() {
        sortAscending.toggle()
    }

    func select(_ id: String) {
        _ = applySelection(id)
    }

    @discardableResult
    private func applySelection(_ id: String) -> Bool {
        guard let field = fields.first(where: { $0.id == id }) else { return false }
        formError = nil
        draft = Draft(label: field.label, dataType: field.dataType, description: field.description)
        mode = CatalogOrigin.isPlugin(field.origin) ? .viewing(id: id) : .editing(id: id)
        return true
    }

    func clearHistorySelection() {
        guard !isAdding else { return }
        mode = .empty
    }

    func openAdd() {
        let resumeID: String? = switch mode {
        case .viewing(let id), .editing(let id): id
        case .empty, .adding: nil
        }
        formError = nil
        mode = .adding(resumeID: resumeID)
        draft = Draft(label: "", dataType: CatalogFieldDataType.text, description: "")
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
        guard case .editing(let id) = mode, let field = fields.first(where: { $0.id == id }) else { return }
        draft = Draft(label: field.label, dataType: field.dataType, description: field.description)
        formError = nil
    }

    func askDelete() {
        guard canDeleteSelectedField, let field = selectedField else { return }
        deleteError = nil
        pendingDeleteID = field.id
    }

    func cancelDelete() {
        guard !isDeleting else { return }
        pendingDeleteID = nil
    }

    @discardableResult
    func confirmDelete() async -> Bool {
        guard let field = pendingDeleteField, !isDeleting else { return false }
        isDeleting = true
        deleteError = nil
        defer { isDeleting = false }
        do {
            try await store.deleteMetadataField(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                fieldID: field.id
            )
            session.apply(.deletedMetadataField(id: field.id))
            patchFieldsList { rows in rows.removeAll { $0.id == field.id } }
            pendingDeleteID = nil
            formError = nil
            mode = .empty
            toast = VocabularyToast(
                title: String(localized: L10n.SourceFields.toastDeletedTitle),
                body: L10n.SourceFields.toastDeletedBody(label: field.label),
                tone: .success
            )
            syncCatalogCounts()
            return true
        } catch {
            deleteError = L10n.Errors.message(for: error)
            return false
        }
    }

    @discardableResult
    func submit() async -> WorkspaceLocation? {
        guard let draft else { return nil }
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !label.isEmpty else {
            formError = String(localized: L10n.SourceFields.errorLabelRequired)
            return nil
        }
        if isAdding && FieldSlug.kebab(label).isEmpty {
            formError = String(localized: L10n.SourceFields.errorUnslugifiable)
            return nil
        }
        isSaving = true
        formError = nil
        defer { isSaving = false }
        do {
            switch mode {
            case .adding:
                let created = try await store.createMetadataField(
                    projectDir: session.projectKey.projectDir, userID: userID,
                    label: label, dataType: draft.dataType, description: draft.description
                )
                session.apply(.createdMetadataField)
                patchFieldsList { rows in rows.append(created) }
                mode = .editing(id: created.id)
                self.draft = Draft(label: created.label, dataType: created.dataType, description: created.description)
                toast = VocabularyToast(
                    title: String(localized: L10n.SourceFields.toastAddedTitle),
                    body: L10n.SourceFields.toastAddedBody(label: created.label, key: created.key),
                    tone: .success
                )
                syncCatalogCounts()
                return WorkspaceLocation(
                    section: .sourceFields,
                    fieldId: created.id,
                    title: created.label
                )
            case .editing(let id):
                let updated = try await store.updateMetadataField(
                    projectDir: session.projectKey.projectDir, userID: userID, fieldID: id,
                    label: label, dataType: draft.dataType, description: draft.description
                )
                session.apply(.updatedMetadataField(id: updated.id))
                patchFieldsList { rows in
                    if let idx = rows.firstIndex(where: { $0.id == updated.id }) {
                        rows[idx] = updated
                    }
                }
                mode = .editing(id: updated.id)
                self.draft = Draft(label: updated.label, dataType: updated.dataType, description: updated.description)
                toast = VocabularyToast(
                    title: String(localized: L10n.SourceFields.toastUpdatedTitle),
                    body: L10n.SourceFields.toastUpdatedBody(label: updated.label, key: updated.key),
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

    func syncCatalogCounts() {
        catalogCounts?.publishSourceFields(.from(fields))
    }

    private func patchFieldsList(_ mutate: (inout [CatalogMetadataField]) -> Void) {
        let key = CatalogQueryKey.metadataFieldsList(project: session.projectKey)
        guard let handle: QueryHandle<[CatalogMetadataField]> = session.queryHandle(key) else { return }
        var rows = handle.value ?? []
        mutate(&rows)
        session.setQueryValue(key, value: rows)
    }
}
