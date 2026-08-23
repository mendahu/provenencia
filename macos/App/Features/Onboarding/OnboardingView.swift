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
                OnboardingChooseFileView(model: model)
            case .identify:
                OnboardingIdentifyView(model: model)
            case .home:
                OnboardingHomeView(model: model)
            }
        }
        .frame(minWidth: 640, minHeight: 420)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(Color(nsColor: .windowBackgroundColor))
        .task {
            await model.load()
        }
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
