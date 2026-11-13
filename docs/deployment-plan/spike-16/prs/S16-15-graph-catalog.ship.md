# Ship log — S16-15 An Evidence-graph catalog in the graph cache

**Plan:** [`S16-15-graph-catalog.md`](S16-15-graph-catalog.md) · **Branch:** `s16-15-graph-catalog` · **PR:** #805
**State:** ready

## Round 0 — implement
Stopped in STEP-4. Commits `97c6c964d` (STEP-1), `352d1fd69` (STEP-2); STEP-3 `index.go` and STEP-4 read wiring written, uncommitted.
Baseline `BenchmarkCacheSizes` on `main` (STEP-1): persons=4000 — nodes 57047832 B, evidence 34674656 B, matchIndex 6161176 B, nodes-ms 645.0; persons=16000 — nodes 76877368 B, evidence 44630000 B, matchIndex 9920952 B, nodes-ms 740.0.

**Stopped on:** "A test fails after STEP-4 for a reason other than D-11." `api/ffi/handlers/evidence_graphs_test.go` `TestListEvidenceGraphs` asserts the project-wide order `[Default Default Smith]` and now gets `[Default Smith Default]`. D-4 changes that order on purpose (Sources by id, then each Source's list order), but T-8 lists the test as "unchanged". The plan contradicts itself; taken to the user.

**User, 2026-11-12:** amend T-8 — `TestListEvidenceGraphs` changes to D-4's order, derived from `bytes.Compare` of the two Source ids. Plan amended in `3feb6d888`; implementer resumes from STEP-3.

Resumed. Head `79aba6f8a`. Gate PASS (orchestrator run, against `97325549b`). Report deviations: `index.go` Apply guards `loadProgress` with `len(graphIDs) > 0` (nil would recount every graph); `plan-docs` failed until the landing commit set the plan `Landed` (plan names the deleted `graphprogress` paths); handler test gained `mustParseID`; `TestFiledCounts` gained a D-11 `Drop`.
After `BenchmarkCacheSizes`: graphIndex 12680 B at both 4,000 and 16,000 persons; others within noise of baseline. `BenchmarkGraphIndex`: build-ms 96.7 (≤150), heap-MB 5.2 (≤8), apply-progress-10k-ms 0.43 (≤20), apply-list-10k-ms 32.7 (≤50).

## Round 1 — review
Reviewers: CON, MNT, SEC+PRF (`api/ffi/handlers/**`, `core/database/**`). Skipped: UX (no `macos/App/**`, strings, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `benchID` is the third copy of the bench id minter | Fix now → F-1 | verified: same body in `evidencegraphs/index_bench_test.go:17`, `sources/records_bench_test.go:18`, `promotealign/bench_test.go:220` (fixed `0xB0`); same-pass cleanup |
| MNT-4 | MNT | minor | `IndexOf`/`ErrIndexOnWrite` copy `RecordsOf`/`ErrRecordsOnWrite` | Fix now → F-2 | verified: identical 12-line bodies (`sources/records.go:134`, `evidencegraphs/index.go:184`); same-pass cleanup, supersedes D-2's "Rejected: a shared guard … logged for the third" |
| MNT-2 | MNT | minor | `GetProgress` writes out its own copy of `catalogIndex` | Fix now → F-3 | verified `progress.go:58-70`; D-9 binds the steps, not their order between id check and `c.DB()` |
| MNT-3 | MNT | minor | "Evidence-graph index" names both the Go entry and the Swift cache | Fix now → F-4 | verified `effects/README.md`, `docs/catalog-caches.md`, `cache-architecture.md:54-55`; lines this PR adds |

SEC+PRF: no findings (bench re-run at Scale: build-ms 86.8, heap-MB 5.2, apply-list-10k-ms 20.9). CON: no findings.

Nits (logged): `MNT-5 · api/ffi/handlers/evidence_graphs_test.go:121 · mustParseID duplicates uuid.MustParse(s)[:]` · `MNT-6 · evidencegraphs/index.go:91-163 · move Apply body into a GraphIndex method` · `CON-1 · index_test.go TestGraphIndexApply · AC-1.3 names GetProgress(def), test reads idx.Progress` · `CON-2 · index.go GraphIndex.Progress · held row returns the held GraphID slice, D-3 says a copy`

### Fix plan 1

**F-1 — One bench id minter in `databasetest`** (from MNT-1)
Files: `core/database/databasetest/heap.go` or a new `core/database/databasetest/ids.go`; `core/database/evidencegraphs/index_bench_test.go`; `core/database/sources/records_bench_test.go`; `core/database/promotealign/bench_test.go`.
Do: add `func BenchID(kind byte, n int) []byte` (16 bytes, `id[0] = kind`, `binary.BigEndian.PutUint64(id[8:], uint64(n))`) to `databasetest`, beside `HeapDelta`. Replace the three local `benchID` functions with it; `promotealign` passes `0xB0`. Delete the local copies.
Check: `grep -rnE "func benchID|PutUint64\(id\[8:\]" core api` prints only the new `databasetest` line. (`metadatafields_test.go:622`'s inline `PutUint64(unknown[8:]…)` is a different, one-off use and stays.) Benchmarks still compile: `go test -tags fts5 -run '^$' -bench 'BenchmarkGraphIndex|BenchmarkSourceRecords|BenchmarkCacheSizes' -benchtime 1x ./core/database/evidencegraphs/ ./core/database/sources/ ./core/database/promotealign/`. scripts/check-pr.sh PASS.

**F-2 — One guard for reading a graph-lifetime whole entry** (from MNT-4; supersedes D-2's Rejected "a shared `graphcache` guard and sentinel … logged for the third")
Files: `core/database` (a small new file beside where `ErrClosed` lives); `core/database/sources/records.go`, `records_test.go`; `core/database/evidencegraphs/index.go`, `index_test.go`; `docs/catalog-caches.md` (whole-entry paragraph).
Failing test first: change `TestIndexOf` and `TestRecordsOf` to expect the shared sentinel on `graphcache.Over(tx)`.
Do: `graphcache` cannot return `database.ErrClosed` (`database` imports `graphcache`), so the helper lives in `database`: `func ReadWhole[T any](w graphcache.Whole[T], g *graphcache.Graph) (T, error)` — nil graph → `ErrClosed`; `Get` error → that error; `ok == false` → `ErrWholeOnWrite`, one sentinel `errors.New("database: a graph-lifetime cache is not read on a write's graph")`. If `Whole[T]` is not a value type that can be passed this way, take `*graphcache.Whole[T]`. `RecordsOf` and `IndexOf` stay as the packages' typed entry points (S16-13 is planned against them) and each becomes one line over `database.ReadWhole`. Delete `ErrRecordsOnWrite` and `ErrIndexOnWrite`; update their doc comments and tests. In `docs/catalog-caches.md`'s whole-entry paragraph, say a graph-lifetime whole entry is read through `database.ReadWhole`. Add a line to the plan's Landed note that D-2's shared-guard rejection was superseded by the same-pass rule (ship log MNT-4).
Check: `grep -rn "OnWrite = errors.New" core` prints only `ErrWholeOnWrite`; `TestIndexOf`, `TestRecordsOf` pass; scripts/check-pr.sh PASS.

**F-3 — `GetProgress` reads through `catalogIndex`** (from MNT-2)
Files: `core/database/evidencegraphs/progress.go`.
Do: validate the id first (`ErrInvalid`), then `idx, err := catalogIndex(c)`, as `List` does. Drop the inline `c.DB()` and `IndexOf` calls.
Check: `TestGraphProgress` (incl. closed catalog) passes; scripts/check-pr.sh PASS.

**F-4 — Name the Go entry by its symbol in prose** (from MNT-3)
Files: `core/database/effects/README.md`, `docs/catalog-caches.md`, `docs/deployment-plan/spike-16/cache-architecture.md`.
Do: call the Go entry "the graph index (`evidencegraphs.Index`)" in the README and in `cache-architecture.md`'s S16-15 lines, in place of "Evidence-graph index" / "Evidence-graph catalog" there. In `docs/catalog-caches.md`, rename the Swift row to "Evidence-graph index (Swift)" so the two rows differ. Do not rename the plan, map row or `completed.md` titles.
Check: `grep -rn "Evidence-graph index" core/database/effects docs/catalog-caches.md docs/deployment-plan/spike-16/cache-architecture.md` prints only the Swift row; `scripts/check-plan-docs.py --links` 0 errors; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `a0d1f9d34`. Gate PASS (orchestrator run). F-1 `bd7b90f69` (`databasetest.BenchID`), F-2 `209e518bf` (`database.ReadWhole`, `ErrWholeOnWrite`; tests red first), F-3 `f3ab4d3c9`, F-4 `a0d1f9d34`. Deviations: F-4's `catalog-caches.md` edit rode in F-2's commit.

## Round 2 — review
Reviewers: MNT (had a Fix-now major). Scope `79aba6f8a...a0d1f9d34`. Not re-reviewed: CON, SEC+PRF (no findings in round 1).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) `benchID` third copy | Resolved | `grep -rnE "func benchID\|PutUint64\(id\[8:\]" core api` prints only `databasetest/ids.go:10` |

No new findings. Minors MNT-2, MNT-3, MNT-4: Resolved (orchestrator read of `f3ab4d3c9`, `209e518bf`, `a0d1f9d34`; the round-2 reviewer confirmed each with grep).

Merged `origin/main` (`df55dbae5`, S15-33 plan docs only); gate PASS against `df55dbae5`.

## Result
Gate PASS · CI green · open: none · follow-ups: none · to user: — (T-8/D-4 amendment answered by the user before review)
