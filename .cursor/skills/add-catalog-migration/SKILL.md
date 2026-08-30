---
name: add-catalog-migration
description: >-
  Adds a Provenencia SQLite catalog format step (user_version, migrations/NNNNNN.sql).
  Use when adding or changing core/database migrations, CREATE TABLE in provenencia.sqlite,
  user_version, schema format, or when the user asks to add a migration.
---

# Add a catalog migration

This is **not** a web `schema_migrations` / timestamp workflow. There is **no** migrate CLI. `Create` / `Open` apply SQL. Old apps must refuse a **newer** `user_version`.

Product SemVer (`VERSION`) is a separate bump. Do not bump it for a format change unless you are cutting a release.

## Steps

1. List `core/database/migrations/*.sql`. Next file is the next integer, **six digits**: `000002.sql` after `000001.sql`.
2. Create **only** that new file. Do **not** edit a shipped `NNNNNN.sql` (rewrite history of existing folders). Unreleased steps on this branch may still be amended.
3. Put additive DDL in the new file (`CREATE TABLE` / `STRICT`). No Source tables unless the spike/PR explicitly includes them.
4. Add a table-driven test: create or open a catalog at the previous version, `Open` it, assert the new table / `user_version`.
5. Run `CGO_ENABLED=1 go test ./core/database/`. Init panics (and tests fail) on gaps, `1.sql` names, empty SQL, or missing files.
6. Do not add embed vars, `formatVersion` constants, or timestamp filenames. The glob in `migrate.go` is the registry.

## Do not

- Run a command against a researcher’s `*.provenencia` folder except by opening it with this engine
- Add `golang-migrate` or a `schema_migrations` table
- Put identity / FFI / ingest in the SQL
