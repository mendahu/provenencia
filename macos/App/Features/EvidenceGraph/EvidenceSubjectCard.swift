import SwiftUI

/// Primary subject card on the Evidence graph (S6-02 / S6-03).
///
/// Paint-only: AppKit ``GraphCanvasPointerController`` owns select / drag.
/// Accessibility labels + rotors remain the keyboard/VoiceOver path.
struct EvidenceSubjectCard: View {
    /// Must match the name on the Evidence graph document `ZStack`.
    static let documentCoordinateSpace = "evidenceGraphDocument"

    /// Fixed card width — keep in sync with content centering / offset.
    static let width: CGFloat = 236
    /// Top of the card sits this far above the grid center (layout + edges share it).
    static let approximateHalfHeight: CGFloat = 36
    /// Edge hit-testing height from that same top. Sized near a header-only card
    /// so the bottom edge is not stranded below short chrome; taller cards still
    /// cover the stroked tip after ``GraphCanvasEdgeGeometry.endpointTuck``.
    static let edgeLayoutHeight: CGFloat = 64

    /// Document-space frame used for edge attachment and AppKit hit targets.
    static func edgeFrame(gridX: Int64, gridY: Int64, dragOffset: CGSize = .zero) -> CGRect {
        let center = contentCenter(gridX: gridX, gridY: gridY)
        return CGRect(
            x: center.x - width / 2 + dragOffset.width,
            y: center.y - approximateHalfHeight + dragOffset.height,
            width: width,
            height: edgeLayoutHeight
        )
    }

    let placed: SourceGraphPlacedSubject
    /// Current target — click, rotor, and VoiceOver share this look.
    var isSelected: Bool
    /// Space/Return opened the card for inner controls (future actions).
    var isActivated: Bool
    /// Connect tool: this card is origin A waiting for B.
    var isConnectingFrom: Bool = false
    /// Live document-space drag offset from AppKit pointer ownership.
    var dragOffset: CGSize = .zero

    private var isDragging: Bool {
        dragOffset != .zero
    }

    var body: some View {
        EvidenceSubjectCardChrome(
            placed: placed,
            isSelected: isSelected || isConnectingFrom,
            isActivated: isActivated,
            isDragging: isDragging,
            isConnectingFrom: isConnectingFrom
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
    static func ghost(placed: SourceGraphPlacedSubject) -> some View {
        EvidenceSubjectCardChrome(
            placed: placed,
            isSelected: false,
            isActivated: false,
            isDragging: false,
            isConnectingFrom: false
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
        return "\(placed.typeLabel), \(name), \(citation)"
    }

    private var ghostOpacity: Double {
        if isDragging { return 0.92 }
        return placed.isCited ? 1 : 0.76
    }
}

/// Visual shell shared by live cards and the placement ghost.
private struct EvidenceSubjectCardChrome: View {
    let placed: SourceGraphPlacedSubject
    var isSelected: Bool
    var isActivated: Bool
    var isDragging: Bool
    var isConnectingFrom: Bool

    private var style: EvidenceSubjectKindStyle {
        EvidenceSubjectKindStyle.forKind(placed.kind)
    }

    private var showsSelectionChrome: Bool {
        isSelected || isActivated || isDragging || isConnectingFrom
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack(alignment: .top, spacing: 10) {
                iconChip
                VStack(alignment: .leading, spacing: 2) {
                    Text(verbatim: displayLabel)
                        .font(PVFont.display(size: 16, weight: PVFontWeight.medium))
                        .foregroundStyle(placed.isCited ? PVColor.textDisplay : PVColor.textSecondary)
                        .lineLimit(2)
                        .multilineTextAlignment(.leading)
                    Text(verbatim: placed.typeLabel)
                        .font(PVFont.mono(size: 10, weight: PVFontWeight.medium))
                        .tracking(1)
                        .textCase(.uppercase)
                        .foregroundStyle(style.ink)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                citationMark
            }
            if isConnectingFrom {
                Text(L10n.EvidenceGraph.connectingFrom)
                    .font(PVFont.mono(size: 11))
                    .foregroundStyle(PVColor.accent)
            } else if let description = nonEmptyDescription {
                Text(verbatim: description)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(3)
                    .multilineTextAlignment(.leading)
            }
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

    private var displayLabel: String {
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? placed.typeLabel : trimmed
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
