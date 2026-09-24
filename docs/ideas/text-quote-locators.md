# Text-quote locators

**Status:** idea only — not roadmapped. Parked here from [`interpretation-graph-ui.md`](interpretation-graph-ui.md) leftover **15**. Not in Spike 8.

The locator vocabulary already includes `text_quote`. Go validates it. The client never writes it. This note is the home for *when* a Citation should pin words instead of (or as well as) a polygon.

## Problem

A Citation always has a locator. Today the client writes `artifact`, optional `page`, optional `region`. That is the right pin for a scan: “this ink on this page.”

A born-digital PDF, book, or other text-layer document is different. The evidence *is* a phrase. Drawing a polygon around “William Robins, carpenter” is slower than selecting the words, leaks neighboring columns, and breaks if the Artifact is replaced with a better export of the same edition.

`text_quote` selects by text, not by rendered coordinates:

```json
{
  "type": "text_quote",
  "exact": "William Robins, carpenter",
  "prefix": "household of ",
  "suffix": " aged 43"
}
```

Authoritative schema: [`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md) §3.6. Validation: [`core/locator`](../../core/locator/). W3C-style `exact` + optional `prefix` / `suffix` so a repeated “John” is not ambiguous.

Spike 8 **Find / I-beam / paste transcription** does **not** seed this. Those fill the reading. They do not change `locator_json`.

## When a quote is better than a polygon

- **Text-layer documents** — registers, books, HTML, plain text. I-beam select is the natural gesture.
- **The Artifact can change** — better scan, re-export, different crop of the same edition. The polygon moves; the phrase still matches (page + prefix/suffix if needed).
- **The locator is readable** — `exact` *is* the snippet. A list of normalized points is not. Audit, share packages, “jump back to this Citation.”
- **Find → cite** — once `PDFSelection` exists, “this selection is the locator” is almost free. A selection → polygon is a lossy bounding box.

## When the polygon still wins

- **Scans and image-only PDFs** — no honest text layer. Most genealogy pages.
- **Visual evidence** — handwriting, stamps, household blocks, overwritten dates, “the name in the margin.”
- **Jumping the camera** — a region is a rectangle to scroll to. A quote has to search the text layer and pick a hit.

They are not exclusive. A locator can already be `page` + `text_quote`, or quote *and* region (identity vs “scroll me to this ink”).

## What this idea still owns

1. **Writer UI** — I-beam selection → append `text_quote` (and keep `page` when paginated). Prefix/suffix from surrounding `PDFSelection` context, as §6 already guessed.
2. **Resolver / highlight** — reopen a Citation and find the quote on the live page (or fail honestly if the text layer changed).
3. **Image / scan policy** — do not invent a quote from Vision. OCR can fill transcription; it is not a locator unless we decide that later.
4. **Relationship to transcription** — whether `exact` should stay aligned with the transcription field, and what happens when the researcher edits one and not the other.
5. **Ambiguous hits** — `exact` matches more than once even with prefix/suffix; researcher pick vs refuse.

## Open questions

- Auto-write a quote whenever the researcher selects text, or an explicit “pin selection” control?
- May a Citation have both `region` and `text_quote`?
- What if Find is on a hit but the researcher never selected — seed from the current hit?
- Plain-text / HTML Artifacts (no PDFKit page) — quote-only, no region?

## Explicitly out of this note

- Spike 8 **S8-D2** / **S8-03…S8-05** (Find, I-beam, paste into transcription only)
- PDF OCR / Vision on page rasters
- `time_range` / audio / video — [`audio-video-sources.md`](audio-video-sources.md)
- Changing the locator schema (already specified)

## Related docs

- [`interpretation-layer-data-model.md`](../interpretation-layer-data-model.md) §3.6
- [`interpretation-graph-ui.md`](interpretation-graph-ui.md) §6
- [`core/locator`](../../core/locator/)
- Spike 8 Find (not this): [`../deployment-plan/spike-8/pdf-text-find.md`](../deployment-plan/spike-8/pdf-text-find.md)
