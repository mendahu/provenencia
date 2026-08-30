# Provenencia Genealogy — Source Layer Data Model

## Status

Draft architecture notes. This document is the authoritative schema and design reference for the Provenencia Source layer.

The Source layer answers:

> What evidence do we possess?

It preserves evidence as acquired and descriptive catalog information about that evidence while deliberately avoiding genealogical interpretation. People, events, places, relationships, transcriptions, identity resolution, and conclusions belong to later layers.

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

Dates are the intentional exception: date-valued metadata may additionally reference Provenencia's shared structured genealogical date representation so dates can be sorted and filtered. That cross-layer value model is defined in [`structured-date-model.md`](structured-date-model.md).

## 1.3 Offline-first ingestion

Digital evidence added to Provenencia is ingested into application-managed local storage. External URLs and locations may be retained as Source provenencia, but they are not runtime dependencies for opening an Artifact.

## 1.4 Immutable digital objects

Stored File bytes are immutable and content-addressed. Replacing a scan means ingesting a new File and updating the Artifact reference. Previously referenced primary Files remain preserved for historical reconstruction.

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
            |                           ^
            |                           |
            |               source_type_metadata_fields
            |                           |
            +---------------------------+
            |
            +--< artifacts >-- files
                                |
                                +--< file_derivatives >-- files

source_metadata.date_value_id --> date_values (shared cross-layer model)
```

The File store is shared infrastructure, but it is defined here because Artifacts are its primary Source-layer consumer. The `date_values` table is not defined here; see [`structured-date-model.md`](structured-date-model.md).

---

# 3. `source_types`

Source types are a controlled but extensible vocabulary used primarily for search and filtering.

```sql
CREATE TABLE source_types (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    key             TEXT UNIQUE NOT NULL,
    label           TEXT NOT NULL,
    description     TEXT,
    builtin         INTEGER NOT NULL DEFAULT 0,

    CHECK (builtin IN (0, 1))
) STRICT;
```

New projects may be seeded with a small set of common types such as `birth_certificate`, `census`, `photograph`, and similar, growing as real cataloging needs appear. The horizon catalog (and suggested metadata fields) lives in [`seeded-vocabulary.md`](seeded-vocabulary.md).

Users may add project-specific types without schema changes. Built-in types are defaults, not an enum and not structurally privileged subclasses.

A source type does not imply a specialized table or interpretation behavior.

---

# 4. `sources`

A Source is the canonical evidentiary object.

```sql
CREATE TABLE sources (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    ref             TEXT UNIQUE NOT NULL,      -- e.g. SRC-F4N2P
    source_type_id  BLOB NOT NULL REFERENCES source_types(id),
    title           TEXT,
    description     TEXT
) STRICT;
```

`source_type_id` provides the broad user-facing classification. More variable catalog information belongs in Source metadata. All Sources have a `ref` with prefix `SRC`; Source *type* (`birth_certificate`, `census`, …) does not change the prefix.

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

Most fields are text. Structured dates are the deliberate exception.

```sql
CREATE TABLE source_metadata_fields (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    key             TEXT UNIQUE NOT NULL,
    label           TEXT NOT NULL,
    data_type       TEXT NOT NULL DEFAULT 'text',
    description     TEXT,
    builtin         INTEGER NOT NULL DEFAULT 0,

    CHECK (data_type IN ('text', 'date')),
    CHECK (builtin IN (0, 1))
) STRICT;
```

Possible seeded fields include `author`, `publisher`, `publication_date`, and similar. The authoritative list is in [`seeded-vocabulary.md`](seeded-vocabulary.md).

The goal is not to create a general typed EAV system. New structured types should only be introduced for concrete use cases.

## 5.2 `source_type_metadata_fields`

A Source type may suggest metadata fields useful for that type.

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
    value_text      TEXT,
    date_value_id   BLOB REFERENCES date_values(id)
) STRICT;
```

Ordinary metadata preserves text directly:

```text
author
  value_text = "Alice Smith and Robert Jones"
```

Date metadata may preserve both the entered/source wording and a structured representation:

```text
publication_date
  value_text = "about the year 1890"
  date_value_id = DateValue(ABT 1890)
```

The text remains useful for fidelity even when a structured date exists.

The schema and semantics of `date_values` are defined in [`structured-date-model.md`](structured-date-model.md).

Application validation should enforce the intended relationship between `source_metadata_fields.data_type` and `date_value_id`; SQLite cannot express that cross-table constraint with a simple `CHECK`.

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
storage: objects/8f/ce/8fce3b...
```

The exact sharding convention (`objects/{first two hex}/{next two}/{full hex}`) is fixed as above unless a migration notes otherwise.

The stored object is the **original file bytes** (JPEG, PNG, PDF, …). Hash names do not encrypt or wrap the payload. A user who opens the object in a normal viewer sees the picture or document. The ingest filename is `files.original_filename`, not the path on disk.

This keeps projects relocatable and prevents `storage_path` and checksum from becoming competing sources of truth.

## Original filename

`original_filename` preserves the filename presented at ingestion while the actual object-store name is checksum-derived.

Generated Files such as thumbnails may have `original_filename = NULL`.

## Immutability and replacement

File bytes never change in place. A different byte stream produces a different checksum and therefore a different File.

Replacing an Artifact's scan is:

```text
CREATE File B
UPDATE Artifact.file_id: File A -> File B
```

The operation is recorded through the audit system. File A remains stored so historical Artifact state can still be materialized and viewed.

The initial implementation does not support destructive deletion of primary Files.

---

# 7. `artifacts`

An Artifact is a concrete evidentiary representation of a Source.

```sql
CREATE TABLE artifacts (
    id              BLOB PRIMARY KEY,          -- UUIDv7, 16 bytes
    ref             TEXT UNIQUE NOT NULL,      -- e.g. ART-3K9M2
    source_id       BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    file_id         BLOB REFERENCES files(id),
    description     TEXT
) STRICT;
```

`ref` is required so Artifacts can be named in discussion independently of their Source (`ART-3K9M2` under `SRC-F4N2P`).

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

For an Artifact whose File is replaced:

```text
Revision 100
  Artifact.file_id: NULL -> File A

Revision 145
  CREATE File B
  Artifact.file_id: File A -> File B

Revision 212
  CREATE File C
  Artifact.file_id: File B -> File C
```

The current row points to File C. Previous File versions are derived from audit history and can be presented directly in the Artifact UI.

No separate `artifact_file_versions` table is required initially. If historical-version queries later become performance-sensitive, a derived projection or cache may be introduced without becoming authoritative state.

Primary File retention must account for historical references, not merely current foreign-key references. Ordinary orphan cleanup must never remove File A or File B in the example above.

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
artifacts
```

with these directly supporting storage tables:

```text
files
file_derivatives
```

`date_values` is shared cross-layer value-object infrastructure and is defined separately in [`structured-date-model.md`](structured-date-model.md).

The audit tables are cross-cutting infrastructure and are defined separately in `audit-revision-history.md`.

---

# 11. Current architectural rules

1. Sources are evidentiary objects and remain free of genealogical interpretation.
2. Source types use a seeded, controlled, user-extensible vocabulary rather than an enum.
3. Source metadata is descriptive and minimally structured.
4. Metadata values are text by default; shared structured genealogical dates are the intentional exception.
5. Source types may suggest metadata fields but do not require them.
6. External provenencia belongs to the Source and must not be required to access ingested evidence.
7. Artifacts are concrete representations of Sources and do not have an `artifact_type` taxonomy.
8. An Artifact has zero or one primary File.
9. Multiple evidentiary representations are multiple Artifacts under the same Source.
10. Digital evidence is ingested into application-managed local storage.
11. Files are content-addressed by SHA-256 and their storage paths are derived rather than persisted.
12. File bytes are immutable.
13. Replacing an Artifact File creates/reuses another File and updates the Artifact under audit.
14. Historically referenced primary Files are retained indefinitely in the initial implementation.
15. Primary File deletion is not supported initially.
16. Historical Artifact File versions are derived from audit history rather than stored in a separate version table.
17. Thumbnails and other generated assets are File derivatives, not Artifacts.
18. File derivatives are reproducible and disposable.
19. Generic creation/update timestamps and user attribution belong to audit history rather than Source-layer rows.
20. Research notes for Sources use a typed `source_notes` table with a real foreign key, not a polymorphic notes table.
21. Sources and Artifacts have required human-readable `ref` values (`SRC-…`, `ART-…`).
22. All tables use SQLite `STRICT` typing.

This schema keeps evidence description, digital representation, storage infrastructure, generated UI assets, and later genealogical interpretation as distinct concerns.