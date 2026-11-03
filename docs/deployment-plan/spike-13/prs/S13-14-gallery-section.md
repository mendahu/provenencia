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
**Context:** [person-detail.md § 4. Gallery](../person-detail.md#4-gallery) (R-29, R-31, R-32; R-1, R-2, R-13b); brief [S13-D5](../design/S13-D5-gallery-and-viewer.md) GA-1 – GA-10, GA-W1, GA-W2; the S14-02 plan ([`../../spike-14/prs/S14-02-pictures-in-the-app.md`](../../spike-14/prs/S14-02-pictures-in-the-app.md)) D-1 – D-5; working rules: "Sections are places", "Redesigns adopt Page columns", "One sectioned Conclusion page, not a Person fork"; Spike 14 working rules "The gallery layer is named gallery, not likeness", "An Event or Place chooses its thumbnail as a Person chooses its portrait"
**Planned:** 2026-11-03 against `main` @ `42eb321` (draft: everything that does not depend on the board's frames; § 12 lists what the board settles)
**Review:** [`S13-14-gallery-section.review.md`](S13-14-gallery-section.review.md)

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

A researcher opens a Person's Gallery section and sees every likeness of the person at a size where faces can be told apart, each captioned with its Source, with the portrait marked, Set as portrait on every other likeness, its own Open citation control, and a visible way back to the earliest likeness. At Wide the grid fills the pane. The slice's Check is bar 5 without the viewer (S13-15 adds the viewer).

**Why now.** Slice 5's first chrome PR. S13-10 moved today's likeness grid into the Gallery section unchanged and left its redesign, and where "Use earliest likeness" lives, to the S13-D5 board (S13-10 § 1 What stays). S14-02 made the gallery a kind-blind slot, so the Person's Gallery can be redrawn while the Event and Place Pictures keep theirs (map scope entry S13-14).

**What stays.** The header, its thumb and the thumb's right-click "Use earliest likeness" menu, shared with Event and Place (user, 2026-11-03). The gallery's order and lead rule (Go, S14-01). `ConclusionDetailGallerySection` and its tiles for the Event and Place pages' Pictures, drawn by `ConclusionDetailDefaultPageSections` (user, 2026-11-03: Person only). `ConclusionLeadModel` and its Toasts. A click on a tile's image still opens the Citation until S13-15 (brief § 3). No Go, proto or schema change.

## 2. User stories

**US-1.** As a researcher, when I open a Person's Gallery, I see every likeness large enough to recognise a face, with where each came from, so that I can compare them.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann Weber with the brief's seven likenesses (§ 5) | Seven tiles in the engine's order, each at the board's tile size (GA-1); the caption of tile 5 is cut at two lines; tile 6's caption "OBS-7Q2KD" is in mono; tile 7 shows the Person mark in Johann's gender colors |
| AC-1.2 | VoiceOver on tile 2 | Reads "Likeness from Hamburger Adressbuch 1861, Volume 3, Entry 214, p. 23" (the citation line, `CitationLineDisplay.line`); on tile 4 the same with ", portrait" |
| AC-1.3 | A person with no likeness | The EmptyState "No likenesses yet" with the photo glyph (GA-8) |
| AC-1.4 | The Event and Place pages | Their Pictures section looks and behaves as before (96pt tiles, click opens the Citation, no Open citation control, no Use earliest control) |

**US-2.** As a researcher, I set a likeness as the portrait, or go back to the earliest, from the Gallery.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Tile 2 → Set as portrait | The header thumb shows likeness 2; tile 2 carries the Portrait Badge; the Toast reads "Portrait updated" · "The likeness from Hamburger Adressbuch 1861 is now the portrait."; focus stays on tile 2 |
| AC-2.2 | Likeness 4 chosen, likeness 1 is the earliest | The Gallery shows the Use earliest likeness control (GA-5); activating it makes tile 1 lead, the control disappears, the Toast reads "Portrait updated" · "The earliest likeness is the portrait again." |
| AC-2.3 | No choice (tile 1 leads), or the chosen likeness is the earliest | No Use earliest likeness control in the Gallery |
| AC-2.4 | A save is in flight | Every Set as portrait and the Use earliest likeness control are disabled |

**US-3.** As a researcher, I open the Citation a likeness was cut from.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Tile 2 → Open citation (its control or its context menu) | The Citation composer opens at the likeness (`ConclusionDetailPicture.location`); Back returns to Johann's Gallery |
| AC-3.2 | A click on tile 2's image | Opens the Citation, as today (S13-15 moves it to the viewer) |

**US-4.** As a researcher on a wide monitor, the gallery uses the width.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Pane 2296pt, 40 likenesses | The grid fills the pane between the gutters with more tiles per row; tiles keep their size; the header keeps its band (GA-W1) |
| AC-4.2 | Resize across 1600 with tile 12 focused | The section, scroll position and focus stay (GA-W2) |

## 3. Scope

| In | Covered by |
| --- | --- |
| The Person's Gallery section from the board: tile, caption, Portrait Badge, Set as portrait (GA-1 – GA-3, GA-7) | STEP-2, STEP-3, AC-1.1, AC-2.1 |
| Open citation as its own control and in the tile's context menu (GA-4) | STEP-3, AC-3.1 |
| The visible Use earliest likeness control (GA-5) | STEP-4, AC-2.2, AC-2.3 |
| The heading line as the board decides (GA-6) | STEP-3 |
| Empty, refreshing, saving states (GA-8; brief § 2) | STEP-3, AC-1.3, AC-2.4 |
| The citation line on each gallery tile, for the tooltip and VoiceOver (GA-2, GA-9) | STEP-1, AC-1.2 |
| Keyboard and VoiceOver (GA-9), dark mode (GA-10) | STEP-3, manual check |
| Full width at Wide (GA-W1, GA-W2) | STEP-5, AC-4.1, AC-4.2 |

| Out | Home |
| --- | --- |
| The likeness viewer; the click on a tile opening it | S13-15 |
| The new grid or Open citation on the Event and Place Pictures | not planned this spike (user, 2026-11-03); the Event and Place redesign (R-1) |
| A change to the header thumb or its right-click menu | not planned (user, 2026-11-03) |
| Delete, reorder or caption editing of a likeness | not planned (brief § 2.1) |
| The provenance control on a likeness | not planned (R-13b) |
| Grouping or filtering the gallery | not planned (brief § 2.1) |
| Retiring `ConclusionDetailGallerySection` | the Event and Place redesign, when they adopt this section (D-2) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Person Gallery case | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` `.gallery` | `PVEmptyState` when `content.gallery.tiles` is empty, else `ConclusionDetailGallerySection(projectDir:pageIdentifier:gallery:lead:)` — the same view the Event and Place pages draw |
| Shared gallery section | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailGallerySection` | `PVSectionHeader(title: gallery.title, meta: count)`; adaptive `LazyVGrid` of 120pt columns; each tile a `GalleryTileView` whose crop is a `Button` → `navigation.go(to: tile.location)`, then the lead `PVBadge` or a ghost `PVButton(gallery.choose.useAs)` → `lead.choose(tile, copy:)`, disabled while `lead.isSaving`; focus returns to the tile after a choice |
| Tile view | `ConclusionDetailPage.swift` `GalleryTileView` (private) | `CachedThumbnail(size: CachedThumbnail.pictureCrop /* 96 */, fallbackMark:)` with no gender colors, a two-line muted caption of `sourceTitle` |
| Default fill | `ConclusionDetailPage.swift` `ConclusionDetailDefaultPageSections` | Event and Place: rows, then `ConclusionDetailGallerySection` when tiles exist, then relationship sections |
| Gallery slot | `ConclusionDetailPage.swift` `ConclusionDetailGallery` (`title`, `openTip`, `fallbackMark`, `tiles`, `choose: LeadCopy`, `lead`, `offersEarliest`), `ConclusionDetailGallery.tiles(from:tileLabel:useAsLabel:chosenBody:)`, `ConclusionDetailPicture` | Each tile carries `id`, `imageRelPath`, `sourceTitle`, `location`, `isLead`, `accessibilityLabel` (from `tileLabel(sourceTitle)`), `useAsLabel`, `chosenToastBody`. No citation line |
| Person gallery | `macos/App/Features/Conclusions/PersonDetailContent.swift` `init` | Builds the gallery with `L10n.Conclusions.likenessesSection`, `likenessOpenTip`, `.subjectPerson`, `likenessTileLabel`, `useAsPortraitLabel`, `toastPortraitChosenBody`, `.personPortrait`; `treatment` holds the gender treatment (`thumbnailColors`) |
| Lead model | `macos/App/Features/Conclusions/ConclusionLeadModel.swift` | `choose(_:copy:)`, `useEarliest(copy:)`, `isSaving`, `toast`; owned by `ConclusionDetailPage`, passed in `ConclusionDetailPageSectionsContext.lead` |
| Header menu | `ConclusionDetailPage.swift` `ConclusionDetailHeader.thumbnail` | `pvRightClickMenu` with `copy.useEarliest` while `content.gallery.offersEarliest` |
| Citation line | `macos/App/Features/Conclusions/CitationLineDisplay.swift` `CitationLineDisplay.line(_:)` | "Source, Artifact, Citation, page" from a `CatalogReconcilerOutcome` (S13-11) |
| Full width | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection.fillsWidth` | `true` for `.timeline` only; `ConclusionDetailPage.swift` passes `primaryFill: pageSection?.fillsWidth` to `PVPageColumns` |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Conclusions` | `likenessesSection` "Likenesses", `likenessOpenTip` "Open citation", `likenessTileLabel(_:)` "Likeness from %@", `portraitBadge`, `useAsPortrait`, `useAsPortraitLabel(_:)`, `useEarliestLikeness` "Use earliest likeness", `toastPortraitUpdatedTitle`, `toastPortraitChosenBody(_:)`, `toastPortraitEarliestBody`, `galleryEmptyTitle`, `galleryEmptyMessage` |
| Context menu | `macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift`, `PVActionMenu.swift` (`pvRightClickMenu`, `PVActionMenuItem`) | The kit's right-click menu, as the header thumb uses |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` `likenessesBecomeTilesEarliestFirst`, `portraitFlagsFollowThePreferredLikeness`, `headerThumbIsTheLeadLikeness`, the `fillsWidth` loop (line 386); `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift` fixture `picture(_:path:title:leads:)` | The gallery content and the lead flags; `fillsWidth == (section == .timeline)` |

**Precedent.** S13-13 ([#691](https://github.com/mendahu/provenencia/pull/691)) replaced an S13-10 placeholder case in `PersonDetailSections` with a section view of its own and opted it into `fillsWidth`; copy that shape (a new file in `Features/Conclusions/`, one case line in `PersonDetailSections`, one `fillsWidth` case, the test loop updated). S14-02 ([#641](https://github.com/mendahu/provenencia/pull/641)) made the gallery kind-blind (`ConclusionDetailGallery`, `LeadCopy`): the new section reads only that slot and the lead model, as `ConclusionDetailGallerySection` does. Do not copy S13-13's own read and host: the gallery comes with the page's detail read.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The Person page only. `PersonDetailSections` `.gallery` draws the new section; `ConclusionDetailDefaultPageSections` keeps drawing `ConclusionDetailGallerySection` for Event and Place | User, 2026-11-03; R-1 (Event and Place keep their look) | All three kinds — the Event and Place pages would change look without a board |
| D-2 | The new section is `ConclusionGallerySection` *(new)* in `macos/App/Features/Conclusions/ConclusionGallerySection.swift` *(new)*, kind-blind: it takes `projectDir`, `pageIdentifier`, `gallery: ConclusionDetailGallery`, `lead: ConclusionLeadModel` and `fallbackColors: PVThumbnail.Colors?`, and speaks only the slot's words. `ConclusionDetailGallerySection` stays unchanged for Event and Place; both are known debt that the Event and Place redesign removes by adopting `ConclusionGallerySection` | The map scope entry: "the old and new galleries can coexist and retiring the old one does not drop the Event and Place galleries"; kind-blind so the second caller adopts it with no fork (working rule "One sectioned Conclusion page, not a Person fork") | A `style:` parameter on `ConclusionDetailGallerySection` — every line of its body would branch on it; a Person-named view — Event and Place could not adopt it |
| D-3 | Each `ConclusionDetailPicture` gains `citationLine: String` *(new)*, set in `ConclusionDetailGallery.tiles(from:…)` from `CitationLineDisplay.line($0.record)`, and `titleIsRef: Bool` *(new)*, true when `$0.record.sourceTitle` is empty (so `sourceTitle` is the Observation ref, `ReconciliationRecord.sourceTitle(for:)`). The `tileLabel` closure receives the citation line instead of the Source title, for every kind | One wording of a citation for the page (S13-11's line); computed once where the tiles are built; Event and Place speak the longer label too, which only adds the Artifact, Citation and page | A second formatter in the section — a near-copy of `CitationLineDisplay` |
| D-4 | The portrait's spoken label adds ", portrait" through `L10n.Conclusions.likenessTileLabelLead(_:)` *(new)*, chosen in the section by `tile.isLead` | GA-9 | A Badge read by VoiceOver separately — it splits one tile into two stops |
| D-5 | Open citation is `navigation.go(to: tile.location)`, the same call the tile's image makes today, from the tile's own control and from its context menu (`pvRightClickMenu` with two `PVActionMenuItem`s: Set as portrait, unless the tile leads, and Open citation) | R-32; one navigation for both routes; the kit menu the header uses | A second navigation path or a new menu component |
| D-6 | The Gallery's Use earliest likeness control calls `lead.useEarliest(copy: gallery.choose)` and shows while `gallery.offersEarliest`; its title is `gallery.choose.useEarliest` | The header menu's rule and copy, so the two controls never disagree (user, 2026-11-03: keep the header menu) | A Gallery-only rule — two rules for one state |
| D-7 | The tile's fallback mark wears the person's gender colors: `PersonDetailSections` passes `content.treatment?.thumbnailColors` as `fallbackColors` | Brief § 2 ("the Person mark on the sunken Thumbnail, in the gender colors"), as the header thumb does | No colors — the tile and the header thumb would disagree |
| D-8 | `ConclusionPageSection.gallery.fillsWidth` is `true` | GA-W1 (rule 3: grids fill the width), the Timeline precedent | A secondary column — the section has nothing supporting to put in it |
| D-9 | The tile size is a `static let` on `ConclusionGallerySection` (the board's GA-1 value). `CachedThumbnail.pictureCrop` stays 96 for Event and Place. `ThumbnailCache.maxPixelSize` (512) is not changed: it covers a tile up to 256pt at 2× | One constant per view; the cache cap is the header thumb's | Changing `pictureCrop` — it would resize the Event and Place tiles |
| D-10 | No new query, read, cache key or navigation place. The gallery stays on `GetConclusionDetail` (working rule "Each section reads its own query": Gallery is part of today's detail read) | The rule names Gallery as part of `GetConclusionDetail` | A gallery read of its own — no data is missing |

The board settles GA-1's size, GA-2's second line, GA-3's and GA-4's placement, GA-5's place and GA-6's heading; § 12 lists them. They are layout inside `ConclusionGallerySection` and change no Decision above.

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionGallerySection` *(new)*, `ConclusionGallerySection.swift` *(new)* | view, file | `ConclusionDetailPage.swift` `ConclusionDetailGallerySection` | D-2: the Event and Place pages keep the old one unchanged; the new one is kind-blind so they adopt it later |
| The new section's tile view (private, in the new file) | view | `GalleryTileView` (private) | Depends on the board: if the agreed tile is `GalleryTileView` with a larger size and colors, the implementer moves `GalleryTileView` to the new file as `internal`, adds `size` and `colors` parameters, and both sections use it; otherwise a private tile in the new file. The board's answer fixes which (§ 12) |
| `ConclusionDetailPicture.citationLine` *(new)*, `.titleIsRef` *(new)* | fields | `CitationLineDisplay.line`, `ReconciliationRecord.sourceTitle(for:)` | Reuse them; the fields carry their results so the view decides nothing about records |
| `L10n.Conclusions.likenessTileLabelLead(_:)` *(new)*, `likenessOpenLabel(_:)` *(new)* | keys | `likenessTileLabel(_:)`, `useAsPortraitLabel(_:)` | New wordings; same shape and namespace |

**Builds on.** The S14-02 gallery slot and lead model; `CitationLineDisplay`; `pvRightClickMenu`; Page columns' full-width mode (S13-13).
**Deletes.** In `PersonDetailSections`, the Person's use of `ConclusionDetailGallerySection`. Nothing else becomes dead: Event and Place still call it.
**Next adds.** S13-15 adds the viewer to `ConclusionGallerySection` (the image's click opens it). The Event and Place redesign passes their gallery to `ConclusionGallerySection` and deletes `ConclusionDetailGallerySection` and `GalleryTileView`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Open citation from the context menu | `WorkspaceNavigation.go(to:)` with the tile's composer location | The same location the tile's image already opens; a Citation deleted since the read is the composer's existing missing-referent path (`fallbackToSectionRoot`) |
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
| Gallery section | `ConclusionGallerySection` *(new)* | `Features/Conclusions/ConclusionGallerySection.swift` | D-2. Heading line per GA-6 (`PVSectionHeader` with `meta` count and the D-6 control as its trailing content, if the board keeps it) |
| Tile | private tile view or the moved `GalleryTileView` | same file | D-9 size; `CachedThumbnail` with `fallbackMark: gallery.fallbackMark`, `colors: fallbackColors`; caption `PVTypeScale.caption`, muted, two lines, `.help(tile.citationLine)`; the caption in `PVFont.mono` when `tile.titleIsRef` (D-3) |
| Portrait mark | `PVBadge(gallery.choose.badge, tone: .accent, icon: .photo)` | — | As today |
| Set as portrait | `PVButton(gallery.choose.useAs, variant: .ghost, size: .sm)` | — | As today, disabled while `lead.isSaving`, focus back to the tile after a choice |
| Open citation | `PVIconButton(.quote, label: …)` or `PVButton`, per the board (GA-4) | — | D-5; tooltip `gallery.openTip` |
| Context menu | `pvRightClickMenu` + `PVActionMenuItem` | — | D-5 |
| Use earliest likeness | `PVButton(gallery.choose.useEarliest, variant: .ghost, size: .sm)` | — | D-6, GA-5 |
| Empty | `PVEmptyState` | `PersonDetailView.swift` | Unchanged |

States: empty (EmptyState), saving (buttons disabled), refreshing (the switcher row's indicator, unchanged), image missing (fallback mark in colors). Accessibility: each tile is one `.accessibilityElement(children: .contain)` labelled per D-3/D-4; the image's button is the open action (Citation now, viewer in S13-15); Set as portrait and Open citation are buttons labelled with the Source; ids `"\(pageIdentifier).picture.\(id)"`, `.leadBadge`, `.useAsLead` (kept), `.openCitation` *(new)*, `"\(pageIdentifier).gallery.useEarliest"` *(new)*.

### Strings

L10n keys *(new)* in `L10n.Conclusions`:
- `likenessTileLabelLead(_:)` = "Likeness from %@, portrait" (`conclusions.person.likenessTile.lead`)
- `likenessOpenLabel(_:)` = "Open the citation of the likeness from %@" (`conclusions.person.likenessOpen.label`)

Changed meaning: `likenessTileLabel(_:)` and `pictureTileLabel(_:)` now receive the citation line (D-3); their comments say so. Reused: `likenessOpenTip`, `useEarliestLikeness`, `portraitBadge`, `useAsPortrait`, `useAsPortraitLabel(_:)`, the portrait Toasts, `galleryEmptyTitle`, `galleryEmptyMessage`. Any heading or caption wording the board adds is listed here on refresh.

## 7. Steps

**STEP-1 — The citation line on each tile**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`ConclusionDetailPicture`, `ConclusionDetailGallery.tiles`), `macos/App/Platform/L10n.swift` (comments of `likenessTileLabel`, `pictureTileLabel`), `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-swift-test`, `add-localized-string`.
Do: D-3. Add T-1.
Check: `xcodebuild test` (as `scripts/check-pr.sh` runs it) passes `PersonDetailTests`.

**STEP-2 — The new section, drawn from the slot**
Files: `macos/App/Features/Conclusions/ConclusionGallerySection.swift` *(new)*, `macos/App/Features/Conclusions/PersonDetailView.swift`. Skill: `add-ui-component`.
Do: D-1, D-2, D-7, D-9: the grid and tiles at the board's size, captions, the image's button opening the Citation, the Portrait Badge and Set as portrait. `PersonDetailSections` `.gallery` draws it. The Xcode project uses synchronized folders (`PBXFileSystemSynchronizedRootGroup`): no `project.pbxproj` edit.
Check: the app builds; the Person's Gallery draws the new tiles; an Event page's Pictures is unchanged.

**STEP-3 — Open citation, the context menu, the heading, accessibility**
Files: `ConclusionGallerySection.swift`, `L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-4, D-5, GA-6 as the board decides, GA-9 labels and ids.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; VoiceOver reads AC-1.2.

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
| T-1 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `galleryTilesCarryTheCitationLine` *(new)* | A picture whose record has Source "Hamburger Adressbuch 1861", Artifact "Volume 3", Citation "Entry 214", page label "23" → `citationLine == "Hamburger Adressbuch 1861, Volume 3, Entry 214, p. 23"`, `titleIsRef == false` and `accessibilityLabel == likenessTileLabel(that line)`; a record with no title and no parts → `citationLine` and `sourceTitle` are the Observation ref, `titleIsRef == true` | AC-1.2, D-3 |
| T-2 | Swift | `PersonDetailTests` the `fillsWidth` loop (updated) | `fillsWidth == (section == .timeline \|\| section == .gallery)` | AC-4.1, D-8 |
| T-3 | Swift | `PersonDetailTests` `portraitFlagsFollowThePreferredLikeness` (extended) | One picture → `offersEarliest == false`; chosen = earliest → false; chosen ≠ earliest → true | AC-2.2, AC-2.3, D-6 |
| T-4 | Swift | `PersonDetailTests` `likenessesBecomeTilesEarliestFirst` (updated) | Its `accessibilityLabel` expectations follow D-3 | AC-1.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonDetailTests` (labels), `ReconciledValueDisplayTests` (shared fixture), any Event or Place test that pins `pictureTileLabel`.

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
3. Open citation from a tile's control and from its right-click menu: the composer opens at the likeness; Back returns to the Gallery. Click a tile's image: the Citation opens as before (AC-3.1, AC-3.2).
4. Turn on VoiceOver: each tile reads its citation line; the portrait adds "portrait" (AC-1.2).
5. Open a Person with no likeness: the EmptyState (AC-1.3).
6. Stretch the window past 1600pt with a tile focused: the grid fills the width, tiles keep their size; shrink back: the section, scroll and focus stay (AC-4.1, AC-4.2).
7. Open an Event and a Place with pictures: their Pictures sections are unchanged (AC-1.4).
8. Dark mode: frames 1, 3 and 5 match the board (GA-10).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The agreed tile is too far from `GalleryTileView` to share | § 6 Codebase fit: a private tile in the new file; the old one stays for Event and Place |
| Passing the citation line to `pictureTileLabel` changes Event and Place VoiceOver | D-3 accepts it; it only adds the Artifact, Citation and page |

**Stop and report if:**

- The agreed board changes the header, its thumb menu, or the Event or Place Pictures (D-1; user, 2026-11-03).
- The board asks for data the detail read does not carry (a likeness date, a Subject name) — that is a follow-up engine PR (working rule "Engine PRs run ahead of their slice").
- The board asks for a tile above 256pt, which `ThumbnailCache.maxPixelSize` would render soft at 2× (D-9).
- `ConclusionDetailGallery.offersEarliest` or `ConclusionLeadModel` no longer do what § 4 says.

## 12. Open questions

The board settles these; the plan stays `Blocked — on board S13-D5` until they are answered on the **Board** line:

1. GA-1: the tile size (recommended 160pt) and whether `GalleryTileView` is shared (§ 6 Codebase fit).
2. GA-2: whether the caption has a second line with the page.
3. GA-3, GA-4: where the Portrait Badge, Set as portrait and Open citation sit; Open citation as an IconButton or a Button.
4. GA-5, GA-6: where Use earliest likeness sits, and whether the heading line stays.
5. GA-8: any rewording of the empty state.

## 13. PR

**Title:** S13-14 — Gallery section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-14-gallery-section.md`.
