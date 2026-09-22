import SwiftUI

/// Subordinate bridge / relationship card on the Evidence graph (S6-04 / S7-09).
///
/// Paint-only: same AppKit pointer ownership as ``EvidenceSubjectCard``.
/// Layout follows the Evidence graph board: provisional cards stack working
/// label over type with an honesty body; durable (cited) cards put type·ref
/// on the chrome row and the full edge-summary sentence as the title below.
struct EvidenceBridgeCard: View {
    static let width: CGFloat = 236
    static let approximateHalfHeight: CGFloat = 36
    /// Edge hit-testing height from the same top as layout — near a one-line
    /// body so bottoms are not stranded below the fill.
    static let edgeLayoutHeight: CGFloat = 88

    static let editActionID = "edit"
    static let deleteActionID = "delete"

    /// Horizontal / top padding on the bridge shell (matches chrome).
    static let shellPaddingX: CGFloat = 11
    static let shellPaddingTop: CGFloat = 9
    static let headerActionHitHeight: CGFloat = 28

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
        let headerHits = EvidenceCardHeaderActionHits.frames(
            cardFrame: frame,
            paddingX: shellPaddingX,
            paddingTop: shellPaddingTop,
            hitHeight: headerActionHitHeight,
            showDelete: !placed.isCited
        )
        var actions: [GraphCanvasActionTarget] = [
            GraphCanvasActionTarget(id: editActionID, frame: headerHits.edit),
        ]
        if let deleteFrame = headerHits.delete {
            actions.append(GraphCanvasActionTarget(id: deleteActionID, frame: deleteFrame))
        }
        return actions
    }

    static func accessibilityLabel(for placed: SourceGraphPlacedBridge) -> String {
        let ref = placed.subject.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        if placed.isCited {
            let summary = EvidenceBridgeEdgeSummary.phrase(for: placed)
            if ref.isEmpty {
                return "\(placed.typeLabel), \(summary)"
            }
            return "\(placed.typeLabel), \(ref), \(summary)"
        }
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        let name = trimmed.isEmpty ? placed.typeLabel : trimmed
        let honesty = String(localized: L10n.EvidenceGraph.bridgeHonestyAccessibility)
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

    private var workingLabel: String {
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? placed.typeLabel : trimmed
    }

    private var typeRefLine: String {
        let ref = placed.subject.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        if ref.isEmpty { return placed.typeLabel }
        return "\(placed.typeLabel) · \(ref)"
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 7) {
            headerRow
            if placed.isCited {
                Text(verbatim: EvidenceBridgeEdgeSummary.phrase(for: placed))
                    .font(PVFont.display(size: 14.5, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(3)
                    .multilineTextAlignment(.leading)
                    .fixedSize(horizontal: false, vertical: true)
            } else {
                Text(L10n.EvidenceGraph.bridgeHonestyBody)
                    .font(PVFont.body(size: 11.5, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(3)
                    .multilineTextAlignment(.leading)
            }
        }
        .padding(.horizontal, EvidenceBridgeCard.shellPaddingX)
        .padding(.vertical, EvidenceBridgeCard.shellPaddingTop)
        .frame(width: EvidenceBridgeCard.width, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .fill(PVColor.surfaceCard)
        )
        .overlay(cardBorder)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .shadow(
            color: Color.black.opacity(showsSelectionChrome || placed.isCited ? 0.08 : 0.06),
            radius: isDragging ? 10 : 4,
            y: isDragging ? 4 : 1
        )
        .background(selectionHalo)
    }

    private var headerRow: some View {
        HStack(alignment: placed.isCited ? .center : .top, spacing: 8) {
            iconChip
            if placed.isCited {
                Text(verbatim: typeRefLine)
                    .font(PVFont.mono(size: 9.5, weight: PVFontWeight.regular))
                    .tracking(0.6)
                    .textCase(.uppercase)
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                    .frame(maxWidth: .infinity, alignment: .leading)
            } else {
                VStack(alignment: .leading, spacing: 1) {
                    Text(verbatim: workingLabel)
                        .font(PVFont.display(size: 14, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(2)
                        .multilineTextAlignment(.leading)
                    Text(verbatim: placed.typeLabel)
                        .font(PVFont.mono(size: 9.5, weight: PVFontWeight.regular))
                        .tracking(1)
                        .textCase(.uppercase)
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
            }
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
            .alignmentGuide(.top) { d in d[.top] }
        }
    }

    private var cardBorder: some View {
        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
            .strokeBorder(
                showsSelectionChrome ? PVColor.accent : PVColor.borderSubtle,
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
            RoundedRectangle(cornerRadius: 3, style: .continuous)
                .fill(PVColor.surfaceSunken)
            RoundedRectangle(cornerRadius: 3, style: .continuous)
                .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
            PVMark(placed.kind.markKey, size: 12)
                .foregroundStyle(PVColor.textMuted)
        }
        .frame(width: 22, height: 22)
    }
}
