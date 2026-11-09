# Review — S16-07 One paged fold that moves focus

**Verdict:** (round 2 pending)
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

Also reviewed: brief [S16-D2](../design/S16-D2-person-page-folds.md) (maintainability Q1, Q3, Q5, Q6; one finding, folded into M-2).

## To user
None.

## Writer's own calls
- D-9: Family's siblings fold counts the siblings other than the person, so a page always adds what its label says (X-5). Neighbor: when the person is within the first 30 in Go's order, 30 others plus the person now show (31 tiles), where today 30 tiles include the person. No board frame draws this boundary.
- M-6 declined: `ConclusionFoldState` keeps the `Conclusion` prefix to match `ConclusionFold`, which the working rule names; S15-30 moves both to `Platform/` and owns any rename.
- D-11: the keeper also hands focus to the fallback when a focused stop disappears for any reason (a reload as well as a width change), keeping today's Timeline behavior (M-5).

## Round 1

### M-1 · blocker · Every caller re-types a fold's numbers; the Show more row would add more call sites
**Answer:** Fixed. D-1 makes `ConclusionFold` a `Hashable` value (`limit`, `minimumHidden`, instance methods); each content declares its folds once (D-1 lists them) and the row takes `fold:` (D-3, D-7). §6 Codebase fit, Deletes, macOS sketch, T-1, T-2 updated.

### M-2 · major · The Sources-only highlight is a flag on the shared row and a copy of the line's background
**Answer:** Fixed. D-16: one private `personSourcesHighlight(_:)` modifier in `PersonSourcesSection.swift` for the line and both Show mores; `ConclusionShowMoreRow` has no highlight parameter. Brief §6 inventory rows reworded.

### M-3 · major · D-7 adds a second focus state; Return would skip the claim
**Answer:** Fixed (with X-1). D-7: one `@FocusState hasKeyboard`, one `press()`, `.focused($hasKeyboard)` and `.pvActivatesOnReturn(when: hasKeyboard, press)`.

### M-4 · major · Sources names its first revealed item differently from the other sections
**Answer:** Fixed. D-6: `firstRevealed(by:)` in all three sections; Sources drops `Row.firstRevealedLineID` and `firstRevealedRowLineID` (Deletes). D-15: `foldedGraphIDs(by: Fold)`.

### M-5 · major · The keeper drops the Timeline's fallback when a card leaves without a resize
**Answer:** Fixed. D-11 rule (b): focus going nil with the last stop no longer shown hands the keyboard to its fallback, whatever the cause; T-10 has the reload case; §6 Reach and Edges list it.

### M-6 · minor · `ConclusionFoldState` is named after Conclusion pages, and S15-30 moves it to `Platform/`
**Answer:** Declined. The working rule "One paged fold that moves focus" names `ConclusionFold`; a neutral name for only the new type would split one pair of names. S15-30's scope is the move to `Platform/`; Next adds now says it owns any rename.

### M-7 · minor · S16-27 will need the fold state to be Hashable
**Answer:** Fixed. D-4 and the sketch declare `Hashable`.

### M-8 · nit · The `PVPersonCard` stop condition is answered by the code
**Answer:** Fixed. Precedent cites `PersonTimelineSection.swift` `card(_:)`; the stop condition is gone.

### X-1 · blocker · `hasKeyboard` conflicts with `.pvActivatesOnReturn(expand)`; signature open
**Answer:** Fixed. D-7 gives the full signature (`focus: FocusState<Focus?>.Binding`, `revealed: Focus?`) and the modifier wiring; Return runs `press()`.

### X-2 · blocker · The keeper modifier's contract and expiry are unspecified
**Answer:** Fixed. D-11 states the modifier's signature, rules (a)–(d), who runs the 400 ms expiry (`hand`), what clears `last` (`forgetLast()` a turn later), and why the order of the two `onChange`s does not matter. The sketch lists every method; T-10 covers each rule.

### X-3 · major · Gallery's deletions disagree between D-11 and Deletes
**Answer:** Fixed. D-12 lists what Gallery deletes and keeps (`scrollTarget`, `scrollRequest`), where Use as portrait → image goes (`reveal`), and the viewer-close call (`FocusKeeper.hand`).

### X-4 · major · Family's `firstRevealed(by:)` underspecified for three folds
**Answer:** Fixed. D-8 defines each fold's first revealed tile; the diagram reads it from the content; `PlacedMore` is unchanged. T-6 adds the parent-set case.

### X-5 · major · The siblings fold with the person appended makes the label wrong
**Answer:** Fixed. D-9 folds the other siblings and places the person outside the fold; T-6 has 130 siblings with the person 40th. Recorded under Writer's own calls.

### X-6 · major · `foldRule` and the Sources test helper are the old forms' last callers
**Answer:** Fixed. STEP-1 rewrites `foldRule`; STEP-5 names the Sources helper and `PersonRecordLinksContentTests`; STEP-3 shows how a test builds a fold state.

### X-7 · major · STEP-2's check cannot run; five call sites, not four
**Answer:** Fixed. §4 lists the five call sites; STEP-2's check is the build and T-1's label cases, and it says a click still reveals all until STEP-3 – STEP-5.

### X-8 · major · `pages(toShow:)` and the dated index are not specified
**Answer:** Fixed. D-10 defines the argument (an event id), the list it indexes, the nil case and what `select` does with it; `isShown(_:)` moves to the content (D-13); T-2 and T-7 carry `minimumHidden` and the nil cases.

### X-9 · major · Undated and participant folds not mapped to the state
**Answer:** Fixed. D-14 names the static `row(…)` helper, the undated fold and the `revealed` map; D-13 gives `row(holdingCard:)` its type and nil case.

### X-10 · major · Sources init and highlight wiring are missing
**Answer:** Fixed. D-4 gives the `folds:` default; D-15 the set's meaning and the row's `showMoreHighlighted` parameter; D-17 the `.id(line.id)` (checked: no `.id(` in the file today).

### X-11 · major · Who moves focus for a Timeline fold after the change
**Answer:** Fixed. D-14: the item and row views take `revealed: (Fold) -> String?`; both Show mores claim through the row; `unfold` keeps the scroll for row folds only.

### X-12 · minor · The just-into-view scroll has no named host
**Answer:** Fixed. D-17 names the state, the modifier's place in each view and the `.id` tags; the page's `ScrollView` is the host, as for the Timeline.

### X-13 · minor · Docs step paths
**Answer:** Fixed. STEP-8 writes the full paths, the new `design/archive/` folder and the README row move. The § 9 sentence goes after the paragraph that follows the rules, so the rules' text and `add-design-brief`'s copy are untouched.

### X-14 · minor · Checks not run; no named fixture for the manual check
**Answer:** Declined in part. The board was read and verified by the writer on 2026-11-09 (plan **Board** line). The manual check names the data it needs instead of a fixture: no shipped fixture has 100+ events on one person. Builds and tests run in `ship-pr`.

### Current state corrections (executability)
**Answer:** Fixed. 14 `FocusClaim` calls; five Show more call sites; the `expanded:` counts per test file; the planning base notes no `macos/` change at `55f7f88`.
