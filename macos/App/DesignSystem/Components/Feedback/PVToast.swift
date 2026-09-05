import SwiftUI

/// A tone. Matches the web `tone` prop — mirrors `components/feedback/Toast.jsx`,
/// read directly out of the compiled component bundle: each tone pairs a
/// semantic color with a Lucide icon (`info`, `check-check`, `triangle-alert`,
/// `octagon-alert`), substituted here to their SF Symbol equivalents in
/// `PVIconMap`.
enum PVToastTone {
    case info, success, warning, danger

    fileprivate var color: Color {
        switch self {
        case .info: return PVColor.info
        case .success: return PVColor.success
        case .warning: return PVColor.warning
        case .danger: return PVColor.danger
        }
    }

    fileprivate var iconName: String {
        switch self {
        case .info: return "info"
        case .success: return "check-check"
        case .warning: return "triangle-alert"
        case .danger: return "octagon-alert"
        }
    }
}

/// A transient notification — mirrors `components/feedback/Toast.jsx`
/// (tone-colored left border, `shadow-lg`, optional title/message/dismiss).
/// `title`/`message` are plain `String` rather than `LocalizedStringResource`
/// because a toast's content is always dynamic (an error, a status update),
/// not fixed UI chrome — callers that do have fixed copy should still route
/// it through `L10n` before passing it in. The web spec's `action` slot
/// (an inline button inside the toast) isn't ported — nothing consumes it
/// yet; add it here, following `PVButton`'s pattern, if a future screen
/// needs one.
///
/// See `DesignSystem/README.md` / `PVButton.swift` for the component
/// pattern this follows.
///
/// Auto-dismisses after `autoDismissAfter` (default 5 seconds) by calling
/// `onDismiss`, unless it's `nil` — pass `autoDismissAfter: nil` to keep a
/// toast on screen until the user (or the caller) dismisses it explicitly.
struct PVToast: View {
    static let defaultAutoDismissDelay: TimeInterval = 5

    let tone: PVToastTone
    var title: String?
    var message: String?
    var onDismiss: (() -> Void)?
    var autoDismissAfter: TimeInterval? = PVToast.defaultAutoDismissDelay

    var body: some View {
        HStack(alignment: .top, spacing: PVSpacing.space5) {
            PVIcon(tone.iconName, size: 15)
                .foregroundStyle(tone.color)
                .padding(.top, 2)
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                if let title {
                    Text(title)
                        .font(PVFont.display(size: PVTypeScale.h4))
                        .foregroundStyle(PVColor.textDisplay)
                }
                if let message {
                    Text(message)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            if let onDismiss {
                Button(action: onDismiss) {
                    PVIcon("x", size: 13)
                        .foregroundStyle(PVColor.textSecondary)
                }
                .buttonStyle(.plain)
                .accessibilityLabel(Text(L10n.DesignSystem.toastDismiss))
            }
        }
        .padding(.vertical, PVSpacing.space5)
        .padding(.horizontal, PVSpacing.space6)
        .frame(maxWidth: 360, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .fill(PVColor.surfaceRaised)
        )
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
        )
        .overlay(alignment: .leading) {
            Rectangle().fill(tone.color).frame(width: 3)
        }
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
        .pvShadow(PVElevation.lg)
        .accessibilityElement(children: .combine)
        .task {
            guard let autoDismissAfter else { return }
            try? await Task.sleep(for: .seconds(autoDismissAfter))
            onDismiss?()
        }
    }
}

#Preview {
    VStack(spacing: PVSpacing.space5) {
        PVToast(tone: .danger, message: "Project already open.", onDismiss: {})
        PVToast(tone: .info, title: "Heads up", message: "This project was last opened on another Mac.")
        PVToast(tone: .success, message: "Project created.", autoDismissAfter: nil)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
