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
                VStack(spacing: PVSpacing.space6) {
                    Text(verbatim: "Provenencia")
                        .font(PVFont.display(size: PVTypeScale.h1))
                        .foregroundStyle(PVColor.textDisplay)
                    ProgressView()
                        .tint(PVColor.accent)
                    Text(L10n.Onboarding.loadingFooterNote)
                        .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                        .foregroundStyle(PVColor.textMuted)
                }
                .accessibilityIdentifier("onboarding.loading")
            case .chooseFile:
                OnboardingChooseFileView(model: model)
            case .identify:
                OnboardingIdentifyView(model: model)
            case .home:
                OnboardingHomeView(model: model)
            }
        }
        .frame(minWidth: 700, minHeight: 460)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
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
        store: FakeStore(identity: InstallIdentity(userID: "1", displayName: "Jane Smith", ref: "USR-A1B2C")),
        folders: .previewEmpty
    )
}

#Preview("Open picker") {
    OnboardingView(
        store: FakeStore(identity: InstallIdentity(userID: "1", displayName: "Jane Smith", ref: "USR-A1B2C")),
        folders: .previewWithProject
    )
}

#Preview("Returning") {
    let folders = OnboardingFolders.previewWithProject
    let project = folders.documentsDirectory.appendingPathComponent("Smith Family.provenencia")
    return OnboardingView(
        store: FakeStore(
            identity: InstallIdentity(userID: "1", displayName: "Jane Smith", ref: "USR-A1B2C"),
            activeProjectDir: project.path
        ),
        folders: folders
    )
}
