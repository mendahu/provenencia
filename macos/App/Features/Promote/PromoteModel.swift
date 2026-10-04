import Foundation
import Observation

/// Frozen arrival query for the Promote place (S9-11). Parsed once from the
/// `WorkspaceLocation` the graph card pushed; fail-closed when the location is
/// not a Promote place for a primary kind.
struct PromoteEntry: Equatable, Sendable {
    let sourceID: String
    let subjectID: String
    let kind: EvidencePrimaryKind
    /// The subject's candidate ref (CPR-…).
    let subjectRef: String
    /// The subject's working label, else its ref.
    let subjectName: String
    let sourceTitle: String?

    init?(location: WorkspaceLocation) {
        guard location.section == .sources,
              location.sourceSurface == .promote,
              let sourceID = location.sourceId,
              let subjectID = location.subjectId,
              let kind = location.subjectTypeKey.flatMap(EvidencePrimaryKind.init(rawValue:))
        else { return nil }
        self.sourceID = sourceID
        self.subjectID = subjectID
        self.kind = kind
        let ref = Self.trimmed(location.ref) ?? ""
        subjectRef = ref
        subjectName = Self.trimmed(location.title) ?? ref
        sourceTitle = Self.trimmed(location.sourceTitle)
    }

    var identityKey: String { "promote|\(sourceID)|\(subjectID)" }

    /// The Evidence graph this subject came from — Back, Done, and after a save.
    var graphLocation: WorkspaceLocation {
        WorkspaceLocation(section: .sources, sourceId: sourceID, sourceSurface: .graph, title: sourceTitle)
    }

    private static func trimmed(_ text: String?) -> String? {
        guard let text = text?.trimmingCharacters(in: .whitespacesAndNewlines), !text.isEmpty else { return nil }
        return text
    }
}

extension WorkspaceLocation {
    /// The Promote place for one primary subject on a Source's Evidence graph.
    static func promote(
        sourceId: String,
        subjectId: String,
        kind: EvidencePrimaryKind,
        ref: String?,
        title: String?,
        sourceTitle: String?
    ) -> WorkspaceLocation {
        WorkspaceLocation(
            section: .sources,
            sourceId: sourceId,
            subjectId: subjectId,
            subjectTypeKey: kind.rawValue,
            sourceSurface: .promote,
            ref: ref,
            title: title,
            sourceTitle: sourceTitle
        )
    }
}

/// The Promote flow: choose where the subject goes, then file it (S9-D9).
///
/// S9-11 has one step. Next files the claim straight away (accepted, no
/// confidence or argument) and returns to the graph; S9-12 inserts the claim
/// step before the save and S9-19 the compare step on the join path. The step
/// row already shows them.
@MainActor
@Observable
final class PromoteModel {
    enum Choice: Equatable, Sendable {
        case none, new, existing
    }

    /// An existing handle the subject would join.
    struct Target: Equatable, Sendable {
        let entityID: String
        let ref: String
        let title: String
        let memberCount: Int
    }

    /// Where the subject stands on its graph. Anything but `.promotable` sends
    /// the place back to the graph.
    enum SubjectStatus: Equatable, Sendable {
        case loading, promotable, missing, alreadyPromoted
    }

    /// The navigation the leave guard is holding.
    struct PendingLeave: Identifiable, Equatable {
        let id = UUID()
        let navigation: PendingNavigation
    }

    let entry: PromoteEntry
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String
    let catalogCounts: CatalogCounts?
    weak var navigation: WorkspaceNavigation?

    private(set) var choice: Choice = .none
    private(set) var target: Target?
    private(set) var searchResults: [CatalogSearchHit] = []
    private(set) var searchFailed = false
    private(set) var isSaving = false
    private(set) var saveError: String?
    /// True once this step is filed: leaving never asks after a save.
    private(set) var saved = false
    var pendingLeave: PendingLeave?

    /// Typing pause before a search runs. Tests set it to zero.
    var searchDebounce: Duration = .milliseconds(200)
    @ObservationIgnored private var searchTask: Task<Void, Never>?

    init(
        entry: PromoteEntry,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String,
        catalogCounts: CatalogCounts?
    ) {
        self.entry = entry
        self.session = session
        self.store = store
        self.userID = userID
        self.catalogCounts = catalogCounts
    }

    var kind: EvidencePrimaryKind { entry.kind }

    var suggestionsKey: CatalogQueryKey {
        .promoteTargets(project: session.projectKey, subjectId: entry.subjectID)
    }

    var graphKey: CatalogQueryKey {
        .sourceGraph(project: session.projectKey, sourceId: entry.sourceID)
    }

    // MARK: Steps

    /// The step row: the join path has a compare step, the mint path doesn't.
    /// Until Existing is chosen the row shows the mint path.
    var steps: [LocalizedStringResource] {
        let choose = L10n.Promote.chooseStep(kind)
        return choice == .existing
            ? [choose, L10n.Promote.compareStep, L10n.Promote.claimStep]
            : [choose, L10n.Promote.claimStep]
    }

    var stepText: String { L10n.Promote.stepOf(current: 1, total: steps.count) }

    /// Next is live once the choice is complete: New, or Existing with a handle.
    var canAdvance: Bool {
        guard !isSaving, !saved else { return false }
        switch choice {
        case .new: return true
        case .existing: return target != nil
        case .none: return false
        }
    }

    /// The footer hint. New and Existing use interim wording while Next files
    /// the claim directly (S9-12 / S9-19 restore the board's copy).
    var hint: String {
        switch choice {
        case .new:
            return L10n.string(L10n.Promote.hintNew(kind))
        case .existing:
            guard let target else { return L10n.string(L10n.Promote.hintChoose(kind)) }
            return L10n.Promote.hintExisting(
                name: entry.subjectName,
                ref: target.ref,
                members: Self.members(target.memberCount)
            )
        case .none:
            return L10n.string(L10n.Promote.hintChoose(kind))
        }
    }

    static func members(_ count: Int) -> String {
        count == 1 ? L10n.Promote.memberOne(count: count) : L10n.Promote.memberOther(count: count)
    }

    // MARK: Subject

    func subjectStatus(rows: SourceGraphRows?) -> SubjectStatus {
        guard let rows else { return .loading }
        guard rows.subjects.contains(where: { $0.id == entry.subjectID }) else { return .missing }
        if !saved, rows.memberships.contains(where: { $0.subjectID == entry.subjectID }) {
            return .alreadyPromoted
        }
        return .promotable
    }

    /// The subject left the graph or was promoted elsewhere: back to the graph.
    func returnToGraph() {
        saved = true
        navigation?.go(to: entry.graphLocation)
    }

    // MARK: Choosing

    func choose(_ newChoice: Choice) {
        guard choice != newChoice else { return }
        choice = newChoice
        saveError = nil
        if newChoice != .existing {
            target = nil
        }
    }

    func select(_ newTarget: Target) {
        choice = .existing
        target = newTarget
        saveError = nil
    }

    static func target(for suggestion: CatalogPromoteTargetSuggestion) -> Target {
        Target(
            entityID: suggestion.entity.id,
            ref: suggestion.entity.ref,
            title: title(for: suggestion),
            memberCount: suggestion.memberCount
        )
    }

    /// Name for Persons (name → label → ref); label, else ref, for the rest.
    static func title(for suggestion: CatalogPromoteTargetSuggestion) -> String {
        if let person = suggestion.person {
            return PersonHeaderDisplay.title(person)
        }
        let label = suggestion.entity.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return label.isEmpty ? suggestion.entity.ref : label
    }

    // MARK: Search

    /// Search results as ComboBox options, keeping the chosen handle among
    /// them so the field goes on showing its name.
    var searchOptions: [PVComboBoxOption] {
        var options = searchResults.map {
            PVComboBoxOption(value: $0.id, label: $0.title, subtext: $0.ref)
        }
        if let target, !options.contains(where: { $0.value == target.entityID }) {
            options.insert(PVComboBoxOption(value: target.entityID, label: target.title, subtext: target.ref), at: 0)
        }
        return options
    }

    /// The handle the search field shows as chosen, if it came from search or
    /// is a suggestion.
    var searchSelection: String { target?.entityID ?? "" }

    func updateQuery(_ query: String) {
        searchTask?.cancel()
        let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else {
            searchResults = []
            searchFailed = false
            return
        }
        let debounce = searchDebounce
        searchTask = Task { [weak self] in
            if debounce > .zero {
                try? await Task.sleep(for: debounce)
            }
            guard !Task.isCancelled else { return }
            await self?.runSearch(trimmed)
        }
    }

    func runSearch(_ query: String) async {
        do {
            let hits = try await store.searchCatalog(
                projectDir: session.projectKey.projectDir,
                query: query,
                location: entry.graphLocation,
                kinds: [kind.rawValue]
            )
            guard !Task.isCancelled else { return }
            searchResults = hits
            searchFailed = false
        } catch {
            guard !Task.isCancelled else { return }
            searchResults = []
            searchFailed = true
        }
    }

    /// The ComboBox committed a value.
    func selectSearchResult(_ entityID: String) {
        guard !entityID.isEmpty else { return }
        if let hit = searchResults.first(where: { $0.id == entityID }) {
            select(Target(entityID: hit.id, ref: hit.ref, title: hit.title, memberCount: hit.memberCount))
        }
    }

    // MARK: Saving

    /// Files the claim — onto a new handle, or the chosen one — then returns
    /// to the graph, which redraws with the subject filed. A failure keeps the
    /// step and shows why.
    @discardableResult
    func next() async -> Bool {
        guard canAdvance else { return false }
        isSaving = true
        defer { isSaving = false }
        do {
            _ = try await store.promoteSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectID: entry.subjectID,
                entityID: choice == .existing ? target?.entityID : nil,
                confidenceGradeID: nil,
                argument: ""
            )
        } catch {
            saveError = L10n.Errors.message(for: error)
            return false
        }
        saved = true
        saveError = nil
        session.apply(.promotedSubject(sourceId: entry.sourceID))
        await catalogCounts?.refreshAll()
        navigation?.go(to: entry.graphLocation)
        return true
    }

    /// Done leaves the flow; the leave guard asks first if a choice is unsaved.
    func done() {
        navigation?.go(to: entry.graphLocation)
    }

    // MARK: Leave guard

    var hasUnsavedChoice: Bool { !saved && choice != .none }

    var leaveTitle: String { L10n.Promote.leaveTitle(name: entry.subjectName) }

    var leaveMessage: String {
        switch choice {
        case .new:
            return L10n.Promote.leaveMessageNew(kind, subjectRef: entry.subjectRef)
        case .existing:
            if let target {
                return L10n.Promote.leaveMessageExisting(target: target.ref, subjectRef: entry.subjectRef)
            }
            return L10n.Promote.leaveMessageUnchosen(subjectRef: entry.subjectRef)
        case .none:
            return L10n.Promote.leaveMessageUnchosen(subjectRef: entry.subjectRef)
        }
    }

    func leave() {
        pendingLeave = nil
        saved = true
        navigation?.resumeHeldNavigation()
    }

    func keepPromoting() {
        pendingLeave = nil
        navigation?.cancelHeldNavigation()
    }
}

extension PromoteModel: WorkspaceLeaveGuard {
    func shouldHoldNavigation(_ pending: PendingNavigation) -> Bool {
        guard hasUnsavedChoice else { return false }
        pendingLeave = PendingLeave(navigation: pending)
        return true
    }
}
