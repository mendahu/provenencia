import SwiftUI

extension View {
    /// The Copy Source form, presented while `model` is non-nil. Both hosts
    /// (the Source page, the Sources list) set `model` and add this one line.
    /// A successful copy clears `model` and opens the copy's page.
    func sourceCopyDialog(_ model: Binding<SourceCopyModel?>) -> some View {
        modifier(SourceCopyDialogModifier(model: model))
    }
}

private struct SourceCopyDialogModifier: ViewModifier {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Binding var model: SourceCopyModel?

    func body(content: Content) -> some View {
        content.pvFormDialog(
            item: $model,
            copy: PVFormDialogCopy(
                title: L10n.Sources.copySourceDialogTitle,
                verbatimSubtitle: model.map { L10n.Sources.copySourceDialogSubtitle($0.original.title) },
                confirm: L10n.Sources.copySourceConfirm,
                cancel: L10n.Sources.cancelAction
            ),
            width: 520,
            isRunning: model?.isSaving ?? false,
            isDirty: model?.isDirty ?? false,
            confirmDisabled: !(model?.canCopy ?? false),
            accessibilityIdentifierPrefix: "sources.copy",
            onConfirm: { current in
                guard let copy = await current.copy() else { return false }
                navigation.go(to: SourcesListNavigation.pageLocation(for: copy))
                return true
            },
            form: { SourceCopyForm(model: $0) }
        )
    }
}

/// The form body: save error, shared title and type fields, then the parts
/// the researcher can bring over from the original.
private struct SourceCopyForm: View {
    @Bindable var model: SourceCopyModel

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if let copyError = model.copyError {
                PVCallout(
                    tone: .danger,
                    title: L10n.Sources.copySourceSaveErrorTitle,
                    message: L10n.Sources.copySourceSaveErrorMessage(copyError)
                )
                .accessibilityIdentifier("sources.copy.error")
            }
            SourceTitleTypeFields(
                draft: $model.choices.draft,
                titleError: $model.titleError,
                typeError: $model.typeError,
                types: model.types,
                accessibilityIdentifierPrefix: "sources.copy"
            )
            .disabled(model.isSaving)
            PVDivider()
            parts
        }
    }

    private var parts: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                Text(L10n.Sources.copySourcePartsHeading)
                    .pvMicroCaps()
                    .foregroundStyle(PVColor.textFaint)
                SourceFormLoadArea(
                    state: model.loadState,
                    loadingText: L10n.Sources.copySourceLoading,
                    errorTitle: L10n.Sources.copySourceLoadErrorTitle,
                    errorMessage: L10n.Sources.copySourceLoadErrorMessage,
                    onReadAgain: { model.reloadWorkspace() },
                    accessibilityIdentifierPrefix: "sources.copy"
                ) {
                    partCheckboxes
                }
            }
            if model.loadState == .ready, let workspace = model.workspace {
                PVArtifactChecklist(
                    artifacts: workspace.artifacts,
                    excluded: $model.choices.excludedArtifactIDs,
                    projectDir: model.projectDir,
                    sourceTypeIconKey: model.sourceTypeIconKey,
                    isDisabled: model.isSaving,
                    accessibilityIdentifierPrefix: "sources.copy.artifacts"
                )
            }
        }
    }

    private var partCheckboxes: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVCheckbox(
                L10n.Sources.copySourceDescription,
                description: model.descriptionAvailable
                    ? L10n.Sources.copySourceDescriptionHint
                    : L10n.Sources.copySourcePartAbsent,
                isChecked: partBinding(\.includeDescription, available: model.descriptionAvailable),
                isDisabled: !model.descriptionAvailable || model.isSaving
            )
            .accessibilityIdentifier("sources.copy.description")
            PVCheckbox(
                L10n.Sources.copySourceMetadata,
                description: model.metadataAvailable
                    ? L10n.Sources.copySourceMetadataHint
                    : L10n.Sources.copySourcePartAbsent,
                isChecked: partBinding(\.includeMetadata, available: model.metadataAvailable),
                isDisabled: !model.metadataAvailable || model.isSaving
            )
            .accessibilityIdentifier("sources.copy.metadata")
            PVCheckbox(
                L10n.Sources.copySourceCredibility,
                description: model.credibilityAvailable
                    ? L10n.Sources.copySourceCredibilityHint
                    : L10n.Sources.copySourcePartAbsent,
                isChecked: partBinding(\.includeCredibility, available: model.credibilityAvailable),
                isDisabled: !model.credibilityAvailable || model.isSaving
            )
            .accessibilityIdentifier("sources.copy.credibility")
        }
    }

    /// An unavailable part reads unchecked whatever the stored choice is.
    private func partBinding(
        _ keyPath: WritableKeyPath<SourceCopyChoices, Bool>,
        available: Bool
    ) -> Binding<Bool> {
        Binding(
            get: { available && model.choices[keyPath: keyPath] },
            set: { model.choices[keyPath: keyPath] = $0 }
        )
    }
}
