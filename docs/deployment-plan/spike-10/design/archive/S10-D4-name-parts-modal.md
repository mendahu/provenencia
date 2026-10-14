# S10-D4 — Name parts modal

**Kind:** Claude Design board (**enhancement** of the name editor)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Name parts modal — the shared name editor form
**Implements later as:** PR **S10-17**
**Depends on:** —
**Related:** [`../../deployment-plan.md`](../../deployment-plan.md) (S10-17). The citation composer is **S10-D3** and already shipped. The date modal already shipped (S10-13). Name-format profiles already shipped (S10-29) and do not change this split.
**Design system layers:** [`docs/design-system-layers.md`](../../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../../.cursor/skills/add-ui-component/SKILL.md)

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

### This brief is an enhancement

Add Split and a stable part-type width to the name editor that exists. Do not draw a second editor. Do not throw away the full-form field, the parts list, Add part, reorder, remove, or the dialog chrome.

---

## 1. Objective

Typing a whole name and then building the parts by hand is the slow step. The type menu also changes width when the selection changes, so the row jumps. This board puts a split of the full-form line onto the editor, and holds the type control at one width.

**What changes.** Where Split sits. The confirm that runs when Split would throw away parts the researcher edited. The type control’s width, sized to the longest label so Surname prefix and Surname occupy the same column.

**What stays.** One name is still a required full form plus optional ordered parts. The part types stay the compiled list, chosen with a select. Untyped is still allowed. Saving is still the only write. The dialog is still Add name / Edit name. The same form body is what the citation composer embeds; this board does not redraw that composer.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The editor is one form, two hosts | `NameValueEditorForm` is the body. The modal wraps it in a form dialog at 560pt (Add name / Edit name, Save or Add, Cancel). The citation composer embeds that same body in the authoring well, with no second dialog. Draw the modal. Draw the same body once in that narrower well so Split and the type column still fit. Do not redraw the composer. |
| The full form is the name as read | One line. Required. Placeholder today is “John William Alderwick”. The hint says the source’s own wording stays on the citation. Parts are optional. |
| Parts are a list under that line | Each row is a value, a type select, move up, move down, and remove. Add part appends a row. Empty parts are a short dashed note (“No parts — the form stands alone.”), not a page empty state. Order is the order of the rows. |
| The type list is eight labels | No type, Prefix, Given name, Nickname, Surname prefix, Surname, Suffix, Undetermined. **Surname prefix** is the longest. The control is a select. Researchers cannot add a type. |
| The type column jumps today | The closed select sizes toward the selected label. Surname prefix is wider than Surname, so changing the selection moves the value field and the row buttons. A fixed frame was tried; the board’s job is a width that holds the longest label and does not change with the selection. |
| Letter keys already belong to Select | With the select focused, a letter jumps to the next label that starts with it. S goes Surname prefix, then Surname, then Suffix, then back around. That cycle is Select, for every select in the app. This board does not draw a name-only keyboard control. It does draw the type column wide enough that those three labels sit in the same width. |
| A long label elsewhere truncates | Other selects ellipsize inside the width they were given, and help shows the full string. These eight labels are short. This field is wide enough that they are not cut. Do not apply “as wide as the longest label” to every select. |
| Split is Western and fixed | A comma means the surname is first. Otherwise the last word is the surname and the earlier words are given names. See NP-3. The rules do not read a name-format profile, and they do not notice a suffix, a particle, or a nickname. |
| Blur fills only an empty list | Leaving the full-form field creates parts only when there are none. If any part is already there, blur does nothing, including when those parts were typed by hand. There is no “suggested” mark. |
| Split refills | Split always reapplies the rules to the current full form. If any part was edited by hand, confirm first. Parts that exist only because blur or Split just filled them are not hand-edited until the researcher changes a value, changes a type, adds, removes, or reorders. |
| Confirm is a replace, not a delete | Replacing hand-edited parts uses Confirm with an item (the split that would be applied), not a Bool, and not a delete-impact check. Nothing in the catalog is removed. Cancel keeps the edited parts. The dialog’s existing dirty-leave confirm stays as it is. |
| Nothing is written until save | Split and blur change the draft. The catalog write is still Save name / Add name on the dialog, or saving the observation in the composer. |

### 2.1 What this board is not

- A smarter parser. No suffix, particle, nickname, or “Jr.” detection. No profile-aware split. Spanish dual surnames and patronymics stay out.
- A combo box, a searchable type list, or **Add type**.
- A suggested-state, a guess badge, or a second readout of “what we split.”
- Sizing every select in the app to its longest option.
- A redraw of the citation composer, the date modal, or the Evidence graph.
- Reading the citation transcription. The line that splits is the full-form field the researcher can edit.

---

## 3. Implementation gate (S10-17)

| Ships in **S10-17** | Stays out |
| --- | --- |
| Blur fills an empty part list. Split refills, and confirms when a part was hand-edited. The type control stays a select at a stable width sized to the longest label. Same-letter cycling stays Select’s own behavior, on every select. The Western rules in NP-3. The form body is shared, so the composer embed gets the same Split and the same width. | Profile-aware splitting. A combo box for part type. User-minted part types. A kit primitive for Split. Applying the longest-label width to every select. Parsing the citation transcription. A new name editor. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| NP-1 | Leaving the full-form field fills parts only when the part list is empty. It uses NP-3. If any part is already there, blur does nothing. There is no suggestion chrome. |
| NP-2 | Split is an explicit action on this form. It refills parts from the current full form using NP-3. If any part was hand-edited, Confirm asks before replacing. Cancel keeps the edited parts. A list that is empty, or that was only filled by blur or Split and not touched since, refills without asking. |
| NP-3 | Western rules, and nothing else. If the trimmed form contains a comma, the text before the first comma is one Surname part (spaces inside it stay in that one part), and the words after that comma are Given parts in order. Otherwise the words split on whitespace: the last word is Surname and each earlier word is Given. `Smith, John Henry` → Surname, Given, Given. `John Henry Smith` → Given, Given, Surname. `De la Cruz, Maria Elena` → Surname “De la Cruz”, Given Maria, Given Elena. A form with no word does not create parts. A suffix, a particle, and a nickname are not detected: the last word of `John Smith Jr` is Surname. |
| NP-4 | The type control stays a select of the eight labels in §2. It does not become a combo box. The researcher cannot add a type. No type (empty) stays a choice. |
| NP-5 | The type control’s width is the width of the longest label, **Surname prefix**, plus the select’s chevron. It does not change when the selection changes. The value field and the move/remove controls stay put from Surname prefix to Surname to Suffix to No type. The label is fully visible. Do not give every other select this width. |
| NP-6 | With the type control focused, pressing the same letter again cycles labels that start with that letter. S moves Surname prefix → Surname → Suffix → back to Surname prefix. Draw those three closed labels in one column width. Do not draw a separate keyboard widget. The cycle is Select, not a name-modal special case. |
| NP-7 | The full form stays required. Parts stay optional. An untyped part stays valid. Row order stays the order on screen. Add, move, and remove stay. The stored-as readout stays on the dialog. |
| NP-8 | Split and blur change the draft only. They do not write the catalog. |

---

## 5. Seeds

Placement only. The rules in section 4 are fixed.

**A. Split beside the full form.** A ghost or secondary button on the full-form row, so the action is next to the line it reads. Add part stays at the bottom of the parts list.

**B. Split in the parts heading.** The heading row gains Split next to the part count, and the empty-parts note still offers it. Use this if the full-form row should stay a single field.

Either seed meets NP-2. A form that hides Split once parts exist does not.

---

## 6. Suggested frames

Draw the modal at 560pt. The last frame is the same form in the composer’s authoring well. Do not draw the rest of the composer.

1. **Empty parts, a full form.** “John Henry Smith” is in the field. Split is visible. Parts are still the dashed empty note.
2. **After leaving that field.** Given John, Given Henry, Surname Smith. No suggestion chrome. The type column is one width on every row.
3. **Comma form.** Split on “Smith, John Henry” shows Surname Smith, Given John, Given Henry.
4. **Multi-word surname.** “De la Cruz, Maria Elena” is one Surname part, then two Given parts. The type column does not grow for the longer value; the value field takes the extra text.
5. **Untouched refill.** Parts were filled from “John Smith”. The researcher changes the full form to “John Henry Smith” and uses Split. No confirm. The parts become Given, Given, Surname.
6. **Hand-edited, then Split.** A part value was changed by hand. Split opens Confirm. One frame is Cancel (the edit remains). One frame is the replace (NP-3 applied). The confirm is not a trash-can delete.
7. **The type column, three labels.** Surname prefix, Surname, and Suffix, each the closed select, same width. The value field and the row buttons do not move. The label is not ellipsized.
8. **No word.** The full form is empty or only spaces. Split does not add a part. The form still cannot be saved.
9. **The same form in the authoring well.** Narrower than 560pt. Split is still obvious. The type column is still the stable width, and the row does not overflow into a horizontal scroll.

---

## 7. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Name editor form | Snowflake | **Extend** | `Features/Names/NameValueEditorForm.swift` | The only form. Split and the type width land here. Both hosts use it. |
| Name dialog | Recipe | Ship | `Features/Names/NameValueFormDialog.swift` (`pvFormDialog`) | 560pt. Add name / Edit name. Do not redraw the dialog chrome. Dirty leave stays the dialog’s existing confirm. |
| Full form | Component | Ship | `PVInput` | NP-1 on blur. Required. |
| Split | Component | Ship | `PVButton` | NP-2. Ghost or secondary. The board places it (section 5). Not a new Split control. |
| Parts empty | Snowflake | Ship | dashed note in `NameValueEditorForm` | A section empty, not `PVEmptyState`. Split stays reachable from it. |
| Add part | Component | Ship | `PVButton` ghost | Stays at the list. |
| Part value | Component | Ship | `PVInput` | |
| Part type | Component | **Extend** this call site | `PVSelect` | NP-4, NP-5, NP-6. Stable width here only. Letter cycle is already Select’s. |
| Move and remove | Component | Ship | `PVIconButton` | Stay on the row. |
| Replace confirm | Component | Ship | `PVConfirm` (`item:`) | NP-2. Tone `irreversible`, so the button is not a trash delete. The item is the pending split. Not DeleteImpact. |
| Parts heading | Snowflake | Ship | caps heading in `NameValueEditorForm` | Keep. Do not swap in a new section-header style. |
| Stored as | Snowflake | Ship | readout in `NameValueEditorForm` | Stays on the dialog. Do not redesign it. |
| Composer host | Snowflake | Ship | `Features/CitationComposer/CitationComposerAuthoringBox.swift` | Embeds the form. Frame 9 only. Do not redraw the composer. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit Split button, name parser, or part-row primitive | One form. Compose Button, Input, Select, Confirm. |
| Longest-label width on every select | NP-5 is this field. Other selects truncate inside the width they were given. |
| A combo box or Add type | The compiled list stays a select. |
| Profile-aware or transcription-aware splitting | NP-3 is the whole parser. |
| A second name editor | The dialog and the composer share this form. |

---

## 8. Handoff

1. Archive this brief under `design/archive/` when the board is agreed.
2. Record it in [`../../completed.md`](../../completed.md).
3. Implement **S10-17** against the board and this inventory.
