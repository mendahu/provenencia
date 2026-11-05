# Ship log — S14-31 Tab reaches a following text area

**Plan:** [`S14-31-tab-reaches-text-area.md`](S14-31-tab-reaches-text-area.md) · **Branch:** `s14-31-tab-reaches-text-area` · **PR:** #718
**State:** review 2

## Round 0 — implement
Head `b485ad89`. Gate PASS (orchestrator run: xcode-test 57s, all nine checks). Report deviations: version goes 0.3.3 → 0.3.4 (`main` was already 0.3.3; the plan allows it); `PVFieldFocus`'s binding init and `neighbours` are `@MainActor` (strict-concurrency warnings otherwise); plan header edited so `check-plan-docs` passes; `focus:` sits before `activateOnAppear:` at P6–P8. **The STEP-4 probe and §10 manual check were not run**: the implementer had no way to drive a window, so three §11 stop conditions are untested (Shift-Tab reaching `.onKeyPress`, `FocusClaim` landing in a `PVTextArea`, a `PVTabPair` `FocusState` moving focus in a sheet).

## Round 1 — review
Base `b6cac73d`, head `b485ad89`. Reviewers: CON (Sonnet), MNT, UX (`macos/App/**`). Skipped: SEC+PRF: no listed security path, no catalog data, no loop or count; the diff moves keyboard focus between two fields, and the `macos/App/Features/**` changes are views, not models.

CON: no code defect; every AC traced, call sites unchanged. Sweep and 8-site count re-checked by CON and MNT.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, UX-3, UX-1 | CON, UX | major | Stop-condition probes never run; Shift-Tab may not reach `.onKeyPress(.tab)` (P4 is the only input with `previous`) | Fix now → F-1; remainder To user | Verified: Landed note says not run. `NSBackTabCharacter` is a separate character and no repo code handles it. The probe is STEP-4, so running it is in plan. A kept hosted Tab test is an Out line, so F-1 is a throwaway. A failed stop condition goes to the user. §10 stays the user's |
| UX-2 | UX | major | Text area's router can attach to the Label's field editor | Fix now → F-2 | Verified: `firstDescendantTextView` has no `isFieldEditor` guard; `clearInsets` writes insets and `tabRouter.attach` to whatever it finds; the clearer now holds closures, so `updateNSView` reruns each body. Hierarchy unobserved, F-1 logs it. One condition either way |
| MNT-1 | MNT | major | "Handle or own focus" decided per field, a new copy in `PVTextArea` | Fix now → F-3 | Same-pass cleanup. Verified at all named lines. Takes the reviewer's smaller fix: D-5 and STEP-3 keep the owned path, so no `AnyView` on it |
| MNT-2 | MNT | major | Tab claim written three times | Fix now → F-4 | Same-pass cleanup. Verified `PVInput`, `PVComboBox`, `PVSelect` |
| MNT-5 | MNT | minor | Router test setup copied further | Fix now → F-5 | Same-pass cleanup |
| MNT-3 | MNT | minor | Clearer and router comments no longer true | Fix now → F-6 | Stale comments in a changed file. Comments only: D-5 names `PVTextEditorInsetClearer`, so the rename is not taken |
| MNT-4 | MNT | minor | `PVTabPair` has no `#Preview` | Fix now → F-6 | `add-ui-component` checklist |
| CON-2 | CON | minor | Landed note omits `focus:` argument order | Fix now → F-7 | Verified |
| UX-4 | UX | minor | P6 Tab skips Save and Cancel with Full Keyboard Access | Fix now → F-7 | Docs only: Landed note records the skip beside P4's; surfaced in the report |

Nits (logged): `CON-3, MNT-9 · plan Context · "dogfood note dogfood note" (F-7 fixes it)` · `MNT-6 · CitationComposerEvidenceColumn.swift:281,293 · make the handle non-optional` · `MNT-7 · PVInput.swift:104 · equalsFocus could be focus` · `MNT-8 · PVFieldFocus.swift · neighbours could use map`

### Fix plan 1

**F-1 — Prove Tab in a real window (throwaway)** (from CON-1, UX-1, UX-3; STEP-4 probe)
Files: a temporary test file only, never committed; save a copy to `/private/tmp/claude-501/-Users-jakerobins-Repos-provenencia/77219f8d-4d51-431f-acec-cf28dc2ae686/scratchpad/PVTabPairProbeTests.swift`, and remove it from the tree (and any project entry) before committing.
Do: host a `PVTabPair` with a `PVInput` and a `PVTextArea` in a titled, key-capable window (not the borderless helper; activate the app and `makeKeyAndOrderFront`), once directly and once inside a SwiftUI `.sheet`. Deliver real `NSEvent.keyEvent` keyDowns through `NSApp.sendEvent` or `window.sendEvent`: Tab (`"\t"`, keyCode 48) and Shift-Tab (`"\u{19}"`, keyCode 48, `.shift`). Also a pair with order swapped (text area first, input second) for P4: Shift-Tab from the input. Log `press.key`, `press.characters` and `press.modifiers` from `.onKeyPress`, and after each key the window's `firstResponder` and `isFieldEditor`. Type into the text area after the move and check its binding updates. Also log `nearestTextView()`'s type and `isFieldEditor` while the Label holds focus.
Report the four results: (a) does `.onKeyPress(.tab)` fire for Shift-Tab, (b) does `FocusClaim` land in the `TextEditor`, (c) the same inside a sheet, (d) does the router ever reach a field editor. If a result is inconclusive (the window cannot become key), say so. **Do not change D-6's filter** and do not stop the round on a failed result: report it and carry on with F-2 to F-7.
Check: the probe ran and the tree has no probe file.

**F-2 — The text area's router never takes a field editor** (from UX-2)
Files: `PVTextArea.swift`, `PVTextAreaTests.swift`.
Failing test first: `PVTextAreaTests/textViewSearchSkipsAFieldEditor` — a view holding an `NSTextView` with `isFieldEditor = true` before a plain `NSTextView`: the search returns the plain one; with only the field editor it returns nil.
Do: `firstDescendantTextView()` skips `isFieldEditor` text views and keeps searching. Widen the extension from `private` only as far as the test needs.
Check: the test passes; scripts/check-pr.sh PASS.

**F-3 — One resolution of handle or own focus** (from MNT-1)
Files: `PVFieldFocus.swift`, `PVInput.swift`, `PVComboBox.swift`, `PVSelect.swift`, `PVTextArea.swift`, `PVFieldFocusTests.swift`.
Failing test first: `PVFieldFocusTests/optionalHandleFallsBackToOwn` — nil handle answers with the own value; a handle answers with its own `isActive`.
Do: add `extension Optional where Wrapped == PVFieldFocus` with `isActive(own: Bool) -> Bool`, `focus(own: FocusState<Bool>.Binding)` and `resign(own: FocusState<Bool>.Binding)`. Every field uses them. `PVComboBox.fieldText` and `PVInput` apply focus through `PVFieldFocusModifier`, so each builds one `TextField`. `PVInput`'s own binding is `externalFocus ?? $ownedFocus`. No handle keeps today's `.focused` path; no `AnyView` on that path.
Check: `rg -n "if let (focus|equalsFocus) \{|focus != nil|focus\?\.isActive\(\) \?\?" macos/App/DesignSystem/Components` prints no focus-resolution line in the four field files (`if let` for an unrelated value is fine; list what it prints); `rg -c "TextField\(text:" PVInput.swift PVComboBox.swift` prints 1 each; scripts/check-pr.sh PASS.

**F-4 — One Tab claim on the handle** (from MNT-2)
Files: `PVFieldFocus.swift`, `PVInput.swift`, `PVComboBox.swift`, `PVSelect.swift`, `PVFieldFocusTests.swift`.
Failing test first: `PVFieldFocusTests/claimTabRunsTheNeighbourAndReportsHandled` — with a neighbour returns `.handled` and runs it once; with none returns `.ignored`.
Do: add `@MainActor func claimTab(shift: Bool) -> KeyPress.Result` on `PVFieldFocus`, built on `tabClaim(shift:)` (`KeyPress` has no test init, so it takes the Bool), and `var hasNeighbour: Bool`. All three fields call it. `PVInputTabClaim` keeps D-6's "no neighbour, no modifier" through `hasNeighbour`. `PVComboBox` sets `isOpen = false` before the claim as today.
Check: `rg -n "tabClaim\(shift" macos/App` prints only `PVFieldFocus.swift`; scripts/check-pr.sh PASS.

**F-5 — One router fixture in the text area tests** (from MNT-5)
Files: `PVTextAreaTests.swift`.
Do: a private fixture holding the text view, spy and router, with `send(_ selector:)`. All router tests use it, F-2's excepted. No assertion changes.
Check: `rg -c "doCommandBy: #selector" PVTextAreaTests.swift` prints 1; scripts/check-pr.sh PASS.

**F-6 — Comments and preview** (from MNT-3, MNT-4)
Files: `PVTextArea.swift`, `PVTabPair.swift`.
Do: the clearer's doc says it clears insets and carries the Tab neighbours to the router; the router's doc says Tab goes to the neighbour when set, else the key loop. Keep the type names. Add a `#Preview` to `PVTabPair` with a `PVInput` and a `PVTextArea`.
Check: scripts/check-pr.sh PASS.

**F-7 — Landed note** (from CON-2, UX-4, CON-3)
Files: `S14-31-tab-reaches-text-area.md`.
Do: Landed note adds that `focus:` sits before `activateOnAppear:` at P6–P8, and that P6's Tab skips the inline Save and Cancel when Full Keyboard Access is on, as P4 does. Fix the doubled "dogfood note dogfood note". Add the F-1 results and replace "were not run" with what was and was not run. Note F-3 and F-4 add handle helpers beyond D-2 (same-pass cleanup).
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `12b1668f`. Gate PASS (orchestrator run). F-1 to F-7 done. F-5's Check printed nothing rather than 1: the one call left passes `selector` as a parameter (`rg -c "doCommandBy:"` prints 1). F-3 and F-4 share one commit.

**F-1 probe** (`NSEvent` key-downs in a hosted titled window; window reported key by subclass, sheet window not key, so the sheet run is weaker):
- (a) **§11 stop tripped.** Shift-Tab arrives as `KeyEquivalent("\u{19}")`, so `.onKeyPress(.tab)` never fires for it (`probe.log` line 16). P4's `previous` claim is dead code; the same holds for the `.tab` cases in `PVComboBox.handleKey` and `PVSelect.handleTriggerKey`. In the probe the key loop still moved focus back.
- (b) `FocusClaim` lands in the `TextEditor`: Tab and Shift-Tab from the area both landed. Typing after Tab updates the binding.
- (c) Same inside a `.sheet`.
- (d) UX-2 confirmed: before F-2 the router sat on the Label's field editor, and typing after Tab raised `unrecognized selector undoManagerForTextView:` and hung. A blocker in the real sense; F-2 cleared it.

**To user:** D-6's `.onKeyPress(.tab)` filter cannot see Shift-Tab, so AC-3.2 rests on the key loop.

**To user, answered 2026-11-05.** "Widen the claim": the kit claims the back-tab key too; this amends D-6. Fix now → F-8.

### Fix plan 1 (continued)

Record the D-6 amendment in the plan's Landed note, marked "user, 2026-11-05". Do not rewrite the Decision row.

**F-8 — The Tab claim sees Shift-Tab** (from UX-1, CON-1; F-1 result (a))
Files: `PVFieldFocus.swift`, `PVInput.swift`, `PVComboBox.swift`, `PVSelect.swift`, `PVFieldFocusTests.swift`, the plan's Landed note.
Failing test first: `PVFieldFocusTests/backTabIsTheShiftTabKey` — `KeyEquivalent.backTab.character == "\u{19}"`, which is what the probe logged for Shift-Tab.
Do: add `extension KeyEquivalent { static let backTab = KeyEquivalent("\u{19}") }` in `PVFieldFocus.swift`, with a comment that AppKit reports Shift-Tab as this character, not as `.tab` with Shift.
- `PVInputTabClaim`: `.onKeyPress(keys: [.tab, .backTab], phases: .down)`, `focus.claimTab(shift: press.key == .backTab)`.
- `PVComboBox.handleKey`: `case .tab` is forward only: drop its two now-dead `press.modifiers.contains(.shift)` reads and claim with `shift: false`. New `case .backTab`: with a `previous` neighbour, close the list and claim; otherwise `.ignored` and leave `isOpen` alone, which is what Shift-Tab does today.
- `PVSelect.handleTriggerKey`: `case .tab` claims with `shift: false`; new `case .backTab` claims with `shift: true`.
- Update the Landed note: result (a) is now fixed, with the probe's re-run result.
Check: re-run the saved probe `/private/tmp/claude-501/-Users-jakerobins-Repos-provenencia/77219f8d-4d51-431f-acec-cf28dc2ae686/scratchpad/PVTabPairProbeTests.swift`, again uncommitted and removed afterwards: the P4 layout logs `CLAIM previous from second`. `rg -n "modifiers.contains\(.shift\)" macos/App/DesignSystem/Components/ComboBox macos/App/DesignSystem/Components/Select macos/App/DesignSystem/Components/Input` prints nothing. scripts/check-pr.sh PASS.

Head `ebd6ada4`. Gate PASS (orchestrator run). F-8 done; the re-run probe logged `CLAIM previous from second -> first` in the P4 layout.

## Round 2 — review
Scope `b485ad89...ebd6ada4`. Reviewers: CON (CON-1), MNT (MNT-1, MNT-2), UX (UX-1, UX-2, UX-3), each from a Fix-now major. Minors (MNT-3 to MNT-5, CON-2, UX-4): read by the orchestrator below.
