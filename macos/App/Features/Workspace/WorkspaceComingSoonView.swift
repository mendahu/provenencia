import SwiftUI

/// Dashed coming-soon placeholder for unbuilt workspace destinations (S5-07).
struct WorkspaceComingSoonView: View {
    let icon: PVSymbol
    let title: LocalizedStringResource
    let message: LocalizedStringResource

    var body: some View {
        VStack {
            Spacer(minLength: 0)
            PVEmptyState(
                icon: icon,
                title: title,
                message: String(localized: message)
            )
            .padding(.horizontal, PVSpacing.space9)
            Spacer(minLength: 0)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
    }
}
