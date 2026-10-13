import SwiftUI

/// Observation dialog body. The form dialog owns Discard and Save.
struct CitationComposerAuthoringBox: View {
    @Bindable var model: CitationComposerModel

    var body: some View {
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
            valueControls(values: draft.values, focusValue: draft.focus == .value)
            PVCheckbox(
                L10n.CitationComposer.negativeLabel,
                description: L10n.CitationComposer.negativeDescription,
                isChecked: negativeBinding
            )
            if let error = draft.error {
                PVCallout(tone: .danger, message: error)
            }
        }
        .accessibilityIdentifier("citationComposer.authoring.property")
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
                    focused: draft.focus == .endA
                )
                endCombo(
                    selection: endBBinding,
                    label: L10n.CitationComposer.endB,
                    prefix: "citationComposer.draft.endB",
                    focused: draft.focus == .endB
                )
            }
            bridgeStatus(pair, draft: draft)
            if let error = draft.error {
                PVCallout(tone: .danger, message: error)
            }
        }
        .accessibilityIdentifier("citationComposer.authoring.bridge")
    }

    private func endCombo(
        selection: Binding<String>,
        label: LocalizedStringResource,
        prefix: String,
        focused: Bool
    ) -> some View {
        PVField(label: label) {
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
                focusNonce: draft.focus == .term ? model.authoringFocusNonce : 0
            )
        }
    }

    @ViewBuilder
    private func valueControls(values: ObservationRowValues, focusValue: Bool) -> some View {
        let property = model.vocabulary.property(id: values.propertyID)
        PVField(label: L10n.CitationComposer.valueLabel) {
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
                        model.beginAddCustomTerm(propertyID: values.propertyID)
                    }
                }
            case PropertyValueType.name.rawValue:
                inlineValue {
                    NameValueEditorForm(
                        draft: nameDraftBinding,
                        accessibilityIdentifierPrefix: "citationComposer.draft.name"
                    )
                }
            case PropertyValueType.date.rawValue:
                inlineValue {
                    DateValueEditorForm(
                        draft: dateDraftBinding,
                        accessibilityIdentifierPrefix: "citationComposer.draft.date"
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

    private func canSwap(_ pair: CitationPair, draft: CitationBridgeDraft) -> Bool {
        guard case .ready(let ready) = pair else { return false }
        return ready.canSwap && !draft.endA.isEmpty && !draft.endB.isEmpty
    }

    private func propertyOptions(subjectID: String) -> [PVComboBoxOption] {
        model.vocabulary.propertyOptions(forSubjectID: subjectID).map {
            PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key)
        }
    }

    private var citationBinding: Binding<String> {
        Binding(
            get: { model.authoringCitationID },
            set: { model.setAuthoringCitation($0) }
        )
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

    private var dateDraftBinding: Binding<DateValueDraft> {
        Binding(
            get: {
                if case .property(let draft) = model.authoring { return draft.values.dateDraft }
                return .empty()
            },
            set: { newValue in model.mutateProperty { $0.values.dateDraft = newValue } }
        )
    }

    private var nameDraftBinding: Binding<NameValueDraft> {
        Binding(
            get: {
                if case .property(let draft) = model.authoring { return draft.values.nameDraft }
                return .empty()
            },
            set: { newValue in model.mutateProperty { $0.values.nameDraft = newValue } }
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
    var focusNonce: Int = 0
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
                ForEach(Array(pieces.enumerated()), id: \.offset) { _, piece in
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
        case .word(let text, let strong):
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
                focusNonce: focusNonce
            )
            .frame(width: 190)
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
            return [
                .word(nameA, strong: true),
                .word(L10n.string(L10n.CitationComposer.bridgeIsA), strong: false),
                .term,
                .word(L10n.string(L10n.CitationComposer.bridgeOf), strong: false),
                .word(nameB, strong: true),
            ]
        case EvidenceBridgeKind.placeRelationship.rawValue:
            return [
                .word(nameA, strong: true),
                .term,
                .word(nameB, strong: true),
            ]
        case EvidenceBridgeKind.participation.rawValue:
            let person = typeA == "person" ? nameA : nameB
            let event = typeA == "event" ? nameA : nameB
            return [
                .word(person, strong: true),
                .word(L10n.string(L10n.CitationComposer.bridgeTookPartIn), strong: false),
                .word(L10n.CitationComposer.bridgeTheEvent(event), strong: true),
                .word(L10n.string(L10n.CitationComposer.bridgeAs), strong: false),
                .term,
            ]
        case EvidenceBridgeKind.location.rawValue:
            let event = typeA == "event" ? nameA : nameB
            let place = typeA == "place" ? nameA : nameB
            return [
                .word(L10n.string(L10n.CitationComposer.bridgeTheWord), strong: false),
                .word(event, strong: true),
                .word(L10n.string(L10n.CitationComposer.bridgeTookPlaceAtWords), strong: false),
                .word(place, strong: true),
            ]
        default:
            return [.word(nameA, strong: true), .word(nameB, strong: true)]
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

    private enum Piece {
        case word(String, strong: Bool)
        case term
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
