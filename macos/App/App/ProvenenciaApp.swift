import SwiftUI

@main
struct ProvenenciaApp: App {
    // Owned here (not inside the view hierarchy) because `.commands` builds
    // the app menu at the Scene level, outside the views that hold
    // `OnboardingModel` / `WorkspaceModel` — see `SignOutCoordinator` and
    // `NavigationCoordinator`.
    @State private var signOutCoordinator = SignOutCoordinator()
    @State private var navigationCoordinator = NavigationCoordinator()
    @State private var omnibarFocusCoordinator = OmnibarFocusCoordinator()

    init() {
        PVFontRegistration.registerBundledFontsIfNeeded()
    }

    var body: some Scene {
        WindowGroup {
            OnboardingView()
                .environment(signOutCoordinator)
                .environment(navigationCoordinator)
                .environment(omnibarFocusCoordinator)
                // `.hiddenTitleBar` alone still reserves the title bar's
                // height as a top safe area, leaving a blank strip above
                // our content instead of letting the stoplights float over
                // it — ignore that inset so our own header row is the one
                // row the stoplights sit in front of (S2-01 Frame 7).
                .ignoresSafeArea(.container, edges: .top)
                .pvAlignsTrafficLights()
        }
        // `contentMinSize` keeps the hard floor from the root view's
        // min frame, but lets `.defaultSize` open larger than content
        // ideal — unlike `.contentSize`, which hugs the loading-phase
        // minimum and never grows to the workspace default.
        .windowResizability(.contentMinSize)
        .defaultSize(WindowSizing.fittedDefaultSize)
        .windowStyle(.hiddenTitleBar)
        .commands {
            CommandGroup(after: .appInfo) {
                Button {
                    signOutCoordinator.signOut()
                } label: {
                    Text(L10n.Onboarding.signOut)
                }
                .disabled(!signOutCoordinator.isAvailable)
            }
            CommandGroup(after: .sidebar) {
                // Never `.disabled` these: Scene `.commands` often keeps the
                // first enabled state, so ⌘] stayed inert after Back created
                // a forward entry. `goBack` / `goForward` already no-op at ends.
                Button {
                    navigationCoordinator.goBack()
                } label: {
                    Text(L10n.Workspace.goBack)
                }
                .keyboardShortcut("[", modifiers: .command)

                Button {
                    navigationCoordinator.goForward()
                } label: {
                    Text(L10n.Workspace.goForward)
                }
                .keyboardShortcut("]", modifiers: .command)

                Button {
                    omnibarFocusCoordinator.focusOmnibar()
                } label: {
                    Text(L10n.Workspace.focusOmnibar)
                }
                .keyboardShortcut("k", modifiers: .command)
            }
        }
    }
}
