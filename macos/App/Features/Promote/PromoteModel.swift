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

/// Runs the Promote page: loads the proposal, keeps the row list, and files Done.
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
    let catalogCounts: CatalogCounts?
    weak var navigation: WorkspaceNavigation?

    private(set) var flow: PromoteFlow
    private(set) var grades: [CatalogClaimConfidenceGrade] = []
    private(set) var sourceRef: String = ""
    private(set) var sourceMark: PVMarkKey = .defaultTypeMark
    private(set) var loadError: String?
    private(set) var saveError: String?

    var graphKey: CatalogQueryKey { .sourceGraph(project: session.projectKey, sourceId: entry.sourceID) }
    var confidenceKey: CatalogQueryKey { .confidenceGradesList(project: session.projectKey) }
    var propertiesKey: CatalogQueryKey { .propertiesWorkspace(project: session.projectKey) }
    var rulesKey: CatalogQueryKey { .connectRules(project: session.projectKey) }
    var sourcesKey: CatalogQueryKey { .sourcesList(project: session.projectKey) }
    var sourceTypesKey: CatalogQueryKey { .sourceTypesList(project: session.projectKey) }

    var hasUnsavedWork: Bool { flow.manual }
    var pendingLeave: PendingLeave? { flow.pendingLeave.map(PendingLeave.init(navigation:)) }
    var isSaving: Bool { flow.saving }

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
        flow = PromoteFlow(entryID: entry.subjectID)
    }

    func load() async {
        do {
            let proposal = try await store.proposePromoteGraphAlignment(
                projectDir: session.projectKey.projectDir,
                sourceID: entry.sourceID,
                fixed: []
            )
            let facts = await facts()
            grades = await session.readyValue(confidenceKey) ?? []
            await attachSource()
            flow.load(entryID: entry.subjectID, proposal: proposal, subjects: facts.subjects, bridges: facts.bridges)
            loadError = nil
        } catch {
            loadError = L10n.Errors.message(for: error)
        }
    }

    func mapRest() { flow.mapRest() }

    func setTarget(subjectID: String, token: String) {
        guard flow.setTarget(subjectID: subjectID, token: token) else { return }
        Task { await repropose() }
    }

    func togglePin(subjectID: String, comparisonID: String) {
        flow.togglePin(subjectID: subjectID, comparisonID: comparisonID)
    }

    func setArgument(subjectID: String, argument: String) {
        flow.setArgument(subjectID: subjectID, argument: argument)
    }

    func setConfidence(subjectID: String, gradeID: String?) {
        flow.setConfidence(subjectID: subjectID, gradeID: gradeID)
    }

    func toggleBridge(_ id: String) { flow.toggleBridge(id) }

    func openSheet(_ subjectID: String) { flow.sheetSubjectID = subjectID }
    func closeSheet() { flow.sheetSubjectID = nil }

    func setGroup(_ group: String) { flow.group = group }

    func setBridgesOpen(_ open: Bool) { flow.bridgesOpen = open }

    func setSheetSubject(_ id: String?) { flow.sheetSubjectID = id }

    func done() {
        guard !flow.saving else { return }
        flow.saving = true
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

    func reasonText(for row: PromoteFlow.Row) -> String {
        let parts = row.reason.split(separator: "|", maxSplits: 1).map(String.init)
        if parts.count == 2, let neighbor = flow.rows.first(where: { $0.subjectID == parts[0] }) {
            return L10n.Promote.via(neighbor: neighbor.name, ref: neighbor.ref, role: parts[1])
        }
        if row.reason == "fixed" { return L10n.string(L10n.Promote.alreadyFiled) }
        if row.reason.hasPrefix("agree ") {
            return L10n.Promote.agreesOn(property: String(row.reason.dropFirst("agree ".count)))
        }
        return row.reason
    }

    private func repropose() async {
        let fixed = flow.rows.compactMap { row -> CatalogPromoteGraphAlignmentFixed? in
            guard row.decided, let id = row.target.handleID else { return nil }
            return CatalogPromoteGraphAlignmentFixed(subjectID: row.subjectID, handleID: id)
        }
        do {
            let proposal = try await store.proposePromoteGraphAlignment(
                projectDir: session.projectKey.projectDir,
                sourceID: entry.sourceID,
                fixed: fixed
            )
            let facts = await facts()
            flow.merge(proposal, subjects: facts.subjects)
        } catch {
            saveError = L10n.Errors.message(for: error)
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
            await catalogCounts?.refreshAll()
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
                await repropose()
                return
            }
            saveError = L10n.Errors.message(for: error)
        }
    }

    private func facts() async -> (subjects: [PromoteFlow.SubjectFact], bridges: [PromoteFlow.BridgeFact]) {
        let rows: SourceGraphRows? = await session.readyValue(graphKey)
        let properties: PropertiesSnapshot? = await session.readyValue(propertiesKey)
        let rules: [CatalogConnectRule]? = await session.readyValue(rulesKey)
        guard let rows else { return ([], []) }
        let snapshot = SourceGraphSnapshot.build(
            rows: rows,
            types: properties?.types ?? [],
            rules: rules ?? []
        )
        let subjects = snapshot.subjects.map { placed -> PromoteFlow.SubjectFact in
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
                name: placed.displayName,
                ref: placed.subject.ref,
                anchor: anchor
            )
        }
        let nameByID = Dictionary(uniqueKeysWithValues: subjects.map { ($0.id, $0.name) })
        let bridges = snapshot.bridges.map { bridge in
            PromoteFlow.BridgeFact(
                id: bridge.subject.id,
                sentence: EvidenceBridgeEdgeSummary.sentence(for: bridge, in: snapshot),
                phrase: EvidenceBridgeEdgeSummary.phrase(for: bridge),
                endAName: bridge.endpointAID.flatMap { nameByID[$0] } ?? "",
                endBName: bridge.endpointBID.flatMap { nameByID[$0] } ?? "",
                mark: bridge.kind.markKey,
                endA: bridge.endpointAID,
                endB: bridge.endpointBID
            )
        }
        return (subjects, bridges)
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
