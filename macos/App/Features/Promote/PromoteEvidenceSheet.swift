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
            if let evidence, isEmpty(evidence) {
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
                if let evidence, PromoteEvidenceDisplay.hasConflict(evidence) {
                    PVCallout(
                        tone: .warning,
                        title: L10n.Promote.conflictCalloutTitle,
                        message: L10n.string(L10n.Promote.conflictCalloutBody),
                        compact: true
                    )
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

    /// Loaded, and nothing to compare: no line, no relationship and no lifespan.
    private func isEmpty(_ evidence: CatalogClaimEvidence) -> Bool {
        evidence.sections.allSatisfy { $0.lines.isEmpty } && evidence.relationships.isEmpty && evidence.lifespan == nil
    }

    /// The Subject's own Cards in Go's order, the lifespan last among them, then one Card per relationship.
    @ViewBuilder
    private func cards(_ evidence: CatalogClaimEvidence) -> some View {
        let context = EvidenceContext(row: row, evidence: evidence, model: model)
        ForEach(evidence.sections.filter { !$0.lines.isEmpty }) { section in
            EvidenceSectionCard(context: context, section: section, expanded: $expanded)
        }
        if let lifespan = evidence.lifespan {
            EvidenceLifespanCard(lifespan: lifespan)
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
    static let crop: CGFloat = 96
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

    /// "Source: value", or the value alone when the Source has no title.
    func memberText(_ pin: CatalogClaimEvidencePin) -> String {
        guard let title = sourceTitle(pin) else { return worded(pin.value) }
        return L10n.Promote.onSource(source: title, value: worded(pin.value))
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
                Text(verbatim: context.propertyLabel(section.propertyKey))
                    .font(PVFont.display(size: PVTypeScale.h4, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textDisplay)
                    .padding(.horizontal, EvidenceLayout.inset)
                    .padding(.vertical, PVSpacing.space3)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(PVColor.surfaceSunken)
                    .accessibilityAddTraits(.isHeader)
                ForEach(Array(section.lines.enumerated()), id: \.element.id) { index, line in
                    if index > 0 { PVDivider() }
                    if case .depiction = line.incoming {
                        EvidenceLikenessLine(context: context, line: line, expanded: $expanded)
                    } else {
                        EvidenceLineRow(context: context, line: line, expanded: $expanded)
                    }
                }
            }
        }
        .accessibilityIdentifier("promote.sheet.section.\(section.propertyKey)")
    }
}

/// One line: a checkbox, the sentence over its detail, a disclosure when several records stand behind it, and the outcome.
private struct EvidenceLineRow: View {
    let context: EvidenceContext
    let line: CatalogClaimEvidenceLine
    @Binding var expanded: Set<String>

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
                .accessibilityIdentifier("promote.sheet.pin.\(key)")
                VStack(alignment: .leading, spacing: PVSpacing.space1) {
                    EvidenceSentence(text: context.sentence(line), count: collapses ? ids.count : nil)
                    Text(verbatim: detail)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                        .fixedSize(horizontal: false, vertical: true)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .accessibilityHidden(true)
                if collapses {
                    PVDisclosureButton(
                        isOpen ? L10n.string(L10n.Promote.hideRecords) : L10n.Promote.showRecords(ids.count),
                        isExpanded: openBinding
                    )
                    .accessibilityIdentifier("promote.sheet.expand.\(key)")
                }
                PVBadge(L10n.Promote.outcome(line.outcome, cause: line.cause), tone: outcomeTone(line.outcome), subtle: true)
                    .frame(width: EvidenceLayout.badgeColumn, alignment: .trailing)
                    .accessibilityHidden(true)
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

    private var openBinding: Binding<Bool> {
        Binding(
            get: { expanded.contains(key) },
            set: { if $0 { expanded.insert(key) } else { expanded.remove(key) } }
        )
    }

    /// "This source: <value>", and for one record its Source and value; a collapsed line shows only the first.
    private var detail: String {
        var parts = [L10n.Promote.thisSource(context.incomingText(line))]
        if line.members.count == 1, let pin = line.members.first {
            parts.append(context.memberText(pin))
        }
        return L10n.Promote.lineDetail(parts: parts)
    }

    private var spoken: String {
        let property = context.propertyLabel(line.propertyKey)
        let sentence = context.sentence(line)
        let here = context.incomingText(line)
        if collapses {
            return L10n.Promote.spokenLineCount(property: property, sentence: sentence, count: ids.count, here: here)
        }
        let base = L10n.Promote.spokenLine(property: property, sentence: sentence, here: here)
        guard let pin = line.members.first else { return base }
        return base + " " + L10n.Promote.spokenMember(source: context.sourceTitle(pin) ?? "", value: context.worded(pin.value))
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

/// One record under an expanded line: its own box, its value and its Source.
private struct EvidenceMemberRow: View {
    let context: EvidenceContext
    let pin: CatalogClaimEvidencePin

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVDivider()
            HStack(alignment: .top, spacing: EvidenceLayout.gap) {
                PVCheckbox(
                    isChecked: context.binding([pin.id]),
                    accessibilitySpokenLabel: L10n.Promote.spokenMember(
                        source: context.sourceTitle(pin) ?? "",
                        value: context.worded(pin.value)
                    )
                )
                .accessibilityIdentifier("promote.sheet.member.\(pin.id)")
                VStack(alignment: .leading, spacing: PVSpacing.space1) {
                    Text(verbatim: context.worded(pin.value))
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                    if let title = context.sourceTitle(pin) {
                        Text(verbatim: title)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .accessibilityHidden(true)
            }
            .padding(.vertical, PVSpacing.space3)
        }
        .padding(.leading, EvidenceLayout.textIndent)
        .padding(.trailing, EvidenceLayout.inset)
    }
}

// MARK: - Likeness

/// A likeness line: this Source's crop beside the first record's, then the outcome.
private struct EvidenceLikenessLine: View {
    let context: EvidenceContext
    let line: CatalogClaimEvidenceLine
    @Binding var expanded: Set<String>

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
                    accessibilitySpokenLabel: L10n.Promote.spokenLikeness(
                        source: memberSource,
                        sentence: context.sentence(line)
                    )
                )
                .accessibilityIdentifier("promote.sheet.likeness.\(key)")
                VStack(alignment: .leading, spacing: PVSpacing.space4) {
                    EvidenceSentence(text: context.sentence(line), count: collapses ? ids.count : nil)
                    HStack(alignment: .top, spacing: PVSpacing.space6) {
                        LikenessTile(
                            projectDir: context.model.session.projectKey.projectDir,
                            relPath: relPath(line.incoming),
                            caption: L10n.string(L10n.Promote.thisSourceCaption),
                            captionWidth: EvidenceLayout.crop
                        )
                        if let pin = line.members.first {
                            LikenessTile(
                                projectDir: context.model.session.projectKey.projectDir,
                                relPath: relPath(pin.value),
                                caption: context.sourceTitle(pin) ?? "",
                                captionWidth: EvidenceLayout.crop
                            )
                        }
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .accessibilityHidden(true)
                if collapses {
                    PVDisclosureButton(
                        isOpen ? L10n.string(L10n.Promote.hideRecords) : L10n.Promote.showRecords(ids.count),
                        isExpanded: openBinding
                    )
                    .accessibilityIdentifier("promote.sheet.expand.\(key)")
                }
                PVBadge(L10n.Promote.outcome(line.outcome, cause: line.cause), tone: outcomeTone(line.outcome), subtle: true)
                    .frame(width: EvidenceLayout.badgeColumn, alignment: .trailing)
                    .accessibilityHidden(true)
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

    private var openBinding: Binding<Bool> {
        Binding(
            get: { expanded.contains(key) },
            set: { if $0 { expanded.insert(key) } else { expanded.remove(key) } }
        )
    }

    /// The first record's Source, or its value when the Source has no title.
    private var memberSource: String {
        guard let pin = line.members.first else { return "" }
        return context.sourceTitle(pin) ?? context.worded(pin.value)
    }

    private func relPath(_ value: CatalogConclusionValue) -> String {
        if case .depiction(let depiction) = value { return depiction.imageRelPath }
        return ""
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
                                if case .depiction = line.incoming {
                                    EvidenceLikenessLine(context: context, line: line, expanded: $expanded)
                                } else {
                                    EvidenceLineRow(context: context, line: line, expanded: $expanded)
                                }
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
        let sentence = L10n.string(L10n.Promote.relationshipSentence(relationship.outcome))
        let link = linkText
        return HStack(alignment: .top, spacing: EvidenceLayout.gap) {
            PVCheckbox(
                isChecked: context.binding(ids),
                isIndeterminate: context.state(ids) == .partial,
                isDisabled: !relationship.pinnable || ids.isEmpty,
                accessibilitySpokenLabel: L10n.Promote.spokenRelationship(sentence: sentence, link: link)
            )
            .accessibilityIdentifier("promote.sheet.relationship.\(relationship.neighborSubjectID).\(relationship.key)")
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                Text(verbatim: sentence)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textPrimary)
                HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
                    Text(verbatim: link)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textSecondary)
                    if ids.count > 1 {
                        Text(verbatim: L10n.Promote.recordCount(ids.count))
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
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
            PVBadge(L10n.Promote.outcome(relationship.outcome), tone: outcomeTone(relationship.outcome), subtle: true)
                .frame(width: EvidenceLayout.badgeColumn, alignment: .trailing)
                .accessibilityHidden(true)
        }
        .padding(.horizontal, EvidenceLayout.inset)
        .padding(.vertical, PVSpacing.space3)
    }

    /// The link on this Source as the page names it, else the neighbor by name.
    private var linkText: String {
        if let bridge = PromoteEvidenceDisplay.bridge(
            for: relationship,
            row: context.row.subjectID,
            in: context.model.flow.bridges
        ), !bridge.endAName.isEmpty, !bridge.endBName.isEmpty, !bridge.phrase.isEmpty {
            return L10n.Promote.connectionLink(first: bridge.endAName, phrase: bridge.phrase, second: bridge.endBName)
        }
        return L10n.Promote.relationshipWith(neighbor?.name ?? L10n.string(L10n.Promote.sheetNeighbor))
    }
}

// MARK: - Lifespan

/// The lifespan check: one line, no checkbox, never pinned.
private struct EvidenceLifespanCard: View {
    let lifespan: CatalogClaimEvidenceLifespan

    var body: some View {
        let sentence = PromoteEvidenceDisplay.lifespanSentence(lifespan)
        PVCard(cornerRadius: PVRadius.sm) {
            VStack(alignment: .leading, spacing: 0) {
                Text(L10n.Promote.lifespanTitle)
                    .font(PVFont.display(size: PVTypeScale.h4, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textDisplay)
                    .padding(.horizontal, EvidenceLayout.inset)
                    .padding(.vertical, PVSpacing.space3)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(PVColor.surfaceSunken)
                    .accessibilityAddTraits(.isHeader)
                HStack(alignment: .top, spacing: EvidenceLayout.gap) {
                    PVIcon(.info, size: EvidenceLayout.box)
                        .foregroundStyle(PVColor.textFaint)
                        .accessibilityHidden(true)
                    VStack(alignment: .leading, spacing: PVSpacing.space1) {
                        Text(verbatim: sentence)
                            .font(PVFont.body(size: PVTypeScale.bodySmall))
                            .foregroundStyle(PVColor.textPrimary)
                        Text(L10n.Promote.lifespanCaption)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .italic()
                            .foregroundStyle(PVColor.textMuted)
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                    PVBadge(L10n.Promote.outcome(lifespan.outcome), tone: outcomeTone(lifespan.outcome), subtle: true)
                        .frame(width: EvidenceLayout.badgeColumn, alignment: .trailing)
                        .accessibilityHidden(true)
                }
                .padding(.horizontal, EvidenceLayout.inset)
                .padding(.vertical, PVSpacing.space3)
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(Text(verbatim: sentence))
            }
        }
        .accessibilityIdentifier("promote.sheet.lifespan")
    }
}
