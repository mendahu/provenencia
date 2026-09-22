import SwiftUI

/// Subordinate bridge / relationship card on the Evidence graph (S6-04 / S7-09).
///
/// Paint-only: same AppKit pointer ownership as ``EvidenceSubjectCard``.
struct EvidenceBridgeCard: View {
    static let width: CGFloat = 212
    static let approximateHalfHeight: CGFloat = 40
    /// Edge hit-testing height from the same top as layout — near a one-line
    /// honesty body so bottoms are not stranded below the fill.
    static let edgeLayoutHeight: CGFloat = 96

    static let editActionID = "edit"
    static let deleteActionID = "delete"

    let placed: SourceGraphPlacedBridge
    var isSelected: Bool
    var isActivated: Bool
    /// Live document-space drag offset from AppKit pointer ownership.
    var dragOffset: CGSize = .zero
    /// Nested action id under the pointer (idle hover), if any.
    var hoveredActionID: String? = nil

    private var isDragging: Bool {
        dragOffset != .zero
    }

    var body: some View {
        EvidenceBridgeCardChrome(
            placed: placed,
            isSelected: isSelected,
            isActivated: isActivated,
            isDragging: isDragging,
            hoveredActionID: hoveredActionID
        )
        .offset(dragOffset)
        .zIndex(isDragging || isActivated ? 1 : 0)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: Self.accessibilityLabel(for: placed)))
        .accessibilityAddTraits((isSelected || isActivated) ? .isSelected : [])
        .accessibilityIdentifier("evidenceGraph.bridge.\(placed.id)")
    }

    static func contentCenter(gridX: Int64, gridY: Int64) -> CGPoint {
        GraphCanvasGridMapping.contentPoint(gridX: gridX, gridY: gridY)
    }

    static func topLeadingOffset(gridX: Int64, gridY: Int64) -> CGSize {
        let center = contentCenter(gridX: gridX, gridY: gridY)
        return CGSize(
            width: center.x - width / 2,
            height: center.y - approximateHalfHeight
        )
    }

    /// Document-space frame used for edge attachment and AppKit hit targets.
    static func contentFrame(gridX: Int64, gridY: Int64, dragOffset: CGSize = .zero) -> CGRect {
        let center = contentCenter(gridX: gridX, gridY: gridY)
        return CGRect(
            x: center.x - width / 2 + dragOffset.width,
            y: center.y - approximateHalfHeight + dragOffset.height,
            width: width,
            height: edgeLayoutHeight
        )
    }

    static func actionTargets(
        for placed: SourceGraphPlacedBridge,
        dragOffset: CGSize = .zero
    ) -> [GraphCanvasActionTarget] {
        let frame = contentFrame(gridX: placed.gridX, gridY: placed.gridY, dragOffset: dragOffset)
        var actions: [GraphCanvasActionTarget] = [
            GraphCanvasActionTarget(
                id: editActionID,
                frame: CGRect(
                    x: frame.maxX - (placed.isCited ? 40 : 68),
                    y: frame.minY + 8,
                    width: 28,
                    height: 28
                )
            ),
        ]
        if !placed.isCited {
            actions.append(
                GraphCanvasActionTarget(
                    id: deleteActionID,
                    frame: CGRect(
                        x: frame.maxX - 36,
                        y: frame.minY + 8,
                        width: 28,
                        height: 28
                    )
                )
            )
        }
        return actions
    }

    static func accessibilityLabel(for placed: SourceGraphPlacedBridge) -> String {
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        let name = trimmed.isEmpty ? placed.typeLabel : trimmed
        let honesty = String(localized: L10n.EvidenceGraph.bridgeHonestyAccessibility)
        let ref = placed.subject.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        if ref.isEmpty {
            return "\(placed.typeLabel), \(name), \(honesty)"
        }
        return "\(placed.typeLabel), \(ref), \(name), \(honesty)"
    }
}

private struct EvidenceBridgeCardChrome: View {
    let placed: SourceGraphPlacedBridge
    var isSelected: Bool
    var isActivated: Bool
    var isDragging: Bool
    var hoveredActionID: String?

    private var showsSelectionChrome: Bool {
        isSelected || isActivated || isDragging
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(alignment: .top, spacing: 8) {
                iconChip
                VStack(alignment: .leading, spacing: 2) {
                    Text(verbatim: displayLabel)
                        .font(PVFont.display(size: 14, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textSecondary)
                        .lineLimit(2)
                        .multilineTextAlignment(.leading)
                    Text(verbatim: typeLine)
                        .font(PVFont.mono(size: 10, weight: PVFontWeight.medium))
                        .tracking(1)
                        .textCase(.uppercase)
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                HStack(spacing: 6) {
                    PVIcon(.penLine, size: 12)
                        .foregroundStyle(PVColor.textMuted)
                    if !placed.isCited {
                        PVIcon(.trash, size: 12)
                            .foregroundStyle(
                                hoveredActionID == EvidenceBridgeCard.deleteActionID
                                    ? PVColor.danger
                                    : PVColor.textMuted
                            )
                    }
                }
            }
            Text(L10n.EvidenceGraph.bridgeHonestyBody)
                .font(PVFont.body(size: PVTypeScale.caption))
                .italic()
                .foregroundStyle(PVColor.textFaint)
                .lineLimit(3)
                .multilineTextAlignment(.leading)
        }
        .padding(.horizontal, 11)
        .padding(.vertical, 10)
        .frame(width: EvidenceBridgeCard.width, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .fill(PVColor.surfaceCard)
        )
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

    private var typeLine: String {
        let ref = placed.subject.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        if ref.isEmpty { return placed.typeLabel }
        return "\(placed.typeLabel) · \(ref)"
    }

    private var cardBorder: some View {
        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
            .strokeBorder(
                showsSelectionChrome ? PVColor.accent : PVColor.borderDefault,
                lineWidth: showsSelectionChrome ? 1.5 : 1
            )
    }

    @ViewBuilder
    private var selectionHalo: some View {
        if showsSelectionChrome {
            RoundedRectangle(cornerRadius: PVRadius.md + 2, style: .continuous)
                .fill(PVColor.graphRing)
                .padding(-3)
        }
    }

    private var iconChip: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 4, style: .continuous)
                .fill(PVColor.surfacePage)
            RoundedRectangle(cornerRadius: 4, style: .continuous)
                .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
            PVMark(placed.kind.markKey, size: 12)
                .foregroundStyle(PVColor.textMuted)
        }
        .frame(width: 22, height: 22)
    }
}
