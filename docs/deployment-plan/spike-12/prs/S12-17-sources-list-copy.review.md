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
