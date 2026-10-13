import SwiftUI

/// Observation dialog body. The form dialog owns Discard and Save.
struct CitationComposerAuthoringBox: View {
    @Bindable var model: CitationComposerModel
    @FocusState private var focus: CitationAuthoringFocus?
    /// The dialog's first field, placed once. Putting it in `.defaultFocus`
    /// reapplies it whenever a value editor or name part is inserted, which
    /// sends Tab back to Subject.
    @State private var didPlaceOpeningFocus = false

    var body: some View {
        dialog
            .onAppear { placeOpeningFocusIfNeeded() }
    }

    /// Once, when the sheet appears. A Property / Bridge switch and a newly
    /// inserted value editor must leave focus where the user just put it.
    private func placeOpeningFocusIfNeeded() {
        guard !didPlaceOpeningFocus else { return }
        didPlaceOpeningFocus = true
        let initial = liveOpeningFocus
        DispatchQueue.main.async { focus = initial }
    }

    @ViewBuilder
    private var dialog: some View {
        switch model.authoring {
        case .rest:
            EmptyView()
        case .property(let draft):
            form(origin: draft.note, originIcon: draft.isDuplicate ? .duplicate : nil) {
                propertyFields(draft)
            }
        case .bridge(let draft):
            form(origin: draft.note, originIcon: originIcon(draft.origin)) {
                bridgeFields(draft)
            }
        }
    }

    private var liveOpeningFocus: CitationAuthoringFocus {
        switch model.authoring {
        case .rest:
            .subject
        case .property(let draft):
            CitationAuthoringNavigation.initialPropertyFocus(subjectID: draft.values.subjectID)
        case .bridge(let draft):
            CitationAuthoringNavigation.initialBridgeFocus(
                endA: draft.endA,
                endB: draft.endB,
                hasTermField: bridgeHasTerm(draft),
                termID: draft.termID
            )
        }
    }

    private func form<Content: View>(
        origin: String?,
        originIcon: PVSymbol?,
        @ViewBuilder content: () -> Content
    ) -> some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            CitationComposerModalLead(
                model: model,
                isProperty: { if case .property = model.authoring { return true }; return false }(),
                lockKind: false,
                citationSelection: citationBinding,
                onProperty: { model.showPropertyAuthoring() },
                onBridge: { model.showBridgeAuthoring() }
            )
            if let origin, !origin.isEmpty {
                HStack(alignment: .top, spacing: PVSpacing.space3) {
                    if let originIcon {
                        PVIcon(originIcon, size: 14)
                            .foregroundStyle(PVColor.textSecondary)
                            .padding(.top, PVSpacing.space1)
                    }
                    Text(verbatim: origin)
                        .font(PVFont.body(size: 12, italic: true))
                        .foregroundStyle(PVColor.textSecondary)
                        .fixedSize(horizontal: false, vertical: true)
                }
            }
            content()
        }
    }

    private func originIcon(_ origin: CitationBridgeOrigin) -> PVSymbol? {
        switch origin {
        case .graph: .link
        case .duplicate: .duplicate
        case .blank: nil
        }
    }

    private func propertyFields(_ draft: CitationPropertyDraft) -> some View {
        CitationPropertyFields(
            model: model,
            values: propertyValuesBinding,
            subjectField: .propertySubject,
            error: draft.error,
            identifierPrefix: "citationComposer.draft",
            focus: $focus,
            applyTermToEdit: false
        )
        .onChange(of: draft.isSaving) { wasSaving, saving in
            guard wasSaving, !saving, draft.error == nil else { return }
            DispatchQueue.main.async { focus = .property }
        }
    }

    private func bridgeFields(_ draft: CitationBridgeDraft) -> some View {
        let pair = CitationComposerAuthoring.resolvePair(
            endA: draft.endA,
            endB: draft.endB,
            termID: draft.termID,
            vocabulary: model.vocabulary
        )
        return VStack(alignment: .leading, spacing: 10) {
            HStack(alignment: .top, spacing: PVSpacing.space5) {
                endCombo(
                    selection: endABinding,
                    label: L10n.CitationComposer.endA,
                    prefix: "citationComposer.draft.endA",
                    field: .endA,
                    draft: draft
                )
                endCombo(
                    selection: endBBinding,
                    label: L10n.CitationComposer.endB,
                    prefix: "citationComposer.draft.endB",
                    field: .endB,
                    draft: draft
                )
            }
            bridgeStatus(pair, draft: draft)
            if let error = draft.error {
                PVCallout(tone: .danger, message: error)
            }
        }
        .accessibilityIdentifier("citationComposer.authoring.bridge")
        .onChange(of: draft.isSaving) { wasSaving, saving in
            guard wasSaving, !saving, draft.error == nil else { return }
            DispatchQueue.main.async { focus = .endB }
        }
    }

    private func bridgeHasTerm(_ draft: CitationBridgeDraft) -> Bool {
        bridgeHasTerm(endA: draft.endA, endB: draft.endB, termID: draft.termID)
    }

    private func bridgeHasTerm(endA: String, endB: String, termID: String) -> Bool {
        let pair = CitationComposerAuthoring.resolvePair(
            endA: endA,
            endB: endB,
            termID: termID,
            vocabulary: model.vocabulary
        )
        if case .ready(let ready) = pair, ready.termProperty != nil { return true }
        return false
    }

    private func endCombo(
        selection: Binding<String>,
        label: LocalizedStringResource,
        prefix: String,
        field: CitationAuthoringFocus,
        draft: CitationBridgeDraft
    ) -> some View {
        let endA = draft.endA
        let endB = draft.endB
        let termID = draft.termID
        return PVField(label: label) {
            PVComboBox(
                selection: selection,
                options: model.subjectOptions,
                size: .sm,
                placeholder: L10n.CitationComposer.subjectPlaceholder,
                emptyLabel: L10n.CitationComposer.subjectPlaceholder,
                label: label,
                accessibilityIdentifierPrefix: prefix,
                focus: PVComboFocus($focus, equals: field),
                onCommit: { value, via in
                    let hasTerm = field == .endB
                        ? self.bridgeHasTerm(endA: endA, endB: value, termID: termID)
                        : self.bridgeHasTerm(endA: value, endB: endB, termID: termID)
                    return self.claimTab(from: field, committed: value, via: via, bridgeHasTerm: hasTerm)
                }
            )
        }
        .frame(maxWidth: .infinity)
    }

    @ViewBuilder
    private func bridgeStatus(_ pair: CitationPair, draft: CitationBridgeDraft) -> some View {
        switch pair {
        case .incomplete(let hint):
            Text(verbatim: hint)
                .font(PVFont.body(size: 12, italic: true))
                .foregroundStyle(PVColor.textMuted)
                .fixedSize(horizontal: false, vertical: true)
        case .refused(let title, let message):
            Text(L10n.CitationComposer.noBridgeJoins)
                .font(PVFont.body(size: 12, italic: true))
                .foregroundStyle(PVColor.textMuted)
            PVCallout(tone: .warning, title: title, message: message)
        case .ready(let ready):
            CitationBridgeSentence(
                ready: ready,
                subjectA: model.vocabulary.subjectsByID[draft.endA],
                subjectB: model.vocabulary.subjectsByID[draft.endB],
                terms: model.vocabulary.termsByPropertyID[ready.termProperty?.id ?? ""] ?? [],
                termID: termBinding,
                canSwap: canSwap(pair, draft: draft),
                onSwap: { model.swapBridgeEnds() },
                onAddTerm: {
                    if let property = ready.termProperty {
                        model.beginAddCustomTerm(propertyID: property.id)
                    }
                },
                termFocus: PVComboFocus($focus, equals: .term)
            )
        }
    }


    private func canSwap(_ pair: CitationPair, draft: CitationBridgeDraft) -> Bool {
        guard case .ready(let ready) = pair else { return false }
        return ready.canSwap && !draft.endA.isEmpty && !draft.endB.isEmpty
    }

    private var propertyValuesBinding: Binding<ObservationRowValues> {
        Binding(
            get: {
                if case .property(let draft) = model.authoring { return draft.values }
                return ObservationRowValues()
            },
            set: { newValue in
                model.mutateProperty { $0.values = newValue }
            }
        )
    }

    private var citationBinding: Binding<String> {
        Binding(
            get: { model.authoringCitationID },
            set: { model.setAuthoringCitation($0) }
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

    private func claimTab(
        from field: CitationAuthoringFocus,
        committed: String,
        via: PVComboCommitVia,
        bridgeHasTerm: Bool
    ) -> Bool {
        guard let next = CitationAuthoringNavigation.tabDestination(
            from: field,
            committed: committed,
            via: via,
            draftAfterCommit: CitationAuthoringDraftAfterCommit(bridgeHasTerm: bridgeHasTerm)
        ) else { return false }
        // The next field may be inserted by this same commit. A name or date
        // editor is not in the hierarchy until the following turn, and a deep
        // form can miss the first write.
        let nextField = next
        DispatchQueue.main.async {
            focus = nextField
            DispatchQueue.main.async {
                if focus != nextField { focus = nextField }
            }
        }
        return true
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

/// Property observation fields shared by the new-observation dialog and the edit dialog.
struct CitationPropertyFields: View {
    @Bindable var model: CitationComposerModel
    @Binding var values: ObservationRowValues
    var subjectField: CitationSubjectField
    var error: String?
    var identifierPrefix: String
    var focus: FocusState<CitationAuthoringFocus?>.Binding
    var applyTermToEdit: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(alignment: .top, spacing: PVSpacing.space5) {
                PVField(label: L10n.CitationComposer.subjectLabel) {
                    PVComboBox(
                        selection: subjectBinding,
                        options: model.subjectOptions,
                        size: .sm,
                        placeholder: L10n.CitationComposer.subjectPlaceholder,
                        emptyLabel: L10n.CitationComposer.subjectPlaceholder,
                        label: L10n.CitationComposer.subjectLabel,
                        accessibilityIdentifierPrefix: "\(identifierPrefix).subject",
                        focus: PVComboFocus(focus, equals: .subject),
                        onCommit: { value, via in
                            self.claimTab(from: .subject, committed: value, via: via)
                        }
                    )
                }
                .frame(maxWidth: .infinity)
                PVField(label: L10n.CitationComposer.propertyLabel) {
                    PVComboBox(
                        selection: propertyBinding,
                        options: propertyOptions,
                        size: .sm,
                        placeholder: L10n.CitationComposer.propertyPlaceholder,
                        emptyLabel: values.subjectID.isEmpty
                            ? L10n.CitationComposer.chooseSubjectFirst
                            : L10n.CitationComposer.propertyPlaceholder,
                        label: L10n.CitationComposer.propertyLabel,
                        accessibilityIdentifierPrefix: "\(identifierPrefix).property",
                        focus: PVComboFocus(focus, equals: .property),
                        onCommit: { value, via in
                            self.claimTab(from: .property, committed: value, via: via)
                        }
                    )
                }
                .frame(maxWidth: .infinity)
            }
            valueControls
            PVCheckbox(
                L10n.CitationComposer.negativeLabel,
                description: L10n.CitationComposer.negativeDescription,
                isChecked: negativeBinding
            )
            if let error {
                PVCallout(tone: .danger, message: error)
            }
        }
        .accessibilityIdentifier("\(identifierPrefix).propertyForm")
    }

    @ViewBuilder
    private var valueControls: some View {
        let property = model.vocabulary.property(id: values.propertyID)
        PVField(label: L10n.CitationComposer.valueLabel) {
            switch property?.valueType {
            case PropertyValueType.text.rawValue:
                PVInput(
                    text: textBinding,
                    size: .sm,
                    prompt: L10n.CitationComposer.valueAsWritten,
                    focus: PVInputFocus(focus, equals: .value)
                )
                .accessibilityIdentifier("\(identifierPrefix).text")
            case PropertyValueType.integer.rawValue:
                PVInput(
                    text: integerBinding,
                    size: .sm,
                    mono: true,
                    suffix: property?.key == "age" ? L10n.string(L10n.CitationComposer.yearsSuffix) : nil,
                    focus: PVInputFocus(focus, equals: .value)
                )
            case PropertyValueType.term.rawValue:
                HStack(spacing: PVSpacing.space4) {
                    PVComboBox(
                        selection: termBinding,
                        options: model.vocabulary.termComboOptionsByPropertyID[values.propertyID] ?? [],
                        size: .sm,
                        placeholder: L10n.CitationComposer.termPlaceholder,
                        emptyLabel: L10n.CitationComposer.termEmpty,
                        label: L10n.CitationComposer.valueLabel,
                        accessibilityIdentifierPrefix: "\(identifierPrefix).valueTerm",
                        focus: PVComboFocus(focus, equals: .value)
                    )
                    PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm, icon: .plus) {
                        model.beginAddCustomTerm(
                            propertyID: values.propertyID,
                            applyToEdit: applyTermToEdit
                        )
                    }
                }
            case PropertyValueType.name.rawValue:
                inlineValue {
                    NameValueEditorForm(
                        draft: nameBinding,
                        accessibilityIdentifierPrefix: "\(identifierPrefix).name",
                        formFocus: PVInputFocus(focus, equals: .value)
                    )
                }
            case PropertyValueType.date.rawValue:
                inlineValue {
                    DateValueEditorForm(
                        draft: dateBinding,
                        accessibilityIdentifierPrefix: "\(identifierPrefix).date",
                        yearFocus: PVInputFocus(focus, equals: .value)
                    )
                }
            default:
                PVInput(
                    text: .constant(""),
                    size: .sm,
                    isReadOnly: true,
                    prompt: L10n.CitationComposer.choosePropertyFirst
                )
                .disabled(true)
            }
        }
    }

    private func inlineValue<Content: View>(@ViewBuilder content: () -> Content) -> some View {
        content()
            .padding(.horizontal, 14)
            .padding(.vertical, PVSpacing.space5)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(PVColor.surfaceSunken)
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .stroke(PVColor.borderSubtle, lineWidth: 1)
            )
    }

    private func claimTab(
        from field: CitationAuthoringFocus,
        committed: String,
        via: PVComboCommitVia
    ) -> Bool {
        guard let next = CitationAuthoringNavigation.tabDestination(
            from: field,
            committed: committed,
            via: via,
            draftAfterCommit: CitationAuthoringDraftAfterCommit()
        ) else { return false }
        // The next field may be inserted by this same commit. A name or date
        // editor is not in the hierarchy until the following turn, and a deep
        // form can miss the first write.
        let binding = focus
        let nextField = next
        DispatchQueue.main.async {
            binding.wrappedValue = nextField
            DispatchQueue.main.async {
                if binding.wrappedValue != nextField { binding.wrappedValue = nextField }
            }
        }
        return true
    }

    private var propertyOptions: [PVComboBoxOption] {
        model.vocabulary.propertyOptions(forSubjectID: values.subjectID).map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key)
        }
    }

    private var subjectBinding: Binding<String> {
        Binding(
            get: { values.subjectID },
            set: { raw in
                var next = values
                model.applyObservationSubject(raw, field: subjectField, values: &next)
                values = next
            }
        )
    }

    private var propertyBinding: Binding<String> {
        Binding(
            get: { values.propertyID },
            set: { raw in
                var next = values
                model.applyObservationProperty(raw, values: &next)
                values = next
            }
        )
    }

    private var textBinding: Binding<String> {
        Binding(get: { values.valueText }, set: { values.valueText = $0 })
    }

    private var integerBinding: Binding<String> {
        Binding(get: { values.valueIntegerText }, set: { values.valueIntegerText = $0 })
    }

    private var termBinding: Binding<String> {
        Binding(get: { values.valueTermID }, set: { values.valueTermID = $0 })
    }

    private var dateBinding: Binding<DateValueDraft> {
        Binding(get: { values.dateDraft }, set: { values.dateDraft = $0 })
    }

    private var nameBinding: Binding<NameValueDraft> {
        Binding(get: { values.nameDraft }, set: { values.nameDraft = $0 })
    }

    private var negativeBinding: Binding<Bool> {
        Binding(
            get: { values.polarity == ObservationPolarity.negative.rawValue },
            set: { checked in
                values.polarity = checked
                    ? ObservationPolarity.negative.rawValue
                    : ObservationPolarity.positive.rawValue
            }
        )
    }
}

/// Citation picker and Property | Bridge chips at the top of the observation dialog.
struct CitationComposerModalLead: View {
    @Bindable var model: CitationComposerModel
    var isProperty: Bool
    var lockKind: Bool
    var citationSelection: Binding<String>
    var citationEnabled: Bool = true
    var onProperty: () -> Void
    var onBridge: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVField(label: L10n.CitationComposer.citationEyebrow) {
                VStack(alignment: .leading, spacing: PVSpacing.space3) {
                    PVSelect(
                        selection: citationSelection,
                        options: citationOptions,
                        size: .sm,
                        displayLabel: model.citationStepLabel(for: citationSelection.wrappedValue),
                        menuWidth: 460,
                        maxVisibleRows: 5,
                        rowHeight: 72,
                        isDisabled: !citationEnabled,
                        accessibilityLabel: L10n.CitationComposer.citationEyebrow,
                        accessibilityIdentifier: "citationComposer.modal.citation"
                    ) { option in
                        if option.id.isEmpty {
                            HStack(spacing: PVSpacing.space4) {
                                PVIcon(.plus, size: 14)
                                Text(L10n.CitationComposer.newCitation)
                                    .font(PVFont.body(size: PVTypeScale.caption))
                                Text(L10n.CitationComposer.onThisArtifact)
                                    .font(PVFont.body(size: 12, italic: true))
                                    .foregroundStyle(PVColor.textMuted)
                            }
                        } else if let listed = model.orderedCitations.first(where: { $0.citation.id == option.id }) {
                            CitationStepMenuRow(listed: listed, locator: menuLocator(listed.citation))
                        } else {
                            Text(verbatim: option.label)
                        }
                    }
                    citationReading
                }
            }
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                PVChipGroup(style: .segmented) {
                    PVChip(
                        L10n.CitationComposer.kindProperty,
                        isSelected: isProperty,
                        expands: true,
                        selectionLift: true,
                        action: { if !lockKind { onProperty() } }
                    )
                    .disabled(lockKind && !isProperty)
                    .accessibilityIdentifier("citationComposer.kind.property")
                    PVChip(
                        L10n.CitationComposer.kindBridge,
                        isSelected: !isProperty,
                        expands: true,
                        selectionLift: true,
                        action: { if !lockKind { onBridge() } }
                    )
                    .disabled(lockKind && isProperty)
                    .accessibilityIdentifier("citationComposer.kind.bridge")
                }
                .accessibilityLabel(Text(L10n.CitationComposer.observationKind))
                if lockKind {
                    Text(L10n.CitationComposer.kindLocked)
                        .font(PVFont.body(size: 12, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            PVDivider(axis: .horizontal, color: PVColor.borderSubtle)
        }
    }

    /// The open citation's transcription, under the menu, so the reading is
    /// in view while the observation is filled in.
    private var citationReading: some View {
        let text = model.transcription.trimmingCharacters(in: .whitespacesAndNewlines)
        return Group {
            if text.isEmpty {
                Text(L10n.CitationComposer.noTranscriptionLine)
                    .font(PVFont.body(size: 13, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            } else {
                Text(verbatim: text)
                    .font(PVFont.body(size: 13))
                    .foregroundStyle(PVColor.textSecondary)
            }
        }
        .lineLimit(8)
        .multilineTextAlignment(.leading)
        .fixedSize(horizontal: false, vertical: true)
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, PVSpacing.space5)
        .padding(.vertical, PVSpacing.space4)
        .background(PVColor.surfaceSunken)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .stroke(PVColor.borderDefault, lineWidth: 1)
        )
        .accessibilityIdentifier("citationComposer.modal.citationReading")
    }

    private var citationOptions: [PVSelectOption] {
        [PVSelectOption(value: "", label: L10n.string(L10n.CitationComposer.newCitation))]
            + model.orderedCitations.map {
                PVSelectOption(value: $0.citation.id, label: $0.citation.ref)
            }
    }

    private func menuLocator(_ citation: CatalogCitation) -> String {
        let locator = CitationLocatorDraft.decode(citation.locatorJSON)
        var parts: [String] = []
        if let page = locator.page {
            parts.append(L10n.CitationComposer.menuPage(page))
        }
        if locator.region != nil {
            parts.append(L10n.string(L10n.CitationComposer.menuRegion))
        }
        return parts.joined(separator: " · ")
    }
}

/// The bridge sentence, with the type sitting in the line.
struct CitationBridgeSentence: View {
    let ready: CitationReadyBridge
    let subjectA: CitationComposerModel.GraphSubjectOption?
    let subjectB: CitationComposerModel.GraphSubjectOption?
    let terms: [CatalogPropertyTerm]
    @Binding var termID: String
    var canSwap: Bool
    var onSwap: () -> Void
    var onAddTerm: () -> Void
    var termFocus: PVComboFocus? = nil
    var termIdentifierPrefix: String = "citationComposer.draft.term"

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            HStack(spacing: PVSpacing.space3) {
                PVMark(PVMarkKey(catalogKey: ready.markKey), size: 14, decorative: true)
                Text(verbatim: ready.kindLabel)
                    .pvMicroCaps()
                    .foregroundStyle(PVColor.textSecondary)
                Spacer(minLength: 0)
                if canSwap {
                    PVButton(L10n.CitationComposer.swap, variant: .ghost, size: .sm, icon: .swap) {
                        onSwap()
                    }
                }
            }
            FlowLayout(spacing: PVSpacing.space4, lineSpacing: PVSpacing.space3) {
                ForEach(pieces) { piece in
                    pieceView(piece)
                }
            }
            if ready.termProperty != nil {
                HStack {
                    Spacer(minLength: 0)
                    PVButton(L10n.CitationComposer.addCustomTerm, variant: .ghost, size: .sm, icon: .plus) {
                        onAddTerm()
                    }
                }
            }
        }
        .padding(.horizontal, 14)
        .padding(.vertical, PVSpacing.space5)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PVColor.surfaceSunken)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .stroke(PVColor.borderSubtle, lineWidth: 1)
        )
        .accessibilityElement(children: .contain)
    }

    @ViewBuilder
    private func pieceView(_ piece: Piece) -> some View {
        switch piece {
        case .word(let text, let strong, _):
            Text(verbatim: text)
                .font(PVFont.body(size: 15, weight: strong ? PVFontWeight.semibold : PVFontWeight.regular))
                .foregroundStyle(PVColor.textPrimary)
        case .term:
            PVComboBox(
                selection: $termID,
                options: termOptions,
                size: .sm,
                placeholder: termPlaceholder,
                emptyLabel: L10n.CitationComposer.termEmpty,
                label: termPlaceholder,
                accessibilityIdentifierPrefix: termIdentifierPrefix,
                focus: termFocus
            )
            .frame(width: 190)
            .id("term")
        }
    }

    private var termOptions: [PVComboBoxOption] {
        terms.map { term in
            PVComboBoxOption(value: term.id, label: Self.displayLabel(term))
        }
    }

    private var termPlaceholder: LocalizedStringResource {
        switch ready.rule.bridgeTypeKey {
        case EvidenceBridgeKind.participation.rawValue:
            L10n.CitationComposer.termPlaceholderRole
        case EvidenceBridgeKind.placeRelationship.rawValue:
            L10n.CitationComposer.termPlaceholderRelation
        default:
            L10n.CitationComposer.termPlaceholderRelationship
        }
    }

    private var pieces: [Piece] {
        let nameA = subjectA?.label ?? ""
        let nameB = subjectB?.label ?? ""
        let typeA = subjectA?.typeKey ?? ""
        switch ready.rule.bridgeTypeKey {
        case EvidenceBridgeKind.relationship.rawValue:
            return sentence([
                .word(text: nameA, strong: true, slot: 0),
                .word(text: L10n.string(L10n.CitationComposer.bridgeIsA), strong: false, slot: 0),
                .term,
                .word(text: L10n.string(L10n.CitationComposer.bridgeOf), strong: false, slot: 0),
                .word(text: nameB, strong: true, slot: 0),
            ])
        case EvidenceBridgeKind.placeRelationship.rawValue:
            return sentence([
                .word(text: nameA, strong: true, slot: 0),
                .term,
                .word(text: nameB, strong: true, slot: 0),
            ])
        case EvidenceBridgeKind.participation.rawValue:
            let person = typeA == "person" ? nameA : nameB
            let event = typeA == "event" ? nameA : nameB
            return sentence([
                .word(text: person, strong: true, slot: 0),
                .word(text: L10n.string(L10n.CitationComposer.bridgeTookPartIn), strong: false, slot: 0),
                .word(text: L10n.CitationComposer.bridgeTheEvent(event), strong: true, slot: 0),
                .word(text: L10n.string(L10n.CitationComposer.bridgeAs), strong: false, slot: 0),
                .term,
            ])
        case EvidenceBridgeKind.location.rawValue:
            let event = typeA == "event" ? nameA : nameB
            let place = typeA == "place" ? nameA : nameB
            return sentence([
                .word(text: L10n.string(L10n.CitationComposer.bridgeTheWord), strong: false, slot: 0),
                .word(text: event, strong: true, slot: 0),
                .word(text: L10n.string(L10n.CitationComposer.bridgeTookPlaceAtWords), strong: false, slot: 0),
                .word(text: place, strong: true, slot: 0),
            ])
        default:
            return sentence([
                .word(text: nameA, strong: true, slot: 0),
                .word(text: nameB, strong: true, slot: 0),
            ])
        }
    }

    /// Word slots follow the sentence so two words never share an identity.
    /// The term keeps id "term" wherever the bridge kind places it.
    private func sentence(_ parts: [Piece]) -> [Piece] {
        parts.enumerated().map { index, piece in
            if case .word(let text, let strong, _) = piece {
                return .word(text: text, strong: strong, slot: index)
            }
            return piece
        }
    }

    static func displayLabel(_ term: CatalogPropertyTerm) -> String {
        switch term.key {
        case "part_of":
            return L10n.string(L10n.CitationComposer.placeTermPartOf)
        case "succeeded_by":
            return L10n.string(L10n.CitationComposer.placeTermSucceededBy)
        default:
            return term.label.localizedLowercase
        }
    }

    private enum Piece: Identifiable {
        case word(text: String, strong: Bool, slot: Int)
        case term

        var id: String {
            switch self {
            case .word(_, _, let slot): "word-\(slot)"
            case .term: "term"
            }
        }
    }
}

/// Wraps children onto new lines when the row runs out of width.
private struct FlowLayout: Layout {
    var spacing: CGFloat
    var lineSpacing: CGFloat

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let rows = rows(in: proposal.width ?? 0, subviews: subviews)
        let height = rows.reduce(CGFloat(0)) { $0 + $1.height } + lineSpacing * CGFloat(max(rows.count - 1, 0))
        return CGSize(width: proposal.width ?? 0, height: height)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let rows = rows(in: bounds.width, subviews: subviews)
        var y = bounds.minY
        for row in rows {
            var x = bounds.minX
            for item in row.items {
                let originY = y + (row.height - item.size.height) / 2
                item.view.place(
                    at: CGPoint(x: x, y: originY),
                    anchor: .topLeading,
                    proposal: ProposedViewSize(item.size)
                )
                x += item.size.width + spacing
            }
            y += row.height + lineSpacing
        }
    }

    private func rows(in width: CGFloat, subviews: Subviews) -> [Row] {
        var rows: [Row] = []
        var items: [Item] = []
        var rowWidth: CGFloat = 0
        var rowHeight: CGFloat = 0
        for view in subviews {
            let size = view.sizeThatFits(.unspecified)
            let next = rowWidth == 0 ? size.width : rowWidth + spacing + size.width
            if rowWidth > 0, next > width {
                rows.append(Row(items: items, height: rowHeight))
                items = []
                rowWidth = 0
                rowHeight = 0
            }
            items.append(Item(view: view, size: size))
            rowWidth = rowWidth == 0 ? size.width : rowWidth + spacing + size.width
            rowHeight = max(rowHeight, size.height)
        }
        if !items.isEmpty {
            rows.append(Row(items: items, height: rowHeight))
        }
        return rows
    }

    private struct Item {
        var view: LayoutSubview
        var size: CGSize
    }

    private struct Row {
        var items: [Item]
        var height: CGFloat
    }
}
