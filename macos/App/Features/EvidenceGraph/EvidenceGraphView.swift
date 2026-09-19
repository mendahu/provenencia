import AppKit
import SwiftUI

/// Workspace destination for a Source’s Evidence graph.
///
/// Product place under Sources (`sourceSurface: .graph`). Composes the
/// reusable [`GraphCanvas`](../GraphCanvas/) shell with primary subject cards,
/// palette place/create, and drag persist (S6-02 / S6-03).
struct EvidenceGraphView: View {
    /// Large enough to pan; Source-scoped graphs stay small (design note §7.1).
    private static let contentSize = CGSize(width: 4_000, height: 4_000)

    let sourceID: String
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String

    @State private var model: EvidenceGraphModel
    @State private var selectedSubjectID: String?
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
                    sourceTitle: sourceTitle,
                    contentSize: Self.contentSize,
                    selectedSubjectID: $selectedSubjectID
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
            } else {
                NSCursor.pop()
            }
        }
        .onDisappear {
            if model.armedKind != nil {
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
    let sourceTitle: String
    let contentSize: CGSize
    @Binding var selectedSubjectID: String?
    @FocusState private var canvasFocused: Bool

    private var snapshot: SourceGraphSnapshot {
        handle.value ?? SourceGraphSnapshot(sourceId: "")
    }

    private var subjects: [SourceGraphPlacedSubject] {
        snapshot.subjects
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
        .focusable()
        .focused($canvasFocused)
        .onAppear { canvasFocused = true }
        .onKeyPress(phases: .down) { press in
            handleKey(press)
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(accessibilityGraphLabel)
        .accessibilityRepresentation { accessibilityList }
        .accessibilityRotor(String(localized: L10n.EvidenceGraph.subjectsRotor)) {
            ForEach(subjects) { placed in
                AccessibilityRotorEntry(
                    EvidenceSubjectCard.accessibilityLabel(for: placed),
                    id: placed.id
                ) {
                    selectedSubjectID = placed.id
                }
            }
        }
        .pvDialog(
            isPresented: createPresented,
            copy: PVDialogCopy(
                title: model.armedKind.map { model.createDialogTitle(for: $0) }
                    ?? L10n.EvidenceGraph.addPersonTitle,
                confirm: L10n.EvidenceGraph.createConfirm,
                cancel: L10n.EvidenceGraph.createCancel
            ),
            isRunning: model.isSaving,
            confirmDisabled: model.draft.label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
            accessibilityIdentifierPrefix: "evidenceGraph.create",
            onConfirm: {
                Task {
                    if let id = await model.confirmCreate() {
                        selectedSubjectID = id
                        canvasFocused = true
                    }
                }
            }
        ) {
            createForm
        }
    }

    private var accessibilityGraphLabel: Text {
        if let kind = model.armedKind {
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
            Text(L10n.EvidenceGraph.subjectCount(count: subjects.count))
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textFaint)
        }
        .padding(.horizontal, PVSpacing.space7)
        .padding(.vertical, PVSpacing.space5)
        .accessibilityElement(children: .combine)
    }

    private var canvasPane: some View {
        ZStack(alignment: .top) {
            GraphCanvasScrollView(contentSize: contentSize) {
                document
            }

            VStack(spacing: PVSpacing.space3) {
                EvidenceGraphPalette(model: model)
                if let kind = model.armedKind {
                    armedBanner(kind)
                }
            }
            .padding(.top, PVSpacing.space5)
        }
    }

    private func armedBanner(_ kind: EvidencePrimaryKind) -> some View {
        HStack(spacing: 12) {
            Text(model.armedHint(for: kind))
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

    private var document: some View {
        ZStack(alignment: .topLeading) {
            GraphCanvasGridView(contentSize: contentSize)
                .accessibilityHidden(true)

            Color.clear
                .contentShape(Rectangle())
                .frame(width: contentSize.width, height: contentSize.height)
                .onTapGesture(coordinateSpace: .local) { location in
                    if model.armedKind != nil {
                        model.beginCreate(at: location)
                    } else {
                        selectedSubjectID = nil
                    }
                }
                .onContinuousHover { phase in
                    switch phase {
                    case .active(let location):
                        model.updateHover(contentPoint: location)
                    case .ended:
                        model.clearHover()
                    }
                }

            ForEach(subjects) { placed in
                cardView(for: placed)
            }

            if let ghost = model.ghostPlacedSubject() {
                EvidenceSubjectCard.ghost(placed: ghost)
                    .position(
                        GraphCanvasGridMapping.contentPoint(
                            gridX: ghost.gridX,
                            gridY: ghost.gridY
                        )
                    )
            }

            if subjects.isEmpty, handle.status == .ready, model.armedKind == nil {
                emptyOverlay
            }
        }
        .frame(width: contentSize.width, height: contentSize.height)
    }

    private func cardView(for placed: SourceGraphPlacedSubject) -> some View {
        let origin = GraphCanvasGridMapping.contentPoint(
            gridX: placed.gridX,
            gridY: placed.gridY
        )
        let dragging = model.draggingSubjectID == placed.id
        let position = CGPoint(
            x: origin.x + (dragging ? model.dragTranslation.width : 0),
            y: origin.y + (dragging ? model.dragTranslation.height : 0)
        )
        return EvidenceSubjectCard(
            placed: placed,
            isSelected: selectedSubjectID == placed.id,
            isDragging: dragging,
            dragEnabled: model.armedKind == nil && !model.isCreating,
            onSelect: { selectedSubjectID = placed.id },
            onDragChanged: { translation in
                if model.draggingSubjectID != placed.id {
                    model.beginDrag(subjectID: placed.id)
                    selectedSubjectID = placed.id
                }
                model.updateDrag(translation: translation)
            },
            onDragEnded: {
                Task {
                    _ = await model.endDrag(
                        subjectID: placed.id,
                        originGridX: placed.gridX,
                        originGridY: placed.gridY
                    )
                }
            }
        )
        .position(position)
        .zIndex(dragging ? 1 : 0)
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

    @ViewBuilder
    private var createForm: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if let kind = model.armedKind {
                let style = EvidenceSubjectKindStyle.forKind(kind)
                HStack(spacing: 8) {
                    ZStack {
                        RoundedRectangle(cornerRadius: 4, style: .continuous)
                            .fill(style.chip)
                        PVSubjectIcon(kind: iconKind(kind), size: 15)
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

    private func iconKind(_ kind: EvidencePrimaryKind) -> PVSubjectIconKind {
        switch kind {
        case .person: .person
        case .event: .event
        case .place: .place
        }
    }

    private var accessibilityList: some View {
        List(selection: $selectedSubjectID) {
            ForEach(subjects) { placed in
                Text(verbatim: EvidenceSubjectCard.accessibilityLabel(for: placed))
                    .tag(placed.id)
                    .accessibilityIdentifier("evidenceGraph.subject.\(placed.id)")
                    .accessibilityAddTraits(
                        selectedSubjectID == placed.id ? [.isSelected] : []
                    )
            }
        }
    }

    private func handleKey(_ press: KeyPress) -> KeyPress.Result {
        if model.isCreating { return .ignored }

        if press.key == .escape {
            if model.armedKind != nil {
                model.disarm()
                return .handled
            }
            if selectedSubjectID != nil {
                selectedSubjectID = nil
                return .handled
            }
            return .ignored
        }

        guard model.armedKind == nil,
              let selectedID = selectedSubjectID,
              let placed = subjects.first(where: { $0.id == selectedID })
        else { return .ignored }

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
                subjectID: placed.id,
                fromGridX: placed.gridX,
                fromGridY: placed.gridY,
                deltaX: dx,
                deltaY: dy
            )
        }
        return .handled
    }
}
