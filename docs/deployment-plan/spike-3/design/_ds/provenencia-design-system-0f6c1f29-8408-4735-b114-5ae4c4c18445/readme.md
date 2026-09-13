# Provenencia — Design System

**Provenencia** is a robust, evidence-based genealogical research application for documenting a family history. Where consumer family-tree products optimise for fast, confident-looking trees, Provenencia optimises for *defensible* ones: the record sits beside the conclusion, every asserted fact carries an evidence grade and a citation, and conflicts stay visible until someone resolves them in writing.

The brand should feel **modern but archival** — the discipline of the reading room, not the pastiche of a scroll-and-quill logo. Warm parchment neutrals, serif type throughout, hairline rules, and a palette drawn from historical pigments (iron gall, madder, verdigris, ochre, lapis, plum, copper) doing real informational work.

## Sources used

This system was authored **from a written brief only** — no codebase, Figma file, screenshots, decks, fonts or logo files were supplied. Consequences to know about:

- **Logo:** no mark existed, so an original one was drawn for the brand: a seal roundel enclosing a four-generation pedigree bracket (`assets/logo-mark.svg`). It is a starting point, not a finished identity — see *Open asks*.
- **Fonts:** substituted from Google Fonts (see *Typography*). Flagged for replacement if licensed files exist.
- **Icons:** substituted with Lucide via CDN (see *Iconography*).
- **Products:** two surfaces were inferred from the brief — a desktop research workspace and a marketing site. Both UI kits are original designs consistent with the brand, not recreations of an existing product.

## Index

| Path | What it is |
|---|---|
| `styles.css` | The single entry point consumers link. `@import` list only. |
| `tokens/` | `fonts.css`, `colors.css`, `typography.css`, `spacing.css`, `radii.css`, `elevation.css`, `motion.css`, `base.css` |
| `assets/` | `logo-mark.svg`, `logo-mark-solid.svg`, `logo-mark-inverse.svg`, `logo-lockup.svg`, `logo-lockup-inverse.svg`, `icons.css` |
| `guidelines/` | 24 foundation specimen cards (Colors, Type, Spacing, Brand) |
| `components/` | React primitives, grouped `core/ forms/ navigation/ feedback/ research/` |
| `ui_kits/app/` | Research workspace — 4 click-through views |
| `ui_kits/site/` | Marketing home page |
| `swift/` | **macOS (SwiftUI) port** — token + component mirror, SF Symbols map, platform notes. See `swift/README.md`. |
| `thumbnail.html` | Homepage tile |
| `SKILL.md` | Agent-Skills front matter for use outside this project |

### Components

**core/** — `Button`, `IconButton`, `Icon`, `Badge`, `Tag`, `Card`, `Tooltip`
**forms/** — `Field`, `Input`, `Select`, `Checkbox`, `Radio`, `Switch`
**navigation/** — `Tabs`, `SidebarNav`, `Breadcrumbs`
**feedback/** — `Dialog`, `Callout`, `Toast`, `EmptyState`
**research/** — `EvidenceBadge`, `SourceCitation`, `PersonChip`, `FactRow`

Each has a sibling `.d.ts` (props contract) and `.prompt.md` (what & when + usage). Import from the compiled bundle: `const { Button } = window.ProvenenciaDesignSystem_*`.

#### Intentional additions

No source defined a component inventory, so a standard set was authored. Five entries are product-specific rather than generic:

- **`EvidenceBadge`** — the five-grade confidence marker; the spine of the product.
- **`SourceCitation`** — citation + repository + scan thumbnail + grade, as one unit.
- **`PersonChip`** — a person with life dates and a lineage-coloured left rule.
- **`FactRow`** — one asserted fact: type glyph, date, value, place, grade, conflict note.
- **`Icon`** — a wrapper over the substituted Lucide set, so swapping icon systems later is a one-file change.
- **`Callout`** — the persistent inline counterpart to `Toast`. Provenencia has standing conditions (an open conflict, an unreceived certificate) that must stay visible in the content flow rather than float past; a transient-only feedback system would hide exactly the information the product exists to surface.

Deliberately **not** built (no evidence they are needed): Avatar, Accordion, Pagination, Popover, Slider, DatePicker, DataGrid, Command palette. Add them when a real screen demands one.

---

## Platforms

The system ships two implementations of the same tokens:

- **Web / prototyping** — `styles.css` + the React components in `components/`. This is what the design-system compiler indexes and what agents build mocks from.
- **macOS app** — `swift/` mirrors every token and the product-specific primitives in SwiftUI (`PVColor`, `PVFont`, `PVButton`, `PVEvidenceBadge`, `PVFactRow`, …). Two deliberate platform deviations: control heights drop to AppKit metrics (22/28/36pt) and icons become **SF Symbols** via `swift/SymbolMap.swift`. Everything else — colour, type, spacing, radii, evidence grades, record types, motion — is identical. Read `swift/README.md` before writing Mac UI.

**When you change a token, change it in both places.** `tokens/*.css` is the source of truth; the Swift file restates it.

---

## Visual foundations

**Colour.** Warm parchment neutrals (`--paper-0` → `--paper-950`); never blue-grey. The primary is **iron gall** `--iron-700` — the blue-slate of oxidised manuscript ink — used for buttons, links, selection and paternal lineage lines. Seven pigment hues each ship 100/300/500/700/900 and carry meaning: **madder** (conflict, destructive actions), **verdigris** (proof, confirmation), **ochre** (inference, marginalia, highlight), **lapis** (sources, citations, links), **plum** (unions, marriages), **copper** (migration, work in progress). Two colour systems sit on top: five **evidence grades** (`--evidence-*`) and eight **record types** (`--rec-*`, deliberately non-sequential so classes are told apart at a glance). Light and dark are both first-class: dark mode re-points every semantic alias, lifts hue lightness, and keeps the warm cast (`--surface-card: #191510`, not neutral grey). Never hard-code a hex in a component — always the semantic alias.

**Type.** All serif. **Newsreader** for display and headings (500–600 weight, `-0.02em`), **Spectral** for running text and UI at 16/14/13px, **IBM Plex Mono** for anything a researcher would cite exactly: dates, GEDCOM ids, folio references, counts. Italic Spectral is reserved for hedged or uncertain statements and place lines — a typographic convention borrowed from record abstracts. Eyebrows are 11px, `0.11em`, uppercase, semibold (`.eyebrow`). Prose measures cap at 66ch.

**Spacing & layout.** 2px base scale (`--space-1`…`--space-15`), non-linear above 32px. Fixed rails: 264px sidebar, 340px inspector, 1180px content max, 32px page gutter (56px on marketing pages). The app is a three-pane, fixed-chrome layout: sticky top bar, scrolling centre, sticky inspector. Table rows use 12px vertical / 16px horizontal padding.

**Backgrounds.** Flat warm surfaces only. **No gradients, no photographic hero images, no repeating textures, no hand-drawn illustration.** Depth comes from surface steps (page → card → raised) and hairline rules. The one high-contrast moment is the closing marketing band on `--paper-950`. Record scans are the only imagery in the product, and they appear as themselves — cropped square-ish, 1px bordered, never filtered or duotoned.

**Cards.** `--surface-card` fill, 1px `--border-subtle`, `--radius-md` (5px), `--shadow-sm`. Optional header with a hairline bottom rule; footers use the sunken surface for metadata. Hoverable cards lift `-1px` and go to `--shadow-md`. Never nest a hoverable card inside another card.

**Borders & radii.** Documents and tables stay square; controls take 3–5px; only avatars/dots/switches go fully round. Four border weights: `subtle` (default divider), `default` (control outline), `strong` (interactive outline), `inked` (emphasis). Dashed `--border-default` means *absence of evidence* — empty states, inferred relationships, unconfirmed nodes. That semantic is reserved; do not use dashes decoratively.

**Shadows.** Warm-tinted and shallow — paper resting on paper, never glass floating in space. `--shadow-sm` at rest, `--shadow-md` on hover, `--shadow-lg` for sticky bars and toasts, `--shadow-overlay` for dialogs. Inputs use `--shadow-inset` so they read as a slot cut into the page. Focus is a 3px `--ring-focus` halo in iron gall, never a colour change alone.

**Motion.** Restrained. 130ms for hover and colour, 200ms for panels and lifts, 320ms for dialogs and drawers, all on `--ease-standard` (`cubic-bezier(.2,.6,.2,1)`). Fades and short slides only — **nothing bounces, nothing springs, nothing bobs**. Hover = background tint plus at most a `-1px` lift; press = `scale(.985)` with no colour flash. Loading is a static count or skeleton rule, not a spinner where avoidable. All durations collapse to 0 under `prefers-reduced-motion`.

**Transparency & blur.** Two sanctioned uses: the dialog scrim (`--surface-overlay` + 2px blur) and the sticky marketing nav (88% surface mix + 8px blur). Hover/active tints are alpha on the accent hue (`--surface-hover`, `--surface-active`) so they work on any surface. No frosted cards, no translucent panels.

**Protection.** Text over a record scan sits in a solid capsule on `--surface-card`, not on a gradient scrim — legibility over atmosphere.

---

## Content fundamentals

**Voice: a careful colleague, not a marketer.** Provenencia is talking to someone who has read a parish register at 2am. It is precise, plain, and never hypes a find.

- **Person.** Address the reader as *you*; the product refers to itself in the third person ("Provenencia writes the citation in Evidence Explained form"). Never *we*, never first-person from the app.
- **Casing.** Sentence case everywhere — buttons, headings, menus. Table headers are the exception, set as 11px caps eyebrows (a typographic style, not a copy rule). No Title Case Buttons. No ALL CAPS in body copy.
- **Buttons are verb-first and specific:** "Attach source", "Add log entry", "Export GEDCOM", "Order certificate". Never "Submit", "OK", "Learn more".
- **Hedging is a feature.** Copy uses the vocabulary of the discipline exactly: *proven, probable, possible, disputed, undocumented*; *direct, indirect, negative* evidence; *original, derivative, authored* sources. "abt 1912" stays "abt 1912" — never silently normalised to "1912".
- **Empty states state an absence, then a next step:** "No sources yet — every fact should trace back to a record."
- **Errors name the record and the line:** "Line 412 of the GEDCOM file could not be parsed." Never "Something went wrong."
- **Transient vs standing feedback.** A `Toast` reports an *event* in the past tense ("Source attached"). A `Callout` states a *condition* in the present tense ("Two sources give different death dates") and stays until the condition changes. If the user cannot clear it, it is not dismissible.
- **Numbers are exact and mono:** "52% of 1,904 facts are proven. 94 conflicts await review."
- **Hints are italic, sentence case, no trailing period:** "Enter the date as written on the record".
- **Tooltips are 2–4 words, no punctuation:** "Copy citation".
- **No emoji. Ever** — not in UI, not in marketing, not in docs. Unicode is used only where typographically correct: en-dash in date ranges (1847–1912), · as a metadata separator, — for unknown values.
- **Marketing headlines are short declaratives** with a full stop: "A family history that can be checked." "Confidence is stated, not implied." No questions, no puns on roots/branches/trees, no "unlock your past".

---

## Iconography

- **System:** [Lucide](https://lucide.dev) — 24px grid, 1.5px stroke, rounded caps, outline only. Chosen because its stroke weight sits comfortably next to a serif at small sizes. **This is a substitution**: no icon set was supplied. Flagged in *Open asks*.
- **Delivery:** glyphs are masked images from `https://unpkg.com/lucide-static@0.469.0/icons/<name>.svg`, painted with `currentColor`. Use the `Icon` component in React (`<Icon name="git-branch" size={16} />`) or `assets/icons.css` in plain HTML. Because it is a mask, icons inherit text colour and work in dark mode with no second asset.
- **No hand-drawn SVG.** If a needed glyph is missing from Lucide, add the closest Lucide name and raise it — never inline a bespoke path.
- **Sizes:** 12px inside badges, 14px in dense rows, 15–16px in nav and buttons, 18–20px in feature blocks, 26px in empty states. Icons are decorative and always aria-hidden; the adjacent label carries the meaning (`IconButton` requires a `label` prop).
- **Record-type glyphs are fixed** so a class is recognisable without reading: birth `baby` · marriage `heart-handshake` · death `cross` · census `table-2` · migration `ship` · military `shield` · probate `gavel` · DNA `dna`. Evidence grades: proven `shield-check` · probable `circle-check` · possible `circle-help` · disputed `circle-slash` · undocumented `circle-dashed`.
- **No icon font, no emoji, no PNG icons.** The only brand art is the logo set in `assets/`.

---

## Typography — substitution notice

| Role | Family | Status |
|---|---|---|
| Display / headings | **Newsreader** | Google Fonts. Transitional, slightly bookish contrast; variable optical sizing. |
| Body / UI | **Spectral** | Google Fonts. A screen-first serif that stays legible at 13px — the reason the whole UI can be serif. |
| Mono | **IBM Plex Mono** | Google Fonts. Slab-ish, pairs with Spectral, unambiguous digits. |

All three load via `@import` in `tokens/fonts.css`, so there are **no local font binaries and no `@font-face` rules** in the project. If Provenencia licenses type, replace that file's `@import` with local `@font-face` rules and drop the files into `assets/fonts/`.

## Open asks

1. **Logo.** The mark is an original proposal, not a delivered identity. If you have a real mark, drop it into `assets/` and it replaces the placeholder everywhere.
2. **Fonts.** Confirm Newsreader / Spectral / IBM Plex Mono, or send licensed files.
3. **Icons.** Confirm Lucide, or name your set.
4. **Product scope.** The two UI kits are inferred surfaces. Real screens (or a repo) would let the app kit become a recreation rather than a proposal.
