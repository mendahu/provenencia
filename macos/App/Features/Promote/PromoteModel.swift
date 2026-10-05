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
        case loading, promotable, missing
        /// Filed on this handle (its ref).
        case alreadyPromoted(onto: String)
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
    var controls: PromoteFlow.Controls { flow.controls }
    var canAdvance: Bool { flow.canAdvance }
    var isSaving: Bool { flow.isSaving }
    var saveError: String? { flow.error }
    var isBlocked: Bool { flow.isBlocked }
    var hasUnsavedWork: Bool { flow.hasUnsavedWork }
    var pendingLeave: PendingLeave? { flow.pendingLeave.map(PendingLeave.init(navigation:)) }

    var suggestionsKey: CatalogQueryKey {
        .promoteTargets(project: session.projectKey, subjectId: flow.subject.id)
    }

    var graphKey: CatalogQueryKey {
        .sourceGraph(project: session.projectKey, sourceId: entry.sourceID)
    }

    var confidenceKey: CatalogQueryKey {
        .confidenceGradesList(project: session.projectKey)
    }

    /// Subject types (their handle ref prefixes) come with the Properties snapshot.
    var propertiesKey: CatalogQueryKey {
        .propertiesWorkspace(project: session.projectKey)
    }

    /// The step row: the plan as designed (later steps show before they are
    /// built), and where the flow is in it.
    var steps: [LocalizedStringResource] { flow.plan.map(label(for:)) }
    var currentStepIndex: Int { flow.stepNumber - 1 }
    var stepText: String { L10n.Promote.stepOf(current: flow.stepNumber, total: flow.plan.count) }

    func label(for step: PromoteStep) -> LocalizedStringResource {
        switch step {
        case .chooseTarget: L10n.Promote.chooseStep(kind)
        case .compare: L10n.Promote.compareStep
        case .claim: L10n.Promote.claimStep
        }
    }

    /// The footer hint for where the flow is.
    var hint: String {
        let name = subject.name
        switch flow.phase {
        case .saving:
            return L10n.Promote.hintSaving(name: name)
        case .blocked:
            return L10n.string(L10n.Promote.hintBlocked)
        case .editing, .confirmingLeave, .finished:
            break
        }
        if flow.step == .claim {
            if let target, choice == .existing {
                return L10n.Promote.hintSaveExisting(name: name, ref: target.ref)
            }
            return L10n.Promote.hintSaveNew(kind, name: name)
        }
        switch choice {
        case .new:
            return L10n.string(L10n.Promote.hintNew(kind))
        case .existing:
            guard let target else { return L10n.string(L10n.Promote.hintChoose(kind)) }
            // Interim until Compare (S9-19) restores the board's wording.
            return L10n.Promote.hintExisting(
                name: name,
                ref: target.ref,
                members: Self.members(target.memberCount)
            )
        case .none:
            return L10n.string(L10n.Promote.hintChoose(kind))
        }
    }

    /// The footer's Back: "Back to {previous step}".
    var backLabel: String? {
        controls.back.map { L10n.Promote.backTo(step: L10n.string(label(for: $0))) }
    }

    /// The primary button: moving on, or writing (and while it writes).
    var nextLabel: LocalizedStringResource {
        if isSaving { return L10n.Promote.saving }
        switch controls.advance {
        case .step: return L10n.Promote.next
        case .save: return L10n.Promote.saveAndNext
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

    func choose(_ choice: Choice) { send(.edit(.choose(choice))) }
    func select(_ target: Target) { send(.edit(.selectTarget(target))) }
    /// The Confidence Select's value; empty means not stated.
    func setConfidence(_ gradeID: String) { send(.edit(.setConfidence(gradeID.isEmpty ? nil : gradeID))) }
    func setArgument(_ argument: String) { send(.edit(.setArgument(argument))) }
    func stepBack() { send(.stepBack) }
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
        switch status {
        case .missing: send(.subjectUnavailable(filedOn: nil))
        case .alreadyPromoted(let ref): send(.subjectUnavailable(filedOn: ref))
        case .loading, .promotable: break
        }
    }

    func subjectStatus(rows: SourceGraphRows?) -> SubjectStatus {
        guard let rows else { return .loading }
        let id = flow.subject.id
        guard rows.subjects.contains(where: { $0.id == id }) else { return .missing }
        if let membership = rows.memberships.first(where: { $0.subjectID == id }) {
            return .alreadyPromoted(onto: membership.entity.ref)
        }
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
            case .announceFiled(let name, let ref):
                session.noticeToast = VocabularyToast(
                    title: L10n.Promote.filedToast(name: name, ref: ref),
                    body: "",
                    tone: .success
                )
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
        case .save, .refreshAfterSave, .announceFiled: break
        }
    }

    private func write(_ save: PromoteFlow.Save) async {
        let result: CatalogPromoteResult
        do {
            result = try await store.promoteSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectID: save.subjectID,
                entityID: save.entityID,
                confidenceGradeID: save.confidenceGradeID,
                argument: save.argument
            )
        } catch {
            await perform(flow.send(.saveFailed(
                message: L10n.Errors.message(for: error),
                retryable: Self.isRetryable(error)
            )))
            return
        }
        await perform(flow.send(.saveSucceeded(entityRef: result.entity.ref)))
    }

    /// Whether saving again could succeed. A subject already filed elsewhere
    /// stays filed, so that refusal is final.
    static func isRetryable(_ error: Error) -> Bool {
        if case CoreInvokeError.coded(_, let code, _, _) = error, code == "identityclaims.already_member" {
            return false
        }
        return true
    }

    // MARK: Claim fields

    /// The Status Select: one option until Provisional and Rejected ship.
    var statusOptions: [PVSelectOption] {
        PromoteFlow.ClaimStatus.allCases.map { status in
            switch status {
            case .accepted: PVSelectOption(value: status.rawValue, label: L10n.string(L10n.Promote.statusAccepted))
            }
        }
    }

    var statusSelection: String { flow.draft.status.rawValue }

    /// Not stated first, then the grades in their vocabulary order.
    static func confidenceOptions(_ grades: [CatalogClaimConfidenceGrade]) -> [PVSelectOption] {
        [PVSelectOption(value: "", label: L10n.string(L10n.Promote.confidenceNone))]
            + grades.sorted { $0.sortOrder < $1.sortOrder }.map { PVSelectOption(value: $0.id, label: $0.label) }
    }

    var confidenceSelection: String { flow.draft.confidenceGradeID ?? "" }
    var argument: String { flow.draft.argument }

    var argumentHint: String {
        choice == .existing
            ? L10n.string(L10n.Promote.argumentHintExisting)
            : L10n.string(L10n.Promote.argumentHintNew(kind))
    }

    /// The handle ref prefix for this kind (PER, EVT, PLC), from the subject types.
    func refPrefix(in properties: PropertiesSnapshot?) -> String? {
        let prefix = properties?.types.first { $0.key == kind.rawValue }?.refPrefix
            .trimmingCharacters(in: .whitespacesAndNewlines)
        return prefix?.isEmpty == false ? prefix : nil
    }

    /// The write summary's target: a new handle's name and pending ref, or the chosen one.
    func summaryTarget(prefix: String?) -> (title: String, ref: String?) {
        if choice == .existing, let target {
            return (target.title, target.ref)
        }
        return (L10n.string(L10n.Promote.newOption(kind)), prefix.map { L10n.Promote.claimNewRef(prefix: $0) })
    }

    func summaryLine(prefix: String?) -> String {
        if choice == .existing, let target {
            return L10n.Promote.claimExistingLine(kind, members: Self.members(target.memberCount))
        }
        if let prefix {
            return L10n.Promote.claimNewLine(prefix: prefix, name: subject.name)
        }
        return L10n.Promote.claimNewLineUnprefixed(name: subject.name)
    }

    /// The refusal callout when the subject was filed elsewhere.
    var blockedTitle: String? {
        guard let failure = flow.failure else { return nil }
        if let ref = failure.filedOn {
            return L10n.Promote.blockedTitle(name: subject.name, ref: ref)
        }
        return L10n.Promote.blockedTitleUnknown(kind, name: subject.name)
    }

    var blockedMessage: String { L10n.Promote.blockedMessage(subjectRef: subject.ref) }

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
        case .announceFiled, .navigateToGraph, .resumeNavigation, .cancelNavigation: false
        }
    }
}
