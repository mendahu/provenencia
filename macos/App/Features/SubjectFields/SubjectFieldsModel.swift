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

enum SubjectFieldsOriginFilter: String, CaseIterable, Identifiable {
    case all
    case seeded
    case user

    var id: String { rawValue }

    var label: LocalizedStringResource {
        switch self {
        case .all: return L10n.SubjectFields.originFilterAll
        case .seeded: return L10n.SubjectFields.originFilterSeeded
        case .user: return L10n.SubjectFields.originFilterUser
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
    }

    private(set) var selectedTypeKey: String?
    var searchQuery = ""
    private(set) var originFilter: SubjectFieldsOriginFilter = .all
    private(set) var valueTypeFilter: String?
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

    var types: [CatalogSubjectType] {
        snapshot.types.sorted { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }
    }

    var selectedType: CatalogSubjectType? {
        guard let selectedTypeKey else { return nil }
        return snapshot.types.first { $0.key == selectedTypeKey }
    }

    var visibleProperties: [CatalogProperty] {
        let q = searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        return snapshot.properties
            .filter { property in
                if let selectedTypeKey,
                   let type = snapshot.types.first(where: { $0.key == selectedTypeKey }) {
                    let bound = snapshot.fieldsByTypeID[type.id]?.contains { $0.property.id == property.id } == true
                    if !bound { return false }
                }
                switch originFilter {
                case .all: break
                case .seeded:
                    if property.origin != CatalogOrigin.provenencia { return false }
                case .user:
                    if property.origin != CatalogOrigin.user { return false }
                }
                if let valueTypeFilter, property.valueType != valueTypeFilter { return false }
                if !q.isEmpty {
                    let hay = (property.label + " " + property.key).lowercased()
                    if !hay.contains(q) { return false }
                }
                return true
            }
            .sorted { $0.label.localizedCaseInsensitiveCompare($1.label) == .orderedAscending }
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
        selectedTypeKey = key
        lockedCallout = nil
    }

    func setOriginFilter(_ filter: SubjectFieldsOriginFilter) {
        originFilter = filter
    }

    func setValueTypeFilter(_ valueType: String?) {
        valueTypeFilter = valueType
    }

    func selectProperty(_ id: String?) {
        selectedPropertyID = id
        lockedCallout = nil
        formError = nil
    }

    func openCreate() {
        formError = nil
        lockedCallout = nil
        createOpen = true
        draft = Draft(label: "", valueType: "text", description: "")
    }

    func closeCreate() {
        createOpen = false
        draft = nil
        formError = nil
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
            warmWorkspaceQuery()
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
            if let type = selectedType {
                try await store.assignSubjectTypeField(
                    projectDir: session.projectKey.projectDir,
                    userID: userID,
                    subjectTypeID: type.id,
                    propertyID: created.id
                )
                session.apply(.mutatedSubjectTypeFields)
            } else {
                session.apply(.createdProperty)
            }
            warmWorkspaceQuery()
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

    /// Toggles binding for the selected property to `type`. Locked bindings show a callout.
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
                warmWorkspaceQuery()
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
            warmWorkspaceQuery()
        } catch {
            formError = L10n.Errors.message(for: error)
        }
    }

    func bindingLocked(propertyID: String, typeID: String) -> Bool {
        snapshot.binding(propertyID: propertyID, typeID: typeID)?.locked == true
    }

    func isBound(propertyID: String, typeID: String) -> Bool {
        snapshot.binding(propertyID: propertyID, typeID: typeID) != nil
    }
}
