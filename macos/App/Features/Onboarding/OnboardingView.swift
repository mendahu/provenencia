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
            case .form:
                form
            case .home:
                home
            }
        }
        .frame(minWidth: 640, minHeight: 420)
        .background(Color(nsColor: .windowBackgroundColor))
        .task {
            await model.load()
        }
    }

    private var form: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text(model.researcherLocked ? "Welcome Back!" : "Welcome to Provenance")
                .font(.title2)
            Text(model.researcherLocked
                ? "You're signed in as \(model.displayName). Create a project, open an existing one, or sign out. Existing folders stay on disk."
                : "Your researcher name is how work is attributed. Create a new project folder or open an existing one.")
                .foregroundStyle(.secondary)
            TextField("Researcher name", text: $model.displayName, prompt: Text("Jane Smith"))
                .textFieldStyle(.roundedBorder)
                .disabled(model.researcherLocked)
                .accessibilityIdentifier("onboarding.displayName")
            Picker("Project", selection: $model.mode) {
                Text("Create new").tag(OnboardingModel.Mode.create)
                Text("Open existing").tag(OnboardingModel.Mode.open)
            }
            .pickerStyle(.segmented)
            .accessibilityIdentifier("onboarding.mode")
            if model.mode == .create {
                TextField("Family / project name", text: $model.familyName, prompt: Text("Smith Family"))
                    .textFieldStyle(.roundedBorder)
                    .accessibilityIdentifier("onboarding.familyName")
            } else {
                openPicker
            }
            if let errorText = model.errorText {
                Text(errorText)
                    .foregroundStyle(.red)
                    .accessibilityIdentifier("onboarding.error")
            }
            HStack {
                if model.isBusy {
                    ProgressView()
                }
                if model.researcherLocked {
                    Button("Sign out") {
                        Task { await model.signOut() }
                    }
                    .accessibilityIdentifier("onboarding.signOut")
                }
                Button("Continue") {
                    Task { await model.submit() }
                }
                .disabled(!model.canContinue)
                .accessibilityIdentifier("onboarding.continue")
            }
        }
        .padding(32)
        .frame(maxWidth: 480)
    }

    private var openPicker: some View {
        VStack(alignment: .leading, spacing: 8) {
            if model.availableProjects.isEmpty {
                Text("No project folders in Documents.")
                    .foregroundStyle(.secondary)
            } else {
                Picker("Existing project", selection: $model.selectedProject) {
                    Text("Choose a project").tag(Optional<URL>.none)
                    ForEach(model.availableProjects, id: \.path) { url in
                        Text(url.lastPathComponent).tag(Optional(url))
                    }
                }
                .accessibilityIdentifier("onboarding.existingProject")
            }
            Button("Choose…") {
                chooseFolder()
            }
            .accessibilityIdentifier("onboarding.chooseFolder")
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
        Task { await model.open(url) }
    }
}

#Preview("Form") {
    OnboardingView(store: FakeStore(), folders: .previewEmpty)
}

#Preview("Identity only") {
    OnboardingView(
        store: FakeStore(identity: InstallIdentity(userID: "1", displayName: "Jane Smith")),
        folders: .previewEmpty
    )
}

#Preview("Picker with folders") {
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
