# Review — S10-62 Refuse symlinked objects in the app

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus, pr-reviewer), executability (sonnet, Explore)

## To user
- None.

## Writer's own calls
- D-1: the check lives inside `ProjectFiles.objectURL`; signature and `URL?` result unchanged, so every reader gains it and no weaker entry point remains.
- D-2: `relPath` must start with `objects` and have at least one more segment (every Go producer goes through `objectpath.Rel` or `files.StorageRelPath`).
- D-3: Darwin `lstat` on each component from `projectDir/objects` down; folders must be real folders, the object a regular file; the project folder path itself is not checked (Go's `os.OpenRoot` allows it; macOS temp folders sit behind `/var` → `/private/var`).
- D-4: check, then read by URL; no descriptor reads (static prepared folder is the threat; `NSWorkspace` cannot open a descriptor; PDFs would be held in memory).
- D-5: the `lstat` calls run where `objectURL` is called today, main actor included.
- D-6: the `fileExists` checks in `openObject` and `ArtifactViewerModel.load` are deleted.
- D-7: no new copy or state; a refused object shows the existing missing state.
- D-8 (changed in round 1): no version bump.
- D-9 (round 1): tests use `ArtifactFileFixtures` plus one new `linkToOutsidePNG(at:)` fixture.
- D-10 (round 1): `objectsDir` pinned to `objectpath.Dir` by a Go test; `openRegular` names the Swift mirror.

## Round 1

### X-1 · blocker · The version bump contradicts the other plans in the same spike
D-8 bumped to 0.2.1 while S10-60, S10-61 and Spike 11 leave `VERSION` at 0.2.0.
**Answer:** Fixed — header, D-8, STEP-4 and DoD now say none. Evidence: `git log -- VERSION` last moved in #402; the sibling plans and every Spike 11 plan say none. The `MARKETING_VERSION` count no longer matters.

### X-2 · major · The `isSafeObject` signature and sketch cannot be applied as written
Two signatures; no `segments` array or `objectsDir` constant in the code; `lstat` call shape and depth rule unstated.
**Answer:** Fixed — §6 Codebase fit and macOS now name one signature `isSafeObject(segments:projectURL:)`; the existing loop appends to `var segments` (empty parts skipped); `private static let objectsDir = "objects"`; the sketch shows the `stat()` / `lstat(url.path, &st)` / `S_IFMT` walk; any folder depth is allowed and each folder is checked.

### X-3 · major · The test-helper instruction contradicts the `add-swift-test` skill
The plan said to write files with suite-private `FileManager` calls and the inline PNG literal; the skill requires `ArtifactFileFixtures`.
**Answer:** Fixed — D-9 and the §8 preamble: `uniqueProjectDir()` + `writePNG`, one new `ArtifactFileFixtures.linkToOutsidePNG(at:)` that makes and returns the outside folder, removed in a `defer`.

### X-4 · major · A Current state row is not true: the Go tests to mirror
The named subtests belong to `TestWriteIfAbsent` (write path); the read refusals are `TestOpenObject`.
**Answer:** Fixed — §4 row now cites `TestOpenObject` cases "symlink to an outside file", "missing object", "object path is a folder", "objects folder symlinked outside", "objects/ab symlinked outside", and says why `TestWriteIfAbsent` is not mirrored.

### X-5 · major · None of the Step checks can run off macOS
**Answer:** Fixed — Stop condition added: no macOS host with Xcode → stop (`scripts/check-pr.sh` skips Xcode tests off macOS, line 99). STEP-3's check now also runs `CitationComposerDepictionTests` and `PersonDetailTests`, and STEP-1 says the old join tests fail until STEP-2.

### X-6 · minor · The Status field contradicts the Implementer block
**Answer:** Declined — `plan-pr` §7 and `review-plan` keep a plan `Draft` until review passes; the writer sets `Planned` after this log, and `ship-pr` refuses a non-Planned plan. Nothing reaches an implementer as Draft.

### X-7 · minor · The STEP-4 docs edit is under-specified
**Answer:** Fixed — STEP-4 gives the full replacement sentence for `docs/source-layer-data-model.md` (keeping the shard-folder clause and the following depiction sentence) and the exact insertion point in `docs/application-stack.md` **Read path**.

### X-8 · minor · The Landed lines need a PR number and a merge date that do not exist yet
**Answer:** Declined — `.agents/skills/implement-pr/SKILL.md` step 8 already says these lines are written after the draft PR opens, in one "Record S#-NN as landed…" commit. STEP-4 now points there.

### X-9 · minor · AC-1.6 names a file that does not exist
**Answer:** Fixed — AC-1.6 uses `provenencia.sqlite`.

### X-10 · minor · Other small gaps
FIFO hang guard; T-11 spans two files; wrong-reason pass; `openObject` doc comment.
**Answer:** Fixed — T-6 carries `.timeLimit(.minutes(1))`; STEP-3 names `ProjectFilesTests.swift` for the `openObject` half of T-11; the §8 preamble requires a regular-file control in every link test; STEP-3 says the `openObject` comment stays as is.

### M-1 · major · The test setup re-solves what `ArtifactFileFixtures` already does
**Answer:** Fixed — same change as X-3 (D-9); the `ArtifactViewerModelTests` private helpers are left alone (no tidy outside the plan).

### M-2 · major · The mirrored Go value `objects` has no drift guard
**Answer:** Fixed — D-10, T-13 `TestSwiftObjectsDirMatchesGo` in `api/ffi/swift_vocabulary_test.go`, and one sentence in the `openRegular` doc comment. The Out line now says no Go behavior change.

### M-3 · major · The guidance the next reader and reviewer use does not name the new guard
**Answer:** Fixed — STEP-4 edits the security lens Object store row, adds `ProjectFiles.swift` and `ThumbnailCache.swift` to the ship-pr Security routing paths, and adds a `macos-client.mdc` bullet; DoD line added.

### M-4 · major · The kept lexical tests stop proving the lexical rules
**Answer:** Fixed — T-12 rewrites the three lexical tests onto a real file (`objects/./ab/cd/x`, `objects/~/x`, `objects/ab/../ab/cd/x` → `nil`, control `objects/ab/cd/x` → URL).

### M-5 · minor · The helper's name and signature disagree
**Answer:** Fixed — with X-2: `isSafeObject(segments:projectURL:)` in both places.

### M-6 · nit · Two ways to detect a symlink in `Platform/`
**Answer:** Logged — no ideas file fits (`ls docs/ideas` has no ingest or security file); recorded here and as an Out line ("not planned").
