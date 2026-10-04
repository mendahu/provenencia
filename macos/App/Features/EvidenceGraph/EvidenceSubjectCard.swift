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
    nonisolated static let width: CGFloat = 264
    /// Top of the card sits this far above the grid center (layout + edges share it).
    nonisolated static let approximateHalfHeight: CGFloat = 36
    /// Minimum edge hit-testing height for a header-only shell.
    static let edgeLayoutHeight: CGFloat = 88

    /// Horizontal / top padding on the card shell (matches board).
    nonisolated static let shellPaddingX: CGFloat = 13
    static let shellPaddingTop: CGFloat = 11
    static let shellPaddingBottom: CGFloat = 12
    static let headerHeight: CGFloat = 28
    static let headerToBodyGap: CGFloat = 9
    /// Ruled property / Add-property stack (board: 1px kind-line hairlines).
    static let propertyRowHeight: CGFloat = 46
    static let addPropertyInStackHeight: CGFloat = 32
    static let stackHairline: CGFloat = 1
    /// S9-D8 footer slot: Promote (unpromoted) or the membership row
    /// (promoted). Both are this height, so promoting never moves an edge.
    static let footerHeight: CGFloat = 36
    /// Horizontal inset for the ghost Promote: 8pt less than the card's padding,
    /// so the borderless button's label lines up with the content above (S9-D8 rev 1).
    static let promoteInset: CGFloat = shellPaddingX - 8
    /// Uncited cards: the footer sits this far below the quiet Add property
    /// row (on top of the stack gap), under a 1pt dashed rule.
    static let uncitedFooterTopMargin: CGFloat = 2
    static let uncitedFooterRule: CGFloat = 1
    /// Painted height of the quiet (uncited) Add property row.
    static let addPropertyQuietHeight: CGFloat = 26

    static let editActionID = "edit"
    static let deleteActionID = "delete"
    static let addPropertyActionID = "addProperty"
    /// Footer on an unpromoted card. v1 opens a Confirm; S9-11 routes it to the flow.
    static let promoteActionID = "promote"
    /// Footer on a promoted card: open the handle's page.
    static let openHandleActionID = "openHandle"

    static func editPropertyActionID(observationID: String) -> String {
        "editProperty.\(observationID)"
    }

    static func observationID(fromEditPropertyAction actionID: String) -> String? {
        let prefix = "editProperty."
        guard actionID.hasPrefix(prefix) else { return nil }
        let id = String(actionID.dropFirst(prefix.count))
        return id.isEmpty ? nil : id
    }

    /// Document-space top-leading corner of the painted card.
    static func origin(for placed: SourceGraphPlacedSubject, dragOffset: CGSize = .zero) -> CGPoint {
        let offset = topLeadingOffset(gridX: placed.gridX, gridY: placed.gridY)
        return CGPoint(x: offset.width + dragOffset.width, y: offset.height + dragOffset.height)
    }

    /// Document-space frame used for edge attachment and AppKit hit targets.
    /// Uses the card's measured `layout` when it has reported one; the
    /// constant-based ``contentHeight(for:)`` is only the first-frame fallback.
    static func edgeFrame(
        for placed: SourceGraphPlacedSubject,
        dragOffset: CGSize = .zero,
        layout: EvidenceCardLayout? = nil
    ) -> CGRect {
        let origin = origin(for: placed, dragOffset: dragOffset)
        if let layout {
            return layout.cardFrame(origin: origin)
        }
        return CGRect(origin: origin, size: CGSize(width: width, height: contentHeight(for: placed)))
    }

    /// Nested action frames in document space (same coordinate as ``edgeFrame``).
    /// With a measured `layout`, every row / Add property / footer target is the
    /// painted band of the piece it belongs to, so hits move with the paint.
    static func actionTargets(
        for placed: SourceGraphPlacedSubject,
        canCite: Bool,
        dragOffset: CGSize = .zero,
        layout: EvidenceCardLayout? = nil
    ) -> [GraphCanvasActionTarget] {
        let frame = edgeFrame(for: placed, dragOffset: dragOffset, layout: layout)
        if let layout {
            return measuredActionTargets(for: placed, canCite: canCite, frame: frame, layout: layout)
        }
        let headerHits = EvidenceCardHeaderActionHits.frames(
            cardFrame: frame,
            paddingX: shellPaddingX,
            paddingTop: shellPaddingTop,
            hitHeight: headerHeight,
            showDelete: true
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
            let addFrame: CGRect
            if placed.isCited {
                let rows = CGFloat(placed.observations.count)
                let addY = stackTop
                    + stackHairline
                    + rows * (propertyRowHeight + stackHairline)
                addFrame = CGRect(x: frame.minX, y: addY, width: frame.width, height: addPropertyInStackHeight)
            } else {
                // Quiet row sits above the footer block; its hit stays clear of the footer.
                let footerTop = frame.maxY - uncitedFooterBlockHeight
                addFrame = CGRect(
                    x: frame.minX,
                    y: footerTop - addPropertyQuietHeight - 2,
                    width: frame.width,
                    height: addPropertyQuietHeight + 2
                )
            }
            actions.append(GraphCanvasActionTarget(id: addPropertyActionID, frame: addFrame))
        }

        // Promote needs no Artifact, so it ignores `canCite`.
        actions.append(
            GraphCanvasActionTarget(
                id: placed.membership == nil ? promoteActionID : openHandleActionID,
                frame: footerFrame(cardFrame: frame)
            )
        )
        return actions
    }

    /// Header icons stay on fixed top-anchored metrics; everything below them
    /// comes from the measured bands.
    private static func measuredActionTargets(
        for placed: SourceGraphPlacedSubject,
        canCite: Bool,
        frame: CGRect,
        layout: EvidenceCardLayout
    ) -> [GraphCanvasActionTarget] {
        let origin = frame.origin
        let headerHits = EvidenceCardHeaderActionHits.frames(
            cardFrame: frame,
            paddingX: shellPaddingX,
            paddingTop: shellPaddingTop,
            hitHeight: headerHeight,
            showDelete: true
        )
        var actions: [GraphCanvasActionTarget] = [
            GraphCanvasActionTarget(id: editActionID, frame: headerHits.edit),
        ]
        if let deleteFrame = headerHits.delete {
            actions.append(GraphCanvasActionTarget(id: deleteActionID, frame: deleteFrame))
        }
        if placed.isCited {
            for observation in placed.observations {
                let id = editPropertyActionID(observationID: observation.id)
                if let band = layout.band(id, origin: origin) {
                    actions.append(GraphCanvasActionTarget(id: id, frame: band))
                }
            }
        }
        if canCite, let band = layout.band(addPropertyActionID, origin: origin) {
            actions.append(GraphCanvasActionTarget(id: addPropertyActionID, frame: band))
        }
        let footerID = placed.membership == nil ? promoteActionID : openHandleActionID
        if let band = layout.band(footerID, origin: origin) {
            actions.append(GraphCanvasActionTarget(id: footerID, frame: band))
        }
        return actions
    }

    /// Footer hit / paint rect: full card width × ``footerHeight``, bottom-anchored.
    static func footerFrame(cardFrame: CGRect) -> CGRect {
        CGRect(
            x: cardFrame.minX,
            y: cardFrame.maxY - footerHeight,
            width: cardFrame.width,
            height: footerHeight
        )
    }

    /// Uncited cards: stack gap + margin + dashed rule + footer, replacing the bottom padding.
    static var uncitedFooterBlockHeight: CGFloat {
        headerToBodyGap + uncitedFooterTopMargin + uncitedFooterRule + footerHeight
    }

    /// Constant-based height estimate: the first-frame fallback before the card
    /// reports its measured ``EvidenceCardLayout``. Hits and edges use the
    /// measured layout once it exists; do not tune this to chase paint.
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
            // Top hairline + N property rows with inter-hairlines + Add property,
            // then a hairline and the footer as the stack's last row.
            height += stackHairline
                + rows * (propertyRowHeight + stackHairline)
                + addPropertyInStackHeight
                + stackHairline + footerHeight
        } else {
            height += headerToBodyGap + addPropertyQuietHeight // quiet Add property under body
            height += uncitedFooterBlockHeight // footer replaces the bottom padding
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
    /// Copy formatted with the graph snapshot (``EvidenceGraphText``).
    let text: EvidenceCardText
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
    /// Being dragged: lifted paint. AppKit moves the card's view; the card
    /// never offsets itself.
    var isDragging: Bool = false
    /// Nested action id currently under the pointer (idle hover), if any.
    var hoveredActionID: String? = nil
    /// Nested action id held down by the pointer, if any (pressed paint).
    var pressedActionID: String? = nil
    /// Receives the painted layout (card size + tagged hit regions) for hit targets and edges.
    var onLayout: ((EvidenceCardLayout) -> Void)? = nil

    var body: some View {
        EvidenceSubjectCardChrome(
            placed: placed,
            presentation: presentation,
            isSelected: isSelected || isConnectingFrom,
            isActivated: isActivated,
            isDragging: isDragging,
            isConnectingFrom: isConnectingFrom,
            canCite: canCite,
            hoveredActionID: hoveredActionID,
            pressedActionID: pressedActionID
        )
        .reportsEvidenceCardLayout(onLayout)
        .opacity(ghostOpacity)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: text.accessibilityLabel))
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
            hoveredActionID: nil,
            pressedActionID: nil
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
            ? L10n.string(L10n.EvidenceGraph.citedAccessibility)
            : L10n.string(L10n.EvidenceGraph.uncitedAccessibility)
        let ref = placed.subject.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        if ref.isEmpty {
            return "\(placed.typeLabel), \(name), \(citation)"
        }
        return "\(placed.typeLabel), \(ref), \(name), \(citation)"
    }

    /// VoiceOver name for the footer action: "Promote James Robins, CPR-…" or "Open person PER-…".
    static func footerAccessibilityActionName(for placed: SourceGraphPlacedSubject) -> String {
        if let membership = placed.membership {
            let target = membershipName(membership).map { "\($0), \(membership.entity.ref)" } ?? membership.entity.ref
            return L10n.EvidenceGraph.openHandleAccessibility(kind: placed.kind, ref: target)
        }
        let label = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return L10n.EvidenceGraph.promoteAccessibility(
            label: label.isEmpty ? placed.subject.ref : label,
            ref: placed.subject.ref
        )
    }

    /// The handle's resolved name for the membership row (S9-09), or nil when
    /// it has none — the row then keeps its *Open … page* copy.
    static func membershipName(_ membership: CatalogSubjectMembership) -> String? {
        guard let name = membership.name else { return nil }
        let text = NameValueDisplay.string(for: name)
        return text.isEmpty ? nil : text
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
    var pressedActionID: String?

    @Environment(\.accessibilityReduceMotion) private var reduceMotion

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
                uncitedFooter
            }
        }
        .padding(.horizontal, EvidenceSubjectCard.shellPaddingX)
        .padding(.top, EvidenceSubjectCard.shellPaddingTop)
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
                PVIcon(.trash, size: 12)
                    .foregroundStyle(
                        hoveredActionID == EvidenceSubjectCard.deleteActionID
                            ? PVColor.danger
                            : PVColor.textMuted
                    )
            }
            .frame(height: EvidenceSubjectCard.headerHeight, alignment: .top)
        }
    }

    /// Full-bleed ruled stack: kind-line hairlines between tint rows (board).
    private var citedPropertyStack: some View {
        EvidenceCitedPropertyStack(
            observations: placed.observations,
            style: style,
            hoveredActionID: hoveredActionID
        ) {
            // The footer is the stack's last row, one hairline below Add property.
            VStack(spacing: EvidenceSubjectCard.stackHairline) {
                addPropertyStackRow
                cardFooter
            }
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
        .evidenceCardHitRegion(EvidenceSubjectCard.addPropertyActionID)
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
        .frame(height: EvidenceSubjectCard.addPropertyQuietHeight, alignment: .topLeading)
        .evidenceCardHitRegion(EvidenceSubjectCard.addPropertyActionID)
        .accessibilityHidden(true)
    }

    // MARK: Footer (S9-D8)

    /// Uncited: the footer bleeds to the card edges under a 1pt kind-line rule
    /// (cited cards get the same rule from the ruled stack's hairline).
    private var uncitedFooter: some View {
        VStack(spacing: 0) {
            Rectangle()
                .fill(style.line)
                .frame(height: EvidenceSubjectCard.uncitedFooterRule)
            cardFooter
        }
        .padding(.top, EvidenceSubjectCard.uncitedFooterTopMargin)
        .padding(.horizontal, -EvidenceSubjectCard.shellPaddingX)
    }

    /// One fixed slot: Promote until the subject has an accepted Identity
    /// Claim, then the membership row. Paint only — AppKit owns the hits.
    @ViewBuilder
    private var cardFooter: some View {
        if let membership = placed.membership {
            membershipRow(membership)
        } else {
            promoteFooter
        }
    }

    /// Kit ghost Button on the kind band: still an action on this subject.
    private var promoteFooter: some View {
        HStack(spacing: 0) {
            Button {} label: {
                HStack(spacing: PVSpacing.space3) {
                    PVIcon(.arrowUpRight, size: PVControlSize.sm.iconGlyphSize)
                    Text(L10n.EvidenceGraph.promote)
                }
            }
            .buttonStyle(.pv(.ghost, size: .sm))
            .pvHostInteraction(
                hovered: hoveredActionID == EvidenceSubjectCard.promoteActionID,
                pressed: pressedActionID == EvidenceSubjectCard.promoteActionID
            )
            .allowsHitTesting(false)
            .help(Text(L10n.EvidenceGraph.promoteHelp(kind: placed.kind)))
            Spacer(minLength: 0)
        }
        .padding(.horizontal, EvidenceSubjectCard.promoteInset)
        .frame(maxWidth: .infinity, minHeight: EvidenceSubjectCard.footerHeight,
               maxHeight: EvidenceSubjectCard.footerHeight, alignment: .leading)
        .background(style.band)
        .evidenceCardHitRegion(EvidenceSubjectCard.promoteActionID)
        .accessibilityHidden(true)
    }

    /// Conclusion link on the kind band (S9-D8 rev 1 / rev 2). Told apart from the
    /// Observation rows by shape — outline mono ref, no micro-caps label,
    /// disclosure chevron — while keeping the card's colour identity.
    private func membershipRow(_ membership: CatalogSubjectMembership) -> some View {
        let hovered = hoveredActionID == EvidenceSubjectCard.openHandleActionID
        let pressed = pressedActionID == EvidenceSubjectCard.openHandleActionID
        return HStack(spacing: 8) {
            PVBadge(text: membership.entity.ref, tone: .neutral, subtle: true, foreground: style.ink)
            // The handle's resolved name (S9-09) in the slot the link copy holds
            // until a name exists — same slot, so promoting never relayouts.
            Group {
                if let name = EvidenceSubjectCard.membershipName(membership) {
                    Text(verbatim: name)
                } else {
                    Text(L10n.EvidenceGraph.openHandlePage(kind: placed.kind)).italic()
                }
            }
            .font(PVFont.body(size: 12.5))
            .foregroundStyle(style.ink)
            .lineLimit(1)
            .truncationMode(.tail)
            Spacer(minLength: 0)
            PVIcon(.chevronForward, size: 14)
                .foregroundStyle(style.ink)
        }
        .scaleEffect(pressed && !reduceMotion ? PVMotion.pressScale : 1)
        .padding(.horizontal, EvidenceSubjectCard.shellPaddingX)
        .frame(maxWidth: .infinity, minHeight: EvidenceSubjectCard.footerHeight,
               maxHeight: EvidenceSubjectCard.footerHeight, alignment: .leading)
        .background {
            // Kind band, mixed with kind ink at 10% (hover) / 18% (pressed).
            // One always-present layer; only its opacity follows the state.
            style.ink
                .opacity(pressed ? 0.18 : (hovered ? 0.10 : 0))
                .background(style.band)
        }
        .pvAnimation(PVMotion.instantStandard, value: pressed)
        .pvAnimation(PVMotion.instantStandard, value: hovered)
        .evidenceCardHitRegion(EvidenceSubjectCard.openHandleActionID)
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
