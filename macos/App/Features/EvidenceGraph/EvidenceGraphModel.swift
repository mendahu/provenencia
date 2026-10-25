import Foundation
import Observation

/// Create / place / drag / connect logic for the Evidence graph (S6-03 / S6-04).
@MainActor
@Observable
final class EvidenceGraphModel {
    struct CreateDraft: Equatable {
        var label: String = ""
        var description: String = ""
    }

    /// What the Unpromote confirm names: the subject, and the bridge sentences
    /// that come off with it. Built fresh on open, so `error` starts nil.
    struct PendingUnpromote: Identifiable, Equatable {
        /// The subject's id.
        let id: String
        let name: String
        let bridges: [String]
        var error: String?
    }

    /// The Evidence graph this canvas shows. Its Source reads (`sourceRow`,
    /// citation counts, the Source page) stay on `sourceID`.
    let graphID: String
    let sourceID: String
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String

    var armedKind: EvidencePrimaryKind?
    /// Connect tool armed (mutually exclusive with `armedKind`).
    var armedConnect = false
    /// Select tool armed from the palette: an empty-canvas drag draws a rectangle,
    /// and its release disarms it (exclusive with Place and Connect). Holding Shift
    /// engages the same mode without arming it.
    var armedSelect = false
    /// Shift is down over the graph. Engages Select without arming it.
    private(set) var shiftSelectHeld = false
    /// First primary chosen while connecting.
    var connectOriginID: String?

    var isCreating = false
    /// Edit label/description sheet open for this subject (mutually exclusive with create).
    var editingSubjectID: String?
    /// Kind used for edit dialog title when editing a primary; nil when editing a bridge.
    var editingPrimaryKind: EvidencePrimaryKind?
    var editingBridgeKind: EvidenceBridgeKind?
    /// After a valid pair, consumed once to open the composer.
    private(set) var pendingComposerHandoff: WorkspaceLocation?
    private var positionGeneration: [String: Int] = [:]
    private var confirmedPositions: [String: CatalogGridCell] = [:]
    /// Store write started by the latest `commitDrag`; tests await it instead of sleeping.
    @ObservationIgnored
    private(set) var lastPositionWrite: Task<Void, Never>?
    /// Cards a drag in flight moves: the anchor first, then the other selected
    /// subjects in snapshot order. Cleared when the drop commits.
    private(set) var groupDragIDs: [String] = []
    @ObservationIgnored
    private var snapshotMemo: (
        rows: EvidenceGraphRows,
        types: [CatalogSubjectType],
        rules: [CatalogConnectRule],
        orders: NameFormatOrders,
        value: EvidenceGraphSnapshot,
        text: EvidenceGraphText
    )?
    /// Subject whose filing the Unpromote confirm names (`.pvConfirm(item:)`).
    var pendingUnpromote: PendingUnpromote?
    var isUnpromoting = false
    /// Subject pending DeleteImpact confirm or notice.
    let deleteImpact = DeleteImpactFlow()
    var pendingImpact: PVDeleteImpactRequest? {
        get { deleteImpact.request }
        set { deleteImpact.applyRequest(newValue) }
    }
    var isDeleting: Bool { deleteImpact.isRunning }
    var deleteError: String? { deleteImpact.error }
    var pendingGridX: Int64 = 0
    var pendingGridY: Int64 = 0
    var draft = CreateDraft()

    /// Differs from the seed: the default create label, or the subject as it
    /// was when Edit opened. An untouched sheet is clean.
    var formSheetIsDirty: Bool {
        draft != formSheetBaseline
    }
    var isSaving = false
    var labelError: String?
    var createError: String?
    var hoverGridX: Int64?
    var hoverGridY: Int64?
    /// Grid-space point for the connect rubber-band (under cursor).
    var connectHoverPoint: CGPoint?
    /// Selected cards. A click or a rotor pick holds one; a rectangle holds several.
    private(set) var selectedSubjectIDs: Set<String> = []
    /// The only selected card, or nil when none or several are selected. Return,
    /// Space, and the bridge highlight read this.
    var selectedSubjectID: String? {
        selectedSubjectIDs.count == 1 ? selectedSubjectIDs.first : nil
    }
    /// Space/Return opened this card for inner controls (cleared on Esc / other select).
    var activatedSubjectID: String?
    /// The Artifact file the header select shows. May name a file that is gone;
    /// read ``resolvedFileArtifactID``.
    var selectedFileArtifactID = ""
    /// Set when Open could not find the selected file. Cleared by choosing a file.
    var fileOpenError: String?
    /// Shown when drag / arrow persist fails after an optimistic patch.
    var toast: VocabularyToast?
    /// Canvas while a drag has carried a card past the resting document.
    /// Cleared when the drop lands; the snapshot's own fit takes over.
    var liveCanvas: EvidenceGraphCanvas?

    /// Seeded type ids keyed by primary or bridge kind rawValue.
    var typeIDByKind: [String: String] {
        Dictionary(uniqueKeysWithValues: placeableTypes.map { ($0.key, $0.id) })
    }

    var typeLabelByKind: [String: String] {
        Dictionary(uniqueKeysWithValues: placeableTypes.map { ($0.key, $0.label) })
    }

    /// Registry presentation tokens keyed by type key (S7-09).
    var presentationByKind: [String: CatalogSubjectTypePresentation] {
        fieldsSnapshot?.presentationsByKey ?? [:]
    }

    /// False when this graph links no Artifact — disables cite controls (S7-09).
    /// True while the graph list has no value (unknown); a graph missing from a
    /// loaded list cannot be cited.
    var canCite: Bool {
        guard let graph = session.evidenceGraph(id: graphID) else {
            return session.evidenceGraphsHandle?.value == nil
        }
        return graph.hasArtifact
    }

    var connectRules: [CatalogConnectRule] {
        let handle: QueryHandle<[CatalogConnectRule]>? = session.queryHandle(connectRulesKey)
        return handle?.value ?? []
    }

    var connectRulesError: String? {
        guard let handle: QueryHandle<[CatalogConnectRule]> = session.queryHandle(connectRulesKey),
              handle.status == .error
        else { return nil }
        return handle.error.map { L10n.Errors.message(for: $0) }
    }

    var isSheetPresented: Bool {
        isCreating || editingSubjectID != nil
    }

    var isEditingBridge: Bool { editingBridgeKind != nil }

    var sourceTitle: String { sourceRow?.title ?? "" }

    var sourceRef: String? {
        let ref = sourceRow?.ref.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        return ref.isEmpty ? nil : ref
    }

    private var graphKey: CatalogQueryKey {
        CatalogQueryKey.evidenceGraph(project: session.projectKey, graphId: graphID)
    }

    private var fieldsKey: CatalogQueryKey {
        CatalogQueryKey.propertiesWorkspace(project: session.projectKey)
    }

    private var sourcesListKey: CatalogQueryKey {
        CatalogQueryKey.sourcesList(project: session.projectKey)
    }

    private var connectRulesKey: CatalogQueryKey {
        CatalogQueryKey.connectRules(project: session.projectKey)
    }

    private var workspaceKey: CatalogQueryKey {
        CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: sourceID)
    }

    private var citationCountsKey: CatalogQueryKey {
        CatalogQueryKey.citationCounts(project: session.projectKey, graphId: graphID)
    }

    private var fieldsSnapshot: PropertiesSnapshot? {
        let handle: QueryHandle<PropertiesSnapshot>? = session.queryHandle(fieldsKey)
        return handle?.value
    }

    private var sourceRow: CatalogSource? { session.source(id: sourceID) }

    private var placeableTypes: [CatalogSubjectType] {
        (fieldsSnapshot?.types ?? []).filter {
            EvidencePrimaryKind(rawValue: $0.key) != nil
                || EvidenceBridgeKind(rawValue: $0.key) != nil
        }
    }

    init(
        graphID: String,
        sourceID: String,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String
    ) {
        self.graphID = graphID
        self.sourceID = sourceID
        self.session = session
        self.store = store
        self.userID = userID
    }

    /// Snapshot as loaded. Edges come from cited observations.
    func displaySnapshot(rows: EvidenceGraphRows?, types: [CatalogSubjectType]) -> EvidenceGraphSnapshot {
        let rules = connectRules
        let normalized = rows ?? EvidenceGraphRows(graphId: graphID)
        if let snapshotMemo,
           snapshotMemo.rows == normalized,
           snapshotMemo.types == types,
           snapshotMemo.rules == rules,
           snapshotMemo.orders == session.nameFormat
        {
            return snapshotMemo.value
        }
        let value = EvidenceGraphSnapshot.build(
            rows: normalized,
            types: types,
            rules: rules
        )
        snapshotMemo = (normalized, types, rules, session.nameFormat, value, EvidenceGraphText(snapshot: value, orders: session.nameFormat))
        return value
    }

    /// Card copy for the same snapshot as ``displaySnapshot(rows:types:)``,
    /// formatted once when that snapshot is built.
    func displayText(rows: EvidenceGraphRows?, types: [CatalogSubjectType]) -> EvidenceGraphText {
        _ = displaySnapshot(rows: rows, types: types)
        return snapshotMemo?.text ?? EvidenceGraphText(snapshot: displaySnapshot(rows: rows, types: types), orders: session.nameFormat)
    }

    func prepare() async {
        let _: QueryHandle<EvidenceGraphRows> = session.query(graphKey)
        let rules: QueryHandle<[CatalogConnectRule]> = session.query(connectRulesKey)
        let _: QueryHandle<PropertiesSnapshot> = session.query(fieldsKey)
        let _: QueryHandle<[CatalogSource]> = session.query(sourcesListKey)
        let _: QueryHandle<[CatalogEvidenceGraph]> = session.query(session.evidenceGraphsKey)
        let _: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
        let _: QueryHandle<[String: Int]> = session.query(citationCountsKey)
        _ = await session.readyValue(graphKey) as EvidenceGraphRows?
        _ = await session.readyValue(fieldsKey) as PropertiesSnapshot?
        _ = await session.readyValue(connectRulesKey) as [CatalogConnectRule]?
        _ = await session.readyValue(sourcesListKey) as [CatalogSource]?
        _ = await session.readyValue(session.evidenceGraphsKey) as [CatalogEvidenceGraph]?
        _ = await session.readyValue(workspaceKey) as CatalogSourceWorkspace?
        _ = await session.readyValue(citationCountsKey) as [String: Int]?
        if rules.status == .error, let message = connectRulesError {
            toast = VocabularyToast(
                title: L10n.string(L10n.EvidenceGraph.connectRulesUnavailableTitle),
                body: message,
                tone: .danger
            )
        }
        if !canCite {
            disarmCiteTools()
        }
        if let id = session.consumePendingGraphSelection(graphID: graphID) {
            selectedSubjectIDs = [id]
        }
    }

    // MARK: Files (S10-20)

    /// The Artifacts this graph links that have a file, in page order. Empty
    /// until the workspace loads and when it fails to, and while the graph list
    /// has no value.
    var fileRows: [EvidenceGraphFiles.Row] {
        guard let handle: QueryHandle<CatalogSourceWorkspace> = session.queryHandle(workspaceKey),
              handle.status == .ready,
              let workspace = handle.value
        else { return [] }
        return EvidenceGraphFiles.fileBearing(
            session.linkedArtifacts(graphID: graphID, artifacts: workspace.artifacts) ?? [],
            coverArtifactID: sourceRow?.coverArtifactID
        )
    }

    /// The selected file when it is still listed, otherwise the cover or the first file.
    var resolvedFileArtifactID: String {
        let rows = fileRows
        if rows.contains(where: { $0.id == selectedFileArtifactID }) {
            return selectedFileArtifactID
        }
        return EvidenceGraphFiles.initialID(in: rows, coverArtifactID: sourceRow?.coverArtifactID) ?? ""
    }

    func citationCount(for artifactID: String) -> Int {
        let handle: QueryHandle<[String: Int]>? = session.queryHandle(citationCountsKey)
        return handle?.value?[artifactID] ?? 0
    }

    func selectFile(_ id: String) {
        selectedFileArtifactID = id
        fileOpenError = nil
    }

    /// Opens the selected file in its own app. Leaves the card selection and camera alone.
    func openSelectedFile() {
        let id = resolvedFileArtifactID
        guard let row = fileRows.first(where: { $0.id == id }) else { return }
        selectedFileArtifactID = id
        fileOpenError = nil
        if !ProjectFiles.openObject(projectDir: session.projectKey.projectDir, relPath: row.relPath) {
            fileOpenError = L10n.EvidenceGraph.fileNotFoundMessage(
                name: row.title,
                filename: row.displayFilename
            )
        }
    }

    func presentation(for typeKey: String) -> CatalogSubjectTypePresentation? {
        presentationByKind[typeKey]
    }

    func toggleArm(_ kind: EvidencePrimaryKind) {
        guard canCite else { return }
        if armedKind == kind {
            disarm()
        } else {
            clearConnectState()
            armedSelect = false
            fileOpenError = nil
            armedKind = kind
            hoverGridX = nil
            hoverGridY = nil
        }
    }

    func toggleConnect() {
        guard canCite else { return }
        if let connectRulesError {
            toast = VocabularyToast(
                title: L10n.string(L10n.EvidenceGraph.connectRulesUnavailableTitle),
                body: connectRulesError,
                tone: .danger
            )
            return
        }
        if armedConnect {
            disarm()
        } else {
            armedKind = nil
            armedSelect = false
            fileOpenError = nil
            hoverGridX = nil
            hoverGridY = nil
            armedConnect = true
            connectOriginID = nil
            connectHoverPoint = nil
        }
    }

    /// Arms or disarms Select. Available when the Source cannot be cited:
    /// selecting cards writes nothing to the catalog. Arming it disarms Place and Connect.
    func toggleSelect() {
        if armedSelect {
            armedSelect = false
        } else {
            disarmCiteTools()
            fileOpenError = nil
            armedSelect = true
        }
    }

    /// Disarms every tool.
    func disarm() {
        disarmCiteTools()
        armedSelect = false
    }

    /// Disarms Place and Connect and leaves Select as it is.
    func disarmCiteTools() {
        armedKind = nil
        hoverGridX = nil
        hoverGridY = nil
        clearConnectState()
    }

    private func clearConnectState() {
        armedConnect = false
        connectOriginID = nil
        connectHoverPoint = nil
    }

    /// Replaces the selection with one card (a bridge is allowed), or empties it.
    /// Leaves the Select tool as it was.
    func selectSubject(id: String?) {
        selectedSubjectIDs = id.map { [$0] } ?? []
        if activatedSubjectID != id {
            activatedSubjectID = nil
        }
    }

    /// Replaces the selection with the cards among `ids` that are on this graph,
    /// including relationship bridges.
    func replacePrimarySelection(_ ids: Set<String>) {
        let snapshot = currentSnapshot()
        let cardIDs = Set(
            (snapshot?.subjects ?? []).map(\.id) + (snapshot?.bridges ?? []).map(\.id)
        )
        selectedSubjectIDs = ids.intersection(cardIDs)
        if let activated = activatedSubjectID, !selectedSubjectIDs.contains(activated) {
            activatedSubjectID = nil
        }
    }

    func setShiftSelectHeld(_ held: Bool) {
        shiftSelectHeld = held
    }

    /// Shift-click: adds a card (a bridge is allowed) to the selection, or removes it.
    func toggleSubjectInSelection(id: String) {
        var next = selectedSubjectIDs
        if next.contains(id) {
            next.remove(id)
        } else {
            next.insert(id)
        }
        replacePrimarySelection(next)
    }

    /// A rectangle was released: replaces the selection with its cards, or adds
    /// them when Shift was down. One rectangle per arm, even one that caught nothing.
    func finishRectangleSelection(_ ids: Set<String>, additive: Bool) {
        replacePrimarySelection(additive ? selectedSubjectIDs.union(ids) : ids)
        armedSelect = false
    }

    func activateSubject(id: String) {
        selectedSubjectIDs = [id]
        activatedSubjectID = id
    }

    func deactivateSubject() {
        activatedSubjectID = nil
    }

    enum Key: Equatable {
        case escape
        case `return`
        case space
        case arrow(dx: Int64, dy: Int64)
    }

    struct KeyEffect: Equatable {
        var handled = false
        var clearFocus = false
        /// Cells to move, in snapshot subject order. Empty when nothing moves.
        var moves: [SubjectMove] = []

        struct SubjectMove: Equatable {
            var subjectID: String
            var fromX: Int64
            var fromY: Int64
            var deltaX: Int64
            var deltaY: Int64
        }
    }

    /// Keyboard policy for the graph. The view forwards `KeyPress` and applies focus.
    func handleKey(_ key: Key, hasFocus: Bool) -> KeyEffect {
        if isSheetPresented { return KeyEffect() }
        switch key {
        case .escape:
            if armedConnect || armedKind != nil {
                disarm()
                return KeyEffect(handled: true)
            }
            if armedSelect {
                armedSelect = false
                return KeyEffect(handled: true)
            }
            if activatedSubjectID != nil {
                deactivateSubject()
                return KeyEffect(handled: true)
            }
            if !selectedSubjectIDs.isEmpty || hasFocus {
                selectSubject(id: nil)
                return KeyEffect(handled: true, clearFocus: true)
            }
            return KeyEffect()
        case .return, .space:
            guard inputMode == .idle, let selectedID = selectedSubjectID else {
                return KeyEffect()
            }
            if key == .return {
                beginEdit(subjectID: selectedID)
            } else {
                activateSubject(id: selectedID)
            }
            return KeyEffect(handled: true)
        case .arrow(let dx, let dy):
            guard allowsCardMoves, let snapshot = currentSnapshot() else { return KeyEffect() }
            // Subjects first, then bridges, in snapshot order.
            let selected = selectedSubjectIDs
            var moves = snapshot.subjects
                .filter { selected.contains($0.id) }
                .map {
                    KeyEffect.SubjectMove(
                        subjectID: $0.id, fromX: $0.gridX, fromY: $0.gridY, deltaX: dx, deltaY: dy
                    )
                }
            moves += snapshot.bridges
                .filter { selected.contains($0.id) }
                .map {
                    KeyEffect.SubjectMove(
                        subjectID: $0.id, fromX: $0.gridX, fromY: $0.gridY, deltaX: dx, deltaY: dy
                    )
                }
            guard !moves.isEmpty else { return KeyEffect() }
            return KeyEffect(handled: true, moves: moves)
        }
    }

    /// Card action id from the pointer layer. Returns a composer place when the action navigates.
    func performCardAction(subjectID: String, actionID: String) -> WorkspaceLocation? {
        switch actionID {
        case EvidenceSubjectCard.editActionID, EvidenceBridgeCard.editActionID:
            beginEdit(subjectID: subjectID)
            return nil
        case EvidenceBridgeCard.editCitationActionID:
            return composerLocationForBridgeCitation(subjectID: subjectID)
        case EvidenceSubjectCard.deleteActionID, EvidenceBridgeCard.deleteActionID:
            Task { await beginDelete(subjectID: subjectID) }
            return nil
        case EvidenceSubjectCard.addPropertyActionID:
            return composerLocation(for: subjectID)
        case EvidenceSubjectCard.promoteActionID:
            return promoteLocation(for: subjectID)
        case EvidenceSubjectCard.openHandleActionID:
            return openHandle(subjectID: subjectID)
        case EvidenceSubjectCard.unpromoteActionID:
            beginUnpromote(subjectID: subjectID)
            return nil
        default:
            guard let observationID = EvidenceSubjectCard.observationID(fromEditPropertyAction: actionID)
            else { return nil }
            return composerLocation(forObservationID: observationID, subjectID: subjectID)
        }
    }

    func gridCell(for subjectID: String) -> (x: Int64, y: Int64)? {
        let snapshot = currentSnapshot()
        if let placed = snapshot?.subjects.first(where: { $0.id == subjectID }) {
            return (placed.gridX, placed.gridY)
        }
        if let bridge = snapshot?.bridges.first(where: { $0.id == subjectID }) {
            return (bridge.gridX, bridge.gridY)
        }
        return nil
    }

    /// Connect pick resolved from the current snapshot (primaries only).
    func handleConnectPick(subjectID: String) async {
        guard let placed = currentSnapshot()?.subjects.first(where: { $0.id == subjectID }) else { return }
        await handleConnectPick(subjectID: placed.id, kind: placed.kind, label: placed.subject.label)
    }

    /// Pointer / keyboard pick while Connect is armed.
    func handleConnectPick(subjectID: String, kind: EvidencePrimaryKind, label: String) async {
        guard canCite, armedConnect, !isCreating, !isSaving, editingSubjectID == nil
        else { return }
        if connectOriginID == nil {
            connectOriginID = subjectID
            selectSubject(id: subjectID)
            return
        }
        guard let originID = connectOriginID, originID != subjectID else { return }
        await completeConnectPair(
            originID: originID,
            targetID: subjectID,
            targetKind: kind,
            targetLabel: label
        )
    }

    func updateHover(gridPoint: CGPoint) {
        guard armedKind != nil, !isCreating else { return }
        let cell = GraphCanvasGridMapping.gridCell(gridPoint: gridPoint)
        hoverGridX = cell.gridX
        hoverGridY = cell.gridY
    }

    func updateConnectHover(gridPoint: CGPoint) {
        guard armedConnect, connectOriginID != nil, !isCreating else {
            connectHoverPoint = nil
            return
        }
        connectHoverPoint = gridPoint
    }

    func clearHover() {
        hoverGridX = nil
        hoverGridY = nil
        connectHoverPoint = nil
    }

    /// Opens the create dialog at the snapped cell (tool must be armed).
    func beginCreate(at gridPoint: CGPoint) {
        guard canCite, let kind = armedKind, !isCreating, !isSaving, editingSubjectID == nil else { return }
        let cell = GraphCanvasGridMapping.gridCell(gridPoint: gridPoint)
        pendingGridX = cell.gridX
        pendingGridY = cell.gridY
        draft = CreateDraft(
            label: defaultLabel(for: kind),
            description: ""
        )
        labelError = nil
        createError = nil
        isCreating = true
        clearHover()
    }

    /// Opens the shared create/edit sheet prefilled for an existing subject.
    func beginEdit(subjectID: String) {
        guard !isCreating, !isSaving, editingSubjectID == nil else { return }
        let snapshot = currentSnapshot()
        if let primary = primary(in: snapshot, id: subjectID) {
            editingSubjectID = subjectID
            editingPrimaryKind = primary.kind
            editingBridgeKind = nil
            draft = CreateDraft(
                label: primary.subject.label,
                description: primary.subject.description
            )
        } else if let bridge = snapshot?.bridges.first(where: { $0.id == subjectID }) {
            editingSubjectID = subjectID
            editingPrimaryKind = nil
            editingBridgeKind = bridge.kind
            draft = CreateDraft(
                label: bridge.subject.label,
                description: bridge.subject.description
            )
        } else {
            return
        }
        selectSubject(id: subjectID)
        labelError = nil
        createError = nil
        disarm()
    }

    func completeConnectPair(
        originID: String,
        targetID: String,
        targetKind: EvidencePrimaryKind,
        targetLabel: String
    ) async {
        guard canCite else { return }
        guard let origin = primary(in: currentSnapshot(), id: originID) else { return }
        let rule = CatalogConnectRule.match(
            from: origin.kind.rawValue,
            to: targetKind.rawValue,
            in: connectRules
        )
        if rule.refuse || rule.bridgeTypeKey.isEmpty {
            toast = VocabularyToast(
                title: L10n.string(L10n.EvidenceGraph.connectInvalidPairTitle),
                body: L10n.string(L10n.EvidenceGraph.connectInvalidPairBody),
                tone: .danger
            )
            return
        }

        connectHoverPoint = nil
        pendingComposerHandoff = connectComposerLocation(
            fromID: originID,
            toID: targetID,
            fromKind: origin.kind,
            toKind: targetKind,
            fromLabel: origin.subject.label,
            toLabel: targetLabel,
            rule: rule
        )
    }

    func cancelSheet() {
        if editingSubjectID != nil {
            cancelEdit()
        } else {
            cancelCreate()
        }
    }

    func cancelEdit() {
        guard !isSaving else { return }
        editingSubjectID = nil
        editingPrimaryKind = nil
        editingBridgeKind = nil
        labelError = nil
        createError = nil
    }

    func cancelCreate() {
        guard !isSaving else { return }
        isCreating = false
        labelError = nil
        createError = nil
        disarm()
    }

    /// Creates a primary subject + position.
    @discardableResult
    func confirmCreate() async -> String? {
        await confirmPrimaryCreate()
    }

    @discardableResult
    func confirmEdit() async -> String? {
        guard let subjectID = editingSubjectID, !isSaving else { return nil }
        let snapshot = currentSnapshot()
        let storedLabel: String
        if let bridge = snapshot?.bridges.first(where: { $0.id == subjectID }) {
            storedLabel = bridge.subject.label
        } else {
            let trimmed = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
            if trimmed.isEmpty {
                labelError = L10n.string(L10n.EvidenceGraph.labelRequired)
                return nil
            }
            storedLabel = trimmed
        }

        isSaving = true
        defer { isSaving = false }
        labelError = nil
        createError = nil
        do {
            let (_, effects) = try await store.updateSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectID: subjectID,
                label: storedLabel,
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            session.apply(effects)
            editingSubjectID = nil
            editingPrimaryKind = nil
            editingBridgeKind = nil
            selectSubject(id: subjectID)
            return subjectID
        } catch {
            createError = L10n.Errors.message(for: error)
            return nil
        }
    }

    /// Citation composer place for Add property.
    func composerLocation(for subjectID: String) -> WorkspaceLocation? {
        composerLocation(for: subjectID, citationID: nil)
    }

    /// Consumes a Connect handoff into the citation composer (subjectId is nil).
    func consumeComposerHandoff() -> WorkspaceLocation? {
        guard canCite, let handoff = pendingComposerHandoff else { return nil }
        pendingComposerHandoff = nil
        return handoff
    }

    /// Citation composer place for editing an existing citation (property row pencil).
    func composerLocation(
        for subjectID: String,
        citationID: String?,
        observationID: String? = nil,
        artifactID: String? = nil
    ) -> WorkspaceLocation? {
        guard canCite else { return nil }
        let snapshot = currentSnapshot()
        let title: String?
        let ref: String?
        if let primary = primary(in: snapshot, id: subjectID) {
            title = primary.subject.label
            ref = primary.subject.ref
        } else if let snapshot, let bridge = snapshot.bridges.first(where: { $0.id == subjectID }) {
            title = EvidenceBridgeEdgeSummary.sentence(for: bridge, in: snapshot, orders: session.nameFormat)
            ref = bridge.subject.ref
        } else {
            title = nil
            ref = nil
        }
        return .citationComposer(
            sourceId: sourceID,
            graphId: graphID,
            subjectId: subjectID,
            citationId: citationID,
            artifactId: artifactID,
            observationId: observationID,
            ref: ref,
            title: title,
            sourceTitle: sourceRow?.title
        )
    }

    /// Opens the citation composer for the Observation's citation (shared across rows).
    func composerLocation(forObservationID observationID: String, subjectID: String) -> WorkspaceLocation? {
        let snapshot = currentSnapshot()
        if let primary = primary(in: snapshot, id: subjectID),
           let observation = primary.observations.first(where: { $0.id == observationID })
        {
            return composerLocation(
                for: subjectID,
                citationID: observation.citationID,
                observationID: observation.id
            )
        }
        if let bridge = snapshot?.bridges.first(where: { $0.id == subjectID }),
           let observation = bridge.observations.first(where: { $0.id == observationID })
        {
            return composerLocation(
                for: subjectID,
                citationID: observation.citationID,
                observationID: observation.id
            )
        }
        return nil
    }

    /// Opens the citation that cites this bridge's edge Observations (or a new cite).
    func composerLocationForBridgeCitation(subjectID: String) -> WorkspaceLocation? {
        let snapshot = currentSnapshot()
        guard let bridge = snapshot?.bridges.first(where: { $0.id == subjectID }) else {
            return nil
        }
        let first = bridge.observations.first
        return composerLocation(
            for: subjectID,
            citationID: first?.citationID,
            observationID: first?.id
        )
    }

    /// Fetches Impact and presents the DeleteImpact recipe. The card never gates.
    func beginDelete(subjectID: String) async {
        let snapshot = currentSnapshot()
        let title: String
        let ref: String
        if let primary = primary(in: snapshot, id: subjectID) {
            title = primary.subject.label.isEmpty ? primary.subject.ref : primary.subject.label
            ref = primary.subject.ref
        } else if let snapshot, let bridge = snapshot.bridges.first(where: { $0.id == subjectID }) {
            title = EvidenceBridgeEdgeSummary.sentence(for: bridge, in: snapshot, orders: session.nameFormat)
            ref = bridge.subject.ref
        } else {
            return
        }
        await deleteImpact.ask(kind: "subject", id: subjectID, ref: ref, title: title) {
            try await self.store.getDeleteImpact(
                projectDir: self.session.projectKey.projectDir,
                kind: "subject",
                id: subjectID
            )
        }
    }

    // MARK: Promote (S9-04 v1)

    /// The Promote place for an unpromoted primary (S9-11): the card's Promote
    /// opens the flow, which files the subject and returns here. Nil for a
    /// promoted or unknown subject.
    func promoteLocation(for subjectID: String) -> WorkspaceLocation? {
        guard let placed = primary(in: currentSnapshot(), id: subjectID),
              placed.membership == nil
        else { return nil }
        return .promote(
            sourceId: sourceID,
            graphId: graphID,
            subjectId: placed.id,
            kind: placed.kind,
            ref: placed.subject.ref,
            title: placed.displayName(orders: session.nameFormat),
            sourceTitle: sourceRow?.title
        )
    }

    /// Promote with every subject expanded. Works after the cards are filed,
    /// so a relationship added later can still be filed from this page.
    /// Nil when the graph has no primary subjects.
    func promoteAllLocation() -> WorkspaceLocation? {
        let subjects = currentSnapshot()?.subjects ?? []
        let placed = subjects
            .filter { $0.membership == nil }
            .min { $0.id < $1.id }
            ?? subjects.min { $0.id < $1.id }
        guard let placed else { return nil }
        return .promote(
            sourceId: sourceID,
            graphId: graphID,
            subjectId: placed.id,
            kind: placed.kind,
            ref: placed.subject.ref,
            title: placed.displayName(orders: session.nameFormat),
            sourceTitle: sourceRow?.title,
            mapAll: true
        )
    }

    // MARK: Unpromote (S10-44)

    /// Opens the Unpromote confirm for a promoted primary. The bridge list is
    /// the inverse of Promote's forecast, read from the same snapshot fields:
    /// filed bridges that have this subject as an end. The engine recomputes
    /// the authoritative set in the write. Does nothing otherwise.
    func beginUnpromote(subjectID: String) {
        guard let snapshot = currentSnapshot(),
              let placed = primary(in: snapshot, id: subjectID),
              placed.membership != nil
        else { return }
        let bridges = snapshot.bridges
            .filter { ($0.endpointAID == subjectID || $0.endpointBID == subjectID) && $0.membership != nil }
            .sorted { $0.id < $1.id }
            .map { EvidenceBridgeEdgeSummary.sentence(for: $0, in: snapshot, orders: session.nameFormat) }
        pendingUnpromote = PendingUnpromote(
            id: placed.id,
            name: placed.displayName(orders: session.nameFormat),
            bridges: bridges
        )
    }

    /// Writes the Unpromote. Success refreshes through the write's effects and
    /// closes the confirm; a failure stays in the confirm as its error.
    func confirmUnpromote() async {
        guard let pending = pendingUnpromote, !isUnpromoting else { return }
        isUnpromoting = true
        defer { isUnpromoting = false }
        do {
            let effects = try await store.unpromoteSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectID: pending.id
            )
            session.apply(effects)
            pendingUnpromote = nil
        } catch {
            pendingUnpromote?.error = L10n.Errors.message(for: error)
        }
    }

    /// Location of a promoted subject's handle page. Persons, Events, and
    /// Places open their pages.
    func openHandle(subjectID: String) -> WorkspaceLocation? {
        guard let placed = primary(in: currentSnapshot(), id: subjectID),
              let membership = placed.membership
        else { return nil }
        switch membership.kind {
        case "person":
            return .personDetail(
                entityId: membership.entity.id,
                ref: membership.entity.ref,
                title: EvidenceSubjectCard.membershipName(membership, orders: session.nameFormat)
            )
        case "event":
            return .eventDetail(
                entityId: membership.entity.id,
                ref: membership.entity.ref,
                title: placed.displayName(orders: session.nameFormat)
            )
        case "place":
            return .placeDetail(
                entityId: membership.entity.id,
                ref: membership.entity.ref,
                title: placed.displayName(orders: session.nameFormat)
            )
        default:
            return nil
        }
    }

    @discardableResult
    func confirmPendingImpact() async -> Bool {
        return await deleteImpact.confirm { target in
            let effects = try await store.deleteSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectID: target.id
            )
            session.apply(effects, .deletedSubject)
            selectedSubjectIDs.remove(target.id)
            if activatedSubjectID == target.id {
                activatedSubjectID = nil
            }
        }
    }

    /// Source page recovery target when the graph has no Artifacts.
    func sourcePageLocation(title: String, ref: String?) -> WorkspaceLocation {
        .sourcePage(sourceId: sourceID, ref: ref, title: title)
    }

    @discardableResult
    private func confirmPrimaryCreate() async -> String? {
        guard let kind = armedKind, !isSaving else { return nil }
        let trimmed = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty {
            labelError = L10n.string(L10n.EvidenceGraph.labelRequired)
            return nil
        }
        guard let typeID = typeIDByKind[kind.rawValue] else {
            createError = L10n.string(L10n.EvidenceGraph.typesUnavailable)
            return nil
        }

        isSaving = true
        defer { isSaving = false }
        labelError = nil
        createError = nil
        do {
            let (created, effects) = try await store.createSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                graphID: graphID,
                subjectTypeID: typeID,
                label: trimmed,
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines),
                placement: CatalogGridCell(gridX: pendingGridX, gridY: pendingGridY)
            )
            session.apply(effects)
            isCreating = false
            disarm()
            return created.id
        } catch {
            createError = L10n.Errors.message(for: error)
            return nil
        }
    }

    private func connectComposerLocation(
        fromID: String,
        toID: String,
        fromKind: EvidencePrimaryKind,
        toKind: EvidencePrimaryKind,
        fromLabel: String,
        toLabel: String,
        rule: CatalogConnectRule
    ) -> WorkspaceLocation {
        let sentence = ConnectEndpointBinding.pendingSentence(
            rule: rule,
            fromID: fromID,
            fromTypeKey: fromKind.rawValue,
            fromLabel: fromLabel,
            toID: toID,
            toTypeKey: toKind.rawValue,
            toLabel: toLabel
        )
        return .citationComposer(
            sourceId: sourceID,
            graphId: graphID,
            connectFromSubjectId: fromID,
            connectToSubjectId: toID,
            connectBridgeTypeKey: rule.bridgeTypeKey,
            title: sentence,
            sourceTitle: sourceRow?.title
        )
    }

    /// Cards move (drag, arrow keys) while no sheet is up and no Place or
    /// Connect tool is armed. Select armed still moves them.
    private var allowsCardMoves: Bool {
        inputMode == .idle || inputMode == .selecting
    }

    /// Starts a card drag. Keeps the selection when the anchor is a selected
    /// card (a person, event, place, or bridge), otherwise the anchor alone
    /// becomes the selection. Returns the cards that move, anchor included.
    @discardableResult
    func beginGroupDrag(anchor: String) -> Set<String> {
        let handle: QueryHandle<EvidenceGraphRows>? = session.queryHandle(graphKey)
        guard handle?.value != nil, let snapshot = currentSnapshot() else {
            selectSubject(id: anchor)
            groupDragIDs = [anchor]
            return [anchor]
        }
        let cardIDs = snapshot.subjects.map(\.id) + snapshot.bridges.map(\.id)
        if cardIDs.contains(anchor), selectedSubjectIDs.contains(anchor) {
            groupDragIDs = [anchor] + cardIDs.filter {
                $0 != anchor && selectedSubjectIDs.contains($0)
            }
        } else {
            selectSubject(id: anchor)
            groupDragIDs = [anchor]
        }
        return Set(groupDragIDs)
    }

    /// Snaps a completed card drag. Patches the session cache **synchronously**
    /// so `@GestureState` can clear onto the new cell, then writes the store.
    @discardableResult
    func commitDrag(
        subjectID: String,
        originGridX: Int64,
        originGridY: Int64,
        delta: CGSize
    ) -> (gridX: Int64, gridY: Int64)? {
        guard allowsCardMoves else { return nil }
        let snapped = snapAndPatch(
            subjectID: subjectID,
            originGridX: originGridX,
            originGridY: originGridY,
            delta: delta
        )
        persistPositions([
            PositionWrite(
                subjectID: subjectID,
                generation: snapped.generation,
                gridX: snapped.gridX,
                gridY: snapped.gridY
            ),
        ])
        return (snapped.gridX, snapped.gridY)
    }

    /// Snaps a completed drag of the anchor and every card that moved with it,
    /// each from its own cell by the same document delta. One write task
    /// persists them in ``groupDragIDs`` order.
    @discardableResult
    func commitGroupDrag(
        anchor: String,
        delta: CGSize
    ) -> [(subjectID: String, gridX: Int64, gridY: Int64)] {
        let ids = groupDragIDs.isEmpty ? [anchor] : groupDragIDs
        groupDragIDs = []
        guard allowsCardMoves else { return [] }
        var writes: [PositionWrite] = []
        for id in ids {
            guard let cell = gridCell(for: id) else { continue }
            let snapped = snapAndPatch(
                subjectID: id,
                originGridX: cell.x,
                originGridY: cell.y,
                delta: delta
            )
            writes.append(
                PositionWrite(
                    subjectID: id,
                    generation: snapped.generation,
                    gridX: snapped.gridX,
                    gridY: snapped.gridY
                )
            )
        }
        persistPositions(writes)
        return writes.map { ($0.subjectID, $0.gridX, $0.gridY) }
    }

    @discardableResult
    func moveSubject(
        subjectID: String,
        fromGridX: Int64,
        fromGridY: Int64,
        deltaX: Int64,
        deltaY: Int64
    ) async -> (gridX: Int64, gridY: Int64)? {
        guard allowsCardMoves else { return nil }
        let nextX = fromGridX + deltaX
        let nextY = fromGridY + deltaY
        let generation = bumpPositionGeneration(subjectID: subjectID)
        applyPositionPatch(subjectID: subjectID, gridX: nextX, gridY: nextY)
        let ok = await persistPositionOrRevert(
            subjectID: subjectID,
            generation: generation,
            gridX: nextX,
            gridY: nextY
        )
        return ok ? (nextX, nextY) : nil
    }

    /// Arrow-key move of every selected card by one cell step. Patches each
    /// cell synchronously, then one task persists them in order.
    func moveSubjects(_ moves: [KeyEffect.SubjectMove]) {
        guard allowsCardMoves else { return }
        var writes: [PositionWrite] = []
        for move in moves {
            let nextX = move.fromX + move.deltaX
            let nextY = move.fromY + move.deltaY
            let generation = bumpPositionGeneration(subjectID: move.subjectID)
            applyPositionPatch(subjectID: move.subjectID, gridX: nextX, gridY: nextY)
            writes.append(
                PositionWrite(
                    subjectID: move.subjectID,
                    generation: generation,
                    gridX: nextX,
                    gridY: nextY
                )
            )
        }
        persistPositions(writes)
    }

    func ghostPlacedSubject() -> EvidenceGraphPlacedSubject? {
        guard let kind = armedKind,
              let x = hoverGridX,
              let y = hoverGridY,
              !isCreating
        else { return nil }
        let label = typeLabelByKind[kind.rawValue] ?? kind.rawValue
        return EvidenceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "ghost",
                ref: "",
                sourceID: sourceID,
                subjectTypeID: typeIDByKind[kind.rawValue] ?? "",
                label: defaultLabel(for: kind),
                description: ""
            ),
            kind: kind,
            typeLabel: label,
            gridX: x,
            gridY: y,
            isCited: false
        )
    }

    func createDialogTitle() -> LocalizedStringResource {
        if editingSubjectID != nil {
            return editDialogTitle()
        }
        switch armedKind {
        case .person: return L10n.EvidenceGraph.addPersonTitle
        case .event: return L10n.EvidenceGraph.addEventTitle
        case .place: return L10n.EvidenceGraph.addPlaceTitle
        case .none: return L10n.EvidenceGraph.addPersonTitle
        }
    }

    func editDialogTitle() -> LocalizedStringResource {
        if let kind = editingPrimaryKind {
            switch kind {
            case .person: return L10n.EvidenceGraph.editPersonTitle
            case .event: return L10n.EvidenceGraph.editEventTitle
            case .place: return L10n.EvidenceGraph.editPlaceTitle
            }
        }
        if let kind = editingBridgeKind {
            switch kind {
            case .relationship: return L10n.EvidenceGraph.editRelationshipTitle
            case .participation: return L10n.EvidenceGraph.editParticipationTitle
            case .location: return L10n.EvidenceGraph.editLocationTitle
            case .placeRelationship: return L10n.EvidenceGraph.editPlaceRelationshipTitle
            }
        }
        return L10n.EvidenceGraph.editPersonTitle
    }

    func sheetConfirmLabel() -> LocalizedStringResource {
        editingSubjectID != nil
            ? L10n.EvidenceGraph.editConfirm
            : L10n.EvidenceGraph.createConfirm
    }

    func createDialogTitle(for kind: EvidencePrimaryKind) -> LocalizedStringResource {
        switch kind {
        case .person: L10n.EvidenceGraph.addPersonTitle
        case .event: L10n.EvidenceGraph.addEventTitle
        case .place: L10n.EvidenceGraph.addPlaceTitle
        }
    }

    func bridgeCreateTitle(for kind: EvidenceBridgeKind) -> LocalizedStringResource {
        switch kind {
        case .relationship: L10n.EvidenceGraph.addRelationshipTitle
        case .participation: L10n.EvidenceGraph.addParticipationTitle
        case .location: L10n.EvidenceGraph.addLocationTitle
        case .placeRelationship: L10n.EvidenceGraph.addPlaceRelationshipTitle
        }
    }

    /// Spoken label of a palette tool: `name, tool, on|off`.
    func toolAccessibilityLabel(name: LocalizedStringResource, armed: Bool) -> String {
        let state = armed
            ? L10n.string(L10n.EvidenceGraph.toolOn)
            : L10n.string(L10n.EvidenceGraph.toolOff)
        return "\(L10n.string(name)), \(L10n.string(L10n.EvidenceGraph.toolRole)), \(state)"
    }

    func toolName(for kind: EvidencePrimaryKind) -> LocalizedStringResource {
        switch kind {
        case .person: L10n.EvidenceGraph.toolPerson
        case .event: L10n.EvidenceGraph.toolEvent
        case .place: L10n.EvidenceGraph.toolPlace
        }
    }

    func armedHint(for kind: EvidencePrimaryKind) -> LocalizedStringResource {
        switch kind {
        case .person: L10n.EvidenceGraph.armedHintPerson
        case .event: L10n.EvidenceGraph.armedHintEvent
        case .place: L10n.EvidenceGraph.armedHintPlace
        }
    }

    var connectArmedHint: LocalizedStringResource {
        if connectOriginID != nil {
            return L10n.EvidenceGraph.armedHintConnectPickB
        }
        return L10n.EvidenceGraph.armedHintConnect
    }

    private var formSheetBaseline: CreateDraft {
        if let id = editingSubjectID {
            let snapshot = currentSnapshot()
            if let primary = primary(in: snapshot, id: id) {
                return CreateDraft(
                    label: primary.subject.label,
                    description: primary.subject.description
                )
            }
            if let bridge = snapshot?.bridges.first(where: { $0.id == id }) {
                return CreateDraft(
                    label: bridge.subject.label,
                    description: bridge.subject.description
                )
            }
        }
        if isCreating, let kind = armedKind {
            return CreateDraft(label: defaultLabel(for: kind), description: "")
        }
        return CreateDraft()
    }

    private func defaultLabel(for kind: EvidencePrimaryKind) -> String {
        switch kind {
        case .person: L10n.string(L10n.EvidenceGraph.defaultLabelPerson)
        case .event: L10n.string(L10n.EvidenceGraph.defaultLabelEvent)
        case .place: L10n.string(L10n.EvidenceGraph.defaultLabelPlace)
        }
    }

    private func currentSnapshot() -> EvidenceGraphSnapshot? {
        guard let handle: QueryHandle<EvidenceGraphRows> = session.queryHandle(graphKey) else {
            return nil
        }
        return displaySnapshot(rows: handle.value, types: fieldsSnapshot?.types ?? [])
    }

    private func primary(in snapshot: EvidenceGraphSnapshot?, id: String) -> EvidenceGraphPlacedSubject? {
        snapshot?.subjects.first { $0.id == id }
    }

    private struct PositionWrite {
        var subjectID: String
        var generation: Int
        var gridX: Int64
        var gridY: Int64
    }

    /// Drops a card at `origin + delta`, snapped to the grid, and patches
    /// the session cache. The store write is the caller's.
    private func snapAndPatch(
        subjectID: String,
        originGridX: Int64,
        originGridY: Int64,
        delta: CGSize
    ) -> (gridX: Int64, gridY: Int64, generation: Int) {
        let origin = GraphCanvasGridMapping.gridPoint(gridX: originGridX, gridY: originGridY)
        let dropped = CGPoint(
            x: origin.x + delta.width,
            y: origin.y + delta.height
        )
        let cell = GraphCanvasGridMapping.gridCell(gridPoint: dropped)
        let generation = bumpPositionGeneration(subjectID: subjectID)
        applyPositionPatch(subjectID: subjectID, gridX: cell.gridX, gridY: cell.gridY)
        return (cell.gridX, cell.gridY, generation)
    }

    /// One store task for a move: the writes persist in order.
    private func persistPositions(_ writes: [PositionWrite]) {
        lastPositionWrite = Task {
            for write in writes {
                _ = await persistPositionOrRevert(
                    subjectID: write.subjectID,
                    generation: write.generation,
                    gridX: write.gridX,
                    gridY: write.gridY
                )
            }
        }
    }

    private func bumpPositionGeneration(subjectID: String) -> Int {
        if confirmedPositions[subjectID] == nil, let cell = gridCell(for: subjectID) {
            confirmedPositions[subjectID] = CatalogGridCell(gridX: cell.x, gridY: cell.y)
        }
        let next = (positionGeneration[subjectID] ?? 0) + 1
        positionGeneration[subjectID] = next
        return next
    }

    private func applyPositionPatch(subjectID: String, gridX: Int64, gridY: Int64) {
        if let handle: QueryHandle<EvidenceGraphRows> = session.queryHandle(graphKey),
           let rows = handle.value
        {
            session.setQueryValue(
                graphKey,
                value: rows.updatingPosition(subjectID: subjectID, gridX: gridX, gridY: gridY)
            )
        }
    }

    @discardableResult
    private func persistPositionOrRevert(
        subjectID: String,
        generation: Int,
        gridX: Int64,
        gridY: Int64
    ) async -> Bool {
        do {
            _ = try await store.setSubjectPosition(
                projectDir: session.projectKey.projectDir,
                subjectID: subjectID,
                gridX: gridX,
                gridY: gridY
            )
            confirmedPositions[subjectID] = CatalogGridCell(gridX: gridX, gridY: gridY)
            if positionGeneration[subjectID] == generation {
                applyPositionPatch(subjectID: subjectID, gridX: gridX, gridY: gridY)
            }
            return true
        } catch {
            if positionGeneration[subjectID] == generation,
               let confirmed = confirmedPositions[subjectID]
            {
                applyPositionPatch(
                    subjectID: subjectID,
                    gridX: confirmed.gridX,
                    gridY: confirmed.gridY
                )
                toast = VocabularyToast(
                    title: L10n.string(L10n.EvidenceGraph.positionPersistFailedTitle),
                    body: L10n.Errors.message(for: error),
                    tone: .danger
                )
            }
            return false
        }
    }
}
