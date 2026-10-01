import Observation
import SwiftUI

/// Painted geometry of one graph card, measured from SwiftUI layout.
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
    /// Action id → painted rect of the piece tagged with ``View/evidenceCardHitRegion(_:)``.
    var regions: [String: CGRect]

    /// Document-space frame of the whole card whose top-leading corner sits at `origin`.
    func cardFrame(origin: CGPoint) -> CGRect {
        CGRect(origin: origin, size: size)
    }

    /// Document-space hit band for a tagged region: full card width, at the
    /// region's painted top and height. `nil` when the region is not painted.
    func band(_ id: String, origin: CGPoint) -> CGRect? {
        guard let rect = regions[id] else { return nil }
        return CGRect(x: origin.x, y: origin.y + rect.minY, width: size.width, height: rect.height)
    }

    /// Ids of tagged regions whose id starts with `prefix` (e.g. property rows).
    func regionIDs(withPrefix prefix: String) -> [String] {
        regions.keys.filter { $0.hasPrefix(prefix) }.sorted()
    }
}

/// Collects tagged regions from inside a card.
private struct EvidenceCardHitRegionKey: PreferenceKey {
    static let defaultValue: [String: Anchor<CGRect>] = [:]

    static func reduce(value: inout [String: Anchor<CGRect>], nextValue: () -> [String: Anchor<CGRect>]) {
        value.merge(nextValue(), uniquingKeysWith: { _, new in new })
    }
}

extension View {
    /// Tags this painted piece as the hit region for `actionID` on its card.
    func evidenceCardHitRegion(_ actionID: String) -> some View {
        anchorPreference(key: EvidenceCardHitRegionKey.self, value: .bounds) { [actionID: $0] }
    }

    /// Card root: resolves every tagged region plus the card's own size and
    /// reports the layout whenever it changes. Apply before any drag offset.
    func reportsEvidenceCardLayout(_ report: ((EvidenceCardLayout) -> Void)?) -> some View {
        overlayPreferenceValue(EvidenceCardHitRegionKey.self) { anchors in
            if let report {
                GeometryReader { proxy in
                    let layout = EvidenceCardLayout(
                        size: proxy.size,
                        regions: anchors.mapValues { proxy[$0] }
                    )
                    Color.clear
                        .onChange(of: layout, initial: true) { _, measured in
                            report(measured)
                        }
                }
                .allowsHitTesting(false)
                .accessibilityHidden(true)
            }
        }
    }
}

/// Latest measured layout per card id for one graph. Hit targets, edges and
/// the connect rubber band read it; the row-height constants are only the
/// first-frame fallback before a card has reported.
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
