# Review — S14-12 File work before the catalog session

**Verdict:** Round 1: Fits with changes — Not ready → Fixed. Round 2 pending.
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- None open. M-1 reversed one of the plan's own Out lines (keeping `NormalizeCover` generating). That line was the writer's call, not a user answer, so the writer fixed it. The user can still veto it: the cost is one setup line each in three core tests (`sources_test.go` `runSetCover`, `artifacts_test.go` line 626, `sourcecopy_test.go` line 174). Recommendation: keep the fix.

## Writer's own calls

- D-3: every stage, including depiction PNGs and thumbnails, goes through a temp file in the objects folder, so only a rename runs inside the session.
- D-10: thumbnails are staged and committed in batches of 16. This bounds open handles on the objects folder. A Source page with N missing thumbnails takes ⌈N/16⌉ extra short sessions.
- D-10 / D-8: a source object missing between the two sessions fails that one call, as a missing object does today. A deleted File row is `ErrInvalid` and gives an empty path.
- D-13: `IngestArtifactFile` stages before checking the Artifact. When a request has two faults, the file error now wins, and an already-attached Artifact pays the copy before the refusal.
- Out: a refused decode is not remembered (PRF-1's repeat cost stays, off the session).
- Out: `ingest.File` and `derivatives.Ensure` stay in place as test-and-fixture compositions with doc comments and a skill "Do not". They were not moved to test packages (M-3).

## Round 1

### M-1 · blocker · The cover keeps a second copy of the thumbnail rule, and its in-session decode stays
`requireRasterCoverArtifact` still ensured (decoded) inside `Do` and mapped the three errors. The handler pre-check added a third place deciding "cover refused when no thumbnail".
**Answer:** Fixed. D-12 adds `sources.CoverArtifactFile`. `requireRasterCoverArtifact` becomes lookup-only (`filederivatives.Lookup`), `sources` stops importing `derivatives`, and `SetSourceCover` plans the cover File in a first session, with `NormalizeCover` deciding alone in the pass's commit session. D-9 adds `derivatives.Refused` as the one error mapping. The three core tests named above call `EnsureThumbnail` first (STEP-5); T-12 is new. The Out line is replaced.

### M-2 · blocker · `thumbnailPass.run` hand-rolls stage, commit and discard beside `withStaged`
**Answer:** Fixed. D-10: each batch runs through `withStaged`. D-7 lists `thumbnailPass` among its users, so no code outside `withStaged` calls `Purge` or `Discard`.

### M-3 · major · `ingest.File` and `derivatives.Ensure` stay as in-session entry points with no handler caller
**Answer:** Fixed (the reviewer's minimum option). D-5 and D-8 give both a doc comment marking them test-and-fixture only. STEP-7 adds both to `use-catalog-session` "Do not", and STEP-5's check greps production code for `Ensure` callers. Moving them to helper packages is an Out line (about 35 call sites of churn).

### M-4 · major · `Staged.Close` / `Staged.Discard` reuse `objectstore.Temp` names with other meanings
**Answer:** Fixed. D-6: `Staged.Discard()` runs outside the session and removes only the temp, matching `Temp.Discard`. `Staged.Purge(c)` runs inside the session, matching `files.PurgeUnreferenced`. `Stages` follows.

### M-5 · major · Five tests hand-build "hold the session on a channel"
**Answer:** Fixed. D-15 adds `catalogsessiontest.Hold` (STEP-2), and T-1, T-3, T-4 and T-8 use it. I checked with `go list -deps ./core/catalogsession` that it includes neither `ingest`, `derivatives` nor `sources`.

### M-6 · minor · Staging copies the session's folder normalization
**Answer:** Fixed. D-7: `catalogsession.key` is exported as `Key`. `withStaged` takes a `stage func(dir string)` and normalizes the folder once.

### M-7 · minor · `derivatives` → `ingest` pulls `observations` under `derivatives`
**Answer:** Logged. §6 Edges records the dependency. With M-1's fix, `sources` no longer imports `derivatives`, which removes the `sources` → `derivatives` → `ingest` → `observations` chain the finding worried about. The stop condition guards both `go list` directions.

### M-8 · nit · `Commit(c, job, staged)` takes two halves of one thing
**Answer:** Fixed. D-8: `Job.Stage` keeps its stage on the `Job`, and `Commit(c, j)`.

### X-1 · blocker · `Commit` via `Staged.InsertTx` changes thumbnail audit
I confirmed this in `core/writes/run.go`: a non-none change records a revision.
**Answer:** Fixed. D-8: `Commit` drops the `file` change `InsertTx` returns. T-5 asserts no revision, and T-7 asserts the revision count is unchanged by a Source page read.

### X-2 · blocker · `SetSourceCover` first session unspecified
**Answer:** Fixed with M-1. D-12 spells out each branch. Artifact mode with a parsed id: `CoverArtifactFile` (its `ErrInvalid` is returned for a missing, foreign or file-less Artifact), then `need`. Any other mode skips the first session. `NormalizeCover` in `then` decides. T-9 rows cover PDF, a late-failing JPEG, another Source's Artifact, and type-icon mode.

### X-3 · major · `need` has no place for `Skipped`; the existing-link path is undefined
**Answer:** Fixed. D-10: `need` takes `fill func(relPath string, skipped bool)`, and `derivedRelPath` holds `thumbnailRelPath`'s tail. D-11: `EnsureFileThumbnail`'s fill sets both fields. T-10 adds a PDF row.

### X-4 · major · Error order in `IngestArtifactFile`
**Answer:** Fixed by documenting it. D-13 accepts the two-fault precedence change and the copy on an attached Artifact, with its reason. AC-1.4 and T-8's attached row pin it. `ErrNotAProject` from `RequireProject` is the same error `Do` returns today.

### X-5 · major · STEP-8 sweep check cannot pass
**Answer:** Fixed. STEP-7's check greps `.agents docs/*.md`, so plan history under `docs/deployment-plan/` keeps its names. This matches the DoD sweep.

### X-6 · major · Steps 2–4 leave `api/` not compiling
**Answer:** Fixed. STEP-2 now moves the depiction handlers and `observationstest` with the type change. `IngestArtifactFile` keeps compiling through `ingest.File` until STEP-3, and `thumbnailRelPath` through `Ensure` until STEP-5. Every step's check is `go test ./...`, and §7 states that every step builds.

### X-7 · major · `Staged` behaviors left to design
**Answer:** Fixed. D-5: `CommitFile` reads every field from the stage; a stage without a filename → `ErrInvalid` (T-2); the reuse path's install repairs the object or removes the temp. D-6: `Discard` is a no-op after `Install`. D-10: temp lifetime is bounded by batches of 16.

### X-8 · major · `derivatives.Commit` details
**Answer:** Fixed. D-8: `Commit` keeps the `errDerivativeLinkExists` → `Lookup` fallback. The caller purges, not `Commit`. The source File re-check is `files.Lookup` outside any tx, before the Run, and compares checksums. `Job` fields are listed.

### X-9 · minor · Test rows not writable from the row alone
**Answer:** Fixed. D-15 names the hold helper and the poll. T-7 names the fixture builder (`sourceWithPNG`'s shape) and replaces "no second session" with observable counts. T-2 names the scaffolding (`writes.Run` with a failing fn). T-10 gains a PDF row.

### X-10 · minor · Undefined edges and leftovers
**Answer:** Fixed. D-10: dedup per File id, and the behavior of a stage or commit failing mid-batch (batch discarded and purged by `withStaged`, call fails). The pass runs only when the first session returned nil (D-11). The DoD brief line is marked not applicable. STEP-2 now also updates the `observations.go` line 177 and `depictionvalue.go` line 33 comments.
