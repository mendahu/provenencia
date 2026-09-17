# Provenencia Genealogy — Narrative Layer Data Model

## Status

**Exploratory draft — not roadmapped, not shipped.** This document captures architecture discussion (September 2026) about a possible fourth research layer. Table shapes, ref prefixes, and product scope here are **proposals for stewing**, not commitments. No SQLite migrations, FFI handlers, or UI specs exist for this layer yet.

When this idea is promoted, update [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md) and sibling layer docs so they do not compete on schema ownership.

The Narrative layer would answer:

> How does the researcher choose to **arrange, emphasize, and communicate** what they believe — in prose, diagrams, and interactive views?

Cross-layer philosophy: [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md). Factual substrate: Source → Interpretation → Conclusion. Audit history: [`audit-revision-history.md`](audit-revision-history.md). Ref minting rules: [`catalog-refs.md`](catalog-refs.md).

---

# 1. Why a fourth layer

Genealogy produces at least two distinct outputs:

1. **Structured belief** — who existed, how they relate, what happened when and where, grounded in evidence where possible (Source → Interpretation → Conclusion).
2. **Told history** — prose, emphasis, omission, visualization, and synthesis that turn facts into something readable and meaningful.

The existing Person / Event / Place vocabulary fits the **family tree** and proof-oriented workflows cleanly. It fits **expressive** material awkwardly: census household amenities, oral-history color (“they loved the beach on Sundays”), pets, generational “hub” places, migration stories drawn on a map. Those details are often real and citable, but they are not timeless Person attributes and they are not always worth normalizing into Observations.

Rather than stretching Conclusion Properties to carry everything, Narrative gives **composition and presentation** a documented home downstream of the graph.

---

# 2. The bias gradient

Each layer adds legitimate researcher interpretation:

| Layer | Question | Bias introduced (examples) |
| --- | --- | --- |
| **Source** | What evidence do we possess? | What was kept, titled, cataloged |
| **Interpretation** | What does this source appear to say? | Reading, transcription, normalization |
| **Conclusion** | What do I currently conclude? | Identity commitment, reconciliation, merge |
| **Narrative** | What story am I telling, and how am I showing it? | Arrangement, emphasis, visualization, prose |

Narrative is the most subjective layer by design. That is not a flaw — it is the point — as long as it **references** lower layers rather than silently rewriting them.

---

# 3. Layer boundaries

## 3.1 One-way flow

```text
Source → Interpretation → Conclusion → Narrative
                              ↑              │
                              └─ references ─┘
```

Narrative **points at** catalog and graph rows. It does **not** automatically promote prose or layout choices into Observations, Reconciliations, or Source metadata.

A sentence in an essay does not create an Observation. A “hub” highlight on a map does not set a Place Property. A tree layout does not create parent/child Claims. When research while writing leads to new evidence or conclusions, the researcher follows the normal lower-layer workflows explicitly.

## 3.2 Not Source, not Interpretation, not Conclusion

| Layer | Why Narrative is different |
| --- | --- |
| **Source** | Researcher-authored composition is output, not acquired evidence. (An externally authored book ingested as a Source is still Source — your own essay while researching is Narrative.) |
| **Interpretation** | No Observation shape; prose is not an atomic cited assertion backed by one Citation per sentence. |
| **Conclusion** | No Sameness or Reconciliation Claims; narrative does not commit historical truth to the working tree. |

Existing **notes** (`source_notes`, `citation_notes`, `observation_notes`, `canonical_entity_notes`) and Claim **`argument`** fields remain scoped commentary on a single object or decision. Narrative is **cross-cutting composition** at a different grain.

## 3.3 Rigorous science and expressive storytelling

The product stance in [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md) §1.0 still applies: good science should be convenient; a casual tree is valid. Narrative extends that — **expressive storytelling should also be convenient** without corrupting the fact layers.

Encourage links from prose to Observations; badge unlinked narrative claims; never block saving a story because the user skipped citations.

---

# 4. Projections versus compositions

Narrative contains two related ideas:

## 4.1 Projections (derived views)

**Projections** render Conclusion (+ Interpretation for drill-down) with **no persisted Narrative row** until the researcher customizes something.

Examples:

- Default **family tree** from parent/child and spouse relationships
- **Timeline** from dated Events
- **Map** pins from Locations
- Person list sorted by name

These are Narrative in the **epistemological** sense: layout algorithms, default filters, and “which spouse displays left” are presentation choices. They may persist **zero** catalog state for a new project.

**Tree edges are not Narrative data.** Parent/child and unions live in Conclusion (or are projected from event participations). The tree **reads** them; it does not store them.

## 4.2 Compositions (authored artifacts)

**Compositions** persist researcher intent beyond defaults:

- Prose essays, chapters, person biographies
- Saved tree views (focus ancestor, collapsed branches, print layout)
- Curated **migration maps**, regional hub highlights, time sliders
- Annotated presentations for export or share packages

Compositions are first-class catalog objects (proposed ref prefix `NAR-…`) with audit history and **references** into the graph.

| Kind | Persisted? | Example |
| --- | --- | --- |
| Default tree / timeline / map | Usually no | Open workspace → see tree |
| Light UI prefs | Maybe session/local | Last focus person |
| Prose narrative | Yes | `NAR-…` “The Robins migration” |
| Saved visual composition | Yes | `NAR-…` map with hub annotations |
| Tree relationships | **No** | Always from Conclusion |

---

# 5. Family tree placement

The family tree sits at the **low-bias** end of Narrative:

```text
Conclusion     relationships, identities, reconciled dates/places
                    ↓ read
Narrative      default tree projection (often no rows)
               saved tree composition (when customized)
```

**Conclusion owns:** who is related to whom, sameness, committed Property values, existence handles.

**Narrative owns:** how that structure is **shown** — root person, inclusion filters, collapse state, emphasis, labels for export, optional prose under subjects.

A researcher who only builds Conclusion still gets a tree “for free” as a projection. That is Narrative-layer behavior with minimal authorship.

---

# 6. Maps, migration, and generational hubs

Geographic visualization follows the same split.

**Lower layers (factual substrate):**

- Places with toponyms and coordinates when known
- Events with structured dates
- Locations linking events to places
- Migration as a **sequence of dated Locations**, not one drawn arrow

**Narrative (meaning and emphasis):**

- “These four towns were generational hubs for our family” — synthesis across decades
- Migration **paths** as drawn routes (which moves count, ordering, smoothing)
- Regional shading, time sliders, color scales
- Curated “important areas” that are not the same as “person lived here 1882–1901”

A **hub** is usually **framing**, not a reconcilable Place fact. Prefer a map composition that **references** canonical Places and Events rather than a `is_hub` Reconciliation on Place.

Example composition (conceptual):

```text
Narrative N1 (kind = map_composition)
  title: "Robins hubs across Canada"
  config: time range, layer visibility, path ordering
  references:
    → canonical Places (Stratford, Winnipeg, …)
    → canonical Events (immigration, land grant, …)
    → optional Observations for cited “why we moved”
  annotations:
    → region highlight "Prairie settlement era"
```

---

# 7. Prose narratives and references

## 7.1 Authoring model

Prose compositions are **freeform text** (Markdown or rich text — product choice) plus **references** to lower-layer rows.

The researcher writes; they **tag** spans or insert links:

- **Canonical entity** — “William Smith” → `PER-…`
- **Observation** — ground a sentence in a specific cited assertion → `OBS-…`
- **Citation** — point at the evidence slice → `CIT-…`
- **Source** — catalog-level pointer → `SRC-…`

Linking to an **Observation** is the finest bridge between story and proof: “the 1931 census shows they had a radio” attaches to the exact normalized reading, not only the person.

## 7.2 Inverse of Citation locators

Citations use locators (for example `text_quote`) to point **into** an Artifact. Narrative references point **out from** researcher prose **to** graph entities — same “select and attach” gesture, opposite direction.

Stored anchors should survive edits better than raw character offsets alone (marker tokens, block IDs, or embedded link subjects in a structured document — product decision).

## 7.3 Multiple narratives, conflicting stories

Two compositions may describe the same period differently. Both can reference the same Observations with different prose. Conclusion Reconciliation picks committed **facts**; Narrative does not arbitrate **storytelling**.

When canonical entities merge, references resolve by machine id; UI may show surviving `ref` and stale-link warnings.

---

# 8. Worked examples (where layers divide)

Discussion examples that motivated this draft:

## 8.1 Census household radio (1931)

| Layer | Home |
| --- | --- |
| Source | Census catalog metadata (sheet, dwelling number, …) |
| Interpretation | Event subject (`event_type = census`, date, `has_radio → true` Observation) |
| Conclusion | Optional Reconciliation on canonical **Event**, not Person |
| Narrative | “By 1931 the household had a radio” in prose, linked to `OBS-…` |

Time-binding stays on the census **event**, not a timeless Person attribute.

## 8.2 “Enjoyed the beach on Sundays”

| Layer | Home |
| --- | --- |
| Source | Oral testimony / interview Artifact |
| Interpretation | Citation transcription (required); optional Relationship + text Observation |
| Conclusion | Usually none |
| Narrative | The telling — Grandma’s voice, Sunday routine, meaning |

Prefer **Citation-first** capture; compose in Narrative rather than forcing a Person Property.

## 8.3 Family pets

| Layer | Home |
| --- | --- |
| Interpretation | Optional `animal` subject + Relationship when named/recurring; vague mention → text only |
| Conclusion | Light touch; avoid reconciling `species` onto Person |
| Narrative | “Rex was part of the family” in biography; link Animal subject or photo Event |

Tier by evidence: vague mention → Narrative + Citation; named pet across sources → Interpretation graph + Narrative links.

---

# 9. Proposed schema sketch

**Draft only.** Names and columns will change if this layer ships.

## 9.1 `narratives`

One table for prose and visual compositions; discriminate by `kind`.

```sql
CREATE TABLE narratives (
    id              BLOB PRIMARY KEY,          -- UUIDv7
    ref             TEXT UNIQUE NOT NULL,      -- e.g. NAR-3K9M2 (proposed prefix)
    kind            TEXT NOT NULL,             -- prose | tree_view | map_composition | timeline_view | …
    title           TEXT NOT NULL,
    body            TEXT,                      -- prose / Markdown; NULL for config-only kinds
    config_json     TEXT,                      -- layout, filters, map layers, focus entity ids, …
    primary_entity_id BLOB REFERENCES canonical_entities(id) ON DELETE SET NULL,
    CHECK (trim(title) != '')
) STRICT;
```

`primary_entity_id` is optional scope (for example a person-centric biography). Cross-cutting essays leave it null.

`body` and `config_json` are mutually exclusive by convention for some kinds; application validation may require at least one.

## 9.2 `narrative_references`

Typed outbound links — not a polymorphic “link anything” JSON blob without FKs.

```sql
CREATE TABLE narrative_references (
    id              BLOB PRIMARY KEY,
    narrative_id    BLOB NOT NULL REFERENCES narratives(id) ON DELETE CASCADE,
    anchor_json     TEXT NOT NULL,             -- span, block id, or footnote key — versioned schema
    target_kind     TEXT NOT NULL,             -- canonical_entity | observation | citation | source
    target_id       BLOB NOT NULL,             -- FK enforced per kind in application layer or split tables
    label           TEXT                       -- optional display override
) STRICT;
```

Alternative later: separate nullable FK columns per target type if SQLite CHECK wiring is preferred over application validation.

## 9.3 `narrative_notes`

Optional researcher commentary on a composition (same pattern as other typed note tables). Creation and edit attribution via audit history.

## 9.4 Ref prefix

Proposed catalog prefix: **`NAR`** (`narratives`). Register in [`catalog-refs.md`](catalog-refs.md) when promoted. Must not collide with reserved prefixes or `subject_types.ref_prefix` values.

## 9.5 Search and share

- **FTS:** index `title` + `body` when prose kinds ship; see omnibar registry patterns in deployment-plan spike 3 notes.
- **Share packages:** narratives may be the most human-readable export unit — prose plus optional bundled Sources. See [`ideas/share-packages.md`](ideas/share-packages.md).

---

# 10. Product surfaces (horizon)

Narrative is not only essays. Possible workspace destinations:

| Surface | Narrative role |
| --- | --- |
| Family tree | Default projection; saved compositions for focus/layout |
| Timeline | Event sequence view; curated “story timelines” |
| Map / GIS | Pins from Locations; authored migration and hub compositions |
| Person page | Tabs: Facts (Conclusion), Evidence (Interpretation), Story (linked Narrative excerpts) |
| Source reader | Evidence-first; link out to narratives that cite this Source |

Implementation order is a product decision. This document does not schedule spikes.

---

# 11. Open questions

1. **Prefs vs. `NAR` rows** — Are saved tree focus and map time range full compositions or lighter session/chrome state under Application Support?
2. **Rich text format** — Markdown with reference tokens vs. structured document AST vs. HTML storage.
3. **Anchor durability** — Best anchor model for edited prose (block IDs, CRDT, footnote-only links).
4. **Auto-extract** — Resist promoting narrative text to Reconciliation/Observation automatically in v1?
5. **Embedded vs. standalone** — One schema for long-form essays and short person biographies, or separate kinds/UI?
6. **External authorship** — Import of published family history as Source vs. Narrative when the researcher is the author.
7. **Living people** — Redaction and warn-before-share for narrative exports.
8. **GEDCOM and interop** — `NOTE` records as export adapter, not as the in-catalog Narrative model.

---

# 12. Documentation ownership

To avoid competing schema definitions:

- [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md) summarizes cross-layer philosophy (including this proposed fourth layer when promoted).
- [`source-layer-data-model.md`](source-layer-data-model.md) — Source tables and Artifact/File storage.
- [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md) — Interpretation tables and vocabulary.
- [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) — Conclusion tables and Claims.
- **This document** — Narrative layer projections, compositions, references, and proposed tables.
- [`structured-date-model.md`](structured-date-model.md), [`structured-name-model.md`](structured-name-model.md), [`audit-revision-history.md`](audit-revision-history.md), [`research-judgment-model.md`](research-judgment-model.md) — shared infrastructure unchanged.

---

# 13. Summary diagram

```text
Source → Interpretation → Conclusion → Narrative
  evidence    cited atoms     working belief    arrangement & communication
                                         ├── projections (tree, map, timeline)
                                         └── compositions (prose, curated views)
                                              └── references → graph (never auto-upstream)
```
