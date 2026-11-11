# Review — S15-07 Metadata operations take an owner

**Verdict:** (pending round 2) — round 1: Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None.

## Writer's own calls

- D-4: `Input` and `Row` lose `SourceID`; `TestCreateAll` "input for another source is refused" is deleted because its state cannot be expressed.
- D-7: `Apply` writes layout rows verbatim, never deletes a layout row the draft leaves out, and never appends one for a new value; a board on a draft sends its whole order (M-8, declined below).
- D-11: `ListWorkspace` returns `ErrInvalid` for a missing owner; the wire keeps `sources.not_found` through `GetSourceWorkspace`'s `sources.Get`.
- D-10: no chunking of `IN (…)` lists; S15-16's largest call binds 300 ids.
- D-5: `ListBySource`, `ListBySourceTx` and `ListLayoutTx` go; `ReadDraftTx` replaces them.

## Round 1

### M-1 · blocker · Apply's layout update drops the owner column effects and audit scope read
**Plan:** D-7 (4) "fields carry only the columns that differ"; AC-4.5, AC-4.6, T-2.
**Codebase:** `effects/registry.go` `"source_metadata_layout"` `SourceScope: field("source_id")`; `effects/path.go` `readChangeColumn` falls back to `pkOf`, which the layout table (no `PK` in `catalogmodel`) cannot serve.
**Better shape:** one `layoutChange` builder that always writes the owner column and `field_id`.
**Answer:** Fixed — D-8 `layoutChange`; D-7 uses it; AC-4.5, AC-4.6 name the exact keys; T-2 checks them. (Same as X-1.)

### M-2 · blocker · Layout change shapes stay in three hand-built places
**Plan:** D-8 kept `Reorder` and `DismissSuggestion` building changes inline.
**Codebase:** `sourcemetadata.go` `CreateAll`, `DismissSuggestion`, `Reorder`; rule "Metadata saves with its owner's write".
**Answer:** Fixed — D-8: all four layout writers build changes through `layoutChange`; `Reorder` and `DismissSuggestion` keep their upsert SQL, which is the rule they implement, and their change fields stay today's.

### M-3 · blocker · ListWorkspace is per owner; S15-08 needs 60 Artifact boards in one read
**Codebase:** map scope entry S15-08 "returns its Artifacts' boards in one query (60 Artifacts at Scale)"; `listLayout` single-owner.
**Answer:** Fixed — D-11 `listWorkspaces(tx, owners)` with `ListWorkspace` as its one-owner call; D-10 `listLayoutByOwners`; T-9; Out row "an exported many-owner workspace read" → S15-08; Scale and Next adds updated.

### M-4 · blocker · The Out row drops the skill's one-read-transaction rule
**Codebase:** `.agents/skills/add-catalog-query/SKILL.md` "A read made of more than one query opens one read transaction".
**Answer:** Fixed — the Out row is removed; D-11 opens one read transaction; D-12 adds `sourcevocab.ListSuggestionsTx`; T-11; a risk row checks no handler calls it inside `writes.Run`.

### M-5 · blocker · A third field-row scan outside `metadatafields`
**Codebase:** `metadatafields` `sqlGetByID` / `getByIDTx`; `sourcemetadata` `sqlFieldGet` / `getField`.
**Answer:** Fixed — D-13 `metadatafields.ListByIDsTx`; `getField` and `sqlFieldGet` deleted; T-10.

### M-6 · major · `List(c, owner)` would have only a test caller
**Answer:** Fixed — D-5 drops it; `ReadDraftTx` and a test helper `values` (D-16) cover the tests.

### M-7 · major · `notFound sources.ErrNotFound` keeps `sourcemetadata` importing `sources`, which imports `artifacts` and `evidencegraphs`
**Answer:** Fixed — D-2 drops `notFound`; D-11: a missing owner is `ErrInvalid`; AC-1.3 is now checked at `GetSourceWorkspace` (its existing test); Next adds says S15-08 calls `Apply` from handlers; a Stop condition names an import cycle.

### M-8 · major · Apply never appends a layout row for a new value, while `Set` does
**Better shape:** say where a new field goes on a draft board, or append in `Apply` when the owner already has layout.
**Answer:** Declined — `CreateAll` must write layout verbatim (`TestCreateAll` "values with no layout rows write no layout"), and it is `Apply`'s empty-owner case; appending only when layout exists gives one draft two meanings. A board on a draft already holds the whole order and sends it (D-7; Out row → S15-26). No order a researcher sees changes: a field added at the end of a board is at the end of its draft layout.

### M-9 · minor · `CreateAll` checks emptiness with a separate query
**Answer:** Fixed — D-9 reads `stored` once and calls `apply`; `sqlHasAny` deleted.

### M-10 · minor · T-5 pins `artifact` and `graph` as refused
**Answer:** Fixed — T-5 iterates the levels with no `stores` entry; T-8 `TestStoresAreRegistered` checks each entry against `catalogmodel`.

### X-1 · blocker · Layout update shape contradicts AC-4.5, AC-4.6, T-2
**Answer:** Fixed — with M-1.

### X-2 · blocker · Stop condition 1 fires on a clean checkout
**Answer:** Fixed — `rg -ln "sourcemetadata\.[A-Z]"` with `sources_test.go` and the `sourcecopy` test listed.

### X-3 · major · STEP-2's check needs STEP-3's code; the tree does not build between steps
**Answer:** Fixed — signatures and callers change in one step (STEP-2); T-5 moved to STEP-3.

### X-4 · major · `store` fields and `newStore` left to guesswork
**Answer:** Fixed — D-2 names every SQL field, the `IN` prefixes and their suffixes, and `newStore`'s signature.

### X-5 · major · Apply edges unsettled (empty draft, missing owner on reads, nil tx)
**Answer:** Fixed — AC-4.10, AC-4.11; D-6 and D-10 nil `tx`; T-5.

### X-6 · minor · Handler call count; domain doc `date_value_id`; `rowFacet` wording
**Answer:** Fixed — §4 Handlers (eight calls), Schema row and STEP-4 (§ 5.3 gains the column and a corrected sentence), Precedent.

### X-7 · minor · Greps that need eyeballing; Xcode on Linux
**Answer:** Fixed — STEP-2's `rg` prints nothing on success; STEP-4 says what to report without macOS.

### X-8 · minor · Smaller choices
**Answer:** Fixed — T-6 field types; T-5 counts. The rest is covered by "Do only In".
