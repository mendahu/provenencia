# S8-D5 — Sources list design refresh

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-08** (one Sources-list chrome pass; more items join this brief)  
**Depends on:** Shipped split-row list ([`SourcesListView`](../../../../macos/App/Features/Sources/SourcesListView.swift), [`SourcesSplitRow`](../../../../macos/App/Features/Sources/SourcesSplitRow.swift)); session cache ([`CatalogQueryKey`](../../../../macos/App/Features/Workspace/Session/CatalogQueryKey.swift), [`macos-client-patterns.md`](../../../macos-client-patterns.md) §1)  
**Related:** interpretation-graph-ui leftover “12 subjects, 3 uncited”; S5-D2 **SL-8** deferred graph columns; pair with Source-page jump (**S8-D4**)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md); cache keys via [`add-workspace-place`](../../../../.cursor/skills/add-workspace-place/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of the shipped Sources list.

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
| Short create / edit form | FormDialog |
| Status / count / polarity mark | Badge; compact token → Chip |
| Cover or file thumb | Thumbnail |
| Grouping / raised or sunken row | Card |
| Section title | SectionHeader |
| Transient after-save notice | Toast |
| Native menu of actions | ContextMenu |

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

---

## 1. Objective

A **Sources list design refresh**. This board and **S8-08** are a **bundle**: more items will be added here as they are scoped. Do **not** open a second list PR unless a later item cannot share the same pass.

The list already has dual action (file vs graph). What it cannot show is **whether the Evidence graph has been worked**. Some Sources are filled in; some are still empty. That has to be visible at a glance without opening the graph.

**Items so far**

1. **Graph-progress counts.** Show **subject count** and **observation count** for each Source (exact placement, density, and empty/zero treatment are board findings). The point is “has this graph been filled out?”, not a second Interpretation list. Optional extra if it stays quiet: uncited subject count (the original leftover phrasing was “12 subjects, 3 uncited”).

Do **not** throw away the split-row contract (page zone | graph zone, no-Artifact disabled). Refresh chrome around it; do not invent a third destination.

```text
[ cover ]  1851 England Census          │  Evidence graph
           SRC-…  ·  Census             │  12 subjects · 48 observations

[ cover ]  Unread parish register       │  Evidence graph
           SRC-…  ·  Register           │  0 subjects

[ cover ]  Fileless placeholder         │  Needs an Artifact
           SRC-…  ·  Certificate        │
```

Exact typography, whether counts live in the graph zone or a new column, and how zero vs “not started” reads are board findings. Prefer existing type scale / muted color; no new kit primitive unless raised.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Graph is Source-scoped | Counts are **this** Source’s graph, not project-wide. |
| Canvas subjects are `source_id = ?` minus `source` type | Subject count matches what the graph draws (person / event / place / bridges). Do **not** count `source` subjects. |
| Observation count is “how much has been cited” | Count Observations that belong with this Source’s graph (subjects homed here, or Citations under this Source’s Artifacts — **S8-08** picks the SQL; both should agree with “filled out”). |
| Uncited = subject with zero Observations | Optional third number. Working `label` is not an Observation. |
| No Artifact → no Citations | Graph zone stays blocked. Counts may still be 0 / omitted; do not imply the graph is openable. |
| Dual action stays | Counts are status, not a third click target. Click still opens page or graph. |
| S5 deferred this on purpose | Folding counts into `[CatalogSource]` would restale the **whole** Sources list on every canvas write. That is the bug to avoid. |

### 2.1 Cache (implementation constraint — plan it on this brief / S8-08)

The list payload (`CatalogQueryKey.sourcesList`) must **not** grow subject/observation numbers. Canvas create / connect / Add property already invalidates `.sourceGraph` for one `sourceId`. The list must not ride that wave.

**S8-08** designs a **narrow count cache** so a write invalidates **that Source’s numbers only**:

- Precedent: `.citationCounts(project:sourceId:)` is already a per-Source derived key; composer mutations do not bust `sourcesList`.
- One cache owns each list ([`macos-client-patterns.md`](../../../macos-client-patterns.md) §1). Derived counts are a **new key** (or a small map that can `setQueryValue` / invalidate one `sourceId`), not columns on `CatalogSource`.
- Mutations already carry `sourceId` (`.createdSubject`, `.createdCitation`, `.addedObservations`). Wire `invalidateOn` to the count key for that id — not `.sourcesList`, not `allCached(.sourceGraph)`.
- Warm counts with the Sources place (visible rows, or all — **S8-08** chooses; do not N+1 from `body`).
- Stale-while-revalidate is fine: show last counts while the one Source refetches.

The board does **not** invent the key enum. It must leave room for counts to arrive asynchronously (placeholder / zero / “—” while loading) so the list chrome does not assume they are on the row model today.

### 2.2 What this board is not

- Not graph chrome (**S8-D3**) or Source-page jump (**S8-D4**).
- Not source-to-source commentary ([`source-to-source-relationships.md`](../../../ideas/source-to-source-relationships.md)).
- Not composer rethink (**S8-D7**).
- Not a Subjects / Observations list destination.
- Not putting counts on the Source **page** unless a later bundle item says so.

---

## 3. Implementation gate (S8-08)

| Ships in **S8-08** | Does **not** ship there |
| --- | --- |
| List chrome refresh + subject + observation counts (uncited if the board keeps it) | Folding counts into `GetSources` / `CatalogSource` |
| Per-Source (or patchable) count cache; invalidate one Source on graph writes | Invalidating `sourcesList` on subject / Observation create |
| L10n + VoiceOver for the numbers | A third list destination; filtering the list by “empty graph” unless added to this brief |
| Further items **added to this brief** before the PR starts | A second Sources-list visuals PR |

If more bundle items land after the board is first drawn, **amend this brief** and redraw those frames — still one **S8-08**.

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SL-1 | Sources list shows **subject count** and **observation count** per Source so a worked graph is obvious next to an empty one. |
| SL-2 | Zero / not-started is honest (not a missing cell that looks like a load error). No-Artifact rows keep the blocked graph zone. |
| SL-3 | Counts are not a third navigation action. Page and graph zones stay the two clicks. |
| SL-4 | VoiceOver: counts are in the row tree (graph zone or named group). Do not leave them decorative-only. |
| SL-5 | Prefer existing type / color tokens. No new kit primitive unless the board raises one. |
| SL-6 | Refresh may tighten split-row spacing, caption band, or graph-zone copy to fit the numbers — keep S5-D2 dual-action meaning. |
| SL-7 | Implementation: counts live on their **own cache identity**; a canvas write updates **that Source** only. |
| SL-8 | Further items get their own `SL-n` rows when scoped. |

---

## 5. Suggested frames

1. Mix of worked / empty / no-Artifact rows — counts readable at list density.
2. Close-up of the graph zone with counts (and optional uncited).
3. Counts still loading (async key) vs settled.
4. After adding a subject on the graph and returning — **only that row’s** numbers changed.
5. *(Add frames here as more bundle items are scoped.)*

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Sources list | Snowflake | **Extend** | `Features/Sources/SourcesListView.swift` | Warm count keys with the list. |
| Split row | Snowflake | **Extend** | `Features/Sources/SourcesSplitRow.swift` | Likely host for counts (graph zone). |
| Split layout | Snowflake | **Extend** | `SourcesSplitLayout` | Graph zone is 210pt today; board may widen. |
| Sources model | Snowflake | **Extend** | `Features/Sources/SourcesModel.swift` | Do not hang counts off `CatalogSource`. |
| Count cache key | Session | **Add** | `CatalogQueryKey` + registry | Per `sourceId` (or patchable map). Follow `citationCounts`. |
| Go count query | Engine | **Add** | `core/database` + FFI | Aggregate only; no new tables. |
| Type / muted text | Token | Ship | `PVFont` / `PVColor` | Status numbers, not headlines. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Counts on `CatalogSource` / `sourcesList` payload | Restales every row on one canvas write. |
| A Subjects column that opens a list | Graph is the only Subject surface. |
| Live-updating the list while the graph is open every keystroke | Invalidate / refetch on mutation end is enough. |

---

## 7. Out of scope

- Redesigning Add Source / empty catalog
- Source-page commentary or Evidence graph card chrome
- Composer / pinning
- Sorting or filtering the list by count (unless a later SL item adds it)

---

## 8. Handoff

1. Keep this brief open until the **S8-08** bundle is frozen, then archive it.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S8-08** against the board (Go aggregates + cache key + list chrome).
