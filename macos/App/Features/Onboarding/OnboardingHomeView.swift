import SwiftUI

struct OnboardingHomeView: View {
    @Bindable var model: OnboardingModel

    var body: some View {
        VStack(spacing: 12) {
            Text(L10n.Onboarding.signedInTitle)
                .font(.title2)
                .accessibilityIdentifier("onboarding.home")
            Text(model.sessionDisplayName)
                .font(.title3)
                .accessibilityIdentifier("onboarding.home.displayName")
            if !model.projectBasename.isEmpty {
                Text(model.projectBasename)
                    .foregroundStyle(.secondary)
                    .accessibilityIdentifier("onboarding.home.project")
            }
            Button(L10n.Onboarding.signOut) {
                Task { await model.signOut() }
            }
            .accessibilityIdentifier("onboarding.signOut")
        }
        .padding(32)
    }
}
