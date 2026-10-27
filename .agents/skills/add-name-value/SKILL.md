---
name: add-name-value
description: >-
  Uses or extends Provenencia shared genealogical name_values (pure type in core/namevalue, queries in core/database/namevalues).
  Use when inserting/looking up NameValues, name_value_parts, form/parts rules,
  name_value_id FKs, open part-type vocabulary, structured-name-model, or
  Source/Interpretation/Conclusion name metadata.
---

# Add or use a NameValue

Shared cross-layer personal names live in **`core/namevalue`** (the pure type, validation and part-type registry) and **`core/database/namevalues`** (the queries), not in Source-only packages. Authoritative design: [`docs/structured-name-model.md`](../../../docs/structured-name-model.md).

New **tables/columns** need a catalog migration first (`.agents/skills/add-catalog-migration/SKILL.md`). Do not invent a second name representation.

## Package

```
core/namevalue/              # pure: no core/database import
  namevalue.go               # Value, Part, PartType*, ErrInvalid, Validate
  registry.go                # PartTypeInfo, PartTypes, KnownPartType
  namevaluetest/             # Western and Given fixture helpers
  validate_test.go
  registry_test.go
core/database/namevalues/    # queries
  namevalues.go
  namevalues_test.go
```

Types and checks: `namevalue.Value`, `namevalue.Validate(v)`. Call sites for storage: `namevalues.Insert(c, v)` / `namevalues.Lookup(c, id)`. Minted id is UUIDv7 BLOB. Domain errors: `namevalue.ErrInvalid` (`namevalues.invalid`). Matching layers import `core/namevalue`, never `core/database/namevalues`.

`TestLayerPackages` in `core/matchregistry/guard_test.go` keeps this package pure; a new pure value package adds a row.

## Model (authoritative)

Every NameValue is **one** personal-name assertion.

| Field | Meaning |
| --- | --- |
| `form` | Required full-form normalized reading (e.g. `James K. Robins`). |
| `parts` | Optional ordered segments (`idx`, `value`, optional open `type`). |

Absence of parts is valid when further segmentation is unnecessary or culturally unclear. Citation transcription stays on the Citation; NameValue is the Interpretation (or later Conclusion) normalization.

### Part types

`type` is a **product-known key** from the compiled registry in `namevalue` (`PartTypes()` / `KnownPartType`). Keys match [`seeded-vocabulary.md`](../../../docs/seeded-vocabulary.md) §4.1 (`prefix`, `given`, `surname`, …). Empty type = untyped segment (allowed). `Validate` (and so Insert) **rejects** unknown non-empty types.

UI (S7-02b) picks from the registry with **localized labels** (`L10nKey`). Do **not** offer free-text type entry. User-minted part types / a catalog `name_part_types` table are a later extensibility PR — keep DDL as open TEXT until then.

Culture-specific *display order* is `name_format_profiles` (out of this package; see structured-name-model §4), which reference the same keys.

### Identity

NameValues are value objects (UUID for persistence only). Prefer **new** rows when a name assertion changes rather than mutating an existing NameValue.

## Cross-layer use

Referencing tables use `name_value_id BLOB REFERENCES name_values(id)`. Observations with `properties.value_type = 'name'` store `value_name_id`.

## Extending helpers

1. Prefer extending validation + the part-type **registry** in `namevalue` (pure) and queries in `namevalues` over ad-hoc SQL elsewhere.
2. Keep `apperr` codes in `core/apperr`; `core/namevalue` owns `ErrInvalid`.
3. Functions take `*database.Catalog` today (same as `datevalues`). Tx-scoped inserts wait until an audited write path needs them.
4. Pure checks: table-driven tests in `core/namevalue/validate_test.go` (no catalog); run `go test ./core/namevalue/...`. Query tests: `namevalues_test.go`; run `CGO_ENABLED=1 go test ./core/database/namevalues/`.
5. Do **not** add shipped-SQL substring greps in `migrate_test.go`.
6. Adding a product part type: append to `core/namevalue/registry.go` + consts + seeded-vocabulary §4.1 + L10n in S7-02b — not a migration.

## Do not

- Require parts when `form` alone is enough
- Close part `type` in DDL (CHECK enum) or invent a parallel name store
- Offer free-text part types in the UI — use the product registry picker
- Put NameValue CRUD on `Catalog` or under `core/database/*.go` root
- Seed or own `name_format_profiles` / project defaults in this package
- Grep `00000N.sql` contents in unit tests to “prove” schema
- Add `name_part_types` catalog rows until user extensibility is scheduled
