# S7-D2 — Subject fields (Properties + bindings)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PR **S7-05** only  
**Depends on:** Spike 5 seeded `subject_types` (fixed product set; **no** Subject types editor); S7-01 schema/seed; **S7-01b Property terms** (sixth `value_type = term`)  
**Related briefs:** archived [`S7-D1`](../archive/S7-D1-subject-types.md) (**descoped**); Source fields (S2-02) is **contrast only** — do not copy its layout; composer term picker lives in **S7-D4** (not this board)

**Already running a board from an older brief?** Paste the delta only: [`S7-D2-subject-fields-addendum-property-terms.md`](S7-D2-subject-fields-addendum-property-terms.md).

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](../README.md) first.

---

## 1. Objective

Design the **Subject fields** destination: configure `properties` and which of them bind to which **seeded** Subject types.

**Asymmetric scale (design around this):**

| Axis | Count | Implication |
| --- | --- | --- |
| **Subject types** | **Seven**, product-fixed (person, event, place, relationship, participation, location, source). Not researcher-extensible; plugins are hypothetical and far off. | Types can be **first-class chrome** — cards, a strip, segments, a small matrix — not a long scrollable type list. Treat “few types” as a creative opportunity. |
| **Properties** | **Many** (seeded + user; plan for ~20–40+) | Properties need findability and density. This is where Source fields’ tall narrow stack fails. |

**Do not reuse the Source fields UI.** Invent a new IA. Domain grammar (`label`, `key`, `origin`, `value_type`) may rhyme with Source fields; chrome must not.

**Encourage exploration.** Horizontal type cards, type-first vs property-first, filters-as-types, a bindings matrix, inspector patterns — all fair game. Prefer 2–3 distinct layout directions on the board before converging. Do not default to “another master–detail list” unless you justify it against the seven-type constraint.

This is config only — not the citation composer, not Observation entry, not a Subject types admin (descoped), and **not** an Event types / Roles vocabulary browser.

---

## 2. Domain facts the UI must reflect

Authoritative schema: interpretation-layer Properties / `subject_type_fields` / **`property_terms`**; seeds: [`seeded-vocabulary.md`](../../../../../seeded-vocabulary.md) §3.2–3.6. Direction: [`interpretation-graph-ui.md`](../../../../../ideas/archive/interpretation-graph-ui.md) decision **23**.

| Fact | UI implication |
| --- | --- |
| Seven fixed types | Show them as a closed, knowable set. Icons/colors may align with Evidence graph presentation tokens later; for this board, clear type identity is enough. |
| Many Properties | Search / filter / density required for the property side. |
| Bindings | Many-to-few: Properties ↔ the seven types. Locked/required bridge bindings from the Interpretation subject registry are not freely removable. |
| `value_type` | Product schema has six: `text`, `integer`, `date`, `name`, `subject`, **`term`**. **Create Property UI offers five** — never `term`. `term` Properties exist only via registry Install (product/plugin). Immutable after create. |
| **`term` Properties** | Kind/edge Properties (`event_type`, `role`, `relationship_type`) are registry-seeded and term-valued. Show them in the list/bindings like other seeded Properties. **Do not** let researchers create a Property with `value_type = term`. Term *values* (including user-minted rows) are managed in the **composer picker** (S7-D4), not as a Subject fields CatalogVocabulary. |
| Origin | Seeded (`provenencia`) vs user — visible without cluttering a dense property surface. |
| Avoid "claim" | **Subject fields** / Property language only. |

### 2.1 What this board is not

- Not a reskin of Source fields / CatalogVocabulary list panes.
- Not a scrollable **Subject types** vocabulary browser (types aren’t editable; there are only seven).
- Not Subject types CRUD — **descoped**.
- Not **Event types**, **Roles**, or **Relationship types** admin destinations — Property terms are not CatalogVocabulary places.
- Not the citation composer or NameValue editor — **S7-D4 / S7-D5**.
- Not Evidence graph cited rows — **S7-D3**.

### 2.2 Implementation gate (S7-05)

| Ships in S7-05 | Does **not** ship there |
| --- | --- |
| Create/edit Properties (**five** researcher value_types — no `term`); bind to the seven seeded types; show seeded `term` Properties read-only as to value_type | Citation composer; NameValue/DateValue editors; term-value CRUD; offering `term` in create Property |
| Layout from **this** board (whatever IA you choose) | Source-fields chrome; a types list that pretends types are a large catalog; Event types / Roles browsers |

---

## 3. Creative brief (authoritative)

**Constraint that should excite the layout:** seven types is small enough to **see all at once**. Forty Properties is not. Lead with that asymmetry.

Questions the board should answer (without prescribing answers):

1. Are **types** the primary navigation (pick a type card → see/bind its Properties), or are **Properties** primary (with types as chips/columns/bindings)?
2. Could types be a **horizontal card strip**, tab bar, or other non-list treatment?
3. How does someone **find** one Property among dozens once a type (or “all”) is in focus?
4. How are **locked** registry bindings shown without looking like broken checkboxes?
5. What does create Property feel like in this IA (five value types — **no** `term`)?
6. How are seeded **`term`** Properties shown so it’s clear they’re product vocabulary (not “broken” or missing an edit affordance for value type)?
7. Narrow workspace column: how do you use width for types vs depth for Properties?

Still mounts in the existing **Subject fields** sidebar destination — no new top-level section.

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SF-1 | Mount in **Subject fields**. **New UI** — not Source fields / CatalogVocabulary list chrome. |
| SF-2 | **Seven types** treated as a fixed, small set — visible together without a long type list. Explore non-list type chrome (e.g. horizontal cards). |
| SF-3 | **Properties** usable at ~30–40 rows: search and/or filter required; justify if omitted. |
| SF-4 | Researcher can create/edit a Property: label, description, key (slug), origin=user, value_type (**five** options: `text` / `integer` / `date` / `name` / `subject` — **not** `term`, immutable after create). |
| SF-5 | Researcher can see and change bindings to the seven types; **locked** bindings non-removable (show why). |
| SF-6 | Delete Property when unused; refuse while in use / locked. |
| SF-7 | Seeded vs user Properties distinguishable; seeded `term` Properties visible without inviting value_type edit. |
| SF-8 | Keyboard / VoiceOver complete for the chosen IA. |
| SF-9 | Board includes **at least two layout directions** explored; annotate why the chosen one fits “few types / many properties.” |
| SF-10 | Contrast frame vs Source fields (short stack) — optional but useful. |
| SF-11 | **No** Event types / Roles / Relationship-types destinations. **No** create-Property path for `value_type = term`. |

---

## 5. Screen / frame inventory (minimum)

1. Destination overview showing **all seven types** in the chosen type chrome + a dense Property surface.
2. Focused state (one type selected / filtered) with its Properties and bindings.
3. Search / filter among many Properties.
4. Create Property (**five** researcher value_types — no `term`).
5. Locked binding treatment on a bridge type (e.g. participation).
6. Optional: a seeded `term` Property in the list showing value type as product vocabulary (not creatable).
7. Alternate layout exploration (second direction) — even if not chosen.

---

## 6. Out of scope

- Composer / NameValue / graph cards.
- Designing the term picker **Add custom…** / rename / delete flow (S7-D4) — only acknowledge it exists.
- Conclusion `name_format`.
- Plugin type expansion UI (types stay seven for this product horizon).
- Cloning Source fields.
- Event types / Roles admin UIs.

---

## 7. Deliverable

Claude Design board + short notes for S7-05. Archive this brief when done.
