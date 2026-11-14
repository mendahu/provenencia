# S16-32 — Promote re-ranks only the Subjects a write moved

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 2d — Promote
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine change; Promote's suggestions read the same as today)
**Board:** —
**Depends on:** S16-20 (landed [#827](https://github.com/mendahu/provenencia/pull/827)), S16-13 (landed)
**Blocks:** —
**Branch:** `s16-32-narrow-candidate-rerank`
**Version bump:** none — no change a researcher can see, and no wire or catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) (the graph cache, its registry and Verify), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (the layer package `core/shortlist`)
**Rules that bind:** `.cursor/rules/go-tests.mdc` (table-driven tests), `.cursor/rules/catalog-queries.mdc` (a derived cache reads and writes only through its registry handle; STEP-3 adds `Keyed.Update` to the handle's methods it names), `.cursor/rules/match-rules.mdc` (`core/shortlist` stays pure)
**Context:** [design § Promote](../cache-architecture.md#promote), [`docs/catalog-caches.md`](../../../catalog-caches.md), [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §8, [S16-20 plan](S16-20-match-index-apply.md) (its **Landed** note and **Next adds**), working rules: "The point of this batch is simpler code", "Cache entities and results per key; compose answers", "Per-graph records and memos are registry entries keyed by graph", "Every cache names the signal that updates it, and updates exactly", "Apply states its fan-out", "Orders are defined and results do not move", "Memory is budgeted", "Measure growth"
**Planned:** 2026-11-14 against `main` @ `36d6ec2`
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

Promote remembers each Evidence graph's ranked candidates per Subject (`promotealign.candidates`). Since S16-20 a handle commit patches the match index instead of rebuilding it, but every commit that changes any profile still empties every graph's remembered candidates, so the next proposal re-ranks every Subject of its Source: 3.2 times a warm proposal at 4,000 persons and 5.3 times at 16,000 (measured at S16-20's planning). This PR makes the match index's Apply hand its readers what moved instead of emptying them: the keys each changed handle was filed under before and after the commit. Each remembered Subject keeps its own lookup keys, and only the Subjects whose lookup keys meet a moved key are forgotten; the next proposal ranks just those again, through the path that already ranks a Subject the memo lacks.

**Why now.** Bar 10's second clause ("the first proposal after a handle commit is no slower than 1.5 times a warm one"), moved here from S16-20 by the user on 2026-11-13; the map's rule "Every cache names the signal that updates it, and updates exactly" — `candidates` is the last entry that empties itself on any profile change.

**What stays.** Every proposal, candidate list, row, alternative and connection reads as today (the goldens, `TestAlignInvariants`, the warm-equals-fresh tests). The match index, its Build, its Check and what its Apply re-profiles stay as S16-20 left them. A vocabulary, alias or name format change, and `Drop`, still empty the candidates (through the match index's `DerivedFrom`). A write that names a graph (`Set.EvidenceGraphs`) still drops that graph's memo. No SQL, wire or Swift change.

## 2. User stories

**US-1.** As a researcher, when I save a change to a person who is no candidate of the Source I am promoting, the next Promote proposal of that Source is as fast as a warm one, because nothing it ranked moved.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `newWarmFixture`; graphs B (`srcB`: "John Smith" and a 1817 birth) and C (a new Source with "Mary Jones") each proposed; then a person "Zebedee Quarrington" filed on a third Source (`filedBirth`) | Every Subject of B and C is still held in the candidates memo (`SubjectRankedForTest`); with `PROVENENCIA_GRAPH_VERIFY=1` the commit's Verify passes; the next proposal of B equals one on a dropped graph |
| AC-1.2 | Large fixture at 4,000 persons (`TestLargeSourceBudgets` "after a handle commit") | The proposal of Source A after a person is added to the other Source takes at most 1.5 times a warm proposal measured in the same subtest |
| AC-1.3 | Large fixture at 4,000 and 16,000 persons (`BenchmarkLargeSourcePropose`) | `warm-ms`, `after-commit-ms` and `after-commit-x` (their ratio) are recorded before and after in the ship log; `after-commit-x` is at most 1.5 at both sizes |

**US-2.** As the engine, after a commit that changes a handle, only the remembered Subjects that could rank that handle differently are ranked again.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | As AC-1.1's setup; the member's name Observation edited from "John Smith" to "Jon Smyth" (`renameMember`) | B's person Subject is not held; B's birth Subject and C's "Mary Jones" Subject are held; Verify passes; the next proposal of B equals one on a dropped graph |
| AC-2.2 | As AC-1.1's setup; the member's birth event date cited as 1818 (`Cite` on `memberBirth`) | B's birth Subject and B's person Subject are not held (the event's profile changed, and the member's lifespan with it); C's "Mary Jones" Subject is held; Verify passes; the next proposal of B equals one on a dropped graph |
| AC-2.3 | As AC-1.1's setup; a person "John Smith" filed on a third Source (`filedBirth`) | B's person Subject is not held; the next proposal of B offers the new handle as target or alternative (as `TestCandidatesMemoFollowsCommits` "a handle commit" checks); C's "Mary Jones" Subject is held |
| AC-2.4 | As AC-1.1's setup; the member's handle merged (`mergePerson`), and in another case the member unpromoted (`unpromote`) | B's person Subject is not held; Verify passes; the next proposal of B equals one on a dropped graph |
| AC-2.5 | As AC-1.1's setup; a vocabulary commit, a name format change | No graph's candidates are held (unchanged from today: `TestEntriesFollowSignals`) |

**US-3.** As the engine, a commit's work on the remembered candidates is bounded by what the cache holds, and it does not reorder which graph leaves the cache next.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | 63 memos of 100 Subjects and one of 10,000 Subjects, 7 lookup keys each, held through a test hook; one forget with 20 moved keys | It runs in at most 100 ms (best of three), and forgets exactly the Subjects whose keys meet the moved keys |
| AC-3.2 | A keyed entry with three keys at its Cap; `Update` over every key; then a fourth key stored | The key that was least recently used before `Update` leaves; `Update` replaced each value |

## 3. Scope

| In | Covered by |
| --- | --- |
| `shortlist.Index.Filed` and `LookupKeys`; `Lookup` reads its keys through `LookupKeys` | STEP-2, D-1, D-2 |
| `graphcache.Keyed.Update`: each held value replaced, no use counted | STEP-3, AC-3.2, D-3 |
| The candidates memo keeps each Subject's kind and lookup keys beside its ranked ids | STEP-4, D-4 |
| The match index's Apply collects the moved keys and forgets the Subjects they meet, in place of `w.Set` | STEP-5, AC-1.1, AC-2.1 – AC-2.5, D-5, D-6, D-7 |
| The forget's fan-out stated and timed | STEP-6, AC-3.1, D-8 |
| Bar 10's 1.5× gate at 4,000 persons; figures before and after at both sizes | STEP-1, STEP-7, AC-1.2, AC-1.3, D-9 |
| Docs this makes false | STEP-8 |

| Out | Home |
| --- | --- |
| Alike-pair memos (B-11) | [`docs/ideas/promote-ranking.md`](../../../ideas/promote-ranking.md) |
| The canon walk, a per-handle canon view; bar 10's "three times faster warm" clause | S16-22 |
| `graphalign.CanonEdge` onto `kinship.Link`; `Connections` reading the place entries | S16-21 |
| Ranking a forgotten Subject again inside Apply (eagerly) | not planned (D-6) |
| Changing any other `TestLargeSourceBudgets` budget, `budgetAfterCommit` included | not planned |
| A per-memo inverted index from lookup key to Subjects | not planned (D-8: the forward scan is measured at Scale) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| The memo | `core/database/promotealign/canon.go` `candidatesMemo{bySubject map[string][][]byte}`, `candidates` | Each Subject's ranked handle ids by Subject id. Registered in `index.go`'s `init` after `MatchIndex`: `DerivedFrom: MatchIndex`, `Apply: graphcache.DropNamedGraphs`, `Check: checkCandidates`, `Cap: graphfill.MaxEvidenceGraphs` (64). "A stored memo is never changed" |
| Fill | `canon.go` `loadCandidatesMemo` | `candidates.Peek`; ranks only the Subjects of `todo` the memo lacks (`rankCandidates`) and `Set`s a new map holding the old and new rankings |
| Ranking | `canon.go` `rankCandidates(m, layer, subjects, cfg) map[string][][]byte` | `seedHandles(m, subjects)` then `graphalign.Candidates` per Subject; every Subject gets an entry, `[][]byte{}` for none |
| Reader | `canon.go` `loadCanon` (lines 96 – 109) | Reads `known[string(s.ID)]` as the Subject's ranked ids |
| Check | `canon.go` `checkCandidates` | Ranks the stored Subjects again from `evidencelayer.Fresh` and compares with `reflect.DeepEqual` against `stored.bySubject` |
| Seeds | `core/database/promotealign/index.go` `seedHandles` | `m.shortlist.Lookup(kind, s.Values)` per Subject, then `m.handles[id]`, sorted by Ref then ID |
| Apply | `index.go` `applyMatchIndex` | Re-profiles the roots and `p.readers(roots)`; `delete` + `m.shortlist.Remove(id)` for a handle that left; `m.shortlist.Put` for each live handle whose Subject is not `reflect.DeepEqual` to its old one; `changed` set by either; at the end `w.Set(g, m)` when `changed`, which empties every `candidates` memo (`Whole.Set` → `dropReaders`) |
| Shortlist | `core/shortlist/index.go` `Index`, `filing{kind, keys}`, `Put`, `Remove`, `unfile`, `Lookup` | `filed` holds each handle's kind and sorted index keys; `Put` returns early when the keys are unchanged; `Lookup` computes `Keys(...)`'s lookup list itself and unions `byKey[kind][k].ids` |
| Keys | `core/shortlist/shortlist.go` `Keys` | Index and lookup keys, each sorted and unique |
| Registry | `core/database/graphcache/cache.go` `Keyed` (`Get`, `GetMany`, `Peek`, `Set`, `Drop`, `Clear`, `SetCapForTest`), `keyedValues` (`items`, LRU list `front` … `back`, `index`), `applyEntries` | No method visits held values; `Peek` and `Set` count a use (`touch`) |
| Rule | `.cursor/rules/catalog-queries.mdc` (the "A derived cache registers a typed handle" bullet) | "reads only through its `Get`, `GetMany` or `Peek` (each counts as a use; a Check reads none of them)" |
| Tests | `core/database/promotealign/memo_test.go` `TestMatchIndexFollowsCommits` ("a commit that changes no profile"), `TestCandidatesMemoFollowsCommits`, `TestEntriesFollowSignals`, `TestVerifyLeavesEntriesAlone`, `TestMemoFollowsCommit` | Candidates follow commits by being emptied; a no-change commit keeps them; LRU unchanged by a Source-only commit |
| Tests | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` "after a handle commit", `largeFixture.proposalAfterCommit`, `largeFixture.warmProposal`, `TestMatchIndexApplyFanOut` | `proposalAfterCommit`: best of two, a person added to the other Source and promoted, then a timed proposal with no GC before it; `warmProposal` is `databasetest.Warm` (a GC before each of five runs) |
| Tests | `core/database/promotealign/bench_test.go` `BenchmarkLargeSourcePropose` | Reports `warm-ms` and `after-commit-ms` at 4,000 and 16,000 |
| Tests | `core/database/promotealign/export_test.go` `CandidatesHeldForTest`, `HoldCandidatesForTest`, `SetCandidatesCapForTest` | Peek and store whole memos |
| Tests | `core/shortlist/shortlist_test.go` `TestLookup`, `TestIndexPut`, `TestIndexPatchesEqualBuild`; `core/database/graphcache/cache_test.go` `TestKeyedEvictsLeastRecentlyUsed`, `TestKeyedInputs` | Index and registry behaviour |
| Docs | `docs/catalog-caches.md` rows `promotealign.candidates` and `promotealign.matchindex`; `docs/promote-graph-alignment.md` §8 "The match index" and "The candidates" bullets; `docs/deployment-plan/spike-16/cache-architecture.md` line 66 | "`DerivedFrom` `promotealign.matchindex`"; "…, then `Set`"; "dropped … whenever the match index changes"; "and the candidates, derived from it, are ranked again (S16-32 narrows that)" |

**Measured on the planning base** (scratch test, removed; Xeon 2.1 GHz, 4 cores): Source A's 100 Subjects have 677 lookup keys (259 person, 418 event), at most 8 per Subject, 7,225 bytes of key text. A scan of 64 memos of 10,000 Subjects with 7 keys each against a set of 20 moved keys takes about 200 ms (scratch program); 64 memos of 100 Subjects, about 2 ms by the same rate. S16-20's Landed figures (Apple M1): `warm-ms` 114.1 / 124.2, `after-commit-ms` 390.8 / 666.9 at 4,000 / 16,000.

**Precedent.** S16-20 ([#827](https://github.com/mendahu/provenencia/pull/827)): `applyMatchIndex`'s patch-in-place Apply and its tests (`TestMatchIndexFollowsCommits` with Verify on, `TestMatchIndexApplyFanOut` with a `// decided:` ceiling). Copy the test shapes. S16-16 ([#814](https://github.com/mendahu/provenencia/pull/814)) is the precedent for figures before and after in the ship log.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `core/shortlist/index.go`: `func (x *Index) Filed(id []byte) (kind vocabkey.SubjectType, keys []string, ok bool)` *(new)*: the kind and sorted index keys `id` is filed under, `ok` false when it is filed nowhere or `x` is nil. The keys are the index's own strings; the doc says the caller must not change them. `Put` and `Remove` keep their signatures | The caller needs a changed handle's keys both before and after, including when `Put` leaves the keys as they were but the values changed (D-5); a read accessor serves both and costs `buildMatchIndex`'s 30,233 `Put`s nothing | `Put` and `Remove` returning the keys that gained or lost the handle (S16-20's **Next adds**, the map's scope entry) — misses a handle whose values change and keys do not, which can move it into or out of a Subject's top `TopK`; returning old and new keys from `Put` — an allocation per handle at Build for no reader |
| D-2 | `core/shortlist/index.go`: `func (x *Index) LookupKeys(kind vocabkey.SubjectType, v subjectcompare.Values) []string` *(new)*: `Keys(kind, v, x.metas, x.rules, x.subject, x.value)`'s lookup list, nil for a nil index. `Lookup` calls it in place of its own `Keys` call; nothing else in `Lookup` changes. Doc on `Filed`: "A Lookup of kind finds a handle when one of its LookupKeys is among the keys the handle is Filed under in that kind." | The index owns the rules it keys with, so a Subject's stored keys are the ones `Lookup` reads | `promotealign` calling `shortlist.Keys` with the profiler's rules — a second copy of the index's rules |
| D-3 | `core/database/graphcache/cache.go`: `func (k Keyed[T]) Update(g *Graph, fn func(id []byte, v T) T)` *(new)*: for each held key, from the most to the least recently used, calls `fn` and stores its result in the item in place (`it.v = …`, then `kv.index(it.key, it.v)`), counting no use and keeping the use order. Nil graph or no data: nothing. Doc: "fn must not read or write this entry." `.cursor/rules/catalog-queries.mdc`'s bullet gains: "an Apply that rewrites held values uses `Keyed.Update`, which counts no use" | An Apply runs on every commit and must not reorder eviction (`TestVerifyLeavesEntriesAlone` holds that for Check); the registry owns the LRU list, so it owns the walk | `Peek` then `Set` per graph — two uses per commit, the eviction order would follow commits; an exported key list — the caller still needs a no-use read and write |
| D-4 | `canon.go`: `candidatesMemo{bySubject map[string]rankedSubject}`; `rankedSubject{kind vocabkey.SubjectType; keys []string; ids [][]byte}` *(new, unexported)*: the Subject's kind, its `m.shortlist.LookupKeys(kind, s.Values)`, and its ranked handle ids (`[][]byte{}` for none, as today). `rankCandidates` returns `map[string]rankedSubject`; `loadCandidatesMemo` returns it; `loadCanon` reads `.ids`; `checkCandidates` compares the whole map as today. `seedHandles` keeps calling `Lookup` | The keys are what a forget reads (D-5); the kind picks the postings map they meet; Check compares them with the ranking, so a stale key fails Verify | Storing the shortlisted handle ids instead — a handle that gains a key is in no stored list; recomputing keys at forget time — needs each Subject's values, which live in `evidencelayer.Layers`, a separate entry that may have evicted the graph; `seedHandles` taking keys to save a second `Keys` call — measured at a few microseconds per ranked Subject |
| D-5 | `index.go` `applyMatchIndex`: a `moved map[movedKey]bool` *(new type `movedKey{kind vocabkey.SubjectType; key string}`)*. For each handle that leaves (`!p.holds(n)` and held): add its `Filed` keys, then `delete` and `Remove`. For each live handle whose Subject changed: add its `Filed` keys, `Put`, add its `Filed` keys again. The `changed` flag and `w.Set(g, m)` go; the last line is `forgetCandidates(g, moved)`; the `w` parameter becomes `_`. The doc comment's last sentence becomes: "It hands the keys every changed handle was filed under, before and after, to forgetCandidates." | A Subject's ranking reads exactly the handles filed under its lookup keys (`Lookup`) and their Subjects (`graphalign.Candidates`): a handle can enter, leave or move in a list only if it was or is filed under one of them. The old and new keys of every changed handle are therefore the whole signal. The match index is patched in place, so `Set` stored the same pointer and only emptied the readers; this replaces the emptying with the exact one | Keeping `Set` and adding a narrow drop beside it — `Set` would still empty everything; storing the commit's moved keys on `matchIndex` for `candidates`' own Apply to read — per-commit scratch state in a cached value, which Check's `DeepEqual` against a fresh build would see; a registry mechanism for a whole entry to pass a change to its readers — one reader, one package |
| D-6 | `canon.go` `forgetCandidates(g *graphcache.Graph, moved map[movedKey]bool)` *(new)*: nothing when `moved` is empty; else `candidates.Update(g, …)`: for a memo with no Subject any of whose `keys` meets `moved` under its `kind`, return it as is; otherwise return a new `candidatesMemo` whose map, made with `make(map[string]rankedSubject, n)` so it is never nil, holds every other Subject (the stored map is never changed; `checkCandidates` compares with `reflect.DeepEqual` against `rankCandidates`'s non-nil map, so a memo forgotten to empty still passes Verify). The Subject is ranked again lazily by `loadCandidatesMemo` on the graph's next proposal. A memo left with no Subject stays held | Lazy re-rank reuses the one ranking path; a graph that is never proposed again costs nothing | Ranking inside Apply — Apply would read Evidence layers and rank at every commit for graphs nobody looks at; dropping the whole graph's memo — the map's Check asks that only the Subjects whose keys meet re-rank |
| D-7 | `candidates` keeps `DerivedFrom: []graphcache.Source{MatchIndex}`, `Apply: graphcache.DropNamedGraphs`, `Check`, `Cap`. Its declaration's doc and the `candidates` var doc say: "emptied when the match index is dropped (a vocabulary, alias or name format change, or `Drop`); a commit that changes a handle forgets the Subjects whose lookup keys meet the keys it moved (`forgetCandidates`, called from the match index's Apply); a write that names a graph drops that graph's memo" | With D-5 the match index is never `Set` by Apply, so `DerivedFrom MatchIndex` fires only when the index is dropped or rebuilt, which is when every ranking is stale; one declaration covers the vocabulary, the alias index and the name format through `dropReaders` | Declaring `DerivedFrom` the three sources directly (S16-20's **Next adds**) — three declarations that must match the match index's; a commit-time `Set` of the match index elsewhere (`EmptyShortlistForTest`) still empties the candidates, which is right for a replaced index |
| D-8 | The forget's cost is a forward scan: every held Subject's keys against `moved`, only on a commit that moved a key. Stated in §6 Scale and held by `TestCandidatesForgetFanOut` *(new, perf_test.go)* with ceiling `ceilingForgetCandidates = 100 * time.Millisecond` and a `// decided: S16-32 D-8` comment, over 63 memos of 100 Subjects and one of 10,000, 7 keys each (bar 2's Source with 10,000 Observations holds at most 10,000 Subjects; the batch's Scale averages 5 Observations a graph), through two hooks (D-11) | Simplest shape that is linear in what the cache holds and bounded by the entry's Cap of 64 graphs (rule "The point of this batch is simpler code") | An inverted index per memo from key to Subjects — a second structure to keep equal to the first, and Check would compare it too; the cost it saves is milliseconds at Scale |
| D-9 | Bar 10's gate. `largeFixture.proposalAfterCommit` runs `runtime.GC()` after the commit and before the timed proposal, and takes the best of three (today two), so it is measured as `databasetest.Warm` is. `TestLargeSourceBudgets` "after a handle commit" measures `warm := lf.warmProposal()` and `after := lf.proposalAfterCommit(&n)` once each, passes the same `after` to the existing `budgetAfterCommit` call, and calls `databasetest.WithinBudget(t, "proposal after a handle commit, against 1.5 times a warm one", after, warm*3/2)` beside the existing `budgetAfterCommit` call, with the comment `// decided: bar 10 (deployment-plan.md), moved to S16-32 by the user on 2026-11-13`. `BenchmarkLargeSourcePropose` reports `after-commit-x` = after ÷ warm | The map's Check names the 1.5× ratio at 4,000 and 16,000; the test gates 4,000 as every Promote budget does, the benchmark records both. Without the GC, a collection from the warm proposal's 120 MB of garbage lands inside a single timed run (comment above the "warm" subtest) | Gating 16,000 in a test — the suite runs the large fixture at 4,000 only (`TestLargeSourceBudgets`); a fixed millisecond budget — the CI runner and a Mac differ by three times, the ratio does not |
| D-10 | Tests. **Change:** `TestMatchIndexFollowsCommits` "a commit that changes no profile" also asserts B's person Subject is held (`SubjectRankedForTest`). **Add:** `TestCandidatesFollowMovedKeys` *(new, memo_test.go)*, T-1; `TestIndexFiled` *(new)* and `TestLookupKeys` *(new)* in `shortlist_test.go`, T-3, T-4; a `TestKeyedUpdate` *(new)* in `cache_test.go`, T-5; T-6, T-7 per §8. No existing test is removed | Each AC has a test; Verify (`PROVENENCIA_GRAPH_VERIFY=1`) checks every forget against a fresh ranking | — |
| D-11 | Hooks added to the existing `core/database/promotealign/export_test.go`: `SubjectRankedForTest(g *graphcache.Graph, graphID, subjectID []byte) bool` (the held memo of the graph holds the Subject; reads `candidates.Peek`); `HoldRankedForTest(g *graphcache.Graph, graphID []byte, subjects, keysEach int)` (stores a memo of `subjects` person Subjects; Subject i's id is `graphID[:8]` followed by `i` as 8 bytes big-endian; its keys are `fmt.Sprintf("name|w:%x-%d-%d", graphID[:4], i, j)` for j from 0 to keysEach−1, and its ranked ids `[][]byte{}`; the memo's map is made with `make`); `ForgetForTest(g *graphcache.Graph, kind string, keys []string)` (calls `forgetCandidates` with those keys); `HeldSubjectsForTest(g, graphID []byte) int` (the number of Subjects the memo holds, 0 when none) | The fan-out test times the forget alone, on synthetic memos of the stated size, without building 64 Sources | A production counter of re-ranked Subjects (the map Check's "counter") — the held-Subject probe shows the same thing (a Subject not held is ranked again on the next proposal) with no code in the engine |

**Rules this plan bends.** None. D-1 and D-7 differ from the shape S16-20's **Next adds** and the map's scope entry sketched ("`Put` and `Remove` return the keys"; "a changed handle is in its list"; "`candidates` drops its `DerivedFrom MatchIndex`"); neither is a working rule. Planning rewrote the map's S16-32 row, scope entry and Check, and the design's line ([`cache-architecture.md`](../cache-architecture.md) "Candidates re-rank only what moved"), to the rule as built (D-5).

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `shortlist.Index.Filed` *(new)* | method | `Index.filed` (unexported map) | The read accessor of that map |
| `shortlist.Index.LookupKeys` *(new)* | method | `shortlist.Keys`, `Index.Lookup` | `Lookup`'s own key step, exported so the caller stores what `Lookup` reads |
| `graphcache.Keyed.Update` *(new)* | method | `Keyed.Set` (counts a use), `keyedValues.touch` | Set touches the LRU list; Update is the no-use write an Apply needs |
| `rankedSubject` *(new)* | type | `candidatesMemo.bySubject`'s `[][]byte` | The same ids with the Subject's kind and keys beside them |
| `movedKey`, `forgetCandidates` *(new)* | type / func | `graphcache.DropNamedGraphs` | That drops whole graphs by an effects field; this forgets Subjects by the match index's moved keys |
| `TestCandidatesFollowMovedKeys`, `TestCandidatesForgetFanOut`, `TestIndexFiled`, `TestLookupKeys`, `TestKeyedUpdate` *(new)* | tests | `TestMatchIndexFollowsCommits`, `TestMatchIndexApplyFanOut`, `TestIndexPut`, `TestKeyedEvictsLeastRecentlyUsed` | New behaviour; they reuse `newWarmFixture`, `renameMember`, `mergePerson`, `unpromote`, `filedBirth` and the `build` helper |
| `SubjectRankedForTest`, `HoldRankedForTest`, `ForgetForTest`, `HeldSubjectsForTest` *(new)* | test hooks | `CandidatesHeldForTest`, `HoldCandidatesForTest` | Those read and store whole memos; these read one Subject and build sized memos |

**Builds on.** The cache registry (S16-10) and its keyed entries' LRU (S16-13); S16-20's patch-in-place Apply and the shortlist's `filed` map; the existing lazy ranking of the Subjects a memo lacks (`loadCandidatesMemo`).
**Deletes.** In `applyMatchIndex`: the `changed` flag, `w.Set(g, m)` and the use of `w`. The emptying of every graph's candidates on any profile change. Docs: the "whenever the match index changes" clause in `promote-graph-alignment.md` §8 and the `catalog-caches.md` row's "Updated by" text. No function loses its last caller (`Whole.Set` keeps `EmptyShortlistForTest`, `sources.Records`, `evidencegraphs.Index` and others).
**Next adds.** S16-22's view reads the same match index; a later memo keyed per Subject (alike pairs, B-11) can forget by the same `moved` keys.

### Reach

None — no new path into existing code. `forgetCandidates` runs only inside the match index's Apply, which runs only on the session graph.

### Scale

At the map's Scale (large fixture at 4,000 and 16,000 persons; a Source with 10,000 Observations; 64 held graphs, `graphfill.MaxEvidenceGraphs`):

- **Per commit.** The match index's Apply adds two `Filed` reads per changed handle (map lookups) and one entry in `moved` per key: a rename, about 10 keys; an event with 500 participants re-dated, the event's and 500 people's keys, about 4,000. Then one scan of every held Subject's keys: at the measured rate, 64 graphs the size of Source A (6,400 Subjects, 43,000 keys) is about 2 ms; one graph of 10,000 Subjects among 63 of Source A's size, about 5 ms (D-8, `TestCandidatesForgetFanOut`). A memo that loses a Subject is copied once (its Subject count). A commit that moves no key costs nothing.
- **Memory.** Each remembered Subject gains its kind and about 7 key strings: about 230 bytes, 23 KB for Source A, 2.3 MB for a 10,000-Subject graph. STEP-7 records `evidence-B` (which holds the candidates) before and after at both sizes.
- **After a commit.** The proposal of a Source none of whose Subjects' keys met the commit ranks nothing again; one whose Subjects met it ranks only those. Bar 10's ratio is gated at 4,000 and recorded at 16,000 (D-9).

### Edges

- **Assumptions.** A Subject's ranking reads only the handles `Lookup` returns for its kind and lookup keys and those handles' Subjects (`rankCandidates` → `seedHandles` → `graphalign.Candidates`), the Subject's own values (dropped with its graph on `Set.EvidenceGraphs`), and the config and metas (vocabulary, through `DerivedFrom`). The match index's Apply sees every handle whose Subject changed (S16-20's T-3 with Verify). A handle is filed under the same keys in only one kind (`filing.kind`). Held by T-1 with Verify on: `checkCandidates` ranks every held Subject again and compares.
- **Neighbors.** A handle whose values change but whose keys do not (a gender retyped, a lifespan from a re-dated event): its unchanged keys are moved, so the Subjects that shortlist it re-rank (AC-2.2 holds the lifespan case). A handle filed under no key before or after: nothing moves, and no Subject could have shortlisted it. A Subject of another kind with the same key text (an event's `t:…` keys and a person's `name|…` keys never share text; `movedKey` carries the kind): not forgotten. A commit that changes no profile: `moved` is empty, nothing is forgotten (T-1 "a commit that changes no profile").
- **Removed inputs.** A merged or deleted handle's keys are read before `Remove`, so the Subjects that ranked it re-rank without it (AC-2.4). A Subject forgotten from a memo keeps nothing: it is ranked from scratch on the next proposal, exactly as on a first proposal.
- **Real data.** None — no rule, seed or registry value changes; the keys are the index's own.
- **Untrusted size.** The scan is bounded by the entry's Cap (64 graphs) times the Subjects per graph times their lookup keys; keys per Subject are bounded by `Rules.MaxBuckets` (100) per date and the name's words. The moved set is bounded by the handles Apply re-profiles (S16-20 D-12 states and times those).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/shortlist` `index.go` | D-1 `Filed`, D-2 `LookupKeys`; `Lookup` reads its keys through `LookupKeys` |
| `core/database/graphcache` `cache.go` | D-3 `Keyed.Update` |
| `core/database/promotealign` `canon.go` | D-4 `rankedSubject`, `candidatesMemo`, `rankCandidates`, `loadCandidatesMemo`, `loadCanon` (`.ids`), `checkCandidates`; D-6 `forgetCandidates`; D-7 doc of `candidates` |
| `core/database/promotealign` `index.go` | D-5 `movedKey`, `applyMatchIndex`; D-7 doc in `init` |

```go
// shortlist
func (x *Index) Filed(id []byte) (kind vocabkey.SubjectType, keys []string, ok bool)
func (x *Index) LookupKeys(kind vocabkey.SubjectType, v subjectcompare.Values) []string

// graphcache
func (k Keyed[T]) Update(g *Graph, fn func(id []byte, v T) T)

// promotealign
type rankedSubject struct {
	kind vocabkey.SubjectType
	keys []string // the Subject's shortlist lookup keys (Index.LookupKeys)
	ids  [][]byte // ranked candidate handle ids
}
type candidatesMemo struct{ bySubject map[string]rankedSubject }
type movedKey struct {
	kind vocabkey.SubjectType
	key  string
}
func forgetCandidates(g *graphcache.Graph, moved map[movedKey]bool)
func rankCandidates(m *matchIndex, layer graphalign.Layer, subjects []graphalign.Subject, cfg graphalign.Config) map[string]rankedSubject
```

Errors: none new.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Measure the base the way the gate will**
Files: `core/database/promotealign/perf_test.go` (`proposalAfterCommit`), `core/database/promotealign/bench_test.go` (`BenchmarkLargeSourcePropose`); the ship log `S16-32-narrow-candidate-rerank.ship.md` (beside this plan; the orchestrator creates it).
Do: D-9's measurement half only: `proposalAfterCommit` GCs before the timed proposal and takes the best of three; the benchmark reports `after-commit-x`. No ratio gate yet. Run `BenchmarkLargeSourcePropose` and `BenchmarkCacheSizes` on the base and write `warm-ms`, `after-commit-ms`, `after-commit-x` and `evidence-B` at both sizes, with `hostname`, into the ship log.
Check: `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'BenchmarkLargeSourcePropose|BenchmarkCacheSizes' -benchtime 1x -timeout 60m ./core/database/promotealign` prints both sizes; `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets ./core/database/promotealign` passes.

**STEP-2 — The shortlist index says where a handle is filed and what a Subject looks up**
Files: `core/shortlist/index.go`, `core/shortlist/shortlist_test.go`. Skill: `add-match-rule` (the package stays pure).
Do: D-1, D-2. Add T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/shortlist` passes.

**STEP-3 — A keyed entry can rewrite its held values without counting a use**
Files: `core/database/graphcache/cache.go`, `core/database/graphcache/cache_test.go`, `.cursor/rules/catalog-queries.mdc`. Skill: `use-catalog-session`.
Do: D-3. Add T-5. Add the rule's clause.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache` passes.

**STEP-4 — The memo keeps each Subject's kind and lookup keys**
Files: `core/database/promotealign/canon.go`, `core/database/promotealign/export_test.go`.
Do: D-4; the `SubjectRankedForTest` and `HeldSubjectsForTest` hooks (D-11). No behaviour change: Apply still `Set`s.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign` passes, and again with `PROVENENCIA_GRAPH_VERIFY=1`.

**STEP-5 — A commit forgets only the Subjects whose keys it moved**
Files: `core/database/promotealign/index.go`, `canon.go`, `export_test.go`, `memo_test.go`.
Do: D-5, D-6, D-7; the `ForgetForTest` hook; D-10's change to "a commit that changes no profile"; T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign ./core/catalogsession` passes, and again with `PROVENENCIA_GRAPH_VERIFY=1`; `rg -n "w\.Set\(g, m\)" core/database/promotealign` finds nothing.

**STEP-6 — The forget's fan-out is timed**
Files: `core/database/promotealign/perf_test.go`, `export_test.go`.
Do: D-8 with the `HoldRankedForTest` hook: T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestCandidatesForgetFanOut -v ./core/database/promotealign` passes and logs the time.

**STEP-7 — Gate bar 10 and measure after**
Files: `core/database/promotealign/perf_test.go`.
Do: D-9's gate in `TestLargeSourceBudgets` (T-7). Run the STEP-1 commands on the same machine; write the after figures beside the base in the ship log. Every figure comes from a run with `PROVENENCIA_GRAPH_VERIFY` unset.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets -v ./core/database/promotealign` passes and logs the ratio; the benchmark's `after-commit-x` is at most 1.5 at both sizes.

**STEP-8 — Docs**
Files: `docs/catalog-caches.md`, `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-16/cache-architecture.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row status only: planning already rewrote the row's Scope, the scope entry and `cache-architecture.md` line 67, the "Candidates re-rank only what moved" bullet), `docs/deployment-plan/spike-16/completed.md`, this plan.
Do:
- `catalog-caches.md`, row `promotealign.candidates`: Fill "`loadCandidatesMemo`: `rankCandidates` for the Subjects not held, each with its kind and lookup keys"; Updated by "`DerivedFrom` `promotealign.matchindex` (emptied when it is dropped); the match index's Apply forgets the Subjects whose lookup keys meet a key a changed handle was filed under, before or after (`forgetCandidates`, through `Keyed.Update`); Apply drops the graphs in `effects.Set.EvidenceGraphs`". Row `promotealign.matchindex`, Updated by: "refiles the handles whose profile changed, then `Set`" → "refiles the handles whose profile changed and hands their old and new keys to `forgetCandidates`". The size column of `evidencelayer.layers` (which holds the candidates' figure): keep its format and replace the leading pair with STEP-7's `evidence-B` at 4,000 / 16,000, and its parenthesis with "(one proposal; <STEP-1's pair> MB before S16-32)".
- `promote-graph-alignment.md` §8, "The candidates" bullet: "They are dropped on `Set.EvidenceGraphs` for that graph and whenever the match index changes (they are derived from it)" → "Each Subject keeps its lookup keys beside its ranking. A commit that changes a handle forgets only the Subjects whose lookup keys meet a key that handle was filed under, before or after, and the next proposal ranks just those again; `Set.EvidenceGraphs` drops that graph's, and a rebuilt match index drops all". The match index bullet's "and refiles only the handles whose profile changed" gains ", handing their old and new keys to the candidates".
- `cache-architecture.md` line 66: "and the candidates, derived from it, are ranked again (S16-32 narrows that)" → "and the candidates forget only the Subjects whose lookup keys meet a key it moved (S16-32)".
- `deployment-plan.md`: the S16-32 row's Status → `Landed #N`.
- `completed.md`: one paragraph.
- This plan: Status, Landed note.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-32-narrow-candidate-rerank.md` reports no error; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/promotealign/memo_test.go` `TestCandidatesFollowMovedKeys` *(new)* | Table, each case on a fresh `newWarmFixture` plus graph C (a Source "Register" with `f.Person(…, "Mary Jones")`), with `PROVENENCIA_GRAPH_VERIFY=1`; B and C proposed before the write. Columns: write, and the expected held state of B's person (`w.bPerson`), B's birth (`w.bBirth`), C's person. "an unrelated person filed" (`filedBirth` "Zebedee Quarrington" 1790 on a new Source) → held, held, held. "the member renamed" (`renameMember` "Jon Smyth") → forgotten, held, held. "the member's birth re-dated" (`Cite` date 1818 on `memberBirth`) → forgotten, forgotten, held. "a John Smith filed" (`filedBirth` "John Smith" 1817 on a new Source) → forgotten, —, held, and the next proposal of B names the new handle as target or alternative. "the member merged" (`mergePerson`) → forgotten, —, held. "the member unpromoted" (`unpromote` `w.member`) → forgotten, —, held. After each: the next proposal of B equals one after `g.Drop()` ("—" is not asserted) | AC-1.1, AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-5, D-6 |
| T-2 | Go | `memo_test.go` `TestMatchIndexFollowsCommits` "a commit that changes no profile" *(changed)* | As today, plus `SubjectRankedForTest(g, graphB, w.bPerson.ID)` is true after the commit | AC-1.1, D-5 |
| T-3 | Go | `core/shortlist/shortlist_test.go` `TestIndexFiled` *(new)* | Table: a handle `Put` with "John Smith" → `Filed` returns kind person and the index keys `Keys` gives, sorted; after `Put` with no name → `ok` false; after `Remove` → `ok` false; a nil index → `ok` false; a `Put` under another kind → the new kind | D-1 |
| T-4 | Go | `shortlist_test.go` `TestLookupKeys` *(new)* | For a person probe and a dated event probe: `LookupKeys` equals `Keys`'s lookup list under the index's rules; a handle is in `Lookup`'s result exactly when one of the probe's `LookupKeys` is among its `Filed` keys (over `TestLookup`'s entries); a nil index → nil | D-2 |
| T-5 | Go | `core/database/graphcache/cache_test.go` `TestKeyedUpdate` *(new)* | A keyed int entry with Cap 3; `Set` a=1, b=2, c=3 (a least recently used); `Update` with `v*2`; then `Set` d=4 → `Peek` a is false (a left, not the key `Update` visited last), and `Peek` b, c read 4, 6. `Update` on a nil graph and on an entry that holds nothing does nothing and does not panic | AC-3.2, D-3 |
| T-6 | Go | `core/database/promotealign/perf_test.go` `TestCandidatesForgetFanOut` *(new)* | `graphcachetest.OpenCatalog`; graph id n (n from 0 to 63) is 16 bytes with `n` big-endian in its first 4 bytes and zeros after; `HoldRankedForTest` for ids 1 to 63 with 100 Subjects and id 0 with 10,000, 7 keys each (`SetCandidatesCapForTest` 64); `ForgetForTest(g, "person", keys)` with 20 keys: the key (i, j=0) of graph 0 for i from 0 to 9, and 10 keys `name|w:none-k` (k from 0 to 9) that meet nothing; `databasetest.Fastest(3, …)` (re-holding the memos before each run) `WithinBudget` `ceilingForgetCandidates`; after it `HeldSubjectsForTest` of the large memo is 9,990 and of each small one 100 | AC-3.1, D-8 |
| T-7 | Go | `perf_test.go` `TestLargeSourceBudgets` "after a handle commit" *(changed)* | `after` ≤ 1.5 × `warm`, both measured in the subtest (D-9); `budgetAfterCommit` still holds | AC-1.2, D-9 |
| T-8 | Go | `core/database/promotealign/bench_test.go` `BenchmarkLargeSourcePropose` *(changed)* | Reports `after-commit-x` at 4,000 and 16,000 | AC-1.3 |
| T-9 | Go | `memo_test.go` `TestEntriesFollowSignals` "a vocabulary commit", "a name format change" *(unchanged)* | No candidates held after either | AC-2.5, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestCandidatesMemoFollowsCommits`, `TestEntriesFollowSignals`, `TestVerifyLeavesEntriesAlone`, `TestMatchIndexFollowsCommits`, the `core/graphalign` goldens, `TestAlignInvariants`, and every `promotealign` test under `PROVENENCIA_GRAPH_VERIFY=1`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] CGO_ENABLED=1 PROVENENCIA_GRAPH_VERIFY=1 go test -tags fts5 ./core/database/promotealign ./core/catalogsession passes
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `rg -n "whenever the match index changes|changed handle is in (its|their) list|S16-32 narrows that" docs/*.md docs/deployment-plan/spike-16/cache-architecture.md .agents .cursor` finds nothing; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says (none)
- [ ] The ship log holds warm-ms, after-commit-ms, after-commit-x and evidence-B at 4,000 and 16,000, before and after, with the hostname
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated: none)
- [ ] Domain doc updated: docs/catalog-caches.md and docs/promote-graph-alignment.md §8 per STEP-8
- [ ] PR description: what shipped, AC evidence, the figures before and after, deviations, manual check
```

## 10. Manual check

1. Open a project with at least two Sources that have unpromoted people, and a filed person who appears in neither (any dogfood project). Open Promote on the first Source, then close it.
2. Open the filed person's Vitals and edit their name Observation; save.
3. Open Promote on the first Source again: the suggestions read as before the edit, and the page draws as quickly as the second opening of an unchanged Source.
4. Edit the name of a person who is one of that Source's suggested candidates; reopen Promote on it: the row's suggestion reflects the new name as a fresh open of the project would.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A ranking input the forget does not see leaves a stale Subject | Verify ranks every held Subject again after every commit in T-1 and T-2 (`checkCandidates`); a miss fails there |
| The ratio gate is noisy on the CI runner | Both sides are best-of runs with a GC first (D-9); the after-commit proposal ranks nothing, so it is a warm proposal plus a commit's worth of other cache work |
| Memo memory grows | About 230 bytes per Subject (§6 Scale); STEP-7 records `evidence-B` |

**Stop and report if:**

- `applyMatchIndex` no longer ends in `w.Set(g, m)` when something changed, or `candidates` is no longer registered with `DerivedFrom: MatchIndex` (§4 is stale).
- With `PROVENENCIA_GRAPH_VERIFY=1`, `checkCandidates` fails after any T-1 write: a ranking reads something D-5's moved keys do not cover. Report the write and the Subject; do not widen the forget to the whole graph.
- `after-commit-x` is above 1.5 at either size after STEP-7, or T-7 fails on the local run: report the figures and what the after-commit proposal still does (a CPU profile of `BenchmarkLargeSourcePropose`).
- T-1's expected held states do not hold because a fixture's name or date shares a lookup key with a moved one (`LookupKeys` of "Mary Jones" meets `Filed` of the member, or B's 1817 birth meets a key of the 1790 birth): report the shared key; do not change the names or years without saying so.
- `evidence-B` after STEP-7 exceeds STEP-1's base figure on the same machine by more than 1 MB at either size.
- `TestCandidatesForgetFanOut` takes over its 100 ms ceiling.

## 12. Open questions

None.

## 13. PR

**Title:** S16-32 — Promote re-ranks only the Subjects a write moved
**Body:** objective in one sentence; what shipped against each US; AC evidence; the figures before and after; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-32-narrow-candidate-rerank.md`.
