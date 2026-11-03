# S13-15 — Likeness viewer

**Status:** Blocked — on board S13-D5 and on S13-14
**Spike / slice:** Spike 13 · Slice 5 — Gallery
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D5](../design/S13-D5-gallery-and-viewer.md), board not drawn yet
**Board:** not drawn yet. When agreed, the Claude Design project, file, turn and frames go here; the implementer may not be able to open it, so **Design → macOS** will describe every frame in words
**Depends on:** S13-14 (not landed: `ConclusionGallerySection`, the tile's citation line), S13-D5 (board)
**Blocks:** —
**Branch:** `s13-15-likeness-viewer`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go or proto change
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 4. Gallery](../person-detail.md#4-gallery) (R-30, R-31, R-32); brief [S13-D5](../design/S13-D5-gallery-and-viewer.md) LV-1 – LV-9, LV-W1; the [S13-14 plan](S13-14-gallery-section.md) and, once it lands, its **Landed** note; working rules: "Sections are places" (the viewer is not one), "One sectioned Conclusion page, not a Person fork"
**Planned:** 2026-11-03 against `main` @ `42eb321` (draft: written before S13-14 lands; § 4 rows marked *(S13-14)* are that plan's design, and are re-read in the code on refresh)
**Review:** [`S13-15-likeness-viewer.review.md`](S13-15-likeness-viewer.review.md)

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

**What stays.** Everything S13-14 shipped except the tile image's click, which moves from opening the Citation to opening the viewer; Open citation stays on its own control and in the tile's context menu. The header and its thumb menu. The Event and Place Pictures (user, 2026-11-03: Person only). `ConclusionLeadModel` and its Toasts. No navigation place is added: the viewer is not in the history.

## 2. User stories

**US-1.** As a researcher, I open a likeness large and step through the gallery.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann's seven likenesses (brief § 5); click tile 2's image | The viewer opens on likeness 2: the image whole, fit, at most its pixel size ÷ 2 in points (300 × 380 pixels → at most 150 × 190pt); counter "2 of 7"; caption "Hamburger Adressbuch 1861" over "Volume 3, Entry 214, p. 23" and "OBS-…" |
| AC-1.2 | In the viewer on 2, press → | Likeness 3 shows, counter "3 of 7", VoiceOver reads the new likeness |
| AC-1.3 | On 1 | Previous is disabled and ← does nothing; on 7, next is disabled and → does nothing (no wrap) |
| AC-1.4 | A gallery of one | No previous, next or counter |
| AC-1.5 | Likeness 4 (1400 × 2048 pixels) at Regular, then at Wide | The image fills the viewer's image area, capped at 700 × 1024pt, larger at Wide (LV-W1) |
| AC-1.6 | Likeness 7 (missing image) | The Person mark in Johann's gender colors on the sunken surface; caption and actions stay |
| AC-1.7 | Tab to a tile's image, press Return | The viewer opens on that likeness |

**US-2.** As a researcher, I act on the likeness I am viewing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Viewer on 2 → Set as portrait | The header thumb and the grid behind show likeness 2 as the portrait; the viewer stays on 2 and shows the Portrait Badge in place of the button; the portrait Toast shows |
| AC-2.2 | Viewer on 2 → Open citation | The viewer closes and the Citation composer opens at the likeness; Back returns to Johann's Gallery with the viewer closed |
| AC-2.3 | Viewer on 4 (the portrait) | The Portrait Badge, no Set as portrait |

**US-3.** As a researcher, closing the viewer puts me back where I was.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Open on 2, step to 5, press Escape | The viewer closes; focus is on tile 5, scrolled into view |
| AC-3.2 | Click outside the viewer, or its close button | The same as AC-3.1 |
| AC-3.3 | Viewer open, then Back, then Forward | Back leaves the Person page (the viewer is not a history step); Forward returns to the Gallery with the viewer closed |

**US-4.** As a researcher, the viewer survives the gallery changing under it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Viewer on 5 of 7; likeness 5 is deleted in another window | The viewer shows the likeness now at position 5 (the old 6), "5 of 6" |
| AC-4.2 | Viewer on 7 of 7; likeness 7 is deleted | The viewer shows the new last (the old 6), "6 of 6" |
| AC-4.3 | Viewer on the only likeness; it is deleted | The viewer closes; the Gallery shows its EmptyState |

## 3. Scope

| In | Covered by |
| --- | --- |
| The viewer in the board's form, opened from a tile's image (LV-1) | STEP-3, STEP-4, AC-1.1, AC-1.7 |
| The image, fit and capped, the loading and missing states (LV-2) | STEP-2, STEP-3, AC-1.1, AC-1.5, AC-1.6 |
| Stepping, the counter, the ends (LV-3) | STEP-1, STEP-3, AC-1.2 – AC-1.4 |
| The caption (LV-4) | STEP-3, AC-1.1 |
| Set as portrait and Open citation in the viewer (LV-5) | STEP-3, AC-2.1 – AC-2.3 |
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
| Prefetching the neighbours' full images | not planned (D-5: the tile's image shows while one loads) |

## 4. Current state

Rows marked *(S13-14)* are S13-14's planned design; re-read them in the code once it lands.

| Area | Where | Today |
| --- | --- | --- |
| Gallery section *(S13-14)* | `macos/App/Features/Conclusions/ConclusionGallerySection.swift` *(new)* in S13-14 `ConclusionGallerySection` | Kind-blind; the tile's image is a `Button` → `navigation.go(to: tile.location)`; Open citation control and context menu; `@FocusState focusedID` per tile; `fallbackColors` |
| Tile data | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPicture` (+ `citationLine`, `titleIsRef` *(S13-14)*), `ConclusionDetailGallery` | `id`, `imageRelPath`, `sourceTitle`, `location`, `isLead`, labels; `lead`, `offersEarliest` |
| Record parts | `ConclusionDetailPicture` has no Artifact, Citation or page parts; `CitationLineDisplay.line(_:)` joins them from `CatalogReconcilerOutcome` (`artifactLabel`, `citationLabel`, `pageLabel`, `artifactPage`, `observationRef`) | The viewer caption needs the parts after the title and the ref separately (LV-4) |
| Lead model | `macos/App/Features/Conclusions/ConclusionLeadModel.swift` | `choose(_:copy:)` → `Bool`, `isSaving`, `toast` |
| Sheet presenter | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` `pvSheet(item:…)`, `PVSheetChrome` | Escape and a click outside close a sheet through one rule |
| Sheet precedent | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailRowsSection` (`provenanceItem`, `.pvSheet(item:)`, `.onChange(of: rows)` closing a sheet whose row left); `macos/App/Features/Conclusions/ProvenanceSheet.swift` `ProvenanceSheetItem.row(in:)`, `ProvenanceSheet.width(_:)` (880 / 1280), `open(_:)` = `close()` then `navigation.go(to:)` | An item that holds an id, resolved against the live rows on every draw |
| Thumbnail cache | `macos/App/Platform/ThumbnailCache.swift` `ThumbnailCache.image(projectDir:relPath:)`, `cachedImage(projectDir:relPath:)`, `maxPixelSize` (512); its decode: `ProjectFiles.objectURL` → `CGImageSourceCreateWithURL` → `ThumbnailDecodePolicy.accepts` → `ImageHeaderDecode.decode(_:maxPixelSize:)` in a detached task | Holds tiles at 512 pixels; never a full-size image (its doc) |
| Decode policy | `ThumbnailCache.swift` `ThumbnailDecodePolicy` (PNG or JPEG, edges ≤ 4096); `macos/App/Platform/ImageHeaderDecode.swift` `ImageHeaderDecode.decode` | The checks a full-size decode must also pass |
| Scroll into view | `macos/App/DesignSystem/Components/ScrollRowIntoView/PVScrollRowIntoView.swift` `PVScrollRowIntoView` | Scrolls the enclosing `ScrollView` to a row tagged `.id(row)` |
| Icons | `macos/App/DesignSystem/Components/Icon/PVIcon.swift` `.chevronBack`, `.chevronForward`, `.dismiss` | For previous, next, close |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift`, `ReconciledValueDisplayTests.swift` fixture `picture(_:path:title:leads:)` | Gallery content; no viewer |

**Precedent.** S13-11 ([#650](https://github.com/mendahu/provenencia/pull/650)): the provenance sheet is an item holding an id, presented with `.pvSheet(item:)`, resolved against the live rows on every draw, closed when its row leaves, sized by width class, and closing before it navigates. Copy that shape for form (A). Do not copy the Artifact viewer (`Features/ArtifactViewer/`): it zooms, pans and draws regions.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The viewer is `ConclusionPictureViewer` *(new)* in `macos/App/Features/Conclusions/ConclusionPictureViewer.swift` *(new)*, kind-blind: it takes `projectDir`, `gallery: ConclusionDetailGallery`, `lead: ConclusionLeadModel`, `fallbackColors`, the shown picture's id and closures `step(to:)`, `openCitation(_:)` and `close()`. Only `ConclusionGallerySection` presents it, so only the Person page has it | User, 2026-11-03 (Person only); kind-blind so Event and Place adopt it with the section | A Person-named viewer — the second caller would fork |
| D-2 | Stepping is a pure value, `ConclusionPictureStep` *(new, in the same file)*: `init?(tiles: [ConclusionDetailPicture], showing id: String, lastIndex: Int)` resolves the shown tile; `index`, `count`, `hasPrevious`, `hasNext`, `previousID`, `nextID`; when `id` is no longer in `tiles`, it resolves to the tile at `min(lastIndex, count − 1)`, and is nil for an empty gallery | LV-3, LV-7 as one testable rule; the view does no index arithmetic | Index math in the view — untestable; wrapping — brief § 2 says no wrap |
| D-3 | `ConclusionGallerySection` holds `@State viewing: ConclusionPictureViewerItem?` *(new)* = `{ id: String, lastIndex: Int }`. The tile image's button sets it (replacing `navigation.go`). On every draw the item resolves through `ConclusionPictureStep`; `.onChange(of: gallery.tiles)` replaces the item with the resolved tile's id and index, or sets it nil when the step is nil | The S13-11 item pattern; AC-4.1 – AC-4.3 | A snapshot of the tile — it would show a deleted likeness |
| D-4 | Form (A), a kit sheet, presents with `.pvSheet(item: $viewing)`, sized by width class as `ProvenanceSheet.width(_:)` does, with the board's sizes. Form (B), an overlay, is drawn by `ConclusionGallerySection` over the page with Escape and a click outside routed the same way. The board's LV-1 answer fixes which (§ 12); the rest of this plan holds for both | One presenter, one dismiss rule | A third presenter |
| D-5 | The full image loads through `ProjectImageDecode.load(projectDir:relPath:maxPixelSize:)` *(new)* in `macos/App/Platform/ProjectImageDecode.swift` *(new)*: the URL, policy and header decode that `ThumbnailCache.image` runs today, moved out of the cache's closure into one `nonisolated` async function that both call. The viewer passes `ThumbnailDecodePolicy.maxDeclaredEdge` (4096; stored crops are at most 2048). It is not cached; the viewer holds the one decoded image in `@State` keyed by the shown id. While it loads, the viewer draws `ThumbnailCache.shared.cachedImage(…)` scaled; on failure the fallback mark | One decode path and one policy for every project image; the cache stays thumbnail-only (its doc) | A larger `ThumbnailCache.maxPixelSize` — every tile would hold a full image; a second copy of the decode in the viewer |
| D-6 | The image is drawn with `.resizable().scaledToFit()` inside a frame capped at the image's pixel size ÷ 2 in points, centered on `PVColor` sunken | LV-2 ("never above its pixel size at 2×") | The screen's backing scale — the cap would change between displays |
| D-7 | The viewer is not a `WorkspaceLocation` and is not persisted; navigating away (Open citation, Back) closes it | Brief § 2; S13-11's sheet is not a place either | A history step — Back would reopen a modal |
| D-8 | ← and → step through `.onKeyPress(.leftArrow)` / `.onKeyPress(.rightArrow)` on the viewer, returning `.ignored` at the ends; the previous and next `PVIconButton`s are disabled at the ends | LV-3; the kit has no stepper | A wrap — brief § 2 |
| D-9 | Set as portrait in the viewer calls `lead.choose(tile, copy: gallery.choose)`; the viewer stays on the same id (D-3 resolves it after the refresh). Open citation runs `close()` then `navigation.go(to: tile.location)`, as `ProvenanceSheet.open(_:)` | AC-2.1, AC-2.2; the precedent | Navigating with the sheet open — the sheet would outlive the page |
| D-10 | On close, `ConclusionGallerySection` sets `focusedID` to the last shown id and scrolls it into view with `PVScrollRowIntoView` (each tile `.id(tile.id)`) | LV-6, AC-3.1 | Focus to the first tile — loses the place |
| D-11 | The caption reads `ConclusionDetailPicture.citationDetail: String` *(new)*: the citation line without the Source title (the Artifact, Citation and page joined as `CitationLineDisplay` joins them), and `observationRef: String` *(new)*, both set in `ConclusionDetailGallery.tiles(from:…)`. `CitationLineDisplay` gains `detail(_:)` *(new)*, which `line(_:)` then uses after the title | LV-4 needs the title and the rest on two lines; one formatter | Splitting the joined line in the view — breaks with a separator inside a label |

The board settles LV-1's form and sizes, LV-5's order and primary action, and the viewer's measures; § 12 lists them. They change no Decision above.

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionPictureViewer` *(new)*, `ConclusionPictureViewer.swift` *(new)* | view, file | `ProvenanceSheet`; `Features/ArtifactViewer/ArtifactViewer` | The provenance sheet shows rows, not an image; the Artifact viewer zooms and draws regions (brief non-goals). It follows `ProvenanceSheet`'s shape |
| `ConclusionPictureStep` *(new)* | value | none (searched `index(after`, `previousID`, `hasNext` in `Features/`) | No stepper exists |
| `ConclusionPictureViewerItem` *(new)* | value | `ProvenanceSheetItem` | Same shape, different id space; one line each |
| `ProjectImageDecode.load` *(new)*, `ProjectImageDecode.swift` *(new)* | func, file | the decode closure inside `ThumbnailCache.image` | Extracted from it (D-5), so it is not a second copy |
| `ConclusionDetailPicture.citationDetail`, `.observationRef` *(new)* | fields | `citationLine` *(S13-14)* | Parts of the same record the caption shows on separate lines |
| `CitationLineDisplay.detail(_:)` *(new)* | func | `CitationLineDisplay.line(_:)` | `line` is rebuilt on it, so one formatter |
| L10n keys (§ Strings) | keys | `L10n.Conclusions` likeness keys | New wordings |

**Builds on.** The S13-11 sheet item pattern; `pvSheet`; `ThumbnailDecodePolicy`; `CitationLineDisplay`; `ConclusionLeadModel`; `PVScrollRowIntoView`.
**Deletes.** The tile image's `navigation.go(to:)` in `ConclusionGallerySection` (Open citation keeps its own control). The decode closure body inside `ThumbnailCache.image`, replaced by a call to `ProjectImageDecode.load`.
**Next adds.** The Event and Place redesign gets the viewer by adopting `ConclusionGallerySection`. A zoom, if ever wanted, replaces the image view inside `ConclusionPictureViewer`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| The viewer on a likeness deleted meanwhile | `ConclusionPictureStep` | D-2, D-3: replaced by its neighbour, closed when none; T-2 |
| Open citation from the viewer | `WorkspaceNavigation.go(to:)` | The location the tile already opens; the viewer closes first (D-9) |
| A full-size decode of a project object | `ProjectImageDecode.load` | The same `ProjectFiles.objectURL` path check and `ThumbnailDecodePolicy` the cache runs; a refused file returns nil → fallback mark |

### Scale

At the brief's sizes (40 likenesses at most): one full decode (≤ 2048 × 2048, 16 MB decoded) per likeness shown, released when the viewer moves on; no cache growth. Stepping fast through 40 cancels each superseded load (`.task(id:)` on the shown id). No write is added.

### Edges

- **Assumptions.** The tiles have unique ids (Observation ids) and keep the engine's order across a refresh; `ConclusionPictureStep` relies on both. T-1, T-2.
- **Neighbors.** First, last, one, two likenesses; the shown one removed at the start, middle and end; all removed (T-2).
- **Removed inputs.** A removed likeness's image is never shown again: the resolved id comes from the live tiles (D-3).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** The image file: edges ≤ 4096 checked from the header before any pixel decodes (`ThumbnailDecodePolicy.accepts`), decode capped by `maxPixelSize`.

### Schema

None — no catalog change.

### Go core

None —.

### FFI

None —.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Viewer | `ConclusionPictureViewer` *(new)* | `Features/Conclusions/ConclusionPictureViewer.swift` | D-1, D-4; sizes per the board |
| Image area | `Image(nsImage:)` `.resizable().scaledToFit()`; fallback `PVThumbnail` with the mark | same | D-5, D-6 |
| Previous, next, close | `PVIconButton(.chevronBack / .chevronForward / .dismiss, label:)` | same | D-8; close calls `close()` |
| Counter, ref | `Text(verbatim:)` in `PVFont.mono(size: PVTypeScale.caption)` | same | `L10n.Conclusions.likenessViewerPosition` |
| Caption | title in `PVFont.body`, mono when `titleIsRef`; `citationDetail` muted | same | D-11 |
| Set as portrait / Portrait Badge | `PVButton(gallery.choose.useAs)` / `PVBadge(gallery.choose.badge, tone: .accent, icon: .photo)` | same | D-9; disabled while `lead.isSaving` |
| Open citation | `PVButton(gallery.openTip)` | same | D-9 |

States: loading (the cached tile image scaled, else a small `ProgressView`), missing (fallback mark in `fallbackColors`), saving (Set as portrait disabled). Accessibility: the image area is one element labelled `likenessViewerLabel(index, count, citationLine)` and is focused when the viewer opens; stepping posts the new label (`AccessibilityNotification.Announcement` through the kit's `PVAnnounce` (`DesignSystem/Components/Announce/PVAnnounce.swift`) if the board's form loses focus on redraw); ids `"\(pageIdentifier).viewer"`, `.viewer.previous`, `.viewer.next`, `.viewer.close`, `.viewer.useAsLead`, `.viewer.openCitation`.

### Strings

L10n keys *(new)* in `L10n.Conclusions`:
- `likenessViewerPosition(_:_:)` = "%1$d of %2$d" (`conclusions.person.likenessViewer.position`)
- `likenessViewerLabel(_:_:_:)` = "Likeness %1$d of %2$d, from %3$@" (`conclusions.person.likenessViewer.label`)
- `likenessViewerPrevious` = "Previous likeness", `likenessViewerNext` = "Next likeness", `likenessViewerClose` = "Close the viewer"
- `likenessViewHint` = "Opens the likeness in the viewer" (the tile image's accessibility hint)

Reused: `likenessOpenTip` ("Open citation"), `useAsPortrait`, `portraitBadge`, the portrait Toasts. Changed meaning: none. The tile's spoken label keeps S13-14's wording; the hint changes from opening the Citation to opening the viewer.

## 7. Steps

**STEP-1 — The step rule**
Files: `macos/App/Features/Conclusions/ConclusionPictureViewer.swift` *(new)*, `macos/ProvenenciaTests/ConclusionPictureStepTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-2 (`ConclusionPictureStep` only). T-1, T-2.
Check: `xcodebuild test` passes `ConclusionPictureStepTests`.

**STEP-2 — One decode for project images**
Files: `macos/App/Platform/ProjectImageDecode.swift` *(new)*, `macos/App/Platform/ThumbnailCache.swift`, `macos/ProvenenciaTests/ThumbnailCacheTests.swift` (unchanged), `macos/ProvenenciaTests/ProjectImageDecodeTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-5: move the decode out of `ThumbnailCache.image`'s closure into `ProjectImageDecode.load`; the cache calls it with `ThumbnailCache.maxPixelSize`. T-3.
Check: every existing `ThumbnailCache` test passes unchanged; T-3 passes.

**STEP-3 — The viewer**
Files: `ConclusionPictureViewer.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`ConclusionDetailPicture`, `ConclusionDetailGallery.tiles`), `macos/App/Features/Conclusions/CitationLineDisplay.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skills: `add-ui-component`, `add-localized-string`.
Do: D-1, D-6, D-8, D-9, D-11; the board's frames; T-4.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; a preview shows frames 6 – 8.

**STEP-4 — Opening and closing from the Gallery**
Files: `macos/App/Features/Conclusions/ConclusionGallerySection.swift` *(new)* in S13-14.
Do: D-3, D-4, D-7, D-10: the tile image opens the viewer; the hint; the item resolution and `.onChange`; focus return and scroll.
Check: AC-1.7, AC-3.1 – AC-3.3, AC-4.1 – AC-4.3 by hand.

**STEP-5 — Wide and dark**
Files: `ConclusionPictureViewer.swift`.
Do: LV-W1 sizes by `PVWidthClass`; check LV-9 in dark.
Check: AC-1.5 at a 2296pt pane.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (map row), this plan (status, Landed note), `docs/deployment-plan/spike-13/design/S13-D5-gallery-and-viewer.md` → `design/archive/`, `docs/deployment-plan/spike-13/design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/ConclusionPictureStepTests.swift` `stepsThroughTheGalleryWithoutWrapping` *(new)* | Tiles a…g showing b → index 1, count 7, previous a, next c; showing a → no previous; showing g → no next; one tile → neither | AC-1.2 – AC-1.4, D-2, D-8 |
| T-2 | Swift | `ConclusionPictureStepTests` `followsTheGalleryWhenTheShownLikenessLeaves` *(new)* | Showing e (lastIndex 4) of a…g, e removed → f at index 4; showing g (6), g removed → f at 5; showing the only tile, removed → nil | AC-4.1 – AC-4.3, D-2, D-3 |
| T-3 | Swift | `macos/ProvenenciaTests/ProjectImageDecodeTests.swift` `decodesWithinThePolicy` *(new)* | A 2048 × 1024 PNG in a temp project → an image with long edge ≤ the cap passed; a GIF → nil; a path outside `objects/` → nil | D-5 |
| T-4 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `galleryTilesCarryTheCitationDetail` *(new)* | The S13-14 T-1 record → `citationDetail == "Volume 3, Entry 214, p. 23"`, `observationRef` set; no parts → `""`; and `CitationLineDisplay.line` unchanged for the S13-11 cases | AC-1.1, D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `ThumbnailCacheTests` (STEP-2), `PersonDetailTests` (the `CitationLineDisplay.line` cases, D-11).

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

1. Use the S13-14 manual check's Person (at least three likenesses, one with a missing image). Open the Gallery and click the second tile's image: the viewer opens on it with the counter, caption and actions; a small likeness is not blown up (AC-1.1). Tab to a tile's image and press Return: the same (AC-1.7).
2. Press → and ←, and the buttons: it steps; at the ends the buttons are disabled and the keys do nothing (AC-1.2, AC-1.3). On a person with one likeness there is no stepping (AC-1.4). On the missing image: the fallback mark in the person's colors (AC-1.6).
3. Set as portrait in the viewer: the header and the grid behind change, the viewer stays, the Badge replaces the button; on the portrait itself there is only the Badge (AC-2.1, AC-2.3). Open citation: the viewer closes, the composer opens; Back returns to the Gallery, viewer closed (AC-2.2).
4. Step to another likeness and press Escape; then reopen and click outside; then reopen and press close: each time focus lands on the last shown tile, scrolled into view (AC-3.1, AC-3.2). With the viewer open press Back, then Forward: the viewer is closed (AC-3.3).
5. With the viewer open in one window, delete the shown likeness's depiction in a second window on the same project: the viewer moves to the neighbour; with the viewer on the last of several, delete it: it shows the new last; on the only one, delete it: it closes (AC-4.1, AC-4.2, AC-4.3).
6. Stretch past 1600pt: the large likeness grows, the small one stays capped (AC-1.5). Dark mode: frames 6 and 8 match the board (LV-9). VoiceOver reads each likeness as it steps (LV-8).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The sheet's focus or arrow keys are eaten by the presenter | Stop condition below; the overlay form (D-4 B) is the fallback only with the user's answer |
| Moving the decode out of `ThumbnailCache` changes its concurrency or cancellation | D-5 moves only the inner detached body; the slot, waiters and cancellation stay in the cache; its tests must pass unchanged (STEP-2) |

**Stop and report if:**

- S13-14 landed differently from the § 4 rows marked *(S13-14)* (its **Landed** note says so).
- `.onKeyPress` inside `pvSheet` does not receive ← and → with focus on the image area (D-8), or Escape does not close through `PVSheetChrome`.
- The board asks for zoom, autoplay or a viewer on Event or Place pages (user, 2026-11-03).
- `ThumbnailCache`'s tests need changing for STEP-2 to pass.

## 12. Open questions

The board settles these; the plan stays `Blocked — on board S13-D5 and on S13-14` until they are answered and S13-14 has landed:

1. LV-1: the viewer's form, (A) the kit sheet or (B) an overlay (D-4), and its sizes at Regular and Wide (LV-W1).
2. LV-3: where previous and next sit.
3. LV-5: the order of Set as portrait and Open citation, and which is primary.

## 13. PR

**Title:** S13-15 — Likeness viewer
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-15-likeness-viewer.md`.
