# S8-D3 — Evidence graph visual enhancements

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-06** (one graph-chrome pass; more items join this brief)  
**Depends on:** Shipped Evidence graph ([`EvidenceGraphView`](../../../../macos/App/Features/EvidenceGraph/EvidenceGraphView.swift), [`EvidenceSubjectCard`](../../../../macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift), [`EvidenceBridgeCard`](../../../../macos/App/Features/EvidenceGraph/EvidenceBridgeCard.swift), [`EvidenceBridgeEdgeSummary`](../../../../macos/App/Features/EvidenceGraph/EvidenceBridgeEdgeSummary.swift)); snapshot already lists every Observation ([`SourceGraphSnapshot`](../../../../macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift)); Source page vs graph already distinct (`sourceSurface`)  
**Related:** leftover conflicted / negated honesty from [`interpretation-graph-ui.md`](../../../ideas/interpretation-graph-ui.md) §3.3 — **row badges only**. Incomplete-bridge chrome and collapse/expand are **descoped**.  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

A **visual-enhancements pass** on the Evidence graph. This board and **S8-06** are a **bundle**. Do **not** open a second graph-chrome PR unless a later item cannot share the same pass.

**Items**

1. **Conflict badge.** Competing Observations on one Property are expected, not an error. The card already shows them as **two rows**. When the same Property appears more than once, mark **each** of those rows with a small conflict badge. Example: two `name` rows → both badged; a single `name` does not.
2. **Negated badge.** `polarity = negative` is a denial, not a missing line. Today the cited row only *italicizes in danger color*. Make negation as explicit as conflict (composer already has a “Negates” chip).
3. **Jump to Source page.** The graph is Source-scoped but has no always-on path to that Source’s filing page. The header already shows the Source title (muted, not a control). The no-Artifact callout already jumps via `sourcePageLocation`. Add a **quick link** that is available even when the graph is usable.
4. **Richer bridge sentences.** Cited bridge cards use a composite sentence from endpoint **working labels**. The three connect pairs are fixed, and each primary has a seeded identity Property. Prefer that vocabulary, then fall back to `subjects.label`:
   - person ↔ person: **name** as {relationship_type} of **name**
   - person → event: **name** participated as {role} at **event_type**
   - event → place: **event_type** took place in **toponym**
5. **Add property on bridges.** Bridge subjects can take Observations (seeded edges plus any extra Subject-field bindings). Primary cards have **Add property**; cited bridges only have edit-citation on the first Citation. There is no control to add another Property. `composerLocation(for:)` already accepts a bridge id — the card just never calls it.

Do **not** redesign the canvas, palette, connect gesture, or composer form. Do **not** invent a “denied edge” drawing. **Collapse/expand** of bridge cards is descoped.

```text
Evidence graph     1851 England Census          [ Open Source ]     12 subjects

  William Robins
  ┌─────────────────────────────────────────────┐
  │  NAME                              [!]      │
  │  Wm Robins                                  │
  ├─────────────────────────────────────────────┤
  │  NAME                              [!]      │
  │  William Robins                             │
  ├─────────────────────────────────────────────┤
  │  BIRTH DATE                        [¬]      │
  │  1 JAN 1800                                 │
  └─────────────────────────────────────────────┘

  (relationship)
  Wm Robins is the father of John Robins
  ┌─────────────────────────────────────────────┐
  │  NOTE                                       │
  │  Named in the household block               │
  │  + Add property                             │
  └─────────────────────────────────────────────┘
```

Exact badge, link chrome, sentence wrapping, and where Add property sits on the smaller bridge card are board findings. Prefer `PVBadge` and reuse the primary **Add property** pattern.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| One Citation → many Observations; one Property may be asserted twice | Two rows, same `propertyKey`. That *is* the conflict. |
| Competing values are legal | Badge is a **notice**, not a resolve action. |
| Snapshot already lists all Observations | Count keys on the card; resolve identity Properties from **endpoint** primaries, not only the bridge row. |
| Negative polarity already italicizes in danger | Not enough. Board designs an explicit mark. |
| `WorkspaceLocation` already distinguishes `.graph` vs `.page` | Jump is `go(to:)` with the same `sourceId` and `sourceSurface: .page`. Back returns to the graph. |
| No-Artifact callout already opens the Source page | Do not invent a second navigation helper. Reuse `sourcePageLocation`. |
| Bridge sentence today uses edge-observation display (working labels) | Look up the referenced person/event/place on the snapshot and prefer `name` / `event_type` / `toponym`. |
| `name` / `event_type` / `toponym` are seeded identity Properties | Not every binding is registry-locked, but they are the first-class identity fields for those types. |
| Two names on one person | Sentence still picks **one** (first positive, else first). Conflict badges stay on the person card. |
| Connect-time crumb runs before names exist | Composer **prefill** title may still use working labels. Cited card (and post-save crumb if it already calls `sentence`) use identity Properties. |
| Incomplete “person but no event” bridges | **Descoped.** Connect is atomic; the UI cannot write that shape. No half-line chrome. |
| Collapse / expand bridge cards | **Descoped.** Sentence + extra rows stay visible. |
| Bridges have no Add property | Composer already opens for a bridge `subjectId`. Wire the same action as primaries. |
| Extra Observations on a bridge are invisible today | Show **non-edge** rows under the sentence (not `person` / `event` / `place` / `related_to` / `role` / `relationship_type` — those are the sentence). Conflict / negated apply to those rows. |

### 2.1 What this board is not

- Not incomplete-bridge / denied-line language, or collapse/expand (descoped).
- Not Source-page → graph (that leftover stays out).
- Not Citation pinning.
- Not density filters, undo, unplaced tray, or minimap (**descoped**).
- Not Conclusion Reconciliation Claims.
- Not PDF Find / OCR (**S8-D1** / **S8-D2**).

---

## 3. Implementation gate (S8-06)

| Ships in **S8-06** | Does **not** ship there |
| --- | --- |
| Conflict badge on every row whose `propertyKey` appears more than once | Merge / pick-a-winner UI |
| Negated badge on every `polarity = negative` row | Denied connect-line; flipping polarity |
| Always-on jump to the same Source’s detail page | Source-page control that opens the graph |
| Bridge sentence prefers endpoint `name` / `event_type` / `toponym`, then label | New L10n sentence *shapes* unless the board finds the shipped templates insufficient |
| Add property on bridge cards; extra (non-edge) Observation rows | Collapse/expand; restyling the composer |
| L10n + VoiceOver for new chrome | Schema / FFI; incomplete-bridge states |

If more bundle items land after the board is first drawn, **amend this brief** and redraw those frames — still one **S8-06**.

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| GV-1 | **Conflict:** if `propertyKey` occurs ≥ 2 times on a cited card, **each** matching row shows a small conflict badge. Singletons stay unmarked. |
| GV-2 | Key is **`propertyKey`**, not the localized label. Values need not differ. |
| GV-3 | Badges are visual + accessible only. No click-to-resolve. Row still opens Edit. |
| GV-4 | Prefer `PVBadge`. No new kit primitive unless the row cannot host it. |
| GV-5 | VoiceOver: row still reads property + value; each mark has a short name (e.g. “Conflict”, “Negated”). |
| GV-6 | Uncited cards have no property stack — no badge work there. |
| GV-7 | **Negated:** every cited row with `polarity = negative` shows an explicit negated mark. |
| GV-8 | Conflict and negated may **stack**. |
| GV-9 | Distinct tones so conflict ≠ negated. |
| GV-11 | **Source jump:** Evidence graph chrome includes a control that opens the **same** Source’s detail page (`sourceSurface: .page`). Available when the graph is usable, not only in the no-Artifact callout. |
| GV-12 | Jump uses existing `sourcePageLocation` + `navigation.go(to:)`. Toolbar Back returns to the graph. L10n + VoiceOver. |
| GV-13 | Board picks placement (header title as link vs trailing “Open Source”). Do not add a second breadcrumb trail. |
| GV-14 | **Bridge sentence — relationship:** prefer each endpoint’s `name` (NameValue form), then that subject’s working label. Phrase still uses `relationship_type`. |
| GV-15 | **Bridge sentence — participation:** prefer person’s `name` then label; prefer event’s `event_type` term then label. Phrase still uses `role`. |
| GV-16 | **Bridge sentence — location:** prefer event’s `event_type` then label; prefer place’s `toponym` then label. |
| GV-17 | Identity pick: first **positive** Observation of that key on the endpoint, else first, else label, else today’s bare phrase. Do not concatenate competing names. |
| GV-18 | Cited card height / a11y use the new sentence. Keep `EvidenceBridgeEdgeSummary` as the single helper (card + any crumb that already calls `sentence`). |
| GV-19 | **Add property on bridges:** cited (and uncited-but-`canCite`) bridge cards show **Add property**. Same L10n as primaries. Opens `composerLocation(for: bridgeID)` (new Citation, not the connect Citation). Disabled when `!canCite`. |
| GV-20 | Extra Observations on the bridge (not the edge keys used in the sentence) paint as cited rows under the sentence so they can be read, edited, and take conflict/negated badges. Edge keys stay in the sentence only — do not duplicate them as rows. |
| GV-21 | Hit targets + VoiceOver for Add property / extra-row edit; `contentHeight` grows with extra rows. |
| GV-10 | Further items get their own `GV-n` rows when scoped. |

---

## 5. Suggested frames

1. Cited person — two `name` rows, both conflict-badged; one `occupation`, unmarked.
2. Two `birth_date` values that disagree.
3. Two `name` rows with the **same** string (still conflict).
4. Negative singleton — negated mark, no conflict.
5. Negative + positive same Property — both conflict; only the negative row also negated.
6. Graph header with Source jump (has Artifacts).
7. Graph header + no-Artifact callout — jump still present; callout action still works.
8. Relationship card: both people have `name` → “Wm Robins is the father of John Robins” (not working labels).
9. Relationship card: one person has only a working label → that side falls back.
10. Participation card: name + `event_type` (e.g. Census Enumeration).
11. Location card: `event_type` + `toponym`.
12. Cited relationship with **Add property** and one extra (non-edge) row.
13. Bridge, no Artifact — Add property disabled (same as primaries).
14. *(Add frames here as more bundle items are scoped.)*

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Evidence graph view / header | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceGraphView.swift` | Host Source jump; wire bridge Add property a11y like primaries. |
| Evidence graph model | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceGraphModel.swift` | Reuse `sourcePageLocation` and `composerLocation(for:)`. |
| Evidence subject card | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceSubjectCard.swift` | Conflict / negated on rows. |
| Cited property row | Snowflake | **Extend** | same file (`EvidenceCitedPropertyRow`) | Host marks; keep hover / edit hit. |
| Bridge card | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceBridgeCard.swift` | Sentence + extra rows + Add property; height/hits. |
| Bridge edge summary | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceBridgeEdgeSummary.swift` | Resolve identity from snapshot primaries. |
| Source graph snapshot | Session | Ship | `Features/Workspace/Session/SourceGraphSnapshot.swift` | Already has endpoint observations. |
| Badge | Component | Ship | `DesignSystem/Components/Badge/PVBadge.swift` | Conflict + negated. |
| Button | Component | Ship | `Components/Button/` | Source jump if not a text link. |
| Card height / summary tests | Test | **Extend** | `ProvenenciaTests` | Badges, sentence fallbacks, jump location. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| `PVConflictBadge` / `PVNegatedBadge` | Compose `PVBadge`. |
| Incomplete / half-line chrome | UI cannot produce that graph. |
| Collapse / expand | Descoped. |
| Filters / undo / tray / minimap | Descoped. UI cannot write unplaced subjects; density lives in dogfood. |
| Source-page → graph button | Different leftover. |
| New sentence catalog keys unless templates cannot take the new nouns | Reuse shipped `bridgeSummary*` strings. |

---

## 7. Out of scope

- Resolving or hiding a competing Observation
- Schema / FFI
- Composer Observation-list restyle
- Incomplete-bridge visual states (**descoped**)
- Collapse / expand bridge cards (**descoped**)
- Density filters, undo, unplaced tray, minimap (**descoped**)
- Pinning
- Source-page Evidence graph entry

---

## 8. Handoff

1. Keep this brief open until the **S8-06** bundle is frozen, then archive it.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S8-06** against the board.
