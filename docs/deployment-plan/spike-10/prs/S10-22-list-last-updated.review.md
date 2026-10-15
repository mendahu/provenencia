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

## Rework review — round 1 (2026-10-15)

The plan was amended for the user's rework (one timestamp formatter; surgical cache updates). Fresh maintainability and executability reviewers. Verdicts: Fits with changes · Not ready. The user answered M-5: move the Sources list onto the same patch and engine cache in this PR.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | `database.Commit` makes `graphcache` import `database` (cycle) | **Fixed.** `effects.Commit` (D-14). |
| X-2 | blocker | `applyTriggerEvicts…` breaks under stale-when-hidden too | **Fixed.** D-19 names both tests and the observable assertion. |
| X-3 | major | Promote test has People hidden | **Fixed.** Renamed outright (D-19). |
| X-4 | major | `en_US` stamp is month-first | **Fixed.** Verified with `Date.FormatStyle`; AC-1.1 and AC-3.4 now expect `Oct 12, 2026 · 9:00 AM MDT`. |
| X-5 | major | AC-1.1 named the deleted `boardStamp` | **Fixed.** |
| X-6 | major | `audit.Record` shape and ~22 callers | **Fixed.** Signature, empty-set meaning, `recordResolved` test helper (D-13). |
| X-7 | major | Parity test not named | **Fixed.** T-8 is `checkNewEffects` with its callers. |
| X-8 | major | Other `LatestByScope` / `audit.Latest` users | **Fixed.** `handleRevision` SQL, error and nil cases for `Graph.Activity`, `TestListHeadersCarryUpdated` (D-15, T-9). |
| X-9, M-6 | major | Docs and rules made false not listed | **Fixed.** STEP-15 lists each file and the rule it rewrites, then an `rg` sweep. |
| X-10 | minor | Seven call sites, not eight | **Fixed.** |
| X-11 | minor | `project.pbxproj` entries | **Fixed.** STEP-11 names the four entries and their source. |
| X-12 | minor | `Context.current`, AC-3.8 wording, existing symbols, fetch recorder | **Fixed.** `isFetching` is the observable; AC-3.8 lists the non-timestamp exceptions. |
| M-1 | blocker | Patch ordered by time string; drops revision; doesn't fit Sources | **Fixed.** `CatalogActivity {revision, at}` everywhere, ordered by revision (D-4, D-17). |
| M-2 | major | Third patch primitive with different in-flight rules | **Fixed.** `patchCached` (D-18); `patchUpdatedSource` moves onto it; T-12 covers in-flight. |
| M-3 | major | A fourth per-kind `switch` | **Fixed.** `CatalogLastUpdatedPatch.rows(Row.self, scope:)` declared on the spec (D-17). |
| M-4 | major | Two Booleans on one axis | **Fixed.** `CatalogHiddenPolicy` (D-19). Other hidden keys: Out, follow-up idea. |
| M-5 | major | Sources keeps the reload mechanism | **To user → do it here.** D-20; mutation-path Sources writes keep `markStaleOn` (Out, follow-up). |
| M-7 | major | Next adds planned a copy of the cell | **Fixed.** D-21: S10-41 moves it to a shared home. |
| M-8 | minor | Stale sections | **Fixed.** §6, Strings, STEP-3 title, risks updated or marked superseded. |
| M-9 | minor | Keep `Graph.Activity` off `persons()` | **Fixed.** D-15. |
| M-10 | nit | `.shortDate` / `.longDate` have no caller | **Declined.** The user named short and long date as part of the API. |

Lint after fixes: 0 errors.

## Rework review — round 2 (2026-10-15)

One fresh reviewer (executability + maintainability) checked the round-1 answers; they hold. New findings, all fixed in the plan:

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-13 | blocker | STEP-12 deleted symbols `LastUpdatedDisplay` still used until STEP-13 | **Fixed.** STEP-12 is additive; STEP-13 deletes `LastUpdatedDisplay`, `boardStamp`, and the two keys. |
| X-14 | blocker | STEP-11 renamed the header field without the three Kind files | **Fixed.** Added to STEP-11. |
| M-11 | major | `patchCached` copied `setQueryValue`'s in-flight block | **Fixed.** Private `replaceValue(_:_:clearStale:)` shared by both (D-18). |
| X-15 | minor | `Source.updated_at` mapped before it exists | **Fixed.** STEP-10 maps revision only; `handlers/sources.go` and the early return named. |
| X-16 | minor | A third test uses `evictsWhenHidden` | **Fixed.** `hiddenPolicies` (D-19). |
| X-17 | minor | `patchUpdatedSource`'s stale clear | **Fixed.** Kept after the workspace patch (D-18). |
| M-12 | nit | Leftovers | **Fixed.** §1, Codebase fit, DoD, `Notify` parameter name, `lastUpdatedPatch(_ key:)`. |

## Rework verdict

Two rounds; no open blocker; the user answered M-5. The plan returns to the implementer.
