# Credibility argument field truncates and ignores Enter

- **Date:** 2026-10-18
- **Type:** Bug (form UX)
- **Where:** Source detail page / credibility field
- **Annoyance:** (1) The credibility argument is a single-line field, so a longer note scrolls off and gets visually truncated instead of wrapping. (2) While editing it, pressing Enter does nothing; the only way to commit is the Save button.
- **Wanted:** (1) The field grows or wraps (bounded, with a scroll past that) so a longer credibility argument is fully visible and editable, like the Source description and notes editors. (2) Enter commits the edit like a form submission (Save when dirty), consistent with the other Source page inline fields.

## Code notes (observed, not fixed)

- `macos/App/Features/Sources/SourcePageCredibilityView.swift` — the argument is a `PVInput(text: $argumentLocal, size: .sm, prompt: …)`, i.e. a single-line `TextField`. That is the truncation: there is no `lineLimit`, wrap, or growth. The sibling Source sections use `PVTextArea` for prose — `SourcePageDescriptionView.swift` (`lineLimit: 3...12`, `typography: .bodyLarge`) and `SourcePageNotesView.swift` — so swapping to `PVTextArea(text:lineLimit:prompt:)` is the in-kit fix.
- Same file — save is only wired through `PVInlineEditActions.onSave` (sets `model.credibility.argumentDraft = argumentLocal`, then `save()`). There is no `.onSubmit` on the input, which is why Enter is inert. Other Source page fields do attach one: `SourcePageMetadataView.swift` (`.onSubmit { onSave(draft) }`), `SourcePageArtifactsView.swift` (`.onSubmit { commitEdit(art) }`), `SourcePageIdentityHeader.swift`. An `.onSubmit` that runs the same closure when `isDirty` would match them. If the field becomes a `PVTextArea`, Enter inserts a newline by default, so Enter-to-save vs. Shift/Option-Enter-for-newline needs a deliberate choice (see `PVTextAreaTabRouter` in `DesignSystem/Components/TextArea/PVTextArea.swift` for the existing key-routing hook).
- `macos/App/Features/Sources/SourceCredibilitySection.swift` — the model: `argumentDraft`, `isDirty`, `save()`. No change implied beyond whatever the view calls.
- `macos/App/DesignSystem/Components/InlineEdit/PVInlineEdit.swift` — `PVInlineEditActions`; worth checking whether the recipe should own a default-action keyboard shortcut for Save so every inline editor gets Enter for free rather than each view adding `.onSubmit`.
