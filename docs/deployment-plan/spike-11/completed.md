# Spike 11 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S11-02 — Value compare core and match registry

**Landed 2026-10-17** in [#441](https://github.com/mendahu/provenencia/pull/441). Two values of one type can now be related without the database. `core/valuecompare` takes text, term and integer values and returns a relation, a similarity and a structured detail (what paired with what, what is on one side only, why), and gives each value a normalized `Key` that agrees with `Compare`. Text compares words in order, so `Upper Canada, York` and `York, Upper Canada` are `Different`. `core/matchregistry` opens with `Default()` and the value section's text, term and integer knobs; `core/match` reads the same numbers from it. The word helpers moved out of `core/autoreconcile` into value compare. `TestLayerPackages` keeps the layer packages free of catalog imports and stray numbers, `TestKnobInventoryMatchesRegistry` keeps the knob inventory in step with the registry, and the `add-match-rule` skill writes down the steps each layer PR repeats. Nothing changes for researchers.

## S11-23 — One home for primary kinds

**Landed 2026-10-17** in [#438](https://github.com/mendahu/provenencia/pull/438). The primary Subject kinds (`person`, `event`, `place`), the eight seeded Property keys matching reads by name, and the vocabulary origins now live in `core/vocabkey`, which imports only `strings`. The ten per-package `OriginProvenencia` declarations, six `originOK` copies, and the hand-written kind lists in Promote, progress, alignment, search and the Swift fake are gone; `search.KindSpec` gains `Handle` to mark handle kinds. `TestVocabKeysMatchSeed` ties the constants to the seed and `TestSwiftVocabularyMatchesGo/primary_kinds` ties them to `EvidencePrimaryKind`. Nothing changes for researchers.

## S11-22 — Pure value types

**Landed 2026-10-17** in [#434](https://github.com/mendahu/provenencia/pull/434). Dates, names and property types can now be used without the database. The date value, the name value with its part-type registry, and the property value-type and cardinality constants moved into `core/datevalue`, `core/namevalue` and `core/propertymeta`, which import nothing from `core/database`, so they build and test with no build tags. Validation is exported (`datevalue.Validate`, `namevalue.Validate`) and each package owns its `ErrInvalid`. The database packages keep only their queries; nothing changes for researchers.

## S11-01 — One-page matching model

**Landed 2026-10-17** in [#432](https://github.com/mendahu/provenencia/pull/432). The matching algorithm now has one page to read. [`matching-model.md`](../../matching-model.md) draws the six layers in one diagram, defines the terms (relation, similarity, evidence, segment, shortlist, candidate, band, contested), walks one Promote row from value compare to band with real numbers, and lists every registry knob with its section, starting value or estimate, and what it changes. The model also fixes lifespan under `subject` (`subject.person.lifespan.*`) and keys the sex rows on `sex_at_birth`, which covers `gender` through the pair. `matching.md` and `promote-graph-alignment.md` carry banners saying which sections Spike 11 replaces and that they describe the running code until then. Docs only; no code changed.
