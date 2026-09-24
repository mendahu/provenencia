# S8-D6 — Interpretation delete paths

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-09** (one delete-semantics pass; more items join this brief)  
**Depends on:** Shipped uncited-subject delete ([`EvidenceGraphModel.beginDelete`](../../../../macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift)); Go `subjects.Delete` (`ErrInUse` when any Observation references the subject as `subject_id` **or** `value_subject_id`); composer draft-row remove only  
**Related:** interpretation-graph-ui leftover **18** / §4.3 / hiccup 3. Leftover **19** is **folded here, descoped as a feature:** no Change type UI — delete and place again. Leftover **20** (adopt imports) is **descoped**.  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

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
| Composer `removeObservation` | Draft rows only. No persisted Observation/Citation delete FFI. |
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
| One Observation on a card | No UI | Leave the Citation if other Observations remain? |
| Last Observation on a Citation | No UI | Delete the empty Citation, or keep leftover **10** (empty Citation — dogfood)? |
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

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Graph delete confirm | Snowflake | **Extend** | `EvidenceGraphView` + `EvidenceGraphModel` | Today uncited only. |
| Subject / bridge trash | Snowflake | **Extend** | `EvidenceSubjectCard` / `EvidenceBridgeCard` | Cited cards have no trash. |
| Composer observation remove | Snowflake | **Extend** if persisted delete ships | `CitationComposerFormPane` | Draft-only today. |
| Confirm | Recipe | Ship | `.pvConfirm` / `PVConfirm` | Damage copy. |
| `subjects.Delete` | Engine | **Extend** | `core/database/subjects` | Cascade vs keep `ErrInUse` per matrix. |
| Citation / Observation delete | Engine | **Add** if matrix allows | `core/database/citations`, `observations` | No Citation `Delete` today. |
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
