import Foundation
import Observation

/// State for the **Sources** workspace destination (S2-04 board / S2-17 PR):
/// browse/filter/sort Sources as an evidence list and create via a thin dialog.
/// List rows come from `WorkspaceSession`; navigation to detail is history-driven.
@MainActor
@Observable
final class SourcesModel {
    enum Sort: String, CaseIterable, Identifiable, Codable {
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

    /// Where a Source's Evidence graph stands, from its primary and filed counts.
    enum EvidenceStatus: String, CaseIterable, Identifiable, Codable {
        case empty
        case inProgress = "in_progress"
        case filed

        var id: String { rawValue }

        var label: LocalizedStringResource {
            switch self {
            case .empty: L10n.Sources.statusEmpty
            case .inProgress: L10n.Sources.statusInProgress
            case .filed: L10n.Sources.statusFullyPromoted
            }
        }

        /// A Source missing from the progress map has zero counts (Go omits it).
        init(_ progress: SourceGraphProgress?) {
            guard let progress, progress.primaryCount > 0 else {
                self = .empty
                return
            }
            self = progress.filedCount < progress.primaryCount ? .inProgress : .filed
        }
    }

    struct Draft: Equatable {
        var sourceTypeID: String
        var title: String
        var description: String
    }

    /// Sort, type (`source_types.id`, empty for all), evidence status, and
    /// search. Saved per project and user whenever it changes.
    var settings: SourcesListSettings {
        didSet {
            if settings != oldValue { settingsStore?.save(settings) }
        }
    }

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
    private let settingsStore: SourcesListSettingsStore?

    init(
        session: WorkspaceSession,
        userID: String,
        store: any GenealogyStore,
        catalogCounts: CatalogCounts? = nil,
        settingsStore: SourcesListSettingsStore? = nil
    ) {
        self.session = session
        self.userID = userID
        self.store = store
        self.catalogCounts = catalogCounts
        self.settingsStore = settingsStore
        self.settings = settingsStore?.load() ?? SourcesListSettings()
    }

    // MARK: Session handles

    static func sourcesListKey(for session: WorkspaceSession) -> CatalogQueryKey {
        CatalogQueryKey.sourcesList(project: session.projectKey)
    }

    static func sourceTypesListKey(for session: WorkspaceSession) -> CatalogQueryKey {
        CatalogQueryKey.sourceTypesList(project: session.projectKey)
    }

    static func sourceGraphProgressKey(for session: WorkspaceSession) -> CatalogQueryKey {
        CatalogQueryKey.sourceGraphProgress(project: session.projectKey)
    }

    /// Starts list queries once (`.task` / dialog). Do not call from view `body`.
    func warmListQueries() {
        let _: QueryHandle<[CatalogSource]> = session.query(Self.sourcesListKey(for: session))
        let _: QueryHandle<[CatalogSourceType]> = session.query(Self.sourceTypesListKey(for: session))
        let _: QueryHandle<[String: SourceGraphProgress]> = session.query(
            Self.sourceGraphProgressKey(for: session)
        )
    }

    func graphProgress(for sourceID: String) -> SourceGraphProgress? {
        progressHandle?.value?[sourceID]
    }

    var graphCountsLoading: Bool {
        guard let handle = progressHandle else { return true }
        return handle.value == nil && handle.status != .ready
    }

    private var typesHandle: QueryHandle<[CatalogSourceType]>? {
        session.queryHandle(Self.sourceTypesListKey(for: session))
    }

    private var progressHandle: QueryHandle<[String: SourceGraphProgress]>? {
        session.queryHandle(Self.sourceGraphProgressKey(for: session))
    }

    /// The progress handle's failure while a status filter needs its counts.
    var statusCountsError: Error? {
        guard settings.status != nil else { return nil }
        guard let handle = progressHandle, handle.status == .error else { return nil }
        return handle.error
    }

    /// A status or restored type is set and the data to apply it has not arrived.
    var isAwaitingFilterData: Bool {
        if settings.status != nil, graphCountsLoading { return true }
        if !settings.typeFilterID.isEmpty {
            guard let typesHandle else { return true }
            return typesHandle.value == nil && typesHandle.status != .error
        }
        return false
    }

    /// The saved type id when it is still a Source type, else all types. The
    /// stored id is left alone until the researcher changes the type.
    var effectiveTypeFilterID: String {
        types.contains { $0.id == settings.typeFilterID } ? settings.typeFilterID : ""
    }

    var isFiltering: Bool {
        !effectiveTypeFilterID.isEmpty
            || settings.status != nil
            || !trimmedQuery.isEmpty
    }

    private var trimmedQuery: String {
        settings.query.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    func clearFilters() {
        var next = settings
        next.typeFilterID = ""
        next.status = nil
        next.query = ""
        settings = next
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
        let typeID = effectiveTypeFilterID
        if !typeID.isEmpty {
            rows = rows.filter { $0.sourceTypeID == typeID }
        }
        if let status = settings.status {
            let progress = progressHandle?.value
            rows = rows.filter { EvidenceStatus(progress?[$0.id]) == status }
        }
        let query = trimmedQuery
        if !query.isEmpty {
            rows = rows.filter {
                $0.title.localizedStandardContains(query) || $0.ref.localizedStandardContains(query)
            }
        }
        switch settings.sort {
        case .added:
            // UUIDv7 string form preserves create-time byte order (newest first).
            rows = rows.sorted { $0.id > $1.id }
        case .updated:
            rows = rows.sorted {
                if $0.updated.revision != $1.updated.revision {
                    return $0.updated.revision > $1.updated.revision
                }
                return $0.id > $1.id
            }
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
        let typeID = effectiveTypeFilterID
        if typeID.isEmpty {
            return L10n.string(L10n.Sources.filterAllTypes)
        }
        return types.first { $0.id == typeID }?.label
            ?? L10n.string(L10n.Sources.filterAllTypes)
    }

    var sortControlLabel: String {
        L10n.Sources.sortedBy(L10n.string(settings.sort.label).lowercased())
    }

    func countLine(visibleCount visible: Int) -> String {
        if isAwaitingFilterData || statusCountsError != nil {
            return L10n.Sources.countLine(total: sources.count)
        }
        if !isFiltering || visible == sources.count {
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

    /// The sole auto-selected type is a seed, not an edit. A typed name,
    /// description, or a type the researcher picked counts.
    var addSheetIsDirty: Bool {
        let title = draft.title.trimmingCharacters(in: .whitespacesAndNewlines)
        let description = draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
        if !title.isEmpty || !description.isEmpty { return true }
        if types.count == 1 { return false }
        return !draft.sourceTypeID.isEmpty
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
            typeErr = L10n.string(L10n.Sources.typeRequired)
        }
        if trimmedTitle.isEmpty {
            titleErr = L10n.string(L10n.Sources.titleRequired)
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
