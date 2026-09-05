import SwiftUI

// MARK: - `PVButton` — Provenencia Design System / Core
//
// This file is the **canonical example** for every component under
// `DesignSystem/Components/`. Read this comment before adding component
// #2 (`PVIcon`, `PVField`, `PVInput`, and `PVSelect` all follow the same
// shape, with lighter comments pointing back here); it's also written up
// in `DesignSystem/README.md` under "The component pattern."
//
// **What this mirrors**: the web design system's `components/core/Button.jsx`
// (variants `primary`/`secondary`/`ghost`/`danger`/`link`, sizes
// `sm`/`md`/`lg`) — read directly out of the compiled component bundle
// during planning, so the tables below are transcribed, not guessed.
// `danger`/`link` are included for parity even though the Onboarding Flow
// board (the only screens this scaffolding targets so far) only uses
// `primary`/`secondary`/`ghost`. The web prop surface also includes
// `iconLeft`/`iconRight`/`loading`/`fullWidth`, which aren't needed by
// anything in scope yet — add them here if a future screen needs one,
// following `PVIcon` for how a leading/trailing icon slot should look.
//
// **The pattern**:
// - File lives at `Components/<Category>/PV<Name>.swift`, where
//   `<Category>` mirrors the web system's `components/<category>/` folder
//   (`core`, `forms`, `navigation`, `feedback`, `research`). `PVButton` is
//   `core/Button.jsx` → `Components/Core/PVButton.swift`.
// - The type only ever reaches for `PVColor`/`PVFont`/`PVSpacing`/
//   `PVRadius`/`PVElevation`/`PVMotion` — never a literal color, font, or
//   size pulled from thin air.
// - User-facing text is always a `LocalizedStringResource`, never a raw
//   `String` (`docs/macos-client-patterns.md` §6); AppKit stays inside a
//   small wrapper, not scattered through the view (§4).
// - `.accessibilityIdentifier` is left to the call site
//   (`docs/macos-client-patterns.md` §5) — components don't invent their
//   own identifiers.
// - Hover/pressed state lives in a private `ButtonStyle` + backing `View`
//   below the public type, not on the public type itself — see
//   `PVButtonBody`.

/// Visual weight. Matches the web `variant` prop.
enum PVButtonVariant {
    case primary, secondary, ghost, danger, link
}

/// Control size. Matches the web `size` prop; heights come from
/// `PVSpacing`'s macOS-native control heights, not the web pixel values —
/// see `DesignSystem/README.md` "Platform deviations".
enum PVButtonSize {
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
        switch self {
        case .sm: return PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium)
        case .md: return PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium)
        case .lg: return PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.medium)
        }
    }
}

/// A styled push button. Always takes a `LocalizedStringResource` label —
/// see the pattern notes above.
struct PVButton: View {
    private let titleKey: LocalizedStringResource
    private let variant: PVButtonVariant
    private let size: PVButtonSize
    private let action: () -> Void

    init(
        _ titleKey: LocalizedStringResource,
        variant: PVButtonVariant = .secondary,
        size: PVButtonSize = .md,
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
        .buttonStyle(PVButtonStyle(variant: variant, size: size))
    }
}

/// Rest/hover colors for one variant. `danger`/`link` reach for raw
/// `PVPalette` values because `tokens/colors.css` doesn't define semantic
/// aliases for a couple of the literals the web `Button.jsx` uses directly
/// (`--madder-700` on danger hover, `--paper-0` as danger's fixed
/// foreground) — everything else goes through `PVColor`.
private struct PVButtonPalette {
    let background: Color
    let hoverBackground: Color?
    let foreground: Color
    let border: Color
    let hoverBorder: Color?
    let underline: Bool

    static func palette(for variant: PVButtonVariant) -> PVButtonPalette {
        switch variant {
        case .primary:
            return PVButtonPalette(
                background: PVColor.accent, hoverBackground: PVColor.accentHover,
                foreground: PVColor.accentForeground,
                border: PVColor.accent, hoverBorder: PVColor.accentHover,
                underline: false
            )
        case .secondary:
            return PVButtonPalette(
                background: PVColor.surfaceRaised, hoverBackground: PVColor.surfaceSunken,
                foreground: PVColor.textPrimary,
                border: PVColor.borderDefault, hoverBorder: PVColor.borderStrong,
                underline: false
            )
        case .ghost:
            return PVButtonPalette(
                background: .clear, hoverBackground: PVColor.surfaceHover,
                foreground: PVColor.textSecondary,
                border: .clear, hoverBorder: nil,
                underline: false
            )
        case .danger:
            return PVButtonPalette(
                background: PVColor.danger, hoverBackground: PVPalette.madder700,
                foreground: PVPalette.paper0,
                border: PVColor.danger, hoverBorder: PVPalette.madder700,
                underline: false
            )
        case .link:
            return PVButtonPalette(
                background: .clear, hoverBackground: nil,
                foreground: PVColor.textLink,
                border: .clear, hoverBorder: nil,
                underline: true
            )
        }
    }
}

private struct PVButtonStyle: ButtonStyle {
    let variant: PVButtonVariant
    let size: PVButtonSize

    func makeBody(configuration: Configuration) -> some View {
        PVButtonBody(configuration: configuration, variant: variant, size: size)
    }
}

/// Backs `PVButtonStyle`. Hover state has to live on a `View` (via
/// `@State`), not directly on the `ButtonStyle` value itself.
private struct PVButtonBody: View {
    let configuration: ButtonStyleConfiguration
    let variant: PVButtonVariant
    let size: PVButtonSize

    @State private var isHovering = false
    @Environment(\.isEnabled) private var isEnabled

    var body: some View {
        let palette = PVButtonPalette.palette(for: variant)
        let showHover = isHovering && isEnabled && variant != .link
        let isLink = variant == .link

        configuration.label
            .font(size.font)
            .foregroundStyle(palette.foreground)
            .underline(palette.underline)
            .padding(.horizontal, isLink ? 0 : size.horizontalPadding)
            .frame(height: isLink ? nil : size.height)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .fill(showHover ? (palette.hoverBackground ?? palette.background) : palette.background)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(showHover ? (palette.hoverBorder ?? palette.border) : palette.border, lineWidth: 1)
            )
            .opacity(isEnabled ? 1 : 0.45)
            .scaleEffect(configuration.isPressed ? PVMotion.pressScale : 1)
            .animation(PVMotion.fastStandard, value: isHovering)
            .animation(PVMotion.instantStandard, value: configuration.isPressed)
            .onHover { isHovering = $0 }
    }
}

#Preview {
    VStack(spacing: PVSpacing.space5) {
        PVButton("Continue", variant: .primary, size: .md) {}
        PVButton("Continue", variant: .primary, size: .md) {}.disabled(true)
        PVButton("Choose…", variant: .secondary, size: .sm) {}
        PVButton("Back", variant: .ghost, size: .md) {}
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
