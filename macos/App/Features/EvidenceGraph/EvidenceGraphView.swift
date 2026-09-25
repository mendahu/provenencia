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

    private var fieldsKey: CatalogQueryKey {
        CatalogQueryKey.subjectFieldsWorkspace(project: session.projectKey)
    }

    private var sourcesListKey: CatalogQueryKey {
        CatalogQueryKey.sourcesList(project: session.projectKey)
    }

    private var connectRulesKey: CatalogQueryKey {
        CatalogQueryKey.connectRules(project: session.projectKey)
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
               let fieldsHandle: QueryHandle<SubjectFieldsSnapshot> = session.queryHandle(fieldsKey)
            {
                EvidenceGraphContent(
                    graphHandle: graphHandle,
                    fieldsHandle: fieldsHandle,
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
            let _: QueryHandle<SourceGraphRows> = session.query(graphKey)
            let _: QueryHandle<SubjectFieldsSnapshot> = session.query(fieldsKey)
            let _: QueryHandle<[CatalogSource]> = session.query(sourcesListKey)
            let _: QueryHandle<[CatalogConnectRule]> = session.query(connectRulesKey)
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
        let _: QueryHandle<[CatalogSource]> = session.query(sourcesListKey)
        let sources: [CatalogSource]? = await session.readyValue(sourcesListKey)
        if let match = sources?.first(where: { $0.id == sourceID }) {
            sourceTitle = match.title
            sourceRef = match.ref
        }
    }
}
// MARK: - Content

private struct EvidenceGraphContent: View {
    @Bindable var graphHandle: QueryHandle<SourceGraphRows>
    @Bindable var fieldsHandle: QueryHandle<SubjectFieldsSnapshot>
    @Bindable var model: EvidenceGraphModel
    let sourceID: String
    let sourceTitle: String
    let sourceRef: String?
    let contentSize: CGSize
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

    private var sheetPresented: Binding<Bool> {
        Binding(
            get: { model.isSheetPresented },
            set: { newValue in
                if !newValue { model.cancelSheet() }
            }
        )
    }

    private var disambiguationPresented: Binding<Bool> {
        Binding(
            get: { model.pendingDisambiguation != nil },
            set: { newValue in
                if !newValue { model.cancelDisambiguation() }
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
            isPresented: disambiguationPresented,
            copy: PVFormDialogCopy(
                title: model.disambiguationTitle(),
                subtitle: L10n.EvidenceGraph.connectDisambiguationSubtitleBare,
                confirm: L10n.EvidenceGraph.connectDisambiguationConfirm,
                cancel: L10n.EvidenceGraph.createCancel
            ),
            isRunning: false,
            confirmDisabled: !model.canConfirmDisambiguation,
            accessibilityIdentifierPrefix: "evidenceGraph.connect.disambiguation",
            onConfirm: {
                if let location = model.confirmDisambiguation() {
                    navigation.go(to: location)
                }
            }
        ) {
            EvidenceConnectDisambiguationForm(model: model)
        }
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
        .pvConfirm(
            item: Binding(
                get: { model.pendingDelete },
                set: { model.pendingDelete = $0 }
            ),
            copy: { pending in
                PVConfirmCopy(
                    title: String(localized: L10n.EvidenceGraph.deleteConfirmTitle),
                    message: L10n.EvidenceGraph.deleteConfirmMessage(
                        label: pending.label,
                        ref: pending.ref
                    ),
                    confirm: L10n.EvidenceGraph.deleteConfirm,
                    cancel: L10n.EvidenceGraph.deleteCancel
                )
            },
            isRunning: model.isDeleting,
            accessibilityIdentifierPrefix: "evidenceGraph.delete",
            onConfirm: {
                Task { _ = await model.confirmDeleteSubject() }
            }
        ) { _ in
            if let deleteError = model.deleteError {
                PVCallout(tone: .danger, message: deleteError)
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
                    graphHandle: graphHandle,
                    fieldsHandle: fieldsHandle,
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
        if let subjectID = effect.moveSubjectID {
            Task {
                _ = await model.moveSubject(
                    subjectID: subjectID,
                    fromGridX: effect.moveFromX,
                    fromGridY: effect.moveFromY,
                    deltaX: effect.moveDeltaX,
                    deltaY: effect.moveDeltaY
                )
            }
        }
        return effect.handled ? .handled : .ignored
    }
}

// MARK: - Document (paint-only; AppKit owns pointer)

private struct EvidenceGraphDocument: View {
    @Bindable var graphHandle: QueryHandle<SourceGraphRows>
    @Bindable var fieldsHandle: QueryHandle<SubjectFieldsSnapshot>
    @Bindable var model: EvidenceGraphModel
    let contentSize: CGSize
    let navigation: WorkspaceNavigation
    @Environment(\.graphCanvasPointer) private var pointer

    var body: some View {
        Group {
            if let pointer {
                EvidenceGraphDocumentBody(
                    graphHandle: graphHandle,
                    fieldsHandle: fieldsHandle,
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
    @Bindable var graphHandle: QueryHandle<SourceGraphRows>
    @Bindable var fieldsHandle: QueryHandle<SubjectFieldsSnapshot>
    @Bindable var model: EvidenceGraphModel
    let contentSize: CGSize
    let navigation: WorkspaceNavigation
    @Bindable var pointer: GraphCanvasPointerController

    private var snapshot: SourceGraphSnapshot {
        model.displaySnapshot(rows: graphHandle.value, types: fieldsHandle.value?.types ?? [])
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

    private var selectedBridgeID: String? {
        bridges.contains(where: { $0.id == model.selectedSubjectID })
            ? model.selectedSubjectID
            : nil
    }

    var body: some View {
        ZStack(alignment: .topLeading) {
            GraphCanvasGridView(contentSize: contentSize)
                .accessibilityHidden(true)

            EvidenceGraphEdgesHost(
                pointer: pointer,
                snapshot: snapshot,
                selectedBridgeID: selectedBridgeID
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

            if subjects.isEmpty, bridges.isEmpty, graphHandle.status == .ready, inputMode == .idle {
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
        .onChange(of: bridges.map { "\($0.id):\($0.gridX),\($0.gridY):\($0.observations.count)" }) { _, _ in
            publishHitTargets()
        }
    }

    private func wirePointer() {
        syncPointerMode()
        publishHitTargets()
        let model = model
        let navigation = navigation
        pointer.onSelect = { id in
            model.selectSubject(id: id)
        }
        pointer.onDeselect = {
            model.selectSubject(id: nil)
        }
        pointer.onDragEnded = { id, delta in
            guard let cell = model.gridCell(for: id) else { return }
            _ = model.commitDrag(
                subjectID: id,
                originGridX: cell.x,
                originGridY: cell.y,
                documentDelta: delta
            )
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
                    frame: EvidenceBridgeCard.contentFrame(for: placed, dragOffset: offset),
                    acceptsConnect: false,
                    actions: EvidenceBridgeCard.actionTargets(
                        for: placed,
                        canCite: model.canCite,
                        dragOffset: offset
                    )
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
            dragOffset: drag,
            hoveredActionID: pointer.hoveredCardAction?.cardID == placed.id
                ? pointer.hoveredCardAction?.actionID
                : nil
        )
        .accessibilityAction(named: Text(L10n.EvidenceGraph.editAccessibility)) {
            model.beginEdit(subjectID: placed.id)
        }
        .accessibilityAction(named: Text(L10n.EvidenceGraph.addProperty)) {
            if let location = model.composerLocation(for: placed.id) {
                navigation.go(to: location)
            }
        }
        .accessibilityAction(named: Text(L10n.EvidenceGraph.deleteAccessibility)) {
            model.beginDelete(subjectID: placed.id)
        }
        .accessibilityAction(named: Text(L10n.EvidenceGraph.editPropertyAccessibility)) {
            if let observation = placed.observations.first,
               let location = model.composerLocation(
                   forObservationID: observation.id,
                   subjectID: placed.id
               )
            {
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
            dragOffset: drag,
            hoveredActionID: pointer.hoveredCardAction?.cardID == placed.id
                ? pointer.hoveredCardAction?.actionID
                : nil,
            canCite: model.canCite
        )
        .accessibilityAction(named: Text(L10n.EvidenceGraph.editAccessibility)) {
            model.beginEdit(subjectID: placed.id)
        }
        .accessibilityAction(named: Text(L10n.EvidenceGraph.editCitationAccessibility)) {
            if let location = model.composerLocationForBridgeCitation(subjectID: placed.id) {
                navigation.go(to: location)
            }
        }
        .accessibilityAction(named: Text(L10n.EvidenceGraph.deleteAccessibility)) {
            model.beginDelete(subjectID: placed.id)
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
