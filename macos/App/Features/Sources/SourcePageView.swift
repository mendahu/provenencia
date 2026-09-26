import SwiftUI

/// Individual Source page (S2-18 / S2-25): sticky identity header, Description +
/// Credibility beside Metadata, Artifacts accordion, Notes. Section bodies live
/// in sibling `SourcePage*View` files; this shell owns cache sync, dialogs, and layout.
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
                    workspaceHandle: workspaceHandle,
                    sourceID: sourceID,
                    model: model
                )
            } else {
                SourcePageLoadingShell(model: model)
            }
        }
        .task(id: sourceID) {
            model.prepare(for: sourceID)
            let handle: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
            model.sync(from: handle, sourceID: sourceID)
        }
    }
}

/// Spinner-only shell before the workspace handle exists.
private struct SourcePageLoadingShell: View {
    let model: SourcePageModel

    var body: some View {
        VStack(spacing: 0) {
            if model.isLoading || model.loadError != nil {
                SourcePageIdentityHeader(model: model)
            }
            ScrollView {
                ProgressView()
                    .frame(maxWidth: .infinity, alignment: .center)
                    .padding(.top, PVSpacing.space9)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .background(PVColor.surfacePage)
        .accessibilityIdentifier("sources.page")
    }
}

/// Observes the workspace handle so cache loads repaint the page.
private struct SourcePageContent: View {
    @Bindable var workspaceHandle: QueryHandle<CatalogSourceWorkspace>
    let sourceID: String
    @Bindable var model: SourcePageModel
    /// Metadata column height so the left column can stretch and pin Credibility.
    @State private var metadataColumnHeight: CGFloat = 0

    var body: some View {
        VStack(spacing: 0) {
            if model.workspace != nil || model.isLoading || model.loadError != nil {
                SourcePageIdentityHeader(model: model)
            }
            ScrollView {
                VStack(alignment: .leading, spacing: PVSpacing.space9) {
                    if model.isLoading && model.workspace == nil {
                        ProgressView()
                            .frame(maxWidth: .infinity, alignment: .center)
                            .padding(.top, PVSpacing.space9)
                    } else if let loadError = model.loadError, model.workspace == nil {
                        PVCallout(tone: .danger, message: L10n.Errors.message(for: loadError))
                            .accessibilityIdentifier("sources.page.loadError")
                    } else {
                        if let pageError = model.pageError {
                            PVCallout(tone: .danger, message: pageError)
                                .accessibilityIdentifier("sources.page.error")
                        }
                        overviewColumns
                        SourcePageArtifactsView(model: model)
                        SourcePageNotesView(model: model)
                    }
                }
                .frame(maxWidth: PVSpacing.widthContentMax, alignment: .leading)
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.horizontal, PVSpacing.gutterPage)
                .padding(.top, PVSpacing.space8)
                .padding(.bottom, PVSpacing.space10)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .background(PVColor.surfacePage)
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
            isRunning: model.metadata.isSavingAdd,
            confirmDisabled: !model.metadata.canSubmitAdd,
            accessibilityIdentifierPrefix: "sources.page.addMetadata",
            onConfirm: { Task { await model.metadata.createFromAdd() } }
        ) {
            SourcePageMetadataView(model: model).addForm
        }
        .pvConfirm(
            item: Binding(
                get: { model.metadata.pendingDelete },
                set: { model.metadata.pendingDelete = $0 }
            ),
            copy: { item in
                PVConfirmCopy(
                    title: L10n.Sources.deleteMetadataConfirmTitle(label: item.label),
                    message: String(localized: L10n.Sources.deleteMetadataConfirmMessage),
                    confirm: L10n.Sources.deleteMetadataConfirm,
                    cancel: L10n.Sources.deleteMetadataKeep
                )
            },
            isRunning: model.metadata.isClearing,
            accessibilityIdentifierPrefix: "sources.page.metadata.delete",
            onConfirm: { Task { await model.metadata.confirmClear() } }
        ) { _ in
            EmptyView()
        }
        .onChange(of: workspaceHandle.status) { _, _ in
            model.sync(from: workspaceHandle, sourceID: sourceID)
        }
        .onChange(of: workspaceHandle.value) { _, newValue in
            guard newValue != nil else { return }
            model.sync(from: workspaceHandle, sourceID: sourceID)
        }
        .onAppear {
            model.sync(from: workspaceHandle, sourceID: sourceID)
        }
        .accessibilityIdentifier("sources.page")
    }

    /// Board: Description grows; Credibility pins to the bottom of the left
    /// column so it tracks the metadata column's height.
    private var overviewColumns: some View {
        HStack(alignment: .top, spacing: PVSpacing.space11) {
            VStack(alignment: .leading, spacing: PVSpacing.space10) {
                SourcePageDescriptionView(model: model)
                Spacer(minLength: 0)
                SourcePageCredibilityView(model: model)
            }
            .frame(minHeight: metadataColumnHeight, alignment: .top)
            .frame(maxWidth: .infinity, alignment: .leading)
            SourcePageMetadataView(model: model)
                .background {
                    GeometryReader { proxy in
                        Color.clear.preference(
                            key: SourcePageOverviewHeightKey.self,
                            value: proxy.size.height
                        )
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .onPreferenceChange(SourcePageOverviewHeightKey.self) { metadataColumnHeight = $0 }
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
            get: { model.metadata.isAdding },
            set: { presented in
                if presented {
                    model.metadata.openAdd()
                } else {
                    model.metadata.cancelAdd()
                }
            }
        )
    }

}

private struct SourcePageOverviewHeightKey: PreferenceKey {
    static var defaultValue: CGFloat = 0
    static func reduce(value: inout CGFloat, nextValue: () -> CGFloat) {
        value = max(value, nextValue())
    }
}
