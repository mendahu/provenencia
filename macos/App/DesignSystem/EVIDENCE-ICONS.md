# Evidence icons (macOS)

Curated marks for **evidence representation** in thumbnail slots. Not SF Symbols —
see `Components/Research/PVEvidenceIcon.swift`.

## PR1 (`file_*` only)

Nine MIME / extension stand-ins when a File exists but has no raster derivative:

`file_pdf`, `file_doc`, `file_txt`, `file_sheet`, `file_slides`, `file_video`,
`file_audio`, `file_image_missing`, `file_generic`.

Assets: `Resources/Assets.xcassets/EvidenceIcons/provenencia_file_*.imageset`
(template-rendered SVG). Labels are composited in SwiftUI, not baked into the SVG.

Map MIME / filename → key via `PVFileTypeGlyph.key(mediaType:originalFilename:imageThumbFailed:)`.

## PR2 (`type_*`)

Source-type marks (`type_certificate`, `type_book`, …) and `source_types.icon_key`
are deferred. Do not invent ad-hoc SF Symbol stand-ins for physical evidence.
