# S11-04 — Value compare: dates

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 2 — Value compare, everywhere
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a Go comparer, moved date helpers, docs)
**Board:** —
**Depends on:** S11-02 (landed [#441](https://github.com/mendahu/provenencia/pull/441))
**Blocks:** S11-05
**Branch:** `s11-04-value-compare-dates`
**Version bump:** none — no caller uses the new comparer yet; the moved date helpers and knobs behave exactly as before ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [layer 1](../comparison-layers.md#layer-1-value-comparison), [relations](../comparison-layers.md#relations) and the `ABT` rule, [detail](../comparison-layers.md#detail), [layer 1 tests](../comparison-layers.md#tests), [reuse: date windows](../comparison-layers.md#reuse-across-the-app); [`docs/matching-model.md`](../../../matching-model.md) knob inventory `value.Dates.*`; [`docs/matching.md`](../../../matching.md) `DateComparer` row; [S11-02 plan](S11-02-value-compare-core.md) D-8, D-10, D-12, D-15; working rules "Build, then migrate", "High tolerance for churn", "Registry numbers: carry over, then estimate", "Value compare returns a relation, a similarity and a detail", "A normalized `Key` per value type", "One match registry"
**Planned:** 2026-10-17 against `main` @ `458c20f` (S11-02 landed)
**Review:** [`S11-04-value-compare-dates.review.md`](S11-04-value-compare-dates.review.md) — Fits with changes / Ready with changes → fixed, 2 rounds

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

Value compare learns dates. `valuecompare.Compare` relates two date values as spans of days and returns `Same`, `Contains`, `Overlaps`, `Near`, `Different` or `Unknown`, a similarity, and a detail with the broader side and the gap in days. `ABT` reads as a range of one unit of its most precise part, which replaces "ABT doubles the tolerance". The date ends and the window algebra (`DateEnds`, `Window`, overlap, containment, intersection) move from `core/autoreconcile` into value compare, and the calendar arithmetic they rest on moves into `core/datevalue`. The date knobs join the registry's value section.

**Why now.** S11-02 landed the contract and the registry. S11-05 moves every date caller onto this comparer, and S11-06 moves place chains onto day windows; both need the comparer and the windows in the layer first.

**What stays.** Nothing changes for researchers. `core/match`'s `DateComparer` keeps scoring Promote until S11-05, reading the registry's numbers, so every score is today's. The reconciler's own date fold (`dateModule`, `DateBounds`) and the place-chain helpers built on it (`WindowOfDate`, `PeriodWindow`, `LinkMembership`) keep their behavior and stay in `core/autoreconcile` until S11-06.

## 2. User stories

Numbers use the registry's starting values: `Tolerance` 2 years, `MissingDay` 0.9, `MissingMonth` 0.8, `OtherDay` 0.7, `OtherMonth` 0.6, `YearsApartFrom` 0.8, `SpanOneYear` 0.8, `SpanPerExtraYear` 0.05, `SpanFloor` 0.2, `AboutUnits` 1, `NearMinimum` 0. "Span score" is max(0.8 − 0.05 × (w − 1), 0.2) for the wider closed span of w years (w at least 1), and 0.2 when a span is open (D-6).

**US-1.** As a contributor building a matching layer, I compare two dates and read whether one contains, overlaps or is near the other, how close, and by how many days, with no database.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `1817` / `1817`; `BEF 1820` / `BEF 1820`; `ABT 1817` / `ABT 1817`; `BET 1817 AND 1817` / `1817`; `BET 1820 AND 1815` / `BET 1815 AND 1820` | `Same`, 1, cause `exact` |
| AC-1.2 | `May 1817` / `14 May 1817`; `1817` / `Sep 1817` | `Contains`, broader A, cause `precision`: 0.9; 0.8 |
| AC-1.3 | `ABT 1817` / `1818`; `ABT 1817` / `1817` | `Contains`, broader A, `precision`, span score of 3 years = 0.8 − 0.05 × 2 = 0.7 |
| AC-1.4 | `BEF 1820` / `BEF 1830`; `BEF 1820` / `1815` | `Contains`, `precision`, 0.2: broader B; broader A |
| AC-1.5 | `BET 1815 AND 1820` / `1818` | `Contains`, broader A, 0.8 − 0.05 × 5 = 0.55 |
| AC-1.6 | `ABT MAR 1817` / `APR 1817`; `ABT 3 MAR 1817` / `4 MAR 1817`; `ABT JAN 1817` / `15 DEC 1816`; `ABT 1 JAN 1817` / `31 DEC 1816` | `Contains`, broader A, 0.8 each (the last two `ABT` ranges cross a year end) |
| AC-1.7 | `BET 1815 AND 1820` / `BET 1819 AND 1825` | `Overlaps`, cause `overlap`, 0.8 − 0.05 × 6 = 0.5 |
| AC-1.8 | `14 MAY 1817` / `20 MAY 1817`; `MAY 1817` / `SEP 1817` | `Near`, cause `gap`: 0.7, `Distance` 6 days; 0.6, `Distance` 93 days |
| AC-1.9 | `1817` / `1818`; `DEC 1817` / `JAN 1818` | `Near`, `gap`: 0.8 × (1 − 1/3) ≈ 0.533, `Distance` 1; 0.8 × (1 − (31/365) / 3) ≈ 0.777, `Distance` 1 |
| AC-1.10 | `AFT 1820` / `1819`; `BET 1815 AND 1820` / `1822`; `ABT 1817` / `ABT 1820` | `Near`, `gap`: 0.2 × (1 − 1/3) ≈ 0.133; 0.55 × (1 − 2/3) ≈ 0.183; 0.7 × (1 − 1/3) ≈ 0.467 |
| AC-1.11 | `1817` / `1820`; `BEF 1820` / `1823`; `BEF 1800` / `AFT 1810` | `Different`, 0, cause `beyond_tolerance`; `Distance` 731, 731 and 3288 days |
| AC-1.12 | A phrase with no year (`spring`); a range with no year on either side; no date | `Unknown`, 0: cause `phrase_only`; `empty`; `empty` |
| AC-1.13 | Year 1817 with day 14 and no month / year 1817 with day 20 and no month | `Same`, 1 (a day with no month reads as its year, as `DateEnds` does today) |
| AC-1.14 | Tolerance 0 on a copy of `Default().Value`: `1817` / `1818` | `Different`, `beyond_tolerance` |

**US-2.** As a contributor, the date comparer keeps the layer 1 contract, and a date key agrees with `Compare`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The date sample table and 500 generated pairs per seed | The S11-02 contract bounds, swap (`Broader` flipped) and self rules (`TestComparerContract`) |
| AC-2.2 | The same pairs through `Key` | Equal at `Same` exactly when `Same`; equal at `Equivalent` exactly when `Same` (dates have no `Equivalent`) (`TestKeyContract`) |
| AC-2.3 | `Key` of `1817` and of `BET JAN 1817 AND DEC 1817` | Equal |

**US-3.** As a contributor, I find the date windows and the calendar in one place each.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | STEP-1's first `rg` | Nothing; `core/datevalue/calendar.go` (new) holds `DaysFromCivil`, `DaysInMonth`, `BoundDays`, `YearLine`, `AddMonths`, `AddDays` |
| AC-3.2 | STEP-2's two `rg` commands | Nothing: no caller reaches the moved code through `core/autoreconcile`, and `HoldKind`, `Relate`, `HoldsAt` exist nowhere |
| AC-3.3 | `go test ./core/matchregistry/...` | `TestKnobInventoryMatchesRegistry` passes with no `value.Dates.` prefix in `notYetBuilt` |

**US-4.** As a researcher, nothing behaves differently.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | The whole Go and Xcode suite | `scripts/check-pr.sh` prints `check-pr: PASS`; no expected value changes in `core/match`, `core/autoreconcile`, `core/database/conclusionheaders`, `core/database/autoreconciler` or `core/database/promotealign` tests |

## 3. Scope

| In | Covered by |
| --- | --- |
| Calendar arithmetic moves to `core/datevalue/calendar.go` (new) (D-2) | STEP-1, AC-3.1 |
| `DateEnd`, `DateEnds` and the window algebra move to `core/valuecompare/datewindow.go` (new); every caller repointed; `HoldKind`, `Relate`, `HoldsAt` deleted (D-3) | STEP-2, AC-3.2 |
| `AboutEnds` and `EndsWindow`: a date as compared, and its window (D-6) | STEP-4, AC-1.3, AC-1.6 |
| The date knobs in `Rules` and the registry; `match` reads them (D-4, D-5) | STEP-3, AC-3.3 |
| The date comparer, `ABT` as a range, and its `Key` (D-1, D-6 – D-9) | STEP-4, AC-1.1 – AC-1.14, AC-2.3 |
| Contract rows and generator for dates | STEP-4, AC-2.1, AC-2.2 |
| Today's date ladder and invariants, ported onto relations | STEP-5 |
| The matching and date-value skills; docs that name the moved code | STEP-6, STEP-8 |

| Out | Home |
| --- | --- |
| Pooling `date` with `start_date` / `end_date` (`match.poolPeriods`) | S11-10 |
| Moving any caller onto the date comparer (`match.DateComparer`, `graphalign`, `promotealign`) and deleting `core/match/dates.go` | S11-05 |
| The reconciler's date fold and grouping (`dateModule`, `dateWindow`, `DateBounds`), and place chains on day windows (`WindowOfDate`, `PeriodWindow`, `LinkMembership`, `conclusionheaders` `laterStart` / `earlierEnd`) | S11-06 |
| `ApproxToleranceFactor` | S11-05 deletes it with `match.DateComparer` (already under "Removed knobs") |
| Time zones and clock fields in date comparison | not planned: `DateEnds` reads civil days, as today |
| `Parts` for dates (from and to ends as parts) | S11-19 adds them if the evidence wording needs them; relation, cause, `Broader` and `Distance` word every case today ([wording](../claim-evidence.md#wording)) |

## 4. Current state

Verified at `458c20f`.

| Area | Where | Today |
| --- | --- | --- |
| Date ends | `core/autoreconcile/dates.go` `DateEnd`, `DateEnds`, `boundEnd` | Ends as written, with day numbers and a line of years; lives in the reconciler, which a layer cannot import |
| Window algebra | `core/autoreconcile/dates.go` `Window`, `Always`, `Overlaps`, `Intersect`, `Contains`, `HoldKind`, `HoldMiss`, `HoldYes`, `HoldAmbiguous`, `Relate`, `HoldsAt`, `laterBound`, `earlierBound`, `copyBound` | Pure day-span algebra in the reconciler. `HoldsAt` is exactly `Overlaps` (`Relate` is `HoldMiss` only when the windows do not overlap); `Relate` and `HoldKind` have no production reader, and `HoldsAt` one (`placechains.go` `linkHolds`) |
| Reconciler date windows | `core/autoreconcile/dates.go` `dateModule`, `dateWindow`, `contains`, `moreSpecificThan`, `openEnds`, `windowOf`, `pointWindow`, `rangeWindow`, `spanOf`, `DateBounds`, `WindowOfDate`, `PeriodWindow`, `LinkMembership` | The reconciler's fold window: `ABT` is its point, a range with a missing year is not placed |
| Calendar | `core/autoreconcile/dates.go` `spanOf` (precision 1 / 2 / 3, then clock fields), `lastDay`, `isLeap`, `daysFromCivil`, `dateSortOffset` | Calendar facts in the reconciler, written with literals |
| Today's date comparer | `core/match/dates.go` `DateComparer`, `dateRules`, `resolve`, `dateSpan`, `spanOf`, `centre`, `width`, `shares`, `spanScore`, `apart`, `Compare` | Bare 0…1 score; `ABT` on a point doubles the tolerance (`ApproxToleranceFactor`) |
| Date knobs today | `core/match/registry.go` `DefaultDates` (`Tolerance` 2, `ApproxToleranceFactor` 2, `MissingDay` 0.9, `MissingMonth` 0.8, `OtherDay` 0.7, `OtherMonth` 0.6, `YearsApartFrom` 0.8, `SpanOneYear` 0.8, `SpanPerExtraYear` 0.05, `SpanFloor` 0.2) | Literals in `match` |
| Callers of the moved code | `core/match/dates.go` L68, L73, L84 (`DateEnd`, `DateEnds`); `core/match/periods.go` L72, L75, L80, L107 (`DateEnd`, `DateEnds`); `core/database/conclusionheaders/placechains.go` (`Window` ×5, `Always` ×3, `HoldsAt`; `WindowOfDate`, `LinkMembership`, `PeriodWindow` stay); `core/database/conclusionheaders/compose.go` (`PeriodWindow` ×3, stays); `core/database/autoreconciler/autoreconciler.go` L519 (`DateBounds`, stays) | Import `core/autoreconcile` for them |
| Value compare | `core/valuecompare/valuecompare.go` `Value` (no date field), `comparers`, `Unit` (`UnitCount` only), `falloff`; `core/valuecompare/rules.go` `Rules{Words, Text, Integers}` | No date comparer |
| Registry | `core/matchregistry/value.go` `defaultValue`; `core/matchregistry/inventory_test.go` `notYetBuilt` (`value.Dates.`) | Dates not built |
| Inventory | `docs/matching-model.md` rows `value.Dates.Tolerance` through `value.Dates.AboutUnits`; "Removed knobs" lists `DefaultDates.ApproxToleranceFactor` | Rows wait for this PR |
| Tests | `core/autoreconcile/dates_test.go` `TestDateEnds` (its `year` subtest also calls `DateBounds`), `TestDateBounds`, `TestPeriodWindowAndLinkMembership` (calls `Overlaps`, `Intersect`, `HoldsAt`, `Relate`, `HoldAmbiguous`, `Always` unqualified, L334 – L375), `TestDateFold`, `TestDateReconcile`; `core/match/dates_test.go` `TestDateComparerScores`, `TestDateComparerDayWithoutMonth`, `TestDateComparerABT`, `TestDateComparerLadder`, `TestDateComparerInvariants`; `core/valuecompare/contract_test.go` `samples`, `generators`; `core/valuecompare/valuecompare_test.go` `TestCompareUnsupported` (expects a date value `unsupported`) | — |
| Docs naming moved code | `docs/deployment-plan/spike-11/comparison-layers.md` L454, L460; `docs/matching.md` L61 (`DateComparer` row) | Name `core/autoreconcile` as the home |
| Date-value skill | `.agents/skills/add-date-value/SKILL.md` | Names no date window or calendar code; D-11 adds both |

**Precedent.** S11-02 ([plan](S11-02-value-compare-core.md), [#441](https://github.com/mendahu/provenencia/pull/441)): helpers moved into `valuecompare` with callers repointed in the same commit (D-8), `match` reading the registry instead of its literals (D-10), the shared `falloff` curve (D-12), and calendar facts kept out of the numbers guard by living in a pure value package (D-15). Copy each.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Date values.** `Value` gains `Date *datevalue.Value`. `comparers[propertymeta.ValueTypeDate] = date{}` in `core/valuecompare/date.go` (new) | One entry point (S11-02 D-3) | — |
| D-2 | **Calendar facts move to `core/datevalue/calendar.go` (new)**, outside the numbers guard (S11-02 D-15): `DaysFromCivil(year, month, day int) int` (today's `daysFromCivil`, 1970-01-01 is 0), `DaysInMonth(year, month int) int` (today's `lastDay` with `isLeap`), `BoundDays(year, month, day *int) (lo, hi int, ok bool)` (the inclusive day span one bound covers at its precision: a nil year is not ok; no month → the whole year, a day ignored; a month failing the package's existing `monthOK` or a day outside 1 – `DaysInMonth` → not ok), `YearLine(lo, hi int) (from, to, mid float64)` (the span on a line of years where day d of year y covers [y + (d−1)/n, y + d/n) with n the days in y — today's `boundEnd` arithmetic — and `mid` its centre, so `valuecompare` needs no literal 2), `AddMonths(year, month, n int) (int, int)`, `AddDays(year, month, day, n int) (int, int, int)` (through an unexported `civilFromDays`, the inverse of `DaysFromCivil`, Howard Hinnant's `civil_from_days`). `autoreconcile.spanOf` keeps its precision count and takes `lo`, `hi`, `ok` from `BoundDays`; `daysFromCivil`, `lastDay`, `isLeap` are deleted from `core/autoreconcile`; `dateSortOffset` stays | The date comparer needs month and day arithmetic (`ABT MAR` is Feb–Apr) and cannot hold the literals; one calendar for the reconciler and value compare | Calendar code in `valuecompare` — because the numbers guard forbids 12, 31 and 4800 there and a calendar fact is not a tunable |
| D-3 | **Window algebra and date ends move to `core/valuecompare/datewindow.go` (new).** `DateEnd` gains `Mid float64` (the centre on the line of years); `DateEnds`, `boundEnd` move unchanged except that `boundEnd` takes `Lo`, `Hi` from `datevalue.BoundDays` and `From`, `To`, `Mid` from `datevalue.YearLine`. `Window` and `Always` move unchanged. `Overlaps`, `Intersect`, `Contains` become methods on `Window`, because `valuecompare.Contains` and `valuecompare.Overlaps` are relation constants: `(a Window) Overlaps(b Window) bool`, `(a Window) Intersect(b Window) (Window, bool)`, `(outer Window) Contains(inner Window) bool`; `laterBound`, `earlierBound`, `copyBound` move unexported. `HoldKind`, `HoldMiss`, `HoldYes`, `HoldAmbiguous`, `Relate` and `HoldsAt` are deleted: `HoldsAt` is exactly `Overlaps`, and `placechains.go` `linkHolds` calls `query.Overlaps(mem)`. `DateBounds`, `WindowOfDate`, `PeriodWindow`, `LinkMembership` stay in `core/autoreconcile` (they read the reconciler's `windowOf` rule, which S11-06 replaces) and return `valuecompare.Window`, calling its methods. Callers in §4 are repointed in the same commit. `TestPeriodWindowAndLinkMembership` calls the methods, and its `Relate(...) != HoldAmbiguous` checks read `q.Overlaps(m) && !m.Contains(q)`, with no expected outcome changed. `TestDateEnds` moves to `core/valuecompare/datewindow_test.go` (new), `package valuecompare_test`; its `year` subtest compares with `datevalue.BoundDays` instead of `DateBounds` (a layer test may not import `core/autoreconcile`) | The map scope entry; place chains and the reconciler use the layer's windows (reuse inventory). Methods avoid the name clash; one relation vocabulary (`Relation`) in layer 1, so no `HoldKind` beside it | Renaming to `WindowsOverlap` / `WindowContains` functions — because methods need no new names; moving `HoldsAt` and `Relate` — because they say what `Overlaps` and `Contains` already say; moving `WindowOfDate` and its kin now — because their `ABT` and missing-year rules are the reconciler's and change in S11-06 |
| D-4 | **Knobs** (`Rules.Dates`, `DateRules`, inventory keys and start values): `Tolerance int` 2, `NearMinimum float64` 0, `MissingDay` 0.9, `MissingMonth` 0.8, `OtherDay` 0.7, `OtherMonth` 0.6, `YearsApartFrom` 0.8, `SpanOneYear` 0.8, `SpanPerExtraYear` 0.05, `SpanFloor` 0.2 (`float64` each), `AboutUnits int` 1. `core/matchregistry/value.go` sets each with a comment naming its source (`DefaultDates.*`; `graphalign` `ValueTypeScale["date"].Floor` for `NearMinimum`; the design's `ABT` rule for `AboutUnits`). `notYetBuilt` loses `value.Dates.` | The inventory rows S11-01 wrote; carry over | — |
| D-5 | **`core/match` keeps no second copy** (S11-02 D-10): `DefaultDates` reads `value.Dates.Tolerance`, `MissingDay`, `MissingMonth`, `OtherDay`, `OtherMonth`, `YearsApartFrom`, `SpanOneYear`, `SpanPerExtraYear`, `SpanFloor`; `ApproxToleranceFactor` keeps its literal 2 (a removed knob, deleted with `match.DateComparer` in S11-05). Every `match` expected value stays | One number, one place | — |
| D-6 | **Reading a date: one exported step.** `datewindow.go` gains `func AboutEnds(d datevalue.Value, units int) (first, last DateEnd, ok bool)` — "a date as compared": `DateEnds`, then an `ABT` point widens by `units` of its most precise part on each side: year precision (no month; a day with no month reads as its year) → the first end is the bound of year − n, the last the bound of year + n; month precision → `AddMonths` by −n and +n; day precision → `AddDays` by −n and +n; each widened end is built by `boundEnd` on its civil components. `DateEnds` stays "as written" (`match` still reads `ABT` as a point until S11-05). And `func EndsWindow(first, last DateEnd) Window`: `Lo` = `first.Lo`, `Hi` = `last.Hi`, nil where an end is `Open`. The comparer calls `AboutEnds(d, r.Dates.AboutUnits)` then `EndsWindow`; S11-06 and S11-12 call the same two. A date is a **plain point** when its kind is not `range` and it has no qualifier (the test helpers leave `Kind` empty, as `match`'s do). The **later** and **earlier** of two disjoint windows: the later is the one whose first end is closed and lies after the other's closed last end. Two windows **share a day** when `a.Overlaps(b)`. A **span score** is `SpanFloor` when either window is open, else max(`SpanOneYear` − `SpanPerExtraYear` × (w − 1), `SpanFloor`) where w is the wider of the two spans' `last.To − first.From`, at least 1 | The [`ABT` rule](../comparison-layers.md#relations); the span score is today's `spanScore` | Keeping `ApproxToleranceFactor` — because the design replaces it |
| D-7 | **The date rule**, in order. (1) Either date nil, or `AboutEnds` not ok → `Unknown`; cause `phrase_only` when a's date (else b's) is the one that failed and has a `Phrase` and no year on either bound, else `empty`. (2) Equal windows (`a.Contains(b) && b.Contains(a)`) → `Same`, 1, `exact`. (3) Windows share a day: when one covers the other (`Contains`) → `Contains`, `Broader` the covering side, cause `precision`; else `Overlaps`, cause `overlap`. Similarity: both plain points → `MissingDay` when the broader names a month and no day, `MissingMonth` when it names neither; otherwise the span score. (4) Disjoint: `years` = the later window's first end `Year` − the earlier window's last end `Year`; above `Tolerance` → `Different`, `beyond_tolerance`. (5) Otherwise `Near`, cause `gap`: both plain points in one civil year → `OtherDay` when both name the same month, else `OtherMonth`; both plain points in different years → `YearsApartFrom` × `falloff(gap, Tolerance)`; else the span score × `falloff(gap, Tolerance)`, where gap = the later window's first end `Mid` − the earlier window's last end `Mid`, in years. (6) A `Near` below `NearMinimum` → `Different`, 0, `below_minimum`. **Detail:** for (4) – (6) `Distance` = the later window's first `Lo` − the earlier window's last `Hi` (days) and `Unit` = `UnitDays`; for `Same`, `Contains`, `Overlaps` and `Unknown`, `Distance` is 0 and `Unit` empty; `Parts` is empty | Every date cell of the relations table: same span, one inside the other, overlapping, apart within tolerance, beyond it. The tiers and the fall-off are today's numbers, regrouped: a missing day or month is containment, a different day or month is a gap. `falloff` is S11-02's shared curve | Similarity from `falloff` over days for every case — because it would drop today's month and day tiers, which the inventory carries over |
| D-8 | **Date `Key`.** At `Same` and `Equivalent`: the window as `lo..hi`, each a decimal day number (`strconv.Itoa`), `*` for an open end (`BEF 1820` is `*..-54422`). An unknown date has no key | Equal windows are exactly D-7 step 2; dates have no `Equivalent` | — |
| D-9 | **New codes.** `Unit` gains `UnitDays = "days"`; `Cause` gains `CauseBelowMinimum = "below_minimum"` with the doc comment "a pair close enough to be Near but under its type's minimum, so Different", which fits both uses. S11-03 declares `CauseBelowMinimum` with the same comment: whichever lands second keeps the one declaration on rebase. **The contract's self rule** in `contract_test.go` ("a value against itself") accepts `Same`, or `Unknown` with cause `empty` or `phrase_only` | S11-02 D-5 reserved `UnitDays` for this PR | — |
| D-10 | **Tests.** `core/match/dates_test.go` stays as it is (it guards live code until S11-05). The new comparer's tests are in `core/valuecompare/date_test.go` (new), `package valuecompare_test`, with one set of builders in `date_test.go` — `ymd(y, m, d)`, `about(y, m, d)`, `bound(qualifier, y)`, `between(y1, m1, d1, y2, m2, d2)`, `phrase(s)` (each returns a `datevalue.Value`), `noYearRange()` (a range with no year on either side), and `ofDate(a, b *datevalue.Value)` wrapping two of them into `valuecompare.Value`s for sample pairs, a nil pointer meaning no date (nil month or day as `0`) — used by `datewindow_test.go` and `contract_test.go`, with no second copy. A shared `core/datevalue/datevaluetest` package is logged for S11-05 / S11-06, which delete `match`'s and `autoreconcile`'s date tests | One copy in this PR | — |
| D-11 | **Docs and skills.** `add-match-rule` gains: dates compute similarity through `falloff`; calendar facts live in `core/datevalue/calendar.go` (new); the window algebra is `valuecompare.Window`. `add-date-value` names `core/datevalue/calendar.go` (new) and `core/valuecompare/datewindow.go` (new). `comparison-layers.md` L454 reads: "The date window primitives (`Window` with `Overlaps`, `Contains`, `Intersect`) moved into value compare in S11-04, with `AboutEnds` and `EndsWindow`; `DateBounds`, `WindowOfDate`, `PeriodWindow` and `LinkMembership` follow in S11-06, so place chains and reconciliation use the same day-precision rules as matching." L460's cell gains "(`DateEnds` and windows moved in S11-04)" after `DateEnds` and windows; the row is not split. `matching.md` L61 gains one sentence: "S11-04's `valuecompare` date comparer replaces this in S11-05; it reads `ABT` as a range." The cause list gains `below_minimum` (once, shared with S11-03) | The working rule on docs a migration makes false; the map's S11-04 and S11-06 scope entries already say what moves when | — |
| D-12 | Commits, each passing the full suite: (1) calendar; (2) windows and ends move; (3) knobs and `match` reads them; (4) the date comparer with its contract rows; (5) ladder and invariants; (6) skills and docs | One concern per commit | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/datevalue/calendar.go` *(new)* | file | `core/autoreconcile/dates.go` `daysFromCivil`, `lastDay`, `isLeap`, `spanOf` | The calendar **moves** here (D-2); `YearLine` is `boundEnd`'s arithmetic; `AddMonths`, `AddDays` are new and only `ABT` widening needs them |
| `core/valuecompare/datewindow.go` *(new)* | file | `core/autoreconcile/dates.go` window algebra, `DateEnds` | **Moves** here (D-3) |
| `Window.Overlaps`, `Intersect`, `Contains` *(methods)* | funcs | the `autoreconcile` functions of the same names | The same code as methods (D-3); `HoldsAt`, `Relate`, `HoldKind` are deleted, not moved |
| `AboutEnds`, `EndsWindow` *(new)* | funcs | `DateEnds`; `autoreconcile.WindowOfDate` | One exported reading of a date as compared, and the one ends-to-window conversion (D-6) |
| `DateEnd.Mid`, `datevalue.YearLine` *(new)* | field, func | `match` `centre` | The centre without a literal in the layer (D-2, D-3) |
| `core/valuecompare/date.go` *(new)*: `date` comparer | file, type | `core/match/dates.go` `DateComparer` | Returns a bare score and doubles the tolerance for `ABT`; build, then migrate: S11-05 deletes it. `date.go` uses `Window.Overlaps` / `Contains` for sharing and covering, `DateEnd.Mid` for the gap, and keeps only the span width, the span score and the year gap; S11-05 deletes `match`'s `centre`, `shares`, `spanScore`, `apart` |
| `valuecompare.DateRules` *(new)* | type | `match.DateComparer` pointer settings | Whole values passed in (S11-02 D-1) |
| `UnitDays`, `CauseBelowMinimum` *(new)* | codes | `UnitCount`, cause list | Extends the code sets (D-9) |
| `core/valuecompare/date_test.go`, `datewindow_test.go` *(new)* | test files | `core/match/dates_test.go`, `core/autoreconcile/dates_test.go` `TestDateEnds` | New comparer's tests; `TestDateEnds` moves |
| `core/datevalue/calendar_test.go` *(new)* | test file | `core/autoreconcile/dates_test.go` `TestDateBounds` | Tests the moved calendar directly |

Searches run for this table: `rg -n 'func (DaysFromCivil|YearLine|AddMonths|AddDays|DaysInMonth)' core` (none), `rg -n 'daysFromCivil|lastDay' core --type go` (only `core/autoreconcile/dates.go`), `rg -n 'type Window' core` (only `core/autoreconcile`).

**Builds on.** S11-02's contract, `comparers`, `falloff`, registry and guards; S11-22's pure `core/datevalue`; the `add-match-rule` skill.
**Deletes.** `daysFromCivil`, `lastDay`, `isLeap`, `HoldKind`, `HoldMiss`, `HoldYes`, `HoldAmbiguous`, `Relate`, `HoldsAt` and the window algebra and date ends from `core/autoreconcile/dates.go`; `TestDateEnds` from `core/autoreconcile/dates_test.go`; the literal values of nine `DefaultDates` knobs; the `value.Dates.` prefix. `match.DateComparer` becomes dead once S11-05 moves its callers, and S11-05 deletes it.
**Next adds.** S11-05 builds date values and deletes `core/match/dates.go`. S11-06 moves `DateBounds`, `WindowOfDate`, `PeriodWindow` and `LinkMembership` onto `AboutEnds` and `EndsWindow` and makes place chains fold by day. S11-10 pools `date` with `start_date` / `end_date` as a range of this comparer. S11-12 builds lifespans from `AboutEnds` windows.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/datevalue` | `calendar.go` *(new)* (D-2) |
| `core/autoreconcile` | `dates.go`: calendar and moved code gone; `spanOf` calls `datevalue.BoundDays`; `DateBounds`, `WindowOfDate`, `PeriodWindow`, `LinkMembership` return and use `valuecompare.Window` (D-3) |
| `core/valuecompare` | `datewindow.go`, `date.go` *(new)*; `valuecompare.go`: `Value.Date`, the `date` entry, `UnitDays`, `CauseBelowMinimum`; `rules.go`: `DateRules`, `Rules.Dates` |
| `core/matchregistry` | `value.go`: the D-4 values; `inventory_test.go`: `value.Dates.` gone |
| `core/match` | `dates.go`, `periods.go` call `valuecompare.DateEnd` / `DateEnds`; `registry.go` reads `value.Dates` (D-5) |
| `core/database/conclusionheaders` | `placechains.go`: `valuecompare.Window`, `valuecompare.Always`, `query.Overlaps(mem)` |

```go
type DateRules struct {
	// Tolerance is the most civil years apart two dates may be and still be Near.
	Tolerance int
	// NearMinimum is the least similarity of a Near date; below it the pair is Different.
	NearMinimum float64
	// MissingDay is a point inside a month that names no day (May 1817 / 14 May 1817).
	MissingDay float64
	// MissingMonth is a point inside a year that names no month (1817 / Sep 1817).
	MissingMonth float64
	// OtherDay is two points in one month on different days.
	OtherDay float64
	// OtherMonth is two points in one year in different months.
	OtherMonth float64
	// YearsApartFrom is where two points in different years start their fall-off.
	YearsApartFrom float64
	// SpanOneYear is a range, bound or ABT on either side spanning a year or less.
	SpanOneYear float64
	// SpanPerExtraYear is taken off SpanOneYear for each further year of span.
	SpanPerExtraYear float64
	// SpanFloor is the least span score, and the score of an open span (BEF, AFT).
	SpanFloor float64
	// AboutUnits is how many units of its most precise part ABT widens each side.
	AboutUnits int
}
```

Errors: none. Every input has a result.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — The calendar moves to `datevalue`**
Files: `core/datevalue/calendar.go` (new), `core/datevalue/calendar_test.go` (new), `core/autoreconcile/dates.go`. Skill: `add-date-value`.
Do: D-2. T-1.
Check: `go test ./core/datevalue/...` passes with no tags; `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/... ./core/database/conclusionheaders/... ./core/database/autoreconciler/...` passes with no expected value changed; `rg -n 'func (daysFromCivil|lastDay|isLeap)\b' core/autoreconcile` prints nothing; `boundEnd` (still in `autoreconcile` until STEP-2) calls `datevalue.BoundDays` and `datevalue.YearLine` (AC-3.1).

**STEP-2 — Date ends and windows move to value compare**
Files: `core/valuecompare/datewindow.go` (new), `core/valuecompare/datewindow_test.go` (new), `core/valuecompare/date_test.go` (new: only D-10's builders in this step), `core/autoreconcile/dates.go`, `core/autoreconcile/dates_test.go`, `core/match/dates.go`, `core/match/periods.go`, `core/database/conclusionheaders/placechains.go`.
Do: D-3, and D-10's builders in `date_test.go`. Move, do not copy. T-2.
Check: `rg -n 'autoreconcile\.(DateEnds?|Window\b|Always|Overlaps|Intersect|Contains)' core api` prints nothing and `rg -n 'HoldKind|HoldsAt|HoldMiss|HoldYes|HoldAmbiguous|\bRelate\(' core api` prints nothing (AC-3.2); `go test ./core/valuecompare/...` passes with no tags; `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/... ./core/match/... ./core/database/conclusionheaders/... ./core/database/autoreconciler/... ./core/database/promotealign/...` passes with no expected value changed; `go test -run TestLayerPackages ./core/matchregistry/` passes.

**STEP-3 — Date knobs in the registry; `match` reads them**
Files: `core/valuecompare/rules.go`, `core/matchregistry/value.go`, `core/matchregistry/inventory_test.go`, `core/match/registry.go`. Skill: `add-match-rule` "Add a knob".
Do: D-4, D-5.
Check: `go test ./core/matchregistry/...` passes; `CGO_ENABLED=1 go test -tags fts5 ./core/match/... ./core/database/promotealign/...` passes with no expected value changed.

**STEP-4 — The date comparer**
Files: `core/valuecompare/date.go` (new), `core/valuecompare/valuecompare.go`, `core/valuecompare/date_test.go` (new), `core/valuecompare/valuecompare_test.go`.
Files also: `core/valuecompare/contract_test.go`, `core/valuecompare/datewindow_test.go` (created in STEP-2) (new).
Do: D-1, D-6 – D-9. In `TestCompareUnsupported`, `ComparerTypes()` gains `date` and the date value leaves the unsupported list. In the same commit, because `TestComparerContract` fails on a comparer with no row: a `propertymeta.ValueTypeDate` row in `samples` holding every AC-1.x pair, and a generator — each side is, by one draw: a point (four in ten), an `ABT` point, a `BEF` point, an `AFT` point (one in ten each), a range (two in ten) or a phrase-only date `spring` (one in ten); every bound has a year 1815 – 1821, a month (none, or 1 – 12, even odds) and, when it has a month, a day (none, or 1 – 28, even odds); a range draws its two bounds and orders them so start ≤ end; the contract runs under `Default()` rules (date tolerance is varied only in T-7); the self rule accepts `phrase_only` (D-9). T-3, T-4, T-5, and `AboutEnds` and `EndsWindow` cases in T-2.
Check: `go test ./core/valuecompare/...` passes; `TestDateCompare` has one named subtest per AC-1.x pair (`rg -c 'AC-1\.' core/valuecompare/date_test.go` counts at least 14).

**STEP-5 — Ladder and invariants**
Files: `core/valuecompare/date_test.go` (new).
Do: T-6, T-7, using STEP-4's generator.
Check: `go test -run 'TestDateLadder|TestDateInvariants' -count=1 ./core/valuecompare/` passes.

**STEP-6 — Skills and docs**
Files: `.agents/skills/add-match-rule/SKILL.md`, `.agents/skills/add-date-value/SKILL.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching.md`.
Do: D-11.
Check: every path the skills name exists (`ls`); `rg -n 'AboutEnds' docs/deployment-plan/spike-11/comparison-layers.md` prints the L454 line; `rg -n 'calendar.go' .agents/skills/add-date-value/SKILL.md .agents/skills/add-match-rule/SKILL.md` prints a line in each.

**STEP-7 — Full suite**
Files: none.
Do: run the gate on the user's Mac, as `ship-pr` requires (`xcode-test` cannot pass elsewhere).
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-11/deployment-plan.md`, `docs/deployment-plan/spike-11/completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-04-value-compare-dates.md` reports no errors.

Commits: one per behavior (D-12). The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Compare dates as spans of days in value compare."). No WIP commits.

## 8. Tests

All new tests are table-driven (`.cursor/rules/go-tests.mdc`) with no build tags. Expected values are written as the arithmetic.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/datevalue/calendar_test.go` `TestCalendar` *(new)* | `DaysFromCivil(1970, 1, 1)` → 0; `AddDays(y, m, d, 1)` then `AddDays(y2, m2, d2, -1)` returns every day of 1600 – 2100 unchanged; `DaysInMonth(1900, 2)` → 28, `(2000, 2)` → 29; `BoundDays` of 1985 → 1 Jan – 31 Dec, of May 1985 → 1 – 31 May, of 1985 with day 14 and no month → the year, of month 13 or 30 Feb → not ok; `AddMonths(1817, 1, -1)` → 1816, 12; `AddDays(1817, 1, 1, -1)` → 1816, 12, 31 | AC-3.1, D-2 |
| T-2 | Go | `core/valuecompare/datewindow_test.go` `TestDateEnds` (moved), `TestWindow` *(new)* | `TestDateEnds`'s cases unchanged (its `year` subtest compares with `datevalue.BoundDays`; `Mid` is `(From + To) / 2` within 1e-12); `AboutEnds` of `ABT 1817` (1 units) → 1816 – 1818, of `ABT MAR 1817` → 1 Feb – 30 Apr, of `ABT JAN 1817` → 1 Dec 1816 – 28 Feb 1817, of `ABT 1 JAN 1817` → 31 Dec 1816 – 2 Jan 1817, of `1817` → the year; `EndsWindow` of `BEF 1820` → nil low, high 31 Dec 1820; `TestWindow`: two disjoint windows do not overlap, an open window overlaps every window, `Intersect` of overlapping windows, `Contains` with open ends, `Contains` of a window by itself | AC-3.2, D-3 |
| T-3 | Go | `core/valuecompare/date_test.go` `TestDateCompare` *(new)* | Every AC-1.1 – AC-1.14 pair → relation, similarity (±1e-9), cause, `Broader`, `Distance`, `Unit` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, AC-1.12, AC-1.13, AC-1.14, D-6, D-7 |
| T-4 | Go | `core/valuecompare/date_test.go` `TestDateKey` *(new)* | AC-2.3; `BEF 1820` has an open low end (`*..`); a phrase → `"", false`; level `Near` → `"", false` | AC-2.3, D-8 |
| T-5 | Go | `core/valuecompare/contract_test.go` `TestComparerContract`, `TestKeyContract` (date rows) | The date samples and generated pairs → the S11-02 contract | AC-2.1, AC-2.2 |
| T-6 | Go | `core/valuecompare/date_test.go` `TestDateLadder` *(new)* | Against `14 MAY 1817`, strictly decreasing: `14 MAY 1817` (1, `Same`) > `MAY 1817` (0.9, `Contains`) > `1817` (0.8, `Contains`) > `20 MAY 1817` (0.7, `Near`) > `SEP 1817` (0.6, `Near`) > `BET 1815 AND 1820` (0.55, `Contains`) > `1818` (0.8 × (1 − (1.5 − 133.5/365) / 3), `Near`) > `1819` (0.8 × (1 − (2.5 − 133.5/365) / 3), `Near`) > `BEF 1820` (0.2, `Contains`) > `1820` (0, `Different`) | D-7 |
| T-7 | Go | `core/valuecompare/date_test.go` `TestDateInvariants` *(new)* | The generator of STEP-4 over seeds 1 – 4, 2000 pairs per seed, `Dates.Tolerance` 0 – 3 (the seed index) → every pair that is `Contains` at `AboutUnits` 1 with the broader side an `ABT` date stays `Contains` with the same `Broader` at `AboutUnits` 2 (similarity may drop); raising `Tolerance` never turns `Near` into `Different`; a plain point is never broader than the same point under `ABT` | D-6, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/autoreconcile` dates tests (STEP-1 and STEP-2 move code out), `core/database/conclusionheaders` place chains (STEP-2 repoints), `core/match` dates and periods tests, `TestLayerPackages` (`valuecompare` imports `core/datevalue`, pure).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: matching.md L61, comparison-layers.md L454, L460 and the cause list (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

No app change. The reviewer runs, from the repo root:

1. `go test ./core/valuecompare/... ./core/matchregistry/... ./core/datevalue/...` with no tags: passes.
2. `go list -deps ./core/valuecompare | grep 'provenencia/core/database'`: no output.
3. AC-3.1's and AC-3.2's `rg` commands: no output; `rg -n 'value.Dates' core/matchregistry/inventory_test.go`: no output (AC-3.3).
4. Open a project, open a Place with dated parents, and open Promote on a Source with dated events: the place chain and the suggestions read as before (AC-4.1).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `BoundDays` validates a month or day differently from today's `spanOf`, so a reconciler or place-chain result moves | T-1 pins the edges; STEP-1 runs the reconciler and place-chain suites unchanged |
| Methods on `Window` change a nil-bound edge in place chains | The method bodies are the moved function bodies; STEP-2 runs `conclusionheaders` unchanged |
| An `ABT` range crossing a year makes a contract swap or self case fail | T-5 and AC-1.6's year-crossing cases |
| S11-03 lands first and edits the same files (`valuecompare.go`, `rules.go`, `value.go`, `inventory_test.go`, `contract_test.go`, `matching-model.md`, the skill) | Each edit is additive; on rebase keep both sides, and one `CauseBelowMinimum` (D-9) |

**Stop and report if:**

- A row in §4 is no longer true at your base (for example `DateEnds` has moved or `conclusionheaders` no longer calls `autoreconcile.HoldsAt`).
- STEP-1, STEP-2 or STEP-3 changes any expected value in an existing test.
- An AC-1.x pair cannot hold under D-7's rule as written.
- `TestLayerPackages` fails for `core/datevalue` after STEP-1 (the calendar must stay pure), or `core/valuecompare/numbers` fails and the fix would be anything other than a registry field or a `datevalue` calendar function.
- `valuecompare` needs to import anything under `core/database`, `core/match` or `core/autoreconcile`.

## 12. Open questions

None.

## 13. PR

**Title:** S11-04 — Value compare: dates
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence (test names); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-04-value-compare-dates.md`.
