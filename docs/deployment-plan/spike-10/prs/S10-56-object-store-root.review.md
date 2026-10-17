# Review — S10-56 Object store stays inside objects/; deletes prune empty folders

**Verdict:** Fits — Ready with changes (applied)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None.

## Writer's own calls

- S10-56 is stacked on #424's branch, planned against its head (`39ee4860`), at the user's request to do it now; a Stop condition guards #424 changing first.
- `ingest.installObject` moves into `core/objectstore` (reverses S10-53's Out line), behind a `Temp` handle.
- An unsafe source object or objects/ folder maps to `derivatives.ErrCorruptObject`, so the thumbnail handler skips it.
- Empty folders already in existing projects are left alone.

## Round 1

### Maintainability (Fits with changes)

#### M-1 · major · `WriteIfAbsent` and `Install` would copy one install sequence
**Answer:** Fixed. D-4 adds private `installTemp`; `WriteIfAbsent` (D-5) and `Temp.Install` (D-6) both use it. Temps live at the top of objects/, so `createTempIn` lost its `dir` parameter and the "leftover temp keeps a shard alive" risk is gone.

#### M-2 · major · `MatchesChecksum` stays exported with no outside caller
**Answer:** Fixed. D-3 makes it private `matches(root, inner, checksum)`; the exported form is deleted in STEP-3. T-3 merged into T-4.

#### M-3 · minor · Temp API leaks a string protocol and reuses `Install`
**Answer:** Fixed. D-6 adds `CreateTemp` returning `*objectstore.Temp` with `Write`, `Close`, `Install`, `Discard`; `ingest.File` defers `Discard`. T-6 merged into T-5.

#### M-4 · minor · New derivatives test file rests on a false premise
**Answer:** Fixed. T-7 is a row of `TestEnsureThumbnailRejectsUnsafeSources`; symlink rows stay in the portable `objectstore_test.go`; only FIFO cases use `objectstore_unix_test.go`. AC-3.1 allows new rows.

#### M-5 · minor · A third spelling of the objects folder name
**Answer:** Fixed. D-1 exports `objectpath.Dir`; `objectpath.Rel` and `objectstore` use it. `catalog.go` stays (Out).

### Executability (Ready with changes)

#### X-1 · blocker · STEP-6 grep matches test fixtures
**Answer:** Fixed. The grep names only the non-test files.

#### X-2 · major · `createTempIn` return and `dir` form unstated
**Answer:** Fixed in D-4 (root-relative name, never `File.Name()`); `dir` is gone (M-1).

#### X-3 · major · `Install` leaks the temp when the root cannot open
**Answer:** Fixed by M-3: the `Temp` holds its open root, so `Install` and `Discard` always reach it.

#### X-4 · major · STEP-1's interim `installObject` signature
**Answer:** Fixed. `MatchesChecksum` keeps its absolute-path form until STEP-3, so STEP-1 does not touch `ingest`.

#### X-5 · minor · STEP-1 lists sync files it does not touch
**Answer:** Fixed.

#### X-6 · minor · `TestWriteIfAbsent` rewrite and `OpenObject` symlinked-objects row
**Answer:** Fixed in STEP-2 and T-4.

#### X-7 · minor · T-8 names `writes.Call`; the file uses `runDelete`
**Answer:** Fixed in T-8.

#### X-8 · minor · `readSourceObject` on a symlinked objects/
**Answer:** Fixed. D-2 adds `ErrUnsafeObjects`; D-10 maps it to `ErrCorruptObject`.

#### X-9 · nit
**Answer:** Logged.

## Round 2

### Executability (Ready with changes)

#### X-R2-1 · minor · `objectPath` is a closure
**Answer:** Fixed. T-8 re-declares the closure in the new test.

#### X-R2-2 · minor · `Temp` close bookkeeping
**Answer:** Fixed in D-6: a `closed` flag; `Close` on a closed file returns nil.

#### X-R2-3, X-R2-4 · nit
**Answer:** Logged.

### Maintainability (Fits)

Every round-1 answer confirmed; no new blockers. Notes below the bar, all applied:
- M-4 was only partly applied: T-4's non-FIFO rows now sit in the portable `objectstore_test.go`; only `TestOpenObjectFIFO` is unix-only.
- T-8 is a `t.Run` inside `TestArtifactDeletePurgesDisk`, reusing its setup and `objectPath` closure.
- `Temp.Close`'s doc says `Install` or `Discard` releases the handle.

**Final verdict:** Fits — Ready with changes (applied). No open blocker; nothing to user. Plan → Planned.
