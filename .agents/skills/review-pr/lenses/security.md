# Security lens

The question: **can a file, a project folder, or a value the app did not
create make it read, write, or show something it should not?** Provenencia is
local-first: there is no server and no login. The attack surface is what comes
in from disk and from other people's projects, and what crosses the FFI.

Runs when the diff touches file I/O, ingest, derivatives, the object store,
identity, project open, FFI, SQL, URLs, locators, entitlements, or stored
numbers, weights, and lists the code trusts (the matching layers). Describe the
risk and the fix. Do not write exploits.

## Read first

- [`docs/artifact-file-storage.md`](../../../../docs/artifact-file-storage.md),
  [`docs/user-identity-model.md`](../../../../docs/user-identity-model.md),
  [`docs/application-stack.md`](../../../../docs/application-stack.md).
- The trust boundaries the diff touches:

| Boundary | Where | Untrusted input |
| --- | --- | --- |
| Ingest | `core/ingest`, `core/ingest/mediatypes`, `macos/App/Platform/IngestMediaPolicy.swift`; `ingest.StageImage` decodes and re-encodes app-rendered images | Any file the researcher drops or picks: name, size, type, bytes |
| Object store | `core/objectpath`, `core/objectstore` (every Go object read, write, install, and removal goes through an `os.Root` on objects/); Swift reads: `macos/App/Platform/ProjectFiles.swift` `objectURL` (an `lstat` walk from objects/) | Content hashes and relative paths read back from the catalog |
| Derivatives | `core/derivatives` (thumbnails), `core/raster` (the one bounded codec, also used by staging) | Image and PDF bytes decoded in-process |
| Project open | `core/database` migrate and schema check, `core/catalogsession` | A `*.provenencia` folder from someone else: its SQLite, its files |
| Identity | `core/identity`, `identity.json`, `active-project.json` | Files under Documents another app or user can edit |
| FFI | `api/ffi`, `api/ffi/handlers`, `macos/App/Platform/GoStore.swift`, `CoreInvoke.swift` | Bytes from Swift; errors going back |
| Catalog values | `core/urlshape`, `core/locator`, `valuecodec` | URLs, locator JSON, text that later renders |
| Stored numbers and lists | `core/matchregistry`, `core/valuecompare`, `core/propertycompare`, `core/subjectcompare`, `core/subjectprofile`, `core/autoreconcile`, `core/database/namealiases` | Weights, counts, and lists read back from a catalog another tool may have written: `NaN`, `+Inf`, `1e308`, a negative weight, a 10,000-word name |

## How to work

1. For each changed line that reads input from a boundary above, follow the
   value to where it is used: a path, a query, an allocation, a decoder, an
   `open`, a rendered string.
2. Ask what happens with the hostile version: `../`, an absolute path, a
   symlink, a 4 GB file, a zero-byte file, a file whose extension lies, a
   malformed PDF, a SQLite row another tool wrote, a locator with a huge
   polygon, a URL with a `javascript:` or `file:` scheme.
3. Check that the existing guard is used, not re-implemented.
4. Rate what you find as if the callers named in the map already exist. A hole
   in code this PR adds is this PR's to close, even when the FFI field, the
   renderer, or the handler that reaches it lands later (see
   [`review-pr`](../SKILL.md) "Rate the code as merged").

## Look for

- **Paths.** Joining a catalog or user value into a filesystem path without
  `objectpath`'s validation; following symlinks out of the project folder;
  writing outside the project or the app's container.
- **SQL.** A query built by string concatenation with a value instead of a
  bound parameter. `IN` lists go through `sqlbatch`.
- **Unbounded reads.** Reading a whole file or blob into memory without a size
  limit; decoding an image without dimension limits; a loop driven by a count
  from the file.
- **Type enforcement.** Trusting an extension or the client's MIME claim instead
  of the ingest media policy; a new accepted type without a decoder that
  handles malformed input.
- **Project open.** Trusting rows in a catalog the app did not write: a
  `user_version` newer than the app, a schema that fails `verifySchema`,
  a path or a ref used without validation. Opening must refuse, not repair
  silently.
- **Reads that pass untrusted values through.** A catalog value that the write
  path validates (a locator, a region, a URL, a path) is untrusted again when
  it is read back, because another tool or another person's project may have
  written it. The Go read that hands it to callers validates it there, once,
  so no caller can forget. Leaving the check to a later renderer or cropper is
  the finding, not a mitigation.
- **Cross-project bleed.** State from one project visible in another after a
  switch: the held catalog session, caches, thumbnails, the active identity.
- **FFI.** A handler trusting a length or an id from Swift without a check;
  an internal error, path, or SQL text returned in a message the UI shows.
- **Secrets and personal data.** Tokens, absolute home paths, or a researcher's
  names written to logs or crash text; anything added to the repo that should
  not be there.
- **Entitlements and TCC.** A new entitlement or a broader file-access scope
  than the feature needs; a debug default left on for release.
- **Rendering.** A catalog URL opened without `urlshape`; text from a file
  rendered as markup.

## Allow

The researcher's own catalog writing values they typed. Files inside the
project folder the app created. Test fixtures.

## Severity in this lens

| Severity | Use when |
| --- | --- |
| **blocker** | A read or write outside the project or container; SQL injection; a crash or hang from a crafted file on open or ingest; cross-project data bleed; a secret committed; an untrusted catalog value that a read returns unvalidated, or a guard missing from code this PR adds, that a mapped later PR will reach |
| **major** | A missing size or dimension bound; a guard re-implemented instead of reused; an internal error leaked to the UI; an entitlement wider than needed |
| **minor** | Defense in depth behind a guard that already covers the path |
| **nit** | — (security findings are not nits) |

"No reachable path today" is not a severity. If a later PR in the map will
reach it, rate it as reached.

## Good finding shape

The shape to copy. Names in the example are illustrative.

"`derivatives.CropDepiction` sizes its output buffer from the locator polygon's
bounding box without clamping to the source image, so a locator saved by an
older build with a polygon outside the frame allocates without bound. Clamp to
the image bounds before allocating, as `raster.EncodeJPEG` and `raster.EncodePNG` do."
