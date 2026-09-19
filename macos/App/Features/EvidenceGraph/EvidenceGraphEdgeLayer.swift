import SwiftUI

/// Drawn A—bridge—B segments under subject cards (S6-04).
struct EvidenceGraphEdgeLayer: View {
    let snapshot: SourceGraphSnapshot
    /// Selected bridge id thickens both segments.
    var selectedBridgeID: String?
    /// Live drag offsets keyed by subject/bridge id (document space).
    var dragOffsets: [String: CGSize] = [:]

    var body: some View {
        Canvas { context, _ in
            for bridge in snapshot.bridges {
                guard let aID = bridge.endpointAID,
                      let bID = bridge.endpointBID,
                      let a = snapshot.subjects.first(where: { $0.id == aID }),
                      let b = snapshot.subjects.first(where: { $0.id == bID })
                else { continue }

                let selected = selectedBridgeID == bridge.id
                let lineWidth: CGFloat = selected ? 2 : 1.5
                let color = PVColor.accent

                let aRect = primaryFrame(a, offset: dragOffsets[a.id] ?? .zero)
                let bridgeRect = bridgeFrame(bridge, offset: dragOffsets[bridge.id] ?? .zero)
                let bRect = primaryFrame(b, offset: dragOffsets[b.id] ?? .zero)

                let segA = GraphCanvasEdgeGeometry.segment(fromRect: aRect, toRect: bridgeRect)
                let segB = GraphCanvasEdgeGeometry.segment(fromRect: bridgeRect, toRect: bRect)

                context.stroke(
                    Path(segA.path),
                    with: .color(color),
                    style: StrokeStyle(lineWidth: lineWidth, lineCap: .round)
                )
                context.stroke(
                    Path(segB.path),
                    with: .color(color),
                    style: StrokeStyle(lineWidth: lineWidth, lineCap: .round)
                )
            }
        }
        .allowsHitTesting(false)
        .accessibilityHidden(true)
    }

    private func primaryFrame(_ placed: SourceGraphPlacedSubject, offset: CGSize) -> CGRect {
        let center = EvidenceSubjectCard.contentCenter(gridX: placed.gridX, gridY: placed.gridY)
        return CGRect(
            x: center.x - EvidenceSubjectCard.width / 2 + offset.width,
            y: center.y - EvidenceSubjectCard.approximateHalfHeight + offset.height,
            width: EvidenceSubjectCard.width,
            height: EvidenceSubjectCard.approximateHalfHeight * 2
        )
    }

    private func bridgeFrame(_ placed: SourceGraphPlacedBridge, offset: CGSize) -> CGRect {
        EvidenceBridgeCard.contentFrame(
            gridX: placed.gridX,
            gridY: placed.gridY,
            dragOffset: offset
        )
    }
}

/// Dashed rubber-band from origin card toward the cursor while waiting for B.
struct EvidenceGraphConnectRubberBand: View {
    let origin: SourceGraphPlacedSubject
    let cursor: CGPoint

    var body: some View {
        Canvas { context, _ in
            let fromRect = CGRect(
                x: EvidenceSubjectCard.contentCenter(gridX: origin.gridX, gridY: origin.gridY).x
                    - EvidenceSubjectCard.width / 2,
                y: EvidenceSubjectCard.contentCenter(gridX: origin.gridX, gridY: origin.gridY).y
                    - EvidenceSubjectCard.approximateHalfHeight,
                width: EvidenceSubjectCard.width,
                height: EvidenceSubjectCard.approximateHalfHeight * 2
            )
            let start = GraphCanvasEdgeGeometry.attachmentPoint(fromRect: fromRect, toward: cursor)
            let path = GraphCanvasEdgeGeometry.cubicPath(from: start, to: cursor)
            context.stroke(
                Path(path),
                with: .color(PVColor.accent),
                style: StrokeStyle(lineWidth: 1.75, lineCap: .round, dash: [5, 4])
            )
        }
        .allowsHitTesting(false)
        .accessibilityHidden(true)
    }
}
