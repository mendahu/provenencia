# S7-D8 — PVCallout actions slot

**Kind:** Claude Design board / design-system component (light handoff OK)  
**Spike:** Provenencia Spike 7  
**Implements later as:** PR **S7-14**  
**Depends on:** Shipped macOS [`PVCallout`](../../../../macos/App/DesignSystem/Components/Callout/PVCallout.swift)  
**Unblocks:** Evidence graph No-Artifact **message center** (**S7-D3** / **S7-09**) — recovery CTA without a hand-rolled banner  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md) — **Extend** existing component; do not invent a parallel banner

Paste this document into Claude Design as the requirements for a **Callout** kit update (or component page delta). Read shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Extend **Callout** so it can carry **optional actions** (primary recovery / secondary dismiss patterns) while keeping today’s tone-tinted fill, left rule, icon, title, and body.

macOS already ships `PVCallout` without an actions slot — the Swift header notes that the web kit’s `actions` (and related) props were deferred. This board locks the visual contract so **S7-14** can port the slot without inventing a graph-only banner.

**Primary consumer:** Evidence graph No-Artifact message center (one corner callout + “Add an Artifact” → Source page). Other call sites may stay text-only.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Callout is content-agnostic kit | Actions are a **slot** (`PVButton` / `PVIconButton` at the call site) — not hard-coded graph copy inside `DesignSystem/`. |
| Web kit already has `actions` | Prefer matching Claude Design Callout.jsx; only document deltas. |
| No-Artifact is graph-wide | One callout host; do not design per-button callouts here (that’s **S7-D3**). |
| macOS dialogs | No scrim on the callout itself; actions are inline, not a nested sheet. |

---

## 3. Implementation gate (S7-14)

| Ships in S7-14 | Does **not** ship there |
| --- | --- |
| `PVCallout` optional actions slot (+ preview); DesignSystem README note | Wiring the Evidence graph No-Artifact gate / disable-controls (**S7-09**) |
| Optional: one existing text-only callout left unchanged (compat) | New Callout variants (`plain`, dismiss-X) unless the board proves need |
| Align Swift with board-clarified layout (trailing vs below body) | A second “banner” or “message center” component |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CA-1 | Board shows Callout **without** actions (today) and **with** one / two actions. |
| CA-2 | Compact + default density both support actions without clipping. |
| CA-3 | Action slot is call-site-owned (`PVButton` / links); kit does not invent domain CTAs. |
| CA-4 | Tone / icon / title / message behavior unchanged for existing call sites. |
| CA-5 | UI inventory: Callout = Component **Extend**; no new recipe; graph host remains a snowflake under **S7-D3**. |

---

## 5. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Callout | Component | **Extend** | `DesignSystem/Components/Callout/PVCallout.swift` | Add optional `@ViewBuilder` actions (or equivalent). Follow `PVButton` / `PVEmptyState` action-slot patterns. |
| EmptyState | Component | Ship | `DesignSystem/Components/EmptyState/PVEmptyState.swift` | Cousin with action slot already — **do not** reuse EmptyState for the graph corner message center. |
| Button / IconButton | Component | Ship | `Components/Button/`, `Components/IconButton/` | Fill the Callout actions slot at call sites. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| New `MessageCenter` / `GraphBanner` component | Compose `PVCallout` in `EvidenceGraphView`. |
| Dismiss / `onDismiss` / `detail` unless a real call site needs them | Stay minimal; ship actions first. |
| Evidence graph gate wiring | **S7-09** after this lands. |

---

## 6. Out of scope

- No-Artifact copy, placement, or disabled palette (**S7-D3** / **S7-09**)
- Toast as a substitute for this slot
- Redesigning Callout tones or left-rule chrome

---

## 7. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S7-14** against the board + existing `PVCallout` API.
4. Then **S7-09** may host the No-Artifact message center with a recovery action.
