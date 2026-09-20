import SwiftUI

/// A "nothing here yet" placeholder — mirrors `components/feedback/EmptyState.jsx`
/// (dashed border, faint icon, display-font title, muted prose body, optional
/// `action` slot). `message` is a plain `String` (like `PVToast`'s) since
/// callers often interpolate dynamic content (e.g. a search query) into it;
/// pass `String(localized: …)` for fixed copy.
struct PVEmptyState<Action: View>: View {
    private let icon: PVSymbol
    private let title: LocalizedStringResource?
    private let message: String?
    private let compact: Bool
    private let action: Action

    init(
        icon: PVSymbol,
        title: LocalizedStringResource? = nil,
        message: String? = nil,
        compact: Bool = false,
        @ViewBuilder action: () -> Action
    ) {
        self.icon = icon
        self.title = title
        self.message = message
        self.compact = compact
        self.action = action()
    }

    var body: some View {
        VStack(spacing: PVSpacing.space5) {
            PVIcon(icon, size: compact ? 20 : 26)
                .foregroundStyle(PVColor.textFaint)
            if let title {
                Text(title)
                    .font(PVFont.display(size: PVTypeScale.h3))
                    .foregroundStyle(PVColor.textDisplay)
            }
            if let message {
                Text(message)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textMuted)
                    .multilineTextAlignment(.center)
                    .frame(maxWidth: PVSpacing.measureNarrow)
            }
            action
        }
        .padding(.vertical, compact ? PVSpacing.space9 : PVSpacing.space12)
        .padding(.horizontal, compact ? PVSpacing.space8 : PVSpacing.space9)
        .frame(maxWidth: .infinity)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .fill(PVColor.surfaceSunken)
        )
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .strokeBorder(PVColor.borderDefault, style: StrokeStyle(lineWidth: 1, dash: [4, 3]))
        )
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
    }
}

extension PVEmptyState where Action == EmptyView {
    init(
        icon: PVSymbol,
        title: LocalizedStringResource? = nil,
        message: String? = nil,
        compact: Bool = false
    ) {
        self.init(icon: icon, title: title, message: message, compact: compact) {
            EmptyView()
        }
    }
}

#Preview {
    VStack(spacing: PVSpacing.space9) {
        PVEmptyState(
            icon: .tag,
            title: "No source fields yet",
            message: "This project has no metadata vocabulary. Add the fields your records actually carry."
        )
        PVEmptyState(
            icon: .photo,
            title: "No artifacts yet",
            message: "Attach a scan, photo, or transcript of this source.",
            compact: true
        ) {
            PVButton("Add artifact", variant: .primary, size: .sm, icon: .plus) {}
        }
        PVEmptyState(icon: .searchEmpty, title: "No field selected", compact: true)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
