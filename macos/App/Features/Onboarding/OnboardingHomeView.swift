import SwiftUI

struct OnboardingHomeView: View {
    @Bindable var model: OnboardingModel

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text(L10n.Onboarding.signedInTitle)
                .font(.title2)
                .accessibilityIdentifier("onboarding.home")
            HStack(spacing: 6) {
                Text(model.sessionDisplayName)
                    .font(.title3)
                if !model.sessionRef.isEmpty {
                    Text(L10n.Onboarding.contributorRef(ref: model.sessionRef))
                        .font(.title3)
                        .foregroundStyle(.secondary)
                }
            }
            .accessibilityElement(children: .combine)
            .accessibilityLabel(L10n.Onboarding.contributorOption(
                displayName: model.sessionDisplayName,
                ref: model.sessionRef
            ))
            .accessibilityIdentifier("onboarding.home.displayName")
            OnboardingProjectMetaLines(
                label: model.projectLabel,
                showFolder: true,
                folderName: model.projectBasename,
                createdAt: model.projectCreatedAt,
                updatedAt: model.projectUpdatedAt,
                updatedByDisplayName: model.projectUpdatedByDisplayName,
                updatedByRef: model.projectUpdatedByRef,
                accessibilityPrefix: "onboarding.home"
            )
            Button(String(localized: L10n.Onboarding.signOut)) {
                Task { await model.signOut() }
            }
            .accessibilityIdentifier("onboarding.signOut")
        }
        .padding(32)
    }
}
