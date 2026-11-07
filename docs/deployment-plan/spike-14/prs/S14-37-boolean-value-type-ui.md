# S14-37 — Boolean value type (UI)

**Status:** Blocked — on S14-36 (engine code not on `main`), S14-39, board S14-D8
**Spike / slice:** Spike 14 · Slice 7e — Boolean value type
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S14-D8](../design/S14-D8-composer-boolean-value.md), board not drawn yet
**Board:** the Claude Design project [*Citation Composer*](https://claude.ai/design/p/8b120853-4ba4-4c0d-8492-75e76b9f9b7a?via=share), section "S14-D8 · Yes / no value". The implementer cannot open Claude Design: **Design → macOS** describes each frame in words (user, 2026-11-07), and is refreshed from the agreed board.
**Depends on:** S14-36 (planned; PR [#732](https://github.com/mendahu/provenencia/pull/732) merged its plan only), S14-39 (scoped), S14-D8
**Blocks:** —
**Branch:** `s14-37-boolean-value-type-ui`
**Version bump:** minor — researchers can create and cite a new kind of Property, a feature ([`docs/versioning.md`](../../../versioning.md)); the next minor after `VERSION` at merge (0.3.8 → 0.4.0 at planning)
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (only to read the `propertymeta.Types` row it owns), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** working rule "(Batch 7) Boolean is a first-class value type in two PRs"; working rule "(Batch 5) FakeStore is a seeded fake, not a second engine"; bar outcome 31; dogfood note [`ux/2026-10-25-boolean-property-value-type.md`](../ux/2026-10-25-boolean-property-value-type.md); S14-36 plan [`S14-36-boolean-value-type-engine.md`](S14-36-boolean-value-type-engine.md) (D-4, D-11, D-17, D-18, D-19 and its Out rows)
**Planned:** 2026-11-07 against `main` @ `2ee4adb` (draft: S14-36's code is not on `main`; every row marked *(S14-36)* names what its plan adds)
**Review:** — (not reviewed yet)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

A researcher can make a yes/no Property ("Can read", "Veteran") on the Properties page and cite it from a census column in the Citation composer. The composer's Value field becomes a Yes / No segmented control with nothing chosen at first, so a No is always one the researcher picked. A yes/no Property has no Negative checkbox, because No is how the record says it. The saved answer reads "Yes" or "No" in the composer's evidence column and on Evidence graph cards; Conclusion detail already reads it (S14-36). This completes bar outcome 31 and the dogfood note of 2026-10-25.

**Why now.** S14-36 stores, reconciles and sends booleans; S14-39 makes the Properties page honour the single-only rule. Boolean stays uncreatable until the composer can take the value: this PR flips it.

**What stays.** Every other value type's editor and its Negative checkbox. The observation dialog's layout and the evidence column's line layout. The Properties page (S14-39 owns it). Conclusion detail's value text (S14-36).

## 2. User stories

**US-1.** As a researcher, when I create a Property, I can choose the yes/no type.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `propertymeta.Rule("boolean")` | Found; `ResearcherCreatable` true, `SingleOnly` true |
| AC-1.2 | `properties.Create(label "Can read", valueType "boolean", cardinality "")` | Created, `cardinality = single` |
| AC-1.3 | `properties.Create(label "Can read", valueType "boolean", cardinality "multiple")` | `ErrInvalid`; no row |
| AC-1.4 | `GetPropertiesWorkspace` | `value_types` row `boolean` has `researcher_creatable` true |

**US-2.** As a researcher, when I cite a yes/no Property in the composer, I pick Yes or No, and nothing is saved until I do.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The Property combo box of a Person Subject whose type binds "Can read" (boolean) | "Can read" is offered (it was hidden before: `supportedValueTypes`) |
| AC-2.2 | A new property draft, Property set to "Can read" | `values.valueBoolean == nil`; the dialog cannot save (`propertyValuesComplete` false) |
| AC-2.3 | AC-2.2, then `valueBoolean = false`, then Save | `addObservationsToCitation` (or `createCitationWithObservations`) is called with a draft whose `valueBoolean == false`; the saved row reads "No" in the evidence column |
| AC-2.4 | A draft with Occupation, Negative ticked, then Property changed to "Can read" | `polarity == positive`, `valueBoolean == nil` |
| AC-2.5 | A draft with "Can read" and `valueBoolean = true`, then Property changed to Occupation | `valueBoolean == nil`, `valueText == ""` |
| AC-2.6 | Edit on a saved "Can read: Yes" Observation | The edit dialog's values have `valueBoolean == true`; Save of an unchanged edit does not call `updateObservation` (the row is not dirty) |
| AC-2.7 | `CitationObservationValue.summary(valueType: "boolean", …)` with `true`, `false`, `nil` | "Yes", "No", "" |

**US-3.** As a researcher, a yes/no Property shows no Negative checkbox, unless the line is already negative.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The dialog with Property "Can read", polarity positive | `CitationPropertyFields.showsNegative` *(new)* false |
| AC-3.2 | The dialog with Property Occupation | `showsNegative` true |
| AC-3.3 | Edit on a saved "Can read" Observation with polarity negative (catalog made outside the app) | `showsNegative` true, checkbox checked; the evidence line shows "not" and the Negative Badge |

**US-4.** As a researcher, a saved yes/no answer reads "Yes" or "No" wherever an Observation's value is shown.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `ObservationValueDisplay.string(for:)` an Observation with `propertyValueType "boolean"`, `valueBoolean true` | "Yes"; with `false` → "No" |
| AC-4.2 | The same with `valueBoolean nil` and `valueText ""` | "" (as today) |
| AC-4.3 | A Person with two Citations both saying "Can read: Yes", on Conclusion detail | "Yes", merged (S14-36's text; checked by hand here) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Go marks boolean researcher-creatable (`propertymeta.Types`) | STEP-1, AC-1.1 – AC-1.4 |
| Composer draft value: `valueBoolean` through the row, the value fields, summary, validity and save draft; one `Fields` constructor | STEP-2, AC-2.2, AC-2.3, AC-2.5, AC-2.6, AC-2.7 |
| Composer offers boolean Properties; Property change resets polarity for a boolean | STEP-2, AC-2.1, AC-2.4 |
| FakeStore keeps `valueBoolean` on the Observations it stores | STEP-2, AC-2.3, AC-2.6 |
| Yes / No control in the Value field (BV-1, BV-2, BV-3, BV-5, BV-7) | STEP-3, AC-2.2, AC-2.3 |
| Negative checkbox hidden for a boolean (BV-4) | STEP-3, AC-3.1 – AC-3.3 |
| Saved line text (BV-6) | STEP-2 (summary), AC-2.3, AC-2.7 |
| Wide (BV-W1) | manual check 6 |
| Evidence graph card text | STEP-4, AC-4.1, AC-4.2 |
| Docs: data model, `add-seeded-vocabulary`, map, completed, brief archive | STEP-6 |

| Out | Home |
| --- | --- |
| Properties page: Holds several hidden for a single-only type; boolean pill colour; how the create picker looks with "boolean" in it | S14-39 (brief S14-D9) |
| A "not answered" or "unknown" value | not planned (a blank column is not cited; brief non-goal) |
| Keyboard shortcuts for Yes and No | not planned (brief §2.1) |
| A focus binding on `PVSegmentedControl` | not planned unless STEP-3's Stop condition fires; then the user decides |
| Omnibar indexing of boolean values | not planned (S14-36 Out) |
| Folding a negative boolean into its opposite value | not planned (S14-36 D-9) |
| Integer text differing between `ReconciledValueDisplay` and `ObservationValueDisplay` | [structural debt](../../../ideas/structural-debt.md) (row from S14-36) |

## 4. Current state

Verified at `2ee4adb`, plus S14-36's planned additions marked *(S14-36)*. Redo this table when S14-36 lands.

| Area | Where | Today |
| --- | --- | --- |
| Value-type rules (Go) | `core/propertymeta/propertymeta.go` `Types`, `Rule` *(S14-36 D-4)* | boolean row `ResearcherCreatable: false, SingleOnly: true` |
| Rule tests (Go) | `core/propertymeta/propertymeta_test.go` `TestTypes` *(S14-36 T-1)*; `core/database/properties/properties_test.go` `TestValueTypeRules` *(S14-36 T-2)*; `api/ffi/handlers/subject_defs_test.go` `TestGetPropertiesWorkspaceValueTypes` *(S14-36 T-11)* | T-1 asserts boolean not creatable; T-2 has a row "Create boolean → `ErrInvalid`"; T-11 compares with `Types` field for field, so it follows the flip |
| Swift value kinds | `macos/App/Platform/InterpretationValueKinds.swift` `PropertyValueType.boolean`, `booleanText(_:)` *(S14-36 D-17, D-18)* | Reads "Yes" / "No" from `L10n.Properties.valueBooleanYes` / `valueBooleanNo` *(S14-36)* |
| Swift models | `macos/App/Platform/GenealogyStore.swift` `CatalogObservation.valueBoolean`, `CatalogObservationDraft.valueBoolean` *(S14-36 D-17)* | `Bool?`, declared after `valueInteger`, default nil; `GoStore` maps both ways |
| Composer offered types | `macos/App/Features/CitationComposer/CitationComposerVocabulary.swift` `supportedValueTypes`, `isOffered(valueType:capabilities:)`, `isDepiction(propertyID:)` | No boolean: a boolean Property is not offered in the Property combo box |
| Composer row values | `macos/App/Features/CitationComposer/CitationObservationRows.swift` `ObservationRowValues` (fields, `clearValue`), `ObservationRow` (`isEmptyDraft`, `loaded(from:)`, `valueFields()`, `asDraft(property:)`, `asCatalogObservation(property:)`) | One field per value type; no boolean |
| Composer value logic | `macos/App/Features/CitationComposer/CitationObservationValue.swift` `Fields`, `fields(from:)`, `summary`, `isDialogValid`, `apply` | `switch valueType` per type, `default` refuses |
| `Fields` copies | `CitationObservationValue.fields(from:)`, `ObservationRow.valueFields()`, `CitationComposerModel+Authoring.swift` `propertyValuesComplete` (~124) | Three hand-written memberwise copies of the same seven fields |
| Property change | `CitationComposerModel+Authoring.swift` `applyObservationProperty(_:values:)` (~233) | Sets the Property and clears the value; polarity stays |
| Dialog form | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationPropertyFields` (`body`, `valueControls` ~366, `claimTab` ~448, bindings ~495–530) | Value switch per type; Negative `PVCheckbox` always drawn under the Value field |
| Tab order | `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` `CitationAuthoringDraftAfterCommit.propertyValueTakesFocus`, `CitationAuthoringNavigation.next` | False only for a depiction whose Draw is blocked; then the combo does not claim Tab |
| Saved lines | `CitationComposerAuthoring.swift` `groups(observations:…)` (~330); `CitationComposerSubjectGroups.swift` line row (~150–190) | Text from `summary`; `mono` for integer and date; "not" word and Negative Badge when negative |
| Card text | `macos/App/Features/EvidenceGraph/ObservationValueDisplay.swift` `string(for:locale:orders:)` | Name, date, term, text, integer; a boolean reads "" |
| Fake | `macos/App/Platform/FakeStore.swift` `fakeObservation(…row:)` (~3083), the add path's `CatalogObservation(…)` (~3185) | Copies `valueInteger`, not `valueBoolean` (S14-36 left it: no editor) |
| Kit | `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift` | Generic `Value: Hashable`; a selection matching no item selects nothing; no focus binding; each segment is a `PVChip` |
| Tests | `macos/ProvenenciaTests/CitationComposerModelTests.swift` (`makeStore`, `firstRowSaveOnNewCreatesCitationAndExposesRef` ~274, `nameRowUsesDisplayOrderNotTheStoredForm` ~929, `editedRowSaveCallsUpdateOnly`); `macos/ProvenenciaTests/ObservationValueDisplayTests.swift` (`integerUsesValueInteger` ~14) | Integer and text cases are the templates |
| Docs | `docs/interpretation-layer-data-model.md` value-type list; `.agents/skills/add-seeded-vocabulary/SKILL.md` Create sentence | Say "boolean from S14-37" *(S14-36 STEP-10)* |

**Precedent.** S10-32 added the depiction value to the composer: a `supportedValueTypes` entry, a `DepictionValueDraft` on `ObservationRowValues`, an arm in each `CitationObservationValue` switch, a `valueControls` case, and `isDepiction(propertyID:)` for the Tab rule. Copy that shape with the integer arms as the line-by-line template. `DateValueEditorForm`'s kind control (`macos/App/Features/Dates/DateValueEditorForm.swift` ~40) is the precedent for a `PVSegmentedControl` inside a value editor.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `propertymeta.Types` boolean row becomes `ResearcherCreatable: true` (keeps `SingleOnly: true`). No other Go change | S14-36 D-5 left this one flag for this PR; Swift reads creatability from the snapshot (S14-36 D-11), so the picker offers it with no Swift change | — |
| D-2 | Editor: a Yes / No segmented control with nothing chosen at first (user, 2026-11-07) | An unticked checkbox cannot tell "No" from "not answered", and the form already has a Negative checkbox | `PVCheckbox` (unticked saves No); `PVSwitch` (the kit says a saved form uses Checkbox) |
| D-3 | Draft value: `valueBoolean: Bool? = nil` *(new)* on `ObservationRowValues` (after `valueIntegerText`) and on `CitationObservationValue.Fields` (after `valueIntegerText`). nil is "not chosen". `clearValue` sets nil; `isEmptyDraft` requires nil; `loaded(from:)` copies `observation.valueBoolean`; `asCatalogObservation` passes `valueBoolean: draft.valueBoolean` | `Bool?` is the wire's own shape (S14-36 D-17); one field, no "chosen" flag | `Bool` + `hasBoolean` — two fields for one optional |
| D-4 | `CitationObservationValue.Fields.init(_ values: ObservationRowValues)` *(new)* is the one copy of the row's value fields. `fields(from:)` returns `Fields(row.values)`; `ObservationRow.valueFields()` returns `Fields(values)`; `propertyValuesComplete` passes `Fields(values)`. The memberwise init stays for tests | Adding a field to three hand copies is the drift this PR would otherwise repeat; the owning type gets one constructor | Adding `valueBoolean` to each of the three copies |
| D-5 | `CitationObservationValue`: `summary` boolean arm → `fields.valueBoolean.map(PropertyValueType.booleanText) ?? ""`; `isDialogValid` → `fields.valueBoolean != nil`; `apply` → `guard let value = fields.valueBoolean else { return .unsupported }; draft.valueBoolean = value` | The integer arms' shape; the dialog never saves without a choice (Save is disabled), so `.unsupported` is a guard, not a path a researcher sees | A new `Failure` case and string — unreachable copy |
| D-6 | `CitationComposerVocabulary.supportedValueTypes` adds `PropertyValueType.boolean.rawValue`; `isOffered` needs no boolean branch. `func isBoolean(propertyID: String) -> Bool` *(new)* beside `isDepiction` | The composer offers what it can edit; the depiction precedent | Checking `valueType == "boolean"` inline at each call site |
| D-7 | `valueControls` case `PropertyValueType.boolean.rawValue`: `PVSegmentedControl<Bool?>(L10n.CitationComposer.valueLabel, items: [.init(L10n.Properties.valueBooleanYes, value: true, accessibilityIdentifier: "\(identifierPrefix).boolean.yes"), .init(L10n.Properties.valueBooleanNo, value: false, accessibilityIdentifier: "\(identifierPrefix).boolean.no")], selection: values.valueBoolean, size: .sm, select: { values.valueBoolean = $0 })`, not wrapped in `inlineValue` | BV-1, BV-2; the kit control selects nothing when the selection matches no item, so no kit change; the segments reuse S14-36's Yes / No keys, the one place a boolean's words live | New `L10n.CitationComposer` Yes / No keys — a second copy of the words |
| D-8 | Negative: `CitationPropertyFields.showsNegative` *(new, computed)* = `!model.vocabulary.isBoolean(propertyID: values.propertyID) \|\| values.polarity == ObservationPolarity.negative.rawValue`; the `PVCheckbox` is drawn only when true. `applyObservationProperty` sets `values.polarity = positive` when `propertyID` differs from the old one and the new Property is boolean | User, 2026-11-07: No is how the record says it. Resetting on change stops a hidden tick from saving a negative boolean; checking for a real change keeps a re-commit of the same Property from clearing a stored negative | Hiding the checkbox always — a negative row from another catalog would be negative with nothing on screen saying so |
| D-9 | Tab: `CitationPropertyFields.claimTab` passes `propertyValueTakesFocus: false` when `model.vocabulary.isBoolean(propertyID: committed)` (beside the depiction rule), so the combo does not claim Tab and the key-view loop moves to the Yes segment | `PVSegmentedControl` has no focus binding; the depiction rule already uses this path | Adding a focus binding to the kit control — a kit change the brief does not mark |
| D-10 | `ObservationValueDisplay.string`: after the term arm and before the text arm, `if let value = observation.valueBoolean { return PropertyValueType.booleanText(value) }` | One wording (S14-36 D-18); `valueText` is empty for a boolean, so the order only matters for a malformed row, where the typed value wins as the date and term arms already do | Reading `propertyValueType` first — the value field already says it |
| D-11 | FakeStore copies `valueBoolean` from the row in `fakeObservation` and from the draft in the add path | A seeded fake still returns what it was given (working rule "FakeStore is a seeded fake"), so composer tests see the saved value | — |
| D-12 | No new L10n keys. The group's spoken label is `L10n.CitationComposer.valueLabel` ("Value") | BV-7; the Field's label already says it | — |
| D-13 | Evidence graph card and the evidence column show the word in their existing text slot; neither is a chrome change for the canvas. The composer's evidence column is on the S14-D8 board (frame 5) | The card draws a value it already has room for; Conclusion detail did the same in S14-36 | A separate canvas board |

**Rules this plan bends.** None. D-2 and D-8 carry the user's answers of 2026-11-07, recorded in the map's Batch 7 boolean working rule (this planning commit).

## 6. Design

### Codebase fit

Searched: `supportedValueTypes`, `isDepiction`, `valueIntegerText`, `CitationObservationValue.Fields(`, `PVSegmentedControl(`, `negativeBinding`, `propertyValueTakesFocus`, `booleanText`, `valueBooleanYes`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ObservationRowValues.valueBoolean`, `Fields.valueBoolean` | field | `valueIntegerText` | Sibling field (D-3) |
| `CitationObservationValue.Fields.init(_:)` | init | `fields(from:)`, `valueFields()`, `propertyValuesComplete`'s copy | Replaces the three copies (D-4) |
| `CitationComposerVocabulary.isBoolean(propertyID:)` | func | `isDepiction(propertyID:)` | Sibling (D-6) |
| `CitationPropertyFields.showsNegative` | computed | the always-drawn `PVCheckbox` | D-8 |
| Boolean arms in `summary`, `isDialogValid`, `apply`, `valueControls`, `ObservationValueDisplay` | switch arms | integer arms | Extends each switch |
| `CitationComposerModelTests` `canReadPropertyID` and its property row | test fixture | `occupationPropertyID` | Same fixture table |

**Builds on.** S14-36's `propertymeta.Types` (one flag), `PropertyValueType.booleanText` and the Yes / No keys; the composer's per-type switches; `PVSegmentedControl` as `DateValueEditorForm` uses it.
**Deletes.** The three memberwise `Fields` copies (D-4). S14-36's doc wording "boolean from S14-37" in `docs/interpretation-layer-data-model.md` and `.agents/skills/add-seeded-vocabulary/SKILL.md`.
**Next adds.** A later value type with a composer editor is one `supportedValueTypes` entry, one field on `ObservationRowValues` and `Fields` (now one copy), and one arm per switch.

### Reach

| New path or state | Destination | Checked |
| --- | --- | --- |
| A boolean Property in the composer's Property combo box | `CitationPropertyFields.valueControls`, `propertyValuesComplete`, `asDraft` | Each has a boolean arm (D-5, D-7); AC-2.2, AC-2.3 |
| A boolean Observation in the evidence column | `CitationComposerAuthoring.groups`, line row | `summary` gives "Yes" / "No"; `mono` false; negative lines unchanged (AC-2.7, AC-3.3) |
| A boolean Observation on an Evidence graph card and in `EvidenceBridgeEdgeSummary` | `ObservationValueDisplay.string` | D-10, AC-4.1 |
| A negative boolean Observation (catalog made outside the app) | `CitationPropertyFields` | Checkbox shows, checked (AC-3.3). Clearing it hides it on the next draw (D-8): the row is then positive, as intended |
| A boolean Property in Properties create | `PropertiesView` create picker (S14-36 D-11, S14-39) | Offered once D-1 lands; Holds several hidden by S14-39 |

### Scale

None — no list, loop or write path changes; a boolean Observation saves through the existing calls with one more field.

### Edges

- **Assumptions.** `isEmptyDraft` treats every empty field as an untouched draft: `valueBoolean == nil` joins it, so a draft with only a Property picked and no answer is still "touched" by its `propertyID` (as today). `propertyValuesComplete` assumes `isDialogValid` covers every offered type: the boolean arm keeps it total (T-2). `PVSegmentedControl` assumes the selection is one of the item values or none: `nil` selects none (T-4 by hand).
- **Neighbors.** Integer 0/1 is unaffected: integers keep `valueIntegerText`. A text Property keeps its Negative checkbox (AC-3.2). Changing from boolean to another type clears `valueBoolean` (AC-2.5); changing to a boolean resets polarity (AC-2.4); re-committing the same boolean Property keeps a stored negative (T-3).
- **Removed inputs.** None — no rule reads a deleted or excluded input.
- **Real data.** None — no seed drives the rule; no seed Property is boolean.
- **Untrusted size.** None — one optional bit per row.

### Schema

None — no catalog change (S14-36 added the columns).

### Go core

| Package | Change |
| --- | --- |
| `core/propertymeta` | D-1 |

Errors: `properties.invalid` for AC-1.3 (S14-36's `cardinalityAllowed`). No new codes.

### FFI

None — the wire carries the flip in `value_types` (S14-36 D-10).

### macOS

From the S14-D8 inventory.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Value field | `PVField` | `CitationComposerAuthoringBox.swift` `CitationPropertyFields.valueControls` | BV-1; label "Value" unchanged |
| Yes / No control | `PVSegmentedControl<Bool?>` (Ship) | same | D-7; BV-1, BV-2, BV-7 |
| Property observation form | `CitationPropertyFields` (Extend) | same | D-7, D-8 |
| Negative checkbox | `PVCheckbox` (Ship) | same | D-8; BV-4 |
| Dialog Save | existing footer `PVButton` | `CitationComposerAuthoringBox.swift` | Disabled while `propertyValuesComplete` is false; BV-3, BV-5 |
| Saved line | line row (Ship) | `CitationComposerSubjectGroups.swift` | Text from `summary`; BV-6 |

Frames, in words (from the brief; refresh from the agreed board):

1. **New yes/no observation.** Subject John Robins, Property Can read. Under them, the Value field holds a full-width two-segment control, "Yes" then "No", small size, neither selected. No Negative checkbox; the dialog is shorter by its height. Save disabled.
2. **Answer chosen.** As 1, with No selected (the kit's selected chip). Save enabled.
3. **Text Property.** Property Occupation: the text input and the Negative checkbox, as today.
4. **Edit a saved answer.** Opens with Yes selected and Save enabled.
5. **Evidence column.** John: "Occupation Carpenter", "Can read Yes", "Can write Yes"; Mary: "Can read No", "Can write No". Value in the body face, no Badge.
6. **Negative line from elsewhere.** Mary's "Can read": "not", "Yes" and the Negative Badge; its Edit dialog shows the Negative checkbox, checked.
7. **Dark mode** of 1 and 5. 8. **Wide** of 2 and 5: unchanged arrangement.

States: nothing chosen → Save disabled; chosen → enabled. Accessibility: the group is labelled "Value"; each segment reads "Yes" or "No" with the selected trait (`PVChip`); Tab from Property reaches Yes (D-9).

### Strings

None new (D-12). Reused: `L10n.Properties.valueBooleanYes`, `L10n.Properties.valueBooleanNo` *(S14-36)*, `L10n.CitationComposer.valueLabel`.

## 7. Steps

**STEP-1 — Boolean is researcher-creatable**
Files: `core/propertymeta/propertymeta.go`, `core/propertymeta/propertymeta_test.go`, `core/database/properties/properties_test.go`.
Do: D-1. In `TestTypes`, boolean's expected row becomes creatable. In `TestValueTypeRules`, the row "Create boolean → `ErrInvalid`" becomes "Create boolean, cardinality "" → created, single" and a row "Create boolean, cardinality multiple → `ErrInvalid`" is added (T-1).
Check: `go test ./core/propertymeta/` and `CGO_ENABLED=1 go test -tags fts5 ./core/database/properties/ ./api/ffi/...` pass.

**STEP-2 — The composer carries a boolean value**
Files: `macos/App/Features/CitationComposer/CitationObservationRows.swift`, `CitationObservationValue.swift`, `CitationComposerModel+Authoring.swift`, `CitationComposerVocabulary.swift`; `macos/App/Platform/FakeStore.swift`. Skill: `add-swift-test`.
Do: D-3, D-4, D-5, D-6, D-8 (the `applyObservationProperty` half), D-11.
Check: the app builds; `rg -n "valueIntegerText: (row|values)\.valueIntegerText" macos/App` finds one hit, inside `Fields.init(_:)`.

**STEP-3 — The Yes / No control and the Negative checkbox**
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`. Skill: `add-ui-component`.
Do: D-7, D-8 (the view half), D-9.
Check: the app builds with no Swift warning; in the app, a boolean Property shows Yes / No with neither selected and no Negative checkbox (manual check 2).

**STEP-4 — Cards read Yes / No**
Files: `macos/App/Features/EvidenceGraph/ObservationValueDisplay.swift`.
Do: D-10.
Check: T-5 passes.

**STEP-5 — Tests**
Files: `macos/ProvenenciaTests/CitationComposerModelTests.swift`, `macos/ProvenenciaTests/ObservationValueDisplayTests.swift`. Skill: `add-swift-test`.
Do: in `makeStore`, add `canReadPropertyID = "prop-can-read"` (key `can_read`, label "Can read", value type `boolean`) to `propertiesByProject` and bind it to the Person type after Occupation. Add T-2 – T-5.
Check: `scripts/check-pr.sh` passes the Xcode tests.

**STEP-6 — Docs**
Files:
- `docs/interpretation-layer-data-model.md`: the value-type list's boolean entry reads "single only; researcher-creatable" (drop "from S14-37").
- `.agents/skills/add-seeded-vocabulary/SKILL.md`: the Create sentence lists boolean among the creatable types (drop "boolean from S14-37").
- `../deployment-plan.md`: S14-37 row status.
- `../completed.md`, this plan; `../design/S14-D8-composer-boolean-value.md` → `../design/archive/`, `../design/README.md` row moved to Archived.
- `VERSION`, `core/version.go`, `MARKETING_VERSION`: the minor bump.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-37-boolean-value-type-ui.md` reports no errors; `rg -n "from S14-37" docs .agents` finds nothing.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/propertymeta/propertymeta_test.go` `TestTypes`; `core/database/properties/properties_test.go` `TestValueTypeRules` (rows changed / added) | `Rule("boolean")` creatable, single-only; Create boolean "" → single; Create boolean multiple → `ErrInvalid`; T-11 of S14-36 stays green | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-1 |
| T-2 | Swift | `CitationComposerModelTests` `booleanAnswerSavesOnlyOnceChosen` *(new)* | draft Subject + Can read → not complete, `valueBoolean` nil; set false, save → recorded add call, saved row `valueBoolean == false`; `summary` of the saved row "No" | AC-2.1, AC-2.2, AC-2.3, D-3, D-5, D-6, D-11 |
| T-3 | Swift | `CitationComposerModelTests` `switchingToABooleanResetsPolarityAndValue` *(new)* | Occupation + negative → Can read: positive, nil; Can read + true → Occupation: nil, ""; a loaded negative Can read row re-committed to Can read → stays negative | AC-2.4, AC-2.5, D-8 |
| T-4 | Swift | `CitationComposerModelTests` `editingABooleanOpensOnItsAnswer` *(new)* | seeded Can read Yes Observation, Edit → `valueBoolean == true`, not dirty, Save calls no `updateObservation`; seeded negative → `isNegated` | AC-2.6, AC-3.3 |
| T-5 | Swift | `ObservationValueDisplayTests` `booleanReadsYesOrNo` *(new)*; `CitationComposerModelTests` `booleanSummaryReadsYesNoOrNothing` *(new)* | `valueBoolean` true → "Yes", false → "No", nil → ""; `summary` true/false/nil → "Yes"/"No"/"" | AC-2.7, AC-4.1, AC-4.2, D-10 |
| — | manual | §10 steps 2 – 5 | `showsNegative`, Tab order, VoiceOver | AC-3.1, AC-3.2, AC-4.3, D-7, D-9 |

Assert copy with `L10n.string(L10n.Properties.valueBooleanYes)` and `valueBooleanNo`, not literals. Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CitationComposerModelTests` tests that count the Person type's offered Properties, `PropertiesModelTests` (S14-36 T-12 seeds its own rules and stays as written), `TestTypes`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none new)
- [ ] Sweep: `rg -n "from S14-37" docs .agents .cursor/rules` finds nothing; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the S14-D8 board and inventory; no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — minor
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S14-37 row status Landed #N; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc and skill updated per STEP-6
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a new project, with S14-36 and S14-39 landed:

1. Properties → Create: the picker offers boolean. Create "Can read", bind it to Person. Holds offers one value only (S14-39). (AC-1.2)
2. Open a Source's Citation composer, add an observation for a Person, pick Can read. The Value field shows Yes / No, neither chosen; no Negative checkbox; Save disabled. Pick Occupation: the Negative checkbox is back. (AC-2.2, AC-3.1, AC-3.2)
3. From the Property combo box, press Tab: focus lands on Yes. Space chooses it. VoiceOver reads "Value", "Yes", selected. (D-9, BV-7)
4. Save "Can read: No". The evidence column reads "Can read No". Edit it: No is selected. (AC-2.3, AC-2.6)
5. On a second Citation, save "Can read: No" for the same Person. Open the Evidence graph: the card reads "No". Promote; on the Person's Conclusion detail, Can read reads "No", merged. Cite "Yes" on a third: it reads mixed. (AC-4.1, AC-4.3)
6. Widen the window past 1600pt of pane: the dialog and evidence column look the same. (BV-W1)

This is the slice 7e Check: bar outcome 31 on a new project.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S14-36 lands with different names | Redo §4 on landing (plan-pr "Refresh"); Stop condition below |
| A hidden Negative tick saves a negative boolean | D-8 resets polarity on Property change; T-3 |
| Tab skips the segmented control | D-9; manual check 3; Stop condition below |

**Stop and report if:**

- At your base, any *(S14-36)* name in §4 does not exist or has another shape (`propertymeta.Types`, `PropertyValueType.booleanText`, `CatalogObservation.valueBoolean`, `CatalogObservationDraft.valueBoolean`, `L10n.Properties.valueBooleanYes`).
- S14-39 has not landed (the create form would still offer Holds several for boolean).
- Tab from the Property combo box does not reach the Yes segment (it lands on the footer or skips the control). The fix is a focus binding on `PVSegmentedControl`, a kit change for the user to decide.
- `PVSegmentedControl` with a `nil` selection draws a selected segment.
- The agreed S14-D8 board differs from §6 macOS frames in anything but wording.

## 12. Open questions

None from the code. The plan stays Blocked until S14-36 and S14-39 land and the S14-D8 board is agreed; then §4 and §6 macOS are refreshed and `review-plan` runs again.

## 13. PR

**Title:** S14-37 — Boolean value type (UI)
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-37-boolean-value-type-ui.md`.
