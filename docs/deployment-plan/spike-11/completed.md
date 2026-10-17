# Spike 11 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S11-22 — Pure value types

**Landed 2026-10-17** in [#434](https://github.com/mendahu/provenencia/pull/434). Dates, names and property types can now be used without the database. The date value, the name value with its part-type registry, and the property value-type and cardinality constants moved into `core/datevalue`, `core/namevalue` and `core/propertymeta`, which import nothing from `core/database`, so they build and test with no build tags. Validation is exported (`datevalue.Validate`, `namevalue.Validate`) and each package owns its `ErrInvalid`. The database packages keep only their queries; nothing changes for researchers.

## S11-01 — One-page matching model

**Landed 2026-10-17** in [#432](https://github.com/mendahu/provenencia/pull/432). The matching algorithm now has one page to read. [`matching-model.md`](../../matching-model.md) draws the six layers in one diagram, defines the terms (relation, similarity, evidence, segment, shortlist, candidate, band, contested), walks one Promote row from value compare to band with real numbers, and lists every registry knob with its section, starting value or estimate, and what it changes. The model also fixes lifespan under `subject` (`subject.person.lifespan.*`) and keys the sex rows on `sex_at_birth`, which covers `gender` through the pair. `matching.md` and `promote-graph-alignment.md` carry banners saying which sections Spike 11 replaces and that they describe the running code until then. Docs only; no code changed.
