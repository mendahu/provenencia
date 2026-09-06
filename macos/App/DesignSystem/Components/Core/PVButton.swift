import SwiftUI

/// Visual weight for `PVButton` / `PVButtonStyle`.
enum PVButtonVariant {
    case primary, secondary, ghost
}

/// Shared control size for buttons, inputs, and selects.
enum PVControlSize {
    case sm, md, lg

    var height: CGFloat {
        switch self {
        case .sm: return PVSpacing.controlHeightSmall
        case .md: return PVSpacing.controlHeightMedium
        case .lg: return PVSpacing.controlHeightLarge
        }
    }

    var horizontalPadding: CGFloat {
        switch self {
        case .sm: return 10
        case .md: return 14
        case .lg: return 20
        }
    }

    var font: Font {
        self == .sm ? PVFont.body(size: PVTypeScale.caption) : PVFont.body(size: PVTypeScale.bodySmall)
    }

    var buttonFont: Font {
        switch self {
        case .sm: return PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium)
        case .md: return PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium)
        case .lg: return PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.medium)
        }
    }
}

/// Provenencia button chrome. Prefer `.buttonStyle(.pv(.primary))` on any `Button`;
/// `PVButton` remains a label convenience for `LocalizedStringResource` titles.
struct PVButtonStyle: ButtonStyle {
    var variant: PVButtonVariant = .secondary
    var size: PVControlSize = .md

    func makeBody(configuration: Configuration) -> some View {
        PVButtonBody(configuration: configuration, variant: variant, size: size)
    }
}

extension ButtonStyle where Self == PVButtonStyle {
    static func pv(_ variant: PVButtonVariant = .secondary, size: PVControlSize = .md) -> PVButtonStyle {
        PVButtonStyle(variant: variant, size: size)
    }
}

/// A styled push button. Always takes a `LocalizedStringResource` label.
struct PVButton: View {
    private let titleKey: LocalizedStringResource
    private let variant: PVButtonVariant
    private let size: PVControlSize
    private let action: () -> Void

    init(
        _ titleKey: LocalizedStringResource,
        variant: PVButtonVariant = .secondary,
        size: PVControlSize = .md,
        action: @escaping () -> Void
    ) {
        self.titleKey = titleKey
        self.variant = variant
        self.size = size
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            Text(titleKey)
        }
        .buttonStyle(.pv(variant, size: size))
    }
}

private struct PVButtonPalette {
    let background: Color
    let hoverBackground: Color?
    let foreground: Color
    let border: Color
    let hoverBorder: Color?

    static func palette(for variant: PVButtonVariant) -> PVButtonPalette {
        switch variant {
        case .primary:
            return PVButtonPalette(
                background: PVColor.accent, hoverBackground: PVColor.accentHover,
                foreground: PVColor.accentForeground,
                border: PVColor.accent, hoverBorder: PVColor.accentHover
            )
        case .secondary:
            return PVButtonPalette(
                background: PVColor.surfaceRaised, hoverBackground: PVColor.surfaceSunken,
                foreground: PVColor.textPrimary,
                border: PVColor.borderDefault, hoverBorder: PVColor.borderStrong
            )
        case .ghost:
            return PVButtonPalette(
                background: .clear, hoverBackground: PVColor.surfaceHover,
                foreground: PVColor.textSecondary,
                border: .clear, hoverBorder: nil
            )
        }
    }
}

private struct PVButtonBody: View {
    let configuration: ButtonStyleConfiguration
    let variant: PVButtonVariant
    let size: PVControlSize

    @State private var isHovering = false
    @Environment(\.isEnabled) private var isEnabled
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    var body: some View {
        let palette = PVButtonPalette.palette(for: variant)
        let showHover = isHovering && isEnabled

        configuration.label
            .font(size.buttonFont)
            .foregroundStyle(palette.foreground)
            .padding(.horizontal, size.horizontalPadding)
            .frame(height: size.height)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .fill(showHover ? (palette.hoverBackground ?? palette.background) : palette.background)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(showHover ? (palette.hoverBorder ?? palette.border) : palette.border, lineWidth: 1)
            )
            .opacity(isEnabled ? 1 : 0.45)
            .scaleEffect(configuration.isPressed && !reduceMotion ? PVMotion.pressScale : 1)
            .pvAnimation(PVMotion.fastStandard, value: isHovering)
            .pvAnimation(PVMotion.instantStandard, value: configuration.isPressed)
            .onHover { isHovering = $0 }
    }
}

#Preview {
    VStack(spacing: PVSpacing.space5) {
        PVButton("Continue", variant: .primary, size: .md) {}
        Button("Custom label") {}
            .buttonStyle(.pv(.secondary, size: .sm))
        PVButton("Back", variant: .ghost, size: .md) {}
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
