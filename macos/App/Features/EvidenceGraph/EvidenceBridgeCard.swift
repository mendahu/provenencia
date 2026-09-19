import SwiftUI

/// Subordinate bridge / relationship card on the Evidence graph (S6-04).
///
/// Same focusable (not Button) + drag pattern as ``EvidenceSubjectCard``, but
/// quieter: 188px, no kind wash, honesty body instead of description.
struct EvidenceBridgeCard: View {
    static let width: CGFloat = 188
    static let approximateHalfHeight: CGFloat = 40

    let placed: SourceGraphPlacedBridge
    var isSelected: Bool
    var isActivated: Bool
    var dragEnabled: Bool
    var keyboardFocus: FocusState<String?>.Binding
    var onSelect: () -> Void
    var onActivate: () -> Void
    var onEscape: () -> Void = {}
    var onDragEnded: ((CGSize) -> Void)?

    @GestureState private var dragOffset: CGSize = .zero

    private var isDragging: Bool {
        dragOffset != .zero
    }

    var body: some View {
        EvidenceBridgeCardChrome(
            placed: placed,
            isSelected: isSelected,
            isActivated: isActivated,
            isDragging: isDragging
        )
        .offset(dragOffset)
        .zIndex(isDragging || isActivated ? 1 : 0)
        .contentShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .gesture(dragGesture)
        .onTapGesture(perform: onSelect)
        .background {
            Color.clear
                .focusable()
                .focusEffectDisabled()
                .focused(keyboardFocus, equals: placed.id)
                .onKeyPress(.space) {
                    onActivate()
                    return .handled
                }
                .onKeyPress(.return) {
                    onActivate()
                    return .handled
                }
                .onKeyPress(.escape) {
                    onEscape()
                    return .handled
                }
                .accessibilityHidden(true)
        }
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

    static func contentFrame(gridX: Int64, gridY: Int64, dragOffset: CGSize = .zero) -> CGRect {
        let center = contentCenter(gridX: gridX, gridY: gridY)
        return CGRect(
            x: center.x - width / 2 + dragOffset.width,
            y: center.y - approximateHalfHeight + dragOffset.height,
            width: width,
            height: approximateHalfHeight * 2
        )
    }

    static func accessibilityLabel(for placed: SourceGraphPlacedBridge) -> String {
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        let name = trimmed.isEmpty ? placed.typeLabel : trimmed
        let honesty = String(localized: L10n.EvidenceGraph.bridgeHonestyAccessibility)
        return "\(placed.typeLabel), \(name), \(honesty)"
    }

    private var dragGesture: some Gesture {
        DragGesture(
            minimumDistance: dragEnabled ? 4 : 10_000,
            coordinateSpace: .named(EvidenceSubjectCard.documentCoordinateSpace)
        )
        .updating($dragOffset) { value, state, _ in
            guard dragEnabled else { return }
            state = CGSize(
                width: value.location.x - value.startLocation.x,
                height: value.location.y - value.startLocation.y
            )
        }
        .onEnded { value in
            guard dragEnabled else { return }
            onSelect()
            let delta = CGSize(
                width: value.location.x - value.startLocation.x,
                height: value.location.y - value.startLocation.y
            )
            onDragEnded?(delta)
        }
    }
}

private struct EvidenceBridgeCardChrome: View {
    let placed: SourceGraphPlacedBridge
    var isSelected: Bool
    var isActivated: Bool
    var isDragging: Bool

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
                    Text(verbatim: placed.typeLabel)
                        .font(PVFont.mono(size: 10, weight: PVFontWeight.medium))
                        .tracking(1)
                        .textCase(.uppercase)
                        .foregroundStyle(PVColor.textMuted)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
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
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .strokeBorder(
                    showsSelectionChrome ? PVColor.accent : PVColor.borderSubtle,
                    lineWidth: showsSelectionChrome ? 1.5 : 1
                )
        )
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .shadow(
            color: showsSelectionChrome && !isSelected
                ? Color.black.opacity(0.08)
                : .clear,
            radius: isDragging ? 8 : 4,
            y: isDragging ? 3 : 1
        )
        .background {
            if showsSelectionChrome {
                RoundedRectangle(cornerRadius: PVRadius.md + 2, style: .continuous)
                    .fill(PVColor.graphRing)
                    .padding(-3)
            }
        }
    }

    private var displayLabel: String {
        let trimmed = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? placed.typeLabel : trimmed
    }

    private var iconChip: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 4, style: .continuous)
                .fill(PVColor.surfacePage)
            RoundedRectangle(cornerRadius: 4, style: .continuous)
                .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
            PVSubjectIcon(kind: placed.kind.subjectIconKind, size: 12)
                .foregroundStyle(PVColor.textMuted)
        }
        .frame(width: 22, height: 22)
    }
}
