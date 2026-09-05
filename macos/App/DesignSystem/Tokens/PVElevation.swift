import SwiftUI

/// One shadow layer: color/radius/offset, mirroring a single CSS
/// `box-shadow` layer. `PVElevation` levels above `sm` compose two of
/// these (matching the two-layer CSS values), applied via `.pvShadow(_:)`.
struct PVShadowLayer {
    let color: Color
    let radius: CGFloat
    let x: CGFloat
    let y: CGFloat
}

/// Provenencia design-system elevation tokens.
///
/// Transcribed from `tokens/elevation.css`: "Shadows are warm and shallow —
/// paper on paper, not glass in space." Dark mode drops the warm tint and
/// uses plain black at higher opacity (also transcribed, not approximated).
enum PVElevation {
    static let hairlineWidth: CGFloat = 1

    static let sm: [PVShadowLayer] = [
        layer(light: (0x2B2620, 0.05), dark: (0x000000, 0.4), radius: 1, x: 0, y: 1)
    ]
    static let md: [PVShadowLayer] = [
        layer(light: (0x2B2620, 0.06), dark: (0x000000, 0.5), radius: 2, x: 0, y: 1),
        layer(light: (0x2B2620, 0.08), dark: (0x000000, 0.5), radius: 10, x: 0, y: 4)
    ]
    static let lg: [PVShadowLayer] = [
        layer(light: (0x2B2620, 0.06), dark: (0x000000, 0.62), radius: 4, x: 0, y: 2),
        layer(light: (0x2B2620, 0.14), dark: (0x000000, 0.62), radius: 28, x: 0, y: 12)
    ]
    static let overlay: [PVShadowLayer] = [
        layer(light: (0x1B1712, 0.34), dark: (0x000000, 0.7), radius: 60, x: 0, y: 24),
        layer(light: (0x1B1712, 0.1), dark: (0x000000, 0.7), radius: 6, x: 0, y: 2)
    ]

    /// CSS `--shadow-inset` has no direct SwiftUI equivalent (there's no
    /// inset variant of `.shadow()`) — see `View.pvInsetShadow(cornerRadius:)`
    /// below for the approximation used at call sites like `PVInput`.
    static let inset = layer(light: (0x2B2620, 0.07), dark: (0x000000, 0.4), radius: 2, x: 0, y: 1)

    /// `--ring-focus`: a 3pt solid ring in `PVColor.borderFocus` at 32%
    /// opacity, applied via `View.pvFocusRing(cornerRadius:)`.
    static let focusRingWidth: CGFloat = 3
    static let focusRingOpacity: Double = 0.32

    private static func layer(
        light: (UInt32, Double),
        dark: (UInt32, Double),
        radius: CGFloat,
        x: CGFloat,
        y: CGFloat
    ) -> PVShadowLayer {
        let color = Color.pvDynamic(
            light: colorWithOpacity(light.0, light.1),
            dark: colorWithOpacity(dark.0, dark.1)
        )
        return PVShadowLayer(color: color, radius: radius, x: x, y: y)
    }

    private static func colorWithOpacity(_ rgb: UInt32, _ opacity: Double) -> Color {
        let r = Double((rgb & 0xFF0000) >> 16) / 255
        let g = Double((rgb & 0x00FF00) >> 8) / 255
        let b = Double(rgb & 0x0000FF) / 255
        return Color(.sRGB, red: r, green: g, blue: b, opacity: opacity)
    }
}

extension View {
    /// Applies a `PVElevation` level (one or two composited shadow layers).
    func pvShadow(_ layers: [PVShadowLayer]) -> some View {
        layers.reduce(AnyView(self)) { view, layer in
            AnyView(view.shadow(color: layer.color, radius: layer.radius, x: layer.x, y: layer.y))
        }
    }

    /// Approximates CSS `--shadow-inset` (a shadow cast *into* the surface,
    /// e.g. a text field's resting state) by overlaying a soft dark stroke
    /// along the shape's top edge, masked to the shape.
    func pvInsetShadow(cornerRadius: CGFloat) -> some View {
        let layer = PVElevation.inset
        return overlay(
            RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
                .stroke(layer.color, lineWidth: layer.radius)
                .blur(radius: layer.radius / 2)
                .offset(y: layer.y)
                .mask(RoundedRectangle(cornerRadius: cornerRadius, style: .continuous))
        )
    }

    /// Approximates CSS `--ring-focus` (a solid, non-blurred ring outside
    /// the element on focus) as a stroked overlay in `PVColor.borderFocus`.
    func pvFocusRing(cornerRadius: CGFloat) -> some View {
        overlay(
            RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
                .strokeBorder(PVColor.borderFocus.opacity(PVElevation.focusRingOpacity), lineWidth: PVElevation.focusRingWidth)
        )
    }
}
