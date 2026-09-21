# Marks (macOS)

Curated research artwork for Provenencia — not SF Symbols. One recipe:
`PVMark` / `PVMarkKey` in this folder. Assets:
`Resources/Assets.xcassets/Marks/provenencia_<key>.imageset` (template-rendered SVG).

## Families

| Prefix | Meaning | Persistence |
| --- | --- | --- |
| `file_*` | MIME / extension stand-in when a File has no raster derivative | Resolved client-side via `PVFileTypeGlyph` |
| `type_*` | What a **Source type** is as a record (register, reel, headstone, …) | `source_types.icon_key` |
| `subject_*` | What a **node** is in the evidence graph / Subject fields | Registry `IconSymbol` (product-fixed) |

**Do not confuse** `type_*` with `subject_source`. `type_*` names a held Source type;
`subject_source` is the reified source **kind** on the graph (folio mark from Subject
fields / S7-D2). Different families — never substituted.

### `file_*` (9)

`file_pdf`, `file_doc`, `file_txt`, `file_sheet`, `file_slides`, `file_video`,
`file_audio`, `file_image_missing`, `file_generic`.

Mono band labels are composited in SwiftUI (Xcode’s SVG importer does not resolve
`<text>`). Map MIME / filename → key via `PVFileTypeGlyph.key(mediaType:originalFilename:)`.

### `type_*` (22)

`type_certificate`, `type_book`, `type_document`, `type_scroll`, `type_photograph`,
`type_newspaper`, `type_map`, `type_microfilm`, `type_cassette`, `type_oral_history`,
`type_video`, `type_website`, `type_census`, `type_dna`, `type_gedcom`, `type_grave`,
`type_scrapbook`, `type_evidence`, `type_folder_archive`, `type_email`, `type_postcard`,
`type_passport`.

Fallback / new custom type default: `type_evidence` (`PVMarkKey.fallback` /
`PVMarkKey.defaultTypeMark`).

### `subject_*` (7)

| Key | Role | Provenance |
| --- | --- | --- |
| `subject_person` | node | S6 Evidence graph board |
| `subject_event` | node | S6 Evidence graph board |
| `subject_place` | node | S6 Evidence graph board |
| `subject_relationship` | bridge | S6 Evidence graph board |
| `subject_participation` | bridge | S6 Evidence graph board |
| `subject_location` | bridge | S6 Evidence graph board |
| `subject_source` | reification | S7-D2 Subject fields board (`SourceMark`) |

## Tint / color

Assets are **monochrome templates**. Ink comes from SwiftUI environment /
`.foregroundStyle` at the call site. Kind pigment (person wash vs place wash)
stays on the parent chrome (registry presentation tokens → `PVColor`) — never
baked into mark assets. Do not enable multicolor “original” rendering.

## Size

| Preset | pt | Typical use |
| --- | --- | --- |
| `PVMarkSize.inline` | 16 | Compact chrome |
| raw ~12–17 | — | Graph cards (~12–15), palette (~17), Subject fields strip (~14–16) |
| `PVMarkSize.row` | 24 | List rows |
| `PVMarkSize.tile` | 40 | Pickers / form tiles |

File mono labels show by default at ≥22pt (`showLabel:` overrides).

## Accessibility

Match decorative vs named: when an adjacent label already names the kind (graph
card title, Subject fields strip), pass `decorative: true`. When the mark is the
sole affordance (picker cell), leave decorative false so `accessibilityName` is
spoken.

## Board links (art SoT)

- Evidence graph: https://claude.ai/design/p/2239e965-3b09-4c13-b85a-d54316ffd8fb?via=share
- Subject fields (source folio): https://claude.ai/design/p/6dceb4b9-d08a-40ad-a46f-430651ea9b3c?via=share
