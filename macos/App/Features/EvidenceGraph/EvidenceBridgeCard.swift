import AppKit
import SwiftUI

/// Subordinate bridge / relationship card on the Evidence graph (S6-04 / S7-09).
///
/// Geometry, action ids and VoiceOver copy for a bridge card.
/// ``EvidenceCardView`` paints the card; hit targets and edges read these.
/// Layout follows the Evidence graph board: provisional cards stack working
/// label over type with an honesty body; durable (cited) cards put type·ref
/// on the chrome row and the full edge-summary sentence as the title below.
@MainActor
enum EvidenceBridgeCard {
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

    static func contentCenter(gridX: Int64, gridY: Int64) -> CGPoint {
        GraphCanvasGridMapping.gridPoint(gridX: gridX, gridY: gridY)
    }

    static func topLeadingOffset(gridX: Int64, gridY: Int64) -> CGSize {
        let center = contentCenter(gridX: gridX, gridY: gridY)
        return CGSize(
            width: center.x - width / 2,
            height: center.y - approximateHalfHeight
        )
    }

    /// Grid-space frame used for edge attachment and AppKit hit targets.
    /// Uses the card's measured `layout` when it has reported one; the wrapped-text
    /// estimate below is only the first-frame fallback.
    static func contentFrame(
        for placed: EvidenceGraphPlacedBridge,
        in snapshot: EvidenceGraphSnapshot,
        dragOffset: CGSize = .zero,
        layout: EvidenceCardLayout? = nil
    ) -> CGRect {
        let center = contentCenter(gridX: placed.gridX, gridY: placed.gridY)
        if let layout {
            return layout.cardFrame(origin: CGPoint(
                x: center.x - width / 2 + dragOffset.width,
                y: center.y - approximateHalfHeight + dragOffset.height
            ))
        }
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
        for placed: EvidenceGraphPlacedBridge,
        in snapshot: EvidenceGraphSnapshot
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

    static func extraObservations(for placed: EvidenceGraphPlacedBridge) -> [CatalogObservation] {
        EvidenceCitedPropertyMarks.extraObservations(in: placed.observations)
    }

    static func bodyTextHeight(
        for placed: EvidenceGraphPlacedBridge,
        in snapshot: EvidenceGraphSnapshot
    ) -> CGFloat {
        let text = placed.isCited
            ? EvidenceBridgeEdgeSummary.sentence(for: placed, in: snapshot)
            : L10n.string(L10n.EvidenceGraph.bridgeHonestyBody)
        let font = placed.isCited
            ? PVFont.nsDisplay(size: 14.5, weight: PVFontWeight.medium)
            : PVFont.nsBody(size: 11.5, weight: PVFontWeight.regular, italic: true)
        guard !text.isEmpty else { return 0 }
        return PVTextMetrics.textHeight(text, width: bodyTextWidth, font: font, maxLines: 3)
    }

    /// Body copy width inside the shell, leaving room for the trailing icon column.
    private static var bodyTextWidth: CGFloat {
        width - shellPaddingX * 2 - 8 - 12
    }

    static func actionTargets(
        for placed: EvidenceGraphPlacedBridge,
        in snapshot: EvidenceGraphSnapshot,
        canCite: Bool,
        dragOffset: CGSize = .zero,
        layout: EvidenceCardLayout? = nil
    ) -> [GraphCanvasActionTarget] {
        let frame = contentFrame(for: placed, in: snapshot, dragOffset: dragOffset, layout: layout)
        let headerHits = EvidenceCardHeaderActionHits.frames(
            cardFrame: frame,
            paddingX: shellPaddingX,
            paddingTop: shellPaddingTop,
            hitHeight: headerActionHitHeight,
            showDelete: true
        )
        var actions: [GraphCanvasActionTarget] = [
            GraphCanvasActionTarget(id: editActionID, frame: headerHits.edit),
        ]
        if let layout {
            // Measured: the citation pencil is the tagged icon, not the header
            // pencil. Rows and Add property are their painted bands.
            if canCite, let pencil = layout.centeredHit(
                editCitationActionID,
                origin: frame.origin,
                minSize: EvidenceCardHeaderActionHits.hitSize
            ) {
                actions.append(GraphCanvasActionTarget(id: editCitationActionID, frame: pencil))
            }
            if placed.isCited {
                for observation in extraObservations(for: placed) {
                    let id = EvidenceSubjectCard.editPropertyActionID(observationID: observation.id)
                    if let band = layout.band(id, origin: frame.origin) {
                        actions.append(GraphCanvasActionTarget(id: id, frame: band))
                    }
                }
            }
            if canCite, let band = layout.band(EvidenceSubjectCard.addPropertyActionID, origin: frame.origin) {
                actions.append(GraphCanvasActionTarget(id: EvidenceSubjectCard.addPropertyActionID, frame: band))
            }
            if let deleteFrame = headerHits.delete {
                actions.append(GraphCanvasActionTarget(id: deleteActionID, frame: deleteFrame))
            }
            return actions
        }
        let headerHeight = placed.isCited ? headerContentHeightCited : headerContentHeightUncited
        let bodyTop = frame.minY + shellPaddingTop + headerHeight + headerToBodySpacing
        if canCite {
            actions.append(
                GraphCanvasActionTarget(
                    id: editCitationActionID,
                    frame: citationHitFrame(cardFrame: frame, bodyTop: bodyTop, uncited: !placed.isCited)
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

    /// Fallback before the card reports a layout. The body pencil sits in the
    /// trailing icon column (flush when cited; one spacer left of the edge
    /// when uncited), not under the header edit icon.
    private static func citationHitFrame(cardFrame: CGRect, bodyTop: CGFloat, uncited: Bool) -> CGRect {
        var iconTrailingX = cardFrame.maxX - shellPaddingX
        if uncited {
            iconTrailingX -= EvidenceCardHeaderActionHits.iconSize + EvidenceCardHeaderActionHits.iconSpacing
        }
        let iconMidX = iconTrailingX - EvidenceCardHeaderActionHits.iconSize / 2
        let hit = EvidenceCardHeaderActionHits.hitSize
        return CGRect(
            x: iconMidX - hit / 2,
            y: bodyTop,
            width: hit,
            height: bodyActionHitHeight
        )
    }

    static func accessibilityLabel(
        for placed: EvidenceGraphPlacedBridge,
        in snapshot: EvidenceGraphSnapshot
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
        let honesty = L10n.string(L10n.EvidenceGraph.bridgeHonestyAccessibility)
        if ref.isEmpty {
            return "\(placed.typeLabel), \(name), \(honesty)"
        }
        return "\(placed.typeLabel), \(ref), \(name), \(honesty)"
    }
}
