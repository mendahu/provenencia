# S15-D14 — Source page Artifacts: a list that shows each Artifact's copy details, and the Artifact sheet

**Kind:** Claude Design board (**rethink** of the Source page's Artifacts section: the card grid S2-25 drew, with S15-09's Copy details disclosure on each card. Replace those frames. The rest of the Source page is unchanged)
**Spike:** Provenencia Spike 15, slices 2 and 7
**View:** the Source page's Artifacts section, the Artifact sheet it opens (add and edit), and the inline Add field row on every metadata board (the Source's own Metadata section included)
**Implements later as:** PR **S15-34** (SA-4, SA-6, SA-8, SA-9, SA-12 – SA-15), then S15-35 (SA-1 – SA-3, SA-5, SA-10, SA-11, SA-W1), S15-10 (SA-7) and S15-25 (SA-5's Move item)
**Board:** agreed 2026-11-13 — [Claude Design — Source Detail](https://claude.ai/artifact/443NPNYmGErAmxgV1fBsFK), page *Source page*: rows "Artifacts" (art-list, art-list-wide, art-sixty, art-menu, art-empty, art-gone), "Artifact sheet" (sheet-edit to art-refused), "Add an Artifact" (sheet-add to add-attach) and "Metadata" (meta-add-row, meta-add-row-wide). Board decisions: **rows**, no compact cards (SA-1); a **64pt** square thumbnail (SA-2); a row's grid is thumbnail, label and description, a copy-details column of **320pt at Regular and 440pt at Wide** (field name 96pt, mono value), then the actions; Set cover keeps its space when hidden so the copy-details column lines up row to row; the Add field row is a bordered card row: a field ComboBox (168pt in the sheet, 150pt in the Source's section), a mono value Input, ✓ and ✕. User calls (2026-11-13): the board's own discard wording ("…are not kept. Nothing has been saved.", "Keep editing") is **not adopted**, the kit's standard discard confirm stays (SA-8); the Source's own Metadata section has **no** Add field button at its foot, only the header's (SA-13); a row's ✓ is disabled while its value is empty (SA-6). Also on the board, and scoped into the PRs (user, 2026-11-13): the Metadata section's intro is a help tooltip beside its title and its field count is gone (S15-34); the header's type pill is the kit Tag (not the Badge; both stay in the kit), a long-standing mismatch fixed in S15-35 with the meta row's clipping and the description's smaller caption text. Cleanup checked 2026-11-13: the board's discard confirm is now the kit's standard one ("Discard changes?", "Your changes will be lost.").
**Review:** [`S15-D14-source-page-artifacts.review.md`](S15-D14-source-page-artifacts.review.md) — Fits with changes — Ready with changes, 2 rounds; every blocker Fixed
**Depends on:** S15-08 (landed [#791](https://github.com/mendahu/provenencia/pull/791)): every Artifact holds its own metadata (its copy details); S15-09 ([#802](https://github.com/mendahu/provenencia/pull/802), trimmed before merge): a board write patches the page's cached workspace; S15-36: a metadata board can hold a draft; S15-33: an Artifact update carries its draft and names a refused field
**Related:** working rules "An Artifact is added and edited in one sheet", "A board adds a field in an inline row", "One metadata board", "Metadata saves with its owner's write" in [`../deployment-plan.md`](../deployment-plan.md#working-rules); [`S15-D1-source-page.md`](S15-D1-source-page.md) (SD-9 and SD-10 are replaced by this brief; SD-14's Move placement carries over; SD-W1's columns); [`../evidence-explained-citation-map.md`](../evidence-explained-citation-map.md) § 6.6; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide)
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
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths, unless the section opts into Page columns' full-width mode (`primaryFill`: at Wide, with no secondary, the primary fills the pane between the gutters and the header keeps its band; with `fillSplitSideMax` the primary and the band are the main-and-side group the page lays out inside it, as the Source page does); its text still keeps its measure. Wide adds columns; it does not stretch text or fields.
3. **Fill the width only with something useful.** A list, table, grid or canvas takes the extra width when Wide has more to show in it: more columns, more cards, the detail beside the list. When it has nothing more to show, it keeps a maximum width (1600pt of content for a list) and is centered, with the page gutters outside it. Empty space stretched between a row's columns is never the Wide design. Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is a rethink

Replace the Source page's "Artifact cards" frames (S15-D1 frames 9a, 9b and 12's Artifact menu). The tall card with a square preview well, its edit-on-the-card mode and its Copy details disclosure all go. Draw the new Artifacts section inside the existing Source page frames, and draw the Artifact sheet over them. Also redraw the Source's own Metadata section for one change only: its Add field becomes the inline row SA-13 describes. Every other part of the Source page stays as S15-D1 drew it.

---

## 1. Objective

On a Source page, a researcher sees each Artifact's copy details (provider, URL, accessed date on a census) at a glance, without the Artifacts section growing into a column of tall forms. The list reads like a list. Editing, and filling the type's suggestions, moves into one sheet that also adds new Artifacts.

**What changes.**

- **The list.** Each Artifact shows what it shows today (thumbnail, label, ref, description, Cover, its graphs or "Not in an evidence graph") plus its saved copy details, read-only and compact.
- **Thumbnails.** A square tile of at most 80pt, cropped like every other thumbnail in the app. The app makes thumbnails 256 pixels on the long edge; a square crop fills from the short edge, about 170 pixels on a census page, so 80pt (160 pixels at 2x) is the largest square that stays sharp. Today's cards draw them 264pt wide or more, which is why they look soft.
- **The Artifact sheet.** The edit pencil opens a sheet with label, description, the Evidence graphs the Artifact is in, and its copy details: saved values to edit, clear and reorder, the type's suggestions to fill or dismiss, and an inline row to add any other field. One Save writes everything; Cancel discards everything.
- **Add artifact** opens the same sheet in add mode, with the file.
- **Add field is an inline row** on every metadata board: in the sheet, and on the Source's own Metadata section. The Add metadata dialog goes.

**What stays.** The Artifacts section's place on the page and its header (title, count, Add artifact). The empty state. Clicking a thumbnail opens the file in the viewer, or Add file on an Artifact with no file. Set cover and the Cover badge. Delete and its confirm. The Source's own Metadata section apart from Add field: its rows still save the moment you confirm them. The Evidence graphs section.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| Artifact | One copy of the record the Source describes: a downloaded census image, a microfilm printout, a scan. It has a ref (`ART-` and five letters or digits, such as `ART-7QK2M`), a label, an optional description, an optional file, and the Evidence graphs of this Source it is in. A Source holds 1 to 60 Artifacts. |
| Copy details | The Artifact's own metadata: field–value pairs from the same field list as the Source's metadata. Every value is text; a URL field's value is a link. Most Artifacts have 0 to 4. |
| Suggestions | The Source's type suggests fields for its Artifacts. On the board's Census type they are Provider, URL, Accessed and Image number, in that order. Suggestions show only in the sheet, to fill or dismiss; the list shows saved values only, never an empty suggested field. |
| Other fields | Any field from the project's field list can be added to an Artifact with Add field, beyond its type's suggestions (for example Film number, Roll, Collection). |
| URLs | A URL value is a link in the list and in the sheet; it opens the browser. Long URLs truncate in the list. |
| Cover | One Artifact's image is the Source's cover. "Set cover" makes another image Artifact the cover; the cover shows a "Cover" badge. |
| Graphs | An Artifact is in zero or more of the Source's Evidence graphs. Zero shows "Not in an evidence graph" with "Link to a graph". In the sheet, a graph the Artifact cannot leave (a Citation on the graph is drawn on this Artifact) is checked and locked, with a hint saying why; the locks load when the sheet opens, with a loading line and a failure line. |
| Writes | In the sheet, nothing saves until Save. Editing saves label, description, graphs and copy details in one write. Adding saves the Artifact with its copy details in one write, then attaches the file. The engine judges each value: a refused value comes back naming its field. |
| Files | A file is attached when the Artifact is added (optional), or later once through Add file. A file the app refuses (wrong type, unreadable) shows a danger Callout in the sheet; the Artifact is already made, so a second Add saves onto it and its graph boxes stay as they were. |
| Scale | Up to 60 Artifacts on one Source. The list must stay scannable at 60. |

### 2.1 Copy

| Where | Text |
| --- | --- |
| Section title / button | "Artifacts" / "Add artifact" |
| Empty state | Title "No artifacts yet"; message as today; button "Add artifact" |
| No graph | "Not in an evidence graph"; button "Link to a graph" |
| Cover | Badge "Cover"; button and menu item "Set cover" |
| Sheet titles | "Edit ART-7QK2M" (the ref in the title), "Add artifact" |
| Sheet fields | "Label" (required), "Description", "Evidence graphs", "Copy details", "File" (add mode) |
| Board | "Suggested by this type"; "Add field" |
| Sheet buttons | "Save artifact" / "Add artifact" (primary), "Cancel" |
| Menu | "Edit…", "Set cover", "Move to another Source…", "Delete…" |
| Refused value | "This metadata change can't be saved." |
| Missing label | "Give the artifact a label." |
| Artifact gone | Toast, title "Artifact no longer here", message "“Sheet 14 (image 1)” was deleted or moved in another window." |

### 2.2 What this board is not

- The Move confirm (S15-D10). Here, only where **Move to another Source…** sits on an Artifact.
- The Source's own Metadata section beyond its Add field row (S15-D1).
- The Evidence graph form (S15-D5) and the graph cards. The graph form will reuse this sheet's board and scrolling; it is drawn there.
- The file viewer, and the Add file dialog for an existing Artifact (first-attach only): unchanged.
- The Source type page's suggestion lists (S15-D6).
- Filling copy details automatically from a URL ([`registry-source-import.md`](../../../ideas/registry-source-import.md)).
- Reordering saved rows by keyboard ([`keyboard-reorder.md`](../../../ideas/keyboard-reorder.md)).

---

## 3. Implementation gate

| PR | Ships | Not that PR |
| --- | --- | --- |
| **S15-34** | SA-4, SA-6, SA-8, SA-9, SA-12 – SA-15: the sheet in edit mode; the read-only copy details on today's cards; the inline Add field on every board; card edit goes. It starts from #802 as trimmed: no Copy details disclosure and no `.compact` board style remain | The list's new layout (S15-35), add mode (S15-10) |
| S15-35 | SA-1 – SA-3, SA-5 (without Move), SA-10, SA-11, SA-W1: the list | Move (S15-25) |
| S15-10 | SA-7: add mode | Filling values automatically |
| S15-25 | SA-5's Move to another Source… | The confirm (S15-D10) |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SA-1 | **The list.** One entry per Artifact, in today's order. Rows are the default; the board may draw compact cards beside them as an alternative, and the user picks one when agreeing the board. An entry's height depends on its content, never on a form; nothing is editable in place. Draw it with 12 Artifacts and with 60. |
| SA-2 | **Thumbnail.** A square tile, cropped to fill, at one size the board picks up to 80pt (the largest square the 256-pixel thumbnail fills sharply at 2x from a 3:2 page); the type icon when there is no image. Clicking it opens the file, or Add file when the Artifact has none, as today. |
| SA-3 | **What an entry shows.** Label (title), ref (mono, muted), description (at most two lines, then truncated), the Cover badge, the graphs it is in by name (the first two, then "+n"), or "Not in an evidence graph" with "Link to a graph". |
| SA-4 | **Copy details, read-only.** The saved values only, in their saved order, each as field name and value; a URL is a link. Each value is one line, truncated. An entry shows at most four values, then "+n more". An Artifact with none shows nothing for them. Draw 0, 3 and 8 values (8 shows four and "+4 more"). |
| SA-5 | **Actions.** Visible on the entry: Edit (pencil) and Delete (trash), as today; Set cover on an image Artifact that is not the cover. The entry's right-click menu: Edit…, Set cover, Move to another Source… (S15-25, between Edit… and Delete…, as S15-D1 SD-14 placed it), Delete…. Delete opens today's confirm. |
| SA-6 | **The sheet, edit mode.** Title "Edit" and the ref ("Edit ART-7QK2M"). In order: Label (required), Description (text area), Evidence graphs (today's checkboxes, the none line, and locked graphs with their hint, their loading line and their failure line), Copy details (the board: saved rows to edit, clear and reorder by drag, where a row being edited cannot be confirmed empty — its ✓ is disabled until it has text, and a value is cleared with the row's trash (user, 2026-11-13); "Suggested by this type" to fill or dismiss; the inline Add field row). The sheet is 560pt wide; its body scrolls between a fixed title and a fixed button row when tall; the saved rows lay out at full height inside that scroll. Buttons: Save artifact (primary), Cancel. The edit sheet shows no file row. |
| SA-7 | **The sheet, add mode.** Title "Add artifact". In order: File (today's drop row, optional, with its refused-file Callout under it), then the edit mode's fields. Graphs start checked, as today; after a refused file the graph boxes are locked. Copy details start as the type's suggestions, empty. Button: Add artifact, enabled once Label has text. |
| SA-8 | **Nothing saves until Save.** Every change in the sheet is a draft: a row's own confirm (✓), a filled suggestion, a dismissed suggestion, a cleared row, a reorder and an added field all change the draft and write nothing. Clearing a row needs no confirm: Cancel restores it. Cancel, Escape or closing asks to discard with the kit's standard discard confirm when the draft differs from what opened; a value typed and then put back counts as unchanged. |
| SA-9 | **Saving and failure.** While saving, Save shows its loading state and the fields are disabled. A refused copy-details value shows "This metadata change can't be saved." under its row, in the danger tone the board uses today; a missing label shows under Label; a graph the Artifact cannot leave keeps today's message naming the graph, in the top Callout; any other failure shows a danger Callout at the top of the sheet's body. The sheet stays open with everything typed. After a save the sheet closes and the entry shows the new values. |
| SA-10 | **Empty.** No Artifacts: today's empty state with Add artifact. |
| SA-11 | **VoiceOver.** An entry reads as one group: label, ref, then each copy detail as "Provider, Ancestry". The Edit and Delete buttons name the Artifact ("Edit ART-7QK2M"). |
| SA-12 | **The Artifact goes while its sheet is open.** If another window deletes or moves the Artifact, the sheet closes, the draft is lost, and the page shows the "Artifact no longer here" Toast (§2.1). |
| SA-14 | **Shortcuts into the sheet.** "Link to a graph" opens the edit sheet scrolled to Evidence graphs; "+n more" opens it scrolled to Copy details. Both land in S15-34, on today's cards first. |
| SA-15 | **Keys inside the sheet.** Escape closes an open row edit or the open Add field row first; only with none open does it close the sheet (asking to discard per SA-8). Return inside a row confirms that row, never the sheet's Save. The kit owns this order (`PVSheet` stands down while an inner edit is open), so the Evidence graph form gets it too. |
| SA-13 | **Add field is an inline row.** Every board has one Add field row, the Source's own Metadata section included. In the sheet (and on the graph form) the board ends with an "Add field" ghost button that turns into a row with a searchable field picker (fields not already on the board) and a value input, with ✓ and ✕; ✓ is disabled until a field is picked and a value typed. On the Source's own board ✓ saves at once, as its rows do; in the sheet it adds to the draft. Escape or ✕ closes the row. The Add metadata dialog goes. On the Source's own Metadata section the header's Add field button is the only one: it opens the row at the foot of the section, and there is no second button there (user, 2026-11-13: redundant with the header). In the sheet the board shows no empty state, only the row; on the Source's own Metadata section the empty state keeps its title and message, its button goes, and the row sits below it. |
| SA-W1 | **Wide.** The Artifacts section stays in the Source page's main column (S15-D1 SD-W1: main 1180pt beside the 480pt Metadata column at a 2296 pane). A row places its copy details in a column to the right of the label and description at both widths; Wide gives that column more room. Compact cards, if drawn, keep their size and add columns. The sheet is the same at both widths. |

### 4.1 Wide layout

| View | Regular (pane below 1600 pt) | Wide (pane 1600 pt and up) | What moves where |
| --- | --- | --- | --- |
| Source page Artifacts section | Rows at the main column's width below the Evidence graphs section; copy details in the row's right column | The same rows in the main column beside the Metadata side column (SD-W1); the copy details column is wider | Nothing moves between sections; the row's copy details column grows (compact cards: more columns) |
| Source page Metadata section | As S15-D1 draws it, with the inline Add field row | As S15-D1 draws it, with the inline Add field row | Nothing |
| Artifact sheet | A 560pt sheet over the page | The same sheet | Nothing: a sheet does not change with the window |

---

## 5. Suggested frames

Frames with an a / b pair are drawn at **Regular** (pane 1102 × 900) and **Wide** (pane 2296 × 1440); the others at Regular.

**Data.** The Source "1880 U.S. census, Ottawa County, Kansas, population schedule", ref `SRC-4KD2P`, type Census. Evidence graphs: "Robins household, dwelling 112", "Hale household, dwelling 113", "Ostrander household, dwelling 114". Twelve Artifacts, labels "Sheet 14 (image 1)" through "Sheet 25 (image 12)", refs `ART-7QK2M` onward, descriptions of one or two sentences on half of them; ten JPEG images, "Sheet 19" a PDF (type icon), "Sheet 24" with no file; the first is the cover; "Sheet 20" is in no graph. Copy details: `ART-7QK2M` has Provider "Ancestry", URL "https://www.ancestry.com/discoveryui-content/view/26754213:6742", Accessed "12 March 2025"; "Sheet 17" has eight values: Provider "FamilySearch", URL, Accessed "3 Jan 2024", Image number "17 of 642", Film number "1254392", Roll "392", Collection "United States Census, 1880", Archive reference "T9-0392"; "Sheet 22" has Provider "Ancestry" only. For 60 Artifacts, continue the labels to "Sheet 73 (image 60)". Invent every person and record.

| Frame | Requirement | Draw |
| --- | --- | --- |
| 1a / 1b | SA-1 – SA-5, SA-W1 | The Source page scrolled to the Artifacts section, twelve Artifacts as rows |
| 1c | SA-1 | (Only if the board offers it) the same section as compact cards |
| 2 | SA-1 | The section with 60 Artifacts, scrolled to the middle |
| 3 | SA-5 | One row's right-click menu with Move to another Source… (no action is hover-only) |
| 4 | SA-6 | The sheet in edit mode for `ART-7QK2M`: Provider, URL, Accessed saved; Image number suggested; the Robins and Hale graphs checked, Robins locked with its hint, Ostrander unchecked |
| 5 | SA-6, SA-13 | The same sheet with the URL row being edited, Image number filled, and the inline Add field row open on "Film number" |
| 6 | SA-8 | The discard confirm after Cancel on frame 5's sheet |
| 7 | SA-9, SA-12 | The sheet saving, then with "This metadata change can't be saved." under the URL row, then with the top Callout after a failed save; and the page with the "Artifact no longer here" Toast |
| 8 | SA-6 | "Sheet 17"'s sheet with eight values, scrolled |
| 9 | SA-7 | The sheet in add mode: a file dropped, the four suggestions empty |
| 10 | SA-7 | Add mode with the refused-file Callout and the graph boxes locked |
| 11a / 11b | SA-13 | The Source's own Metadata section with its inline Add field row open; at Wide it sits in the 480pt side column |
| 12 | SA-10 | The empty Artifacts section |
| 13 | SA-13 | The sheet for an Artifact with no copy details and a type with no Artifact suggestions: only the Add field row |

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Artifacts section | Snowflake | Rethink | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `SourcePageArtifactsView` | SectionHeader with count and Add artifact as today; the card grid (`LazyVGrid`, `cardMinWidth` 264) is replaced by the list |
| Artifact entry (row, or compact card if chosen) | Snowflake | Rethink (the Artifact card) | `macos/App/Features/Sources/` | Composed of kit pieces: Thumbnail, Text, Badge, Button, IconButton. A row has no hover of its own (it is not a button; its controls hover); a compact card keeps `pvLiftHover` and is a Card |
| Artifact thumbnail | Component | Ship | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail` layout `.tile` (`PVThumbnail`) | Square, cropped, at most 80pt, as `SourcesListRow` uses it at 40pt. `ArtifactSquarePreview` and the `.well` use on cards go |
| Metadata value | Snowflake | Extend (pulled out of `MetadataBoardView`'s `displayValue`) | `macos/App/Features/Sources/MetadataBoardView.swift` | One view draws a value for the board's rows and the read-only copy details, with a line limit: today's mono caption text, and for a URL today's plain link button in the link color (not `PVButton` `.link`, whose font would shift the board's row when it enters edit) |
| Copy details, read-only | Snowflake | Rethink (inside the entry; was the Copy details disclosure) | `macos/App/Features/Sources/` | Field name Text + the metadata value view; "+n more" is a `PVButton` `.link`. Not a Table, not the board |
| Cover badge | Component | Ship | `macos/App/DesignSystem/Components/Badge/` | "Cover", as today |
| Not in an evidence graph | Component | Ship | Badge + ghost Button | As today; Link to a graph opens the sheet |
| Set cover | Component | Ship | `macos/App/DesignSystem/Components/Button/` | Ghost, small; copy `L10n.Sources.useAsThumbnail` |
| Edit / Delete | Component | Ship | `macos/App/DesignSystem/Components/IconButton/` | As today |
| Entry menu | Component | Ship (first use on Artifacts) | `.pvRightClickMenu(items:)` + `PVActionMenuItem`, `macos/App/DesignSystem/Components/ContextMenu/` | As `SourcesListRow` and `SourcePageIdentityHeader` use it |
| Delete confirm | Recipe | Ship | DeleteImpact recipe, as today | Unchanged |
| Artifact sheet | Component | Extend (S15-34: an opt-in scrolling body with a fixed footer; Escape and Return yield to an inner edit, SA-15) | `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift`, `macos/App/DesignSystem/Components/Panel/PVSheet.swift` | FormDialog at 560pt with `isDirty`. The scrolling body is opt-in and the kit's, not a local ScrollView; the Evidence graph form (S15-11) is its second caller. The sheet's model follows `SourceGraphFormModel` (`Mode` create / edit, one draft) |
| Label | Component | Ship | Field + Input | Required marker |
| Description | Component | Ship | Field + TextArea | TextArea in both modes |
| Evidence graphs | Snowflake | Ship | `macos/App/Features/Sources/SourceArtifactGraphLinks.swift` `SourceArtifactGraphLinks`; `SourceFormLoadArea` | Today's checkboxes, none line, locks and their load states |
| Copy details board | Snowflake | Extend (S15-36: a draft and the add row's state; S15-34: the inline Add field) | `macos/App/Features/Sources/MetadataBoardView.swift` `MetadataBoardView` (no style: `MetadataBoardStyle` leaves with the #802 trim) | The same board as the Source's Metadata section, in the sheet |
| Inline Add field row | Snowflake | Rethink (was the Add metadata dialog, `MetadataBoardDialogs`) | `macos/App/Features/Sources/MetadataBoardView.swift` | ComboBox (field) + Input (value) + IconButtons ✓ ✕, on every board |
| File drop | Snowflake | Ship | `macos/App/Features/Sources/IngestFileDropRow.swift` `IngestFileDropRow` | Add mode only |
| Errors | Component | Ship | `macos/App/DesignSystem/Components/Callout/`; Field error | SA-9 |
| Copy details disclosure | Component | Remove (its S15-09 use; trimmed from #802 before merge) | `macos/App/DesignSystem/Components/Disclosure/` | The kit piece stays |
| Card edit mode | Snowflake | Remove | `SourceArtifactsSection` (`editingArtifactID`, `PendingEditSwitch`, `LinkEdit`, `saveFields`), `SourcePageArtifactsView.editFields` | Becomes the sheet |
| Add Artifact dialog | Snowflake | Remove (S15-10) | `SourcePageArtifactsView.addForm`, `SourceArtifactsSection.Draft` | Becomes the sheet's add mode |
| Empty state | Component | Ship | `macos/App/DesignSystem/Components/EmptyState/` | As today |

**Non-goals.** No new kit primitive: the entry, the read-only copy details and the inline Add field row are Source page snowflakes composed of kit pieces; the sheet is FormDialog with a scrolling body. No inline editing in the list. No second metadata board design: the sheet instances the existing board. No thumbnail that keeps the image's shape, and no larger thumbnail derivative.

---

## 7. Handoff

When the board is agreed: record the frames and the board's decisions (rows or compact cards, the thumbnail size, the entry's spacing, the inline Add field row's look) on the **Board** line here and in [`README.md`](README.md). S15-34's plan copies SA-4, SA-6, SA-8, SA-9, SA-12, SA-13 and the inventory and moves from `Blocked — on board S15-D14` to `Planned`; S15-35, S15-10 and S15-25 copy theirs. The brief stays in `design/` until S15-25, the last PR it gates, lands; then it moves to the spike's design archive.
