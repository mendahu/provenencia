import SwiftUI

/// Grows a card to `scale` on hover and lifts it with the medium shadow. Hover
/// state stays on this card. The hit shape is always the grown size, so the
/// shrinking edge never crosses the pointer and retriggers the hover. The card
/// is one layer before it scales, so the glyphs are not redrawn on every frame.
/// Reduce motion keeps the shadow and skips the grow. The hovered card draws
/// above its neighbours. Content-agnostic: `resting` is the shadow the card
/// keeps at rest, nil for none.
struct PVLiftHover: ViewModifier {
    var scale: CGFloat
    var resting: [PVShadowLayer]?

    @State private var isHovering = false
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    func body(content: Content) -> some View {
        let lifted = PVElevation.md[1]
        let grow = isHovering && !reduceMotion
        content
            .compositingGroup()
            .pvShadow(resting ?? [])
            .shadow(
                color: lifted.color.opacity(isHovering ? 1 : 0),
                radius: lifted.radius,
                x: lifted.x,
                y: lifted.y
            )
            .scaleEffect(grow ? scale : 1, anchor: .center)
            .animation(reduceMotion ? nil : .easeOut(duration: PVMotion.durationFast), value: isHovering)
            .contentShape(GrownHitShape(scale: scale))
            .onHover { hovering in
                guard hovering != isHovering else { return }
                isHovering = hovering
            }
            .zIndex(isHovering ? 1 : 0)
    }
}

extension View {
    /// Grows to `scale` and lifts with `PVElevation.md` on hover (see `PVLiftHover`).
    func pvLiftHover(scale: CGFloat, resting: [PVShadowLayer]? = nil) -> some View {
        modifier(PVLiftHover(scale: scale, resting: resting))
    }
}

/// Layout frame expanded to the hovered scale. It does not change with the
/// animation, so pointer exit happens once.
private struct GrownHitShape: Shape {
    var scale: CGFloat

    func path(in rect: CGRect) -> Path {
        let dx = rect.width * (scale - 1) / 2
        let dy = rect.height * (scale - 1) / 2
        return Path(rect.insetBy(dx: -dx, dy: -dy))
    }
}
