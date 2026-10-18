# S11-10 — Subject compare

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 3 — Property, Subject and entity compare
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a new pure Go package and a registry section with no caller yet; nothing reaches the app)
**Board:** —
**Depends on:** S11-09 ([plan](S11-09-property-compare.md)), in progress at planning time. This plan assumes S11-09 lands as its plan says: `core/propertycompare` (`Compare`, `Value`, `Result`, `SpeakingPair`, the `Outcome*` and `Cause*` constants), `propertymeta.Meta`, `core/datevalue/datevaluetest` (`YMD`, `About`, `Phrase`), its `layerPackages` rows, and its STEP-8 text in `add-match-rule`, `match-rules.mdc`, `comparison-layers.md` §Layer 2 and `matching-model.md`. Branch from `main` after S11-09 merges, and read its **Landed** note first
**Blocks:** S11-11
**Branch:** `s11-10-subject-compare`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Layer 3](../comparison-layers.md#layer-3-subject-comparison) (requirements, the two axes, conflict weighs its penalty, registry shape); [match registry](../comparison-layers.md#match-registry) (typed keys, data-model facts stay in the catalog); [`docs/matching-model.md`](../../../matching-model.md) stack table, glossary and knob inventory (`subject.` rows); working rules "Sparse data first", "Registry numbers: carry over, then estimate", "Subject compare returns similarity (0–1) and evidence (points)", "One match registry", "One matching skill, grown with the layers", "Build, then migrate", "One guard, not one per package"
**Planned:** 2026-10-18 against `main` @ `95d1f34`, with S11-09 assumed landed
**Review:** [`S11-10-subject-compare.review.md`](S11-10-subject-compare.review.md) — see the log for the verdict

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

Layer 3 of matching exists: `core/subjectcompare` says how alike two Subjects of one type are over the properties both record. It runs property compare (layer 2) once per property the registry lists for the type, pools sex at birth with gender, reads an event's `start_date` and `end_date` as one period when it has no `date`, and returns two numbers: a **similarity** from 0 to 1 (are they alike?) and **evidence** in points (how strong is the connection?), with each property's layer 2 result and points. The registry gains its `subject` section, keyed by the `core/vocabkey` constants.

**Why now.** Slice 3 builds entity compare (S11-11) and lifespan (S11-12) on this layer, and S11-13 moves Promote onto it. "Build, then migrate" puts the layer in its own PR with its own tests before any caller moves.

**What stays.** `core/match` (`DefaultProfile`, `Score`, `Rank`, `poolPeriods`, `period`) and `core/graphalign/registry.go` (`PropertyScale`) are untouched; S11-13 moves Promote onto this layer and deletes them, and S11-16 deletes the graphalign registry. `subjectvocab`'s `TestDefaultProfileCardinality` stays while `match.DefaultProfile` exists. No catalog, wire, Swift or researcher-visible change.

## 2. User stories

All values are compared under `matchregistry.Default()` (`Subject` and `Value` sections) with the test metas of D-11 (the seeded value type and cardinality of each property). Names are built with `namevaluetest.Western`, dates with `datevaluetest.YMD` (`1817` is `YMD(1817, 0, 0)`). "¬x" is a negative value x; "(n)" is a neutral value. "Sim" is `Result.Similarity` (within 0.0001), "ev" is `Result.Evidence` (within 0.0001). Layer 1 relations and similarities below were read from `valuecompare.Compare` at `95d1f34`.

**US-1.** As the engine (entity compare S11-11, Promote S11-13), when I compare two Subjects of one type, I get a similarity and evidence from the properties both record, each property counted by its registry agree weight or conflict penalty.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | person: name `{James Robins}` / `{James Robins}` | sim 1, ev 2.5; `Properties`: `name` agree, points 2.5; `sex_at_birth` unknown (`no_values`), points 0 |
| AC-1.2 | person: name `{James Robins}` / `{James Robins}`, gender `{female}` / `{female}` | sim 1, ev 3.1 (gender counts under `sex_at_birth`, D-6) |
| AC-1.3 | person: name `{James Robins}` / `{James Robins}`, sex at birth `{male}` / `{female}` | sim 0.7143 (2.5 / 3.5), ev 1.5 |
| AC-1.4 | person: name `{J. Robins}` / `{James Robins}` | name partial (`Contains` 0.8): sim 0.8, ev 2.0 |
| AC-1.5 | person: name `{James Robins}` / `{Mary Robins}` | name conflict: sim 0, ev −3 |
| AC-1.6 | event: event type `{birth}` / `{birth}` | sim 1, ev 0.5 |
| AC-1.7 | event: event type `{birth}` / `{death}`, date `{1817}` / `{1817}` | sim 0.0909 (2 / 22), ev −18 |
| AC-1.8 | event: event type `{birth}` / `{death}` | sim 0, ev −20 |
| AC-1.9 | place: toponym `{Montréal}` / `{Montreal}` | toponym agree (`Equivalent` 0.95): sim 1, ev 3 (agree earns the full weight, D-5) |
| AC-1.10 | place: toponym `{York}` / `{London}` | toponym unknown (cause `cardinality`): sim 0, ev 0 |
| AC-1.11 | person: name `{James Robins}` / `{}`, sex at birth `{male}` / `{male}` | sim 1, ev 0.6; `name` unknown (`no_values`), points 0 |
| AC-1.12 | person: both sides empty | sim 0, ev 0; `Properties` = `name`, `sex_at_birth`, in that order, both unknown |
| AC-1.13 | kind `relationship` (no registry entry): name `{James Robins}` / `{James Robins}` | sim 0, ev 0, `Properties` nil |
| AC-1.14 | person, metas with no `name` entry: name `{James Robins}` / `{James Robins}` | `name` unknown, `Speaking.Result.Relation` `Unknown`, `Speaking.Detail.Cause` `unsupported`; sim 0, ev 0 |
| AC-1.15 | event: `event_name` `{Great Fire}` / `{Great Fire}` only | sim 0, ev 0; `Properties` = `date`, `event_type`, both unknown (an unlisted property is never read) |

**US-2.** As the engine, the properties the registry pairs are compared as one: sex at birth with gender, and an event's `date` with its `start_date` / `end_date`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | person: gender `{female}` / sex at birth `{female}` | `sex_at_birth` agree: sim 1, ev 0.6 |
| AC-2.2 | person: sex at birth `{male}` / gender `{female}` | `sex_at_birth` conflict: sim 0, ev −1 |
| AC-2.3 | person: sex at birth `{unknown (n)}` / `{male}` | `sex_at_birth` unknown (`no_values`), `Excluded` the `unknown` value; sim 0, ev 0 |
| AC-2.4 | person: sex at birth `{¬male}` / gender `{male}` | `sex_at_birth` conflict, cause `negative`: ev −1 |
| AC-2.5 | event: date `{1817}` / start `{1815}`, end `{1820}` | `date` partial: `datevalue.CompactDisplay(*Speaking.B.Date)` is `Between 1815 and 1820`, `Speaking.B.Ref` nil, relation `Contains` 0.55; sim 0.55, ev 1.1 |
| AC-2.6 | event: start `{1815}`, end `{1820}` / start `{1815}`, end `{1820}` | `date` agree (`Same`): sim 1, ev 2 |
| AC-2.7 | event: date `{1817}`, start `{1700}`, end `{1701}` / date `{1817}` | `date` agree (`Same`): side A's own date speaks; no period is built for it |
| AC-2.8 | event: start `{1815}` / date `{1817}` | `CompactDisplay(*Speaking.A.Date)` is `After 1815`; `Contains` 0.2: ev 0.4 |
| AC-2.9 | event: end `{1820}` / date `{1817}` | `CompactDisplay(*Speaking.A.Date)` is `Before 1820`; `Contains` 0.2: ev 0.4 |
| AC-2.10 | event: start `{1821}`, end `{1820}` / date `{1817}` | no period: `date` unknown (`no_values`); ev 0 |
| AC-2.11 | event: start `{1816, 1815}` (and `{1815, 1816}`), end `{1820}` / date `{1815}` | the period is 1815 – 1820 (earliest start): `Contains` 0.55, ev 1.1 (1816 – 1820 would be `Near` 0.4) |
| AC-2.12 | event: date `{¬1817}`, start `{1815}`, end `{1820}` / date `{1817}` | the period joins side A beside its negative: `date` conflict, cause `negative`; ev −3 |
| AC-2.13 | event: start `{1815}`, end `{1820}` / date `{1830}` | `date` conflict (`Different`): ev −3 |

**US-3.** As a contributor, the layer's invariants hold on generated cases, so a later change that breaks one fails a test.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Any generated case; swap the sides | Similarity and evidence equal (within 1e-9); each property's outcome equal |
| AC-3.2 | Any generated case; shuffle the values of every property on both sides | Similarity, evidence and each property's outcome equal |
| AC-3.3 | Any generated case | 0 ≤ similarity ≤ 1 |
| AC-3.4 | Any generated case; pick a lead whose group (the lead, its pooled properties, its period start and end) records nothing on the other side, and add one value to a property of that group on this side | Similarity, evidence and each property's outcome unchanged |
| AC-3.5 | Any generated case where a lead's group records nothing on either side; give both sides the same one value for the lead (name `James Robins`, sex at birth `male`, event type `birth`, date `1817`, toponym `York`) | Evidence does not go down, similarity does not go down |
| AC-3.6 | Any generated person case with name `{James Robins}` / `{Mary Robins}`; any event case with event type `{birth}` / `{death}` | Evidence < 0, whatever the other properties say |

## 3. Scope

| In | Covered by |
| --- | --- |
| `core/subjectcompare`: typed keys, `Rules`, `Compare`, similarity and evidence (D-1 – D-5, D-9, D-10) | STEP-2, STEP-4, US-1 |
| Pools (sex with gender) and periods (`date` with `start_date` / `end_date`) (D-6 – D-8) | STEP-5, US-2 |
| Registry `subject` section with today's and estimated numbers; inventory rows renamed to the built keys (D-12, D-13) | STEP-3, AC-1.1 – AC-1.9 |
| `vocabkey.PropertyKeys`; the registry key check (the port of `TestDefaultProfileCardinality`) (D-14) | STEP-1, STEP-3, T-6, T-7 |
| Exact table and invariant tests (D-11, D-15) | STEP-4 – STEP-6, all ACs |
| Guard row; `match-rules.mdc` glob and check line (D-16) | STEP-2 |
| Docs the layer makes true or false: `comparison-layers.md` §Layer 3, `matching-model.md`, the `add-match-rule` skill (D-17) | STEP-7 |

| Out | Home |
| --- | --- |
| Any caller: Promote scoring, `match.Score`, `Rank`, `Evaluate`, `alikeNew`, `pairNode` | S11-13 |
| Deleting `match.DefaultProfile`, `match.DatePeriods`, `poolPeriods`, `period` and `subjectvocab`'s `TestDefaultProfileCardinality` | S11-13 (deletes `core/match`) |
| Deleting `graphalign` `PropertyScale` and the rest of `core/graphalign/registry.go` | S11-16 |
| Profiles from reconciler output; `Excluding` | S11-11 |
| Lifespan: its registry fields, its inventory rows' keys, its comparison | S11-12 |
| Evidence lines that name the Observations behind a pooled period (refs of `start_date` / `end_date`) | S11-19 |
| An accept bar, bands, a 0–1 confidence over the total | S11-16 (bands, accept bar); confidence not planned (labeler idea, [`docs/ideas/matching-test-suite.md`](../../../ideas/matching-test-suite.md)) |
| Weights for `start_date` / `end_date` on places and place relationships | Not planned — the inventory lists none; places match on toponym and `part_of` (S11-14) |
| Tuning any number beyond the inventory's start values | Not planned — working rule "Registry numbers: carry over, then estimate" |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Layer 3 package | `core/subjectcompare` | Does not exist |
| Layer 2 | `core/propertycompare/propertycompare.go` *(new)* in S11-09: `Compare(a, b []Value, meta propertymeta.Meta, rules valuecompare.Rules) Result`; `Value{valuecompare.Value; Negative, Neutral bool}`; `Result{Outcome, Cause, Speaking *SpeakingPair, Excluded}`; `OutcomeAgree`, `OutcomePartial`, `OutcomeConflict`, `OutcomeUnknown` | Sets every value's `Type` from `meta.ValueType`; a value type with no comparer (the zero `Meta`) speaks as `Unknown` / `unsupported`; reads no number |
| Property facts | `core/propertymeta/propertymeta.go` `Meta{ValueType, Cardinality}` *(new)* in S11-09 | Empty cardinality reads as single |
| Registry | `core/matchregistry/registry.go` `Registry{Value valuecompare.Rules}`, `Default()`; `core/matchregistry/registry_test.go` `TestDefaultIsFresh` | One section. The design file list puts the subject section in `subject.go` |
| Knob inventory | `docs/matching-model.md` §Knob inventory rows `subject.person.name.Agree` … `subject.place.toponym.Conflict`, `subject.Pairs.event.date`, `subject.Pairs.person.sex_at_birth`, `subject.person.lifespan.*`; §Worked example steps 3 – 4 | Keys written before the section's Go shape existed: a field (`Pairs`) and kind map keys (`person`) at the same level, which no Go type can produce (D-13) |
| Inventory guard | `core/matchregistry/inventory_test.go` `notYetBuilt` (`"subject."` first), `registryKeys`, `TestKnobInventoryMatchesRegistry` | `registryKeys` lowercases the top field, walks maps whose element is a struct or a map by sorted key, and stops at any other leaf |
| Purity guard | `core/matchregistry/guard_test.go` `layerPackages` | After S11-09: `core/valuecompare` and `core/propertycompare` with numbers true, the pure value and fixture packages false |
| Keys | `core/vocabkey/vocabkey.go`: kinds `Person`, `Event`, `Place`, `PrimaryKinds()`; untyped constants `PropertyName`, `PropertySexAtBirth`, `PropertyGender`, `PropertyEventType`, `PropertyDate`, `PropertyStartDate`, `PropertyEndDate`, `PropertyToponym`, `PropertyLikeness` | No list of the Property constants |
| Seed check | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` | Checks eight Property constants against the seed with a hand-written list (`PropertyLikeness` missing) |
| Old profile check | `core/database/subjectvocab/profile_test.go` `TestDefaultProfileCardinality` | Checks `match.DefaultProfile` features against the seed's value type and cardinality |
| Old Subject scoring | `core/match/match.go` `Score` (Weight × similarity, −Contradiction at 0); `core/match/profiles.go` `Feature`, `Profile`; `core/match/registry.go` `DefaultProfile`, `DatePeriods` | `Rank`'s 10-point scale; `start_date` / `end_date` scored as their own features |
| Old periods | `core/match/periods.go` `poolPeriods`, `pools`, `period`, `endFields`, `firstLeader` | Pools only when the other side has a voting `date` and no start or end; picks the first leader by support |
| Promote's weights | `core/graphalign/registry.go` `DefaultConfig` `PropertyScale` (name 2.5 / 3, sex and gender 0.6 / 1) | The source of the person numbers (D-12) |
| Date ends | `core/valuecompare/datewindow.go` `DateEnd{Year, Month, Day, Lo, Hi, …, Open}`, `DateEnds(d) (first, last DateEnd, ok bool)` | `DateEnds` returns an open first end for `BEF` and an open last end for `AFT` |
| Fixtures | `core/namevalue/namevaluetest/namevaluetest.go` `Western`; `core/datevalue/datevaluetest` *(new)* in S11-09: `YMD`, `About`, `Phrase` | Pure; listed in `layerPackages` |
| Docs | `docs/deployment-plan/spike-11/comparison-layers.md` §Layer 3; `docs/matching-model.md` stack table row 3, glossary **pair (compared together)**; `.agents/skills/add-match-rule/SKILL.md` (as S11-09 leaves it); `.cursor/rules/match-rules.mdc` | §Layer 3's table reads 0.81 for "name agrees, sex conflicts", written before "a conflict weighs its penalty in similarity" was decided; with the penalty it is 0.71 (AC-1.3) |

**Precedent.** S11-09 (`core/propertycompare`) and [#441](https://github.com/mendahu/provenencia/pull/441) (S11-02, `core/valuecompare` with the registry's `value` section). Copy from S11-02: the layer declares its `Rules` types with doc comments and no numbers, `core/matchregistry/<section>.go` sets the values with a source comment on each, `Registry` gains one field, and `notYetBuilt` loses the built prefix. Copy from S11-09: the package doc naming the layer and its purity, an external `_test` package reading `matchregistry.Default()`, a `layerPackages` row with numbers true, private test helpers, and a fixed-seed invariants test.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Package and entry.** `core/subjectcompare`, files `subjectcompare.go` (types, `Compare`) and `pairs.go` (pools and periods), plus tests. `func Compare(kind SubjectType, a, b Values, metas Metas, rules Rules, value valuecompare.Rules) Result`. Types (all *new*): `SubjectType`, `PropertyKey`, `Weights`, `Period`, `Rules`, `Values`, `Metas`, `Result`, `PropertyResult` (sketch in §6) | One entry like layers 1 and 2. Layer 3 reads its `subject` section and passes `value` to layer 2 unread, as layer 2 passes it to layer 1. `kind` is one argument because both Subjects are of one type | A `Subject{Type, Values}` struct per side — the caller would have to keep two types equal; one combined rules argument — layer 2 already takes `valuecompare.Rules` alone |
| D-2 | **Typed keys.** `type SubjectType string`, `type PropertyKey string`, declared in `subjectcompare`. The registry writes the `vocabkey` constants directly as keys (they are untyped, so they convert with no copy). No key constants are declared in `subjectcompare` | Map row and S11-23: `vocabkey` is the one home for the key strings; a typed map key stops a value type string or a term key landing in a key position | Declaring `subjectcompare.PropertyName` and the like — a second copy of `vocabkey`; plain `string` keys — the registry rule asks for typed keys |
| D-3 | **Catalog facts come in as `Metas`.** `type Metas map[PropertyKey]propertymeta.Meta`, filled by the caller (S11-13's loader, the tests' `metas()`). A lead property with no entry is compared with the zero `Meta`, which layer 2 reads as `Unknown` / `unsupported` (AC-1.14). Pooled and period properties are compared under their lead's `Meta` | The registry rule: value type and cardinality are data-model facts that stay in the catalog; the registry names properties by key only. No special case: layer 2 already reports a missing type | Value type and cardinality in `Weights` — duplicates the seed, the drift `TestDefaultProfileCardinality` guards against; skipping a property with no meta — hides a caller bug |
| D-4 | **Which properties.** `Compare` reads only the lead properties in `rules.Weights[kind]`, in ascending key order, and returns one `PropertyResult` for each, unknown ones included. A kind with no `Weights` entry returns the zero `Result` (AC-1.13). Values for properties the registry does not list are never read (AC-1.15) | "One registry defines every Subject type, its properties, and how each one counts." Sorted order makes the float sums and `Properties` deterministic; unknown rows let S11-19 say what was missing | Iterating the map — a nondeterministic sum order; comparing every property both sides have — a property with no registry weight would count for nothing anyway |
| D-5 | **Points and similarity, from the outcome.** For lead property `p` with weights `w` and layer 2 result `r`: agree → points `w.Agree`, counts `w.Agree` × 1 in similarity; partial → points `w.Agree × s`, counts `w.Agree` × `s`, where `s` = `r.Speaking.Result.Similarity`; conflict → points `−w.Conflict`, counts `w.Conflict` × 0; unknown → points 0, not counted. `Evidence` = Σ points. `Similarity` = Σ (weight × score) / Σ weight over counted properties, 0 when nothing is counted. An agree earns the full weight whatever its relation (`Equivalent` 0.95 earns 3 on a toponym, AC-1.9) | §Layer 3 "decided": two axes, a conflict weighs its penalty in similarity (birth vs death with matching dates ≈ 0.09). Layer 2's outcome is the verdict; layer 1's similarity grades only a partial. The glossary's **agree weight** is "points a property earns when it agrees" | Agree × similarity on agree — an accent fold would earn less than an exact match though layer 2 says both agree; counting unknown as 0 in similarity — "a property on one side only neither helps nor hurts" |
| D-6 | **Pools.** `rules.Pools[kind][lead]` lists properties whose values join the lead's on the same side before layer 2 runs: side A's values are A's lead values, then A's values for each pooled property in list order, all with their flags; the same for side B. Default: person `sex_at_birth` pools `gender` | §Layer 3: "both sides' sex and gender values are pooled into one direct comparison, so a Subject recording only gender still meets an entity recording only sex at birth" (AC-2.1). Layer 2 is order-independent, so list order changes no outcome | A sex-only and a gender-only comparison — the two would never meet; pooling on one side only — asymmetric |
| D-7 | **Periods.** `rules.Periods[kind][lead]` is a `Period{Start, End PropertyKey}` that stands for the lead on a side with **no positive lead value**: that side's values are its lead values (negatives included) plus one built period value, when one can be built. The other side is untouched and builds its own. Default: event `date` with `Start` `start_date`, `End` `end_date` | §Layer 3: "an event's `date` against another's `start_date` + `end_date`, read as one period". Each side decides alone, so the rule is symmetric (AC-3.1) and two start / end-only events compare period with period (AC-2.6). Keeping the side's negatives lets a denied date still contradict (AC-2.12) | `match.poolPeriods`' rule (pool only when the other side has a date and no start or end) — asymmetric, and two start / end-only events would not compare at all once start and end lose their own weights; building a period on a side that has a date — a side's own date says more than its span |
| D-8 | **Building the period** (`period`, unexported, `pairs.go`). Read the side's positive values of `Start` and `End` that have a `Date`. A start end is `first` of `valuecompare.DateEnds(*v.Date)` with `ok` and not `Open`; the start used is the one with the lowest `Lo`, then the lowest `Hi`. An end end is `last` with `ok` and not `Open`; the end used is the one with the highest `Hi`, then the highest `Lo`. Both found and start `Lo` ≤ end `Hi` → a range (`Kind` `datevalue.KindRange`, `Calendar` `"gregorian"`, start fields from the start end, end fields from the end end). Both found and start `Lo` > end `Hi` → no period (AC-2.10). Start only → a point with `Qualifier` `datevalue.QualifierAFT` on the start; end only → `datevalue.QualifierBEF` on the end. A month or day of 0 on a `DateEnd` is left nil (`endFields`, as `match` does). The value has no `Ref`, `Negative` or `Neutral` | The widest span the side records, picked by day number, so the input order never matters (AC-2.11). Same construction as `match.period`, which S11-13 deletes | `match.firstLeader` — reads support, which never reaches layers 2 and 3; one period per start × end combination — several made-up values for one fact; a ref copied from the start — the period is no one Observation (S11-19 words it) |
| D-9 | **`Result` shape.** `Result{Similarity, Evidence float64; Properties []PropertyResult}`; `PropertyResult{Key PropertyKey; Result propertycompare.Result; Points float64}`. `Key` is the lead property. `Properties` is nil only when the kind has no `Weights` entry | Rule 4, "output is the explanation": the layer above and S11-19 read which values spoke, why a property was unknown, and what it added, without redoing the work | A map of results — no order; similarity per property — layer 2's speaking pair already carries it |
| D-10 | **Purity and numbers.** `core/subjectcompare` imports only the standard library, `core/propertycompare`, `core/valuecompare`, `core/propertymeta` and `core/datevalue`; its non-test files hold no numeric literal other than 0 and 1 | Working rules "Pure means the package builds and tests with no build tags" and "One guard": the layer is a `layerPackages` row with numbers true | — |
| D-11 | **Tests and fixtures.** External package `subjectcompare_test`, table-driven (`.cursor/rules/go-tests.mdc`), no build tags. Rules: `matchregistry.Default()` (`.Subject`, `.Value`). Helpers in `core/subjectcompare/helpers_test.go` *(new)*: `term(key)`, `text(s)`, `name(form)` (`namevaluetest.Western`), `date(d *datevalue.Value)` (`datevaluetest.YMD`), `neg(v)`, `neutral(v)`, each returning a `propertycompare.Value` with `Type` empty; `metas() subjectcompare.Metas`, a fresh map of the seeded facts: `name` name single; `sex_at_birth`, `gender`, `event_type` term single; `date`, `start_date`, `end_date` date single; `toponym` text multiple; `event_name` text single. `TestCompare` holds every US-1 and US-2 row; each row checks sim, ev, and for the property it names: outcome, cause, points, and the speaking relation or date it names (`datevalue.CompactDisplay` for AC-2.5, AC-2.8, AC-2.9) | The S11-09 test layer. Layer 2's helpers live in an external test package no one can import; one small set per layer package is the S11-09 precedent | Importing layer 2's helpers — not importable; a shared `propertycomparetest` package — a new package for six one-line builders |
| D-12 | **Registry `subject` section.** `Registry.Subject subjectcompare.Rules` *(new field)*, set in `core/matchregistry/subject.go` *(new)* `defaultSubject()`. Weights: person `name` 2.5 / 3 and `sex_at_birth` 0.6 / 1 (`graphalign` `PropertyScale`); event `event_type` 0.5 / 20 and `date` 2 / 3; place `toponym` 3 / 3 (inventory estimates with their reasoning as the source comment). Pools: person `sex_at_birth` → `[gender]`. Periods: event `date` → `{start_date, end_date}`. Every key is a `vocabkey` constant. `Default()` builds fresh maps and slices | The knob inventory's start values (working rule "Registry numbers: carry over, then estimate"); the section file the design names | Carrying `Rank`'s 10-point scale — the inventory retired it ("one weight set") |
| D-13 | **Inventory keys follow the Go shape.** `registryKeys` can't produce both `subject.person.…` (kind map keys under the section) and `subject.Pairs.…` (a field under the section). The built keys are `subject.Weights.<kind>.<property>.Agree` / `.Conflict`, `subject.Pools.person.sex_at_birth`, `subject.Periods.event.date.Start` and `.End`. STEP-3 renames the eleven rows (values, sources and Changes cells unchanged; the two `Pairs` rows become three), the worked example's key names, and the intro sentence. The `subject.person.lifespan.*` rows keep their keys; `notYetBuilt`'s `"subject."` becomes `"subject.person.lifespan."` | The guard is the source of truth for keys; the rows predate the type. Lifespan's shape is S11-12's to decide | A custom walk in `registryKeys` for the subject section — a special case in the one guard; one `Pairs` map with a join kind — period needs a start and an end, which a list of keys says only by position |
| D-14 | **Registry keys are `vocabkey` keys** (the port of `TestDefaultProfileCardinality`). `vocabkey.PropertyKeys() []string` *(new)* returns a fresh slice of the nine Property constants in declaration order. `core/matchregistry/subject_test.go` `TestSubjectSection` *(new)* checks `Default().Subject`: every kind in `Weights`, `Pools` and `Periods` is in `vocabkey.PrimaryKinds()`; every property key (leads, pooled properties, `Start`, `End`) is in `vocabkey.PropertyKeys()`; each `Pools` and `Periods` lead has a `Weights` entry for its kind; no pooled property, `Start` or `End` has a `Weights` entry of its own or is a lead or member of another pair in its kind. `TestVocabKeysMatchSeed` iterates `vocabkey.PropertyKeys()` instead of its hand-written list | The map's scope entry: "every registry `PropertyKey` is a `vocabkey` constant"; S11-23's `TestVocabKeysMatchSeed` ties the constants to the seed. One list of the constants replaces the hand-written one (which missed `likeness`) | A third hand-written list in the new test; checking the seed from `matchregistry` — impure |
| D-15 | **Generator.** `rand.New(rand.NewSource(1))`; `TestInvariants` has one subtest per AC-3.1 – AC-3.6 per kind (`AC-3.1/person`, `AC-3.1/event`, `AC-3.1/place`, and the same for each AC; AC-3.6 runs person and event only), 500 cases each. Pools: person name `James Robins`, `James Robbins`, `J. Robins`, `Mary Robins`; sex at birth and gender `male`, `female`, `unknown (n)`. Event type `birth`, `death`, `baptism`; date `1817`, `1818`, `1830`, `ABT 1817`; start `1815`, `1816`; end `1820`. Place toponym `York`, `york`, `Montréal`, `Montreal`, `London`. Each side draws 0 – 2 pool values for each of its kind's properties (person: `name`, `sex_at_birth`, `gender`; event: `event_type`, `date`, `start_date`, `end_date`; place: `toponym`); each drawn value is negative when `r.Intn(4) == 0`. AC-3.4 picks, with `r.Intn`, a lead whose group is empty on side B and adds one positive pool value to a property of that group on side A (a case with no such lead is drawn again). AC-3.5 picks a lead whose group is empty on both sides and adds the AC's value to both (a case with no such lead is drawn again). AC-3.6 fixes name or event type as the AC says and draws the rest. A failure reports `t.Errorf("case %d: …", i)` with both sides' values | The S11-09 generator shape; fixed pools cover agree, partial, conflict, unknown, negatives, neutrals, pools and periods | Random strings — would rarely relate |
| D-16 | **Guard and editor rule.** `layerPackages` gains `{"core/subjectcompare", true}` after `core/propertycompare`. `.cursor/rules/match-rules.mdc`: `globs` gains `core/subjectcompare/**`; the Check line becomes "`go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/matchregistry/...` with no tags." | One guard; the editor rule loads for the new layer | — |
| D-17 | **Docs.** `comparison-layers.md` §Layer 3 gains a **Built in S11-10** paragraph and its "name agrees, sex conflicts" similarity becomes 0.71; `matching-model.md` stack table row 3 reads `subject` and passes `value` through, the glossary's **pair (compared together)** row names the two kinds; the `add-match-rule` skill gains layer 3 (description, a rule bullet, an "Add a Subject property or pair" recipe, Check). Exact text in STEP-7 | The design doc and the model page are kept true as layers land; the map's working rule "One matching skill" names S11-10 | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/subjectcompare` *(new)* | package | `core/match` `Score`, `Profile` | `match` is impure (imports `autoreconcile`), reads support, and is deleted in S11-13 ("Build, then migrate") |
| `subjectcompare.Compare` *(new)* | func | `match.Score`, `graphalign` `ApplyScale` | Replaces both in S11-13; reads layer 2 outcomes, not a similarity number with a floor |
| `subjectcompare.Rules`, `Weights`, `Period` *(new)* | type | `match.Profile`, `match.Feature`, `graphalign.Scale`, `match.PeriodProperties` | Same concepts on the retired scale; the registry rule puts the section's types in the layer |
| `subjectcompare.SubjectType`, `PropertyKey` *(new)* | type | `match.Property{Key, Origin}` | Typed keys the registry rule asks for; matching reads product properties only, so the origin is the caller's filter (S11-13) |
| `subjectcompare.Values`, `Metas` *(new)* | type | `match.Values`, `match.PropertyMeta` | `match.Values` carries reconciler fields; `Metas` holds S11-09's pure `propertymeta.Meta` |
| `subjectcompare.Result`, `PropertyResult` *(new)* | type | `match.Match`, `match.Reason` | Two axes (similarity and evidence) and layer 2's result per property, where `Match` has one score |
| `pool`, `period`, `endFields` *(new, unexported, `pairs.go`)* | func | `match.poolPeriods`, `match.period`, `match.endFields` | Same date construction; a symmetric rule with no support (D-7, D-8). The `match` copies go in S11-13 |
| `matchregistry.Registry.Subject`, `defaultSubject` *(new)* | field, func | `matchregistry.Registry.Value`, `defaultValue` | The second section, the same way |
| `vocabkey.PropertyKeys` *(new)* | func | `vocabkey.PrimaryKinds`; the hand list in `TestVocabKeysMatchSeed` | Same shape as `PrimaryKinds`; replaces the hand list |
| `core/matchregistry/subject_test.go` `TestSubjectSection` *(new)* | test | `TestWesternPatternCoversEveryPartType`; `subjectvocab` `TestDefaultProfileCardinality` | Checks the new section's keys; the old test leaves with `match.DefaultProfile` in S11-13 |
| `core/subjectcompare/helpers_test.go` *(new)* | helper | `core/propertycompare/helpers_test.go` (S11-09) | Not importable (external test package); private per package, as S11-09 D-12 does |

**Builds on.** `propertycompare.Compare` and its result (S11-09), `propertymeta.Meta`, `valuecompare.DateEnds`, the `vocabkey` constants (S11-23), the registry section pattern (S11-02), the `layerPackages` and inventory guards, `namevaluetest`, `datevaluetest`, the `add-match-rule` skill.
**Deletes.** The hand-written key list in `TestVocabKeysMatchSeed`. Nothing else becomes dead: `core/match` and `graphalign` keep their callers until S11-13 and S11-16.
**Next adds.** S11-11 passes a profile built from the conclusion as one side's `Values`, so this layer does not change. S11-12 adds the lifespan fields to `subjectcompare.Rules` (renaming the `subject.person.lifespan.*` rows to the shape it picks) and a people-only comparison beside the property loop. S11-13 calls `Compare` from Promote with metas from the catalog and deletes `match.DefaultProfile`, `poolPeriods` and `TestDefaultProfileCardinality`. S11-14's `segment` section follows the same section pattern. S11-19 reads `Properties` for evidence lines. A new property weight or pair is a registry entry and an inventory row.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/vocabkey` | `func PropertyKeys() []string` *(new)* (D-14) |
| `core/subjectcompare` *(new)* | `subjectcompare.go`: types and `Compare` (D-1 – D-5, D-9); `pairs.go`: `pool`, `period`, `endFields` (D-6 – D-8) |
| `core/matchregistry` | `registry.go` `Registry.Subject` *(new)*; `subject.go` *(new)* `defaultSubject`; `inventory_test.go` `notYetBuilt`; `guard_test.go` `layerPackages`; `registry_test.go` `TestDefaultIsFresh`; `subject_test.go` *(new)* |
| `core/database/subjectvocab` | `vocabkey_test.go` `TestVocabKeysMatchSeed` reads `vocabkey.PropertyKeys()` |

```go
// Package subjectcompare is layer 3 of matching: how alike two Subjects of one
// type are over the properties both record. Pure: no I/O, no catalog, no
// numbers; the subject section and layer 1's rules come in as arguments.
package subjectcompare

// SubjectType is a Subject type key: a vocabkey kind.
type SubjectType string

// PropertyKey is a product Property key: a vocabkey Property constant.
type PropertyKey string

// Weights are one lead property's part in a Subject comparison, in points.
type Weights struct {
	Agree    float64 // earned on agree; Agree × similarity on partial
	Conflict float64 // lost on conflict; also its weight in Similarity
}

// Period reads two date properties as one range that stands for the lead on
// a side with no positive lead value.
type Period struct{ Start, End PropertyKey }

// Rules is the match registry's subject section.
type Rules struct {
	// Weights lists, per Subject type, the lead properties compared and
	// what each counts.
	Weights map[SubjectType]map[PropertyKey]Weights
	// Pools lists properties whose values join the lead's on each side.
	Pools map[SubjectType]map[PropertyKey][]PropertyKey
	// Periods lists the start and end that stand for a lead date.
	Periods map[SubjectType]map[PropertyKey]Period
}

// Values are one Subject's values by property. Product properties only.
type Values map[PropertyKey][]propertycompare.Value

// Metas are the catalog's value type and cardinality for each property.
type Metas map[PropertyKey]propertymeta.Meta

type Result struct {
	Similarity float64          // 0 to 1: weighted mean over counted properties; 0 when none
	Evidence   float64          // points: agree weights minus conflict penalties
	Properties []PropertyResult // one per lead property, ascending key
}

type PropertyResult struct {
	Key    PropertyKey // the lead property
	Result propertycompare.Result
	Points float64 // what it added to Evidence
}

func Compare(kind SubjectType, a, b Values, metas Metas, rules Rules, value valuecompare.Rules) Result
```

`Compare`, per lead key of `rules.Weights[kind]` in ascending order: build each side with `pool` (D-6) and, when the lead has a `Periods` entry, `period` (D-7, D-8); call `propertycompare.Compare(sideA, sideB, metas[key], value)`; add points and the similarity terms (D-5). Errors: none — `Compare` cannot fail.

```go
// core/matchregistry/subject.go
func defaultSubject() subjectcompare.Rules {
	return subjectcompare.Rules{
		Weights: map[subjectcompare.SubjectType]map[subjectcompare.PropertyKey]subjectcompare.Weights{
			vocabkey.Person: {
				vocabkey.PropertyName:       {Agree: 2.5, Conflict: 3}, // graphalign name PropertyScale
				vocabkey.PropertySexAtBirth: {Agree: 0.6, Conflict: 1}, // graphalign sex / gender PropertyScale; pools gender
			},
			vocabkey.Event: {
				vocabkey.PropertyEventType: {Agree: 0.5, Conflict: 20}, // estimate: generic alone; a birth is never a death
				vocabkey.PropertyDate:      {Agree: 2, Conflict: 3},     // estimate: cold-start ln(0.85 / 0.1) rounded; ConflictPenalty
			},
			vocabkey.Place: {
				vocabkey.PropertyToponym: {Agree: 3, Conflict: 3}, // estimate: cold-start + toponym bonus, rounded; ConflictPenalty
			},
		},
		Pools: map[subjectcompare.SubjectType]map[subjectcompare.PropertyKey][]subjectcompare.PropertyKey{
			vocabkey.Person: {vocabkey.PropertySexAtBirth: {vocabkey.PropertyGender}},
		},
		Periods: map[subjectcompare.SubjectType]map[subjectcompare.PropertyKey]subjectcompare.Period{
			vocabkey.Event: {vocabkey.PropertyDate: {Start: vocabkey.PropertyStartDate, End: vocabkey.PropertyEndDate}},
		},
	}
}
```

### FFI

None — no caller yet.

### macOS

None — no app change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — One list of the Property keys**
Files: `core/vocabkey/vocabkey.go`, `core/vocabkey/vocabkey_test.go`, `core/database/subjectvocab/vocabkey_test.go`.
Do: add `PropertyKeys()` (D-14) with a doc comment in the words of `PrimaryKinds` ("PropertyKeys returns a new slice of the Property keys on each call."); add `TestPropertyKeys` *(new)* (T-7) in the style of `TestPrimaryKinds`; change `TestVocabKeysMatchSeed` to range over `vocabkey.PropertyKeys()`.
Check: `go test ./core/vocabkey/...` passes; `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/ -run TestVocabKeysMatchSeed` passes.

**STEP-2 — Package skeleton and the guard**
Files: `core/subjectcompare/subjectcompare.go` *(new)*, `core/matchregistry/guard_test.go`, `.cursor/rules/match-rules.mdc`. Skill: `add-match-rule` ("Add a layer package").
Do: the package doc, every type of §6, and `Compare` with D-4 only (it returns a `PropertyResult` per lead key with layer 2's result and points 0, similarity 0, evidence 0). Add the `layerPackages` row and the `match-rules.mdc` changes (D-16).
Check: `go test ./core/matchregistry/ -run TestLayerPackages -v` passes and lists `core/subjectcompare/pure` and `core/subjectcompare/numbers`.

**STEP-3 — The registry's subject section**
Files: `core/matchregistry/registry.go`, `core/matchregistry/subject.go` *(new)*, `core/matchregistry/registry_test.go`, `core/matchregistry/subject_test.go` *(new)*, `core/matchregistry/inventory_test.go`, `docs/matching-model.md`. Skill: `add-match-rule` ("Add a knob").
Do: `Registry.Subject` with the doc comment "Subject is layer 3: what each property counts when two Subjects of one type are compared." and `Default()` filling it from `defaultSubject()` (D-12), each number with its source comment from the inventory's Source column. Extend `TestDefaultIsFresh`: change `first.Subject.Weights[vocabkey.Person][vocabkey.PropertyName]`, append to `first.Subject.Pools[vocabkey.Person][vocabkey.PropertySexAtBirth]` and set `first.Subject.Periods[vocabkey.Event][vocabkey.PropertyDate]`, then check a new `Default()` is unchanged. Add `TestSubjectSection` (D-14, T-6). Rename the inventory rows, worked example keys and intro sentence per D-13 (exact text below) and change `notYetBuilt`.
- §Knob inventory intro, the sentence "A lowercase part is a typed map key: a Subject kind, a property key, a derived property (`lifespan`), or a bridge or participation key; a pair compared together is keyed by its lead property (`sex_at_birth` covers `gender`)." stays, and after it: "The subject section's weights are `subject.Weights.<kind>.<property>`; a pooled pair is keyed by its lead in `subject.Pools`, a period by its lead in `subject.Periods`."
- Rows: `subject.person.name.Agree` → `subject.Weights.person.name.Agree`, and the same `subject.Weights.` prefix for `name.Conflict`, `sex_at_birth.Agree`, `sex_at_birth.Conflict`, `event.event_type.Agree`, `event.event_type.Conflict`, `event.date.Agree`, `event.date.Conflict`, `place.toponym.Agree`, `place.toponym.Conflict`. `subject.Pairs.event.date` becomes two rows: "| `subject.Periods.event.date.Start` | `start_date` | `match.DatePeriods` | Start of the period that stands for `date` on a side with no date |" and "| `subject.Periods.event.date.End` | `end_date` | `match.DatePeriods` | End of that period |". `subject.Pairs.person.sex_at_birth` becomes the row `subject.Pools.person.sex_at_birth`, Start `gender`, the Source cell copied unchanged from today's row (the "decided" link to `comparison-layers.md`), Changes "Properties whose values join `sex_at_birth`'s on each side".
- §Worked example step 3: `subject.person.name.Agree` → `subject.Weights.person.name.Agree`; step 4: `subject.event.event_type.Agree` → `subject.Weights.event.event_type.Agree`, `subject.event.date.Agree` → `subject.Weights.event.date.Agree`. Step 3's `subject.person.lifespan.Agree` stays.
Check: `go test ./core/matchregistry/...` passes (`TestKnobInventoryMatchesRegistry`, `TestDefaultIsFresh`, `TestSubjectSection`, `TestLayerPackages`).

**STEP-4 — Points and similarity**
Files: `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/helpers_test.go` *(new)*, `core/subjectcompare/subjectcompare_test.go` *(new)*.
Do: D-5 in `Compare`. Write the helpers (D-11) and `TestCompare` with the US-1 rows (T-1).
Check: `go test ./core/subjectcompare/...` passes with no tags.

**STEP-5 — Pools and periods**
Files: `core/subjectcompare/pairs.go` *(new)*, `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/subjectcompare_test.go` *(new)*, created in STEP-4.
Do: `pool`, `period`, `endFields` (D-6 – D-8) and their use in `Compare`. Add the US-2 rows to `TestCompare` (T-2).
Check: `go test ./core/subjectcompare/...` passes; `go vet ./core/subjectcompare/...` is clean.

**STEP-6 — Invariants**
Files: `core/subjectcompare/invariants_test.go` *(new)*.
Do: `TestInvariants` per D-15 (T-3).
Check: `go test ./core/subjectcompare/... -run TestInvariants -v` passes; then `go test ./core/subjectcompare/... ./core/propertycompare/... ./core/valuecompare/... ./core/matchregistry/... ./core/vocabkey/...` passes with no tags.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching-model.md`, `.agents/skills/add-match-rule/SKILL.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do:
- `comparison-layers.md` §Layer 3, the options table under "Decided (2026-10-16): two axes": the row "| Name agrees, sex conflicts | 0.81 | 1.5 |" becomes "| Name agrees, sex conflicts | 0.71 | 1.5 |". After the **Registry shape** paragraph, a paragraph *(new)*: "**Built in S11-10.** `subjectcompare.Compare(kind, a, b, metas, rules, value)` runs layer 2 once per lead property in `subject.Weights[kind]`, in key order. Agree earns the agree weight, partial the agree weight × the speaking pair's similarity, conflict loses the conflict penalty, unknown counts for nothing. Similarity is the weighted mean over counted properties: agree 1 and partial its similarity, each at the agree weight; conflict 0 at the penalty. `subject.Pools` adds a property's values to the lead's on each side (sex with gender). `subject.Periods` builds, on a side with no positive `date`, one range from its earliest start and latest end, `AFT` the start or `BEF` the end when it has only one. Value types and cardinalities come from the catalog as `Metas`."
- `matching-model.md` §The stack, table row 3: the Registry section cell `` `subject` `` becomes "`subject`, and `value` passed through to layer 2". §Glossary, the **pair (compared together)** row becomes "| **pair (compared together)** | Two or more properties compared as one, keyed by the lead: a pool adds one property's values to the lead's on each side (sex at birth with gender); a period reads `start_date` and `end_date` as one range for `date` on a side with no date |".
- `add-match-rule` SKILL.md (as S11-09 leaves it):
  - Front matter `description`: "(core/valuecompare, core/propertycompare, core/matchregistry)" becomes "(core/valuecompare, core/propertycompare, core/subjectcompare, core/matchregistry)".
  - "Rules that hold everywhere", a bullet *(new)* after S11-09's layer 2 bullet: "**Layer 3 (`core/subjectcompare`) turns outcomes into points.** It reads only the lead properties in `subject.Weights` for the kind; agree earns the agree weight, partial agree weight × similarity, conflict loses the penalty and weighs it in similarity. Keys are `subjectcompare.SubjectType` and `PropertyKey`, written in the registry as `vocabkey` constants, never as string literals. Value type and cardinality come from the catalog as `Metas`, never from the registry."
  - A section *(new)* after "Add a knob", titled "Add a Subject property or pair": "1. Add the key to `core/vocabkey` (and `PropertyKeys`) if it has none; `TestVocabKeysMatchSeed` ties it to the seed. 2. Add its `Weights` (or a `Pools` / `Periods` entry under an existing lead) in `core/matchregistry/subject.go`, with a source comment. 3. Add the inventory rows (`subject.Weights.<kind>.<property>.Agree` and `.Conflict`, or the pair's row). 4. `TestSubjectSection` fails on a key that is not a `vocabkey` constant, a pair whose lead has no weights, and a pooled or period property with weights of its own. 5. Add a `TestCompare` row for it in `core/subjectcompare`."
  - "Check": the command becomes `go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/matchregistry/...`.
- Per the Definition of done: `completed.md` paragraph (newest first), map row status, this plan's status and **Landed** note.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-10-subject-compare.md` reports no errors; `go test ./core/matchregistry/ -run TestKnobInventoryMatchesRegistry` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Add layer 3, Subject compare, with the registry's subject section."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/subjectcompare/subjectcompare_test.go` `TestCompare` *(new)* | Rows AC-1.1 – AC-1.15, inputs and expectations as in §2 | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, AC-1.12, AC-1.13, AC-1.14, AC-1.15, D-3, D-4, D-5, D-9, D-12 |
| T-2 | Go | same, `TestCompare` | Rows AC-2.1 – AC-2.13; AC-2.11 runs both start orders | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, AC-2.10, AC-2.11, AC-2.12, AC-2.13, D-6, D-7, D-8 |
| T-3 | Go | `core/subjectcompare/invariants_test.go` `TestInvariants` *(new)* | 500 generated cases per invariant per kind (D-15) → AC-3.1 – AC-3.6 hold | AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.6, D-15 |
| T-4 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` | New row → `core/subjectcompare` pure and free of numbers | D-10, D-16 |
| T-5 | Go | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry`; `core/matchregistry/registry_test.go` `TestDefaultIsFresh` | Every `subject.Weights`, `subject.Pools`, `subject.Periods` field has a row; lifespan rows covered by `notYetBuilt`; a changed copy leaves the next `Default()` unchanged | D-12, D-13 |
| T-6 | Go | `core/matchregistry/subject_test.go` `TestSubjectSection` *(new)* | `Default().Subject` → every kind primary, every key a `vocabkey.PropertyKeys()` entry, every pair lead weighted, no pooled or period property weighted or paired twice | D-14 |
| T-7 | Go | `core/vocabkey/vocabkey_test.go` `TestPropertyKeys` *(new)*; `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` | The nine constants in declaration order; a changed slice leaves the next call unchanged; every one is a seeded Property | D-14 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestKnobInventoryMatchesRegistry` (a row key that does not match `registryKeys`), `TestLayerPackages` (a literal in `pairs.go` or a stray import), `TestDefaultProfileCardinality` (must not change: `core/match` is untouched).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] go test ./core/subjectcompare/... ./core/propertycompare/... ./core/valuecompare/... ./core/matchregistry/... ./core/vocabkey/... passes with no build tags
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without macOS every check but xcode-test passes, and CI runs the Xcode tests on the PR (this PR changes no Swift)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored; core/match and core/graphalign are unchanged
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none (no copy)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: comparison-layers.md §Layer 3, matching-model.md (stack table, glossary, inventory keys, worked example), add-match-rule skill, match-rules.mdc (STEP-2, STEP-3, STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — nothing reaches it until S11-13. The reviewer runs `go test ./core/subjectcompare/... -v` with no build tags and reads the `TestCompare` row names for AC-1.1, AC-1.2 and AC-1.7: these are the map's Check ("name 1.0 / 2.5; name + gender 1.0 / 3.1; birth vs death with matching dates ≈ 0.09"). Then `go test ./core/matchregistry/... -v` to see `TestSubjectSection`, `TestKnobInventoryMatchesRegistry` and the new `TestLayerPackages` rows pass.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S11-09 lands with names or behavior other than its plan's | Implementer step 2 and the first stop condition: read S11-09's **Landed** note before STEP-2 |
| A literal (`2` in a sum, `"gregorian"` is fine) fails the numbers guard | D-10; a doubled value is `w + w` (skill rule); the period builder reads only `DateEnd` fields |
| Float sums differ by order | D-4: ascending key order; tests compare within 0.0001 and AC-3.1 within 1e-9 |
| An AC similarity drifts because layer 1 changed after the planning base | §2 values were read from `valuecompare.Compare` at `95d1f34`; a drift is a stop condition, not a test edit |

**Stop and report if:**

- S11-09 has not landed, or its **Landed** note or code differs from §4's S11-09 rows: `propertycompare.Compare`'s signature, `Value`'s `Negative` / `Neutral`, `Result.Speaking.Result.Similarity`, the `Outcome*` or `Cause*` constants, `propertymeta.Meta`, or `datevaluetest.YMD`. Report each difference.
- `valuecompare.Compare` at your base gives a relation or similarity different from §2 for an AC pair (for example `1817` / range 1815 – 1820 is not `Contains` 0.55, or `J. Robins` / `James Robins` is not `Contains` 0.8): report the pair and the output; do not change the AC.
- `core/subjectcompare` cannot pass `TestLayerPackages` with only the imports D-10 lists.
- `registryKeys` produces a key for the subject section other than D-13's (for example a `With` or a slice index): report the keys it prints; do not change the guard.
- An invariant of US-3 fails under D-5 – D-8 as written: report the generated case (seed, index, both sides) and which invariant.
- `core/subjectcompare` already exists, or `vocabkey` already has `PropertyKeys`.
- Merging `main` conflicts in a paragraph STEP-7 rewrites, so keeping both sides' text is not possible.

## 12. Open questions

None.

## 13. PR

**Title:** S11-10 — Subject compare
**Body:** objective in one sentence; what shipped against each US; AC evidence (test names); deviations from the plan; manual check (the two `go test` runs of §10); `Plan: docs/deployment-plan/spike-11/prs/S11-10-subject-compare.md`.
