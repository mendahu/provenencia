import SwiftUI

/// Primary subject card on the Evidence graph (S6-02 / S7-09).
///
/// Paint-only: AppKit ``GraphCanvasPointerController`` owns select / drag /
/// nested action hits (edit, Add property). Accessibility remains the keyboard path.
struct EvidenceSubjectCard: View {
    /// Must match the name on the Evidence graph document `ZStack`.
    static let documentCoordinateSpace = "evidenceGraphDocument"

    /// Fixed card width — keep in sync with content centering / offset.
    static let width: CGFloat = 264
    /// Top of the card sits this far above the grid center (layout + edges share it).
    static let approximateHalfHeight: CGFloat = 36
    /// Minimum edge hit-testing height for a header-only shell.
    static let edgeLayoutHeight: CGFloat = 88

    static let editActionID = "edit"
    static let addPropertyActionID = "addProperty"

    /// Document-space frame used for edge attachment and AppKit hit targets.
    static func edgeFrame(
        for placed: SourceGraphPlacedSubject,
        dragOffset: CGSize = .zero
    ) -> CGRect {
        let center = contentCenter(gridX: placed.gridX, gridY: placed.gridY)
        let height = contentHeight(for: placed)
        return CGRect(
            x: center.x - width / 2 + dragOffset.width,
            y: center.y - approximateHalfHeight + dragOffset.height,
            width: width,
            height: height
        )
    }

    /// Nested action frames in document space (same coordinate as ``edgeFrame``).
    static func actionTargets(
        for placed: SourceGraphPlacedSubject,
        canCite: Bool,
        dragOffset: CGSize = .zero
    ) -> [GraphCanvasActionTarget] {
        let frame = edgeFrame(for: placed, dragOffset: dragOffset)
        var actions: [GraphCanvasActionTarget] = [
            GraphCanvasActionTarget(
                id: editActionID,
                frame: CGRect(
                    x: frame.maxX - 52,
                    y: frame.minY + 8,
                    width: 40,
                    height: 28
                )
            ),
        ]
        if canCite {
            actions.append(
                GraphCanvasActionTarget(
                    id: addPropertyActionID,
                    frame: CGRect(
                        x: frame.minX,
                        y: frame.maxY - 34,
                        width: frame.width,
                        height: 30
                    )
                )
            )
        }
        return actions
    }

    /// Approximate painted height so edges / hits track cited-row growth.
    static func contentHeight(for placed: SourceGraphPlacedSubject) -> CGFloat {
        var height: CGFloat = 24 + 28 // vertical padding + header
        if placed.isCited {
            height += 10 // divider spacing
            height += CGFloat(max(placed.observations.count, 1)) * 34
        } else if !placed.subject.description.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            height += 10 + 40
        }
        height += 10 + 26 // Add property row
        return max(edgeLayoutHeight, height)
    }

    let placed: SourceGraphPlacedSubject
    /// Registry presentation when available (S7-09).
    var presentation: CatalogSubjectTypePresentation?
    /// Current target — click, rotor, and VoiceOver share this look.
    var isSelected: Bool
    /// Space/Return opened the card for inner controls.
    var isActivated: Bool
    /// Connect tool: this card is origin A waiting for B.
    var isConnectingFrom: Bool = false
    /// When false, Add property paints disabled (No-Artifact gate).
    var canCite: Bool = true
    /// Live document-space drag offset from AppKit pointer ownership.
    var dragOffset: CGSize = .zero

    private var isDragging: Bool {
        dragOffset != .zero
    }

    var body: some View {
        EvidenceSubjectCardChrome(
            placed: placed,
            presentation: presentation,
            isSelected: isSelected || isConnectingFrom,
            isActivated: isActivated,
            isDragging: isDragging,
            isConnectingFrom: isConnectingFrom,
            canCite: canCite
        )
        .opacity(ghostOpacity)
        .offset(dragOffset)
        .zIndex(isDragging || isActivated ? 1 : 0)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: Self.accessibilityLabel(for: placed)))
        .accessibilityAddTraits((isSelected || isActivated) ? .isSelected : [])
        .accessibilityIdentifier("evidenceGraph.subject.\(placed.id)")
    }

    /// Non-interactive placement preview while a tool is armed.
    static func ghost(
        placed: SourceGraphPlacedSubject,
        presentation: CatalogSubjectTypePresentation? = nil
    ) -> some View {
        EvidenceSubjectCardChrome(
            placed: placed,
            presentation: presentation,
            isSelected: false,
            isActivated: false,
            isDragging: false,
            isConnectingFrom: false,
            canCite: true
        )
        .opacity(0.62)
        .accessibilityHidden(true)
    }

    /// Content-space center for a placed (or ghost) card.
    static func contentCenter(gridX: Int64, gridY: Int64) -> CGPoint {
        GraphCanvasGridMapping.contentPoint(gridX: gridX, gridY: gridY)
    }

    /// Top-leading offset so the card keeps a real layout frame.
    static func topLeadingOffset(gridX: Int64, gridY: Int64) -> CGSize {
        let center = contentCenter(gridX: gridX, gridY: gridY)
        return CGSize(
            width: center.x - width / 2,
            height: center.y - approximateHalfHeight
        )
    }

    /// VoiceOver / rotor label for a placed subject.
    static func accessibilityLabel(for placed: SourceGraphPlacedSubject) -> String {
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        let name = trimmed.isEmpty ? placed.typeLabel : trimmed
        let citation = placed.isCited
            ? String(localized: L10n.EvidenceGraph.citedAccessibility)
            : String(localized: L10n.EvidenceGraph.uncitedAccessibility)
        let ref = placed.subject.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        if ref.isEmpty {
            return "\(placed.typeLabel), \(name), \(citation)"
        }
        return "\(placed.typeLabel), \(ref), \(name), \(citation)"
    }

    private var ghostOpacity: Double {
        if isDragging { return 0.92 }
        return placed.isCited ? 1 : 0.76
    }
}

/// Visual shell shared by live cards and the placement ghost.
private struct EvidenceSubjectCardChrome: View {
    let placed: SourceGraphPlacedSubject
    var presentation: CatalogSubjectTypePresentation?
    var isSelected: Bool
    var isActivated: Bool
    var isDragging: Bool
    var isConnectingFrom: Bool
    var canCite: Bool

    private var style: EvidenceSubjectKindStyle {
        EvidenceSubjectKindStyle.resolve(typeKey: placed.kind.rawValue, presentation: presentation)
    }

    private var showsSelectionChrome: Bool {
        isSelected || isActivated || isDragging || isConnectingFrom
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 9) {
            headerRow
            if isConnectingFrom {
                Text(L10n.EvidenceGraph.connectingFrom)
                    .font(PVFont.mono(size: 11))
                    .foregroundStyle(PVColor.accent)
            } else if placed.isCited {
                Rectangle()
                    .fill(style.line.opacity(0.55))
                    .frame(height: 1)
                ForEach(placed.observations) { observation in
                    EvidenceCitedPropertyRow(observation: observation)
                }
            } else if let description = nonEmptyDescription {
                Text(verbatim: description)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(3)
                    .multilineTextAlignment(.leading)
            }
            addPropertyRow
        }
        .padding(.horizontal, 13)
        .padding(.vertical, 12)
        .frame(width: EvidenceSubjectCard.width, alignment: .leading)
        .background(cardBackground)
        .overlay(cardBorder)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .shadow(
            color: showsSelectionChrome && !isSelected
                ? Color.black.opacity(0.1)
                : .clear,
            radius: isDragging ? 10 : 6,
            y: isDragging ? 4 : 2
        )
        .background(selectionHalo)
    }

    private var headerRow: some View {
        HStack(alignment: .top, spacing: 10) {
            iconChip
            VStack(alignment: .leading, spacing: 2) {
                Text(verbatim: displayLabel)
                    .font(PVFont.display(size: 16, weight: PVFontWeight.medium))
                    .foregroundStyle(placed.isCited ? PVColor.textDisplay : PVColor.textSecondary)
                    .lineLimit(2)
                    .multilineTextAlignment(.leading)
                Text(verbatim: typeLine)
                    .font(PVFont.mono(size: 10, weight: PVFontWeight.medium))
                    .tracking(1)
                    .textCase(.uppercase)
                    .foregroundStyle(style.ink)
                    .lineLimit(1)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            HStack(spacing: 6) {
                Image(systemName: "pencil")
                    .font(.system(size: 11, weight: .medium))
                    .foregroundStyle(PVColor.textMuted)
                    .accessibilityHidden(true)
                citationMark
            }
        }
    }

    private var addPropertyRow: some View {
        HStack(spacing: 6) {
            Image(systemName: "plus")
                .font(.system(size: 11, weight: .medium))
            Text(L10n.EvidenceGraph.addProperty)
                .font(PVFont.body(size: 12))
        }
        .foregroundStyle(canCite ? PVColor.textMuted : PVColor.textFaint)
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.top, 2)
        .accessibilityHidden(true)
    }

    private var displayLabel: String {
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? placed.typeLabel : trimmed
    }

    private var typeLine: String {
        let ref = placed.subject.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        if ref.isEmpty { return placed.typeLabel }
        return "\(placed.typeLabel) · \(ref)"
    }

    private var nonEmptyDescription: String? {
        let trimmed = placed.subject.description.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? nil : trimmed
    }

    private var iconChip: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 4, style: .continuous)
                .fill(style.chip)
            RoundedRectangle(cornerRadius: 4, style: .continuous)
                .strokeBorder(
                    placed.isCited ? style.line : PVColor.borderDefault,
                    style: StrokeStyle(
                        lineWidth: 1,
                        dash: placed.isCited ? [] : [3, 2]
                    )
                )
            PVMark(placed.kind.markKey, size: 15)
                .foregroundStyle(style.ink)
        }
        .frame(width: 28, height: 28)
    }

    @ViewBuilder
    private var citationMark: some View {
        if placed.isCited {
            Image(systemName: "checkmark")
                .font(.system(size: 11, weight: .semibold))
                .foregroundStyle(style.ink)
                .accessibilityHidden(true)
        } else {
            Image(systemName: "circle.dashed")
                .font(.system(size: 13, weight: .regular))
                .foregroundStyle(PVColor.evidenceUndocumented)
                .accessibilityHidden(true)
        }
    }

    private var cardBackground: some View {
        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
            .fill(PVColor.surfaceCard)
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .fill(style.tint.opacity(placed.isCited ? 0.45 : 0.38))
            )
    }

    private var cardBorder: some View {
        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
            .strokeBorder(
                borderColor,
                style: StrokeStyle(
                    lineWidth: showsSelectionChrome ? 1.5 : 1,
                    dash: (placed.isCited || showsSelectionChrome) ? [] : [4, 3]
                )
            )
    }

    private var borderColor: Color {
        if showsSelectionChrome { return PVColor.accent }
        return placed.isCited ? style.line : PVColor.borderDefault
    }

    @ViewBuilder
    private var selectionHalo: some View {
        if showsSelectionChrome {
            RoundedRectangle(cornerRadius: PVRadius.md + 2, style: .continuous)
                .fill(PVColor.graphRing)
                .padding(-3)
        }
    }
}

/// Compact Observation summary row on a cited primary card (feature snowflake).
private struct EvidenceCitedPropertyRow: View {
    let observation: CatalogObservation

    var body: some View {
        VStack(alignment: .leading, spacing: 1) {
            Text(verbatim: propertyLabel)
                .font(PVFont.mono(size: 9, weight: PVFontWeight.medium))
                .tracking(0.7)
                .textCase(.uppercase)
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(1)
            Text(verbatim: valueSummary)
                .font(PVFont.body(size: 13))
                .foregroundStyle(isNegative ? PVColor.textMuted : PVColor.textPrimary)
                .lineLimit(2)
                .multilineTextAlignment(.leading)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .accessibilityElement(children: .combine)
        .accessibilityLabel(Text(verbatim: "\(propertyLabel), \(valueSummary)"))
    }

    private var propertyLabel: String {
        let trimmed = observation.propertyLabel.trimmingCharacters(in: .whitespacesAndNewlines)
        if !trimmed.isEmpty { return trimmed }
        return observation.propertyKey
    }

    private var valueSummary: String {
        let text = observation.valueText.trimmingCharacters(in: .whitespacesAndNewlines)
        if !text.isEmpty {
            return isNegative ? "¬ \(text)" : text
        }
        if let value = observation.valueInteger {
            let rendered = "\(value)"
            return isNegative ? "¬ \(rendered)" : rendered
        }
        return String(localized: L10n.EvidenceGraph.citedValueUnavailable)
    }

    private var isNegative: Bool {
        observation.polarity.trimmingCharacters(in: .whitespacesAndNewlines).lowercased() == "negative"
    }
}
