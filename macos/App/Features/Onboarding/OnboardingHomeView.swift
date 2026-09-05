import SwiftUI

struct OnboardingHomeView: View {
    @Bindable var model: OnboardingModel

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            Text(L10n.Onboarding.signedInTitle)
                .font(PVFont.display(size: PVTypeScale.h2))
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityIdentifier("onboarding.home")
            HStack(spacing: PVSpacing.space3) {
                Text(model.sessionDisplayName)
                    .font(PVFont.body(size: PVTypeScale.h3))
                    .foregroundStyle(PVColor.textPrimary)
                if !model.sessionRef.isEmpty {
                    Text(L10n.Onboarding.contributorRef(ref: model.sessionRef))
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
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
            PVButton(L10n.Onboarding.signOut, variant: .ghost) {
                Task { await model.signOut() }
            }
            .accessibilityIdentifier("onboarding.signOut")
        }
        .padding(32)
    }
}
