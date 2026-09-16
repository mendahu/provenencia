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
    @State private var session: WorkspaceSession
    @State private var catalogCounts: CatalogCounts
    @State private var countsToast: VocabularyToast?
    @Environment(SignOutCoordinator.self) private var signOutCoordinator
    @Environment(NavigationCoordinator.self) private var navigationCoordinator

    init(model: OnboardingModel, projectDir: String, userID: String) {
        self.model = model
        self.projectDir = projectDir
        self.userID = userID
        _workspace = State(initialValue: WorkspaceModel())
        let session = WorkspaceSession(
            projectKey: ProjectKey(projectDir: projectDir),
            store: model.store
        )
        let navigation = WorkspaceNavigation()
        navigation.onLocationCommit = { location in
            session.apply(location: location)
        }
        _session = State(initialValue: session)
        _navigation = State(initialValue: navigation)
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
        // Leading, not the default center: if a destination ever does insist
        // on more width than the window has, the overflow has to spill off
        // the trailing edge. Centering it slides the fixed-width sidebar
        // under the traffic lights.
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
        .environment(catalogCounts)
        .environment(navigation)
        .environment(session)
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
            presentHistoryIssueIfNeeded()
            signOutCoordinator.isAvailable = true
            signOutCoordinator.action = { [model] in
                Task { await model.signOut() }
            }
            bindNavigationCommands()
        }
        .onChange(of: navigation.canGoBack) { _, _ in bindNavigationCommands() }
        .onChange(of: navigation.canGoForward) { _, _ in bindNavigationCommands() }
        .onChange(of: navigation.currentLocation) { _, _ in bindNavigationCommands() }
        .onChange(of: navigation.lastHistoryIssue) { _, _ in presentHistoryIssueIfNeeded() }
        .onChange(of: model.project?.uuid) { _, uuid in
            if let uuid, !uuid.isEmpty {
                navigation.attachProject(uuid: uuid)
                presentHistoryIssueIfNeeded()
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

    private func presentHistoryIssueIfNeeded() {
        guard let issue = navigation.lastHistoryIssue else { return }
        switch issue {
        case .loadFailed:
            countsToast = VocabularyToast(
                title: String(localized: L10n.Workspace.navigationHistoryLoadFailedTitle),
                body: String(localized: L10n.Workspace.navigationHistoryLoadFailedBody),
                tone: .danger
            )
        case .persistFailed:
            countsToast = VocabularyToast(
                title: String(localized: L10n.Workspace.navigationHistoryPersistFailedTitle),
                body: String(localized: L10n.Workspace.navigationHistoryPersistFailedBody),
                tone: .danger
            )
        }
        navigation.acknowledgeHistoryIssue()
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
        let model = OnboardingModel(store: store, folders: .previewEmpty())
        model.session = PreviewFixture.identity
        model.project = PreviewFixture.project
        model.activeProjectDir = projectDir
        model.phase = .home(projectDir: projectDir, userID: PreviewFixture.identity.userID)
        return model
    }
}
#endif
