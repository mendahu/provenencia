import SwiftUI

/// A link that keeps the page's ink: body-sm type in primary (or secondary) ink,
/// underlined at 35%, link ink on hover and press. It wraps, never truncates, and
/// Return activates it as Space does; its caller names it for VoiceOver and tests. A record's citation line and a Source's title
/// are its callers. Its label sets no ink of its own: a plain `Text`, or a
/// `PVTitleText` in `.rowLink`, takes the link's ink and font. Content-agnostic.
struct PVInkLink<Label: View>: View {
    enum Ink {
        case primary
        case secondary

        fileprivate var color: Color {
            switch self {
            case .primary: PVColor.textPrimary
            case .secondary: PVColor.textSecondary
            }
        }
    }

    let ink: Ink
    let action: () -> Void
    @ViewBuilder var label: () -> Label

    var body: some View {
        Button(action: action) {
            label()
                .multilineTextAlignment(.leading)
                .fixedSize(horizontal: false, vertical: true)
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .buttonStyle(PVInkLinkStyle(ink: ink.color))
        .pvActivatesOnReturn(when: true, action)
    }
}

/// Kit hover and press on the link: link ink on hover, else the caller's ink.
private struct PVInkLinkStyle: ButtonStyle {
    let ink: Color

    func makeBody(configuration: Configuration) -> some View {
        PVInkLinkBody(configuration: configuration, ink: ink)
    }
}

private struct PVInkLinkBody: View {
    let configuration: ButtonStyleConfiguration
    let ink: Color

    /// Read here, in the button's focus scope (see `PVButtonBody`).
    @Environment(\.isFocused) private var isFocused

    var body: some View {
        PVHoverEffect(isPressed: configuration.isPressed, hoverAnimation: PVMotion.fastStandard) { showHover in
            let color = showHover || configuration.isPressed ? PVColor.textLinkHover : ink
            configuration.label
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(color)
                .underline(true, color: color.opacity(0.35))
                .pvFocusRing(isFocused, cornerRadius: PVRadius.xs)
        }
    }
}

#Preview("Ink link") {
    VStack(alignment: .leading, spacing: PVSpacing.space4) {
        PVInkLink(ink: .primary, action: {}) {
            Text(verbatim: "Hamburg census 1867")
        }
        .accessibilityIdentifier("preview.primary")
        PVInkLink(ink: .secondary, action: {}) {
            Text(verbatim: "Hamburger Adressbuch 1861, a longer title that wraps when the column is narrow")
        }
        .accessibilityLabel(Text(verbatim: "Open the Adressbuch"))
        .accessibilityIdentifier("preview.secondary")
    }
    .frame(width: 240)
    .padding()
    .background(PVColor.surfacePage)
}
