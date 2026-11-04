# Citation composer follow-ups

**Status:** idea / bugs, not scheduled. Dogfood 2026-10-15. Field hints from transcription are under [`evidence-graph-structure-shortcuts.md`](evidence-graph-structure-shortcuts.md). Depiction page-change confirm is under [`depictions-and-likenesses.md`](depictions-and-likenesses.md).

## Bridge mode subject selectors — primary subjects only

While creating a bridge observation, the two subject selectors list existing bridge subjects (relationships, participations, locations, place links) alongside primary subjects. A bridge is between two primary subjects, so picking a bridge there can only end in a refusal.

**Wanted:** In bridge mode the subject selectors offer primary subjects only; bridge subjects are excluded (or shown disabled with a reason).

**Shape:**

- `CitationComposerVocabulary.make` builds one `graphSubjects` list from `snapshot.subjects` **and** `snapshot.bridges`, then one `subjectComboOptions` menu for every observation row — that design leaks bridges into bridge mode.
- `CitationComposerModel.subjectOptions` returns that single list; there is no mode-specific variant.
- Bridge end pickers and the property-mode subject picker both pass `model.subjectOptions`. A fix is a filtered list for the bridge pickers: `graphSubjects` whose `typeKey` is not an `EvidenceBridgeKind` raw value. Whether "New \<type\>…" rows belong in bridge mode is a separate product call.
- `resolvePair` already refuses a bridge on either end (`refusedPairTitle` / `refusedPairBody`); keep that as a backstop after filtering the menu.
- Tests: a "bridge mode options exclude bridges" case beside the existing citation composer vocabulary / authoring tests.

## Citation preview in the new observation modal is cropped

The citation preview under the citation picker is cropped after roughly eight lines. The rest of the transcription is cut off and there is no way to see it from the modal.

**Wanted:** A bounded scroll region (or expand / "show more") so the full citation is readable while the observation is filled in. Letting the preview grow unbounded would push the whole dialog taller.

**Shape:**

- `CitationComposerAuthoringBox.citationReading` renders `model.transcription` as `Text` with `.lineLimit(8)` and `.fixedSize(horizontal: false, vertical: true)` inside a sunken bordered box — the eight-line cap is the crop.
- The block sits under the citation `PVSelect` in a `.pvFormDialog(width: 520, …)`.
- `PVFormDialogContent` wraps the form in a `PVPanel` with no internal scroll; check panel behavior when the form outgrows the window before choosing expand vs scroll.
