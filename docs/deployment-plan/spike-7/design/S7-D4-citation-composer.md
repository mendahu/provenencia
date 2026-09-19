# S7-D4 — Citation composer place (Option B)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PR **S7-08** (thin submit path); viewers/locators/NameValue fill in via **S7-06 / S7-07 / S7-02b**  
**Depends on:** Evidence graph handoff designed in **S7-D3**; NameValue editor designed separately in **S7-D5** (composer only hosts it)  
**Related briefs:** [`S7-D3`](S7-D3-evidence-graph-updates.md) — graph entry / return; [`S7-D5`](S7-D5-name-value-editor.md) — NameValue modal  
**Phase the board:** prioritize shell + form + text/term Observations + breadcrumbs (enough for S7-08). Viewer, locator tools, and NameValue host can be later frames on the same board.

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
| Value types | Editors for text, integer, **term** (searchable term picker), date (existing DateValue), name (**open S7-D5 NameValue modal**), subject (graph-scoped picker). No real/boolean. |
| Property terms | Kind/edge Properties (`event_type`, `role`, `relationship_type`) pick a `property_terms` row — not free text. Product terms are fixed; **Add custom…** / rename / delete (when unused) for `origin=user` only. No Event types / Roles destinations. |
| Transcription ≠ Observation | Citation: transcription, uncertain flag, note, description. Observations: Property + polarity + typed value. |
| Media MVP | Image + PDF only. Unsupported Artifact types: honest empty / disable with explanation. |
| Own a11y tree | This place is not an overlay on the canvas — design focus order for viewer + form. |
| Connect entry | May open with two Observations pre-filled (edge Properties); citation still required. |

### 2.1 What this board is not

- Not graph card growth / Add property chrome — **S7-D3**.
- Not Subject fields admin — **S7-D2**. Not Subject types admin (descoped).
- Not the **NameValue editor** internals — **S7-D5** (only the host control that opens it).
- Not audio/video players.
- Not Citation pinning across multiple graph sessions.
- Not a sheet/modal covering the still-visible graph (Option A rejected).

### 2.2 Implementation gates (thin → thick)

| Ships in **S7-08** (thin) | Later PRs | Never this board |
| --- | --- | --- |
| Place + breadcrumbs + form; Artifact pick; text Observations; submit | Image/PDF viewers (**S7-06**); locators (**S7-07**); NameValue host (**S7-02b**) | Add property / card growth (**S7-D3** / **S7-09**) |
| Left pane may be placeholder or Artifact title only | Full viewer\|form composition | Connect disambiguation (**S7-10**) |
| DateValue reuse if needed | NameValue modal chrome | Designing NameValue (**S7-D5**); companion NSWindow |

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
| CC-9 | **Term picker** for `value_type = term`: searchable list of installed terms for that Property; **Add custom…** creates `origin=user` term; rename/delete user terms when unused; product/plugin terms not editable. |
| CC-10 | **Submit** writes one Citation + N Observations; **Cancel / Back** writes nothing. |
| CC-11 | Breadcrumbs + page title per §3; toolbar Back works. |
| CC-12 | Accessibility: complete keyboard path for form and tools; VoiceOver structure for this place alone. |
| CC-13 | Connect-prefilled mode: show that two edge Observations are required/pre-filled without looking like a broken empty form. |

---

## 5. Screen / frame inventory (minimum)

**Must for S7-08 (thin):**

1. Composer shell with form + text or **term** Observation; left pane placeholder or Artifact metadata only.
2. Artifact picker (multi-Artifact Source).
3. Breadcrumb states: Add-property path (and Connect path if sketched).
4. No-Artifact empty state.

**Later frames (same board, thicken after thin ships):**

5. Composer with PDF, page selected.
6. Composer with image + polygon region tool active.
7. Observations list with two rows (e.g. name + occupation); annotation that name opens the **S7-D5** modal.
8. Connect-prefilled composer (two subject-valued edges).
9. Unsupported media empty states.

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
