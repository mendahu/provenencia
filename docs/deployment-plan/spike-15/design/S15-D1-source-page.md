# S15-D1 — Source page: the Source's date, the Private switch, Page columns, an Artifact's metadata, the citation preview and Move

**Kind:** Claude Design board (**enhancement** of the Source page as shipped: S2-18 / S2-25 page, S8 metadata board, S15-01 credibility text area. Draw on the existing Source page frames; add states and one header row, and give the page its first Wide frame)
**Spike:** Provenencia Spike 15, slices 1, 2, 6 and 7
**View:** the Source page (one Source: identity header, Description, Credibility, Metadata, Evidence graphs, Artifacts, Notes)
**Implements later as:** PR **S15-03** (SD-1 – SD-8, SD-W1, SD-W2), then S15-09 (SD-9, SD-10), S15-23 (SD-11 – SD-13) and S15-25 (SD-14)
**Board:** — (not drawn yet)
**Depends on:** S15-02 (landed [#762](https://github.com/mendahu/provenencia/pull/762)): every Source carries an optional date and an `is_public` flag
**Related:** [`../source-provenance-date.md`](../source-provenance-date.md); [`../2026-10-27-source-privacy-toggle.md`](../2026-10-27-source-privacy-toggle.md); [`../evidence-explained-citation-map.md`](../evidence-explained-citation-map.md) § 6.3, § 6.6; [`../move-artifact-between-sources.md`](../move-artifact-between-sources.md); [`../../../ideas/wide-layout.md`](../../../ideas/wide-layout.md); [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); working rules in [`../deployment-plan.md`](../deployment-plan.md#working-rules); plan [`../prs/S15-03-source-page-date-and-private-switch.md`](../prs/S15-03-source-page-date-and-private-switch.md)
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
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths, unless the section opts into Page columns' full-width mode (`primaryFill`: at Wide, with no secondary, the primary fills the pane between the gutters and the header keeps its band); its text still keeps its measure. Wide adds columns; it does not stretch text or fields.
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is an enhancement

Draw on the Source page frames as a new turn "S15-D1 · Source page", starting from the page as it ships today (described in § 1). Keep the cover thumbnail, the title's inline edit, the ref and type pill with its pencil, Copy Source and Delete, the Description and Credibility sections, the Metadata board's rows, suggestions and Add, the Evidence graph cards, the Artifact cards and their on-card edit, and Notes. Add the date and the Private switch to the header, put the page on Page columns, and draw the later states (instance the kit's **Switch**, **Disclosure** and **ClipboardButton** for them; do not draw local ones) (an Artifact's metadata, the citation preview, Move) in the same turn so one board covers the page.

---

## 1. Objective

The Source page is one Source's record. Today, from the top: a header pinned above the scroll (a 72pt cover thumbnail, the title as a large inline-editable heading, and under it a meta row: the mono ref `SRC-4KD2P`, a dot, the type pill with a pencil; Copy Source and Delete as small icon buttons at the right). Under the header, in one left-aligned column up to the page's content width: Description and Credibility on the left beside the Metadata board on the right, then the Evidence graph cards, the Artifact cards (wrapping, four across), and Notes. The page has no Wide arrangement: at 2296pt its column sits at the left with the rest of the pane empty.

**What changes.**
- **The Source's own date.** A Source has one optional date: when the record was made or first published ("1880", "abt 12 Mar 1880", "bet 1834 and 1840"). It is the date the Sources list sorts by and the date citations print. It reads as the Source's own fact, in the header beside the ref and the type, never as another Metadata row. It is set, changed and cleared with the app's structured date form (point or range; a qualifier: as stated, about, before, after; year, month, day; optional time; calendar and phrase under More).
- **The Private switch.** A Source can be marked private. One on/off switch labeled "Private", off by default. It is saved the moment it flips. It changes nothing else in the app yet: no badge, no filtering, no hiding (that comes with sharing, later).
- **Page columns.** The page moves onto the app's Page columns with its header in the pinned header slot. Regular keeps today's arrangement. Wide gets a second column (SD-W1).
- **Later on this board** (drawn now, built by later PRs): each Artifact card shows its own metadata board (provider, URL, accessed date); a citation preview shows how this Source's citations read, which fields its format expects and is missing, and Copy citation; Artifact cards and Evidence graph cards offer **Move to another Source…**.

**What stays.** Every section's content, controls and order at Regular; the header's look apart from the new row items; the Metadata board's rows and suggestion behavior; the Credibility text area (S15-01).

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The Source date | One optional structured date per Source: the original publication, or when the record was made. Not a Metadata row: Metadata keeps its own text dates ("Issue date", "Enumeration date") for filing and search. |
| How a date reads | As the app's lists and page summaries write dates (the Event page, the Events list, a Person's life line): "1880", "12 Mar 1880", "abt 1880", "bef 3 Mar 1830", "bet 1834 and 1840", or a phrase ("in the reign of Victoria"). The Sources list (S15-04) and the Person page (S15-05) will show it the same way. |
| Undated | Most Sources start undated; there is no backfill. Undated sorts last in the Sources list. |
| Private | A stored on/off fact on the Source. Off (public) for every existing Source. Nothing in the app reads it yet. A copied Source starts not private. |
| Writes | The date and the switch each save on their own, at once, with their own write. Title, type and description keep their shared Save. |
| A Source deleted elsewhere | Another window can delete the Source while this page is open. A write then finds no Source; the page says the Source was deleted and goes back to the Sources list. |
| Artifact metadata (S15-09) | An Artifact (one copy of the record: a download, a microfilm printout) has its own metadata board with the same fields and behavior as the Source's: saved rows, the type's suggestions for an Artifact (provider, URL, accessed date on a census), Add. Up to 60 Artifacts on a Source. |
| Citation preview (S15-23) | The Source's type has citation formats (full, short, source-list entry). The preview renders this Source's source-list entry and, for a chosen Citation of this Source, its full and short notes. It names the fields the format expects that are empty ("Missing: Enumeration district, Page"). A type with no format shows today's plain citation line and no Missing list. |
| Move (S15-25) | An Artifact and the Evidence graphs drawn on it move together to another Source; the confirm (its own board, S15-D10) lists the bundle. On this page the board decides only where **Move to another Source…** is offered on an Artifact card and a graph card. |

### 2.1 What this board is not

- The date dialog's form itself: it is the app's shared date form, drawn as it is today. Do not redesign it.
- The Sources list, the Person page's Sources section, Copy Source's form (S15-D2, S15-D3, S15-D13).
- The Move confirm (S15-D10) and the provenance sheet (S15-D9).
- A privacy badge, glyph, banner or explanation callout: the switch does nothing else yet.
- A date on Artifacts, Evidence graphs or Notes.

---

## 3. Implementation gate

| PR | Ships | Not that PR |
| --- | --- | --- |
| **S15-03** | SD-1 – SD-8, SD-W1, SD-W2 | Artifact metadata (S15-09), the citation preview (S15-23), Move (S15-25) |
| S15-09 | SD-9, SD-10 | The add flow's copy fields (S15-D4, S15-10) |
| S15-23 | SD-11 – SD-13 | The format editor (S15-D6, S15-17) |
| S15-25 | SD-14 | The confirm (S15-D10) |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SD-1 | **The date in the header.** The header's meta row reads: ref · type pill (pencil) · date · Private switch. Private ends the row; it does not join Copy and Delete. When the row does not fit (a long type name, a range date, the type picker open at Regular), it wraps after the type, keeping the ref first. A dated Source shows its date as text ("abt 12 Mar 1880") with a small pencil IconButton "Edit date". It sits with the ref and the type so it reads as the Source's own fact. **The board decides** the date's look (recommended: body-small secondary text with a calendar icon before it, no pill, so it does not look like the type). |
| SD-2 | **Undated.** An undated Source shows a small ghost Button "Add date" with a calendar icon in the date's place. No "Undated" text. |
| SD-3 | **The date dialog.** "Add date" and the pencil open a FormDialog titled "Source date" (subtitle: "When the record was made or first published.") holding the shared date form, prefilled with the saved date (empty for Add). Confirm "Save" is disabled while the form is blank or invalid; Cancel; a dirty form asks before it closes, as every FormDialog does. On save the dialog closes and the header shows the new date. |
| SD-4 | **Clear the date.** A dated Source's dialog has a small ghost danger Button "Clear date" under the form. It empties the form and shows a caption "Saving removes this Source's date."; Save then saves the Source undated and closes the dialog, and Cancel keeps the date. No Confirm over the dialog: its Save is the confirmation. Not offered when the Source is undated. |
| SD-5 | **A date the engine refuses.** If the save fails, the dialog stays open with its draft and shows a danger Callout above the form with the error's message; a refused date reads "This date isn't valid. Check the year, month and day." (the app-wide message for a refused date, also shown in the composer). |
| SD-6 | **The Private switch.** A Switch labeled "Private", off for a public Source, on for a private one. It saves when flipped; while the write runs it is disabled; if the write fails it flips back and the page's error Callout says why. A tooltip on hover: "Private Sources will be kept out of sharing. Nothing else changes yet." No other mark anywhere on the page. |
| SD-7 | **A Source deleted elsewhere.** A date, switch, title, type, description or cover write to a Source deleted in another window closes any open dialog, leaves for the Sources list and shows the workspace notice toast "“1880 census, Ottawa County” was deleted." (the title as it was shown). Opening a Source page whose Source no longer exists (Back to it) leaves the same way with "This Source no longer exists." |
| SD-8 | **Keyboard and VoiceOver.** Tab order in the meta row: type pencil, date pencil (or Add date), Private switch. VoiceOver reads the date as "Date, about 12 March 1880" and the switch as "Private, off". After the dialog closes, focus returns to the date's pencil (or Add date). |
| SD-9 | **(S15-09) An Artifact's metadata.** Each Artifact card shows its metadata board under its label and file: the same rows, suggestions and Add as the Source's board, at the card's width. **The board decides** whether it is always open or behind a Disclosure "Copy details (3)" on the card (recommended: Disclosure, closed, with the count of saved values, because a Source can hold 60 Artifacts). |
| SD-10 | **(S15-09) Many Artifacts.** A Source with 12 Artifacts, three with values: cards keep their size; an open board grows its card only. |
| SD-11 | **(S15-23) The citation preview.** A section "Citation" with the Source's source-list entry rendered with its italics, and a Select "For" listing this Source's Citations by label; choosing one shows its full note and short note. Each rendered form has a ClipboardButton "Copy citation" (rich text). |
| SD-12 | **(S15-23) Missing fields.** Under each rendered form, a warning Callout "Missing: Enumeration district, Page" naming the empty fields the format expects. No Callout when nothing is missing. A type with no format shows today's citation line and the caption "This type has no citation format." with a link "Edit formats" to the type's page. |
| SD-13 | **(S15-23) No Citations yet.** The Select is absent; the list entry still renders; a caption "Add a Citation to preview its notes." |
| SD-14 | **(S15-25) Move.** Each Artifact card's menu and each Evidence graph card's menu gains **Move to another Source…**, after Edit and before Delete. No other chrome. |
| SD-W1 | **Wide: two columns.** At Wide a main column at today's content width holds Description, Credibility, Evidence graphs, Artifacts and Notes; a side column beside it (340–480pt) holds Metadata and, under it (S15-23), the Citation preview. The group is centered in the pane. The header is Page columns' pinned band: it keeps the content width, centered, with the kit's rule under it and no card background of its own. The page error Callout sits at the top of the main column. Recommended; **the board decides** what goes in the side column, by the rule that it holds reference material the researcher reads beside the record. |
| SD-W2 | **Regular: as today.** At Regular, one column at the content width: Description and Credibility beside Metadata as today, then graphs, Artifacts, Notes; (S15-23) the Citation preview under Metadata in Metadata's column. The header band looks as today without its own card background. Crossing 1600pt keeps scroll, focus, an open dialog and an open inline edit. |

### 4.1 Wide layout

| View | Regular (pane below 1600 pt) | Wide (pane 1600 pt and up) | What moves where |
| --- | --- | --- | --- |
| Source page | Header pinned; one column at the content width: Description + Credibility beside Metadata (Citation preview under Metadata, S15-23); graphs; Artifacts (four across); Notes | Header pinned at the content width, centered; main column at the content width: Description, Credibility, graphs, Artifacts, Notes; side column (`widthInspector` to `widthSecondaryMax`): Metadata, then Citation preview | Metadata (and the preview) move from beside Description to beside the whole main column; Description and Credibility take the main column's full width |

---

## 5. Suggested frames

Frames with an a / b pair are drawn at **Regular** (pane 1102 × 900) and **Wide** (pane 2296 × 1440); the others at Regular. Data: a Source "1880 U.S. census, Ottawa County, Kansas, population schedule", ref `SRC-4KD2P`, type Census, eight metadata values, three Evidence graphs, four Artifacts, two Notes. Invent every person and record.

| Frame | Requirement | Draw |
| --- | --- | --- |
| 1a / 1b | SD-1, SD-6, SD-W1, SD-W2 | The page with the Source dated "1 Jun 1880", Private off |
| 2 | SD-2 | The header only: an undated Source, "Add date" in the date's place, Private on |
| 3 | SD-3 | The date dialog over the page, opened from Add date, empty |
| 4 | SD-3, SD-4 | The dialog opened from the pencil on "abt 12 Mar 1880", prefilled (point, About, 1880, March, 12), with Clear date |
| 5 | SD-4 | The dialog after Clear date: the form empty, the caption, Save enabled |
| 6 | SD-5 | The dialog with the danger Callout over a refused date |
| 7 | SD-6 | The header with the Private switch disabled while saving, and the page's error Callout after a failed flip |
| 8 | SD-7 | The Sources list with the notice toast "“1880 U.S. census, Ottawa County, Kansas, population schedule” was deleted." (toolbar not needed) |
| 9a / 9b | SD-9, SD-10 | (S15-09) Twelve Artifacts; one card's Copy details open with provider "Ancestry", URL, accessed "12 March 2025" and one suggestion |
| 10a / 10b | SD-11, SD-12 | (S15-23) The Citation preview with a Citation chosen; the full note missing "Enumeration district" |
| 11 | SD-12, SD-13 | (S15-23) The preview for a type with no format, and for a Source with no Citations |
| 12 | SD-14 | (S15-25) An Artifact card's menu and a graph card's menu with Move to another Source… |

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Page columns | Component | Ship | `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` | The page's shell: `header` with `headerPinned: true`, full-width mode; the kit draws the band and its rule. The main and side columns are the Source page's own arrangement inside it, so no section changes column when the window crosses 1600 |
| Identity header | Snowflake | Extend | `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `SourcePageIdentityHeader` | Gains the date and the Private switch in its meta row (SD-1, SD-2, SD-6). Moves into Page columns' header slot and drops its own column frame, card background and rule. The date is text through the app's row date display (`DateRowDisplay`) with an Icon: not a Chip, not a Badge, not the type pill |
| Edit date | Component | Ship | `macos/App/DesignSystem/Components/IconButton/` | Small pencil, as the type's pencil |
| Add date | Component | Ship | `macos/App/DesignSystem/Components/Button/` | Ghost, small, calendar icon |
| Date dialog | Component | Ship | `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift` | FormDialog; no bespoke sheet |
| Date form | Snowflake (shared feature form) | Ship | `macos/App/Features/Dates/DateValueEditorForm.swift` `DateValueEditorForm` | Unchanged. The composer's date entry uses the same form |
| Clear date | Component | Ship | `Button` (ghost, danger tone, small) | Inside the dialog; no Confirm (SD-4) |
| Refused-date message | Component | Ship | `macos/App/DesignSystem/Components/Callout/` | Danger tone, in the dialog |
| Private switch | Component | Ship | `macos/App/DesignSystem/Components/Switch/PVSwitch.swift` `PVSwitch` | Labeled "Private". First Features caller. Not a Checkbox (it saves at once), not a local toggle |
| Deleted notice | Component | Ship | `macos/App/DesignSystem/Components/Toast/` via the workspace notice toast | The same notice the Person page shows when its record is gone |
| Metadata board | Snowflake | Ship / Extend (S15-26, S15-09) | `macos/App/Features/Sources/SourcePageMetadataView.swift` | Moves to the secondary column at Wide. S15-09 instances it per Artifact |
| Artifact card | Snowflake | Extend (S15-09, S15-25) | `macos/App/Features/Sources/SourcePageArtifactsView.swift` | Gains the metadata Disclosure and a Move menu item |
| Copy details disclosure | Component | Ship | `macos/App/DesignSystem/Components/Disclosure/` | S15-09; not a local chevron row |
| Evidence graph card | Snowflake | Extend (S15-25) | `macos/App/Features/Sources/SourcePageGraphsView.swift` `SourcePageGraphRow` | Gains Move in its menu |
| Card menus | Component | Ship | `macos/App/DesignSystem/Components/ContextMenu/` | Move to another Source… (S15-25) |
| Citation preview | Snowflake | Extend (S15-23: the Source page gains a section) | `macos/App/Features/Sources/` | A snowflake, not a kit primitive. SectionHeader + rendered text + Select + ClipboardButton + Callout; no new kit piece |
| Copy citation | Component | Ship | `macos/App/DesignSystem/Components/ClipboardButton/` | S15-23 |
| Missing fields | Component | Ship | `Callout`, warning tone | S15-23 |
| Section titles | Component | Ship | `macos/App/DesignSystem/Components/SectionHeader/` | As today |

**Non-goals.** No new kit primitive: the date text is part of the header snowflake; the Switch, FormDialog, Callout, Disclosure, Select and ClipboardButton exist. No privacy badge or glyph. No date field inside the Metadata board. No redesign of the shared date form.

---

## 7. Handoff

When the board is agreed: record the frames and the board's decisions (SD-1's look, SD-4's placement, SD-9's disclosure, SD-W1's secondary) on the **Board** line here, in [`README.md`](README.md) and on S15-03's plan **Board** line; the plan copies SD-1 – SD-8, SD-W1 and SD-W2 and moves from `Blocked — on board S15-D1` to `Planned`. The brief stays in `design/` until S15-25, the last PR it gates, lands; then it moves to the spike's design archive.
