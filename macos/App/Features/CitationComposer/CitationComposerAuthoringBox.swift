import SwiftUI

/// The single draft above the saved subject groups.
struct CitationComposerAuthoringBox: View {
    @Bindable var model: CitationComposerModel

    var body: some View {
        switch model.authoring {
        case .rest:
            rest
        case .property(let draft):
            propertyDraft(draft)
        case .bridge(let draft):
            bridgeDraft(draft)
        }
    }

    private var rest: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            HStack(spacing: PVSpacing.space3) {
                PVButton(L10n.CitationComposer.addObservation, variant: .secondary, size: .sm, icon: .plus) {
                    model.beginAddObservation()
                }
                .accessibilityIdentifier("citationComposer.addObservation")
                PVButton(L10n.CitationComposer.addBridge, variant: .secondary, size: .sm, icon: .link) {
                    model.beginAddBridge()
                }
                .accessibilityIdentifier("citationComposer.addBridge")
            }
            Text(L10n.CitationComposer.authoringRestHint)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
        }
    }

    private func propertyDraft(_ draft: CitationPropertyDraft) -> some View {
        PVCard(tone: .raised, border: .dashed, padding: PVSpacing.space5) {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                HStack {
                    Text(L10n.CitationComposer.newObservationTitle)
                        .font(PVFont.display(size: PVTypeScale.h4, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textDisplay)
                    PVBadge(L10n.CitationComposer.draftBadge, tone: .info)
                    Spacer(minLength: 0)
                }
                if let note = draft.note {
                    Text(verbatim: note)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
                PVComboBox(
                    selection: subjectBinding,
                    options: model.subjectOptions,
                    size: .sm,
                    placeholder: L10n.CitationComposer.subjectPlaceholder,
                    emptyLabel: L10n.CitationComposer.subjectPlaceholder,
                    label: L10n.CitationComposer.subjectLabel,
                    accessibilityIdentifierPrefix: "citationComposer.draft.subject",
                    focusNonce: draft.focus == .subject ? model.authoringFocusNonce : 0
                )
                PVComboBox(
                    selection: propertyBinding,
                    options: propertyOptions(subjectID: draft.values.subjectID),
                    size: .sm,
                    placeholder: L10n.CitationComposer.propertyPlaceholder,
                    emptyLabel: L10n.CitationComposer.propertyPlaceholder,
                    label: L10n.CitationComposer.propertyLabel,
                    accessibilityIdentifierPrefix: "citationComposer.draft.property",
                    focusNonce: draft.focus == .property ? model.authoringFocusNonce : 0
                )
                valueControls(values: draft.values, applyToEdit: false, focusValue: draft.focus == .value)
                PVCheckbox(
                    L10n.CitationComposer.negativeLabel,
                    description: L10n.CitationComposer.negativeDescription,
                    isChecked: negativeBinding
                )
                if let error = draft.error {
                    PVCallout(tone: .danger, message: error)
                }
                HStack {
                    PVButton(L10n.CitationComposer.discardDraft, variant: .ghost, size: .sm) {
                        model.discardAuthoring()
                    }
                    Spacer(minLength: 0)
                    PVButton(
                        L10n.CitationComposer.addObservation,
                        variant: .primary,
                        size: .sm,
                        loading: draft.isSaving
                    ) {
                        Task { await model.saveAuthoring() }
                    }
                    .disabled(draft.isSaving)
                    .accessibilityIdentifier("citationComposer.draft.saveObservation")
                }
            }
        }
        .accessibilityIdentifier("citationComposer.authoring.property")
    }

    private func bridgeDraft(_ draft: CitationBridgeDraft) -> some View {
        let pair = CitationComposerAuthoring.resolvePair(
            endA: draft.endA,
            endB: draft.endB,
            termID: draft.termID,
            vocabulary: model.vocabulary
        )
        return PVCard(tone: .raised, border: .dashed, padding: PVSpacing.space5) {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                HStack {
                    Text(L10n.CitationComposer.newBridgeTitle)
                        .font(PVFont.display(size: PVTypeScale.h4, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textDisplay)
                    PVBadge(L10n.CitationComposer.draftBadge, tone: .info)
                    Spacer(minLength: 0)
                }
                if let note = draft.note {
                    Text(verbatim: note)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
                PVComboBox(
                    selection: endABinding,
                    options: model.subjectOptions,
                    size: .sm,
                    placeholder: L10n.CitationComposer.subjectPlaceholder,
                    emptyLabel: L10n.CitationComposer.subjectPlaceholder,
                    label: endALabel(pair),
                    accessibilityIdentifierPrefix: "citationComposer.draft.endA",
                    focusNonce: draft.focus == .endA ? model.authoringFocusNonce : 0
                )
                PVComboBox(
                    selection: endBBinding,
                    options: model.subjectOptions,
                    size: .sm,
                    placeholder: L10n.CitationComposer.subjectPlaceholder,
                    emptyLabel: L10n.CitationComposer.subjectPlaceholder,
                    label: endBLabel(pair),
                    accessibilityIdentifierPrefix: "citationComposer.draft.endB",
                    focusNonce: draft.focus == .endB ? model.authoringFocusNonce : 0
                )
                bridgePair(pair, draft: draft)
                if let error = draft.error {
                    PVCallout(tone: .danger, message: error)
                }
                HStack {
                    PVButton(L10n.CitationComposer.discardDraft, variant: .ghost, size: .sm) {
                        model.discardAuthoring()
                    }
                    Spacer(minLength: 0)
                    if case .ready(let ready) = pair, ready.canSwap, !draft.endA.isEmpty, !draft.endB.isEmpty {
                        PVButton(L10n.CitationComposer.swapEnds, variant: .ghost, size: .sm, icon: .swap) {
                            model.swapBridgeEnds()
                        }
                    }
                    PVButton(
                        L10n.CitationComposer.saveBridge,
                        variant: .primary,
                        size: .sm,
                        loading: draft.isSaving
                    ) {
                        Task { await model.saveAuthoring() }
                    }
                    .disabled(draft.isSaving || !canSaveBridge(pair, termID: draft.termID))
                    .accessibilityIdentifier("citationComposer.draft.saveBridge")
                }
            }
        }
        .accessibilityIdentifier("citationComposer.authoring.bridge")
    }

    @ViewBuilder
    private func bridgePair(_ pair: CitationPair, draft: CitationBridgeDraft) -> some View {
        switch pair {
        case .incomplete(let hint):
            Text(verbatim: hint)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
        case .refused(let title, let message):
            PVCallout(tone: .warning, title: title, message: message)
        case .ready(let ready):
            if let property = ready.termProperty {
                HStack(spacing: PVSpacing.space3) {
                    PVComboBox(
                        selection: termBinding,
                        options: model.vocabulary.termComboOptionsByPropertyID[property.id] ?? [],
                        size: .sm,
                        placeholder: L10n.CitationComposer.termPlaceholder,
                        emptyLabel: L10n.CitationComposer.termEmpty,
                        label: ConnectionTermCopy.label(disambiguationKey: property.key),
                        accessibilityIdentifierPrefix: "citationComposer.draft.term",
                        focusNonce: draft.focus == .term ? model.authoringFocusNonce : 0
                    )
                    PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm) {
                        model.beginAddCustomTerm(propertyID: property.id)
                    }
                }
                Text(L10n.CitationComposer.termStartsEmpty)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            if let reading = ready.reading {
                Text(verbatim: reading)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textSecondary)
            }
        }
    }

    @ViewBuilder
    private func valueControls(values: ObservationRowValues, applyToEdit: Bool, focusValue: Bool) -> some View {
        let property = model.vocabulary.property(id: values.propertyID)
        switch property?.valueType {
        case PropertyValueType.text.rawValue:
            PVInput(text: textBinding, size: .sm)
                .accessibilityIdentifier("citationComposer.draft.text")
        case PropertyValueType.integer.rawValue:
            PVInput(text: integerBinding, size: .sm, mono: true)
        case PropertyValueType.term.rawValue:
            HStack(spacing: PVSpacing.space3) {
                PVComboBox(
                    selection: valueTermBinding,
                    options: model.vocabulary.termComboOptionsByPropertyID[values.propertyID] ?? [],
                    size: .sm,
                    placeholder: L10n.CitationComposer.termPlaceholder,
                    emptyLabel: L10n.CitationComposer.termEmpty,
                    label: L10n.CitationComposer.termLabel,
                    accessibilityIdentifierPrefix: "citationComposer.draft.valueTerm",
                    focusNonce: focusValue ? model.authoringFocusNonce : 0
                )
                PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm) {
                    model.beginAddCustomTerm(propertyID: values.propertyID, applyToEdit: applyToEdit)
                }
            }
        case PropertyValueType.name.rawValue, PropertyValueType.date.rawValue:
            Button {
                model.beginValueEditor(applyToEdit: applyToEdit)
            } label: {
                HStack {
                    let summary = CitationObservationValue.summary(
                        valueType: property?.valueType ?? "",
                        fields: CitationObservationValue.Fields(
                            valueText: values.valueText,
                            valueIntegerText: values.valueIntegerText,
                            valueTermID: values.valueTermID,
                            valueSubjectID: values.valueSubjectID,
                            dateDraft: values.dateDraft,
                            nameDraft: values.nameDraft
                        ),
                        termLabel: nil
                    )
                    if summary.isEmpty {
                        Text(L10n.CitationComposer.editValuePlaceholder)
                    } else {
                        Text(verbatim: summary)
                    }
                    Spacer(minLength: 0)
                    PVIcon(.penLine, size: 12)
                }
            }
            .buttonStyle(.pv(.secondary, size: .sm))
        default:
            Text(L10n.CitationComposer.choosePropertyFirst)
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
        }
    }

    private func propertyOptions(subjectID: String) -> [PVComboBoxOption] {
        model.vocabulary.propertyOptions(forSubjectID: subjectID).map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key)
        }
    }

    private func canSaveBridge(_ pair: CitationPair, termID: String) -> Bool {
        guard case .ready(let ready) = pair else { return false }
        if ready.termProperty != nil { return !termID.isEmpty }
        return true
    }

    private func endALabel(_ pair: CitationPair) -> String {
        if case .ready(let ready) = pair { return ready.endATitle }
        return L10n.string(L10n.CitationComposer.endA)
    }

    private func endBLabel(_ pair: CitationPair) -> String {
        if case .ready(let ready) = pair { return ready.endBTitle }
        return L10n.string(L10n.CitationComposer.endB)
    }

    private var subjectBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let draft) = model.authoring { return draft.values.subjectID }
                return ""
            },
            set: { model.selectAuthoringSubject($0, field: .propertySubject) }
        )
    }

    private var propertyBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let draft) = model.authoring { return draft.values.propertyID }
                return ""
            },
            set: { newValue in
                model.mutateProperty { draft in
                    draft.values.propertyID = newValue
                    draft.values.valueText = ""
                    draft.values.valueIntegerText = ""
                    draft.values.valueTermID = ""
                    draft.values.dateDraft = .empty()
                    draft.values.nameDraft = .empty()
                }
            }
        )
    }

    private var textBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let draft) = model.authoring { return draft.values.valueText }
                return ""
            },
            set: { newValue in model.mutateProperty { $0.values.valueText = newValue } }
        )
    }

    private var integerBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let draft) = model.authoring { return draft.values.valueIntegerText }
                return ""
            },
            set: { newValue in model.mutateProperty { $0.values.valueIntegerText = newValue } }
        )
    }

    private var valueTermBinding: Binding<String> {
        Binding(
            get: {
                if case .property(let draft) = model.authoring { return draft.values.valueTermID }
                return ""
            },
            set: { newValue in model.mutateProperty { $0.values.valueTermID = newValue } }
        )
    }

    private var negativeBinding: Binding<Bool> {
        Binding(
            get: {
                if case .property(let draft) = model.authoring {
                    return draft.values.polarity == ObservationPolarity.negative.rawValue
                }
                return false
            },
            set: { checked in
                model.mutateProperty {
                    $0.values.polarity = checked
                        ? ObservationPolarity.negative.rawValue
                        : ObservationPolarity.positive.rawValue
                }
            }
        )
    }

    private var endABinding: Binding<String> {
        Binding(
            get: {
                if case .bridge(let draft) = model.authoring { return draft.endA }
                return ""
            },
            set: { model.selectAuthoringSubject($0, field: .bridgeEndA) }
        )
    }

    private var endBBinding: Binding<String> {
        Binding(
            get: {
                if case .bridge(let draft) = model.authoring { return draft.endB }
                return ""
            },
            set: { model.selectAuthoringSubject($0, field: .bridgeEndB) }
        )
    }

    private var termBinding: Binding<String> {
        Binding(
            get: {
                if case .bridge(let draft) = model.authoring { return draft.termID }
                return ""
            },
            set: { newValue in model.mutateBridge { $0.termID = newValue } }
        )
    }
}
