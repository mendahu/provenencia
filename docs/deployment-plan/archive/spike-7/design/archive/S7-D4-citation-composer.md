# S7-D4 — Citation composer place (Option B)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PR **S7-08** (thin submit path); viewers/locators/NameValue fill in via **S7-06 / S7-07 / S7-02b**  
**Depends on:** Evidence graph handoff designed in **S7-D3**; NameValue editor designed separately in **S7-D5** (composer only hosts it)  
**Related briefs:** [`S7-D3`](../archive/S7-D3-evidence-graph-updates.md) — graph entry / return; [`S7-D5`](S7-D5-name-value-editor.md) — NameValue modal  
**Phase the board:** prioritize shell + form + text/term Observations + breadcrumbs (enough for S7-08). Viewer, locator tools, and NameValue host can be later frames on the same board.

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](../README.md) first.

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
| Value types | Editors for text, integer, **term** (searchable term picker), date (existing DateValue), name (**open S7-D5 NameValue modal**), subject (graph-scoped picker **only when the Property is not a connect edge** — see §2.3). No real/boolean. |
| Property terms | Kind/edge Properties (`event_type`, `role`, `relationship_type`) pick a `property_terms` row — not free text. Product terms are fixed; **Add custom…** / rename / delete (when unused) for `origin=user` only. No Event types / Roles destinations. |
| Transcription ≠ Observation | Citation: transcription, uncertain flag, note, description. Observations: Property + polarity + typed value. |
| Media MVP | Image + PDF only. Unsupported Artifact types: honest empty / disable with explanation. |
| Own a11y tree | This place is not an overlay on the canvas — design focus order for viewer + form. |
| Connect entry | Opens with **two edge Observations pre-filled and fixed** (endpoint Properties); citation still required. Researcher does **not** pick those Properties from Add observation. |

### 2.1 What this board is not

- Not graph card growth / Add property chrome — **S7-D3**.
- Not Subject fields admin — **S7-D2**. Not Subject types admin (descoped).
- Not the **NameValue editor** internals — **S7-D5** (only the host control that opens it).
- Not audio/video players.
- Not Citation pinning across multiple graph sessions.
- Not a sheet/modal covering the still-visible graph (Option A rejected).
- Not inventing a second admin surface for “which Properties are edges” — that is the Interpretation subject registry connect matrix (**S7-01**).

### 2.2 Implementation gates (thin → thick)

| Ships in **S7-08** (thin) | Later PRs | Never this board |
| --- | --- | --- |
| Place + breadcrumbs + form; Artifact pick; text Observations; submit; **Property picker excludes connect-edge endpoints** (§2.3) | Image/PDF viewers (**S7-06**); locators (**S7-07**); NameValue host (**S7-02b**) | Add property / card growth (**S7-D3** / **S7-09**) |
| Left pane may be placeholder or Artifact title only | Full viewer\|form composition | Connect **disambiguation sheet** on the graph (**S7-10**) — but design the **Connect-prefilled composer** frame now |
| DateValue reuse if needed | NameValue modal chrome | Designing NameValue (**S7-D5**); companion NSWindow |

### 2.3 Connect-edge Properties are system-owned (authoritative)

Bridge endpoint Properties — the subject-valued edges that attach a primary to a bridge (`person`, `event`, `place`, `related_to` on participation / location / relationship) — are **not** researcher-picked Observations.

| Path | Behavior |
| --- | --- |
| **Add property** (primary or bridge) | Property picker = bindings for that Subject type **minus** every Property key that appears in the registry connect matrix as an **edge endpoint** for that type (`ConnectRule.EdgePropertyKeys` where `BridgeTypeKey` matches). Researcher cannot manually add a `person` / `related_to` / … Observation. |
| **Connect** (S7-10 → this place) | Composer opens with the two endpoint Observations **already scoped** (values = the two connected subjects). Rows are visible, read as fixed / system, and **not** choosable or replaceable via the Property picker. Cancel / Back writes nothing. |
| **Disambiguation terms** (`role`, `relationship_type`) | **Not** edge endpoints. Collected on the graph Connect sheet (or shown as a normal term Observation once the composer opens). Still offered when Add property lands on a bridge **if** bound and not an edge key. |
| **Locked ≠ connect-edge** | Registry `Locked` on bindings (e.g. event `date` / `start_date` / `end_date`) means Subject fields must not unbind — those Properties **remain** manually citable from Add property. Do not conflate lock with “Connect only.” |

**Implementation preference (for S7-08 / S7-10, not this board’s UI):** derive the exclude set from the existing connect matrix (`EdgePropertyKeys` per bridge type). Do **not** hard-code Property keys in the Swift picker. A dedicated binding flag is unnecessary while the matrix already names the endpoints; plugins that add bridges must declare edges on their connect rules the same way.

**Board must show:** (1) Add-property composer Property picker with **no** edge endpoint Properties; (2) Connect-prefilled composer with two fixed endpoint rows that do not look like empty broken form fields.

---

## 3. Navigation / breadcrumbs (decide on this board)

**Locked default (same leaf shape for Add property and Connect):**

```text
Sources › {Source title} › Evidence graph › Citation for {scope}
```

The citation is **subject-scoped** — the leaf names what the Citation is *for*, not “citing the subject as a Source.” Use one crumb shape for both entry paths; do **not** insert a separate `Connect ›` segment.

| `{scope}` | When | Copy source |
| --- | --- | --- |
| Primary subject | Add property on person / event / place | Working **subject label** (same string the card leads with) |
| Bridge subject | Connect handoff, or Add property on a bridge | Full sentence: **`{endpoint A label} {edge phrase} {endpoint B label}`**, using the same edge-phrase templates as the durable bridge card ([S7-D3 §3.1](../archive/S7-D3-evidence-graph-updates.md)) plus the two connected subjects’ working labels. Example: “John is the father of Mary”, “Margt. participated as head of household at 1851 census”, “Marriage took place in Leeds”. Do **not** use the bridge working `label` or a bare type word (“Relationship”). The bridge **card** still omits endpoint names in its body (those stay on the root cards); this crumb is the one place that joins endpoints + phrase for orientation. Until the phrase or an endpoint label is incomplete, fall back to the best readable partial (e.g. honesty / missing term), still under **Citation for …**. |

Examples:

```text
Sources › Alderwick family bible › Evidence graph › Citation for Margt. Alderwick
Sources › Alderwick family bible › Evidence graph › Citation for John is the father of Mary
Sources › Alderwick family bible › Evidence graph › Citation for Margt. participated as head of household at 1851 census
Sources › Alderwick family bible › Evidence graph › Citation for Marriage took place in Leeds
```

| Topic | Default | Board must decide |
| --- | --- | --- |
| History | Composer is a **first-class** `WorkspaceLocation` — enter with `go(to:)`, Back/Forward like any other deep place. **No** special stack surgery for Cancel, submit, or “draft” entries. Cancelled / empty visits may remain in history; Forward returns to the same place (fresh form chrome — unsaved field values are UI ephemera, not history). | Confirm |
| After submit | Navigate to the Evidence graph with normal `go(to:)` / Back — do **not** pop/replace to scrub the composer from the stack. | Confirm |
| Cancel / Back | No catalog writes; stack unchanged beyond normal Back. | Confirm |
| Relaunch | Same as any deep place: restore composer if `subjectId` still exists; else fallback to Evidence graph for that Source. | Confirm |
| Disambiguation | Stays on graph; not a history entry. | Confirm |
| Leaf copy | **Citation for {scope}** — one template; primary = label; bridge = **endpoints + edge phrase** as above | Confirm truncation / VoiceOver if the sentence is long |

Workspace location needs a discriminant beyond page|graph (e.g. `citationComposer` + `subjectId`). Connect-prefilled edge rows must be **reconstructible from that location** (bridge subject + endpoints / provisional links) — that is place payload, not a parallel draft stack. Design should not invent a new sidebar section.

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
| CC-7 | **Observations list:** Add/remove rows; Property picker filtered by subject type bindings **and excluding connect-edge endpoint Properties** (§2.3); polarity; typed value editor. |
| CC-8 | **Host** NameValue modal designed in **S7-D5**; reuse DateValue; subject picker scoped to this Source’s graph **only for non-edge subject Properties** (if any). Do not redesign NameValue on this board. |
| CC-9 | **Term picker** for `value_type = term`: searchable list of installed terms for that Property; **Add custom…** creates `origin=user` term; rename/delete user terms when unused; product/plugin terms not editable. |
| CC-10 | **Submit** writes one Citation + N Observations; **Cancel / Back** writes nothing. |
| CC-11 | Breadcrumbs + page title per §3; toolbar Back works. |
| CC-12 | Accessibility: complete keyboard path for form and tools; VoiceOver structure for this place alone. |
| CC-13 | **Connect-prefilled mode:** two endpoint edge Observations are present, **fixed** (not Property-picker rows), with clear system/owned chrome so the form does not look empty or broken. Disambiguation term (`role` / `relationship_type`) may appear as a normal editable term row. |
| CC-14 | **Add observation** never offers connect-edge endpoint Properties for the subject’s type; empty picker states must not invite “pick the other person” as a manual Observation. |

---

## 5. Screen / frame inventory (minimum)

**Must for S7-08 (thin):**

1. Composer shell with form + text or **term** Observation; left pane placeholder or Artifact metadata only. Property picker shows **no** connect-edge endpoints (§2.3).
2. Artifact picker (multi-Artifact Source).
3. Breadcrumb states: **Citation for {label}** (primary) and **Citation for {A} {edge phrase} {B}** (bridge / Connect) — same crumb depth, no Connect segment.
4. No-Artifact empty state.

**Later frames (same board, thicken after thin ships):**

5. Composer with PDF, page selected.
6. Composer with image + polygon region tool active.
7. Observations list with two rows (e.g. name + occupation); annotation that name opens the **S7-D5** modal.
8. **Connect-prefilled** composer: two **fixed** endpoint edge rows + optional disambiguation term; Property picker still excludes those endpoints if Add observation is offered.
9. Unsupported media empty states.

---

## 6. Out of scope

- Graph card layouts (S7-D3).
- NameValue editor layout / parts UX (S7-D5).
- Audio/video/`time_range`.
- Companion window.
- Editing an existing Citation in place (create path first unless trivial).
- Manual creation of connect-edge endpoint Observations (system / Connect only — §2.3).
- Connect disambiguation sheet chrome on the graph (S7-10); this board only designs how the composer **receives** that handoff.
---

## 7. Deliverable

Claude Design board + short notes for S7-08 (and breadcrumb/history decisions locked for `add-workspace-location`). Archive this brief when done.
