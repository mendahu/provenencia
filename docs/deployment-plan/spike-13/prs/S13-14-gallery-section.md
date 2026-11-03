# S13-14 — Gallery section

**Status:** Blocked — on board S13-D5
**Spike / slice:** Spike 13 · Slice 5 — Gallery
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D5](../design/S13-D5-gallery-and-viewer.md), board not drawn yet
**Board:** not drawn yet. When agreed, the Claude Design project, file, turn and frames go here; the implementer may not be able to open it, so **Design → macOS** will describe every frame in words
**Depends on:** S13-10 (landed [#634](https://github.com/mendahu/provenencia/pull/634): the sections, Gallery as a place), S14-02 (landed [#641](https://github.com/mendahu/provenencia/pull/641): the kind-blind `gallery` slot and `ConclusionLeadModel`), S13-13 (landed [#691](https://github.com/mendahu/provenencia/pull/691): Page columns' full-width mode per section, `ConclusionPageSection.fillsWidth`), S13-D5 (board)
**Blocks:** S13-15
**Branch:** `s13-14-gallery-section`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go or proto change
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-12 and S13-13 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-depiction-value`](../../../../.agents/skills/add-depiction-value/SKILL.md) (read only: what a gallery picture is)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 4. Gallery](../person-detail.md#4-gallery) (R-29, R-31, R-32; R-1, R-2, R-13b); brief [S13-D5](../design/S13-D5-gallery-and-viewer.md) GA-1 – GA-10, GA-W1, GA-W2; the S14-02 plan ([`../../spike-14/prs/S14-02-pictures-in-the-app.md`](../../spike-14/prs/S14-02-pictures-in-the-app.md)) D-1 – D-5; working rules: "The Person's gallery is redrawn on the Person page only", "Sections are places", "Redesigns adopt Page columns", "One sectioned Conclusion page, not a Person fork"; Spike 14 working rules "The gallery layer is named gallery, not likeness", "An Event or Place chooses its thumbnail as a Person chooses its portrait"
**Planned:** 2026-11-03 against `main` @ `42eb321` (draft: everything that does not depend on the board's frames; § 12 lists what the board settles)
**Review:** [`S13-14-gallery-section.review.md`](S13-14-gallery-section.review.md) — round 1 Fits with changes / Ready with changes, round 2 Fits with changes / Ready with changes; every finding Fixed or Declined; reviewed again on the board refresh

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

A researcher opens a Person's Gallery section and sees every likeness of the person at a size where faces can be told apart, each captioned with its Source, with the portrait marked, Set as portrait on every other likeness, and a visible way back to the earliest likeness. A click on a likeness's image opens its Citation, as today, until S13-15's viewer takes the click. At Wide the grid fills the pane. The slice's Check is bar 5 without the viewer (S13-15 adds the viewer).

**Why now.** Slice 5's first chrome PR. S13-10 moved today's likeness grid into the Gallery section unchanged and left its redesign, and where "Use earliest likeness" lives, to the S13-D5 board (S13-10 § 1 What stays). S14-02 made the gallery a kind-blind slot, so the Person's Gallery can be redrawn while the Event and Place Pictures keep theirs (map scope entry S13-14).

**What stays.** The header, its thumb and the thumb's right-click "Use earliest likeness" menu, shared with Event and Place (user, 2026-11-03). The gallery's order and lead rule (Go, S14-01). `ConclusionDetailGallerySection` and its tiles for the Event and Place pages' Pictures, drawn by `ConclusionDetailDefaultPageSections` (user, 2026-11-03: Person only). `ConclusionLeadModel` and its Toasts. Until S13-15 a click on a tile's image opens its Citation, as today (brief § 3). The user's calls of 2026-11-03 (R-13b, R-32 updated): tiles get no provenance control and no separate Open citation; a likeness's provenance shows in S13-15's viewer, under the image, and on the Vitals Likeness row's sheet meanwhile. No Go, proto or schema change.

## 2. User stories

**US-1.** As a researcher, when I open a Person's Gallery, I see every likeness large enough to recognise a face, with where each came from, so that I can compare them.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann Weber with the brief's seven likenesses (§ 5) | Seven tiles in the engine's order, each at the board's tile size (GA-1); the caption of tile 5 is cut at two lines; tile 6's caption "OBS-7Q2KD" is in mono; tile 7 shows the Person mark in Johann's gender colors |
| AC-1.2 | VoiceOver on tile 2 | Reads "Likeness from Hamburger Adressbuch 1861, Volume 3, Entry 214, p. 23" (the citation line) with the hint "Open citation"; on tile 4 "Likeness from Carte de visite, Atelier Strumper, Hamburg, Front, portrait" |
| AC-1.3 | A person with no likeness | The EmptyState "No likenesses yet" with the photo glyph (GA-8) |
| AC-1.4 | The Event and Place pages | Their Pictures section looks and behaves as before (96pt tiles, click opens the Citation, no Open citation control, no Use earliest control). Only their VoiceOver changes: the label names the citation line, the lead tile adds ", thumbnail", and "Open citation" moves from the label to the hint (D-3, D-4): "Picture from Hamburger Nachrichten, Issue 221, p. 4, thumbnail", hint "Open citation" |

**US-2.** As a researcher, I set a likeness as the portrait, or go back to the earliest, from the Gallery.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Tile 2 → Set as portrait | The header thumb shows likeness 2; tile 2 carries the Portrait Badge; the Toast reads "Portrait updated" · "The likeness from Hamburger Adressbuch 1861 is now the portrait."; focus stays on tile 2 |
| AC-2.2 | Likeness 4 chosen, likeness 1 is the earliest | The Gallery shows the Use earliest likeness control (GA-5); activating it makes tile 1 lead, the control disappears, the Toast reads "Portrait updated" · "The earliest likeness is the portrait again." |
| AC-2.3 | No choice (tile 1 leads), or the chosen likeness is the earliest | No Use earliest likeness control in the Gallery |
| AC-2.4 | A save is in flight | Every Set as portrait and the Use earliest likeness control are disabled |

**US-3.** As a researcher, I open the Citation a likeness was cut from, until the viewer replaces the click.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A click on tile 2's image, or Return with it focused | The Citation composer opens at the likeness (`ConclusionDetailPicture.location`), as today; Back returns to Johann's Gallery |
| AC-3.2 | Any tile | No provenance control, no Open citation control and no context menu (user, 2026-11-03) |

**US-4.** As a researcher on a wide monitor, the gallery uses the width.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Pane 2296pt, 40 likenesses | The grid fills the pane between the gutters with more tiles per row; tiles keep their size; the header keeps its band (GA-W1) |
| AC-4.2 | Resize across 1600 with tile 12 focused | The section, scroll position and focus stay (GA-W2) |

## 3. Scope

| In | Covered by |
| --- | --- |
| The Person's Gallery section from the board: tile, caption, Portrait Badge, Set as portrait (GA-1 – GA-3, GA-7) | STEP-2, STEP-3, AC-1.1, AC-2.1 |
| The image's click and Return open the Citation, as today, until S13-15 (GA-4) | STEP-2, AC-3.1, AC-3.2 |
| The visible Use earliest likeness control (GA-5) | STEP-4, AC-2.2, AC-2.3 |
| The heading line as the board decides (GA-6) | STEP-3 |
| Empty, refreshing, saving states (GA-8; brief § 2) | STEP-3, AC-1.3, AC-2.4 |
| The citation line on each gallery tile, for the tooltip and VoiceOver (GA-2, GA-9) | STEP-1, AC-1.2 |
| Keyboard and VoiceOver (GA-9), dark mode (GA-10) | STEP-3, manual check |
| Full width at Wide (GA-W1, GA-W2) | STEP-5, AC-4.1, AC-4.2 |

| Out | Home |
| --- | --- |
| The likeness viewer; the click on a tile opening it | S13-15 |
| The new grid on the Event and Place Pictures | not planned this spike (user, 2026-11-03); the Event and Place redesign (R-1) |
| A direct Open citation control | not planned (user, 2026-11-03) |
| A provenance control or sheet on a tile, a context menu | not planned (user, 2026-11-03): the viewer's record row (S13-15) shows a likeness's provenance |
| A change to the header thumb or its right-click menu | not planned (user, 2026-11-03) |
| Delete, reorder or caption editing of a likeness | not planned (brief § 2.1) |
| Grouping or filtering the gallery | not planned (brief § 2.1) |
| Retiring `ConclusionDetailGallerySection` | the Event and Place redesign, when they adopt this section (D-2) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Person Gallery case | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` `.gallery` | `PVEmptyState` when `content.gallery.tiles` is empty, else `ConclusionDetailGallerySection(projectDir:pageIdentifier:gallery:lead:)` — the same view the Event and Place pages draw |
| Shared gallery section | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailGallerySection` | `PVSectionHeader(title: gallery.title, meta: count)`; adaptive `LazyVGrid` of 120pt columns; each tile a `GalleryTileView` whose crop is a `Button` → `navigation.go(to: tile.location)`, then the lead `PVBadge` or a ghost `PVButton(gallery.choose.useAs)` → `lead.choose(tile, copy:)`, disabled while `lead.isSaving`; focus returns to the tile after a choice |
| Tile view | `ConclusionDetailPage.swift` `GalleryTileView` (private) | `CachedThumbnail(size: CachedThumbnail.pictureCrop /* 96 */, fallbackMark:)` with no gender colors, a two-line muted caption of `sourceTitle` |
| Default fill | `ConclusionDetailPage.swift` `ConclusionDetailDefaultPageSections` | Event and Place: rows, then `ConclusionDetailGallerySection` when tiles exist, then relationship sections |
| Gallery slot | `ConclusionDetailPage.swift` `ConclusionDetailGallery` (`title`, `openTip`, `fallbackMark`, `tiles`, `choose: LeadCopy`, `lead`, `offersEarliest`), `ConclusionDetailGallery.tiles(from:tileLabel:useAsLabel:chosenBody:)`, `ConclusionDetailPicture` | Each tile carries `id`, `imageRelPath`, `sourceTitle` (read by `ConclusionDetailThumbnail.leading` and the old section's caption), `location`, `isLead`, `accessibilityLabel` (from `tileLabel(sourceTitle)`), `useAsLabel`, `chosenToastBody`. No citation line |
| Person gallery | `macos/App/Features/Conclusions/PersonDetailContent.swift` `init` | Builds the gallery with `L10n.Conclusions.likenessesSection`, `likenessOpenTip`, `.subjectPerson`, `likenessTileLabel`, `useAsPortraitLabel`, `toastPortraitChosenBody`, `.personPortrait`; `treatment` holds the gender treatment (`thumbnailColors`) |
| Lead model | `macos/App/Features/Conclusions/ConclusionLeadModel.swift` | `choose(_:copy:)`, `useEarliest(copy:)`, `isSaving`, `toast`; owned by `ConclusionDetailPage`, passed in `ConclusionDetailPageSectionsContext.lead` |
| Header menu | `ConclusionDetailPage.swift` `ConclusionDetailHeader.thumbnail` | `pvRightClickMenu` with `copy.useEarliest` while `content.gallery.offersEarliest` |
| Citation line | `macos/App/Features/Conclusions/CitationLineDisplay.swift` `CitationLineDisplay.line(_:)` | "Source, Artifact, Citation, page" from a `CatalogReconcilerOutcome` (S13-11) |
| Full width | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection.fillsWidth` | `true` for `.timeline` only; `ConclusionDetailPage.swift` passes `primaryFill: pageSection?.fillsWidth` to `PVPageColumns` |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Conclusions` | `likenessesSection` "Likenesses", `likenessOpenTip` "Open citation", `likenessTileLabel(_:)` "Likeness from %@. Open citation", `pictureTileLabel(_:)` "Picture from %@. Open citation", `portraitBadge`, `useAsPortrait`, `useAsPortraitLabel(_:)`, `useEarliestLikeness` "Use earliest likeness", `toastPortraitUpdatedTitle`, `toastPortraitChosenBody(_:)`, `toastPortraitEarliestBody`, `galleryEmptyTitle`, `galleryEmptyMessage` |
| Tile button | `macos/App/DesignSystem/Components/FocusRingButton/` `PVFocusRingButtonStyle`; `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` `pvActivatesOnReturn(when:_:)` | Today's tile uses the focus ring style but not Return activation; a SwiftUI `Button` does not activate on Return on macOS, so `PersonTimelineSection` adds `pvActivatesOnReturn(when:)` with its focused id |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` `likenessesBecomeTilesEarliestFirst`, `portraitFlagsFollowThePreferredLikeness`, `headerThumbIsTheLeadLikeness`, the `fillsWidth` loop (line 387); `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift` fixture `picture(_:path:title:leads:)` | The gallery content and the lead flags; `fillsWidth == (section == .timeline)` |

**Precedent.** S13-13 ([#691](https://github.com/mendahu/provenencia/pull/691)) replaced an S13-10 placeholder case in `PersonDetailSections` with a section view of its own and opted it into `fillsWidth`; copy that shape (a new file in `Features/Conclusions/`, one case line in `PersonDetailSections`, one `fillsWidth` case, the test loop updated). S14-02 ([#641](https://github.com/mendahu/provenencia/pull/641)) made the gallery kind-blind (`ConclusionDetailGallery`, `LeadCopy`): the new section reads only that slot and the lead model, as `ConclusionDetailGallerySection` does. Do not copy S13-13's own read and host: the gallery comes with the page's detail read.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The Person page only. `PersonDetailSections` `.gallery` draws the new section; `ConclusionDetailDefaultPageSections` keeps drawing `ConclusionDetailGallerySection` for Event and Place | User, 2026-11-03; R-1 (Event and Place keep their look) | All three kinds — the Event and Place pages would change look without a board |
| D-2 | The new section is `ConclusionGallerySection` *(new)* in `macos/App/Features/Conclusions/ConclusionGallerySection.swift` *(new)*, kind-blind: it takes `projectDir`, `pageIdentifier`, `gallery: ConclusionDetailGallery`, and `lead: ConclusionLeadModel`, and speaks only the slot's words (every string comes from `ConclusionDetailGallery`, its `LeadCopy` or its tiles). `ConclusionDetailGallerySection` stays unchanged for Event and Place; both are known debt that the Event and Place redesign removes by adopting `ConclusionGallerySection` | The map scope entry: "the old and new galleries can coexist and retiring the old one does not drop the Event and Place galleries"; kind-blind so the second caller adopts it with no fork (working rule "One sectioned Conclusion page, not a Person fork") | A `style:` parameter on `ConclusionDetailGallerySection` — every line of its body would branch on it; a Person-named view — Event and Place could not adopt it |
| D-3 | `CitationLineDisplay` gains `parts(_:) -> CitationLineDisplay.Parts` *(new)*: `title` (`ReconciliationRecord.sourceTitle(for:)`), `titleIsRef` (`outcome.sourceTitle.isEmpty`, the same untrimmed test `sourceTitle(for:)` uses), and `line` (today's output). `line(_:)` returns `parts(_:).line`. Each `ConclusionDetailPicture` gains one field `citation: CitationLineDisplay.Parts` *(new)*, set in `ConclusionDetailGallery.tiles(from:…)`, and loses `sourceTitle` (the same value as `citation.title`): `ConclusionDetailThumbnail.leading` reads `leadLabel(lead.citation.title)` and `ConclusionDetailGallerySection`'s caption reads `tile.citation.title`. The tile label closures receive `citation.line` instead of the Source title, for every kind | One formatter, one output for the Spike 15 citation formatter to replace (`CitationLineDisplay`'s doc; `spike-15/citation-formats.md`); one value instead of loose fields | A second formatter in the section; separate `citationLine` / `titleIsRef` fields |
| D-4 | `tiles(from:…)` takes two more kind closures: `leadTileLabel: (String) -> String` and `recordTitle: (String) -> String`, and sets `accessibilityLabel` from `leadTileLabel` when `$0.leads`, else `tileLabel`. Person passes `likenessTileLabelLead` *(new)* and `L10n.Conclusions.likenessLabel` ("Likeness from %@", as its header thumb); `pictures(from:…)` passes `pictureTileLabelLead` *(new)* and `L10n.Conclusions.pictureFrom` (as their header thumbs) | Every per-tile string is built where the others are, by kind (M-1); the section stays kind-blind; Event and Place adopt with no new plumbing | Picking a Person key in the section by `tile.isLead` — the section would speak Person words |
| D-6 | The Gallery's Use earliest likeness control calls `lead.useEarliest(copy: gallery.choose)` and shows while `gallery.offersEarliest`; its title is `gallery.choose.useEarliest` | The header menu's rule and copy, so the two controls never disagree (user, 2026-11-03: keep the header menu) | A Gallery-only rule — two rules for one state |
| D-7 | `ConclusionDetailGallery` gains `fallbackColors: PVThumbnail.Colors? = nil` *(new)* beside `fallbackMark`; `PersonDetailContent` sets it from `treatment?.thumbnailColors`; `ConclusionDetailThumbnail.leading` reads it and drops its `colors:` parameter (its only caller with colors is `PersonDetailContent`). The tile's fallback wears it | Brief § 2 (the Person mark in the gender colors, as the header thumb); one slot fact instead of a parameter on every view (M-4) | A `fallbackColors` parameter on the section and the viewer |
| D-8 | `ConclusionPageSection.gallery.fillsWidth` is `true` | GA-W1 (rule 3: grids fill the width), the Timeline precedent | A secondary column — the section has nothing supporting to put in it |
| D-9 | One tile view for both sections: `GalleryTileView` becomes `internal` and gains `size: CGFloat`, `colors: PVThumbnail.Colors?`, `help: String` and `captionIsMono: Bool`; Both pass `caption: tile.citation.title`. `ConclusionDetailGallerySection` passes `CachedThumbnail.pictureCrop`, nil, `help: tile.citation.title` and false (no visible change for Event and Place) and adds `.accessibilityHint(gallery.openTip)` to its image button (Event and Place keep "Open citation"); `ConclusionGallerySection` passes its `static let tile` (the board's GA-1 value), `gallery.fallbackColors`, `help: tile.citation.line` and `tile.citation.titleIsRef`. `CachedThumbnail.pictureCrop`'s doc says it is the Event and Place size. `ThumbnailCache.maxPixelSize` (512) is not changed: it covers a tile up to 256pt at 2× | One tile, not two copies of crop, caption and focus (M-5) | A private tile in the new file |
| D-11 | The tile image's button keeps `PVFocusRingButtonStyle`, adds `pvActivatesOnReturn(when: focusedID == tile.id)` and `.accessibilityHint(gallery.openTip)` (the Person's `likenessOpenTip`, "Open citation"); it alone carries `tile.accessibilityLabel`. The tile's container is `.accessibilityElement(children: .contain)` with no label of its own, so the label is read once | GA-9 (Return opens it); the Timeline precedent. The button's action is `navigation.go(to: tile.location)`, as today, until S13-15 | `onKeyPress(.return)` by hand |
| D-10 | No new query, read, cache key or navigation place. The gallery stays on `GetConclusionDetail` (working rule "Each section reads its own query": Gallery is part of today's detail read) | The rule names Gallery as part of `GetConclusionDetail` | A gallery read of its own — no data is missing |

The board settles GA-1's size, GA-2's second line, GA-3's and GA-4's placement, GA-5's place and GA-6's heading; § 12 lists them. They are layout inside `ConclusionGallerySection` and change no Decision above. A tile the board draws that `GalleryTileView`'s parameters cannot express is a stop condition, not a second tile.

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionGallerySection` *(new)*, `ConclusionGallerySection.swift` *(new)* | view, file | `ConclusionDetailPage.swift` `ConclusionDetailGallerySection` | D-2: the Event and Place pages keep the old one unchanged; the new one is kind-blind so they adopt it later |
| `GalleryTileView` parameters (`size`, `colors`, `help`, `captionIsMono`) | view change | `GalleryTileView` | Extended, not copied (D-9) |
| `CitationLineDisplay.Parts` *(new)*, `parts(_:)` *(new)* | type, func | `CitationLineDisplay.line(_:)` | `line` is rebuilt on it (D-3) |
| `ConclusionDetailPicture.citation` *(new)*; `ConclusionDetailGallery.fallbackColors` *(new)* | fields | `fallbackMark`, `accessibilityLabel`, `useAsLabel` | Same place and shape as the slot's existing facts (D-3, D-7) |
| `L10n.Conclusions.likenessTileLabelLead(_:)`, `pictureTileLabelLead(_:)` *(new)* | keys | `likenessTileLabel(_:)`, `pictureTileLabel(_:)` | New wordings only (D-4) |

**Builds on.** The S14-02 gallery slot and lead model; `CitationLineDisplay`; Page columns' full-width mode (S13-13).
**Deletes.** In `PersonDetailSections`, the Person's use of `ConclusionDetailGallerySection`; the `colors:` parameter of `ConclusionDetailThumbnail.leading` (D-7); `ConclusionDetailPicture.sourceTitle` (D-3); ". Open citation" at the end of `likenessTileLabel` and `pictureTileLabel` (moved to the hint). Nothing else becomes dead: Event and Place still call the old section. The two sections side by side are recorded debt: the map's working rule "The Person's gallery is redrawn on the Person page only" (added with this plan) records it, owned by the Event and Place redesign, as "The provenance sheet replaces the Why on the Person page only" is.
**Next adds.** S13-15 adds the viewer to `ConclusionGallerySection` (the image's click opens it). The Event and Place redesign passes their gallery to `ConclusionGallerySection` and deletes `ConclusionDetailGallerySection`; their per-tile strings already exist (D-4).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Use earliest likeness from the Gallery | `ConclusionLeadModel.useEarliest(copy:)` | The header menu's call; the save's effects refresh the page |

### Scale

At the brief's sizes (1 – 6 typical, 40 large): one `CachedThumbnail` per tile through `ThumbnailCache` (100 entries, 4 decodes at once) in a `LazyVGrid`, as today. The cache holds 40 tiles at 512 pixels. No write is added: Set as portrait and Use earliest are today's writes.

### Edges

- **Assumptions.** `offersEarliest` assumes the tiles are in the engine's order with the earliest first and exactly one `isLead` (S14-01's lead rule); the section does not reorder. Held by `portraitFlagsFollowThePreferredLikeness`.
- **Neighbors.** A gallery of one: the tile leads, no Set as portrait, no Use earliest (T-3). The chosen likeness is the earliest: no Use earliest (AC-2.3, T-3).
- **Removed inputs.** A likeness deleted elsewhere drops out on the page's refresh; focus on a removed tile falls to the grid (SwiftUI default). No state of the section holds a tile id across refreshes except `focusedID`, which then matches nothing.
- **Real data.** None — no table-driven rule.
- **Untrusted size.** The tile count is bounded by the detail read; the image decode by `ThumbnailDecodePolicy` (4096 edge) and `ThumbnailCache.maxPixelSize`.

### Schema

None — no catalog change.

### Go core

None — the gallery's data and order are S14-01's.

### FFI

None —.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Gallery section | `ConclusionGallerySection` *(new)* | `Features/Conclusions/ConclusionGallerySection.swift` | D-2. Heading line per GA-6 (`PVSectionHeader(title:meta:actions:)` with the count as `meta` and the D-6 control in `actions:`, if the board keeps it) |
| Tile | `GalleryTileView` (now internal), `caption: tile.citation.title` | `ConclusionDetailPage.swift` | D-9; the image button per D-11 |
| Portrait mark | `PVBadge(gallery.choose.badge, tone: .accent, icon: .photo)` | — | As today |
| Set as portrait | `PVButton(gallery.choose.useAs, variant: .ghost, size: .sm)` | — | As today, disabled while `lead.isSaving`, focus back to the tile after a choice |
| Tile button | `macos/App/DesignSystem/Components/FocusRingButton/` `PVFocusRingButtonStyle`; `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` `pvActivatesOnReturn(when:_:)` | Today's tile uses the focus ring style but not Return activation; a SwiftUI `Button` does not activate on Return on macOS, so `PersonTimelineSection` adds `pvActivatesOnReturn(when:)` with its focused id |
| Use earliest likeness | `PVButton(gallery.choose.useEarliest, variant: .ghost, size: .sm)` | — | D-6, GA-5 |
| Empty | `PVEmptyState` | `PersonDetailView.swift` | Unchanged |

States: empty (EmptyState), saving (buttons disabled), refreshing (the switcher row's indicator, unchanged), image missing (fallback mark in colors). Accessibility: each tile is one unlabelled `.accessibilityElement(children: .contain)`; its image button carries the D-3/D-4 label (D-11); the image's button is the open action (the Citation now, with the hint `gallery.openTip`; the viewer in S13-15); Set as portrait is labelled with the Source; ids `"\(pageIdentifier).picture.\(id)"`, `.leadBadge`, `.useAsLead` (kept), `"\(pageIdentifier).gallery.useEarliest"` *(new)*.

### Strings

L10n keys *(new)* in `L10n.Conclusions` (`%@` is the citation line):
- `likenessTileLabelLead(_:)` = "Likeness from %@, portrait" (`conclusions.person.likenessTile.lead`)
- `pictureTileLabelLead(_:)` = "Picture from %@, thumbnail" (`conclusions.picture.tile.lead`)

Reused: `likenessOpenTip` ("Open citation") as the image's hint until S13-15.

Changed: `likenessTileLabel(_:)` becomes "Likeness from %@" and `pictureTileLabel(_:)` "Picture from %@"; ". Open citation" moves to the image's hint in both sections (D-9, D-11), and both receive the citation line. Their comments say the argument is the citation line, and `likenessTileLabel`'s says it differs from `likenessLabel(_:)` (same wording, the Source title, the header thumb) by its argument; `Localizable.xcstrings` is updated with them. Also reused: `useEarliestLikeness`, `portraitBadge`, `useAsPortrait`, `useAsPortraitLabel(_:)`, the portrait Toasts, `galleryEmptyTitle`, `galleryEmptyMessage`. Any heading or caption wording the board adds is listed here on refresh.

## 7. Steps

**STEP-1 — The citation line on each tile**
Files: `macos/App/Features/Conclusions/CitationLineDisplay.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`ConclusionDetailPicture`, `ConclusionDetailGallery`, `tiles`, `pictures`, `ConclusionDetailThumbnail.leading`), `macos/App/Features/Conclusions/PersonDetailContent.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-swift-test`, `add-localized-string`.
Do: D-3, D-4, D-7 and the § Strings keys; the old section's caption, help and hint (D-9). Add T-1, T-4, T-5, T-6.
Check: `xcodebuild test` (as `scripts/check-pr.sh` runs it) passes `PersonDetailTests`.

**STEP-2 — The new section, drawn from the slot**
Files: `macos/App/Features/Conclusions/ConclusionGallerySection.swift` *(new)*, `macos/App/Features/Conclusions/PersonDetailView.swift`. Skill: `add-ui-component`.
Do: D-1, D-2, D-9, D-11: `GalleryTileView`'s new parameters (the old section passes today's values), the grid and tiles at the board's size, captions, the Portrait Badge and Set as portrait; the image's button opens the Citation (D-11). `PersonDetailSections` `.gallery` draws it. The Xcode project uses synchronized folders (`PBXFileSystemSynchronizedRootGroup`): no `project.pbxproj` edit.
Check: the app builds; the Person's Gallery draws the new tiles; an Event page's Pictures is unchanged.

**STEP-3 — The heading and accessibility**
Files: `ConclusionGallerySection.swift`, `L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-11's label and hint, GA-6 as the board decides, GA-9 ids.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; AC-3.1, AC-3.2 by hand; VoiceOver reads AC-1.2.

**STEP-4 — Use earliest likeness in the Gallery**
Files: `ConclusionGallerySection.swift`.
Do: D-6.
Check: AC-2.2, AC-2.3 by hand; T-3 passes.

**STEP-5 — Full width at Wide**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift` (`fillsWidth`), `macos/ProvenenciaTests/PersonDetailTests.swift` (the `fillsWidth` loop). Skill: `add-swift-test`.
Do: D-8; T-2.
Check: `PersonDetailTests` passes; at a 2296pt pane the grid fills the width.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (map row), this plan (status, Landed note), `docs/macos-client-patterns.md` only if it names the Person gallery's view.
Do: per the Definition of done. The brief stays in `design/` for S13-15.
Check: `python3 scripts/check-plan-docs.py` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `galleryTilesCarryTheCitationParts` *(new)* | `D.picture("a", path:, title: "Hamburger Adressbuch 1861")` with its `record.artifactLabel = "Volume 3"`, `record.citationLabel = "Entry 214"`, `record.pageLabel = "23"` set on the returned value → `citation.title == "Hamburger Adressbuch 1861"`, `citation.line == "Hamburger Adressbuch 1861, Volume 3, Entry 214, p. 23"`, `titleIsRef == false`, `accessibilityLabel == likenessTileLabel(citation.line)`; the same picture with `leads = true` → `likenessTileLabelLead(citation.line)`; title "" and no parts → `title` and `line` are the Observation ref, `titleIsRef == true` | AC-1.2, D-3, D-4 |
| T-2 | Swift | `PersonDetailTests` the `fillsWidth` loop (updated) | `fillsWidth == (section == .timeline \|\| section == .gallery)` | AC-4.1, D-8 |
| T-3 | Swift | `PersonDetailTests` `portraitFlagsFollowThePreferredLikeness` (one case added) | One picture that leads → `offersEarliest == false` (the two-picture cases are already there) | AC-2.3, D-6 |
| T-4 | Swift | `PersonDetailTests` `likenessesBecomeTilesEarliestFirst` (updated) | Its `accessibilityLabel` expectations follow D-3, and `\.sourceTitle` reads `\.citation.title`; and `L10n.Conclusions.likenessTileLabel("X") == "Likeness from X"` (pins the wording) | AC-1.2 |
| T-5 | Swift | `PersonDetailTests` `eventPicturesKeepTheirWording` *(new)* | `EventDetailContent` with one leading picture → `accessibilityLabel == "Picture from Register, thumbnail"`; a second, not leading → `"Picture from Register 2"`; `gallery.fallbackColors == nil` | AC-1.4, D-4, D-7 |
| T-6 | Swift | `PersonDetailTests` `headerThumbIsTheLeadLikeness` (updated) | The thumb's `colors` come from `gallery.fallbackColors` | D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonDetailTests` (labels, the citation line cases at line 671), `ReconciledValueDisplayTests` (shared fixture).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief stays in design/ (S13-15 archives it); design/README.md unchanged
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Person who has at least three likenesses from different Citations, one with no Source title, one whose image file is missing, and one chosen as the portrait that is not the earliest. Open the Person → Gallery. See the board's tiles in the engine's order, the cut caption, the mono ref caption, the fallback mark in the person's colors, the Portrait Badge on the chosen tile and the Use earliest likeness control (AC-1.1, AC-2.2).
2. Set another likeness as the portrait: the header thumb changes, the Toast reads as AC-2.1, focus stays; while it saves every Set as portrait and Use earliest is disabled (AC-2.1, AC-2.4). Use earliest likeness: the first tile leads and the control goes (AC-2.2, AC-2.3).
3. Click a tile's image, and press Return on another: the composer opens at the likeness, and Back returns to the Gallery; no tile has a provenance control, an Open citation control or a context menu (AC-3.1, AC-3.2).
4. Turn on VoiceOver: each tile reads its citation line; the portrait adds "portrait" (AC-1.2).
5. Open a Person with no likeness: the EmptyState (AC-1.3).
6. Stretch the window past 1600pt with a tile focused: the grid fills the width, tiles keep their size; shrink back: the section, scroll and focus stay (AC-4.1, AC-4.2).
7. Open an Event and a Place with pictures: their Pictures sections are unchanged (AC-1.4).
8. Dark mode: frames 1, 3 and 5 match the board (GA-10).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The agreed tile is too far from `GalleryTileView` to share | Stop and report (D-9); a second tile is not written |
| Passing the citation line to `pictureTileLabel` changes Event and Place VoiceOver | Accepted (AC-1.4): it only adds the Artifact, Citation and page, and ", thumbnail" on the lead; T-5 pins it |

**Stop and report if:**

- The agreed board changes the header, its thumb menu, or the Event or Place Pictures (D-1; user, 2026-11-03).
- The board asks for data the detail read does not carry (a likeness date, a Subject name) — that is a follow-up engine PR (working rule "Engine PRs run ahead of their slice").
- The board's tile cannot be drawn through `GalleryTileView`'s parameters (D-9).
- The board asks for a tile above 256pt, which `ThumbnailCache.maxPixelSize` would render soft at 2× (D-9).
- `ConclusionDetailGallery.offersEarliest` or `ConclusionLeadModel` no longer do what § 4 says.

## 12. Open questions

The board settles these; the plan stays `Blocked — on board S13-D5` until they are answered on the **Board** line:

1. GA-1: the tile size (recommended 160pt).
2. GA-2: whether the caption has a second line with the page.
3. GA-3: where the Portrait Badge and Set as portrait sit.
4. GA-5, GA-6: where Use earliest likeness sits, and whether the heading line stays.
5. GA-8: any rewording of the empty state.

## 13. PR

**Title:** S13-14 — Gallery section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-14-gallery-section.md`.
