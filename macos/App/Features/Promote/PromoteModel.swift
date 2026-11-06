import Foundation
import Observation

/// Frozen arrival query for the Promote place. Parsed once from the
/// `WorkspaceLocation` the graph card pushed.
struct PromoteEntry: Equatable, Sendable {
    let sourceID: String
    let graphID: String
    let subjectID: String
    let kind: EvidencePrimaryKind
    let subjectRef: String
    let subjectName: String
    let sourceTitle: String?
    /// Opened from Promote all: every subject starts expanded.
    let mapAll: Bool

    init?(location: WorkspaceLocation) {
        guard location.section == .sources,
              location.sourceSurface == .promote,
              let sourceID = location.sourceId,
              let graphID = location.graphId,
              let subjectID = location.subjectId,
              let kind = location.subjectTypeKey.flatMap(EvidencePrimaryKind.init(rawValue:))
        else { return nil }
        self.sourceID = sourceID
        self.graphID = graphID
        self.subjectID = subjectID
        self.kind = kind
        let ref = Self.trimmed(location.ref) ?? ""
        subjectRef = ref
        subjectName = Self.trimmed(location.title) ?? ref
        sourceTitle = Self.trimmed(location.sourceTitle)
        mapAll = location.promoteAll
    }

    var identityKey: String {
        mapAll ? "promote|\(graphID)|all" : "promote|\(graphID)|\(subjectID)"
    }

    var graphLocation: WorkspaceLocation {
        .evidenceGraph(sourceId: sourceID, graphId: graphID, title: sourceTitle)
    }

    private static func trimmed(_ text: String?) -> String? {
        guard let text = text?.trimmingCharacters(in: .whitespacesAndNewlines), !text.isEmpty else { return nil }
        return text
    }
}

extension WorkspaceLocation {
    static func promote(
        sourceId: String,
        graphId: String,
        subjectId: String,
        kind: EvidencePrimaryKind,
        ref: String?,
        title: String?,
        sourceTitle: String?,
        mapAll: Bool = false
    ) -> WorkspaceLocation {
        WorkspaceLocation(
            section: .sources,
            sourceId: sourceId,
            graphId: graphId,
            subjectId: subjectId,
            subjectTypeKey: kind.rawValue,
            sourceSurface: .promote,
            promoteAll: mapAll,
            ref: ref,
            title: title,
            sourceTitle: sourceTitle
        )
    }
}

/// Runs the Promote page: loads the proposal, keeps the row list, and promotes Done.
@MainActor
@Observable
final class PromoteModel {
    struct PendingLeave: Identifiable, Equatable {
        let navigation: PendingNavigation
        var id: String { String(describing: navigation) }
    }

    let entry: PromoteEntry
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String
    weak var navigation: WorkspaceNavigation?

    private(set) var flow: PromoteFlow
    private(set) var loadError: String?
    private(set) var saveError: String?
    private(set) var proposeError: String?
    private(set) var isLoading = false
    private(set) var loadAttempt = 0
    private(set) var noticeToken = 0
    /// A proposal after a decision is in flight; Done waits for it.
    private(set) var isProposing = false
    private var pendingProposal: PromoteProposalInput?

    private struct PromoteProposalInput {
        let fixed: [CatalogPromoteGraphAlignmentFixed]
        let onlySubjectID: String?
    }

    var graphKey: CatalogQueryKey { .evidenceGraph(project: session.projectKey, graphId: entry.graphID) }
    var propertiesKey: CatalogQueryKey { .propertiesWorkspace(project: session.projectKey) }
    var rulesKey: CatalogQueryKey { .connectRules(project: session.projectKey) }
    var sourcesKey: CatalogQueryKey { .sourcesList(project: session.projectKey) }
    var sourceTypesKey: CatalogQueryKey { .sourceTypesList(project: session.projectKey) }

    /// What the header shows of the Source, read from one lookup of its row in the
    /// session (which every rename and type change patches). While the Sources list is
    /// not warm, or once the Source is gone: the carried title, no ref, the default mark.
    var sourceHeader: PromoteSourceHeader {
        let row = session.source(id: entry.sourceID)
        let types: QueryHandle<[CatalogSourceType]>? = session.queryHandle(sourceTypesKey)
        let mark = types?.value?.iconKey(forTypeID: row?.sourceTypeID).map { PVMarkKey(catalogKey: $0) }
        return PromoteSourceHeader(
            title: WorkspaceLocationLabels.sourceTitle(live: row?.title, carried: entry.sourceTitle),
            ref: row?.ref ?? "",
            mark: mark ?? .defaultTypeMark
        )
    }

    var hasUnsavedWork: Bool { flow.manual }
    var pendingLeave: PendingLeave? { flow.pendingLeave.map(PendingLeave.init(navigation:)) }
    var isSaving: Bool { flow.saving }
    var canFinish: Bool { !flow.saving && !isProposing && loadError == nil && flow.targetsChosen }

    init(
        entry: PromoteEntry,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String
    ) {
        self.entry = entry
        self.session = session
        self.store = store
        self.userID = userID
        var flow = PromoteFlow(entryID: entry.subjectID)
        flow.mappedRest = entry.mapAll
        self.flow = flow
    }

    func load() async {
        defer { isLoading = false }
        do {
            let proposal = try await store.proposePromoteGraphAlignment(
                projectDir: session.projectKey.projectDir,
                graphID: entry.graphID,
                fixed: [],
                onlySubjectID: flow.onlySubjectID
            )
            // Rows are built from the graph's Subjects. Without them every row
            // would drop and the page would read as empty: say so instead.
            guard let snapshot = await graphSnapshot() else {
                loadError = L10n.string(L10n.Promote.graphUnavailable)
                loadAttempt += 1
                return
            }
            await warmSourceKeys()
            syncNameOrders()
            let subjects = subjectFacts(snapshot)
            flow.load(
                entryID: entry.subjectID, proposal: proposal, subjects: subjects,
                bridges: Self.bridgeFacts(snapshot, subjects: subjects)
            )
            loadError = nil
        } catch {
            loadError = L10n.Errors.message(for: error)
            loadAttempt += 1
        }
    }

    /// Loads the page again after a failed read or proposal on open.
    func retryLoad() {
        guard !isLoading else { return }
        isLoading = true
        let _: QueryHandle<EvidenceGraphRows> = session.reread(graphKey)
        Task { await load() }
    }

    /// Maps the rest of the page's Subjects. The scope is part of what Done files, so Go is asked again.
    func mapRest() {
        flow.mapRest()
        proposeAfterDecision()
    }

    func setTarget(subjectID: String, token: String) {
        if token == PromoteFlow.Target.findToken {
            openRecordSearch(subjectID: subjectID)
            return
        }
        guard flow.setTarget(subjectID: subjectID, token: token) else { return }
        proposeAfterDecision()
    }

    /// A search of one row's kind, open in that row's target cell.
    struct RecordSearch {
        let subjectID: String
        let search: CatalogSearchModel
    }

    /// The one open record search; nil when no row is searching. Opening or closing it decides nothing.
    private(set) var recordSearch: RecordSearch?

    /// Turns the row's target cell into a search over every record of its kind. Any open search closes.
    func openRecordSearch(subjectID: String) {
        guard let row = flow.row(subjectID), !row.anchor else { return }
        closeRecordSearch()
        recordSearch = RecordSearch(subjectID: subjectID, search: CatalogSearchModel(kinds: [row.kind.rawValue]))
    }

    func recordSearchQueryChanged(_ text: String) {
        guard let search = recordSearch?.search else { return }
        search.query = text
        search.scheduleSearch(projectDir: session.projectKey.projectDir, location: nil, store: store)
    }

    /// Holds the picked hit as the open search's row's decided target, then closes the search.
    func pickSearchHit(id: String) {
        guard let open = recordSearch,
              let hit = open.search.hits.first(where: { $0.id == id }) else { return }
        closeRecordSearch()
        hold(subjectID: open.subjectID, id: hit.id, ref: hit.ref, summary: hit.summary)
    }

    func closeRecordSearch() {
        recordSearch?.search.cancel()
        recordSearch = nil
    }

    /// Holds a record Promote did not offer as the row's decided target.
    func hold(subjectID: String, id: String, ref: String, summary: CatalogConclusionSummary?) {
        guard flow.hold(subjectID: subjectID, id: id, ref: ref, summary: summary) else { return }
        proposeAfterDecision()
    }

    /// A decision changed the fixed pairs: propose again.
    private func proposeAfterDecision() {
        flow.staleNote = nil
        proposeError = nil
        submitProposal()
    }

    private func submitProposal() {
        pendingProposal = PromoteProposalInput(
            fixed: flow.fixed(for: flow.rows.filter(\.decided)),
            onlySubjectID: flow.onlySubjectID
        )
        guard !isProposing else { return }
        isProposing = true
        Task { [weak self] in
            while let self {
                guard let input = pendingProposal else {
                    isProposing = false
                    return
                }
                pendingProposal = nil
                do {
                    let proposal = try await store.proposePromoteGraphAlignment(
                        projectDir: session.projectKey.projectDir,
                        graphID: entry.graphID,
                        fixed: input.fixed,
                        onlySubjectID: input.onlySubjectID
                    )
                    if pendingProposal != nil { continue }
                    let snapshot = await graphSnapshot()
                    if pendingProposal != nil { continue }
                    if let snapshot {
                        syncNameOrders()
                        flow.merge(proposal, subjects: subjectFacts(snapshot))
                        showingNotices { self.proposeError = nil }
                    } else {
                        showingNotices {
                            self.flow.staleNote = nil
                            self.proposeError = L10n.string(L10n.Promote.graphUnavailableRepropose)
                        }
                    }
                } catch {
                    if pendingProposal != nil { continue }
                    showingNotices {
                        self.flow.staleNote = nil
                        self.proposeError = L10n.Errors.message(for: error)
                    }
                }
                if pendingProposal == nil {
                    isProposing = false
                    return
                }
            }
        }
    }

    private func showingNotices(_ change: () -> Void) {
        let hadStale = flow.staleNote != nil
        let hadSave = saveError != nil
        let hadPropose = proposeError != nil
        change()
        if (flow.staleNote != nil && !hadStale)
            || (saveError != nil && !hadSave)
            || (proposeError != nil && !hadPropose)
        {
            noticeToken += 1
        }
    }

    /// The one place the page takes the researcher's name order from the session: when it loads
    /// and proposes again.
    private func syncNameOrders() {
        flow.nameOrders = session.nameFormat
    }

    func done() {
        closeRecordSearch()
        guard canFinish else { return }
        flow.saving = true
        flow.staleNote = nil
        saveError = nil
        proposeError = nil
        Task { await promote() }
    }

    func leave() {
        closeRecordSearch()
        let pending = flow.pendingLeave
        flow.cancelLeave()
        navigation?.resumeHeldNavigation()
        if pending == nil { navigation?.go(to: entry.graphLocation) }
    }

    func keepPromoting() { flow.cancelLeave() }

    /// The page row a connected Subject is, or nil when it is not on the page.
    func neighbor(_ subjectID: String) -> PromoteFlow.Row? {
        flow.row(subjectID)
    }

    func reasonText(_ reason: PromoteFlow.Reason, on row: PromoteFlow.Row) -> String {
        switch reason {
        case .via(let neighborID):
            guard let neighbor = neighbor(neighborID) else {
                return L10n.string(L10n.Promote.reasonVia)
            }
            let role = flow.bridge(between: row.subjectID, and: neighborID)?.term ?? ""
            return role.isEmpty
                ? L10n.Promote.viaNeighbor(neighbor: neighbor.name, ref: neighbor.shownRef)
                : L10n.Promote.via(neighbor: neighbor.name, ref: neighbor.shownRef, role: role)
        case .neighborPlaced(let id):
            guard let neighbor = neighbor(id), let ref = neighbor.target.handleRef else {
                return L10n.string(L10n.Promote.reasonMoveUnknown)
            }
            return L10n.Promote.neighborPlaced(neighbor: neighbor.name, ref: ref)
        case .neighborMoved(let id):
            guard let neighbor = neighbor(id), let ref = neighbor.target.handleRef else {
                return L10n.string(L10n.Promote.reasonMoveUnknown)
            }
            return L10n.Promote.neighborMoved(neighbor: neighbor.name, ref: ref)
        case .neighborReleased(let id):
            guard let neighbor = neighbor(id) else { return L10n.string(L10n.Promote.reasonMoveUnknown) }
            return L10n.Promote.neighborReleased(neighbor: neighbor.name)
        }
    }

    private func promote() async {
        let batch = flow.batchRows()
        let seen = flow.revision
        do {
            let effects = try await store.applyPromoteGraphAlignment(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                graphID: entry.graphID,
                seenRevision: seen,
                rows: batch
            )
            session.apply(effects)
            session.noticeToast = VocabularyToast(
                title: L10n.Promote.promotedTitle(promoted: flow.promoteCount, connections: flow.connectionCount),
                body: flow.skipCount > 0 ? L10n.Promote.promotedBody(skipped: flow.skipCount) : "",
                tone: .success
            )
            flow.saving = false
            flow.manual = false
            navigation?.go(to: entry.graphLocation)
        } catch {
            flow.saving = false
            if case CoreInvokeError.coded(_, let code, _, _) = error, code == "promote.stale" {
                showingNotices { flow.staleNote = L10n.string(L10n.Promote.stale) }
                submitProposal()
                return
            }
            showingNotices { saveError = L10n.Errors.message(for: error) }
        }
    }

    /// The Source graph as a snapshot. Nil
    /// when the graph read isn't ready (never loaded, or failed with nothing cached).
    private func graphSnapshot() async -> EvidenceGraphSnapshot? {
        await session.evidenceGraphSnapshot(graphID: entry.graphID)
    }

    /// The snapshot's Subjects as rows read them.
    private func subjectFacts(_ snapshot: EvidenceGraphSnapshot) -> [PromoteFlow.SubjectFact] {
        snapshot.subjects.map { placed -> PromoteFlow.SubjectFact in
            let anchor: PromoteFlow.Anchor?
            if let membership = placed.membership {
                let title = membership.name?.form ?? membership.entity.ref
                anchor = PromoteFlow.Anchor(id: membership.entity.id, ref: membership.entity.ref, title: title)
            } else {
                anchor = nil
            }
            return PromoteFlow.SubjectFact(
                id: placed.subject.id,
                kind: placed.kind,
                name: placed.displayName(orders: flow.nameOrders),
                ref: placed.subject.ref,
                anchor: anchor,
                date: placed.dateLine
            )
        }
    }

    /// The snapshot's connections as the page reads them, with each end named from `subjects`.
    static func bridgeFacts(
        _ snapshot: EvidenceGraphSnapshot,
        subjects: [PromoteFlow.SubjectFact]
    ) -> [PromoteFlow.BridgeFact] {
        let nameByID = Dictionary(subjects.map { ($0.id, $0.name) }, uniquingKeysWith: { first, _ in first })
        let placedByID = EvidenceBridgeEdgeSummary.subjects(of: snapshot)
        return snapshot.bridges.map { bridge in
            PromoteFlow.BridgeFact(
                id: bridge.subject.id,
                sentence: EvidenceBridgeEdgeSummary.sentence(for: bridge, in: snapshot, subjects: placedByID, bridges: [:]),
                term: EvidenceBridgeEdgeSummary.reasonRole(for: bridge),
                endAName: bridge.endpointAID.flatMap { nameByID[$0] } ?? "",
                endBName: bridge.endpointBID.flatMap { nameByID[$0] } ?? "",
                mark: bridge.kind.markKey,
                endA: bridge.endpointAID,
                endB: bridge.endpointBID
            )
        }
    }

    /// Waits for the Sources and source-types keys so the header has values when the page settles.
    private func warmSourceKeys() async {
        let _: [CatalogSource]? = await session.readyValue(sourcesKey)
        let _: [CatalogSourceType]? = await session.readyValue(sourceTypesKey)
    }
}

extension PromoteModel: WorkspaceLeaveGuard {
    func shouldHoldNavigation(_ pending: PendingNavigation) -> Bool {
        flow.requestLeave(pending)
    }
}

/// The Source facts the Promote page header shows.
struct PromoteSourceHeader: Equatable {
    let title: String?
    let ref: String
    let mark: PVMarkKey
}
