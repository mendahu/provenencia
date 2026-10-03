import Foundation
import Observation

/// Researcher-creatable Property value types (S7-05). Schema also has `term`
/// (registry-only) — never offered here.
enum SubjectPropertyValueType {
    static let researcherCreatable = ["text", "integer", "date", "name", "subject"]

    static func label(_ valueType: String) -> LocalizedStringResource {
        switch valueType {
        case "text": return L10n.Properties.valueTypeText
        case "integer": return L10n.Properties.valueTypeInteger
        case "date": return L10n.Properties.valueTypeDate
        case "name": return L10n.Properties.valueTypeName
        case "subject": return L10n.Properties.valueTypeSubject
        case "term": return L10n.Properties.valueTypeTerm
        default: return L10n.Properties.valueTypeText
        }
    }
}

/// State for the **Properties** workspace destination (S7-D2 board / S7-05).
@MainActor
@Observable
final class PropertiesModel {
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
    private(set) var isEditingIdentity = false
    private(set) var isSaving = false
    var formError: String?
    var toast: VocabularyToast?
    private(set) var lockedCallout: String?
    let deleteImpact = DeleteImpactFlow()
    var pendingImpact: PVDeleteImpactRequest? {
        get { deleteImpact.request }
        set { deleteImpact.applyRequest(newValue) }
    }
    var isDeleting: Bool { deleteImpact.isRunning }
    var deleteError: String? { deleteImpact.error }
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
        .propertiesWorkspace(project: session.projectKey)
    }

    func warmWorkspaceQuery() {
        let _: QueryHandle<PropertiesSnapshot> = session.query(Self.workspaceKey(for: session))
    }

    var snapshot: PropertiesSnapshot {
        session.queryHandle(Self.workspaceKey(for: session))?.value ?? .empty
    }

    var isLoading: Bool {
        guard let handle: QueryHandle<PropertiesSnapshot> = session.queryHandle(Self.workspaceKey(for: session))
        else { return true }
        return handle.status == .loading && snapshot.properties.isEmpty
    }

    var loadError: Error? {
        guard snapshot.properties.isEmpty else { return nil }
        let handle: QueryHandle<PropertiesSnapshot>? = session.queryHandle(Self.workspaceKey(for: session))
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
                    let bound = snapshot.propertiesByTypeID[type.id]?.contains { $0.property.id == property.id } == true
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
        Self.draftKey(label: draft?.label ?? "")
    }

    static func draftKey(label: String) -> String {
        FieldSlug.kebab(label)
    }

    var canSubmitCreate: Bool {
        canSubmitCreate(draft)
    }

    func canSubmitCreate(_ draft: Draft?) -> Bool {
        guard let draft, !isSaving else { return false }
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return !label.isEmpty
            && !FieldSlug.kebab(label).isEmpty
            && SubjectPropertyValueType.researcherCreatable.contains(draft.valueType)
    }

    var showsDelete: Bool { selectedProperty != nil }
    var canDeleteSelected: Bool { selectedProperty != nil }

    var deleteTooltip: LocalizedStringResource { L10n.Properties.deleteProperty }

    var deleteAccessibilityLabel: String {
        guard let property = selectedProperty else {
            return String(localized: L10n.Properties.deleteProperty)
        }
        return L10n.Properties.deletePropertyAccessibility(label: property.label)
    }

    var canEditSelected: Bool {
        guard let property = selectedProperty else { return false }
        return !CatalogOrigin.isPlugin(property.origin)
    }

    func isEditDirty(label: String, description: String) -> Bool {
        guard let property = selectedProperty else { return false }
        return label.trimmingCharacters(in: .whitespacesAndNewlines) != property.label
            || description.trimmingCharacters(in: .whitespacesAndNewlines) != property.description
    }

    func canSubmitEdit(label: String, description: String) -> Bool {
        guard canEditSelected, !isSaving else { return false }
        let trimmed = label.trimmingCharacters(in: .whitespacesAndNewlines)
        return !trimmed.isEmpty && isEditDirty(label: label, description: description)
    }

    var pendingDeleteProperty: CatalogProperty? {
        guard let id = deleteImpact.request?.target.id else { return nil }
        return snapshot.properties.first { $0.id == id }
    }

    func syncCatalogCounts() {
        catalogCounts?.publishProperties(.from(snapshot.properties))
    }

    // MARK: Selection

    /// Reconciles strip category + inspector row to `location` once the snapshot is cached.
    /// Both levels are history identity; a missing type key or property id prunes to the list root.
    @discardableResult
    func syncSelection(from location: WorkspaceLocation) -> WorkspaceLocationReconcile {
        guard location.section == .properties else { return .ignored }
        guard let handle: QueryHandle<PropertiesSnapshot> = session.queryHandle(Self.workspaceKey(for: session)),
              handle.status == .ready || !snapshot.properties.isEmpty else { return .ignored }
        if createOpen { return .ignored }
        if let key = location.subjectTypeKey, !snapshot.types.contains(where: { $0.key == key }) {
            applySelection(typeKey: nil, propertyID: nil)
            return .missingDeepId
        }
        if let id = location.propertyId, !snapshot.properties.contains(where: { $0.id == id }) {
            applySelection(typeKey: location.subjectTypeKey, propertyID: nil)
            return .missingDeepId
        }
        applySelection(typeKey: location.subjectTypeKey, propertyID: location.propertyId)
        return .applied
    }

    /// Committed place for a strip press. Board: pressing the focused type again clears the filter.
    func location(afterPressingType key: String?) -> WorkspaceLocation {
        location(typeKey: typeKey(afterPressing: key), propertyID: selectedPropertyID)
    }

    /// Committed place for an inspector row pick; keeps the strip category.
    func location(selectingProperty id: String?) -> WorkspaceLocation {
        location(typeKey: selectedTypeKey, propertyID: id)
    }

    private func location(typeKey: String?, propertyID: String?) -> WorkspaceLocation {
        let property = propertyID.flatMap { id in snapshot.properties.first { $0.id == id } }
        let type = typeKey.flatMap { key in snapshot.types.first { $0.key == key } }
        return WorkspaceLocation(
            section: .properties,
            subjectTypeKey: typeKey,
            propertyId: propertyID,
            title: property?.label ?? type?.label
        )
    }

    private func typeKey(afterPressing key: String?) -> String? {
        (key != nil && key == selectedTypeKey) ? nil : key
    }

    /// Applies UI state for a strip press. History owns the committed place — see `location(afterPressingType:)`.
    func selectType(_ key: String?) {
        applyType(typeKey(afterPressing: key))
    }

    /// Applies UI state for a row pick. History owns the committed place — see `location(selectingProperty:)`.
    func selectProperty(_ id: String?) {
        cancelEdit()
        selectedPropertyID = id
        lockedCallout = nil
        formError = nil
    }

    /// Idempotent: re-applying the current place must not discard an in-progress identity edit.
    private func applySelection(typeKey: String?, propertyID: String?) {
        if typeKey != selectedTypeKey {
            applyType(typeKey)
        }
        if propertyID != selectedPropertyID {
            selectProperty(propertyID)
        }
    }

    private func applyType(_ key: String?) {
        selectedTypeKey = key
        lockedCallout = nil
        addPropertySelection = ""
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

    func beginEdit() {
        guard canEditSelected else { return }
        isEditingIdentity = true
        formError = nil
    }

    func cancelEdit() {
        isEditingIdentity = false
        formError = nil
    }

    @discardableResult
    func submitEdit(label: String, description: String) async -> Bool {
        guard let property = selectedProperty, canSubmitEdit(label: label, description: description) else {
            return false
        }
        let trimmedLabel = label.trimmingCharacters(in: .whitespacesAndNewlines)
        let trimmedDescription = description.trimmingCharacters(in: .whitespacesAndNewlines)
        isSaving = true
        formError = nil
        defer { isSaving = false }
        do {
            let updated = try await store.updateProperty(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                propertyID: property.id,
                label: trimmedLabel,
                valueType: property.valueType,
                description: trimmedDescription
            )
            session.apply(.updatedProperty)
            isEditingIdentity = false
            toast = VocabularyToast(
                title: String(localized: L10n.Properties.toastUpdatedTitle),
                body: L10n.Properties.toastUpdatedBody(label: updated.label, key: updated.key),
                tone: .success
            )
            return true
        } catch {
            formError = L10n.Errors.message(for: error)
            return false
        }
    }

    func askDelete() async {
        guard let property = selectedProperty else { return }
        cancelEdit()
        await deleteImpact.ask(
            kind: "property",
            id: property.id,
            ref: property.key,
            title: property.label
        ) {
            try await self.store.getDeleteImpact(
                projectDir: self.session.projectKey.projectDir,
                kind: "property",
                id: property.id
            )
        }
    }

    func cancelDelete() {
        deleteImpact.cancel()
    }

    @discardableResult
    func confirmPendingImpact() async -> Bool {
        guard let property = pendingDeleteProperty else { return false }
        return await deleteImpact.confirm { target in
            try await store.deleteProperty(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                propertyID: target.id
            )
            session.apply(.deletedProperty)
            selectedPropertyID = nil
            formError = nil
            cancelEdit()
            toast = VocabularyToast(
                title: String(localized: L10n.Properties.toastDeletedTitle),
                body: L10n.Properties.toastDeletedBody(label: property.label),
                tone: .success
            )
            syncCatalogCounts()
        }
    }

    /// ComboBox picked an unbound property — bind it to the focused type.
    /// Returns the place the caller must commit so the new inspector row survives reload.
    @discardableResult
    func addPropertyFromCombo() async -> WorkspaceLocation? {
        guard let type = selectedType,
              !addPropertySelection.isEmpty,
              let property = snapshot.properties.first(where: { $0.id == addPropertySelection })
        else { return nil }
        selectedPropertyID = property.id
        await toggleBinding(to: type)
        addPropertySelection = ""
        return location(selectingProperty: property.id)
    }

    /// Returns the created property's place for the caller to commit; nil when nothing was created.
    @discardableResult
    func submitCreate(_ incoming: Draft? = nil) async -> WorkspaceLocation? {
        guard let draft = incoming ?? draft else { return nil }
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !label.isEmpty else {
            formError = String(localized: L10n.Properties.errorLabelRequired)
            return nil
        }
        guard SubjectPropertyValueType.researcherCreatable.contains(draft.valueType) else {
            formError = String(localized: L10n.Properties.errorValueType)
            return nil
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
                try await store.assignSubjectTypeProperty(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    subjectTypeID: typeID,
                    propertyID: created.id
                )
            }
            if draft.bindTypeIDs.isEmpty {
                session.apply(.createdProperty)
            } else {
                session.apply(.mutatedSubjectTypeProperties)
            }
            closeCreate()
            selectedPropertyID = created.id
            toast = VocabularyToast(
                title: String(localized: L10n.Properties.toastCreatedTitle),
                body: L10n.Properties.toastCreatedBody(label: created.label),
                tone: .success
            )
            syncCatalogCounts()
            return WorkspaceLocation(
                section: .properties,
                subjectTypeKey: selectedTypeKey,
                propertyId: created.id,
                title: created.label
            )
        } catch {
            formError = L10n.Errors.message(for: error)
            return nil
        }
    }

    func toggleBinding(to type: CatalogSubjectType) async {
        guard let property = selectedProperty else { return }
        lockedCallout = nil
        if let existing = snapshot.binding(propertyID: property.id, typeID: type.id) {
            if existing.locked {
                lockedCallout = L10n.Properties.lockedBindingReason(typeLabel: type.label)
                return
            }
            do {
                try await store.removeSubjectTypeProperty(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    subjectTypeID: type.id,
                    propertyID: property.id
                )
                session.apply(.mutatedSubjectTypeProperties)
            } catch {
                formError = L10n.Errors.message(for: error)
            }
            return
        }
        do {
            try await store.assignSubjectTypeProperty(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectTypeID: type.id,
                propertyID: property.id
            )
            session.apply(.mutatedSubjectTypeProperties)
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
