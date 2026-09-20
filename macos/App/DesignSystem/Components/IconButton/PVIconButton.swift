import SwiftUI

/// Tone for `PVIconButton` — `neutral` is the default chrome color,
/// `danger` tints the glyph for a destructive action (the web board sets
/// `style={{ color: 'var(--danger)' }}` on the delete button), and
/// `accent` is a filled primary control (metadata inline Save).
enum PVIconButtonTone {
    case neutral, danger, accent

    fileprivate var foreground: Color {
        switch self {
        case .neutral: PVColor.textSecondary
        case .danger: PVColor.danger
        case .accent: PVColor.accentForeground
        }
    }

    fileprivate var restingFill: Color {
        switch self {
        case .accent: PVColor.accent
        case .neutral, .danger: .clear
        }
    }

    fileprivate var hoverFill: Color {
        switch self {
        case .accent: PVColor.accentHover
        case .neutral, .danger: PVColor.surfaceHover
        }
    }
}

/// Provenencia icon-only button chrome (toolbar/chrome actions — the
/// sidebar's collapse toggle is the first call site). Follows
/// `PVButton.swift`'s pattern: hover/pressed state lives in a private
/// backing view (`PVIconButtonBody`), not on the public type.
///
/// `label` is required, matching `IconButton.jsx` — the glyph is decorative
/// (`PVIcon` is `accessibilityHidden`), so the label is the only thing a
/// screen reader has. It doubles as the hover tooltip, which is how the web
/// board's `Tooltip` wrapper is expressed on macOS (a native `.help()`
/// tooltip rather than a custom hover card).
struct PVIconButton: View {
    private let icon: PVSymbol
    private let label: LocalizedStringResource
    private let accessibilityLabel: LocalizedStringResource?
    private let size: PVControlSize
    private let tone: PVIconButtonTone
    private let action: () -> Void

    /// `label` is both the tooltip and, by default, what VoiceOver reads.
    /// Pass `accessibilityLabel` only when the two should differ — a tooltip
    /// can lean on what the user is already looking at ("Assign Author to
    /// this type"), where a spoken label has to name its target in full.
    init(
        _ icon: PVSymbol,
        label: LocalizedStringResource,
        accessibilityLabel: LocalizedStringResource? = nil,
        size: PVControlSize = .md,
        tone: PVIconButtonTone = .neutral,
        action: @escaping () -> Void
    ) {
        self.icon = icon
        self.label = label
        self.accessibilityLabel = accessibilityLabel
        self.size = size
        self.tone = tone
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            PVIcon(icon, size: size.iconGlyphSize)
        }
        .buttonStyle(PVIconButtonStyle(size: size, tone: tone))
        .accessibilityLabel(Text(accessibilityLabel ?? label))
        // `.help` on the button itself stops firing once it is disabled, and
        // the disabled tooltip is exactly where the reason lives — so the
        // hit area carries it instead.
        .contentShape(Rectangle())
        .help(Text(label))
    }
}

private struct PVIconButtonStyle: ButtonStyle {
    var size: PVControlSize = .md
    var tone: PVIconButtonTone = .neutral

    func makeBody(configuration: Configuration) -> some View {
        PVIconButtonBody(configuration: configuration, size: size, tone: tone)
    }
}

private struct PVIconButtonBody: View {
    let configuration: ButtonStyleConfiguration
    let size: PVControlSize
    let tone: PVIconButtonTone

    var body: some View {
        PVHoverEffect(isPressed: configuration.isPressed) { showHover in
            configuration.label
                .foregroundStyle(tone.foreground)
                .frame(width: size.height, height: size.height)
                .background(
                    RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                        .fill(showHover ? tone.hoverFill : tone.restingFill)
                )
        }
    }
}

#Preview {
    HStack(spacing: PVSpacing.space5) {
        PVIconButton(.sidebarToggle, label: "Collapse sidebar") {}
        PVIconButton(.dismiss, label: "Dismiss", size: .sm) {}
        PVIconButton(.check, label: "Save value", size: .sm, tone: .accent) {}
        PVIconButton(.account, label: "Account", size: .lg) {}
        PVIconButton(.trash, label: "Delete field", size: .sm, tone: .danger) {}
        PVIconButton(.trash, label: "In use on 3 sources", size: .sm, tone: .danger) {}
            .disabled(true)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
