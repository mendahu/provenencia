# S8-D12 — Source page resource delete

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-14** (Source + Artifact official Delete + DeleteImpact on this page)  
**Depends on:** Frozen policy + Impact report in [`../deployment-plan.md`](../deployment-plan.md) §S8-09.1–§S8-09.4; **S8-12** `GetDeleteImpact`; **S8-D9** / **S8-13** DeleteImpact recipe  
**Related:** Shared confirm / notice is [`S8-D9-impact.md`](S8-D9-impact.md) — **instance it, do not restyle.** Metadata-value delete already ships (`ClearSourceMetadata`) and stays on `.pvConfirm` (facet). Vocab / composer / graph are other boards.  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of the shipped Source page (**S8-D4** / **S8-07**). Do not start a second page.

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

---

## 1. Objective

Give the Source page official **resource** delete for the Source and for each Artifact. Both go through `GetDeleteImpact` + DeleteImpact.

**This board designs**

1. **Delete Source** — confirm when inbound resources are empty (no Artifact, no Subject on this Source). Notes / metadata / credibility / layout CASCADE; do not ask the researcher to clear those first. Blocked notice lists Artifacts (`ART-…`) and/or Subjects (`CPR-…`) with links.
2. **Delete Artifact** — confirm when no Citation points at it. File releases `ifUnused` (pool). Primary thumb `SET NULL` if this was the cover. Blocked notice lists Citations (`CIT-…`).
3. Placement of those controls on the **shipped** page (identity / artifact list). Do not hide trash when inbound is non-empty.

**Shipped metadata-row delete stays.** That is a facet write, not this recipe.

---

## 2. Domain facts (this board)

| Fact | UI implication |
| --- | --- |
| Policy §S8-09.1 / report §S8-09.4 | Confirm iff Impact.allowed. Notice names blockers. |
| `artifacts.source_id` and `subjects.source_id` block the Source | A Source that still has an Artifact or a graph card cannot go. |
| `citations.artifact_id` blocks the Artifact | Empty Artifact (no Citation) can go. |
| Facets CASCADE | Notes / metadata / layout do not appear in the notice. |
| `files` are a pool | Deleting the last Artifact that uses a checksum may remove the file; another Artifact keeps it. Do not explain checksums unless the board needs one quiet line. |
| After Source erase | Leave the page — Sources list (or history Back). Do not keep a stale `sourceId`. |
| After Artifact erase | Stay on the Source page. If the composer had that Artifact open, **S8-D11** / **S8-18** must recover (`missingDeepId` / identity fallback). Do not leave that unspecified. |
| Sources list | No list-row delete on this board unless the frames find a natural one that uses the same recipe. Prefer the page. |

### 2.1 What this board is not

- Not restyling DeleteImpact (**S8-D9**).
- Not vocab, composer, or graph delete.
- Not metadata / note Confirm (already shipped).
- Not redesigning identity, credibility, or the artifact viewer.

---

## 3. Implementation gate (S8-14)

| Ships in **S8-14** | Does **not** ship there |
| --- | --- |
| `sources.Delete` + `artifacts.Delete` through `deleteimpact` (new writers) | Vocab / composer / graph chrome |
| Source page trash + DeleteImpact | The recipe itself (**S8-13**) |
| File `ifUnused` release on Artifact erase | Changing FK actions (**S8-09**) |
| L10n + VoiceOver + leave-page after Source erase | Metadata-row Confirm rewrite |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SD-1 | Source offers a delete control. DeleteImpact confirm when allowed; notice when Artifacts or Subjects remain. |
| SD-2 | Each Artifact offers a delete control. Confirm when no Citations; notice lists `CIT-…`. |
| SD-3 | Facets are not listed and are not a pre-step. |
| SD-4 | Successful Source erase does not leave the app on a ghost Source page. |
| SD-5 | Successful Artifact erase keeps the Source page; cover thumb clears if it was primary. |
| SD-6 | VoiceOver names the Source or Artifact and erase vs blocked. |
| SD-7 | Instance DeleteImpact. No page-local refuse sheet. |

---

## 5. Suggested frames

1. Empty-ish Source (notes/metadata only, no Artifact, no Subject) — Delete Source confirm.
2. Source with two Artifacts — Delete Source notice names those `ART-…`.
3. Source with an uncited Subject — Delete Source notice names the subject.
4. Artifact with zero Citations — Delete Artifact confirm.
5. Artifact with Citations — notice names `CIT-…` (link to composer).
6. After Source erase — Sources list / Back, no crash.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Source identity / page chrome | Snowflake | **Extend** | `SourcePageView` / identity header | Home for Delete Source. |
| Artifact row / list | Snowflake | **Extend** | shipped artifact chrome | Home for Delete Artifact. |
| DeleteImpact | Recipe | Ship | `DesignSystem/Recipes/DeleteImpact/` | **S8-13**. |
| Metadata-value Confirm | Component | Ship | `.pvConfirm` | Facet only. Do not move it onto DeleteImpact. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Cascade “delete Source and 4 Artifacts” | Policy is refuse. |
| List-only Source delete as a second pattern | Prefer one place (the page) unless the board proves a list control that uses the same recipe. |
| Rewriting metadata / notes Confirm | Not inbound-resource checks. |

---

## 7. Out of scope

- Graph / composer / vocab
- Redesigning the Source page
- ⌘Z

---

## 8. Handoff

1. Archive this brief under [`archive/`](archive/) and write [`../completed.md`](../completed.md).
2. Implement **S8-14** against these frames. New `Delete`s go through [`add-catalog-delete`](../../../../.cursor/skills/add-catalog-delete/SKILL.md).
