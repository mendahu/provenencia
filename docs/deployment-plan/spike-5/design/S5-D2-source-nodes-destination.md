# S5-D2 — Source nodes destination

**Kind:** Claude Design board
**Spike:** Provenencia Spike 5 (Interpretation foundation)
**Implements later as:** PR S5-09
**Depends on:** S5-D1 (how the researcher arrives here)
**Related brief:** [`S5-D1-interpretation-section.md`](S5-D1-interpretation-section.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design the destination behind a Source in the Interpretation section: **the list of candidate things this one Source appears to mention.**

The researcher can create a person, event, or place, give it a working label, rename it, and delete it. That is the whole capability. There are no facts, no dates, no names-as-evidence, and no relationships — those all require Citations and Observations, which arrive in a later spike.

**Read this part carefully, because it shapes every decision on the board.** In Spike 6 this destination becomes a **spatial canvas**: the same Nodes as draggable bubbles on a grid, connected by lines. This list does not get deleted when that happens — it becomes the **structured alternate view**: the keyboard-navigable, VoiceOver-legible, test-drivable representation of the same graph. So design something honest and small that survives demotion, not a flagship screen.

---

## 2. Domain facts the UI must reflect

### 2.1 What a Node is

A Node is a **candidate** — a thing one Source appears to mention, before any judgement about whether it is real or who it matches elsewhere.

| Concept | UI implication |
| --- | --- |
| Type | One of **person**, **event**, or **place** in this spike. Chosen at creation and **permanently immutable** — there is no "change type." Correcting a mistake means delete and re-create. |
| `ref` | Required, mono, never editable. Carries the type and the candidate marker: `PER-C-7KD45`, `EVT-C-…`, `PLC-C-…`. |
| `label` | **Optional and non-evidentiary.** See §2.2 — this is the single most important constraint on the board. |
| Facts | **None exist.** No name, no date, no place name, no relationships. A Node in this spike is genuinely just a typed, labelled placeholder. |

### 2.2 The label is a working handle, not a name

`label` is scratch. A researcher looking at a census page types things like *"head of household, line 14"* or *"the illegible one"*. It is a sticky note, not an assertion.

Later, a person's **name** will be recorded as a cited Observation — evidence, backed by a transcription and a locator, that the researcher can defend. The whole product exists to keep those two things apart.

**So the label must never be styled like an asserted name.** If the board makes the label look like the person's name, it teaches exactly the wrong mental model at the exact moment the researcher forms it, and the distinction the layer exists to protect is blurred at the point of entry. Treat it as annotation: secondary weight, clearly informal, comfortable being empty. A Node with no label at all is completely normal and must read as normal — the `ref` is the identity.

### 2.3 Nothing here is cited

Every Node on this screen is uncited, because citation does not exist yet. Do **not** design cited/uncited badges, evidence indicators, confidence, or completeness meters — there is nothing yet for them to reflect, and a "0 citations" badge on every row is noise that will need removing.

What is worth saying **once**, at the screen level, is that this is the beginning of the workflow and evidence comes next. Propose that framing.

---

## 3. Requirements

### 3.1 Screen shell

| ID | Requirement |
| --- | --- |
| N-1 | Make it unmistakable **which Source** is being interpreted — title and `SRC-…` — and offer a way back to that Source's page. The researcher arrived here from one of two places and must not lose the thread. |
| N-2 | Reuse the shipped destination chrome. Do not redraw the window, sidebar, toolbar, or breadcrumbs. |
| N-3 | Frame the screen's purpose in a line of copy: this is where you map what this Source appears to mention. Propose it. |

### 3.2 The node list

| ID | Requirement |
| --- | --- |
| N-4 | Use the shipped **`PVList`** pattern. No `PVTable`, no new component. |
| N-5 | Row anatomy: **type** (icon and/or label), the **working label** if present, and the **`PER-C-…`** ref in mono. Propose the slot assignment, respecting §2.2 — the label must not occupy the row's headline slot in a way that reads as a name. |
| N-6 | Design the **unlabelled row** as a first-class state, not a degraded one. It will be extremely common; a researcher placing six household members may label none of them. |
| N-7 | Types must be **distinguishable at a glance** — a person, an event, and a place should never be confused while scanning. Icons exist in the design system under the evidence icon set; reuse them. |
| N-8 | Grouping or sorting by type is **optional**. Propose only if it is cheap; a flat list in creation order is an acceptable answer. |
| N-9 | Empty state: this Source has no Nodes yet. This is the screen's most important frame — it is what the researcher sees the very first time they enter the Interpretation layer at all. Use `PVEmptyState` and make the next action obvious. |

### 3.3 Creating a node

| ID | Requirement |
| --- | --- |
| N-10 | Creating requires choosing a **type**, which is the one irreversible decision (§2.1). Propose the affordance — three distinct actions (Add person / Add event / Add place) or one action with a type picker. Say which you chose and why. |
| N-11 | Note for context: in Spike 6 this becomes a **tool palette** — pick a tool, then click the grid to place a bubble. Prefer a shape that foreshadows that rather than one that fights it. |
| N-12 | A label is **optional at creation.** Creating a Node with nothing but a type must be fast and must not feel like skipping a required field. |
| N-13 | Keep creation lightweight. If a dialog is used at all, reuse the existing form-dialog chrome; do not invent a modal system. Inline creation is also acceptable — propose. |

### 3.4 Editing and deleting

| ID | Requirement |
| --- | --- |
| N-14 | The label is **renameable in place**. Reuse the inline-edit pattern from the Source page title. |
| N-15 | There is **no type editing**. Do not design a type picker on an existing row. If the board wants to help a researcher who picked wrong, the honest answer is delete and re-create — propose how that is communicated, if at all. |
| N-16 | Delete uses the existing **`PVConfirm`** chrome. In this spike a Node has no dependents, so the confirmation is simple — but note on the board that once Observations exist, deleting a Node will destroy cited evidence and will need a confirmation that counts the damage. Do not design that now; do not design a pattern that cannot grow into it. |

---

## 4. Screen / frame inventory (minimum)

1. Destination with several Nodes — a mix of all three types, some labelled, some not.
2. Destination **empty** (first entry into the Interpretation layer).
3. The **create** affordance mid-flow, showing the type choice.
4. A row in **inline rename**.
5. Delete **confirmation**.
6. Optional: a note or annotation showing how this list relates to the canvas that replaces it in Spike 6.

---

## 5. Out of scope

- **The canvas.** No grid, no bubbles, no drag, no connecting lines, no pan/zoom, no spatial arrangement of any kind.
- Citations, Observations, properties, transcription, locators, the artifact viewer.
- Relationships between Nodes — including `participation`, `location`, and `relationship`. Those types exist in the vocabulary but nothing creates them in this spike.
- Names, dates, places-as-values, or any typed value editor.
- Cited/uncited badges, evidence counts, confidence, completeness (§2.3).
- Cross-Source display. Every Node here belongs to this Source; Nodes from other Sources are never shown.
- Node descriptions, notes, or per-Node detail pages.
- Bulk selection, multi-delete, import, export.

---

## 6. Acceptance checklist

- [ ] The working label is visibly **not** an asserted name, and an unlabelled Node reads as normal rather than incomplete.
- [ ] Person, event, and place are distinguishable at a glance using existing evidence icons.
- [ ] The empty state is strong — it is the first thing a researcher ever sees of this layer.
- [ ] Creation makes the type decision explicit and foreshadows a Spike 6 tool palette.
- [ ] Nothing on the board offers to change a Node's type.
- [ ] Nothing implies citation, evidence, confidence, or completeness.
- [ ] Delete reuses `PVConfirm` and could grow into a damage-counting confirmation later.
- [ ] The whole surface reuses shipped components — no new design-system primitives.
- [ ] The screen would still make sense as a secondary structured view once the canvas exists.

---

## 7. Implementation notes (for PRs — not Design homework)

| Item | Status |
| --- | --- |
| Why this screen exists | It proves the full vertical path — migration, Go, FFI, store, session cache, view — before any canvas code sits on top of it. It is also the structured editing path the design note commits to as the accessibility representation of the graph. |
| Type immutability | Enforced by the data model: `node_type_id` is immutable after insert because the type is baked into the public `ref`. N-15 is a schema fact, not a simplification. |
| Ref format | `{prefix}-C-{token}` from the Node Type's `ref_prefix`: `PER` person, `EVT` event, `PLC` place. Minted in Go; never user-editable. |
| Vocabulary | Seven Node Types are seeded, but only person, event, and place are creatable here. The other four (`relationship`, `participation`, `location`, `source`) need Observations to mean anything and must not appear in any picker. |
| Positions | The catalog stores grid coordinates for these Nodes, but nothing in this spike reads or writes them — that is Spike 6. Do not surface position anywhere on this board. |
