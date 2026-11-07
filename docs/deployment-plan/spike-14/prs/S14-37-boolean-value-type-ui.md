# S14-37 — Boolean value type (UI)

**Status:** Blocked — on S14-39
**Spike / slice:** Spike 14 · Slice 7e — Boolean value type
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S14-D8](../design/S14-D8-composer-boolean-value.md), board agreed 2026-11-07
**Board:** S14-D8 — the *Citation Composer* project, now a [migrated design on claude.ai](https://claude.ai/artifact/3GSFQZjqdQjL8v2rpgMiDc), page *Citation Composer.dc.html*, section "S14-D8 · Yes / no value", frames 1, 2, 3, 4, 5, 7a, 7b, 8a, 8b. Change from the brief: frame 6 (a negative yes/no) removed; a yes/no is never negative (user, 2026-11-07). The implementer may not be able to open it: **Design → macOS** describes each frame in words.
**Depends on:** S14-36 (landed [#733](https://github.com/mendahu/provenencia/pull/733)), S14-39 (scoped), S14-D8
**Blocks:** —
**Branch:** `s14-37-boolean-value-type-ui`
**Version bump:** minor — researchers can create and cite a new kind of Property, a feature ([`docs/versioning.md`](../../../versioning.md)); the next minor after `VERSION` at merge (0.3.8 → 0.4.0 at planning)
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (the `propertymeta.Types` table), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md) (`observations`), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** working rule "(Batch 7) Boolean is a first-class value type in two PRs" (Go owns the value-type rules; the composer's Yes / No and no Negative); working rule "(Batch 5) FakeStore is a seeded fake, not a second engine"; bar outcome 31; dogfood note [`ux/2026-10-25-boolean-property-value-type.md`](../ux/2026-10-25-boolean-property-value-type.md); S14-36 plan and its **Landed** note [`S14-36-boolean-value-type-engine.md`](S14-36-boolean-value-type-engine.md)
**Planned:** 2026-11-07 against `main` @ `aaf3b15`
**Review:** [`S14-37-boolean-value-type-ui.review.md`](S14-37-boolean-value-type-ui.review.md) — Fits with changes — Ready with changes, round 1 answered; To user answered (M-1, M-2, M-4); round 2 runs once the board is agreed

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

A researcher can make a yes/no Property ("Can read", "Veteran") on the Properties page and cite it from a census column in the Citation composer. The composer's Value field becomes a Yes / No segmented control with nothing chosen at first, so a No is always one the researcher picked. A yes/no Property cannot be negated: No is how the record says it. Go owns that rule beside the other value-type rules and refuses a negative boolean; the composer reads it and draws no Negative checkbox. The saved answer reads "Yes" or "No" in the composer's evidence column and on Evidence graph cards; Conclusion detail already reads it (S14-36). This completes bar outcome 31 and the dogfood note of 2026-10-25.

The PR also makes the composer keep its value fields in one place (the value part of a row becomes `CitationObservationValue.Fields`), deletes the composer's dead row-mutation paths, and gives the kit's segmented control the focus handle every other kit field has.

**Why now.** S14-36 stores, reconciles and sends booleans; S14-39 makes the Properties page honour the single-only rule. Boolean stays uncreatable until the composer can take the value: this PR flips it.

**What stays.** Every other value type's editor and its Negative checkbox. The observation dialog's layout and the evidence column's line layout. The Properties page (S14-39). Conclusion detail's value text (S14-36). A negative Observation of any other type.

## 2. User stories

**US-1.** As the engine, I own whether a value type can be created by a researcher and whether it can be negated, and I refuse what my rules forbid.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `propertymeta.Rule("boolean")` | `ResearcherCreatable` true, `SingleOnly` true, `Negatable` false; every other row `Negatable` true |
| AC-1.2 | `properties.Create(tx, userID, "Can read", "boolean", "", "")` | Created, `cardinality = single` |
| AC-1.3 | The same with cardinality `"multiple"` | `ErrInvalid`; no row |
| AC-1.4 | `GetPropertiesWorkspace` | the `boolean` row of `value_types` has `researcher_creatable` true and `negatable` false; `text` has `negatable` true |
| AC-1.5 | `AddToCitation` with a boolean Observation, polarity `negative`, value true | `ErrInvalid`; no row |
| AC-1.6 | `Update` of a stored negative boolean Observation (written by raw SQL) keeping polarity `negative` | `ErrInvalid`; row unchanged. The same update with polarity `positive` succeeds |
| AC-1.7 | `AddToCitation` with a text Observation, polarity `negative` | Stored, as today |

**US-2.** As a researcher, when I cite a yes/no Property in the composer, I pick Yes or No, and nothing is saved until I do.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The Property combo box for a Person Subject whose type binds "Can read" (boolean) | "Can read" is offered (`supportedValueTypes`) |
| AC-2.2 | A new property draft: Subject, then Property "Can read" | `values.valueFields.valueBoolean == nil`; `authoringCanConfirm` false |
| AC-2.3 | AC-2.2, then `valueBoolean = false`, then Save | `recordedCalls` has the `addObservationsToCitation` (or `createCitationWithObservations`) prefix; `store.observationsByGraph[graphID]` holds one Can read row with `valueBoolean == false`; the saved composer row's summary is "No" |
| AC-2.4 | A draft with Occupation and Negative ticked, then Property changed to "Can read", Yes chosen, Save | `valueBoolean` was nil after the change; the saved Observation's polarity is `positive` |
| AC-2.5 | A draft with "Can read" and `valueBoolean = true`, then Property changed to Occupation | `valueBoolean == nil`, `valueText == ""` |
| AC-2.6 | A draft with "Can read" and `valueBoolean = true`, then "Can read" committed again in the Property combo box | `valueBoolean == true` (nothing changes) |
| AC-2.7 | Edit on a saved "Can read: Yes" Observation | The edit's `valueBoolean == true`; the row's `state == .saved` (not dirty); Save of the unchanged edit records no `updateObservation` |
| AC-2.8 | `CitationObservationValue.summary(valueType: "boolean", …)` with `true`, `false`, `nil` | "Yes", "No", "" |

**US-3.** As a researcher, a yes/no Property never shows a Negative checkbox and is never saved negative (user, 2026-11-07).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `CitationComposerVocabulary.offersNegative(propertyID: canRead)` with rules seeded | false |
| AC-3.2 | The same for Occupation | true |
| AC-3.3 | `ObservationRow.asDraft` of a Can read row whose `polarity` is `negative` (a row edited by hand in the database) | the draft's polarity is `positive` |
| AC-3.4 | The same for "Can read" with no rules loaded (`valueTypeRules` empty) | true (Go decides, as before the first load) |

**US-4.** As a researcher, a saved yes/no answer reads "Yes" or "No" wherever an Observation's value is shown.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `ObservationValueDisplay.string(for:)` an Observation with `propertyValueType "boolean"`, `valueBoolean true` | "Yes"; with `false` → "No" |
| AC-4.2 | The same with `valueBoolean nil` and `valueText ""` | "" (as today) |
| AC-4.3 | A Person with two Citations both saying "Can read: Yes", on Conclusion detail | "Yes", merged (S14-36's text; checked by hand here) |

**US-5.** As a keyboard user, Tab from the Property combo box lands on the Yes segment.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `PVSegmentedControl` given `focus:` and the host sets that focus | The first enabled segment takes keyboard focus |
| AC-5.2 | In the dialog, Property "Can read" committed with Tab | Focus is on Yes; Space chooses it (manual check 3) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `propertymeta.TypeRule.Negatable` *(new)*; boolean creatable and not negatable | STEP-1, AC-1.1, AC-1.2, AC-1.3 |
| `observations` refuses a negative Observation of a non-negatable type | STEP-1, AC-1.5, AC-1.6, AC-1.7 |
| Proto `PropertyValueTypeRule.negatable`; handler; Swift rule mapping | STEP-2, AC-1.4 |
| The composer row's value part is one `CitationObservationValue.Fields` | STEP-3, all US-2 tests |
| Dead composer row-mutation paths deleted | STEP-3 |
| Composer draft value `valueBoolean`, summary, validity, save; boolean offered; Property change resets polarity and value only when the Property changes | STEP-4, AC-2.1 – AC-2.8 |
| `CitationComposerVocabulary.offersNegative` from Go's rule (BV-4) | STEP-4, AC-3.1 – AC-3.4 |
| `PVSegmentedControl` focus handle | STEP-5, AC-5.1 |
| Yes / No control, Negative checkbox, Tab (BV-1 – BV-5, BV-7) | STEP-6, AC-2.2, AC-5.2 |
| Saved line text (BV-6) | STEP-4 (summary), AC-2.3, AC-2.8 |
| Wide (BV-W1) | manual check 6 |
| Evidence graph card text | STEP-7, AC-4.1, AC-4.2 |
| Docs, version bump | STEP-8 |

| Out | Home |
| --- | --- |
| Properties page: Holds several hidden for a single-only type; boolean pill colour; the picker's label for boolean | S14-39 (brief S14-D9) |
| A "not answered" or "unknown" value | not planned (a blank column is not cited; brief non-goal) |
| Keyboard shortcuts for Yes and No | not planned (brief §2.1) |
| Using the new focus handle in `DateValueEditorForm` and `PropertiesView` | not planned (their Tab order works today) |
| Folding a stored negative boolean into its opposite value | not planned (S14-36 D-9; the engine now refuses new ones) |
| `FakeStore.appendFakeObservations` built on `fakeObservation`, and its `displayText` synthesis | [structural debt](../../../ideas/structural-debt.md) (row added in STEP-8) |
| Two text paths for a saved Observation (`CitationObservationValue.summary` and `ObservationValueDisplay`) | [structural debt](../../../ideas/structural-debt.md) (S14-36's display row gains `summary`, STEP-8) |
| Omnibar indexing of boolean values | not planned (S14-36 Out) |

## 4. Current state

Verified at `aaf3b15`.

| Area | Where | Today |
| --- | --- | --- |
| Value-type rules (Go) | `core/propertymeta/propertymeta.go` `TypeRule` (`ValueType`, `ResearcherCreatable`, `SingleOnly`), `Types`, `Rule`, `ValueTypes` | boolean row `{ValueType: ValueTypeBoolean, SingleOnly: true}` |
| Rule tests (Go) | `core/propertymeta/propertymeta_test.go` `TestTypes` (`flags` table pins every row); `core/database/properties/properties_test.go` `TestValueTypeRules` (row "create boolean is not researcher-creatable yet") | Both assert boolean not creatable |
| Observation polarity | `core/database/observations/observations.go` `insertOne` (~273: polarity check, ~300 `getPropertyTx`, ~314 `resolveValue`); `core/database/observations/update.go` `updateOne` (~27 polarity, ~49 `getPropertyTx`) | Any type takes `negative` |
| Observation tests | `core/database/observations/observations_test.go` `TestBooleanObservationValue` (~1530, table with `booleanFixture`) | No polarity rows |
| Proto / handler | `api/proto/engine.proto` `PropertyValueTypeRule` (~2325, fields 1–3); `api/ffi/handlers/subject_defs.go` `GetPropertiesWorkspace` (~268); `api/ffi/handlers/subject_defs_test.go` `TestGetPropertiesWorkspaceValueTypes` (~266, compares field for field) | No `negatable` |
| Swift rule | `macos/App/Platform/GenealogyStore.swift` `CatalogPropertyValueTypeRule` (~944); `macos/App/Platform/GoStore.swift` `getPropertiesWorkspace` (~1940); `macos/ProvenenciaTests/PropertiesModelTests.swift` `goValueTypeRules` (~257) | Three fields; two construction sites |
| Swift snapshot | `macos/App/Features/Properties/PropertiesSnapshot.swift` `valueTypeRules`, `creatableValueTypes`; `valueTypeRule(_:)` *(S14-39)* | One lookup by type (from S14-39) |
| Swift value kinds | `macos/App/Platform/InterpretationValueKinds.swift` `PropertyValueType.boolean`, `booleanText(_:)`; `macos/App/Platform/L10n.swift` `L10n.Properties.valueBooleanYes`, `valueBooleanNo` | "Yes" / "No" |
| Swift models | `GenealogyStore.swift` `CatalogObservation.valueBoolean`, `CatalogObservationDraft.valueBoolean` (`Bool? = nil`); `GoStore` maps both ways; `FakeStore.fakeObservation` (~3123) and `appendFakeObservations` (~3198) copy it | Wire complete |
| Composer offered types | `macos/App/Features/CitationComposer/CitationComposerVocabulary.swift` `supportedValueTypes`, `isOffered`, `isDepiction(propertyID:)`, `make(fields:snapshot:rules:termsByPropertyID:)` | No boolean; no value-type rules kept |
| Composer row values | `macos/App/Features/CitationComposer/CitationObservationRows.swift` `ObservationRowValues` (`subjectID`, `propertyID`, `polarity`, then seven value fields; `clearValue`), `ObservationRow` (seven value-field forwarding accessors ~60–93, `isEmptyDraft`, `loaded(from:)`, `valueFields()`, `asDraft`, `asCatalogObservation`) | The same seven fields as `Fields`, kept by hand |
| Composer value logic | `macos/App/Features/CitationComposer/CitationObservationValue.swift` `Fields`, `fields(from:)` (no caller), `summary`, `isDialogValid`, `apply` | `switch valueType` per type |
| Model copy of `Fields` | `macos/App/Features/CitationComposer/CitationComposerModel+Authoring.swift` `propertyValuesComplete` (~124), `authoringCanConfirm`, `applyObservationProperty(_:values:)` (~233) | Hand copy of seven fields; Property change always clears the value |
| Dead row paths | `CitationObservationRows.swift` `applyProperty(rowID:propertyID:)`, `applyInteger(rowID:text:)`, `applyTerm(rowID:termID:)`, `togglePolarity(rowID:)`, `ObservationRow.polarityMenuTitle`; `L10n.CitationComposer.negateObservation`, `affirmObservation` | No caller in `macos/App`; `polarityMenuTitle` is called only from its own test; `applyText` only from two row-edit tests (~331, ~348) |
| Dialog form | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationPropertyFields` (`body`, `valueControls` ~366, `claimTab` ~448, bindings ~495–530) | Negative `PVCheckbox` always drawn; each value editor takes `PVFieldFocus(focus, equals: .value)` |
| Tab order | `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` `CitationAuthoringDraftAfterCommit.propertyValueTakesFocus` | False only when the value has no control (a blocked depiction) |
| Saved lines | `CitationComposerAuthoring.swift` `groups(observations:…)` (~330); `CitationComposerSubjectGroups.swift` line row (~150–190) | Text from `summary`; "not" word and Negative Badge on a negative line |
| Kit | `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift`; `macos/App/DesignSystem/Components/Input/PVFieldFocus.swift`; `macos/App/DesignSystem/README.md` SegmentedControl row (~173) | No focus handle; a selection matching no item selects nothing; each segment is a `PVChip` (draws keyboard focus) |
| Card text | `macos/App/Features/EvidenceGraph/ObservationValueDisplay.swift` `string(for:locale:orders:)` | A boolean reads "" |
| Tests | `macos/ProvenenciaTests/CitationComposerModelTests.swift` (`makeStore` ~29 builds `propertiesByProject` as a list and binds by index `![0]`…`![7]`; `occupationObservation` ~1375; `polarityMenuTitleFollowsPolarity` ~920); `macos/ProvenenciaTests/CitationComposerDepictionTests.swift`; `macos/ProvenenciaTests/ObservationValueDisplayTests.swift` | Integer and text cases are the templates |
| Docs | `docs/interpretation-layer-data-model.md`; `.agents/skills/add-seeded-vocabulary/SKILL.md`; `docs/deployment-plan/spike-14/deployment-plan.md` bar outcome 31 | Say boolean is creatable "from S14-37" |

**Precedent.** S14-36 ([#733](https://github.com/mendahu/provenencia/pull/733)) added `SingleOnly` end to end: a `TypeRule` field, a refusal in the domain package, a proto field, a Swift field mapped in `GoStore`: copy that path for `Negatable`. S10-32 added the depiction value to the composer (`supportedValueTypes`, a draft field, one arm per `CitationObservationValue` switch, a `valueControls` case, `isDepiction`): copy its shape with the integer arms as the line-by-line template. `PVSelect`'s `focus: PVFieldFocus? = nil` (`macos/App/DesignSystem/Components/Select/PVSelect.swift` ~109) is the precedent for the kit focus handle.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `propertymeta.TypeRule` gains `Negatable bool` *(new)*, declared after `SingleOnly`, comment "Negatable is whether an Observation of the type may be negative: a yes/no says No with its value". Every row of `Types` sets `Negatable: true` except boolean. The boolean row becomes `{ValueType: ValueTypeBoolean, ResearcherCreatable: true, SingleOnly: true}` | User, 2026-11-07: Go owns the rule, beside creatable and single-only; S14-36 D-5 left the creatable flag for this PR | A Swift-only rule — the user chose Go; a negative-sense `NoNegative` field — the other two fields read positive |
| D-2 | `observations.negationAllowed(valueType, polarity string) bool` *(new, observations.go)*: true unless `polarity == PolarityNegative` and `propertymeta.Rule(valueType)` is found with `Negatable` false. `insertOne` and `updateOne` call it right after `getPropertyTx` and return `ErrInvalid` when false | A rule Go owns is enforced by Go; the composer never sends one (D-9), so the refusal only meets crafted calls | Letting the engine accept it and only hiding the checkbox — the rule would be advice |
| D-3 | Proto: `PropertyValueTypeRule` `bool negatable = 4;`. `GetPropertiesWorkspace` sets `Negatable: r.Negatable`. Swift: `CatalogPropertyValueTypeRule.negatable: Bool` *(new, no default)*; `GoStore` maps `$0.negatable`; `goValueTypeRules` passes `negatable: true`. | The S14-36 path for `singleOnly`; no default, so a missing mapping does not compile | A default of `true` — it hides a missing mapping |
| D-4 | The composer row's value part is one struct (user, 2026-11-07): `ObservationRowValues` keeps `subjectID`, `propertyID`, `polarity` and gains `var valueFields = CitationObservationValue.Fields()` in place of its seven value fields. `clearValue()` becomes `valueFields = .init()`; `isEmptyDraft` tests `values.valueFields == .init()`; `ObservationRow`'s seven value-field forwarding accessors and `valueFields()` are deleted (callers read `row.values.valueFields`); `propertyValuesComplete` passes `values.valueFields`; `CitationObservationValue.fields(from:)` is deleted. Each caller is renamed mechanically (`values.valueText` → `values.valueFields.valueText`); no behaviour changes in this step | One declaration per value field; the "empty" and "clear" rules become one `==` and one assignment instead of hand lists (review M-1) | Adding `valueBoolean` to seven places; one `Fields.init(_:)` replacing three copies — still five places |
| D-5 | `CitationObservationValue.Fields.valueBoolean: Bool? = nil` *(new)*, after `valueIntegerText`; nil is "not chosen". `loaded(from:)` sets it from `observation.valueBoolean`; `asCatalogObservation` passes `valueBoolean: draft.valueBoolean`. `summary` → `fields.valueBoolean.map(PropertyValueType.booleanText) ?? ""`; `isDialogValid` → `fields.valueBoolean != nil`; `apply` → `guard let value = fields.valueBoolean else { return .unsupported }; draft.valueBoolean = value` | The wire's own shape (S14-36 D-17); the integer arms' shape. Save is disabled without a choice, so `.unsupported` is a guard no researcher meets | `Bool` + a chosen flag; a new `Failure` case and string — unreachable copy |
| D-6 | `CitationComposerVocabulary`: `supportedValueTypes` adds `PropertyValueType.boolean.rawValue`; `func valueType(propertyID: String) -> PropertyValueType?` *(new)*; `var fields: PropertiesSnapshot` kept from `make(fields:…)` *(new stored property; `.empty` holds `PropertiesSnapshot.empty`)*; `func offersNegative(propertyID: String) -> Bool` *(new)* = `fields.valueTypeRule(valueType)?.negatable ?? true`, through S14-39's one lookup `PropertiesSnapshot.valueTypeRule(_:)` (S14-39 review M-1). `ObservationRow.asDraft(property:)` becomes `asDraft(property:vocabulary:)` (and `asCatalogObservation(property:)` likewise; `performCommit` passes `context.vocabulary`, the new-draft save passes `vocabulary`) and writes `positive` when `!vocabulary.offersNegative(propertyID:)`, whatever the row's `polarity`: the one place the composer applies the rule to what it saves (user, 2026-11-07: a yes/no is never negative, no exception). `isDepiction` stays | The rule's one Swift reader; with no rules loaded Go decides (AC-3.4). A typed lookup instead of a second `isX` (review M-6) | `isBoolean(propertyID:)` — one predicate per type |
| D-7 | `applyObservationProperty(_:values:)`: returns without change when `propertyID == values.propertyID`; otherwise sets the Property and clears the value, as today. Polarity is left alone: `asDraft` applies the rule (D-6) | A re-commit of the same Property (Shift-Tab back, Tab on) no longer wipes an answer, for every type | Resetting polarity here too — a second place for the rule |
| D-8 | Dead paths deleted: `ObservationRows.applyProperty(rowID:propertyID:)`, `applyInteger(rowID:text:)`, `applyTerm(rowID:termID:)`, `togglePolarity(rowID:)`, `ObservationRow.polarityMenuTitle`, `L10n.CitationComposer.negateObservation` and `affirmObservation` (and their `Localizable.xcstrings` entries), test `polarityMenuTitleFollowsPolarity`. `applyText(rowID:text:)` stays: `editedRowSaveCallsUpdateOnly` and `revertDraftRemovesAndRevertEditedRestores` drive a row edit through it | Twins of the paths D-4 and D-7 change; a revived one would skip the polarity rule (review M-3) | Leaving them — a second negation path |
| D-9 | Dialog: `valueControls` case `PropertyValueType.boolean.rawValue` → `PVSegmentedControl<Bool?>(L10n.CitationComposer.valueLabel, items: [.init(L10n.Properties.valueBooleanYes, value: true, accessibilityIdentifier: "\(identifierPrefix).boolean.yes"), .init(L10n.Properties.valueBooleanNo, value: false, accessibilityIdentifier: "\(identifierPrefix).boolean.no")], selection: values.valueFields.valueBoolean, size: .sm, focus: PVFieldFocus(focus, equals: .value), select: { values.valueFields.valueBoolean = $0 })`, not wrapped in `inlineValue`, with no extra frame (the segments expand). The Negative `PVCheckbox` sits in `if model.vocabulary.offersNegative(propertyID: values.propertyID)`, so the `VStack` closes up. `claimTab` needs no change: the Yes segment takes `.value` | BV-1 – BV-5, BV-7; the segments reuse S14-36's Yes / No keys; one focus mechanism for every value editor | A `PVCheckbox` (user chose segmented); the key-view loop (user, 2026-11-07) |
| D-10 | Kit: `PVSegmentedControl` gains `focus: PVFieldFocus? = nil` *(new init parameter, after `size`)*, applied with `focus.apply(AnyView(chip))` to the first item that is not disabled; the README SegmentedControl row names it | Every other kit field takes the handle (`PVInput`, `PVTextArea`, `PVComboBox`, `PVSelect`); the user chose the kit change (2026-11-07) | A composer-local wrapper — a second copy of the control |
| D-11 | `ObservationValueDisplay.string`: after the term arm and before the text arm, `if let value = observation.valueBoolean { return PropertyValueType.booleanText(value) }` | One wording (S14-36 D-18) | Reading `propertyValueType` first — the value field already says it |
| D-12 | No new L10n keys. Two removed (D-8) | — | New Yes / No keys under `CitationComposer` — a second copy |
| D-13 | The Evidence graph card shows the word in its existing text slot; no canvas board. The composer's evidence column is on S14-D8 (frame 5) | The card draws a value it already has room for, as Conclusion detail did in S14-36 | A separate canvas board |

**Rules this plan bends.** None. D-1, D-2, D-4, D-9 and D-10 carry the user's answers of 2026-11-07. Planning added "negatable" to the map's Batch 7 boolean working rule beside creatable and single-only.

## 6. Design

### Codebase fit

Searched: `TypeRule`, `SingleOnly`, `single_only`, `singleOnly`, `supportedValueTypes`, `isDepiction`, `valueIntegerText`, `CitationObservationValue.Fields(`, `PVSegmentedControl(`, `focus: PVFieldFocus?`, `negativeBinding`, `applyProperty(`, `togglePolarity`, `booleanText`, `valueBooleanYes`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `TypeRule.Negatable`, proto `negatable`, `CatalogPropertyValueTypeRule.negatable` | field | `SingleOnly` / `single_only` / `singleOnly` | Extends the same table and wire (D-1, D-3) |
| `observations.negationAllowed` | func | `properties.cardinalityAllowed` (S14-36) | The same shape in the package that writes Observations (D-2) |
| `ObservationRowValues.valueFields` | field | the seven value fields it replaces | D-4 |
| `Fields.valueBoolean` | field | `valueIntegerText` | D-5 |
| `CitationComposerVocabulary.valueType(propertyID:)`, `fields`, `offersNegative` | func / field | `isDepiction(propertyID:)` | D-6 |
| `PVSegmentedControl` `focus:` | init parameter | `PVSelect` `focus:` | D-10 |
| Boolean arms in `summary`, `isDialogValid`, `apply`, `valueControls`, `ObservationValueDisplay` | switch arms | integer arms | Extends each switch |
| `CitationComposerModelTests` `canReadPropertyID`, `canReadObservation(id:ref:citationID:value:)` | test fixture | `occupationPropertyID`, `occupationObservation` | Same fixture table |

**Builds on.** S14-36's value-type table and wire; the composer's per-type switches; the kit's `PVFieldFocus`.
**Deletes.** The seven value fields of `ObservationRowValues` and `ObservationRow`'s seven forwarding accessors and `valueFields()` (D-4); `CitationObservationValue.fields(from:)`; the hand list in `propertyValuesComplete`; the dead row paths, two strings and one test of D-8; the docs' "from S14-37".
**Next adds.** S14-39 reads `singleOnly` from the same snapshot. A later value type with a composer editor is one `supportedValueTypes` entry, one field on `Fields`, one arm per switch, and one `Types` row that says whether it is negatable.

### Reach

| New path or state | Destination | Checked |
| --- | --- | --- |
| A boolean Property in the composer's Property combo box | `valueControls`, `propertyValuesComplete`, `asDraft` | A boolean arm in each (D-5, D-9); AC-2.2, AC-2.3 |
| A boolean Observation in the evidence column | `groups`, line row | `summary` gives "Yes" / "No"; `mono` false (AC-2.8) |
| A boolean Observation on an Evidence graph card and in `EvidenceBridgeEdgeSummary` | `ObservationValueDisplay.string` | D-11, AC-4.1 |
| A negative boolean row edited into the database by hand (no app path writes one) | evidence line; composer dialog; `observations.updateOne` | The line shows its stored "not" and Badge, as any stored negative; the dialog shows no checkbox; Save writes it positive (AC-3.3), which Go accepts (AC-1.6) |
| A boolean Property in Properties create | `PropertiesView` create picker | Offered once D-1 lands; Holds several hidden by S14-39 |
| A re-commit of the same Property in any dialog | `applyObservationProperty` | Value and polarity kept (AC-2.6) |

### Scale

None — no list, loop or write path changes; one comparison per Observation write, one bit per value-type row.

### Edges

- **Assumptions.** `isEmptyDraft` treats an untouched draft as one with no Property and empty values: `valueFields == .init()` includes `valueBoolean == nil` (T-6). `propertyValuesComplete` assumes `isDialogValid` covers every offered type: the boolean arm keeps it total (T-5). `PVSegmentedControl` assumes the selection is one item's value or none: `nil` selects none (manual check 2). `negationAllowed` assumes every stored value type has a rule: the `properties` CHECK and `TestTypes` hold that.
- **Neighbors.** A negative text, date, name, term, subject or depiction Observation is stored as today (AC-1.7, T-2). Integer 0/1 stays an integer. Changing from boolean to another type clears the answer (AC-2.5); a boolean always saves positive, whatever a hidden tick left (AC-2.4, AC-3.3); the same Property keeps its answer (AC-2.6).
- **Removed inputs.** None — no rule reads a deleted or excluded input.
- **Real data.** `propertymeta.Types` is the whole table: `TestTypes` pins all eight rows' three flags.
- **Untrusted size.** None — one optional bit per row.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/propertymeta` | D-1 |
| `core/database/observations` | D-2 |

Errors: `observations.invalid` (AC-1.5, AC-1.6), `properties.invalid` (AC-1.3). No new codes.

### FFI

`api/proto/engine.proto` `PropertyValueTypeRule.negatable = 4`; `api/ffi/handlers/subject_defs.go` sets it. Regenerate with `./scripts/generate-proto.sh` (needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`); commit `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`.

### macOS

From the S14-D8 inventory.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Value field | `PVField` | `CitationComposerAuthoringBox.swift` `CitationPropertyFields.valueControls` | BV-1 |
| Yes / No control | `PVSegmentedControl<Bool?>` (Extend: `focus:`) | same; kit in `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift` | D-9, D-10 |
| Property observation form | `CitationPropertyFields` (Extend) | same | D-9 |
| Negative checkbox | `PVCheckbox` (Ship) | same | D-6, D-9; BV-4 |
| Dialog Save | existing footer `PVButton` | `CitationComposerAuthoringBox.swift` | Disabled while `authoringCanConfirm` is false; BV-3, BV-5 |
| Saved line | line row (Ship) | `CitationComposerSubjectGroups.swift` | Text from `summary`; BV-6 |

Frames, in words (from the agreed board; the board's dialog is the existing observation dialog: title "New observation" or "Edit observation", the Citation picker and its transcription, Property | Bridge, then the fields; footer Discard or Cancel, and Save observation):

1. **New yes/no observation.** Subject John Robins, Property Can read. Under them, the Value field holds a two-segment control, "Yes" then "No", small, as wide as the field, neither selected. No Negative checkbox; the dialog is shorter by its height. Save observation disabled. Note on the board: Tab from Property lands on Yes, then No, then the footer; Space chooses; VoiceOver reads "Value", "Yes" / "No", selected; only the kit focus ring shows.
2. **Answer chosen.** As 1, with No selected (the segmented control's selected chip). Save observation enabled. Clicking Yes moves the selection; nothing returns to none.
3. **Text Property** (half size, beside 1). Occupation: the text input in Value and the Negative checkbox under it, as today.
4. **Edit a saved answer.** Edit on John's Can write Yes: Yes selected, Save observation enabled.
5. **Evidence column.** John Robins: Occupation Carpenter, Can read Yes, Can write Yes. Mary Robins: Can read No, Can write No. Yes / No in the body face, like Carpenter; no Badge, no glyph.
7a, 7b. **Dark** of 1 and 5. 8a, 8b. **Wide** of 2 and 5: the three-pane composer as today; the dialog and the observations column keep their widths, the value does not stretch.

States: nothing chosen → Save disabled; chosen → enabled; a refused save shows the existing save error. Accessibility: the group is labelled "Value"; each segment reads "Yes" or "No" with the selected trait; Tab from Property lands on Yes (D-9, D-10).

### Strings

None new. Removed: `citationComposer.negateObservation`, `citationComposer.affirmObservation` (D-8). Reused: `L10n.Properties.valueBooleanYes`, `valueBooleanNo`, `L10n.CitationComposer.valueLabel`.

## 7. Steps

**STEP-1 — Go: boolean is creatable and cannot be negated**
Files: `core/propertymeta/propertymeta.go`, `propertymeta_test.go`; `core/database/observations/observations.go`, `update.go`, `observations_test.go`; `core/database/properties/properties_test.go`. Skill: `add-match-rule`, `add-catalog-query`.
Do: D-1, D-2. `TestTypes`' `flags` table gains a `negatable` column (true for every row but boolean) and boolean's creatable becomes true. In `TestValueTypeRules`, "create boolean is not researcher-creatable yet" becomes "create boolean empty cardinality is single" (created, single) and a row "create boolean multiple is refused" is added. Add T-2.
Check: `go test ./core/propertymeta/` and `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` pass.

**STEP-2 — The rule reaches the app**
Files: `api/proto/engine.proto`, `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/subject_defs.go`, `subject_defs_test.go`; `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`; `macos/ProvenenciaTests/PropertiesModelTests.swift`. Skill: `add-ffi-handler`.
Do: D-3; run `./scripts/generate-proto.sh`. `TestGetPropertiesWorkspaceValueTypes` compares `GetNegatable()` too and asserts the boolean row by name (T-1).
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; the app builds.

**STEP-3 — One copy of the composer's value fields; dead paths go**
Files: `macos/App/Features/CitationComposer/CitationObservationRows.swift`, `CitationObservationValue.swift`, `CitationComposerModel.swift`, `CitationComposerModel+Authoring.swift`, `CitationComposerModel+Depiction.swift`, `CitationComposerAuthoring.swift`, `CitationComposerAuthoringBox.swift`; `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`; `macos/ProvenenciaTests/CitationComposerModelTests.swift`, `CitationComposerDepictionTests.swift`. Skill: `add-localized-string` (removal).
Do: D-4 and D-8 only; no boolean yet. Rename by compiling: every error names a site.
Check: Xcode tests pass unchanged except the deleted test; `rg -n "values\.(valueText|valueIntegerText|valueTermID|valueSubjectID|dateDraft|nameDraft|depictionDraft)\b" macos/App macos/ProvenenciaTests` finds nothing; `rg -n "togglePolarity|polarityMenuTitle|negateObservation|affirmObservation|fields\(from:" macos` finds nothing.

**STEP-4 — The composer carries a boolean value**
Files: `CitationObservationValue.swift`, `CitationObservationRows.swift`, `CitationComposerModel+Authoring.swift`, `CitationComposerVocabulary.swift`; `macos/ProvenenciaTests/CitationComposerModelTests.swift`. Skill: `add-swift-test`.
Do: D-5, D-6, D-7. In `makeStore`, append the Can read row **last** in `propertiesByProject` (`canReadPropertyID = "prop-can-read"`, key `can_read`, label "Can read", value type `boolean`, index 8) and append `field(store.propertiesByProject[projectDir]![8], 2)` to the Person binding; seed `store.valueTypeRules` with `text` and `boolean` rows (boolean `negatable: false`). Add `canReadObservation(id:ref:citationID:value:)` beside `occupationObservation` (`valueBoolean:` directly after `valueInteger:`, `propertyValueType: "boolean"`, `valueText: ""`). Add T-3 – T-6.
Check: Xcode tests pass.

**STEP-5 — The segmented control takes a focus handle**
Files: `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift`, `macos/App/DesignSystem/README.md`. Skill: `add-ui-component`.
Do: D-10.
Check: the app builds; every existing `PVSegmentedControl` call site compiles unchanged.

**STEP-6 — The Yes / No control and the Negative checkbox**
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`. Skill: `add-ui-component`.
Do: D-9. Then run manual checks 2 and 3 before STEP-7.
Check: the app builds with no Swift warning; manual checks 2 and 3 pass.

**STEP-7 — Cards read Yes / No**
Files: `macos/App/Features/EvidenceGraph/ObservationValueDisplay.swift`, `macos/ProvenenciaTests/ObservationValueDisplayTests.swift`.
Do: D-11. Add T-7.
Check: T-7 passes.

**STEP-8 — Docs and version**
Files:
- `docs/interpretation-layer-data-model.md` (~630, ~632): the `Types` sentence reads "with which a researcher may create, which are single-valued only, and which may be negative"; the boolean paragraph's last sentence becomes "A researcher creates boolean Properties; no seeded Property is boolean. A boolean Observation is never negative: No is the value that says it (`observations` refuses a negative one)."
- `.agents/skills/add-seeded-vocabulary/SKILL.md` (~60): the Create sentence reads "(text, integer, date, name, subject, boolean)".
- `.agents/skills/add-match-rule/SKILL.md` (~61): "A new value type adds a row to `propertymeta.Types`" gains "and says whether it is researcher-creatable, single-only and negatable".
- `docs/ideas/structural-debt.md`: add the FakeStore row (`appendFakeObservations` near-copies `fakeObservation`, with an engine-like `displayText` chain; S14-37 review M-7); add `CitationObservationValue.summary` to S14-36's display row.
- `../deployment-plan.md`: S14-37 row status.
- `../completed.md`; this plan; `../design/S14-D8-composer-boolean-value.md` → `../design/archive/`; `../design/README.md` row → Archived.
- `VERSION`, `core/version.go`, `MARKETING_VERSION`: the minor bump.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-37-boolean-value-type-ui.md` reports no errors; `rg -n "S14-37" docs/interpretation-layer-data-model.md .agents/skills` finds nothing.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/propertymeta/propertymeta_test.go` `TestTypes`; `core/database/properties/properties_test.go` `TestValueTypeRules` (rows changed / added); `api/ffi/handlers/subject_defs_test.go` `TestGetPropertiesWorkspaceValueTypes` (+ `negatable`, + boolean row by name) | boolean creatable, single-only, not negatable, others negatable; Create boolean "" → single; Create boolean multiple → `ErrInvalid`; wire boolean `researcher_creatable` true, `negatable` false | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-1, D-3 |
| T-2 | Go | `core/database/observations/observations_test.go` `TestNegationFollowsTheValueType` *(new)*, table-driven | negative boolean add → `ErrInvalid`, no row; raw-SQL negative boolean updated still negative → `ErrInvalid`, row unchanged; same update positive → ok; negative text add → stored | AC-1.5, AC-1.6, AC-1.7, D-2 |
| T-3 | Swift | `CitationComposerModelTests` `booleanAnswerSavesOnlyOnceChosen` *(new)* | Subject + Can read → `authoringCanConfirm` false, `valueBoolean` nil; set false, save → `addObservationsToCitation` prefix in `recordedCalls`, `observationsByGraph[graphID]` Can read row `valueBoolean == false`; saved row summary "No" | AC-2.1, AC-2.2, AC-2.3, D-5, D-6 |
| T-4 | Swift | `CitationComposerModelTests` `propertyChangeResetsOnlyWhenThePropertyChanges` *(new)* | Occupation + negative → Can read: `valueBoolean` nil; Yes, save → stored row polarity `positive`; Can read + true → Occupation: nil, ""; Can read + true → Can read again: true | AC-2.4, AC-2.5, AC-2.6, D-7 |
| T-5 | Swift | `CitationComposerModelTests` `editingABooleanOpensOnItsAnswer` *(new)* | seeded `canReadObservation(value: true)`, Edit → `valueBoolean == true`, `state == .saved`, save records no `updateObservation` | AC-2.7 |
| T-6 | Swift | `CitationComposerModelTests` `negativeIsOfferedByGoRule` *(new)*, `booleanSummaryReadsYesNoOrNothing` *(new)* | `offersNegative` Can read → false, Occupation → true, no rules → true; `asDraft` of a negative Can read row → positive; `summary` true/false/nil → "Yes"/"No"/""; a fresh draft `isEmptyDraft` | AC-2.8, AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-4 |
| T-7 | Swift | `ObservationValueDisplayTests` `booleanReadsYesOrNo` *(new)* | `valueBoolean` true → "Yes", false → "No", nil → "" | AC-4.1, AC-4.2, D-11 |
| — | manual | §10 steps 2 – 6 | Yes / No drawing, Tab, VoiceOver, Wide, Conclusion detail | AC-4.3, AC-5.1, AC-5.2, D-9, D-10 |

Assert copy with `L10n.string(L10n.Properties.valueBooleanYes)` and `valueBooleanNo`, not literals. Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CitationComposerModelTests` and `CitationComposerDepictionTests` (the STEP-3 rename), `TestTypes`, `TestGetPropertiesWorkspaceValueTypes`, `PropertiesModelTests`, `scripts/check-localizable-xcstrings.py`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (two removed)
- [ ] Sweep: `rg -n "togglePolarity|polarityMenuTitle|negateObservation|affirmObservation|applyProperty\(rowID|fields\(from:" .agents .cursor/rules docs/*.md macos` finds nothing; `rg -n "S14-37" docs/interpretation-layer-data-model.md .agents/skills` finds nothing; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the S14-D8 board and inventory; the only kit change is D-10
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — minor
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S14-37 row status Landed #N; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc, skills and structural debt updated per STEP-8
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a new project, with S14-39 landed:

1. Properties → Create: the picker offers boolean. Create "Can read", bind it to Person. Holds offers one value only (S14-39). (AC-1.2)
2. Open a Source's Citation composer, add an observation for a Person, pick Can read. The Value field shows Yes / No, neither chosen; no Negative checkbox; Save disabled. Pick Occupation: the Negative checkbox is back. (AC-2.2, AC-3.1, AC-3.2)
3. Pick Can read again and press Tab: focus lands on Yes. Space chooses it. VoiceOver reads "Value", "Yes", selected. (AC-5.1, AC-5.2)
4. Save "Can read: No". The evidence column reads "Can read No". Edit it: No is selected. (AC-2.3, AC-2.7)
5. On a second Citation, save "Can read: No" for the same Person. The Evidence graph card reads "No". Promote; on the Person's Conclusion detail, Can read reads "No", merged. Cite "Yes" on a third: it reads mixed. (AC-4.1, AC-4.3)
6. Widen the window past 1600pt of pane: the dialog and evidence column look the same. (BV-W1)

This is the slice 7e Check: bar outcome 31 on a new project.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The STEP-3 rename changes behaviour | It is a type change only; the compiler finds every site; STEP-3's check runs the whole suite before any boolean code |
| A hidden Negative tick saves a negative boolean | `asDraft` writes positive (D-6); D-2 refuses it in Go; T-2, T-3 |
| The focus handle does not land on a segment | STEP-6 runs manual check 3 before going on; Stop condition below |

**Stop and report if:**

- `rg -n "INSERT INTO observations|UPDATE observations SET" core --type go -g '!*_test.go'` finds a writer outside `core/database/observations` and `core/database/graphcache/graphcachetest` (D-2 would not cover it).
- S14-39 has not landed (the create form would still offer Holds several for boolean).
- With `focus:` applied, Tab from the Property combo box does not land on the Yes segment, or `focus.apply` on a `PVChip` does not take keyboard focus.
- `PVSegmentedControl` with a `nil` selection draws a selected segment.
- STEP-3 needs any change other than the rename and D-8's deletions to keep the suite green.

## 12. Open questions

None from the code. The plan stays Blocked until S14-39 lands and the S14-D8 board is agreed; then §4 and §6 macOS are refreshed and `review-plan` runs round 2.

## 13. PR

**Title:** S14-37 — Boolean value type (UI)
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-37-boolean-value-type-ui.md`.
