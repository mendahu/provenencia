# Provenencia Genealogy — Source Layer Data Model

## Status

Draft architecture notes. This document is the authoritative schema and design reference for the Provenencia Source layer.

The Source layer answers:

> What evidence do we possess?

It preserves evidence as acquired and descriptive catalog information about that evidence while deliberately avoiding genealogical interpretation. People, events, places, relationships, transcriptions, identity resolution, conclusions, and structured Source **credibility** assessments belong to later layers (credibility: [`research-judgment-model.md`](research-judgment-model.md)).

---

# 1. Source-layer principles

## 1.1 Evidence first

A Source is the evidentiary object. Examples include a birth certificate, census, photograph, oral testimony, DNA match report, book, website capture, GEDCOM file, or family-tree export.

A Source may have zero or more Artifacts. An Artifact is a concrete representation of the Source: a scan, photograph, PDF, audio recording, downloaded export, or a placeholder for a physical representation that has not been digitized.

```text
Source
  ├── Artifact -> File
  ├── Artifact -> File
  └── Artifact -> no File
```

## 1.2 Descriptive, not interpretive

Source metadata is catalog information. It should faithfully preserve useful descriptive values without trying to resolve them into the canonical genealogical graph.

If a book cover says:

```text
Alice Smith and Robert Jones
```

then its `author` metadata may preserve that exact string. The Source layer does not need to split the authors or resolve either one to a canonical Person.

Catalog dates are the same: `publication_date`, `issue_date`, `census_date`, and similar stay as-written text on `source_metadata.value_text` (`"about the year 1890"`, `"15 May 1880"`). They exist so the researcher can file a reference date without parsing it from the Artifact, and so omnibar search can match that wording. They do **not** attach a structured DateValue.

Structured dates belong on Interpretation Observations and later Conclusion claims ([`structured-date-model.md`](structured-date-model.md)). A later project-wide sort or group by when evidence was created or published is a first-class Source attribute, not an EAV metadata key — the printed name of that date changes by type (publication vs issue vs registration vs enumeration). Parked: [`ideas/source-provenance-date.md`](ideas/source-provenance-date.md).

## 1.3 Offline-first ingestion

Digital evidence added to Provenencia is ingested into application-managed local storage. External URLs and locations may be retained as Source provenencia, but they are not runtime dependencies for opening an Artifact.

## 1.4 Immutable digital objects

Stored File bytes are immutable and content-addressed. A different byte stream is always a different File. An Artifact may receive a primary File at most once (create-time or first attach while fileless). A clearer or newer scan of the same document is a **separate Artifact** under the same Source — do not pointer-swap `file_id` under an existing Artifact, because Citations resolve as `artifact_id` + locator into that Artifact’s media.

## 1.5 Generated derivatives are infrastructure

Thumbnails, previews, waveforms, and similar generated assets belong to Files, not Artifacts. They are reproducible application data and may be regenerated or purged.

## 1.6 Strict SQLite schema

All tables use SQLite `STRICT` typing. UUIDv7 identifiers are stored as 16-byte `BLOB` values.

Generic `created_at`, `updated_at`, `created_by`, and `updated_by` bookkeeping does not belong on these domain tables. Creation and modification history is represented by the audit/revision model.

---

# 2. Relationship overview

```text
source_types
    |
    +--< sources
            |
            +--< source_notes
            |
            +--< source_metadata >-- source_metadata_fields
            |                                  ^  ^
            +--< source_metadata_layout >------+  |
            |                                     |
            |                         source_type_metadata_fields
            |                                     |
            +-------------------------------------+
            |
            +--< artifacts >-- files
                                |
                                +--< file_derivatives >-- files
```

The File store is shared infrastructure, but it is defined here because Artifacts are its primary Source-layer consumer.

---

# 3. `source_types`

Source types are a controlled but extensible vocabulary used primarily for search and filtering.

```sql
CREATE TABLE source_types (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    key             TEXT NOT NULL,
    origin          TEXT NOT NULL,             -- provenencia | user | plugin:<id>
    label           TEXT NOT NULL,
    description     TEXT,

    UNIQUE (key, origin)
) STRICT;
```

`origin` is the vocabulary namespace (product seed, researcher, or future plugin). Rules and reserved values: [`seeded-vocabulary.md`](seeded-vocabulary.md) §1.1. Uniqueness is `(key, origin)` so a later product seed can share a `key` with an existing user or plugin term without colliding. Domain rows (`sources.source_type_id`) reference the UUID `id`, not bare `key`.

New projects may be seeded at create time with a small starter (today: `birth_certificate` and a few suggested fields; `origin = 'provenencia'`). The horizon catalog (and suggested metadata fields) lives in [`seeded-vocabulary.md`](seeded-vocabulary.md). Existing catalogs are not healed or backfilled on open.

Users may add project-specific types (`origin = 'user'`) without schema changes. Product-seeded types are create-time defaults, not an enum and not structurally privileged subclasses beyond optional first-class UX for well-known keys. Any origin may be deleted when unused.

A source type does not imply a specialized table or interpretation behavior.

---

# 4. `sources`

A Source is the canonical evidentiary object.

```sql
CREATE TABLE sources (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    ref             TEXT UNIQUE NOT NULL,      -- e.g. SRC-F4N2P
    source_type_id  BLOB NOT NULL REFERENCES source_types(id),
    title           TEXT NOT NULL,
    description     TEXT
) STRICT;
```

`source_type_id` provides the broad user-facing classification. `title` is the required human headline for lists and search (always non-empty after trim). More variable catalog information belongs in Source metadata. All Sources have a `ref` with prefix `SRC`; Source *type* (`birth_certificate`, `census`, …) does not change the prefix.

`description` is catalog text about the Source itself. Researcher commentary that may accumulate over time belongs in `source_notes` rather than a single inline notes field.

External provenencia such as an originating URL belongs conceptually to the Source rather than the Artifact. The exact Source-level acquisition/provenencia representation can be refined separately when its use cases require more structure.

## 4.1 `source_notes`

Research notes attached to a Source. Multiple notes are allowed so commentary can accumulate without overwriting earlier remarks.

```sql
CREATE TABLE source_notes (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    source_id       BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    body            TEXT NOT NULL
) STRICT;
```

Notes use a typed table with a real foreign key rather than a polymorphic notes table. Creation, edit, and deletion attribution belong to audit history.

---

# 5. Source metadata

## 5.1 `source_metadata_fields`

Metadata fields form a controlled but extensible vocabulary.

Most fields are text, including catalog dates. `url` is text-shaped (still stored
in `value_text`) and exists so clients can treat those fields as external links.

```sql
CREATE TABLE source_metadata_fields (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    key             TEXT NOT NULL,
    origin          TEXT NOT NULL,             -- provenencia | user | plugin:<id>
    label           TEXT NOT NULL,
    data_type       TEXT NOT NULL DEFAULT 'text',
    description     TEXT,

    UNIQUE (key, origin),
    CHECK (data_type IN ('text', 'url'))
) STRICT;
```

Same `origin` / `UNIQUE (key, origin)` rules as `source_types` ([`seeded-vocabulary.md`](seeded-vocabulary.md) §1.1). `source_metadata.field_id` and `source_type_metadata_fields.field_id` reference the UUID `id`. **`data_type` is immutable after create** so existing `source_metadata` values stay consistent with validation.

Possible seeded fields include `author`, `publisher`, `publication_date`, and similar. Date-named keys are still `data_type = 'text'`. The authoritative list is in [`seeded-vocabulary.md`](seeded-vocabulary.md).

The goal is not to create a general typed EAV system. New structured types should only be introduced for concrete use cases. Do not reintroduce `data_type = 'date'` on catalog fields so a later sort can reuse those keys — that sort cannot be global across `publication_date` / `issue_date` / `census_date` without a first-class attribute ([`ideas/source-provenance-date.md`](ideas/source-provenance-date.md)).

## 5.2 `source_type_metadata_fields`

A Source type may suggest metadata fields useful for that type. This is a join table, not a vocabulary-definition table: it has no `origin` of its own; both ends already carry `origin`.

```sql
CREATE TABLE source_type_metadata_fields (
    source_type_id  BLOB NOT NULL REFERENCES source_types(id) ON DELETE CASCADE,
    field_id        BLOB NOT NULL REFERENCES source_metadata_fields(id) ON DELETE CASCADE,
    sort_order      INTEGER,

    PRIMARY KEY (source_type_id, field_id)
) STRICT;
```


For example, `book` might suggest:

```text
author
publisher
publication_date
edition
isbn
```

These are UI/cataloging suggestions, not mandatory fields. A Source may use metadata fields not associated with its Source type.

## 5.3 `source_metadata`

```sql
CREATE TABLE source_metadata (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    source_id       BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    field_id        BLOB NOT NULL REFERENCES source_metadata_fields(id),
    value_text      TEXT
) STRICT;
```

Metadata preserves text directly, including catalog dates:

```text
author
  value_text = "Alice Smith and Robert Jones"

publication_date
  value_text = "about the year 1890"
```

There is no `date_value_id` on this table. Structured DateValues are Interpretation (and later Conclusion) values, not catalog filing. Application validation should enforce that `data_type = 'url'` and `'text'` both store in `value_text` only.

## 5.4 `source_metadata_layout`

Type suggestions (§5.2) are shared by every Source of a type, and `source_metadata` rows exist only once a field holds a value. Neither can record how one Source's metadata area should look. That presentation state — which suggestions this Source has waved off, and in what order its fields read — is per-Source and belongs to its own table.

```sql
CREATE TABLE source_metadata_layout (
    source_id       BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    field_id        BLOB NOT NULL REFERENCES source_metadata_fields(id) ON DELETE CASCADE,
    sort_order      INTEGER NOT NULL,
    dismissed       INTEGER NOT NULL DEFAULT 0 CHECK (dismissed IN (0, 1)),

    PRIMARY KEY (source_id, field_id)
) STRICT;
```

A row may exist for a field with no value, which is what makes a dismiss durable. Rows are sparse: a Source the researcher has never reordered or dismissed from has none at all.

`ListWorkspace` is the single reader that merges the three tables into the Source page metadata area:

- **Dismiss** — a suggested field with `dismissed = 1` is omitted while it holds no value. A field that has a value stays visible regardless, so dismissing an already-filled field is a no-op rather than a way to hide catalog data.
- **Order** — once a Source has any layout row, entries are returned by `sort_order`; fields with no row yet sort after them, keeping suggestion-then-extra order among themselves. A Source with no layout rows keeps the type's suggestion order followed by extra values.

Reorder rewrites `sort_order` as `0..n-1` over the fields it is given and leaves `dismissed` alone. Setting a value for a field with no layout row appends one at the end of the existing order, so filling in a field does not disturb a hand-sorted Source.

Dismiss and reorder are researcher decisions about their own catalog, so both are audited (`dismiss_source_metadata_suggestion`, `reorder_source_metadata`).

---

# 6. `files`

A File is a locally managed immutable digital storage object.

```sql
CREATE TABLE files (
    id                  BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    checksum_sha256     TEXT NOT NULL UNIQUE,
    original_filename   TEXT,
    media_type          TEXT,
    byte_size           INTEGER NOT NULL
) STRICT;
```

## Content-addressed storage

The managed storage location is derived deterministically from `checksum_sha256` rather than persisted in the database.

For example:

```text
SHA-256: 8fce3b...
storage: objects/8f/ce/8fce3b....jpg
```

The basename is the full lowercase hex checksum plus an optional MIME-derived extension (`.jpg`, `.pdf`, …) from `media_type`. Unknown or empty media types keep the bare hex name. The exact sharding convention (`objects/{first two hex}/{next two}/{full hex}{ext}`) is fixed as above unless a migration notes otherwise.

The stored object is the **original file bytes** (JPEG, PNG, PDF, …). Hash names do not encrypt or wrap the payload. A user who opens the object in a normal viewer sees the picture or document. The ingest filename is `files.original_filename`, not the path on disk.

This keeps projects relocatable and prevents `storage_path` and checksum from becoming competing sources of truth.

## Original filename

`original_filename` preserves the filename presented at ingestion while the actual object-store name is checksum-derived.

Generated Files such as thumbnails may have `original_filename = NULL`.

## Immutability and first attach

File bytes never change in place. A different byte stream produces a different checksum and therefore a different File.

An Artifact may gain a primary File while still fileless:

```text
CREATE File A
UPDATE Artifact.file_id: NULL -> File A
```

Once `file_id` is set, it must not be pointer-swapped to another File. A better or clearer digitization is modeled as an additional Artifact under the same Source:

```text
Source
  ├── Artifact A -> File (older photocopy)
  └── Artifact B -> File (newer clear scan)
```

Interpretation may later offer flows to move or duplicate Citations from Artifact A to Artifact B (and adjust locators). That is outside the Source layer.

The initial implementation does not support destructive deletion of primary Files.

---

# 7. `artifacts`

An Artifact is a concrete evidentiary representation of a Source.

```sql
CREATE TABLE artifacts (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    ref             TEXT UNIQUE NOT NULL,      -- e.g. ART-3K9M2
    source_id       BLOB NOT NULL REFERENCES sources(id),
    file_id         BLOB REFERENCES files(id),
    label           TEXT NOT NULL,             -- required list headline
    description     TEXT
) STRICT;
```

`ref` is required so Artifacts can be named in discussion independently of their Source (`ART-3K9M2` under `SRC-F4N2P`). `source_id` is `NO ACTION`: deleting a Source while Artifacts remain fails. Facet rows (notes, metadata) still CASCADE with the Source.

`label` is the required researcher-facing list headline (distinct from optional `description`).

An Artifact has zero or one primary File.

Fileless Artifacts are valid and can represent physical-only evidence:

```text
Source: Smith family Bible
Artifact: physical copy held by Mary Smith
file_id = NULL
```

Multiple evidentiary representations are separate Artifacts under the same Source:

```text
Source: family photograph
  Artifact A -> archival TIFF scan
  Artifact B -> independently retained JPEG representation
```

Artifacts do not have an `artifact_type`. Technical properties such as MIME type, checksum, and byte size belong to the File.

Application-generated thumbnails and previews are not additional Artifacts.

---

# 8. `file_derivatives`

Generated assets are relationships between a source File and another File generated from it.

```sql
CREATE TABLE file_derivatives (
    id                  BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    source_file_id      BLOB NOT NULL REFERENCES files(id),
    derived_file_id     BLOB NOT NULL REFERENCES files(id),
    derivative_type     TEXT NOT NULL,

    UNIQUE (source_file_id, derivative_type),
    CHECK (source_file_id <> derived_file_id)
) STRICT;
```

Initial derivative types may include:

```text
thumbnail
preview
pdf_page_preview
waveform
```

This vocabulary is application infrastructure, not a user-managed genealogy taxonomy unless future requirements justify making it one.

Derivatives belong to Files rather than Artifacts because the Artifact's primary File can change over time:

```text
Artifact A
  Revision 100 -> File A -> Thumbnail T1
  Revision 145 -> File B -> Thumbnail T2
```

A historical Artifact view can therefore resolve the correct derivative from the historical File version.

Derivative Files use the same content-addressed `files` store, but unlike primary Files they are reproducible and disposable. They may be regenerated or purged without changing research history.

Derivative generation and cache eviction do not need to create research audit events.

---

# 9. Audit and historical file versions

The Source-layer tables do not contain generic version columns. The append-only audit system records their mutations.

For an Artifact that receives a primary File:

```text
Revision 100
  Artifact.file_id: NULL -> File A
```

The current row points to File A. There is no product path that later updates that Artifact to File B. A newer scan is a second Artifact (and its own first-attach revision).

Primary Files that remain referenced by any current Artifact (or that must be retained for other product reasons) are not garbage-collected in the initial implementation. Ordinary orphan cleanup must never remove Files that are still referenced.

No separate `artifact_file_versions` table is required. If historical reconstruction of past Artifact rows from audit ever becomes a product need, present it from the audit stream rather than inventing pointer-swap “versions” on Artifacts.

---

# 10. Current Source-layer schema

The Source layer currently consists of:

```text
source_types
sources
source_notes
source_metadata_fields
source_type_metadata_fields
source_metadata
source_metadata_layout
artifacts
```

with these directly supporting storage tables:

```text
files
file_derivatives
```

The audit tables are cross-cutting infrastructure and are defined separately in `audit-revision-history.md`. Structured `date_values` are Interpretation / Conclusion infrastructure ([`structured-date-model.md`](structured-date-model.md)), not a Source-layer attachment.

---

# 11. Current architectural rules

1. Sources are evidentiary objects and remain free of genealogical interpretation. Structured Source **credibility** is an Interpretation assessment entity, not a column on `sources`; see [`research-judgment-model.md`](research-judgment-model.md).
2. Source types and metadata fields use a seeded, controlled, origin-namespaced vocabulary (`UNIQUE (key, origin)`) rather than an enum; see [`seeded-vocabulary.md`](seeded-vocabulary.md) §1.1.
3. Source metadata is descriptive and minimally structured.
4. Metadata values are text (`value_text`); `url` is text-shaped for external-link affordances. Catalog dates are text. Do not attach DateValue (or NameValue) to `source_metadata`.
5. Source types may suggest metadata fields but do not require them; how one Source presents those suggestions (dismissed, ordered) lives in `source_metadata_layout` rather than on the shared type join or on `source_metadata`.
6. External provenencia belongs to the Source and must not be required to access ingested evidence.
7. Artifacts are concrete representations of Sources and do not have an `artifact_type` taxonomy.
8. An Artifact has zero or one primary File.
9. Multiple evidentiary representations are multiple Artifacts under the same Source.
10. Digital evidence is ingested into application-managed local storage.
11. Files are content-addressed by SHA-256 and their storage paths are derived rather than persisted.
12. File bytes are immutable.
13. An Artifact’s primary File may be set only while fileless (first attach); better scans are additional Artifacts under the same Source — not `file_id` pointer-swaps (Citations locate into Artifacts).
14. Primary Files referenced by Artifacts are retained indefinitely in the initial implementation.
15. Primary File deletion is not supported initially.
16. Historical Artifact File versions are derived from audit history rather than stored in a separate version table.
17. Thumbnails and other generated assets are File derivatives, not Artifacts.
18. File derivatives are reproducible and disposable.
19. Generic creation/update timestamps and user attribution belong to audit history rather than Source-layer rows.
20. Research notes for Sources use a typed `source_notes` table with a real foreign key, not a polymorphic notes table.
21. Sources and Artifacts have required human-readable `ref` values (`SRC-…`, `ART-…`).
22. All tables use SQLite `STRICT` typing.

This schema keeps evidence description, digital representation, storage infrastructure, generated UI assets, and later genealogical interpretation as distinct concerns.