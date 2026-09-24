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

### One reading, many subjects — composer is per-subject

- **Date:** 2026-09-23
- **Where:** Evidence graph / citation composer
- **Annoyance:** A newspaper engagement notice is one reading (names, relationships, the engagement event) but the composer is tied to one subject. Observations that belong on other cards required a second Citation and a second trip through the composer. Filling the model meant typing the same names, roles, and event over and over.
- **Wanted:** Keep the one-Citation-to-many-Observations model (including Observations on different subjects) and make a Source-reading session cheap: see the artifact while placing, reuse one Citation across the cards it actually supports, and stop re-keying the same facts to satisfy the graph. This may be a fairly large reimagining of the graph + composer together, not a small composer tweak. The schema already allows shared Citations; the shipped composer is one Citation × one subject per trip.
- **Folded in:** **Citation pinning** across graph edits and **Citation with zero Observations** (“I transcribed this line; I have not interpreted it yet”). Today Save refuses an empty Observation list. Both are UI policy on a per-subject composer, not schema gaps. A more flexible composer absorbs them; do not schedule a standalone pinning or empty-Citation story.

### Automatic structure from the Artifact (beyond transcription)

- **Date:** 2026-09-23
- **Where:** Citation composer / Evidence graph
- **Annoyance:** Even after transcription is filled, the researcher still retypes names, dates, and roles onto cards. Spike 8 only dumps text into **transcription**.
- **Wanted:** Cheap field hints after there is a string (`NSDataDetector`, name/date patterns) and, later, guided extract. Not catalog writes.
- **Still out of Spike 8:** PDF OCR / Vision on page rasters; Live Text / VisionKit; Foundation Models (`macOS 26+`, raise-the-floor). Pipeline if we ever did it: string → suggestions the researcher accepts → real Citation + Observations. Experiment: “obituary → draft cards,” not “obituary → catalog writes.”

## Done

### Pulled into Spike 8 (2026-09-24)

- Image **Auto Transcribe** (Vision → transcription)
- PDF **Find** + **select** + **paste transcription**
- Source page **Open Evidence graph** (was “graph starts blind”)
