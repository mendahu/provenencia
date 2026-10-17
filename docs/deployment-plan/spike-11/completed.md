# Spike 11 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

**Landed 2026-10-17** in [#434](https://github.com/mendahu/provenencia/pull/434). Dates, names and property types can now be used without the database. The date value, the name value with its part-type registry, and the property value-type and cardinality constants moved into `core/datevalue`, `core/namevalue` and `core/propertymeta`, which import nothing from `core/database`, so they build and test with no build tags. Validation is exported (`datevalue.Validate`, `namevalue.Validate`) and each package owns its `ErrInvalid`. The database packages keep only their queries; nothing changes for researchers.
