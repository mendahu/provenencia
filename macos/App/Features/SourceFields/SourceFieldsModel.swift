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

    private(set) var fields: [CatalogMetadataField] = []
    private(set) var isLoading = false
    /// True after the first `load()` finishes — gates history prune until rows exist.
    private(set) var hasCompletedInitialLoad = false
    var loadError: Error?

    private(set) var sortAscending = true

    /// Detail-pane mode — see `VocabularyPaneMode` for the invariants.
    private(set) var mode: VocabularyPaneMode = .empty
    /// Non-nil whenever a field is selected or the add form is open.
    /// Cleared only for `.empty`. Do not nil this while the edit/add form
    /// may still be in the hierarchy — `@Bindable` projections into an
    /// optional trap if it becomes nil mid-update (edit/add → view).
    var draft: Draft?
    private(set) var isSaving = false
    var formError: String?
    var toast: VocabularyToast?

    /// The field the delete confirmation is open for. Held as an id (not a
    /// `Bool`) so the dialog keeps naming the right field even if selection
    /// moves underneath it.
    private(set) var pendingDeleteID: String?
    private(set) var isDeleting = false
    /// Only set when a confirmed delete failed — the dialog stays open and
    /// says why. The in-use case is normally caught before this by
    /// `canDeleteSelectedField`.
    private(set) var deleteError: String?

    private let projectDir: String
    private let userID: String
    private let store: any GenealogyStore
    /// Shared sidebar / header totals. Nil in isolated unit tests and
    /// previews that don't mount a workspace.
    private let catalogCounts: CatalogCounts?

    init(
        projectDir: String,
        userID: String,
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil
    ) {
        self.projectDir = projectDir
        self.userID = userID
        self.store = store
        self.catalogCounts = catalogCounts
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

    /// Live key preview while adding — mirrors what the engine will mint.
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

    /// The detail pane shows a delete affordance whenever a saved field is
    /// selected — disabled, with a reason, when it cannot be deleted.
    var showsDelete: Bool { !isAdding && selectedField != nil }

    /// A field is deletable when this project owns its definition (a plugin
    /// owns its own) and no source carries a value for it yet. The engine
    /// enforces the second rule too (`sourcefields.in_use`).
    var canDeleteSelectedField: Bool {
        guard let field = selectedField else { return false }
        return !CatalogOrigin.isPlugin(field.origin) && field.usedBy == 0
    }

    /// Tooltip on the delete button — the action when it is available, the
    /// reason it is not when it is disabled.
    var deleteTooltip: LocalizedStringResource {
        guard let field = selectedField else { return L10n.SourceFields.deleteField }
        if CatalogOrigin.isPlugin(field.origin) { return L10n.SourceFields.deleteOwnedByPlugin }
        if field.usedBy > 0 { return L10n.SourceFields.deleteInUse(count: field.usedBy) }
        return L10n.SourceFields.deleteField
    }

    /// The field the open confirmation refers to, if any.
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

    // MARK: Actions

    func load() async {
        isLoading = true
        loadError = nil
        defer {
            isLoading = false
            hasCompletedInitialLoad = true
        }
        do {
            fields = try await store.listMetadataFields(projectDir: projectDir)
            publishCounts()
        } catch {
            loadError = error
        }
    }

    func toggleLabelSort() {
        sortAscending.toggle()
    }

    func select(_ id: String) {
        guard let field = fields.first(where: { $0.id == id }) else { return }
        formError = nil
        // Always keep `draft` non-nil here. The locked (.viewing) panel does
        // not bind it, but going edit/add → view with `draft = nil` in the
        // same turn tears down `Binding($model.draft)` and traps.
        draft = Draft(label: field.label, dataType: field.dataType, description: field.description)
        mode = CatalogOrigin.isPlugin(field.origin) ? .viewing(id: id) : .editing(id: id)
    }

    /// Clears master–detail selection when history restores a section root.
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
            // Leave `draft` in place — nilling it in the same turn as
            // removing the form races `@Bindable` optional projections.
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
        // `deleteError` is deliberately left in place: the sheet is still
        // animating out and reads it live, so nilling it here blanks the
        // error callout mid-dismissal. `askDelete` resets it anyway.
    }

    /// Deletes the pending field. Returns `true` on success so the view can
    /// `fallbackToSectionRoot()` and keep history aligned with empty detail.
    @discardableResult
    func confirmDelete() async -> Bool {
        guard let field = pendingDeleteField, !isDeleting else { return false }
        isDeleting = true
        deleteError = nil
        defer { isDeleting = false }
        do {
            try await store.deleteMetadataField(projectDir: projectDir, userID: userID, fieldID: field.id)
            fields.removeAll { $0.id == field.id }
            pendingDeleteID = nil
            formError = nil
            // Leave `draft` in place, same reason as `cancelAdd`: nilling it
            // in the same turn the form leaves the hierarchy races
            // `@Bindable` optional projections.
            mode = .empty
            toast = VocabularyToast(
                title: String(localized: L10n.SourceFields.toastDeletedTitle),
                body: L10n.SourceFields.toastDeletedBody(label: field.label),
                tone: .success
            )
            publishCounts()
            return true
        } catch {
            deleteError = L10n.Errors.message(for: error)
            return false
        }
    }

    /// Saves the draft. On successful **create**, returns the location the
    /// view must `go(to:)` so history matches the new selection. Edits and
    /// failures return `nil` (place unchanged).
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
                    projectDir: projectDir, userID: userID,
                    label: label, dataType: draft.dataType, description: draft.description
                )
                fields.append(created)
                mode = .editing(id: created.id)
                self.draft = Draft(label: created.label, dataType: created.dataType, description: created.description)
                toast = VocabularyToast(
                    title: String(localized: L10n.SourceFields.toastAddedTitle),
                    body: L10n.SourceFields.toastAddedBody(label: created.label, key: created.key),
                    tone: .success
                )
                publishCounts()
                return WorkspaceLocation(
                    section: .sourceFields,
                    fieldId: created.id,
                    title: created.label
                )
            case .editing(let id):
                let updated = try await store.updateMetadataField(
                    projectDir: projectDir, userID: userID, fieldID: id,
                    label: label, dataType: draft.dataType, description: draft.description
                )
                if let idx = fields.firstIndex(where: { $0.id == updated.id }) {
                    fields[idx] = updated
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

    private func publishCounts() {
        catalogCounts?.publishSourceFields(.from(fields))
    }
}
