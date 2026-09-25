import Foundation
import Observation

/// Researcher-creatable Property value types (S7-05). Schema also has `term`
/// (registry-only) — never offered here.
enum SubjectPropertyValueType {
    static let researcherCreatable = ["text", "integer", "date", "name", "subject"]

    static func label(_ valueType: String) -> LocalizedStringResource {
        switch valueType {
        case "text": return L10n.SubjectFields.valueTypeText
        case "integer": return L10n.SubjectFields.valueTypeInteger
        case "date": return L10n.SubjectFields.valueTypeDate
        case "name": return L10n.SubjectFields.valueTypeName
        case "subject": return L10n.SubjectFields.valueTypeSubject
        case "term": return L10n.SubjectFields.valueTypeTerm
        default: return L10n.SubjectFields.valueTypeText
        }
    }
}

/// State for the **Subject fields** workspace destination (S7-D2 board / S7-05).
@MainActor
@Observable
final class SubjectFieldsModel {
    struct Draft: Equatable {
        var label: String
        var valueType: String
        var description: String
        var bindTypeIDs: Set<String>
    }

    private(set) var selectedTypeKey: String?
    var searchQuery = ""
    /// ComboBox selection for “Add a property to {Type}” (resets after bind).
    var addPropertySelection = ""
    private(set) var selectedPropertyID: String?
    private(set) var createOpen = false
    var draft: Draft?
    private(set) var isSaving = false
    var formError: String?
    var toast: VocabularyToast?
    private(set) var lockedCallout: String?
    private(set) var pendingDeleteID: String?
    private(set) var isDeleting = false
    private(set) var deleteError: String?
    var searchFocused = false

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

    static func workspaceKey(for session: WorkspaceSession) -> CatalogQueryKey {
        .subjectFieldsWorkspace(project: session.projectKey)
    }

    func warmWorkspaceQuery() {
        let _: QueryHandle<SubjectFieldsSnapshot> = session.query(Self.workspaceKey(for: session))
    }

    var snapshot: SubjectFieldsSnapshot {
        session.queryHandle(Self.workspaceKey(for: session))?.value ?? .empty
    }

    var isLoading: Bool {
        guard let handle: QueryHandle<SubjectFieldsSnapshot> = session.queryHandle(Self.workspaceKey(for: session))
        else { return true }
        return handle.status == .loading && snapshot.properties.isEmpty
    }

    var loadError: Error? {
        guard snapshot.properties.isEmpty else { return nil }
        let handle: QueryHandle<SubjectFieldsSnapshot>? = session.queryHandle(Self.workspaceKey(for: session))
        return handle?.error
    }

    var types: [CatalogSubjectType] { snapshot.typesInPaletteOrder }

    var selectedType: CatalogSubjectType? {
        guard let selectedTypeKey else { return nil }
        return snapshot.types.first { $0.key == selectedTypeKey }
    }

    /// Board: type filter + keyword search only (no origin / value-type filters).
    var visibleProperties: [CatalogProperty] {
        let q = searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        return snapshot.properties
            .filter { property in
                if let selectedTypeKey,
                   let type = snapshot.types.first(where: { $0.key == selectedTypeKey }) {
                    let bound = snapshot.fieldsByTypeID[type.id]?.contains { $0.property.id == property.id } == true
                    if !bound { return false }
                }
                if !q.isEmpty {
                    let hay = (property.label + " " + property.key).lowercased()
                    if !hay.contains(q) { return false }
                }
                return true
            }
            .sorted { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }
    }

    /// Unbound properties offered by the “Add a property to {Type}” combo.
    var addPropertyOptions: [PVComboBoxOption] {
        guard let type = selectedType else { return [] }
        return snapshot.properties
            .filter { snapshot.binding(propertyID: $0.id, typeID: type.id) == nil }
            .sorted { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }
            .map {
                PVComboBoxOption(
                    value: $0.id,
                    label: $0.label,
                    subtext: "\($0.key) · \($0.valueType) · \($0.origin)"
                )
            }
    }

    var selectedProperty: CatalogProperty? {
        guard let selectedPropertyID else { return nil }
        return snapshot.properties.first { $0.id == selectedPropertyID }
    }

    var draftKey: String {
        FieldSlug.kebab(draft?.label ?? "")
    }

    var canSubmitCreate: Bool {
        guard let draft, !isSaving else { return false }
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return !label.isEmpty
            && !FieldSlug.kebab(label).isEmpty
            && SubjectPropertyValueType.researcherCreatable.contains(draft.valueType)
    }

    var canDeleteSelected: Bool {
        guard let property = selectedProperty else { return false }
        return property.origin == CatalogOrigin.user && property.usedBy == 0
    }

    var pendingDeleteProperty: CatalogProperty? {
        guard let pendingDeleteID else { return nil }
        return snapshot.properties.first { $0.id == pendingDeleteID }
    }

    func syncCatalogCounts() {
        catalogCounts?.publishSubjectFields(.from(snapshot.properties))
    }

    func selectType(_ key: String?) {
        // Board: pressing the focused type again clears the filter.
        selectedTypeKey = (key != nil && key == selectedTypeKey) ? nil : key
        lockedCallout = nil
        addPropertySelection = ""
    }

    func selectProperty(_ id: String?) {
        selectedPropertyID = id
        lockedCallout = nil
        formError = nil
    }

    func focusSearch() {
        searchFocused = true
    }

    func openCreate(prefillLabel: String = "") {
        formError = nil
        lockedCallout = nil
        createOpen = true
        var bind = Set<String>()
        if let selectedType {
            bind.insert(selectedType.id)
        }
        let label = prefillLabel
        draft = Draft(
            label: label,
            valueType: "text",
            description: "",
            bindTypeIDs: bind
        )
    }

    func closeCreate() {
        createOpen = false
        draft = nil
        formError = nil
    }

    func toggleDraftBind(typeID: String) {
        guard draft != nil else { return }
        if draft!.bindTypeIDs.contains(typeID) {
            draft!.bindTypeIDs.remove(typeID)
        } else {
            draft!.bindTypeIDs.insert(typeID)
        }
    }

    func askDelete() {
        guard canDeleteSelected, let property = selectedProperty else { return }
        deleteError = nil
        pendingDeleteID = property.id
    }

    func cancelDelete() {
        guard !isDeleting else { return }
        pendingDeleteID = nil
        deleteError = nil
    }

    /// ComboBox picked an unbound property — bind it to the focused type.
    func addPropertyFromCombo() async {
        guard let type = selectedType,
              !addPropertySelection.isEmpty,
              let property = snapshot.properties.first(where: { $0.id == addPropertySelection })
        else { return }
        selectedPropertyID = property.id
        await toggleBinding(to: type)
        addPropertySelection = ""
    }

    @discardableResult
    func confirmDelete() async -> Bool {
        guard let property = pendingDeleteProperty, !isDeleting else { return false }
        isDeleting = true
        deleteError = nil
        defer { isDeleting = false }
        do {
            try await store.deleteProperty(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                propertyID: property.id
            )
            session.apply(.deletedProperty)
            selectedPropertyID = nil
            pendingDeleteID = nil
            toast = VocabularyToast(
                title: String(localized: L10n.SubjectFields.toastDeletedTitle),
                body: L10n.SubjectFields.toastDeletedBody(label: property.label),
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
    func submitCreate() async -> Bool {
        guard let draft else { return false }
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !label.isEmpty else {
            formError = String(localized: L10n.SubjectFields.errorLabelRequired)
            return false
        }
        guard SubjectPropertyValueType.researcherCreatable.contains(draft.valueType) else {
            formError = String(localized: L10n.SubjectFields.errorValueType)
            return false
        }
        isSaving = true
        formError = nil
        defer { isSaving = false }
        do {
            let created = try await store.createProperty(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                label: label,
                valueType: draft.valueType,
                description: draft.description
            )
            for typeID in draft.bindTypeIDs {
                try await store.assignSubjectTypeField(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    subjectTypeID: typeID,
                    propertyID: created.id
                )
            }
            if draft.bindTypeIDs.isEmpty {
                session.apply(.createdProperty)
            } else {
                session.apply(.mutatedSubjectTypeFields)
            }
            closeCreate()
            selectedPropertyID = created.id
            toast = VocabularyToast(
                title: String(localized: L10n.SubjectFields.toastCreatedTitle),
                body: L10n.SubjectFields.toastCreatedBody(label: created.label),
                tone: .success
            )
            syncCatalogCounts()
            return true
        } catch {
            formError = L10n.Errors.message(for: error)
            return false
        }
    }

    func toggleBinding(to type: CatalogSubjectType) async {
        guard let property = selectedProperty else { return }
        lockedCallout = nil
        if let existing = snapshot.binding(propertyID: property.id, typeID: type.id) {
            if existing.locked {
                lockedCallout = L10n.SubjectFields.lockedBindingReason(typeLabel: type.label)
                return
            }
            do {
                try await store.removeSubjectTypeField(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    subjectTypeID: type.id,
                    propertyID: property.id
                )
                session.apply(.mutatedSubjectTypeFields)
            } catch {
                formError = L10n.Errors.message(for: error)
            }
            return
        }
        do {
            try await store.assignSubjectTypeField(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectTypeID: type.id,
                propertyID: property.id
            )
            session.apply(.mutatedSubjectTypeFields)
        } catch {
            formError = L10n.Errors.message(for: error)
        }
    }

    func toggleFocusedTypeBinding() async {
        guard let type = selectedType else { return }
        await toggleBinding(to: type)
    }

    func bindingLocked(propertyID: String, typeID: String) -> Bool {
        snapshot.binding(propertyID: propertyID, typeID: typeID)?.locked == true
    }

    func isBound(propertyID: String, typeID: String) -> Bool {
        snapshot.binding(propertyID: propertyID, typeID: typeID) != nil
    }
}
