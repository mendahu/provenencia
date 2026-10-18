# S11-09 — Property compare

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 3 — Property, Subject and entity compare
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a new pure Go package with no caller yet; nothing reaches the app)
**Board:** —
**Depends on:** S11-05 (landed [#451](https://github.com/mendahu/provenencia/pull/451)). Planned assuming S11-08 ([plan](S11-08-swift-normalization.md)), in progress at planning time, lands first. S11-08 is Swift-only, but it also edits [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) and [`comparison-layers.md`](../comparison-layers.md) in paragraphs STEP-8 does not touch. Branch from `main` after S11-08 merges. If it has not merged when you start, branch from `main` anyway; when it merges, merge `main` into this branch and keep both sides' text
**Blocks:** S11-10
**Branch:** `s11-09-property-compare`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships, as Spike 10 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Layer 2](../comparison-layers.md#layer-2-property-comparison) (contract, steps, outcome table, output, tests); [rules for every layer](../comparison-layers.md#rules-for-every-layer); [entities](../comparison-layers.md#entities-compare-against-the-conclusion) (why support and standing never reach layer 2); [`docs/matching-model.md`](../../../matching-model.md) stack table and glossary; working rules "Matching is a stack of pure layers", "Pure means the package builds and tests with no build tags", "Property compare holds only what a Subject says about itself", "Build, then migrate", "One guard, not one per package"
**Planned:** 2026-10-18 against `main` @ `d49fde7`, with S11-08 assumed landed on top
**Review:** [`S11-09-property-compare.review.md`](S11-09-property-compare.review.md) — Fits with changes / Ready with changes → fixed, 1 round

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

Layer 2 of matching exists: `core/propertycompare` decides whether two Subjects agree on one property from every value each one records for it. It drops values that are not evidence (a neutral term such as sex `unknown`), runs value compare on every pair, picks the pair that speaks, lets a negative Observation contradict a value it matches, treats a difference on a multiple-value property as no conflict, and returns one outcome (agree, partial, conflict, unknown) with the speaking pair and the excluded values. It is pure, holds no number, and reads no support or standing.

**Why now.** Slice 3 builds Subject compare (S11-10) on this layer; S11-10 cannot start without it. The rule "Build, then migrate" puts the layer in its own PR with its own tests before any caller moves.

**What stays.** `core/match` (`pairValues`, `Evaluate`, `OutcomeOf`, `PropertyMeta`) and its tests are untouched: S11-13 moves Promote onto the layers and deletes `core/match`. `promotealign.pairExhibits` stays until S11-19. `core/valuecompare` changes only where its relation order lives (D-16); every comparer returns what it returns today. No catalog, wire, Swift or registry value changes. No behavior a researcher can see changes.

## 2. User stories

All values below are compared under `matchregistry.Default().Value` and built with the D-12 helpers: names with `namevaluetest.Western` (`"J. Robins"` is given `J.`, surname `Robins`), dates with `datevaluetest` (`1817` is `datevaluetest.YMD(1817, 0, 0)`, `ABT 1817` is `datevaluetest.About(1817, 0, 0)`, `spring` is `datevaluetest.Phrase("spring")`). "¬x" is a negative value x. "Pair" is `Result.Speaking`; its relation and similarity are `Speaking.Result`. Similarities are value compare's at the planning base (checked by running `valuecompare.Compare`).

**US-1.** As the engine (Subject compare, S11-10), when I compare two Subjects' values for one property, I get one outcome read from the relation of the pair that speaks, so a later layer turns it into points without a per-type floor.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | term, single: `{male}` / `{male}` | agree, cause `relation`; Pair `male` / `male`, `Same`, 1 |
| AC-1.2 | term, single: `{male}` / `{female}` | conflict, cause `relation`; Pair `male` / `female`, `Different`, 0 |
| AC-1.3 | name, single: `{James Robins}` / `{James Robbins}` | partial, cause `relation`; Pair `Near`, 0.9143 (±0.001), detail cause `spelling` |
| AC-1.4 | name, single: `{J. Robins}` / `{James Robins}` | partial, cause `relation`; Pair `Contains`, 0.8, `Speaking.Detail.Broader` = `a` |
| AC-1.5 | text, single: `{Montréal}` / `{Montreal}` | agree, cause `relation`; Pair `Equivalent`, 0.95 |
| AC-1.6 | text, multiple: `{York}` / `{London}` | unknown, cause `cardinality`; Pair `York` / `London`, `Different` |
| AC-1.7 | text, multiple: `{York, Eboracum}` / `{Eboracum}` | agree, cause `relation`; Pair `Eboracum` / `Eboracum`, `Same` |
| AC-1.8 | date, single: `{spring}` (phrase only) / `{1817}` | unknown, cause `relation`; Pair `Unknown`, detail cause `phrase_only` |
| AC-1.9 | term, single: `{}` / `{male}`; and `{}` / `{}` | unknown, cause `no_values`; `Speaking` nil; Excluded empty |
| AC-1.10 | meta value type `subject`, single: one value on each side | unknown, cause `relation`; Pair `Unknown`, detail cause `unsupported` |
| AC-1.11 | meta cardinality `""` (empty), term: `{male}` / `{female}` | conflict, cause `relation` (empty reads as single) |
| AC-1.12 | text, single, values built with `Type` left empty and meta value type `text`: `{York}` / `{york}` | agree, `Same`; `Speaking.A.Type` and `Speaking.B.Type` are `text` |

**US-2.** As the engine, when a side records several values for a property, the best pair speaks, and the same pair speaks whatever order the values arrive in.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | name, single: `{Mary Smith, James Robins}` / `{James Robins}` | agree; Pair `James Robins` / `James Robins`, `Same` |
| AC-2.2 | name, single: `{James Robins}` / `{J. Robins, James Robbins}` | partial; Pair `James Robins` / `James Robbins`, `Near` 0.9143 (beats `J. Robins`, `Contains` 0.8) |
| AC-2.3 | name, single: `{James Robbins}` / `{J. Robins, James Robins}` | partial; Pair `James Robbins` / `James Robins`, `Near` 0.9143 (`J. Robins` is `Different`, `below_minimum`) |
| AC-2.4 | date, single: `{spring, 1817}` / `{1830}` | conflict; Pair `1817` / `1830`, `Different`, detail cause `beyond_tolerance` (a `Different` pair speaks before an `Unknown` one) |
| AC-2.5 | name, single: `{Anna Smith, Bea Smith, Cora Smith}` / `{Zed Jones}`, in each of the six orders of side A | conflict; Pair A is `Anna Smith` every time (lowest `Same` key on a tie) |
| AC-2.6 | term, single: `{male}` with ref `0x07` / `{male}` with ref `0x09` | `Speaking.A.Ref` and `Speaking.Detail.RefA` are `0x07`; `Speaking.B.Ref` and `Speaking.Detail.RefB` are `0x09` |

**US-3.** As the engine, a negative Observation contradicts a value on the other side that it matches, on any property, and nothing on its own side rescues the match.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | term, single: `{¬male}` / `{male}` | conflict, cause `negative`; Pair A `male` with `Negative` true, Pair B `male`, `Same` |
| AC-3.2 | term, single: `{female, ¬male}` / `{male}` | conflict, cause `negative`; Pair A is the negative `male` |
| AC-3.3 | name, single: `{¬J. Robins}` / `{James Robins}` | conflict, cause `negative`; Pair `Contains` |
| AC-3.4 | text, single: `{¬Montréal}` / `{Montreal}` | conflict, cause `negative`; Pair `Equivalent` |
| AC-3.5 | name, single: `{¬James Robbins}` / `{James Robins}` | unknown, cause `no_values` (`Near` does not contradict; A has no positive) |
| AC-3.6 | term, single: `{¬female}` / `{male}` | unknown, cause `no_values` (`Different` does not contradict) |
| AC-3.7 | term, single: `{¬male}` / `{¬male}` | unknown, cause `no_values` (two negatives never meet) |
| AC-3.8 | term, single: `{male}` / `{¬male, female}` | conflict, cause `negative`; Pair B is the negative `male` |
| AC-3.9 | text, multiple: `{¬Eboracum}` / `{York, Eboracum}` | conflict, cause `negative`; Pair `Eboracum` / `Eboracum` |

**US-4.** As the engine, a neutral value is not evidence: it is dropped before any comparison and listed with its reason.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | term, single: `{unknown (neutral)}` / `{male}` | unknown, cause `no_values`; Excluded = one entry: side `a`, value `unknown`, reason `neutral` |
| AC-4.2 | term, single: `{unknown (neutral), female}` / `{male}` | conflict, cause `relation`, Pair `female` / `male`; Excluded = the `unknown` entry |
| AC-4.3 | term, single: `{¬unknown (neutral)}` / `{unknown (neutral)}` | unknown, cause `no_values`; Excluded = side `a` entry, then side `b` entry |

**US-5.** As a contributor, the layer's invariants hold on generated cases, so a later change that breaks one fails a test.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | Any generated case; add a neutral value to either side | Outcome, cause and speaking pair (relation, similarity, both `Same` keys) unchanged |
| AC-5.2 | Any generated case; add a negative value to either side | Outcome is not more favorable (agree > partial > unknown > conflict) |
| AC-5.3 | Any generated case; shuffle the values on either side | Outcome, cause and speaking pair (relation, similarity, both `Same` keys) unchanged |
| AC-5.4 | Any generated case; swap the sides | Outcome, cause, speaking pair relation and similarity unchanged |
| AC-5.5 | Any generated multiple-value case with no negative; add any positive value to either side | Outcome is never conflict |

## 3. Scope

| In | Covered by |
| --- | --- |
| `propertymeta.Meta` (D-2) | STEP-1, AC-1.10, AC-1.11 |
| `valuecompare.Stronger` and one relation order (D-16) | STEP-1, T-7 |
| `core/datevalue/datevaluetest` (D-12) | STEP-2 |
| `core/propertycompare`: types, `Compare`, outcome table, best pair, negatives, neutral values, cardinality (D-1, D-3 – D-11) | STEP-3 – STEP-6, US-1 – US-4 |
| Exact table and invariant tests (D-12, D-15) | STEP-4 – STEP-7, all ACs |
| Purity and numbers guard rows; `match-rules.mdc` glob, text and check line (D-13) | STEP-2, STEP-3 |
| Docs the layer makes true or false (D-14) | STEP-8 |

| Out | Home |
| --- | --- |
| Any caller: Promote scoring, `match.Evaluate`, `pairValues`, `alikeNew` | S11-13 |
| `promotealign.pairExhibits` onto the speaking pair | S11-19 |
| Deleting `core/match` (`pairs.go`, `pairs_test.go`, `match.PropertyMeta`, `match.Outcome`) | S11-13 |
| Support counts, leaders, standing, denied values | S11-11 (`subjectprofile.Build` turns them into positives and negatives) |
| Pairs compared together (sex with gender; `date` with `start_date` / `end_date`) and any weight or points | S11-10 |
| A `property` section in `core/matchregistry` | Not planned — layer 2 has no tunable (D-3) |
| Provisional or outvoted values as non-evidence | Not planned — the profile is strict and leaves them out before layer 2 ([entities](../comparison-layers.md#entities-compare-against-the-conclusion)) |
| Moving the other hand-written date builders (`core/autoreconcile/dates_test.go` `ymd`, `core/database/conclusionheaders/placechains_internal_test.go` `ymd`, `placechains_test.go` `year`, `core/database/promote/batch_test.go` `yearDate`) onto `datevaluetest` | Not planned — each moves when its file is next edited |
| Wording outcomes or causes in the UI | S11-20 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Layer 2 package | `core/propertycompare` | Does not exist |
| Property facts | `core/propertymeta/propertymeta.go` | Package doc says it "holds the property value-type and cardinality constants". Only constants: `ValueType*` and `CardinalitySingle`, `CardinalityMultiple`. No struct. `comparison-layers.md` §Layer 2 already names `propertymeta.Meta` |
| Old pair selection | `core/match/pairs.go` `pairValues`, `weigh`, `refutedBy`, `plainPair`, `compatible` | Mixes Subject-level rules (negatives R1, multiple cardinality, neutral via `votes`) with entity rules (support, leaders, denied R2). Uses `autoreconcile.Compatible` to decide what a negative matches. Stays until S11-13 |
| Old outcome | `core/match/pair.go` `Outcome` (`agree`, `partial`, `conflict`, `unknown`), `OutcomeOf`, `PropertyMeta{Property, ValueType, Cardinality}` (empty cardinality means single) | Outcome from a similarity number. The outcome strings reach Swift through `api/ffi/handlers/promote_align.go` (`macos/App/Platform/L10n.swift` `case "partial"`) |
| Relation order | `core/valuecompare/name.go` `relationOrder` (`[]Relation{Same, Equivalent, Contains, Near}`), read by `weakestOf` with `slices.Index` | The one relation order today, private to names, and missing `Overlaps`, `Different`, `Unknown` |
| Date fixtures | `core/valuecompare/date_test.go` `optInt`, `ymd`, `about`, `phrase`, `dp` (package `valuecompare_test`) | Not importable; `core/datevalue/datevaluetest` does not exist |
| Layer 1 | `core/valuecompare/valuecompare.go` `Compare(a, b Value, rules Rules) (Result, Detail)`, `Key(v Value, level Relation, rules Rules) (string, bool)`, `Value`, `Result`, `Detail` (`RefA`, `RefB`, `Broader`, `Cause`), `Relation` constants, `Side` (`SideA`, `SideB`) | `Compare` returns `Unknown` / `unsupported` for two types or a type with no comparer; `Key` has no key for an empty value, a phrase-only date or a type with no comparer |
| Registry | `core/matchregistry/registry.go` `Registry{Value valuecompare.Rules}`, `Default()` | One section; no change in this PR |
| Guard | `core/matchregistry/guard_test.go` `layerPackages`, `TestLayerPackages` | Lists `core/valuecompare` (numbers true) and the pure value packages |
| Knob inventory guard | `core/matchregistry/inventory_test.go` `notYetBuilt`, `TestKnobInventoryMatchesRegistry` | `subject.` and up are not built; this PR adds no field, so it changes nothing here |
| Editor rule | `.cursor/rules/match-rules.mdc` | `globs: core/valuecompare/**,core/matchregistry/**,docs/matching-model.md`; says layer packages import no `core/matchregistry` (the `valuecompare` tests do); Check line names `valuecompare` and `matchregistry` only |
| Test fixtures | `core/namevalue/namevaluetest/namevaluetest.go` `Western(form string) *namevalue.Value` | Pure; given names then one surname |
| Tests to port | `core/match/pairs_test.go` `TestPairValues` ("AC-1.7", "only neutral", "empty probe"), `TestPairValuesNegatives` (AC-2.1, AC-2.3, "compatible folds count", "a candidate negative against a probe leader", "negative alone…"), `TestPairValuesMultiple`, `TestPairValuesOrderIndependent` | The Subject-level cases become this PR's exact table (US-1 – US-4); the support, leader, weak and denied cases are entity rules and go to S11-11 |
| Docs | `docs/deployment-plan/spike-11/comparison-layers.md` §Layer 2 "What it does, in order", Contract and Output; `docs/matching-model.md` §Glossary (**cause** row); `.agents/skills/add-match-rule/SKILL.md` (description, "Add a relation, cause or unit code", "Add a layer package", "Check") | The step list puts negatives after choosing the pair; the contract sketch names `Side`, `PropertyMeta`, `PropertyResult`, `SpokenPair`, `ExcludedValue`, `OutcomeCause`; the **cause** row names one vocabulary; the skill describes layer 1 only |

**Precedent.** [#441](https://github.com/mendahu/provenencia/pull/441) (S11-02) built `core/valuecompare` as a pure layer with no caller: copy its shape (a package doc comment naming the layer and its purity, constants as typed strings, an external `_test` package that reads `matchregistry.Default().Value`, a `layerPackages` row, the generated-case style of `core/valuecompare/contract_test.go` with a fixed seed). Do not copy its registry section or knob rows: this layer has none (D-3).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Package and entry.** `core/propertycompare`, one file `propertycompare.go` plus tests. `func Compare(a, b []Value, meta propertymeta.Meta, rules valuecompare.Rules) Result`. Types (all *new*): `Value`, `Result`, `SpeakingPair`, `Excluded`, `Outcome`, `Cause`, `ExcludeReason` (sketch in §6); the result's field is `Speaking` | `propertycompare.Result` reads cleanly where `PropertyResult` stutters. `SpeakingPair` is the glossary's **speaking pair**, and leaves the word "pair" to S11-10's pairs compared together. A side is a plain slice: it carries nothing but its values | A `Side` type — a slice with no behavior, and the name clashes in meaning with `valuecompare.Side`; `PropertyResult`, `ExcludedValue` — stutter; `Pair` — collides with the glossary's **pair (compared together)** |
| D-2 | **`propertymeta.Meta`** *(new)*: `type Meta struct { ValueType string; Cardinality string }`, doc comments as in §6. The package doc becomes "Package propertymeta holds what matching and the catalog share about a property: the value-type and cardinality constants and Meta. It is a pure value package: queries live in core/database/properties." | Pure, already named by the design doc; S11-10's registry and S11-13's loader fill it | `match.PropertyMeta` — impure package deleted in S11-13, and carries a `Property` the layer never reads |
| D-3 | **No registry section.** `Compare` takes `valuecompare.Rules` (the registry's `value` section) and passes it to layer 1. The package holds no numeric literal other than 0 and 1 | Layer 2 has no tunable: the outcome comes from the relation, and layer 1's `Near` minimum replaced the floors (`comparison-layers.md` §Outcome from relation). `matching-model.md` already says layer 2 reads `value`, passed through | A `property` section — an empty struct no knob would fill |
| D-4 | **The property sets the value type.** Before comparing, `Compare` sets each value's `Type` to `meta.ValueType` on its own copy; the caller's `Type` is ignored. The values in `Pair` and `Excluded` carry the set type | A value of a property has that property's type; one place sets it, so no caller can pass a mismatch | Trusting the caller's `Type` — `meta.ValueType` would be unread, and a mismatch would surface as `Unknown` far from its cause |
| D-5 | **Step 1, non-evidence.** A value with `Neutral` true, positive or negative, is dropped and listed in `Excluded` as `{Side, Value, Reason: ExcludeNeutral}`, side A's in input order, then side B's. Only the flag is read; the layer does not check the value type | "Is this value evidence?" is a property decision (layer 1 table); callers set `Neutral` only on terms. Input order keeps `Excluded` stable without sorting | Dropping only positive neutrals — a neutral negative is no evidence either |
| D-6 | **Ranking, used for every "best".** A compared pair `x` speaks before `y` when, in order: (1) its relation's outcome comes earlier in `speakingOrder` *(new, unexported)* `= []Outcome{OutcomeAgree, OutcomePartial, OutcomeConflict, OutcomeUnknown}`, the outcome read by `outcomeOf` (D-8) as on a single-value property; (2) its similarity is higher; (3) `valuecompare.Stronger(x, y)` (D-16); (4) side A's `valuecompare.Key(…, Same, rules)` sorts lower, a value with no key after any value with a key; (5) the same for side B. Pairs still equal keep the first offered, offering A's values in input order and, for each, B's in input order. Positions are read with `slices.Index`, as `weakestOf` does, so no literal is written. **Refs are never read** | Outcome first, so an `Equivalent` (agree) beats a 0.96 `Near` (partial); a comparable `Different` speaks before an `Unknown`, so a phrase-only date never hides a real conflict; keys make the pair independent of input order without reading the caller's refs, which the glossary says the layers never read | Ordering by similarity alone — a high `Near` would outrank `Equivalent`; tie-breaking on refs (`match.order` does, on Observation ids) — reads an opaque ref; values with equal keys are the same value, and claim evidence collapses identical values anyway |
| D-7 | **Step 4, negatives, before the outcome.** After step 1, every pair of a negative value on one side and a positive value on the other is compared. A pair whose relation is `Same`, `Equivalent` or `Contains` (either broader side) contradicts: the best such pair by D-6 is `Speaking`, outcome conflict, cause `negative`, on any cardinality, whatever the positive pairs say. Two negatives are never compared. A negative that matches nothing does nothing | The layer 2 steps and outcome table; "a positive does not rescue the negative" (`pairs_test.go` AC-2.3); `Near` and `Overlaps` say the values differ, so a denial of one is not a denial of the other | `autoreconcile.Compatible` (today's R1 test) — the reconciler's grouping policy, not layer 2's, and a layer may not import `autoreconcile` |
| D-8 | **Steps 2, 3 and 6, the positive pairs.** With no contradiction, every positive A value is compared with every positive B value; the best by D-6 is `Speaking`. `outcomeOf(r valuecompare.Relation) Outcome` *(new, unexported)* is the one relation → outcome mapping: `Same`, `Equivalent` → agree; `Contains`, `Overlaps`, `Near` → partial; `Different` → conflict; `Unknown` → unknown. `Compare` then applies cardinality: conflict on a multiple-value property becomes unknown with cause `cardinality`. Cause is `relation` unless named. A side with no positive value: outcome unknown, cause `no_values`, `Speaking` nil | The outcome table in `comparison-layers.md` §Outcome from relation; an `Unknown` pair still names what could not be compared (a phrase-only date, an unsupported type) | Reporting no pair for `Unknown` — the evidence view could not say why |
| D-9 | **Cardinality.** A property is multiple only when `meta.Cardinality == propertymeta.CardinalityMultiple`; anything else, the empty string included, is single | Matches `match.Evaluate` (empty means single) | Rejecting an empty cardinality — every caller would have to default it |
| D-10 | **Codes.** `Outcome` strings are `agree`, `partial`, `conflict`, `unknown`, the same as `match.Outcome`. `Cause` strings: `relation`, `negative`, `no_values`, `cardinality`. `ExcludeReason`: `neutral` | Outcome strings already reach Swift; S11-13 swaps `match.Outcome` for these with no wire change. Codes are strings, as in `valuecompare` | New outcome words — a wire and copy change for nothing |
| D-11 | **Output carries values, not indexes.** `Speaking.A` and `Speaking.B` are the two `Value`s (side A's and side B's, whichever is negative), `Speaking.Result` and `Speaking.Detail` are layer 1's for `Compare(A, B)` in that order. `Excluded[i].Value` is the dropped value | The layer above words lines and reads refs without holding the input slices; `Detail.RefA` / `RefB` already echo refs | Indexes into the input — callers would need the slices to explain a result |
| D-12 | **Tests and fixtures.** External package `propertycompare_test`, table-driven (`.cursor/rules/go-tests.mdc`). `TestCompare` holds every US-1 – US-4 row; `TestInvariants` generates cases from a fixed seed (D-15) for AC-5.1 – AC-5.5. Rules: `matchregistry.Default().Value`. No build tags. **Dates:** `core/datevalue/datevaluetest` *(new)*, beside `namevaluetest`, with `YMD(y, m, d int) *datevalue.Value` (kind `point`, calendar `gregorian`, a 0 month or day left nil), `About(y, m, d int) *datevalue.Value` (`YMD` with qualifier `datevalue.QualifierABT`) and `Phrase(s string) *datevalue.Value` (kind `point`, only `Phrase` set). **Helpers** in `core/propertycompare/helpers_test.go` *(new)*: `term(key string) propertycompare.Value` (sets `Term`), `text(s string)` (sets `Text`), `name(form string)` (sets `Name: namevaluetest.Western(form)`, `NamePattern` empty), `date(d *datevalue.Value)` (sets `Date`), `neg(v) Value` (sets `Negative`), `neutral(v) Value` (sets `Neutral`), `ref(v, b byte) Value` (sets `Ref: []byte{b}`), and `label(v) string` (term key, text, name form, or `datevalue.CompactDisplay`). Helpers leave `Type` empty: `Compare` sets it (D-4) | The S11-02 test layer; `valuecompare` tests import `matchregistry` the same way; the purity guard walks test imports and all are pure. S11-10 – S11-12 need date fixtures too, and the `valuecompare` builders sit in an external test package no one can import | Porting `pairs_test.go` by moving it — `core/match` keeps its tests until S11-13; date builders private to this package — a sixth copy | The S11-02 test layer; `valuecompare` tests import `matchregistry` the same way; the purity guard walks test imports and both are pure | Porting `pairs_test.go` by moving it — `core/match` keeps its tests until S11-13 |
| D-13 | **Guard and editor rule.** `layerPackages` gains `{"core/propertycompare", true}` after `core/valuecompare`, and `{"core/datevalue/datevaluetest", false}` after `core/datevalue`. `.cursor/rules/match-rules.mdc`: `globs` gains `core/propertycompare/**`; the first bullet becomes "Layer packages are pure: no `core/database`, no build tags; non-test files import no `core/matchregistry` or `core/autoreconcile`."; the Check line becomes "`go test ./core/valuecompare/... ./core/propertycompare/... ./core/matchregistry/...` with no tags." | The one guard keeps the layer pure and free of numbers (working rule "One guard, not one per package"); the editor rule loads for the new layer | A package-local guard — a second mechanism |
| D-14 | **Docs.** `comparison-layers.md` §Layer 2 step list, Contract and Output match the built layer; `matching-model.md` glossary gains **negative value** and **neutral value** and its **cause** row names both vocabularies; the `add-match-rule` skill covers layer 2 (description, a layer 2 bullet, codes per layer, the full layer package recipe, Check). Exact text in STEP-8 | The design doc and the model page are kept true as layers land; S11-10 extends the skill and needs the layer 2 rule there | Leaving the sketch names — S11-10's planner would read names that do not exist |
| D-15 | **Generator.** `rand.New(rand.NewSource(1))`; one subtest per invariant per value type (`AC-5.1/term`, `AC-5.1/name`, and so on for each AC and type), 500 cases each, over fixed pools: term `male`, `female`, `intersex`; name `James Robins`, `James Robbins`, `J. Robins`, `Mary Robins`, `Robins`, `Mary Smith`; text `York`, `york`, `Montréal`, `Montreal`, `York, Upper Canada`, `London`; date `1817`, `1818`, `1830`, `ABT 1817`, phrase `spring`. Each side draws 0 – 3 pool values; each drawn value is negative with chance 1 in 4 (`r.Intn(4) == 0`) and gets a distinct one-byte ref. Pool values are never neutral; the value AC-5.1 adds is a pool value of the subtest's type with `Neutral` set (for terms, `unknown`). Cardinality is single for even case indexes and multiple for odd, except AC-5.5 (always multiple, no negatives drawn). A failure reports `t.Errorf("case %d: a=%v b=%v: …", i, labels(a), labels(b))` with the seed fixed at 1 | Fixed pools cover every relation each comparer returns; a fixed seed makes a failure reproducible, as `contract_test.go` does | Random strings — rarely relate, so the invariants would be tested mostly on `Different` |
| D-16 | **One relation order.** `relationOrder` moves from `name.go` to `valuecompare.go`, below the `Relation` constants, and lists all seven, strongest first: `Same, Equivalent, Contains, Overlaps, Near, Different, Unknown`. `func Stronger(a, b Relation) bool` *(new, exported)*: `slices.Index(relationOrder, a) < slices.Index(relationOrder, b)`. `weakestOf` keeps its `slices.Index(relationOrder, …)` comparison unchanged | One place says which relation is stronger; layer 2 ranks with it and S11-10 and S11-19 read the same. Name relations keep their relative order, so `weakestOf` returns what it does today | A second order in `propertycompare` — two lists to keep in step; an exported slice — callers could change it | Fixed pools cover every relation each comparer returns; a fixed seed makes a failure reproducible, as `contract_test.go` does | Random strings — rarely relate, so the invariants would be tested mostly on `Different` |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/propertycompare` *(new)* | package | `core/match` `pairs.go` | `match` is impure (imports `autoreconcile`), mixes entity rules in, and is deleted in S11-13 ("Build, then migrate") |
| `propertycompare.Compare` *(new)* | func | `match.pairValues` | Replaces it in S11-13; it reads no support, leaders or denied (working rule "Property compare holds only what a Subject says about itself") |
| `propertycompare.Value` *(new)* | type | `match.Value`, `valuecompare.Value` | Embeds `valuecompare.Value` and adds only `Negative`, `Neutral`; `match.Value` carries reconciler fields that must never reach layer 2 |
| `propertycompare.Result`, `SpeakingPair`, `Excluded` *(new)* | type | `match.PairResult`, `match.PropertyComparison` | Those carry a weighted similarity and `Comparable`; layer 2 reports an outcome, the layer 1 result and detail, and the excluded values (§Output) |
| `propertycompare.Outcome` *(new)* | type | `match.Outcome` | Same strings (D-10); `match.Outcome` goes in S11-13 |
| `propertycompare.Cause`, `ExcludeReason` *(new)* | type | `valuecompare.Cause` | `valuecompare.Cause` names the rule behind a relation; these name why an outcome or an exclusion happened, a different vocabulary at a different layer |
| `propertymeta.Meta` *(new)* | type | `match.PropertyMeta` | Pure, without `Property` (D-2) |
| ranking helper `better`, `speakingOrder`, `outcomeOf` *(new, unexported)* | func, var | `match.order`, `pairSet.before`, `match.OutcomeOf` | Those rank by support and Observation id and read the outcome from a number; D-6 and D-8 read the relation, with the relation order from `valuecompare.Stronger` |
| `valuecompare.Stronger` *(new)* | func | `valuecompare` `relationOrder` (`name.go`) | Extends and exports the existing order (D-16) instead of adding one |
| `core/datevalue/datevaluetest` *(new)* | package | `core/namevalue/namevaluetest`; `core/valuecompare/date_test.go` `ymd`, `about`, `phrase` | Same shape as `namevaluetest`; the `valuecompare` builders cannot be imported (D-12) |
| `core/propertycompare/helpers_test.go` *(new)* | helper | `core/valuecompare` `textValue`, `termValue` | Build `propertycompare.Value`, not `valuecompare.Value`; private test helpers per package, as in `valuecompare` |

**Builds on.** `valuecompare.Compare`, `Key` and the relation order (S11-02 – S11-04, S11-07), `valuecompare.Side` for `Excluded.Side`, `namevaluetest.Western`, the `layerPackages` guard, the `add-match-rule` skill.
**Deletes.** None — nothing becomes dead: `core/match` keeps its callers until S11-13, which deletes `pairs.go` and `pair.go`.
**Next adds.** S11-10 calls `Compare` once per property (and once per declared pair, pooling two properties' values into one side) and turns `Outcome` into points from the registry `subject` section. S11-11 builds profiles whose concluded values are positives and denied values negatives, so layer 2 needs no change. S11-13 replaces `match.PropertyMeta` with `propertymeta.Meta` and `match.Outcome` with `propertycompare.Outcome`. S11-19 reads `Speaking` and `Excluded` for evidence lines. S11-10 – S11-12 build dates with `datevaluetest`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/propertymeta` | `type Meta struct` *(new)* and the package doc (D-2) |
| `core/valuecompare` | `relationOrder` moves to `valuecompare.go` with all seven relations; `func Stronger(a, b Relation) bool` *(new)* (D-16) |
| `core/datevalue/datevaluetest` *(new)* | `YMD`, `About`, `Phrase` (D-12) |
| `core/propertycompare` *(new)* | `propertycompare.go`: the types below and `Compare` (D-1, D-3 – D-11) |
| `core/matchregistry` | `guard_test.go` `layerPackages` rows (D-13) |

```go
// core/propertymeta
// Meta is what matching reads about a property.
type Meta struct {
	ValueType   string // a ValueType* constant
	Cardinality string // a Cardinality* constant; empty reads as single
}

// core/propertycompare
// Package propertycompare is layer 2 of matching: whether two Subjects agree
// on one property, from every value each one records for it. Pure: no I/O,
// no catalog, no numbers; layer 1's rules come in as an argument.

// Value is one value a Subject records for the property. Type is ignored:
// Compare compares every value as the property's value type.
type Value struct {
	valuecompare.Value
	Negative bool // the Subject says the property does not have this value
	Neutral  bool // a term that is no evidence (sex "unknown")
}

type Outcome string // OutcomeAgree "agree", OutcomePartial "partial", OutcomeConflict "conflict", OutcomeUnknown "unknown"
type Cause string   // CauseRelation "relation", CauseNegative "negative", CauseNoValues "no_values", CauseCardinality "cardinality"
type ExcludeReason string // ExcludeNeutral "neutral"

type Result struct {
	Outcome  Outcome
	Cause    Cause
	Speaking *SpeakingPair // the pair that spoke; nil when a side has no positive value
	Excluded []Excluded    // values dropped as non-evidence, side A's then side B's
}

// SpeakingPair is the one pair of values that decided the outcome.
type SpeakingPair struct {
	A, B   Value // side A's and side B's value; one may be negative
	Result valuecompare.Result
	Detail valuecompare.Detail
}

type Excluded struct {
	Side   valuecompare.Side
	Value  Value
	Reason ExcludeReason
}

func Compare(a, b []Value, meta propertymeta.Meta, rules valuecompare.Rules) Result
```

`Compare`, in order: copy each side, setting `Type` (D-4); drop neutral values into `Excluded` (D-5); negatives (D-7); positive pairs and outcome (D-8, D-9). Each pair is compared once with `valuecompare.Compare(x.Value, y.Value, rules)`, side A's value first. The `Same` keys D-6 reads are computed once per value. `Excluded` is nil when nothing is dropped.

Errors: none — `Compare` cannot fail.

### FFI

None — no caller yet; the outcome strings already on the wire are unchanged (D-10).

### macOS

None — no app change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — One relation order; `propertymeta.Meta`**
Files: `core/valuecompare/valuecompare.go`, `core/valuecompare/name.go`, `core/valuecompare/valuecompare_test.go`, `core/propertymeta/propertymeta.go`. Skill: `add-match-rule`.
Do: D-16 (move and extend `relationOrder`, add `Stronger`); add `TestStronger` (T-7). Add `Meta` and the package doc (D-2).
Check: `go test ./core/valuecompare/... ./core/matchregistry/...` passes with no tags (`TestComparerContract` and the name tables unchanged).

**STEP-2 — Date fixtures**
Files: `core/datevalue/datevaluetest/datevaluetest.go` *(new)*, `core/matchregistry/guard_test.go`.
Do: the package per D-12, with a package doc in the words of `namevaluetest`'s ("Package datevaluetest builds date values for test fixtures."); its `layerPackages` row (D-13).
Check: `go test ./core/matchregistry/ -run TestLayerPackages -v` passes and lists `core/datevalue/datevaluetest/pure`.

**STEP-3 — The package skeleton and the guard**
Files: `core/propertycompare/propertycompare.go` *(new)*, `core/matchregistry/guard_test.go`, `.cursor/rules/match-rules.mdc`. Skill: `add-match-rule` ("Add a layer package").
Do: package doc comment, the types and constants of §6, and `Compare` with D-4 and D-5 only (it returns unknown / `no_values` with `Excluded` filled). Add the `layerPackages` row and the `match-rules.mdc` changes (D-13).
Check: `go test ./core/matchregistry/ -run TestLayerPackages -v` passes and lists `core/propertycompare/pure` and `core/propertycompare/numbers`.

**STEP-4 — Positive pairs and the outcome**
Files: `core/propertycompare/propertycompare.go`, `core/propertycompare/helpers_test.go` *(new)*, `core/propertycompare/propertycompare_test.go` *(new)*.
Do: `outcomeOf`, `speakingOrder`, `better` (D-6), the positive pairs and cardinality (D-8, D-9). Write the helpers (D-12) and `TestCompare` with the US-1 and US-2 rows (T-1, T-2).
Check: `go test ./core/propertycompare/...` passes with no tags.

**STEP-5 — Negatives**
Files: `core/propertycompare/propertycompare.go` *(new)*, created in STEP-3, `core/propertycompare/propertycompare_test.go` *(new)*, created in STEP-4.
Do: D-7. Add the US-3 rows to `TestCompare` (T-3).
Check: `go test ./core/propertycompare/...` passes with no tags.

**STEP-6 — Neutral values in the table**
Files: `core/propertycompare/propertycompare_test.go` *(new)*, created in STEP-4.
Do: add the US-4 rows to `TestCompare` (T-4); fix `Compare` in this step if a row fails.
Check: `go test ./core/propertycompare/...` passes; `go vet ./core/propertycompare/...` is clean.

**STEP-7 — Invariants**
Files: `core/propertycompare/invariants_test.go` *(new)*.
Do: `TestInvariants` per D-15 (T-5).
Check: `go test ./core/propertycompare/... -run TestInvariants -v` passes; then `go test ./core/propertycompare/... ./core/matchregistry/... ./core/valuecompare/... ./core/datevalue/...` passes with no tags.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching-model.md`, `.agents/skills/add-match-rule/SKILL.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do:
- `comparison-layers.md` §Layer 2 **What it does, in order**: the six numbered items become: "1. **Drop non-evidence.** Remove neutral values, and record each one with its reason." / "2. **Apply negatives.** A negative value whose relation to a positive value on the other side is `Same`, `Equivalent` or `Contains` contradicts it: the outcome is conflict, whatever the positive values say." / "3. **Compare pairs.** Run layer 1 on every pair of positive values." / "4. **Choose the pair that speaks:** by outcome (agree, partial, conflict, unknown), then similarity, then the stronger relation, then each side's `Same` key. Step 2 picks among contradicting pairs the same way." / "5. **Apply cardinality.** On a multiple-value property, a difference is not a conflict." / "6. **Decide the outcome** from the relation (below)."
- Same section **Contract**: the code block becomes `func Compare(a, b []Value, meta propertymeta.Meta, rules valuecompare.Rules) Result // propertycompare.Compare`; the bullet "**`Side`** is one Subject's values for the property. Each value carries only what a Subject can say about itself:" becomes "**Each side** is one Subject's values for the property (`[]propertycompare.Value`). Each value carries only what a Subject can say about itself:"; the `PropertyMeta` bullet becomes "**`propertymeta.Meta`:** value type and cardinality. The property sets each value's type."
- Same section **Output**: the code block becomes the `Result`, `SpeakingPair` and `Excluded` structs of §6 (fields and comments only), followed by one sentence: "Built in S11-09. Outcome causes: `relation`, `negative`, `no_values`, `cardinality` (a `Different` pair on a multiple-value property). Refs are never read."
- `matching-model.md` §Glossary: the **cause** row becomes "| **cause** | A code naming the rule behind a result. Layer 1's relation causes (`exact`, `fold`, `alias`, `initial`, `spelling`, `precision`, `missing_part`, `too_long`, and so on) say why two values relate as they do; layer 2's outcome causes (`relation`, `negative`, `no_values`, `cardinality`) say why a property comparison came out as it did |". After the **outcome** row, two rows *(new)*: "| **negative value** | A value a Subject says the property does not have; it contradicts a value on the other side that is `Same`, `Equivalent` or `Contains` |" and "| **neutral value** | A term that is no evidence (sex `unknown`); layer 2 drops it before comparing |".
- `add-match-rule` SKILL.md:
  - Front matter `description`: "(core/valuecompare, core/matchregistry)" becomes "(core/valuecompare, core/propertycompare, core/matchregistry)".
  - "Rules that hold everywhere", a bullet *(new)* after "Pure helpers are shared, not copied.": "**Layer 2 (`core/propertycompare`) reads relations, never numbers.** `Compare` takes the `value` section and has no section of its own; the outcome comes from the speaking pair's relation through `outcomeOf`, and relations rank with `valuecompare.Stronger`, the one relation order. A value carries only `Negative` and `Neutral` beside its layer 1 value: support, standing and denied are turned into positives and negatives by the profile builder before they reach it."
  - "Add a relation, cause or unit code": the paragraph becomes "A relation, a relation cause or a unit is layer 1's: add the constant in `core/valuecompare/valuecompare.go`, a new relation also in `relationOrder` at its strength, and a cause to the cause list in `comparison-layers.md` §Detail. An outcome cause or an exclude reason is layer 2's: add it in `core/propertycompare/propertycompare.go` *(new)* and to `comparison-layers.md` §Layer 2 Output. Codes are strings; they reach the wire and UI copy later."
  - "Add a layer package": the paragraph becomes "Add a row to `layerPackages` in `core/matchregistry/guard_test.go`, with `numbers` true for a layer and false for a pure value or fixture package. Add the package's glob to `.cursor/rules/match-rules.mdc` and its path to the Check commands there and below. Give it a package doc naming its layer and saying it is pure. Test it from an external `_test` package with an exact table and a fixed-seed invariants test, building fixtures with `namevaluetest` and `datevaluetest` (`core/propertycompare` is the model)."
  - "Check": the command becomes `go test ./core/valuecompare/... ./core/propertycompare/... ./core/matchregistry/...`.
- Per the Definition of done: `completed.md` paragraph (newest first), map row status, this plan's status and **Landed** note.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-09-property-compare.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Add layer 2, property compare, as a pure package."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/propertycompare/propertycompare_test.go` `TestCompare` *(new)* | Rows AC-1.1 – AC-1.12, inputs and expectations as in §2 | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, AC-1.12, D-4, D-8, D-9, D-10 |
| T-2 | Go | same, `TestCompare` | Rows AC-2.1 – AC-2.6; AC-2.5 runs all six orders of side A | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, D-6, D-11 |
| T-3 | Go | same, `TestCompare` | Rows AC-3.1 – AC-3.9 | AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.6, AC-3.7, AC-3.8, AC-3.9, D-7 |
| T-4 | Go | same, `TestCompare` | Rows AC-4.1 – AC-4.3, `Excluded` compared field by field | AC-4.1, AC-4.2, AC-4.3, D-5 |
| T-5 | Go | `core/propertycompare/invariants_test.go` `TestInvariants` *(new)* | 500 generated cases per invariant per value type (D-15) → AC-5.1 – AC-5.5 hold | AC-5.1, AC-5.2, AC-5.3, AC-5.4, AC-5.5, D-6, D-15 |
| T-6 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` | New rows → `core/propertycompare` pure and free of numbers; `core/datevalue/datevaluetest` pure | D-3, D-13 |

| T-7 | Go | `core/valuecompare/valuecompare_test.go` `TestStronger` *(new)* | Each adjacent pair of `Same, Equivalent, Contains, Overlaps, Near, Different, Unknown` → `Stronger(x, y)` true, `Stronger(y, x)` false; `Stronger(Same, Same)` false | D-16 |

Each `TestCompare` row compares `Outcome`, `Cause`, `Speaking == nil`, and when a pair is expected its `A` and `B` labels (`label`, D-12), `A.Negative` / `B.Negative`, `Result.Relation`, `Result.Similarity` (within 0.001), the `Detail` field the row names, and `Excluded`.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestLayerPackages` (a literal 2 or a stray import), the `valuecompare` name tables (D-16 must not change `weakestOf`), `TestKnobInventoryMatchesRegistry` (must not change: this PR adds no registry field).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] go test ./core/propertycompare/... ./core/matchregistry/... ./core/valuecompare/... ./core/datevalue/... passes with no build tags
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without macOS every check but xcode-test passes, and CI runs the Xcode tests on the PR (this PR changes no Swift)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored; core/match is unchanged
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none (no copy)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: comparison-layers.md §Layer 2, matching-model.md glossary, add-match-rule skill, match-rules.mdc (STEP-3, STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — nothing reaches it until S11-13. The reviewer runs `go test ./core/propertycompare/... -v` with no build tags, reads the `TestCompare` row names against §2, and runs `go test ./core/matchregistry/ -run TestLayerPackages -v` to see the new row's `pure` and `numbers` subtests pass. This makes the slice 3 Check's "invariants and the outcome table in the layer 2 section hold in tests" (scope entry **Check**) true.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A ranking constant written as a literal (`2`, `3`) fails the numbers guard | D-6, D-16: positions come from `slices.Index` on declared lists |
| An invariant fails on a generated case because layer 1 is not symmetric for some pair | AC-5.4 checks relation and similarity only; layer 1's contract test already guarantees symmetry |
| AC similarities drift because layer 1 changed after the planning base | §2 values were read from `valuecompare.Compare` at `d49fde7`; a drift is a stop condition, not a test edit |

**Stop and report if:**

- `valuecompare.Compare` at your base gives a relation or similarity different from §2 for any AC pair (for example `James Robins` / `James Robbins` is not `Near` 0.9143): report the pair and the output; do not change the AC.
- `core/propertycompare` cannot pass `TestLayerPackages` without importing anything beyond the standard library, `core/valuecompare` and `core/propertymeta` (tests may add `core/matchregistry`, `core/datevalue`, `core/datevalue/datevaluetest` and `core/namevalue/namevaluetest`).
- Moving `relationOrder` (D-16) changes any `core/valuecompare` test result.
- An invariant of US-5 fails under D-6 – D-9 as written: report the generated case (seed, index, both sides) and which invariant.
- `core/propertymeta` already has a type named `Meta`, or `core/propertycompare` already exists.
- Merging `main` after S11-08 conflicts in a paragraph STEP-8 rewrites (`add-match-rule` SKILL.md or `comparison-layers.md`), so keeping both sides' text is not possible.

## 12. Open questions

None.

## 13. PR

**Title:** S11-09 — Property compare
**Body:** objective in one sentence; what shipped against each US; AC evidence (test names); deviations from the plan; manual check (the two `go test` runs of §10); `Plan: docs/deployment-plan/spike-11/prs/S11-09-property-compare.md`.
