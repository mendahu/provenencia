# Depictions and likenesses

**Status:** pulled into [Spike 10](deployment-plan.md). The PR split and the calls for this spike are there. This note stays the design. Reshaped 2026-10-16: a depiction is an owned value row holding the region and an extracted image, not a pointer back to the Citation.

A photograph of a person, a house, a gravestone, or a wedding should be citable evidence whose value is the picture. Today a Citation can already draw a polygon around a region, and the composer still expects the Citation to become text.

## Problem

The evidence chain is Source → Artifact → Citation → Observation.

- The Artifact holds the scan or born-digital image ([`source-layer-data-model.md`](../../source-layer-data-model.md)).
- The Citation locator can be `artifact`, optional `page`, and a `region` polygon ([`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) §3).
- `transcription` is the reading of words or speech. It is nullable. The composer treats a Citation as something to transcribe.
- Observation values are `text`, `integer`, `date`, `name`, `subject`, or `term`. None of them is a picture.

A group photo is the case that hurts. One Artifact, several faces. The researcher wants those faces on the people, not a sentence that describes each face. A `likeness` property key alone still has to store one of the existing value types, so it would be a caption.

## Shape

**`depiction` is a value type with its own value row.** A depiction is stored the way a date or a name is: the Observation owns one `depiction_values` row (`observations.value_depiction_id`). The row holds two things:

- **The region**: page and polygon on the Citation's Artifact, validated by `core/locator`. It records where the picture came from, so the original stays the evidence and the face can be shown in place, re-extracted, or checked.
- **The image**: a File holding the extracted picture. It is a `depiction` derivative of the Artifact's File, recorded in `file_derivatives`, so the lineage from the original scan is explicit.

**The Citation is the photo, not the face.** A Citation on a photograph cites the photo or the page, with a transcription when there is one (*Wedding party, 1902*), or none. Each face is a likeness Observation on that Citation with its own region and image. A group photo is one Citation with one likeness per person, the same way one passage yields several names.

**`likeness` is the seeded person property of that type**, cardinality `multiple`. That key is what puts a portrait on a person. Place, event, and signature keys are later ([`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md)). Rendering, the image, and side-by-side comparison hang off the value type. The key decides where it shows up.

```text
Photograph Source
  Artifact: wedding.jpg            (File F)
  Citation: the photo              transcription "Wedding party, 1902" or empty
    Observation  N1 -- likeness --> depiction value: region R1, image C1
    Observation  N2 -- likeness --> depiction value: region R2, image C2
  file_derivatives: F → C1 (depiction), F → C2 (depiction)
```

Several likenesses of one person are a gallery. They do not compete the way two birth dates do. The auto-reconciler reconciles them like any value (one image is one value; negatives, provisional members, and weak evidence apply), and `multiple` keeps every distinct one.

One of them can be the preferred portrait. That is a display pointer on the person, not a claim: the gallery still holds every likeness. The Person page lead and the Persons list thumb use that image. With no choice, both use the earliest likeness. Removing the chosen likeness clears the pointer and the thumb falls back. The choice is made on the Person page.

On promote, a likeness lines up like a name: Promote's evidence sheet ([`promote-graph-alignment.md`](../../promote-graph-alignment.md)) shows the two images as one comparison, and pinning it pins the two Observations that already exist. It does not create a picture and it does not write a transcription.

## The image

**The client extracts; Go stores its own copy.** The composer already has the Artifact open, so it renders the crop: a pixel crop for an image, and PDFKit drawing the page region for a PDF. Go has no PDF engine and does not grow one. The client sends the crop the way Artifact import sends a file: as a path to a file it wrote, never as bytes over the FFI.

**Client bytes are untrusted** (user, 2026-10-16). The crop goes through every protection a file import has: an absolute path, no symlinks, a regular file, a size cap, and the content sniff and allowlist. Go then decodes it with the raster package's format allowlist and pixel budget, checks the longest edge, and stores **its own re-encoded image**, not the client's bytes.

**Format:** the polygon's bounding box, PNG, longest edge at most 2048. A transparent polygon mask is later; the polygon itself is kept in the region.

**Lifetime follows the evidence chain.** The image is stored in the same write as its Observation, so a failed save leaves no file behind. A region edit stores a new image; the old one is released when nothing else points at it. Deleting the Observation releases its value row, its derivative link, and the image File when unused. A Citation cannot be deleted while it has Observations, so by the time a Citation goes, every image made from it has gone. Identical pixels share one stored File (the store is content-addressed); the value rows and derivative links stay separate.

Person page:

```text
the person's cached likeness values (auto_reconciler_values, kept or weak)
  → each value's image File
```

One read, then a handful of small images. No crop cache, and no "fill it in the next time a client can open the Artifact."

## Settled (do not reopen)

| Call | Why |
| --- | --- |
| **Value type, not only a property key** | `likeness` still needs somewhere to put the picture. Existing value types are text-like or structured scalars. |
| **The value is an owned row: region plus image** | Every value type stores its value; a depiction is no exception. The region keeps the original as the evidence. The image is what every reader shows. (Replaces "The Citation is the value", 2026-10-16.) |
| **The Citation is the photo; each face is an Observation** | One photo, many faces, like one passage with many names. (Replaces "several Citations on one Artifact".) |
| **The image is a derivative of the Artifact's File** | `file_derivatives` already records lineage from an original; many depictions per File, released when nothing points at them. (Replaces the crop cache keyed by File plus locator.) |
| **The client renders; Go validates and re-encodes** | The document is open in the composer. Go does not grow a PDF engine, and does not store bytes it did not produce. |
| **Gallery, not a single concluded face** | Photos accumulate. `likeness` is cardinality `multiple`, so the auto-reconciler keeps every distinct image; a Reconciliation Claim stays for values that compete. |
| **Preferred portrait is a display pointer** | One likeness leads the page and the list. The others stay in the gallery. It is not a Reconciliation Claim. |

## Later

Place and event keys, signatures, a transparent polygon, and a Windows PDF draw are in [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md).

## Explicitly out of this note

- Audio/video playback ([`audio-video-sources.md`](../../ideas/audio-video-sources.md)) — a video frame plus region may later be a depiction; the player is a separate idea
- Go-side PDF thumbnail generation ([`artifact-pdf-thumbnails.md`](../../ideas/artifact-pdf-thumbnails.md))
- Text-quote locators ([`text-quote-locators.md`](../../ideas/text-quote-locators.md))

## Related docs

- [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) — Citations, locators, Observation value types
- [`source-layer-data-model.md`](../../source-layer-data-model.md) — `files`, `file_derivatives`
- [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) — members and Reconciliation, when a likeness is shown on a person
- [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) — value-type modules
