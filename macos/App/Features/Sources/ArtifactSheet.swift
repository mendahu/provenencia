import SwiftUI

extension View {
    /// The Artifact sheet, presented while `model` is non-nil. The Source page
    /// sets `model` through `SourceArtifactsSection.openSheet`. A successful
    /// Save clears it; Cancel clears it at once; Escape and a click outside ask
    /// first when the sheet holds changes.
    func artifactSheet(_ model: Binding<ArtifactSheetModel?>) -> some View {
        modifier(ArtifactSheetModifier(model: model))
    }
}

private struct ArtifactSheetModifier: ViewModifier {
    @Binding var model: ArtifactSheetModel?

    private var copy: PVFormDialogCopy {
        switch model?.mode {
        case .create:
            PVFormDialogCopy(
                title: L10n.Sources.addArtifactDialogTitle,
                verbatimSubtitle: model?.subtitle,
                confirm: L10n.Sources.addArtifactConfirm,
                cancel: L10n.Sources.cancelAction
            )
        case .edit(let art):
            editCopy(ref: art.ref)
        case nil:
            editCopy(ref: "")
        }
    }

    private func editCopy(ref: String) -> PVFormDialogCopy {
        PVFormDialogCopy(
            verbatimTitle: L10n.Sources.editArtifactLabel(ref: ref),
            verbatimSubtitle: model?.subtitle,
            confirm: L10n.Sources.saveArtifact,
            cancel: L10n.Sources.cancelAction
        )
    }

    func body(content: Content) -> some View {
        content.pvFormDialog(
            item: $model,
            copy: copy,
            width: 560,
            isRunning: model?.isSaving ?? false,
            isDirty: model?.isDirty ?? false,
            confirmDisabled: !(model?.canSave ?? false),
            accessibilityIdentifierPrefix: "sources.page.artifactSheet",
            scrollsBody: true,
            scrollTarget: model?.scrollTarget.map { AnyHashable($0) },
            scrollRequest: model.map { AnyHashable($0.scrollRequest) },
            onConfirm: { await $0.save() },
            form: { current in ArtifactSheetForm(model: current) }
        )
    }
}

/// The sheet's body: save error, label and description, the Evidence graphs the
/// Artifact is in, and its copy details. Each section carries its `.id` for the
/// scroll target.
private struct ArtifactSheetForm: View {
    @Bindable var model: ArtifactSheetModel

    private let prefix = "sources.page.artifactSheet"

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            VStack(alignment: .leading, spacing: PVSpacing.space6) {
                if let saveError = model.saveError {
                    PVCallout(
                        tone: .danger,
                        title: L10n.Sources.artifactCardSaveErrorTitle,
                        message: saveError,
                        announces: true
                    )
                    .accessibilityIdentifier("\(prefix).saveError")
                }
                fields
                    .disabled(model.isSaving)
            }
            .id(ArtifactSheetModel.Section.top)
            graphLinks
                .id(ArtifactSheetModel.Section.graphs)
            copyDetails
                .id(ArtifactSheetModel.Section.copyDetails)
        }
    }

    private var fields: some View {
        PVTabPair { label, description in
            VStack(alignment: .leading, spacing: PVSpacing.space6) {
                PVField(
                    label: L10n.Sources.artifactLabel,
                    error: model.labelError,
                    marker: .required
                ) {
                    PVInput(
                        text: $model.draft.label,
                        isInvalid: model.labelError != nil,
                        focus: label,
                        activateOnAppear: model.opensAt == .top
                    )
                    .accessibilityIdentifier("\(prefix).label")
                    .onChange(of: model.draft.label) { _, _ in
                        if model.labelError != nil { model.labelError = nil }
                    }
                }
                PVField(
                    label: L10n.Sources.artifactDescription,
                    hint: L10n.Sources.artifactDescriptionHint
                ) {
                    PVTextArea(text: $model.draft.description, lineLimit: 3...8, focus: description)
                        .accessibilityIdentifier("\(prefix).description")
                }
            }
        }
    }

    @ViewBuilder
    private var graphLinks: some View {
        Group {
            if model.awaitingGraphs {
                PVField(
                    label: L10n.Sources.artifactGraphsLabel,
                    hint: L10n.Sources.artifactGraphsHint
                ) {
                    SourceFormLoadArea(
                        state: .loading,
                        loadingText: L10n.Sources.artifactGraphsLoading,
                        errorTitle: L10n.Sources.artifactGraphsLoadErrorTitle,
                        errorMessage: L10n.Sources.artifactGraphsLoadErrorMessage,
                        onReadAgain: {},
                        compact: true,
                        accessibilityIdentifierPrefix: "\(prefix).graphs"
                    ) {
                        EmptyView()
                    }
                }
            } else if model.graphIDs != nil {
                PVField(
                    label: L10n.Sources.artifactGraphsLabel,
                    hint: L10n.Sources.artifactGraphsHint
                ) {
                    SourceFormLoadArea(
                        state: model.linkLoadState,
                        loadingText: L10n.Sources.artifactGraphsLoading,
                        errorTitle: L10n.Sources.artifactGraphsLoadErrorTitle,
                        errorMessage: L10n.Sources.artifactGraphsLoadErrorMessage,
                        onReadAgain: { Task { await model.loadLinkLocks() } },
                        compact: true,
                        accessibilityIdentifierPrefix: "\(prefix).graphs"
                    ) {
                        SourceArtifactGraphLinks(
                            graphs: model.graphs,
                            excluded: $model.excludedGraphIDs,
                            locked: model.lockedGraphs,
                            isDisabled: model.isSaving,
                            accessibilityIdentifierPrefix: "\(prefix).graphs"
                        )
                    }
                    if model.inNoGraph {
                        SourceArtifactNoGraphLine(accessibilityIdentifier: "\(prefix).graphs.none")
                    }
                }
            }
        }
        .onChange(of: model.visibleGraphs?.map(\.id)) { _, _ in
            model.graphsChanged()
        }
    }

    private var copyDetails: some View {
        PVField(label: L10n.Sources.artifactCopyDetailsLabel) {
            MetadataBoardView(
                board: model.board,
                placement: .form,
                accessibilityIdentifierPrefix: "\(prefix).copyDetails",
                onClear: { id in
                    // A draft board clears with no confirm.
                    Task { _ = await model.board.requestClear(fieldID: id) }
                }
            )
            .disabled(model.isSaving)
        }
    }
}
