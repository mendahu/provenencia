# S16-10 — One cache registry

**Status:** Planned
**Spike / slice:** Spike 16 · Slice 2a — One home for caches
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine refactor; no view, wire or catalog changes)
**Board:** —
**Depends on:** S14-42 (Spike 14, landed [#748](https://github.com/mendahu/provenencia/pull/748))
**Blocks:** S16-12, S16-15, S16-16, S16-18, S16-20 (each registers its cache as an entry and adds its row to `docs/catalog-caches.md` *(new)*)
**Branch:** `s16-10-one-cache-registry`
**Version bump:** none — a preparatory refactor with no change a researcher can see, no wire and no catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (only for the match index's home in `promotealign`; no knob changes)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [design § One cache registry](../cache-architecture.md#one-cache-registry-two-ways-to-fill-s16-10); working rules: "(Batch 2) The point of this batch is simpler code", "(Batch 2) One cache registry, two ways to fill", "(Batch 2) Every cache names the signal that updates it, and updates exactly", "(Batch 2) Apply states its fan-out", "(Batch 2) A write's one-shot graph fills only what it reads", "(Batch 2) Memory is budgeted", "(Batch 2) Warm only what every proposal needs; no background re-warm"; goal bar 15; [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §7–§8
**Planned:** 2026-11-11 against `main` @ `6c8736d`
**Review:** [`S16-10-one-cache-registry.review.md`](S16-10-one-cache-registry.review.md) — Fits with changes — Ready with changes, 2 rounds; no open blocker; To user: answered (write path, typed get-or-fill)

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

`graphcache` keeps five kinds of derived cache, each with its own slot and its own drop rule: three `NodeMemo` slots with flags (`DropOnVocabulary`, `DropOnHandleCommit`), a match index field dropped by hand in `OnCommit`, and an alias index and name format whose setters drop the match index and every Evidence-graph memo as a side effect (`dropDerived`) and raise a `warmNeeded` flag that `catalogsession.Do` answers with a re-warm. This PR gives them one registry with two typed kinds of entry, keyed (one value per node id) and whole (one value). An entry declares what a commit does to it (Apply, on the roots `OnCommit` reloaded), what it is derived from (DerivedFrom), how long it lives (the graph or the catalog connection), how Verify checks it, and, for keyed entries, its cap; a whole entry may Build itself on first read. The five move onto it with today's results; the flags, `dropDerived`, `warmNeeded` and the dead `canonicalgraph` package go, and `docs/catalog-caches.md` *(new)* lists every Go and Swift cache with its signal and its size at 4,000 and 16,000 persons, held to the registry by a test.

**Why now.** Every later Batch 2 Go PR (S16-12, S16-15, S16-16, S16-18, S16-20) adds or reshapes a cache. Without one registry each would add a sixth slot with a seventh drop rule.

**What stays.** Every read returns what it returns today: the same Promote suggestions, Family, summaries and search rows. The match index still drops on every handle commit (until S16-20); the `for` and `siblings` answers still drop on every handle commit (until S16-16). `Activity`, the summary displays, the labels, the vocabulary, the node maps and the Evidence-graph store (with its per-graph memos, S16-13's) stay typed fields of `Graph`. No SQL, wire or Swift code changes.

## 2. User stories

**US-1.** As a contributor adding a cache in Batch 2, I register one entry that states its commit rule, what it is derived from, its lifetime, its check and its cap, so that I write no new drop path in `OnCommit` and no new flag.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A test registers whole entry A (catalog lifetime), whole entry B derived from A, keyed entry C derived from B, keyed entry D derived from the vocabulary, keyed entry E derived from nothing; all hold a value; `A.Set(g, v)` | B and C are empty; A holds `v`; D and E keep their values |
| AC-1.2 | Same entries filled; a commit with `Lists: [ListPropertyVocabulary]` | D is empty; A, B, C, E keep their values. A commit with `Lists: [ListSourceVocabulary]` empties nothing |
| AC-1.3 | Same entries filled; `Drop()` | B, C, D, E are empty; A keeps its value |
| AC-1.4 | Keyed entry with `Cap: 2` holding keys k1, k2; `Set(k3)`; then `Set(k1, v2)` | k3 is not stored; k1 now holds `v2` |
| AC-1.5 | Whole entry W with a Build that counts its calls, and whole entry R derived from W holding a value; `W.Get(g)` twice on the session graph; then `W.Set(g, v)` | Build ran once; R keeps its value after the Build (a first fill drops no reader) and is empty after `Set`; a Build error returns the error and stores nothing |
| AC-1.6 | Keyed entry with an Apply that records its arguments; a participation association commit; a merge commit; a delete of a loaded node; a Source-only commit (`SourceScopes`, no `Handles`); a first load | Apply gets the association's ends, the merged handle's neighbours, the deleted node's link neighbours (today's `dropMemos` roots), and on the Source-only commit is called with no roots and the commit's `effects.Set`; a first load calls nothing |
| AC-1.7 | `PROVENENCIA_GRAPH_VERIFY=1`; a keyed and a whole entry each hold a value their Check rejects; any commit | `OnCommit` returns the Check's error |
| AC-1.8 | `RegisterKeyed` with no Cap, or with no Apply; a `DerivedFrom` naming an unregistered slot; a catalog-lifetime whole entry derived from a graph-lifetime entry or the vocabulary; two entries with one Name | Each panics at registration |
| AC-1.10 | On `graphcache.Over(tx)`: `Get` of a catalog-lifetime whole entry with a Build; `Get` of a graph-lifetime whole entry with a Build; `Set` of the graph-lifetime entry | The first builds once through the transaction; the second does not call its Build and returns `ok == false`; the `Set` stores nothing |
| AC-1.9 | `docs/catalog-caches.md` *(new)* lacks the row for a registered entry, or names an entry nobody registered | `TestCacheDocListsEveryEntry` fails naming it |

**US-2.** As the engine, after this PR every moved cache reads and drops as it does today.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | An open project with the alias index, name format and match index warm, kinship pivots stored; `nameformats.Default.Set(g, …)` | The match index is dropped; the alias index and every kinship pivot stay (the map's Check "setting the name format drops exactly the entries that Read it") |
| AC-2.2 | An open project after `Do` returns from a vocabulary commit, a `Drop`, a handle commit, a `ListSourceVocabulary` commit, and a failed `fn` that called `Drop` | `valuerules.Load(g)` issues 0 SELECTs in each case (the alias index and name format survive); the match index is absent after the vocabulary commit, the `Drop`, the handle commit and the failed `fn`, and present after the Source vocabulary commit |
| AC-2.3 | A participation commit naming an event | The event's and its neighbours' kinship pivots are dropped, an unrelated person's pivot stays; every `for` and `siblings` answer is dropped |
| AC-2.4 | A Source-only commit, then a handle commit | The match index stays after the Source-only commit and is gone after the handle commit |
| AC-2.5 | An Evidence graph's layer memo stored; `namealiases.Index.Set(g, …)`; then a vocabulary commit | The memo stays after the alias index change and is gone after the vocabulary commit |
| AC-2.6 | `PROVENENCIA_GRAPH_VERIFY=1`; the stored alias index, name format or match index replaced by one that differs from a fresh build | `Verify` returns an error naming the entry |
| AC-2.7 | `TestLargeSourceBudgets`, the canonlinks, promotealign, conclusionsummaries, searchindex and catalogsession suites | Pass unchanged in their expectations (only API renames in the tests) |

**US-3.** As a contributor, I read one table that names every cache, so that I know where a value lives, what updates it and what it costs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `docs/catalog-caches.md` *(new)* | Lists every registry entry (by its `Name`) and every typed `Graph` field, the held catalog sessions, and the Swift caches, with key, fill, updating signal, lifetime, cap, and the Go sizes at 4,000 and 16,000 persons from `BenchmarkCacheSizes`; says why `Activity`, displays, labels and the Evidence-graph store stay typed fields |
| AC-3.2 | `rg -n 'canonicalgraph\|RegisterNodeMemo\|NodeMemo\b\|MemoSlot\|DropOnVocabulary\|DropOnHandleCommit\|dropHandleMemos\|dropDerived\|WarmNeeded\|MarkWarm\|SetAliasIndex\|SetNameFormat\|SetMatchIndex' core api .agents .cursor docs/*.md` | No hit |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache` registry: typed handles `Keyed[T]` (`RegisterKeyed[T]`, get-or-fill `Get`, batch `GetMany`) and `Whole[T]` (`RegisterWhole[T]`, `Get` with an optional Build), Apply after `OnCommit`'s reload, DerivedFrom with transitive drops, lifetime, Check in Verify, Cap, `EntryNames`, `KeyCounts` | STEP-1, AC-1.1 – AC-1.10 |
| `RegisterNodeMemo`, `NodeMemo`, `MemoSlot`, `MemoCheck`, `MemoDeps`, `dropMemos`, `dropVocabularyMemos`, `dropHandleMemos`, `verifyMemos` fold into it; `memo.go` deleted | STEP-1, STEP-2 |
| The three `canonlinks` slots as keyed entries; `deps` becomes the kinship entry's Apply; the `for` and `siblings` drop-on-any-handle becomes their Apply | STEP-2, AC-2.3 |
| Alias index and name format as catalog-lifetime whole entries with a Build, owned by `namealiases` and `nameformats`; the cold-graph fallbacks and the `q` parameters of `namealiases.Rules`, `nameformats.LoadFor`, `valuerules.Load`, `promotealign.loadConfig` go | STEP-3, AC-2.2, AC-1.5 |
| `warmNeeded`, `WarmNeeded`, `MarkWarm` and `catalogsession`'s re-warm after `fn` go; `WarmOnOpen` stays | STEP-3, AC-2.2 |
| The match index as a graph-lifetime whole entry owned by `promotealign`, derived from the vocabulary, the alias index and the name format; its handle-commit drop as its Apply; a Check | STEP-4, AC-2.1, AC-2.4, AC-2.6 |
| `dropDerived` goes: Evidence-graph memos drop on a vocabulary commit only, listed in the doc as the store's one remaining hand-written drop, owned by S16-13 | STEP-4, STEP-8, AC-2.5 |
| Checks for the alias index and the name format | STEP-3, AC-2.6 |
| `canonicalgraph` deleted; `TestInputsMatchIsolatedComposition` keeps an inline association query | STEP-5, AC-2.7 |
| `BenchmarkCacheSizes` reports each cache's heap at 4,000 and 16,000 persons | STEP-6, AC-3.1 |
| `docs/catalog-caches.md` *(new)* and `TestCacheDocListsEveryEntry` *(new)* | STEP-7, AC-3.1, AC-1.9 |
| Docs and guides this makes false; a new Known trap in the maintainability lens, "a cache that empties on any write"; the map's S16-13, S16-16 and S16-20 scope entries and the registry rule reworded to the registry's names | STEP-8, AC-3.2 |

| Out | Home |
| --- | --- |
| Any new cache (Source records, outcome records, graph catalog, place entries) | S16-11, S16-12, S16-15, S16-18 |
| The `for` and `siblings` entries and their drop-on-any-handle Apply | S16-16 |
| Dropping the match index on every handle commit; `indexGen`; the profiler's shared map | S16-20 |
| Evidence-graph memos following the vocabulary through the registry instead of `OnCommit`'s loop | S16-13 (the records tier) |
| Exporting `nodeKey` and `followKeys` | S16-18 |
| Moving the large fixture into an importable `promotealigntest` package | not planned until a PR outside `promotealign` needs it at Scale (review M-10) |
| Callers for the dead `namealiases.Invalidate` and `nameformats.Invalidate` (no writer changes alias tables or the default name format in-app) | not planned |
| Any Swift change; Swift caches are only listed in the doc | S16-24 – S16-29 |
| A speed budget for any cache | S16-11 – S16-32 (rule "Measure growth") |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Per-node memos | `core/database/graphcache/memo.go` `NodeMemo`, `RegisterNodeMemo`, `Memo`, `SetMemo`, `Cached`, `dropMemos`, `dropVocabularyMemos`, `dropHandleMemos`, `verifyMemos` | Per-node results with `Deps`, `Check` and two flags; a commit drops every slot's roots, `Deps` widens it; the flags are the drop rules this PR replaces |
| Graph fields | `core/database/graphcache/graph.go` `Graph.matchIndex`, `matchIndexOK`, `warmNeeded`, `memos`, `aliasIndex`, `aliasIndexOK`, `nameFormat`, `nameFormatOK` | Five hand-kept slots. `MatchIndex`, `SetMatchIndex`, `ResetMatchIndexForTest`, `AliasIndex`, `SetAliasIndex`, `NameFormat`, `SetNameFormat`, `WarmNeeded`, `MarkWarm`, `dropMatchIndex`, `dropDerived` |
| `dropDerived` | `graph.go` `dropDerived` | Empties every Evidence-graph memo and the match index and sets `warmNeeded`; called by `clear`, `SetAliasIndex`, `SetNameFormat` and the vocabulary branch of `OnCommit` |
| `clear` / `Drop` | `graph.go` `clear` | Empties nodes, kinds, displays, memos, Evidence graphs, labels, activity, vocab, and through `dropDerived` the match index; leaves the alias index and name format |
| Commit | `core/database/graphcache/commit.go` `OnCommit` | Vocabulary list → `DropLabels`, vocab reset, `dropDerived`, `dropVocabularyMemos`. Structure → `Drop`. Then `dropEvidenceGraphs`; returns early when `set.Handles` is empty; else `dropMatchIndex` and `dropHandleMemos` before the reload, and `dropMemos(roots)` after it, roots built by `detachAssociation`, `mergedNeighbors` and removed nodes' links |
| Commit error | `core/database/listen.go` `Catalog.Notify` | The first `OnCommit` error calls `Drop()` on every listener |
| Verify | `core/database/graphcache/verify.go` `Verify` | Nodes, kinds, displays, `verifyMemos`, Evidence graphs. The match index, alias index and name format are not checked |
| Evidence-graph memos | `core/database/graphcache/evidence_graph.go` `loadedEvidenceGraph.memos`, `EvidenceGraphMemo`, `SetEvidenceGraphMemo` | Dropped with the graph, by `Drop`, and by `dropDerived` |
| Evidence-graph memo users | `core/database/evidencelayer/evidencelayer.go` `Load` (key `memoKey + namePattern`; its comment says the memo drops on "new alias tables"); `core/database/promotealign/canon.go` `loadCandidatesMemo` (key `candidatesMemoKey + namePattern`, valid for one `matchIndex.gen`) | The layer reads the vocabulary, terms and the name pattern key, not the alias index; the candidates memo ignores itself when the match index generation changes |
| Kinship slots | `core/database/canonlinks/reader.go` `kinshipSlot`, `forSlot`, `siblingSlot`, `init`, `deps`, `check`, `checkFor`, `checkSiblings`, `Reader.pivot`, `Reader.answer`, `Reader.For` | Three `NodeMemo`s registered in `init`; `kinship` with `Deps` (roots and their one-hop neighbours) and `DropOnVocabulary`; `for` and `siblings` with `DropOnVocabulary` and `DropOnHandleCommit` |
| Match index | `core/database/promotealign/index.go` `matchIndex`, `matchIndexOf`, `loadShortlist`, `Warm`, `indexGen`; `profiles.go` `newProfiler` (shares the stored `profiles` map when `profilesPattern` matches), `profiler.keep` | Stored with `g.SetMatchIndex`; `keep` re-sets it to hold profiles. No Check |
| Alias index | `core/database/namealiases/namealiases.go` `Warm` (returns `database.ErrClosed` on a nil graph), `Invalidate`, `Rules(g, q, base)` | `Warm` loads tables and calls `SetAliasIndex`; `Rules` reads the warm index or, on a cold graph, loads from `q` and builds without storing. `Invalidate` has no caller. Alias tables report no effects (`core/database/effects/registry.go`, `None: true`) |
| Name format | `core/database/nameformats/nameformats.go` `Warm`, `Invalidate`, `LoadFor(g, q)`, `Load(q)` | Same shape as the alias index. `Invalidate` has no caller; `name_format_profile*` tables report no effects |
| Value rules | `core/database/valuerules/valuerules.go` `Querier`, `Load(g, q)` | Calls `nameformats.LoadFor(g, q)` and `namealiases.Rules(g, q, …)`; "g may be nil". `Querier` has no other use |
| Callers of the readers | `core/database/promotealign/promotealign.go` `loadConfig(g, q)` (callers `index.go` `Warm`, `promotealign.go` `loadInputs`); `core/database/persondetail/recordlinks.go:84`; `core/database/conclusionsummaries/compose.go:388` `nameformats.LoadFor(g, g)` (runs on a write's `Over` graph through `searchindex`); `api/ffi/handlers/nameformat.go` `GetNameFormat` | Pass a graph and a querier that is the graph's own connection or `db` |
| Session warm | `core/catalogsession/session.go` `warmGraph`, `WarmOnOpen`, `runLocked` | `warmGraph` runs `namealiases.Warm`, `nameformats.Warm`, `MarkWarm`; `runLocked` runs `warmGraph` again when `fn` succeeded and `WarmNeeded()` |
| One-shot graphs | `core/database/graphcache/graph.go` `Over(q)` (sets `alt`); callers `core/database/searchindex/handles.go` `newGraph`, `core/database/searchindex/inputs.go:115`, `core/database/autoreconciler/autoreconciler.go:318` | A write composes on a throwaway graph over its transaction |
| `canonicalgraph` | `core/database/canonicalgraph/canonicalgraph.go`, `canonicalgraph_test.go`, `export_test.go` | `Walk`, `Targets`. Only caller: `core/database/conclusionsummaries/inputs_test.go` `TestInputsMatchIsolatedComposition` (`assoc` helper at line 163, called with `hops.PlacesOfEvent` and three `hops.ParentsOfPlace`) |
| Tests | `core/database/graphcache/graph_test.go` `TestDropLeavesAliasIndex`, `TestDropLeavesNameFormat`, `TestMatchIndexDrops`, `TestNodeMemoDropsOnAnyCommit`, `TestNodeMemoDropsOnCommit`, `TestWarmNeeded`, the `memoNear`/`memoVocab`/`memoAll` vars; `evidence_graph_memo_test.go` `TestEvidenceGraphMemoDroppedWithSource`; `core/catalogsession/warm_test.go` `TestDoWarmsNameFormat`, `TestDoWarmsMatching`, `TestDoRewarms`, `TestOpenSurvivesABadValue`; `core/catalogsession/alias_cache_test.go` `TestDoKeepsAliasTablesForTheOpenProject` (lines 34, 57); `core/database/promotealign/promotealign_test.go` (line 90), `memo_test.go` `TestProfilesLiveWithTheMatchIndex` and line 175, `export_test.go` `MatchIndexGenForTest` and line 34; `core/database/valuerules/valuerules_test.go` (`Load(nil, db)` at 44, 69); `core/database/conclusionsummaries/conclusionsummaries_test.go:865`; `core/database/evidencelayer/evidencelayertest/evidencelayertest.go:310`; `core/database/canonlinks/reader_test.go` (`hasMemo`, `SetMemo` at 643 and 1124) | Hold today's drop rules; rewritten to the registry's API with the same expectations except where a Decision says otherwise |
| Large fixture | `core/database/promotealign/perf_test.go` `newLargeFixture(t *testing.T)`, `TestLargeSourceBudgets` (subtests "cold", "compose every summary"); `bench_test.go` `largeShape(r, names, treeRoot)` (4,000 persons, 1 city of 6,000 events, 50 towns sharing 2,000 events, a 12-generation tree); `promotealign_test.go` `newFixture(t testing.TB)`, `must` (`evidencelayertest.Must`) | No memory measurement per cache; persons fixed at 4,000 |
| Doc parity precedent | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry` | Holds `docs/matching-model.md` to the registry both ways |
| Guides | `.cursor/rules/catalog-queries.mdc` lines 17–18; `.agents/skills/use-catalog-session/SKILL.md` line 62; `docs/promote-graph-alignment.md` lines 158, 165–169; `.agents/skills/review-pr/lenses/maintainability.md` Known traps (no cache trap yet) | Name `canonicalgraph.Walk`, `NodeMemo`/`RegisterNodeMemo`, `WarmNeeded`, `SetAliasIndex`, `SetNameFormat` and "S9-28 caches it as canonical edges" |

**Precedent.** The registration-from-the-owner pattern is today's `RegisterNodeMemo` from `canonlinks`' `init` and `graphcache.SetTruth` / `SetDisplayCheck` / `SetEvidenceGraphLoader`: `graphcache` never imports the owner. Copy that, registering in `init()`. Do not copy `NodeMemo`'s flags: every rule that was a flag is now either a DerivedFrom or the owner's Apply. The doc test copies `TestKnobInventoryMatchesRegistry`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The registry lives in `core/database/graphcache/cache.go` *(new)*, which replaces `memo.go`. An owner registers a typed handle and reads and writes only through it: `RegisterKeyed[T]`, which takes a `KeyedEntry[T]` and returns a `Keyed[T]` (one value of type `T` per 16-byte id), and `RegisterWhole[T]`, which takes a `WholeEntry[T]` and returns a `Whole[T]` (one value). `DerivedFrom` takes `[]graphcache.Source`, an interface every `Whole[T]` satisfies, plus `graphcache.Vocabulary` *(new, a `Source` naming the typed vocabulary)*. `Lifetime` (`LifeGraph`, the zero value; `LifeCatalog`). Owners register in `init()` (user, 2026-11-11: typed get-or-fill) | Each owner's read is one call with no `any` cast; the handle's type says which methods apply, so kind misuse is a compile error (review M-3); the registry is the one place that holds the cap, the write-path rule and the drops | Untyped slots with the owner reading, casting, computing and storing by hand — that four-step block would repeat in every Batch 2 owner (user's choice, 2026-11-11); one `Entry` with a `PerKey` flag — runtime panics |
| D-2 | Fill. Keyed: `k.Get(g, id, fill func() (T, error)) (T, error)` returns the held value, else calls `fill`, stores its result under the cap and returns it; `k.GetMany(g, ids, fill func(missing [][]byte) (map[string]T, error)) (map[string]T, error)` calls `fill` once with every id not held, so an owner loads a frontier in one batch (S16-18); a `fill` error is returned and nothing is stored. The fill is a closure at the call site, so it uses the reader's own state (the deriver, the profiler) with no registry parameter. Whole: `w.Get(g) (T, bool, error)` returns the held value, else runs the entry's `Build` (if any), stores and returns it; with no Build it returns `ok == false` and the owner stores with `w.Set(g, v)`; a first fill by Build drops no reader | The map rule's "fill per key" with the fill written where its inputs are; batch fills fit S16-18's frontier loads; one Build declared once serves every reader of a whole value | A `Fill func(g, id)` registered once — cannot see the reader's state and fills one key per call, which would turn S16-18's one narrow load into thousands (user, 2026-11-11) |
| D-3 | `Apply` runs for every entry holding data, in registration order, once per commit after the reload, on every commit that is not `Structure` (including one with no `Handles`, then with nil roots). Its first error stops the loop and is `OnCommit`'s error; `Catalog.Notify` then drops the graph, as for any `OnCommit` error today. A keyed entry must have an Apply (`RegisterKeyed` panics without one); a whole entry's nil Apply means commits leave it. Roots are exactly today's `dropMemos` roots. The match index's and `for`/`siblings`' drops move from before the reload to their Apply after it; an error between the two drops the graph through `Notify` | One place a commit reaches a cache; a keyed cache cannot forget its commit rule (review M-2); later entries (Source records, graph catalog) update on `SourceScopes` and `Set.Reads` with no handle | Keeping today's early return when `Handles` is empty — S16-12 and S16-15 would each add a branch; a default "drop the roots" — a second, implicit rule beside Apply |
| D-4 | `DerivedFrom []Source` lists `graphcache.Vocabulary` and earlier-registered whole handles, on both kinds. `w.Set` and `w.Drop` of a whole handle empty every entry derived from it, then entries derived from those, transitively; a vocabulary commit does the same from `graphcache.Vocabulary`. Keyed changes drop no reader | Replaces `DropOnVocabulary` and `dropDerived` with a declaration; key-level changes reach other caches through each one's own Apply, which sees the same roots. Named `DerivedFrom` because `effects.Set.Reads` already means "partial reads a batch moved" on the same commit path (review M-4) | Key-level propagation — no S16-10 entry derives from a keyed entry; `Reads` or `Inputs` as the field name — taken by `effects` and `InputKey` |
| D-5 | `Drop()` empties every `LifeGraph` entry and leaves `LifeCatalog` entries. `RegisterWhole` panics when a `LifeCatalog` entry derives from a `LifeGraph` entry or `graphcache.Vocabulary`. Keyed entries are always `LifeGraph` | The alias index and name format survive `Drop` today; a catalog-lifetime entry derived from graph data would go stale on `Drop` | A per-entry "survives Drop" flag — the lifetime says it |
| D-6 | `KeyedEntry.Cap int` is required (> 0): storing a key not held (through `Get`, `GetMany` or `Set`) when the entry holds `Cap` keys stores nothing and still returns the filled value; a held key is replaced. A whole entry has no Cap; its size is bounded by what it is built from and recorded in the doc. Caps: `canonlinks.kinship`, `canonlinks.for`, `canonlinks.siblings` 65,536 each | Bounds a crafted catalog with no eviction code; results do not move (a key not stored is filled again on read). `BenchmarkCacheSizes` reports each entry's key count at 16,000 persons; a count over half the cap is a Stop condition | LRU eviction — nothing needs it at Scale; a Cap on whole entries — a key count means nothing for one value |
| D-7 | `Check func(g *Graph, key []byte, stored any) error` on both kinds (key nil for a whole entry); Verify runs it over every stored value. Nil Check: Verify skips. Verify never runs a Build | Today's `MemoCheck` shape, extended to whole entries | A separate whole-entry check type |
| D-8 | Handle methods (a nil graph returns the zero value and no error, as today). `Keyed[T]`: `Get`, `GetMany` (D-2), `Peek(g, id) (T, bool)` (no fill), `Set(g, id, v)` (an id that is not 16 bytes stores nothing, as today), `Drop(g, ids [][]byte)`, `Clear(g)`. `Whole[T]`: `Get` (D-2), `Set(g, v)`, `Drop(g)`. Entry signatures are typed: `KeyedEntry[T].Apply func(g *Graph, k Keyed[T], roots [][]byte, set effects.Set) error` (required), `Check func(g *Graph, id []byte, stored T) error`; `WholeEntry[T].Build func(g *Graph) (T, error)`, `Apply func(g *Graph, w Whole[T], roots [][]byte, set effects.Set) error`, `Check func(g *Graph, stored T) error`. `g.KeyCounts() map[string]int` (held keys per keyed entry `Name`, review M-13). None of the drops returns an error. `Cached` moves to `graph.go` beside `Node` (review M-11) | The handle carries the type, so no owner casts `any` and no Check asserts a type | Graph methods taking a slot number — every caller casts the `any` it gets back |
| D-9 | The three `canonlinks` slots become keyed handles registered in `reader.go`'s `init`: `pivots Keyed[pivotResult]` (DerivedFrom `graphcache.Vocabulary`, Apply drops the roots and `deps(g, roots)` through `k.Drop`, Check `check`, Cap 65,536), `forAnswers` and `siblingAnswers Keyed[personAnswer]` (DerivedFrom `graphcache.Vocabulary`, Apply `k.Clear(g)` when `len(roots) > 0`, Checks `checkFor`, `checkSiblings`, Cap 65,536). `Reader.pivot` and `Reader.For` become `Get` with the derive closure; `siblingsOf`'s batch store becomes `GetMany`; `Reader.answer` goes. `NewReader(g)` drops its `rules` parameter and reads `matchregistry.Default().Kinship` itself, since every caller passes that (`core/database/persondetail/family.go`, `core/database/promotealign/canon.go`, `caps_test.go`, `promotealign_test.go`, `reader_test.go`'s `shipped`); the stored values keep `rules` (and `pivotResult.kind`) only for Check, and the read no longer compares them | Today's behavior with one-call reads. The drop-on-any-handle rule lives in the owner until S16-16 deletes those two entries, so `graphcache` carries no `DropOnHandleCommit` mechanism meanwhile | Keeping the rules comparison — a validity callback on every `Get` for a case no caller has; keeping a `DropOnHandleCommit` field until S16-16 — an old mechanism beside its replacement |
| D-10 | The alias index is `namealiases.Index Whole[valuecompare.AliasIndex]` *(new)*, the name format `nameformats.Default Whole[Loaded]` *(new)*: `LifeCatalog`, no DerivedFrom, no Apply, registered in each package's `init`; Build = today's `Load(g)` plus `buildIndex` / the `Loaded` value. `namealiases.Rules(g, base)`, `nameformats.LoadFor(g)` and `valuerules.Load(g)` call `Get`; their `q` parameters and type assertions go; a nil graph returns `database.ErrClosed`, as `Warm` does today. `Warm(c *database.Catalog) error` stays: `database.ErrClosed` on a nil graph, else `Get`'s error. `Invalidate(c) error` becomes `Drop` and returns `database.ErrClosed` on a nil graph, else nil; the next read rebuilds. `valuerules.Querier` goes. On a write's `Over` graph they Build once from its transaction and live with that graph (D-21) | One path for the session graph and a write's graph | Keeping `q` — a second way to the same value; a new `errNoGraph` (review M-8) — `database.ErrClosed` already means it here; dropping `Warm` from `WarmOnOpen` — an alias load error would stop failing the open, since `promotealign.Warm`'s error is ignored |
| D-11 | Neither the alias index nor the name format derives from the vocabulary. `warmNeeded`, `WarmNeeded`, `MarkWarm` and `runLocked`'s re-warm go; `warmGraph` keeps the two `Warm` calls at open, then `WarmOnOpen` runs `promotealign.Warm` as today | Today's re-warm after a vocabulary commit or `Drop` rebuilds the same values: alias tables and name formats read their own tables, report no effects, and survive `Drop`. Nothing is missing on the session graph after open | Keeping the re-warm — work with no effect, and the flag couples five caches |
| D-12 | The match index is `promotealign.MatchIndex Whole[matchIndex]` *(new, exported for DerivedFrom and tests)*, registered in an `init()` in `index.go` (a package-level `var` initializer would make an initialization cycle through `checkMatchIndex` → `newProfiler` → `matchIndexOf`): `LifeGraph`, no Build (its build needs the profiler's shared map, which S16-20 removes), DerivedFrom `graphcache.Vocabulary`, `namealiases.Index`, `nameformats.Default`; Apply `w.Drop(g)` when `len(roots) > 0`; Check `checkMatchIndex`. `matchIndexOf` calls `MatchIndex.Get(g)`, `loadShortlist` and `keep` store with `MatchIndex.Set(g, m)`. `ResetMatchIndexForTest` goes; tests call `promotealign.MatchIndex.Drop(g)` | Today's drops (handle commit, vocabulary, alias index, name format, `Drop`) as declarations | A Build for it now — S16-20 gives it one when the shared map goes; moving `indexGen` or profiles — S16-20 |
| D-13 | `checkMatchIndex(g, nil, stored)`: `cfg, namePattern, err := loadConfig(g)`; `metas, err := valuerules.Metas(g)`; either error is returned as the check's error. `newProfiler`'s body moves to `newProfilerWith(g, metas, rules, namePattern, done map[string]profile)` *(new)*; `newProfiler` calls it with the stored map (or a new one, as today); the check calls it with an empty map, so nothing is shared with the stored value. `buildShortlist(g *graphcache.Graph, p *profiler, cfg graphalign.Config) (*shortlist.Index, error)` *(new)* is `loadShortlist`'s kinds loop, which `loadShortlist` then calls. When `stored.shortlist != nil` the check compares `buildShortlist`'s result with `reflect.DeepEqual`; for each stored profile under `stored.profilesPattern == namePattern` it builds that handle's profile with the fresh profiler and compares. `gen` is not compared. The check may load nodes (Verify runs after the commit's reload) | Verify then covers every moved entry (map Check); S16-20's "Verify compares the patched entry with a fresh build" reuses it | Comparing only counts — misses a wrong key; a check that shares the stored map — compares the value with itself (review X-1) |
| D-14 | Alias index Check: `Load(g)` + `buildIndex` and `reflect.DeepEqual` with the stored index. Name format Check: `Load(g)` and `reflect.DeepEqual` with the stored `Loaded` | Same shape as every other Check | — |
| D-15 | `dropDerived` goes. Evidence-graph memos are emptied on a vocabulary commit by a loop in `OnCommit`'s vocabulary branch, and with their graph and `Drop` as today; an alias index or name format change no longer empties them. The doc lists that loop as the Evidence-graph store's one hand-written drop, owned by S16-13, and STEP-8 adds it to S16-13's scope | The layer reads the vocabulary, terms and the name pattern string, which is in its key; the candidates memo is keyed by name pattern and ignored when the match index generation changes, which a new alias index or name format causes by dropping the match index (D-12). Results do not move. Rule K-18 puts per-graph memos in the store, which S16-13 reshapes (review M-6) | Registering Evidence-graph memos as entries — they live and die with the loaded graph |
| D-16 | `canonicalgraph` is deleted. `TestInputsMatchIsolatedComposition`'s `assoc(h, from, to)` becomes `assocBetween(t, db, a, b []byte) []byte` *(new, in `inputs_test.go`)*: `SELECT x.entity_id FROM auto_reconciler_values x JOIN auto_reconciler_values y ON y.entity_id = x.entity_id WHERE x.value_entity_id = ? AND y.value_entity_id = ? AND x.rank = 1 AND y.rank = 1 AND x.reason = 'kept' AND y.reason = 'kept'`, failing the test unless exactly one row. The four calls drop their hop: `assocBetween(t, db, w.birth, w.town)`, `(w.town, w.city)`, `(w.city, w.region)`, `(w.region, w.country)` | The oracle stays independent of the cache it checks; in that world each pair has one association | Using `graphcache.Follow` — the oracle must not be the cache |
| D-17 | `BenchmarkCacheSizes` *(new)* in `core/database/promotealign/bench_test.go`, sub-benchmarks `persons=4000` and `persons=16000`. `largeShape(r, names, treeRoot, persons int)` sets `Persons: persons` and changes nothing else; `newLargeFixture(tb testing.TB, persons int)` (it already uses `newFixture(testing.TB)` and `must`); `TestLargeSourceBudgets` passes 4,000. The benchmark builds the fixture once per size, then `g.Drop()`, and measures `runtime.MemStats.HeapAlloc` after `runtime.GC()` before and after each fill, in this order, each reported with `b.ReportMetric(delta, "<metric>-B")`: (1) `nodes`: `g.Kind` of `graphcachetest.SubjectTypeID(b, c, "person")`, `"event"`, `"place"`; (2) `aliasIndex`: `namealiases.Warm(c)`; (3) `nameFormat`: `nameformats.Warm(c)`; (4) `matchIndex`: `promotealign.Warm(c)`; (5) `evidence`: one proposal on Source A (`f.propose(f.Graph(f.source), nil)`, as the "cold" subtest); (6) `kinship`: `canonlinks.NewReader(g, matchregistry.Default().Kinship)`, then `Relationships(id, canonlinks.Options{Siblings: true})` for every person id from `conclusionsummariestest.HandleIDs(b, db, vocabkey.Person)`; this fills the `kinship`, `for` and `siblings` entries together and is one metric; (7) `displays`: `conclusionsummaries.RowsForSearch(g, w)` with `w` as in "compose every summary". It also reports `handles` (`SELECT COUNT(*) FROM canonical_entities`) and, looping over `g.KeyCounts()`, `<Name>-keys` for every keyed entry, so a later keyed entry is reported with no benchmark edit and no slot is exported for a test. Reported, not gated; no tolerance | One reusable measure for the 40 MB budget; each later PR adds its fill. Persons is the only axis the map's Scale line grows | A gated budget — rule "Measure growth" binds S16-11 on; scaling events and the tree — the Scale line names persons |
| D-18 | `docs/catalog-caches.md` *(new)*: (a) registry entries, one row per `Name`; (b) typed `Graph` fields (nodes, holders, kinds, vocabulary, labels, displays, activity, the Evidence-graph store with its memos) with a "why typed" column, the Evidence-graph memos' vocabulary loop marked as S16-13's; (c) `core/catalogsession/session.go` `reg`; (d) Swift: one row for the per-view keys that points to `CatalogQueryRegistry.swift` `CatalogQueryKey.Kind.descriptor` as the source of truth, with the 50-hidden-entity cap `WorkspaceSession.maxHiddenEntities`, and one row each for `WorkspaceSession.evidenceGraphIndex`, `WorkspaceSession.nameFormat`, `CatalogCounts`, `ThumbnailCache.shared` (100 entries, LRU), the single-entry `Memo` holders (`SourcesModel.visibleRowsMemo`, `CitationComposerContext.memo`, `CitationComposerModel.depictionPreviewMemo`), `SourcesModel.searchKeyCache` (per row) and `EvidenceGraphModel.snapshotMemo`. Section (a) marks the three drop-on-any-handle Applies (`canonlinks.for`, `canonlinks.siblings`, the match index) as known exceptions owned by S16-16 and S16-20. Columns: Cache, Owner (path · symbol), Key, Fill, Updated by (the `effects.Set` field or DerivedFrom), Lifetime, Cap, Size at 4,000 / 16,000. Swift sizes "not measured" | Goal bar 15; one Swift row avoids 22 rows that S16-24 – S16-29 would re-edit (review M-7) | Listing caches in `cache-architecture.md` — deleted at spike close |
| D-19 | `graphcache.EntryNames() []string` *(new)* returns every registered entry's `Name` in registration order. `TestCacheDocListsEveryEntry` *(new)* in `core/catalogsession/cache_doc_test.go` (the package imports every owner) parses section (a) of `docs/catalog-caches.md` and fails when a registered name has no row or a row names no registered entry. Names are unique (`Register*` panics on a repeat) | No later PR can register a cache without its row (review M-5); precedent `TestKnobInventoryMatchesRegistry` | A manual `rg` in STEP-7 only — drifts with five PRs to come |
| D-21 | On an `Over` graph (`g.alt != nil`), `Whole.Get` runs a `LifeCatalog` entry's Build (it reads catalog configuration through the transaction, once per write graph) and refuses a `LifeGraph` entry's Build: it returns `ok == false` and stores nothing, so the owner composes without that cache. `Whole.Set` of a `LifeGraph` entry on an `Over` graph stores nothing either (review M-12). Keyed handles fill on an `Over` graph as today (a write fills only the keys its composers read). The rule "(Batch 2) A write's one-shot graph fills only what it reads" is reworded in STEP-8 | User, 2026-11-11 (review M-1): catalog configuration is a few rows the write path reads today on every call; graph data must never build per write, and the mechanism holds that for S16-12 and S16-15 | Refusing every Build there — keeps the cold branches, a second read path; sharing the session graph's values through `Over` — changes `searchindex` and `autoreconciler`, which receive only a transaction |
| D-20 | Registry tests run against a snapshot: `graphcache/export_test.go` *(new)* has `SnapshotRegistryForTest() (restore func())`; `cache_test.go` registers its test entries A–E and the invalid entries of AC-1.8 between snapshot and restore, so no test entry stays registered for other tests | Package-level registry state must not leak across tests (review X-4) | Registering test entries globally in `var` — they would run their Apply on every other test's commits |

**Rules this plan bends.** D-21 → the working rule "(Batch 2) A write's one-shot graph fills only what it reads" ("No whole Build runs on a write path"): a catalog-lifetime Build (catalog configuration) now runs once on a write's graph, and the registry refuses any graph-lifetime Build there → the user chose this on 2026-11-11 (review M-1) → STEP-8 rewords the rule to "No whole Build of graph data runs on a write path: the registry neither builds nor stores a `LifeGraph` entry on a write's graph; catalog configuration (`LifeCatalog`) builds once there".

D-1, D-2 → the rule "(Batch 2) One cache registry, two ways to fill" names a per-key "Fill and `Deps`" → the user chose typed get-or-fill handles on 2026-11-11 (review M-14): the fill is a closure at the call site, `Deps`' work is the entry's Apply → STEP-8 rewords the rule to "keyed, filled through `Keyed[T].Get` / `GetMany` with the caller's fill and updated by an Apply; or whole, `Whole[T]` with a lazy Build". Not a bend: the map's S16-16 scope says "`DropOnHandleCommit` and `dropHandleMemos` go" there; D-9 removes those names here and leaves S16-16 deleting the two entries and their Apply (map wording, STEP-8).

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/graphcache/cache.go` `Keyed[T]`, `Whole[T]`, `KeyedEntry[T]`, `WholeEntry[T]`, `RegisterKeyed`, `RegisterWhole`, `Source`, `Vocabulary`, `Lifetime`, `EntryNames` *(new)* | type / func / file | `core/database/graphcache/memo.go` `NodeMemo`, `MemoSlot`, `RegisterNodeMemo` | It is that mechanism extended, typed and split by kind; `memo.go` is deleted (searched `Register`, `Slot`, `memo` in `core/database`) |
| Handle methods `Get`, `GetMany`, `Peek`, `Set`, `Drop`, `Clear`; `Graph.KeyCounts` *(new)* | func | `Memo`/`SetMemo`, `MatchIndex`/`SetMatchIndex`, `AliasIndex`/`SetAliasIndex`, `NameFormat`/`SetNameFormat`, `dropMemos`, `canonlinks` `Reader.answer` | Those accessors and the owners' read-cast-compute-store blocks are deleted; the handles replace them |
| `namealiases.Index`, `nameformats.Default`, `promotealign.MatchIndex` *(new)* | var | `canonlinks` `kinshipSlot` | Same pattern, exported because other entries derive from them |
| `checkMatchIndex`, `buildShortlist`, `newProfilerWith` *(new)* | func | `loadShortlist`, `newProfiler` | Split out of those so the check and the load share one builder |
| Alias index and name format Checks *(new, unexported)* | func | `canonlinks.check` | Same shape |
| `SnapshotRegistryForTest` *(new, test)* | helper | `SetEvidenceGraphCapForTest` | Test-only hook in `export_test.go` |
| `TestCacheDocListsEveryEntry` *(new, test)* | test | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry` | Same pattern for another doc |
| `assocBetween` *(new, test)* | helper | `canonicalgraph.Walk` | The package is deleted; one query is the oracle |
| `BenchmarkCacheSizes` *(new, test)* | benchmark | `BenchmarkProposeColdCity` heap metric | Reports per cache, at two sizes |
| `docs/catalog-caches.md` *(new)* | doc | `cache-architecture.md` (spike design) | Spike docs are deleted at close; this lasts |

**Builds on.** The owner-registers pattern (`RegisterNodeMemo`, `SetTruth`); `OnCommit`'s root computation unchanged; the S11-21 large fixture; the knob-inventory parity test.
**Deletes.** `core/database/graphcache/memo.go` (`NodeMemo`, `MemoSlot`, `MemoCheck`, `MemoDeps`, `RegisterNodeMemo`, `dropMemos`, `dropVocabularyMemos`, `dropHandleMemos`, `verifyMemos`); in `graph.go` the fields `matchIndex`, `matchIndexOK`, `warmNeeded`, `aliasIndex`, `aliasIndexOK`, `nameFormat`, `nameFormatOK` and `MatchIndex`, `SetMatchIndex`, `ResetMatchIndexForTest`, `dropMatchIndex`, `dropDerived`, `WarmNeeded`, `MarkWarm`, `AliasIndex`, `SetAliasIndex`, `NameFormat`, `SetNameFormat`; `catalogsession.runLocked`'s re-warm; the `q` parameters of `Rules`, `LoadFor`, `valuerules.Load` and `promotealign.loadConfig`; `valuerules.Querier` and its `database/sql` import; the false clause of `evidencelayer.Load`'s comment ("and new alias tables"); `core/database/canonicalgraph/` (three files); `TestWarmNeeded`, `TestNodeMemoDropsOnAnyCommit`, `TestNodeMemoDropsOnCommit`, `TestDropLeavesAliasIndex`, `TestDropLeavesNameFormat` (folded into T-1, T-2, T-9). The cold branches of `namealiases.Rules` and `nameformats.LoadFor`.
**Next adds.** S16-12 and S16-15 register whole entries with a Build and an Apply on `SourceScopes` / `Set.Reads`; S16-16 deletes the `for` and `siblings` entries; S16-18 registers keyed place entries; S16-20 gives the match index an Apply that patches instead of drops and reuses `checkMatchIndex`; S16-13 moves the Evidence-graph memos' vocabulary drop off `OnCommit`. Each adds a row to `docs/catalog-caches.md` *(new)* (held by `TestCacheDocListsEveryEntry`) and a fill to `BenchmarkCacheSizes`.

### Reach

None — no new path into existing code, no new state reachable. A write's `Over` graph now holds the alias index and name format it builds (D-21) and is discarded with the write, as today.

### Scale

At the map's Scale (large fixture, 4,000 and 16,000 persons; crafted: a 20,000-child hub, an event with 500 participants): Apply per commit does what `OnCommit` does today. The kinship Apply deletes the roots and their one-hop neighbours from cached nodes: a commit naming the hub touches 20,001 keys, an event with 500 participants 501, as `Deps` does today. `for`, `siblings` and the match index each clear one map. DerivedFrom propagation walks at most the six registered entries. A Build runs on the session graph only at open (D-11). No cap binds at Scale (D-6; `BenchmarkCacheSizes` reports the key counts). Memory is unchanged; `BenchmarkCacheSizes` records it.

### Edges

- **Assumptions.** Roots passed to Apply equal today's `dropMemos` roots (AC-1.6, T-2). DerivedFrom forms a DAG because a slot may derive only from earlier slots (AC-1.8). A catalog-lifetime entry never derives from graph data (D-5). Apply errors drop the graph through `Notify` (D-3).
- **Neighbors.** A vocabulary commit drops the kinship entries and the match index and keeps the alias index and name format (T-1, T-9). A Source vocabulary commit (`ListSourceVocabulary`) drops nothing (T-1, T-9). A first load drops nothing (T-2). A Build's first fill drops no reader; a later `Set` does (T-4).
- **Removed inputs.** An Evidence-graph memo now outlives an alias index or name format change; it does not read them (D-15, T-10).
- **Real data.** None — no table- or seed-driven rule changes.
- **Untrusted size.** Keyed entries are capped (D-6, T-3); whole entries are bounded by the catalog they are built from.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` | `cache.go` *(new)*, `memo.go` deleted, `export_test.go` *(new)*; `graph.go` fields and methods per **Deletes**, `Cached` moved here, one `entries` map from entry index to its held data *(new)* on `Graph`; `clear` empties `LifeGraph` entries; `commit.go` vocabulary branch: `DropLabels`, vocab reset, Evidence-graph memos emptied, entries derived from `graphcache.Vocabulary` dropped; Apply loop after the reload, also on a commit with no handles; `verify.go` runs every entry's Check |
| `core/database/canonlinks` | `reader.go`: three keyed entries per D-9 |
| `core/database/namealiases`, `nameformats` | `Index` / `Default` handles, Build, Check, `Warm`, `Invalidate`, `Rules(g, base)` / `LoadFor(g)` per D-10, D-14 |
| `core/database/valuerules` | `Load(g)`; `Querier` goes |
| `core/database/promotealign` | `MatchIndexSlot` in `init`, `buildShortlist`, `newProfilerWith`, `checkMatchIndex`; `loadConfig(g)`; `matchIndexOf`, `loadShortlist`, `keep` through the registry |
| `core/database/evidencelayer` | `evidencelayer.go` `Load` comment |
| `core/database/persondetail`, `conclusionsummaries`, `evidencelayer/evidencelayertest`, `api/ffi/handlers/nameformat.go` | Call sites of the readers |
| `core/catalogsession` | `warmGraph` without `MarkWarm`; `runLocked` without the re-warm; comments of `Do`, `WarmOnOpen`; `cache_doc_test.go` *(new)* |
| `core/database/canonicalgraph` | Deleted |

```go
// cache.go sketch
type Lifetime int
const (LifeGraph Lifetime = iota; LifeCatalog)
type Source interface{ source() int }      // Whole[T] and Vocabulary
var Vocabulary Source                      // the typed vocabulary, for DerivedFrom
type KeyedEntry[T any] struct {
	Name        string
	Apply       func(g *Graph, k Keyed[T], roots [][]byte, set effects.Set) error // required
	DerivedFrom []Source
	Check       func(g *Graph, id []byte, stored T) error
	Cap         int // required, > 0
}
type WholeEntry[T any] struct {
	Name        string
	Build       func(g *Graph) (T, error) // nil: the owner calls Set
	Apply       func(g *Graph, w Whole[T], roots [][]byte, set effects.Set) error
	DerivedFrom []Source
	Lifetime    Lifetime
	Check       func(g *Graph, stored T) error
}
func RegisterKeyed[T any](e KeyedEntry[T]) Keyed[T]
func RegisterWhole[T any](e WholeEntry[T]) Whole[T]
func (k Keyed[T]) Get(g *Graph, id []byte, fill func() (T, error)) (T, error)
func (k Keyed[T]) GetMany(g *Graph, ids [][]byte, fill func(missing [][]byte) (map[string]T, error)) (map[string]T, error)
func (w Whole[T]) Get(g *Graph) (T, bool, error)
func EntryNames() []string
```

Errors: none new; a nil graph in the readers returns `database.ErrClosed`.

### FFI

None — `api/ffi/handlers/nameformat.go` `GetNameFormat` drops its `db` read and calls `nameformats.LoadFor(c.Graph())`; the wire does not change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — The registry.**
Files: `core/database/graphcache/cache.go` *(new)*, `export_test.go` *(new)*, `memo.go` (delete), `graph.go`, `commit.go`, `verify.go`, `cache_test.go` *(new)*, `graph_test.go`. Skill: `add-catalog-query` (for package layout only).
Do: D-1 – D-8, D-19's `EntryNames`, D-20, D-21. `Cached` moves into `graph.go`. `OnCommit`: delete the `dropHandleMemos` call; after the reload run the Apply loop with `roots` and `set`; on a commit with no handles run it with nil roots before returning. Vocabulary branch: keep `DropLabels`, the vocab reset and the `dropDerived()` call (it still drops the match index and Evidence-graph memos until STEP-4), replace `dropVocabularyMemos` with dropping the entries derived from `graphcache.Vocabulary`. Verify runs Checks. Keep `Memo`/`SetMemo`/`MemoSlot`/`RegisterNodeMemo` as thin shims over the new registry only until STEP-2 moves `canonlinks` (delete them in STEP-2). Leave `matchIndex`, `aliasIndex`, `nameFormat`, `warmNeeded`, `dropMatchIndex` and their accessors in place for now so the tree builds (the match index keeps its pre-reload drop until STEP-4). `TestNodeMemoDropsOnCommit` and `TestNodeMemoDropsOnAnyCommit` become T-1 – T-6 with registered test entries under `SnapshotRegistryForTest`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/...` passes, T-1 – T-6 included.

**STEP-2 — Kinship entries.**
Files: `core/database/canonlinks/reader.go`, `reader_test.go`.
Do: D-9. Delete the STEP-1 shims (`Memo`, `SetMemo`, `MemoSlot`, `RegisterNodeMemo`). Tests that stored memos directly (`reader_test.go` lines 643, 1124) use the handle's `Set`; `hasMemo` uses `Peek`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/... ./core/database/persondetail/... ./core/database/promotealign/...` passes; T-7.

**STEP-3 — Alias index and name format.**
Files: `core/database/namealiases/namealiases.go`, `namealiases_test.go`; `core/database/nameformats/nameformats.go`, `nameformats_test.go`; `core/database/valuerules/valuerules.go`, `valuerules_test.go`; `core/database/promotealign/promotealign.go`, `index.go`, `memo_test.go`; `core/database/persondetail/recordlinks.go`; `core/database/conclusionsummaries/compose.go`, `conclusionsummaries_test.go`; `core/database/evidencelayer/evidencelayertest/evidencelayertest.go`; `api/ffi/handlers/nameformat.go`; `core/catalogsession/session.go`, `warm_test.go`, `alias_cache_test.go`; `core/database/graphcache/graph.go`, `graph_test.go`. Skill: `use-catalog-session`.
Do: D-10, D-11, D-14. In `dropDerived` delete the `g.warmNeeded = true` line (the function itself goes in STEP-4). Delete `aliasIndex*`, `nameFormat*`, `AliasIndex`, `SetAliasIndex`, `NameFormat`, `SetNameFormat`, `warmNeeded`, `WarmNeeded`, `MarkWarm` from `graph.go`, and `TestDropLeavesAliasIndex`, `TestDropLeavesNameFormat`, `TestWarmNeeded` (their cases are in T-1 and T-9). `valuerules_test.go` passes `graphcache.Over(db)`. `TestDoRewarms` becomes T-9. `graph_test.go` `TestMatchIndexDrops` loses its `SetAliasIndex` and `SetNameFormat` cases (they return in T-11, STEP-4). Between STEP-3 and STEP-4 nothing drops the match index when the alias index or name format changes; this is intended, since nothing in the app changes either, and STEP-4's DerivedFrom restores it: add no patch.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes.

**STEP-4 — Match index; `dropDerived` goes.**
Files: `core/database/promotealign/index.go`, `profiles.go`, `export_test.go`, `promotealign_test.go`, `memo_test.go`; `core/database/graphcache/graph.go`, `commit.go`, `graph_test.go`, `evidence_graph.go` (doc comments), `evidence_graph_memo_test.go`; `core/database/evidencelayer/evidencelayer.go` (comment); `core/catalogsession/warm_test.go`. Skill: `add-match-rule` (home of the match index only).
Do: D-12, D-13, D-15. Delete `matchIndex*`, `MatchIndex`, `SetMatchIndex`, `ResetMatchIndexForTest`, `dropMatchIndex`, `dropDerived`; `TestMatchIndexDrops` becomes T-11 in `promotealign`. Update `EvidenceGraphMemo`'s and `evidencelayer.Load`'s comments and T-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/... ./core/catalogsession/...` passes.

**STEP-5 — `canonicalgraph` goes.**
Files: `core/database/canonicalgraph/` (delete), `core/database/conclusionsummaries/inputs_test.go`.
Do: D-16.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/...` passes; `rg -n canonicalgraph core api` has no hit.

**STEP-6 — Measure.**
Files: `core/database/promotealign/bench_test.go`, `perf_test.go`.
Do: D-17.
Check: `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench BenchmarkCacheSizes -benchtime 1x ./core/database/promotealign/` prints every metric of D-17 at both sizes; `TestLargeSourceBudgets` passes. Paste the output in the PR description.

**STEP-7 — The cache table.**
Files: `docs/catalog-caches.md` *(new)*, `core/catalogsession/cache_doc_test.go` *(new)*.
Do: D-18, D-19's test. Sizes from STEP-6. Link the doc from `docs/promote-graph-alignment.md` §8 and from `cache-architecture.md` § One cache registry.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestCacheDocListsEveryEntry ./core/catalogsession/` passes; it fails when one row is deleted (try it, then restore).

**STEP-8 — Guides.**
Files: `.cursor/rules/catalog-queries.mdc`, `.agents/skills/use-catalog-session/SKILL.md`, `docs/promote-graph-alignment.md`, `.agents/skills/review-pr/lenses/maintainability.md`, `docs/deployment-plan/spike-16/cache-architecture.md`, `docs/deployment-plan/spike-16/deployment-plan.md`.
Do: `catalog-queries.mdc` line 17 drops `canonicalgraph.Walk` (a walk past the cache reads SQL in its own package, like `searchindex`'s input lookups); line 18 says a derived cache registers a typed handle (`graphcache.RegisterKeyed[T]` or `RegisterWhole[T]`), reads only through its `Get`/`GetMany`, and declares its Apply, DerivedFrom, lifetime, Check and Cap, adds its row to `docs/catalog-caches.md` *(new)*, and never adds a field on `Graph`. `use-catalog-session` line 62: the warm list without `WarmNeeded`, the readers' new signatures, `Invalidate` drops the entry. `promote-graph-alignment.md`: line 158's "S9-28 caches it as canonical edges" → the graph cache holds each node's links (`graphcache.Node.Links`); lines 165–169 describe the match index as a registry entry derived from the vocabulary, alias index and name format, and the open-time warm without the re-warm. Maintainability lens Known traps, a new line: "A cache that empties itself on any write, or a drop rule written in `OnCommit` instead of an entry's Apply and DerivedFrom — `docs/catalog-caches.md`, rule 'Every cache names the signal that updates it'". `cache-architecture.md` § One cache registry: typed keyed and whole handles, "Fill: keyed (`Keyed[T].Get` / `GetMany` with the caller's fill closure, and an Apply in place of `Deps`) or whole (`Whole[T]` with a lazy Build)", "Reads" → "DerivedFrom". Map: the Batch 2 rule "One cache registry, two ways to fill" says "DerivedFrom" for "Reads" and "keyed, filled through `Keyed[T].Get` / `GetMany` with the caller's fill and updated by an Apply" for "per key, today's `NodeMemo` with Fill and `Deps`" (user, 2026-11-11); the rule "Every cache names the signal that updates it" says "the drop-on-any-handle Applies of `for` and `siblings` (S16-16) and of the match index (S16-20) are the known exceptions" in place of its `DropOnHandleCommit` clause; the reworded write-path rule says "the registry neither builds nor stores graph data on a write's graph"; the rule "A write's one-shot graph fills only what it reads" is reworded per **Rules this plan bends**, and `cache-architecture.md` § "The write path fills only what it reads" with it; S16-13's Ships gains "the Evidence-graph memos' vocabulary drop moves off `OnCommit`"; S16-16 Ships "the `for` and `siblings` entries and their drop-on-any-handle Apply go"; S16-20's "dropping the match index on every handle commit" → "the match index's drop-on-any-handle Apply".
Check: the `rg` of AC-3.2 has no hit; `python3 scripts/check-guidance-names.py` passes.

**STEP-9 — Docs.**
Files: `completed.md`, `deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: links resolve; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period ("Hold the graph's derived caches in one registry of entries."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphcache/cache_test.go` `TestEntryDerivedFrom` *(new)* | Table: A.Set / vocabulary commit / Source vocabulary commit / B.Drop(g) / `Drop()` over entries A–E of AC-1.1 → the held set after each | AC-1.1, AC-1.2, AC-1.3, D-4, D-5 |
| T-2 | Go | `cache_test.go` `TestEntryApply` *(new)* | Recording Apply; association commit, merge, delete of a loaded node, Source-only commit, first load → roots and set per AC-1.6 | AC-1.6, D-3 |
| T-3 | Go | `cache_test.go` `TestEntryCap` *(new)* | Cap 2: k1, k2, k3, then k1 again → k3 absent, k1 replaced | AC-1.4, D-6 |
| T-4 | Go | `cache_test.go` `TestEntryBuild` *(new)* | Counting Build; two `Get` calls → 1 call, reader R kept; then `Set` → R dropped; a failing Build → error, nothing stored; on `Over(tx)`: a `LifeCatalog` Build runs reading the transaction, a `LifeGraph` Build is not called and `Value` returns `ok == false`; `Set` of a `LifeGraph` entry on `Over(tx)` → `Get` still `ok == false` | AC-1.5, AC-1.10, D-2, D-21 |
| T-5 | Go | `cache_test.go` `TestEntryVerify` *(new)* | `t.Setenv("PROVENENCIA_GRAPH_VERIFY", "1")`; rejecting Check on a keyed and a whole entry; a commit → error | AC-1.7, D-7 |
| T-6 | Go | `cache_test.go` `TestRegisterRejects` *(new)* | Each invalid entry of AC-1.8, under `SnapshotRegistryForTest` → panic | AC-1.8, D-20 |
| T-7 | Go | `core/database/canonlinks/reader_test.go` (existing memo cases, renamed API) | Participation commit naming an event → its and neighbours' pivots gone, unrelated pivot kept, `for`/`siblings` empty | AC-2.3, D-9 |
| T-8 | Go | `core/catalogsession/warm_test.go` `TestNameFormatDropsItsReaders` *(new)* | Open project, a proposal and a Family read run; `nameformats.Default.Set(g, loaded)` → match index absent; alias index (0 SELECTs for `valuerules.Load`) and a stored kinship pivot present | AC-2.1, D-12 |
| T-9 | Go | `core/catalogsession/warm_test.go` `TestDoKeepsWarmEntries` (was `TestDoRewarms`) | The five cases of AC-2.2 → 0 SELECTs for `valuerules.Load` each; match index per AC-2.2; a drop leaves the person kind unloaded | AC-2.2, D-11 |
| T-10 | Go | `core/database/graphcache/evidence_graph_memo_test.go` `TestEvidenceGraphMemoDroppedWithSource` | The `SetAliasIndex` case becomes: a test `LifeCatalog` whole entry is Set → memo kept; a vocabulary commit → memo gone | AC-2.5, D-15 |
| T-11 | Go | `core/database/promotealign/memo_test.go` `TestMatchIndexDrops` *(moved from graphcache)* | Stored match index; Source-only commit → kept; handle commit → gone; refill; vocabulary commit → gone; refill; `namealiases.Index.Set` → gone; refill; `Drop()` → gone | AC-2.4, D-12 |
| T-12 | Go | `core/database/namealiases/namealiases_test.go` `TestVerifyChecksAliasIndex` *(new)*, `core/database/nameformats/nameformats_test.go` `TestVerifyChecksNameFormat` *(new)*, `core/database/promotealign/memo_test.go` `TestVerifyChecksMatchIndex` *(new)* | Verify on; stored value equal to a fresh build → commit passes; an alias index with no tables over a catalog with one / a `Loaded` with another key / a match index whose shortlist lacks a handle → commit fails naming the entry | AC-2.6, D-13, D-14 |
| T-13 | Go | `core/database/conclusionsummaries/inputs_test.go` `TestInputsMatchIsolatedComposition` | Same cases, `assocBetween` in place of `canonicalgraph.Walk` | AC-2.7, D-16 |
| T-14 | Go | `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes` *(new)* | 4,000 and 16,000 persons → every metric of D-17 | AC-3.1, D-17 |
| T-15 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` | Unchanged budgets, fixture at 4,000 | AC-2.7 |
| T-16 | Go | `core/catalogsession/cache_doc_test.go` `TestCacheDocListsEveryEntry` *(new)* | Registered names vs section (a) rows → equal sets | AC-1.9, AC-3.1, D-19 |
| T-17 | Sweep | STEP-8 check: the `rg` of AC-3.2 over `core api .agents .cursor docs/*.md` | → no hit | AC-3.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/catalogsession` (warm order), `core/database/promotealign` (`memo_test.go`, `TestLargeSourceBudgets`), `core/database/canonlinks` (memo tests), `core/database/searchindex` (rebuild-equals-upkeep).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/... passes
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (AC-3.2's list); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status, the two reworded rules, and the S16-13, S16-16, S16-20 scope rewording (STEP-8); completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated)
- [ ] Domain doc: docs/catalog-caches.md (new) with sizes from BenchmarkCacheSizes; docs/promote-graph-alignment.md §7–§8
- [ ] PR description: what shipped, AC evidence, deviations, the BenchmarkCacheSizes output
```

## 10. Manual check

1. Open a real project in the app. Promote a Source with a person, open that person's Family section, rename the Source, edit an Observation of the person, and propose again. Each view shows what it showed before this branch (compare with `main` on the same project).
2. Run `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench BenchmarkCacheSizes -benchtime 1x ./core/database/promotealign/` and check the Go caches' sum at 16,000 persons against the 40 MB budget; report the sum. A sum over 40 MB is reported, not fixed here (the caches are today's).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Build on the session graph runs inside an open write transaction and waits on the catalog's one connection | D-11: the session graph's alias index and name format are filled at open and never dropped in-app; T-9 asserts 0 SELECTs after every case |
| A moved drop rule misses a case | T-1, T-7, T-9, T-11 carry today's cases; `PROVENENCIA_GRAPH_VERIFY=1` over `./core/...` |
| `reflect.DeepEqual` cannot compare a stored value (a func field) | Stop condition below |

**Stop and report if:**

- `OnCommit`'s roots differ from §4's description (another root source was added since `6c8736d`).
- `reflect.DeepEqual` on `valuecompare.AliasIndex`, `nameformats.Loaded`, `*shortlist.Index` or `profile` reports a difference for two fresh builds of the same catalog.
- A caller of `valuerules.Load`, `namealiases.Rules` or `nameformats.LoadFor` passes a querier that is not its graph's own connection or transaction (a transaction with the session graph), so dropping `q` would change what it reads.
- Any test's expected result (not its API) has to change, beyond T-9's `WarmNeeded` column and T-10's alias case.
- A caller of `canonlinks.NewReader` passes rules other than `matchregistry.Default().Kinship`, or a test depends on a Reader with other rules (D-9 drops the parameter).
- `BenchmarkCacheSizes` reports more than 32,768 keys in any keyed entry at 16,000 persons (half its cap).
- `BenchmarkCacheSizes` at 16,000 persons does not finish within 10 minutes on the implementer's machine.

## 12. Open questions

None. Answered by the user on 2026-11-11: the write path (D-21) and typed get-or-fill handles (D-1, D-2).

## 13. PR

**Title:** S16-10 — One cache registry
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; the `BenchmarkCacheSizes` output; `Plan: docs/deployment-plan/spike-16/prs/S16-10-one-cache-registry.md`.
