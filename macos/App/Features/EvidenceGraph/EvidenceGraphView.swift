import AppKit
import SwiftUI

/// Workspace destination for a Source’s Evidence graph.
///
/// Product place under Sources (`sourceSurface: .graph`). Composes the
/// reusable [`GraphCanvas`](../GraphCanvas/) shell with primary + bridge cards,
/// palette place/create/connect, and drag persist (S6-02–S6-04).
struct EvidenceGraphView: View {
    /// Large enough to pan; Source-scoped graphs stay small (design note §7.1).
    private static let contentSize = CGSize(width: 4_000, height: 4_000)

    let sourceID: String
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String

    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: EvidenceGraphModel
    @State private var sourceTitle: String = ""
    @State private var sourceRef: String?

    private var graphKey: CatalogQueryKey {
        CatalogQueryKey.sourceGraph(project: session.projectKey, sourceId: sourceID)
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
        _model = State(
            initialValue: EvidenceGraphModel(
                sourceID: sourceID,
                session: session,
                store: store,
                userID: userID
            )
        )
    }

    var body: some View {
        Group {
            if let handle: QueryHandle<SourceGraphSnapshot> = session.queryHandle(graphKey) {
                EvidenceGraphContent(
                    handle: handle,
                    model: model,
                    sourceID: sourceID,
                    sourceTitle: sourceTitle,
                    sourceRef: sourceRef,
                    contentSize: Self.contentSize,
                    navigation: navigation
                )
            } else {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .background(PVColor.surfacePage)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .accessibilityIdentifier("workspace.destination.evidenceGraph")
        .task(id: sourceID) {
            let _: QueryHandle<SourceGraphSnapshot> = session.query(graphKey)
            await model.prepare()
            await refreshSourceChrome()
        }
        .onChange(of: model.armedKind) { _, kind in
            if kind != nil {
                NSCursor.crosshair.push()
            } else if !model.armedConnect {
                NSCursor.pop()
            }
        }
        .onChange(of: model.armedConnect) { _, armed in
            if armed {
                NSCursor.crosshair.push()
            } else if model.armedKind == nil {
                NSCursor.pop()
            }
        }
        .onDisappear {
            if model.armedKind != nil || model.armedConnect {
                NSCursor.pop()
            }
        }
    }

    private func refreshSourceChrome() async {
        let listKey = CatalogQueryKey.sourcesList(project: session.projectKey)
        if let handle: QueryHandle<[CatalogSource]> = session.queryHandle(listKey),
           let sources = handle.value,
           let match = sources.first(where: { $0.id == sourceID })
        {
            sourceTitle = match.title
            sourceRef = match.ref
            return
        }
        let handle: QueryHandle<[CatalogSource]> = session.query(listKey)
        var waited = 0
        while handle.value == nil && waited < 40 {
            try? await Task.sleep(nanoseconds: 25_000_000)
            waited += 1
        }
        if let match = handle.value?.first(where: { $0.id == sourceID }) {
            sourceTitle = match.title
            sourceRef = match.ref
        }
    }
}
// MARK: - Content

private struct EvidenceGraphContent: View {
    @Bindable var handle: QueryHandle<SourceGraphSnapshot>
    @Bindable var model: EvidenceGraphModel
    let sourceID: String
    let sourceTitle: String
    let sourceRef: String?
    let contentSize: CGSize
    let navigation: WorkspaceNavigation
    @FocusState private var focus: EvidenceGraphFocus?

    private var snapshot: SourceGraphSnapshot {
        model.displaySnapshot(from: handle.value)
    }

    private var subjects: [SourceGraphPlacedSubject] {
        snapshot.subjects
    }

    private var bridges: [SourceGraphPlacedBridge] {
        snapshot.bridges
    }

    private var sheetPresented: Binding<Bool> {
        Binding(
            get: { model.isSheetPresented },
            set: { newValue in
                if !newValue { model.cancelSheet() }
            }
        )
    }

    var body: some View {
        VStack(spacing: 0) {
            header
            canvasPane
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .background(PVColor.surfacePage)
        .onKeyPress(phases: .down) { press in
            handleKey(press)
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(accessibilityGraphLabel)
        .accessibilityRotor(String(localized: L10n.EvidenceGraph.subjectsRotor)) {
            ForEach(subjects) { placed in
                AccessibilityRotorEntry(
                    EvidenceSubjectCard.accessibilityLabel(for: placed),
                    id: placed.id
                ) {
                    model.selectSubject(id: placed.id)
                }
            }
        }
        .accessibilityRotor(String(localized: L10n.EvidenceGraph.linksRotor)) {
            ForEach(bridges) { placed in
                AccessibilityRotorEntry(
                    EvidenceBridgeCard.accessibilityLabel(for: placed),
                    id: placed.id
                ) {
                    model.selectSubject(id: placed.id)
                }
            }
        }
        .vocabularyToastOverlay($model.toast, identifier: "evidenceGraph.toast")
        .pvFormDialog(
            isPresented: sheetPresented,
            copy: PVFormDialogCopy(
                title: model.createDialogTitle(),
                confirm: model.sheetConfirmLabel(),
                cancel: L10n.EvidenceGraph.createCancel
            ),
            isRunning: model.isSaving,
            confirmDisabled: model.draft.label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
            accessibilityIdentifierPrefix: model.editingSubjectID != nil
                ? "evidenceGraph.edit"
                : "evidenceGraph.create",
            onConfirm: {
                Task {
                    if model.editingSubjectID != nil {
                        _ = await model.confirmEdit()
                    } else if let id = await model.confirmCreate() {
                        model.selectSubject(id: id)
                    }
                }
            }
        ) {
            createForm
        }
    }
    private var accessibilityGraphLabel: Text {
        if model.armedConnect {
            Text(model.connectArmedHint)
        } else if let kind = model.armedKind {
            Text(model.armedHint(for: kind))
        } else {
            Text(L10n.Workspace.evidenceGraphTitle)
        }
    }

    private var header: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space5) {
            Text(L10n.Workspace.evidenceGraphTitle)
                .font(PVFont.display(size: PVTypeScale.h1, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
            if !sourceTitle.isEmpty {
                Text(verbatim: sourceTitle)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
            }
            Spacer(minLength: 0)
            Text(L10n.EvidenceGraph.subjectCount(count: subjects.count + bridges.count))
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textFaint)
        }
        .padding(.horizontal, PVSpacing.space7)
        .padding(.vertical, PVSpacing.space5)
        .accessibilityElement(children: .combine)
    }

    private var canvasPane: some View {
        ZStack(alignment: .top) {
            GraphCanvasScrollView(contentSize: contentSize, contentID: sourceID) {
                EvidenceGraphDocument(
                    handle: handle,
                    model: model,
                    contentSize: contentSize,
                    navigation: navigation
                )
            }

            HStack(alignment: .top, spacing: PVSpacing.space5) {
                EvidenceGraphPalette(model: model, focus: $focus)
                Spacer(minLength: 0)
                if !model.canCite {
                    noArtifactCallout
                        .frame(width: 360, alignment: .trailing)
                } else if case .placing(let kind) = model.inputMode {
                    armedBanner(Text(model.armedHint(for: kind)))
                } else if case .connecting = model.inputMode {
                    armedBanner(Text(model.connectArmedHint))
                }
            }
            .padding(.horizontal, PVSpacing.space5)
            .padding(.top, PVSpacing.space5)
        }
    }

    private var noArtifactCallout: some View {
        PVCallout(
            tone: .warning,
            title: L10n.EvidenceGraph.noArtifactTitle,
            message: String(localized: L10n.EvidenceGraph.noArtifactMessage)
        ) {
            PVButton(L10n.EvidenceGraph.noArtifactAction, variant: .secondary, size: .sm) {
                navigation.go(
                    to: model.sourcePageLocation(title: sourceTitle, ref: sourceRef)
                )
            }
        }
        .accessibilityIdentifier("evidenceGraph.noArtifact.callout")
    }
    private func armedBanner(_ hint: Text) -> some View {
        HStack(spacing: 12) {
            hint
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textPrimary)
            Button {
                model.disarm()
                focus = nil
            } label: {
                Text(L10n.EvidenceGraph.createCancel)
                    .font(PVFont.mono(size: 11, weight: PVFontWeight.medium))
            }
            .buttonStyle(.plain)
            .foregroundStyle(PVColor.textSecondary)
            .accessibilityIdentifier("evidenceGraph.armed.cancel")
            .keyboardShortcut(.cancelAction)
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 7)
        .background(PVColor.surfaceCard)
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .shadow(color: Color.black.opacity(0.06), radius: 6, y: 1)
    }

    @ViewBuilder
    private var createForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if model.editingSubjectID == nil, let bridgeKind = model.pendingBridgeKind {
                HStack(spacing: 8) {
                    ZStack {
                        RoundedRectangle(cornerRadius: 4, style: .continuous)
                            .fill(PVColor.surfacePage)
                        PVMark(bridgeKind.markKey, size: 15)
                            .foregroundStyle(PVColor.textMuted)
                    }
                    .frame(width: 28, height: 28)
                    VStack(alignment: .leading, spacing: 2) {
                        Text(verbatim: model.typeLabelByKind[bridgeKind.rawValue] ?? bridgeKind.rawValue)
                            .font(PVFont.mono(size: 10, weight: PVFontWeight.medium))
                            .tracking(1)
                            .textCase(.uppercase)
                            .foregroundStyle(PVColor.textMuted)
                        if let line = model.bridgeEndpointLine() {
                            Text(verbatim: line)
                                .font(PVFont.mono(size: 11))
                                .foregroundStyle(PVColor.textFaint)
                        }
                    }
                }
            } else if model.editingSubjectID == nil, let kind = model.armedKind {
                let style = EvidenceSubjectKindStyle.resolve(
                    typeKey: kind.rawValue,
                    presentation: model.presentation(for: kind.rawValue)
                )
                HStack(spacing: 8) {
                    ZStack {
                        RoundedRectangle(cornerRadius: 4, style: .continuous)
                            .fill(style.chip)
                        PVMark(kind.markKey, size: 15)
                            .foregroundStyle(style.ink)
                    }
                    .frame(width: 28, height: 28)
                    Text(verbatim: styleTypeLabel(kind))
                        .font(PVFont.mono(size: 10, weight: PVFontWeight.medium))
                        .tracking(1)
                        .textCase(.uppercase)
                        .foregroundStyle(style.ink)
                }
            } else if let kind = model.editingPrimaryKind {
                let style = EvidenceSubjectKindStyle.resolve(
                    typeKey: kind.rawValue,
                    presentation: model.presentation(for: kind.rawValue)
                )
                HStack(spacing: 8) {
                    ZStack {
                        RoundedRectangle(cornerRadius: 4, style: .continuous)
                            .fill(style.chip)
                        PVMark(kind.markKey, size: 15)
                            .foregroundStyle(style.ink)
                    }
                    .frame(width: 28, height: 28)
                    Text(verbatim: styleTypeLabel(kind))
                        .font(PVFont.mono(size: 10, weight: PVFontWeight.medium))
                        .tracking(1)
                        .textCase(.uppercase)
                        .foregroundStyle(style.ink)
                }
            }
            if let createError = model.createError {
                Text(verbatim: createError)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.danger)
            }
            PVField(
                label: L10n.EvidenceGraph.labelField,
                error: model.labelError,
                required: true
            ) {
                PVInput(
                    text: Binding(
                        get: { model.draft.label },
                        set: { model.draft.label = $0 }
                    ),
                    isInvalid: model.labelError != nil
                )
            }
            PVField(label: L10n.EvidenceGraph.descriptionField) {
                PVTextArea(
                    text: Binding(
                        get: { model.draft.description },
                        set: { model.draft.description = $0 }
                    )
                )
            }
        }
    }
    private func styleTypeLabel(_ kind: EvidencePrimaryKind) -> String {
        model.typeLabelByKind[kind.rawValue]
            ?? String(localized: model.toolName(for: kind))
    }

    private func handleKey(_ press: KeyPress) -> KeyPress.Result {
        if model.isSheetPresented { return .ignored }

        if press.key == .escape {
            if model.armedConnect || model.armedKind != nil {
                model.disarm()
                return .handled
            }
            if model.activatedSubjectID != nil {
                model.deactivateSubject()
                return .handled
            }
            if model.selectedSubjectID != nil || focus != nil {
                model.selectSubject(id: nil)
                focus = nil
                return .handled
            }
            return .ignored
        }

        if press.key == .return || press.key == .space {
            guard model.inputMode == .idle,
                  let selectedID = model.selectedSubjectID
            else { return .ignored }
            if press.key == .return {
                model.beginEdit(subjectID: selectedID)
            } else {
                model.activateSubject(id: selectedID)
            }
            return .handled
        }

        guard model.inputMode == .idle,
              let selectedID = model.selectedSubjectID
        else { return .ignored }

        let grid: (Int64, Int64)?
        if let placed = subjects.first(where: { $0.id == selectedID }) {
            grid = (placed.gridX, placed.gridY)
        } else if let bridge = bridges.first(where: { $0.id == selectedID }) {
            grid = (bridge.gridX, bridge.gridY)
        } else {
            grid = nil
        }
        guard let (fromX, fromY) = grid else { return .ignored }

        let delta: (Int64, Int64)?
        switch press.key {
        case .leftArrow: delta = (-1, 0)
        case .rightArrow: delta = (1, 0)
        case .upArrow: delta = (0, -1)
        case .downArrow: delta = (0, 1)
        default: delta = nil
        }
        guard let (dx, dy) = delta else { return .ignored }

        Task {
            _ = await model.moveSubject(
                subjectID: selectedID,
                fromGridX: fromX,
                fromGridY: fromY,
                deltaX: dx,
                deltaY: dy
            )
        }
        return .handled
    }
}

// MARK: - Document (paint-only; AppKit owns pointer)

private struct EvidenceGraphDocument: View {
    @Bindable var handle: QueryHandle<SourceGraphSnapshot>
    @Bindable var model: EvidenceGraphModel
    let contentSize: CGSize
    let navigation: WorkspaceNavigation
    @Environment(\.graphCanvasPointer) private var pointer

    var body: some View {
        Group {
            if let pointer {
                EvidenceGraphDocumentBody(
                    handle: handle,
                    model: model,
                    contentSize: contentSize,
                    navigation: navigation,
                    pointer: pointer
                )
            } else {
                Color.clear
                    .frame(width: contentSize.width, height: contentSize.height)
            }
        }
        .frame(width: contentSize.width, height: contentSize.height)
    }
}

/// Observes ``GraphCanvasPointerController`` so live drag offsets refresh paint.
private struct EvidenceGraphDocumentBody: View {
    @Bindable var handle: QueryHandle<SourceGraphSnapshot>
    @Bindable var model: EvidenceGraphModel
    let contentSize: CGSize
    let navigation: WorkspaceNavigation
    @Bindable var pointer: GraphCanvasPointerController

    private var snapshot: SourceGraphSnapshot {
        model.displaySnapshot(from: handle.value)
    }

    private var subjects: [SourceGraphPlacedSubject] {
        snapshot.subjects
    }

    private var bridges: [SourceGraphPlacedBridge] {
        snapshot.bridges
    }

    private var inputMode: EvidenceCanvasInputMode {
        model.inputMode
    }

    var body: some View {
        ZStack(alignment: .topLeading) {
            GraphCanvasGridView(contentSize: contentSize)
                .accessibilityHidden(true)

            EvidenceGraphEdgesHost(
                pointer: pointer,
                snapshot: snapshot,
                selectedBridgeID: bridges.contains(where: { $0.id == model.selectedSubjectID })
                    ? model.selectedSubjectID
                    : nil
            )
            .frame(width: contentSize.width, height: contentSize.height)

            if case .connecting = inputMode,
               let originID = model.connectOriginID,
               let origin = subjects.first(where: { $0.id == originID }),
               let cursor = model.connectHoverPoint
            {
                EvidenceGraphConnectRubberBand(origin: origin, cursor: cursor)
                    .frame(width: contentSize.width, height: contentSize.height)
            }

            ForEach(bridges) { placed in
                bridgeCardView(for: placed)
            }

            ForEach(subjects) { placed in
                cardView(for: placed)
            }

            if let ghost = model.ghostPlacedSubject() {
                EvidenceSubjectCard.ghost(
                    placed: ghost,
                    presentation: model.presentation(for: ghost.kind.rawValue)
                )
                    .position(EvidenceSubjectCard.contentCenter(gridX: ghost.gridX, gridY: ghost.gridY))
            }

            if subjects.isEmpty, bridges.isEmpty, handle.status == .ready, inputMode == .idle {
                emptyOverlay
            }
        }
        .frame(width: contentSize.width, height: contentSize.height)
        .coordinateSpace(name: EvidenceSubjectCard.documentCoordinateSpace)
        .onAppear { wirePointer() }
        .onChange(of: inputMode) { _, _ in
            syncPointerMode()
        }
        .onChange(of: model.canCite) { _, _ in
            publishHitTargets()
        }
        .onChange(of: subjects.map(\.id)) { _, _ in
            publishHitTargets()
        }
        .onChange(of: bridges.map(\.id)) { _, _ in
            publishHitTargets()
        }
        .onChange(of: pointer.offsets.count) { _, _ in
            publishHitTargets()
        }
        .onChange(of: subjects.map { "\($0.id):\($0.gridX),\($0.gridY):\($0.observations.count)" }) { _, _ in
            publishHitTargets()
        }
        .onChange(of: bridges.map { "\($0.id):\($0.gridX),\($0.gridY)" }) { _, _ in
            publishHitTargets()
        }
    }

    private func wirePointer() {
        syncPointerMode()
        publishHitTargets()
        let model = model
        let handle = handle
        let navigation = navigation
        pointer.onSelect = { id in
            model.selectSubject(id: id)
        }
        pointer.onDeselect = {
            model.selectSubject(id: nil)
        }
        pointer.onDragEnded = { id, delta in
            let snap = model.displaySnapshot(from: handle.value)
            let grid: (Int64, Int64)?
            if let placed = snap.subjects.first(where: { $0.id == id }) {
                grid = (placed.gridX, placed.gridY)
            } else if let bridge = snap.bridges.first(where: { $0.id == id }) {
                grid = (bridge.gridX, bridge.gridY)
            } else {
                grid = nil
            }
            guard let (ox, oy) = grid else { return }
            _ = model.commitDrag(
                subjectID: id,
                originGridX: ox,
                originGridY: oy,
                documentDelta: delta
            )
        }
        pointer.onPlace = { point in
            model.beginCreate(at: point)
        }
        pointer.onConnectPick = { id in
            let snap = model.displaySnapshot(from: handle.value)
            guard let placed = snap.subjects.first(where: { $0.id == id }) else { return }
            model.handleConnectPick(
                subjectID: placed.id,
                kind: placed.kind,
                label: placed.subject.label
            )
        }
        pointer.onCardAction = { id, actionID in
            switch actionID {
            case EvidenceSubjectCard.editActionID, EvidenceBridgeCard.editActionID:
                model.beginEdit(subjectID: id)
            case EvidenceSubjectCard.addPropertyActionID:
                if let location = model.composerLocation(for: id) {
                    navigation.go(to: location)
                }
            default:
                break
            }
        }
        pointer.onHover = { point in
            switch model.inputMode {
            case .placing:
                if let point {
                    model.updateHover(contentPoint: point)
                } else {
                    model.clearHover()
                }
            case .connecting:
                if let point {
                    model.updateConnectHover(contentPoint: point)
                } else {
                    model.clearHover()
                }
            case .idle:
                break
            }
        }
    }

    private func syncPointerMode() {
        switch inputMode {
        case .idle:
            pointer.mode = .idle
        case .placing:
            pointer.mode = .placing
            model.selectSubject(id: nil)
        case .connecting:
            pointer.mode = .connecting
        }
    }

    private func publishHitTargets() {
        var targets: [GraphCanvasHitTarget] = []
        for placed in bridges {
            let offset = pointer.offsets[placed.id] ?? .zero
            targets.append(
                GraphCanvasHitTarget(
                    id: placed.id,
                    frame: EvidenceBridgeCard.contentFrame(
                        gridX: placed.gridX,
                        gridY: placed.gridY,
                        dragOffset: offset
                    ),
                    acceptsConnect: false,
                    actions: EvidenceBridgeCard.actionTargets(for: placed, dragOffset: offset)
                )
            )
        }
        for placed in subjects {
            let offset = pointer.offsets[placed.id] ?? .zero
            targets.append(
                GraphCanvasHitTarget(
                    id: placed.id,
                    frame: EvidenceSubjectCard.edgeFrame(for: placed, dragOffset: offset),
                    acceptsConnect: true,
                    actions: EvidenceSubjectCard.actionTargets(
                        for: placed,
                        canCite: model.canCite,
                        dragOffset: offset
                    )
                )
            )
        }
        pointer.hitTargets = targets
    }

    @ViewBuilder
    private func cardView(for placed: SourceGraphPlacedSubject) -> some View {
        let layout = EvidenceSubjectCard.topLeadingOffset(gridX: placed.gridX, gridY: placed.gridY)
        let drag = pointer.offsets[placed.id] ?? .zero
        EvidenceSubjectCard(
            placed: placed,
            presentation: model.presentation(for: placed.kind.rawValue),
            isSelected: model.selectedSubjectID == placed.id,
            isActivated: model.activatedSubjectID == placed.id,
            isConnectingFrom: model.connectOriginID == placed.id,
            canCite: model.canCite,
            dragOffset: drag
        )
        .accessibilityAction(named: Text(L10n.EvidenceGraph.editAccessibility)) {
            model.beginEdit(subjectID: placed.id)
        }
        .accessibilityAction(named: Text(L10n.EvidenceGraph.addProperty)) {
            if let location = model.composerLocation(for: placed.id) {
                navigation.go(to: location)
            }
        }
        .offset(x: layout.width, y: layout.height)
    }

    @ViewBuilder
    private func bridgeCardView(for placed: SourceGraphPlacedBridge) -> some View {
        let layout = EvidenceBridgeCard.topLeadingOffset(gridX: placed.gridX, gridY: placed.gridY)
        let drag = pointer.offsets[placed.id] ?? .zero
        EvidenceBridgeCard(
            placed: placed,
            isSelected: model.selectedSubjectID == placed.id,
            isActivated: model.activatedSubjectID == placed.id,
            dragOffset: drag
        )
        .accessibilityAction(named: Text(L10n.EvidenceGraph.editAccessibility)) {
            model.beginEdit(subjectID: placed.id)
        }
        .offset(x: layout.width, y: layout.height)
    }

    private var emptyOverlay: some View {
        PVEmptyState(
            icon: .shapes,
            title: L10n.EvidenceGraph.emptyTitle,
            message: String(localized: L10n.EvidenceGraph.emptyMessage)
        )
        .frame(width: 420)
        .position(x: contentSize.width / 2, y: contentSize.height / 2)
    }
}
