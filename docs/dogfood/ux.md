# UX notes

Scratch log of annoyances from dogfooding. Newest at the top of **Open**. Not a commitment and not a test plan.

Once a note is pulled into a spike or an idea, delete it here. The plan is the copy.

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

### Guarding a Citation page change after depictions are saved on it

- **Date:** 2026-10-17
- **Where:** Citation composer
- **Annoyance:** A likeness is drawn on the Citation's PDF page. Moving the Citation to another page leaves the saved likenesses pointing at a region that no longer matches; Go hides it and the edit dialog asks for a redraw, but nothing warns before the page changes (deferred from S10-34).
- **Wanted:** A confirm before the page change when depictions are saved on it.

### Folding the other test copies of PNG / PDF writers and `waitUntil` into `ArtifactFileFixtures`

- **Date:** 2026-10-17
- **Where:** Tooling (macOS tests)
- **Annoyance:** `ArtifactFileFixtures` now holds the PNG and PDF writers and `waitUntil`, but older tests keep their own copies: `ArtifactViewerModelTests.swift`, `PromoteModelTests.swift`, `CitationAutoTranscribeTests.swift`, `EvidenceGraphModelTests.swift`, and `seedArtifact` in `CitationComposerModelTests.swift` (deferred from S10-34).
- **Wanted:** One shared copy.

### No keyboard way to add one card to a multi-selection

- **Date:** 2026-10-17
- **Where:** Evidence graph
- **Annoyance:** Shift-click adds or removes one card, but keyboard and VoiceOver users have no way to do the same (deferred from S10-57).

### Automatic structure from the Artifact (beyond transcription)

- **Date:** 2026-09-23
- **Where:** Citation composer / Evidence graph
- **Annoyance:** Even after transcription is filled, the researcher still retypes names, dates, and roles onto cards. Spike 8 only dumps text into **transcription**.
- **Wanted:** Cheap field hints after there is a string (`NSDataDetector`, name/date patterns) and, later, guided extract. Not catalog writes.
- **Still out:** PDF OCR / Vision on page rasters; Live Text / VisionKit; Foundation Models (`macOS 26+`, raise-the-floor). Pipeline if we ever did it: string → suggestions the researcher accepts → real Citation + Observations. Experiment: “obituary → draft cards,” not “obituary → catalog writes.”
- **Spike 10 took the two cheap slices:** date prefill from the transcription, and splitting a full name line into parts. The rest stays here.
