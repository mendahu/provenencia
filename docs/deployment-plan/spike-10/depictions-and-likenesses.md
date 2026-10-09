# Depictions and likenesses

**Status:** pulled into [Spike 10](deployment-plan.md). The PR split and the calls for this spike are there. This note stays the design.

A photograph of a person, a house, a gravestone, or a wedding should be citable evidence whose value is the picture. Today a Citation can already draw a polygon around that region, and the composer still expects the region to become text.

## Problem

The evidence chain is Source → Artifact → Citation → Observation.

- The Artifact holds the scan or born-digital image ([`source-layer-data-model.md`](../../source-layer-data-model.md)).
- The Citation locator can be `artifact`, optional `page`, and a `region` polygon ([`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) §3). That polygon is the crop: a face, a house, a seal.
- `transcription` is the reading of words or speech. It is nullable. The composer treats a Citation as something to transcribe.
- Observation values are `text`, `integer`, `date`, `name`, `subject`, or `term`. None of them is a picture.

A group photo is the case that hurts. One Artifact, several faces. Each face is a polygon. The researcher wants those crops on the person, not a sentence that describes the face. A `likeness` property key alone still has to store one of the existing value types, so it would be a caption.

## Shape

**`depiction` is a value type.** The Observation does not gain an image column and does not copy file bytes. Its value is its own Citation: the Artifact plus the locator. The UI renders that region. Empty `transcription` is valid. Optional `description` can still say “seated man, left.”

**`likeness` is the seeded person property of that type.** That key is what puts a portrait on a person. Place, event, and signature keys are later ([`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md)). Rendering, empty transcription, and side-by-side comparison hang off the value type. The key decides where it shows up.

```text
Photograph Source
  Artifact: scan.jpg
  Citation: polygon around one face
    transcription empty
  Observation
    person subject N1 -- likeness --> this Citation's region
```

A group photo is several Citations on one Artifact, one likeness Observation per person. The whole frame can be a depiction of the event.

Several likenesses of one person are a gallery. They do not compete the way two birth dates do, and they do not need a Reconciliation Claim to be useful.

One of them can be the preferred portrait. That is a display pointer on the person, not a claim: the gallery still holds every likeness. The Person page lead and the Persons list thumb use that crop. With no choice, both use the earliest likeness. Removing the chosen likeness clears the pointer and the thumb falls back. The choice is made on the Person page.

On promote, a likeness lines up like a name: Promote's evidence sheet ([`promote-graph-alignment.md`](../../promote-graph-alignment.md)) shows the two crops as one comparison, and pinning it pins the two Observations that already exist. It does not create a picture and it does not write a transcription.

## Crop cache

The person page finds pictures through the evidence chain, then shows a small raster. It should not open the original scan and reapply the polygon on every visit.

`file_derivatives` today is one row per source File and derivative type (`UNIQUE (source_file_id, derivative_type)` — one `thumbnail` per File). A photograph has many polygons, so a crop cannot use that key. The cache key is the **File plus the locator**.

Same content-addressed `files` store. Disposable, not audited, not an Artifact, not the Observation’s value. Purge it and the Citation still names the face. If the polygon changes, that cache entry is stale and a new crop is generated. Two Citations with the same pixels can share the stored bytes. The derivative rows stay separate.

Mint the crop when the depiction Observation is saved. The composer is the right moment for a PDF: the document is already open. `PDFView` is a live view, not a bitmap, so save asks PDFKit to draw the locator into an image at a fixed size. That does not require a PDF renderer in the Go core. A JPEG scan can be clipped the same way, in the composer or in Go. A missing cache is filled the next time a client that can open that Artifact draws it again.

Spike 10 saves the bounding box of the polygon as a JPEG, longest edge 512. A transparent polygon crop is later ([`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md)).

Person page, cache present:

```text
accepted identity claims for the person
  → depiction Observations on those member subjects
  → each Observation's Citation
  → crop cached for that File and locator
```

That is one join, then a handful of small rasters. The slow path is a missing cache, not the query.

## Settled (do not reopen)

| Call | Why |
| --- | --- |
| **Value type, not only a property key** | `likeness` still needs somewhere to put the picture. Existing value types are text-like or structured scalars. |
| **The Citation is the value** | The locator is already the crop. Copying bytes onto the Observation makes a second original. |
| **Crop is a derivative cache** | Person page stays fast. Evidence stays the Artifact plus the polygon. |
| **Cache key is File + locator** | One File has many regions. The current one-type-per-File unique key cannot hold them. |
| **Composer draws the PDF crop** | The document is open. Go does not grow a PDF engine for this. |
| **Gallery, not a single concluded face** | Photos accumulate. Reconciliation stays for values that compete. |
| **Preferred portrait is a display pointer** | One likeness leads the page and the list. The others stay in the gallery. It is not a Reconciliation Claim. |

## Later

Place and event keys, signatures, a transparent polygon, and the Windows PDF draw are in [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md).

## Explicitly out of this note

- Audio/video playback ([`audio-video-sources.md`](../../ideas/audio-video-sources.md)) — a video frame plus region may later be a depiction; the player is a separate idea
- Go-side PDF thumbnail generation ([`artifact-pdf-thumbnails.md`](../../ideas/artifact-pdf-thumbnails.md))
- Text-quote locators ([`text-quote-locators.md`](../../ideas/text-quote-locators.md))

## Related docs

- [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) — Citations, locators, Observation value types
- [`source-layer-data-model.md`](../../source-layer-data-model.md) — `files`, `file_derivatives`
- [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) — members and Reconciliation, when a likeness is shown on a person
