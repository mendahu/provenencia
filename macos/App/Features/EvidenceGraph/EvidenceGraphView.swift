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

    @State private var model: EvidenceGraphModel
    @State private var sourceTitle: String = ""

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
                    contentSize: Self.contentSize
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
            await refreshSourceTitle()
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

    private func refreshSourceTitle() async {
        let listKey = CatalogQueryKey.sourcesList(project: session.projectKey)
        if let handle: QueryHandle<[CatalogSource]> = session.queryHandle(listKey),
           let sources = handle.value,
           let match = sources.first(where: { $0.id == sourceID })
        {
            sourceTitle = match.title
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
        }
    }
}

// MARK: - Content

private struct EvidenceGraphContent: View {
    @Bindable var handle: QueryHandle<SourceGraphSnapshot>
    @Bindable var model: EvidenceGraphModel
    let sourceID: String
    let sourceTitle: String
    let contentSize: CGSize
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

    private var createPresented: Binding<Bool> {
        Binding(
            get: { model.isCreating },
            set: { newValue in
                if !newValue { model.cancelCreate() }
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
        .pvDialog(
            isPresented: createPresented,
            copy: PVDialogCopy(
                title: model.createDialogTitle(),
                confirm: L10n.EvidenceGraph.createConfirm,
                cancel: L10n.EvidenceGraph.createCancel
            ),
            isRunning: model.isSaving,
            confirmDisabled: model.draft.label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
            accessibilityIdentifierPrefix: "evidenceGraph.create",
            onConfirm: {
                Task {
                    if let id = await model.confirmCreate() {
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
                    contentSize: contentSize
                )
            }

            HStack(alignment: .top, spacing: PVSpacing.space5) {
                EvidenceGraphPalette(model: model, focus: $focus)
                Spacer(minLength: 0)
                if case .placing(let kind) = model.inputMode {
                    armedBanner(Text(model.armedHint(for: kind)))
                } else if case .connecting = model.inputMode {
                    armedBanner(Text(model.connectArmedHint))
                }
            }
            .padding(.horizontal, PVSpacing.space5)
            .padding(.top, PVSpacing.space5)
        }
    }

    private func armedBanner(_ hint: Text) -> some View {
        HStack(spacing: 12) {
            hint
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textPrimary)
            Text(L10n.EvidenceGraph.armedEscHint)
                .font(PVFont.mono(size: 11))
                .foregroundStyle(PVColor.textMuted)
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
        .allowsHitTesting(false)
    }

    @ViewBuilder
    private var createForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if let bridgeKind = model.pendingBridgeKind {
                HStack(spacing: 8) {
                    ZStack {
                        RoundedRectangle(cornerRadius: 4, style: .continuous)
                            .fill(PVColor.surfacePage)
                        PVSubjectIcon(kind: bridgeKind.subjectIconKind, size: 15)
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
            } else if let kind = model.armedKind {
                let style = EvidenceSubjectKindStyle.forKind(kind)
                HStack(spacing: 8) {
                    ZStack {
                        RoundedRectangle(cornerRadius: 4, style: .continuous)
                            .fill(style.chip)
                        PVSubjectIcon(kind: kind.subjectIconKind, size: 15)
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
        if model.isCreating { return .ignored }

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

// MARK: - Document (inside NSHostingView)

private struct EvidenceGraphDocument: View {
    @Bindable var handle: QueryHandle<SourceGraphSnapshot>
    @Bindable var model: EvidenceGraphModel
    let contentSize: CGSize
    @Environment(\.graphCanvasViewport) private var viewport
    @FocusState private var focusedSubjectID: String?

    @State private var panTranslation: CGSize = .zero
    @State private var isPanning = false

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
            gridLayer

            EvidenceGraphEdgeLayer(
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
                EvidenceSubjectCard.ghost(placed: ghost)
                    .position(EvidenceSubjectCard.contentCenter(gridX: ghost.gridX, gridY: ghost.gridY))
            }

            if subjects.isEmpty, bridges.isEmpty, handle.status == .ready, inputMode == .idle {
                emptyOverlay
            }

            if case .placing = inputMode {
                placeOverlay
            } else if case .connecting = inputMode {
                connectHoverOverlay
            }
        }
        .frame(width: contentSize.width, height: contentSize.height)
        .coordinateSpace(name: EvidenceSubjectCard.documentCoordinateSpace)
        .onChange(of: focusedSubjectID) { _, id in
            model.selectSubject(id: id)
        }
        .onChange(of: inputMode) { _, mode in
            if case .placing = mode {
                focusedSubjectID = nil
                model.selectSubject(id: nil)
            }
        }
        .onChange(of: viewport?.documentContainsKeyboardFocus ?? false) { _, inside in
            if !inside {
                focusedSubjectID = nil
                model.selectSubject(id: nil)
            }
        }
    }

    @ViewBuilder
    private var gridLayer: some View {
        let grid = GraphCanvasGridView(contentSize: contentSize)
            .accessibilityHidden(true)
            .contentShape(Rectangle())

        switch inputMode {
        case .idle:
            grid
                .onTapGesture {
                    model.selectSubject(id: nil)
                    focusedSubjectID = nil
                }
                .gesture(backgroundPanGesture)
        case .placing, .connecting:
            grid.allowsHitTesting(false)
        }
    }

    private var placeOverlay: some View {
        Color.clear
            .contentShape(Rectangle())
            .frame(width: contentSize.width, height: contentSize.height)
            .onTapGesture {
                if let point = viewport?.contentPointUnderCursor() {
                    model.beginCreate(at: point)
                }
            }
            .onContinuousHover { phase in
                switch phase {
                case .active:
                    if let point = viewport?.contentPointUnderCursor() {
                        model.updateHover(contentPoint: point)
                    }
                case .ended:
                    model.clearHover()
                }
            }
    }

    private var connectHoverOverlay: some View {
        Color.clear
            .frame(width: contentSize.width, height: contentSize.height)
            .allowsHitTesting(false)
            .onContinuousHover { phase in
                switch phase {
                case .active:
                    if let point = viewport?.contentPointUnderCursor() {
                        model.updateConnectHover(contentPoint: point)
                    }
                case .ended:
                    model.clearHover()
                }
            }
    }

    private var backgroundPanGesture: some Gesture {
        DragGesture(minimumDistance: 4, coordinateSpace: .global)
            .onChanged { value in
                guard let viewport else { return }
                if !isPanning {
                    isPanning = true
                    NSCursor.closedHand.push()
                }
                let delta = CGSize(
                    width: value.translation.width - panTranslation.width,
                    height: value.translation.height - panTranslation.height
                )
                panTranslation = value.translation
                viewport.panByViewDelta(delta)
            }
            .onEnded { _ in
                panTranslation = .zero
                if isPanning {
                    NSCursor.pop()
                    isPanning = false
                }
            }
    }

    @ViewBuilder
    private func cardView(for placed: SourceGraphPlacedSubject) -> some View {
        let idle = inputMode == .idle
        let connecting = inputMode == .connecting
        let offset = EvidenceSubjectCard.topLeadingOffset(gridX: placed.gridX, gridY: placed.gridY)
        EvidenceSubjectCard(
            placed: placed,
            isSelected: model.selectedSubjectID == placed.id,
            isActivated: model.activatedSubjectID == placed.id,
            isConnectingFrom: model.connectOriginID == placed.id,
            dragEnabled: idle,
            keyboardFocus: $focusedSubjectID,
            onSelect: {
                if connecting {
                    model.handleConnectPick(
                        subjectID: placed.id,
                        kind: placed.kind,
                        label: placed.subject.label
                    )
                } else {
                    model.selectSubject(id: placed.id)
                }
            },
            onActivate: {
                if connecting {
                    model.handleConnectPick(
                        subjectID: placed.id,
                        kind: placed.kind,
                        label: placed.subject.label
                    )
                } else {
                    model.activateSubject(id: placed.id)
                }
            },
            onEscape: {
                if model.armedConnect {
                    model.disarm()
                } else if model.activatedSubjectID != nil {
                    model.deactivateSubject()
                } else {
                    focusedSubjectID = nil
                    model.selectSubject(id: nil)
                }
            },
            onDragEnded: idle
                ? { delta in
                    _ = model.commitDrag(
                        subjectID: placed.id,
                        originGridX: placed.gridX,
                        originGridY: placed.gridY,
                        documentDelta: delta
                    )
                }
                : nil
        )
        .offset(x: offset.width, y: offset.height)
    }

    @ViewBuilder
    private func bridgeCardView(for placed: SourceGraphPlacedBridge) -> some View {
        let idle = inputMode == .idle
        let offset = EvidenceBridgeCard.topLeadingOffset(gridX: placed.gridX, gridY: placed.gridY)
        EvidenceBridgeCard(
            placed: placed,
            isSelected: model.selectedSubjectID == placed.id,
            isActivated: model.activatedSubjectID == placed.id,
            dragEnabled: idle,
            keyboardFocus: $focusedSubjectID,
            onSelect: {
                // Bridges are not Connect endpoints.
                guard inputMode != .connecting else { return }
                model.selectSubject(id: placed.id)
            },
            onActivate: {
                guard inputMode != .connecting else { return }
                model.activateSubject(id: placed.id)
            },
            onEscape: {
                if model.activatedSubjectID != nil {
                    model.deactivateSubject()
                } else {
                    focusedSubjectID = nil
                    model.selectSubject(id: nil)
                }
            },
            onDragEnded: idle
                ? { delta in
                    _ = model.commitDrag(
                        subjectID: placed.id,
                        originGridX: placed.gridX,
                        originGridY: placed.gridY,
                        documentDelta: delta
                    )
                }
                : nil
        )
        .offset(x: offset.width, y: offset.height)
    }

    private var emptyOverlay: some View {
        PVEmptyState(
            icon: .shapes,
            title: L10n.EvidenceGraph.emptyTitle,
            message: String(localized: L10n.EvidenceGraph.emptyMessage)
        )
        .frame(width: 420)
        .position(x: contentSize.width / 2, y: contentSize.height / 2)
        .allowsHitTesting(false)
    }
}
