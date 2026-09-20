import SwiftUI

/// Shared interaction chrome for `ButtonStyle` bodies: hover tracking,
/// disabled opacity, reduce-motion-aware press scale, and a single
/// `contentShape(Rectangle())` so hover and click share one hit target.
/// Callers build content from `showHover` (`isHovering && isEnabled`).
struct PVHoverEffect<Content: View>: View {
    let isPressed: Bool
    /// Defaults to snappy timing; `PVButtonBody` passes `PVMotion.fastStandard`
    /// to keep its existing feel.
    var hoverAnimation: Animation = PVMotion.instantStandard
    @ViewBuilder var content: (_ showHover: Bool) -> Content

    @State private var isHovering = false
    @Environment(\.isEnabled) private var isEnabled
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    var body: some View {
        content(isHovering && isEnabled)
            .contentShape(Rectangle())
            .opacity(isEnabled ? 1 : 0.45)
            .scaleEffect(isPressed && !reduceMotion ? PVMotion.pressScale : 1)
            .pvAnimation(hoverAnimation, value: isHovering)
            .pvAnimation(PVMotion.instantStandard, value: isPressed)
            .onHover { isHovering = $0 }
    }
}
