import Foundation
import Observation

/// State for the **Sources** workspace destination (S2-04 board / S2-17 PR):
/// browse/filter/sort Sources as an evidence list and create via a thin dialog.
/// List rows come from `WorkspaceSession`; navigation to detail is history-driven.
@MainActor
@Observable
final class SourcesModel {
    enum Sort: String, CaseIterable, Identifiable {
        case added
        case updated
        case az
        case za

        var id: String { rawValue }

        var label: LocalizedStringResource {
            switch self {
            case .added: L10n.Sources.sortAdded
            case .updated: L10n.Sources.sortUpdated
            case .az: L10n.Sources.sortAZ
            case .za: L10n.Sources.sortZA
            }
        }
    }

    struct Draft: Equatable {
        var sourceTypeID: String
        var title: String
        var description: String
    }

    /// `nil` / empty string means all types. Otherwise a `source_types.id`.
    var typeFilterID = ""
    var sort: Sort = .added

    var isAdding = false
    var draft = Draft(sourceTypeID: "", title: "", description: "")
    var typeError: String?
    var titleError: String?
    /// FFI / store failure on Create Source (not field validation).
    var createError: String?
    private(set) var isSaving = false
    var toast: VocabularyToast?

    var pageProjectDir: String { session.projectKey.projectDir }

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

    static func sourcesListKey(for session: WorkspaceSession) -> CatalogQueryKey {
        CatalogQueryKey.sourcesList(project: session.projectKey)
    }

    static func sourceTypesListKey(for session: WorkspaceSession) -> CatalogQueryKey {
        CatalogQueryKey.sourceTypesList(project: session.projectKey)
    }

    /// Starts list queries once (`.task` / dialog). Do not call from view `body`.
    func warmListQueries() {
        let _: QueryHandle<[CatalogSource]> = session.query(Self.sourcesListKey(for: session))
        let _: QueryHandle<[CatalogSourceType]> = session.query(Self.sourceTypesListKey(for: session))
    }

    var sources: [CatalogSource] {
        session.queryHandle(Self.sourcesListKey(for: session))?.value ?? []
    }

    var types: [CatalogSourceType] {
        session.queryHandle(Self.sourceTypesListKey(for: session))?.value ?? []
    }

    var isLoading: Bool {
        guard let handle: QueryHandle<[CatalogSource]> = session.queryHandle(Self.sourcesListKey(for: session))
        else { return true }
        return handle.status == .loading && sources.isEmpty
    }

    var loadError: Error? {
        guard sources.isEmpty else { return nil }
        let sourcesHandle: QueryHandle<[CatalogSource]>? = session.queryHandle(Self.sourcesListKey(for: session))
        let typesHandle: QueryHandle<[CatalogSourceType]>? = session.queryHandle(Self.sourceTypesListKey(for: session))
        return sourcesHandle?.error ?? typesHandle?.error
    }

    // MARK: Derived

    func typeLabel(for source: CatalogSource) -> String {
        types.first { $0.id == source.sourceTypeID }?.label ?? ""
    }

    func typeIconKey(for source: CatalogSource) -> String? {
        types.first { $0.id == source.sourceTypeID }?.iconKey
    }

    var visibleSources: [CatalogSource] {
        var rows = sources
        if !typeFilterID.isEmpty {
            rows = rows.filter { $0.sourceTypeID == typeFilterID }
        }
        switch sort {
        case .added:
            break
        case .updated:
            rows = rows.reversed()
        case .az:
            rows = rows.sorted {
                $0.title.localizedCaseInsensitiveCompare($1.title) == .orderedAscending
            }
        case .za:
            rows = rows.sorted {
                $0.title.localizedCaseInsensitiveCompare($1.title) == .orderedDescending
            }
        }
        return rows
    }

    var isCatalogEmpty: Bool { sources.isEmpty }

    var filterLabel: String {
        if typeFilterID.isEmpty {
            return String(localized: L10n.Sources.filterAllTypes)
        }
        return types.first { $0.id == typeFilterID }?.label
            ?? String(localized: L10n.Sources.filterAllTypes)
    }

    var sortControlLabel: String {
        L10n.Sources.sortedBy(String(localized: sort.label).lowercased())
    }

    var countLine: String {
        let visible = visibleSources.count
        if visible == sources.count {
            return L10n.Sources.countLine(total: sources.count)
        }
        return L10n.Sources.countLineFiltered(visible: visible, total: sources.count)
    }

    var typeComboOptions: [PVComboBoxOption] {
        types.map { PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key) }
    }

    var canSubmitAdd: Bool {
        !isSaving
            && !draft.sourceTypeID.isEmpty
            && !draft.title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    // MARK: Actions

    /// Ensures the type pool query is warm for the Add Source dialog.
    func refreshTypes() {
        let _: QueryHandle<[CatalogSourceType]> = session.query(Self.sourceTypesListKey(for: session))
    }

    func openAdd() {
        typeError = nil
        titleError = nil
        createError = nil
        draft = Draft(
            sourceTypeID: types.count == 1 ? (types.first?.id ?? "") : "",
            title: "",
            description: ""
        )
        isAdding = true
    }

    /// After a late type refresh, commit the only option so Create is one less step.
    func selectSoleTypeIfNeeded() {
        guard types.count == 1, draft.sourceTypeID.isEmpty else { return }
        draft.sourceTypeID = types[0].id
    }

    func cancelAdd() {
        guard !isSaving else { return }
        isAdding = false
        typeError = nil
        titleError = nil
        createError = nil
    }

    /// On success returns the created source for the view to `go(to:)`.
    @discardableResult
    func create() async -> CatalogSource? {
        guard !isSaving else { return nil }
        let trimmedTitle = draft.title.trimmingCharacters(in: .whitespacesAndNewlines)
        var typeErr: String?
        var titleErr: String?
        if draft.sourceTypeID.isEmpty {
            typeErr = String(localized: L10n.Sources.typeRequired)
        }
        if trimmedTitle.isEmpty {
            titleErr = String(localized: L10n.Sources.titleRequired)
        }
        typeError = typeErr
        titleError = titleErr
        createError = nil
        guard typeErr == nil, titleErr == nil else { return nil }

        isSaving = true
        defer { isSaving = false }
        do {
            let created = try await store.createSource(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                sourceTypeID: draft.sourceTypeID,
                title: trimmedTitle,
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            session.apply(.createdSource)
            let listKey = CatalogQueryKey.sourcesList(project: session.projectKey)
            if let listHandle: QueryHandle<[CatalogSource]> = session.queryHandle(listKey) {
                var rows = listHandle.value ?? []
                rows.insert(created, at: 0)
                session.setQueryValue(listKey, value: rows)
            }
            syncCatalogCounts()
            isAdding = false
            typeError = nil
            titleError = nil
            createError = nil
            toast = VocabularyToast(
                title: L10n.Sources.toastCreatedTitle(ref: created.ref),
                body: L10n.Sources.toastCreatedBody(title: created.title),
                tone: .success
            )
            return created
        } catch {
            createError = L10n.Errors.message(for: error)
            return nil
        }
    }

    /// Refreshes sidebar totals from the cached list (after warm load or create).
    func syncCatalogCounts() {
        catalogCounts?.publishSources(sources.count)
    }
}
