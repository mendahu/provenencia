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
        PVPanel(title: Text(verbatim: sheetTitle), width: 880) {
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
                    if case .handle(_, let ref, let title) = row.target {
                        evidenceArea(ref: ref, title: title)
                    }
                }
                .padding(PVSpacing.space5)
            }
        } footer: {
            PVButton(L10n.Promote.backToRows, variant: .primary) { model.closeSheet() }
                .accessibilityIdentifier("promote.sheet.close")
        } titleAccessory: {
            marks
        }
    }

    // MARK: Title

    private var sheetTitle: String {
        switch row.target {
        case .handle(_, let ref, let title):
            return L10n.Promote.sheetTitleHandle(name: row.name, ref: ref, title: title)
        case .newKind:
            return L10n.Promote.sheetTitleNew(name: row.name, target: L10n.string(L10n.Promote.newOption(row.kind)))
        case .skip, .unset:
            return row.name
        }
    }

    /// The band and Contested marks at the right end of the title. The reason is their tooltip;
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

    private func evidenceArea(ref: String, title: String) -> some View {
        let evidence = row.evidence
        return VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVSectionHeader(
                title: L10n.Promote.evidenceHeading(ref: ref, title: title),
                meta: evidence.map {
                    L10n.Promote.checkedCount(PromoteEvidenceDisplay.checkedCount($0, pins: row.pins))
                }
            )
            .accessibilityAddTraits(.isHeader)
            .accessibilityIdentifier("promote.sheet.heading")
            if let evidence {
                if PromoteEvidenceDisplay.hasNothingToCompare(evidence) {
                    PVEmptyState(
                        icon: .searchEmpty,
                        title: L10n.Promote.nothingToCompareTitle,
                        message: L10n.string(L10n.Promote.nothingToCompareBody),
                        compact: true
                    )
                } else {
                    Text(L10n.Promote.evidenceIntro)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textSecondary)
                        .fixedSize(horizontal: false, vertical: true)
                    if PromoteEvidenceDisplay.hasConflict(evidence) {
                        PVCallout(
                            tone: .warning,
                            title: L10n.Promote.conflictCalloutTitle,
                            message: L10n.string(L10n.Promote.conflictCalloutBody),
                            compact: true
                        )
                    }
                }
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
    static let box: CGFloat = 16
    static let inset = PVSpacing.space4
    static let gap = PVSpacing.space5
    /// A relationship's own properties start here.
    static let relationshipIndent: CGFloat = 28
    /// Where a line's text starts: after its inset, box and gap.
    static let textIndent = inset + box + gap
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
        model.propertyLabel(key: key, origin: "provenencia")
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
                EvidenceCardHeader(title: Text(verbatim: context.propertyLabel(section.propertyKey)))
                ForEach(Array(section.lines.enumerated()), id: \.element.id) { index, line in
                    if index > 0 { PVDivider() }
                    EvidenceLineView(context: context, line: line, expanded: $expanded)
                }
            }
        }
        .accessibilityIdentifier("promote.sheet.section.\(section.propertyKey)")
    }
}

/// Picks a line's middle column: the two crops of a likeness, else the sentence over its detail.
private struct EvidenceLineView: View {
    let context: EvidenceContext
    let line: CatalogClaimEvidenceLine
    @Binding var expanded: Set<String>

    private var count: Int? { line.members.count > 1 ? line.members.count : nil }

    var body: some View {
        if case .depiction = line.incoming {
            EvidenceLine(
                context: context, line: line, idPrefix: "likeness", spoken: likenessSpoken, expanded: $expanded
            ) {
                likenessMiddle
            }
        } else {
            EvidenceLine(context: context, line: line, idPrefix: "pin", spoken: spoken, expanded: $expanded) {
                textMiddle
            }
        }
    }

    private var textMiddle: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space1) {
            EvidenceSentence(text: context.sentence(line), count: count)
            Text(verbatim: PromoteEvidenceDisplay.detail(
                here: context.incomingText(line),
                member: line.members.count == 1 ? line.members.first.map(context.member) : nil
            ))
            .font(PVFont.body(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textMuted)
            .fixedSize(horizontal: false, vertical: true)
        }
    }

    private var likenessMiddle: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            EvidenceSentence(text: context.sentence(line), count: count)
            HStack(alignment: .top, spacing: PVSpacing.space6) {
                LikenessTile(
                    projectDir: context.model.session.projectKey.projectDir,
                    relPath: relPath(line.incoming),
                    caption: L10n.string(L10n.Promote.thisSourceCaption),
                    captionWidth: LikenessTile<EmptyView>.crop
                )
                if let pin = line.members.first {
                    LikenessTile(
                        projectDir: context.model.session.projectKey.projectDir,
                        relPath: relPath(pin.value),
                        caption: PromoteEvidenceDisplay.likenessSource(title: context.sourceTitle(pin)),
                        captionWidth: LikenessTile<EmptyView>.crop
                    )
                }
            }
        }
    }

    private func relPath(_ value: CatalogConclusionValue) -> String {
        if case .depiction(let depiction) = value { return depiction.imageRelPath }
        return ""
    }

    private var spoken: String {
        PromoteEvidenceDisplay.spokenLabel(
            property: context.propertyLabel(line.propertyKey),
            sentence: context.sentence(line),
            here: context.incomingText(line),
            count: line.members.count,
            member: line.members.first.map(context.member)
        )
    }

    private var likenessSpoken: String {
        L10n.Promote.spokenLikeness(
            source: PromoteEvidenceDisplay.likenessSource(title: line.members.first.flatMap(context.sourceTitle)),
            sentence: context.sentence(line)
        )
    }
}

/// One line: a checkbox, its middle column, a disclosure when several records stand behind it, and the outcome.
private struct EvidenceLine<Middle: View>: View {
    let context: EvidenceContext
    let line: CatalogClaimEvidenceLine
    /// The identifier family of the line's box: `pin` or `likeness`.
    let idPrefix: String
    let spoken: String
    @Binding var expanded: Set<String>
    private let middle: Middle

    init(
        context: EvidenceContext,
        line: CatalogClaimEvidenceLine,
        idPrefix: String,
        spoken: String,
        expanded: Binding<Set<String>>,
        @ViewBuilder middle: () -> Middle
    ) {
        self.context = context
        self.line = line
        self.idPrefix = idPrefix
        self.spoken = spoken
        self._expanded = expanded
        self.middle = middle()
    }

    private var ids: [String] { line.members.map(\.id) }
    private var key: String { ids.first ?? line.id }
    private var isOpen: Bool { expanded.contains(key) }
    private var collapses: Bool { ids.count > 1 }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack(alignment: .top, spacing: EvidenceLayout.gap) {
                PVCheckbox(
                    isChecked: context.binding(ids),
                    isIndeterminate: context.state(ids) == .partial,
                    isDisabled: !line.pinnable || ids.isEmpty,
                    accessibilitySpokenLabel: spoken
                )
                .accessibilityIdentifier("promote.sheet.\(idPrefix).\(key)")
                middle
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .accessibilityHidden(true)
                if collapses {
                    PVDisclosureButton(
                        isOpen ? L10n.string(L10n.Promote.hideRecords) : L10n.Promote.showRecords(ids.count),
                        isExpanded: openBinding
                    )
                    .accessibilityIdentifier("promote.sheet.expand.\(key)")
                }
                EvidenceOutcomeBadge(outcome: line.outcome, cause: line.cause)
            }
            .padding(.horizontal, EvidenceLayout.inset)
            .padding(.vertical, PVSpacing.space3)
            if collapses && isOpen {
                ForEach(line.members) { pin in
                    EvidenceMemberRow(context: context, pin: pin, isDisabled: PromoteEvidenceDisplay.memberDisabled(line: line))
                }
            }
        }
    }

    private var openBinding: Binding<Bool> {
        Binding(
            get: { expanded.contains(key) },
            set: { if $0 { expanded.insert(key) } else { expanded.remove(key) } }
        )
    }
}

/// A sentence with its muted record count after it.
private struct EvidenceSentence: View {
    let text: String
    let count: Int?

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
            Text(verbatim: text)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textPrimary)
            if let count {
                Text(verbatim: L10n.Promote.recordCount(count))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
    }
}

/// The sunken title row at the top of a Card.
private struct EvidenceCardHeader: View {
    let title: Text

    var body: some View {
        title
            .font(PVFont.display(size: PVTypeScale.h4, weight: PVFontWeight.medium))
            .foregroundStyle(PVColor.textDisplay)
            .padding(.horizontal, EvidenceLayout.inset)
            .padding(.vertical, PVSpacing.space3)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(PVColor.surfaceSunken)
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
private struct EvidenceMemberRow: View {
    let context: EvidenceContext
    let pin: CatalogClaimEvidencePin
    let isDisabled: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVDivider()
            HStack(alignment: isLikeness ? .top : .center, spacing: EvidenceLayout.gap) {
                PVCheckbox(
                    isChecked: context.binding([pin.id]),
                    isDisabled: isDisabled,
                    accessibilitySpokenLabel: PromoteEvidenceDisplay.spokenMember(spokenSource)
                )
                .accessibilityIdentifier("promote.sheet.member.\(pin.id)")
                content
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .accessibilityHidden(true)
            }
            .padding(.vertical, PVSpacing.space2)
        }
        .padding(.leading, EvidenceLayout.textIndent)
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
    private var content: some View {
        if case .depiction(let depiction) = pin.value {
            LikenessTile(
                projectDir: context.model.session.projectKey.projectDir,
                relPath: depiction.imageRelPath,
                caption: PromoteEvidenceDisplay.likenessSource(title: context.sourceTitle(pin)),
                captionWidth: LikenessTile<EmptyView>.crop
            )
        } else {
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
                Text(verbatim: context.worded(pin.value))
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(1)
                    .truncationMode(.tail)
                    .layoutPriority(1)
                if let title = context.sourceTitle(pin) {
                    Spacer(minLength: PVSpacing.space3)
                    Text(verbatim: title)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                        .truncationMode(.middle)
                        .multilineTextAlignment(.trailing)
                }
            }
        }
    }
}

// MARK: - Relationships

/// A connection this Subject and the record's members share, named by the link on the page, with the neighbor's lines under it.
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
                        .padding(.leading, EvidenceLayout.textIndent)
                        .padding(.trailing, EvidenceLayout.inset)
                        .padding(.bottom, PVSpacing.space3)
                } else {
                    ForEach(Array(sections.enumerated()), id: \.element.id) { index, section in
                        PVDivider()
                        VStack(alignment: .leading, spacing: 0) {
                            Text(verbatim: context.propertyLabel(section.propertyKey))
                                .pvMicroCaps()
                                .foregroundStyle(PVColor.textMuted)
                                .padding(.horizontal, EvidenceLayout.inset)
                                .padding(.top, PVSpacing.space3)
                            ForEach(Array(section.lines.enumerated()), id: \.element.id) { lineIndex, line in
                                if lineIndex > 0 { PVDivider() }
                                EvidenceLineView(context: context, line: line, expanded: $expanded)
                            }
                        }
                        .padding(.leading, EvidenceLayout.relationshipIndent)
                        .accessibilityIdentifier("promote.sheet.section.\(relationship.id).\(section.propertyKey).\(index)")
                    }
                }
            }
        }
    }

    private var header: some View {
        let heading = PromoteEvidenceDisplay.relationshipHeader(
            bridge: PromoteEvidenceDisplay.bridge(for: relationship, row: context.row.subjectID, in: context.model.flow.bridges),
            outcome: relationship.outcome,
            neighborName: neighbor?.name
        )
        return HStack(alignment: heading.subtitle == nil ? .center : .top, spacing: EvidenceLayout.gap) {
            PVCheckbox(
                isChecked: context.binding(ids),
                isIndeterminate: context.state(ids) == .partial,
                isDisabled: !relationship.pinnable || ids.isEmpty,
                accessibilitySpokenLabel: heading.spoken
            )
            .accessibilityIdentifier("promote.sheet.relationship.\(relationship.neighborSubjectID).\(relationship.key)")
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
                    Text(verbatim: heading.title)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textPrimary)
                    if heading.subtitle == nil, ids.count > 1 { recordCount }
                }
                if let subtitle = heading.subtitle {
                    HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
                        Text(verbatim: subtitle)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textSecondary)
                        if ids.count > 1 { recordCount }
                    }
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .accessibilityHidden(true)
            if let neighbor {
                Text(verbatim: neighbor.shownRef)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .accessibilityHidden(true)
            }
            EvidenceOutcomeBadge(outcome: relationship.outcome)
        }
        .padding(.horizontal, EvidenceLayout.inset)
        .padding(.vertical, PVSpacing.space3)
    }

    private var recordCount: some View {
        Text(verbatim: L10n.Promote.recordCount(ids.count))
            .font(PVFont.body(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textMuted)
    }
}
