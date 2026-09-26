---
name: add-date-value
description: >-
  Uses or extends Provenencia shared genealogical date_values (core/database/datevalues).
  Use when inserting/looking up DateValues, changing kind/qualifier/precision/timezone
  rules, date_value_id FKs, ABT/BEF/AFT, point/range dates, structured-date-model,
  or Interpretation/Conclusion date properties (not Source catalog metadata).
---

# Add or use a DateValue

Shared cross-layer civil dates live in **`core/database/datevalues`**, not in Source-only packages. Authoritative design: [`docs/structured-date-model.md`](../../../docs/structured-date-model.md).

New **columns** need a catalog migration first (`.cursor/skills/add-catalog-migration/SKILL.md`). Do not invent a second date representation.

## Package

```
core/database/datevalues/
  datevalues.go
  datevalues_test.go
```

Call sites: `datevalues.Insert(c, v)` / `datevalues.Lookup(c, id)`. Minted id is UUIDv7 BLOB. Domain errors: `datevalues.ErrInvalid` (`datevalues.invalid`).

## Model (authoritative)

Every DateValue is **one** genealogical date.

| `kind` | Meaning |
| --- | --- |
| `point` | One date; precision = which components are set (year / month / day / time…). |
| `range` | One date in a bounded window (`BET`); start = earliest, end = latest. Not an event duration. |

| `qualifier` (point only) | Meaning |
| --- | --- |
| `""` | As stated |
| `ABT` | About |
| `BEF` | Before / no later than the point |
| `AFT` | After / no earlier than the point |

`range` → qualifier empty (**between** is the kind, not a qualifier).

### Precision

Any subset of year / month / day / hour / minute / second / millisecond may be set; **gaps are allowed**. At least one of year / month / day / hour / minute / second (or a `phrase`) is required. Missing fields = **unknown**, not midnight. `point` forbids `end_*`. Domain `value_text` holds attachment fidelity; DateValue `phrase` is optional wording on the value itself.

### Timezone

`StartTZ` / `EndTZ` are **free text**. Empty = unspecified. Do **not** parse to UTC in this package.

Phrase-only `point` values (no civil components) are valid when `Phrase` is set. `range` needs at least one civil component on each side (year is not required).

## Cross-layer use

Referencing tables use `date_value_id BLOB REFERENCES date_values(id)`. Observation / Claim callers may keep as-written wording on the domain row and attach structured `date_value_id`. **Do not** add `date_value_id` to `source_metadata` — catalog dates are `value_text` only ([`docs/source-layer-data-model.md`](../../../docs/source-layer-data-model.md) §5).

DateValues are value objects (UUID for persistence only). Concluded/refined zones or dates later should usually be **new** DateValue rows (or later-layer assertions), not silent mutation of Source evidence.

## Display (macOS)

User-visible DateValue strings (list rows, composer summaries, previews) go through **`DateValueDisplay`** in [`macos/App/Features/Dates/DateValueDisplay.swift`](../../../macos/App/Features/Dates/DateValueDisplay.swift). Do not ad-hoc interpolate `y-m-d` or hard-code English ABT/BEF/AFT/between wrappers — call `DateValueDisplay.string(for:locale:)` (accepts `DateValueDraft` or `CatalogDateValueInput`). Qualifier/range templates live under `L10n.Dates`. Invalid drafts return `""`; call sites that need a placeholder keep their own L10n (e.g. composer “No date set”).

Evidence graph cited rows use **`ObservationValueDisplay`**, which prefers structured `CatalogObservation.date` → `DateValueDisplay`, then denormalized `valueText` / `nameForm` / integer. List RPCs fill `value_text` for term/name/date/subject so other clients stay correct without type-specific UI.

## Extending helpers

1. Prefer extending validation + constants in `datevalues` over ad-hoc SQL elsewhere.
2. Keep `apperr` codes in `core/apperr`; package owns `ErrInvalid`.
3. Functions take `*database.Catalog` today (same as `users`). Tx-scoped inserts wait until an audited write path needs them.
4. Table-driven tests in `datevalues_test.go`; run `CGO_ENABLED=1 go test ./core/database/datevalues/`.
5. Do **not** add shipped-SQL substring greps in `migrate_test.go`.

## Do not

- Store only UTC and discard civil components
- Require timezone whenever hour is set
- Put DateValue CRUD on `Catalog` or under `core/database/*.go` root
- Treat `range` as FROM–TO event duration
- Reintroduce `exact`/`year` kinds — use `point` + components
- Grep `00000N.sql` contents in unit tests to “prove” schema
- Attach DateValue to `source_metadata` (catalog dates are text)
