# S8-D6 — Interpretation delete paths

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-09** (one delete-semantics pass; more items join this brief)  
**Depends on:** Shipped uncited-subject delete ([`EvidenceGraphModel.beginDelete`](../../../../macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift)); Go `subjects.Delete` (`ErrInUse` when any Observation references the subject as `subject_id` **or** `value_subject_id`); composer draft-row remove only  
**Related:** interpretation-graph-ui leftover **18** / §4.3 / hiccup 3. Leftover **19** is **folded here, descoped as a feature:** no Change type UI — delete and place again. Leftover **20** (adopt imports) is **descoped**.  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of shipped graph delete chrome (plus new confirms).

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

A **delete-paths** pass for the Interpretation layer. Researchers will add the wrong bubble, cite the wrong line, or need to undo a connect. That has to stay possible. What dies with a delete is not one row — Observations, shared Citations, and bridge edges fan out — so this board’s job is to **see every path** and design an honest confirm (or an honest refuse).

This board and **S8-09** are a **bundle**. The matrix will be **refined when the story starts** (what is allowed, what is refused, what cascades, what the confirm says). Do **not** open a second delete PR unless a later item cannot share the same pass.

**Items so far**

1. **Delete matrix.** Walk every Interpretation delete the product can offer. Seed from leftover **18** (cited subject: today trash is hidden and Go returns `subjects.in_use`). Broaden to the other entities and to “this Citation is shared.” For each path: allowed / refused / cascade-in-a-transaction, and what the confirm counts (“removes 7 Observations across 3 Citations”).
2. **Wrong type.** `subject_type_id` stays immutable. There is **no** Change type control. The researcher deletes the subject (matrix rules) and places a new one. That *is* leftover **19** — not a second feature.

Do **not** redesign the graph, composer, or cards except for delete chrome (trash, confirm, disabled + reason). Prefer shipped `.pvConfirm` / `PVConfirm`. Audit stays a **forward revision** ([`audit-revision-history.md`](../../../audit-revision-history.md) §9) — delete is not `UndoManager`.

```text
Delete William Robins (CPR-…)?

Removes 7 Observations across 3 Citations.
Also removes the “father of” bridge to John Robins.

[ Cancel ]     [ Delete ]
```

Exact copy, which paths get a confirm vs a refuse, and whether shared Citations are left empty or deleted are **board + later refinement**. Do not invent silent `ON DELETE CASCADE` for evidence.

---

## 2. Domain facts (today)

| Fact | UI implication |
| --- | --- |
| FKs are `NO ACTION` | Deleting a referenced subject/Citation **fails** unless the app deletes dependents first in one transaction. |
| A subject is referenced two ways | `observations.subject_id` **and** `observations.value_subject_id`. Damage count is two queries. |
| Uncited primary (and theoretically uncited bridge) | Trash + confirm shipped. `subjects.Delete` succeeds. Positions `ON DELETE CASCADE` (layout, not evidence). |
| Cited subject / cited bridge | No trash. Engine `ErrInUse`. Researcher who placed the wrong person and then cited them is stuck. |
| Connect is atomic | UI does not persist an uncited bridge. Deleting a bridge is deleting a cited subgraph. |
| One Citation → many Observations, possibly many subjects | Deleting one Observation must not silently delete a shared Citation. Deleting a Citation deletes every Observation on it. |
| Composer Observation delete | **S8-11** ships row Delete… with confirm (`DeleteObservation`, one audited revision; edge rows refused). No Citation delete FFI. |
| `citations` package has no `Delete` | Citation delete is greenfield on both sides. |
| Source with Subjects | Source delete stays blocked (`NO ACTION`). Not this story unless a later item adds Source-layer delete. |
| Wrong type | Immutable `subject_type_id`. **No Change type UI.** Delete + place a new subject. Refs are cheap to burn. |

### 2.1 Paths the matrix must name

Fill allow / refuse / cascade when this story is refined. Rows may be added.

| Path | Today | Must decide |
| --- | --- | --- |
| Uncited primary | Allowed + confirm | Keep; copy only? |
| Cited primary | Hidden + engine fail | Allow with damage count, or refuse with a path to strip citations first? |
| Cited bridge | Hidden + engine fail | Same. Endpoints stay? |
| One Observation on a card | **Decided by S8-D8 / S8-11:** composer row Delete… with confirm; the Citation always stays | Card-level delete affordance only (reuse `observations.Delete`) |
| Last Observation on a Citation | **Decided by S8-D8 / S8-11:** the empty Citation stays (empty Citations are legal) | Whether a later whole-Citation delete is offered from here |
| Whole Citation (shared) | No UI | Count Observations **and** other subjects that lose support. |
| Citation notes / observation notes | Follow parent? | Silent with parent vs listed in the confirm. |
| Subject that is only an edge object | Engine counts `value_subject_id` | Deleting John while he is `related_to` on a bridge. |

### 2.2 What this board is not

- Not undo / ⌘Z (descoped).
- Not composer pinning / empty-Citation policy (**S8-D7**) except where a delete would *create* an empty Citation.
- Not Source / Artifact / File delete (Source layer), unless a later item pulls a gate in.
- Not a Change type action (delete + recreate only).
- Not adopting imported / unplaced subjects (**20** — descoped with the tray).
- Not graph visual badges (**S8-D3**).

---

## 3. Implementation gate (S8-09)

| Ships in **S8-09** | Does **not** ship there |
| --- | --- |
| The refined matrix: each in-scope path is allowed-with-confirm or refused-with-reason | Silent SQL CASCADE on Observations / Citations |
| Damage-count confirm where a cascade is allowed (counts before delete) | ⌘Z / audit rollback |
| Engine deletes in **one transaction** when cascade is allowed; `ErrInUse` (or equivalent) when refused | A second delete PR for “just citations” |
| L10n + VoiceOver on trash / confirm / refuse | `UPDATE` of `subject_type_id`; a Change type control |
| Further items **added to this brief** before the PR starts | |

Freeze the matrix on this brief before **S8-09** starts. If refinement adds paths after the first board, amend and redraw — still one PR.

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| DL-1 | The board (then the brief) lists every in-scope delete path with **allow + confirm**, **refuse + reason**, or **out**. |
| DL-2 | Any allowed cascade **counts the damage** before the researcher confirms. Subject *and* object Observation refs. Shared Citations named when they would go or be emptied. |
| DL-3 | Allowed deletes are one engine transaction. Partial failure must not leave a half-deleted graph. |
| DL-4 | Refused deletes do not look like a broken trash can. Short reason; no navigation to a dead end. |
| DL-5 | Uncited primary delete stays possible (already shipped). Do not regress it. |
| DL-6 | Prefer `.pvConfirm` / `PVConfirm`. No new kit primitive. |
| DL-7 | VoiceOver: destructive control + confirm dialog name what will be removed. |
| DL-8 | Audit records the delete as a revision, same as other catalog writes. |
| DL-9 | Further items get their own `DL-n` rows when scoped. |

---

## 5. Suggested frames

1. Uncited primary — today’s confirm (baseline; do not make it worse).
2. Cited primary — proposed confirm **or** refuse (board picks after the matrix is sketched).
3. Cited bridge — same.
4. Shared Citation — “this reading also supports N other Observations / subjects.”
5. Single Observation remove when the Citation has more rows.
6. *(Add frames as the matrix is refined.)*

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Graph delete confirm | Snowflake | **Extend** | `EvidenceGraphView` + `EvidenceGraphModel` | Today uncited only. |
| Subject / bridge trash | Snowflake | **Extend** | `EvidenceSubjectCard` / `EvidenceBridgeCard` | Cited cards have no trash. |
| Composer observation remove | Snowflake | Ship (**S8-11**) | `CitationComposerObservationRow` | Row Delete… + confirm ships in S8-11. Edge rows are engine-locked (`observations.edge_locked`). Do not add a second Observation delete. |
| Confirm | Recipe | Ship | `.pvConfirm` / `PVConfirm` | Damage copy. |
| `subjects.Delete` | Engine | **Extend** | `core/database/subjects` | Cascade vs keep `ErrInUse` per matrix. |
| Citation / Observation delete | Engine | Citation: **Add** if matrix allows. Observation: Ship (**S8-11** `observations.Delete`) | `core/database/citations`, `observations` | No Citation `Delete` today. Reuse S8-11's lossless audit shape for any cascade. |
| Delete-impact query | Engine | **Add** | counts before confirm | Two Observation FKs + shared Citation. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Schema `ON DELETE CASCADE` on evidence FKs | Evidence must not vanish without a counted confirm. Layout already cascades. |
| Soft-delete / tombstone tables | Audit is the history. |
| Change type / in-place type UPDATE | Delete and place again. |
| Multi-select delete | Scope creep unless a later DL item adds it. |

---

## 7. Out of scope

- ⌘Z / undo stack
- Source-page commentary deletes
- Composer rethink (**S8-D7**)
- Redesigning cards except delete affordances
- Change type UI; adopt/import tray

---

## 8. Handoff

1. Keep this brief open. **Refine the matrix** before freezing **S8-09**.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S8-09** against the frozen matrix (Go transaction + confirms + tests per path).
