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

    /// Where an Evidence graph stands: no Subjects, Subjects and none promoted,
    /// or promoted (at least one Subject filed by a Promote).
    enum EvidenceStatus: String, CaseIterable, Identifiable, Codable {
        case empty
        case inProgress = "in_progress"
        case promoted

        var id: String { rawValue }

        var label: LocalizedStringResource {
            switch self {
            case .empty: L10n.Sources.statusEmpty
            case .inProgress: L10n.Sources.statusInProgress
            case .promoted: L10n.Sources.statusPromoted
            }
        }

        /// A graph missing from the progress map has zero counts (Go omits it).
        init(_ progress: EvidenceGraphProgress?) {
            self = SourcesGraphCounts(progress: progress, countsLoading: false, countsError: nil).status ?? .empty
        }
    }

    struct Draft: Equatable {
        var sourceTypeID: String
        var title: String
        var description: String = ""

        var trimmedTitle: String { title.trimmingCharacters(in: .whitespacesAndNewlines) }

        /// The shared title and type rule for Add Source and Copy Source.
        var validation: SourceDraftValidation {
            SourceDraftValidation(
                titleError: trimmedTitle.isEmpty ? L10n.string(L10n.Sources.titleRequired) : nil,
                typeError: sourceTypeID.isEmpty ? L10n.string(L10n.Sources.typeRequired) : nil
            )
        }
    }

    struct SourceDraftValidation: Equatable {
        var titleError: String?
        var typeError: String?

        var isValid: Bool { titleError == nil && typeError == nil }
    }

    /// Sort, type (`source_types.id`, empty for all), evidence status, and
    /// search. Saved per project and user whenever it changes.
    var settings: SourcesListSettings {
        didSet {
            if settings != oldValue { settingsStore?.save(settings) }
        }
    }

    var isAdding = false
    /// The Copy Source form while it is open.
    var copying: SourceCopyModel?
    var draft = Draft(sourceTypeID: "", title: "")
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

    /// Starts list queries once (`.task` / dialog). Do not call from view `body`.
    func warmListQueries() {
        let _: QueryHandle<[CatalogSource]> = session.query(Self.sourcesListKey(for: session))
        let _: QueryHandle<[CatalogSourceType]> = session.query(Self.sourceTypesListKey(for: session))
        let _: QueryHandle<[String: EvidenceGraphProgress]> = session.query(session.evidenceGraphProgressKey)
        let _: QueryHandle<[CatalogEvidenceGraph]> = session.query(session.evidenceGraphsKey)
    }

    private var graphsHandle: QueryHandle<[CatalogEvidenceGraph]>? {
        session.evidenceGraphsHandle
    }

    /// True while the graph list or progress map can still arrive.
    var graphCountsLoading: Bool { session.evidenceGraphCountsLoading }

    private var typesHandle: QueryHandle<[CatalogSourceType]>? {
        session.queryHandle(Self.sourceTypesListKey(for: session))
    }

    /// The progress handle's failure while a status filter needs its counts. It
    /// reports a failure even beside a stale value, because a status filter
    /// must not apply stale counts silently.
    var statusCountsError: Error? {
        guard settings.status != nil else { return nil }
        if let failure = graphsHandle?.failure { return failure }
        return session.evidenceGraphProgressHandle?.failure
    }

    /// A status or restored type is set and the data to apply it has not arrived.
    var isAwaitingFilterData: Bool {
        if settings.status != nil, graphCountsLoading { return true }
        if !settings.typeFilterID.isEmpty {
            guard let typesHandle else { return true }
            return typesHandle.read.isLoading
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
        return sourcesHandle?.failure ?? typesHandle?.failure
    }

    // MARK: Derived

    func typeLabel(for source: CatalogSource) -> String {
        types.first { $0.id == source.sourceTypeID }?.label ?? ""
    }

    func typeIconKey(for source: CatalogSource) -> String? {
        types.iconKey(forTypeID: source.sourceTypeID)
    }

    /// A Source's title and ref keyed once. Two keys, so a match cannot cross fields.
    private struct SearchKey {
        let revision: Int64
        let title: String
        let ref: String
    }

    /// Keyed Sources by id, reused while a Source's revision and the locale hold.
    /// Not observed: filling it while the view reads `visibleSources` must not
    /// invalidate that view.
    @ObservationIgnored private var searchKeyCache: [String: SearchKey] = [:]
    @ObservationIgnored private var searchKeyLocale: String?

    private func searchKeys(for rows: [CatalogSource], locale: Locale) -> [String: SearchKey] {
        if searchKeyLocale != locale.identifier {
            searchKeyCache = [:]
            searchKeyLocale = locale.identifier
        }
        var next: [String: SearchKey] = [:]
        next.reserveCapacity(rows.count)
        for row in rows {
            if let hit = searchKeyCache[row.id], hit.revision == row.updated.revision {
                next[row.id] = hit
            } else {
                next[row.id] = SearchKey(
                    revision: row.updated.revision,
                    title: TextMatch.key(row.title, locale: locale),
                    ref: TextMatch.key(row.ref, locale: locale)
                )
            }
        }
        searchKeyCache = next
        return next
    }

    /// The statuses of the Source's graphs; empty while the Source's graph list
    /// is loading or failed.
    func statuses(for source: CatalogSource) -> Set<EvidenceStatus> {
        guard case .loaded(let graphs) = session.evidenceGraphsRead(sourceID: source.id) else { return [] }
        let progress = session.evidenceGraphProgressHandle?.value
        return Set(graphs.map { EvidenceStatus(progress?[$0.id]) })
    }

    /// Every input `visibleSources` reads: the handles' revisions, the settings
    /// and the locale. A hit means the build would return the same rows.
    private struct VisibleRowsToken: Equatable {
        let sources: Int
        let types: Int
        let settings: SourcesListSettings
        let locale: String
        /// The graph list and progress revisions, read only under a status filter.
        let graphs: Int
        let progress: Int
    }

    @ObservationIgnored private var visibleRowsToken: VisibleRowsToken?
    @ObservationIgnored private var visibleRows: [CatalogSource] = []
    /// How many times `visibleSources` built its rows (a test hook).
    @ObservationIgnored private(set) var visibleSourcesBuildCount = 0

    /// The filtered, sorted rows. Built once per change of an input; typing in
    /// Add Source or reading it again changes none. Reading the revisions here
    /// keeps the view observing them.
    var visibleSources: [CatalogSource] {
        let sourcesHandle: QueryHandle<[CatalogSource]>? = session.queryHandle(Self.sourcesListKey(for: session))
        let typesHandle: QueryHandle<[CatalogSourceType]>? = session.queryHandle(Self.sourceTypesListKey(for: session))
        let filtersByStatus = settings.status != nil
        let token = VisibleRowsToken(
            sources: sourcesHandle?.revision ?? -1,
            types: typesHandle?.revision ?? -1,
            settings: settings,
            locale: Locale.autoupdatingCurrent.identifier,
            graphs: filtersByStatus ? graphsHandle?.revision ?? -1 : -1,
            progress: filtersByStatus ? session.evidenceGraphProgressHandle?.revision ?? -1 : -1
        )
        if token == visibleRowsToken { return visibleRows }
        let rows = buildVisibleSources()
        visibleRowsToken = token
        visibleRows = rows
        visibleSourcesBuildCount += 1
        return rows
    }

    private func buildVisibleSources() -> [CatalogSource] {
        var rows = sources
        let typeID = effectiveTypeFilterID
        if !typeID.isEmpty {
            rows = rows.filter { $0.sourceTypeID == typeID }
        }
        if let status = settings.status {
            rows = rows.filter { statuses(for: $0).contains(status) }
        }
        let query = TextMatch.Query(trimmedQuery)
        if !query.isEmpty {
            let keys = searchKeys(for: sources, locale: query.locale)
            rows = rows.filter { row in
                guard let key = keys[row.id] else { return false }
                return query.matchesKey(key.title) || query.matchesKey(key.ref)
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

    var canSubmitAdd: Bool {
        !isSaving && draft.validation.isValid
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

    /// Opens the Copy Source form for a row's Source.
    func openCopy(_ source: CatalogSource) {
        copying = SourceCopyModel(original: source, session: session, userID: userID, store: store)
    }

    func openAdd() {
        typeError = nil
        titleError = nil
        createError = nil
        draft = Draft(
            sourceTypeID: types.count == 1 ? (types.first?.id ?? "") : "",
            title: ""
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
        let validation = draft.validation
        typeError = validation.typeError
        titleError = validation.titleError
        createError = nil
        guard validation.isValid else { return nil }

        isSaving = true
        defer { isSaving = false }
        do {
            let (created, effects) = try await store.createSource(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                sourceTypeID: draft.sourceTypeID,
                title: draft.trimmedTitle,
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            session.apply(effects, .createdSource(created))
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

    /// Publishes the count of a settled list to the sidebar. The view calls it
    /// through `onSettled`: a refetch keeps the cached `.ready` value, which
    /// may be stale.
    func syncCatalogCounts() {
        catalogCounts?.publishSources(sources.count)
    }
}
