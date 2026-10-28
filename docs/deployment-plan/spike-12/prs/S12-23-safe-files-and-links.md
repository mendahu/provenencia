# S12-23 — The app opens only safe files and links

**Status:** Draft
**Spike / slice:** Spike 12 · Slice 5 — Audit fixes: High
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the refusal reuses the Source page's error banner and the canvas's file callout, with new copy only)
**Board:** —
**Depends on:** —
**Blocks:** S12-28 (map: "After S12-23, whose `files` filter the viewer relies on")
**Branch:** `s12-23-safe-files-and-links`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-38 did; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [app-health audit](../app-health-audit.md) A-1, A-5 (and L-6, left out); working rules "What the app hands to macOS to open is checked first, and Go decides the type", "FakeStore is a seeded fake, not a second engine", "High tolerance for churn"; [map review](../deployment-plan.review.md) K-8; [`artifact-file-storage.md`](../../../artifact-file-storage.md); [`application-stack.md`](../../../application-stack.md) "Read path"; [`source-layer-data-model.md`](../../../source-layer-data-model.md) (object files paragraph)
**Planned:** 2026-10-28 against `main` @ `7adb3bc`
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

A project folder someone shares can no longer make Provenencia launch a program. Today a crafted catalog can give a stored file any `media_type`; Go then returns an object path with no extension, and Open hands it to `NSWorkspace`, which runs an extensionless executable in Terminal (audit A-1). A crafted metadata URL value opens with any scheme, `file:` and `smb:` included (audit A-5). After this PR, Go returns an object path only for a type on the ingest allowlist, both layers refuse an object with an execute bit, Swift refuses to open an object with no extension or of an executable, application or script type and names the file, and a metadata link opens only through the web link Go computes.

**Why now.** Bar outcome 11 ("A shared project is safe to open"), slice 5. The audit rated both items High.

**What stays.** Ingest, the allowlist itself, how Artifacts and metadata look, Open's behavior on every allowed file, and the missing-file messages. The viewer's decode checks are S12-28.

## 2. User stories

**US-1.** As a researcher who opens a project someone sent me, when I open an Artifact whose stored file type is not one Provenencia stores, the app refuses and names the file, so that nothing runs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A `files` row damaged to `media_type = 'application/x-sh'`; `GetSourceWorkspace` on its Source | The read succeeds; the Artifact's `file` carries its `original_filename` and `media_type` and an empty `rel_path` |
| AC-1.2 | `objectpath.Rel(sum, t)` for `t` in `application/x-sh`, `application/octet-stream`, `""`, `image/jpg` | Each returns `objectpath.ErrDisallowedType`; `image/jpeg` returns `objects/…/<sum>.jpg` |
| AC-1.3 | `files.RelPaths` over a PNG row and a row damaged to `application/x-sh` | Only the PNG's id is in the map |
| AC-1.4 | On the Source page, Open on that Artifact's tile (empty `relPath`, file `run.sh`) | The page error reads `Provenencia won't open “run.sh”. Its file type could run a program, or it isn't a type Provenencia stores.` |
| AC-1.5 | `ListSources` on a project whose cover thumbnail's `files` row is damaged to `application/x-sh` | The read succeeds; that Source's `thumbnail_rel_path` is empty |

**US-2.** As a researcher, when a stored object has an execute bit or a dangerous type on disk, the app refuses to open it and names the file.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `objectstore.OpenObject` on a regular object with mode `0o755` | Returns an error for which `errors.Is(err, objectstore.ErrExecutable)` and `errors.Is(err, objectstore.ErrNotRegular)` both hold |
| AC-2.2 | `objectstore.WriteIfAbsent` when the object already exists with the right bytes and mode `0o755` | `wrote` is true and the object's mode has no execute bit |
| AC-2.3 | `ProjectFiles.objectURL` on a regular `.png` object with mode `0o755` | `nil` |
| AC-2.4 | `ProjectFiles.openableObjectURL` on an existing object `objects/aa/bb/aabb` (no extension) | `.failure(.unsafe)` |
| AC-2.5 | `ProjectFiles.openableObjectURL` on an existing `objects/aa/bb/aabb.sh`, and on `objects/aa/bb/aabb.command` | `.failure(.unsafe)` for each |
| AC-2.6 | `ProjectFiles.openableObjectURL` on an existing `objects/aa/bb/aabb.pdf` with mode `0o755` | `.failure(.unsafe)` |
| AC-2.7 | `ProjectFiles.openableObjectURL` on an existing `objects/aa/bb/aabb.pdf` with mode `0o644`; on a missing one | `.success(url)`; `.failure(.missing)` |
| AC-2.8 | On the canvas Files pane, Open on a row whose object is refused as `.unsafe` | The file callout's message is the AC-1.4 text with that row's file name |

**US-3.** As a researcher, when I click a saved metadata link, it opens only on the web.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `urlshape.OpenHref` on `file:///System/Applications/Calculator.app`, `smb://server.example.org/share`, `vnc://host.example.org` | `ErrInvalid` for each |
| AC-3.2 | `GetSourceWorkspace` with a url-type field whose value is `www.example.org` | That entry's `open_href` is `https://www.example.org` |
| AC-3.3 | A url-type value damaged to `file:///System/Applications/Calculator.app` | `GetSourceWorkspace` returns the entry with that `value_text` and an empty `open_href` |
| AC-3.4 | A text-type field with value `www.example.org` | `open_href` is empty |
| AC-3.5 | `SetSourceMetadata` on a url field with `example.org/records` | The returned entry's `open_href` is `https://example.org/records` |
| AC-3.6 | `CatalogMetadataEntry.openURL` for a url field with `openHref` `""`; with `https://example.org` | `nil`; `URL(string: "https://example.org")` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `objectpath.Rel` refuses a type not on the ingest allowlist (A-1) | STEP-1, AC-1.2 |
| Every caller of the path maps that refusal to "no path" | STEP-2, AC-1.1, AC-1.3, AC-1.5 |
| `objectstore.openRegular` refuses an execute bit (A-1) | STEP-3, AC-2.1, AC-2.2 |
| `MetadataWorkspaceEntry.open_href` from `urlshape.OpenHref` (A-5) | STEP-4, AC-3.1 – AC-3.5 |
| Swift `isSafeObject` refuses an execute bit; `openableObjectURL` refuses no extension and executable, application and script types; `openObject` reports why (A-1) | STEP-5, AC-2.3 – AC-2.7 |
| The Source page and the canvas name a refused file (A-1) | STEP-6, AC-1.4, AC-2.8 |
| Swift opens only `open_href`; `openSavedURL` goes (A-5) | STEP-7, AC-3.6 |
| Docs and rules that describe the read path | STEP-8 |

| Out | Home |
| --- | --- |
| Links inside PDFs (`ArtifactPDFViewport` has no `PDFViewDelegate`; audit L-6) | [audit, Not scheduled](../app-health-audit.md#not-scheduled) |
| An App Sandbox | not planned (map scope entry) |
| The viewer's sniff-vs-media-type and pixel-budget checks | S12-28 |
| Listing a refused file in the canvas Files pane (it lists only Artifacts with a path, as today) | not planned |
| Removing the bytes of a damaged row on purge (`PurgeUnreferenced` skips a path it cannot build) | not planned |
| Folding `objectpath.Extension`'s type list into the allowlist package | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Object path | `core/objectpath/objectpath.go` `Rel`, `Extension` | Builds `objects/hh/hh/<hex><ext>`; an unknown type gives no extension. The one path builder: `files` and `searchindex` both call it |
| Allowlist | `core/ingest/mediatypes/mediatypes.go` `Allowed`, `Normalize` | The ingest allowlist; imports only `path/filepath` and `strings`. Checked on write only |
| Files path | `core/database/files/files.go` `StorageRelPath`, `RelPaths` | `StorageRelPath` passes `objectpath.Rel` through (maps only `ErrInvalidChecksum`). `RelPaths` omits an id whose path fails |
| Purge | `core/database/files/purge.go` `PurgeUnreferenced` | `continue` when `StorageRelPath` fails |
| Artifact read | `api/ffi/handlers/artifacts.go` `artifactProto` | Returns the error when `StorageRelPath` fails, which fails the whole `GetSourceWorkspace` |
| Thumbnails | `api/ffi/handlers/thumbnails.go` `thumbnailRelPath`, `existingThumbnailRelPath` | Return the `StorageRelPath` error; `existingThumbnailRelPath` feeds `sourceCoverThumbnail` |
| Sources list cover | `core/database/searchindex/searchindex.go` `objectsRelPath` (called by `existingSourceCoverThumb`) | Wraps any `objectpath.Rel` error as "bad checksum", which fails the search document write |
| Derivatives | `core/derivatives/ensure.go` `Ensure` | Skips types `raster.Supported` rejects, then calls `StorageRelPath`; `raster.supportedMIME` holds `image/jpg`, which is not on the allowlist, so a damaged `image/jpg` row would return the new error |
| Other path callers | `core/ingest/file.go` (`mediatypes.Resolve` output), `core/ingest/stage.go` (`image/png`), `core/derivatives/ensure.go` (`image/jpeg` derived) | Always allowed types; unaffected |
| Object open (Go) | `core/objectstore/objectstore.go` `openRegular`, `ErrNotRegular`, `matches` | Refuses a non-regular file or a symlinked shard; no mode-bit check. `matches` treats `ErrNotRegular` as "no match", so `installTemp` replaces the entry |
| Object open (Swift) | `macos/App/Platform/ProjectFiles.swift` `objectURL`, `isSafeObject`, `openObject` | `isSafeObject` checks folders, regular file and Finder alias, not the execute bit. `openObject` returns `Bool` and passes any URL `objectURL` accepts to `NSWorkspace.shared.open` |
| Source page Open | `macos/App/Features/Sources/SourceArtifactsSection.swift` `open(_:)` | Returns silently on an empty `relPath`; on `false` sets `context.pageError` to `L10n.Sources.fileOpenMissing` |
| Canvas Open | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `openSelectedFile`; `EvidenceCanvasInputMode.swift` `EvidenceGraphCanvasNotice.fileNotFound`; `EvidenceGraphView.swift` (callout titled `L10n.EvidenceGraph.fileNotFoundTitle`) | On `false` sets `fileOpenError` to `L10n.EvidenceGraph.fileNotFoundMessage(name:filename:)`. Rows come from `EvidenceGraphFiles.fileBearing`, which drops an Artifact with no `openableRelPath` |
| Metadata wire | `api/proto/engine.proto` `MetadataWorkspaceEntry` (fields 1, 2, 4, 5, 6; reserved 3, 7, 8); `api/ffi/handlers/sources.go` `metadataEntryProto` (used by `GetSourceWorkspace`, `SetSourceMetadata` and two more reads in that file) | No open link; `urlshape.OpenHref` has no caller |
| URL shape | `core/urlshape/urlshape.go` `OpenHref` | Already returns `ErrInvalid` for any scheme but `http` / `https` and adds `https://` to a scheme-less value |
| Metadata link (Swift) | `macos/App/Features/Sources/SourcePageMetadataView.swift` `displayValue`, `openSavedURL` | Prefixes `https://` when the text has no `://`, then opens any scheme through `openURL` |
| Metadata model (Swift) | `macos/App/Platform/GenealogyStore.swift` `CatalogMetadataEntry`; `GoStore.swift` `mapMetadataEntry`; `FakeStore.swift` `setSourceMetadata`; `SourceMetadataSection.swift` `applyCleared` | No href field; `applyCleared` blanks `valueText` on a suggested row |
| Tests | `core/objectpath/objectpath_test.go` `TestRel`; `core/database/files/files_test.go` `TestStorageRelPath` ("unknown" and "empty" rows expect a bare path), `TestRelPaths`; `core/objectstore/objectstore_unix_test.go`; `core/urlshape/urlshape_test.go` `TestOpenHref`; `api/ffi/handlers/sources_test.go` `TestGetSourceWorkspace`, `TestSetSourceMetadata`, `TestListSources`; `macos/ProvenenciaTests/ProjectFilesTests.swift` (`openObjectReturnsFalseFor…`); `EvidenceGraphModelTests.swift` (`openSelectedFileReportsAMissingFileAndKeepsTheSelection`, `armingAToolClearsTheMissingFileError`); `SourcePageModelTests.swift`; `MetadataModelTests.swift` | `TestStorageRelPath` "unknown"/"empty" rows and `TestRel`'s `image/jpg` alias case change meaning in this PR |

**Precedent.** S10-62 ([#464](https://github.com/mendahu/provenencia/pull/464)), commits `200c4687` and `6cd8c2b1`: it added the symlink and Finder-alias refusals to `openRegular` and `isSafeObject` together, with `ArtifactFileFixtures` helpers and `ProjectFilesTests` cases, and updated `.cursor/rules/macos-client.mdc`, `application-stack.md` "Read path" and `source-layer-data-model.md` in the same PR. Copy that shape: one rule in both mirrors, one fixture helper, the same three docs. The damaged-row Go tests copy `api/ffi/handlers/citations_test.go`'s `UPDATE files SET media_type = …` damage rows.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The allowlist check lives in `objectpath.Rel`: it returns `ErrDisallowedType` *(new)* when `!mediatypes.Allowed(mediaType)` or `Extension(mediaType) == ""`. `objectpath` imports `core/ingest/mediatypes` | `Rel` is the one place every object path is built, `files` and `searchindex` both; one check covers every path the FFI returns, including Source covers, which bypass `files`. `mediatypes` is a leaf package, so no cycle. The `Extension` half guarantees a path always has an extension even if the two lists drift | Filtering in `files.scanFile` / `Lookup` — because `searchindex.existingSourceCoverThumb` reads `files` with its own SQL, and dropping rows would also break `LookupByChecksum` dedupe for a damaged row. Filtering per handler — four copies |
| D-2 | Each caller maps `ErrDisallowedType` to "no path": `files.RelPaths` and `PurgeUnreferenced` already skip; `artifactProto` returns the `SourceFileRef` with `rel_path` empty; `thumbnailRelPath` returns `("", true, nil)`; `existingThumbnailRelPath` returns `("", nil)`; `searchindex.objectsRelPath` returns `("", nil)`; `derivatives.Ensure` returns `ErrUnprocessable` | A damaged row must not fail a whole page or list read; it hides one file. Keeping the file ref (name, type, size) lets Swift name the file it refuses (bar outcome 11) | Omitting `file` from the Artifact — Swift could not name the file. A new wire field for "refused" — empty `rel_path` already means "no openable object" in Swift (`openableRelPath`) |
| D-3 | Go refuses an execute bit in `openRegular` with `ErrExecutable` *(new)* = `fmt.Errorf("%w: execute bit set", ErrNotRegular)`, checked on both the `Lstat` and the opened file's `Stat` (`Mode().Perm()&0o111 != 0`) | Wrapping `ErrNotRegular` keeps every caller that already handles it (`matches`, `derivatives.readSourceObject`) correct with no edit, and `matches` returning false makes `installTemp` replace such an object with a fresh `0o600` file on re-import | A new unrelated sentinel — every caller would need a second `errors.Is` |
| D-4 | Swift: `ProjectFiles.ObjectRefusal` *(new)* `enum { case missing, unsafe }`. `isSafeObject` returns `ObjectRefusal?`: an execute bit (`st.st_mode & 0o111 != 0`) is `.unsafe`; every refusal it has today stays `.missing`. `objectURL` keeps its signature and returns `nil` on either | A symlink or alias stays "missing", as S10-62 decided; only the new refusal needs its own message. Thumbnails, the viewer and the PDF crop gain the execute-bit refusal through `objectURL` with no edit | Making every existing refusal `.unsafe` — changes S10-62's copy decision, not in scope |
| D-5 | `ProjectFiles.openableObjectURL(projectDir:relPath:) -> Result<URL, ObjectRefusal>` *(new)* runs the `objectURL` checks, then refuses `.unsafe` when `url.pathExtension` is empty, or `UTType(filenameExtension:)` conforms to `.executable`, `.application` or `.script`. `openObject(projectDir:relPath:)` returns `ObjectRefusal?` (nil when opened): it calls `openableObjectURL`, then `NSWorkspace.shared.open`; a `false` from `open` is `.missing` | Working rule: Swift checks extension, execute bit and executable types only; Go owns the allowlist, so there is no type list in Swift. A pure check function is testable without launching an app | A Swift allowlist — the working rule forbids it (map review K-8). Keeping `Bool` — callers could not tell missing from refused |
| D-6 | Copy: one shared refusal message `L10n.Errors.objectOpenRefused(filename:)` *(new)*. The Source page shows it in `context.pageError` for an Artifact with a `file` and an empty `relPath`, and for `.unsafe`; `.missing` keeps `fileOpenMissing`. The canvas shows it in `fileOpenError` for `.unsafe`; `.missing` keeps `fileNotFoundMessage` | One rule, one string, in the shared `Errors` namespace both features read | Two copies in `Sources` and `EvidenceGraph` |
| D-7 | The canvas callout title becomes one title for both cases: `L10n.EvidenceGraph.fileNotFoundTitle` is renamed `fileOpenFailedTitle` *(new key `evidenceGraph.fileOpenFailed.title`)* = "Can't open the file", and `EvidenceGraphCanvasNotice.fileNotFound` is renamed `.fileOpenFailed`. `fileOpenError` stays `String?` | "File not found" is false for a refusal; one neutral title keeps the notice a single string with no new type | A title per refusal kind — needs a struct in `fileOpenError` and in the notice for no researcher gain |
| D-8 | The Source page names the file by `file.originalFilename`, or the Artifact's `label` when that is empty; the canvas uses `EvidenceGraphFiles.Row.displayFilename` | The names each surface already shows | — |
| D-9 | `MetadataWorkspaceEntry` gains `string open_href = 9;`. `metadataEntryProto` sets it from `urlshape.OpenHref(e.Value.ValueText)` when the field's `DataType` is `metadatafields.DataTypeURL` and the value is present; an error leaves it empty. Every response that uses `metadataEntryProto` carries it | Go decides what is a web link (working rule); the projection is the one place every metadata read and `SetSourceMetadata`'s returned entry pass through | Computing it in `sourcemetadata.WorkspaceEntry` — the domain read has no wire concern; Swift computing it — the audit's defect |
| D-10 | Swift: `CatalogMetadataEntry.openHref: String = ""` *(new)* and a computed `openURL: URL?` *(new)*: `nil` unless the field is a url field and `openHref` is non-empty, else `URL(string: openHref)`. `displayValue` shows the link button only when `openURL` is non-nil and opens that URL; otherwise it shows the plain-text branch. `openSavedURL` is deleted. `applyCleared` blanks `openHref` with `valueText` | Swift opens only what Go returned; the default value keeps the 25 existing initializers unchanged | A scheme check in Swift — Go already decided |
| D-11 | `FakeStore.setSourceMetadata` sets `openHref = ""` on the entry it writes; it computes no href. A test that needs a link seeds `openHref` | Working rule "FakeStore is a seeded fake, not a second engine" | Calling a Swift copy of `OpenHref` in the fake |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `objectpath.ErrDisallowedType` *(new)* | error | `objectpath.ErrInvalidChecksum` | Extends the same package beside it; a different refusal |
| `objectstore.ErrExecutable` *(new)* | error | `objectstore.ErrNotRegular` | Wraps it (D-3), so it extends rather than adds a mechanism |
| `ProjectFiles.ObjectRefusal` *(new)* | type | `ProjectFiles.objectURL`'s `nil` | `nil` cannot say why; searched `macos/App` for an open-result type (`openObject`, `OpenResult`, `Refusal`): none |
| `ProjectFiles.openableObjectURL` *(new)* | func | `ProjectFiles.objectURL` | Calls it, then adds the open-only checks; `objectURL` stays the read path for thumbnails and the viewer, which must not refuse by extension |
| `MetadataWorkspaceEntry.open_href` *(new)* | wire field | `SourceFileRef.rel_path` (Go computes, Swift uses) | Same pattern on another message |
| `CatalogMetadataEntry.openHref`, `openURL` *(new)* | Swift fields | `CatalogArtifact.openableRelPath` | Same pattern: the model says what is openable |
| `L10n.Errors.objectOpenRefused(filename:)` *(new)* | L10n | `L10n.Sources.fileOpenMissing`, `L10n.EvidenceGraph.fileNotFoundMessage` | Those say "missing"; the refusal is a different fact, shared by both surfaces |
| `ArtifactFileFixtures.writeObject(projectDir:relPath:mode:)` *(new)* | test helper | `ArtifactFileFixtures.writePNG`, `writeBlankPDF` | They write image or PDF bytes at a fixed mode; the open tests need arbitrary bytes at a chosen mode |

**Builds on.** `mediatypes.Allowed` (the one allowlist), `objectpath.Rel` (the one path builder), S10-62's `openRegular` / `isSafeObject` mirror, `urlshape.OpenHref`, `metadataEntryProto`.
**Deletes.** `SourcePageMetadataView.openSavedURL`; the `Bool` return of `ProjectFiles.openObject`; `L10n.EvidenceGraph.fileNotFoundTitle` and its `evidenceGraph.fileNotFound.title` catalog entry (renamed, D-7); `EvidenceGraphCanvasNotice.fileNotFound` (renamed); `TestRel`'s `image/jpg` alias assertion and `TestStorageRelPath`'s bare-path rows (their meaning is reversed, not kept).
**Next adds.** S12-28's viewer and PDF crop reach objects through `objectURL` and inherit the execute-bit refusal; its sniff check relies on Go returning only allowed types. Audit L-6 (PDF link annotations) would route a link through the same "Go-computed web link" rule.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An Artifact with a `file` and an empty `rel_path` | Source page tile (`SourcePageArtifactsView`), card, `SourceArtifactsSection.open` | Tile already paints the type icon when `thumbnailRelPath` is empty; `open` now shows the refusal (AC-1.4) instead of returning silently |
| Same Artifact | Canvas Files pane (`EvidenceGraphFiles.fileBearing`) | Dropped, as an Artifact with no path is today (Out) |
| Same Artifact as a Source cover | Sources list row, Source header | `thumbnail_rel_path` empty → the type icon, as for a missing derivative (AC-1.5) |
| A metadata url entry with an empty `open_href` | `SourcePageMetadataView.displayValue` | Plain text, as a text field shows (D-10) |

### Scale

None — no write, list or loop changes; `open_href` is one string function per metadata row on reads that already build the row, and the allowlist check is one map lookup per path.

### Schema

None — no catalog change. The filter is on read; no migration rewrites stored types.

### Go core

| Package | Change |
| --- | --- |
| `core/objectpath` | `var ErrDisallowedType = errors.New("objectpath: media type not allowed")` *(new)*. `Rel` checks the checksum first, then returns `ErrDisallowedType` when `!mediatypes.Allowed(mediaType) \|\| Extension(mediaType) == ""`. Doc comments on `Rel` and the package say so |
| `core/database/files` | `StorageRelPath` doc: a disallowed type returns `objectpath.ErrDisallowedType` (passed through). No code change; `RelPaths` and `PurgeUnreferenced` already skip |
| `core/database/searchindex` | `objectsRelPath`: `errors.Is(err, objectpath.ErrDisallowedType)` → `("", nil)`; other errors as today |
| `core/derivatives` | `Ensure`: the source `StorageRelPath` error `ErrDisallowedType` → `ErrUnprocessable` |
| `core/objectstore` | `var ErrExecutable = fmt.Errorf("%w: execute bit set", ErrNotRegular)` *(new)*. `openRegular` returns it after each `IsRegular` check when `Mode().Perm()&0o111 != 0`. `ErrNotRegular`'s and `openRegular`'s doc comments name it. `matches` unchanged (D-3) |
| `core/urlshape` | No code change; `TestOpenHref` gains rows (T-6) |

Errors: none reach the FFI. `ErrDisallowedType` is mapped at each caller (D-2); `ErrExecutable` surfaces as today's `ErrNotRegular` handling.

### FFI

`api/proto/engine.proto`: `MetadataWorkspaceEntry` gains `string open_href = 9; // web link to open, from urlshape.OpenHref; empty when the value is not one`. Regenerate with `./scripts/generate-proto.sh`.

| Handler | Change |
| --- | --- |
| `api/ffi/handlers/sources.go` `metadataEntryProto` | Sets `OpenHref` per D-9 |
| `api/ffi/handlers/artifacts.go` `artifactProto` | `ErrDisallowedType` → `rel = ""`, still `out.File = fileRefProto(f, rel)`; still asks `thumbnailRelPath` |
| `api/ffi/handlers/thumbnails.go` `thumbnailRelPath`, `existingThumbnailRelPath` | `ErrDisallowedType` from the derived file's path → empty path (D-2) |

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Object checks | `ProjectFiles.isSafeObject`, `objectURL`, `openableObjectURL` *(new)*, `openObject`, `ObjectRefusal` *(new)* | `macos/App/Platform/ProjectFiles.swift` | D-4, D-5. `import UniformTypeIdentifiers` is already there |
| Source page Open | `SourceArtifactsSection.open(_:)` | `macos/App/Features/Sources/SourceArtifactsSection.swift` | D-6, D-8: `guard let file = art.file else { return }`; empty `relPath` → refusal; else switch on `openObject` |
| Canvas Open | `EvidenceGraphModel.openSelectedFile`; `EvidenceGraphCanvasNotice.fileOpenFailed`; the callout in `EvidenceGraphView` | `macos/App/Features/EvidenceGraph/` | D-6, D-7; the existing `PVCallout(tone: .danger, title:message:)`, identifier `evidenceGraph.fileOpen.callout` unchanged |
| Metadata link | `SourcePageMetadataView.displayValue` | `macos/App/Features/Sources/SourcePageMetadataView.swift` | D-10; the link button and plain-text branches exist today |
| Metadata model | `CatalogMetadataEntry`, `GoStore.mapMetadataEntry`, `FakeStore.setSourceMetadata`, `SourceMetadataSection.applyCleared` | `macos/App/Platform/`, `macos/App/Features/Sources/` | D-10, D-11 |

States: a refused file shows the shared refusal; a missing file keeps its message. Accessibility: the Source page error banner and the canvas callout announce as they do today (S12-30 owns announcing).

### Strings

L10n keys *(new)*:

- `L10n.Errors.objectOpenRefused(filename:)`, key `errors.objectOpenRefused`, = "Provenencia won't open “%@”. Its file type could run a program, or it isn't a type Provenencia stores." Comment: "Error when Open refuses a stored file; argument is the file name".
- `L10n.EvidenceGraph.fileOpenFailedTitle`, key `evidenceGraph.fileOpenFailed.title`, = "Can't open the file". Comment: "Title of the danger callout when Open cannot find or refuses the Artifact file".

Removed: `L10n.EvidenceGraph.fileNotFoundTitle` (`evidenceGraph.fileNotFound.title`). Unchanged: `sources.page.fileOpenMissing`, `evidenceGraph.fileNotFound.message`, `L10n.Sources.openMetadataURL`.

## 7. Steps

**STEP-1 — The path builder refuses a disallowed type**
Files: `core/objectpath/objectpath.go`, `core/objectpath/objectpath_test.go`, `core/database/files/files_test.go`.
Do: D-1. Rewrite `TestRel` as a table (T-1). In `TestStorageRelPath`, the "unknown" and "empty" rows expect `objectpath.ErrDisallowedType` (T-2).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/objectpath/ ./core/database/files/` passes; `go list -deps ./core/objectpath | grep provenencia` shows only `core/objectpath` and `core/ingest/mediatypes`.

**STEP-2 — Reads drop a disallowed path instead of failing**
Files: `core/database/searchindex/searchindex.go`, `core/database/searchindex/objectpath_test.go` *(new)*, `core/derivatives/ensure.go`, `api/ffi/handlers/artifacts.go`, `api/ffi/handlers/thumbnails.go`, tests in `core/database/files/files_test.go`, `api/ffi/handlers/sources_test.go`. Skill: `add-ffi-handler`.
Do: D-2 at each listed caller. Add T-3, T-4, T-5, T-16.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; `rg -n 'StorageRelPath\(|objectpath\.Rel\(' core api --glob '!*_test.go'` lists only the callers in §4.

**STEP-3 — Go refuses an executable object**
Files: `core/objectstore/objectstore.go`, `core/objectstore/objectstore_unix_test.go`.
Do: D-3. Add T-7 and T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/objectstore/ ./core/derivatives/ ./core/ingest/...` passes.

**STEP-4 — Metadata entries carry the web link Go computes**
Files: `api/proto/engine.proto`, generated Go and Swift from `./scripts/generate-proto.sh`, `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`, `core/urlshape/urlshape_test.go`. Skill: `add-ffi-handler`.
Do: D-9. Add T-6, T-9, T-10.
Check: `./scripts/generate-proto.sh` leaves no diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/urlshape/` passes.

**STEP-5 — Swift refuses an executable or typeless object and says why**
Files: `macos/App/Platform/ProjectFiles.swift`, `macos/ProvenenciaTests/ProjectFilesTests.swift`, `macos/ProvenenciaTests/ArtifactFileFixtures.swift`. Skill: `add-swift-test`.
Do: D-4, D-5. Add `ArtifactFileFixtures.writeObject`. Change `openObjectReturnsFalseForSymlinkedObject` and `openObjectReturnsFalseForRejectedRelPath` to expect `.missing` and rename `false` to `Missing` in their names. Add T-11 – T-13. Update `openRegular`'s and `isSafeObject`'s "mirrors" comments to name the execute bit.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO -only-testing:ProvenenciaTests/ProjectFilesTests` passes. (The two callers will not compile until STEP-6; do STEP-5 and STEP-6 in one commit if the build blocks the test run.)

**STEP-6 — The Source page and the canvas name a refused file**
Files: `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `EvidenceCanvasInputMode.swift`, `EvidenceGraphView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/SourcePageModelTests.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-6, D-7, D-8 and §6 Strings. Update the `.fileNotFound` match in `EvidenceGraphModelTests` (near line 977). Add T-14.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO -only-testing:ProvenenciaTests/SourcePageModelTests -only-testing:ProvenenciaTests/EvidenceGraphModelTests` passes; `rg -n 'fileNotFoundTitle|fileNotFound\.title|\.fileNotFound\b' macos` finds nothing.

**STEP-7 — Swift opens only Go's web link**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/App/Features/Sources/SourcePageMetadataView.swift`, `macos/App/Features/Sources/SourceMetadataSection.swift`, `macos/ProvenenciaTests/MetadataModelTests.swift`.
Do: D-10, D-11. Add T-15.
Check: `rg -n openSavedURL macos` finds nothing; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO -only-testing:ProvenenciaTests/MetadataModelTests` passes.

**STEP-8 — Docs**
Files: `.cursor/rules/macos-client.mdc` (the `objectURL` line: it also refuses an execute bit; open an object only through `ProjectFiles.openObject`; open a metadata link only through `openHref`), `docs/application-stack.md` "Read path" (Go returns a path only for an allowed type; Swift's open checks), `docs/source-layer-data-model.md` (object files paragraph: the execute bit, in both mirrors), `docs/artifact-file-storage.md` (a stored type off the allowlist has no path on read), `completed.md`, `deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-23-safe-files-and-links.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/objectpath/objectpath_test.go` `TestRel` (rewritten as a table) | `image/jpeg` → `.jpg` path; `application/pdf` → `.pdf`; `application/x-sh`, `application/octet-stream`, `""`, `image/jpg` → `ErrDisallowedType`; short checksum → `ErrInvalidChecksum` | AC-1.2, D-1 |
| T-2 | Go | `core/database/files/files_test.go` `TestStorageRelPath` | "unknown" and "empty" rows → `errors.Is(err, objectpath.ErrDisallowedType)` | AC-1.2 |
| T-3 | Go | `core/database/files/files_test.go` `TestRelPaths` | new row `"disallowed type omitted"`: a PNG row and a row inserted with `application/x-sh` → map holds only the PNG | AC-1.3 |
| T-4 | Go | `api/ffi/handlers/sources_test.go` `TestGetSourceWorkspace` | new case: ingest a PNG Artifact, `UPDATE files SET media_type = 'application/x-sh'` → no error; `artifacts[0].file.rel_path == ""`, `original_filename` and `media_type` kept | AC-1.1, D-2 |
| T-5 | Go | `api/ffi/handlers/sources_test.go` `TestListSources` | new case: a Source whose cover thumbnail's `files` row is damaged to `application/x-sh` → no error; that row's `thumbnail_rel_path == ""` | AC-1.5, D-2 |
| T-6 | Go | `core/urlshape/urlshape_test.go` `TestOpenHref` | `file:///System/Applications/Calculator.app`, `smb://server.example.org/share`, `vnc://host.example.org` → `ErrInvalid` | AC-3.1 |
| T-7 | Go | `core/objectstore/objectstore_unix_test.go` `TestOpenObjectExecutable` *(new)* | object written then `chmod 0o755` → `errors.Is(err, ErrExecutable)` and `errors.Is(err, ErrNotRegular)` | AC-2.1, D-3 |
| T-8 | Go | `core/objectstore/objectstore_unix_test.go` `TestWriteIfAbsentReplacesExecutable` *(new)* | matching bytes at mode `0o755`, `WriteIfAbsent` → `wrote == true`, mode `&0o111 == 0` | AC-2.2, D-3 |
| T-9 | Go | `api/ffi/handlers/sources_test.go` `TestGetSourceWorkspace` | new cases: url field `www.example.org` → `open_href` `https://www.example.org`; value damaged by SQL to `file:///System/Applications/Calculator.app` → `open_href ""`, `value_text` kept; text field `www.example.org` → `open_href ""` | AC-3.2, AC-3.3, AC-3.4, D-9 |
| T-10 | Go | `api/ffi/handlers/sources_test.go` `TestSetSourceMetadata` | url field set to `example.org/records` → returned `entry.open_href == "https://example.org/records"` | AC-3.5 |
| T-11 | Swift | `macos/ProvenenciaTests/ProjectFilesTests.swift` `objectURLRefusesExecutableObject` *(new)* | `.png` at `0o755` → `objectURL == nil` | AC-2.3, D-4 |
| T-12 | Swift | `ProjectFilesTests.swift` `openableObjectURLRefusesUnsafeObjects` *(new, parameterized)* | `aabb` (no extension), `aabb.sh`, `aabb.command`, `aabb.pdf` at `0o755` → `.failure(.unsafe)` | AC-2.4, AC-2.5, AC-2.6, D-5 |
| T-13 | Swift | `ProjectFilesTests.swift` `openableObjectURLAcceptsAStoredPDF` *(new)* | `aabb.pdf` at `0o644` → `.success`; missing `aabb.pdf` → `.failure(.missing)` | AC-2.7 |
| T-14 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `openRefusesAFileWithNoPath` *(new)* | FakeStore Artifact with `file.originalFilename "run.sh"`, `relPath ""`; `model.artifacts.open(art)` → `context.pageError == L10n.Errors.objectOpenRefused(filename: "run.sh")` | AC-1.4, D-6, D-8 |
| T-15 | Swift | `macos/ProvenenciaTests/MetadataModelTests.swift` `openURLUsesOnlyOpenHref` *(new)* | url field, `openHref ""` → `nil`; `openHref "https://example.org"` → that URL; text field with `openHref "https://example.org"` → `nil` | AC-3.6, D-10 |
| T-16 | Go | `core/database/searchindex/objectpath_test.go` `TestObjectsRelPath` *(new)* in package `searchindex` | `image/jpeg` → `.jpg` path; `application/x-sh` → `("", nil)`; short checksum → error | AC-1.5, D-2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `ProjectFilesTests`, `EvidenceGraphModelTests` (the renamed notice), `ThumbnailCacheTests` and `ArtifactViewerModelTests` (fixtures written at a mode with an execute bit would now be refused; none are today).

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
- [ ] Domain doc updated if a rule in docs/*.md changed: application-stack.md "Read path", source-layer-data-model.md object files paragraph, artifact-file-storage.md; .cursor/rules/macos-client.mdc
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

Sweep names for this PR: `openSavedURL`, `fileNotFoundTitle`, `evidenceGraph.fileNotFound.title`, `EvidenceGraphCanvasNotice.fileNotFound`.

## 10. Manual check

1. Copy a project with a Source that has a PNG Artifact and a url metadata field set to `www.example.org`. Close the app.
2. In the copy's `provenencia.sqlite`: `UPDATE files SET media_type = 'application/x-sh' WHERE original_filename LIKE '%.png';` and set the url field's `value_text` to `file:///System/Applications/Calculator.app`.
3. Open the copy. The Sources list and the Source page load. The Artifact tile shows its type icon; Open on it shows "Provenencia won't open “<name>.png”. Its file type could run a program, or it isn't a type Provenencia stores." Nothing launches (AC-1.1, AC-1.4).
4. The metadata value reads `file:///System/Applications/Calculator.app` as plain text, not a link. Clicking it does nothing.
5. In the original project, `chmod +x` an Artifact's object under `objects/`. Open it from the canvas Files pane: the callout titled "Can't open the file" shows the refusal with that file's name; Calculator or Terminal does not open (AC-2.8).
6. In the original project, a url field set to `www.example.org` opens `https://www.example.org` in the browser.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A real catalog holds a legacy type not on today's allowlist, and its files stop opening | Ingest has resolved types through `mediatypes.Resolve` since the allowlist landed, and the product owner is the only user; the refusal names the file, so a legacy type shows up at once |
| A damaged row's bytes are never purged (`PurgeUnreferenced` cannot build their path) | Out, not planned: the bytes are inert while nothing opens them |
| `UTType(filenameExtension:)` returns a dynamic type for an unknown extension | Go returns only allowlisted extensions, so an unknown extension reaches Swift only from a hand-edited path; a dynamic type conforms to none of the refused types and opens like any document, which is the rule's intent |
| A test fixture writes an object with an execute bit | STEP-5's check runs the file-reading suites; fix the fixture's mode, not the rule |

**Stop and report if:**

- `go list -deps ./core/ingest/mediatypes` lists any other `provenencia` package (D-1 would then risk an import cycle).
- A path caller exists that §4 does not list (`rg 'StorageRelPath\(|objectpath\.Rel\('` finds another non-test hit).
- `objectpath.Extension` returns `""` for a type `mediatypes.Allowed` accepts (D-1's second half would hide an allowed type).
- A Swift test or fixture relies on opening an extensionless object through `openObject` for a reason other than refusal.
- `metadataEntryProto` is not the only builder of `MetadataWorkspaceEntry` in `api/ffi/handlers`.

## 12. Open questions

None.

## 13. PR

**Title:** S12-23 — The app opens only safe files and links
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-23-safe-files-and-links.md`.
