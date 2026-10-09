# Depictions and likenesses — later

**Status:** idea, not scheduled. Spike 10 takes the value type, the `likeness` property, the bounding-box crop cache, the composer row, the Person gallery, a preferred portrait, and Promote's side-by-side crops. Design for that work: [`spike-10/depictions-and-likenesses.md`](../deployment-plan/spike-10/depictions-and-likenesses.md).

## Still open

- **Other keys of the same value type.** `likeness` is the person key. A photo of a place or an event needs its own key, or one shared `depiction` key. The value type does not decide that.
- **Signature.** Same question: another key of type `depiction`, or something of its own.
- **Transparent polygon.** Spike 10 saves the bounding box of the region as a JPEG. A crop that follows the polygon, with transparency outside it, can replace that once the gallery exists.
- **Windows.** The cache row is the same. That client draws the PDF locator with its own stack at save time. Spike 10 draws it with PDFKit on the Mac.

## Still out

Audio and video playback ([`audio-video-sources.md`](audio-video-sources.md)). A video frame plus a region may later be a depiction. The player is separate.

Go-side PDF thumbnail generation ([`artifact-pdf-thumbnails.md`](artifact-pdf-thumbnails.md)). Spike 10 does not put a PDF engine in Go.
