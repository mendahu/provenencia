# Spike 14 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S14-13 — A crafted project can't crash the app

**Landed 2026-10-29** in [#632](https://github.com/mendahu/provenencia/pull/632). `subjectpositions.MaxCell` (±1,000,000) is refused on write and clamped on every read, and Swift placement and moves use saturating cells (`EvidenceGraphPlacement.cellLimit`, held to Go's by `TestSwiftVocabularyMatchesGo`). The Artifact viewer decodes once through `ArtifactImageDecode`: the sniffed type must match Go's media type (else Go's extension), a declaration over 256 MP is refused, and the decode is held to `raster.MaxDecodePixels` (64 MP) for display, crop and OCR, with `displayImage` keeping the original point size; `ThumbnailDecodePolicy` now owns the shared header read and decode. `ffi.Call` recovers a panic into `internal.unknown` (the switch is `route`), and a panicking `catalogsession.Do` callback drops its session so the next call reopens. The PDF cut reopens its file through `ProjectFiles.objectURL`. Version 0.2.2. Audit A-26 to A-28.

## S14-01 — Pictures engine

**Landed 2026-10-28** in [#615](https://github.com/mendahu/provenencia/pull/615). New catalogs seed locked `event_picture` / `place_picture` (Label Picture). Package `core/database/gallery` admits `GalleryKeys()`, reads each tile from `depiction_values.image_file_id` (SEC-4), and `leadPictures` fills `PictureImageFileID` on Person, Event and Place headers. Proto/Swift wire rename to `PictureDetail` / `pictures` / `pictureImageRelPath`; preferred column and `SetPreferredLikeness` stay. View-layer gallery chrome remains S14-02.

## S14-08 — Other-link support calibration

**Landed 2026-10-28** in [#613](https://github.com/mendahu/provenencia/pull/613). `segment.OtherSupport` is 0.75 (`OtherCredit` stays 0.25), so a Source person with only a short given name plus a lined-up grandparent (or other non-birth/death link) clears Weak on Promote for the Jake + Carol dogfood shape; inventory and `promote-graph-alignment.md` quote the new total, `TestPartialNameWithGrandparentClearsWeak` locks the band, and product version is 0.2.1.

## S14-24 — Workspace width class

**Landed 2026-10-29** in [#628](https://github.com/mendahu/provenencia/pull/628). `PVSpacing.widthWideBreakpoint` (1600) and `PVWidthClass` with the `pvWidthClass` environment value; `WorkspaceDestinationHost` fills the detail area and measures the pane once, outside the place's `.id`, so any page reads `.regular` or `.wide` without measuring itself. No page reads it yet and nothing on screen changes; a shared `hostInBorderlessWindow` test helper replaces three hand-written window mounts.

## S14-D1 — Wide layout in the design system (board)

**Agreed 2026-10-29** in the Provenencia [design-system project](https://claude.ai/design/p/0f6c1f29-8408-4735-b114-5ae4c4c18445); brief archived in [`design/archive/`](design/archive/S14-D1-wide-layout.md). Per the brief, the kit gains `--width-wide-breakpoint` (1600) and `--width-secondary-max` (480), a Width classes README section with the six Wide rules, and the Page columns component: a centered primary column at Regular, primary plus a 340–480 secondary column at Wide, the group centered past both maximums, an optional header slot that scrolls or pins. No PR of its own: Spike 13's S13-10 builds `PVPageColumns` from it with the Person page frame (S14-25 descoped).
