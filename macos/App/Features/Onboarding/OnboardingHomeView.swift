import SwiftUI

struct OnboardingHomeView: View {
    var model: OnboardingModel

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            PVLogoMark(size: 36)
                .opacity(0.85)
            Text(L10n.Onboarding.signedInTitle)
                .font(PVFont.display(size: PVTypeScale.h2))
                .foregroundStyle(PVColor.textDisplay)
                .accessibilityIdentifier("onboarding.home")
            HStack(spacing: PVSpacing.space3) {
                Text(model.session?.displayName ?? "")
                    .font(PVFont.body(size: PVTypeScale.h3))
                    .foregroundStyle(PVColor.textPrimary)
                if let ref = model.session?.ref, !ref.isEmpty {
                    Text(L10n.Onboarding.contributorRef(ref: ref))
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            .accessibilityElement(children: .combine)
            .accessibilityLabel(L10n.Onboarding.contributorOption(
                displayName: model.session?.displayName ?? "",
                ref: model.session?.ref ?? ""
            ))
            .accessibilityIdentifier("onboarding.home.displayName")
            if let project = model.project {
                OnboardingProjectMetaLines(
                    project,
                    showFolder: true,
                    accessibilityPrefix: "onboarding.home"
                )
            }
            PVButton(L10n.Onboarding.signOut, variant: .ghost) {
                Task { await model.signOut() }
            }
            .accessibilityIdentifier("onboarding.signOut")
        }
        .padding(PVSpacing.space9)
        .frame(maxWidth: PVSpacing.measureForm, alignment: .leading)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .center)
    }
}
