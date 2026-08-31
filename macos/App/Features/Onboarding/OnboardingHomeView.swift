import SwiftUI

struct OnboardingHomeView: View {
    @Bindable var model: OnboardingModel

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text(L10n.Onboarding.signedInTitle)
                .font(.title2)
                .accessibilityIdentifier("onboarding.home")
            Text(researcherLine)
                .font(.title3)
                .accessibilityIdentifier("onboarding.home.displayName")
            if !model.projectLabel.isEmpty {
                Text(model.projectLabel)
                    .accessibilityIdentifier("onboarding.home.projectLabel")
            }
            if !model.projectBasename.isEmpty {
                Text(L10n.Onboarding.homeFolder(folderName: model.projectBasename))
                    .foregroundStyle(.secondary)
                    .accessibilityIdentifier("onboarding.home.project")
            }
            if !model.projectCreatedAt.isEmpty {
                Text(L10n.Onboarding.homeCreated(date: formatDate(model.projectCreatedAt)))
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .accessibilityIdentifier("onboarding.home.created")
            }
            if !model.projectUpdatedAt.isEmpty {
                Text(L10n.Onboarding.homeUpdated(date: formatDate(model.projectUpdatedAt)))
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .accessibilityIdentifier("onboarding.home.updated")
            }
            if !model.projectUpdatedByDisplayName.isEmpty {
                Text(L10n.Onboarding.homeUpdatedBy(
                    displayName: model.projectUpdatedByDisplayName,
                    ref: model.projectUpdatedByRef
                ))
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .accessibilityIdentifier("onboarding.home.updatedBy")
            }
            Button(String(localized: L10n.Onboarding.signOut)) {
                Task { await model.signOut() }
            }
            .accessibilityIdentifier("onboarding.signOut")
        }
        .padding(32)
    }

    private var researcherLine: String {
        if model.sessionRef.isEmpty {
            return model.sessionDisplayName
        }
        return L10n.Onboarding.contributorOption(
            displayName: model.sessionDisplayName,
            ref: model.sessionRef
        )
    }

    private func formatDate(_ rfc3339: String) -> String {
        let parser = ISO8601DateFormatter()
        parser.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = parser.date(from: rfc3339) ?? ISO8601DateFormatter().date(from: rfc3339) {
            return date.formatted(date: .abbreviated, time: .shortened)
        }
        return rfc3339
    }
}
