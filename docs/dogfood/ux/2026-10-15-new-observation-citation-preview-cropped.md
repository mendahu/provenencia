# Citation preview in the new observation modal is cropped

- **Date:** 2026-10-15
- **Type:** Bug (layout)
- **Where:** New observation modal / citation preview
- **Annoyance:** The citation preview under the citation picker is cropped after roughly eight lines. The rest of the transcription is cut off and there is no way to see it from the modal.
- **Wanted:** Either let the preview expand to fit the content, or give it a scroll bar so the full citation is readable while the observation is filled in.

## Code notes (observed, not fixed)

- `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` — `citationReading` renders `model.transcription` as a `Text` with a hard `.lineLimit(8)` and `.fixedSize(horizontal: false, vertical: true)`, inside a sunken bordered box (`citationComposer.modal.citationReading`). The eight-line cap is the crop; there is no `ScrollView`, truncation indicator, or expand affordance.
- Same file — `citationReading` sits inside the `PVField(label: citationEyebrow)` block directly under the `PVSelect` citation picker, so any fix (scroll, "show more", or a taller limit) lands there.
- `macos/App/Features/CitationComposer/CitationComposerView.swift` — the modal is a `.pvFormDialog(width: 520, …)` hosting `CitationComposerAuthoringBox`. Letting the preview grow unbounded would push the whole dialog taller, so a bounded scroll region is probably the safer shape.
- `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift` — `PVFormDialogContent` wraps the form in a `PVPanel` with no internal scroll; worth checking how the panel behaves when the form outgrows the window before choosing expand vs scroll.
