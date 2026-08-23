import SwiftUI

struct OnboardingIdentifyView: View {
    @Bindable var model: OnboardingModel

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            VStack(alignment: .leading, spacing: 16) {
                if model.mode == .create {
                    Text("Name this research")
                        .font(.title2)
                    Text("Your researcher name is how work is attributed. The family name labels this project folder.")
                        .foregroundStyle(.secondary)
                    TextField("Researcher name", text: $model.displayName, prompt: Text("Jane Smith"))
                        .textFieldStyle(.roundedBorder)
                        .disabled(model.researcherLocked)
                        .accessibilityIdentifier("onboarding.displayName")
                    TextField("Family / project name", text: $model.familyName, prompt: Text("Smith Family"))
                        .textFieldStyle(.roundedBorder)
                        .accessibilityIdentifier("onboarding.familyName")
                } else {
                    Text("Who are you in this file?")
                        .font(.title2)
                    Text("These are the contributors already in \(model.selectedProject?.lastPathComponent ?? "this project"). Choose yourself to keep the same ID, or add a new contributor.")
                        .foregroundStyle(.secondary)
                    Picker("Contributor", selection: $model.selectedContributorID) {
                        ForEach(model.catalogUsers, id: \.userID) { user in
                            Text(user.displayName).tag(user.userID)
                        }
                        Text("I’m not listed — add me").tag(OnboardingModel.newContributorID)
                    }
                    .pickerStyle(.radioGroup)
                    .accessibilityIdentifier("onboarding.contributor")
                    if model.selectedContributorID == OnboardingModel.newContributorID {
                        TextField("Researcher name", text: $model.displayName, prompt: Text("Jane Smith"))
                            .textFieldStyle(.roundedBorder)
                            .disabled(model.researcherLocked)
                            .accessibilityIdentifier("onboarding.displayName")
                    }
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
                    Button("Back") {
                        model.goBack()
                    }
                    .accessibilityIdentifier("onboarding.back")
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
