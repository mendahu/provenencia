# S7-D1 — Subject types (vocabulary editor)

> **Superseded / descoped.** Subject types are **product-seeded only** with first-class app plumbing (palette, cards, connect macros). They are not a user-extensible vocabulary like Source types. Do **not** implement a Subject types CatalogVocabulary editor. Spike 5 sidebar stub may remain; removing it from the rail is optional follow-on, not Spike 7. See [`../README.md`](../README.md) and the Spike 7 deployment plan *Descoped* section.

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** ~~PR **S7-04**~~ — **cancelled**  
**Depends on:** Spike 5 nested Sources config stubs (S5-D3); S7-01 schema/seed may land in parallel  
**Related briefs:** [`S7-D2`](S7-D2-subject-fields.md) — Subject fields; Source types pattern from Spike 2

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](../README.md) first.

---

## 1. Objective

Design the **Subject types** destination: replace the Spike 5 stub with a real CatalogVocabulary master–detail that mirrors **Source types**.

Researchers browse seeded and user-defined Subject types (`person`, `event`, `place`, bridges, `source`, plus custom), open a type to view/edit, and create new types with **ref prefixes**.

**Keep this board on this destination only.** Do not design Subject fields, the Evidence graph, or the citation composer.

---

## 2. Domain facts the UI must reflect

Authoritative schema: [`interpretation-layer-data-model.md`](../../../../../interpretation-layer-data-model.md); origin: [`seeded-vocabulary.md`](../../../../../seeded-vocabulary.md) §1.1 / §3.1.

| Fact | UI implication |
| --- | --- |
| `label`, `description`, `key`, `origin` | Same vocabulary grammar as Source types. Origin badge (provenencia / user / plugin). |
| `ref_prefix` + `candidate_ref_prefix` | Both required. Globally unique across origins and across both columns. Show on detail; validate on create/edit. |
| Seeded rows | Seven provenencia types already exist at project create. List them with system origin; deletable only when unused. |
| Subject ≠ Source | Copy and chrome must not say "claim." Engine words (`subject_types`) stay out of the rail. |
| Nested config | Destination already lives under Sources as de-emphasized config — do not redesign the sidebar hierarchy (S5-D3). |

### 2.1 What this board is not

- Not Subject fields / Properties / bindings — **S7-D2**.
- Not Evidence graph cards or Add property — **S7-D3**.
- Not the citation composer — **S7-D4**.
- Not Conclusion canonical entity UI.

### 2.2 Implementation gate (S7-04)

| Ships in S7-04 | Does **not** ship there |
| --- | --- |
| List + detail for Subject types | Subject fields editor |
| Create / edit / delete-when-unused | Observation or Citation UI |
| Origin + dual ref-prefix presentation | Canvas chrome |

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| ST-1 | Mount in the existing **Subject types** sidebar destination (stub replaced). Prefer existing CatalogVocabulary / `PVTable` patterns from Source types. |
| ST-2 | List shows all origins together; origin is visible (badge/column). |
| ST-3 | Detail shows label, description, key (read-only after create), origin, both ref prefixes, and usage/delete affordance when unused. |
| ST-4 | Create flow collects label (+ description); key auto-slugged; both prefixes required and validated; origin = user for researcher-created rows. |
| ST-5 | Seeded (`provenencia`) rows are editable only where product policy allows — propose: label/description editable, prefixes locked if that matches Source types; justify if different. |
| ST-6 | Delete refused while in use (subjects referencing the type); conflict messaging, not silent failure. |
| ST-7 | Accessibility: list and detail operable by VoiceOver/keyboard like Source types. |
| ST-8 | Empty / loading / error states consistent with CatalogVocabulary. |

---

## 4. Screen / frame inventory (minimum)

1. Subject types list (seeded + empty user set).
2. Detail of a seeded type (e.g. Person) showing both prefixes.
3. Create Subject type form (prefixes prominent).
4. Delete blocked / in-use state.
5. Annotation: how this mirrors Source types without looking identical (Subject vs Source vocabulary).

---

## 5. Out of scope

- Binding Properties to types (S7-D2).
- Icons per Subject type on the canvas (already Spike 6).
- Search registry / omnibar hit kinds for new types (follow-on unless trivial).

---

## 6. Deliverable

Claude Design board + short notes for S7-04. Archive this brief when done.
