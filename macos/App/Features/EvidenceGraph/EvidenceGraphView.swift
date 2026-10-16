import AppKit
import SwiftUI

/// Workspace destination for a Source’s Evidence graph.
///
/// Product place under Sources (`sourceSurface: .graph`). Composes the
/// reusable [`GraphCanvas`](../GraphCanvas/) shell with primary + bridge cards,
/// palette place/create/connect, and drag persist (S6-02–S6-04).
struct EvidenceGraphView: View {
    let sourceID: String
    let session: WorkspaceSession
    let store: any GenealogyStore
    let userID: String

    @Environment(WorkspaceNavigation.self) private var navigation
    @State private var model: EvidenceGraphModel

    private var graphKey: CatalogQueryKey {
        CatalogQueryKey.sourceGraph(project: session.projectKey, sourceId: sourceID)
    }

    private var fieldsKey: CatalogQueryKey {
        CatalogQueryKey.propertiesWorkspace(project: session.projectKey)
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
            if let graphHandle: QueryHandle<SourceGraphRows> = session.queryHandle(graphKey),
               let fieldsHandle: QueryHandle<PropertiesSnapshot> = session.queryHandle(fieldsKey)
            {
                EvidenceGraphContent(
                    graphHandle: graphHandle,
                    fieldsHandle: fieldsHandle,
                    model: model,
                    sourceID: sourceID,
                    sourceTitle: model.sourceTitle,
                    sourceRef: model.sourceRef,
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
            await model.prepare()
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
}
// MARK: - Content

private struct EvidenceGraphContent: View {
    @Bindable var graphHandle: QueryHandle<SourceGraphRows>
    @Bindable var fieldsHandle: QueryHandle<PropertiesSnapshot>
    @Bindable var model: EvidenceGraphModel
    let sourceID: String
    let sourceTitle: String
    let sourceRef: String?
    let navigation: WorkspaceNavigation
    @FocusState private var focus: EvidenceGraphFocus?

    private var snapshot: SourceGraphSnapshot {
        model.displaySnapshot(rows: graphHandle.value, types: fieldsHandle.value?.types ?? [])
    }

    private var subjects: [SourceGraphPlacedSubject] {
        snapshot.subjects
    }

    private var bridges: [SourceGraphPlacedBridge] {
        snapshot.bridges
    }

    /// Downtown 4000×4000, one screen in from grid zero, until a card is within
    /// one card-width of an edge. Left and top spend that reserve in 800-point
    /// steps. A drag that has not landed yet can be larger.
    private var canvas: EvidenceGraphCanvas {
        var frames = EvidenceGraphPlacement.cardFrames(in: snapshot)
        if let ghost = model.ghostPlacedSubject() {
            frames.append(
                EvidenceGraphPlacement.subjectFrame(gridX: ghost.gridX, gridY: ghost.gridY)
            )
        }
        let fitted = EvidenceGraphPlacement.canvas(fitting: frames)
        guard let live = model.liveCanvas else { return fitted }
        return EvidenceGraphPlacement.union(fitted, live)
    }

    private var contentSize: CGSize { canvas.size }

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
        .vocabularyToastOverlay($model.toast, identifier: "evidenceGraph.toast")
        .onChange(of: canvas.origin, initial: true) { _, origin in
            model.canvasOrigin = origin
        }
        .pvFormDialog(
            isPresented: sheetPresented,
            copy: PVFormDialogCopy(
                title: model.createDialogTitle(),
                confirm: model.sheetConfirmLabel(),
                cancel: L10n.EvidenceGraph.createCancel
            ),
            isRunning: model.isSaving,
            isDirty: model.formSheetIsDirty,
            confirmDisabled: model.isEditingBridge
                ? false
                : model.draft.label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
            accessibilityIdentifierPrefix: model.editingSubjectID != nil
                ? "evidenceGraph.edit"
                : "evidenceGraph.create",
            onConfirm: {
                Task {
                    if model.editingSubjectID != nil {
                        _ = await model.confirmEdit()
                    } else if let id = await model.confirmCreate() {
                        if let location = model.consumeComposerHandoff() {
                            navigation.go(to: location)
                        } else {
                            model.selectSubject(id: id)
                        }
                    }
                }
            }
        ) {
            createForm
        }
        .pvDeleteImpact(
            flow: model.deleteImpact,
            accessibilityIdentifierPrefix: "evidenceGraph.deleteImpact",
            onConfirm: {
                Task { _ = await model.confirmPendingImpact() }
            },
            onNavigate: { location in
                navigation.go(to: location)
            }
        )
        .pvConfirm(
            item: $model.pendingUnpromote,
            copy: { item in
                PVConfirmCopy(
                    title: L10n.EvidenceGraph.unpromoteConfirmTitle(name: item.name),
                    message: L10n.string(L10n.EvidenceGraph.unpromoteConfirmMessage),
                    confirm: L10n.EvidenceGraph.unpromote,
                    cancel: L10n.EvidenceGraph.unpromoteKeep
                )
            },
            tone: .irreversible,
            isRunning: model.isUnpromoting,
            accessibilityIdentifierPrefix: "evidenceGraph.unpromote",
            onConfirm: { Task { await model.confirmUnpromote() } },
            detail: unpromoteDetail
        )
    }

    /// The bridges that come off with the filing, then a write error. Nothing
    /// when there are neither.
    @ViewBuilder
    private func unpromoteDetail(_ item: EvidenceGraphModel.PendingUnpromote) -> some View {
        if !item.bridges.isEmpty || item.error != nil {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                if !item.bridges.isEmpty {
                    Text(L10n.EvidenceGraph.unpromoteConfirmBridgesHeader)
                        .pvMicroCaps()
                        .foregroundStyle(PVColor.textMuted)
                    VStack(alignment: .leading, spacing: PVSpacing.space2) {
                        ForEach(Array(item.bridges.enumerated()), id: \.offset) { _, sentence in
                            Text(verbatim: sentence)
                                .font(PVFont.body(size: PVTypeScale.bodySmall))
                                .foregroundStyle(PVColor.textPrimary)
                                .frame(maxWidth: .infinity, alignment: .leading)
                        }
                    }
                    Text(L10n.EvidenceGraph.unpromoteConfirmBridgesNote)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textSecondary)
                }
                if let error = item.error {
                    PVCallout(tone: .danger, message: error)
                }
            }
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
        EvidenceGraphHeader(
            model: model,
            navigation: navigation,
            sourceTitle: sourceTitle,
            sourceRef: sourceRef,
            subjectCount: subjects.count + bridges.count
        )
    }

    private var canvasPane: some View {
        ZStack(alignment: .top) {
            // The Subjects and Links rotors live on the canvas group (AppKit),
            // next to the card views they navigate.
            GraphCanvasScrollView(
                contentSize: contentSize,
                contentOrigin: canvas.origin,
                contentID: sourceID,
                restoredCamera: model.session.graphCamera(for: sourceID),
                onCameraChange: { camera in
                    model.session.setGraphCamera(camera, for: sourceID)
                },
                accessibilityLabel: L10n.string(L10n.Workspace.evidenceGraphTitle)
            ) {
                EvidenceGraphDocument(
                    graphHandle: graphHandle,
                    fieldsHandle: fieldsHandle,
                    model: model,
                    canvas: canvas,
                    navigation: navigation
                )
            }

            HStack(alignment: .top, spacing: PVSpacing.space5) {
                EvidenceGraphPalette(model: model, focus: $focus)
                Spacer(minLength: 0)
                if model.pendingImpact == nil, let deleteError = model.deleteError {
                    PVCallout(tone: .danger, message: deleteError)
                        .frame(width: 360, alignment: .trailing)
                } else if model.pendingImpact == nil, model.deleteError == nil,
                          let fileOpenError = model.fileOpenError,
                          model.resolvedFileArtifactID == model.selectedFileArtifactID
                {
                    PVCallout(
                        tone: .danger,
                        title: L10n.EvidenceGraph.fileNotFoundTitle,
                        message: fileOpenError
                    )
                    .frame(width: 360, alignment: .trailing)
                    .accessibilityIdentifier("evidenceGraph.fileOpen.callout")
                } else if !model.canCite {
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

            if EvidenceCitedPropertyMarks.graphHasConflict(subjects: subjects, bridges: bridges) {
                VStack {
                    Spacer(minLength: 0)
                    HStack {
                        EvidenceGraphRowMarksLegend()
                        Spacer(minLength: 0)
                    }
                }
                .padding(.horizontal, PVSpacing.space5)
                // space5 would sit on the canvas scroller; clear it plus a gap.
                .padding(.bottom, PVSpacing.space8)
            }
        }
    }

    private var noArtifactCallout: some View {
        PVCallout(
            tone: .warning,
            title: L10n.EvidenceGraph.noArtifactTitle,
            message: L10n.string(L10n.EvidenceGraph.noArtifactMessage)
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
            if model.editingSubjectID == nil, let kind = model.armedKind {
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
            if !model.isEditingBridge {
                PVField(
                    label: L10n.EvidenceGraph.labelField,
                    hint: L10n.EvidenceGraph.workingLabelHint,
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
            ?? L10n.string(model.toolName(for: kind))
    }

    private func handleKey(_ press: KeyPress) -> KeyPress.Result {
        let command: EvidenceGraphModel.Key?
        switch press.key {
        case .escape: command = .escape
        case .return: command = .return
        case .space: command = .space
        case .leftArrow: command = .arrow(dx: -1, dy: 0)
        case .rightArrow: command = .arrow(dx: 1, dy: 0)
        case .upArrow: command = .arrow(dx: 0, dy: -1)
        case .downArrow: command = .arrow(dx: 0, dy: 1)
        default: command = nil
        }
        guard let command else { return .ignored }
        let effect = model.handleKey(command, hasFocus: focus != nil)
        if effect.clearFocus {
            focus = nil
        }
        if !effect.moves.isEmpty {
            model.moveSubjects(effect.moves)
        }
        return effect.handled ? .handled : .ignored
    }
}

// MARK: - Header

/// Title and the trailing file controls. Open, Promote all, and Jump are
/// either full labels or glyphs. The choice is the window width against the
/// width of the full row, measured once off to the side. `ViewThatFits` is
/// not used: its choice changes the width it is offered, so the row flips
/// every pass and the canvas never gets a quiet layout.
private struct EvidenceGraphHeader: View {
    @Bindable var model: EvidenceGraphModel
    let navigation: WorkspaceNavigation
    let sourceTitle: String
    let sourceRef: String?
    let subjectCount: Int

    @State private var headerWidth: CGFloat = 0
    @State private var labeledWidth: CGFloat = 0

    /// Full labels until the measured row is wider than the header.
    private var iconOnly: Bool {
        headerWidth > 1 && labeledWidth > headerWidth + 1
    }

    var body: some View {
        HStack(alignment: .center, spacing: PVSpacing.space5) {
            pageTitle
                .fixedSize(horizontal: true, vertical: false)
            if !sourceTitle.isEmpty {
                sourceTitleText
                    .lineLimit(1)
                    .layoutPriority(-1)
            }
            subjectCountLabel
                .fixedSize(horizontal: true, vertical: false)
            Spacer(minLength: PVSpacing.space3)
            actionCluster(iconOnly: iconOnly)
        }
        .background {
            GeometryReader { proxy in
                Color.clear.preference(key: EvidenceGraphHeaderWidthKey.self, value: proxy.size.width)
            }
        }
        .background(alignment: .leading) {
            labeledProbe
        }
        .padding(.horizontal, PVSpacing.space7)
        .padding(.vertical, PVSpacing.space5)
        .onPreferenceChange(EvidenceGraphHeaderWidthKey.self) { width in
            guard width > 1, abs(width - headerWidth) > 0.5 else { return }
            headerWidth = width
        }
        .onPreferenceChange(EvidenceGraphLabeledHeaderWidthKey.self) { width in
            guard width > 1, abs(width - labeledWidth) > 0.5 else { return }
            labeledWidth = width
        }
    }

    /// The labeled row at its natural width, hidden. Its geometry is the
    /// cutoff for `iconOnly` and does not depend on which buttons are showing.
    private var labeledProbe: some View {
        HStack(spacing: PVSpacing.space5) {
            pageTitle
            if !sourceTitle.isEmpty {
                sourceTitleText
                    .fixedSize(horizontal: true, vertical: false)
            }
            subjectCountLabel
            Color.clear
                .frame(width: PVSpacing.space3)
                .accessibilityHidden(true)
            actionCluster(iconOnly: false)
        }
        .fixedSize(horizontal: true, vertical: false)
        .background {
            GeometryReader { proxy in
                Color.clear.preference(key: EvidenceGraphLabeledHeaderWidthKey.self, value: proxy.size.width)
            }
        }
        .hidden()
        .allowsHitTesting(false)
        .accessibilityHidden(true)
    }

    private var pageTitle: some View {
        Text(L10n.Workspace.evidenceGraphTitle)
            .font(PVFont.display(size: PVTypeScale.h1, weight: PVFontWeight.medium))
            .foregroundStyle(PVColor.textDisplay)
    }

    private var sourceTitleText: some View {
        Text(verbatim: sourceTitle)
            .font(PVFont.mono(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textMuted)
    }

    private var subjectCountLabel: some View {
        HStack(alignment: .firstTextBaseline, spacing: 0) {
            Text(verbatim: "(")
            Text(verbatim: L10n.EvidenceGraph.subjectCount(count: subjectCount))
            Text(verbatim: ")")
        }
        .font(PVFont.mono(size: PVTypeScale.caption))
        .foregroundStyle(PVColor.textFaint)
    }

    private func actionCluster(iconOnly: Bool) -> some View {
        HStack(spacing: PVSpacing.space3) {
            if !model.fileRows.isEmpty {
                fileSelect(model.fileRows)
            }
            headerButtons(iconOnly: iconOnly)
        }
        .fixedSize(horizontal: true, vertical: false)
    }

    private func headerButtons(iconOnly: Bool) -> some View {
        HStack(spacing: PVSpacing.space3) {
            if !model.fileRows.isEmpty {
                openFileButton(model.fileRows, iconOnly: iconOnly)
            }
            promoteAllButton(iconOnly: iconOnly)
            jumpToSourcePageButton(iconOnly: iconOnly)
        }
    }

    private func fileSelect(_ rows: [EvidenceGraphFiles.Row]) -> some View {
        let resolved = model.resolvedFileArtifactID
        let projectDir = model.session.projectKey.projectDir
        return PVSelect(
            selection: Binding(
                get: { model.resolvedFileArtifactID },
                set: { model.selectFile($0) }
            ),
            options: rows.map { PVSelectOption(value: $0.id, label: $0.title) },
            size: .sm,
            menuWidth: 360,
            maxVisibleRows: 6,
            rowHeight: 64,
            accessibilityLabel: L10n.EvidenceGraph.artifactFile,
            accessibilityIdentifier: "evidenceGraph.header.fileSelect",
            closedHelp: rows.first(where: { $0.id == resolved })?.displayFilename
        ) { option in
            if let row = rows.first(where: { $0.id == option.id }) {
                EvidenceGraphFileMenuRow(
                    row: row,
                    projectDir: projectDir,
                    citationCount: model.citationCount(for: row.id)
                )
            } else {
                Text(verbatim: option.label)
            }
        }
        .frame(width: 208)
    }

    private func openFileButton(_ rows: [EvidenceGraphFiles.Row], iconOnly: Bool) -> some View {
        let resolved = model.resolvedFileArtifactID
        let filename = rows.first(where: { $0.id == resolved })?.displayFilename ?? ""
        return PVButton(
            L10n.EvidenceGraph.openFile,
            variant: .secondary,
            size: .sm,
            icon: .externalLink,
            iconOnly: iconOnly
        ) {
            model.openSelectedFile()
        }
        .help(Text(L10n.EvidenceGraph.openFileHelp))
        .accessibilityLabel(Text(verbatim: L10n.EvidenceGraph.openFileAccessibility(filename: filename)))
        .accessibilityIdentifier("evidenceGraph.header.openFile")
    }

    private func promoteAllButton(iconOnly: Bool) -> some View {
        PVButton(
            L10n.EvidenceGraph.promoteAll,
            variant: .secondary,
            size: .sm,
            icon: .promote,
            iconOnly: iconOnly
        ) {
            if let location = model.promoteAllLocation() {
                navigation.go(to: location)
            }
        }
        .disabled(model.promoteAllLocation() == nil)
        .accessibilityIdentifier("evidenceGraph.header.promoteAll")
    }

    private func jumpToSourcePageButton(iconOnly: Bool) -> some View {
        let title = sourceTitle.trimmingCharacters(in: .whitespacesAndNewlines)
        return PVButton(
            L10n.EvidenceGraph.jumpToSourcePage,
            variant: .secondary,
            size: .sm,
            icon: .library,
            iconOnly: iconOnly
        ) {
            navigation.go(to: model.sourcePageLocation(title: sourceTitle, ref: sourceRef))
        }
        .accessibilityLabel(
            title.isEmpty
                ? Text(L10n.EvidenceGraph.jumpToSourcePage)
                : Text(verbatim: L10n.EvidenceGraph.jumpToSourcePageAccessibility(title: title))
        )
        .accessibilityIdentifier("evidenceGraph.header.jumpToSourcePage")
    }
}

private struct EvidenceGraphHeaderWidthKey: PreferenceKey {
    static let defaultValue: CGFloat = 0
    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = max(value, nextValue())
    }
}

private struct EvidenceGraphLabeledHeaderWidthKey: PreferenceKey {
    static let defaultValue: CGFloat = 0
    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = max(value, nextValue())
    }
}

// MARK: - Document (paint-only; AppKit owns pointer)

/// Reads live drag offsets so the document can grow before the drop, without
/// the card `ForEach` following every pointer event.
private struct EvidenceGraphDragCanvasSize: View {
    let snapshot: SourceGraphSnapshot
    @Bindable var pointer: GraphCanvasPointerController
    @Bindable var model: EvidenceGraphModel

    var body: some View {
        let board = EvidenceGraphPlacement.canvas(
            fitting: EvidenceGraphPlacement.cardFrames(in: snapshot, dragOffsets: pointer.offsets)
        )
        Color.clear
            .frame(width: 0, height: 0)
            .accessibilityHidden(true)
            .onChange(of: board) { _, board in
                guard pointer.draggingItemID != nil else { return }
                model.liveCanvas = board
            }
            .onChange(of: pointer.draggingItemID) { _, id in
                if id == nil {
                    model.liveCanvas = nil
                }
            }
    }
}

/// The selection rectangle, in document space so it scales with zoom. This is
/// the only reader of ``GraphCanvasPointerController/marqueeRect``: the document
/// body must not follow a property that changes every pointer event.
private struct EvidenceGraphMarquee: View {
    @Bindable var pointer: GraphCanvasPointerController

    var body: some View {
        if let rect = pointer.marqueeRect {
            Rectangle()
                .fill(PVColor.accent.opacity(0.09))
                .overlay(Rectangle().stroke(PVColor.accent, lineWidth: 1.5))
                .frame(width: rect.width, height: rect.height)
                .offset(x: rect.minX, y: rect.minY)
                .allowsHitTesting(false)
                .accessibilityHidden(true)
        }
    }
}

private struct EvidenceGraphDocument: View {
    @Bindable var graphHandle: QueryHandle<SourceGraphRows>
    @Bindable var fieldsHandle: QueryHandle<PropertiesSnapshot>
    @Bindable var model: EvidenceGraphModel
    let canvas: EvidenceGraphCanvas
    let navigation: WorkspaceNavigation
    @Environment(\.graphCanvasPointer) private var pointer
    @Environment(\.graphCanvasItems) private var items

    var body: some View {
        Group {
            if let pointer, let items {
                EvidenceGraphDocumentBody(
                    graphHandle: graphHandle,
                    fieldsHandle: fieldsHandle,
                    model: model,
                    canvas: canvas,
                    navigation: navigation,
                    pointer: pointer,
                    items: items
                )
            } else {
                Color.clear
                    .frame(width: canvas.size.width, height: canvas.size.height)
            }
        }
        .frame(width: canvas.size.width, height: canvas.size.height)
    }
}

/// The graph's base layer (grid, edges, ghost, rubber band, empty state) and
/// the controller for its cards.
///
/// Cards are ``GraphCanvasItem``s: AppKit hosts and positions each one, and a
/// drag moves its view without running this body. Nothing here may read
/// ``GraphCanvasPointerController/offsets`` — observation is per property, so a
/// read would re-run this body on every pointer event. ``EvidenceGraphEdgesHost``
/// is the one view that follows live offsets.
private struct EvidenceGraphDocumentBody: View {
    @Bindable var graphHandle: QueryHandle<SourceGraphRows>
    @Bindable var fieldsHandle: QueryHandle<PropertiesSnapshot>
    @Bindable var model: EvidenceGraphModel
    let canvas: EvidenceGraphCanvas
    let navigation: WorkspaceNavigation
    @Bindable var pointer: GraphCanvasPointerController
    let items: GraphCanvasItemLayer
    /// Painted card layouts; hit targets and edges follow these, not constants.
    @State private var cardLayouts = EvidenceCardLayoutStore()

    private var snapshot: SourceGraphSnapshot {
        model.displaySnapshot(rows: graphHandle.value, types: fieldsHandle.value?.types ?? [])
    }

    private var subjects: [SourceGraphPlacedSubject] {
        snapshot.subjects
    }

    private var bridges: [SourceGraphPlacedBridge] {
        snapshot.bridges
    }

    /// Card copy formatted with the snapshot; card views format nothing.
    private var text: EvidenceGraphText {
        model.displayText(rows: graphHandle.value, types: fieldsHandle.value?.types ?? [])
    }

    private var inputMode: EvidenceCanvasInputMode {
        model.inputMode
    }

    private var selectedBridgeID: String? {
        bridges.contains(where: { $0.id == model.selectedSubjectID })
            ? model.selectedSubjectID
            : nil
    }

    private var contentSize: CGSize { canvas.size }

    /// Grid-space frames are drawn at `frame − canvas.origin`.
    private func inDocument(_ rect: CGRect) -> CGRect {
        rect.offsetBy(dx: -canvas.origin.x, dy: -canvas.origin.y)
    }

    var body: some View {
        ZStack(alignment: .topLeading) {
            GraphCanvasGridView(contentSize: contentSize)
                .accessibilityHidden(true)

            EvidenceGraphEdgesHost(
                pointer: pointer,
                snapshot: snapshot,
                selectedBridgeID: selectedBridgeID,
                layouts: cardLayouts.layouts,
                documentOrigin: canvas.origin
            )
            .frame(width: contentSize.width, height: contentSize.height)

            if case .connecting = inputMode,
               let originID = model.connectOriginID,
               let origin = subjects.first(where: { $0.id == originID }),
               let cursor = model.connectHoverPoint
            {
                EvidenceGraphConnectRubberBand(
                    origin: origin,
                    cursor: cursor,
                    layout: cardLayouts.layout(for: origin.id),
                    documentOrigin: canvas.origin
                )
                    .frame(width: contentSize.width, height: contentSize.height)
            }

            if let ghost = model.ghostPlacedSubject() {
                let center = EvidenceSubjectCard.contentCenter(gridX: ghost.gridX, gridY: ghost.gridY)
                EvidenceGhostCard(card: .ghost(ghost, presentation: model.presentation(for: ghost.kind.rawValue)))
                    .id(ghost.kind.rawValue)
                    .fixedSize()
                    .allowsHitTesting(false)
                    .accessibilityHidden(true)
                    .position(CGPoint(x: center.x - canvas.origin.x, y: center.y - canvas.origin.y))
            }

            if subjects.isEmpty, bridges.isEmpty, graphHandle.status == .ready, inputMode == .idle {
                emptyOverlay
            }

            EvidenceGraphMarquee(pointer: pointer)

            EvidenceGraphDragCanvasSize(snapshot: snapshot, pointer: pointer, model: model)
        }
        .frame(width: contentSize.width, height: contentSize.height)
        .onAppear { wirePointer() }
        .onChange(of: inputMode) { _, _ in
            syncPointerMode()
        }
        .onChange(of: model.canCite) { _, _ in
            publishHitTargets()
        }
        .onChange(of: pointer.draggingItemID) { _, _ in
            publishHitTargets()
        }
        // Cards change only when what they draw does — never per drag frame.
        .onChange(of: cards, initial: true) { _, cards in
            pushItems(cards)
        }
        .onChange(of: canvas.origin) { _, _ in
            pushItems(cards)
            publishHitTargets()
        }
        .onChange(of: EvidenceGraphHitRefresh.token(subjects: subjects, bridges: bridges)) { _, _ in
            cardLayouts.retain(only: Set(subjects.map(\.id) + bridges.map(\.id)))
            publishHitTargets()
        }
        // A card's painted layout moved (wrap, new row, footer swap): re-anchor its hits.
        .onChange(of: cardLayouts.layouts) { _, _ in
            publishHitTargets()
        }
    }

    private func wirePointer() {
        syncPointerMode()
        publishHitTargets()
        let model = model
        let navigation = navigation
        items.onRotorSelect = { id in
            model.selectSubject(id: id)
        }
        pointer.onSelect = { id in
            model.selectSubject(id: id)
        }
        pointer.onDeselect = {
            model.selectSubject(id: nil)
        }
        pointer.onDragBegan = { id in
            model.beginGroupDrag(anchor: id)
        }
        pointer.onMarqueeEnded = { ids in
            model.replacePrimarySelection(ids)
        }
        pointer.onDragEnded = { id, delta in
            _ = model.commitGroupDrag(anchor: id, documentDelta: delta)
        }
        pointer.onPlace = { point in
            model.beginCreate(at: point)
        }
        pointer.onConnectPick = { id in
            Task {
                await model.handleConnectPick(subjectID: id)
                if let location = model.consumeComposerHandoff() {
                    navigation.go(to: location)
                }
            }
        }
        pointer.onCardAction = { id, actionID in
            if let location = model.performCardAction(subjectID: id, actionID: actionID) {
                navigation.go(to: location)
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
            case .idle, .selecting:
                break
            }
        }
    }

    private func syncPointerMode() {
        switch inputMode {
        case .idle:
            pointer.mode = .idle
        case .selecting:
            pointer.mode = .selecting
        case .placing:
            pointer.mode = .placing
            model.selectSubject(id: nil)
        case .connecting:
            pointer.mode = .connecting
        }
    }

    /// Hit rects at rest. Mid-drag the pointer owns the gesture and never
    /// hit-tests, and offsets are cleared before the drop's update.
    private func publishHitTargets() {
        var targets: [GraphCanvasHitTarget] = []
        for placed in bridges {
            targets.append(
                GraphCanvasHitTarget(
                    id: placed.id,
                    frame: inDocument(
                        EvidenceBridgeCard.contentFrame(
                            for: placed,
                            in: snapshot,
                            layout: cardLayouts.layout(for: placed.id)
                        )
                    ),
                    acceptsConnect: false,
                    actions: EvidenceBridgeCard.actionTargets(
                        for: placed,
                        in: snapshot,
                        canCite: model.canCite,
                        layout: cardLayouts.layout(for: placed.id)
                    ).map { action in
                        action.mapFrame(inDocument)
                    }
                )
            )
        }
        for placed in subjects {
            targets.append(
                GraphCanvasHitTarget(
                    id: placed.id,
                    frame: inDocument(
                        EvidenceSubjectCard.edgeFrame(
                            for: placed,
                            layout: cardLayouts.layout(for: placed.id)
                        )
                    ),
                    acceptsConnect: true,
                    actions: EvidenceSubjectCard.actionTargets(
                        for: placed,
                        canCite: model.canCite,
                        layout: cardLayouts.layout(for: placed.id)
                    ).map { action in
                        action.mapFrame(inDocument)
                    }
                )
            )
        }
        pointer.hitTargets = targets
    }

    /// What every card draws, bridges first so subjects paint above them.
    private var cards: [EvidenceCanvasCard] {
        let text = text
        let hovered = pointer.hoveredCardAction
        let pressed = pointer.pressedCardAction
        func state(_ id: String) -> EvidenceCanvasCard.State {
            EvidenceCanvasCard.State(
                isSelected: model.selectedSubjectIDs.contains(id),
                isActivated: model.activatedSubjectID == id,
                isConnectingFrom: model.connectOriginID == id,
                isDragging: pointer.draggingItemID == id || pointer.draggingPeerIDs.contains(id),
                canCite: model.canCite,
                hoveredActionID: hovered?.cardID == id ? hovered?.actionID : nil,
                pressedActionID: pressed?.cardID == id ? pressed?.actionID : nil
            )
        }
        return bridges.map { placed in
            EvidenceCanvasCard(
                placed: .bridge(placed),
                text: text[bridge: placed, in: snapshot],
                state: state(placed.id)
            )
        } + subjects.map { placed in
            EvidenceCanvasCard(
                placed: .subject(placed, model.presentation(for: placed.kind.rawValue)),
                text: text[subject: placed],
                state: state(placed.id),
                nameOrders: model.session.nameFormat
            )
        }
    }

    private func pushItems(_ cards: [EvidenceCanvasCard]) {
        let content = EvidenceCanvasCardContent(model: model, navigation: navigation, layouts: cardLayouts)
        let subjectsRotor = L10n.string(L10n.EvidenceGraph.subjectsRotor)
        let linksRotor = L10n.string(L10n.EvidenceGraph.linksRotor)
        items.update(cards.map { card in
            GraphCanvasItem(
                id: card.id,
                origin: CGPoint(
                    x: card.origin.x - canvas.origin.x,
                    y: card.origin.y - canvas.origin.y
                ),
                isRaised: card.state.isDragging || card.state.isActivated,
                rotor: GraphCanvasRotorEntry(
                    rotor: card.isBridge ? linksRotor : subjectsRotor,
                    label: card.text.accessibilityLabel
                ),
                model: card,
                makeView: { content.cardView(for: card) }
            )
        })
    }

    private var emptyOverlay: some View {
        PVEmptyState(
            icon: .shapes,
            title: L10n.EvidenceGraph.emptyTitle,
            message: L10n.string(L10n.EvidenceGraph.emptyMessage)
        )
        .frame(width: 420)
        .position(x: contentSize.width / 2, y: contentSize.height / 2)
    }
}

/// One file in the header select's menu: thumbnail, title, Cover, then
/// `filename · KIND · size` and the citation count.
private struct EvidenceGraphFileMenuRow: View {
    let row: EvidenceGraphFiles.Row
    let projectDir: String
    let citationCount: Int

    var body: some View {
        HStack(alignment: .center, spacing: 10) {
            CachedThumbnail(
                projectDir: projectDir,
                relPath: row.thumbnailRelPath,
                mediaType: row.mediaType,
                originalFilename: row.filename,
                size: 48,
                layout: .tile
            )
            VStack(alignment: .leading, spacing: 1) {
                HStack(spacing: PVSpacing.space3) {
                    Text(verbatim: row.title)
                        .font(
                            row.titleIsRef
                                ? PVFont.mono(size: PVTypeScale.caption)
                                : PVFont.body(size: PVTypeScale.caption)
                        )
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                    if row.isCover {
                        Text(L10n.Sources.coverBadge)
                            .pvMicroCaps()
                            .foregroundStyle(PVColor.textMuted)
                    }
                }
                Text(verbatim: L10n.EvidenceGraph.fileMeta(
                    filename: row.displayFilename,
                    kind: EvidenceGraphFiles.kindToken(mediaType: row.mediaType),
                    size: IngestMediaPolicy.formatByteSize(row.byteSize)
                ))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                Text(verbatim: L10n.CitationComposer.artifactCitationCount(count: citationCount))
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
            }
        }
        .padding(.vertical, PVSpacing.space2)
    }
}

/// Canvas legend for the ochre conflict bracket. One strip, bottom-left, while
/// any competing Property run is on the graph (S8-D3 Frame 09).
private struct EvidenceGraphRowMarksLegend: View {
    var body: some View {
        HStack(alignment: .center, spacing: 7) {
            EvidenceLegendConflictMark()
            Text(L10n.EvidenceGraph.legendConflict)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .fill(PVColor.surfaceCard)
                .overlay(
                    RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                        .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
                )
                .shadow(color: Color.black.opacity(0.08), radius: 8, y: 2)
        )
        .accessibilityElement(children: .combine)
        .accessibilityIdentifier("evidenceGraph.legend")
    }
}

private struct EvidenceLegendConflictMark: View {
    var body: some View {
        Path { path in
            path.move(to: CGPoint(x: 8, y: 2))
            path.addLine(to: CGPoint(x: 3, y: 2))
            path.addLine(to: CGPoint(x: 3, y: 14))
            path.addLine(to: CGPoint(x: 8, y: 14))
        }
        .stroke(
            PVColor.warning,
            style: StrokeStyle(lineWidth: 1.5, lineCap: .round, lineJoin: .round)
        )
        .frame(width: 12, height: 16)
        .accessibilityHidden(true)
    }
}

/// Canvas hit targets must refresh when observation ids change, not only when
/// the row count does. Saving a citation deletes and reinserts its observations.
enum EvidenceGraphHitRefresh {
    static func token(
        subjects: [SourceGraphPlacedSubject],
        bridges: [SourceGraphPlacedBridge]
    ) -> String {
        var parts: [String] = []
        parts.reserveCapacity(subjects.count + bridges.count)
        for placed in subjects {
            parts.append(cardToken(id: placed.id, gridX: placed.gridX, gridY: placed.gridY, observations: placed.observations))
        }
        for placed in bridges {
            parts.append(cardToken(id: placed.id, gridX: placed.gridX, gridY: placed.gridY, observations: placed.observations))
        }
        return parts.joined(separator: "|")
    }

    private static func cardToken(
        id: String,
        gridX: Int64,
        gridY: Int64,
        observations: [CatalogObservation]
    ) -> String {
        let ids = observations.map(\.id).joined(separator: ",")
        return "\(id)@\(gridX),\(gridY):\(ids)"
    }
}

/// The placement ghost: the live card, built in its ghost mode. The look
/// depends only on the armed kind, so SwiftUI never needs to update it.
private struct EvidenceGhostCard: NSViewRepresentable {
    let card: EvidenceCanvasCard

    func makeNSView(context: Context) -> EvidenceCardView { EvidenceCardView(ghostOf: card) }
    func updateNSView(_ view: EvidenceCardView, context: Context) {}
}
