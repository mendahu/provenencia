import SwiftUI

/// Tone for `PVBadge` — mirrors `components/core/Badge.jsx`'s `tone` prop.
enum PVBadgeTone {
    case neutral, accent, success, warning, danger, info

    fileprivate var colors: (background: Color, foreground: Color, line: Color) {
        switch self {
        case .neutral: (PVColor.surfaceInset, PVColor.textSecondary, PVColor.borderDefault)
        case .accent: (PVColor.accentSoft, PVColor.accentSoftForeground, PVColor.accentLine)
        case .success: (PVColor.successSoft, PVColor.successForeground, PVColor.success)
        case .warning: (PVColor.warningSoft, PVColor.warningForeground, PVColor.warning)
        case .danger: (PVColor.dangerSoft, PVColor.dangerForeground, PVColor.danger)
        case .info: (PVColor.infoSoft, PVColor.infoForeground, PVColor.info)
        }
    }
}

/// A small status/label pill — mirrors `components/core/Badge.jsx` (tone
/// fill, optional leading icon, uppercase micro caption). `subtle` swaps
/// the filled background for a bordered outline, same as the web prop.
struct PVBadge: View {
    private let label: Text?
    private let tone: PVBadgeTone
    private let icon: PVSymbol?
    private let subtle: Bool
    /// Only set for the icon-only badge, which has no text to read out.
    private let iconLabel: LocalizedStringResource?

    /// For fixed UI copy (e.g. "provenencia", "you").
    init(_ titleKey: LocalizedStringResource, tone: PVBadgeTone = .neutral, icon: PVSymbol? = nil, subtle: Bool = false) {
        label = Text(titleKey)
        self.tone = tone
        self.icon = icon
        self.subtle = subtle
        iconLabel = nil
    }

    /// For badge content that is data, not UI copy (e.g. a raw `plugin:…` origin id).
    init(text: String, tone: PVBadgeTone = .neutral, icon: PVSymbol? = nil, subtle: Bool = false) {
        label = Text(text)
        self.tone = tone
        self.icon = icon
        self.subtle = subtle
        iconLabel = nil
    }

    /// Glyph-only pill — `PVIcon` is `accessibilityHidden`, so `label` carries
    /// the meaning for VoiceOver and doubles as the hover tooltip.
    init(icon: PVSymbol, label: LocalizedStringResource, tone: PVBadgeTone = .neutral, subtle: Bool = false) {
        self.label = nil
        self.tone = tone
        self.icon = icon
        self.subtle = subtle
        iconLabel = label
    }

    var body: some View {
        let colors = tone.colors
        HStack(spacing: 5) {
            if let icon {
                PVIcon(icon, size: 11)
            }
            label?
                .tracking(PVTypeScale.micro * PVTracking.caps)
                .textCase(.uppercase)
        }
        .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
        .foregroundStyle(colors.foreground)
        .padding(.horizontal, label == nil ? PVSpacing.space2 : PVSpacing.space4)
        .padding(.vertical, PVSpacing.space1)
        .background(subtle ? Color.clear : colors.background)
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous)
                .stroke(subtle ? colors.line : .clear, lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous))
        .fixedSize()
        .modifier(PVBadgeIconLabel(label: iconLabel))
    }
}

/// Applies the VoiceOver label + tooltip a glyph-only badge needs, and is a
/// no-op for the text badges (which read out their own label).
private struct PVBadgeIconLabel: ViewModifier {
    let label: LocalizedStringResource?

    func body(content: Content) -> some View {
        if let label {
            content
                .accessibilityElement()
                .accessibilityLabel(Text(label))
                .help(Text(label))
        } else {
            content
        }
    }
}

#Preview {
    HStack(spacing: PVSpacing.space5) {
        PVBadge("text", tone: .neutral, icon: .textType, subtle: true)
        PVBadge("date", tone: .info, icon: .calendar, subtle: true)
        PVBadge("provenencia", tone: .accent)
        PVBadge("you", tone: .warning)
        PVBadge(text: "plugin:findagrave", tone: .info)
        PVBadge(icon: .shieldCheck, label: "Seeded by Provenencia", tone: .accent, subtle: true)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
