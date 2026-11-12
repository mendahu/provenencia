# Ship log — S16-12 Source records in the graph cache

**Plan:** [`S16-12-source-records.md`](S16-12-source-records.md) · **Branch:** `s16-12-source-records` · **PR:** #800
**State:** ready

Base: `main` @ `83895e4`.

## Round 0 — implement
Head `c01f4f07f` (STEP-1 committed, local only; STEP-2 code and T-1–T-3, T-13 written, uncommitted). Gate not run.

**Stopped on** §11: `BenchmarkSourceRecords` heap over the 5 MB stop at 3,000 Sources × 10 Artifacts. Three runs: Build 25.4–26.9 ms (budget 50), heap 5.74–5.78 MB (budget 5), Apply 0.05–0.06 ms ordinary, 7.8–8.4 ms for the 10,000-Artifact Source (budget 20). Cause: D-3's `ArtifactLabels` map per Source plus D-5's `sourceOfArtifact` map, 30,000 string-keyed entries each. Pre-sizing did not help.

Baseline `BenchmarkCacheSizes` on `main` (STEP-1): evidence 34.4 / 44.6 MB, nodes 57.0 / 76.9 MB, matchIndex 6.2 / 9.7 MB, kinship 1.4 / 4.4 MB, aliasIndex 0.77 MB (4,000 / 16,000 persons).

**User, 2026-11-12:** raise the stop to 8 MB. Plan §7 and §11 updated in `6e1314400`. Implementer resumed from STEP-2.

Resumed: head `d104bf9`. Gate PASS (orchestrator run). Push blocked: the Mac's system resolver cannot resolve github.com (`curl`, `ssh`, `dscacheutil` fail; `host` resolves), so no draft PR yet. Report deviations: `TestUpdateAndCoverReturnEffects` updated for D-9 (a rename names no graph); `TestEvidenceGraphsResolve` gains a search-document assertion for AC-2.3; `PersonRecordLinksContentTests` `graph()` helper now sets `sourceTitle: "T"`. `BenchmarkSourceRecords`: Build 24 ms, heap 5.7 MB (stop 8 MB), Apply 0.05 / 8.0 ms. `sourceRecords` 16 KB in `BenchmarkCacheSizes` (large fixture holds one Source). Manual check not run.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Skipped: none (UX: `macos/App/**`; SEC+PRF: `api/ffi/**`, `core/database/**`, `core/catalogsession`, loads stored credibility order).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | major | A crafted grade `sort_order` overflows the relative column to REAL; the scan into `int` fails every `RecordsOf` | Fix now → F-1 | verified: `RelativeOrderColumn` (`sourcecredibilitygrades/registry.go:16`) has no bound; reviewer's sqlite run gives `real`; `listen.go:34-38` drops all listeners on an `OnCommit` error. Crafted-catalog only, so one high clamp, not a design |
| CON-4 | CON (from SEC outside-lens) | major | Apply removes and re-inserts per Source, so an Artifact swap between two named Sources loses a mapping | Fix now → F-2 | verified: `sources/records.go:96-108` interleaves delete and insert per id; A S1→S2, B S2→S1 with `[S1,S2]` deletes B's new mapping. Code this PR adds; S15-24 reaches it, so rated as merged |
| MNT-1 | MNT | minor | Credibility join written twice; reconciler's three fragments have one caller each | Fix now → F-3 | same-pass cleanup: this PR touches both copies (`sources.sqlRecordSources`, `autoreconciler.provenanceJoins`) |
| MNT-3 | MNT | minor | `queryEach` copies `sqlbatch.Query`'s scan loop | Fix now → F-4 | same-pass cleanup: new copy added by this PR |
| MNT-5 | MNT | nit→minor | `BenchmarkSourceRecords` hand-rolls the heap measure `promotealign` has | Fix now → F-4 | same-pass cleanup (any severity): this PR adds the copy |
| MNT-2 | MNT | minor | `Record.ArtifactLabels` shares the cache's map with answer values | Fix now → F-5 (comment) | D-3 fixes the field's shape, so the shape change is out; the read-only note on `Record` is in a file this PR adds |
| UX-1 | UX | minor | Translator comment for `recordLinks.nodeHelp` still describes the "else ref" fallbacks | Fix now → F-5 | verified: `L10n.swift:9789`, `Localizable.xcstrings:53` |
| CON-1 | CON | minor | `evidencelayer.go:52-53` says the memo drops on a commit naming its Source | Fix now → F-5 | verified: `"sources"` has no `EvidenceGraph:` path after D-9 |
| PRF-1 | PRF | minor | Record links recomputed on any write in a cited Source | Later PR → S16-25 | plan Out line defers a narrower signal; written into S16-25's **Check** in the map (no Decision reopened) |

Nits (logged): `MNT-4 · conclusiondetails.go:428 · name the param sourceRecs` · `MNT-6 · recordlinks.go:72-76 · take []SourcedGraph (D-12 kept it)` · `CON-2 · graphcache/evidence_graph.go:91-92 · drop "so no Source drop removes it"` · `CON-3 · docs/catalog-caches.md:19 · say the large fixture has one Source`

### Fix plan 1

**F-1 — Bound the relative credibility column** (from SEC-1)
Files: `core/database/sourcecredibilitygrades/registry.go`, `core/database/sources/records_test.go`, `core/database/autoreconciler/autoreconciler_test.go` (if it has a provenance read test).
Failing test first: `TestRecordsBuild/out-of-range grade` — a grade with `sort_order` = 9223372036854775807 assessed on a Source while the default is negative-large (or set the default's `sort_order` to −9223372036854775807): Build errors today.
Do: clamp each `sort_order` operand inside `RelativeOrderColumn` to one high bound before subtracting (e.g. `MAX(-1000000, MIN(1000000, …))` on both the assessed grade and the default subquery), so the difference is always an INTEGER. One constant names the bound, with a comment that real grades are small and the bound only stops crafted catalogs. Build succeeds and `Credibility` reads the clamped difference. The reconciler reads the same column, so it is covered by the same change.
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — Apply removes every named Source before inserting any** (from CON-4)
Files: `core/database/sources/records.go`, `core/database/sources/records_test.go`.
Failing test first: `TestRecordsApply/artifact swap` — Artifacts A (in S1) and B (in S2) swap Sources in one write whose `SourceScopes` is `[S1, S2]`; after Apply, `Artifact(B)` → S1 and `Artifact(A)` → S2, and `g.Verify()` is nil. If no production write can swap today, drive it with a raw `UPDATE artifacts` in the test's tx and the `effects.Set` it would produce. Also the order `[S2, S1]`.
Do: in Apply, first loop over `set.SourceScopes` deleting each old record and its Artifact mappings; then a second loop inserting each fresh record and its mappings.
Check: the test passes, also with `PROVENENCIA_GRAPH_VERIFY=1`; scripts/check-pr.sh PASS.

**F-3 — The grades package owns the credibility join** (from MNT-1; same-pass cleanup)
Files: `core/database/sourcecredibilitygrades/registry.go`, `core/database/sources/records.go`, `core/database/autoreconciler/autoreconciler.go`.
Do: add `RelativeOrderJoins(sourceIDExpr string) string` (or equivalent) beside `RelativeOrderColumn`, producing the `LEFT JOIN source_credibility_assessments sca ON sca.source_id = <expr> LEFT JOIN source_credibility_grades sg …` pair that puts `sg` in scope; update `RelativeOrderColumn`'s comment to point at it. `sources.sqlRecordSources` and the reconciler compose from it. Inline the reconciler's `provenanceColumns`, `provenanceJoins` and `provenanceDest` into `sqlLoadCandidates` and its scan, dropping the stale "sg is in scope for a caller's own column" sentence.
Check: `rg -n "JOIN source_credibility_assessments" core --type go | grep -v _test` prints only the line in `sourcecredibilitygrades/registry.go`; `rg -n "provenance(Columns|Joins|Dest)" core` prints nothing; autoreconciler and sources tests pass; scripts/check-pr.sh PASS.

**F-4 — One query-and-scan loop, one heap measure** (from MNT-3, MNT-5; same-pass cleanups)
Files: `core/database/sqlbatch/sqlbatch.go`, `core/database/sources/records.go`, `core/database/sources/records_bench_test.go`, `core/database/promotealign/bench_test.go`, a shared test-support file (`core/database/databasetest/` holds `budget.go`; put it there unless a better existing home exists).
Do: export `sqlbatch.Scan(q, query string, args []any, scan func(*sql.Rows) error) error`, used by `sqlbatch.Query` per batch and by `sources.loadRecords` for the unbatched read; delete `queryEach`. Move the GC/`HeapAlloc` before-and-after measure into one exported helper in the shared test-support package; `promotealign`'s `measure` and `BenchmarkSourceRecords` both call it.
Check: `rg -n "func\(\*sql.Rows\) error" core --type go | grep -v _test` prints only `sqlbatch.go` lines; `rg -n "HeapAlloc" core --type go` prints only the shared helper; `BenchmarkSourceRecords` and `BenchmarkCacheSizes` still report the same metrics; scripts/check-pr.sh PASS.

**F-5 — Stale comments and the read-only note** (from UX-1, CON-1, MNT-2)
Files: `macos/App/Platform/L10n.swift:9789` and `macos/App/Resources/Localizable.xcstrings:53` (comment only: "arguments are the graph's name and the Source's title"; key and value unchanged); `core/database/evidencelayer/evidencelayer.go:52-53` ("dropped with the graph (a commit that moves the graph's rows, `Set.EvidenceGraphs`)"); `core/database/sources/records.go` (on `Record.ArtifactLabels`: owned by the held `RecordIndex`; readers treat it as read-only).
Check: `rg "else ref" macos/App` prints nothing; `rg -n "names its Source\)" core` prints nothing; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `46c0c1a`. Gate PASS (orchestrator run). F-1 to F-5 done: `TestRecordsBuildOutOfRangeGrade` and `TestRecordsApplySwap` red first, then green (swap also under `PROVENENCIA_GRAPH_VERIFY=1`). Deviations: F-1–F-3 in one commit (`eb359bce9`); `sqlLoadCandidates` and `sqlRecordSources` became `var`s because `RelativeOrderJoins` is a function; `databasetest.HeapDelta`/`LiveHeap` replace every hand-rolled heap read in `promotealign` benches too; `rg "else ref"` still hits unrelated "Nothing else references it." copy. Network back: branch pushed, draft PR #800 opened.

Minors read by the orchestrator: MNT-1 Resolved (one join in `sourcecredibilitygrades`, trio gone); MNT-3 Resolved (`sqlbatch.Scan`, `queryEach` gone); MNT-5 Resolved (`databasetest/heap.go`); MNT-2 Resolved (read-only note on `Record.ArtifactLabels`); UX-1 Resolved; CON-1 Resolved. MNT and UX not sent back.

## Round 2 — review
Reviewers: CON (CON-4 major), SEC+PRF (SEC-1 major). Scope `d104bf9...46c0c1a`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-4 | CON | major | Artifact swap loses a mapping | Resolved | `records.go:95-112` two passes; `TestRecordsApplySwap` both orders, also under `PROVENENCIA_GRAPH_VERIFY=1` |
| SEC-1 | SEC | major | Crafted `sort_order` overflows to REAL | Resolved | `registry.go:12-23` clamps both operands to ±1,000,000; sqlite3 at int64 extremes gives `integer`; STRICT `INTEGER` column; `TestRecordsBuildOutOfRangeGrade` |

No new findings in round 2. No Fix plan 2.

## Result
Gate PASS · CI pending · open: none · later PR: PRF-1 (S16-25 Check) · follow-ups: none · to user: — (heap stop raised to 8 MB, answered 2026-11-12)
