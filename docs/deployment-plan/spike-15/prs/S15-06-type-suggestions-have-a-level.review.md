# Review — S15-06 Type suggestions have a level

**Verdict:** Fits with changes — Ready with changes (round 1)
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **M-1: does `Level` include `citation`?** Recommended: four values with a metadata subset. **Answered (user, 2026-11-11): four values, subset.** D-1, D-2, D-5, D-10 and the map's rule "One level type" follow.

## Writer's own calls

- D-2: no `DEFAULT 'source'` on `level` (the map sketch had one); the backfill names `'source'`.
- D-3: drop the redundant `source_type_metadata_fields_type_id_idx`.
- D-8: each RPC takes one level and returns that level's list; `TypeSuggestion` gains no level.
- D-9: the Fields column counts `source` rows only until S15-12.
- D-11: `CatalogQueryKey.typeSuggestions` and the type-suggestion mutations keep no level until S15-12.
- D-12: `FakeStore.suggestionsByType` stays as an alias over the `source` level until S15-12.
- Swift `MetadataLevel` has no `citation` case until a Swift caller needs it (S15-17).

## Round 1

### M-1 · blocker · `Level` has no slot for the `citation` level that S15-14 and S15-17 need
**Plan:** D-1 defined three values; "Next adds" had S15-14 map prefixes to `vocabkey.Level`.
**Codebase:** rule "One level type" (`deployment-plan.md`) serves the format prefixes and the built-ins table; map § 6.3 has `{citation.page}`, `{citation.image}`, `{citation.label}` and an Insert field group per four levels.
**Better shape:** four values; `MetadataLevels()` subset for suggestions, the `CHECK`, validation and the Swift enum.
**Answer:** To user → answered (2026-11-11): four values, subset. Fixed — D-1 (`LevelCitation`, `MetadataLevels`, `HoldsMetadata`), D-2, D-5, D-10, T-1, T-2, T-8, T-9, Out row for S15-14/S15-17, **Rules this plan bends**; the map's rule is amended in the wiring commit.

### M-2 · blocker · The suggested-field count lives in two Go places
**Plan:** D-9 added the level to `sqlList` and to `sourcevocab.CountSuggestions` separately.
**Codebase:** `sourcetypes.sqlList` subquery; `sourcevocab.CountSuggestions` (only caller `UpdateSourceType`).
**Better shape:** one `sqlSuggestedFields` and `sourcetypes.SuggestedFields` beside `UsedBy`; delete `CountSuggestions`.
**Answer:** Fixed — D-9, Codebase fit, Deletes, Go core, FFI, T-6, a stop condition. The FakeStore copy (`withSuggestedFieldCount`) keeps deriving from the `source` list (D-12); it is the fake, not a second engine path.

### M-3 · major · Seed rows hand-number `SortOrder`
**Plan:** D-6 added `Level` to the flat row.
**Codebase:** `sourcevocab/registry.go` `seedSuggestion`; S15-19 seeds 26 types per level in format order.
**Better shape:** one entry per (type, level) with ordered `FieldKeys`; position is the order.
**Answer:** Fixed — D-6, Codebase fit, Deletes, T-4, STEP-7's skill line.

### M-4 · minor · `suggestionsByType` kept as an alias for the `source` level
**Answer:** Declined — S15-12 draws the three lists and re-keys the fake with them; renaming ~20 fixture sites here adds churn to a PR that shows one list. Out row added with home S15-12.

### M-5 · minor · Current state misses raw writers of the join
**Answer:** Fixed — §4 Tests row lists the two raw inserts and raw count in `sourcetypes_test.go` and the raw count in `metadatafields_test.go`; STEP-3 names them.

### M-6 · nit · `Level.Valid()` is a new shape in `vocabkey`
**Answer:** Fixed — D-1 says methods on typed keys are the form for new typed keys.

### X-1 · blocker · An unlisted handler test sends no `level`
**Codebase:** `api/ffi/handlers/sources_test.go` metadata fixture; `source_defs_test.go` `UpdateSourceType` case and `TestListTypeSuggestions`.
**Answer:** Fixed — §4 Tests row and STEP-4 list `sources_test.go`; STEP-4 gives the `rg` that finds every request literal.

### X-2 · major · Raw insert in `sourcetypes_test.go`; T-6's second row
**Answer:** Fixed — STEP-3 and T-6: the same field inserted by raw SQL at `source` and at `graph`.

### X-3 · major · The version bump contradicts the migration skill
**Answer:** Declined — `add-catalog-migration` says not to bump for a format change *alone*; this PR also adds a request field on three RPCs, and `.cursor/rules/versioning.mdc` with `docs/versioning.md` ("new RPCs old clients can ignore, additive schema" → MINOR) govern. Spike precedent: S15-02 (migration + RPCs) and S15-32 (a request field) each bumped minor.

### X-4 · minor · `completed.md` without a path
**Answer:** Fixed — STEP-7 and the DoD give spike-15 paths.

### X-5 · minor · Parity row placement
**Answer:** No change — the reviewer found the row workable as written.

### X-6 · minor · FakeStore locking
**Answer:** Fixed — D-12: the computed property takes no lock; callers already hold `stateLock` through `withState`, which is not reentrant.

### X-7 · minor · Validation order in the handlers
**Answer:** No change — the reviewer confirmed AC-2.5 holds.

Also fixed from X-5's notes: §4's proto row now says fields 1, 3, 4 with 2 reserved.
