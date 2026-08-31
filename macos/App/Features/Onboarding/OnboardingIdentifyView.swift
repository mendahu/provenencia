import SwiftUI

struct OnboardingIdentifyView: View {
    @Bindable var model: OnboardingModel

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            VStack(alignment: .leading, spacing: 16) {
                if model.mode == .create {
                    Text(L10n.Onboarding.nameResearchTitle)
                        .font(.title2)
                    Text(L10n.Onboarding.nameResearchBody)
                        .foregroundStyle(.secondary)
                    VStack(alignment: .leading, spacing: 6) {
                        Text(L10n.Onboarding.researcherName)
                            .font(.headline)
                        TextField(
                            "",
                            text: $model.displayName,
                            prompt: Text(L10n.Onboarding.researcherNamePrompt)
                        )
                            .textFieldStyle(.roundedBorder)
                            .disabled(model.researcherLocked)
                            .accessibilityIdentifier("onboarding.displayName")
                    }
                    VStack(alignment: .leading, spacing: 6) {
                        Text(L10n.Onboarding.familyName)
                            .font(.headline)
                        TextField(
                            "",
                            text: $model.familyName,
                            prompt: Text(L10n.Onboarding.familyNamePrompt)
                        )
                            .textFieldStyle(.roundedBorder)
                            .accessibilityIdentifier("onboarding.familyName")
                        if !model.folderNamePreview.isEmpty {
                            Text(L10n.Onboarding.folderNamePreview(folderName: model.folderNamePreview))
                                .font(.caption)
                                .foregroundStyle(.secondary)
                                .accessibilityIdentifier("onboarding.folderPreview")
                        }
                    }
                } else {
                    Text(L10n.Onboarding.whoAreYouTitle)
                        .font(.title2)
                    Text(
                        L10n.Onboarding.contributorsBody(
                            projectName: model.selectedProject?.lastPathComponent
                                ?? String(localized: L10n.Onboarding.thisProject)
                        )
                    )
                        .foregroundStyle(.secondary)
                    Picker(String(localized: L10n.Onboarding.contributorPicker), selection: $model.selectedContributorID) {
                        ForEach(model.catalogUsers, id: \.userID) { user in
                            HStack(spacing: 6) {
                                Text(user.displayName)
                                if !user.ref.isEmpty {
                                    Text(L10n.Onboarding.contributorRef(ref: user.ref))
                                        .foregroundStyle(.secondary)
                                }
                            }
                            .tag(user.userID)
                            .accessibilityLabel(L10n.Onboarding.contributorOption(
                                displayName: user.displayName,
                                ref: user.ref
                            ))
                        }
                        Text(L10n.Onboarding.notListed).tag(OnboardingModel.newContributorID)
                    }
                    .pickerStyle(.radioGroup)
                    .accessibilityIdentifier("onboarding.contributor")
                    if model.selectedContributorID == OnboardingModel.newContributorID {
                        VStack(alignment: .leading, spacing: 6) {
                            Text(L10n.Onboarding.researcherName)
                                .font(.headline)
                            TextField(
                                "",
                                text: $model.displayName,
                                prompt: Text(L10n.Onboarding.researcherNamePrompt)
                            )
                                .textFieldStyle(.roundedBorder)
                                .disabled(model.researcherLocked)
                                .accessibilityIdentifier("onboarding.displayName")
                        }
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
                    Button(String(localized: L10n.Onboarding.back)) {
                        model.goBack()
                    }
                    .accessibilityIdentifier("onboarding.back")
                },
                trailing: {
                    Button(String(localized: L10n.Onboarding.continueAction)) {
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
