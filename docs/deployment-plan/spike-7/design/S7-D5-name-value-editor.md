# S7-D5 — NameValue editor (reusable modal)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PR **S7-02b** only (schema + Go are **S7-02**, ungated)  
**Depends on:** [`structured-name-model.md`](../../../../structured-name-model.md) §1–3; shipped DateValue editor as the twin pattern (`Features/Dates/`); thin composer (**S7-08**) already dogfoodable  
**Related briefs:** [`S7-D4`](S7-D4-citation-composer.md) — composer hosts this modal for `value_type = name`; [`S7-D2`](S7-D2-subject-fields.md) — researchers create `name`-typed Properties (none in create-time seed) 
**Schedule:** Late — after Add property + thin composer work; not on the path to S7-05 or first cite dogfood.

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design the reusable **NameValue** editor as its own surface — a modal/sheet parallel to **DateValue**, not as chrome invented inside the citation composer board.

```text
NameValue
  form          ← required full-form normalized reading
  parts[]       ← optional ordered parts (value + open type)
```

The composer (and any later host) opens this editor when an Observation’s Property has `value_type = name`. Transcription of what the Source said stays on the **Citation**; this editor is the Interpretation normalization.

**Keep this board small.** Do not design the composer place, Subject fields admin, or Conclusion `name_format` profiles.

---

## 2. Domain facts the UI must reflect

Authoritative schema: [`structured-name-model.md`](../../../../structured-name-model.md). Starter part types: [`seeded-vocabulary.md`](../../../../seeded-vocabulary.md) §4.

| Fact | UI implication |
| --- | --- |
| `form` is required | Primary field; user can save with form alone and empty parts. |
| Parts are optional and ordered | Add/reorder/remove part rows; each has `value` + open `type` (starter: prefix, given, surname, …). |
| Not Western-only | Do not force given/surname; culture-agnostic part types; form can stand alone for names that do not segment cleanly. |
| Value object, not a subject | No ref, no canvas card — just a structured value draft. |
| Twin to DateValue | Same product grammar: nested modal from a form field, confirm/cancel, validation before accept. |
| Transcription ≠ NameValue | Copy should not invite pasting the Citation transcription as the only mental model — form is normalized reading. |

### 2.1 What this board is not

- Not the citation composer place — **S7-D4** (only a host that opens this modal).
- Not Subject fields vocabulary — **S7-D2**.
- Not Evidence graph cards — **S7-D3**.
- Not Conclusion `name_format` / display profiles.
- Not search / FTS projection for names (engine follow-on).

### 2.2 Implementation gate (S7-02b)

| Ships in S7-02b | Does **not** ship there |
| --- | --- |
| `NameValueDraft` + editor modal UI | Wiring into composer Observation rows (**S7-02b** after thin **S7-08**) |
| Part list add/reorder/remove | `name_format` profiles |
| Validation UX for empty form / empty part values | Schema/Go (S7-02) |

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| NV-1 | Modal/sheet editor reusable from any host (composer first). Prefer patterns from `DateValueEditorForm` / Dates feature. |
| NV-2 | **Form** field required; clear empty/error state. |
| NV-3 | **Parts** list: add, remove, reorder; each row value + type (open text with starter suggestions OK). |
| NV-4 | Confirm returns a draft the host can store; Cancel discards. |
| NV-5 | Propose how a collapsed/summary control looks when embedded in an Observation row (form preview, “Edit name…”). |
| NV-6 | Accessibility: full keyboard path; VoiceOver names for form, parts, add/remove/reorder. |
| NV-7 | Localization via `L10n` — no hard-coded English in the design handoff notes for implementers. |

---

## 4. Screen / frame inventory (minimum)

1. Empty NameValue editor (form only).
2. Form + three parts (e.g. given / initial / surname).
3. Form-only culturally unsegmented name (no parts).
4. Validation: missing form.
5. Embedded summary control as it would appear in an Observation row (annotation OK — full composer is S7-D4).
6. Side annotation vs DateValue: what is shared chrome vs name-specific.

---

## 5. Out of scope

- Composer layout / breadcrumbs (S7-D4).
- Part-type vocabulary admin table (open text + starters only).
- Conclusion name format profiles.

---

## 6. Deliverable

Claude Design board + short notes for S7-02b. Archive this brief when done.
