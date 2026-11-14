# Ship log — S16-20 One match index that Apply keeps current

**Plan:** [`S16-20-match-index-apply.md`](S16-20-match-index-apply.md) · **Branch:** `s16-20-match-index-apply` · **PR:** #827
**State:** ready

## Round 0 — implement
Dispatched `pr-implementer` against `main` @ `c09b83b9e` (plan's Planned base `64a91ae`).
Head `2fdf6d460`. Gate PASS (orchestrator run against `c09b83b9e`). Report deviations: `profiler.readers` reads a neighbour's kind from `Link.NeighborTypeKey`; no pointer-returning test hook, the index tests compare two `MatchIndex.Peek` results (D-14 allows); those tests set `PROVENENCIA_GRAPH_VERIFY=1`; `inputs.namePattern` stays in `promotealign.go` unread by `Propose`; §8 of `promote-graph-alignment.md` changed one extra sentence and dropped the `promote-ranking.md#name-index-on-commit` link; §10 manual check not run.

Figures (Apple M1; base `c09b83b9e`):

| | base | after |
| --- | --- | --- |
| `after-commit-ms`, 4,000 / 16,000 | 557.9 / 919.9 | 390.8 / 666.9 |
| `warm-ms`, 4,000 / 16,000 | 112.5 / 124.5 | 114.1 / 124.2 |
| `matchIndex-B`, 4,000 | 6,161,224 | 35,539,136 |
| `matchIndex-B`, 16,000 | 9,724,200 | 57,943,640 (stop 60 MB) |
| `WarmOnOpen` best of 3, 4,000 / 16,000 | 615 / 791 ms | 641 / 835 ms |
| Go caches' `-B` sum, 16,000 | 162 MB | 179 MB (stop 250 MB) |

Apply fan-out: 11 ms for an event with 500 participants, 32 ms for a father of 2,000 births (ceiling 250 ms).

## Round 1 — review
Reviewers: CON, MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, `core/apperr`, or error codes in the diff).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1, CON-2 | MNT, CON | major | `inputs.namePattern` is written by `loadInputs` and read by nothing | Fix now → F-1 | verified: `grep -n namePattern promotealign.go` shows only the field, `loadConfig`, the local and the literal; `Propose` was its last reader (D-9). D-3 keeps `namePattern` on `profiler`, not on `inputs` |
| MNT-2 | MNT | minor | "a handle merged" test body copied; `unpromote` only on `*warmFixture`, inline copy in `perf_test.go` | Fix now → F-2 | same-pass cleanup: this PR added the second copy (`memo_test.go:498-515`) |
| MNT-3 | MNT | minor | Index membership rule (non-nil, unmerged, primary kind) written in Build and again in Apply | Fix now → F-3 | D-3's Rejected column already rules out "the order in Build and again in Apply — two copies of one rule"; `index.go:74-76` vs `:142` |
| MNT-4 | MNT | minor | `loadCanon`'s `kinds` map duplicates `m.handles[id].Kind` | To user | true (`canon.go:72,178` write `h.Kind`/`nh.Kind`, one read at `:140`), but D-9 says `loadCanon` "keeps … its `kinds` map". Not fixed around the Decision; asked at finish (does not block the other fixes) |
| MNT-5 | MNT | minor | `profiler` stores both `rules` (= `cfg.Subject`) and `cfg` | To user | true (`profiles.go:44,47,67`), but D-3 names both fields in `profiler{…}`. Asked at finish with MNT-4 |
| MNT-6 | MNT | minor | `TestIndexPut` is hand-written subtests, not a table | Fix now → F-5 | `.cursor/rules/go-tests.mdc` binds; T-4 says "table" |
| PRF-1 | PRF | minor | `filed` holds its own copy of every key string; D-4's "shares bytes with `byKey`" is false; index at 57.9 MB vs 60 MB stop | Fix now → F-4 | measured by reviewer: `filed` alone 15.5 MB at 30,000 handles; code this PR adds; D-4's premise made true rather than the Decision changed |
| SEC-1 | SEC | minor | `Put`/`Remove` index catalog values with no fuzz or generated-input test | Fix now → F-5 | rubric asks for one on code keying untrusted values; `grep "func Fuzz" core/shortlist` empty |
| CON-1 | CON | minor | `TestMatchIndexApplyFanOut` never checks the index survived the commit | Fix now → F-6 | reviewer's mutation (`len(roots) > 2` → `Drop`) still passes the test |
| CON-3 | CON | minor | §10 manual check not run | To user | the manual check is the user's at finish; listed in the report |

Nits (logged): `MNT-7 · core/database/graphcache/cache.go:548,564 · Peek only forwards to held; rename or move` · `MNT-8 · promotealign/index.go:112 · read through w.Peek(g) like siblings` · `MNT-9 · catalogsession/session.go:83 · "matching's match index"` · `MNT-10 · docs/promote-graph-alignment.md:170, catalogsession/warm_test.go:55,83 · "the shortlist kinds" now every primary kind` · `CON-4 · promotealign/memo_test.go TestMatchIndexHoldsEveryHandle · AssertKindsLoaded runs after the test's own g.Kind calls`

MNT-4 and MNT-5 contradict D-9 and D-3, so they wait for the user; Fix plan 1 does not touch `kinds` or `profiler`'s fields and proceeds.

### Fix plan 1

**F-1 — Delete the dead `inputs.namePattern`** (from MNT-1, CON-2)
Files: `core/database/promotealign/promotealign.go`.
Failing test first: none (dead field; the compiler is the check).
Do: delete the `namePattern` field from `inputs`, its local and its struct-literal entry in `loadInputs`; take `cfg, _, err := loadConfig(...)` there. Leave `profiler.namePattern` (D-3).
Check: `grep -n "namePattern" core/database/promotealign/promotealign.go` prints only `loadConfig`'s lines; `go build ./...`; scripts/check-pr.sh PASS.

**F-2 — One merge helper and one `unpromote` for both test files** (from MNT-2)
Files: `core/database/promotealign/memo_test.go`, `core/database/promotealign/perf_test.go` (and the shared `*fixture` file).
Failing test first: none (test refactor).
Do: add `mergePerson(t)` beside `renameMember` holding the `UPDATE canonical_entities SET merged_into_id` and the hand-built `OnCommit` at `Revision: 1 << 40`; call it from `TestMemoFollowsCommit` "a handle merged" and `TestMatchIndexFollowsCommits` "a handle merged". Move `unpromote` from `*warmFixture` to `*fixture` and call it in `perf_test.go` in place of the inline `writes.Call(... promote.Unpromote ...)`.
Check: `grep -n "merged_into_id" core/database/promotealign/*_test.go` prints one line (the helper); `grep -n "unpromote_subject\|promote.Unpromote" core/database/promotealign/*_test.go` prints one call site (the helper); all promotealign tests pass, also with `PROVENENCIA_GRAPH_VERIFY=1`.

**F-3 — One `profiler.holds(n)` for which handles the index holds** (from MNT-3)
Files: `core/database/promotealign/profiles.go`, `core/database/promotealign/index.go`.
Failing test first: none (behavior unchanged; T-3 and `TestVerifyChecksMatchIndex` hold it).
Do: add `func (p *profiler) holds(n *graphcache.Node) bool` — non-nil, not merged, `vocabkey.IsPrimaryKind(string(p.kind(n)))` — in `profiles.go`. Use it in `applyMatchIndex`'s removal test (`!p.holds(n)`) and to filter the nodes `buildMatchIndex` passes to `p.subjects`. `subjects` keeps its own nil/merged skip (D-3).
Check: `grep -n "IsPrimaryKind" core/database/promotealign/*.go` prints only `holds`; promotealign tests pass with and without `PROVENENCIA_GRAPH_VERIFY=1`; scripts/check-pr.sh PASS.

**F-4 — `filed` shares its key strings with the postings** (from PRF-1)
Files: `core/shortlist/index.go`, `core/shortlist/shortlist_test.go`, `core/database/promotealign/perf_test.go` (figures only).
Failing test first: `TestIndexPut` (or a new subtest) asserting, through `unsafe.StringData`, that each key in a handle's filing points at the same bytes as the posting map's key for that key — red today.
Do: intern keys per index: when `Put` inserts into a posting, reuse the existing map key string (keep it on the posting, e.g. `posting{key string; ids [][]byte}`, or one `map[string]string` intern held by the `Index`), and store that string in `filing.keys`. A patched index must still `reflect.DeepEqual` a fresh build (T-4, Verify). Re-run `BenchmarkCacheSizes` and record `matchIndex-B` at 4,000 and 16,000 in the plan's Landed note and the PR body; correct D-4's Landed note wording if the saving is not the full copy.
Check: the new test passes; `TestIndexPut` and F-5's test pass; `matchIndex-B` at 16,000 below its round-0 57,943,640; scripts/check-pr.sh PASS.

**F-5 — `TestIndexPut` as a table, plus a seeded random `Put`/`Remove` check** (from MNT-6, SEC-1)
Files: `core/shortlist/shortlist_test.go`.
Failing test first: n/a (tests over code that passes the reviewers' probes); confirm the random check fails by temporarily breaking `unfile` (mutation), then restore.
Do: fold the five same-shape subtests of `TestIndexPut` into one table `{name, start []entry, op func(*shortlist.Index), want []entry}` with one loop; keep "a handle is refiled" and the nil-index check separate. Add `FuzzIndexPut` (seed corpus of a few byte strings decoding to op sequences) or a seeded table over fixed seeds: random `Put`/`Remove` sequences over ~8 ids with person, place and event kinds, empty values and kind changes, then `reflect.DeepEqual` against an index built by `Put` of the final state in sorted id order. It runs under plain `go test`.
Check: tests pass; the mutation fails the random check; scripts/check-pr.sh PASS.

**F-6 — The fan-out test asserts the index survived** (from CON-1)
Files: `core/database/promotealign/perf_test.go`.
Failing test first: with `applyMatchIndex` temporarily dropping on `len(roots) > 2`, `TestMatchIndexApplyFanOut` must fail; restore.
Do: after each timed `k.Commit` in both subtests, `if _, ok := promotealign.MatchIndex.Peek(g); !ok { t.Fatal("the commit dropped the match index") }` (model: `stored` in `memo_test.go`).
Check: test passes; the mutation fails it; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `6e1e95f99`. Gate PASS (orchestrator run). F-1 to F-6 done. `matchIndex-B` 33,979,848 at 4,000 and 55,095,528 at 16,000 (was 35.5 / 57.9 MB); Go caches' sum at 16,000 about 176.5 MB. Deviations: F-4 saved 2.85 MB at 16,000, not the full copy (a string header per key and the posting's own `key` field remain), recorded in the Landed note, D-4 not reworded; F-5 is a seeded test (100 seeds × 60 ops), not a `Fuzz` function; F-6's mutation used `len(roots) > 0`, since the timed commits carry at most 2 roots; F-6 helper named `indexSurvived`.

Minors read by the orchestrator (no round 2 for them):
- PRF-1 — **Resolved.** `core/shortlist/index.go`: `posting{key, ids}`; `Put` and `unfile` assign by `post.key` so the map keeps the first key string; `filing.keys` takes `post.key`. Share test red→green per report.
- SEC-1 — **Resolved.** `TestIndexPatchesEqualBuild`; `unfile` mutation fails it.
- CON-1 — **Resolved.** `indexSurvived(t, g)` after each timed commit in both subtests of `TestMatchIndexApplyFanOut`.
- MNT-2, MNT-3, MNT-6 ride with MNT-1 into the MNT round-2 review.

## Round 2 — review
Reviewers: MNT (had the round-1 major MNT-1). Not sent: CON and SEC+PRF (minors only in round 1).

MNT-1 — **Resolved** (`inputs.namePattern` gone; `loadInputs` takes `cfg, _, err := loadConfig(...)`; the name pattern lives only on `profiler`, per D-3). MNT-2, MNT-3, MNT-6 — **Resolved** (one `mergePerson`, `unpromote` on `*fixture`; `profiler.holds` the only `IsPrimaryKind` in promotealign; `TestIndexPut` one table). No new findings. No Fix plan 2.

## Result
Gate PASS · `origin/main` unchanged since `c09b83b9e` · CI pending at ready · open blockers/majors: none · follow-ups: none · to user: MNT-4 (`loadCanon`'s `kinds` map vs D-9), MNT-5 (`profiler.rules` beside `cfg` vs D-3), CON-3 (§10 manual check not run).
