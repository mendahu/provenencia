---
name: add-name-value
description: >-
  Uses or extends Provenencia shared genealogical name_values (core/database/namevalues).
  Use when inserting/looking up NameValues, name_value_parts, form/parts rules,
  name_value_id FKs, open part-type vocabulary, structured-name-model, or
  Source/Interpretation/Conclusion name metadata.
---

# Add or use a NameValue

Shared cross-layer personal names live in **`core/database/namevalues`**, not in Source-only packages. Authoritative design: [`docs/structured-name-model.md`](../../../docs/structured-name-model.md).

New **tables/columns** need a catalog migration first (`.cursor/skills/add-catalog-migration/SKILL.md`). Do not invent a second name representation.

## Package

```
core/database/namevalues/
  namevalues.go
  namevalues_test.go
```

Call sites: `namevalues.Insert(c, v)` / `namevalues.Lookup(c, id)`. Minted id is UUIDv7 BLOB. Domain errors: `namevalues.ErrInvalid` (`namevalues.invalid`).

## Model (authoritative)

Every NameValue is **one** personal-name assertion.

| Field | Meaning |
| --- | --- |
| `form` | Required full-form normalized reading (e.g. `James K. Robins`). |
| `parts` | Optional ordered segments (`idx`, `value`, optional open `type`). |

Absence of parts is valid when further segmentation is unnecessary or culturally unclear. Citation transcription stays on the Citation; NameValue is the Interpretation (or later Conclusion) normalization.

### Part types

`type` is an **open vocabulary**, not a database enum. Starter keys (`prefix`, `given`, `initial`, `nick`, `surname_prefix`, `surname`, `suffix`, `undetermined`) are package consts for callers — Insert does not refuse unknown types. Culture-specific *display order* is `name_format_profiles` (out of this package; see structured-name-model §4).

### Identity

NameValues are value objects (UUID for persistence only). Prefer **new** rows when a name assertion changes rather than mutating an existing NameValue.

## Cross-layer use

Referencing tables use `name_value_id BLOB REFERENCES name_values(id)`. Observations with `properties.value_type = 'name'` store `value_name_id`.

## Extending helpers

1. Prefer extending validation + starter consts in `namevalues` over ad-hoc SQL elsewhere.
2. Keep `apperr` codes in `core/apperr`; package owns `ErrInvalid`.
3. Functions take `*database.Catalog` today (same as `datevalues`). Tx-scoped inserts wait until an audited write path needs them.
4. Table-driven tests in `namevalues_test.go`; run `CGO_ENABLED=1 go test ./core/database/namevalues/`.
5. Do **not** add shipped-SQL substring greps in `migrate_test.go`.

## Do not

- Require parts when `form` alone is enough
- Close part `type` to a CHECK enum or refuse unknown types
- Put NameValue CRUD on `Catalog` or under `core/database/*.go` root
- Seed or own `name_format_profiles` / project defaults in this package
- Grep `00000N.sql` contents in unit tests to “prove” schema
