# S12-D4 — Artifacts section graph links

**Kind:** Claude Design board (**enhancement** of the Source page's Artifacts section)
**Spike:** Provenencia Spike 12 (Source structure)
**View:** Source page — the Add artifact dialog and an Artifact card in edit
**Implements later as:** PR **S12-13**
**Board:** not drawn yet
**Depends on:** S12-12 (a second graph can be made, board S12-D3), S12-11 (the Evidence graphs section, board S12-D2)
**Related:** [`../source-structure.md`](../source-structure.md#rules) (the link rules), [`../deployment-plan.md`](../deployment-plan.md) working rules "Artifact ↔ graph links", "Single-graph Sources hide link checkboxes". [S12-D3](archive/S12-D3-graph-form.md) drew the graph form, its locked rows and their hint; this board reuses the lock and the hint and does not redraw the form. The Add artifact dialog and the Artifact cards are as the Source page draws them today.
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

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

### This brief is an enhancement

The Source page, its Artifacts section, the Add artifact dialog and the Artifact cards stay as they are drawn today. This board adds one **Evidence graphs** group of checkboxes to the Add artifact dialog and to an Artifact card in edit, shown only when the Source has two or more Evidence graphs. On a Source with one graph, nothing on this board changes. The board also rewords two "no artifact" callouts on the canvas and the Citation composer (LK-11, LK-12); it does not redraw them.

---

## 1. Objective

A census has one Evidence graph per household, each linked to the page scans it draws on (S12-D3). When the researcher adds page 7, it is linked to every graph today, and the only way to change an Artifact's graphs is to open each graph's Edit form. The researcher wants to say, from the page itself, "page 7 belongs to the Jones household only".

**What changes.** On a Source with two or more graphs, the Add artifact dialog lists the Source's graphs as checkboxes, all checked; the Artifact is linked to the checked ones. An Artifact card in edit lists the same checkboxes, checked for the graphs the Artifact is linked to; Save writes the label, description and links together, Cancel drops them all. A graph that must keep the Artifact shows the box locked with the reason.

**What stays.** The section header, the card grid, the cards at rest, the Add file dialog, delete, cover and Use as thumbnail are unchanged. A one-graph Source looks and works exactly as now: its Artifacts link to its graph with no choice.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A Source has one or more Evidence graphs | Each has a name ("Jones household") and a `GPH-` ref. The Source page's Evidence graphs section lists them by name; the checkboxes list them in the same order. |
| With one graph, linking is automatic | No checkbox group is drawn at all, in the dialog or on the card. Not a single disabled checkbox. |
| A new Artifact starts linked to every graph | In Add artifact every box starts checked. The researcher unchecks the households the page is not part of. |
| A new Artifact needs a graph | With every box unchecked, **Add** is disabled and a muted line under the group reads "Check at least one evidence graph." |
| A graph whose Citations use the Artifact keeps it | On the card, that graph's box is checked and locked (the kit checkbox's lock mark, as in the graph form) with the hint under its name, the graph form's hint: "1 citation in this graph uses this artifact" / "3 citations in this graph use this artifact". No Citation list on the card; the graph's own Edit form lists them. |
| A graph keeps at least one Artifact | When the Artifact is the only one a graph links, that graph's box is checked and locked with the hint "The only artifact in this evidence graph". |
| The card reads before it can edit links | Opening edit on a card asks which graphs' Citations use the Artifact. Until that read returns, label and description are editable and the group shows a small loading line, "Reading the evidence graphs…"; Save is disabled. If the read fails, a compact danger callout with **Read again**, titled "Couldn’t read the evidence graphs", body "Label and description are kept. Read the evidence graphs again to change the links.", takes the group's place, and Save stays disabled. |
| Save writes everything, Cancel nothing | Save writes the label and description, then each changed link. A refusal leaves the card in edit with the card's danger callout and the app's message for it; every field and box keeps its value. Cancel (or Esc) restores the stored label, description and links. |
| A link change shows elsewhere at once | After Save, each graph's composer and canvas offer the Artifact or stop offering it; the graph form shows the new links. No toast after a card save (as today). |
| After Add | The dialog closes, the card appears and the existing toast shows, as today. |

### 2.1 What this board is not

- Showing an Artifact's graphs on a card at rest (product owner, 2026-10-25: links show only in the dialog and in edit).
- Listing the Citations that keep a link, on the card (the graph form's Kept linked group does that).
- An "All evidence graphs" row; a search or filter over graphs.
- Creating a graph from the dialog or the card (the Evidence graphs section's **Add evidence graph**, S12-D3).
- The graph form, the Copy Source form, the Add file dialog, or the Evidence graphs section rows.
- Per-graph progress (S12-D5).

---

## 3. Implementation gate (S12-13)

| Ships in **S12-13** | Stays out |
| --- | --- |
| The Evidence graphs checkbox group in the Add artifact dialog and in an Artifact card's edit, shown with two or more graphs; locked boxes with their hints; the card's loading and read-error states; Add disabled with no graph checked; the canvas and composer callout copy (LK-11, LK-12). | Links on a card at rest. A Citation list on the card. Graph create or edit. Per-graph progress (S12-14). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| LK-1 | The group is drawn only when the Source has two or more Evidence graphs. With one graph, the dialog and the card are exactly today's frames. |
| LK-2 | The group is one `PVField` labelled **Evidence graphs** with the hint "Which evidence graphs can cite this artifact", holding one checkbox row per graph in name order. Each row: the graph's name as the label, the `GPH-` ref as the row's muted description line. No thumbnail, no All row. Past six rows the list scrolls inside its field. |
| LK-3 | In **Add artifact** the group sits after **Description** and before **File (optional)**. Every box starts checked. |
| LK-4 | In Add artifact, with every box unchecked, **Add** is disabled and the muted line "Check at least one evidence graph." shows under the group. The dialog's discard check counts an unchecked box as a change. |
| LK-5 | In an Artifact card's edit the group sits after **Description** and before **Save** / **Cancel**, at the card's width (264pt and up). Boxes start checked for the graphs the Artifact is linked to. |
| LK-6 | A graph whose Citations use the Artifact: the box is checked, shows the kit's lock mark, can't be clicked, and its description line reads the Citation hint ("{n} citation(s) in this graph use(s) this artifact") in place of the ref. |
| LK-7 | A graph that links only this Artifact: the box is checked and locked the same way, with the hint "The only artifact in this evidence graph". When both apply, the Citation hint shows. |
| LK-8 | Card loading: label and description editable; in the group's place, the forms' loading area at its compact size (one skeleton bar and the caption "Reading the evidence graphs…"); Save disabled. Read error: the forms' compact danger callout with **Read again**, title "Couldn’t read the evidence graphs", body "Label and description are kept. Read the evidence graphs again to change the links." and the error; Save disabled. |
| LK-9 | Save is enabled when the label is not blank and any field or box changed. A refused save keeps the card in edit with the page's danger callout and the app's message; boxes keep their values. Cancel and Esc restore everything. |
| LK-10 | VoiceOver: each row reads the graph name, "checked" or "unchecked", and its ref or hint; a locked row adds "locked". The group reads as "Evidence graphs". |
| LK-11 | The canvas's no-artifact callout names the graph: title "No artifacts in this evidence graph", body "Citing needs an artifact linked to this evidence graph. Place tools, Connect, and Add property stay disabled until one is linked.", action **Go to Source page** (secondary). Layout unchanged. |
| LK-12 | The Citation composer's no-artifact callout body reads "Every citation needs an artifact linked to this evidence graph. Add one on the Source page, or link an existing one from its card, then return here." Layout and button unchanged. |

---

## 5. Suggested frames

Draw over the Source page's current frames.

1. **Add, one graph** — a Source with one graph: the Add artifact dialog exactly as today (LK-1).
2. **Add, census** — a census with graphs "Brown household", "Jones household", "Smith household"; label "Page 7"; all three checked (LK-2, LK-3).
3. **Add, Jones only** — Brown and Smith unchecked (LK-3).
4. **Add, nothing checked** — every box unchecked; Add disabled; "Check at least one evidence graph." (LK-4).
5. **Card edit, one graph** — today's card edit (LK-1).
6. **Card edit, loading** — Page 3 in edit on the census; the group's loading line; Save disabled (LK-8).
7. **Card edit, ready** — Page 3 linked to Jones and Smith; Smith locked with "2 citations in this graph use this artifact"; Brown unchecked (LK-5, LK-6).
8. **Card edit, only artifact** — Page 9 is Brown household's only link: Brown locked with "The only artifact in this evidence graph" (LK-7).
9. **Card edit, read error** and **save refused** (LK-8, LK-9).
10. **Twelve households** — the group scrolling inside the card and inside the dialog (LK-2).
11. **Canvas callout** and **composer callout** with the new copy (LK-11, LK-12).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Add artifact dialog | Component | Ship | `DesignSystem/Components/FormDialog/PVFormDialog.swift`; form in `Features/Sources/SourcePageArtifactsView.swift` `addForm` | Gains the group. No new dialog. |
| Card edit fields | Snowflake | Ship | `SourcePageArtifactsView.swift` `editFields` | Gains the group between Description and the buttons. No new card. |
| Group field | Component | Ship | `PVField` | Label and hint. |
| Graph row | Component | Ship | `DesignSystem/Components/Checkbox/PVCheckbox.swift` `init(verbatim:description:isChecked:isLocked:)` | Name as label, ref or hint as `description`, `isLocked` for LK-6 / LK-7. Not a new row; not `PVArtifactChecklist` (Artifacts with thumbnails). |
| Graph list | Snowflake | **New** | `Features/Sources/SourceArtifactGraphLinks.swift` *(new)* | One view used by the dialog and the card: rows, scroll past six, the "Check at least one" line. First call site: `SourcePageArtifactsView.addForm`; second call site: `SourcePageArtifactsView.editFields`. Both in one feature, so a feature view, not a kit recipe. |
| Loading and read error | Snowflake | **Extend** | `Features/Sources/SourceFormLoadArea.swift` `SourceFormLoadArea` | Gains a compact size (one skeleton bar) for the card. Not a second loading body. |
| Save refused | Component | Ship | The page's danger `PVCallout` (`pageError`) | As card save today. |
| Locked mark and hint | Component | Ship | `PVCheckboxMark` `isLocked`; hint copy `PVDeleteImpactCopy.heading` | The graph form's lock and hint. |
| Canvas callout | Component | Ship | `Features/EvidenceGraph/EvidenceGraphView.swift` `noArtifactCallout` | Copy only. |
| Composer callout | Component | Ship | `Features/CitationComposer/CitationComposerView.swift` `noArtifactGate` | Copy only. |

**Non-goals.** No new kit primitive. No local checkbox. No graph chip or badge on resting cards. No toast after a card save.

---

## 7. Handoff

When the board is agreed: move this brief to the spike's `design/archive` folder, update [`README.md`](README.md), and move the S12-13 plan from `Blocked — on board S12-D4` to `Planned` with the frames' final copy in its Strings section.
