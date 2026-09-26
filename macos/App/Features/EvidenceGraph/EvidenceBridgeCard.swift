import AppKit
import SwiftUI

/// Subordinate bridge / relationship card on the Evidence graph (S6-04 / S7-09).
///
/// Paint-only: same AppKit pointer ownership as ``EvidenceSubjectCard``.
/// Layout follows the Evidence graph board: provisional cards stack working
/// label over type with an honesty body; durable (cited) cards put type·ref
/// on the chrome row and the full edge-summary sentence as the title below.
struct EvidenceBridgeCard: View {
    nonisolated static let width: CGFloat = 236
    nonisolated static let approximateHalfHeight: CGFloat = 36

    static let editActionID = "edit"
    static let editCitationActionID = "editCitation"
    static let deleteActionID = "delete"

    /// Horizontal / top padding on the bridge shell (matches chrome).
    nonisolated static let shellPaddingX: CGFloat = 11
    static let shellPaddingTop: CGFloat = 9
    static let headerActionHitHeight: CGFloat = 28
    /// Painted header row (type·ref / working label) before the body sentence.
    static let headerContentHeightCited: CGFloat = 22
    static let headerContentHeightUncited: CGFloat = 26
    static let headerToBodySpacing: CGFloat = 7
    static let bodyActionHitHeight: CGFloat = 28

    let placed: SourceGraphPlacedBridge
    let snapshot: SourceGraphSnapshot
    var isSelected: Bool
    var isActivated: Bool
    /// Live document-space drag offset from AppKit pointer ownership.
    var dragOffset: CGSize = .zero
    /// Nested action id under the pointer (idle hover), if any.
    var hoveredActionID: String? = nil
    /// When false (no Artifact), the citation pencil is omitted.
    var canCite: Bool = false

    private var isDragging: Bool {
        dragOffset != .zero
    }

    var body: some View {
        EvidenceBridgeCardChrome(
            placed: placed,
            snapshot: snapshot,
            isSelected: isSelected,
            isActivated: isActivated,
            isDragging: isDragging,
            hoveredActionID: hoveredActionID,
            canCite: canCite
        )
        .offset(dragOffset)
        .zIndex(isDragging || isActivated ? 1 : 0)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: Self.accessibilityLabel(for: placed, in: snapshot)))
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
    /// Height follows the wrapped body so lines and hits stay on the painted card.
    static func contentFrame(
        for placed: SourceGraphPlacedBridge,
        in snapshot: SourceGraphSnapshot,
        dragOffset: CGSize = .zero
    ) -> CGRect {
        let center = contentCenter(gridX: placed.gridX, gridY: placed.gridY)
        return CGRect(
            x: center.x - width / 2 + dragOffset.width,
            y: center.y - approximateHalfHeight + dragOffset.height,
            width: width,
            height: contentHeight(for: placed, in: snapshot)
        )
    }

    /// Painted height: header plus up to three wrapped body lines, then extra
    /// non-edge rows and Add property. No minimum floor — a frame taller than
    /// the fill strands bottom-approach edge terminals below the card.
    static func contentHeight(
        for placed: SourceGraphPlacedBridge,
        in snapshot: SourceGraphSnapshot
    ) -> CGFloat {
        let header = placed.isCited ? headerContentHeightCited : headerContentHeightUncited
        let body = bodyTextHeight(for: placed, in: snapshot)
        var height = shellPaddingTop + header + headerToBodySpacing + body
        if placed.isCited {
            let extras = CGFloat(extraObservations(for: placed).count)
            height += EvidenceSubjectCard.stackHairline
                + extras * (EvidenceSubjectCard.propertyRowHeight + EvidenceSubjectCard.stackHairline)
                + EvidenceSubjectCard.addPropertyInStackHeight
        } else {
            height += 26 + shellPaddingTop
        }
        return height
    }

    static func extraObservations(for placed: SourceGraphPlacedBridge) -> [CatalogObservation] {
        EvidenceCitedPropertyMarks.extraObservations(in: placed.observations)
    }

    static func bodyTextHeight(
        for placed: SourceGraphPlacedBridge,
        in snapshot: SourceGraphSnapshot
    ) -> CGFloat {
        let text = placed.isCited
            ? EvidenceBridgeEdgeSummary.sentence(for: placed, in: snapshot)
            : String(localized: L10n.EvidenceGraph.bridgeHonestyBody)
        let font = placed.isCited
            ? PVFont.nsDisplay(size: 14.5, weight: PVFontWeight.medium)
            : PVFont.nsBody(size: 11.5, weight: PVFontWeight.regular, italic: true)
        return wrappedTextHeight(text, width: bodyTextWidth, font: font, maxLines: 3)
    }

    /// Body copy width inside the shell, leaving room for the trailing icon column.
    private static var bodyTextWidth: CGFloat {
        width - shellPaddingX * 2 - 8 - 12
    }

    private static func wrappedTextHeight(
        _ text: String,
        width: CGFloat,
        font: NSFont,
        maxLines: Int
    ) -> CGFloat {
        let style = NSMutableParagraphStyle()
        style.lineBreakMode = .byWordWrapping
        let rect = (text as NSString).boundingRect(
            with: CGSize(width: max(width, 1), height: .greatestFiniteMagnitude),
            options: [.usesLineFragmentOrigin, .usesFontLeading],
            attributes: [
                .font: font,
                .paragraphStyle: style,
            ]
        )
        let line = max(font.boundingRectForFont.height, 1)
        return min(ceil(rect.height), ceil(line) * CGFloat(maxLines))
    }

    static func actionTargets(
        for placed: SourceGraphPlacedBridge,
        in snapshot: SourceGraphSnapshot,
        canCite: Bool,
        dragOffset: CGSize = .zero
    ) -> [GraphCanvasActionTarget] {
        let frame = contentFrame(for: placed, in: snapshot, dragOffset: dragOffset)
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
        let headerHeight = placed.isCited ? headerContentHeightCited : headerContentHeightUncited
        let bodyTop = frame.minY + shellPaddingTop + headerHeight + headerToBodySpacing
        if canCite {
            actions.append(
                GraphCanvasActionTarget(
                    id: editCitationActionID,
                    frame: CGRect(
                        x: headerHits.edit.minX,
                        y: bodyTop,
                        width: headerHits.edit.width,
                        height: bodyActionHitHeight
                    )
                )
            )
        }
        if placed.isCited {
            let extras = extraObservations(for: placed)
            let stackTop = bodyTop + bodyTextHeight(for: placed, in: snapshot)
            for (index, observation) in extras.enumerated() {
                let y = stackTop
                    + EvidenceSubjectCard.stackHairline
                    + CGFloat(index) * (EvidenceSubjectCard.propertyRowHeight + EvidenceSubjectCard.stackHairline)
                actions.append(
                    GraphCanvasActionTarget(
                        id: EvidenceSubjectCard.editPropertyActionID(observationID: observation.id),
                        frame: CGRect(
                            x: frame.minX,
                            y: y,
                            width: frame.width,
                            height: EvidenceSubjectCard.propertyRowHeight
                        )
                    )
                )
            }
            if canCite {
                let addY = stackTop
                    + EvidenceSubjectCard.stackHairline
                    + CGFloat(extras.count)
                    * (EvidenceSubjectCard.propertyRowHeight + EvidenceSubjectCard.stackHairline)
                actions.append(
                    GraphCanvasActionTarget(
                        id: EvidenceSubjectCard.addPropertyActionID,
                        frame: CGRect(
                            x: frame.minX,
                            y: addY,
                            width: frame.width,
                            height: EvidenceSubjectCard.addPropertyInStackHeight
                        )
                    )
                )
            }
        } else if canCite {
            actions.append(
                GraphCanvasActionTarget(
                    id: EvidenceSubjectCard.addPropertyActionID,
                    frame: CGRect(
                        x: frame.minX,
                        y: frame.maxY - 30 - 4,
                        width: frame.width,
                        height: 30
                    )
                )
            )
        }
        if let deleteFrame = headerHits.delete {
            actions.append(GraphCanvasActionTarget(id: deleteActionID, frame: deleteFrame))
        }
        return actions
    }

    static func accessibilityLabel(
        for placed: SourceGraphPlacedBridge,
        in snapshot: SourceGraphSnapshot
    ) -> String {
        let ref = placed.subject.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        if placed.isCited {
            let summary = EvidenceBridgeEdgeSummary.sentence(for: placed, in: snapshot)
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
    let snapshot: SourceGraphSnapshot
    var isSelected: Bool
    var isActivated: Bool
    var isDragging: Bool
    var hoveredActionID: String?
    var canCite: Bool

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
        VStack(alignment: .leading, spacing: EvidenceBridgeCard.headerToBodySpacing) {
            headerRow
            bodyRow
            if placed.isCited {
                extraPropertyStack
            } else {
                addPropertyQuietRow
            }
        }
        .padding(.horizontal, EvidenceBridgeCard.shellPaddingX)
        .padding(.top, EvidenceBridgeCard.shellPaddingTop)
        .padding(.bottom, placed.isCited ? 0 : EvidenceBridgeCard.shellPaddingTop)
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
                    .foregroundStyle(
                        hoveredActionID == EvidenceBridgeCard.editActionID
                            ? PVColor.textSecondary
                            : PVColor.textMuted
                    )
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

    private var bodyRow: some View {
        HStack(alignment: .top, spacing: 8) {
            if placed.isCited {
                Text(verbatim: EvidenceBridgeEdgeSummary.sentence(for: placed, in: snapshot))
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
            Spacer(minLength: 0)
            HStack(spacing: 6) {
                if canCite {
                    PVIcon(.penLine, size: 12)
                        .foregroundStyle(
                            hoveredActionID == EvidenceBridgeCard.editCitationActionID
                                ? PVColor.accent
                                : PVColor.textMuted
                        )
                        .padding(.top, 2)
                }
                if !placed.isCited {
                    Color.clear.frame(width: 12, height: 12)
                }
            }
        }
    }

    private var extraPropertyStack: some View {
        EvidenceCitedPropertyStack(
            observations: EvidenceBridgeCard.extraObservations(for: placed),
            style: nil,
            hoveredActionID: hoveredActionID
        ) {
            addPropertyStackRow
        }
        .padding(.top, EvidenceSubjectCard.stackHairline)
        .background(PVColor.borderSubtle)
        .clipShape(
            UnevenRoundedRectangle(
                bottomLeadingRadius: PVRadius.md,
                bottomTrailingRadius: PVRadius.md,
                style: .continuous
            )
        )
        .padding(.horizontal, -EvidenceBridgeCard.shellPaddingX)
    }

    private var addPropertyStackRow: some View {
        let hovered = hoveredActionID == EvidenceSubjectCard.addPropertyActionID
        return HStack(spacing: 6) {
            PVIcon(.plus, size: 12)
            Text(L10n.EvidenceGraph.addProperty)
                .font(PVFont.body(size: 12))
        }
        .foregroundStyle(
            canCite
                ? (hovered ? PVColor.textPrimary : PVColor.textMuted)
                : PVColor.textFaint
        )
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, EvidenceBridgeCard.shellPaddingX)
        .padding(.top, 9)
        .padding(.bottom, 11)
        .background(hovered && canCite ? PVColor.surfaceHover : PVColor.surfaceCard)
        .accessibilityHidden(true)
    }

    private var addPropertyQuietRow: some View {
        let hovered = hoveredActionID == EvidenceSubjectCard.addPropertyActionID
        return HStack(spacing: 6) {
            PVIcon(.plus, size: 12)
            Text(L10n.EvidenceGraph.addProperty)
                .font(PVFont.body(size: 12))
        }
        .foregroundStyle(
            canCite
                ? (hovered ? PVColor.textPrimary : PVColor.textMuted)
                : PVColor.textFaint
        )
        .frame(maxWidth: .infinity, alignment: .leading)
        .accessibilityHidden(true)
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
