import SwiftUI

/// Provenencia design-system motion tokens.
///
/// Transcribed from `tokens/motion.css`: "Restrained: fades and short
/// slides. Nothing bounces." `prefers-reduced-motion` collapses every
/// duration to 0 on the web; there's no static-token equivalent for that on
/// macOS, so call sites should gate their own animations on
/// `@Environment(\.accessibilityReduceMotion)` instead.
enum PVMotion {
    static let durationInstant: TimeInterval = 0.08
    static let durationFast: TimeInterval = 0.13
    static let durationNormal: TimeInterval = 0.2
    static let durationSlow: TimeInterval = 0.32
    static let durationPage: TimeInterval = 0.46

    /// `--press-scale`: the scale applied to a pressed control (buttons).
    static let pressScale: CGFloat = 0.985
    /// `--lift-hover`: the y-offset applied to a hoverable card on hover.
    static let liftHover: CGFloat = -1

    static let easeStandard = Animation.timingCurve(0.2, 0.6, 0.2, 1, duration: durationNormal)
    static let easeOut = Animation.timingCurve(0, 0.55, 0.45, 1, duration: durationNormal)
    static let easeIn = Animation.timingCurve(0.55, 0, 1, 0.45, duration: durationNormal)
    static let easeInOut = Animation.timingCurve(0.5, 0, 0.5, 1, duration: durationNormal)

    /// `easeStandard` at `durationFast` — the pairing used by hover/color
    /// transitions across the web components (e.g. `Button.jsx`'s
    /// `transition: background var(--dur-fast) var(--ease-standard)...`).
    static let fastStandard = Animation.timingCurve(0.2, 0.6, 0.2, 1, duration: durationFast)
    /// `easeStandard` at `durationInstant` — used for the button press scale.
    static let instantStandard = Animation.timingCurve(0.2, 0.6, 0.2, 1, duration: durationInstant)
}

extension View {
    /// Applies `animation` unless Reduce Motion is enabled.
    func pvAnimation<V: Equatable>(_ animation: Animation?, value: V) -> some View {
        modifier(PVAnimationModifier(animation: animation, value: value))
    }
}

private struct PVAnimationModifier<V: Equatable>: ViewModifier {
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    let animation: Animation?
    let value: V

    func body(content: Content) -> some View {
        content.animation(reduceMotion ? nil : animation, value: value)
    }
}
