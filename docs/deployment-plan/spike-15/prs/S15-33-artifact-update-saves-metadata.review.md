# Review — S15-33 Editing an Artifact can save its metadata

**Verdict:** (after round 2)
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
None.

## Writer's own calls
- Version bump is a patch (0.11.1), following S15-26's response fields, not a minor.
- Both Artifact write replies always carry the board, draft or not (D-2).
- Only value refusals (blank, malformed for the type) name their field; repeats, unknown fields and a missing owner stay bare (D-3).
- A field id rides `Error.params` as a locator, written into the proto comment, `ErrInvalid`'s doc and `add-ffi-handler` (D-6), rather than a new `Error` field.
- The board is read after `writes.Run`, outside the write transaction, as `SetMetadata` does (D-4).

## Round 1

### X-1 · blocker · STEP-2 and STEP-4 checks cannot run in this environment
**Answer:** Declined — the plan's **Runs on** header already says STEP-2 needs `protoc` and its plugins and `check-pr.sh` needs Xcode, and to stop rather than hand-edit generated files; S15-26's plan carries the same line. Where the PR runs is the dispatcher's call (ship-pr), not a gap in the plan.

### X-2 · blocker · AC-3.5 "missing Artifact" cannot produce `sourcemetadata.invalid` through `UpdateArtifact`
**Answer:** Fixed — AC-3.5 now targets `sourcemetadata.Apply` and says `artifacts.Get` refuses a missing Artifact first, as today.

### X-3 · major · AC-1.1 and T-2 change order and layout actions are underspecified
**Answer:** Fixed — AC-1.1 seeds A by two `SetMetadata` calls (layout `provider` 0, `url` 1, from `layoutAppend`), gives the draft's values order and the exact `latestRevisionTypes` list, and says why `url` leaves the board.

### X-4 · minor · The rewritten `UpdateArtifactRequest` comment drops existing text
**Answer:** Fixed — §6 FFI keeps the `IngestArtifactFile` line; §4 quotes the real comment.

### X-5 · minor · The DoD brief line has no referent
**Answer:** Fixed — "n/a, no brief (S15-D14 moves with S15-34)".

### X-6 · minor · Test choices left open
**Answer:** Fixed — T-1 names `applyWorld`, `runApply`/`runCreateAll`, `errors.As` and `countRows`; `notes` dropped; T-6 names the page and graph; AC-1.5 and T-4 build the equal draft, layout included, from `GetMetadataBoard`.

### X-7 · minor · Status, review and base
**Answer:** Fixed — status moves to Planned after round 2; base unchanged.

### M-1 · major · A refused value should name its field in one place
**Answer:** Fixed — D-3 adds `normalizeFor`, called by `Set` and `apply`; AC-3.6, T-8, T-9; the Out row for `Set` is gone.

### M-2 · major · `Error.params` gains a second meaning
**Answer:** Fixed — D-6: the proto comment, `ErrInvalid`'s doc comment and one `add-ffi-handler` line; §5 says why the Delete-scoped line does not apply (the user-reviewed map scope chose the param).

### M-3 · major · Sixth and seventh copy of "read the owner's board after the write"
**Answer:** Fixed — `ownerBoardProto` in `metadata.go`; the five board handlers and both Artifact handlers use it (D-2, STEP-3, Codebase fit, Deletes).

### M-4 · major · Url test fields created inline again
**Answer:** Fixed in part — `newField(.., dataType)` with `newUserField` delegating; new tests use it. Declined moving the four existing inline creations in `sources_test.go`: they build requests inside `runRPC` table closures from `req`, so the move is churn in tests this PR does not touch (Out row).

### M-5 · minor · "Reply equals the board" written three times
**Answer:** Fixed — `assertReplyBoard` beside `draftOf`; §4 lists `metadataByField` and `metadataFieldIDs`.

### M-6 · minor · Artifact replies carry the board, graph replies do not
**Answer:** Logged — S15-11's scope entry now asks its planner to decide it; the Out row points there.
