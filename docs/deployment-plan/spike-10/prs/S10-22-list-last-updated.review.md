# S10-22 — review log

Plan: [`S10-22-list-last-updated.md`](S10-22-list-last-updated.md). Briefs: [S10-D9](../design/S10-D9-persons-list.md), [S10-D10](../design/S10-D10-events-list.md), [S10-D11](../design/S10-D11-places-list.md).

## Round 1 (2026-10-14)

Lint: `python3 scripts/check-plan-docs.py` — 0 errors on the plan and the three briefs (the briefs have one expected warning: D9's New inventory row names its second call site, S10-41).

Maintainability verdict: **Fits with changes.** Executability verdict: **Ready with changes.**

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | D-5 formatted in the system time zone; "yesterday" used the real clock, not `now` | **Fixed.** D-5 builds every style with `Date.FormatStyle(locale:calendar:timeZone: calendar.timeZone)` and names the day tests (`isDate(_:inSameDayAs:)` against `now` and `now − 1 day`). T-4 pins a Gregorian calendar in `America/Edmonton`. |
| X-2 | major | AC-1.3 locale forms vs. the brief's "12 Oct 2026" | **Fixed.** D-5 says the brief's example is the `en_GB` form and the locale decides. D-11 and brief PSL-1 size the column for the US form too. |
| X-3 | major | `detailWidth` default, two inits, argument order, nil help | **Fixed** with M-3: one `detail: PVListDetailColumn<Item>? = nil` on both inits, after `meta`; help attached only when non-nil; no `accessibilityHidden`. |
| X-4 | major | The em dash had no home | **Fixed** with M-1: `LastUpdatedDisplay.text` is `"—"` for unknown (D-6). |
| X-5 | major | The label wrap untested; overload name clashes | **Fixed** with M-1: `ConclusionListRow.labelWithUpdated(_:_:)` (no `accessibilityLabel` overload), T-6 tests it; the wrap in `ConclusionListBody` is one call, named in D-8. |
| X-6 | minor | Small choices: Resolve helper, invalidation shape, `touchedHandles` unchanged, field placement, skeleton sites, L10n placement | **Fixed.** T-1 calls `Resolve` inline; D-3 gives the shape and says `touchedHandles` does not change; D-4 places `updatedAt` last; D-8 names both skeleton sites via `ConclusionListPage.skeleton()`; Strings names the spot and comments. |
| X-cs | note | Precedent paragraph: `header_handles` comes from `HeaderDeps` after `Resolve` | **Fixed.** Precedent says so and limits the copy to proto, mapper, and invalidation. |
| M-1 | major | Dash, help, and spoken suffix lived in Conclusions; S10-41 needs them | **Fixed.** `LastUpdatedDisplay` *(new)* in `Platform/TimestampFormat.swift` carries text, help, spoken, and `columnWidth` (D-5, D-6, D-10, D-11). `ConclusionListRow` only maps it. T-5 moved to `TimestampFormatTests`. |
| M-2 | major | `PVList` comment and DesignSystem README say "no columns" | **Fixed.** STEP-6 edits both; DoD names the README. |
| M-3 | minor | Width defined twice; skeleton sites drift | **Fixed.** One width (`LastUpdatedDisplay.columnWidth`), one `PVListDetailColumn` parameter, one `ConclusionListPage.skeleton()`. |
| M-4 | minor | STEP-6 named `PVList` callers that do not exist | **Fixed.** Verified with `rg "PVList\("`: `ConclusionListPage.swift` is the only caller. Step and risk row reworded. |
| M-5 | minor | `scope_handles` reverses Go's `HandleScopes` | **Fixed.** `handle_scopes` / `handleScopes` at every layer; D-1 adds the `Set` field comment. |

Open after round 1: no blocker. To user: none beyond the plan's one open question (the boards).

## Round 2

Run because round 1 had a blocker (X-1). One fresh reviewer checked every round-1 answer against the plan and the code: all Fixed answers hold, and the names agree across D-5, D-8, D-10, D-11, §6, and T-4 to T-6.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-7 | minor | `L10n.LastUpdated` accessor signatures not declared | **Fixed.** §6 Strings lists each accessor's signature in the `a11yList` style. |
| X-8 | minor | "one parse per visible row" no longer true; cell and label each build the value | **Fixed.** §6 States says two builds per row and that a midnight mismatch for one render is accepted (D-9). |
| X-9 | nit | D-4 did not say `var` | **Fixed.** D-4 reads `var updatedAt: String = ""`. |

## Refresh — boards agreed (2026-10-15)

The three agreed boards were folded in: a Lucide `history` icon in the column (`PVSymbol.history`), a 136pt column, “Last updated …” help, a faint dash for unknown, a 64pt skeleton bar. One fresh dry-run reviewer checked the changed parts against the code: no blockers, verdict Ready with changes.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| R-1 | minor | D-5 field list missed `isKnown` | **Fixed.** D-5 lists it. |
| R-2 | minor | AC-1.5 did not say faint and icon like AC-1.4 | **Fixed.** AC-1.5 matches; T-5 checks `isKnown == false` for `"garbage"` too. |
| R-3 | minor | No mechanism kept the column from shrinking | **Fixed.** D-7: `.fixedSize()` on the text, `layoutPriority(1)` on the column. |
| R-4 | nit | STEP-5 said “width from the board” | **Fixed.** `columnWidth = 136`. |

## Verdict

**Fits with changes, Ready with changes — all applied.** No open blocker, nothing to the user. After the refresh the plan is `Planned`: boards agreed 2026-10-15, no open question.
