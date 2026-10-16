# Review — S10-32 Depiction value type and `likeness` on Person

**Verdict:** Fits with changes — Ready with changes (all applied)
**Rounds:** 1 (no blocker, so no round 2). **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None. The one product call came up during planning and was answered on 2026-10-16: `likeness` is new catalogs only, and a later one-off CLI updates the dogfood projects (working rule "`likeness` is new catalogs only").

## Writer's own calls

- D-1: a DDL-only migration (`000052.sql`) widens the `value_type` CHECK. The map said "Install path only"; that still holds for the seed row, and the map's In line now names the migration.
- D-4: a depiction needs no region, page, or Artifact file.
- D-5: negative likenesses are not special-cased.
- D-6: researchers cannot create a depiction Property.
- D-7: the autoreconciler loader skips depictions, so the Person page lists no empty Likeness field.
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
