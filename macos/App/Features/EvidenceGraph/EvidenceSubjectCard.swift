import SwiftUI

/// Geometry, action ids and VoiceOver copy for a primary subject card.
/// ``EvidenceCardView`` paints the card; hit targets and edges read these.
@MainActor
enum EvidenceSubjectCard {
    /// Fixed card width — keep in sync with content centering / offset.
    nonisolated static let width: CGFloat = 264
    /// Top of the card sits this far above the grid center (layout + edges share it).
    nonisolated static let approximateHalfHeight: CGFloat = 36
    /// Minimum edge hit-testing height for a header-only shell.
    static let edgeLayoutHeight: CGFloat = 88

    /// Horizontal / top padding on the card shell (matches board).
    nonisolated static let shellPaddingX: CGFloat = 13
    static let shellPaddingTop: CGFloat = 11
    static let headerHeight: CGFloat = 28
    static let headerToBodyGap: CGFloat = 9
    /// Ruled property / Add-property stack (board: 1px kind-line hairlines).
    static let propertyRowHeight: CGFloat = 46
    static let addPropertyInStackHeight: CGFloat = 32
    static let stackHairline: CGFloat = 1
    /// S9-D8 footer slot: Promote (unpromoted) or the membership row
    /// (promoted). Both are this height, so promoting never moves an edge.
    static let footerHeight: CGFloat = 36
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
    /// Icon button at the trailing edge of a promoted card's footer: take the filing back.
    static let unpromoteActionID = "unpromote"
    /// Width of the Unpromote square; also the footer height, so it is square.
    static let unpromoteSlotWidth: CGFloat = 36

    /// Where a promoted footer's pieces sit inside `footer`: the handle row, the
    /// 1pt hairline, and the Unpromote slot. Paint, regions, and hits all read this.
    static func promotedFooterParts(footer: CGRect) -> (handle: CGRect, hairline: CGRect, slot: CGRect) {
        let slot = CGRect(
            x: footer.maxX - unpromoteSlotWidth, y: footer.minY,
            width: unpromoteSlotWidth, height: footer.height
        )
        let hairline = CGRect(x: slot.minX - 1, y: footer.minY + 8, width: 1, height: footer.height - 16)
        let handle = CGRect(
            x: footer.minX, y: footer.minY,
            width: hairline.minX - footer.minX, height: footer.height
        )
        return (handle, hairline, slot)
    }

    static func editPropertyActionID(observationID: String) -> String {
        "editProperty.\(observationID)"
    }

    static func observationID(fromEditPropertyAction actionID: String) -> String? {
        let prefix = "editProperty."
        guard actionID.hasPrefix(prefix) else { return nil }
        let id = String(actionID.dropFirst(prefix.count))
        return id.isEmpty ? nil : id
    }

    /// Grid-space top-leading corner of the painted card.
    static func origin(for placed: EvidenceGraphPlacedSubject, dragOffset: CGSize = .zero) -> CGPoint {
        let offset = topLeadingOffset(gridX: placed.gridX, gridY: placed.gridY)
        return CGPoint(x: offset.width + dragOffset.width, y: offset.height + dragOffset.height)
    }

    /// Grid-space frame used for edge attachment and AppKit hit targets.
    /// Uses the card's measured `layout` when it has reported one; the
    /// constant-based ``contentHeight(for:)`` is only the first-frame fallback.
    static func edgeFrame(
        for placed: EvidenceGraphPlacedSubject,
        dragOffset: CGSize = .zero,
        layout: EvidenceCardLayout? = nil
    ) -> CGRect {
        let origin = origin(for: placed, dragOffset: dragOffset)
        if let layout {
            return layout.cardFrame(origin: origin)
        }
        return CGRect(origin: origin, size: CGSize(width: width, height: contentHeight(for: placed)))
    }

    /// Nested action frames in grid space (same coordinate as ``edgeFrame``).
    /// With a measured `layout`, every row / Add property / footer target is the
    /// painted band of the piece it belongs to, so hits move with the paint.
    static func actionTargets(
        for placed: EvidenceGraphPlacedSubject,
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
        if placed.membership == nil {
            actions.append(GraphCanvasActionTarget(id: promoteActionID, frame: footerFrame(cardFrame: frame)))
        } else {
            let parts = promotedFooterParts(footer: footerFrame(cardFrame: frame))
            actions.append(GraphCanvasActionTarget(id: openHandleActionID, frame: parts.handle))
            actions.append(unpromoteTarget(frame: parts.slot))
        }
        return actions
    }

    /// Header icons stay on fixed top-anchored metrics; everything below them
    /// comes from the measured bands.
    private static func measuredActionTargets(
        for placed: EvidenceGraphPlacedSubject,
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
        if placed.membership == nil {
            if let band = layout.band(promoteActionID, origin: origin) {
                actions.append(GraphCanvasActionTarget(id: promoteActionID, frame: band))
            }
        } else {
            // The handle row and the Unpromote square are the painted rects, not the full-width band.
            if let handle = layout.centeredHit(openHandleActionID, origin: origin, minSize: 0) {
                actions.append(GraphCanvasActionTarget(id: openHandleActionID, frame: handle))
            }
            if let slot = layout.centeredHit(unpromoteActionID, origin: origin, minSize: 0) {
                actions.append(unpromoteTarget(frame: slot))
            }
        }
        return actions
    }

    private static func unpromoteTarget(frame: CGRect) -> GraphCanvasActionTarget {
        GraphCanvasActionTarget(
            id: unpromoteActionID, frame: frame, help: L10n.string(L10n.EvidenceGraph.unpromote)
        )
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
    static func contentHeight(for placed: EvidenceGraphPlacedSubject) -> CGFloat {
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

    private static func stackOriginY(for placed: EvidenceGraphPlacedSubject, frame: CGRect) -> CGFloat {
        var y = frame.minY + shellPaddingTop + headerHeight + headerToBodyGap
        let hasDescription = !placed.subject.description
            .trimmingCharacters(in: .whitespacesAndNewlines)
            .isEmpty
        if hasDescription {
            y += 18 + headerToBodyGap
        }
        return y
    }

    /// Grid-space center for a placed (or ghost) card.
    static func contentCenter(gridX: Int64, gridY: Int64) -> CGPoint {
        GraphCanvasGridMapping.gridPoint(gridX: gridX, gridY: gridY)
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
    static func accessibilityLabel(for placed: EvidenceGraphPlacedSubject) -> String {
        let name = placed.titleOrTypeLabel
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
    static func footerAccessibilityActionName(
        for placed: EvidenceGraphPlacedSubject,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> String {
        if let membership = placed.membership {
            let target = membershipName(membership, orders: orders).map { "\($0), \(membership.entity.ref)" } ?? membership.entity.ref
            return L10n.EvidenceGraph.openHandleAccessibility(kind: placed.kind, ref: target)
        }
        return L10n.EvidenceGraph.promoteAccessibility(
            label: placed.subject.trimmedLabel ?? placed.subject.ref,
            ref: placed.subject.ref
        )
    }

    /// The handle's auto-reconciled name for the membership row (S9-09), or nil when
    /// it has none — the row then keeps its *Open … page* copy.
    static func membershipName(_ membership: CatalogSubjectMembership) -> String? {
        membershipName(membership, orders: NameFormatOrders())
    }

    static func membershipName(
        _ membership: CatalogSubjectMembership,
        orders: NameFormatOrders
    ) -> String? {
        guard let name = membership.name else { return nil }
        let text = NameValueDisplay.string(for: name, style: .natural, orders: orders)
        return text.isEmpty ? nil : text
    }
}

/// Grid-space hit frames for the trailing edit / delete icons on graph cards.
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
