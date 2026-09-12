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

Every value in these files is transcribed from the CSS, not invented.
Reach for `PVColor`/`PVFont`/`PVSpacing`/`PVRadius`/`PVElevation`/`PVMotion`
at call sites — never a literal color, font, or number.

## The component pattern

`Components/Core/PVButton.swift` is the **canonical example** — read its
header comment before adding a new component. In short:

- A component lives at `Components/<Category>/PV<Name>.swift`, where
  `<Category>` mirrors the web design system's own `components/<category>/`
  folder (`core`, `forms`, `navigation`, `feedback`, `research`).
  `PVButton` is `components/core/Button.jsx` → `Components/Core/PVButton.swift`.
- The file's header comment names the `.jsx` it mirrors and calls out any
  deliberate deviation (a prop that isn't ported yet, a platform-specific
  approximation).
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

To add another component: pick its category folder (create it if this is the
first component in that category), copy `PVButton.swift`'s shape, and read
its exact CSS/JS spec out of the source design-system project first — don't
guess at colors/spacing/sizes.

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

## What's built vs. not yet

The components the **Onboarding Flow** board actually uses, plus `Toast`
(added when onboarding needed a way to surface errors that wasn't an inline
red `Text`), plus `Badge`/`EmptyState`/`Callout` (added for the S2-02
**Source fields** board — see `Features/SourceFields/`):

| Component | File | 
|---|---|
| Button | `Components/Core/PVButton.swift` (icon-left, loading spinner, and a chrome-less `link` variant added for S2-02) |
| Icon | `Components/Core/PVIcon.swift` |
| Field | `Components/Forms/PVField.swift` |
| Input | `Components/Forms/PVInput.swift` |
| Select | `Components/Forms/PVSelect.swift` |
| Toast | `Components/Feedback/PVToast.swift` |
| LogoMark | `Components/Core/PVLogoMark.swift` |
| SidebarNav | `Components/Navigation/PVSidebarNav.swift` (added for the S2-01 workspace chrome; ports that board's revised `collapsed`-capable `SidebarNav.jsx`) |
| IconButton | `Components/Core/PVIconButton.swift` (added for the workspace sidebar's collapse toggle, which needed real hover feedback; `label` is required per `IconButton.jsx` and doubles as the `.help` tooltip; pass `accessibilityLabel` when the spoken label has to name a target the tooltip can leave implicit; `tone: .danger` tints a destructive action) |
| Badge | `Components/Core/PVBadge.swift` (added for S2-02's data-type/origin badges; a glyph-only variant carries the S2-22 seeded pill) |
| Divider | `Components/Core/PVDivider.swift` (1pt hairline; horizontal/vertical) |
| EmptyState | `Components/Feedback/PVEmptyState.swift` (added for S2-02's empty/no-match states; the web spec's `action` slot isn't ported — see the file's header comment) |
| Callout | `Components/Feedback/PVCallout.swift` (added for S2-02's "this field is locked" note; only the subset S2-02 needs is ported — see the file's header comment) |
| Table | `Components/Data/PVTable.swift` (added for S2-22, extracted from the Source fields list; see "The table tradeoff" below) |
| Confirm | `Components/Feedback/PVConfirm.swift` (added for S2-22's delete confirmation; the macOS answer to `ConfirmDialog.jsx`, which the web spec says not to port — see "Confirmations are system chrome" below) |
| ComboBox | `Components/Forms/PVComboBox.swift` (added for S2-16's assign-field control, where the pool is the whole Source fields vocabulary; single-select subset only — see "The combo box subset" below) |
| Thumbnail | `Components/Core/PVThumbnail.swift` (added for S2-17 Sources list rows; image / evidence glyph / SF glyph / empty / loading tile) |
| EvidenceIcon | `Components/Research/PVEvidenceIcon.swift` (PR1 `file_*` MIME stand-ins; see `EVIDENCE-ICONS.md`; not SF Symbols) |
| List | `Components/Data/PVList.swift` (added for S2-17 evidence browse — not `PVTable`; Files remounts it in S2-21) |
| Dialog | `Components/Feedback/PVDialog.swift` (added for S2-17 Add Source; sheet form with content slot + footer — see note below) |
| Breadcrumbs | `Components/Navigation/PVBreadcrumbs.swift` (added for S2-18 Source page trail; Sources → `SRC-…`) |
| ReorderableList | `Components/Data/PVReorderableList.swift` (+ `PVReorderHandle`; added for S2-25 Metadata drag order) |

The other design-system components have **no files yet** — add them on
demand, following the pattern above, when a screen needs one:

| Component | Category | Purpose |
|---|---|---|
| Card | Core | Bordered content container with optional header/footer |
| Tag | Core | Removable/interactive pill with a color dot |
| Tooltip | Core | Hover label — on macOS this is usually SwiftUI's own `.help()`, which is what `PVIconButton` uses; port the web hover card only if a call site needs richer content |
| Checkbox | Forms | Checkbox control |
| Radio | Forms | Radio control |
| Switch | Forms | Toggle switch |
| Tabs | Navigation | Tab strip |
| EvidenceBadge | Research | The 5-grade confidence marker (proven/probable/possible/disputed/undocumented) |
| FactRow | Research | One asserted fact: type glyph, date, value, place, grade, conflict note |
| PersonChip | Research | A person with life dates and a lineage-colored rule |
| SourceCitation | Research | Citation + repository + scan thumbnail + grade, as one unit |

**Dialog on macOS:** the web `Dialog.jsx` draws a scrim because the browser
gives it none. Here `PVDialog` is a `.sheet` — same family as `PVConfirm`'s
rich sheet — and leaves window chrome to the system. Prefer `PVConfirm` for
destructive confirmations; use `PVDialog` for short create/edit forms.

## The combo box subset

`PVComboBox` ports `components/forms/ComboBox.jsx`, but only the
**single-select** half of it. The token/multi-select field, the "use what you
typed" create row, the `person` row kind and async `loading` are all in the
web component and none of them have a call site here yet — add them the way
`PVSidebarNav` was added, rather than porting speculatively.

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
  `NSHostingView.fittingSize` can be trusted to say. `PVComboBox` lays a
  hidden copy of the rows out behind the field purely to measure them, hands
  that number to the window, and lets the scroller simply fill whatever it
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
`ScrollViewReader`. Like a native table, the highlight dims when the table
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

`Components/Feedback/PVConfirm.swift` carries the web component's **copy
rules** across without its chrome, and offers the two right answers:

| Modifier | Use |
|---|---|
| `.pvConfirm(isPresented:copy:tone:onConfirm:)` | **The default.** A system alert — Apple's own pattern, fully system-drawn, inherits keyboard, VoiceOver and Reduce Motion for free. Plain-text message only. |
| `.pvConfirmSheet(item:copy:tone:isRunning:onConfirm:detail:)` | When the consequence needs rich content — a mono-set key (`PVConfirmKeyChip`), a list of affected records. Chrome still belongs to the window; only content and the button row are ours. Keyed to the record it names (`item:`, not an `isPresented` Bool) so the copy and detail render from a snapshot and the sheet animates out still showing them, rather than blanking the instant the model clears. |

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
system *alert* form still draws its own buttons and stays fully native.)
In-content shapes keep Provenencia radii as before — `PVConfirmKeyChip`
uses `PVRadius.xs`, matching the design's square treatment for citable
values. Because a custom `ButtonStyle` draws no focus indication of its
own, `PVButtonStyle` shows `pvFocusRing` when focused, so the sheet's
cancel-first focus stays visible to keyboard users.

The copy rules travel in `PVConfirmCopy`: the title is a question naming the
record ("Delete Photographer?", never "Are you sure?"), the message says what
is *and is not* lost, confirm repeats the verb ("Delete field", never "OK"),
and cancel names the safe outcome ("Keep field"). **Focus starts on cancel** in
both forms, so Return cannot complete a destructive action by reflex.

Two further rules from the spec: a blocked action never reaches a confirmation
— disable the control and explain why in its tooltip, the way Source fields'
delete button does. And a *reversible* action should not confirm at all: act,
then offer undo in a `PVToast`.

One deviation from the design system's `swift/ProvenenciaConfirm.swift`
reference: it clears `isPresented` before invoking `onConfirm`, which closes
the sheet the instant a confirm starts and leaves its own `isRunning` spinner
nowhere to appear. `pvConfirmSheet` leaves dismissal to the caller's binding
so an async action can stay on screen while it runs and report a failure in
`detail`.

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
PostScript names between versions), falling back to the closest system
serif/monospace face if a bundled font is ever unavailable.

To add a new weight or style later: download it from
`fonts.googleapis.com/css2?family=...` (see this project's own
`tokens/fonts.css` for the exact family query), drop the `.ttf` into
`Resources/Fonts/`, and add it to the Xcode project's Resources phase (see
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

`macos/Provenencia.xcodeproj/project.pbxproj` is a classic manually-authored
Xcode project (not Xcode 16 synchronized groups), so a new file on disk is
invisible to the build until it's added as a `PBXFileReference` +
`PBXBuildFile` (and, for source files, a `Sources` build phase entry). Use
Xcode itself (File → Add Files…) or the `xcodeproj` Ruby gem — do not hand
edit `project.pbxproj`'s GUIDs.
