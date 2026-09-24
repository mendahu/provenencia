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
- **Wanted:** Layers or filters when a Source actually hurts. Not scheduled. Parked here until dogfood proves we need it.

### Automatic structure from the Artifact (beyond transcription)

- **Date:** 2026-09-23
- **Where:** Citation composer / Evidence graph
- **Annoyance:** Even after transcription is filled, the researcher still retypes names, dates, and roles onto cards. Spike 8 only dumps text into **transcription**.
- **Wanted:** Cheap field hints after there is a string (`NSDataDetector`, name/date patterns) and, later, guided extract. Not catalog writes.
- **Still out of Spike 8:** PDF OCR / Vision on page rasters; Live Text / VisionKit; Foundation Models (`macOS 26+`, raise-the-floor). Pipeline if we ever did it: string → suggestions the researcher accepts → real Citation + Observations. Experiment: “obituary → draft cards,” not “obituary → catalog writes.”

## Done

### Pulled into Spike 8 (2026-09-24)

- **Composer rethink** — Citation as the document; multi-subject rows; reuse / pinning; empty Save ([`S8-D7`](../deployment-plan/spike-8/design/archive/S8-D7-composer-rethink.md) / [`S8-10`](../deployment-plan/spike-8/deployment-plan.md))
- Image **Auto Transcribe** (Vision → transcription)
- PDF **Find** + **select** + **paste transcription**
- Source page **Open Evidence graph** (was “graph starts blind”)
