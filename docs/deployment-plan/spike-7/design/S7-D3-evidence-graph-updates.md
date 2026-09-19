# S7-D3 — Evidence graph updates (Add property, cited rows, connect handoff)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PRs **S7-09**, **S7-10** only  
**Depends on:** Spike 6 canvas + cards (S6-D1/D2); composer place designed in **S7-D4** (handoff target)  
**Related briefs:** [`S7-D4`](S7-D4-citation-composer.md) — composer place (do not design it here)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **Evidence graph chrome updates** now that Citations / Observations exist:

- **Add property** affordance on primary (and bridge) cards
- **Cited-data rows** that make cards grow vertically
- Uncited → cited shell transition once Observations exist
- **No-Artifact** empty / disabled state for citing
- **Connect** disambiguation sheet, then **navigate away** to the citation composer place (Option B) — this board designs the graph-side handoff only

**Do not design the composer layout, viewer, or form.** That is **S7-D4**.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Graph is the only Subject surface | No parallel list of Observations. Rows live on the card. |
| Uncited = zero Observations about the subject | Working `label` is not an Observation. Keep Spike 6 uncited treatment until first Observation. |
| Cards grow | Cited property rows stack; width stays fixed (~236pt family); height grows; edge attachment must still read correctly. |
| Add property needs an Artifact | Source with no Artifact: disable or explain; route toward adding a file on the Source page. |
| Composer is another place | Add property / connect completion **navigates** (leaves the canvas). Propose transition cue (brief), not a sheet over the graph. |
| Bridges require Citation at create | Connect never creates an uncited bridge. Disambiguation on-graph → composer with two edge Observations pre-scoped. |
| Provisional Spike 6 links | Replace honesty labels with real cited edges after S7-10; propose interim vs final bridge body. |

### 2.1 What this board is not

- Not the citation composer place (viewer\|form, breadcrumbs) — **S7-D4**.
- Not Subject types / fields admin — **S7-D1 / D2**.
- Not NameValue editor chrome (nested in composer) — **S7-D4**.
- Not conflicted / negated Observation visual language (thin OK; full polish later).
- Not unplaced-subjects tray / minimap.

### 2.2 Implementation gate

| Ships in S7-09 / S7-10 | Does **not** ship there |
| --- | --- |
| Add property control → navigate to composer | Composer layout (S7-08 / D4) |
| Cited-data rows + card growth | Locator tools |
| Connect disambiguation sheet + handoff | Pinning Citations across edits |
| Bridge card body once edges are real | Companion window / modal composer |

---

## 3. Card model updates (authoritative for D3)

Build on Spike 6 primary / bridge cards.

| Element | Intent |
| --- | --- |
| **Add property** | Clear control on activated/selected card (and/or always-visible quiet control). Opens navigation to composer for that subject. |
| **Cited rows** | Each Observation (or Property summary) as a compact row: property label + value summary + polarity hint if negative. Propose density at canvas zoom. |
| **Growth** | Card height grows with rows; propose max before scroll-inside-card vs always grow (prefer grow for tens of rows on a Source). |
| **Uncited → cited** | After ≥1 Observation about the subject, drop uncited shell (Spike 6 contrast frame). |
| **Bridge cards** | Subordinate chrome remains; body shows relationship/role summary from Observations once durable. |

---

## 4. Flows (authoritative)

### 4.1 Add property (primary)

```text
1. Select / activate card
2. Add property
3. If Source has no Artifact → gate (do not enter composer empty)
4. If multiple Artifacts → may pick on graph or defer pick to composer (propose; prefer composer-owned pick if simpler)
5. Navigate to citation composer place for this subject
6. On return (submit or Back): graph shows updated rows / selection restored if possible
```

### 4.2 Connect (durable)

```text
1. Connect tool: click A, then B
2. Disambiguation sheet on graph (role / relationship_type / refuse / shared-event choice) — §3.2 matrix
3. Confirm → navigate to composer with bridge + two edge Observations pre-filled
4. Submit → Back to graph with bridge card + edges; cancel → nothing written
```

Disambiguation is **not** a history entry. Composer place is.

---

## 5. Requirements

| ID | Requirement |
| --- | --- |
| EG-1 | **Add property** on primary cards; propose bridge-card variant (bridges may add further properties later). |
| EG-2 | **Cited-data rows** with property + value summary; cards grow. |
| EG-3 | Uncited vs cited shell wired to Observations (not to working label alone). |
| EG-4 | No-Artifact gate with clear recovery path. |
| EG-5 | Connect disambiguation sheet covering person→event, person→person, event→place, refusals. |
| EG-6 | Handoff to composer is navigational (leave canvas); propose any transient confirmation. |
| EG-7 | Graph a11y: new controls named; rotor/list still works; **do not** design composer a11y here. |
| EG-8 | Edge geometry still attaches sensibly to taller cards (note for implementers). |

---

## 6. Screen / frame inventory (minimum)

1. Uncited Person card with Add property.
2. Same card after two cited properties (grown body).
3. No-Artifact gate.
4. Connect disambiguation (person→person choice).
5. Bridge card after durable connect (cited body).
6. Annotation of navigation away (graph → composer) without drawing the composer.

---

## 7. Out of scope

- Composer layout, PDF viewer, polygon tools, observation form — **S7-D4**.
- Conflicted / competing Observation badges (later honesty spike).
- Pinning a Citation while staying on the graph.

---

## 8. Deliverable

Claude Design board + short notes for S7-09 / S7-10. Archive this brief when done.
