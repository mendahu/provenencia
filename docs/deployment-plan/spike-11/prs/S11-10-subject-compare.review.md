# Review — S11-10 Subject compare

Plan: [`S11-10-subject-compare.md`](S11-10-subject-compare.md). Lint (`scripts/check-plan-docs.py`) clean before round 1, except the review log link.

**Verdict:** Fits with changes — Ready with changes (round 2), all fixed; no open blocker.
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **The registry check is a seed check, not a key-list check (M-2).** The map's S11-10 entry said: port `TestDefaultProfileCardinality` as "every registry `PropertyKey` is a `vocabkey` constant". The plan instead ports it in place, in `core/database/subjectvocab`: every kind is a seeded type, every key is bound to that kind, and pooled and period properties have the right value type. This is stronger, and S11-14's bridge kinds pass it unchanged. The map's scope entry was reworded in the same commit. Recommendation: keep it.

## Writer's own calls

- An agree earns the full agree weight, even for an `Equivalent` (D-5). A partial earns agree weight × similarity.
- Each side decides alone whether to build a period. The rule is symmetric; `match.poolPeriods` was not (D-7).
- The period runs from the earliest start to the latest end, picked by day number (D-8).
- Value type and cardinality reach the layer as `Metas` from the caller, never from the registry (D-3).
- `comparison-layers.md`'s stale 0.81 becomes 0.71, the value under the decided "conflict weighs its penalty" rule.

## Round 1

Maintainability: Fits with changes. Executability: Not ready.

### M-1 · blocker · Lifespan (S11-12) gets no slot in this shape
**Answer:** Fixed. D-13 makes lifespan a derived lead. Its weight rows become `subject.Weights.person.lifespan.Agree` / `.Conflict` now, covered by `notYetBuilt`. The window rows stay for S11-12. Next adds says S11-12 adds a `Weights` entry, a meta and a builder value, with no second scoring path.

### M-2 · major · `TestSubjectSection` hard-codes today's reach and drops half the ported test
**Answer:** Fixed. D-14 splits the check. Structure is `TestSubjectSection` in `matchregistry`; seed is `TestSubjectSectionMatchesSeed` in `subjectvocab`, which checks types, bindings and value types. `vocabkey.PropertyKeys` is dropped as no longer needed. To user (above), because the map named the vocabkey form.

### M-3 · major · A second private copy of layer-2 value builders and an unguarded metas fixture
**Answer:** Fixed. D-11 and STEP-3 add `core/propertycompare/propertycomparetest` and move S11-09's helpers onto it. `core/subjectcompare/subjectcomparetest.Metas` is pinned to the seed by AC-4.2.

### M-4 · major · `period` / `endFields` copy `match`'s construction
**Answer:** Fixed. D-8 adds `valuecompare.EndsDate`, the inverse of `DateEnds` (STEP-1, AC-4.4, T-8). `pairs.go` only picks the ends.

### M-5 · minor · Typed keys belong where every section can use them
**Answer:** Fixed. D-2 declares `SubjectType` and `PropertyKey` in `core/vocabkey`.

### M-6 · minor · Record why pools and periods are separate maps
**Answer:** Fixed. D-6's Rejected cell records it: fields on `Weights` would demand ten empty inventory rows.

### M-7 · minor · Skill trigger text does not name layer 3 work
**Answer:** Fixed. STEP-8 adds "a Subject property weight, pool or period" to the description.

### M-8 · nit · The per-layer test command is spelled out in three places
Logged for S11-11's planner (plan §3 Out).

### X-1 · blocker · `subject.Pools.person.sex_at_birth` is not a key `registryKeys` produces
**Answer:** Fixed. D-6 makes the element `Pool{With []PropertyKey}`, so the key is `subject.Pools.person.sex_at_birth.With`. A scratch run of `registryKeys`' walk over the §6 types printed exactly D-13's thirteen keys.

### X-2 · major · STEP-4's check needs pooling, which STEP-5 builds
**Answer:** Fixed. STEP-5 holds US-1 rows except AC-1.2, and STEP-6 adds AC-1.2.

### X-3 · major · `matching-model.md` keeps stale keys in §Removed knobs
**Answer:** Fixed. STEP-4 lists each renamed key in §Removed knobs. Its check `rg` must print only the five lifespan window rows.

### X-4 · minor · Row-rename text miscounted and unclear
**Answer:** Fixed. STEP-4 now lists every full key, gives the cell text for the new rows, and quotes the sentence the intro insertion follows.

### X-5 · minor · `check-plan-docs` fails on the missing log
**Answer:** Fixed by this file.

### X-6 · minor · The `Pools` freshness check cannot fail with `append`
**Answer:** Fixed. STEP-4 sets element `With[0]`.

### X-7 · Choices the plan does not make
**Answer:** Fixed:
- `TestCompare` row names and fields are in D-11.
- A lead with both a pool and a period is forbidden (D-7, AC-4.3).
- A "positive" value is defined as neither negative nor neutral (D-7).
- The period value's wrapping is in D-8.
- The key-order call is in D-4.
- The string-to-typed-key conversions are in STEP-4.
- AC-3.6 says what "exactly" means.
- The `completed.md` text and the Landed note are in STEP-8.

### X-8 · Assumptions about S11-09 it does not state
**Answer:** Fixed:
- §4's layer 2 row cites S11-09 D-4, D-8 and D-11 for the zero-`Meta` and speaking-value behavior.
- A stop condition covers the zero `Meta` (AC-1.14) and the eight helpers STEP-3 moves.
- AC-2.5's speaking value is the only one on its side, so pair identity is not at stake.

## Round 2

Maintainability: Fits with changes. Executability: Ready with changes. The round-1 answers were checked and hold.

### M-9 · blocker · The seed check reads `seedBindings`, which leaves out the bridge bindings S11-14 will weigh
**Answer:** Fixed. D-14, AC-4.1, T-7 and §4 now check against `subjectvocab.allBindings()`, which is `seedBindings` plus the product `connectrules.BridgeBindings()`, the set Install materializes. Verified at `core/database/subjectvocab/subjectvocab.go` `allBindings`.

### M-10 · blocker · `subjectcompare` re-derives layer 2's "positive value" rule
**Answer:** Fixed. D-18 adds `propertycompare.Value.Positive()`. Layer 2's checks call it with no change in behavior, and so do D-7 and D-8. It is in STEP-3 and has a stop condition.

### Minor (noted by maintainability) · the recipe's "`TestVocabKeysMatchSeed` ties it to the seed"
**Answer:** Fixed. STEP-8's recipe says to add the key to that test's list.

### X-9 · minor · STEP-3's PASS-count baseline
**Answer:** Fixed. STEP-3's Do now records the count first.

### X-10 · minor · Generator choices
**Answer:** Fixed. D-15 now settles four things:
- draws are with replacement;
- the label format is fixed;
- AC-3.4 adds a value that is neither negative nor neutral;
- AC-3.5 compares with a 1e-9 slack.

### X-11 · nit · The header Review line was stale
**Answer:** Fixed.

## After review

- **Refreshed against `main` @ `847b5f0` (S11-09 landed in #468).** Its code matches the plan's assumptions. `helpers_test.go` has ten helpers (`labels`, `dates` added), so `propertycomparetest` exports ten. The fixture `var` block stays in a new `fixtures_test.go`.
- **Scope added at the user's request (2026-10-18):** `core/valuecompare`'s private date builders move onto `datevaluetest`. This was S11-09 review finding MNT-3, logged only in its ship log. It is new STEP-1, D-19 and T-9; the later steps are renumbered (the round-1 and round-2 answers above use the old numbers).
- **Writer's call on `optInt`:** it is called directly in about ten inline date literals, not only by the builders. Deleting it would mean keeping a private copy, so `datevaluetest` exports it as `Part` and gains `Between`. `bound` and `noYearRange` stay private.
