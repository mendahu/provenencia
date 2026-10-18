# S11-11 — Subject profile and entity compare

**Status:** Blocked — on S10-66
**Spike / slice:** Spike 11 · Slice 3 — Property, Subject and entity compare
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a new pure Go package with no caller yet; nothing reaches the app)
**Board:** —
**Depends on:** S11-10 (in review as [#471](https://github.com/mendahu/provenencia/pull/471), branch `s11-10-subject-compare` @ `3bd1348`; [plan](S11-10-subject-compare.md) and its **Landed** note: it shipped as planned, with `Compare` complete in one commit). S10-66 ([plan](../../spike-10-followups/prs/S10-66-pure-depiction-value.md), Planned): it moves the depiction value type into pure `core/depictionvalue`, which makes `core/autoreconcile` pure. This PR imports `core/autoreconcile` from a layer package, so it cannot pass the purity guard before S10-66 lands (D-2). Branch from `main` after both land
**Blocks:** S11-12
**Branch:** `s11-11-subject-profile`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) ("Add a layer package")
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Entities: compare against the conclusion](../comparison-layers.md#entities-compare-against-the-conclusion) (the profile table, "the profile is strict", "the profile builder is its own package"); [reuse across the app](../comparison-layers.md#reuse-across-the-app) ("Leave-one-out profiles", "Profiles on both sides"); [Layer 2 contract](../comparison-layers.md#layer-2-property-comparison) ("No reconciler context … Layer 4 translates them before comparing"); [`docs/matching-model.md`](../../../matching-model.md) stack table row 4, glossary **profile**; [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) (reasons); working rules "Entity compare is against the reconciled conclusion", "Pure means the package builds and tests with no build tags and no `core/database` import", "Build, then migrate", "One guard, not one per package", "One matching skill, grown with the layers"
**Planned:** 2026-10-18 against `main` @ `dc67a09` with S11-10's branch `s11-10-subject-compare` @ `3bd1348` as the expected base for every S11-10 name; S10-66 names from its plan
**Review:** — (not reviewed yet)

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

Layer 4 of matching exists: `subjectprofile.Build` *(new)* turns an entity's reconciled conclusion into a Subject-shaped profile (`subjectcompare.Values`), so a Subject is compared with the entity by `subjectcompare.Compare` against that profile. The profile is strict: displayed values are positive, values a stronger negative Observation denied are negative, and every other value is left out. `Excluding(member)` builds the profile as if that member's Observations were not in the group, by running the reconciler again without them.

**Why now.** S11-12 derives a lifespan inside the profile builder, and S11-13 moves Promote's candidate scoring onto entity compare. "Build, then migrate" puts the builder in its own PR with its own tests before any caller moves.

**What stays.** `core/match` (`Value.SetEvidence`, `pairValues`, `weigh`, the leaders and support weights) and `core/database/matching` are untouched; S11-13 moves Promote onto this layer and deletes them. `core/autoreconcile` and `core/subjectcompare` are unchanged. No catalog, wire, Swift or researcher-visible change.

## 2. User stories

Conventions for every AC. Members are `M1`, `M2`, `M3` (member Subject ids `{0,2,0,n}`), Sources `S1`, `S2`, `S3` (`{0,1,0,n}`), Observations numbered in the order written (`{0,0,0,n}`). "male (M1, S1)" is a candidate on member M1, cited under Source S1, positive, with default provenance. "weak" is `Provenance.Credibility` −1; "provisional" is `Candidate.Provisional`; "negative" is `Candidate.Negative`. Unless an AC says otherwise, a property's `Result` is `autoreconcile.Reconcile(matchregistry.Default().Value, …)` over its candidates with the property's `subjectcomparetest.Metas()` value type and cardinality and no concluded value, and its `Candidates` are those candidates. Profile values are written as their term key, text, year or name form; "¬x" is a negative value. "sim" and "ev" are `subjectcompare.Result.Similarity` and `.Evidence` (within 0.0001) under `matchregistry.Default()` and `subjectcomparetest.Metas()`. Reconciler outcomes and every sim / ev below were read from the code on `s11-10-subject-compare` @ `3bd1348`.

**US-1.** As the engine (Promote S11-13, merge hints later), when I build an entity's profile, I get its conclusion as positives, its denied values as negatives, and nothing else.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `sex_at_birth`: male (M1, S1), male (M2, S2), female (M3, S3) | Reconciler: male kept, female outvoted. Profile `sex_at_birth` = `[male]` |
| AC-1.2 | `sex_at_birth`: female weak (M1, S1), female negative (M2, S2) | Reconciler: female denied. Profile `sex_at_birth` = `[¬female]` |
| AC-1.3 | `date`: 1817 (M1, S1), 1830 (M2, S2) | Reconciler: both kept (mixed). Profile `date` = `[1817, 1830]`, in that order |
| AC-1.4 | `toponym` (multiple): York (M1, S1), York (M2, S2), London (M3, S3) | Reconciler: both kept. Profile `toponym` = `[York, London]` |
| AC-1.5 | `sex_at_birth`: male (M1, S1), female weak (M2, S2) | Reconciler: female weak. Profile `sex_at_birth` = `[male]` |
| AC-1.6 | `sex_at_birth`: male (M1, S1), female provisional (M2, S2) | Reconciler: female provisional. Profile `sex_at_birth` = `[male]` |
| AC-1.7 | `sex_at_birth`: male (M1, S1), male (M2, S2), concluded female | Reconciler: `Concluded`, female kept at rank 1, male kept. Profile `sex_at_birth` = `[female]` (D-4) |
| AC-1.8 | `sex_at_birth`: female provisional (M1, S1) only | `sex_at_birth` is not a key of the profile |
| AC-1.9 | `name`: James Robins (M1, S1); `Entity.NamePattern` `"pattern-1"` | Profile `name` has one value: `Name` is the reconciled value's `Name` pointer, `NamePattern` `"pattern-1"`, `Type` `""`, `Ref` nil, not negative, not neutral |
| AC-1.10 | `sex_at_birth`: male (M1, S1) | The value's `Term` is `"male"`, `Type` `""`, `Ref` nil, `NamePattern` `""` |
| AC-1.11 | `sex_at_birth` with a hand-built `Result{Values: [{Value: {TermKey: "unknown", Neutral: true}, Reason: kept}]}` | Profile `sex_at_birth` = one value, `Term` `"unknown"`, `Neutral` true |
| AC-1.12 | The zero `Entity` | Profile has length 0; error nil |

**US-2.** As the engine (the weak-claim review later), when I build a profile `Excluding(member)`, I get the conclusion the group would reach without that member's Observations.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.1's entity, `Excluding(M1)` | Reconciler without M1: male 1, female 1, no majority. Profile `sex_at_birth` = `[male, female]` |
| AC-2.2 | AC-1.1's entity, `Excluding(M3)` | Profile `sex_at_birth` = `[male]` |
| AC-2.3 | AC-1.1's entity, `Excluding` a member id `{0,2,0,9}` that has no candidate | Profile equals AC-1.1's |
| AC-2.4 | AC-1.1's entity plus `toponym` with `Result` from York (M2, S2) and `Candidates` nil; `Excluding(M1)` | `sex_at_birth` = `[male, female]`; `toponym` = `[York]` (M1 has no `toponym` candidate, so its `Result` stands); error nil |
| AC-2.5 | AC-1.7's entity, `Excluding(M1)` | The concluded female is passed again; profile `sex_at_birth` = `[female]` |
| AC-2.6 | AC-1.1's entity, `Excluding(M1)` and `Excluding(M2)` | Profile `sex_at_birth` = `[female]` |
| AC-2.7 | `date`: 1817 (M1, S1) only, `Excluding(M1)` | `date` is not a key of the profile |
| AC-2.8 | `date` with `Candidates` 1817 (M1, S1) and a candidate (M2, S2) whose `Value` is `{Text: "x", HasText: true}`; `Result` zero; `Excluding(M1)` | `errors.Is(err, autoreconcile.ErrValueMismatch)`; the profile is nil |
| AC-2.9 | Property `occupation` (no entry in the metas passed) with candidate text `x` (M1, S1) and `Result` zero; `Excluding(M1)` | `errors.Is(err, autoreconcile.ErrUnknownValueType)` |
| AC-2.10 | AC-1.1's entity, metas whose `sex_at_birth` `Cardinality` is `""`, `Excluding(M1)` | Read as single: profile `sex_at_birth` = `[male, female]`, error nil |

**US-3.** As the engine, entity compare is `Build` then `subjectcompare.Compare`, and the map's Check holds: a Subject recording an outvoted value conflicts, one recording a denied value conflicts through the negative, and leave-one-out drops that member's support. In each row the profile is side A and the Subject side B, built with `propertycomparetest`; the kind is the property's.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | AC-1.1's profile against person `sex_at_birth` `{female}` | sim 0, ev −1; `sex_at_birth` conflict, cause `relation` |
| AC-3.2 | AC-1.1's profile against person `sex_at_birth` `{male}` | sim 1, ev 0.6 |
| AC-3.3 | AC-1.2's profile against person `sex_at_birth` `{female}` | sim 0, ev −1; `sex_at_birth` conflict, cause `negative` |
| AC-3.4 | AC-1.2's profile against person `gender` `{female}` | sim 0, ev −1; `sex_at_birth` conflict, cause `negative` (the pool carries gender) |
| AC-3.5 | AC-1.2's profile against person `sex_at_birth` `{male}` | sim 0, ev 0; `sex_at_birth` unknown, cause `no_values` |
| AC-3.6 | AC-2.1's profile (`Excluding(M1)`) against person `sex_at_birth` `{female}` | sim 1, ev 0.6 |
| AC-3.7 | Person entity: `name` James Robins (M1, S1), James Robins (M2, S2); `sex_at_birth` male (M1, S1), male (M2, S2); against person name `{James Robins}`, `sex_at_birth` `{female}` | sim 0.7143, ev 1.5 |
| AC-3.8 | Event entity: `event_type` birth (M1, S1); AC-1.3's `date`; against event `event_type` `{birth}`, `date` `{1830}` | sim 1, ev 2.5 |
| AC-3.9 | AC-1.4's place profile against place `toponym` `{London}` | sim 1, ev 3 |

**US-4.** As a contributor, the builder's invariants hold on generated entities, so a later change that breaks one fails a test.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Any generated entity; `Excluding` a member id with no candidate | Profile equal to `Build` with no option |
| AC-4.2 | Any generated entity and any member `m` of it | `Build(e, Excluding(m))` equals `Build(e')`, where `e'` is `e` with `m`'s candidates removed and each property's `Result` reconciled from what is left |
| AC-4.3 | Any generated entity; shuffle every property's `Candidates`; `Excluding` any member | Profile equal |
| AC-4.4 | Any generated entity | Every profile value with `Negative` false comes from a `Result` value with reason `kept`, every value with `Negative` true from one with reason `denied`, and each property's value count equals the count of those two reasons in its `Result` (for a `Concluded` result: 1 plus the denied count) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `core/subjectprofile`: `Entity`, `Property`, `Candidate`, `Option`, `Excluding`, `Build` (D-1, D-3 – D-9) | STEP-1, STEP-2, US-1, US-2 |
| The translation: kept → positive, denied → negative, the rest out; `Concluded` keeps only rank 1 (D-4, D-5) | STEP-1, US-1 |
| Leave-one-out by running `autoreconcile.Reconcile` again (D-6 – D-8) | STEP-2, US-2 |
| Entity compare tests: `Build` then `subjectcompare.Compare` (D-10) | STEP-3, US-3 |
| Invariant tests (D-11) | STEP-4, US-4 |
| Guard row; `match-rules.mdc` glob, autoreconcile line and Check line (D-2, D-12) | STEP-1, STEP-5 |
| Docs the layer makes true: `comparison-layers.md` §Entities, `matching-model.md`, the `add-match-rule` skill; S10-66's **Blocks** line (D-12, D-13) | STEP-5 |

| Out | Home |
| --- | --- |
| Any caller: Promote candidates, `match.Rank`, `matching.ForSubject` / `ForEntity`, `alikeNew`, `pairNode` | S11-13 |
| Loading entity values from the cache (`graphcache.Value` rows → `autoreconcile.ReconciledValue`) and Subject values (`matching.SubjectsValues`) | S11-13 (the loader that feeds `propertycompare` and `subjectprofile`) |
| Deleting `match.Value`'s reconciler fields (`Support`, `Against`, `Reason`, `Denied`, `SetEvidence`) and `pairValues`' leaders and weights | S11-13 (deletes `core/match`) |
| Lifespan in the profile | S11-12 |
| A profile for a plain (non-entity) Subject | S11-13 — a Subject's own Observations are already positives and negatives; its loader builds `subjectcompare.Values` directly |
| Negative Observations that deny no positive value (the reconciler keeps no row for them) as profile negatives | Not planned — the profile is strict and reads only the reconciler's output ([entities](../comparison-layers.md#entities-compare-against-the-conclusion)) |
| Value refs on profile values for evidence lines | S11-19 — the claim evidence query reads members, not matching internals ([claim evidence](../claim-evidence.md#separate-from-matching)) |
| Merge-hint and weak-claim review UIs | [`docs/ideas/identity-claim-review.md`](../../../ideas/identity-claim-review.md) (spike README "Not this spike") |
| A guard that fails when a layer package other than `core/subjectprofile` imports `core/autoreconcile` | Not planned — the rule stays in the skill and `match-rules.mdc`; nothing has tried it |
| A shared `autoreconcile` candidate fixture package | Not planned — `core/autoreconcile`'s own tests are internal; S11-12 adds rows to this package's tests |

## 4. Current state

Rows marked *(S11-10)* are on `s11-10-subject-compare` @ `3bd1348` and land with #471; "(new)" on such a row means new on `main` until #471 merges. Rows marked *(S10-66)* are named by S10-66's plan and land with it.

| Area | Where | Today |
| --- | --- | --- |
| Layer 4 package | `core/subjectprofile` | Does not exist |
| Reconciler | `core/autoreconcile/autoreconcile.go`: `Value` (`Text`, `HasText`, `Integer`, `HasInteger`, `TermID`, `TermKey`, `Neutral`, `SubjectID`, `EntityID`, `Date`, `Name`, `Depiction`), `Candidate` (`ObservationID`, `Value`, `SourceID`, `Negative`, `Provisional`, `Provenance`), `Provenance` (`Credibility`, `Uncertain`, `ClaimConfidence`), `ReconciledValue` (`Value`, `ObservationIDs`, `Support`, `Against`, `Reason`), `Reason` and `ReasonKept`, `ReasonDenied`, `ReasonOutvoted`, `ReasonWeak`, `ReasonProvisional`, `Reason.Displayed()`, `Result` (`Values`, `Outcomes`, `Concluded`), `Reconcile(rules, valueType, candidates, concluded, cardinality) (Result, error)`, `ErrUnknownValueType`, `ErrValueMismatch` | Pure except `Value.Depiction *depictionvalues.Value`, which reaches `core/database` until S10-66. `Reconcile` rejects a cardinality other than `single` / `multiple` and a value type outside the closed set; `Result.Values` holds every distinct positive value with its reason (displayed first); negative candidates and neutral terms produce no value; with a concluded value, rank 1 is it (`ReasonKept`) and the other values keep their reasons, kept ones included (`applyConcluded`) |
| Reconciler purity | *(S10-66)* `core/depictionvalue` `Value`; `autoreconcile.Value.Depiction *depictionvalue.Value` | After S10-66, `go list -deps -test ./core/autoreconcile \| grep 'provenencia/core/database'` prints nothing (S10-66 AC-1.2) |
| Cache shape S11-13 reads | `core/database/graphcache/graph.go` `Value` (`Rank`, `Reason`, `Text`, …, `Support`, `Against`); `core/database/promotealign/canon.go` `handleValues` | One cached row per `ReconciledValue`, with its reason as a string. `Concluded` is never set today: both `Reconcile` callers (`core/database/autoreconciler/autoreconciler.go`, `truth.go`) pass a nil concluded value |
| Layer 3 | *(S11-10)* (new) `core/subjectcompare/subjectcompare.go`: `Values` (`map[vocabkey.PropertyKey][]propertycompare.Value`), `Metas` (`map[vocabkey.PropertyKey]propertymeta.Meta`), `Compare(kind, a, b, metas, rules, value) Result`, `Result{Similarity, Evidence, Properties}`, `PropertyResult{Key, Result, Points}` | `Values`' doc: "one Subject's values by property. Product properties only." |
| Layer 2 values | `core/propertycompare/propertycompare.go` `Value{valuecompare.Value; Negative, Neutral bool}`; *(S11-10)* `Value.Positive()`; `OutcomeConflict`, `OutcomeUnknown`, `OutcomeAgree`, `CauseRelation`, `CauseNegative`, `CauseNoValues` | Sets every value's `Type` from the meta; a negative on one side that relates `Same` / `Equivalent` / `Contains` to a positive on the other is a conflict, cause `negative` |
| Layer 1 values | `core/valuecompare/valuecompare.go` `Value{Type, Text, Term, Integer, Name, NamePattern, Date, Ref}`; `core/valuecompare/term.go` (an empty `Term` is `Unknown`); `core/valuecompare/name.go` `nameWords` (an unknown `NamePattern` reads as `Rules.Names.Western`) | No comparer for `subject` or `depiction` values: they are `Unknown` / `unsupported` |
| Fixtures | *(S11-10)* (new) `core/propertycompare/propertycomparetest` (`Term`, `Text`, `Name`, `Date`, `Neg`, `Labels`, …); *(S11-10)* `core/subjectcompare/subjectcomparetest.Metas()`; `core/datevalue/datevaluetest` `YMD`; `core/namevalue/namevaluetest` `Western`; `core/autoreconcile/autoreconcile_test.go` private `id`, `text`, `name`, `date` (package `autoreconcile`, internal) | `Metas()` covers `name`, `sex_at_birth`, `gender`, `event_type`, `date`, `start_date`, `end_date`, `toponym` (multiple), `event_name`; not `occupation` |
| Registry | *(S11-10)* `core/matchregistry/registry.go` `Registry{Value, Subject}`, `Default()` | No profile section: the builder has no tunable (D-9) |
| Purity guard | *(S11-10)* `core/matchregistry/guard_test.go` `layerPackages` (rows through `core/subjectcompare/subjectcomparetest`), `checkPure` (walks imports and test imports of every module package reached) | After S10-66 also `core/depictionvalue` and `core/depictionvalue/depictionvaluetest` |
| Rules text | *(S11-10)* `.agents/skills/add-match-rule/SKILL.md`: front matter `description`; "Rules that hold everywhere" bullet "**Layers are pure.** … and never imports `core/matchregistry` or `core/autoreconcile`."; the Layer 2 bullet ("… turned into positives and negatives by the profile builder before they reach it."); the Layer 3 bullet; "Check" block; "Do not" line "Import `core/autoreconcile` from a layer package."; `.cursor/rules/match-rules.mdc` `globs`, the line "Layer packages are pure: no `core/database`, no build tags; non-test files import no `core/matchregistry` or `core/autoreconcile`." and the Check line | The Check command is written out in both files (S11-10 review log M-8, left for this planner) |
| Docs | `docs/deployment-plan/spike-11/comparison-layers.md` §The layers table row 4, §Entities: compare against the conclusion (ends with "**The profile builder is its own package:** …"); `docs/matching-model.md` §The stack paragraph "The subject profile builder sits beside layer 4. …", §Glossary row **profile** | Describe the builder before it exists |
| Old profile code | `core/match/match.go` `Value.SetEvidence`, `effective`, `votes`; `core/match/pairs.go` `weigh`, `pairValues` | Reads support and standing at compare time; deleted in S11-13 |

**Precedent.** S11-10 (`core/subjectcompare`, branch `s11-10-subject-compare`): copy the package doc naming its layer and its purity, the external `_test` package with an AC-named table and a `check func(t *testing.T, …)` field, the fixed-seed `TestInvariants` with one subtest per AC, and the `layerPackages` row with numbers true. Copy S11-10 STEP-9 for the shape of the docs step. Do not copy S11-10's registry section: this layer has none (D-9).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Package and entry.** `core/subjectprofile`, one file `subjectprofile.go`, plus tests. `func Build(e Entity, metas subjectcompare.Metas, value valuecompare.Rules, opts ...Option) (subjectcompare.Values, error)`. Types (all *new*): `Entity`, `Property`, `Candidate`, `Option`, and the unexported `options` (sketch in §6). Entity compare is `Build` then `subjectcompare.Compare`; no function joins the two | The map's working rule: "layer 4 is `subjectprofile.Build` plus `subjectcompare.Compare`, with no package of its own". The output is `subjectcompare.Values`, so a profile goes anywhere a Subject's values go, entity against entity included | An `EntityCompare` helper — a second entry for the same comparison; a `Profile` type — a second name for `subjectcompare.Values` |
| D-2 | **The builder imports the reconciler.** `core/subjectprofile` is the one matching package that imports `core/autoreconcile`: it reads `ReconciledValue`, `Reason`, `Result` and calls `Reconcile`. Its non-test files import only the standard library, `core/autoreconcile`, `core/propertycompare`, `core/subjectcompare`, `core/valuecompare`, `core/propertymeta` and `core/vocabkey`. It is a `layerPackages` row with numbers true, so the guard also keeps `core/autoreconcile` pure from then on. The skill's and `match-rules.mdc`'s "never imports `core/autoreconcile`" gain this one exception (STEP-5) | The design puts the translation from standing to polarity in the builder ("Layer 4 translates them before comparing"; `match.Value`'s reconciler fields go "to the profile builder package"). Leave-one-out needs the reconciler itself (D-6). Until S10-66 the import reaches `core/database` | The builder's own copy of the reason strings and value fields — a second vocabulary for the reconciler's output, which S11-13's loader would convert into twice |
| D-3 | **Input shape.** `Entity{Properties map[vocabkey.PropertyKey]Property; NamePattern string}`; `Property{Result autoreconcile.Result; Candidates []Candidate}`; `Candidate{Member []byte; autoreconcile.Candidate}`. `Result` is the reconciler's output for the property: a fresh `Reconcile` result, or the cache's rows as `Values` (S11-13). `Candidates` is the reconciler's input for it, each tagged with the member Subject its Observation is on; only `Excluding` reads it. `NamePattern` is the catalog's name format key. Product properties only, as `subjectcompare.Values` | The cache stores one row per `ReconciledValue` with its reason, and `Result.Concluded` is the one flag the translation needs beyond the values (D-4). The member is not on `autoreconcile.Candidate`, and leave-one-out is by member | `[]autoreconcile.ReconciledValue` alone — loses `Concluded`; candidates keyed by Observation id with a separate member map — two lookups for one fact |
| D-4 | **The translation.** For each property, in `Result.Values` order: a value with `Reason` `autoreconcile.ReasonKept` becomes a positive value; one with `ReasonDenied` becomes a negative value; every other reason is left out. When `Result.Concluded` is true, only `Values[0]` is positive (the concluded value); the other kept values are left out, denied values stay negative. A property with no value left is not a key of the profile (AC-1.8) | The design's profile table: concluded values positive (several when tied leaders show, or when multi-valued), denied negative, outvoted / weak / provisional out. A concluded value is the researcher overriding the auto-reconciler, so its other kept values are no longer the conclusion; `applyConcluded` keeps them kept for display only | Every kept value under `Concluded` — the researcher's conclusion would sit beside the values it replaced; an empty slice for an empty property — two shapes of "nothing" for layer 3 |
| D-5 | **One profile value.** From an `autoreconcile.Value` `v`: `propertycompare.Value{Value: valuecompare.Value{Text: v.Text, Term: v.TermKey, Integer: v.Integer, Name: v.Name, Date: v.Date}, Negative: <D-4>, Neutral: v.Neutral}`, and `NamePattern` = `e.NamePattern` when `v.Name` is not nil. `Type` stays empty (layer 2 sets it from the meta), `Ref` stays nil, `Text` / `Integer` are copied as they are. `Name` and `Date` pointers are shared with the input, not copied. An unexported `value(v autoreconcile.Value, negative bool, namePattern string) propertycompare.Value` does it | Layer 2 sets `Type` from the meta, so the profile never carries one of its own. Layers never write through these pointers. The reconciler never yields a neutral value today, but `Neutral` is the value's own fact and layer 2 reads it | `autoreconcile`'s private `compareValue` — partial on purpose (names and terms go another way there) and the reconciler reads no catalog name format; a value ref from `ObservationIDs` — several ids for one value, and evidence lines are S11-19's |
| D-6 | **Leave-one-out.** `Excluding(member []byte) Option` adds `member` to the members to leave out; several `Excluding` options leave out every named member. With at least one member to leave out, for each property: if no candidate's `Member` equals (`bytes.Equal`) a left-out member, its `Result` stands; otherwise `Build` calls `autoreconcile.Reconcile(value, meta.ValueType, kept, concluded, card)`, where `kept` is the property's remaining candidates' `autoreconcile.Candidate` in input order, `concluded` is `&Result.Values[0].Value` when `Result.Concluded` and nil otherwise, `meta` is `metas[key]` (the zero `Meta` when absent), and `card` is `meta.Cardinality`, or `propertymeta.CardinalitySingle` when that is empty. The new result is translated by D-4 | The reconciler's result depends only on its candidates and concluded value, so a property the member never touched needs no second run (AC-2.4) and leave-one-out costs one `Reconcile` per property the member recorded. A researcher's conclusion does not depend on which members are in the group (AC-2.5). "Empty reads as single" is `propertymeta.Meta`'s rule, and `Reconcile` accepts only the two named cardinalities | Removing the member's Observation ids from each `ReconciledValue` — a majority or a denial can change without them (AC-2.1); re-running every property — one call per property for nothing; one excluded member only — the option form makes several free |
| D-7 | **Errors.** `Build` returns an error only from `Reconcile` under `Excluding`, unwrapped (`autoreconcile.ErrUnknownValueType`, `autoreconcile.ErrValueMismatch`), with a nil profile. With no `Excluding`, the error is always nil | The reconciler's sentinels already say what is wrong; S11-13's loader and the tests check them with `errors.Is` | A new `subjectprofile` sentinel — the same fact under two names; ignoring the property — hides a loader bug |
| D-8 | **Which rules.** `value` is the `valuecompare.Rules` the reconciler runs under. The doc comment of `Build` says callers pass `matchregistry.Default().Value`, the rules the cache was built with (`autoreconciler.valueRules`), so a leave-one-out profile agrees with the cached ones | Rule 5: a layer takes its section as an argument and never reads the registry. The cache reconciles under the shipped registry, never a catalog's alias tables | Reading `matchregistry.Default()` inside `Build` — a layer importing the registry (skill: "never imports `core/matchregistry`") |
| D-9 | **No registry section, no numbers.** The builder has no tunable: which reasons count is the design's rule, not a knob. `layerPackages` gains `{"core/subjectprofile", true}` after `{"core/subjectcompare/subjectcomparetest", false}`. No inventory row, no `notYetBuilt` change | The match registry file list has no profile file; lifespan's numbers (S11-12) live in the subject section's `subject.person.lifespan.` rows, already named | A `profile` section holding the reason lists — a "knob" nobody should turn |
| D-10 | **Tests.** External package `subjectprofile_test`, no build tags. `core/subjectprofile/helpers_test.go` *(new)* holds the private builders: `obs(n byte) []byte` (`{0,0,0,n}`), `src(n byte) []byte` (`{0,1,0,n}`), `mem(n byte) []byte` (`{0,2,0,n}`); `term(key string)`, `text(s string)`, `year(y int)` (`datevaluetest.YMD(y, 0, 0)`), `name(form string)` (`namevaluetest.Western`), each an `autoreconcile.Value` (a term sets `TermID` to `[]byte(key)` and `TermKey`); `cand(m, o, s byte, v autoreconcile.Value) subjectprofile.Candidate`; `weak(c)`, `provisional(c)`, `negative(c)` returning a changed copy; `property(t *testing.T, key vocabkey.PropertyKey, concluded *autoreconcile.Value, cs ...subjectprofile.Candidate) subjectprofile.Property` (reconciles as §2's conventions say, `t.Fatal` on error); `labels(v subjectcompare.Values) string` (per key in ascending order, `key=` and `propertycomparetest.Labels`, joined with `; `). `subjectprofile_test.go` *(new)*: `TestBuild` (US-1, US-2), rows named by AC (`"AC-1.1 outvoted value is left out"`), fields `entity`, `metas` (nil → `subjectcomparetest.Metas()`), `opts []subjectprofile.Option`, `want string` (the `labels` of the profile, or `""`), `wantErr error`, and `check func(t *testing.T, p subjectcompare.Values)` for AC-1.9 – AC-1.11. `entity_test.go` *(new)*: `TestEntityCompare` (US-3), rows with `kind`, `entity`, `opts`, `subject subjectcompare.Values`, `sim`, `ev`, `check` | One `want` string compares order and polarity in one line, the way S11-10's tests read labels. The reconciler's own builders are internal to its package | Exporting test builders from `autoreconcile` — a fixture package for one caller; comparing `subjectcompare.Values` with `reflect.DeepEqual` — pointer sharing (D-5) makes the expected value hard to write |
| D-11 | **Generator.** `invariants_test.go` *(new)* `TestInvariants`, one subtest per AC (`AC-4.1`, `AC-4.2`, `AC-4.3`, `AC-4.4`), 500 cases each, `rand.New(rand.NewSource(1))`. An entity has 1 – 4 members (`mem(1)`, `mem(2)`, `mem(3)`, `mem(4)`); for each of `sex_at_birth` (pool male, female), `date` (pool 1817, 1830, `datevaluetest.About(1817, 0, 0)`) and `toponym` (pool York, London), each member draws 0 – 2 candidates; each candidate's Source is `src(1 + r.Intn(3))`, Observation ids count up from 1 across the entity, and it is negative when `r.Intn(4) == 0`, weak when `r.Intn(4) == 0`, provisional when `r.Intn(6) == 0`. Each property's `Result` comes from `property` (no concluded value). AC-4.1 excludes `mem(9)`; AC-4.2 and AC-4.3 pick the member with `r.Intn`; AC-4.3 shuffles with `r.Shuffle`. A failure reports `t.Errorf("case %d: %s: got %s, want %s", i, what, labels(got), labels(want))` | The S11-09 and S11-10 generator shape; the pools reach kept, outvoted, weak, provisional and denied values, mixed dates, a multiple property and members with no candidate on a property | Random strings — would rarely group |
| D-12 | **Rules text and the Check command.** `match-rules.mdc`: `globs` gains `core/subjectprofile/**` after `core/subjectcompare/**`; the purity line becomes "Layer packages are pure: no `core/database`, no build tags; non-test files import no `core/matchregistry`, and only `core/subjectprofile` imports `core/autoreconcile`, to turn its output into a profile."; the Check line becomes "Check: the commands in the skill's Check section." The skill's Check block gains `./core/subjectprofile/...` and stays the one place the command is written (exact text in STEP-5) | Resolves S11-10 review log M-8: the command is listed in the skill and repeated in the rule, and every layer PR edits both. The rule already says "Follow the project skill" | A script that reads `layerPackages` — new tooling for a one-line list |
| D-13 | **Docs.** `comparison-layers.md` §The layers row 4's Status becomes "Built in S11-11 (§ Entities)", and §Entities gains a **Built in S11-11** paragraph; `matching-model.md`'s §The stack paragraph and glossary **profile** row name the rule; the skill gains layer 4 (description, purity exception, a rule bullet, Check, Do not). S10-66's plan header **Blocks** line gains S11-11. Exact text in STEP-5 | The design doc and the model page are kept true as layers land; the map's working rule "One matching skill" grows the skill with each layer | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/subjectprofile` *(new)* | package | `core/match` `Value.SetEvidence`, `pairValues`, `weigh` | `match` reads support and standing at compare time, is impure, and is deleted in S11-13 ("Build, then migrate"); the design names this package |
| `subjectprofile.Build` *(new)* | func | `core/database/matching` `ForEntity` / `EntitiesValues`; `promotealign` `handleValues` | Those load the catalog and keep every rank with its evidence for `match`; `Build` is pure and keeps only the conclusion. S11-13 makes the loaders feed it |
| `Entity`, `Property` *(new)* | type | `autoreconcile.Result`; `graphcache.Node.Values` | `Property` wraps `autoreconcile.Result` (the reconciler's own output type) with its candidates; no new value type |
| `Candidate` *(new)* | type | `autoreconcile.Candidate` | Embeds it and adds the member Subject, which leave-one-out needs and the reconciler does not |
| `Option`, `Excluding` *(new)* | type, func | none in the repo (no functional options today) | The design names `Excluding(member)` as an option of `Build` ([reuse across the app](../comparison-layers.md#reuse-across-the-app)) |
| `value` *(new, unexported)* | func | `autoreconcile` private `Value.compareValue`; `match.Value` → `valuecompare.Value` in `core/match` | `compareValue` is partial on purpose and never sets a catalog name pattern; `match`'s goes in S11-13 (D-5) |
| Test builders in `core/subjectprofile/helpers_test.go` *(new)* | helper | `core/autoreconcile/autoreconcile_test.go` `id`, `text`, `name`, `date` | Internal to the reconciler's package; these also tag members and Sources (D-10) |
| `TestBuild`, `TestEntityCompare`, `TestInvariants` *(new)* | test | S11-10's `TestCompare`, `TestInvariants` | Same shapes, this layer's cases |

**Builds on.** `autoreconcile.Reconcile` and its `Result` (Spike 9, S11-06), `subjectcompare.Values` / `Metas` / `Compare` (S11-10), `propertycompare.Value` (S11-09), `propertycomparetest` and `subjectcomparetest` (S11-10), the `layerPackages` guard, the `add-match-rule` skill.
**Deletes.** None — nothing becomes dead: `core/match`'s evidence fields and `core/database/matching` keep their callers until S11-13.
**Next adds.** S11-12 adds lifespan to `Build`: `Entity` gains the direct neighbors' dates and `Build` adds one value for the lifespan lead, which `subjectcompare` already weighs by its `subject.Weights.person.lifespan.` rows. S11-13 loads `Property.Result` from `graphcache` rows (`Values` with `autoreconcile.Reason(row.Reason)`), calls `Build` with `matchregistry.Default().Value` and the catalog's metas and name pattern, and deletes `match`'s evidence fields. Merge hints are `subjectcompare.Compare` on two `Build` results. The weak-claim review is `Build` with `Excluding` and the entity's candidates.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/subjectprofile` *(new)* | `subjectprofile.go`: the types, `Excluding`, `Build`, `value` (D-1 – D-8) |
| `core/matchregistry` | `guard_test.go` `layerPackages` row (D-9) |

```go
// Package subjectprofile is the profile builder beside layer 4 of matching:
// it turns an entity's reconciled conclusion into a Subject-shaped profile
// that layer 3 compares like any Subject. It compares nothing. Pure: no I/O,
// no catalog, no numbers; the reconciler's output and rules come in as
// arguments.
package subjectprofile

// Entity is one canonical entity's reconciled product Properties.
type Entity struct {
	Properties  map[vocabkey.PropertyKey]Property
	NamePattern string // the catalog's name format key, set on every name value
}

// Property is the reconciler's output for one Property and, for Excluding,
// its input.
type Property struct {
	Result     autoreconcile.Result // a Reconcile result, or the cache's rows as Values
	Candidates []Candidate          // read only by Excluding
}

// Candidate is a reconciler candidate and the member Subject its Observation
// is on.
type Candidate struct {
	Member []byte
	autoreconcile.Candidate
}

// Option changes how Build reads the entity.
type Option func(*options)

type options struct{ excluding [][]byte }

// Excluding builds the profile as if member's Observations were not in the
// group: each Property member recorded is reconciled again without them.
func Excluding(member []byte) Option

// Build returns the entity's profile: on each Property, the displayed values
// as positives (only the concluded value when there is one) and the values a
// stronger negative denied as negatives; outvoted, weak and provisional values
// are left out, and a Property with nothing left is absent. value is the rules
// the reconciler runs under: pass matchregistry.Default().Value, as the cache
// does. metas give each Property's value type and cardinality for Excluding.
// The error is Reconcile's, and only Excluding runs it.
func Build(e Entity, metas subjectcompare.Metas, value valuecompare.Rules, opts ...Option) (subjectcompare.Values, error)
```

`Build`: apply the options; for each key of `e.Properties` (map order is fine: the output is a map and each property is independent), take `res := p.Result`, re-run per D-6 when a left-out member has a candidate on it, translate `res` per D-4 with `value` (D-5), and set the key only when the slice is not empty. `out` starts as `subjectcompare.Values{}`, so the zero `Entity` gives an empty, non-nil map (AC-1.12).

### FFI

None — no caller yet.

### macOS

None — no app change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — The package, the translation and the guard**
Files: `core/subjectprofile/subjectprofile.go` *(new)*, `core/subjectprofile/helpers_test.go` *(new)*, `core/subjectprofile/subjectprofile_test.go` *(new)*, `core/matchregistry/guard_test.go`. Skill: `add-match-rule` ("Add a layer package").
Do: first run `go list -deps -test ./core/autoreconcile | grep 'provenencia/core/database'` and confirm it prints nothing (S10-66 landed). Then the package doc, every type and function of §6 with their doc comments, `Build` with D-4 and D-5 (with `Excluding` recorded in `options` but not yet read), the `layerPackages` row (D-9), the helpers (D-10), and `TestBuild` with the US-1 rows (T-1).
Check: `go test ./core/subjectprofile/...` passes with no tags; `go test ./core/matchregistry/ -run TestLayerPackages -v` passes and lists `core/subjectprofile/pure` and `core/subjectprofile/numbers`.

**STEP-2 — Leave-one-out**
Files: `core/subjectprofile/subjectprofile.go` and `core/subjectprofile/subjectprofile_test.go`, both (new) in STEP-1.
Do: D-6 and D-7 in `Build`. Add the US-2 rows to `TestBuild` (T-2).
Check: `go test ./core/subjectprofile/...` passes; `go vet ./core/subjectprofile/...` is clean.

**STEP-3 — Entity compare**
Files: `core/subjectprofile/entity_test.go` *(new)*.
Do: `TestEntityCompare` with the US-3 rows (T-3): build the profile with `Build`, then call `subjectcompare.Compare(kind, profile, subject, subjectcomparetest.Metas(), matchregistry.Default().Subject, matchregistry.Default().Value)`; AC-3.1, AC-3.3, AC-3.4 and AC-3.5 check the `sex_at_birth` `PropertyResult`'s outcome and cause in `check`.
Check: `go test ./core/subjectprofile/... -run TestEntityCompare -v` passes.

**STEP-4 — Invariants**
Files: `core/subjectprofile/invariants_test.go` *(new)*.
Do: `TestInvariants` per D-11 (T-4).
Check: `go test ./core/subjectprofile/... -run TestInvariants -v` passes; then `go test ./core/subjectprofile/... ./core/subjectcompare/... ./core/propertycompare/... ./core/valuecompare/... ./core/autoreconcile/... ./core/matchregistry/...` passes with no tags.

**STEP-5 — Docs**
Files: `.cursor/rules/match-rules.mdc`, `.agents/skills/add-match-rule/SKILL.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching-model.md`, `docs/deployment-plan/spike-10-followups/prs/S10-66-pure-depiction-value.md` (header only), `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do:
- `match-rules.mdc` (D-12): in `globs`, insert `core/subjectprofile/**,` after `core/subjectcompare/**,`; the line "- Layer packages are pure: no `core/database`, no build tags; non-test files import no `core/matchregistry` or `core/autoreconcile`." becomes "- Layer packages are pure: no `core/database`, no build tags; non-test files import no `core/matchregistry`, and only `core/subjectprofile` imports `core/autoreconcile`, to turn its output into a profile."; the line "- Check: `go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/matchregistry/...` with no tags." becomes "- Check: the commands in the skill's Check section."
- `add-match-rule` SKILL.md:
  - Front matter `description`: "(core/valuecompare, core/propertycompare, core/subjectcompare, core/matchregistry)" becomes "(core/valuecompare, core/propertycompare, core/subjectcompare, core/subjectprofile, core/matchregistry)", and in its "Use when" sentence, "an entity profile rule," is inserted after "a Subject property weight, pool or period,"; rewrap the paragraph.
  - The "**Layers are pure.**" bullet's last clause "and never imports `core/matchregistry` or `core/autoreconcile`." becomes "and never imports `core/matchregistry`; only `core/subjectprofile` imports `core/autoreconcile`, to turn its output into a profile."
  - A bullet *(new)* after the Layer 3 bullet: "**The profile builder (`core/subjectprofile`) turns the reconciler's output into a Subject.** `Build` keeps a Property's displayed values as positives (only the concluded value when there is one) and its denied values as negatives, and leaves everything else out; layer 3 compares the profile like any Subject, so entity compare is `Build` then `subjectcompare.Compare`. `Excluding(member)` reconciles each Property the member recorded again without its candidates. Callers pass `matchregistry.Default().Value`, the rules the cache was built with. The builder has no registry section."
  - "Check": the command becomes `go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/subjectprofile/... ./core/matchregistry/...`.
  - "Do not": "Import `core/autoreconcile` from a layer package." becomes "Import `core/autoreconcile` from a layer package other than `core/subjectprofile`."
- `comparison-layers.md`: in §The layers, row 4's Status cell "Decided in shape (§ Entities): subject comparison against the entity's reconciled conclusion" becomes "Built in S11-11 (§ Entities): subject comparison against the entity's reconciled conclusion". At the end of §Entities: compare against the conclusion, after the paragraph that starts "**The profile builder is its own package:**", a paragraph *(new)*: "**Built in S11-11.** `subjectprofile.Build(entity, metas, value, opts...)` reads each Property's reconciler `Result`: values with reason `kept` become positives, values with reason `denied` negatives, and the rest are left out. When the result is concluded, only the concluded value is positive. A Property with nothing left is absent. The output is `subjectcompare.Values`, so entity compare is `Build` then `subjectcompare.Compare`, and entity against entity is `Compare` on two profiles. `Excluding(member)` runs `autoreconcile.Reconcile` again on each Property the member recorded, without that member's candidates and with the concluded value kept. A negative Observation that denies no positive value leaves no reconciled value, so it is not in the profile."
- `matching-model.md`: §The stack, the paragraph "The subject profile builder sits beside layer 4. It turns the reconciler's conclusion into a Subject-shaped profile and compares nothing." becomes "The subject profile builder sits beside layer 4. It turns the reconciler's conclusion into a Subject-shaped profile and compares nothing: displayed values are positive (only the concluded value when there is one), denied values negative, and outvoted, weak and provisional values are left out." §Glossary, the **profile** row becomes "| **profile** | A Subject-shaped set of values built from an entity's reconciled conclusion: displayed values positive, denied values negative, nothing else |".
- S10-66 plan header: "**Blocks:** S11-19 (Spike 11 map: the claim evidence query carries likeness lines through the pure matching stack)" becomes "**Blocks:** S11-11 (the profile builder imports `core/autoreconcile` from a layer package), S11-19 (Spike 11 map: the claim evidence query carries likeness lines through the pure matching stack)".
- `completed.md`, newest first: "## S11-11 — Subject profile and entity compare" / "**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N). The profile builder exists: `subjectprofile.Build` turns an entity's reconciled values into a Subject-shaped profile, with displayed values as positives, denied values as negatives and nothing else, so entity compare is `Build` then `subjectcompare.Compare`. `Excluding(member)` reconciles again without that member's Observations, for leave-one-out checks. No caller yet; nothing changes for researchers." with the date and PR number filled in.
- Map row status `Landed #N`; this plan's Status `Landed in [#N](…) (YYYY-MM-DD)` and a **Landed** note under the header listing each difference from the plan (or "None.").
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-11-subject-profile.md` reports no errors; `rg -n 'core/autoreconcile' .cursor/rules/match-rules.mdc .agents/skills/add-match-rule/SKILL.md` shows only lines naming `core/subjectprofile` as the exception and the reconciler bullet; `rg -n 'go test' .cursor/rules/match-rules.mdc` prints nothing; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Add the subject profile builder, which turns an entity's conclusion into a Subject."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/subjectprofile/subjectprofile_test.go` `TestBuild` *(new)* | Rows AC-1.1 – AC-1.12, inputs and expectations as in §2 | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, AC-1.12, D-3, D-4, D-5 |
| T-2 | Go | same, `TestBuild` | Rows AC-2.1 – AC-2.10 | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, AC-2.10, D-6, D-7 |
| T-3 | Go | `core/subjectprofile/entity_test.go` `TestEntityCompare` *(new)* | Rows AC-3.1 – AC-3.9 | AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.6, AC-3.7, AC-3.8, AC-3.9, D-1 |
| T-4 | Go | `core/subjectprofile/invariants_test.go` `TestInvariants` *(new)* | 500 generated entities per AC (D-11) → AC-4.1 – AC-4.4 hold | AC-4.1, AC-4.2, AC-4.3, AC-4.4, D-11 |
| T-5 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` | New row → `core/subjectprofile` pure (through `core/autoreconcile`) and free of numbers | D-2, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestLayerPackages` (a stray import, or S10-66 not landed), `core/autoreconcile`'s tests (must not change: this PR does not touch the package).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] go test ./core/subjectprofile/... ./core/subjectcompare/... ./core/propertycompare/... ./core/valuecompare/... ./core/autoreconcile/... ./core/matchregistry/... passes with no build tags
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without macOS every check but xcode-test passes, and CI runs the Xcode tests on the PR (this PR changes no Swift)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored; core/autoreconcile, core/subjectcompare, core/match and core/database are unchanged
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none (no copy)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: comparison-layers.md §The layers and §Entities, matching-model.md (stack paragraph, glossary), add-match-rule skill, match-rules.mdc, S10-66 plan header (STEP-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — nothing reaches it until S11-13. The reviewer runs `go test ./core/subjectprofile/... -v` with no build tags and reads the `TestEntityCompare` subtests `AC-3.1 …`, `AC-3.3 …` and `AC-3.6 …`: these are the map's Check (a Subject recording an outvoted value conflicts; one recording a denied value conflicts via the negative; leave-one-out drops that member's support). Then `go test ./core/matchregistry/ -run TestLayerPackages -v` to see `core/subjectprofile` pass the purity and numbers guard.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S10-66 has not landed, so the guard fails through `autoreconcile.Value.Depiction` | STEP-1's first command; the plan stays Blocked until S10-66 lands |
| S11-10 lands with names other than §4's *(S11-10)* rows | Implementer step 2 and the first stop condition |
| A reconciler change after the planning base moves an outcome §2 reads (majority, denial, weak, concluded order) | §2's outcomes were read from `Reconcile` at `3bd1348`; a drift is a stop condition, not a test edit |
| Map iteration order leaks into the output | The output is a map and each property is built alone (§6); `labels` sorts keys |

**Stop and report if:**

- `go list -deps -test ./core/autoreconcile | grep 'provenencia/core/database'` prints anything at your base.
- `main` differs from §4's *(S11-10)* rows: `subjectcompare.Values`, `Metas`, `Compare`'s signature or `Result`, `propertycomparetest`'s builders, or `subjectcomparetest.Metas()`'s entries. Report each difference.
- `autoreconcile.Reconcile` at your base gives a reason or an order different from §2 for an AC's candidates (for example AC-1.1's female is not `outvoted`, AC-1.2's female is not `denied`, AC-2.1 is not two kept values, or AC-1.7's rank 1 is not the concluded female): report the candidates and the result; do not change the AC.
- `subjectcompare.Compare` gives a sim or ev different from §2 for a US-3 row: report the row and the output.
- `core/subjectprofile` cannot pass `TestLayerPackages` with only the imports D-2 lists.
- An invariant of US-4 fails under D-4 – D-6 as written: report the generated case (seed, index, the entity's candidates) and which invariant.
- `core/subjectprofile` already exists.
- Merging `main` conflicts in a paragraph STEP-5 rewrites, so keeping both sides' text is not possible.

## 12. Open questions

None.

## 13. PR

**Title:** S11-11 — Subject profile and entity compare
**Body:** objective in one sentence; what shipped against each US; AC evidence (test names); deviations from the plan; manual check (the `go test` runs of §10); `Plan: docs/deployment-plan/spike-11/prs/S11-11-subject-profile.md`.
