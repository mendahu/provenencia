# S15-D13 — Copy Source dialog: carry the published date

**Kind:** Claude Design board (**enhancement** of the Copy Source dialog as shipped in Spike 12: a FormDialog, title and type fields, then "Copy from the original" with three part ticks and the Artifact checklist. Add one tick for the Source's date; nothing else on the form changes)
**Spike:** Provenencia Spike 15, slice 1
**View:** the Copy Source dialog (opened from the Source page header's Copy Source button and from a Sources list row's right-click Copy Source)
**Implements later as:** PR **S15-32** (CS-1 – CS-6, CS-W1)
**Board:** — (not drawn yet). Draw it on the Source page board ([S15-D1, validated 2026-11-10](https://claude.ai/artifact/443NPNYmGErAmxgV1fBsFK)) as a new turn "S15-D13 · Copy Source", since the dialog opens over that page
**Depends on:** S15-02 (landed [#762](https://github.com/mendahu/provenencia/pull/762)): every Source carries one optional date; S15-03 (landed [#766](https://github.com/mendahu/provenencia/pull/766)) and S15-04 (landed [#774](https://github.com/mendahu/provenencia/pull/774)): the date is shown on the Source page and the Sources list, and its one name is **Published** (S15-04 D-14)
**Related:** [`../source-provenance-date.md`](../source-provenance-date.md); board [S15-D1](S15-D1-source-page.md) (the Source page the copy opens on); [S15-D2](archive/S15-D2-sources-list.md) SL-1 – SL-3 (how the published date reads); [`../../../source-layer-data-model.md` § 4.2](../../../source-layer-data-model.md#42-copy-source); [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); working rules in [`../deployment-plan.md`](../deployment-plan.md#working-rules); plan [`../prs/S15-32-copy-source-date.md`](../prs/S15-32-copy-source-date.md)
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

### This brief is an enhancement

Draw the Copy Source dialog as it ships today (described in § 1) over the Source page frames, as a new turn "S15-D13 · Copy Source". Keep the dialog's title, subtitle, width, buttons, the title and type fields, the divider, the "Copy from the original" heading, the three part ticks with their hints, the Artifact checklist, and the loading, read-failed and save-failed states. Add one tick, **Published date**, to the parts.

---

## 1. Objective

Copy Source makes a new Source from an existing one, so a researcher starts a sibling record (another issue of a newspaper, another page of a register) from what they already typed. Today the dialog is a FormDialog 520pt wide titled "Copy Source", with the subtitle naming the original. From the top: the title field (prefilled "Copy of {original title}") and the type Select (prefilled with the original's type); a divider; a micro-caps heading "COPY FROM THE ORIGINAL"; three ticks, each a checkbox with a label and a one-line hint under it: **Description** ("The original's description, as written"), **Metadata** ("Values and field layout"), **Credibility** ("Grade and argument"); a part the original does not have is unticked and disabled with the hint "None on this Source"; then the Artifact checklist (every Artifact of the original, each with a thumbnail and its label, all ticked). While the original's parts load, a loading line stands where the ticks go; a failed read shows a Callout with Read again. The buttons are Cancel and **Copy**. A successful copy closes the dialog and opens the copy's Source page.

Since Spike 15 a Source has its own date: when the record was made or first published, named **Published** everywhere the app shows it ("Published 1 Jun 1880" on the Sources list, the date in the Source page header). Today the copy never carries it: every copy starts undated.

**What changes.** A fourth tick, **Published date**, whose hint is the original's date as the list writes it ("1 Jun 1880", "abt 1895", "bet 1834 and 1840", or the researcher's phrase "in the reign of Victoria"). It is ticked when the original has a date; ticked, the copy gets the same date as its own. An undated original shows the tick unticked and disabled, with "None on this Source", like the other parts.

**What stays.** Everything else on the dialog, its width at both widths, and what the copy carries otherwise. The Private switch is never carried: a copy always starts public, and the dialog says nothing about it.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The Source date | One optional structured date per Source: the original publication, or when the record was made. Named **Published** in the app. Set on the Source page only |
| How a date reads | As the Sources list writes it: "1 Jun 1880", "abt 1895", "bef 3 Mar 1830", "bet 1834 and 1840", or the researcher's phrase ("in the reign of Victoria") |
| A carried date is the copy's own | The copy gets a date equal to the original's, not a link to it. Changing or clearing one Source's date later leaves the other's alone |
| Why tick it | A sibling issue or page usually shares the date (two pages of one census); a different issue usually does not (next week's newspaper). The researcher unticks it, or edits the copy's date on its page afterwards |
| The date is not edited here | The dialog only chooses whether it carries. The copy's date is changed on the copy's Source page |
| Private | Not carried, not shown here (working rule "Not public is a stored switch and nothing else") |

### 2.1 What this board is not

- Editing the date in the dialog, or a date field on the form (the date is set on the copy's page).
- A Private tick, or any note about privacy.
- Ticks for notes, Subjects, Citations or graphs (never copied, [`source-layer-data-model.md` § 4.2](../../../source-layer-data-model.md#42-copy-source)).
- The Add Source dialog (it takes no date).
- The Source page (S15-D1) and the Sources list (S15-D2), except as the backdrop the dialog sits on.

---

## 3. Implementation gate

| PR | Ships | Not that PR |
| --- | --- | --- |
| **S15-32** | CS-1 – CS-6, CS-W1 | A date field on the form (not planned); carrying the Private flag (refused) |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CS-1 | **The Published date tick.** A checkbox labeled "Published date" among the part ticks, the same kit Checkbox as Description, Metadata and Credibility, same spacing. Its hint under the label is the original's date alone, as the Sources list writes it, in the hint's muted style: "1 Jun 1880". No icon, no "Published" prefix in the hint (the label already says it). Recommended place: first, above Description, because the date is the record's own fact like the title and type above it; the board may place it elsewhere among the ticks and records why |
| CS-2 | **Default.** Ticked when the original has a date, as every available part starts ticked |
| CS-3 | **Undated.** An undated original shows the tick unticked and disabled, with the hint "None on this Source", exactly as an absent Description, Metadata or Credibility does |
| CS-4 | **A long date.** A long phrase date ("in the reign of Victoria, before the mill fire of that year") wraps in the hint like any long hint; the dialog keeps its width |
| CS-5 | **States.** The tick sits inside the parts area with the other three, so it is absent while the parts load and while the read-failed Callout shows, and it is disabled while the copy saves. Ticking or unticking it marks the form changed, so Cancel on a changed form asks as it does today |
| CS-6 | **VoiceOver and keyboard.** The tick reads "Published date, 1 Jun 1880, checkbox, checked" (undated: "Published date, None on this Source, dimmed"). Tab reaches it in visual order |
| CS-W1 | **Wide.** The dialog keeps its 520pt width and is centered over the pane at both widths: a short form has nothing to add at Wide (rule 1 "center its content"), and its fields keep their measure (rule 2) |

### 4.1 Wide layout

| View | Regular (pane below 1600 pt) | Wide (pane 1600 pt and up) | What moves where |
| --- | --- | --- | --- |
| Copy Source dialog | FormDialog 520pt wide, centered over the Source page | The same dialog, 520pt, centered over the Source page at Wide | Nothing: a short form centers (rule 1) |

---

## 5. Suggested frames

Draw the dialog over the Source page frames of this board (the page dimmed behind the dialog), at **Regular** (pane 1102 × 900) for every frame and at **Wide** (pane 2296 × 1440) for frame 1. Data: invented. The original is "Ottawa County Herald, 4 July issue", a Newspaper, dated `4 Jul 1881`, with a description, three metadata values, a credibility grade and two Artifacts ("Front page", "Page 3").

| Frame | Requirement | Draw |
| --- | --- | --- |
| 1a / 1b | CS-1, CS-2, CS-W1 | The dialog for the dated original, every part ticked: Published date ("4 Jul 1881"), Description, Metadata, Credibility, both Artifacts |
| 2 | CS-3 | An undated original ("Notes on the old mill" with no date): Published date unticked and disabled, "None on this Source" |
| 3 | CS-4 | An original whose date is the phrase "in the reign of Victoria, before the mill fire of that year": the hint wraps |
| 4 | CS-1, CS-5 | Frame 1a with Published date unticked by the researcher (the next week's issue): the form is changed, Copy enabled |

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Copy Source dialog | Component | Ship | `macos/App/DesignSystem/Components/FormDialog/` `pvFormDialog`, used by `macos/App/Features/Sources/SourceCopyDialog.swift` `SourceCopyDialogModifier` | Unchanged: 520pt, title, subtitle, Cancel and Copy, dirty-close confirm |
| Copy form body | Snowflake | Extend | `macos/App/Features/Sources/SourceCopyDialog.swift` `SourceCopyForm` `partCheckboxes` | Gains the Published date tick beside the other three |
| Part tick | Component | Ship | `macos/App/DesignSystem/Components/Checkbox/PVCheckbox.swift` `PVCheckbox` | The Published date tick is a PVCheckbox with a label and a description, like the other three. The date hint is a value, not fixed copy: the `verbatim:` initializer, as `PVChecklistRows` does. Not a new "date tick" control |
| Published date text | Snowflake helper | Ship | `macos/App/Features/Sources/SourcesRowDate.swift` `SourcesRowDate.line` | The one place a Source's date becomes its row text; the hint reads it |
| Parts load area | Snowflake | Ship | `macos/App/Features/Sources/SourceFormLoadArea.swift` `SourceFormLoadArea` | Unchanged: loading line, read-failed Callout with Read again |
| Artifact checklist | Recipe | Ship | `macos/App/DesignSystem/Recipes/ArtifactChecklist/` `PVArtifactChecklist` | Unchanged |
| Save failed | Component | Ship | `PVCallout` | Unchanged |

**Non-goals.** No new kit component and no new token: the tick is a fourth PVCheckbox. No date field, no date picker, no Private tick, no icon on the date.

---

## 7. Handoff

When the board is agreed: record the link, the frames and the board's decisions on the **Board** line here, in [`README.md`](README.md) and on S15-32's plan **Board** line; the plan copies CS-1 – CS-6 and CS-W1 and moves from `Blocked — on board S15-D13` to `Planned`. When S15-32 lands, the brief moves to the spike's design archive.
