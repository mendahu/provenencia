# S16-20 — One match index that Apply keeps current

**Status:** Landed in [#827](https://github.com/mendahu/provenencia/pull/827) (2026-11-13)
**Spike / slice:** Spike 16 · Slice 2d — Promote
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine refactor; Promote's suggestions read the same as today)
**Board:** —
**Depends on:** S16-10 (landed [#789](https://github.com/mendahu/provenencia/pull/789))
**Blocks:** S16-32 (narrows the candidates' re-rank with the keys `Put` and `Remove` change), S16-22 (serves each handle's profile from this entry)
**Branch:** `s16-20-match-index-apply`
**Version bump:** none — no change a researcher can see, and no wire or catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) (the graph cache, its registry and Verify), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (the layer packages `core/shortlist` and `core/graphalign`)
**Rules that bind:** `.cursor/rules/go-tests.mdc` (table-driven tests), `.cursor/rules/catalog-queries.mdc` (a derived cache registers a typed handle from `init()`, reads only through its handle, keeps its row in `docs/catalog-caches.md` and its fill in `BenchmarkCacheSizes`), `.cursor/rules/match-rules.mdc` (`core/shortlist` and `core/graphalign` stay pure)
**Context:** [design § Promote](../cache-architecture.md#promote), [design § Warming](../cache-architecture.md#warming), [`docs/catalog-caches.md`](../../../catalog-caches.md), [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §7 and §8, working rules: "The point of this batch is simpler code", "Cache entities and results per key; compose answers", "One cache registry, two ways to fill", "Every cache names the signal that updates it, and updates exactly", "Apply states its fan-out", "Warm only what every proposal needs; no background re-warm", "Orders are defined and results do not move", "Memory is budgeted", "No new persisted projections, and no stored answer without a measurement", "Measure growth"
**Planned:** 2026-11-13 against `main` @ `64a91ae`
**Review:** [`S16-20-match-index-apply.review.md`](S16-20-match-index-apply.review.md) — Fits with changes / Ready with changes after two rounds; no open blocker

## Landed

- The ship log was not edited by the implementer (the orchestrator owns it), so the STEP-1 and STEP-6 figures are in the PR description: `after-commit-ms` 557.9 / 919.9 before and 390.8 / 666.9 after at 4,000 / 16,000 persons; `warm-ms` 112.5 / 124.5 before and 114.1 / 124.2 after; `matchIndex-B` 6.2 / 9.7 MB before and 34.0 / 55.1 MB after fix round 1 (33,979,848 and 55,095,528; 35.5 / 57.9 MB before it; under the 60 MB stop at 16,000); `WarmOnOpen` at 16,000, best of three, 791 ms before and 835 ms after; the Go caches' total at 16,000, 162 MB before and 176 MB after fix round 1 (179 MB before it). They ran on an Apple M1 (`Jakes-M1-Macbook-Pro.local`), not the planning base's Xeon, with the base measured from `c09b83b9e`.
- `main` had moved from the Planned commit to `c09b83b9e`; no **Current state** row changed.
- `profiler.readers` reads a neighbor's kind from `Link.NeighborTypeKey`, not from a loaded node.
- No test hook returns the held index as a pointer: `TestMatchIndexHoldsEveryHandle` and `TestWarmBuildsTheIndex` compare two `MatchIndex.Peek` results.
- `TestMatchIndexFollowsCommits` and `TestMatchIndexHoldsEveryHandle` set `PROVENENCIA_GRAPH_VERIFY=1` for the test, so each commit in them is verified as well as the explicit `g.Verify()`.
- `Index.Put` diffs the new keys against the filed ones and touches only the postings that moved; it returns early when the keys are unchanged.
- `inputs.namePattern` was deleted in fix round 1 (`Propose` was its last reader); `profiler.namePattern` stays (D-3).
- D-4's "`filed`'s key strings share their bytes with `byKey`'s" holds from fix round 1: `Index` keeps each key once, on its `posting{key, ids}`, and `filing.keys` names those strings. The saving is 2.8 MB at 16,000 persons, not the whole copy: `filed` still holds a slice header per handle and a string header per key.
- Fix round 1 also added `profiler.holds` (the one rule for which handles the index holds) and `TestIndexPatchesEqualBuild`.
- `promote-graph-alignment.md` §8's open-time warms bullet also says "the match index is built at open" in the sentence between the two the plan named.
- The §10 manual check was not run by the implementer.

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

Promote's match index (the shortlist postings that find each Source Subject's candidate handles) is thrown away by every commit that names a handle, and the next proposal rebuilds it from every person, event and place in the catalog. Handle profiles (the values a handle is compared by) are built lazily by a profiler that shares a map with the index and dies with it, and a generation number (`indexGen`) tells the candidates memo that the index it was ranked against is gone. This PR makes the match index one whole registry entry with a **Build**: the profile of every unmerged person, event and place, and the shortlist postings filed from those profiles. `promotealign.Warm` fills it at open. A commit no longer drops it: its **Apply** re-profiles the commit's roots and the people at the root events, and refiles only the handles whose profile changed. Proposals read profiles from the entry; the lazy profiler, its shared map, `keep`, `indexGen` and the generation checks go, and the candidates entry is declared `DerivedFrom` the match index instead. Verify compares the patched entry with a fresh Build. `graphalign.Handle` becomes an alias of `subjectcompare.Subject`, the profile the entry stores.

**Why now.** Bar 9 ("the next Promote proposal … reuse[s] … the match index instead of rebuilding" it) and the map's last drop-on-any-handle Apply (rule "Every cache names the signal that updates it, and updates exactly"; `docs/catalog-caches.md` § Known exceptions). S16-32 narrows the candidates' re-rank on top of this entry, and S16-22 serves each walked handle's profile from it.

**What stays.** Every proposal, candidate list, row, alternative and connection reads as today (the goldens, `TestAlignInvariants`, the warm-equals-fresh tests). The candidates are still ranked again after any commit that changes a profile or a key: narrowing that is S16-32's, so bar 10's "no slower than 1.5 times a warm one" moves to S16-32's Check (user, 2026-11-13: measured on the planning base, the re-rank alone costs 3.2× a warm proposal at 4,000 persons and 5.3× at 16,000). `TestLargeSourceBudgets`' budgets stay. No SQL, wire or Swift change.

## 2. User stories

**US-1.** As a researcher, when I save a change to a person and then open Promote, the engine does not rebuild the catalog's match index, so the proposal waits only for its own ranking.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Warm fixture (`newWarmFixture`), `Warm`; a member's name Observation edited from "John Smith" to "Jon Smyth" (`observations.Update`) | The match index is still held (no Build ran: `MatchIndex` was not emptied); `PROVENENCIA_GRAPH_VERIFY=1` Verify passes; the handle's held profile changed, and a `Lookup` with the new profile's values returns it |
| AC-1.2 | Warm fixture, `Warm`; each write in the T-3 table (a new filed person, a name edit, a birth date edit on an event with participants, a participation added, a handle merged, a Subject unpromoted, a place renamed) | After each: the index is held, `g.Verify()` returns nil, and the next proposal equals one on a dropped graph |
| AC-1.3 | Warm fixture, `Warm`; a commit whose roots change no profile (`effects.Set{Handles: [the person handle]}` with nothing written) | The index is held and unchanged, and the candidates entry still holds the graph it held before |
| AC-1.4 | Large fixture at 4,000 and 16,000 persons | `after-commit-ms` (`BenchmarkLargeSourcePropose`) is recorded before and after in the ship log; `TestLargeSourceBudgets` passes |

**US-2.** As the engine, a proposal reads every handle's profile from one entry built once, so no proposal builds a profile and no reader shares a mutable profile map.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A dropped graph; `Warm` | The entry holds a profile for every unmerged person, event and place handle (count equals the three kinds' unmerged nodes), and `AssertKindsLoaded(person, event, place)` |
| AC-2.2 | After `Warm`, a proposal on another Source | The entry is the same value (pointer equal) before and after; the proposal issues no graph query beyond its own Source's loads (`TestProposeWarmIssuesNoQuery`, `TestLargeSourceBudgets` "warm queries" unchanged) |
| AC-2.3 | A vocabulary commit, a new alias index, a name format change, `Drop` | The entry is emptied, and so is the candidates entry; the next proposal builds both and equals a fresh one (`TestMatchIndexDrops`, `TestCandidatesMemoFollowsCommits`) |
| AC-2.4 | The stored index loses a handle's postings (`EmptyShortlistForTest`), then a Source-only commit with Verify on | `OnCommit` returns an error naming `promotealign.matchindex` |

**US-3.** As a contributor, `graphalign.Handle` is the type every layer already calls a Subject, so a profile moves from the entry into `Align` with no copy.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `core/graphalign/types.go` | `type Handle = subjectcompare.Subject`; `go build ./...` passes with no other change to `graphalign` or its tests |

## 3. Scope

| In | Covered by |
| --- | --- |
| `matchIndex` holds every unmerged person, event and place profile (`subjectcompare.Subject` by handle id) and the shortlist index filed from them; a Build | STEP-3, AC-2.1, D-1, D-2, D-3 |
| `shortlist.Index` keeps each handle's filed keys and sorted postings; `Put` and `Remove` replace `Add` | STEP-2, D-4 |
| Apply re-profiles the roots and the people at root events, refiles changed handles, and `Set`s only when something changed | STEP-4, AC-1.1 – AC-1.3, D-5, D-6 |
| Candidates `DerivedFrom` the match index; `indexGen`, `candidatesMemo.gen` and the generation checks go | STEP-3, AC-2.3, D-7 |
| The lazy profiler goes: `done`, `profile`, `keep`, `keyValues`, `values`, `newProfilerWith`, `shortlistInputs`, `loadShortlist`, `buildShortlist`; keys come from the full profile | STEP-3, D-3, D-8 |
| Proposals read profiles from the entry (`seedHandles`, `loadCanon`); `propose` folds into `Propose` | STEP-3, AC-2.2, D-9 |
| The shortlist-lead rule keeps its real reason in `add-match-rule` and `TestShortlistSection` | STEP-7, D-8 |
| Check compares the entry with a fresh Build; `checkCandidates` drops its generation guard | STEP-3, AC-2.4, D-10 |
| `graphalign.Handle` is an alias of `subjectcompare.Subject` | STEP-1, AC-3.1, D-11 |
| Apply's fan-out stated and timed at bar 2's sizes | STEP-5, D-12 |
| Figures before and after (open, after-commit, warm, the entry's heap) | STEP-1, STEP-6, AC-1.4 |
| Docs this makes false; the map's S16-20 and S16-32 Checks per the user's answer | STEP-7 |

| Out | Home |
| --- | --- |
| Re-ranking only the Subjects a write moved; recording the changed keys; bar 10's 1.5× gate | S16-32 |
| The canon walk, `loadCanon`'s link merge, a per-handle canon view | S16-22 |
| `graphalign.CanonEdge` onto `kinship.Link`; `Connections` reading the place entries | S16-21 |
| Memoising a person's composed kinship | [`ideas/promote-ranking.md`](../../../ideas/promote-ranking.md) "Memoise a person's composed kinship" |
| Alike-pair memos (B-11) | [`ideas/promote-ranking.md`](../../../ideas/promote-ranking.md) |
| Replacing `subjectcompare.Subject` in `graphalign`'s own code and tests by name | not planned (the alias is the fold) |
| Changing any `TestLargeSourceBudgets` budget | not planned |
| `canonlinks` owning the event-type property lookup that `newDeriver` and `loadCanon` each make (review M-6) | S16-21, which reshapes `EdgeOf` |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| The entry | `core/database/promotealign/index.go` `matchIndex`, `MatchIndex`, `init` | `Whole[matchIndex]` (a value, not a pointer): `shortlist`, `gen`, `profiles map[string]profile`, `profilesPattern`. No Build. DerivedFrom the vocabulary, `namealiases.Index`, `nameformats.Default`. Apply: `w.Drop(g)` when `len(roots) > 0` — the known exception |
| Generation | `index.go` `indexGen` | `atomic.Uint64`; each `loadShortlist` build gets a new one |
| Fill | `index.go` `loadShortlist`, `buildShortlist`, `Warm`, `shortlistInputs`, `matchIndexOf` | `loadShortlist` builds the shortlist from `p.keyValues(n)` (a profile with no life) of every unmerged node of the rule kinds, stores it with `MatchIndex.Set`, then `p.keep()`. `Warm` calls it at open. `shortlistInputs(g, fresh)` makes config, pattern and a profiler (shared map or fresh) |
| Check | `index.go` `checkMatchIndex` | Rebuilds the shortlist and each stored profile with a fresh profiler; `reflect.DeepEqual` |
| Seeds | `index.go` `seedHandles` | `ix.Lookup` per Subject, `g.Nodes(union)`, skips nil and merged, `p.profile(present)`, makes `graphalign.Handle{ID, Ref, Kind, Values}`, sorts by Ref then ID |
| Profiler | `core/database/promotealign/profiles.go` `profiler`, `newProfiler`, `newProfilerWith`, `keep`, `kind`, `profile`, `entity`, `keyValues`, `build`, `values`, `profile` type, `hopPersonToEvent` (the same hop as `core/hops` `EventsOfParticipant`), `reconciled` | Builds each node's `subjectcompare.Values` once into `done`, the index's map, shared across proposals; a lifespan kind (person, `matchregistry` `Subject.Lifespans`) reads its life from the profiles of the events it joins (`hopPersonToEvent`), loading and building those first |
| Candidates | `core/database/promotealign/canon.go` `candidatesMemo{gen, bySubject}`, `candidates`, its `init`, `loadCandidatesMemo`, `rankCandidates`, `checkCandidates` | Registered in `canon.go`'s `init` (which runs before `index.go`'s: files init in name order). DerivedFrom the vocabulary and `nameformats.Default`; Apply `DropNamedGraphs`. A memo is used only while `gen` equals the stored index's; `checkCandidates` skips a stale one |
| Canon | `canon.go` `loadCanon` | Takes the profiler; `prof.kind(n)` decides primary kind; at the end `prof.profile(nodes)` and builds `graphalign.Handle` from `prof.values` |
| Propose | `core/database/promotealign/promotealign.go` `Propose`, `propose` | `propose` makes a profiler with `newProfiler(c.Graph(), layer.Metas, cfg.Subject, namePattern)` and returns it for `ProfiledRefsForTest`; `Propose` forwards to it |
| Shortlist | `core/shortlist/index.go` `Index`, `NewIndex`, `Add`, `Lookup` | `byKey` per kind of `[][]byte` postings in insertion order; "read-only once built"; `Lookup` dedups and sorts |
| Keys | `core/shortlist/shortlist.go` `Keys` | Index and lookup keys, each sorted and unique (`sortedUnique`) |
| Handle | `core/graphalign/types.go` `Handle` | `struct{ID []byte; Ref string; Kind string; Values subjectcompare.Values}` — the fields of `subjectcompare.Subject` (`core/subjectcompare/subjectcompare.go:75`); `Subject` is already an alias on line 13. The map's Batch 2 sources table names the fold; the `structural-debt.md` row moved to this PR and is no longer in that file |
| Open | `core/catalogsession/session.go` `WarmOnOpen` | Calls `promotealign.Warm` and ignores its error |
| Registry | `core/database/graphcache/cache.go` `Whole.Get`, `Whole.held` (unexported), `Keyed.Peek`, `Whole.Set`, `dropReaders`, `register` | `Whole.Get` runs a Build when nothing is held, so once an entry has a Build, `Get`'s `ok` no longer says whether it was held; `Whole` has no non-building probe, `Keyed` has `Peek`.  `Set` stores and empties every entry `DerivedFrom` it; a Build's first fill empties none. `register` panics when a source is not registered earlier |
| Commit | `core/database/graphcache/commit.go` `OnCommit` | Roots: the written ids; an association's old holders and new ends; a removed node's cached neighbours; a merged handle's neighbours. Reloads loaded ones, then `applyEntries` |
| Errors | `core/writes/run.go` (after `Notify`) | An `OnCommit` error is logged, every listener drops, and the write still returns success |
| Tests | `core/database/promotealign/memo_test.go` `TestProfilesLiveWithTheMatchIndex`, `TestCandidatesMemoFollowsCommits`, `TestWarmBuildsTheIndex`, `TestMatchIndexDrops`, `TestVerifyChecksMatchIndex`, `TestMemoFollowsCommit` | Profiles kept and dropped by a handle commit; warm equals fresh after each commit kind; Warm's generation; the declared drops (the handle commit drops); Verify fails on an emptied shortlist |
| Tests | `core/database/promotealign/promotealign_test.go` `TestProposeColdIndexBuildsNoUnshortlistedLife`, `TestProposeWithOnlyFixedSubjectsBuildsNoIndex` | A cold proposal profiles only the shortlisted handles and what the walk reaches (`ProfiledRefsForTest`); a proposal whose Subjects are all fixed leaves the index unbuilt |
| Tests | `core/database/promotealign/export_test.go` | `ProfiledRefsForTest`, `ProfilesKeptForTest`, `MatchIndexGenForTest`, `EmptyShortlistForTest`, `HoldCandidatesForTest` read `done`, `profiles`, `gen`; `CandidatesHeldForTest` peeks the candidates entry. `memo_test.go` `stored(g)` probes the index with `MatchIndex.Get`'s `ok` |
| Tests | `core/shortlist/shortlist_test.go` `TestLookup` (`build` helper calls `Add`), `TestLookupCopiesAndSorts`, `TestKeys` | Index behaviour |
| Tests | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`, `bench_test.go` `BenchmarkCacheSizes` (`matchIndex` fill), `BenchmarkLargeSourcePropose` | Budgets at 4,000; sizes and timings at both |
| Tests | `core/catalogsession/warm_test.go` `TestDoWarmsMatching` (line 60), `TestNameFormatDropsItsReaders` (lines 145, 153), `TestDoKeepsWarmEntries` | Each probes the index with `MatchIndex.Get`'s `ok`. In `TestDoKeepsWarmEntries`, case "handle commit" (a nonexistent handle id) expects `wantIndex` false; the doc comment above the cases says a handle commit empties the index |
| Skill and registry test | `.agents/skills/add-match-rule/SKILL.md` "Add a shortlist key" step 2; `core/matchregistry/registry_test.go` `TestShortlistSection` (comment line 46, message line 62) | "The lead is a catalog property, never a derived one (`Lifespans`): `keyValues` builds no life"; "the index builds no life" |
| Siblings | `core/database/sources/records.go` `RecordsOf`, `core/database/evidencegraphs/index.go` `IndexOf` | Each graph-lifetime whole entry has one named read over `database.ReadWhole` |

**Measured on the planning base** (scratch benchmarks, removed; Xeon 2.1 GHz, 4 cores, 2026-11-13): full profiles of every unmerged person, event and place take 121 ms and 20.7 MB at 4,000 persons (18,233 handles) and 202 ms and about 34 MB at 16,000 (30,233 handles); today's `Warm` keying takes 169 / 290 ms. The index keys of the full profile equal those of `keyValues` for all 18,233 and 30,233 handles. A proposal of Source A after only the candidates are cleared: 723 ms against a 226 ms warm one at 4,000, 1,187 ms against 226 ms at 16,000.

**Precedent.** S16-12's Source records ([#800](https://github.com/mendahu/provenencia/pull/800), `core/database/sources/records.go` `Records`): a whole `LifeGraph` entry over a pointer with a Build, an Apply that patches the held value in place and calls `w.Set(g, idx)`, and a Check that compares with a fresh build by `reflect.DeepEqual`. Copy that shape. Do not copy its `w.Get` inside Apply: Apply runs only when the entry is held (`applyEntries`), so read the held value with `MatchIndex.Get` once and use it. S16-16 ([#814](https://github.com/mendahu/provenencia/pull/814)) is the precedent for the before / after figures in the ship log (`BenchmarkLargeSourcePropose`).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `type matchIndex struct { handles map[string]subjectcompare.Subject; shortlist *shortlist.Index }`; `var MatchIndex graphcache.Whole[*matchIndex]`. `handles` holds every unmerged person, event and place handle (`vocabkey.PrimaryKinds()`), by `string(id)`; each Subject carries the node's `ID` (a copy), `Ref`, `Kind` (the subject type key) and full profile `Values`. `gen`, `profiles`, `profilesPattern` go | One map of the type every layer reads (US-3); a pointer as S16-12's records, so Apply patches in place. `profilesPattern` is redundant: `DerivedFrom nameformats.Default` already empties the entry on a pattern change | Keeping only the rule kinds — the walk reaches every primary kind, so a place would need a lazy profile again; a separate keyed entry for profiles — two caches that must change together, and the map's design names one entry |
| D-2 | `MatchIndex` gets `Build: buildMatchIndex` *(new)* and `Lifetime: graphcache.LifeGraph`. `buildMatchIndex(g)`: `newProfiler(g)` (D-3); for each kind in `vocabkey.PrimaryKinds()` whose type id is 16 bytes, `g.Kind(typeID)`; collect every unmerged node of the three kinds into one slice and call `p.subjects(all, built)` (D-3) once into a new map — never once per kind, since `subjects` alone orders events before people and a per-kind call would build every person before any event (`PrimaryKinds()` lists person first), giving each an empty life (review M-10); then `ix := shortlist.NewIndex(cfg.Alignment.Shortlist, cfg.Subject, cfg.Value, metas)` and `ix.Put(vocabkey.SubjectType(s.Kind), s.ID, s.Values)` for every handle in sorted id order. `matchIndexOf(g) (*matchIndex, error)` stays as the one named read, now `database.ReadWhole(MatchIndex, g)`, as `sources.RecordsOf` and `evidencegraphs.IndexOf` are; `Warm(c)` becomes the `c.DB()` check plus `matchIndexOf(c.Graph())`; `WarmOnOpen` is unchanged | Map: "filled by `promotealign.Warm` at open"; rule "Warm only what every proposal needs" names this fill. `ReadWhole` is the one guarded read of a graph-lifetime whole entry (S16-12, S16-15) | Building lazily on the first proposal — the map puts it at open; building only keys at open and profiles later — the shared lazy map comes back |
| D-3 | `profiles.go` owns what a profile reads and in what order, in one builder: `profiler{g, props, terms, metas, types, typeKeys, rules, reconcile, namePattern, cfg graphalign.Config}` (`eventTypeProp` goes, D-9) and `newProfiler(g) (*profiler, error)`, which reads `loadConfig(g)` and `valuerules.Metas(g)` itself (`shortlistInputs` folds into it). Methods: `kind(n)`, `entity(n)`; `subject(n, built map[string]subjectcompare.Subject) (subjectcompare.Subject, error)` *(today's `build`)*: a kind with a `Lifespans` rule follows `hops.EventsOfParticipant` and reads each event's `Values` from `built` (an event absent from `built` is skipped, as today); `subjects(nodes []*graphcache.Node, built map[string]subjectcompare.Subject) error` *(today's `profile` without `done` and without loading neighbours)*: skips nil and merged nodes, builds the nodes whose kind has no `Lifespans` rule first, then the rest, each stored into `built`; `readers(ids [][]byte) ([][]byte, error)` *(new; the inverse of `subject`'s read)*: for each id whose node's kind has no `Lifespans` rule, the neighbours across `hops.ParticipantsOfEvent` whose kind has one, deduplicated by `string(id)` in first-seen order. `hopPersonToEvent`, `done`, `profile`, `keep`, `keyValues`, `values`, `newProfilerWith`, the `profile` type, `loadShortlist`, `buildShortlist`, `shortlistInputs` go | One file says what a profile reads, its build order, and who must be rebuilt when a read input changes (review M-1); one declared hop (M-2) | Writing the order in Build and again in Apply — two copies of one rule; keying the reverse read on `kind == Event` — the `Lifespans` rule drives the read; keeping `keyValues` for keys — measured equal on every handle of both fixture sizes (D-8) |
| D-4 | `core/shortlist/index.go`: `Index` gains `filed map[string]filing` *(new; `filing{kind vocabkey.SubjectType; keys []string}`)*. `Put(kind, id, v)` *(new; replaces `Add`)* computes `Keys`, removes `id` from each previously filed key no longer in the new list and inserts it into each new key's posting at its `bytes.Compare` position (no duplicate), and records `filing{kind, keys}`; a `Put` that yields no keys unfiles the handle and leaves it out of `filed`; a `Put` under another kind unfiles it from the old kind first. `Remove(id)` *(new)* takes it out of each filed posting, deletes an emptied posting and an emptied kind map, and forgets it. Postings stay sorted by `bytes.Compare`. A nil `*Index` ignores both. The type comment: "It is patched by Put and Remove; the caller serializes them with Lookup." `Lookup` is unchanged. Neither returns the changed keys (S16-32 adds that when it reads them) | The pure package owns its postings and their order (rule "Orders are defined"); sorted postings make a patched index `DeepEqual` a fresh build, so Verify can compare | Keeping `Add` beside `Put` — two ways to file; postings as sets — more memory for 196,526 postings; diffing keys in `promotealign` — the owner of the postings is `shortlist`; `Put(kind, id, old, new)` with the caller's old values (review M-8) — `filed`'s key strings share their bytes with `byKey`'s, so it costs slice headers (STEP-6 records `matchIndex-B`), and an index that trusts its caller's old values is corrupted by one wrong caller |
| D-5 | Apply `applyMatchIndex(g, w, roots, _ effects.Set) error` *(new)*: nothing when `len(roots) == 0`. Else, in this order: `m, err := matchIndexOf(g)`; `p, err := newProfiler(g)`; `g.Nodes(roots)`; `more, err := p.readers(roots)` (D-3; it reads the roots' nodes, now loaded); `g.Nodes(more)`; `reached` = roots then `more`, deduplicated by `string(id)`. For each reached id whose node (`g.Node`) is nil, merged, or of a kind that is not primary, and that `m.handles` holds: `delete(m.handles, id)`, `m.shortlist.Remove(id)`, `changed = true`. The other reached nodes are `live`: copy their held Subjects into `old` (by id), then `p.subjects(live, m.handles)`, which overwrites them in place, events before people, so a person reads the events just rebuilt. For each live id whose new Subject is not `reflect.DeepEqual` to `old[id]` (or `old` had none): `m.shortlist.Put(vocabkey.SubjectType(s.Kind), s.ID, s.Values)`, `changed = true`. At the end, `w.Set(g, m)` when `changed`. Errors are returned (D-13). A root the index never held and that changes nothing (a nonexistent id) is a no-op: no `Set` | Map: "re-profiles the commit's roots and the people at the events they touched, once"; keys are a function of the profile, so an unchanged profile needs no refiling. `Set` empties the candidates (D-7) only when something changed. Loading the roots before `readers` classifies them: a root the commit created may not be loaded, since `OnCommit` reloads only loaded nodes (review X-3) | Re-keying only the written handles from `set.Handles` — the roots are what `OnCommit` reloads and hands every Apply; dropping and rebuilding — the exception this PR removes |
| D-6 | `register` order: one `init()` in `index.go` registers `MatchIndex` and then `candidates`; `canon.go`'s `init` goes (the `candidates` var and its doc stay in `canon.go`) | `register` panics when a `DerivedFrom` source is registered later, and `canon.go`'s `init` runs before `index.go`'s | Renaming files to force init order — hidden coupling |
| D-7 | `candidates` declares `DerivedFrom: []graphcache.Source{MatchIndex}` (the match index is itself derived from the vocabulary, the alias index and the name format, so those reach it through `dropReaders`), keeps `Apply: graphcache.DropNamedGraphs[candidatesMemo]`, `Check: checkCandidates`, `Cap: graphfill.MaxEvidenceGraphs`. `candidatesMemo` keeps `bySubject` only; `loadCandidatesMemo` reads `candidates.Peek` with no generation test | The declaration replaces the hand-written generation (rule "One cache registry": DerivedFrom replaces hand-written drops). Correctness is today's: a ranking reads keys and profiles, and any change to either `Set`s the index | A generation kept beside DerivedFrom — two mechanisms; narrowing now — S16-32 |
| D-8 | A handle's shortlist keys are `shortlist.Keys` of its full profile (the entry's `Values`), at Build and at Apply. The rule "a shortlist lead is a catalog property, never a derived one" (`add-match-rule` "Add a shortlist key" step 2, `TestShortlistSection`) stays, with its real reason: a derived lead's keys would move with every edit to the events it is read from, so each event edit would refile all its participants, and the keys would stop equalling today's measured keys | One profile per handle; measured equal to today's `keyValues` keys on all handles at both sizes (§4). Goldens and `TestLargeSourceBudgets` "fixture" hold the results. The old reason ("`keyValues` builds no life") is false after this PR (review M-3) | Two builds per handle (full and key-only); lifting the rule — no lead needs it, and it would widen Apply's refiling |
| D-9 | Readers. `Propose(c, graphID, fixed, only)` keeps its signature and absorbs `propose` (its last other caller, `ProfiledRefsForTest`, goes). It reads `m, err := matchIndexOf(c.Graph())` once, always, and passes it down: `loadCanon(g, graphID, m *matchIndex, layer, fixed, cfg, terms)`, `loadCandidatesMemo(g, graphID, m, layer, todo, cfg)`, `rankCandidates(m, layer, subjects, cfg)`, `seedHandles(m, subjects) [][]graphalign.Handle`. `seedHandles` reads `m.shortlist.Lookup` and `m.handles[id]` (an id absent from `handles` is skipped), sorts by Ref then ID, and loads no node. `loadCanon` keeps its `handles map[string]*graphcache.Node` (the walk reads `n.Links`) and its `kinds` map; a node is primary when `m.handles` holds its id, and its kind is that Subject's `Kind` (no `prof.kind`); it reads `eventTypeProp` itself as `vocab.PropertyID(vocabkey.PropertyEventType, vocabkey.OriginProvenencia)` from `g.Vocabulary()` (as `canonlinks.newDeriver` does); `out.Handles` are `m.handles[id]` for each walked id, sorted as today, shared with the entry, not copied: `graphalign` reads a Handle's `ID` and `Values` and never writes them (T-2 holds it: Verify after two proposals) | Proposals compose from the entry and build nothing (rule "The point of this batch is simpler code"). Reading the entry even for a Source whose Subjects are all fixed: the app builds it at open, so the old saving (no index built for a fixed-only proposal) no longer exists (review X-1) | Reading `m` only when there are unfixed Subjects — a second way to decide primary kind in the walk; copying each Handle — a copy per walked handle per proposal for no reader |
| D-10 | `MatchIndex`'s `Check(g, stored *matchIndex)`: `fresh, err := buildMatchIndex(g)`; error when `!reflect.DeepEqual(stored.handles, fresh.handles)` ("promotealign: the stored handle profiles do not match a fresh build") or `!reflect.DeepEqual(stored.shortlist, fresh.shortlist)` ("promotealign: the stored shortlist index does not match a fresh build"). `checkCandidates` reads `m, err := matchIndexOf(g)` and ranks against `m`, with no generation guard | Map: "Verify compares the patched entry with a fresh build" | Comparing only counts |
| D-11 | `core/graphalign/types.go`: `type Handle = subjectcompare.Subject` with the comment "Handle is one canonical entity in the canon: a Subject whose ID is a handle's." | The structural-debt row's `Handle` half; an alias needs no change in `graphalign`'s code or its 77 test uses, as `Subject` on line 13 | Renaming every use to `Subject` — churn in a pure package for no reader gain |
| D-12 | Fan-out (rule "Apply states its fan-out"), stated in §6 Scale and held by `TestMatchIndexApplyFanOut` *(new)* with ceilings, `Fastest(3, …)` of the `OnCommit` alone: an event with 500 participants, its date edited → ≤ 250 ms; a person father of 2,000 births, a gender term retyped → ≤ 250 ms. Ceilings are package constants in `perf_test.go` with a `// decided:` comment naming this plan | Bar 2's crafted sizes; profile cost measured at about 7 µs a handle (202 ms / 30,233), so 500 people is a few milliseconds and the ceiling only catches a quadratic slip | No test — the rule asks for the case stated; a cap that falls back to a full drop — a second path for a cost that stays linear in what the commit reloaded |
| D-13 | Apply returns its error. `writes.Run` already logs an `OnCommit` error and every listener drops, so the next proposal builds a fresh entry | Today's behaviour for every entry (S16-12) | Swallowing it inside Apply |
| D-14 | Tests. **Go:** `ProfiledRefsForTest`, `ProfilesKeptForTest`, `MatchIndexGenForTest`, `TestProposeColdIndexBuildsNoUnshortlistedLife` and `TestProposeWithOnlyFixedSubjectsBuildsNoIndex` (both held the lazy fill the map removes: the entry is built at open). `TestProfilesLiveWithTheMatchIndex` is replaced by `TestMatchIndexHoldsEveryHandle` (T-2). **Change:** `TestWarmBuildsTheIndex` compares the entry's pointer before and after a proposal instead of a generation, and asserts kinds person, event and place. `TestMatchIndexDrops`: the handle commit case asserts the index is held (message "a handle commit dropped the match index"). `core/catalogsession/warm_test.go` `TestDoKeepsWarmEntries`: case "handle commit" → `wantIndex` true (a nonexistent id changes nothing, D-5), and the cases' doc comment drops "and a handle commit" from what empties the index (review X-2). `EmptyShortlistForTest` stores, through `MatchIndex.Set`, a new `*matchIndex` whose `handles` is the held map and whose `shortlist` is a new empty index. `HoldCandidatesForTest` stores `candidatesMemo{}`. **Add** to `export_test.go`: `IndexedForTest(g *graphcache.Graph, kind string, probe subjectcompare.Values) [][]byte` (the held shortlist's `Lookup`, nil when none is held), `HandlesHeldForTest(g) int`, and `HandleForTest(g, id []byte) (subjectcompare.Subject, bool)` (the held Subject of a handle). All three read through `MatchIndex.Peek`. **Fixture:** `warmFixture` gains `memberBirth subjects.Subject` (the second value `filedBirth` returns, discarded today) and a helper `renameMember(t, w, name string)` *(new)*, the body of `TestMemoFollowsCommit`'s "a member's Observation edited" closure with the name a parameter; that subtest calls it (review X-4) | Every test that read the removed fields either reads the entry or goes with the behaviour it held | Keeping the two lazy-fill tests — they hold what the map removes |
| D-15 | `graphcache` gains `func (w Whole[T]) Peek(g *Graph) (T, bool)` *(new)*: the held value, never building (the exported form of `held`, as `Keyed.Peek`). Every test that asks "is the index held" uses `MatchIndex.Peek`: `memo_test.go` `stored(g)` (so `TestMatchIndexDrops`, T-2, T-3, T-5, T-7), `warm_test.go` `TestDoWarmsMatching` (line 60), `TestDoKeepsWarmEntries` (line 109) and `TestNameFormatDropsItsReaders` (lines 145 and 153), and the `export_test.go` hooks. No `MatchIndex.Get` probe is left in a test | With a Build, `Get` rebuilds on a miss, so every "emptied" assertion would pass or fail for the wrong reason (review X-10); one probe for every whole entry with a Build | A test hook per package — `warm_test.go` is in `catalogsession`, outside `promotealign`'s `export_test.go` |

**Rules this plan bends.** None. The map's S16-20 Check names bar 10's second half; the user moved that gate to S16-32 (2026-11-13), and STEP-7 rewords both Checks.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `buildMatchIndex` *(new)* | func | `buildShortlist` (goes) | It is that function extended to profiles, renamed for what it builds |
| `applyMatchIndex` *(new)* | func | `sources.Records`' Apply | Same shape (patch, `Set`); different signal (roots, not `SourceScopes`) |
| `profiler.subject`, `profiler.subjects` *(new names)* | methods | `profiler.build`, `profiler.profile` | The same bodies with the built map passed in and no shared `done` |
| `profiler.readers` *(new)* | method | `hops.ParticipantsOfEvent` (used as is) | It is the inverse of `subject`'s read, kept beside it so the two change together |
| `shortlist.Index.Put`, `Remove`, `filing` *(new)* | methods / type | `Index.Add` (goes) | `Add` cannot unfile; `Put` is `Add` that replaces |
| `TestMatchIndexApplyFanOut`, `TestMatchIndexHoldsEveryHandle`, `TestMatchIndexFollowsCommits`, `TestIndexPut` *(new)* | tests | `TestMemoFollowsCommit`, `TestLookup` | New behaviour; they reuse `newWarmFixture`, `newFixture` and the `build` helper |
| `graphcache.Whole.Peek` *(new)* | method | `Keyed.Peek`, `Whole.held` | `held` is unexported; `Peek` is its exported name, matching `Keyed` |
| `IndexedForTest`, `HandlesHeldForTest`, `HandleForTest` *(new)* | test hooks | `ProfilesKeptForTest` (goes) | Read the new fields |

**Builds on.** The cache registry (S16-10): a whole entry with Build, Apply, DerivedFrom and Check; S16-12's patch-and-`Set` Apply and its named read (`RecordsOf`); `database.ReadWhole`; the declared hops `hops.EventsOfParticipant` and `hops.ParticipantsOfEvent`.
**Deletes.** In `promotealign`: `indexGen`, `matchIndex.gen`, `.profiles`, `.profilesPattern`, `loadShortlist`, `buildShortlist`, `shortlistInputs`, `checkMatchIndex`'s profile loop, the `profile` type, `hopPersonToEvent`, `profiler.done`, `.eventTypeProp`, `profile`, `keep`, `keyValues`, `values`, `build` (now `subject`), `newProfilerWith`, the drop-on-any-handle Apply, `candidatesMemo.gen`, `canon.go`'s `init`, `propose` (folded into `Propose`); `seedHandles`' node load. In `shortlist`: `Index.Add`. In `graphalign`: the `Handle` struct. Tests and hooks per D-14. Docs: `docs/catalog-caches.md` § Known exceptions' text (the section says none is left).
**Next adds.** S16-32: `Put` and `Remove` return the keys that gained or lost a handle, Apply hands them to the candidates' Apply, and `candidates` drops its `DerivedFrom MatchIndex` for a narrow re-rank. S16-22: a canon view reads `m.handles[id]` per walked handle.

### Reach

None — no new path into existing code. Apply is a new caller of `p.subject` and `shortlist.Index.Put`/`Remove`, only on the session graph (a write's graph holds no graph-lifetime entry).

### Scale

At the map's Scale (large fixture at 4,000 and 16,000 persons; 18,233 and 30,233 person, event and place handles) and bar 2's crafted sizes:

- **Open.** Build profiles every handle (121 / 202 ms measured) and files their keys (today's `Warm`: 169 / 290 ms, including a key-only profile per handle that goes). Expected `matchIndex` fill ≈ the two profile and keying passes less the key-only profiles; STEP-1 and STEP-6 record `WarmOnOpen` and the `matchIndex` fill at both sizes. Open is 4.16 s at 16,000 today.
- **Memory.** Profiles add about 21 MB at 4,000 and 34 MB at 16,000 to today's 6.2 / 9.8 MB; the Go caches stay under 250 MB (catalog-caches.md totals about 118 MB at 16,000 before this PR).
- **Apply per commit.** One `newProfiler` (the vocabulary maps, `loadConfig`, `Metas`: held entries, no query), then one profile per root and per person at a root event; `Put` only for changed profiles, each O(keys × posting length) for the memmove. Real data (10 participants per event, a handful of roots): under a millisecond of profiles. Crafted: an edited event with 500 participants → 501 profiles; a 20,000-child hub's own edit → one profile reading 20,000 life events; deleting or merging that hub → its 20,000 events are roots, each re-profiled with its own participants (about 60,000 profiles, ≈ 0.4 s at 7 µs), once, for a write that already reloads those nodes. D-12 times the first two.
- **After a commit.** The next proposal no longer rebuilds keys or profiles; it still re-ranks every Subject of the proposed Source when the commit changed a profile (723 / 1,187 ms measured at the base on the scratch run against a 226 ms warm proposal). S16-32 removes that.

### Edges

- **Assumptions.** A profile reads only its node's values and, for a lifespan kind, the profiles of the events it joins and the roles of those links (`profiler.build`). Any commit that changes them names the node or the association as a root (`OnCommit`: written ids, association ends, a removed node's neighbours, a merged handle's neighbours). An event's own profile reads no other node, so re-profiling the people at a root event is exactly what an event change reaches. Held by T-3 with Verify, over every write kind in the table.
- **Neighbors.** A person two hops from a write (the parent of a re-dated event's subject): its profile reads no link of its child's, so it is not re-profiled and does not change (T-3 "an event date edited": Verify passes). A commit naming a Source only: no roots, nothing done (AC-2.3's `TestMatchIndexDrops` "a Source-only commit"). A root whose profile is unchanged: no `Put`, no `Set`, the candidates are kept (AC-1.3).
- **Removed inputs.** A merged or deleted handle leaves `handles` and every posting (T-3 "a handle merged", "a Subject unpromoted"); a deleted event leaves its participants re-profiled with no life from it, since their participations are committed with it and their ends are roots (T-3 "an event deleted"); a probe that matched only it no longer returns it. A posting left empty is deleted, so a fresh build and a patched one are equal (T-4).
- **Real data.** None — no rule, seed or registry value changes; the keys are the same keys (§4 measurement).
- **Untrusted size.** A crafted event's participants and a hub's events are bounded by the commit's roots, stated above and timed (D-12). Keys per handle are bounded by `shortlist.Rules.MaxBuckets` (100) per date and by the name's words.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/shortlist` `index.go` | D-4: `filed`, `Put`, `Remove`; `Add` goes; postings sorted |
| `core/graphalign` `types.go` | D-11: `type Handle = subjectcompare.Subject` |
| `core/database/promotealign` `index.go` | D-1, D-2, D-5, D-6, D-10: the entry, `buildMatchIndex`, `applyMatchIndex`, one `init`, `Warm`, the Check, `seedHandles` (D-9) |
| `core/database/promotealign` `profiles.go` | D-3: one builder, `newProfiler(g)`, `subject`, `subjects`, `readers` |
| `core/database/promotealign` `canon.go` | D-7, D-9: `candidatesMemo{bySubject}`, `loadCanon`, `loadCandidatesMemo`, `rankCandidates`, `checkCandidates`; `init` goes; the `candidatesMemo` and `loadCanon` comments say profiles come from the match index |
| `core/database/promotealign` `promotealign.go` | D-9: `Propose` absorbs `propose`; reads `matchIndexOf` once |

```go
// shortlist
type filing struct {
	kind vocabkey.SubjectType
	keys []string // the index keys id is filed under, sorted
}
func (x *Index) Put(kind vocabkey.SubjectType, id []byte, v subjectcompare.Values)
func (x *Index) Remove(id []byte)

// promotealign
type matchIndex struct {
	handles   map[string]subjectcompare.Subject // every unmerged person, event and place, by handle id
	shortlist *shortlist.Index
}
var MatchIndex graphcache.Whole[*matchIndex]
func buildMatchIndex(g *graphcache.Graph) (*matchIndex, error)
func applyMatchIndex(g *graphcache.Graph, w graphcache.Whole[*matchIndex], roots [][]byte, _ effects.Set) error
func matchIndexOf(g *graphcache.Graph) (*matchIndex, error)
func (p *profiler) subject(n *graphcache.Node, built map[string]subjectcompare.Subject) (subjectcompare.Subject, error)
func (p *profiler) subjects(nodes []*graphcache.Node, built map[string]subjectcompare.Subject) error
func (p *profiler) readers(ids [][]byte) ([][]byte, error)
```

Errors: none new.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Measure the base, and alias the handle**
Files: `core/graphalign/types.go`; the ship log `S16-20-match-index-apply.ship.md` (beside this plan; the orchestrator creates it).
Do: on the planning base, run `BenchmarkLargeSourcePropose`, `BenchmarkCacheSizes` and `TestLargeSourceBudgets -v`, and write into the ship log, with `hostname`: `warm-ms`, `after-commit-ms`, `matchIndex-B` at both sizes, and the "warm at open" figure `TestLargeSourceBudgets` logs. Then D-11.
Check: `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'BenchmarkLargeSourcePropose|BenchmarkCacheSizes' -benchtime 1x -timeout 60m ./core/database/promotealign` prints both sizes; `go build ./... && CGO_ENABLED=1 go test -tags fts5 ./core/graphalign` passes.

**STEP-2 — The shortlist index can refile a handle**
Files: `core/shortlist/index.go`, `core/shortlist/shortlist_test.go`. Skill: `add-match-rule` (the package stays pure).
Do: D-4. `shortlist_test.go`'s `build` helper calls `Put`. Add `TestIndexPut` (T-4). `buildShortlist` in `promotealign` calls `Put` for now (it goes in STEP-3).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/shortlist ./core/database/promotealign` passes.

**STEP-3 — One entry holds every profile, built at open; proposals read it**
Files: `core/database/graphcache/cache.go`, `cache_test.go`; `core/database/promotealign/index.go`, `profiles.go`, `canon.go`, `promotealign.go`, `export_test.go`, `memo_test.go`, `promotealign_test.go`; `core/catalogsession/warm_test.go` (the `Peek` probes only). Skill: `use-catalog-session`.
Do: D-15 first (`cache.go`, with a `cache_test.go` case: `Peek` on an entry with a Build returns false and builds nothing). Then D-1, D-2, D-3, D-6, D-7, D-8, D-9, D-10, and D-14 except the two handle-commit flips (`TestMatchIndexDrops`, `TestDoKeepsWarmEntries`), which STEP-4 makes; Apply stays `w.Drop(g)` on any root until STEP-4. Order within the step: the profiler and the entry with its Build and Check (`profiles.go`, `index.go`), then the readers (`canon.go`, `promotealign.go`), then the test hooks and tests; `go build ./core/database/promotealign` after each. Delete every name **Deletes** lists for these files. Add `TestMatchIndexHoldsEveryHandle` (T-2).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign ./core/catalogsession` passes, and again with `PROVENENCIA_GRAPH_VERIFY=1`; `rg -n "indexGen|profilesPattern|keyValues|loadShortlist|shortlistInputs|newProfilerWith|hopPersonToEvent|ProfiledRefsForTest|ProfilesKeptForTest|MatchIndexGenForTest|\.keep\(\)" core` finds nothing; the goldens (`core/graphalign`) pass.

**STEP-4 — Apply patches the index instead of dropping it**
Files: `core/database/promotealign/index.go`, `profiles.go` (`readers`), `memo_test.go`, `core/catalogsession/warm_test.go`.
Do: D-5, D-13. `TestMatchIndexDrops` and `TestDoKeepsWarmEntries` per D-14 (the handle commit keeps the index). Add `TestMatchIndexFollowsCommits` (T-3) and the AC-1.3 subtest (T-5).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign ./core/catalogsession ./core/database/graphcache` passes, and again with `PROVENENCIA_GRAPH_VERIFY=1`.

**STEP-5 — Apply's fan-out is timed**
Files: `core/database/promotealign/perf_test.go`.
Do: D-12: `TestMatchIndexApplyFanOut` (T-6), on its own catalog: `graphcachetest.OpenCatalog`, `nameformats.Install` and `namealiases.Install` (as `newLargeFixture`), handles built with `graphcachetest.NewCanon` inside `Canon.Bulk` (`Person`, `Event`, `Join`, `Birth`), then `catalogsession.WarmOnOpen`; the edits and commits are T-6's.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestMatchIndexApplyFanOut -v ./core/database/promotealign` passes and logs both timings.

**STEP-6 — Measure after**
Files: none changed unless a figure fails a Stop condition.
Do: run the STEP-1 commands on the same machine; write the after figures beside the base in the ship log.
Check: `TestLargeSourceBudgets` passes; §11's limits hold. Every figure comes from a run with `PROVENENCIA_GRAPH_VERIFY` unset (`WithinBudget` skips its budget under Verify).

**STEP-7 — Docs**
Files: `docs/catalog-caches.md`, `docs/promote-graph-alignment.md`, `.agents/skills/use-catalog-session/SKILL.md`, `core/catalogsession/session.go` (comment), `docs/deployment-plan/spike-16/deployment-plan.md`, `docs/deployment-plan/spike-16/cache-architecture.md`, `docs/deployment-plan/spike-16/completed.md`, this plan.
Do:
- `catalog-caches.md`: the `promotealign.matchindex` row — Fill "Build (`buildMatchIndex`): every unmerged person, event and place profile and the shortlist postings filed from them; `promotealign.Warm` fills it at open"; Updated by "`DerivedFrom` the vocabulary, `namealiases.index` and `nameformats.default`; Apply re-profiles the commit's roots and the people at root events and refiles the handles whose profile changed, then `Set`"; size the STEP-6 `matchIndex-B` at both sizes. The `promotealign.candidates` row — Updated by "`DerivedFrom` `promotealign.matchindex`; Apply drops the graphs in `effects.Set.EvidenceGraphs`" (the generation clause goes). § Known exceptions: "None. Every registry entry updates from a named signal."
- `promote-graph-alignment.md` §8: the match index bullet → "holds every person, event and place profile and the shortlist postings filed from them (`promotealign.MatchIndex`), derived from the vocabulary, the alias index and the name format; a commit re-profiles its roots and the people at root events and refiles only the handles whose profile changed"; the candidates bullet → "dropped on `Set.EvidenceGraphs` for that graph and whenever the match index changes (they are derived from it), held at most 64 at a time"; the open-time warms bullet: "builds the shortlist index" → "builds the match index (every profile and the shortlist postings)", and its last sentence → "A handle commit patches the match index; nothing rebuilds it but a vocabulary, alias or name format change, or `Drop`." §7 item 2: "(`core/shortlist`, built by `loadShortlist`)" → "(`core/shortlist`, held in the match index)"; item 3's last sentence → "Handles carry the profiles the match index holds."
- `promote-graph-alignment.md` §12 (line 242, "Query count (§7)"): "After a handle or vocabulary commit drops the shortlist index, the next proposal reads it again (one read of every person, event and place node)." → "After a vocabulary, alias or name format change empties the match index, the next proposal builds it again (one read of every person, event and place node); a handle commit patches it."
- `catalog-caches.md` line 5: "The signal is a field of `effects.Set` that the entry's Apply reads, a moved key" → "The signal is a field of `effects.Set` that the entry's Apply reads (`Handles` through the commit's roots, as the map's rule says), a moved key"; and "A cache that empties itself on any write is a known exception, listed below." → "A cache that empties itself on any write would be a known exception, listed below."
- `add-match-rule` "Add a shortlist key" step 2: "`keyValues` builds no life (`TestShortlistSection`)" → "a derived lead's keys would move with every edit to the events it is read from, so an event edit would refile all its participants (`TestShortlistSection`)". `core/matchregistry/registry_test.go` `TestShortlistSection`: the comment's "the index builds no life" and the message's "the shortlist builds no life" → "its keys would move with every event edit".
- `use-catalog-session` line 62: "matching's shortlist kinds and shortlist index (`promotealign.Warm`)" → "the match index: every person, event and place profile and the shortlist postings (`promotealign.Warm`)"; "the shortlist index stays unbuilt" → "the match index stays unbuilt".
- `session.go`: the `WarmOnOpen` comment, "matching's shortlist kinds and index" → "matching's match index" and "leaves the shortlist index unbuilt" → "leaves the match index unbuilt"; the comment before `WarmOnOpen(c)` in `openResearcher` (line 112), "the shortlist index too when it builds" → "the match index too when it builds".
- Map: the S16-20 row's status (`Landed #N`, as landed rows read). The scope wording and the moved Check were changed when this plan was written. The rule "Every cache names the signal that updates it, and updates exactly": "Nothing empties itself on any handle commit; the match index's drop-on-any-handle Apply (S16-20) is the known exception." → "Nothing empties itself on any handle commit."
- `cache-architecture.md` § Promote: already reworded when this plan was written; nothing more.
- `completed.md`: one paragraph. This plan: status, Landed note.
Check: the DoD sweep finds nothing; `CGO_ENABLED=1 go test -tags fts5 -run 'TestCacheDoc' ./core/catalogsession` passes; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-20-match-index-apply.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/graphalign` all tests; `core/database/promotealign` `TestProposeWarmIssuesNoQuery`, `TestMemoFollowsCommit`, `TestCandidatesMemoFollowsCommits`, `TestMemoIsNotMutated`, the goldens (unchanged) | Their cases → today's results | AC-2.2, AC-2.3, AC-3.1, D-8, D-11 |
| T-2 | Go | `memo_test.go` `TestMatchIndexHoldsEveryHandle` *(new)* | Warm fixture plus a place (`f.Place(f.source, f.artifact, "Springfield")`, `f.promote`); `Drop`, `Warm` → `HandlesHeldForTest` equals the count of unmerged person, event and place nodes (`g.Kind` of each); two proposals on Source B → the held `*matchIndex` pointer is the same before and after, and `g.Verify()` with `PROVENENCIA_GRAPH_VERIFY=1` returns nil (no proposal changed a shared Subject); after `Warm`, the held Subject of `personHandle` has a non-empty `vocabkey.PropertyLifespan` value (Build read its birth) | AC-2.1, AC-2.2, D-2, D-9 |
| T-3 | Go | `memo_test.go` `TestMatchIndexFollowsCommits` *(new)*, table; each row on a fresh `newWarmFixture`, `Warm`, then the write, then: `MatchIndex` held, `g.Verify()` nil, next proposal on Source B `reflect.DeepEqual` to one after `Drop` | rows: "a new filed person" (`filedBirth` on a new Source, "Jane Doe" 1820); "a name edited" (AC-1.1: `before := HandleForTest(g, personHandle)`; `renameMember(t, w, "Jon Smyth")`; `after := HandleForTest(g, personHandle)` → `after.Values` differ from `before.Values`, and `IndexedForTest(g, "person", after.Values)` contains `personHandle`; Verify proves no stale posting is left); "an event date edited" (`w.f.Cite(w.f.artifact, w.memberBirth, observations.Input{PropertyID: w.f.Prop("date").ID, Date: datevaluetest.YMD(1818, 0, 0)})`, a second date on the filed birth); "a participation added" (`d := w.f.DatedOn(w.f.source, w.f.artifact, "death", 1880, 0, 0)`; `w.f.Participation(w.f.source, w.f.artifact, w.member, d, "subject")`; `w.f.promote(d)`, which files the participation onto the person's handle); "a handle merged" (as `TestMemoFollowsCommit`'s, committing the merged handle; it leaves `handles` and every posting); "a Subject unpromoted" (`promote.Unpromote` of the member, which deletes the handle); "a place renamed" (a place made and promoted as in T-2, then a second toponym "Shelbyville" cited on its Subject with `w.f.Cite`); "an event deleted" (`promote.Unpromote` of `memberBirth`, which deletes the birth handle and its participation: the person is re-profiled with no life) | AC-1.1, AC-1.2, D-5, D-10 |
| T-4 | Go | `core/shortlist/shortlist_test.go` `TestIndexPut` *(new)*, table | Put A "James Robins", Put B "Robins" → `Lookup("Robins")` [A, B]; Put A again as "James Smith" → `Lookup("Robins")` [B], `Lookup("Smith")` [A]; Remove B → `Lookup("Robins")` nil and the index `reflect.DeepEqual` to one built by `Put` of A "James Smith" alone; Put of a value with no keys → not filed; Put A filed, then Put A with no keys → unfiled and equal to an index without A; Put A as a person, then as a place → filed only under the place keys; Put A twice with the same value → equal to Put once; ids put in order 3, 1, 2 under one key → `reflect.DeepEqual` to put in order 1, 2, 3 | D-4 |
| T-5 | Go | `memo_test.go` `TestMatchIndexFollowsCommits` subtest "a commit that changes no profile" *(new)* | Propose on Source B (candidates held); `OnCommit(Set{Handles: [personHandle]})` with nothing written → the same `*matchIndex` held, `CandidatesHeldForTest(g, graphB)` true | AC-1.3, D-5, D-7 |
| T-6 | Go | `perf_test.go` `TestMatchIndexApplyFanOut` *(new)*, table | inside `Canon.Bulk`: `e := k.Event(vocabkey.EventMarriage)` and 500 persons each `k.Join(k.Person(), e, "witness")`; then `k.Date(e, "date", datevaluetest.YMD(1850, 0, 0))` and the timed `k.Commit(e)` → ≤ 250 ms. Inside `Canon.Bulk`: `f := k.Person()` and 2,000 `k.Birth(k.Person(), f, nil)`; then `k.Retype(f, "gender", vocabkey.GenderMan)` and the timed `k.Commit(f)` → ≤ 250 ms. `g.Verify()` returns nil after each | D-12 |
| T-7 | Go | `memo_test.go` `TestMatchIndexDrops` (per D-14), `TestWarmBuildsTheIndex` (per D-14), `TestVerifyChecksMatchIndex` (with the new `EmptyShortlistForTest`) | handle commit → held; vocabulary, alias index, `Drop` → emptied; Source-only and Source-vocabulary commits → held; emptied shortlist → error naming `promotealign.matchindex` | AC-2.3, AC-2.4, D-10 |
| T-8 | Go | `perf_test.go` `TestLargeSourceBudgets`, `bench_test.go` `BenchmarkLargeSourcePropose`, `BenchmarkCacheSizes` | 4,000 → budgets pass; both sizes → figures in the ship log | AC-1.4 |
| T-9 | Go | `core/catalogsession/warm_test.go` `TestDoKeepsWarmEntries` (per D-14: "handle commit" keeps the index), `cache_doc_test.go` `TestCacheDocListsEveryEntry`, `TestCacheDocSeesEveryOwner` | → pass | D-2, D-5 |
| T-10 | Go | `core/matchregistry/registry_test.go` `TestShortlistSection` (comment and message per D-8) | → pass | D-8 |

Must stay green: everything `scripts/check-pr.sh` runs, and the Go suites with `PROVENENCIA_GRAPH_VERIFY=1`. Most likely to break: `core/database/promotealign` (memo tests, goldens, budgets), `core/catalogsession` (warm and cache doc).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] The promotealign, shortlist, graphcache and catalogsession suites pass with PROVENENCIA_GRAPH_VERIFY=1
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `rg` over `core/`, `.agents/`, `.cursor/rules/`, `docs/catalog-caches.md`, `docs/promote-graph-alignment.md` and `docs/matching-model.md` finds none of `loadShortlist`, `indexGen`, `shortlistInputs`, `keyValues`, `hopPersonToEvent`, `ProfiledRefsForTest`, "builds no life", "drops the shortlist index", "match index generation", "drop-on-any-handle"; `scripts/check-guidance-names.py` passes (check-pr runs it). Map and review history keep their old text
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory (no UI)
- [ ] Version bump: none
- [ ] Ship log: STEP-1 and STEP-6 figures with `hostname`
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the rule; completed.md: one short paragraph of what shipped
- [ ] Domain docs updated: docs/catalog-caches.md, docs/promote-graph-alignment.md §7 and §8, use-catalog-session, cache-architecture.md
- [ ] PR description: what shipped, AC evidence, deviations, the before / after figures
```

## 10. Manual check

1. Open a project with filed people (the dogfood project). Open Promote on a Source with unpromoted people: the suggestions show as before.
2. In another window, edit a filed person's name. Back in Promote, re-open the same Source: the suggestions match what a project reopened from scratch shows for that Source (close and reopen the project, open Promote on it, compare the rows).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A patched profile misses a dependency (a link or role a profile reads) | T-3 with Verify over every write kind; Verify compares with a fresh Build |
| Open grows | STEP-1 / STEP-6 figures; limit below |
| Memory | `matchIndex-B` recorded; limit below |
| A test reads a field this PR removes from outside `promotealign` | `rg` in STEP-3's check |

**Stop and report if:**

- On the large fixture, a handle's keys from its full profile differ from today's `keyValues` keys (T-1's goldens or `TestLargeSourceBudgets` "fixture" change), so D-8 does not hold.
- `matchIndex-B` at 16,000 persons is over 60 MB, or the Go caches' total at 16,000 passes 250 MB.
- `WarmOnOpen` at 16,000 persons takes more than 400 ms longer than on the base, or the `TestLargeSourceBudgets` "warm at open" or "warm" subtest fails where the base passed.
- `after-commit-ms` at either size is higher than on the base (figures from runs with `PROVENENCIA_GRAPH_VERIFY` unset).
- `TestMatchIndexApplyFanOut` passes its ceilings only by changing D-5.
- A Promote golden or `TestAlignInvariants` changes.
- `register` order cannot be met by one `init` in `index.go` (D-6).

## 12. Open questions

None.

## 13. PR

**Title:** S16-20 — One match index that Apply keeps current
**Body:** objective in one sentence; what shipped against each US; AC evidence; the ship log's before / after figures; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-20-match-index-apply.md`.
