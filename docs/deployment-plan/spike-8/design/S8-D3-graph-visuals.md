# S8-D3 — Evidence graph visual enhancements

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-06** (one card-chrome pass; more items join this brief)  
**Depends on:** Shipped Evidence graph cards ([`EvidenceSubjectCard`](../../../../macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift), [`EvidenceCitedPropertyRow`](../../../../macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift)); snapshot already lists **every** Observation ([`SourceGraphSnapshot`](../../../../macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift))  
**Related:** leftover conflicted / negated honesty from [`interpretation-graph-ui.md`](../../../ideas/interpretation-graph-ui.md) §3.3 — **row badges only**, not that slice  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

A **small visual-enhancements pass** on Evidence graph cards. This board and **S8-06** are a **bundle**: more items will be added here as they are scoped. Do **not** open a second graph-chrome PR unless a later item cannot share the same card pass.

**Items so far**

1. **Conflict badge.** Competing Observations on one Property are expected, not an error. The card already shows them as **two rows**. When the same Property appears more than once, mark **each** of those rows with a small conflict badge so the researcher notices. Example: two `name` rows → both badged; a single `name` does not.
2. **Negated badge.** An Observation with `polarity = negative` is a denial, not the absence of a line. Today the cited row only *italicizes in danger color*. Make negation as explicit as conflict (composer already has a “Negates” chip). A negative singleton must still read as negated even when there is no conflict.

Do **not** redesign the canvas, palette, connect, or composer. Do **not** invent a “denied edge” drawing.

```text
  William Robins
  ┌─────────────────────────────────────────────┐
  │  NAME                              [!]      │
  │  Wm Robins                                  │
  ├─────────────────────────────────────────────┤
  │  NAME                              [!]      │
  │  William Robins                             │
  ├─────────────────────────────────────────────┤
  │  BIRTH DATE                        [¬]      │
  │  1 JAN 1800   (negated)                     │
  ├─────────────────────────────────────────────┤
  │  OCCUPATION                                 │
  │  carpenter                                  │
  └─────────────────────────────────────────────┘
```

Exact badge glyph, copy, tone, and placement (and whether italic/danger stays) are board findings. Prefer shipped `PVBadge`. Conflict and negated may stack on one row.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| One Citation → many Observations; one Property may be asserted twice | Two rows, same `propertyKey`. That *is* the conflict. |
| Competing values are legal | Badge is a **notice**, not a warning that data is wrong. No resolve/merge action. |
| Snapshot already lists all Observations | No new query. Count `propertyKey` on the card. |
| Primary cards paint one row per Observation | Both badges live on `EvidenceCitedPropertyRow`. |
| Bridge cards summarize edges (often `.first` per key) | Conflict if a listed Property repeats; negated if that Observation is negative. Do not invent a denied-line drawing. |
| Negative polarity already italicizes in danger | Not enough. Board designs an explicit mark; may keep or drop italic. |
| Composer already shows “Negates” | Match that meaning on the card; do not restyle the composer in this PR. |

### 2.1 What this board is not

- Not a full honesty language (incomplete lines, collapse/expand bridges, filters, canvas-level denied edges).
- Not Citation pinning, tray, or composer work.
- Not Conclusion Reconciliation Claims.
- Not PDF Find / OCR (**S8-D1** / **S8-D2**).

---

## 3. Implementation gate (S8-06)

| Ships in **S8-06** | Does **not** ship there |
| --- | --- |
| Conflict badge on every row whose `propertyKey` appears more than once on that card | Merge / pick-a-winner UI |
| Negated badge (or equivalent mark) on every `polarity = negative` row | Drawing a missing/denied connect line; flipping polarity |
| L10n + VoiceOver for both marks | Schema flag; Go changes |
| Further visual items **added to this brief** before the PR starts | A second graph-visuals PR |

If more bundle items land after the board is first drawn, **amend this brief** and redraw those frames — still one **S8-06**.

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| GV-1 | **Conflict:** on a subject card, if `propertyKey` occurs ≥ 2 times, **each** matching cited row shows a small conflict badge. Singletons stay unmarked. |
| GV-2 | Key is **`propertyKey`**, not the localized label (labels can collide). Values need not differ — two identical names still badge. |
| GV-3 | Badge is visual + accessible only. No click-to-resolve. Row still opens Edit as today. |
| GV-4 | Prefer `PVBadge` (tone a board finding: warning vs danger vs subtle). No new kit primitive unless the row cannot host `PVBadge`. |
| GV-5 | VoiceOver: row still reads property + value; each mark has a short name (e.g. “Conflict”, “Negated”). |
| GV-6 | Uncited cards have no property stack — no badge work there. |
| GV-7 | **Negated:** every cited row with `polarity = negative` shows an explicit negated mark. Positive / default rows do not. |
| GV-8 | Conflict and negated are independent and may **stack** (two `birth_date` rows, one negative → both conflict; the negative row also shows negated). |
| GV-9 | Prefer `PVBadge` for both marks. Distinct tones so conflict ≠ negated. |
| GV-10 | Inventory each bundle item. Further items get their own `GV-n` rows when scoped. |

---

## 5. Suggested frames

1. Cited person — two `name` rows, both badged; one `occupation`, unmarked.
2. Same card — two `birth_date` values that disagree.
3. Same card — two `name` rows with the **same** string (still badge).
4. Negative singleton — negated mark, **no** conflict badge.
5. Negative + positive same Property — both conflict; only the negative row also shows negated.
6. Two negative rows on the same Property — both conflict + both negated.
7. Bridge card if a listed Property repeats and/or an edge Observation is negative (or a note that bridges usually do not).
8. *(Add frames here as more bundle items are scoped.)*

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Evidence subject card | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceSubjectCard.swift` | Pass which keys are duplicated; do not restyle the shell. |
| Cited property row | Snowflake | **Extend** | same file (`EvidenceCitedPropertyRow`) | Host conflict + negated marks; keep hover / edit hit. |
| Source graph snapshot | Session | Ship / tiny helper | `Features/Workspace/Session/SourceGraphSnapshot.swift` | Already has all Observations; count keys in the view or a pure helper. |
| Bridge card | Snowflake | **Extend** only if rows list a repeated key | `Features/EvidenceGraph/EvidenceBridgeCard.swift` | Do not change edge-sentence `.first` in this item unless a later GV says so. |
| Badge | Component | Ship | `DesignSystem/Components/Badge/PVBadge.swift` | Prefer existing tones. |
| Card height / a11y tests | Test | **Extend** | `ProvenenciaTests` (card height + snapshot helpers) | Two name rows + badge must not clip. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| `PVConflictBadge` / `PVNegatedBadge` kit types | One call site; compose `PVBadge`. |
| Conflict on the canvas chrome / header | Rows are enough. |
| Auto-layout / tray / minimap | Different leftovers. |

---

## 7. Out of scope

- Resolving or hiding a competing Observation
- Schema / FFI
- Composer Observation list chrome
- Full slice-8 honesty (incomplete bridges, filters, undo, canvas denied-edges)

---

## 8. Handoff

1. Keep this brief open until the **S8-06** bundle is frozen, then archive it.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S8-06** against the board (conflict + negated + any other GV items added here).
