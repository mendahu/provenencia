# Review — S11-04 Value compare: dates

**Verdict:** round 1 — Fits with changes / Not ready → every finding fixed; round 2 pending
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- None.

## Writer's own calls

- Calendar arithmetic moves to `core/datevalue/calendar.go`, because the numbers guard keeps calendar literals out of the layer (D-2).
- `HoldKind`, `Relate` and `HoldsAt` are deleted, not moved: `HoldsAt` is exactly `Overlaps`. The map's S11-04 scope entry changed with it (D-3).
- `DateBounds`, `WindowOfDate`, `PeriodWindow`, `LinkMembership` stay in `autoreconcile` until S11-06; the map's S11-06 scope entry says so (D-3).
- Today's tiers are regrouped: a missing day or month is `Contains`, a different day or month is `Near` (D-7).
- `ABT 1817` / `1817` drops from 1 to `Contains` 0.7: the design's `ABT` rule makes ABT a 3-year span (AC-1.3).

## Round 1

### M-1 · blocker · `ABT` widening private beside `DateEnds`
**Answer:** Fixed. D-6: exported `AboutEnds(d, units)` in `datewindow.go`; `DateEnds` stays as written for `match`. T-2 covers the widening, year and month crossings included.

### M-2 · blocker · `HoldKind` / `Relate` / `HoldsAt` duplicate `Overlaps`
**Answer:** Fixed. D-3 deletes them; `linkHolds` calls `query.Overlaps(mem)`; `TestPeriodWindowAndLinkMembership` reads `Overlaps` and `!Contains`. The map line and `comparison-layers.md` L454 (D-11) drop `HoldsAt`. A technical call, so it is listed under the writer's calls rather than To user.

### M-3 · major · Reuse the window methods through one conversion
**Answer:** Fixed. D-6 adds `EndsWindow`; D-7 uses `Overlaps`, `Contains` and mutual `Contains` for equality; D-8 formats the same window; Codebase fit row rewritten.

### M-4 · major · One set of date builders
**Answer:** Fixed. D-10 names one set in `date_test.go`; `datevaluetest` is logged for S11-05 / S11-06.

### M-5 · minor · `BoundDays` uses `monthOK`
**Answer:** Fixed (D-2).

### M-6 · minor · `CivilFromDays` exported with one caller
**Answer:** Fixed. Unexported `civilFromDays`; T-1 goes through `AddDays`.

### M-7 · minor · One doc comment per knob
**Answer:** Fixed (§6 sketch).

### X-1 · blocker · STEP-4 breaks the contract test
**Answer:** Fixed. The contract row and generator are in STEP-4 and its commit (D-12).

### X-2 · blocker · Self rule against `phrase_only`
**Answer:** Fixed. D-9: the self rule accepts `Unknown` with `empty` or `phrase_only`.

### X-3 · blocker · `centre` needs a literal 2
**Answer:** Fixed. D-2 `datevalue.YearLine` returns `mid`; D-3 adds `DateEnd.Mid`; D-7 uses it.

### X-4 · major · Plain point and `Kind ""`
**Answer:** Fixed. D-6: plain is kind not `range` and no qualifier.

### X-5 · major · `Distance` and `Unit` for the other relations
**Answer:** Fixed (D-7 Detail).

### X-6 · major · `rg` checks vacuous
**Answer:** Fixed. The commands live in STEP-1 and STEP-2 unescaped; the ACs point at them; D-3 names the `TestPeriodWindowAndLinkMembership` edits.

### X-7 · major · `add-date-value` row false
**Answer:** Fixed. Own row: the skill names no window or calendar code; D-11 says the L460 edit and that the row is not split.

### X-8 · major · One wording for `CauseBelowMinimum`
**Answer:** Fixed. D-9 gives the shared doc comment; S11-03 uses the same.

### X-9 · minor · Smaller choices
**Answer:** Fixed: the generator's shape, the builders, `Default()` rules in the contract, a day with no month reads as its year, later / earlier windows, the cause when a fails first, `civilFromDays`' algorithm, STEP-4's and STEP-6's checks, the D-8 example (`*..-54422`), AC-1.6's wording, and 2000 pairs per seed in T-7.
