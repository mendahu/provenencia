# S7-D5 — NameValue editor (reusable modal)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PR **S7-02b** only (schema + Go are **S7-02**, ungated; part-type registry hardened in Go alongside S7-02b)  
**Depends on:** [`structured-name-model.md`](../../../../structured-name-model.md) §1–3; shipped DateValue editor as the twin pattern (`Features/Dates/`); thin composer (**S7-08**) already dogfoodable; product part-type registry in `core/database/namevalues`  
**Related briefs:** [`S7-D4`](archive/S7-D4-citation-composer.md) — composer hosts this modal for `value_type = name`; [`S7-D2`](archive/S7-D2-subject-fields.md) — Property `name` exists in vocabulary 
**Schedule:** Late — after Add property + thin composer work; not on the path to S7-05 or first cite dogfood.

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design the reusable **NameValue** editor as its own surface — a modal/sheet parallel to **DateValue**, not as chrome invented inside the citation composer board.

```text
NameValue
  form          ← required full-form normalized reading
  parts[]       ← optional ordered parts (value + product part-type key)
```

The composer (and any later host) opens this editor when an Observation’s Property has `value_type = name`. Transcription of what the Source said stays on the **Citation**; this editor is the Interpretation normalization.

**Keep this board small.** Do not design the composer place, Subject fields admin, or Conclusion `name_format` profiles.

---

## 2. Domain facts the UI must reflect

Authoritative schema: [`structured-name-model.md`](../../../../structured-name-model.md). Product part types: [`seeded-vocabulary.md`](../../../../seeded-vocabulary.md) §4.1 + `namevalues.PartTypes()`.

| Fact | UI implication |
| --- | --- |
| `form` is required | Primary field; user can save with form alone and empty parts. |
| Parts are optional and ordered | Add/reorder/remove part rows; each has `value` + **type picker** from the product registry (prefix, given, surname, …) with **localized labels**. |
| Type is first-class | Do **not** use free-text type entry. Empty / omit type only as an untyped-segment escape (or use `undetermined`). Unknown keys are invalid. |
| Not Western-only | Do not force given/surname rows; form can stand alone for names that do not segment cleanly. Registry keys stay culture-agnostic. |
| Value object, not a subject | No ref, no canvas card — just a structured value draft. |
| Twin to DateValue | Same product grammar: nested modal from a form field, confirm/cancel, validation before accept. |
| Transcription ≠ NameValue | Copy should not invite pasting the Citation transcription as the only mental model — form is normalized reading. |

### 2.1 What this board is not

- Not the citation composer place — **S7-D4** (only a host that opens this modal).
- Not Subject fields vocabulary — **S7-D2**.
- Not Evidence graph cards — **S7-D3**.
- Not Conclusion `name_format` / display profiles.
- Not search / FTS projection for names (engine follow-on).
- Not user-minted part-type admin (Property-terms-shaped table) — deferred; product registry only.

### 2.2 Implementation gate (S7-02b)

| Ships in S7-02b | Does **not** ship there |
| --- | --- |
| `NameValueDraft` + editor modal UI | User-minted `name_part_types` catalog table |
| Part list add/reorder/remove; type **picker** from Go/Swift registry + L10n labels | `name_format` profiles |
| Validation UX for empty form / empty part values / unknown type | Schema DDL change (column stays TEXT) |
| Wire into composer Observation rows | — |

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| NV-1 | Modal/sheet editor reusable from any host (composer first). Prefer patterns from `DateValueEditorForm` / Dates feature. |
| NV-2 | **Form** field required; clear empty/error state. |
| NV-3 | **Parts** list: add, remove, reorder; each row value + **type picker** (product registry keys; localized labels via `L10n`). No free-text type field. |
| NV-4 | Confirm returns a draft the host can store; Cancel discards. |
| NV-5 | Propose how a collapsed/summary control looks when embedded in an Observation row (form preview, “Edit name…”). |
| NV-6 | Accessibility: full keyboard path; VoiceOver names for form, parts, type picker, add/remove/reorder. |
| NV-7 | Localization via `L10n` — part-type labels from registry `L10nKey`s; no hard-coded English type names in the design handoff. |

---

## 4. Screen / frame inventory (minimum)

1. Empty NameValue editor (form only).
2. Form + three parts (e.g. given / initial / surname) with type pickers.
3. Form-only culturally unsegmented name (no parts).
4. Validation: missing form.
5. Embedded summary control as it would appear in an Observation row (annotation OK — full composer is S7-D4).
6. Side annotation vs DateValue: what is shared chrome vs name-specific.

---

## 5. Out of scope

- Composer layout / breadcrumbs (S7-D4).
- Part-type vocabulary **admin** / user-minted types (future extensibility PR).
- Conclusion name format profiles.

---

## 6. Deliverable

Claude Design board + short notes for S7-02b. Archive this brief when done.
