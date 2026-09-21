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

    let sourceID: String
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String
    let linkStore: any EvidenceProvisionalLinkStoring

    var armedKind: EvidencePrimaryKind?
    /// Connect tool armed (mutually exclusive with `armedKind`).
    var armedConnect = false
    /// First primary chosen while connecting.
    var connectOriginID: String?
    /// Inferred bridge kind while the create sheet is open for a connect.
    var pendingBridgeKind: EvidenceBridgeKind?
    /// Endpoint B id while creating a bridge (A is `connectOriginID`).
    var pendingEndpointBID: String?

    var isCreating = false
    /// Edit label/description sheet open for this subject (mutually exclusive with create).
    var editingSubjectID: String?
    /// Kind used for edit dialog title when editing a primary; nil when editing a bridge.
    var editingPrimaryKind: EvidencePrimaryKind?
    var editingBridgeKind: EvidenceBridgeKind?
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
    private(set) var typeIDByKind: [String: String] = [:]
    private(set) var typeLabelByKind: [String: String] = [:]
    /// Registry presentation tokens keyed by type key (S7-09).
    private(set) var presentationByKind: [String: CatalogSubjectTypePresentation] = [:]
    /// False when the Source has zero Artifacts — disables cite controls (S7-09).
    private(set) var canCite = true

    var isSheetPresented: Bool {
        isCreating || editingSubjectID != nil
    }

    private var graphKey: CatalogQueryKey {
        CatalogQueryKey.sourceGraph(project: session.projectKey, sourceId: sourceID)
    }

    init(
        sourceID: String,
        session: WorkspaceSession,
        store: any GenealogyStore,
        userID: String,
        linkStore: (any EvidenceProvisionalLinkStoring)? = nil
    ) {
        self.sourceID = sourceID
        self.session = session
        self.store = store
        self.userID = userID
        if let linkStore {
            self.linkStore = linkStore
        } else if let live = try? EvidenceProvisionalLinkStore(
            projectDir: session.projectKey.projectDir
        ) {
            self.linkStore = live
        } else {
            self.linkStore = InMemoryEvidenceProvisionalLinkStore()
        }
    }

    /// Snapshot with provisional endpoints attached for drawing.
    func displaySnapshot(from raw: SourceGraphSnapshot?) -> SourceGraphSnapshot {
        let base = raw ?? SourceGraphSnapshot(sourceId: sourceID)
        return base.attaching(links: linkStore.links(for: sourceID))
    }

    func prepare() async {
        do {
            let types = try await store.listSubjectTypes(projectDir: session.projectKey.projectDir)
            var ids: [String: String] = [:]
            var labels: [String: String] = [:]
            for type in types {
                let isPrimary = EvidencePrimaryKind(rawValue: type.key) != nil
                let isBridge = EvidenceBridgeKind(rawValue: type.key) != nil
                guard isPrimary || isBridge else { continue }
                ids[type.key] = type.id
                labels[type.key] = type.label
            }
            typeIDByKind = ids
            typeLabelByKind = labels
        } catch {
            typeIDByKind = [:]
            typeLabelByKind = [:]
        }

        var presentations: [String: CatalogSubjectTypePresentation] = [:]
        if let placeable = try? await store.listPlaceableSubjectTypes() {
            for item in placeable {
                presentations[item.typeKey] = item
            }
        }
        let keys = EvidencePrimaryKind.allCases.map(\.rawValue)
            + EvidenceBridgeKind.allCases.map(\.rawValue)
        for key in keys where presentations[key] == nil {
            if let presentation = try? await store.getSubjectTypePresentation(typeKey: key) {
                presentations[key] = presentation
            }
        }
        presentationByKind = presentations

        await refreshCanCite()
    }

    func refreshCanCite() async {
        let listKey = CatalogQueryKey.sourcesList(project: session.projectKey)
        let sources: [CatalogSource]?
        if let handle: QueryHandle<[CatalogSource]> = session.queryHandle(listKey) {
            sources = handle.value
        } else {
            let handle: QueryHandle<[CatalogSource]> = session.query(listKey)
            var waited = 0
            while handle.value == nil && waited < 40 {
                try? await Task.sleep(nanoseconds: 25_000_000)
                waited += 1
            }
            sources = handle.value
        }
        guard let match = sources?.first(where: { $0.id == sourceID }) else {
            // Keep previous / default until the Sources list includes this id.
            return
        }
        canCite = match.hasArtifact
        if !match.hasArtifact {
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
        if armedConnect {
            disarm()
        } else {
            armedKind = nil
            hoverGridX = nil
            hoverGridY = nil
            armedConnect = true
            connectOriginID = nil
            connectHoverPoint = nil
            pendingBridgeKind = nil
            pendingEndpointBID = nil
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
        pendingBridgeKind = nil
        pendingEndpointBID = nil
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

    /// Pointer / keyboard pick while Connect is armed.
    func handleConnectPick(subjectID: String, kind: EvidencePrimaryKind, label: String) {
        guard canCite, armedConnect, !isCreating, !isSaving, editingSubjectID == nil else { return }
        if connectOriginID == nil {
            connectOriginID = subjectID
            selectSubject(id: subjectID)
            return
        }
        guard let originID = connectOriginID, originID != subjectID else { return }
        beginBridgeCreate(
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
        pendingBridgeKind = nil
        pendingEndpointBID = nil
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

    private func beginBridgeCreate(
        originID: String,
        targetID: String,
        targetKind: EvidencePrimaryKind,
        targetLabel: String
    ) {
        guard canCite else { return }
        guard let origin = primary(in: currentSnapshot(), id: originID) else { return }
        guard let bridgeKind = EvidenceBridgeKindInference.kind(origin.kind, targetKind) else {
            toast = VocabularyToast(
                title: String(localized: L10n.EvidenceGraph.connectInvalidPairTitle),
                body: String(localized: L10n.EvidenceGraph.connectInvalidPairBody),
                tone: .danger
            )
            return
        }

        let mid = midpointCell(origin: origin, targetID: targetID)
        pendingGridX = mid.gridX
        pendingGridY = mid.gridY
        pendingBridgeKind = bridgeKind
        pendingEndpointBID = targetID
        draft = CreateDraft(
            label: defaultBridgeLabel(for: bridgeKind),
            description: ""
        )
        labelError = nil
        createError = nil
        isCreating = true
        connectHoverPoint = nil
        _ = targetLabel
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
        let wasBridge = pendingBridgeKind != nil
        isCreating = false
        labelError = nil
        createError = nil
        pendingBridgeKind = nil
        pendingEndpointBID = nil
        if wasBridge {
            // Keep Connect armed with A held (S6-D2).
            return
        }
        disarm()
    }

    /// Creates a primary subject + position, or a bridge + provisional link.
    @discardableResult
    func confirmCreate() async -> String? {
        if pendingBridgeKind != nil {
            return await confirmBridgeCreate()
        }
        return await confirmPrimaryCreate()
    }

    @discardableResult
    func confirmEdit() async -> String? {
        guard let subjectID = editingSubjectID, !isSaving else { return nil }
        let trimmed = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty {
            labelError = String(localized: L10n.EvidenceGraph.labelRequired)
            return nil
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
                label: trimmed,
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            session.apply(.createdSubject(sourceId: sourceID))
            let _: QueryHandle<SourceGraphSnapshot> = session.query(graphKey)
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

    /// Citation composer place for Add property (stub destination until S7-08).
    func composerLocation(for subjectID: String) -> WorkspaceLocation? {
        guard canCite else { return nil }
        let snapshot = currentSnapshot()
        let title: String?
        let ref: String?
        if let primary = primary(in: snapshot, id: subjectID) {
            title = primary.subject.label
            ref = primary.subject.ref
        } else if let bridge = snapshot?.bridges.first(where: { $0.id == subjectID }) {
            title = bridge.subject.label
            ref = bridge.subject.ref
        } else {
            title = nil
            ref = nil
        }
        return WorkspaceLocation(
            section: .sources,
            sourceId: sourceID,
            subjectId: subjectID,
            sourceSurface: .citationComposer,
            ref: ref,
            title: title
        )
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
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            _ = try await store.setSubjectPosition(
                projectDir: session.projectKey.projectDir,
                subjectID: created.id,
                gridX: pendingGridX,
                gridY: pendingGridY
            )
            session.apply(.createdSubject(sourceId: sourceID))
            let _: QueryHandle<SourceGraphSnapshot> = session.query(graphKey)
            isCreating = false
            disarm()
            return created.id
        } catch {
            createError = L10n.Errors.message(for: error)
            return nil
        }
    }

    @discardableResult
    private func confirmBridgeCreate() async -> String? {
        guard let bridgeKind = pendingBridgeKind,
              let originID = connectOriginID,
              let endpointBID = pendingEndpointBID,
              !isSaving
        else { return nil }

        let trimmed = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty {
            labelError = String(localized: L10n.EvidenceGraph.labelRequired)
            return nil
        }
        guard let typeID = typeIDByKind[bridgeKind.rawValue] else {
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
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            _ = try await store.setSubjectPosition(
                projectDir: session.projectKey.projectDir,
                subjectID: created.id,
                gridX: pendingGridX,
                gridY: pendingGridY
            )
            linkStore.upsert(
                EvidenceProvisionalLink(
                    bridgeSubjectID: created.id,
                    endpointAID: originID,
                    endpointBID: endpointBID
                ),
                sourceID: sourceID
            )
            session.apply(.createdSubject(sourceId: sourceID))
            let _: QueryHandle<SourceGraphSnapshot> = session.query(graphKey)
            isCreating = false
            pendingBridgeKind = nil
            pendingEndpointBID = nil
            connectOriginID = nil
            connectHoverPoint = nil
            armedConnect = false
            return created.id
        } catch {
            createError = L10n.Errors.message(for: error)
            return nil
        }
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
        applyPositionPatch(subjectID: subjectID, gridX: cell.gridX, gridY: cell.gridY)
        Task {
            await persistPositionOrRevert(
                subjectID: subjectID,
                gridX: cell.gridX,
                gridY: cell.gridY,
                revertGridX: originGridX,
                revertGridY: originGridY
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
        applyPositionPatch(subjectID: subjectID, gridX: nextX, gridY: nextY)
        let ok = await persistPositionOrRevert(
            subjectID: subjectID,
            gridX: nextX,
            gridY: nextY,
            revertGridX: fromGridX,
            revertGridY: fromGridY
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
        if let bridgeKind = pendingBridgeKind {
            return bridgeCreateTitle(for: bridgeKind)
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

    /// Mono endpoint line under the bridge create title.
    func bridgeEndpointLine() -> String? {
        guard pendingBridgeKind != nil,
              let originID = connectOriginID,
              let endpointBID = pendingEndpointBID,
              let snapshot = currentSnapshot()
        else { return nil }
        let a = displayLabel(forPrimaryID: originID, in: snapshot)
        let b = displayLabel(forPrimaryID: endpointBID, in: snapshot)
        return "\(a) → \(b)"
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

    private func defaultBridgeLabel(for kind: EvidenceBridgeKind) -> String {
        switch kind {
        case .relationship: String(localized: L10n.EvidenceGraph.defaultLabelRelationship)
        case .participation: String(localized: L10n.EvidenceGraph.defaultLabelParticipation)
        case .location: String(localized: L10n.EvidenceGraph.defaultLabelLocation)
        }
    }

    private func currentSnapshot() -> SourceGraphSnapshot? {
        guard let handle: QueryHandle<SourceGraphSnapshot> = session.queryHandle(graphKey) else {
            return nil
        }
        return displaySnapshot(from: handle.value)
    }

    private func primary(in snapshot: SourceGraphSnapshot?, id: String) -> SourceGraphPlacedSubject? {
        snapshot?.subjects.first { $0.id == id }
    }

    private func displayLabel(forPrimaryID id: String, in snapshot: SourceGraphSnapshot) -> String {
        guard let placed = snapshot.subjects.first(where: { $0.id == id }) else { return id }
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? placed.typeLabel : trimmed
    }

    private func midpointCell(
        origin: SourceGraphPlacedSubject,
        targetID: String
    ) -> (gridX: Int64, gridY: Int64) {
        guard let target = primary(in: currentSnapshot(), id: targetID) else {
            return (origin.gridX, origin.gridY)
        }
        let pa = GraphCanvasGridMapping.contentPoint(gridX: origin.gridX, gridY: origin.gridY)
        let pb = GraphCanvasGridMapping.contentPoint(gridX: target.gridX, gridY: target.gridY)
        let mid = CGPoint(x: (pa.x + pb.x) / 2, y: (pa.y + pb.y) / 2)
        return GraphCanvasGridMapping.gridCell(contentPoint: mid)
    }

    private func applyPositionPatch(subjectID: String, gridX: Int64, gridY: Int64) {
        if let handle: QueryHandle<SourceGraphSnapshot> = session.queryHandle(graphKey),
           let snapshot = handle.value
        {
            session.setQueryValue(
                graphKey,
                value: snapshot.updatingPosition(subjectID: subjectID, gridX: gridX, gridY: gridY)
            )
        }
    }

    @discardableResult
    private func persistPositionOrRevert(
        subjectID: String,
        gridX: Int64,
        gridY: Int64,
        revertGridX: Int64,
        revertGridY: Int64
    ) async -> Bool {
        do {
            _ = try await store.setSubjectPosition(
                projectDir: session.projectKey.projectDir,
                subjectID: subjectID,
                gridX: gridX,
                gridY: gridY
            )
            return true
        } catch {
            applyPositionPatch(subjectID: subjectID, gridX: revertGridX, gridY: revertGridY)
            toast = VocabularyToast(
                title: String(localized: L10n.EvidenceGraph.positionPersistFailedTitle),
                body: L10n.Errors.message(for: error),
                tone: .danger
            )
            return false
        }
    }
}
