import SwiftUI

/// Tone for `PVCallout` — mirrors `components/feedback/Callout.jsx`'s `tone` prop.
enum PVCalloutTone {
    case info, success, warning, danger, neutral

    fileprivate var color: Color {
        switch self {
        case .info: PVColor.info
        case .success: PVColor.success
        case .warning: PVColor.warning
        case .danger: PVColor.danger
        case .neutral: PVColor.borderStrong
        }
    }

    fileprivate var soft: Color {
        switch self {
        case .info: PVColor.infoSoft
        case .success: PVColor.successSoft
        case .warning: PVColor.warningSoft
        case .danger: PVColor.dangerSoft
        case .neutral: PVColor.surfaceSunken
        }
    }

    fileprivate var foreground: Color {
        switch self {
        case .info: PVColor.infoForeground
        case .success: PVColor.successForeground
        case .warning: PVColor.warningForeground
        case .danger: PVColor.dangerForeground
        case .neutral: PVColor.textSecondary
        }
    }

    fileprivate var defaultIcon: PVSymbol {
        switch self {
        case .danger: .danger
        case .warning: .warning
        case .success: .success
        case .info, .neutral: .info
        }
    }
}

/// An inline status/explanation banner — mirrors `components/feedback/Callout.jsx`
/// (tone-tinted fill, 3pt left rule, optional title, optional actions row under the
/// body). `onDismiss`, `detail`, and `plain` variant remain deferred until a call
/// site needs them.
struct PVCallout<Actions: View>: View {
    private let tone: PVCalloutTone
    private let icon: PVSymbol?
    private let title: LocalizedStringResource?
    private let message: String
    private let compact: Bool
    private let actions: Actions

    init(
        tone: PVCalloutTone = .info,
        icon: PVSymbol? = nil,
        title: LocalizedStringResource? = nil,
        message: String,
        compact: Bool = false,
        @ViewBuilder actions: () -> Actions
    ) {
        self.tone = tone
        self.icon = icon
        self.title = title
        self.message = message
        self.compact = compact
        self.actions = actions()
    }

    var body: some View {
        HStack(alignment: .top, spacing: compact ? PVSpacing.space5 : PVSpacing.space6) {
            PVIcon(icon ?? tone.defaultIcon, size: compact ? 14 : 15)
                .foregroundStyle(tone.color)
                .padding(.top, 2)
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                if let title {
                    Text(title)
                        .font(PVFont.display(size: compact ? PVTypeScale.h4 : PVTypeScale.h3, weight: PVFontWeight.semibold))
                        .foregroundStyle(tone.foreground)
                }
                Text(message)
                    .font(PVFont.body(size: compact ? PVTypeScale.caption : PVTypeScale.bodySmall))
                    .foregroundStyle(tone.foreground.opacity(0.92))
                    .fixedSize(horizontal: false, vertical: true)
                if Actions.self != EmptyView.self {
                    HStack(spacing: PVSpacing.space5) {
                        actions
                    }
                    .padding(.top, PVSpacing.space5)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(.vertical, compact ? PVSpacing.space5 : PVSpacing.space6)
        .padding(.horizontal, compact ? PVSpacing.space6 : PVSpacing.space7)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .fill(tone.soft)
        )
        .overlay(alignment: .leading) {
            RoundedRectangle(cornerRadius: PVRadius.xs / 2)
                .fill(tone.color)
                .frame(width: 3)
        }
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
    }
}

extension PVCallout where Actions == EmptyView {
    init(
        tone: PVCalloutTone = .info,
        icon: PVSymbol? = nil,
        title: LocalizedStringResource? = nil,
        message: String,
        compact: Bool = false
    ) {
        self.init(
            tone: tone,
            icon: icon,
            title: title,
            message: message,
            compact: compact
        ) {
            EmptyView()
        }
    }
}

#Preview {
    VStack(spacing: PVSpacing.space6) {
        PVCallout(
            tone: .neutral, icon: .lock,
            message: "Seeded by Provenencia. Its label, data type, and description are fixed.",
            compact: true
        )
        PVCallout(tone: .info, title: "Heads up", message: "This project was last opened on another Mac.")
        PVCallout(
            tone: .danger,
            title: "No artifact linked to this record",
            message: "The evidence graph needs a source before this fact can be graded."
        ) {
            PVButton("Add an Artifact", variant: .secondary, size: .sm) {}
        }
        PVCallout(
            tone: .danger,
            title: "Two sources give different death dates",
            message: "Provenencia keeps this fact at disputed until the conflict is resolved in writing."
        ) {
            PVButton("Record a conclusion", variant: .secondary, size: .sm) {}
            PVButton("View both sources", variant: .link, size: .sm) {}
        }
        PVCallout(
            tone: .info,
            message: "Index entries are derivative sources.",
            compact: true
        ) {
            PVButton("Order the original", variant: .link, size: .sm) {}
        }
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
