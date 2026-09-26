import SwiftUI

/// Paint tokens for a cited Observation row on a primary or bridge card.
struct EvidenceCitedPropertyRowStyle {
    var tint: Color
    var chip: Color
    var ink: Color
    var leadingInset: CGFloat
    var trailingInset: CGFloat

    static func primary(_ style: EvidenceSubjectKindStyle, hasConflict: Bool) -> EvidenceCitedPropertyRowStyle {
        EvidenceCitedPropertyRowStyle(
            tint: style.tint,
            chip: style.chip,
            ink: style.ink,
            leadingInset: hasConflict
                ? EvidenceCitedPropertyMarks.conflictGutterWidth
                : EvidenceSubjectCard.shellPaddingX,
            trailingInset: EvidenceSubjectCard.shellPaddingX
        )
    }

    static func bridge(hasConflict: Bool) -> EvidenceCitedPropertyRowStyle {
        EvidenceCitedPropertyRowStyle(
            tint: PVColor.surfaceCard,
            chip: PVColor.surfaceHover,
            ink: PVColor.textMuted,
            leadingInset: hasConflict
                ? EvidenceCitedPropertyMarks.conflictGutterWidth
                : EvidenceBridgeCard.shellPaddingX,
            trailingInset: EvidenceBridgeCard.shellPaddingX
        )
    }
}

/// Compact Observation summary row on a cited graph card (feature snowflake).
///
/// Board: ruled tint row; hover lifts to chip + 2px kind-ink inset. Conflict
/// is an ochre gutter bracket on the stack; negation is a "Not" prefix.
struct EvidenceCitedPropertyRow: View {
    let observation: CatalogObservation
    let style: EvidenceCitedPropertyRowStyle
    var conflictCount: Int? = nil
    var isHovered: Bool = false

    var body: some View {
        HStack(alignment: .center, spacing: 8) {
            VStack(alignment: .leading, spacing: 3) {
                Text(verbatim: propertyLabel)
                    .font(PVFont.mono(size: 9))
                    .tracking(0.7)
                    .textCase(.uppercase)
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                valueLine
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            PVIcon(.penLine, size: 12)
                .foregroundStyle(PVColor.textFaint)
        }
        .padding(.leading, style.leadingInset)
        .padding(.trailing, style.trailingInset)
        .padding(.vertical, 8)
        .frame(maxWidth: .infinity, minHeight: EvidenceSubjectCard.propertyRowHeight, alignment: .leading)
        .background(isHovered ? style.chip : style.tint)
        .overlay(alignment: .leading) {
            if isHovered {
                Rectangle()
                    .fill(style.ink)
                    .frame(width: 2)
            }
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel(Text(verbatim: spokenLabel))
    }

    private var valueLine: some View {
        HStack(alignment: .firstTextBaseline, spacing: 4) {
            if isNegative {
                Text(L10n.EvidenceGraph.negatedPrefix)
                    .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                    .tracking(PVTypeScale.micro * PVTracking.caps)
                    .textCase(.uppercase)
                    .foregroundStyle(PVColor.danger)
            }
            Text(verbatim: valueSummary)
                .font(PVFont.body(size: 13))
                .italic(isNegative)
                .foregroundStyle(isNegative ? PVColor.danger : PVColor.textPrimary)
                .lineLimit(2)
                .multilineTextAlignment(.leading)
        }
    }

    private var propertyLabel: String {
        let trimmed = observation.propertyLabel.trimmingCharacters(in: .whitespacesAndNewlines)
        if !trimmed.isEmpty { return trimmed }
        return observation.propertyKey
    }

    private var valueSummary: String {
        let rendered = ObservationValueDisplay.string(for: observation)
        if rendered.isEmpty {
            return String(localized: L10n.EvidenceGraph.citedValueUnavailable)
        }
        return rendered
    }

    private var spokenValue: String {
        if isNegative {
            return "\(String(localized: L10n.EvidenceGraph.negatedPrefix)) \(valueSummary)"
        }
        return valueSummary
    }

    private var spokenLabel: String {
        EvidenceCitedPropertyMarks.accessibilityLabel(
            propertyLabel: propertyLabel,
            spokenValue: spokenValue,
            conflictCount: conflictCount,
            isNegated: isNegative
        )
    }

    private var isNegative: Bool {
        EvidenceCitedPropertyMarks.isNegated(observation)
    }
}

/// Ruled cited-row stack plus ochre brackets for competing `propertyKey` runs.
struct EvidenceCitedPropertyStack<Footer: View>: View {
    let observations: [CatalogObservation]
    let style: EvidenceSubjectKindStyle?
    let hoveredActionID: String?
    @ViewBuilder var footer: () -> Footer

    private var counts: [String: Int] {
        EvidenceCitedPropertyMarks.conflictCounts(in: observations)
    }

    private var hasConflict: Bool {
        counts.values.contains { $0 >= 2 }
    }

    private var rowStyle: EvidenceCitedPropertyRowStyle {
        if let style {
            return .primary(style, hasConflict: hasConflict)
        }
        return .bridge(hasConflict: hasConflict)
    }

    var body: some View {
        ZStack(alignment: .topLeading) {
            VStack(spacing: EvidenceSubjectCard.stackHairline) {
                ForEach(observations) { observation in
                    EvidenceCitedPropertyRow(
                        observation: observation,
                        style: rowStyle,
                        conflictCount: conflictCount(for: observation),
                        isHovered: hoveredActionID
                            == EvidenceSubjectCard.editPropertyActionID(observationID: observation.id)
                    )
                }
                footer()
            }
            ForEach(Array(EvidenceCitedPropertyMarks.conflictRuns(in: observations).enumerated()), id: \.offset) { _, run in
                EvidenceConflictBracket(height: bracketHeight(for: run))
                    .offset(y: bracketOriginY(for: run))
                    .accessibilityHidden(true)
            }
        }
    }

    private func conflictCount(for observation: CatalogObservation) -> Int? {
        let count = counts[observation.propertyKey, default: 0]
        return count >= 2 ? count : nil
    }

    private func bracketOriginY(for run: Range<Int>) -> CGFloat {
        EvidenceSubjectCard.stackHairline
            + CGFloat(run.lowerBound) * (EvidenceSubjectCard.propertyRowHeight + EvidenceSubjectCard.stackHairline)
    }

    private func bracketHeight(for run: Range<Int>) -> CGFloat {
        let rows = CGFloat(run.count)
        return rows * EvidenceSubjectCard.propertyRowHeight
            + max(rows - 1, 0) * EvidenceSubjectCard.stackHairline
    }
}

/// Ochre `[` in the left gutter joining a competing Property run.
private struct EvidenceConflictBracket: View {
    var height: CGFloat

    var body: some View {
        Path { path in
            let x: CGFloat = 5
            let tick: CGFloat = 4
            let top: CGFloat = 6
            let bottom = max(height - 6, top + 2)
            path.move(to: CGPoint(x: x + tick, y: top))
            path.addLine(to: CGPoint(x: x, y: top))
            path.addLine(to: CGPoint(x: x, y: bottom))
            path.addLine(to: CGPoint(x: x + tick, y: bottom))
        }
        .stroke(
            PVColor.warning,
            style: StrokeStyle(lineWidth: 1.5, lineCap: .round, lineJoin: .round)
        )
        .frame(width: EvidenceCitedPropertyMarks.conflictGutterWidth, height: height, alignment: .leading)
        .allowsHitTesting(false)
    }
}
