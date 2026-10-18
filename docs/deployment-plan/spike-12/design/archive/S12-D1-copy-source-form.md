# S12-D1 — Copy Source form (Source page)

**Kind:** Claude Design board (**enhancement** of the Add Source form and the Source page header)
**Spike:** Provenencia Spike 12 (Source structure)
**View:** Source page — the Copy action and the Copy Source form it opens
**Implements later as:** PR **S12-02**
**Board:** agreed 2026-10-18 — [Claude Design](https://claude.ai/design/p/de1e1ccc-aa35-455f-9b9c-3ae269593dc7?via=share)
**Depends on:** S12-01 (Copy Source engine)
**Related:** [`../source-structure.md`](../../source-structure.md#copy-source) (the design), [`../deployment-plan.md`](../../deployment-plan.md) working rules "Copy Source". S12-D6 (Sources list) opens this same form from a row's right-click menu and does not redraw it. S12-D3 (Graph form) reuses the Artifact checklist drawn here.
**Design system layers:** [`docs/design-system-layers.md`](../../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../../.agents/skills/add-ui-component/SKILL.md)

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

The Copy form is the Add Source form's title and type fields with a "what to copy" section under them. Do not draw a second title field, a second type picker or a new dialog shell. The Source page header gains one icon button beside Delete. Nothing else on the page moves. The Sources list entry point is its own board, S12-D6.

---

## 1. Objective

A researcher with twelve obituaries from one newspaper retypes the publication, issue, page and URL each time. Copy Source starts the next obituary from the last one in one step: same type, same metadata and layout, same credibility, a chosen subset of the Artifacts, and a new title.

**What changes.** A **Copy…** action on the Source page. A Copy Source form dialog, which S12-D6 also opens from the Sources list. Saving opens the new Source's page.

**What stays.** The Add Source form looks and behaves as it does today. The Source page header keeps its cover, title, type, Delete and Evidence graph controls.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A copy is a new, ordinary Source | It gets its own `SRC-` ref and its own Evidence graph. It is not linked to the original afterwards. Editing one never changes the other. |
| Title and type come from Add Source | Add Source is a form dialog at 480pt: **Title** (required, a single-line input) and **Source type** (required, a searchable combo box with each type's icon), then an optional description. The Copy form reuses Title and Source type exactly, with the same required marks and the same inline errors under each field. |
| The title is prefilled | `Copy of <original title>`. The field is focused when the form opens. |
| The type is preselected | The original's type. The researcher can pick another. Metadata values are still copied as they are when the type changes. |
| Description is a checkbox, not a field | The Copy form copies the original's description or leaves it empty. It does not show an editable description. |
| Parts are checkboxes, all on | Description; Metadata (values and field layout); Credibility (grade and argument); one checkbox per Artifact. |
| A part can be absent | An original with no description, no metadata, or no credibility shows that checkbox disabled and unchecked, with the hint "None on this Source". |
| Artifacts can be many | A census Source can hold 40 page scans. Each row is a small thumbnail, the Artifact's label (or its `ART-` ref when it has no label) and a checkbox. The list scrolls inside the dialog. An **All Artifacts** checkbox above the rows is checked, unchecked or mixed. |
| An Artifact copy shares the file | The copy gets a new Artifact row pointing at the same file. Nothing is stored twice. The board does not need to say so. |
| Never copied | Notes, Subjects, Citations, Observations, the Evidence graph's content and the cover. The copy shows its type icon until a cover is chosen. The dialog subtitle says notes and the Evidence graph are not copied; it does not list every table. |
| The form can open before the parts are loaded | On the Source page they are already loaded. From the Sources list (S12-D6) the form loads the original's Artifacts and parts when it opens. Title and type show at once; the parts section shows a small loading state; Copy is disabled until it loads. |
| Saving is one write | Copy creates the Source and its parts together, then opens the new Source's page. There is no toast. A failure keeps the dialog open with a danger callout at the top. |

### 2.1 What this board is not

- Choosing graphs to copy, or copying Subjects, Citations or notes.
- Editing the description, metadata, credibility or an Artifact inside the Copy form.
- A bulk copy of several Sources, or "copy as template".
- A new toolbar or overflow menu on the Source page.
- The Sources list row and its right-click menu (S12-D6).
- A redraw of Add Source.

---

## 3. Implementation gate (S12-02)

| Ships in **S12-02** | Stays out |
| --- | --- |
| The Copy Source form dialog with the Add Source title and type fields, the part checkboxes and the Artifact checklist. Copy… on the Source page header. Saving opens the copy. | Graph choices (a copy always starts with one Evidence graph). The Sources list entry point (S12-17, board S12-D6). A toast after copy. A kit Checklist primitive (it is promoted when the Graph form, S12-12, becomes its second caller). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CS-1 | **Copy…** on the Source page is an icon button (duplicate icon) in the header's action row, before Delete, same size as Delete, neutral tone. Its tooltip and accessibility label are "Copy…". |
| CS-2 | The same form opens from a Sources list row (board S12-D6). This board draws it once, opened from the Source page; it does not depend on where it was opened, except for CS-8. |
| CS-3 | The form is a form dialog titled **Copy Source**, subtitle "Start a new Source from <original title>. Notes and the Evidence graph are not copied." Confirm **Copy**, Cancel **Cancel**. |
| CS-4 | Title and Source type are the Add Source fields, same order, same required marks, same errors, same combo rows with type icons. Title is prefilled `Copy of <title>`; type is preselected to the original's. |
| CS-5 | Under them, a **Copy from the original** section: Description, Metadata (hint "Values and field layout"), Credibility (hint "Grade and argument"), each a labelled checkbox, all on. A part the original lacks is disabled and unchecked with the hint "None on this Source". |
| CS-6 | An **Artifacts** group under the parts: an All Artifacts checkbox (checked / unchecked / mixed), then one row per Artifact in the Source page's order: thumbnail, label (or ref), checkbox; all on. With no Artifacts, one muted line, "This Source has no Artifacts.", in place of the group. |
| CS-7 | With many Artifacts, the Artifact rows scroll inside the dialog; the title, type, parts and the dialog buttons stay visible. |
| CS-8 | Loading (the form opened before the parts are loaded, as from the Sources list): title and type are shown and editable; the parts and Artifacts area shows a small loading state; Copy is disabled. A load failure shows a danger callout in that area. |
| CS-9 | Copy writes and opens the new Source's page. A failure shows a danger callout at the top of the form and keeps every choice. |
| CS-10 | Discard follows the kit's form dialog: Cancel closes at once; a click outside a changed form asks "Discard changes?". (Corrected 2026-10-18: this line first said Cancel asks, as Add Source does; Add Source does not. The board's "Discard this copy?" on Cancel is not built.) |

---

## 5. Suggested frames

Draw the dialog at 520pt.

1. **Source page header** with Copy… beside Delete (CS-1). Draw only the header.
2. **Copy form, an obituary**: title "Copy of Obituary — Mary Alderwick", type Obituary, all three parts on, two Artifacts on (CS-3 – CS-6).
3. **Some unchecked**: Credibility off, one Artifact off, All Artifacts mixed.
4. **Parts absent**: an original with no description and no credibility, both disabled with "None on this Source"; no Artifacts line (CS-5, CS-6).
5. **Census, 40 Artifacts**: the Artifact rows scroll; the buttons stay visible (CS-7).
6. **Loading**, then **load error** (CS-8).
7. **Validation**: title cleared, Add Source's title error under it; Copy disabled or showing the error as Add Source does (CS-4).
8. **Save failed**: danger callout at the top, choices kept (CS-9).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Form dialog shell | Component | Ship | `DesignSystem/Components/FormDialog/PVFormDialog.swift` | The Copy Source dialog, 520pt. Same shell as Add Source. |
| Title and type fields | Snowflake | **Extend** | `Features/Sources/SourcesListView.swift` `addForm` (to become a shared Sources view) | One title-and-type block for Add Source and Copy Source. Not a second form. |
| Source type combo row | Snowflake | **Extend** | `SourcesListView.swift` `sourceTypeComboRow` and the inline row in `Features/Sources/SourcePageIdentityHeader.swift` | One row (type icon + label) for every Source type picker, including the Source page's type editor. |
| Title input | Component | Ship | `PVField` + `PVInput` | Prefilled, required. |
| Source type pick | Component | Ship | `PVComboBox` with the type-icon row | Same rows as Add Source. Not a Select. |
| Part checkbox | Component | Ship | `DesignSystem/Components/Checkbox/PVCheckbox.swift` | Label + description (the hint). Disabled when the part is absent. |
| Artifact checklist | Snowflake | **New** | `Features/Sources/` | All Artifacts (PVCheckbox, mixed state) + one row per Artifact: Thumbnail + PVCheckbox. Built once here; its second call site is the Graph form (S12-D3 / S12-12), which promotes it per `add-ui-component`. Not a kit primitive yet. |
| Artifact thumbnail | Component | Ship | `Platform/CachedThumbnail.swift` `CachedThumbnail` (32pt) | Its `typeIconKey` fallback already shows the Source type icon for a fileless Artifact, as on the Source page. Not a raw `PVThumbnail`. |
| Error and load-error notice | Component | Ship | `PVCallout` (danger) | Top of the form for a save failure; in the parts area for a load failure. |
| Page Copy… button | Component | Ship | `PVIconButton` (`duplicate`) in `Features/Sources/SourcePageIdentityHeader.swift` | Beside Delete. |
| Discard-changes confirm | Component | Ship | `PVFormDialog` `isDirty` | As Add Source. |

**Non-goals.** No new kit primitive: the Artifact checklist stays a Sources snowflake until its second caller. No local checkbox or dialog. No "copied" badge or link between the copy and the original.

---

## 7. Handoff

When the board is agreed: move this brief to the spike's `design/archive` folder, update [`README.md`](README.md), and move the S12-02 plan from `Blocked — on board S12-D1` to `Planned` with the frames' final copy in its Strings section.
