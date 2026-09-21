import Observation
import SwiftUI

/// Drawn A—bridge—B segments under subject cards (S6-04).
///
/// Each segment gradients from the **primary** kind ink at the primary end into
/// a neutral at the bridge end (design board). The two spokes leaving a bridge
/// therefore carry the two endpoint colors. Selecting a bridge brightens the
/// same hues (wider soft under-stroke + slightly thicker stroke) — never remaps
/// to accent.
struct EvidenceGraphEdgeLayer: View {
    let snapshot: SourceGraphSnapshot
    var selectedBridgeID: String?
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

                let aRect = primaryFrame(a, offset: dragOffsets[a.id] ?? .zero)
                let bridgeRect = bridgeFrame(bridge, offset: dragOffsets[bridge.id] ?? .zero)
                let bRect = primaryFrame(b, offset: dragOffsets[b.id] ?? .zero)

                // Always primary → bridge so gradient mapping stays unambiguous.
                strokePrimaryToBridge(
                    context: context,
                    primary: a,
                    primaryRect: aRect,
                    bridgeRect: bridgeRect,
                    selected: selected,
                    lineWidth: lineWidth
                )
                strokePrimaryToBridge(
                    context: context,
                    primary: b,
                    primaryRect: bRect,
                    bridgeRect: bridgeRect,
                    selected: selected,
                    lineWidth: lineWidth
                )
            }
        }
        .allowsHitTesting(false)
        .accessibilityHidden(true)
    }

    private var bridgeNeutral: Color { PVColor.borderStrong }

    private func strokePrimaryToBridge(
        context: GraphicsContext,
        primary: SourceGraphPlacedSubject,
        primaryRect: CGRect,
        bridgeRect: CGRect,
        selected: Bool,
        lineWidth: CGFloat
    ) {
        let seg = GraphCanvasEdgeGeometry.segment(fromRect: primaryRect, toRect: bridgeRect)
        let style = EvidenceSubjectKindStyle.forKind(primary.kind)
        // Selected: lift toward the lighter kind line color — same hue family.
        let primaryInk = selected ? style.line : style.ink
        let neutral = selected ? PVColor.borderDefault : bridgeNeutral
        let shading = GraphicsContext.Shading.linearGradient(
            Gradient(stops: [
                .init(color: primaryInk, location: 0),
                .init(color: primaryInk, location: 0.2),
                .init(color: neutral, location: 0.65),
                .init(color: neutral, location: 1),
            ]),
            startPoint: seg.start,
            endPoint: seg.end
        )
        let path = Path(seg.path)
        if selected {
            context.stroke(
                path,
                with: .linearGradient(
                    Gradient(stops: [
                        .init(color: primaryInk.opacity(0.35), location: 0),
                        .init(color: primaryInk.opacity(0.2), location: 0.35),
                        .init(color: neutral.opacity(0.25), location: 1),
                    ]),
                    startPoint: seg.start,
                    endPoint: seg.end
                ),
                style: StrokeStyle(lineWidth: lineWidth + 2.5, lineCap: .round)
            )
        }
        context.stroke(
            path,
            with: shading,
            style: StrokeStyle(lineWidth: lineWidth, lineCap: .round)
        )
    }

    private func primaryFrame(_ placed: SourceGraphPlacedSubject, offset: CGSize) -> CGRect {
        EvidenceSubjectCard.edgeFrame(for: placed, dragOffset: offset)
    }

    private func bridgeFrame(_ placed: SourceGraphPlacedBridge, offset: CGSize) -> CGRect {
        EvidenceBridgeCard.contentFrame(
            gridX: placed.gridX,
            gridY: placed.gridY,
            dragOffset: offset
        )
    }
}

/// Observes AppKit live drag offsets without invalidating the card `ForEach` parent.
struct EvidenceGraphEdgesHost: View {
    @Bindable var pointer: GraphCanvasPointerController
    let snapshot: SourceGraphSnapshot
    var selectedBridgeID: String?

    var body: some View {
        EvidenceGraphEdgeLayer(
            snapshot: snapshot,
            selectedBridgeID: selectedBridgeID,
            dragOffsets: pointer.offsets
        )
    }
}

/// Dashed rubber-band from origin card toward the cursor while waiting for B.
struct EvidenceGraphConnectRubberBand: View {
    let origin: SourceGraphPlacedSubject
    let cursor: CGPoint

    var body: some View {
        Canvas { context, _ in
            let fromRect = EvidenceSubjectCard.edgeFrame(for: origin)
            let edgeStart = GraphCanvasEdgeGeometry.attachmentPoint(fromRect: fromRect, toward: cursor)
            let start = GraphCanvasEdgeGeometry.tuckInside(edgeStart, rect: fromRect)
            let path = GraphCanvasEdgeGeometry.cubicPath(from: start, to: cursor)
            let ink = EvidenceSubjectKindStyle.forKind(origin.kind).ink
            context.stroke(
                Path(path),
                with: .linearGradient(
                    Gradient(stops: [
                        .init(color: ink, location: 0),
                        .init(color: PVColor.accent.opacity(0.45), location: 1),
                    ]),
                    startPoint: start,
                    endPoint: cursor
                ),
                style: StrokeStyle(lineWidth: 1.75, lineCap: .round, dash: [5, 4])
            )
        }
        .allowsHitTesting(false)
        .accessibilityHidden(true)
    }
}
