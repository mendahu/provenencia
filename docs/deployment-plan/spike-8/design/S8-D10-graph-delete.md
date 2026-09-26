# S8-D10 — Evidence graph delete chrome

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-19** (graph trash + cut over `subjects.Delete`)  
**Depends on:** Frozen policy, path matrix, and table register in [`../deployment-plan.md`](../deployment-plan.md) §S8-09.1–§S8-09.4; **S8-09**, **S8-12**, and the **S8-D9** recipe (**S8-13**) landed or mocked as the same contract; shipped graph cards ([`EvidenceSubjectCard`](../../../../macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift), [`EvidenceBridgeCard`](../../../../macos/App/Features/EvidenceGraph/EvidenceBridgeCard.swift), [`EvidenceGraphModel.beginDelete`](../../../../macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift))  
**Related:** Shared confirm / blocked notice is [`S8-D9-impact.md`](S8-D9-impact.md) — **instance it, do not restyle.** Composer delete chrome is a **different board** ([`S8-D11-composer-delete.md`](S8-D11-composer-delete.md) → **S8-18**). Do not draw the composer. Wrong type is delete + place (leftover **19**). Leftover **20** (adopt) is descoped.  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of shipped graph delete chrome (uncited trash + confirm). Do not fork a second canvas.

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
| Resource delete with inbound check | DeleteImpact recipe — confirm if allowed, notice if blocked |
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

Make graph delete **honest**. Policy is already frozen in the plan: erase only when nothing points at the subject; otherwise refuse with the Impact list (refs + links). This board draws that on the **cards**, not a new matrix.

**Today:** uncited primary (and theoretically uncited bridge) show trash and confirm. Cited cards hide trash. A person who is only a bridge endpoint looks uncited (`isCited` = “has my own Observations”) but the engine returns `subjects.in_use`. That trash is a lie.

**This board designs**

1. Trash on **every** primary and bridge card (cited included). Do not hide it.
2. **Erase confirm** when inbound is empty — the **allowed** branch of DeleteImpact (today’s uncited primary — keep it, do not make it worse).
3. **Blocked notice** when inbound is not empty (cited primary, endpoint-only person, bridge with extra Observations). Instance DeleteImpact; do not invent a graph-only refuse. A cited bridge that only has connection facets uses the **confirm** branch (those rows go with it — quiet line, not a cascade checklist).
4. VoiceOver names the subject and whether this will erase or is blocked.

Do **not** redesign cards except delete chrome. Do **not** add per-row Observation trash on the card (that is the composer). Do **not** invent Change type.

---

## 2. Domain facts (this board)

| Fact | UI implication |
| --- | --- |
| Policy lives in the plan §S8-09.1 | Boards do not invent cascade. Erase or refuse. |
| Path matrix is §S8-09.2 **G1–G6** | G1 erase; G2–G4 refuse; G5 erase if it exists; G6 erase when only connection facets remain, refuse when extras remain. |
| `isCited` is “has Observations as `subject_id`” | Endpoint-only (G2) looks empty and still must refuse. Do not use the dashed shell as “safe to delete.” |
| Inbound for a subject is ordinary `subject_id` rows **or** `value_subject_id` | Connection facets (edges + role / relationship_type on the bridge) are **not** inbound. A bridge edge pointing at John still blocks **John**. |
| Connect is atomic | A cited bridge with only its connection facets **can** be erased; official `subjects.Delete` releases those rows. Extra Add-property rows on the bridge still block. |
| Endpoints stay when the bridge erases | Do not draw “also delete William.” |
| Positions are facets | They CASCADE with the subject. Do not mention layout in the confirm. |
| Wrong type | Delete this card (if allowed) and place a new one. No type control. |
| Observation rows on the card | Click opens the composer. No card-level Observation delete. |

### 2.1 What this board is not

- Not the citation composer (**S8-D11**).
- Not restyling DeleteImpact (**S8-D9**).
- Not Source / Artifact delete UI.
- Not undo / ⌘Z.
- Not Change type, adopt/import, graph badges (**S8-D3** already shipped).
- Not a counted cascade that also deletes Observations.

---

## 3. Implementation gate (S8-19)

| Ships in **S8-19** | Does **not** ship there |
| --- | --- |
| Trash on cited and uncited cards; DeleteImpact confirm vs notice; `subjects.Delete` through the register | Other screens; the recipe itself (**S8-13**) |
| L10n + VoiceOver for graph trash | Schema repair (**S8-09**); registry (**S8-12**); recipe chrome (**S8-13**) |
| G2 no longer looks like a working trash | Per-row Observation delete on the card |
| | `UPDATE subject_type_id`; Source-layer delete UI |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| GD-1 | Every primary and bridge card shows a delete control. Cited is not hidden. |
| GD-2 | Inbound empty → DeleteImpact **confirm** branch. Copy names the subject (`label` or sentence + ref). |
| GD-3 | Inbound non-empty → DeleteImpact **notice**, not confirm-then-engine-fail. `via` splits “cites this card” from “used as an endpoint.” Short path: strip those cites, or leave it. |
| GD-4 | G2 (endpoint-only) uses the refuse, even though the card is dashed / uncited-looking. |
| GD-5 | Uncited primary erase stays (G1). Do not regress it. |
| GD-6 | Card Observation rows stay “open composer.” No trash on a property row. |
| GD-7 | VoiceOver: destructive control + dialog say erase vs blocked, and what is named. |
| GD-8 | Header trash stays. Present DeleteImpact; do not fork a graph-only sheet. |

---

## 5. Suggested frames

1. Uncited primary — today’s confirm (baseline).
2. Endpoint-only primary (dashed, connected by a cited bridge) — refuse, not confirm.
3. Cited primary — trash visible; refuse with named Observation refs (cites vs endpoint groups).
4. Cited bridge, edges + role only — confirm (connection facets go with it; endpoints stay).
4b. Cited bridge with an extra Add-property row — notice names that `OBS-…`.
5. *(Optional)* Disabled-looking trash vs always-enabled + refuse dialog — board picks one pattern and uses it on every blocked card.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Subject / bridge header trash | Snowflake | **Extend** | `EvidenceSubjectCard` / `EvidenceBridgeCard` | Show even when `isCited`. Same hit target as today. |
| Graph delete present | Snowflake | **Extend** | `EvidenceGraphView` + `EvidenceGraphModel` | Calls DeleteImpact with `GetDeleteImpact`. |
| DeleteImpact | Recipe | Ship | `DesignSystem/Recipes/DeleteImpact/` | **S8-13**. Confirm or notice; do not restyle. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Local graph refuse sheet | Shared recipe is **S8-D9**. |
| Multi-select delete | Scope. |
| Cascade copy that lists Observations to destroy | Policy is refuse, not cascade. |

---

## 7. Out of scope

- Composer chrome
- Source / Artifact delete
- ⌘Z
- Change type UI
- Redesigning card body / badges

---

## 8. Handoff

1. Archive this brief under [`archive/`](archive/) and write [`../completed.md`](../completed.md).
2. Implement **S8-19** against these frames + §S8-09.2 G-rows. Cut over `subjects.Delete` in this PR. Reuse `Impact` from **S8-12**.
