# S7-D2 — Subject fields (Properties + bindings)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PR **S7-05** only  
**Depends on:** Subject types destination chrome (S7-D1 / S7-04 may land first); S7-01 schema/seed  
**Related briefs:** [`S7-D1`](S7-D1-subject-types.md); Source fields (S2-02)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design the **Subject fields** destination: the product name for `properties` plus `subject_type_fields` bindings. Mirror **Source fields** / type↔field suggestions: browse Properties, create/edit them, and bind which Properties apply to which Subject types.

This is config for what can be asserted on a subject — not the citation composer and not Observation value entry on the graph.

---

## 2. Domain facts the UI must reflect

Authoritative schema: interpretation-layer Properties / `subject_type_fields`; seeds: [`seeded-vocabulary.md`](../../../../seeded-vocabulary.md) §3.2–3.3.

| Fact | UI implication |
| --- | --- |
| Property `label`, `key`, `origin`, `description` | Same vocabulary grammar as Source fields. |
| `value_type` | **Exactly five:** `text`, `integer`, `date`, `name`, `subject`. Chosen at create; **immutable** afterward. No `real` / `boolean`. |
| Bindings | `subject_type_fields` joins Subject type ↔ Property (like Source type ↔ metadata field suggestions). |
| Seeded Properties | Install at project create; list with system origin. |
| Subject-valued Properties | `person`, `event`, `place`, `participant` are edges; UI may hint target kind in copy, but no SQL allow-list chrome required for v1. |
| Avoid "claim" | Use **Subject fields** / Property language. |

### 2.1 What this board is not

- Not Subject types CRUD — **S7-D1**.
- Not the Observation form inside the composer — **S7-D4** (value editors live there).
- Not Evidence graph cited rows — **S7-D3**.
- Not open-value picker lists for `event_type` / `role` as separate admin (those are free-text starters in the composer).

### 2.2 Implementation gate (S7-05)

| Ships in S7-05 | Does **not** ship there |
| --- | --- |
| Properties list/detail/create | Citation composer |
| Bind Properties to Subject types | NameValue / DateValue editors (reuse in composer) |
| Five value_type picker on create | Graph Add property chrome |

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| SF-1 | Mount in existing **Subject fields** destination; reuse CatalogVocabulary patterns from Source fields. |
| SF-2 | List Properties across origins; show `value_type` and origin. |
| SF-3 | Detail: label, description, key, origin, value_type (read-only after create), bindings to Subject types. |
| SF-4 | Create: label → key slug; **value_type** picker with only the five types; origin = user. |
| SF-5 | Bind / unbind Properties to Subject types (propose UX: on Property detail, on Subject type detail, or both — pick one primary). |
| SF-6 | Delete Property when unused; refuse while Observations or bindings require it (propose clear conflict copy). |
| SF-7 | Seeded Properties discoverable; first-class keys (`name`, `birth_date`, …) need no special chrome beyond origin. |
| SF-8 | Accessibility parity with Source fields. |

---

## 4. Screen / frame inventory (minimum)

1. Subject fields list (seeded Properties visible).
2. Detail of `name` (value_type = name) with person binding.
3. Create Property — value_type picker showing five options only.
4. Binding UI (Property ↔ Subject type).
5. Annotation vs Source fields: what is the same pattern, what differs (`value_type` vs `data_type`, bindings shape).

---

## 5. Out of scope

- Composer Observation rows and typed editors (S7-D4).
- Conclusion Properties / name_format.
- Plugin origin management UI.

---

## 6. Deliverable

Claude Design board + short notes for S7-05. Archive this brief when done.
