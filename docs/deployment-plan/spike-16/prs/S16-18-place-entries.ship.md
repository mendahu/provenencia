# Ship log — S16-18 Place entries

**Plan:** [`S16-18-place-entries.md`](S16-18-place-entries.md) · **Branch:** `s16-18-place-entries` · **PR:** #817
**State:** fixing 1

## Round 0 — implement
Head `3b58acf2b` (7 local commits on `b9a881b65`, not pushed, no PR). Gate not run.

**Stopped on** §11: "`BenchmarkPlaceMesh` cold, after the hub commit, or after a location at the hub is more than 1.25 × slower after than before."

`Jakes-M1-Macbook-Pro.local`, `-benchtime 1x`, ms:

| BenchmarkPlaceMesh | base | after 1 | after 2 | after 3 |
| --- | --- | --- | --- | --- |
| list-cold | 1522 | 1865 (1.23×) | 2366 (1.55×) | 1822 (1.20×) |
| list-warm | 230.9 | 100.7 | 110.7 | 106.2 |
| page-cold | 2026 | 2031 | 2008 | 1898 |
| list-after-hub-commit | 142.4 | 560.2 (3.9×) | 667.9 (4.7×) | 667.8 (4.7×) |
| list-after-location | 137.7 | 88.6 | 98.4 | 101.3 |

Cause (implementer): renaming the 20,000-part hub drops its entry (D-5); the refill (D-2, D-3) reads the hub's 20,000 association value sets through `NodesOnBridges`. At base the hub reload kept the associations' nodes cached. The plan's Performance section anticipated this read ("its fill reads 20,000 association value sets") but not against the 1.25× stop.

Other stops hold: `BenchmarkListPlaces` cold +7.7 ms at 4,000, +7.6 ms at 16,000 (limit 50 ms); warm and write within; `placeEntries-B` ≈ 124 KB at both sizes (limit 5 MB).

Deviations reported: `eventPlaceGraph` batches `g.Nodes(placeIDs)` before `buildPlaceGraph` (query-count tests); `buildPlaceGraph`'s `rec` unused; self-loop filed in both lists; extra hook `FillPlaceEntriesForTest`; `byID` locals in `composePersons` unrelated.

Taken to the user 2026-11-13.

**User, 2026-11-13:** keep the 20,000-part hub (the spike's bar 2 size); replace the 1.25× stop for "after the hub commit" with an absolute 1 s ceiling, since the map's Scale asks crafted data only to stay bounded. Cold and after-location keep 1.25×, now on the median of three runs. Plan §11 and Performance "Hub entry" amended. Implementer resumes at STEP-9.

## Round 0 — resume
Head `58af54ac0`. Gate not run. STEP-9 measured; STEP-10 not done.

**Stopped on** amended §11 cold row. `list-cold` ms: base 1522 / 1362 / 1328 (median 1362); after 1865 / 2366 / 1822 (median 1865, 1.37×) and 1706 / 2022 / 1974 (median 1974, 1.45×). Cause: a cold read of the hub's entry reads its 20,000 association value sets through uncached `NodesOnBridges` (D-2, D-3), on top of the whole-node load.

Other rows: after-hub-commit 560–668 ms (under 1 s); after-location median ≈ 98 vs 210 base (pass); page-cold unchanged; warm faster. `BenchmarkListPlaces` 16,000 cold +7.6 ms (limit 50), warm faster. `placeEntries-B` ≈ 124 KB. `nodes-B` unchanged. `TestLargeSourceBudgets` passes.

Taken to the user 2026-11-13.

**User, 2026-11-13:** the cold crafted-hub row also gets an absolute ceiling, 3 s on any of three runs. Plan §11 amended. Implementer resumes at STEP-10.

## Round 0 — resume 2
Head `018ef0703`, pushed; draft #817. Gate PASS (orchestrator, against `b9a881b65`). Every §11 stop holds under the two amendments. `TestLargeSourceBudgets` passes. Manual check: none (Go-only).

Report deviations: as above, plus the Landed note. The report's deviation 6 says cold compares medians; the plan's §11 gives cold the 3 s ceiling (any of three runs) and keeps medians for after-location only.

## Round 1 — review
Reviewers: CON, MNT, SEC+PRF (`core/database/**`, graph and cache code; fills loop over catalog-stored `part_of` links of unbounded count). Skipped: UX (no `macos/App/**`, strings, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1, PRF-1 | SEC, PRF | major | The entry fill decodes and reads every link's association on all four hops, though a chain reads only `partFrom`: one bad date on a sibling or succession link fails `ListPlaces` and every event/person summary through that place (base: only that place's page); writes under an N-part whole read N association value sets (2,500 parts 27→45 ms, 10,000 148→218 ms) | To user → Fix now → F-1 (user, 2026-11-13: split the entry) | verified: `fillPlaceEntries` collects `assocIDs` over all four `byHop` and returns `orientPlaceLink`'s error for each; reviewer's probe errors at head, nil at base. Fix (split the chain read from the place-page read) contradicts D-1/D-3 (one entry with four lists, one fill per frontier) |
| MNT-1 | MNT | major | Two ways to drop keyed entries by dependency: new `Inputs` and `canonlinks.kinship`'s `deps` Apply | Follow-up | Plan's Out line; this PR adds the debt row in `structural-debt.md` (already homed). Same-pass rule does not apply: `canonlinks` is untouched by this PR. Reviewer confidence medium on whether kinship's reads all fit `FollowKeys` |
| MNT-2 | MNT | minor | `buildPlaceGraph`'s dead `rec` parameter | Fix now → F-2 | compose.go:1295 `_ *inputBudget` |
| MNT-3 | MNT | minor | Private `sortedKeys` dedupe+sort beside graphcache's | Fix now → F-3, same-pass | this PR adds a copy |
| MNT-4 | MNT | minor | `placeInputs` builds fake `Link`s for `FollowKeys` | Fix now → F-1 (closed by the split) | placeentries.go:203-207 |
| MNT-5, CON-3 | MNT, CON | minor | `cache-architecture.md:47` names deleted `pg.inputs` | Fix now → F-4 | `rg pg\.inputs` |
| CON-1 | CON | minor | Self-loop `part_of` untested | Fix now → F-5 | no test; Verify rebuilds through the same fill |
| CON-2 | CON | minor | `TestKeyedInputLinks` counts keys instead of naming them | Fix now → F-6 | `held() != 1` |
| MNT (outside lens → PRF) | PRF | — | `placeStep` fills full entries for every neighbour on the place page | Rejected | page-cold unchanged at base (≈2.0 s base vs ≈1.9–2.0 s head, three runs each) |
| CON (outside lens) | CON | — | Chained merge of a self-looped place leaves a stale cached node (Verify "links: count 2, catalog 0"); reproduces at base | Follow-up | pre-existing; place entries never read cached nodes; probe `scratchpad/ov/node_probe_test.go` |

Nits (logged): `MNT-6 · placeentries.go:64 · use the bridge constant, not "place_relationship"` · `MNT-7 · placeentries.go:118 · i < 2 ties part/succession to byHop order` · `MNT-8 · export_test.go:54 · HeldPlaceEntryForTest repeats PlaceEntryForTest` · `CON-4 · cache_test.go TestKeyedInputs · rename "a key's own handle drops it"`

Landed note correction for the fix round: deviation 6 says cold compares medians; §11 gives cold the 3 s ceiling.

**User, 2026-11-13:** split the place entry (SEC-1, PRF-1). Plan D-1, D-3, D-5, D-8, D-12, D-18 amended in place ("Amended (user, 2026-11-13, round 1 …)").

### Fix plan 1

**F-1 — Split the place entry into a chain entry and a page entry** (from SEC-1, PRF-1, MNT-4)
Files: `core/database/conclusionsummaries/placeentries.go`, `compose.go` (`placeStep`), `placechains.go` (`placeGraph`, `ImmediateParts`, `succNeighbors`, `AttachPlaceRelationships`), `export_test.go`, `placeentries_test.go`, `core/database/promotealign/bench_test.go` (`BenchmarkCacheSizes`), `core/catalogsession/cache_doc_test.go` if it lists entries, `docs/catalog-caches.md`.
Failing tests first, in `TestPlaceEntries` (table cases):
- "bad succession date": A `succeeded_by` B with a start date; set that association's `auto_reconciler_values.value_date` to `x'ff'` (raw SQL on `f.C`), `Graph().Drop()` → `ListPlaces` returns no error and A's row reads as before; `AttachPlaceRelationships` of A returns the decode error.
- "bad part date": the same on a child's `part_of` link of City (Town `part_of` City, dates on the link) → `ParentsAtDate` of City and `ListPlaces` succeed; Town's own chain (which climbs that link) errors, as at base.
- "chain fills no page": after `Drop` and `ListPlaces`, the `conclusionsummaries.placePages` key count is 0.
Do: per amended D-1, D-3, D-5, D-8, D-12, D-18. `placeEntry` keeps `found`, `shell`, `period`, `partFrom`, `inputs`; its fill follows only `ParentsOfPlace` and reads only those links' associations. New `placePages` keyed entry (`init` registration copying `placeEntries`': `DerivedFrom` vocabulary, `Cap: maxPlaceEntries`, `Check` by rebuild, `Inputs` returning its stored `inputs`) whose fill follows `PartsOfPlace`, `SuccessorsOfPlace`, `PredecessorsOfPlace` and builds `inputs` from `NodeKey(id)` plus `FollowKeys` over the real `Follow` results. Delete `placeInputs`. `placeGraph` gains `pages`; only `placeStep` fills it, for its own id. Rename or add the test hooks as needed (`PlaceEntryForTest` returns the chain entry's inputs; add `PlacePageForTest(g, id) (held bool, inputs []graphcache.InputKey)`); T-8 (b) checks the page entry per amended D-18; T-11's link moves also hold a page for City and assert it drops. Add a `placePages` fill to `BenchmarkCacheSizes` and a row to `docs/catalog-caches.md`.
Add sub-benchmark `write-under-hub` to `BenchmarkPlaceMesh`: `RowsForSearch(graphcache.Over(db), Want{Events: [an event located at the first part], Places: [first part]})`. Measure it at head of this fix and at base (`1dd02710d` with only this sub-benchmark added, in a `git worktree add` under the scratchpad, removed after); record three runs each.
Check: the three new cases pass; `go test ./core/database/...` and `./api/...` pass, also with `PROVENENCIA_GRAPH_VERIFY=1`; `BenchmarkPlaceMesh` (all six sub-benchmarks, three runs) and `BenchmarkCacheSizes` re-run and every §11 stop holds; `write-under-hub` median after ≤ 1.25× base median; scripts/check-pr.sh PASS. Report all figures.

**F-2 — Drop `buildPlaceGraph`'s and `eventPlaceGraph`'s dead `rec` parameter** (from MNT-2)
Files: `compose.go`, `placechains.go` call sites (compose.go ~855, ~1000, ~1278; placechains.go ~412).
Do: remove the parameter and whatever builds a budget only to pass it.
Check: `rg -n "buildPlaceGraph\(|eventPlaceGraph\(" core` shows no budget argument; tests pass.

**F-3 — One dedupe-and-sort for InputKeys** (from MNT-3; same-pass cleanup)
Files: `core/database/graphcache/inputs.go`, `conclusionsummaries/placeentries.go`, `core/database/searchindex/inputs.go` (`dependentsOf`), `core/database/conclusionsummaries/conclusionsummariestest/inputs.go`.
Do: export `graphcache.UniqueInputKeys(keys []InputKey) []InputKey` (dedupe by `Key()`, then `SortInputKeys`; returns a new slice, does not alias the input). Delete `sortedKeys`; use it there, in the isolator, and in `dependentsOf` where its order is acceptable to callers (if `dependentsOf` must stay unsorted, leave it and say so in the report).
Check: `rg -n "\.Key\(\)\]|seen\[.*Key\(\)" core --type go` prints only `InputScope.add`'s map and `UniqueInputKeys` (and `dependentsOf` if left, with the reason); tests pass.

**F-4 — Design doc names the entry's inputs** (from MNT-5, CON-3)
Files: `docs/deployment-plan/spike-16/cache-architecture.md:47`.
Do: "(`pg.inputs`, `inputBudget.addClosure`)" → "(each place entry's `inputs`, which `inputBudget.addClosure` hands to the row)".
Check: `rg -n "pg\.inputs" docs .agents .cursor core` prints nothing.

**F-5 — Self-loop test** (from CON-1)
Files: `placeentries_test.go`.
Do: `TestPlaceEntries` case "self-loop": `placeRel(town, town, part_of)` → `ParentsAtDate(town)` reads as at base (write the expected value from base behaviour: run it at `b9a881b65` if unsure); `AttachPlaceRelationships` lists the link once in `PartOf` and once in `Contains`; `PlaceEntryForTest` and `PlacePageForTest` inputs equal the scope reads.
Check: passes, also under Verify.

**F-6 — Name the kept keys in `TestKeyedInputLinks`** (from CON-2)
Files: `core/database/graphcache/cache_test.go`.
Do: in the "removed location" and "merge" steps, assert with `kv.Peek` which key is held and which is not, as `TestKeyedInputs`' `held` helper does.
Check: passes.

Also: correct the plan's Landed note (deviation 6: cold has a 3 s ceiling, after-location compares medians) and add the round-1 amendments to it; update the Performance "Hub entry" line with the new figures.
