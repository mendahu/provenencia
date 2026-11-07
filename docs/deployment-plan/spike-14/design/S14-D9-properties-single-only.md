# S14-D9 — Properties page: a single-only value type

**Kind:** Claude Design board (**enhancement** of the Properties page on the *Properties View* board, S7-D2 Subject Fields)
**Spike:** Provenencia Spike 14 (odds and ends), batch 7 — Boolean value type (slice 7e)
**View:** the Properties page: the New property dialog's Value type and Holds fields, the inspector's Value type and Holds rows, and the table's Value type pill
**Implements later as:** PR **S14-39** (Properties page reads single-only)
**Board:** — (not drawn yet). Draw on the *Properties View* project, now a [migrated design on claude.ai](https://claude.ai/artifact/XRS4WQDZbaDa7jAzVKYQdc), page *S7-D2 Subject Fields.dc.html*, as a new section "S14-D9 · Single-only value type".
**Depends on:** S14-36 (landed [#733](https://github.com/mendahu/provenencia/pull/733): Go sends each value type's rules, `single_only` among them, with the Properties read).
**Related:** [`../deployment-plan.md`](../deployment-plan.md) working rule "(Batch 7) Boolean is a first-class value type in two PRs", bar outcome 31; [S14-D8](S14-D8-composer-boolean-value.md) (the composer's Yes / no editor, S14-37); [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); plan [`../prs/S14-39-properties-single-only.md`](../prs/S14-39-properties-single-only.md).
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

Draw on the *Properties View* board, beside the S7-D2 frames, as a new section "S14-D9 · Single-only value type". Change the New property dialog, the inspector and the table's Value type pill in place. Do not redraw the type strip, the search, the bindings or Delete. Do not start a second Properties page.

---

## 1. Objective

Spike 14 adds a value type for yes/no facts from a record: Can read, Veteran, Owns home. Its key is `boolean`, and it is **single-only**: a yes/no Property holds one value, because two different answers are a conflict to show, not two values to keep. The engine already refuses a yes/no Property that holds several values. This board makes the Properties page say so, instead of offering a choice the engine refuses.

**What changes.**
- In the New property dialog, when the chosen value type is single-only, the Holds field reads "One value", read-only, with the reason as its hint. Choosing a type that is not single-only brings back the two radios.
- In the inspector, a researcher's own yes/no Property shows Holds as the existing locked row ("One value" and the lock), and the lock says why: "Yes / no Properties hold one value".
- The value type pill and the create picker show the key `boolean` in mono, like every other type, in a pill colour of its own.

**What stays.** Every other value type's Holds radios, in the dialog and the inspector. The seeded Property's locked Holds row ("Set by Provenencia"). The pills of the other types. The table, the type strip, bindings and Delete. VoiceOver still reads a type by its label ("Yes / no").

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| Value type | What a Property's values are: `text`, `integer`, `date`, `name`, `subject` (a researcher may create these), `term` and `depiction` (seeded only), and now `boolean`, whose label is "Yes / no". The page shows the key in mono on a pill coloured per type; VoiceOver reads the label. A Property's type cannot change after it is created. |
| Holds | Whether a Property keeps one value per person, event or place ("One value: records are narrowed to one value, or shown as mixed") or every distinct value ("Several values"). A researcher sets it on their own Properties; a seeded Property's is set by Provenencia and shown locked. |
| Single-only | A value type whose Properties must hold one value. Today only `boolean` is. The engine decides which types are single-only and sends the list with the page's data, so the page never keeps its own list. |
| When `boolean` appears | A researcher can create a yes/no Property once the composer can cite one (S14-37). Until then the picker does not offer it, but the page draws it correctly if one exists. |
| The New property dialog | A FormDialog: Label, Key (read-only, from the label), Description, Value type (a segmented control of the creatable types), Holds, and Bind to subject types; Cancel and Create property. |
| The inspector | The side pane for the selected Property: label and description, then meta rows (Value type with its pill and a lock, Holds, Origin, Used on, Terms, Bound to), then Delete. |
| Two widths | The Properties page is a list with its inspector beside it. This board's changes sit inside the dialog and the inspector and do not move between widths. |

### 2.1 What this board is not

- The composer's Yes / no editor (S14-D8, S14-37).
- Showing value type labels ("Text", "Yes / no") in place of keys: every type keeps its mono key (user, 2026-11-07).
- Changing a Property's value type, or its Holds, from anything other than the inspector's existing radios.
- Any other single-only type: the rule is drawn generally, and boolean is its one case today.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S14-39** | SO-1 … SO-6 and SO-W1, at Regular and Wide, light and dark. |
| Ships in **S14-37** | Offering `boolean` in the create picker (Go marks it creatable); the composer (S14-D8). |
| Stays out | Any engine change. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SO-1 | **Create: Holds for a single-only type.** When the dialog's Value type is single-only, the Holds field shows "One value" as read-only text with the lock glyph the inspector's locked row uses, and its hint reads "Yes / no Properties hold one value" (the type's label, then "Properties hold one value"). The two radios are not drawn. |
| SO-2 | **Create: switching type.** Choosing a type that is not single-only brings back the two radios with "One value" selected, and the hint "Facts that change over time, like occupation or residence, usually belong in events". Choosing a single-only type after "Several values" was selected sets Holds back to one value. Nothing else in the dialog changes or moves. |
| SO-3 | **Inspector: a researcher's own yes/no Property.** Holds is the existing locked row: "One value" and the lock. VoiceOver reads the lock as "Yes / no Properties hold one value" (a seeded Property's lock keeps "Set by Provenencia"). No radios, no "reconciled again" notice. |
| SO-4 | **Inspector: everything else.** A researcher's text, integer, date, name or subject Property keeps its two Holds radios and their notice, as today. |
| SO-5 | **The `boolean` pill.** In the table's Value type column, the inspector's Value type row and the create picker's segment, `boolean` is drawn as the other keys are (mono, on the sunken pill), in its own pigment from the palette: not a pigment another type uses (`text` paper, `integer` copper, `date` lapis, `name` iron, `subject` plum, `term` ochre), and with text contrast of at least 4.5:1 in light and dark. |
| SO-6 | **The create picker with `boolean`.** The Value type segmented control offers `text`, `integer`, `date`, `name`, `subject`, `boolean` in that order, each segment expanding to share the row. Draw it so the six fit the dialog's width without truncating. |
| SO-W1 | **Wide.** The Properties page at Wide keeps its arrangement today; the dialog keeps its width (rule 2) and the inspector its width. Every SO requirement looks the same at both widths; draw frames 2 and 4 at Wide. |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Properties page, table and inspector | Type strip, table, inspector at its width | The same; the table takes the extra width | Nothing |
| New property dialog | Fields at the dialog's width | The same dialog, centred | Nothing |

---

## 5. Suggested frames

Draw each as the Properties page pane only, with no app sidebar, at Regular (1102 × 900) unless it says Wide (2296 × 1440). Use the Person type selected, with seeded Properties (Name, Birth, Occupation), a researcher's own **Languages spoken** (text, several values) and a researcher's own **Can read** (boolean).

1. **Create, a single-only type.** New property: Label "Can read", Key `can-read`, Value type `boolean` selected among the six, Holds as SO-1's read-only "One value" with its hint, Person ticked under Bind (SO-1, SO-6).
2. **Create, back to text.** Frame 1 after choosing `text`: Holds shows the two radios, "One value" selected, and the events hint (SO-2).
3. **Inspector, Can read.** The table with Can read selected; the inspector's Value type row shows the `boolean` pill and its lock; Holds is the locked row; a note for SO-3's spoken reason (SO-3, SO-5).
4. **Inspector, Languages spoken.** For comparison: the two radios, "Several values" selected (SO-4).
5. **Table pills.** The table's Value type column with every key's pill, `boolean` among them (SO-5).
6. **Dark mode.** Frames 1 and 3 in dark mode.
7. **Wide.** Frames 2 and 4 at Wide (SO-W1).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| New property dialog | Component | Ship | `.pvFormDialog` (`DesignSystem/Components/FormDialog/PVFormDialog.swift`) in `Features/Properties/PropertiesView.swift` | Unchanged frame. |
| Value type picker | Component | Ship | `PVSegmentedControl` in `PropertiesView` `createForm` | SO-6. One more segment when Go offers it. |
| Holds field (create) | Component | **Extend** | `PVField` + `PVRadio` in `PropertiesView` `createForm` | SO-1, SO-2. For a single-only type: the field's text and lock in place of the radios, the reason as the field's hint. **Read-only Holds → Field with text and lock, not a disabled Radio.** |
| Holds row (inspector) | Snowflake | **Extend** | `PropertiesView` `holdsRow`, its locked branch (`inspectorMetaRow`) | SO-3. The locked branch now also serves a researcher's single-only Property, with its own spoken reason. Not a second row design. |
| Value type pill | Snowflake | **Extend** | `Features/Properties/PropertiesSnapshot.swift` `PropertiesValueTypePill`; colour in `Features/Catalog/SubjectTypeChrome.swift` `PropertiesTypeChrome.valueTypeForeground` | SO-5. One more pigment. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A disabled "Several values" radio | The user chose read-only "One value" (SO-1). |
| A Callout or banner explaining single-only | The hint says it in place (SO-1); the inspector's lock says it (SO-3). |
| Labels in place of keys | Every type keeps its key (user, 2026-11-07). |
| A new kit primitive (a locked field) | The inspector's locked row already draws it. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the frame numbers, the pigment chosen for `boolean`, and every change from this brief with the user's answer.
2. Refresh the [S14-39 plan](../prs/S14-39-properties-single-only.md): copy the pigment and the answers into its Decisions and **Design → macOS**, describe each frame in words, run `review-plan`, and set it to `Planned`.
3. S14-39 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md).
