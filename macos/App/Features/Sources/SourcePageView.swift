import SwiftUI

/// Individual Source page (S2-18 / S2-25): sticky identity header, Description +
/// Credibility beside Metadata, Artifacts accordion, Notes. Section bodies live
/// in sibling `SourcePage*View` files; this shell owns load, dialogs, and layout.
struct SourcePageView: View {
    @State private var model: SourcePageModel
    let onBackToList: () -> Void
    /// Date dialog confirm wiring — kept off the page observation graph so
    /// structure/wording keystrokes stay local to the dialog form.
    @State private var dateEditorCanSave = false
    @State private var dateEditorSaveAction: (() async -> Void)?

    init(
        sourceID: String,
        projectDir: String,
        userID: String,
        sessionDisplayName: String = "",
        store: any GenealogyStore,
        onBackToList: @escaping () -> Void,
        onSourceUpdated: ((CatalogSource) -> Void)? = nil
    ) {
        _model = State(
            initialValue: SourcePageModel(
                sourceID: sourceID,
                projectDir: projectDir,
                userID: userID,
                sessionDisplayName: sessionDisplayName,
                store: store,
                onSourceUpdated: onSourceUpdated
            )
        )
        self.onBackToList = onBackToList
    }

    var body: some View {
        VStack(spacing: 0) {
            if model.workspace != nil || model.isLoading || model.loadError != nil {
                SourcePageIdentityHeader(model: model, onBackToList: onBackToList)
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
        .pvDialog(
            isPresented: addArtifactPresented,
            copy: PVDialogCopy(
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
        .pvDialog(
            isPresented: addFilePresented,
            copy: PVDialogCopy(
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
        .pvDialog(
            isPresented: addMetadataPresented,
            copy: PVDialogCopy(
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
        .pvDialog(
            isPresented: dateEditorPresented,
            copy: PVDialogCopy(
                title: model.metadata.isDateEditMode
                    ? L10n.Sources.editDateDialogTitle
                    : L10n.Sources.addDateDialogTitle,
                subtitle: model.metadata.dateEditorSubtitle,
                confirm: L10n.Sources.saveDateConfirm,
                cancel: L10n.Sources.cancelAction
            ),
            isRunning: model.metadata.isSavingDate,
            confirmDisabled: !dateEditorCanSave,
            accessibilityIdentifierPrefix: "sources.page.date",
            onConfirm: {
                Task { await dateEditorSaveAction?() }
            }
        ) {
            SourcePageMetadataView(model: model).dateEditorForm(
                canSave: $dateEditorCanSave,
                saveAction: $dateEditorSaveAction
            )
        }
        .onChange(of: model.metadata.isEditingDate) { _, open in
            if !open {
                dateEditorCanSave = false
                dateEditorSaveAction = nil
            }
        }
        .task { await model.load() }
        .accessibilityIdentifier("sources.page")
    }

    /// Board: Description + Credibility | Metadata.
    private var overviewColumns: some View {
        LazyVGrid(
            columns: [
                GridItem(.flexible(minimum: 400), spacing: PVSpacing.space11, alignment: .top),
                GridItem(.flexible(minimum: 480), spacing: PVSpacing.space11, alignment: .top),
            ],
            alignment: .leading,
            spacing: PVSpacing.space10
        ) {
            VStack(alignment: .leading, spacing: PVSpacing.space10) {
                SourcePageDescriptionView(model: model)
                SourcePageCredibilityView(model: model)
            }
            SourcePageMetadataView(model: model)
        }
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

    private var dateEditorPresented: Binding<Bool> {
        Binding(
            get: { model.metadata.isEditingDate },
            set: { presented in
                if !presented {
                    model.metadata.cancelDateEditor()
                }
            }
        )
    }

}
