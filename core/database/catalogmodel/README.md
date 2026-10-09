# catalogmodel

The declared list of catalog tables and foreign keys. SQLite remains the schema (`core/database/migrations`, `schemahash.go`). This package is the classified copy other registries read: table name, primary key, kind, bucket, and each foreign key's target, `ON DELETE` action, and whether a cascade is audited.

It imports nothing else in `core/database`. `TestPragmaHonesty` checks `Tables` and `FKs` against `sqlite_schema` and `PRAGMA foreign_key_list`.

## How the schema packages sit

These are siblings under `core/database`. Each one is its own package so the imports stay one way.

| Package | Role | Imports |
| --- | --- | --- |
| `catalogmodel` | Declared tables and foreign keys | nothing in `core/database` |
| [`rowchange`](../rowchange) | One row's old and new fields | nothing in `core/database` |
| [`deleteimpact`](../deleteimpact) | Whether a delete is allowed, and what it must release | `catalogmodel`, `rowchange` |
| [`effects`](../effects) | What a written change touches | `catalogmodel`, `rowchange` |
| `audit` | Stores the revision | `rowchange` |

`deleteimpact` and `effects` do not import each other. `effects` does not import `audit`. Writers (`sources`, `observations`, and the rest) speak `rowchange` and call `audit.Record`. Official deletes call `deleteimpact`. Nothing calls `effects` yet.

A wrapping directory would group the folders in the tree and leave this direction unchanged. Nesting one of these packages inside another would cycle.
