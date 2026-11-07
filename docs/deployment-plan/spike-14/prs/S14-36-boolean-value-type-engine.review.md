# Review — S14-36 Boolean value type (engine)

**Verdict:** Fits with changes — Ready with changes (round 2 raised four blockers: all Fixed; no open blocker)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **M-2 (answered 2026-11-07).** How Swift learns the value-type rules (creatable, single-only). The writer recommended a pinned Swift mirror. The user chose: **Go sends the rules on the wire.** The plan's D-10 and D-11 carry it: `GetPropertiesWorkspaceResponse.value_types`, Swift's own list deleted.

## Writer's own calls

- **D-5:** Boolean ships `ResearcherCreatable: false`, and S14-37 flips it. This follows from the user's M-2 answer: Swift now offers whatever Go marks creatable, so `true` here would show boolean in the Properties picker before its composer editor exists. It moves "researcher-creatable in Go" from S14-36's map scope to S14-37's (STEP-10).
- **D-11 cold load:** before the first workspace load, the create picker offers the draft's "text", and Go validates. ⌘N and the header Add button open create on a cold page, so Swift cannot wait for the rules.
- **D-1:** a dedicated `value_boolean` column (the working rule left this to `plan-pr`).
- **D-9:** negative polarity gets no boolean-specific rule.
- **D-13:** a boolean's delete-impact title is its Property label, with no Go "yes"/"no".
- **D-14:** the update diff stays per column, not derived from row maps.
- **D-12 (round 2):** FakeStore carries no default value-type rules; tests seed them.
- **Map:** planning moved "researcher-creatable in Go" and the Holds-several hide to S14-37 and rewrote the Batch 7 boolean working rule (same commit as this log).

## Round 1

### M-1 · blocker · The set of value types and their per-type rules has no owner, and this PR adds a fourth hand-kept list
**Plan:** STEP-2 / D-4 / D-5 / STEP-9. Boolean goes into the `valueTypeOK` and `researcherCreatable` switches, a new `cardinalityAllowed` wrapper, and the hand-written "value types" want list.
**Codebase:** `properties.go` `valueTypeOK` and `researcherCreatable`; `swift_vocabulary_test.go` "value types" row; the stale `engine.proto` `Property.value_type` comment. The pattern that fixes it is `vocabkey.PrimaryKinds()`.
**Better shape:** a declared table in `propertymeta` that the rules and the test read.
**Answer:** Fixed. D-4 adds `propertymeta.TypeRule`, `Types`, `ValueTypes`, `Rule`. D-6 has `properties` read the table, D-20 has the test read `ValueTypes()`, and D-10 rewrites the proto comment. T-1 covers the table.

### M-2 · blocker · Swift copies two engine rules about value types; D-5 opens a drift window, and S14-37 will add a third copy
**Plan:** D-5 (old) left Go and Swift creatable lists diverging; D-4 (old) made single-only Go-only.
**Codebase:** `PropertiesModel.swift:7` `SubjectPropertyValueType.researcherCreatable`; `PropertiesView.swift` ~527 offers "Holds several" for every Property.
**Better shape:** Go owns both facts; Swift reads them from the wire, or a pinned mirror.
**Answer:** To user, then Fixed per the user's answer (wire). D-10: proto `PropertyValueTypeRule` and `GetPropertiesWorkspaceResponse.value_types`. D-11 and D-12: the snapshot carries the rules, Swift's list is deleted, FakeStore seeds them. D-5: boolean not creatable until S14-37. Hiding "Holds several" is an Out row for S14-37, since it changes the Properties page.

### M-3 · blocker · `observations` builds the Observation audit row in two places, and the resolved-value fill in two more
**Plan:** STEP-3 threads `value_boolean` through the inline insert row map, `observationRowMap`, the update diff and two fills.
**Codebase:** `observations.go` ~376–392 and ~420–440; `write.go` `observationRowMap`; `update.go` ~118–132; `deleteimpact/facets.go` `eraseObservation`.
**Answer:** Fixed for `observations` (D-14): the insert records `FullRow(observationRowMap(obs))`, and `applyResolved` serves insert and update. Declined on deriving the update diff, with evidence: `observationUpdateChange` records date and name changes on their own value rows, not as column diffs. The `deleteimpact` copy cannot import `observations`, so it is logged as a structural-debt row in STEP-10.

### M-4 · major · D-3's `nullBool` stays private to one package and its signature doesn't match its sibling
**Answer:** Fixed. D-3 is now `database.NullBool(p *bool) any` in `core/database/null.go`, used by the Observation writes and `autoreconciler.insertRow`. `resolvedValue.Boolean` is `*bool`.

### M-5 · major · "Yes"/"No" sit inside `ReconciledValueDisplay` under `L10n.Conclusions`
**Answer:** Fixed. D-18 adds `PropertyValueType.booleanText(_:)` with `L10n.Properties.valueBooleanYes` and `valueBooleanNo`; S14-37's `ObservationValueDisplay` calls the same function. The existing integer divergence is logged as a structural-debt row (STEP-10).

### M-6 · major · The cache value columns are copied by hand across three `graphcache` structs, `promotealign.decode` and `conclusiondetails`
**Answer:** Fixed as a log, as the reviewer's Fit says. STEP-10 adds a structural-debt row naming the sites; the refactor is an Out row.

### M-7 · major · STEP-10's doc sweep misses a skill and a domain doc
**Answer:** Fixed. STEP-10 and the DoD now list `.agents/skills/add-seeded-vocabulary/SKILL.md` (~58) and `docs/matching-model.md` (~49).

### M-8 · minor · The value-type pill colour is a string switch the Reach row doesn't check
**Answer:** Fixed. The Reach row names `SubjectTypeChrome.valueTypeForeground`; the colour is an Out row for S14-37, because it is chrome.

### M-9 · nit · `optionalBool` next to `optionalInt64`
**Answer:** Logged. Out row "not planned".

### X-1 · major · D-2 and the Go core table disagree on field names
**Answer:** Fixed. D-2 now gives the naming rule: `ValueBoolean` in `observations`, `Boolean` everywhere the sibling is `Integer`.

### X-2 · major · No helper for binding a boolean in the autoreconciler cache
**Answer:** Fixed with M-4 (D-3 names `insertRow`).

### X-3 · major · The contract-test sample and generator are not specified
**Answer:** Fixed. STEP-4 now spells out `boolValue`, `ofBool`, the samples, the generator, and why there is no "no value" pair.

### X-4 · minor · STEP-8 file list is missing `PropertiesModelTests.swift`
**Answer:** Fixed. STEP-8 lists it and says how to assert copy (`L10n.string`). D-17 places `valueBoolean` after `valueInteger`, and §6 Strings places the keys beside `valueTypeInteger`.

### X-5 · minor · STEP-10 doc instructions are partly wrong
**Answer:** Fixed. STEP-10 now covers:
- the already-struck structural-debt row ("Done in S14-36.")
- both "`real` and `boolean` are not" sentences
- the stale storage clause in the working rule, given its replacement text
- the map status is in the DoD

### X-6 · minor · Checks I cannot run in this environment
**Answer:** Fixed as a note. §6 FFI says that `generate-proto.sh` needs the protoc toolchain on a Mac, and that both generated files are committed.

### X-7 · minor · Behaviors the plan does not specify
**Answer:** Fixed. D-6 places `cardinalityAllowed` where `cardinalityOK` is today, before the origin lock check. D-13 names `facets.go` only, and §4 notes that `observations.Delete` audits through `observationRowMap`. AC-3.4 is held by T-7.

### Writer's own find after round 1
The old Reach row said the create form opens only from a loaded page. ⌘N and the header Add button (`PropertiesView` `header`) open it at any time. D-11 now falls back to the draft's type and lets Go validate before the first load, and T-12 covers it.

## Round 2

Reviewers: maintainability (opus) — Fits with changes; executability (sonnet) — Not ready. Both checked every round-1 answer. They found the answers sound except M-1 for the tests (M-10), M-2 for the fake (M-11), the partial M-7 and the M-8 symbol name.

### X-8 · blocker · Subject-profile conversion omits Boolean
**Codebase:** `core/subjectprofile/subjectprofile.go` `Value` copies `Integer` into `valuecompare.Value` and is missing from §4.
**Answer:** Fixed. It is in §4 and the Go core table, and STEP-5 copies `Boolean: v.Boolean`.

### X-9 · blocker · The "source surfaces" want expression does not compile, and its import is not named
**Answer:** Fixed. D-20 now gives `sorted(append([]string{…three constants…}, swiftOnlySurfaces...)...)` and the `workspacelocation` import.

### M-10 · blocker · The new value-type table is not wired into the three tests that list every value type
**Codebase:** `contract_test.go` `allTypes` (a missing type is skipped silently), `modules_test.go` `TestModuleFor`, `valuecompare_test.go` `TestCompareUnsupported`.
**Answer:** Fixed. STEP-4: `allTypes` = `propertymeta.ValueTypes()`, and `TestCompareUnsupported` wants `ValueTypes()` minus a declared `noComparer`. STEP-5: `TestModuleFor` ranges over `ValueTypes()`. §4 records the silent skip.

### M-11 · blocker · `FakeStore.defaultValueTypeRules` is an unpinned Swift copy of `propertymeta.Types`
**Answer:** Fixed with the reviewer's second option. D-12: FakeStore `valueTypeRules` is empty unless a test seeds it, so there is no default list. T-12 and the create-flow test seed the rows themselves.

### X-10 · major · AC-4.5 / T-14 needs a private function
**Answer:** Fixed. T-14 round-trips through the existing `GoStore.updateObservationRequest`, and `mapCatalogObservation` stays private (D-17).

### X-11 · major · Adding `PropertiesSnapshot.valueTypeRules` breaks call sites the plan does not list
**Answer:** Fixed. D-11 declares `valueTypeRules = []` last, so the memberwise call sites compile unchanged.

### X-12 · major · T-12 and the replaced assertions do not say how the snapshot gets loaded
**Answer:** Fixed. STEP-8 spells out both replacements (seed then assert at ~273; drop the line at ~797), and T-12 seeds `store.valueTypeRules`.

### X-13 · minor · Cache column position is only half-specified
**Answer:** Fixed. `value_boolean` sits directly after `value_integer` in every SELECT, scan and insert.

### X-14 · minor · Handler-test file, `valueInput` position, and "is it new" markers
**Answer:** Fixed. `propertymeta_test.go` is an existing file. T-11 is in `subject_defs_test.go`. D-16 gives the call shape. FakeStore needs no boolean Observation handling (no editor until S14-37).

### X-15 · minor · Doc steps
**Answer:** Fixed. STEP-10 now gives:
- the exact `matching-model.md` wording
- the skill's Create sentence
- the working rule's "from S14-37"

### M-7 (partial), M-8 (symbol)
**Answer:** Fixed. The skill line 59 is in STEP-10, and the symbol is `PropertiesTypeChrome.valueTypeForeground`.

### Round-2 below-blocker notes
- **Existing test uses "boolean" as its invalid type:** Fixed. `properties_test.go` "rejects bad value type" uses `"real"` (STEP-2).
- **`conclusion-layer-data-model.md` ~362:** Fixed. It is the claims sketch and already has `value_boolean`, so the edit was removed.
- **`real` sentence:** Fixed. It is reworded, not deleted.
- **Guard written twice:** Fixed. It is one `PropertiesModel.acceptsValueType(_:)`.
- **`nullInt64` → `database.NullInt64`:** Declined. It moves a working private helper with four call sites that this PR does not otherwise change. The new column's binder is shared, which was M-4's point.
- **Fold `SubjectPropertyValueType.label` into `PropertyValueType`:** Declined for this PR. It renames call sites outside the boolean path; S14-37 edits these files and can take it.
