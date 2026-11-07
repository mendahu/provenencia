# S14-D8 — Citation composer: the Yes / no value editor

**Kind:** Claude Design board (**enhancement** of the observation dialog and the evidence column on the *Citation Composer* board)
**Spike:** Provenencia Spike 14 (odds and ends), batch 7 — Boolean value type (slice 7e)
**View:** the Citation composer: the observation dialog's Value field and Negative checkbox, and the saved lines in the evidence column
**Implements later as:** PR **S14-37** (Boolean value type (UI))
**Board:** drawn 2026-11-07, not agreed: the section "S14-D8 · Yes / no value" in the *Citation Composer* project, now on claude.ai as a [migrated design](https://claude.ai/artifact/3GSFQZjqdQjL8v2rpgMiDc) (frames 1, 2, 3, 4, 5, 6, 7a, 7b, 8a, 8b). Change asked 2026-11-07: remove frame 6 and the negative case (BV-4, BV-6), and drop "Checkbox (Negative, only when already checked)" from the board's Kit used note. Originally drawn on the Claude Design project [*Citation Composer*](https://claude.ai/design/p/8b120853-4ba4-4c0d-8492-75e76b9f9b7a?via=share), file *Citation Composer.dc.html*, as a new section "S14-D8 · Yes / no value", beside the S14-D6 frames.
**Depends on:** S14-36 (landed [#733](https://github.com/mendahu/provenencia/pull/733): the engine stores, reconciles and sends a `boolean` value; it reads "Yes" and "No").
**Related:** [`../deployment-plan.md`](../deployment-plan.md) working rule "(Batch 7) Boolean is a first-class value type in two PRs", bar outcome 31; dogfood note [`../ux/2026-10-25-boolean-property-value-type.md`](../ux/2026-10-25-boolean-property-value-type.md); [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); plan [`../prs/S14-37-boolean-value-type-ui.md`](../prs/S14-37-boolean-value-type-ui.md). The Properties page (offering the type, its pill, Holds) is brief S14-D9 and PR S14-39, not this board.
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

Draw on the *Citation Composer* board, beside the S14-D6 "Drawing feedback" frames, as a new section "S14-D8 · Yes / no value". Change the observation dialog's Value field and Negative checkbox, and the evidence column's saved lines, in place. Do not redraw the viewer pane, the dialog's lead (Citation picker, Property | Bridge), or the Subject and Property combo boxes. Do not start a second composer or a second dialog.

---

## 1. Objective

Census and register columns record yes/no facts: "Can read", "Can write", "Veteran", "Owns home". Spike 14 adds a value type for them (S14-36). A researcher makes a Property of that type on the Properties page (S14-39) and then cites it in the composer. This board draws how the composer takes the answer, and how a saved answer reads in the evidence column.

**What changes.**
- When the dialog's Property is a yes/no Property, the Value field holds a two-segment control, **Yes** and **No**. Neither is chosen until the researcher picks one, and Save waits until they do.
- The Negative checkbox ("The record says this is not so") is not drawn for a yes/no Property: **No** is how the record says it.
- A saved yes/no line in the evidence column reads "Can read  Yes" or "Can read  No", in the body face like a text value.

**What stays.** The dialog's layout: lead, Subject and Property side by side, Value field, error Callout, footer buttons. Every other value type's editor. The evidence column's grouping by Subject and its line layout (label, value, badges, actions). The composer's panes and their widths.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A yes/no Property | A Property whose value type is `boolean`. Its type label is "Yes / no"; its values read "Yes" and "No" everywhere in the app. Examples: Can read, Can write, Veteran, Owns home. |
| One answer | A yes/no Property holds one value per Subject. Two Sources that disagree show as mixed on the Person page; this board does not draw that. |
| No is a recorded fact | "Can read: No" is what the census says. It is not the same as a column left blank, which the researcher simply does not cite. So the control starts with nothing chosen: a click the researcher never made must not save a No. |
| Negative | Every other Property has a Negative checkbox: "The record says this is not so" (a record that denies a birth date, say). A yes/no Property already says No with its value, so the engine marks the type as one that is never negative, and the dialog hides the checkbox for it. There is no negative yes/no anywhere: the engine refuses one, and the composer always saves a yes/no as stated (user, 2026-11-07). |
| The observation dialog | Opens from the evidence column's Add and from a saved line's Edit. New and edit share the same fields. Save is the footer's primary Button; it is disabled while a required value is missing. |
| The evidence column | Lists this Citation's saved lines, grouped by Subject. Each line: Property label (secondary, 12pt), value (body, or mono for integers and dates), a "not" word and Negative Badge on a negative line, and hover actions. |
| Two widths | The composer has three panes side by side on a wide window (viewer, evidence, observations) and two on a narrower one. The dialog keeps its width at both. |

### 2.1 What this board is not

- The Properties page: offering "Yes / no" in Create, its value-type pill colour, hiding Holds several (S14-D9, S14-39).
- The Person page or Conclusion detail: a yes/no Conclusion reads "Yes" or "No" in the existing value slot (S14-36), with the existing merged and mixed marks.
- The Evidence graph canvas: a card's value line reads "Yes" or "No" in its existing text slot.
- A tri-state or "unknown" answer. A blank column is not cited.
- Keyboard shortcuts for Yes and No.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S14-37** | BV-1 … BV-7 and BV-W1, at Regular and Wide, light and dark. |
| Ships in **S14-39** | The Properties page (brief S14-D9). |
| Stays out | Any engine change (S14-36). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| BV-1 | **Yes / No control.** When the dialog's Property is a yes/no Property, the Value field ("Value") holds the kit Segmented control with two segments, "Yes" then "No", at the small size the dialog's other value inputs use, as wide as the Value field. It sits where the text input sits for a text Property. |
| BV-2 | **Nothing chosen at first.** On a new observation, and whenever the Property changes to a yes/no Property, neither segment is selected. Clicking one selects it; clicking the other moves the selection. There is no way back to nothing chosen except changing the Property. |
| BV-3 | **Save waits.** The footer's Save Button shows its disabled state until Yes or No is chosen, as it does for an empty text value today. No hint, Callout or caption is added for it. |
| BV-4 | **No Negative checkbox.** For a yes/no Property the Negative checkbox and its description are not drawn; the dialog closes up the space. There is no exception: a yes/no is never negative. |
| BV-5 | **Edit opens on the saved answer.** Edit on a saved yes/no line opens the dialog with that answer's segment selected and Save enabled. |
| BV-6 | **Saved line.** In the evidence column a yes/no line shows the Property label and "Yes" or "No" in the body face (not mono), with no Badge and no glyph. A yes/no line never has the "not" word or the Negative Badge. |
| BV-7 | **Keyboard and VoiceOver.** Tab from the Property combo box reaches the Yes segment, then No, then the dialog's next stop. Space chooses the focused segment. VoiceOver reads the group as "Value", each segment as "Yes" or "No", and the chosen one as selected. Nothing visible changes beyond the kit's focus ring; draw it as a note on frame 1. |
| BV-W1 | **Wide.** The composer at Wide is its three-pane arrangement today. The dialog keeps its Regular width in the observations column (rule 2), and the evidence column's lines keep their layout (rule 3: the column takes the width, a line does not stretch its value). Every BV requirement looks the same at both widths; draw frames 2 and 5 at Wide. |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Citation composer, observation dialog | Subject and Property side by side, the Value field below with Yes / No | The same, at the dialog's fixed width | Nothing |
| Citation composer, evidence column | Saved lines grouped by Subject | The same lines in the wider column | Nothing |

---

## 5. Suggested frames

Draw each as the composer pane only, with no app sidebar, at Regular (1102 × 900) unless it says Wide (2296 × 1440). Use a 1921 census page ("Robins household, 1921 census", image) with Subjects John Robins and Mary Robins, and two yes/no Properties, **Can read** and **Can write**, beside a text Property **Occupation**.

1. **New yes/no observation.** The dialog with Subject John Robins, Property Can read; the Value field holds Yes / No with neither selected; no Negative checkbox; Save disabled. A note for BV-7's keyboard order and spoken labels (BV-1, BV-2, BV-3, BV-4, BV-7).
2. **Answer chosen.** Frame 1 after clicking No: No selected, Save enabled (BV-2, BV-3).
3. **Text Property for comparison.** The same dialog with Property Occupation: the text input in the Value field and the Negative checkbox below it, unchanged from today. Draw it small, beside frame 1, so the difference reads (BV-4).
4. **Edit a saved answer.** Edit on John's saved "Can write  Yes": the dialog opens with Yes selected and Save enabled (BV-5).
5. **Evidence column.** John Robins: "Occupation  Carpenter", "Can read  Yes", "Can write  Yes"; Mary Robins: "Can read  No", "Can write  No" (BV-6).
6. *(Removed 2026-11-07.)* Delete frame 6 ("A negative line from elsewhere") and its note "The line itself": a yes/no is never negative, so there is no such line or dialog. Keep the other frame numbers as they are.
7. **Dark mode.** Frames 1 and 5 in dark mode.
8. **Wide.** Frames 2 and 5 at Wide (BV-W1).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Value field | Component | Ship | `PVField` in `Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationPropertyFields.valueControls` | BV-1. Same label "Value". |
| Yes / No control | Component | **Extend** | `DesignSystem/Components/Chip/PVSegmentedControl.swift` | BV-1, BV-2, BV-7. Two items, small size, no selection until chosen. Gains the kit's field focus handle (the one `Input`, `Select` and `ComboBox` take) so Tab lands on its first segment; nothing visible changes. **Yes / No → Segmented control, not a Checkbox, Switch, Radio pair or Select.** |
| Property observation form | Snowflake | **Extend** | `CitationComposerAuthoringBox.swift` `CitationPropertyFields` | BV-1, BV-4. One new value case; the Negative checkbox is conditional. Not a second form. |
| Negative checkbox | Component | Ship | `PVCheckbox` in `CitationPropertyFields` | BV-4. Never drawn for a yes/no Property. |
| Dialog Save | Component | Ship | the observation dialog's Save `PVButton` | BV-3, BV-5. Disabled state only. |
| Saved line | Snowflake | Ship | `Features/CitationComposer/CitationComposerSubjectGroups.swift` line row | BV-6. Value text in the body face. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A Checkbox or Switch for the value | An unticked box cannot tell "No" from "not answered yet", and it would sit beside the Negative checkbox (BV-2, BV-4). |
| A default selection | A No the researcher never chose would be saved as a recorded fact (BV-2). |
| A "Clear" or "Unknown" segment | A blank column is not cited; changing the Property clears the answer (BV-2). |
| A hint, Callout or caption saying "choose Yes or No" | Save's disabled state says it, as for every other empty value (BV-3). |
| A coloured, iconed or badged Yes / No in the evidence column | The value reads as text like every other value (BV-6). |
| A new kit primitive (a "boolean field" or toggle group) | One call site; the Segmented control already does it. Its focus handle is an invisible extension, not a new control. |
| Any change to the composer's panes at Wide | BV-W1. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the frame numbers, and every change from this brief with the user's answer.
2. Refresh the [S14-37 plan](../prs/S14-37-boolean-value-type-ui.md): copy the answers into its Decisions and **Design → macOS**, describe each frame in words, run `review-plan` again, and set it to `Planned` once S14-36 and S14-39 have landed.
3. S14-37 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md).
