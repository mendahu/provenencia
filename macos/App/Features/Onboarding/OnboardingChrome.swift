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
        }
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
    model.choose(url)
}
