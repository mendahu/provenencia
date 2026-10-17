# Deployment Plan — Spike 11

Promote's matching is rebuilt as a stack of pure, registry-driven layers that the rest of the app borrows from, and the old matching code is deleted as each layer lands. The design is in [`comparison-layers.md`](comparison-layers.md) (layers 1–6, bands, the match registry, the reuse inventory) and [`claim-evidence.md`](claim-evidence.md) (the Promote sheet's pins). It pulls in the "make the algorithm graspable" part of [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md). PRs follow a build-then-migrate rhythm: a PR builds a layer, the next moves every caller onto it and deletes what it replaces.

**A brief covers one view.** A PR that changes chrome waits on a brief for each view it touches. Briefs and PR plans are written just before the PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)). A bugfix or a behavior change with no new chrome is not gated.

## Status

**Open.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** Promote's matching runs on one layered, registry-driven module the rest of the app borrows from, and the old matching code is gone.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project:

1. **Sparse matches work.** A Source person with only a name and a birth date is proposed strong on the person whose name and birth match. A look-alike with the same name and a different birth year is not.
2. **Impossible matches never surface.** A birth is never proposed on a death. A person who died before the Subject was born is never proposed for them.
3. **Fast, and it explains itself.** A Source with 100 Subjects proposes without a noticeable wait. Deciding one row re-proposes only what moved, each with "updated because…".
4. **Conflicts are visible.** A row whose comparisons include a conflict shows *contested*, whatever its band.
5. **The evidence sheet shows all the evidence.** Joining an entity with five named members shows five name matches (identical values collapsed into one line), worded as sentences. A relationship line says what matched ("They have a matching birth event") and checking it pins everything behind it. Pins land on every member.
6. **"The same" means the same everywhere.** José and Jose, Montréal and Montreal, compare the same way in Promote, search and the reconciler, under the reconciler's own grouping policy.
7. **One place for every knob.** Every matching number is in `core/matchregistry` and listed in the knob inventory. A guard test fails on a number anywhere else.
8. **The old code is gone:** `core/match`, `core/database/matching`, `core/graphalign/registry.go`, the frequency stats, and `pairExhibits`.

## Where the notes went

| Idea or note | PR | Brief |
| --- | --- | --- |
| [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md) — "make the algorithm graspable" (one-page model, glossary, knob inventory) | S11-01 | — |
| Same — "are the levers right?" | S11-02 – S11-16 (the layer rewrite) | — |
| Same — known gaps | [`comparison-layers.md`](comparison-layers.md) open questions | — |
| Claim evidence shows only the first member; "Through <name>" wording | S11-19, S11-20 | S11-D2 |
| Reuse inventory ([`comparison-layers.md`](comparison-layers.md#reuse-across-the-app)) | S11-05 – S11-08, S11-13, S11-19 | — |

**Stays in the ideas folder** — the rest of [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md): scenario files and goldens runner, random-graph invariants, the simulated world, the scoreboard, the labeler, fitting, and dogfood export. Whether they join this spike is decided after the re-implementation.

## Working rules

These are the calls for this spike. PR plans treat them as binding.

### Spike-wide

- **High tolerance for churn.** No one uses the app yet. No adapters, shims, wrappers, feature flags, deprecated aliases or kept-for-now paths. Migration PRs move every caller and delete the old code in the same PR. Behavior may change where the design says so; goldens are updated, not preserved. Catalog shapes matching used to need may be reshaped or dropped.
- **Build, then migrate.** A build PR adds a layer and its registry section with its own tests. The next PR moves every caller onto it and deletes what it replaces. Between the two, the old code may call the new layer (for example `pairExhibits` uses value compare after S11-05 and is deleted in S11-19); nothing gets a compatibility layer.
- **Sparse data first.** Core properties and links (person name and sex; event type and date; place name; subject of a birth or death, parent / child, spouse, place part of place) carry most weight; secondary ones only reinforce and never outweigh a core conflict. The obituary golden plus sparse cases (name + birth date, or less) are each layer PR's acceptance tests. A person's birth date lives on the birth event, so the commonest sparse match (same name, same birth date) is half subject compare and half one link; both halves must be solid alone. Invariant: an agreeing secondary property never lowers a score and never turns a core conflict into a match.
- **Registry numbers: carry over, then estimate.** Each layer PR carries over today's value where the knob exists (name agree 2.5 / conflict 3, sex 0.6 / 1, accept bar 2, bands 4 / 3 / 2, link support 1.5 and credit 0.5 for one-cardinality links). New knobs start from an estimate written in the knob inventory with its reasoning. No tuning beyond the goldens and sparse cases; numbers are adjusted as we review.
- **The test suite is deferred.** Only each layer's unit tests and invariants ship in this spike for now ([`comparison-layers.md`](comparison-layers.md) Tests sections).

### Layers and packages

- **Matching is a stack of pure layers** ([rules for every layer](comparison-layers.md#rules-for-every-layer)). Each takes only what it needs; a rule that needs more moves up a layer. Comparison layers make no workflow decisions; Promote, the reconciler and the evidence view compose them.
- **Packages:** `core/valuecompare` (1), `core/propertycompare` (2), `core/subjectcompare` (3), `core/subjectprofile` (the profile builder; layer 4 is `subjectprofile.Build` plus `subjectcompare.Compare`, with no package of its own), `core/segmentcompare` (5), `core/graphalign` (6, rewritten in place in S11-16), `core/matchregistry` (every tunable).
- **Pure means the package builds and tests with no build tags and no `core/database` import.** An import guard test in each layer package enforces it. Value types the layers compare live in pure packages ([purity](comparison-layers.md#purity-requirements)).
- **Value compare returns a relation, a similarity and a detail** ([layer 1](comparison-layers.md#layer-1-value-comparison)). Relations: `Same`, `Equivalent`, `Contains`, `Overlaps`, `Near`, `Different`, `Unknown`. `Contains` scores below 1; `Near` needs a per-type minimum similarity (today's floors: name 0.75, spelling 0.8). `Detail` carries matched parts, cause codes rich enough to word evidence lines, and the callers' opaque value refs. Relations per type: [table](comparison-layers.md#relations). `ABT` is a range of one unit of its most precise part.
- **A normalized `Key` per value type** sits beside `Compare`: same key exactly when `Compare` says `Same` (or `Equivalent`, by level). Shortlisting, the reconciler, search and term dedupe use it.
- **Property compare holds only what a Subject says about itself** ([layer 2](comparison-layers.md#layer-2-property-comparison)): negatives, neutral terms, cardinality, missing values. No support or standing; no points.
- **Subject compare returns similarity (0–1) and evidence (points)** ([layer 3](comparison-layers.md#layer-3-subject-comparison)). Similarity weighs a conflict by its penalty. No knockout flag: a decisive conflict (event type) is a large penalty. Pairs compared together: sex at birth with gender; `date` with `start_date` / `end_date`.
- **Entity compare is against the reconciled conclusion** ([entities](comparison-layers.md#entities-compare-against-the-conclusion)). The profile is strict: concluded values positive, researcher-denied values negative, nothing else. `subjectprofile.Build` also takes an `Excluding(member)` option.
- **Lifespan is for people only** ([lifespan](comparison-layers.md#lifespan-a-derived-property-for-people)): derived from direct neighbors' dates by registry rules; an impossible overlap is a large conflict, agreement weighs almost nothing.
- **Segment compare is one hop and pure** ([segments](comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6)). Only subject-of-birth and subject-of-death are declared one-cardinality; a differing neighbor there subtracts. Neighbor credit travels one hop; chains come from the walk.
- **Graph alignment follows the layer 6 procedure** ([procedure](comparison-layers.md#layer-6-procedure)): shortlist → score, keep top K with ties → one queue → segments around accepted pairs → settle dirty pairs. The answer doesn't depend on where the walk starts.
- **Bands come from evidence** ([bands](comparison-layers.md#bands)): similarity only gates candidates and gives the first sort. The accept bar is its own setting. A conflict shows a contested marker; it never caps the band.
- **Frequency weighting is gone**, for names, event types and toponyms ([frequency](comparison-layers.md#frequency)).
- **One match registry** ([match registry](comparison-layers.md#match-registry)): `core/matchregistry`, one file per section (value, subject, segment, alignment, evidence), typed keys, each layer gets its section as an argument. Catalog data (aliases, name patterns) is passed in; data-model facts (cardinality, term flags, bridge kinds) stay in the catalog. A guard test fails on numeric literals other than 0 and 1 in layer packages; another fails when the knob inventory drifts from the registry.
- **`core/database/matching` is deleted, not migrated.** Nothing calls `ForSubject` or `ForEntity` outside its tests. Merge hints are rebuilt on subject compare over two profiles when their UI exists.
- **The reconciler keeps its own decisions** and builds them on value compare. Its grouping policy is decided in S11-06's plan.

### Claim evidence

- **A separate, lazy query per row** ([claim evidence](claim-evidence.md#separate-from-matching)): inputs are the page's targets, not matching internals.
- **Lines** ([lines](claim-evidence.md#lines)): one per member per property, grouped by property, identical values collapsed; `Same` / `Equivalent` start checked, `Contains` / `Overlaps` / `Near` pinnable and unchecked, disagreements shown and not pinnable. No "select all".
- **Relationships** ([relationships](claim-evidence.md#evidence-through-relationships)): a corresponding segment is its own checkable line that pins the bundled plumbing (bridge ends, everything on the bridge, the neighbor's defining properties, e.g. `event_type`); the neighbor's other properties are checkable lines under it.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S11-01 | One-page matching model | 1 | Diagram, glossary, knob inventory with starting values and estimates | — | — | — | Scoped |
| S11-02 | Value compare core and match registry | 2 | `valuecompare` contract, `Key`, text / term / integer, `matchregistry` skeleton and value section, guard tests | — | S11-01 | — | Scoped |
| S11-03 | Value compare: names | 2 | Name comparer on relations; word helpers move in from `autoreconcile` | — | S11-02 | — | Scoped |
| S11-04 | Value compare: dates | 2 | Date comparer on relations; date windows move in; `ABT` as a range | — | S11-02 | — | Scoped |
| S11-05 | Migrate matching onto value compare | 2 | `match`, `graphalign`, `promotealign` call `valuecompare`; `match`'s comparers deleted | — | S11-03, S11-04 | — | Scoped |
| S11-06 | Migrate the reconciler onto value compare | 2 | `Compatible`, keys, grouping policy, place-chain windows, `hasValue` copy | — | S11-05 | — | Scoped |
| S11-07 | One normalization key across the core | 2 | Search rescoring, search-index dedupe, term dedupe, alias tokens use `Key` | — | S11-05 | — | Scoped |
| S11-08 | One normalization helper in Swift | 2 | Shared Swift normalize for filters; `FakeStore` groups names by parts | — | S11-06 | — | Scoped |
| S11-09 | Property compare | 3 | `propertycompare`: negatives, neutral terms, cardinality, outcome table | — | S11-05 | — | Scoped |
| S11-10 | Subject compare | 3 | `subjectcompare` + registry subject section; pairs; similarity and evidence | — | S11-09 | — | Scoped |
| S11-11 | Subject profile and entity compare | 3 | `subjectprofile.Build`: strict profile, denied as negatives, `Excluding` | — | S11-10 | — | Scoped |
| S11-12 | Lifespan for people | 3 | Profile builder derives a lifespan; subject compare checks it | — | S11-11 | — | Scoped |
| S11-13 | Migrate Promote scoring onto subject compare | 3 | Property scoring, seeding, duplicate rows; delete `core/match`, `core/database/matching`, frequency | — | S11-12 | — | Scoped |
| S11-14 | Segment compare | 4 | `segmentcompare`, typed bridge keys, birth / death cardinality, one-hop credit | — | S11-13 | — | Scoped |
| S11-15 | Shortlist index | 4 | Key → handles per kind on `graphcache`; registry shortlist keys | — | S11-13 | — | Scoped |
| S11-16 | Rewrite graph alignment | 4 | Layer 6 procedure, bands from evidence, accept bar; old walk and registry deleted | — | S11-14, S11-15 | — | Scoped |
| S11-17 | Contested flag through FFI | 4 | Proposal rows carry contested and move reasons to Swift | — | S11-16 | — | Scoped |
| S11-18 | Contested marker on Promote rows | 4 | The marker on the page | S11-D1 | S11-17, S11-D1 | — | Scoped |
| S11-19 | Claim evidence query | 5 | Lazy per-row RPC; member lines, collapse, relationship bundles, pin policy; `attachExhibits` removed | — | S11-16 | — | Scoped |
| S11-20 | Claim evidence sheet | 5 | The sheet's new layout and copy | S11-D2 | S11-19, S11-D2 | — | Scoped |
| S11-21 | Large-Source performance check | 6 | 100-Subject benchmark; memoize only if it fails | — | S11-16 | — | Scoped |
| S11-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S11-D1 | Promote page — row | **S11-18** | Contested marker beside the band |
| S11-D2 | Promote page — evidence sheet | **S11-20** | Whole sheet: "why suggested" summary, property sections, relationship lines, collapsed members, copy |

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify when the slice lands.

```text
SLICE 1 — Graspable model
  S11-01  One-page matching model
  Check: one doc explains every layer, term and knob; each knob has a starting value.

SLICE 2 — Value compare, everywhere
  S11-02  Value compare core and match registry
  S11-03  Value compare: names          (parallel with S11-04)
  S11-04  Value compare: dates
  S11-05  Migrate matching onto value compare
  S11-06  Migrate the reconciler        (parallel with S11-07, S11-09)
  S11-07  One normalization key across the core
  S11-08  One normalization helper in Swift
  Check: Promote proposes as before or better; Montréal / Montreal and José / Jose
         compare the same way in Promote, search and the reconciler; a place chain
         with same-year dates folds by day; core/match has no comparers left.

SLICE 3 — Property, Subject and entity compare
  S11-09  Property compare
  S11-10  Subject compare
  S11-11  Subject profile and entity compare
  S11-12  Lifespan for people
  S11-13  Migrate Promote scoring onto subject compare
  Check: a birth is never proposed on a death; a person who died before the
         Subject's birth is not proposed; core/match, core/database/matching and
         the frequency stats are gone.

SLICE 4 — Segments and graph alignment
  S11-14  Segment compare               (parallel with S11-15)
  S11-15  Shortlist index
  S11-16  Rewrite graph alignment
  S11-17  Contested flag through FFI
  ✎ S11-D1 ──▶ S11-18  Contested marker on Promote rows
  Check: name + birth date proposes strong; the same proposal whether or not a
         row starts decided; a decision moves only its neighborhood, with reasons;
         a conflicting row shows contested; graphalign/registry.go is gone.

SLICE 5 — Claim evidence
  S11-19  Claim evidence query
  ✎ S11-D2 ──▶ S11-20  Claim evidence sheet
  Check: five named members show as matches with identical values collapsed;
         "They have a matching birth event" pins its bundle; pins land on every member.

SLICE 6 — Performance
  S11-21  Large-Source performance check
  Check: a 100-Subject Source proposes within the benchmark; memoization exists
         only if the benchmark needed it.

S11-99 — Dogfood close / docs
```

Parallel: S11-03 with S11-04; S11-06, S11-07 and S11-09 after S11-05; S11-14 with S11-15; slice 5 with slice 6 after S11-16. S11-D1 and S11-D2 can be drawn while slice 4's engine PRs run.

### Dependencies at a glance

```text
S11-01 ──▶ S11-02 ──▶ S11-03 ──┐
                 └──▶ S11-04 ──┴──▶ S11-05 ──▶ S11-06 ──▶ S11-08
                                       ├──▶ S11-07
                                       └──▶ S11-09 ──▶ S11-10 ──▶ S11-11 ──▶ S11-12 ──▶ S11-13
S11-13 ──▶ S11-14 ──┐
S11-13 ──▶ S11-15 ──┴──▶ S11-16 ──▶ S11-17 ──▶ S11-18
                            ├──▶ S11-19 ──▶ S11-20
                            └──▶ S11-21
S11-D1 ──▶ S11-18
S11-D2 ──▶ S11-20
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Link, do not restate.

### Slice 1 — Graspable model

#### S11-01 — One-page matching model

**Ships.** A domain doc (`docs/matching.md` rewritten, or a new `docs/matching-model.md` that replaces it) with one diagram of the six layers, the glossary (relation, similarity, evidence, segment, shortlist, candidate, band, contested…), and the knob inventory: every registry field, its section, starting value or estimate with reasoning, and what it changes.
**Not this PR.** Any code. `core/matchregistry` (S11-02).
**Check.** A reader can trace a Promote row from value compare to band on one page; every knob in the comparison-layers doc appears in the inventory with a value.
**Context.** Rule "Registry numbers: carry over, then estimate"; [`comparison-layers.md`](comparison-layers.md) whole; [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md) "First: make the algorithm graspable"; [`promote-graph-alignment.md`](../../promote-graph-alignment.md) (superseded sections are marked).

### Slice 2 — Value compare, everywhere

#### S11-02 — Value compare core and match registry

**Ships.** `core/valuecompare`: pure value types, `Result` (relation + similarity), `Detail` (parts, cause codes, value refs), `Key`, and the text, term and integer comparers. `core/matchregistry` with `Default()` and the value section. Import guard, contract tests, the no-magic-numbers guard, and the inventory-drift guard.
**Not this PR.** Names (S11-03), dates (S11-04), any caller (S11-05).
**Check.** `go test ./core/valuecompare/... ./core/matchregistry/...` passes with no build tags; the contract tests run against every comparer.
**Context.** [Layer 1](comparison-layers.md#layer-1-value-comparison), [relations](comparison-layers.md#relations), [detail](comparison-layers.md#detail), [purity](comparison-layers.md#purity-requirements), [match registry](comparison-layers.md#match-registry).

#### S11-03 — Value compare: names

**Ships.** The name comparer on relations and parts detail: folds, particles, aliases (tokens normalized), initials, spelling (`Near` minimum 0.8 for words, 0.75 for names), generation conflict, swapped part types → `Near`. `SpellingSimilarity`, `EditDistance`, `NormalizeForm` and `IsInitial` move in from `core/autoreconcile`. No frequency.
**Not this PR.** Callers (S11-05). The reconciler's grouping (S11-06).
**Check.** Today's `names_test.go` ladders and generated invariants pass on the new relations; the worked name cases in the relations table hold.
**Context.** [Relations](comparison-layers.md#relations) worked cases; [`matching.md`](../../matching.md) Names.

#### S11-04 — Value compare: dates

**Ships.** The date comparer on range relations; `DateEnds`, windows, `Overlaps`, `Contains`, `Intersect`, `HoldsAt` move in from `core/autoreconcile`; `ABT` read as one unit of its most precise part; tolerance in the registry.
**Not this PR.** Period pooling across `date` / `start_date` / `end_date` (S11-10). Callers (S11-05, S11-06).
**Check.** `ABT 1817` / `1818` is `Contains`; `BEF 1820` / `BEF 1830` is `Contains`; today's `dates_test.go` cases pass, updated where the ABT rule changes them.
**Context.** [Relations](comparison-layers.md#relations) and the `ABT` rule; [`matching.md`](../../matching.md) `DateComparer` row.

#### S11-05 — Migrate matching onto value compare

**Ships.** `core/match` (`Score`, `Evaluate`, `pairValues`, `poolPeriods`), `core/graphalign` and `core/database/promotealign` (`pairExhibits`, stats) call `valuecompare`; `nameformats.Loaded.Comparer` builds its rules. `match`'s comparers, `DiscountComparer` and the name frequency discount are deleted.
**Not this PR.** Layer 2 and up (slice 3). The reconciler (S11-06).
**Check.** The obituary golden and sparse cases pass (expected values updated where relations change them); `core/match` has no comparer types left.
**Context.** Rules "Build, then migrate", "High tolerance for churn"; [what moves where](comparison-layers.md#what-moves-where-in-todays-code).

#### S11-06 — Migrate the reconciler onto value compare

**Ships.** `autoreconcile` keys, `Compatible`, `subsumes` / `partFits` and the pipeline's grouping use `valuecompare` relations under a grouping policy the plan decides (recommendation on record: group on `Same` and `Contains`, not `Equivalent`). `conclusionheaders` place chains use day-precision windows. `autoreconciler.hasValue` calls `carries`.
**Not this PR.** Pins (S11-19).
**Check.** `Sainte-Anne` / `Sainte Anne` group; `J. Robins` folds into `James Robins`; a place chain with two same-year dates folds by day; `Montréal` / `Montreal` follow the chosen policy.
**Context.** [Reuse inventory](comparison-layers.md#reuse-across-the-app); [decided 2026-10-17](comparison-layers.md#decided-on-2026-10-17) (reconciler grouping); rule "The reconciler keeps its own decisions"; [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md).

#### S11-07 — One normalization key across the core

**Ships.** `core/search` Go rescoring, `searchindex` alternate-value dedupe, `propertyterms.Create` duplicate check and `namealiases` tokens use `valuecompare.Key`.
**Not this PR.** Search's fuzzy ranker stays ([reuse inventory](comparison-layers.md#reuse-across-the-app)). Swift (S11-08).
**Check.** Searching `Montreal` ranks `Montréal` like an exact hit; adding "god-father" beside "godfather" is caught; a capitalized alias token hits.
**Context.** [Reuse inventory](comparison-layers.md#reuse-across-the-app) text-equality table.

#### S11-08 — One normalization helper in Swift

**Ships.** One Swift normalize helper (case and diacritic fold, matching `Key`'s text rules) used by `PVComboBox`, the list filters (`PropertiesModel`, `SourcesModel`, `GraphCanvasItemLayer`, `PVFloatingMenuSelection`); `FakeStore` groups names by parts, not by lowercase form.
**Not this PR.** Go changes.
**Check.** Filtering `jose` finds `José` in every list; `FakeStore`-backed tests group names like Go does.
**Context.** [Reuse inventory](comparison-layers.md#reuse-across-the-app) "Lighter touch"; [`add-swift-test`](../../../.agents/skills/add-swift-test/SKILL.md).

### Slice 3 — Property, Subject and entity compare

#### S11-09 — Property compare

**Ships.** `core/propertycompare`: one property between two Subjects from their value sets; negatives, neutral terms, cardinality, missing values; outcome from relation; output with the speaking pair and excluded values.
**Not this PR.** Callers (S11-13). Support or standing (gone with the strict profile).
**Check.** The invariants and the outcome table in the layer 2 section hold in tests.
**Context.** [Layer 2](comparison-layers.md#layer-2-property-comparison).

#### S11-10 — Subject compare

**Ships.** `core/subjectcompare` and the registry's subject section: per Subject type and property, agree weight and conflict penalty (event type: low agree, very large conflict); the pairs (sex with gender; `date` with `start_date` / `end_date`); similarity and evidence.
**Not this PR.** Profiles (S11-11), lifespan (S11-12), callers (S11-13).
**Check.** The layer 3 examples hold (name 1.0 / 2.5; name + gender 1.0 / 3.1; birth vs death with matching dates ≈ 0.09).
**Context.** [Layer 3](comparison-layers.md#layer-3-subject-comparison); rule "Sparse data first".

#### S11-11 — Subject profile and entity compare

**Ships.** `core/subjectprofile.Build` from reconciler output: concluded values positive (tied leaders and multi-valued properties included), denied values negative, others left out; `Excluding(member)`. Entity compare is `Build` then `subjectcompare.Compare`.
**Not this PR.** Lifespan (S11-12), callers (S11-13).
**Check.** A Subject recording an outvoted value conflicts; one recording a denied value conflicts via negative; leave-one-out drops that member's support.
**Context.** [Entities](comparison-layers.md#entities-compare-against-the-conclusion).

#### S11-12 — Lifespan for people

**Ships.** The profile builder derives a person's lifespan from direct neighbors' dates by registry rules (birth, death, marriage, parent of, residence); subject compare checks it as a people-only pair with a large conflict penalty and near-zero agree weight.
**Not this PR.** Events and places. Two-hop rules.
**Check.** An entity who died in 1840 against a Subject born in 1850 conflicts; born 1850 / died 1990 conflicts; born 1850 / died 1920 is consistent.
**Context.** [Lifespan](comparison-layers.md#lifespan-a-derived-property-for-people).

#### S11-13 — Migrate Promote scoring onto subject compare

**Ships.** `graphalign` scores candidates' properties with entity compare; `promotealign` seeds with it instead of `match.Rank`; duplicate-row checks (`alikeNew`, `pairNode`) use subject compare. Delete `core/match`, `core/database/matching`, `Stats.ValueFreq`, `Stats.Words`, the m / u priors and frequency scales.
**Not this PR.** Segments and the walk (slice 4): the old walk stays until S11-16, calling the new scoring.
**Check.** Golden and sparse cases pass; a birth is never proposed on a death; the packages are gone and `go build ./...` passes.
**Context.** Rules "`core/database/matching` is deleted", "Frequency weighting is gone"; [frequency](comparison-layers.md#frequency).

### Slice 4 — Segments and graph alignment

#### S11-14 — Segment compare

**Ships.** `core/segmentcompare`: neighbor by entity compare, bridge by subject compare on its properties plus direction and inverse terms; typed `BridgeKey`; registry segment section (subject of birth / death is one-cardinality, neighbor credit one hop, link support). The layer loader reads link kinds from `connectrules` instead of a list in code. Output carries value refs for the evidence bundle.
**Not this PR.** The walk (S11-16).
**Check.** A spouse never corresponds to a sibling; parent ↔ child corresponds with ends swapped; differing births subtract; differing residences are unknown.
**Context.** [Segments](comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6).

#### S11-15 — Shortlist index

**Ships.** A key → handles map per kind on `graphcache`, built with `valuecompare.Key` from the registry's shortlist keys (person name words, not initials; event type + year bucket; toponym), dropped on a handle commit.
**Not this PR.** Using it in the walk (S11-16). A coarse misspelling key (open; [decided 2026-10-17](comparison-layers.md#decided-on-2026-10-17)).
**Check.** "J. Robins" shortlists every handle with `robins`; aliases share a key; no `Same` / `Equivalent` pair is missed (property test).
**Context.** [Layer 6 procedure](comparison-layers.md#layer-6-procedure) step 1; [vocabulary](comparison-layers.md#vocabulary-finding-candidates).

#### S11-16 — Rewrite graph alignment

**Ships.** `graphalign.Align` follows the layer 6 procedure: shortlist and nominate, top K with ties, one queue, segments around accepted pairs, dirty-pair settling with reasons, bands from evidence, the accept bar, the contested flag on rows. Registry alignment section. Delete the old walk, fallback, `refineOneHop` and `core/graphalign/registry.go`.
**Not this PR.** FFI and UI for contested (S11-17, S11-18). Memoization (S11-21).
**Check.** The invariants hold (order, decided rows, re-run, start-independence); name + birth date proposes strong; a decision moves only its neighborhood.
**Context.** [Layer 6 procedure](comparison-layers.md#layer-6-procedure), [bands](comparison-layers.md#bands).

#### S11-17 — Contested flag through FFI

**Ships.** The proposal's contested flag and move reasons reach Swift through `engine.proto` and the Promote handler.
**Not this PR.** The marker (S11-18).
**Check.** A handler test returns contested on a row with a conflicting comparison.
**Context.** [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md); [bands](comparison-layers.md#bands).

#### S11-18 — Contested marker on Promote rows

**Ships.** The contested marker beside the band, as S11-D1 draws it, with L10n copy.
**Not this PR.** The evidence sheet (S11-20).
**Check.** A row with a conflict shows contested at any band; a clean row does not.
**Context.** S11-D1; [`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md).

### Slice 5 — Claim evidence

#### S11-19 — Claim evidence query

**Ships.** A lazy per-row RPC from the page's targets: property lines per member through property compare, identical values collapsed, relationship lines through segment compare with bundles, the registry evidence section (pin defaults, defining properties), wording keys from cause codes. Pins write one incoming Observation against many member Observations. `attachExhibits`, `pairExhibits`, `withRowLines` and `refLine` are removed from the proposal read.
**Not this PR.** The sheet (S11-20).
**Check.** Five named members give one collapsed name line of five; a matching birth gives one relationship line whose pin writes its whole bundle; Done backfills every member's claim.
**Context.** [`claim-evidence.md`](claim-evidence.md); [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §9.

#### S11-20 — Claim evidence sheet

**Ships.** The Promote sheet as S11-D2 draws it: "why suggested" summary, then "What's the evidence that this record is <entity>?", property sections, relationship lines with their checkboxes, expandable collapsed lines; all copy through L10n.
**Not this PR.** Engine (S11-19).
**Check.** The sheet reads as sentences; no "Through <name>"; checking a relationship line pins it.
**Context.** S11-D2; [wording](claim-evidence.md#wording).

### Slice 6 — Performance

#### S11-21 — Large-Source performance check

**Ships.** A benchmark on a generated 100-Subject Source against a few thousand handles, with a budget, in the test suite. Memoization of layers 1–4 across re-runs only if the benchmark fails; if built, a test that cached and uncached runs propose the same.
**Not this PR.** Anything the benchmark doesn't need.
**Check.** The benchmark passes; the PR says whether memoization was needed.
**Context.** [Layer 6 procedure](comparison-layers.md#layer-6-procedure) step 6.
