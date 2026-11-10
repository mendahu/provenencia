# Ship log — S16-34 Tests run without showing a window or opening the researcher's project

**Plan:** [`S16-34-hidden-test-windows.md`](S16-34-hidden-test-windows.md) · **Branch:** `s16-34-hidden-test-windows` · **PR:** —
**State:** implementing (resumed after D-12)

Base: `main` @ `e70bba8c` (the plan was written against `16efd08f`; S16-03 (#769) and the plan (#771) landed since).

## Round 0 — implement
**Stopped** — §11 first stop condition: `PVSegmentedControlTests` fails only because its window is off screen.
- Off screen (the plan's design), alone: 2 tests, 3 issues — `selectSendsTheItemValue()` `:49` `waitUntil { chosen.values == [3] }`, `:51` `[3, 1]`; `disabledItemDoesNotSelect()` `:60` `[3]`. The test's `click(segment:in:)` builds mouse-down/up `NSEvent`s and calls `window.sendEvent`; nothing selects.
- Same run with `orderFrontRegardless()` restored in the SwiftUI helper: 2 tests pass. Reverted.

State: commit `7ab3d7e4` "Launch the test host with no window and no project." (STEP-2, T-1/T-2 pass). STEP-3/4 edits uncommitted in the tree; `HostedWindowTests` (T-3–T-5) pass; ten touched suites pass alone. Not pushed, no PR.

STEP-1 "before" (base `e70bba8c`, one run at a time): 5 of 5 PASS, 1873 tests — 16.2 s, 18.1 s, 22.9 s, 23.3 s, 16.9 s.

Implementer deviations (uncommitted): `provideScope` builds `HostedWindows()` without `await` (the sketch's `await` warns "no 'async' operations"); T-4 uses one `willCloseNotification` observer per window with a nonisolated counter (a shared observer warned "sending 'note'").

**User answer (2026-11-10):** unit and integration tests should not simulate mouse clicks or run end-to-end; they should be low-cost. Plan amended on this branch: D-12 deletes `PVSegmentedControlTests.swift` (its two tests clicked SwiftUI chips through `window.sendEvent`); `add-swift-test` gains the rule (STEP-6); the Batch 3 working rule gains "no mouse or key events synthesized through a window". `ArtifactRegionOverlayTests` calls the view's own `mouseDown(with:)` directly (not window dispatch) and stays. Resumed the implementer.

**Stopped again** (STEP-5) — §11 second stop condition: T-2 `noWindowIsVisible` sees a visible titled 200 × 152 `NSWindow` (bare `NSView` content) in 4 of 5 "after" runs: `PVFormDialogClickTests`' `window()`; `beginSheet` on a never-shown window orders it in. Commits so far: `7ab3d7e4`, `e387cde5` (plan amendment D-12), `c55a2cbe` (tracker, trait, suites, `PVSegmentedControlTests` deleted); STEP-4 rg checks clean; each suite passes alone.

STEP-5 "after" (head `c55a2cbe`, 1876 tests): run 1 pass 16.1 s; runs 2–5 fail on T-2 only, 15.3 / 16.2 / 14.7 / 14.7 s. `WorkspaceDestinationHostTests` failed in no run, before or after.

**Orchestrator decision (under the user's 2026-11-10 bar, D-12):** plan amended with D-13 — `PVOutsideClick.isInsideSheet` takes a `parent` closure defaulting to `sheetParent ?? parent`; the two tests pass a map-backed closure instead of real child windows and sheets; assertions unchanged. Also fixed the Decisions table split before D-12. Reported to the user; resumed the implementer.
