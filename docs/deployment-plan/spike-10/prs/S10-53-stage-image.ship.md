# Ship log — S10-53 Stage a client image

**Plan:** [`S10-53-stage-image.md`](S10-53-stage-image.md) · **Branch:** `s10-53-stage-image` · **PR:** #424
**State:** ready

## Round 0 — implement
Base: branch `s10-53-stage-image` from `main` @ `1c99ef4e`, with the re-plan committed at `7c8cf08a`.
Head `e68f8c7b`. Gate PASS (orchestrator run). Report deviations:
- `objectstore.WriteIfAbsent` hashes `data` and refuses a mismatched checksum: D-6 said it would not hash, but T-3 requires the mismatch error (a contradiction in the plan, not the code). Landed note records it.
- `stageMaxEdge = 2048` constant instead of a literal.
- T-9 reuses the package's `testCatalog` / `runArtifactCreate` helpers.
- Removed code-span backticks from four Current state paths the PR moved or deleted, so `check-plan-docs.py` passes after landing; Landed note records it.
- `Discard` resets `wrote` after purging.
- In-app manual check (delete an Artifact, thumbnail leaves `objects/`) not run.

## Round 1 — review
Lenses: all five. CON, MNT, SEC, UX minors/nits; PERF clean. No blocker or major → no Fix plan 1, no round 2.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | minor | T-1's `fake` row is refused by `supportedFormats` before the `Formats` check, so nothing pins the check before a full decode | Follow-up | verified: `raster.go` `supportedFormats` has no `fake`; the plan's T-1 wording caused it; code is correct today |
| CON-2 | CON | minor | `TestEncodePNG`, `TestWriteIfAbsent`, `TestLookupByChecksumTx`, `TestStagedImageInsertTx` not table-driven | Follow-up | coverage complete; style rule only |
| CON-3 | CON | minor | §10 in-app manual check not run | To user (manual) | user runs it before merge |
| CON-4 | CON | nit | Landed note omits `Discard` resetting `wrote` | Logged | recorded in round 0 above |
| MNT-1 | MNT | minor | `LookupByChecksumTx` copies the trim and hex check | Follow-up | verified `files.go:91-115`; D-7 asked for reuse; a few lines |
| MNT-2 | MNT | minor | private `releaseClaim` doc names `ReleaseFacets` and the old callers | Follow-up | verified `facets.go:252-255`; comment only |
| MNT-3 | MNT | minor | PNG hashed twice; three inline SHA-256 helpers | Follow-up | `objectstore.Checksum` helper |
| MNT-4 | MNT | minor | image fixtures duplicated in `ingest` and `raster` tests | Follow-up | when S10-32 adds a third caller (`core/raster/rastertest`) |
| MNT-5 | MNT | nit | callers still name the change list `released` | Logged | — |
| MNT-6 | MNT | nit | `"image/png"` three times in `stage.go` | Logged | — |
| SEC-1 | SEC | minor | `WriteIfAbsent` follows symlinks in `objects/`; `RemoveObject` does not | Follow-up | behavior moved, not new (old `writeObjectIfAbsent`, `installObject`); fix through `os.Root` with S10-55 SEC-R2-1 |
| UX-1 | UX | minor | S10-54's map row does not own Swift copy for `ingest.unsupported_image` or depiction wording for reused ingest errors (512 MB cap, "choose that file") | Follow-up | verified: `grep unsupported_image macos/` empty; S10-54 map row |

## Result
Gate PASS · CI see PR #424 · open: none · follow-ups: CON-1, CON-2, MNT-1, MNT-2, MNT-3, MNT-4, SEC-1, UX-1 · to user: CON-3 (manual check) · logged: CON-4, MNT-5, MNT-6
