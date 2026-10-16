# Review — S10-53 Stage a client image

**Verdict:** Fits — Ready with changes (all applied).
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None.

## Writer's own calls

- D-1: staging lives in `core/ingest`, not `core/derivatives`, because a depiction image is not a disposable derivative.
- D-3: the staged-file cap is 32 MiB.
- D-4: the codec moves to `core/raster`.
- D-8: only PNG and JPEG are accepted for staging.
- D-9: Go downscales anything over 2048 on the long edge.
- S10-53 now depends on S10-51, so the two run in sequence.

## Round 1

### M-1 · blocker · `Discard` copies S10-51's purge
**Answer:** Fixed. D-11 calls `files.PurgeUnreferenced`, and the header and map now depend on S10-51.

### M-2 · major · The tx checksum lookup is inline SQL in `ingest`
**Answer:** Fixed. D-7 adds `files.LookupByChecksumTx` (`add-catalog-query`). The inline `lookupFile` by id is logged in Out.

### M-3 · major · The `file` create change would be copied
**Answer:** Fixed. D-2 extracts `insertFile`, which both callers use.

### M-4 · major · A third byte write into the object store
**Answer:** Fixed. D-6 moves `writeObjectIfAbsent` and `syncDir` into `core/objectstore`. `installObject` stays in `ingest` (Out).

### M-5 · major · The shared codec stays under `core/derivatives`
**Answer:** Fixed. D-4 moves it to `core/raster` and rewrites its doc, and STEP-7 updates the security lens trust table.

### M-6 · major · `Options.Formats` is honored by one encoder; the preamble is copied
**Answer:** Fixed. D-4 adds private `decode` and `prepare`, and both encoders honor `Formats`.

### M-7 · minor · The import media policy is skipped without a decision
**Answer:** Fixed. D-8 records it, and T-4 checks that `stageFormats` stays inside `mediatypes.Allowed`.

### M-8 · minor · The cap breaks the const-plus-override pattern
**Answer:** Fixed. D-3 adds `imageMaxBytes` and `maxImageBytes`, both unexported. The shared `LimitReader` helper is logged in Out.

### M-9 · minor · One calling shape for S10-32
**Answer:** Fixed. D-11 makes `Discard` safe to call unconditionally (`defer s.Discard()`), and T-8 has the double-call case.

### X-1 · blocker · The checksum lookup on a Tx
**Answer:** Fixed by D-7.

### X-2 · major · How the PNG reaches disk
**Answer:** Fixed by D-6 and D-9 (`objectstore.WriteIfAbsent`).

### X-3 · major · `StagedImage` fields and repeat calls
**Answer:** Fixed. D-10 lists the fields and the second-call behavior, and D-11 covers the second-call behavior of `Discard`.

### X-4 · major · Where the slot is taken
**Answer:** Fixed. D-5 takes it after `openChecked` and releases it after `EncodePNG`.

### X-5 · major · Test helpers are closures in `TestFile`
**Answer:** Fixed. STEP-5 adds package-level helpers in `stage_test.go` and leaves `file_test.go` unedited.

### X-6 · minor · The change fields are inline
**Answer:** Fixed by D-2 (`insertFile`).

### X-7 · minor · Package doc: replace all or part
**Answer:** Fixed. D-1 keeps the first two sentences and gives the replacement.

### X-8 · minor · Doc placement and the "original bytes" sentence
**Answer:** Fixed. STEP-7 qualifies "original file bytes" and places the `completed.md` section above the current first one.

### X-9 · minor · Two `ErrTooLarge`s
**Answer:** Fixed. D-12 has tests use `errors.Is` and keeps the two errors distinct.

### X-10 · nit · Fixtures
**Answer:** Fixed. The test cap override replaces the 32 MiB fixture, and AC-1.2 and STEP-5 require real TIFF, GIF, and BMP encodings.

## Round 2

### X-11 · blocker · STEP-7's grep cannot pass, and an idea file's link breaks
**Answer:** Fixed. STEP-7 adds `docs/ideas/large-jpeg-thumbnails.md` with the new path and the ImageIO note (M-10), and the grep is narrowed to `core`, `.agents`, and `docs/ideas`.

### X-12 · major · Error precedence in STEP-1
**Answer:** Fixed. D-2 keeps `c.DB()`, the trim, and `RequireUserID` in `File`, in today's order, and D-9 has `StageImage` call `c.DB()` first.

### X-13 · major · A failure after the object write leaves an orphan
**Answer:** Fixed. D-9 has `StageImage` discard its own object on any error after `WriteIfAbsent` reports `wrote`.

### X-14 · minor · `WriteIfAbsent` details
**Answer:** Fixed in D-6: the `.write-*` temp prefix, its own `sha256Hex`, and the `SyncDir` error ignored as today. M-14's package doc change is in the same decision.

### X-15 · minor · T-7 and T-8 setup
**Answer:** Fixed. The `Op.Action` is named, T-7 reads the changes from `InsertTx`'s return, T-8 makes the existing File by staging and committing, and D-10 uses `files.NewID()`'s two returns.

### X-16 · minor · STEP-2 and D-4 wording
**Answer:** Fixed. `raster_test.go` moves unchanged and is then extended, and D-4 keeps the security paragraph.

### X-17 · minor · Two concurrent stages
**Answer:** Accepted as a Risks row. `catalogsession.Do` serializes a project's operations.

### M-10 · major · Grep and idea link
**Answer:** Fixed with X-11.

### M-11 · major · The codec leaves the security routing table
**Answer:** Fixed. STEP-7 adds `core/raster` to the Security row of the ship-pr routing table.

### M-12 · minor · T-4 restates the rule instead of enforcing it
**Answer:** Fixed. T-4 loops over `stageFormats`.

### M-13 · minor · `ingest.File`'s rollback is a second remover
**Answer:** Fixed. D-2 moves it onto `files.PurgeUnreferenced`, and S10-51 D-1's doc line now names the purge as the only remover.

### M-14 · nit · `objectstore` package doc
**Answer:** Fixed in D-6.

## After S10-32's plan

### From S10-32 · `InsertTx` takes a `*sql.Tx`
**Answer:** Fixed. S10-32 inserts the staged File from `observations.resolveValue`, which holds a `*sql.Tx` (as does every Observation write below the handler). D-2's `insertFile` and D-10's `InsertTx` now take `*sql.Tx`; `ingest.File` passes `tx.Tx`. Nothing in `ingest` needed the wrapper.
