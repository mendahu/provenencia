# S7-D2 — Subject fields (Properties + bindings)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PR **S7-05** only  
**Depends on:** Spike 5 seeded `subject_types` (listable for bindings; **no** Subject types editor); S7-01 schema/seed  
**Related briefs:** archived [`S7-D1`](archive/S7-D1-subject-types.md) (**descoped**); Source fields (S2-02) is **contrast only** — do not copy its layout

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design the **Subject fields** destination: the product name for `properties` plus `subject_type_fields` bindings. Researchers browse Properties, create/edit them, and bind which Properties apply to which **seeded** Subject types (person, event, place, bridges, source — fixed product set).

**Do not reuse the Source fields UI as the template.** Source fields live in a narrow sidebar destination where a short list (often ~1–5 fields) stacks fine. Subject fields may grow to **dozens** of Properties (20–40+) as researchers track more assertions. The Source fields master–detail / tall narrow list will fill and thrash. This board must invent a **new IA / layout** that stays usable at that scale (search, filter, density, and how detail/bindings appear).

Domain grammar can still rhyme with Source fields (`label`, `key`, `origin`, typed value) — that is data shape, not chrome.

This is config for what can be asserted on a subject — not the citation composer and not Observation value entry on the graph. **There is no Subject types admin** in this spike.

---

## 2. Domain facts the UI must reflect

Authoritative schema: interpretation-layer Properties / `subject_type_fields`; seeds: [`seeded-vocabulary.md`](../../../../seeded-vocabulary.md) §3.2–3.3.

| Fact | UI implication |
| --- | --- |
| Scale | Expect **many** Properties. Design for scan + find, not a short stacked list that works for Source fields. |
| Property `label`, `key`, `origin`, `description` | Same vocabulary grammar as Source fields (data), different presentation. |
| `value_type` | **Exactly five:** `text`, `integer`, `date`, `name`, `subject`. Chosen at create; **immutable** afterward. No `real` / `boolean`. |
| Bindings | `subject_type_fields` joins Subject type ↔ Property. Subject types are a **fixed seeded picker**, not a types editor. **Required/locked** seeded bindings (bridge edges) come from the Interpretation subject registry — not freely deletable. |
| Seeded Properties | Install at project create; list with system origin. |
| Subject-valued Properties | `person`, `event`, `place`, `participant` are edges; UI may hint target kind in copy. |
| Avoid "claim" | Use **Subject fields** / Property language. |

### 2.1 What this board is not

- Not a reskin of Source fields / CatalogVocabulary list panes — **explicitly out**.
- Not Subject types CRUD — **descoped** (S7-D1 archived).
- Not the Observation form inside the composer — **S7-D4**.
- Not Evidence graph cited rows — **S7-D3**.
- Not open-value picker admin for `event_type` / `role` (composer starters).

### 2.2 Implementation gate (S7-05)

| Ships in S7-05 | Does **not** ship there |
| --- | --- |
| Properties browse/create/edit at scale | Citation composer |
| Bind Properties to **seeded** Subject types | NameValue / DateValue editors (**S7-D5** / Dates) |
| Five value_type picker on create | Graph Add property chrome; Subject types destination |
| Layout from **this** board | Copying Source fields chrome |

---

## 3. Layout problem (authoritative)

The workspace content column for nested config is relatively **narrow**. Source fields tolerate that because the list is short. Subject fields will not.

Propose an IA that answers:

1. How does a researcher **find** one Property among dozens (search, filters by origin / value_type / bound Subject type, grouping)?
2. Where does **detail + bindings** live without forcing endless vertical scroll of the whole catalog (inspector, split that uses width differently, table with row expansion, type-centric vs property-centric view, …)?
3. How do **locked** registry bindings stay visible without dominating a long list?
4. What does **empty / first-run** look like when seeds already populate many rows?

Feel free to use more of the destination’s width, alternate density, or a different split than Source fields — still inside the existing Subject fields sidebar destination (do not invent a new top-level section).

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SF-1 | Mount in existing **Subject fields** destination. **New UI** — do not reuse Source fields / CatalogVocabulary list layout as the default. |
| SF-2 | Scale: usable with **~30–40** Properties (seeded + user). Search and/or filter required; justify if omitted. |
| SF-3 | List/browse shows enough to discriminate rows: label, `value_type`, origin (at minimum). |
| SF-4 | Detail: label, description, key, origin, value_type (read-only after create), bindings to Subject types. |
| SF-5 | Create: label → key slug; **value_type** picker with only the five types; origin = user. |
| SF-6 | Bind / unbind to **seeded** Subject types. **Locked/required** registry bindings non-removable (show why). |
| SF-7 | Delete Property when unused; refuse while Observations or bindings require it. Locked provenencia edge Properties refuse the same way. |
| SF-8 | Seeded Properties discoverable among a long list (origin facet or equivalent). |
| SF-9 | Full keyboard / VoiceOver path for browse, open, edit, bind — scale must not break a11y. |
| SF-10 | Annotate **why this is not Source fields**: show a frame contrasting short Source-fields stack vs this large-list IA. |

---

## 5. Screen / frame inventory (minimum)

1. Full Subject fields destination with a **dense** seeded list (~dozens of rows visible as a problem to solve).
2. Search / filter in use (narrowed list).
3. Detail of `name` (value_type = name) with person binding + locked binding treatment if applicable.
4. Create Property — five value_types only.
5. Binding UI at scale (many Properties, few Subject types).
6. Side-by-side or annotated contrast: Source fields (short) vs Subject fields (long) — why the new IA wins.

---

## 6. Out of scope

- Composer Observation rows and typed editors (S7-D4).
- Conclusion Properties / name_format.
- Plugin origin management UI.
- Cloning Source fields chrome.

---

## 7. Deliverable

Claude Design board + short notes for S7-05. Archive this brief when done.
