# S8-D4 — Source page enhancements

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-07** (one Source-page chrome pass; more items join this brief)  
**Depends on:** Shipped Source page ([`SourcePageView`](../../../../macos/App/Features/Sources/SourcePageView.swift), [`SourcePageIdentityHeader`](../../../../macos/App/Features/Sources/SourcePageIdentityHeader.swift)); list dual-action already opens the graph ([`SourcesListNavigation.graphLocation`](../../../../macos/App/Features/Sources/SourcesListNavigation.swift)); `sourceSurface` page vs graph  
**Related:** interpretation-graph-ui leftover “control on the Source page”; pair with graph → page on **S8-D3** / **S8-06**  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of the shipped Source page.

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

A **Source-page enhancements** pass. This board and **S8-07** are a **bundle**: more items will be added here as they are scoped. Do **not** open a second Source-page PR unless a later item cannot share the same pass.

**Items so far**

1. **Jump to Evidence graph.** The Sources list already has dual action (file vs graph). The Source **page** has no control that opens that Source’s Evidence graph. Dogfood starts “blind”: you are on the filing page and bounce back to the list, or create a card just so Add property can open an Artifact. Add a **quick link** on the page that `go(to:)` the same Source with `sourceSurface: .graph`.
2. **Metadata dates are text.** Catalog date fields (`record_date`, `issue_date`, `publication_date`, …) use the **same text editor** as author. Drop the wording + DateValue modal, the date pencil, and any `data_type = date` chrome on this page. The researcher types a reference date; omnibar search matches that string. Structured DateValue stays on Interpretation Observations — not here. Decision: [`source-layer-data-model.md`](../../../source-layer-data-model.md) §1.2 / §5; later sort-by-provenance-time is parked in [`source-provenance-date.md`](../../../ideas/source-provenance-date.md).

Do **not** redesign the identity header, artifacts accordion, or notes stream. Do **not** restyle the metadata column — only collapse date rows onto the existing text editor (and later bundle items).

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
| Catalog metadata is filing text | Date-named fields are `value_text` only. No DateValue picker, no second value column. |
| DateValue lives on Observations | The composer / graph date editor stays. Do not delete `DateValueEditorForm`; remove it from **this** page. |
| Provenance-time sort is not this spike | Do not invent a first-class publication-date control or a Sources-list date sort. Parked: [`source-provenance-date.md`](../../../ideas/source-provenance-date.md). |

### 2.1 What this board is not

- Not graph chrome (**S8-D3**).
- Not Source-page commentary (`mentions` / `remark`) — parked in [`source-to-source-relationships.md`](../../../ideas/source-to-source-relationships.md).
- Not composer rethink / pinning (**S8-D7**).
- Not PDF Find on this page.
- Not Sources-list counts (“12 subjects”) — **S8-D5** / **S8-08**.
- Not a first-class Source provenance date or list sort by catalog time.

---

## 3. Implementation gate (S8-07)

| Ships in **S8-07** | Does **not** ship there |
| --- | --- |
| Always-visible jump to this Source’s Evidence graph when `hasArtifact` | Opening the graph with zero Artifacts |
| Disabled + honest copy when no Artifact | A second Sources list action |
| L10n + VoiceOver; history Back to the page | Commentary surface |
| Metadata date rows use the same text editor as other fields; no DateValue modal on this page | A first-class provenance-date field or Sources-list date sort |
| Catalog + FFI: drop `data_type = date` / `date_value_id` on source metadata; keep `value_text` | Removing DateValue from Observations / the composer |
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
| SP-8 | Every metadata row (including former date fields) edits as **text** with the existing saved-row / suggestion editor. No date modal, no structure pencil, no `data_type` badge that implies a DateValue. |
| SP-9 | Suggestion and extra-field add for a date-named key is the same as author: type, Save. Empty stays a suggestion. |
| SP-10 | Do not show or edit DateValue components on this page. Composer / graph date chrome is unchanged. |
| SP-11 | Further items get their own `SP-n` rows when scoped. |

---

## 5. Suggested frames

1. Source with Artifacts — jump enabled in the identity header (or chosen slot).
2. Source with no Artifact — jump disabled + reason; Add Artifact still available.
3. After jump — Evidence graph for the same Source; Back returns to the page.
4. Metadata — date-named field empty suggestion: same dashed text row as author.
5. Metadata — date-named field with a value: inline text edit / save / clear; no date dialog.
6. *(Add frames here as more bundle items are scoped.)*

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Source page shell | Snowflake | **Extend** | `Features/Sources/SourcePageView.swift` | Pass navigation if the header does not already have it. Drop the date-dialog host. |
| Identity header | Snowflake | **Extend** | `Features/Sources/SourcePageIdentityHeader.swift` | Likely host for the jump. |
| Source page model | Snowflake | **Extend** | `Features/Sources/SourcePageModel.swift` | `hasArtifact` from workspace / source row. |
| Metadata section / view | Snowflake | **Rethink** (date rows only) | `Features/Sources/SourceMetadataSection.swift`, `SourcePageMetadataView.swift` | All values through the text editor. Remove `dateValueCell`, `SourcePageDateEditorForm`, `openDateEditor`. |
| Metadata text editor | Snowflake | Ship | `SourcePageMetadataTextEditor` (same file) | Date-named fields use this. |
| DateValue editor form | Snowflake | **Remove** from this page | `Features/Dates/DateValueEditorForm.swift` | Keep for the citation composer. Do not instance here. |
| Sources list navigation | Snowflake | Ship | `Features/Sources/SourcesListNavigation.swift` | Reuse `graphLocation`. |
| Button | Component | Ship | `DesignSystem/Components/Button/` | Evidence graph action. |
| Input / Field | Component | Ship | `DesignSystem/Components/Input/`, `Field/` | Metadata text rows. |
| Workspace location tests | Test | **Extend** | `ProvenenciaTests` | Page → graph location; disabled when no Artifact. |
| Source page metadata tests | Test | **Extend** | `ProvenenciaTests/SourcePageModelTests.swift` | Date-named field saves as text; no date payload. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Page-local Back that bypasses history | Toolbar Back / `go(to:)` only. |
| Graph-specific counts on the page | List counts are **S8-D5**; not this page. |
| `mentions` / `remark` commentary | Parked: [`source-to-source-relationships.md`](../../../ideas/source-to-source-relationships.md). |
| DateValue (or NameValue) on metadata | Catalog filing is text. |
| First-class provenance-date control | Parked: [`source-provenance-date.md`](../../../ideas/source-provenance-date.md). |

---

## 7. Out of scope

- Restyling filing (notes, artifacts ingest, metadata **layout**) — date rows becoming text is in scope
- Composer / graph DateValue chrome
- Source-page Find
- Commentary (“what other Sources say about this one”) — [`source-to-source-relationships.md`](../../../ideas/source-to-source-relationships.md)
- Sorting or grouping Sources by catalog time — [`source-provenance-date.md`](../../../ideas/source-provenance-date.md)

---

## 8. Handoff

1. Keep this brief open until the **S8-07** bundle is frozen, then archive it.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S8-07** against the board.
