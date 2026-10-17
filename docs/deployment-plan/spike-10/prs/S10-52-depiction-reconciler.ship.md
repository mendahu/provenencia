# Ship log — S10-52 Depiction auto-reconciler module

**Plan:** [`S10-52-depiction-reconciler.md`](S10-52-depiction-reconciler.md) · **Branch:** `s10-52-depiction-reconciler` · **PR:** #437
**State:** ready

## Round 0 — implement
Head `2df46b5a`. Gate PASS (orchestrator rerun). Report: no deviations from Decisions/Scope; T-3 cases share a `world` struct and `finish` helper (noted in Landed); one earlier xcode-test flake on `WorkspaceSessionTests.invalidateAllMatching()` (no Swift touched; rerun passed); manual check not run.

## Round 1 — review
Base `51cbf1e6`, head `2df46b5a`, five lenses. PRF and UX clean. No majors.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | blocker → minor | Cached `value_image_file_id` returned by `graphcache.loadValues` without the `Readable` check | Rejected (as a blocker); Follow-up for S10-35 | Verifier: Refuted. `Readable` tests only `files.media_type`, which a crafted folder writes as easily as the cache, so the check would add no protection. Every cached column is trusted as derived data (D-4 "same shape as `value_entity_id`"; D-8, S10-54 D-6: graphcache imports no value package). No non-crafted path gives a non-PNG id: `files.media_type` is never updated, and the cache is filled from `FromJoin` output. No reader in this PR hands the id to the app. `files.RelPaths` builds paths from checksum and media type, so no path can escape `objects/`. Home: the S10-35 gallery drops non-PNG Files when it calls `RelPaths` and decodes defensively |
| CON-1 | CON | minor | AC-1.7 subtest does not assert which image survives | Follow-up | No fix plan; outcome check already pins B as `no_evidence` |
| CON-2 | CON | minor | AC-1.1 does not assert `DateLo`/`DateHi`/`SortKey` NULL | Follow-up | Follows T-3's field list; `insertRow` leaves them NULL (D-3) |
| MNT-1 | MNT | minor | `insertRow` and `truthRow` each encode a cache row; depiction branch in both | Follow-up | Plan Out row (review M-5); `Verify` catches drift. Suggest a `docs/ideas/` note: one `cacheRow` encoder |
| MNT-2 | MNT | minor | `outcomeOf` closure copies `reasonOf` in `autoreconciler_test.go` | Follow-up | No fix plan touches the file; lift to `(*fixture).outcomeOf` next time it is edited |
| MNT-3 | MNT | minor | `observationstest` copies "stage, then write" from four FFI handlers | Follow-up | Moved code, not new; suggest an idea for a core stage-then-write entry point (thin handlers) |
| CON-3 | CON | nit | `TestReconcileDepictionMultiple` is three `t.Run`s, not a table | Logged | |
| MNT-4 | MNT | nit | Handler tests keep private PNG/likeness helpers | Logged | Plan Next adds defers this; ride with MNT-3 |
| MNT-5 | MNT | nit | Package doc line in `autoreconcile.go:22` not wrapped | Logged | |

No Fix-now findings, so there is no Fix plan 1 and no round 2: round 2 goes only to lenses with a Fix-now finding.

## Result
Gate PASS · CI <pending> · open: none · follow-ups: SEC-1 (S10-35 note), CON-1, CON-2, MNT-1, MNT-2, MNT-3 · to user: —
