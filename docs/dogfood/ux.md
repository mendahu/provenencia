# UX notes

Scratch log of annoyances from dogfooding. Newest at the top of **Open**. Not a commitment and not a test plan.

## How to add

Copy the stub. Surface + what got in the way is enough. A wanted fix is optional.

```markdown
### Short title

- **Date:** YYYY-MM-DD
- **Where:** Evidence graph / citation composer / Sources list / …
- **Annoyance:** What you were doing and what felt wrong.
- **Wanted:** (optional) What would have been better.
```

## Open

<!-- add below this line -->

### Dense Evidence graphs (census-scale)

- **Date:** 2026-09-24
- **Where:** Evidence graph
- **Annoyance:** A real census page can put dozens of Subjects and hundreds of Observations on one canvas. Nothing is wrong with the model — it just gets hard to see the household you care about.
- **Wanted:** Layers or filters when a Source actually hurts. Not scheduled (descoped from the interpretation-graph-ui leftover list). Parked here until dogfood proves we need it.

### One reading, many subjects — composer is per-subject

- **Date:** 2026-09-23
- **Where:** Evidence graph / citation composer
- **Annoyance:** A newspaper engagement notice is one reading (names, relationships, the engagement event) but the composer is tied to one subject. Observations that belong on other cards required a second Citation and a second trip through the composer. Filling the model meant typing the same names, roles, and event over and over. Starting a graph is also blind: there is no source in view, so you either bounce back to the Source page and remember, or create a card just so Add property can open the artifact.
- **Wanted:** Keep the one-Citation-to-many-Observations model (including Observations on different subjects) and make a Source-reading session cheap: see the artifact while placing, reuse one Citation across the cards it actually supports, and stop re-keying the same facts to satisfy the graph. This may be a fairly large reimagining of the graph + composer together, not a small composer tweak. The schema already allows shared Citations; Spike 7 deferred pinning across graph edits.
- **Also wanted:** Some automatic data entry from the Artifact itself (image / PDF). The text is already on the page. An LLM is the obvious extractor, but it is not obvious how far we can get without one — even simple field population from a text layer (regex, known date/name patterns, paste-from-selection) would cut the retyping. Image-only scans have no text until OCR.
- **Pulled into Spike 8:** image **Auto Transcribe** ([`S8-D1`](../deployment-plan/spike-8/design/S8-D1-auto-transcribe.md) / [`S8-01`](../deployment-plan/spike-8/deployment-plan.md)); PDF **Find** + **select** + **paste transcription** ([`S8-D2`](../deployment-plan/spike-8/design/S8-D2-pdf-text-find.md) / [`S8-03…S8-05`](../deployment-plan/spike-8/deployment-plan.md); notes: [`pdf-text-find.md`](../deployment-plan/spike-8/pdf-text-find.md)). PDF OCR / Vision on page rasters stay out. Shared Citation, graph+composer rethink, and Foundation Models stay here.

**OCR / text off the Artifact (macOS, no third party required):**

- **PDFKit first** — not OCR. A born-digital PDF already has a text layer (`string`, `selection(for:)`, `findString`). A newspaper scan in a PDF is just pictures of pages; you still have to rasterize a page and run Vision.
- **Vision `VNRecognizeTextRequest`** — Apple on-device OCR, system framework, available on our **macOS 14** deployment target. Input a `CGImage`; get lines with bounding boxes, confidence, and alternate candidates. Fast vs accurate, languages, language correction, `customWords` (project surnames would fit). No cloud, fits offline-first.
- **Newer Vision** — `RecognizeTextRequest` is the same family with newer Swift concurrency. **`RecognizeDocumentsRequest`** (paragraphs / tables / lists) is **macOS 26+**, so it is off the table unless we raise the deployment target.
- **After you have a string:** `NSDataDetector` (dates, addresses) and Natural Language (tokens, weak name hints). That is the “regex or something” layer. It will not turn an engagement notice into people + relationship + event.
- **Live Text / VisionKit** — same Vision stack with system “select text on an image” UI. Useful for click-to-transcribe, not for filling the graph.
- **What Apple will not do:** invent Observations. You get text and geometry. Mapping onto subjects still needs our rules or an LLM. Third-party OCR (Tesseract, cloud) is optional if Vision fails a class of scans — not the default, and it fights offline-first more than Vision does.

Practical split: PDF text layer when it exists; Vision on the page image when it does not; detectors for cheap date/name fills; on-device Foundation Models only if we want structure (and only after we have a string).

**Apple Intelligence / Foundation Models (structured extract, not catalog writes):**

- **`import FoundationModels`** — on-device Apple Intelligence LLM. Apple lists entity extraction as a designed job (with summarization and classification). Small on-device model (~3B, quantized), not a cloud reasoner. Stays on-device unless we opt into Private Cloud Compute for a harder prompt.
- **Guided generation** is the fit for an obituary. `@Generable` / `@Guide` Swift types plus constrained decoding: you ask for people, events, dates, stated relationships and get `[PersonDraft]` / `[EventDraft]`, not a paragraph to parse. **Dynamic schema** could be built from our seeded Properties and terms (`relationship_type`, date kinds) instead of a one-off prompt. The content-tagging adapter is topics/actions, not genealogy.
- **It does not know Provenencia.** It will not mint subjects, pick locators, or keep one Citation honest. 19th-century newspaper voice, abbreviations, and “the late widow of…” will be uneven. Context is small: one obituary is fine; a whole page of classifieds may not be. Treat output as **suggestions the researcher accepts**.
- **Availability:** **macOS 26+**, Apple Intelligence–capable Mac, Intelligence actually enabled. Our deployment target is **14.0**. This is a raise-the-floor decision and a “some users have no model” path — not a drop-in on today’s app.
- **Pipeline if we ever did it:** PDF text layer or Vision OCR → string → Foundation Models guided extract using our vocabulary → researcher confirms → real Citation + Observations. The LLM does not see pixels unless we add a separate image-understanding step.

Interesting experiment: “obituary → draft cards.” Not “obituary → catalog writes.”

## Done

_Nothing yet._
