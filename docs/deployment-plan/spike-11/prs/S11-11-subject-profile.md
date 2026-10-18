# S11-11 — Subject profile and entity compare

**Status:** Blocked — on S10-66
**Spike / slice:** Spike 11 · Slice 3 — Property, Subject and entity compare
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a new pure Go package with no caller yet; nothing reaches the app)
**Board:** —
**Depends on:** S11-10 (landed [#471](https://github.com/mendahu/provenencia/pull/471); [plan](S11-10-subject-compare.md) and its **Landed** note: it shipped as planned, with `Compare` complete in one commit). S10-66 ([plan](../../spike-10-followups/prs/S10-66-pure-depiction-value.md), Planned): it moves the depiction value type into pure `core/depictionvalue`, which makes `core/autoreconcile` pure. This PR imports `core/autoreconcile` from a layer package, so it cannot pass the purity guard before S10-66 lands (D-2). Branch from `main` after S10-66 lands
**Blocks:** S11-12
**Branch:** `s11-11-subject-profile`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) ("Add a layer package")
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Entities: compare against the conclusion](../comparison-layers.md#entities-compare-against-the-conclusion) (the profile table, "the profile is strict", "the profile builder is its own package"); [lifespan](../comparison-layers.md#lifespan-a-derived-property-for-people) ("for a person Subject or person entity"); [reuse across the app](../comparison-layers.md#reuse-across-the-app) ("Leave-one-out profiles", "Profiles on both sides"); [Layer 2 contract](../comparison-layers.md#layer-2-property-comparison) ("No reconciler context … Layer 4 translates them before comparing"); [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) (reasons; a claim's value is laid over the cache, never written into it); [`docs/matching-model.md`](../../../matching-model.md) stack table row 4, glossary **profile**; working rules "Entity compare is against the reconciled conclusion", "Pure means the package builds and tests with no build tags and no `core/database` import", "Build, then migrate", "One guard, not one per package", "One matching skill, grown with the layers"
**Planned:** 2026-10-18 against `main` @ `bfe09bd` (S11-10 merged); S10-66 names from its plan
**Review:** [`S11-11-subject-profile.review.md`](S11-11-subject-profile.review.md) — Fits with changes / Ready with changes → fixed, 1 round

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

Layer 4 of matching exists. `subjectprofile.Build` *(new)* turns an entity's auto-reconciled values into a Subject-shaped profile (`subjectcompare.Values`), so a Subject is compared with the entity by `subjectcompare.Compare` against that profile. The profile is strict: displayed values are positive, values a stronger negative Observation denied are negative, and every other value is left out. `Options.Excluding` builds the profile as if some members' Observations were not in the group, by running the reconciler again without them. `subjectprofile.ForSubject` *(new)* builds the other side, a plain Subject's own values, through the same value conversion, so both sides of entity compare (and of S11-12's lifespan) come from one package.

**Why now.** S11-12 derives a lifespan inside the profile builder for both sides, and S11-13 moves Promote's candidate scoring onto entity compare. "Build, then migrate" puts the builder in its own PR with its own tests before any caller moves.

**What stays.** `core/match` (`Value.SetEvidence`, `pairValues`, `weigh`, the leaders and support weights) and `core/database/matching` are untouched; S11-13 moves Promote onto this layer and deletes them. `core/autoreconcile` and `core/subjectcompare`'s non-test code are unchanged. No catalog, wire, Swift or researcher-visible change.

## 2. User stories

Conventions for every AC. Members are `M1` … `M4` (member Subject ids `{0,2,0,n}`), Sources `S1` … `S3` (`{0,1,0,n}`), Observations numbered in the order written (`{0,0,0,n}`). "male (M1, S1)" is a candidate on member M1, cited under Source S1, positive, with default provenance. "weak" is `Provenance.Credibility` −1; "provisional" is `Candidate.Provisional`; "negative" is `Candidate.Negative`. Unless an AC says otherwise, a property's `Values` are `autoreconcile.Reconcile(matchregistry.Default().Value, …).Values` over its candidates with the property's `subjectcomparetest.Metas()` value type and cardinality and no concluded value, and its `Candidates` are those candidates. "Excluding [M1]" is `Options{Excluding: [][]byte{M1}}`. Profile values are written as `subjectcomparetest.Labels` writes them; "¬x" is a negative value. "sim" and "ev" are `subjectcompare.Result.Similarity` and `.Evidence` (within 0.0001) under `matchregistry.Default()` and `subjectcomparetest.Metas()`. Reconciler outcomes and every sim / ev below were read from the code at `bfe09bd`.

**US-1.** As the engine (Promote S11-13, merge hints later), when I build an entity's profile, I get its displayed values as positives, its denied values as negatives, and nothing else.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `sex_at_birth`: male (M1, S1), male (M2, S2), female (M3, S3) | Reconciler: male kept, female outvoted. Profile `sex_at_birth=male` |
| AC-1.2 | `sex_at_birth`: female weak (M1, S1), female negative (M2, S2) | Reconciler: female denied. Profile `sex_at_birth=¬female` |
| AC-1.3 | `date`: 1817 (M1, S1), 1830 (M2, S2) | Reconciler: both kept (mixed). Profile `date=1817,1830` |
| AC-1.4 | `toponym` (multiple): York (M1, S1), York (M2, S2), London (M3, S3) | Reconciler: both kept. Profile `toponym=York,London` |
| AC-1.5 | `sex_at_birth`: male (M1, S1), female weak (M2, S2) | Reconciler: female weak. Profile `sex_at_birth=male` |
| AC-1.6 | `sex_at_birth`: male (M1, S1), female provisional (M2, S2) | Reconciler: female provisional. Profile `sex_at_birth=male` |
| AC-1.7 | `date`: 1817 (M1, S1), 1830 weak (M2, S2), 1830 negative (M3, S3) | Reconciler: 1817 kept, 1830 denied. Profile `date=1817,¬1830` |
| AC-1.8 | `sex_at_birth`: female provisional (M1, S1) only | `sex_at_birth` is not a key of the profile; the profile is non-nil with length 0 |
| AC-1.9 | `name`: James Robins (M1, S1); `Entity.NamePattern` `"pattern-1"` | Profile `name` has one value: `Name` is the reconciled value's `Name` pointer, `NamePattern` `"pattern-1"`, `Type` `""`, `Ref` nil, not negative, not neutral |
| AC-1.10 | `sex_at_birth`: male (M1, S1) | The value's `Term` is `"male"`, `Type` `""`, `Ref` nil, `NamePattern` `""` |
| AC-1.11 | `sex_at_birth` with hand-built `Values` `[{Value: {TermKey: "unknown", Neutral: true}, Reason: kept}]` | One value, `Term` `"unknown"`, `Neutral` true |
| AC-1.12 | The zero `Entity` | Profile non-nil, length 0; error nil |

**US-2.** As the engine (the weak-claim review later), when I build a profile with `Options.Excluding`, I get the conclusion the group would reach without those members' Observations.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.1's entity, Excluding [M1] | Reconciler without M1: male 1, female 1, no majority. Profile `sex_at_birth=male,female` |
| AC-2.2 | AC-1.1's entity, Excluding [M3] | Profile `sex_at_birth=male` |
| AC-2.3 | AC-1.1's entity, Excluding [`{0,2,0,9}`] (no candidate) | Profile equals AC-1.1's |
| AC-2.4 | AC-1.1's entity plus `toponym` with `Values` reconciled from York (M2, S2) and `Candidates` set to nil by hand; Excluding [M1] | `sex_at_birth=male,female; toponym=York` (M1 has no `toponym` candidate, so its `Values` stand); error nil |
| AC-2.5 | AC-1.2's entity, Excluding [M2] (the negative's member) | Reconciler: the weak female alone is kept. Profile `sex_at_birth=female` |
| AC-2.6 | AC-1.1's entity, Excluding [M1, M2] | Profile `sex_at_birth=female` |
| AC-2.7 | `date`: 1817 (M1, S1) only, Excluding [M1] | Profile non-nil, length 0 |
| AC-2.8 | `date` with `Candidates` 1817 (M1, S1) and a candidate (M2, S2) whose `Value` is `{Text: "x", HasText: true}`; `Values` nil; Excluding [M1] | `errors.Is(err, autoreconcile.ErrValueMismatch)`; the profile is nil |
| AC-2.9 | Property `occupation` (no entry in the metas passed) with candidate text `x` (M1, S1) and `Values` nil; Excluding [M1] | `errors.Is(err, autoreconcile.ErrUnknownValueType)`; the profile is nil |
| AC-2.10 | AC-1.1's entity, metas whose `sex_at_birth` `Cardinality` is `""`, Excluding [M1] | Read as single: profile `sex_at_birth=male,female`, error nil |

**US-3.** As the engine, entity compare is `Build` then `subjectcompare.Compare`, and the map's Check holds: a Subject recording an outvoted value conflicts, one recording a denied value conflicts through the negative, and leave-one-out drops that member's support. In each row the profile is side A and the Subject side B; the kind is the property's. Side B is built with `ForSubject` from positive candidates with default provenance.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | AC-1.1's profile against person `sex_at_birth` `{female}` | sim 0, ev −1; `sex_at_birth` conflict, cause `relation` |
| AC-3.2 | AC-1.1's profile against person `sex_at_birth` `{male}` | sim 1, ev 0.6 |
| AC-3.3 | AC-1.2's profile against person `sex_at_birth` `{female}` | sim 0, ev −1; `sex_at_birth` conflict, cause `negative` |
| AC-3.4 | AC-1.2's profile against person `gender` `{female}` | sim 0, ev −1; `sex_at_birth` conflict, cause `negative` (the pool carries gender) |
| AC-3.5 | AC-1.2's profile against person `sex_at_birth` `{male}` | sim 0, ev 0; `sex_at_birth` unknown, cause `no_values` |
| AC-3.6 | AC-2.1's profile (Excluding [M1]) against person `sex_at_birth` `{female}` | sim 1, ev 0.6 |
| AC-3.7 | Person entity: `name` James Robins (M1, S1), James Robins (M2, S2); `sex_at_birth` male (M1, S1), male (M2, S2); against person `name` `{James Robins}`, `sex_at_birth` `{female}` | sim 0.7143, ev 1.5 |
| AC-3.8 | Event entity: `event_type` birth (M1, S1); AC-1.3's `date`; against event `event_type` `{birth}`, `date` `{1830}` | sim 1, ev 2.5 |
| AC-3.9 | AC-1.4's place profile against place `toponym` `{London}` | sim 1, ev 3 |

**US-4.** As a contributor, the builder's invariants hold on generated entities, so a later change that breaks one fails a test.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Any generated entity; Excluding a member id with no candidate | Profile labels equal to `Build` with the zero `Options` |
| AC-4.2 | Any generated entity and any member `m` of it | `Build(e, Excluding [m])` labels equal `Build(e')` labels, where `e'` is `e` with `m`'s candidates removed and each property's `Values` reconciled from what is left |
| AC-4.3 | Any generated entity; shuffle every property's `Candidates`; Excluding any member | Profile labels equal |
| AC-4.4 | Any generated entity, zero `Options` | For each property, the profile's positive count equals the count of its `Values` for which `Displayed()` is true, and its negative count the count with reason `denied` |

**US-5.** As the engine (S11-13's loader, S11-12's lifespan), when I build a plain Subject's side, its own Observations become positive and negative values through the same conversion as a profile's.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `sex_at_birth`: male, female negative | `sex_at_birth=male,¬female`, in input order |
| AC-5.2 | `name`: James Robins; name pattern `"pattern-1"` | One value with `NamePattern` `"pattern-1"` and the input's `Name` pointer |
| AC-5.3 | `sex_at_birth`: female weak, provisional | `sex_at_birth=female` (provenance and provisional are not read) |
| AC-5.4 | `date` with an empty candidate slice, and a nil map | Both give a non-nil profile of length 0 |
| AC-5.5 | AC-1.2's profile against `ForSubject` of `sex_at_birth`: female negative | sim 0, ev 0; `sex_at_birth` unknown, cause `no_values` (two negatives never meet) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `propertymeta.Meta.CardinalityOrSingle` (D-14) | STEP-1 |
| `subjectcomparetest.Labels`; `core/subjectcompare`'s private `labels` goes (D-10) | STEP-1 |
| `core/subjectprofile`: `Entity`, `Property`, `Candidate`, `Options`, `Build`, `ForSubject` (D-1, D-3 – D-8, D-15) | STEP-2, STEP-3, US-1, US-2, US-5 |
| The translation: displayed → positive, denied → negative, the rest out (D-4, D-5) | STEP-2, US-1 |
| Leave-one-out by running `autoreconcile.Reconcile` again (D-6, D-7) | STEP-3, US-2 |
| Entity compare tests: `Build` then `subjectcompare.Compare` (D-10) | STEP-4, US-3 |
| Invariant tests (D-11) | STEP-5, US-4 |
| Guard row; `match-rules.mdc` glob, autoreconcile line and Check line (D-2, D-12) | STEP-2, STEP-6 |
| Docs the layer makes true: `comparison-layers.md`, `matching-model.md`, the `add-match-rule` skill, the map's working rule and scope entry, the spike README, `docs/ideas/identity-claim-review.md`; S10-66's **Blocks** line (D-12, D-13) | STEP-6 |

| Out | Home |
| --- | --- |
| Any caller: Promote candidates, `match.Rank`, `matching.ForSubject` / `ForEntity`, `alikeNew`, `pairNode` | S11-13 |
| Loading entity values from the cache (`graphcache.Value` rows → `autoreconcile.ReconciledValue`) and Subjects' Observations (`matching.SubjectsValues` → `autoreconcile.Candidate`) | S11-13 (the loader that feeds `propertycompare` and `subjectprofile`) |
| Deleting `match.Value`'s reconciler fields (`Support`, `Against`, `Reason`, `Denied`, `SetEvidence`) and `pairValues`' leaders and weights | S11-13 (deletes `core/match`) |
| Lifespan in either profile, and the rules argument it needs | S11-12 |
| Reconciliation Claims (concluded values) laid over the profile | Not planned in Spike 11 — the claims PR adds them as an input that replaces a property's list ([`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md)); no caller produces a concluded value today |
| Negative Observations that deny no positive value (the reconciler keeps no row for them) as profile negatives | Not planned — the profile is strict and reads only the reconciler's output ([entities](../comparison-layers.md#entities-compare-against-the-conclusion)) |
| Value refs on profile values for evidence lines | S11-19 — the claim evidence query reads members, not matching internals ([claim evidence](../claim-evidence.md#separate-from-matching)) |
| Moving the other "empty cardinality reads as single" copies (`core/database/promotealign/exhibits.go`, `core/match/pair.go`, `core/database/properties/properties.go`) onto `CardinalityOrSingle` | `core/match/pair.go` goes with `core/match` in S11-13; `exhibits.go` with `pairExhibits` in S11-25; `properties.go` normalizes a catalog read and moves when next edited |
| The `conclusiondetails` string copy of `Displayed` (`core/database/conclusiondetails/conclusiondetails.go` `ReconciledValue.Displayed`) | Logged in the review log (M-2) — not matching code |
| A guard that fails when a layer package imports `core/matchregistry`, or one other than `core/subjectprofile` imports `core/autoreconcile` | Logged in the review log (M-9) for S11-16's planner, whose rewrite of `core/graphalign` (which imports `matchregistry` today) is where the rule first bites |
| Merge-hint and weak-claim review UIs | [`docs/ideas/identity-claim-review.md`](../../../ideas/identity-claim-review.md) (spike README "Not this spike") |
| A shared `autoreconcile` candidate fixture package | Not planned — `core/autoreconcile`'s own tests are internal; S11-12 adds rows to this package's tests |

## 4. Current state

Rows marked *(S10-66)* are named by S10-66's plan and land with it.

| Area | Where | Today |
| --- | --- | --- |
| Layer 4 package | `core/subjectprofile` | Does not exist |
| Reconciler | `core/autoreconcile/autoreconcile.go`: `Value` (`Text`, `HasText`, `Integer`, `HasInteger`, `TermID`, `TermKey`, `Neutral`, `SubjectID`, `EntityID`, `Date`, `Name`, `Depiction`), `Candidate` (`ObservationID`, `Value`, `SourceID`, `Negative`, `Provisional`, `Provenance`), `Provenance` (`Credibility`, `Uncertain`, `ClaimConfidence`), `ReconciledValue` (`Value`, `ObservationIDs`, `Support`, `Against`, `Reason`) and `ReconciledValue.Displayed()`, `Reason` and `ReasonKept`, `ReasonDenied`, `ReasonOutvoted`, `ReasonWeak`, `ReasonProvisional`, `Reason.Displayed()` ("the one rule the matcher and the reconciler share"), `Result` (`Values`, `Outcomes`, `Concluded`), `Reconcile(rules, valueType, candidates, concluded, cardinality) (Result, error)`, `ErrUnknownValueType`, `ErrValueMismatch` | Pure except `Value.Depiction *depictionvalues.Value`, which reaches `core/database` until S10-66; its tests import `core/matchregistry` (pure). `Reconcile` checks the value type first (`ErrUnknownValueType`), then the cardinality (only `single` / `multiple`), then each candidate (`ErrValueMismatch`). `Result.Values` holds every distinct positive value with its reason, displayed first; negative candidates and neutral terms produce no value |
| Reconciler purity | *(S10-66)* `core/depictionvalue` `Value`; `autoreconcile.Value.Depiction *depictionvalue.Value` | After S10-66, `go list -deps -test ./core/autoreconcile \| grep 'provenencia/core/database'` prints nothing (S10-66 AC-1.2) |
| Cache shape S11-13 reads | `core/database/graphcache/graph.go` `Value` (`Rank`, `Reason`, `Text`, …, `Support`, `Against`); `core/database/promotealign/canon.go` `handleValues`; `core/database/autoreconciler/autoreconciler.go` `valueRules` (`matchregistry.Default().Value`) | One cached row per `ReconciledValue`, with its reason as a string. Both `Reconcile` callers (`autoreconciler.go`, `truth.go`) pass a nil concluded value; claims are laid over the cache by composers, never written into it |
| Subject Observations | `core/database/autoreconciler/autoreconciler.go` `sqlLoadCandidates` (Observations as `autoreconcile.Candidate`); `core/database/matching/matching.go` `SubjectsValues` (as `match.Value`) | S11-13's loader reads a Subject's Observations; `ForSubject` takes them as candidates |
| Layer 3 | `core/subjectcompare/subjectcompare.go`: `Values` (`map[vocabkey.PropertyKey][]propertycompare.Value`), `Metas`, `Compare(kind, a, b, metas, rules, value) Result`, `Result{Similarity, Evidence, Properties}`, `PropertyResult{Key, Result, Points}` | `Values`' doc: "one Subject's values by property. Product properties only." |
| Layer 3 test labels | `core/subjectcompare/invariants_test.go` `labels(vals Vals) string` (sorted keys, `string(p)+"="+pt.Labels(vals[p])`, joined with `"; "`), called twice; `Vals` is an alias in `subjectcompare_test.go` | Private to `subjectcompare_test` |
| Layer 2 values | `core/propertycompare/propertycompare.go` `Value{valuecompare.Value; Negative, Neutral bool}`, `Value.Positive()`; `OutcomeConflict`, `OutcomeUnknown`, `CauseRelation`, `CauseNegative`, `CauseNoValues` | Sets every value's `Type` from the meta; skips a pair of two negatives; a negative that relates `Same` / `Equivalent` / `Contains` to a positive on the other side is a conflict, cause `negative` |
| Layer 1 values | `core/valuecompare/valuecompare.go` `Value{Type, Text, Term, Integer, Name, NamePattern, Date, Ref}`; `core/valuecompare/term.go` (an empty `Term` is `Unknown`); `core/valuecompare/name.go` `nameWords` (an unknown `NamePattern` reads as `Rules.Names.Western`) | No comparer for `subject` or `depiction` values: they are `Unknown` / `unsupported` |
| Property facts | `core/propertymeta/propertymeta.go` `Meta{ValueType, Cardinality}` ("empty reads as single", a comment only) | The same `if card == ""` fallback is written in `core/database/promotealign/exhibits.go`, `core/match/pair.go`, `core/database/properties/properties.go` |
| Fixtures | `core/propertycompare/propertycomparetest` (`Term`, `Text`, `Name`, `Date`, `Neg`, `Label`, `Labels`); `core/subjectcompare/subjectcomparetest/subjectcomparetest.go` `Metas()`; `core/datevalue/datevaluetest` `YMD`, `About`; `core/namevalue/namevaluetest` `Western`; `core/autoreconcile/autoreconcile_test.go` private `id`, `text`, `name`, `date` (internal package) | `Metas()` covers `name`, `sex_at_birth`, `gender`, `event_type`, `date`, `start_date`, `end_date`, `toponym` (multiple), `event_name`; not `occupation` |
| Registry | `core/matchregistry/registry.go` `Registry{Value, Subject}`, `Default()` | No profile section: the builder has no tunable (D-9) |
| Purity guard | `core/matchregistry/guard_test.go` `layerPackages` (last subject rows `{"core/subjectcompare", true}`, `{"core/subjectcompare/subjectcomparetest", false}`), `checkPure` (walks imports and test imports of every module package reached) | After S10-66 also `core/depictionvalue` rows |
| Rules text | `.agents/skills/add-match-rule/SKILL.md`: front matter `description`; bullet "**Layers are pure.** … and never imports `core/matchregistry` or `core/autoreconcile`."; the Layer 3 bullet; "Add a layer package" ("… and its path to the Check commands there and below."); "Check" block; "Do not" line "Import `core/autoreconcile` from a layer package."; `.cursor/rules/match-rules.mdc` `globs`, the purity line and the Check line | The Check command is written out in both files (S11-10 review log M-8, left for this planner) |
| Docs | `docs/deployment-plan/spike-11/comparison-layers.md` §The layers row 4, §Entities (ends with "**The profile builder is its own package:** …"), §Reuse across the app bullet "**Leave-one-out profiles:** `subjectprofile.Build(entity, Excluding(member))`. …"; `docs/matching-model.md` §The stack paragraph "The subject profile builder sits beside layer 4. …", §Glossary row **profile**; `docs/deployment-plan/spike-11/deployment-plan.md` working rule "**Entity compare is against the reconciled conclusion**" and §S11-11 scope entry; `docs/deployment-plan/spike-11/README.md` line 25; `docs/ideas/identity-claim-review.md` line 22 | Written before the builder existed: they name an `Excluding(member)` option and "concluded values positive (tied leaders … included)" |
| Old profile code | `core/match/match.go` `Value.SetEvidence`, `effective`, `votes`; `core/match/pairs.go` `weigh`, `pairValues` | Reads support and standing at compare time; deleted in S11-13 |

**Precedent.** S11-10 (`core/subjectcompare`, [#471](https://github.com/mendahu/provenencia/pull/471)): copy the package doc naming its layer and its purity, the external `_test` package with an AC-named table and a `check` field, the fixed-seed `TestInvariants` with one subtest per AC, and the `layerPackages` row with numbers true. Copy S11-10 STEP-9 for the shape of the docs step. Do not copy S11-10's registry section: this layer has none (D-9).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Package and entries.** `core/subjectprofile`, one file `subjectprofile.go`, plus tests. `func Build(e Entity, metas subjectcompare.Metas, reconcile valuecompare.Rules, opts Options) (subjectcompare.Values, error)` and `func ForSubject(values map[vocabkey.PropertyKey][]autoreconcile.Candidate, namePattern string) subjectcompare.Values`. Types (all *new*): `Entity`, `Property`, `Candidate`, `Options` (sketch in §6). Entity compare is `Build` then `subjectcompare.Compare`; no function joins the two | The map's working rule: "layer 4 is `subjectprofile.Build` plus `subjectcompare.Compare`, with no package of its own". The output is `subjectcompare.Values`, so a profile goes anywhere a Subject's values go, entity against entity included. The design builds lifespan "for a person Subject or person entity", so both sides come from this package | An `EntityCompare` helper — a second entry for the same comparison; a `Profile` type — a second name for `subjectcompare.Values`; leaving the Subject side to S11-13's loader — a second Observation → layer 2 conversion and a fork of S11-12's lifespan hook |
| D-2 | **The builder imports the reconciler.** `core/subjectprofile` is the one matching package that imports `core/autoreconcile`. Its non-test files import only the standard library, `core/autoreconcile`, `core/propertycompare`, `core/subjectcompare`, `core/valuecompare`, `core/propertymeta` and `core/vocabkey`. It is a `layerPackages` row with numbers true, so the guard also keeps `core/autoreconcile` pure from then on. The skill's and `match-rules.mdc`'s "never imports `core/autoreconcile`" gain this one exception (STEP-6) | The design puts the translation from standing to polarity in the builder ("Layer 4 translates them before comparing"; `match.Value`'s reconciler fields go "to the profile builder package"). Leave-one-out needs the reconciler itself (D-6). Until S10-66 the import reaches `core/database` | The builder's own copy of the reason strings and value fields — a second vocabulary for the reconciler's output, which S11-13's loader would convert into twice |
| D-3 | **Input shape.** `Entity{Properties map[vocabkey.PropertyKey]Property; NamePattern string}`; `Property{Values []autoreconcile.ReconciledValue; Candidates []Candidate}`; `Candidate{Member []byte; autoreconcile.Candidate}`. `Values` is the reconciler's output for the property: a fresh `Reconcile` result's `Values`, or the cache's rows (S11-13). `Candidates` is the reconciler's input for it, each tagged with the member Subject its Observation is on; only `Options.Excluding` reads it. `NamePattern` is the catalog's name format key. Product properties only, as `subjectcompare.Values` | The cache stores one row per `ReconciledValue` with its reason. The member is not on `autoreconcile.Candidate`, and leave-one-out is by member | `autoreconcile.Result` — its `Concluded` flag and `Outcomes` are never read (D-4); candidates keyed by Observation id with a separate member map — two lookups for one fact |
| D-4 | **The translation.** For each property, in `Values` order: a value for which `ReconciledValue.Displayed()` is true becomes a positive value; one with `Reason` `autoreconcile.ReasonDenied` becomes a negative value; every other value is left out. A property with no value left is not a key of the profile (AC-1.8). The builder reads only auto-reconciled values: Reconciliation Claims are laid over them by their own PR (Out) | The design's profile table: displayed values positive (several when tied leaders show, or when multi-valued), denied negative, outvoted / weak / provisional out. `Displayed()` is "the one rule the matcher and the reconciler share", so the profile follows any change to it. The cache never holds a claim's value (`conclusion-reconciliation.md`) | `Reason == ReasonKept` — a second copy of `Displayed`; a concluded-value branch — no caller produces one, and claims replace a list rather than one value; an empty slice for an empty property — two shapes of "nothing" for layer 3 |
| D-5 | **One value conversion.** An unexported `value(v autoreconcile.Value, negative bool, namePattern string) propertycompare.Value` returns `propertycompare.Value{Value: valuecompare.Value{Text: v.Text, Term: v.TermKey, Integer: v.Integer, Name: v.Name, Date: v.Date}, Negative: negative, Neutral: v.Neutral}`, with `NamePattern` = `namePattern` when `v.Name` is not nil. `Type` stays empty (layer 2 sets it from the meta), `Ref` stays nil. `Name` and `Date` pointers are shared with the input, not copied. `Build` and `ForSubject` both call it | One conversion from the reconciler's value shape to layer 2's for both sides. Layers never write through these pointers. The reconciler never yields a neutral value, but `Neutral` is the value's own fact and layer 2 reads it; a Subject's neutral term reaches `ForSubject` | `autoreconcile`'s private `compareValue` — partial on purpose (names and terms go another way there) and the reconciler reads no catalog name format; a value ref from `ObservationIDs` — several ids for one value, and evidence lines are S11-19's |
| D-6 | **Leave-one-out.** `Options{Excluding [][]byte}`: the member Subjects to leave out; nil or empty leaves none out. With at least one, for each property: if no candidate's `Member` equals (`bytes.Equal`) a left-out member, its `Values` stand; otherwise `Build` calls `autoreconcile.Reconcile(reconcile, meta.ValueType, kept, nil, meta.CardinalityOrSingle())`, where `kept` is the property's remaining candidates' `autoreconcile.Candidate` in input order and `meta` is `metas[key]` (the zero `Meta` when absent), and translates the new result's `Values` by D-4 | The reconciler's result depends only on its candidates, so a property the members never touched needs no second run (AC-2.4), and leave-one-out costs one `Reconcile` per property they recorded. A slice field gives several members for free | Removing the members' Observation ids from each `ReconciledValue` — a majority or a denial can change without them (AC-2.1, AC-2.5); re-running every property — one call per property for nothing; functional options (`Excluding(member) Option`) — a new options idiom beside the repo's `Options` structs (`core/database/matching` `Options`, `core/raster` `Options`) |
| D-7 | **Order and errors.** `Build` walks `e.Properties` in ascending key order (`slices.Sorted(maps.Keys(…))`) and returns the first `Reconcile` error unwrapped (`autoreconcile.ErrUnknownValueType`, `autoreconcile.ErrValueMismatch`), with a nil profile. With no member left out, the error is always nil. `ForSubject` cannot fail | Sorted order makes the returned error the same on every run. The reconciler's sentinels already say what is wrong; S11-13's loader and the tests check them with `errors.Is` | A new `subjectprofile` sentinel — the same fact under two names; skipping the property — hides a loader bug |
| D-8 | **Which rules.** `Build`'s `reconcile valuecompare.Rules` is the rules the reconciler runs under. Its doc comment says callers pass `matchregistry.Default().Value`, the rules the cache was built with (`autoreconciler.valueRules`), not `valuerules.Load`'s catalog rules that `subjectcompare.Compare` takes | Rule 5: a layer takes its rules as an argument and never reads the registry. The cache reconciles under the shipped registry, never a catalog's alias tables; S11-13's caller holds both rule sets side by side, so the name says which | Reading `matchregistry.Default()` inside `Build` — a layer importing the registry; calling the parameter `value` — the same name as `Compare`'s, for different rules |
| D-9 | **No registry section, no numbers.** The builder has no tunable: which reasons count is the design's rule, not a knob. `layerPackages` gains `{"core/subjectprofile", true}` after `{"core/subjectcompare/subjectcomparetest", false}`. No inventory row, no `notYetBuilt` change | The match registry file list has no profile file; lifespan's numbers (S11-12) are the `subject.person.lifespan.` rows, already named | A `profile` section holding the reason lists — a knob nobody should turn |
| D-10 | **Tests and fixtures.** **`subjectcomparetest.Labels(v subjectcompare.Values) string`** *(new)*, doc "Labels names a Subject's values for a failure message and an expected value: each property in ascending key order as `key=` and its values' `propertycomparetest.Labels`, joined with `; `.", the body of `core/subjectcompare`'s private `labels`, which is deleted; its two calls in `invariants_test.go` call `subjectcomparetest.Labels`. **This package's tests**: external package `subjectprofile_test`, no build tags. `core/subjectprofile/helpers_test.go` *(new)*: `obs(n byte) []byte` (`{0,0,0,n}`), `src(n byte) []byte` (`{0,1,0,n}`), `mem(n byte) []byte` (`{0,2,0,n}`); `term(key string)` (`TermID` `[]byte(key)`, `TermKey` key), `text(s string)` (`Text` s, `HasText` true), `year(y int)` (`Date` `datevaluetest.YMD(y, 0, 0)`), `name(form string)` (`Name` `namevaluetest.Western(form)`), each an `autoreconcile.Value` with every other field zero; `cand(m, o, s byte, v autoreconcile.Value) subjectprofile.Candidate` (`Member` `mem(m)`, `ObservationID` `obs(o)`, `SourceID` `src(s)`); `weak(c)`, `provisional(c)`, `negative(c)` returning a changed copy of a `subjectprofile.Candidate`; `property(t *testing.T, key vocabkey.PropertyKey, cs ...subjectprofile.Candidate) subjectprofile.Property` (reconciles as §2's conventions say, `t.Fatal` on error, `Candidates` = `cs`). `subjectprofile_test.go` *(new)*: `TestBuild` (US-1, US-2), rows named by AC (`"AC-1.1 outvoted value is left out"`), fields `entity`, `metas` (nil → `subjectcomparetest.Metas()`), `opts subjectprofile.Options`, `want string` (`subjectcomparetest.Labels` of the profile), `wantErr error`, and `check func(t *testing.T, p subjectcompare.Values)`; rows AC-1.8, AC-1.9, AC-1.10, AC-1.11, AC-1.12, AC-2.7, AC-2.8, AC-2.9 have a `check` (non-nil and length, or nil, or the field values the AC names). `TestForSubject` (US-5 except AC-5.5), same fields minus `entity`, `metas` and `opts`, plus `values map[vocabkey.PropertyKey][]autoreconcile.Candidate` and `namePattern`. `entity_test.go` *(new)*: `TestEntityCompare` (US-3, AC-5.5), fields `kind vocabkey.SubjectType`, `entity`, `opts`, `subject map[vocabkey.PropertyKey][]autoreconcile.Candidate` (built into side B with `ForSubject(subject, "")`), `sim`, `ev`, `check func(t *testing.T, r subjectcompare.Result)` | One label format for profiles and Subjects, so expected strings read the same in layer 3's and layer 4's tests, S11-12's and S11-13's. The reconciler's own builders are internal to its package | A private copy of `labels` here — a second copy of S11-10's; exporting test builders from `autoreconcile` — a fixture package for one caller; `reflect.DeepEqual` on profiles — pointer sharing (D-5) makes the expected value hard to write |
| D-11 | **Generator.** `invariants_test.go` *(new)* `TestInvariants`, one subtest per AC (`AC-4.1`, `AC-4.2`, `AC-4.3`, `AC-4.4`), 500 cases each, `rand.New(rand.NewSource(1))`. An entity has 1 – 4 members (`mem(1)`, `mem(2)`, `mem(3)`, `mem(4)`); for each of `sex_at_birth` (pool male, female), `date` (pool 1817, 1830, `datevaluetest.About(1817, 0, 0)`) and `toponym` (pool York, London), each member draws 0 – 2 candidates; each candidate's Source is `src(1 + r.Intn(3))`, Observation ids count up from 1 across the entity, and it is negative when `r.Intn(4) == 0`, weak when `r.Intn(4) == 0`, provisional when `r.Intn(6) == 0`. Each property comes from `property`. AC-4.1 excludes `mem(9)`; AC-4.2 and AC-4.3 pick the member with `r.Intn`; AC-4.3 shuffles with `r.Shuffle`. AC-4.1 – AC-4.3 compare `subjectcomparetest.Labels` of the two profiles; AC-4.4 counts. A failure reports `t.Errorf("case %d: %s: got %s, want %s", i, what, got, want)` with labels | The S11-09 and S11-10 generator shape; the pools reach kept, outvoted, weak, provisional and denied values, mixed dates, a multiple property and members with no candidate on a property. Labels show order and polarity; the generator has no names, so `NamePattern` cannot differ | Random strings — would rarely group |
| D-12 | **Rules text and the Check command.** `match-rules.mdc`: `globs` gains `core/subjectprofile/**` after `core/subjectcompare/**`; the purity line names the one exception; the Check line points at the skill's Check section. The skill's Check block gains `./core/subjectprofile/...` and stays the one place the command is written, and its "Add a layer package" recipe stops sending the next PR to the rule's Check (exact text in STEP-6) | Resolves S11-10 review log M-8: every layer PR edited the command in two files. The rule already says "Follow the project skill" | A script that reads `layerPackages` — new tooling for a one-line list |
| D-13 | **Docs.** `comparison-layers.md` §The layers row 4 and §Entities say it is built and how; its "Leave-one-out profiles" bullet, the map's working rule and S11-11 scope entry, the spike README and `docs/ideas/identity-claim-review.md` name `Options.Excluding` instead of `Excluding(member)`, and the map scope entry says "displayed values positive" instead of "concluded values positive"; `matching-model.md` names the rule; the skill gains layer 4. S10-66's plan header **Blocks** line gains S11-11. Exact text in STEP-6 | The design doc, the map and the model page are kept true as layers land; the map's working rule "One matching skill" grows the skill with each layer | — |
| D-14 | **`func (m Meta) CardinalityOrSingle() string`** *(new)* in `core/propertymeta/propertymeta.go`, doc "CardinalityOrSingle is m's cardinality, with empty read as single.", returns `CardinalitySingle` when `m.Cardinality` is `""`, else `m.Cardinality`. `Build` calls it. Test `TestCardinalityOrSingle` *(new)* in `core/propertymeta/propertymeta_test.go` *(new)*: `""` → `single`, `single` → `single`, `multiple` → `multiple` | The rule lives with its type, not in a fifth call site | An inline `if card == ""` in `Build` — a fifth copy |
| D-15 | **`ForSubject`.** For each key of `values`, each candidate becomes `value(c.Value, c.Negative, namePattern)`, in input order; every other `Candidate` field (provenance, provisional, Source, Observation) is not read. A key with no candidates is not a key of the profile; the result is never nil | A Subject speaks only for itself: all its Observations count, positive or negative, which is what layer 2 takes. S11-13's loader reads Observations as `autoreconcile.Candidate` the way the reconciler's loader does, so one conversion serves both sides | `map[…][]autoreconcile.Value` plus a negative flag — `Negative` is on `Candidate`, not on `Value` |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/subjectprofile` *(new)* | package | `core/match` `Value.SetEvidence`, `pairValues`, `weigh` | `match` reads support and standing at compare time, is impure, and is deleted in S11-13 ("Build, then migrate"); the design names this package |
| `subjectprofile.Build` *(new)* | func | `core/database/matching` `ForEntity` / `EntitiesValues`; `promotealign` `handleValues` | Those load the catalog and keep every rank with its evidence for `match`; `Build` is pure and keeps only the conclusion. S11-13 makes the loaders feed it |
| `subjectprofile.ForSubject` *(new)* | func | `core/database/matching` `SubjectsValues` (Observations as `match.Value`) | `SubjectsValues` is the catalog read; `ForSubject` is the pure half it will call in S11-13, sharing `value` with `Build` (D-15) |
| `Entity`, `Property` *(new)* | type | `autoreconcile.Result`; `graphcache.Node.Values` | `Property` holds `[]autoreconcile.ReconciledValue` (the reconciler's own output type) with its candidates; no new value type |
| `Candidate` *(new)* | type | `autoreconcile.Candidate` | Embeds it and adds the member Subject, which leave-one-out needs and the reconciler does not |
| `Options` *(new)* | type | `core/database/matching` `Options`, `core/raster` `Options` | Same struct-of-options idiom; one field |
| `value` *(new, unexported)* | func | `autoreconcile` private `Value.compareValue`; `match.Value` → `valuecompare.Value` in `core/match` | `compareValue` is partial on purpose and never sets a catalog name pattern; `match`'s goes in S11-13 (D-5) |
| `propertymeta.Meta.CardinalityOrSingle` *(new)* | method | inline `if card == ""` in `exhibits.go`, `match/pair.go`, `properties.go` | The rule's home; the copies move per the Out row (D-14) |
| `subjectcomparetest.Labels` *(new)* | helper | `core/subjectcompare/invariants_test.go` `labels` | The same body, made importable; the private copy is deleted (D-10) |
| Test builders in `core/subjectprofile/helpers_test.go` *(new)* | helper | `core/autoreconcile/autoreconcile_test.go` `id`, `text`, `name`, `date` | Internal to the reconciler's package; these also tag members and Sources (D-10) |
| `TestBuild`, `TestForSubject`, `TestEntityCompare`, `TestInvariants`, `TestCardinalityOrSingle` *(new)* | test | S11-10's `TestCompare`, `TestInvariants` | Same shapes, this layer's cases |

**Builds on.** `autoreconcile.Reconcile`, `ReconciledValue.Displayed` and the reasons (Spike 9, S11-06), `subjectcompare.Values` / `Metas` / `Compare` (S11-10), `propertycompare.Value` (S11-09), `propertycomparetest` and `subjectcomparetest` (S11-10), the `layerPackages` guard, the `add-match-rule` skill.
**Deletes.** `core/subjectcompare/invariants_test.go` `labels` (now `subjectcomparetest.Labels`). Nothing else becomes dead: `core/match`'s evidence fields and `core/database/matching` keep their callers until S11-13.
**Next adds.** S11-12 adds lifespan to both entries: `Entity` and `ForSubject` gain the direct neighbors' dates, both take the subject section's lifespan rules as a new argument (the `subject.person.lifespan.` window rows), and each adds one value for the lifespan lead, which `subjectcompare` already weighs by its `subject.Weights.person.lifespan.` rows. S11-13 loads `Property.Values` from `graphcache` rows (`autoreconcile.Reason(row.Reason)`) and Subjects' Observations as candidates, calls `Build` with `matchregistry.Default().Value` and `ForSubject` with the catalog's name pattern, and deletes `match`'s evidence fields. Merge hints are `subjectcompare.Compare` on two `Build` results. The weak-claim review is `Build` with `Options.Excluding` and the entity's candidates. Reconciliation Claims add an input to `Property` that replaces its list.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/propertymeta` | `Meta.CardinalityOrSingle` *(new)* (D-14) |
| `core/subjectcompare/subjectcomparetest` | `Labels` *(new)* (D-10) |
| `core/subjectcompare` | tests only: `invariants_test.go` calls `subjectcomparetest.Labels`; private `labels` deleted |
| `core/subjectprofile` *(new)* | `subjectprofile.go`: the types, `Build`, `ForSubject`, `value` (D-1 – D-8, D-15) |
| `core/matchregistry` | `guard_test.go` `layerPackages` row (D-9) |

```go
// Package subjectprofile is the profile builder beside layer 4 of matching:
// it turns an entity's reconciled conclusion, or a Subject's own
// Observations, into the Subject-shaped values layer 3 compares. It compares
// nothing. Pure: no I/O, no catalog, no numbers; the reconciler's output and
// rules come in as arguments.
package subjectprofile

// Entity is one canonical entity's auto-reconciled product Properties.
type Entity struct {
	Properties  map[vocabkey.PropertyKey]Property
	NamePattern string // the catalog's name format key, set on every name value
}

// Property is the reconciler's output for one Property and, for leave-one-out,
// its input.
type Property struct {
	Values     []autoreconcile.ReconciledValue // a Reconcile result's Values, or the cache's rows
	Candidates []Candidate                     // read only by Options.Excluding
}

// Candidate is a reconciler candidate and the member Subject its Observation
// is on.
type Candidate struct {
	Member []byte
	autoreconcile.Candidate
}

// Options change how Build reads the entity.
type Options struct {
	// Excluding builds the profile as if these members' Observations were not
	// in the group: each Property they recorded is reconciled again without them.
	Excluding [][]byte
}

// Build returns the entity's profile: on each Property, the displayed values
// as positives and the values a stronger negative denied as negatives;
// outvoted, weak and provisional values are left out, and a Property with
// nothing left is absent. reconcile is the rules the reconciler runs under:
// pass matchregistry.Default().Value, as the cache does, not a catalog's
// valuerules. metas give each Property's value type and cardinality for
// Options.Excluding. The error is Reconcile's, and only Excluding runs it.
func Build(e Entity, metas subjectcompare.Metas, reconcile valuecompare.Rules, opts Options) (subjectcompare.Values, error)

// ForSubject returns a Subject's own values: each candidate positive or
// negative as recorded, in order. Only Value and Negative are read.
func ForSubject(values map[vocabkey.PropertyKey][]autoreconcile.Candidate, namePattern string) subjectcompare.Values
```

`Build`: for each key of `e.Properties` in ascending order, take `vals := p.Values`; when `opts.Excluding` names a member that has a candidate on it, re-run per D-6 and take the new `Values`; translate `vals` per D-4 with `value` (D-5); set the key only when the slice is not empty. `out` starts as `subjectcompare.Values{}`, so the zero `Entity` gives an empty, non-nil map (AC-1.12).

### FFI

None — no caller yet.

### macOS

None — no app change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Shared fixtures and the cardinality rule**
Files: `core/propertymeta/propertymeta.go`, `core/propertymeta/propertymeta_test.go` *(new)*, `core/subjectcompare/subjectcomparetest/subjectcomparetest.go`, `core/subjectcompare/invariants_test.go`.
Do: first run `go test ./core/subjectcompare/... -v | grep -c -- '--- PASS'` and write the count down. D-14 with `TestCardinalityOrSingle` (T-6). D-10's `subjectcomparetest.Labels`; delete the private `labels` and call `subjectcomparetest.Labels` at its two call sites (no test case changes).
Check: `go test ./core/propertymeta/... ./core/subjectcompare/... ./core/matchregistry/...` passes with no tags; the `subjectcompare` PASS count is unchanged; `rg -n 'func labels' core/subjectcompare` prints nothing.

**STEP-2 — The package, the translation, `ForSubject` and the guard**
Files: `core/subjectprofile/subjectprofile.go` *(new)*, `core/subjectprofile/helpers_test.go` *(new)*, `core/subjectprofile/subjectprofile_test.go` *(new)*, `core/matchregistry/guard_test.go`. Skill: `add-match-rule` ("Add a layer package").
Do: first run `go list -deps -test ./core/autoreconcile | grep 'provenencia/core/database'` and confirm it prints nothing (S10-66 landed). Then the package doc, every type and function of §6 with their doc comments, `Build` with D-4, D-5 and D-7's order (`Options.Excluding` not yet read), `ForSubject` (D-15), the `layerPackages` row (D-9), the helpers (D-10), and `TestBuild` with the US-1 rows (T-1) and `TestForSubject` (T-2).
Check: `go test ./core/subjectprofile/...` passes with no tags; `go test ./core/matchregistry/ -run TestLayerPackages -v` passes and lists `core/subjectprofile/pure` and `core/subjectprofile/numbers`.

**STEP-3 — Leave-one-out**
Files: `core/subjectprofile/subjectprofile.go` and `core/subjectprofile/subjectprofile_test.go` (both (new) in STEP-2).
Do: D-6 and D-7 in `Build`. Add the US-2 rows to `TestBuild` (T-3).
Check: `go test ./core/subjectprofile/...` passes; `go vet ./core/subjectprofile/...` is clean.

**STEP-4 — Entity compare**
Files: `core/subjectprofile/entity_test.go` *(new)*.
Do: `TestEntityCompare` with the US-3 rows and AC-5.5 (T-4): build side A with `Build`, side B with `ForSubject(subject, "")`, then call `subjectcompare.Compare(kind, a, b, subjectcomparetest.Metas(), matchregistry.Default().Subject, matchregistry.Default().Value)`; AC-3.1, AC-3.3, AC-3.4, AC-3.5 and AC-5.5 check the `sex_at_birth` `PropertyResult`'s outcome and cause in `check`.
Check: `go test ./core/subjectprofile/... -run TestEntityCompare -v` passes.

**STEP-5 — Invariants**
Files: `core/subjectprofile/invariants_test.go` *(new)*.
Do: `TestInvariants` per D-11 (T-5).
Check: `go test ./core/subjectprofile/... -run TestInvariants -v` passes; then `go test ./core/subjectprofile/... ./core/subjectcompare/... ./core/propertycompare/... ./core/propertymeta/... ./core/valuecompare/... ./core/autoreconcile/... ./core/matchregistry/...` passes with no tags.

**STEP-6 — Docs**
Files: `.cursor/rules/match-rules.mdc`, `.agents/skills/add-match-rule/SKILL.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching-model.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (working rule, scope entry, map row), `docs/deployment-plan/spike-11/README.md`, `docs/ideas/identity-claim-review.md`, `docs/deployment-plan/spike-10-followups/prs/S10-66-pure-depiction-value.md` (header only), `docs/deployment-plan/spike-11/completed.md`, this plan (status).
Do:
- `match-rules.mdc` (D-12): in `globs`, insert `core/subjectprofile/**,` after `core/subjectcompare/**,`; the line "- Layer packages are pure: no `core/database`, no build tags; non-test files import no `core/matchregistry` or `core/autoreconcile`." becomes "- Layer packages are pure: no `core/database`, no build tags; non-test files import no `core/matchregistry`, and only `core/subjectprofile` imports `core/autoreconcile`, to turn its output into a profile."; the line "- Check: `go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/matchregistry/...` with no tags." becomes "- Check: the commands in the skill's Check section."
- `add-match-rule` SKILL.md:
  - Front matter `description`: "(core/valuecompare, core/propertycompare, core/subjectcompare, core/matchregistry)" becomes "(core/valuecompare, core/propertycompare, core/subjectcompare, core/subjectprofile, core/matchregistry)", and in its "Use when" sentence, "an entity profile rule," is inserted after "a Subject property weight, pool or period,"; rewrap the paragraph to the width of the other lines (80 columns).
  - The "**Layers are pure.**" bullet's last clause "and never imports `core/matchregistry` or `core/autoreconcile`." becomes "and never imports `core/matchregistry`; only `core/subjectprofile` imports `core/autoreconcile`, to turn its output into a profile."
  - A bullet *(new)* after the Layer 3 bullet: "**The profile builder (`core/subjectprofile`) turns the reconciler's output into a Subject.** `Build` keeps a Property's displayed values (`ReconciledValue.Displayed`) as positives and its denied values as negatives, and leaves everything else out; `ForSubject` turns a Subject's own Observations into positives and negatives through the same conversion. Layer 3 compares either like any Subject, so entity compare is `Build` then `subjectcompare.Compare`. `Options.Excluding` reconciles each Property those members recorded again without their candidates. Callers pass `Build` `matchregistry.Default().Value`, the rules the cache was built with. The builder has no registry section."
  - "Add a layer package": "and its path to the Check commands there and below." becomes "and its path to the Check command below."
  - "Check": the command becomes `go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/subjectprofile/... ./core/matchregistry/...`.
  - "Do not": "Import `core/autoreconcile` from a layer package." becomes "Import `core/autoreconcile` from a layer package other than `core/subjectprofile`."
- `comparison-layers.md`:
  - §The layers, row 4's Status cell "Decided in shape (§ Entities): subject comparison against the entity's reconciled conclusion" becomes "Built in S11-11 (§ Entities): subject comparison against the entity's reconciled conclusion".
  - §Reuse across the app, in the bullet that starts "- **Leave-one-out profiles:**", the call `subjectprofile.Build(entity, Excluding(member))` becomes `subjectprofile.Build(entity, metas, rules, Options{Excluding: members})`; the rest of the bullet stays.
  - At the end of §Entities: compare against the conclusion, after the paragraph that starts "**The profile builder is its own package:**", a paragraph *(new)*: "**Built in S11-11.** `subjectprofile.Build(entity, metas, rules, opts)` reads each Property's auto-reconciled values: values the reconciler displays become positives, values with reason `denied` negatives, and the rest are left out. A Property with nothing left is absent. The output is `subjectcompare.Values`, so entity compare is `Build` then `subjectcompare.Compare`, and entity against entity is `Compare` on two profiles. `Options.Excluding` runs `autoreconcile.Reconcile` again on each Property those members recorded, without their candidates. `subjectprofile.ForSubject` builds a plain Subject's side from its own Observations through the same conversion. A negative Observation that denies no positive value leaves no reconciled value, so it is not in the profile. Reconciliation Claims, when they ship, are laid over the profile as a replacing list."
- `matching-model.md`: §The stack, the paragraph "The subject profile builder sits beside layer 4. It turns the reconciler's conclusion into a Subject-shaped profile and compares nothing." becomes "The subject profile builder sits beside layer 4. It turns the reconciler's conclusion into a Subject-shaped profile and compares nothing: displayed values are positive, denied values negative, and outvoted, weak and provisional values are left out. It also builds a plain Subject's side from its own Observations." §Glossary, the **profile** row becomes "| **profile** | A Subject-shaped set of values built from an entity's reconciled conclusion: displayed values positive, denied values negative, nothing else |".
- `deployment-plan.md`: in the working rule "**Entity compare is against the reconciled conclusion**", "`subjectprofile.Build` also takes an `Excluding(member)` option." becomes "`subjectprofile.Build` also takes `Options.Excluding`, a list of members to leave out."; in §S11-11's **Ships.** line, "concluded values positive (tied leaders and multi-valued properties included), denied values negative, others left out; `Excluding(member)`." becomes "displayed values positive (tied leaders and multi-valued properties included), denied values negative, others left out; `Options.Excluding`; `ForSubject` for a plain Subject's side."
- `README.md` (spike 11) line 25: "`subjectprofile.Build` with `Excluding`" becomes "`subjectprofile.Build` with `Options.Excluding`". `docs/ideas/identity-claim-review.md` line 22: "`subjectprofile.Build` with `Excluding(member)`, S11-11" becomes "`subjectprofile.Build` with `Options.Excluding`, S11-11".
- S10-66 plan header: "**Blocks:** S11-19 (Spike 11 map: the claim evidence query carries likeness lines through the pure matching stack)" becomes "**Blocks:** S11-11 (the profile builder imports `core/autoreconcile` from a layer package), S11-19 (Spike 11 map: the claim evidence query carries likeness lines through the pure matching stack)". Skip this line if S10-66's plan is already deleted or archived at your base.
- `completed.md`, newest first: "## S11-11 — Subject profile and entity compare" / "**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N). The profile builder exists: `subjectprofile.Build` turns an entity's reconciled values into a Subject-shaped profile, with displayed values as positives, denied values as negatives and nothing else, so entity compare is `Build` then `subjectcompare.Compare`. `Options.Excluding` reconciles again without some members' Observations, for leave-one-out checks, and `ForSubject` builds a plain Subject's side. No caller yet; nothing changes for researchers." with the date and PR number filled in.
- Map row status `Landed #N`; this plan's Status `Landed in [#N](…) (YYYY-MM-DD)` and a **Landed** note under the header listing each difference from the plan (or "None.").
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-11-subject-profile.md` reports no errors; `rg -c 'core/autoreconcile' .agents/skills/add-match-rule/SKILL.md` prints 2 (the "Layers are pure" bullet and the Do not line) and `rg -c 'core/autoreconcile' .cursor/rules/match-rules.mdc` prints 1; `rg -n 'go test' .cursor/rules/match-rules.mdc` prints nothing; `rg -n 'Excluding\(member\)' docs` prints only lines in `prs/S11-10-subject-compare.md` and this plan; links resolve.

Commits: one per step, in step order. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Add the subject profile builder, which turns an entity's conclusion into a Subject."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/subjectprofile/subjectprofile_test.go` `TestBuild` *(new)* | Rows AC-1.1 – AC-1.12, inputs and expectations as in §2 | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, AC-1.12, D-3, D-4, D-5 |
| T-2 | Go | same file, `TestForSubject` *(new)* | Rows AC-5.1 – AC-5.4 | AC-5.1, AC-5.2, AC-5.3, AC-5.4, D-15 |
| T-3 | Go | same file, `TestBuild` | Rows AC-2.1 – AC-2.10 | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, AC-2.10, D-6, D-7 |
| T-4 | Go | `core/subjectprofile/entity_test.go` `TestEntityCompare` *(new)* | Rows AC-3.1 – AC-3.9, AC-5.5 | AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.6, AC-3.7, AC-3.8, AC-3.9, AC-5.5, D-1 |
| T-5 | Go | `core/subjectprofile/invariants_test.go` `TestInvariants` *(new)* | 500 generated entities per AC (D-11) → AC-4.1 – AC-4.4 hold | AC-4.1, AC-4.2, AC-4.3, AC-4.4, D-11 |
| T-6 | Go | `core/propertymeta/propertymeta_test.go` `TestCardinalityOrSingle` *(new)* | `""` → `single`; `single` → `single`; `multiple` → `multiple` | D-14 |
| T-7 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` | New row → `core/subjectprofile` pure (through `core/autoreconcile`) and free of numbers | D-2, D-9 |
| T-8 | Go | `core/subjectcompare` `TestInvariants` | Unchanged cases on `subjectcomparetest.Labels` → same PASS count | D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestLayerPackages` (a stray import, or S10-66 not landed), `core/subjectcompare`'s tests after the labels move, `core/autoreconcile`'s tests (must not change: this PR does not touch the package).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] go test ./core/subjectprofile/... ./core/subjectcompare/... ./core/propertycompare/... ./core/propertymeta/... ./core/valuecompare/... ./core/autoreconcile/... ./core/matchregistry/... passes with no build tags
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without macOS every check but xcode-test passes, and CI runs the Xcode tests on the PR (this PR changes no Swift)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored; core/autoreconcile, core/match, core/database and core/subjectcompare's non-test files are unchanged
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none (no copy)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: comparison-layers.md, matching-model.md, deployment-plan.md (working rule, scope entry), spike README, docs/ideas/identity-claim-review.md, add-match-rule skill, match-rules.mdc, S10-66 plan header (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — nothing reaches it until S11-13. The reviewer runs `go test ./core/subjectprofile/... -v` with no build tags and reads the `TestEntityCompare` subtests `AC-3.1 …`, `AC-3.3 …` and `AC-3.6 …`: these are the map's Check (a Subject recording an outvoted value conflicts; one recording a denied value conflicts via the negative; leave-one-out drops that member's support). Then `go test ./core/matchregistry/ -run TestLayerPackages -v` to see `core/subjectprofile` pass the purity and numbers guard.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S10-66 has not landed, so the guard fails through `autoreconcile.Value.Depiction` | STEP-2's first command; the plan stays Blocked until S10-66 lands |
| S10-66 lands with names other than its plan's | The first stop condition |
| A reconciler change after the planning base moves an outcome §2 reads (majority, denial, weak) | §2's outcomes were read from `Reconcile` at `bfe09bd`; a drift is a stop condition, not a test edit |
| Map iteration order leaks into the output or the error | D-7: sorted keys; labels sort keys |

**Stop and report if:**

- `go list -deps -test ./core/autoreconcile | grep 'provenencia/core/database'` prints anything at your base, or `autoreconcile.Value.Depiction` is not `*depictionvalue.Value`.
- `main` differs from §4: `subjectcompare.Values`, `Metas`, `Compare`'s signature or `Result`, `propertycomparetest`'s builders, `subjectcomparetest.Metas()`'s entries, `subjectcompare`'s private `labels`, `ReconciledValue.Displayed`, or `Reconcile`'s signature. Report each difference.
- `autoreconcile.Reconcile` at your base gives a reason or an order different from §2 for an AC's candidates (for example AC-1.1's female is not `outvoted`, AC-1.2's female or AC-1.7's 1830 is not `denied`, AC-2.1 is not two kept values, or AC-2.5's female is not kept): report the candidates and the result; do not change the AC.
- `subjectcompare.Compare` gives a sim or ev different from §2 for a US-3 row or AC-5.5: report the row and the output.
- `core/subjectprofile` cannot pass `TestLayerPackages` with only the imports D-2 lists.
- An invariant of US-4 fails under D-4 – D-7 as written: report the generated case (seed, index, the entity's candidates) and which invariant.
- `core/subjectprofile` already exists, or `propertymeta.Meta` already has a cardinality method.
- Merging `main` conflicts in a paragraph STEP-6 rewrites, so keeping both sides' text is not possible.

## 12. Open questions

None.

## 13. PR

**Title:** S11-11 — Subject profile and entity compare
**Body:** objective in one sentence; what shipped against each US; AC evidence (test names); deviations from the plan; manual check (the `go test` runs of §10); `Plan: docs/deployment-plan/spike-11/prs/S11-11-subject-profile.md`.
