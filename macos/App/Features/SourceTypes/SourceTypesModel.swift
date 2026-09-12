import Foundation
import Observation

/// State for the **Source types** workspace destination (S2-03 board /
/// S2-16 PR): browse/search/sort the project's `source_types` vocabulary,
/// create or edit project rows (`user` and create-time `provenencia`
/// starters), delete unused ones, and assign or remove the
/// `source_type_metadata_fields` suggestions each type carries.
/// Plugin-origin types stay view-only — see `isSelectedTypeLocked`.
@MainActor
@Observable
final class SourceTypesModel {
    /// The add/edit form's in-progress values. `key` is never part of this —
    /// it's minted server-side from `label` (`FieldSlug.kebab` mirrors the
    /// preview client-side; the engine is the source of truth).
    struct Draft: Equatable {
        var label: String
        var description: String
        var iconKey: String = PVEvidenceIconKey.defaultTypeIcon.rawValue
    }

    /// Column ids the list can sort by. They double as `PVTable` column ids.
    enum SortColumn: String {
        case label
        case key
        case fields
    }

    private(set) var types: [CatalogSourceType] = []
    /// The assignment pool — the whole Source fields vocabulary (S2-02).
    /// This destination never edits it, only reads it to offer suggestions.
    private(set) var fields: [CatalogMetadataField] = []
    private(set) var isLoading = false
    var loadError: Error?

    var query = ""
    private(set) var sortColumn: SortColumn = .label
    private(set) var sortAscending = true

    /// Detail-pane mode — see `VocabularyPaneMode` for the invariants.
    private(set) var mode: VocabularyPaneMode = .empty
    /// Non-nil whenever a type is selected or the add form is open. Cleared
    /// only for `.empty`. Do not nil this while the edit/add form may still
    /// be in the hierarchy — `@Bindable` projections into an optional trap if
    /// it becomes nil mid-update (edit/add → view).
    var draft: Draft?
    private(set) var isSaving = false
    var formError: String?
    var toast: VocabularyToast?

    /// The selected type's suggestions, in the engine's `sort_order`.
    private(set) var suggestions: [CatalogTypeSuggestion] = []
    private(set) var isLoadingSuggestions = false
    /// Set when assigning, removing, or loading suggestions failed. Shown as
    /// a callout inside the suggestions section, not as a form error — the
    /// label/description form is a separate concern.
    var suggestionError: String?
    /// The field id picked in the assign combo box; empty when nothing is
    /// chosen yet.
    var assignPick = ""
    private(set) var isAssigning = false
    /// The suggestion whose remove is in flight, so only that row disables.
    private(set) var removingFieldID: String?

    /// The type the delete confirmation is open for. Held as an id (not a
    /// `Bool`) so the dialog keeps naming the right type even if selection
    /// moves underneath it.
    private(set) var pendingDeleteID: String?
    private(set) var isDeleting = false
    /// Only set when a confirmed delete failed — the dialog stays open and
    /// says why. The in-use case is normally caught before this by
    /// `canDeleteSelectedType`.
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

    var visibleTypes: [CatalogSourceType] {
        types.matching(query).sorted { a, b in
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

    /// Live key preview while adding — mirrors what the engine will mint.
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

    /// Associations are editable for exactly the types whose definition this
    /// project owns — a plugin owns its own suggestions too (S2-03 §7).
    var canEditAssociations: Bool {
        guard let type = selectedType else { return false }
        return !CatalogOrigin.isPlugin(type.origin)
    }

    /// The fields not yet suggested for the selected type, by label. The pool
    /// is the Source fields vocabulary and nothing else — this destination
    /// never creates a field inline (S2-03 T-9).
    var assignPool: [CatalogMetadataField] {
        let assigned = Set(suggestions.map(\.field.id))
        return fields
            .filter { !assigned.contains($0.id) }
            .sorted { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }
    }

    var canAssign: Bool {
        !assignPick.isEmpty && !isAssigning && canEditAssociations
    }

    /// The field the assign control would attach, if one is picked.
    var pickedField: CatalogMetadataField? {
        assignPool.first { $0.id == assignPick }
    }

    /// Tooltip on the assign button — the action when it is available, the
    /// reason it is not when it is disabled. Same disabled-not-hidden pattern
    /// as `deleteTooltip`.
    var assignTooltip: LocalizedStringResource {
        if assignPool.isEmpty { return L10n.SourceTypes.assignTipPoolEmpty }
        guard let field = pickedField else { return L10n.SourceTypes.assignTipChoose }
        return L10n.SourceTypes.assignTipField(label: field.label)
    }

    /// Spoken label for the same button. The tooltip can lean on what the eye
    /// already has — the picked field sits right beside it — but VoiceOver
    /// has to name both ends of the association.
    var assignAccessibilityLabel: LocalizedStringResource {
        guard let field = pickedField, let type = selectedType else {
            return L10n.SourceTypes.assignField
        }
        return L10n.SourceTypes.assignFieldNamed(field: field.label, type: type.label)
    }

    /// The detail pane shows a delete affordance whenever a saved type is
    /// selected — disabled, with a reason, when it cannot be deleted.
    var showsDelete: Bool { !isAdding && selectedType != nil }

    /// A type is deletable when this project owns its definition (a plugin
    /// owns its own) and no source is classified as it yet. The engine
    /// enforces the second rule too (`sourcetypes.in_use`).
    var canDeleteSelectedType: Bool {
        guard let type = selectedType else { return false }
        return !CatalogOrigin.isPlugin(type.origin) && type.usedBy == 0
    }

    /// Tooltip on the delete button — the action when it is available, the
    /// reason it is not when it is disabled.
    var deleteTooltip: LocalizedStringResource {
        guard let type = selectedType else { return L10n.SourceTypes.deleteType }
        if CatalogOrigin.isPlugin(type.origin) { return L10n.SourceTypes.deleteOwnedByPlugin }
        if type.usedBy > 0 { return L10n.SourceTypes.deleteInUse(count: type.usedBy) }
        return L10n.SourceTypes.deleteType
    }

    /// The type the open confirmation refers to, if any.
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

    // MARK: Actions

    func load() async {
        isLoading = true
        loadError = nil
        defer { isLoading = false }
        do {
            types = try await store.listSourceTypes(projectDir: projectDir)
            fields = try await store.listMetadataFields(projectDir: projectDir)
            publishCounts()
        } catch {
            loadError = error
        }
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
        guard let type = types.first(where: { $0.id == id }) else { return }
        formError = nil
        suggestionError = nil
        assignPick = ""
        // Always keep `draft` non-nil here. The locked (.viewing) panel does
        // not bind it, but going edit/add → view with `draft = nil` in the
        // same turn tears down `Binding($model.draft)` and traps.
        draft = Draft(label: type.label, description: type.description, iconKey: type.iconKey)
        mode = CatalogOrigin.isPlugin(type.origin) ? .viewing(id: id) : .editing(id: id)
        Task { await loadSuggestions(for: id) }
    }

    func openAdd() {
        let resumeID: String? = switch mode {
        case .viewing(let id), .editing(let id): id
        case .empty, .adding: nil
        }
        formError = nil
        suggestionError = nil
        assignPick = ""
        suggestions = []
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
            // Leave `draft` in place — nilling it in the same turn as
            // removing the form races `@Bindable` optional projections.
        }
    }

    func revertEdit() {
        guard case .editing(let id) = mode, let type = types.first(where: { $0.id == id }) else { return }
        draft = Draft(label: type.label, description: type.description, iconKey: type.iconKey)
        formError = nil
    }

    func submit() async {
        guard let draft else { return }
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !label.isEmpty else {
            formError = String(localized: L10n.SourceTypes.errorLabelRequired)
            return
        }
        if isAdding && FieldSlug.kebab(label).isEmpty {
            formError = String(localized: L10n.SourceTypes.errorUnslugifiable)
            return
        }
        isSaving = true
        formError = nil
        defer { isSaving = false }
        do {
            switch mode {
            case .adding:
                let created = try await store.createSourceType(
                    projectDir: projectDir, userID: userID,
                    label: label, description: draft.description, iconKey: draft.iconKey
                )
                types.append(created)
                query = ""
                mode = .editing(id: created.id)
                self.draft = Draft(label: created.label, description: created.description, iconKey: created.iconKey)
                // A new type suggests nothing yet (S2-03 T-15) — assigning
                // happens on the detail the save lands you on.
                suggestions = []
                toast = VocabularyToast(
                    title: String(localized: L10n.SourceTypes.toastAddedTitle),
                    body: L10n.SourceTypes.toastAddedBody(label: created.label, key: created.key),
                    tone: .success
                )
                publishCounts()
            case .editing(let id):
                let updated = try await store.updateSourceType(
                    projectDir: projectDir, userID: userID, typeID: id,
                    label: label, description: draft.description, iconKey: draft.iconKey
                )
                replace(updated)
                mode = .editing(id: updated.id)
                self.draft = Draft(label: updated.label, description: updated.description, iconKey: updated.iconKey)
                toast = VocabularyToast(
                    title: String(localized: L10n.SourceTypes.toastUpdatedTitle),
                    body: L10n.SourceTypes.toastUpdatedBody(label: updated.label, key: updated.key),
                    tone: .success
                )
            case .empty, .viewing:
                break
            }
        } catch {
            formError = L10n.Errors.message(for: error)
        }
    }

    // MARK: Suggestions

    func loadSuggestions(for typeID: String) async {
        isLoadingSuggestions = true
        defer { isLoadingSuggestions = false }
        do {
            let loaded = try await store.listTypeSuggestions(projectDir: projectDir, typeID: typeID)
            // Selection may have moved while the read was in flight; a stale
            // reply must not overwrite the pane the researcher is looking at.
            guard selectedType?.id == typeID else { return }
            suggestions = loaded
            suggestionError = nil
        } catch {
            guard selectedType?.id == typeID else { return }
            suggestions = []
            suggestionError = L10n.Errors.message(for: error)
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
                projectDir: projectDir, userID: userID, typeID: type.id, fieldID: fieldID
            )
            apply(updated, to: type.id)
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

    /// Detaching a suggestion is not destructive, so it takes no confirmation
    /// — the toast carries the reassurance instead (S2-03 T-10 / T-20): the
    /// field stays in the vocabulary and its existing values are untouched.
    func removeSuggestion(fieldID: String) async {
        guard canEditAssociations, removingFieldID == nil, let type = selectedType else { return }
        guard let field = suggestions.first(where: { $0.field.id == fieldID })?.field else { return }
        removingFieldID = fieldID
        suggestionError = nil
        defer { removingFieldID = nil }
        do {
            let updated = try await store.removeTypeField(
                projectDir: projectDir, userID: userID, typeID: type.id, fieldID: fieldID
            )
            apply(updated, to: type.id)
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

    // MARK: Delete

    func askDelete() {
        guard canDeleteSelectedType, let type = selectedType else { return }
        deleteError = nil
        pendingDeleteID = type.id
    }

    func cancelDelete() {
        guard !isDeleting else { return }
        pendingDeleteID = nil
        // `deleteError` is deliberately left in place: the sheet is still
        // animating out and reads it live, so nilling it here blanks the
        // error callout mid-dismissal. `askDelete` resets it anyway.
    }

    func confirmDelete() async {
        guard let type = pendingDeleteType, !isDeleting else { return }
        isDeleting = true
        deleteError = nil
        defer { isDeleting = false }
        do {
            try await store.deleteSourceType(projectDir: projectDir, userID: userID, typeID: type.id)
            types.removeAll { $0.id == type.id }
            pendingDeleteID = nil
            formError = nil
            suggestionError = nil
            suggestions = []
            // Leave `draft` in place, same reason as `cancelAdd`: nilling it
            // in the same turn the form leaves the hierarchy races
            // `@Bindable` optional projections.
            mode = .empty
            toast = VocabularyToast(
                title: String(localized: L10n.SourceTypes.toastDeletedTitle),
                body: L10n.SourceTypes.toastDeletedBody(label: type.label),
                tone: .success
            )
            publishCounts()
        } catch {
            deleteError = L10n.Errors.message(for: error)
        }
    }

    // MARK: Helpers

    /// Types load also lists the fields pool — publish both so the fields
    /// sidebar badge stays honest without a second round trip.
    private func publishCounts() {
        catalogCounts?.publishSourceTypes(.from(types))
        catalogCounts?.publishSourceFields(.from(fields))
    }

    /// Adopts a suggestion list the engine just returned, keeping the list
    /// row's count column in step without a second round trip.
    private func apply(_ updated: [CatalogTypeSuggestion], to typeID: String) {
        suggestions = updated
        guard let idx = types.firstIndex(where: { $0.id == typeID }) else { return }
        types[idx].suggestedFieldCount = updated.count
    }

    private func replace(_ type: CatalogSourceType) {
        guard let idx = types.firstIndex(where: { $0.id == type.id }) else { return }
        types[idx] = type
    }

    private func compare(_ a: CatalogSourceType, _ b: CatalogSourceType) -> ComparisonResult {
        switch sortColumn {
        case .label:
            return a.label.localizedCaseInsensitiveCompare(b.label)
        case .key:
            return a.key.localizedCaseInsensitiveCompare(b.key)
        case .fields:
            // Ties fall back to label so the order stays stable while the
            // counts (which repeat a lot) do not decide it.
            if a.suggestedFieldCount == b.suggestedFieldCount {
                return a.label.localizedCaseInsensitiveCompare(b.label)
            }
            return a.suggestedFieldCount < b.suggestedFieldCount ? .orderedAscending : .orderedDescending
        }
    }
}
