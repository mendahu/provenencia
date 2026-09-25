import SwiftUI

/// Primary subject card on the Evidence graph (S6-02 / S7-09 / board chrome).
///
/// Paint-only: AppKit ``GraphCanvasPointerController`` owns select / drag /
/// nested action hits (edit, delete, property edit, Add property).
/// Accessibility remains the keyboard path.
struct EvidenceSubjectCard: View {
    /// Must match the name on the Evidence graph document `ZStack`.
    static let documentCoordinateSpace = "evidenceGraphDocument"

    /// Fixed card width — keep in sync with content centering / offset.
    static let width: CGFloat = 264
    /// Top of the card sits this far above the grid center (layout + edges share it).
    static let approximateHalfHeight: CGFloat = 36
    /// Minimum edge hit-testing height for a header-only shell.
    static let edgeLayoutHeight: CGFloat = 88

    /// Horizontal / top padding on the card shell (matches board).
    static let shellPaddingX: CGFloat = 13
    static let shellPaddingTop: CGFloat = 11
    static let shellPaddingBottom: CGFloat = 12
    static let headerHeight: CGFloat = 28
    static let headerToBodyGap: CGFloat = 9
    /// Ruled property / Add-property stack (board: 1px kind-line hairlines).
    static let propertyRowHeight: CGFloat = 46
    static let addPropertyInStackHeight: CGFloat = 32
    static let stackHairline: CGFloat = 1

    static let editActionID = "edit"
    static let deleteActionID = "delete"
    static let addPropertyActionID = "addProperty"

    static func editPropertyActionID(observationID: String) -> String {
        "editProperty.\(observationID)"
    }

    static func observationID(fromEditPropertyAction actionID: String) -> String? {
        let prefix = "editProperty."
        guard actionID.hasPrefix(prefix) else { return nil }
        let id = String(actionID.dropFirst(prefix.count))
        return id.isEmpty ? nil : id
    }

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
        let headerHits = EvidenceCardHeaderActionHits.frames(
            cardFrame: frame,
            paddingX: shellPaddingX,
            paddingTop: shellPaddingTop,
            hitHeight: headerHeight,
            showDelete: !placed.isCited
        )
        var actions: [GraphCanvasActionTarget] = [
            GraphCanvasActionTarget(id: editActionID, frame: headerHits.edit),
        ]
        if let deleteFrame = headerHits.delete {
            actions.append(GraphCanvasActionTarget(id: deleteActionID, frame: deleteFrame))
        }

        let stackTop = stackOriginY(for: placed, frame: frame)
        if placed.isCited {
            for (index, observation) in placed.observations.enumerated() {
                let y = stackTop
                    + stackHairline
                    + CGFloat(index) * (propertyRowHeight + stackHairline)
                // Full-row hit so hover/edit match the board (not pencil-only).
                actions.append(
                    GraphCanvasActionTarget(
                        id: editPropertyActionID(observationID: observation.id),
                        frame: CGRect(
                            x: frame.minX,
                            y: y,
                            width: frame.width,
                            height: propertyRowHeight
                        )
                    )
                )
            }
        }

        if canCite {
            let addY: CGFloat
            if placed.isCited {
                let rows = CGFloat(placed.observations.count)
                addY = stackTop
                    + stackHairline
                    + rows * (propertyRowHeight + stackHairline)
            } else {
                addY = frame.maxY - addPropertyInStackHeight - 4
            }
            actions.append(
                GraphCanvasActionTarget(
                    id: addPropertyActionID,
                    frame: CGRect(
                        x: frame.minX,
                        y: addY,
                        width: frame.width,
                        height: placed.isCited ? addPropertyInStackHeight : 30
                    )
                )
            )
        }
        return actions
    }

    /// Approximate painted height so edges / hits track cited-row growth.
    static func contentHeight(for placed: SourceGraphPlacedSubject) -> CGFloat {
        var height = shellPaddingTop + headerHeight
        let hasDescription = !placed.subject.description
            .trimmingCharacters(in: .whitespacesAndNewlines)
            .isEmpty
        if hasDescription {
            height += headerToBodyGap + 18
        }
        if placed.isCited {
            height += headerToBodyGap
            let rows = CGFloat(max(placed.observations.count, 0))
            // Top hairline + N property rows with inter-hairlines + Add property.
            height += stackHairline
                + rows * (propertyRowHeight + stackHairline)
                + addPropertyInStackHeight
        } else {
            height += headerToBodyGap + 26 // quiet Add property under body
            height += shellPaddingBottom
        }
        return max(edgeLayoutHeight, height)
    }

    private static func stackOriginY(for placed: SourceGraphPlacedSubject, frame: CGRect) -> CGFloat {
        var y = frame.minY + shellPaddingTop + headerHeight + headerToBodyGap
        let hasDescription = !placed.subject.description
            .trimmingCharacters(in: .whitespacesAndNewlines)
            .isEmpty
        if hasDescription {
            y += 18 + headerToBodyGap
        }
        return y
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
    /// Nested action id currently under the pointer (idle hover), if any.
    var hoveredActionID: String? = nil

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
            canCite: canCite,
            hoveredActionID: hoveredActionID
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
            canCite: true,
            hoveredActionID: nil
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
    var hoveredActionID: String?

    private var style: EvidenceSubjectKindStyle {
        EvidenceSubjectKindStyle.resolve(typeKey: placed.kind.rawValue, presentation: presentation)
    }

    private var showsSelectionChrome: Bool {
        isSelected || isActivated || isDragging || isConnectingFrom
    }

    var body: some View {
        VStack(alignment: .leading, spacing: EvidenceSubjectCard.headerToBodyGap) {
            headerRow
            if isConnectingFrom {
                Text(L10n.EvidenceGraph.connectingFrom)
                    .font(PVFont.mono(size: 11))
                    .foregroundStyle(PVColor.accent)
            } else if let description = nonEmptyDescription {
                Text(verbatim: description)
                    .font(PVFont.body(size: 12.5))
                    .italic()
                    .foregroundStyle(PVColor.textSecondary)
                    .lineLimit(3)
                    .multilineTextAlignment(.leading)
            }
            if placed.isCited {
                citedPropertyStack
            } else {
                addPropertyQuietRow
            }
        }
        .padding(.horizontal, EvidenceSubjectCard.shellPaddingX)
        .padding(.top, EvidenceSubjectCard.shellPaddingTop)
        .padding(.bottom, placed.isCited ? 0 : EvidenceSubjectCard.shellPaddingBottom)
        .frame(width: EvidenceSubjectCard.width, alignment: .leading)
        .background(cardBackground)
        .overlay(cardBorder)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .shadow(
            color: showsSelectionChrome
                ? Color.black.opacity(0.1)
                : (placed.isCited ? Color.black.opacity(0.06) : .clear),
            radius: isDragging ? 10 : 6,
            y: isDragging ? 4 : 2
        )
        .background(selectionHalo)
    }

    private var headerRow: some View {
        HStack(alignment: .center, spacing: 10) {
            iconChip
            VStack(alignment: .leading, spacing: 2) {
                Text(verbatim: displayLabel)
                    .font(PVFont.display(size: 16, weight: PVFontWeight.medium))
                    .foregroundStyle(placed.isCited ? PVColor.textPrimary : PVColor.textSecondary)
                    .lineLimit(2)
                    .multilineTextAlignment(.leading)
                Text(verbatim: typeLine)
                    .font(PVFont.mono(size: 10))
                    .tracking(0.6)
                    .textCase(.uppercase)
                    .foregroundStyle(style.ink)
                    .lineLimit(1)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            HStack(spacing: 6) {
                PVIcon(.penLine, size: 12)
                    .foregroundStyle(PVColor.textMuted)
                if !placed.isCited {
                    PVIcon(.trash, size: 12)
                        .foregroundStyle(
                            hoveredActionID == EvidenceSubjectCard.deleteActionID
                                ? PVColor.danger
                                : PVColor.textMuted
                        )
                }
            }
            .frame(height: EvidenceSubjectCard.headerHeight, alignment: .top)
        }
    }

    /// Full-bleed ruled stack: kind-line hairlines between tint rows (board).
    private var citedPropertyStack: some View {
        VStack(spacing: EvidenceSubjectCard.stackHairline) {
            ForEach(placed.observations) { observation in
                EvidenceCitedPropertyRow(
                    observation: observation,
                    style: style,
                    isHovered: hoveredActionID
                        == EvidenceSubjectCard.editPropertyActionID(observationID: observation.id)
                )
            }
            addPropertyStackRow
        }
        .padding(.top, EvidenceSubjectCard.stackHairline)
        .background(style.line)
        .clipShape(
            UnevenRoundedRectangle(
                bottomLeadingRadius: PVRadius.md,
                bottomTrailingRadius: PVRadius.md,
                style: .continuous
            )
        )
        .padding(.horizontal, -EvidenceSubjectCard.shellPaddingX)
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
        .padding(.horizontal, EvidenceSubjectCard.shellPaddingX)
        .padding(.top, 9)
        .padding(.bottom, 11)
        .background(hovered && canCite ? style.chip : style.tint)
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
        // Badge overlays the chip; do not use ZStack alignment or the mark shifts.
        .overlay(alignment: .bottomLeading) {
            citationBadge
                .offset(x: -5, y: 5)
        }
    }

    /// Board: 14pt circular notification-dot on the mark corner.
    /// Cited = kind-ink fill + chip check; uncited = surface fill + circle-dashed.
    private var citationBadge: some View {
        ZStack {
            Circle()
                .fill(placed.isCited ? style.ink : PVColor.surfaceCard)
            Circle()
                .strokeBorder(style.tint, lineWidth: 1.5)
            if placed.isCited {
                PVIcon(.check, size: 9)
                    .foregroundStyle(style.chip)
            } else {
                PVIcon(.circleDashed, size: 12)
                    .foregroundStyle(PVColor.evidenceUndocumented)
            }
        }
        .frame(width: 14, height: 14)
        .accessibilityHidden(true)
    }

    private var cardBackground: some View {
        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
            .fill(placed.isCited ? style.tint : PVColor.surfaceCard)
            .overlay {
                if !placed.isCited {
                    RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                        .fill(style.tint.opacity(0.38))
                }
            }
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
///
/// Board: ruled tint row; hover lifts to chip + 2px kind-ink inset at the leading edge.
private struct EvidenceCitedPropertyRow: View {
    let observation: CatalogObservation
    let style: EvidenceSubjectKindStyle
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
                Text(verbatim: valueSummary)
                    .font(PVFont.body(size: 13))
                    .italic(isNegative)
                    .foregroundStyle(isNegative ? PVColor.danger : PVColor.textPrimary)
                    .lineLimit(2)
                    .multilineTextAlignment(.leading)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            PVIcon(.penLine, size: 12)
                .foregroundStyle(PVColor.textFaint)
        }
        .padding(.horizontal, EvidenceSubjectCard.shellPaddingX)
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
        .accessibilityLabel(Text(verbatim: "\(propertyLabel), \(valueSummary)"))
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

    private var isNegative: Bool {
        observation.polarity.trimmingCharacters(in: .whitespacesAndNewlines).lowercased() == ObservationPolarity.negative.rawValue
    }
}

/// Document-space hit frames for the trailing edit / delete icons on graph cards.
///
/// Paint places 12pt ``PVIcon``s in an `HStack(spacing: 6)` flush to the card’s
/// trailing shell padding. Older hard-coded `maxX - N` offsets sat left of those
/// icons after the card widened — keep hits centered on the painted icons.
enum EvidenceCardHeaderActionHits {
    static let iconSize: CGFloat = 12
    static let iconSpacing: CGFloat = 6
    static let hitSize: CGFloat = 28

    static func frames(
        cardFrame: CGRect,
        paddingX: CGFloat,
        paddingTop: CGFloat,
        hitHeight: CGFloat,
        showDelete: Bool
    ) -> (edit: CGRect, delete: CGRect?) {
        var iconTrailingX = cardFrame.maxX - paddingX
        let deleteFrame: CGRect?
        if showDelete {
            deleteFrame = hitFrame(
                iconTrailingX: iconTrailingX,
                cardFrame: cardFrame,
                paddingTop: paddingTop,
                hitHeight: hitHeight
            )
            iconTrailingX -= iconSize + iconSpacing
        } else {
            deleteFrame = nil
        }
        let editFrame = hitFrame(
            iconTrailingX: iconTrailingX,
            cardFrame: cardFrame,
            paddingTop: paddingTop,
            hitHeight: hitHeight
        )
        return (editFrame, deleteFrame)
    }

    private static func hitFrame(
        iconTrailingX: CGFloat,
        cardFrame: CGRect,
        paddingTop: CGFloat,
        hitHeight: CGFloat
    ) -> CGRect {
        let iconMidX = iconTrailingX - iconSize / 2
        return CGRect(
            x: iconMidX - hitSize / 2,
            y: cardFrame.minY + paddingTop,
            width: hitSize,
            height: hitHeight
        )
    }
}
