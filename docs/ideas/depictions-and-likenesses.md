# Depictions and likenesses — later

**Status:** idea, not scheduled. Spike 10 shipped the `depiction` value type, the `likeness` property, the composer row, the Person gallery, and a preferred portrait ([decisions](../deployment-plan/archive/spike-10/)). Promote comparing two likenesses is Spike 11 (S11-20). Event and Place pictures (`event_picture`, `place_picture`) a chosen Event or Place thumbnail, and rectangle-only depiction drawing are [Spike 14](../deployment-plan/spike-14/) batch 1.

## Still open

- **Signature.** Another key of type `depiction`, or something of its own. Spike 14 gave each kind its own picture key (`likeness`, `event_picture`, `place_picture`); a signature would follow that or not.
- **Every saved face on the photo.** The composer's viewer shows one region at a time; showing every saved depiction on the photo needs a read-only region layer.
- **A hidden value's reason and support.** A hidden reconciler row takes its first entry's reason and counts every entry's Source, so a value carried by an accepted weak photo reads `provisional` when a provisional member's copy of the same picture ranks above it. The gallery reads each record's own outcome instead (S10-60); other readers of `auto_reconciler_values.reason` still see the row's.
- **One media type for the stored image.** `ingest.StageImage` writes `image/png` as a literal in three places and `depictionvalue.ReadableImage` checks the same literal; one constant in `core/depictionvalue` would tie them, and the transparent-polygon crop would change one place.
- **Windows.** That client draws the PDF region with its own stack before it sends the image. Spike 10 draws it with PDFKit on the Mac.


## From the 2026-10-18 likeness audit

The Spike 10 follow-ups (S10-60 – S10-66) fixed the High items and Medium items 5, 7 and 8. Spike 14 takes the gallery image source (S14-01) and the composer feedback, copy and cap items (S14-03); Spike 12 takes decoding the crop before the session (S12-27). These remain:

- **A refused image (UX-11).** No distinct glyph or copy for an image the thumbnail check refuses.
- **Thumbnail work (PRF-2, Medium items 6 and 13).** Cancel a decode that is already running, decode at the drawn size, bound the cache by bytes, and keep the evidence card's loaded likeness.
- **Go crop work (PRF-4).** Store crops below 2,048 px. (Decoding the crop before the catalog session, PRF-3, and one stage-then-write entry point went to Spike 12 S12-27, 2026-10-27.)
- **Raster byte limit (SEC-3).** `core/raster` limits decodes by pixel size, not by bytes.
- **The PDF cut reopens its file.** The off-main cut opens the PDF by URL without running `ProjectFiles.objectURL` again; carry the project folder and object path in the snapshot instead.
- **`CitationLocatorDraft.encodeJSON()` key order.** It serializes without `.sortedKeys`, so two encodings of one locator can differ as strings.

## Still out

Audio and video playback ([`audio-video-sources.md`](audio-video-sources.md)). A video frame plus a region may later be a depiction. The player is separate.

Go-side PDF thumbnail generation ([`artifact-pdf-thumbnails.md`](artifact-pdf-thumbnails.md)). Spike 10 does not put a PDF engine in Go.
