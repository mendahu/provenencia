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
        PVCard(padding: PVSpacing.space5) {
            HStack(alignment: .center, spacing: PVSpacing.space4) {
                PVButton(L10n.CitationComposer.addObservation, variant: .primary, size: .sm, icon: .plus) {
                    model.beginAddObservation()
                }
                .accessibilityIdentifier("citationComposer.addObservation")
                PVButton(L10n.CitationComposer.addBridge, variant: .secondary, size: .sm, icon: .link) {
                    model.beginAddBridge()
                }
                .accessibilityIdentifier("citationComposer.addBridge")
                Text(L10n.CitationComposer.authoringRestHint)
                    .font(PVFont.body(size: 12, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
        }
    }

    private func propertyDraft(_ draft: CitationPropertyDraft) -> some View {
        draftCard(
            title: L10n.CitationComposer.newObservationTitle,
            note: draft.note,
            noteIcon: draft.note == nil ? nil : .duplicate
        ) {
            HStack(alignment: .top, spacing: PVSpacing.space5) {
                PVField(label: L10n.CitationComposer.subjectLabel) {
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
                }
                .frame(maxWidth: .infinity)
                PVField(label: L10n.CitationComposer.propertyLabel) {
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
                    .disabled(draft.values.subjectID.isEmpty)
                }
                .frame(maxWidth: .infinity)
            }
            valueControls(values: draft.values, applyToEdit: false, focusValue: draft.focus == .value)
            if let error = draft.error {
                PVCallout(tone: .danger, message: error)
            }
            HStack(spacing: PVSpacing.space4) {
                PVCheckbox(L10n.CitationComposer.negativeLabel, isChecked: negativeBinding)
                    .help(L10n.string(L10n.CitationComposer.negativeDescription))
                Spacer(minLength: 0)
                PVButton(L10n.CitationComposer.discardDraft, variant: .ghost, size: .sm) {
                    model.discardAuthoring()
                }
                PVButton(
                    L10n.CitationComposer.saveObservation,
                    variant: .primary,
                    size: .sm,
                    icon: .check,
                    loading: draft.isSaving
                ) {
                    Task { await model.saveAuthoring() }
                }
                .disabled(draft.isSaving || !propertyCanSave(draft.values))
                .accessibilityIdentifier("citationComposer.draft.saveObservation")
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
        let noteIcon: PVSymbol? = switch draft.origin {
        case .graph: .link
        case .duplicate: .duplicate
        case .blank: nil
        }
        return draftCard(title: L10n.CitationComposer.newBridgeTitle, note: draft.note, noteIcon: noteIcon) {
            endCombo(
                selection: endABinding,
                label: endALabel(pair),
                prefix: "citationComposer.draft.endA",
                focused: draft.focus == .endA
            )
            kindRow(pair, canSwap: canSwap(pair, draft: draft))
            endCombo(
                selection: endBBinding,
                label: endBLabel(pair),
                prefix: "citationComposer.draft.endB",
                focused: draft.focus == .endB
            )
            bridgeTail(pair, draft: draft)
            if let error = draft.error {
                PVCallout(tone: .danger, message: error)
            }
            HStack(spacing: PVSpacing.space4) {
                Spacer(minLength: 0)
                PVButton(L10n.CitationComposer.discardDraft, variant: .ghost, size: .sm) {
                    model.discardAuthoring()
                }
                PVButton(
                    L10n.CitationComposer.saveBridge,
                    variant: .primary,
                    size: .sm,
                    icon: .check,
                    loading: draft.isSaving
                ) {
                    Task { await model.saveAuthoring() }
                }
                .disabled(draft.isSaving || !canSaveBridge(pair, termID: draft.termID))
                .accessibilityIdentifier("citationComposer.draft.saveBridge")
            }
        }
        .accessibilityIdentifier("citationComposer.authoring.bridge")
    }

    private func draftCard<Content: View>(
        title: LocalizedStringResource,
        note: String?,
        noteIcon: PVSymbol?,
        @ViewBuilder content: () -> Content
    ) -> some View {
        PVCard(tone: .raised, border: .dashed, elevated: true, padding: PVSpacing.space6) {
            VStack(alignment: .leading, spacing: PVSpacing.space5) {
                HStack(spacing: PVSpacing.space4) {
                    Text(title)
                        .pvMicroCaps()
                        .foregroundStyle(PVColor.textSecondary)
                    PVBadge(L10n.CitationComposer.draftBadge, tone: .neutral, subtle: true)
                    Spacer(minLength: PVSpacing.space4)
                    Text(verbatim: CitationComposerAuthoring.headNote(
                        authoring: model.authoring,
                        vocabulary: model.vocabulary,
                        citationRef: model.activeCitationRef
                    ))
                    .font(PVFont.body(size: 12, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                }
                if let note {
                    HStack(alignment: .top, spacing: PVSpacing.space3) {
                        if let noteIcon {
                            PVIcon(noteIcon, size: 14)
                                .foregroundStyle(PVColor.textSecondary)
                                .padding(.top, PVSpacing.space1)
                        }
                        Text(verbatim: note)
                            .font(PVFont.body(size: 12, italic: true))
                            .foregroundStyle(PVColor.textSecondary)
                    }
                }
                content()
            }
        }
    }

    private func endCombo(
        selection: Binding<String>,
        label: String,
        prefix: String,
        focused: Bool
    ) -> some View {
        fieldLabel(label) {
            PVComboBox(
                selection: selection,
                options: model.subjectOptions,
                size: .sm,
                placeholder: L10n.CitationComposer.subjectPlaceholder,
                emptyLabel: L10n.CitationComposer.subjectPlaceholder,
                label: label,
                accessibilityIdentifierPrefix: prefix,
                focusNonce: focused ? model.authoringFocusNonce : 0
            )
        }
    }

    private func fieldLabel<Content: View>(_ label: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            Text(verbatim: label)
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textSecondary)
            content()
        }
    }

    @ViewBuilder
    private func kindRow(_ pair: CitationPair, canSwap: Bool) -> some View {
        HStack(spacing: PVSpacing.space4) {
            switch pair {
            case .incomplete(let hint):
                Text(verbatim: hint)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: .infinity, alignment: .leading)
            case .refused:
                Text(L10n.CitationComposer.noBridgeJoins)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .frame(maxWidth: .infinity, alignment: .leading)
            case .ready(let ready):
                PVMark(PVMarkKey(catalogKey: ready.markKey), size: 16, decorative: true)
                Text(verbatim: ready.kindLabel)
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(1)
                Text(verbatim: ready.pairHint)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(2)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
            if canSwap {
                PVIconButton(.sort, label: L10n.CitationComposer.swapEnds, size: .sm) {
                    model.swapBridgeEnds()
                }
            }
        }
        .frame(minHeight: PVSpacing.controlHeightSmall)
    }

    @ViewBuilder
    private func bridgeTail(_ pair: CitationPair, draft: CitationBridgeDraft) -> some View {
        switch pair {
        case .incomplete:
            EmptyView()
        case .refused(let title, let message):
            PVCallout(tone: .warning, title: title, message: message)
        case .ready(let ready):
            if let property = ready.termProperty {
                fieldLabel(ConnectionTermCopy.label(disambiguationKey: property.key)) {
                    VStack(alignment: .leading, spacing: PVSpacing.space2) {
                        HStack(spacing: PVSpacing.space4) {
                            PVComboBox(
                                selection: termBinding,
                                options: model.vocabulary.termComboOptionsByPropertyID[property.id] ?? [],
                                size: .sm,
                                placeholder: termPlaceholder(property),
                                emptyLabel: L10n.CitationComposer.termEmpty,
                                label: ConnectionTermCopy.label(disambiguationKey: property.key),
                                accessibilityIdentifierPrefix: "citationComposer.draft.term",
                                focusNonce: draft.focus == .term ? model.authoringFocusNonce : 0
                            )
                            PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm, icon: .plus) {
                                model.beginAddCustomTerm(propertyID: property.id)
                            }
                        }
                        if draft.termID.isEmpty {
                            Text(L10n.CitationComposer.termStartsEmpty)
                                .font(PVFont.body(size: PVTypeScale.micro, italic: true))
                                .foregroundStyle(PVColor.textMuted)
                        }
                    }
                }
            }
            if let reading = ready.reading {
                reads(reading)
            }
        }
    }

    private func reads(_ reading: String) -> some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            Text(L10n.CitationComposer.readsEyebrow)
                .pvMicroCaps()
                .foregroundStyle(PVColor.textMuted)
            Text(verbatim: reading)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textPrimary)
        }
        .padding(.horizontal, PVSpacing.space5)
        .padding(.vertical, PVSpacing.space3)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PVColor.surfaceSunken)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
    }

    private func canSwap(_ pair: CitationPair, draft: CitationBridgeDraft) -> Bool {
        guard case .ready(let ready) = pair else { return false }
        return ready.canSwap && !draft.endA.isEmpty && !draft.endB.isEmpty
    }

    private func termPlaceholder(_ property: CatalogProperty) -> LocalizedStringResource {
        property.key == "role" ? L10n.CitationComposer.searchRoles : L10n.CitationComposer.searchTypes
    }

    @ViewBuilder
    private func valueControls(values: ObservationRowValues, applyToEdit: Bool, focusValue: Bool) -> some View {
        let property = model.vocabulary.property(id: values.propertyID)
        let summary = valueSummary(values: values, property: property)
        PVField(label: L10n.CitationComposer.valueLabel, hint: valueHint(property)) {
            switch property?.valueType {
            case PropertyValueType.text.rawValue:
                PVInput(text: textBinding, size: .sm, prompt: L10n.CitationComposer.valueAsWritten)
                    .accessibilityIdentifier("citationComposer.draft.text")
            case PropertyValueType.integer.rawValue:
                PVInput(
                    text: integerBinding,
                    size: .sm,
                    mono: true,
                    suffix: property?.key == "age" ? L10n.string(L10n.CitationComposer.yearsSuffix) : nil
                )
            case PropertyValueType.term.rawValue:
                HStack(spacing: PVSpacing.space4) {
                    PVComboBox(
                        selection: valueTermBinding,
                        options: model.vocabulary.termComboOptionsByPropertyID[values.propertyID] ?? [],
                        size: .sm,
                        placeholder: L10n.CitationComposer.termPlaceholder,
                        emptyLabel: L10n.CitationComposer.termEmpty,
                        label: L10n.CitationComposer.valueLabel,
                        accessibilityIdentifierPrefix: "citationComposer.draft.valueTerm",
                        focusNonce: focusValue ? model.authoringFocusNonce : 0
                    )
                    PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm, icon: .plus) {
                        model.beginAddCustomTerm(propertyID: values.propertyID, applyToEdit: applyToEdit)
                    }
                }
            case PropertyValueType.name.rawValue:
                structuredValue(
                    summary: summary,
                    empty: summary.isEmpty,
                    mono: false,
                    title: summary.isEmpty ? L10n.CitationComposer.enterName : L10n.CitationComposer.editName,
                    icon: .signature,
                    applyToEdit: applyToEdit
                )
            case PropertyValueType.date.rawValue:
                structuredValue(
                    summary: summary,
                    empty: summary.isEmpty,
                    mono: true,
                    title: summary.isEmpty ? L10n.CitationComposer.enterDate : L10n.CitationComposer.editDateButton,
                    icon: .calendar,
                    applyToEdit: applyToEdit
                )
            default:
                PVInput(text: .constant(""), size: .sm, isReadOnly: true, prompt: L10n.CitationComposer.choosePropertyFirst)
                    .disabled(true)
            }
        }
    }

    private func structuredValue(
        summary: String,
        empty: Bool,
        mono: Bool,
        title: LocalizedStringResource,
        icon: PVSymbol,
        applyToEdit: Bool
    ) -> some View {
        HStack(spacing: PVSpacing.space5) {
            Text(verbatim: empty ? L10n.string(L10n.CitationComposer.notEnteredYet) : summary)
                .font(mono ? PVFont.mono(size: PVTypeScale.caption) : PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(empty ? PVColor.textFaint : PVColor.textPrimary)
                .lineLimit(1)
                .frame(maxWidth: .infinity, alignment: .leading)
            PVButton(title, variant: .secondary, size: .sm, icon: icon) {
                model.beginValueEditor(applyToEdit: applyToEdit)
            }
        }
        .frame(minHeight: PVSpacing.controlHeightSmall)
    }

    private func valueSummary(values: ObservationRowValues, property: CatalogProperty?) -> String {
        CitationObservationValue.summary(
            valueType: property?.valueType ?? "",
            fields: CitationObservationValue.Fields(
                valueText: values.valueText,
                valueIntegerText: values.valueIntegerText,
                valueTermID: values.valueTermID,
                valueSubjectID: values.valueSubjectID,
                dateDraft: values.dateDraft,
                nameDraft: values.nameDraft
            ),
            termLabel: model.vocabulary.termLabel(termID: values.valueTermID, propertyID: values.propertyID)
        )
    }

    private func valueHint(_ property: CatalogProperty?) -> LocalizedStringResource? {
        switch property?.valueType {
        case PropertyValueType.name.rawValue: L10n.CitationComposer.nameEditorHint
        case PropertyValueType.date.rawValue: L10n.CitationComposer.dateEditorHint
        default: nil
        }
    }

    private func propertyCanSave(_ values: ObservationRowValues) -> Bool {
        guard !values.subjectID.isEmpty, !values.propertyID.isEmpty else { return false }
        guard let property = model.vocabulary.property(id: values.propertyID) else { return false }
        return CitationObservationValue.isDialogValid(
            valueType: property.valueType,
            fields: CitationObservationValue.Fields(
                valueText: values.valueText,
                valueIntegerText: values.valueIntegerText,
                valueTermID: values.valueTermID,
                valueSubjectID: values.valueSubjectID,
                dateDraft: values.dateDraft,
                nameDraft: values.nameDraft
            )
        )
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
