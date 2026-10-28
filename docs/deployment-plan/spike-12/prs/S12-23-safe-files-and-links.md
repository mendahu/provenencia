# S12-23 — The app opens only safe files and links

**Status:** Landed in [#608](https://github.com/mendahu/provenencia/pull/608) (2026-10-28)
**Landed:** Differences from the plan: `ProjectFiles.ObjectRefusal` also conforms to `Error`, which `Result<URL, ObjectRefusal>` needs. Beyond T-1 – T-18 there is one more Swift test (`openObjectReportsUnsafeForExecutableObject`) and the Go test helpers `sourceWithPNG`, `damageCatalog`, `urlMetadataSetFixture` and `writeObject`. STEP-7's check `rg -n 'bare hex' docs core` still hits this plan and its review log, which name the phrase; no code or other doc holds it. `add-ffi-handler` still tells handlers to return one path through `files.StorageRelPath`; reads now use `files.ReadRelPath`, and the plan did not name that line. D-3's "`searchindex` cannot import `files`" was false: `searchindex` and `derivatives.Ensure` call `files.ReadRelPath`, so `objectsRelPath` and its test are gone (review MNT-1). D-10's "label" fallback became the shared `CatalogArtifact.fileDisplayName` (original name, else the path's last component, else the ref), which the Source page's Open and its preview label and the canvas's `Row.displayFilename` all use (review MNT-2, UX-4). Otherwise none.
**Spike / slice:** Spike 12 · Slice 5 — Audit fixes: High
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the refusal reuses the Source page's error banner and the canvas's file callout, with new copy only)
**Board:** —
**Depends on:** —
**Blocks:** S12-28 (map: "After S12-23, whose `files` filter the viewer relies on")
**Branch:** `s12-23-safe-files-and-links`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-38 did; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`
**Context:** [app-health audit](../app-health-audit.md) A-1, A-5 (and L-6, left out); working rules "What the app hands to macOS to open is checked first, and Go decides the type", "FakeStore is a seeded fake, not a second engine", "High tolerance for churn"; [map review](../deployment-plan.review.md) K-8; [`artifact-file-storage.md`](../../../artifact-file-storage.md); [`application-stack.md`](../../../application-stack.md) "Read path" and "Object names"; [`source-layer-data-model.md`](../../../source-layer-data-model.md) (object files paragraphs)
**Planned:** 2026-10-28 against `main` @ `7adb3bc`
**Review:** [`S12-23-safe-files-and-links.review.md`](S12-23-safe-files-and-links.review.md) — Fits with changes / Ready with changes, 1 round; every blocker Fixed

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

A project folder someone shares can no longer make Provenencia launch a program. Today a crafted catalog can give a stored file any `media_type`; Go then returns an object path with no extension, and Open hands it to `NSWorkspace`, which runs an extensionless executable in Terminal (audit A-1). A crafted metadata URL value opens with any scheme, `file:` and `smb:` included (audit A-5). After this PR, the ingest allowlist is the one table of stored types and their extensions, Go returns an object path only for a type on it, both layers refuse an object with an execute bit, Swift refuses to open an object with no extension or of a type that could run a program and names the file, and a metadata link opens only through the web link Go computes.

**Why now.** Bar outcome 11 ("A shared project is safe to open"), slice 5. The audit rated both items High.

**What stays.** Ingest's accepted types, how Artifacts and metadata look, Open's behavior on every allowed file, and the missing-file messages. The viewer's decode checks are S12-28.

## 2. User stories

**US-1.** As a researcher who opens a project someone sent me, when I open an Artifact whose stored file type is not one Provenencia stores, the app refuses and names the file, so that nothing runs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A `files` row damaged to `media_type = 'application/x-sh'`; `GetSourceWorkspace` on its Source | The read succeeds; the Artifact's `file` carries its `original_filename` and `media_type` and an empty `rel_path` |
| AC-1.2 | `objectpath.Rel(sum, t)` for `t` in `application/x-sh`, `application/octet-stream`, `""`, `image/jpg` | Each returns `objectpath.ErrDisallowedType`; `image/jpeg` returns `objects/…/<sum>.jpg`; `text/plain; charset=utf-8` returns `….txt` |
| AC-1.3 | `files.RelPaths` over a PNG row and a row inserted with `application/x-sh` | Only the PNG's id is in the map |
| AC-1.4 | On the Source page, Open on that Artifact's tile (empty `relPath`, file `run.sh`) | The page error reads `Provenencia won't open “run.sh”. Its file type could run a program, or it isn't a type Provenencia stores.` |
| AC-1.5 | `ListSources` on a project whose cover thumbnail's `files` row is damaged to `application/x-sh` | The read succeeds; that Source's `thumbnail_rel_path` is empty |
| AC-1.6 | `mediatypes.StoredExtension(t)` | `image/jpeg` → `(".jpg", true)`; `IMAGE/PNG` → `(".png", true)`; `audio/x-wav` → `(".wav", true)`; `image/jpg`, `application/zip`, `""` → `("", false)`; every type `Allowed` accepts has a non-empty extension |

**US-2.** As a researcher, when a stored object has an execute bit or a dangerous type on disk, the app refuses to open it and names the file.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `objectstore.OpenObject` on a regular object with mode `0o755` | Returns an error for which `errors.Is(err, objectstore.ErrExecutable)` and `errors.Is(err, objectstore.ErrNotRegular)` both hold |
| AC-2.2 | `objectstore.WriteIfAbsent` when the object already exists with the right bytes and mode `0o755` | `wrote` is true and the object's mode has no execute bit |
| AC-2.3 | `ProjectFiles.objectURL` on a regular `.png` object with mode `0o755` | `nil` |
| AC-2.4 | `ProjectFiles.openableObjectURL` on an existing object `objects/aa/bb/aabb` (no extension) at mode `0o644` | `.failure(.unsafe)` |
| AC-2.5 | `ProjectFiles.openableObjectURL` on existing regular files `objects/aa/bb/aabb.sh`, `aabb.command` and `aabb.app` at mode `0o644` | `.failure(.unsafe)` for each |
| AC-2.6 | `ProjectFiles.openableObjectURL` on an existing `objects/aa/bb/aabb.pdf` with mode `0o755` | `.failure(.unsafe)` |
| AC-2.7 | `ProjectFiles.openableObjectURL` on an existing `objects/aa/bb/aabb.pdf` at `0o644`; on a missing one; on a symlinked one | `.success(url)`; `.failure(.missing)`; `.failure(.missing)` |
| AC-2.8 | On the canvas Files pane, Open on a row whose object has an execute bit | The callout titled "Can't open the file" shows the AC-1.4 text with that row's file name (manual only, §10 step 5) |
| AC-2.9 | `IngestMediaPolicy.isExecutableType` for `.shellScript`, `.unixExecutable`, `.applicationBundle`, `.pdf` | `true`, `true`, `true`, `false` |

**US-3.** As a researcher, when I click a saved metadata link, it opens only on the web.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `urlshape.OpenHref` on `file:///System/Applications/Calculator.app`, `smb://server.example.org/share`, `vnc://host.example.org` | `ErrInvalid` for each |
| AC-3.2 | `GetSourceWorkspace` with a url-type field whose value is `www.example.org` | That entry's `open_href` is `https://www.example.org` |
| AC-3.3 | A url-type value damaged by SQL to `file:///System/Applications/Calculator.app` | `GetSourceWorkspace` returns the entry with that `value_text` and an empty `open_href` |
| AC-3.4 | A text-type field with value `www.example.org` | `open_href` is empty |
| AC-3.5 | `SetSourceMetadata` on a url field with `example.org/records` | The returned entry's `open_href` is `https://example.org/records` |
| AC-3.6 | `CatalogMetadataEntry.openURL` with `openHref` `""`; with `https://example.org` | `nil`; `URL(string: "https://example.org")` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `mediatypes` owns one table of allowed type → stored extension; `objectpath.Extension` goes | STEP-1, AC-1.6 |
| `objectpath.Rel` refuses a type off that table (A-1) | STEP-1, AC-1.2 |
| One read-side path function in `files`; every read maps the refusal to "no path" | STEP-2, AC-1.1, AC-1.3, AC-1.5 |
| `objectstore.openRegular` refuses an execute bit (A-1) | STEP-3, AC-2.1, AC-2.2 |
| `sourcemetadata.WorkspaceEntry.OpenHref` from `urlshape.OpenHref`, on the wire as `MetadataWorkspaceEntry.open_href` (A-5) | STEP-4, AC-3.1 – AC-3.5 |
| Swift refuses an execute bit in `objectURL`; `openableObjectURL` refuses no extension and executable types through one shared predicate; `openObject` reports why (A-1) | STEP-5, AC-2.3 – AC-2.7, AC-2.9 |
| The Source page and the canvas name a refused file (A-1) | STEP-5, AC-1.4, AC-2.8 |
| Swift opens only `open_href`; `openSavedURL` goes (A-5) | STEP-6, AC-3.6 |
| Docs and rules that describe object paths and the read path | STEP-7 |

| Out | Home |
| --- | --- |
| Links inside PDFs (`ArtifactPDFViewport` has no `PDFViewDelegate`; audit L-6) | [audit, Not scheduled](../app-health-audit.md#not-scheduled) |
| An App Sandbox | not planned (map scope entry) |
| The viewer's sniff-vs-media-type and pixel-budget checks | S12-28 |
| `raster.Supported`'s own normalize copy and its `image/jpg` alias | not planned (review log M-10; S12-28's planner reads it there) |
| Listing a refused file in the canvas Files pane (it lists only Artifacts with a path, as today) | not planned |
| Removing the bytes of a damaged row on purge (`PurgeUnreferenced` skips a path it cannot build) | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Allowlist | `core/ingest/mediatypes/mediatypes.go` `allowed` (a set), `extAllowed` (extension → type), `Allowed`, `Normalize` | The ingest allowlist; imports only `path/filepath` and `strings`. Checked on write only |
| Object path | `core/objectpath/objectpath.go` `Extension`, `Rel` | `Extension` is a third hand-kept type → extension switch with its own copy of `Normalize`; it maps `image/jpg`, which the allowlist does not hold. Its only non-test caller is `Rel`. `Rel` builds `objects/hh/hh/<hex><ext>`; an unknown type gives no extension. `files` and `searchindex` both call `Rel` |
| Files path | `core/database/files/files.go` `StorageRelPath`, `RelPaths` | `StorageRelPath` passes `objectpath.Rel` through (maps only `ErrInvalidChecksum`); its doc says unknown types keep a bare hex name. `RelPaths` omits an id whose path fails |
| Purge | `core/database/files/purge.go` `PurgeUnreferenced` | `continue` when `StorageRelPath` fails |
| Artifact read | `api/ffi/handlers/artifacts.go` `artifactProto` | `files.Lookup` then `StorageRelPath`; returns the error, which fails the whole `GetSourceWorkspace` |
| Thumbnails | `api/ffi/handlers/thumbnails.go` `thumbnailRelPath`, `existingThumbnailRelPath` | Each `files.Lookup` then `StorageRelPath`, returning its error; `existingThumbnailRelPath` feeds `sourceCoverThumbnail` |
| Sources list cover | `core/database/searchindex/searchindex.go` `objectsRelPath` (called by `existingSourceCoverThumb`) | Wraps any `objectpath.Rel` error as "bad checksum", which fails the search document write. `searchindex` reads `files` with its own SQL, so it cannot call `files` |
| Derivatives | `core/derivatives/ensure.go` `Ensure` | Skips types `raster.Supported` rejects, then calls `StorageRelPath` for the source; `raster.supportedMIME` holds `image/jpg`, which is off the allowlist |
| Other path callers | `core/ingest/file.go` (`mediatypes.Resolve` output), `core/ingest/stage.go` (`image/png`), `core/derivatives/ensure.go` (`image/jpeg` derived) | Always allowed types; keep `StorageRelPath` |
| Object open (Go) | `core/objectstore/objectstore.go` `openRegular`, `ErrNotRegular`, `matches` | Refuses a non-regular file or a symlinked shard; no mode-bit check. `matches` treats `ErrNotRegular` as "no match", so `installTemp` replaces the entry |
| Object open (Swift) | `macos/App/Platform/ProjectFiles.swift` `objectURL`, `isSafeObject`, `openObject` | `objectURL` has six early `return nil` guards (empty, `.`/`..`, absolute-ish segment, not under `objects/`, prefix) then `isSafeObject`, which checks folders, regular file and Finder alias, not the execute bit. `openObject` returns `Bool` and passes any URL `objectURL` accepts to `NSWorkspace.shared.open` |
| Executable types (Swift) | `macos/App/Platform/IngestMediaPolicy.swift` `validate(url:)` | Rejects a picked file as `.executable` when its type conforms to `.executable`, `.unixExecutable` or `.application`, inline; no `.script` |
| Source page Open | `macos/App/Features/Sources/SourceArtifactsSection.swift` `open(_:)` | Returns silently on an empty `relPath`; on `false` sets `context.pageError` to `L10n.Sources.fileOpenMissing` |
| Canvas Open | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `openSelectedFile`; `EvidenceCanvasInputMode.swift` `EvidenceGraphCanvasNotice.fileNotFound`; `EvidenceGraphView.swift` (callout titled `L10n.EvidenceGraph.fileNotFoundTitle`) | On `false` sets `fileOpenError` to `L10n.EvidenceGraph.fileNotFoundMessage(name:filename:)`. Rows come from `EvidenceGraphFiles.fileBearing`, which drops an Artifact with no `openableRelPath` |
| Metadata domain | `core/database/sourcemetadata/sourcemetadata.go` `WorkspaceEntry`, `ListWorkspace`, `normalizeValue` | `normalizeValue` is the one place that knows a url-typed value (`metadatafields.DataTypeURL` → `urlshape.Canonical`); `WorkspaceEntry` has no link |
| Metadata wire | `api/proto/engine.proto` `MetadataWorkspaceEntry` (fields 1, 2, 4, 5, 6; reserved 3, 7, 8); `api/ffi/handlers/sources.go` `metadataEntryProto` (four call sites, including `SetSourceMetadata`'s returned entry; the only builder of the message) | No open link; `urlshape.OpenHref` has no caller |
| URL shape | `core/urlshape/urlshape.go` `OpenHref` | Already returns `ErrInvalid` for any scheme but `http` / `https` and adds `https://` to a scheme-less value |
| Metadata link (Swift) | `macos/App/Features/Sources/SourcePageMetadataView.swift` `displayValue`, `openSavedURL` | Prefixes `https://` when the text has no `://`, then opens any scheme through `openURL` |
| Metadata model (Swift) | `macos/App/Platform/GenealogyStore.swift` `CatalogMetadataEntry` (25 initializer call sites); `GoStore.swift` `mapMetadataEntry`; `FakeStore.swift` `setSourceMetadata`; `SourceMetadataSection.swift` `applyCleared` | No href field; `applyCleared` blanks `valueText` on a suggested row |
| Go tests that this PR breaks | `core/objectpath/objectpath_test.go` `TestRel` (`image/jpg` alias), `TestExtension`; `core/database/files/files_test.go` `TestStorageRelPath` ("unknown" and "empty" rows expect a bare path); `core/database/searchindex/searchindex_test.go` `TestObjectsRelPathAgreesWithObjectpath` (loops `image/jpg`, `application/octet-stream`, `""` and fatals on any error); `core/derivatives/thumbnail_test.go` `TestEnsureThumbnail…` skip case (`installSourceFile(…, "application/octet-stream", "blob.bin")` fatals on the path error) | Each is changed in STEP-1 or STEP-2 as stated there |
| Other tests | `core/database/files/files_test.go` `TestRelPaths`; `core/objectstore/objectstore_unix_test.go`; `core/urlshape/urlshape_test.go` `TestOpenHref`; `api/ffi/handlers/sources_test.go` `TestGetSourceWorkspace`, `TestSetSourceMetadata`, `TestListSources`; `macos/ProvenenciaTests/ProjectFilesTests.swift` (`openObjectReturnsFalseFor…`); `ArtifactFileFixtures.swift` `write(_:projectDir:relPath:)`; `EvidenceGraphModelTests.swift` (`.fileNotFound` match near line 977); `SourcePageModelTests.swift`; `MetadataModelTests.swift` | Extended, not rewritten |

**Precedent.** S10-62 ([#464](https://github.com/mendahu/provenencia/pull/464)), commits `200c4687` and `6cd8c2b1`: it added the symlink and Finder-alias refusals to `openRegular` and `isSafeObject` together, with `ArtifactFileFixtures` helpers and `ProjectFilesTests` cases, and updated `.cursor/rules/macos-client.mdc`, `application-stack.md` "Read path" and `source-layer-data-model.md` in the same PR. Copy that shape: one rule in both mirrors, the same fixtures file, the same docs. For the damaged-row Go tests, copy the `UPDATE files SET media_type = …` damage pattern from `api/ffi/handlers/citations_test.go` (its "image not a PNG" row).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `mediatypes` holds one table, `allowed map[string]string` (type → stored extension with its dot), replacing the `allowed` set. `Allowed(mt)` reads it. `StoredExtension(mt string) (string, bool)` *(new)* normalizes and returns the entry. `objectpath.Extension` and its normalize copy are deleted; `objectpath` imports `core/ingest/mediatypes` | One list decides both "may we store it" and "what it is called on disk"; the next allowed type is one entry. `Extension`'s only caller is `Rel`, which this PR rewrites. `mediatypes` imports only the standard library, so no cycle | Keeping `Extension` and checking `Allowed && Extension != ""` — two lists that must agree, already drifted on `image/jpg` |
| D-2 | `objectpath.Rel` checks the checksum, then returns `ErrDisallowedType` *(new)* when `StoredExtension` is not ok | `Rel` is the one place every object path is built, `files` and `searchindex` both; a disallowed type never gets a path, so every path has an extension | Filtering in `files.scanFile` / `Lookup` — `searchindex.existingSourceCoverThumb` reads `files` with its own SQL, and dropping rows would break `LookupByChecksum` dedupe for a damaged row |
| D-3 | `files.ReadRelPath(f File) (string, error)` *(new)* is the read-side path: it returns `("", nil)` for `ErrDisallowedType` and other errors as `StorageRelPath` does. `RelPaths`, `artifactProto`, `thumbnailRelPath` and `existingThumbnailRelPath` call it in place of `StorageRelPath`. `artifactProto` returns the `SourceFileRef` with `rel_path` empty. `thumbnailRelPath` returns `("", true, nil)` when the derived path is empty. `searchindex.objectsRelPath` maps `ErrDisallowedType` to `("", nil)` itself (it cannot import `files`). `derivatives.Ensure` maps it to `ErrUnprocessable` (it reads source bytes, so it keeps `StorageRelPath`). Ingest, stage and purge keep `StorageRelPath` | A damaged row hides one file instead of failing a page or list read, and the next read that returns a path gets the refusal by calling one function. Keeping the file ref (name, type, size) lets Swift name the file it refuses (bar outcome 11) | An `errors.Is` branch at each handler — five copies. Omitting `file` from the Artifact — Swift could not name the file. A wire "refused" flag — empty `rel_path` already means "no openable object" in Swift (`openableRelPath`) |
| D-4 | Go refuses an execute bit in `openRegular` with `ErrExecutable` *(new)* = `fmt.Errorf("%w: execute bit set", ErrNotRegular)`, checked after both `IsRegular` checks (the `Lstat` and the opened file's `Stat`) as `Mode().Perm()&0o111 != 0` | Wrapping `ErrNotRegular` keeps every caller that handles it (`matches`, `derivatives.readSourceObject`) correct with no edit, and `matches` returning false makes `installTemp` replace such an object with a fresh `0o600` file on re-import | A new unrelated sentinel — every caller would need a second `errors.Is` |
| D-5 | Swift: `ProjectFiles.ObjectRefusal` *(new)* `enum { case missing, unsafe }`. `objectURL`'s body moves into `private static func checkedObject(projectDir:relPath:) -> Result<URL, ObjectRefusal>` *(new)*; its six early guards return `.failure(.missing)`. `isSafeObject` is renamed `objectRefusal(segments:projectURL:) -> ObjectRefusal?`: an execute bit (`st.st_mode & 0o111 != 0` on the object) is `.unsafe`; every refusal it has today stays `.missing`. `objectURL` becomes `try? checkedObject(…).get()`, signature unchanged | A symlink or alias stays "missing", as S10-62 decided; only the new refusal needs its own message. Thumbnails, the viewer and the PDF crop gain the execute-bit refusal through `objectURL` with no edit | Making every existing refusal `.unsafe` — changes S10-62's copy decision |
| D-6 | `IngestMediaPolicy.isExecutableType(_ type: UTType) -> Bool` *(new)*: true when the type conforms to `.executable`, `.unixExecutable`, `.application` or `.script`. `validate(url:)` calls it in place of its inline check | One Swift answer to "could this run a program", for ingest and open alike. Ingest gains `.script`, which Go's allowlist already refuses, so no accepted file changes | A second inline conformance list in `ProjectFiles` — two lists that already differ (`.script`, `.unixExecutable`) |
| D-7 | `ProjectFiles.openableObjectURL(projectDir:relPath:) -> Result<URL, ObjectRefusal>` *(new)* runs `checkedObject`, then returns `.failure(.unsafe)` when `url.pathExtension` is empty or `UTType(filenameExtension:)` is non-nil and `IngestMediaPolicy.isExecutableType` holds. `openObject(projectDir:relPath:)` returns `ObjectRefusal?` (nil when opened): it calls `openableObjectURL`, then `NSWorkspace.shared.open`; a `false` from `open` is `.missing` | Working rule: Swift checks extension, execute bit and executable types only; Go owns the allowlist. A pure check is testable without launching an app | A Swift allowlist — the working rule forbids it (map review K-8). Keeping `Bool` — callers could not tell missing from refused |
| D-8 | Copy: one shared refusal message `L10n.ObjectOpen.refused(filename:) -> String` *(new namespace and function)*, key `objectOpen.refused.message`, built with `L10n.format` like `fileNotFoundMessage`. The Source page shows it in `context.pageError` for an Artifact with a `file` and an empty `relPath`, and for `.unsafe`; `.missing` keeps `L10n.string(L10n.Sources.fileOpenMissing)`. The canvas shows it in `fileOpenError` for `.unsafe`; `.missing` keeps `fileNotFoundMessage` | One rule, one string, in a namespace both features read. `L10n.Errors` is reserved for FFI codes (`add-localized-string` step 8), and keys follow `<feature>.<area>.<name>` (step 1) | `L10n.Errors.…` with an `errors.` key — breaks the skill. Two copies in `Sources` and `EvidenceGraph` |
| D-9 | The canvas callout gets one title for both cases: `L10n.EvidenceGraph.fileNotFoundTitle` is renamed `fileOpenFailedTitle` (key `evidenceGraph.fileOpenFailed.title`) = "Can't open the file", and `EvidenceGraphCanvasNotice.fileNotFound` is renamed `.fileOpenFailed`. `fileOpenError` stays `String?` | "File not found" is false for a refusal; one neutral title keeps the notice a single string with no new type | A title per refusal kind — a struct in `fileOpenError` and the notice for no researcher gain |
| D-10 | The Source page names the file by `file.originalFilename`, or the Artifact's `label` when that is empty; the canvas uses `EvidenceGraphFiles.Row.displayFilename` | The names each surface already shows | — |
| D-11 | `sourcemetadata.WorkspaceEntry` gains `func (e WorkspaceEntry) OpenHref() string` *(new)*, beside `normalizeValue`: `""` unless `e.Value != nil` and `e.Field.DataType == metadatafields.DataTypeURL`, else `urlshape.OpenHref(e.Value.ValueText)` with an error giving `""`. `MetadataWorkspaceEntry` gains `string open_href = 9;`, and `metadataEntryProto` copies `e.OpenHref()` | Which value is a web link is a domain rule; it lives in the package that already decides what a url-typed value is, and handlers stay thin (`ffi-handlers.mdc`). A method covers every place a `WorkspaceEntry` is built | Computing it in `metadataEntryProto` — domain logic in a handler; a struct field filled in `ListWorkspace` — two construction sites to keep in step |
| D-12 | Swift: `CatalogMetadataEntry.openHref: String = ""` *(new)* and a computed `openURL: URL?` *(new)* = `openHref.isEmpty ? nil : URL(string: openHref)`. `displayValue` shows the link button when the field is a url field and `openURL` is non-nil, opening that URL; otherwise it shows the plain-text branch. `openSavedURL` is deleted. `applyCleared` blanks `openHref` with `valueText` | Go alone decides what is a link; the default value keeps the 25 initializers unchanged | A data-type or scheme check in `openURL` — a Swift copy of Go's rule |
| D-13 | `FakeStore.setSourceMetadata` sets `openHref = ""` on the entry it writes, new or existing; it computes no href. A test that needs a link seeds `openHref` | Working rule "FakeStore is a seeded fake, not a second engine"; resetting on update leaves no stale link | A Swift copy of `OpenHref` in the fake |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `mediatypes.StoredExtension` *(new)* | func | `objectpath.Extension` (deleted), `mediatypes.Allowed` | Replaces `Extension`; reads the same table as `Allowed` |
| `objectpath.ErrDisallowedType` *(new)* | error | `objectpath.ErrInvalidChecksum` | Same package, beside it; a different refusal |
| `files.ReadRelPath` *(new)* | func | `files.StorageRelPath`, `files.RelPaths` | `StorageRelPath` stays the write and purge path; `RelPaths` now calls `ReadRelPath` |
| `objectstore.ErrExecutable` *(new)* | error | `objectstore.ErrNotRegular` | Wraps it (D-4) |
| `sourcemetadata.WorkspaceEntry.OpenHref` *(new)* | method | `sourcemetadata.normalizeValue` | Same package and same url rule; a read, not a write normalization |
| `MetadataWorkspaceEntry.open_href` *(new)* | wire field | `SourceFileRef.rel_path` (Go computes, Swift uses) | Same pattern on another message |
| `ProjectFiles.ObjectRefusal`, `checkedObject`, `openableObjectURL` *(new)* | type, funcs | `ProjectFiles.objectURL`'s `nil` | `nil` cannot say why; searched `macos/App` for an open-result type (`openObject`, `OpenResult`, `Refusal`): none. `objectURL` stays the read path for thumbnails and the viewer, which must not refuse by extension |
| `IngestMediaPolicy.isExecutableType` *(new)* | func | `IngestMediaPolicy.validate(url:)`'s inline check | Extracted from it, so ingest and open share it |
| `CatalogMetadataEntry.openHref`, `openURL` *(new)* | Swift fields | `CatalogArtifact.openableRelPath` | Same pattern: the model says what is openable |
| `L10n.ObjectOpen.refused(filename:)` *(new)* | L10n | `L10n.Sources.fileOpenMissing`, `L10n.EvidenceGraph.fileNotFoundMessage` | Those say "missing"; the refusal is a different fact, shared by both surfaces |
| `ArtifactFileFixtures.write(_:projectDir:relPath:mode:)` | test helper | `ArtifactFileFixtures.write(_:projectDir:relPath:)` | Extended with a `mode: mode_t = 0o644` parameter (applied with `chmod` after the write); no new helper |

**Builds on.** `mediatypes` (the one allowlist), `objectpath.Rel` (the one path builder), S10-62's `openRegular` / `isSafeObject` mirror, `IngestMediaPolicy`'s executable check, `urlshape.OpenHref`, `sourcemetadata`'s url rule.
**Deletes.** `objectpath.Extension` and `TestExtension` (its rows move to `TestStoredExtension`); the `mediatypes.allowed` set (becomes the table); `SourcePageMetadataView.openSavedURL`; the `Bool` return of `ProjectFiles.openObject`; `ProjectFiles.isSafeObject` (renamed); the inline executable check in `IngestMediaPolicy.validate`; `L10n.EvidenceGraph.fileNotFoundTitle` and its `evidenceGraph.fileNotFound.title` catalog entry (renamed, D-9); `EvidenceGraphCanvasNotice.fileNotFound` (renamed); `TestRel`'s `image/jpg` alias assertion, `TestStorageRelPath`'s bare-path rows and `TestObjectsRelPathAgreesWithObjectpath`'s "every type gives a path" loop (their meaning is reversed); every doc line that says an unknown type keeps a bare hex name.
**Next adds.** S12-28's viewer and PDF crop reach objects through `objectURL` and inherit the execute-bit refusal; its sniff check relies on Go returning only allowed types. A new allowed type is one `mediatypes.allowed` entry. Audit L-6 (PDF link annotations) would route a link through the same Go-computed web link rule.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An Artifact with a `file` and an empty `rel_path` | Source page tile (`SourcePageArtifactsView`), card, `SourceArtifactsSection.open` | Tile already paints the type icon when `thumbnailRelPath` is empty; `open` now shows the refusal (AC-1.4) instead of returning silently |
| Same Artifact | Canvas Files pane (`EvidenceGraphFiles.fileBearing`) | Dropped, as an Artifact with no path is today (Out) |
| Same Artifact as a Source cover | Sources list row, Source header | `thumbnail_rel_path` empty → the type icon, as for a missing derivative (AC-1.5) |
| A metadata url entry with an empty `open_href` | `SourcePageMetadataView.displayValue` | Plain text, as a text field shows (D-12) |

### Scale

None — no write, list or loop changes; `OpenHref` is one string function per metadata row on reads that already build the row, and the allowlist check is one map lookup per path.

### Schema

None — no catalog change. The filter is on read; no migration rewrites stored types.

### Go core

| Package | Change |
| --- | --- |
| `core/ingest/mediatypes` | `allowed` becomes `map[string]string` (type → `.ext`): every key it holds today, each with the extension `objectpath.Extension` gives it (`image/tif` → `.tiff`, `application/csv` → `.csv`, `text/x-markdown` → `.md`, `audio/wave` and `audio/x-wav` → `.wav`, `audio/x-m4a` → `.m4a`, `audio/x-aiff` → `.aiff`). `Allowed` reads it. `func StoredExtension(mediaType string) (string, bool)` *(new)* |
| `core/objectpath` | `Extension` deleted. `var ErrDisallowedType = errors.New("objectpath: media type not allowed")` *(new)*. `Rel` per D-2; its doc and the package doc say a disallowed type has no path |
| `core/database/files` | `func ReadRelPath(f File) (string, error)` *(new)* per D-3; `RelPaths` calls it and omits an empty path. `StorageRelPath`'s doc: a disallowed type returns `objectpath.ErrDisallowedType`; no bare hex name |
| `core/database/searchindex` | `objectsRelPath`: `errors.Is(err, objectpath.ErrDisallowedType)` → `("", nil)`; other errors as today |
| `core/derivatives` | `Ensure`: the source `StorageRelPath` error `ErrDisallowedType` → `ErrUnprocessable` |
| `core/objectstore` | `var ErrExecutable = fmt.Errorf("%w: execute bit set", ErrNotRegular)` *(new)*; `openRegular` per D-4. `ErrNotRegular`'s and `openRegular`'s doc comments name it, and the "mirrors" comment names the execute bit |
| `core/database/sourcemetadata` | `func (e WorkspaceEntry) OpenHref() string` *(new)* per D-11 |
| `core/urlshape` | No code change; `TestOpenHref` gains rows (T-7) |

Errors: none reach the FFI. `ErrDisallowedType` is mapped by `ReadRelPath`, `objectsRelPath` and `Ensure` (D-3); `ErrExecutable` surfaces as today's `ErrNotRegular` handling.

### FFI

`api/proto/engine.proto`: `MetadataWorkspaceEntry` gains `string open_href = 9; // web link to open, from sourcemetadata OpenHref; empty when the value is not one`. Regenerate with `./scripts/generate-proto.sh`.

| Handler | Change |
| --- | --- |
| `api/ffi/handlers/sources.go` `metadataEntryProto` | `OpenHref: e.OpenHref()` |
| `api/ffi/handlers/artifacts.go` `artifactProto` | `files.ReadRelPath(f)` in place of `StorageRelPath`; `out.File = fileRefProto(f, rel)` with `rel` possibly empty; still asks `thumbnailRelPath` |
| `api/ffi/handlers/thumbnails.go` `thumbnailRelPath`, `existingThumbnailRelPath` | `files.ReadRelPath(f)` in place of `StorageRelPath`; `thumbnailRelPath` returns `skipped` true when the path is empty |

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Object checks | `ProjectFiles.ObjectRefusal`, `checkedObject`, `objectRefusal(segments:projectURL:)`, `objectURL`, `openableObjectURL`, `openObject` | `macos/App/Platform/ProjectFiles.swift` | D-5, D-7. `import UniformTypeIdentifiers` is already there |
| Executable types | `IngestMediaPolicy.isExecutableType` | `macos/App/Platform/IngestMediaPolicy.swift` | D-6 |
| Source page Open | `SourceArtifactsSection.open(_:)` | `macos/App/Features/Sources/SourceArtifactsSection.swift` | D-8, D-10: `guard let file = art.file else { return }`; empty trimmed `relPath` → refusal; else switch on `openObject` |
| Canvas Open | `EvidenceGraphModel.openSelectedFile`; `EvidenceGraphCanvasNotice.fileOpenFailed`; the callout in `EvidenceGraphView` | `macos/App/Features/EvidenceGraph/` | D-8, D-9; the existing `PVCallout(tone: .danger, title:message:)`, identifier `evidenceGraph.fileOpen.callout` unchanged |
| Metadata link | `SourcePageMetadataView.displayValue` | `macos/App/Features/Sources/SourcePageMetadataView.swift` | D-12; the link button and plain-text branches exist today |
| Metadata model | `CatalogMetadataEntry`, `GoStore.mapMetadataEntry`, `FakeStore.setSourceMetadata`, `SourceMetadataSection.applyCleared` | `macos/App/Platform/`, `macos/App/Features/Sources/` | D-12, D-13 |

States: a refused file shows the shared refusal; a missing file keeps its message. Accessibility: the Source page error banner and the canvas callout announce as they do today (S12-30 owns announcing).

### Strings

L10n *(new)*:

- `enum ObjectOpen` in `macos/App/Platform/L10n.swift` with `static func refused(filename: String) -> String`, key `objectOpen.refused.message`, = "Provenencia won't open “%@”. Its file type could run a program, or it isn't a type Provenencia stores." Comment: "Error when Open refuses a stored file; argument is the file name".
- `L10n.EvidenceGraph.fileOpenFailedTitle`, key `evidenceGraph.fileOpenFailed.title`, = "Can't open the file". Comment: "Title of the danger callout when Open cannot find or refuses the Artifact file".

Removed: `L10n.EvidenceGraph.fileNotFoundTitle` (`evidenceGraph.fileNotFound.title`). Unchanged: `sources.page.fileOpenMissing`, `evidenceGraph.fileNotFound.message`, `L10n.Sources.openMetadataURL`.

## 7. Steps

**STEP-1 — One table of stored types; the path builder refuses the rest**
Files: `core/ingest/mediatypes/mediatypes.go`, `core/ingest/mediatypes/mediatypes_test.go`, `core/objectpath/objectpath.go`, `core/objectpath/objectpath_test.go`, `core/database/files/files.go` (doc only), `core/database/files/files_test.go`, `core/database/searchindex/searchindex.go`, `core/database/searchindex/searchindex_test.go`, `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`.
Do: D-1, D-2, and the `searchindex` and `derivatives` mappings of D-3. Move `TestExtension`'s rows into `TestStoredExtension` *(new)* (T-1), dropping `image/jpg` to the not-ok side. Rewrite `TestRel` as a table (T-2). In `TestStorageRelPath`, the "unknown" and "empty" rows expect `objectpath.ErrDisallowedType` (T-3). Rewrite `TestObjectsRelPathAgreesWithObjectpath` in place as T-4. In `thumbnail_test.go`, the skip case's `installSourceFile(…, "application/octet-stream", "blob.bin")` becomes `installSourceFile(…, "audio/mpeg", "blob.mp3")` (still skipped: `raster.Supported` rejects it).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/...` passes; `go list -deps ./core/objectpath | grep provenencia` shows only `core/objectpath` and `core/ingest/mediatypes`; `rg -n 'objectpath\.Extension' core api` finds nothing.

**STEP-2 — Reads drop a disallowed path instead of failing**
Files: `core/database/files/files.go`, `core/database/files/files_test.go`, `api/ffi/handlers/artifacts.go`, `api/ffi/handlers/thumbnails.go`, `api/ffi/handlers/sources_test.go`. Skill: `add-ffi-handler`.
Do: the `files` and handler parts of D-3. Add T-5, T-6, and the damaged-row cases of T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; `rg -n 'StorageRelPath\(' api` finds nothing.

**STEP-3 — Go refuses an executable object**
Files: `core/objectstore/objectstore.go`, `core/objectstore/objectstore_unix_test.go`.
Do: D-4. Add T-10 and T-11.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/objectstore/ ./core/derivatives/ ./core/ingest/...` passes.

**STEP-4 — Metadata entries carry the web link Go computes**
Files: `core/database/sourcemetadata/sourcemetadata.go`, `core/database/sourcemetadata/sourcemetadata_test.go`, `api/proto/engine.proto`, generated Go and Swift from `./scripts/generate-proto.sh`, `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`, `core/urlshape/urlshape_test.go`. Skill: `add-ffi-handler`.
Do: D-11. Add T-7, T-8, the metadata cases of T-9, and T-12.
Check: `./scripts/generate-proto.sh` leaves no diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/urlshape/ ./core/database/sourcemetadata/` passes.

**STEP-5 — Swift refuses an executable or typeless object and names the file** (one commit: changing `openObject`'s return type breaks its two callers until they change)
Files: `macos/App/Platform/ProjectFiles.swift`, `macos/App/Platform/IngestMediaPolicy.swift`, `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `EvidenceCanvasInputMode.swift`, `EvidenceGraphView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/ProjectFilesTests.swift`, `macos/ProvenenciaTests/ArtifactFileFixtures.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`, `macos/ProvenenciaTests/IngestMediaPolicyTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-5 – D-10 and §6 Strings. Add the `mode:` parameter to `ArtifactFileFixtures.write`. Rename `openObjectReturnsFalseForSymlinkedObject` and `openObjectReturnsFalseForRejectedRelPath` to `…ReportsMissingFor…` and expect `.missing`. Update the `.fileNotFound` match in `EvidenceGraphModelTests` (near line 977). Add T-13 – T-17.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO -only-testing:ProvenenciaTests` passes; `rg -n 'fileNotFoundTitle|fileNotFound\.title|isSafeObject|\.fileNotFound\(' macos` finds nothing.

**STEP-6 — Swift opens only Go's web link**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/App/Features/Sources/SourcePageMetadataView.swift`, `macos/App/Features/Sources/SourceMetadataSection.swift`, `macos/ProvenenciaTests/MetadataModelTests.swift`.
Do: D-12, D-13. Add T-18.
Check: `rg -n openSavedURL macos` finds nothing; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO -only-testing:ProvenenciaTests/MetadataModelTests` passes.

**STEP-7 — Docs**
Files: `.cursor/rules/macos-client.mdc` (the `objectURL` line: it also refuses an execute bit; open an object only through `ProjectFiles.openObject`; open a metadata link only through `openHref`), `docs/application-stack.md` "Object names" (every object has an extension from the allowlist; no bare hex) and "Read path" (Go returns a path only for an allowed type; Swift's open checks and the execute bit), `docs/source-layer-data-model.md` (the basename paragraph: no bare hex name, a type off the allowlist has no path; the object files paragraph: the execute bit in both mirrors), `docs/artifact-file-storage.md` (a stored type off the allowlist has no path on read), `completed.md`, `deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: `rg -n 'bare hex' docs core` finds nothing; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-23-safe-files-and-links.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/ingest/mediatypes/mediatypes_test.go` `TestStoredExtension` *(new)* | `TestExtension`'s rows (`IMAGE/PNG` → `.png`, `text/plain; charset=utf-8` → `.txt`, …) ok; `image/jpg`, `application/zip`, `""` → not ok; a loop over `Allowed`'s table asserts every extension is non-empty and starts with `.` | AC-1.6, D-1 |
| T-2 | Go | `core/objectpath/objectpath_test.go` `TestRel` (rewritten as a table) | `image/jpeg` → `.jpg` path; `application/pdf` → `.pdf`; `application/x-sh`, `application/octet-stream`, `""`, `image/jpg` → `ErrDisallowedType`; short checksum → `ErrInvalidChecksum` | AC-1.2, D-2 |
| T-3 | Go | `core/database/files/files_test.go` `TestStorageRelPath` | "unknown" and "empty" rows → `errors.Is(err, objectpath.ErrDisallowedType)` | AC-1.2 |
| T-4 | Go | `core/database/searchindex/searchindex_test.go` `TestObjectsRelPathAgreesWithObjectpath` (rewritten in place) | allowed types → path equals `objectpath.Rel`'s; `image/jpg`, `application/octet-stream`, `""` → `("", nil)`; short checksum → error | AC-1.5, D-3 |
| T-5 | Go | `core/database/files/files_test.go` `TestRelPaths` and `TestReadRelPath` *(new)* | `RelPaths`: new row "disallowed type omitted" (a PNG row and an `application/x-sh` row → only the PNG). `ReadRelPath`: PNG → path; `application/x-sh` → `("", nil)`; bad checksum → `ErrInvalid` | AC-1.3, D-3 |
| T-6 | Go | `api/ffi/handlers/sources_test.go` `TestListSources` | new case: a Source whose cover thumbnail's `files` row is damaged to `application/x-sh` → no error; that row's `thumbnail_rel_path == ""` | AC-1.5, D-3 |
| T-7 | Go | `core/urlshape/urlshape_test.go` `TestOpenHref` | `file:///System/Applications/Calculator.app`, `smb://server.example.org/share`, `vnc://host.example.org` → `ErrInvalid` | AC-3.1 |
| T-8 | Go | `core/database/sourcemetadata/sourcemetadata_test.go` `TestWorkspaceEntryOpenHref` *(new)* | url field `www.example.org` → `https://www.example.org`; url field `file:///System/Applications/Calculator.app` → `""`; text field `www.example.org` → `""`; `Value == nil` → `""` | AC-3.2 – AC-3.4, D-11 |
| T-9 | Go | `api/ffi/handlers/sources_test.go` `TestGetSourceWorkspace` | new cases: ingest a PNG Artifact, `UPDATE files SET media_type = 'application/x-sh'` → no error; `artifacts[0].file.rel_path == ""`, `original_filename` and `media_type` kept. A url field `www.example.org` → `open_href` `https://www.example.org`; its value damaged by SQL to `file:///System/Applications/Calculator.app` → `open_href ""`, `value_text` kept | AC-1.1, AC-3.2, AC-3.3, D-3, D-11 |
| T-10 | Go | `core/objectstore/objectstore_unix_test.go` `TestOpenObjectExecutable` *(new)* | object written then `chmod 0o755` → `errors.Is(err, ErrExecutable)` and `errors.Is(err, ErrNotRegular)` | AC-2.1, D-4 |
| T-11 | Go | `core/objectstore/objectstore_unix_test.go` `TestWriteIfAbsentReplacesExecutable` *(new)* | matching bytes at mode `0o755`, `WriteIfAbsent` → `wrote == true`, mode `&0o111 == 0` | AC-2.2, D-4 |
| T-12 | Go | `api/ffi/handlers/sources_test.go` `TestSetSourceMetadata` | url field set to `example.org/records` → returned `entry.open_href == "https://example.org/records"` | AC-3.5 |
| T-13 | Swift | `macos/ProvenenciaTests/ProjectFilesTests.swift` `objectURLRefusesExecutableObject` *(new)* | `.png` at `0o755` → `objectURL == nil` | AC-2.3, D-5 |
| T-14 | Swift | `ProjectFilesTests.swift` `openableObjectURLRefusesUnsafeObjects` *(new)*, `@Test(arguments:)` over (relPath, mode) | (`aabb`, `0o644`), (`aabb.sh`, `0o644`), (`aabb.command`, `0o644`), (`aabb.app` as a regular file, `0o644`), (`aabb.pdf`, `0o755`) → `.failure(.unsafe)` | AC-2.4, AC-2.5, AC-2.6, D-7 |
| T-15 | Swift | `ProjectFilesTests.swift` `openableObjectURLAcceptsAStoredPDF` *(new)* | `aabb.pdf` at `0o644` → `.success`; missing `aabb.pdf` → `.failure(.missing)`; symlinked `aabb.pdf` → `.failure(.missing)` | AC-2.7, D-5 |
| T-16 | Swift | `macos/ProvenenciaTests/IngestMediaPolicyTests.swift` `isExecutableTypeCoversScriptsAndApps` *(new)* | `.shellScript`, `.unixExecutable`, `.applicationBundle` → `true`; `.pdf` → `false` | AC-2.9, D-6 |
| T-17 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `openRefusesAFileWithNoPath` *(new)* | FakeStore Artifact with `file.originalFilename "run.sh"`, `relPath ""`; `model.artifacts.open(art)` → `context.pageError == L10n.ObjectOpen.refused(filename: "run.sh")` | AC-1.4, D-8, D-10 |
| T-18 | Swift | `macos/ProvenenciaTests/MetadataModelTests.swift` `openURLUsesOnlyOpenHref` *(new)* | `openHref ""` → `nil`; `openHref "https://example.org"` → that URL | AC-3.6, D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `ProjectFilesTests`, `EvidenceGraphModelTests` (the renamed notice), the `IngestMediaPolicy` tests (the shared predicate), `ThumbnailCacheTests` and `ArtifactViewerModelTests` (a fixture written with an execute bit would now be refused; none are today).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: application-stack.md "Object names" and "Read path", source-layer-data-model.md basename and object files paragraphs, artifact-file-storage.md; .cursor/rules/macos-client.mdc
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

Sweep names for this PR: `objectpath.Extension`, `isSafeObject`, `openSavedURL`, `fileNotFoundTitle`, `evidenceGraph.fileNotFound.title`, `EvidenceGraphCanvasNotice.fileNotFound`, `bare hex`.

## 10. Manual check

1. Copy a project with a Source that has a PNG Artifact and a url metadata field set to `www.example.org`. Close the app.
2. In the copy's `provenencia.sqlite`: `UPDATE files SET media_type = 'application/x-sh' WHERE original_filename LIKE '%.png';` and set the url field's `value_text` to `file:///System/Applications/Calculator.app`.
3. Open the copy. The Sources list and the Source page load. The Artifact tile shows its type icon; Open on it shows "Provenencia won't open “<name>.png”. Its file type could run a program, or it isn't a type Provenencia stores." Nothing launches (AC-1.1, AC-1.4).
4. The metadata value reads `file:///System/Applications/Calculator.app` as plain text, not a link. Clicking it does nothing (AC-3.3).
5. In the original project, `chmod +x` an Artifact's object under the project's objects folder. Open it from the canvas Files pane: the callout titled "Can't open the file" shows the refusal with that file's name; Calculator or Terminal does not open (AC-2.8).
6. In the original project, a url field set to `www.example.org` opens `https://www.example.org` in the browser (AC-3.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A real catalog holds a legacy type not on today's allowlist, and its files stop opening | Ingest has resolved types through `mediatypes.Resolve` since the allowlist landed, and the product owner is the only user; the refusal names the file, so a legacy type shows up at once |
| A damaged row's bytes are never purged (`PurgeUnreferenced` cannot build their path) | Out, not planned: the bytes are inert while nothing opens them |
| `UTType(filenameExtension:)` returns a dynamic type for an unknown extension | Go returns only allowlisted extensions, so an unknown extension reaches Swift only from a hand-edited path; a dynamic type conforms to none of the refused types and opens like any document, which is the rule's intent |
| Ingest now also refuses `.script` types | Go's allowlist already refuses every script type, so no file that ingests today is newly refused |
| A test fixture writes an object with an execute bit | STEP-5's check runs every Swift suite; fix the fixture's mode, not the rule |

**Stop and report if:**

- `go list -deps ./core/ingest/mediatypes` lists any other `provenencia` package (D-1 would then risk an import cycle).
- A non-test caller of `objectpath.Rel` or `files.StorageRelPath` exists that §4 does not list.
- An existing test other than those in §4's "Go tests that this PR breaks" row fails because a type has no path.
- A Swift test or fixture relies on opening an extensionless object through `openObject` for a reason other than refusal.
- `xcodebuild` is not available on the machine you run on: finish STEP-1 – STEP-4, write STEP-5 and STEP-6 without running their checks, and report the Swift checks as not run instead of marking them done.

## 12. Open questions

None.

## 13. PR

**Title:** S12-23 — The app opens only safe files and links
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-23-safe-files-and-links.md`.
