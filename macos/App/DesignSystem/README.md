# Provenencia design system (macOS)

## What this is

Provenencia's brand is "modern but archival": warm parchment neutrals, serif
type throughout, hairline rules, and a palette of seven historical pigment
hues (iron gall, madder, verdigris, ochre, lapis, plum, copper) doing real
informational work — evidence grades, record types, lineage lines. See
`tokens/colors.css`'s comment block (referenced below) for the full
rationale; this file only covers how that system is expressed in Swift.

## Provenance

The design system was authored in Claude Design as project **"Genealogy app
onboarding flow"** (`b51790c9-7f65-4e91-a7ef-a900095c5870`), design-system
bundle `provenencia-design-system-0f6c1f29-8408-4735-b114-5ae4c4c18445`. That
project represents the system as CSS custom properties and a React
component library — that's how Claude Design renders an interactive browser
preview for review, not something that ships. `tokens/*.css` in that project
is the source of truth for every value in `Tokens/`; there's no live sync
between it and this folder, so **when a token changes on the design-system
side, re-transcribe it here by hand.**

The project also contains one mockup board, `Onboarding Flow.dc.html`,
showing the 8 window states of this app's actual onboarding flow. Per that
project's own sync log, it was built by reading the shipped onboarding
SwiftUI files. Those screens have been restyled to consume `PV*` tokens and
components directly.

## Token file map

| CSS file | Swift file | Holds |
|---|---|---|
| `tokens/colors.css` | `Tokens/PVColor.swift` | Palette ramps (`PVPalette`) + semantic aliases (`PVColor`), light/dark aware |
| `tokens/fonts.css` + `tokens/typography.css` | `Tokens/PVTypography.swift` | Font families, type scale, weights, tracking, line-heights, `PVFont` factories |
| `tokens/spacing.css` | `Tokens/PVSpacing.swift` | Space scale, gutters, measures, layout widths, control heights |
| `tokens/radii.css` | `Tokens/PVRadius.swift` | Corner radii |
| `tokens/elevation.css` | `Tokens/PVElevation.swift` | Shadows, focus ring, `.pvShadow`/`.pvInsetShadow`/`.pvFocusRing` modifiers |
| `tokens/motion.css` | `Tokens/PVMotion.swift` | Durations, easing curves, press-scale, lift-hover |
| — | `Tokens/PVStroke.swift` | The 1 pt dashed outline (`dashed`, `dashPattern`), used by `PVCard`, `PVEmptyState`, `PVChip`, the Evidence card and the Sources list blocked graph buttons |

Every value in these files is transcribed from the CSS, not invented. One exception: `PVSpacing.widthWideBreakpoint` was set in Swift first; S14-D1 WL-1 adds `--width-wide-breakpoint` to the CSS, and S14-25 checks the two agree.
Reach for `PVColor`/`PVFont`/`PVSpacing`/`PVRadius`/`PVElevation`/`PVMotion`/`PVStroke`
at call sites — never a literal color, font, or number.

## The component pattern

**Layering (authoritative):** **Components / Recipes / Snowflakes** with **one
folder per control** — see [`docs/design-system-layers.md`](../../../docs/design-system-layers.md).
Classify before adding ([`add-ui-component`](../../../.agents/skills/add-ui-component/SKILL.md));
audit with [`evaluate-ui-component`](../../../.agents/skills/evaluate-ui-component/SKILL.md).
Do not fork a `PV*` for a slight variant; compose down. **No UI-category
subfolders** (`Core`, `Forms`, `Feedback`, `Research`, …). **No loose `.swift`
at a layer root.**

```text
DesignSystem/Components/<Name>/   # e.g. Button/PVButton.swift + optional docs/helpers
DesignSystem/Recipes/<Name>/      # e.g. Marks/PVMark.swift
DesignSystem/Snowflakes/<Name>/   # named kit-side one-offs (rare; prefer Features/)
Features/<Feature>/               # typical snowflake home
```

`Components/Button/PVButton.swift` is the **canonical example** — read its
header comment before adding a new component. In short:

- A **component** lives at `DesignSystem/Components/<Name>/PV<Name>.swift`
  (folder PascalCase without `PV`; **type** uses the `PV` prefix — see
  [`docs/design-system-layers.md`](../../../docs/design-system-layers.md) § PV
  type prefix: disambiguate from SwiftUI/AppKit and mark kit ownership). A
  **recipe** at `DesignSystem/Recipes/<Name>/…` likewise uses a `PV*` public
  type. Colocate helpers and component docs in that folder.
- The file's header comment names the `.jsx` it mirrors (when ported from the
  Claude Design kit) and calls out any deliberate deviation.
- The type only ever reaches for `PV*` tokens — never a literal color,
  font, or size.
- User-facing text is a `LocalizedStringResource`, never a raw `String`
  (`docs/macos-client-patterns.md` §6); a select/list of *data* (a project
  name, a person's name) stays a plain `String` — it isn't translatable UI
  copy.
- `.accessibilityIdentifier` is left to the call site
  (`docs/macos-client-patterns.md` §5) — components don't invent their own.
- Hover/pressed/focus state lives in a private backing `View` or
  `ButtonStyle`, not on the public type (see `PVButtonBody` in
  `PVButton.swift`).
- Each component file ends with a `#Preview` using static sample data (no
  live `GenealogyStore` needed).

To add another component: classify the layer, create `…/<Name>/`, drop the
primary Swift file (and any colocated docs/helpers) there, and read its CSS/JS
spec from the source design-system project first when porting — don't guess at
colors/spacing/sizes.

**Check `swift/` in the source project before porting anything.** That
project ships a Swift reference implementation (`ProvenenciaTokens.swift`,
`ProvenenciaCore.swift`, …) alongside the CSS/JSX, and its own `SKILL.md`
says to use those APIs "rather than translating CSS by hand." Several
components here were re-derived from the CSS instead, and at least one of
them lost a detail the Swift reference had already got right — see below.

## Interaction state

**Never branch view structure on focus, hover, pressed, or selected.**
Those states change *values* — color, opacity, offset — on a view that is
**always present**:

```swift
// WRONG — swaps view identity when focus flips
if isFocused { content.pvFocusRing(…) } else { content.pvInsetShadow(…) }

// RIGHT — one view, always applied; only the opacity changes
content
    .pvInsetShadow(cornerRadius: PVRadius.sm, visible: !isFocused)
    .pvFocusRing(isFocused, cornerRadius: PVRadius.sm)
```

An `if/else` in a `ViewBuilder` compiles to `_ConditionalContent` — two
different view types — so SwiftUI rebuilds that subtree whenever the
condition flips. Around a text field that tears down the underlying
`NSTextField` *at the instant focus arrives*, destroying the first
responder it just gained: the field won't accept a click, or takes exactly
one keystroke and then every key beeps. Nothing catches this — the code
reads correctly, there are no UI tests, and `#Preview` doesn't exercise
focus. `PVInput` shipped with this bug for exactly that reason.

The same applies to `.animation(_:value:)` keyed on an interaction flag:
scope it to the decorative shape, never to a wrapper containing the
control. `pvFocusRing(_:cornerRadius:)` and `pvInsetShadow(cornerRadius:
visible:)` both take the state as a parameter so the call site never needs
a branch — that mirrors the design system's own `pvFocusRing(_:radius:)`.

This generalizes: the web kit's CSS is full of ternaries like
`boxShadow: focus ? 'var(--ring-focus)' : 'var(--shadow-inset)'`. In CSS
that's a *property value* swap and costs nothing. Port it as a value swap,
not as a conditional view. `Select`, `Checkbox`, and `Switch` all have the
same shape waiting in their specs.


### Host-owned pointer

Paint-only surfaces own their pointer: see the item contract in the [GraphCanvas README](../Features/GraphCanvas/README.md) and `GraphCanvasItemView`.

## What's built vs. not yet

The components the **Onboarding Flow** board actually uses, plus `Toast`
(added when onboarding needed a way to surface errors that wasn't an inline
red `Text`), plus `Badge`/`EmptyState`/`Callout` (added for the S2-02
**Metadata** board, then called Source fields — see `Features/Metadata/`):

| Component | File | 
|---|---|
| Button | `Components/Button/PVButton.swift` (icon-left, loading spinner, and a chrome-less `link` variant added for S2-02; the label is `PVCopy`, so a formatted `String` such as Promote's "Back to Compare" works as well as a catalog resource, and an `AttributedString` from `StyledFormat.attributed` carries its own styling; `flushLeading` pulls the button left by its own horizontal padding so a ghost button's label starts on its column edge) |
| Copy | `Components/Copy/PVCopy.swift` (`PVCopy`: the copy a component shows or speaks, a `LocalizedStringResource`, an already formatted `String` or an `AttributedString`, drawn through `pvText`; kit titles and labels take it) |
| Disclosure | `Components/Disclosure/PVDisclosureButton.swift` (added for S9-16's Person page: a `PVButton` that owns an `isExpanded` binding, its chevron, and the spoken state; the caller places the disclosed content. `flushLeading` is forwarded to the button; `pvExpandedState(_:)` is the shared VoiceOver piece, also used by `PVSelect` and the Source page's artifact rows. No JSX twin: boards compose `Button.jsx` with `aria-expanded`) |
| Icon | `Components/Icon/PVIcon.swift` |
| Painter | `Components/Painter/PVPainter.swift`, `PVTextMetrics.swift` (AppKit drawing for views that paint themselves, such as a canvas item: token colors resolved per appearance through `PVPainter.color(_:in:)`, continuous rounded shapes, text, symbols and marks; `PVTextMetrics` sizes text by the rule it draws with. `PVSymbolImages` in `Components/Icon/` caches the configured SF Symbol images) |
| Field | `Components/Field/PVField.swift` (`marker: .required / .optional` marks the label) |
| Input | `Components/Input/PVInput.swift` |
| TextArea | `Components/TextArea/PVTextArea.swift` (multiline field with `PVInput`'s chrome; grows with its text between `lineLimit`'s bounds, or with `fillsHeight` takes the height its container offers, never less than the lower bound, and scrolls inside it; added for inline edit, and for the Promote sheet's argument box beside a column of fields) |
| InlineEdit | `Components/InlineEdit/PVInlineEdit.swift` (display and pencil → editor and Save / Cancel; `PVInlineEditActions` is the Save / Cancel row for always-editable surfaces; accessibility ids stay at the call site unless `accessibilityIdentifierPrefix` is set) |
| SectionHeader | `Components/SectionHeader/PVSectionHeader.swift` (detail-page section chrome: display title, mono meta, aside, trailing actions and a bottom rule; the title is a VoiceOver heading; a `String` or localized title is semibold; an `AttributedString` title draws regular and its strongly emphasized runs carry the weight) |
| Select | `Components/Select/PVSelect.swift` (native popup contract on `PVContextMenu`; see "The Select contract" below) |
| ContextMenu | `Components/ContextMenu/PVContextMenu.swift`, `PVActionMenu.swift` (action menus are `PVActionMenuItem` lists: a right-click menu that opens at the pointer is one `.pvRightClickMenu`, which adds the state, the trigger and one VoiceOver action per enabled item (the toolbar's Back/Forward history menu anchors to its button and keeps its own presenter); a button-opened menu draws `PVActionMenuRows` inside `.pvContextMenu`, with `keyboardTitles` from `PVActionMenuItem.keyboardTitles`; `PVSelect` uses `.pvContextMenu` directly) |
| FloatingMenuSelection | `Components/FloatingMenuSelection/PVFloatingMenuSelection.swift` (clamped ↑/↓ highlight movement with no wrap for lists and floating menus; `PVTypeSelectMatcher` is the type-to-select buffer; do not write another highlight mover) |
| Toast | `Components/Toast/PVToast.swift` (announces its title and message to VoiceOver when it appears) |
| LogoMark | `Components/LogoMark/PVLogoMark.swift` |
| SidebarNav | `Components/SidebarNav/PVSidebarNav.swift` (added for the S2-01 workspace chrome; ports that board's revised `collapsed`-capable `SidebarNav.jsx`); an item's icon is a `PVGlyph`, a symbol or a mark (S10-19) |
| IconButton | `Components/IconButton/PVIconButton.swift` (added for the workspace sidebar's collapse toggle, which needed real hover feedback; `label` is required per `IconButton.jsx` and doubles as the `.help` tooltip; pass `accessibilityLabel` when the spoken label has to name a target the tooltip can leave implicit; `tone: .danger` tints a destructive action; `isSelected` is the armed/radio fill used by Artifact region tools) |
| ClipboardButton | `Components/ClipboardButton/PVClipboardButton.swift` (a `PVIconButton` that copies `text` to the pasteboard, shows a green check for 1.6s, then the copy glyph again, and announces `copiedAnnouncement` to VoiceOver; `label` names what is copied and is the tooltip; the pasteboard write and the timer live in `PVClipboardCopyState` so a test can inject both; added for the provenance sheet's citation line) |
| HoverEffect | `Components/HoverEffect/PVHoverEffect.swift` (the interaction chrome inside a `ButtonStyle` body: hover tracking, disabled opacity, reduce-motion-aware press scale and one hit shape; callers draw from `showHover`) |
| Badge | `Components/Badge/PVBadge.swift` (added for S2-02's data-type/origin badges; a glyph-only variant carries the S2-22 seeded pill) |
| Chip | `Components/Chip/PVChip.swift` (a meaning-free selectable pill for exclusive options, and `PVChipGroup` to lay a set out; sizes from `PVControlSize` (`size:`, default `.md`), an optional leading `glyph:`, keyboard focus drawn through its button style, and the selected trait; sibling to the display-only `PVBadge`; domain vocabulary stays in the feature that composes it) |
| SegmentedControl | `Components/Chip/PVSegmentedControl.swift` (a few exclusive options in one sunken track: `PVSegmentedControl(label, items:, selection:, size:, select:)`; items carry a title, glyph, ⌘ shortcut, disabled flag and accessibility identifier; the group reads its label to VoiceOver; use it instead of hand-building `PVChipGroup(style: .segmented)`) |
| Divider | `Components/Divider/PVDivider.swift` (1pt hairline; horizontal/vertical) |
| EmptyState | `Components/EmptyState/PVEmptyState.swift` (added for S2-02's empty/no-match states; the web spec's `action` slot isn't ported — see the file's header comment); the icon slot takes a `PVSymbol` or a `PVGlyph` (S10-19) |
| Callout | `Components/Callout/PVCallout.swift` (tone / icon / title / body / compact; the title is `PVCopy`; optional `@ViewBuilder` actions under the body — call-site `PVButton`s; `announces: true` speaks the title and message when it appears or changes (error callouts); `onDismiss` / `detail` / `plain` deferred) |
| ErrorText | `Components/ErrorText/PVErrorText.swift` (a validation or failure line in the danger color; `size:` `.micro` for a field-level line, `.caption` for a form-level line; call sites keep their own padding and identifiers; `announces: true` speaks the line when it appears or changes) |
| Announce | `Components/Announce/PVAnnounce.swift` *(new)* (`pvAnnounce(_:priority:)` speaks a text when it appears or changes; toasts, callouts and error lines use it) |
| Table | `Components/Table/PVTable.swift` (added for S2-22, extracted from the Metadata (then Source fields) list; see "The table tradeoff" below) |
| Confirm | `Components/Confirm/PVConfirm.swift` (added for S2-22's delete confirmation; the macOS answer to `ConfirmDialog.jsx`, which the web spec says not to port — see "Confirmations are system chrome" below) |
| Panel | `Components/Panel/PVPanel.swift` (sheet content shell: title / subtitle / body / optional footer; optional title accessory, drawn after the title, filling the row; put a Spacer in it to push a control right; no window chrome); `Components/Panel/PVSheet.swift` (`.pvSheet`, the one sheet presenter: Escape, outside click and the discard question) |
| FormDialog | `Components/FormDialog/PVFormDialog.swift` (short create/edit form sheet on `PVPanel`; optional `width`, default 480; `pvFormDialog(item:)` for model-backed forms — see note below) |
| ComboBox | `Components/ComboBox/PVComboBox.swift` (added for S2-16's assign-field control, where the pool is the whole Metadata vocabulary; single-select subset only — see "The combo box subset" below). The box is content-agnostic: rows (`row`) and the empty line (`empty`, given the typed query) are caller-built views; the kit ships only the generic `PVComboBoxPlainRow` (label + subtext) and a plain-text empty line, and never grows domain row kinds (the web kit's `person` row lives in features as their own `row`). S9-11 added **remote results**: `onQueryChange` hands each typed query to the caller, which passes back the search results as `options`, keeping the committed option among them, and `isLoading` shows the searching line while that search runs |
| Radio | `Components/Radio/PVRadio.swift` (`PVRadio` label + description, and the bare `PVRadioMark` for rows that are themselves the control; added for S9-11's Promote target choice) |
| Checkbox | `Components/Checkbox/PVCheckbox.swift` (`PVCheckbox` label + description, and the bare `PVCheckboxMark`; `isLocked` is the lock mark in that same box for a link or binding the researcher cannot clear, spoken "Locked"; `isIndeterminate` is spoken "partly checked") |
| Switch | `Components/Switch/PVSwitch.swift` (32×18 track; immediate on/off — a saved form still uses Checkbox) |
| List | `Components/List/PVList.swift` (added for S9-09, ported from the kit's `PVList`; its internal `PVListRowStyle(focused: nil)` is the row chrome the Sources list's Source and Type button also wears: navigation rows — thumbnail slot (a view; S10-35), title, secondary line, optional fixed-width detail column (icon and one line of text, with help) before the trailing mono ref, chevron — as buttons, not selection; one focus stop with ↑/↓ · Home/End · Page Up/Down · Return/Space; `PVListSkeleton` for first load. The counterpart to Table) |
| ScrollRowIntoView | `Components/ScrollRowIntoView/PVScrollRowIntoView.swift` (`.pvScrollsRowIntoView`: keeps the highlighted or selected row in view as the keyboard moves it, can send a new list back to its first row, and restarts a page at its top when a key changes (`restartsAt:whenChanged:`); `PVTable`, `PVList`, `PVSelect`, `PVComboBox` and the omnibar results scroll through it — do not add another `ScrollViewReader` for this) |
| SkeletonBar | `Components/SkeletonBar/PVSkeletonBar.swift` (added for S12-02: a rounded `borderSubtle` placeholder bar, fixed or fill width; used by `PVListSkeleton` and the Copy Source loading body) |
| FlowLayout | `Components/FlowLayout/PVFlowLayout.swift` (S12-14: promoted from the Citation composer; wraps children by ideal width, callers cap long children; used by the composer) |
| WidthClass | `Components/WidthClass/PVWidthClass.swift` (the Regular / Wide class of the pane, measured once by `WorkspaceDestinationHost`; read `\.pvWidthClass`) |
| PageColumns | `Components/PageColumns/PVPageColumns.swift` (the record-page scaffold: a header slot, a primary and an optional secondary column, centered in the pane; stacked at Regular, side by side at Wide; the header scrolls with the page or is pinned (`headerPinned`) and receives `isScrolled` to compact; owns the scroll and `scrollResetKey`; sizing lives in `PVPageColumnsMetrics`; reads `\.pvWidthClass`; sets no background; see [§ 9 of macos-client-patterns.md](../../../docs/macos-client-patterns.md#9-window-widths-regular-and-wide)) |
| Thumbnail | `Components/Thumbnail/PVThumbnail.swift` (added for S2-17 Sources list rows; image / mark / SF glyph / empty / loading tile); content takes an image or a `PVGlyph` (S10-49); optional `colors:`: tint and line on mark and symbol states; the line also outlines an image (S10-23, S10-35) |
| Card | `Components/Card/PVCard.swift` (surface container: tone / border / radius / optional elevation + padding; mirrors web `Card.jsx`; no header/footer slots yet — compose those outside) |
| Marks | `Recipes/Marks/PVMark.swift` (`file_*` + `type_*` + `subject_*`; see colocated `MARKS.md`; not SF Symbols); `PVGlyph` / `PVGlyphView` for a symbol-or-mark slot |
| OmnibarHitRow | `Recipes/OmnibarHitRow/PVOmnibarHitRow.swift` |
| LastUpdated | `Recipes/LastUpdated/PVLastUpdated.swift` (the last-updated cell of a list row: history icon, relative time with the full stamp as help, a faint dash when none, and the row label's spoken suffix; `PVListDetailCell` draws it) |
| DeleteImpact | `Recipes/DeleteImpact/PVDeleteImpact.swift` (S8-13: confirm when `GetDeleteImpact` is allowed, notice when blocked; composes `.pvConfirm`; `PVDeleteImpactGroupList` is its list of rows, for in-form refusal lists) |
| ArtifactChecklist | `Recipes/ArtifactChecklist/PVArtifactChecklist.swift`, `PVChecklistRows.swift` (S12-12: a Source's Artifacts as a checklist with an All row, thumbnails, count and an optional caption; `locked` rows show the lock mark with a hint and survive the All row; `PVChecklistState` is the pure arithmetic of every id checklist; `PVChecklistFrame` is the bordered, capped, divided row list and `PVChecklistCheckbox` the row checkbox with its toggle and locked spoken value, both also composed by the artifact graph group in `SourceArtifactGraphLinks`; used by the Copy Source and Evidence graph forms) |
| Breadcrumbs | `Components/Breadcrumbs/PVBreadcrumbs.swift` (added for S2-18 Source page trail; Sources → `SRC-…`) |
| ReorderableList | `Components/ReorderableList/PVReorderableList.swift` (+ `PVReorderHandle`; added for S2-25 Metadata drag order) |

The other design-system components have **no files yet** — add them on
demand, following the pattern above, when a screen needs one:

| Component | Category | Purpose |
|---|---|---|
| Tag | Core | Removable/interactive pill with a color dot |
| Tooltip | Core | Hover label — on macOS this is usually SwiftUI's own `.help()`, which is what `PVIconButton` uses; port the web hover card only if a call site needs richer content |
| Tabs | Navigation | Tab strip |
| EvidenceBadge | Research | The 5-grade confidence marker (proven/probable/possible/disputed/undocumented) |
| FactRow | Research | One asserted fact: type glyph, date, value, place, grade, conflict note |
| PersonChip | Research | A person with life dates and a lineage-colored rule |
| SourceCitation | Research | Citation + repository + scan thumbnail + grade, as one unit |

**Form dialog on macOS:** the web `Dialog.jsx` draws a scrim because the browser
gives it none. Here `PVFormDialog` presents through `.pvSheet` on ``PVPanel`` — same
family as `PVConfirm`'s rich sheet — and leaves window chrome to the system. Prefer
`PVConfirm` for destructive confirmations; use `PVFormDialog` for short
create/edit forms.
A form backed by a model its host sets presents through `pvFormDialog(item:)`, which
owns the presented binding (cleared by Escape, Cancel or an outside click, never
while the save runs) and clears it when `onConfirm` answers true.

**Every sheet presents through `.pvSheet`**
(`Components/Panel/PVSheet.swift`, `item:` and `isPresented:` forms), and
`pvFormDialog`, `pvConfirm` and the delete-impact notice are built on it. It owns
the two ways a researcher closes a sheet without a button: **Escape** (a hidden
cancel-action button) and **a click outside the sheet**
(`.pvOutsideClick(isEnabled:perform:)`, `Components/Panel/PVOutsideClick.swift`; a
macOS sheet gives that click to no one, so the modifier watches the event
stream; menus and popovers are not this path). Both call one rule,
`PVSheetDismiss.action(isRunning:isDirty:isConfirmingDiscard:)`: running work is
ignored, a dirty sheet shows the discard question (and Escape on that question
returns to the form), a clean sheet closes. The visible Cancel button closes at
once and carries no key of its own. A new sheet gets all of this by presenting
with `.pvSheet`; do not add a raw `.sheet`, `.pvOutsideClick` or
`.keyboardShortcut(.cancelAction)` for a sheet.

## The combo box subset

`PVComboBox` ports `components/forms/ComboBox.jsx`, but only the
**single-select** half of it. The token/multi-select field, the "use what you
typed" create row and the `person` row kind are all in the web component and
none of them have a call site here yet — add them the way `PVSidebarNav` was
added, rather than porting speculatively. Async loading is ported for remote
results (`isLoading`): with nothing in yet the list shows a "Searching…" line
in place of the empty line, and options already in stay until new ones
replace them, as the omnibar keeps its hits. A failed search is the caller's
empty line, and the general init takes `empty` as a view given the typed query,
so a hint or a failure's message can be that line. When the box resets the
field itself (a click away, the second Escape) it reports the restored text
through `onQueryChange` too, so the caller's query and results follow the
field. The popup window carries the field's environment, so a row sees the
same session and `isEnabled` the field does.

What is ported is the interaction contract, and it is worth restating because
every clause of it is a decision rather than an accident:

- **Seeded text is not typed text.** The field shows the committed option's
  label, so re-opening it must show the whole list, not just the row already
  chosen. An `isDirty` flag separates the two, and only the text binding's
  *setter* raises it — code that assigns the query (mirroring a label,
  reverting on Escape, restoring on blur) leaves it alone.
- **Spotlight rule.** With a query, the top hit is pre-highlighted so Return
  takes it. With no query nothing is highlighted: it is a browse list, and
  Return falls through to whatever is behind it.
- **Arrow keys clamp, they do not cycle** — `NSMenu` does not wrap, and
  neither does this. `PVTable`'s list *does* wrap, which is right for a browse
  table and wrong for a menu; the two are deliberately different. The
  jump-to-end modifier is **not** yet consistent between them, though:
  `ComboBox.jsx` specifies `⌘↑/⌘↓` and this port follows it, while `PVTable`
  shipped `⌥↑/⌥↓`. Both also accept Home/End. Worth settling on one before a
  third list component copies whichever it meets first.
- **One highlight, two drivers.** The pointer and the arrow keys move the same
  `active` row, so there is no second hover state competing with it.
- **Escape is two-stage** — close the list, then abandon the typed text — and
  once there is nothing left to undo it stops consuming the key, so an
  enclosing sheet can still be dismissed from inside the field.
- **Clicking away blurs.** A click outside the field and its list doesn't
  just close the list, it resigns the field's focus — outside means done.
  Clicking the field itself opens (or reopens) the list; focus alone does
  not, so a sheet that auto-focuses the field on present won't pop the
  menu. Typing or pressing an arrow / Home / End / Page key also opens it.

**The list is a child window, not an overlay.** This is the part that took two
passes to get right. A SwiftUI overlay is clipped by any enclosing
`ScrollView` and by the app window, so for a field near the bottom of a
scrolling inspector — which is exactly where the first call site puts it — the
list was cut off at the window edge and had nowhere legal to go. Flipping it
upward only moves the problem.

`NSMenu`, `NSComboBox` and Spotlight all draw into their own borderless window
for this reason, and so does this. `PVComboBoxPlacement.popupFrame` positions
a non-activating `NSPanel` in **screen** coordinates against the field, which
buys three things at once: it clips to the display rather than to any view, it
may extend past the app window's edge, and it is placed strictly *outside* the
anchor so it can never cover what is being typed. Height is trimmed to the
room actually available on whichever side it lands, and the list scrolls
inside that. The panel returns `false` from `canBecomeKey`, so the text field
keeps focus and keeps handling arrows, Return and Escape while it is open.

Two things the window costs us, both deliberate:

- **Sizing is the caller's job.** The window has to know its height *before*
  it is shown, and neither a `ScrollView` inside it nor
  `NSHostingView.fittingSize` can be trusted to say. While the menu is open,
  `PVComboBox` lays a hidden copy of the rows out behind the field purely to
  measure them, waits for that height, then hands it to the window. A closed
  field does not build the list. The scroller simply fills whatever height it
  gets. An earlier version made one view both report an ideal height *and*
  lay out inside the resulting window; any disagreement between the two clips
  a row, and it did.
- **Dismissal converges on focus.** SwiftUI does not resign a `TextField`'s
  focus when you click inert content, so "clicked away" is observed at the
  event level: a local mouse-down monitor treats anything that is neither in
  the popup nor on the field as a click away, and — rather than merely hiding
  the list — resigns the field's focus. Losing focus is the one close path
  everything funnels into, whether the click moved first responder by itself
  or the monitor had to do it. The panel itself never decides to be visible:
  the coordinator mirrors a `wantsPresented` flag that SwiftUI state (and the
  monitor's dismissal) writes synchronously, so a deferred show can never
  race a dismissal and re-open the list. The coordinator also follows the
  field on window move/resize and on scroll, and dismisses when the parent
  window resigns key — which is also what covers a click in another app, so
  there is no global event monitor.
- **Clicks on rows are ordinary tap gestures.** The popup never becomes key,
  but mouse events are still delivered to the window under the pointer, so
  the row's `.onTapGesture` is the one commit path; the mouse monitor passes
  panel clicks through untouched. What *had* made gestures inside the popup
  undependable was re-framing: an early version had the active-row index in
  the token that triggers re-measurement, so every mouse move resized the
  window and reassigned the hosting view's root, which tore down click
  handling mid-gesture and made rows unclickable. Hence the rule that
  **nothing that changes on mouse move may re-frame the window** — the
  coordinator re-frames only when the row count or measured height actually
  changed.
- **The shadow is AppKit's, not `--shadow-overlay`.** A transparent window
  casts its shadow from its own alpha, so the rounded list gets the native
  popup shadow for free and correctly shaped. Transcribing the CSS shadow
  would mean padding the window out to make room for it; the native shadow is
  the more Mac-like answer anyway.

The pure parts — filtering, movement, placement, match highlighting — live in
`PVComboBoxMatch` / `PVComboBoxPlacement` / `PVComboBoxHighlight` so they can
be unit-tested without a view, the same split `PVTable` uses for its
type-select matcher. `popupFrame` in particular is worth the test: the
"never overlaps the field" and "never escapes the screen" properties are
swept across anchor positions rather than spot-checked.

## The Select contract

`PVSelect` is custom Frost chrome on `PVContextMenu` that implements a
**native macOS popup-button contract**, minus press-drag-release. It is
**not** SwiftUI `Picker` / `Menu` and **not** `NSPopUpButton`.
`PVSelectSession` owns open / commit / keys. `PVSelectPlacement` owns
where the panel sits. The menu is a normal click-to-open list: click a
row, or scroll and then click.

**Interaction contract:**

| Input | Closed (focused trigger) | Open (menu showing) |
|---|---|---|
| Click the field | Open | Close without commit (same as Escape) |
| Click a row | — | Commit that row and close |
| Scroll the list | — | Reveal more rows; click to commit |
| Space / Return | Open | Commit the highlighted row |
| ↑ / ↓ | Commit next/prev. Stay closed. Clamp — do **not** wrap. | Move highlight. Clamp. |
| a–z, 0–9 | Commit the match. Stay closed. | Jump highlight only |
| Home / End | Commit first/last. Stay closed. | First / last highlight |
| Escape | No-op | Dismiss; restore the snapshot taken at open |
| Click away | — | Dismiss without commit (same as Escape) |

**Menu-only options.** A `PVSelectOption` with `isMenuOnly` is chosen only
from the open menu; the closed keys (↑ / ↓, Home / End, type-select) and
VoiceOver adjust skip it.

**Selected vs highlight.** The committed row keeps a leading checkmark and
the `.isSelected` trait, with **no** fill. Keyboard highlight uses
`surfaceSelected` fill. A committed+highlighted row shows both; two rows
must not look equally selected.

**Placement.** Width is `max(trigger, menuWidth)` then clamped to the
screen. Gap 4 pt, left-aligned. Prefer below; flip above when the
preferred height does not fit; scroll when neither side fits
(`maxVisibleRows` × row height, default 10). The closed field is
`PVInputChrome`. A full-width field takes the width it is offered. A
fixed field hugs its label, and truncates when that offer is tighter.
Help shows the full closed label, unless `closedHelp` is set: then help is
that string (the Evidence graph's file select helps with the filename).

**Empty / disabled.** Empty options cannot open. Disabled is opacity 0.45
and ignores keys and pointer.

**Accessibility.** The trigger is the VoiceOver surface (`aria-activedescendant`
twin). Spoken value is the committed label when closed and the highlighted
option when open. Expanded / collapsed is custom content, not stuffed into
the value (`pvExpandedState`, shared with `PVDisclosureButton`). Increment / decrement call the same session moves as ↑ / ↓
(closed commit; open highlight). Open menus expose `N of M`. One tab stop
on the trigger; the panel is not a second stop. Icon-only chips require a
spoken label (`tableFilterColumn`). Decorative chevron, chip icon, and
checkmark stay `accessibilityHidden`.

**Honest accessibility limit.** A SwiftUI `Button` cannot claim AppKit's
`NSAccessibilityPopUpButtonRole`. We will **not** wrap `NSPopUpButton` or
hide a `Picker` as `accessibilityRepresentation` to fake that role.
VoiceOver will likely say **button** (or “adjustable”) rather than
“pop-up button”. Same honesty bar as `PVTable` not claiming a table grid.

Not `PVComboBox` (filter-as-you-type), not `PVContextMenu` action menus
(`PVActionMenuItem` lists, button-opened or `.pvRightClickMenu`), and not `PVSegmentedControl` (a few exclusive options chosen inline).
Date-value calendar / month and `PVTable.filterMenu` remount on
`PVSelect`. Do not add a second popup-select kit.

## The table tradeoff

`PVTable` is custom chrome on purpose. SwiftUI `Table` and AppKit
`NSTableView` bring their own header, row and selection styling, and none of
it can be pushed all the way to the design system's look: micro-caps headers,
a 2pt accent bar on the selected row, hairline `borderSubtle` rules, badge
cells, the warm hover tint. Visual fidelity won; the cost is that keyboard
and screen-reader parity had to be built by hand.

**Interaction contract** (mirrors `components/data/Table.prompt.md`):

| Input | Behavior |
|---|---|
| ↑ / ↓ | Move selection through the visible rows. Clamps at both ends — does **not** wrap. |
| ⌥↑ / Home | Select the first row. |
| ⌥↓ / End | Select the last row. |
| Page up / Page down | Move ten rows, clamped. |
| a–z, 0–9 | Type-to-select on `primaryText`. The buffer resets after 800ms; a single keystroke advances to the *next* match, so repeated presses cycle. |
| Escape | Clears the type-select buffer. |

The row list is one focus stop, so tab order reads: search field → sort
headers → table → detail pane. Selection always scrolls into view via
`.pvScrollsRowIntoView`. Like a native table, the highlight dims when the table
does not have focus — `surfaceSelected` + accent bar when focused,
`surfaceSelectedInactive` + `borderStrong` when not. Focus shows
`pvFocusRing` on the container; the view structure never changes on focus
(same rule as `PVInput`).

The caller owns the data: `rows` arrive already filtered and sorted, and the
table reports sort/filter intent through `onSortChange` /
`PVTableColumnFilter.onChange`. Column definitions are the single source of
truth for width — never restate a width at the call site. Search bars and
result footers are feature chrome and stay in the pane.

**Honest accessibility limit.** A custom view cannot claim AppKit's table
grid role, so VoiceOver will not announce "table, row 3 of 12, column 2".
Rows are `.accessibilityElement(children: .combine)` instead, so each reads
as a single element ("Author, author, text") with `.isSelected` on the
selected one, and sortable headers announce their direction via
`.accessibilityValue` because `PVIcon` is `accessibilityHidden`. That
list-like semantic is accepted for single-selection browse lists; do not
reach for `PVTable` where cell-level navigation matters.

`PVTableSelection.moveIndex` and `PVTableTypeSelectMatcher` are pure and
sit outside the view so selection movement and prefix matching are unit
tested without mounting UI (`ProvenenciaTests/PVTableTests.swift`).

Not implemented, deliberately — same scope line as the web component:
multi-select, column resize/reorder, drag-and-drop, inline editing.

## Confirmations are system chrome

`components/feedback/ConfirmDialog.jsx` is **deliberately not ported**, on its
own `prompt.md`'s instruction. The web component draws a scrim, a backdrop
blur, a corner radius and a drop shadow only because a browser gives it none
of them. On macOS all four belong to the window, and redrawing them is exactly
what makes a native dialog look off:

- macOS does **not** dim the parent window behind a sheet — the parent's
  controls simply go inactive. A dark scrim is a web/iOS idiom. The absent
  scrim is correct, not missing.
- The sheet window supplies its own corner radius, shadow and material.
  Setting `.background` / `.cornerRadius` / `.shadow` on sheet *content* is
  what yields the double-rounded, double-shadowed panel.
- Sheets are modal to their window, not the app, and slide from the titlebar —
  all free from `.sheet`.

`Components/Confirm/PVConfirm.swift` carries the web component's **copy
rules** across without its chrome. The kit confirm is a **rich sheet on
``PVPanel``** — not a system alert wrapper:

| Modifier | Use |
|---|---|
| `.pvConfirm(item:copy:tone:isRunning:onConfirm:detail:)` | Standard confirm. Title + consequence message + optional detail (`PVConfirmKeyChip`, error callout, …). Chrome belongs to the window; content and the button row are ours. Keyed to the record it names (`item:`, not an `isPresented` Bool) so copy and detail render from a snapshot and the sheet animates out still showing them, rather than blanking the instant the model clears. |

Plain SwiftUI `.alert` remains available at a call site if a future flow needs
text-only system chrome; the kit does not wrap it.

The action bar keeps `Dialog.jsx`'s footer treatment — `surfaceSunken` with a
hairline top rule — because that band is *content*, not window chrome. The
"draw none of it" rule names four things the window owns: scrim, backdrop
blur, corner radius, drop shadow. Anything inside the panel is still ours to
style. (`swift/ProvenenciaConfirm.swift` omits the band; we restore it.)

**The sheet's corner radius is not ours; its buttons are.** The panel's
rounded corners come from the sheet *window* — there is no supported API to
change them, and redrawing the panel to get square corners means giving up
`.sheet` entirely. Accept the window radius as platform chrome, the same
category as the titlebar. The action-bar buttons, though, sit *inside* the
panel in a band we already paint, so by the content rule above they take
`.buttonStyle(.pv(…))` — `PVRadius.sm` corners, DS palette — not the system
`.borderedProminent` pill. Nothing native is lost: shortcuts, focus, roles
and disabled state live on `Button`, not the style, and the destructive
tint was already `PVColor.danger` rather than the system role tint. (This
reverses an earlier decision to keep native buttons in the sheet; the
system *alert* if a call site uses SwiftUI `.alert` directly; the kit confirm
is sheet-only.) In-content shapes keep Provenencia radii as before — `PVConfirmKeyChip`
uses `PVRadius.xs`, matching the design's square treatment for citable
values. Because a custom `ButtonStyle` draws no focus indication of its
own, `PVButtonStyle` shows `pvFocusRing` when focused, so the sheet's
cancel-first focus stays visible to keyboard users.

The copy rules travel in `PVConfirmCopy`: the title is a question naming the
record ("Delete Photographer?", never "Are you sure?"), the message says what
is *and is not* lost, confirm repeats the verb ("Delete field", never "OK"),
and cancel names the safe outcome ("Keep field"). A danger confirm's default action is Cancel: Return keeps the record, and the destructive button has no Return shortcut. Focus starts on Cancel. An irreversible confirm keeps Return on its confirm button.

Two further rules from the spec: a blocked action never reaches a confirmation
— disable the control and explain why in its tooltip, the way Metadata's
delete button does. And a *reversible* action should not confirm at all: act,
then offer undo in a `PVToast`.

One deviation from the design system's `swift/ProvenenciaConfirm.swift`
reference: it clears presentation state before invoking `onConfirm`, which
closes the sheet the instant a confirm starts and leaves its own `isRunning`
spinner nowhere to appear. `.pvConfirm` leaves dismissal to the caller's
binding so an async action can stay on screen while it runs and report a
failure in `detail`.

## Fonts

Real font files are bundled (not a system-font placeholder) at
`macos/App/Resources/Fonts/`, downloaded from `fonts.gstatic.com` (the same
files Google Fonts serves the web version — SIL Open Font License):

- **Newsreader** (display/headings): weights 300/400/500/600/700, regular only
- **Spectral** (body/UI): weights 300/400/500/600/700 regular, plus 400/600
  italic (italic is a real content convention — hedged statements, place
  lines — not decorative)
- **IBM Plex Mono** (anything cited exactly — dates, ids, folio refs,
  counts): weights 400/500/600

They're registered at process launch via `PVFontRegistration
.registerBundledFontsIfNeeded()` (call this once, early — e.g. from
`ProvenenciaApp.init`), which uses `CTFontManagerRegisterFontsForURL` to
register each `.ttf` under the process scope. This was used instead of the
`ATSApplicationFontsPath` Info.plist key because the project's Info.plist is
auto-generated (`GENERATE_INFOPLIST_FILE = YES`) from `INFOPLIST_KEY_*`
build settings, and per-file registration avoids relying on that key's
exact folder-name semantics.

`PVFont.display/body/mono(size:weight:)` resolve a weight by family name +
numeric weight against whatever's registered (via `NSFontDescriptor`, not a
hardcoded PostScript name — robust to Google Fonts renaming internal
PostScript names between versions), falling back to the system font (display and
body) or the system monospaced face (mono) if a bundled font is ever
unavailable. AppKit faces
are resolved once per family, size, weight and italic and reused.

To add a new weight or style later: download it from
`fonts.googleapis.com/css2?family=...` (see this project's own
`tokens/fonts.css` for the exact family query), drop the `.ttf` into
`Resources/Fonts/`; it is bundled through the `Fonts` folder reference (see
"Xcode project registration" below) — no code changes needed since
`PVFont` resolves by family + weight, not by filename.

## Brand assets

The logo mark (a seal roundel enclosing a stylized "P," drawn in the `iron`
palette) lives in two places:

- **`Assets.xcassets/LogoMark.imageset`** — the in-app mark, used via
  `PVLogoMark`. Two SVGs: `logo-mark.svg` (light, `iron-700`/`iron-300`) and
  `logo-mark-dark.svg` (dark, `iron-300`/`iron-100` — the same light↔dark
  relationship `PVColor.accent` uses), registered as light/dark `appearances`
  in `Contents.json` exactly like `AccentColor.colorset`. Never reference
  `Image("LogoMark")` directly — go through `PVLogoMark` so new placements
  stay consistent.
- **`Assets.xcassets/AppIcon.appiconset`** — the real macOS app icon, the
  classic 10-slot format (16/32/128/256/512pt, each @1x/@2x). Generated (not
  hand-exported) from the *same* `logo-mark.svg`, composed onto a
  warm-parchment background (`PVPalette.paper50`) at ~80% fill via
  `scripts/generate-app-icon.sh` (requires `rsvg-convert`, `brew install
  librsvg`). Rerun that script and commit the resulting PNGs if the mark or
  brand color ever changes — don't hand-edit the PNGs.

## Platform deviations

Three are deliberate; the first two match what the source design system's own
readme documents as the intended macOS port:

- **Control heights**: `PVSpacing.controlHeightSmall/Medium/Large` are
  22/28/36pt (AppKit-native metrics), not the web `--control-h-sm/md/lg`
  pixel values (28/34/42).
- **Icons**: `PVIcon` wraps SF Symbols, not the web's Lucide set — Lucide is
  itself a web-only substitution (no icon set was supplied to the original
  brief), so this is a second substitution, not a compromise.
- **`PVComboBox`'s popup shadow** is AppKit's window shadow rather than
  `--shadow-overlay`. The list lives in its own transparent window, which
  casts a correctly-shaped shadow from its alpha; see "The combo box subset".

## Xcode project registration

`App/App`, `App/Platform`, `App/Features`, `App/DesignSystem` and `ProvenenciaTests` are Xcode synchronized folders: a Swift file saved in one is built into its target with no edit to `project.pbxproj`. Any other file in them is copied into the app bundle, so a README or other note is listed in its folder's membership exceptions (`PBXFileSystemSynchronizedBuildFileExceptionSet`); `scripts/check-guidance-names.py` fails on one that is not.

`Resources/` is a classic group. A new `.ttf` dropped into `Resources/Fonts/` is bundled through the `Fonts` folder reference with no project edit.
