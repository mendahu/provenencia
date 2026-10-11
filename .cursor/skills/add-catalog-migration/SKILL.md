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
   - **Classify the FK before you write `ON DELETE`.** See [`docs/catalog-deletes.md`](../../../docs/catalog-deletes.md). **Resource → resource** is `NO ACTION`. **Facet → parent** (notes, metadata, layout, positions, name parts, vocab joins) is `ON DELETE CASCADE`; also mark it `Audited` (research rows: Go releases and audits them, the CASCADE is a backstop) or silent — see [`add-catalog-delete`](../add-catalog-delete/SKILL.md). **Owned outbound** (parent column → child, e.g. `observations.value_date_id`, `artifacts.file_id`) stays `NO ACTION`; register it on the parent’s release list (`always` or `ifUnused`) so official `Delete` removes the child. **Optional back-pointer** is `SET NULL` (`sources.primary_artifact_id`). In the same PR, register the table and every FK in `core/database/catalogmodel` — [`add-catalog-model`](../add-catalog-model/SKILL.md). Resource probes, owned-outbound releases, and title/location projectors stay in `deleteimpact` — [`add-catalog-delete`](../add-catalog-delete/SKILL.md). If researchers can delete the new row, add `Delete` there too — no private `sqlInUse`. Unregistered tables and FKs fail CI.
   - **New audited table?** Its writes need a non-`None` effect — see [`add-audit-scope`](../add-audit-scope/SKILL.md). A `None` entry is not recorded. Unmapped entity types fail CI.
   - **Table rebuilds drop indexes.** `DROP TABLE` + `RENAME` (SQLite cannot `ALTER` `ON DELETE`) also drops that table’s indexes. Recreate hot FK indexes in the same step or a follow-up (`000030` dropped `artifacts_source_id_idx` / `artifacts_file_id_idx`; `000031` restored them and added Impact-probe indexes). Honesty tests require a covering index (explicit, UNIQUE, or PK leftmost) on every FK child column.
4. **Do not write a test for the migration itself.** No upgrade-from-version-N test, no per-migration table-presence test, and no grepping shipped `.sql` contents in `migrate_test.go` (that file only tests `parseMigrations` rules with a fake FS). The existing suite covers a migration:
   - the schema hash self-check (Create/Open `verifySchema`)
   - `catalogmodel.TestPragmaHonesty` (tables and FKs against the live schema) and the `deleteimpact` honesty tests (probes, releases, projectors)
   - the domain package tests that read and write the new or changed table
   
   Test the **behavior** a migration enables (the store helpers for a new table) in its domain package, as for any other code. Data backfills follow the same rule — `000036.sql` backfills audit scopes untested; the effect paths it mirrors are tested in `core/database/audit/`.
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
- **Seed vocabulary rows** (`subject_types`, `properties`, `property_terms`, bindings, source types, grades, …) in SQL — use create-time `Install` / registries (`subjectvocab`, `sourcevocab`, …). Pre-production: refresh existing projects by re-running Install on the CLI; do not backfill seeds in migrations until the product owner says the app is in production.
