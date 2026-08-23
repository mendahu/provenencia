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
            Text(model.researcherLocked ? "You're signed in" : "Welcome to Provenance")
                .font(.title2)
            Text(model.researcherLocked
                ? "You're signed in as \(model.displayName). Name this project folder, or sign out to use a different researcher on this Mac. Existing project folders stay in Documents."
                : "Your researcher name is how work is attributed. The family name labels this project folder.")
                .foregroundStyle(.secondary)
            TextField("Researcher name", text: $model.displayName, prompt: Text("Jane Smith"))
                .textFieldStyle(.roundedBorder)
                .disabled(model.researcherLocked)
                .accessibilityIdentifier("onboarding.displayName")
            TextField("Family / project name", text: $model.familyName, prompt: Text("Smith Family"))
                .textFieldStyle(.roundedBorder)
                .accessibilityIdentifier("onboarding.familyName")
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

#Preview("Returning") {
    OnboardingView(
        store: FakeStore(identity: InstallIdentity(userID: "1", displayName: "Jane Smith")),
        folders: .previewWithProject
    )
}
