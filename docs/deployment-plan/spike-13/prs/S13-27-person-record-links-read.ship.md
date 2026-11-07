# Ship log — S13-27 Person record links read

**Plan:** [`S13-27-person-record-links-read.md`](S13-27-person-record-links-read.md) · **Branch:** `s13-27-plan` (the plan PR's branch, by the user's choice, 2026-11-07; the plan header names `s13-27-person-record-links-read`) · **PR:** #743
**State:** fixing 1

## Round 0 — implement
Head `fa3dbb91` (8 commits on `cf7a27ca`). Gate PASS (implementer, then orchestrator on the pushed head). Report deviations: `pairKey{a, b int}` instead of `[2]int` (the numbers guard); `readWork` wraps an internal `readPlan` that also returns the shared pairs; `Pairs` returns nil with no shared person; T-8 files its registers with `evidencelayertest.Open` and `promotetest.Save` directly; Landed docs written for #743.

Measured: 50-graph read work 31,850 → `MaxReadWork` 640,000. `BenchmarkPairsAtCap` 12.8 s/op, m = 7 graphs (names axis, M1; under the 60 s stop). `TestRecordLinksBudgets`: cold 50.5 ms (1.5 s), repeat 36.7 ms (300 ms), 65-graph repeat 81.0 ms (1.5 s).

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Base `cf7a27ca`, head `fa3dbb91`. UX: the diff touches `macos/App/**` (store, key, registry). SEC+PRF: `api/ffi/**`, `core/recordlinks`, `core/database/**`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC+PRF | blocker | `MaxReadWork` charges no work for links to entities the other graph lacks; a crafted person in 64 graphs holds the read 57.7 s at 2,016 of 640,000 counted work | Fix now → F-1 | verified by the reviewer's measured probe (`go test -overlay`, 16 / 32 / 64 graphs × 5,000 – 20,000 unshared links: 0.41 s → 57.7 s, ReadWork = pairs); `costOf`, `against` and `links` each walk every link of a member, `pairWork` charges only products |
| PRF-1, MNT-1 | SEC+PRF, MNT | minor | Each pair's `MaxPairWork` result is computed twice (`readPlan`, then `linesOf` → `overWork`) | Fix now → F-2 | read: `readPlan` drops `w`; `linesOf` recomputes. Code this PR adds |
| MNT-2 | MNT | minor | `readWork` has no production caller | Fix now → F-3 | `rg readWork core/recordlinks`: only `export_test.go` |
| CON-1 | CON | minor | The 65-graph equality compares graphs only by count | Fix now → F-4 | read the "past the cache" subtest |
| CON-2 | CON | minor | `TestPairsInvariants` never reaches either cap | Fix now → F-5 | uses the shipped config; the reviewer's random-cap probe passed 3,000 rounds |
| CON-3 | CON | minor | `TestRecordLinks` subtests share and mutate one fixture | Fix now → F-6 | AC-1.3 asserts `len(graphs) != 6` from earlier rows; AC-1.6 writes `merged_into_id` |
| UX | UX | — | No findings | — | no view, copy or error code |

Nits (logged): `MNT-3 · api/ffi/handlers/person_detail_test.go · second copy of AC-1.1's register builder beside persondetail's registerOf` · `MNT-4 · CatalogQueryRegistry.swift:583,629 · personGraphsSection callers match the key case twice` · `CON-4 · TestGetPersonRecordLinks · add a malformed entity_id row`

### Fix plan 1

**F-1 — Charge the read for every link and property a compare walks** (from SEC-1)
Files: `core/recordlinks/recordlinks.go`, `core/recordlinks/recordlinks_test.go`, `core/matchregistry/recordlinks.go`, `docs/matching-model.md` (line-33 paragraph, `MaxPairWork` and `MaxReadWork` rows), `.agents/skills/add-match-rule/SKILL.md` if it describes what work counts. Skill: `add-match-rule`.
Failing test first: `TestLinesWorkCap` row "unshared links" and `TestPairsReadCap` row "unshared links": a shared person F who, in each graph, is a "witness" on k events only that graph holds, each event on its own entity (the reviewer's probe at `/private/tmp/claude-501/-Users-jakerobins--claude-spares-spare-exgiSE/1df29678-6114-41d4-97b5-a1bf452a254b/scratchpad/probe/probe_test.go` builds it; copy its builder into `recordlinks_test.go` as a named axis). Today: 64 graphs, k = 5,000 take seconds with ReadWork = 2,016. After the fix: the pair passes `MaxPairWork` (its lines at 0 facts, weak) and the read passes `MaxReadWork` (every line 0 facts) before any compare, and the row finishes in well under a second.
Do:
- Each member carries a walk count, read once per read with its tally (the `holdings` cache): the number of its links (every `memberLink` under every far entity) plus the number of property keys it holds that count. `pairWork` charges, per member pair, `walk(x) + walk(y)` **before** it builds either member's `cost` (`costOf`), and stops past `limit` there, so a member whose links alone pass the cap is never walked per pair. Then it adds the products as today.
- Through `readPlan`, this charge reaches `MaxReadWork` with no other change (each pair still contributes `min(w, MaxPairWork)`).
- Add the "unshared links" axis to `axes` (`underCap` finds its n) so `TestLinesWorkCap`, `BenchmarkLinesAtCap` and `BenchmarkPairsAtCap` cover it; `BenchmarkPairsAtCap` runs it as a second sub-benchmark beside "names".
- Re-measure: `ReadWork(householdGraphs(50))` and the obituary's `PairWork`. Set `MaxReadWork` again to 20 × the new household work, rounded up to two significant digits (plan D-3). Keep `MaxPairWork` at 25,000 if it is still at least 20 × the obituary's new work; if it is not, **stop and report** the numbers (that value is the user's).
- Update the `Rules` field docs for `MaxPairWork` and `MaxReadWork`, the matching-model paragraph and both rows to say work also counts each member's links and properties per member pair.
Check: the new rows pass; every existing `TestLines*`, `TestPairs*`, `FuzzLines` seed and `TestRecordLinksBudgets` passes (a `TestLinesWorkCap` row whose n moved because the count grew is updated with the new n, and named in the report); `BenchmarkPairsAtCap` both axes reported with m and ns/op; `TestKnobInventoryMatchesRegistry` passes; scripts/check-pr.sh PASS.

**F-2 — One owner for each pair's cap result** (from PRF-1, MNT-1)
Files: `core/recordlinks/recordlinks.go`.
Do: `readPlan` stores `over bool` (`w > MaxPairWork`) on `sharedPair`; `Pairs` passes `skip || p.over` to `linesOf` as its capped flag; `linesOf` takes the decision and no longer calls `overWork`. `pairLines` (the one-pair count) still calls `overWork` itself.
Check: `rg -n 'overWork\(' core/recordlinks/recordlinks.go` prints only the definition and the call in `pairLines`; `TestPairsInvariants` and `TestPairsReadCap` pass.

**F-3 — Delete readWork** (from MNT-2)
Files: `core/recordlinks/recordlinks.go`, `core/recordlinks/export_test.go`.
Do: delete `readWork`; `ReadWork` in `export_test.go` calls `readPlan(string(page), graphs, holdings{}, cfg, math.MaxInt)` and returns its count.
Check: `rg -n 'readWork' core/recordlinks` prints nothing; tests pass.

**F-4 — Compare the 65-graph read by graph id** (from CON-1)
Files: `core/database/persondetail/recordlinks_perf_test.go`.
Do: the "past the cache" subtest compares the `GraphID` sequence of `first.Graphs` and `cold.Graphs` as well as `Pairs`.
Check: `TestRecordLinksBudgets` passes.

**F-5 — Hold the read's caps under generated input** (from CON-2)
Files: `core/recordlinks/invariants_test.go`.
Do: in `TestPairsInvariants`, each round sets `MaxReadWork` to a random value in `[0, ReadWork + 2]` and, one round in three, `MaxPairWork` to a random value below 40; assert: nil when the number of lines passes the cap; every pair listed with every line at 0 facts and weak when the work passes it; equal to `Lines` per pair otherwise; the reversed read equals the mirrored read. The reviewer's `/private/tmp/claude-501/-Users-jakerobins--claude-spares-spare-exgiSE/1df29678-6114-41d4-97b5-a1bf452a254b/scratchpad/probe/zz_probe_test.go` is that body; move it in.
Check: the test passes over its rounds.

**F-6 — One fixture per record-links case** (from CON-3)
Files: `core/database/persondetail/recordlinks_test.go`.
Do: `TestRecordLinks` becomes a table of `{name, build func(f *fixture) (person []byte), check}`, each row on a fresh `newFixture`; AC-1.3's count reads only its own graphs; the merged-handle row writes `merged_into_id` on its own fixture.
Check: `go test -run TestRecordLinks ./core/database/persondetail/` passes; each row passes alone (`-run 'TestRecordLinks/<row>'`).

## Round 1 — fix
