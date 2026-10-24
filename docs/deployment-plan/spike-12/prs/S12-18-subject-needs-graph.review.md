# Review — S12-18 A Subject needs its graph

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

Round 1 had one blocker (X-1), a one-clause fix to D-10, and no reshape, so no round 2 ran.

## To user
- M-1: fold the ~25 copied Subject-create test helpers into one shared fixture in this PR? Recommendation: keep S12-18 mechanical and log the fold. **Answered (user, 2026-10-24):** keep it mechanical; logged in [`ideas/test-subject-fixtures.md`](../../../ideas/test-subject-fixtures.md).

## Writer's own calls
- D-2: `requireSource` and `sqlSourceExists` are deleted, since `SourceOf` proves the graph exists, and so its Source.
- D-4: `SoleGraphTx` and `SoleGraph` stay exported; only the comment changes (S12-09 deletes both).
- D-5: graph reads run before the write, never inside a write closure, because the catalog has one connection (`SetMaxOpenConns(1)`).
- D-6: `artifactstest.SubjectOn` drops its dead `sourceID` parameter.
- D-7: `sources_test.go` uses `evidencegraphs.List` directly (import cycle with `evidencegraphstest`).
- D-10: the tests for the removed fallback and for a Source / graph mismatch are deleted; a wrong-length case is added.

## Round 1

### M-1 · major · D-5 adds a graph read to about 25 copies of `artifactstest.SubjectOn` instead of moving them onto it
About 25 test helpers copy `SubjectOn`'s body; D-5 adds a `SoleGraph` line to each. Better shape: helpers call `SubjectOn` (or a type-id sibling).
**Answer:** To user → user kept this PR mechanical (2026-10-24). Logged in `docs/ideas/test-subject-fixtures.md`; §3 Out and D-5 Rejected cite it.

### M-2 · major · Fixtures with one fixed Source should store the graph once
`batchFixture`, `bridgeWorld`, canonicalgraph `fixture`, conclusionheaders `fixture`, `handleFixture` hold one Source; S12-09 changes the Promote fixtures next.
**Answer:** Fixed — D-5 shape (a): those five fixtures gain `graph []byte`, set where they are built; Codebase fit row added.

### M-3 · minor · `SubjectOn` lives in `artifactstest`
**Answer:** Declined for this PR with M-1 (user's call); logged in the same idea file.

### M-4 · minor · Two placement writes in test fixtures (log it)
**Answer:** Logged in `docs/ideas/test-subject-fixtures.md` ("One placement write").

### X-1 · blocker · Deleting "another Source's id" leaves `other` unused
**Answer:** Fixed — D-10: `src := newSource("Deed")`; the `newSource` closure stays.

### X-2 · major · Unused helpers and imports after the case deletions
**Answer:** Fixed — D-5 and D-10: remove any import, helper or variable the change leaves unused, and nothing else.

### X-3 · major · STEP-2's check cannot see the `SubjectOn` callers
**Answer:** Fixed — STEP-2 changes the function only; STEP-3 moves the 10 callers and lists "too many arguments in a call to `artifactstest.SubjectOn`" as a second vet signal.

### X-4 · major · D-5 names `mustPerson` / `mk` as helpers with Source parameters
**Answer:** Fixed — D-5 now has three shapes: fixed-Source fixtures, helpers or closures that take or close over a Source id, and tests that already hold a graph. Reads go outside and before the write.

### X-5 · major · The handler-test graph read is not located
**Answer:** Fixed — STEP-4 names each `withProjectCatalog` closure, the line for the read, and the `evidencegraphstest` import.

### X-6 · minor · Literals that set neither id
**Answer:** Fixed — STEP-3: `go test` reports them with `subjects.invalid`; they are fixed the same way.

### X-7 · minor · `SoleGraphTx` comment placement
**Answer:** Fixed — D-4 quotes the whole comment.

### X-8 · minor · `completed.md` position and format
**Answer:** Fixed — STEP-5: top of `completed.md`, above S12-08, in its format.

### X-9 · minor · `sources_test.go` failure message
**Answer:** Fixed — D-7 gives the `t.Fatalf` message and the read's position.
