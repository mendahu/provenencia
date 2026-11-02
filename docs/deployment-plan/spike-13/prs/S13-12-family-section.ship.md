# Ship log — S13-12 Family section

**Plan:** [`S13-12-family-section.md`](S13-12-family-section.md) · **Branch:** `s13-12-family-section` · **PR:** #676
**State:** ready

## Round 0 — implement
Base `018e5176` (`main`).
Head `4cbf430a`. PR #676 (draft). Gate PASS (orchestrator re-ran on the pushed head). Report deviations: (1) `a11yList` was renamed by #675; the tile uses `L10n.DesignSystem.spokenJoin`. (2) D-8 (b) added `AnyQueryHandle.isFetching` and `WorkspaceSession.isRefreshing(_:)`, files outside STEP-5's list. (3) T-2 (b): kinship folds a shared parent into the first set, so the second set is "Other parent". (4) `a11yLifeLineKnown` takes `part:`, as Strings has it (D-6 text says `parts:`). (5) Shared Swift fixtures in `PersonFamilyFixtures.swift`. (6) L10n keys landed in the step that first used them. Not run: AC-1.6, AC-2.6, AC-3.2, AC-3.3, AC-4.1, visual frames (manual check).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX for `macos/App/**` and `Localizable.xcstrings`; SEC+PRF for `api/ffi/**`, `core/database/**`, and `macos/App/Features/**` models.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC+PRF | major | Unbounded tiles on one Family page; layout quadratic in parent sets; Go `originSiblings` O(sets × siblings) | Linear: Fix now → F-2. Cap: To user → answered → F-1 | verified: `kinship/families.go:205` stops pairing past `MaxPivotEnd`, so each extra parent becomes its own set with no limit. The plan's Edges premise ("parents bounded at `MaxPivotEnd`") is wrong. `PersonFamilyLayout.init` loops over n sets with `filter { sets.contains(k) }` and `sets.min()`. `originSiblings` runs `place` per set over `all`. **User 2026-11-01:** "I don't ever want to hide actual data, so we should put a cap on it, but the cap should be really reasonable", then chose "30 per list + Show more". Supersedes D-14 |
| UX-1 | UX | major | Return does nothing on a focused relative's tile (AC-3.3) | Fix now → F-3 | verified: `PersonTile.swift` is a plain `Button`. `docs/deployment-plan/spike-14/app-health-audit.md:166` records that Return goes to the default button. `PVList.swift:331` and `CitationComposerSubjectGroups.swift:204` handle Return themselves |
| MNT-1 | MNT | major | Family fixtures copied in the app previews and the tests | Fix now → F-4 (same-pass) | verified: `FamilyPreviewData` and `FamilyFixtures` build the same four families. The `PVDeleteImpactPreviewData` pattern exists |
| MNT-2 | MNT | major | Fourth hand fold of `spokenJoin` over a list | Fix now → F-5 (same-pass) | verified: `rg 'reduce\(parts\[0\]\)'` finds 3 copies in `Features/Conclusions`, plus the tile's variant |
| MNT-3 | MNT | major | `SortByBirth` copies `sortEvents`' sort loop | Fix now → F-6 (same-pass) | verified by reading `order.go:40` and `compose.go:960` |
| MNT-4 | MNT | major | `macos-client-patterns.md:103` still says Person pages read only the detail key | Fix now → F-7 | verified: line 103 is unchanged and contradicts the added sentence |
| UX-3 | UX | minor | Place-only life line reads "died Hamburg" | Fix now → F-8 | `PersonLifeDisplay.swift` uses one key for date-only and place-only |
| CON-1 | CON | minor | No Go test for one parent in two multi-parent sets (frame 11c) | Fix now → F-9 | `families.go:206-210`; the existing "two sets" subtest has no spouse link |
| SEC-2, CON-2 | SEC, CON | minor | `Dictionary(uniqueKeysWithValues:)` traps on a duplicate tile id | Fix now → F-10 | `PersonFamilySection.swift:102`. Ids are scoped by column and group, so none duplicates today. Defense in code this PR adds |
| MNT-5 | MNT | minor | Siblings caption is a Bool; layout owns its text | Fix now → F-11 | `PersonFamilyLayout.swift:131` |
| MNT-6 | MNT | minor | 92pt tile height declared twice | Fix now → F-11 | `PersonTile.swift:57`, `PersonFamilyLayout.swift:32` |
| MNT-7 | MNT | minor | `lifeBefore` is a second date comparator in `compose.go` | Fix now → F-6 (same-pass) | `compose.go:719`. Refs are uppercase Crockford tokens (`core/ref`), so `dateKeyLess`'s lowercasing changes no researcher-visible order |
| UX-2 | UX | minor | Tab order may follow screen position | To user (manual check step 4) | low confidence; only a Tab pass settles it. If the order alternates between columns, it becomes a fix round |

Nits (logged): `CON-3 · S13-12-family-section.md:18 · Landed note could list the manual-only ACs` · `UX-4 · ConclusionDatePlaceLine.swift:28 · layoutPriority now also in the header` · `UX-5 · PersonTile.swift:115 · .clipped() cuts the focus card's shadow` · `UX-6 · PersonTile.swift:64,70 · .help shows only the name` · `MNT-8 · PVLiftHover.swift:18 · resting [] not Optional` · `MNT-9 · PersonFamilyContent.swift:104 · focus id built by hand` · `MNT-10 · PersonFamilyLayout.swift:78 · name the 120 floor` · `MNT-11 · ConclusionDatePlaceLine.swift:14 · Missing named for Person` · `MNT-12 · PersonTile.swift · PersonTileContent.id unread`

### Fix plan 1

Supersedes **D-14** ("No cap on tiles") and its § 6 Edges line, by the user's decision on SEC-1 (above). Record both in the plan's Landed note as superseded, with the user's words. Go still returns every member; only the first paint is bounded.

**F-1 — Fold every Family list at 30 behind "Show N more"** (from SEC-1)
Files: `macos/App/Features/Conclusions/PersonFamilyContent.swift`, `PersonFamilyLayout.swift`, `PersonFamilySection.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonFamilyContentTests.swift`, `PersonFamilyLayoutTests.swift`. Skills: `add-localized-string`, `add-swift-test`, `add-ui-component`.
Failing tests first, in `PersonFamilyContentTests`:
- `foldsEachListAtThirty`: 50 children with one partner → 30 child tiles and `hidden == 20`. 40 parent sets → 30 sets. 40 siblings → 30. 40 partner groups → 30 groups.
- `expandedShowsAll`: the same content with every fold expanded → every tile.
- `foldKeepsTheFocus`: 50 siblings with the person at index 45 → the first 30 siblings plus the person (31 tiles), and hidden 19.
- `foldDropsHiddenSetIndexes`: a sibling whose only set is beyond 30 keeps `sets == []`.

Do:
- `PersonFamilyContent.foldLimit = 30`.
- `enum Fold: Hashable { case parentSets, siblings, partners, children(Int) }`. `children` takes the group's index in Go's order, before the partner fold.
- `init(family:expanded: Set<Fold> = [], …)`.
- Each folded list carries the count it hides: `parentSetsHidden`, `siblingsHidden`, `groupsHidden`, and `Group.childrenHidden`.
- Siblings: keep the first 30 in Go's order. Also keep the person when it falls after that, in its order, after the 30. A sibling's `sets` keeps only indexes of the visible parent sets.
- `PersonFamilyLayout` places one "more" row of `PersonFamilyLayout.moreRowHeight` (28) after the last visible member of each folded list, inside the same VoiceOver group, and grows the height for it. No rail or line goes to a more row.
- The view draws each more row as a plain borderless kit button (reuse an existing `PV*` link or plain button style; add none). It reads `L10n.Conclusions.familyShowMore(count)`, "Show %lld more" with plural variants (one: "Show 1 more"). Return and Space activate it, as in F-3. Clicking inserts its `Fold` into an `@State var expanded: Set<PersonFamilyContent.Fold>` in `PersonFamilySection`. That state resets when `entityId` changes (`.onChange(of: entityId) { expanded = [] }`, or key the state's owner by `entityId`).
- Update T-9 `fiftyMembersKeepOrder` and T-13 `fiftySiblingsStack` to pass every fold expanded, so they keep testing 50 members. Add one `fiftySiblingsFolded` layout case: 30 tiles plus a more row, no overlap.

Check: the new and updated tests pass; `scripts/check-pr.sh` PASS.

**F-2 — Make the origin placement and its lines linear** (from SEC-1)
Files: `core/database/persondetail/family.go`, `macos/App/Features/Conclusions/PersonFamilyLayout.swift`.
Failing test first: none. Behavior is unchanged; `TestPersonFamilyOriginSiblings`, T-12 and T-13 must still pass unchanged.
Do:
- In `originSiblings`, replace the per-set `place` passes with one pass that buckets each entry:
  - `len(ParentSets) == sets` → the full bucket.
  - Any other entry with sets → the bucket of its largest set index. This matches the current descending loop, which places an entry at the first k that matches.
  - No sets → the recorded bucket.
- Emit the full bucket first, then the per-set buckets from `sets-1` down to 0, then the recorded bucket, each sorted by `byBirth`.
- In `PersonFamilyLayout.init`, build `members[k]` (sibling indexes per set) and each sibling's `minSet` in one pass over `content.siblings`. Use them instead of `filter { sets.contains(k) }` and `sets.min()` inside the `for k` loop.

Check: Go tests and Swift layout tests unchanged and passing; `scripts/check-pr.sh` PASS.

**F-3 — Return opens a focused tile** (from UX-1)
Files: `macos/App/Features/Conclusions/PersonTile.swift` (and the F-1 more row).
Do: on the relative's button, add `@FocusState` and `.onKeyPress(.return) { open(); return .handled }`, guarded by that focus state, following `macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift:204`. Apply the same handling to F-1's more row.
Check: builds; `scripts/check-pr.sh` PASS. AC-3.3 stays in the manual check.

**F-4 — One Family fixture set, in the app** (from MNT-1, same-pass)
Files: `macos/App/Features/Conclusions/PersonFamilySection.swift` (`FamilyPreviewData`), `macos/ProvenenciaTests/PersonFamilyFixtures.swift` (delete), every test that uses `FamilyFixtures`.
Do:
- Move the builders and the four families into one non-private `enum PersonFamilyPreviewData` in a new `macos/App/Features/Conclusions/PersonFamilyPreviewData.swift`, under `#if DEBUG` like `macos/App/Platform/PreviewFixtures.swift`.
- Keep the tests' `link:` option.
- Point the `#Preview`s and T-4 – T-13 at it through `@testable import`, as `PVDeleteImpactTests.swift:8` uses `PVDeleteImpactPreviewData`.
- Delete `PersonFamilyFixtures.swift`.

Check:
- `rg -n "enum FamilyPreviewData|enum FamilyFixtures" macos` prints nothing.
- `rg -n "enum PersonFamilyPreviewData" macos` prints exactly one line.
- Tests pass; `scripts/check-pr.sh` PASS.

**F-5 — One spoken-list join** (from MNT-2, same-pass)
Files: `macos/App/Platform/L10n.swift`, `PersonTile.swift`, `ReconciliationRecord.swift:58`, `ReconciledValueDisplay.swift:235`, `ReconciledValueRowModel.swift:259` (all in `macos/App/Features/Conclusions/`).
Failing test first: a `spokenJoinList` case in an existing L10n or design-system test file: `[]` → nil, `["a"]` → "a", `["a","b","c"]` → the two-part join folded.
Do: add `static func spokenJoin(_ parts: [String]) -> String?` beside the two-part `L10n.DesignSystem.spokenJoin`, and move the four callers onto it. Leave `AccessibilityAnnouncement` alone unless it holds the same fold.
Check:
- `rg -n "reduce\(parts\[0\]\)|reduce\(nil\).*spokenJoin|spokenJoin\(\\\$0" macos/App` prints only the new helper's own body, or nothing.
- Tests pass; `scripts/check-pr.sh` PASS.

**F-6 — One date-ordered sort in `conclusionsummaries`** (from MNT-3, MNT-7, same-pass)
Files: `core/database/conclusionsummaries/order.go`, `compose.go`.
Do:
- Add a private `sortByDateKey[T any](rows []T, key func(T) dateKey)`. Make `sortEvents` and `SortByBirth` call it.
- Give `lifePick` a `dateKey` built the way `eventKeyOf` builds one. Replace `lifeBefore` with `dateKeyLess` over that key. Refs are uppercase Crockford tokens, so the lowercasing changes no order.
- If a `conclusionsummaries` test's expected order changes, stop that part: keep `lifeBefore` and report it.

Check:
- `rg -n "sort.SliceStable" core/database/conclusionsummaries` prints only the helper's line.
- `rg -n "func lifeBefore" core` prints nothing.
- Go tests pass; `scripts/check-pr.sh` PASS.

**F-7 — Fix the "One detail key" guide line** (from MNT-4)
Files: `docs/macos-client-patterns.md:103`.
Do: narrow the sentence to the page's header and default sections. Say that a page section may read its own key, declared in `ConclusionPageSection.queryKey`, and keep "no per-entity summary key beside it".
Check: `rg -n "read only this key" docs` prints nothing; `scripts/check-pr.sh` PASS.

**F-8 — Place-only life line reads "in"** (from UX-3)
Files: `macos/App/Features/Conclusions/PersonLifeDisplay.swift`, `L10n.swift`, `Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`.
Failing test first: extend `lifeLinesOmitMissingHalves` so a death with a place and no date reads `"died in Hamburg"`.
Do: add `a11yLifeLinePlaceOnly(word, place:)` = "%1$@ in %2$@" for the place-only case. Keep `a11yLifeLineKnown` for the date-only case.
Check: the test passes; `scripts/check-pr.sh` PASS.

**F-9 — Test one parent in two multi-parent sets** (from CON-1)
Files: `core/database/persondetail/family_test.go`.
Do: add the subtest "two sets sharing a parent" to `TestPersonFamilyOriginSiblings`, with `w.k.Marriage(heinrich, kath)` added to the "two sets" world. Expect `Johann[Anna+Heinrich,Heinrich+Katharina]`, a sibling in both sets, a sibling in the second set only, then the first set's other siblings, in the order the code gives. If the code's order differs from frame 11c's reading, report it. Do not change the code to fit the test.
Check: passes; `scripts/check-pr.sh` PASS.

**F-10 — No trap on a duplicate tile id** (from SEC-2, CON-2)
Files: `PersonFamilySection.swift:102`.
Do: `Dictionary(content.allTiles.map { ($0.id, $0) }, uniquingKeysWith: { first, _ in first })`.
Check: `scripts/check-pr.sh` PASS.

**F-11 — Layout takes caption text and the tile's height** (from MNT-5, MNT-6)
Files: `PersonFamilyContent.swift`, `PersonFamilyLayout.swift`, the tests that read `siblingsCaption`.
Do:
- `siblingsCaption: String?` holds the text, built in the content.
- The layout places what it is given and has no L10n use left.
- Drop `PersonFamilyLayout.tileHeight` and read `PersonTile.height`.

Check: `rg -n "L10n" macos/App/Features/Conclusions/PersonFamilyLayout.swift` prints nothing; tests pass; `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `7ca95a03`. Gate PASS (orchestrator re-ran). F-1 – F-11 done. Deviations:
- **F-9:** with the Heinrich–Katharina marriage, kinship numbers Katharina's set first. The order is therefore `Johann, Friedrich, Margarethe, Wilhelm, Christian`, and the pre-F-2 code gives the same. CON-1 is **Resolved**: the shape is now tested. Whether that order reads as frame 11c intends goes to the user's manual check (frame 11c).
- **F-6:** the check's `rg` also hits `SortByTitle` and `placechains.go`, which are unrelated. The copied loops are gone.
- **Commits:** F-1, F-2 (Swift), F-3, F-10 and F-11 landed in one commit (`f411fcdf`).
- **F-1:** the partner fold counts the one-parent children group as a group. The more row is `.pv(.ghost, size: .sm)`.
- **Docs:** the Landed note records D-14 and the Edges premise as superseded, with the user's words.

Minors read by the orchestrator: F-9 (CON-1) Resolved, F-10 (SEC-2/CON-2) Resolved. F-8 and F-11 ride with the round-2 UX and MNT reviewers.

## Round 2 — review
Reviewers: MNT, UX, SEC+PRF (each had a Fix-now major). Not sent: CON (minors only). Scope `4cbf430a...7ca95a03`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | Return on a focused tile | Resolved | `PersonTile.swift:57-73` guarded `.onKeyPress(.return)`; the more row has the same handling |
| UX-3 | UX | minor | Place-only line | Resolved | `PersonLifeDisplay.swift:87`, `a11yLifeLinePlaceOnly` |
| MNT-1 – MNT-7 | MNT | major / minor | Fixtures, spoken join, date sort, guide line, caption, height, `lifeBefore` | Resolved | each `rg` check from Fix plan 1 holds |
| SEC-1 | SEC+PRF | major | Unbounded tiles; quadratic placement | Resolved | first paint bounded per list at 30; Swift and Go linear. A 200,000-case seeded probe showed the old and new Go placement give the same output |
| SEC-2 | SEC+PRF | minor | Duplicate-id trap | Resolved | `uniquingKeysWith` |

No new findings. No Fix plan 2.

## Result
Gate PASS (head `7ca95a03`; `main` unmoved at `018e5176`) · CI <pending> · open: none · follow-ups: none · to user: UX-2 (Tab order, manual check step 4); F-9 sibling order on frame 11c (manual check step 3) · CON miss: none
