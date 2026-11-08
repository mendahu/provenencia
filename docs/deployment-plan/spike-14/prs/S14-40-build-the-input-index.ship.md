# Ship log — S14-40 Build the input index

**Plan:** [`S14-40-build-the-input-index.md`](S14-40-build-the-input-index.md) · **Branch:** `s14-40-build-the-input-index` · **PR:** #746
**State:** ready

## Round 0 — implement
Base `main` @ `78f26812` (plan was cut against `26ada55`). Local head `f6fc77d6`, unpushed, plus uncommitted T-9/T-10 in `core/database/promotealign/perf_test.go` and the single-graph isolator in `conclusionsummaries/conclusionsummariestest/inputs.go`. Gate not run. STEP-1 through STEP-6 committed; STEP-7 stopped; STEP-8 and STEP-9 not started.

**Stopped on.** In `TestLargeSourceBudgets`, the "inputs cover Dependents" subtest (T-10) fails with 9,987 misses. Every miss is a person that `Dependents([place])` returns, but whose recorded inputs do not include that place.

Orchestrator check:
- `conclusionsummaries/dependents.go:191` `sqlBirthOrDeath` matches an event on its kept `event_type` term (birth or death) only. It does not check whether the event has an accepted member.
- The person row's `lifePicks` keeps only events where `Node.Canonical` is true (`graphcache/graph.go:170` requires `len(AcceptedMembers()) > 0`).
- `bulkCanon` in `promotealign/bench_test.go` builds the fixture's events with no accepted `identity_claims`. The implementer spot-checked two events (`accepted=0`), not all 9,987.

So `Dependents` reports more dependents than the composers actually read. The index looks more precise, not short. T-2, T-3 and T-9 (the isolated-composition oracle on all 8,117 persons, 10,070 events and 52 places) pass.

Deviations so far:
1. Base drift: perf helpers moved to `databasetest.Fastest`, `Since` and `WithinBudget`.
2. New package `conclusionsummaries/conclusionsummariestest`, shared by T-2/T-3 and T-9/T-10.
3. `InputScope.Node` and `Follow` call the exported graph methods, so the recorder also hears reads made through `NoInputs()`.
4. `writeInputs` uses `INSERT OR IGNORE` in 200-row statements.
5. `lifeOwners` takes an `*InputScope`.

Provisional timings (noisy machine):
- First part_of from the city: 2.25 s → 1.22 s.
- Compose all rows: 146 → 188 ms (budget 500 ms).

**To user.** How should D-13 treat birth and death events that have no accepted member?

**User (2026-11-08):** "Exclude them in D-13". The plan's D-13, AC-1.3, T-10 and Stop condition 2 are amended. A person missing for X is excused when every birth or death subject event of theirs that `Dependents([X])` returns has no accepted member. Excused persons are counted in the log. Any other miss is still a stop. Committed with this log as "Amend S14-40's D-13 to excuse non-canonical birth and death events."

## Round 0 — resume
Head `251cd6c0`; the PR is #746 (draft). The orchestrator ran the gate on `2a50791a` and again on `251cd6c0`: PASS both times.

T-10 now passes. It excuses 9,987 persons, the same number as the earlier misses. Extras: event 100, person 100, place 52. Final figures (single runs on a noisy machine):
- T-11, first part_of from the city: 2.25 s → 1.33 s (×0.59).
- T-8, compose every row: 146 → 203 ms (budget 500 ms).

New deviation 6: the test isolator reuses one graph for all rows. STEP-8 and STEP-9 are done; the plan is marked Landed in #746.

## Round 1 — review
Base `78f26812`, head `251cd6c0`. Reviewers: CON, MNT, SEC+PRF. Security is routed for the migration and `catalogmodel`; performance for `core/database/**`, `core/search` and the graph cache. Skipped: UX (nothing under `macos/App/**`, no `Localizable.xcstrings`, no error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC (+PRF) | blocker | Every row copies its place closure; a dense part_of mesh makes the key count rows × links, and the rebuild on open writes all of it | Fix now → F-1 | The reviewer's probe measured it: a complete DAG of 40 places (780 part_of links) with 2,000 events gives 1.73M keys and 698 ms. Only a crafted catalog builds a mesh like that; the large fixture's max is 17 keys per row. Working rule (user, S13-26): one high total cap, counted before the work, and over the cap the result counts nothing |
| CON-1 | CON | major | In a batch, a row shows part_of ancestors deeper than 8 that its inputs lack | Fix now → F-2 | The orchestrator confirmed `placechains.go` `parentChain` and `isAncestor` walk `partFrom` with no bound, while `closure` stops at `summaryChainDepth`. The reviewer reproduced it with an 11-place chain. The doc on `summaryChainDepth` (`dependents.go:16`) says it is how far a place row's parent chain climbs, so bounding the chain makes the code match that doc. Only chains deeper than 8 change, and there the batch display already disagreed with the isolated one |
| MNT-1 | MNT | major | A fifth copy of the bridge-origin default, already drifted (trim / no trim) | Fix now → F-3 | Same-pass cleanup. `rg -n 'origin == ""' core` finds 5 sites; this PR adds `graphcache/inputs.go:155` |
| MNT-2 | MNT | major | "Places of these events, then closure" and the birth-plus-death merge are copied 3× in the composers | Fix now → F-4 | Same-pass cleanup. `rg -n 'append\(append\(\[\]lifePick\(nil\)' compose.go` returns 3 lines |
| PRF-1 | PRF (+MNT outside-lens) | minor | List and ByIDs paths record inputs, then discard them | Fix now → F-5 | The diff discards the map with `_`. T-8 rose from 146 to 203 ms. Keeps D-3's single path by passing `NoInputs()` scopes |
| CON-3 | CON | minor | AC-1.2 asserts only some of the keys it lists | Fix now → F-6 | A test this PR adds |
| MNT-3 | MNT | minor | The D-13 excuse re-derives `Node.Canonical` and `sqlBirthOrDeath` in raw SQL | Fix now → F-7 | A test helper this PR adds |
| MNT-4 | MNT | minor | `InputKey` equality and map key are re-implemented by 5 consumers | Fix now → F-8 | Same-pass cleanup. `rg -n 'func (sameKeys\|holds\|hasRow)' core` |
| CON-2 | CON | minor | Large-fixture checks never reach a person's place inputs; `summary-inputs.md` gives the wrong reason | Fix now → F-9 | Fixture events have `accepted=0`. Doc fix plus a Landed note. The `bulkCanon` fixture is not changed, because every existing budget runs on it |

Nits (logged): `CON-4 · graphcache/inputs_test.go:31 · TestInputScope is t.Run blocks, not a table` · `MNT-5 · graphcache/inputs.go:221 · NoInputs doc says "reads and records nothing"; it reads` · `MNT-6 · searchindex/inputs.go:18 · inputRowsPerInsert cites SQLite's variable limit; 800 binds is far below it` · `MNT-7 · searchindex/inputs.go:30-37 · parallel rows/deps slices could be one slice` · `MNT-8 · conclusionsummariestest/inputs.go:185 · hex8 hand-rolls hex.EncodeToString` · `MNT-9 · conclusionsummariestest/inputs.go:23 · exported Querier alias has no caller that needs it`

### Fix plan 1

Order: the blocker first, then the majors, then the minors.

**F-1 — One high total cap on recorded input keys per composition call** (from SEC-1)

Files:
- `core/database/conclusionsummaries/placechains.go`
- `core/database/conclusionsummaries/compose.go`
- `core/database/conclusionsummaries/inputs_test.go`
- `core/database/promotealign/perf_test.go` (log the measurement)

Failing test first: `conclusionsummaries` `TestInputsCapCountsNothing`. Build a complete part_of DAG over places (place i part_of every j<i) and put enough births at the deepest place for the total to pass the cap. Use a test seam that lowers the cap, so the world stays small. `RowsForSearch` → `s.Inputs(id)` is empty for every row of that call, and the rows and documents are unchanged.

Do:
- Add `maxInputKeysPerCall` (a const beside `summaryChainDepth`).
- One `RowsForSearch` call counts the keys it records across all its rows.
- Before each `in.Add(pg.closure(...)...)`, count the closure's size without copying it; give `closure` a size-only form or a precomputed per-place count. If the running total plus that size would pass the cap, the call stops recording and returns no inputs for any row. That is the "counts nothing" result, so `writeInputs` writes nothing for that batch.
- Do not truncate per row.
- Set the number from a measurement: log the largest per-call total in `TestLargeSourceBudgets` T-9 (500-row batches on the large fixture), then set the cap to at least 50× that and at least 250,000.
- Write the measured figure and the cap in the const's doc and in the ship log.
- Over the cap, the index for that batch is empty. That is an invalidation miss S14-42 inherits only for a crafted catalog. Say so in the const's doc.

Check: the test passes; T-9 and T-10 still pass, with no call over the cap on the large fixture; `scripts/check-pr.sh` PASS.

**F-2 — Bound the parent chain to `summaryChainDepth`** (from CON-1)

Files: `core/database/conclusionsummaries/placechains.go`, `core/database/conclusionsummaries/inputs_test.go`.

Failing test first: `TestInputsHoldEveryShownParent`. Use an 11-place part_of chain P0…P10, event A at P0 and event B at P3, both in one `RowsForSearch`. Every parent name A's place shows must be a place whose values key A's inputs hold, and A's parents must equal its isolated composition. The reviewer's code is in CON-1.

Do: `parentChain` and `isAncestor` (and so `FoldLocations` and `ParentChain`) stop after `summaryChainDepth` levels above the queried place, the same depth `closure` and the build use. Add a depth counter; do not add a second constant. Update the `summaryChainDepth` doc to name these readers.

Check: the test passes; existing place-chain and fold tests pass unchanged; `scripts/check-pr.sh` PASS.

**F-3 — One bridge-origin default in `connectrules`** (from MNT-1)

Files:
- `core/connectrules/connectrules.go`
- `core/hops/hops.go`
- `core/database/graphcache/inputs.go`
- `core/database/graphcache/load.go` (`bridgePredicates`)
- `core/database/conclusionsummaries/dependents.go`

Failing test first: `connectrules` `TestBridgeOriginKey`, a table: `""` → `provenencia`, `" plugin "` → `plugin`, `"x"` → `x`.

Do:
- Add `func (b Bridge) OriginKey() string` (trim, then default to `provenencia`).
- Use it at every site that defaults or reads `b.Origin`: `hops.go:42`, `connectrules.go:121` and `:342`, `dependents.go:118`, `graphcache/inputs.go:155`, and `load.go` `bridgePredicates`.

Check:
- `rg -n 'origin == ""|Origin == ""' core` prints only `OriginKey`'s own body.
- `rg -n '\.Origin\b' core --type go` prints only `OriginKey` and struct literals or registrations.
- `scripts/check-pr.sh` PASS.

**F-4 — One "places of events, then closure" helper and one life-events merge** (from MNT-2)

Files: `core/database/conclusionsummaries/placechains.go`, `core/database/conclusionsummaries/compose.go`.

No new test: this is a refactor. T-2, T-3, T-9 and T-10 are the check.

Do:
- In `placechains.go`, add `func (pg *placeGraph) addChains(in *graphcache.InputScope, eventIDs [][]byte) error`. It follows `PlacesOfEvent` through `in` and adds `pg.closure`, going through F-1's cap.
- Add `func lifeEvents(births, deaths []lifePick) [][]byte`.
- Use both in `composePersons`, `composeEvents` and `PersonLifeOwners`.
- Have `eventPlaceGraph` reuse the same "places of events" step through `g.NoInputs()`.
- `composeEvents` reuses the places `oneEvent` / `eventPlaces` already followed through the row's scope, rather than following `PlacesOfEvent` a second time.

Check:
- `rg -n 'append\(append\(\[\]lifePick\(nil\)' core/database/conclusionsummaries` prints only `lifeEvents`.
- `rg -n 'hops.PlacesOfEvent\)' core/database/conclusionsummaries/compose.go` prints no per-row follow-and-closure loop outside the helper.
- `scripts/check-pr.sh` PASS.

**F-5 — List and ByIDs paths record nothing** (from PRF-1)

Files: `core/database/conclusionsummaries/compose.go`.

Failing test first: `TestListsRecordNoInputs`. Use a seam that reports whether any recording scope was created, or count `Inputs` scopes through an export_test hook. `ListPersons`, `ListEvents`, `ListPlaces` and the `*ByIDs` paths create none; `RowsForSearch` does.

Do:
- `persons`, `events` and `places` take a `record bool`. When it is false, rows get `g.NoInputs()` and the code skips `closure`, `addChains` and `Keys()`.
- Only `RowsForSearch` passes true.
- This keeps D-3's one code path.

Check: the test passes; T-8 logged again in the ship log; `scripts/check-pr.sh` PASS.

**F-6 — AC-1.2 asserts every key it lists** (from CON-3)

Files: `core/database/conclusionsummaries/inputs_test.go`.

Do: in `TestInputsMatchIsolatedComposition`'s James table, look up the birth's location association and the Town→City, City→Region and Region→Country part_of association ids with `canonicalgraph.Walk`. Add `holds` rows for:
- each association's `(A, values)`;
- the City and Region `(place_relationship, from)` ends.

Check: the test passes; `scripts/check-pr.sh` PASS.

**F-7 — The D-13 excuse reads through the graph** (from MNT-3)

Files: `core/database/conclusionsummaries/conclusionsummariestest/inputs.go`.

Do: `excusable` reads each event through a `graphcache.Over(q)` graph. It uses `n.Canonical(eventTypeID)` and the kept `event_type` term from the node, rather than its two raw SQL strings.

Check: T-3 passes; T-10 passes with the same excused count (9,987); `scripts/check-pr.sh` PASS.

**F-8 — `InputKey.Equal` and `InputKey.Key`** (from MNT-4)

Files:
- `core/database/graphcache/inputs.go`
- `core/database/graphcache/inputs_test.go`
- `core/database/conclusionsummaries/inputs_test.go`
- `core/database/conclusionsummaries/conclusionsummariestest/inputs.go`
- `core/database/searchindex/inputs_test.go`

Do:
- Export `func (k InputKey) Equal(o InputKey) bool` and `func (k InputKey) Key() string`. `Key` replaces `mapKey`.
- Move `sameKeys` (both copies), `holds`, `hasRow` and the isolator's hand-built key string onto `Key`/`Equal` with `slices.EqualFunc` / `slices.ContainsFunc`.

Check:
- `rg -n 'func (sameKeys|holds|hasRow)' core` prints only thin wrappers over `slices.*Func`, or nothing.
- `rg -n '"\\\\x00"' core/database` prints only `Key`.
- `scripts/check-pr.sh` PASS.

**F-9 — Say why the large fixture's person rows hold no place inputs** (from CON-2)

Files: `docs/deployment-plan/spike-14/summary-inputs.md`, and the plan's Landed note.

Do:
- Replace "A person's keys are fewer than planned because its place graph now comes from its birth and death events only" with the real reason: the fixture's `bulkCanon` events have no accepted member, so no person has a life pick, and person place inputs are proven by T-2 and T-3 only.
- Add the same line to the plan's **Landed** note.

Check: `scripts/check-plan-docs.py` on the plan passes.

### Fix round 1 — figures

- F-1: on the large fixture the largest 500-row call counts 1,000 closure keys and holds 3,500 keys in all (events; persons 3,044; places 51); a whole-fixture call counts 16,102. `maxInputKeysPerCall` is 250,000 (the plan's floor; it is above 50 times 3,500).
- F-5: T-8, compose every person and event row, 184 ms (was 203 ms; budget 500 ms), single run on a noisy machine.
- T-10 still excuses 9,987 persons; extras: event 100, person 100, place 52.

## Round 1 — fix
Head `4edc866e`. The orchestrator ran the gate: PASS. F-1 through F-9 are done; none was dropped as Not reproduced.

Deviations:
- F-2 stops `parentChain` at 7 named ancestors (`summaryChainDepth-1`), so the batch matches the isolated read. `isAncestor` keeps 8 levels, so its level-8 edge reads P8's period.
- F-5 uses a nil-able `*inputBudget` instead of `record bool`.
- F-1's cap counts closure keys before dedup; the large fixture's largest call counts 3,500. The cap is 250,000.
- New test seams: `Graph.SetScopeCounterForTest` and `SetInputKeyCapForTest`.

## Round 2 — review
Scope `251cd6c0...4edc866e`. Reviewers: CON (CON-1 major), MNT (MNT-1 and MNT-2 majors), SEC+PRF (SEC-1 blocker).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major (r1) | Shown ancestors past depth 8 | Resolved | `TestInputsHoldEveryShownParent` passes at head. With 88f099fb's `placechains.go` hunk reversed, it fails |
| MNT-1 | MNT | major (r1) | Bridge-origin default copies | Resolved | One `Bridge.OriginKey`; every bridge reader calls it |
| MNT-2 | MNT | major (r1) | Places-of-events and life-merge copies | Resolved | `lifeEvents`, `placesOfEvents`, `addChains`; every closure goes through `inputBudget.addClosure` |
| SEC-1 | SEC+PRF | blocker (r1) | Unbounded closure keys | Resolved | One per-call cap, sized before copying. Over the cap, the call records nothing. `TestInputsCapCountsNothing` |
| PRF-1 | PRF | minor (r1) | Lists record inputs | Resolved (orchestrator) | List and ByIDs paths pass a nil `*inputBudget` (`compose.go:232`, `:368`, `:398`, `:415`, `:476`, `:493`). `TestListsRecordNoInputs` |
| CON-3 | CON | minor (r1) | AC-1.2 partial asserts | Resolved (orchestrator) | The James table asserts the location and part_of associations and the City and Region `place_relationship` ends |
| MNT-3 | MNT | minor (r1) | Excuse re-derives Canonical | Resolved (orchestrator) | `rg 'identity_claims\|sqlBirthOrDeath' conclusionsummariestest/` finds nothing |
| MNT-4 | MNT | minor (r1) | `InputKey` equality copies | Resolved (orchestrator) | `sameKeys`, `holds` and `hasRow` are thin wrappers over `slices.*Func` and `InputKey.Equal` |
| CON-2 | CON | minor (r1) | Fixture reason wrong | Resolved (orchestrator) | `summary-inputs.md:114` gives the `bulkCanon` reason. One clause is repeated (nit) |

No new blockers in round 2.

Nits (logged): `CON-5 · conclusionsummaries/inputs_test.go:240 · TestInputsCapCountsNothing checks event rows only for "rows unchanged"` · `summary-inputs.md:114 · "the fixture's bulkCanon events have no accepted claims" said twice`

## Result
Gate PASS on `d55fb493`. `origin/main` is unchanged since the branch was cut (`78f26812`). CI: see PR #746. Open: none. Follow-ups: none. To user: D-13 (answered 2026-11-08, "Exclude them in D-13").
