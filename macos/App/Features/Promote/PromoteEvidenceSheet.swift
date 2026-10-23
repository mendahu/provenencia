import SwiftUI

/// The Promote evidence sheet (S11-D2): why the row was suggested, in the
/// title, then the evidence that this row is its record, as sentences the
/// researcher checks or unchecks. Checked evidence is pinned on Done.
struct PromoteEvidenceSheet: View {
    let row: PromoteFlow.Row
    let grades: [CatalogClaimConfidenceGrade]
    let model: PromoteModel

    /// Lines opened to their records, by a line's first pin id. Closing the sheet drops it.
    @State private var expanded: Set<String> = []

    var body: some View {
        PVPanel(
            title: titleText,
            subtitle: subtitleText,
            width: 880,
            titleAccessoryPlacement: .afterTitle
        ) {
            ScrollView {
                VStack(alignment: .leading, spacing: PVSpacing.space5) {
                    if let readout = row.selectionReadout {
                        EvidenceTitleLines(row: row, readout: readout, model: model)
                    }
                    if !hasRecord {
                        PVEmptyState(
                            icon: .searchEmpty,
                            title: L10n.Promote.pickRecordTitle,
                            message: L10n.string(L10n.Promote.pickRecordBody),
                            compact: true
                        )
                    }
                    if showsClaimFields {
                        claimFields
                    }
                    if case .handle(_, _, let title) = row.target {
                        evidenceArea(title: title)
                    }
                }
                .padding(PVSpacing.space5)
            }
        } footer: {
            PVButton(L10n.Promote.backToRows, variant: .primary) { model.closeSheet() }
                .accessibilityIdentifier("promote.sheet.close")
        } titleAccessory: {
            HStack(alignment: .top, spacing: PVSpacing.space4) {
                marks
                Spacer(minLength: PVSpacing.space4)
                PVIconButton(.dismiss, label: L10n.Promote.closeSheet, size: .sm) { model.closeSheet() }
                    .accessibilityIdentifier("promote.sheet.dismiss")
            }
        }
        // Nothing is lost by closing: pins and the claim fields already live in the model.
        .pvOutsideClick { model.closeSheet() }
    }

    // MARK: Title

    /// "<row name> joins <record title> <REF>" for a record; the row's name and its New option otherwise.
    private var titleText: Text {
        switch row.target {
        case .handle(_, let ref, let title):
            return Text(PromoteEvidenceDisplay.title(name: row.name, recordTitle: title, ref: ref))
        case .newKind:
            return Text(verbatim: L10n.Promote.sheetTitleNew(name: row.name, target: L10n.string(L10n.Promote.newOption(row.kind))))
        case .skip, .unset:
            return Text(verbatim: row.name)
        }
    }

    /// The claim in one sentence, under the title of a row filed on a record.
    private var subtitleText: Text? {
        guard case .handle(_, _, let title) = row.target else { return nil }
        return Text(PromoteEvidenceDisplay.subtitle(
            name: row.name, sourceTitle: model.entry.sourceTitle, kind: row.kind, recordTitle: title
        ))
    }

    /// The band and Contested marks after the title. The reason is their tooltip;
    /// the pair is one element for VoiceOver.
    @ViewBuilder
    private var marks: some View {
        if let readout = row.selectionReadout {
            let reason = model.reasonText(readout.reason, on: row)
            PromoteReadoutBadges(readout: readout)
                .help(Text(verbatim: reason))
                .focusable()
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(Text(verbatim: L10n.Promote.whySuggestedSpoken(
                    assessment: model.assessmentText(readout),
                    reason: reason
                )))
                .accessibilityIdentifier("promote.sheet.marks")
        }
    }

    private var hasRecord: Bool {
        if case .handle = row.target { return true }
        return false
    }

    // MARK: Claim fields

    private var showsClaimFields: Bool {
        switch row.target {
        case .handle, .newKind: true
        case .skip, .unset: false
        }
    }

    private var claimFields: some View {
        HStack(alignment: .top, spacing: PVSpacing.space6) {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                PVField(label: L10n.Promote.status) {
                    PVSelect(
                        selection: .constant("accepted"),
                        options: [PVSelectOption(value: "accepted", label: L10n.string(L10n.Promote.statusAccepted))]
                    )
                }
                PVField(label: L10n.Promote.confidence) {
                    PVSelect(
                        selection: confidenceBinding,
                        options: confidenceOptions,
                        accessibilityIdentifier: "promote.sheet.confidence"
                    )
                }
            }
            .frame(width: 200)
            PVField(label: L10n.Promote.argument, marker: .optional) {
                PVTextArea(text: argumentBinding)
                    .accessibilityIdentifier("promote.sheet.argument")
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }

    // MARK: Evidence

    private func evidenceArea(title: String) -> some View {
        let evidence = row.evidence
        return VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVDivider()
            PVSectionHeader(
                title: PromoteEvidenceDisplay.evidenceHeading(
                    name: row.name, sourceTitle: model.entry.sourceTitle, kind: row.kind, recordTitle: title
                ),
                meta: evidence.map {
                    L10n.Promote.checkedCount(PromoteEvidenceDisplay.checkedCount($0, pins: row.pins))
                }
            )
            .accessibilityAddTraits(.isHeader)
            .accessibilityIdentifier("promote.sheet.heading")
            if let evidence, PromoteEvidenceDisplay.hasNothingToCompare(evidence) {
                PVEmptyState(
                    icon: .searchEmpty,
                    title: L10n.Promote.nothingToCompareTitle,
                    message: L10n.string(L10n.Promote.nothingToCompareBody),
                    compact: true
                )
            }
            if case .failed(let message) = model.sheetEvidence {
                PVCallout(tone: .danger, title: L10n.Promote.evidenceUnavailable, message: message, compact: true)
            }
            if let evidence {
                cards(evidence)
            } else if model.sheetEvidence == .loading {
                VStack(alignment: .leading, spacing: PVSpacing.space3) {
                    PVSkeletonBar(width: nil, height: 12)
                    PVSkeletonBar(width: 320, height: 12)
                    PVSkeletonBar(width: 240, height: 12)
                }
                .accessibilityIdentifier("promote.sheet.loading")
            }
        }
    }

    /// The Subject's own Cards in Go's order, then one Card per relationship.
    @ViewBuilder
    private func cards(_ evidence: CatalogClaimEvidence) -> some View {
        let context = EvidenceContext(row: row, evidence: evidence, model: model)
        ForEach(evidence.sections.filter { !$0.lines.isEmpty }) { section in
            EvidenceSectionCard(context: context, section: section, expanded: $expanded)
        }
        ForEach(evidence.relationships) { relationship in
            EvidenceRelationshipCard(context: context, relationship: relationship, expanded: $expanded)
        }
    }

    // MARK: Bindings

    private var argumentBinding: Binding<String> {
        Binding(get: { row.argument }, set: { model.setArgument(subjectID: row.subjectID, argument: $0) })
    }

    private var confidenceBinding: Binding<String> {
        Binding(
            get: { row.confidenceGradeID ?? "" },
            set: { model.setConfidence(subjectID: row.subjectID, gradeID: $0.isEmpty ? nil : $0) }
        )
    }

    private var confidenceOptions: [PVSelectOption] {
        [PVSelectOption(value: "", label: L10n.string(L10n.Promote.confidenceUnset))]
            + grades.map { PVSelectOption(value: $0.id, label: $0.label) }
    }
}

// MARK: - Layout

private enum EvidenceLayout {
    /// The outcome Badge's column, so every line's Badge lines up.
    static let badgeColumn: CGFloat = 116
    static let box: CGFloat = 18
    static let inset = PVSpacing.space4
    static let gap = PVSpacing.space4
    /// A line's value column, and the Source column on the right of every line.
    static let valueColumn: CGFloat = 200
    static let sourceColumn: CGFloat = 200
    /// An expanded record's value column, and where it starts.
    static let memberValueColumn: CGFloat = 170
    static let memberIndent: CGFloat = 44
    /// A relationship's own properties start here.
    static let relationshipIndent: CGFloat = 28
    /// A likeness's crops sit side by side in the value column.
    static let cropGap: CGFloat = 8
}

private func outcomeTone(_ outcome: EvidenceOutcome) -> PVBadgeTone {
    switch outcome {
    case .agree: .success
    case .partial: .info
    case .conflict: .danger
    case .unknown, .unrecognized: .neutral
    }
}

// MARK: - Context

/// What every line needs from the sheet: the row, its evidence and the model's words.
@MainActor
private struct EvidenceContext {
    let row: PromoteFlow.Row
    let evidence: CatalogClaimEvidence
    let model: PromoteModel

    /// A value as the Conclusion page words it, in the researcher's locale and name order.
    func worded(_ value: CatalogConclusionValue) -> String {
        ReconciledValueDisplay.string(for: value, orders: model.session.nameFormat)
    }

    func incomingText(_ line: CatalogClaimEvidenceLine) -> String {
        let text = worded(line.incoming)
        return line.incomingNegative ? L10n.Promote.negatedValue(text) : text
    }

    func propertyLabel(_ key: String) -> String {
        model.propertyLabel(key: key, origin: CatalogOrigin.provenencia)
    }

    func sourceTitle(_ pin: CatalogClaimEvidencePin) -> String? {
        guard let title = evidence.sourceTitles[pin.sourceID], !title.isEmpty else { return nil }
        return title
    }

    /// A record behind a line, for the detail and the spoken labels.
    func member(_ pin: CatalogClaimEvidencePin) -> PromoteEvidenceDisplay.Member {
        PromoteEvidenceDisplay.Member(source: sourceTitle(pin), value: worded(pin.value))
    }

    func sentence(_ line: CatalogClaimEvidenceLine) -> String {
        L10n.string(L10n.Promote.lineSentence(
            relation: line.relation,
            cause: line.cause,
            propertyCause: line.propertyCause,
            value: line.incoming
        ))
    }

    /// Where a Card says what this Source has, from its lines' incoming values.
    func thisSourcePlacement(_ section: CatalogClaimEvidenceSection) -> PromoteEvidenceDisplay.ThisSourcePlacement {
        PromoteEvidenceDisplay.thisSourcePlacement(values: section.lines.map { line in
            if case .depiction = line.incoming { return nil }
            return incomingText(line)
        })
    }

    func state(_ ids: [String]) -> PromoteEvidenceDisplay.CheckState {
        PromoteEvidenceDisplay.checkState(pinIDs: ids, pins: row.pins)
    }

    /// Checked when every pin is; one press flips them all, and a partly checked box checks them all.
    func binding(_ ids: [String]) -> Binding<Bool> {
        Binding(
            get: { state(ids) == .checked },
            set: { _ in model.togglePin(subjectID: row.subjectID, pinIDs: ids) }
        )
    }
}

// MARK: - Lines under the title

/// What the row shows beside its band, kept off the badges: the Updated caption and the contested line.
private struct EvidenceTitleLines: View {
    let row: PromoteFlow.Row
    let readout: PromoteFlow.SelectionReadout
    let model: PromoteModel

    var body: some View {
        if row.updated || readout.contested {
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                if row.updated {
                    Text(verbatim: L10n.Promote.updatedDetail(model.reasonText(row.captionReason, on: row)))
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .italic()
                        .foregroundStyle(PVColor.infoForeground)
                }
                if readout.contested {
                    Text(L10n.Promote.somethingDisagrees)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .accessibilityElement(children: .contain)
            .accessibilityLabel(Text(L10n.Promote.whySuggested))
        }
    }
}

// MARK: - Property Cards

private struct EvidenceSectionCard: View {
    let context: EvidenceContext
    let section: CatalogClaimEvidenceSection
    @Binding var expanded: Set<String>

    var body: some View {
        PVCard(cornerRadius: PVRadius.sm) {
            VStack(alignment: .leading, spacing: 0) {
                EvidenceCardHeader(
                    title: PromoteEvidenceDisplay.similarityHeader(property: context.propertyLabel(section.propertyKey)),
                    thisSource: placement.headerValue
                )
                ForEach(Array(section.lines.enumerated()), id: \.element.id) { index, line in
                    if index > 0 { PVDivider() }
                    EvidenceLine(context: context, line: line, showsThisSource: placement == .lines, expanded: $expanded)
                }
            }
        }
        .accessibilityIdentifier("promote.sheet.section.\(section.propertyKey)")
    }

    private var placement: PromoteEvidenceDisplay.ThisSourcePlacement {
        context.thisSourcePlacement(section)
    }
}

/// One line on the board's grid: box, value, sentence, Source and outcome. A disclosure in the
/// Source column opens the records behind a line that bundles several.
private struct EvidenceLine: View {
    let context: EvidenceContext
    let line: CatalogClaimEvidenceLine
    /// This Source's value goes under the sentence, because the Card does not say it once.
    let showsThisSource: Bool
    @Binding var expanded: Set<String>

    private var ids: [String] { line.members.map(\.id) }
    private var key: String { ids.first ?? line.id }
    private var isOpen: Bool { expanded.contains(key) }
    private var collapses: Bool { ids.count > 1 }
    private var isLikeness: Bool {
        if case .depiction = line.incoming { return true }
        return false
    }
    private var idPrefix: String { isLikeness ? "likeness" : "pin" }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack(alignment: .top, spacing: EvidenceLayout.gap) {
                PVCheckbox(
                    isChecked: context.binding(ids),
                    isIndeterminate: context.state(ids) == .partial,
                    isDisabled: ids.isEmpty,
                    accessibilitySpokenLabel: spoken
                )
                .frame(width: EvidenceLayout.box)
                .accessibilityIdentifier("promote.sheet.\(idPrefix).\(key)")
                valueColumn
                    .frame(width: EvidenceLayout.valueColumn, alignment: .leading)
                    .accessibilityHidden(true)
                sentenceColumn
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .accessibilityHidden(true)
                sourceColumn
                    .frame(width: EvidenceLayout.sourceColumn, alignment: .leading)
                EvidenceOutcomeBadge(outcome: line.outcome, cause: line.cause)
            }
            .padding(.horizontal, EvidenceLayout.inset)
            .padding(.vertical, PVSpacing.space3)
            if collapses && isOpen {
                ForEach(line.members) { pin in
                    EvidenceMemberRow(context: context, pin: pin)
                }
            }
        }
    }

    /// The first record's value; a likeness shows its two crops, this Source's and the record's.
    @ViewBuilder
    private var valueColumn: some View {
        if isLikeness {
            HStack(alignment: .top, spacing: EvidenceLayout.cropGap) {
                EvidenceCrop(context: context, value: line.incoming)
                if let pin = line.members.first {
                    EvidenceCrop(context: context, value: pin.value)
                }
            }
        } else if let pin = line.members.first {
            Text(verbatim: context.worded(pin.value))
                .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textPrimary)
                .fixedSize(horizontal: false, vertical: true)
        }
    }

    private var sentenceColumn: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space1) {
            Text(verbatim: context.sentence(line))
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textSecondary)
                .fixedSize(horizontal: false, vertical: true)
            if showsThisSource, !isLikeness {
                Text(verbatim: L10n.Promote.thisSource(context.incomingText(line)))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }

    @ViewBuilder
    private var sourceColumn: some View {
        if collapses {
            PVDisclosureButton(
                isOpen ? L10n.Promote.hideRecords(ids.count) : L10n.Promote.showRecords(ids.count),
                isExpanded: openBinding
            )
            .accessibilityIdentifier("promote.sheet.expand.\(key)")
        } else if let pin = line.members.first {
            EvidenceSourceTitle(title: isLikeness ? PromoteEvidenceDisplay.likenessSource(title: context.sourceTitle(pin)) : context.sourceTitle(pin))
                .accessibilityHidden(true)
        }
    }

    private var openBinding: Binding<Bool> {
        Binding(
            get: { expanded.contains(key) },
            set: { if $0 { expanded.insert(key) } else { expanded.remove(key) } }
        )
    }

    private var spoken: String {
        if isLikeness {
            return L10n.Promote.spokenLikeness(
                source: PromoteEvidenceDisplay.likenessSource(title: line.members.first.flatMap(context.sourceTitle)),
                sentence: context.sentence(line)
            )
        }
        return PromoteEvidenceDisplay.spokenLabel(
            property: context.propertyLabel(line.propertyKey),
            sentence: context.sentence(line),
            here: context.incomingText(line),
            count: line.members.count,
            member: line.members.first.map(context.member)
        )
    }
}

/// A Source's title, italic and muted; nothing when there is none.
private struct EvidenceSourceTitle: View {
    let title: String?

    var body: some View {
        if let title, !title.isEmpty {
            Text(verbatim: title)
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(2)
                .truncationMode(.middle)
                .fixedSize(horizontal: false, vertical: true)
        }
    }
}

/// A likeness's stored crop, with no caption.
private struct EvidenceCrop: View {
    let context: EvidenceContext
    let value: CatalogConclusionValue

    var body: some View {
        if case .depiction(let depiction) = value {
            CachedThumbnail(
                projectDir: context.model.session.projectKey.projectDir,
                relPath: depiction.imageRelPath,
                size: LikenessTile<EmptyView>.crop,
                fallbackMark: .subjectPerson
            )
            .accessibilityHidden(true)
        }
    }
}

/// The sunken title row at the top of a Card, with what this Source has after it when every line agrees on it.
private struct EvidenceCardHeader: View {
    let title: String
    let thisSource: String?

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            Text(verbatim: title)
                .font(PVFont.display(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
            if let thisSource {
                Text(verbatim: L10n.Promote.thisSource(thisSource))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .padding(.horizontal, EvidenceLayout.inset)
        .padding(.vertical, PVSpacing.space3)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PVColor.surfaceSunken)
        .accessibilityElement(children: .combine)
        .accessibilityAddTraits(.isHeader)
    }
}

/// A line's or relationship's outcome, in the column every Badge lines up in.
private struct EvidenceOutcomeBadge: View {
    let outcome: EvidenceOutcome
    var cause: ValueCause = .none

    var body: some View {
        PVBadge(L10n.Promote.outcome(outcome, cause: cause), tone: outcomeTone(outcome), subtle: true)
            .frame(width: EvidenceLayout.badgeColumn, alignment: .trailing)
            .accessibilityHidden(true)
    }
}

/// One record under an expanded line: its own box, its value (the crop, for a likeness) and its Source.
/// Every box can be unchecked; none is disabled.
private struct EvidenceMemberRow: View {
    let context: EvidenceContext
    let pin: CatalogClaimEvidencePin

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVDivider()
            HStack(alignment: .top, spacing: EvidenceLayout.gap) {
                PVCheckbox(
                    isChecked: context.binding([pin.id]),
                    accessibilitySpokenLabel: PromoteEvidenceDisplay.spokenMember(spokenSource)
                )
                .frame(width: EvidenceLayout.box)
                .accessibilityIdentifier("promote.sheet.member.\(pin.id)")
                value
                    .frame(width: EvidenceLayout.memberValueColumn, alignment: .leading)
                    .accessibilityHidden(true)
                Spacer(minLength: 0)
                EvidenceSourceTitle(title: isLikeness ? PromoteEvidenceDisplay.likenessSource(title: context.sourceTitle(pin)) : context.sourceTitle(pin))
                    .frame(width: EvidenceLayout.sourceColumn, alignment: .leading)
                    .accessibilityHidden(true)
                Color.clear.frame(width: EvidenceLayout.badgeColumn, height: 0)
            }
            .padding(.vertical, PVSpacing.space2)
        }
        .padding(.leading, EvidenceLayout.memberIndent)
        .padding(.trailing, EvidenceLayout.inset)
    }

    private var isLikeness: Bool {
        if case .depiction = pin.value { return true }
        return false
    }

    /// The record as it is spoken: a likeness with no Source title is "a record", not "Likeness".
    private var spokenSource: PromoteEvidenceDisplay.Member {
        if case .depiction = pin.value {
            return PromoteEvidenceDisplay.Member(
                source: PromoteEvidenceDisplay.likenessSource(title: context.sourceTitle(pin)),
                value: context.worded(pin.value)
            )
        }
        return context.member(pin)
    }

    @ViewBuilder
    private var value: some View {
        if isLikeness {
            EvidenceCrop(context: context, value: pin.value)
        } else {
            Text(verbatim: context.worded(pin.value))
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textPrimary)
                .fixedSize(horizontal: false, vertical: true)
        }
    }
}

// MARK: - Relationships

/// A connection this Subject and the record's members share, worded by the link on the page, with the neighbor's lines under it.
private struct EvidenceRelationshipCard: View {
    let context: EvidenceContext
    let relationship: CatalogClaimEvidenceRelationship
    @Binding var expanded: Set<String>

    private var ids: [String] { relationship.bundles.map(\.id) }
    private var neighbor: PromoteFlow.Row? { context.model.neighbor(relationship.neighborSubjectID) }
    private var sections: [CatalogClaimEvidenceSection] { relationship.sections.filter { !$0.lines.isEmpty } }

    var body: some View {
        PVCard(cornerRadius: PVRadius.sm) {
            VStack(alignment: .leading, spacing: 0) {
                header
                if sections.isEmpty {
                    Text(L10n.Promote.nothingUnder)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .italic()
                        .foregroundStyle(PVColor.textMuted)
                        .padding(.leading, EvidenceLayout.inset + EvidenceLayout.box + EvidenceLayout.gap)
                        .padding(.trailing, EvidenceLayout.inset)
                        .padding(.bottom, PVSpacing.space3)
                } else {
                    ForEach(Array(sections.enumerated()), id: \.element.id) { index, section in
                        PVDivider()
                        let placement = context.thisSourcePlacement(section)
                        VStack(alignment: .leading, spacing: 0) {
                            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
                                Text(verbatim: PromoteEvidenceDisplay.similarityHeader(property: context.propertyLabel(section.propertyKey)))
                                    .font(PVFont.display(size: PVTypeScale.caption))
                                    .foregroundStyle(PVColor.textSecondary)
                                if let value = placement.headerValue {
                                    Text(verbatim: L10n.Promote.thisSource(value))
                                        .font(PVFont.body(size: PVTypeScale.caption))
                                        .foregroundStyle(PVColor.textMuted)
                                }
                            }
                            .padding(.horizontal, EvidenceLayout.inset)
                            .padding(.top, PVSpacing.space3)
                            .accessibilityElement(children: .combine)
                            ForEach(Array(section.lines.enumerated()), id: \.element.id) { lineIndex, line in
                                if lineIndex > 0 { PVDivider() }
                                EvidenceLine(context: context, line: line, showsThisSource: placement == .lines, expanded: $expanded)
                            }
                        }
                        .padding(.leading, EvidenceLayout.relationshipIndent)
                        .accessibilityIdentifier("promote.sheet.section.\(relationship.id).\(section.propertyKey).\(index)")
                    }
                }
            }
        }
    }

    /// "…they share a spouse (CEV-3B5RT and PER-7QH3D)", with the outcome as a badge.
    private var header: some View {
        let phrase = PromoteEvidenceDisplay.sharePhrase(
            bridge: PromoteEvidenceDisplay.bridge(for: relationship, row: context.row.subjectID, in: context.model.flow.bridges),
            neighbor: relationship.neighborSubjectID,
            neighborName: neighbor?.name ?? "",
            outcome: relationship.outcome
        )
        let title = PromoteEvidenceDisplay.shareHeader(
            phrase: phrase,
            ownRef: neighbor?.ref,
            recordRef: neighbor?.target.handleRef
        )
        return HStack(alignment: .center, spacing: EvidenceLayout.gap) {
            PVCheckbox(
                isChecked: context.binding(ids),
                isIndeterminate: context.state(ids) == .partial,
                isDisabled: ids.isEmpty,
                accessibilitySpokenLabel: L10n.Promote.spokenRelationship(
                    sentence: PromoteEvidenceDisplay.spokenHeader(title),
                    link: L10n.string(L10n.Promote.outcome(relationship.outcome))
                )
            )
            .frame(width: EvidenceLayout.box)
            .accessibilityIdentifier("promote.sheet.relationship.\(relationship.neighborSubjectID).\(relationship.key)")
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
                Text(verbatim: title)
                    .font(PVFont.display(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textDisplay)
                if ids.count > 1 {
                    Text(verbatim: L10n.Promote.recordCount(ids.count))
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .accessibilityHidden(true)
            EvidenceOutcomeBadge(outcome: relationship.outcome)
        }
        .padding(.horizontal, EvidenceLayout.inset)
        .padding(.vertical, PVSpacing.space3)
        .background(PVColor.surfaceSunken)
    }
}
