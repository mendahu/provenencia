# Deployment Plan — Spike 11

Promote's matching is rebuilt as a stack of pure, registry-driven layers that the rest of the app borrows from, and the old matching code is deleted as each layer lands. The design is in [`comparison-layers.md`](comparison-layers.md) (layers 1–6, bands, the match registry, the reuse inventory) and [`claim-evidence.md`](claim-evidence.md) (the Promote sheet's pins). It pulls in the "make the algorithm graspable" part of [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md), [connection subrows](promote-connection-subrows.md) on the Promote page, Spike 10's likeness crops on the evidence sheet (S10-36), and, added 2026-10-24, the evidence button's hover that S11-27 missed (S11-34) [find another record](promote-find-another-record.md) in the target menu (S11-35 – S11-37), and [kinship shapes](kinship-shapes.md): a shared marriage, a birth's parents and a parent chain match spouse, parent and grandparent links (S11-38 – S11-42), and [name candidates](name-candidates.md): a fuller record and a shortened given name reach the target menu (S11-43 – S11-47). PRs follow a build-then-migrate rhythm: a PR builds a layer, the next moves every caller onto it and deletes what it replaces.

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
10. **The rows match their board.** The Promote row's evidence button answers the pointer as S11-D1 draws it.
11. **Any record can be the target.** A Subject Promote offers nothing for (Promote says New) can be filed on a record the researcher finds by search. The row holds that record through later decisions and shows its band, Contested when it conflicts, and its evidence.
12. **One kinship fact, any shape.** A Source that records a marriage proposes its two people on a couple the tree records as spouses at the band a spouse relationship would get, and the reverse. A birth record naming the father matches a tree that says he is the child's parent. A Source with A parent of B and B parent of C matches a tree that records A as C's grandparent. A fact recorded twice on one side counts once.
13. **A plausible name is offered.** A Subject named Amy is offered Amy Lee Anne Long, and one named Desiree is offered Des, as alternatives in the target menu even with nothing else to compare. Promote never drafts New or Skip; a row it does not place opens unset.

## Where the notes went

| Idea or note | PR | Brief |
| --- | --- | --- |
| [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md) — "make the algorithm graspable" (one-page model, glossary, knob inventory) | S11-01 | — |
| Same — "are the levers right?" | S11-02 – S11-16 (the layer rewrite) | — |
| Same — known gaps | [`comparison-layers.md`](comparison-layers.md) open questions | — |
| Claim evidence shows only the first member; "Through <name>" wording | S11-19, S11-25, S11-32, S11-20 | S11-D2 |
| [`ideas/shared-list-helpers.md`](../../ideas/shared-list-helpers.md) — "Primary Subject kinds", "In-list text matching" | S11-23, S11-08 | — |
| [Connection subrows](promote-connection-subrows.md) (was `ideas/promote-connection-subrows.md`) | S11-26, S11-27 | S11-D1 |
| Spike 10 S10-36 — Promote compares two likeness crops (moved 2026-10-17) | S11-32 (lines), S11-20 (crops) | S11-D2 |
| Reuse inventory ([`comparison-layers.md`](comparison-layers.md#reuse-across-the-app)) | S11-05 – S11-08, S11-13, S11-19 | — |
| S11-27 follow-up — the evidence button's hover on the S11-D1 board was not built (user, 2026-10-24) | S11-34 | — (S11-D1, agreed) |
| Dogfood, 2026-10-24: "Amy" not offered "Amy Lee Anne Long"; "Desiree" not offered "Des" ([name candidates](name-candidates.md)) | S11-43, S11-44, S11-46, S11-45, S11-47 | — |
| [Kinship shapes](kinship-shapes.md) (was `ideas/relationships-from-shared-events.md`, from S11-14; grandparent added by the user; pulled in 2026-10-24) | S11-41, S11-38, S11-39, S11-40, S11-42 | — |
| Dogfood, 2026-10-27 (S11-40 review): a spouse's name lines under a relationship Card compare only with the member from the Source that records the link, so "Jake Robins" conflicts with the marriage certificate's "James Kenneth Robins" while the entity's three "Jake" Sources go unseen | S11-52 | — |
| [Find another record](promote-find-another-record.md) (was `ideas/promote-find-another-record.md`, from S11-05 D-16; pulled in 2026-10-24) | S11-35, S11-37, S11-36 | S11-D3 |

**Stays in the ideas folder** — the kinship view for trees and person pages, and patterns beyond spouse, parent and grandparent: [`ideas/kinship-view.md`](../../ideas/kinship-view.md). And the rest of [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md): scenario files and goldens runner, random-graph invariants, the simulated world, the scoreboard, the labeler, fitting, and dogfood export. Whether they join this spike is decided after the re-implementation.

## Working rules

These are the calls for this spike. PR plans treat them as binding.

### Spike-wide

- **High tolerance for churn.** No one uses the app yet. No adapters, shims, wrappers, feature flags, deprecated aliases or kept-for-now paths. Migration PRs move every caller and delete the old code in the same PR. Behavior may change where the design says so; goldens are updated, not preserved. Catalog shapes matching used to need may be reshaped or dropped.
- **Build, then migrate.** A build PR adds a layer and its registry section with its own tests. The next PR moves every caller onto it and deletes what it replaces. Between the two, the old code may call the new layer (for example `pairExhibits` uses value compare after S11-05 and is deleted in S11-25); nothing gets a compatibility layer.
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
- **Entity compare is against the reconciled conclusion** ([entities](comparison-layers.md#entities-compare-against-the-conclusion)). The profile is strict: displayed values positive, denied values negative, nothing else. `subjectprofile.Build` also takes `Options.Excluding`, a list of members to leave out.
- **Lifespan is for people only** ([lifespan](comparison-layers.md#lifespan-a-derived-property-for-people)): a window and a span from every dated event the person takes part in (user, 2026-10-18); a span outside the other side's window is a large conflict, agreement weighs almost nothing.
- **Segment compare is one hop and pure** ([segments](comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6)). Only subject-of-birth and subject-of-death are declared one-cardinality, from both ends (user, 2026-10-19); a differing neighbor there subtracts. Neighbor credit travels one hop; chains come from the walk.
- **Graph alignment follows the layer 6 procedure** ([procedure](comparison-layers.md#layer-6-procedure)): shortlist → score, keep top K with ties → one queue → segments around accepted pairs → settle dirty pairs. The answer doesn't depend on where the walk starts.
- **Bands come from evidence** ([bands](comparison-layers.md#bands)): similarity only gates candidates and gives the first sort. The accept bar is its own setting. A conflict shows a contested marker; it never caps the band.
- **Frequency weighting is gone**, for names, event types and toponyms ([frequency](comparison-layers.md#frequency)).
- **One match registry** ([match registry](comparison-layers.md#match-registry)): `core/matchregistry`, one file per section (value, subject, segment, alignment, evidence), typed keys, each layer gets its section as an argument. Catalog data (aliases, name patterns) is passed in; data-model facts (cardinality, term flags, bridge kinds) stay in the catalog. A guard test fails on numeric literals other than 0 and 1 in layer packages; another fails when the knob inventory drifts from the registry.
- **`core/database/matching` is deleted.** `ForSubject` and `ForEntity` have no callers outside tests; merge hints are rebuilt on subject compare over two profiles when their UI exists. Its one live function, `SubjectsValues` (called from `promotealign/layer.go`), is replaced in S11-13 by the loader that feeds `subjectprofile`, which reads the layer's Observations from the cached Source graph.
- **The reconciler keeps its own decisions** and builds them on value compare. Grouping policy (user, 2026-10-17, S11-06): values group on `Same` and fold on `Contains`; `Equivalent` (an accent fold, an alias) never groups; a nickname or abbreviation folds into the formal name the product seed gives it (user, 2026-10-19, S11-28), and an accent fold never folds; text never folds; the reconciler reads `ABT` as a range, as matching does.
- **Pure value types and kinds come first.** The date and name value structs, their display helpers, the name part-type registry, and the value-type and cardinality constants move out of `core/database` into pure packages (S11-22), and primary Subject kinds get one home (S11-23), before any layer is built. No layer defines its own copy.
- **One matching skill, grown with the layers.** S11-02 adds a skill (working name `add-match-rule`) for the steps every change repeats: the knob in its registry section, its inventory row, contract tests, the guard list entry. S11-10, S11-14 and S11-16 extend it. Each migration PR also updates the domain docs and idea files it makes false (`matching.md`, `promote-graph-alignment.md`, `conclusion-reconciliation.md`, `seeded-vocabulary.md`, `ideas/daitch-mokotoff.md`, `ideas/archive/name-matching-enhancements.md`).
- **One guard, not one per package.** The purity and numeric-literal checks are one repo-walking test (like `core/writes/begin_test.go` `TestBeginOnlyInDatabaseAndWrites`) with a list of layer packages; each layer PR adds its line.
- **One wire vocabulary.** Promote's flags and reason codes extend `graphalign.RowFlags` and `graphalign.Reason`, and Swift parses them in `Platform/ConclusionVocabulary.swift` under `api/ffi/swift_vocabulary_test.go`, along with the relation and cause codes the evidence sheet words.
- **The pin policy has one owner, in Go.** Default pins come from the registry evidence section, applied in Go at Apply for rows whose sheet was never opened, and by the evidence RPC for rows that were.

### Find another record

- **A found record is a decided row, nothing new** ([shape](promote-find-another-record.md#shape)). It goes to Go as a `PromoteGraphAlignmentFixed` pair like a menu pick. Go holds it and never moves it, and scores it for the row's band, Contested and evidence like any decided handle. No new proposal field, flag or reason code.
- **No new search RPC, and one search model.** The picker calls `SearchCatalog` with `kinds` set to the row's kind; hits already carry their header (S9-35). It reuses the omnibar's results model and hit row, moved to a shared home by S11-37, not a copy in Promote.
- **An inline picker, every record of the kind, no warning** (user, 2026-10-24; [decided](promote-find-another-record.md#decided-user-2026-10-24)). The search opens from the row on the Promote page, not the omnibar, which is not connected to the page. It lists every canonical entity of the row's kind, ranked by `SearchCatalog`'s match to the search string, with nothing filtered out. A differently named pick shows no warning beyond the band and Contested.
- **The menu entry is a sentinel option.** `PVSelect` gets no action item; the row's binding catches the entry's token before `setTarget`, so opening or cancelling the search never marks the row decided. A kit PR is cut only if S11-D3 draws what a sentinel cannot do.
- **The picked record joins the row's menu**, through one `PromoteFlow` call that adds the record to `row.menu` and sets the target, with `setTarget`'s follow-up. The menu stays the one list of a row's records (S11-35).
- **Engine first, proven by tests.** S11-35 adds no Go code unless a test finds a gap: a fixed handle no shortlist key reaches is loaded, banded, contested when it conflicts, has evidence and files. A gap found there is fixed in S11-35.

### Kinship shapes

- **Implied links, not a new comparison** ([shape](kinship-shapes.md#shape)). Each side's links are read through the kinship patterns before alignment; a pattern that holds adds a relationship link with the key a direct link would have. `segmentcompare` is unchanged and both stay one hop; `graphalign` learns only the count-once rule below. An implied grandparent is one hop made from two parent links.
- **Three patterns, product terms only** (user, 2026-10-24): two `subject`s of one `marriage` imply `spouse` (the couple are the marriage's subjects, user decision 2026-10-26); a `father` or `mother` and the `subject` of one `birth` imply `parent`, directed; a parent of a parent (each direct or implied) implies `grandparent`. A grandparent never implies a parent; nothing deeper than grandparent.
- **The subject is an event's principal** (user, 2026-10-26): a birth's subject is the person born, a death's the person dying, a marriage's subjects the couple, a migration's the person migrating. No `spouse` participation role.
- **One pure package, `core/kinship`**, holds the patterns and the derivation, joins the purity guard, and is what trees and person pages read later. It returns only implied links, keyed through `segmentcompare.Key` with the catalog's `Terms`.
- **The canon's implied links live in the graph cache** (user, 2026-10-26; [where implied links live](kinship-shapes.md#where-implied-links-live)). Every pattern reads one pivot node's own links: a marriage's or birth's event node, or the parent between in a chain (a second stage over that person's direct and implied parent links). The graph cache keeps one result per pivot, derived on first read and dropped when the pivot or a node one hop out changes; graph verify mode checks each against a fresh derivation. `canonlinks.Reader.For` answers one person's implied links from those results; `loadCanon` (S11-39) and the evidence query's `canonLinks` (S11-40) call it, and neither derives. One copy per open catalog, never written to the catalog. The Source side derives when its layer loads, memoized with the layer.
- **Bounded per pivot** (user, 2026-10-26). A pivot with more than `kinship.MaxPivotEnd` (10, a `matchregistry` value passed in as `kinship.Rules`) participants at either end, or a person with more than 10 parents, implies nothing, so derivation stays linear in its input.
- **Matching warms behind the open spinner** (user, 2026-10-26): "Move everything behind the spinner so that it's honest." `catalogsession.WarmOnOpen` loads the shortlist's kinds into the graph cache and builds the name index while the project opens; an operation that drops them (not a handle commit) warms them again in the same call. A longer open is accepted; it is revisited only if it grows too long (S11-50).
- **One fact, one link per side.** A direct link and an implied one between the same two people with the same key merge into one link carrying both shapes' bridge Subjects, credited once, through the one `promotealign` `linkSet` merge both loaders use. From S11-51 that one merge is `kinship.LinkSet` (implied links enter through `AddImplied`, or `PlaceImplied` in `loadCanon`), and `canonlinks.Reader.Relationships` gives one person's merged links, so the Person page (Spike 13) reads the same links (user, 2026-10-26, S11-51 plan review).
- **Counted once across sides** (user, 2026-10-26; [counting once](kinship-shapes.md#counting-once)). When both sides imply a link through the same pattern and the walk has paired the event (or middle parent) it passes through, the implied link adds nothing: the shape's own links already line up. Graph alignment's `corresponding` skips that pair (S11-39); the evidence lines follow (S11-40). When the pivots are not paired (a sparse marriage whose event row is not placed), the implied link credits. Two recorded links are never skipped.
- **No new weight.** An implied link weighs as the direct link it stands for (the segment registry's Other pair). The one new registry value is the pivot bound above.

### Name candidates

- **The candidate gate drops to 0.4** (user, 2026-10-24; [case 1](name-candidates.md#case-1--amy-against-amy-lee-anne-long)). `alignment.MinSimilarity` 0.5 → 0.4, so a one-name Subject reaches a record with three more names. The gate only picks candidates; no band moves, and the row still opens unset (S11-47).
- **Des is a nickname of Desiree and of Desirae** (user, 2026-10-24): two sets, `{desiree, des}` and `{desirae, des}`, in `nicknameSets`, the product's already-reviewed edit of the upstream list, with a `CacheVersion` bump. An existing catalog gains them from a one-off command the user runs by hand with the app closed; nothing is shipped for it and nothing backfills on open (user, 2026-10-27; S11-44).
- **Promote suggests a record or nothing** (user, 2026-10-24; [detail](name-candidates.md#promote-suggests-a-record-or-nothing)). A proposal row carries a handle or no target, with its alternatives; Go no longer sends New or Skip, nor the row reasons `weak` and `no_match`, and Swift never drafts them. New and Skip are only the researcher's choice, still sent back as fixed pairs.
- **Gate only; the curve is held** (user, 2026-10-24) until dogfood after S11-43 shows a fuller record still missing.
- **A prefix is an initial generalised** (user, 2026-10-24; [prefix rule](name-candidates.md#prefix-rule-in-detail)). A word of a given, nickname or untyped part (the alias tables' roles) of at least `PrefixMinLength` letters that is the exact start of another, after folding, is `Contains` at `PrefixCredit` with `CausePrefix`, checked after alias, fold and spelling. No surname prefixes; text compare never gets it. Both knobs live in the registry's value section. `PrefixMinLength` is 2 (user, 2026-10-24). A prefix folds in the reconciler like an initial (user, 2026-10-24). The shortlist finds prefixes both ways through separate index and lookup keys (S11-46).

### Claim evidence

- **A separate, lazy query per row** ([claim evidence](claim-evidence.md#separate-from-matching)): inputs are the page's targets, not matching internals.
- **Lines** ([lines](claim-evidence.md#lines)): one per member per property, grouped by property, identical values collapsed; `Same` / `Equivalent` start checked; every line can be pinned, only the default check depends on the relation. No "select all".
- **Relationships** ([relationships](claim-evidence.md#evidence-through-relationships)): a corresponding segment is its own checkable line that pins the bundled plumbing (bridge ends, everything on the bridge, the neighbor's defining properties, e.g. `event_type`); the neighbor's other properties are checkable lines under it.
- **A related person's lines compare with every member** (user, 2026-10-27; S11-52). Under a relationship Card whose neighbor is a person, the neighbor's other properties are compared with all of that entity's accepted members, as the row's own property lines are: a spouse is one person recorded across many Sources, and the evidence that the Source's "Jake Robins" is that person lives in every Source that names him. An event neighbor keeps S11-19 D-13's scope, the members of the Source that records the link ("this birth with that birth only"). Which kinds widen is a registry entry in the evidence section, not a kind check in `bundles`. Supersedes S11-19 D-13 for person neighbors.

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
| S11-09 | Property compare | 3 | `propertycompare`: negatives, neutral terms, cardinality, outcome table | — | S11-05 | [plan](prs/S11-09-property-compare.md) | Landed #468 |
| S11-10 | Subject compare | 3 | `subjectcompare` + registry subject section; pairs; similarity and evidence | — | S11-09 | [plan](prs/S11-10-subject-compare.md) | Landed #471 |
| S11-11 | Subject profile and entity compare | 3 | `subjectprofile.Build`: strict profile, denied as negatives, `Excluding` | — | S11-10, S10-66 | [plan](prs/S11-11-subject-profile.md) | Landed #476 |
| S11-12 | Lifespan for people | 3 | Profile builder derives a lifespan; subject compare checks it | — | S11-11 | [plan](prs/S11-12-lifespan.md) | Landed #480 |
| S11-24 | Delete frequency weighting | 3 | `Stats.ValueFreq`, m / u priors, frequency scales and their benchmark go | — | S11-05 | [plan](prs/S11-24-delete-frequency.md) | Landed #483 |
| S11-13 | Migrate Promote scoring onto subject compare | 3 | Property scoring, seeding, duplicate rows; values loader moves; delete `core/match`, `core/database/matching` | — | S11-12, S11-24 | [plan](prs/S11-13-promote-subject-compare.md) | Landed #487 |
| S11-14 | Segment compare | 4 | `segmentcompare`, typed bridge keys, birth / death cardinality, one-hop credit | — | S11-13 | [plan](prs/S11-14-segment-compare.md) | Landed #493 |
| S11-15 | Shortlist index | 4 | Key → handles per kind on `graphcache`; registry shortlist keys | — | S11-13 | [plan](prs/S11-15-shortlist-index.md) | Landed #497 |
| S11-16 | Rewrite graph alignment | 4 | Layer 6 procedure, bands from evidence, accept bar; old walk and registry deleted | — | S11-14, S11-15 | [plan](prs/S11-16-graph-alignment.md) | Landed #501 |
| S11-17 | Contested flag through FFI | 4 | Proposal rows carry contested and move reasons to Swift | — | S11-16 | [plan](prs/S11-17-contested-ffi.md) | Landed #504 |
| S11-18 | Contested marker on Promote rows | 4 | The marker on the page | [S11-D1](design/archive/S11-D1-promote-rows.md) | S11-17, S11-D1 | [plan](prs/S11-18-contested-marker.md) | Landed #507 |
| S11-19 | Claim evidence query | 5 | Go query: member lines, collapse, relationship bundles; registry evidence section | — | S11-16 | [plan](prs/S11-19-claim-evidence-query.md) | Landed #513 |
| S11-25 | Claim evidence RPC and pin policy | 5 | Lazy per-row RPC; Go default pins at Apply; Swift drafting moves onto it; proposal field 9 and `attachExhibits` removed | — | S11-19 | [plan](prs/S11-25-claim-evidence-rpc.md) | Landed #517 |
| S11-32 | Lifespan and likeness evidence lines | 5 | The lifespan line (says when a side contradicts itself) and likeness lines (depiction compare by image) in `claimevidence` and the RPC; `ConclusionValue` gains a depiction case | — | S11-25, S10-54 (landed #433) | [plan](prs/S11-32-lifespan-likeness-lines.md) | Landed #523 |
| S11-20 | Claim evidence sheet | 5 | The sheet's new layout and copy, with likeness lines as two crops | [S11-D2](design/archive/S11-D2-evidence-sheet.md) | S11-25, S11-32, S11-D2 | [plan](prs/S11-20-claim-evidence-sheet.md) | Landed #525 |
| S11-33 | Evidence sheet audit follow-up | 5 | Items 1 – 10, A, B, C, F, G, H, I of the evidence sheet area audit: one sheet dismissal, engine-named relationships, one seeded-term wording table, whole-phrase copy, an evidence board for responsiveness, unread wire fields removed, docs | — | S11-20 | [plan](prs/S11-33-evidence-sheet-audit.md) | Landed #525 |
| S11-26 | Connections file by their ends | 6 | Drop the bridge switch-off: `SkipBridgeIDs`, `skip_bridge_ids`, the `skip` parameters | — | — | [plan](prs/S11-26-connections-by-ends.md) | Landed #527 |
| S11-27 | Connection subrows on Promote | 6 | Each connection as a subrow of a row it joins, its state from its ends; the panel and toggles go | S11-D1 | S11-26, S11-D1 | [plan](prs/S11-27-connection-subrows.md) | Landed #530 |
| S11-21 | Large-Source performance check | 7 | 100-Subject budgeted test; bound nominations, memoize candidates, warm the name format and the alias index | — | S11-16, S11-19 | [plan](prs/S11-21-large-source-performance.md) | Landed #536 |
| S11-31 | Bound the loaded Evidence graphs | 7 | Loaded Evidence graphs and their memos are evicted least recently used beyond a cap | — | S11-19, S11-21 | [plan](prs/S11-31-evidence-graph-cap.md) | Landed #541 |
| S11-28 | Fold a nickname into its given name | 2 | Reconciler folds a nickname or abbreviation into its formal name, keeping the set's head; a reviewed table directs two-way pairs; cache reads the product alias tables | — | S11-06 | [plan](prs/S11-28-alias-fold.md) | Landed #589 |
| S11-29 | Quote nicknames and add former surnames | 2 | Nicknames in quotes everywhere; `former_surname` part type in parentheses, italic on the Person titles | — | — | [plan](prs/S11-29-name-display.md) | Planned |
| S11-30 | Gender meets sex at birth in its own words | 3 | The sex / gender pool maps gender terms into sex at birth terms through a registry crosswalk; `man` agrees with `male` | — | S11-10 | — (bugfix, user 2026-10-21) | In review |
| S11-34 | Evidence button hover on Promote rows | 6 | The row's assessment button (band, Contested, reason, chevron) takes the hover the S11-D1 board draws, on the kit's `PVHoverEffect` | — (S11-D1, agreed) | S11-27 | — | Scoped |
| S11-35 | Hold a record Promote did not offer | 8 | `PromoteFlow` takes a record from outside the menu and holds it as a decided row; Go tests pin that a fixed off-shortlist handle is loaded, banded, has evidence and files; `FakeStore` holds it | — | — | [plan](prs/S11-35-hold-found-record.md) | Landed #543 |
| S11-37 | One kind-scoped catalog search | 8 | The omnibar's results model and hit row move out of `Features/Workspace` to a shared home, with a `kinds` input and an optional location; the omnibar uses them unchanged | — | — | [plan](prs/S11-37-kind-scoped-search.md) | Landed #551 |
| S11-36 | Find another record in the target menu | 8 | "Find another record…" at the end of the target menu opens a search of the row's kind on S11-37's model; the pick goes through S11-35 | [S11-D3](design/archive/S11-D3-promote-target-menu.md) | S11-35, S11-37, S11-D3 | [plan](prs/S11-36-find-another-record.md) | Landed #558 |
| S11-41 | One link reading for bridges | 9 | `promotealign`'s private `linkOf` and `eventEndpoint` move to `segmentcompare`, so the loaders and `core/kinship` read a bridge as a `Link` one way | — | — | [plan](prs/S11-41-link-reading.md) | Landed #561 |
| S11-38 | Kinship patterns | 9 | `core/kinship`: the three patterns and the derivation over one side's links, returning only implied links with the links and event behind them; `vocabkey` term constants; the matching skill | — | S11-41 | [plan](prs/S11-38-kinship-patterns.md) | Landed #564 |
| S11-49 | Kinship links in the graph cache | 9 | `graphcache` keeps one kinship result per pivot, derived on first read and dropped with the pivot or a node one hop out; `canonlinks.Reader.For` answers one person's implied links; verify mode checks them; the bound per pivot, linear grandparent dedupe and the `subject` marriage role (moved from S11-39) | — | S11-38 | [plan](prs/S11-49-kinship-graph-cache.md) | Landed #569 |
| S11-50 | Matching warms while the project opens | 9 | `WarmOnOpen` loads the shortlist's kinds and builds the name index behind the open spinner; drops re-warm in the same call; budgets gain an open budget and cold becomes the first proposal after open | — | — | [plan](prs/S11-50-warm-matching-at-open.md) | Landed #570 |
| S11-39 | Matching reads kinship shapes | 9 | The layer loader derives the Source side's implied links and `loadCanon` reads the canon's from the graph cache, both through one `linkSet` merge; graph alignment counts a paired shape once; goldens for marriage ↔ spouse, birth parents ↔ parent, chain ↔ grandparent; the large-source fixture gains a 12-generation tree | — | S11-49, S11-50 | [plan](prs/S11-39-matching-reads-kinship.md) | Landed #572 |
| S11-51 | One person's links for every reader | 9 | S11-39's `linkSet` merge moves from `promotealign` to `core/kinship` and takes in `loadCanon`'s implied-link merge; `canonlinks.Reader.Relationships` returns a person's recorded and implied links for S11-40 and Spike 13's Person page | — | S11-39 | [plan](prs/S11-51-person-links.md) | Landed #575 |
| S11-40 | Implied links in claim evidence | 9 | `canonLinks` reads the canon's implied links from the graph cache; a corresponding implied link gives a relationship line with how each side recorded it and the event behind it, a bundle of every bridge behind it, through the RPC to Swift | — | S11-39, S11-51 | [plan](prs/S11-40-implied-links-evidence.md) | Landed #579 |
| S11-42 | Implied links on the evidence sheet | 9 | The relationship Card says how each side recorded the fact ("This source: married in the marriage dated 12 May 1842 · The record: recorded as spouses") in a muted line under its header; L10n | — | S11-40 | [plan](prs/S11-42-implied-links-sheet.md) | Landed #585 |
| S11-52 | A related person's lines compare with every member | 5 | Under a relationship Card, a person neighbor's property lines compare with all of the entity's accepted members, not only the linking Source's; event neighbors unchanged; a registry entry names the kinds | — | S11-40 | — | Scoped |
| S11-43 | Lower the candidate gate | 10 | `alignment.MinSimilarity` 0.5 → 0.4 with its inventory row; a test that "Amy" gets "Amy Lee Anne Long" as an alternative | — | — | [plan](prs/S11-43-candidate-gate.md) | Landed #588 |
| S11-44 | Des as a nickname of Desiree and Desirae | 10 | `{desiree, des}` and `{desirae, des}` in `nicknameSets` with a `CacheVersion` bump. Existing catalogs get them, and S11-48's Properties, from a one-off command the user runs by hand, not shipped (user 2026-10-27) | — | S11-28 | [plan](prs/S11-44-des-nickname.md) | Landed #594 |
| S11-47 | Promote suggests a record or nothing | 10 | Unplaced proposal rows carry no target and no `weak` / `no_match` reason; possible duplicates found among unplaced rows; Swift drafts a handle or unset only | — | — | [plan](prs/S11-47-suggest-a-record-or-nothing.md) | Landed #598 |
| S11-46 | Prefix keys in the shortlist | 10 | `WordKeys` splits into index and lookup lists; given-name words are filed and looked up under their prefixes both ways; `PrefixMinLength`; extra keys measured | — | — | [plan](prs/S11-46-prefix-shortlist-keys.md) | Landed #602 |
| S11-45 | Prefix names | 10 | A given name that starts another is a `Contains` hit at `PrefixCredit` (`CausePrefix`), with the broader side right; the cause worded on the sheet; the reconciler's call | — | S11-46, S11-28 | [plan](prs/S11-45-prefix-names.md) | Landed in [#607](https://github.com/mendahu/provenencia/pull/607) |
| S11-48 | Religion and languages spoken on a person | — | Term Properties `religion` and `languages_spoken` on person, cardinality multiple, seeded with Catholic and English | — | — | [plan](prs/S11-48-religion-languages.md) | Landed #563 |
| S11-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| [S11-D1](design/archive/S11-D1-promote-rows.md) | Promote page — rows | **S11-18**, **S11-27** | Board agreed 2026-10-20 ([Claude Design](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share)). Contested marker beside the band; move reasons in the Updated caption; connections as subrows of the rows they join, answering the [design questions](promote-connection-subrows.md#open-questions-for-design) |
| [S11-D2](design/archive/S11-D2-evidence-sheet.md) | Promote page — evidence sheet | **S11-20** | Board agreed 2026-10-22 ([Claude Design](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share), frames E01 – E14). Whole sheet: "why suggested" marks at the title, property sections, relationship lines, collapsed members, likeness crops (from Spike 10's S10-D13), copy |
| [S11-D3](design/archive/S11-D3-promote-target-menu.md) | Promote page — target menu | **S11-36** | Board agreed 2026-10-25 ([Claude Design](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share), frames 20 – 26b). The "Find another record…" entry, the inline picker it opens on the row, how a found record reads in the menu once picked, and how a held found record reads, which today is band None as "No match" beside "Your choice" (it can also read Contested after Go answers, S11-35 T-2): S11-D3 draws what replaces "No match" for it. Product calls are [decided](promote-find-another-record.md#decided-user-2026-10-24) |
| [S11-D4](design/S11-D4-kit-search-parity.md) | Design system — ComboBox and omnibar | — | Written 2026-10-25. Brings the main kit's ComboBox (remote results, searching line, caller empty line, custom rows) and the omnibar field, panel and hit row up to the shipped Swift (#554, S11-37). Drawn before S11-D3, which instances them |

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
  S11-32  Lifespan and likeness evidence lines
  ✎ S11-D2 ──▶ S11-20  Claim evidence sheet
  Check: five named members show as matches with identical values collapsed;
         "They have a matching birth event" pins its bundle; pins land on every member; two people with a likeness show both crops.
  S11-52  A related person's lines compare with every member   (follow-up, after S11-40)
  Check (S11-52): a spouse recorded as "Jake Robins" or "Jake" in three Sources and
         "James Kenneth Robins" in three shows the Jake lines as matches under
         "they share a spouse", beside the James Kenneth Robins line.

SLICE 6 — Connection subrows
  S11-26  Connections file by their ends   (any time)
  ✎ S11-D1 ──▶ S11-27  Connection subrows on Promote
  S11-34  Evidence button hover on Promote rows   (follow-up, after S11-27)
  Check: setting John to Skip dims "Mary · parent of · John" under its row; nothing
         toggles by itself; the bottom connections panel is gone; the evidence
         button answers the pointer as the board draws it.

SLICE 7 — Performance
  S11-21  Large-Source performance check
  S11-31  Bound the loaded Evidence graphs
  Check: a 100-Subject Source proposes within the benchmark, and a warm evidence
         open issues no catalog query; a Source whose events are all located at one place with many filed events leaves at most N Sources loaded.

SLICE 8 — Find another record
  S11-35  Hold a record Promote did not offer   (any time)
  S11-37  One kind-scoped catalog search        (any time, parallel with S11-35)
  ✎ S11-D3 ──▶ S11-36  Find another record in the target menu
  Check: on a row Promote opens New, "Find another record…" finds a differently
         named record of the row's kind; the row holds it after other rows are
         decided, shows its band and Contested, opens its evidence, and Done files it.

SLICE 9 — Kinship shapes
  S11-41  One link reading for bridges          (any time)
  S11-38  Kinship patterns
  S11-49  Kinship links in the graph cache
  S11-50  Matching warms while the project opens   (any time)
  S11-39  Matching reads kinship shapes
  S11-51  One person's links for every reader   (right after S11-39)
  S11-40  Implied links in claim evidence
  S11-42  Implied links on the evidence sheet
  Check: a Source marriage proposes its couple strong on two people the tree has as
         spouses; a birth naming the father matches the tree's parent link; a parent
         chain matches a grandparent link; the sheet says "married in the marriage
         on …" and checking it pins both participations. A project with a
         12-generation tree opens, then proposes within the large-source budgets.

SLICE 10 — Name candidates
  S11-43  Lower the candidate gate              (any time)
  S11-44  Des as a nickname of Desiree and Desirae   (after S11-28)
  S11-47  Promote suggests a record or nothing  (any time)
  S11-46  Prefix keys in the shortlist          (any time)
  S11-45  Prefix names
  Check: "Amy" is offered "Amy Lee Anne Long"; "Desiree" and "Desirae" are offered
         "Des" in a new catalog (and in the dogfood catalog after the one-off command); a three-letter start of a given name
         with no nickname set is offered through the prefix rule, and the sheet says why.

PERSON VOCABULARY (one-off, user 2026-10-25)
  S11-48  Religion and languages spoken on a person   (any time)
  Check: a new project's person takes Religion → Catholic and a custom Anglican, and
         Languages spoken → English; the Person page keeps both religions.

S11-99 — Dogfood close / docs
```

Parallel: S11-01 with S11-22; S11-03 with S11-04; S11-24 with slice 3 after S11-05; S11-06, S11-07 and S11-09 after S11-05; S11-14 with S11-15; slice 5 with slice 7 after S11-16; slice 6 any time (S11-26 has no dependency); S11-34 any time after S11-27; S11-43, S11-46 and S11-47 any time, S11-44 and S11-45 after S11-28; S11-41 and S11-50 any time, and slice 9 runs beside slice 8; S11-35 and S11-37 any time, in parallel, and S11-D3 can be drawn while they run. S11-D1 and S11-D2 can be drawn while slice 4's engine PRs run.

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
                            ├──▶ S11-19 ──▶ S11-25 ──▶ S11-32 ──▶ S11-20
                            └──▶ S11-21
S11-19, S11-21 ──▶ S11-31
S11-D1 ──▶ S11-18, S11-27
S11-26 ──▶ S11-27
S10-54 ──▶ S11-32   (Spike 10)
S11-D2 ──▶ S11-20
S11-27 ──▶ S11-34
S11-35 ──┐
S11-37 ──┤
S11-D3 ──┴──▶ S11-36
S11-41 ──▶ S11-38 ──▶ S11-49 ──┐
S11-50 ─────────────────────┴──▶ S11-39 ──▶ S11-51 ──▶ S11-40 ──▶ S11-42
                                                        S11-40 ──▶ S11-52
                                            S11-51 ──▶ S13-04   (Spike 13)
S11-28 ──▶ S11-44
S11-46, S11-28 ──▶ S11-45
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
Ports `subjectvocab`'s `TestDefaultProfileCardinality` as a seed check of the subject section (every kind a seeded type, every key bound to it, pooled and period properties of the right value type); the registry writes the `vocabkey` constants. Extends the matching skill. Also moves `core/valuecompare`'s private date builders (`optInt`, `ymd`, `about`, `phrase`, `between`, `dp`) onto `core/datevalue/datevaluetest` (S11-09 review finding MNT-3).
**Context.** [Layer 3](comparison-layers.md#layer-3-subject-comparison); rule "Sparse data first".

#### S11-11 — Subject profile and entity compare

**Ships.** `core/subjectprofile.Build` from reconciler output: displayed values positive (tied leaders and multi-valued properties included), denied values negative, others left out; `Options.Excluding`; `ForSubject` for a plain Subject's side. Entity compare is `Build` then `subjectcompare.Compare`.
**Not this PR.** Lifespan (S11-12), callers (S11-13).
**Check.** A Subject recording an outvoted value conflicts; one recording a denied value conflicts via negative; leave-one-out drops that member's support.
**Context.** [Entities](comparison-layers.md#entities-compare-against-the-conclusion).

#### S11-12 — Lifespan for people

**Ships.** The profile builder derives a person's lifespan from every dated event they take part in: birth starts it, death or burial ends it, `MaxYears` pads an open end; subject compare checks each side's span against the other's window, with a large conflict penalty and near-zero agree weight (user, 2026-10-18).
**Not this PR.** Events and places. Age rules and two-hop rules. Loading life events (S11-13).
**Check.** An entity who died in 1840 against a Subject born in 1850 conflicts; born 1850 / died 1990 conflicts; died 1900 / resided 1950 conflicts; born 1850 / died 1920 is consistent.
Also builds `core/subjectcompare`'s `invariants_test.go` lead groups (`kindGroups`) from `matchregistry.Default().Subject` (its `Weights` leads, `Pools` and `Periods`), keeping only each lead's test value by hand, so the lifespan lead cannot be silently skipped by AC-3.4 / AC-3.5 (S11-10 review finding MNT-2). And folds `subjectcompare`'s `period` start and end selection loops into one helper, keeping the earliest-start and latest-end tie rules (S11-10 review finding MNT-3).
**Context.** [Lifespan](comparison-layers.md#lifespan-a-derived-property-for-people).

#### S11-13 — Migrate Promote scoring onto subject compare

**Ships.** `graphalign` scores candidates' properties with entity compare; `promotealign` seeds with it instead of `match.Rank`; duplicate-row checks (`alikeNew`, `pairNode`) use subject compare. `matching.SubjectsValues` moves to the loader that feeds `propertycompare` and `subjectprofile`. Delete `core/match` and `core/database/matching`.
**Not this PR.** Segments and the walk (slice 4): the old walk stays until S11-16, calling the new scoring.
**Check.** Golden and sparse cases pass; a birth is never proposed on a death; the packages are gone and `go build ./...` passes.
**Lifespan on the page.** Moved to S11-20 (user, 2026-10-18): this PR keeps the derived lifespan off the sheet and the row reason.
**Context.** Rule "`core/database/matching` is deleted".

#### S11-24 — Delete frequency weighting

**Ships.** Remove `Stats.ValueFreq`, the m / u priors (`MPrior`, `UPrior`), the frequency scales and their lift, and `BenchmarkLoadStatsNames`; event type and toponym get plain weights.
**Not this PR.** Fan-out (S11-14).
**Check.** Golden and sparse cases pass with plain weights; no frequency symbol remains (grep).
**Context.** Rule "Frequency weighting is gone"; [frequency](comparison-layers.md#frequency).

### Slice 4 — Segments and graph alignment

#### S11-14 — Segment compare

**Ships.** `core/segmentcompare`: neighbor by entity compare, the bridge as a typed `BridgeKey` (role or relationship term, direction and inverse terms, event type); registry segment section (subject of birth / death is one-cardinality from both ends, user 2026-10-19; neighbor credit one hop, link support). The layer loader reads link kinds from `connectrules` instead of a list in code. Output carries value refs for the evidence bundle. Fan-out goes: `Stats.FanOut`, `promotealign/stats.go`'s `addFanOut` / `sqlFanOut`, `FanOutUnknown`. Extends the matching skill.
**Not this PR.** The walk (S11-16). Comparing a bridge's own properties (place relationship dates) through subject compare (S11-16; user, 2026-10-19).
**Check.** A spouse never corresponds to a sibling; parent ↔ child corresponds with ends swapped; differing births subtract; differing residences are unknown.
**Context.** [Segments](comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6).

#### S11-15 — Shortlist index

**Ships.** A key → handles map per kind (`core/shortlist`), built with value compare's keys from the registry's shortlist keys (person name words, not initials; event type + year bucket; toponym words). It reuses `graphcache`'s stats slot (`Stats` / `SetStats` / `statsOK`, renamed), dropped on a handle commit like today. Promote's seeding looks handles up in it instead of scanning the kind. When the shortlist replaces the kind scan in `graphalign.Candidates`, the Promote profiler (`core/database/promotealign/profiles.go`, `profiler.profile`) loads and builds life events only for shortlisted handles (and the fixed and BFS-reached ones), not for every handle of the kind. S11-13 profiles every person and their events per repropose (PRF-1 in its ship log: warm repropose 40 to 160 ms on the 20k-person fixture).
**Not this PR.** Using it in the walk's fallback and queue (S11-16). A coarse misspelling key (open; [decided 2026-10-17](comparison-layers.md#decided-on-2026-10-17)).
**Check.** "J. Robins" shortlists every handle with `robins`; aliases share a key; no `Same` / `Equivalent` pair is missed (property test).
**Context.** [Layer 6 procedure](comparison-layers.md#layer-6-procedure) step 1; [vocabulary](comparison-layers.md#vocabulary-finding-candidates).

#### S11-16 — Rewrite graph alignment

**Ships.** `graphalign.Align` follows the layer 6 procedure: shortlist and nominate, top K with ties, one queue, segments around accepted pairs, dirty-pair settling with reasons, bands from evidence, the accept bar. Contested is a `RowFlags` field and move reasons are `Reason` codes; the plan decides whether `ConflictWithFixed` is subsumed. Grows the registry alignment section S11-15 opened, and moves its type into `graphalign`. Delete the old walk, fallback, `refineOneHop` and `core/graphalign/registry.go`. [`promote-graph-alignment.md`](../../promote-graph-alignment.md) and the matching skill are updated.
Segment compare learns unplaced neighbors (entity compare decides). A bridge's own properties (place relationship dates) are not compared: no golden needs them (planning, 2026-10-19). Candidates gate on similarity (user, 2026-10-19).
**Not this PR.** FFI and UI for contested (S11-17, S11-18). Memoization (S11-21).
**Check.** The invariants hold (order, decided rows, re-run, start-independence); name + birth date proposes strong; a decision moves only its neighborhood.
**Context.** [Layer 6 procedure](comparison-layers.md#layer-6-procedure), [bands](comparison-layers.md#bands).

#### S11-17 — Contested flag through FFI

**Ships.** The contested flag and move reasons reach Swift through `engine.proto` and the Promote handler. Promote's reason codes move from `PromoteFlow.swift` into `Platform/ConclusionVocabulary.swift`, under `api/ffi/swift_vocabulary_test.go`.
**Not this PR.** The marker and the move-reason copy (S11-18). Cause codes (S11-25).
**Check.** A handler test returns contested on a row with a conflicting comparison.
**Context.** [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md); [bands](comparison-layers.md#bands).

#### S11-18 — Contested marker on Promote rows

**Ships.** The contested marker beside the band, as S11-D1 draws it, with L10n copy. It words the three move reasons, naming the neighbor (`PromoteFlow.Row.move`, S11-17). A move is settled within one proposal and `Row.updated` is a change between proposals, so S11-D1 decides where a move shows: in the "Updated · <reason>" line, on its own, or both.
**Not this PR.** The evidence sheet (S11-20).
**Check.** A row with a conflict shows contested at any band; a clean row does not; a row that moved says which neighbor moved it.
**Context.** S11-D1; [`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md).

### Slice 5 — Claim evidence

#### S11-19 — Claim evidence query

**Ships.** A Go query from a row's target and its neighbors' targets: property lines per member through property compare, identical values collapsed, relationship lines through segment compare with bundles, the registry evidence section (pin defaults, defining properties), wording keys from cause codes. A line's points come from `subjectcompare.Points` given the property's `propertycompare.Result`: reshape `Points` (or add one helper beside it) so no caller reads the speaking similarity itself, and drop `Compare`'s inline partial arithmetic and `promotealign`'s private `similarity` (S11-13 MNT-3). The Source side's Observations come from the Source graph through `promotealign`'s `decode`, the one row decoder, not a second SQL read like `loadExhibitObservations` (S11-13 MNT-4).
**Not this PR.** The RPC, Apply defaults and Swift (S11-25). If the plan finds the pin tables can't hold one incoming Observation against many member Observations, a migration PR is cut ahead of it.
**Check.** Five named members give one collapsed name line of five; a matching birth gives one relationship line whose bundle holds every Observation behind it.
**Context.** [`claim-evidence.md`](claim-evidence.md); [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §9.

#### S11-25 — Claim evidence RPC and pin policy

**Ships.** A lazy per-row RPC over S11-19's query. Apply computes default pins in Go for rows whose sheet was never opened. Each evidence line carries its value or segment cause code, parsed in `ConclusionVocabulary.swift` under `swift_vocabulary_test.go` (moved from S11-17). Swift's pin drafting (`PromoteFlow.swift` `draftedPins`, `draftedArgument`, `kept.pins`) moves onto the RPC. `PromoteGraphAlignmentRow.comparisons` (field 9), `attachExhibits`, `pairExhibits`, `withRowLines` and `refLine` are deleted. Pins write one incoming Observation against many member Observations; Done backfills every member's claim. Deleting `pairExhibits` also retires its compare-before-`Compatible` loop (S11-13 PRF-2) and the pin-test helper `obs` in `exhibits_pin_test.go` that copies how `exhibitObs` is built (S11-13 MNT-5); any pin test that survives builds its values through the production constructor.
Apply runs S11-19's query once per row and reuses the per-Source layer memo and the profile memo (S11-19 §14), and a test shows that Apply over a page builds the layer once.
**Not this PR.** The new sheet layout (S11-20); the shipped sheet reads the new lines unchanged in shape.
**Check.** Done on a row never opened pins its matching lines; opening the sheet shows the same defaults; every member's claim is backfilled.
**Context.** Rule "The pin policy has one owner, in Go"; [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md).

#### S11-32 — Lifespan and likeness evidence lines

Split from S11-20 when it was planned (user, 2026-10-22): the engine half of the lines S11-19's query leaves out, so S11-20 is the sheet only.
**Ships.** No new chrome; the sheet S11-25 left draws the new lines in its current shape. The contract below is what S11-20's plan builds on; this PR's plan may name types differently but must deliver each item.
- **Lifespan line.** People only. `claimevidence.Evidence` gains one lifespan result from the row against the record's profile (the `lifespan` lead of `subjectcompare.Compare`, which `Build` already runs): its outcome, relation and value cause, and each side's window and span as date values. `subjectprofile` reports why a side has no lifespan: no dated events, or its own dated events contradict each other (S11-12 D-12); the result says which side contradicts itself. The RPC carries it as one message on the response (not a section line: it has no member and no pin), parsed into a Swift `CatalogClaimEvidence` field. It gets no offer and is never pinned; each side's span (its earliest and latest dated event) is sent as date values, and a side with none says whether it had no dated events or contradicts itself. The sheet shows spans, not windows (S11-D2 §5).
- **Likeness lines.** `likeness` values reach the evidence: `decode` and the profile values carry the depiction's image File id (pure `core/depictionvalue`; layers never import `core/database/depictionvalues`). A depiction comparer says `Same` when both values name one image File (the reconciler's rule in `autoreconcile/modules.go`), and otherwise that the layer cannot judge, with a new cause code. In `claimevidence`, a likeness line with the same image is agree, pinnable and checked; with two images it is shown, pinnable and unchecked (user, 2026-10-22: the researcher judges by eye). Likeness earns no points and does not change matching.
- **Section order.** Sections follow the catalog's binding order for the kind (`subject_type_properties.sort_order`, as the conclusion page orders fields), cached on the vocabulary; the seeded person binding puts `likeness` second, after `name` (user, 2026-10-22, from the S11-32 plan review). The sheet renders wire order (S11-20 review M-2).
- **Binding writes.** Assign, remove and reorder of a type's bindings run through `writes.Run`, audited as `subject_type_property`, so a binding edit drops the cached vocabulary; a Property delete releases its bindings through `deleteimpact`; a new `ReorderSubjectTypeProperties` RPC sets a type's order. No gesture calls it yet (`docs/ideas/binding-order-editing.md`). Added when the ship loop found binding writes unaudited (user, 2026-10-22).
- **Wire.** `ConclusionValue` gains `DepictionValue depiction = 6` (the S10-54 message: `region_json`, `image_rel_path`, the path resolved as `observations.go` `depictionPaths` does); Swift `CatalogConclusionValue` gains its case and `mapConclusionValue` maps it; the new cause code joins `ConclusionVocabulary.swift` and `TestSwiftVocabularyMatchesGo`. Likeness offers come from `PinOffers` with no change to Apply; the lifespan has none.
**Not this PR.** The sheet's new layout and copy (S11-20).
**Check.** A person whose own death precedes their birth gives a lifespan result naming that side as contradicting itself; two members with likenesses, one the same image as this Source's, give two likeness lines, one checked and one unchecked; Done on an unopened row pins the same-image likeness only.
**Context.** [`comparison-layers.md`](comparison-layers.md#lifespan-a-derived-property-for-people); [`add-depiction-value`](../../../.agents/skills/add-depiction-value/SKILL.md), [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md); S11-12 D-12.

#### S11-20 — Claim evidence sheet

**Ships.** The Promote sheet as S11-D2 draws it (board agreed 2026-10-22) and as S11-20's ship rounds 16 – 20 left it: the title "<row> joins <record>" with the band and Contested marks after it (the reason is their tooltip), a subtitle that states the claim, the claim fields, "<row> from <Source> belongs in the concluded <kind> <record> because…", property Cards of lines on a grid (box, value, sentence, Source, outcome) whose several records open from "Show N records", relationship Cards with the neighbor's lines under them, and likeness lines as two crops. Every line, record and relationship can be pinned; only the default check depends on the relation. All copy goes through L10n. No engine change: the lines come from S11-25 and S11-32 (split, user 2026-10-22). The sheet's audit follow-up is S11-33.
**Naming and band (from S11-18 review M-2, M-6).** The sheet's subtitle speaks the band through `PromoteModel.assessmentText`, so a contested row reads "Strong, contested"; neighbor groups name a row by `PromoteFlow.Row.shownRef`, as the row reasons do.
**Lifespan (withdrawn).** The sheet carries no lifespan line: S11-20 ship round 11 removed it from the wire and the app at the user's request (S11-32 had carried it). The lifespan stays a matching signal.
**Empty evidence (from S11-25 review UX-3).** When a row's evidence has no lines and no relationships, say there is nothing to compare instead of drawing the column headers alone.
**Not this PR.** Engine and wire (S11-25, S11-32). Opening the sheet from a connection subrow (not planned, decided here for S11-D1's note).
**Check.** The sheet reads as sentences; no "Through <name>"; checking a relationship line pins it; two people with a likeness show both images.
**Context.** [S11-D2](design/archive/S11-D2-evidence-sheet.md); [wording](claim-evidence.md#wording); [plan](prs/S11-20-claim-evidence-sheet.md).

### Slice 5 follow-up — A related person's lines

#### S11-52 — A related person's lines compare with every member

Found in dogfood on 2026-10-27, reviewing S11-40 (user): under "they share a spouse", the spouse's name section showed one line, the marriage certificate's "James Kenneth Robins", as a conflict with the Source's "Jake Robins", while three of the entity's six Sources say "Jake Robins" or "Jake". The Card itself agrees, because the relationship compares with the entity's merged values; the lines under it do not.
**Ships.** Working rule "A related person's lines compare with every member". `bundles` in `core/claimevidence/build.go` passes `Sections` all of a person neighbor's accepted members (`Neighbor.Members`), not only the ones a bundle names; an event neighbor keeps today's scope. The evidence section of `core/matchregistry` gains the entry that says which neighbor kinds compare with every member (person yes, event no), read by `bundles`. The evidence query loads values for those members: S11-19 PRF-1's `bridgedOnly` trim stays for event neighbors and is skipped for the kinds the entry widens. Pins and default checks follow the existing line rules (matches checked, the rest unchecked); bundles and their pin ids do not change. `claim-evidence.md` § Evidence through relationships says it. No new chrome: the Card already draws several records per section.
**Not this PR.** Place neighbors (an event's place link) keep today's scope unless the plan finds a reason and the user agrees. Whether "Jake" and "James" compare as the same given name ([name candidates](name-candidates.md), S11-28's nickname sets).
**Check.** The Jake Robins case above shows the "Jake Robins" and "Jake" members as matching lines, checked, and the three "James Kenneth Robins" members as one collapsed line, unchecked, as the row's own Name section would; a birth neighbor's date lines still show only the linking Source's birth; `TestLargeSourceBudgets` stays within budget.
**Context.** [`claim-evidence.md`](claim-evidence.md#evidence-through-relationships); S11-19 plan D-13 and its ship log PRF-1; `core/claimevidence/build.go` `bundles`, `core/database/promotealign/evidence.go`; [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md).

### Slice 3 follow-up — Sex and gender

#### S11-30 — Gender meets sex at birth in its own words

**Ships.** Found in dogfood on 2026-10-21: the evidence sheet showed a Source's gender *Man* against a filed sex at birth *Male* as a conflict. S11-10's pool joined gender's values to sex at birth's but compared their term keys as they are, and the two vocabularies differ. `subjectcompare.Pool` gains `Terms`, a per-pooled-property crosswalk the registry fills (`man` → `male`, `woman` → `female`); `Side` maps a pooled term before comparing and leaves out one with no entry (`non_binary`). Tests use the seeded gender keys.
**Not this PR.** How the old evidence sheet labels a mapped value (it now reads the sex at birth label); the sheet is rebuilt in S11-20.
**Check.** A Source person recorded as gender Man against a filed person with sex at birth Male agrees.

### Slice 6 — Connection subrows

#### S11-26 — Connections file by their ends

**Ships.** Remove the switch-off: `promote.Batch.SkipBridgeIDs`, `ApplyPromoteGraphAlignmentRequest.skip_bridge_ids`, and the `skip` parameter of `FileSourceBridgesTx` / `FileBridgesTx`. Swift's `bridgeOff` and `toggleBridge` go with the request field. [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §3 and §9.1 and [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5.4 lose the "switched off" rule.
**Not this PR.** Subrows (S11-27).
**Check.** Done files every connection whose ends are both handles; no request field or parameter can skip one.
**Context.** [Connection subrows](promote-connection-subrows.md) "The problem" §2.

#### S11-27 — Connection subrows on Promote

**Ships.** Each connection shows as a subrow under a row it joins, as S11-D1 draws it, with its state from `ConnectionState` (files, waits, self-link). `PromoteFlow.connectionsByRow` (each connection under every end that is a row, built once per render, no pointers); the bottom connections panel goes; the footer keeps its count, with an unselected part (board frame 19). A subrow words its connection with the existing connection sentence (user, 2026-10-23). VoiceOver reads why a connection waits.
**Not this PR.** Opening the evidence sheet at a relationship line (S11-20 decides).
**Check.** Setting John to Skip dims his connections; setting him to New turns them on; a self-link warns.
**Context.** [Connection subrows](promote-connection-subrows.md); S11-D1.

#### S11-34 — Evidence button hover on Promote rows

Follow-up to S11-27 (user, 2026-10-24): the S11-D1 board draws a hover on the row's evidence button, and neither S11-18 nor S11-27 built it.
**Ships.** The assessment-column button in `PromoteAlignmentRow` (`PromoteView.swift`, identifier `promote.row.<ref>.evidence`: band and Contested Badges, the reason, the chevron) shows the board's hover state, reusing a hover style the kit already has (`PVMenuRowButtonStyle` in `PVButton.swift`, `PVListRowStyle` in `PVList.swift`; feature precedent `SourcesSplitZoneButtonStyle`) or extending one through `add-ui-component`, not a third hover fill. `PVHoverEffect` adds a press scale; the plan checks the board wants it. Its hit target stays the whole column.
**Not this PR.** Hover on subrows (S11-D1 §6 non-goals: no hover action on a subrow), on the target menu, or on the anchor rows. A new kit primitive (S11-D1 non-goals); if the hover is a reusable shape, the plan says whether it belongs on an existing kit style.
**Check.** Pointing at the evidence button shows the board's hover; moving off clears it; clicking still opens the sheet; VoiceOver reads as before.
**Context.** [S11-D1](design/archive/S11-D1-promote-rows.md) and its board, [Claude Design](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share) (user, 2026-10-24). The planner may not be able to open it; the implementing agent runs locally and reads the hover from the board, so the plan names the button and the board and leaves the hover's look to the board, not to prose; `macos/App/DesignSystem/Components/HoverEffect/PVHoverEffect.swift`; [`add-ui-component`](../../../.agents/skills/add-ui-component/SKILL.md) if the kit changes.

### Slice 7 — Performance

#### S11-21 — Large-Source performance check

**Ships.** The D-1 budgets in `TestLargeSourceBudgets` on the large fixture (a 100-Subject Source, a Source of 30 births all located at one place with 6,000 filed events, Zipf names, BEF / AFT dates); bounded nominations (a placed pair offers each layer neighbor its most similar `TopK` plus ties, capped at `TieCap`, at any similarity); a pair memo per `Align`; a candidates memo across proposals; the name format warmed at open by `nameformats`; the alias index built once at open by `namealiases`.
**Not this PR.** In-place update of the shortlist index after a handle commit (680 ms against a 1.5 s budget); scoring shortlisted handles on key values first (cold 1.3 s against 2.5 s); a cross-run neighbor memo (declined by the user); a commit effect for name-format writes (no writer after onboarding).
**Check.** `TestLargeSourceBudgets` passes in `scripts/check-pr.sh`; a warm Evidence pass issues no catalog query.
**Context.** [Layer 6 procedure](comparison-layers.md#layer-6-procedure) step 6.

#### S11-31 — Bound the loaded Evidence graphs

**Ships.** `graphcache` holds at most N loaded Evidence graphs, and evicts the least recently used one together with its `EvidenceGraphMemo` entries (S11-19 D-21). Today `g.evidenceGraphs` has no eviction, and only a commit or a reset clears it, and the claim evidence query loads every member's Source (S11-19 §14). N lives in one named constant, sized from S11-21's fixture. `Verify` checks only graphs that are still loaded. A test shows that an evicted graph reloads the same, and that its memo is rebuilt.
**Not this PR.** Evicting nodes, or the match index.
**Check.** Opening the evidence sheet on every row of a page whose records hold members from many Sources leaves at most N Evidence graphs loaded, and the sheet shows the same evidence as before.
**Context.** [`graphcache`](../../../core/database/graphcache/graph.go); S11-19 §14.

### Slice 2 follow-ups — Names on the Person page

Asked by the user on 2026-10-19, after S11-06. Planned as two PRs, with no migration for existing catalogs.

#### S11-28 — Fold a nickname into its given name

**Ships.** The name module folds a given-name or nickname unit into another when every word pair is the same, an initial, or a nickname or abbreviation of the other side's formal name (*Chris* → *Christopher*, *Wm* → *William*, *Ebbie* → *Abel*). Two given names the list pairs fold, keeping the set's head; for the 143 pairs listed both ways a reviewed table says which to keep. The upstream nickname seed is not edited. The reconciler cache reads the product alias tables; `CacheVersion` 18.
**Not this PR.** Storing the formal name in the catalog; reading a catalog's own alias rows; display changes (S11-29).
**Check.** A Person recorded as *Chris Robins* and *Christopher Robins* shows *Christopher Robins*.
**Context.** [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) §7.2; rule "The reconciler keeps its own decisions".

#### S11-29 — Quote nicknames and add former surnames

**Ships.** Nickname parts display in curly quotes. A `former_surname` product part type (role `family`, after the nickname in both Western orders) displays in parentheses, with the surname italic on the Persons list row and the Person page title.
**Not this PR.** Migrating existing catalogs; italics on other surfaces.
**Check.** *James “Jake” Robins*; *Jones, Mary (Smith)* with *Smith* italic.
**Context.** [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §4; [`ideas/international-names.md`](../../ideas/international-names.md).

### Slice 8 — Find another record

Pulled in from the ideas folder on 2026-10-24. Design: [`promote-find-another-record.md`](promote-find-another-record.md).

#### S11-35 — Hold a record Promote did not offer

**Ships.** A `PromoteFlow` call that takes a record from outside the row's menu (id, ref, sorted and reading titles, subtitle), adds it to `row.menu` with reason `decided`, sets the target and marks the row decided, with `setTarget`'s duplicate-note and re-propose follow-up; `PromoteModel` exposes it. A decided row keeps the record through re-proposals with its band refreshed (`makeRow`'s decided branch; [why the reason is `decided`](promote-find-another-record.md#where-the-swift-side-changes)). Go tests show that a fixed handle no shortlist key reaches is loaded into the canon, banded, flagged contested when a property conflicts, and answers `GetPromoteClaimEvidence` (extending `promote_align_test.go` and `promotealign_test.go`), and that it files at Apply (handler or `promote` tests). `FakeStore`'s unscripted propose already holds a fixed handle of any filed record (it has no candidates); a model test pins it. No chrome.
**Not this PR.** The menu entry and the search (S11-36). Any new proposal field or reason (working rule "A found record is a decided row, nothing new").
**Check.** `PromoteFlowTests`: a row opened New, given an outside record, holds it, shows it in its menu, and keeps it after another row is decided; the Go tests above pass.
**Context.** Working rules under [Find another record](#find-another-record); [shape](promote-find-another-record.md#shape) and [what the engine already does](promote-find-another-record.md#what-the-engine-already-does); [`add-swift-test`](../../../.agents/skills/add-swift-test/SKILL.md).

#### S11-36 — Find another record in the target menu

**Ships.** "Find another record…" at the end of every non-anchor row's target menu, as S11-D3 draws it: a sentinel option the row's binding catches (working rule "The menu entry is a sentinel option"), opening an inline picker on the row that searches every record of the row's kind, ranked by match, on S11-37's shared model and hit row (working rule "An inline picker, every record of the kind, no warning"). A pick goes through S11-35's call. How a held found record reads is S11-D3's: band None reads "No match" beside "Your choice" today, and it can be Contested after Go answers (S11-35 T-2); this PR ships what S11-D3 draws in place of "No match". All copy through L10n; VoiceOver reads the entry and the hits.
**Not this PR.** Searching from the evidence sheet; records of another kind; merging records ([what stays later](promote-find-another-record.md#what-stays-later)). A warning for a `Different` pick (refused, user 2026-10-24).
**Check.** On a row Promote opens New, the entry finds a differently named record; picking it selects it in the menu, the band and Contested show, the sheet opens on it, and Done files it. Opening and cancelling the search leaves the row undecided.
The picker keeps a picked value, so it is a `PVComboBox` with remote results: `onQueryChange` drives `CatalogSearchModel(kinds:)`, `isLoading` shows the combo box's searching line, and a failed search reads in its `empty` line. It does not build a third popup list. The picker's Escape and outside-click handling reuses an existing dismissal (`PVComboBox`, or a shared popup presenter if one exists by then), not another hand-rolled monitor ([`ideas/shared-list-helpers.md`](../../ideas/shared-list-helpers.md); S11-37 review M-6).
Each picker hit gets the same VoiceOver label as the omnibar's: `CatalogSearchHitPresentation.accessibilityLabel(for:)`. The plan either applies it at the picker's call site, as `OmnibarResultsPanel` does, or moves it onto `CatalogSearchHitRow` for both callers. Without it, `PVOmnibarHitRow`'s combined element reads the hit in layout order (S11-37 ship log UX-1). The plan also decides whether `CatalogSearchHitRow` keeps its `selected` parameter. The omnibar always passes `false`, because `PVMenuRowButtonStyle` draws selection (S11-37 ship log MNT-3).
**Context.** S11-D3; [decided](promote-find-another-record.md#decided-user-2026-10-24); S11-37; [`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md).

#### S11-37 — One kind-scoped catalog search

Preparatory refactor, from the map review (K-1, 2026-10-24).
**Ships.** `OmnibarResultsModel`'s search state (debounce, generation guard, errors, empty state) and the private `OmnibarHitRowView` move out of `Features/Workspace` to a shared home, with a `kinds` input and an optional location (Promote has none to give `SearchCatalog`). The omnibar keeps its behavior on them. [`add-searchable-kind`](../../../.agents/skills/add-searchable-kind/SKILL.md) says the omnibar is no longer the only consumer and points at the shared search.
**Not this PR.** Promote (S11-36). Any change to ranking or hit kinds.
**Check.** The omnibar works as before (its tests pass); a test drives the shared model with `kinds: ["person"]` and gets only person hits.
**Context.** `macos/App/Features/Workspace/OmnibarResultsModel.swift`, `OmnibarResultsHost.swift`, `OmnibarHitPresentation.swift`; [`add-ui-component`](../../../.agents/skills/add-ui-component/SKILL.md) for where the row view lands.

### Slice 9 — Kinship shapes

Pulled in from the ideas folder on 2026-10-24, with grandparent added by the user. Design: [`kinship-shapes.md`](kinship-shapes.md).

#### S11-41 — One link reading for bridges

Preparatory refactor, from the slice 9 map review (K-4, 2026-10-24).
**Ships.** `linkOf` and `eventEndpoint` (`core/database/promotealign/terms.go`), which read a filed or layer bridge as a `segmentcompare.Link`, move to `core/segmentcompare` beside `Key`. The layer loader, `loadCanon`'s `canonKeys` and the evidence query call the moved functions. No behavior change.
**Not this PR.** `core/kinship` (S11-38).
**Check.** `go test ./core/...` passes; `promotealign` has no private link reading left (grep).
**Context.** [`segmentcompare/key.go`](../../../core/segmentcompare/key.go); working rule "One wire vocabulary" is untouched.

#### S11-38 — Kinship patterns

**Ships.** `core/kinship`: the three patterns ([shape](kinship-shapes.md#shape)) and one derivation that takes a side's links (ends, `vocabkey.BridgeKey`, bridge Subjects) and `segmentcompare.Terms`, and returns only the implied links, each with the links and the event it was made from, keyed through S11-41's link reading and `segmentcompare.Key`. `core/vocabkey` gains constants for the terms the patterns name (all already seeded), checked by `TestVocabKeysMatchSeed`. The package joins the one purity guard. [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) lists `core/kinship` and gains "add a kinship pattern". Unit tests: each pattern, direction through `Key` (parent reads as `child` with ends swapped), grandparent from direct, implied and mixed parent links, no parent from a grandparent, nothing from catalog-only terms.
**Not this PR.** Any caller (S11-39). Patterns beyond the three ([`ideas/kinship-view.md`](../../ideas/kinship-view.md)).
**Check.** `go test ./core/kinship/...` passes with no build tags; the guard lists the package.
**Context.** Working rules under [Kinship shapes](#kinship-shapes); [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) (a new layer-side package and its guard line).

#### S11-49 — Kinship links in the graph cache

From the S11-39 ship loop (2026-10-26): deriving the canon's implied links on every proposal, and memoizing the walk per Source, both measured badly on a 12-generation tree ([where implied links live](kinship-shapes.md#where-implied-links-live)).
**Ships.** The patterns gain the user's 2026-10-26 decisions from #567: `kinship.Rules{MaxPivotEnd}` read from `matchregistry` (10), a linear grandparent dedupe, and the marriage pattern reading `subject` with the `spouse` participation role removed from the seed, `vocabkey` and Swift. The graph cache keeps one kinship result per pivot (an event pivot: a `marriage` or `birth` node's own participations; a chain pivot: one person's direct parent links and the birth parents its event neighbors imply), derived on first read and dropped when a commit touches the pivot or a node one hop out, and on a vocabulary commit, through a registered per-node memo slot that replaces the dead `details` memo. A new `core/database/canonlinks` answers one person's implied links from those results (`Reader.For`), owns the link reading and terms that `promotealign` moves to it, and registers the Verify check. Tests: each pattern; commits that change a marriage, an event type, a filed parent link and a birth; an oracle against `kinship.Derive`; the 10 vs 11 bound; verify mode.
**Not this PR.** Any reader (S11-39, S11-40). The Source side (derived per layer, S11-39). Patterns beyond the three.
**Check.** `go test -tags fts5 ./core/...` passes, and `canonlinks` passes again with `PROVENENCIA_GRAPH_VERIFY=1`; `For` equals `kinship.Derive` over the whole canon for every person in the oracle test.
**Context.** [Where implied links live](kinship-shapes.md#where-implied-links-live); `core/database/graphcache/commit.go` (`OnCommit`, `entityNeedsReload`), `load.go` (`install`), `verify.go`; [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md).

#### S11-50 — Matching warms while the project opens

From the S11-39 ship loop (2026-10-26). The first proposal after open paid for loading every person, event and place and building the name index (about 0.65 s of 1.05 s on a 12-generation tree, three times that on CI).
**Ships.** `catalogsession.WarmOnOpen` also loads the shortlist's kinds into the graph cache and builds the name index, so the open spinner covers them ("Move everything behind the spinner so that it's honest", user 2026-10-26). Every operation that drops the graph cache, the name index or their vocabulary (a vocabulary commit, `Drop`, `SetAliasIndex`, `SetNameFormat`) warms them again before it returns. `TestLargeSourceBudgets` gains an open budget (opening the project, warming included) and its cold case becomes the first proposal after open; S11-21's other budgets are unchanged. The plan sets the open budget from CI's measurement and asks the user to confirm it.
**Not this PR.** Ranking candidates ahead of a Source (it depends on the Source). Keeping the name index current on commit ([`ideas/name-index-on-commit.md`](../../ideas/name-index-on-commit.md)). A background warm (only if opens grow too long).
**Check.** A test catalog opened through `catalogsession.Do` proposes with no kind load and no index build; the budgets pass on CI.
**Context.** `core/catalogsession/session.go` (`openResearcher`, `WarmOnOpen`); `core/database/promotealign/index.go` (`loadShortlist`); `perf_test.go`, `bench_test.go`; S11-21 D-1.

#### S11-39 — Matching reads kinship shapes

Re-cut 2026-10-26 after its first ship loop ([#567](https://github.com/mendahu/provenencia/pull/567), closed; ship log on its branch); re-planned on S11-49 and S11-50, porting #567's shapes, count-once rule, merge, goldens and budget fixture.
**Ships.** Both loaders merge links through one `promotealign.linkSet` (implied links enter only through `addImplied` and carry a `kinship.Recording`): the layer loader derives the Source side's implied links (their `Associations` are the bridges behind them, for S11-40), and `loadCanon` reads the canon's through `canonlinks.Reader.For` (S11-49) for each walked person and keeps those whose two ends it walked. Implied bridges are never filed and never shown as connections; a test pins what the row's "via" reason says when the neighbor was reached through an implied link. Goldens in `graphalign` / `promotealign`: a Source marriage against a filed spouse relationship, a birth's father against a filed parent link, a Source parent chain against a filed grandparent, each the reverse too, and a fact recorded in two shapes on one side credited once. Each implied link records its shape (`kinship.Shape`), and graph alignment's `corresponding` skips a link pair both sides imply through the same pattern once the walk has paired its pivots (working rule "Counted once across sides"). One type-id → kind lookup (`Vocabulary.TypeKeysByID`) replaces `typeKey`'s scan. The large-source fixture gains a 12-generation tree with a marriage per couple, and every budget passes on CI. [`matching-model.md`](../../matching-model.md) and [`promote-graph-alignment.md`](../../promote-graph-alignment.md) say links are read through the kinship patterns.
**Not this PR.** The evidence query's canon side and the shape on the line (S11-40); a Source-side implied link already reaches the sheet as today's relationship line (user, 2026-10-26). A memo of the walk. A second link loader.
**Check.** The goldens above pass; a Source with only a marriage proposes the couple at the band the same Source with a spouse relationship gets; `TestLargeSourceBudgets` passes on CI with the tree.
**Context.** [Counting once](kinship-shapes.md#counting-once), [weight](kinship-shapes.md#weight); [layer 6 procedure](comparison-layers.md#layer-6-procedure); S11-14 D-11 (link weights); #567's ship log (measurements and the user's decisions).

#### S11-51 — One person's links for every reader

Added 2026-10-26 from Spike 13's map review (was S13-19, [review K-2](../spike-13/deployment-plan.review.md)): the Person page's Family section needs a person's links, direct and implied, merged once, and would otherwise be the third merge.
**Ships.** A move with no behavior change: `sideLink`, `linkSet` (`add`, `addImplied`, `withImplied`) and the merge key leave `promotealign/links.go` for the pure `core/kinship` as `kinship.LinkSet` (user, 2026-10-26, in the plan's review: the Source loader uses it too, and it reads no catalog), and `loadCanon`'s implied-link merge becomes `LinkSet.PlaceImplied`. `canonlinks` gains `Reader.Relationships`, a canonical person's relationship links, recorded (read as `Reader.For` reads them) and implied (`Reader.For`), merged through one `LinkSet`. `loadLayerBridges` and `loadCanon` use the moved set; S11-40's canon side reads `Relationships`. [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md)'s kinship bullet names `kinship.LinkSet` as the one merge.
**Not this PR.** Siblings or any other pattern (Spike 13, S13-03, page-only); any RPC (S13-04).
**Check.** Goldens, `TestLinkSet` and `TestLargeSourceBudgets` unchanged; `canonlinks` passes with `PROVENENCIA_GRAPH_VERIFY=1`; a test that a spouse link recorded directly and implied by a marriage comes back once from the export.
**Context.** Working rule "One fact, one link per side"; S11-39 plan D-5 – D-7; Spike 13 working rule ["One person's links, direct and implied, from one function"](../spike-13/deployment-plan.md#engine).

#### S11-40 — Implied links in claim evidence

**Ships.** The evidence query's canon side (`canonLinks` in `promotealign/evidence.go`, which reads graph nodes, not `Canon.Edges`) reads a person's recorded and implied links through S11-51's `canonlinks` export, as the Source side reads its layer's. In `claimevidence`, `Relationship` gains how each side recorded the fact (directly, through an event, through a chain) and the event behind it (its Subject or entity, type and date), and a corresponding implied link gives a relationship line whose bundle holds every bridge behind it and the event's defining properties. The plan decides the bundle's shape (one bundle per underlying bridge, or one bundle of several) and its pin id (`BundlePinID` keys on one bridge today). The RPC carries it (proto field, handler, `GoStore` mapping; the tree's pivot travels as its existing `EventHeader` or `PersonHeader`, read through a new `conclusionheaders.ByIDs` that `proposalProto` and the search hits' header read also move onto), and the shape code joins `ConclusionVocabulary.swift` under `swift_vocabulary_test.go`. The plan decided one bundle per Source recording of an implied link, keyed by the member's own bridge. Default pins at Apply follow. A relationship line follows the count-once rule (working rule "Counted once across sides"). The sheet keeps today's sentence until S11-42.
**Not this PR.** The sheet's new sentences (S11-42). Trees and person pages ([`ideas/kinship-view.md`](../../ideas/kinship-view.md)).
**Check.** A Source marriage against a filed spouse relationship returns one relationship line marked "through an event" whose bundle holds both participations; a tree marriage against a Source spouse relationship returns the reverse; Done on an unopened row pins the bundle.
**Context.** [On the evidence sheet](kinship-shapes.md#on-the-evidence-sheet); [`claim-evidence.md`](claim-evidence.md#evidence-through-relationships); `core/claimevidence/build.go` `bundles`, `pins.go`; [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md).

#### S11-42 — Implied links on the evidence sheet

**Ships.** The relationship Card keeps its header ("…they share a spouse (refs)") and, when either side records the fact through a kinship pattern, says how each side recorded it in a muted line under the header: "This source: married in the marriage dated 12 May 1842 · The record: recorded as spouses" (an event by its date, a chain by the parent in between; user, 2026-10-27). Through L10n. No new chrome.
**Not this PR.** Engine or wire (S11-40).
**Check.** A Source marriage against a filed spouse relationship shows "This source: married in the marriage dated <date> · The record: recorded as spouses" under the Card header; checking it pins both participations.
**Context.** [On the evidence sheet](kinship-shapes.md#on-the-evidence-sheet); [`claim-evidence.md`](claim-evidence.md#evidence-through-relationships); [`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md).

### Person vocabulary

#### S11-48 — Religion and languages spoken on a person

The user's call on 2026-10-25: record religion and languages as term Properties first; Organization as a Subject type is parked in [`ideas/organizations.md`](../../ideas/organizations.md).
**Ships.** `religion` and `languages_spoken` in `subjectvocab`'s registry: term, cardinality multiple, bound to person unlocked, one product term each (Catholic, English). Researchers add the rest as `origin=user` terms. [`seeded-vocabulary.md`](../../seeded-vocabulary.md) and the cardinality paragraph of [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) follow.
**Not this PR.** Existing catalogs (a one-off command the user runs by hand, S11-44). Matching weights. Swift wording.
**Check.** A new project's person takes Religion → Catholic and a custom Anglican, and the Person page shows both.
**Context.** [`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md).

### Slice 10 — Name candidates

From dogfood on 2026-10-24. Design and diagnosis: [`name-candidates.md`](name-candidates.md).

#### S11-43 — Lower the candidate gate

**Ships.** `alignment.MinSimilarity` 0.4 in `core/matchregistry/alignment.go`; its comment (and the inventory row in [`matching-model.md`](../../matching-model.md)) drops "at least half alike" and gives the Amy case as the reason. A `graphalign` test: a one-given-name Subject and a record with three more names is a candidate and shows as an alternative on an unplaced row; a record whose first given name differs is still not a candidate. The obituary golden and sparse cases keep their targets (the map review ran the suite at 0.4 with no flips).
**Not this PR.** The name curve (held, user 2026-10-24; [what stays later](name-candidates.md#what-stays-later)). Surname-only ties ([`ideas/coverage-aware-similarity.md`](../../ideas/coverage-aware-similarity.md)).
**Check.** On a real project, a row for "Amy" lists "Amy Lee Anne Long".
**Context.** [Case 1](name-candidates.md#case-1--amy-against-amy-lee-anne-long); [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md).

#### S11-44 — Des as a nickname of Desiree and Desirae

**Ships.** `{desiree, des}` and `{desirae, des}` in `nicknameSets` (`namealiases/nicknames_seed.go`, already the product's reviewed edit of upstream; two sets because a shared nickname does not join two formal names), the reconciler `CacheVersion` bump and the seed hash re-pinned (S11-28's recipe).
**Not this PR.** Reaching an existing catalog: the user runs a one-off command by hand, with the app closed, that adds the new nickname sets and S11-48's Properties, terms and person bindings (user 2026-10-27: "We don't need to build anything for this."). An editor for alias tables.
**Check.** In a new catalog, "Desiree" and "Desirae" are offered "Des" and the reverse.
**Context.** [Case 2](name-candidates.md#case-2--desiree-against-des); S11-28 (the reviewed table beside upstream).

#### S11-47 — Promote suggests a record or nothing

The user's call on 2026-10-24, answering S11-43's open question.
**Ships.** `graphalign/rows.go`: a row Align does not place has no target (neither `TargetSkip` nor `TargetNew`) and no row reason `weak` or `no_match`; its alternatives stay as they are, each with its own reason (`ReasonWeak` on an alternative stays). The possible-duplicate check finds unplaced rows instead of `TargetNew` rows. `TargetNew` and `TargetSkip` stay for `Fixed`. `PromoteGraphAlignmentRow.target`'s comment says it is `handle` or empty; the handler and `FakeStore` follow. Swift: `PromoteFlow.draftTarget` drafts the proposed handle, the weak-band first alternative as today, or unset, never `.skip` or `.newKind`; `Reason`'s row cases `weak` and `noMatch` and their L10n go if nothing else reads them (`ConclusionVocabulary.swift`, `swift_vocabulary_test.go`, `PromoteModel.reasonText`). Tests asserting `TargetNew` / `TargetSkip` / `ReasonNoMatch` on a proposal row (for example `TestLifespanRulesOutAPerson`) assert an unplaced row instead. [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §3 says Promote drafts no New or Skip.
**Not this PR.** Decided rows: New and Skip as fixed pairs are unchanged. Which record a weak band drafts.
**Check.** A row with no candidate and a row whose candidate fell short both open unset; two look-alike unplaced rows still get the "also new" note once both are set to New; Done files a decided New and leaves a decided Skip off.
**Context.** [Promote suggests a record or nothing](name-candidates.md#promote-suggests-a-record-or-nothing); [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md).

#### S11-46 — Prefix keys in the shortlist

Split from S11-45 by the map review (K-2, K-7, 2026-10-24).
**Ships.** `valuecompare.WordKeys` returns separate index and lookup lists, and `shortlist.Keys` files and looks up each (no behavior change for today's keys). Words of the alias roles are also filed under their prefixes from `PrefixMinLength` up (a new registry value knob, 2 by the user's call, with its inventory row), with a tag apart from whole words; a lookup asks its whole word against filed prefixes and its own prefixes against filed whole words, so Des reaches Desiree and Desiree reaches Des. The extra keys and shortlist hits are measured on S11-21's fixture or `bench_test.go`'s, with a cap from the measurement if needed. Surnames are not keyed by prefix.
**Not this PR.** The comparer rule (S11-45): until it lands, a prefix shortlists and the gate drops it.
**Check.** A shortlist property test: Des ↔ Desiree both ways; "Mac" does not shortlist "MacDonald" by prefix; today's keys still pair as before.
**Context.** [Prefix rule](name-candidates.md#prefix-rule-in-detail); `core/shortlist/shortlist.go` `Keys`; `core/valuecompare/shortlist.go`.

#### S11-45 — Prefix names

**Ships.** The prefix rule ([detail](name-candidates.md#prefix-rule-in-detail)): in `valuecompare`'s word comparer after alias, fold and spelling, behind a `prefixes` word option names set and text never does, for the alias roles; `PrefixCredit` in the registry value section with its inventory row; `CausePrefix`; the broader side from word length, not `IsInitial` (`name.go`). The cause joins `ConclusionVocabulary.swift` under `swift_vocabulary_test.go`, worded through L10n on the sheet's existing name line. `CausePrefix` joins the reconciler's fold allow-list (`autoreconcile/names.go`, after S11-28 widens it), so "Des Long" shows as "Desiree Long" within one entity (user, 2026-10-24). Extends the name ladders in `name_test.go` and the contract tests; "Ann" / "Anne" stays a spelling `Near`.
**Not this PR.** Surname prefixes. (Two-letter prefixes count: `PrefixMinLength` is 2, user 2026-10-24.)
**Check.** "Desiree" against "Des" (no nickname set) is `Contains` with `CausePrefix` and is offered; "Desmond" against "Desiree" is not a prefix; "Mac" does not reach "MacDonald".
**Context.** [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md); [`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md).
