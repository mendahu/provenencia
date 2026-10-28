# Review — S12-23 The app opens only safe files and links

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

Round 1 raised blockers (M-1 – M-5, X-1), all answered Fixed with changes that move code into an existing owner or name an existing test. No fix reshaped the plan, so the writer did not run round 2.

## To user

None.

## Writer's own calls

- D-2 / D-3: a disallowed type gives no path. The Artifact still arrives with its file name and type, so the Source page can name the file it refuses. The canvas Files pane drops it, as it drops any Artifact with no path.
- D-5: a symlink or Finder alias stays "missing" (S10-62's call). Only an execute bit or a dangerous type reads as refused.
- D-6: ingest also refuses `.script` types now, through the shared predicate. Go already refused them.
- D-8 / D-9: the copy is "Provenencia won't open “x”. Its file type could run a program, or it isn't a type Provenencia stores." The canvas callout title becomes "Can't open the file" for both the missing case and the refused case.
- Out: `PurgeUnreferenced` leaves a damaged row's bytes on disk. Its refusal is not listed in the canvas Files pane.

## Round 1

### M-1 · blocker · Make the allowlist the single table of type → extension, not two lists that must agree
**Plan:** D-1 checked `Allowed && Extension != ""`, and folding `Extension` into `mediatypes` was Out.
**Codebase:** `core/objectpath/objectpath.go` `Extension` is a third type list. It has its own `Normalize` copy, it maps `image/jpg`, and `Rel` is its only caller.
**Answer:** Fixed. D-1 now makes `mediatypes.allowed` a type → extension table with `StoredExtension`. `objectpath.Extension` is deleted, the Out row and the drift Stop condition are removed, and T-1 is added.

### M-2 · blocker · A second Swift "is this an executable type" check beside `IngestMediaPolicy`'s
**Codebase:** `macos/App/Platform/IngestMediaPolicy.swift` `validate(url:)` checks executable, unixExecutable and application inline.
**Answer:** Fixed. D-6 adds `IngestMediaPolicy.isExecutableType` (executable, unixExecutable, application, script), which both `validate` and `openableObjectURL` (D-7) use. T-16 is added.

### M-3 · blocker · Deciding a metadata value's web link is a domain rule, and the plan puts it in the FFI handler
**Codebase:** `core/database/sourcemetadata/sourcemetadata.go` `normalizeValue` owns the url rule, and `.cursor/rules/ffi-handlers.mdc` keeps handlers thin.
**Answer:** Fixed. D-11 adds the `WorkspaceEntry.OpenHref()` method in `sourcemetadata`, and `metadataEntryProto` only copies it. T-8 is added.

### M-4 · blocker · Swift `openURL` re-checks the url data type that Go already decided
**Answer:** Fixed. In D-12, `openURL` is `openHref.isEmpty ? nil : URL(string:)`. The view's existing url-field branch decides only the styling. T-18 drops its text-field row.

### M-5 · blocker · The new string's key and namespace break `add-localized-string`
**Codebase:** every key in `L10n.Errors` is `error.<domain>.<name>` and maps an FFI code (skill step 8).
**Answer:** Fixed. D-8 now uses `L10n.ObjectOpen.refused(filename:)` with the key `objectOpen.refused.message`.

### M-6 · major · Five call sites each map `ErrDisallowedType` to "no path"
**Answer:** Fixed. D-3 adds `files.ReadRelPath`, used by `RelPaths`, `artifactProto` and both thumbnail functions. `searchindex` and `derivatives` keep one mapping each, because they cannot or must not call it.

### M-7 · major · An existing test breaks and the plan adds a parallel test instead of editing it
**Answer:** Fixed. T-4 rewrites `TestObjectsRelPathAgreesWithObjectpath` in place, and no new file is added. §4 lists it, and so does §6 Deletes.

### M-8 · major · Docs this PR makes false are not in STEP-8
**Answer:** Fixed. STEP-7 now names `application-stack.md` "Object names" and the `source-layer-data-model.md` basename paragraph. The files and objectpath doc comments are in STEP-1, and `bare hex` is in the sweep with an `rg` check.

### M-9 · minor · `isSafeObject` returning `ObjectRefusal?` reads backwards
**Answer:** Fixed. D-5 renames it `objectRefusal(segments:projectURL:)`.

### M-10 · minor · `raster.Supported` is another normalize-and-list copy
**Answer:** Logged. This is an Out row ("not planned") for S12-28's planner, which reworks the viewer decode policy. This PR does not touch `raster`.

### X-1 · blocker · Existing Go tests that STEP-1 and STEP-2 break are not named
**Codebase:** `searchindex_test.go` `TestObjectsRelPathAgreesWithObjectpath`; `derivatives/thumbnail_test.go` skip case with `application/octet-stream`.
**Answer:** Fixed. §4 has a "Go tests that this PR breaks" row, and STEP-1 says what happens to each. The derivatives skip case moves to `audio/mpeg`. A Stop condition covers any other test that breaks.

### X-2 · major · T-12 does not say what mode the first three cases use
**Answer:** Fixed. T-14 lists each case with its mode: the extension and type cases are at `0o644`, the PDF is at `0o755`, and an `.app` case is added for the application branch.

### X-3 · major · `openableObjectURL` needs a refusal-returning core under `objectURL`
**Answer:** Fixed. D-5 names `checkedObject(projectDir:relPath:) -> Result<URL, ObjectRefusal>`. Its six guards are `.missing`, and `objectURL` is `try? …get()`.

### X-4 · major · AC-2.8 has no named test
**Answer:** Fixed. AC-2.8 is marked "manual only, §10 step 5".

### X-5 · major · The checks for STEP-5 and later cannot run without `xcodebuild`
**Answer:** Fixed. A Stop condition says how to proceed and what to report when `xcodebuild` is missing.

### X-6 · minor · STEP-5/6 ordering
**Answer:** Fixed. The Source page and canvas callers are merged into STEP-5, which is one commit.

### X-7 · minor · Choices left to the implementer
**Answer:** Fixed. The L10n function returns `String` through `L10n.format`, and T-14 uses `@Test(arguments:)`. The fake resets `openHref` on new and existing entries (D-13). `.unixExecutable` is explicit in D-6.

### X-8 · nit · Citation style
**Answer:** Logged. The precedent line names the "image not a PNG" row as a pattern.
