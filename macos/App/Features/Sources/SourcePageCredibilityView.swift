import SwiftUI

/// Credibility assessment: grade chips, argument, save/cancel.
struct SourcePageCredibilityView: View {
    @Bindable var model: SourcePageModel
    /// Local argument so typing does not rewrite the page observation graph.
    @State private var argumentLocal = ""

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVSectionHeader(
                title: L10n.Sources.credibilityHeading,
                aside: {
                    Text(
                        model.credibility.hasSavedAssessment
                            ? L10n.Sources.credibilityHint
                            : L10n.Sources.credibilityHintUnset
                    )
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(2)
                }
            )

            PVChipGroup(style: .loose, spacing: PVSpacing.space4) {
                ForEach(model.credibility.grades, id: \.id) { grade in
                    PVChip(
                        text: grade.label,
                        isSelected: model.credibility.draftKey == grade.key,
                        tone: CatalogCredibility.chipTone(for: grade.key),
                        isDashed: CatalogCredibility.isDashedUnset(
                            key: grade.key,
                            hasSavedAssessment: model.credibility.hasSavedAssessment,
                            draftKey: model.credibility.draftKey
                        ),
                        action: { model.credibility.selectDraft(key: grade.key) }
                    )
                    .accessibilityIdentifier("sources.page.credibility.\(grade.key)")
                }
            }
            .accessibilityIdentifier("sources.page.credibility.grades")

            PVTextArea(
                text: $argumentLocal,
                lineLimit: 2...12,
                prompt: L10n.Sources.credibilityArgumentPlaceholder,
                accessibilityLabel: L10n.Sources.credibilityArgumentPlaceholder
            )
            .accessibilityIdentifier("sources.page.credibility.argument")

            HStack(spacing: PVSpacing.space4) {
                PVInlineEditActions(
                    isSaving: model.credibility.isSaving,
                    saveLabel: L10n.Sources.saveAssessment,
                    cancelLabel: L10n.Sources.cancelEdit,
                    saveDisabled: !model.credibility.isDirty(argument: argumentLocal),
                    showsCancel: model.credibility.isDirty(argument: argumentLocal),
                    accessibilityIdentifierPrefix: "sources.page.credibility",
                    onSave: {
                        model.credibility.argumentDraft = argumentLocal
                        Task { await model.credibility.save() }
                    },
                    onCancel: {
                        model.credibility.cancel()
                        argumentLocal = model.credibility.argumentDraft
                    }
                )

                if !model.credibility.isDirty(argument: argumentLocal) {
                    Text(
                        model.credibility.hasSavedAssessment
                            ? L10n.Sources.credibilitySavedStatus
                            : L10n.Sources.credibilityUnsavedStatus
                    )
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textFaint)
                }
            }
        }
        .onAppear { argumentLocal = model.credibility.argumentDraft }
        .onChange(of: model.credibility.argumentDraft) { _, newValue in
            // Sync after load / cancel / successful save (model owns the source of truth).
            argumentLocal = newValue
        }
    }
}
