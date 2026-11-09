# Ship log — S16-07 One paged fold that moves focus

**Plan:** [`S16-07-paged-fold-focus.md`](S16-07-paged-fold-focus.md) · **Branch:** `s16-07-paged-fold-focus` · **PR:** —
**State:** review 2

## Round 0 — implement
Head `a0837fe4` (local only: `git push` fails, github.com does not resolve from this Mac). Gate: implementer PASS on `4d60b7a1`; orchestrator re-run on `a0837fe4` below.
Report deviations:
1. Pre-existing test expectations changed beyond `expanded:` → `folds:`: Family siblings 31 tiles / 9 hidden (`foldsEachListAtThirty`), `expandedShowsAll` siblingsHidden 10; Timeline `foldsAndOrder` `firstRevealed` nil when nothing hidden. Accepted: D-9 and D-6 require them, and plan §6 "Neighbors" names the 31-tile change.
2. T-6 parent sets: "3 sets folded at 1" cannot fold at the fixed limit 30; test uses 32 sets folded at 30, asserting `l-p30-a30`. Accepted: same AC, the plan's fixture was infeasible.
3. Test helpers open "everything" with 100 pages.
4. Three behavior commits split by file, not by behavior.
5. Show more scrolls the revealed item into view on every press (D-17).
Not done in round 0: manual checks §10 (no app session); STEP-8 landing docs (need the PR number).
Orchestrator gate on `a0837fe4`: PASS (9/9; xcode-test 98s).

## Round 1 — review
Reviewers: CON, MNT, UX (`macos/App/**`), SEC+PRF (`macos/App/Features/**` content models paging lists up to 2,000 items). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Lifeline `select` writes fold state when it reveals nothing (D-10 "greater than the current") | Fix now → F-1 | Failing test given; `reveal(_:pages:)` writes `max(current, pages)` unconditionally |
| UX-1 | UX | major | Family `.id` after `.offset`: scroll targets the tile's pre-offset place | Fix now → F-2 | verified: `PersonFamilySection.swift:94-96`; `.offset` is a geometry effect, `.id` outside it sees the layout frame. D-17 does not fix the order |
| MNT-2 | MNT | major | Family respells its tile-id format in four `revealed` literals | Fix now → F-3 | verified: `PersonFamilyContent.swift` 125/148/181/186 beside `tile(...)` line 111 |
| MNT-1, CON-4, UX-4, SEC note | MNT/CON/UX/SEC | major | `FocusKeeper` calls the Gallery's scrolling `reveal` on every focus move while a stop is pending | To user → user chose split (2026-11-09) → Fix now → F-8 | verified: `FocusKeeper.swift:125`, Gallery `reveal` bumps `scrollRequest`. D-11 rule (a) specifies `took(…, revealedPending: pending.map(reveal))` and D-12 puts the scroll in `reveal`; the fix changes D-11's signature |
| CON-2 | CON | major | DoD: STEP-8 landing docs and §10 manual check not done | Fix now → F-4 (docs that need no PR number); PR-number lines at finish; manual check is the user's (step 8) | verified: map row still Planned, S16-D2 not archived |
| CON-3 | CON | minor | No test for `row(holdingCard:)` or Family `firstRevealed(by: .partners)` | Fix now → F-5 | code this PR adds |
| UX-2 | UX | minor | Translator comment on `conclusions.showMore` still says "shows the rest" / "how many are hidden" | Fix now → F-6 | this PR changes the number to the next page's size (D-2) |
| UX-3 | UX | minor | VoiceOver activation may not move focus (`hasKeyboard` false) | Fix now → VoiceOver line added to the PR body's manual check; code only if it fails | low confidence; settled by the check |
| MNT-3, UX-5 | MNT/UX | minor | Family and Sources each hand-roll the Show more scroll; it fires on pointer clicks too | Later PR → S15-30 | D-17 prescribes the per-section pair; S15-30 moves the row to the kit. Written into S15-30's map scope entry (`spike-15/deployment-plan.md`), commit `a09d5e4a` |
| CON-5, MNT-5 | CON/MNT | nit | Stale `revealed` doc comment in `PersonTimelineContent.swift:110` | Fix now → F-7 | stale comment in a file this PR changes |

Nits (logged): `MNT-4 · FocusKeeper.swift:321 · lost(isShown:) needn't be mutating` · `MNT-6 · ConclusionShowMoreRow pages(toShow:count:) · arithmetic instead of loop` · `MNT-7 · PersonSourcesSection · highlighted.map(foldedGraphIDs.contains) twice` · `MNT-8 · FocusKeeper.swift:293 · Hashable unneeded` · `CON-6 · PersonTimelineContentTests.swift:11 · pages param unused under max(pages, 100)`

SEC+PRF: no findings.

### Fix plan 1

**F-1 — A reveal that adds nothing leaves the fold state equal** (from CON-1)
Files: `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift`, `macos/ProvenenciaTests/ConclusionFoldTests.swift`.
Failing test first: `ConclusionFoldTests/revealThatAddsNothingLeavesStateEqual` — `var s = ConclusionFoldState<String>(); let b = s; s.reveal("a", pages: 0); #expect(s == b)`; also assert a reveal to fewer pages than current leaves the state equal.
Do: in `reveal(_:pages:)`, write only when `pages > self.pages(fold)` (D-10's condition, in one place).
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — Family's scroll id sits inside the offset** (from UX-1)
Files: `macos/App/Features/Conclusions/PersonFamilySection.swift`.
Do: move `.id(placed.id)` above `.offset(x:y:)` (next to `.focused`). Leave `.accessibilityIdentifier` where it is.
Check: build; scripts/check-pr.sh PASS. (Verified on screen in the manual check.)

**F-3 — One spelling of Family's tile id** (from MNT-2)
Files: `macos/App/Features/Conclusions/PersonFamilyContent.swift`.
Do: add `static func tileID(column:group:personID:) -> String`, the counterpart of `PersonTimelineContent.participantLinkID`. Route `tile(...)`, the person's own tile, and the four `revealed` assignments through it.
Check: `rg -n '"l-|"r-|"\\\(column\)-' macos/App/Features/Conclusions/PersonFamilyContent.swift` prints only the line inside `tileID`; existing Family tests pass unchanged.

**F-4 — Landing docs that need no PR number** (from CON-2)
Files: `docs/deployment-plan/spike-16/design/S16-D2-person-page-folds.md` → `design/archive/`, `docs/deployment-plan/spike-16/design/README.md`.
Do: the STEP-8 lines that do not name the PR number: `git mv` the brief to `design/archive/` and update `design/README.md` and any links to it (`rg -n "S16-D2-person-page-folds" docs .agents .cursor`). Leave the map row, `completed.md`, plan Status and Landed note for the orchestrator's finish step.
Check: scripts/check-pr.sh PASS (plan-docs).

**F-5 — Tests for the two untested focus mappings** (from CON-3)
Files: `macos/ProvenenciaTests/PersonTimelineContentTests.swift`, `macos/ProvenenciaTests/PersonFamilyContentTests.swift`.
Do: Timeline: `row(holdingCard:)` returns the event row for a participant card id, nil for a row id and for an unknown id (beside T-5). Family: 31 partner units with no children → `firstRevealed(by: .partners)` is the 31st unit's partner tile; a one-parent unit at index 30 → its first child's tile (beside T-6). Build expected ids with F-3's `tileID`.
Check: tests pass.

**F-6 — Translator comment says the next page** (from UX-2)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: change the `conclusions.showMore` comment in both to "Button under a folded list on a Conclusion page that shows the next page of it; argument is how many that page adds". Key and value unchanged.
Check: scripts/check-pr.sh PASS (localizable).

**F-7 — Stale `revealed` doc comment** (from CON-5, MNT-5)
Files: `macos/App/Features/Conclusions/PersonTimelineContent.swift:110`.
Do: say it is the first item the next page would reveal, absent when the fold hides nothing.
Check: scripts/check-pr.sh PASS.

**F-8 — The keeper maps without scrolling** (from MNT-1, CON-4, UX-4; user decision 2026-11-09 amends D-11 and D-12)
Files: `macos/App/Platform/FocusKeeper.swift`, `macos/App/Features/Conclusions/ConclusionGallerySection.swift`, `macos/App/Features/Conclusions/PersonTimelineSection.swift`, `macos/ProvenenciaTests/FocusKeeperTests.swift`, `docs/macos-client-patterns.md`.
Do: replace the `reveal: (Stop) -> Stop` parameter on `focusKeeper(...)`, `FocusKeeper.reclaim` and `FocusKeeper.hand` with two: `map: (Stop) -> Stop` (pure; `took`'s `revealedPending` and the claim use it) and `scroll: (Stop) -> Void` (only `reclaim` calls it, before the claim). Gallery: `map` is today's `.useAs` → `.image` rule; `scroll` sets `scrollTarget` and bumps `scrollRequest`. Timeline: `map: { $0 }`, `scroll: { _ in }` (or its current scroll, if `reveal` did one). Update the doc comment on `focusKeeper(...)` and the `macos-client-patterns.md` keeper sentence to the new names. The plan's D-11/D-12 text stays; the orchestrator records the amendment in the Landed note.
Failing test first: `FocusKeeperTests/tookDoesNotScroll` — a counting `scroll` closure: `hand` a stop (count 1), then `took` the mapped stop and `took` another stop; count stays 1.
Check: the test passes; `rg -n "reveal:" macos/App/Platform/FocusKeeper.swift macos/App/Features/Conclusions` prints no keeper call site; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `b2eb6377` (local only). Gate PASS (orchestrator re-run; xcode-test 87s). F-1 – F-8 done.
- F-1: `revealThatAddsNothingLeavesStateEqual` red first (`["a": 0]` != `[:]`), green after.
- F-8: `focusKeeper`/`reclaim`/`hand` take `map:` and `scroll:`; `tookDoesNotScroll` red (no compile) first. Deviation: added non-mutating `FocusKeeper.claimed(map:scroll:) -> Stop?` that `reclaim` routes through, so the test can run without a `FocusState`.
- F-2 not checked on screen (manual check 2).
- F-4: brief moved to `design/archive/`, README and links updated; PR-number lines left for finish.

## Round 2 — review
Reviewers: CON, MNT, UX (each had a Fix-now major). Skipped: SEC+PRF (no findings in round 1). Scope `a0837fe4...b2eb6377`.
