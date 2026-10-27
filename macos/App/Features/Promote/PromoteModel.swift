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

/// Runs the Promote page: loads the proposal, keeps the row list, and files Done.
@MainActor
@Observable
final class PromoteModel {
    /// Where the open sheet's evidence read stands.
    enum SheetEvidence: Equatable {
        case idle
        case loading
        case failed(String)
    }

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
    private(set) var grades: [CatalogClaimConfidenceGrade] = []
    private(set) var sourceRef: String = ""
    private(set) var sourceMark: PVMarkKey = .defaultTypeMark
    /// Property labels by "key|origin", from the Properties snapshot.
    private(set) var propertyLabels: [String: String] = [:]
    /// Property labels by key alone: the product's, else a researcher's.
    private var propertyLabelsByKey: [String: String] = [:]
    private(set) var loadError: String?
    private(set) var saveError: String?
    private(set) var proposeError: String?
    /// The open sheet's claim evidence read: the loaded value lives on its row.
    private(set) var sheetEvidence = SheetEvidence.idle
    private var evidenceGeneration = 0
    /// A proposal after a decision is in flight; Done waits for it.
    private(set) var isProposing = false
    /// Bumped per proposal; only the latest one's answer is merged.
    private var proposeGeneration = 0

    var graphKey: CatalogQueryKey { .evidenceGraph(project: session.projectKey, graphId: entry.graphID) }
    var confidenceKey: CatalogQueryKey { .confidenceGradesList(project: session.projectKey) }
    var propertiesKey: CatalogQueryKey { .propertiesWorkspace(project: session.projectKey) }
    var rulesKey: CatalogQueryKey { .connectRules(project: session.projectKey) }
    var sourcesKey: CatalogQueryKey { .sourcesList(project: session.projectKey) }
    var sourceTypesKey: CatalogQueryKey { .sourceTypesList(project: session.projectKey) }

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
        do {
            let proposal = try await store.proposePromoteGraphAlignment(
                projectDir: session.projectKey.projectDir,
                graphID: entry.graphID,
                fixed: []
            )
            // Rows are built from the graph's Subjects. Without them every row
            // would drop and the page would read as empty: say so instead.
            guard let snapshot = await graphSnapshot() else {
                loadError = L10n.string(L10n.Promote.graphUnavailable)
                return
            }
            grades = await session.readyValue(confidenceKey) ?? []
            await attachSource()
            syncNameOrders()
            let subjects = subjectFacts(snapshot)
            flow.load(
                entryID: entry.subjectID, proposal: proposal, subjects: subjects,
                bridges: Self.bridgeFacts(snapshot, subjects: subjects)
            )
            loadError = nil
        } catch {
            loadError = L10n.Errors.message(for: error)
        }
    }

    func mapRest() { flow.mapRest() }

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
        hold(subjectID: open.subjectID, id: hit.id, ref: hit.ref, header: hit.header)
    }

    func closeRecordSearch() {
        recordSearch?.search.cancel()
        recordSearch = nil
    }

    /// Holds a record Promote did not offer as the row's decided target.
    func hold(subjectID: String, id: String, ref: String, header: CatalogConclusionHeader?) {
        guard flow.hold(subjectID: subjectID, id: id, ref: ref, header: header) else { return }
        proposeAfterDecision()
    }

    /// A decision changed the fixed pairs: propose again.
    private func proposeAfterDecision() {
        flow.staleNote = nil
        let generation = nextProposal()
        Task { await repropose(generation: generation) }
    }

    func togglePin(subjectID: String, pinIDs: [String]) {
        flow.togglePin(subjectID: subjectID, pinIDs: pinIDs)
    }

    func setArgument(subjectID: String, argument: String) {
        flow.setArgument(subjectID: subjectID, argument: argument)
    }

    func setConfidence(subjectID: String, gradeID: String?) {
        flow.setConfidence(subjectID: subjectID, gradeID: gradeID)
    }

    func openSheet(_ subjectID: String) { setSheetSubject(subjectID) }
    func closeSheet() { setSheetSubject(nil) }

    /// Reads the open sheet's evidence again, after a failed read.
    func reloadEvidence() { setSheetSubject(flow.sheetSubjectID) }

    /// Opens, switches or closes the sheet. A row filed on a handle reads its
    /// claim evidence each time its sheet opens; any other row has none.
    func setSheetSubject(_ id: String?) {
        evidenceGeneration += 1
        sheetEvidence = .idle
        syncNameOrders()
        flow.sheetSubjectID = id
        guard let id, let row = flow.row(id), !row.anchor, case .handle(let handleID, _, _) = row.target else { return }
        loadEvidence(subjectID: id, handleID: handleID)
    }

    private func loadEvidence(subjectID: String, handleID: String) {
        let generation = evidenceGeneration
        let neighbors = flow.evidenceNeighbors(for: subjectID)
        sheetEvidence = .loading
        Task {
            do {
                let evidence = try await store.getPromoteClaimEvidence(
                    projectDir: session.projectKey.projectDir,
                    graphID: entry.graphID,
                    subjectID: subjectID,
                    handleID: handleID,
                    neighbors: neighbors
                )
                guard generation == evidenceGeneration else { return }
                flow.setEvidence(subjectID: subjectID, handleID: handleID, board: board(of: evidence))
                sheetEvidence = .idle
            } catch {
                guard generation == evidenceGeneration else { return }
                sheetEvidence = .failed(L10n.Errors.message(for: error))
            }
        }
    }

    /// The evidence worded for the sheet, once: the page's names for the connected rows and the
    /// researcher's name order.
    private func board(of evidence: CatalogClaimEvidence) -> EvidenceBoard {
        EvidenceBoard.build(
            evidence: evidence,
            neighbor: { [flow] id in flow.row(id) },
            labels: { [self] key in propertyLabel(key: key) },
            worded: { [session] value in ReconciledValueDisplay.string(for: value, orders: session.nameFormat) }
        )
    }

    /// The one place the page takes the researcher's name order from the session: when it loads,
    /// proposes again and opens a sheet, whose evidence is worded in the same order.
    private func syncNameOrders() {
        flow.nameOrders = session.nameFormat
    }

    func done() {
        closeRecordSearch()
        guard canFinish else { return }
        flow.saving = true
        flow.staleNote = nil
        saveError = nil
        Task { await file() }
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
        case .filed: return L10n.string(L10n.Promote.alreadyFiled)
        case .decided: return L10n.string(L10n.Promote.reasonDecided)
        case .agrees(let key, let origin): return L10n.Promote.agreesOn(property: propertyLabel(key: key, origin: origin))
        case .weak: return L10n.string(L10n.Promote.reasonWeak)
        case .taken: return L10n.string(L10n.Promote.reasonTaken)
        case .noMatch: return L10n.string(L10n.Promote.reasonNoMatch)
        case .empty: return L10n.string(L10n.Promote.reasonEmpty)
        case .neighborPlaced(let id), .neighborMoved(let id), .neighborReleased(let id):
            guard let neighbor = neighbor(id) else { return L10n.string(L10n.Promote.reasonMoveUnknown) }
            switch reason {
            case .neighborPlaced:
                return neighbor.target.handleRef.map { L10n.Promote.neighborPlaced(neighbor: neighbor.name, ref: $0) }
                    ?? L10n.string(L10n.Promote.reasonMoveUnknown)
            case .neighborMoved:
                return neighbor.target.handleRef.map { L10n.Promote.neighborMoved(neighbor: neighbor.name, ref: $0) }
                    ?? L10n.string(L10n.Promote.reasonMoveUnknown)
            case .neighborReleased:
                return L10n.Promote.neighborReleased(neighbor: neighbor.name)
            default:
                // Only the three move reasons reach this arm.
                return L10n.string(L10n.Promote.reasonMoveUnknown)
            }
        }
    }

    /// The band as the row's button speaks it: "Strong", or "Strong, contested".
    func assessmentText(_ readout: PromoteFlow.SelectionReadout) -> String {
        let band = L10n.string(L10n.Promote.assessment(readout))
        return readout.contested ? L10n.Promote.assessmentContested(band) : band
    }

    /// The catalog label for a Property, or its key when the snapshot lacks it.
    func propertyLabel(key: String, origin: String) -> String {
        propertyLabels[key + "|" + origin] ?? key
    }

    /// The label for a Property known by its key alone: the product's Property, else a researcher's
    /// own with that key, else the key.
    func propertyLabel(key: String) -> String {
        propertyLabelsByKey[key] ?? key
    }

    private func nextProposal() -> Int {
        proposeGeneration += 1
        isProposing = true
        return proposeGeneration
    }

    private func repropose(generation: Int) async {
        defer {
            if generation == proposeGeneration {
                isProposing = false
            }
        }
        let fixed = flow.fixed(for: flow.rows.filter(\.decided))
        do {
            let proposal = try await store.proposePromoteGraphAlignment(
                projectDir: session.projectKey.projectDir,
                graphID: entry.graphID,
                fixed: fixed
            )
            let snapshot = await graphSnapshot()
            guard generation == proposeGeneration else { return }
            if let snapshot {
                syncNameOrders()
                // Bridges are the page's, built once at load: a proposal moves rows, not connections.
                flow.merge(proposal, subjects: subjectFacts(snapshot))
                proposeError = nil
            } else {
                // Merging without the graph would drop every row and its
                // decisions. Keep the rows as they are.
                proposeError = L10n.string(L10n.Promote.graphUnavailableRepropose)
            }
        } catch {
            guard generation == proposeGeneration else { return }
            proposeError = L10n.Errors.message(for: error)
        }
    }

    private func file() async {
        let batch = flow.batchRows()
        let seen = flow.revision
        do {
            _ = try await store.applyPromoteGraphAlignment(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                graphID: entry.graphID,
                seenRevision: seen,
                rows: batch
            )
            session.apply(.promotedSubject(graphId: entry.graphID))
            session.noticeToast = VocabularyToast(
                title: L10n.Promote.filedTitle(filed: flow.fileCount, connections: flow.connectionCount),
                body: flow.skipCount > 0 ? L10n.Promote.filedBody(skipped: flow.skipCount) : "",
                tone: .success
            )
            flow.saving = false
            flow.manual = false
            navigation?.go(to: entry.graphLocation)
        } catch {
            flow.saving = false
            if case CoreInvokeError.coded(_, let code, _, _) = error, code == "promote.stale" {
                flow.staleNote = L10n.string(L10n.Promote.stale)
                await repropose(generation: nextProposal())
                return
            }
            saveError = L10n.Errors.message(for: error)
        }
    }

    /// The Source graph as a snapshot, with the property labels refreshed from the same read. Nil
    /// when the graph read isn't ready (never loaded, or failed with nothing cached).
    private func graphSnapshot() async -> EvidenceGraphSnapshot? {
        let rows: EvidenceGraphRows? = await session.readyValue(graphKey)
        let properties: PropertiesSnapshot? = await session.readyValue(propertiesKey)
        let rules: [CatalogConnectRule]? = await session.readyValue(rulesKey)
        if let properties {
            propertyLabels = Dictionary(
                properties.properties.map { ($0.key + "|" + $0.origin, $0.label) },
                uniquingKeysWith: { first, _ in first }
            )
            propertyLabelsByKey = [:]
            for property in properties.properties where propertyLabelsByKey[property.key] == nil || CatalogOrigin.isProduct(property.origin) {
                propertyLabelsByKey[property.key] = property.label
            }
        }
        guard let rows else { return nil }
        return EvidenceGraphSnapshot.build(
            rows: rows,
            types: properties?.types ?? [],
            rules: rules ?? []
        )
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
                endB: bridge.endpointBID,
                alreadyFiled: bridge.membership != nil
            )
        }
    }

    private func attachSource() async {
        let sources: [CatalogSource]? = await session.readyValue(sourcesKey)
        let types: [CatalogSourceType]? = await session.readyValue(sourceTypesKey)
        guard let source = sources?.first(where: { $0.id == entry.sourceID }) else { return }
        sourceRef = source.ref
        if let iconKey = types?.iconKey(forTypeID: source.sourceTypeID) {
            sourceMark = PVMarkKey(catalogKey: iconKey)
        }
    }
}

extension PromoteModel: WorkspaceLeaveGuard {
    func shouldHoldNavigation(_ pending: PendingNavigation) -> Bool {
        flow.requestLeave(pending)
    }
}
