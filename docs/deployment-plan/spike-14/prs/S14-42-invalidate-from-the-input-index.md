# S14-42 — Invalidate from the input index

**Status:** Planned
**Spike / slice:** Spike 14 · Slice 8 — Summary rows know their inputs
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the app reloads what `summary_handles` names, as today; no view looks different)
**Board:** —
**Depends on:** S14-40 (on the same branch, [#746](https://github.com/mendahu/provenencia/pull/746), not merged)
**Blocks:** —
**Branch:** `s14-40-build-the-input-index` — the same PR as S14-40 (user, 2026-11-08: "S14-42 on the same pr")
**Version bump:** patch — a researcher-visible bugfix (stale search hits after Unpromote and after a link delete) ([`docs/versioning.md`](../../../versioning.md)): `0.4.0` → `0.4.1`
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md) (an unreleased step is amended), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md) (`subjects.Delete`), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [design § Invalidation](../summary-inputs.md#invalidation); working rules "(Batch 8) The structural fix comes before the patch", "(Batch 8) Summary rows record their inputs; invalidation is a lookup", "(Batch 8) A row is its own input", "(Batch 8) One input index, owned by `searchindex`", "(Batch 8) Every write tells the app what it moved through `summary_handles`"; S14-38's scope entry (its scenarios are this PR's acceptance tests); S14-40's plan, its **Landed** note and ship log
**Planned:** 2026-11-08 against `s14-40-build-the-input-index` @ `3a4cc943` (S14-40 complete on the branch; `main` @ `78f26812`)
**Review:** [`S14-42-invalidate-from-the-input-index.review.md`](S14-42-invalidate-from-the-input-index.review.md) — Fits with changes — Ready with changes, 2 rounds; no open blockers; To user none

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

A write finds the summary rows and search documents it moved by looking them up in `catalog_search_inputs`, the index S14-40 built, instead of walking a hand-written inverse of the summary reads. That fixes the stale search hits S14-38 found: after Unpromote, the event or person at the other end of a filed link kept its old text ("Birth of James Robins" where a rebuild reads "Unspecified birth"), because the link's rows were gone before the walk ran. A link delete at a city with thousands of events rewrites only the event it touched and tells the app to reload the city page.

**Why now.** S14-40 put the index in place and proved it against isolated composition. Working rule "(Batch 8) The structural fix comes before the patch" descoped S14-38 into this PR.

**What stays.** Layer 1 (`auto_reconciler_values`) and the effects registry; `Set.Search` for Source and Evidence graph documents; the wire (`WriteEffects.summary_handles`); `graphcache` `detachAssociation` and `mergedNeighbors` for node reloads and memo roots; the kinship memos; every composed row and document text for unchanged data.

## 2. User stories

**US-1.** As a researcher, when I unpromote a Subject that has filed links, every search hit and list row that showed it changes to what a rebuild shows, so the omnibar and the lists never show a link that is gone.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | James Robins (person) and a birth dated `1885-05`, both promoted, a filed participation (role subject). Unpromote James | The birth's search document equals a rebuild's (`Unspecified birth`, no `James Robins`), and the write's `SummaryDeps` holds the birth |
| AC-1.2 | The same world. Unpromote the birth instead | James's search document equals a rebuild's (no `1885-05`), and the write's `SummaryDeps` holds James |
| AC-1.3 | Each of: Unpromote either end of a participation; Unpromote a location's place; Unpromote a part_of's whole; Unpromote either end of a succeeded_by | Handle documents and `catalog_search_inputs` equal a rebuild after the write (`assertUpkeepEqualsRebuild`) |

**US-2.** As a researcher, when I delete a link at a place with many located events, only the event the link names is rewritten and an open page for the place reloads, so the delete is correct and stays cheap.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A place with three located, promoted events. `subjects.Delete` of one event's location Subject | Documents and inputs equal a rebuild; the write's `SummaryDeps` holds that event and the place, and none of the other two events |
| AC-2.2 | The large fixture's city with about 6,000 located events. Unpromote one located event; delete one event's location Subject | Each write finishes under its budget (D-12) |

**US-3.** As a contributor, I have one way to find what a write moved, so a new read in a composer needs no matching edit anywhere else.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | T-11's `rg` over `core` and `api` | No hit |
| AC-3.2 | `subjects.Delete` of a Subject with released facets | It calls no `autoreconciler` function; documents and inputs equal a rebuild |
| AC-3.3 | A town three part_of steps under a country, a birth located at the town, the person whose birth it is; displays stored. Rename the country | The stored person, birth and town displays are dropped; the next read shows the new name (the moved `TestDisplayCacheFollowsAPlaceThreeStepsUp`) |
| AC-3.4 | A promoted person with two likenesses. `canonicalentities.SetPreferredLikeness` to the second | The person's stored display is dropped and its search document is rewritten with the new lead picture, although no value row moved |
| AC-3.5 | `graphcachetest` kit: a person with a birth, displays stored. `Merge` the birth into a second birth and `Commit` | The person's stored display is dropped (it read the merged birth) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `searchindex` finds the rows a write moved from the index, reprojects them with every written handle on one graph, and returns them with the changed links' ends | STEP-3, AC-1.1–1.3, AC-2.1 |
| `catalog_search_inputs` marks an association's end rows (amends the unreleased `000062.sql`) | STEP-2, D-4 |
| `autoreconciler.RecomputeReported` calls it; `AssociationEndpoints`, `Dependents` and the `movedRoots` block leave it; `Rebuild` reprojects every handle directly | STEP-4 |
| S14-38's scenarios as rebuild-equals-upkeep scenarios; the reproduction; preferred picture | STEP-4, AC-1.1–1.3, AC-2.1, AC-3.4 |
| Display rows drop from `SummaryDeps` only; `DepsFunc`, `SetDisplayDeps`, `forgetDisplays`, `SummaryDepsWalked` deleted; the kit passes `SummaryDeps`; kit merge | STEP-5, AC-3.3, AC-3.5 |
| `subjects.Delete` stops snapshotting and recomputing; `HandlesObservingSubject` and `linkedHandles` deleted | STEP-6, AC-3.2 |
| `Dependents`, `walkDepth`, `birthOrDeath`, `AssociationEndpoints`, `linkEnds` and `dependents.go` deleted; their tests moved onto writes | STEP-7, AC-3.1 |
| S14-33 budgets plus an Unpromote and a location delete at the city | STEP-1, STEP-4, STEP-8, AC-2.2 |
| Docs and version bump | STEP-9 |

## 4. Current state

Verified at `3a4cc943`.

| Area | Where | Today |
| --- | --- | --- |
| Run | `core/writes/run.go` `Run` | After `audit.Record`, calls `autoreconciler.RecomputeReported(tx.Tx, set.Handles)`, sets `set.SummaryDeps` and `set.SummaryDepsWalked = true` |
| Recompute | `core/database/autoreconciler/autoreconciler.go` `RecomputeReported` | `AssociationEndpoints` before and after `recomputeBatch`; `Dependents(roots)`; narrows to `movedRoots` when not every id moved; `searchindex.ReprojectHandles(roots ∪ ends ∪ reproject)`; returns `deps ∪ ends`. `RecomputeTx` wraps it; `Rebuild` calls `RecomputeTx` with every id |
| Moved | same file `recomputeBatch` | Returns the ids whose displayed rows differ (`graphcache.SameDisplayRows`): a created or deleted handle with values counts as moved |
| Inverse | `core/database/conclusionsummaries/dependents.go` | `Dependents`, `walkDepth`, `birthOrDeath`, `AssociationEndpoints`, `linkEnds`; also holds `summaryChainDepth`, `maxInputKeysPerCall`, `inputKeyCap` |
| Display drop | `core/database/graphcache/commit.go` `dropCommittedDisplays` | With `SummaryDepsWalked`, drops `roots ∪ SummaryDeps`; otherwise `forgetDisplays(roots)` (`display.go`), which calls the registered `depsFn` |
| Registration | `core/database/conclusionsummaries/display_register.go` `init` | `SetDisplayDeps(Dependents)` and `SetDisplayCheck(checkDisplay)` |
| Set | `core/database/effects/registry.go` `Set.SummaryDeps`, `Set.SummaryDepsWalked` | Run fills both |
| Wire | `api/ffi/handlers/write_effects.go` | `SummaryHandles: idStrings(set.SummaryDeps)` — unchanged |
| Index | `core/database/searchindex/inputs.go` `writeInputs`; `core/database/migrations/000062.sql` | Rows `(dependent_id, input_id, bridge, endpoint)`; an association's rows are its ends (`graphcache.Graph.AssociationInputs`); index on `(input_id, bridge, endpoint)`. Nothing reads it |
| Reproject | `core/database/searchindex/handles.go` `ReprojectHandles` | One graph per call (`newGraph`), batches through `reprojectHandleBatch`, writes documents and inputs |
| Subject delete | `core/database/subjects/subjects.go` `Delete` | Snapshots `HandlesObservingSubject`, `linkedHandles`, `Dependents(ends)` before `ReleaseFacets`; after the `DELETE` calls `autoreconciler.RecomputeTx(released ∪ inbound ∪ ends ∪ deps)` |
| Observers | `core/database/autoreconciler/autoreconciler.go` `HandlesObservingSubject` | Only caller is `subjects.Delete`. `observations.value_subject_id → subjects` is `NO ACTION`, `BucketResource` (`catalogmodel/model.go:168`), so a Subject an Observation points at is refused; the set is empty for every delete that succeeds |
| Unpromote | `core/database/promote/unpromote.go` `Unpromote` | Releases the Subject's accepted claim (`ReleaseClaim`, deletes the handle if last member) and unfiles every bridge with it as an end (`identityclaims.UnfileBridges`), so the participation's association handle is deleted too |
| Test kit | `core/database/graphcache/graphcachetest/graphcachetest.go` `observe`, `Commit`, `commit`, `Merge`, `Remove` | `observe` calls `RecomputeTx`; `Commit(ids)` sends `effects.Set{Handles: ids}` with no `SummaryDeps`, so `forgetDisplays` walks `Dependents` |
| Tests on `Dependents` | `core/database/conclusionsummaries/walks_test.go` `TestCanonicalWalks` (last block); `placechains_test.go` `TestDependentsIncludeChildPlaces`; `display_cache_test.go` both tests; `filedlinks_test.go` three tests via `filedLinkDeps` (calls `RecomputeReported` a second time); `core/search/handles_test.go` `TestJoinReprojectsDependentsOnlyWhenTheRootMoves`; `core/database/conclusionsummaries/inputs_test.go` `TestInputsCoverDependents`; `core/database/conclusionsummaries/conclusionsummariestest/inputs.go` `CheckCoversDependents`, `excusable`; `core/database/promotealign/perf_test.go` subtest "inputs cover Dependents" | Hold today's walk |
| Upkeep tests | `core/database/autoreconciler/autoreconciler_test.go` `TestRebuildEqualsUpkeep_Scenarios`, `TestRebuildEqualsUpkeep_SeededSequences`, `assertUpkeepEqualsRebuild` | Compare documents and inputs with `searchindex.RebuildAll`. The fixture embeds `evidencelayertest.Records` (`Participation`, `Locate`, `PlaceRel`) |
| Budgets | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` | "locate an event at the city" files a location on a promoted event at `lf.citySubject` (`budgetLocateAtCity` 500 ms) |
| Docs | `.cursor/rules/catalog-queries.mdc:17`; `core/writes/README.md:15`; `docs/catalog-deletes.md:46`; `core/database/autoreconciler/autoreconciler.go` package doc lines 15–18 and `RecomputeTx` doc; `subjects.Delete` doc; `docs/ideas/structural-debt.md` rows 19 and 25 | Name the walk, the snapshot, or `RecomputeTx` in Subject delete |

**Precedent.** S14-40 (on this branch) for the index and `ReprojectHandles` composing on one graph; follow its `writeInputs` statement shape for the new lookups (`sqlbatch.Query`, `database.SQLInPlaceholders`). Do not wrap the old walk: the index replaces it.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `searchindex.ReprojectChanged(q Querier, written, moved [][]byte) ([][]byte, error)` *(new)* in `core/database/searchindex/inputs.go` does the lookup, the reprojection and the index write. `recompute(q, ids) (moved [][]byte, err error)` *(new, unexported)* in `autoreconciler.go` runs `recomputeBatch` per `sqlbatch` batch and collects `moved`. `autoreconciler.RecomputeReported` keeps its signature: `recompute`, then `searchindex.ReprojectChanged(q, ids, moved)`. `Run` and the kit reach it through `RecomputeReported`. `Rebuild` clears the tables, calls `recompute` over every id, and calls `searchindex.ReprojectHandles(q, ids)` directly. `RecomputeTx` is deleted: after this PR no production code calls it, and its test callers (`autoreconciler_test.go:1911`, `conclusionsummaries_test.go:273`, `:1172`) call `RecomputeReported` | The index's owner reads and writes it (working rule "One input index, owned by `searchindex`"). A rebuild moves every row and reprojects every handle, so the lookups would find nothing outside `ids` (review M-4) | Logic in `writes.Run` — the kit would need a copy; `Rebuild` through `ReprojectChanged` — about 37,000 dead lookups on every cache rebuild |
| D-2 | **Changed keys.** `ReprojectChanged` builds one graph, `g := newGraph(q)`, on the written data. (a) `gone` = ids of `written` whose `g` node is nil or `Merged` (the test `graphcache.Graph.AssociationInputs` uses); every index row whose `input_id` is in `gone` names a changed dependent, whatever its bridge and endpoint. (b) `moved` ids not in `gone`: the values key `(id, '', '')`. (c) For every id in `moved` ∪ `gone`, the link keys `(end, bridge, endpoint)` of its old ends (its index rows with `dependent_is_association = 1`; a summary row has none) and, for a live association, of its new ends (`g.AssociationInputs`). A handle in `written` that is neither moved nor gone adds no key | Design § Invalidation step 1. Values moving changes no link, so a moved person or event adds no link key: an event's own place-chain keys would reach the city's other 6,000 events. One reader of "gone or merged" (review M-1) | Link keys of every moved id — reprojects the city on an event's date edit; a raw `canonical_entities` read for `gone` — a second definition of "gone" |
| D-3 | **Order.** `ReprojectChanged` reads every key's dependents and every id's old ends from the index **before** it reprojects, because reprojecting deletes a gone or rewritten dependent's rows | A deleted association's ends exist only in its index rows until it is reprojected | Lookups after — loses every deleted link's ends (S14-38 again) |
| D-4 | `000062.sql` *(amended; unreleased on this branch, `add-catalog-migration` step 2)* gains `dependent_is_association INTEGER NOT NULL DEFAULT 0 CHECK (dependent_is_association IN (0, 1))`. `writeInputs` collects one slice of `inputRow` *(new unexported struct: `dep []byte`, `key graphcache.InputKey`, `association bool`)* in place of its parallel `rows` / `deps` slices, and inserts five columns: `1` on the rows from `AssociationInputs`, `0` on a summary's rows. `inputRowsPerInsert` drops to 160 (5 × 160 = 800 binds, as today). `ProjectionVersion` stays 11 | A gone handle's type is gone with its row, and a summary row's link keys look like an association's ends; the flag says which rows are ends | Inferring "an association never holds its own id" — an unstated invariant; reading types from the change list — couples the index to change shapes; a second table — two writers for one owner |
| D-5 | **Reprojected and reported sets.** `ReprojectHandles` splits into the exported entry point (builds its graph, as today) and `reprojectOn(q Querier, g *graphcache.Graph, ids [][]byte) error` *(new, unexported)*. `ReprojectChanged` calls `reprojectOn(q, g, written ∪ dependents)` on the same `g` (working rule "A row is its own input"). Its return value, which `Run` stores as `set.SummaryDeps`, is `dependents ∪ ends`, where `ends` is the old and new end ids of (c). Ends are not reprojected unless they are dependents | One graph per write, the rule S14-40 set (review M-1). A place does not read its location links, so the city's document does not change; the app still reloads its page from `summary_handles` | Reprojecting ends — rewrites the city's document for nothing (today's code does); `ReprojectHandles` from `ReprojectChanged` — a second graph |
| D-6 | **Display drops.** `dropCommittedDisplays` always drops `roots ∪ set.SummaryDeps`. Deleted: `DepsFunc`, `depsFn`, `SetDisplayDeps`, `forgetDisplays` (`core/database/graphcache/display.go`), `Set.SummaryDepsWalked` (`core/database/effects/registry.go`) and its assignment in `Run`, and the `SetDisplayDeps` call in `display_register.go` (`SetDisplayCheck` stays). `set.SummaryDeps`'s doc says it is the rows the input index found plus the changed links' ends | Working rule "(Batch 8) Every write tells the app what it moved through `summary_handles`"; one source for app reloads and cache drops | Keeping `forgetDisplays` as a fallback — a second inverse |
| D-7 | **Kit.** `graphcachetest.Canon` gains `pending [][]byte` *(new field)*. `observe` calls `RecomputeReported` and appends its result to `pending`. `Commit(ids...)` calls `autoreconciler.RecomputeReported(k.db, ids)` (which sees raw `Remove`, `Merge`, `Provisional` writes) and commits `effects.Set{Handles: ids, SummaryDeps: pending ∪ that result}`; `pending` is then emptied. A test that never commits leaves `pending` unread. `CommitVocabulary` is unchanged | The kit writes rows outside `Run`, and `observe` keeps values current at each Observation (its doc: "rewrites the entity's kept rows from its Observations"), so a recompute at `Commit` alone sees no moved row for an observed entity | Committing through `writes.Run` — the kit writes raw SQL by design; deferring `observe`'s recompute to `Commit` (review M-5) — changes when every kit user sees values |
| D-8 | **Subject delete.** `subjects.Delete` keeps `Impact`, `Refuse`, `ReleaseFacets`, the `DELETE` and its change; it drops `HandlesObservingSubject`, `linkedHandles`, `Dependents`, `effects.Handles(released)` and `RecomputeTx`. `autoreconciler.HandlesObservingSubject`, its `sqlHandlesObservingSubject` and `subjects.linkedHandles` are deleted. `Run` resolves the released claim changes: `identity_claims` → `entity_id` names the association handle, which D-2 (c) turns into its ends | A Subject an Observation points at cannot be deleted (`core/database/catalogmodel/model.go:168`, `BucketResource`), so `HandlesObservingSubject` is empty for every delete that succeeds; the ends are the released association's ends, which the index holds | A `subjects` registry path for observers — a path that can never resolve a row |
| D-9 | `dependents.go` is deleted. `summaryChainDepth`, `maxInputKeysPerCall` and `inputKeyCap` move unchanged, with their docs, to the top of `placechains.go`; the `summaryChainDepth` doc drops "and how far the reverse of that read walks back down" | Their last neighbour goes | A `consts.go` — one more file for three lines |
| D-10 | **Index reads**, all unexported in `inputs.go` beside `writeInputs`: `dependentsOf(q, keys []graphcache.InputKey) ([][]byte, error)` (exact keys, by `catalog_search_inputs_input_idx`); `dependentsOfHandles(q, ids [][]byte) ([][]byte, error)` (every key whose `input_id` is in `ids`, same index); `oldEnds(q, ids [][]byte) (map[string][]graphcache.InputKey, error)` (rows with `dependent_id` in `ids` and `dependent_is_association = 1`, by the primary key). Statements: `dependentsOf` — `SELECT DISTINCT dependent_id FROM catalog_search_inputs WHERE (input_id, bridge, endpoint) IN (VALUES (?, ?, ?), …)`, 250 keys (750 binds) per statement; `dependentsOfHandles` — `… WHERE input_id IN (…)` through `sqlbatch.Query`; `oldEnds` — `SELECT dependent_id, input_id, bridge, endpoint … WHERE dependent_is_association = 1 AND dependent_id IN (…)` through `sqlbatch.Query`. Results sorted and unique. `export_test.go` exposes them as `DependentsOfForTest`, `DependentsOfHandlesForTest`, `OldEndsForTest` (the tests are `package searchindex_test`) | Three questions, three reads, one place (review M-2, M-3); only `ReprojectChanged` calls them | Exported lookups — tests outside the owner would encode the composers' keys a second time |
| D-11 | **Moved tests** assert the `Result.Effects.SummaryDeps` of a write, read from D-13's `Records.Reported`: <br>• `display_cache_test.go` `TestDisplayCacheFollowsAPlaceThreeStepsUp`: the country rename it already makes → holds the town, the birth and the person. <br>• `TestBirthBecomingMarriageDropsThePerson`: its `update_observation` → holds the person. <br>• `placechains_test.go` `TestDependentsIncludeChildPlaces` → `TestPlaceRenameReachesChildPlaces`: cite a second toponym on Ontario through `writes.Run` → holds Toronto. <br>• `walks_test.go` `TestCanonicalWalks` last block (lines 253–275): a toponym cited on York (`f.Cite`) → holds the birth, the fire and James; a name cited on James → holds the birth, the death, the marriage and the census; a date cited on the birth → holds James; the negative assertions go (T-6 and the budgets hold precision). <br>• `filedlinks_test.go`: `filedLinkDeps` is deleted; each test reads `f.Reported` right after the `f.Participation`, `f.Locate` or `f.PlaceRel` that filed the link. <br>• `core/search/handles_test.go` `TestJoinReprojectsDependentsOnlyWhenTheRootMoves` stays unchanged and must pass. <br>• `TestInputsCoverDependents`, `CheckCoversDependents`, `excusable` and the "inputs cover Dependents" subtest are deleted; S14-40's D-13 retires. <br>Where an expected id is missing, S14-40's `TestInputsMatchIsolatedComposition` decides: if it passes for that row, the expectation goes | The one invalidation path the app uses is what tests hold (review M-3, X-4) | Hand-picked keys passed to an exported lookup — a second copy of the composers' key choices |
| D-12 | **Budgets.** `budgetUnpromoteAtCity` *(new)* = 500 ms and `budgetDeleteLocationAtCity` *(new)* = 500 ms in `perf_test.go`, warm, fastest of 3, as `budgetLocateAtCity`. STEP-1 adds the subtests first and logs the base figures | The same rows as filing one location at the city: one event, one association, the city reported, not reprojected | No budget — the 6,000-event fan-out D-2 avoids would pass |
| D-13 | **Write reports in tests.** `evidencelayertest.Records` gains `Reported [][]byte` *(new field)*: the `Result.Effects.SummaryDeps` of the last write a Records helper made, and `func (r *Records) Run(action string, fn func(tx *database.Tx) ([]rowchange.Change, error)) [][]byte` *(new)*: `writes.Run` with `writes.Op{Action: action, UserID: UserID}`, sets and returns `Reported`. `Cite`, `Participation`, `Locate`, `PlaceRel` and `Relationship` make their write through `writes.Run` and set `Reported` (`Cite` calls `citations.CreateWithObservations` inside it, as `observationstest.Cite` does; `observationstest.Cite` is unchanged). Scenario actions: `unpromote_subject` (`promote.Unpromote`), `delete_subject` (`subjects.Delete`), `set_preferred_likeness` (`canonicalentities.SetPreferredLikeness`). A location Subject is the return value of `Records.Locate` | One helper for every fixture that embeds `Records` (`autoreconciler_test.go`, `conclusionsummaries_test.go`); action strings from `api/ffi/handlers/promote.go:56`, `subjects.go:110`, `conclusion_detail.go:69` | A `run` helper per fixture — two copies; changing `observationstest.Cite` — many callers that need no Result |

**Rules this plan bends.** None. `catalog-queries.mdc` line 17 changes because its subject goes ("`canonicalgraph.Walk` stays for dependent discovery"); the rule's other uses stay. The map's working rule "One input index, owned by `searchindex`" says the index is "read by the graph cache's display drops"; STEP-9 makes it "read by `searchindex.ReprojectChanged`, whose set the graph cache's display drops use", which is what the rule meant.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `searchindex.ReprojectChanged` *(new)* | func | `autoreconciler.RecomputeReported`'s walk; `searchindex.ReprojectHandles` | It replaces the walk; it reprojects through `reprojectOn` |
| `searchindex.reprojectOn` *(new, unexported)* | func | `ReprojectHandles` | Split out of it so both entry points share one body and `ReprojectChanged` passes its graph |
| `dependentsOf`, `dependentsOfHandles`, `oldEnds` *(new, unexported)* | func | `conclusionsummaries.Dependents` | That walks the live graph; these read the index (`rg -n 'catalog_search_inputs' core`: only `inputs.go`, `migrations`, `catalogmodel` and tests name it) |
| `inputRow` *(new, unexported)* | struct | `writeInputs`'s parallel slices | Carries the flag; replaces the slices (S14-40 nit MNT-7) |
| `catalog_search_inputs.dependent_is_association` *(new column)* | column | `catalog_search_inputs` (S14-40) | Amends the same unreleased step |
| `Canon.pending` *(new)* | field | `Canon.rev` | — |
| `Records.Run`, `Records.Reported` *(new)* | test helper | `writes.Call` inside `Records` helpers | `Call` drops the Result |
| `autoreconciler.recompute` *(new, unexported)* | func | the batch loop in `RecomputeReported` | `Rebuild` and `RecomputeReported` share it |
| `budgetUnpromoteAtCity`, `budgetDeleteLocationAtCity` *(new)* | const | `budgetLocateAtCity` | Separate writes, separate budgets |

**Builds on.** S14-40's index, `graphcache.Graph.AssociationInputs`, `sqlbatch.Query`, `ReprojectHandles`; the effects registry's `identity_claims` → `entity_id` path for released claims.
**Deletes.** `conclusionsummaries.Dependents`, `walkDepth`, `birthOrDeath`, `AssociationEndpoints`, `linkEnds`, `sqlBirthOrDeath` and `dependents.go`; the `movedRoots` block in `RecomputeReported`; `autoreconciler.RecomputeTx` (its last production callers go: `Rebuild`, `subjects.Delete`, the kit's `observe`); `graphcache.DepsFunc`, `depsFn`, `SetDisplayDeps`, `forgetDisplays`; `effects.Set.SummaryDepsWalked`; `autoreconciler.HandlesObservingSubject`, `sqlHandlesObservingSubject`; `subjects.linkedHandles`; `conclusionsummariestest.CheckCoversDependents`, `excusable`; `TestInputsCoverDependents`; the "inputs cover Dependents" subtest; `writeInputs`'s parallel slices. The `canonicalgraph` and `hops` imports go from files that no longer use them.
**Next adds.** Kinship memos can read the index instead of `DropOnHandleCommit` roots ([`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md)); a new searchable kind needs only its composer's scope reads.

### Reach

| New path or state | Destination | Checked |
| --- | --- | --- |
| A write on a batch whose inputs were over `maxInputKeysPerCall` | `ReprojectChanged` | Finds no dependents for that batch's places; only a crafted catalog reaches it (S14-40 F-1) |
| A gone association with no index rows (never reprojected) | D-2 (c) | Old ends empty; new ends empty; its dependents still come from (a) |
| `Rebuild` | `ReprojectHandles(all)` | Every handle reprojected, as today |

### Scale

At the map's Scale (37,000 Subjects; the large fixture's city with about 6,000 located events): a location delete at the city looks up the association's two end keys and `input_id` = the association: one event row and the city (reported). An Unpromote of a located event: `input_id` ∈ {event, its associations}: the event's own rows and its participants. No lookup reads `(city, location, place)`'s dependents, because no row records it. A place rename at the city reads `(city, '', '')`: every event and person under it, as the walk does today. Each lookup is one indexed query per 500 keys.

### Edges

- **Assumptions.** Every reprojected dependent writes its rows again (S14-40 `writeInputs` deletes then inserts), so an index row is never older than its dependent's last reprojection; the rebuild-equals-upkeep tests hold it (T-7).
- **Neighbors.** An association whose values did not move (a second record of a filed link) adds no key and no end (D-2), as `movedRoots` did. A created association moves (no stored rows before), so its new ends' link keys reach the rows that will now read it.
- **Removed inputs.** A deleted handle's dependents keep their rows until reprojected in the same write (D-3). A merged handle is treated as gone.
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** The keys come from the index the app wrote; a crafted catalog is bounded by S14-40's per-call cap.

### Schema

Amended `core/database/migrations/000062.sql` (unreleased):

```sql
CREATE TABLE catalog_search_inputs (
    dependent_id BLOB NOT NULL,
    input_id BLOB NOT NULL,
    bridge TEXT NOT NULL DEFAULT '',
    endpoint TEXT NOT NULL DEFAULT '',
    dependent_is_association INTEGER NOT NULL DEFAULT 0 CHECK (dependent_is_association IN (0, 1)),
    PRIMARY KEY (dependent_id, input_id, bridge, endpoint)
) STRICT, WITHOUT ROWID;
```

The index and the `catalogmodel` row are unchanged.

### Go core

| Package | Change |
| --- | --- |
| `core/database/searchindex` | `ReprojectChanged` *(new)*; unexported `reprojectOn`, `dependentsOf`, `dependentsOfHandles`, `oldEnds`, `inputRow` *(new)*; `writeInputs` writes the flag |
| `core/database/autoreconciler` | `recompute` *(new)*, `RecomputeReported` and `Rebuild` per D-1; `RecomputeTx` and `HandlesObservingSubject` deleted; package doc |
| `core/database/conclusionsummaries` | `dependents.go` deleted (D-9); `display_register.go` registers the check only |
| `core/database/graphcache` | D-6; `graphcachetest` D-7 |
| `core/database/effects` | `SummaryDepsWalked` deleted; `SummaryDeps` doc |
| `core/writes` | `Run` sets `SummaryDeps` only |
| `core/database/subjects` | D-8 |

Errors: none new.

### FFI

None — `summary_handles` keeps its meaning and shape.

### macOS

None — no view change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Budgets first, on the old code**
Files: `core/database/promotealign/perf_test.go`.
Do: add subtests "unpromote an event at the city" (promote three Births located at `lf.citySubject` as "locate an event at the city" does, then `promote.Unpromote` each through `writes.Call` with action `unpromote_subject`, fastest of 3) and "delete a location at the city" (three located Births; `subjects.Delete` of each location Subject, the return of `f.Locate`, through `writes.Call` with action `delete_subject`) with D-12's budgets. Log each figure.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign -run TestLargeSourceBudgets -v` logs both base figures. A new budget may fail here; record the figures in the commit message and continue.

**STEP-2 — Mark an association's end rows**
Files: `core/database/migrations/000062.sql`, `core/database/searchindex/inputs.go`, `core/database/searchindex/inputs_test.go` (its test `inputRow` gains `association bool`, `inputRowsOf` reads it, `hasRow` ignores it), `core/database/autoreconciler/autoreconciler_test.go` (`inputRows` appends `|0` or `|1` to each line). Skill: `add-catalog-migration`.
Do: D-4.
Check: `TestReprojectWritesInputs` asserts `dependent_is_association = 1` on the association's rows and `0` on the birth's; `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes (schema hash, `catalogmodel.TestPragmaHonesty`, `TestRebuildEqualsUpkeep_*`).

**STEP-3 — Find the rows a write moved**
Files: `core/database/searchindex/inputs.go`, `handles.go`, `export_test.go`, `inputs_test.go`. Skill: `add-catalog-query`.
Do: (1) D-10's three reads; (2) split `reprojectOn` out of `ReprojectHandles` (D-5); (3) `ReprojectChanged` per D-2, D-3, D-5.
Check: T-1 (through the `…ForTest` exports), T-2 pass; `TestReprojectHandlesBuildsOneGraph` extended: one `ReprojectChanged` call → `newGraph` called once.

**STEP-4 — Recompute calls it; S14-38's scenarios**
Files: `core/database/autoreconciler/autoreconciler.go`, `autoreconciler_test.go`, `core/database/conclusionsummaries/conclusionsummaries_test.go`, `core/database/graphcache/graphcachetest/graphcachetest.go` (`observe` calls `RecomputeReported`), `core/database/evidencelayer/evidencelayertest/evidencelayertest.go`.
Do: D-1 (`recompute`; `RecomputeReported` no longer calls `AssociationEndpoints` or `Dependents`; the `movedRoots` block goes; `Rebuild` per D-1; `RecomputeTx` deleted, its test callers moved). D-13's helper; T-3 – T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/... ./core/database/searchindex/... ./core/search/...` passes, every `TestRebuildEqualsUpkeep_*` included; `go test -tags fts5 ./core/database/promotealign -run TestLargeSourceBudgets` passes every budget that passed at STEP-1.

**STEP-5 — Displays drop from SummaryDeps only**
Files: `core/database/graphcache/commit.go`, `display.go`, `core/database/effects/registry.go`, `core/writes/run.go`, `core/database/conclusionsummaries/display_register.go`, `core/database/graphcache/graphcachetest/graphcachetest.go`, `core/database/conclusionsummaries/display_cache_test.go`.
Do: D-6, D-7; T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/conclusionsummaries/... ./core/database/persondetail/... ./core/database/canonlinks/... ./core/kinship/... ./core/catalogsession/... ./core/writes/...` passes.

**STEP-6 — Subject delete recomputes nothing itself**
Files: `core/database/subjects/subjects.go`, `core/database/autoreconciler/autoreconciler.go`. Skill: `add-catalog-delete`.
Do: D-8. Rewrite `Delete`'s doc: it releases facets and deletes the row; `writes.Run` recomputes what the released changes name.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjects/... ./core/database/deleteimpact/... ./core/database/autoreconciler/... ./api/ffi/...` passes, T-6 included.

**STEP-7 — Delete the inverse; move its tests**
Files: `core/database/conclusionsummaries/dependents.go` (deleted), `placechains.go`, `walks_test.go`, `placechains_test.go`, `display_cache_test.go`, `filedlinks_test.go`, `inputs_test.go`, `core/database/conclusionsummaries/conclusionsummariestest/inputs.go`, `core/database/promotealign/perf_test.go`.
Do: D-9, D-11.
Check: this command prints nothing (T-11):

```bash
rg -n 'func Dependents\b|AssociationEndpoints|linkEnds|SetDisplayDeps|forgetDisplays|SummaryDepsWalked|movedRoots|HandlesObservingSubject|func linkedHandles|RecomputeTx' core api
```

and `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes.

**STEP-8 — Budgets on the new code**
Files: `core/database/promotealign/perf_test.go`.
Do: run the budgets; record before (STEP-1) and after for both new subtests and for "locate an event at the city", "second part_of from the city" and "compose every summary".
Check: every `TestLargeSourceBudgets` budget passes.

**STEP-9 — Docs and version**
Files and edits:
- `.cursor/rules/catalog-queries.mdc` line 17 → "`canonicalgraph.Walk` stays for any walk that must see rows the cache has not loaded, including inside a write. A write finds the summary rows it moved in `catalog_search_inputs` (`searchindex.ReprojectChanged`)."
- `core/database/autoreconciler/autoreconciler.go`: package doc lines 15–18 (Run calls `RecomputeReported`, which recomputes and hands `searchindex.ReprojectChanged` the moved ids; Subject delete recomputes nothing itself) and the `Rebuild` doc.
- `core/database/effects/README.md` line 15: at the end of the paragraph that says `writes.Run` calls `Resolve` once, add "`Run` sets `Set.SummaryDeps` from the input index (`searchindex.ReprojectChanged`)."
- `core/writes/README.md` line 15: the sentence that begins "Subject delete still calls `RecomputeTx`" goes; "Summary dependents are still outside the registry" becomes "Summary dependents come from `catalog_search_inputs`."
- `docs/catalog-deletes.md` line 46 → "Handles come from the released changes. `effects.Handles` reads them and `Run` recomputes them; the rows that read a released link come from the input index."
- `docs/deployment-plan/spike-14/summary-inputs.md` § Invalidation: its opening names `autoreconciler.RecomputeReported` → `searchindex.ReprojectChanged` and the callers it serves (`Run`, the kit); line 91 is replaced whole: `Rebuild` recomputes and calls `ReprojectHandles` for every handle, and `ReprojectChanged` reprojects through `reprojectOn` on its own graph, with D-4's flag telling an association's end rows apart.
- `docs/deployment-plan/spike-14/deployment-plan.md`: the working rule "One input index, owned by `searchindex`" (see **Rules this plan bends**); S14-42 row `Landed #746`.
- `docs/ideas/structural-debt.md` row 19: the Subject delete clause goes, the seed `Upsert` part stays.
- `VERSION`, `core/version.go`, `MARKETING_VERSION` → `0.4.1`.
- `completed.md`, this plan.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-42-invalidate-from-the-input-index.md` reports no errors; `scripts/check-pr.sh` PASS.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/searchindex/inputs_test.go` `TestIndexReads` *(new)* | S14-40's `TestReprojectWritesInputs` world (James, a birth, filed participation A): `dependentsOf([(James,'','')])` holds the birth; `dependentsOf([(birth, participation, event)])` holds the birth; `dependentsOfHandles([A])` holds the birth and James; `oldEnds([A])` → A's two ends; `oldEnds([birth])` → none; an unknown id → empty | D-10 |
| T-2 | Go | same file `TestUnpromoteReportsTheDeletedLinksEnds` *(new)* | same world; `writes.Run` (`unpromote_subject`) of `promote.Unpromote(James)` → `SummaryDeps` holds the birth and James; no index row has James or A as dependent; the birth's document equals a rebuild's | D-2, D-3, D-5 |
| T-3 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestRebuildEqualsUpkeep_Scenarios` "unpromote james leaves the birth unnamed" *(new)* | James Robins + birth `1885-05`, both promoted, participation role subject filed; `f.run("unpromote_subject", Unpromote(James))` → `assertUpkeepEqualsRebuild`; the birth's document title is `Unspecified birth`; the returned set holds the birth | AC-1.1 |
| T-4 | Go | same, "unpromote the birth leaves james undated" *(new)* | same world; Unpromote the birth → equals rebuild; James's document has no `1885-05`; the returned set holds James | AC-1.2 |
| T-5 | Go | same, one scenario each *(new)*: unpromote the person end of a participation; the event end; a location's place; a part_of's whole; each end of a succeeded_by | → `assertUpkeepEqualsRebuild` after each | AC-1.3 |
| T-6 | Go | same, "delete a location at a place with other events" *(new)* | a place, three promoted events located there (`Records.Locate` returns each location Subject); `f.run("delete_subject", subjects.Delete(one location))` → equals rebuild; the returned set holds that event and the place and neither other event | AC-2.1, AC-3.2 |
| T-7 | Go | same, "preferred picture rebuilds the lead" *(new)* | a promoted person, two likenesses (`f.likenessIn`, `f.citeOn`); `f.run("set_preferred_likeness", SetPreferredLikeness(second))` → equals rebuild | AC-3.4 |
| T-8 | Go | `core/database/conclusionsummaries/display_cache_test.go` `TestKitMergeDropsReadersDisplays` *(new)* | `graphcachetest` kit: person P, birth E (`Birth`), a second birth E2; `Commit(P, E, E2)` (writes P's inputs, which hold E); list persons (displays stored); `Merge(E, E2)`, `Commit(E)` → `Display(P)` is gone | AC-3.5, D-7 |
| T-9 | Go | `display_cache_test.go` `TestDisplayCacheFollowsAPlaceThreeStepsUp` *(moved)* | the country rename through `writes.Run` → `SummaryDeps` holds the town, the birth and the person; the next People list shows the new name | AC-3.3, D-11 |
| T-10 | Go | `core/database/promotealign/perf_test.go` "unpromote an event at the city", "delete a location at the city" *(new)* | → under D-12's budgets | AC-2.2, D-12 |
| T-11 | Shell | STEP-7's check (the `rg` command written there) | → no output | AC-3.1 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `conclusionsummaries` (moved tests), `persondetail` and `canonlinks` (kit display drops), `core/search` handle tests, `TestRebuildEqualsUpkeep_SeededSequences`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump 0.4.1 (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #746 (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (no brief)
- [ ] Domain doc updated: docs/catalog-deletes.md, summary-inputs.md § Invalidation
- [ ] PR description: what shipped for S14-40 and S14-42, AC evidence, deviations, manual check, the STEP-1 / STEP-8 figures
```

## 10. Manual check

1. Open a project with a promoted person (James Robins) and a promoted birth dated 1885-05 with a filed participation (role subject). Search the omnibar for "Birth of James". Unpromote James. Search again: the birth shows as "Unspecified birth" and no hit reads "Birth of James Robins"; the Events list agrees.
2. With the birth located at a place that has other located events, open the place's page, then delete that birth's location from the birth's page. The place page reloads without the birth; the other events' hits are unchanged.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A row reads an input the index does not hold, so upkeep misses it | T-3 – T-7 compare with a rebuild; S14-40's oracle holds every recorded set |
| A test kit user relied on `forgetDisplays` walking from roots | D-7 sends the recompute's set; STEP-5's suites |
| A budget fails because a lookup reaches the city's events | D-2 keeps link keys to associations only; STEP-8 measures |

**Stop and report if:**

- A `TestRebuildEqualsUpkeep_*` scenario differs from a rebuild at STEP-4: name the write and the dependent. Do not add a lookup outside D-2.
- A moved test (D-11) needs an id the write does not report while S14-40's `TestInputsMatchIsolatedComposition` fails for that row.
- A `subjects.Delete` scenario needs `HandlesObservingSubject`'s set (a successful delete of a Subject an Observation points at).
- A D-12 budget fails at STEP-8, or a `TestLargeSourceBudgets` budget that passed at STEP-1 fails at STEP-4 or later.
- A test outside `graphcachetest` calls `OnCommit` with `Handles` and relies on dependents' displays being dropped without `SummaryDeps`.

## 12. Open questions

None.

## 13. PR

**Title:** S14-40 and S14-42 — Build the input index and invalidate from it
**Body:** S14-40's body, plus: objective in one sentence; what shipped against each US; AC evidence; deviations; manual check; the STEP-1 / STEP-8 figures; `Plan: docs/deployment-plan/spike-14/prs/S14-42-invalidate-from-the-input-index.md`.
