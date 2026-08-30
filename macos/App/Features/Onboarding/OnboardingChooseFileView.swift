import SwiftUI

struct OnboardingChooseFileView: View {
    @Bindable var model: OnboardingModel

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            VStack(alignment: .leading, spacing: 20) {
                Text(L10n.Onboarding.welcomeTitle)
                    .font(.title2)
                Text(model.researcherLocked
                    ? L10n.Onboarding.bodySignedIn(displayName: model.displayName)
                    : String(localized: L10n.Onboarding.bodyChoose))
                    .foregroundStyle(.secondary)
                    .fixedSize(horizontal: false, vertical: true)
                HStack(spacing: 16) {
                    OnboardingFileChoice(
                        L10n.Onboarding.createNewTitle,
                        subtitle: L10n.Onboarding.createNewSubtitle,
                        selected: model.mode == .create
                    ) {
                        Task { await model.selectMode(.create) }
                    }
                    OnboardingFileChoice(
                        L10n.Onboarding.haveFileTitle,
                        subtitle: L10n.Onboarding.haveFileSubtitle,
                        selected: model.mode == .open
                    ) {
                        Task { await model.selectMode(.open) }
                    }
                }
                .accessibilityElement(children: .contain)
                .accessibilityIdentifier("onboarding.mode")
                if model.mode == .open {
                    OnboardingOpenPicker(model: model)
                }
                if let errorText = model.errorText {
                    Text(errorText)
                        .foregroundStyle(.red)
                        .accessibilityIdentifier("onboarding.error")
                }
            }
            Spacer(minLength: 24)
            OnboardingFooter(
                model: model,
                leading: {
                    if model.researcherLocked {
                        Button(L10n.Onboarding.signOut) {
                            Task { await model.signOut() }
                        }
                        .accessibilityIdentifier("onboarding.signOut")
                    }
                },
                trailing: {
                    Button(L10n.Onboarding.continueAction) {
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
