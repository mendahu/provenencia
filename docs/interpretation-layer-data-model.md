# Provenencia Genealogy — Interpretation Layer Data Model

## Status

Draft architecture notes. This document is the authoritative schema and design reference for the Provenencia Interpretation layer.

The Interpretation layer answers:

> What does this particular source appear to say?

Cross-layer philosophy and the relationship among Source, Interpretation, and Conclusion are summarized in [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md).

The authoritative Source-layer schema is [`source-layer-data-model.md`](source-layer-data-model.md). Conclusion-layer schema is [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md). Shared date and name value models are [`structured-date-model.md`](structured-date-model.md) and [`structured-name-model.md`](structured-name-model.md). Seeded keys and open-vocabulary starters are [`seeded-vocabulary.md`](seeded-vocabulary.md). Audit history is [`audit-revision-history.md`](audit-revision-history.md).

---

# 1. Interpretation-layer principles

## 1.1 Preserve evidence; make interpretation revisable

Source evidence must never be rewritten to match a later interpretation.

For ambiguous handwriting, the Citation may preserve a faithful researcher transcription such as `Robins [?]` or `[William?] Smith`. Normalization of that reading belongs to an Observation rather than rewriting the Citation transcription.

## 1.2 Interpretation is a cited property graph

Nodes provide stable identity for source-local things. Observations provide atomic, cited assertions about those Nodes. Properties define the meaning and primitive value type of those assertions.

The database should enforce generic graph integrity and primitive typing. It should not attempt to encode the entire genealogy ontology into rigid table structure.

**Invariants** are that graph shape (Citation, subject Node, Property, typed value, polarity). **Conventions** are how to get useful genealogy out of it (one grain of Place, Observations that match what the Citation appears to say, seeded `location` ends). UI may warn; writers must not reject rows because a convention was skipped. See [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) §1.

## 1.3 Interpretation vocabulary is extensible

Node Types and Properties are first-class data. Provenencia ships with a useful seeded vocabulary, but researchers may add new Node Types and Properties without a database schema migration.

All Properties, including user-defined Properties, declare a value type. Unknown or custom vocabulary remains preservable and generically usable even when the core application has no specialized semantics for it.

Predicate _values_ such as `event_type` and participation `role` are likewise open text vocabulary: seeded with useful defaults, researcher-extensible, and not enforced as closed database enums. The Interpretation layer must remain ready for event kinds and roles the product cannot anticipate.

Core application logic and future plugins may provide first-class behavior for recognized keys and values while leaving storage generic. For example, the application may treat an Event with `event_type = birth` and a Participation with `role = subject` as that person's birth, without requiring the schema to encode birth-specific tables or mandatory role constraints.

## 1.4 Derived semantics belong to the application layer

The Interpretation schema stores explicit normalized assertions derived from evidence. Relationships that can be inferred from those assertions are application-level projections rather than duplicated persisted Interpretation data.

For example, a birth Event with subject and father Participations can allow the application to infer a father/child relationship without separately persisting that inferred relationship.

## 1.5 Negation is explicit

Absence of a positive assertion is not equivalent to a negative assertion.

```text
No Observation:
  unknown whether the name is Jake

Positive:
  the name is Jake

Negative:
  the name is not Jake
```

Polarity belongs on the Observation: it records what a particular source appears to assert or deny. Conflicting evidence remains a separate concern from negation. Two positive Observations with different values conflict; a negative Observation denies a specific proposition.

## 1.6 Persistence conventions

Persistent rows use globally unique machine identifiers, currently UUIDv7 stored as 16-byte SQLite `BLOB` values. Ordinary schema tables use SQLite `STRICT` typing.

Structured genealogical dates use [`structured-date-model.md`](structured-date-model.md). Structured personal names use [`structured-name-model.md`](structured-name-model.md).

Generic `created_at` / `updated_at` / user bookkeeping does not belong on these domain tables; see [`audit-revision-history.md`](audit-revision-history.md).

Selected user-facing entities receive a required short human-readable `ref`, unique within the project. This layer uses `CIT`, `OBS`, and Node refs of the form `{type_prefix}-C-{token}` (candidate). Shared rules are in [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md).

---

# 2. Layer overview

The Interpretation layer is a generic, schema-described property graph:

```text
Source
  └── Artifact
        └── Citation
              └── Observation
                    ├── subject Node
                    ├── Property
                    └── typed value
                          └── may be another Node
```

There are no specialized `person_records`, `event_records`, `place_records`, `relationship_records`, or `participation_records` tables. A Node's type and its cited Observations provide the structure that those Record tables previously attempted to encode.

---

# 3. `citations`

A Citation selects an addressable portion of exactly one Artifact.

Each Citation is independently resolvable from its `artifact_id` and locator. Citations are not nested; the locator contains all information required to identify its target within the Artifact.

```sql
CREATE TABLE citations (
    id              BLOB PRIMARY KEY,
    ref             TEXT UNIQUE NOT NULL,      -- e.g. CIT-3K9M2
    artifact_id     BLOB NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    locator_json    TEXT NOT NULL,
    transcription   TEXT,
    description     TEXT
) STRICT;
```

`ref` is a required short human-readable reference for UI and discussion (prefix `CIT`).
`transcription` preserves the researcher's reading of textual or spoken content within the cited evidence. It is intended to remain faithful to the evidence, including uncertainty where appropriate, rather than silently normalizing abbreviations, names, places, or other values. For example, `Wm Robins` may be transcribed as written and normalized to `William Robins` later through an Observation.

`description` records what the researcher observes in the cited evidence. It is media-neutral and may describe visual, textual, audio, or other characteristics. It is not specifically an accessibility `alt_text` field.

Researcher commentary about the Citation itself belongs in `citation_notes`.

Conceptually:

```text
Artifact
    raw evidence

Citation.transcription
    what I read/hear

Citation.description
    what I observe

citation_notes
    researcher commentary about this Citation

Observation
    what I think it means
```

Research notes attached to a Citation live in a typed child table. Multiple notes are allowed so commentary can accumulate without overwriting earlier remarks.

```sql
CREATE TABLE citation_notes (
    id              BLOB PRIMARY KEY,
    citation_id     BLOB NOT NULL REFERENCES citations(id) ON DELETE CASCADE,
    body            TEXT NOT NULL
) STRICT;
```

Notes use a typed table with a real foreign key rather than a polymorphic notes table. Creation, edit, and deletion attribution belong to audit history.

## 3.1 Locator design

A locator is a versioned JSON document containing an ordered list of selectors. Each selector narrows the context established by the selectors before it.

This makes locators composable rather than forcing every Citation into one mutually exclusive `locator_type`.

For example:

```text
Artifact: PDF
  → artifact page 37 / marked page 23
  → polygon region on that page
```

is one Citation with two selectors, not a page Citation containing a child region Citation.

The top-level locator schema is:

```json
{
  "version": 1,
  "selectors": [{ "type": "..." }]
}
```

Semantically:

```text
LocatorV1 {
    version: 1
    selectors: Selector[]   // ordered, at least one
}
```

The application must validate the selector chain as a whole. A selector is interpreted relative to the context produced by the preceding selector. For example, `page` can select a page from a PDF and `region` can then select a polygon within that page.

The selector vocabulary is application-defined but extensible. The MVP should implement only selectors needed by concrete workflows. Unknown selector types must be preserved losslessly even if the current client cannot render or edit them.

There is intentionally no separate `locator_type` database column. A composable locator may contain several selector types, so a single database-level type would be ambiguous and redundant.

## 3.2 `page` selector

The `page` selector identifies a page by its position in the Artifact itself.

For a PDF, `artifact_page` is the 1-based page position the application uses to navigate the PDF and is the authoritative locator value.

A scanned book, register, newspaper, or archival document may display a different page number or label on the scanned page itself. That source pagination is retained separately as optional descriptive information in `page_label`.

```json
{
  "type": "page",
  "artifact_page": 37,
  "page_label": "23"
}
```

Schema:

```text
PageSelector {
    type: "page"
    artifact_page: integer >= 1
    page_label?: string
}
```

`page_label` is deliberately text rather than an integer because source pagination may contain values such as:

```text
iv
xii
A-3
23a
folio 17r
```

An unnumbered page simply omits `page_label`.

The general rule is:

> `artifact_page` answers where the Citation is in the digital Artifact. `page_label` records how that page identifies itself in the underlying source.

## 3.3 `region` selector

Selects an arbitrary polygonal region from an image-like context, including a standalone image, a rendered PDF page, or a video frame context.

A polygon is used rather than a rectangle so the same selector can represent simple rectangular crops as well as irregular evidence such as handwriting blocks, seals, marginal notes, damaged fragments, or people in photographs.

```json
{
  "type": "region",
  "points": [
    { "x": 0.31, "y": 0.18 },
    { "x": 0.73, "y": 0.18 },
    { "x": 0.7, "y": 0.39 },
    { "x": 0.34, "y": 0.42 }
  ],
  "unit": "normalized"
}
```

Schema:

```text
RegionSelector {
    type: "region"
    points: Point[]       // at least 3 points
    unit: "normalized"
}

Point {
    x: number between 0 and 1
    y: number between 0 and 1
}
```

The polygon is implicitly closed by connecting the final point back to the first point. The first point should not be repeated at the end of the array.

For version 1:

- `points` must contain at least three distinct points;
- every point must lie within the normalized media bounds;
- points are interpreted in array order as the polygon boundary;
- the polygon must not self-intersect;
- degenerate polygons with zero area are invalid.

Normalized coordinates keep Citations stable across rendering resolutions and generated previews.

A rectangle is represented as an ordinary four-point polygon. The application may provide rectangular drag-selection as a UI convenience and serialize it as four points.

A crop within a PDF is represented compositionally:

```json
{
  "version": 1,
  "selectors": [
    {
      "type": "page",
      "artifact_page": 37,
      "page_label": "23"
    },
    {
      "type": "region",
      "points": [
        { "x": 0.31, "y": 0.18 },
        { "x": 0.73, "y": 0.18 },
        { "x": 0.7, "y": 0.39 },
        { "x": 0.34, "y": 0.42 }
      ],
      "unit": "normalized"
    }
  ]
}
```

The same `region` selector works directly against a standalone image without a preceding `page` selector.

## 3.4 `time_range` selector

Selects an interval from time-based media such as audio or video.

```json
{
  "type": "time_range",
  "start_ms": 802400,
  "end_ms": 845100
}
```

Schema:

```text
TimeRangeSelector {
    type: "time_range"
    start_ms: integer >= 0
    end_ms: integer > start_ms
}
```

Milliseconds are used as the canonical stored unit so the representation is unambiguous and integer-based.

Selectors may be composed. For example, a Citation could identify a polygonal region of a video frame during a particular interval:

```json
{
  "version": 1,
  "selectors": [
    {
      "type": "time_range",
      "start_ms": 15120,
      "end_ms": 19440
    },
    {
      "type": "region",
      "points": [
        { "x": 0.12, "y": 0.08 },
        { "x": 0.42, "y": 0.08 },
        { "x": 0.42, "y": 0.53 },
        { "x": 0.12, "y": 0.53 }
      ],
      "unit": "normalized"
    }
  ]
}
```

## 3.5 `text_quote` selector

Selects textual content by its text rather than by unstable paragraph numbering or rendered coordinates.

```json
{
  "type": "text_quote",
  "exact": "William Robins, carpenter",
  "prefix": "household of ",
  "suffix": " aged 43"
}
```

Schema:

```text
TextQuoteSelector {
    type: "text_quote"
    exact: non-empty string
    prefix?: string
    suffix?: string
}
```

`exact` contains the text being selected. Optional `prefix` and `suffix` provide surrounding context to disambiguate repeated text without becoming part of the selected content.

For a PDF with a text layer, a page plus text quote can identify a paragraph or phrase without relying on a paragraph number:

```json
{
  "version": 1,
  "selectors": [
    {
      "type": "page",
      "artifact_page": 37,
      "page_label": "23"
    },
    {
      "type": "text_quote",
      "exact": "William Robins, carpenter",
      "prefix": "household of ",
      "suffix": " aged 43"
    }
  ]
}
```

## 3.6 Future selectors

The selector system is intentionally open to additional addressable media and structured data. Possible future selectors include:

```text
text_position
table_row
table_cell
csv_row
json_pointer
xpath
```

These are not part of the version 1 supported vocabulary until a concrete workflow requires them.

Adding a selector type does not require changing the `citations` table. It requires defining that selector's JSON shape, validation rules, and application behavior.

## 3.7 Locator invariants

For locator version 1:

1. `version` must equal `1`.
2. `selectors` must contain at least one selector.
3. Selectors are ordered and interpreted from the Artifact inward.
4. Every selector must contain a string `type` discriminator.
5. Known selector types must satisfy their type-specific schema and context requirements.
6. Unknown selector types are preserved losslessly for forward compatibility.
7. A Citation must be independently resolvable from `artifact_id` plus `locator_json`; it never depends on another Citation.
8. Region vertices use normalized coordinates relative to the selected media context, not a particular UI rendering.
9. Region polygons contain at least three distinct points, are non-self-intersecting, and have non-zero area.
10. For paginated digital Artifacts, the Artifact page position is authoritative for navigation.
11. Printed or marked source pagination is supplementary descriptive data and does not replace the Artifact page position.
12. Locator JSON identifies where the evidence is; transcription, description, and interpretation remain separate concerns.

Conceptually:

```text
Citation = Artifact + complete composable locator
```

UI hierarchy or containment can be derived from locator selector chains when useful without making Citation hierarchy part of the persisted evidence model.

---

# 4. Node vocabulary

## 4.1 `node_types`

Node Types define the semantic category of a Node. They are data rather than a database enum so the vocabulary can be extended without schema migrations.

```sql
CREATE TABLE node_types (
    key             TEXT PRIMARY KEY,
    label           TEXT NOT NULL,
    description     TEXT,
    ref_prefix      TEXT NOT NULL UNIQUE
) STRICT;
```

`ref_prefix` is the type token in human-readable refs for Nodes and canonical entities of this kind, for example `PER` for `person` → Person `PER-7KD45`, candidate Node `PER-C-7KD45`. It is required when defining a Node Type, including researcher-defined types. Prefixes are uppercase ASCII letters, unique among Node Types, and must not use the reserved prefixes `SRC`, `ART`, `CIT`, `OBS`, or the candidate layer code `C`.

The Node Type `source` (reification of a Source row) must not use `SRC`; a distinct prefix such as `SRN` keeps Source catalog refs (`SRC-…`) distinguishable from source-Nodes in speech.

Application semantics attach to stable `key` values rather than a persisted built-in flag. Horizon Node Types and their prefixes are catalogued in [`seeded-vocabulary.md`](seeded-vocabulary.md). That set is expected to include at least `person`, `event`, `place`, `relationship`, `participation`, `location`, and `source`.

These names describe application semantics, not different SQL structures. Every instance is stored in the same `nodes` table.

`relationship`, `participation`, and `location` are examples of bridge-like Nodes. `source` is a reification Node: it lets the Interpretation graph talk about evidentiary objects, not only historical persons, events, and places. Structurally, however, the database does not distinguish these categories. Any Node can be related to any other Node through a Node-valued Observation. The application vocabulary defines what those relationships mean and which combinations are semantically useful.

Inferred associations that no Source asserted (for example an extra Location grain) belong on Conclusion handles and Reconciliation Claims, not as extra Observations on the wrong Citation. See [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md).

## 4.2 `nodes`

A Node gives stable identity to a source-local thing or association encountered during interpretation.

```sql
CREATE TABLE nodes (
    id              BLOB PRIMARY KEY,
    ref             TEXT UNIQUE NOT NULL,      -- e.g. PER-C-7KD45
    source_id       BLOB NOT NULL REFERENCES sources(id),
    node_type_key   TEXT NOT NULL REFERENCES node_types(key),
    label           TEXT,
    description     TEXT,

    UNIQUE (id, node_type_key)
) STRICT;
```

`ref` is required. It is assembled as `{ref_prefix}-C-{token}` (`C` = candidate). Users talk about a person Node as a candidate person (`PER-C-…`), distinct from the canonical Person (`PER-…`).

`UNIQUE (id, node_type_key)` exists so Sameness Claims can use a composite foreign key that pins both endpoints to the same type. It is redundant with the primary key for uniqueness of `id`; it does not allow two types per Node.

`node_type_key` is immutable after insert. Correcting a wrong type means a new Node (and new `ref`), not an UPDATE of the type. The UUID remains the machine identity; the type is part of the public identity encoded in `ref`.

`source_id` is the Node's home Source — typically the Source being interpreted when the Node was created. It exists so the application can efficiently surface Nodes that belong with a given Source during common same-source workflows. It does not restrict which Citations or Observations may reference the Node; cross-source Observations remain valid.

For Nodes of type `source`, `source_id` has a stronger meaning: it identifies the Source this Node reifies. The application should maintain at most one `source` Node per Source row. Other Sources may then make cited Observations about that Node—bare references such as “this book mentions that marriage certificate,” free-text remarks about authenticity or errors, or both—without collapsing that material into Observations about historical persons alone, and without requiring structured Source-quality columns.

Nodes deliberately contain little domain data. A person's name, an Event's date, a Place's name, or a Participation's role belongs in cited Observations rather than fixed Node columns.

`description` is an optional short summary of the Node itself. Nodes do not have a multi-note table; researcher commentary about interpreted assertions belongs on the supporting Observations and can be aggregated from `observation_notes` when a Node-centric view is needed.

A Node can therefore be sparse. Creating a `person` Node does not require knowing a name, date, or any other property.

---

# 5. Property vocabulary

## 5.1 `properties`

Properties are first-class, user-extensible definitions of predicates that may appear in Observations.

```sql
CREATE TABLE properties (
    key             TEXT PRIMARY KEY,
    label           TEXT NOT NULL,
    description     TEXT,
    value_type      TEXT NOT NULL,

    CHECK (value_type IN (
        'text',
        'integer',
        'real',
        'boolean',
        'date',
        'name',
        'node'
    ))
) STRICT;
```

A Property's `value_type` is intrinsic to the Property. Seeded Properties (for example `name`, `birth_date`, `event_type`, `role`, `person`, `mentions`, `remark`) and their `node_type_properties` bindings are listed in [`seeded-vocabulary.md`](seeded-vocabulary.md).

The semantic vocabulary is open, but the primitive value system is intentionally constrained. A researcher may define a new Property without introducing a new storage type.

`value_type = 'date'` always means the shared structured DateValue model in [`structured-date-model.md`](structured-date-model.md), not a SQL date or free-text date string.

`value_type = 'name'` always means the shared structured NameValue model in [`structured-name-model.md`](structured-name-model.md), not a single undifferentiated text string. A NameValue always has a full-form `form` and may optionally include ordered parts with an open part-type vocabulary for search and reconciliation.

`name_format` is primarily a Conclusion Property (Reconciliation Claim on a person entity). It need not appear in `node_type_properties` for Interpretation unless a Source itself asserts a naming convention.

Application semantics attach to stable `key` values, matching `node_types`. Seeded Properties may receive first-class application behavior. User-defined Properties remain first-class persisted data and can be generically displayed, searched, audited, synced, and referenced. Plugins may add specialized semantics for additional Properties later.

Open text values such as `event_type` and `role` are seeded with common defaults for pickers but remain researcher-extensible; see [`seeded-vocabulary.md`](seeded-vocabulary.md). The schema does not close those sets or require particular roles for particular event types; first-class workflows recognize well-known values in application logic.

Interactions between `source` Nodes should stay deliberately lightweight. Provenencia does not seed a structured source-quality ontology (`is_authentic`, defect codes, and similar).

Two common shapes:

```text
mentions   -> node   # bare source-to-source reference; target is typically a source Node
remark     -> text   # free-text commentary about a source Node
```

A book that merely cites a marriage certificate can record `BookSource -- mentions --> CertificateSource` with no remark. A letter that challenges a certificate can add text `remark` Observations and, when needed, ordinary person-level Observations as well. Structured Properties beyond this may be added later only if a concrete workflow requires them.

## 5.2 `node_type_properties`

This table defines which Properties are valid for which Node Types.

```sql
CREATE TABLE node_type_properties (
    node_type_key   TEXT NOT NULL REFERENCES node_types(key),
    property_key    TEXT NOT NULL REFERENCES properties(key),

    PRIMARY KEY (node_type_key, property_key)
) STRICT;
```

For example (full matrix in [`seeded-vocabulary.md`](seeded-vocabulary.md)):

```text
person        -> name
person        -> birth_date

event         -> event_type
event         -> date

place         -> toponym

participation -> person
participation -> event
participation -> role

location      -> event
location      -> place

source        -> mentions
source        -> remark
```

This is a vocabulary/schema relationship, not historical research data. It says that `participation.person` is a meaningful shape in the Interpretation graph; it does not assert that any particular Person participated in any particular Event.

The vocabulary should be seeded with common definitions but remain researcher-extensible.

For Node-valued Properties, allowed **target** Node Types (for example, `participation.person` should target a `person` Node) are an **application invariant** for now — the same posture as Observation value population. Seeded Properties get first-class UI/validation behavior; user-defined node Properties may remain unconstrained or warn-only. Provenencia does not persist target-type allow-lists in SQL yet (no `target_node_type_key` on `properties`, and no target join table). That can be added later if pickers and importers need a shared declarative vocabulary.

Malformed edges (wrong target type) may be warned about or ignored by typed workflows; the generic graph still stores the Observation.

---

# 6. Relationships between Nodes

There is intentionally no separate table containing historical graph edges.

A relationship between two Nodes is an Observation whose Property has `value_type = 'node'`:

```text
subject Node -- Property --> object Node
```

For example:

```text
Participation PT1 -- person --> Person P1
Participation PT1 -- event  --> Event E1
Participation PT1 -- role   --> "subject"
```

or, where the evidence gives only an indeterminate/general association:

```text
Relationship R1 -- participant --> Person P1
Relationship R1 -- participant --> Person P2
Relationship R1 -- relationship_type --> "cousin"
```

The graph stores only explicit normalized interpretation. It does not need to persist an additional `father_of` edge if application logic can infer that relationship from a birth Event and its Participations.

This allows the same structural model to represent nuclear family roles, extended kinship, step relationships, guardianship, employment, friendship, household roles, and unanticipated historical associations without adding bridge tables.

The same mechanism covers source-to-source evidence. That includes bare references and optional free-text commentary:

```text
Source Node SB1 (reifies book B)
Source Node SC1 (reifies marriage certificate C)
Person Node P1

Book Citation B1 supports a bare reference:
  SB1 -- mentions --> SC1

Letter Citation L1 supports commentary and historical-world facts:
  SC1 -- remark --> "birth certificate is fake"
  # or, when the document is trusted but a fact is disputed:
  SC1 -- remark --> "date of birth on certificate mistyped"
  P1  -- birth_date --> DateValue(2 JAN 1800)          # polarity positive
  P1  -- birth_date --> DateValue(1 JAN 1800)          # polarity negative
```

`mentions` is a Node-valued edge and need not say anything further about the referenced Source. `remark` is ordinary text-valued Interpretation of what a citing Source appears to say about another Source. Neither is a Source-layer column or a closed authenticity taxonomy. Historical-world Observations remain separate and independently citable.

---

# 7. `observations`

An Observation is the atomic unit of interpreted evidence. Every Observation is independently addressable and is supported by one Citation.

Conceptually:

```text
Observation {
    citation
    subject Node
    Property
    polarity          -- positive | negative
    typed value
}
```

The Citation preserves what the source actually contains; the Observation may normalize that evidence into a domain value. For example:

```text
Citation.transcription
    "Age: 42"

Observation
    polarity = positive
    Person P1 -- birth_date --> DateValue(...)
```

A Source may also appear to deny a proposition. That is still Interpretation, not a Conclusion-layer judgment:

```text
Citation.transcription
    "not Jake"

Observation
    polarity = negative
    Person P1 -- name --> NameValue(form = "Jake")
```

Absence of any name Observation means the name is unknown from that evidence. It does not mean the name is not Jake.

This distinction lets Observation values remain strongly typed without attempting to reproduce every possible source representation.

The intended primitive value categories are currently:

```text
text
integer
real
boolean
date
name
node
```

A provisional SQL shape is:

```sql
CREATE TABLE observations (
    id              BLOB PRIMARY KEY,
    ref             TEXT UNIQUE NOT NULL,      -- e.g. OBS-2F8Q1
    citation_id     BLOB NOT NULL REFERENCES citations(id),
    subject_node_id BLOB NOT NULL REFERENCES nodes(id),
    property_key    TEXT NOT NULL REFERENCES properties(key),
    polarity        TEXT NOT NULL DEFAULT 'positive',

    value_text      TEXT,
    value_integer   INTEGER,
    value_real      REAL,
    value_boolean   INTEGER,
    value_date_id   BLOB REFERENCES date_values(id),
    value_name_id   BLOB REFERENCES name_values(id),
    value_node_id   BLOB REFERENCES nodes(id),

    CHECK (polarity IN ('positive', 'negative')),
    CHECK (value_boolean IS NULL OR value_boolean IN (0, 1))
) STRICT;
```

`ref` is a required short human-readable reference for UI and discussion (prefix `OBS`).

**Value population (application invariant):** exactly one value representation must be populated, and it must match `properties.value_type`. Provenencia does **not** enforce that cross-table rule in SQLite for now (no XOR/`value_type` trigger or typed bridge tables). Writers — repository code, importers, sync — are responsible for correct population.

**Read-side tolerance:** if a row is malformed, readers should prefer the column that matches the Property's `value_type` and ignore any other non-null value columns. A row with *no* usable value for that `value_type` is invalid and should be surfaced as an error or omitted rather than guessed. Reconciliation Claims use the same sparse-column idea, except `value_type = 'node'` is stored as `value_entity_id` (a canonical handle) rather than `value_node_id`. See [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md).

Reading uncertainty belongs in Citation transcription or description when needed. Alternative or competing interpretations are modeled as separate Observations rather than a numeric confidence score on a single row. Researcher commentary about a particular Observation belongs in `observation_notes`.

```sql
CREATE TABLE observation_notes (
    id              BLOB PRIMARY KEY,
    observation_id  BLOB NOT NULL REFERENCES observations(id) ON DELETE CASCADE,
    body            TEXT NOT NULL
) STRICT;
```

Notes use a typed table with a real foreign key rather than a polymorphic notes table. Creation, edit, and deletion attribution belong to audit history.

For Node-valued Observations, the object Node must exist and should satisfy any target Node Type constraint defined by the vocabulary. Polarity applies to Node-valued Observations as well: a negative Observation denies that particular edge rather than deleting or omitting it.

A single Citation may support many atomic Observations:

```text
Citation C1
  -> Person P1 -- name       --> NameValue(form = "William Robins", …)
  -> Person P1 -- occupation --> "Carpenter"
  -> Person P1 -- birth_date --> DateValue(...)
```

Multiple Observations may also make different assertions about the same Property without forcing a single value onto the Node. Conflicting or alternative interpretations therefore remain independently citable and auditable. Multiple name Observations on one person Node are expected when Sources use different forms, nicknames, or name changes.

---

# 8. Interpretation-layer invariants

The current design aims to preserve these invariants:

1. Every Node has exactly one Node Type. `node_type_key` is immutable after insert.
2. Every Node has a required `ref` of the form `{ref_prefix}-C-{token}` (candidate).
3. Every Node has a home `source_id`; that home Source does not confine which Observations may target the Node.
4. A Node of type `source` reifies the Source identified by its `source_id`; the application should keep at most one such Node per Source.
5. Node Types and Properties are extensible persisted vocabulary, not closed application enums. A new Node Type includes a unique `ref_prefix` that is not a reserved catalog prefix or layer code.
6. Every Observation has its own stable identity and a required `OBS-…` `ref`.
7. Every Observation is supported by exactly one Citation.
8. Every Citation has a required `CIT-…` `ref`.
9. Every Observation has exactly one subject Node and one Property.
10. Every Observation has exactly one typed value.
11. Every Observation has an explicit polarity of `positive` or `negative`; absence of an Observation is not negation.
12. The Observation value must match the Property's declared `value_type` (application write rule). Readers prefer the matching column if multiple value columns are set.
13. Date-valued Observations reference a structured DateValue; they do not store SQL dates or free-text dates as the typed value.
14. Name-valued Observations reference a structured NameValue; they do not store an undifferentiated name string as the typed value.
15. A Property used on a Node must be allowed for that Node's Node Type.
16. A Node-valued Observation forms a graph edge and its object Node must exist.
17. Application logic may constrain the target Node Type of Node-valued Properties for seeded vocabulary; that is not enforced as SQL allow-lists in this draft.
18. Citation text/description preserves the evidence representation; Observations contain normalized interpretation.
19. Derived genealogical semantics are not duplicated into the Interpretation graph merely for convenience.
20. Unknown/custom Node Types, Properties, and open vocabulary values (such as event types and roles) remain preservable and generically usable without first-class application support.
21. First-class application behavior may recognize seeded keys and values; it must not require the schema to close those vocabularies or encode every genealogical edge case.

---

# 9. Documentation ownership

To avoid competing schema definitions:

- [`source-layer-data-model.md`](source-layer-data-model.md) is authoritative for Source-layer tables and Artifact/File storage.
- This document is authoritative for Interpretation-layer tables and vocabulary schema.
- [`seeded-vocabulary.md`](seeded-vocabulary.md) is the horizon catalog for intended Node Types, Properties, bindings, and open-value starters (not a v1 ship list).
- [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) is authoritative for Conclusion-layer tables and Claims.
- [`structured-date-model.md`](structured-date-model.md) is authoritative for shared DateValue persistence.
- [`structured-name-model.md`](structured-name-model.md) is authoritative for shared NameValue persistence.
- [`audit-revision-history.md`](audit-revision-history.md) is authoritative for audit and revision history.
- [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md) summarizes the three-layer philosophy and cross-layer examples.
