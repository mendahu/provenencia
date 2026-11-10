# S16-34 — Tests run without showing a window or opening the researcher's project

**Status:** Landed in [#772](https://github.com/mendahu/provenencia/pull/772) (2026-11-10)
**Spike / slice:** Spike 16 · Slice 3a — Quiet test runs
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the app looks the same; only the test host and the test helpers change)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s16-34-hidden-test-windows`
**Version bump:** none — no change a researcher can see in the shipped app; the launch only differs when Xcode hosts the test bundle ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 3) A test run shows no window and touches no install"; S16-03's ship log on [#769](https://github.com/mendahu/provenencia/pull/769) (the `WorkspaceDestinationHostTests` timeouts under load)
**Landed:** No version bump, as planned. Differences from the plan: `HostedWindows()` is built without `await` in `provideScope` (the sketched `await` warns); T-4 counts `willClose` with one observer per window and a small counter (the shared observer warned); the `add-swift-test` sentence leaves out the literal `window.sendEvent` because `check-guidance-names` rejects an undefined name. D-12 and D-13 were added while implementing.
**Planned:** 2026-11-10 against `main` @ `16efd08f`
**Review:** [`S16-34-hidden-test-windows.review.md`](S16-34-hidden-test-windows.review.md) — Fits with changes — Ready with changes, 2 rounds; no open blocker; To user: none

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

Running the Swift tests stops putting windows on the researcher's screen and stops opening their real project. Today the test bundle is hosted inside Provenencia.app, which launches as it does for a researcher: it reads the install identity, reopens the active project and mounts the workspace, so every test run loads real catalog data beside the tests. And nine test files put borderless windows (up to 2000 × 400 and 1200 × 800 points) in front of whatever the researcher is doing, then never close them. This PR gives the test host a launch with no window and no project, keeps every test window off screen, and closes each one when its test ends, through one Swift Testing trait.

**Why now.** On 2026-11-09 and 2026-11-10 the full suite failed `WorkspaceDestinationHostTests`' 15-second waits under load, on `main` and on the S16-03 branch alike (2 of 3 runs with three suites at once), and three suites at once took the researcher's Mac down on 2026-11-10. The host app loading a real project during the run is the heaviest unrequested work in it. The researcher never asked for the windows and wants them gone (user, 2026-11-10).

**What stays.** The shipped app's launch order, scenes, commands and windows; `ProvenenciaApp.init()`; every test's assertions and timeouts; Swift Testing's parallel runs; the test bundle id and its host (`TEST_HOST` stays Provenencia.app).

## 2. User stories

**US-1.** As a researcher running the tests, when the suite runs, Provenencia does not open my project or show its window, so that my data and my screen are left alone.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Xcode or `xcodebuild test` hosts the test bundle | `ProvenenciaMain.main()` runs `TestHostApp`, whose only scene is `Settings`; no `OnboardingView` is built, so no install identity is read and no project is opened |
| AC-1.2 | The shipped app launches with no `XCTestConfigurationFilePath` in its environment | `ProvenenciaApp` runs as today, its `init()` unchanged: same scenes, commands, fonts and `WindowDrag` tracking |
| AC-1.3 | The test host is running | Its activation policy is `.accessory`: no Dock icon, and it does not take focus from the researcher's app |

**US-2.** As a researcher running the tests, no test window appears on my screen, and each one is closed when its test ends.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Any test hosts a view with `hostInBorderlessWindow`, or builds a window with `testWindow(size:styleMask:)` | The window is never ordered in: `window.isVisible` is false for the whole test |
| AC-2.2 | A test under the `.hostedWindows` trait ends, passing or throwing | Every window it built is closed and dropped; no test calls `withExtendedLifetime` to hold one |
| AC-2.3 | A test builds a test window without the `.hostedWindows` trait | The test records an issue naming the trait, so it fails instead of leaking a window |
| AC-2.4 | The full suite runs | At any point, `NSApplication.shared.windows` holds no visible window |

**US-3.** As a contributor, I add a window-hosted test the same way every time.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A contributor reads `add-swift-test` | It names `.hostedWindows` and the helpers, and says the window is never shown and is closed by the trait |

## 3. Scope

| In | Covered by |
| --- | --- |
| The suite measured one run at a time on the base, before any change | STEP-1 |
| `@main enum ProvenenciaMain` *(new)* choosing `ProvenenciaApp` or `TestHostApp` *(new)* | STEP-2, AC-1.1 – AC-1.3 |
| `HostedWindows` *(new)* tracker and the `.hostedWindows` trait *(new)*; the builders never order a window in and register it | STEP-3, AC-2.1 – AC-2.3 |
| Every window-building test moved onto the trait, its `withExtendedLifetime` holds and `orderFront:` arguments removed; `PVFormDialogClickTests`' titled windows and panels tracked by the same mechanism | STEP-4, AC-2.2, AC-2.4 |
| The suite measured one run at a time at the head | STEP-5 |
| `add-swift-test` and `docs/macos-client-patterns.md` say where the process entry is and how to host a test window | STEP-6, AC-3.1 |

| Out | Home |
| --- | --- |
| The root cause of `WorkspaceDestinationHostTests`' timeouts, if it survives this PR | S16-99 triages it from this PR's ship log (STEP-5 records whether it still fails) |
| Changing any test's timeout, wait or assertion | not planned |
| A test plan (`.xctestplan`), test parallelism settings, or CI changes | not planned |
| Go tests | not planned — they open no window |
| Changing rig return types (`mountHost`, `ArtifactRegionOverlayTests`' rigs, `GraphCanvasViewportControllerTests`' rig) | not planned — unused tuple elements are bound to `_` (D-10) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| App entry | `macos/App/App/ProvenenciaApp.swift` `ProvenenciaApp` | `@main`. `init()` registers bundled fonts (`PVFontRegistration.registerBundledFontsIfNeeded()`) and arms `WindowDrag.startTrackingPresses()`. `body` is one `WindowGroup { OnboardingView() … }` with `.commands`. The same launch runs when Xcode hosts the test bundle |
| Onboarding | `macos/App/Features/Onboarding/OnboardingView.swift` `OnboardingView` | `init(store: any GenealogyStore = GoStore(), …)`; `.task { await model.load() }` |
| Onboarding load | `macos/App/Features/Onboarding/OnboardingModel.swift` `OnboardingModel.load()` | Reads the live identity folder (`OnboardingFolders.liveIdentity()`), `store.installIdentity`, then `store.activeProject`; when the active project exists it calls `enterHome`, which sets `.home` and mounts `WorkspaceView` |
| Workspace | `macos/App/Features/Workspace/WorkspaceView.swift` `WorkspaceView` | Builds a `WorkspaceSession` on the real `GoStore`; `.task` loads the name format and refreshes catalog counts; `.onAppear` applies the restored location. Under tests this runs on the researcher's real active project |
| Test host | `macos/Provenencia.xcodeproj/project.pbxproj` `TEST_HOST` | `$(BUILT_PRODUCTS_DIR)/Provenencia.app/Contents/MacOS/Provenencia`; the test bundle loads into the running app. The project's folders are synchronized groups (`ProvenenciaApp.swift` has no `project.pbxproj` entry) |
| Test detection | `macos/App` | No code checks for a test run (`rg "XCTestConfigurationFilePath\|isRunningTests" macos/App` has no hit) |
| Press tracking | `macos/App/Platform/WindowDragRegion.swift` `WindowDrag.startTrackingPresses()` | Adds an `NSEvent` local monitor and a `didMove` observer; its doc says it is installed from `ProvenenciaApp`. No test calls it: `WindowDragTests.swift` drives `pressBegan()` and `windowMovedUnderPress()` directly |
| Window helper | `macos/ProvenenciaTests/HostedWindow.swift` `hostInBorderlessWindow(_:windowSize:hostSize:orderFront:)`, `hostInBorderlessWindow(nsView:size:)`, private `borderlessWindow(size:)` | Builds a borderless `NSWindow` (`defer: false`) at the screen origin and calls `orderFrontRegardless()` (the SwiftUI overload unless `orderFront: false`; the AppKit overload always). Never closes it. Its doc tells callers to hold the window with `withExtendedLifetime(window) {}` |
| Callers | `WorkspaceDestinationHostTests.swift` (`mountHost`, 1200 × 800), `PVWidthClassTests.swift` (2000 × 400), `PVPageColumnsTests.swift`, `SourcePageOverviewLayoutTests.swift`, `PVSegmentedControlTests.swift`, `PVSelectPlacementTests.swift` (passes `orderFront: false` at lines 144 and 162), `GraphCanvasViewportControllerTests.swift` (`nsView:`, `@Suite(.serialized)`), `ArtifactPDFViewportPageTests.swift` (`nsView:`), `ArtifactRegionOverlayTests.swift` (`nsView:`) — all in `macos/ProvenenciaTests/` | 17 calls; 24 `withExtendedLifetime` holds across these nine files, plus one in the helper's doc. All suites are `@MainActor` except `PVSelectPlacementTests`, whose tests are each `@MainActor`; none builds a window from a detached task |
| Titled test windows | `macos/ProvenenciaTests/PVPanelFormDialogTests.swift` `PVFormDialogClickTests` | `window()` builds a 200 × 120 `.titled` `NSWindow` with `defer: true`. `aParentWindowClickIsOutside` adds a child window and removes it in a `defer`; `aSheetOnASheetIsInside` calls `workspace.makeKeyAndOrderFront(nil)`, begins two sheets and ends them in a `defer`, which shows a titled window on screen. `anOpenPanelModalCountsAsInside` builds `NSOpenPanel()` and `aSavePanelEventWindowIsInside` builds `NSSavePanel()`; neither is shown or closed |
| Guidance | `.agents/skills/add-swift-test/SKILL.md` line 16 (its last sentence), line 53 | Line 16 points at `hostInBorderlessWindow(nsView:size:)` for AppKit layout; nothing about the trait, closing, or the window being shown. Line 53: "Host: Provenencia.app." |
| Client patterns | `docs/macos-client-patterns.md` line 13 (`App/` row), line 232 | "Process entry: scenes, menus … `ProvenenciaApp.swift`"; "Keep `ProvenenciaApp` thin (scenes, menus)" |
| Generated code | `macos/App/Platform/Generated/engine.pb.swift` | Holds 102 `withExtendedLifetime` lines; protoc output, never edited, so every sweep excludes it |

**Precedent.** `6f173924` ("Share one borderless window builder between the test hosting helpers.") made `borderlessWindow(size:)` the one builder; this PR keeps one builder and extends it. `b14ca03a` ("Hold test windows with withExtendedLifetime in every width-class and select test.") added the per-test holds that the trait replaces; use its file list as the sweep's starting point, not its mechanism.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The test-host check is `ProcessInfo.processInfo.environment["XCTestConfigurationFilePath"] != nil`, read once in `ProvenenciaMain.main()` | Xcode and `xcodebuild` set it for every hosted test run, XCTest and Swift Testing alike; it needs no scheme edit, so CI and local runs agree | A launch argument in the scheme — one more setting to keep in step with CI; `NSClassFromString("XCTestCase")` — true in any process that links XCTest, says nothing about hosting |
| D-2 | `@main enum ProvenenciaMain` *(new)* in `macos/App/App/ProvenenciaMain.swift` *(new)*: `static func main()` calls `TestHostApp.main()` when D-1's check holds, else `ProvenenciaApp.main()`. `ProvenenciaApp` loses only `@main`; its `init()` and `body` are unchanged | The test host never builds the app's scene, its coordinators or `OnboardingView`, so nothing reads the install; the shipped app's launch order does not move, because its fonts and press monitor still start inside `App.main()` | A shared `prepare()` called before `App.main()` — moves the press monitor ahead of `NSApplication` and adds a launch concept for one shared call; an `if` inside `WindowGroup` — still opens an empty window and builds the coordinators and commands; skipping work inside `OnboardingModel.load()` — the window still shows, and every later launch path needs its own guard |
| D-3 | `TestHostApp` *(new)*, in the same file, is an `App` whose `init()` calls `PVFontRegistration.registerBundledFontsIfNeeded()` and `NSApplication.shared.setActivationPolicy(.accessory)`, and whose `body` is `Settings { EmptyView() }`. It does not arm `WindowDrag` | A `Settings` scene opens no window at launch, and `App.main()` still runs `NSApplication`, which the hosted tests need. Layout tests measure text in the bundled fonts, so the host registers them; no test relies on the press monitor (§4). `.accessory` gives no Dock icon and no focus stealing while windows still work. The `init()` runs inside `App.main()`, after `NSApplication` exists | `.prohibited` — AppKit may refuse windows to a prohibited app; a `WindowGroup` with `EmptyView` — opens a window |
| D-4 | Test windows are never ordered in. The helper's `orderFront:` parameter goes; no builder calls `orderFrontRegardless()`, and `aSheetOnASheetIsInside` drops `makeKeyAndOrderFront(nil)` | SwiftUI and AppKit lay out, run `.task` and `onAppear`, and deliver `sendEvent` and `mouseDown` to a window that is not on screen: on 2026-11-10 the four window-hosted suites that had failed (35 tests) passed with `window.orderOut(nil)` in the helper | Ordering the window in at an off-screen origin — still a window the window server composites, and still visible on a second display |
| D-5 | `HostedWindows` *(new)* is a `@MainActor final class` with `private(set) var windows: [NSWindow]`, `func register(_ window: NSWindow)` (sets `isReleasedWhenClosed = false`, then appends) and `func closeAll()` (calls `close()` on each window, then empties the list). Its task-local lives on a nonisolated enum, `enum HostedWindowScope { @TaskLocal static var current: HostedWindows? }` *(new)*, so the trait can bind it from a nonisolated context | A `@MainActor` class is `Sendable`, so it can be a task-local value and cross into `provideScope`; keeping the `@TaskLocal` off the actor-isolated class avoids an isolated static under strict concurrency | `@TaskLocal` as a static on `HostedWindows` — isolated to the main actor, unreachable from the nonisolated `provideScope` |
| D-6 | `HostedWindowsTrait` *(new)* conforms to `TestTrait`, `SuiteTrait` and `TestScoping`; `isRecursive` is true; `scopeProvider(for:testCase:)` returns `self` only when `testCase` is non-nil, so each test case gets its own tracker. `provideScope` is the §6 sketch: it builds the tracker on the main actor, binds `HostedWindowScope.$current` around the test body, and calls `closeAll()` after the body returns or throws. `extension Trait where Self == HostedWindowsTrait { static var hostedWindows: Self }`. A suite opts in with `.hostedWindows` beside its other traits | One place opens and closes every test window, on pass and on throw; the per-test `withExtendedLifetime` holds go; a parallel test cannot close another's window | A closure-scoped `withHostedWindow { }` — `ArtifactRegionOverlayTests` and `GraphCanvasViewportControllerTests` return windows from rig functions, so every rig would be rebuilt; closing in `deinit` — a nonisolated `deinit` cannot touch `NSWindow` |
| D-7 | One tracking function and one builder. `trackTestWindow<W: NSWindow>(_ window: W) -> W` *(new)* registers the window with `HostedWindowScope.current` (which sets `isReleasedWhenClosed = false`); with no tracker it calls `Issue.record("Build test windows under the .hostedWindows trait.")`; it returns the window either way. `testWindow(size:styleMask:)` *(new name for `borderlessWindow(size:)`, now internal)*, default `styleMask: .borderless`, `defer: false`, origin `.zero`, builds the window and returns `trackTestWindow(window)`. Both `hostInBorderlessWindow` overloads build through `testWindow`. Tests never call `HostedWindows.register` directly | A missing trait fails the test loudly instead of leaking a window; `register` setting `isReleasedWhenClosed = false` keeps `close()` from over-releasing any window the tracker holds, the hand-built panels included | `preconditionFailure` — takes down the whole test process; a silent fallback — the leak this PR removes |
| D-8 | `PVFormDialogClickTests` gains `.hostedWindows`. `window()` keeps its name and signature and returns `testWindow(size: CGSize(width: 200, height: 120), styleMask: [.titled])` (its `defer: true` becomes the builder's `false`). The `NSOpenPanel()` and `NSSavePanel()` in its two panel tests are built as `trackTestWindow(NSOpenPanel())` and `trackTestWindow(NSSavePanel())`. The tests' own `defer` blocks (`removeChildWindow`, `endSheet`) stay and run before the trait's `closeAll()`; `close()` on a window whose sheet has ended is harmless | Every window a test builds goes through `trackTestWindow`, so the missing-trait check and the release flag cover panels too, and a future dialog test copies a call with the check built in | Leaving that suite on its own `NSWindow(` — the one test that still shows a window |
| D-9 | No test window carries an identifier, and T-2 asserts that no window in the app is visible, test windows included | D-4 says nothing is ever ordered in, so "no visible window" is the whole check; exempting test windows would hide the failure US-2 is about | A test-window identifier that T-2 filters out — skips exactly the sheets that could reach the screen |
| D-10 | Where a removed hold leaves a tuple element unread, bind it to `_` at the call (`let (_, host) = …`, `let _ = mountHost(…)`); rig return types stay | Removing the hold must not leave an unused-variable warning, and check-pr allows no Swift warning; changing rigs is Out | Changing rig return types — a refactor this PR does not need |
| D-11 | This PR changes no test's timeout, wait or assertion, and does not claim to fix the timeouts. STEP-1 runs the full suite five times on the base before any change; STEP-5 runs it five times at the head. Each run is recorded with its result, duration and failing tests; a failing run is recorded, not fixed. Never two `xcodebuild`s at once | The cause is not proven; the host's real project load is the leading suspect, so the two tables say whether this PR removed it. Three suites at once took the researcher's Mac down on 2026-11-10 | Raising the 15-second timeouts — hides the cause; the three-parallel repro — it crashed the machine |
| D-12 | The `PVSegmentedControlTests` suite and its file are deleted. Unit and integration tests do not synthesize mouse or key events through a window (`window.sendEvent`) (user, 2026-11-10). Its two tests clicked SwiftUI chips through `window.sendEvent`, which needs a window on screen; what they checked is SwiftUI's `Button` and `.disabled`, not this app's logic. `add-swift-test` gains the rule (STEP-6). Tests that call an AppKit view's own handler directly (`ArtifactRegionOverlayTests`' `overlay.mouseDown(with:)`) are not window event dispatch and stay | Found in implementation: with the window off screen the two tests fail and pass again once it is ordered in (ship log, round 0). Low-cost tests are the user's bar for this suite | An opt-in off-display window for this suite — keeps an end-to-end click test the user does not want; pressing segments through accessibility — still an end-to-end test of SwiftUI's control |
| D-13 | `PVOutsideClick.isInsideSheet(eventWindow:sheet:modalWindow:)` gains a last parameter `parent: (NSWindow) -> NSWindow? = { $0.sheetParent ?? $0.parent }`, which its loop calls instead of reading `sheetParent ?? parent`; the production caller in `PVOutsideClick.swift` passes nothing and behaves as today. `PVFormDialogClickTests.aParentWindowClickIsOutside` and `aSheetOnASheetIsInside` stop calling `addChildWindow` and `beginSheet` (and their `defer` blocks go): each passes a `parent` closure backed by a `[ObjectIdentifier: NSWindow]` map (sheet → workspace; child sheet → parent sheet), with the same assertions | Found in implementation: `beginSheet` on a never-shown window orders a titled window on screen, so T-2 fails in 4 of 5 full runs (ship log). Under the user's bar (D-12), these are unit tests of the walk, not of AppKit's sheets | A real sheet that the test orders out — still an end-to-end test of AppKit, and the window is on screen until then; exempting it in T-2 — D-9 |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `macos/App/App/ProvenenciaMain.swift` *(new)*: `@main enum ProvenenciaMain`, `struct TestHostApp` | file / type | `macos/App/App/ProvenenciaApp.swift` `ProvenenciaApp` | `ProvenenciaApp` is the researcher's scene; the choice of app has to sit above it. `TestHostApp` repeats one call of `ProvenenciaApp.init()` (font registration), the one mechanism called from two places |
| `HostedWindows`, `HostedWindowScope`, `HostedWindowsTrait`, `Trait.hostedWindows` *(new)* in `macos/ProvenenciaTests/HostedWindow.swift` | type / trait | `hostInBorderlessWindow` and `borderlessWindow(size:)` in the same file; `rg "TestScoping\|TestTrait\|SuiteTrait\|@TaskLocal" macos` — no hit | Extends the one helper file; the repo has no Swift Testing trait to reuse |
| `testWindow(size:styleMask:)` | func (renamed) | `borderlessWindow(size:)` | Same builder, one more parameter, now internal so `PVFormDialogClickTests` uses it |
| `trackTestWindow(_:)` *(new)* | func | `borderlessWindow(size:)`'s window setup | The tracking half of the builder, split out so a window the builder cannot make (a panel) gets the same check |

**Builds on.** The shared builder from `6f173924`; Swift Testing's `TestScoping` (Swift 6.1+; the toolchain is Swift 6.3.3 with Xcode 26.6).
**Deletes.** the `PVSegmentedControlTests` file (D-12); the `addChildWindow`, `beginSheet` and `endSheet` calls in `PVFormDialogClickTests` (D-13); `@main` on `ProvenenciaApp`; the helper's `orderFront:` parameter and both `orderFrontRegardless()` calls; the two `orderFront: false` arguments in `PVSelectPlacementTests`; every `withExtendedLifetime` hold in the nine caller files and the helper's doc sentence about it; `PVFormDialogClickTests`' own `NSWindow(` builder and its `makeKeyAndOrderFront(nil)`.
**Next adds.** A new window-hosted test is `.hostedWindows` on its suite plus a helper; nothing else.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| The test host launches `TestHostApp` | Every hosted test | Tests that needed the app's launch work get fonts from `TestHostApp.init()`; no test reads app windows (`rg "NSApp.windows\|NSApplication.shared.windows" macos/ProvenenciaTests` has no hit at the base) or calls `startTrackingPresses` |
| A test window that is never visible | The nine caller files and `PVFormDialogClickTests` | D-4's evidence covers four suites; STEP-4 runs the rest alone, and a failure that needs an on-screen window is a stop condition |

### Scale

None — no catalog write, list or loop changes. The suite (1,873 tests at the S16-03 head) is measured in STEP-1 and STEP-5, one run at a time.

### Edges

- **Assumptions.** A task-local reaches the builder only from the test's own task and its child tasks. Every caller builds from the test body or a rig it calls; none uses `Task.detached` (`rg "Task.detached" macos/ProvenenciaTests` has no hit at the base). Parallel tests each get their own tracker, so `closeAll()` closes only its test's windows. `PVSelectPlacementTests` is not a `@MainActor` suite, but its tests are, and the trait hops to the main actor itself (§6 sketch).
- **Neighbors.** `GraphCanvasViewportControllerTests` becomes `@Suite(.serialized, .hostedWindows)`. Suites in `PVPanelFormDialogTests.swift` that build no window (`PVPanelTests`, `PVFormDialogCopyTests`, `PVFormDialogControlsTests`, `PVSheetDismissTests`, `PVFormDialogAccessibilityTests`, `PVFormDialogContentTests`) do not gain the trait.
- **Removed inputs.** None — no data is decided, excluded or deleted.
- **Real data.** None — no table, seed or configuration drives a rule here.
- **Untrusted size.** None — no data from outside the app is read.

### Schema

None — no catalog change.

### Go core

None — no Go change.

### FFI

None — no wire change.

### macOS

```swift
// macos/App/App/ProvenenciaMain.swift (new)
@main
enum ProvenenciaMain {
    static func main() {
        if ProcessInfo.processInfo.environment["XCTestConfigurationFilePath"] != nil {
            TestHostApp.main()
        } else {
            ProvenenciaApp.main()
        }
    }
}

/// The app Xcode hosts the unit tests in: no window, no identity, no project.
struct TestHostApp: App {
    init() {
        PVFontRegistration.registerBundledFontsIfNeeded()
        NSApplication.shared.setActivationPolicy(.accessory)
    }
    var body: some Scene { Settings { EmptyView() } }
}
```

```swift
// macos/ProvenenciaTests/HostedWindow.swift
@MainActor final class HostedWindows {
    private(set) var windows: [NSWindow] = []
    func register(_ window: NSWindow) { window.isReleasedWhenClosed = false; windows.append(window) }
    func closeAll() { windows.forEach { $0.close() }; windows.removeAll() }
}

enum HostedWindowScope {
    @TaskLocal static var current: HostedWindows?
}

struct HostedWindowsTrait: TestTrait, SuiteTrait, TestScoping {
    var isRecursive: Bool { true }
    func scopeProvider(for test: Test, testCase: Test.Case?) -> Self? { testCase == nil ? nil : self }
    func provideScope(
        for test: Test, testCase: Test.Case?,
        performing function: @Sendable () async throws -> Void
    ) async throws {
        let windows = await HostedWindows()
        do {
            try await HostedWindowScope.$current.withValue(windows) { try await function() }
        } catch {
            await windows.closeAll()
            throw error
        }
        await windows.closeAll()
    }
}

extension Trait where Self == HostedWindowsTrait {
    static var hostedWindows: Self { Self() }
}

@MainActor func trackTestWindow<W: NSWindow>(_ window: W) -> W
@MainActor func testWindow(size: CGSize, styleMask: NSWindow.StyleMask = .borderless) -> NSWindow
@MainActor func hostInBorderlessWindow<V: View>(_ view: V, windowSize: CGSize, hostSize: CGSize? = nil) -> (NSWindow, NSHostingView<V>)
@MainActor func hostInBorderlessWindow(nsView view: NSView, size: CGSize) -> NSWindow
```

States: none — the app's views do not change. Accessibility: none.

### Strings

None — no researcher-facing copy. The `Issue.record` text is for contributors.

## 7. Steps

**STEP-1 — Measure the suite on the base**
Files: none committed. Results go in the PR body.
Do: D-11. Before any change, on the branch cut from the planning base with a clean tree, run `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` five times, one after another. Never start a second `xcodebuild` while one runs. Record each run: pass or fail, the seconds on its "Test run with" line, and the failing tests.
Check: the five-row "before" table is written down.

**STEP-2 — The test host launches with no window and no project**
Files: `macos/App/App/ProvenenciaMain.swift` *(new)*, `macos/App/App/ProvenenciaApp.swift`, `macos/ProvenenciaTests/AppLaunchTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-1 – D-3: add the file (synchronized groups need no project edit), move `@main` from `ProvenenciaApp` to `ProvenenciaMain`. Write T-1.
Check: `-only-testing:ProvenenciaTests/AppLaunchTests` and `-only-testing:ProvenenciaTests/PVFontTests` pass (STEP-1's command with those filters); building and running the app from Xcode (⌘R) still opens onboarding or the last project, with its Dock icon.

**STEP-3 — One tracker builds and closes every test window**
Files: `macos/ProvenenciaTests/HostedWindow.swift`, `macos/ProvenenciaTests/HostedWindowTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-4 – D-7: the tracker, the scope enum, the trait, `testWindow(size:styleMask:)`, both helper overloads without ordering in; rewrite the helper's doc comment to name the trait and drop the `withExtendedLifetime` sentence. Write T-3 – T-5. `PVSelectPlacementTests` passes the removed `orderFront:`, so the test target builds again only after STEP-4: STEP-3 and STEP-4 land as one commit, and STEP-3's tests run in STEP-4's check.
Check: the app target builds (`xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO`).

**STEP-4 — Every window-building test runs under the trait**
Files: the nine caller files in §4 and `macos/ProvenenciaTests/PVPanelFormDialogTests.swift`.
Do: delete `PVSegmentedControlTests.swift` (D-12); add `.hostedWindows` to each remaining window-building suite (D-6, Edges); delete every `withExtendedLifetime` hold and both `orderFront:` arguments, binding unread tuple elements to `_` (D-10); make `PVFormDialogClickTests` follow D-8. Write T-2.
Check: `rg -n "withExtendedLifetime|orderFrontRegardless|makeKeyAndOrderFront|orderFront:" macos/ProvenenciaTests` has no hit; `rg -n "NS(Window|Panel|OpenPanel|SavePanel)\(" macos/ProvenenciaTests` prints only the builder in `HostedWindow.swift` and the two `trackTestWindow(NSOpenPanel())` / `trackTestWindow(NSSavePanel())` lines in `PVPanelFormDialogTests.swift`; `HostedWindowTests` and each touched suite pass alone with `-only-testing:ProvenenciaTests/<Suite>`, one at a time; the build has no warning.

**STEP-5 — Measure the suite at the head**
Files: none committed. Results go in the PR body.
Do: D-11, STEP-1's procedure at this PR's head.
Check: the five-row "after" table is written down beside the "before" table; a failing run names its tests and stays recorded, not fixed.

**STEP-6 — Docs**
Files: `.agents/skills/add-swift-test/SKILL.md`, `docs/macos-client-patterns.md`, `docs/deployment-plan/spike-16/completed.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row), this plan (status).
Do: in `add-swift-test`, replace the last sentence of line 16 with: "To lay a SwiftUI or AppKit view out in a window, use `hostInBorderlessWindow` or `hostInBorderlessWindow(nsView:size:)` from `macos/ProvenenciaTests/HostedWindow.swift` (or `testWindow(size:styleMask:)` for a bare window, and `trackTestWindow(_:)` around a panel the builder cannot make) on a suite marked `.hostedWindows`: the window is never shown and the trait closes it when the test ends. Do not hold the window yourself or build an untracked window. Do not synthesize mouse or key events through a window (`window.sendEvent`): test the model, the action closure, or the view's own handler instead." Change line 53 to "Host: Provenencia.app, which runs `TestHostApp` under tests: no window, no install identity, no project." In `docs/macos-client-patterns.md`, the `App/` row (line 13) becomes "Process entry (`ProvenenciaMain` picks the app or the test host), scenes, menus. Stay thin." with `ProvenenciaMain.swift`, `ProvenenciaApp.swift` in its last column. Then the Definition of done's docs lines.
Check: `python3 scripts/check-guidance-names.py` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/AppLaunchTests.swift` `hostRunsTheTestApp` *(new)*, in Swift Testing `@Suite @MainActor struct AppLaunchTests` | The hosted test process → `ProcessInfo.processInfo.environment["XCTestConfigurationFilePath"]` is non-nil and `NSApplication.shared.activationPolicy()` is `.accessory` | AC-1.1, AC-1.3, D-1, D-3 |
| T-2 | Swift | `macos/ProvenenciaTests/AppLaunchTests.swift` `noWindowIsVisible` *(new)* | `NSApplication.shared.windows.filter(\.isVisible)` → empty | AC-1.1, AC-2.4, D-4, D-9 |
| T-3 | Swift | `macos/ProvenenciaTests/HostedWindowTests.swift` `hostedWindowIsNeverShownAndIsTracked` *(new)*, in `@Suite(.hostedWindows) @MainActor struct HostedWindowTests` | Under `.hostedWindows`, host a 200 × 100 `Text` → the window's `isVisible` is false and `HostedWindowScope.current?.windows.count` is 1 | AC-2.1, D-4, D-7 |
| T-4 | Swift | `macos/ProvenenciaTests/HostedWindowTests.swift` `closeAllClosesAndForgets` *(new)* | A fresh `HostedWindows` bound with `HostedWindowScope.$current.withValue(fresh)` around two `testWindow(size: CGSize(width: 10, height: 10))` calls (so they register with it, not the suite's tracker), and an observer counting `NSWindow.willCloseNotification` for them; `closeAll()` → the count is 2 and `windows` is empty | AC-2.2, D-5 |
| T-5 | Swift | `macos/ProvenenciaTests/HostedWindowTests.swift` `buildingWithoutTheTraitRecordsAnIssue` *(new)* | Inside `withKnownIssue`, `HostedWindowScope.$current.withValue(nil) { testWindow(size: CGSize(width: 10, height: 10)) }` → exactly one issue is recorded; the test closes the returned window itself | AC-2.3, D-7 |
| T-6 | Swift | the eight remaining caller suites and `PVFormDialogClickTests`, unchanged assertions | Run under `.hostedWindows`, off screen → pass | AC-2.2, AC-2.4, D-4, D-8, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PVFormDialogClickTests.aSheetOnASheetIsInside` (sheets on a window never ordered in), `PVFontTests` and the width-class suites (fonts now registered by `TestHostApp`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: `rg -n "withExtendedLifetime|orderFrontRegardless|orderFront:|borderlessWindow\(" macos/ProvenenciaTests .agents .cursor/rules docs/*.md` and `rg -n "orderFrontRegardless|borderlessWindow\(" macos/App --glob '!Platform/Generated/**'` have no hit outside this plan; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI change
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: `docs/macos-client-patterns.md` `App/` row (STEP-6); `add-swift-test` updated
- [ ] PR description: what shipped, AC evidence, the before-and-after tables from STEP-1 and STEP-5, deviations, manual check
```

## 10. Manual check

1. Run the tests once in Xcode (⌘U) with another app in front: no Provenencia window, Dock icon or test window appears, and the front app keeps focus.
2. While they run, open Activity Monitor's Open Files and Ports for the Provenencia process: no file in your project folder is open.
3. (AC-1.2) Build and run the app (⌘R): it opens onboarding or your last project as before, with its Dock icon and menus.
4. (AC-3.1) Read `add-swift-test`'s window sentence: it names `.hostedWindows` and the helpers, and says the window is never shown and the trait closes it.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A test needs a window on screen (key status, sheets, popovers) | STEP-4 runs each touched suite alone; D-4's evidence covers four suites |
| The machine is overloaded by test runs | Every run is one `xcodebuild` at a time (D-11) |
| `TestScoping`'s API differs from the §6 sketch in this toolchain | Stop condition below |

**Stop and report if:**

- A test in §4's callers or `PVFormDialogClickTests` fails only because its window is off screen (it passes with the window ordered in and fails without): name the test and the failing expectation; do not order the window in.
- T-2 finds a visible window (a sheet attached in `aSheetOnASheetIsInside`, or an app window): name the window's class and the test that built it.
- `XCTestConfigurationFilePath` is absent, or the activation policy is not `.accessory`, in the hosted test process (T-1 fails).
- `TestScoping`, `scopeProvider(for:testCase:)`, `HostedWindowScope.$current.withValue` or `await HostedWindows()` do not compile as §6 sketches with Xcode 26.6, or compile with a concurrency warning.
- The shipped app built from Xcode does not open onboarding or the last project after STEP-2.
- A test outside §4's files fails after STEP-2 because the host no longer builds the app's scene or arms `WindowDrag`.

## 12. Open questions

None.

## 13. PR

**Title:** S16-34 — Tests run without showing a window or opening the researcher's project
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; STEP-1's and STEP-5's tables; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-34-hidden-test-windows.md`.
