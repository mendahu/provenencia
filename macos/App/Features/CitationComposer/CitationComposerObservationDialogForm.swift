import SwiftUI

/// Observation dialog body (Frames 3–6, 13) for `.pvFormDialog`.
struct CitationComposerObservationDialogForm: View {
    @Bindable var model: CitationComposerModel
    var onAddCustomTerm: () -> Void
    var onEditName: () -> Void

    var body: some View {
        if let draft = model.observationDialog {
            VStack(alignment: .leading, spacing: PVSpacing.space6) {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    PVComboBox(
                        selection: propertyBinding,
                        options: model.propertyOptions,
                        size: .sm,
                        placeholder: L10n.CitationComposer.propertyPlaceholder,
                        emptyLabel: L10n.CitationComposer.propertyEmpty,
                        label: L10n.CitationComposer.propertyLabel,
                        accessibilityIdentifierPrefix: "citationComposer.dialog.property"
                    )
                    if let error = model.dialogPropertyError {
                        Text(verbatim: error)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.danger)
                    }
                }

                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    Text(L10n.CitationComposer.polarityLabel)
                        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textSecondary)
                    PVChipGroup(style: .segmented) {
                        PVChip(
                            L10n.CitationComposer.polarityAsserts,
                            isSelected: draft.polarity != "negative",
                            expands: true,
                            selectionLift: true,
                            action: {
                                var next = draft
                                next.polarity = "positive"
                                model.updateObservationDialog(next)
                            }
                        )
                        PVChip(
                            L10n.CitationComposer.polarityNegates,
                            isSelected: draft.polarity == "negative",
                            expands: true,
                            selectionLift: true,
                            action: {
                                var next = draft
                                next.polarity = "negative"
                                model.updateObservationDialog(next)
                            }
                        )
                    }
                }

                valueEditor(draft)
            }
        }
    }

    @ViewBuilder
    private func valueEditor(_ draft: CitationComposerModel.ObservationDialogState) -> some View {
        let property = model.catalogProperty(id: draft.propertyID)
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            Text(valueTypeLabel(property?.valueType))
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textSecondary)

            if let property {
                switch property.valueType {
                case "text":
                    PVTextArea(text: textBinding, lineLimit: 1...4)
                        .accessibilityIdentifier("citationComposer.dialog.valueText")
                case "integer":
                    PVInput(text: integerBinding, size: .sm)
                        .accessibilityIdentifier("citationComposer.dialog.valueInteger")
                case "term":
                    VStack(alignment: .leading, spacing: PVSpacing.space3) {
                        PVComboBox(
                            selection: termBinding,
                            options: model.termOptions(for: property.id),
                            size: .sm,
                            placeholder: L10n.CitationComposer.termPlaceholder,
                            emptyLabel: L10n.CitationComposer.termEmpty,
                            label: L10n.CitationComposer.termLabel,
                            accessibilityIdentifierPrefix: "citationComposer.dialog.term"
                        )
                        PVButton(
                            L10n.CitationComposer.addCustomTerm,
                            variant: .ghost,
                            size: .sm,
                            action: onAddCustomTerm
                        )
                    }
                case "date":
                    DateValueEditorForm(
                        draft: dateBinding,
                        accessibilityIdentifierPrefix: "citationComposer.dialog.date"
                    )
                case "name":
                    NameValueHostControl(
                        draft: draft.nameDraft,
                        accessibilityIdentifierPrefix: "citationComposer.dialog.name",
                        onEdit: onEditName
                    )
                default:
                    Text(L10n.CitationComposer.unsupportedValueTypeError)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
            } else {
                Text(L10n.CitationComposer.dialogValuePickPropertyFirst)
                    .font(PVFont.body(size: PVTypeScale.body))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(PVSpacing.space5)
                    .background(
                        RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                            .fill(PVColor.surfaceSunken)
                    )
            }

            if let error = model.dialogValueError {
                Text(verbatim: error)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.danger)
            }
        }
    }

    private func valueTypeLabel(_ type: String?) -> LocalizedStringResource {
        switch type {
        case "text": return L10n.CitationComposer.valueLabelText
        case "integer": return L10n.CitationComposer.valueLabelInteger
        case "term": return L10n.CitationComposer.valueLabelTerm
        case "date": return L10n.CitationComposer.valueLabelDate
        case "name": return L10n.CitationComposer.valueLabelName
        default: return L10n.CitationComposer.valueLabel
        }
    }

    private var propertyBinding: Binding<String> {
        Binding(
            get: { model.observationDialog?.propertyID ?? "" },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.propertyID = $0
                model.updateObservationDialog(draft)
            }
        )
    }

    private var textBinding: Binding<String> {
        Binding(
            get: { model.observationDialog?.valueText ?? "" },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.valueText = $0
                model.updateObservationDialog(draft)
            }
        )
    }

    private var integerBinding: Binding<String> {
        Binding(
            get: { model.observationDialog?.valueIntegerText ?? "" },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.valueIntegerText = $0
                model.updateObservationDialog(draft)
            }
        )
    }

    private var termBinding: Binding<String> {
        Binding(
            get: { model.observationDialog?.valueTermID ?? "" },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.valueTermID = $0
                model.updateObservationDialog(draft)
            }
        )
    }

    private var dateBinding: Binding<DateValueDraft> {
        Binding(
            get: { model.observationDialog?.dateDraft ?? .empty() },
            set: {
                guard var draft = model.observationDialog else { return }
                draft.dateDraft = $0
                model.updateObservationDialog(draft)
            }
        )
    }
}
