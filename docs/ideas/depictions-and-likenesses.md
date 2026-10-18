# Depictions and likenesses — later

**Status:** idea, not scheduled. Spike 10 shipped the `depiction` value type, the `likeness` property, the composer row, the Person gallery, and a preferred portrait ([decisions](../deployment-plan/archive/spike-10/)). Promote comparing two likenesses is Spike 11 (S11-20).

## Still open

- **Other keys of the same value type.** `likeness` is the person key. A photo of a place or an event needs its own key, or one shared `depiction` key. The value type does not decide that.
- **Signature.** Same question: another key of type `depiction`, or something of its own.
- **Transparent polygon.** Spike 10 saves the bounding box of the region as a PNG. A crop that follows the polygon, with transparency outside it, can replace that once the gallery exists.
- **Every saved face on the photo.** The composer's viewer shows one region at a time; showing every saved depiction on the photo needs a read-only region layer.
- **Gallery image from the record.** The gallery reads the image File id from the reconciler cache; joining it to the record's own `depiction_values.image_file_id` keeps an edited cache row from showing another PNG (audit SEC-4, 2026-10-18).
- **A hidden value's reason and support.** A hidden reconciler row takes its first entry's reason and counts every entry's Source, so a value carried by an accepted weak photo reads `provisional` when a provisional member's copy of the same picture ranks above it. The gallery reads each record's own outcome instead (S10-60); other readers of `auto_reconciler_values.reason` still see the row's.
- **One media type for the stored image.** `ingest.StageImage` writes `image/png` as a literal in three places and `depictionvalue.ReadableImage` checks the same literal; one constant in `core/depictionvalue` would tie them, and the transparent-polygon crop would change one place.
- **Windows.** That client draws the PDF region with its own stack before it sends the image. Spike 10 draws it with PDFKit on the Mac.

## Still out

Audio and video playback ([`audio-video-sources.md`](audio-video-sources.md)). A video frame plus a region may later be a depiction. The player is separate.

Go-side PDF thumbnail generation ([`artifact-pdf-thumbnails.md`](artifact-pdf-thumbnails.md)). Spike 10 does not put a PDF engine in Go.
