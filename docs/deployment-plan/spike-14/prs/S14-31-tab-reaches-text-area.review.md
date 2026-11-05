# Review — S14-31 Tab reaches a following text area

**Verdict:** Round 1: Reshape — Not ready. Plan rewritten on the reshaped design (user, 2026-11-05). Round 2 pending.
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **Reshape (M-1, M-2, M-4, M-5).** Use `.onKeyPress` like `PVComboBox`, move focus only through `FocusClaim`, and give each handle its Tab neighbour. Recommended. **Answered 2026-11-05: Reshape.**
- **Focus handles (M-3).** Unify `PVInputFocus`, `PVComboFocus` and `PVSelectFocus` into `PVFieldFocus` in this PR, in a preparatory PR, or not at all. Recommended in this PR. **Answered 2026-11-05: in this PR** (D-2, STEP-1).
- **Kit-wide rule (M-6).** Accept opt-in wiring and reword the rule, or add a lasting guard. Recommended opt-in and reword. **Answered 2026-11-05: opt-in, reword** (§5 Rules this plan bends, STEP-7).

## Writer's own calls

- `PVTabPair` and `PVFieldFocus` live in `Components/Input/` (M-7): the pair is a field-focus helper, not a control, so it has no README row of its own; the Input and TextArea rows carry the contract.
- The opening caret is left as it is today (D-10); the manual check clicks into the first field.
- `PVComboBox` does not read `next` / `previous`; it keeps its `onCommit` Tab claim (Out).
- `PVInput.focusField()` keeps its existing one-turn write for `activateOnAppear` (Out): it is unchanged code; the new moves all go through `FocusClaim`.

## Round 1

### M-1 · blocker · AppKit Tab router for PVInput beside PVComboBox's `.onKeyPress` claim
**Answer:** Fixed — D-6: `PVInput` adds `.onKeyPress(.tab, phases: .down)` in `PVComboBox.handleKey`'s shape. `PVInputTabRouter`, `PVInputTabHook`, the geometry match, their tests and two stop rows are gone. The stop condition is flipped: stop if `.onKeyPress` does not see Tab.

### M-2 · blocker · Hand-rolled one-turn focus writes beside FocusClaim
**Answer:** Fixed — D-3: neighbours are `{ FocusClaim.claim(binding, neighbour) }`. No new `DispatchQueue` in the kit; the Risk row's "only allowed variation" is gone.

### M-3 · blocker · Third focus-handle copy extended, not extracted
**Answer:** Fixed (user's answer) — D-2: `PVFieldFocus` replaces all three; both `install` closures go; STEP-1 is the rename; DoD sweep names the old types.

### M-4 · major · Neighbour on the handle, not a parallel `tabPair:` parameter
**Answer:** Fixed — D-3, D-4: handles carry `next` / `previous`; fields keep one `focus:` parameter; `PVTabPair { first, second in }` orders by argument, so P4 (`panelForm`) is wired (D-8, US-3). D-7 (assertion) and D-10 (nil `panelForm`) are gone.

### M-5 · major · D-9 put one call site's state in the router
**Answer:** Fixed — D-7: P1 passes `focus: model.isEditingBridge ? nil : second`; the router keeps one rule.

### M-6 · blocker · Opt-in wiring may bend "Tab into a text area is kit-wide"
**Answer:** To user → answered: opt-in, reword. §5 Rules this plan bends names it; STEP-7 rewords the rule and the scope entry. `PVFormDialog` row added to §4 (it touches only `cancelFocused`).

### M-7 · minor · Where PVTabPair and its doc live
**Answer:** Fixed — README Input / TextArea rows carry the contract, plus a `PVFieldFocus` row; § Interaction state is left alone. File stays in `Components/Input/` (Writer's own calls).

### X-1 · blocker · One-turn versus two-turn write contradictory
**Answer:** Fixed — every move is `FocusClaim.claim` (D-3).

### X-2 · blocker · Where the one-turn delay lives contradicts T-1
**Answer:** Fixed — the delay is inside `FocusClaim`; the router calls `onTab` / `onBacktab` synchronously (D-5), as T-1 expects.

### X-3 · blocker · D-9 fallback has no home
**Answer:** Fixed — fallback removed (D-7).

### X-4 · major · Hook attach point and coordinates
**Answer:** Fixed — the hook is gone (M-1).

### X-5 · major · Opening caret preconditions unverified
**Answer:** Fixed — D-10; AC-1.1, AC-1.2 and §10 click into the first field.

### X-6 · major · P6 inline edit not walked
**Answer:** Fixed — §4 row: `PVInlineEdit` has no blur commit and the input has no `onBlur`; D-9 says the pair wraps `PropertiesIdentityEditor.body`'s `VStack` content.

### X-7 · major · No test for PVTextArea(tabPair:) / PVInput wiring
**Answer:** Fixed — T-3 tests `PVInput.tabClaim(shift:focus:)` as a pure function; §8 says what is manual-only and why (a `FocusState` exists only in a mounted view). D-7's assertion is gone.

### X-8 · major · `labelField` is a computed property
**Answer:** Fixed — D-8 names `labelField(focus:)` and `transcriptionColumn(rows:focus:)` and both layouts.

### X-9 · major · PVTabPair init access
**Answer:** Fixed — §6 sketch gives an explicit `init(@ViewBuilder content:)`.

### X-10 · minor · Where P1's wrapper goes
**Answer:** Fixed — §6: inside the existing `VStack`, around the two `PVField`s.

### X-11 · minor · Map row, scope entry, working rule
**Answer:** Fixed — STEP-7 rewords the rule and the scope entry; STEP-8 names the map row and the "Where the notes went" row.

### X-12 · minor · Checks without verifiable output
**Answer:** Fixed — STEP-6 check counts `PVTabPair {` hits; STEP-7 check `rg`s `PVTabPair` in each doc; build checks are the `xcodebuild test` line, whose warnings `check-pr.sh` already gates.

### X-13 · minor · Temptations not closed
**Answer:** Fixed — the handles are unified (D-2); `PVTextArea` keeps its owned `FocusState` for the no-handle path (STEP-3); `isFieldFocused` reads the handle's `isActive` as today; `SourcePageIdentityHeader` is in Out.

### X-14 · minor · Stale wording
**Answer:** Fixed — §10 items cite AC-2.2; the sweep line names the three old types.
