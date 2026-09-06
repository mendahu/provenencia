# S2-04 — Extensible Source types and metadata fields

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 2 (Source layer validation)  
**Implements later as:** PR S2-18 (Swift UI); seed choices feed PR S2-07  
**Depends on:** S2-02 Source catalog (uses types/fields while editing Sources)  
**Related briefs:** S2-01 (prefer not adding a top-level nav item unless justified here)

Paste this entire document into Claude Design as the requirements for one board/flow.

---

## 1. Objective

Design how researchers **extend the project’s catalog vocabulary**: add Source types, add metadata fields (`text` / `date`), attach suggested fields to a type, and use custom vocabulary while creating/editing Sources. Builtins are convenient defaults, **not** privileged subclasses.

Also **lock the dogfood seed set** (which builtin types/fields appear in mocks and first engineering seed).

---

## 2. Domain model (UI must reflect)

### 2.1 `source_types`

| Column / rule | UI implication |
| --- | --- |
| `key` | Stable machine id (`photograph`); show rarely (advanced); **label** is primary. |
| `label` | What pickers and chips show. |
| `description` | Optional help text in admin UI. |
| `builtin` | Builtin vs project-custom **affordance only** (badge/caption). Same edit rules for using them on Sources; builtins should not be deletable (or delete is refused). |
| Not an enum | Users can add types without engineering. |

### 2.2 `source_metadata_fields`

| Column / rule | UI implication |
| --- | --- |
| `key` / `label` | Same as types — label in forms; key secondary. |
| `data_type` | Closed for Spike 2: **`text`** or **`date`** only. Choosing type is part of create-field UX. |
| `builtin` | Same affordance rules as types. |
| Not a general EAV playground | Do not design arbitrary “number / boolean / person-link” field types. |

### 2.3 `source_type_metadata_fields`

| Rule | UI implication |
| --- | --- |
| Type **suggests** fields with optional `sort_order` | Admin UI: attach/reorder suggested fields for a type. |
| Suggestions are not validation | Source forms never block save because a suggested field is empty. |
| Source may use non-suggested fields | Editing a Source can include values outside the suggestion list (S2-02 “more fields”). |

### 2.4 Product principles (vocabulary doc)

1. **Implement small; grow from use** — do not dump the entire horizon catalog into the first seed.
2. Seeds are data, not SQL enums.
3. Do not invent Source-quality defect ontologies or credibility grades in this UI.

---

## 3. Requirements

### 3.1 Information architecture

| ID | Requirement |
| --- | --- |
| V-1 | Decide where vocabulary admin lives. **Default:** sheet, secondary screen, or settings subsection **reached from Sources** (e.g. “Manage types & fields”) — **not** a new top-level sidebar item unless the board argues strongly. |
| V-2 | Document that decision for S2-01/S2-14 implementers. |
| V-3 | Using a type/field on a Source stays in the S2-02 create/edit flows; admin is for defining vocabulary. |

### 3.2 Manage Source types

| ID | Requirement |
| --- | --- |
| V-4 | List existing types with **label**, builtin vs custom affordance, and optional description. |
| V-5 | Create type: require **label**; `key` may be auto-derived from label (show advanced override only if needed). |
| V-6 | Builtin types: visible; **not deletable** (disable delete or explain refusal). |
| V-7 | Custom types: editable label/description; delete only if safe — if delete rules are unclear, hide delete in Spike 2 and note “omit delete.” |
| V-8 | Newly created types appear in the Source type picker (show a frame proving round-trip). |

### 3.3 Manage metadata fields

| ID | Requirement |
| --- | --- |
| V-9 | List fields with label, **text/date** type, builtin vs custom. |
| V-10 | Create field: label + data type (`text` \| `date`). |
| V-11 | Builtin fields not deletable (same as types). |
| V-12 | Do not offer field data types beyond text/date. |

### 3.4 Suggest fields for a type

| ID | Requirement |
| --- | --- |
| V-13 | For a selected type, show suggested fields in order; support add/remove suggestion and reorder (or explicit sort). |
| V-14 | Suggesting a field does not create a value on existing Sources until the user fills it. |
| V-15 | Frame: Source edit form for that type showing suggestions reflecting the admin configuration. |

### 3.5 Using non-suggested fields on a Source

| ID | Requirement |
| --- | --- |
| V-16 | From Source edit (S2-02 surface), user can add a metadata value for a field **not** in the type’s suggestions (picker of existing fields). |
| V-17 | If the field does not exist yet, user may jump to create-field (or combined flow) — keep it light. |

### 3.6 Dogfood seed decision (required output)

| ID | Requirement |
| --- | --- |
| V-18 | Board (or caption) states the **seed set** for engineering S2-07. |
| V-19 | **Proposed default** (amend freely, but decide): |

**Types**

- Photograph (`photograph`)
- Book (`book`)
- Birth certificate (`birth_certificate`)

**Fields** (minimum to make those types feel real — trim or extend slightly if needed)

| Field label | key (suggested) | data_type | Suggested on |
| --- | --- | --- | --- |
| Photographer | `photographer` | text | photograph |
| Taken date | `taken_date` | date | photograph |
| Medium | `medium` | text | photograph |
| Author | `author` | text | book |
| Publisher | `publisher` | text | book |
| Publication date | `publication_date` | date | book |
| Jurisdiction | `jurisdiction` | text | birth_certificate |
| Certificate number | `certificate_number` | text | birth_certificate |
| Record date | `record_date` | date | birth_certificate |
| Repository | `repository` | text | birth_certificate, book (optional) |

| ID | Requirement |
| --- | --- |
| V-20 | Do **not** seed the full horizon list from the vocabulary doc (census ED fields, DNA kit numbers, etc.) unless dogfood explicitly needs them. |

---

## 4. Screen / frame inventory (minimum)

1. Entry point from Sources → vocabulary admin.
2. Types list (builtin + one custom).
3. Create custom type.
4. Fields list + create field (`text` and `date` examples).
5. Type → suggested fields editor (reorder/add).
6. Source create/edit using a **custom** type and a **custom** field.
7. Source edit adding a **non-suggested** existing field.
8. Seed set callout (table or notes panel) for engineering handoff.

---

## 5. Out of scope

- Interpretation node types / properties.
- Credibility grades, confidence grades.
- Arbitrary new `data_type`s.
- Syncing vocabulary across projects.
- Translating researcher-defined labels (still plain project data).

---

## 6. Acceptance checklist

- [ ] IA decision recorded (Sources subflow vs sidebar).
- [ ] Builtin vs custom is visible but not a second class of behavior for *using* vocabulary.
- [ ] Create type + create field + suggest fields covered.
- [ ] Only `text` and `date` field types.
- [ ] Round-trip: custom vocabulary appears on Source forms.
- [ ] Non-suggested field use is possible.
- [ ] Dogfood seed set explicitly decided for S2-07.
- [ ] Visual language matches workspace / onboarding.
