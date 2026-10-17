# Review — S10-57 Hold Shift to select; Select disarms after one rectangle

**Verdict:** Fits — Ready with changes (applied)
**Rounds:** 2. **Reviewers:** maintainability (claude-opus), executability (claude-sonnet)

## To user

None. The three product calls (Shift-click toggles, Shift rectangle adds, every rectangle release disarms) were answered on 2026-10-17 before review.

## Writer's own calls

- D-2: Shift is watched with a local `flagsChanged` monitor in `GraphCanvasDocumentView`, ignored while a text field has focus, reset on resign-key.
- D-6: Shift-click on a card's nested action (pencil, Add property) toggles the card and does not run the action.
- D-7: Shift-click on empty canvas still clears the selection (requirement 4 read literally; Finder keeps it).
- D-11: a Shift-press on the canvas counts even while the omnibar keeps focus, and lights Select for that press.
- Shift-drag of an unselected card keeps today's behavior (the drag replaces the selection with that card). It is an Out line.
- Select armed from the palette stays armed after a plain click on a card or on empty canvas; only a rectangle disarms it.

## Round 1

### X-1 · blocker · The Xcode checks and the manual check cannot run in this environment
**Answer:** Fixed — §7 gains a "Host" paragraph: on a host without `xcodebuild`, report each Swift check and §10 step as not run, never as passed.

### X-2 · major · D-2's monitor behavior is covered by no automated test and no step check
**Answer:** Fixed — `reportedShift(_:textHasFocus:)` is a pure static on `GraphCanvasDocumentView`, tested by T-9; STEP-3's check now includes §10 steps 1, 9, 10, 11.

### X-3 · major · Working rule text in STEP-5 leaves stale, contradictory sentences
**Answer:** Fixed — STEP-5 replaces the three sentences through the stale bridge sentence, and names every GraphCanvas README row it touches.

### X-4 · minor · Test construction choices not made
**Answer:** Fixed — T-1 names `marqueeController()`, T-2 names `GraphCanvasPointerPressTests.controller()` and the press point, T-8 names its setup.

### X-5 · minor · STEP-1 Check sentence is garbled
**Answer:** Fixed — the closure change is an unconditional Do line.

### X-6 · minor · Choices left open in the document view
**Answer:** Fixed — §6 names `nonisolated(unsafe)` tokens, the block observer on `queue: .main` with `MainActor.assumeIsolated`, seeding from `NSEvent.modifierFlags` on install, and that `deinit` does not touch `pointer`.

### X-7 · minor · D-1 and AC-1.4 wording
**Answer:** Fixed — T-8 asserts that a rectangle finished with Shift held disarms `armedSelect` and keeps `inputMode == .selecting`.

### X-8 · nit · Sheet and Shift
Logged. Select re-engages on sheet close if Shift is still held; that is intended.

### M-1 · major · Two sources of truth for "Shift is down" can disagree
**Answer:** Fixed — D-11: press-time Shift owns behavior, unguarded; `mouseDown` also sets the held state so paint agrees. AC-1.7 and §10 step 11 cover the omnibar-focused case.

### M-2 · major · The GraphCanvas README contract goes stale
**Answer:** Fixed — STEP-5 updates the Reuse contract callbacks cell and adds the Shift sentence, plus the Input mapping rows.

### M-3 · minor · Shift is deduped twice and stored twice
**Answer:** Fixed — only `GraphCanvasPointerController.setShiftHeld(_:)` dedupes; `setShiftSelectHeld` assigns plainly (D-8).

### M-4 · minor · `isSelectEngaged` only forwards `inputMode` for one caller
**Answer:** Fixed — removed; the palette reads `model.inputMode == .selecting` (D-9).

### M-5 · minor · Third "local monitor + resign-key observer" pairing
**Answer:** Fixed — Codebase fit cites `PVComboBox.startWatching(parent:)`, the count is 9, and the shared helper is an Out line logged here for a later hardening pass.

### M-6 · nit · "Next adds" overclaims
**Answer:** Fixed — reworded: a future host gets the signal and the toggle, and supplies its own engage precedence.

## Round 2

Maintainability: every round-1 answer holds; no new blocker. Verdict Fits.
Executability: every round-1 answer holds; no new blocker. Verdict Ready with changes.

### X-9 · major · Shift seed on install may fire before `onShiftHeldChanged` is wired
**Answer:** Fixed — STEP-3: `wirePointer()` calls `model.setShiftSelectHeld(pointer.shiftHeld)` after wiring; §6 notes the seed only sets the pointer flag.

### X-10 · major · STEP-3's check needs STEP-4's paint
**Answer:** Fixed — STEP-3 checks pointer behavior only; the paint halves of §10 steps 1, 10, 11 are STEP-4's check.

### X-11 · minor · STEP-3 does not name the `mouseDown` change
**Answer:** Fixed — STEP-3 Do names `mouseDown(with:)` (D-11).

### X-12 · minor · "three sentences" should be four
**Answer:** Fixed.

### X-13 · minor · Landed status wording not given
**Answer:** Fixed — STEP-5 gives the map suffix, the scope-entry line (copied from S10-20's), and the `completed.md` section form.

### X-14 · minor · Test suite placement
**Answer:** Fixed — T-1, T-2, T-3, T-4, T-5 are in `GraphCanvasPointerPressTests` (where `marqueeController()` and `controller()` live); T-9 is in `GraphCanvasPointerHitTestingTests`.

### X-15 · minor · T-8 call order
**Answer:** Fixed — T-8 lists the calls in order.

### X-16 · nit · README cell wording
**Answer:** Fixed — only the parenthetical changes.
