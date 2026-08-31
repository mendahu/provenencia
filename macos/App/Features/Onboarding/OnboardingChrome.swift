import AppKit
import SwiftUI

struct OnboardingFileChoice: View {
    let title: LocalizedStringResource
    let subtitle: LocalizedStringResource
    let selected: Bool
    let action: () -> Void

    init(_ title: LocalizedStringResource, subtitle: LocalizedStringResource, selected: Bool, action: @escaping () -> Void) {
        self.title = title
        self.subtitle = subtitle
        self.selected = selected
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            VStack(alignment: .leading, spacing: 6) {
                Text(title)
                    .font(.headline)
                    .foregroundStyle(.primary)
                Text(subtitle)
                    .font(.callout)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.leading)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(16)
            .background(
                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .fill(Color(nsColor: .controlBackgroundColor))
            )
            .overlay(
                RoundedRectangle(cornerRadius: 10, style: .continuous)
                    .strokeBorder(selected ? Color.accentColor : Color(nsColor: .separatorColor), lineWidth: selected ? 2 : 1)
            )
        }
        .buttonStyle(.plain)
    }
}

struct OnboardingOpenPicker: View {
    @Bindable var model: OnboardingModel

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(L10n.Onboarding.projectFolder)
                .font(.headline)
            if model.availableProjects.isEmpty {
                HStack(alignment: .center, spacing: 12) {
                    Text(L10n.Onboarding.noProjectsInDocuments)
                        .foregroundStyle(.secondary)
                    Spacer(minLength: 0)
                    Button(String(localized: L10n.Onboarding.chooseFolder)) {
                        onboardingChooseFolder(model)
                    }
                    .accessibilityIdentifier("onboarding.chooseFolder")
                }
            } else {
                HStack(alignment: .center, spacing: 12) {
                    Picker(String(localized: L10n.Onboarding.existingProjectPicker), selection: $model.selectedProject) {
                        Text(L10n.Onboarding.selectProject).tag(Optional<URL>.none)
                        ForEach(model.availableProjects, id: \.path) { url in
                            Text(url.lastPathComponent).tag(Optional(url))
                        }
                    }
                    .labelsHidden()
                    .accessibilityIdentifier("onboarding.existingProject")
                    Button(String(localized: L10n.Onboarding.chooseFolder)) {
                        onboardingChooseFolder(model)
                    }
                    .accessibilityIdentifier("onboarding.chooseFolder")
                }
            }
            if model.selectedProject != nil {
                OnboardingProjectMetaLines(
                    label: model.projectLabel,
                    showFolder: false,
                    folderName: model.projectBasename,
                    createdAt: model.projectCreatedAt,
                    updatedAt: model.projectUpdatedAt,
                    updatedByDisplayName: model.projectUpdatedByDisplayName,
                    updatedByRef: model.projectUpdatedByRef,
                    accessibilityPrefix: "onboarding.open"
                )
            }
        }
        .onChange(of: model.selectedProject) { _, _ in
            Task { await model.refreshSelectedProjectInfo() }
        }
    }
}

/// Shared project bookkeeping lines for open-picker preview and home.
struct OnboardingProjectMetaLines: View {
    let label: String
    var showFolder: Bool = true
    let folderName: String
    let createdAt: String
    let updatedAt: String
    let updatedByDisplayName: String
    let updatedByRef: String
    var accessibilityPrefix: String = "onboarding.project"

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            if !label.isEmpty {
                Text(label)
                    .accessibilityIdentifier("\(accessibilityPrefix).projectLabel")
            }
            if showFolder, !folderName.isEmpty {
                Text(L10n.Onboarding.homeFolder(folderName: folderName))
                    .foregroundStyle(.secondary)
                    .accessibilityIdentifier("\(accessibilityPrefix).project")
            }
            if !createdAt.isEmpty {
                Text(L10n.Onboarding.homeCreated(date: Self.formatDate(createdAt)))
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .accessibilityIdentifier("\(accessibilityPrefix).created")
            }
            if !updatedAt.isEmpty {
                Text(L10n.Onboarding.homeUpdated(date: Self.formatDate(updatedAt)))
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .accessibilityIdentifier("\(accessibilityPrefix).updated")
            }
            if !updatedByDisplayName.isEmpty {
                Text(L10n.Onboarding.homeUpdatedBy(
                    displayName: updatedByDisplayName,
                    ref: updatedByRef
                ))
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .accessibilityIdentifier("\(accessibilityPrefix).updatedBy")
            }
        }
        .padding(.top, 4)
    }

    private static func formatDate(_ rfc3339: String) -> String {
        let parser = ISO8601DateFormatter()
        parser.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = parser.date(from: rfc3339) ?? ISO8601DateFormatter().date(from: rfc3339) {
            return date.formatted(date: .abbreviated, time: .shortened)
        }
        return rfc3339
    }
}

struct OnboardingFooter<Leading: View, Trailing: View>: View {
    @Bindable var model: OnboardingModel
    @ViewBuilder var leading: Leading
    @ViewBuilder var trailing: Trailing

    var body: some View {
        HStack(spacing: 12) {
            leading
            Spacer()
            if model.isBusy {
                ProgressView()
                    .controlSize(.small)
            }
            trailing
                .buttonStyle(.borderedProminent)
        }
    }
}

@MainActor
func onboardingChooseFolder(_ model: OnboardingModel) {
    let panel = NSOpenPanel()
    panel.canChooseFiles = false
    panel.canChooseDirectories = true
    panel.allowsMultipleSelection = false
    panel.treatsFilePackagesAsDirectories = true
    panel.prompt = String(localized: L10n.Onboarding.openPanelPrompt)
    panel.message = String(localized: L10n.Onboarding.openPanelMessage)
    guard panel.runModal() == .OK, let url = panel.url else {
        return
    }
    Task { await model.choose(url) }
}
