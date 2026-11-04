# S13-15 — Likeness viewer

**Status:** Landed in #703 (2026-11-03)
**Spike / slice:** Spike 13 · Slice 5 — Gallery
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D5](../design/archive/S13-D5-gallery-and-viewer.md), board drawn 2026-11-03 (turn 13) and agreed by the user the same day
**Board:** Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Gallery.dc.html*, turn 13, viewer frames 13f – 13k, 13o, 13p, 13t – 13x. The implementer may not be able to open it: D-6, D-9 and § 6 macOS give every measure
**Depends on:** S13-36 (landed [#702](https://github.com/mendahu/provenencia/pull/702): the kit provenance recipe), S13-14 (landed [#701](https://github.com/mendahu/provenencia/pull/701): `ConclusionGallerySection`, the tile's citation line, the tile's focus restore), S13-D5 (board)
**Blocks:** —
**Branch:** `s13-15-likeness-viewer`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go code or proto change; one Go test's path constant follows the moved Swift file (STEP-2)
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 4. Gallery](../person-detail.md#4-gallery) (R-30, R-31, R-32); brief [S13-D5](../design/archive/S13-D5-gallery-and-viewer.md) LV-1 – LV-9, LV-W1; the [S13-14](S13-14-gallery-section.md) and [S13-36](S13-36-provenance-sheet-in-the-kit.md) **Landed** notes; working rules: "The Person's gallery is redrawn on the Person page only", "Sections are places" (the viewer is not one), "One sectioned Conclusion page, not a Person fork"
**Planned:** 2026-11-03; refreshed after S13-14 and S13-36 landed, against `main` @ `4c245957`, every § 4 row re-read in the code
**Review:** [`S13-15-likeness-viewer.review.md`](S13-15-likeness-viewer.review.md) — reviewed twice before the board; refreshed after S13-14 and S13-36 and reviewed twice more: Fits with changes / Ready with changes, every finding Fixed, Declined with evidence, or Logged

**Landed:** the code follows the plan; the notes below. (1) `ConclusionPictureImage.sized` and `.decoded` hold the file's pixel size, not the final frame (D-6), and the view calls `ConclusionPictureImage.frame` on each draw, so the image follows a width-class or pane-height change without a reload; `frame` and `bandHeight` are the pure functions the plan names. (2) `PVIconButton` takes the `ViewerCopy` resources directly (a `PVCopy`), with no `L10n.string`. (3) The § 4 row for `ArtifactImageDecode` names its new path, because `check-plan-docs.py` fails on a path the move deleted; the moved file's header comment also drops "Artifact viewer's". (4) The close handler clears `lead.toast` and `toastInViewer` for every close, and skips the focus restore when the gallery is empty (AC-4.3: nothing to focus). (5) The manual check (§ 10), the frames against the board, the `.refused` preview at both widths, and the stop conditions that need the running app (← and → inside the sheet, `FocusClaim` on opening, the focus return while the sheet dismisses, the sheet closing when the section is swapped for its EmptyState) were not run in the app; they are unverified. Fix round 1 (the first review) changed the following. (6) The decode is `BoundedImageDecode` (was `ArtifactImageDecode`), admitted by `ImageDecodeLimits` (types, edge, `maxBytes`) through one `ImageHeaderDecode.admit`; the viewer passes `ThumbnailDecodePolicy.limits`, whose 128 MiB byte bound is held to `raster.MaxDecodeBytes`; the Artifact viewer passes no byte bound. (7) Supersedes D-5's "steps more than 120 ms apart can overlap" and the matching Scale line: one decode is in flight (`decodeTask`, awaited before the next file opens; a `Task<Result<...>, Never>`), both detached tasks run at `.userInitiated`. (8) Moves the writer of D-10's and D-11's `toastInViewer` and Toast clearing to the viewer's own mount: `ConclusionLeadModel.viewerOpened()` and `viewerClosed()`; the rule is unchanged. (9) Supersedes D-8's "on the panel body" and `.ignored`: both arrow handlers sit on the panel and return `.handled` at the ends; a `@FocusState ViewerStop` covers the band and both step buttons, and the image reclaims focus after Use as portrait and when a focused step button is disabled or gone; this focus behavior is unverified in the app. (10) `ConclusionPictureImage` states carry the tile id and a mismatch draws as `.waiting`; the sheet falls back to the first tile for one draw when the shown one leaves. (11) The viewer label also speaks the portrait (`likenessViewerLabelLead`) and a refused image ("Image unavailable"); `spokenJoin(_ parts:)` drops empty parts for all callers; `PVPanelMetrics.bodyInset` (an enum beside `PVPanel`, since a generic type holds no stored static) names the panel padding. Fix round 2 changed the following. (12) `ImageHeaderDecode.admit(url:limits:)` is the one admission of a picture file by URL (size, then open, then type, header and fit); `BoundedImageDecode` and `ThumbnailCache` both call it, the cache with `ThumbnailDecodePolicy.limits`, and `ThumbnailDecodePolicy.accepts` is gone. The thumbnail cache now has the byte limit, so a file over 128 MiB shows the fallback on both the tile and the viewer. (13) `ConclusionLeadModel.viewerClosed()` does nothing unless the viewer is open, so leaving the Gallery section keeps a page Toast; the viewer's own Toast overlay is gated on `toastInViewer`, as the page's is. Fix round 3 changed the following (the user's manual check, 2026-11-04). (14) Supersedes D-5's `declaredSize` step and D-6's thumbnail placeholder: the viewer never shows the tile's thumbnail and `BoundedImageDecode.declaredSize` is gone; `ConclusionPictureImage` is `.waiting`, `.decoded`, `.refused`, and one `decode` gives the image and its `pixelSize`. `.waiting` draws an empty band, with a spinner only after 250 ms (`showsSpinner`). (15) Supersedes the Out line on prefetching, and Scale's "no cache growth": the viewer holds `@State decoded` for the shown tile and its previous and next (`ConclusionPictureStep.heldIDs`, at most three, the rest dropped on each step). A held likeness shows at once, with no 120 ms wait and no spinner; after the shown one is up, the same `.task(id:)` decodes next, then previous, one at a time through `decodeTask`, and a step cancels a prefetch not yet started. A refused neighbour stores nothing. The manual feel of stepping is unverified in the app.

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

A click on a likeness in a Person's Gallery opens it large in a viewer over the page. The researcher steps through the gallery with previous and next or the arrow keys, sees where each likeness came from, makes one the portrait, or opens its Citation, and closes the viewer back onto the tile. This makes bar 5 true: "the Gallery shows every likeness; one opens a viewer that steps through them; from the gallery or viewer a likeness can be set as the portrait or opened at its Citation."

**Why now.** The last PR of slice 5; R-30 asks for a built-in viewer, and the user chose stepping without autoplay (2026-11-03).

**What stays.** Everything S13-14 shipped except the tile image's click, which moves from opening the Citation to opening the viewer. Tiles have no provenance control, no Open citation and no context menu: the viewer's record row is where a likeness's provenance and its Citation link live (user, 2026-11-03). The header and its thumb menu. The Event and Place Pictures, which keep the `.small` look of `ConclusionGallerySection` (user, 2026-11-03: Person only). `ConclusionLeadModel` and its Toasts. No navigation place is added: the viewer is not in the history.

## 2. User stories

**US-1.** As a researcher, I open a likeness large and step through the gallery.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann's seven likenesses (brief § 5); click tile 2's image | The viewer opens on likeness 2: the image whole, fit, at most its pixel size ÷ 2 in points (300 × 380 pixels → at most 150 × 190pt); the title row "Likeness of Johann Weber", the counter "2 of 7" in mono, close; under the image the likeness's provenance record row: "Hamburger Adressbuch 1861, Volume 3, Entry 214, p. 23", its outcome, its refs and its Source, Citation and Observation links |
| AC-1.2 | In the viewer on 2, press → | Likeness 3 shows, counter "3 of 7", VoiceOver reads the new likeness |
| AC-1.3 | On 1 | Previous is disabled and ← does nothing; on 7, next is disabled and → does nothing (no wrap) |
| AC-1.4 | A gallery of one | No previous, next or counter |
| AC-1.5 | Likeness 4 (1400 × 2048 pixels) at Regular, then at Wide | The cap is 700 × 1024pt; the band's height less 40pt binds first: about 273 × 400pt at Regular, 622 × 910pt at Wide (LV-2, LV-W1) |
| AC-1.6 | Likeness 7 (missing image) | The Person mark in Johann's gender colors on the sunken surface; the record row and actions stay |
| AC-1.7 | Tab to a tile's image, press Return | The viewer opens on that likeness |

**US-2.** As a researcher, I act on the likeness I am viewing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Viewer on 2 → Use as portrait | The header thumb and the grid behind show likeness 2 as the portrait; the viewer stays on 2 and shows the Portrait Badge in place of the button; the portrait Toast shows over the viewer and is spoken once |
| AC-2.2 | Viewer on 2 → the record row's Citation link | The viewer closes and the Citation composer opens at the likeness; Back returns to Johann's Gallery with the viewer closed. The Source and Observation links do the same for their pages |
| AC-2.3 | Viewer on 4 (the portrait) | The Portrait Badge, no Use as portrait |

**US-3.** As a researcher, closing the viewer puts me back where I was.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Open on 2, step to 5, press Escape | The viewer closes; focus is on tile 5, scrolled into view |
| AC-3.2 | Click outside the viewer, or its close button | The same as AC-3.1 |
| AC-3.3 | Open the viewer, close it, press Back; then quit and reopen the project on Johann's Gallery | Back leaves the Person page (the viewer added no history step); the restored page shows the Gallery with the viewer closed |

**US-4.** As a researcher, the viewer survives the gallery changing under it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Viewer on 5 of 7; likeness 5 is deleted in another window | The viewer shows the likeness now at position 5 (the old 6), "5 of 6" |
| AC-4.2 | Viewer on 7 of 7; likeness 7 is deleted | The viewer shows the new last (the old 6), "6 of 6" |
| AC-4.3 | Viewer on the only likeness; it is deleted | The viewer closes; the Gallery shows its EmptyState |

## 3. Scope

| In | Covered by |
| --- | --- |
| The viewer in the board's form, opened from a tile's image (LV-1) | STEP-5, STEP-8, AC-1.1, AC-1.7 |
| The image, fit and capped, the loading and missing states (LV-2) | STEP-1, STEP-2, STEP-6, AC-1.1, AC-1.5, AC-1.6 |
| Stepping, the counter, the ends (LV-3) | STEP-1, STEP-5, AC-1.2 – AC-1.4 |
| The provenance record row under the image (LV-4; user, 2026-11-03) | STEP-4, STEP-7, AC-1.1, AC-2.2 |
| Use as portrait in the viewer (LV-5) | STEP-7, AC-2.1, AC-2.3 |
| Closing and focus return (LV-6) | STEP-8, AC-3.1 – AC-3.3 |
| A likeness leaving while shown (LV-7) | STEP-1, STEP-8, AC-4.1 – AC-4.3 |
| Keyboard and VoiceOver (LV-8), dark mode (LV-9), Wide (LV-W1) | STEP-5, STEP-6, STEP-8, manual check |

| Out | Home |
| --- | --- |
| Autoplay, a timer, transitions | not planned (user, 2026-11-03) |
| Zoom, pan, rotate, crop, redraw | not planned (brief § 2.1); the Artifact viewer owns zoom, S14-03 drawing |
| A viewer on the Event and Place Pictures | not planned this spike (user, 2026-11-03); the Event and Place redesign (R-1) |
| Use earliest likeness in the viewer | not planned (brief LV-5: GA-5 owns it, S13-14) |
| The viewer as a navigation place or restored state | not planned (D-7) |
| An overlay presenter (form B) | not planned; a new kit presenter, asked of the user if the board picks it (D-4) |
| Prefetching beyond the shown likeness's two neighbours | not planned (Landed 14: the neighbours' full images decode ahead, three held) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Gallery section | `macos/App/Features/Conclusions/ConclusionGallerySection.swift` `ConclusionGallerySection(look:projectDir:pageIdentifier:gallery:lead:)` | One section for every kind. `tileView(_:)` builds `let open = { navigation.go(to: tile.location) }` for the image's `Button` and `pvActivatesOnReturn(when: look.opensOnReturn && focused == .image(tile.id), open)`; hint `gallery.openTip`. Each tile has `.id(tile.id)`. `@FocusState focused: TileFocus?` (`.image(id)`, `.useAs(id)`), `lastFocused`, `pendingFocus`, `scrollTarget`, `scrollRequest`; `keepFocus()` sets `pendingFocus`, `requestScroll()` and `applyPending()` (through `FocusClaim`), and ends the pending stop after 400ms; `.pvScrollsRowIntoView(scrollTarget, animated: false, whenRequested: scrollRequest)` is on the grid |
| Tile data | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPicture` (`citation: CitationLineDisplay.Parts`, `location`), `ConclusionDetailGallery` (`openTip`, `fallbackColors`, `tiles`, `choose`), `tiles(from:…)` | Every per-tile string is built by kind in `tiles(from:…)` from `CatalogPicture.record`; the tile keeps no outcome. The slot's static words are `title`, `openTip`, `LeadCopy` |
| A likeness's record | `core/database/conclusiondetails/conclusiondetails.go` `loadProperties`, `loadOutcomes`; `macos/App/Platform/GenealogyStore.swift` `CatalogPicture.record` | A depiction Property has **no field**: `loadProperties` skips it, so `detail.fields` never holds a likeness and there is no Vitals likeness row. The picture's outcome reaches Swift only as `CatalogPicture.record`, and it is `kept` or `weak` (`autoreconcile.Reason.InGallery`), the likeness's earliest gallery record |
| Record assembler | `macos/App/DesignSystem/Recipes/ProvenanceSheet/PVProvenanceSheetContent.swift` `Record(outcome:in:locale:orders:)`, `Record.spoken`; `macos/App/Platform/ReconciledValueDisplay.swift` `counted`, `deniedSomething`, `outcomePhrase`, `outcomeMark` | Each takes a non-optional field. `kept` and `weak` read nothing from it. `spoken` joins `readAs` even when empty |
| Record row | `macos/App/DesignSystem/Recipes/ProvenanceSheet/PVProvenanceRecordRow.swift` `PVProvenanceRecordRow(record:face:columns:idPrefix:open:)`, `PVProvenanceColumns.swift` (`showsReadAs`, `widthClass`, `panelWidth`) | Placeable alone. With `showsReadAs` false the record's spoken line sits at the row's leading edge, and `face` is unused. `open` reports `(Record, PVProvenanceLink)`; `WorkspaceLocation.provenance(_:_:)` in `macos/App/Features/Workspace/WorkspaceLocation.swift` maps it |
| Lead model | `macos/App/Features/Conclusions/ConclusionLeadModel.swift` | `choose(_:copy:)` → `Bool`, `isSaving`, `toast` |
| Toast | `ConclusionDetailPage.swift` line 323 `.vocabularyToastOverlay($lead.toast, identifier:)` (`Features/CatalogVocabulary/VocabularyChrome.swift`) | Drawn on the page, so a sheet covers it |
| Sheet presenter | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` `pvSheet(item:…)` (wraps `sheet(item:)`: a changed `Identifiable.id` is a new sheet), `PVSheetChrome` (Escape and a click outside; no visible close) | One dismiss rule |
| Panel | `macos/App/DesignSystem/Components/Panel/PVPanel.swift` `PVPanel(title:subtitle:width:footerChrome:body:footer:titleAccessory:)` | The body every app sheet uses; a title is required; the title accessory holds a trailing control |
| Sheet precedent | `ConclusionDetailPage.swift` `ConclusionDetailRowsSection` (`provenanceItem`, `.pvSheet(item:)`, `.onChange(of: rows)`); `macos/App/Features/Conclusions/ProvenanceSheetItem.swift` `ProvenanceSheetItem` (a fixed id while open), `PVProvenanceColumns.panelWidth` (880 / 1280), a link = `sheetClose { open(record, link) }` in the kit's `PVProvenanceSheet` | An item that holds an id, resolved against the live rows on every draw |
| Go parity test | `api/ffi/swift_vocabulary_test.go` `swiftBoundedImageDecode` (line 61) | Reads `BoundedImageDecode.swift` by path to check `maxDecodePixels` against Go |
| Full-size decode | `macos/App/Platform/BoundedImageDecode.swift` (moved here by STEP-2, renamed by fix round 1) `BoundedImageDecode.decode(url:limits:) -> Result<Decoded, Refusal>` (`raster`, `image`, `pixelSize`; the sniffed type must be one of the identifiers; declared size and decoded pixels bounded); called by `ArtifactViewerModel` after `ProjectFiles.objectURL` | The bounded full-size decode of a project object; synchronous |
| Header read | `macos/App/Platform/ImageHeaderDecode.swift` `ImageHeaderDecode.declaredImage(_:)` → `DeclaredImage` (`width`, `height`, …) | Reads the header only; no pixels decode |
| Thumbnail cache | `macos/App/Platform/ThumbnailCache.swift` `cachedImage(projectDir:relPath:)`, `ThumbnailDecodePolicy.allowedTypes` (PNG, JPEG) | The tile's 512-pixel image, already decoded |
| Path rule | `.cursor/rules/macos-client.mdc` line 21 | A catalog `relPath` is read only through `ProjectFiles.objectURL` (or the cache); names who may |
| Scroll into view | `macos/App/DesignSystem/Components/ScrollRowIntoView/PVScrollRowIntoView.swift` `pvScrollsRowIntoView(_:anchor:animated:…whenRequested:)` | With a token, scrolls to the row when the token changes, on the next main-actor turn. The section already applies it to its grid |
| Announce | `macos/App/DesignSystem/Components/Announce/PVAnnounce.swift` `pvAnnounce(_:priority:)` | A modifier that posts its text when it changes |
| Position copy | `macos/App/Platform/L10n.swift` `L10n.DesignSystem.selectOptionPosition(current:count:)` "%1$lld of %2$lld" | Reused as a visible counter by `CitationComposerEvidenceColumn.citationPosition` |
| Icons | `macos/App/DesignSystem/Components/Icon/PVIcon.swift` `PVSymbol` `.chevronBack`, `.chevronForward`, `.dismiss` | For previous, next, close |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift`, `ProvenanceSheetTests.swift`, `ReconciledValueDisplayTests.swift` fixture `picture(_:path:title:leads:)` | Gallery content, records; no viewer |

**Precedent.** S13-11 ([#650](https://github.com/mendahu/provenencia/pull/650)) and S13-36 ([#702](https://github.com/mendahu/provenencia/pull/702)): the provenance sheet is an item with a fixed id, presented with `.pvSheet(item:)` around a `PVPanel`, resolved against the live rows on every draw, closed when its row leaves, sized by width class, and closing through `\.pvSheetClose` before it navigates. Copy that shape. From the Artifact viewer reuse only `ArtifactImageDecode`; do not copy its zoom, pan or region tools.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The viewer is `ConclusionPictureViewer` *(new)* in `macos/App/Features/Conclusions/ConclusionPictureViewer.swift` *(new)*, kind-blind: it takes `projectDir`, `pageIdentifier`, `gallery: ConclusionDetailGallery`, `viewer: ConclusionDetailGallery.ViewerCopy`, `@Bindable var lead: ConclusionLeadModel`, `step: ConclusionPictureStep` (built by the section, D-3), `widthClass`, `paneHeight` (from `\.pvPageViewportSize`), and two closures: `show(_ id: String)` and `open: (PVProvenanceSheetContent.Record, PVProvenanceLink) -> Void`. It closes through `@Environment(\.pvSheetClose)` (`PVSheetClose`), the kit's one close route, from its close button and from a record link. It carries `.accessibilityIdentifier("\(pageIdentifier).viewer")` on its panel. Only `ConclusionGallerySection` presents it, and it passes `open: { navigation.go(to: .provenance($0, $1)) }` | User, 2026-11-03 (Person only); kind-blind so Event and Place adopt it with the section; the viewer names no navigation type, as `PVProvenanceSheet` | A Person-named viewer — the second caller would fork |
| D-2 | Stepping is a pure value, `ConclusionPictureStep` *(new, in the same file)*: `init?(tiles: [ConclusionDetailPicture], showing id: String)`, nil when `id` is not in `tiles`; `tile`, `index`, `count`, `previousID`, `nextID` | LV-3 as one testable rule; position is computed from the shown id, not stored | Index math in the view; wrapping (brief § 2) |
| D-3 | `ConclusionGallerySection` holds `@State viewing: ConclusionPictureViewerItem?` *(new, in `ConclusionPictureViewer.swift`)*: `struct: Identifiable, Equatable { let id = "viewer"; var shownID: String; func resolved(from old: [ConclusionDetailPicture], in new: [ConclusionDetailPicture]) -> Self? }`. The `Identifiable.id` is constant, so stepping never re-presents the sheet. `resolved` keeps `shownID` when it is in `new`, else moves to `new[min(oldPosition, new.count − 1)]`, else nil. `tileView`'s `open` closure sets `viewing` to `{shownID: tile.id}` when `gallery.viewer` is non-nil, else navigates to `tile.location` as today; Return and the click share it, and Space comes from the `Button`. The sheet closure builds `ConclusionPictureStep(tiles: gallery.tiles, showing: item.shownID)` on each draw, `if let` as the precedent does for `row`; `show(_:)` sets `shownID`; `.onChange(of: gallery.tiles) { old, new in viewing = viewing?.resolved(from: old, in: new) }` | The `ProvenanceSheetItem` shape (X-1, M-10); AC-4.1, AC-4.2; `.onChange` hands over the old tiles | The picture id as the item's id — every step would dismiss and rebuild the sheet; a stored `lastIndex` — a second field that can drift |
| D-4 | The viewer is form (A): `.pvSheet(item: $viewing)` around a `PVPanel` whose `width` is `PVProvenanceColumns.panelWidth` (880pt Regular, 1280pt Wide), so the record row under the image keeps its columns (S13-D13 PSK-6; board 13f, 13t). Form (B), an overlay that covers the pane, is a new page-level presenter the kit does not have; if the board picks it, stop and ask | One presenter and one dismiss rule (`PVSheetChrome`); an overlay drawn by the section cannot cover the pane (M-7) | An overlay in `ConclusionGallerySection` |
| D-5 | The full image decodes with `ArtifactImageDecode.decode(url:typeIdentifiers: Array(ThumbnailDecodePolicy.allowedTypes))`. STEP-2 moves `ArtifactImageDecode.swift` from `Features/ArtifactViewer/` to `macos/App/Platform/` beside `ImageHeaderDecode`, and gives both paths one admission: a private `admit(url:typeIdentifiers:maxEdge:) -> Result<(CGImageSource, ImageHeaderDecode.DeclaredImage), Refusal>` holds the type sniff, header read and `fits` check `decode` runs today, and `ImageHeaderDecode.DeclaredImage` gains computed `isTurned` and `uprightSize` *(new)*, for EXIF orientations 5 – 8. `decode` is `admit` plus the raster, reads both, and gains `maxEdge: Int = maxEdge`, so the Artifact viewer's call and result do not change. `declaredSize(url:typeIdentifiers:maxEdge:) -> Result<(width: Int, height: Int), Refusal>` *(new)* is `admit` plus `uprightSize`. The viewer passes `ThumbnailDecodePolicy.maxDeclaredEdge`, so a likeness the tile refuses is refused here too. The viewer's loader runs in `.task(id: step.tile.id)`, and its first line sets `image = .waiting` (D-6): (1) `try await Task.sleep(for: .milliseconds(120))`, so a held arrow key cancels each superseded load before any file is opened; (2) `ProjectFiles.objectURL` on the main actor; (3) `Task.detached` runs `ArtifactImageDecode.declaredSize(url:typeIdentifiers:)` *(new)*; back on the main actor `image = .sized(frame)` (D-6); (3b) a second `Task.detached` runs `decode`; (4) after it returns, a cancelled task discards the result. At most one result is shown; a decode already started runs to its end (bounded by `maxDeclaredEdge`) and is dropped, so steps more than 120 ms apart can overlap. Not cached | One full-size decode mechanism (M-2), bounded by its own pixel budget; the sleep is the cancellation point a detached task lacks (X-3, X-4) | Extracting a new decode out of `ThumbnailCache` — a third mechanism beside `ArtifactImageDecode`; a larger `ThumbnailCache.maxPixelSize` |
| D-6 | One `@State image: ConclusionPictureImage` *(new, in the same file)*: `.waiting`, `.sized(CGSize)`, `.decoded(NSImage, CGSize)`, `.refused`. The image band is 440pt tall at Regular and 950pt at Wide (the board's LV-2), less what a short pane lacks: `ConclusionPictureImage.bandHeight(widthClass:paneHeight:)` *(new, pure)* is `min(board height, max(320, paneHeight − 320))`, since Wide is a width and the panel's chrome is about 280pt. The band takes the panel body's full width inside `PVPanel`'s padding, on `PVColor.surfaceSunken` with a `PVColor.borderSubtle` 1pt border and `PVRadius.sm`; its middle column (the width less two 56pt step columns) holds the image with 20pt of padding above and below. `ConclusionPictureImage.frame(pixelWidth:pixelHeight:widthClass:bandHeight:) -> CGSize` *(new, pure)* is `min(pixels ÷ 2, available)` with the aspect kept, on `declaredSize`'s turned edges (the cap, LV-2); available is `panelWidth − 2 × PVSpacing.space7 − 112` wide (728pt, 1128pt) and `bandHeight − 40` tall. `.waiting` (a load starting, on opening and on each step, so the last likeness never lingers) draws `ThumbnailCache.shared.cachedImage(…)` `.resizable().scaledToFit()` at most 256pt square, else a small `ProgressView`, centered. `.sized` draws that image in the final frame. `.decoded` draws the decoded image in the same frame. `.refused` draws `CachedThumbnail` at 160pt with `gallery.fallbackMark` and `fallbackColors`; a nil `objectURL` or any `Refusal`, including a `decode` refusal after `.sized`, sets it | No jump between the placeholder and the image (X-2); LV-2's cap as one tested function; no combination of optionals to reset | Sizing by the 512-pixel thumbnail — small crops would jump on load |
| D-7 | The viewer is not a `WorkspaceLocation` and is not persisted; a record link closes it before navigating | Brief § 2; S13-11's sheet is not a place either | A history step |
| D-8 | Focus: the whole image band is `.focusable()` with `@FocusState imageFocused` and `.onAppear { FocusClaim.claim($imageFocused, true) }`, once, as the composer sheets place opening focus; `.onKeyPress(.leftArrow)` / `.onKeyPress(.rightArrow)` on the panel body call `show(previousID / nextID)` and return `.ignored` at the ends; the previous and next `PVIconButton`s are disabled at the ends; the image area keeps focus across steps (its view identity does not change). The image area's label is `tile.viewerLabel`, and `.pvAnnounce(that label)` speaks it on each step (the label changes with the index, so no repeat is lost) | LV-3, LV-8 (X-11) | A wrap; a hand-posted `AccessibilityNotification` |
| D-9 | `PVPanel` title: `Text(verbatim: viewer.title)` ("Likeness of Johann Weber", the person's title; board 13f); a long title wraps, as every panel's does. Title accessory, an `HStack(spacing: PVSpacing.space4)` in order: `Spacer(minLength: 0)`; the counter `L10n.DesignSystem.selectOptionPosition(current: index + 1, count:)` in `PVFont.mono` at body-small size, secondary ink, hidden from VoiceOver (the image area's label says it), shown only for two or more likenesses; then the close `PVIconButton(.dismiss, label: L10n.string(viewer.close), size: .md)` running `sheetClose()`. Footer (`PVPanel`'s sunken footer): Use as portrait as a **primary** `PVButton` (disabled while `lead.isSaving`), or the Portrait Badge (`PVBadge(gallery.choose.badge, tone: .accent, icon: .photo)`) on the portrait. Body, top to bottom with 16pt between: the image band (D-6), its previous and next `PVIconButton(.chevronBack / .chevronForward, size: .lg)` centered in the band's two 56pt side columns (absent for one likeness); then the likeness's provenance record between a `PVColor.borderDefault` rule above and a `PVColor.borderSubtle` rule below: the kit's `PVProvenanceRecordRow` for `PVProvenanceSheetContent.Record(outcome: step.tile.outcome, in: nil)`, built on each draw, with `PVProvenanceColumns(showsReadAs: false, widthClass: widthClass)`, `idPrefix: "\(pageIdentifier).viewer.record"` and `open: { record, link in sheetClose { open(record, link) } }`. A likeness has no field (§ 4), so `Record(outcome:in:)` takes `CatalogConclusionField?`. `counted`, `deniedSomething`, `outcomePhrase` and `outcomeMark` in `ReconciledValueDisplay` take it too; with nil, nothing is denied and `.folded`, `.denied` and `.against` read their plain phrases (`outcomeFolded`, `outcomeDenied`, `outcomeAgainst`). `readAs` is "" and `readAsIsPhrase` false with nil. STEP-4 makes `spoken` filter empty parts before `spokenJoin`; no field's record has one (its `readAs` is at least "—"), so `aRecordSpeaksAsOneLine` holds. `ConclusionDetailPicture` holds `outcome: CatalogReconcilerOutcome` *(new)*, and `id` and `location` become computed from it. No Open citation button (user, 2026-11-03: the viewer shows the citation itself) | LV-4, LV-5, LV-6 (X-5); board 13f – 13k; one position wording (M-8); the record reads and links exactly as on the sheet; one rulebook for a record with or without a field | The Source title as the panel's title (the record row carries it); a third "N of M" key; a caption of its own beside the record row; a synthetic `CatalogConclusionField` for the likeness — a field the engine never sent; a depiction field from Go — a wire and Vitals change this PR has no story for |
| D-10 | Use as portrait in the viewer calls `lead.choose(step.tile, copy: gallery.choose)`; the viewer stays on the same id (D-3). One Toast has one host: `ConclusionLeadModel` gains `toastInViewer: Bool` *(new)*. The section sets it to `new != nil` in `.onChange(of: viewing)`, and false in its own `.onDisappear`, which covers the gallery emptying. The viewer's panel carries `.vocabularyToastOverlay($lead.toast, identifier: "\(pageIdentifier).viewer.toast")`, and the page's overlay at `ConclusionDetailPage.swift:323` takes `lead.toastInViewer ? .constant(nil) : $lead.toast`. A close clears `lead.toast`: moving it to the page would mount a second `PVToast` and speak it twice, and the header thumb already shows the choice. A record link runs `sheetClose { … }` (`PVSheetClose.callAsFunction(then:)`), as `PVProvenanceSheet` does | AC-2.1, AC-2.2 (X-10); the page's overlay sits under the sheet; `PVToast` announces and times itself, so two mounts speak twice | Navigating with the sheet open; a second overlay with no gate — one Toast spoken twice |
| D-11 | Every close (Escape, a click outside, the close button, a record link) sets `viewing` nil through the kit, so the focus return runs in `.onChange(of: viewing) { old, new in }` when `old` is set and `new` is nil, which also sets `lead.toastInViewer`. It reuses the section's restore: `keepFocus()` becomes `restoreFocus(to stop: TileFocus)`, and the width change calls it with `focused ?? lastFocused`; a close calls it with `.image(old.shownID)`. No second `pvScrollsRowIntoView` and no second `.id`. When the gallery empties (AC-4.3), `PersonDetailSections` swaps the section for the EmptyState, which removes the section and closes the sheet; nothing restores focus | LV-6, AC-3.1 (X-9); one restore path; the kit's `FocusClaim` writes again a turn later for a sheet still dismissing | A second scroll token and a second focus write beside `keepFocus()` |
| D-12 | The viewer's static words live on the slot: `ConclusionDetailGallery.viewer: ViewerCopy?` *(new)*, a `struct: Equatable` of a finished `title: String` and `LocalizedStringResource`s `previous`, `next`, `close`, as `LeadCopy`. The viewer's spoken line is per tile, so it is built with the other per-tile strings: `ConclusionDetailPicture.viewerLabel: String` *(new)*, set in `tiles(from:…)` through a `viewerLabel: (Int, Int, String) -> String = { _, _, line in line }` parameter (position, count, citation line), so `pictures(from:…)` needs no edit; the Person passes `L10n.Conclusions.likenessViewerLabel(index:count:line:)`. `PersonDetailContent` sets `.likeness(title: L10n.Conclusions.likenessViewerTitle(title.text))` (*new static func*, taking the finished title); Event and Place leave it nil. The section opens the viewer only when it is non-nil, else the image keeps opening the Citation. The Person's `openTip` becomes `likenessViewHint`, so the image's hint says what it does on the one slot field that already holds it; `likenessOpenTip` is deleted | Kind-blind words on the slot, as `LeadCopy` (M-1); no copy left naming an action the image no longer does (M-9); no second hint field | Person keys inside the viewer; a `viewHint` on `ViewerCopy` beside `openTip` |

The board settles LV-1's sizes, LV-3's button placement and LV-5's order; § 12 lists them. They change no Decision above.

**Rules this plan bends.** None. `ProjectFiles.objectURL` is still how a `relPath` is read, so `.cursor/rules/macos-client.mdc` needs no edit.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionPictureViewer` *(new)*, `ConclusionPictureViewer.swift` *(new)* | view, file | `PVProvenanceSheet`; `ArtifactViewer` | Follows `PVProvenanceSheet`'s shape (`pvSheet` + `PVPanel`, `\.pvSheetClose`); the Artifact viewer zooms and draws regions (brief non-goals) |
| `ConclusionPictureStep` *(new)* | value | `ArtifactViewerModel.goToPreviousPage`, `CitationComposerEvidenceColumn` `stepper` | A page field on one model and a picker's control; neither is a value over a list |
| `ConclusionPictureImage` *(new)* | value | `CitationComposerModel.Phase` | Same shape, this viewer's states |
| `ConclusionPictureViewerItem` *(new)* | value | `ProvenanceSheetItem` | Same shape; a different payload |
| `ConclusionDetailGallery.ViewerCopy` *(new)*, `.viewer` *(new)* | type, field | `ConclusionDetailGallery.LeadCopy` | Same pattern: static words of the slot per kind |
| `ArtifactImageDecode` moved to `Platform/` | file move | — | Reused, not copied (D-5) |
| `ConclusionDetailPicture.outcome`, `viewerLabel` *(new)* | fields | `citation`, `accessibilityLabel` | The record builds when the viewer draws it; the label is one more per-tile string |
| `ConclusionLeadModel.toastInViewer` *(new)* | field | `toast` | One host for one Toast (D-10) |
| `DeclaredImage.isTurned`, `uprightSize`, `admit` *(new)* | properties, func | the `turned` test and guards in `decode` | Moved, so `decode` and `declaredSize` share them (D-5) |
| `Record(outcome:in:)` with an optional field | signature | `ReconciledValueDisplay`'s four field-reading functions | Widened, not forked (D-9) |
| `restoreFocus(to:)` | func | `ConclusionGallerySection.keepFocus()` | The same function, taking its stop (D-11) |
| L10n keys (§ Strings) | keys | `L10n.Conclusions` likeness keys; `selectOptionPosition` reused | New wordings only |

**Builds on.** The S13-11 sheet pattern; `pvSheet`, `PVPanel`; `ArtifactImageDecode`, `ImageHeaderDecode`, `ThumbnailDecodePolicy.allowedTypes`; `CitationLineDisplay.Parts` (S13-14); `ConclusionLeadModel`; the section's focus restore and `pvScrollsRowIntoView`; `PVProvenanceRecordRow`, `WorkspaceLocation.provenance`; `pvAnnounce`.
**Deletes.** `L10n.Conclusions.likenessOpenTip` and its catalog entry (D-12); `keepFocus()`'s name (D-11). On the Person page the tile image no longer navigates; `tile.location` stays for Event and Place.
**Next adds.** The Event and Place redesign gets the viewer by adopting `.large`, setting a `ViewerCopy` (`.picture`) and `openTip`, and passing `viewerLabel`; until then their nil `viewer` keeps opening the Citation. A zoom, if wanted, replaces the image area inside `ConclusionPictureViewer`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| The viewer on a likeness deleted meanwhile | `ConclusionPictureViewerItem.resolved` | D-3: replaced by its neighbour; T-2 |
| A step to another likeness | `ConclusionPictureViewer`'s loader | `.task(id:)` cancels; `image = .waiting` first, so the last likeness never shows under the new title (D-6) |
| A record link from the viewer | `WorkspaceNavigation.go(to:)` | `WorkspaceLocation.provenance`, the sheet's own mapping (Source, Citation, Observation); the viewer closes first (D-10) |
| A full-size decode of a likeness file | `ArtifactImageDecode.decode` | Through `ProjectFiles.objectURL` (path rule); refused type or size → fallback mark (D-6) |

### Scale

At the brief's sizes (40 likenesses at most): one decode per likeness that stays shown for 120 ms — ≤ 2048 × 2048 (16 MB) for crops the app wrote (`ArtifactRegionCrop.maxLongEdge`, `core/ingest/stage.go` `stageMaxEdge`), ≤ 4096 × 4096 (`ThumbnailDecodePolicy.maxDeclaredEdge`, 64 MB) for a crafted file —, released when the viewer moves on; one is shown, and a superseded one finishes and is dropped (D-5). Landed 15 holds the shown likeness's two neighbours as well: three decodes at most, a gallery-size bound. No write is added.

### Edges

- **Assumptions.** Tile ids are unique (Observation ids) and keep the engine's order across a refresh; `ConclusionPictureStep` relies on both. T-1, T-2.
- **Neighbors.** First, last, one, two likenesses; the shown one removed at the start, middle and end; all removed (T-2).
- **Removed inputs.** A removed likeness is never shown again: the step resolves from the live tiles (D-3).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** The file: its type is sniffed and its declared edges are held to `ThumbnailDecodePolicy.maxDeclaredEdge` before any pixel decodes.

### Schema

None — no catalog change.

### Go core

None —.

### FFI

None —.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Viewer | `ConclusionPictureViewer` *(new)* in `pvSheet` + `PVPanel` | `Features/Conclusions/ConclusionPictureViewer.swift` | D-1, D-4, D-9 |
| Image area | `Image(nsImage:)` `.resizable().scaledToFit()`; fallback `CachedThumbnail` with the mark | same | D-5, D-6 |
| Previous, next, close | `PVIconButton(.chevronBack / .chevronForward / .dismiss, label:)` | same | D-8, D-9 |
| Counter | `Text` in `PVPanel`'s title accessory | same | D-9 |
| Provenance record | `PVProvenanceRecordRow` | same | D-9 |
| Use as portrait / Portrait Badge | `PVButton(gallery.choose.useAs)` / `PVBadge(gallery.choose.badge, tone: .accent, icon: .photo)` | `PVPanel` footer | D-10; disabled while `lead.isSaving` |
| Toast | `vocabularyToastOverlay` | on the panel | D-10 |

States: loading (the tile's image in the final frame), missing (fallback mark), saving (Use as portrait disabled). Ids `"\(pageIdentifier).viewer"`, `.viewer.previous`, `.viewer.next`, `.viewer.close`, `.viewer.useAsLead`, `.viewer.toast`; the record row's under `.viewer.record`.

### Strings

L10n keys *(new)* in `L10n.Conclusions`, used by `ConclusionDetailGallery.ViewerCopy.likeness`:
- `likenessViewerTitle(_:)` = "Likeness of %@" (`conclusions.person.likenessViewer.title`; the argument is the person's title)
- `likenessViewerLabel(index:count:line:)` = "Likeness %1$lld of %2$lld, from %3$@" (`conclusions.person.likenessViewer.label`; `line` is the citation line)
- `likenessViewerPrevious` = "Previous likeness", `likenessViewerNext` = "Next likeness", `likenessViewerClose` = "Close the viewer"
- `likenessViewHint` = "Opens the likeness in the viewer", the Person's `openTip`

Reused: `L10n.DesignSystem.selectOptionPosition(current:count:)` (the counter), `useAsPortrait`, `portraitBadge`, the portrait Toasts. Removed: `likenessOpenTip` (D-12). `likenessLabel` keeps S13-14's wording. The `.large` look has no tooltip on the image, so only the hint changes.

## 7. Steps

**STEP-1 — The step rule**
Files: `macos/App/Features/Conclusions/ConclusionPictureViewer.swift` *(new)*, `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`ConclusionDetailPicture`), `macos/ProvenenciaTests/ConclusionPictureStepTests.swift` *(new)*. Skill: `add-swift-test`.
Do: `ConclusionDetailPicture.outcome` with computed `id` and `location` (D-9); D-2, D-3's `ConclusionPictureViewerItem`, and `ConclusionPictureImage.frame` (D-6), values only, no view. The tests build tiles through `ConclusionDetailGallery.tiles(from:…)` on a detail of `ReconciledValueDisplayTests.picture`. T-1, T-2, T-7.
Check: `xcodebuild test` passes `ConclusionPictureStepTests`.

**STEP-2 — The full-size decode moves to Platform**
Files: `macos/App/Features/ArtifactViewer/ArtifactImageDecode.swift` → `macos/App/Platform/ArtifactImageDecode.swift` *(new)* (`git mv`), `macos/App/Platform/ImageHeaderDecode.swift`, `api/ffi/swift_vocabulary_test.go` (the `swiftArtifactImageDecode` path constant, line 61), `macos/ProvenenciaTests/ArtifactViewerModelTests.swift`, `macos/ProvenenciaTests/ArtifactFileFixtures.swift`. Skill: `add-swift-test`.
Do: D-5's move, `admit`, `uprightSize`, `declaredSize` (T-5). Both files' header comments stop naming only the Artifact viewer. The landed S14-13 plan keeps its old path.
Check: the app builds; `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes (`TestSwiftVocabularyMatchesGo` reads the moved file); every existing Artifact viewer test passes unchanged.

**STEP-3 — The viewer's copy on the slot**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`ConclusionDetailGallery.ViewerCopy`, `viewer`, `tiles(from:…)`), `macos/App/Features/Conclusions/PersonDetailContent.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-localized-string`.
Do: D-12; T-4.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; T-4 passes.

**STEP-4 — A record with no field**
Files: `macos/App/Platform/ReconciledValueDisplay.swift`, `macos/App/DesignSystem/Recipes/ProvenanceSheet/PVProvenanceSheetContent.swift`, `macos/ProvenenciaTests/ProvenanceSheetTests.swift`. Skill: `add-swift-test`.
Do: D-9's optional field; T-6.
Check: T-6 and the existing `ProvenanceSheetTests` pass.

**STEP-5 — The viewer's panel and stepping**
Files: `macos/App/Features/Conclusions/ConclusionPictureViewer.swift` *(new)*, from STEP-1. Skill: `add-ui-component`.
Do: D-1, D-4, D-8, D-9's title row and band, with the band drawing `.refused`.
Check: the app builds; a preview on `.refused` shows the panel at both widths.

**STEP-6 — The image**
Files: `macos/App/Features/Conclusions/ConclusionPictureViewer.swift` *(new)*.
Do: D-5's loader, D-6.
Check: the app builds; AC-1.1, AC-1.5, AC-1.6 are checked in STEP-8.

**STEP-7 — The record row and the footer**
Files: `macos/App/Features/Conclusions/ConclusionPictureViewer.swift` *(new)*, `macos/App/Features/Conclusions/ConclusionLeadModel.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (the page's overlay). Skill: `add-ui-component`.
Do: D-9's record row and footer, D-10.
Check: the app builds; `PersonDetailTests` pass.

**STEP-8 — Opening and closing from the Gallery**
Files: `macos/App/Features/Conclusions/ConclusionGallerySection.swift`.
Do: D-3, D-4, D-7, D-11. The width change keeps working through `restoreFocus(to:)`. The section's header comment, `GalleryTileLook.opensOnReturn`'s and `GalleryTileView`'s stop saying the image opens the Citation for every kind.
Check: AC-1.1, AC-1.5 – AC-1.7, AC-3.1 – AC-3.3, AC-4.1 – AC-4.3 by hand, at both widths and in dark.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (map row), this plan (status, Landed note), `docs/deployment-plan/spike-13/design/S13-D5-gallery-and-viewer.md` → `design/archive/`, `docs/deployment-plan/spike-13/design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/ConclusionPictureStepTests.swift` `stepsThroughTheGalleryWithoutWrapping` *(new)* | Tiles a…g showing b → index 1, count 7, previous a, next c; showing a → no previous; showing g → no next; one tile → neither; an id not in the tiles → nil | AC-1.2, AC-1.3, AC-1.4, D-2, D-8 |
| T-2 | Swift | `ConclusionPictureStepTests` `followsTheGalleryWhenTheShownLikenessLeaves` *(new)* | Item on e, a…g to a…g without e → item on f; item on g, g removed → item on f; item on the only tile, removed → nil; tiles unchanged → equal to the input, with the same `id` | AC-4.1, AC-4.2, AC-4.3, D-3 |
| T-5 | Swift | `macos/ProvenenciaTests/ArtifactViewerModelTests.swift` `declaredSizeTurnsRotatedImages` *(new)*, with `ArtifactFileFixtures` | A 300 × 200 PNG → (300, 200); a JPEG 300 × 200 with EXIF orientation 6 → (200, 300), equal to `decode`'s `pixelSize`; a GIF → `.unreadable`; a 5000 × 10 PNG with `maxEdge` 4096 → `.tooLarge` | D-5, D-6 |
| T-6 | Swift | `macos/ProvenenciaTests/ProvenanceSheetTests.swift` `aGalleryRecordNeedsNoField` *(new)* | A `kept` outcome, `in: nil` → `counted` true, phrase "kept", `.check`, `spoken` starting at the citation line; a `weak` one with a low-trust Source → `counted` false, the weak phrase and `.signalLow` | AC-1.1, D-9 |
| T-7 | Swift | `ConclusionPictureStepTests` `framesCapAtHalfThePixelsAndFitTheBand` *(new)* | 300 × 380 → 150 × 190 at both widths; 1400 × 2048 → height 400 at Regular, 910 at Wide, with the aspect kept; 4000 × 1000 → 728 × 182, 1128 × 282; `bandHeight` at Wide on a 1000pt pane → 680, on a 1440pt pane → 950, on a 500pt pane → 320 | AC-1.1, AC-1.5, D-6 |
| T-4 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `personGalleryOffersTheViewer` *(new)* | `PersonDetailContent` for a Person with seven pictures → `gallery.viewer?.title` is "Likeness of " plus the person's title, the second of seven tiles' `viewerLabel` is "Likeness 2 of 7, from " plus its citation line, `gallery.openTip` is `likenessViewHint`; `EventDetailContent`, `PlaceDetailContent` → `gallery.viewer == nil` | AC-1.1, D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestSwiftVocabularyMatchesGo` and the Artifact viewer tests (STEP-2's move), `PersonDetailTests`.

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
- [ ] Brief moved to design/archive/ and design/README.md updated (Scheduled → Archived, built by S13-14 and S13-15)
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Use the S13-14 manual check's Person (at least three likenesses, one with a missing image). Open the Gallery and click the second tile's image: the viewer opens on it with the counter, the provenance record under the image and Use as portrait; a small likeness is not blown up (AC-1.1). Tab to a tile's image and press Return: the same (AC-1.7).
2. Press → and ←, and the buttons: it steps; at the ends the buttons are disabled and the keys do nothing (AC-1.2, AC-1.3). On a person with one likeness there is no stepping (AC-1.4). On the missing image: the fallback mark in the person's colors (AC-1.6).
3. Use as portrait in the viewer: the header and the grid behind change, the viewer stays, the Badge replaces the button; on the portrait itself there is only the Badge (AC-2.1, AC-2.3). The record row's Citation link: the viewer closes, the composer opens; Back returns to the Gallery, viewer closed (AC-2.2).
4. Step to another likeness and press Escape; then reopen and click outside; then reopen and press close: each time focus lands on the last shown tile, scrolled into view (AC-3.1, AC-3.2). Close the viewer and press Back: it leaves the Person page; reopen the project: the Gallery shows with no viewer (AC-3.3).
5. After step 3, close the viewer within five seconds: no Toast follows on the page. With the viewer open in one window, delete the shown likeness's depiction in a second window on the same project: the viewer moves to the neighbour; with the viewer on the last of several, delete it: it shows the new last; on the only one, delete it: it closes; add a likeness back and Use as portrait shows its Toast on the page (AC-4.1, AC-4.2, AC-4.3).
6. On a Wide but short window (a 1080p display), the record row and footer stay on screen. Stretch past 1600pt with the viewer open: the large likeness grows, the small one stays capped (AC-1.5). Dark mode: the viewer matches the board's dark frames (LV-9). VoiceOver reads each likeness as it steps (LV-8).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The sheet's focus or arrow keys are taken by the presenter | Stop condition below |
| The decode takes long on a 2048-pixel PNG | The tile's image shows in the final frame meanwhile (D-6) |
| The Toast sits over the counter and close button for its 5 seconds | Accepted; Escape and a click outside still close |

**Stop and report if:**

- A close does not return focus to the tile through `restoreFocus(to:)` while the sheet is dismissing (D-11): give the sequence you saw.
- The sheet stays open when `PersonDetailSections` swaps the section for its EmptyState (D-11, AC-4.3).
- `.onKeyPress` inside `pvSheet` does not receive ← and → with focus on the image area, or `FocusClaim` does not place it there on opening (D-8), or Escape does not close through `PVSheetChrome`.
- The board picks form (B), or asks for zoom, autoplay or a viewer on Event or Place pages (D-4; user, 2026-11-03).
- Moving `ArtifactImageDecode` breaks an Artifact viewer test or `TestSwiftVocabularyMatchesGo` after the path constant is updated (STEP-2).

## 12. Open questions

None. The board's answers are in D-4, D-6, D-9 and D-11.

## 13. PR

**Title:** S13-15 — Likeness viewer
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-15-likeness-viewer.md`.
