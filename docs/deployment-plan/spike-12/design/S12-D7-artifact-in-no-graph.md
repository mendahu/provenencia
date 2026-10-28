# S12-D7 — Source page Artifacts section: an Artifact in no graph

**Kind:** Claude Design board (**enhancement** of the Source page's Artifacts section, over the S12-D4 frames)
**Spike:** Provenencia Spike 12 (Source structure)
**View:** Source page — the Add artifact dialog, an Artifact card at rest and in edit; in the graph form, its locks and one line of its Kept linked group
**Implements later as:** PR **S12-38**
**Board:** not drawn yet. Draw it on the Source page board (the [Claude Design project](https://claude.ai/design/p/de1e1ccc-aa35-455f-9b9c-3ae269593dc7?via=share) that holds the S12-D3 and S12-D4 frame rows), as a new S12-D7 frame row
**Depends on:** S12-22 (the engine lets an Artifact stand in no graph, landed [#600](https://github.com/mendahu/provenencia/pull/600)); [S12-D4](archive/S12-D4-artifact-graph-links.md) (the graph checkboxes this board extends)
**Related:** [`../source-structure.md`](../source-structure.md#rules) (the link rules), [`../deployment-plan.md`](../deployment-plan.md) working rules "An Artifact is the parent; it may stand in no graph", "An Artifact in no graph says so on its card", "An Artifact with Citations keeps at least one graph", "An Artifact card's Save is one write", "A new graph starts with the Artifacts in no graph checked". [S12-D4](archive/S12-D4-artifact-graph-links.md) drew the checkbox group, its locks, the card's loading and read-error states and the refused-save callout; this board changes when they show and what they say. [S12-D3](archive/S12-D3-graph-form.md) drew the graph form; this board changes which of its rows lock and one line of its Kept linked group, and draws nothing else there. [`docs/ideas/artifact-outside-graph-marker.md`](../../../ideas/artifact-outside-graph-marker.md) holds the part left out (the graph delete confirm).
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

The Source page, its Artifacts section, the Add artifact dialog and the Artifact cards stay as the S12-D4 frames draw them. This board changes three things on them: the **Evidence graphs** group now shows on a Source with one graph too, and may be left with no box checked; a card at rest whose Artifact is in no graph says so and offers linking it; and the card's locks and refused save read differently because a card Save is now one write. In the graph form it locks every Artifact the engine says the graph must keep and rewords one line (ND-12); nothing else there changes.

---

## 1. Objective

A researcher adds a new page of a census and does not yet know which household it belongs to. Today the page must join a graph to exist, and a new graph must start from a page, so a new page and a new graph cannot be made together. From S12-22 the engine lets a page stand in no Evidence graph. This board lets the researcher do that, and makes such a page visible on its card so it is not forgotten.

**What changes.** The Add artifact dialog and the card in edit list the Source's graphs on every Source, one graph included. Every box starts checked in the dialog; unchecking all of them is allowed and the dialog says what that means. A card at rest whose Artifact is in no graph shows a "Not in an evidence graph" mark and a **Link to a graph** action that opens the card in edit with the boxes. Every lock on a box is the engine's: the card and the graph form lock each graph the engine says must keep the Artifact, worded by why, and the lock does not lift while the card is open. A cited page in one graph moves to another in two Saves: link the new graph, then drop the old. A card Save writes everything or nothing, so its refusal no longer talks about a partial save.

**What stays.** The section header, the card grid, the card's preview well, ref, cover badge, Use as thumbnail, edit and delete; the group's field, rows, scroll and order (S12-D4 LK-2); the card's loading and read-error states (LK-8); Cancel and Esc; the toast after Add; no toast after a card Save.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| Every Source has at least one Evidence graph | The engine refuses to delete a Source's last graph. The group shows on every Source; the only time it does not is while the page's graph list is still loading. |
| An Artifact may be in no graph | It stays on the Source page with its file and its cover role. No composer or canvas offers it until a graph links it. |
| A new Artifact starts in every graph | In Add artifact every box starts checked. The researcher unchecks the graphs the page is not part of, or all of them. |
| No box checked is allowed | **Add** stays enabled. A muted line under the group says the page will be in no graph (ND-3). |
| The card at rest shows links only when there are none | A card whose Artifact is in at least one graph looks exactly as today: no graph names, chips or counts on resting cards (S12-D4 §2.1). |
| A graph keeps at least one Artifact | When the Artifact is the only one a graph links, that graph's box is checked and locked. |
| A graph whose Citations use the Artifact keeps it | That graph's box is checked and locked, with the count of those Citations, as in S12-D4 LK-6. |
| An Artifact with Citations keeps at least one graph | When the Artifact is in exactly one graph and has Citations, that graph's box is locked. Checking another graph does not unlock it. To move the page, the researcher checks the new graph and saves; the next time the card opens the old graph is no longer locked, and the researcher unchecks it and saves (user, 2026-10-28: the app holds no rule of its own). |
| A card Save is one write | The label, the description and every link save together, or nothing saves. A refusal leaves every field and box as the researcher left them and reads the boxes' locks again. |
| A new graph starts with the pages in no graph | The graph form's Add evidence graph starts with the Artifacts in no graph checked when there are any, otherwise with every Artifact (as today). No new chrome: S12-D3's frames, different boxes checked. |

### 2.1 What this board is not

- Naming, in a graph delete's confirm, the Artifacts the delete leaves in no graph ([`docs/ideas/artifact-outside-graph-marker.md`](../../../ideas/artifact-outside-graph-marker.md)).
- Showing an Artifact's graphs on a card at rest when it has any.
- A filter or count of the Artifacts in no graph on the section header or the Sources list.
- Creating a graph from the dialog or the card (the Evidence graphs section's **Add evidence graph**, S12-D3).
- Redrawing the graph form or its Kept linked group; only ND-12 changes there.
- A toast after a card Save.

---

## 3. Implementation gate (S12-38)

| Ships in **S12-38** | Stays out |
| --- | --- |
| The group on a one-graph Source in the dialog and the card; no box checked allowed with its muted line; the resting card's mark and **Link to a graph**; every engine lock with its hint by reason; the one-write refused-save callout; the Kept linked line in the graph form (ND-12). | The graph delete confirm naming Artifacts. Any resting-card mark for an Artifact in some graph. A header count. Graph create from the card. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| ND-1 | The **Evidence graphs** group (S12-D4 LK-2) shows in the Add artifact dialog and in a card's edit on every Source whose graph list has loaded, one graph included. While the list loads, the dialog shows no group and the card shows the group's loading state (LK-8). |
| ND-2 | In Add artifact every box starts checked, as before. |
| ND-3 | With every box unchecked, **Add** stays enabled and a muted italic line under the group reads "This artifact won’t be in an evidence graph. Citations can’t use it until it is linked." It replaces "Check at least one evidence graph." An unchecked box still counts as a change for the dialog's discard check. |
| ND-4 | A card at rest whose Artifact no graph links shows, below its description, one row: a neutral Badge with the kit's `link` icon reading "Not in an evidence graph", and a ghost small Button **Link to a graph** at the row's end. The row wraps at the card's narrowest width (264pt) with the Button under the Badge. Cards whose Artifact is in a graph show no row. A card with the row is taller than its neighbours; the grid aligns card tops and does not level heights, as with a card in edit. The row sits under the description, not in the card's top row, which already holds the ref, the cover badge, Use as thumbnail and the edit and delete buttons at 264pt. |
| ND-5 | **Link to a graph** opens that card in edit exactly as the pencil does (focus in Label, as today), with every box unchecked. |
| ND-6 | The card in edit, on any Source: boxes start checked for the graphs that link the Artifact. With every box unchecked the same muted line as ND-3 shows under the group (below it, never over the loading or read-error state, which replace the group), and Save stays enabled when anything changed. |
| ND-7 | Lock hints by reason, in the row's description line in place of the ref: the graph's Citations use the Artifact → "1 citation in this graph uses this artifact" / "3 citations in this graph use this artifact" (as LK-6); the Artifact is the graph's only one → "An evidence graph needs at least one artifact."; the Artifact's Citations need a graph → "1 citation would be left in no evidence graph" / "3 citations would be left in no evidence graph". When more than one reason applies, the first in this order shows. |
| ND-8 | No lock lifts while the card is open: checking or unchecking other boxes never changes a locked row. A lock changes only when the card reads the engine again (edit opens, Read again, or a refused Save). A reason the app has no words for still locks, with the app's message for its code as the hint. |
| ND-9 | A refused Save keeps the card in edit with the compact danger callout as the first item of its edit fields (LK-9 position), title "Couldn’t save the artifact". Body when the locks read again name a graph the researcher unchecked: "{graph name} keeps this artifact. Your changes are kept. {error}"; otherwise: "Your changes are kept. {error}". No body says "save again to finish". Every field and box keeps its value; a row the re-read locks shows checked and locked. |
| ND-10 | VoiceOver: the resting mark reads "Not in an evidence graph"; **Link to a graph** reads "Link {artifact ref} to an evidence graph". The muted line of ND-3 / ND-6 is read after the group. A locked row reads "locked" and its hint. |
| ND-11 | Keyboard: **Link to a graph** is in the card's tab order after the description and before the next card. Return on it opens edit as a click does. |
| ND-12 | Graph form: every Artifact the engine says the graph must keep is locked, with the ND-7 hint, including the graph's only Artifact ("An evidence graph needs at least one artifact."). The Kept linked group lists the locked Artifacts that have Citations, under the ND-7 heading; its callout line reads "These artifacts stay linked: citations need them in this evidence graph." (was "These artifacts stay linked while citations in this graph use them."). No other change to the form. |

---

## 5. Suggested frames

Draw over the S12-D4 frame row. A census with graphs "Brown household", "Jones household", "Smith household" unless a frame says otherwise.

1. **Add, one graph** — a Source with one graph "Default": the dialog now shows the group with one checked box (ND-1, ND-2).
2. **Add, nothing checked** — the census, every box unchecked: Add enabled, the ND-3 line under the group.
3. **Resting cards** — four cards: Page 3 (in Jones and Smith), Page 8 in no graph with the ND-4 row, Page 9 cover with a long label in no graph (the row wraps), Page 4 (in Jones). Only Pages 8 and 9 show the row (ND-4).
4. **Link to a graph** — Page 8 in edit after **Link to a graph**: every box unchecked, the ND-6 line (ND-5, ND-6).
5. **Card edit, one graph** — Page 3 on a one-graph Source "Default": one box, checked, unlocked (ND-1, ND-6).
6. **Cited page, locked** — Page 5 has two Citations and is only in Jones: Jones checked and locked with "2 citations would be left in no evidence graph"; Brown and Smith unchecked (ND-7).
7. **Cited page, first Save** — the same card after checking Smith: Jones still locked, Smith checked, Save enabled (ND-8).
8. **Only artifact** — Page 9 is Brown household's only link: Brown locked with "An evidence graph needs at least one artifact." (ND-7).
9. **Save refused** — Page 3: Jones unchecked; a Citation took it since edit opened; the ND-9 callout naming Jones household; Jones now checked and locked with "1 citation in this graph uses this artifact".
10. **Graph form, Kept linked** — S12-D3's edit frame with the ND-12 line, one Artifact listed under the ND-7 Citations heading, and the graph's only other Artifact locked with "An evidence graph needs at least one artifact." but not listed.
11. **Loading** — the resting cards and the Add dialog while the page's graph list is loading: no mark row, no group (ND-1).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Add artifact dialog | Component | Ship | `DesignSystem/Components/FormDialog/PVFormDialog.swift`; form in `Features/Sources/SourcePageArtifactsView.swift` `addForm` | Group shows on one graph; the ND-3 line replaces the needs-one line. No new dialog. |
| Card edit fields | Snowflake | Ship | `Features/Sources/SourcePageArtifactsView.swift` `editFields` | Group shows on one graph; ND-6 line. No new card. |
| Graph list | Snowflake | Ship | `Features/Sources/SourceArtifactGraphLinks.swift` `SourceArtifactGraphLinks` | Same rows, scroll and lock. No new row state. |
| Graph row | Recipe | Ship | `DesignSystem/Recipes/ArtifactChecklist/PVChecklistRows.swift` `PVChecklistCheckbox` | Lock mark, "Locked" value and hint as in S12-D4. Not `PVCheckbox` directly. |
| No-graph line (ND-3, ND-6) | Snowflake | Ship | `Features/Sources/SourcePageArtifactsView.swift` | The muted italic caption S12-D4 drew for the needs-one line, new copy. Not a Callout. |
| Resting mark | Component | Ship | `DesignSystem/Components/Badge/PVBadge.swift` `PVBadge` | Neutral tone, `link` icon (the kit has no link-off icon; do not draw one). Not a Chip (not a token), not a Callout (not an error). |
| Link to a graph | Component | Ship | `DesignSystem/Components/Button/PVButton.swift` `PVButton` | Ghost, small. Not an IconButton: the words are the action. |
| Resting card text | Snowflake | Extend | `Features/Sources/SourcePageArtifactsView.swift` `restingText` | Gains the ND-4 row under the description. One call site, so not a kit recipe and not a new view type. |
| Mark row layout | Component | Ship | `DesignSystem/Components/FlowLayout/PVFlowLayout.swift` `PVFlowLayout` | Wraps the Badge and the Button. Not a local wrap. |
| Loading and read error | Snowflake | Ship | `Features/Sources/SourceFormLoadArea.swift` `SourceFormLoadArea` | Compact size, as S12-D4. |
| Save refused | Component | Ship | `DesignSystem/Components/Callout/PVCallout.swift` `PVCallout` | Danger, compact, ND-9 copy. |
| Locked hint (Citations) | Recipe | Ship | `DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` `PVDeleteImpactCopy.heading` | Both Citation hints of ND-7 are its headings. |
| Kept linked group | Snowflake | Ship | `Features/Sources/SourceGraphFormDialog.swift` `keptLinked` | ND-12 copy; lists only locked Artifacts with Citations. |

**Non-goals.** No new kit primitive. No local checkbox or badge. No graph chip or badge on a card whose Artifact is in a graph. No toast after a card Save.

---

## 7. Handoff

When the board is agreed: move this brief to the spike's `design/archive` folder, update [`README.md`](README.md), write the agreed copy into the S12-38 plan's Strings, and move the plan from `Blocked — on board S12-D7` to `Planned`.
