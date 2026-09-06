# S2-02 — Source catalog IA and create/edit

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 2 (Source layer validation)  
**Implements later as:** PRs S2-15 (list/detail shell), S2-16 (create/edit + notes + metadata)  
**Depends on:** S2-01 workspace chrome (content mounts inside that host)  
**Related briefs:** S2-03 (Artifacts section of detail), S2-04 (extending types/fields)

Paste this entire document into Claude Design as the requirements for one board/flow.

---

## 1. Objective

Design the **Sources** destination inside the app workspace: catalog list, empty states, create-Source flow, and Source detail for descriptive catalog fields (type, title, description, notes, metadata). This validates the Source-layer **catalog** model — evidence description without genealogical interpretation.

Assume S2-01 chrome exists. **Do not redraw a second window shell.**

---

## 2. Domain model (UI must reflect)

Authoritative schema: Source layer data model. Relevant concepts:

### 2.1 Source

A **Source** is the evidentiary object (“what evidence do we possess?”), not a person or event.

| Field / concept | UI implication |
| --- | --- |
| `ref` (`SRC-…`) | Always visible on detail; list may show it secondary. Monospace. Never treat as editable. |
| `source_type_id` → type `label` | Primary classification control (picker). Types are extensible vocabulary, not a fixed enum. |
| `title` | Optional text; primary list headline when present; fallback needed when empty (e.g. type + ref). |
| `description` | Catalog text about the Source itself (not researcher commentary). Multiline. |
| Notes (`source_notes`) | **Multiple** accumulating notes — not a single overwriteable notes field. Each note has a body; attribution is audit (no per-note author chrome required in Spike 2). |
| Metadata | Descriptive key/value catalog fields; see §2.3. |
| Artifacts | Concrete representations; **design placement** on detail (section/tab), but ingest UI is S2-03. |

### 2.2 Source types

| Rule | UI implication |
| --- | --- |
| Seeded builtins (`photograph`, `book`, `birth_certificate`, …) | Show **labels**, not raw keys, in pickers. |
| User-extensible | Type picker must allow choosing existing types; adding types is S2-04 (link/entry point OK). |
| Type does not change `SRC-` prefix | Never invent type-specific id formats. |
| Type does not imply specialized forms beyond suggested metadata | One Source detail layout; metadata suggestions change with type. |

### 2.3 Metadata

| Rule | UI implication |
| --- | --- |
| Fields have `label`, `key`, `data_type` (`text` \| `date`) | Render text inputs vs date-capable inputs by type. |
| Type may **suggest** fields via join table | Detail form should prefer suggested fields for the current type; order by `sort_order` when present. |
| Suggestions are not required | User may leave suggested fields empty; may add values for fields not suggested (S2-04 covers “add field”; this board should leave room for “additional fields” / non-suggested values). |
| Values: `value_text` always useful; dates may also have structured date | For date fields, preserve human wording (e.g. “about the year 1890”) — do not force a single calendar widget that destroys phrase fidelity. Minimal structured entry is OK (year / exact day / “about”). |
| No Source credibility grade here | Do not show Low/Standard/High trust on Source catalog UI (Interpretation assessment). |

### 2.4 What Source UI must not become

- No person picker, relationship graph, or “attach to tree.”
- No Citation / Observation / transcription workspace (Interpretation).
- No GEDCOM import in this flow.
- No audit timeline required on the detail screen for Spike 2.

---

## 3. Requirements

### 3.1 Catalog list (Sources destination)

| ID | Requirement |
| --- | --- |
| S-1 | List Sources for the open project inside the S2-01 content host. |
| S-2 | Each row shows enough to scan: **title** (or fallback), **type label**, **`SRC-…`**. |
| S-3 | Empty state when the project has zero Sources: explain what a Source is in one short sentence; primary CTA to create. |
| S-4 | Selecting a row opens Source detail (push, split, or in-place — pick one pattern and stay consistent with macOS density). |
| S-5 | Provide an obvious **Create Source** action from the list (toolbar, empty state, or both). |

### 3.2 Create Source flow

| ID | Requirement |
| --- | --- |
| S-6 | Create flow collects at least: **type** (required), **title** (optional but encouraged), optional short **description**. |
| S-7 | On save, the user lands on the new Source’s detail with a visible minted **`SRC-…`** (ref is system-assigned, not typed by the user). |
| S-8 | Cancel / dismiss create without leaving orphan chrome. |
| S-9 | Do not require Artifacts or metadata to create a Source (fileless catalog stub is valid). |

### 3.3 Source detail — identity and core fields

| ID | Requirement |
| --- | --- |
| S-10 | Detail header/identity strip shows **`SRC-…`**, **type**, and **title** (editable). |
| S-11 | **Description** editable as multiline catalog text. |
| S-12 | Changing type is allowed; if metadata suggestions change, the board should show how existing values remain (values are not deleted silently in UX copy — warn or keep “additional” section). |
| S-13 | No `created_by` / timestamp fields on the Source form — those live in audit infrastructure, not domain columns. |

### 3.4 Notes

| ID | Requirement |
| --- | --- |
| S-14 | Notes are a **list** of entries the researcher can add over time. |
| S-15 | Support add note and edit/delete affordances (exact pattern: inline, sheet, or stacked cards). |
| S-16 | Empty notes state is calm, not an error. |
| S-17 | Do not merge notes into `description`. Copy should distinguish catalog description vs research commentary. |

### 3.5 Metadata on detail

| ID | Requirement |
| --- | --- |
| S-18 | Show metadata fields **suggested for the Source’s type** as the primary form (labels from vocabulary, not keys). |
| S-19 | Text fields: single-line or short text as appropriate (`author`, `url`, `certificate_number`, …). |
| S-20 | Date fields: allow phrase-friendly entry; show that structured precision is optional enhancement, not a replacement for the wording. |
| S-21 | Leave a clear path for **fields beyond suggestions** (even if full “add field” UI is S2-04 — e.g. “More fields…” stub). |
| S-22 | Saving metadata should feel like part of Source editing, not a separate product. |

### 3.6 Artifacts placement (stub for S2-03)

| ID | Requirement |
| --- | --- |
| S-23 | Reserve a visible **Artifacts** section or tab on Source detail. |
| S-24 | For this board, Artifacts may be empty-state only (“No representations yet”) with a disabled or note CTA “Add in next board” — or a static mock list. Full ingest UX is S2-03. |

### 3.7 Errors and feedback

| ID | Requirement |
| --- | --- |
| S-25 | Validation failures (missing type, etc.) use calm inline or toast patterns consistent with onboarding (`Toast` danger/warning) — not modal panic. |
| S-26 | Busy/saving states should disable duplicate submits. |

---

## 4. Suggested dogfood seed (for realistic mock content)

Use these labels in mocks unless S2-04 changes them:

**Types:** Photograph, Book, Birth certificate  

**Example metadata labels:** Photographer, Taken date; Author, Publisher, Publication date; Jurisdiction, Certificate number, Record date  

---

## 5. Screen / frame inventory (minimum)

1. Sources list with several rows (mixed types, one untitled fallback).
2. Sources **empty** state.
3. **Create Source** flow (at least type + title).
4. Source detail — core fields + notes + metadata suggestions populated.
5. Source detail — empty notes + empty metadata suggestions.
6. Annotated region showing `SRC-…` treatment (mono, copyable look optional).

---

## 6. Out of scope

- File ingest, replace scan, thumbnails (S2-03).
- Creating custom types/fields (S2-04) — entry points OK.
- Credibility / proof standards UI.
- Interpretation (citations, people from this source).
- Bulk import.

---

## 7. Acceptance checklist

- [ ] Lives inside workspace content host; no second app chrome.
- [ ] List ↔ detail ↔ create covers the catalog loop.
- [ ] `SRC-…` is visible and non-editable.
- [ ] Notes are multi-entry; description is separate.
- [ ] Metadata follows type suggestions without requiring every field.
- [ ] Date fields preserve wording fidelity.
- [ ] No Interpretation or credibility controls.
- [ ] Artifacts section reserved but not fully designed.
