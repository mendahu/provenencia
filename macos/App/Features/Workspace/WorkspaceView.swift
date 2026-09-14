import SwiftUI

/// The post-onboarding application shell: one leading sidebar plus one
/// content host (W-1). Replaces `OnboardingHomeView` as the permanent
/// chrome once a project is open (`OnboardingModel.phase == .home`) — see
/// `docs/deployment-plan/archive/spike-2/design/archive/S2-01-workspace-chrome.md`.
///
/// `projectDir` and `userID` are non-optional: `OnboardingView` only
/// mounts this from `.home(projectDir:userID:)`, which is set solely via
/// `OnboardingModel.enterHome`.
///
/// Catalog access is a held Go session (`catalogsession`): destination
/// content mounts immediately and may overlap `refreshAll` / feature loads.
/// Leave closes the session via `closeCatalogSession`.
struct WorkspaceView: View {
    var model: OnboardingModel
    let projectDir: String
    let userID: String
    @State private var workspace: WorkspaceModel
    @State private var navigation: WorkspaceNavigation
    @State private var catalogCounts: CatalogCounts
    @State private var countsToast: VocabularyToast?
    @Environment(SignOutCoordinator.self) private var signOutCoordinator
    @Environment(NavigationCoordinator.self) private var navigationCoordinator

    init(model: OnboardingModel, projectDir: String, userID: String) {
        self.model = model
        self.projectDir = projectDir
        self.userID = userID
        _workspace = State(initialValue: WorkspaceModel())
        _navigation = State(initialValue: WorkspaceNavigation())
        _catalogCounts = State(initialValue: CatalogCounts(projectDir: projectDir, store: model.store))
    }

    var body: some View {
        HStack(spacing: 0) {
            WorkspaceSidebar(session: model.session, workspace: workspace, catalogCounts: catalogCounts)
            WorkspaceContent(
                projectDir: projectDir,
                userID: userID,
                sessionDisplayName: model.session?.displayName ?? "",
                store: model.store,
                catalogCounts: catalogCounts
            )
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .environment(catalogCounts)
        .environment(navigation)
        .vocabularyToastOverlay($countsToast, identifier: "workspace.counts.toast")
        .task {
            await catalogCounts.refreshAll()
            if let message = catalogCounts.lastRefreshError {
                countsToast = VocabularyToast(
                    title: String(localized: L10n.Workspace.countsRefreshFailedTitle),
                    body: message,
                    tone: .danger
                )
            }
        }
        .onAppear {
            if let uuid = model.project?.uuid, !uuid.isEmpty {
                navigation.attachProject(uuid: uuid)
            }
            signOutCoordinator.isAvailable = true
            signOutCoordinator.action = { [model] in
                Task { await model.signOut() }
            }
            bindNavigationCommands()
        }
        .onChange(of: navigation.canGoBack) { _, _ in bindNavigationCommands() }
        .onChange(of: navigation.canGoForward) { _, _ in bindNavigationCommands() }
        .onChange(of: navigation.currentLocation) { _, _ in bindNavigationCommands() }
        .onChange(of: model.project?.uuid) { _, uuid in
            if let uuid, !uuid.isEmpty {
                navigation.attachProject(uuid: uuid)
                bindNavigationCommands()
            }
        }
        .onDisappear {
            signOutCoordinator.isAvailable = false
            navigationCoordinator.canGoBack = false
            navigationCoordinator.canGoForward = false
            navigationCoordinator.goBack = {}
            navigationCoordinator.goForward = {}
            let store = model.store
            let dir = projectDir
            Task {
                do {
                    try await store.closeCatalogSession(projectDir: dir)
                } catch {
                    #if DEBUG
                    assertionFailure("closeCatalogSession failed: \(error)")
                    #endif
                }
            }
        }
        .accessibilityIdentifier("workspace")
    }

    private func bindNavigationCommands() {
        navigationCoordinator.canGoBack = navigation.canGoBack
        navigationCoordinator.canGoForward = navigation.canGoForward
        navigationCoordinator.goBack = { navigation.goBack() }
        navigationCoordinator.goForward = { navigation.goForward() }
    }
}

#if DEBUG
#Preview("Expanded, Sources") {
    let model = WorkspaceView.previewModel()
    let projectDir = model.activeProjectDir ?? ""
    WorkspaceView(
        model: model,
        projectDir: projectDir,
        userID: PreviewFixture.identity.userID
    )
    .environment(SignOutCoordinator())
    .environment(NavigationCoordinator())
    .environment(OmnibarFocusCoordinator())
    .frame(width: PVSpacing.widthWorkspaceDefault, height: PVSpacing.heightWorkspaceDefault)
}

extension WorkspaceView {
    static func previewModel() -> OnboardingModel {
        let store = FakeStore(identity: PreviewFixture.identity)
        let projectDir = store.lastResult.projectDir
        store.sourcesByProject[projectDir] = (1 ... 12).map {
            CatalogSource(id: "\($0)", ref: "SRC-FAKE\($0)", sourceTypeID: "", title: "Source \($0)", description: "")
        }
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(id: "1", key: "photograph", origin: "provenencia", label: "Photograph", description: ""),
            CatalogSourceType(id: "2", key: "book", origin: "provenencia", label: "Book", description: ""),
        ]
        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(id: "1", key: "date_taken", origin: "provenencia", label: "Date taken", dataType: "date", description: ""),
        ]
        store.fileCountByProject[projectDir] = 8

        let model = OnboardingModel(store: store, folders: .previewEmpty())
        model.session = PreviewFixture.identity
        model.project = PreviewFixture.project
        model.activeProjectDir = projectDir
        model.phase = .home(projectDir: projectDir, userID: PreviewFixture.identity.userID)
        return model
    }
}
#endif
