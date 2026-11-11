# S16-10 — One cache registry

**Status:** Draft
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
**Context:** [design § One cache registry](../cache-architecture.md#one-cache-registry-two-ways-to-fill-s16-10); working rules: "(Batch 2) The point of this batch is simpler code", "(Batch 2) One cache registry, two ways to fill", "(Batch 2) Every cache names the signal that updates it, and updates exactly", "(Batch 2) Apply states its fan-out", "(Batch 2) Memory is budgeted", "(Batch 2) Warm only what every proposal needs; no background re-warm"; goal bar 15; [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §7–§8
**Planned:** 2026-11-11 against `main` @ `6c8736d`
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

`graphcache` keeps five kinds of derived cache, each with its own slot and its own drop rule: three `NodeMemo` slots with flags (`DropOnVocabulary`, `DropOnHandleCommit`), a match index field dropped by hand in `OnCommit`, and an alias index and name format whose setters drop the match index and every Evidence-graph memo as a side effect (`dropDerived`) and raise a `warmNeeded` flag that `catalogsession.Do` answers with a re-warm. This PR gives them one registry: an entry declares how it fills (per key, or whole with a lazy Build), what a commit does to it (Apply, on the roots `OnCommit` reloaded), what it is derived from (Reads), how long it lives (the graph or the catalog connection), how Verify checks it and its cap. The five move onto it with today's results; the flags, `dropDerived`, `warmNeeded` and the dead `canonicalgraph` package go, and `docs/catalog-caches.md` *(new)* lists every Go and Swift cache with its signal and its size at 4,000 and 16,000 persons.

**Why now.** Every later Batch 2 Go PR (S16-12, S16-15, S16-16, S16-18, S16-20) adds or reshapes a cache. Without one registry each would add a sixth slot with a seventh drop rule.

**What stays.** Every read returns what it returns today: the same Promote suggestions, Family, summaries and search rows. The match index still drops on every handle commit (until S16-20); the `for` and `siblings` answers still drop on every handle commit (until S16-16). `Activity`, the summary displays, the labels, the vocabulary, the node maps and the Evidence-graph store stay typed fields of `Graph`. No SQL, wire or Swift code changes.

## 2. User stories

**US-1.** As a contributor adding a cache in Batch 2, I register one `graphcache.Entry` that states its fill, its commit rule, what it reads, its lifetime, its check and its cap, so that I write no new drop path in `OnCommit` and no new flag.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A test registers whole entry A (catalog lifetime), whole entry B reading A, per-key entry C reading B, per-key entry D reading the vocabulary, per-key entry E reading nothing; all hold a value; `SetEntry(A, v)` | B and C are empty; A holds `v`; D and E keep their values |
| AC-1.2 | Same entries filled; a commit with `Lists: [ListPropertyVocabulary]` | D is empty; A, B, C, E keep their values |
| AC-1.3 | Same entries filled; `Drop()` | B, C, D, E are empty; A keeps its value |
| AC-1.4 | Per-key entry with `Cap: 2` holding keys k1, k2; `SetMemo(k3)`; then `SetMemo(k1, v2)` | k3 is not stored; k1 now holds `v2` |
| AC-1.5 | Whole entry with a Build that counts its calls; `Entry` twice on the session graph; then `Entry` on `graphcache.Over(tx)` | Build ran once on the session graph and once on the `Over` graph, the second reading through the transaction; a Build error returns the error and stores nothing |
| AC-1.6 | Per-key entry with an Apply that records its arguments; a participation association commit; a merge commit; a delete of a loaded node; a Source-only commit (`SourceScopes`, no `Handles`) | Apply gets the association's ends, the merged handle's neighbours, the deleted node's link neighbours (today's `dropMemos` roots), and on the Source-only commit is called with no roots and the commit's `effects.Set` |
| AC-1.7 | `PROVENENCIA_GRAPH_VERIFY=1`; a per-key and a whole entry each hold a value their Check rejects; any commit | `OnCommit` returns the Check's error |
| AC-1.8 | `Register` with a per-key entry and no Cap; a Build on a per-key entry; a Cap on a whole entry; Reads naming an unregistered slot; a catalog-lifetime entry reading a graph-lifetime entry or the vocabulary | Each panics at registration |

**US-2.** As the engine, after this PR every moved cache reads and drops as it does today.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | An open project with the alias index, name format and match index warm, kinship pivots stored; `SetEntry(nameformats.Slot, …)` | The match index is dropped; the alias index and every kinship pivot stay (the map's Check "setting the name format drops exactly the entries that Read it") |
| AC-2.2 | An open project after `Do` returns from a vocabulary commit, a `Drop`, a handle commit, and a failed `fn` | `valuerules.Load(g)` issues 0 SELECTs in each case (the alias index and name format survive); the match index is absent in each case |
| AC-2.3 | A participation commit naming an event | The event's and its neighbours' kinship pivots are dropped, an unrelated person's pivot stays; every `for` and `siblings` answer is dropped |
| AC-2.4 | A Source-only commit, then a handle commit | The match index stays after the Source-only commit and is gone after the handle commit |
| AC-2.5 | An Evidence graph's layer memo stored; `SetEntry` of the alias index; then a vocabulary commit | The memo stays after the alias index change and is gone after the vocabulary commit |
| AC-2.6 | `PROVENENCIA_GRAPH_VERIFY=1`; the stored alias index, name format or match index replaced by one that differs from a fresh build | `Verify` returns an error naming the entry |
| AC-2.7 | `TestLargeSourceBudgets`, the canonlinks, promotealign, conclusionsummaries, searchindex and catalogsession suites | Pass unchanged in their expectations (only API renames in the tests) |

**US-3.** As a contributor, I read one table that names every cache, so that I know where a value lives, what updates it and what it costs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `docs/catalog-caches.md` *(new)* | Lists every Go cache (registry entries and typed fields) and every Swift cache with key, fill, updating signal, lifetime, cap, and the Go sizes at 4,000 and 16,000 persons from `BenchmarkCacheSizes`; says why `Activity`, displays and labels stay typed fields |
| AC-3.2 | `rg -n 'canonicalgraph\|RegisterNodeMemo\|NodeMemo\b\|DropOnVocabulary\|DropOnHandleCommit\|dropHandleMemos\|dropDerived\|WarmNeeded\|MarkWarm\|SetAliasIndex\|SetNameFormat\|SetMatchIndex' core api .agents .cursor docs/*.md` | No hit |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache` registry: `Entry`, `Slot`, `Register`, per-key and whole storage, Apply after `OnCommit`'s reload, Reads with transitive drops, lifetime, Check in Verify, Cap | STEP-1, AC-1.1 – AC-1.8 |
| `RegisterNodeMemo`, `NodeMemo`, `MemoSlot`, `MemoCheck`, `MemoDeps`, `dropMemos`, `dropVocabularyMemos`, `dropHandleMemos`, `verifyMemos` fold into it; `memo.go` deleted | STEP-1, STEP-2 |
| The three `canonlinks` slots as per-key entries; `deps` becomes the kinship entry's Apply; the `for` and `siblings` drop-on-any-handle becomes their Apply | STEP-2, AC-2.3 |
| Alias index and name format as catalog-lifetime whole entries with a Build, owned by `namealiases` and `nameformats`; the cold-graph fallbacks and the `q` parameters of `namealiases.Rules`, `nameformats.LoadFor`, `valuerules.Load` go | STEP-3, AC-2.2, AC-1.5 |
| `warmNeeded`, `WarmNeeded`, `MarkWarm` and `catalogsession`'s re-warm after `fn` go; `WarmOnOpen` stays | STEP-3, AC-2.2 |
| The match index as a graph-lifetime whole entry owned by `promotealign`, reading the vocabulary, the alias index and the name format; its handle-commit drop as its Apply; a Check | STEP-4, AC-2.1, AC-2.4, AC-2.6 |
| `dropDerived` goes: Evidence-graph memos drop on a vocabulary commit only | STEP-4, AC-2.5 |
| Checks for the alias index and the name format | STEP-3, AC-2.6 |
| `canonicalgraph` deleted; `TestInputsMatchIsolatedComposition` keeps an inline association query | STEP-5, AC-2.7 |
| `BenchmarkCacheSizes` reports each cache's heap at 4,000 and 16,000 persons | STEP-6, AC-3.1 |
| `docs/catalog-caches.md` *(new)* | STEP-7, AC-3.1 |
| Docs and guides this makes false; the maintainability lens's Known trap "a cache that empties on any write"; the map's S16-16 and S16-20 scope entries reworded to the registry's names | STEP-8, AC-3.2 |

| Out | Home |
| --- | --- |
| Any new cache (Source records, outcome records, graph catalog, place entries) | S16-11, S16-12, S16-15, S16-18 |
| The `for` and `siblings` entries and their drop-on-any-handle Apply | S16-16 |
| Dropping the match index on every handle commit; `indexGen`; the profiler's shared map | S16-20 |
| Evidence-graph memos moving into a records tier | S16-13 |
| Exporting `nodeKey` and `followKeys` | S16-18 |
| Callers for the dead `namealiases.Invalidate` and `nameformats.Invalidate` (no writer changes alias tables or the default name format in-app) | not planned |
| Any Swift change; Swift caches are only listed in the doc | S16-24 – S16-29 |
| A speed budget for any cache | S16-11 – S16-32 (rule "Measure growth") |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Per-node memos | `core/database/graphcache/memo.go` `NodeMemo`, `RegisterNodeMemo`, `Memo`, `SetMemo`, `dropMemos`, `dropVocabularyMemos`, `dropHandleMemos`, `verifyMemos` | Per-node results with `Deps`, `Check` and two flags; the flags are the drop rules this PR replaces |
| Graph fields | `core/database/graphcache/graph.go` `Graph.matchIndex`, `matchIndexOK`, `warmNeeded`, `memos`, `aliasIndex`, `aliasIndexOK`, `nameFormat`, `nameFormatOK` | Five hand-kept slots. `MatchIndex`, `SetMatchIndex`, `ResetMatchIndexForTest`, `AliasIndex`, `SetAliasIndex`, `NameFormat`, `SetNameFormat`, `WarmNeeded`, `MarkWarm`, `dropMatchIndex`, `dropDerived` |
| `dropDerived` | `graph.go` `dropDerived` | Empties every Evidence-graph memo and the match index and sets `warmNeeded`; called by `clear`, `SetAliasIndex`, `SetNameFormat` and the vocabulary branch of `OnCommit` |
| `clear` / `Drop` | `graph.go` `clear` | Empties nodes, kinds, displays, memos, Evidence graphs, labels, activity, vocab, and through `dropDerived` the match index; leaves the alias index and name format |
| Commit | `core/database/graphcache/commit.go` `OnCommit` | Vocabulary list → `DropLabels`, vocab reset, `dropDerived`, `dropVocabularyMemos`. Structure → `Drop`. Then `dropEvidenceGraphs`; returns early when `set.Handles` is empty; else `dropMatchIndex` and `dropHandleMemos` before the reload, and `dropMemos(roots)` after it, roots built by `detachAssociation`, `mergedNeighbors` and removed nodes' links |
| Verify | `core/database/graphcache/verify.go` `Verify` | Nodes, kinds, displays, `verifyMemos`, Evidence graphs. The match index, alias index and name format are not checked |
| Evidence-graph memos | `core/database/graphcache/evidence_graph.go` `loadedEvidenceGraph.memos`, `EvidenceGraphMemo`, `SetEvidenceGraphMemo` | Dropped with the graph, by `Drop`, and by `dropDerived` |
| Evidence-graph memo users | `core/database/evidencelayer/evidencelayer.go` `Load` (key `memoKey + namePattern`); `core/database/promotealign/canon.go` `loadCandidatesMemo` (key `candidatesMemoKey + namePattern`, valid for one `matchIndex.gen`) | The layer reads the vocabulary, terms and the name pattern key, not the alias index; the candidates memo ignores itself when the match index generation changes |
| Kinship slots | `core/database/canonlinks/reader.go` `kinshipSlot`, `forSlot`, `siblingSlot`, `init`, `deps`, `check`, `checkFor`, `checkSiblings`, `Reader.pivot`, `Reader.answer`, `Reader.For` | Three `NodeMemo`s; `kinship` with `Deps` (roots and their one-hop neighbours) and `DropOnVocabulary`; `for` and `siblings` with `DropOnVocabulary` and `DropOnHandleCommit` |
| Match index | `core/database/promotealign/index.go` `matchIndex`, `matchIndexOf`, `loadShortlist`, `Warm`, `indexGen`; `profiles.go` `newProfiler`, `profiler.keep` | Stored with `g.SetMatchIndex`; `keep` re-sets it to hold profiles. No Check |
| Alias index | `core/database/namealiases/namealiases.go` `Warm`, `Invalidate`, `Rules(g, q, base)` | `Warm` loads tables and calls `SetAliasIndex`; `Rules` reads the warm index or, on a cold graph, loads from `q` and builds without storing. `Invalidate` has no caller |
| Name format | `core/database/nameformats/nameformats.go` `Warm`, `Invalidate`, `LoadFor(g, q)`, `Load(q)` | Same shape as the alias index. `Invalidate` has no caller |
| Value rules | `core/database/valuerules/valuerules.go` `Load(g, q)` | Calls `nameformats.LoadFor(g, q)` and `namealiases.Rules(g, q, …)`; "g may be nil" |
| Callers of the three readers | `core/database/promotealign/promotealign.go` `loadConfig(g, q)` (and its callers in `index.go` `Warm`, `promotealign.go` `loadInputs`); `core/database/persondetail/recordlinks.go:84`; `core/database/conclusionsummaries/compose.go:388` `nameformats.LoadFor(g, g)`; `api/ffi/handlers/nameformat.go` `GetNameFormat` | Pass a graph and a querier |
| Session warm | `core/catalogsession/session.go` `warmGraph`, `WarmOnOpen`, `runLocked` | `warmGraph` runs `namealiases.Warm`, `nameformats.Warm`, `MarkWarm`; `runLocked` runs `warmGraph` again when `fn` succeeded and `WarmNeeded()` |
| `canonicalgraph` | `core/database/canonicalgraph/canonicalgraph.go`, `canonicalgraph_test.go`, `export_test.go` | `Walk`, `Targets`. Only caller: `core/database/conclusionsummaries/inputs_test.go` `TestInputsMatchIsolatedComposition` (`assoc` helper, line 163) |
| Tests | `core/database/graphcache/graph_test.go` `TestDropLeavesAliasIndex`, `TestDropLeavesNameFormat`, `TestMatchIndexDrops`, `TestNodeMemoDropsOnAnyCommit`, `TestNodeMemoDropsOnCommit`, `TestWarmNeeded`, the `memoNear`/`memoVocab`/`memoAll` vars; `evidence_graph_memo_test.go` `TestEvidenceGraphMemoDroppedWithSource`; `core/catalogsession/warm_test.go` `TestDoWarmsNameFormat`, `TestDoWarmsMatching`, `TestDoRewarms`, `TestOpenSurvivesABadValue`; `core/catalogsession/alias_cache_test.go` `TestDoKeepsAliasTablesForTheOpenProject`; `core/database/promotealign/promotealign_test.go` (line 90), `memo_test.go` `TestProfilesLiveWithTheMatchIndex`, `export_test.go` `MatchIndexGenForTest` and the `MatchIndex` check at line 34; `core/database/valuerules/valuerules_test.go` (`Load(nil, db)` at lines 44, 69); `core/database/conclusionsummaries/conclusionsummaries_test.go:865`; `core/database/evidencelayer/evidencelayertest/evidencelayertest.go:310`; `core/database/canonlinks/reader_test.go` (`hasMemo`, direct `SetMemo` at 643 and 1124) | Hold today's drop rules; rewritten to the registry's API with the same expectations except where a Decision says otherwise |
| Large fixture | `core/database/promotealign/perf_test.go` `newLargeFixture(t *testing.T)`, `TestLargeSourceBudgets`; `bench_test.go` `largeShape` (4,000 persons fixed) | No memory measurement per cache; no 16,000-person shape |
| Guides | `.cursor/rules/catalog-queries.mdc` lines 17–18; `.agents/skills/use-catalog-session/SKILL.md` line 62; `docs/promote-graph-alignment.md` lines 158, 165–169, 241; `.agents/skills/review-pr/lenses/maintainability.md` Known traps | Name `canonicalgraph.Walk`, `NodeMemo`/`RegisterNodeMemo`, `WarmNeeded`, `SetAliasIndex`, `SetNameFormat` and "S9-28 caches it as canonical edges" |

**Precedent.** The registration-from-the-owner pattern is today's `RegisterNodeMemo` from `canonlinks`' `init` and `graphcache.SetTruth` / `SetDisplayCheck` / `SetEvidenceGraphLoader`: `graphcache` never imports the owner. Copy that: each owner registers its entry in a package-level `var`, and `graphcache` holds `any`. Do not copy `NodeMemo`'s flags: every rule that was a flag is now either a Read or the owner's Apply.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The registry lives in `core/database/graphcache/cache.go` *(new)*, which replaces `memo.go`. Names: `Entry`, `Slot`, `Register`, `Lifetime` (`LifeGraph`, the zero value; `LifeCatalog`), `VocabularySlot` (a reserved `Slot` naming the typed vocabulary in Reads) | One file owns the mechanism; `Entry` and `Slot` say what they are once node memos are not the only kind | Keeping `NodeMemo`/`MemoSlot` names — they name one of two kinds; a new package — `Graph` holds the storage, so the methods belong on it |
| D-2 | Per-key fill stays the owner's `SetMemo` after it computes; the registry has no per-key Fill callback. A whole entry has an optional `Build func(g *Graph) (any, error)`: `Entry` returns the stored value, else runs Build, stores and returns it; with no Build it returns `ok == false` and the owner stores with `SetEntry` | A per-key Fill needs the reader's state (`kinship.Rules`, the deriver), which the registry cannot hold without `any`-typed parameters. The design's "Fill plus Deps" becomes "owner's Set plus Apply" | A `Fill func(g, id) (any, error)` — needs the rules the `Reader` was built with; a Build on the match index now — its build needs the profiler's shared map, which S16-20 removes |
| D-3 | `Apply func(g *Graph, slot Slot, roots [][]byte, set effects.Set) error` runs for every entry holding data, once per commit, after the reload, on every commit that is not `Structure` (including one with no `Handles`, then with nil roots). Nil Apply: commits leave the entry. Roots are exactly today's `dropMemos` roots | One place a commit reaches a cache; later entries (Source records, graph catalog) update on `SourceScopes` and `Reads` with no handle | Keeping today's early return when `Handles` is empty — S16-12 and S16-15 would each add a branch |
| D-4 | `Reads []Slot` lists `VocabularySlot` and earlier-registered slots. `SetEntry(s)` and `DropEntry(s)` empty every entry whose Reads include `s`, then their readers, transitively; a vocabulary commit does the same from `VocabularySlot`. A per-key `SetMemo` or `DropMemos` drops no reader | Replaces `DropOnVocabulary` and `dropDerived` with a declaration; key-level changes reach readers through each reader's own Apply, which sees the same roots | Key-level propagation — no S16-10 entry reads a per-key entry, and S16-16's composition from pivots reads at read time |
| D-5 | `Drop()` empties every `LifeGraph` entry and leaves `LifeCatalog` entries. `Register` panics when a `LifeCatalog` entry Reads a `LifeGraph` entry or `VocabularySlot` | The alias index and name format survive `Drop` today; a catalog-lifetime entry derived from graph data would go stale on `Drop` | A per-entry "survives Drop" flag — the lifetime says it |
| D-6 | `Cap int` is required (> 0) on a per-key entry: `SetMemo` of a new key when the entry holds `Cap` keys stores nothing. A whole entry has no Cap (`Register` panics on one); its size is bounded by what it is built from and recorded in `docs/catalog-caches.md` *(new)*. Caps: `canonlinks.kinship` 65,536; `canonlinks.for` 65,536; `canonlinks.siblings` 65,536 | Bounds a crafted catalog with no eviction code; at 16,000 persons the large fixture has 37,000 handles, so no cap binds at Scale and results do not move (a capped key is recomputed on read) | LRU eviction — more code, nothing needs it at Scale; a Cap on whole entries — a key count means nothing for one value |
| D-7 | `Check func(g *Graph, key []byte, stored any) error`; Verify runs it over every stored per-key value and every stored whole value (key nil). Nil Check: Verify skips. Verify never runs a Build | Today's `MemoCheck` shape, extended to whole entries | A separate whole-entry check type — one signature does both |
| D-8 | Accessors on `*Graph`: per key `Memo(slot, id)`, `SetMemo(slot, id, v)`, `DropMemos(slot, ids)`; whole `Entry(slot) (any, bool, error)`, `SetEntry(slot, v)`; either `DropEntry(slot)`. A nil graph returns zero values and no error. Calling a per-key accessor on a whole slot, or the reverse, panics | Keeps the per-key names `canonlinks` uses; `DropMemos` and `DropEntry` are what Applies call | One `Get(slot, key)` for both — a nil key meaning "whole" is easy to misread |
| D-9 | The three `canonlinks` slots become per-key entries in `reader.go`'s `init` (or package vars). `kinship`: Reads `VocabularySlot`, Apply drops the roots and `deps(g, roots)` (today's `Deps`), Check `check`. `for` and `siblings`: Reads `VocabularySlot`, Apply `DropEntry(slot)` when `len(roots) > 0`, Checks `checkFor` and `checkSiblings` | Today's behavior. The drop-on-any-handle rule lives in the owner until S16-16 deletes those two entries, so `graphcache` carries no `DropOnHandleCommit` mechanism meanwhile | Keeping a `DropOnHandleCommit` field until S16-16 — an old mechanism kept beside its replacement (rule "simpler code"); the map's S16-16 entry is reworded (STEP-8) |
| D-10 | The alias index is `namealiases.Slot` *(new)*, the name format `nameformats.Slot` *(new)*: whole, `LifeCatalog`, no Reads, Build = today's `Load` plus `buildIndex` / the `Loaded` value, reading through `g` (which reads its transaction on an `Over` graph). `namealiases.Rules(g, base)`, `nameformats.LoadFor(g)` and `valuerules.Load(g)` read `g.Entry`; their `q` parameters and cold branches go; a nil graph returns `errNoGraph` *(new, per package)*. `Warm` becomes `g.Entry(Slot)` returning its error; `Invalidate` becomes `g.DropEntry(Slot)` | One path for the session graph and a write's `Over` graph; the `Over` graph now builds once per write instead of on each call, with the same rows | Keeping `q` — a second way to the same value; dropping `Warm` from `WarmOnOpen` — an alias load error would stop failing the open, since `promotealign.Warm`'s error is ignored |
| D-11 | Neither the alias index nor the name format Reads the vocabulary. `warmNeeded`, `WarmNeeded`, `MarkWarm` and `runLocked`'s re-warm go; `warmGraph` keeps the two `Warm` calls at open | Today's re-warm after a vocabulary commit or `Drop` rebuilds the same values: alias tables and name formats read `name_alias_*` and `name_formats`/`project_settings`, not the property vocabulary, and both survive `Drop`. With D-10's lazy Build, nothing is ever missing on the session graph after open | Keeping the re-warm — work with no effect, and the flag couples five caches |
| D-12 | The match index is `promotealign.MatchIndexSlot` *(new, exported for Reads and tests)*: whole, `LifeGraph`, no Build, Reads `VocabularySlot`, `namealiases.Slot`, `nameformats.Slot`; Apply `DropEntry(slot)` when `len(roots) > 0`; Check `checkMatchIndex` *(new)*. `matchIndexOf` reads `g.Entry`, `loadShortlist` and `keep` store with `g.SetEntry`. `ResetMatchIndexForTest` goes; tests call `g.DropEntry(promotealign.MatchIndexSlot)` | Today's drops (handle commit, vocabulary, alias index, name format, `Drop`) as declarations | A Build for it (D-2); moving `indexGen` or profiles — S16-20 |
| D-13 | `checkMatchIndex`: when `shortlist` is set, rebuild a shortlist with a fresh profiler through `buildShortlist` *(new, split out of `loadShortlist`'s loop, which then calls it)* and compare with `reflect.DeepEqual`; for each stored profile, rebuild it with a fresh profiler and compare. `gen` is not compared | Verify then covers every moved entry (map Check); S16-20's "Verify compares the patched entry with a fresh build" reuses it | Comparing only counts — misses a wrong key |
| D-14 | Alias index Check: `Load(g)` + `buildIndex` and `reflect.DeepEqual` with the stored index. Name format Check: `Load(g)` and `reflect.DeepEqual` with the stored `Loaded` | Same shape as every other Check | — |
| D-15 | `dropDerived` goes. Evidence-graph memos are emptied on a vocabulary commit (a loop in `OnCommit`'s vocabulary branch) and with their graph and `Drop` as today; an alias index or name format change no longer empties them | The layer reads the vocabulary, terms and the name pattern string, which is in its key; the candidates memo is keyed by name pattern and ignored when the match index generation changes, which a new alias index or name format causes by dropping the match index (D-12). Results do not move | Registering Evidence-graph memos as an entry — they live and die with the loaded graph; S16-13 gives them the records tier |
| D-16 | `canonicalgraph` is deleted. `TestInputsMatchIsolatedComposition`'s `assoc(h, from, to)` becomes `assocBetween(t, db, a, b []byte) []byte` *(new, in `inputs_test.go`)*: `SELECT x.entity_id FROM auto_reconciler_values x JOIN auto_reconciler_values y ON y.entity_id = x.entity_id WHERE x.value_entity_id = ? AND y.value_entity_id = ? AND x.rank = 1 AND y.rank = 1 AND x.reason = 'kept' AND y.reason = 'kept'`, failing unless exactly one row | The oracle stays independent of the cache it checks; in that world each pair has one association | Using `graphcache.Follow` — the test checks what the cache records, so the oracle must not be the cache |
| D-17 | `BenchmarkCacheSizes` *(new)* in `core/database/promotealign/bench_test.go`, with sub-benchmarks `persons=4000` and `persons=16000`. `newLargeFixture` takes `(tb testing.TB, persons int)` and `largeShape` a `persons int`; `TestLargeSourceBudgets` passes 4,000. The benchmark measures `runtime.MemStats.HeapAlloc` after `runtime.GC()` before and after each fill, in this order, and reports each delta with `b.ReportMetric(…, "<name>-B")`: `nodes` (`g.Kind` of person, event, place), `aliasIndex` + `nameFormat` + `matchIndex` (`catalogsession.WarmOnOpen`, each Entry measured by dropping and re-filling it alone), `kinship` + `for` + `siblings` (`canonlinks.NewReader(g, matchregistry.Default().Kinship)` then `Relationships(id, canonlinks.Options{Siblings: true})` for every person), `evidence` (a cold proposal on Source A), `displays` (the "compose every summary" call). Reported, not gated | One reusable measure for the 40 MB budget; each later PR adds its fill | A gated budget — rule "Measure growth" binds S16-11 on; a separate fixture — the S11-21 fixture is the batch's Scale |
| D-18 | `docs/catalog-caches.md` *(new)*: one table for registry entries, one for typed `Graph` fields (nodes, holders, kinds, vocabulary, labels, displays, activity, Evidence-graph store with its memos) with a "why typed" column, one for Swift caches; columns Cache, Owner (path · symbol), Key, Fill, Updated by (the `effects.Set` field or Read), Lifetime, Cap, Size at 4,000 / 16,000. Each later Batch 2 PR edits its row | Goal bar 15 | Listing caches in `cache-architecture.md` — that file is deleted at spike close; a domain doc lasts |

**Rules this plan bends.** None. The map's S16-16 scope says "`DropOnHandleCommit` and `dropHandleMemos` go" there; D-9 removes those names here and leaves S16-16 deleting the two entries and their Apply. STEP-8 rewords the map entry; the rule "Every cache names the signal that updates it" already allows the behavior until S16-16.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `graphcache/cache.go` `Entry`, `Slot`, `Register`, `Lifetime`, `VocabularySlot` *(new)* | type / func / file | `core/database/graphcache/memo.go` `NodeMemo`, `MemoSlot`, `RegisterNodeMemo` | It is that mechanism extended and renamed; `memo.go` is deleted (searched `Register`, `Slot`, `memo` in `core/database`) |
| `Graph.Entry`, `SetEntry`, `DropEntry`, `DropMemos` *(new)* | func | `MatchIndex`/`SetMatchIndex`, `AliasIndex`/`SetAliasIndex`, `NameFormat`/`SetNameFormat`, `dropMemos` | Those six typed accessors are deleted; four generic ones replace them |
| `namealiases.Slot`, `nameformats.Slot`, `promotealign.MatchIndexSlot` *(new)* | var | `canonlinks` `kinshipSlot` | Same pattern, exported because other entries Read them |
| `checkMatchIndex`, `buildShortlist` *(new)* | func | `loadShortlist` | `buildShortlist` is `loadShortlist`'s loop split out so the Check and the load share it |
| Alias index and name format Checks *(new, unexported)* | func | `canonlinks.check` | Same shape |
| `errNoGraph` in `namealiases`, `nameformats` *(new)* | var | `evidencelayer.errNilGraph` | Package-local, like it |
| `assocBetween` *(new, test)* | helper | `canonicalgraph.Walk` | The package is deleted; one query is the oracle |
| `BenchmarkCacheSizes` *(new, test)* | benchmark | `BenchmarkProposeColdCity` heap metric | Reports per cache, at two sizes |
| `docs/catalog-caches.md` *(new)* | doc | `cache-architecture.md` (spike design) | Spike docs are deleted at close; this lasts |

**Builds on.** The owner-registers pattern (`RegisterNodeMemo`, `SetTruth`); `OnCommit`'s root computation unchanged; the S11-21 large fixture.
**Deletes.** `core/database/graphcache/memo.go` (`NodeMemo`, `MemoSlot`, `MemoCheck`, `MemoDeps`, `RegisterNodeMemo`, `dropMemos`, `dropVocabularyMemos`, `dropHandleMemos`, `verifyMemos`); in `graph.go` the fields `matchIndex`, `matchIndexOK`, `warmNeeded`, `aliasIndex`, `aliasIndexOK`, `nameFormat`, `nameFormatOK` and `MatchIndex`, `SetMatchIndex`, `ResetMatchIndexForTest`, `dropMatchIndex`, `dropDerived`, `WarmNeeded`, `MarkWarm`, `AliasIndex`, `SetAliasIndex`, `NameFormat`, `SetNameFormat`; `catalogsession.runLocked`'s re-warm; the cold branches of `namealiases.Rules` and `nameformats.LoadFor`; the `q` parameters of `Rules`, `LoadFor`, `valuerules.Load` and `promotealign.loadConfig`; `core/database/canonicalgraph/` (three files); `TestWarmNeeded`, `TestNodeMemoDropsOnAnyCommit` (folded into the registry tests).
**Next adds.** S16-12 and S16-15 register whole entries with a Build and an Apply on `SourceScopes` / `Reads`; S16-16 deletes the `for` and `siblings` entries; S16-18 registers per-key place entries; S16-20 gives the match index an Apply that patches instead of drops and reuses `checkMatchIndex`. Each adds a row to `docs/catalog-caches.md` *(new)* and a fill to `BenchmarkCacheSizes`.

### Reach

None — no new path into existing code, no new state reachable. The `Over` graph now stores the alias index and name format it builds (D-10); it is discarded with the write, as today.

### Scale

At the map's Scale (large fixture, 4,000 and 16,000 persons; crafted: a 20,000-child hub, an event with 500 participants): Apply per commit does what `OnCommit` does today. The kinship Apply deletes the roots and their one-hop neighbours from cached nodes: a commit naming the hub touches 20,001 keys, an event with 500 participants 501, as `Deps` does today. `for`, `siblings` and the match index each drop one map. Reads propagation walks at most the six registered entries. Lazy Build runs on the session graph only at open (D-11) and on an `Over` graph at most once per write (two small reads, as today's cold path did per call). No cap binds at Scale (D-6). Memory is unchanged; `BenchmarkCacheSizes` records it.

### Edges

- **Assumptions.** Roots passed to Apply equal today's `dropMemos` roots (AC-1.6, T-2). Reads form a DAG because a slot may Read only earlier slots (AC-1.8). A catalog-lifetime entry never reads graph data (D-5).
- **Neighbors.** A vocabulary commit drops the kinship slots and the match index and keeps the alias index and name format (T-1, T-9). A Source vocabulary commit (`ListSourceVocabulary`) drops nothing (T-9, from `TestWarmNeeded`'s last case). A first load drops nothing (T-2).
- **Removed inputs.** An Evidence-graph memo now outlives an alias index or name format change; it does not read them (D-15, T-10).
- **Real data.** None — no table- or seed-driven rule changes.
- **Untrusted size.** Per-key entries are capped (D-6, T-3); whole entries are bounded by the catalog they are built from.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` | `cache.go` *(new)*, `memo.go` deleted; `graph.go` fields and methods per **Deletes**, `entries map[Slot]any` *(new: per key `map[string]any`, whole the value)* on `Graph`; `clear` empties `LifeGraph` entries; `commit.go` vocabulary branch: `DropLabels`, vocab reset, Evidence-graph memos emptied, readers of `VocabularySlot` dropped; Apply loop after the reload, also on a commit with no handles; `verify.go` runs every entry's Check |
| `core/database/canonlinks` | `reader.go`: three per-key entries per D-9 |
| `core/database/namealiases`, `nameformats` | `Slot`, Build, Check, `Warm`, `Invalidate`, `Rules(g, base)` / `LoadFor(g)` per D-10, D-14 |
| `core/database/valuerules` | `Load(g)` |
| `core/database/promotealign` | `MatchIndexSlot`, `buildShortlist`, `checkMatchIndex`; `loadConfig(g)`; `matchIndexOf`, `loadShortlist`, `keep` through the registry |
| `core/database/persondetail`, `conclusionsummaries`, `evidencelayer/evidencelayertest` | Call sites of the readers |
| `core/catalogsession` | `warmGraph` without `MarkWarm`; `runLocked` without the re-warm; comments of `Do`, `WarmOnOpen` |
| `core/database/canonicalgraph` | Deleted |

```go
// cache.go sketch
type Slot int
const VocabularySlot Slot = -1
type Lifetime int
const (LifeGraph Lifetime = iota; LifeCatalog)
type Entry struct {
	Name     string
	PerKey   bool
	Build    func(g *Graph) (any, error)                                   // whole only
	Apply    func(g *Graph, slot Slot, roots [][]byte, set effects.Set) error
	Reads    []Slot
	Lifetime Lifetime
	Check    func(g *Graph, key []byte, stored any) error
	Cap      int // per key only, > 0
}
func Register(e Entry) Slot
```

Errors: `namealiases`/`nameformats` `errNoGraph` when the graph is nil.

### FFI

None — `api/ffi/handlers/nameformat.go` `GetNameFormat` drops its `db` read and calls `nameformats.LoadFor(c.Graph())`; the wire does not change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — The registry.**
Files: `core/database/graphcache/cache.go` *(new)*, `memo.go` (delete), `graph.go`, `commit.go`, `verify.go`, `cache_test.go` *(new)*, `graph_test.go`. Skill: `add-catalog-query` (for package layout only).
Do: write D-1 – D-8. Move `Memo`, `SetMemo`, `Cached` into `cache.go`. `OnCommit`: drop `dropMatchIndex`/`dropHandleMemos`; after the reload call the Apply loop with `roots` and `set`; on a commit with no handles call it with nil roots before returning. Vocabulary branch per the **Go core** row. Verify runs Checks. Leave `matchIndex`, `aliasIndex`, `nameFormat` and their accessors in place for now so the tree builds. Move the `memoNear`/`memoVocab`/`memoAll` tests into `cache_test.go` as T-1 – T-6 with registered test entries; `TestNodeMemoDropsOnAnyCommit` folds into T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/...` passes, T-1 – T-6 included.

**STEP-2 — Kinship entries.**
Files: `core/database/canonlinks/reader.go`, `reader_test.go`.
Do: D-9. `deps` becomes the kinship Apply's body (drop roots and `deps(g, roots)` through `g.DropMemos`). Rename types in tests.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/... ./core/database/persondetail/... ./core/database/promotealign/...` passes; T-7.

**STEP-3 — Alias index and name format.**
Files: `core/database/namealiases/namealiases.go`, `namealiases_test.go`; `core/database/nameformats/nameformats.go`, `nameformats_test.go`; `core/database/valuerules/valuerules.go`, `valuerules_test.go`; `core/database/promotealign/promotealign.go`, `index.go`, `memo_test.go`; `core/database/persondetail/recordlinks.go`; `core/database/conclusionsummaries/compose.go`, `conclusionsummaries_test.go`; `core/database/evidencelayer/evidencelayertest/evidencelayertest.go`; `api/ffi/handlers/nameformat.go`; `core/catalogsession/session.go`, `warm_test.go`, `alias_cache_test.go`; `core/database/graphcache/graph.go`, `graph_test.go`. Skill: `use-catalog-session`.
Do: D-10, D-11, D-14. Delete `aliasIndex*`, `nameFormat*`, `AliasIndex`, `SetAliasIndex`, `NameFormat`, `SetNameFormat`, `warmNeeded`, `WarmNeeded`, `MarkWarm` from `graph.go`, and `TestDropLeavesAliasIndex`, `TestDropLeavesNameFormat`, `TestWarmNeeded` (their cases move to T-1, T-8, T-9). `valuerules_test.go` passes `graphcache.Over(db)`. `TestDoRewarms` becomes T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes.

**STEP-4 — Match index; `dropDerived` goes.**
Files: `core/database/promotealign/index.go`, `profiles.go`, `export_test.go`, `promotealign_test.go`, `memo_test.go`; `core/database/graphcache/graph.go`, `commit.go`, `graph_test.go`, `evidence_graph.go` (doc comments), `evidence_graph_memo_test.go`; `core/catalogsession/warm_test.go`. Skill: `add-match-rule` (home of the match index only).
Do: D-12, D-13, D-15. Delete `matchIndex*`, `MatchIndex`, `SetMatchIndex`, `ResetMatchIndexForTest`, `dropMatchIndex`, `dropDerived`; `TestMatchIndexDrops` becomes T-11 in `promotealign`. Update `EvidenceGraphMemo`'s comment and T-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/... ./core/catalogsession/...` passes.

**STEP-5 — `canonicalgraph` goes.**
Files: `core/database/canonicalgraph/` (delete), `core/database/conclusionsummaries/inputs_test.go`.
Do: D-16.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/...` passes; `rg -n canonicalgraph core api` has no hit.

**STEP-6 — Measure.**
Files: `core/database/promotealign/bench_test.go`, `perf_test.go`.
Do: D-17.
Check: `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench BenchmarkCacheSizes -benchtime 1x ./core/database/promotealign/` prints a `<name>-B` metric for every cache at both sizes; `TestLargeSourceBudgets` passes. Paste the output in the PR description.

**STEP-7 — The cache table.**
Files: `docs/catalog-caches.md` *(new)*.
Do: D-18. Go rows from this PR's registry and the typed fields in `core/database/graphcache/graph.go`; sizes from STEP-6. Also a row for `core/catalogsession/session.go` `reg` (held sessions, one project at a time). Swift rows: one per `CatalogQueryKey` kind from `CatalogQueryRegistry.swift` `CatalogQueryKey.Kind.descriptor` (policy, follows, patches, the 50-hidden-entity cap `WorkspaceSession.maxHiddenEntities`); `WorkspaceSession.evidenceGraphIndex`, `WorkspaceSession.nameFormat`, `CatalogCounts`, `ThumbnailCache.shared` (100 entries, LRU, no invalidation), and the single-entry `Memo` holders (`SourcesModel.visibleRowsMemo`, `SourcesModel.searchKeyCache`, `CitationComposerContext.memo`, `EvidenceGraphModel.snapshotMemo`, `CitationComposerModel.depictionPreviewMemo`); sizes "not measured" for Swift. The Go table notes that alias tables and name formats report no effects (`core/database/effects/registry.go`, `None: true`), so their entries change only through `Invalidate`. Link it from `docs/promote-graph-alignment.md` §8 and from `cache-architecture.md` § One cache registry.
Check: every registered `Slot` (`rg -n 'graphcache.Register\(' core`) and every `Graph` field in `graph.go` has a row; `python3 scripts/check-plan-docs.py` passes.

**STEP-8 — Guides.**
Files: `.cursor/rules/catalog-queries.mdc`, `.agents/skills/use-catalog-session/SKILL.md`, `docs/promote-graph-alignment.md`, `.agents/skills/review-pr/lenses/maintainability.md`, `docs/deployment-plan/spike-16/cache-architecture.md`, `docs/deployment-plan/spike-16/deployment-plan.md`.
Do: `catalog-queries.mdc` line 17 drops `canonicalgraph.Walk` (a walk past the cache reads SQL in its own package, like `searchindex`'s input lookups); line 18 says a derived cache registers a `graphcache.Entry` with its Apply, Reads, lifetime, Check and Cap, and never adds a field on `Graph`; links `docs/catalog-caches.md` *(new)*. `use-catalog-session` line 62: the warm list without `WarmNeeded`, the readers' new signatures, `Invalidate` drops the entry. `promote-graph-alignment.md`: line 158's "S9-28 caches it as canonical edges" → the graph cache holds each node's links (`graphcache.Node.Links`); lines 165–169 describe the match index as a registry entry reading the vocabulary, alias index and name format, and the open-time warm without the re-warm; line 241 unchanged in substance. Maintainability lens Known traps: "A cache that empties itself on any write, or a new drop rule written in `OnCommit` instead of an entry's Apply on the roots and Reads — `docs/catalog-caches.md`, rule 'Every cache names the signal that updates it'". `cache-architecture.md` § One cache registry: "Fill: per key (the owner's `SetMemo`, with an Apply in place of `Deps`) or whole (a lazy Build)". Map: S16-16 Ships "the `for` and `siblings` entries and their drop-on-any-handle Apply go"; S16-20's "dropping the match index on every handle commit" → "the match index's drop-on-any-handle Apply".
Check: the `rg` of AC-3.2 has no hit; `python3 scripts/check-guidance-names.py` passes.

**STEP-9 — Docs.**
Files: `completed.md`, `deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: links resolve; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period ("Hold the graph's derived caches in one registry of entries."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphcache/cache_test.go` `TestEntryReads` *(new)* | Table: SetEntry(A) / vocabulary commit / DropEntry(B) / `Drop()` / `ListSourceVocabulary` commit over entries A–E of AC-1.1 → the held set after each | AC-1.1, AC-1.2, AC-1.3, D-4, D-5 |
| T-2 | Go | `cache_test.go` `TestEntryApply` *(new)* | Recording Apply; association commit, merge, delete of a loaded node, Source-only commit, first load → roots and set per AC-1.6; first load calls nothing | AC-1.6, D-3 |
| T-3 | Go | `cache_test.go` `TestEntryCap` *(new)* | Cap 2: k1, k2, k3, then k1 again → k3 absent, k1 replaced | AC-1.4, D-6 |
| T-4 | Go | `cache_test.go` `TestEntryBuild` *(new)* | Counting Build; two `Entry` on the session graph, one on `Over(tx)`; a failing Build → counts 1 and 1 (tx read); error returned, nothing stored | AC-1.5, D-2 |
| T-5 | Go | `cache_test.go` `TestEntryVerify` *(new)* | `t.Setenv("PROVENENCIA_GRAPH_VERIFY", "1")`; rejecting Check on a per-key and a whole entry; a commit → error | AC-1.7, D-7 |
| T-6 | Go | `cache_test.go` `TestRegisterRejects` *(new)* | Each invalid Entry of AC-1.8 → panic (registered on a copy of the registry the test restores) | AC-1.8 |
| T-7 | Go | `core/database/canonlinks/reader_test.go` (existing memo cases, renamed API) | Participation commit naming an event → its and neighbours' pivots gone, unrelated pivot kept, `for`/`siblings` empty | AC-2.3, D-9 |
| T-8 | Go | `core/catalogsession/warm_test.go` `TestNameFormatDropsItsReaders` *(new)* | Open project, a proposal and a Family read run; `SetEntry(nameformats.Slot, loaded)` → match index absent; alias index (0 SELECTs for `valuerules.Load`) and a stored kinship pivot present | AC-2.1, D-12 |
| T-9 | Go | `core/catalogsession/warm_test.go` `TestDoKeepsWarmEntries` (was `TestDoRewarms`) | Cases AC-2.2 1–4, plus a `ListSourceVocabulary` commit → 0 SELECTs for `valuerules.Load` each; match index absent in the four, present after the Source vocabulary commit; a drop leaves the person kind unloaded | AC-2.2, D-11 |
| T-10 | Go | `core/database/graphcache/evidence_graph_memo_test.go` `TestEvidenceGraphMemoDroppedWithSource` | The `SetAliasIndex` case becomes: a test whole entry with `LifeCatalog` is Set → memo kept; a vocabulary commit → memo gone | AC-2.5, D-15 |
| T-11 | Go | `core/database/promotealign/memo_test.go` `TestMatchIndexDrops` *(moved from graphcache)* | Stored match index; Source-only commit → kept; handle commit → gone; refill; vocabulary commit → gone; refill; `SetEntry(namealiases.Slot)` → gone; refill; `Drop()` → gone | AC-2.4, D-12 |
| T-12 | Go | `core/database/namealiases/namealiases_test.go` `TestVerifyChecksAliasIndex` *(new)*, `core/database/nameformats/nameformats_test.go` `TestVerifyChecksNameFormat` *(new)*, `core/database/promotealign/memo_test.go` `TestVerifyChecksMatchIndex` *(new)* | Verify on; stored value equal to a fresh build → commit passes; an alias index with no tables over a catalog with one / a `Loaded` with another key / a match index whose shortlist lacks a handle → commit fails naming the entry | AC-2.6, D-13, D-14 |
| T-13 | Go | `core/database/conclusionsummaries/inputs_test.go` `TestInputsMatchIsolatedComposition` | Same cases, `assocBetween` in place of `canonicalgraph.Walk` | AC-2.7, D-16 |
| T-14 | Go | `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes` *(new)* | 4,000 and 16,000 persons → a `-B` metric per cache | AC-3.1, D-17 |
| T-15 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` | Unchanged budgets, fixture at 4,000 | AC-2.7 |
| T-16 | Sweep | STEP-8 check: the `rg` of AC-3.2 over `core api .agents .cursor docs/*.md` | → no hit | AC-3.2 |

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
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status, and the S16-16 and S16-20 scope rewording (STEP-8); completed.md: one short paragraph of what shipped
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
- `BenchmarkCacheSizes` at 16,000 persons does not finish within 10 minutes on the implementer's machine.

## 12. Open questions

None.

## 13. PR

**Title:** S16-10 — One cache registry
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; the `BenchmarkCacheSizes` output; `Plan: docs/deployment-plan/spike-16/prs/S16-10-one-cache-registry.md`.
