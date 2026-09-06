import SwiftUI

struct OnboardingFileChoice: View {
    let title: LocalizedStringResource
    let subtitle: LocalizedStringResource
    let icon: PVSymbol
    let selected: Bool
    let action: () -> Void

    init(
        _ title: LocalizedStringResource,
        subtitle: LocalizedStringResource,
        icon: PVSymbol,
        selected: Bool,
        action: @escaping () -> Void
    ) {
        self.title = title
        self.subtitle = subtitle
        self.icon = icon
        self.selected = selected
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                HStack(spacing: PVSpacing.space4) {
                    PVIcon(icon, size: 16)
                    titleText
                    if selected {
                        Spacer(minLength: 0)
                        checkIcon
                    }
                }
                Text(subtitle)
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(selected ? PVColor.accentSoftForeground : PVColor.textSecondary)
                    .multilineTextAlignment(.leading)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(PVSpacing.space6)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .fill(selected ? PVColor.surfaceSelected : PVColor.surfaceCard)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .strokeBorder(selected ? PVColor.accent : PVColor.borderSubtle, lineWidth: selected ? 1.5 : 1)
            )
            .pvShadow(selected ? PVElevation.sm : [])
        }
        .buttonStyle(.plain)
    }

    private var titleText: some View {
        Text(title)
            .font(PVFont.body(size: PVTypeScale.h4, weight: PVFontWeight.semibold))
            .foregroundStyle(selected ? PVColor.accentSoftForeground : PVColor.textPrimary)
    }

    private var checkIcon: some View {
        PVIcon(.check, size: 15)
            .foregroundStyle(PVColor.accent)
    }
}
