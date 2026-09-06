import SwiftUI

/// Whether a divider is drawn above an `OnboardingContributorRow`, and in
/// what style — solid between ordinary rows, dashed above the trailing
/// "not listed" row (the design system's "absence of evidence" convention).
enum OnboardingContributorRowDivider {
    case none, solid, dashed
}

/// A single selectable row in the "who are you in this file?" contributor
/// list. Bespoke to Onboarding, not a shared `DesignSystem/Components` file
/// — the source design system's own mockup hand-styles these rows as
/// page-specific markup too (it doesn't reach for its own `Radio`
/// component here), and `DesignSystem/README.md` lists `Radio` as
/// "add on demand." Only reaches for `PV*` tokens.
struct OnboardingContributorRow: View {
    let title: String
    var subtitle: String?
    let selected: Bool
    var divider: OnboardingContributorRowDivider = .none
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: PVSpacing.space5) {
                Circle()
                    .strokeBorder(selected ? PVColor.accent : PVColor.borderStrong, lineWidth: selected ? 4 : 1)
                    .background(Circle().fill(selected ? PVColor.surfaceCard : PVColor.surfaceRaised))
                    .frame(width: 13, height: 13)
                Text(title)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: selected ? PVFontWeight.semibold : PVFontWeight.regular))
                    .foregroundStyle(PVColor.textPrimary)
                if let subtitle {
                    Text(subtitle)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                }
                Spacer(minLength: 0)
            }
            .padding(.horizontal, PVSpacing.space5)
            .padding(.vertical, PVSpacing.space4)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(selected ? PVColor.surfaceSelected : Color.clear)
            .overlay(alignment: .leading) {
                if selected {
                    Rectangle().fill(PVColor.accent).frame(width: 2)
                }
            }
            .overlay(alignment: .top) {
                switch divider {
                case .none:
                    EmptyView()
                case .solid:
                    Rectangle().fill(PVColor.borderSubtle).frame(height: 1)
                case .dashed:
                    OnboardingHorizontalLine()
                        .stroke(PVColor.borderDefault, style: StrokeStyle(lineWidth: 1, dash: [4, 3]))
                        .frame(height: 1)
                }
            }
        }
        .buttonStyle(.plain)
    }
}

/// A single horizontal line spanning its frame's width — used with a dashed
/// `StrokeStyle` for the "not listed" row's divider (dashed = absence of
/// evidence, per the design system's border-weight convention).
private struct OnboardingHorizontalLine: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        path.move(to: CGPoint(x: rect.minX, y: rect.midY))
        path.addLine(to: CGPoint(x: rect.maxX, y: rect.midY))
        return path
    }
}
