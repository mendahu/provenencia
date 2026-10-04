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

/// Runs the Promote place. `flow` (`PromoteFlow`) is the single source of
/// truth for progress, the draft, and what is in flight; this model sends it
/// events and performs the effects it returns — the write, the refresh after
/// it, and navigation. It also holds the picker's search results, which are
/// transient view data rather than flow progress.
@MainActor
@Observable
final class PromoteModel {
    typealias Choice = PromoteFlow.Choice
    typealias Target = PromoteFlow.Target

    /// The guard's question for the confirm sheet, stable while it is open.
    struct PendingLeave: Identifiable, Equatable {
        let navigation: PendingNavigation
        var id: String { String(describing: navigation) }
    }

    /// Where the current subject stands on its graph.
    enum SubjectStatus: Equatable, Sendable {
        case loading, promotable, missing, alreadyPromoted
    }

    let entry: PromoteEntry
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String
    let catalogCounts: CatalogCounts?
    weak var navigation: WorkspaceNavigation?

    private(set) var flow: PromoteFlow
    private(set) var searchResults: [CatalogSearchHit] = []
    private(set) var searchFailed = false

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
        flow = PromoteFlow(subject: PromoteFlow.Subject(
            id: entry.subjectID,
            ref: entry.subjectRef,
            name: entry.subjectName,
            kind: entry.kind
        ))
    }

    // MARK: Reading the flow

    var subject: PromoteFlow.Subject { flow.subject }
    var kind: EvidencePrimaryKind { flow.subject.kind }
    var choice: Choice { flow.draft.choice }
    var target: Target? { flow.draft.target }
    var canAdvance: Bool { flow.canAdvance }
    var isSaving: Bool { flow.isSaving }
    var saveError: String? { flow.error }
    var hasUnsavedWork: Bool { flow.hasUnsavedWork }
    var pendingLeave: PendingLeave? { flow.pendingLeave.map(PendingLeave.init(navigation:)) }

    var suggestionsKey: CatalogQueryKey {
        .promoteTargets(project: session.projectKey, subjectId: flow.subject.id)
    }

    var graphKey: CatalogQueryKey {
        .sourceGraph(project: session.projectKey, sourceId: entry.sourceID)
    }

    /// The step row: the plan as designed (later steps show before they are
    /// built), and where the flow is in it.
    var steps: [LocalizedStringResource] { flow.plan.map(label(for:)) }
    var currentStepIndex: Int { flow.stepNumber - 1 }
    var stepText: String { L10n.Promote.stepOf(current: flow.stepNumber, total: flow.plan.count) }

    private func label(for step: PromoteStep) -> LocalizedStringResource {
        switch step {
        case .chooseTarget: L10n.Promote.chooseStep(kind)
        case .compare: L10n.Promote.compareStep
        case .claim: L10n.Promote.claimStep
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
                name: subject.name,
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

    var leaveTitle: String { L10n.Promote.leaveTitle(name: subject.name) }

    var leaveMessage: String {
        switch choice {
        case .new:
            return L10n.Promote.leaveMessageNew(kind, subjectRef: subject.ref)
        case .existing:
            if let target {
                return L10n.Promote.leaveMessageExisting(target: target.ref, subjectRef: subject.ref)
            }
            return L10n.Promote.leaveMessageUnchosen(subjectRef: subject.ref)
        case .none:
            return L10n.Promote.leaveMessageUnchosen(subjectRef: subject.ref)
        }
    }

    // MARK: Events

    func choose(_ choice: Choice) { send(.choose(choice)) }
    func select(_ target: Target) { send(.selectTarget(target)) }
    func done() { send(.done) }
    func leave() { send(.leaveConfirmed) }
    func keepPromoting() { send(.leaveCancelled) }

    /// Next: the following step, or the write. Returns once any write and its
    /// follow-up have run; true when the flow moved on.
    @discardableResult
    func next() async -> Bool {
        let before = flow
        await perform(flow.send(.advance))
        return flow.savedCount > before.savedCount || flow.step != before.step
    }

    /// Feeds the graph's view of the current subject to the flow.
    func subjectStatusChanged(_ status: SubjectStatus) {
        if status == .missing || status == .alreadyPromoted {
            send(.subjectUnavailable)
        }
    }

    func subjectStatus(rows: SourceGraphRows?) -> SubjectStatus {
        guard let rows else { return .loading }
        let id = flow.subject.id
        guard rows.subjects.contains(where: { $0.id == id }) else { return .missing }
        if rows.memberships.contains(where: { $0.subjectID == id }) { return .alreadyPromoted }
        return .promotable
    }

    /// Sends an event from the view. Navigation effects run in the same turn
    /// so the place reacts at once; an event that starts a write runs its
    /// effects in order on a task.
    private func send(_ event: PromoteFlow.Event) {
        let effects = flow.send(event)
        if effects.contains(where: \.isAsync) {
            Task { await perform(effects) }
        } else {
            effects.forEach(runNavigation)
        }
    }

    /// Runs effects in order, awaiting the write and the refresh after it.
    private func perform(_ effects: [PromoteFlow.Effect]) async {
        for effect in effects {
            switch effect {
            case .save(let save):
                await write(save)
            case .refreshAfterSave:
                session.apply(.promotedSubject(sourceId: entry.sourceID))
                await catalogCounts?.refreshAll()
            case .navigateToGraph, .resumeNavigation, .cancelNavigation:
                runNavigation(effect)
            }
        }
    }

    private func runNavigation(_ effect: PromoteFlow.Effect) {
        switch effect {
        case .navigateToGraph: navigation?.go(to: entry.graphLocation)
        case .resumeNavigation: navigation?.resumeHeldNavigation()
        case .cancelNavigation: navigation?.cancelHeldNavigation()
        case .save, .refreshAfterSave: break
        }
    }

    private func write(_ save: PromoteFlow.Save) async {
        do {
            _ = try await store.promoteSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectID: save.subjectID,
                entityID: save.entityID,
                confidenceGradeID: save.confidenceGradeID,
                argument: save.argument
            )
        } catch {
            await perform(flow.send(.saveFailed(message: L10n.Errors.message(for: error))))
            return
        }
        await perform(flow.send(.saveSucceeded))
    }

    // MARK: Targets

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
        guard !entityID.isEmpty,
              let hit = searchResults.first(where: { $0.id == entityID })
        else { return }
        select(Target(entityID: hit.id, ref: hit.ref, title: hit.title, memberCount: hit.memberCount))
    }
}

extension PromoteModel: WorkspaceLeaveGuard {
    func shouldHoldNavigation(_ pending: PendingNavigation) -> Bool {
        flow.requestLeave(pending) == .hold
    }
}

private extension PromoteFlow.Effect {
    var isAsync: Bool {
        switch self {
        case .save, .refreshAfterSave: true
        case .navigateToGraph, .resumeNavigation, .cancelNavigation: false
        }
    }
}
