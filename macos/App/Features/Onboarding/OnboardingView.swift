import AppKit
import SwiftUI

struct OnboardingView: View {
    @State private var model: OnboardingModel

    init(store: any GenealogyStore = GoStore(), folders: OnboardingFolders? = nil) {
        _model = State(initialValue: OnboardingModel(store: store, folders: folders))
    }

    var body: some View {
        Group {
            switch model.phase {
            case .loading:
                ProgressView()
                    .accessibilityIdentifier("onboarding.loading")
            case .chooseFile:
                chooseFile
            case .identify:
                identify
            case .home:
                home
            }
        }
        .frame(minWidth: 640, minHeight: 420)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color(nsColor: .windowBackgroundColor))
        .task {
            await model.load()
        }
    }

    private var chooseFile: some View {
        VStack(alignment: .leading, spacing: 0) {
            VStack(alignment: .leading, spacing: 20) {
                Text("Welcome to Provenance")
                    .font(.title2)
                Text(model.researcherLocked
                    ? "You're signed in as \(model.displayName). Open a project you already have, or create a new one."
                    : "Do you already have a Provenance project, or do you want to start a new one?")
                    .foregroundStyle(.secondary)
                    .fixedSize(horizontal: false, vertical: true)
                HStack(spacing: 16) {
                    fileChoice("Create new", subtitle: "Start a new project folder", selected: model.mode == .create) {
                        Task { await model.selectMode(.create) }
                    }
                    fileChoice("I have a file", subtitle: "Open a project you already have", selected: model.mode == .open) {
                        Task { await model.selectMode(.open) }
                    }
                }
                .accessibilityElement(children: .contain)
                .accessibilityIdentifier("onboarding.mode")
                if model.mode == .open {
                    openPicker
                }
                if let errorText = model.errorText {
                    Text(errorText)
                        .foregroundStyle(.red)
                        .accessibilityIdentifier("onboarding.error")
                }
            }
            Spacer(minLength: 24)
            footer(
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

    private func fileChoice(_ title: String, subtitle: String, selected: Bool, action: @escaping () -> Void) -> some View {
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

    private var identify: some View {
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
            footer(
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

    private var openPicker: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Project folder")
                .font(.headline)
            if model.availableProjects.isEmpty {
                HStack(alignment: .center, spacing: 12) {
                    Text("No project folders in Documents.")
                        .foregroundStyle(.secondary)
                    Spacer(minLength: 0)
                    Button("Choose…") {
                        chooseFolder()
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
                        chooseFolder()
                    }
                    .accessibilityIdentifier("onboarding.chooseFolder")
                }
            }
        }
    }

    private func footer<Leading: View, Trailing: View>(
        @ViewBuilder leading: () -> Leading,
        @ViewBuilder trailing: () -> Trailing
    ) -> some View {
        HStack(spacing: 12) {
            leading()
            Spacer()
            if model.isBusy {
                ProgressView()
                    .controlSize(.small)
            }
            trailing()
                .buttonStyle(.borderedProminent)
        }
    }

    private var home: some View {
        VStack(spacing: 12) {
            Text("You're signed in")
                .font(.title2)
                .accessibilityIdentifier("onboarding.home")
            Text(model.sessionDisplayName)
                .font(.title3)
                .accessibilityIdentifier("onboarding.home.displayName")
            if !model.projectBasename.isEmpty {
                Text(model.projectBasename)
                    .foregroundStyle(.secondary)
                    .accessibilityIdentifier("onboarding.home.project")
            }
            Button("Sign out") {
                Task { await model.signOut() }
            }
            .accessibilityIdentifier("onboarding.signOut")
        }
        .padding(32)
    }

    private func chooseFolder() {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        panel.treatsFilePackagesAsDirectories = true
        panel.prompt = "Open"
        panel.message = "Choose a Provenance project folder."
        guard panel.runModal() == .OK, let url = panel.url else {
            return
        }
        model.choose(url)
    }
}

#Preview("Choose file") {
    OnboardingView(store: FakeStore(), folders: .previewEmpty)
}

#Preview("Identity, no active") {
    OnboardingView(
        store: FakeStore(identity: InstallIdentity(userID: "1", displayName: "Jane Smith")),
        folders: .previewEmpty
    )
}

#Preview("Open picker") {
    OnboardingView(
        store: FakeStore(identity: InstallIdentity(userID: "1", displayName: "Jane Smith")),
        folders: .previewWithProject
    )
}

#Preview("Returning") {
    let folders = OnboardingFolders.previewWithProject
    let project = folders.documentsDirectory.appendingPathComponent("Smith Family.provenance")
    return OnboardingView(
        store: FakeStore(
            identity: InstallIdentity(userID: "1", displayName: "Jane Smith"),
            activeProjectDir: project.path
        ),
        folders: folders
    )
}
