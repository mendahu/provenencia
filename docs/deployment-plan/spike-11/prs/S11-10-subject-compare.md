# S11-10 — Subject compare

**Status:** Landed in [#471](https://github.com/mendahu/provenencia/pull/471) (2026-10-18)
**Spike / slice:** Spike 11 · Slice 3 — Property, Subject and entity compare
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a new pure Go package and a registry section with no caller yet; nothing reaches the app)
**Board:** —
**Depends on:** S11-09 (landed [#468](https://github.com/mendahu/provenencia/pull/468); [plan](S11-09-property-compare.md) and its **Landed** note). It shipped as its plan says: `core/propertycompare` (`Compare`, `Value`, `Result`, `SpeakingPair`, the `Outcome*` and `Cause*` constants, private test helpers in `core/propertycompare/helpers_test.go`: `term`, `text`, `name`, `date`, `neg`, `neutral`, `ref`, `label`, `labels`, `dates`), `propertymeta.Meta`, `core/datevalue/datevaluetest` (`YMD`, `About`, `Phrase`), its `layerPackages` rows, and its STEP-9 text in `add-match-rule`, `match-rules.mdc`, `comparison-layers.md` §Layer 2 and `matching-model.md`. Branch from `main`
**Blocks:** S11-11
**Branch:** `s11-10-subject-compare`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md) (read only: the seed test of STEP-5 reads `seedTypes`, `seedProperties`, `allBindings()`)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Layer 3](../comparison-layers.md#layer-3-subject-comparison) (requirements, the two axes, conflict weighs its penalty, registry shape); [lifespan](../comparison-layers.md#lifespan-a-derived-property-for-people); [match registry](../comparison-layers.md#match-registry) (typed keys, data-model facts stay in the catalog); [`docs/matching-model.md`](../../../matching-model.md) stack table, glossary, worked example, knob inventory (`subject.` rows) and removed knobs; working rules "Sparse data first", "Registry numbers: carry over, then estimate", "Subject compare returns similarity (0–1) and evidence (points)", "One match registry", "One matching skill, grown with the layers", "Build, then migrate", "One guard, not one per package"
**Planned:** 2026-10-18 against `main` @ `95d1f34`; refreshed 2026-10-18 against `main` @ `847b5f0` (S11-09 landed; layer 1 changed only by `Stronger`, the relation order and `Relations()`, so §2's values hold)
**Review:** [`S11-10-subject-compare.review.md`](S11-10-subject-compare.review.md) — Fits with changes / Ready with changes → fixed, 2 rounds

**Landed:** None — the code is as planned.

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

Layer 3 of matching exists: `core/subjectcompare` says how alike two Subjects of one type are over the properties both record. It runs property compare (layer 2) once per property the registry lists for the type, pools sex at birth with gender, reads an event's `start_date` and `end_date` as one period when it has no `date`, and returns two numbers: a **similarity** from 0 to 1 (are they alike?) and **evidence** in points (how strong is the connection?), with each property's layer 2 result and points. The registry gains its `subject` section, keyed by `core/vocabkey`'s typed keys and constants, and a seed test ties it to the product vocabulary.

**Why now.** Slice 3 builds entity compare (S11-11) and lifespan (S11-12) on this layer, and S11-13 moves Promote onto it. "Build, then migrate" puts the layer in its own PR with its own tests before any caller moves.

**What stays.** `core/match` (`DefaultProfile`, `Score`, `Rank`, `poolPeriods`, `period`, `endFields`) and `core/graphalign/registry.go` (`PropertyScale`) are untouched; S11-13 moves Promote onto this layer and deletes them, and S11-16 deletes the graphalign registry. `subjectvocab`'s `TestDefaultProfileCardinality` stays while `match.DefaultProfile` exists. No catalog, wire, Swift or researcher-visible change.

## 2. User stories

All values are compared under `matchregistry.Default()` (`Subject` and `Value` sections) with `subjectcomparetest.Metas()` (D-11). Values are built with `propertycomparetest` (D-11): names with `namevaluetest.Western`, dates with `datevaluetest.YMD` (`1817` is `YMD(1817, 0, 0)`). "¬x" is a negative value x; "(n)" is a neutral value. "Sim" is `Result.Similarity` (within 0.0001), "ev" is `Result.Evidence` (within 0.0001). Layer 1 relations and similarities below were read from `valuecompare.Compare` at `95d1f34`; date displays from `datevalue.CompactDisplay`.

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
| AC-1.14 | person, metas with the `name` entry deleted: name `{James Robins}` / `{James Robins}` | `name` unknown, `Speaking` not nil, `Speaking.Result.Relation` `Unknown`, `Speaking.Detail.Cause` `unsupported`; sim 0, ev 0 |
| AC-1.15 | event: `event_name` `{Great Fire}` / `{Great Fire}` only | sim 0, ev 0; `Properties` = `date`, `event_type`, both unknown (an unlisted property is never read) |

**US-2.** As the engine, the properties the registry pairs are compared as one: sex at birth with gender, and an event's `date` with its `start_date` / `end_date`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | person: gender `{female}` / sex at birth `{female}` | `sex_at_birth` agree: sim 1, ev 0.6 |
| AC-2.2 | person: sex at birth `{male}` / gender `{female}` | `sex_at_birth` conflict: sim 0, ev −1 |
| AC-2.3 | person: sex at birth `{unknown (n)}` / `{male}` | `sex_at_birth` unknown (`no_values`), `Excluded` the `unknown` value; sim 0, ev 0 |
| AC-2.4 | person: sex at birth `{¬male}` / gender `{male}` | `sex_at_birth` conflict, cause `negative`: ev −1 |
| AC-2.5 | event: date `{1817}` / start `{1815}`, end `{1820}` | `date` partial: `CompactDisplay(*Speaking.B.Date)` is `Between 1815 and 1820`, `Speaking.B.Ref` nil, relation `Contains` 0.55; sim 0.55, ev 1.1 |
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
| AC-3.4 | Any generated case; pick a lead whose group (the lead, its pooled properties, its period start and end) records nothing on side B, and add one positive value to a property of that group on side A | Similarity, evidence and each property's outcome unchanged |
| AC-3.5 | Any generated case where a lead's group records nothing on either side; give both sides the same one value for the lead (name `James Robins`, sex at birth `male`, event type `birth`, date `1817`, toponym `York`) | Evidence does not go down, similarity does not go down |
| AC-3.6 | Any generated person case with name exactly `{James Robins}` / `{Mary Robins}` (positive, nothing else on name); any event case with event type exactly `{birth}` / `{death}` | Evidence < 0, whatever the other properties say |

**US-4.** As a contributor, the registry's subject section cannot drift from the product vocabulary, and a later layer reuses one set of builders.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `matchregistry.Default().Subject` against the seed | Every kind is a seeded type; every (kind, property) the section names is in `allBindings()` (the seed's bindings plus the product bridge bindings Install adds); a pooled property has its lead's seeded value type; a period's lead, start and end are seeded `date` |
| AC-4.2 | `subjectcomparetest.Metas()` against the seed | Each entry's value type and cardinality equal the seeded Property's (empty seeded cardinality reads `single`) |
| AC-4.3 | `matchregistry.Default().Subject` | Each `Pools` and `Periods` lead has a `Weights` entry for its kind; no pooled property, `Start` or `End` has a `Weights` entry of its own; no property is in more than one pair of its kind, and no lead has both a pool and a period |
| AC-4.4 | `valuecompare.EndsDate(DateEnds(d))` for `1817`, `3 Mar 1817`, `BEF 1820`, `AFT 1815`, range 1815 – 1820 | Returns a value whose `DateEnds` equal the input's ends; both ends open → not ok |

## 3. Scope

| In | Covered by |
| --- | --- |
| `core/valuecompare` tests build dates with `datevaluetest`; their private builders go (D-19; S11-09 review finding MNT-3) | STEP-1 |
| `valuecompare.EndsDate`, the inverse of `DateEnds` (D-8) | STEP-2, AC-4.4 |
| `vocabkey.SubjectType`, `vocabkey.PropertyKey` (D-2) | STEP-3 |
| `propertycompare.Value.Positive`, layer 2's one rule for a positive value (D-18) | STEP-4 |
| `core/subjectcompare`: `Rules`, `Compare`, similarity and evidence (D-1, D-3 – D-5, D-9, D-10) | STEP-3, STEP-6, US-1 |
| `propertycomparetest` builders (S11-09's private helpers move onto them) and `subjectcomparetest.Metas` (D-11) | STEP-4, AC-4.2 |
| Registry `subject` section; inventory rows renamed to the built keys; lifespan weight rows renamed for S11-12 (D-12, D-13) | STEP-5, AC-1.1 – AC-1.9 |
| Registry checks: structure in `matchregistry`, seed in `subjectvocab` (the port of `TestDefaultProfileCardinality`) (D-14) | STEP-5, AC-4.1 – AC-4.3 |
| Pools (sex with gender) and periods (`date` with `start_date` / `end_date`) (D-6 – D-8) | STEP-7, US-2 |
| Exact table and invariant tests (D-11, D-15) | STEP-6 – STEP-8, US-1 – US-3 |
| Guard rows; `match-rules.mdc` glob and check line (D-16) | STEP-3, STEP-4 |
| Docs the layer makes true or false: `comparison-layers.md` §Layer 3, `matching-model.md`, the `add-match-rule` skill (D-17) | STEP-5, STEP-9 |

| Out | Home |
| --- | --- |
| Any caller: Promote scoring, `match.Score`, `Rank`, `Evaluate`, `alikeNew`, `pairNode` | S11-13 |
| Deleting `match.DefaultProfile`, `match.DatePeriods`, `poolPeriods`, `period`, `endFields` and `subjectvocab`'s `TestDefaultProfileCardinality` | S11-13 (deletes `core/match`) |
| Deleting `graphalign` `PropertyScale` and the rest of `core/graphalign/registry.go` | S11-16 |
| Profiles from reconciler output; `Excluding` | S11-11 |
| Lifespan: its derived key, its window rules and their rows, its profile-builder values | S11-12 |
| Evidence lines that name the Observations behind a pooled period (refs of `start_date` / `end_date`) | S11-19 |
| An accept bar, bands, a 0–1 confidence over the total | S11-16 (bands, accept bar); confidence not planned (labeler idea, [`docs/ideas/matching-test-suite.md`](../../../ideas/matching-test-suite.md)) |
| Moving the other private date builders (`core/autoreconcile/dates_test.go` `ymd`, `core/database/conclusionheaders/placechains_internal_test.go` `ymd`, `placechains_test.go` `year`, `core/database/promote/batch_test.go` `yearDate`) onto `datevaluetest` | Not planned — their signatures differ; each moves when its file is next edited (S11-09 Out) |
| Weights for `start_date` / `end_date` on places and place relationships | Not planned — the inventory lists none; places match on toponym and `part_of` (S11-14) |
| One test command for every layer package, instead of a list each layer PR extends | Logged in the review log (M-8) for S11-11's planner |
| Tuning any number beyond the inventory's start values | Not planned — working rule "Registry numbers: carry over, then estimate" |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Layer 3 package | `core/subjectcompare` | Does not exist |
| Layer 2 | `core/propertycompare/propertycompare.go`: `Compare(a, b []Value, meta propertymeta.Meta, rules valuecompare.Rules) Result`; `Value{valuecompare.Value; Negative, Neutral bool}`; `Result{Outcome, Cause, Speaking *SpeakingPair, Excluded}`; `SpeakingPair{A, B Value; Result; Detail}`; `OutcomeAgree`, `OutcomePartial`, `OutcomeConflict`, `OutcomeUnknown` | Sets every value's `Type` from `meta.ValueType` (S11-09 D-4); an `Unknown` pair still speaks (D-8); `Speaking.A` / `B` are the input values with `Type` set (D-11) |
| Layer 2 test helpers | `core/propertycompare/helpers_test.go`: `term`, `text`, `name`, `date`, `neg`, `neutral`, `ref`, `label`, `labels`, `dates`, and a fixture `var` block (package `propertycompare_test`) | Private to one external test package; S11-10, S11-11, S11-12 and S11-14 need the same builders |
| Property facts | `core/propertymeta/propertymeta.go` `Meta{ValueType, Cardinality}` | Empty cardinality reads as single |
| Registry | `core/matchregistry/registry.go` `Registry{Value valuecompare.Rules}`, `Default()`; `core/matchregistry/registry_test.go` `TestDefaultIsFresh` (changes `Particles[0]`, a map entry, then compares a new `Default()`) | One section. The design file list puts the subject section in `subject.go` |
| Knob inventory | `docs/matching-model.md` §Knob inventory: the intro paragraph; rows `subject.person.name.Agree` … `subject.place.toponym.Conflict` (ten weight rows), `subject.Pairs.event.date`, `subject.Pairs.person.sex_at_birth`, `subject.person.lifespan.*` (seven rows); §Worked example steps 3 – 4; §Removed knobs rows for `graphalign` `PropertyAgreement`, `DefaultConfig.ConflictPenalty` and `DefaultProfile` `start_date` / `end_date` features | Keys written before the section's Go shape existed: a field (`Pairs`) and kind map keys (`person`) at the same level, which no Go type can produce (D-13) |
| Inventory guard | `core/matchregistry/inventory_test.go` `notYetBuilt` (`"subject."` first), `registryKeys`, `TestKnobInventoryMatchesRegistry` | `registryKeys` lowercases the top field, walks a map only when its element is a struct or a map (by sorted key), and stops at any other leaf, so a map of slices is one leaf |
| Purity guard | `core/matchregistry/guard_test.go` `layerPackages` | After S11-09: `core/valuecompare` and `core/propertycompare` with numbers true, the pure value and fixture packages false |
| Keys | `core/vocabkey/vocabkey.go`: untyped constants `Person`, `Event`, `Place`, `PropertyName`, `PropertySexAtBirth`, `PropertyGender`, `PropertyEventType`, `PropertyDate`, `PropertyStartDate`, `PropertyEndDate`, `PropertyToponym`, `PropertyLikeness`; `PrimaryKinds()` | No key types. The map's working rule names `SubjectType` and `PropertyKey` as the registry's typed keys; S11-02 D-14 left them to S11-10 |
| Seed | `core/database/subjectvocab/registry.go` `seedTypes` (`Key`), `seedProperties` (`Key`, `ValueType`, `Cardinality`), `seedBindings` (`TypeKey`, `PropertyKey`); `core/database/subjectvocab/subjectvocab.go` `allBindings()` (`seedBindings` plus product bridge bindings from `connectrules.BridgeBindings()`, read by `LockedBinding`) | Person binds `name`, `sex_at_birth`, `gender`; event binds `event_type`, `event_name`, `date`, `start_date`, `end_date`; place binds `toponym`; `toponym` is the only multiple one of these |
| Old profile check | `core/database/subjectvocab/profile_test.go` `TestDefaultProfileCardinality` | Checks `match.DefaultProfile` features against the seed's value type and cardinality |
| Old Subject scoring | `core/match/match.go` `Score` (Weight × similarity, −Contradiction at 0); `core/match/profiles.go` `Feature`, `Profile`; `core/match/registry.go` `DefaultProfile`, `DatePeriods` | `Rank`'s 10-point scale; `start_date` / `end_date` scored as their own features |
| Old periods | `core/match/periods.go` `poolPeriods`, `pools`, `period`, `endFields`, `firstLeader` | Pools only when the other side has a voting `date` and no start or end; picks the first leader by support; converts `DateEnd`s back to a date privately |
| Promote's weights | `core/graphalign/registry.go` `DefaultConfig` `PropertyScale` (name 2.5 / 3, sex and gender 0.6 / 1) | The source of the person numbers (D-12) |
| Date ends | `core/valuecompare/datewindow.go` `DateEnd{Year, Month, Day, Lo, Hi, From, To, Mid, Open}`, `DateEnds(d) (first, last DateEnd, ok bool)`; `core/valuecompare/datewindow_test.go` `TestDateEnds` | `DateEnds` returns an open first end for `BEF` and an open last end for `AFT`; no inverse exists outside `match` |
| Layer 1 date fixtures | `core/valuecompare/date_test.go` `optInt`, `ymd`, `about`, `bound`, `between`, `phrase`, `noYearRange`, `dp` (package `valuecompare_test`), used by `date_test.go`, `contract_test.go` and `datewindow_test.go`; `optInt` is also called directly in about ten inline `datevalue.Value` literals (`date_test.go` AC-1.13, `datewindow_test.go` `TestDateEnds`, `TestDateWindow`) | `ymd`, `about`, `phrase` are exact copies of `datevaluetest.YMD`, `About`, `Phrase` but return `datevalue.Value`, wrapped with `dp(…)` where a pointer is needed. `datevaluetest`'s own `optInt` is unexported |
| Fixtures | `core/namevalue/namevaluetest/namevaluetest.go` `Western`; `core/datevalue/datevaluetest`: `YMD`, `About`, `Phrase` | Pure; listed in `layerPackages` |
| Docs | `docs/deployment-plan/spike-11/comparison-layers.md` §Layer 3; `docs/matching-model.md` stack table row 3, glossary **pair (compared together)**; `.agents/skills/add-match-rule/SKILL.md` (as S11-09 leaves it); `.cursor/rules/match-rules.mdc` | §Layer 3's table reads 0.81 for "name agrees, sex conflicts", written before "a conflict weighs its penalty in similarity" was decided; with the penalty it is 0.71 (AC-1.3) |

**Precedent.** S11-09 (`core/propertycompare`) and [#441](https://github.com/mendahu/provenencia/pull/441) (S11-02, `core/valuecompare` with the registry's `value` section). Copy from S11-02: the layer declares its `Rules` types with doc comments and no numbers, `core/matchregistry/<section>.go` sets the values with a source comment on each, `Registry` gains one field, and `notYetBuilt` loses the built prefix. Copy from S11-09: the package doc naming the layer and its purity, an external `_test` package reading `matchregistry.Default()`, a `layerPackages` row with numbers true, a fixture package beside its value package (`datevaluetest` beside `datevalue`), and a fixed-seed invariants test. Copy from S11-23 (`TestVocabKeysMatchSeed`): a `subjectvocab` test that reads the seed slices directly and checks pure keys against them.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Package and entry.** `core/subjectcompare`, files `subjectcompare.go` (types, `Compare`) and `pairs.go` (`side`, `period`), plus tests. `func Compare(kind vocabkey.SubjectType, a, b Values, metas Metas, rules Rules, value valuecompare.Rules) Result`. Types (all *new*): `Weights`, `Pool`, `Period`, `Rules`, `Values`, `Metas`, `Result`, `PropertyResult` (sketch in §6) | One entry like layers 1 and 2. Layer 3 reads its `subject` section and passes `value` to layer 2 unread, as layer 2 passes it to layer 1. `kind` is one argument because both Subjects are of one type | A `Subject{Type, Values}` struct per side — the caller would have to keep two types equal; one combined rules argument — layer 2 already takes `valuecompare.Rules` alone |
| D-2 | **Typed keys live in `vocabkey`.** `type SubjectType string` and `type PropertyKey string` *(new)* in `core/vocabkey/vocabkey.go`, with doc comments "SubjectType is a Subject type key, typed for registry maps." and "PropertyKey is a Property key, typed for registry maps.". The existing constants stay untyped, so they convert into either type with no copy, and every caller that uses them as `string` is unchanged. `subjectcompare` and every later registry section (segment, alignment, evidence) key their maps by these types; no other package declares a key type | `vocabkey` is the one home for key strings (S11-23); the segment, alignment and evidence sections key by kind too, and must not grow a second `SubjectType` or depend on layer 3 for a key type | Declaring them in `subjectcompare` — S11-19's evidence policy would import layer 3 for a type; typing the constants — every string caller of `vocabkey` would need a conversion |
| D-3 | **Catalog facts come in as `Metas`.** `type Metas map[vocabkey.PropertyKey]propertymeta.Meta`, filled by the caller (S11-13's loader, the tests' `subjectcomparetest.Metas()`). A lead property with no entry is compared with the zero `Meta`: layer 2 sets `Type` to `""`, and `valuecompare.Compare` speaks `Unknown` / `unsupported` (AC-1.14). Pooled and period properties are compared under their lead's `Meta` | The registry rule: value type and cardinality are data-model facts that stay in the catalog; the registry names properties by key only. No special case | Value type and cardinality in `Weights` — duplicates the seed; skipping a property with no meta — hides a caller bug |
| D-4 | **Which properties.** `Compare` reads only the lead properties in `rules.Weights[kind]`, in ascending key order (`slices.Sorted(maps.Keys(rules.Weights[kind]))`), and returns one `PropertyResult` for each, unknown ones included. A kind with no `Weights` entry returns the zero `Result` (AC-1.13). Values for properties the registry does not reach are never read (AC-1.15) | "One registry defines every Subject type, its properties, and how each one counts." Sorted order makes the float sums and `Properties` deterministic; unknown rows let S11-19 say what was missing | Iterating the map — a nondeterministic sum order; comparing every property both sides have — a property with no registry weight would count for nothing anyway |
| D-5 | **Points and similarity, from the outcome.** For lead `p` with weights `w` and layer 2 result `r`: agree → points `w.Agree`, counts `w.Agree` × 1 in similarity; partial → points `w.Agree × s`, counts `w.Agree` × `s`, where `s` = `r.Speaking.Result.Similarity`; conflict → points `−w.Conflict`, counts `w.Conflict` × 0; unknown → points 0, not counted. `Evidence` = Σ points. `Similarity` = Σ (weight × score) / Σ weight over counted properties, 0 when nothing is counted. An agree earns the full weight whatever its relation (`Equivalent` 0.95 earns 3 on a toponym, AC-1.9) | §Layer 3 "decided": two axes, a conflict weighs its penalty in similarity (birth vs death with matching dates ≈ 0.09). Layer 2's outcome is the verdict; layer 1's similarity grades only a partial. The glossary's **agree weight** is "points a property earns when it agrees" | Agree × similarity on agree — an accent fold would earn less than an exact match though layer 2 says both agree; counting unknown as 0 in similarity — "a property on one side only neither helps nor hurts" |
| D-6 | **Pools.** `rules.Pools[kind][lead]` is a `Pool{With []vocabkey.PropertyKey}`: on each side, the values passed to layer 2 are that side's lead values, then its values for each `With` property in list order, all with their flags. Default: person `sex_at_birth` with `[gender]` | §Layer 3: "both sides' sex and gender values are pooled into one direct comparison" (AC-2.1). Layer 2's outcome is order-independent (S11-09 AC-5.3). A struct element, not a bare slice, because `registryKeys` stops at a slice one level up: `Pool` gives the row `subject.Pools.person.sex_at_birth.With` | A bare `[]PropertyKey` element — its inventory key would be `subject.Pools.person`, one row per kind; pool and period fields on `Weights` — `registryKeys` would demand a `Pool` and a `Period` row for every weighted lead, empty ones included (ten extra rows); a sex-only and a gender-only comparison — the two would never meet |
| D-7 | **Periods.** `rules.Periods[kind][lead]` is a `Period{Start, End vocabkey.PropertyKey}` that stands for the lead on a side with **no positive lead value** (`propertycompare.Value.Positive`, D-18): that side's values are its lead values (negatives and neutrals included) plus one built period value, when one can be built. The other side decides alone. Default: event `date` with `Start` `start_date`, `End` `end_date`. A lead has a pool or a period, never both (AC-4.3) | §Layer 3: "an event's `date` against another's `start_date` + `end_date`, read as one period". Each side decides alone, so the rule is symmetric (AC-3.1) and two start / end-only events compare period with period (AC-2.6). Keeping negatives lets a denied date still contradict (AC-2.12) | `match.poolPeriods`' rule (pool only when the other side has a date and no start or end) — asymmetric, and two start / end-only events would not compare once start and end lose their own weights; building a period on a side that has a date — its own date says more than its span |
| D-8 | **Building the period.** `period` (unexported, `pairs.go`) reads the side's values of `Start` and `End` for which `Positive()` is true and that have a non-nil `Date`. A start candidate is `first` of `valuecompare.DateEnds(*v.Date)` with `ok` and not `first.Open`; the start used has the lowest `Lo`, then the lowest `Hi`. An end candidate is `last` with `ok` and not `last.Open`; the end used has the highest `Hi`, then the highest `Lo`. No start and no end → no period. Otherwise, with an absent start or end as `valuecompare.DateEnd{Open: true}`, `d, ok := valuecompare.EndsDate(start, end)`; not ok → no period (AC-2.10). The built value is `propertycompare.Value{Value: valuecompare.Value{Date: &d}}` (no `Ref`, not `Negative`, not `Neutral`; layer 2 sets `Type`). **`valuecompare.EndsDate(first, last DateEnd) (datevalue.Value, bool)`** *(new, `datewindow.go`)* is the inverse of `DateEnds`: both closed and `first.Lo` ≤ `last.Hi` → `Kind` `datevalue.KindRange`, start fields from `first`, end fields from `last`; both closed and `first.Lo` > `last.Hi` → not ok; first open, last closed → `KindPoint`, `Qualifier` `datevalue.QualifierBEF`, start fields from `last`; last open, first closed → `KindPoint`, `datevalue.QualifierAFT`, start fields from `first`; both open → not ok. `Calendar` is `"gregorian"`. A `Month` or `Day` of 0 is left nil | The widest span the side records, picked by day number, so input order never matters (AC-2.11). One place turns ends back into a date: `match.period` / `endFields` go in S11-13 with nothing left to copy, and S11-12's lifespan builder (another package) reuses it | A private `endFields` copy in `subjectcompare` — a second copy of the conversion, unreachable from S11-12; `match.firstLeader` — reads support; one period per start × end combination — several made-up values for one fact |
| D-9 | **`Result` shape.** `Result{Similarity, Evidence float64; Properties []PropertyResult}`; `PropertyResult{Key vocabkey.PropertyKey; Result propertycompare.Result; Points float64}`. `Key` is the lead property. `Properties` is nil only when the kind has no `Weights` entry | Rule 4, "output is the explanation": the layer above and S11-19 read which values spoke, why a property was unknown, and what it added | A map of results — no order; similarity per property — layer 2's speaking pair already carries it |
| D-10 | **Purity and numbers.** `core/subjectcompare` non-test files import only the standard library, `core/propertycompare`, `core/valuecompare`, `core/propertymeta` and `core/vocabkey`; they hold no numeric literal other than 0 and 1 | Working rules "Pure means the package builds and tests with no build tags" and "One guard": a `layerPackages` row with numbers true | — |
| D-11 | **Tests and fixtures.** **`core/propertycompare/propertycomparetest`** *(new)*, package doc "Package propertycomparetest builds property compare values for test fixtures.": `Term(key string) propertycompare.Value`, `Text(s string)`, `Name(form string)` (`namevaluetest.Western`), `Date(d *datevalue.Value)`, `Neg(v propertycompare.Value) propertycompare.Value`, `Neutral(v)`, `Ref(v, b byte)`, `Label(v) string`, `Labels(vs []propertycompare.Value) string`, `Dates(ds ...*datevalue.Value) []propertycompare.Value` — the bodies of S11-09's private `term`, `text`, `name`, `date`, `neg`, `neutral`, `ref`, `label`, `labels` and `dates`; the fixture `var` block (`y1817`, `y1818`, `y1830`, `about1817`, `spring`) stays in `core/propertycompare`'s tests, in a new `fixtures_test.go`, while `core/propertycompare/helpers_test.go` loses is deleted (the `propertycompare_test` files call the exported names). **`core/subjectcompare/subjectcomparetest`** *(new)*, package doc "Package subjectcomparetest holds Subject compare fixtures.": `Metas() subjectcompare.Metas`, a fresh map: `name` name single; `sex_at_birth`, `gender`, `event_type` term single; `date`, `start_date`, `end_date` date single; `toponym` text multiple; `event_name` text single (cardinality written as the constant, `propertymeta.CardinalitySingle` included). Tests: external package `subjectcompare_test`, table-driven, no build tags, rules `matchregistry.Default()`. `TestCompare` rows are named by their AC (`"AC-1.1 name agrees"`), with fields `kind`, `a`, `b` (`subjectcompare.Values`), `metas` (nil → `Metas()`), `sim`, `ev`, and a `check func(t *testing.T, r subjectcompare.Result)` for the property-level assertions the AC names (outcome, cause, points, speaking relation, `CompactDisplay` of a speaking date) | S11-11, S11-12 and S11-14 build the same values and need the same metas; one builder set instead of four (the `datevaluetest` precedent). AC-4.2 pins `Metas()` to the seed, so the fixture cannot drift | Private helpers per package — S11-09's choice for one package, a fourth copy by S11-14; metas written in each test file — unguarded copies of seed facts |
| D-12 | **Registry `subject` section.** `Registry.Subject subjectcompare.Rules` *(new field)*, set in `core/matchregistry/subject.go` *(new)* `defaultSubject()`. Weights: person `name` 2.5 / 3 and `sex_at_birth` 0.6 / 1 (`graphalign` `PropertyScale`); event `event_type` 0.5 / 20 and `date` 2 / 3; place `toponym` 3 / 3 (the inventory's estimates, its reasoning as the source comment). Pools: person `sex_at_birth` `With` `[gender]`. Periods: event `date` `{start_date, end_date}`. Every key is a `vocabkey` constant. `Default()` builds fresh maps and slices | The knob inventory's start values (working rule "Registry numbers: carry over, then estimate"); the section file the design names | `Rank`'s 10-point scale — retired ("one weight set") |
| D-13 | **Inventory keys follow the Go shape.** The built keys are `subject.Weights.<kind>.<property>.Agree` / `.Conflict`, `subject.Pools.person.sex_at_birth.With`, `subject.Periods.event.date.Start` and `.End`. STEP-5 renames the ten weight rows and turns the two `Pairs` rows into three (exact text there), and fixes every other place `matching-model.md` names an old `subject.` key. **Lifespan is a derived lead:** its two weight rows become `subject.Weights.person.lifespan.Agree` and `.Conflict` now; its five window rows (`MaxYears`, `ParentMinYears`, `ParentMaxYears`, `MotherMaxYears`, `MarriageMinYears`) keep `subject.person.lifespan.` for S11-12 to place. `notYetBuilt`'s `"subject."` becomes two entries, `"subject.Weights.person.lifespan."` and `"subject.person.lifespan."` | The guard is the source of truth for keys; the old rows mix a field and kind keys at one level, which no Go type produces. The S11-12 map row calls lifespan "a people-only pair with a large conflict penalty and near-zero agree weight" — a `Weights` entry. S11-12 then adds a registry entry, a meta and a builder value, with no second scoring path in `Compare` | A custom walk in `registryKeys` — a special case in the one guard; leaving the lifespan weights under the old shape — two key shapes in the worked example and a second scoring path in S11-12 |
| D-14 | **Registry checks, two places.** **Structure** — `core/matchregistry/subject_test.go` `TestSubjectSection` *(new)*: AC-4.3 on `Default().Subject`. **Seed** — `core/database/subjectvocab/subject_section_test.go` *(new)* `TestSubjectSectionMatchesSeed`, the port of `TestDefaultProfileCardinality`: AC-4.1 on `matchregistry.Default().Subject` (kinds against `seedTypes` keys; every lead, `With`, `Start` and `End` against `allBindings()` of that kind — `seedBindings` plus the provenencia-origin `connectrules.BridgeBindings()`, the set Install materializes; value types from `seedProperties`), and AC-4.2 on `subjectcomparetest.Metas()` against `seedProperties`. `TestDefaultProfileCardinality` stays until S11-13 | The seed is the source of truth for which properties a kind has and their types; checking it catches `toponym` weighted on `person` or a pooled term beside a text, which a key-list check cannot. Reading the seed needs `core/database`, so the seed half lives in `subjectvocab` (impure imports pure), as `TestDefaultProfileCardinality` and `TestVocabKeysMatchSeed` do. Bridge kinds that S11-14 weighs are seeded types, so the check needs no change for them | "Every registry key is a `vocabkey` constant" with a new `vocabkey.PropertyKeys()` (the map's first wording) — weaker, and S11-14's bridge kinds and properties are not `vocabkey` constants; the whole check in `matchregistry` — cannot read the seed |
| D-15 | **Generator.** `rand.New(rand.NewSource(1))`; `TestInvariants` has one subtest per AC and kind (`AC-3.1/person`, `AC-3.1/event`, `AC-3.1/place`, and the same for each AC; AC-3.6 runs person and event only), 500 cases each. Pools: name `James Robins`, `James Robbins`, `J. Robins`, `Mary Robins`; sex at birth and gender `male`, `female`, `unknown (n)`; event type `birth`, `death`, `baptism`; date `1817`, `1818`, `1830`, `ABT 1817` (`datevaluetest.About`); start `1815`, `1816`; end `1820`; toponym `York`, `york`, `Montréal`, `Montreal`, `London`. Each side draws 0 – 2 pool values for each of its kind's properties (person: `name`, `sex_at_birth`, `gender`; event: `event_type`, `date`, `start_date`, `end_date`; place: `toponym`); draws are with replacement, so one side may hold the same value twice; a drawn value is negative when `r.Intn(4) == 0`. AC-3.4 picks with `r.Intn` a lead whose group is empty on side B and adds one pool value that is neither negative nor neutral to a property of that group on side A; AC-3.5 picks a lead whose group is empty on both sides and adds the AC's value to both; a case with no such lead is drawn again and not counted. AC-3.6 sets the name or event type exactly as the AC says and draws the rest. A failure reports `t.Errorf("case %d: a=%v b=%v: %s", i, labelsA, labelsB, what)`, where a side's labels are `property=label,label` per property, joined with `; ` in ascending key order (`propertycomparetest.Labels` for each property). AC-3.5 compares with a 1e-9 slack with both sides' labels (`propertycomparetest.Label`) | The S11-09 generator shape; fixed pools cover agree, partial, conflict, unknown, negatives, neutrals, pools and periods | Random strings — would rarely relate |
| D-16 | **Guard and editor rule.** `layerPackages` gains `{"core/subjectcompare", true}` after `core/propertycompare`, `{"core/propertycompare/propertycomparetest", false}` after `core/propertycompare`, and `{"core/subjectcompare/subjectcomparetest", false}` after `core/subjectcompare`. `.cursor/rules/match-rules.mdc`: `globs` gains `core/subjectcompare/**`; the Check line becomes "`go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/matchregistry/...` with no tags." | One guard; fixture packages are pure rows with numbers false, as `datevaluetest`; the editor rule loads for the new layer | — |
| D-17 | **Docs.** `comparison-layers.md` §Layer 3 gains a **Built in S11-10** paragraph and its "name agrees, sex conflicts" similarity becomes 0.71; `matching-model.md`'s stack table row 3 passes `value` through, the glossary's **pair (compared together)** row names pools and periods, and every old `subject.` key is renamed (STEP-5); the `add-match-rule` skill gains layer 3 (description and trigger, a rule bullet, an "Add a Subject property or pair" recipe, Check). Exact text in STEP-5 and STEP-9 | The design doc and the model page are kept true as layers land; the map's working rule "One matching skill" names S11-10 | — |
| D-18 | **One rule for "positive".** `func (v Value) Positive() bool` *(new)* on `propertycompare.Value` in `core/propertycompare/propertycompare.go`, doc comment "Positive reports whether v speaks for the property: neither negative nor neutral.", returns `!v.Negative && !v.Neutral`. `propertycompare.Compare`'s own checks for a positive value (the split after dropping neutrals, and "a side with no positive value") call it, with no change in behavior; `subjectcompare`'s `period` and D-7's lead test call it | Layers 2 and 3 must agree on which values are evidence for a property; S11-11's profile builder is the next writer of these flags, and a private copy in layer 3 would drift from layer 2's verdict | Writing `!v.Negative && !v.Neutral` in `pairs.go` — a second copy of layer 2's rule |
| D-19 | **Layer 1 tests build dates with `datevaluetest`.** `core/datevalue/datevaluetest` gains `Part(n int) *int` *(new)*, its `optInt` exported and renamed, doc "Part is a date part for a datevalue.Value literal: nil for 0." (`YMD` calls it), and `Between(y1, m1, d1, y2, m2, d2 int) *datevalue.Value` *(new)*, the body of `valuecompare_test`'s `between` (a Gregorian range; 0 parts unset). In `core/valuecompare`'s tests: `optInt`, `ymd`, `about`, `phrase`, `between` and `dp` are deleted; `bound` and `noYearRange` stay private and return `*datevalue.Value` (`bound` from `datevaluetest.YMD`, `noYearRange` with `datevaluetest.Part`). Call sites: `dp(ymd(…))` → `datevaluetest.YMD(…)`, `dp(about(…))` → `datevaluetest.About(…)`, `dp(phrase(…))` → `datevaluetest.Phrase(…)`, `dp(between(…))` → `datevaluetest.Between(…)`, `dp(bound(…))` → `bound(…)`, `dp(noYearRange())` → `noYearRange()`, `dp(datevalue.Value{…})` → `&datevalue.Value{…}`; where a value is needed (`DateEnds`, `DateWindow`, `AboutEnds`, a field of type `datevalue.Value`) the call is dereferenced, `*datevaluetest.YMD(…)`; `optInt(n)` in a literal → `datevaluetest.Part(n)`. No test case, name or expectation changes | The S11-09 review (MNT-3, in its ship log) found the copies; `datevaluetest` is the fixture home and already a pure `layerPackages` row, and `valuecompare_test` imports it with no cycle. Exporting `Part` lets the inline literals drop `optInt` instead of keeping a private copy beside the fixture's; `Between` is the one builder the package lacks | Keeping a private `optInt` for the inline literals — a second copy of the fixture's helper; moving `bound` and `noYearRange` too — one caller package each, and `bound`'s qualifier argument has no fixture shape yet |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/subjectcompare` *(new)* | package | `core/match` `Score`, `Profile` | `match` is impure (imports `autoreconcile`), reads support, and is deleted in S11-13 ("Build, then migrate") |
| `subjectcompare.Compare` *(new)* | func | `match.Score`, `graphalign` `ApplyScale` | Replaces both in S11-13; reads layer 2 outcomes, not a similarity number with a floor |
| `subjectcompare.Rules`, `Weights`, `Pool`, `Period` *(new)* | type | `match.Profile`, `match.Feature`, `graphalign.Scale`, `match.PeriodProperties` | Same concepts on the retired scale; the registry rule puts a section's types in its layer |
| `vocabkey.SubjectType`, `vocabkey.PropertyKey` *(new)* | type | `match.Property{Key, Origin}` | The registry's typed keys, in the one key package (D-2); matching reads product properties only, so the origin is the caller's filter (S11-13) |
| `subjectcompare.Values`, `Metas` *(new)* | type | `match.Values`, `match.PropertyMeta` | `match.Values` carries reconciler fields; `Metas` holds S11-09's pure `propertymeta.Meta` |
| `subjectcompare.Result`, `PropertyResult` *(new)* | type | `match.Match`, `match.Reason` | Two axes and layer 2's result per property, where `Match` has one score |
| `side`, `period` *(new, unexported, `pairs.go`)* | func | `match.poolPeriods`, `match.period` | Symmetric, with no support (D-7, D-8); the date conversion moves to `valuecompare.EndsDate` |
| `propertycompare.Value.Positive` *(new)* | method | the inline checks in `propertycompare.Compare` (S11-09) | Names layer 2's rule so layer 3 calls it (D-18) |
| `datevaluetest.Part`, `datevaluetest.Between` *(new)* | func | `valuecompare_test` `optInt`, `between` | The same helpers, made importable (D-19); the private copies are deleted |
| `valuecompare.EndsDate` *(new)* | func | `valuecompare.DateEnds`; `match.period` / `endFields` | The inverse belongs beside `DateEnds`; `match`'s private copy goes in S11-13 |
| `matchregistry.Registry.Subject`, `defaultSubject` *(new)* | field, func | `Registry.Value`, `defaultValue` | The second section, the same way |
| `core/propertycompare/propertycomparetest` *(new)* | package | `core/propertycompare/helpers_test.go` (S11-09); `namevaluetest`, `datevaluetest` | The same builders, made importable; the private file is deleted |
| `core/subjectcompare/subjectcomparetest` `Metas` *(new)* | package, func | `match.DefaultProfile` feature value types | A fixture of seed facts, checked against the seed (AC-4.2) |
| `TestSubjectSection`, `TestSubjectSectionMatchesSeed` *(new)* | test | `TestWesternPatternCoversEveryPartType`; `TestDefaultProfileCardinality`, `TestVocabKeysMatchSeed` | The new section's structure and seed checks; the old profile test leaves with `match` |

**Builds on.** `propertycompare.Compare` and its result (S11-09), `propertymeta.Meta`, `valuecompare.DateEnds`, the `vocabkey` constants (S11-23), the registry section pattern (S11-02), the `layerPackages` and inventory guards, `namevaluetest`, `datevaluetest`, the `add-match-rule` skill.
**Deletes.** `core/propertycompare/helpers_test.go` (S11-09's private builders, now `propertycomparetest`); `valuecompare_test`'s `optInt`, `ymd`, `about`, `phrase`, `between` and `dp` (D-19). Nothing else becomes dead: `core/match` and `graphalign` keep their callers until S11-13 and S11-16.
**Next adds.** S11-11 passes a profile built from the conclusion as one side's `Values` and builds its tests with `propertycomparetest` and `subjectcomparetest.Metas`; this layer does not change. S11-12 adds lifespan as a derived lead: a key constant for it, its `Weights` entry (rows already named), a `Meta` (date, single) and a builder value per side, its window-rule fields and rows, and `TestSubjectSectionMatchesSeed` learns that the derived key is not a seeded binding; it reuses `valuecompare.EndsDate`. S11-13 calls `Compare` from Promote with metas from the catalog and deletes `match`'s profile, periods and `TestDefaultProfileCardinality`. S11-14's `segment` section keys by `vocabkey.SubjectType` and `PropertyKey`, and its bridge kinds pass the seed check as they are. S11-19 reads `Properties` for evidence lines. A new property weight or pair is a registry entry and an inventory row.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/datevalue/datevaluetest` | `Part`, `Between` *(new)* (D-19) |
| `core/valuecompare` | tests move onto `datevaluetest` (D-19); `datewindow.go` `func EndsDate(first, last DateEnd) (datevalue.Value, bool)` *(new)* (D-8) |
| `core/vocabkey` | `type SubjectType string`, `type PropertyKey string` *(new)* (D-2) |
| `core/propertycompare` | `func (v Value) Positive() bool` *(new)*; `Compare` calls it (D-18) |
| `core/subjectcompare` *(new)* | `subjectcompare.go`: types and `Compare` (D-1, D-3 – D-5, D-9); `pairs.go`: `side`, `period` (D-6 – D-8) |
| `core/propertycompare/propertycomparetest` *(new)* | The builders of D-11 |
| `core/subjectcompare/subjectcomparetest` *(new)* | `Metas()` (D-11) |
| `core/matchregistry` | `registry.go` `Registry.Subject` *(new)*; `subject.go` *(new)* `defaultSubject`; `inventory_test.go` `notYetBuilt`; `guard_test.go` `layerPackages`; `registry_test.go` `TestDefaultIsFresh`; `subject_test.go` *(new)* |
| `core/database/subjectvocab` | `subject_section_test.go` *(new)* `TestSubjectSectionMatchesSeed` |

```go
// Package subjectcompare is layer 3 of matching: how alike two Subjects of one
// type are over the properties both record. Pure: no I/O, no catalog, no
// numbers; the subject section and layer 1's rules come in as arguments.
package subjectcompare

// Weights are one lead property's part in a Subject comparison, in points.
type Weights struct {
	Agree    float64 // earned on agree; Agree × similarity on partial
	Conflict float64 // lost on conflict; also its weight in Similarity
}

// Pool lists properties whose values join the lead's on each side.
type Pool struct{ With []vocabkey.PropertyKey }

// Period reads two date properties as one range that stands for the lead on
// a side with no positive lead value.
type Period struct{ Start, End vocabkey.PropertyKey }

// Rules is the match registry's subject section.
type Rules struct {
	// Weights lists, per Subject type, the lead properties compared and what
	// each counts.
	Weights map[vocabkey.SubjectType]map[vocabkey.PropertyKey]Weights
	Pools   map[vocabkey.SubjectType]map[vocabkey.PropertyKey]Pool
	Periods map[vocabkey.SubjectType]map[vocabkey.PropertyKey]Period
}

// Values are one Subject's values by property. Product properties only.
type Values map[vocabkey.PropertyKey][]propertycompare.Value

// Metas are the catalog's value type and cardinality for each property.
type Metas map[vocabkey.PropertyKey]propertymeta.Meta

type Result struct {
	Similarity float64          // 0 to 1: weighted mean over counted properties; 0 when none
	Evidence   float64          // points: agree weights minus conflict penalties
	Properties []PropertyResult // one per lead property, ascending key
}

type PropertyResult struct {
	Key    vocabkey.PropertyKey // the lead property
	Result propertycompare.Result
	Points float64 // what it added to Evidence
}

func Compare(kind vocabkey.SubjectType, a, b Values, metas Metas, rules Rules, value valuecompare.Rules) Result
```

`Compare`, per lead key of `rules.Weights[kind]` in ascending order: build each side with `side(kind, key, values, rules)` (the lead's values, then its pool's, D-6, or its period when it has no positive lead value, D-7, D-8); call `propertycompare.Compare(sideA, sideB, metas[key], value)`; add points and the similarity terms (D-5). `side` copies into a new slice; it never appends to the caller's. Errors: none — `Compare` cannot fail.

```go
// core/matchregistry/subject.go
func defaultSubject() subjectcompare.Rules {
	return subjectcompare.Rules{
		Weights: map[vocabkey.SubjectType]map[vocabkey.PropertyKey]subjectcompare.Weights{
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
		Pools: map[vocabkey.SubjectType]map[vocabkey.PropertyKey]subjectcompare.Pool{
			vocabkey.Person: {vocabkey.PropertySexAtBirth: {With: []vocabkey.PropertyKey{vocabkey.PropertyGender}}},
		},
		Periods: map[vocabkey.SubjectType]map[vocabkey.PropertyKey]subjectcompare.Period{
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

**STEP-1 — Layer 1 date fixtures from `datevaluetest`**
Files: `core/datevalue/datevaluetest/datevaluetest.go`, `core/valuecompare/date_test.go`, `core/valuecompare/contract_test.go`, `core/valuecompare/datewindow_test.go`.
Do: first run `go test ./core/valuecompare/... -v | grep -c -- '--- PASS'` and write the count down. Then D-19: add `Part` and `Between`, delete the six private builders, and rewrite every call site as D-19 says.
Check: `go test ./core/valuecompare/... ./core/datevalue/... ./core/matchregistry/...` passes with no tags; the PASS count is unchanged; `rg -n 'func (optInt|ymd|about|phrase|between|dp)\(' core/valuecompare` prints nothing.

**STEP-2 — The inverse of `DateEnds`**
Files: `core/valuecompare/datewindow.go`, `core/valuecompare/datewindow_test.go`. Skill: `add-match-rule`.
Do: `EndsDate` per D-8, doc comment "EndsDate is the date whose DateEnds are first and last: a range when both are closed, BEF last when first is open, AFT first when last is open. ok is false when both are open or first begins after last ends." Add `TestEndsDate` *(new)* (T-8): the AC-4.4 round trips and the two not-ok cases, building dates with `datevaluetest` and `bound`.
Check: `go test ./core/valuecompare/...` passes with no tags; `go test ./core/matchregistry/ -run TestLayerPackages` passes.

**STEP-3 — Typed keys, the package skeleton and the guard**
Files: `core/vocabkey/vocabkey.go`, `core/subjectcompare/subjectcompare.go` *(new)*, `core/matchregistry/guard_test.go`, `.cursor/rules/match-rules.mdc`. Skill: `add-match-rule` ("Add a layer package").
Do: the two key types (D-2). The package doc, every type of §6, and `Compare` with D-4 only (one `PropertyResult` per lead key with layer 2's result on the lead's own values, points 0; similarity and evidence 0). The `core/subjectcompare` `layerPackages` row and the `match-rules.mdc` changes (D-16).
Check: `go build ./core/...` succeeds; `go test ./core/matchregistry/ -run TestLayerPackages -v` passes and lists `core/subjectcompare/pure` and `core/subjectcompare/numbers`.

**STEP-4 — Shared fixtures**
Files: `core/propertycompare/propertycomparetest/propertycomparetest.go` *(new)*, `core/propertycompare/propertycompare.go`, `core/propertycompare/helpers_test.go` (deleted), `core/propertycompare/fixtures_test.go` *(new)*, `core/propertycompare/propertycompare_test.go`, `core/propertycompare/invariants_test.go`, `core/subjectcompare/subjectcomparetest/subjectcomparetest.go` *(new)*, `core/matchregistry/guard_test.go`.
Do: first run `go test ./core/propertycompare/... -v | grep -c -- '--- PASS'` and write the count down. D-11: move S11-09's ten helpers into `propertycomparetest` under the exported names, delete `helpers_test.go`, and change every call in `core/propertycompare`'s tests to the exported names (no test case changes). Add `subjectcomparetest.Metas`. Add both `layerPackages` rows (D-16). Add `Value.Positive` (D-18) and make `propertycompare.Compare`'s positive checks call it.
Check: `go test ./core/propertycompare/... ./core/matchregistry/...` passes with no tags, and `go test ./core/propertycompare/... -v | grep -c -- '--- PASS'` prints the same count as before the move.

**STEP-5 — The registry's subject section and its checks**
Files: `core/matchregistry/registry.go`, `core/matchregistry/subject.go` *(new)*, `core/matchregistry/registry_test.go`, `core/matchregistry/subject_test.go` *(new)*, `core/matchregistry/inventory_test.go`, `core/database/subjectvocab/subject_section_test.go` *(new)*, `docs/matching-model.md`. Skills: `add-match-rule` ("Add a knob"), `add-seeded-vocabulary` (read only).
Do:
- `Registry.Subject` with the doc comment "Subject is layer 3: what each property counts when two Subjects of one type are compared." and `Default()` filling it from `defaultSubject()` (D-12, §6 sketch), each number with its source comment.
- `TestDefaultIsFresh`: on `first`, set `first.Subject.Weights[vocabkey.Person][vocabkey.PropertyName] = subjectcompare.Weights{}`, set element `first.Subject.Pools[vocabkey.Person][vocabkey.PropertySexAtBirth].With[0] = "changed"`, and set `first.Subject.Periods[vocabkey.Event][vocabkey.PropertyDate] = subjectcompare.Period{}`, before the existing comparison of a new `Default()` with `want`.
- `TestSubjectSection` (AC-4.3, T-6) and `TestSubjectSectionMatchesSeed` (AC-4.1, AC-4.2, T-7), per D-14. `vocabkey.PrimaryKinds()` and the seed hold `string`s: convert with `vocabkey.SubjectType(k)` / `vocabkey.PropertyKey(k)` where a map is keyed by the typed keys.
- `notYetBuilt`: replace `"subject."` with `"subject.Weights.person.lifespan."` and `"subject.person.lifespan."` (D-13).
- `docs/matching-model.md` §Knob inventory rows: `subject.person.name.Agree` → `subject.Weights.person.name.Agree`; `subject.person.name.Conflict` → `subject.Weights.person.name.Conflict`; `subject.person.sex_at_birth.Agree` → `subject.Weights.person.sex_at_birth.Agree`; `subject.person.sex_at_birth.Conflict` → `subject.Weights.person.sex_at_birth.Conflict`; `subject.event.event_type.Agree` → `subject.Weights.event.event_type.Agree`; `subject.event.event_type.Conflict` → `subject.Weights.event.event_type.Conflict`; `subject.event.date.Agree` → `subject.Weights.event.date.Agree`; `subject.event.date.Conflict` → `subject.Weights.event.date.Conflict`; `subject.place.toponym.Agree` → `subject.Weights.place.toponym.Agree`; `subject.place.toponym.Conflict` → `subject.Weights.place.toponym.Conflict`; `subject.person.lifespan.Agree` → `subject.Weights.person.lifespan.Agree`; `subject.person.lifespan.Conflict` → `subject.Weights.person.lifespan.Conflict`. Start, Source and Changes cells unchanged. The row `subject.Pairs.event.date` is replaced by two rows: Key `subject.Periods.event.date.Start`, Start `start_date`, Source `match.DatePeriods`, Changes "Start of the period that stands for `date` on a side with no date"; Key `subject.Periods.event.date.End`, Start `end_date`, Source `match.DatePeriods`, Changes "End of that period". The row `subject.Pairs.person.sex_at_birth` becomes Key `subject.Pools.person.sex_at_birth.With`, Start `gender`, Source unchanged (its "decided" link), Changes "Properties whose values join `sex_at_birth`'s on each side". The five other `subject.person.lifespan.` rows are unchanged.
- §Knob inventory intro: after the sentence ending "a pair compared together is keyed by its lead property (`sex_at_birth` covers `gender`).", insert the sentence "The subject section's weights are `subject.Weights.<kind>.<property>`, a derived property's (`lifespan`) included; a pool is keyed by its lead in `subject.Pools`, a period by its lead in `subject.Periods`."
- §Worked example step 3: `subject.person.name.Agree` → `subject.Weights.person.name.Agree`, `subject.person.lifespan.Agree` → `subject.Weights.person.lifespan.Agree`; step 4: `subject.event.event_type.Agree` → `subject.Weights.event.event_type.Agree`, `subject.event.date.Agree` → `subject.Weights.event.date.Agree`.
- §Removed knobs: in the `graphalign` `PropertyAgreement` row `subject.place.toponym.Agree` → `subject.Weights.place.toponym.Agree`; in the `DefaultConfig.ConflictPenalty` row `subject.event.date.Conflict` → `subject.Weights.event.date.Conflict` and `subject.place.toponym.Conflict` → `subject.Weights.place.toponym.Conflict`; in the `DefaultProfile` `start_date` / `end_date` row "the `date` pair in `subject.Pairs.event.date`" → "the `date` period in `subject.Periods.event.date`".
Check: `go test ./core/matchregistry/...` passes (`TestKnobInventoryMatchesRegistry`, `TestDefaultIsFresh`, `TestSubjectSection`, `TestLayerPackages`); `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` passes; `rg -n 'subject\.(person|event|place|Pairs)\.' docs/matching-model.md` prints only the five lifespan window rows.

**STEP-6 — Points and similarity**
Files: `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/subjectcompare_test.go` *(new)*.
Do: D-5 in `Compare`. `TestCompare` (D-11) with the US-1 rows except AC-1.2, which needs pools (STEP-7) (T-1).
Check: `go test ./core/subjectcompare/...` passes with no tags.

**STEP-7 — Pools and periods**
Files: `core/subjectcompare/pairs.go` *(new)*, `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/subjectcompare_test.go` *(new)*, created in STEP-6.
Do: `side` and `period` (D-6 – D-8) and their use in `Compare`. Add AC-1.2 and the US-2 rows to `TestCompare` (T-1, T-2).
Check: `go test ./core/subjectcompare/...` passes; `go vet ./core/subjectcompare/...` is clean.

**STEP-8 — Invariants**
Files: `core/subjectcompare/invariants_test.go` *(new)*.
Do: `TestInvariants` per D-15 (T-3).
Check: `go test ./core/subjectcompare/... -run TestInvariants -v` passes; then `go test ./core/subjectcompare/... ./core/propertycompare/... ./core/valuecompare/... ./core/matchregistry/... ./core/vocabkey/...` passes with no tags.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching-model.md`, `.agents/skills/add-match-rule/SKILL.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do:
- `comparison-layers.md` §Layer 3, the options table under "Decided (2026-10-16): two axes": the row "| Name agrees, sex conflicts | 0.81 | 1.5 |" becomes "| Name agrees, sex conflicts | 0.71 | 1.5 |". After the **Registry shape** paragraph, a paragraph *(new)*: "**Built in S11-10.** `subjectcompare.Compare(kind, a, b, metas, rules, value)` runs layer 2 once per lead property in `subject.Weights[kind]`, in key order. Agree earns the agree weight, partial the agree weight × the speaking pair's similarity, conflict loses the conflict penalty, unknown counts for nothing. Similarity is the weighted mean over counted properties: agree 1 and partial its similarity, each at the agree weight; conflict 0 at the penalty. `subject.Pools` adds properties' values to the lead's on each side (sex with gender). `subject.Periods` builds, on a side with no positive `date`, one range from its earliest start and latest end, or `AFT` the start or `BEF` the end when it has only one (`valuecompare.EndsDate`). Value types and cardinalities come from the catalog as `Metas`. A derived property (lifespan) is a lead in `subject.Weights` like any other."
- `matching-model.md` §The stack, table row 3: the Registry section cell `` `subject` `` becomes "`subject`, and `value` passed through to layer 2". §Glossary, the **pair (compared together)** row becomes "| **pair (compared together)** | Two or more properties compared as one, keyed by the lead: a pool adds properties' values to the lead's on each side (sex at birth with gender); a period reads `start_date` and `end_date` as one range for `date` on a side with no date |".
- `add-match-rule` SKILL.md (as S11-09 leaves it):
  - Front matter `description`: "(core/valuecompare, core/propertycompare, core/matchregistry)" becomes "(core/valuecompare, core/propertycompare, core/subjectcompare, core/matchregistry)", and in its "Use when" sentence, "a Subject property weight, pool or period," is inserted after "a Relation or Cause code," (the YAML block wraps that phrase across two lines; rewrap the paragraph).
  - "Rules that hold everywhere", a bullet *(new)* after S11-09's layer 2 bullet: "**Layer 3 (`core/subjectcompare`) turns outcomes into points.** It reads only the lead properties in `subject.Weights` for the kind; agree earns the agree weight, partial agree weight × similarity, conflict loses the penalty and weighs it in similarity. Registry maps are keyed by `vocabkey.SubjectType` and `vocabkey.PropertyKey`, written as `vocabkey` constants, never as string literals; no other package declares a key type. Value type and cardinality come from the catalog as `Metas`, never from the registry. Tests build values with `propertycomparetest` and metas with `subjectcomparetest.Metas`."
  - "Pure helpers are shared, not copied." bullet: append "`valuecompare.EndsDate` is the one way back from date ends to a date."
  - A section *(new)* after "Add a knob", titled "Add a Subject property or pair": "1. Add the key to `core/vocabkey` if it has none, and to the key list in `TestVocabKeysMatchSeed`. 2. Add its `Weights` (or a `Pools` / `Periods` entry under an existing weighted lead) in `core/matchregistry/subject.go`, with a source comment. 3. Add the inventory rows: `subject.Weights.<kind>.<property>.Agree` and `.Conflict`, `subject.Pools.<kind>.<lead>.With`, or `subject.Periods.<kind>.<lead>.Start` and `.End`. 4. If the property is new to `subjectcomparetest.Metas`, add it there. `TestSubjectSection` and `TestSubjectSectionMatchesSeed` (`core/database/subjectvocab`) fail on a key the seed does not bind to the kind, a pair whose lead has no weights, a pooled property of another value type, and a fixture meta that differs from the seed. 5. Add a `TestCompare` row for it in `core/subjectcompare`."
  - "Check": the command becomes `go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/matchregistry/...`, followed by the line "and `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` after a subject section change."
- `completed.md`, newest first: "## S11-10 — Subject compare" / "**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N). Layer 3 of matching exists: `subjectcompare.Compare` says how alike two Subjects of one type are as a similarity (0 to 1) and evidence (points), from layer 2's outcome on each property the registry's new `subject` section lists. Sex at birth and gender are pooled, and an event with no date compares its start and end as one period. The section keys by the new `vocabkey.SubjectType` and `PropertyKey`, and a seed test ties it to the product vocabulary. `valuecompare.EndsDate` turns date ends back into a date, and `propertycomparetest` and `subjectcomparetest` hold the layers' test builders. No caller yet; nothing changes for researchers." with the date and PR number filled in.
- Map row status `Landed #N`; this plan's Status `Landed in [#N](…) (YYYY-MM-DD)` and a **Landed** note under the header listing each difference from the plan (or "None.").
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-10-subject-compare.md` reports no errors; `go test ./core/matchregistry/ -run TestKnobInventoryMatchesRegistry` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Add layer 3, Subject compare, with the registry's subject section."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/subjectcompare/subjectcompare_test.go` `TestCompare` *(new)* | Rows AC-1.1 – AC-1.15, inputs and expectations as in §2 | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, AC-1.12, AC-1.13, AC-1.14, AC-1.15, D-3, D-4, D-5, D-9, D-12 |
| T-2 | Go | same, `TestCompare` | Rows AC-2.1 – AC-2.13; AC-2.11 runs both start orders | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, AC-2.10, AC-2.11, AC-2.12, AC-2.13, D-6, D-7, D-8 |
| T-3 | Go | `core/subjectcompare/invariants_test.go` `TestInvariants` *(new)* | 500 generated cases per invariant per kind (D-15) → AC-3.1 – AC-3.6 hold | AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.6, D-15 |
| T-4 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` | New rows → `core/subjectcompare` pure and free of numbers; both fixture packages pure | D-10, D-16 |
| T-5 | Go | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry`; `core/matchregistry/registry_test.go` `TestDefaultIsFresh` | Every `subject.Weights`, `subject.Pools`, `subject.Periods` field has a row; lifespan rows covered by `notYetBuilt`; a changed copy leaves the next `Default()` unchanged | D-12, D-13 |
| T-6 | Go | `core/matchregistry/subject_test.go` `TestSubjectSection` *(new)* | `Default().Subject` → AC-4.3 | AC-4.3, D-14 |
| T-7 | Go | `core/database/subjectvocab/subject_section_test.go` `TestSubjectSectionMatchesSeed` *(new)* | `Default().Subject` and `subjectcomparetest.Metas()` against `seedTypes`, `allBindings()`, `seedProperties` → AC-4.1, AC-4.2 | AC-4.1, AC-4.2, D-11, D-14 |
| T-8 | Go | `core/valuecompare/datewindow_test.go` `TestEndsDate` *(new)* | AC-4.4 round trips; both open → not ok; first after last → not ok | AC-4.4, D-8 |
| T-9 | Go | `core/valuecompare` `TestDateEnds`, `TestDateWindow`, the date tables, `TestComparerContract` | Unchanged cases on `datevaluetest` builders → same results, same PASS count | D-19 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestKnobInventoryMatchesRegistry` (a row key that does not match `registryKeys`), `TestLayerPackages` (a literal in `pairs.go` or a stray import), S11-09's `core/propertycompare` tests after the helpers move (STEP-4), `TestDefaultProfileCardinality` (must not change: `core/match` is untouched).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] go test ./core/subjectcompare/... ./core/propertycompare/... ./core/valuecompare/... ./core/datevalue/... ./core/matchregistry/... ./core/vocabkey/... passes with no build tags
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without macOS every check but xcode-test passes, and CI runs the Xcode tests on the PR (this PR changes no Swift)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored; core/match and core/graphalign are unchanged
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none (no copy)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: comparison-layers.md §Layer 3, matching-model.md (stack table, glossary, inventory keys, worked example, removed knobs), add-match-rule skill, match-rules.mdc (STEP-3, STEP-5, STEP-9)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — nothing reaches it until S11-13. The reviewer runs `go test ./core/subjectcompare/... -v` with no build tags and reads the `TestCompare` subtests `AC-1.1 …`, `AC-1.2 …` and `AC-1.7 …`: these are the map's Check ("name 1.0 / 2.5; name + gender 1.0 / 3.1; birth vs death with matching dates ≈ 0.09"). Then `go test ./core/matchregistry/... -v` and `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/ -run TestSubjectSectionMatchesSeed -v` to see the section, inventory, guard and seed checks pass.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S11-09 lands with names or behavior other than its plan's | Implementer step 2 and the first stop condition: read S11-09's **Landed** note before STEP-3 |
| A literal fails the numbers guard | D-10; a doubled value is `w + w` (skill rule); `pairs.go` compares `DateEnd` fields and calls `EndsDate` |
| Float sums differ by order | D-4: ascending key order; tests compare within 0.0001 and AC-3.1 within 1e-9 |
| An AC similarity drifts because layer 1 changed after the planning base | §2 values were read from `valuecompare.Compare` at `95d1f34`; a drift is a stop condition, not a test edit |

**Stop and report if:**

- `main` differs from §4's S11-09 rows: `propertycompare.Compare`'s signature, `Value`'s `Negative` / `Neutral`, `Result.Speaking` (an `Unknown` pair still speaks, with `Speaking.A` / `B` the input values), `Speaking.Result.Similarity`, the `Outcome*` or `Cause*` constants, `propertymeta.Meta`, `datevaluetest.YMD` / `About`, or the ten helpers in `core/propertycompare/helpers_test.go`. Report each difference.
- Layer 2 with the zero `Meta` does not give `Speaking.Result.Relation` `Unknown` and `Speaking.Detail.Cause` `unsupported` (AC-1.14).
- `valuecompare.Compare` at your base gives a relation or similarity different from §2 for an AC pair (for example `1817` / range 1815 – 1820 is not `Contains` 0.55, or `J. Robins` / `James Robins` is not `Contains` 0.8): report the pair and the output; do not change the AC.
- `core/subjectcompare` cannot pass `TestLayerPackages` with only the imports D-10 lists.
- `registryKeys` produces a key for the subject section other than D-13's: report the keys it prints; do not change the guard.
- Moving `core/valuecompare`'s tests onto `datevaluetest` (STEP-1) changes any case, name or result, or the PASS count.
- Moving S11-09's helpers or adding `Positive` (STEP-4) changes any `core/propertycompare` test case or its result.
- An invariant of US-3 fails under D-5 – D-8 as written: report the generated case (seed, index, both sides) and which invariant.
- `core/subjectcompare` already exists, or `vocabkey` already declares `SubjectType` or `PropertyKey`.
- Merging `main` conflicts in a paragraph STEP-5 or STEP-9 rewrites, so keeping both sides' text is not possible.

## 12. Open questions

None.

## 13. PR

**Title:** S11-10 — Subject compare
**Body:** objective in one sentence; what shipped against each US; AC evidence (test names); deviations from the plan; manual check (the `go test` runs of §10); `Plan: docs/deployment-plan/spike-11/prs/S11-10-subject-compare.md`.
