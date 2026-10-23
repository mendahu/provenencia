import SwiftUI

// The evidence sheet's Cards: property Cards and relationship Cards, their lines, the records
// under a line, and the one place their words get a color and an emphasis. The sheet
// (`PromoteEvidenceSheet`) holds the panel, the title, the claim fields and the bindings.

// MARK: - Layout

private enum EvidenceLayout {
    static let inset = PVSpacing.space4
    static let gap = PVSpacing.space4
    /// A relationship's own properties start here.
    static let relationshipIndent = PVSpacing.space6 + PVSpacing.space5
    // The board's grid: no token equals these, so each is named once with where it comes from.
    /// The checkbox column (board: 18 px).
    static let box: CGFloat = 18
    /// The outcome Badge's column, so every line's Badge lines up (board: 116 px).
    static let badgeColumn: CGFloat = 116
    /// A line's value column, and the Source column on the right of every line (board: 200 px each).
    static let valueColumn: CGFloat = 200
    static let sourceColumn: CGFloat = 200
    /// An expanded record's value column (board: 170 px), and where it starts (board: 44 px).
    static let memberValueColumn: CGFloat = 170
    static let memberIndent: CGFloat = 44
    /// A likeness's crops sit side by side in the value column.
    static let cropGap = PVSpacing.space4
}

private func outcomeTone(_ outcome: EvidenceOutcome) -> PVBadgeTone {
    switch outcome {
    case .agree: .success
    case .partial: .info
    case .conflict: .danger
    case .unknown, .unrecognized: .neutral
    }
}

// MARK: - What the board does not carry

/// What a line needs from the sheet that is not in the board: what its boxes and its disclosure do.
/// The closures never decide whether a line draws again.
struct EvidenceEnv {
    /// Checks or unchecks the pins of a line, a record or a relationship together.
    let toggle: ([String]) -> Void
    /// Opens or closes the records behind a line, by its key.
    let setOpen: (String, Bool) -> Void
}

/// A box that checks or unchecks a line's, a record's or a relationship's pins together: checked
/// when every pin is, and a partly checked box checks them all.
private struct EvidencePinBox: View {
    let ids: [String]
    /// Which of `ids` are checked.
    let checked: Set<String>
    let spoken: String
    let identifier: String
    let env: EvidenceEnv

    var body: some View {
        let state = PromoteEvidenceDisplay.checkState(pinIDs: ids, pins: checked)
        PVCheckbox(
            isChecked: Binding(get: { state == .checked }, set: { _ in env.toggle(ids) }),
            isIndeterminate: state == .partial,
            isDisabled: ids.isEmpty,
            accessibilitySpokenLabel: spoken
        )
        .frame(width: EvidenceLayout.box)
        .accessibilityIdentifier(identifier)
    }
}

// MARK: - Property Cards

struct EvidenceSectionCard: View {
    let group: EvidenceBoard.Group
    let pins: Set<String>
    let expanded: Set<String>
    /// Where a likeness's crops are stored.
    let projectDir: String
    let env: EvidenceEnv

    var body: some View {
        PVCard(cornerRadius: PVRadius.sm) {
            EvidenceSectionLines(group: group, style: .card, pins: pins, expanded: expanded, projectDir: projectDir, env: env)
        }
        .accessibilityIdentifier("promote.sheet.section.\(group.id)")
    }
}

/// One property's header and its lines, divided: the body of a Card and of a group under a
/// relationship. A line draws again only when its own value, its own checked pins or its own
/// disclosure change.
private struct EvidenceSectionLines: View {
    let group: EvidenceBoard.Group
    let style: EvidenceCardHeader.Style
    let pins: Set<String>
    let expanded: Set<String>
    let projectDir: String
    let env: EvidenceEnv

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            EvidenceCardHeader(title: group.header, thisSource: group.thisSource, style: style)
            ForEach(Array(group.lines.enumerated()), id: \.element.id) { index, line in
                if index > 0 { PVDivider() }
                EvidenceLine(
                    line: line,
                    checked: pins.intersection(line.pinIDs),
                    isOpen: expanded.contains(line.key),
                    projectDir: projectDir,
                    env: env
                )
                .equatable()
            }
        }
    }
}

/// One line on the board's grid: box, value, sentence, Source and outcome. A disclosure in the
/// Source column opens the records behind a line that bundles several.
struct EvidenceLine: View, Equatable {
    let line: EvidenceBoard.Line
    /// The line's own pins that are checked.
    let checked: Set<String>
    let isOpen: Bool
    let projectDir: String
    let env: EvidenceEnv

    nonisolated static func == (lhs: EvidenceLine, rhs: EvidenceLine) -> Bool {
        lhs.line == rhs.line && lhs.checked == rhs.checked && lhs.isOpen == rhs.isOpen
            && lhs.projectDir == rhs.projectDir
    }

    private var idPrefix: String { line.isLikeness ? "likeness" : "pin" }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack(alignment: .top, spacing: EvidenceLayout.gap) {
                EvidencePinBox(
                    ids: line.pinIDs, checked: checked, spoken: line.spoken,
                    identifier: "promote.sheet.\(idPrefix).\(line.key)", env: env
                )
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
            if line.collapses && isOpen {
                ForEach(line.members) { member in
                    EvidenceMemberRow(member: member, checked: checked, projectDir: projectDir, env: env)
                }
            }
        }
    }

    /// The first record's value; a likeness shows its two crops, this Source's and the record's.
    private var valueColumn: some View {
        EvidenceValueView(value: line.value, weight: PVFontWeight.medium, projectDir: projectDir)
    }

    private var sentenceColumn: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space1) {
            Text(verbatim: line.sentence)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textSecondary)
                .fixedSize(horizontal: false, vertical: true)
            if let thisSource = line.thisSource {
                Text(verbatim: thisSource)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }

    @ViewBuilder
    private var sourceColumn: some View {
        if line.collapses {
            // The label starts on the Source column's edge, where the record titles start.
            PVDisclosureButton(
                isOpen ? L10n.Promote.hideRecords(line.pinIDs.count) : L10n.Promote.showRecords(line.pinIDs.count),
                isExpanded: Binding(get: { isOpen }, set: { env.setOpen(line.key, $0) }),
                flushLeading: true
            )
            .accessibilityIdentifier("promote.sheet.expand.\(line.key)")
        } else if !line.pinIDs.isEmpty {
            EvidenceSourceTitle(title: line.source)
                .accessibilityHidden(true)
        }
    }
}

/// A Source's title, italic and muted, with the whole title as its tooltip when the column cuts it off.
private struct EvidenceSourceTitle: View {
    let title: String

    var body: some View {
        Text(verbatim: title)
            .font(PVFont.body(size: PVTypeScale.caption, italic: true))
            .foregroundStyle(PVColor.textMuted)
            .lineLimit(2)
            .truncationMode(.middle)
            .fixedSize(horizontal: false, vertical: true)
            .help(Text(verbatim: title))
    }
}

/// Draws the roles ``PromoteEvidenceDisplay`` names: the one place the sheet's words get a color
/// and an emphasis. A name is set by weight, so it takes the size of the text around it. Typed
/// `AttributeContainer` subscripts, as `PVComboBox` uses, not dynamic members: those warn under
/// Swift 6 key-path rules.
enum EvidenceTextStyle {
    static func painted(_ text: AttributedString) -> AttributedString {
        var result = text
        for run in text.runs {
            guard let role = run[PromoteEvidenceDisplay.RoleAttribute.self] else { continue }
            var attributes = AttributeContainer()
            switch role {
            case .name:
                attributes[AttributeScopes.FoundationAttributes.InlinePresentationIntentAttribute.self] = .stronglyEmphasized
                attributes[AttributeScopes.SwiftUIAttributes.ForegroundColorAttribute.self] = PVColor.textDisplay
            case .source:
                attributes[AttributeScopes.FoundationAttributes.InlinePresentationIntentAttribute.self] = .emphasized
                attributes[AttributeScopes.SwiftUIAttributes.ForegroundColorAttribute.self] = PVColor.textLink
            case .muted:
                attributes[AttributeScopes.FoundationAttributes.InlinePresentationIntentAttribute.self] = .emphasized
                attributes[AttributeScopes.SwiftUIAttributes.ForegroundColorAttribute.self] = PVColor.textMuted
            case .ref:
                attributes[AttributeScopes.SwiftUIAttributes.FontAttribute.self] = PVFont.mono(size: PVTypeScale.caption)
                attributes[AttributeScopes.SwiftUIAttributes.ForegroundColorAttribute.self] = PVColor.textMuted
            }
            result[run.range].mergeAttributes(attributes)
        }
        return result
    }
}

/// A record's value in a line or under it: its words at `weight`, or a likeness's crops.
private struct EvidenceValueView: View {
    let value: EvidenceBoard.Value
    let weight: CGFloat
    let projectDir: String

    var body: some View {
        switch value {
        case .none:
            EmptyView()
        case .text(let text):
            Text(verbatim: text)
                .font(PVFont.body(size: PVTypeScale.bodySmall, weight: weight))
                .foregroundStyle(PVColor.textPrimary)
                .fixedSize(horizontal: false, vertical: true)
        case .crops:
            EvidenceCrops(value: value, projectDir: projectDir)
        }
    }
}

/// A likeness's stored crops side by side, with no caption.
private struct EvidenceCrops: View {
    let value: EvidenceBoard.Value
    let projectDir: String

    var body: some View {
        if case .crops(let paths) = value {
            HStack(alignment: .top, spacing: EvidenceLayout.cropGap) {
                ForEach(Array(paths.enumerated()), id: \.offset) { _, path in
                    CachedThumbnail(
                        projectDir: projectDir,
                        relPath: path,
                        size: CachedThumbnail.likenessCrop,
                        fallbackMark: .subjectPerson
                    )
                    .accessibilityHidden(true)
                }
            }
        }
    }
}

/// The title row of a Card or of a group under a relationship, with what this Source has after it
/// when every line agrees on it.
private struct EvidenceCardHeader: View {
    /// The sunken row at the top of a Card, or the lighter label over a relationship's group.
    enum Style {
        case card
        case group
    }

    /// The header's words, without the ellipsis it is drawn after; VoiceOver reads them as they are.
    let title: String
    let thisSource: String?
    var style: Style = .card

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            Text(verbatim: L10n.Promote.cardHeaderLead(title))
                .font(style == .card
                    ? PVFont.display(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium)
                    : PVFont.display(size: PVTypeScale.caption))
                .foregroundStyle(style == .card ? PVColor.textDisplay : PVColor.textSecondary)
                .accessibilityLabel(Text(verbatim: title))
            if let thisSource {
                Text(verbatim: L10n.Promote.thisSource(thisSource))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .padding(.horizontal, EvidenceLayout.inset)
        .padding(.top, PVSpacing.space3)
        .padding(.bottom, style == .card ? PVSpacing.space3 : 0)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(style == .card ? PVColor.surfaceSunken : Color.clear)
        .accessibilityElement(children: .combine)
        .accessibilityAddTraits(style == .card ? .isHeader : [])
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
    let member: EvidenceBoard.Member
    /// The expanded line's checked pins.
    let checked: Set<String>
    let projectDir: String
    let env: EvidenceEnv

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVDivider()
            HStack(alignment: .top, spacing: EvidenceLayout.gap) {
                EvidencePinBox(
                    ids: [member.id], checked: checked, spoken: member.spoken,
                    identifier: "promote.sheet.member.\(member.id)", env: env
                )
                EvidenceValueView(value: member.value, weight: PVFontWeight.regular, projectDir: projectDir)
                    .frame(width: EvidenceLayout.memberValueColumn, alignment: .leading)
                    .accessibilityHidden(true)
                Spacer(minLength: 0)
                EvidenceSourceTitle(title: member.source)
                    .frame(width: EvidenceLayout.sourceColumn, alignment: .leading)
                    .accessibilityHidden(true)
                Color.clear.frame(width: EvidenceLayout.badgeColumn, height: 0)
            }
            .padding(.vertical, PVSpacing.space2)
        }
        .padding(.leading, EvidenceLayout.memberIndent)
        .padding(.trailing, EvidenceLayout.inset)
    }
}

// MARK: - Relationships

/// A connection this Subject and the record's members share, worded by the engine's reading of the link from the neighbor's side, with the neighbor's lines under it.
struct EvidenceRelationshipCard: View {
    let relationship: EvidenceBoard.RelationshipCard
    let pins: Set<String>
    let expanded: Set<String>
    let projectDir: String
    let env: EvidenceEnv

    var body: some View {
        PVCard(cornerRadius: PVRadius.sm) {
            VStack(alignment: .leading, spacing: 0) {
                header
                if relationship.groups.isEmpty {
                    Text(L10n.Promote.nothingUnder)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .italic()
                        .foregroundStyle(PVColor.textMuted)
                        .padding(.leading, EvidenceLayout.inset + EvidenceLayout.box + EvidenceLayout.gap)
                        .padding(.trailing, EvidenceLayout.inset)
                        .padding(.bottom, PVSpacing.space3)
                } else {
                    ForEach(relationship.groups) { group in
                        PVDivider()
                        EvidenceSectionLines(
                            group: group, style: .group, pins: pins, expanded: expanded, projectDir: projectDir, env: env
                        )
                            .padding(.leading, EvidenceLayout.relationshipIndent)
                        .accessibilityIdentifier("promote.sheet.section.\(group.id)")
                    }
                }
            }
        }
    }

    /// "…they share a spouse (CEV-3B5RT and PER-7QH3D)" as drawn, with the outcome as a badge.
    private var header: some View {
        HStack(alignment: .center, spacing: EvidenceLayout.gap) {
            EvidencePinBox(
                ids: relationship.pinIDs, checked: pins.intersection(relationship.pinIDs), spoken: relationship.spoken,
                identifier: "promote.sheet.relationship.\(relationship.neighborSubjectID).\(relationship.key)", env: env
            )
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
                Text(verbatim: L10n.Promote.cardHeaderLead(relationship.header))
                    .font(PVFont.display(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textDisplay)
                if relationship.recordCount > 1 {
                    Text(verbatim: L10n.Promote.recordCount(relationship.recordCount))
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
