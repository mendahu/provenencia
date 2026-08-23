import SwiftUI

struct OnboardingView: View {
    @State private var model: OnboardingModel

    init(store: any GenealogyStore = GoStore()) {
        _model = State(initialValue: OnboardingModel(store: store))
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
            Text("Welcome to Provenance")
                .font(.title2)
            Text("Your researcher name is how work is attributed. The family name labels this project folder.")
                .foregroundStyle(.secondary)
            TextField("Researcher name", text: $model.displayName, prompt: Text("Jane Smith"))
                .textFieldStyle(.roundedBorder)
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
            Text("You're in")
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
        }
        .padding(32)
    }
}

#Preview("Form") {
    OnboardingView(store: FakeStore())
}

#Preview("Returning") {
    OnboardingView(store: FakeStore(identity: InstallIdentity(userID: "1", displayName: "Jake Robins")))
}
