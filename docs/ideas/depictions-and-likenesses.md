# Depictions and likenesses — later

**Status:** idea, not scheduled. Spike 10 shipped the `depiction` value type, the `likeness` property, the composer row, the Person gallery, and a preferred portrait ([decisions](../deployment-plan/archive/spike-10/)). Promote comparing two likenesses is Spike 11 (S11-20). Event and Place pictures (`event_picture`, `place_picture`) a chosen Event or Place thumbnail, and rectangle-only depiction drawing shipped in [Spike 14](../deployment-plan/archive/spike-14/README.md#decisions).

## Still open

- **Signature.** Another key of type `depiction`, or something of its own. Spike 14 gave each kind its own picture key (`likeness`, `event_picture`, `place_picture`); a signature would follow that or not.
- **Every saved face on the photo.** The composer's viewer shows one region at a time; showing every saved depiction on the photo needs a read-only region layer.
- **Confirm before a Citation page change when depictions are saved on it.** A likeness is drawn on the Citation's PDF page. Moving the Citation to another page leaves the saved likenesses pointing at a region that no longer matches; Go hides it and the edit dialog asks for a redraw, but nothing warns before the page changes (deferred from S10-34; dogfood 2026-10-17).
- **A hidden value's reason and support.** A hidden reconciler row takes its first entry's reason and counts every entry's Source, so a value carried by an accepted weak photo reads `provisional` when a provisional member's copy of the same picture ranks above it. The gallery reads each record's own outcome instead (S10-60); other readers of `auto_reconciler_values.reason` still see the row's.
- **Windows.** That client draws the PDF region with its own stack before it sends the image. Spike 10 draws it with PDFKit on the Mac.
- **Depiction captions.** While the viewer loads, the composer's depiction caption falls through to the empty case ("use a region button") and needs a loading string; `depictionDrawingHint` says "tools above" though a depiction has one tool. Both wait for the composer board's next pass.
- **EXIF orientation.** No JPEG honours EXIF orientation; a fix covers every JPEG, not only the oversized ones the reduced decoder reads.
- **A larger cover derivative.** Whether a 256 px thumbnail is enough for a Source cover cut from a very large scan.
- **Pictures as matching evidence.** `subjectcompare` gives an Event's or Place's pictures no weight.


## From the 2026-10-18 likeness audit

The Spike 10 follow-ups (S10-60 – S10-66) fixed the High items and Medium items 5, 7 and 8. Spike 14 took the gallery image source (S14-01) and the composer feedback, copy and cap items (S14-03); Spike 12 takes decoding the crop before the session and one constant for the stored-image type (S12-27), and the PDF cut's file check (S12-28). S14-23 landed a raster byte limit. These remain:

- **A refused image (UX-11).** No distinct glyph or copy for an image the thumbnail check refuses.
- **Thumbnail work (PRF-2, Medium items 6 and 13).** Cancel a decode that is already running, decode at the drawn size, bound the cache by bytes, and keep the evidence card's loaded likeness.
- **Go crop work (PRF-4).** Store crops below 2,048 px. (Decoding the crop before the catalog session, PRF-3, and one stage-then-write entry point went to Spike 12 S12-27, 2026-10-27.)
- **`CitationComposerViewerPane.regionSymbol` repeats `ArtifactRegionTool.symbol`.** The composer maps a region kind to a glyph a second time; one mapping on the kind would serve both.
- **`CitationLocatorDraft.encodeJSON()` key order.** It serializes without `.sortedKeys`, so two encodings of one locator can differ as strings.

## PDF Artifact thumbnails

Descoped from Spike 8 (**S8-02** retired; do not reuse the id). PDF Artifacts keep the **file-type glyph**. Image Files still get a raster derivative. A Source cover stays a user-pinned raster (or the type icon) — so a PDF still cannot be cover until an engine renderer exists.

Telling two PDFs apart on Artifact rows is real, but not worth a client-only path. macOS **PDFKit** → `file_derivatives` would leave Windows on the glyph and put generation in the frontend. Thumbnails are catalog derivatives; they belong in **Go** if they come back, same as JPEG/PNG today.

**If revived:** Page 1 (media box) only unless a later story picks a cover page. Same `ThumbnailSpec()` as images: JPEG, longest edge ≤ 256. Persist as unaudited `file_derivatives.thumbnail`. Encrypted, empty, corrupt, or over the source-byte budget → skip, keep the glyph. **Generate in `core/derivatives`**, not Swift. No `PutFileThumbnail` client bytes path. Out: Quick Look, video / Office posters, user-chosen thumb page, PDF OCR. Image pipeline: [`core/derivatives`](../../core/derivatives/).

## Still out

Audio and video playback ([`audio-video-sources.md`](audio-video-sources.md)). A video frame plus a region may later be a depiction. The player is separate.
