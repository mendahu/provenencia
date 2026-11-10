# Review — S16-34 Tests run without showing a window or opening the researcher's project

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- S16-34 is a new Batch 3 ("Test infrastructure", slice 3a) in Spike 16, with one working rule, "(Batch 3) A test run shows no window and touches no install", from the user's 2026-11-10 request and the machine crash that day.
- The test host's activation policy is `.accessory` (no Dock icon, no focus stealing) (D-3).
- The PR does not claim to fix `WorkspaceDestinationHostTests`' timeouts; it measures the suite five times before and after, one run at a time (D-11).
- No version bump.

## Round 1

### X-1 · blocker · `AppLaunch.prepare()` runs before `NSApplication` exists
D-2 and D-4 called a shared `prepare()` (fonts and `WindowDrag.startTrackingPresses()`) from `main()` before `App.main()`, without saying whether the press monitor is safe before `NSApplication` is up.
**Answer:** Fixed — with M-3: no shared `prepare()`. `ProvenenciaApp.init()` stays as it is, so the shipped launch order does not move (D-2); `TestHostApp.init()` registers fonts and sets `.accessory` inside `App.main()` (D-3).

### X-2 · blocker · The trait's isolation and construction are not specified
How the main-actor tracker is built, bound to the task-local and closed from the nonisolated `provideScope`, and on a throw, was left to the implementer; `PVSelectPlacementTests` is not a `@MainActor` suite.
**Answer:** Fixed — D-5 puts the `@TaskLocal` on a nonisolated `enum HostedWindowScope`; D-6 and the §6 sketch give `provideScope` (build with `await`, bind, `closeAll()` after return and in `catch`); Edges says `PVSelectPlacementTests`' tests are `@MainActor` and the trait hops itself; a stop condition covers a concurrency warning.

### X-3 · major · Removing the holds leaves unused bindings
**Answer:** Fixed — D-10: bind unread tuple elements to `_`; rig return types are Out.

### X-4 · major · `PVFormDialogClickTests` windows: defaults and lifecycle undecided
**Answer:** Fixed — D-8: `window()` keeps its name, signature and 200 × 120 size, builds through `testWindow(size:styleMask:)` (`defer: false`), and the tests' own `defer` blocks run before `closeAll()`.

### X-5 · major · T-4 and T-5 are weaker than their rows imply
A never-shown window is already not visible, so T-4's assertion passed without `close()`. T-5 had no size.
**Answer:** Fixed — T-4 counts `NSWindow.willCloseNotification` (2) and checks the list is empty; T-5 uses 10 × 10, expects exactly one issue, and closes the window itself. Same point as M's "Outside my lens" CON note.

### X-6 · major · STEP-4's "before" run has no safe procedure, and its check can fail
**Answer:** Fixed — the "before" runs are STEP-1, on the clean branch before any change; the "after" runs are STEP-5; each Check is that the table is recorded, and a failing run is recorded, not fixed (D-11). Results go in the PR body (the ship log is the orchestrator's).

### X-7 · minor · Comments name `ProvenenciaApp.init`
**Answer:** Fixed — by M-3 `ProvenenciaApp.init()` stays, so `PVTypography.swift`'s and `WindowDragRegion.swift`'s comments stay true.

### X-8 · minor · Small gaps
**Answer:** Fixed — STEP-4's search widened and its expected output named; STEP-6 says "the last sentence of line 16"; STEP-2's check adds `PVFontTests`; a stop condition covers the activation policy. `docs/macos-client-patterns.md:223` stays true and is not touched.

### M-1 · major · T-2 filters out tagged test windows, so it cannot catch a test window that shows
**Answer:** Fixed — D-9: no identifier; T-2 asserts no visible window at all; a stop condition names a sheet that shows.

### M-2 · major · Moving `ProvenenciaApp.init()` makes comments and a doc row false
**Answer:** Fixed — with M-3 the two code comments stay true; STEP-6 updates `docs/macos-client-patterns.md`'s `App/` row. Line 232 ("Keep `ProvenenciaApp` thin (scenes, menus)") stays true.

### M-3 · major · `AppLaunch` adds three types and a shared `prepare()` the test host only half needs
**Answer:** Fixed — D-2, D-3: one `@main enum ProvenenciaMain` and `TestHostApp` in `ProvenenciaMain.swift`; `ProvenenciaApp.init()` unchanged; the speculative Next adds line is gone.

### M-4 · major · The DoD sweep hits 102 lines of generated code
**Answer:** Fixed — §9's sweep searches `macos/ProvenenciaTests` and the docs, and `macos/App` without `Platform/Generated/**`; §4 names the generated file.

### M-5 · minor · D-8 leaves out the two panels
**Answer:** Fixed — D-8 registers `NSOpenPanel()` and `NSSavePanel()` with the tracker; STEP-4's search covers `NS(Window|Panel|OpenPanel|SavePanel)\(`.

Nits (logged): `M-6 · macos/ProvenenciaTests/HostedWindow.swift · testWindow could read hostedWindow`.

## Round 2

Both reviewers checked every round-1 answer against the plan and the code: X-1 – X-8 and M-1 – M-5 resolved, M-4 partly (see M-8), M-5 reshaped (see M-7).

### X-9 · major · STEP-3's check cannot run where the step puts it
The test bundle does not compile until STEP-4 removes `PVSelectPlacementTests`' `orderFront:` arguments.
**Answer:** Fixed — STEP-3's check is the app target build; `HostedWindowTests` runs in STEP-4's check; STEP-3 and STEP-4 are one commit.

### X-10 · major · The panels are closed without `isReleasedWhenClosed = false`
**Answer:** Fixed — D-5: `HostedWindows.register` sets `isReleasedWhenClosed = false` on every window it receives.

### X-11 · minor · T-1 and T-2 name no framework or isolation
**Answer:** Fixed — §8 names `@Suite @MainActor struct AppLaunchTests` and `@Suite(.hostedWindows) @MainActor struct HostedWindowTests`; T-1 is `hostRunsTheTestApp`.

### X-12 · minor · `aSheetOnASheetIsInside` sheet-attach path
**Answer:** Declined — the first and second stop conditions cover a failing expectation and a visible sheet; no plan change needed.

### M-7 · blocker · D-8 registers the panels a second way, which skips the missing-trait check
**Answer:** Fixed — D-7 splits the builder: `trackTestWindow(_:)` *(new)* registers, sets the release flag through `register`, and records the missing-trait issue; `testWindow` returns `trackTestWindow(window)`; D-8's panels are `trackTestWindow(NSOpenPanel())` and `trackTestWindow(NSSavePanel())`; tests never call `register`; STEP-6's skill sentence names `trackTestWindow`.

### M-8 · blocker · The DoD sweep fails on the skill sentence STEP-6 adds
**Answer:** Fixed — STEP-6's sentence no longer names `withExtendedLifetime` ("Do not hold the window yourself or build an untracked window"), so the §9 sweep can return no hit.

Outside-lens notes: T-4's windows are built with `testWindow` under a fresh tracker bound by `HostedWindowScope.$current.withValue`, so they do not touch STEP-4's `NSWindow(` check (Fixed in T-4). `aSheetOnASheetIsInside` relying on `sheetParent` for a parent never ordered in is the likeliest stop; the stop conditions already cover it.
