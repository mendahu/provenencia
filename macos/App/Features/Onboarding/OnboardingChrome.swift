import AppKit
import SwiftUI

struct OnboardingFileChoice: View {
    let title: String
    let subtitle: String
    let selected: Bool
    let action: () -> Void

    init(_ title: String, subtitle: String, selected: Bool, action: @escaping () -> Void) {
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
            Text("Project folder")
                .font(.headline)
            if model.availableProjects.isEmpty {
                HStack(alignment: .center, spacing: 12) {
                    Text("No project folders in Documents.")
                        .foregroundStyle(.secondary)
                    Spacer(minLength: 0)
                    Button("Choose…") {
                        onboardingChooseFolder(model)
                    }
                    .accessibilityIdentifier("onboarding.chooseFolder")
                }
            } else {
                HStack(alignment: .center, spacing: 12) {
                    Picker("Existing project", selection: $model.selectedProject) {
                        Text("Select a project").tag(Optional<URL>.none)
                        ForEach(model.availableProjects, id: \.path) { url in
                            Text(url.lastPathComponent).tag(Optional(url))
                        }
                    }
                    .labelsHidden()
                    .accessibilityIdentifier("onboarding.existingProject")
                    Button("Choose…") {
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
    panel.prompt = "Open"
    panel.message = "Choose a Provenencia project folder."
    guard panel.runModal() == .OK, let url = panel.url else {
        return
    }
    model.choose(url)
}
