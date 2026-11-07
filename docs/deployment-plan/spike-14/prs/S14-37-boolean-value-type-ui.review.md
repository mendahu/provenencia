# Review — S14-37 Boolean value type (UI)

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1 (round 2 runs once S14-39 lands and the S14-D8 board is agreed). **Reviewers:** maintainability (opus), executability (sonnet). Both read the plan and brief S14-D8.

## To user

- **M-1** Nest the composer's value fields in this PR, a preparatory PR, or log it? Writer recommended this PR. **Answered (2026-11-07): this PR.** → D-4, STEP-3.
- **M-2** Where does "a boolean has no Negative" live? Writer recommended one Swift function. **Answered (2026-11-07): a Go `TypeRule` field.** S14-36 had merged, so the field is added here, and the plan was refreshed against `main` @ `aaf3b15`. → D-1, D-2, D-3, D-6.
- **M-4** Give `PVSegmentedControl` a focus handle, or rely on the key-view loop? Writer recommended the kit change. **Answered (2026-11-07): the kit change.** → D-9, D-10, STEP-5.

## Writer's own calls

- **D-2.** Go refuses a negative Observation of a non-negatable type, on both add and update. The user asked for the rule field; enforcing it in the engine is the writer's call.
- **D-6 / AC-3.4.** With no rules loaded, the composer offers Negative, as today, and Go decides.
- **D-7.** Re-committing the same Property in the combo box now keeps the value and polarity, for every value type, not only boolean. Today a Shift-Tab, Tab round trip clears a typed value.
- **D-8.** `applyText(rowID:text:)` stays, because two row-edit tests drive it. The other dead row paths and their two strings are deleted.

## Board check (2026-11-07)

- The writer read the drawn board (export and the migrated design) against the brief. Frames 1–5, 7a, 7b, 8a and 8b match BV-1 – BV-7 and BV-W1.
- Frame 6 drew the brief's own BV-4 exception: a negative yes/no from an outside catalog, shown with its checkbox. **User (2026-11-07): a yes/no is never negative, no exception.**
- Fixed in the brief: BV-4 and BV-6 have no exception, and frame 6 is to be removed from the board.
- Fixed in the plan: D-6 `offersNegative(propertyID:)` drops polarity, and `asDraft` always writes a boolean positive. D-7 no longer resets polarity. AC-2.4, AC-3.3, T-4 and T-6 are restated.
- The board is not agreed yet.

## Round 1

### Maintainability

**M-1 · major · Fields and ObservationRowValues hold the same seven value fields.** The reviewer proposed nesting the value part. **Answer:** To user, then Fixed per the answer: D-4 and STEP-3 (a rename only, before any boolean code).

**M-2 · major · The no-Negative rule is split between the view and the model, and can't be unit-tested.** **Answer:** To user, then Fixed per the answer:
- The rule now lives in Go: `TypeRule.Negatable` (D-1), enforced by `observations.negationAllowed` (D-2) and sent on the wire (D-3).
- Swift has one reader, `CitationComposerVocabulary.offersNegative` (D-6). Both the view (D-9) and the Property change (D-7) call it.
- AC-3.1 – AC-3.4 are unit tests (T-6).

**M-3 · major · Dead `fields(from:)` and twin mutation paths.** **Answer:** Fixed. D-8 deletes `fields(from:)`, `applyProperty`, `applyInteger`, `applyTerm`, `togglePolarity`, `polarityMenuTitle` and their two strings and one test. `applyText` stays (test-driven; see Writer's calls). The DoD sweep names them.

**M-4 · major · D-9 bent `propertyValueTakesFocus` instead of giving the kit control a focus handle.** **Answer:** To user, then Fixed per the answer:
- D-10 adds `focus: PVFieldFocus? = nil` to `PVSegmentedControl`.
- D-9 passes `.value` like every other value editor, and `propertyValueTakesFocus` keeps its meaning.
- The brief's inventory row is now **Extend**.

**M-5 · minor · Bar outcome 31 says "checkbox"; S14-36's lines route S14-39's work to S14-37.** **Answer:**
- Fixed: bar outcome 31 now reads "a Yes / No choice" (this planning commit).
- Declined for the S14-36 plan: it has landed (#733) and its Out rows are a record. The map is the copy, and it routes the Properties page to S14-39.

**M-6 · minor · `isBoolean` starts a family of one predicate per type.** **Answer:** Fixed. D-6 adds `valueType(propertyID:)`; `isDepiction` stays (its call sites are not this PR's).

**M-7 · minor · FakeStore builds Observations in two near-copies.** **Answer:**
- The landed S14-36 already copies `valueBoolean` in both, so this PR does not edit FakeStore.
- Logged instead: STEP-8 adds a structural-debt row for the merge and the `displayText` chain.

**M-8 · minor · Two text paths for a saved Observation.** **Answer:** Logged. STEP-8 adds `summary` to S14-36's display row.

### Executability

**X-1 · blocker · The DoD sweep can never come back empty.** **Answer:** Fixed. Both sweeps are scoped to `docs/interpretation-layer-data-model.md` and `.agents/skills`, and they match "S14-37" because the landed doc says "until S14-37".

**X-2 · major · The fixture is indexed by position.** **Answer:** Fixed. STEP-4 appends Can read last (index 8) and binds it with `![8]`.

**X-3 · major · FakeStore records no draft.** **Answer:** Fixed. AC-2.3 and T-3 assert on `observationsByGraph` and the `recordedCalls` prefix.

**X-4 · major · No Can read Observation helper.** **Answer:** Fixed. STEP-4 adds `canReadObservation(...)`, with `valueBoolean:` after `valueInteger:`. AC-2.7 reads `state == .saved`.

**X-5 · minor · `Fields.init` placement; dead `fields(from:)`.** **Answer:** Fixed by D-4 (no new init: the row holds a `Fields`) and D-8 (deleted).

**X-6 · minor · `ObservationRow` pass-through for `valueBoolean`?** **Answer:** Fixed. D-4 deletes all seven forwarding accessors; callers read `row.values.valueFields`.

**X-7 · minor · A second Property-change path skips the polarity reset.** **Answer:** Fixed. D-8 deletes `applyProperty(rowID:)`.

**X-8 · minor · Re-commit wipes a chosen answer.** **Answer:** Fixed. D-7 returns early on the same Property; AC-2.6 and T-4 pin it.

**X-9 · minor · The Tab claim is unverifiable.** **Answer:** Fixed. D-10's focus handle replaces the key-view loop, and STEP-6 runs manual check 3 before STEP-7. A Stop condition names the failure.

**X-10 · minor · The `properties.Create` call shape; S14-36 tests not on `main`.** **Answer:** Fixed. S14-36 has landed; AC-1.2 uses the real signature and STEP-1 names the real test rows.

**X-11 · minor · AC-1.4 has no direct assertion.** **Answer:** Fixed. T-1 asserts the boolean row by name.

**X-12 · minor · S14-36's Out table is stale.** **Answer:** Declined, as for M-5: the plan has landed, and the map is the copy.

**X-13 · minor · Width and checkbox placement.** **Answer:** Fixed. D-9 says no extra frame, and puts the checkbox in an `if` so the `VStack` closes up.

**X-14 · nit.** Logged. The §4 Fake row is now verified against the landed code.
