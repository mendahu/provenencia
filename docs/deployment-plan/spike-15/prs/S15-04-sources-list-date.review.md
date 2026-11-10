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

**Note after the board.** The agreed board (2026-11-10) changed the design under several round-1 answers: M-1, X-1, X-3 and X-8 describe a row date on the second line drawn by `PVListDetailCell` and "no Wide code". Round 2 below supersedes them; the plan's D-5 – D-10 are the current design. X-10's steps are now STEP-7 and STEP-8.

## Round 2 (2026-11-10, after the board)

Lint clean. Maintainability verdict: **Fits with changes**. Executability verdict: **Not ready** (two blockers), Ready with changes once fixed.

### Maintainability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M2-1 | blocker | Round-1 blocker M-1 reversed by D-6 without the user; the rejection reason was partly wrong (`PVListDetail.icon` is optional) and `PVDatePlaceLine` was not considered | **To user → answered** (user, 2026-11-10: draw the published date as the board says). D-6 and Codebase fit give the true trade-off (the cell lacks a label, the secondary tone and truncation; `PVDatePlaceLine`'s leading is an abbreviation and its missing wording "unknown"); promoted at S15-05. Stale round-1 answers noted above |
| M2-2 | blocker | `widthListWideMax` would mean content for Sources and the outer frame for Promote; T-8's Promote assertion cannot reach the `private` `PromoteLayout` | **To user → answered** (user, 2026-11-10: Promote adopts it). D-9: the token is content width; Promote frames its content at it before its padding, `listMaxWide` deleted, Regular unchanged, Wide 80pt wider. T-8 drops the Promote assertion; manual check 7 covers Promote |
| M2-3 | major | Wide centering hand-rolled a fourth time beside `pvKeepsPageBand` | **Fixed.** D-9: kit modifier `pvListBand(_:)` in `PVPageColumns.swift`; brief inventory row; § 9 names it |
| M2-4 | major | Columns cached in `@State` with the width class captured in the geometry closure | **Fixed.** D-7: measure only `listWidth`; build `columns` in `body` |
| M2-5 | major | "Published", "publication date", "Source date" and "Date" name one concept | **To user → answered** (user, 2026-11-10: "Published" everywhere). D-2 sort labels, D-14 Source page copy, T-9, AC-1.8; the S15-D1 and S15-D2 **Board** lines record the copy change |

### Executability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X2-1 | blocker | T-5's counts contradict (every multiple of 30 is even) | **Fixed.** 100 dated, all even; 1,400 undated; key 0 is present |
| X2-2 | blocker | STEP-4 changes the `SourcesListColumns` init but its call sites were in STEP-6 | **Fixed.** No default; STEP-4 updates `SourcesListView` (with M2-4's `listWidth`) |
| X2-3 | major | "Titles stay on one line" not carried into a Decision | **Fixed.** D-8: titles keep today's wrapping; Out line: the board's one line is the effect of the width |
| X2-4 | major | Wide container mechanics open; identity at the crossing | **Fixed.** D-9: one always-present structure, full-height hairlines in an overlay, modifiers order named |
| X2-5 | major | `SourcesLinkLabel` parameters and the empty-type rule unspecified | **Fixed.** D-8: `typeLabel` only at `.recordLine`, `date:`, `showsRecordLine:`; an empty type drops its separator |
| X2-6 | major | The board may be unreadable to the implementer | **Fixed.** Stop condition: §6 is the authority when the board cannot be read |
| X2-7 | minor | Key naming and comments | **Fixed.** camelCase keys beside their siblings; a comment for each in §6 Strings |
| X2-8 | minor | T-7 against `swift-tests.mdc` | **Fixed.** T-7 cites `ObservationDepictionMappingTests` as the allowed exception |
| X2-9 | minor | T-6 hedged; T-3 fields; locale | **Fixed.** `rowDate`'s guard returns `""` for a range missing a side, so T-6 expects nil; T-3 gives the fields; `SourcesRowDate` takes a locale and T-6 passes `en_US` |
| X2-10 | minor | One token, two meanings | **Fixed** with M2-2 |
| X2-11 | minor | Field order in the memberwise init | **Fixed.** D-1 |
| X2-12 | minor | Temptations | **Fixed.** Out lines for `.az`/`.za` and the search keys |

## Verdict

Fits with changes / Ready with changes after two rounds. Every blocker Fixed or answered by the user; nothing open. Status `Planned`.
