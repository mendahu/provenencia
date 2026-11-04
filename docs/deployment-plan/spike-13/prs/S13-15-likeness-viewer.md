# S13-15 — Likeness viewer

**Status:** Blocked — on S13-36 only (S13-14 landed in #701) ([planned](S13-36-provenance-sheet-in-the-kit.md); board S13-D13 drawn, its Record row and ClipboardButton are in the kit) (the provenance sheet joins the design system; brief S13-D13). Once S13-36 lands, D-9's record table becomes one kit `PVProvenanceRecordRow` built with `PVProvenanceSheetContent.Record(outcome:in:)` from the picture's outcome and the likeness field that holds it in `detail.fields` (S13-14 no longer builds a per-tile row, so this plan adds that lookup in `ConclusionDetailGallery.tiles(from:…)`; [`../provenance-sheet.md`](../provenance-sheet.md) § 5.1), and the plan's `ProvenanceSheet` and `ReconciliationRecordRows` references are refreshed to the kit recipe's names (working rule "The provenance sheet is a kit recipe")
**Spike / slice:** Spike 13 · Slice 5 — Gallery
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D5](../design/S13-D5-gallery-and-viewer.md), board drawn 2026-11-03 (turn 13) and agreed by the user the same day
**Board:** Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Gallery.dc.html*, turn 13, viewer frames 13f – 13k, 13o, 13p, 13t – 13x. The implementer may not be able to open it: **Design → macOS** describes the viewer in words and gives every measure
**Depends on:** S13-36 ([planned](S13-36-provenance-sheet-in-the-kit.md), not landed: the kit provenance sheet), S13-14 (landed [#701](https://github.com/mendahu/provenencia/pull/701): `ConclusionGallerySection`, the tile's citation line), S13-D5 (board)
**Blocks:** —
**Branch:** `s13-15-likeness-viewer`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go code or proto change; one Go test's path constant follows the moved Swift file (STEP-2)
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 4. Gallery](../person-detail.md#4-gallery) (R-30, R-31, R-32); brief [S13-D5](../design/S13-D5-gallery-and-viewer.md) LV-1 – LV-9, LV-W1; the [S13-14 plan](S13-14-gallery-section.md) and, once it lands, its **Landed** note; working rules: "The Person's gallery is redrawn on the Person page only", "Sections are places" (the viewer is not one), "One sectioned Conclusion page, not a Person fork"
**Planned:** 2026-11-03 against `main` @ `42eb321`, refreshed after the board against `fe93195` (draft: written before S13-14 lands; § 4 rows marked *(S13-14)* are that plan's design, and are re-read in the code on refresh)
**Review:** [`S13-15-likeness-viewer.review.md`](S13-15-likeness-viewer.review.md) — round 1 Fits with changes / Not ready, round 2 Fits with changes / Not ready, every finding Fixed; reviewed again on the board refresh

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
| AC-1.5 | Likeness 4 (1400 × 2048 pixels) at Regular, then at Wide | The image fits the image area's height (440pt at Regular, 950pt at Wide, less 40pt of padding), never above 700 × 1024pt (LV-2, LV-W1) |
| AC-1.6 | Likeness 7 (missing image) | The Person mark in Johann's gender colors on the sunken surface; the record row and actions stay |
| AC-1.7 | Tab to a tile's image, press Return | The viewer opens on that likeness |

**US-2.** As a researcher, I act on the likeness I am viewing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Viewer on 2 → Use as portrait | The header thumb and the grid behind show likeness 2 as the portrait; the viewer stays on 2 and shows the Portrait Badge in place of the button; the portrait Toast shows |
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
| AC-4.3 | Viewer on the only likeness; it is deleted | The viewer closes; the Gallery shows its EmptyState, and focus goes to the section |

## 3. Scope

| In | Covered by |
| --- | --- |
| The viewer in the board's form, opened from a tile's image (LV-1) | STEP-3, STEP-4, AC-1.1, AC-1.7 |
| The image, fit and capped, the loading and missing states (LV-2) | STEP-2, STEP-3, AC-1.1, AC-1.5, AC-1.6 |
| Stepping, the counter, the ends (LV-3) | STEP-1, STEP-3, AC-1.2 – AC-1.4 |
| The provenance record row under the image (LV-4; user, 2026-11-03) | STEP-4, AC-1.1, AC-2.2 |
| Use as portrait in the viewer (LV-5) | STEP-4, AC-2.1, AC-2.3 |
| Closing and focus return (LV-6) | STEP-4, AC-3.1 – AC-3.3 |
| A likeness leaving while shown (LV-7) | STEP-1, STEP-4, AC-4.1 – AC-4.3 |
| Keyboard and VoiceOver (LV-8), dark mode (LV-9), Wide (LV-W1) | STEP-3, STEP-5, manual check |

| Out | Home |
| --- | --- |
| Autoplay, a timer, transitions | not planned (user, 2026-11-03) |
| Zoom, pan, rotate, crop, redraw | not planned (brief § 2.1); the Artifact viewer owns zoom, S14-03 drawing |
| A viewer on the Event and Place Pictures | not planned this spike (user, 2026-11-03); the Event and Place redesign (R-1) |
| Use earliest likeness in the viewer | not planned (brief LV-5: GA-5 owns it, S13-14) |
| The viewer as a navigation place or restored state | not planned (D-7) |
| An overlay presenter (form B) | not planned; a new kit presenter, asked of the user if the board picks it (D-4) |
| Prefetching the neighbours' full images | not planned (D-5, D-6: the tile's image shows while one loads) |

## 4. Current state

Rows marked *(S13-14)* are S13-14's planned design; re-read them in the code once it lands.

| Area | Where | Today |
| --- | --- | --- |
| Gallery section *(S13-14)* | `macos/App/Features/Conclusions/ConclusionGallerySection.swift` *(new)* in S13-14 `ConclusionGallerySection` | Kind-blind; the tile image is a `Button` (`PVFocusRingButtonStyle`, `pvActivatesOnReturn(when:)`, hint `gallery.openTip`) → `navigation.go(to: tile.location)` (the Citation); no provenance control and no context menu; `@FocusState focusedID` |
| Tile data *(S13-14)* | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPicture` (`citation: CitationLineDisplay.Parts` with `title`, `titleIsRef`, `detail`, `line`), `ConclusionDetailGallery` (`fallbackColors`) | Every per-tile string is built by kind in `tiles(from:…)`; the slot's static words are `title`, `openTip`, `LeadCopy` |
| Record rows | `macos/App/Features/Conclusions/ReconciliationRecordRows.swift` `ReconciliationRecordRows(groups:style:provenance:open:widthClass:)`, `SheetHeadings` | The provenance sheet's record rows, with their Source, Citation and Observation links calling `open` |
| Lead model | `macos/App/Features/Conclusions/ConclusionLeadModel.swift` | `choose(_:copy:)` → `Bool`, `isSaving`, `toast` |
| Toast | `ConclusionDetailPage.swift` line 317 `.vocabularyToastOverlay($lead.toast, identifier:)` (`Features/CatalogVocabulary/VocabularyChrome.swift`) | Drawn on the page, so a sheet covers it |
| Sheet presenter | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` `pvSheet(item:…)` (wraps `sheet(item:)`: a changed `Identifiable.id` is a new sheet), `PVSheetChrome` (Escape and a click outside; no visible close) | One dismiss rule |
| Panel | `macos/App/DesignSystem/Components/Panel/PVPanel.swift` `PVPanel(title:subtitle:width:footerChrome:body:footer:titleAccessory:)` | The body every app sheet uses; a title is required; the title accessory holds a trailing control |
| Sheet precedent | `ConclusionDetailPage.swift` `ConclusionDetailRowsSection` (`provenanceItem`, `.pvSheet(item:)`, `.onChange(of: rows)`); `macos/App/Features/Conclusions/ProvenanceSheetItem.swift` `ProvenanceSheetItem` (a fixed id while open), `PVProvenanceColumns.panelWidth` (880 / 1280), a link = `sheetClose { open(record, link) }` in the kit's `PVProvenanceSheet` | An item that holds an id, resolved against the live rows on every draw |
| Go parity test | `api/ffi/swift_vocabulary_test.go` `swiftArtifactImageDecode` (line 61) | Reads `ArtifactImageDecode.swift` by path to check `maxDecodePixels` against Go |
| Full-size decode | `macos/App/Features/ArtifactViewer/ArtifactImageDecode.swift` `ArtifactImageDecode.decode(url:typeIdentifiers:) -> Result<Decoded, Refusal>` (`raster`, `image`, `pixelSize`; the sniffed type must be one of the identifiers; declared size and decoded pixels bounded); called by `ArtifactViewerModel` after `ProjectFiles.objectURL` | The bounded full-size decode of a project object; synchronous |
| Header read | `macos/App/Platform/ImageHeaderDecode.swift` `ImageHeaderDecode.declaredImage(_:)` → `DeclaredImage` (`width`, `height`, …) | Reads the header only; no pixels decode |
| Thumbnail cache | `macos/App/Platform/ThumbnailCache.swift` `cachedImage(projectDir:relPath:)`, `ThumbnailDecodePolicy.allowedTypes` (PNG, JPEG) | The tile's 512-pixel image, already decoded |
| Path rule | `.cursor/rules/macos-client.mdc` line 21 | A catalog `relPath` is read only through `ProjectFiles.objectURL` (or the cache); names who may |
| Scroll into view | `macos/App/DesignSystem/Components/ScrollRowIntoView/PVScrollRowIntoView.swift` `pvScrollRowIntoView(row:…request:)` | Scrolls when `row` changes, or when the `request` token changes |
| Announce | `macos/App/DesignSystem/Components/Announce/PVAnnounce.swift` `pvAnnounce(_:priority:)` | A modifier that posts its text when it changes |
| Position copy | `macos/App/Platform/L10n.swift` `L10n.DesignSystem.selectOptionPosition(current:count:)` "%1$lld of %2$lld" | Reused as a visible counter by `CitationComposerEvidenceColumn.citationPosition` |
| Icons | `macos/App/DesignSystem/Components/Icon/PVIcon.swift` `PVSymbol` `.chevronBack`, `.chevronForward`, `.dismiss` | For previous, next, close |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift`, `ReconciledValueDisplayTests.swift` fixture `picture(_:path:title:leads:)` | Gallery content; no viewer |

**Precedent.** S13-11 ([#650](https://github.com/mendahu/provenencia/pull/650)): the provenance sheet is an item with a fixed id, presented with `.pvSheet(item:)` around a `PVPanel`, resolved against the live rows on every draw, closed when its row leaves, sized by width class, and closing before it navigates. Copy that shape. From the Artifact viewer reuse only `ArtifactImageDecode`; do not copy its zoom, pan or region tools.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The viewer is `ConclusionPictureViewer` *(new)* in `macos/App/Features/Conclusions/ConclusionPictureViewer.swift` *(new)*, kind-blind: it takes `projectDir`, `pageIdentifier`, `gallery: ConclusionDetailGallery`, `viewer: ConclusionDetailGallery.ViewerCopy`, `lead: ConclusionLeadModel`, `step: ConclusionPictureStep` (built by the section, D-3), `widthClass` and one closure `show(_ id: String)`. It closes through `@Environment(\.pvSheetClose)` (`PVSheetClose`), the kit's one close route, from its close button and from a record link. Only `ConclusionGallerySection` presents it | User, 2026-11-03 (Person only); kind-blind so Event and Place adopt it with the section | A Person-named viewer — the second caller would fork |
| D-2 | Stepping is a pure value, `ConclusionPictureStep` *(new, in the same file)*: `init?(tiles: [ConclusionDetailPicture], showing id: String, lastIndex: Int)`; `tile`, `index`, `count`, `hasPrevious`, `hasNext`, `previousID`, `nextID`; when `id` is no longer in `tiles` it resolves to the tile at `min(lastIndex, count − 1)`; nil for an empty gallery | LV-3, LV-7 as one testable rule | Index math in the view; wrapping (brief § 2) |
| D-3 | `ConclusionGallerySection` holds `@State viewing: ConclusionPictureViewerItem?` *(new, in `ConclusionPictureViewer.swift`)*: `struct: Identifiable, Equatable { let id = "viewer"; var shownID: String; var lastIndex: Int; func resolved(in tiles: [ConclusionDetailPicture]) -> Self? }` — the `Identifiable.id` is constant, so stepping never re-presents the sheet; `resolved(in:)` returns the item moved to `ConclusionPictureStep`'s tile and index, or nil. The grid iterates `Array(gallery.tiles.enumerated())` (id `\.element.id`) so the image's button opens `{shownID: tile.id, lastIndex: index}`. The section builds `ConclusionPictureStep(tiles: gallery.tiles, showing: shownID, lastIndex:)` on each draw and passes it; `show(_:)` sets `shownID` and `lastIndex` together; `.onChange(of: gallery.tiles)` sets `viewing = viewing?.resolved(in: gallery.tiles)` | The S13-11 item with a fixed id (X-1, M-10); AC-4.1 – AC-4.3 | The picture id as the item's id — every step would dismiss and rebuild the sheet |
| D-4 | The viewer is form (A): `.pvSheet(item: $viewing)` around a `PVPanel` whose `width` is the provenance sheet's (880pt Regular, 1280pt Wide, by `PVWidthClass`), so the record row under the image keeps its columns (S13-D13 PSK-6; board 13f, 13t). Form (B), an overlay that covers the pane, is a new page-level presenter the kit does not have; if the board picks it, stop and ask | One presenter and one dismiss rule (`PVSheetChrome`); an overlay drawn by the section cannot cover the pane (M-7) | An overlay in `ConclusionGallerySection` |
| D-5 | The full image decodes with `ArtifactImageDecode.decode(url:typeIdentifiers: Array(ThumbnailDecodePolicy.allowedTypes))`. STEP-2 moves `ArtifactImageDecode.swift` from `Features/ArtifactViewer/` to `macos/App/Platform/` beside `ImageHeaderDecode`, and splits out `declaredSize(url:typeIdentifiers:) -> Result<(width: Int, height: Int), Refusal>` *(new)*: the type sniff, header read and size check `decode` runs today, with the edges swapped for EXIF orientations 5 – 8 as `decode`'s `pixelSize` is; `decode` keeps its signature and result (the Artifact viewer keeps calling it). The viewer's loader runs in `.task(id: step.tile.id)`: (1) `try await Task.sleep(for: .milliseconds(120))`, so a held arrow key cancels each superseded load before any file is opened; (2) `ProjectFiles.objectURL` on the main actor; (3) `Task.detached` runs `ArtifactImageDecode.declaredSize(url:typeIdentifiers:)` *(new)*; back on the main actor the frame is published (D-6); (3b) a second `Task.detached` runs `decode`; (4) after it returns, a cancelled task discards the result. At most one result is shown; a decode already started runs to its end (bounded by `maxDecodePixels`) and is dropped. Not cached | One full-size decode mechanism (M-2), bounded by its own pixel budget; the sleep is the cancellation point a detached task lacks (X-3, X-4) | Extracting a new decode out of `ThumbnailCache` — a third mechanism beside `ArtifactImageDecode`; a larger `ThumbnailCache.maxPixelSize` |
| D-6 | The image area's size is fixed before pixels arrive: from `declaredSize`'s (already turned) `width` × `height` ÷ 2 points (the cap, LV-2), fit inside the panel's image area: a band 440pt tall at Regular and 950pt at Wide (the board's LV-2), the panel's full width, on `PVColor.surfaceSunken` with a `PVColor.borderSubtle` 1pt border and `PVRadius.sm`; its middle column (the width less two 56pt step columns) holds the image with 20pt of padding above and below. While the decode runs the area draws `ThumbnailCache.shared.cachedImage(…)` (else a small `ProgressView`) `.resizable().scaledToFit()` in that frame; then the decoded image in the same frame, on the sunken surface (`PVColor.surfaceSunken`). A refused file or header shows the fallback mark (`gallery.fallbackMark`, `gallery.fallbackColors`) in a square area | No jump between the placeholder and the image (X-2); LV-2's cap | Sizing by the 512-pixel thumbnail — small crops would jump on load |
| D-7 | The viewer is not a `WorkspaceLocation` and is not persisted; a record link closes it before navigating | Brief § 2; S13-11's sheet is not a place either | A history step |
| D-8 | Focus: the image area is `.focusable()` with `@FocusState imageFocused` and `.defaultFocus($imageFocused, true)`; `.onKeyPress(.leftArrow)` / `.onKeyPress(.rightArrow)` on the panel body call `show(previousID / nextID)` and return `.ignored` at the ends; the previous and next `PVIconButton`s are disabled at the ends; the image area keeps focus across steps (its view identity does not change). The image area's label is `viewer.label(index: index + 1, count: count, line: citation.line)`, and `.pvAnnounce(that label)` speaks it on each step (the label changes with the index, so no repeat is lost) | LV-3, LV-8 (X-11) | A wrap; a hand-posted `AccessibilityNotification` |
| D-9 | `PVPanel` title: `Text(verbatim: viewer.title)` ("Likeness of Johann Weber", the person's title; board 13f). Title accessory, in order: the counter `L10n.DesignSystem.selectOptionPosition(current: index + 1, count:)` in `PVFont.mono` at body-small size, secondary ink, hidden from VoiceOver (the image area's label says it), shown only for two or more likenesses; then the close `PVIconButton(.dismiss, label: L10n.string(viewer.close), size: .md)` running `sheetClose()`. Footer (`PVPanel`'s sunken footer): Use as portrait as a **primary** `PVButton` (disabled while `lead.isSaving`), or the Portrait Badge (`PVBadge(gallery.choose.badge, tone: .accent, icon: .photo)`) on the portrait. Body, top to bottom with 16pt between: the image band (D-6), its previous and next `PVIconButton(.chevronBack / .chevronForward, size: .lg)` centered in the band's two 56pt side columns (absent for one likeness); then the likeness's provenance record between a `PVColor.borderDefault` rule above and a `PVColor.borderSubtle` rule below: the kit's `PVProvenanceRecordRow` (S13-36) for `tile.record`, a `PVProvenanceSheetContent.Record` *(new field on `ConclusionDetailPicture`)* that `ConclusionDetailGallery.tiles(from:…)` builds with `Record(outcome:in:)` from the picture's outcome and the `detail.fields` field holding it (nil when none), with `PVProvenanceColumns` showing no Read as column and the sheet's width class; drawn only when `tile.record` is non-nil. `open` is `{ location in sheetClose { navigation.go(to: location) } }`. No Open citation button (user, 2026-11-03: the viewer shows the citation itself) | LV-4, LV-5, LV-6 (X-5); board 13f – 13k; one position wording (M-8); the record reads and links exactly as on the sheet | The Source title as the panel's title (the record row carries it); a third "N of M" key; a caption of its own beside the record row |
| D-10 | Use as portrait in the viewer calls `lead.choose(step.tile, copy: gallery.choose)`; the viewer stays on the same id (D-3). The viewer's panel carries its own `.vocabularyToastOverlay($lead.toast, identifier: "\(pageIdentifier).viewer.toast")`, since the page's overlay sits under the sheet. A record link runs `sheetClose { navigation.go(to: location) }` (`PVSheetClose.callAsFunction(then:)`), the kit's form of `ProvenanceSheet.open(_:)`'s close-then-go | AC-2.1, AC-2.2 (X-10); the precedent | Navigating with the sheet open |
| D-11 | Every close (Escape, a click outside, the close button, a record link) sets `viewing` nil through the kit, so the focus return runs in `.onChange(of: viewing) { old, new in }` when `old` is set and `new` is nil: `focusedID = old.shownID` and `closeRequest += 1` (`@State closeRequest = 0`). The page's scroll view is `PVPageColumns`'s (the section has none), so the rows scroll it: each tile view inside the `LazyVGrid` carries `.id(tile.id)`, and the grid applies `.pvScrollsRowIntoView(focusedID, whenRequested: AnyHashable(closeRequest))`, which scrolls even when the row is the one it opened from. When the gallery empties while the viewer is open (AC-4.3), `focusedID` is nil and focus goes to the section's EmptyState | LV-6, AC-3.1 (X-9) | Scrolling on `row` alone — no scroll when the id repeats |
| D-12 | The viewer's static words live on the slot: `ConclusionDetailGallery.viewer: ViewerCopy?` *(new)*, a `struct: Equatable` of a `title: String` (built by the kind: `L10n.Conclusions.likenessViewerTitle(content.title.text)`, "Likeness of %@") and `LocalizedStringResource`s `previous`, `next`, `close`, `viewHint`, `labelFormat`, with a method `label(index: Int, count: Int, line: String) -> String` that returns `L10n.format(labelFormat, index, count, line)` (no stored closure, as `LeadCopy`); `PersonDetailContent` sets `.likeness(title:)` *(new static func)*; Event and Place leave it nil. The section opens the viewer only when it is non-nil, else the image keeps opening the Citation. With a viewer, the tile image's hint becomes `viewer.viewHint` | Kind-blind words on the slot, as `LeadCopy` (M-1); no copy left naming an action the image no longer does (M-9) | Person keys inside the viewer |

The board settles LV-1's sizes, LV-3's button placement and LV-5's order; § 12 lists them. They change no Decision above.

**Rules this plan bends.** None. STEP-2 updates `.cursor/rules/macos-client.mdc` line 21 to name `ArtifactImageDecode`'s new home; the rule itself is unchanged.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionPictureViewer` *(new)*, `ConclusionPictureViewer.swift` *(new)* | view, file | `ProvenanceSheet`; `ArtifactViewer` | Follows `ProvenanceSheet`'s shape (`pvSheet` + `PVPanel`); the Artifact viewer zooms and draws regions (brief non-goals) |
| `ConclusionPictureStep` *(new)* | value | none (searched `previousID`, `hasNext`, `index(after` in `Features/`) | No stepper exists |
| `ConclusionPictureViewerItem` *(new)* | value | `ProvenanceSheetItem` | Same shape; a different payload |
| `ConclusionDetailGallery.ViewerCopy` *(new)*, `.viewer` *(new)* | type, field | `ConclusionDetailGallery.LeadCopy` | Same pattern: static words of the slot per kind |
| `ArtifactImageDecode` moved to `Platform/` | file move | — | Reused, not copied (D-5) |
| L10n keys (§ Strings) | keys | `L10n.Conclusions` likeness keys; `selectOptionPosition` reused | New wordings only |

**Builds on.** The S13-11 sheet pattern; `pvSheet`, `PVPanel`; `ArtifactImageDecode`, `ImageHeaderDecode`, `ThumbnailDecodePolicy.allowedTypes`; `CitationLineDisplay.Parts` (S13-14); `ConclusionLeadModel`; `pvScrollRowIntoView`; `pvAnnounce`.
**Deletes.** On the Person page, the tile image's `navigation.go(to:)` and its "Open citation" hint (D-12); `likenessOpenTip` if nothing else reads it.
**Next adds.** S13-15 changes only the `.large` look's open action; the `.small` look (Event and Place) keeps opening the Citation. The Event and Place redesign gets the viewer by moving to the `.large` look and setting a `ViewerCopy` (`.picture`). A zoom, if wanted, replaces the image area inside `ConclusionPictureViewer`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| The viewer on a likeness deleted meanwhile | `ConclusionPictureStep` | D-2, D-3: replaced by its neighbour, closed when none; T-2 |
| A record link from the viewer | `WorkspaceNavigation.go(to:)` | The provenance sheet's own locations (Source, Citation, Observation); the viewer closes first (D-10) |
| A full-size decode of a likeness file | `ArtifactImageDecode.decode` | Through `ProjectFiles.objectURL` (path rule); refused type or size → fallback mark (D-6) |

### Scale

At the brief's sizes (40 likenesses at most): one decode per likeness that stays shown for 120 ms — ≤ 2048 × 2048 (16 MB) for crops the app wrote (`ArtifactRegionCrop.maxLongEdge`, `core/ingest/stage.go` `stageMaxEdge`), ≤ `ArtifactImageDecode.maxDecodePixels` (64 MP) for a crafted file —, released when the viewer moves on; at most one at a time per viewer (D-5). No cache growth. No write is added.

### Edges

- **Assumptions.** Tile ids are unique (Observation ids) and keep the engine's order across a refresh; `ConclusionPictureStep` relies on both. T-1, T-2.
- **Neighbors.** First, last, one, two likenesses; the shown one removed at the start, middle and end; all removed (T-2).
- **Removed inputs.** A removed likeness is never shown again: the step resolves from the live tiles (D-3).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** The file: its type is sniffed and its declared size and decoded pixels are bounded by `ArtifactImageDecode` before any pixel decodes.

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
| Counter | `PVPanel` subtitle | same | D-9 |
| Provenance record | `ReconciliationRecordRows` (reused) | same | D-9 |
| Use as portrait / Portrait Badge | `PVButton(gallery.choose.useAs)` / `PVBadge(gallery.choose.badge, tone: .accent, icon: .photo)` | `PVPanel` footer | D-10; disabled while `lead.isSaving` |
| Toast | `vocabularyToastOverlay` | on the panel | D-10 |

States: loading (the tile's image in the final frame), missing (fallback mark), saving (Use as portrait disabled). Ids `"\(pageIdentifier).viewer"`, `.viewer.previous`, `.viewer.next`, `.viewer.close`, `.viewer.useAsLead`, `.viewer.toast`; the record row keeps its own ids.

### Strings

L10n keys *(new)* in `L10n.Conclusions`, used by `ConclusionDetailGallery.ViewerCopy.likeness`:
- `likenessViewerTitle(_:)` = "Likeness of %@" (`conclusions.person.likenessViewer.title`; the argument is the person's title)
- `likenessViewerLabelFormat` = "Likeness %1$lld of %2$lld, from %3$@" (`conclusions.person.likenessViewer.label`, a `LocalizedStringResource`; the third argument is the citation line)
- `likenessViewerPrevious` = "Previous likeness", `likenessViewerNext` = "Next likeness", `likenessViewerClose` = "Close the viewer"
- `likenessViewHint` = "Opens the likeness in the viewer"

Reused: `L10n.DesignSystem.selectOptionPosition(current:count:)` (the counter), `useAsPortrait`, `portraitBadge`, the portrait Toasts. Changed: the Person tile image's hint and help (D-12); `likenessLabel` keeps S13-14's wording.

## 7. Steps

**STEP-1 — The step rule**
Files: `macos/App/Features/Conclusions/ConclusionPictureViewer.swift` *(new)*, `macos/ProvenenciaTests/ConclusionPictureStepTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-2 and D-3's `ConclusionPictureViewerItem` (values only, no view). T-1, T-2, T-3.
Check: `xcodebuild test` passes `ConclusionPictureStepTests`.

**STEP-2 — The full-size decode moves to Platform**
Files: `macos/App/Features/ArtifactViewer/ArtifactImageDecode.swift` → `macos/App/Platform/ArtifactImageDecode.swift` *(new)* (`git mv`), `api/ffi/swift_vocabulary_test.go` (the `swiftArtifactImageDecode` path constant, line 61), `.cursor/rules/macos-client.mdc` line 21, `macos/ProvenenciaTests/ArtifactViewerModelTests.swift`, `macos/ProvenenciaTests/ArtifactFileFixtures.swift`. Skill: `add-swift-test`.
Do: D-5's move and `declaredSize` split (T-5). In `macos-client.mdc` line 21, after "(or `ThumbnailCache` / `CachedThumbnail`, which call it)", add: "; the likeness viewer also resolves its URL through `ProjectFiles.objectURL` before `ArtifactImageDecode` (`Platform/`) reads it".
Check: the app builds; `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes (`TestSwiftVocabularyMatchesGo` reads the moved file); every existing Artifact viewer test passes unchanged.

**STEP-3 — The viewer's copy on the slot**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`ConclusionDetailGallery.ViewerCopy`, `viewer`), `macos/App/Features/Conclusions/PersonDetailContent.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-localized-string`.
Do: D-12; T-4.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; T-4 passes.

**STEP-4 — The viewer**
Files: `ConclusionPictureViewer.swift`. Skill: `add-ui-component`.
Do: D-1, D-5, D-6, D-8, D-9, D-10; the board's frames.
Check: a preview shows frames 6 – 8.

**STEP-5 — Opening and closing from the Gallery**
Files: `macos/App/Features/Conclusions/ConclusionGallerySection.swift` *(new)* in S13-14.
Do: D-3, D-4, D-7, D-11, D-12's hint and help.
Check: AC-1.7, AC-3.1 – AC-3.3, AC-4.1 – AC-4.3 by hand.

**STEP-6 — Wide and dark**
Files: `ConclusionPictureViewer.swift`.
Do: LV-W1 sizes by `PVWidthClass`; LV-9 in dark.
Check: AC-1.5 at a 2296pt pane.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (map row), this plan (status, Landed note), `docs/deployment-plan/spike-13/design/S13-D5-gallery-and-viewer.md` → `design/archive/`, `docs/deployment-plan/spike-13/design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/ConclusionPictureStepTests.swift` `stepsThroughTheGalleryWithoutWrapping` *(new)* | Tiles a…g showing b → index 1, count 7, previous a, next c; showing a → no previous; showing g → no next; one tile → neither | AC-1.2, AC-1.3, AC-1.4, D-2, D-8 |
| T-2 | Swift | `ConclusionPictureStepTests` `followsTheGalleryWhenTheShownLikenessLeaves` *(new)* | Showing e (lastIndex 4) of a…g, e removed → f at index 4; showing g (6), g removed → f at 5; showing the only tile, removed → nil | AC-4.1, AC-4.2, AC-4.3, D-2, D-3 |
| T-3 | Swift | `ConclusionPictureStepTests` `viewerItemResolvesAgainstTheGallery` *(new)* | Item `{shownID: "e", lastIndex: 4}`, tiles a…g without e → `resolved(in:)` is `{shownID: "f", lastIndex: 4}` with the same `id`; tiles empty → nil; tiles unchanged → equal to the input | D-3 |
| T-5 | Swift | `macos/ProvenenciaTests/ArtifactViewerModelTests.swift` `declaredSizeTurnsRotatedImages` *(new)*, with `ArtifactFileFixtures` | A 300 × 200 PNG → (300, 200); a JPEG 300 × 200 with EXIF orientation 6 → (200, 300), equal to `decode`'s `pixelSize`; a GIF → `.unreadable` | D-5, D-6 |
| T-4 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `personGalleryOffersTheViewer` *(new)* | `PersonDetailContent` for Johann Weber → `gallery.viewer == .likeness(title: "Likeness of Johann Weber")`, `label(index: 2, count: 7, line: "X") == "Likeness 2 of 7, from X"`; `EventDetailContent`, `PlaceDetailContent` → `gallery.viewer == nil` | AC-1.1, D-12 |

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
5. With the viewer open in one window, delete the shown likeness's depiction in a second window on the same project: the viewer moves to the neighbour; with the viewer on the last of several, delete it: it shows the new last; on the only one, delete it: it closes (AC-4.1, AC-4.2, AC-4.3).
6. Stretch past 1600pt: the large likeness grows, the small one stays capped (AC-1.5). Dark mode: frames 6 and 8 match the board (LV-9). VoiceOver reads each likeness as it steps (LV-8).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The sheet's focus or arrow keys are taken by the presenter | Stop condition below |
| The decode takes long on a 2048-pixel PNG | The tile's image shows in the final frame meanwhile (D-6); one decode at a time (D-5) |

**Stop and report if:**

- S13-14 landed differently from the § 4 rows marked *(S13-14)* (its **Landed** note says so).
- `.onKeyPress` inside `pvSheet` does not receive ← and → with focus on the image area (D-8), or Escape does not close through `PVSheetChrome`.
- The board picks form (B), or asks for zoom, autoplay or a viewer on Event or Place pages (D-4; user, 2026-11-03).
- Moving `ArtifactImageDecode` breaks an Artifact viewer test or `TestSwiftVocabularyMatchesGo` after the path constant is updated (STEP-2).

## 12. Open questions

None. The board's answers are in D-4, D-6, D-9 and D-11. The plan stays `Blocked` until S13-36 is planned and lands; it is refreshed then (its § 4 rows marked *(S13-14)*, the kit names) and reviewed again.

## 13. PR

**Title:** S13-15 — Likeness viewer
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-15-likeness-viewer.md`.
