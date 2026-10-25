# S12-D2 — Source page Graphs section

**Kind:** Claude Design board (**enhancement** of the Source page)
**Spike:** Provenencia Spike 12 (Source structure)
**View:** Source page — a new Graphs section, and the header without its Evidence graph jump
**Implements later as:** PR **S12-11**
**Board:** not drawn yet
**Depends on:** S12-05 (graph RPCs and the graph list), S12-08 (a graph is a place)
**Related:** [`../source-structure.md`](../source-structure.md#where-graphs-show) (the design), [`../deployment-plan.md`](../deployment-plan.md) working rules "Source page", "One graph list", "Name and description are editable; graphs list alphabetically by name". S12-D1 drew the Source page header this board edits. S12-D3 (Graph form) adds Add graph and Edit to this section and does not redraw it. S12-D5 (Sources list accordion) draws the same graphs as list rows on another view.
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

The Source page keeps its header, its Description / Credibility / Metadata columns, its Artifacts section and its Notes. This board adds one section, **Evidence graphs**, between the overview columns and Artifacts, and removes one button from the header. Do not redraw the other sections, and do not start a second Source page.

---

## 1. Objective

A census page scan can hold several households. The researcher splits the Source into several Evidence graphs (Jones, Smith, …) so each household is worked and promoted on its own. Today the Source page has a single **Jump to graph** button in the header, which opens the Source's first graph. With several graphs there is no way to see them or choose one.

**What changes.** A **Evidence graphs** section on the Source page with one row per graph: its name, `GPH-` ref and description, an **Open** button and a **Delete** button. The header's **Jump to graph** button is removed; this section replaces it.

**What stays.** The header keeps its cover, title, type, Copy… and Delete. Every other section is unchanged. On an existing project every Source has one graph named Default, so the section shows one row.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| Every Source has at least one graph | The section is never empty once loaded. A new or migrated Source has one graph named **Default**. |
| A graph has a name, a ref and a description | Name: free text, shown as the researcher typed it. **Default** is stored text, not a label: it is shown as "Default", the same as any other name. Ref: `GPH-` plus five characters (`GPH-4K2PQ`), shown in mono. Description: optional free text, often one sentence ("Household 14, Jones family, page 3"). |
| Graphs list alphabetically by name | The row order is by name, A to Z. The order is not draggable. |
| Open opens the graph's canvas | It is the same canvas **Jump to graph** opens today. |
| Open needs a linked Artifact | A graph opens only when at least one of the Source's Artifacts is linked to it, as **Jump to graph** today. A Source with no Artifacts has a Default graph with none linked, so its Open is disabled with the reason "Add an artifact to open this evidence graph". |
| Delete is refused in two cases | A graph that still holds Subjects can't be deleted; the notice lists them ("This graph has 3 subjects"), each linking to its graph. A Source's only graph can't be deleted; the notice reads "This is the source’s only evidence graph" / "A source keeps at least one evidence graph. Add another graph before you delete this one." |
| Delete is always enabled | Like Source and Artifact delete, the button is never greyed out. The DeleteImpact recipe asks first and explains a refusal. |
| An allowed delete asks first | An empty graph that is not the last one: the recipe's confirm, "Delete evidence graph GPH-4K2PQ?", then the row leaves the section and a success toast reads "Evidence graph deleted" / "GPH-4K2PQ was erased from this source." (the same verb as the Artifact toast on this page). |
| The graph list loads separately from the page | The page's own sections can be ready before the graph list is. Until it is loaded, the section shows a short loading state. If the read fails, the section shows a danger callout; the rest of the page still works. |
| No progress on this page | Per-graph progress (Subjects filed, Citations) is on the Sources list (S12-D5), not here. |
| A graph name is the canvas's breadcrumb | Breadcrumbs read **Sources › {Source title} › {graph name}**. They are drawn in the PR plan, not on this board. |

### 2.1 What this board is not

- Creating or editing a graph: **Add graph** in the section header and **Edit** on a row (S12-D3, Graph form). Leave the section header's trailing side clear for it.
- Linking Artifacts to a graph, or graph checkboxes on Artifact cards (S12-D4).
- Per-graph progress, counts of Subjects or Citations, or a Promote button on a row.
- Reordering, pinning or collapsing graphs.
- The Sources list accordion (S12-D5).
- The toolbar breadcrumbs (no new chrome: the existing breadcrumb component with new labels).

---

## 3. Implementation gate (S12-11)

| Ships in **S12-11** | Stays out |
| --- | --- |
| The Evidence graphs section with one row per graph (name, ref, description, Open, Delete), loading and error states, the delete flow through the DeleteImpact recipe, and the header without **Jump to graph**. | Add graph and Edit (S12-12, board S12-D3). Artifact link checkboxes (S12-13, board S12-D4). Per-graph progress (S12-14, board S12-D5). A kit graph-row component. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| GS-1 | The **Evidence graphs** section sits between the Description / Credibility / Metadata columns and the Artifacts section, at the same width and spacing as Artifacts. Its section header reads "Evidence graphs" with the count as meta in words, as Artifacts does ("1 evidence graph", "3 evidence graphs"). The header's trailing side stays empty (S12-D3 puts Add graph there). |
| GS-2 | One row per graph in name order. A row is a card: the name (body, semibold, up to 2 lines then truncated), the ref in mono, muted, on its own line under the name, and the description under the ref (body small, muted, up to 3 lines then truncated). A graph with no description has no description line. |
| GS-3 | Each row ends with **Open** (secondary button, small, graph icon) and **Delete** (icon button, trash, danger tone, small), in that order, matching the header's Copy… and Delete sizes. |
| GS-4 | Open opens the graph's canvas. It is disabled when the graph links no Artifact, with the tooltip and VoiceOver hint "Add an artifact to open this evidence graph". |
| GS-5 | Delete opens the DeleteImpact recipe for that graph: the confirm when allowed; the blocked notice listing the graph's Subjects when it has any; the "only evidence graph" notice for the Source's last graph. Delete is never disabled. |
| GS-6 | After a confirmed delete the row leaves the section and a success toast shows "Evidence graph deleted" / "{ref} was erased from this source." |
| GS-7 | The header loses **Jump to graph**. Copy… and Delete stay where they are, right-aligned. |
| GS-8 | Loading: until the graph list loads, the section header shows with no count and one skeleton row in place of the rows. Error: the section header stays, and a danger callout in place of the rows carries the app's standard message for the failure (the board writes "Couldn’t read the evidence graphs." as a stand-in); the rest of the page is unchanged. |
| GS-9 | VoiceOver: each row reads its name and ref. Open reads "Open {name}"; Delete reads "Delete {name}". |
| GS-10 | Many graphs (12) stack in the section with no inner scroll; the page scrolls as it does today. In a narrow window the name and description wrap within their line limits and Open and Delete keep their size. |

---

## 5. Suggested frames

Draw at the Source page's content width.

1. **Obituary, one graph** — the common case. Header without Jump to graph (Copy… and Delete only); Evidence graphs section with one row, "Default", `GPH-4K2PQ`, no description; Artifacts below (GS-1 – GS-3, GS-7).
2. **Census, three graphs** — "Default", "Jones household", "Smith household", with descriptions "Pages 1–2, every household", "Household 14, page 3", "Household 15, page 3" (GS-2).
3. **No Artifacts** — a new Source with no Artifacts; the Default row's Open is disabled with its tooltip shown (GS-4).
4. **Delete allowed** — the recipe's confirm over the page for "Jones household", titled "Delete evidence graph GPH-JONES?" (GS-5).
5. **Delete blocked by Subjects** — the recipe's notice listing three Subjects of "Smith household" (GS-5).
6. **Last graph** — the recipe's "only evidence graph" notice on frame 1 (GS-5).
7. **After delete** — frame 2 with two rows and the success toast (GS-6).
8. **Loading** and **error** states of the section (GS-8).
9. **Long text** — a 60-character name and a four-sentence description, truncated, at the page's minimum width (GS-2, GS-10).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Section title | Component | Ship | `DesignSystem/Components/SectionHeader/PVSectionHeader.swift` | "Evidence graphs" with the count in words as `meta`, as Artifacts does ("3 evidence graphs"); Notes' bare number is not the model. |
| Graph row | Snowflake | **New** | `Features/Sources/SourcePageGraphsView.swift` (new) | A Card with the name, ref, description, Open and Delete. No second call site: the Sources list accordion (S12-D5) draws list rows with progress, not this card, and S12-D3 adds Edit to this same row. Not a kit primitive. |
| Row surface | Component | Ship | `DesignSystem/Components/Card/PVCard.swift` | One card per graph. Not a local rounded rectangle. |
| Open | Component | Ship | `PVButton` (`.secondary`, `.sm`, `gitBranch` icon) | The same button and icon **Jump to graph** used. |
| Delete | Component | Ship | `PVIconButton` (`trash`, `.danger`, `.sm`) | The same as the header's Source Delete. |
| Delete confirm and refusals | Recipe | Ship | `DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `DeleteImpactFlow.swift` | The page's existing `pvDeleteImpact`. Kind `evidence_graph`. Inbound Subjects and the `last_graph` gate already have copy; the noun "evidence graph" in the confirm title is new copy (today it would read "item"). Not a second Confirm. |
| Loading row | Component | Ship | `DesignSystem/Components/SkeletonBar/PVSkeletonBar.swift` | One skeleton row while the graph list loads. Not a spinner. |
| Load error | Component | Ship | `PVCallout` (danger) | In place of the rows. |
| After-delete notice | Component | Ship | The page's `vocabularyToastOverlay` toast | As Artifact delete. |
| Jump to graph | Snowflake | **Remove** | `Features/Sources/SourcePageIdentityHeader.swift` `evidenceGraphJump` | Replaced by the section's Open. |

**Non-goals.** No new kit primitive: the graph row is a Sources snowflake. No list component with trailing actions for one call site. No local confirm: graph delete is the DeleteImpact recipe. No graph badge, progress meter or chip.

---

## 7. Handoff

When the board is agreed: move this brief to the spike's `design/archive` folder, update [`README.md`](README.md), and move the S12-11 plan from `Blocked — on board S12-D2` to `Planned` with the frames' final copy in its Strings section.
