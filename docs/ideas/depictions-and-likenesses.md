# Depictions and likenesses — later

**Status:** idea, not scheduled. Spike 10 shipped the `depiction` value type, the `likeness` property, the composer row, the Person gallery, and a preferred portrait ([decisions](../deployment-plan/archive/spike-10/)). Promote comparing two likenesses is Spike 11 (S11-20).

## Still open

- **Other keys of the same value type.** `likeness` is the person key. A photo of a place or an event needs its own key, or one shared `depiction` key. The value type does not decide that.
- **Signature.** Same question: another key of type `depiction`, or something of its own.
- **Transparent polygon.** Spike 10 saves the bounding box of the region as a PNG. A crop that follows the polygon, with transparency outside it, can replace that once the gallery exists.
- **Every saved face on the photo.** The composer's viewer shows one region at a time; showing every saved depiction on the photo needs a read-only region layer.
- **Windows.** That client draws the PDF region with its own stack before it sends the image. Spike 10 draws it with PDFKit on the Mac.

## Still out

Audio and video playback ([`audio-video-sources.md`](audio-video-sources.md)). A video frame plus a region may later be a depiction. The player is separate.

Go-side PDF thumbnail generation ([`artifact-pdf-thumbnails.md`](artifact-pdf-thumbnails.md)). Spike 10 does not put a PDF engine in Go.
