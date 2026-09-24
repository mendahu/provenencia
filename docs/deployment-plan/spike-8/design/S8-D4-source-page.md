# S8-D4 — Source page enhancements

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-07** (one Source-page chrome pass; more items join this brief)  
**Depends on:** Shipped Source page ([`SourcePageView`](../../../../macos/App/Features/Sources/SourcePageView.swift), [`SourcePageIdentityHeader`](../../../../macos/App/Features/Sources/SourcePageIdentityHeader.swift)); list dual-action already opens the graph ([`SourcesListNavigation.graphLocation`](../../../../macos/App/Features/Sources/SourcesListNavigation.swift)); `sourceSurface` page vs graph  
**Related:** interpretation-graph-ui leftover “control on the Source page”; pair with graph → page on **S8-D3** / **S8-06**  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

A **Source-page enhancements** pass. This board and **S8-07** are a **bundle**: more items will be added here as they are scoped. Do **not** open a second Source-page PR unless a later item cannot share the same pass.

**Items so far**

1. **Jump to Evidence graph.** The Sources list already has dual action (file vs graph). The Source **page** has no control that opens that Source’s Evidence graph. Dogfood starts “blind”: you are on the filing page and bounce back to the list, or create a card just so Add property can open an Artifact. Add a **quick link** on the page that `go(to:)` the same Source with `sourceSurface: .graph`.

Do **not** redesign the identity header, metadata, artifacts accordion, or notes stream — only add chrome for this jump (and later bundle items).

```text
[ cover ]  1851 England Census     [ Open Evidence graph ]
           SRC-…  ·  Census
```

Exact placement (identity header vs page toolbar), label, and no-Artifact disabled treatment are board findings. Prefer a shipped `PVButton`. Reuse list copy (**Evidence graph**) so the two entry points say the same thing.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Page and graph are already distinct places | Same `sourceId`, `sourceSurface: .graph`. Back returns to the page. |
| List already opens the graph | Reuse `SourcesListNavigation.graphLocation` (or the same `hasArtifact` rule). Do not invent a second location builder. |
| No Artifact → no graph | List disables the graph zone. Page control **disabled** + short reason; do not navigate to an empty/blocked graph. Shortcut to add an Artifact is already on this page. |
| Graph → page is **S8-D3** / **S8-06** | This brief is the other direction only. The pair should feel like the same product (same name, same history rules). |
| `hasArtifact` is already on `CatalogSource` | Drive enablement from the page’s loaded Source / workspace, not a new query. |

### 2.1 What this board is not

- Not graph chrome (**S8-D3**).
- Not Source-page commentary (`mentions` / `remark`).
- Not composer rethink / pinning (dogfood).
- Not PDF Find on this page.
- Not Sources-list counts (“12 subjects”).

---

## 3. Implementation gate (S8-07)

| Ships in **S8-07** | Does **not** ship there |
| --- | --- |
| Always-visible jump to this Source’s Evidence graph when `hasArtifact` | Opening the graph with zero Artifacts |
| Disabled + honest copy when no Artifact | A second Sources list action |
| L10n + VoiceOver; history Back to the page | Schema / FFI; commentary surface |
| Further items **added to this brief** before the PR starts | A second Source-page visuals PR |

If more bundle items land after the board is first drawn, **amend this brief** and redraw those frames — still one **S8-07**.

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SP-1 | Source page shows a control that opens **Evidence graph** for **this** Source (`sourceSurface: .graph`, same `sourceId` / ref / title). |
| SP-2 | Enabled only when the Source has at least one Artifact (same rule as the list). Disabled otherwise; short reason; no navigation. |
| SP-3 | Uses `navigation.go(to:)` and the existing graph location helper. Toolbar Back returns to the Source page. |
| SP-4 | Board picks placement (identity header is the default guess). Do not add a second breadcrumb trail. |
| SP-5 | Copy says **Evidence graph** (not Interpret / Interpretation). Align with the list action. |
| SP-6 | VoiceOver: control is in the page tree; disabled state is announced. |
| SP-7 | Prefer `PVButton`. No new kit primitive. |
| SP-8 | Further items get their own `SP-n` rows when scoped. |

---

## 5. Suggested frames

1. Source with Artifacts — jump enabled in the identity header (or chosen slot).
2. Source with no Artifact — jump disabled + reason; Add Artifact still available.
3. After jump — Evidence graph for the same Source; Back returns to the page.
4. *(Add frames here as more bundle items are scoped.)*

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Source page shell | Snowflake | **Extend** | `Features/Sources/SourcePageView.swift` | Pass navigation if the header does not already have it. |
| Identity header | Snowflake | **Extend** | `Features/Sources/SourcePageIdentityHeader.swift` | Likely host for the jump. |
| Source page model | Snowflake | **Extend** | `Features/Sources/SourcePageModel.swift` | `hasArtifact` from workspace / source row. |
| Sources list navigation | Snowflake | Ship | `Features/Sources/SourcesListNavigation.swift` | Reuse `graphLocation`. |
| Button | Component | Ship | `DesignSystem/Components/Button/` | Evidence graph action. |
| Workspace location tests | Test | **Extend** | `ProvenenciaTests` | Page → graph location; disabled when no Artifact. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Page-local Back that bypasses history | Toolbar Back / `go(to:)` only. |
| Graph-specific counts on the page | Deferred leftover; not this item. |
| `mentions` / `remark` commentary | Different leftover. |

---

## 7. Out of scope

- Redesigning filing (metadata, notes, artifacts ingest)
- Composer / graph card chrome
- Source-page Find
- Commentary (“what other Sources say about this one”)

---

## 8. Handoff

1. Keep this brief open until the **S8-07** bundle is frozen, then archive it.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S8-07** against the board.
