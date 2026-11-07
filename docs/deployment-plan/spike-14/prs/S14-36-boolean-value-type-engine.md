# S14-36 — Boolean value type (engine)

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 7e — Boolean value type
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: no view offers, edits or draws a boolean until S14-37; the one Swift text path added here is reachable only from a catalog that already holds a boolean Property)
**Board:** —
**Depends on:** —
**Blocks:** S14-37
**Branch:** `s14-36-boolean-value-type-engine`
**Version bump:** none — engine and wire only; a researcher cannot create a boolean Property until S14-37 offers it ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** working rule "(Batch 7) Boolean is a first-class value type in two PRs"; working rule "(Batch 5) FakeStore is a seeded fake, not a second engine"; bar outcome 31 (completed by S14-37); dogfood note [`ux/2026-10-25-boolean-property-value-type.md`](../ux/2026-10-25-boolean-property-value-type.md); [structural debt](../../../ideas/structural-debt.md) SourceSurface row; [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) (value types, Property create rules); [`seeded-vocabulary.md`](../../../seeded-vocabulary.md) §3.2; [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) (auto-reconciler cache)
**Planned:** 2026-11-07 against `main` @ `f971f40`
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

Yes/no facts from a record ("literate", "can sign name", "veteran" on a census) get their own value type, `boolean`, stored, compared, reconciled and sent over the wire like `integer`. This PR is the engine half: Go accepts a researcher-created boolean Property (single cardinality only), stores true/false Observations in a dedicated column, reconciles them, and sends them to Swift, whose wire mapping and value text learn the new case. It also pins Swift's `SourceSurface` raw values to Go's `workspacelocation.Surface*` constants.

**Why now.** Batch 7 pulled the dogfood note in; S14-37 (composer checkbox, Properties create, Conclusion display) cannot start until the engine and wire carry booleans.

**What stays.** Every existing value type reads, writes and reconciles as today. Swift's Properties create list (`SubjectPropertyValueType.researcherCreatable`) does not offer boolean yet, and the composer has no boolean editor (S14-37). No seed Property is boolean.

## 2. User stories

**US-1.** As the engine, when a researcher creates a Property typed boolean, I store it single-valued, so a later yes/no fact has a home.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `properties.Create(label "Literate", valueType "boolean", cardinality "")` | Created, `value_type = boolean`, `cardinality = single` |
| AC-1.2 | `properties.Create(label "Veteran", valueType "boolean", cardinality "multiple")` | `ErrInvalid`; no row |
| AC-1.3 | `properties.Update` on a boolean Property with cardinality `multiple` | `ErrInvalid`; row unchanged |
| AC-1.4 | `properties.Upsert` (Install path) of a boolean Property with cardinality `multiple` | `ErrInvalid` |
| AC-1.5 | An existing catalog (format before this PR) is opened | It migrates; existing Properties and Observations read back unchanged; `verifySchema` passes |

**US-2.** As the engine, when a Citation records a boolean Observation, I store exactly true or false and give it back.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `AddToCitation` with `{Property Literate, ValueBoolean false, HasBoolean true}` | Row stored with `value_boolean = 0`; returned and listed Observation has `HasBoolean true, ValueBoolean false` |
| AC-2.2 | Same, with `HasBoolean false` (no value) | `ErrInvalid` |
| AC-2.3 | A boolean Observation that also carries `HasText` or `HasInteger` | `ErrInvalid` |
| AC-2.4 | A text Property Observation that carries `HasBoolean` | `ErrInvalid` |
| AC-2.5 | `Update` of that Observation to `true` | Row `value_boolean = 1`; audit change `value_boolean: {Old: false, New: true}` |
| AC-2.6 | `CreateCitationWithObservations` over FFI with an `ObservationDraft` whose `value_boolean` is set to `false` | Stored false; the response Observation has `value_boolean` present and false |

**US-3.** As the engine, when members of one handle cite a boolean Property, I reconcile their values as I do integers.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `valuecompare.Compare(true, true)` | `Same`, similarity 1, cause `exact` |
| AC-3.2 | `valuecompare.Compare(true, false)` | `Different`, cause `no_match` |
| AC-3.3 | A handle with Sources A and B both saying `true` | One displayed value `true`, state `merged`, support 2 |
| AC-3.4 | A handle with Source A `true`, Source B `false` | State `mixed` (single cardinality; the existing conflict path) |
| AC-3.5 | `GetConclusionDetail` on AC-3.3's handle | `ConclusionValue` kind `boolean = true`; a `false` value is sent as `boolean = false`, not as an empty value |

**US-4.** As the app, when Go sends a boolean, I map it and word it "Yes" or "No".

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Proto `ConclusionValue{boolean: false}` → `GoStore.mapConclusionValue` | `CatalogConclusionValue.boolean(false)` |
| AC-4.2 | `ReconciledValueDisplay.string(for: .boolean(true), …)` / `(.boolean(false))` | "Yes" / "No" |
| AC-4.3 | Proto `Observation` with `value_boolean = false` → `GoStore.mapObservation` → `mapCatalogObservation` | `valueBoolean == false` both ways; a proto without it maps to `nil` |
| AC-4.4 | `PropertyValueType(rawValue: "boolean")?.label` | "Yes / no" |

**US-5.** As a contributor, when either side's vocabulary drifts, a Go test fails.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `TestSwiftVocabularyMatchesGo` "value types" | Passes with `boolean` on both sides; fails if either side drops it |
| AC-5.2 | `TestSwiftVocabularyMatchesGo` "source surfaces" *(new row)* | Swift `SourceSurface` cases equal Go's `Surface*` constants plus the Swift-only list `{promote}`; renaming `citationComposer` on either side, or adding a Swift case, fails it |

## 3. Scope

| In | Covered by |
| --- | --- |
| `propertymeta.ValueTypeBoolean`; `valueTypeOK`, `researcherCreatable`; single-only cardinality for boolean in `Upsert`, `Create`, `Update` | STEP-1, STEP-2, AC-1.1–1.4 |
| Migration `000061.sql`: `properties` CHECK gains `boolean`; `observations.value_boolean`; `auto_reconciler_values.value_boolean` | STEP-1, AC-1.5 |
| Observation storage: insert, update, get, list, audit row maps and diffs, value resolve | STEP-3, AC-2.1–2.5 |
| `valuecompare` boolean comparer | STEP-4, AC-3.1, AC-3.2 |
| `autoreconcile` boolean value: `Carries`, `SortKey`, module, compare value | STEP-5, AC-3.3, AC-3.4 |
| Reconciler cache and readers: `autoreconciler` load/insert/truth, `graphcache`, `graphfill`, `promotealign`, `conclusiondetails`, `deleteimpact` facet row | STEP-6, AC-3.3, AC-3.5 |
| Proto: `ConclusionValue.boolean`, `Observation.value_boolean`, `ObservationDraft.value_boolean`, value-type comments; handlers | STEP-7, AC-2.6, AC-3.5 |
| Swift wire mapping and value text: `PropertyValueType.boolean` + label, `CatalogConclusionValue.boolean`, `CatalogObservation`/`CatalogObservationDraft.valueBoolean`, `GoStore` mappings, `ReconciledValueDisplay` | STEP-8, AC-4.1–4.4 |
| `swift_vocabulary_test.go`: boolean in "value types"; new "source surfaces" row | STEP-9, AC-5.1, AC-5.2 |
| Docs: interpretation and conclusion data models, seeded vocabulary §3.2, structural-debt row, completed.md, map | STEP-10 |

| Out | Home |
| --- | --- |
| Properties create offers boolean (`SubjectPropertyValueType.researcherCreatable`) | S14-37 |
| Composer checkbox editor, `CitationObservationValue` / `CitationComposerVocabulary.supportedValueTypes` | S14-37 (brief S14-D8) |
| Evidence graph card text for a boolean Observation (`ObservationValueDisplay`) | S14-37 |
| Conclusion detail chrome for a boolean beyond the "Yes"/"No" text | S14-37 |
| Omnibar indexing of boolean values | not planned (integers are not indexed either; `core/database/searchindex/handles.go` reads text, names and dates only) |
| A boolean seed Property | not planned (researcher-created only, per the working rule) |
| Subject matching weight for a boolean Property (`matchregistry` / `subjectcompare`) | not planned (researcher Properties carry no weights; seeds own weights) |

## 4. Current state

Verified at `f971f40`.

| Area | Where | Today |
| --- | --- | --- |
| Value-type constants | `core/propertymeta/propertymeta.go` const block | Seven `ValueType*`; no boolean |
| Properties CHECK | `core/database/migrations/000052.sql` `properties__new` | Last rebuild; `CHECK (value_type IN ('text','integer','date','name','subject','term','depiction'))`; no later migration touches `properties` |
| Observations table | `000026.sql` `observations`, `000052.sql` adds `value_depiction_id` | One column per scalar type; `value_integer INTEGER` |
| Reconciler cache | `000034.sql` `conclusion_resolved_values`, renamed `auto_reconciler_values` in `000040.sql`, `value_image_file_id` added in `000053.sql` | `value_integer INTEGER` sibling columns; derived, `BucketSkip` |
| Property rules | `core/database/properties/properties.go` `valueTypeOK`, `researcherCreatable`, `cardinalityOK`, `Upsert`, `Create`, `Update` | No boolean; cardinality checked independently of value type |
| Observation write | `core/database/observations/observations.go` `sqlInsert`, `sqlGetRow`, `sqlListSelect`, `Observation`, `Listed`, `Input`, `resolveValue`, `hasOtherScalars`, `scanListed`, insert change map | `value_integer` / `HasInteger` / `ValueInteger` threaded through each |
| Observation update / raw read | `core/database/observations/update.go` (SQL line 21, `resolved.Integer`, field diff); `core/database/observations/write.go` `scanRawObservation`, `observationRowMap` | Same threading |
| Comparers | `core/valuecompare/valuecompare.go` `comparers`, `Value`; `term.go` (exact-or-different shape) | No boolean comparer |
| Comparer tests | `core/valuecompare/contract_test.go` `samples` + generators; `valuecompare_test.go` `TestCompareUnsupported` (expected `ComparerTypes()`) | Fail on a comparer without a row |
| Reconcile | `core/autoreconcile/autoreconcile.go` `Value`, `Carries`, `SortKey`; `modules.go` `moduleFor`, `Value.compareValue` | Integer via `keyModule` |
| Cache load / write | `core/database/autoreconciler/autoreconciler.go` `sqlInsert`, `sqlLoadCandidates`, candidate scan (~line 460), row insert (~line 574–615); `truth.go` (~line 75) | `value_integer` threaded |
| Cache readers | `core/database/graphcache/graph.go` (row type ~line 122), `evidence_graph.go` `EvidenceObservation`, `load.go` (~line 120–156), `verify.go` (~lines 23, 259, 271); `core/database/graphfill/load.go` (~119–149); `core/database/promotealign/profiles.go` `decode` | Integer fields threaded; `decode` takes `integer, hasInteger` positionally |
| Conclusion detail | `core/database/conclusiondetails/conclusiondetails.go` `Value`, recorded SQL line 151, scan ~442–458, rank copy ~399 | Integer threaded |
| Delete impact | `core/database/deleteimpact/facets.go` (~416–456) audit row; `projectors.go` `projectObservation` | Facet row lists `value_integer`; projector title falls back through value columns (depiction contributes nothing) |
| Proto | `api/proto/engine.proto` `ConclusionValue` (oneof fields 1–6), `Property.value_type` comment, `Observation` (fields 1–21), `ObservationDraft` (fields 1–13) | No boolean |
| Handlers | `api/ffi/handlers/observations.go` `valueInput`, two callers, `observationProto`; `source_read.go` `listedFromSource`; `conclusion_detail.go` `conclusionValueProto` | Integer threaded |
| Swift value kinds | `macos/App/Platform/InterpretationValueKinds.swift` `PropertyValueType` | Seven cases with `label` |
| Swift models | `macos/App/Platform/GenealogyStore.swift` `CatalogConclusionValue`, `CatalogObservation`, `CatalogObservationDraft` | No boolean |
| Swift mapping | `macos/App/Platform/GoStore.swift` `mapConclusionValue` (private), `mapObservation`, `mapCatalogObservation`, `mapObservationDraft` | Exhaustive switch over the proto oneof; fails to compile once the proto gains a case |
| Swift text | `macos/App/Platform/ReconciledValueDisplay.swift` `string(for:propertyKey:locale:orders:)` | Exhaustive switch over `CatalogConclusionValue` |
| Vocabulary test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` "value types" row; no `SourceSurface` row | `rawCase` reads bare `case x` lines |
| Surfaces | `core/workspacelocation/workspacelocation.go` `SurfacePage`, `SurfaceGraph`, `SurfaceCitationComposer`; `macos/App/Features/Workspace/WorkspaceLocation.swift` `SourceSurface` (`page`, `graph`, `citationComposer`, `promote`) | Matched by a comment only |
| Tests | `core/database/properties/properties_test.go` `TestProperties`; `core/database/observations/observations_test.go` `TestObservations`; `core/autoreconcile/autoreconcile_test.go` `TestReconcile`, `TestSortKey`; `core/database/autoreconciler/autoreconciler_test.go` `TestReconciledValues`; `api/ffi/handlers/conclusion_detail_test.go` `TestConclusionValueProtoCoversEveryKind`; `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift`; `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift` | Integer cases are the template |

**Precedent.** S10-32 added `depiction` (`000052.sql`: `properties` rebuild + `ALTER TABLE observations ADD COLUMN`) and S10-52 added its cache column (`000053.sql`). Copy their DDL shape. For every Go and Swift site, copy the `integer` threading line for line (`HasInteger`/`ValueInteger` → `HasBoolean`/`ValueBoolean`); copy `term.go` for the comparer.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A dedicated column `value_boolean INTEGER CHECK (value_boolean IN (0, 1))` on `observations` and `auto_reconciler_values` (no CHECK on the derived cache column) | Each column names its type, as every other value type does; `hasOtherScalars` and every reader stay one-column-per-type; the working rule and the note prefer it | Reusing `value_integer` 0/1 — because every reader would have to branch on the Property's type to tell 1 from true, and the delete-impact title would read "1" |
| D-2 | Go carries a boolean as `ValueBoolean bool` + `HasBoolean bool` on every struct that carries `ValueInteger`/`HasInteger` (`autoreconcile.Value`, `valuecompare.Value` gets `Boolean bool`, `observations.Observation`/`Input`, `conclusiondetails.Value`, `graphcache` rows) | Same shape as integer; false must be distinguishable from absent | `*bool` — because no sibling field uses pointers |
| D-3 | SQL binds `nil` or `int64` 0/1 through `nullBool(b bool, has bool) any` *(new, in `observations`)*; scans read `sql.NullBool` | Explicit about what is stored in an `INTEGER` STRICT column | Binding a Go `bool` — because it relies on driver conversion |
| D-4 | A boolean Property is single cardinality only. `cardinalityAllowed(valueType, cardinality string) bool` *(new, `properties`)* replaces the bare `cardinalityOK` calls in `Upsert`, `Create`, `Update`; it returns `cardinalityOK(c) && !(valueType == boolean && c == multiple)`. Error: `ErrInvalid` | User, 2026-11-07: a boolean with two answers is a conflict to show, not two kept values | Allowing `multiple` — the user chose single only |
| D-5 | `researcherCreatable` includes boolean in Go now; Swift's `SubjectPropertyValueType.researcherCreatable` stays without it until S14-37. The Go comment says Swift offers boolean from S14-37 | The map puts "researcher-creatable in Go" here and the Properties offer in S14-37 | Adding it to Swift now — because that is S14-37's UI |
| D-6 | Comparer `boolean{}` in `core/valuecompare/boolean.go` *(new)*: equal → `Same`, sim 1, `CauseExact`; unequal → `Different`, `CauseNoMatch`. Key at `Same` and `Equivalent`: `"true"` / `"false"` | A yes/no fact has no nearness; same shape as `term` | Treating false as "no evidence" — because "not literate" is a recorded fact |
| D-7 | `autoreconcile`: boolean uses `keyModule` (with text, integer, depiction); `Carries` reads `HasBoolean`; `SortKey` is `"0"` for false and `"1"` for true | False before true; same module as integer | A new module — because nothing folds |
| D-8 | Negative polarity gets no special case: a negative `true` Observation denies positive `true` candidates as any negative does; it does not create a `false` value | One rule for every type; the composer (S14-37) decides what it offers | Folding "not true" into "false" — because that changes the reconciler for one type |
| D-9 | Proto: `ConclusionValue` `bool boolean = 7;` in the oneof; `Observation` `optional bool value_boolean = 22;`; `ObservationDraft` `optional bool value_boolean = 14;`; the `Property.value_type` comment lists every type including `depiction` and `boolean` | Next free numbers; `optional` and the oneof keep false distinct from unset | Plain `bool` fields — because false would read as absent |
| D-10 | `conclusionValueProto` checks `v.HasBoolean` before `v.HasInteger` | A boolean value has no other field set; order only matters for clarity, so put it next to integer | — |
| D-11 | `valueInput` takes `valueBoolean *bool` as a new parameter after `valueInteger` | Same shape as integer | A separate setter — because both callers already pass every scalar |
| D-12 | `deleteimpact` facet row gains `"value_boolean"` (true/false/nil); `projectObservation`'s title adds no boolean arm, so a boolean Observation's impact row is titled by its Property label alone | Go keeps no English "Yes"/"No"; depiction already titles this way | A Go "yes"/"no" literal — because Go copy is not localized |
| D-13 | The Swift side is wire mapping and value text only (user, 2026-11-07): `PropertyValueType.boolean`; `CatalogConclusionValue.boolean(Bool)`; `CatalogObservation.valueBoolean: Bool? = nil`; `CatalogObservationDraft.valueBoolean: Bool? = nil`; `GoStore` maps both ways; `ReconciledValueDisplay.string` returns "Yes"/"No" | Regenerating the proto makes the Swift switch non-exhaustive; S14-01 set the precedent that the engine PR owns the Swift wire mapping | Mapping the proto case to `.none` until S14-37 — the user chose the full wire mapping |
| D-14 | `GoStore.mapConclusionValue` changes from `private static` to `static` so a Swift test calls it, as `mapObservation` is | Tests the mapping directly | Testing through a fake store — because FakeStore does not map protos |
| D-15 | Copy (user, 2026-11-07): `L10n.Properties.valueTypeBoolean` = "Yes / no"; `L10n.Conclusions.booleanYes` = "Yes", `L10n.Conclusions.booleanNo` = "No" | Researcher wording | "Boolean", "True"/"False" — the user chose "Yes / no" |
| D-16 | The "source surfaces" vocabulary row compares `rawCase` matches in `SourceSurface` with `sorted(SurfacePage, SurfaceGraph, SurfaceCitationComposer, swiftOnlySurfaces...)` where `swiftOnlySurfaces = []string{"promote"}` *(new, in the test file, with a comment: Promote is a Swift-only place Go never opens)*. Add `swiftWorkspaceLocation = "../../macos/App/Features/Workspace/WorkspaceLocation.swift"` to the const block | User, 2026-11-07; exact equality catches drift on both sides | A Go `SurfacePromote` constant — because nothing in Go would use it |
| D-17 | No cache-version or search projection-version bump | No catalog can hold a boolean before this migration (the CHECK forbids it), so no cached row is stale | Bumping `auto_reconciler_meta.cache_version` — because it rebuilds every catalog for nothing |

**Rules this plan bends.** None. The working rule leaves storage to `plan-pr` (D-1). D-4, D-13, D-15 and D-16 are the user's answers of 2026-11-07; D-4 also goes into the working rule (STEP-10).

## 6. Design

### Codebase fit

Searched: `valueTypeOK`, `researcherCreatable`, `cardinalityOK`, `nullInt64`, `comparers`, `keyModule`, `CatalogConclusionValue`, `rawCase`, `Surface*`, `L10n.Properties.valueType*`, `L10n.Conclusions`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `propertymeta.ValueTypeBoolean` | const | `ValueTypeInteger` | Extends the const block |
| `000061.sql` | migration | `000052.sql`, `000053.sql` | New format step; shipped files are not edited |
| `observations.value_boolean`, `auto_reconciler_values.value_boolean` | column | `value_integer` | D-1 |
| `properties.cardinalityAllowed` | func | `cardinalityOK` | Wraps it; one place for the type rule |
| `observations.nullBool` | func | `nullInt64` | Same helper for the new column type |
| `HasBoolean`/`ValueBoolean` fields | struct fields | `HasInteger`/`ValueInteger` | D-2 |
| `valuecompare.boolean` (`boolean.go`) | comparer | `term` | A registry entry, per `add-match-rule` |
| Proto fields (D-9) | proto | `int64 integer = 2`, `value_integer` | Next field numbers |
| `CatalogConclusionValue.boolean`, `valueBoolean` | Swift case / fields | `.integer`, `valueInteger` | D-13 |
| `L10n.Properties.valueTypeBoolean`, `L10n.Conclusions.booleanYes`, `.booleanNo` | keys | `valueTypeInteger`, `Conclusions.likenessValue` | Same namespaces |
| `swiftOnlySurfaces`, `swiftWorkspaceLocation`, "source surfaces" row | test | "value types" row | One more table row |
| `core/valuecompare/boolean_test.go` *(new)* | test file | `integer_test.go` | Mirrors the per-type file |

**Builds on.** `add-match-rule` "Add a comparer" (registry entry, contract row, `TestCompareUnsupported`); `add-catalog-migration` table-rebuild rules; `TestSwiftVocabularyMatchesGo` table.
**Deletes.** The comment "Surface values match macOS SourceSurface raw values." becomes "Surface values match macOS SourceSurface raw values; TestSwiftVocabularyMatchesGo holds them." The sentence "`real` and `boolean` are not used." in `interpretation-layer-data-model.md`. The structural-debt SourceSurface row is struck through as done.
**Next adds.** S14-37 adds `"boolean"` to `SubjectPropertyValueType.researcherCreatable`, a `PropertyValueType.boolean.rawValue` arm in `CitationObservationValue` and `CitationComposerVocabulary.supportedValueTypes`, and an `ObservationValueDisplay` arm reading `valueBoolean`. A later value type repeats this PR's column-per-type threading.

### Reach

| New path or state | Destination | Checked |
| --- | --- | --- |
| A boolean Property in `ListProperties` | `PropertiesModel` / `PropertiesView` | `SubjectPropertyValueType.label("boolean")` now returns "Yes / no" through `PropertyValueType`; the create picker is unchanged (D-5) |
| A boolean value on a Conclusion detail | `ReconciledValueDisplay.string` | Words it "Yes"/"No" (AC-4.2) |
| A boolean Observation on an Evidence graph card | `ObservationValueDisplay` | Shows `valueText`, which is empty for a boolean; S14-37 adds the arm. Reachable only from a catalog created outside the app before S14-37 |

### Scale

None — no new list, loop or write path; each existing write and read carries one more column.

### Edges

- **Assumptions.** `hasOtherScalars` assumes one value column per type: boolean adds `keep != "boolean" && in.HasBoolean` (T-3). `keyModule` assumes `Key` agrees with `Compare` (`TestKeyContract`, T-5). `conclusionValueProto` assumes exactly one field set (T-9).
- **Neighbors.** `false` vs absent: an Observation with `HasBoolean true, ValueBoolean false` stores 0 and is carried; `HasBoolean false` is refused (AC-2.1, AC-2.2). `ConclusionValue{boolean:false}` is a set oneof, not empty (AC-3.5, T-9). Integer 0 and 1 stay integers (existing integer rows in T-3, T-9).
- **Removed inputs.** A deleted boolean Observation removes its candidate; the cache is rewritten by the existing recompute (T-7 recomputes after delete).
- **Real data.** None — no seed or table drives the rule.
- **Untrusted size.** None — one bit per row.

### Schema

Migration `core/database/migrations/000061.sql` *(new)*:

```sql
-- S14-36: value_type = boolean. A yes/no fact: observations.value_boolean
-- holds 0 or 1, and the auto-reconciler cache keeps it beside value_integer.
-- No seed Property is boolean; researchers create them.
-- migrate: foreign_keys off

CREATE TABLE properties__new ( … as 000052.sql, CHECK (value_type IN (
	'text', 'integer', 'date', 'name', 'subject', 'term', 'depiction', 'boolean'
)) ) STRICT;
INSERT INTO properties__new (…) SELECT … FROM properties;
DROP TABLE properties;
ALTER TABLE properties__new RENAME TO properties;

ALTER TABLE observations ADD COLUMN value_boolean INTEGER CHECK (value_boolean IN (0, 1));
ALTER TABLE auto_reconciler_values ADD COLUMN value_boolean INTEGER;
```

`properties` has no index besides its inline `UNIQUE (key, origin)`, which the rebuild keeps. No new table or FK, so no `catalogmodel` registration; `TestPragmaHonesty` must stay green. Delete class: unchanged. Audit: the Observation row maps gain `value_boolean`.

### Go core

| Package | Change |
| --- | --- |
| `core/propertymeta` | `ValueTypeBoolean = "boolean"` *(new)* |
| `core/database/properties` | `valueTypeOK` and `researcherCreatable` include boolean; `cardinalityAllowed` *(new)* used in `Upsert`, `Create`, `Update` (D-4); `Create` doc comment updated |
| `core/database/observations` | `value_boolean` in `sqlInsert`, `sqlGetRow`, `sqlListSelect`, update SQL; `HasBoolean`/`ValueBoolean` on `Observation`, `Input`; `resolvedValue.Boolean *bool` *(new field)*; `resolveValue` case `ValueTypeBoolean` (refuse when `!in.HasBoolean || hasOtherScalars(in, "boolean")`); `hasOtherScalars` boolean line; `scanListed`, `scanRawObservation`, `observationRowMap`, insert change map, update diff `value_boolean`; `nullBool` *(new)* |
| `core/valuecompare` | `Value.Boolean bool`; `boolean{}` in `comparers` (D-6) |
| `core/autoreconcile` | `Value.Boolean`, `Value.HasBoolean`; `Carries`, `SortKey` (D-7), `moduleFor`, `compareValue` |
| `core/database/autoreconciler` | `value_boolean` in `sqlInsert`, `sqlLoadCandidates`, candidate scan, row insert; `truth.go` copies it |
| `core/database/graphcache` | `Boolean`/`HasBoolean` on the row in `graph.go`, `EvidenceObservation`, `verify.go` value and its equality and conversion; `load.go` selects and scans `r.value_boolean` |
| `core/database/graphfill` | `load.go` selects and scans `o.value_boolean` |
| `core/database/promotealign` | `decode` takes `boolean bool, hasBoolean bool` after `hasInteger`; both callers pass them |
| `core/database/conclusiondetails` | `Value.Boolean`/`HasBoolean`; recorded SQL and scan; rank copy |
| `core/database/deleteimpact` | `facets.go` selects `value_boolean`, row map `"value_boolean"` (D-12) |

Errors: `properties.invalid` (`ErrInvalid`) for D-4; `observations` `ErrInvalid` for AC-2.2–2.4. No new codes.

### FFI

`api/proto/engine.proto` per D-9. Handlers: `observations.go` `valueInput(…, valueBoolean *bool, …)`, both callers pass `o.ValueBoolean` / `d.ValueBoolean`; `observationProto` sets `ValueBoolean: optionalBool(o.HasBoolean, o.ValueBoolean)` *(new helper beside `optionalInt64`)*; `source_read.go` `listedFromSource` copies `Boolean`/`HasBoolean`; `conclusion_detail.go` `conclusionValueProto` arm `v.HasBoolean` → `ConclusionValue_Boolean` (D-10). Regenerate with `./scripts/generate-proto.sh` (Go and Swift).

### macOS

No view changes.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Value type case | `PropertyValueType.boolean` + `label` | `macos/App/Platform/InterpretationValueKinds.swift` | `L10n.Properties.valueTypeBoolean` |
| Conclusion value case | `CatalogConclusionValue.boolean(Bool)` | `macos/App/Platform/GenealogyStore.swift` | — |
| Observation fields | `valueBoolean: Bool? = nil` on `CatalogObservation`, `CatalogObservationDraft` | `GenealogyStore.swift` | Defaults keep every call site compiling |
| Mapping | `mapConclusionValue` (`.boolean(let b)` → `.boolean(b)`, now `static`), `mapObservation` (`o.hasValueBoolean ? o.valueBoolean : nil`), `mapCatalogObservation`, `mapObservationDraft` | `macos/App/Platform/GoStore.swift` | — |
| Value text | `.boolean(let b)` → `L10n.string(b ? L10n.Conclusions.booleanYes : L10n.Conclusions.booleanNo)` | `macos/App/Platform/ReconciledValueDisplay.swift` | — |

States: none new. Accessibility: the value text is what VoiceOver reads already.

### Strings

L10n keys *(new)*, with `Localizable.xcstrings` entries per `add-localized-string`:

- `L10n.Properties.valueTypeBoolean` — key `properties.valueType.boolean` = "Yes / no", comment "Property value type label: boolean (a yes/no fact)".
- `L10n.Conclusions.booleanYes` — key `conclusions.value.booleanYes` = "Yes", comment "A boolean Property value that is true".
- `L10n.Conclusions.booleanNo` — key `conclusions.value.booleanNo` = "No", comment "A boolean Property value that is false".

Removed: none.

## 7. Steps

**STEP-1 — Constant and migration**
Files: `core/propertymeta/propertymeta.go`, `core/database/migrations/000061.sql` *(new)*. Skill: `add-catalog-migration`, `add-catalog-model`.
Do: add `ValueTypeBoolean`; write the migration per §6 Schema, copying `000052.sql`'s `properties__new` columns exactly.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes (schema hash, `TestPragmaHonesty`).

**STEP-2 — Property rules**
Files: `core/database/properties/properties.go`, `properties_test.go`.
Do: D-4, D-5. Add T-1, T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/properties/` passes.

**STEP-3 — Observation storage**
Files: `core/database/observations/observations.go`, `update.go`, `write.go`, `observations_test.go`. Skill: `add-catalog-query`.
Do: thread `value_boolean` per §6 Go core (D-2, D-3). Add T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/observations/` passes.

**STEP-4 — Comparer**
Files: `core/valuecompare/valuecompare.go`, `boolean.go` *(new)*, `boolean_test.go` *(new)*, `contract_test.go`, `valuecompare_test.go`. Skill: `add-match-rule` (Add a comparer).
Do: D-6; contract sample row `ofBool(true,true), ofBool(true,false), ofBool(false,false)` and a generator; add boolean to `TestCompareUnsupported`'s want list. Add T-5.
Check: `go test ./core/valuecompare/` passes.

**STEP-5 — Reconcile**
Files: `core/autoreconcile/autoreconcile.go`, `modules.go`, `autoreconcile_test.go`.
Do: D-7, D-8. Add T-6.
Check: `go test ./core/autoreconcile/` passes.

**STEP-6 — Cache and readers**
Files: `core/database/autoreconciler/autoreconciler.go`, `truth.go`, `autoreconciler_test.go`; `core/database/graphcache/graph.go`, `evidence_graph.go`, `load.go`, `verify.go`; `core/database/graphfill/load.go`; `core/database/promotealign/profiles.go`; `core/database/conclusiondetails/conclusiondetails.go`; `core/database/deleteimpact/facets.go`.
Do: thread `value_boolean` per §6 Go core (D-12, D-17). Add T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/...` passes.

**STEP-7 — Proto and handlers**
Files: `api/proto/engine.proto`, generated Go and Swift, `api/ffi/handlers/observations.go`, `source_read.go`, `conclusion_detail.go`, `conclusion_detail_test.go`, `citations_test.go`. Skill: `add-ffi-handler`.
Do: D-9 – D-11; run `./scripts/generate-proto.sh`. Add T-8, T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes (Swift does not compile yet: STEP-8).

**STEP-8 — Swift wire mapping and text**
Files: `InterpretationValueKinds.swift`, `GenealogyStore.swift`, `GoStore.swift`, `ReconciledValueDisplay.swift`, `L10n.swift`, `Localizable.xcstrings`, `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift`, `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-13 – D-15, §6 macOS and Strings. Add T-10, T-11.
Check: Xcode tests pass with no Swift warning (`scripts/check-pr.sh` Xcode step).

**STEP-9 — Vocabulary pins**
Files: `api/ffi/swift_vocabulary_test.go`, `core/workspacelocation/workspacelocation.go` (comment).
Do: add `propertymeta.ValueTypeBoolean` to the "value types" want list; D-16 row. Add T-12.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; temporarily renaming Swift `citationComposer` fails it (revert).

**STEP-10 — Docs**
Files: `docs/interpretation-layer-data-model.md` (value-type list ~line 613–635: add `boolean`, single only, researcher-creatable; drop "`real` and `boolean` are not used"; `value_boolean` in the observations sketch ~line 880), `docs/conclusion-layer-data-model.md` (`value_boolean` in the cache sketch ~line 362), `docs/seeded-vocabulary.md` §3.2 (`boolean` is a valid `value_type` with no seed row), `docs/ideas/structural-debt.md` (SourceSurface row → done by S14-36), `../deployment-plan.md` (map row; append "Boolean Properties are single cardinality only (user, 2026-11-07)." to the Batch 7 boolean working rule), `../completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-36-boolean-value-type-engine.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/properties/properties_test.go` `TestCreateBooleanProperty` *(new)*, table-driven | boolean + "" → single; boolean + multiple → `ErrInvalid`; integer + multiple → ok | AC-1.1, AC-1.2, D-4, D-5 |
| T-2 | Go | same file `TestBooleanCardinalityStaysSingle` *(new)*, table-driven | `Update` to multiple → `ErrInvalid`; `Upsert` boolean multiple → `ErrInvalid` | AC-1.3, AC-1.4 |
| T-3 | Go | `core/database/observations/observations_test.go` `TestBooleanObservationValue` *(new)*, table-driven | false stored/listed/raw-read as `HasBoolean, false`; true likewise; missing → `ErrInvalid`; boolean + text → `ErrInvalid`; text Property + boolean → `ErrInvalid` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-3 |
| T-4 | Go | same file `TestUpdateBooleanObservation` *(new)* | false → true: row 1, diff `{false, true}`; unchanged → no change | AC-2.5 |
| T-5 | Go | `core/valuecompare/boolean_test.go` `TestBoolean` *(new)*, table-driven; `TestComparerContract`, `TestKeyContract`, `TestCompareUnsupported` | t/t Same exact; t/f Different no_match; keys "true"/"false" | AC-3.1, AC-3.2, D-6 |
| T-6 | Go | `core/autoreconcile/autoreconcile_test.go` `TestReconcile` rows *(new)* + `TestSortKey` rows | two Sources true → merged support 2; true vs false single → mixed; negative true denies true; sort keys "0" < "1" | AC-3.3, AC-3.4, D-7, D-8 |
| T-7 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestReconciledBoolean` *(new)* | catalog: boolean Property, two members true → cache row `value_boolean = 1`, support 2; delete one Observation → support 1; `graphcache` verify passes | AC-3.3, D-17 |
| T-8 | Go | `api/ffi/handlers/citations_test.go` `TestCreateCitationBooleanObservation` *(new)* | draft `value_boolean = false` → response `GetValueBoolean() == false` and `ValueBoolean != nil` | AC-2.6, D-9, D-11 |
| T-9 | Go | `api/ffi/handlers/conclusion_detail_test.go` `TestConclusionValueProtoCoversEveryKind` rows *(new)* | `Value{HasBoolean, Boolean false}` → `ConclusionValue_Boolean` false; true likewise | AC-3.5, D-10 |
| T-10 | Swift | `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift` `aBooleanReadsAsYesOrNo` *(new)*; `macos/ProvenenciaTests/PropertiesModelTests.swift` `booleanLabelReadsYesNo` *(new)* | `.boolean(true)` → "Yes", `.boolean(false)` → "No"; `PropertyValueType(rawValue:"boolean")?.label` → "Yes / no"; `researcherCreatable` does not contain "boolean" | AC-4.2, AC-4.4, D-5, D-15 |
| T-11 | Swift | `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift` `booleanValuesMapBothWays` *(new)* | proto `boolean=false` → `.boolean(false)`; Observation `valueBoolean=false` round-trips; absent → nil | AC-4.1, AC-4.3, D-14 |
| T-12 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` "value types", "source surfaces" *(new row)* | Swift = Go (+ `promote`) | AC-5.1, AC-5.2, D-16 |
| — | Go | existing suites via the schema hash | migrated catalog opens | AC-1.5 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestComparerContract`, `TestCompareUnsupported`, `TestPragmaHonesty`, `graphcache` verify tests, `TestSwiftVocabularyMatchesGo`, `scripts/check-localizable-xcstrings.py`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief; no view change)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the single-cardinality line in the boolean working rule; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain docs updated: interpretation-layer-data-model.md, conclusion-layer-data-model.md, seeded-vocabulary.md §3.2; ideas/structural-debt.md SourceSurface row
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Booleans cannot be created in the app until S14-37, so the check is through tests and an existing project:

1. Open an existing project made before this branch. It opens; the Properties list and a Person page look and read as before (AC-1.5).
2. Create a new project. It opens; the Properties create picker offers the same five types as before (D-5).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A reader of `observations` or `auto_reconciler_values` misses the new column | §4 lists every site found by `rg value_integer` / `HasInteger`; STEP-6 check runs all of `./core/...`; `graphcache` verify compares cache with truth |
| The `properties` rebuild breaks an FK | `-- migrate: foreign_keys off` runs `foreign_key_check` before commit |
| Generated Swift adds a case another switch misses | The compiler reports it; §4 lists the two known switches |

**Stop and report if:**

- `rg -n "value_integer|HasInteger" --type go` at your base finds a non-test site not listed in §4.
- `000061.sql` is taken on `main` when you start, or a later migration than `000060.sql` recreated `properties`.
- `TestPragmaHonesty` or `verifySchema` fails after the migration and the fix is anything other than the migration's own DDL.
- A Swift `switch` over `CatalogConclusionValue` or `PropertyValueType` outside §6 macOS fails to compile (its arm is S14-37's view work).
- `TestComparerContract` needs a change to the contract itself, not a new sample row and generator.

## 12. Open questions

None.

## 13. PR

**Title:** S14-36 — Boolean value type (engine)
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-36-boolean-value-type-engine.md`.
