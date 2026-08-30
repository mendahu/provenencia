import SwiftUI

struct OnboardingChooseFileView: View {
    @Bindable var model: OnboardingModel

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            VStack(alignment: .leading, spacing: 20) {
                Text("Welcome to Provenencia")
                    .font(.title2)
                Text(model.researcherLocked
                    ? "You're signed in as \(model.displayName). Open a project you already have, or create a new one."
                    : "Do you already have a Provenencia project, or do you want to start a new one?")
                    .foregroundStyle(.secondary)
                    .fixedSize(horizontal: false, vertical: true)
                HStack(spacing: 16) {
                    OnboardingFileChoice("Create new", subtitle: "Start a new project folder", selected: model.mode == .create) {
                        Task { await model.selectMode(.create) }
                    }
                    OnboardingFileChoice("I have a file", subtitle: "Open a project you already have", selected: model.mode == .open) {
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
                        Button("Sign out") {
                            Task { await model.signOut() }
                        }
                        .accessibilityIdentifier("onboarding.signOut")
                    }
                },
                trailing: {
                    Button("Continue") {
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
