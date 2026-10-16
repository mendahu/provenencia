# Review — S10-32 Depiction value and `likeness` on Person

**Verdict:** Fits with changes — Ready with changes (all applied)
**Rounds on this shape:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

This log reviews the shape planned after the Slice 8 re-cut: an owned `depiction_values` row (region plus a staged image File). The earlier shape (the value is the Citation, with a reconciler module) is kept below under "Earlier shape (superseded)".

## To user

None. The shape follows the user's answers of 2026-10-16: the value is its own row holding the region and an extracted image; the image points at `files`, not `file_derivatives`; a client-sent image goes through every file protection; `likeness` is new catalogs only.

## Writer's own calls

- D-3: researcher Create uses an allow-list, `researcherCreatable`, the same list as Swift's.
- D-5: a depiction shows one page; a Citation with two page selectors cannot hold one, and the region need not lie inside the Citation's own region.
- D-7: callers stage with `ingest.StageDepictions` before `writes.Run`; the staged image rides in `Input.Depiction` behind `observations.StagedFile`, so no write signature changes and no `core/database` package imports `ingest`.
- D-9: date and name updates move onto `releaseReplacedValue` with depictions (one release path on update).
- D-11: the auto-reconciler skips depiction Observations until S10-52; `CacheVersion` is not bumped.
- Out: a Citation locator edit after a likeness is saved is not kept in step; the stored image stays a true copy.
- Out: the Swift value type and label move to S10-54, so this PR has no Swift change.

## Round 1

### M-1 · blocker · Value-row release on update becomes three mechanisms in `updateOne`
**Answer:** Fixed. D-9 adds `releaseReplacedValue`, used for date, name, and depiction values; `releaseDateValue`, `releaseNameValue`, and their four SQL constants are deleted. S10-51's Out row now points here. A Stop condition guards the date and name update tests.

### M-2 · major · AC-2.5 / T-9 test a Subject delete that refuses
**Answer:** Fixed with X-1. AC-2.5 and T-9 (`TestDeleteSubjectWithLikeness`) expect `ErrInUse`, then success after `observations.Delete`. §4 and D-10 say `eraseObservation` never sees a likeness today.

### M-3 · major · The new error code breaks the naming rule
**Answer:** Fixed. `depictionvalues.invalid` (D-6, §6). Swift copy is an Out row for S10-54 (X-10).

### M-4 · major · `registryOnly` is a deny-list
**Answer:** Fixed. D-3 adds `researcherCreatable`, matching the Swift list; T-1 also refuses an unknown type.

### M-5 · minor · Third copy of the owned-value effects block
**Answer:** Fixed. D-6 adds `effects.ownedValue`, and the three entries use it.

### M-6 · minor · The depiction update case differs from date and name
**Answer:** Fixed with X-13. D-9 uses the date branches: with no existing value, the insert path runs. AC-2.7 and T-7 cover both directions.

### M-7 · minor · `depictionvalues.LookupManyTx` has no caller
**Answer:** Fixed with X-11. Dropped and logged in Out.

### M-8 · minor · The "every reader" Risk row is false
**Answer:** Fixed. The Risks row names the readers that skip a likeness on purpose, and an Out row forbids adding branches to them. The declared value-column list is logged here, not planned: one list driving the SQL column lists, scans, and audit maps in `observations` would turn about 12 edits per value column into one entry.

### X-1 · blocker · AC-2.5 and T-9 cannot pass
**Answer:** Fixed (see M-2). The Stop condition now names T-8.

### X-2 · major · D-5's page comparison leaves cases undecided
**Answer:** Fixed. D-5 states every case; T-3 has a row for each.

### X-3 · major · STEP-6 and STEP-7 split `update.go`
**Answer:** Fixed. All `update.go` work is in STEP-7; STEP-6 says the depiction case refuses `existing != nil` until then. The data-loss note does not apply: `sqlUpdateObservation` names its columns, so an update before STEP-7 leaves `value_depiction_id` untouched.

### X-4 · major · The manual check is false; the app mishandles `likeness`
**Answer:** Partly declined, with evidence. `PropertiesValueTypePill` prints `Text(verbatim: valueType)`, so the pill does show `depiction`. Only VoiceOver reads *Text*, through `SubjectPropertyValueType.label`. The composer filters on `supportedValueTypes`, which has no `depiction`. The Person field list is built from `auto_reconciler_outcomes`, which has no likeness row while D-11 holds. So nothing in the app offers Likeness as free text. Fixed: manual check 1 states the VoiceOver interim, a Risks row accepts it, and §4 has a Swift app row.

### X-5 · minor · D-8 needs the Citation locator too
**Answer:** Fixed. One query returns `a.file_id, c.locator_json`.

### X-6 · minor · `listedFromSource` is not covered
**Answer:** Fixed. D-10 and STEP-6 copy `DepictionID` into `ValueDepictionID`. An Out row says the source-graph read has no `Depiction` until S10-54.

### X-7 · minor · STEP-9 anchors
**Answer:** Fixed. STEP-9 names the DDL sketch, the index, the exclusivity sentence, and the `add-audit-scope` paragraph.

### X-8 · minor · Reader coverage
**Answer:** Fixed with M-8.

### X-9 · minor · Change order
**Answer:** Fixed. Insert emits `observation`, `file`, `depiction_value`. Update emits the observation change first, then the value changes, with release delete changes in `extra` as today (D-8, D-9, AC-1.1).

### X-10 · minor · Error code name and Swift copy
**Answer:** Fixed with M-3.

### X-11 · minor · `LookupManyTx`
**Answer:** Fixed with M-7.

### X-12 · minor · STEP-1 check
**Answer:** Fixed. The check names the unregistered table and both FKs.

### X-13 · nit · D-9 departs from the date precedent
**Answer:** Fixed with M-6.

### X-14 · minor · Test fixtures
**Answer:** Fixed. STEP-6 adds `likeness` to `mustSeed`'s `seed`. T-10 gives the order: cite, raw rows, promote.

### X-15 · minor · Map row and plan disagree on staging
**Answer:** Fixed in the map now: the S10-32 row says staging happens before the caller's write.

## Round 2

Both reviewers accepted every round-1 answer.

### M-10 · blocker · Image staging in a query package that would import a use-case package
`.agents/skills/add-catalog-query/SKILL.md` "Do not": put ingest in the query package; `core/ingest` imports `core/writes`, and no `core/database` package imports either today.
**Answer:** Fixed. D-7: `observations` declares `StagedFile` (only `InsertTx`) and `DepictionInput.Image`; `StageDepictions` moves to `core/ingest/depictions.go`, so `ingest` imports `observations`. AC-1.7 is `Image == nil`. The map's S10-32 row names `ingest.StageDepictions`. The Stop condition checks that no `core/database` package imports `ingest` or `writes`. Declined in part: `ImagePath` stays on `DepictionInput` as a plain string the write never reads; moving it out would need a second slice parallel to `inputs` in every handler (S10-54).


### X-16 · major · A test crop could equal the Artifact's own File
**Answer:** Fixed. STEP-6's `crop` takes a gray shade and `fileArtifact` uses red, so a staged crop never equals the Artifact's File; AC-2.4 reuses one size and shade on purpose.

### X-17 · minor · STEP-6's check skips the `api` edit
**Answer:** Fixed. STEP-6 also tests `./api/ffi/handlers/...`.

### X-18 · minor · Undeclared fields and signatures
**Answer:** Fixed. D-8 lists the three new `resolvedValue` fields; T-10 passes an `at` label to `assertUpkeepEqualsRebuild` and names the `files` columns.

## Earlier shape (superseded)

**Verdict (superseded shape):** Fits with changes — Ready with changes (all applied)
**Rounds:** 2. Round 1 reviewed the first shape (the loader skipped depictions). The user then asked for a shape that keeps every value type reconciled; the plan was re-planned with a `depiction` module, and round 2 reviewed it in full. **Reviewers:** maintainability (opus), executability (sonnet)

### To user

None open. Answered 2026-10-16:
- `likeness` is new catalogs only; a later one-off CLI updates the dogfood projects (working rule "`likeness` is new catalogs only").
- Depictions are reconciled, not skipped (the user questioned the exclusion).
- The gallery shows `weak` crops (working rule "A gallery is the reconciled likeness values").

### Writer's own calls

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

### Round 1

#### M-1 · major · Domain doc made false: "Every value type needs an auto-reconciler"
`docs/conclusion-reconciliation.md` §1, §5 and §7, and the value-population sentence in `docs/conclusion-layer-data-model.md`, become false under D-7.
**Answer:** Fixed. §4 has a "Reconcile docs" row. STEP-7 gives the exact edits for both docs, and the DoD names them.

#### M-2 · major · The new Swift label goes into a string switch whose `default` caused the mislabel
`SubjectPropertyValueType.label` switches on raw strings, and `.cursor/rules/macos-client.mdc` says views use the typed `PropertyValueType`. The pill shows the raw key, so only VoiceOver used the label.
**Answer:** Fixed. D-10 now switches exhaustively over `PropertyValueType`, and `researcherCreatable` becomes `[PropertyValueType]` with call sites named. §4 has a Pill row. AC-3.5, §6 and the manual check now say the pill shows `depiction` and VoiceOver reads *Depiction*. `valueTypeForeground` and `CitationObservationValue` are Out lines (default colour, and S10-34).

#### M-3 · minor · The registry-only rule lives in two hand-kept lists
**Answer:** Fixed for Go. D-6 adds `registryOnly`, and `Create` calls it. The Swift↔Go parity subtest for `researcherCreatable` is logged here and in Out, not planned. Swift keeps the same five types, now typed.

#### M-4 · nit · T-6 pins a non-behavior that S10-36 reverses
**Answer:** Fixed. T-6 carries a comment that S10-36 replaces the assertion.

#### X-1 · major · T-5 reads `f.props["likeness"]`, which the fixture never loads
**Answer:** Fixed. STEP-5 adds `"likeness"` to the `newFixture` key loop. T-5 names the Subjects, `nameIn` with its spec, and the inputs it uses.

#### X-2 · minor · T-2 does not say how to build its locator and counts
**Answer:** Fixed. T-2 uses `mustSeed`'s artifact and person, `citations.CreateWithObservations` through `writes.Run`, and raw `SELECT`s on `c.DB()`.

#### X-3 · major · STEP-6's check cannot run off macOS, and its Swift cases are underspecified
**Answer:** Fixed. STEP-6 says `xcodebuild` is macOS only and what to report when it does not run. T-9 asserts through `L10n.string`. The xcstrings entry copies the `term` block and goes after `properties.valueType.date`. The `PropertiesModelTests` `term` line is updated.

#### X-4 · minor · STEP-7 docs leave the wording to the implementer
**Answer:** Fixed. STEP-7 gives each sentence, the §3.2 and §3.3 rows, the map lines, and the `completed.md` shape.

#### X-5 · minor · "Nothing is written" has no assertion form
**Answer:** Fixed. T-3 compares the `citations` and `observations` counts before and after.

#### X-6 · nit · The Swift doc comment goes stale
**Answer:** Fixed. D-10 updates the `SubjectPropertyValueType` doc comment.

### Round 2 (reconciled shape)

#### M-5 · major · The loader's `hasValue` is a copy of `autoreconcile.carries`
**Answer:** Fixed. D-7 exports `Carries`. D-12 has `load` call it and deletes `hasValue`. Codebase fit and **Deletes** list both.

#### M-6 · major · The gallery's "kept or weak" rule lives only in prose
**Answer:** Fixed, by changing the map. The S10-35 row now binds the rule to one `autoreconcile.Reason` method beside `Displayed`, which S10-39 reuses, and the Out row points there. This PR adds no method with no caller.

#### M-7 · minor · `knownValueType` is a third hand-kept list
**Answer:** Fixed. `knownValueType(vt)` is `moduleFor(vt) != nil`. Declined: an exported `properties.ValueTypes` slice. It reaches beyond this PR's layers, and Out logs it.

#### M-8 · minor · Stale doc and code comments
**Answer:** Fixed. STEP-9 adds the "Read-side tolerance" clause. D-14 updates the `attachValues` comment, and D-7 updates the package doc.

#### M-9 · nit · The loader branches on value type
**Answer:** Fixed. `sqlLoadCandidates` selects `CASE WHEN p.value_type = 'depiction' THEN c.id END` (D-12).

#### M-10 · nit · Bookkeeping
**Answer:** Fixed. D-12 to D-14 follow D-11, and this log restates the writer's calls.

#### X-1 · major · T-5's `Verify` is vacuous
**Answer:** Fixed. T-5 loads `Graph().Node(h)` before the last write and again before `Verify`, and asserts the node's likeness `CitationID`s.

#### X-2 · major · T-5 lifecycle and setups
**Answer:** Fixed. Each subtest gets its own fixture. B is filed only by `identityclaims.Create` in AC-2.4. AC-2.3 gives the `promote.Input` for `c`. `f.credibility` is wrapped in `must`, and rows are looked up by `ValueCitationID`.

#### X-3 · minor · STEP-6 bundles five changes
**Answer:** Fixed. STEP-6 is the cache and graph mirror; STEP-7 is the Person field skip and the Promote guard. There are now 9 steps.

#### X-4 · minor · T-11 does not exercise version 14
**Answer:** Fixed. T-11 adds `14` to the stale-version list.

#### X-5 · minor · T-1 overlaps an existing row
**Answer:** Fixed. T-1 extends "create refuses term value type" into a loop and adds an Upsert row.

#### X-6 · minor · Small choices in T-12, T-6, and T-2 to T-4
**Answer:** Fixed. T-12 is its own function, using `properties.Lookup`. T-6 spells out `srcB, artB`. T-2 to T-4 are table rows.

#### X-7 · minor · Docs and log drift
**Answer:** Fixed. STEP-9 covers line 891. This log and the header are current. The draft PR is opened before STEP-9.

#### X-8 · minor · The DoD box off macOS
**Answer:** Fixed. The DoD line says how to report it.

#### X-9 · minor · No old catalog for manual check 4
**Answer:** Fixed. The check builds `main` at the Planned commit first, or reports itself as not run.

#### X-10 · nit · Others
**Answer:** Fixed: `canSubmitCreate`; STEP-4 to STEP-6 land together; Out names the proto comment and the inspector note. The `.contains(.term)` edit stays, because the old line no longer compiles.
