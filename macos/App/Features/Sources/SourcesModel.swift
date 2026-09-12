import Foundation
import Observation

/// State for the **Sources** workspace destination (S2-04 board / S2-17 PR):
/// browse/search/filter/sort Sources as an evidence list, create via a thin
/// dialog, and open a separate Source page (S2-18).
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

    private(set) var sources: [CatalogSource] = []
    private(set) var types: [CatalogSourceType] = []
    private(set) var isLoading = false
    var loadError: Error?

    var query = ""
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

    /// Non-nil when the Source page is showing that id.
    private(set) var openedSourceID: String?

    /// Exposed so `SourcesView` can mount `SourcePageView` without duplicating init.
    var pageProjectDir: String { projectDir }
    var pageUserID: String { userID }
    var pageStore: any GenealogyStore { store }

    private let projectDir: String
    private let userID: String
    private let store: any GenealogyStore
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

    func typeLabel(for source: CatalogSource) -> String {
        types.first { $0.id == source.sourceTypeID }?.label ?? ""
    }

    var visibleSources: [CatalogSource] {
        var rows = sources
        if !typeFilterID.isEmpty {
            rows = rows.filter { $0.sourceTypeID == typeFilterID }
        }
        let q = query.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if !q.isEmpty {
            rows = rows.filter { source in
                source.title.lowercased().contains(q)
                    || source.ref.lowercased().contains(q)
                    || typeLabel(for: source).lowercased().contains(q)
            }
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

    var openedSource: CatalogSource? {
        guard let openedSourceID else { return nil }
        return sources.first { $0.id == openedSourceID }
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

    func load() async {
        isLoading = true
        loadError = nil
        defer { isLoading = false }
        // Load independently so a busy catalog on one call does not leave
        // the type pool empty for Add Source when sources already arrived
        // (or vice versa).
        var firstError: Error?
        do {
            sources = try await store.listSources(projectDir: projectDir)
            publishCounts()
        } catch {
            firstError = error
        }
        do {
            types = try await store.listSourceTypes(projectDir: projectDir)
        } catch {
            firstError = firstError ?? error
        }
        loadError = firstError
    }

    /// Refreshes the type vocabulary for the Add Source combo. Always hits
    /// the store so a failed/racy initial `load` cannot leave the dialog
    /// with an empty pool while Source types elsewhere show rows.
    func refreshTypes() async {
        do {
            types = try await store.listSourceTypes(projectDir: projectDir)
            if loadError != nil, !types.isEmpty {
                loadError = nil
            }
        } catch {
            loadError = error
        }
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

    func create() async {
        guard !isSaving else { return }
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
        guard typeErr == nil, titleErr == nil else { return }

        isSaving = true
        defer { isSaving = false }
        do {
            let created = try await store.createSource(
                projectDir: projectDir,
                userID: userID,
                sourceTypeID: draft.sourceTypeID,
                title: trimmedTitle,
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            sources.insert(created, at: 0)
            publishCounts()
            isAdding = false
            typeError = nil
            titleError = nil
            createError = nil
            toast = VocabularyToast(
                title: L10n.Sources.toastCreatedTitle(ref: created.ref),
                body: L10n.Sources.toastCreatedBody(title: created.title),
                tone: .success
            )
            openedSourceID = created.id
        } catch {
            createError = L10n.Errors.message(for: error)
        }
    }

    func openSource(id: String) {
        openedSourceID = id
    }

    func closeSource() {
        openedSourceID = nil
    }

    /// Keeps the list row in sync when the Source page edits identity or cover.
    /// Identity `updateSource` responses omit thumbnail enrichment — preserve
    /// existing cover fields when the incoming row leaves them empty.
    func applyUpdatedSource(_ source: CatalogSource) {
        if let idx = sources.firstIndex(where: { $0.id == source.id }) {
            var merged = source
            let incomingHasCover = !merged.thumbnailRelPath.isEmpty
                || !merged.thumbnailMediaType.isEmpty
                || !merged.thumbnailOriginalFilename.isEmpty
            if !incomingHasCover {
                merged.thumbnailRelPath = sources[idx].thumbnailRelPath
                merged.thumbnailMediaType = sources[idx].thumbnailMediaType
                merged.thumbnailOriginalFilename = sources[idx].thumbnailOriginalFilename
            }
            sources[idx] = merged
        }
    }

    private func publishCounts() {
        catalogCounts?.publishSources(sources.count)
    }
}
