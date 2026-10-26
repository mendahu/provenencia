# S12-D5 — Sources list graph tokens

**Kind:** Claude Design board (**rethink** of the Sources list row; **enhancement** of the Source page's Evidence graphs cards)
**Spike:** Provenencia Spike 12 (Source structure)
**View:** Sources list — every row becomes a Source link over a line of graph tokens, one per Evidence graph; the Status filter reads each graph. Source page — each Evidence graphs card gains the count line.
**Implements later as:** PR **S12-14**
**Board:** agreed 2026-10-26 — [Claude Design](https://claude.ai/design/p/118d2eb7-3b46-4912-a688-09ec7b0070fa?via=share) (`S2-04 Sources list`, frames 1–13)
**Depends on:** S12-08 (a graph is a place), S12-D2 (the Source page graph cards this board adds a line to)
**Related:** [`../../source-structure.md`](../../source-structure.md#where-graphs-show) (the design), [`../../deployment-plan.md`](../../deployment-plan.md) working rules "Sources list", "Status is per graph", "Source page", "One graph list". S5-D2 drew the split row this board replaces; S12-D6 added its right-click menu, which this board keeps.
**Design system layers:** [`docs/design-system-layers.md`](../../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../../.agents/skills/add-ui-component/SKILL.md)

**History.** The brief as first written (2026-10-26) asked for an accordion: an expander in the row's 248 pt graph zone over one sub-row per graph. The board drew a line of graph tokens instead, and the user adopted it at board review (2026-10-26), with compact token counts that explain themselves in a hover tooltip. This file records the agreed board; requirement IDs are the board's.

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

### This brief is a rethink of the row

The Sources list keeps its header, count line, toolbar (search, type, status, sort), empty and page-level states, Add source sheet and Copy… menu. The **row** is redrawn: the two-zone split row (page zone | 248 pt Evidence graph zone) and its two-column caption band are replaced. On the Source page this board adds **one line** to each Evidence graphs card and changes nothing else.

---

## 1. Objective

A census Source can hold several Evidence graphs. The Sources list had one **Open graph** zone per Source, which opened the alphabetically first graph and showed only its counts. The Source page listed the graphs without progress.

**What changes.**

- **Sources list row.** Two lines. The top line is the **Source link** (cover, title, type; opens the Source page), then Updated and the ref. The second line is the **token line**: one **graph token** per Evidence graph, in name order, always shown, wrapping. A token shows the graph's name and its compact count ("12 · 40", or "not started") and opens that graph's canvas. Hovering a token shows what the count means ("12 subjects · 40 observations"). Nothing expands.
- **Status filter.** Read per graph (§2); the last option reads **Promoted**.
- **Source page.** Each Evidence graphs card shows the full count line under its ref, above its description.

**What stays.** Search matches Source title and ref only. The row's Copy… menu. The Updated cell and the ref. A Source with no Artifact still says "Needs an artifact".

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| Every Source has at least one graph | Every row with an Artifact has at least one token; a migrated project's rows show one "Default" token. |
| Graphs list alphabetically by name | Tokens are in name order. "Default" sorts with the other names. |
| Open needs a linked Artifact | A graph that links no Artifact has a blocked token: dashed, faint, inert, reading "{name} needs an artifact", with help "Add an artifact to open this evidence graph". |
| A Source with no Artifact | One dashed "Needs an artifact" token in place of its graphs, with today's help. |
| Counts are per graph | A token counts its graph only. There is no Source-level total. |
| Token count | "{subjects} · {observations}" in mono micro, or italic "not started" with 0 subjects. Its help spells it out: "12 subjects · 40 observations", "0 subjects · not started". Before counts arrive: a 32 × 6 rule in the count's place; the token still opens. |
| Full count line | On the Source page cards: "N subjects · M observations", or "0 subjects · not started", or a loading rule. |
| Graph status | Empty: no subjects. In progress: at least one subject, none promoted. Promoted: at least one subject promoted by any run. A Source shows under a status when any of its graphs has it, so it can show under several; its token line always shows every graph. |
| The graph list loads apart from the Sources | Rows paint first. Until the graph list arrives, one token-sized skeleton (96 × 6 rule) stands in for the token line. If it fails: one inert token with the danger alert icon, carrying the app's standard error as help. |
| Counts update live | Adding an Observation changes only that graph's token, and its Source page card, without a reload. |

### 2.1 What this board is not

- Searching graphs (S12-15, omnibar).
- Create, edit, delete or link graphs from the list; Promote from the list; status badges; a Source-level total.
- Selection, multi-select, a detail pane, or anything that expands.
- Any change to the Source page card other than the count line.

---

## 3. Implementation gate (S12-14)

| Ships in **S12-14** | Stays out |
| --- | --- |
| The two-line row (Source link, Updated, ref; token line); graph tokens with counts, help, blocked, loading and failed states; the one-line caption; the per-graph status rule and the Promoted label; the count line on the Source page cards. | Graph search (S12-15). Graph create, edit, delete and links from the list. Promote from the list. Shift-F10 for the row menu (the kit's right-click menu has no keyboard trigger; logged). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| GT-1 | **Row.** A group, not a button: no row hover and no row hit area. Padding 12 × 32, 8 between its two lines, a 1 px `--border-subtle` divider. Top line: Source link · Updated (136) · ref, gap 12. Second line: the token line, indented 56 to the title. |
| GT-2 | **Caption.** One caption band above the list: "Source · evidence graphs". |
| GT-3 | **Source link.** Cover 44, title (14 medium, one line, truncates) and type (13 muted) are one button that opens the Source page. Hover in the kit's link style: the title turns `--text-link-hover` and is underlined at a 2 px offset, the cover outline steps to `--border-strong`, over 130 ms; no fill. Focus ring around the button, 4 px pad. |
| GT-4 | **Token.** One per graph, name order, all shown, wrapping (gap 6). 26 tall, `--radius-sm`, 1 px `--border-default` on `--surface-card`, padding 0 × 9: network glyph 13, name (13, max 300 wide, truncates), count (mono 11) "12 · 40" or italic "not started". Hover fills `--surface-hover` and the border goes to `--border-strong`. A click opens that graph's canvas. Help: the full count ("12 subjects · 40 observations"). |
| GT-5 | **Blocked token.** A graph that links no Artifact: dashed, faint, inert token "{name} needs an artifact"; help "Add an artifact to open this evidence graph". A Source with no Artifact: one dashed "Needs an artifact" token with today's help. Blocked tokens stay focusable so their hint can be read. |
| GT-6 | **Loading and failure.** Graph list pending: one skeleton token (96 × 6 rule on `--surface-inset`). Counts pending: each token shows a 32 × 6 rule in the count's place and still opens. Graph list failed: one inert token with the alert glyph in `--danger`, the standard error as help. |
| GT-7 | **Status filter.** Options: All statuses, Empty, In progress, Promoted. A Source shows under a status when any of its graphs has it (§2). |
| GT-8 | **Copy… menu.** Right-click on a row opens the existing Copy… menu (not on a token). |
| GT-9 | **Keyboard.** Tab order: Source link, then tokens in name order, then the next row. Return or Space activates. |
| GT-10 | **VoiceOver.** The row group is labelled with the title. Source link: today's label and value ("Open source page"; "{title}, {ref}, updated {when}"), with Copy… in its actions. Token: "Open graph {name}, {N} subjects, {M} observations", or "…, 0 subjects, not started", or "…, counts loading". A blocked token adds its hint. |
| GT-11 | **Minimum width (650).** Title and type truncate; Updated and the ref keep their width. Twelve tokens wrap onto as many lines as they need; a 60-character name truncates inside its token. |
| GT-12 | **Source page.** Each Evidence graphs card shows the full count line on its own line under the ref, above the description. Open, Edit and Delete do not move. |

---

## 5. Frames (as agreed)

Board file `S2-04 Sources list`, content host 1136 wide.

1. **Sources list, interactive** — eight Sources; the 1851 census with tokens Default "not started", Jones household "12 · 40", Smith household "7 · 19"; two Sources with "Needs an artifact" (GT-1 – GT-5).
2. **Row states** — 2a hover on the census Source link; 2b help on "Yesterday"; 2c a graph with no linked Artifact, help shown; 2d graph lists pending and census counts pending; 2e graph list failed; 2f row menu at the pointer; 2g after Copy…: the copy at the top with one empty Default token; 2h the Copy Source form (Source page board) (GT-3, GT-5, GT-6, GT-8).
3. **Toolbar** — status menu open; "In progress: 2 of 8"; search "smith" (title or ref only); nothing matches; status set with counts pending; status set with counts failed (GT-7).
4. **Status, read per graph** — the census (Default empty, Jones in progress, Smith promoted) under Empty, In progress and Promoted, every token shown each time (GT-7).
5. **Minimum width** — 650 host; twelve tokens wrap; the 60-character Pryke name truncates (GT-11).
6. **Empty destination**, 7. **Page-level states**, 8. **Add source**, 10. **Thumbnail** — unchanged from today.
9. **Row metrics** — row padding 12 × 32, 8 between lines; top line `minmax(0,1fr) 136 auto`, gap 12; cover 44; title 14 medium / type 13 muted; Source link pad 4; Updated 136; ref 11 mono muted; token line indent 56, gap 6, wraps; token 26, padding 0 × 9, `--radius-sm`, max 300; glyph 13 / name 13 / count 11 mono; loading rules 96 × 6 (list) and 32 × 6 (count) on `--surface-inset`; divider 1 px `--border-subtle`. "Two hover styles, two kinds of target": the Source link hovers like a link (ink and underline, no fill), a token like a control (fill and outline); dashed outlines only on blocked tokens.
11. **Navigation** — Source link → Source page; token → that graph's canvas by graph id, inert when it links no Artifact. Nothing expands; no pane opens.
12. **Source page count line** — the census cards: Default "0 subjects · not started", Jones "12 subjects · 40 observations" with a description, Smith "7 subjects · 19 observations" (GT-12).
13. **Building blocks** — the inventory below.

---

## 6. UI building-block inventory

This table is **binding**. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Row | Snowflake | Rethink | `Features/Sources/SourcesListRow.swift` (new) | Source link + Updated + ref, then the token line. Replaces `SourcesSplitRow` and `SourcesSplitLayout`. |
| Source link | Snowflake | Rethink | `Features/Sources/SourcesListRow.swift` (new) | A button with a feature link style (title ink and underline on hover, cover outline), replacing `SourcesSplitZoneButtonStyle`. Not `PVButton(.link)`: that underlines at rest and has no cover. |
| Graph token | Snowflake | **New** | `Features/Sources/SourcesGraphToken.swift` (new) | Its states (open, counts pending, blocked, list pending, list failed) and spoken label. One call site (the row), no second call site known; a Sources snowflake, not a kit primitive. Not `PVChip`: a chip is a selectable pill, and this is a link to a place with no selection state. |
| Token line wrap | Component | **New** (promoted) | `DesignSystem/Components/Flow/PVFlowLayout.swift` (new) | The Citation composer's private `FlowLayout` (`CitationComposerAuthoringBox.swift`), promoted at its second call site (the Sources token line; the first is the Citation composer). |
| Dashed stroke | Token | **New** | `DesignSystem/Tokens/PVStroke.swift` (new) | The blocked token's dash, and the five existing `[4, 3]` copies (`PVCard`, `PVEmptyState`, `PVChip` and two features) moved onto it: a second call site already exists. |
| Full count line | Snowflake | **New** (lifted) | `Features/Sources/SourcesGraphCountLine.swift` (new) | Today's private `SourcesSplitRow.graphCountLine`, for the Source page cards; no second call site, a Sources snowflake. |
| Cover | Component | Ship | `Platform/CachedThumbnail.swift` over `PVThumbnail` | 44. |
| Updated | Recipe | Ship | `PVListDetailCell` + `PVLastUpdated` | Today's cell, 136. |
| Row menu | Component | Ship | `pvRightClickMenu` / `PVActionMenuItem` | Today's Copy…. |
| Toolbar, states, Add source | Component | Ship | `PVInput`, `PVSelect`, `PVEmptyState`, `PVCallout`, `PVFormDialog` | Unchanged; only the status option label changes. |
| Loading rule | Component | Ship | `PVSkeletonBar` | 96 × 6 and 32 × 6. |
| Failure glyph | Component | Ship | `PVIcon(.danger)` in `PVColor.danger` | The kit's alert glyph. |
| Source page graph card | Snowflake | Extend | `Features/Sources/SourcePageGraphsView.swift` `SourcePageGraphRow` | Adds the full count line. |

**Non-goals.** No disclosure, accordion or tree. No status badge, progress bar or meter. No `PVChip` for tokens. No local Button or Card.

---

## 7. Handoff

Board agreed 2026-10-26. This brief is archived; the S12-14 plan carries the requirement IDs, the inventory and the frames in words.
