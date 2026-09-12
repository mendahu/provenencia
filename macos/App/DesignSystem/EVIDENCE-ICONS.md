# Evidence icons (macOS)

Curated marks for **evidence representation** in thumbnail slots. Not SF Symbols —
see `Components/Research/PVEvidenceIcon.swift`.

## PR1 (`file_*`) — done

Nine MIME / extension stand-ins when a File exists but has no raster derivative:

`file_pdf`, `file_doc`, `file_txt`, `file_sheet`, `file_slides`, `file_video`,
`file_audio`, `file_image_missing`, `file_generic`.

Assets: `Resources/Assets.xcassets/EvidenceIcons/provenencia_file_*.imageset`
(template-rendered SVG). Labels are composited in SwiftUI, not baked into the SVG.

Map MIME / filename → key via `PVFileTypeGlyph.key(mediaType:originalFilename:imageThumbFailed:)`.

## PR2 (`type_*`) — done

Source-type marks on `source_types.icon_key` (required, default `type_evidence`):

`type_certificate`, `type_book`, `type_document`, `type_scroll`, `type_photograph`,
`type_newspaper`, `type_map`, `type_microfilm`, `type_cassette`, `type_oral_history`,
`type_video`, `type_website`, `type_census`, `type_dna`, `type_gedcom`, `type_grave`,
`type_scrapbook`, `type_evidence`, `type_folder_archive`, `type_email`, `type_postcard`.

Assets: `provenencia_type_*.imageset` in the same catalog. No ink-band label.

Cover preference for Sources / identity / `publishCoverToList`: **raster → type icon →
MIME (error path) → empty**. Artifact rows **with a File** stay raster → MIME; **fileless**
Artifact rows use the parent Source’s type icon.

Source types create/edit: form shows a 44px tile (28px mark + name + key + Change);
Change opens a **Choose an icon** dialog with the 21 marks at 40px (names under each),
immediate selection, metaphor footer, Done to dismiss.
