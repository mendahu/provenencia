---
name: add-date-value
description: >-
  Uses or extends Provenencia shared genealogical date_values (pure type in core/datevalue, queries in core/database/datevalues).
  Use when inserting/looking up DateValues, changing kind/qualifier/precision/timezone
  rules, date_value_id FKs, ABT/BEF/AFT, point/range dates, structured-date-model,
  or Interpretation/Conclusion date properties (not Source catalog metadata).
---

# Add or use a DateValue

Shared cross-layer civil dates live in **`core/datevalue`** (the pure type, validation and compact display) and **`core/database/datevalues`** (the queries), not in Source-only packages. Authoritative design: [`docs/structured-date-model.md`](../../../docs/structured-date-model.md).

New **columns** need a catalog migration first (`.agents/skills/add-catalog-migration/SKILL.md`). Do not invent a second date representation.

## Package

```
core/datevalue/              # pure: no core/database import
  datevalue.go               # Value, kinds, qualifiers, ErrInvalid, Validate
  display.go                 # CompactDisplay
  calendar.go                # DaysFromCivil, DaysInMonth, BoundDays, YearLine, AddMonths, AddDays, AddYears, CompletedYears
  validate_test.go
  display_test.go
core/database/datevalues/    # queries
  datevalues.go
  datevalues_test.go
```

Types and checks: `datevalue.Value`, `datevalue.Validate(v)`, `datevalue.CompactDisplay(v)`. Call sites for storage: `datevalues.Insert(c, v)` / `datevalues.Lookup(c, id)`. Minted id is UUIDv7 BLOB. Domain errors: `datevalue.ErrInvalid` (`datevalues.invalid`). Matching layers import `core/datevalue`, never `core/database/datevalues`.

Calendar arithmetic (day numbers, month lengths, the days a bound covers) lives in `core/datevalue/calendar.go`; the date ends and day windows that compare dates live in `core/valuecompare/datewindow.go` (`DateEnds`, `AboutEnds`, `Window`, `ShiftYears` for a year move on an end; `datevalue.AddYears` moves 29 February to 1 March in a common year, `CompletedYears` is its inverse, `Earliest` / `Latest` for the one earliest / latest pick of date ends). Do not write a second calendar. Where a date sorts is `core/valuecompare/dateorder.go` (`DateKeyOf`, `DateKey.Compare`, `DateKey.Int64`, beside `datewindow.go`): timelines, birth order, a Source's `date_sort_key` and the reconciler's stored date `sort_key` all read it. Do not write a second order.

`TestLayerPackages` in `core/matchregistry/guard_test.go` keeps this package pure; a new pure value package adds a row.

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

## Owning a DateValue

An Observation (`value_date_id`) and a Source (`sources.date_value_id`) each own their date row, unique to the owner (a unique index on the column). An owner stores it with `datevalues.PutTx(tx, prevID, v)`: with the id it holds, the row is edited in place and the update change names only the fields that differ (nil, and no write, when nothing differs; strings are trimmed before the comparison); with none, a new row and its create change. `datevalues.RowMap` / `CreateChange` / `UpdateChange` build the audit rows. An owner that drops or replaces its date releases the old row with `deleteimpact.ReleaseReplaced`, which returns the audited delete; an owner that is deleted takes `SnapshotOwned` and `ReleaseSnapshot`. A second kind of owner passes its path to `effects.ownedValue` ([`add-audit-scope`](../add-audit-scope/SKILL.md)).

## Cross-layer use

Referencing tables use `date_value_id BLOB REFERENCES date_values(id)` (a Source's own date included; its catalog `source_metadata` dates are still text). Observation / Claim callers may keep as-written wording on the domain row and attach structured `date_value_id`. **Do not** add `date_value_id` to `source_metadata` — catalog metadata dates are `value_text` only ([`docs/source-layer-data-model.md`](../../../docs/source-layer-data-model.md) §5).

DateValues are value objects (UUID for persistence only). Concluded/refined zones or dates later should usually be **new** DateValue rows (or later-layer assertions), not silent mutation of Source evidence.

## Display (macOS)

User-visible DateValue strings (list rows, composer summaries, previews) go through **`DateValueDisplay`** in [`macos/App/Platform/DateValueDisplay.swift`](../../../macos/App/Platform/DateValueDisplay.swift). Do not ad-hoc interpolate `y-m-d` or hard-code English ABT/BEF/AFT/between wrappers — call `DateValueDisplay.string(for:locale:)` (accepts `DateValueDraft` or `CatalogDateValueInput`). Qualifier/range templates live under `L10n.Dates`. Invalid drafts return `""`; call sites that need a placeholder keep their own L10n (e.g. composer “No date set”).

Evidence graph cited rows use **`ObservationValueDisplay`**, which prefers a structured NameValue via `NameValueDisplay` (profile order, else the form), then structured `CatalogObservation.date` → `DateValueDisplay`, then denormalized `valueText` / integer. List RPCs still fill `value_text` with the name form so other clients stay correct without type-specific UI.

## Extending helpers

1. Prefer extending validation + constants in `datevalue` (pure) and queries in `datevalues` over ad-hoc SQL elsewhere.
2. Keep `apperr` codes in `core/apperr`; `core/datevalue` owns `ErrInvalid`.
3. Functions take `*database.Catalog` today (same as `users`). Tx-scoped inserts wait until an audited write path needs them.
4. Pure checks: table-driven tests in `core/datevalue/validate_test.go` (no catalog); run `go test ./core/datevalue/`. Query tests: `datevalues_test.go`; run `CGO_ENABLED=1 go test ./core/database/datevalues/`.
5. Do **not** add shipped-SQL substring greps in `migrate_test.go`.

## Do not

- Store only UTC and discard civil components
- Require timezone whenever hour is set
- Put DateValue CRUD on `Catalog` or under `core/database/*.go` root
- Treat `range` as FROM–TO event duration
- Reintroduce `exact`/`year` kinds — use `point` + components
- Grep `00000N.sql` contents in unit tests to “prove” schema
- Attach DateValue to `source_metadata` (catalog dates are text)
