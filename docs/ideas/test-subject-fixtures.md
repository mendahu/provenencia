# Test Subject fixtures

**Status:** idea only — not roadmapped. Found by the S12-18 plan review (M-1, M-3, M-4); the user kept S12-18 mechanical (2026-10-24).

## One way to make a test Subject

About 25 Go test helpers copy one body: look up a provenencia Subject type, run `writes.Call` with `create_subject`, call `subjects.Create`. Examples: `(*fixture).bare` in `core/database/promotealign/promotealign_test.go` and `core/database/conclusionheaders/walks_test.go`, `(*batchFixture).bare` and `(*bridgeWorld).node` in `core/database/promote/`, `(*walkFixture).subject` in `core/database/graphcache/evidence_graph_walk_test.go`, `mustSubject` in `core/database/identityclaims/identityclaims_test.go`, and the `mk` closures in `core/database/effects/`. `artifactstest.SubjectOn` already does the job. Move it to a `core/database/subjects/subjectstest` package (it has nothing to do with Artifacts), add a sibling that takes a type id, and point the copies at it. Then the next `subjects.CreateInput` rule is one edit, not a 40-file sweep.

## One placement write

Several of those helpers call `subjects.Create` with no placement and then run a second write with `subjectpositions.Set`. `subjects.Create` takes a `*subjects.Placement`; the shared fixture can take one and the second write goes.
