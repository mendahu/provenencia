import SwiftUI

/// Individual Source page (S2-18 / S2-25): pinned identity header, Description +
/// Credibility beside Metadata (Metadata in a side column at Wide), graphs, artifact cards, Notes,
/// on Page columns. Section bodies live in sibling `SourcePage*View` files; this shell owns cache
/// sync, dialogs, and layout.
struct SourcePageView: View {
    @Environment(WorkspaceSession.self) private var session
    let sourceID: String
    @State private var model: SourcePageModel

    init(
        sourceID: String,
        session: WorkspaceSession,
        userID: String,
        sessionDisplayName: String = "",
        store: any GenealogyStore
    ) {
        self.sourceID = sourceID
        _model = State(
            initialValue: SourcePageModel(
                sourceID: sourceID,
                session: session,
                userID: userID,
                sessionDisplayName: sessionDisplayName,
                store: store
            )
        )
    }

    private var workspaceKey: CatalogQueryKey {
        CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: sourceID)
    }

    var body: some View {
        Group {
            if let workspaceHandle: QueryHandle<CatalogSourceWorkspace> = session.queryHandle(workspaceKey) {
                SourcePageContent(
                    sourceID: sourceID,
                    workspaceHandle: workspaceHandle,
                    model: model
                )
            } else {
                SourcePageLoadingShell(sourceID: sourceID, model: model)
            }
        }
        .task {
            let handle: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
            model.sync(from: handle)
        }
    }
}

/// Spinner-only shell before the workspace handle exists.
private struct SourcePageLoadingShell: View {
    let sourceID: String
    let model: SourcePageModel

    var body: some View {
        PVPageColumns(
            headerPinned: true,
            scrollResetKey: sourceID,
            header: { SourcePageIdentityHeader(model: model) },
            primary: {
                ProgressView()
                    .frame(maxWidth: .infinity, alignment: .center)
                    .padding(.top, PVSpacing.space9)
            }
        )
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .background(PVColor.surfacePage)
        .accessibilityIdentifier("sources.page")
    }
}

/// Observes the workspace handle so cache loads repaint the page.
private struct SourcePageContent: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    let sourceID: String
    @Bindable var workspaceHandle: QueryHandle<CatalogSourceWorkspace>
    @Bindable var model: SourcePageModel
    /// True while a file drag is over the content pane.
    @State private var pageFileDrag = false

    var body: some View {
        PVPageColumns(
            headerPinned: true,
            // A new page error scrolls up to the banner it shows in.
            scrollResetKey: "\(sourceID)/\(model.pageErrorShownCount)",
            primaryFill: true,
            fillSplitSideMax: SourcePageOverviewLayout.sideMax,
            header: { SourcePageIdentityHeader(model: model) },
            primary: {
                if model.isLoading && model.workspace == nil {
                    ProgressView()
                        .frame(maxWidth: .infinity, alignment: .center)
                        .padding(.top, PVSpacing.space9)
                } else if let loadError = model.loadError, model.workspace == nil {
                    PVCallout(tone: .danger, message: L10n.Errors.message(for: loadError))
                        .accessibilityIdentifier("sources.page.loadError")
                } else {
                    SourcePageOverview(model: model)
                }
            }
        )
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .background(PVColor.surfacePage)
        .overlay {
            if showPageDropHighlight {
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .fill(PVColor.accentSoft.opacity(0.55))
                    .overlay(
                        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                            .strokeBorder(PVColor.accent, lineWidth: 2)
                    )
                    .allowsHitTesting(false)
            }
        }
        .overlay(alignment: .bottom) {
            if showPageDropHighlight {
                pageDropHint
            }
        }
        .dropDestination(for: URL.self, action: { urls, _ in
            guard model.acceptsArtifactPageDrop, !urls.isEmpty else { return false }
            model.artifacts.stagePageDrop(urls)
            return true
        }, isTargeted: { targeted in
            pageFileDrag = targeted
        })
        .vocabularyToastOverlay($model.toast, identifier: "sources.page.toast")
        .pvFormDialog(
            isPresented: addArtifactPresented,
            copy: PVFormDialogCopy(
                title: L10n.Sources.addArtifactDialogTitle,
                subtitle: L10n.Sources.addArtifactDialogSubtitle,
                confirm: L10n.Sources.addArtifactConfirm,
                cancel: L10n.Sources.cancelAction
            ),
            isRunning: model.artifacts.isSavingDraft,
            isDirty: model.artifacts.addSheetIsDirty,
            confirmDisabled: !model.artifacts.canSubmitDraft,
            accessibilityIdentifierPrefix: "sources.page.addArtifact",
            onConfirm: { Task { await model.artifacts.create() } }
        ) {
            SourcePageArtifactsView(model: model).addForm
        }
        .pvFormDialog(
            isPresented: addFilePresented,
            copy: PVFormDialogCopy(
                title: L10n.Sources.addFileDialogTitle,
                subtitle: L10n.Sources.addFileDialogSubtitle,
                confirm: L10n.Sources.addFileConfirm,
                cancel: L10n.Sources.cancelAction
            ),
            isRunning: model.artifacts.isSavingAttach,
            isDirty: model.artifacts.attachSheetIsDirty,
            confirmDisabled: !model.artifacts.canSubmitAttach,
            accessibilityIdentifierPrefix: "sources.page.addFile",
            onConfirm: { Task { await model.artifacts.confirmAttach() } }
        ) {
            SourcePageArtifactsView(model: model).addFileForm
        }
        .pvFormDialog(
            isPresented: addMetadataPresented,
            copy: PVFormDialogCopy(
                title: L10n.Sources.addMetadataDialogTitle,
                subtitle: L10n.Sources.addMetadataDialogSubtitle,
                confirm: L10n.Sources.addMetadataConfirm,
                cancel: L10n.Sources.cancelAction
            ),
            isRunning: model.metadataDialogs.isSavingAdd,
            isDirty: model.metadataDialogs.addSheetIsDirty,
            confirmDisabled: !model.metadataDialogs.canSubmitAdd,
            accessibilityIdentifierPrefix: "sources.page.addMetadata",
            onConfirm: { Task { await model.metadataDialogs.createFromAdd() } }
        ) {
            MetadataBoardAddForm(dialogs: model.metadataDialogs, accessibilityIdentifierPrefix: "sources.page.addMetadata")
        }
        .sourceCopyDialog($model.copying)
        .sourceGraphFormDialog($model.graphs.form)
        .sourceDateDialog($model.identity.dateForm)
        .pvConfirm(
            item: Binding(
                get: { model.metadataDialogs.pendingDelete },
                set: { model.metadataDialogs.pendingDelete = $0 }
            ),
            copy: { item in
                PVConfirmCopy(
                    title: L10n.Sources.deleteMetadataConfirmTitle(label: item.label),
                    message: L10n.string(L10n.Sources.deleteMetadataConfirmMessage),
                    confirm: L10n.Sources.deleteMetadataConfirm,
                    cancel: L10n.Sources.deleteMetadataKeep
                )
            },
            isRunning: model.metadataDialogs.isClearing,
            accessibilityIdentifierPrefix: "sources.page.metadata.delete",
            onConfirm: { Task { await model.metadataDialogs.confirmClear() } }
        ) { _ in
            EmptyView()
        }
        .pvConfirm(
            item: Binding(
                get: { model.artifacts.pendingEditSwitch },
                set: { if $0 == nil { model.artifacts.cancelEditSwitch() } }
            ),
            copy: { _ in .discardChanges },
            tone: .danger,
            accessibilityIdentifierPrefix: "sources.page.artifact.discard",
            onConfirm: { model.artifacts.confirmEditSwitch() }
        ) { _ in
            EmptyView()
        }
        .pvConfirm(
            item: Binding(
                get: { model.notes.pendingDelete },
                set: { model.notes.pendingDelete = $0 }
            ),
            copy: { _ in
                PVConfirmCopy(
                    title: L10n.string(L10n.Sources.deleteNoteConfirmTitle),
                    message: L10n.string(L10n.Sources.deleteNoteConfirmMessage),
                    confirm: L10n.Sources.deleteNoteConfirm,
                    cancel: L10n.Sources.deleteNoteKeep
                )
            },
            isRunning: model.notes.isSaving,
            accessibilityIdentifierPrefix: "sources.page.note.delete",
            onConfirm: { Task { await model.notes.confirmDelete() } }
        ) { _ in
            EmptyView()
        }
        .pvDeleteImpact(
            flow: model.deleteImpact,
            accessibilityIdentifierPrefix: "sources.page.deleteImpact",
            onConfirm: {
                Task { await model.confirmResourceDelete(leaving: navigation.fallbackToSectionRoot) }
            },
            onNavigate: { location in
                navigation.go(to: location)
            }
        )
        .onChange(of: workspaceHandle.status) {
            model.sync(from: workspaceHandle)
        }
        .leavesForSectionRoot(
            when: workspaceHandle.read.isFailed || model.gone != nil,
            notice: model.goneNotice(read: workspaceHandle.read)
        )
        .onChange(of: workspaceHandle.value) { _, newValue in
            guard newValue != nil else { return }
            model.sync(from: workspaceHandle)
        }
        .onAppear {
            model.sync(from: workspaceHandle)
        }
        .accessibilityIdentifier("sources.page")
    }

    private var showPageDropHighlight: Bool {
        pageFileDrag && model.acceptsArtifactPageDrop
    }

    private var pageDropHint: some View {
        HStack(spacing: PVSpacing.space4) {
            PVIcon(.fileUp, size: 18)
                .foregroundStyle(PVColor.accent)
            Text(L10n.Sources.pageDropHint)
                .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                .foregroundStyle(PVColor.textPrimary)
        }
        .padding(.horizontal, PVSpacing.space7)
        .padding(.vertical, PVSpacing.space5)
        .background(PVColor.surfaceCard)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .strokeBorder(PVColor.accent, lineWidth: 1)
        )
        .pvShadow(PVElevation.lg)
        .padding(.bottom, PVSpacing.space12)
        .allowsHitTesting(false)
        .accessibilityIdentifier("sources.page.dropHint")
    }

    private var addArtifactPresented: Binding<Bool> {
        Binding(
            get: { model.artifacts.isAdding },
            set: { presented in
                if presented {
                    model.artifacts.openAdd()
                } else {
                    model.artifacts.cancelAdd()
                }
            }
        )
    }

    private var addFilePresented: Binding<Bool> {
        Binding(
            get: { model.artifacts.isAttaching },
            set: { presented in
                if !presented {
                    model.artifacts.cancelAttach()
                }
            }
        )
    }

    private var addMetadataPresented: Binding<Bool> {
        Binding(
            get: { model.metadataDialogs.isAdding },
            set: { presented in
                if presented {
                    model.metadataDialogs.openAdd(on: model.metadata)
                } else {
                    model.metadataDialogs.cancelAdd()
                }
            }
        )
    }

}

/// The page's sections as four fixed subviews for `SourcePageOverviewLayout`: the page error,
/// Description above Credibility, Metadata, then graphs, Artifacts and Notes. No section
/// changes parent between widths, so its drafts survive crossing the Wide breakpoint.
private struct SourcePageOverview: View {
    @Environment(\.pvWidthClass) private var widthClass
    @Bindable var model: SourcePageModel

    var body: some View {
        SourcePageOverviewLayout(widthClass: widthClass) {
            VStack(spacing: 0) {
                if let pageError = model.pageError {
                    PVCallout(
                        tone: .danger,
                        title: pageError.title,
                        message: pageError.message,
                        announces: true
                    )
                    .accessibilityIdentifier("sources.page.error")
                }
            }
            VStack(alignment: .leading, spacing: PVSpacing.space10) {
                SourcePageDescriptionView(model: model)
                Spacer(minLength: 0)
                SourcePageCredibilityView(model: model)
            }
            SourcePageMetadataView(model: model)
            VStack(alignment: .leading, spacing: PVSpacing.space9) {
                SourcePageGraphsView(model: model)
                SourcePageArtifactsView(model: model)
                SourcePageNotesView(model: model)
            }
        }
    }
}

/// Places the page's four subviews (error, Description and Credibility, Metadata, rest).
///
/// Regular: a group of up to `widthContentMax`, left-aligned; the error across it, then
/// Description and Credibility beside Metadata, each half the group, both as tall as the
/// taller (Credibility sits at the bottom of its column), then the rest below.
/// Wide: the group `PVPageColumnsMetrics.split` computes: the error, Description and
/// Credibility, and the rest stacked in a main column, Metadata in a side column beside it.
/// An empty error adds no gap. It sizes itself from the proposed width alone, so it holds no
/// state and crossing the breakpoint keeps the page's scroll and focus.
struct SourcePageOverviewLayout: Layout {
    let widthClass: PVWidthClass

    private static let rowGap = PVSpacing.space9
    /// The side column's cap; `PVPageColumns` is handed the same value to size its band.
    static let sideMax = PVSpacing.widthSecondaryMax

    private func width(_ proposal: ProposedViewSize) -> CGFloat {
        proposal.width.flatMap { $0.isFinite ? $0 : nil } ?? PVSpacing.widthContentMax
    }

    /// The `Subviews` adapter for `arrange`: no frames unless there are four subviews.
    private func placed(width: CGFloat, subviews: Subviews) -> (frames: [CGRect], height: CGFloat) {
        guard subviews.count == 4 else { return ([], 0) }
        return Self.arrange(width: width, widthClass: widthClass, fit: subviews.fit)
    }

    /// The frame of each of the four subviews (error, body, side, rest) and the overall height,
    /// for `width`. `fit` is a subview's size at a proposed width.
    static func arrange(
        width: CGFloat, widthClass: PVWidthClass, fit: (Int, CGFloat) -> CGSize
    ) -> (frames: [CGRect], height: CGFloat) {
        var frames = [CGRect](repeating: .zero, count: 4)
        if widthClass == .wide {
            let split = PVPageColumnsMetrics.split(available: width, sideMax: Self.sideMax)
            let x = PVPageColumnsMetrics.groupX(in: width, group: split.group)
            let error = fit(0, split.main), body = fit(1, split.main), rest = fit(3, split.main)
            var y: CGFloat = 0
            frames[0] = CGRect(x: x, y: y, width: split.main, height: error.height)
            if error.height > 0 { y += error.height + Self.rowGap }
            frames[1] = CGRect(x: x, y: y, width: split.main, height: body.height)
            y += body.height + Self.rowGap
            frames[3] = CGRect(x: x, y: y, width: split.main, height: rest.height)
            y += rest.height
            let side = fit(2, split.side)
            frames[2] = CGRect(
                x: x + split.group - split.side, y: 0, width: split.side, height: side.height
            )
            return (frames, max(y, side.height))
        }
        let group = min(PVSpacing.widthContentMax, width)
        let half = max(0, (group - PVSpacing.space11) / 2)
        let error = fit(0, group)
        var y: CGFloat = 0
        frames[0] = CGRect(x: 0, y: y, width: group, height: error.height)
        if error.height > 0 { y += error.height + Self.rowGap }
        let left = fit(1, half), right = fit(2, half)
        let row = max(left.height, right.height)
        frames[1] = CGRect(x: 0, y: y, width: half, height: row)
        frames[2] = CGRect(x: half + PVSpacing.space11, y: y, width: half, height: row)
        y += row + Self.rowGap
        let rest = fit(3, group)
        frames[3] = CGRect(x: 0, y: y, width: group, height: rest.height)
        return (frames, y + rest.height)
    }

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let w = width(proposal)
        return CGSize(width: w, height: placed(width: w, subviews: subviews).height)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let laidOut = placed(width: bounds.width, subviews: subviews)
        for (index, frame) in laidOut.frames.enumerated() {
            subviews[index].place(
                at: CGPoint(x: bounds.minX + frame.minX, y: bounds.minY + frame.minY),
                anchor: .topLeading,
                proposal: ProposedViewSize(width: frame.width, height: frame.height)
            )
        }
    }
}
