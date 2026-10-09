---
name: add-catalog-model
description: >-
  Keeps Provenencia's declared catalog model (core/database/catalogmodel
  Tables, FKs, Kind, Bucket) in step with the live SQLite schema. Use when
  adding, changing, or dropping a catalog table or foreign key, editing
  catalogmodel, registering a Kind or Bucket, or when TestPragmaHonesty
  fails on an unregistered table or FK.
---

# Maintain the catalog model

`core/database/catalogmodel` is the hand-written list of catalog tables and foreign keys. `TestPragmaHonesty` checks it against SQLite (`sqlite_schema`, `PRAGMA foreign_key_list`). It is not the SQLite schema itself (`schemahash.go`, `catalog.schema_mismatch`).

The package imports nothing above it. It does not import `deleteimpact`. Probes, the deletable-kind set, releases, and `Impact` stay in `deleteimpact` — [`add-catalog-delete`](../add-catalog-delete/SKILL.md). DDL lives in a migration — [`add-catalog-migration`](../add-catalog-migration/SKILL.md). FK classification policy: [`docs/catalog-deletes.md`](../../../docs/catalog-deletes.md).

Update the model in the **same PR** as the migration that adds, changes, or drops the table or FK.

## Add

1. One `Tables` row: `Name`, `PK` when the table has one, `Bucket`. Skip `sqlite_%` and `catalog_search_fts%`.
2. A `Kind` constant when delete policy or a probe names the table. The value is the stable machine key (`source`), not the table name (`sources`). Set `Table.Kind` to that constant, and set `PK` to the primary-key column. `tableByKind` returns the row only when that kind is in `deleteimpact`'s deletable set; add it there only for an official delete lookup. The existence query is `SELECT 1 FROM <name> WHERE <pk> = ?`, taken from the table row. Facet, skip, and pool tables usually have a bucket and no kind.
3. One `FKs` row per live FK: `From`, `Column` (the leading column), `To`, `OnDelete`, `Bucket`. A composite key sets `FromCols` to the full column tuple in `PRAGMA` order and is keyed by seq 0 (`identity_claims` `subject_id` + `subject_type_id`). `Audited: true` only on a `CASCADE` whose rows official deletes release and audit. Empty and `RESTRICT` compare as `NO ACTION`.
4. The leading column needs a covering index (explicit, `UNIQUE`, or primary key leftmost).
5. A resource FK needs a list probe, an owned-outbound FK needs a release entry, and an audited `CASCADE` needs one facet release. Those stay in `deleteimpact`.
6. A table outside the skip bucket also needs an `effects` entry in the same PR (`core/database/effects`). One `Effect` per table: `Source`, `Handles`, `Search`, `Vocabulary`, `Structure`. An empty field means that job is unaffected. `None: true` is the explicit empty entry. Skip tables stay out of the registry. A path step that names an edge has to be a real `catalogmodel.FKs` row; package init panics when it is not.

```go
{Name: "citation_notes", Bucket: BucketFacet},
{From: "citation_notes", Column: "citation_id", To: "citations", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
```

Call sites use `catalogmodel.KindCitation`. Do not alias `Kind` or `Bucket` from `deleteimpact`.

## Change

A renamed column or a new `ON DELETE` updates `Column`, `FromCols`, `To`, and `OnDelete` so they match `PRAGMA foreign_key_list`. A new bucket updates `Bucket` and `Audited`, then the probe or release `deleteimpact` requires for that bucket. Do not edit a shipped migration to match the model. The model follows the schema the new SQL file produces.

## Remove

Drop the `Tables` row, its `FKs` rows, and the `Kind` constant when nothing names the table. Drop the kind from `deleteimpact`'s deletable set, and drop its inbound probes, owned releases, facet releases, and projectors, plus any call site that passed the kind. Drop the table's `effects` entry. Do not leave a zero-count probe for a table that is not in the catalog yet.

## Check

```
CGO_ENABLED=1 go test -tags fts5 ./core/database/catalogmodel ./core/database/deleteimpact ./core/database/effects
```

`catalogmodel.TestPragmaHonesty`: every non-FTS catalog table is registered; every live FK matches `To`, `OnDelete`, and `FromCols`; no registered FK is missing; the leading column has a covering index.

`deleteimpact` then checks that a resource FK has a list probe, an owned-outbound FK is on the release list, a deletable kind has a projector section, and each audited `CASCADE` has one facet release. Every deletable kind resolves to one table with a primary key.

`effects` checks that every non-skip table has an entry, and that a `CASCADE` into a table with effects is `Audited`. Package init already panics when a path names a missing foreign key.

## Do not

- Import `deleteimpact`, `audit`, or a domain package from `catalogmodel`.
- Put `Exists` SQL, probes, or releases on `Table` or `FK`.
- Register a kind, table, or FK for a table that does not exist yet.
- Re-export `type Kind = catalogmodel.Kind` or `const KindSource = catalogmodel.KindSource`.
