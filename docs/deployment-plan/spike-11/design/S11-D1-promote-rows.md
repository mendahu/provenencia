# S11-D1 — Promote page rows: contested marker, move reasons, connection subrows

**Kind:** Claude Design board (**enhancement** of the Promote page board, S9-D16)
**Spike:** Provenencia Spike 11 (Promote matching layers)
**View:** Promote page — the proposal rows
**Implements later as:** PR **S11-18** (contested marker, move reasons) and PR **S11-27** (connection subrows)
**Depends on:** S11-17 (landed [#504](https://github.com/mendahu/provenencia/pull/504): rows carry `contested` and `move`); S11-26 (connections file by their ends) before S11-27 ships
**Related:** [`../promote-connection-subrows.md`](../promote-connection-subrows.md) (the subrow idea and its design questions), [`../comparison-layers.md`](../comparison-layers.md#bands) (bands and contested), [`../../../promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3. S11-D2 later redraws the evidence sheet the row opens.
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

Extend the Promote page frames that exist (S9-D16): the header, the sections by kind or by assessment, the rows with their kind tile, name and ref, target menu, assessment column and Updated / Decided badges, the anchors block, the footer. Do not redraw the target menu, the header, the section headers or the evidence sheet. Two things change on a row: a **contested** marker joins the assessment column, and the **Updated** caption can say which neighbor moved the row. One thing changes on the page: the bottom **connections panel** goes, and each connection becomes a **subrow** under a row it joins.

---

## 1. Objective

A researcher on the Promote page should see, on the row itself, everything that decides what Done files.

**What changes.**

1. A row whose comparisons include a conflict shows **Contested** beside its band, at any band. A strong match with one contradicting value no longer looks clean.
2. When deciding one row moves another, the moved row's **Updated** line says which neighbor moved it and how.
3. Each connection (participation, location, relationship, place relationship) shows as a **subrow** under a row it joins. Its state follows its ends: it files, it waits for an end, or both ends are one record. There is no switch. The bottom connections panel goes; the footer keeps its count.

**What stays.** The row's columns and their widths, the target menu and its items, band badges and their tones, the Decided badge, the conflict-with-decided and duplicate callouts under a row, grouping by kind or by assessment, the anchors block, Done and Cancel.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A band is evidence strength | Strong, Medium, Weak, None, drawn as today's Badge tones (success, info, warning, neutral). A conflict never lowers the band. |
| Contested is separate from the band | A row is contested when any comparison behind its proposed pair conflicts: a property (sex at birth, a birth year) or a link (a different spouse on a one-of link). Any band can be contested, Strong included. |
| Contested describes the row's handle | It is about the record in the target menu. It shows only when the assessment column shows, which is when the target is a record. A row set to New or Skip shows no band and no Contested. |
| A changed target clears it | When the researcher picks another record, the marker disappears until the page re-proposes, then shows again if the new pair conflicts. The re-proposal takes well under a second. |
| Updated compares proposals | A suggested row whose target changed since the last proposal shows the Updated badge and a caption "Updated · <why>". Today the why is the row's reason ("via Mary Smith → PER-4KX2M, mother of"). |
| A move is why it changed | Re-proposing can release a row's earlier pair because of a neighbor: the neighbor was matched, was matched to a different record, or was released. The row knows which neighbor. |
| Where a move shows | Only in the Updated caption, replacing the reason. A move on the first proposal (nothing the researcher saw has changed) shows nothing. |
| A connection has two ends | Each end is a row on this page, a record already filed (an anchor), or a Subject not on the page. |
| A connection's state | **Files**: both ends will be records on Done. **Waits**: an end is Skip, unset, or not on this page. **Both ends one record**: both ends land on the same record; Done refuses it. |
| One host per connection | Participation and location: under the **event**. Place relationship ("Toronto · part of · Ontario"): under the **part** place (Toronto). Relationship: directed types ("Mary · parent of · John") under the end named **last** (John); spouse and other symmetric types under the end named first. |
| The other end points at it | A row that is the non-host end of one or more connections shows a compact count, "+2 connections", that scrolls to and highlights the host's subrows. |
| An end that is not a row | When the host end is not a row on this page (an anchor, or not on this page), the subrow goes under the visible end and names the other: "Toronto, already filed". |
| Already-filed connections | Hidden, as today. |
| Busy rows | A Source event can have eight participants. A row shows its first **three** subrows, then "+5 more" to expand in place. |
| Grouping | Subrows travel with their host row in both groupings. |
| Disputing a connection | Not on this page. The way back is the Evidence graph (edit or delete the bridge). This board draws no switch, menu or action on a subrow. |
| Footer | Today: "File 12 · Skip 3 · 22 connections" and a ghost button "22 connections will be filed" that opens the panel. The button goes; the summary text stays. |

### 2.1 What this board is not

- The evidence sheet (S11-D2). A subrow does not open it.
- A contested marker per alternative in the target menu. The proposal only says whether the row's own pair is contested.
- Why a row is contested (which comparison). The evidence sheet shows the conflicting lines; S11-D2 words them.
- A switch, menu, or edit action on a connection.
- Changes to the target menu, the header, or the section headers.

---

## 3. Implementation gate

| Ships in **S11-18** | Ships in **S11-27** | Stays out |
| --- | --- | --- |
| CM-1 – CM-6: the Contested marker in the assessment column, its VoiceOver reading, the move sentences in the Updated caption. | CS-1 – CS-9: connection subrows under their host, the "+N connections" pointer, "+N more", states, VoiceOver, removing the bottom panel and its footer button. | A per-alternative contested mark (not planned). Which comparison conflicts (S11-20, S11-D2). Opening the sheet from a subrow (S11-20 decides). |

---

## 4. Requirements

### 4.1 Contested and moves (S11-18)

| ID | Requirement |
| --- | --- |
| CM-1 | A row with a record target whose pair is contested shows a **Contested** Badge right after the band Badge in the assessment column, before the reason text. Tone **danger**, icon **warning** triangle, text "Contested". It shows at Strong, Medium and Weak. None appears only on a row where the researcher picked a record the evidence rates below Weak; it shows Contested too. |
| CM-2 | A clean row, a row set to New or Skip, and an anchor row show no Contested Badge. The assessment column of those rows is unchanged. |
| CM-3 | The band Badge, Contested Badge and reason text share one line. When the column is narrow, the reason text truncates first; both Badges stay whole. |
| CM-4 | VoiceOver reads the assessment button as today, with "contested" after the band: "Evidence for Mary Smith: Strong, contested, via John Smith → PER-7QH3D (son)". A clean row reads as today: "Evidence for Mary Smith: Strong, via John Smith → PER-7QH3D (son)". |
| CM-5 | A row with the Updated badge and a move reads its caption from the move, naming the neighbor as the via reason does (the neighbor's name, then its record ref, or its Source ref when it has no record): **placed** "Updated · after John Smith → PER-7QH3D was matched"; **moved** "Updated · after John Smith moved to PER-7QH3D"; **released** "Updated · after John Smith was unmatched". |
| CM-6 | A row with the Updated badge and no move keeps today's caption, "Updated · <reason>". A row with a move and no Updated badge shows no caption. |

### 4.2 Connection subrows (S11-27)

| ID | Requirement |
| --- | --- |
| CS-1 | Each connection whose host end is a row on this page shows as a subrow under that row, indented under the name column: the connection mark, "Mary Smith · parent of · John Smith", and its state at the right. |
| CS-2 | **Files**: the subrow reads in primary text with state "Files on Done". **Waits**: the subrow is dimmed (muted text) with state "Waits for John Smith", naming the end that is Skip, unset, or not on this page. **Both ends one record**: the subrow shows a compact warning Callout, "Not filed · both ends are PER-7QH3D". |
| CS-3 | A row that is the other end of connections hosted elsewhere shows "+N connections" as a ghost Button after its badges. Pressing it scrolls to the first host row and highlights its subrows for a moment. N counts connections, not host rows. |
| CS-4 | When the host end is not a row on this page, the subrow goes under the end that is a row and names the other end with its status: "Toronto, already filed" or "Ontario, not on this page". |
| CS-5 | A row shows its first three subrows, ordered files, waits, both-ends-one-record, then by sentence. More than three adds a "+N more" ghost Button that expands the rest in place and becomes "Show fewer". |
| CS-6 | Changing a row's target updates every subrow that row is an end of, under its own row and elsewhere, without a re-layout jump. Setting John to Skip dims "Mary Smith · parent of · John Smith"; setting him to New turns it back to Files. |
| CS-7 | Subrows move with their host in both the kind grouping and the assessment grouping. |
| CS-8 | VoiceOver reads a subrow as one element that names its host: "Mary Smith, parent of John Smith. Files on Done." / "… Waits for John Smith, who is skipped." / "… Not filed, both ends are PER-7QH3D." |
| CS-9 | The bottom connections panel and the footer's "N connections will be filed" button are gone. The footer summary stays: "File 12 · Skip 3 · 22 connections", where 22 counts the connections that file. |

---

## 5. Suggested frames

Frames 1 and 2 are S11-18's; agreeing them unblocks S11-18. Frames 3–6 are S11-27's.

1. **Rows by assessment**, Strong section: a clean "Mary Smith" row and a contested "Ann Ames" row (CM-1, CM-2); Medium section: one contested row; Weak section: one contested row whose reason is too long for the assessment column at a 1200 pt page, truncated (CM-3). The Updated and Decided badges keep their own fixed column at the right.
2. **After deciding John**: Mary's row with Updated and "Updated · after John Smith moved to PER-7QH3D" (CM-5); a second Updated row with today's via caption (CM-6).
3. **An event row with subrows**: a Birth with four participations, three shown and "+1 more" (CS-1, CS-5), one subrow dimmed because its person is Skip (CS-2).
4. **A person row as the other end**: John's row with "+2 connections", and the event's subrows highlighted after pressing it (CS-3).
5. **Edge subrows**: one "Toronto, already filed" subrow (CS-4) and one "both ends are one record" subrow (CS-2).
6. **Footer** without the connections button (CS-9).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Promote row | Snowflake | **Extend** | `Features/Promote/PromoteView.swift` `PromoteAlignmentRow` | Gains the Contested Badge (S11-18), subrows and the "+N connections" Button (S11-27). Columns unchanged. |
| Band, Updated, Decided marks | Component | Ship | `DesignSystem/Components/Badge/PVBadge.swift` | As today. |
| Contested mark | Component | Ship | `DesignSystem/Components/Badge/PVBadge.swift` (`init(_:tone:icon:subtle:)`, tone `.danger`, icon `.warning`); connection marks are `DesignSystem/Recipes/Marks/PVMark.swift` | A Badge, not a new control and not a Chip. Not a Callout: it sits in the assessment line. |
| Updated caption | Snowflake | **Extend** | `PromoteAlignmentRow` (caption under the row) | Same type and indent; the text may be a move sentence (CM-5). |
| Connection subrow | Snowflake | **Rethink** | `Features/Promote/PromoteView.swift` | Composes `PVMark` (the connection mark), text, and the state text; dimmed with muted text, not opacity on the whole row. One call site: not a kit primitive. Rethinks `connectionRow` from the panel into a subrow. |
| Self-link warning | Component | Ship | `DesignSystem/Components/Callout/PVCallout.swift` (`compact: true`, tone `.warning`) | As the row's conflict and duplicate callouts. |
| Subrow highlight (CS-3) | Snowflake | **Extend** | `Features/Promote/PromoteView.swift` `PromoteAlignmentRow` | Short-lived view state with a `PVColor.surfaceSelected` fill. No new token, color, animation or kit modifier: one call site. |
| "+N connections", "+N more", "Show fewer" | Component | Ship | `DesignSystem/Components/Button/PVButton.swift` (`.ghost`, `.sm`) | Not links, not Chips. |
| Connections panel | Snowflake | **Remove** | `PromoteView` `connectionsPanel`, `connectionRow`, footer toggle | Goes in S11-27. |
| Connection switch | Component | **Remove** (here) | `PVSwitch` on `connectionRow` | Goes with S11-26 / S11-27. `PVSwitch` stays in the kit. |
| Footer summary | Snowflake | Ship | `PromoteView` `footer` | Text unchanged. |

**Non-goals.** No new kit primitive. No local Badge, Button or Callout. No switch, menu or hover action on a subrow. No disclosure triangle on a row for its subrows.

---

## 7. Handoff

When the board is agreed: record the board link in the S11-18 plan's **Board** line, move this brief to the spike's design archive folder when S11-27 (the last PR it gates) lands, update [`README.md`](README.md), and move S11-18 from `Blocked — on board S11-D1` to `Planned` with `plan-pr`.
