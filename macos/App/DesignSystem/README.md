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
project's own sync log, it was built by reading the shipped
`OnboardingView`/`OnboardingChooseFileView`/`OnboardingIdentifyView`/
`OnboardingHomeView`/`OnboardingChrome` SwiftUI files — so it's meant to
restyle those exact screens once this scaffolding is reviewed. **This task
does not touch those files.** Restyling them to use `PV*` tokens and
components is the natural next task.

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
header comment before adding component #2. In short:

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

To add component #2: pick its category folder (create it if this is the
first component in that category), copy `PVButton.swift`'s shape, and read
its exact CSS/JS spec out of the source design-system project first — don't
guess at colors/spacing/sizes.

## What's built vs. not yet

The components the **Onboarding Flow** board actually uses, plus `Toast`
(added when onboarding needed a way to surface errors that wasn't an inline
red `Text`):

| Component | File | 
|---|---|
| Button | `Components/Core/PVButton.swift` |
| Icon | `Components/Core/PVIcon.swift` |
| Field | `Components/Forms/PVField.swift` |
| Input | `Components/Forms/PVInput.swift` |
| Select | `Components/Forms/PVSelect.swift` |
| Toast | `Components/Feedback/PVToast.swift` |

The other 17 design-system components have **no files yet** — add them on
demand, following the pattern above, when a screen needs one:

| Component | Category | Purpose |
|---|---|---|
| Badge | Core | Small status/label pill |
| Card | Core | Bordered content container with optional header/footer |
| IconButton | Core | Icon-only button (toolbar actions) |
| Tag | Core | Removable/interactive pill with a color dot |
| Tooltip | Core | Hover label |
| Dialog | Feedback | Modal dialog |
| EmptyState | Feedback | "Nothing here yet" placeholder |
| Checkbox | Forms | Checkbox control |
| Radio | Forms | Radio control |
| Switch | Forms | Toggle switch |
| Breadcrumbs | Navigation | Path trail |
| SidebarNav | Navigation | Sidebar navigation rail |
| Tabs | Navigation | Tab strip |
| EvidenceBadge | Research | The 5-grade confidence marker (proven/probable/possible/disputed/undocumented) |
| FactRow | Research | One asserted fact: type glyph, date, value, place, grade, conflict note |
| PersonChip | Research | A person with life dates and a lineage-colored rule |
| SourceCitation | Research | Citation + repository + scan thumbnail + grade, as one unit |

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

## Platform deviations

Two are deliberate, matching what the source design system's own readme
documents as the intended macOS port:

- **Control heights**: `PVSpacing.controlHeightSmall/Medium/Large` are
  22/28/36pt (AppKit-native metrics), not the web `--control-h-sm/md/lg`
  pixel values (28/34/42).
- **Icons**: `PVIcon` wraps SF Symbols, not the web's Lucide set — Lucide is
  itself a web-only substitution (no icon set was supplied to the original
  brief), so this is a second substitution, not a compromise.

## Xcode project registration

`macos/Provenencia.xcodeproj/project.pbxproj` is a classic manually-authored
Xcode project (not Xcode 16 synchronized groups), so a new file on disk is
invisible to the build until it's added as a `PBXFileReference` +
`PBXBuildFile` (and, for source files, a `Sources` build phase entry). Use
Xcode itself (File → Add Files…) or the `xcodeproj` Ruby gem — do not hand
edit `project.pbxproj`'s GUIDs.
