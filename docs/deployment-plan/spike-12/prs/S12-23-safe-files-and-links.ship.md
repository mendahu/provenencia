# Ship log — S12-23 The app opens only safe files and links

**Plan:** [`S12-23-safe-files-and-links.md`](S12-23-safe-files-and-links.md) · **Branch:** `s12-23-safe-files-and-links` · **PR:** #608
**State:** ready

## Round 0 — implement
Head `acfe12bb` (seven step commits `e82013a9`…`825dcfc5`, landing record `acfe12bb`). Gate PASS (orchestrator, on the pushed head). Report: deviations — `ProjectFiles.ObjectRefusal` also conforms to `Error` (`Result` needs it); extra tests `openObjectReportsUnsafeForExecutableObject` and Go helpers `sourceWithPNG`, `damageCatalog`, `urlMetadataSetFixture`, `writeObject`; STEP-7's `rg 'bare hex' docs core` still hits the plan and its review log (no other hit); `TestListSources` cover case passes without the handler change (`RelPaths` already omitted an erroring path; `TestGetSourceWorkspace`'s artifact case is the red one); D-10 literal, so empty `originalFilename` and empty `label` give empty quotes. Flagged, not edited: `.agents/skills/add-ffi-handler/SKILL.md:67` still says to return one path through `files.StorageRelPath` (now `files.ReadRelPath`). AC-2.8 is manual only, not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none — UX for `macos/App/**` and `Localizable.xcstrings`; SEC for `core/objectpath`, `core/objectstore`, `core/derivatives`, `api/ffi/**`, `ProjectFiles.swift`, URL code; PRF for `core/database/**`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1, CON-3 | UX, CON (SEC outside-lens agrees) | blocker | The execute-bit refusal refuses every object in a project on, or copied from, an exFAT or FAT volume, and the copy blames the type | To user | verified: UX mounted an `hdiutil -fs ExFAT` image; a 0644 file lists `-rwx------`, `cp -R` and `ditto` to APFS keep it; `core/objectstore` tests fail on that volume at head, pass at base. Contradicts D-4, D-5, AC-2.1–2.3, AC-2.6, AC-2.8, the map working rule and bar outcome 11. **User (2026-10-28): open with the app for the type and drop the execute-bit refusal** → Fix now → F-11 |
| SEC-1 | SEC | major | Open lets a per-file `com.apple.LaunchServices.OpenWith` attribute choose the app | To user (with UX-1) → Fix now → F-12 | verifier **Confirmed** (Launch Services queries only, nothing launched): a `.txt` carrying the attribute resolves to Preview, Terminal, JavaLauncher, Calculator or Script Editor while `UTType.plainText` resolves to TextEdit; bundle id alone suffices; the attribute survives ditto zip, `cp -p`, `cp -R` and exFAT (`xattr_preserve_for_intent` SHARE=1). No stock app found that runs an allowlisted, non-executable document at head; Terminal was held off only by the execute bit. User answer as UX-1 |
| UX-2 | UX | major | An execute-bit refusal says "file type could run a program" on Open and "File missing" in the viewer | To user (with UX-1) | verified: `objectURL` is `try? checkedObject(…).get()`; `ArtifactViewerModel:162-166` maps nil to `.missingFile`. User answer to UX-1 drops the bit refusal, so every remaining `.unsafe` cause is the type and the shared copy is true → resolved by F-11, no copy change |
| MNT-1 | MNT | blocker | `searchindex.objectsRelPath` and `derivatives.Ensure` copy `files.ReadRelPath`'s rule | Fix now → F-1 | verified: `go list -deps ./core/database/files` does not list `searchindex` (it already imports `effects`, `catalogmodel`), so D-3's "cannot import files" premise is false; same-pass cleanup (this PR adds all three copies). Supersedes D-3's `searchindex` and `Ensure` clauses |
| MNT-2, UX-4 | MNT, UX | blocker | A third rule for naming an Artifact's file; the refusal can read “” | Fix now → F-2 | verified: `SourceArtifactsSection.swift:687`, `SourcePageArtifactsView.swift:479-484` `previewLabel`, `EvidenceGraphFiles.swift:19-28` `Row.displayFilename`; same-pass cleanup. Supersedes D-10's "label" fallback |
| MNT-3 | MNT | major | `open(_:)` re-derives `openableRelPath` | Fix now → F-3 | verified: `SourceArtifactsSection.swift:688` against `GenealogyStore.swift:84-88` |
| MNT-4, CON-1 | MNT, CON | major | `add-ffi-handler` still teaches `files.StorageRelPath` for a read | Fix now → F-4 | verified: `.agents/skills/add-ffi-handler/SKILL.md:67`; no `StorageRelPath(` caller left in `api` |
| CON-2 | CON | major | `existingThumbnailRelPath` and `Ensure`'s refusal have no guarding test | Fix now → F-5 | verified by CON's mutation runs in a scratch copy: both reverts leave the suite green |
| MNT-5 | MNT | minor | `extAllowed` must stay a subset of `allowed`; no test holds it | Fix now → F-6 | `grep extAllowed core/ingest/mediatypes/*_test.go` empty |
| MNT-6 | MNT | minor | New handler-test helpers sit in `sources_test.go`; older copies stay inline | Fix now → F-7 | same-pass cleanup: `damageCatalog` = `citations_test.go:271-290`; `sourceWithPNG` = the inline sequences in `thumbnails_test.go`, `cover_test.go` |
| CON-4 | CON | minor | The Source page and canvas `.unsafe` branches have no unit test | Fix now → F-8 | `openObject` returns `.unsafe` before `NSWorkspace` for an `.sh` fixture |
| CON-5 | CON | minor | `application-stack.md:348`, `:691` still say "MIME-derived extension when known" | Fix now → F-9 | verified |
| SEC-2 | SEC | minor | `isExecutableType` misses stored-URL (`.webloc`, `.fileloc`, `.inetloc`, `.url`) and Terminal (`.terminal`, `.term`) types | Fix now → F-10 | code this PR adds; conformance read from CoreTypes and Terminal `Info.plist` |
| SEC-3 | SEC | minor | Search hits return a stored thumbnail path with no check; the new Read path sentence overstates | Follow-up (read check); doc wording Fix now → F-9 | search read path untouched by this PR; worst case another object's thumbnail in the same project |
| UX-3 | UX | minor | The Open refusal is not announced and may be off-screen | Later PR (S12-30) | written into S12-30's map scope **Ships** and **Check** (`deployment-plan.md`): the Source page error and the canvas file-open callout opt into `PVCallout` announcing |
| PRF | PRF | — | No findings | — | — |

Nits (logged): `CON-6 · core/urlshape/urlshape_test.go:65 · new loop sits in a non-table TestOpenHref` · `CON-7 · deployment-plan.md:505 · S12-23 Ships names isSafeObject, openSavedURL (record)` · `MNT-7 · SourcePageMetadataView.swift:286 · dataType == .url check redundant with openHref`. MNT-8 (`mediatypes.go:1` package doc) and MNT-9 (`IngestMediaPolicy.swift:4` type doc) are stale comments in files this PR changes: ride along in F-6 and F-10.

### Fix plan 1

Work order: F-1, F-2, F-11, F-12 (blockers and the user's answer), then F-3 to F-10. User answer to UX-1 / SEC-1 / UX-2 (2026-10-28): "App by type, drop x-bit" — Open picks the app from the file's type, and neither layer refuses an object for its execute bit; the extension and type checks stay. This amends D-4, D-5, D-7, AC-2.1, AC-2.2, AC-2.3, AC-2.6, AC-2.8, the map working rule "What the app hands to macOS to open is checked first, and Go decides the type" and bar outcome 11's wording.

**F-1 — One read-side path rule** (from MNT-1)
Files: `core/database/searchindex/searchindex.go`, `core/database/searchindex/searchindex_test.go`, `core/derivatives/ensure.go`, `api/ffi/handlers/` test for the search projection, the plan's Landed note.
Failing test first: none (refactor). Guard added: a handler test where a Source's cover derivative `files` row is damaged to `application/x-sh` (`sourceWithPNG`, `SetSourceCover`, `damageCatalog`), then an RPC that rewrites that Source's search document (find it: e.g. `UpdateSource`) succeeds, and `SearchCatalog`'s hit for the Source has an empty `thumbnail_rel_path`.
Do: `existingSourceCoverThumb` lowercases and trims the checksum as `objectsRelPath` does today, then returns `files.ReadRelPath(files.File{ChecksumSHA256: sum, MediaType: mediaType})`, wrapping a non-nil error as today's "searchindex: bad checksum for object path". Delete `objectsRelPath` and `TestObjectsRelPathAgreesWithObjectpath` (its rows are `TestReadRelPath`'s). `derivatives.Ensure` calls `files.ReadRelPath(src)` and maps an empty path to `ErrUnprocessable`; drop its `objectpath` import if unused. Keep `ReadRelPath(f File)`'s signature. Landed note: D-3's "searchindex cannot import files" was false; `searchindex` and `Ensure` call `ReadRelPath` (review MNT-1).
Check: `rg -n 'ErrDisallowedType' core api --glob '!*_test.go'` prints only `core/objectpath/objectpath.go` and `files.ReadRelPath`; `go list -deps ./core/database/files | grep searchindex` prints nothing; scripts/check-pr.sh PASS.

**F-2 — One rule names an Artifact's file** (from MNT-2, UX-4)
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphFiles.swift`, a Swift test file beside the existing `CatalogArtifact` / `EvidenceGraphFiles` tests, the plan's Landed note. Skill: `add-swift-test`.
Failing test first: `SourcePageModelTests` — an Artifact with `originalFilename "  "`, `label ""`, `relPath ""`, ref `ART-X` → `pageError == L10n.ObjectOpen.refused(filename: "ART-X")` (today it reads “”).
Do: move `EvidenceGraphFiles.Row.displayFilename`'s body to `extension CatalogArtifact { static func fileDisplayName(originalFilename: String, relPath: String, ref: String) -> String }` beside `displayTitle` (trimmed name, else the path's last component, else the ref), plus `var fileDisplayName: String` calling it with `file?.originalFilename ?? ""`, `file?.relPath ?? ""`, `ref`. `Row.displayFilename` calls the static func; `open(_:)` and `previewLabel` use `art.fileDisplayName`. Add a table test of `fileDisplayName` (name; blank name with a path → last component; both blank → ref). Landed note: D-10's label fallback became the shared `fileDisplayName` (review MNT-2, UX-4).
Check: `rg -n 'originalFilename\.isEmpty|originalFilename ?? ""' macos/App` prints nothing outside `fileDisplayName`; scripts/check-pr.sh PASS.

**F-3 — Source page Open asks `openableRelPath`** (from MNT-3)
Files: `macos/App/Features/Sources/SourceArtifactsSection.swift`.
Failing test first: none (refactor); T-17 and F-2's test stay green.
Do: `guard art.file != nil else { return }`; `guard let rel = art.openableRelPath else { refusal; return }`; `ProjectFiles.openObject(projectDir:relPath: rel)`.
Check: `rg -n 'relPath.trimmingCharacters' macos/App/Features/Sources/SourceArtifactsSection.swift` prints nothing; scripts/check-pr.sh PASS.

**F-4 — The handler skill teaches the read path** (from MNT-4, CON-1)
Files: `.agents/skills/add-ffi-handler/SKILL.md:67`, the plan's Landed note.
Do: the clause reads "Return `objects/…` paths through `files.RelPaths` (many Files, one query) or `files.ReadRelPath` (one; a type off the allowlist returns no path). `files.StorageRelPath` is the ingest, stage and purge path and refuses a type off the allowlist." Remove the Landed note's sentence about this line being left stale.
Check: `rg -n 'StorageRelPath' .agents .cursor docs/*.md` prints only the new sentence; scripts/check-pr.sh PASS.

**F-5 — Guard the Source header cover and `Ensure`'s refusal** (from CON-2)
Files: `api/ffi/handlers/sources_test.go`, `core/derivatives/` test file.
Failing test first: guards (pass at head). For each, revert its branch locally (`existingThumbnailRelPath` back to `StorageRelPath`; `Ensure`'s empty-path → `ErrUnprocessable` mapping removed), run it, see it fail, restore. Report the failure lines.
Do: two `TestGetSourceWorkspace` cases as CON-2 wrote them: "cover thumbnail with a damaged type has no path" (`SetSourceCover`, derived `files` row → `application/x-sh`, `source.thumbnail_rel_path == ""`), and "a raster type off the allowlist keeps the file and drops its path" (`DELETE FROM file_derivatives`, artifact's file → `image/jpg`; no error, `artifacts[0].file.rel_path == ""`). One `core/derivatives` case: a source row typed `image/jpg` with no derivative → `Ensure` returns `ErrUnprocessable`.
Check: the three cases pass; scripts/check-pr.sh PASS.

**F-6 — The extension index stays inside the allowlist** (from MNT-5; MNT-8 rides along)
Files: `core/ingest/mediatypes/mediatypes.go`, `core/ingest/mediatypes/mediatypes_test.go`.
Failing test first: none (guard).
Do: `TestStoredExtension` also loops `extAllowed` and asserts `Allowed(mt)` for each value. Package doc: it also owns each stored type's extension, which `objectpath` reads.
Check: scripts/check-pr.sh PASS.

**F-7 — Handler-test fixtures live in the harness and the copies use them** (from MNT-6)
Files: `api/ffi/handlers/harness_test.go`, `sources_test.go`, `citations_test.go`, `thumbnails_test.go`, `cover_test.go`, and any other `api/ffi/handlers/*_test.go` holding the same sequence.
Failing test first: none (refactor).
Do: move `damageCatalog` and `sourceWithPNG` to `harness_test.go`. `sourceWithPNG` returns a small struct (`dir`, `userID`, `sourceID`, `artifactID`, `fileID`). `TestLikenessReadHidesDamagedValue` calls `damageCatalog`. Every inline run of `sourceFixture` → `CreateSource` → `CreateArtifact` → `writePNG` → `IngestArtifactFile` calls `sourceWithPNG`. A site whose assertions read its own Source title or Artifact label, or that ingests a non-PNG, stays inline; list each such site in the report.
Check: `grep -nE 'writePNG\(t, ' api/ffi/handlers/*_test.go` — each remaining line is `sourceWithPNG`'s, `writePNG`'s own test use, or a site listed in the report; `grep -n 'db.Exec(tt.damage' api/ffi/handlers` prints nothing; scripts/check-pr.sh PASS.

**F-8 — Unit tests for the `.unsafe` Open branches** (from CON-4)
Files: `macos/ProvenenciaTests/SourcePageModelTests.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skill: `add-swift-test`.
Failing test first: guards (pass at head); delete each `case .unsafe` locally, see the test fail, restore.
Do: write an existing regular file at `objects/aa/bb/aabb.sh`, mode `0o644`, with `ArtifactFileFixtures.write`. Source page: Open on an Artifact with that `relPath` → `pageError == L10n.ObjectOpen.refused(filename: art.fileDisplayName)`. Canvas: `openSelectedFile` on that row → `fileOpenError == L10n.ObjectOpen.refused(filename: row.displayFilename)` and `canvasNotice == .fileOpenFailed`. (`.sh` is refused by type, whatever the execute-bit answer.)
Check: scripts/check-pr.sh PASS.

**F-9 — Docs match: extensions and the search thumbnail** (from CON-5, SEC-3's doc part)
Files: `docs/application-stack.md:348`, `:691`, `:204`.
Do: lines 348 and 691 say "extension from the ingest allowlist" in place of "MIME-derived extension when known". In the Read path paragraph, after "Go returns that path only for a media type on the ingest allowlist", qualify that the `files` read does this, and that a search hit's thumbnail path is the one stored in the search index when it was projected.
Check: `rg -n 'when known' docs/application-stack.md` prints no object-name line; scripts/check-pr.sh PASS.

**F-10 — Refuse types that open another item** (from SEC-2; MNT-9 rides along)
Files: `macos/App/Platform/IngestMediaPolicy.swift`, `macos/ProvenenciaTests/IngestMediaPolicyTests.swift`, `macos/ProvenenciaTests/ProjectFilesTests.swift`.
Failing test first: T-16 gains `UTType(filenameExtension: "webloc")`, `"fileloc"`, `"inetloc"`, `"url"`, `"terminal"`, `"term"` → `true` (red today); T-14 gains (`aabb.webloc`, `0o644`) and (`aabb.terminal`, `0o644`) → `.failure(.unsafe)`.
Do: `isExecutableType` also returns true for a type conforming to `UTType("public.stored-url")`, `UTType("com.apple.terminal.settings")` or `UTType("com.apple.terminal.session")` (skip an identifier that resolves to nil). Its doc: "could run a program or open another item". `IngestMediaPolicy`'s type doc names the open-time use.
Check: scripts/check-pr.sh PASS.

**F-11 — No refusal for the execute bit alone** (from UX-1, CON-3, UX-2; user answer 2026-10-28)
Files: `core/objectstore/objectstore.go`, `core/objectstore/objectstore_unix_test.go`, `macos/App/Platform/ProjectFiles.swift`, `macos/ProvenenciaTests/ProjectFilesTests.swift`, `.cursor/rules/macos-client.mdc`, `docs/application-stack.md` (Read path), `docs/source-layer-data-model.md` (object files paragraph), `docs/deployment-plan/spike-12/deployment-plan.md` (bar outcome 11 at line 33, the working rule at line 145, S12-23's Ships at line 505), `docs/deployment-plan/spike-12/completed.md`, the plan's Landed note.
Failing test first: Go `TestOpenObjectIgnoresExecuteBit` (replaces T-10): an object written then `chmod 0o755` → `OpenObject` succeeds and reads its bytes (red at head: `ErrExecutable`). Swift `objectURLAcceptsExecuteBit` (replaces T-13): a regular `.png` at `0o755` → `objectURL != nil` (red at head).
Do: Go — delete `ErrExecutable` and the mode check in `openRegular`; `ErrNotRegular`'s, `openRegular`'s and the "mirrors" doc comments go back to naming only what they refuse (not regular, symlinked shard). Delete T-11 `TestWriteIfAbsentReplacesExecutable` (the behavior it pinned is gone). Swift — `objectRefusal` drops the `st_mode & 0o111` check; its doc stops naming the execute bit. T-14's (`aabb.pdf`, `0o755`) row leaves the refused list; T-15 gains `aabb.pdf` at `0o755` → `.success`. `openObjectReportsUnsafeForExecutableObject` is deleted (F-8 covers `.unsafe` with an `.sh` fixture). Keep `ArtifactFileFixtures.write(…mode:)`. Docs — `macos-client.mdc` `objectURL` line and `application-stack.md` Read path stop saying the execute bit is refused and say Open opens with the app for the file's type (F-12); `source-layer-data-model.md` drops the execute bit from both mirrors; the map's bar outcome 11 reads "A stored file whose type is not allowed, or could run a program or open another item, is refused with its name; …"; the working rule's "has an extension, has no execute bit and is not an executable, application or script type" becomes "has an extension and is not a type that could run a program or open another item, and opens it with the app for its type"; the S12-23 Ships paragraph says the same. `completed.md` paragraph and the plan's Landed note: the user's answer (2026-10-28) to review UX-1 / SEC-1 dropped D-4 and D-5's execute-bit refusal (exFAT and FAT report every file executable) and replaced it with opening by type; AC-2.1, AC-2.2, AC-2.3, AC-2.6 and AC-2.8 changed with it. Do not edit the AC table or the Decisions.
Check: `rg -n 'ErrExecutable|0o111|execute bit' core macos .cursor docs/application-stack.md docs/source-layer-data-model.md` prints nothing outside tests that set a mode to prove it is ignored; scripts/check-pr.sh PASS.

**F-12 — Open picks the app from the type, not from the file** (from SEC-1; user answer 2026-10-28)
Files: `macos/App/Platform/ProjectFiles.swift`, `macos/ProvenenciaTests/ProjectFilesTests.swift`, the plan's Landed note.
Failing test first: `openingApplicationIgnoresPerFileOverride` *(new)*: write `objects/aa/bb/aabb.txt` at `0o644`, set `com.apple.LaunchServices.OpenWith` on it with `setxattr` to a binary plist (`PropertyListSerialization`, `.binary`) of `["bundleidentifier": "com.apple.Preview", "path": "/System/Applications/Preview.app", "version": 0]`; `ProjectFiles.openingApplication(for:)` returns `NSWorkspace.shared.urlForApplication(toOpen: UTType.plainText)` and not Preview's URL. (Fails to compile until the helper exists.) Do not launch anything in a test.
Do: add `static func openingApplication(for url: URL) -> URL?` (internal, for the test): `UTType(filenameExtension: url.pathExtension)` then `NSWorkspace.shared.urlForApplication(toOpen: type)`. `openObject` runs `openableObjectURL`; on `.success(url)` it asks `openingApplication(for:)`; nil → `.missing` (as today's `false` from `open`); otherwise `NSWorkspace.shared.open([url], withApplicationAt: app, configuration: NSWorkspace.OpenConfiguration())` with no completion handler, and returns nil. Its signature and both callers stay synchronous. Doc comment: the app comes from the type, so nothing stored on the file (an Open With attribute) chooses it. Landed note: D-7's `NSWorkspace.shared.open(url)` became open-with-the-type's-app (review SEC-1, user answer 2026-10-28).
Check: `rg -n 'NSWorkspace.shared.open\(url\)' macos/App/Platform/ProjectFiles.swift` prints nothing; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `099690ce` (twelve commits `5ee57050`…`099690ce` on top of the orchestrator's doc commit `9f4ad32c`). Gate PASS (orchestrator). F-1 to F-12 done; every red-first or mutation check reported failing before the fix and passing after. Deviations:
- F-2 Check: `originalFilename.isEmpty` / `?? ""` still hits thumbnail and glyph inputs (`PVArtifactChecklist.swift:135`, `SourcePageArtifactsView.swift:186`, `CachedThumbnail.swift:120`), a label-first title rule (`CitationComposerEvidenceColumn.swift:375`) and generated code. None names a file, so all were left as they are.
- F-2: `previewLabel` now falls back to the path's last component before the ref.
- F-3: `openableRelPath` also needs a non-empty `fileID`, so a file ref with an empty `fileID` now gets the refusal.
- F-10: a picked `.webloc`, `.fileloc`, `.inetloc`, `.url`, `.terminal` or `.term` is now rejected at ingest as `.executable` ("Programs and installers can’t be attached") instead of `.genericType`. Recorded in the Landed note.
- Sweep hits that no F-item named, now fixed: `.agents/skills/review-pr/lenses/security.md:105-109` (the known trap now says to pick the app from the type and that the execute bit alone is no signal); `deployment-plan.md` S12-23 map row (line 183) and S12-23's **Check** (line 507).
- F-7 kept inline: `artifacts_test.go:50`, `:72` (fileless then ingest; asserts its own label), `citations_test.go:120` (`likenessFixture`), and the PDF and TXT ingest sites.
- F-11: `objectstore.go` is byte-identical to base.

## Round 2 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF (each had a round-1 Fix-now blocker or major). Scope `acfe12bb...099690ce`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1, CON-3 | UX, CON | blocker | (round 1) execute-bit refusal on exFAT/FAT | Resolved | `objectstore.go` byte-identical to base; no `0o111` in `ProjectFiles`; `TestOpenObjectIgnoresExecuteBit`, `objectURLAcceptsExecuteBit` |
| MNT-1 | MNT | blocker | (round 1) one read-side path rule | Resolved | `ErrDisallowedType` only in `objectpath.go` and `files.ReadRelPath`; CON's mutation of `existingSourceCoverThumb` turns `TestSearchHitHidesDamagedCoverThumbnail` red |
| MNT-2, UX-4 | MNT, UX | blocker | (round 1) one file-name rule | Resolved | `CatalogArtifact.fileDisplayName` used by `open`, `previewLabel`, `Row.displayFilename` |
| SEC-1 | SEC | major | (round 1) Open With attribute picks the app | Resolved | `openingApplication(for:)` reads only `UTType(filenameExtension:)`; only object-open call is `ProjectFiles.swift:129`; CON's read-only Launch Services query: all 26 stored extensions resolve an app by type and none is flagged |
| UX-2 | UX | major | (round 1) exec-bit refusal copy | Resolved | every remaining `.unsafe` cause is the type; viewer no longer sees nil for a mode bit |
| MNT-3, MNT-4/CON-1, CON-2 | MNT, CON | major | (round 1) `openableRelPath`, skill line, guards | Resolved | CON mutated each guarded branch in a scratch copy: each guard went red |
| MNT-5, CON-4, CON-5, SEC-2, SEC-3 doc | — | minor | (round 1) | Resolved | read in round 2 by their lenses |
| MNT-6 | MNT | minor | (round 1) handler-test fixtures | Not resolved (remainder) | `thumbnails_test.go:205-223` still hand-rolls `DELETE FROM file_derivatives` (`sources_test.go:374` uses `damageCatalog`). Follow-up |

No new blockers. No Fix plan 2.

Noted in round 2, not raised (round 2 raises blockers only):
- UX, CON, MNT: since F-10, ingest rejects a picked `.webloc`, `.fileloc`, `.inetloc`, `.url`, `.terminal` or `.term` with "Programs and installers can't be attached" (it was `.genericType`). That copy names the wrong concept; by the UX rubric it would be a major. **To user** at hand-off: split the open-another-item types out of the ingest `.executable` branch (they fall through to `.genericType`) and keep them in the Open check.
- CON, MNT: F-12 opens with no completion handler. A Launch Services failure after an app is found reports nothing; before this PR, `open(url) == false` showed the missing-file message. Follow-up.
- UX, CON: plan §10 step 5 still expects a refusal for a `chmod +x` object. After F-11 it opens in the app for its type. The PR body's manual check carries the corrected step.

Nits (logged): `MNT-10 · ProjectFiles.swift:33,72 · checkedObject / objectRefusal only produce .missing; Result and ObjectRefusal? shapes are vestigial` · `MNT-11 · ProjectFiles.swift:16-17,95-97 · .unsafe and openableObjectURL docs should add "or open another item"` (SEC also flagged it) · `MNT-12 · IngestMediaPolicy.swift:157 · isExecutableType now also covers links; name could say so` · `CON-8 · ProjectFilesTests.swift:~255 · openingApplicationIgnoresPerFileOverride pins the helper, not openObject; the F-12 rg check is the guard`

## Result
Gate PASS (`099690ce`; `origin/main` unmoved since the branch was cut) · CI pending · open: none at blocker or major · to user: ingest copy for link and Terminal files (round-2 note) · follow-ups: SEC-3 (check a search hit's stored thumbnail path on read), MNT-6 remainder (`thumbnails_test.go:205-223`), Open reports no Launch Services failure after F-12, MNT-10 to MNT-12 · CON miss: none
