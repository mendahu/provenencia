import SwiftUI

extension View {
    /// The Evidence graph form, presented while `model` is non-nil. The Source
    /// page sets `model` (Add or Edit) and adds this one line. A successful
    /// save clears `model`.
    func sourceGraphFormDialog(_ model: Binding<SourceGraphFormModel?>) -> some View {
        modifier(SourceGraphFormDialogModifier(model: model))
    }
}

private struct SourceGraphFormDialogModifier: ViewModifier {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Binding var model: SourceGraphFormModel?

    private var copy: PVFormDialogCopy {
        if let graph = model?.graph {
            return PVFormDialogCopy(
                title: L10n.Sources.graphFormEditTitle,
                verbatimSubtitle: graph.ref,
                confirm: L10n.Sources.graphFormSaveConfirm,
                cancel: L10n.Sources.cancelAction
            )
        }
        return PVFormDialogCopy(
            title: L10n.Sources.graphFormAddTitle,
            subtitle: L10n.Sources.graphFormAddSubtitle,
            confirm: L10n.Sources.graphFormAddConfirm,
            cancel: L10n.Sources.cancelAction
        )
    }

    func body(content: Content) -> some View {
        content.pvFormDialog(
            item: $model,
            copy: copy,
            width: 520,
            isRunning: model?.isSaving ?? false,
            isDirty: model?.isDirty ?? false,
            confirmDisabled: !(model?.canSave ?? false),
            accessibilityIdentifierPrefix: "sources.graphForm",
            onConfirm: { await $0.save() },
            form: { current in
                SourceGraphForm(
                    model: current,
                    onNavigate: { navigation.go(to: $0) },
                    dismiss: { model = nil }
                )
            }
        )
    }
}

/// The form body: save error, name, description, then the Artifacts the graph
/// links and, on an edit, the Citations that keep some of them linked.
private struct SourceGraphForm: View {
    @Bindable var model: SourceGraphFormModel
    let onNavigate: (WorkspaceLocation) -> Void
    let dismiss: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            if let saveError = model.saveError {
                PVCallout(
                    tone: .danger,
                    title: L10n.Sources.graphFormSaveErrorTitle,
                    message: saveError,
                    announces: true
                )
                .accessibilityIdentifier("sources.graphForm.error")
            }
            fields
                .disabled(model.isSaving)
            PVDivider()
            VStack(alignment: .leading, spacing: PVSpacing.space6) {
                SourceFormLoadArea(
                    state: model.loadState,
                    loadingText: L10n.Sources.graphFormLoading,
                    errorTitle: L10n.Sources.graphFormLoadErrorTitle,
                    errorMessage: L10n.Sources.graphFormLoadErrorMessage,
                    onReadAgain: { Task { await model.loadKeptLinks() } },
                    accessibilityIdentifierPrefix: "sources.graphForm"
                ) {
                    checklist
                }
                if model.loadState == .ready, !model.keptLinked.isEmpty {
                    keptLinked
                }
            }
        }
    }

    private var fields: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVField(
                label: L10n.Sources.graphFormName,
                error: model.nameError,
                marker: .required
            ) {
                PVInput(
                    text: $model.draft.name,
                    prompt: L10n.Sources.graphFormNamePlaceholder,
                    isInvalid: model.nameError != nil,
                    activateOnAppear: true
                )
                .accessibilityIdentifier("sources.graphForm.name")
                .onChange(of: model.draft.name) { _, _ in
                    if model.nameError != nil { model.nameError = nil }
                }
            }
            PVField(
                label: L10n.Sources.graphFormDescription,
                hint: L10n.Sources.graphFormDescriptionHint,
                marker: .optional
            ) {
                PVTextArea(text: $model.draft.description, lineLimit: 3...3)
                    .accessibilityIdentifier("sources.graphForm.description")
            }
        }
    }

    private var checklist: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            PVArtifactChecklist(
                artifacts: model.artifacts,
                excluded: $model.draft.excludedArtifactIDs,
                projectDir: model.projectDir,
                sourceTypeIconKey: model.sourceTypeIconKey,
                isDisabled: model.isSaving,
                locked: model.locked,
                description: L10n.Sources.graphFormArtifactsHint,
                accessibilityIdentifierPrefix: "sources.graphForm.artifacts"
            )
            if model.needsArtifact {
                Text(L10n.Sources.graphFormNoneChecked)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .accessibilityIdentifier("sources.graphForm.noneChecked")
            }
        }
    }

    private var keptLinked: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVCallout(
                tone: .neutral,
                icon: .lock,
                message: L10n.string(L10n.Sources.graphFormKeptLinkedMessage),
                compact: true
            )
            PVDeleteImpactGroupStack {
                ForEach(model.keptLinked, id: \.artifact.id) { kept in
                    PVDeleteImpactGroupList(
                        title: kept.artifact.displayTitle,
                        group: kept.block.citations,
                        onNavigate: onNavigate,
                        dismiss: dismiss
                    )
                }
            }
        }
        .accessibilityIdentifier("sources.graphForm.keptLinked")
    }
}
