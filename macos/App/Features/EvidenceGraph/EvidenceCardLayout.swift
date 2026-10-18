import Observation
import SwiftUI

/// Painted geometry of one graph card, measured by ``EvidenceCardView``.
///
/// Cards are paint-only: AppKit (``GraphCanvasPointerController``) owns every
/// click and routes it through document-space hit rects. Those rects come
/// from this measured layout, not from row-height constants, so a hit target
/// moves with the painted piece it belongs to — wrapped titles, wrapped values,
/// and taller rows included. Coordinates are card-local (origin = the card's
/// top-leading corner, before any drag offset).
struct EvidenceCardLayout: Equatable, Sendable {
    /// Painted card size.
    var size: CGSize
    /// Action id → painted rect of that piece.
    var regions: [String: CGRect]

    /// Document-space frame of the whole card whose top-leading corner sits at `origin`.
    func cardFrame(origin: CGPoint) -> CGRect {
        CGRect(origin: origin, size: size)
    }

    /// Document-space hit band for a region: full card width, at the
    /// region's painted top and height. `nil` when the region is not painted.
    func band(_ id: String, origin: CGPoint) -> CGRect? {
        guard let rect = regions[id] else { return nil }
        return CGRect(x: origin.x, y: origin.y + rect.minY, width: size.width, height: rect.height)
    }

    /// Document-space hit centered on a control. Grows to `minSize` on
    /// each axis when the painted piece is smaller than the hit. `nil` when
    /// the region is not painted.
    func centeredHit(_ id: String, origin: CGPoint, minSize: CGFloat) -> CGRect? {
        guard let rect = regions[id] else { return nil }
        let translated = rect.offsetBy(dx: origin.x, dy: origin.y)
        let width = max(translated.width, minSize)
        let height = max(translated.height, minSize)
        return CGRect(
            x: translated.midX - width / 2,
            y: translated.midY - height / 2,
            width: width,
            height: height
        )
    }

    /// Ids of regions whose id starts with `prefix` (e.g. property rows).
    func regionIDs(withPrefix prefix: String) -> [String] {
        regions.keys.filter { $0.hasPrefix(prefix) }.sorted()
    }
}

/// Latest measured layout per card id for one graph. Hit targets, edges and
/// the connect rubber band read it; the row-height constants are only the
/// first-frame fallback before its view is built.
@MainActor
@Observable
final class EvidenceCardLayoutStore {
    private(set) var layouts: [String: EvidenceCardLayout] = [:]

    func layout(for cardID: String) -> EvidenceCardLayout? {
        layouts[cardID]
    }

    func record(_ layout: EvidenceCardLayout, for cardID: String) {
        guard layouts[cardID] != layout else { return }
        layouts[cardID] = layout
    }

    /// Drops layouts for cards no longer on the graph.
    func retain(only cardIDs: Set<String>) {
        let stale = layouts.keys.filter { !cardIDs.contains($0) }
        guard !stale.isEmpty else { return }
        for id in stale { layouts[id] = nil }
    }
}
