# Review — S12-17 Copy Source from the Sources list

**Verdict:** Fits with changes — Ready with changes → fixed (round 1). The plan stays `Blocked` on S12-02 landing and on board S12-D6; round 2 runs when it is refreshed after S12-02 lands.
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet). The maintainability reviewer also read brief S12-D6.

## To user
- **A kit right-click menu modifier (M-1).** The row menu is the third hand-wired right-click menu (`SourcePageIdentityHeader.coverMenu`, `ConclusionDetailPage.portraitMenu`). Each names an item three times (keyboard title, item, accessibility action), and the copies have drifted: the cover menu has no VoiceOver action. A `pvActionMenu(items:)` modifier in `DesignSystem/Components/ContextMenu/` would own the state, derive titles and indexes, and add the accessibility actions. Recommendation: a small follow-up PR after S12-17 that adds the modifier and moves all three menus onto it (the cover menu gains its VoiceOver action), rather than folding it into S12-17, because the cover and portrait menus put the trigger and the panel on different views and the modifier needs a placement option designed for that. Until you answer, S12-17 hand-wires the row like the portrait menu and STEP-4 logs the idea.

## Writer's own calls
- No menu title and no hover hint on the row (D-3, D-6).
- The menu sits on the row's page zone only, so S12-14's graph sub-rows do not inherit it (D-2).
- Tests: one list-model test; the prefill, the refresh and the count are owned by S12-02 and existing tests (D-8).

## Round 1

### M-1 · major · Third hand-wired right-click menu; copies drifted
**Answer:** To user (above). The Out line and STEP-4's idea line now describe the kit modifier and the placement constraint; the brief says "kit menu modifier", not "recipe".

### M-2 · major · Duplicate row call site gets a third duplicated closure
Proposal: one `ForEach` call site reading `model.graphProgress(for:)` / `model.graphCountsLoading`; delete `SourcesSplitRowProgress`.
**Answer:** Declined — `SourcesSplitRowProgress` exists so "one Source's Get-one merge repaints that row" (its doc comment, `macos/App/Features/Sources/SourcesListView.swift` L459): it observes the progress handle per row with `@Bindable`. Reading the model getters in the `ForEach` body moves that observation to the list, so every merge repaints every row. The cost of keeping it is one more one-line `onCopy` closure. Out line updated with the reason.

### M-3 · minor · T-2 and part of T-1 re-test S12-02's model
**Answer:** Fixed — T-1 asserts only `copying?.original.id`; T-2 dropped; D-8 names the owners (S12-02 T-3, T-9; `createReturnsSourceAndPublishesCount`).

### M-4 · minor · Brief and plan name different precedents; "recipe" is the wrong layer
**Answer:** Fixed — both point at the portrait menu (the one with a VoiceOver action); "kit menu modifier" in the brief and STEP-4.

### M-5 · nit · No discoverability hint
**Answer:** Fixed — D-6 records that the row has no `.help` / `.accessibilityHint`, and why.

### X-1 · major · T-2's waits race
**Answer:** Fixed — T-2 dropped (M-3).

### X-2 · major · T-1 / T-2 fixtures unspecified
**Answer:** Fixed — T-1 names `source(id:title:typeID:)`, `photoType()` and `makeModel(sources:types:)`.

### X-3 · minor · STEP-2's check
**Answer:** Fixed — STEP-2's check says it proves the build only; the menu is checked in STEP-3's manual check; D-9 covers Linux.

### X-4 · minor · STEP-4 formats and the start precondition
**Answer:** Fixed — STEP-1 opens with the start precondition; STEP-4 names the map status `Landed #N` and the `completed.md` paragraph. The archive folder and README Archived line come from S12-02's STEP-7, which lands first.

### X-5 · minor · Modifier order on `pageZone`
**Answer:** Fixed — D-2 gives the order after `.accessibilityIdentifier`: accessibility action, trigger, menu.

### X-6 · nit · "Four hand-wired menus"; line numbers
**Answer:** Fixed — STEP-4's line names the three right-click menus, the row's included as new. Line numbers are approximate and rechecked on refresh.

## Refresh — 2026-10-18 (after S12-02 landed, board agreed)

**Verdict:** Fits with changes — Ready with changes → fixed. No blocker; no round 2.
**Reviewers:** maintainability (opus), executability (sonnet), fresh context.

**User answers (2026-10-18).** M-1 above: build the kit menu modifier in this PR. Board S12-D6's web-only details (Shift-F10 opening, an "Actions for <title>" panel label): leave both out (D-10).

### Writer's own calls
- D-3: the trigger and the panel go on the same view, with no placement option. The row-level placement in the cover and portrait menus dates from before `71f32c35` moved the panel into a child `NSPanel`. A stop condition covers a panel that opens away from the pointer.
- D-14: the Citation composer Clear menu moves onto the items layer (refresh M-1). This follows from the user's "refactor the hand-wired menus here" answer.
- The right-click modifier is named `pvRightClickMenu`, not `pvActionMenu` (refresh M-3).

### Refresh round
- **M-1 · major · The Clear menu keeps parallel titles and indexes.** Fixed: D-2 is now two layers (`PVActionMenuItem` + `PVActionMenuRows`; `pvRightClickMenu`); D-14 moves the Clear menu onto the items layer; AC-2.6 and manual check step 11 were added.
- **M-2 · major · The `pvContextMenu` doc comment's ancestor advice becomes false.** Fixed: D-15, done in STEP-2.
- **M-3 · minor · "Action menu" already means every non-Select menu.** Fixed: the modifier is renamed `pvRightClickMenu`; D-13 rewords README L363–364.
- **M-4 · minor · The map scope entry doesn't name the kit work.** Fixed: STEP-6 updates the L113 scope cell and the L231 **Ships** line.
- **X-1 · major · The `ForEach` for accessibility actions was pseudo-code.** Fixed: D-2 now spells out the `ForEach` form.
- **X-2 · major · Steps 3 and 5 can't show the menus work.** Fixed: both steps now have a build check (XT), and §7 says the manual checks are the user's.
- **X-3 · major · The manual check had no fixture recipe.** Fixed: §10 now opens with setup steps.
- **X-4 · minor · pbxproj ids weren't given.** Fixed: STEP-2 names `FE005` and `FE006` and the ContextMenu group.
- **X-5 · minor · STEP-6 left the link and PR-number choices open.** Fixed: STEP-6 now gives the map texts, the link moves and when to fill in `#N`.
- **X-6 · minor · The memberwise init was only implied.** Fixed: D-2 now says "synthesized memberwise init, no custom init".
- **X-7, X-8 · notes.** Logged. The README sentence X-8 flagged is now covered by D-13.
