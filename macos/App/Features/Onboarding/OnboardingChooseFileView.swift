import SwiftUI

struct OnboardingChooseFileView: View {
    @Bindable var model: OnboardingModel

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                Text(L10n.Onboarding.welcomeTitle)
                    .font(PVFont.display(size: PVTypeScale.h2))
                    .foregroundStyle(PVColor.textDisplay)
                Text(model.researcherLocked
                    ? L10n.Onboarding.bodySignedIn(displayName: model.displayName)
                    : String(localized: L10n.Onboarding.bodyChoose))
                    .font(PVFont.body(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textSecondary)
                    .fixedSize(horizontal: false, vertical: true)
                HStack(spacing: PVSpacing.space6) {
                    OnboardingFileChoice(
                        L10n.Onboarding.createNewTitle,
                        subtitle: L10n.Onboarding.createNewSubtitle,
                        icon: "folder-plus",
                        selected: model.mode == .create
                    ) {
                        Task { await model.selectMode(.create) }
                    }
                    OnboardingFileChoice(
                        L10n.Onboarding.haveFileTitle,
                        subtitle: L10n.Onboarding.haveFileSubtitle,
                        icon: "folder-open",
                        selected: model.mode == .open
                    ) {
                        Task { await model.selectMode(.open) }
                    }
                }
                .accessibilityElement(children: .contain)
                .accessibilityIdentifier("onboarding.mode")
                if model.mode == .create {
                    HStack(alignment: .top, spacing: PVSpacing.space4) {
                        PVIcon("info", size: 15)
                            .foregroundStyle(PVColor.textFaint)
                            .padding(.top, 1)
                        Text(L10n.Onboarding.createNewFolderNote)
                            .font(PVFont.body(size: PVTypeScale.bodySmall))
                            .foregroundStyle(PVColor.textSecondary)
                    }
                    .padding(PVSpacing.space6)
                    .background(
                        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                            .fill(PVColor.surfaceSunken)
                    )
                    .overlay(
                        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                            .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
                    )
                }
                if model.mode == .open {
                    OnboardingOpenPicker(model: model)
                }
            }
            Spacer(minLength: 24)
            OnboardingFooter(
                model: model,
                leading: {
                    if model.researcherLocked {
                        PVButton(L10n.Onboarding.signOut, variant: .ghost) {
                            Task { await model.signOut() }
                        }
                        .accessibilityIdentifier("onboarding.signOut")
                    }
                },
                trailing: {
                    PVButton(L10n.Onboarding.continueAction, variant: .primary) {
                        Task { await model.submit() }
                    }
                    .keyboardShortcut(.defaultAction)
                    .disabled(!model.canContinue)
                    .accessibilityIdentifier("onboarding.continue")
                }
            )
        }
        .padding(32)
        .frame(maxWidth: 560, alignment: .leading)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .center)
    }
}
