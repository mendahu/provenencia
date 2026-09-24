# S8-D7 — Citation composer rethink

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-10**  
**Depends on:** Shipped Option B composer ([`CitationComposerView`](../../../../macos/App/Features/CitationComposer/CitationComposerView.swift), S7-D4 / S7-08); graph handoff (`composerLocation`); locators S7-07  
**Related:** dogfood “one reading, many subjects” ([`docs/dogfood/ux.md`](../../../dogfood/ux.md)); archived [S7-D4](../../archive/spike-7/design/archive/S7-D4-citation-composer.md)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md); location focus via [`add-workspace-location`](../../../../.cursor/skills/add-workspace-location/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

**Hand this board first.** **S8-D1** (Auto Transcribe) and **S8-D2** (PDF Find / paste) land on this form. Do not design those controls until this layout is agreed.

---

## 1. Objective

Make the citation composer a **single flexible place** that can express the Interpretation layer: one **Citation** (a reading of an Artifact) with **many Observations**, each Observation naming a **Subject**.

Today the composer is **subject-locked**. Add property always mints a Citation. Edit filters the loaded Citation to one card’s rows. Save refuses zero Observations. The schema already allows one Citation × many subjects, reuse, and an empty reading.

**Do this first in Spike 8** so we do not restyle the old form and then throw it away.

The shell can stay Option B: **navigable place**, artifact viewer beside a form, Back to the Evidence graph. **Do not give up the graph.** Do not rebuild Subjects as a fourth list in the composer. **Rethink the form and how identity is chosen** — not the idea of a place.

```text
Today (locks):
  card → [artifact pre-screen] → form for THIS subject only

Target (focus):
  entry pre-selects artifact / citation / focused observation / subject
  researcher can change any of those without leaving the place
```

**Density is the design problem.** Artifact + citation identity + transcription + locator + every observation (with a subject on each row) is a lot of data. A stack of list → selector → list → selector will fail. The board must propose **more than one layout idea**, then pick one. Cleverness is required; a file-browser for the interpretation layer is not.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Citation is the **document** | Artifact + locator + transcription + description. Not subject-scoped. Not source-scoped (source is via the Artifact). |
| Observation is a **child of the Citation** | `subject_id` is a field on the row. It is not a fourth navigation layer. |
| Subject lives on the **other** Source branch | The graph is the subject / relationship view. Composer names a subject; it does not browse the graph as nested lists. |
| One Artifact → many Citations | Need a **citation identity** control on this Artifact (new vs existing `CIT-…`). |
| One Source → many Artifacts | Artifact identity belongs **in** the compose surface. **Replace** today’s full-window pre-screen (`CitationComposerArtifactPicker`). One Artifact still auto-selects. |
| Switching Artifact on a **saved** Citation | Citations are `artifact_id` + `ON DELETE CASCADE`. You cannot move a reading. Changing Artifact is “abandon this Citation / start another,” not a field edit. Viewer + locator reset. Downstream citation + observation selection clears. |
| Switching Citation | Observation list replaces. Focused row clears. Transcription / locator load with the Citation. |
| Switching Observation focus | Editor loads that row. Citation chrome stays. |
| Retargeting `subject_id` | That Observation moves to another card. Nothing above it invalidates. |
| Empty Citation is legal in the schema | Save with zero Observations (“I transcribed this; I have not interpreted it yet”). Today only **Swift** (and `CreateWithObservations` ≥1) refuse. |
| Shared reading | Editing transcription while focused on Mary’s birth date also changes the quote on John’s name. Show the Citation ref and that other cards share this reading. |
| Connect-edge rows stay special | Endpoint Properties stay fixed. Add observation still excludes them. Flexibility is not “those rows become generic.” |
| Scope | Artifacts and Citations for **this Source**. Not the whole project. Subjects offered on a row are this Source’s graph subjects (minus `source` type). |
| Entry is **pre-selection**, not a mode | One form. Do not toggle “Add mode” vs “Edit mode” layouts. |

### 2.1 Hierarchy (do not build four master–details)

```text
Source
  Artifact → Citation → Observation
                          └──► Subject   (same Source, other branch)
```

You need an Artifact to have a Citation, and a Citation to have an Observation. You do **not** open a Subject to finish an Observation — you name it.

**Valuable lists (exactly two):**

1. Citations on this Artifact (identity / reuse).
2. Observations on this Citation (each row: subject + property + value summary).

Everything else is the **viewer** (Artifact) or the **graph** (Subjects + Connect).

### 2.2 Entry points (initial state only)

| Entry | Artifact | Citation | Observation focus | Subject |
| --- | --- | --- | --- | --- |
| **Add property** on a card | pre-select if one Artifact; else pick in-form | **new** (reuse offered) | none, or one draft row aimed at the card | that card |
| **Pencil** on a property | that Citation’s Artifact | that Observation’s Citation | **that row** | that card |
| **Connect** | as today | **new** | two **fixed** edge rows | the two endpoints |
| **Open / pin** a Citation | its Artifact | that Citation | none | none (or last-used filter, if the board wants a quiet chip) |
| **Transcribe first** | picked | **new** | none | none; Save allowed |

Pencil **must not hide** other subjects’ rows. Load the whole Citation; focus the row.

`WorkspaceLocation` already has `sourceId` + `subjectId` + `citationId`. Those are **focus, not locks**. `subjectId` may be nil. Observation focus is **ephemera** (do not add `observationId` to the location unless restore-into-that-row becomes a product requirement). Connect payload stays (`subjectId` nil + endpoint ids).

### 2.3 Replace the artifact pre-screen

Today, creating a Citation with 2+ Artifacts is a **separate phase** (picker, then compose). That is the missing identity control pretending to be a wizard step.

Fold Artifact choice into the compose layout: thumb strip, popover, title menu — board finds the compact control. Changing it later is allowed (new Citation / abandon saved). Do not keep a full-page gate that hides transcription and observations.

### 2.4 Density — the board must be clever

This will not work if every layer is a visible list plus a selector that reveals the next list.

**Explore at least two (preferably three) layout ideas** before picking. Directions worth trying — not a mandate to ship all of them:

- **Citation identity is chrome, not a page.** Compact Artifact switcher + Citation switcher (`CIT-…` + transcription snippet). Observation list is the scroll. Keep the existing **observation dialog** for property/value so the list stays one-line rows (subject · property · summary).
- **Sticky reading, expandable transcription.** One-line quote + Uncertain; expand to the full textarea. Locator stays with the viewer tools, not a second essay in the sidebar.
- **Do not add a third full pane** (graph \| viewer \| form) unless a finding proves the graph hop is still too expensive *after* a subject picker on the row exists. Default remains: viewer \| form, toggle to graph via Back / existing jump.
- **Reuse is a popover**, not a second column of every Citation on the Source. List is **this Artifact only**.
- **Shared-reading hint** is a caption or chip (“also on 3 cards”), not a replica of the graph.
- **Census-scale:** one household Citation may have dozens of rows. Virtualize / quiet grouping by subject if needed; do not open every value editor inline.

**Reject on sight:** four simultaneous master–detail columns; a “subject composer” mode that hides the Citation; tabs that put transcription and observations on different pages so you cannot see the quote while placing.

### 2.5 What this board is not

- Not Auto Transcribe (**S8-D1**) or PDF Find / I-beam / paste (**S8-D2**) — leave a home for those controls on the **new** transcription / viewer chrome.
- Not minting Subjects from the composer (create stays on the graph).
- Not Foundation Models / PDF OCR / field detectors ([`docs/dogfood/ux.md`](../../../dogfood/ux.md)).
- Not density **filters** on the graph (still dogfood).
- Not source-to-source commentary, `text_quote` locators, audio/video players.
- Not a sheet over the live graph (Option A stays rejected).
- Not changing Connect rules or edge-property exclusion.
- Not delete-path matrix (**S8-D6**). Draft-row remove stays; persisted Observation delete follows that brief.

---

## 3. Implementation gate (S8-10)

| Ships in **S8-10** | Does **not** ship there |
| --- | --- |
| In-form Artifact identity; **remove** the create-time pre-screen | OCR / Vision; PDFKit remount / Find / paste |
| Citation identity: new vs existing on this Artifact | Project-wide Citation search |
| Observation list = **all** rows on the Citation; subject control per row | Creating Subjects; subject-valued Properties except shipped Connect edges |
| Save with **zero** Observations; reuse existing Citation from Add property | Pinning as a separate feature (reuse + last-used *is* pinning) |
| Entry points only set focus; pencil does not filter other subjects | `WorkspaceLocation.observationId` unless the board proves restore needs it |
| Connect-prefilled edge rows unchanged in meaning | Graph visual extras (**S8-06**); 3-pane graph+composer |
| FFI/list if the client cannot yet list Citations by Artifact (`ListByArtifact` exists in Go) | New catalog tables |

**Engine:** `citations.CreateWithObservations` requires ≥1 Observation today. Empty Save needs a create-without-observations path (relax that function or add `Create`). Schema already allows it.

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CR-1 | Board shows **one** compose surface (not a wizard) with Artifact identity, Citation identity, reading fields, and the Observation list. |
| CR-2 | At least **two layout alternatives** for density, then a recommendation. Explain what was rejected (especially any 4-stack). |
| CR-3 | Artifact pre-screen is **gone**. 1 Artifact → auto. 2+ → in-form picker. No Artifact → same honest empty as today (go to Source page). |
| CR-4 | Citation switcher: **New** + existing on **this** Artifact (`ref` + short transcription / locator hint). Switching a saved Citation to another Artifact is abandon / new, with confirm if dirty. |
| CR-5 | Observation rows show **which subject**; researcher can change it (graph subjects on this Source). Add observation asks for subject (pre-filled from entry when we have one). |
| CR-6 | Pencil entry: load that Citation, **all** rows, focus the row. Do not hide other subjects. |
| CR-7 | Add property entry: subject pre-selected; Citation **new** by default; **reuse** offered without leaving. |
| CR-8 | Save with empty Observation list is allowed and does not look like an error. Today’s “need an observation” callout goes away as a blocker (empty list can still be a quiet hint). |
| CR-9 | Connect entry still shows two **fixed** endpoint rows + optional term; Property picker still excludes edge keys. |
| CR-10 | Shared-reading honesty: Citation `ref` visible; if other subjects have rows, that is obvious. |
| CR-11 | Viewer \| form (or the chosen alternative) still has an a11y tree as its own place. Cancel / Back returns to the graph. After Save, go to the graph as today. |
| CR-12 | Prefer existing `PV*` (`PVField`, `PVButton`, `PVThumbnail`, popover/menu, observation dialog). No new kit primitive unless a finding says the kit cannot do an identity switcher. |
| CR-13 | VoiceOver: Artifact, Citation, and focused Observation are named. Changing Artifact or Citation is announced as a document change. |
| CR-14 | L10n for new chrome. `Text(verbatim:)` for refs / counts. |

---

## 5. Suggested frames

1. **Add property, one Artifact** — new Citation, subject pre-selected, empty observations (Save enabled).
2. **Add property, many Artifacts** — identity on the form (no pre-screen); pick Artifact, then New vs existing Citation.
3. **Reuse** — same card, pick an existing `CIT-…` on that Artifact; transcription fills; add a row for this subject.
4. **Pencil** — shared Citation with rows on two subjects; focus one row; other subject still listed.
5. **Transcribe first** — reading filled, zero observations, Save, return; reopen from a Citation identity control.
6. **Connect** — two fixed edge rows (regression; do not look broken).
7. **Dirty Artifact change** — confirm abandon vs stay.
8. **Census-ish list** — many observation rows on one Citation; still scannable.
9. **No Artifact** — existing callout + go to Source page.
10. *(Optional third layout exploration as a discarded frame.)*

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Composer place | Snowflake | **Rethink** | `Features/CitationComposer/` | Still Option B. Layout of form + identity is this brief. |
| Artifact pre-screen | Snowflake | **Remove** | `CitationComposerArtifactPicker.swift` | Replace with in-form identity. |
| Form pane | Snowflake | **Rethink** | `CitationComposerFormPane.swift` | Identity + reading + full observation list. |
| Observation row | Snowflake | **Extend** | `CitationComposerObservationRow` | Subject on the row. |
| Observation dialog | Snowflake | **Extend** | `CitationComposerObservationDialogForm` | Subject control; keep value editors here if density needs it. |
| Composer model | Snowflake | **Rethink** | `CitationComposerModel.swift` | Drop subject-lock filter; empty Save; list-by-artifact. |
| Artifact viewer | Snowflake | **Keep** | `Features/ArtifactViewer/` | Slot may move; remount/Find is **S8-D2**. |
| Graph cards / Add property / pencil | Snowflake | **Keep** | `Features/EvidenceGraph/` | Same entries; they only pre-select. |
| `WorkspaceLocation` | Place | **Extend** | `subjectId` / `citationId` optional focus | No new place. |
| List Citations by Artifact | Engine / FFI | **Expose** if missing | `core/database/citations.ListByArtifact` | Client needs it for the switcher. |
| Create Citation, zero Observations | Engine | **Add / relax** | `CreateWithObservations` today requires ≥1 | Schema already allows empty. |
| Thumbnail / menu / field / dialog | Component | Ship | `PVThumbnail`, `PVField`, `PVButton`, existing dialog | Identity switcher composes these. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Subject list destination in the composer | That is the graph. |
| `PVCitationSwitcher` kit control | One call site unless a second appears. Compose menu + thumbs. |
| Four-column file browser | The failure mode this brief exists to avoid. |
| Live graph pane by default | Option B + hop stays until a finding says otherwise. |

---

## 7. Out of scope

- **S8-D1** / **S8-D2** control design (placement hole only)
- Graph chrome (**S8-D3**), Source page jump (**S8-D4**), list counts (**S8-D5**), delete matrix (**S8-D6**)
- Auto Observations / LLM extract
- Creating Subjects or changing Connect rules
- Audio / video Sources; `text_quote`

---

## 8. Handoff

1. Agree the layout (including the discarded alternatives) before **S8-10**.
2. Archive this brief under `archive/` when the board is agreed.
3. Record in [`../completed.md`](../completed.md).
4. Implement **S8-10**, then design **S8-D1** / **S8-D2** against the new chrome.
