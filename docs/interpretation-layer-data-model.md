# Provenencia Genealogy — Interpretation Layer Data Model

## Status

Draft architecture notes. This document is the authoritative schema and design reference for the Provenencia Interpretation layer.

The Interpretation layer answers:

> What does this particular source appear to say?

Cross-layer philosophy and the relationship among Source, Interpretation, and Conclusion are summarized in [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md).

The authoritative Source-layer schema is [`source-layer-data-model.md`](source-layer-data-model.md). Conclusion-layer schema is [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md). Shared date and name value models are [`structured-date-model.md`](structured-date-model.md) and [`structured-name-model.md`](structured-name-model.md). Seeded keys and open-vocabulary starters are [`seeded-vocabulary.md`](seeded-vocabulary.md). Audit history is [`audit-revision-history.md`](audit-revision-history.md). Researcher judgment (Source credibility assessments, Citation transcription certainty, Claim confidence) is [`research-judgment-model.md`](research-judgment-model.md).

---

# 1. Interpretation-layer principles

## 1.1 Preserve evidence; make interpretation revisable

Source evidence must never be rewritten to match a later interpretation.

For ambiguous handwriting, the Citation may preserve a faithful researcher transcription such as `Robins [?]` or `[William?] Smith`. Normalization of that reading belongs to an Observation rather than rewriting the Citation transcription.

## 1.2 Interpretation is a cited property graph

Subjects provide stable identity for source-local things. Observations provide atomic, cited assertions about those subjects. Properties define the meaning and primitive value type of those assertions.

The database should enforce generic graph integrity and primitive typing. It should not attempt to encode the entire genealogy ontology into rigid table structure.

**Invariants** are that graph shape (Citation, subject, Property, typed value, polarity). **Conventions** are how to get useful genealogy out of it (one grain of Place, Observations that match what the Citation appears to say, seeded `location` ends). UI may warn; writers must not reject rows because a convention was skipped. See [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) §1.

## 1.3 Interpretation vocabulary is extensible

Subject types and Properties are first-class data. Provenencia ships with a useful seeded vocabulary (`origin = 'provenencia'`), but researchers may add new Subject types and Properties (`origin = 'user'`) without a database schema migration. Future plugins use the reserved `plugin:<plugin_id>` origin namespace. Vocabulary origin rules: [`seeded-vocabulary.md`](seeded-vocabulary.md) §1.1.

All Properties, including user-defined Properties, declare a value type. Unknown or custom vocabulary remains preservable and generically usable even when the core application has no specialized semantics for it.

Predicate _values_ that carry **kind or edge identity** — notably `event_type`, participation `role`, and `relationship_type` — are **Property terms**: origin-namespaced vocabulary rows (`property_terms`), not free-text Observation strings. Observations store `value_term_id`. Product and plugins seed large term sets; researchers may add `origin=user` terms for the long tail. First-class UI behavior (birthday facets, family-tree edges, connect disambiguation) attaches to recognized `(property key, term key, origin)` pairs in the Interpretation subject registry — without encoding birth-specific tables.

Free text remains appropriate for true prose Properties (`remark`, `toponym`, researcher-defined notes), not for kind/edge vocabulary. Name part `type` strings stay an open part vocabulary on NameValue (see [`structured-name-model.md`](structured-name-model.md)); they are not Property terms.

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

Selected user-facing entities receive a required short human-readable `ref`, unique within the project. This layer uses `CIT`, `OBS`, and Subject refs minted off the Subject type's `candidate_ref_prefix` (`CPR-…`, `CEV-…`, `CPL-…`). Shared rules are in [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md).

---

# 2. Layer overview

The Interpretation layer is a generic, schema-described property graph:

```text
Source
  └── Artifact
        └── Citation
              └── Observation
                    ├── Subject
                    ├── Property
                    └── typed value
                          └── may be another Subject
```

There are no specialized `person_records`, `event_records`, `place_records`, `relationship_records`, or `participation_records` tables. A subject's type and its cited Observations provide the structure that those Record tables previously attempted to encode.

---

# 3. `citations`

A Citation selects an addressable portion of exactly one Artifact.

Each Citation is independently resolvable from its `artifact_id` and locator. Citations are not nested; the locator contains all information required to identify its target within the Artifact.

```sql
CREATE TABLE citations (
    id                        BLOB PRIMARY KEY,
    ref                       TEXT UNIQUE NOT NULL,      -- e.g. CIT-3K9M2
    artifact_id               BLOB NOT NULL REFERENCES artifacts(id),
    locator_json              TEXT NOT NULL,
    transcription             TEXT,
    description               TEXT,
    transcription_uncertain   INTEGER NOT NULL DEFAULT 0,
    transcription_note        TEXT,

    CHECK (transcription_uncertain IN (0, 1))
) STRICT;
```

`ref` is a required short human-readable reference for UI and discussion (prefix `CIT`).
`artifact_id` is `NO ACTION`: deleting an Artifact while Citations remain fails. Citation notes still CASCADE with the Citation.
`transcription` preserves the researcher's reading of textual or spoken content within the cited evidence. It is intended to remain faithful to the evidence, including uncertainty where appropriate, rather than silently normalizing abbreviations, names, places, or other values. For example, `Wm Robins` may be transcribed as written and normalized to `William Robins` later through an Observation.

`description` records what the researcher observes in the cited evidence. It is media-neutral and may describe visual, textual, audio, or other characteristics. It is not specifically an accessibility `alt_text` field.

`transcription_uncertain` is a media-agnostic boolean: when set, the transcription (or equivalent representation of spoken/visual content) may not faithfully represent the cited Artifact portion — faint ink, damaged scan, garbled audio, muddy video, and similar. Optional `transcription_note` records why (for example `garbled audio ~1:02` or `surname damaged`). In-line marks such as `[?]` in `transcription` remain valid and complementary. This is not Claim confidence and not Source credibility; see [`research-judgment-model.md`](research-judgment-model.md).

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

## 3.4 `artifact` selector

Names the **digital Artifact as a whole** as the starting context for the locator chain. Citations always bind an `artifact_id`; `artifact` makes “no further narrowing” an explicit, valid locator (and the UI default) instead of an empty `locator_json`. Not `document` (paper-biased) and not `source` (a Source can have several Artifacts).

```json
{
  "type": "artifact"
}
```

Schema:

```text
ArtifactSelector {
    type: "artifact"
}
```

Rules:

1. When present, `artifact` is the **first** selector in the chain (outermost context).
2. It may stand alone (cite the entire Artifact) or be followed by narrowing selectors such as `page` and/or `region`.
3. Prefer an `artifact`-only chain over an empty locator. Citations require a locator; empty JSON is invalid.
4. Product UI prepopulates `[{ "type": "artifact" }]` and layers **Set Page** / region on top; there is no separate “cite entire artifact” checkbox.
5. On paginated Artifacts (PDF), a `region` selector must be preceded by a `page` selector. If the UI creates a region while no page is set, it must insert `page` for the current viewer page before (or with) the region. Standalone images may use `artifact` → `region` with no `page`.

Example — whole Artifact:

```json
{
  "version": 1,
  "selectors": [{ "type": "artifact" }]
}
```

Example — page then region on that Artifact:

```json
{
  "version": 1,
  "selectors": [
    { "type": "artifact" },
    { "type": "page", "artifact_page": 2 },
    {
      "type": "region",
      "unit": "normalized",
      "points": [
        { "x": 0.1, "y": 0.1 },
        { "x": 0.4, "y": 0.1 },
        { "x": 0.4, "y": 0.3 },
        { "x": 0.1, "y": 0.3 }
      ]
    }
  ]
}
```

## 3.5 `time_range` selector

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

## 3.6 `text_quote` selector

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

## 3.7 Future selectors

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

## 3.8 Locator invariants

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
13. An `artifact` selector, when present, is the first (outermost) selector and may stand alone or be followed by narrowing selectors (`page`, `region`, …).
14. On paginated Artifacts, `region` must follow `page` (a polygon is always relative to a chosen Artifact page).

Conceptually:

```text
Citation = Artifact + complete composable locator
```

UI hierarchy or containment can be derived from locator selector chains when useful without making Citation hierarchy part of the persisted evidence model.

---

# 4. Subject vocabulary

## 4.1 `subject_types`

Subject types define the semantic category of a subject. They are data rather than a database enum so the vocabulary can be extended without schema migrations.

```sql
CREATE TABLE subject_types (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    key             TEXT NOT NULL,
    origin          TEXT NOT NULL,             -- provenencia | user | plugin:<id>
    label                   TEXT NOT NULL,
    description             TEXT,
    ref_prefix              TEXT NOT NULL UNIQUE,
    candidate_ref_prefix    TEXT NOT NULL UNIQUE,

    UNIQUE (key, origin)
) STRICT;
```

`origin` and `UNIQUE (key, origin)` follow [`seeded-vocabulary.md`](seeded-vocabulary.md) §1.1. Domain rows reference `subject_types.id`, not bare `key`, so a later product seed can reuse a `key` already taken by a user or plugin term. Application recognition of shipped types looks up `(key, origin = 'provenencia')`.

A Subject type carries **two** ref prefixes, because this vocabulary is shared with the Conclusion layer: `canonical_entities.subject_type_id` references the same rows. For `person`, `ref_prefix` is `PER` (the Conclusion handle, `PER-7KD45`) and `candidate_ref_prefix` is `CPR` (the Interpretation subject, `CPR-7KD45`). Both are required when defining a Subject type, including researcher-defined types.

There is one ref format; the layers are told apart by prefix, not by shape. Candidate prefixes conventionally begin with `C`, but nothing enforces that — a prefix's layer is a registry lookup, not a property of the string.

Prefixes are three uppercase ASCII letters and must not use the reserved catalog prefixes `USR`, `SRC`, `ART`, `CIT`, `OBS`. Both columns draw from **one namespace**: a prefix must be globally unique among Subject types (across origins) and across *both* columns, so no `ref_prefix` may equal another type's `candidate_ref_prefix`. The two `UNIQUE` constraints above are necessary but not sufficient; enforce the cross-column check in the write path.

The Subject type `source` (reification of a Source row) must not use `SRC`; the distinct prefixes `SRN` and `CSR` keep Source catalog refs (`SRC-…`) distinguishable from source-subjects in speech.

Application semantics attach to stable `key` values within an origin rather than a separate built-in flag. Horizon Subject types and their prefixes are catalogued in [`seeded-vocabulary.md`](seeded-vocabulary.md). That set is expected to include at least `person`, `event`, `place`, `relationship`, `participation`, `location`, and `source`.

These names describe application semantics, not different SQL structures. Every instance is stored in the same `subjects` table.

`relationship`, `participation`, and `location` are examples of bridge-like Subjects. `source` is a reification subject: it lets the Interpretation graph talk about evidentiary objects, not only historical persons, events, and places. Structurally, however, the database does not distinguish these categories. Any subject can be related to any other subject through a Subject-valued Observation. The application vocabulary defines what those relationships mean and which combinations are semantically useful.

Inferred associations that no Source asserted (for example an extra Location grain) belong on Conclusion handles and Reconciliation Claims, not as extra Observations on the wrong Citation. See [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md).

## 4.2 `subjects`

A subject gives stable identity to a source-local thing or association encountered during interpretation.

```sql
CREATE TABLE subjects (
    id              BLOB PRIMARY KEY,
    ref             TEXT UNIQUE NOT NULL,      -- e.g. CPR-7KD45
    source_id       BLOB NOT NULL REFERENCES sources(id),
    subject_type_id BLOB NOT NULL REFERENCES subject_types(id),
    label           TEXT,
    description     TEXT,

    UNIQUE (id, subject_type_id)
) STRICT;
```

`ref` is required. It is minted as `{candidate_ref_prefix}-{token}` from the subject's type. Users talk about a person subject as a candidate person (`CPR-…`), distinct from the canonical Person (`PER-…`).

`UNIQUE (id, subject_type_id)` exists so Sameness Claims can use a composite foreign key that pins both endpoints to the same type. It is redundant with the primary key for uniqueness of `id`; it does not allow two types per subject.

`subject_type_id` is immutable after insert. Correcting a wrong type means a new subject (and new `ref`), not an UPDATE of the type. The UUID remains the machine identity; the type (via `candidate_ref_prefix`) is part of the public identity encoded in `ref`.

`source_id` is the subject's home Source — typically the Source being interpreted when the subject was created. It exists so the application can efficiently surface Subjects that belong with a given Source during common same-source workflows. It does not restrict which Citations or Observations may reference the subject; cross-source Observations remain valid.

For Subjects of type `source`, `source_id` has a stronger meaning: it identifies the Source this subject reifies. The application should maintain at most one `source` subject per Source row. Other Sources may then make cited Observations about that subject—bare references such as “this book mentions that marriage certificate,” free-text remarks about authenticity or errors, or both—without collapsing that material into Observations about historical persons alone, and without requiring structured Source-quality columns.

Subjects deliberately contain little domain data. A person's name, an Event's date, a Place's name, or a Participation's role belongs in cited Observations rather than fixed subject columns.

`description` is an optional short summary of the subject itself. Subjects do not have a multi-note table; researcher commentary about interpreted assertions belongs on the supporting Observations and can be aggregated from `observation_notes` when a subject-centric view is needed.

A subject can therefore be sparse. Creating a `person` subject does not require knowing a name, date, or any other property.

---

# 5. Property vocabulary

## 5.1 `properties`

Properties are first-class, user-extensible definitions of predicates that may appear in Observations.

```sql
CREATE TABLE properties (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    key             TEXT NOT NULL,
    origin          TEXT NOT NULL,             -- provenencia | user | plugin:<id>
    label           TEXT NOT NULL,
    description     TEXT,
    value_type      TEXT NOT NULL,

    UNIQUE (key, origin),
    CHECK (value_type IN (
        'text',
        'integer',
        'date',
        'name',
        'subject',
        'term'
    ))
) STRICT;
```

`origin` and `UNIQUE (key, origin)` follow [`seeded-vocabulary.md`](seeded-vocabulary.md) §1.1. Observations and Reconciliation Claims reference `properties.id`, not bare `key`.

A Property's `value_type` is intrinsic to the Property. Seeded Properties (for example `name`, `event_type`, `role`, `person`, `mentions`, `remark`) and their `subject_type_fields` bindings are listed in [`seeded-vocabulary.md`](seeded-vocabulary.md).

Product value types are **`text`**, **`integer`**, **`date`**, **`name`**, **`subject`**, and **`term`**. `real` and `boolean` are not used.

`value_type = 'term'` means the Observation value is a row in `property_terms` for that Property (§5.1.1). Kind and edge Properties (`event_type`, `role`, `relationship_type`) use `term`.

**Who may create a Property with `value_type = term`:** product/plugin Install (Interpretation subject registry) only. Researcher Property create (`origin=user`) must **not** offer or accept `term` — those Properties stay `text` / `integer` / `date` / `name` / `subject`. Researchers may still mint additional **term rows** (`origin=user`) under an existing registry term Property via the composer picker.

The semantic vocabulary is open, but the primitive value system is intentionally constrained. A researcher may define a new Property without introducing a new storage type.

`value_type = 'date'` always means the shared structured DateValue model in [`structured-date-model.md`](structured-date-model.md), not a SQL date or free-text date string.

`value_type = 'name'` always means the shared structured NameValue model in [`structured-name-model.md`](structured-name-model.md), not a single undifferentiated text string. A NameValue always has a full-form `form` and may optionally include ordered parts with product-registry part types for search and reconciliation.

`name_format` is primarily a Conclusion Property (Reconciliation Claim on a person entity). It need not appear in `subject_type_fields` for Interpretation unless a Source itself asserts a naming convention.

Application semantics attach to stable `key` values within an origin, matching `subject_types`. Seeded Properties may receive first-class application behavior. User-defined Properties remain first-class persisted data and can be generically displayed, searched, audited, synced, and referenced. Plugins may add specialized semantics for additional Properties later under `plugin:<plugin_id>` origins.

### 5.1.1 `property_terms`

Categorical values for Properties with `value_type = 'term'`. Same origin rules as other vocabulary-definition tables ([`seeded-vocabulary.md`](seeded-vocabulary.md) §1.1).

```sql
CREATE TABLE property_terms (
    id           BLOB PRIMARY KEY,
    property_id  BLOB NOT NULL REFERENCES properties(id),
    key          TEXT NOT NULL,
    origin       TEXT NOT NULL,             -- provenencia | user | plugin:<id>
    label        TEXT NOT NULL,
    description  TEXT,

    UNIQUE (property_id, key, origin)
) STRICT;
```

Domain rows (Observations, and later Reconciliation Claims when term-valued) reference `property_terms.id`, never bare `key`. Application recognition of shipped terms looks up `(property key, term key, origin = 'provenencia')`.

**Policy (product default for kind/edge Properties):** Install (registry) is the only path that creates Properties with `value_type = term`. Install seeds a **large** product term set so `other` is rare. Researchers may mint additional `origin=user` **term rows** under those Properties (composer picker **Add custom…**, with rename/delete when unused — not a per-vocabulary sidebar destination). Product/plugin terms are not researcher-editable. First-class behavior (capabilities) is declared in the Interpretation subject registry on recognized terms; user term rows remain valid identity without app specialization.

`subjects.label` stays a free working handle on the Evidence graph. It is **not** event-type or role identity.

Interactions between `source` subjects should stay deliberately lightweight. Provenencia does not seed a fine-grained source-quality ontology (`is_authentic`, defect codes, and similar) as Observation Properties.

**First-class Source credibility** — the researcher's reusable three-point trust rating of a Source — lives in `source_credibility_assessments`, not on the `sources` catalog row and not as a required Observation. Authoritative rules: [`research-judgment-model.md`](research-judgment-model.md). Cited commentary about a Source (another Source challenges authenticity, bare references, free-text remarks) continues to use Observations on the Source subject:

```text
mentions   -> node   # bare source-to-source reference; target is typically a source subject
remark     -> text   # free-text commentary about a source subject
```

A book that merely cites a marriage certificate can record `BookSource -- mentions --> CertificateSource` with no remark. A letter that challenges a certificate can add text `remark` Observations and, when needed, ordinary person-level Observations as well. Structured Properties beyond this may be added later only if a concrete workflow requires them. First-class credibility grades: §5.3.

## 5.2 `subject_type_fields`

This table defines which Properties are valid for which Subject types. It is a join table with no `origin` of its own.

```sql
CREATE TABLE subject_type_fields (
    subject_type_id BLOB NOT NULL REFERENCES subject_types(id) ON DELETE CASCADE,
    property_id     BLOB NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    sort_order      INTEGER NOT NULL,

    PRIMARY KEY (subject_type_id, property_id)
) STRICT;
```

`sort_order` orders Add-property menus and Subject fields lists. Locked bindings (cannot unbind while required for connect macros, Conclusion ordering, or other product integrity) live in the compiled Interpretation subject registry (`subjectvocab`), not as a column here.

For example (full matrix in [`seeded-vocabulary.md`](seeded-vocabulary.md)):

```text
person        -> name
person        -> sex_at_birth  # term

event         -> event_type   # term
event         -> date         # locked (point / ordering)
event         -> start_date   # locked (span start)
event         -> end_date     # locked (span end)

place         -> toponym

participation -> person
participation -> event
participation -> role         # term

location      -> event
location      -> place

relationship  -> person             # locked; who is the X
relationship  -> related_to         # locked; …of this person
relationship  -> relationship_type  # term (directed)

source        -> mentions
source        -> remark
```

This is a vocabulary/schema relationship, not historical research data. It says that `participation.person` is a meaningful shape in the Interpretation graph; it does not assert that any particular Person participated in any particular Event.

The vocabulary should be seeded with common definitions but remain researcher-extensible.

For subject-valued Properties, allowed **target** Subject types (for example, `participation.person` should target a `person` subject) are an **application invariant** for now — the same posture as Observation value population. Seeded Properties get first-class UI/validation behavior; user-defined node Properties may remain unconstrained or warn-only. Provenencia does not persist target-type allow-lists in SQL yet (no `target_subject_type_id` on `properties`, and no target join table). That can be added later if pickers and importers need a shared declarative vocabulary.

Malformed edges (wrong target type) may be warned about or ignored by typed workflows; the generic graph still stores the Observation.

## 5.3 `source_credibility_assessments`

Interpretation-layer entity for the researcher's working credibility grade of a Source. Semantics and grade vocabulary: [`research-judgment-model.md`](research-judgment-model.md). Seed keys: [`seeded-vocabulary.md`](seeded-vocabulary.md) §3.0.

```sql
CREATE TABLE source_credibility_grades (
    id          BLOB PRIMARY KEY,              -- UUIDv7, 16 bytes
    key         TEXT NOT NULL,
    origin      TEXT NOT NULL,                 -- provenencia | user | plugin:<id>
    label       TEXT NOT NULL,
    sort_order  INTEGER NOT NULL,

    UNIQUE (key, origin)
) STRICT;

CREATE TABLE source_credibility_assessments (
    id                      BLOB PRIMARY KEY,
    source_id               BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    credibility_grade_id    BLOB NOT NULL REFERENCES source_credibility_grades(id),
    argument                TEXT,

    UNIQUE (source_id)
) STRICT;
```

At most one working assessment per Source. Updates are audited. Missing assessment may display as baseline (`standard` under `origin = 'provenencia'`) without inserting a row.

---

# 6. Relationships between Subjects

There is intentionally no separate table containing historical graph edges.

A relationship between two Subjects is an Observation whose Property has `value_type = 'subject'`:

```text
subject -- Property --> object subject
```

For example:

```text
Participation PT1 -- person --> Person P1
Participation PT1 -- event  --> Event E1
Participation PT1 -- role   --> "subject"
```

or, where the evidence gives only an indeterminate/general association:

```text
Relationship R1 -- person --> Person P1
Relationship R1 -- related_to --> Person P2
Relationship R1 -- relationship_type --> "cousin"
# meaning: P1 is the cousin of P2 (symmetric; either orientation is fine)
```

The graph stores only explicit normalized interpretation. It does not need to persist an additional `father_of` edge if application logic can infer that relationship from a birth Event and its Participations.

This allows the same structural model to represent nuclear family roles, extended kinship, step relationships, guardianship, employment, friendship, household roles, and unanticipated historical associations without adding bridge tables.

The same mechanism covers source-to-source evidence. That includes bare references and optional free-text commentary:

```text
Source subject SB1 (reifies book B)
Source subject SC1 (reifies marriage certificate C)
Person subject P1

Book Citation B1 supports a bare reference:
  SB1 -- mentions --> SC1

Letter Citation L1 supports commentary and historical-world facts:
  SC1 -- remark --> "birth certificate is fake"
  # or, when the document is trusted but a fact is disputed:
  SC1 -- remark --> "date of birth on certificate mistyped"
  P1  -- birth_date --> DateValue(2 JAN 1800)          # polarity positive
  P1  -- birth_date --> DateValue(1 JAN 1800)          # polarity negative
```

`mentions` is a Subject-valued edge and need not say anything further about the referenced Source. `remark` is ordinary text-valued Interpretation of what a citing Source appears to say about another Source. Neither is a Source-layer column or a closed authenticity taxonomy. Historical-world Observations remain separate and independently citable.

---

# 7. `observations`

An Observation is the atomic unit of interpreted evidence. Every Observation is independently addressable and is supported by one Citation.

Conceptually:

```text
Observation {
    citation
    subject
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

The intended primitive value categories are:

```text
text
integer
date
name
subject
term
```

(`real` and `boolean` are not product value types.)

A provisional SQL shape is:

```sql
CREATE TABLE observations (
    id              BLOB PRIMARY KEY,
    ref             TEXT UNIQUE NOT NULL,      -- e.g. OBS-2F8Q1
    citation_id     BLOB NOT NULL REFERENCES citations(id),
    subject_id     BLOB NOT NULL REFERENCES subjects(id),
    property_id     BLOB NOT NULL REFERENCES properties(id),
    polarity        TEXT NOT NULL DEFAULT 'positive',

    value_text      TEXT,
    value_integer   INTEGER,
    value_date_id   BLOB REFERENCES date_values(id),
    value_name_id   BLOB REFERENCES name_values(id),
    value_subject_id BLOB REFERENCES subjects(id),
    value_term_id   BLOB REFERENCES property_terms(id),

    CHECK (polarity IN ('positive', 'negative'))
) STRICT;

CREATE UNIQUE INDEX observations_value_date_id_uidx
    ON observations(value_date_id) WHERE value_date_id IS NOT NULL;
CREATE UNIQUE INDEX observations_value_name_id_uidx
    ON observations(value_name_id) WHERE value_name_id IS NOT NULL;
```

DateValue and NameValue rows are exclusive to one Observation (`000030`). Do **not** unique `value_subject_id` or `value_term_id` — those are shared.

`ref` is a required short human-readable reference for UI and discussion (prefix `OBS`).

**Value population (application invariant):** exactly one value representation must be populated, and it must match `properties.value_type`. Provenencia does **not** enforce that cross-table rule in SQLite for now (no XOR/`value_type` trigger or typed bridge tables). Writers — repository code, importers, sync — are responsible for correct population.

**Read-side tolerance:** if a row is malformed, readers should prefer the column that matches the Property's `value_type` and ignore any other non-null value columns. A row with *no* usable value for that `value_type` is invalid and should be surfaced as an error or omitted rather than guessed. Reconciliation Claims use the same sparse-column idea, except `value_type = 'subject'` is stored as `value_entity_id` (a canonical handle) rather than `value_subject_id`. See [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md).

Reading uncertainty belongs in Citation `transcription` / `transcription_uncertain` / `transcription_note` (and description when needed). Alternative or competing interpretations are modeled as separate Observations rather than a numeric confidence score on a single Observation row. Researcher commentary about a particular Observation belongs in `observation_notes`. Epistemic confidence about a *conclusion* belongs on Claims; see [`research-judgment-model.md`](research-judgment-model.md).

```sql
CREATE TABLE observation_notes (
    id              BLOB PRIMARY KEY,
    observation_id  BLOB NOT NULL REFERENCES observations(id) ON DELETE CASCADE,
    body            TEXT NOT NULL
) STRICT;
```

Notes use a typed table with a real foreign key rather than a polymorphic notes table. Creation, edit, and deletion attribution belong to audit history.

For subject-valued Observations, the object subject must exist and should satisfy any target Subject type constraint defined by the vocabulary. Polarity applies to subject-valued Observations as well: a negative Observation denies that particular edge rather than deleting or omitting it.

A single Citation may support many atomic Observations:

```text
Citation C1
  -> Person P1 -- name       --> NameValue(form = "William Robins", …)
  -> Person P1 -- occupation --> "Carpenter"
  -> Person P1 -- birth_date --> DateValue(...)
```

Multiple Observations may also make different assertions about the same Property without forcing a single value onto the subject. Conflicting or alternative interpretations therefore remain independently citable and auditable. Multiple name Observations on one person subject are expected when Sources use different forms, nicknames, or name changes.

---

# 8. Interpretation-layer invariants

The current design aims to preserve these invariants:

1. Every subject has exactly one Subject type. `subject_type_id` is immutable after insert.
2. Every subject has a required `ref` of the form `{candidate_ref_prefix}-{token}`.
3. Every subject has a home `source_id`; that home Source does not confine which Observations may target the subject.
4. A subject of type `source` reifies the Source identified by its `source_id`; the application should keep at most one such subject per Source.
5. Subject types and Properties are extensible persisted vocabulary with `origin` namespaces and `UNIQUE (key, origin)`, not closed application enums. A new Subject type includes a globally unique `ref_prefix` that is not a reserved catalog prefix or layer code. See [`seeded-vocabulary.md`](seeded-vocabulary.md) §1.1.
6. Every Observation has its own stable identity and a required `OBS-…` `ref`.
7. Every Observation is supported by exactly one Citation.
8. Every Citation has a required `CIT-…` `ref`.
9. Every Observation has exactly one subject and one Property.
10. Every Observation has exactly one typed value.
11. Every Observation has an explicit polarity of `positive` or `negative`; absence of an Observation is not negation.
12. The Observation value must match the Property's declared `value_type` (application write rule). Readers prefer the matching column if multiple value columns are set.
13. Date-valued Observations reference a structured DateValue; they do not store SQL dates or free-text dates as the typed value.
14. Name-valued Observations reference a structured NameValue; they do not store an undifferentiated name string as the typed value.
15. A Property used on a subject must be allowed for that subject's Subject type.
16. A Subject-valued Observation forms a graph edge and its object subject must exist.
17. Application logic may constrain the target Subject type of subject-valued Properties for seeded vocabulary; that is not enforced as SQL allow-lists in this draft.
18. Citation text/description preserves the evidence representation; Observations contain normalized interpretation.
19. Derived genealogical semantics are not duplicated into the Interpretation graph merely for convenience.
20. Unknown/custom Subject types, Properties, and Property terms (including `origin=user` terms) remain preservable and generically usable without first-class application support; capabilities attach only to recognized `(key, origin)` pairs.
21. First-class application behavior may recognize seeded `(key, origin = 'provenencia')` pairs and open values; it must not require the schema to close those vocabularies or encode every genealogical edge case.
22. Source credibility assessments are Interpretation entities (`source_credibility_assessments`), not columns on `sources` and not Observation confidence scores.
23. Citation transcription certainty is the boolean `transcription_uncertain` (+ optional note), media-agnostic; it is not Claim confidence.

---

# 9. Documentation ownership

To avoid competing schema definitions:

- [`source-layer-data-model.md`](source-layer-data-model.md) is authoritative for Source-layer tables and Artifact/File storage.
- This document is authoritative for Interpretation-layer tables and vocabulary schema.
- [`seeded-vocabulary.md`](seeded-vocabulary.md) is the horizon catalog for intended Subject types, Properties, bindings, and open-value starters (not a v1 ship list).
- [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md) is authoritative for Conclusion-layer tables and Claims.
- [`structured-date-model.md`](structured-date-model.md) is authoritative for shared DateValue persistence.
- [`structured-name-model.md`](structured-name-model.md) is authoritative for shared NameValue persistence.
- [`audit-revision-history.md`](audit-revision-history.md) is authoritative for audit and revision history.
- [`research-judgment-model.md`](research-judgment-model.md) is authoritative for Source credibility, Citation transcription certainty, and Claim confidence semantics.
- [`data-model-source-interpretation-conclusion.md`](data-model-source-interpretation-conclusion.md) summarizes the three-layer philosophy and cross-layer examples.
