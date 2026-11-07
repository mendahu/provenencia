# S14-36 — Boolean value type (engine)

**Status:** Planned
**Spike / slice:** Spike 14 · Slice 7e — Boolean value type
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Properties create picker offers the same five types, now read from Go; no view offers, edits or draws a boolean until S14-37)
**Board:** —
**Depends on:** —
**Blocks:** S14-37
**Branch:** `s14-36-boolean-value-type-engine`
**Version bump:** none — engine and wire only; a researcher cannot create a boolean Property until S14-37 marks it creatable ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (snapshot only), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** working rule "(Batch 7) Boolean is a first-class value type in two PRs"; working rule "(Batch 5) FakeStore is a seeded fake, not a second engine"; bar outcome 31 (completed by S14-37); dogfood note [`ux/2026-10-25-boolean-property-value-type.md`](../ux/2026-10-25-boolean-property-value-type.md); [structural debt](../../../ideas/structural-debt.md) SourceSurface row; [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) (value types, Property create rules); [`seeded-vocabulary.md`](../../../seeded-vocabulary.md) §3.2
**Planned:** 2026-11-07 against `main` @ `f971f40`
**Review:** [`S14-36-boolean-value-type-engine.review.md`](S14-36-boolean-value-type-engine.review.md) — Fits with changes — Ready with changes, 2 rounds; no open blocker; To user: M-2 answered (Go sends the rules on the wire)

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

Yes/no facts from a record ("literate", "can sign name", "veteran" on a census) get their own value type, `boolean`, stored, compared, reconciled and sent over the wire like `integer`. This PR is the engine half. Go declares every value type and its rules (researcher-creatable, single-only) in one table and sends that table to the app, which stops keeping its own list. Go stores true/false Observations in a dedicated column, reconciles them, and sends them to Swift, whose wire mapping and value text learn the new case. The PR also pins Swift's `SourceSurface` raw values to Go's `workspacelocation.Surface*` constants.

**Why now.** Batch 7 pulled the dogfood note in; S14-37 (composer checkbox, Properties create, Conclusion display) cannot start until the engine and wire carry booleans and their rules.

**What stays.** Every existing value type reads, writes and reconciles as today. The Properties create picker offers the same five types (`text`, `integer`, `date`, `name`, `subject`): boolean ships not creatable, and S14-37 flips that one flag. The composer has no boolean editor (S14-37). No seed Property is boolean.

## 2. User stories

**US-1.** As the engine, I own the value types and their rules in one table, and I store a boolean Property single-valued.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `propertymeta.ValueTypes()` | The eight types, `boolean` included, in table order |
| AC-1.2 | `properties.Upsert` of a user-origin boolean Property with cardinality `""` | Stored, `value_type = boolean`, `cardinality = single` |
| AC-1.3 | `properties.Upsert` of a boolean Property with cardinality `multiple` | `ErrInvalid`; no row |
| AC-1.4 | `properties.Update` of a boolean Property to cardinality `multiple` | `ErrInvalid`; row unchanged |
| AC-1.5 | `properties.Create(label "Literate", valueType "boolean")` in this PR | `ErrInvalid` (boolean is not researcher-creatable until S14-37) |
| AC-1.6 | `properties.Create` with `integer` and cardinality `multiple` | Created, as today |
| AC-1.7 | An existing catalog (format before this PR) is opened | It migrates; existing Properties and Observations read back unchanged; `verifySchema` passes |

**US-2.** As the engine, when a Citation records a boolean Observation, I store exactly true or false and give it back.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `AddToCitation` with `{Property Literate, ValueBoolean false, HasBoolean true}` | Row stored with `value_boolean = 0`; the returned, listed and raw-read Observation has `HasBoolean true, ValueBoolean false`; audit row `value_boolean: false` |
| AC-2.2 | Same, with `HasBoolean false` | `ErrInvalid` |
| AC-2.3 | A boolean Observation that also carries `HasText` or `HasInteger` | `ErrInvalid` |
| AC-2.4 | A text Property Observation that carries `HasBoolean` | `ErrInvalid` |
| AC-2.5 | `Update` of that Observation to `true` | Row `value_boolean = 1`; audit diff `value_boolean: {Old: false, New: true}`; `Update` with the same value → no change |
| AC-2.6 | `CreateCitationWithObservations` over FFI with an `ObservationDraft` whose `value_boolean` is set to `false` | Stored false; the response Observation has `value_boolean` present and false |

**US-3.** As the engine, when members of one handle cite a boolean Property, I reconcile their values as I do integers.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `valuecompare.Compare(true, true)` | `Same`, similarity 1, cause `exact` |
| AC-3.2 | `valuecompare.Compare(true, false)` | `Different`, cause `no_match` |
| AC-3.3 | A handle with Sources A and B both saying `true` | One displayed value `true`, state `merged`, support 2 |
| AC-3.4 | A handle with Source A `true`, Source B `false` | State `mixed` (single cardinality) |
| AC-3.5 | `GetConclusionDetail` on AC-3.3's handle | `ConclusionValue` kind `boolean = true`; a `false` value is sent as `boolean = false`, not as an empty value |

**US-4.** As the app, I read the value-type rules from Go and map and word booleans.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `GetPropertiesWorkspace` | `value_types` lists all eight types with `researcher_creatable` true for text, integer, date, name, subject and false for term, depiction, boolean; `single_only` true for boolean only |
| AC-4.2 | `PropertiesModel` over a snapshot from AC-4.1 | Create offers text, integer, date, name, subject in that order; `canSubmitCreate` refuses `"boolean"` and `"term"` |
| AC-4.3 | Proto `ConclusionValue{boolean: false}` → `GoStore.mapConclusionValue` | `CatalogConclusionValue.boolean(false)` |
| AC-4.4 | `ReconciledValueDisplay.string(for: .boolean(true), propertyKey: "literate")` and the same with `.boolean(false)` | "Yes" and "No" |
| AC-4.5 | Proto `Observation` with `value_boolean = false` → `GoStore.mapObservation` → back through `mapCatalogObservation` | `valueBoolean == false` both ways; a proto without it maps to `nil` |
| AC-4.6 | `PropertyValueType.boolean.label` | "Yes / no" |

**US-5.** As a contributor, when either side's vocabulary drifts, a Go test fails.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `TestSwiftVocabularyMatchesGo` "value types" | Swift `PropertyValueType` cases equal `propertymeta.ValueTypes()` |
| AC-5.2 | `TestSwiftVocabularyMatchesGo` "source surfaces" *(new row)* | Swift `SourceSurface` cases equal Go's `Surface*` constants plus the Swift-only list `{promote}`; renaming `citationComposer` on either side, or adding a Swift case, fails it |

## 3. Scope

| In | Covered by |
| --- | --- |
| `propertymeta` value-type table (`ValueTypeBoolean`, `TypeRule`, `Types`, `ValueTypes`, `Rule`) | STEP-1, AC-1.1 |
| Migration `000061.sql`: `properties` CHECK gains `boolean`; `observations.value_boolean`; `auto_reconciler_values.value_boolean` | STEP-1, AC-1.7 |
| `properties` rules read the table: `valueTypeOK`, `researcherCreatable`, `cardinalityAllowed` in `Upsert`, `Create`, `Update` | STEP-2, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6 |
| `database.NullBool`; Observation storage, one audit row map and one resolved-value fill for insert and update | STEP-3, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| `valuecompare` boolean comparer | STEP-4, AC-3.1, AC-3.2 |
| `autoreconcile` boolean value | STEP-5, AC-3.3, AC-3.4 |
| Reconciler cache and readers: `autoreconciler`, `graphcache`, `graphfill`, `promotealign`, `conclusiondetails`, `deleteimpact` facet row | STEP-6, AC-3.3, AC-3.5 |
| Proto: `ConclusionValue.boolean`, `Observation.value_boolean`, `ObservationDraft.value_boolean`, `PropertyValueTypeRule` and `GetPropertiesWorkspaceResponse.value_types`; handlers | STEP-7, AC-2.6, AC-3.5, AC-4.1 |
| Swift: value-type rules from the snapshot (Swift's own list deleted), FakeStore seeds them; wire mapping and "Yes"/"No" text for booleans | STEP-8, AC-4.2, AC-4.3, AC-4.4, AC-4.5, AC-4.6 |
| `swift_vocabulary_test.go`: "value types" reads the table; new "source surfaces" row | STEP-9, AC-5.1, AC-5.2 |
| Docs and skill: data models, seeded vocabulary §3.2, matching model, `add-seeded-vocabulary`, structural debt, map | STEP-10 |

| Out | Home |
| --- | --- |
| Boolean marked researcher-creatable (one table flag) and offered in Properties create | S14-37 |
| Properties page hides "Holds several" for a single-only type (reads `singleOnly` from the snapshot) | S14-37 (brief S14-D8 or its own board: it changes the Properties page) |
| Value-type pill colour for boolean (`PropertiesTypeChrome.valueTypeForeground` in `macos/App/Features/Catalog/SubjectTypeChrome.swift`) | S14-37 (board) |
| Composer checkbox editor (`CitationObservationValue`, `CitationComposerVocabulary.supportedValueTypes`) | S14-37 (brief S14-D8) |
| Evidence graph card text for a boolean Observation (`ObservationValueDisplay`, calling `PropertyValueType.booleanText`) | S14-37 |
| Omnibar indexing of boolean values | not planned (integers are not indexed either; `core/database/searchindex/handles.go` reads text, names and dates only) |
| A boolean seed Property; a Subject-matching weight for one | not planned (researcher-created only; seeds own weights) |
| One shared struct for the cached value columns (`graphcache` `Value`, `TruthRow`, `EvidenceObservation`, `promotealign.decode`) and a leaf package for the Observation row map that `deleteimpact` can share | [structural debt](../../../ideas/structural-debt.md) (row added in STEP-10) |
| Integer text differs between `ReconciledValueDisplay` and `ObservationValueDisplay` | [structural debt](../../../ideas/structural-debt.md) (row added in STEP-10) |
| A generic `optional[T]` in place of `optionalInt64` / `optionalBool` | not planned (review nit M-9) |

## 4. Current state

Verified at `f971f40`.

| Area | Where | Today |
| --- | --- | --- |
| Value-type constants | `core/propertymeta/propertymeta.go` const block | Seven `ValueType*`; no list function, no per-type rules |
| Value-type rules | `core/database/properties/properties.go` `valueTypeOK` (~390), `researcherCreatable` (~402), `cardinalityOK`; used in `Upsert` (~89), `Create` (~128, ~152), `Update` (~217) | Hand-written switches; the comment says `researcherCreatable` "matches Swift's" |
| Properties CHECK | `core/database/migrations/000052.sql` `properties__new` | Last rebuild; `CHECK (value_type IN ('text','integer','date','name','subject','term','depiction'))`; no index besides inline `UNIQUE (key, origin)` |
| Observations table | `000026.sql` `observations`; `000052.sql` adds `value_depiction_id` | One column per scalar type; `value_integer INTEGER` |
| Reconciler cache | `000034.sql` `conclusion_resolved_values` → `auto_reconciler_values` (`000040.sql`); `value_image_file_id` (`000053.sql`) | `value_integer INTEGER` sibling columns; derived, `BucketSkip` |
| Null binders | `core/database/null.go` `NullBlob`, `NullString`; `core/database/observations/observations.go` `nullInt64(p *int64) any` (~788, private) | No bool binder |
| Observation write | `observations.go` `sqlInsert`, `sqlGetRow`, `sqlListSelect`, `Observation`, `Input`, `resolvedValue`, `resolveValue`, `hasOtherScalars`, `scanListed`; insert builds its audit row inline (~376–392) and fills `obs` from `resolved` (~420–440) | Column list in four hand-kept places |
| Observation update / raw read | `core/database/observations/update.go` (SQL ~21, fill ~118–132, `observationUpdateChange` diff ~143–185); `core/database/observations/write.go` `scanRawObservation`, `observationRowMap` (~252) | Second copy of the fill; `observationRowMap` is the delete path's row map |
| Comparers | `core/valuecompare/valuecompare.go` `comparers`, `Value`; `term.go` | No boolean comparer |
| Comparer tests | `core/valuecompare/contract_test.go` `samples`, `generators` (~224), `TestComparerContract` (`allTypes`, ~340, a hand list), `TestKeyContract`; `valuecompare_test.go` `TestCompareUnsupported` (hand `want`); `core/autoreconcile/modules_test.go` `TestModuleFor` (hand list, ~104) | `TestComparerContract` runs only `allTypes ∩ ComparerTypes()`, so a type missing from `allTypes` is silently skipped |
| Reconcile | `core/autoreconcile/autoreconcile.go` `Value` (`Integer`, `HasInteger`), `Carries`, `SortKey`; `modules.go` `moduleFor`, `Value.compareValue` | Integer via `keyModule` |
| Cache load / write | `core/database/autoreconciler/autoreconciler.go` `sqlInsert`, `sqlLoadCandidates`, candidate scan (~460–486), `insertRow` (~571–615, binds each scalar inline); `truth.go` (~75) | `value_integer` threaded |
| Cache readers | `core/database/graphcache/graph.go` `Value` (~117), `evidence_graph.go` `EvidenceObservation` (~44), `verify.go` `TruthRow` (~17), `SameDisplay` (~259), `rowOf` (~271), `load.go` (~120–156); `core/database/graphfill/load.go` (~119–149); `core/database/promotealign/profiles.go` `decode` (~42, positional) and its two callers | `Integer`/`HasInteger` threaded |
| Profile conversion | `core/subjectprofile/subjectprofile.go` `Value` (~219) | The one conversion from `autoreconcile.Value` to `valuecompare.Value`; copies `Integer`, not a boolean |
| Conclusion detail | `core/database/conclusiondetails/conclusiondetails.go` `Value` (~55), recorded SQL (~151), scan (~442–458), rank copy (~399) | Integer threaded |
| Delete impact | `core/database/deleteimpact/facets.go` `eraseObservation` (~403–460): own SELECT, scan and audit row; `projectors.go` `projectObservation` | `deleteimpact` cannot import `observations` (the import runs the other way) |
| Proto | `api/proto/engine.proto` `ConclusionValue` (oneof 1–6), `Property.value_type` comment (stale: no `depiction`), `Observation` (1–21), `ObservationDraft` (1–13), `GetPropertiesWorkspaceResponse` (1–3) | No boolean, no value-type rules |
| Handlers | `api/ffi/handlers/observations.go` `valueInput`, its two callers, `observationProto`, `optionalInt64` (~413); `source_read.go` `listedFromSource`; `conclusion_detail.go` `conclusionValueProto`; `subject_defs.go` `GetPropertiesWorkspace` | Integer threaded |
| Swift value kinds | `macos/App/Platform/InterpretationValueKinds.swift` `PropertyValueType` | Seven cases with `label` |
| Swift creatable list | `macos/App/Features/Properties/PropertiesModel.swift` `SubjectPropertyValueType.researcherCreatable` (line 7, string array), used at ~147 `canSubmitCreate`, ~460 `submitCreate`; `PropertiesView.swift` ~831 create picker | A copy of Go's rule with no test |
| Swift snapshot | `macos/App/Features/Properties/PropertiesSnapshot.swift` `PropertiesSnapshot` (+ `.empty`); `macos/App/Platform/GoStore.swift` `getPropertiesWorkspace` (~1918); `macos/App/Platform/FakeStore.swift` `getPropertiesWorkspace` (~2549) | No value-type rules |
| Swift models / mapping | `macos/App/Platform/GenealogyStore.swift` `CatalogConclusionValue`, `CatalogObservation`, `CatalogObservationDraft`; `GoStore.swift` `mapConclusionValue` (private), `mapObservation`, `mapCatalogObservation`, `mapObservationDraft` | Exhaustive switch over the proto oneof |
| Swift text | `macos/App/Platform/ReconciledValueDisplay.swift` `string(for:propertyKey:locale:orders:)` | Exhaustive switch over `CatalogConclusionValue` |
| Vocabulary test | `api/ffi/swift_vocabulary_test.go` "value types" row (hand-written want list); no `SourceSurface` row | `rawCase` reads bare `case x` lines |
| Surfaces | `core/workspacelocation/workspacelocation.go` `SurfacePage`, `SurfaceGraph`, `SurfaceCitationComposer`; `macos/App/Features/Workspace/WorkspaceLocation.swift` `SourceSurface` (`page`, `graph`, `citationComposer`, `promote`) | Matched by a comment only |
| Tests | `core/database/properties/properties_test.go` `TestProperties`; `core/database/observations/observations_test.go` `TestObservations`; `core/autoreconcile/autoreconcile_test.go` `TestReconcile`, `TestSortKey`; `core/database/autoreconciler/autoreconciler_test.go` `TestReconciledValues`; `api/ffi/handlers/conclusion_detail_test.go` `TestConclusionValueProtoCoversEveryKind`; `api/ffi/handlers/citations_test.go`; `macos/ProvenenciaTests/PropertiesModelTests.swift` (uses `SubjectPropertyValueType.researcherCreatable` at ~273, ~797); `ReconciledValueDisplayTests.swift`; `ObservationDepictionMappingTests.swift` | Integer cases are the template |

**Precedent.** S10-32 added `depiction` (`000052.sql`: `properties` rebuild + `ALTER TABLE observations ADD COLUMN`) and S10-52 added its cache column (`000053.sql`): copy their DDL. For every Go and Swift site, copy the `integer` threading line for line. Copy `term.go` for the comparer. Copy `vocabkey.PrimaryKinds()` and the "primary kinds" vocabulary row for a declared list read by the test.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A dedicated column `value_boolean INTEGER CHECK (value_boolean IN (0, 1))` on `observations`, and `value_boolean INTEGER` (no CHECK) on the derived `auto_reconciler_values` | Each column names its type; `hasOtherScalars` and every reader stay one-column-per-type; the working rule and the note prefer it | Reusing `value_integer` 0/1 — because every reader would branch on the Property's type to tell 1 from true |
| D-2 | Field names follow the integer field in each struct: `observations.Observation` and `Input` get `ValueBoolean bool` + `HasBoolean bool`; every other struct that has `Integer`/`HasInteger` gets `Boolean bool` + `HasBoolean bool` (`autoreconcile.Value`, `conclusiondetails.Value`, `graphcache.Value`, `TruthRow`, `EvidenceObservation`, the autoreconciler candidate/row types). `valuecompare.Value` gets `Boolean bool` only (an integer there has no `Has`) | One rule: the sibling's spelling, and false stays distinct from absent | `*bool` — because no sibling uses pointers |
| D-3 | `database.NullBool(p *bool) any` *(new, `core/database/null.go`)* returns nil, or `int64(0)` / `int64(1)`. `resolvedValue.Boolean *bool` *(new field)*. The Observation insert and update and `autoreconciler.insertRow` bind through `NullBool` (the cache passes `&row.Boolean` when `row.HasBoolean`, else nil). Scans read `sql.NullBool` | One binder for the new column on both write paths, beside its siblings | A private `nullBool` in `observations` — because the cache write needs it too; binding a Go `bool` — because it relies on driver conversion |
| D-4 | `propertymeta` declares the value types once: `type TypeRule struct { ValueType string; ResearcherCreatable bool; SingleOnly bool }` *(new)*, `var Types = []TypeRule{…}` *(new)* in the order text, integer, date, name, subject, term, depiction, boolean; `func ValueTypes() []string` *(new)* and `func Rule(valueType string) (TypeRule, bool)` *(new)*. Rows: text, integer, date, name, subject creatable; term, depiction, boolean not; boolean `SingleOnly` | Every rule about a value type has one owner; the next type is one row (review M-1) | Hand-written switches in `properties` — the four lists the review found |
| D-5 | Boolean ships `ResearcherCreatable: false`; S14-37 flips it when Properties create and the composer can handle it. S14-36's tests create boolean Properties through `Upsert` with origin `user` | Swift now reads creatability from Go (D-11), so `true` here would put boolean in the picker before its editor exists | `true` now — because it would ship S14-37's UI |
| D-6 | `properties`: `valueTypeOK(vt)` = `Rule(vt)` found; `researcherCreatable(vt)` = rule found and `ResearcherCreatable`; `cardinalityAllowed(valueType, cardinality string) bool` *(new)* = `cardinalityOK(c) && !(rule.SingleOnly && c == multiple)`, called where each `cardinalityOK` call is today in `Upsert`, `Create`, `Update` (in `Update`, before the origin lock check, as today). Error: `ErrInvalid`. The comment that says the list "matches Swift's" is deleted | User, 2026-11-07: a boolean with two answers is a conflict to show, not two kept values | — |
| D-7 | Comparer `boolean{}` in `core/valuecompare/boolean.go` *(new)*: equal → `Same`, similarity 1, `CauseExact`; unequal → `Different`, `CauseNoMatch`. Key at `Same` and `Equivalent`: `"true"` / `"false"`, always present | A yes/no fact has no nearness; the `term` shape | Treating false as no evidence — because "not literate" is a recorded fact |
| D-8 | `autoreconcile`: boolean uses `keyModule` (with text, integer, depiction); `Carries` reads `HasBoolean`; `SortKey` is `"0"` for false, `"1"` for true; `compareValue` passes `Boolean` | Same module as integer | A new module — because nothing folds |
| D-9 | Negative polarity gets no special case: a negative `true` denies positive `true` candidates; it does not create a `false` value | One rule for every type; the composer (S14-37) decides what it offers | Folding "not true" into "false" |
| D-10 | Proto: `ConclusionValue` `bool boolean = 7;` in the oneof; `Observation` `optional bool value_boolean = 22;`; `ObservationDraft` `optional bool value_boolean = 14;`; `message PropertyValueTypeRule { string value_type = 1; bool researcher_creatable = 2; bool single_only = 3; }` *(new)*; `GetPropertiesWorkspaceResponse` `repeated PropertyValueTypeRule value_types = 4;`; the `Property.value_type` comment says "a propertymeta value type; GetPropertiesWorkspace lists them" | Next free numbers; `optional` and the oneof keep false distinct from unset; the rules ride the read the Properties page already makes (user, 2026-11-07: Go sends the rules on the wire) | A new RPC — because the Properties workspace read already loads every Property; a Swift mirror pinned by a test — the user chose the wire |
| D-11 | Swift: `CatalogPropertyValueTypeRule` *(new, `GenealogyStore.swift`)* `{ valueType: String; researcherCreatable: Bool; singleOnly: Bool }`; `PropertiesSnapshot.valueTypeRules: [CatalogPropertyValueTypeRule] = []` *(new, declared last, so existing memberwise call sites compile unchanged)* and `var creatableValueTypes: [String]` *(new)* in Go's order; `PropertiesModel.offeredValueTypes` *(new)* is `snapshot.creatableValueTypes`, or `[draft.valueType]` while the snapshot has no rules (before the first load; the draft defaults to "text"); the create picker shows `offeredValueTypes`. `PropertiesModel.acceptsValueType(_:)` *(new)* returns `snapshot.valueTypeRules.isEmpty || snapshot.creatableValueTypes.contains(vt)`, and `canSubmitCreate` and `submitCreate` call it: before the first load Go decides, and its refusal shows through the existing create error path. `SubjectPropertyValueType.researcherCreatable` is deleted (`SubjectPropertyValueType.label` stays). `singleOnly` is mapped and not read until S14-37 | One owner for the rule (review M-2); ⌘N and the header Add open create before the page has loaded (`PropertiesView` `header`, the ⌘N button), so the form must work without the rules | Keeping the Swift list — the drift the review found; disabling create until load — a behaviour change with no board |
| D-12 | FakeStore has `var valueTypeRules: [CatalogPropertyValueTypeRule] = []` *(new)*, returned by `getPropertiesWorkspace`; a test that needs rules seeds them, as tests seed `propertiesByProject`. No default list | Seeded data, per working rule "FakeStore is a seeded fake"; a default list would be an unpinned Swift copy of `propertymeta.Types` (round-2 M-11) | A static default list — an unpinned copy; a pinned default — a mirror the user's M-2 answer turned down |
| D-13 | `deleteimpact` `eraseObservation` selects, scans and records `value_boolean` (true/false/nil); `projectObservation` adds no boolean arm, so a boolean Observation's impact row is titled by its Property label | Go keeps no English "Yes"/"No"; depiction already titles this way | A Go "yes"/"no" literal |
| D-14 | `observations`: the insert builds `obs` first and records `rowchange.FullRow(observationRowMap(obs))`; one `applyResolved(obs *Observation, r resolvedValue)` *(new)* fills the value fields for insert and update. `observationUpdateChange` keeps its per-column diff and gains the `value_boolean` line | One column list for the audit row (review M-3); the update diff stays per column because date and name diffs are value-row diffs, not column diffs | Deriving the update diff from two row maps — because date and name changes are recorded on their own value rows |
| D-15 | `conclusionValueProto` arm `v.HasBoolean` → `ConclusionValue_Boolean`, placed beside the integer arm | Exactly one field is set | — |
| D-16 | `valueInput` takes `valueBoolean *bool` after `valueInteger`; the calls become `valueInput(&in, o.GetValueText(), o.ValueInteger, o.ValueBoolean, o.GetDate(), …)` and the same with `d`; `observationProto` sets `ValueBoolean: optionalBool(o.HasBoolean, o.ValueBoolean)` *(new, beside `optionalInt64`)* | Same shape as integer | — |
| D-17 | Swift mapping and text (user, 2026-11-07): `PropertyValueType.boolean`; `CatalogConclusionValue.boolean(Bool)`; `valueBoolean: Bool? = nil` declared directly after `valueInteger` in `CatalogObservation` and `CatalogObservationDraft`; `GoStore` maps both ways; `GoStore.mapConclusionValue` becomes `static` (not `private`) so a test calls it; the Observation round trip is tested through the existing `GoStore.updateObservationRequest`, as `ObservationDepictionMappingTests` does (~51), and `mapCatalogObservation` stays private | Regenerating the proto makes the Swift switch non-exhaustive; S14-01 set the precedent that the engine PR owns the Swift wire mapping | Mapping to `.none` until S14-37 — the user chose the full wire mapping |
| D-18 | One boolean text function: `extension PropertyValueType { static func booleanText(_ value: Bool) -> String }` *(new, `InterpretationValueKinds.swift`)* returns `L10n.string(L10n.Properties.valueBooleanYes / valueBooleanNo)`. `ReconciledValueDisplay.string` calls it; S14-37's `ObservationValueDisplay` arm calls it too | One place says how a boolean reads (review M-5) | Keys under `L10n.Conclusions` inline in `ReconciledValueDisplay` |
| D-19 | Copy (user, 2026-11-07): type label "Yes / no"; values "Yes", "No" | Researcher wording | "Boolean", "True"/"False" |
| D-20 | "source surfaces" vocabulary row: `swiftWorkspaceLocation = "../../macos/App/Features/Workspace/WorkspaceLocation.swift"` *(new const)*; `swiftOnlySurfaces = []string{"promote"}` *(new, comment: Promote is a Swift-only place Go never opens)*; want `sorted(append([]string{workspacelocation.SurfacePage, workspacelocation.SurfaceGraph, workspacelocation.SurfaceCitationComposer}, swiftOnlySurfaces...)...)`, regex `rawCase`, with the import `"github.com/mendahu/provenencia/core/workspacelocation"`. "value types" want becomes `sorted(propertymeta.ValueTypes()...)` | User, 2026-11-07; exact equality catches drift on both sides | A Go `SurfacePromote` nothing uses |
| D-21 | No cache-version or search projection-version bump | No catalog can hold a boolean before this migration (the CHECK forbids it) | Bumping `auto_reconciler_meta.cache_version` |

**Rules this plan bends.** None. The working rule leaves storage to `plan-pr` (D-1). D-6, D-10, D-17, D-19 and D-20 carry the user's answers of 2026-11-07; STEP-10 writes the single-only rule and "Go sends value-type rules" into the map's boolean working rule.

## 6. Design

### Codebase fit

Searched: `valueTypeOK`, `researcherCreatable`, `cardinalityOK`, `nullInt64`, `NullString`, `PrimaryKinds`, `comparers`, `keyModule`, `CatalogConclusionValue`, `PropertiesSnapshot`, `rawCase`, `Surface*`, `L10n.Properties.valueType*`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `propertymeta.TypeRule`, `Types`, `ValueTypes`, `Rule` | type / var / func | `vocabkey.PrimaryKinds()` | Same pattern, in the package that owns value types |
| `ValueTypeBoolean` | const | `ValueTypeInteger` | Extends the const block |
| `000061.sql` | migration | `000052.sql`, `000053.sql` | New format step |
| `value_boolean` columns | column | `value_integer` | D-1 |
| `database.NullBool` | func | `database.NullString`, `observations.nullInt64` | Shared because two packages bind the column |
| `properties.cardinalityAllowed` | func | `cardinalityOK` | Wraps it with the table's `SingleOnly` |
| `observations.applyResolved` | func | the two fill blocks it replaces | Removes a copy |
| `valuecompare.boolean` | comparer | `term` | A registry entry, per `add-match-rule` |
| `PropertyValueTypeRule`, `value_types` | proto | `SubjectTypePresentation` on the same response | Extends the existing read |
| `optionalBool` | func | `optionalInt64` | Sibling helper |
| `CatalogPropertyValueTypeRule`, `PropertiesSnapshot.valueTypeRules`, `creatableValueTypes`, `PropertiesModel.offeredValueTypes`, `acceptsValueType`, `FakeStore.valueTypeRules` | Swift type / fields / seed | `CatalogSubjectTypePresentation`, `presentationsByKey` | Same snapshot |
| `CatalogConclusionValue.boolean`, `valueBoolean` | Swift case / fields | `.integer`, `valueInteger` | D-17 |
| `PropertyValueType.booleanText` | Swift func | `PropertyValueType.label` | Same type |
| `L10n.Properties.valueTypeBoolean`, `valueBooleanYes`, `valueBooleanNo` | keys | `valueTypeInteger` | Same namespace |
| `boolean_test.go`, `boolValue`, `ofBool`, boolean generator; "source surfaces" row | tests | `integer_test.go`, `ofInt`, "primary kinds" row | Mirror the existing ones |

**Builds on.** `add-match-rule` "Add a comparer"; `add-catalog-migration` table rebuild; the `PrimaryKinds` declared-list pattern; `TestSwiftVocabularyMatchesGo`.
**Deletes.** `SubjectPropertyValueType.researcherCreatable` (Swift) and its uses; the hand-written switches inside `valueTypeOK` and `researcherCreatable` (their bodies read the table); the comment "matches Swift's SubjectPropertyValueType.researcherCreatable"; the inline audit map in `observations` insert and the duplicate fill blocks; the hand-written want list of the "value types" row; the comment "Surface values match macOS SourceSurface raw values." becomes "Surface values match macOS SourceSurface raw values; TestSwiftVocabularyMatchesGo holds them."; the doc sentences that say `boolean` is not used.
**Next adds.** S14-37 flips `ResearcherCreatable` on boolean's row, reads `singleOnly` to hide "Holds several", and adds an `ObservationValueDisplay` arm calling `PropertyValueType.booleanText`. A later value type is one `Types` row plus the column threading (the shared-struct refactor is logged).

### Reach

| New path or state | Destination | Checked |
| --- | --- | --- |
| The create picker reads `snapshot.creatableValueTypes` | `PropertiesView` create form, `PropertiesModel.canSubmitCreate` / `submitCreate` | Same five types in the same order (AC-4.2). Before the first load (⌘N or Add on a cold page) the picker offers the draft's "text" and Go validates (D-11, T-12) |
| A boolean Property in the Properties list (catalog made outside the app) | `SubjectPropertyValueType.label`, `PropertiesTypeChrome.valueTypeForeground` | Label "Yes / no"; pill colour falls to the default, as `depiction` does today (S14-37); "Holds several" stays offered and Go refuses it with `properties.invalid` (S14-37 hides it) |
| A boolean value on a Conclusion detail | `ReconciledValueDisplay.string` | "Yes"/"No" (AC-4.4) |
| A boolean Observation on an Evidence graph card | `ObservationValueDisplay` | Empty value text until S14-37 |

### Scale

None — no new list, loop or write path; each existing write and read carries one more column, and the workspace read carries eight more small rows.

### Edges

- **Assumptions.** `hasOtherScalars` assumes one value column per type: boolean adds `keep != "boolean" && in.HasBoolean` (T-4). `keyModule` assumes `Key` agrees with `Compare` (`TestKeyContract`, T-6). `conclusionValueProto` assumes exactly one field set (T-10). `PropertiesView`'s picker assumes a non-empty list with "text" as its default selection (AC-4.2 keeps text first).
- **Neighbors.** false vs absent: `HasBoolean true, ValueBoolean false` stores 0 and is carried; `HasBoolean false` is refused (AC-2.1, AC-2.2). `ConclusionValue{boolean:false}` is a set oneof (AC-3.5). Integer 0 and 1 stay integers, and integer `multiple` stays allowed (AC-1.6).
- **Removed inputs.** A deleted boolean Observation removes its candidate; the existing recompute rewrites the cache (T-8).
- **Real data.** None — no seed drives the rule; the eight-row table is checked whole by T-1.
- **Untrusted size.** None — one bit per row; eight fixed rules.

### Schema

Migration `core/database/migrations/000061.sql` *(new)*:

```sql
-- S14-36: value_type = boolean. A yes/no fact: observations.value_boolean
-- holds 0 or 1, and the auto-reconciler cache keeps it beside value_integer.
-- No seed Property is boolean; researchers create them (from S14-37).
-- migrate: foreign_keys off

CREATE TABLE properties__new ( /* 000052.sql's columns, unchanged */
	CHECK (value_type IN (
		'text', 'integer', 'date', 'name', 'subject', 'term', 'depiction', 'boolean'
	))
) STRICT;
INSERT INTO properties__new (id, key, origin, label, description, value_type, cardinality)
SELECT id, key, origin, label, description, value_type, cardinality FROM properties;
DROP TABLE properties;
ALTER TABLE properties__new RENAME TO properties;

ALTER TABLE observations ADD COLUMN value_boolean INTEGER CHECK (value_boolean IN (0, 1));
ALTER TABLE auto_reconciler_values ADD COLUMN value_boolean INTEGER;
```

No new table or FK: no `catalogmodel` registration; `TestPragmaHonesty` stays green. Audit: the Observation row map gains `value_boolean`.

### Go core

| Package | Change |
| --- | --- |
| `core/propertymeta` | D-4 |
| `core/database` | `NullBool` (D-3) |
| `core/database/properties` | D-6 |
| `core/database/observations` | `value_boolean` in `sqlInsert`, `sqlGetRow`, `sqlListSelect`, update SQL; fields (D-2); `resolvedValue.Boolean`; `resolveValue` case `ValueTypeBoolean` (refuse when `!in.HasBoolean || hasOtherScalars(in, "boolean")`); `hasOtherScalars` line; `scanListed`, `scanRawObservation`; `observationRowMap` `"value_boolean"` (bool or nil); `applyResolved` and the insert audit row (D-14); update diff line |
| `core/valuecompare` | `Value.Boolean`; `boolean{}` (D-7) |
| `core/autoreconcile` | D-8 |
| `core/database/autoreconciler` | `value_boolean` directly after `value_integer` in `sqlInsert` and its `Exec` arguments, `sqlLoadCandidates` and the scan `dest`; `insertRow` binds via `NullBool`; `truth.go` copies `Boolean`/`HasBoolean` |
| `core/database/graphcache` | `Boolean`/`HasBoolean` on `Value`, `EvidenceObservation`, `TruthRow`; `SameDisplay` compares them; `rowOf` copies them; `load.go` selects and scans `r.value_boolean` directly after `r.value_integer` |
| `core/database/graphfill` | `load.go` selects and scans `o.value_boolean` directly after `o.value_integer` |
| `core/database/promotealign` | `decode` takes `boolean, hasBoolean bool` after `hasInteger`; both callers pass them |
| `core/database/conclusiondetails` | `Value.Boolean`/`HasBoolean`; recorded SQL, scan, rank copy |
| `core/subjectprofile` | `Value` copies `Boolean: v.Boolean` beside `Integer` |
| `core/database/deleteimpact` | D-13 (`facets.go` only) |

Errors: `ErrInvalid` (`properties.invalid`, `observations.invalid`) for D-5, D-6, AC-2.2, AC-2.3, AC-2.4. No new codes.

### FFI

`api/proto/engine.proto` per D-10. Handlers: D-15, D-16; `source_read.go` `listedFromSource` copies `Boolean`/`HasBoolean`; `subject_defs.go` `GetPropertiesWorkspace` appends one `PropertyValueTypeRule` per `propertymeta.Types` row. Regenerate with `./scripts/generate-proto.sh`; it writes `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, both committed. It needs `protoc`, `protoc-gen-go` and `protoc-gen-swift` (a Mac with the toolchain).

### macOS

No view changes; the picker shows what it shows today.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Value-type rules | `CatalogPropertyValueTypeRule`, `PropertiesSnapshot.valueTypeRules`, `creatableValueTypes` | `GenealogyStore.swift`, `PropertiesSnapshot.swift` | D-11 |
| Mapping | `getPropertiesWorkspace` maps `resp.valueTypes`; `mapConclusionValue`, `mapObservation`, `mapCatalogObservation`, `mapObservationDraft` | `GoStore.swift` | D-17 |
| Fake | `valueTypeRules` (empty unless seeded) | `FakeStore.swift` | D-12 |
| Create picker / guards | existing `PVSegmentedControl` | `PropertiesView.swift`, `PropertiesModel.swift` | `offeredValueTypes`; guard per D-11 |
| Value type case | `PropertyValueType.boolean`, `label`, `booleanText` | `InterpretationValueKinds.swift` | D-18 |
| Value text | `.boolean(let b)` → `PropertyValueType.booleanText(b)` | `ReconciledValueDisplay.swift` | — |

### Strings

L10n keys *(new)* in `L10n.Properties`, beside `valueTypeInteger`, with `Localizable.xcstrings` entries per `add-localized-string` (keep the catalog's sort order):

- `valueTypeBoolean` — `properties.valueType.boolean` = "Yes / no", comment "Property value type label: boolean (a yes/no fact)".
- `valueBooleanYes` — `properties.value.booleanYes` = "Yes", comment "A boolean Property value that is true".
- `valueBooleanNo` — `properties.value.booleanNo` = "No", comment "A boolean Property value that is false".

Removed: none.

## 7. Steps

**STEP-1 — Value-type table and migration**
Files: `core/propertymeta/propertymeta.go`, `core/propertymeta/propertymeta_test.go` (add `TestTypes`), `core/database/migrations/000061.sql` *(new)*. Skill: `add-catalog-migration`, `add-catalog-model`.
Do: D-4; the migration per §6 Schema, copying `000052.sql`'s `properties__new` columns exactly. Add T-1.
Check: `go test ./core/propertymeta/` and `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` pass.

**STEP-2 — Property rules read the table**
Files: `core/database/properties/properties.go`, `properties_test.go`.
Do: D-5, D-6; the existing "rejects bad value type" case uses `"real"` instead of `"boolean"`. Add T-2, T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/properties/` passes.

**STEP-3 — Observation storage**
Files: `core/database/null.go`, `core/database/observations/observations.go`, `update.go`, `write.go`, `observations_test.go`. Skill: `add-catalog-query`.
Do: D-2, D-3, D-14; §6 Go core `observations` row. Add T-4, T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/observations/` passes (the existing audit-row tests hold D-14).

**STEP-4 — Comparer**
Files: `core/valuecompare/valuecompare.go`, `core/valuecompare/boolean.go` *(new)*, `core/valuecompare/boolean_test.go` *(new)*, `contract_test.go`, `valuecompare_test.go`. Skill: `add-match-rule` (Add a comparer).
Do: D-7. In `contract_test.go`: `func boolValue(b bool) valuecompare.Value { return valuecompare.Value{Type: propertymeta.ValueTypeBoolean, Boolean: b} }`, `func ofBool(a, b bool) samplePair`, samples `ofBool(true, true), ofBool(true, false), ofBool(false, true), ofBool(false, false)` (no "no value" pair: a boolean always has a key), generator `func(rng *rand.Rand) samplePair { return ofBool(rng.Intn(2) == 0, rng.Intn(2) == 0) }`. `TestComparerContract`'s `allTypes` becomes `propertymeta.ValueTypes()`; `TestCompareUnsupported`'s want becomes `propertymeta.ValueTypes()` minus a declared `noComparer := map[string]bool{propertymeta.ValueTypeSubject: true}`, sorted. Add T-6.
Check: `go test ./core/valuecompare/` passes.

**STEP-5 — Reconcile**
Files: `core/autoreconcile/autoreconcile.go`, `modules.go`, `autoreconcile_test.go`, `modules_test.go`; `core/subjectprofile/subjectprofile.go`.
Do: D-8, D-9; `subjectprofile.Value` copies `Boolean`; `TestModuleFor` ranges over `propertymeta.ValueTypes()` instead of its hand list. Add T-7.
Check: `go test ./core/autoreconcile/ ./core/subjectprofile/` passes.

**STEP-6 — Cache and readers**
Files: `core/database/autoreconciler/autoreconciler.go`, `truth.go`, `autoreconciler_test.go`; `core/database/graphcache/graph.go`, `evidence_graph.go`, `load.go`, `verify.go`; `core/database/graphfill/load.go`; `core/database/promotealign/profiles.go`; `core/database/conclusiondetails/conclusiondetails.go`; `core/database/deleteimpact/facets.go`.
Do: §6 Go core rows for these packages; D-13, D-21. Add T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/...` passes.

**STEP-7 — Proto and handlers**
Files: `api/proto/engine.proto`, `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/observations.go`, `source_read.go`, `conclusion_detail.go`, `subject_defs.go`, `conclusion_detail_test.go`, `citations_test.go`, `subject_defs_test.go`. Skill: `add-ffi-handler`.
Do: D-10, D-15, D-16; run `./scripts/generate-proto.sh`. Add T-9, T-10, T-11.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes (Swift compiles after STEP-8).

**STEP-8 — Swift: rules from Go, booleans on the wire**
Files: `macos/App/Platform/InterpretationValueKinds.swift`, `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `ReconciledValueDisplay.swift`, `L10n.swift`, `Localizable.xcstrings`; `macos/App/Features/Properties/PropertiesSnapshot.swift`, `PropertiesModel.swift`, `PropertiesView.swift`; tests `macos/ProvenenciaTests/PropertiesModelTests.swift`, `ReconciledValueDisplayTests.swift`, `ObservationDepictionMappingTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-11, D-12, D-17, D-18, D-19; §6 Strings. In `PropertiesModelTests.swift`: the create-flow test (~273) seeds `store.valueTypeRules` with the eight rows of D-4 before the model loads and asserts `!model.offeredValueTypes.contains("term")`; the label test (~797) drops its `researcherCreatable` line (the rule is Go's, held by T-1 and T-11). The test at `core/database/properties/properties_test.go` "rejects bad value type" (~49) changes its invalid type from `"boolean"` to `"real"` (STEP-2). Add T-12, T-13, T-14. Assert copy with `L10n.string(...)`, as `PropertiesModelTests` does at ~794.
Check: Xcode tests pass with no Swift warning (`scripts/check-pr.sh`); `rg -n "SubjectPropertyValueType.researcherCreatable" macos/` finds nothing.

**STEP-9 — Vocabulary pins**
Files: `api/ffi/swift_vocabulary_test.go`, `core/workspacelocation/workspacelocation.go` (comment).
Do: D-20. Add T-15.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; temporarily renaming Swift `citationComposer` fails it (revert).

**STEP-10 — Docs**
Files:
- `docs/interpretation-layer-data-model.md`: the value-type list (~613–635) adds `boolean` (single only; researcher-creatable from S14-37); reword both sentences that say `real` and `boolean` are not used (~627, ~866) to say `real` is not used; add `value_boolean` to the observations sketch (~880).
- `docs/seeded-vocabulary.md` §3.2: `boolean` is a valid `value_type` with no seed row.
- `docs/matching-model.md` (~49): the **value** row reads "One recorded fact of one value type: a name, a date, a piece of text, a term, an integer, a picture, a yes/no".
- `.agents/skills/add-seeded-vocabulary/SKILL.md` (~58–60): the value-type list adds `boolean` and says `propertymeta.Types` is the list; the Create sentence reads "Researcher Create Property accepts only the types `propertymeta.Types` marks creatable (text, integer, date, name, subject; boolean from S14-37)."
- `docs/ideas/structural-debt.md`: the SourceSurface row (already struck) ends "Done in S14-36."; add two rows: the cached value columns repeated across `graphcache` `Value` / `TruthRow` / `EvidenceObservation`, `promotealign.decode`, `conclusiondetails.Value` and `deleteimpact.eraseObservation`'s own row map (S14-36 review M-3, M-6); integer text differs between `ReconciledValueDisplay` and `ObservationValueDisplay` (M-5).
- `../deployment-plan.md`: S14-36 row status (planning already updated its scope entries and the Batch 7 boolean working rule).
- `../completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-36-boolean-value-type-engine.md` reports no errors; `rg -n "boolean. are not" docs/` finds nothing.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/propertymeta/propertymeta_test.go` `TestTypes` *(new test, existing file)* | `ValueTypes()` = the eight in order; `Rule("boolean")` = not creatable, single-only; `Rule("nope")` not found; no duplicate | AC-1.1, D-4, D-5 |
| T-2 | Go | `core/database/properties/properties_test.go` `TestValueTypeRules` *(new)*, table-driven | Upsert boolean "" → single; Upsert boolean multiple → `ErrInvalid`; Create boolean → `ErrInvalid`; Create integer multiple → ok; Create term → `ErrInvalid` | AC-1.2, AC-1.3, AC-1.5, AC-1.6, D-6 |
| T-3 | Go | same file `TestBooleanCardinalityStaysSingle` *(new)* | Update boolean to multiple → `ErrInvalid`, row unchanged | AC-1.4 |
| T-4 | Go | `core/database/observations/observations_test.go` `TestBooleanObservationValue` *(new)*, table-driven | false and true stored, listed, raw-read, audit row; missing → `ErrInvalid`; boolean + text → `ErrInvalid`; text Property + boolean → `ErrInvalid` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-3, D-14 |
| T-5 | Go | same file `TestUpdateBooleanObservation` *(new)* | false → true: row 1, diff `{false, true}`; unchanged → no change | AC-2.5 |
| T-6 | Go | `core/valuecompare/boolean_test.go` `TestBoolean` *(new)*; `TestComparerContract`, `TestKeyContract`, `TestCompareUnsupported` | t/t Same exact; t/f Different no_match; keys "true"/"false" | AC-3.1, AC-3.2, D-7 |
| T-7 | Go | `core/autoreconcile/autoreconcile_test.go` `TestReconcile` rows *(new)*, `TestSortKey` rows *(new)* | two Sources true → merged, support 2; true vs false → mixed; negative true denies true; "0" < "1" | AC-3.3, AC-3.4, D-8, D-9 |
| T-8 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestReconciledBoolean` *(new)* | boolean Property, two members true → cache `value_boolean = 1`, support 2; delete one Observation → support 1; `graphcache` verify passes | AC-3.3, D-21 |
| T-9 | Go | `api/ffi/handlers/citations_test.go` `TestCreateCitationBooleanObservation` *(new)* | draft `value_boolean = false` → response `ValueBoolean != nil` and false | AC-2.6, D-16 |
| T-10 | Go | `api/ffi/handlers/conclusion_detail_test.go` `TestConclusionValueProtoCoversEveryKind` rows *(new)* | `HasBoolean` false → `ConclusionValue_Boolean` false; true likewise | AC-3.5, D-15 |
| T-11 | Go | `api/ffi/handlers/subject_defs_test.go` `TestGetPropertiesWorkspaceValueTypes` *(new)* | `value_types` equals `propertymeta.Types` field for field | AC-4.1, D-10 |
| T-12 | Swift | `macos/ProvenenciaTests/PropertiesModelTests.swift` `createOffersGoCreatableTypes` *(new)*, `booleanLabelReadsYesNo` *(new)* | `store.valueTypeRules` seeded with D-4's eight rows, model loaded → `offeredValueTypes` = text, integer, date, name, subject; `canSubmitCreate` false for "boolean"; with no snapshot loaded → `offeredValueTypes` = ["text"] and `canSubmitCreate` true for a labelled text draft; `L10n.string(PropertyValueType.boolean.label)` = "Yes / no" | AC-4.2, AC-4.6, D-11, D-12 |
| T-13 | Swift | `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift` `aBooleanReadsAsYesOrNo` *(new)* | `.boolean(true)` → "Yes"; `.boolean(false)` → "No" | AC-4.4, D-18 |
| T-14 | Swift | `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift` `booleanValuesMapBothWays` *(new)* | proto `boolean=false` → `.boolean(false)`; proto Observation `value_boolean=false` → `mapObservation` → `valueBoolean == false` → `updateObservationRequest(...).observation.valueBoolean == false` and `hasValueBoolean`; absent → nil | AC-4.3, AC-4.5, D-17 |
| T-15 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` "value types", "source surfaces" *(new row)* | Swift = Go (+ `promote`) | AC-5.1, AC-5.2, D-20 |
| — | Go | existing suites through the schema hash | a migrated catalog opens | AC-1.7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestComparerContract`, `TestCompareUnsupported`, `TestPragmaHonesty`, `graphcache` verify tests, observation audit-row tests, `TestSwiftVocabularyMatchesGo`, `PropertiesModelTests`, `scripts/check-localizable-xcstrings.py`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`SubjectPropertyValueType.researcherCreatable`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief; no view change)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S14-36 row status Landed #N; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain docs and skill updated per STEP-10: interpretation data model, seeded-vocabulary.md §3.2, matching-model.md, add-seeded-vocabulary, structural-debt rows
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project made before this branch. It opens; the Properties list and a Person page read as before (AC-1.7).
2. On the Properties page, start Create. The type picker offers Text, Integer, Date, Name, Subject, in that order, as before (AC-4.2).
3. Create a Property of each offered type: each saves.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A reader of `observations` or `auto_reconciler_values` misses the new column | §4 lists every site from `rg value_integer` / `HasInteger`; STEP-6's check runs `./core/...`; `graphcache` verify compares the cache with truth |
| The `properties` rebuild breaks an FK | `-- migrate: foreign_keys off` runs `foreign_key_check` before commit |
| Create opens before the first load | D-11 falls back to the draft's type and lets Go validate; T-12 covers it |

**Stop and report if:**

- `rg -n "value_integer|HasInteger" --type go` at your base finds a non-test site not listed in §4.
- `000061.sql` is taken on `main`, or a migration after `000060.sql` recreated `properties`.
- `TestPragmaHonesty` or `verifySchema` fails and the fix is anything other than the migration's own DDL.
- A Swift `switch` over `CatalogConclusionValue` or `PropertyValueType` outside §6 macOS fails to compile.
- D-14's shared row map changes any existing audit row's values (an existing observation test fails on a field value).
- `TestComparerContract` needs a change to the contract itself, not a sample row and generator.

## 12. Open questions

None.

## 13. PR

**Title:** S14-36 — Boolean value type (engine)
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-36-boolean-value-type-engine.md`.
