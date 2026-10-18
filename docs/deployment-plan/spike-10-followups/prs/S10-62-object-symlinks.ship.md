# Ship log — S10-62 Refuse symlinked objects in the app

**Plan:** [`S10-62-object-symlinks.md`](S10-62-object-symlinks.md) · **Branch:** `s10-62-object-symlinks` · **PR:** #464
**State:** ready

Base: `main` @ `9467e847`. Worktree: `/Users/jakerobins/Repos/provenencia-s10-62`.

## Round 0 — implement
First run stopped: host disk full (180 MiB free), `xcodebuild` could not write DerivedData. User approved deleting DerivedData of removed worktrees (2026-10-18); implementer resumed.
Head `527185da`. Gate PASS (orchestrator rerun). PR #464 (draft). Report deviations: `symlinkedObjectSetsMissingFile` control asserts `emptyReason == .idle` (non-optional) plus `displayImage != nil`; mirror-line and Go test header wording follow S10-64 D-10 (recorded in the plan's Landed note); ship log was swept into f8719a7a and untracked again in 527185da. Manual check (§10) not run.

## Round 1 — review
Verdicts: CON clean · MNT clean · UX clean · SEC changes needed · PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | major | A Finder alias at an object path passes `lstat` as `S_IFREG`, and Open in default app opens the alias's target | Fix now → F-1 | verified by orchestrator probe: bookmark file written with `URL.writeBookmarkData` → `lstat` S_IFREG true, `isAliasFile` true, `NSWorkspace.urlForApplication(toOpen:)` on `obj.pdf` alias → TextEdit.app (the .txt target's app), while a real `.pdf` → Preview.app. Launch Services resolves the alias. Extends D-3 rather than contradicting it: D-3 rejected `resourceValues` for the symlink and type check because values cache on a reused URL; F-1 reads one key from a freshly built URL |
| CON-1 | CON | minor | T-10 `symlinkedObjectCachesSentinel` has no control step, which §8 requires (also MNT outside-lens note) | Fix now → F-2 | read `ThumbnailCacheTests.swift:40`: ends at `entryCount == 1`; plan §8 preamble requires the control |
| MNT-1 | MNT | minor | `objectURL` builds the path twice; the prefix check can no longer fail | Fix now → F-3 (comment only) | F-1 already edits `isSafeObject`; a few lines. Keep the prefix check as defence in depth and say so, rather than removing a guard in security code |
| MNT-2 | MNT | minor | Folder-link tests repeat a 10-line hostile-folder setup | Fix now → F-4 | F-1 adds a fixture beside it in `ArtifactFileFixtures` and a test in `ProjectFilesTests`; S10-64 reuses these fixtures |
| MNT-3 | MNT | nit | `ArtifactViewerModelTests` has suite-private temp-project helpers beside `ArtifactFileFixtures` | Logged | existed before this PR; plan Out says no tidy-up. Follow-up candidate |
| MNT-o1 | MNT→SEC | note | `IngestMediaPolicy.validate(url:)` uses `resourceValues` for symlinks | Rejected | plan Out, logged as M-6 in the review log |

### Fix plan 1

**F-1 — Refuse a Finder alias at the object path** (from SEC-1)
Files: `macos/App/Platform/ProjectFiles.swift`, `macos/ProvenenciaTests/ArtifactFileFixtures.swift`, `macos/ProvenenciaTests/ProjectFilesTests.swift`, `docs/source-layer-data-model.md`, `docs/application-stack.md`, `.cursor/rules/macos-client.mdc`, the plan's Landed note. Skill: `add-swift-test`.
Failing test first: `ProjectFilesTests.objectURLRefusesAliasObject` — add `static func aliasToOutsidePNG(at aliasPath: String) throws -> String` to `ArtifactFileFixtures` (same shape as `linkToOutsidePNG`: fresh outside folder via `uniqueProjectDir()`, `writePNG(projectDir: outside, relPath: "outside.png")`, create the alias's parent folders, write a bookmark file at `aliasPath` with `URL(fileURLWithPath: <outside>/outside.png).bookmarkData(options: .suitableForBookmarkFile, includingResourceValuesForKeys: nil, relativeTo: nil)` and `URL.writeBookmarkData(_:to:)`, return the outside folder). In a temp project, put the alias at `objects/ab/cd/<64 hex>.png`, assert `objectURL` is `nil`; control: remove it, `writePNG` at the same relPath, assert non-nil. Run it red before the code change.
Do: in `isSafeObject`, after the final component passes the `S_IFREG` check, read `URL(fileURLWithPath: url.path).resourceValues(forKeys: [.isAliasFileKey]).isAliasFile` on a freshly built URL; `true` or a thrown error returns `false`. Comment why (Launch Services resolves aliases on open; a fresh URL so no cached value). Update the `objectURL` doc comment and the three STEP-4 doc places that list what the app checks so they say "and is not a Finder alias". Add one line to the plan's Landed note: alias refusal added in fix round 1 from SEC-1.
Check: the test passes; `scripts/check-pr.sh` PASS.

**F-2 — Give T-10 its control** (from CON-1)
Files: `macos/ProvenenciaTests/ThumbnailCacheTests.swift`.
Do: at the end of `symlinkedObjectCachesSentinel`, remove the link, `ArtifactFileFixtures.writePNG(projectDir: project, relPath: rel)`, and assert `await ThumbnailCache().image(projectDir: project, relPath: rel) != nil` (a fresh cache, since the old one keeps `.missing`).
Check: the test passes.

**F-3 — Say the prefix check is defence in depth** (from MNT-1)
Files: `macos/App/Platform/ProjectFiles.swift`.
Do: replace the comment above the `hasPrefix(projectPath + "/")` guard with one saying the segment rules above already guarantee it and it stays as defence in depth. No behavior change.

**F-4 — One fixture for a symlinked folder** (from MNT-2)
Files: `macos/ProvenenciaTests/ArtifactFileFixtures.swift`, `macos/ProvenenciaTests/ProjectFilesTests.swift`.
Do: add `static func linkToOutsideFolder(at linkPath: String, holding relPath: String) throws -> String` beside `linkToOutsidePNG` (fresh outside folder, `writePNG(projectDir: outside, relPath: relPath)`, create the link's parent folders, symlink `linkPath` → outside folder, return outside). Use it in `objectURLRefusesSymlinkedObjectsFolder` and `objectURLRefusesSymlinkedShardFolder`; keep their controls.
Check: both tests pass; `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `87728ef9`. Gate PASS (orchestrator rerun). F-1 done (`objectURLRefusesAliasObject` red first, then green; 6cd8c2b1); F-2 done (a6779436); F-3 done (69de12d5); F-4 done (87728ef9). Deviation: `macos-client.mdc` wording "also refuses symlinks and Finder aliases" (fix plan gave none).

## Round 2 — review
Scope `527185da...87728ef9`, lenses SEC, CON, MNT. Verdicts: SEC clean (SEC-1 resolved) · CON clean (CON-1 resolved) · MNT clean (MNT-1, MNT-2 resolved). No new blockers. MNT noted, not raised: the three outside-folder fixtures share a few opening lines; extract if a fourth appears.

## Result
Gate PASS · CI pending at time of writing · open: none · follow-ups: MNT-3 (move `ArtifactViewerModelTests` and `ThumbnailCacheTests.loadsAndHitsCache` onto `ArtifactFileFixtures`) · to user: —
