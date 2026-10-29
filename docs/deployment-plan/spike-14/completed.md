# Spike 14 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S14-01 — Pictures engine

**Landed 2026-10-28** in [#615](https://github.com/mendahu/provenencia/pull/615). New catalogs seed locked `event_picture` / `place_picture` (Label Picture). Package `core/database/gallery` admits `GalleryKeys()`, reads each tile from `depiction_values.image_file_id` (SEC-4), and `leadPictures` fills `PictureImageFileID` on Person, Event and Place headers. Proto/Swift wire rename to `PictureDetail` / `pictures` / `pictureImageRelPath`; preferred column and `SetPreferredLikeness` stay. View-layer gallery chrome remains S14-02.

## S14-08 — Other-link support calibration

**Landed 2026-10-28** in [#613](https://github.com/mendahu/provenencia/pull/613). `segment.OtherSupport` is 0.75 (`OtherCredit` stays 0.25), so a Source person with only a short given name plus a lined-up grandparent (or other non-birth/death link) clears Weak on Promote for the Jake + Carol dogfood shape; inventory and `promote-graph-alignment.md` quote the new total, `TestPartialNameWithGrandparentClearsWeak` locks the band, and product version is 0.2.1.

## S14-24 — Workspace width class

**Landed 2026-10-29** in [#628](https://github.com/mendahu/provenencia/pull/628). `PVSpacing.widthWideBreakpoint` (1600) and `PVWidthClass` with the `pvWidthClass` environment value; `WorkspaceDestinationHost` fills the detail area and measures the pane once, outside the place's `.id`, so any page reads `.regular` or `.wide` without measuring itself. No page reads it yet and nothing on screen changes; a shared `hostInBorderlessWindow` test helper replaces three hand-written window mounts.
