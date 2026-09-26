---
name: add-catalog-migration
description: >-
  Adds a Provenencia SQLite catalog format step (user_version, migrations/NNNNNN.sql).
  Use when adding or changing core/database migrations, CREATE TABLE in provenencia.sqlite,
  user_version, schema format, schema hash, schemahash.go, catalog.schema_mismatch,
  or when the user asks to add a migration.
---

# Add a catalog migration

This is **not** a web `schema_migrations` / timestamp workflow. There is **no** migrate CLI. `Create` / `Open` apply SQL. Old apps must refuse a **newer** `user_version`.

Product SemVer (`VERSION`) is a separate bump. Do not bump it for a format change unless you are cutting a release.

## Steps

1. List `core/database/migrations/*.sql`. Next file is the next integer, **six digits**: `000002.sql` after `000001.sql`.
2. Create **only** that new file. Do **not** edit a shipped `NNNNNN.sql` (rewrite history of existing folders). Unreleased steps on this branch may still be amended.
3. Put additive DDL in the new file (`CREATE TABLE` / `STRICT`). No Source tables unless the spike/PR explicitly includes them.
   - **Classify the FK before you write `ON DELETE`.** See [`docs/catalog-deletes.md`](../../../docs/catalog-deletes.md). **Resource → resource** is `NO ACTION`. **Facet → parent** (notes, metadata, layout, positions, name parts, vocab joins) is `ON DELETE CASCADE`. **Owned outbound** (parent column → child, e.g. `observations.value_date_id`, `artifacts.file_id`) stays `NO ACTION`; register it on the parent’s release list (`always` or `ifUnused`) so official `Delete` removes the child. **Optional back-pointer** is `SET NULL` (`sources.primary_artifact_id`). After **S8-12**, register the FK in `core/database/deleteimpact` in the same PR (count + list probe if inbound **resource**, plus title/location projectors). If researchers can delete the new row, add `Delete` with [`add-catalog-delete`](../add-catalog-delete/SKILL.md) — no private `sqlInUse`. Unregistered FKs fail CI.
4. Add a table-driven test in the **domain package** that owns the new table: `database.Create` a temp catalog, assert `user_version` / table presence / helpers. Do **not** grep shipped `.sql` contents in `migrate_test.go` (that file only tests `parseMigrations` rules with fake FS).
5. Run `CGO_ENABLED=1 go test -tags fts5 ./core/database/...`. Init panics (and tests fail) on gaps, `1.sql` names, empty SQL, or missing files. Create/Open also run `verifySchema` (see Schema hash below).
6. Do not add embed vars, `formatVersion` constants, or timestamp filenames. The glob in `migrate.go` is the registry.

## Schema hash

After migrate, `Create` / `Open` hash `sqlite_schema` and compare it to an **init-derived** expected digest (`core/database/schemahash.go`). Mismatch → `catalog.schema_mismatch`.

1. After adding the `.sql` file, run `CGO_ENABLED=1 go test -tags fts5 ./core/database/...`. That suite must include Create/Open paths that call `verifySchema`.
2. **Do not** add, edit, or regenerate a golden `expectedSchemaHash` constant, checked-in digest file, or `go:generate` hash artifact. Digest is computed in migrate `init` from embedded migrations via `initExpectedSchemaHash`.
3. **Do not** weaken or skip `verifySchema` to land a migration. If Create fails `catalog.schema_mismatch` after a correct migration, fix migrate/hash encoding — the new DDL should define the new expected schema.
4. Extra researcher indexes / hand-edited `sqlite_schema` are unsupported; Open refuses them on purpose.

Agents “update the hash” by shipping the migration and proving tests pass — never by inventing a parallel digest file or constant.

## Do not

- Run a command against a researcher’s `*.provenencia` folder except by opening it with this engine
- Add `golang-migrate` or a `schema_migrations` table
- Put identity / FFI / ingest in the SQL
- Commit a hand-maintained schema hash / bypass open-time schema verification
