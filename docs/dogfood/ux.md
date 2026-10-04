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

### Evidence graph misrenders when zoomed far out

- **Date:** 2026-10-04
- **Where:** Evidence graph
- **Annoyance:** Zooming far out breaks the rendering in two ways:
  - **Background grid:** parts of it turn solid black and the grid disappears in those areas.
  - **Lines between cards:** past a certain zoom level they start to disappear, but unevenly. Some vanish and others stay.
- **Wanted:** Zoomed-out views look refined and polished. The grid stays visible (or fades out cleanly), and lines behave the same way at every zoom level.

### Evidence graph button hit boxes are wrong

- **Date:** 2026-10-04
- **Where:** Evidence graph (toolbar, bridge cards)
- **Annoyance:** Button hit boxes still don't match the buttons. Two cases stand out:
  - **Toolbar (top left):** Add a place, Add a person, Add an event, Connect. Only the text or icon takes the click. The rest of the button does nothing.
  - **Bridge cards:** the button that opens the citation for the bridge relationship has a hit box shifted about one button-width to the left. Clicking the button does nothing; clicking the empty space to its left opens the citation.
- **Wanted:** The whole visible button accepts the click, and the hit box sits exactly on the button.

### Evidence graph viewport resets after visiting the composer

- **Date:** 2026-10-04
- **Where:** Evidence graph ↔ Citation composer
- **Annoyance:** I pan around the Evidence graph, zoom into a card, open the composer to add a property, then go back. The graph's view has been reset instead of staying where I left it, which is disorienting. This happens whether I use the back button or the breadcrumbs.
- **Wanted:** Coming back puts me at the exact same center coordinates and zoom level. Longer term, the last viewport for each Evidence graph (center + zoom) should be saved persistently per user, either in a dedicated table or by reusing wherever card coordinates are already stored.

### Omnibar ranks fuzzy name matches above exact title words

- **Date:** 2026-10-04
- **Where:** Omnibar search
- **Annoyance:** Searched `marriage` expecting a marriage certificate and a newspaper marriage announcement. Both have "marriage" in the title. They landed at the bottom of the results, under obituaries and memorials for someone named **Marion Margaret**. Looks like the shared leading "Mar…" on both names stacked up and beat the real hits. That feels like a bug: I typed the whole word, so an exact word match in a title should win.
- **Wanted:** A whole-word or exact title match always ranks above fuzzy or prefix matches. Partial matches on several tokens (Marion + Margaret) should not add up to more than one exact hit.

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
- **Still out:** PDF OCR / Vision on page rasters; Live Text / VisionKit; Foundation Models (`macOS 26+`, raise-the-floor). Pipeline if we ever did it: string → suggestions the researcher accepts → real Citation + Observations. Experiment: “obituary → draft cards,” not “obituary → catalog writes.”

## Done

### Pulled into Spike 8 (2026-09-24)

- **Composer rethink** — Citation as the document; multi-subject rows; reuse / pinning; empty Save ([Spike 8](../deployment-plan/archive/spike-8/))
- Image **Auto Transcribe** (Vision → transcription)
- PDF **Find** + **select** + **paste transcription**
- Source page **Open Evidence graph** (was “graph starts blind”)
