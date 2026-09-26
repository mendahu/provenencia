# S8-D4 — Source page enhancements

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-07** (one Source-page chrome pass; more items join this brief)  
**Depends on:** Shipped Source page ([`SourcePageView`](../../../../macos/App/Features/Sources/SourcePageView.swift), [`SourcePageIdentityHeader`](../../../../macos/App/Features/Sources/SourcePageIdentityHeader.swift)); list dual-action already opens the graph ([`SourcesListNavigation.graphLocation`](../../../../macos/App/Features/Sources/SourcesListNavigation.swift)); `sourceSurface` page vs graph  
**Related:** interpretation-graph-ui leftover “control on the Source page”; pair with graph → page on **S8-D3** / **S8-06**. **Jump to Evidence graph is already on this board** — S8-07 still ships it; do not redesign.  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of the shipped Source page. **Metadata frames on the board:** throw away the date-entry experiments and **redraw from the running Mac app**, then add delete. Do not extend those experiments.

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

**Already on this board (do not redraw)**

**Jump to Evidence graph** is already drawn (identity-header control; disabled when no Artifact). **S8-07** still implements it from those frames. Leave placement, label (**Evidence graph**), and disabled treatment alone. Do not add a second jump or restyle the header to “finish” it.

**This board designs**

1. **Restore shipped metadata chrome, then add delete.** Throw away the board’s date-entry experiments (structure modal, date pencil, split wording/DateValue, extra date chrome). **Redraw the metadata section from the shipped Mac app** ([`SourcePageMetadataView.swift`](../../../../macos/App/Features/Sources/SourcePageMetadataView.swift)) — that quick-add / quick-edit is the one we keep. Every field is text, including former date keys. The **one** piece to keep from the current board (and add to the app) is a **delete** on a saved metadata value. Decision: [`source-layer-data-model.md`](../../../source-layer-data-model.md) §1.2 / §5; later sort-by-provenance-time is parked in [`source-provenance-date.md`](../../../ideas/source-provenance-date.md).

Do **not** redesign the identity header, artifacts accordion, or notes stream. Do **not** invent a new metadata layout — match the shipped section and add delete.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Jump to Evidence graph is already on these frames | Leave it. S8-07 ships the existing control; this pass is metadata. |
| Shipped metadata is the baseline | Match it: header + **Add**; reorderable saved rows; `PVInlineEdit` pencil → textarea + Save/Cancel; dashed **suggestion** rows (input + Save + dismiss); **Add** is a FormDialog (field ComboBox + value). Screenshots / running app beat the board’s date-era frames. |
| Catalog metadata is filing text | Date-named fields use that same text row. No DateValue picker. |
| DateValue lives on Observations | Composer / graph date editor stays. Do not instance it here. |
| Saved value vs empty suggestion | **Delete** clears a saved value (`ClearSourceMetadata` — engine already exists). **Dismiss** hides an empty type suggestion. Do not merge those into one control. |
| After delete | A type-suggested field returns to the suggestion list. An extra field disappears. |
| Provenance-time sort is not this spike | Do not invent a first-class publication-date control or a Sources-list date sort. Parked: [`source-provenance-date.md`](../../../ideas/source-provenance-date.md). |

### 2.1 Shipped metadata (match this)

```text
Metadata                              n fields     [ + Add ]

intro caption

[ ≡  Author          Alice Smith          ✎ ]
[ ≡  Record date     15 May 1880          ✎ ]   ← same row as Author (no date chrome)

Suggested
[ Author                                  [value] [Save]  × ]
```

Add dialog (already shipped): field ComboBox + value Input. Keep it.

**Delete** (from the current board; not in the app yet) sits on a **saved** row — `PVIconButton` via `PVInlineEdit`’s `restingTrailing` (next to the pencil) is the default guess. Confirm if the existing board already confirms; use `Confirm` (`item:` snapshot), not a new primitive.

### 2.2 What this board is not

- Not redesigning **Jump to Evidence graph** (already drawn; S8-07 implements).
- Not keeping the board’s DateValue / “easier date entry” metadata experiments.
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
| Jump to Evidence graph from the **existing** board frames (`hasArtifact` gate, **Evidence graph** copy, Back to the page) | Redesigning that jump; opening the graph with zero Artifacts |
| Metadata section matches shipped quick-add / quick-edit; all fields text; **delete** on saved values (`ClearSourceMetadata`) | Rebuilding Add / suggestions / reorder from scratch; a first-class provenance-date field or list sort |
| Catalog + FFI: drop `data_type = date` / `date_value_id` on source metadata; keep `value_text` | Removing DateValue from Observations / the composer |
| L10n + VoiceOver for delete, date-as-text, and the already-designed jump | Commentary surface; deleting a metadata-**field** vocabulary row (Source Fields page) |
| Further items **added to this brief** before the PR starts | A second Source-page visuals PR |

If more bundle items land after the board is first drawn, **amend this brief** and redraw those frames — still one **S8-07**.

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SP-1 | Redraw metadata from the **shipped Mac section**. Keep header + Add, reorderable saved rows, `PVInlineEdit` quick edit, dashed suggestion quick-add, Add FormDialog. Discard the board’s date-entry experiments. |
| SP-2 | Every saved / suggestion / extra field (including former date keys) is that same **text** path. No date modal, no structure pencil, no `data_type` badge that implies a DateValue. |
| SP-3 | Do not show or edit DateValue components on this page. Composer / graph date chrome is unchanged. |
| SP-4 | Each **saved** metadata row has a **delete** that clears the value (`ClearSourceMetadata`). Not the suggestion dismiss ×. Suggested fields return to the suggestion list; extras vanish. |
| SP-5 | Leave the existing Jump to Evidence graph frames untouched. |
| SP-6 | Further items get their own `SP-n` rows when scoped. |

---

## 5. Suggested frames

1. Metadata — shipped section restored (saved + suggestions + Add). Date-named fields look like author. Existing jump stays as drawn.
2. Metadata — saved row: pencil edit as today, plus **delete**. After delete, a suggested field is an empty dashed row again.
3. Metadata — extra (non-suggested) field deleted: row gone; field available in Add again.
4. *(Add frames here as more bundle items are scoped.)*

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Source page shell | Snowflake | **Extend** | `Features/Sources/SourcePageView.swift` | Drop the date-dialog host. Do not restyle around the existing jump. Host delete Confirm if the board keeps one. |
| Identity header | Snowflake | Ship | `Features/Sources/SourcePageIdentityHeader.swift` | Jump already designed here. Leave it. |
| Source page model | Snowflake | **Extend** | `Features/Sources/SourcePageModel.swift` | Wire `clearSourceMetadata` for delete. |
| Metadata section / view | Snowflake | **Extend** | `Features/Sources/SourceMetadataSection.swift`, `SourcePageMetadataView.swift` | Restore shipped rows; remove `dateValueCell` / `SourcePageDateEditorForm` / `openDateEditor`; add delete on saved rows. |
| Metadata text editor | Snowflake | Ship | `SourcePageMetadataTextEditor` (same file) | Date-named fields use this. Delete is `restingTrailing`, not a fork of InlineEdit. |
| Inline edit | Component | Ship | `DesignSystem/Components/InlineEdit/` | Pencil / Save / Cancel as shipped. |
| Reorderable list | Component | Ship | `DesignSystem/Components/ReorderableList/` | Saved rows only. |
| IconButton | Component | Ship | `DesignSystem/Components/IconButton/` | Delete on saved rows; dismiss stays on suggestions. |
| Confirm | Component | Ship | `DesignSystem/Components/Confirm/` | Only if delete confirms. `item:` snapshot. |
| DateValue editor form | Snowflake | **Remove** from this page | `Features/Dates/DateValueEditorForm.swift` | Keep for the citation composer. Do not instance here. |
| Input / Field / ComboBox / FormDialog | Component | Ship | `DesignSystem/Components/` | Add dialog + suggestion input as shipped. |
| Source page metadata tests | Test | **Extend** | `ProvenenciaTests/SourcePageModelTests.swift` | Date-named field saves as text; delete clears; no date payload. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A second Evidence graph jump, or moving the existing one | Already on the board; S8-07 implements those frames. |
| Reinvented metadata layout / date-entry chrome | Match the shipped app; delete is the only add. |
| Graph-specific counts on the page | List counts are **S8-D5**; not this page. |
| `mentions` / `remark` commentary | Parked: [`source-to-source-relationships.md`](../../../ideas/source-to-source-relationships.md). |
| DateValue (or NameValue) on metadata | Catalog filing is text. |
| First-class provenance-date control | Parked: [`source-provenance-date.md`](../../../ideas/source-provenance-date.md). |

---

## 7. Out of scope

- Restyling filing (notes, artifacts ingest) — metadata **returns to the shipped section** + delete
- Redesigning Jump to Evidence graph (already on the board)
- Keeping the board’s DateValue / easier-date-entry metadata experiments
- Composer / graph DateValue chrome
- Source-page Find
- Commentary (“what other Sources say about this one”) — [`source-to-source-relationships.md`](../../../ideas/source-to-source-relationships.md)
- Sorting or grouping Sources by catalog time — [`source-provenance-date.md`](../../../ideas/source-provenance-date.md)

---

## 8. Handoff

1. Keep this brief open until the **S8-07** bundle is frozen, then archive it.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S8-07** against the board.
