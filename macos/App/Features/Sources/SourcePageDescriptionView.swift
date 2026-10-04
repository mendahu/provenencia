import SwiftUI

/// Description section: resting prose or explicit edit/save/cancel.
struct SourcePageDescriptionView: View {
    @Bindable var model: SourcePageModel
    /// Local draft so description keystrokes do not invalidate the Source page.
    @State private var descriptionDraft = ""

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space6) {
            PVSectionHeader(
                title: L10n.Sources.descriptionHeading,
                actions: {
                    if !model.identity.editingDescription {
                        PVButton(
                            L10n.Sources.editDescription,
                            variant: .ghost,
                            size: .sm,
                            icon: .penLine
                        ) {
                            beginEdit()
                        }
                        .accessibilityIdentifier("sources.page.description.edit")
                    }
                }
            )

            if model.identity.editingDescription
                || !model.identity.description.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            {
                PVInlineEdit(
                    isEditing: model.identity.editingDescription,
                    isSaving: model.identity.isSaving,
                    error: model.identity.descriptionError,
                    saveLabel: L10n.Sources.saveDescription,
                    cancelLabel: L10n.Sources.cancelEdit,
                    editLabel: L10n.Sources.editDescription,
                    showsEditControl: false,
                    axis: .vertical,
                    accessibilityIdentifierPrefix: "sources.page.description",
                    onEdit: { beginEdit() },
                    onSave: {
                        model.identity.descriptionDraft = descriptionDraft
                        Task { await model.identity.saveDescription() }
                    },
                    onCancel: {
                        model.identity.cancelEditDescription()
                        descriptionDraft = ""
                    }
                ) {
                    Text(model.identity.description)
                        .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.regular))
                        .foregroundStyle(PVColor.textSecondary)
                        .lineSpacing((PVLineHeight.normal - 1) * PVTypeScale.body)
                        .accessibilityIdentifier("sources.page.description")
                } editor: {
                    PVTextArea(
                        text: $descriptionDraft,
                        lineLimit: 3...12,
                        prompt: L10n.Sources.descriptionPlaceholder,
                        typography: .bodyLarge,
                        activateOnAppear: true
                    )
                    .accessibilityIdentifier("sources.page.description")
                    .disabled(model.identity.isSaving)
                }
            } else {
                PVEmptyState(
                    icon: .scrollText,
                    title: L10n.Sources.descriptionEmptyTitle,
                    message: L10n.string(L10n.Sources.descriptionEmptyMessage),
                    compact: true
                ) {
                    PVButton(
                        L10n.Sources.editDescription,
                        variant: .primary,
                        size: .sm,
                        icon: .penLine
                    ) {
                        beginEdit()
                    }
                    .accessibilityIdentifier("sources.page.description.empty.edit")
                }
                .accessibilityIdentifier("sources.page.description.empty")
            }
        }
    }

    private func beginEdit() {
        descriptionDraft = model.identity.description
        model.identity.beginEditDescription()
    }
}
