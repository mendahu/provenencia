# Review — S10-32 Depiction value type and `likeness` on Person

**Verdict:** Fits with changes — Ready with changes (all applied)
**Rounds:** 2. Round 1 reviewed the first shape (the loader skipped depictions). The user then asked for a shape that keeps every value type reconciled; the plan was re-planned with a `depiction` module, and round 2 reviewed it in full. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None open. Answered 2026-10-16:
- `likeness` is new catalogs only; a later one-off CLI updates the dogfood projects (working rule "`likeness` is new catalogs only").
- Depictions are reconciled, not skipped (the user questioned the exclusion).
- The gallery shows `weak` crops (working rule "A gallery is the reconciled likeness values").

## Writer's own calls

- D-1: a DDL-only migration (`000052.sql`) widens the `value_type` CHECK. The map said "Install path only"; that still holds for the seed row, and the map's In line now names the migration.
- D-4: a depiction needs no region, page, or Artifact file.
- D-5: negative likenesses are not special-cased.
- D-6: researchers cannot create a depiction Property.
- D-7: `depiction` has a key module keyed by the Citation id, not by File plus locator.
- D-12: the cache column has no FK, and `CacheVersion` is bumped to 15 although no catalog holds a depiction yet.
- D-14: the Person field list skips depictions until S10-35 draws the gallery.
- D-8: Promote alignment is not changed; T-6 pins that no likeness line is shown until S10-36.
- D-9: the `likeness` binding is locked, at sort order 4.
- D-10: no pill colour for `depiction`.

## Round 1

### M-1 · major · Domain doc made false: "Every value type needs an auto-reconciler"
`docs/conclusion-reconciliation.md` §1, §5 and §7, and the value-population sentence in `docs/conclusion-layer-data-model.md`, become false under D-7.
**Answer:** Fixed. §4 has a "Reconcile docs" row. STEP-7 gives the exact edits for both docs, and the DoD names them.

### M-2 · major · The new Swift label goes into a string switch whose `default` caused the mislabel
`SubjectPropertyValueType.label` switches on raw strings, and `.cursor/rules/macos-client.mdc` says views use the typed `PropertyValueType`. The pill shows the raw key, so only VoiceOver used the label.
**Answer:** Fixed. D-10 now switches exhaustively over `PropertyValueType`, and `researcherCreatable` becomes `[PropertyValueType]` with call sites named. §4 has a Pill row. AC-3.5, §6 and the manual check now say the pill shows `depiction` and VoiceOver reads *Depiction*. `valueTypeForeground` and `CitationObservationValue` are Out lines (default colour, and S10-34).

### M-3 · minor · The registry-only rule lives in two hand-kept lists
**Answer:** Fixed for Go. D-6 adds `registryOnly`, and `Create` calls it. The Swift↔Go parity subtest for `researcherCreatable` is logged here and in Out, not planned. Swift keeps the same five types, now typed.

### M-4 · nit · T-6 pins a non-behavior that S10-36 reverses
**Answer:** Fixed. T-6 carries a comment that S10-36 replaces the assertion.

### X-1 · major · T-5 reads `f.props["likeness"]`, which the fixture never loads
**Answer:** Fixed. STEP-5 adds `"likeness"` to the `newFixture` key loop. T-5 names the Subjects, `nameIn` with its spec, and the inputs it uses.

### X-2 · minor · T-2 does not say how to build its locator and counts
**Answer:** Fixed. T-2 uses `mustSeed`'s artifact and person, `citations.CreateWithObservations` through `writes.Run`, and raw `SELECT`s on `c.DB()`.

### X-3 · major · STEP-6's check cannot run off macOS, and its Swift cases are underspecified
**Answer:** Fixed. STEP-6 says `xcodebuild` is macOS only and what to report when it does not run. T-9 asserts through `L10n.string`. The xcstrings entry copies the `term` block and goes after `properties.valueType.date`. The `PropertiesModelTests` `term` line is updated.

### X-4 · minor · STEP-7 docs leave the wording to the implementer
**Answer:** Fixed. STEP-7 gives each sentence, the §3.2 and §3.3 rows, the map lines, and the `completed.md` shape.

### X-5 · minor · "Nothing is written" has no assertion form
**Answer:** Fixed. T-3 compares the `citations` and `observations` counts before and after.

### X-6 · nit · The Swift doc comment goes stale
**Answer:** Fixed. D-10 updates the `SubjectPropertyValueType` doc comment.

## Round 2 (reconciled shape)

### M-5 · major · The loader's `hasValue` is a copy of `autoreconcile.carries`
**Answer:** Fixed. D-7 exports `Carries`. D-12 has `load` call it and deletes `hasValue`. Codebase fit and **Deletes** list both.

### M-6 · major · The gallery's "kept or weak" rule lives only in prose
**Answer:** Fixed, by changing the map. The S10-35 row now binds the rule to one `autoreconcile.Reason` method beside `Displayed`, which S10-39 reuses, and the Out row points there. This PR adds no method with no caller.

### M-7 · minor · `knownValueType` is a third hand-kept list
**Answer:** Fixed. `knownValueType(vt)` is `moduleFor(vt) != nil`. Declined: an exported `properties.ValueTypes` slice. It reaches beyond this PR's layers, and Out logs it.

### M-8 · minor · Stale doc and code comments
**Answer:** Fixed. STEP-9 adds the "Read-side tolerance" clause. D-14 updates the `attachValues` comment, and D-7 updates the package doc.

### M-9 · nit · The loader branches on value type
**Answer:** Fixed. `sqlLoadCandidates` selects `CASE WHEN p.value_type = 'depiction' THEN c.id END` (D-12).

### M-10 · nit · Bookkeeping
**Answer:** Fixed. D-12 to D-14 follow D-11, and this log restates the writer's calls.

### X-1 · major · T-5's `Verify` is vacuous
**Answer:** Fixed. T-5 loads `Graph().Node(h)` before the last write and again before `Verify`, and asserts the node's likeness `CitationID`s.

### X-2 · major · T-5 lifecycle and setups
**Answer:** Fixed. Each subtest gets its own fixture. B is filed only by `identityclaims.Create` in AC-2.4. AC-2.3 gives the `promote.Input` for `c`. `f.credibility` is wrapped in `must`, and rows are looked up by `ValueCitationID`.

### X-3 · minor · STEP-6 bundles five changes
**Answer:** Fixed. STEP-6 is the cache and graph mirror; STEP-7 is the Person field skip and the Promote guard. There are now 9 steps.

### X-4 · minor · T-11 does not exercise version 14
**Answer:** Fixed. T-11 adds `14` to the stale-version list.

### X-5 · minor · T-1 overlaps an existing row
**Answer:** Fixed. T-1 extends "create refuses term value type" into a loop and adds an Upsert row.

### X-6 · minor · Small choices in T-12, T-6, and T-2 to T-4
**Answer:** Fixed. T-12 is its own function, using `properties.Lookup`. T-6 spells out `srcB, artB`. T-2 to T-4 are table rows.

### X-7 · minor · Docs and log drift
**Answer:** Fixed. STEP-9 covers line 891. This log and the header are current. The draft PR is opened before STEP-9.

### X-8 · minor · The DoD box off macOS
**Answer:** Fixed. The DoD line says how to report it.

### X-9 · minor · No old catalog for manual check 4
**Answer:** Fixed. The check builds `main` at the Planned commit first, or reports itself as not run.

### X-10 · nit · Others
**Answer:** Fixed: `canSubmitCreate`; STEP-4 to STEP-6 land together; Out names the proto comment and the inspector note. The `.contains(.term)` edit stays, because the old line no longer compiles.
