# Review — S10-53 Stage a client image

**Verdict:** Fits — Ready with changes (applied); refresh after S10-55 below, same verdict.
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None.

## Writer's own calls

- D-1: staging lives in `core/ingest`, not `core/derivatives`, because a depiction image is not a disposable derivative.
- D-3: the staged-file cap is 32 MiB.
- D-4: the codec moves to `core/raster`.
- D-8: only PNG and JPEG are accepted for staging.
- D-9: Go downscales anything over 2048 on the long edge.
- S10-53 now depends on S10-55, so the two run in sequence.

## Round 1

### M-1 · blocker · `Discard` copies S10-55's purge
**Answer:** Fixed. D-11 calls `files.PurgeUnreferenced`, and the header and map now depend on S10-55.

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
**Answer:** Fixed. D-2 moves it onto `files.PurgeUnreferenced`, and S10-55 D-1's doc line now names the purge as the only remover.

### M-14 · nit · `objectstore` package doc
**Answer:** Fixed in D-6.

## After S10-32's plan

### From S10-32 · `InsertTx` takes a `*sql.Tx`
**Answer:** Fixed. S10-32 inserts the staged File from `observations.resolveValue`, which holds a `*sql.Tx` (as does every Observation write below the handler). D-2's `insertFile` and D-10's `InsertTx` now take `*sql.Tx`; `ingest.File` passes `tx.Tx`. Nothing in `ingest` needed the wrapper.

## Refresh after S10-55 (2026-10-16)

Re-planned against `main` @ `1c99ef4e` with S10-55 landed, folding in its two review requirements (MNT-1, SEC-O2). Fresh reviewers: maintainability (opus), executability (sonnet).

### To user (answered)

- **M-1 · How the delete/import race is closed.** The writer's draft held the catalog's one connection across the purge's check and removal and moved `ingest.File`'s and `derivatives.Ensure`'s object writes into their writes, with a test hook. The reviewer showed `catalogsession.Do` already serializes every production object write and purge. **User (2026-10-16):** rely on `catalogsession.Do`; document it as the guarantee; drop the reorders, the hook, and its test; keep or drop the connection hold at the writer's call. **Writer:** dropped it too (D-14 is now the doc comment only), since a second serializer beside the declared one is what M-1 objects to.

### Writer's own calls

- D-13: every exported `deleteimpact` release takes `*database.Tx` and purges; `Released` becomes unexported, so callers cannot drop objects. `UnfileBridgesTx` is renamed `UnfileBridges` (M-6).
- D-15: `StageImage` writes nothing; `InsertTx` writes the object inside the caller's write, so a stage never inserted leaves nothing. This replaces the first plan's write-in-`StageImage`.
- `WriteIfAbsent` takes the checksum (M-4).

### Round 1 — maintainability

#### M-1 · blocker · A second serializer beside `catalogsession`
**Answer:** To user → answered above. D-14 is the purge doc comment; D-15 keeps only staging's in-write placement; T-10/T-11 of the draft, the hook, and the draft STEP-5/STEP-6 are gone; a Risks row, a Stop condition, AC-5.1/5.2, and T-10 (`TestDo`) name the guarantee.

#### M-2 · major · `*sql.Conn` brings a context-taking querier
**Answer:** Moot with M-1 — D-7 is back to `LookupByChecksumTx` alone.

#### M-3 · major · `SetMaxOpenConns(1)` becomes load-bearing undocumented
**Answer:** Moot with M-1 — nothing relies on it now.

#### M-4 · major · `WriteIfAbsent` should take the checksum
**Answer:** Fixed in D-6, D-10, T-3.

#### M-5 · minor · `Ensure`'s pre-write lookup and retry
**Answer:** Logged as Out (M-1 leaves `Ensure`'s order unchanged).

#### M-6 · minor · `UnfileBridgesTx` with `*database.Tx` breaks the suffix convention
**Answer:** Fixed in D-13 and STEP-7: renamed `UnfileBridges`.

#### M-7 · minor · Next adds omits S10-32's signature change
**Answer:** Fixed in the S10-32 Out row: `updateOne` and `releaseReplacedValue` take `*database.Tx`.

#### M-8 · minor · `raster.Slot` is a caller pairing
**Answer:** Declined, with the reason in D-5: the raw read is part of the memory the slot bounds, so it cannot move inside `prepare`.

### Round 1 — executability

#### X-1 · blocker · Map row false
**Answer:** Fixed. The Map row says the working rule and entry were refreshed with this plan; STEP-8 changes only the status and SLICE 8 lines.

#### X-2 · major · `register.go` holds no `Released`
**Answer:** Fixed. The row names the `facets.go` symbols; `register.go` left the file list.

#### X-3 · blocker · `ReleaseOwned` has no unexported body
**Answer:** Fixed. D-13 adds `releaseOwned`; STEP-7 says `release_test.go` near line 348 calls it.

#### X-4 · major · External test move to `BeginWrite` under-specified
**Answer:** Fixed. STEP-7 names the four transactions, says the rest stay, and says to route every query through `tx`.

#### X-5 · major · `EncodeJPEG` error order
**Answer:** Fixed. D-4 adds `checkMaxEdge`; order is MaxEdge, Quality, decode.

#### X-6 · major · D-2 contradicts itself on `RequireUserID`
**Answer:** Fixed. `openChecked` starts at the absolute-path test; `File` calls `RequireUserID` first; both are `ErrInvalid`.

#### X-7 · major · `installObject` failure inside the closure
**Answer:** Moot with M-1 — `installObject` stays before the write.

#### X-8 · minor · `StageImage` details
**Answer:** Fixed in D-9: `defer release()`, read errors returned as is, `ErrTooLarge` with params.

#### X-9 · minor · "No full decode" unobservable
**Answer:** Fixed. AC-1.2 drops the claim; T-1 uses the registered `fake` format to show no full decode ran.

#### X-10 · minor · Connection-count stop condition uncheckable
**Answer:** Moot with M-1.

#### X-11 · minor · STEP-8 doc edits under-specified
**Answer:** Fixed. The sketch's `released.Changes`, the idea file's insertion point, and the missing Scope and DoD lines.

#### X-12 · minor · Fixture choices
**Answer:** Fixed where the test remains; T-11 is gone with M-1.

#### X-13 · nit · STEP-2 listing
**Answer:** Fixed.

### Round 2 — maintainability (verdict: Fits)

Every round-1 answer confirmed; no new blockers. Notes below the bar:
- STEP-7's four test transactions call `BeginWrite` directly instead of `writes.Call`. **Declined:** those tests assert rows inside an uncommitted transaction and some commit raw; `writes.Call` rolls back an empty change list, so each would need a dummy change. Test-only.
- Performance routing should include `core/raster`. **Fixed** in STEP-8.
- `security.md` cites a nonexistent `raster.Thumbnail`. **Fixed** in STEP-8.
- `ReleaseOwned` has no production caller. **Logged;** kept as the documented while-the-parent-exists form.

### Round 2 — executability (verdict: Ready with changes)

#### X-R2-1 · blocker · Duplicate STEP-5 / STEP-6
**Answer:** Fixed. A renumbering slip; steps are STEP-1 to STEP-8 and every reference matches.

#### X-R2-2 · minor · T-9 fixture
**Answer:** Fixed. T-9 names `database.Create`, `files.Insert`, `objectstore.WriteIfAbsent`, and `artifacts.Create` after a Source.

#### X-R2-3 · nit · `Unpromote` wording
**Answer:** Fixed in D-13.

**Final verdict:** Fits — Ready with changes (applied). No open blocker; To user answered. Plan → Planned.
