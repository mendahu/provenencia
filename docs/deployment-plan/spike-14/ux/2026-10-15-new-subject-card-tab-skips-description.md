# Tab in new subject card label does not reach description

- **Date:** 2026-10-15
- **Type:** Bug (keyboard navigation)
- **Where:** Event editor / new subject card
- **Annoyance:** Adding a new subject card for an event, pressing Tab in the label field does not move focus to the description field.
- **Wanted:** Tab moves focus from label to description, and Shift-Tab goes back.

## Code notes (observed, not fixed)

- `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` — `createForm` is the new-subject card form: a `PVInput` for the label and a `PVTextArea` for the description, hosted in a `PVFormDialog`. Neither field is given a shared `FocusState` / `PVInputFocus`, so the Tab order relies entirely on the AppKit key view loop.
- `macos/App/Features/CitationComposer/CitationComposerView.swift` — the "New subject" dialog uses the same `PVInput` + `PVTextArea` pair; worth checking whether it shows the same behaviour.
- `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift` — `PVTextAreaTabRouter` only handles Tab / Shift-Tab *leaving* the text area (`selectNextKeyView` / `selectPreviousKeyView`). Nothing makes the underlying SwiftUI `TextEditor` a reachable next key view *from* the preceding `TextField`, so Tab out of the label likely goes to the dialog buttons or nowhere.
- `macos/App/DesignSystem/Components/Input/PVInput.swift` — single-line `TextField`; supports an optional shared `PVInputFocus`, which would be the natural place to drive label → description focus explicitly if the key loop cannot be fixed.
- `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift` — dialog host; check whether it alters the key view loop or focus on present.
