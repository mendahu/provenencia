import SwiftUI

struct OnboardingFooter<Leading: View, Trailing: View>: View {
    var isBusy: Bool
    @ViewBuilder var leading: Leading
    @ViewBuilder var trailing: Trailing

    var body: some View {
        HStack(spacing: PVSpacing.space5) {
            leading
            Spacer()
            if isBusy {
                ProgressView()
                    .controlSize(.small)
                    .tint(PVColor.accent)
            }
            trailing
        }
    }
}
