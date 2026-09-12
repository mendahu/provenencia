# Evidence thumbnails: file-type fallbacks, Source-type icons, and primary Artifact

**Status:** idea — PR1 (file-type glyphs) done on `feat/file-type-fallback-glyphs`; PR2/PR3 not roadmapped. Related client media work: [`image-optimization.md`](archive/image-optimization.md). Domain: [`source-layer-data-model.md`](../source-layer-data-model.md) §§3, 6–8; vocabulary: [`seeded-vocabulary.md`](../seeded-vocabulary.md) §2.1.

## Problem

### Non-image Files look empty

`derivatives.EnsureThumbnail` only rasterizes common **image** MIME types. PDFs, Office docs, video, and other evidence skip cleanly — Artifact rows and Sources list cells fall back to a generic empty `PVThumbnail`. Researchers still need a **recognizable** stand-in: “this is a PDF,” “this is a video,” not a blank square.

### Fileless / physical Artifacts look empty too

An Artifact with no File (physical book on the shelf, cassette in a box, courthouse register only seen on microfilm) has nothing to rasterize. A MIME glyph does not apply. The **Source type** is the best signal for what kind of evidence it is — birth certificate, oral history, parish register — if types can carry a **representational icon**.

### Sources have many Artifacts; which thumb represents the Source?

A Source can own several Artifacts, each with zero or one primary File, each File optionally with a thumbnail derivative. List rows and Source identity chrome need **one** representation for the Source. Today the list is enriched with *some* `thumbnail_rel_path` (first available raster path in practice) — there is no durable, user-controllable choice when multiple thumbs (or only fallbacks) exist.

## Idea

Three cooperating layers of representation, in preference order when resolving what to show:

```text
1. Raster file_derivatives thumbnail (when present)
2. File-type glyph (PDF / MP4 / …) when there is a File but no raster thumb
3. Source-type icon (book, microfilm, interview, …) for fileless Artifacts — or when chosen as Source cover
4. Generic empty placeholder
```

### 1. File-type fallback thumbnails

When there is a File but no raster thumbnail derivative, render a **typed fallback** in `PVThumbnail` (or a thin wrapper):

| Kind of stand-in | Examples |
| --- | --- |
| Document badge | PDF, DOC/DOCX, RTF, plain text |
| Spreadsheet / slides | XLS, CSV, PPT (if we care) |
| Media | MP4 / MOV / audio waveform-style glyph |
| Generic file | unknown / `application/octet-stream` |

Visual language: archival (not colorful emoji), e.g. document silhouette with a short type label (`PDF`, `MP4`) in mono. Map from `files.media_type` and/or filename extension; one shared map (design-system helper).

**Not** inventing fake JPEG derivatives for PDFs in Go unless we later add real `pdf_page_preview`. MIME fallback is **UI chrome**, not a `file_derivatives` row.

### 2. Source-type icon vocabulary

Design a **curated collection** of icons that represent *kinds of evidence*, independent of digital file format. Examples to explore in Design:

- Generic book / bound volume
- Loose document / certificate
- Parchment / scroll
- Cassette / magnetic tape
- Microfilm / microfiche
- Video / film reel
- Audio / interview / oral history
- Photograph / album (even when no File yet)
- Map / plat
- Newspaper / clipping
- Digital-native / website (if useful)
- Generic “evidence” fallback for custom types

**On Source types:** when creating or editing a **type** (seeded or user), the researcher **picks one icon** from the collection (stored as a stable key on `source_types`, e.g. `icon_key`, not raw image bytes). Seeded `provenencia` types ship with sensible defaults; user types default to a generic document/book until changed.

**Not on Sources:** creating or editing a **Source** does **not** pick a free-floating icon. The researcher picks a **Source type**; the type’s `icon_key` comes along. The create/edit Source type picker should **show** each type’s icon for scanability. Changing a Source’s type later inherits that type’s icon for type-icon fallbacks / cover.

**Where it shows**

| Surface | Behavior |
| --- | --- |
| **Create / edit Source type** | Icon picker — this is where `icon_key` is chosen. |
| **Create / edit Source** | Type picker shows icons next to type labels; no separate icon field on the Source. |
| **Fileless Artifact** | Use the parent Source’s type icon as the Artifact row thumbnail (physical-only stand-in). |
| **Source cover (optional)** | User may set the Source-level thumbnail to the **type icon** (not only to an Artifact raster/MIME glyph) — useful when all Artifacts are fileless or when the type mark is a better list identity than a random scan. |
| **Vocabulary / type lists** | Show the icon next to the type label for scanability. |

Icons live in the **design system** (asset pack + `PVSymbol`-like or dedicated `SourceTypeIcon` keyed enum). Catalog stores only the **key**; clients resolve artwork. Plugins later could register extra keys under a namespaced prefix if needed — keep v1 closed set.

### 3. Source primary thumbnail (rollup + user choice)

Persist what supplies the Source’s representative thumbnail — broader than “which Artifact’s File”:

```text
Source cover resolution
  ├── primary Artifact with raster thumb
  ├── primary Artifact with MIME glyph only
  ├── explicit “use Source type icon”
  └── auto: first raster ensured → else first useful Artifact fallback → else type icon → empty
```

**Default behavior**

- When the **first** raster thumbnail is successfully ensured for any Artifact under the Source, mark that Artifact as the Source’s **primary thumbnail source** (auto).
- If nothing rasterizes (all PDFs / fileless), fall through to MIME glyph or **Source-type icon** so the Sources list is never a wall of blanks.

**User control**

- **Use as Source thumbnail** on an Artifact row (raster or MIME fallback for that Artifact).
- **Use Source type icon as thumbnail** (identity chrome or type picker) so cover can be the type mark even when rasters exist — researcher choice wins over auto.
- Switching updates list + header immediately; one primary mode at a time.

**Schema sketch (when pulled into a spike)**

- `sources.primary_artifact_id` (nullable FK) **and/or** a small cover mode enum (`artifact` | `source_type_icon`) — exact shape TBD.
- `source_types.icon_key` (TEXT, required or defaulted) referencing the curated set.
- List/Get Source still returns enough for Swift to render: either `thumbnail_rel_path` **or** a resolved `cover: { kind: raster|mime|type_icon, … }` so cells stay simple.

## Why these hang together

| Situation | What the researcher sees |
| --- | --- |
| Scanned JPEG Artifact | Real thumb; can be Source cover |
| PDF Artifact | PDF glyph; can be Source cover |
| Fileless “parish register” Artifact | Source-type scroll/book icon |
| Multi-Artifact Source | Explicit primary / type-icon cover instead of opaque “first ensure” |

Without type icons, physical evidence stays blank. Without MIME glyphs, digital non-images stay blank. Without primary/cover control, multi-Artifact Sources stay arbitrary.

## Suggested delivery (when scheduled)

**Multiple PRs**, not one mega-PR. The three layers ship value independently, touch different surfaces (Swift-only vs schema+FFI vs cover UX), and leave the tree reviewable. A single PR would mix design-system assets, catalog migrations, and Source identity chrome.

Design assets: PR1 shipped `file_*`; PR2 still needs the `type_*` keys locked and imported.

| PR | Delivers | Leaves the tree… |
| --- | --- | --- |
| **1 — File-type glyphs** | **Done:** nine `file_*` assets + `PVEvidenceIcon` / `PVFileTypeGlyph`; `PVThumbnail` evidence-glyph state; `CachedThumbnail` MIME fallback; ListSources `thumbnail_media_type` + `thumbnail_original_filename` with first-raster-then-file cover scan; Artifact rows, Sources list, identity header wired | Digital non-images (PDF, video, …) show a typed stand-in; empty only for fileless / unknown-missing |
| **2 — Source-type icons** | Closed `icon_key` set on `source_types`; seed defaults; icon picker on create/edit **type**; type icons visible in create/edit **Source** type picker; fileless Artifact thumbs + type-list marks | Physical / fileless rows and vocabulary are scannable; custom types default to generic |
| **3 — Source primary cover** | Persist cover mode + optional primary Artifact; auto first-raster; user “use as Source thumbnail” / “use type icon”; List/Get returns resolved cover for cells | Multi-Artifact Sources have durable, controllable list identity |

**Do not** fold cover control into PR1/PR2 — resolution rules and open questions below belong in PR3 once both fallbacks exist.

**Do not** invent `file_derivatives` rows for MIME glyphs — UI chrome only until a real `pdf_page_preview` (non-goal).

### PR1 — File-type fallback glyphs (**done**)

Shipped on `feat/file-type-fallback-glyphs`:

1. Closed `file_*` key set + SVG imagesets under `Assets.xcassets/EvidenceIcons/` (from design-system export).
2. `PVFileTypeGlyph` MIME / extension → key map (Swift tests).
3. `PVThumbnail` evidence-glyph content + `PVEvidenceIcon` (ink-band label composited in SwiftUI).
4. Wired: Sources list, Artifact rows / primary-file card, Source identity header (client fallthrough from artifacts when workspace `source` lacks cover fields).
5. ListSources enrichment: prefer first successful raster; else empty path + first file-bearing Artifact’s `media_type` / `original_filename`. No catalog schema migration. Files list still descoped.

### PR2 — Source-type icon vocabulary (actionable)

Schema + vocabulary UI + fileless thumbs. Depends on `type_*` assets (and preferably PR1 so lists already understand non-raster thumbs).

1. Lock the closed `type_*` key set (~12–20) and seeded defaults (`birth_certificate` → `type_certificate`, etc.).
2. Catalog migration: `source_types.icon_key` (TEXT, defaulted); backfill existing rows; registry seeds set defaults for `provenencia` types.
3. Proto / FFI: create/update/list Source types carry `icon_key`; validate against the closed set (reject unknown keys).
4. Design-system `SourceTypeIcon` (or equivalent) keyed enum resolving `icon_key` → artwork — catalog stores only the key.
5. Source **types** create/edit: icon picker (grid of the closed set); user types default to `type_evidence` / `type_document` until changed. This is the only place `icon_key` is chosen.
6. Source create/edit: type picker (and type filter chips if present) **display** each type’s icon next to the label — still selecting a type, not a separate icon.
7. Fileless Artifact thumbnail: resolve parent Source’s type `icon_key` into the thumbnail slot.
8. Source types list / type chips: show the icon next to the type label.
9. Tests: migration backfill; create/update reject bad keys; FakeStore + Swift tests for type icon default, create-Source type picker showing icons, and fileless row.

### PR3 — Source primary thumbnail / cover (actionable)

Depends on PR1 + PR2 so cover resolution can return `raster | mime | type_icon`. Settle open questions in this PR’s design brief before coding.

1. Decide schema: `sources.primary_artifact_id` (nullable FK) **and** cover mode (`artifact` | `source_type_icon`) — or equivalent; document pin-vs-auto rules (see Open questions).
2. Auto: on first successful raster ensure under a Source, set primary Artifact if cover is still auto/unpinned.
3. Fallthrough when nothing rasterizes: primary Artifact MIME glyph → else Source type icon → else empty.
4. User actions: “Use as Source thumbnail” on an Artifact row; “Use Source type icon as thumbnail” in identity chrome (and/or type area); one mode at a time; list + header update immediately.
5. List/Get Source (and workspace enrichment) return a resolved cover payload Swift can render without re-deriving preference order — e.g. `cover: { kind, thumbnail_rel_path?, media_type?, icon_key? }`.
6. Tests: auto-assign on first raster; pin type icon survives later raster; switch Artifact primary; delete primary Artifact falls through cleanly.

## Non-goals (for the parked idea)

- Full in-app PDF/page raster pipeline (`pdf_page_preview`) — heavier; glyphs first
- Replacing content-addressed derivatives with UI-only icons for images that *do* have thumbs
- Async decode cache — stays in [`image-optimization.md`](image-optimization.md)
- Per-Artifact custom uploaded icons (type-level curated set is enough for v1)
- Utility / chrome icons beyond `file_*` and `type_*` (add, pin, repository marks — separate if needed)

## Open questions

**Settle before / in PR3**

- Primary = Artifact id vs File id vs cover-mode enum?
- If user pinned type icon as cover, does a later first raster auto-steal cover or leave the pin?
- If primary Artifact is PDF-only, show PDF glyph or fall through to type icon unless pinned?
- Where do cover actions live — Artifact row, identity header, both?
- Share packages: embed raster only, or also ship `icon_key` / cover mode?

**Settle before / in PR2 (Design)**

- Closed icon set size for v1 (~12–20)? Naming keys (`type_book`, `type_microfilm`, `type_oral_history`, …)?
- Seeded types: which default icons for `birth_certificate` and friends?

## Related docs

- [`source-layer-data-model.md`](../source-layer-data-model.md) §§3, 6–8
- [`seeded-vocabulary.md`](../seeded-vocabulary.md) §2.1 (`source_types`)
- [`image-optimization.md`](archive/image-optimization.md)
- [`artifact-file-storage.md`](../artifact-file-storage.md)
- Spike 2 completed notes for S2-12 / S2-26 (thumbnail ensure + list enrichment)
