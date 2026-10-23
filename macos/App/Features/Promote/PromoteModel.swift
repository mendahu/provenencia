import Foundation
import Observation

/// Frozen arrival query for the Promote place. Parsed once from the
/// `WorkspaceLocation` the graph card pushed.
struct PromoteEntry: Equatable, Sendable {
    let sourceID: String
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
        mapAll = location.promoteAll
    }

    var identityKey: String {
        mapAll ? "promote|\(sourceID)|all" : "promote|\(sourceID)|\(subjectID)"
    }

    var graphLocation: WorkspaceLocation {
        WorkspaceLocation(section: .sources, sourceId: sourceID, sourceSurface: .graph, title: sourceTitle)
    }

    private static func trimmed(_ text: String?) -> String? {
        guard let text = text?.trimmingCharacters(in: .whitespacesAndNewlines), !text.isEmpty else { return nil }
        return text
    }
}

extension WorkspaceLocation {
    static func promote(
        sourceId: String,
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

    var graphKey: CatalogQueryKey { .sourceGraph(project: session.projectKey, sourceId: entry.sourceID) }
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
        flow.nameOrders = session.nameFormat
        flow.mappedRest = entry.mapAll
        self.flow = flow
    }

    func load() async {
        do {
            let proposal = try await store.proposePromoteGraphAlignment(
                projectDir: session.projectKey.projectDir,
                sourceID: entry.sourceID,
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
            flow.nameOrders = session.nameFormat
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
        guard flow.setTarget(subjectID: subjectID, token: token) else { return }
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

    func toggleBridge(_ id: String) { flow.toggleBridge(id) }

    func openSheet(_ subjectID: String) { setSheetSubject(subjectID) }
    func closeSheet() { setSheetSubject(nil) }

    func setGroup(_ group: PromoteFlow.Grouping) { flow.group = group }

    func setBridgesOpen(_ open: Bool) { flow.bridgesOpen = open }

    /// Opens, switches or closes the sheet. A row filed on a handle reads its
    /// claim evidence each time its sheet opens; any other row has none.
    func setSheetSubject(_ id: String?) {
        evidenceGeneration += 1
        sheetEvidence = .idle
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
                    sourceID: entry.sourceID,
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
            labels: { [self] key in propertyLabel(key: key, origin: CatalogOrigin.provenencia) },
            worded: { [session] value in ReconciledValueDisplay.string(for: value, orders: session.nameFormat) }
        )
    }

    /// The researcher's name order changed. An open sheet's words were made in the old one, so
    /// its evidence is read and worded again.
    func nameFormatChanged() {
        guard session.nameFormat != flow.nameOrders else { return }
        flow.nameOrders = session.nameFormat
        if let id = flow.sheetSubjectID { setSheetSubject(id) }
    }

    func done() {
        guard canFinish else { return }
        flow.saving = true
        flow.staleNote = nil
        saveError = nil
        Task { await file() }
    }

    func leave() {
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
        let band = L10n.string(L10n.Promote.assessment(readout.assessment))
        return readout.contested ? L10n.Promote.assessmentContested(band) : band
    }

    /// The catalog label for a Property, or its key when the snapshot lacks it.
    func propertyLabel(key: String, origin: String) -> String {
        propertyLabels[key + "|" + origin] ?? key
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
                sourceID: entry.sourceID,
                fixed: fixed
            )
            let snapshot = await graphSnapshot()
            guard generation == proposeGeneration else { return }
            if let snapshot {
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
                sourceID: entry.sourceID,
                seenRevision: seen,
                rows: batch,
                skipBridgeIDs: flow.skipBridgeIDs()
            )
            session.apply(.promotedSubject(sourceId: entry.sourceID))
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
    private func graphSnapshot() async -> SourceGraphSnapshot? {
        let rows: SourceGraphRows? = await session.readyValue(graphKey)
        let properties: PropertiesSnapshot? = await session.readyValue(propertiesKey)
        let rules: [CatalogConnectRule]? = await session.readyValue(rulesKey)
        if let properties {
            propertyLabels = Dictionary(
                properties.properties.map { ($0.key + "|" + $0.origin, $0.label) },
                uniquingKeysWith: { first, _ in first }
            )
        }
        guard let rows else { return nil }
        return SourceGraphSnapshot.build(
            rows: rows,
            types: properties?.types ?? [],
            rules: rules ?? []
        )
    }

    /// The snapshot's Subjects as rows read them.
    private func subjectFacts(_ snapshot: SourceGraphSnapshot) -> [PromoteFlow.SubjectFact] {
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
                anchor: anchor
            )
        }
    }

    /// The snapshot's connections as the page reads them, with each end named from `subjects`.
    static func bridgeFacts(
        _ snapshot: SourceGraphSnapshot,
        subjects: [PromoteFlow.SubjectFact]
    ) -> [PromoteFlow.BridgeFact] {
        let nameByID = Dictionary(subjects.map { ($0.id, $0.name) }, uniquingKeysWith: { first, _ in first })
        let placedByID = EvidenceBridgeEdgeSummary.subjects(of: snapshot)
        return snapshot.bridges.map { bridge in
            PromoteFlow.BridgeFact(
                id: bridge.subject.id,
                sentence: EvidenceBridgeEdgeSummary.sentence(for: bridge, in: snapshot, subjects: placedByID),
                phrase: EvidenceBridgeEdgeSummary.phrase(for: bridge),
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
        if let type = types?.first(where: { $0.id == source.sourceTypeID }) {
            sourceMark = PVMarkKey(catalogKey: type.iconKey)
        }
    }
}

extension PromoteModel: WorkspaceLeaveGuard {
    func shouldHoldNavigation(_ pending: PendingNavigation) -> Bool {
        flow.requestLeave(pending)
    }
}
