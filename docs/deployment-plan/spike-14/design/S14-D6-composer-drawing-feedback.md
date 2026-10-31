# S14-D6 — Citation composer: depiction drawing feedback

**Kind:** Claude Design board (**enhancement** of the S10-D15 depiction row on the *Citation Composer* board)
**Spike:** Provenencia Spike 14 (odds and ends), batch 1 — Event and Place pictures
**View:** the Citation composer: its viewer pane (tool strip, locator strip, drawing band) and the observation dialog's depiction value area
**Implements later as:** PR **S14-03** (Rectangle-only depiction drawing)
**Board:** — (not drawn yet). Draw on the Claude Design project *Citation Composer*, file *Citation Composer.dc.html*, beside the S10-D15 "Depiction row" frames D1a – D10 (<https://claude.ai/design/p/8b120853-4ba4-4c0d-8492-75e76b9f9b7a?via=share>), as a new turn "S14-D6 · Drawing feedback".
**Depends on:** the agreed S10-D15 board (shipped in S10-34); S14-13 (landed [#632](https://github.com/mendahu/provenencia/pull/632)), which leaves one predicate for whether the viewer can cut the open file.
**Related:** [`../deployment-plan.md`](../deployment-plan.md) working rule "A depiction is drawn as a rectangle", bar outcomes 4 and 9; the 2026-10-18 likeness audit items UX-5, UX-6, UX-7, UX-12, UX-13 and the freehand cap ([`docs/ideas/depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md)); [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); plan [`../prs/S14-03-rectangle-only-depiction-drawing.md`](../prs/S14-03-rectangle-only-depiction-drawing.md).
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board. This file is the whole prompt.

### Claude Design — do this first (in order)

Work **in place** on this board. Do not fork a parallel copy of the surface.
- **Rethink** (this brief says replace): throw away the old frames. Do not keep a before/after to ship.
- **Enhancement**: add to the existing frames. Do not start a second composer / graph / page.

1. **Clear this board’s local design-system cache.** Claude Design keeps a stale pack; drawing against it invents local copies of kit controls.
2. **Delete this board’s reference** to the design-system bundle.
3. **Pull a fresh copy** of the Provenencia design system from the main project. Do not continue until the fetched kit lists current components. If the kit looks stale or empty, delete the cache and refetch. Do **not** draw a replacement kit locally.
4. **Compose from that kit.** Instance existing components. Reach for a **bespoke / local** control only when the use is truly this domain. One call site is not a new design-system primitive.

**Reach for (kit).** Instance these first. The **UI building-block inventory** later in this brief names the snowflakes and which kit piece each situation should use.

| Situation | Use |
| --- | --- |
| Labeled value, textarea, or trailing control | Field + TextArea / Input |
| Primary / secondary / ghost action | Button; icon-only → IconButton |
| Choose one from a short list | Select |
| Searchable pick | ComboBox |
| Warning, error, or inline hint | Callout |
| Page- or pane-level empty | EmptyState |
| Confirm replace or destroy | Confirm (`item:` snapshot, not a Bool) |
| Resource delete with inbound check | DeleteImpact recipe (S8-D9 / **S8-13**) — confirm if allowed, notice if blocked |
| Short create / edit form | FormDialog |
| Status / count / polarity mark | Badge; compact token → Chip |
| Cover or file thumb | Thumbnail |
| Grouping / raised or sunken row | Card |
| Section title | SectionHeader |
| Transient after-save notice | Toast |
| Native menu of actions | ContextMenu |
| Record or detail page at Regular and Wide | Page columns |
| Choose one of a few options inline | Segmented control (ChipGroup segmented) |

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

### Two widths — draw both

The window opens at 1366 × 900pt but is often stretched across a large monitor.
Every view has two width classes, measured on the view's own pane (the window
minus the sidebar), not the screen. A view with a side pane (an inspector, a detail
pane) plans that pane inside its Wide frame:

- **Regular** — pane below 1600pt. Draw the pane at 1102 × 900 (the default 1366pt window minus the open sidebar).
- **Wide** — pane 1600pt and up. Draw the pane at 2296 × 1440 (a 2560pt window minus the sidebar).

Draw only the view's own pane. Do not draw the app's sidebar, its navigation items or the window's title bar: they are not part of the board. A frame that needs the toolbar (breadcrumbs, Back and Forward) says so in the brief, and draws only the toolbar above the pane.

Rules for Wide:

1. **The extra space does a job.** Show more at once: a second column of supporting content, the detail beside its list, more grid or table columns, an inspector docked open. If the view truly has nothing to add, center its content. Content jammed to one side with empty space beside it is never the Wide design.
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths. Wide adds columns; it does not stretch text or fields.
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is an enhancement

Draw on the *Citation Composer* board, beside the S10-D15 "Depiction row" frames, as a new turn "S14-D6 · Drawing feedback". Change the composer's existing viewer pane and depiction value area in place. Do not redraw the evidence column or the observations column, and do not start a second composer. The composer's layout (viewer, evidence and observations panes) is not this board.

---

## 1. Objective

A researcher cuts a likeness or a picture out of a scan by drawing a region on it. Today the composer offers seven region shapes for that, though the saved picture is always the rectangle around the region; it lets them press controls that do nothing while they draw; it lets them start drawing on a file it cannot cut; and when a cut fails it says so only after Save. This board makes the composer guide the drawing and report what happens.

**What changes.**
- A depiction is drawn with the rectangle tool alone: the tool strip shows one radio while the observation dialog holds a depiction. A Citation's own region keeps all seven.
- While any region is drawn (a Citation's or a depiction's), the controls that would change the Citation's locator under the drawing are disabled: Set page, the × on the locator chips, the Clear menu, and Find. The locator strip's trailing hint ("Draw with the region tools to change it") hides.
- Draw and the region buttons are disabled on a file the viewer cannot cut, with a caption that says why.
- A cut that fails shows in the caption at once, and Save is disabled until the region is drawn again.
- An existing PDF Citation saved with no page offers a button in its warning Callout that saves the viewer's page to the Citation, so drawing can start.
- The captions name the region buttons, not only Draw. The preview of an edit that saves no new image is labelled as the saved image.
- VoiceOver hears when the freehand tool reaches its point cap (Citation regions only).

**What stays.** The composer's panes and their widths, the depiction row's layout (96pt preview on the left, caption and buttons on the right), the drawing band (hint and Cancel), the viewer's page and zoom groups, and every other part of the observation dialog. A depiction saved earlier as a polygon still opens, shows and saves as it is until it is drawn again.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A depiction is a crop | A likeness (Person), or a Picture (Event, Place): a region drawn on a Citation's image or PDF page. The saved image is always the rectangle around the region, so a depiction is drawn as a rectangle. |
| A Citation's own region | The part of the file the Citation reads: rectangle, four L shapes, circle or freehand. Its tools do not change. |
| Drawing mode | Draw opens drawing mode: the drawing band ("Draw the region with the tools above. Escape goes back to the observation." and Cancel) shows above the viewer, and the observation dialog waits. A Citation region is drawn by arming a radio on the tool strip, without the band. |
| The region buttons | Beside Draw: "Use the whole photo" (or "Use the whole page" on a PDF), and "Use the citation's region" when the Citation has one. "Use the citation's region" on a circle or freehand region sets the rectangle around it. |
| A file the viewer cannot cut | The viewer shows its empty state (missing file, load failed, image too large to show) or is still opening the file. |
| A failed cut | The viewer showed the file but the cut failed (the file changed on disk after it opened). |
| A PDF Citation with no page | A saved PDF Citation must have a page before it can take a depiction, because the depiction is checked against the saved page. |
| The freehand cap | A freehand region holds at most 256 points. Today a click past that beeps, with nothing for VoiceOver. |
| Two widths | The composer has three panes side by side on a wide window (viewer, evidence, observations) and two on a narrower one. This board's changes sit inside the viewer pane and the observation dialog and do not move between widths. |

### 2.1 What this board is not

- The composer's layout or the breakpoint between its two and three panes ([`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md)).
- Removing shapes from a Citation's own region.
- Showing every saved depiction on the photo.
- A distinct glyph for an image the thumbnail check refuses (likeness audit UX-11, idea file).
- The gallery, Person, Event or Place pages.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S14-03** | DF-1 … DF-11 and DF-W1, at Regular and Wide, light and dark. |
| Later boards | The composer's next layout brief moves its pane switch onto the Wide width class. |
| Stays out | Any engine change. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| DF-1 | **One radio for a depiction.** While the observation dialog holds a depiction Property (in drawing mode or not), the tool strip's region group shows the rectangle radio alone. In drawing mode it is selected; clicking it leaves drawing mode, as Escape and Cancel do. The page group, Set page and zoom group keep their places. |
| DF-2 | **Seven radios for a Citation region.** With no depiction in the dialog, the region group shows all seven radios, as today. |
| DF-3 | **Locked locator while drawing.** While a depiction is drawn (drawing band showing) or a Citation region tool is armed: Set page, the × on the page and region chips in the locator strip, and the Clear menu and the Find toggle are disabled (kit disabled state); Set page is drawn as the kit Chip (accent, small; selected when the page is set) so it has the kit's disabled state; the locator strip's trailing hint "Draw with the region tools to change it" is not drawn. The chips themselves stay. An open Find row closes when drawing starts, as today. |
| DF-4 | **Draw blocked on a file that cannot be cut.** When the viewer shows an empty state, Draw (or Draw again), "Use the whole photo / page" and "Use the citation's region" are disabled, and the caption reads "Draw is unavailable: the file is not showing in the viewer". While the viewer is still opening the file, the buttons are disabled, the preview shows its loading state, the caption is unchanged, and the dialog's Save waits. |
| DF-5 | **A failed cut shows at once.** When the cut fails, the 96pt preview stays the empty Thumbnail (the Thumbnail is never an error display), the caption reads "The region could not be cut from the file. Draw it again." in the warning tone, Draw again and the region buttons stay enabled, and the dialog's Save is disabled. Drawing again, or pressing a region button again, retries the cut. |
| DF-6 | **Save a page from the callout.** On an existing PDF Citation saved with no page, the warning Callout reads "This citation has no page yet. Save the viewer's page to it to draw a region." with one secondary small Button "Save page 3 to this citation" (the viewer's page) in the Callout's actions row. The Button is disabled while the viewer is still opening the file. Pressing it shows the Button's loading state, then the Callout goes away, the caption becomes the empty caption and Draw is enabled. A failure shows the Citation's save error where the composer shows it today. |
| DF-7 | **Empty caption.** "Nothing drawn yet. Draw on the file, or use a region button below". |
| DF-8 | **No-page caption.** "Draw and the region buttons wait until the citation has a page". |
| DF-9 | **Region-hidden Callout.** "This region no longer matches the citation's page. Draw it again or use a region button below." |
| DF-10 | **Saved image label.** On an edit whose region is unchanged, the preview's VoiceOver label is "Saved image", not "Image to save". A new or redrawn region keeps "Image to save". Nothing visible changes; draw it as a note on frame 6. |
| DF-11 | **Freehand cap.** When a Citation's freehand region reaches 256 points, a click past that beeps, as today, and VoiceOver hears "This freehand region holds its most points, 256. Click the first point to close it." once per shape (the number is the app's point limit) Nothing visible changes; draw it as a note on frame 2. |
| DF-W1 | **Wide.** The composer at Wide is its three-pane arrangement today: the viewer pane takes the extra width and the canvas fills it (rule 3); the observation dialog keeps its Regular width in the observations column (rule 2). Every DF requirement looks the same at both widths; draw frames 1 and 5 at Wide. |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Citation composer, viewer pane | Tool strip, locator strip, drawing band, canvas | The same; the canvas fills the wider pane | Nothing |
| Citation composer, observation dialog | The depiction row: preview, caption, buttons | The same, at the observations column's fixed width | Nothing |

---

## 5. Suggested frames

Draw each as the composer pane only, with no app sidebar, at Regular (1102 × 900) unless it says Wide (2296 × 1440). Use one Citation of a family photo ("Robins family, 1952", image, with a circle region round a face, so the locator strip shows a region chip) for frames 1 – 6, and a parish register PDF for frames 7 – 8.

1. **Drawing a likeness.** Drawing band showing; the tool strip's region group has the rectangle radio alone, selected; the region chip's × and the Clear menu are disabled (Set page and Find appear only on a PDF; draw a small PDF variant with both disabled); the locator strip has no trailing hint; a rectangle being dragged on a face (DF-1, DF-3).
2. **Drawing a Citation region.** No depiction dialog; the freehand radio armed among all seven; the locator's chip × and the Clear menu disabled, no trailing hint; a freehand shape in progress, with a note for the cap announcement (DF-2, DF-3, DF-11).
3. **Empty depiction row.** The dialog with Property Likeness and no region: the new empty caption under it, Draw, "Use the whole photo", "Use the citation's region" (DF-7).
4. **File cannot be cut.** The viewer showing its "Image too large to show" empty state; the depiction row's buttons disabled and the DF-4 caption (DF-4).
5. **Failed cut.** A drawn region whose cut failed: the empty Thumbnail, the warning caption, Draw again enabled, the dialog's Save disabled (DF-5).
6. **Unchanged edit.** An edit of a saved likeness: the saved crop in the preview, with a note for the "Saved image" label (DF-10).
7. **PDF Citation with no page.** The register on page 3, an existing Citation with no page: the DF-6 Callout with "Save page 3 to this citation", the DF-8 caption, the buttons disabled (DF-6, DF-8).
8. **After saving the page.** Frame 7 after the press: no Callout, the page chip "Page 3" in the locator strip, the empty caption, Draw enabled (DF-6).
9. **Region hidden.** A saved likeness whose region no longer matches the Citation's page: the DF-9 Callout (DF-9).
10. **Dark mode.** Frames 1 and 5 in dark mode.
11. **Wide.** Frames 1 and 5 at Wide (DF-W1).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Viewer tool strip | Snowflake | **Extend** | `Features/ArtifactViewer/ArtifactViewer.swift` `ArtifactViewerToolChrome` | DF-1, DF-2. Takes the radios to show; the viewer does not know about depictions. Not a second strip. |
| Region radio | Component | Ship | `PVIconButton` with `isSelected` | DF-1, DF-2. The same radios, fewer of them. |
| Set page | Component | **Rethink** | `PVChip` accent, small, in `ArtifactViewerToolChrome` (today a hand-built button with the Chip's colors) | DF-3. Same titles and selected look; gains the kit's disabled state. |
| Clear menu | Component | Ship | `PVButton` link in `CitationComposerViewerPane` `clearMenuButton` | DF-3. Disabled state only. |
| Locator strip and chips | Snowflake | **Extend** | `Features/CitationComposer/CitationComposerViewerPane.swift` `locatorStrip`, `locatorChip` | DF-3. The × is `PVIconButton`, disabled; the trailing hint is not drawn while drawing. |
| Find toggle | Component | Ship | `PVIconButton` in `CitationComposerViewerPane` `findToggleButton` | DF-3. Disabled state only. |
| Drawing band | Snowflake | Ship | `CitationComposerViewerPane` `drawingBand` (`PVCallout` info + `PVButton` ghost) | Unchanged. |
| Depiction row | Snowflake | **Extend** | `Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationDepictionValueArea` | DF-4 – DF-10. Same layout. |
| Needs-page and region-hidden warnings | Component | Ship | `PVCallout` warning, with its actions row for DF-6 | DF-6, DF-9. The Callout's existing actions row; not a new banner. |
| Save page button | Component | Ship | `PVButton` secondary small with `loading` | DF-6. |
| Preview | Component | Ship | `PVThumbnail` (empty for DF-5, loading for DF-4), `CachedThumbnail` for the saved crop | DF-4, DF-5, DF-10. |
| Caption | Snowflake | Ship | the depiction row's italic caption `Text` | DF-4, DF-5, DF-7, DF-8. DF-5 uses the warning foreground. |
| Dialog Save | Component | Ship | the observation dialog's Save `PVButton` | DF-5. Disabled state only. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A shape picker, menu or segmented control for a depiction | A depiction has one tool (working rule "A depiction is drawn as a rectangle"). |
| A new banner, toast or alert, or a warning glyph in the Thumbnail, for a failed cut | The caption says it in place (DF-5); the Thumbnail is never an error display. |
| Hiding the locator chips or the tool strip while drawing | DF-3 disables; it does not hide, so the layout does not jump. |
| A page picker in the needs-page Callout | DF-6 saves the page the viewer shows; the page group already picks it. |
| A visible counter of freehand points | DF-11 is a VoiceOver announcement. |
| Any change to the composer's panes at Wide | DF-W1; the layout move is the composer's next brief. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the frame numbers and every change from this brief with the user's answer.
2. Refresh the [S14-03 plan](../prs/S14-03-rectangle-only-depiction-drawing.md): copy the answers into its Decisions and **Design → macOS**, describe each frame in words, run `review-plan` again, and set it to `Planned`.
3. S14-03 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md).
