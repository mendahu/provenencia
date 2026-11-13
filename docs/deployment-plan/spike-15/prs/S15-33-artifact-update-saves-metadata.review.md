# Review — S15-33 Editing an Artifact can save its metadata

**Verdict:** Fits with changes — Ready with changes
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

## Round 2

Executability (Sonnet) accepted X-1 – X-5, X-7 and M-1 – M-6; X-6 partly, with the test recipes below.

### X-9 · blocker · T-1 asserts "no params" on the blank-value case D-3 gives a param
**Answer:** Fixed — T-1's loop skips "blank value", which its first row covers with `[doc id]`.

### X-10 · major · AC-2.2 and T-7 name different suggested fields
**Answer:** Fixed — T-7 adds a `newField(.., "url")` assigned at the `artifact` level; `provider` stays the drafted text field.

### X-11 · major · The equal-draft recipe can write layout rows
**Answer:** Fixed — AC-1.5 and T-4 name A's stored rows (`provider` by one `SetMetadata`) and list a layout row only for that field.

### X-12 · minor · AC-1.1's label does not match the fixture
**Answer:** Fixed — `page1`, label "Scan" → "Scan (front)".

### X-13 · minor · Wording and placement left open
**Answer:** Fixed — D-6 gives the `add-ffi-handler` sentence; T-9 is a new subtest on `m.artifact`; AC-1.1's draft is built by hand.

Maintainability (Opus) accepted X-1 – X-7, M-1, M-5, M-6; challenged M-2, M-3 and M-4's answers.

### M-8 · blocker · T-1 asserts "no params" on the blank-value case
**Answer:** Fixed — same as X-9.

### M-2 (round 2) · major · `core/apperr`'s doc comments still say params are L10n args
**Answer:** Fixed — STEP-1 rewords `apperr.Error`'s and `apperr.New`'s comments to D-6's wording.

### M-3 (round 2) · major · A sixth inline board read in `GetSourceWorkspace`
**Answer:** Fixed — `sources.go` `GetSourceWorkspace`'s Source board moves onto `ownerBoardProto`; `artifactBoardsProto` named as the batch form that stays.

### M-4 (round 2) · major · The decline's premise was wrong
**Answer:** Fixed — the four inline url-field creations in `sources_test.go` have `t` and `req` in scope; STEP-3 moves them onto `newField`; the Out row is gone.

Optional (M-1, round 2): change `normalizeValue`'s signature instead of adding `normalizeFor`. Logged: `normalizeFor` wraps it, and both stay unexported in one file.
