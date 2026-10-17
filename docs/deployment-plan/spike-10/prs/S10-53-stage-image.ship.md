# Ship log — S10-53 Stage a client image

**Plan:** [`S10-53-stage-image.md`](S10-53-stage-image.md) · **Branch:** `s10-53-stage-image` · **PR:** #424
**State:** review 1

## Round 0 — implement
Base: branch `s10-53-stage-image` from `main` @ `1c99ef4e`, with the re-plan committed at `7c8cf08a`.
Head `e68f8c7b`. Gate PASS (orchestrator run). Report deviations:
- `objectstore.WriteIfAbsent` hashes `data` and refuses a mismatched checksum: D-6 said it would not hash, but T-3 requires the mismatch error (a contradiction in the plan, not the code). Landed note records it.
- `stageMaxEdge = 2048` constant instead of a literal.
- T-9 reuses the package's `testCatalog` / `runArtifactCreate` helpers.
- Removed code-span backticks from four Current state paths the PR moved or deleted, so `check-plan-docs.py` passes after landing; Landed note records it.
- `Discard` resets `wrote` after purging.
- In-app manual check (delete an Artifact, thumbnail leaves `objects/`) not run.
