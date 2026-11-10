# Review — S15-04 The Sources list shows and sorts by the date

Plan: [`S15-04-sources-list-date.md`](S15-04-sources-list-date.md). Brief: [`../design/S15-D2-sources-list.md`](../design/S15-D2-sources-list.md).

## Round 1 (2026-11-10)

Lint: `check-plan-docs.py` clean on both. Maintainability verdict: **Fits with changes**. Executability verdict: **Ready with changes** (board-dependent items excluded: the plan is Blocked on S15-D2).

### Maintainability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | The row date hand-builds icon + text beside `PVListDetailCell`, which already draws the line's last update and says the Sources row's second line draws the same cell | **Fixed.** D-5: the date is a `PVListDetailCell`; `PVListDetail` gains `truncates`; `SourcesListRow.dateDetail` builds it, shaped like `PVLastUpdated.detail`. Brief SL-1 and inventory now name the kit cell; the board no longer decides spacing or tone |
| M-2 | major | Accessibility base built by interpolation beside `L10n.DesignSystem.spokenJoin([String])` | **Fixed.** D-7: `SourcesListRow.spokenBase` uses `spokenJoin`; T-6 asserts it |
| M-3 | major | D-8 had the fake keep an old key or apply a cleared-means-nil rule, against `.cursor/rules/swift-tests.mdc` | **Fixed.** D-8: `FakeStore.nextSourceDateSortKey` scripts the key; T-3 scripts nil and a set key |
| M-4 | minor | Title-then-id tie spelled out twice | **Fixed.** D-3: one private `SourcesModel.titleOrder`; `.az`/`.za` untouched (Out: no refactor) |
| M-5 | minor | "Undated" belongs with the row date strings | **Fixed.** `L10n.Dates.rowUndated` (`dates.row.undated`) |
| M-6 | minor | "A later date filter" under Next adds names no map row or idea | **Fixed.** Dropped |

### Executability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | T-6 tested a helper on the `private` `SourcesLinkLabel` | **Fixed.** Internal statics `SourcesListRow.dateDetail` and `spokenBase` (D-5, D-7) |
| X-2 | major | A set date can give an empty `DateRowDisplay` line | **Fixed.** D-5: an empty line reads "Undated"; AC-1.2, T-6 case; a Stop condition was not needed |
| X-3 | major | Layout choices left open (priorities, spacing, identifier, help) | **Fixed.** D-5 (kit cell owns spacing, tone, help), D-6 (`.layoutPriority(0.5)` against the cell's 1), §6 macOS (no identifier: the link merges children) |
| X-4 | major | Fixtures underspecified | **Fixed.** STEP-3 extends the `source(...)` helper; T-1 – T-5 give literal ids, titles, keys and the phrase-only construction |
| X-5 | major | T-5's cost bound contradicted the Stop condition | **Fixed.** No timing assertion; T-5 asserts order and one build; the Stop condition is gone |
| X-6 | major | Swift checks need macOS | **Fixed.** DoD: report "not run (no macOS)", never mark done |
| X-7 | major | No test of the `mapSource` key gate | **Fixed.** D-1 makes `mapSource` internal; T-7 in `SourceMappingTests.swift` |
| X-8 | minor | Wide in scope but no code | **Fixed.** §6 macOS: no Wide code; a Stop condition if Wide does not fill |
| X-9 | minor | T-4 claimed AC-2.6 | **Fixed.** T-4 cites `unknownFieldValuesFallBack` |
| X-10 | minor | Version and archive steps vague | **Fixed.** STEP-5 names all four `MARKETING_VERSION` lines and the rebase rule; STEP-6 gives the archive row and `completed.md` shape |
| X-11 | minor | Older builds drop a stored date sort | No change: the existing per-field fallback is the intended behavior (AC-2.6) |
| X-12 | nit | Temptations not in Out | Logged |
| (brief) | minor | SL-5 listed `abt 1834` and `bet 1834 and 1840` as distinct places; `valuecompare.DateKeyOf` gives both the first day of 1834, side 0 | **Fixed.** SL-5 says they share a place and go by title |

**Round 2.** Not run yet: round 1 had blockers, all Fixed. Run round 2 when the board is agreed and D-6 takes its order, so both rounds' fixes and the board are checked together.

## Verdict

Fits with changes / Ready with changes. No open blocker; nothing to user. Status stays `Blocked — on board S15-D2`.
