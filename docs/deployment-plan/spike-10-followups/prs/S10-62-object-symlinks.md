# S10-62 — Refuse symlinked objects in the app

**Status:** Draft
**Spike / slice:** Spike 10 follow-ups · likeness hardening
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a refused object shows the missing state each reader already has)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s10-62-object-symlinks`
**Version bump:** patch — a security bugfix with no new feature ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [source layer data model, storage](../../../source-layer-data-model.md), [application stack, read path](../../../application-stack.md), audit 2026-10-18 finding SEC-2 (likeness and depiction area audit, all five lenses)
**Planned:** 2026-10-18 against `main` @ `aa60b63`
**Review:** — (not reviewed yet)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

The app reads object files (likeness images, Source thumbnails, Artifact files) straight from the project's `objects/` folder. Today it checks only the text of the path, so a project folder someone else made can hold a symlinked `objects/`, a symlinked shard folder, or a symlinked object. The app then follows the link and shows any image the user can read as a portrait or a Source cover, or opens any file in its default app. After this PR the app refuses the same things Go's `core/objectstore` refuses, and the reader shows the missing state it already has.

**Why now.** The 2026-10-18 likeness audit (SEC-2) found that S10-34, S10-35, and S10-39 added readers that load on their own: the Persons list thumb, the Person header, the gallery, the composer line, and the evidence card. Opening a shared project is enough to trigger the read.

**What stays.** Catalog paths are still built by Go from the checksum (`objectpath.Rel`). The project folder itself may sit behind a symlink, as Go's `os.OpenRoot(projectDir)` allows. Image type checks, pixel-size checks, and decode concurrency in `ThumbnailCache` are not part of this PR.

## 2. User stories

**US-1.** As a researcher who opens a project folder someone sent me, when that folder links `objects/` or anything under it to a place outside the project, I see the usual missing-file state instead of a file from elsewhere on my Mac.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `objects/` in the project is a symlink to an outside folder holding `ab/cd/x.png`; the app asks for `objects/ab/cd/x.png` | `ProjectFiles.objectURL` returns `nil` |
| AC-1.2 | `objects/ab` is a symlink to an outside folder holding `cd/x.png` | `objectURL` returns `nil` |
| AC-1.3 | `objects/ab/cd/x.png` is a symlink to an outside PNG | `objectURL` returns `nil` |
| AC-1.4 | `objects/ab/cd/x.png` is a folder, or a FIFO made with `mkfifo` | `objectURL` returns `nil`, and returns without blocking |
| AC-1.5 | `objects/ab/cd/x.png` does not exist | `objectURL` returns `nil` |
| AC-1.6 | `relPath` is `catalog.sqlite` or `derivatives/x.jpg` (first segment is not `objects`) | `objectURL` returns `nil` |
| AC-1.7 | A real regular file at `objects/ab/cd/x.png`, and the project folder path itself is a symlink to the real project folder | `objectURL` returns the URL `projectDir/objects/ab/cd/x.png` |
| AC-1.8 | A thumbnail is requested for the symlinked object from AC-1.3 | `ThumbnailCache.image` returns `nil` and caches the missing sentinel; no image of the outside file is produced |
| AC-1.9 | The Artifact viewer loads the symlinked object from AC-1.3 | `ArtifactViewerModel.emptyReason == .missingFile` |
| AC-1.10 | Open in default app on the symlinked object from AC-1.3 | `ProjectFiles.openObject` returns `false`, so the caller shows its existing file-missing message |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ProjectFiles.objectURL` requires `objects` as the first segment and checks each component from `objects/` to the object with `lstat`: every folder is a real folder, the object is a regular file (D-1–D-5) | STEP-1, STEP-2, AC-1.1–AC-1.7 |
| `openObject` and `ArtifactViewerModel.load` drop their separate `fileExists` checks, which `objectURL` now covers (D-6) | STEP-3, AC-1.9, AC-1.10 |
| Tests for each refusal, for `ThumbnailCache` and `ArtifactViewerModel`, and the two existing join tests moved onto a real temp project | STEP-2, STEP-3, T-1–T-11 |
| The source layer storage paragraph and the application stack read path say what the app checks | STEP-4 |

| Out | Home |
| --- | --- |
| `ThumbnailCache` refuses a type other than PNG or JPEG and a declared pixel size over 4096 before decoding, and limits how many decodes run at once (audit SEC-2, second half) | audit 2026-10-18 Medium item 5, not yet mapped |
| Reading object bytes through a no-follow file descriptor instead of a URL (TOCTOU) | not planned (D-4) |
| A symlink or FIFO check on `derivatives/` or on the catalog file | not planned: no Swift reader opens them by catalog path |
| App Sandbox | not planned ([`application-stack.md`](../../../application-stack.md) read path) |
| Go-side changes to `core/objectstore` | not planned: Go already refuses these cases |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Path join | `macos/App/Platform/ProjectFiles.swift` `ProjectFiles.objectURL(projectDir:relPath:)` | Splits `relPath` on `/`, refuses `.`, `..`, and segments starting with `/` or `~`, appends the rest to `projectDir`, then checks that the `standardizedFileURL` path starts with `projectDir + "/"`. It never touches the disk, so a symlink anywhere under the project passes. Any first segment is allowed. |
| Open in app | `macos/App/Platform/ProjectFiles.swift` `ProjectFiles.openObject(projectDir:relPath:)` | `objectURL`, then `FileManager.default.fileExists(atPath:)` (which follows symlinks), then `NSWorkspace.shared.open(url)`. |
| Thumbnail reader | `macos/App/Platform/ThumbnailCache.swift` `ThumbnailCache.image(projectDir:relPath:)` | Calls `objectURL` inside an unstructured `Task` on the main actor; `nil` → returns `nil` and stores `.missing`. Otherwise `CGImageSourceCreateWithURL` in a detached task, which follows links. Every likeness and cover thumb (`CachedThumbnail`, `EvidenceCardView.loadMissingLikenesses`) goes through it. |
| Viewer reader | `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` `ArtifactViewerModel.load(_:)` | `objectURL` `nil` → `.missingFile`; then a separate `fileExists` → `.missingFile`; then `NSImage(contentsOf:)` or `PDFDocument(url:)` in a detached task. |
| Open callers | `macos/App/Features/Sources/SourceArtifactsSection.swift` `open(_:)`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `openSelectedFile()` | `openObject` `false` → `L10n.Sources.fileOpenMissing` or `L10n.EvidenceGraph.fileNotFoundMessage`. |
| Go rules to mirror | `core/objectstore/objectstore.go` `openObjectsRoot`, `openRegular` | `os.OpenRoot(projectDir)` (the project path may be a symlink), `Lstat("objects")` must be a folder, each shard folder top-down must be a folder (`Lstat`), the object must be `IsRegular()` before the open, so a FIFO cannot block. |
| Paths Go returns | `core/objectpath/objectpath.go` `Dir = "objects"`, `Rel`; `core/database/files` `StorageRelPath` | Every `relPath` that reaches Swift (Artifact files, derived thumbnails, depiction images) is `objects/{hh}/{hh}/{hex}{ext}`. |
| Swift precedent | `macos/App/Platform/IngestMediaPolicy.swift` `IngestMediaPolicy.validate(url:)` | Refuses a picked file whose `isSymbolicLink` resource value is true. Only the final component, and only for ingest. |
| Tests | `macos/ProvenenciaTests/ProjectFilesTests.swift` | Lexical cases only, against the non-existent `/tmp/proj.provenencia`. `objectURLJoinsNormalContentAddressedPath` and `objectURLJoinsPathWithExtension` expect a URL for files that do not exist; they fail once `objectURL` checks the disk. No symlink case. |
| Tests | `macos/ProvenenciaTests/ThumbnailCacheTests.swift` `missingFileCachesSentinel`, `loadsAndHitsCache`; `macos/ProvenenciaTests/ArtifactViewerModelTests.swift` `missingFileSetsEmptyReason`, `loadsPNGImageViaProjectFiles`, `makeTempProject`, `writeFile` | Real temp projects under `FileManager.default.temporaryDirectory` (on macOS `/var/folders/…`, itself behind the `/var` → `/private/var` symlink), files under `objects/`. They keep passing because the check starts at `objects/`. |
| Go tests to mirror | `core/objectstore/objectstore_test.go` subtests "objects subfolder symlinked outside", "objects folder symlinked outside", "object path symlinked to an outside file" | The three cases this PR brings to Swift. |

**Precedent.** `core/objectstore` `openRegular` (S10-53 / S10-56, [#425](https://github.com/mendahu/provenencia/pull/425)) is the rule set to copy: a top-down `lstat` of each folder, then `lstat` of the object, regular file required before any open. Copy its order and its refusals; do not copy `os.Root`, which Swift has no equivalent of.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The check lives in `ProjectFiles.objectURL` itself. Its signature and its `URL?` result stay; `nil` now also means "not a safe, existing regular file". | Every reader (`ThumbnailCache`, `ArtifactViewerModel`, `openObject`) already goes through it and already maps `nil` to its missing state, so all of them gain the check with no caller change and no second function a new reader could skip. | A new `readableObjectURL` beside `objectURL` — because the old one would stay callable without the check; a throwing API — because no caller shows a different message for a refused object than for a missing one. |
| D-2 | The first segment of `relPath` must be `objects` (match `objectpath.Dir`), and at least one more segment must follow. | Go never returns any other path to Swift; refusing the rest keeps `objectURL` from becoming a reader for `provenencia.sqlite` or `derivatives/`. | Allowing any project-relative path — because nothing needs it and it widens what a crafted path can reach. |
| D-3 | The disk check uses Darwin `lstat` on each path from `projectDir/objects` down to the object: `objects` and every shard folder must be `S_IFDIR`; the object must be `S_IFREG`. Any `lstat` failure (missing, permission) returns `nil`. The project folder path itself is not checked. | It is the same rule as Go's `openObjectsRoot` and `openRegular`, which open the project with `os.OpenRoot` and so allow a symlinked project path (on macOS every temp folder is behind `/var` → `/private/var`). `lstat` reads fresh from disk; checking the type before any open means a FIFO never blocks a reader. | `URL.resourceValues(forKeys: [.isSymbolicLinkKey, .isRegularFileKey])` — because Foundation caches resource values on the URL object and the API reads less directly than `lstat`; `resolvingSymlinksInPath()` and a prefix check — because it would allow a symlink that points back inside the project, which Go refuses. |
| D-4 | Check, then read by URL. No file-descriptor read. | The threat is a static folder someone else prepared, not a process racing the app; anyone who can swap a link between the check and the read can already rewrite the object bytes. The readers (`CGImageSourceCreateWithURL`, `NSImage(contentsOf:)`, `PDFDocument(url:)`, `NSWorkspace.open`) take URLs, and reading a large PDF into `Data` through an `O_NOFOLLOW` descriptor would hold it all in memory. | `open(O_NOFOLLOW)` and decode from `Data` — because of the memory cost on PDFs and because `NSWorkspace` cannot open a descriptor. |
| D-5 | The `lstat` calls run where `objectURL` is called today, including the main actor in `ThumbnailCache` and `ArtifactViewerModel`. | At most five `lstat` calls per object (`objects`, two shard folders, the object; one more for a deeper path) on a local disk; moving them is a performance change this PR does not need. | Moving `objectURL` into the detached decode task — because it changes `ThumbnailCache`'s structure for no measured cost. |
| D-6 | Delete the `fileExists` check in `openObject` and in `ArtifactViewerModel.load`. | `objectURL` now returns `nil` for a missing file, so both checks are dead, and `fileExists` follows symlinks, so keeping it suggests the wrong rule. | Keeping them as belt and braces — because they are dead code that reads like a second rule. |
| D-7 | No new copy and no new error state. A refused object shows exactly what a missing one shows. | The researcher can do the same thing in both cases (the file is not in this project), and a "symlink refused" message would add a string for a case only a crafted folder hits. | A distinct "unsafe file" message — because it costs an L10n key and a state per reader for no different action. |
| D-8 | Version bump: patch, `0.2.0` → `0.2.1`, in `VERSION`, `core/version.go`, and `MARKETING_VERSION` together. | A security bugfix with no feature, per [`docs/versioning.md`](../../../versioning.md) "What bumps". | None — because the PR ships a behavior change. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ProjectFiles.isSafeObject(at:projectURL:)` *(new)*, private | func | `IngestMediaPolicy.validate(url:)` (final-component symlink check); `core/objectstore` `openRegular` | `IngestMediaPolicy` checks one picked file, not every folder under `objects/`, and returns ingest callouts. Go's rule cannot be called from Swift for a URL read. This private helper extends `objectURL` (D-1); it is not a second entry point. Searched: `isSymbolicLink`, `lstat`, `resolvingSymlinks`, `fileExists` under `macos/App`. |
| `ProjectFilesTests` symlink and FIFO cases *(new)* and a temp-project helper inside that file *(new)* | helper | `ArtifactViewerModelTests` `makeTempProject` / `writeFile` (private to that suite); `ArtifactFileFixtures.uniqueProjectDir` | Use `ArtifactFileFixtures.uniqueProjectDir()` for the project path; the file and link writes are three lines of `FileManager` calls, private to the suite, like `ArtifactViewerModelTests` does. |

**Builds on.** The `core/objectstore` read rules (`openObjectsRoot`, `openRegular`) and `objectpath.Dir`; the existing `nil` → missing mapping in each reader.
**Deletes.** The `fileExists` check in `ProjectFiles.openObject` and in `ArtifactViewerModel.load` (D-6). The two lexical-only join tests become disk tests (T-1, T-2).
**Next adds.** Audit Medium item 5 (type and pixel-size checks before `ThumbnailCache` decodes) runs after `objectURL` on a URL this PR has already made safe; it adds its check in `ThumbnailCache`, not here.

### Schema

None — no catalog change.

### Go core

None — Go already refuses these cases (`core/objectstore`).

### FFI

None — no message changes.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Object path check | `ProjectFiles.objectURL` | `macos/App/Platform/ProjectFiles.swift` | After the existing lexical checks and before the prefix check: require `segments.first == "objects"` and `segments.count >= 2` (D-2). After the prefix check: `isSafeObject` walks `projectURL/objects`, then each further folder, then the object (D-3). |

Sketch:

```swift
static func objectURL(projectDir: String, relPath: String) -> URL? {
    // existing lexical join and refusals, plus: first segment == objectsDir, count >= 2
    // existing prefix check
    guard isSafeObject(segments: segments, projectURL: projectURL) else { return nil }
    return resolved
}

/// Mirrors core/objectstore openObjectsRoot + openRegular: every folder from
/// objects/ down is a real folder and the object is a regular file (lstat,
/// never followed). The project folder path itself may be a symlink.
private static func isSafeObject(segments: [String], projectURL: URL) -> Bool
// lstat(path, &st); (st.st_mode & S_IFMT) == S_IFDIR for segments[0..<count-1]; == S_IFREG for the last
```

States: a refused object is the missing state each reader already has: `ThumbnailCache` `.missing` (placeholder glyph), `ArtifactViewerEmptyReason.missingFile`, `openObject` → `false` (D-7). Accessibility: unchanged.

### Strings

None — no new copy (D-7).

## 7. Steps

**STEP-1 — Check the object on disk in `objectURL`**
Files: `macos/App/Platform/ProjectFiles.swift`.
Do: Add the `objects` first-segment rule (D-2) and the private `isSafeObject` `lstat` walk (D-3). Update the `objectURL` doc comment to say it returns `nil` for a missing file, a symlink at or under `objects/`, or anything that is not a regular file. Keep the signature (D-1).
Check: the Mac target builds: `xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` succeeds.

**STEP-2 — Prove each refusal**
Files: `macos/ProvenenciaTests/ProjectFilesTests.swift`. Skill: `add-swift-test`.
Do: Rewrite `objectURLJoinsNormalContentAddressedPath` and `objectURLJoinsPathWithExtension` to write real files under a temp project from `ArtifactFileFixtures.uniqueProjectDir()` (T-1, T-2). Add T-3 to T-9. Remove each temp project in a `defer`. The lexical tests stay as they are (they still return `nil`).
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/ProjectFilesTests` passes.

**STEP-3 — Drop the dead existence checks and test the readers**
Files: `macos/App/Platform/ProjectFiles.swift` (`openObject`), `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` (`load`), `macos/ProvenenciaTests/ThumbnailCacheTests.swift`, `macos/ProvenenciaTests/ArtifactViewerModelTests.swift`. Skill: `add-swift-test`.
Do: Delete the `fileExists` guard in `openObject` and the `fileExists` guard after `objectURL` in `ArtifactViewerModel.load` (D-6). Add T-10 and T-11.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/ThumbnailCacheTests -only-testing:ProvenenciaTests/ArtifactViewerModelTests -only-testing:ProvenenciaTests/ProjectFilesTests` passes.

**STEP-4 — Version and docs**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION`), `docs/source-layer-data-model.md`, `docs/application-stack.md`, `docs/deployment-plan/spike-10-followups/deployment-plan.md`, this plan.
Do: Bump to `0.2.1` (D-8). In `docs/source-layer-data-model.md`, the storage paragraph that says "Object files are read, written, and removed only through `core/objectstore`" becomes: Go reads, writes, and removes object files only through `core/objectstore`; the app reads them by their relative path through `ProjectFiles.objectURL`, which refuses the same things (a path not under `objects/`, a symlinked `objects/`, shard folder, or object, and anything that is not a regular file). In `docs/application-stack.md` **Read path**, add one sentence: Swift opens a relative path only through `ProjectFiles.objectURL`, which applies those refusals. Then the Definition of done lines for the map row and this plan.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10-followups/prs/S10-62-object-symlinks.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

All Swift Testing, in existing files (no `project.pbxproj` wiring). Each test makes its own project with `ArtifactFileFixtures.uniqueProjectDir()` (it returns a path only; create the folder with `FileManager.default.createDirectory(atPath:withIntermediateDirectories: true)`), writes the 1×1 PNG bytes `ThumbnailCacheTests.loadsAndHitsCache` already uses, makes links with `FileManager.default.createSymbolicLink(atPath:withDestinationPath:)`, and makes the outside target in a second temp folder.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/ProjectFilesTests.swift` `objectURLJoinsNormalContentAddressedPath` (rewritten) | real file `objects/ab/cd/abcd` → URL path `projectDir/objects/ab/cd/abcd` | AC-1.7, D-1 |
| T-2 | Swift | `ProjectFilesTests.swift` `objectURLJoinsPathWithExtension` (rewritten) | real file `objects/ab/cd/abcd.jpg` → URL path ends `objects/ab/cd/abcd.jpg` | AC-1.7 |
| T-3 | Swift | `ProjectFilesTests.swift` `objectURLRefusesSymlinkedObjectsFolder` *(new)* | `objects` → outside folder holding `ab/cd/x.png` → `nil` | AC-1.1, D-3 |
| T-4 | Swift | `ProjectFilesTests.swift` `objectURLRefusesSymlinkedShardFolder` *(new)* | `objects/ab` → outside folder holding `cd/x.png` → `nil` | AC-1.2, D-3 |
| T-5 | Swift | `ProjectFilesTests.swift` `objectURLRefusesSymlinkedObject` *(new)* | `objects/ab/cd/x.png` → outside PNG → `nil` | AC-1.3 |
| T-6 | Swift | `ProjectFilesTests.swift` `objectURLRefusesNonRegularObject` *(new)* | `objects/ab/cd/x.png` is a folder → `nil`; is a FIFO (`mkfifo(path, 0o644)`) → `nil`, and the test finishes (no hang) | AC-1.4, D-3 |
| T-7 | Swift | `ProjectFilesTests.swift` `objectURLRefusesMissingObject` *(new)* | `objects/ab/cd/` exists, `x.png` does not → `nil` | AC-1.5 |
| T-8 | Swift | `ProjectFilesTests.swift` `objectURLRefusesPathOutsideObjects` *(new)* | real files `provenencia.sqlite` and `derivatives/x.jpg` in the project → `nil` for both; `relPath` `objects` alone → `nil` | AC-1.6, D-2 |
| T-9 | Swift | `ProjectFilesTests.swift` `objectURLAllowsSymlinkedProjectFolder` *(new)* | a link to the real project folder used as `projectDir`, real file `objects/ab/cd/x.png` → non-nil URL | AC-1.7, D-3 |
| T-10 | Swift | `macos/ProvenenciaTests/ThumbnailCacheTests.swift` `symlinkedObjectCachesSentinel` *(new)* | object symlinked to an outside PNG → `image` is `nil`, `entryCount == 1` | AC-1.8 |
| T-11 | Swift | `macos/ProvenenciaTests/ArtifactViewerModelTests.swift` `symlinkedObjectSetsMissingFile` *(new)*; `ProjectFilesTests.swift` `openObjectReturnsFalseForSymlinkedObject` *(new)* | object symlinked to an outside PNG, `mediaType` `image/png` → `emptyReason == .missingFile`; `openObject` → `false` | AC-1.9, AC-1.10, D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `ArtifactViewerModelTests`, `ThumbnailCacheTests`, `CitationComposerDepictionTests`, `PersonDetailTests`, and any test that writes an object under a path not starting `objects/`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none added: D-7)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief: no UI change)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together): 0.2.1
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] docs/deployment-plan/spike-10-followups/deployment-plan.md: S10-62 row status → Landed #N (this folder has no completed.md)
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated): not gated
- [ ] Domain doc updated: docs/source-layer-data-model.md storage paragraph and docs/application-stack.md read path (STEP-4)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Copy a project that has a Person with a likeness. In Terminal, move its `objects` folder to `/tmp/elsewhere-objects` and run `ln -s /tmp/elsewhere-objects objects` inside the project folder. Open the copy. The Persons list thumb, the Person header, and the gallery tiles show the placeholder. A Source with a photo shows the missing-file state in the viewer, and Open in default app shows the file-missing message.
2. Remove the link and move the folder back. Reopen. Every thumb and the viewer show the images again.
3. In a fresh copy, replace one likeness object file with `ln -s ~/Pictures/<any>.png <object path>`. Its tile shows the placeholder, not the picture from Pictures.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A real project stores an object under a path that does not start `objects/` | §4 verified every Go producer goes through `objectpath.Rel` or `files.StorageRelPath`; T-8 pins the rule |
| A test writes objects in a temp folder behind `/var` → `/private/var` | D-3 checks from `objects/` down only; T-9 pins it |
| A FIFO in `objects/` hangs a test or a reader | D-3 checks the type with `lstat` before any open; T-6 |
| `lstat` on the main actor is slow on a network volume | D-5 accepts it; at most five calls per object |

**Stop and report if:**

- A Swift caller of `ProjectFiles.objectURL` other than `ThumbnailCache.image`, `ArtifactViewerModel.load`, and `ProjectFiles.openObject` exists at your base (`rg "objectURL\(" macos/App`).
- A `relPath` that reaches Swift from Go does not start with `objects/` (a test or a fixture in `FakeStore.swift` that only passes without D-2).
- `lstat` is not callable from `ProjectFiles.swift` with `import Foundation` (Darwin) on the deployment target.
- An existing test in `ArtifactViewerModelTests`, `ThumbnailCacheTests`, or `CitationComposerDepictionTests` fails after STEP-1 for a reason other than a missing `objects/` file.

## 12. Open questions

None.

## 13. PR

**Title:** S10-62 — Refuse symlinked objects in the app
**Body:** objective in one sentence; what shipped against US-1; AC evidence (test names); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10-followups/prs/S10-62-object-symlinks.md`.
