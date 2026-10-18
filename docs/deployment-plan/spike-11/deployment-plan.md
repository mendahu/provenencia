# Deployment Plan — Spike 11

Promote's matching is rebuilt as a stack of pure, registry-driven layers that the rest of the app borrows from, and the old matching code is deleted as each layer lands. The design is in [`comparison-layers.md`](comparison-layers.md) (layers 1–6, bands, the match registry, the reuse inventory) and [`claim-evidence.md`](claim-evidence.md) (the Promote sheet's pins). It pulls in the "make the algorithm graspable" part of [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md), [connection subrows](promote-connection-subrows.md) on the Promote page, and Spike 10's likeness crops on the evidence sheet (S10-36). PRs follow a build-then-migrate rhythm: a PR builds a layer, the next moves every caller onto it and deletes what it replaces.

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
5. **The evidence sheet shows all the evidence.** Joining an entity with five named members shows five name matches (identical values collapsed into one line), worded as sentences. A relationship line says what matched ("They have a matching birth event") and checking it pins everything behind it. Pins land on every member. Two people with a likeness show both crops.
6. **"The same" means the same everywhere.** José and Jose, Montréal and Montreal, compare the same way in Promote, search and the reconciler, under the reconciler's own grouping policy.
7. **One place for every knob.** Every matching number is in `core/matchregistry` and listed in the knob inventory. A guard test fails on a number anywhere else.
8. **The old code is gone:** `core/match`, `core/database/matching`, `core/graphalign/registry.go`, the frequency stats, and `pairExhibits`.
9. **Connections follow their ends.** Each connection shows under a Promote row it joins. Setting John to Skip dims "Mary · parent of · John"; nothing toggles on its own, and there's no switch-off.

## Where the notes went

| Idea or note | PR | Brief |
| --- | --- | --- |
| [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md) — "make the algorithm graspable" (one-page model, glossary, knob inventory) | S11-01 | — |
| Same — "are the levers right?" | S11-02 – S11-16 (the layer rewrite) | — |
| Same — known gaps | [`comparison-layers.md`](comparison-layers.md) open questions | — |
| Claim evidence shows only the first member; "Through <name>" wording | S11-19, S11-25, S11-20 | S11-D2 |
| [`ideas/shared-list-helpers.md`](../../ideas/shared-list-helpers.md) — "Primary Subject kinds", "In-list text matching" | S11-23, S11-08 | — |
| [Connection subrows](promote-connection-subrows.md) (was `ideas/promote-connection-subrows.md`) | S11-26, S11-27 | S11-D1 |
| Spike 10 S10-36 — Promote compares two likeness crops (moved 2026-10-17) | S11-20 | S11-D2 |
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
- **`core/database/matching` is deleted.** `ForSubject` and `ForEntity` have no callers outside tests; merge hints are rebuilt on subject compare over two profiles when their UI exists. Its one live function, `SubjectsValues` (called from `promotealign/layer.go`), moves to the loader that feeds `propertycompare` and `subjectprofile` in S11-13.
- **The reconciler keeps its own decisions** and builds them on value compare. Grouping policy (user, 2026-10-17, S11-06): values group on `Same` and fold on `Contains`; `Equivalent` (an accent fold, an alias) never groups; text never folds; the reconciler reads `ABT` as a range, as matching does.
- **Pure value types and kinds come first.** The date and name value structs, their display helpers, the name part-type registry, and the value-type and cardinality constants move out of `core/database` into pure packages (S11-22), and primary Subject kinds get one home (S11-23), before any layer is built. No layer defines its own copy.
- **One matching skill, grown with the layers.** S11-02 adds a skill (working name `add-match-rule`) for the steps every change repeats: the knob in its registry section, its inventory row, contract tests, the guard list entry. S11-10, S11-14 and S11-16 extend it. Each migration PR also updates the domain docs and idea files it makes false (`matching.md`, `promote-graph-alignment.md`, `conclusion-reconciliation.md`, `seeded-vocabulary.md`, `ideas/daitch-mokotoff.md`, `ideas/archive/name-matching-enhancements.md`).
- **One guard, not one per package.** The purity and numeric-literal checks are one repo-walking test (like `core/writes/begin_test.go` `TestBeginOnlyInDatabaseAndWrites`) with a list of layer packages; each layer PR adds its line.
- **One wire vocabulary.** Promote's flags and reason codes extend `graphalign.RowFlags` and `graphalign.Reason`, and Swift parses them in `Platform/ConclusionVocabulary.swift` under `api/ffi/swift_vocabulary_test.go`, along with the relation and cause codes the evidence sheet words.
- **The pin policy has one owner, in Go.** Default pins come from the registry evidence section, applied in Go at Apply for rows whose sheet was never opened, and by the evidence RPC for rows that were.

### Claim evidence

- **A separate, lazy query per row** ([claim evidence](claim-evidence.md#separate-from-matching)): inputs are the page's targets, not matching internals.
- **Lines** ([lines](claim-evidence.md#lines)): one per member per property, grouped by property, identical values collapsed; `Same` / `Equivalent` start checked, `Contains` / `Overlaps` / `Near` pinnable and unchecked, disagreements shown and not pinnable. No "select all".
- **Relationships** ([relationships](claim-evidence.md#evidence-through-relationships)): a corresponding segment is its own checkable line that pins the bundled plumbing (bridge ends, everything on the bridge, the neighbor's defining properties, e.g. `event_type`); the neighbor's other properties are checkable lines under it.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S11-01 | One-page matching model | 1 | Diagram, glossary, knob inventory with starting values and estimates | — | — | [plan](prs/S11-01-matching-model.md) | Landed #432 |
| S11-22 | Pure value types | 1 | Date and name value structs, display helpers, part types, value-type and cardinality constants move to pure packages | — | — | [plan](prs/S11-22-pure-value-types.md) | Landed #434 |
| S11-23 | One home for primary kinds | 1 | `person` / `event` / `place`, the Property keys matching reads by name, and the Origin constants and check in one pure place; the hand-written lists use it | — | S11-22 | [plan](prs/S11-23-primary-kinds.md) | Landed #438 |
| S11-02 | Value compare core and match registry | 2 | `valuecompare` contract, `Key`, text / term / integer, `matchregistry` skeleton and value section, the guard, the matching skill | — | S11-01, S11-22, S11-23 | [plan](prs/S11-02-value-compare-core.md) | Landed #441 |
| S11-03 | Value compare: names | 2 | Name comparer on relations; initials and aliases in word similarity; alias seeds to `namealiases` | — | S11-02 | [plan](prs/S11-03-value-compare-names.md) | Landed #447 |
| S11-04 | Value compare: dates | 2 | Date comparer on relations; date windows move in; `ABT` as a range | — | S11-02 | [plan](prs/S11-04-value-compare-dates.md) | Landed #445 |
| S11-05 | Migrate matching onto value compare | 2 | `match`, `graphalign`, `promotealign` call `valuecompare`; `match`'s comparers deleted | — | S11-03, S11-04 | [plan](prs/S11-05-migrate-matching.md) | Landed #451 |
| S11-06 | Migrate the reconciler onto value compare | 2 | `Compatible`, keys, grouping policy, place-chain windows | — | S11-05 | [plan](prs/S11-06-migrate-reconciler.md) | Landed #456 |
| S11-07 | One normalization key across the core | 2 | Search rescoring, search-index dedupe, term dedupe, alias tokens use `Key`; text ignores word breaks | — | S11-05 | [plan](prs/S11-07-normalization-key.md) | Landed #460 |
| S11-08 | One normalization helper in Swift | 2 | Shared Swift normalize for filters; `FakeStore` groups names by parts | — | S11-06, S11-07 | [plan](prs/S11-08-swift-normalization.md) | Landed #466 |
| S11-09 | Property compare | 3 | `propertycompare`: negatives, neutral terms, cardinality, outcome table | — | S11-05 | [plan](prs/S11-09-property-compare.md) | Planned |
| S11-10 | Subject compare | 3 | `subjectcompare` + registry subject section; pairs; similarity and evidence | — | S11-09 | [plan](prs/S11-10-subject-compare.md) | Draft |
| S11-11 | Subject profile and entity compare | 3 | `subjectprofile.Build`: strict profile, denied as negatives, `Excluding` | — | S11-10 | — | Scoped |
| S11-12 | Lifespan for people | 3 | Profile builder derives a lifespan; subject compare checks it | — | S11-11 | — | Scoped |
| S11-24 | Delete frequency weighting | 3 | `Stats.ValueFreq`, m / u priors, frequency scales and their benchmark go | — | S11-05 | — | Scoped |
| S11-13 | Migrate Promote scoring onto subject compare | 3 | Property scoring, seeding, duplicate rows; values loader moves; delete `core/match`, `core/database/matching` | — | S11-12, S11-24 | — | Scoped |
| S11-14 | Segment compare | 4 | `segmentcompare`, typed bridge keys, birth / death cardinality, one-hop credit | — | S11-13 | — | Scoped |
| S11-15 | Shortlist index | 4 | Key → handles per kind on `graphcache`; registry shortlist keys | — | S11-13 | — | Scoped |
| S11-16 | Rewrite graph alignment | 4 | Layer 6 procedure, bands from evidence, accept bar; old walk and registry deleted | — | S11-14, S11-15 | — | Scoped |
| S11-17 | Contested flag through FFI | 4 | Proposal rows carry contested and move reasons to Swift | — | S11-16 | — | Scoped |
| S11-18 | Contested marker on Promote rows | 4 | The marker on the page | S11-D1 | S11-17, S11-D1 | — | Scoped |
| S11-19 | Claim evidence query | 5 | Go query: member lines, collapse, relationship bundles; registry evidence section | — | S11-16 | — | Scoped |
| S11-25 | Claim evidence RPC and pin policy | 5 | Lazy per-row RPC; Go default pins at Apply; Swift drafting moves onto it; proposal field 9 and `attachExhibits` removed | — | S11-19 | — | Scoped |
| S11-20 | Claim evidence sheet | 5 | The sheet's new layout and copy, with likeness lines as two crops | S11-D2 | S11-25, S11-D2, S10-54 | — | Scoped |
| S11-26 | Connections file by their ends | 6 | Drop the bridge switch-off: `SkipBridgeIDs`, `skip_bridge_ids`, the `skip` parameters | — | — | — | Scoped |
| S11-27 | Connection subrows on Promote | 6 | Each connection as a subrow of a row it joins, its state from its ends; the panel and toggles go | S11-D1 | S11-26, S11-D1 | — | Scoped |
| S11-21 | Large-Source performance check | 7 | 100-Subject benchmark; memoize only if it fails | — | S11-16 | — | Scoped |
| S11-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S11-D1 | Promote page — rows | **S11-18**, **S11-27** | Contested marker beside the band; connections as subrows of the rows they join, answering the [design questions](promote-connection-subrows.md#open-questions-for-design) |
| S11-D2 | Promote page — evidence sheet | **S11-20** | Whole sheet: "why suggested" summary, property sections, relationship lines, collapsed members, likeness crops (from Spike 10's S10-D13), copy |

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify when the slice lands.

```text
SLICE 1 — Graspable model and pure foundations
  S11-01  One-page matching model       (parallel with S11-22)
  S11-22  Pure value types
  S11-23  One home for primary kinds
  Check: one doc explains every layer, term and knob; each knob has a starting value;
         date and name values, value types and kinds import nothing from core/database.

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
  S11-24  Delete frequency weighting    (any time after S11-05)
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
  S11-25  Claim evidence RPC and pin policy
  ✎ S11-D2 ──▶ S11-20  Claim evidence sheet
  Check: five named members show as matches with identical values collapsed;
         "They have a matching birth event" pins its bundle; pins land on every member.

SLICE 6 — Connection subrows
  S11-26  Connections file by their ends   (any time)
  ✎ S11-D1 ──▶ S11-27  Connection subrows on Promote
  Check: setting John to Skip dims "Mary · parent of · John" under its row; nothing
         toggles by itself; the bottom connections panel is gone.

SLICE 7 — Performance
  S11-21  Large-Source performance check
  Check: a 100-Subject Source proposes within the benchmark; memoization exists
         only if the benchmark needed it.

S11-99 — Dogfood close / docs
```

Parallel: S11-01 with S11-22; S11-03 with S11-04; S11-24 with slice 3 after S11-05; S11-06, S11-07 and S11-09 after S11-05; S11-14 with S11-15; slice 5 with slice 7 after S11-16; slice 6 any time (S11-26 has no dependency). S11-D1 and S11-D2 can be drawn while slice 4's engine PRs run.

### Dependencies at a glance

```text
S11-22 ──▶ S11-23 ──┐
S11-01 ─────────────┴──▶ S11-02 ──▶ S11-03 ──┐
                 └──▶ S11-04 ──┴──▶ S11-05 ──▶ S11-06 ──▶ S11-08
                                       ├──▶ S11-07
                                       ├──▶ S11-24 ─────────────────────────────────────┐
                                       └──▶ S11-09 ──▶ S11-10 ──▶ S11-11 ──▶ S11-12 ──┴──▶ S11-13
S11-13 ──▶ S11-14 ──┐
S11-13 ──▶ S11-15 ──┴──▶ S11-16 ──▶ S11-17 ──▶ S11-18
                            ├──▶ S11-19 ──▶ S11-25 ──▶ S11-20
                            └──▶ S11-21
S11-D1 ──▶ S11-18, S11-27
S11-26 ──▶ S11-27
S10-54 ──▶ S11-20   (Spike 10)
S11-D2 ──▶ S11-20
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Link, do not restate.

### Slice 1 — Graspable model and pure foundations

#### S11-01 — One-page matching model

**Landed 2026-10-17** in [#432](https://github.com/mendahu/provenencia/pull/432). See [`completed.md`](completed.md).

**Ships.** A domain doc (`docs/matching.md` rewritten, or a new `docs/matching-model.md` that replaces it) with one diagram of the six layers, the glossary (relation, similarity, evidence, segment, shortlist, candidate, band, contested…), and the knob inventory: every registry field, its section, starting value or estimate with reasoning, and what it changes.
**Not this PR.** Any code. `core/matchregistry` (S11-02).
**Check.** A reader can trace a Promote row from value compare to band on one page; every knob in the comparison-layers doc appears in the inventory with a value.
**Context.** Rule "Registry numbers: carry over, then estimate"; [`comparison-layers.md`](comparison-layers.md) whole; [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md) "First: make the algorithm graspable"; [`promote-graph-alignment.md`](../../promote-graph-alignment.md) (superseded sections are marked).

#### S11-22 — Pure value types

**Ships.** `datevalues.Value` and `namevalues.Value`, their display helpers (`CompactDisplay` and kin), the name part-type registry, and `properties`' value-type and cardinality constants move to pure packages with no `core/database` import. Every caller (about 65 files, including `valuecodec`, `eventtitle`, `autoreconcile`) points at them. The database packages keep only their queries. `add-date-value` and `add-name-value` skills say where the types live.
**Not this PR.** Any comparison change.
**Check.** `go list -deps` of the new packages has nothing under `core/database`; `go build ./...` and the tests pass.
**Context.** Rule "Pure value types and kinds come first"; [purity](comparison-layers.md#purity-requirements).

#### S11-23 — One home for primary kinds

**Ships.** `person`, `event`, `place` and the Property keys matching reads by name (their locked keys plus `sex_at_birth` and `gender`) in one pure place, `core/vocabkey`; `promote.primaryKinds`, `match.ProfiledKinds`, `promotealign/layer.go` and the `FakeStore` literals use it. `sourcegraphprogress` drops its `promote` import. `OriginProvenencia` / `OriginUser`, declared today in nine packages (`properties`, `subjecttypes`, `propertyterms`, `sourcetypes`, `metadatafields`, `nameformats`, `claimconfidencegrades`, `sourcecredibilitygrades`, `core/connectrules`), get the same pure home, with the origin check the seven `originOK` copies repeat. Closes "Primary Subject kinds" in [`ideas/shared-list-helpers.md`](../../ideas/shared-list-helpers.md).
**Not this PR.** Registry keys (S11-02 uses these).
**Check.** No hand-written kind list remains (grep); tests pass.
**Context.** [`ideas/shared-list-helpers.md`](../../ideas/shared-list-helpers.md) "Primary Subject kinds".

### Slice 2 — Value compare, everywhere

#### S11-02 — Value compare core and match registry

**Ships.** `core/valuecompare`: pure value types, `Result` (relation + similarity), `Detail` (parts, cause codes, value refs), `Key`, and the text, term and integer comparers. `core/matchregistry` with `Default()` and the value section. The word helpers (`NormalizeForm`, `IsInitial`, `EditDistance`, `SpellingSimilarity` from `core/autoreconcile`; the tokenizer and fold from `core/match/words.go`) move in. The purity guard lists `core/vocabkey`. One repo-walking guard (purity and numeric literals over a list of layer packages; the guard also lists `core/datevalue`, `core/namevalue` and `core/propertymeta`), contract tests, and the inventory-drift guard. The matching skill (`add-match-rule`), first version.
**Not this PR.** Names (S11-03), dates (S11-04), any caller (S11-05).
**Check.** `go test ./core/valuecompare/... ./core/matchregistry/...` passes with no build tags; the contract tests run against every comparer.
**Context.** [Layer 1](comparison-layers.md#layer-1-value-comparison), [relations](comparison-layers.md#relations), [detail](comparison-layers.md#detail), [purity](comparison-layers.md#purity-requirements), [match registry](comparison-layers.md#match-registry).

#### S11-03 — Value compare: names

**Ships.** The name comparer on relations and parts detail: folds, particles, aliases (tokens normalized), initials, spelling (`Near` minimum 0.8 for words, 0.75 for names), generation conflict; a surname or first given name left over on both sides is `Different`, other words left over on both sides `Near`; swapped part types fall under the `Near` minimum and are `Different` (user call 2026-10-17). Builds on S11-02's word helpers; adds initials and aliases to word similarity. `AliasTable`, `NameRole` and `NamePattern` move to `valuecompare`; the alias and nickname seeds, `Register`, and `third_party/nicknames/` move to `namealiases`, with `add-seeded-vocabulary` and [`seeded-vocabulary.md`](../../seeded-vocabulary.md) updated. No frequency.
**Not this PR.** Callers (S11-05). The reconciler's grouping (S11-06).
**Check.** Today's `names_test.go` ladders and generated invariants pass on the new relations; the worked name cases in the relations table hold.
**Context.** [Relations](comparison-layers.md#relations) worked cases; [`matching.md`](../../matching.md) Names.

#### S11-04 — Value compare: dates

**Ships.** The date comparer on range relations; `DateEnds` and the window algebra (`Window` with `Overlaps`, `Contains`, `Intersect`; `HoldsAt`, which is `Overlaps`, and `Relate` are deleted) move in from `core/autoreconcile`, and the calendar arithmetic moves to `core/datevalue`; `ABT` read as one unit of its most precise part; tolerance in the registry. `DateBounds`, `WindowOfDate`, `PeriodWindow` and `LinkMembership` stay with the reconciler until S11-06.
**Not this PR.** Period pooling across `date` / `start_date` / `end_date` (S11-10). Callers (S11-05, S11-06).
**Check.** `ABT 1817` / `1818` is `Contains`; `BEF 1820` / `BEF 1830` is `Contains`; today's `dates_test.go` cases pass, updated where the ABT rule changes them.
**Context.** [Relations](comparison-layers.md#relations) and the `ABT` rule; [`matching.md`](../../matching.md) `DateComparer` row.

#### S11-05 — Migrate matching onto value compare

**Ships.** `core/match` (`Score`, `Evaluate`, `pairValues`, `poolPeriods`), `core/graphalign` and `core/database/promotealign` (`pairExhibits`, stats) call `valuecompare`; a catalog's name pattern and alias tables reach its `Rules` through one loader, `valuerules.Load`. `match`'s comparers, `DiscountComparer` and the name frequency discount (with `Stats.Words` and `countNameWords`, its only inputs) are deleted, and so are Promote's per-type floors, which value compare's `Near` minimum replaces. [`matching.md`](../../matching.md) and `ideas/daitch-mokotoff.md` point at the new code.
**Not this PR.** Layer 2 and up (slice 3). The reconciler (S11-06).
**Check.** The obituary golden and sparse cases pass (expected values updated where relations change them); `core/match` has no comparer types left.
**Context.** Rules "Build, then migrate", "High tolerance for churn"; [what moves where](comparison-layers.md#what-moves-where-in-todays-code).

#### S11-06 — Migrate the reconciler onto value compare

**Ships.** `autoreconcile` keys, `Compatible`, `subsumes` / `partFits` and the pipeline's grouping use `valuecompare` relations under the grouping policy decided 2026-10-17 (working rule "The reconciler keeps its own decisions"). `conclusionheaders` place chains use day-precision windows; `DateBounds` and `WindowOfDate` become `valuecompare.DateWindow`; `PeriodWindow` and `LinkMembership` move to `conclusionheaders` on it (S11-04 left them in `autoreconcile`). The caller passes the value rules in (answers S11-02 D-8). [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) says what the reconciler now uses.
**Not this PR.** Pins (S11-19).
**Check.** `Sainte-Anne` / `Sainte Anne` group; `J. Robins` folds into `James Robins`; a place chain with two same-year dates folds by day; `Montréal` / `Montreal` follow the chosen policy.
**Context.** [Reuse inventory](comparison-layers.md#reuse-across-the-app); [decided 2026-10-17](comparison-layers.md#decided-on-2026-10-17) (reconciler grouping); rule "The reconciler keeps its own decisions"; [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md).

#### S11-07 — One normalization key across the core

**Ships.** `core/search` Go rescoring, `searchindex` alternate-value dedupe, the `propertyterms` duplicate check (create and rename, any origin) and `namealiases` tokens use `valuecompare.Key`. Text value compare ignores word breaks: `god-father` and `godfather` are `Same` (user call 2026-10-17).
**Not this PR.** Search's fuzzy ranker stays ([reuse inventory](comparison-layers.md#reuse-across-the-app)). Swift (S11-08).
**Check.** Searching `Montreal` ranks `Montréal` like an exact hit; adding "god-father" beside "godfather" is caught; a capitalized alias token hits.
**Context.** [Reuse inventory](comparison-layers.md#reuse-across-the-app) text-equality table.

#### S11-08 — One normalization helper in Swift

**Ships.** `TextMatch` (`macos/App/Platform/TextMatch.swift`), a Swift mirror of `valuecompare.TextKey` (case, accents, punctuation and word breaks ignored; S11-07 D-1) that Go checks against its own key, replaces `PVComboBoxMatch.normalize` and the other Swift text rules: `PVComboBox` filter and highlight, the list filters (`PropertiesModel`, `SourcesModel`, `GraphCanvasItemLayer`, `PVFloatingMenuSelection`), `DateValueTranscriptionMatch` and `NameTitleRegistry`; `FakeStore` groups names by parts and searches like Go's rescoring. Closes "In-list text matching" in [`ideas/shared-list-helpers.md`](../../ideas/shared-list-helpers.md).
**Not this PR.** Go behavior changes (a Go test reads the Swift fixtures).
**Check.** Filtering `jose` finds `José` in every list; `FakeStore`-backed tests group names like Go does.
**Context.** [Reuse inventory](comparison-layers.md#reuse-across-the-app) "Lighter touch"; [`add-swift-test`](../../../.agents/skills/add-swift-test/SKILL.md).

### Slice 3 — Property, Subject and entity compare

#### S11-09 — Property compare

**Ships.** `core/propertycompare`: one property between two Subjects from their value sets; negatives, neutral terms, cardinality, missing values; outcome from relation; output with the speaking pair and excluded values.
**Not this PR.** Callers (S11-13). Support or standing (gone with the strict profile).
**Check.** The invariants and the outcome table in the layer 2 section hold in tests.
**Context.** [Layer 2](comparison-layers.md#layer-2-property-comparison).

#### S11-10 — Subject compare

**Ships.** `core/subjectcompare` and the registry's subject section, keyed by `core/vocabkey` constants (S11-23) used directly as typed `SubjectType` / `PropertyKey` keys, with no re-declared key constants: per Subject type and property, agree weight and conflict penalty (event type: low agree, very large conflict); the pairs (sex with gender; `date` with `start_date` / `end_date`); similarity and evidence.
**Not this PR.** Profiles (S11-11), lifespan (S11-12), callers (S11-13).
**Check.** The layer 3 examples hold (name 1.0 / 2.5; name + gender 1.0 / 3.1; birth vs death with matching dates ≈ 0.09).
Ports `subjectvocab`'s `TestDefaultProfileCardinality` as "every registry `PropertyKey` is a `vocabkey` constant"; S11-23's `TestVocabKeysMatchSeed` ties those constants to the seed. Extends the matching skill.
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

**Ships.** `graphalign` scores candidates' properties with entity compare; `promotealign` seeds with it instead of `match.Rank`; duplicate-row checks (`alikeNew`, `pairNode`) use subject compare. `matching.SubjectsValues` moves to the loader that feeds `propertycompare` and `subjectprofile`. Delete `core/match` and `core/database/matching`.
**Not this PR.** Segments and the walk (slice 4): the old walk stays until S11-16, calling the new scoring.
**Check.** Golden and sparse cases pass; a birth is never proposed on a death; the packages are gone and `go build ./...` passes.
**Context.** Rule "`core/database/matching` is deleted".

#### S11-24 — Delete frequency weighting

**Ships.** Remove `Stats.ValueFreq`, the m / u priors (`MPrior`, `UPrior`), the frequency scales and their lift, and `BenchmarkLoadStatsNames`; event type and toponym get plain weights.
**Not this PR.** Fan-out (S11-14).
**Check.** Golden and sparse cases pass with plain weights; no frequency symbol remains (grep).
**Context.** Rule "Frequency weighting is gone"; [frequency](comparison-layers.md#frequency).

### Slice 4 — Segments and graph alignment

#### S11-14 — Segment compare

**Ships.** `core/segmentcompare`: neighbor by entity compare, bridge by subject compare on its properties plus direction and inverse terms; typed `BridgeKey`; registry segment section (subject of birth / death is one-cardinality, neighbor credit one hop, link support). The layer loader reads link kinds from `connectrules` instead of a list in code. Output carries value refs for the evidence bundle. Fan-out goes: `Stats.FanOut`, `promotealign/stats.go`'s `addFanOut` / `sqlFanOut`, `FanOutUnknown`. Extends the matching skill.
**Not this PR.** The walk (S11-16).
**Check.** A spouse never corresponds to a sibling; parent ↔ child corresponds with ends swapped; differing births subtract; differing residences are unknown.
**Context.** [Segments](comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6).

#### S11-15 — Shortlist index

**Ships.** A key → handles map per kind on `graphcache`, built with `valuecompare.Key` from the registry's shortlist keys (person name words, not initials; event type + year bucket; toponym). It reuses `graphcache`'s stats slot (`Stats` / `SetStats` / `statsOK`, renamed), dropped on a handle commit like today.
**Not this PR.** Using it in the walk (S11-16). A coarse misspelling key (open; [decided 2026-10-17](comparison-layers.md#decided-on-2026-10-17)).
**Check.** "J. Robins" shortlists every handle with `robins`; aliases share a key; no `Same` / `Equivalent` pair is missed (property test).
**Context.** [Layer 6 procedure](comparison-layers.md#layer-6-procedure) step 1; [vocabulary](comparison-layers.md#vocabulary-finding-candidates).

#### S11-16 — Rewrite graph alignment

**Ships.** `graphalign.Align` follows the layer 6 procedure: shortlist and nominate, top K with ties, one queue, segments around accepted pairs, dirty-pair settling with reasons, bands from evidence, the accept bar. Contested is a `RowFlags` field and move reasons are `Reason` codes; the plan decides whether `ConflictWithFixed` is subsumed. Registry alignment section. Delete the old walk, fallback, `refineOneHop` and `core/graphalign/registry.go`. [`promote-graph-alignment.md`](../../promote-graph-alignment.md) and the matching skill are updated.
**Not this PR.** FFI and UI for contested (S11-17, S11-18). Memoization (S11-21).
**Check.** The invariants hold (order, decided rows, re-run, start-independence); name + birth date proposes strong; a decision moves only its neighborhood.
**Context.** [Layer 6 procedure](comparison-layers.md#layer-6-procedure), [bands](comparison-layers.md#bands).

#### S11-17 — Contested flag through FFI

**Ships.** The contested flag and move reasons reach Swift through `engine.proto` and the Promote handler. Promote's reason codes move from `PromoteFlow.swift` into `Platform/ConclusionVocabulary.swift`, under `api/ffi/swift_vocabulary_test.go`.
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

**Ships.** A Go query from a row's target and its neighbors' targets: property lines per member through property compare, identical values collapsed, relationship lines through segment compare with bundles, the registry evidence section (pin defaults, defining properties), wording keys from cause codes.
**Not this PR.** The RPC, Apply defaults and Swift (S11-25). If the plan finds the pin tables can't hold one incoming Observation against many member Observations, a migration PR is cut ahead of it.
**Check.** Five named members give one collapsed name line of five; a matching birth gives one relationship line whose bundle holds every Observation behind it.
**Context.** [`claim-evidence.md`](claim-evidence.md); [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §9.

#### S11-25 — Claim evidence RPC and pin policy

**Ships.** A lazy per-row RPC over S11-19's query. Apply computes default pins in Go for rows whose sheet was never opened. Swift's pin drafting (`PromoteFlow.swift` `draftedPins`, `draftedArgument`, `kept.pins`) moves onto the RPC. `PromoteGraphAlignmentRow.comparisons` (field 9), `attachExhibits`, `pairExhibits`, `withRowLines` and `refLine` are deleted. Pins write one incoming Observation against many member Observations; Done backfills every member's claim.
**Not this PR.** The new sheet layout (S11-20); the shipped sheet reads the new lines unchanged in shape.
**Check.** Done on a row never opened pins its matching lines; opening the sheet shows the same defaults; every member's claim is backfilled.
**Context.** Rule "The pin policy has one owner, in Go"; [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md).

#### S11-20 — Claim evidence sheet

**Ships.** The Promote sheet as S11-D2 draws it: "why suggested" summary, then "What's the evidence that this record is <entity>?", property sections, relationship lines with their checkboxes, expandable collapsed lines; all copy through L10n.
**Not this PR.** Engine (S11-19).
Likeness comparisons show the two crops, which S10-36 was to add in Spike 10; pinning one writes the same kind of pin a name pin writes.
**Check.** The sheet reads as sentences; no "Through <name>"; checking a relationship line pins it; two people with a likeness show both images.
**Context.** S11-D2; [wording](claim-evidence.md#wording); Spike 10 S10-54 (depiction values over the FFI) lands first.

### Slice 6 — Connection subrows

#### S11-26 — Connections file by their ends

**Ships.** Remove the switch-off: `promote.Batch.SkipBridgeIDs`, `ApplyPromoteGraphAlignmentRequest.skip_bridge_ids`, and the `skip` parameter of `FileSourceBridgesTx` / `FileBridgesTx`. Swift's `bridgeOff` and `toggleBridge` go with the request field. [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §3 and §9.1 and [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5.4 lose the "switched off" rule.
**Not this PR.** Subrows (S11-27).
**Check.** Done files every connection whose ends are both handles; no request field or parameter can skip one.
**Context.** [Connection subrows](promote-connection-subrows.md) "The problem" §2.

#### S11-27 — Connection subrows on Promote

**Ships.** Each connection shows as a subrow under a row it joins, as S11-D1 draws it, with its state from `ConnectionState` (files, waits, self-link). `PromoteFlow.connections(hostedBy:)`; the bottom connections panel goes; the footer keeps its count. VoiceOver reads why a connection waits.
**Not this PR.** Opening the evidence sheet at a relationship line (S11-20 decides).
**Check.** Setting John to Skip dims his connections; setting him to New turns them on; a self-link warns.
**Context.** [Connection subrows](promote-connection-subrows.md); S11-D1.

### Slice 7 — Performance

#### S11-21 — Large-Source performance check

**Ships.** A budgeted `Test…` (not a `Benchmark…`, which `scripts/check-pr.sh` never runs) on a 100-Subject Source against a few thousand handles, reusing `promotealign/bench_test.go`'s `newFixture` and `bulkCanon`. Memoization of layers 1–4 across re-runs only if the benchmark fails; if built, a test that cached and uncached runs propose the same.
**Not this PR.** Anything the benchmark doesn't need.
**Check.** The benchmark passes; the PR says whether memoization was needed.
**Context.** [Layer 6 procedure](comparison-layers.md#layer-6-procedure) step 6.
