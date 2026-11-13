# S16-16 — Kinship answers are composed from the pivots, and the Family read is one pass

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 2c — Walks
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine refactor; the Family section draws the same units)
**Board:** —
**Depends on:** S16-10 (landed [#789](https://github.com/mendahu/provenencia/pull/789)), S16-01 (landed [#756](https://github.com/mendahu/provenencia/pull/756))
**Blocks:** S16-22 (reads each handle's links from the kinship pivots), S16-24 (instruments the Family read this PR rewrites)
**Branch:** `s16-16-composed-kinship`
**Version bump:** none — no change a researcher can see, no wire and no catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (its kinship lines change; no knob changes), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) (the graph cache's registry and Verify)
**Rules that bind:** `.cursor/rules/go-tests.mdc` (table-driven tests), `.cursor/rules/match-rules.mdc` (no literal knob in `core/kinship`), `.cursor/rules/catalog-queries.mdc` (a derived cache reads only through its registry handle)
**Context:** [design § Results per key](../cache-architecture.md#results-per-key-for-the-walks), [`docs/catalog-caches.md`](../../../catalog-caches.md), [audit P-3 and P-24](../person-page-audit.md), working rules: "The point of this batch is simpler code", "Cache entities and results per key; compose answers", "One cache registry, two ways to fill", "Every cache names the signal that updates it, and updates exactly", "Apply states its fan-out", "Orders are defined and results do not move", "Memory is budgeted", "No new persisted projections, and no stored answer without a measurement", "Batch 1's caps stay where S16-01 puts them"
**Planned:** 2026-11-13 against `main` @ `d1d1347`
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

The graph cache keeps one kinship result per pivot (`canonlinks.kinship`: what an event implies, or what a person implies as the parent between), and a commit drops exactly the pivots one hop from what it touched. On top of that it stores two answers per person, `canonlinks.for` (a person's implied links) and `canonlinks.siblings` (the siblings a shared parent implies), and empties both on every commit that names any handle, because they read nodes too far away for an exact drop. This PR deletes those two entries: `Reader.For` and the siblings are composed from the pivots at each read. The Family read then reads each person's kin once, through one batched `Reader.Relationships` call, instead of loading answers, reading them back, and loading every named relative; `LoadRelationships` and `LoadSiblings` go. `edgeKey` length-prefixes its parts, so crafted ids cannot make two links one key (audit P-24).

**Why now.** Bar 9 ("No cache empties itself on any write") needs the two exceptions gone; S16-22 reads each handle's implied links from the pivots and needs `For` free of a catalog-wide drop; S16-24 instruments the Family read, so it lands after the rewrite.

**What stays.** Every read returns what it returns today: `For` in today's order (so Promote's tie-breaks do not move), the same Family units, origin siblings, Promote suggestions and goldens. The `canonlinks.kinship` entry, its Apply, Check and Cap stay. S16-01's caps stay where they are: `kinship.KinOf`'s `MaxKinPerRole` and `kinship.Siblings`' `MaxPivotEnd`. No SQL, wire, schema or Swift change.

## 2. User stories

**US-1.** As the engine, after an Observation of a filed person is edited, I recompute only the kinship pivots that commit dropped, so Promote and the Family section reuse the rest.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Ann and Bob married (M1); Bea born of Ann and Bob (B1); Cal born of Bea (B2); every person's `For` read once, so every pivot is held. Commit Ann's handle (`w.k.Commit(ann)`, what an Observation edit on Ann commits) | The pivots dropped are exactly Ann's and those of every neighbour on Ann's cached node (M1, B1: the ids `deps` names); the next `For` of every person derives exactly those of them that were held, and no other key |
| AC-1.2 | The same world; commit Cal's handle | `For(Ann)` derives no pivot (Ann is two hops from Cal) and returns the same links as before the commit |
| AC-1.3 | Any world in `TestForFollowsCommits`, `TestSiblingsFollowCommits` | `For` and the siblings after each commit equal a derivation on a fresh graph; no stored answer exists to go stale |
| AC-1.4 | The large fixture at 4,000 and 16,000 persons | A warm proposal and a proposal after a handle commit are measured before and after this PR on the same machine (`BenchmarkLargeSourcePropose`); `TestLargeSourceBudgets` passes |

**US-2.** As a researcher opening a person's Family section, it reads their kin once and draws the same family.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `Reader.Relationships([]{Pat, Bea}, Options{Siblings: true})` cold, on a family of 1 child and on one of 7 children with births | Both issue the same fixed number of graph queries; a second call issues none |
| AC-2.2 | `PersonFamily(Pat)` warm, then a commit naming a person outside Pat's family (`w.k.Commit(stranger)`) | The next `PersonFamily(Pat)` issues no graph query and returns the same `Family`, origin siblings and summaries |
| AC-2.3 | Every existing Family, kinship and Promote test (`TestPersonFamily`, `TestPersonFamilyOriginSiblings`, `TestPersonFamilyQueryCount`, `TestCraftedFamily`, `core/kinship` tests, `TestLargeSourceBudgets` "memo matches a fresh run", `core/graphalign` tests) | Pass unchanged |

**US-3.** As a contributor, the caches table and the kinship guidance name only what exists.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `edgeKey(Edge{A: "ab", B: "c"})` and `edgeKey(Edge{A: "a", B: "bc"})`; an edge with associations `["xy","z"]` and one with `["x","yz"]` | Each pair gives two different keys; the same edge gives the same key twice |
| AC-3.2 | `docs/catalog-caches.md` after this PR | No `canonlinks.for` or `canonlinks.siblings` row, and the Known exceptions paragraph names only `promotealign.matchindex`; `TestCacheDocListsEveryEntry` passes |
| AC-3.3 | `rg -n "LoadRelationships|LoadSiblings|forAnswers|siblingAnswers|canonlinks\.for\b|canonlinks\.siblings|memo slot" core docs .agents .cursor -g '!docs/deployment-plan/**'` | No hit |

## 3. Scope

| In | Covered by |
| --- | --- |
| `BenchmarkLargeSourcePropose` *(new)* and the base figures, before any change | STEP-1, AC-1.4 |
| The `canonlinks.for` and `canonlinks.siblings` entries go; `For` composes from the pivots | STEP-2, AC-1.1 – AC-1.3 |
| `Reader.Relationships` reads many ids in one call; `LoadRelationships`, `LoadSiblings`, `Reader.siblings` go; `siblingsOf` and `forLinks` take the loaded steps (one step derivation) | STEP-2, AC-2.1 |
| `edgeKey` length-prefixes its parts (P-24) | STEP-2, AC-3.1 |
| `persondetail.Families` reads each person's kin once | STEP-2, AC-2.2, AC-2.3 |
| Bar 9's pivot-reuse test and the Family-after-commit test | STEP-3, AC-1.1, AC-1.2, AC-2.2 |
| Measure after; `BenchmarkCacheSizes` kinship figure | STEP-4, AC-1.4 |
| Docs: `catalog-caches.md`, `add-match-rule`, `matching-model.md`, package and `promotealign` comments, the map rule and row | STEP-5, AC-3.2, AC-3.3 |

| Out | Home |
| --- | --- |
| Promote's canon reading each handle's links from the pivots instead of walking the component | S16-22 |
| The match index's drop-on-any-handle Apply | S16-20 |
| The Family read reporting its inputs on the wire | S16-24 |
| A kinship index, reverse map or open-time warm of the pivots | not planned (map review K-20; rule "No stored answer without a measurement") |
| Memoising a composed `For` if it measures slow | not planned unless STEP-4 hits the Stop condition; then the user decides |
| Family trees beyond the nuclear family | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |
| Changing the pivot Apply's one-hop drop | not planned (exact already) |

## 4. Current state

Verified at `d1d1347`.

| Area | Where | Today |
| --- | --- | --- |
| Pivot entry | `core/database/canonlinks/reader.go` `pivots`, its `init` registration (`Name: "canonlinks.kinship"`), `deps`, `check`, `Reader.pivot`, `Reader.event`, `Reader.chain` | One `pivotResult` per event or person id; Apply drops the roots and `deps(g, roots)` (every cached neighbour of each root); Check re-derives from cached nodes. Stays as is |
| For answer entry | `reader.go` `forAnswers` (`canonlinks.for`), `personAnswer`, `checkFor`, `checkPersonAnswer`, `clearOnHandles` | Stores each person's `For` answer; Apply `clearOnHandles` empties it when a commit names any handle. Comment says "S16-16 deletes this entry" |
| Siblings entry | `reader.go` `siblingAnswers` (`canonlinks.siblings`), `checkSiblings`, `Reader.siblings`, `Reader.LoadSiblings` | Stores each person's implied siblings; same clear-on-any-handle Apply. Comment says "S16-16 deletes this entry" |
| `For` | `reader.go` `Reader.For(id)` | `forAnswers.GetMany` with a fill that calls `d.around(id)` and `d.forLinks(id, r.event, r.chain)`; drops nothing for a non-node id (S16-10 Landed note) |
| `forLinks` | `reader.go` `deriver.forLinks` | The composition itself: the links of each event id stands in (`eventLinksAt`), then for each person one parent step away (`stepPeople` over `kinEdges`) `around(m)` and the chain links with id at an end. Already reads only pivots and nodes |
| `Relationships` | `reader.go` `Reader.Relationships(id, opts)` | One id: `around(id)`, its relationship edges, `For(id)`, and with `opts.Siblings` the stored siblings, merged through one `kinship.LinkSet`. Returns nil for a non-node |
| Loaders | `reader.go` `Reader.LoadRelationships(ids, opts)` | `d.loadSteps(ids, r.event)` (four batched loads), then `For` per id to store it, then `LoadSiblings`. Exists so a later `Relationships` per id issues no query |
| Step derivation | `reader.go` `stepped`, `deriver.loadSteps`, `deriver.forLinks`, `deriver.around` | `loadSteps(ids, event)` works out each node's `parentEdges` (`eventLinksAt` then `kinEdges`) and `stepPeople` in four batched loads. `forLinks` works out the same three again for one id, loading through `around(id)` and one `around(m)` per stepped person: a second derivation of the steps, and 2 + N loads cold |
| Siblings derivation | `reader.go` `deriver.siblingsOf(ids, event)` | `siblingsOf` calls `loadSteps` itself, takes each node's parents (`kinship.Parents`), reads each parent's `parentEdges`, merges them into the node's edges deduplicated by `edgeKey`, and calls `kinship.Siblings` |
| `edgeKey` | `reader.go` `edgeKey` | `string(e.A) + string(e.B) + string(bytes.Join(e.Associations, nil))`: no separator, so `"ab"+"c"` and `"a"+"bc"` collide (P-24) |
| Package doc | `core/database/canonlinks/canonlinks.go` package comment | Names `LoadSiblings` and `LoadRelationships` |
| Family read | `core/database/persondetail/family.go` `Families` | `g.Nodes(want)`; keeps the Persons; `r.LoadRelationships(people, {Siblings: true})`; per person `r.Relationships` + `kinship.KinOf`, collecting every relative any role names; `r.LoadRelationships(relatives, {})` (all named relatives, though only parents and children are read next); per person `shownKin`, then per `kin.Around(id)` relative `r.Relationships` + `KinOf` + `shownKin`, then `kinship.FamilyUnits` |
| `shownKin` | `family.go` `shownKin` | Keeps a link only when `g.Cached(other)` is a canonical Person; reads only cached nodes, so every relative it filters must be loaded before |
| Promote's use | `core/database/promotealign/canon.go` `loadCanon` (the `reader.For(id)` loop over `walkedPeople`), doc comment at line 56 | Calls `For` once per walked person after the walk loaded the component |
| Cache doc | `docs/catalog-caches.md` rows `canonlinks.kinship` ("1.4 / 4.4 MB with `canonlinks.for` and `canonlinks.siblings`"), `canonlinks.for`, `canonlinks.siblings`; § Known exceptions | Lists the two entries and the exception |
| Guidance | `.agents/skills/add-match-rule/SKILL.md` line 33 (`LoadSiblings`, `LoadRelationships`), line 88 ("its own memo slot"), line 91 (step 5: `Reader.For` and "how far `Deps` reaches"); `docs/matching-model.md` line 100 (**pivot**: "derived on first read (`canonlinks.Reader.For`)") | Describe the stored answers |
| Map | `docs/deployment-plan/spike-16/deployment-plan.md` rule "Every cache names the signal that updates it, and updates exactly" | "the drop-on-any-handle Applies of `for` and `siblings` (S16-16) and of the match index (S16-20) are the known exceptions" |
| Tests | `reader_test.go` `TestRelationshipsOptionsSiblings` subtest "AC-1.9" (asserts with `siblingAnswers.Peek` that a wrong-length and an unknown id stored nothing) and subtest "an edge in both lists is read once" (its `fromNames` reads the From edges of the stored sibling answer through `siblingAnswers.Peek`); `TestLoadSiblingsQueryCount` AC-2.1 calls `w.r.Relationships(bea, …)` directly | Read the deleted entry |
| Tests | `core/database/promotealign/promotealign_test.go` `TestProposeIgnoresStoredSiblings` | Calls `reader.LoadSiblings([][]byte{beaID, calID})`, then `reader.Relationships(beaID, Options{Siblings: true})`, and checks the proposal is unchanged after siblings were stored |
| Tests | `core/database/canonlinks/reader_test.go` `TestFor`, `TestRelationships`, `TestForFollowsCommits`, `TestForMatchesDerive`, `TestVerifyCatchesAStaleResult`, `TestRelationshipsOptionsSiblings`, `TestSiblingsFollowCommits` (uses `hasSiblings`), `TestLoadSiblingsQueryCount`, `TestLoadRelationshipsQueryCount`, `TestSiblingsLeaveMatchingAlone`, `TestVerifyCatchesAStaleSiblingAnswer`, `TestPivotKindMismatch`; helpers `world.relationships`, `world.siblings`, `world.hasPivot`, `world.hasSiblings` | Hold today's answers and the stored entries |
| Tests | `core/database/persondetail/family_query_test.go` `TestPersonFamilyQueryCount` (AC-2.1 warm zero, AC-2.2 small = large, AC-2.3 one = three people); `family_test.go` `TestPersonFamily`, `TestPersonFamilyOriginSiblings`; `caps_test.go` `TestCraftedFamily` (calls `r.Relationships(pat, …)`) | Hold the Family read |
| Tests | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` (4,000 persons; `budgetWarm` 500 ms, `budgetAfterCommit` 2.5 s), `newLargeFixture`; `bench_test.go` `BenchmarkCacheSizes` "kinship" fill (calls `r.Relationships(id, …)` per person) | Promote's budgets at 4,000 only; no 16,000 proposal measurement exists |
| Tests | `core/catalogsession/cache_doc_test.go` `TestCacheDocListsEveryEntry` | Holds the doc's first table to the registry's names |

**Precedent.** [#789](https://github.com/mendahu/provenencia/pull/789) (S16-10) moved the three `canonlinks` slots onto the registry and left these two entries marked for deletion: copy its `docs/catalog-caches.md` row edits and `BenchmarkCacheSizes` reporting. [#809](https://github.com/mendahu/provenencia/pull/809) (S16-13) deleted a stored mechanism in favour of entries and recorded before/after kinship figures in its ship log: copy that measuring (same machine, before and after, in the ship log). The batched read follows `graphcache.Keyed.GetMany`'s convention: a map keyed by `string(id)`, an id that is no node absent.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Delete `forAnswers`, `siblingAnswers`, their `init` registrations, `personAnswer`, `clearOnHandles`, `checkPersonAnswer`, `checkFor`, `checkSiblings`. `Reader.For(id)` composes through the one step derivation (D-11): `steps, err := r.d.loadSteps([][]byte{id}, r.event)`; no step (an id that is no node) → `nil, nil`; else `r.d.forLinks(steps[0], r.chain)`. Its doc says it composes the answer from the pivot results at each call, the slice is the caller's to read and must not be changed (its links are the pivots'), and the order is the links of each event id stands in, then the chain links through each person one parent step away | `forLinks` is already the composition from pivots; the stored answer was the only thing that needed a catalog-wide drop. Map rule "Cache entities and results per key; compose answers" (K-20) | Keeping a memo with a narrower Apply — a reverse map or three-hop deps, which K-20 refused; memoising before a measurement says slow — rule "No stored answer without a measurement" |
| D-2 | `Reader.Relationships` becomes `Relationships(ids [][]byte, opts Options) (map[string][]kinship.Link, error)`: one call loads every id through `d.loadSteps(ids, r.event)` (four batched loads however many ids and relatives), then, per stepped node in `loadSteps` order, merges its relationship edges, `r.d.forLinks(step, r.chain)` (D-11) and, with `opts.Siblings`, its siblings from one `d.siblingsOf(steps, r.event)` call, through one `kinship.LinkSet` each, as today. An id that is no node has no entry. The single-id form goes; every caller passes a slice | One read that loads and composes replaces the load-then-read pair, so the Family read cannot read something it did not load, and `LoadRelationships` / `LoadSiblings` have no job left | Keeping `Relationships(id)` beside a batch form — two ways to read the same answer; keeping `LoadRelationships` as a loader with no store behind it — a call whose only effect is a side effect on node loading |
| D-3 | `deriver.siblingsOf` becomes `siblingsOf(steps []stepped, event func([]byte) ([]kinship.Implied, error)) (map[string][]kinship.Implied, error)`: it takes the steps `Relationships` loaded instead of calling `loadSteps` itself. The rest of its body stays | `Relationships` already holds the steps; deriving them twice is the duplicate D-11 removes | Calling `loadSteps` twice — four extra no-op passes over cached ids per call |
| D-4 | `edgeKey` writes, for `e.A`, `e.B` and each association in order, `binary.AppendUvarint(len)` then the bytes, and returns the string. The name and its one caller stay | P-24's fix; the audit names length prefixes. Ids are not length-checked when loaded, so a 16-byte assumption is not safe | Skipping ids that are not 16 bytes — drops data silently; a separator byte — ids are arbitrary bytes and can contain it |
| D-5 | `Families` reads in this order: (1) `g.Nodes(want)` and keep the Persons, as today; (2) `links, err := r.Relationships(people, canonlinks.Options{Siblings: true})` and `kin[id] = kinship.KinOf(id, links[string(id)], terms, rules)`; (3) `g.Nodes(named)` where `named` is every relative any role of any person's kin names, first seen, so `shownKin` reads loaded nodes; (4) `shown[id] = shownKin(g, typeID, id, kin[id])`, and `around` is every id of `shown[id].Around(id)` over all people, first seen; (5) `aroundLinks, err := r.Relationships(around, canonlinks.Options{})` and `aroundKin[other] = shownKin(g, typeID, other, kinship.KinOf(other, aroundLinks[string(other)], terms, rules))`, each computed once; (6) per person, `kinship.FamilyUnits(id, shown[id], <aroundKin restricted to shown[id].Around(id)>, rules)` — FamilyUnits reads only the ids it is given, so passing the whole `aroundKin` map gives the same units and is what the code does | Each person's and each relative's kin is read once; only parents and children get a `Relationships` read, which is all `FamilyUnits` reads; the spouses and siblings need only their nodes, for `shownKin`. Today every named relative gets `LoadRelationships` | Reading `Relationships` for every named relative, as today — composes kin for spouses and siblings nobody reads; computing `around` from the unfiltered kin — reads kin of relatives the page does not show |
| D-6 | `around` relatives with no node, or that are not people, are passed to `Relationships` anyway; it skips them (no entry), and `aroundKin` has no entry, which `FamilyUnits` reads as "none" (its documented contract) | One rule for a missing relative, the one `FamilyUnits` already has | A pre-filter — a second place that decides who has kin |
| D-7 | Bar 9's counter is the pivot entry's own keys: the test reads `pivots.Peek` over every id in the world before and after, and a derived key is one held after and not before. No production counter is added | The test is in package `canonlinks` and `Peek` already answers "is it held"; a derive hook would be production code for one test | A package-level derive hook — test plumbing in production; `g.KeyCounts()` — a count cannot say which keys |
| D-8 | The two Scale proposal timings move into `largeFixture` methods in `core/database/promotealign/perf_test.go`: `warmProposal() time.Duration` *(new)* (one untimed `propose` of `lf.a`'s graph, then `databasetest.Fastest(5, …)` of `runtime.GC()` + `databasetest.Since(propose)`: the body of `TestLargeSourceBudgets` "warm") and `proposalAfterCommit(n *int) time.Duration` *(new)* (`databasetest.Fastest(2, …)` of: `*n++`, `f.promote(f.Person(lf.other, lf.otherA, "Added "+letters(*n)+" Person"))`, then `Since(propose)`: the body of "after a handle commit"; the counter is the caller's so names never repeat). `TestLargeSourceBudgets` "warm" and "after a handle commit" call them and keep their budgets. `BenchmarkLargeSourcePropose` *(new)* in `bench_test.go`: for `persons` in 4,000 and 16,000, `b.Run("persons=N")` builds `newLargeFixture(b, persons)`, ignores `b.N` as `BenchmarkCacheSizes` does, and reports `warm-ms` then `after-commit-ms` (that order: after-commit adds people to `lf.other`) with `b.ReportMetric(float64(d.Microseconds())/1000, …)`. Reported, not gated. The ship log (`S16-16-composed-kinship.ship.md`, beside this plan; the orchestrator creates it) records the figures with the output of `hostname` | The map's Check asks for Promote's figures at 4,000 and 16,000; only 4,000 has a budget today. S16-20 and S16-22 need the same figures for bar 10 | Gating at 16,000 — bar 10's budget is S16-22's to set; reusing `BenchmarkProposeWarmCity` — its fixture is not the map's Scale; copying the two bodies into the benchmark — a third copy of the after-commit timing beside `BenchmarkProposeAfterCommit` (review M-2) |
| D-9 | `For` stays a single-id read; Promote's `loadCanon` loop is unchanged | Promote's walk already loads the component before the loop, so a batched `For` saves nothing there, and S16-22 replaces the loop | A batched `For` — an API with no caller that needs it |
| D-10 | Tests that asserted storage of the deleted entries change to assert answers: `TestSiblingsFollowCommits` AC-2.4 and AC-2.5 drop their `hasSiblings` checks and compare `w.relShapes(w.siblings(bea))` after the commit with the same read on a fresh world built by the same subtest setup; `TestRelationshipsOptionsSiblings` "AC-1.9" drops its two `siblingAnswers.Peek` checks and asserts `Relationships([][]byte{{1, 2, 3}, make([]byte, 16)}, Options{Siblings: true})` returns an empty map; its `fromNames` reads the same `[]kinship.Implied` the stored answer held, from the derivation itself: `steps, err := w.r.d.loadSteps([][]byte{id}, w.r.event)`, then `answers, err := w.r.d.siblingsOf(steps, w.r.event)`, and loops over `answers[string(id)]`'s `From` edges as today (a merged `kinship.Link` carries no `From`, so the derivation is the only place to see them); `TestVerifyCatchesAStaleSiblingAnswer` and the `hasSiblings` helper go; `TestLoadSiblingsQueryCount` and `TestLoadRelationshipsQueryCount` become one `TestRelationshipsQueryCount` (AC-2.1); `promotealign_test.go` `TestProposeIgnoresStoredSiblings` becomes `TestProposeIgnoresPageSiblings`: its `LoadSiblings` call goes, `Relationships` takes `[][]byte{beaID, calID}` and the setup reads `links[string(beaID)]`, its comment and failure message say "after the page read siblings" | The behaviour they guard (answers follow commits; cold reads are a fixed number of batches) stays; the storage they also checked is gone | Deleting the tests — loses the commit-following and query-count guards |

| D-11 | One step derivation (review M-1): `stepped` gains `atEnd []kinship.Implied`, the `eventLinksAt` result `loadSteps` already computes inside `parentEdges` (so `loadSteps` calls `eventLinksAt` then `kinEdges` itself and keeps both). `forLinks` becomes `func (d *deriver) forLinks(s stepped, chain func([]byte) ([]kinship.Implied, error)) ([]kinship.Implied, error)`: start from a clone of `s.atEnd`, then, for each of `s.people` in order, append the links of `chain(m)` with `s.node.ID` at an end. `deriver.around` has no caller left and goes | Same links in the same order as today: `forLinks` derived `eventLinksAt`, then `stepPeople(id, kinEdges(n, out))` — what `loadSteps` derives. `For` cold becomes four batched loads instead of 2 + N; one derivation of the steps serves `For`, `Relationships` and the siblings | Keeping `forLinks` with its own `around` loads — a second step derivation beside `loadSteps` in one package |

**Rules this plan bends.** None. The map's scope entry and the rule "Every cache names the signal that updates it, and updates exactly" name this deletion; STEP-5 rewords the rule to drop S16-16's half of the exception.

## 6. Design

### Codebase fit

Searched: `GetMany`, `Relationships(`, `Load` readers in `core/database` returning `map[string]`, `Bench.*Propose`, `newLargeFixture`, `Peek(` in tests, `AppendUvarint` in `core/`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `Reader.Relationships(ids [][]byte, opts Options) (map[string][]kinship.Link, error)` (changed signature) | func | `Reader.Relationships(id, opts)` + `Reader.LoadRelationships` | It is those two folded into one; the map return follows `graphcache.Keyed.GetMany` |
| `BenchmarkLargeSourcePropose` *(new, test)* | benchmark | `TestLargeSourceBudgets`; `BenchmarkCacheSizes`' `persons` loop | Calls the two helpers at both sizes; the budget test runs at 4,000 only and gates, the sizes benchmark measures heap and its fills would warm the proposal it times |
| `TestComposedAnswersReusePivots` *(new, test)* | test | `TestForFollowsCommits` | That test checks answers, not which pivots were derived |
| `largeFixture.warmProposal`, `largeFixture.proposalAfterCommit` *(new, test)* | helper | the bodies of `TestLargeSourceBudgets` "warm" and "after a handle commit" | They are those bodies, moved so the gate and the report time a proposal one way (review M-2) |
| `stepped.atEnd` *(new field)* | field | `loadSteps`' local `eventLinksAt` result | Kept instead of derived again by `forLinks` (D-11) |
| `TestRelationshipsQueryCount` *(new name, test)* | test | `TestLoadSiblingsQueryCount`, `TestLoadRelationshipsQueryCount` | It is the two merged onto the one read |
| `TestEdgeKey` *(new, test)* | test | none for `edgeKey` | — |

**Builds on.** The cache registry's keyed entry `canonlinks.kinship` (S16-10) and its exact one-hop Apply; `deriver.loadSteps` as the one step derivation, with `forLinks` and `siblingsOf` reading its steps (D-11, D-3); `kinship.KinOf`'s cap (S16-01).
**Deletes.** `forAnswers`, `siblingAnswers` and their registrations; `personAnswer`; `clearOnHandles`; `checkPersonAnswer`, `checkFor`, `checkSiblings`; `Reader.LoadRelationships`; `Reader.LoadSiblings`; `Reader.siblings`; the single-id `Relationships`; `deriver.around` and `forLinks`' own step derivation (D-11); in tests `world.hasSiblings`, `TestVerifyCatchesAStaleSiblingAnswer`, `TestLoadSiblingsQueryCount`, `TestLoadRelationshipsQueryCount`; in docs the two `catalog-caches.md` rows and S16-16's half of the Known exceptions paragraph and of the map rule. `stepped`, `loadSteps`, `siblingsOf`, `forLinks`, `stepPeople` stay: the composition uses them.
**Next adds.** S16-22's canon view reads many handles' implied links through `loadSteps(ids)` and `forLinks` per step (exact, no catalog-wide drop), with no new API. S16-24 records the reader's reads by giving `deriver` a `nodeSource` backed by a `graphcache.InputScope` (the reader's one load policy, beside `loadingNodes` and `cachedNodes`), not by adding parameters to `Relationships`; with D-11 every node the composed answers read goes through `loadSteps` and that source. S16-20 and S16-22 report bar 10 with `BenchmarkLargeSourcePropose`.

### Reach

None — no new path into existing code. `Relationships` reaches the same derivations for the same ids; an id that is no node now has no map entry instead of a nil slice, and every caller (`Families`, `TestCraftedFamily`, `BenchmarkCacheSizes`, `TestProposeIgnoresPageSiblings`, the `reader_test.go` helpers) reads a missing entry as no links.

### Scale

At the map's Scale (large fixture at 4,000 and 16,000 persons; 18,000 and 37,000 handles) and bar 2's crafted sizes:

- **Apply per commit.** `canonlinks.kinship`'s Apply is unchanged: `Drop(roots)` and `Drop(deps(roots))`, one map delete per root and per cached neighbour. A commit naming a 20,000-child hub drops about 20,001 keys (map deletes, single-digit milliseconds). The two `Clear` calls on any handle commit go, so a commit no longer throws away up to 20,115 + 20,115 keys (S16-10's figures at 16,000) for the next read to rebuild.
- **Reads.** `For(id)` costs, at each call, `loadSteps`' four load calls (no query when the nodes are cached), then map lookups for each event id stands in and each person one parent step away, plus the filter over those pivots' links: what the first `For` after any commit cost today. A warm Promote proposal pays it for every walked person, where today it hit the stored answer; STEP-1 and STEP-6 measure that at both sizes. The 20,000-child hub: `For(hub)` reads 20,000 chain pivots, each small; the hub's own chain pivot implies nothing past `MaxPivotEnd` (10), so a child's `For` filters an empty list. `TestCraftedFamily` (4,000 children, cold, under a second) holds the Family read there.
- **Family read.** Cold: one `g.Nodes(want)`, four batches for the people, at most one batch for the named relatives, four for `around` — a fixed number, as `TestPersonFamilyQueryCount` AC-2.2 and AC-2.3 hold. Warm: no query (AC-2.1, AC-2.2). Work per person is bounded by `MaxKinPerRole` (100) per role.
- **Memory.** The two entries go: at 16,000 persons up to 40,230 stored answers. STEP-4 records `BenchmarkCacheSizes`' kinship figure before and after; `docs/catalog-caches.md` gets the new one.

### Edges

- **Assumptions.** A composed answer is exact when each pivot it reads is exact and the nodes it reads are current. An event pivot reads only the event node's links (`deriver.eventLinks`, `EdgeOf` with `neighbor` nil); a chain pivot reads the person's links and the event pivots of their events. Any commit that changes those reloads the event or person as a root, and Apply drops each root and its cached neighbours, so the chain pivot of a person beside a changed event is dropped too. The composition reads nodes through `g.Node`/`g.Cached`, which `OnCommit` reloads. Held by `TestForFollowsCommits` (every subtest, including "an event deleted alone"), `TestSiblingsFollowCommits`, `TestForMatchesDerive` and Verify's pivot `check`.
- **Neighbors.** A commit one hop from a person (their event, their parent): their pivots drop and `For` re-derives them (AC-1.1). Two hops (a grandchild): `For` derives nothing new and is unchanged unless the grandchild's link moved (AC-1.2). A vocabulary commit drops every pivot (DerivedFrom), as today (`TestForFollowsCommits` AC-2.5). One handle that is both an event and, as a parent between, a person: the kind-mismatch path derives without storing, as today (`TestPivotKindMismatch`).
- **Removed inputs.** A deleted or unlinked relative: its node reloads or goes, the pivots at its ends drop, and the next composed answer omits it (`TestForFollowsCommits` AC-2.1, AC-2.4; `TestSiblingsFollowCommits` AC-2.7). Nothing stored can keep it, since no answer is stored.
- **Real data.** None — no rule, seed or table changes; kinship patterns and knobs are untouched.
- **Untrusted size.** Ids from crafted data: `edgeKey` no longer collides for any lengths (AC-3.1). A 20,000-child hub, a person with 50,000 parents: bounded by `MaxPivotEnd` in `Derive` and `Siblings` and by `MaxKinPerRole` in `KinOf`, before `FamilyUnits` (S16-01's caps, unchanged). `named` and `around` are built from capped kin, so the extra `g.Nodes(named)` is at most 400 ids per person.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/canonlinks` `reader.go` | D-1: entries and their helpers go; `For` composes. D-2: `func (r *Reader) Relationships(ids [][]byte, opts Options) (map[string][]kinship.Link, error)`. D-3: `func (d *deriver) siblingsOf(steps []stepped, event func([]byte) ([]kinship.Implied, error)) (map[string][]kinship.Implied, error)`. D-4: `edgeKey` length-prefixed. `LoadRelationships`, `LoadSiblings`, `siblings` go |
| `core/database/canonlinks` `canonlinks.go` | Package comment: `Reader.For` composes one person's implied links from the pivot results; `Reader.Relationships` returns each id's relationship links, recorded and implied, merged through `kinship.LinkSet`, with `Options{Siblings: true}` also the siblings a shared parent implies, loading what it reads in a fixed number of batches. No `LoadSiblings` or `LoadRelationships` |
| `core/database/persondetail` `family.go` | D-5, D-6: `Families` in one pass. `shownKin`, `PersonFamily`, `originSiblings`, `dropUnread` unchanged |
| `core/database/promotealign` `canon.go` | Doc comment at line 56: "implied ones from `canonlinks.Reader.For`, which composes them from the graph cache's kinship pivots for each walked person". No code change |

Errors: none new.

### FFI

None — the Family wire is unchanged.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Measure the base**
Files: `core/database/promotealign/perf_test.go`, `core/database/promotealign/bench_test.go`.
Do: D-8: move the two timings into `largeFixture.warmProposal` and `largeFixture.proposalAfterCommit`, `TestLargeSourceBudgets` calls them; add `BenchmarkLargeSourcePropose`. On the planning base (no other change), run it, `BenchmarkCacheSizes` and `TestLargeSourceBudgets`, and write into the ship log, with `hostname`: `warm-ms`, `after-commit-ms` at both sizes; `kinship-B`, `canonlinks.kinship-keys`, `canonlinks.for-keys`, `canonlinks.siblings-keys` at both sizes; `TestLargeSourceBudgets` pass or fail with its logged figures.
Check: `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'BenchmarkLargeSourcePropose|BenchmarkCacheSizes' -benchtime 1x -timeout 60m ./core/database/promotealign` prints both sizes; `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets -v ./core/database/promotealign` passes.

**STEP-2 — The answers compose from the pivots, through one batched read**
Files: `core/database/canonlinks/reader.go`, `canonlinks.go`, `reader_test.go`; `core/database/persondetail/family.go`, `caps_test.go`; `core/database/promotealign/bench_test.go` (`BenchmarkCacheSizes` "kinship": one `r.Relationships(ids, Options{Siblings: true})`), `promotealign_test.go`.
Do: D-1, D-2, D-3, D-4, D-11, and D-5 / D-6 in `Families`, as one commit: the reader and its callers compile only together (S16-10 landed the same way). Delete every name **Deletes** lists for `reader.go`. Package comment per §6. Test helpers `world.relationships` and `world.siblings` call the batched form and read `[string(id)]`; `TestCraftedFamily` reads `links[string(pat)]`. D-10 for every test it names. Add `TestEdgeKey` (T-5). `Families`' doc keeps its contract (ids order, skipped ids, one Family per repeated id, a fixed number of batched loads cold, no query warm).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks ./core/database/persondetail ./core/database/promotealign` passes, and passes again with `PROVENENCIA_GRAPH_VERIFY=1`; `rg -n "LoadRelationships|LoadSiblings|siblingAnswers|forAnswers|d\.around\(" core` finds nothing; `TestLargeSourceBudgets` passes (the Stop condition's first check).

**STEP-3 — Bar 9's tests**
Files: `core/database/canonlinks/reader_test.go`, `core/database/persondetail/family_query_test.go`.
Do: add `TestComposedAnswersReusePivots` (T-2) and the `TestPersonFamilyQueryCount` subtest "a commit elsewhere" (T-6).
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestComposedAnswersReusePivots|TestPersonFamilyQueryCount' ./core/database/canonlinks ./core/database/persondetail` passes.

**STEP-4 — Measure after**
Files: none changed unless a figure fails the Stop conditions.
Do: run the STEP-1 commands on the same machine (same `hostname`); write the after figures beside the base in the ship log.
Check: `TestLargeSourceBudgets` passes; `warm-ms` at 16,000 is at most 1.25 × its base; `kinship-B` does not grow at either size (§11).

**STEP-5 — Docs**
Files: `docs/catalog-caches.md`, `.agents/skills/add-match-rule/SKILL.md`, `docs/matching-model.md`, `core/database/promotealign/canon.go` (comment), `docs/deployment-plan/spike-16/deployment-plan.md`, `docs/deployment-plan/spike-16/completed.md`, this plan.
Do:
- `catalog-caches.md`: delete the `canonlinks.for` and `canonlinks.siblings` rows; the `canonlinks.kinship` row's size cell becomes the STEP-4 keys and heap figures at 4,000 / 16,000, without the words "with `canonlinks.for` and `canonlinks.siblings`" and its Fill says "`Reader.pivot`, read by `For` and `Relationships`, which compose their answers from it"; § Known exceptions: "The drop-on-any-handle Apply of `promotealign.matchindex` (S16-20 patches it instead) is the registry's known exception to updating exactly."
- `add-match-rule` line 33: replace the text that starts at "`Reader.For`, which answers one person's implied links" and ends at "in one fixed set of batches))" with "`Reader.For`, which composes one person's implied links from one cached result per pivot on the graph cache, and `Reader.Relationships`, which returns each id's recorded and implied links merged through `kinship.LinkSet`, loading what it reads in a fixed number of batches (with `Options{Siblings: true}`, also the siblings a shared parent implies, for the page))". Line 88: "and its own memo slot" → "composed from the pivots like `For`". Line 91: "and how far `Deps` reaches" → "and how far the pivot entry's Apply drops (the roots and their neighbours: one hop)"; "changes `canonlinks`' pivots, `For` and its `Deps` only" → "changes `canonlinks`' pivots, `For` and that Apply only".
- `matching-model.md` line 100: "derived on first read (`canonlinks.Reader.For`)" → "derived on first read and composed into a person's answer at each read (`canonlinks.Reader.For`)".
- `canon.go` comment per §6.
- Map: the rule "Every cache names the signal that updates it, and updates exactly": "the drop-on-any-handle Applies of `for` and `siblings` (S16-16) and of the match index (S16-20) are the known exceptions" → "the match index's drop-on-any-handle Apply (S16-20) is the known exception"; the S16-16 row: status, and its scope cell drops "`DropOnHandleCommit` and `dropHandleMemos`" (S16-10 already removed them).
Check: the DoD sweep (`rg` of AC-3.3) finds nothing; `CGO_ENABLED=1 go test -tags fts5 -run 'TestCacheDocListsEveryEntry|TestCacheDocSeesEveryOwner' ./core/catalogsession` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/canonlinks/reader_test.go` `TestForFollowsCommits`, `TestForMatchesDerive`, `TestFor`, `TestPivotKindMismatch`, `TestVerifyCatchesAStaleResult` (unchanged) | Their cases → today's answers | AC-1.3, D-1 |
| T-2 | Go | `reader_test.go` `TestComposedAnswersReusePivots` *(new)*, table of two cases on the AC-1.1 world (Ann + Bob married M1; B1: Bea of Ann and Bob; B2: Cal of Bea): read `For` of every person; record `held` = ids with `pivots.Peek`; commit `ann` → right after the commit, of the ids held before exactly the expected set is no longer held — `{ann} ∪ every l.Neighbor of Ann's cached node` (in this world M1 and B1, Ann's events: what `deps` names), computed by the test from the node and also asserted equal to {Ann, M1, B1} — and every other id held before still is; after reading `For` of every person, `derived` (held now and not before) equals that expected set intersected with the ids held before; commit `cal` → `For(ann)` derives nothing and equals its earlier value | AC-1.1, AC-1.2, D-7 |
| T-3 | Go | `reader_test.go` `TestSiblingsFollowCommits` (rewritten per D-10) | AC-2.3: a new child of Ann → Bea's sibling Cal; AC-2.4: commit Ann → Bea's siblings equal a fresh world's; AC-2.5: vocabulary commit → same; AC-2.7: Ann and her links removed → none, Verify passes | AC-1.3, D-10 |
| T-4 | Go | `reader_test.go` `TestRelationshipsQueryCount` *(new name; merges `TestLoadSiblingsQueryCount` and `TestLoadRelationshipsQueryCount`; their direct single-id `Relationships` calls take `[][]byte{id}`)* | cold `Relationships([parent, first], {Siblings: true})` on 1 child, 7 children with births, three families → equal non-zero query counts; the same call again → 0 queries and equal result | AC-2.1, D-2 |
| T-5 | Go | `reader_test.go` `TestEdgeKey` *(new)*, table | `{A:"ab",B:"c"}` vs `{A:"a",B:"bc"}` → differ; associations `["xy","z"]` vs `["x","yz"]` → differ; an edge vs itself → equal; `{A:"a",B:""}` vs `{A:"",B:"a"}` → differ | AC-3.1, D-4 |
| T-6 | Go | `core/database/persondetail/family_query_test.go` `TestPersonFamilyQueryCount` subtest "a commit elsewhere" *(new)* | `shape("", 2, 2, true)` plus a stranger person; `PersonFamily(pat)`; `w.k.Commit(stranger)`; count queries on `PersonFamily(pat)` → 0, and `Family`, origin siblings and summaries equal the first read | AC-2.2, D-5 |
| T-7 | Go | `TestPersonFamily`, `TestPersonFamilyOriginSiblings`, `TestPersonFamilyQueryCount` AC-2.1 – AC-2.3, `TestCraftedFamily`, `TestRelationships`, `TestRelationshipsOptionsSiblings`, `TestSiblingsLeaveMatchingAlone`, `core/kinship` tests, `core/graphalign` tests, `TestLargeSourceBudgets` (unchanged but for the call signature) | → pass | AC-2.3, D-5, D-6 |
| T-8 | Go | `core/catalogsession/cache_doc_test.go` `TestCacheDocListsEveryEntry`, `TestCacheDocSeesEveryOwner` | the doc without the two rows → pass | AC-3.2 |
| T-10 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeIgnoresPageSiblings` *(renamed from `TestProposeIgnoresStoredSiblings`)* | Bea's siblings read through Ann → the proposal equals the one before | AC-2.3, D-10 |
| T-9 | Go | `core/database/promotealign/bench_test.go` `BenchmarkLargeSourcePropose` *(new)* | 4,000 and 16,000 persons → `warm-ms`, `after-commit-ms` reported; base and after in the ship log | AC-1.4, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs, and the Go suites with `PROVENENCIA_GRAPH_VERIFY=1`. Most likely to break: `core/database/persondetail` (Family order and query counts), `core/database/promotealign` (`TestLargeSourceBudgets` warm), `core/catalogsession` (cache doc).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] The Go suites of canonlinks, persondetail and promotealign pass with PROVENENCIA_GRAPH_VERIFY=1
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (AC-3.3: LoadRelationships, LoadSiblings, forAnswers, siblingAnswers, canonlinks.for, canonlinks.siblings, "memo slot"); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] Ship log: STEP-1 base and STEP-4 after figures (BenchmarkLargeSourcePropose, BenchmarkCacheSizes kinship) on one named machine
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the reworded rule; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/catalog-caches.md, docs/matching-model.md (STEP-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a filed family (a person with both parents from a baptism, siblings, a marriage and children). Open the person's Family section. See the same parents, siblings, partner and children as on `main`.
2. Edit one of the person's Observations in a Source and file it. Return to the Family section: the same family shows, at once.
3. Open a Source with unfiled people and run Promote: the suggestions are the same as on `main` for the same project.
4. Run the AC-3.3 `rg` from the repo root: no hit.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A warm Promote proposal gets slower, since `For` is composed for every walked person instead of read from a stored answer | STEP-1 / STEP-4 measure it at both sizes; S16-22 removes the walk; the Stop condition below sends a slowdown to the user instead of a quiet memo |
| The Family read loses a relative because `shownKin` reads a node that was not loaded | D-5 step (3) loads every named relative before `shownKin`; T-7 holds today's units |
| `Relationships` returning a map changes order somewhere | Callers iterate their own id lists, not the map (D-5); `TestPersonFamily` checks order |

**Stop and report if:**

- `Reader.For`, `Reader.Relationships`, `LoadRelationships`, `forLinks`, `siblingsOf`, `loadSteps` or `Families` no longer do what §4 says.
- `TestLargeSourceBudgets` fails after STEP-2 where it passed at STEP-1 on the same machine, or `warm-ms` at 16,000 is more than 1.25 × its STEP-1 base: report both figures; memoising `For` is the user's call (rule "No stored answer without a measurement").
- `BenchmarkCacheSizes`' kinship heap figure grows at either size.
- A test in T-7 needs its expected values changed to pass.
- The existing `TestPersonFamilyQueryCount` subtests named "AC-2.2" or "AC-2.3" (S16-01's labels, not this plan's ACs) give unequal counts after STEP-2.
- Verify (`PROVENENCIA_GRAPH_VERIFY=1`) fails on the `canonlinks.kinship` Check after any step.

## 12. Open questions

None.

## 13. PR

**Title:** S16-16 — Kinship answers are composed from the pivots, and the Family read is one pass
**Body:** objective in one sentence; what shipped against each US; AC evidence; the STEP-1 / STEP-4 figures; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-16-composed-kinship.md`.
