# Review — S11-12 Lifespan for people

**Verdict:** Fits with changes — Not ready → all findings fixed
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

None. The two product calls were asked and answered before the plan was written (below).

## User's calls (2026-10-18, during planning)

- **Lifespan shape.** The user asked for "a really straightforward" rule: take every event the person is connected to, first and last, and pad the start when there is no birth and the end when there is no death or burial. The four age-rule knobs (`ParentMinYears`, `ParentMaxYears`, `MotherMaxYears`, `MarriageMinYears`) are dropped. Written to the map's working rule, its S11-12 scope entry and `comparison-layers.md` §Lifespan with this plan.
- **Comparison.** Span in window, over window against window. The writer first said windows alone would catch resided-after-death, then corrected that: they catch only fully disjoint lives. The user chose span in window, which keeps the map's three Checks and adds died 1900 / resided 1950.
- **Neighbors.** Participations only, every event type and role ("or even just all of them").

## Writer's own calls

- A side whose own events fall wholly outside its own window has no lifespan (D-12, X-1), the same as a birth after a death. A coarser same-year event is not a contradiction (AC-2.18).
- Consistent lifespans score as agree (0.1) whatever partial relation spoke (D-5), so the worked example keeps its 0.1.
- `Near` (within the 2-year date tolerance) is consistent, not a conflict (D-2).
- The derived key's date meta is fixed in `subjectcompare` (D-5), not passed in `Metas`.
- `Options.Excluding` does not change the lifespan (Out, weak-claim review).

## Round 1

### X-1 · blocker · AC-4.3 contradicts D-12
A side such as birth 1900 with residence 1850 built a window that did not contain its own span, so the invariant would fail.
**Answer:** Fixed. D-12 refuses a lifespan when a date's late end finishes before the window starts or its early end begins after it ends. New AC-2.16, AC-2.17, and AC-2.18 (a coarser same-year event is kept). The invariant is now AC-4.7: the window overlaps every event date.

### X-2 · blocker · The plan fails its own plan-doc check
**Answer:** Fixed. This log exists, and the **Review:** line links it.

### X-3 · major · Test and subtest IDs collide with existing ones
**Answer:** Fixed. D-13: every new row starts with its AC ID, and the IDs continue the existing tables' numbers: US-3 is AC-3.10 – AC-3.14 and US-4 is AC-4.5 – AC-4.7. AC-1.10 names the existing "both sides empty" row.

### X-4 · major · STEP-6 leaves stale lifespan text in `comparison-layers.md`
**Answer:** Fixed with this plan's commit. The intro sentence, the two rule bullets and the "Bounded by construction" bullet now state the 2026-10-18 rule. STEP-6 only adds the "Built in S11-12" paragraph.

### X-5 · minor · A Current state row is false (eight calls)
**Answer:** Fixed. It now says six.

### X-6 · minor · Choices left open in the test steps
**Answer:** Fixed. D-13 types the `TestEntityCompare` fields (`life, subjectLife []subjectprofile.LifeEvent`, with `e.Life = tc.life`). US-3 names the name property and the subject's values. D-10 deletes `group.same`. `yr` is gone (M-10). The inventory rows are in a fenced block.

### X-7 · nit · Docs wording the plan does not cover
**Answer:** Fixed. STEP-6 says the scope entry, including its "Also builds" paragraph, stays as it is.

### M-1 · blocker · `EventRole` is a second, untyped key that S11-14 will duplicate
**Answer:** Fixed. D-3 adds `vocabkey.EventType` and `vocabkey.Role` with typed constants. `EventRole` and `LifeEvent` use them. Next adds names S11-14's `BridgeKey` as their next user, and Codebase fit and Out cite `EdgeSignature` and the `hops.go` literal.

### M-2 · major · Two ways to pick the earliest / latest date end
**Answer:** Fixed. D-6 adds `valuecompare.Earliest` and `Latest`; `period` (D-7) and the builder (D-12) both use them. `pickEnd`, `StartsBefore` and `EndsAfter` are gone.

### M-3 · major · `ShiftYears` puts calendar arithmetic in `valuecompare`
**Answer:** Fixed. D-6 adds `datevalue.AddYears`, and `ShiftYears` calls it and `boundEnd`. The leap-day cases are in `TestAddYears`.

### M-4 · major · Term constants are checked against the seed in the wrong place
**Answer:** Fixed. D-3 has `TestVocabKeysMatchSeed` check the term constants. D-9's seed change only skips the derived lead and checks the kind. STEP-6 adds the skill line for derived keys.

### M-5 · major · `Rules.Lifespan` breaks the `Pools` / `Periods` shape
**Answer:** Fixed. D-4 makes it `Rules.Lifespans map[SubjectType]map[PropertyKey]Lifespan{Span, MaxYears, Starts, Ends}`. `Compare` checks `rules.Lifespans[kind][key]`. The inventory has four `subject.Lifespans.person.lifespan.*` rows and no `Kind` or `Window` rows. The builder takes `map[PropertyKey]Lifespan` (D-11).

### M-6 · minor · `ForSubject`'s positional arguments drift from `Entity`
**Answer:** Fixed. D-11 adds `subjectprofile.Subject{Values, NamePattern, Life}`, and `ForSubject(s Subject, lifespans)` takes it.

### M-7 · minor · `sameValues` beside `propertyPools`
**Answer:** Fixed. D-10 reads each group's test value as `propertyPools[lead][0]`.

### M-8 · minor · Agree kept on a partial relation is undocumented
**Answer:** Fixed. The doc comment in D-5 and the skill's Layer 3 sentence in STEP-6 both explain it.

### M-9 · minor · Docs this change makes false
**Answer:** Fixed. STEP-2 updates the `subjectcompare.Values` doc and the `vocabkey` package doc. STEP-6 updates the skill's `Metas` sentence.

### M-10 · nit · `yr` forwards to an existing helper
**Answer:** Fixed. `ev(…)` is called with `datevaluetest.YMD(y, 0, 0)` directly.
