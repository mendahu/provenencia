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

    struct PendingDelete: Identifiable, Equatable {
        var id: String
        var label: String
        var ref: String
    }

    let sourceID: String
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String

    var armedKind: EvidencePrimaryKind?
    /// Connect tool armed (mutually exclusive with `armedKind`).
    var armedConnect = false
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
    @ObservationIgnored
    private var snapshotMemo: (
        rows: SourceGraphRows,
        types: [CatalogSubjectType],
        rules: [CatalogConnectRule],
        value: SourceGraphSnapshot
    )?
    /// Uncited subject pending delete confirm.
    var pendingDelete: PendingDelete?
    var isDeleting = false
    var deleteError: String?
    var pendingGridX: Int64 = 0
    var pendingGridY: Int64 = 0
    var draft = CreateDraft()
    var isSaving = false
    var labelError: String?
    var createError: String?
    var hoverGridX: Int64?
    var hoverGridY: Int64?
    /// Content-space point for the connect rubber-band (under cursor).
    var connectHoverPoint: CGPoint?
    var selectedSubjectID: String?
    /// Space/Return opened this card for inner controls (cleared on Esc / other select).
    var activatedSubjectID: String?
    /// Shown when drag / arrow persist fails after an optimistic patch.
    var toast: VocabularyToast?

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

    /// False when the Source has zero Artifacts — disables cite controls (S7-09).
    var canCite: Bool {
        guard let match = sourceRow else { return true }
        return match.hasArtifact
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
        CatalogQueryKey.sourceGraph(project: session.projectKey, sourceId: sourceID)
    }

    private var fieldsKey: CatalogQueryKey {
        CatalogQueryKey.subjectFieldsWorkspace(project: session.projectKey)
    }

    private var sourcesListKey: CatalogQueryKey {
        CatalogQueryKey.sourcesList(project: session.projectKey)
    }

    private var connectRulesKey: CatalogQueryKey {
        CatalogQueryKey.connectRules(project: session.projectKey)
    }

    private var fieldsSnapshot: SubjectFieldsSnapshot? {
        let handle: QueryHandle<SubjectFieldsSnapshot>? = session.queryHandle(fieldsKey)
        return handle?.value
    }

    private var sourceRow: CatalogSource? {
        let handle: QueryHandle<[CatalogSource]>? = session.queryHandle(sourcesListKey)
        return handle?.value?.first { $0.id == sourceID }
    }

    private var placeableTypes: [CatalogSubjectType] {
        (fieldsSnapshot?.types ?? []).filter {
            EvidencePrimaryKind(rawValue: $0.key) != nil
                || EvidenceBridgeKind(rawValue: $0.key) != nil
        }
    }

    init(
        sourceID: String,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String
    ) {
        self.sourceID = sourceID
        self.session = session
        self.store = store
        self.userID = userID
    }

    /// Snapshot as loaded. Edges come from cited observations.
    func displaySnapshot(rows: SourceGraphRows?, types: [CatalogSubjectType]) -> SourceGraphSnapshot {
        let rules = connectRules
        let normalized = rows ?? SourceGraphRows(sourceId: sourceID)
        if let snapshotMemo,
           snapshotMemo.rows == normalized,
           snapshotMemo.types == types,
           snapshotMemo.rules == rules
        {
            return snapshotMemo.value
        }
        let value = SourceGraphSnapshot.build(
            rows: normalized,
            types: types,
            rules: rules
        )
        snapshotMemo = (normalized, types, rules, value)
        return value
    }

    func prepare() async {
        let _: QueryHandle<SourceGraphRows> = session.query(graphKey)
        let rules: QueryHandle<[CatalogConnectRule]> = session.query(connectRulesKey)
        let _: QueryHandle<SubjectFieldsSnapshot> = session.query(fieldsKey)
        let _: QueryHandle<[CatalogSource]> = session.query(sourcesListKey)
        _ = await session.readyValue(graphKey) as SourceGraphRows?
        _ = await session.readyValue(fieldsKey) as SubjectFieldsSnapshot?
        _ = await session.readyValue(connectRulesKey) as [CatalogConnectRule]?
        _ = await session.readyValue(sourcesListKey) as [CatalogSource]?
        if rules.status == .error, let message = connectRulesError {
            toast = VocabularyToast(
                title: String(localized: L10n.EvidenceGraph.connectRulesUnavailableTitle),
                body: message,
                tone: .danger
            )
        }
        if !canCite {
            disarm()
        }
    }

    func refreshCanCite() async {
        let _: QueryHandle<[CatalogSource]> = session.query(sourcesListKey)
        _ = await session.readyValue(sourcesListKey) as [CatalogSource]?
        if !canCite {
            disarm()
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
            armedKind = kind
            hoverGridX = nil
            hoverGridY = nil
        }
    }

    func toggleConnect() {
        guard canCite else { return }
        if let connectRulesError {
            toast = VocabularyToast(
                title: String(localized: L10n.EvidenceGraph.connectRulesUnavailableTitle),
                body: connectRulesError,
                tone: .danger
            )
            return
        }
        if armedConnect {
            disarm()
        } else {
            armedKind = nil
            hoverGridX = nil
            hoverGridY = nil
            armedConnect = true
            connectOriginID = nil
            connectHoverPoint = nil
        }
    }

    func disarm() {
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

    func selectSubject(id: String?) {
        selectedSubjectID = id
        if activatedSubjectID != id {
            activatedSubjectID = nil
        }
    }

    func activateSubject(id: String) {
        selectedSubjectID = id
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
        var moveSubjectID: String?
        var moveFromX: Int64 = 0
        var moveFromY: Int64 = 0
        var moveDeltaX: Int64 = 0
        var moveDeltaY: Int64 = 0
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
            if activatedSubjectID != nil {
                deactivateSubject()
                return KeyEffect(handled: true)
            }
            if selectedSubjectID != nil || hasFocus {
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
            guard inputMode == .idle, let selectedID = selectedSubjectID,
                  let cell = gridCell(for: selectedID)
            else { return KeyEffect() }
            return KeyEffect(
                handled: true,
                moveSubjectID: selectedID,
                moveFromX: cell.x,
                moveFromY: cell.y,
                moveDeltaX: dx,
                moveDeltaY: dy
            )
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
            beginDelete(subjectID: subjectID)
            return nil
        case EvidenceSubjectCard.addPropertyActionID:
            return composerLocation(for: subjectID)
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

    func updateHover(contentPoint: CGPoint) {
        guard armedKind != nil, !isCreating else { return }
        let cell = GraphCanvasGridMapping.gridCell(contentPoint: contentPoint)
        hoverGridX = cell.gridX
        hoverGridY = cell.gridY
    }

    func updateConnectHover(contentPoint: CGPoint) {
        guard armedConnect, connectOriginID != nil, !isCreating else {
            connectHoverPoint = nil
            return
        }
        connectHoverPoint = contentPoint
    }

    func clearHover() {
        hoverGridX = nil
        hoverGridY = nil
        connectHoverPoint = nil
    }

    /// Opens the create dialog at the snapped cell (tool must be armed).
    func beginCreate(at contentPoint: CGPoint) {
        guard canCite, let kind = armedKind, !isCreating, !isSaving, editingSubjectID == nil else { return }
        let cell = GraphCanvasGridMapping.gridCell(contentPoint: contentPoint)
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
                title: String(localized: L10n.EvidenceGraph.connectInvalidPairTitle),
                body: String(localized: L10n.EvidenceGraph.connectInvalidPairBody),
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
                labelError = String(localized: L10n.EvidenceGraph.labelRequired)
                return nil
            }
            storedLabel = trimmed
        }

        isSaving = true
        defer { isSaving = false }
        labelError = nil
        createError = nil
        do {
            _ = try await store.updateSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectID: subjectID,
                label: storedLabel,
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            session.apply(.mutatedSourceGraph(sourceId: sourceID))
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
            title = EvidenceBridgeEdgeSummary.sentence(for: bridge, in: snapshot)
            ref = bridge.subject.ref
        } else {
            title = nil
            ref = nil
        }
        return WorkspaceLocation(
            section: .sources,
            sourceId: sourceID,
            subjectId: subjectID,
            citationId: citationID,
            artifactId: artifactID,
            observationId: observationID,
            sourceSurface: .citationComposer,
            ref: ref,
            title: title,
            sourceTitle: resolvedSourceTitle()
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

    /// Queues delete confirm for an uncited subject or bridge card.
    func beginDelete(subjectID: String) {
        let snapshot = currentSnapshot()
        deleteError = nil
        if let primary = primary(in: snapshot, id: subjectID), !primary.isCited {
            pendingDelete = PendingDelete(
                id: subjectID,
                label: primary.subject.label,
                ref: primary.subject.ref
            )
            return
        }
        if let snapshot, let bridge = snapshot.bridges.first(where: { $0.id == subjectID }),
           !bridge.isCited
        {
            pendingDelete = PendingDelete(
                id: subjectID,
                label: EvidenceBridgeEdgeSummary.sentence(for: bridge, in: snapshot),
                ref: bridge.subject.ref
            )
        }
    }

    @discardableResult
    func confirmDeleteSubject() async -> Bool {
        guard let pending = pendingDelete, !isDeleting else { return false }
        isDeleting = true
        deleteError = nil
        defer { isDeleting = false }
        do {
            try await store.deleteSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                subjectID: pending.id
            )
            session.apply(.mutatedSourceGraph(sourceId: sourceID))
            if selectedSubjectID == pending.id {
                selectSubject(id: nil)
            }
            if activatedSubjectID == pending.id {
                activatedSubjectID = nil
            }
            pendingDelete = nil
            return true
        } catch {
            deleteError = L10n.Errors.message(for: error)
            return false
        }
    }

    private func resolvedSourceTitle() -> String? {
        let listKey = CatalogQueryKey.sourcesList(project: session.projectKey)
        guard let handle: QueryHandle<[CatalogSource]> = session.queryHandle(listKey),
              let sources = handle.value,
              let match = sources.first(where: { $0.id == sourceID })
        else { return nil }
        return match.title
    }

    /// Source page recovery target when the graph has no Artifacts.
    func sourcePageLocation(title: String, ref: String?) -> WorkspaceLocation {
        WorkspaceLocation(
            section: .sources,
            sourceId: sourceID,
            sourceSurface: .page,
            ref: ref,
            title: title
        )
    }

    @discardableResult
    private func confirmPrimaryCreate() async -> String? {
        guard let kind = armedKind, !isSaving else { return nil }
        let trimmed = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty {
            labelError = String(localized: L10n.EvidenceGraph.labelRequired)
            return nil
        }
        guard let typeID = typeIDByKind[kind.rawValue] else {
            createError = String(localized: L10n.EvidenceGraph.typesUnavailable)
            return nil
        }

        isSaving = true
        defer { isSaving = false }
        labelError = nil
        createError = nil
        do {
            let created = try await store.createSubject(
                projectDir: session.projectKey.projectDir,
                userID: userID,
                sourceID: sourceID,
                subjectTypeID: typeID,
                label: trimmed,
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines),
                placement: CatalogGridCell(gridX: pendingGridX, gridY: pendingGridY)
            )
            session.apply(.mutatedSourceGraph(sourceId: sourceID))
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
        return WorkspaceLocation(
            section: .sources,
            sourceId: sourceID,
            subjectId: nil,
            connectFromSubjectId: fromID,
            connectToSubjectId: toID,
            connectBridgeTypeKey: rule.bridgeTypeKey,
            sourceSurface: .citationComposer,
            title: sentence,
            sourceTitle: resolvedSourceTitle()
        )
    }

    /// Snaps a completed card drag. Patches the session cache **synchronously**
    /// so `@GestureState` can clear onto the new cell, then writes the store.
    @discardableResult
    func commitDrag(
        subjectID: String,
        originGridX: Int64,
        originGridY: Int64,
        documentDelta: CGSize
    ) -> (gridX: Int64, gridY: Int64)? {
        guard inputMode == .idle else { return nil }
        let origin = GraphCanvasGridMapping.contentPoint(gridX: originGridX, gridY: originGridY)
        let dropped = CGPoint(
            x: origin.x + documentDelta.width,
            y: origin.y + documentDelta.height
        )
        let cell = GraphCanvasGridMapping.gridCell(contentPoint: dropped)
        let generation = bumpPositionGeneration(subjectID: subjectID)
        applyPositionPatch(subjectID: subjectID, gridX: cell.gridX, gridY: cell.gridY)
        Task {
            await persistPositionOrRevert(
                subjectID: subjectID,
                generation: generation,
                gridX: cell.gridX,
                gridY: cell.gridY
            )
        }
        return (cell.gridX, cell.gridY)
    }

    @discardableResult
    func moveSubject(
        subjectID: String,
        fromGridX: Int64,
        fromGridY: Int64,
        deltaX: Int64,
        deltaY: Int64
    ) async -> (gridX: Int64, gridY: Int64)? {
        guard inputMode == .idle else { return nil }
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

    func ghostPlacedSubject() -> SourceGraphPlacedSubject? {
        guard let kind = armedKind,
              let x = hoverGridX,
              let y = hoverGridY,
              !isCreating
        else { return nil }
        let label = typeLabelByKind[kind.rawValue] ?? kind.rawValue
        return SourceGraphPlacedSubject(
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
        }
    }

    func toolAccessibilityLabel(for kind: EvidencePrimaryKind, armed: Bool) -> String {
        let name = String(localized: toolName(for: kind))
        let state = armed
            ? String(localized: L10n.EvidenceGraph.toolOn)
            : String(localized: L10n.EvidenceGraph.toolOff)
        return "\(name), \(String(localized: L10n.EvidenceGraph.toolRole)), \(state)"
    }

    func connectToolAccessibilityLabel(armed: Bool) -> String {
        let name = String(localized: L10n.EvidenceGraph.toolConnect)
        let state = armed
            ? String(localized: L10n.EvidenceGraph.toolOn)
            : String(localized: L10n.EvidenceGraph.toolOff)
        return "\(name), \(String(localized: L10n.EvidenceGraph.toolRole)), \(state)"
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

    private func defaultLabel(for kind: EvidencePrimaryKind) -> String {
        switch kind {
        case .person: String(localized: L10n.EvidenceGraph.defaultLabelPerson)
        case .event: String(localized: L10n.EvidenceGraph.defaultLabelEvent)
        case .place: String(localized: L10n.EvidenceGraph.defaultLabelPlace)
        }
    }

    private func currentSnapshot() -> SourceGraphSnapshot? {
        guard let handle: QueryHandle<SourceGraphRows> = session.queryHandle(graphKey) else {
            return nil
        }
        return displaySnapshot(rows: handle.value, types: fieldsSnapshot?.types ?? [])
    }

    private func primary(in snapshot: SourceGraphSnapshot?, id: String) -> SourceGraphPlacedSubject? {
        snapshot?.subjects.first { $0.id == id }
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
        if let handle: QueryHandle<SourceGraphRows> = session.queryHandle(graphKey),
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
                    title: String(localized: L10n.EvidenceGraph.positionPersistFailedTitle),
                    body: L10n.Errors.message(for: error),
                    tone: .danger
                )
            }
            return false
        }
    }
}
