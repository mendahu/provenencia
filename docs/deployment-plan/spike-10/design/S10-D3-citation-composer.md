# S10-D3 — Citation composer: observation list

**Kind:** Claude Design board (**rethink** of the observation list, the connection list, and how a citation is chosen)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Citation composer — the form. The viewer stays.
**Implements later as:** PR **S10-16**
**Depends on:** —
**Related:** [`../deployment-plan.md`](../deployment-plan.md) (S10-16). Depiction rows are **S10-D15**, a later board. Name parts are **S10-D4**. The date modal already shipped (S10-13).
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

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

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

### This brief is a RETHINK of the lists and the citation chooser

Replace the frames for observations, connections, and the control that picks the citation. Keep the viewer, the transcription and description fields, Done, and the leave guard. Draw them again only where a new arrangement has to show them beside the list.

The requirements below are outcomes. Section 5 is a set of seeds. A layout that meets the requirements and ignores a seed is a successful board.

---

## 1. Objective

Entering observations on a real citation is hard to scan, and adding one is disorienting. This board makes a citation of about 20 observations readable, makes adding the next one obvious, and stops the citation chooser from firing transcription actions.

**What changes.** How saved observations are shown. Where a new observation or a new connection is authored. Where keyboard focus goes when authoring starts. How the researcher tells citations apart, and where Auto Transcribe and Paste sit relative to that chooser.

**What stays.** One citation is still the document: it lists every observation on it, each observation names a subject, and an empty citation may be saved. Writes stay separate: citation fields, one observation, and one connection are three different saves. Nothing deletes an observation because it was absent from a payload. The viewer (image or PDF, region, page) stays. Date and name editors stay the editors they are. Term fields stay a searchable combo box.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The place is viewer \| form | Wide at 1500pt: the form's left column is citation fields, the right column is connections and observations. Narrow: one form column, fields then connections then observations. Design both. |
| Two lists, one researcher action | A **connection** is a bridge (participation, relationship, location, place relationship): a sentence, and a role or type when the bridge has one. A location has no term. A **property observation** is subject + property + value. Today a new connection is inserted at the top of the connection list. **Add observation** appends a draft at the bottom of the property list. That split is the inconsistency to remove. |
| A saved property row is a full form | Subject combo, property combo, value editor, status, save, revert, and an actions menu (negate, remove). Text and integer edit inline. A term is a combo box plus **Add custom term**. Name and date open the existing editors. After save, the researcher rarely changes the row. The form chrome is why 20 rows do not scan. |
| Focus is stored and not applied | Add sets which row is focused. Keyboard focus does not enter the property search, so the researcher clicks before typing. |
| The citation chooser is a select | Closed label is the citation ref, or **New citation**. Each menu row is the ref, an observation count, and up to two lines of transcription. A citation with no transcription is only a ref and a count. The artifact switcher, when the source has several files, is a different control. |
| Auto Transcribe and Paste share one slot | Images show **Auto Transcribe**. PDFs show **Paste transcription**. Both sit on the trailing edge of the transcription field. The citation select sits on the trailing edge of the line above. Reaching for the select often hits the button. Overwrite still asks before replacing a transcription. |
| Subjects are the cards on this Source | Person, event, and place. Several observations often name the same subject, so the list repeats that name. |
| Negation is part of the reading | A negative observation is evidence. A condensed row still shows that it is negative. |
| Leaving with unsaved work asks | Dirty citation fields, a touched observation or a non-empty draft, or a touched connection hold navigation. Discard continues. Keep editing stays. |
| Twenty is the bar | A citation with 70 observations already opens quickly. This board is the layout at 20, not a performance pass and not a split that shows the form before the file loads. |

### 2.1 What this board is not

- The depiction row. That is S10-D15, drawn later, on its own board.
- The name-parts modal (S10-D4 / S10-17) and the date modal (already shipped). If editing opens them, it opens them as they are.
- An adaptive choice field. A term stays a combo box. Short lists do not become radios.
- The Evidence graph, except that a new connection can arrive here from Connect on the graph.
- A second composer. Work on this form.

---

## 3. Implementation gate (S10-16)

| Ships in **S10-16** | Stays out |
| --- | --- |
| The list, the authoring place, focus into the first field, and a citation chooser that does not fire Auto Transcribe or Paste. Confirm-before-overwrite stays. Term fields stay combo boxes. Comfortable at 20 observations. | Depiction row (S10-34). Name-parts modal. Date-modal internals. A new viewer. A faster open. An adaptive choice control. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CC-1 | At about 20 observations, each one is readable as subject, property, value, and negation, and it is clear where one ends. A saved observation does not have to keep its editing controls on the row. |
| CC-2 | Adding a connection and adding a property observation is one kind of action: the same place, and the same resulting position. After reload, the order on screen matches the order that was saved. |
| CC-3 | The observation or connection being authored has one obvious home. The researcher does not scroll to find the draft. |
| CC-4 | Starting that authoring moves keyboard focus into the first field they type. Recording a focused row without moving keyboard focus does not meet this. |
| CC-5 | Choosing a citation does not start Auto Transcribe or Paste. Both actions stay available. Each still confirms before it replaces a transcription. |
| CC-6 | Citations on this artifact are distinguishable when one of them has no transcription. A ref and an observation count are not enough for that citation. |
| CC-7 | The layout is comfortable at 20 observations in the narrow form and in the wide form. |
| CC-8 | A term property keeps a searchable combo box and **Add custom term**. |
| CC-9 | Citation fields, each observation, and each connection still save on their own. An empty citation can still be saved. Leaving with unsaved authoring still asks. |

---

## 5. Seeds

These are starting points from dogfooding. Use one, combine them, or draw something else that meets section 4.

**A. Condense by subject.** Group saved observations under the subject card they belong to, so the name is not repeated on every row. A saved observation is display text. Editing opens a sheet (`FormDialog`). Name and date already open a dialog; this seed extends that to the whole observation. A denser inline row that still scans also meets CC-1.

**B. One authoring area.** The draft lives in a dedicated place. Saving moves it into the list. The researcher always looks at the same spot for the thing they are creating, which is also where focus goes (CC-3, CC-4). This covers a property observation and a connection that just arrived from the graph. Putting every new row at the top of one mixed list was an earlier guess. It is not required.

**C. A citation chooser that shows the reading.** The select is hard to use because the ref does not say what the citation contains, and an empty transcription leaves nothing else to read. Something more visual is welcome: a snippet when it exists, and a clear empty when it does not, enough to pick the right citation without opening each one. The board chooses the control. Moving Auto Transcribe and Paste off that control's corner is the part that meets CC-5.

---

## 6. Suggested frames

Draw the board's answer to each situation. Include narrow and wide.

1. **Twenty saved observations** across a handful of subjects, including two on the same person and one negated. The researcher can read them without opening an editor.
2. **Authoring a property observation.** Focus is in the first field. The draft's home is obvious.
3. **A connection just arrived** from Connect on the graph. It is authored in that same home, then saved into the list.
4. **Editing a saved observation** (text, a term, and a date — the date opens the existing editor).
5. **Choosing a citation** among three: one with a transcription, one with none, and New citation. Auto Transcribe (image) and, in a second frame, Paste (PDF) are on screen and are not what a click on the chooser hits.
6. **Empty list** on a new citation.
7. **Leave** while the draft is dirty: the existing confirm.

---

## 7. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Property observation row | Snowflake | **Rethink** | `Features/CitationComposer/CitationComposerObservationRow.swift` | Saved rows scan (CC-1). Editing controls follow the board. |
| Connection row | Snowflake | **Rethink** | `Features/CitationComposer/CitationComposerConnectionRow.swift` | Same authoring home as a property observation (CC-2). Still a bridge write. |
| Form pane | Snowflake | **Rethink** | `Features/CitationComposer/CitationComposerFormPane.swift` | Where the lists, the authoring home, and the citation chooser sit. Narrow and wide. |
| Add observation | Component | Ship | `PVButton` | Starts a property observation. |
| Subject and property pick | Component | Ship | `PVComboBox` | The fields inside authoring. |
| Term pick | Component | Ship | `PVComboBox` + ghost `PVButton` | CC-8. Add custom term stays a separate small dialog. |
| Text and integer value | Component | Ship | `PVInput` | |
| Name and date value | Snowflake | Ship | existing observation dialog | Open them. Do not redraw their internals. |
| Whole-observation edit | Recipe | Extend | `pvFormDialog` | Use this if editing leaves the row (seed A). |
| Negated mark | Component | Ship | `PVBadge` | Visible on a saved row. |
| Row actions | Component | Ship | `PVContextMenu` | Negate and remove. |
| Row status | Component | Ship | `PVBadge` | Draft, edited, saving, error. A saved ref may stay as quiet text. |
| Citation chooser | Component | **Rethink** | `PVSelect`, `CitationComposerCitationMenuRow` | CC-5 and CC-6. Stay on Select if the row content is enough. |
| Auto Transcribe / Paste | Component | **Rethink** placement | `PVButton` on the transcription `PVField` | One slot: Auto Transcribe on an image, Paste on a PDF. Confirm before overwrite stays `PVConfirm`. |
| Transcription and description | Component | Ship | `PVField`, `PVTextArea`, `PVCheckbox` | Uncertain checkbox and note stay. |
| Section title and count | Component | Ship | `PVSectionHeader` | |
| Empty list | Component | Ship | `PVEmptyState` | A new citation with nothing on it yet. |
| Group or row chrome | Component | Ship | `PVCard` | A subject group or a row, if the board uses one. |
| Subject kind | Component | Ship | `PVMark` | Person, event, place, if a group header needs a kind. |
| Errors | Component | Ship | `PVCallout` | |
| Leave and overwrite | Component | Ship | `PVConfirm` | |
| Viewer | Snowflake | Ship | `Features/CitationComposer/CitationComposerViewerPane.swift` | On the board only so the form has its neighbor. Do not redesign it. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit observation-row, subject-group, or citation-picker primitive | One call site. Compose Card, Field, Select, ComboBox, Badge. |
| A depiction or likeness row | S10-D15. |
| Radios in place of a term combo box | Descoped with S10-14. |
| A second form column that exists to hide loading | Open speed is already fine. |

---

## 8. Handoff

1. Archive this brief under `design/archive/` when the board is agreed.
2. Record it in [`../completed.md`](../completed.md).
3. Implement **S10-16** against the board and this inventory.
