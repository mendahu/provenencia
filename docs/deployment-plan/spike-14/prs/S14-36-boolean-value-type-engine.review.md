# Review — S14-36 Boolean value type (engine)

**Verdict:** round 1: Fits with changes — Ready with changes. Round 2 pending.
**Rounds:** 1 (round 2 to run). **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **M-2 (answered 2026-11-07).** How Swift learns the value-type rules (creatable, single-only). The writer recommended a pinned Swift mirror. The user chose: **Go sends the rules on the wire.** The plan's D-10 and D-11 carry it: `GetPropertiesWorkspaceResponse.value_types`, Swift's own list deleted.

## Writer's own calls

- **D-5:** Boolean ships `ResearcherCreatable: false`, and S14-37 flips it. This follows from the user's M-2 answer: Swift now offers whatever Go marks creatable, so `true` here would show boolean in the Properties picker before its composer editor exists. It moves "researcher-creatable in Go" from S14-36's map scope to S14-37's (STEP-10).
- **D-11 cold load:** before the first workspace load, the create picker offers the draft's "text", and Go validates. ⌘N and the header Add button open create on a cold page, so Swift cannot wait for the rules.
- **D-1:** a dedicated `value_boolean` column (the working rule left this to `plan-pr`).
- **D-9:** negative polarity gets no boolean-specific rule.
- **D-13:** a boolean's delete-impact title is its Property label, with no Go "yes"/"no".
- **D-14:** the update diff stays per column, not derived from row maps.

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
