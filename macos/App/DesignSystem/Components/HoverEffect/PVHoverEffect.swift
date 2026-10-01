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
    @Environment(\.pvHostInteraction) private var hostInteraction

    var body: some View {
        let hovering = hostInteraction?.isHovered ?? isHovering
        let pressed = hostInteraction?.isPressed ?? isPressed
        content(hovering && isEnabled)
            .contentShape(Rectangle())
            .opacity(isEnabled ? 1 : 0.45)
            .scaleEffect(pressed && !reduceMotion ? PVMotion.pressScale : 1)
            .pvAnimation(hoverAnimation, value: hovering)
            .pvAnimation(PVMotion.instantStandard, value: pressed)
            .onHover { isHovering = $0 }
    }
}

/// Hover / press supplied by a host that owns the pointer itself — paint-only
/// surfaces such as the graph canvas, where AppKit hit-tests and SwiftUI
/// `onHover` / `isPressed` never fire. ``PVHoverEffect`` (and so every kit
/// control built on it) prefers this over its own tracking when set.
struct PVHostInteraction: Equatable, Sendable {
    var isHovered: Bool
    var isPressed: Bool
}

private struct PVHostInteractionKey: EnvironmentKey {
    static let defaultValue: PVHostInteraction? = nil
}

extension EnvironmentValues {
    var pvHostInteraction: PVHostInteraction? {
        get { self[PVHostInteractionKey.self] }
        set { self[PVHostInteractionKey.self] = newValue }
    }
}

extension View {
    /// Drives kit hover / press chrome from a host-owned pointer (paint-only canvas cards).
    func pvHostInteraction(hovered: Bool, pressed: Bool) -> some View {
        environment(\.pvHostInteraction, PVHostInteraction(isHovered: hovered, isPressed: pressed))
    }
}
