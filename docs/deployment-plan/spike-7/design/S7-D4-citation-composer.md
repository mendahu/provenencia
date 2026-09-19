# S7-D4 — Citation composer place (Option B)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PR **S7-08** only (viewers/locators land in S7-06/S7-07; this board owns place chrome)  
**Depends on:** Evidence graph handoff designed in **S7-D3**; NameValue editor designed separately in **S7-D5** (composer only hosts it)  
**Related briefs:** [`S7-D3`](S7-D3-evidence-graph-updates.md) — graph entry / return; [`S7-D5`](S7-D5-name-value-editor.md) — NameValue modal

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design the **citation composer** as a **first-class workspace place** (Option B): navigate away from the Evidence graph to a full-window side-by-side **artifact viewer \| form**.

```text
┌─────────────────────────────┬──────────────────────────┐
│  Artifact viewer            │  Citation fields         │
│  (image / PDF)              │  + locator tools summary │
│  + locator drawing overlay  │  + Observations list     │
│                             │  + Add observation       │
│                             │  Submit / Cancel         │
└─────────────────────────────┴──────────────────────────┘
```

Also decide **breadcrumbs**, title, and cancel/submit chrome so Back/Forward and relaunch behave honestly.

**Do not redesign the Evidence graph cards** — only the composer place and how it titles itself in the workspace chrome.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Citation needs an Artifact | First step or left-rail: pick Artifact when Source has many; auto-select when one. |
| Locator is composable | Tools: **page** (PDF), **region** polygon (image or PDF page). Nestable (page then region). |
| One Citation → many Observations | Form: citation block once; observation list with Add; single submit. |
| Value types | Editors for text, integer, date (existing DateValue), name (**open S7-D5 NameValue modal**), subject (graph-scoped picker). No real/boolean. |
| Transcription ≠ Observation | Citation: transcription, uncertain flag, note, description. Observations: Property + polarity + typed value. |
| Media MVP | Image + PDF only. Unsupported Artifact types: honest empty / disable with explanation. |
| Own a11y tree | This place is not an overlay on the canvas — design focus order for viewer + form. |
| Connect entry | May open with two Observations pre-filled (edge Properties); citation still required. |

### 2.1 What this board is not

- Not graph card growth / Add property chrome — **S7-D3**.
- Not Subject types / fields admin — **S7-D1 / D2**.
- Not the **NameValue editor** internals — **S7-D5** (only the host control that opens it).
- Not audio/video players.
- Not Citation pinning across multiple graph sessions.
- Not a sheet/modal covering the still-visible graph (Option A rejected).

### 2.2 Implementation gate (S7-08)

| Ships in S7-08 | Does **not** ship there |
| --- | --- |
| Place + breadcrumbs + viewer\|form layout | Wiring Add property on cards (S7-09) |
| Artifact pick, citation fields, N Observations, submit | Connect disambiguation sheet (S7-10 / D3) |
| Host affordance for NameValue + DateValue editors | Designing NameValue chrome (S7-D5 / S7-02b); companion NSWindow |

---

## 3. Navigation / breadcrumbs (decide on this board)

**Defaults from the deployment plan — confirm or revise:**

```text
Add property:
  Sources › {Source title} › Evidence graph › Cite {subject label}

Connect:
  Sources › {Source title} › Evidence graph › Connect › Cite
```

| Topic | Default | Board must decide |
| --- | --- | --- |
| History push | Entering composer pushes; Back returns to graph | Confirm |
| After submit | Pop/replace to graph (no draft stack) | Confirm |
| Cancel / Back | No catalog writes | Confirm |
| Relaunch | Restore composer if subject still exists; else Evidence graph | Confirm |
| Disambiguation | Stays on graph; not a crumb | Confirm |
| Leaf title | “Cite …” vs “Add citation” vs Property-led | Propose final copy |

Workspace location needs a discriminant beyond page|graph (e.g. `citationComposer` + `subjectId` + optional connect context). Design should not invent a new sidebar section.

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CC-1 | Full-window **viewer \| form** composition; usable on a typical laptop width; propose collapse/stack behavior if narrow. |
| CC-2 | **Artifact picker** when multiple Artifacts; single-Artifact auto-select; empty Artifact state with recovery. |
| CC-3 | **Image viewer:** zoom + pan; region polygon tool overlay. |
| CC-4 | **PDF viewer:** page prev/next + page number entry; zoom + pan; region on current page. |
| CC-5 | **Locator tool chrome:** select tool (page / region); clear/edit locator; show active selector chain readably. |
| CC-6 | **Citation fields:** transcription, uncertain + note, description (and notes if in scope). |
| CC-7 | **Observations list:** Add/remove rows; Property picker filtered by subject type bindings; polarity; typed value editor. |
| CC-8 | **Host** NameValue modal designed in **S7-D5**; reuse DateValue; subject picker scoped to this Source’s graph. Do not redesign NameValue on this board. |
| CC-9 | **Submit** writes one Citation + N Observations; **Cancel / Back** writes nothing. |
| CC-10 | Breadcrumbs + page title per §3; toolbar Back works. |
| CC-11 | Accessibility: complete keyboard path for form and tools; VoiceOver structure for this place alone. |
| CC-12 | Connect-prefilled mode: show that two edge Observations are required/pre-filled without looking like a broken empty form. |

---

## 5. Screen / frame inventory (minimum)

1. Composer with PDF, page selected, one Observation (text).
2. Composer with image + polygon region tool active.
3. Artifact picker (multi-Artifact Source).
4. Observations list with two rows (e.g. name + occupation); annotation that name opens the **S7-D5** modal (do not detail that modal here).
5. Connect-prefilled composer (two subject-valued edges).
6. Breadcrumb states: Add-property path and Connect path.
7. Unsupported media / no Artifact empty states.

---

## 6. Out of scope

- Graph card layouts (S7-D3).
- NameValue editor layout / parts UX (S7-D5).
- Audio/video/`time_range`.
- Companion window.
- Editing an existing Citation in place (create path first unless trivial).

---

## 7. Deliverable

Claude Design board + short notes for S7-08 (and breadcrumb/history decisions locked for `add-workspace-location`). Archive this brief when done.
