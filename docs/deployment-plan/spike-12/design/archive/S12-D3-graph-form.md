# S12-D3 — Graph form

**Kind:** Claude Design board (**enhancement** of the Source page's Evidence graphs section)
**Spike:** Provenencia Spike 12 (Source structure)
**View:** Source page — Add evidence graph and Edit on the Evidence graphs section, and the form dialog they open
**Implements later as:** PR **S12-12**
**Board:** agreed 2026-10-25 — [Claude Design](https://claude.ai/design/p/de1e1ccc-aa35-455f-9b9c-3ae269593dc7?via=share). The board changed this brief's copy and layout in a few places; GF-4, GF-5, GF-6, GF-8, GF-10 and GF-11 below read as agreed
**Depends on:** S12-11 (the Evidence graphs section, board S12-D2), S12-02 (the Artifact checklist, board S12-D1)
**Related:** [`../source-structure.md`](../../source-structure.md#rules) (the link rules), [`../deployment-plan.md`](../../deployment-plan.md) working rules "Artifact ↔ graph links", "One title and type form, one Artifact checklist", "Name and description are editable; graphs list alphabetically by name". [S12-D2](S12-D2-source-page-graphs.md) drew the section and its rows; this board adds to them and does not redraw them. [S12-D1](S12-D1-copy-source-form.md) drew the Artifact checklist this form reuses. S12-D4 (Artifacts section graph links) draws graph checkboxes on the Artifacts section; this board does not.
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

The Source page and its Evidence graphs section stay as the S12-D2 frames draw them. This board adds **Add evidence graph** to the section header, **Edit** to each graph row, and the one form dialog both open. The form reuses the Copy Source form's Artifact checklist (S12-D1); do not draw a second checklist. The only change to an existing dialog is the body copy of the "only evidence graph" delete notice (GF-12).

---

## 1. Objective

A census page scan holds several households. The researcher makes one Evidence graph per household ("Jones household", "Smith household"), each linked to the page scans it draws on, so each household is worked and promoted on its own. S12-D2 lists the graphs on the Source page. Today nothing creates a graph or changes one.

**What changes.** **Add evidence graph** in the Evidence graphs section header opens a form: name, description, and a checklist of the Source's Artifacts, all checked. **Edit** on a row opens the same form filled from that graph, with its linked Artifacts checked. An Artifact that the graph's Citations use can't be unchecked; the form says why and lists those Citations.

**What stays.** The section, its rows, Open and Delete are as S12-D2 draws them. The Copy Source form's checklist looks the same. Every other section of the Source page is unchanged.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A graph has a name and a description | Name: required free text, one line ("Jones household"). Two graphs may share a name. Description: optional free text, often one sentence ("Household 14, page 3"). Both are shown as typed. A graph's `GPH-` ref is made when it is saved; the form does not show or edit it. |
| A graph is linked to some of its Source's Artifacts | The links decide which page scans the graph's Citation composer offers. The form's checklist is the link set. |
| A new graph needs an Artifact | **Add evidence graph** is disabled while the Source has no Artifact, with the tooltip and VoiceOver hint "Add an artifact before you add an evidence graph". In the form, Add is disabled while no Artifact is checked. |
| A new graph starts with every Artifact checked | The researcher unchecks the pages the household is not on. |
| An edit needs an Artifact too | When the Source has Artifacts, Save is disabled while none is checked. A Source with no Artifacts can still rename its Default graph: the checklist area shows the line "This Source has no Artifacts." and Save is available. |
| Some links can't be removed | When the graph's Citations use an Artifact, unlinking it is refused. The form finds these Artifacts when it opens and shows each one with the kit checkbox's lock mark (a lock on a paper fill, the look the Properties inspector uses for a binding the researcher cannot clear), with the hint "1 citation in this graph uses this artifact" / "3 citations in this graph use this artifact". Under the checklist, a **Kept linked** group lists those Citations (ref and title), grouped by Artifact, the way the delete notice lists what blocks a delete. |
| A Citation in that list opens it | Clicking a Citation row closes the form without saving and opens that Citation, as the delete notice's rows do. |
| The form reads before it can edit | Edit asks which Artifacts are kept linked. Until that read returns, the name and description are shown and editable; the checklist shows a small loading state and Save is disabled. If the read fails, a danger callout with **Read again** takes the checklist's place. Add needs no read. |
| A new Artifact starts linked | An Artifact added to the Source while the form is open appears checked. |
| Saving is one write | Add creates the graph and its links together; Save writes the name, description and every link together. A refusal saves nothing: the form stays open, keeps every choice and shows a danger callout at the top with the reason. |
| After Add | The dialog closes, the new graph's row appears in the section in name order, and a success toast reads "Evidence graph added" / "GPH-4K2PQ is ready to open." The researcher stays on the Source page. |
| After Save | The dialog closes and the row shows the new name and description. No toast. |
| The Source's only graph can't be deleted | The S12-D2 notice keeps its title, "This is the source’s only evidence graph". Its body names the new control and says a new graph needs an Artifact: "A source keeps at least one evidence graph. Use Add evidence graph to make another, then delete this one. A new evidence graph needs an artifact." |
| Graphs list by name | A renamed graph moves to its new place in the section. |

### 2.1 What this board is not

- Graph checkboxes in the Add Artifact dialog or on Artifact cards (S12-D4).
- Showing the graph's ref, Subjects, Citations or progress inside the form.
- Moving Subjects between graphs, splitting a graph or merging two.
- Copying a graph, or choosing graphs in the Copy Source form.
- A new Citation list design: the Kept linked group is the delete notice's group of rows.
- A redraw of the section, its rows, or the Copy Source form.

---

## 3. Implementation gate (S12-12)

| Ships in **S12-12** | Stays out |
| --- | --- |
| Add evidence graph in the section header; Edit on each row; the graph form dialog for both, with name, description and the Artifact checklist; locked rows and the Kept linked group for Artifacts in use; loading, read-error and save-error states; the after-add toast; the "only evidence graph" notice body. The Artifact checklist becomes a shared recipe for its two forms. | Graph checkboxes on the Artifacts section (S12-13, board S12-D4). Per-graph progress (S12-14). Any change to the Copy Source form's look. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| GF-1 | The Evidence graphs section header's trailing side holds **Add evidence graph**: a primary button, small, plus icon, the same as **Add artifact** in the Artifacts section header. |
| GF-2 | Add evidence graph is disabled while the Source has no Artifact, with the tooltip and VoiceOver hint "Add an artifact before you add an evidence graph". |
| GF-3 | Each graph row gains **Edit**: an icon button, pencil, neutral tone, small, between Open and Delete. Its tooltip and VoiceOver label are "Edit {name}". It is the same button the Artifact cards use for edit. |
| GF-4 | Add and Edit open one form dialog, 520pt wide. Add: title **Add evidence graph**, subtitle "An evidence graph is a discrete group of subjects and their relationships. Use different graphs to isolate groups of data from a large source.", confirm **Add**. Edit: title **Edit evidence graph**, subtitle the graph's ref, confirm **Save**. Cancel **Cancel** in both. |
| GF-5 | The form body, top to bottom: **Name** (required, single-line input, empty on Add, focused when the form opens; placeholder "Jones household"); **Description** (optional, with the field's Optional marker; a three-line text area; hint "Which part of the source this graph covers"); a divider; the **Artifacts** checklist. |
| GF-6 | The checklist is the Copy Source form's: a heading with "{checked} of {total}", the caption "Select all the artifacts that are relevant to this evidence graph." under it (this form only), an All Artifacts checkbox (checked, unchecked or mixed), then one row per Artifact in the Source page's order with its thumbnail, label (or `ART-` ref) and checkbox; the rows scroll inside the dialog past about eight. On Add every row is checked. On Edit the graph's linked Artifacts are checked. With no Artifacts, the line "This Source has no Artifacts." replaces it. |
| GF-7 | A row whose Artifact the graph's Citations use shows the checkbox's lock mark and can't be clicked, with the hint under its label "{n} citation(s) in this graph use(s) this artifact". All Artifacts never unchecks it: with every other row off, All Artifacts reads mixed. |
| GF-8 | When any row is locked, a **Kept linked** group sits under the checklist: a neutral compact callout with a lock icon and no title, "These artifacts stay linked while citations in this graph use them.", then for each locked Artifact its label and, under it, the delete notice's list of Citation rows (ref, title, arrow) with its "and N more" line past twenty. Clicking a row closes the form without saving and opens the Citation. |
| GF-9 | Name validation follows Add artifact: Add or Save with a blank name shows "Give the evidence graph a name." under the field. With Artifacts on the Source and none checked, Add and Save are disabled and a muted line under the checklist reads "Check at least one artifact." |
| GF-10 | Edit's loading state: name and description editable; the checklist area shows the Copy form's loading area (three skeleton bars and a caption, here "Reading the artifacts…"); Save disabled. Read error: the Copy form's compact danger callout with **Read again**, titled "Couldn’t read the artifacts", body "Name and description are kept. Read the artifacts again to change the links." and the error, in the checklist's place; Save disabled. One loading and read-error area for both forms. |
| GF-11 | A save failure shows a danger callout at the top of the form: title "Couldn’t save the evidence graph", body "Nothing was saved. Your changes are kept." and the app's message for the refusal. Every field and checkbox keeps its value. |
| GF-12 | The "only evidence graph" delete notice keeps its title and reads "A source keeps at least one evidence graph. Use Add evidence graph to make another, then delete this one. A new evidence graph needs an artifact." |
| GF-13 | After Add, the dialog closes, the new row shows in name order, and the toast reads "Evidence graph added" / "{ref} is ready to open.". After Save, the dialog closes with no toast. |
| GF-14 | Discard follows the kit's form dialog: Cancel closes at once; a click outside a changed form asks "Discard changes?". |
| GF-15 | VoiceOver: a locked row reads its label, "checked", "locked" and the hint. The Kept linked rows read "{ref}, {title}". |

---

## 5. Suggested frames

Draw the dialog at 520pt over the Source page from the S12-D2 frames.

1. **Section header** — frame 2 of S12-D2 (census, three graphs) with Add evidence graph in the header and Edit on each row (GF-1, GF-3).
2. **No Artifacts** — frame 3 of S12-D2 with Add evidence graph disabled and its tooltip shown (GF-2).
3. **Add, census** — name "Jones household", description "Household 14, page 3", six page scans with pages 1, 2, 4, 5 and 6 unchecked; All Artifacts mixed; heading "1 of 6" (GF-4 – GF-6).
4. **Add, nothing checked** — every row unchecked; Add disabled; "Check at least one artifact." (GF-9).
5. **Add, blank name** — the name error under the field (GF-9).
6. **Edit, loading** — Edit on "Smith household": name and description filled, checklist skeleton, Save disabled (GF-10).
7. **Edit, kept links** — "Smith household" linked to pages 3 and 4; page 3 with the lock mark and "2 citations in this graph use this artifact"; Kept linked group listing `CIT-7RQ2M` "Smith, John — 1851 census" and `CIT-8TW4N` "Smith, Mary — 1851 census" under "Page 3" (GF-7, GF-8).
8. **Edit, Default with no Artifacts** — renaming Default on a Source with none; "This Source has no Artifacts."; Save available (GF-6, GF-9).
9. **Read error** and **save failed** (GF-10, GF-11).
10. **After add** — the section with four rows and the success toast (GF-13).
11. **Only evidence graph** — the S12-D2 notice with the new body (GF-12).
12. **Forty pages** — a census with 40 page scans; the rows scroll and the buttons stay visible (GF-6).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Form dialog shell | Component | Ship | `DesignSystem/Components/FormDialog/PVFormDialog.swift` | Add and Edit, 520pt, as Copy Source. One dialog, two modes. Not two dialogs. |
| Name | Component | Ship | `PVField` + `PVInput` | Required mark and inline error, as Add artifact's Label. |
| Description | Component | Ship | `PVField` + `PVTextArea` | Optional, three lines. |
| Artifact checklist | Recipe | **Extend** (promote) | `Features/Sources/SourceArtifactChecklist.swift` → `DesignSystem/Recipes/ArtifactChecklist/PVArtifactChecklist.swift` *(new)* | Its second caller is this form, so it is promoted per `add-ui-component`. Gains locked rows (checked, disabled, with a hint). The Copy Source form shows it unchanged. Not a second checklist. |
| Row checkbox | Component | **Extend** | `DesignSystem/Components/Checkbox/PVCheckbox.swift` | Label, hint as `description`. A locked row uses the kit's lock mark (`PVCheckboxMark` `isLocked`, already drawn by the Properties inspector), which the labelled checkbox gains as a parameter. Not a dimmed check, not a local lock glyph. |
| Artifact thumbnail | Component | Ship | `Platform/CachedThumbnail.swift` `CachedThumbnail` (32pt) | As in the Copy form. |
| Kept linked Citation list | Recipe | **Extend** | `DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` (the notice's private group view, made a recipe view) | The delete notice's rows of ref, title and arrow, with its "and N more" line. One list for the notice and this form. Not a new Citation row. |
| Kept linked intro line | Component | Ship | `PVCallout` (neutral, compact) | Above the per-Artifact lists. |
| Loading and read error | Snowflake | **Extend** | `Features/Sources/SourceCopyDialog.swift` `SourceCopyForm.loading` and its failed callout → `Features/Sources/SourceFormLoadArea.swift` *(new)* | The Copy form's area (`PVSkeletonBar` ×3 + caption; compact danger `PVCallout` + **Read again**), shared by both forms. Not a second loading body. |
| Save error | Component | Ship | `PVCallout` (danger) | At the top of the form. |
| Add evidence graph | Component | Ship | `PVButton` (`.primary`, `.sm`, `plus`) in `PVSectionHeader` `actions` | The same as Add artifact. |
| Edit | Component | Ship | `PVIconButton` (`penLine`, `.sm`) | The same as the Artifact card's edit. |
| Graph row | Snowflake | Ship | `Features/Sources/SourcePageGraphsView.swift` `SourcePageGraphRow` | Gains Edit between Open and Delete. Nothing else changes. |
| After-add notice | Component | Ship | The page's `vocabularyToastOverlay` toast | As Add artifact. |
| Only-graph notice | Recipe | Ship | `PVDeleteImpact` `last_graph` gate | Body copy only (GF-12). |
| Discard-changes confirm | Component | Ship | `PVFormDialog` `isDirty` | As Copy Source. |

**Non-goals.** No new kit primitive: the checklist and the Citation list are recipes because each now has two callers. No local checkbox, list or dialog. No graph ref field. No toast after an edit.

---

## 7. Handoff

When the board is agreed: move this brief to the spike's `design/archive` folder, update [`README.md`](../README.md), and move the S12-12 plan from `Blocked — on board S12-D3` to `Planned` with the frames' final copy in its Strings section.
