# Audio and video Sources

**Status:** idea only — not roadmapped. Parked here from [`interpretation-graph-ui.md`](interpretation-graph-ui.md) leftover **16**. Not in Spike 8.

Oral history, recorded interviews, home movies, and digitized tapes are real Sources. The catalog already accepts those Files. The app does not play them, cite a moment in them, or show more than a type glyph. This note is the whole media cluster — not only a locator type.

## Problem

A Source may have an audio or video Artifact today (ingest MIME already allows it). The composer classifies `audio` / `video` and shows a coming-soon empty state. Save refuses those kinds. There is no player, no transport, no `time_range` write, no poster thumbnail.

The model already knows the pin:

```json
{
  "type": "time_range",
  "start_ms": 802400,
  "end_ms": 845100
}
```

Authoritative schema: [`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md) §3.5. Milliseconds, `end_ms > start_ms`. A video Citation may compose `time_range` + `region` (a frame interval *and* a box on that frame). Go `core/locator` does not yet type-check `time_range` (unknown types are preserved); that lands with the first writer.

Image Vision and PDF Find do not apply. Speech-to-text, if we ever want it, is a different recognizer than `VNRecognizeTextRequest`.

## What this idea still owns

1. **Composer playback** — AVKit `AVPlayer` (or equivalent) in the viewer pane: play / pause / scrub, current time, duration. Replace the coming-soon empty state.
2. **`time_range` writer** — mark in / mark out (or drag on a timeline) → append the selector. Reopen a Citation and seek to that interval.
3. **Video frame + region** — optional: pause, draw a polygon on the current frame, store `time_range` + `region` as the model already allows.
4. **Transcription** — the Citation reading may be a spoken excerpt. Manual type-in first. Dictation / speech-to-text later, not Vision OCR, not Spike 8 Auto Transcribe.
5. **Thumbnails / cover** — a video poster or audio glyph that is not a blank `PVThumbnail`. User-picked poster frame is a sub-item (PDF cover page has the same shape).
6. **Source-page playback** — whether filing can play the Artifact without opening the composer. Nice; not required to cite.
7. **Locator validation** — teach `core/locator` `time_range` when the client first writes it.

## Settled (do not reopen)

| Call | Why |
| --- | --- |
| **Do not use QuickLook as the composer** | `QLPreviewView` has no selection or coordinate API. Leftover **17** is **descoped**. Image and PDF already have real viewers; A/V gets AVKit when this idea ships. |
| **Do not fold this into Spike 8** | Image OCR and PDF text-layer work stay independent. Audio/video stay disabled there. |
| **Locator schema is already specified** | No new selector type. Implementation is player + writer + validate. |

## Open questions

- In/out marks vs a range slider vs both?
- Does `artifact`-only remain legal for “cite the whole tape,” same as a whole image?
- Waveform / filmstrip chrome, or transport only for v1?
- Speech-to-text: on-device only, or never until we have a privacy story?
- Windows later: same locator JSON; player is a different stack.

## Explicitly out of this note

- Spike 8 image Auto Transcribe and PDF Find/paste
- [`text-quote-locators.md`](text-quote-locators.md)
- QuickLook preview (descoped)
- Changing `citations` / adding a new locator type

## Related docs

- [`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md) §3.5
- [`interpretation-graph-ui.md`](interpretation-graph-ui.md) §6
- [`source-layer-data-model.md`](../source-layer-data-model.md) (audio/video as Artifact representations)
- [`research-judgment-model.md`](../research-judgment-model.md) (`transcription_uncertain` for garbled audio / muddy video)
- Spike 7 composer: image/PDF only; A/V empty state shipped
