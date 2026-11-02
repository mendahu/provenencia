# S14-17 — Vocabulary writes return their effects; dead RPCs go

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes (batch 5; was Spike 12's S12-33)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: every view draws the same rows and copy; only what reloads after a vocabulary write changes, and the Source types and Metadata header lines count the rows the page already shows)
**Board:** —
**Depends on:** S14-16 (branch `s14-16-promote-source-effects`, PR [#686](https://github.com/mendahu/provenencia/pull/686), open at planning)
**Blocks:** S14-27
**Branch:** `s14-17-vocabulary-effects`
**Version bump:** patch — a refresh fix and dead-code removal, no feature ([`docs/versioning.md`](../../../versioning.md)); the next patch after `VERSION` at merge (0.2.10 → 0.2.11 at planning)
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [`app-health-audit.md`](../app-health-audit.md) A-29, A-35; working rules "Every audited write returns its effects, and a test holds them to it", "FakeStore is a seeded fake, not a second engine"; [`macos-client-patterns.md`](../../../macos-client-patterns.md) cache contract rule 2 and "Conclusion keys"; [`core/database/effects/README.md`](../../../../core/database/effects/README.md) (`Set.Lists`); [`catalog-deletes.md`](../../../catalog-deletes.md) (the type–field join is never audited); S14-16's plan **Next adds** ([`S14-16-promote-source-effects.md`](S14-16-promote-source-effects.md))
**Planned:** 2026-11-02 against `s14-16-promote-source-effects` @ `a5e057b` (S14-16's head; `main` @ `04a2c99` plus S14-16)
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

Every vocabulary write tells the app what it changed, the way every other audited write does after S14-16. Creating, renaming or deleting a Source type, a metadata field, a Property or a Property term, and binding or unbinding a Property on a Subject type, return `WriteEffects`. The app reloads from those effects and no longer from a hand-named mutation per write. Eleven `CatalogMutation` cases and `CatalogQueryRegistry.conclusionTriggers` go. `TestEveryAuditedWriteReturnsEffects` runs with no allow-list. The two unaudited type–field join writes stop sending a `user_id`, so the test does not count them. The vocabulary counts nothing draws go: `CatalogCounts` keeps only the sidebar's Research badges, and the Source types and Metadata header lines count the rows their page shows. Eight RPCs with no Swift caller go, with their handlers, tests and Swift store methods, and their method numbers are reserved.

**Why now.** S14-16 added the pin test with sixteen vocabulary writes on its allow-list (`effectsPending`) and left `.updatedProperty` in `conclusionTriggers`; this PR empties both (working rule, map review K-10). Audit A-35 lists the unclaimed RPCs. The user ruled on 2026-11-02 that the vocabulary counts nobody draws are removed, not moved onto the recount, and that the type–field join stays unaudited and drops its `user_id`.

**What stays.** What every view draws. `.assignedTypeSuggestion(typeId:)` and `.removedTypeSuggestion(typeId:)` stay: the join they name is not audited, so it has no effects (D-2). `.createdSource`, `.updatedSource`, `.changedSourceType`, `.deletedSource`, `.mutatedSourceWorkspace`, `.mutatedSourceMetadata` and `.deletedSubject` stay. The models' `patchTypesList`, `patchSuggestions` and other list patches after a write stay. `UpdatePropertyTerm`, `DeletePropertyTerm` and `ReorderSubjectTypeProperties` stay (named by [`term-management.md`](../../../ideas/term-management.md) and [`binding-order-editing.md`](../../../ideas/binding-order-editing.md)); the first two now return effects, and the third already does. The Go effects registry (`core/database/effects/registry.go`) is unchanged. The sidebar recount (`CatalogCounts.refreshAll`) and its `counts` trigger are unchanged.

## 2. User stories

**US-1.** As a researcher, when I rename a Property, a Source type or a metadata field, every page that shows the label shows the new one the next time I look at it, because the rename reports what it changed.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A warmed Person page (`conclusionDetail` for `e1`) and the warmed Persons list; on the Properties page the researcher renames the user Property "Occupation" to "Trade" | The rename's effects carry `lists: ["vocabulary"]`; the hidden `conclusionDetail` key is evicted and the Persons list is marked stale, so opening the Person shows "Trade" |
| AC-1.2 | A visited Source page (`sourceWorkspace` for `s1`) and a warmed `typeSuggestions` for type `t1`; on Metadata the researcher renames the field "Volume" to "Book" | `metadataFieldsList` reloads; the `sourceWorkspace` key is marked stale and every warmed `typeSuggestions` key reloads |
| AC-1.3 | The researcher adds a Source type "Diary" on Source types | The Source types list reloads from the write's effects; the new row shows, selected, as today |
| AC-1.4 | In the citation composer the researcher adds a custom term "Cooper" to Occupation | Every warmed `propertyTerms` key reloads and the combo offers "Cooper", as today |
| AC-1.5 | The researcher sets a user Property's Holds from single to multiple, and one Person observes it | The effects name that Person's handle in `handles` and carry `vocabulary`; the Person's page reloads its values |
| AC-1.6 | The researcher assigns the field "Volume" to type `t1` | The `typeSuggestions` key for `t1` reloads and every visited Source page is marked stale, as today (`.assignedTypeSuggestion(typeId:)`) |

**US-2.** As the engine, every request with a `user_id` returns `WriteEffects`, and no write is excused.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `TestEveryAuditedWriteReturnsEffects` on `engine.File_engine_proto` | Passes, and the file has no `effectsPending` map |
| AC-2.2 | `CreateSourceType`, `UpdateSourceType` (label changed), `DeleteSourceType`, `CreateMetadataField`, `UpdateMetadataField` (label changed), `DeleteMetadataField`, `CreateProperty`, `UpdateProperty` (label changed), `DeleteProperty`, `CreatePropertyTerm`, `UpdatePropertyTerm` (label changed), `DeletePropertyTerm`, `AssignSubjectTypeProperty`, `RemoveSubjectTypeProperty` | Each response's `effects.lists` is `["vocabulary"]` and `effects.revision` > 0 |
| AC-2.3 | `AssignTypeFieldRequest` and `RemoveTypeFieldRequest` | Have no `user_id` field; number 2 and the name are reserved |

**US-3.** As a contributor, every RPC in `engine.proto` has a Swift caller, except the three an idea file names.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `rg` for each of `EnsureFileThumbnail`, `GetSubjectTypePresentation`, `ListPlaceableSubjectTypes`, `ListProperties`, `ListSubjectTypeProperties`, `RemoveInstallIdentity`, `ListSubjectTypes`, `ClearSubjectPosition` as a handler or method name | No hit in `api/`, `core/` or `macos/App/` outside the generated files' reserved lines; each `METHOD_…` number and name is reserved |
| AC-3.2 | `rg -n 'listSubjectTypes|clearSubjectPosition' macos` | No hit |

**US-4.** As a researcher, the Source types and Metadata headers count the rows they show.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | The Source types page lists 3 seeded and 2 user types | The header reads the 5 · 3 seeded · 2 yours line from those rows, with no `CatalogCounts` |
| AC-4.2 | The Metadata page lists 4 seeded fields and 1 plugin field | The header reads the line with the plugin count from those rows |
| AC-4.3 | The sidebar after any write | Shows badges for Sources, Persons, Events and Places as today, and none on Configure rows |

## 3. Scope

| In | Covered by |
| --- | --- |
| Fourteen vocabulary handlers return `WriteEffects` (D-1) | STEP-1, AC-2.2 |
| `effectsPending` and `missingEffects`'s `pending` parameter go (D-3) | STEP-1, AC-2.1 |
| `AssignTypeField` / `RemoveTypeField` drop `user_id` (D-2) | STEP-1, AC-2.3, AC-1.6 |
| `.vocabulary` reaches every key a vocabulary label feeds (D-5) | STEP-2, AC-1.1, AC-1.2, AC-1.4 |
| Swift store vocabulary writes return `CatalogWriteEffects`; callers apply them (D-4, D-7) | STEP-3, AC-1.1–AC-1.5 |
| Eleven `CatalogMutation` cases and `conclusionTriggers` deleted (D-6) | STEP-3, AC-1.1 |
| Vocabulary counts removed: `CatalogCounts`, the models' publish hooks, `GetWorkspaceNavCounts`'s vocabulary fields, `CountByOrigin` (D-8) | STEP-4, AC-4.1–AC-4.3 |
| Eight RPCs with no Swift caller deleted, numbers reserved; `listSubjectTypes` and `clearSubjectPosition` deleted from the Swift store (D-9, D-10, D-11) | STEP-5, AC-3.1, AC-3.2 |
| `macos-client-patterns.md` rule 2 and "Conclusion keys", `add-workspace-place`, `add-ffi-handler` (D-13) | STEP-6 |

| Out | Home |
| --- | --- |
| Deleting `UpdatePropertyTerm`, `DeletePropertyTerm`, `ReorderSubjectTypeProperties` | not planned ([`term-management.md`](../../../ideas/term-management.md), [`binding-order-editing.md`](../../../ideas/binding-order-editing.md)) |
| Moving `lists` / `structure` onto descriptor `follows` slots; deleting the `invalidations(for list:)` switch | S14-27 |
| Splitting `effects.ListVocabulary` per table, or keyed reads for `typeSuggestions` / `propertyTerms` (D-5 Rejected) | not planned |
| A visited canvas (`evidenceGraph`) and `citationsByArtifact` keep a Property's old label after a rename: their rows carry `propertyLabel`, and no write reloads them for it today (Edges) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md), a line STEP-7 adds |
| Auditing the type–field join (`source_type_metadata_fields`) | not planned (user, 2026-11-02) |
| `.deletedSubject` on `propertiesWorkspace`, now that Subject deletes report `property_usage` | not planned in this PR; S14-27 reads every remaining `invalidateOn` |
| The audit's test-net items (L-16) | not planned (map scope entry) |
| A sidebar recount that fails after a write is silent (S14-16 UX-1) | S14-16's follow-up |

## 4. Current state

Verified on `s14-16-promote-source-effects` @ `a5e057b`.

| Area | Where | Today |
| --- | --- | --- |
| Pin test | `api/ffi/handlers/write_effects_test.go` `effectsPending`, `missingEffects`, `TestEveryAuditedWriteReturnsEffects` | Sixteen `…Request` names map to `"S14-17"`; `missingEffects(file, pending)` skips them and reports a stale entry; five fixture subtests, three of which test `pending` |
| Reserved fields test | `api/ffi/handlers/write_effects_test.go` `TestRetiredFieldsStayReserved` | Pins S14-16's reserved fields and `METHOD_PROMOTE_SUBJECT` (85) |
| Effects mapping | `api/ffi/handlers/write_effects.go` `writeEffectsProto(res writes.Result)` | The one mapper every handler uses |
| Source type / field handlers | `api/ffi/handlers/source_defs.go` `CreateSourceType`, `UpdateSourceType`, `DeleteSourceType`, `CreateMetadataField`, `UpdateMetadataField`, `DeleteMetadataField` | Run through `writes.Run` and discard the `writes.Result` (`got, _, err :=`); responses carry no effects |
| Type–field join handlers | `api/ffi/handlers/source_defs.go` `AssignTypeField`, `RemoveTypeField`, `parseSuggestionJoin(userID, typeID, fieldID)` | Validate `user_id` and never use it; write through `sourcevocab.AppendSuggestion(c, …)` / `DeleteSuggestion(c, …)` (`core/database/sourcevocab/sourcevocab.go`), a plain `db.Exec` outside `writes.Run`, unaudited; the table is `None: true` in the effects registry; `docs/catalog-deletes.md` calls it "a vocab join, a UI hint, never audited" |
| Property / binding handlers | `api/ffi/handlers/subject_defs.go` `CreateProperty`, `UpdateProperty`, `DeleteProperty`, `AssignSubjectTypeProperty`, `RemoveSubjectTypeProperty` | Discard the `writes.Result`; `ReorderSubjectTypeProperties` already returns `Effects: writeEffectsProto(wrote)` (the precedent in the same file) |
| Term handlers | `api/ffi/handlers/property_terms.go` `CreatePropertyTerm`, `UpdatePropertyTerm`, `DeletePropertyTerm` | Discard the `writes.Result` |
| Registry | `core/database/effects/registry.go` entries `source_types`, `source_metadata_fields`, `subject_types`, `properties`, `property_terms`, `subject_type_properties` | Each sets `Lists: []List{ListVocabulary}`; `properties` adds `Handles: onField("cardinality", …)`; `property_terms` adds `Handles` and `Structure` on `neutral` / `directed` / `inverse_key`; `source_type_metadata_fields` is `None` |
| Go display cache | `core/database/graphcache/commit.go` `OnCommit` | `set.Has(effects.ListVocabulary)` drops every cached label and vocabulary memo: Go already treats `vocabulary` as "any label may have changed" |
| Dead RPC handlers | `api/ffi/handlers/thumbnails.go` `EnsureFileThumbnail`; `subject_defs.go` `ListProperties`, `ListSubjectTypeProperties`, `ListPlaceableSubjectTypes`, `GetSubjectTypePresentation`; `identity.go` `RemoveInstallIdentity`; `subjects.go` `ListSubjectTypes`, `ClearSubjectPosition` | No Swift caller (`rg` in `macos/App`), except `ListSubjectTypes` and `ClearSubjectPosition`, whose Swift store methods only `macos/ProvenenciaTests/SubjectStoreTests.swift` calls; each has a `dispatch.go` constant and case (`api/ffi/dispatch.go`) and a `METHOD_…` value in `api/proto/engine.proto` (5, 41, 45, 51, 53, 57, 60, 61) |
| Go test fixtures on dead RPCs | `api/ffi/handlers/property_terms_test.go` `propertyIDByKey`; `ListProperties` / `ListSubjectTypes` calls in `citations_test.go`, `connect_test.go`, `conclusion_detail_test.go`, `person_detail_test.go`, `promote_test.go`, `subjects_test.go`, `subject_defs_test.go` | Read seeded Property and Subject type ids through the RPCs this PR deletes; `GetPropertiesWorkspaceResponse` carries the same `properties` and `types` |
| Callees of dead RPCs | `core/database/subjectvocab/subjectvocab.go` `PlaceableTypes`; `core/database/subjectpositions/subjectpositions.go` `Clear`, `sqlClear` | Only the deleted handlers call them outside tests; `core/database/connect/connect_test.go` ("missing position invalid") and `subjectpositions_test.go` call `Clear`; `subjectvocab_test.go` ("registry placeable presentation connect") calls `PlaceableTypes` |
| Kept callees | `identity.Remove` (`core/identity/identity.go`), `thumbnailPass.need` | Still called by `api/ffi/handlers/session.go` (sign-out) and by `artifacts.go` / `sources.go` |
| Nav counts | `api/ffi/handlers/workspace_counts.go` `GetWorkspaceNavCounts`, `typeOriginProto`, `fieldOriginProto`; proto `GetWorkspaceNavCountsResponse.source_types` (2), `.metadata_fields` (3), `VocabularyOriginCounts` | Count Source types and metadata fields by origin on every recount; `sourcetypes.CountByOrigin`, `metadatafields.CountByOrigin` and their `OriginCounts`; `properties.CountByOrigin` / `OriginCounts` have no caller at all |
| Swift store | `macos/App/Platform/GenealogyStore.swift` (`createSourceType` … `removeSubjectTypeProperty`, `listSubjectTypes`, `clearSubjectPosition`, `WorkspaceNavCounts`, `WorkspaceNavOriginCounts`); `GoStore.swift` (same methods, `mapVocabularyOriginCounts`-style mapper at the `Provenencia_Engine_V1_VocabularyOriginCounts` parameter); `FakeStore.swift` (same methods, `listSubjectTypesCalls`, `originCounts(from:)`); `CoreInvoke.swift` (`listSubjectTypes`, `clearSubjectPosition`) | The twelve vocabulary writes Swift calls return a row, a list or nothing; no effects. S14-16's Source writes return `(Row, CatalogWriteEffects)` or `CatalogWriteEffects` and `FakeStore` returns `nextWriteEffects` (the precedent) |
| Mutations | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogMutation`, `CatalogMutationKind`, `invalidationKind` | Thirteen vocabulary cases: `createdSourceType`, `updatedSourceType(id:)`, `deletedSourceType(id:)`, `createdMetadataField`, `updatedMetadataField(id:)`, `deletedMetadataField(id:)`, `assignedTypeSuggestion(typeId:)`, `removedTypeSuggestion(typeId:)`, `createdProperty`, `updatedProperty`, `deletedProperty`, `mutatedSubjectTypeProperties`, `createdPropertyTerm(propertyId:)` |
| Registry | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `conclusionTriggers`, `invalidations(for list:project:)`, descriptors `sourceTypesList`, `metadataFieldsList`, `sourceWorkspace`, `typeSuggestions`, `propertiesWorkspace`, `propertyTerms`, the seven Conclusion kinds | `conclusionTriggers == [.updatedProperty]`; `.vocabulary` and `.propertyUsage` both return only `.key(.propertiesWorkspace)`; each vocabulary key lists its mutation kinds in `invalidateOn` |
| Callers | `macos/App/Features/SourceTypes/SourceTypesModel.swift` (`submit` adding / editing, `assignPickedField`, `removeSuggestion`, `confirmPendingImpact`), `macos/App/Features/Metadata/MetadataModel.swift` (delete, create, update), `macos/App/Features/Properties/PropertiesModel.swift` (`submitEdit`, `setCardinality`, delete, create + assign, remove, assign), `macos/App/Features/CitationComposer/CitationComposerModel.swift` `confirmCustomTerm` | Each applies its mutation with `session.apply(.…)` |
| Vocabulary counts | `macos/App/Features/Catalog/CatalogCounts.swift` `metadata`, `sourceTypes`, `properties`, `publishMetadata`, `publishSourceTypes`, `publishProperties`, `badge(for:)`, `CatalogCountSummary.init(_ counts: WorkspaceNavOriginCounts)`; `SourceTypesModel.publishCounts()`, `countLine`; `MetadataModel.syncCatalogCounts()`, `countLine`; `PropertiesModel.syncCatalogCounts()`; `MetadataView` / `PropertiesView` `.onSettled(…) { model.syncCatalogCounts() }`; `catalogCounts:` parameters on the three models and views, `WorkspaceDestinationHost.catalogCounts`, `WorkspaceContent.catalogCounts`, `WorkspaceView` `.environment(catalogCounts)` | `WorkspaceSidebarSections.make` draws badges only for Sources, Persons, Events, Places (`count: nil` on Configure rows). `properties` is written and never read outside tests. `sourceTypes` / `metadata` are read only by the two `countLine`s, which fall back to `.from(types)` / `.from(fields)`. No view reads `CatalogCounts` from the environment |
| Docs | `docs/macos-client-patterns.md` rule 2 and "Conclusion keys"; `.agents/skills/add-workspace-place/SKILL.md` checklist line 41, "Sidebar badges" (line 156), the "writes that still report none" clause (line 174); `.agents/skills/add-ffi-handler/SKILL.md` line 57 | Name `.updatedProperty` / `conclusionTriggers` "until S14-17" and the vocabulary `publish*` path |
| Tests | `api/ffi/handlers/source_defs_test.go`, `subject_defs_test.go`, `property_terms_test.go`, `workspace_counts_test.go`, `identity_test.go` `TestRemoveInstallIdentity`, `thumbnails_test.go` `TestEnsureFileThumbnail`, `subjects_test.go` `TestListSubjectTypes`, `TestClearSubjectPositionBadProto`; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` (`blanketTriggersAreGone`, `applyFieldCRUDInvalidatesMetadataList`, `applyTypeCRUDInvalidatesTypesList`, `listCases`, the `.updatedProperty` assertions near lines 916–995), `WorkspaceSessionTests.swift` `aKeyInvalidatedAndRefreshedRevalidatesOnce`, `SourceTypesModelTests.swift`, `MetadataModelTests.swift`, `PropertiesModelTests.swift` (`createPublishesTheCountOnceTheReloadSettles`, `deletePublishesTheCountOnceTheReloadSettles`), `CatalogCountsTests.swift` (`publishReplacesVocabularySummariesWithoutStoreRoundTrip`, `refreshAllPopulatesEveryDestination`), `SubjectStoreTests.swift` (`setPositionPersistsAcrossSessionCloseThenClear`, `listSubjectTypesReturnsSeededRows`), `OnboardingModelTests.swift` (its store stub), `CitationComposerModelTests.swift` | Cover today's mutation-driven behavior |

**Precedent.** S14-16 (PR [#686](https://github.com/mendahu/provenencia/pull/686)): `823f243` added `WriteEffects effects` to the Source write responses and `Effects: writeEffectsProto(wrote)` in their handlers; `37fb20f` changed the Swift store to `(Row, CatalogWriteEffects)` and the callers to `session.apply(effects, …)`; `794f797` maps every response's effects with `Self.mapWriteEffects(resp.effects)` (no `hasEffects` guard). Copy those three shapes. In the same file as the vocabulary handlers, `ReorderSubjectTypeProperties` already does it (`_, wrote, err := writes.Run(…)`). The method reservation copies S14-16's `reserved 85; reserved "METHOD_PROMOTE_SUBJECT";` and its `TestRetiredFieldsStayReserved` lines. Do not copy S14-16's `effectsPending` (this PR deletes it).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The fourteen audited vocabulary handlers keep the `writes.Result` (`got, wrote, err :=` / `_, wrote, err :=`) and set `Effects: writeEffectsProto(wrote)` on their response. Proto: `WriteEffects effects = 2;` on `CreateSourceTypeResponse`, `UpdateSourceTypeResponse`, `CreateMetadataFieldResponse`, `UpdateMetadataFieldResponse`, `CreatePropertyResponse`, `UpdatePropertyResponse`, `CreatePropertyTermResponse`, `UpdatePropertyTermResponse`; `WriteEffects effects = 1;` on `DeleteSourceTypeResponse`, `DeleteMetadataFieldResponse`, `DeletePropertyResponse`, `DeletePropertyTermResponse`, `AssignSubjectTypePropertyResponse`, `RemoveSubjectTypePropertyResponse`. The effects registry is not changed | Working rule; each write already resolves its effects in `writes.Run`; the registry already gives every vocabulary table `ListVocabulary` | A hand-built `WriteEffects{Lists: ["vocabulary"]}` in handlers — a second source of truth beside the registry |
| D-2 | `AssignTypeFieldRequest` and `RemoveTypeFieldRequest` reserve `user_id`: `reserved 2; reserved "user_id";`. `parseSuggestionJoin(typeID, fieldID string)` loses its user parameter. Swift `assignTypeField` / `removeTypeField` lose `userID:` in `GenealogyStore`, `GoStore`, `FakeStore`, the `OnboardingModelTests` stub and `SourceTypesModel`. The writes stay outside `writes.Run`, and `SourceTypesModel` keeps applying `.assignedTypeSuggestion(typeId:)` / `.removedTypeSuggestion(typeId:)` | User, 2026-11-02: the join stays a UI hint, never audited (`catalog-deletes.md`). The pin test treats `user_id` as "audited"; a request that is not audited does not carry one. `add-ffi-handler` already says an unaudited write has no effects to return | Auditing the join through `writes.Run` (refused by the user); keeping `user_id` with a test exception — an allow-list under another name |
| D-3 | `effectsPending` is deleted. `missingEffects(file protoreflect.FileDescriptor) []string` drops its `pending` parameter and the two pending branches. `TestEveryAuditedWriteReturnsEffects` keeps the two fixture subtests "a response without effects fails naming the request" and "a response with effects passes" and deletes the three pending subtests | S14-16 **Next adds**: when the map is empty it goes; a parameter with one caller passing nil is dead | Keeping an empty map "for later" — the working rule wants a write to land with its effects, not an excuse slot |
| D-4 | Swift store signatures: `createSourceType`, `updateSourceType` → `(CatalogSourceType, CatalogWriteEffects)`; `deleteSourceType` → `CatalogWriteEffects`; `createMetadataField`, `updateMetadataField` → `(CatalogMetadataField, CatalogWriteEffects)`; `deleteMetadataField` → `CatalogWriteEffects`; `createProperty`, `updateProperty` → `(CatalogProperty, CatalogWriteEffects)`; `deleteProperty` → `CatalogWriteEffects`; `createPropertyTerm` → `(CatalogPropertyTerm, CatalogWriteEffects)`; `assignSubjectTypeProperty`, `removeSubjectTypeProperty` → `CatalogWriteEffects`. `GoStore` maps with `Self.mapWriteEffects(resp.effects)`. `FakeStore` returns `nextWriteEffects` from each (S14-16's seeded-fake shape); the `OnboardingModelTests` stub returns `.empty` for the effects-only methods and keeps throwing for the rest. `UpdatePropertyTerm` and `DeletePropertyTerm` get no Swift method | S14-16 shape (`(CatalogSource, CatalogWriteEffects)`); working rule "FakeStore is a seeded fake, not a second engine" | A `FakeStore` that builds `.vocabulary` effects itself — a second engine |
| D-5 | `CatalogQueryRegistry.invalidations(for list:project:)` splits `.vocabulary` from `.propertyUsage`. `.propertyUsage` keeps `[.key(.propertiesWorkspace(project:))]`. `.vocabulary` returns, in this order: `.key(.sourceTypesList)`, `.key(.metadataFieldsList)`, `.key(.propertiesWorkspace)`, `.allCached(.typeSuggestions)`, `.allCached(.sourceWorkspace)`, `.allCached(.propertyTerms)`, `.key(.personsList)`, `.key(.eventsList)`, `.key(.placesList)`, `.allCached(.conclusionDetail)`, `.allCached(.personClaims)`, `.allCached(.personFamily)`, `.allCached(.personEvents)` (each with `project:`). The doc comment on `CatalogList.vocabulary` says: a vocabulary label or binding changed; every key whose payload holds a vocabulary label reloads, as Go's `graphcache` drops every label. That set is exactly the union of what the eleven deleted mutations reached today | Go already reads `vocabulary` as "any label may have changed" (`graphcache.OnCommit`). Vocabulary writes are rare and, except a composer term, made on a Configure page where none of these keys is on screen: the hidden lists mark stale, details evict, and only the small vocabulary lists revalidate (Scale). The union keeps every reload a researcher sees today | Splitting `effects.ListVocabulary` into per-table lists — changes the wire, `graphcache` and S14-27's input for writes made a few times a session; keyed reads (`Read`) for `typeSuggestions` and `propertyTerms` — two new key tables to save reloading lists of tens of rows |
| D-6 | Deleted: `CatalogMutation` cases `createdSourceType`, `updatedSourceType(id:)`, `deletedSourceType(id:)`, `createdMetadataField`, `updatedMetadataField(id:)`, `deletedMetadataField(id:)`, `createdProperty`, `updatedProperty`, `deletedProperty`, `mutatedSubjectTypeProperties`, `createdPropertyTerm(propertyId:)`, their `CatalogMutationKind` cases and `invalidationKind` arms; `CatalogQueryRegistry.conclusionTriggers` and its doc. New `invalidateOn` sets: `sourceTypesList` `[.createdSource, .deletedSource, .changedSourceType]`; `metadataFieldsList` `[.mutatedSourceMetadata]`; `sourceWorkspace` `[.mutatedSourceWorkspace, .mutatedSourceMetadata, .changedSourceType, .assignedTypeSuggestion, .removedTypeSuggestion, .deletedSource]`; `typeSuggestions` `[.assignedTypeSuggestion, .removedTypeSuggestion]` (its closure unchanged); `propertiesWorkspace` `[.deletedSubject]`; `propertyTerms` `[]` with `invalidation: { _, _ in .allCached(.propertyTerms) }`; the seven Conclusion kinds `[]`. Each descriptor comment that names a deleted mutation says the key follows `vocabulary` instead | Each deleted case's keys are in D-5's set; nothing else applies them (STEP-3 check) | Keeping the cases beside the effects — two paths to one reload, the thing rule 2's exceptions list is meant to shrink |
| D-7 | Callers apply the write's effects where they applied the mutation, and nothing else changes: `session.apply(effects)` in `SourceTypesModel` (add, edit, delete), `MetadataModel` (add, edit, delete), `PropertiesModel` (`submitEdit`, `setCardinality`, delete, create, each assign and remove), `CitationComposerModel.confirmCustomTerm` (then `await warmTerms(for:)` as today). `PropertiesModel`'s create-then-assign applies the create's effects, then the assign's, as it applied two mutations. Patches after `apply` stay | Rule 2: `apply(effects)` for a write that returns effects | Merging the create and assign effects into one `apply` — changes when the first reload starts, outside this PR |
| D-8 | Vocabulary counts go (user, 2026-11-02: "We don't use those counts, so if they're not in use, let's remove them"). `CatalogCounts` loses `metadata`, `sourceTypes`, `properties`, `publishMetadata`, `publishSourceTypes`, `publishProperties`; `badge(for:)` returns `nil` for `.sourceTypes`, `.metadata`, `.properties`; `refreshAll` assigns only sources and handles; its doc drops the `publish*` sentence. `CatalogCountSummary` stays with `from(_:)`; its `init(_ counts: WorkspaceNavOriginCounts)` goes. `SourceTypesModel.countLine` uses `.from(types)`, `MetadataModel.countLine` uses `.from(fields)`. `SourceTypesModel.publishCounts()` and its calls, `MetadataModel.syncCatalogCounts()`, `PropertiesModel.syncCatalogCounts()`, the two `.onSettled(…) { model.syncCatalogCounts() }` modifiers, the `catalogCounts` property and init parameter of the three models and views, `WorkspaceDestinationHost.catalogCounts`, `WorkspaceContent.catalogCounts` and `WorkspaceView`'s `.environment(catalogCounts)` go; `WorkspaceView` keeps the `@State` and passes it to `WorkspaceSidebar` and the recount. Go: `GetWorkspaceNavCountsResponse` reserves 2, 3 and `"source_types", "metadata_fields"`; `message VocabularyOriginCounts` is deleted; `typeOriginProto`, `fieldOriginProto`, `sourcetypes.CountByOrigin` / `OriginCounts` / `sqlCountByOrigin`, the same three in `metadatafields` and `properties`, and their test cases are deleted. Swift: `WorkspaceNavCounts.sourceTypes` / `.metadataFields`, `WorkspaceNavOriginCounts`, the `GoStore` mapper taking `Provenencia_Engine_V1_VocabularyOriginCounts`, and `FakeStore.originCounts(from:)` go | No view draws a Configure badge (`WorkspaceSidebarSections`); the headers already fall back to their rows; nothing reads the environment value. One count source per number | Making the vocabulary writes report `counts` (the map's first scope line) — a full recount on each vocabulary create and delete for numbers nothing draws |
| D-9 | Eight RPCs go: `RemoveInstallIdentity` (5), `EnsureFileThumbnail` (41), `ListSubjectTypes` (45), `ClearSubjectPosition` (51), `ListProperties` (53), `ListSubjectTypeProperties` (57), `ListPlaceableSubjectTypes` (60), `GetSubjectTypePresentation` (61). For each: the `Method` value becomes `reserved N;` + `reserved "METHOD_…";` beside S14-16's, the request and response messages are deleted, the handler, its `dispatch.go` constant and case, and its handler tests go. `SubjectTypePresentation`, `SubjectType`, `Property` and `SubjectTypeProperty` messages stay (used by `GetPropertiesWorkspace`). `RemoveInstallIdentity` is not planned: no file under `docs/ideas/`, `docs/*.md` or the open spikes names it, and sign-out removes the identity through `identity.Remove` in `api/ffi/handlers/session.go`, which stays | Audit A-35; the map's Check "no RPC is left without a Swift caller except the three kept" — `ListSubjectTypes` and `ClearSubjectPosition` lose their last Swift caller with the test-only store methods (D-10), so they go too | Leaving the two RPCs whose Swift methods go — two new unclaimed RPCs |
| D-10 | Swift: `GenealogyStore.listSubjectTypes`, `clearSubjectPosition`, their `GoStore` and `FakeStore` bodies, `FakeStore.listSubjectTypesCalls`, `CoreMethod.listSubjectTypes` / `.clearSubjectPosition`, and the `OnboardingModelTests` stubs go. `SubjectStoreTests.listSubjectTypesReturnsSeededRows` is deleted; `setPositionPersistsAcrossSessionCloseThenClear` keeps its set-and-reopen half and is renamed `setPositionPersistsAcrossSessionClose` | The map's scope entry | — |
| D-11 | Go tests that read seeded ids through the deleted RPCs read them through `GetPropertiesWorkspace`: `propertyIDByKey` (`property_terms_test.go`) and a new `subjectTypeIDByKey(t, dir, key) string` beside it *(new, same file)* call it, and each `ListProperties` / `ListSubjectTypes` call in `citations_test.go`, `connect_test.go`, `conclusion_detail_test.go`, `person_detail_test.go`, `promote_test.go`, `subjects_test.go` and `subject_defs_test.go` uses the two helpers or reads `GetPropertiesWorkspaceResponse.properties` / `.types` directly where it needs the whole list. Dead callees: `subjectvocab.PlaceableTypes` and its assertions in `subjectvocab_test.go` go (the `PresentationFor` assertions in that case stay); `subjectpositions.Clear` and `sqlClear` go with the "clear" half of `subjectpositions_test.go` "get missing is tray; clear returns to tray" (renamed "get missing is tray"); `connect_test.go` "missing position invalid" removes the position with `DELETE FROM subject_positions WHERE subject_id = ?` on the test catalog's DB | Same payload (`GetPropertiesWorkspaceResponse` carries `properties` and `types`); a callee whose last caller goes is deleted | Keeping `ListProperties` as a test-only RPC — the wire is not a test fixture |
| D-12 | No new copy, no new error codes, no migration | Behavior and wire only | — |
| D-13 | Docs: rule 2's exceptions paragraph names only `.createdSource`, `.updatedSource`, `.changedSourceType`, `.deletedSource` (rows and the Source page), `.mutatedSourceWorkspace` / `.mutatedSourceMetadata` (a Source page's own workspace, the metadata field list), `.deletedSubject` and the type–field join (`.assignedTypeSuggestion` / `.removedTypeSuggestion`, an unaudited write with no effects); it adds that a vocabulary write applies `apply(effects)` and `vocabulary` reloads every key holding a label. "Conclusion keys" first bullet ends at "reloads only the handles it reports; a vocabulary label change reaches them through `vocabulary`" (no `conclusionTriggers`). `add-workspace-place`: line 41 drops the `syncCatalogCounts` clause; "Sidebar badges" says the sidebar badges come only from the recount and Configure rows have none; line 174's "writes that still report none" clause names only the type–field join. `add-ffi-handler` line 57 adds: "A write that is not audited takes no `user_id`." | The sweep rule; S14-16 wrote these lines with "until S14-17" | — |

**Rules this plan bends.** None. The map's scope line "The source type and metadata field writes report `counts` (Go `effects.ListCounts` on their registry entries), so their sidebar badges come from the recount" is replaced by D-8 on the user's answer (2026-11-02); STEP-7 updates the scope entry.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `effects` field on 14 responses *(new)* | proto field | `ReorderSubjectTypeProperties.effects`, S14-16's Source responses | Reused shape, one field per response |
| `TestVocabularyWritesReturnEffects` in `api/ffi/handlers/vocabulary_write_effects_test.go` *(new)* | Go test file | `api/ffi/handlers/source_write_effects_test.go` | Same table shape; a separate file because its fixtures (vocabulary rows) share nothing with `filedSubject` beyond the T-1 cardinality row, which calls it |
| `subjectTypeIDByKey` *(new)* | test helper | `propertyIDByKey` (`property_terms_test.go`) | Its sibling, same read |
| `.vocabulary` arm of `invalidations(for list:)` | switch arm | the same switch | Extends it; S14-27 replaces the switch |

Searched: `rg 'Effects: writeEffectsProto' api/ffi/handlers` (the shape), `rg 'IDByKey' api/ffi/handlers` (helpers), `rg 'CountByOrigin|OriginCounts'`, `rg 'catalogCounts'`.

**Builds on.** `writeEffectsProto` and `writes.Run`; the effects registry's `ListVocabulary` entries; S14-16's Swift store shape and `GoStore.mapWriteEffects`; `CatalogQueryRegistry.invalidations(for:project:)`; `add-ffi-handler` method reservation.
**Deletes.** `effectsPending`, `missingEffects`'s `pending` parameter and three subtests; eleven `CatalogMutation` cases, their kinds and arms; `conclusionTriggers`; `CatalogCounts.metadata` / `.sourceTypes` / `.properties` and `publish*`; `CatalogCountSummary.init(_:)` from origin counts; `SourceTypesModel.publishCounts`; `MetadataModel.syncCatalogCounts`; `PropertiesModel.syncCatalogCounts`; two `.onSettled` hooks; the `catalogCounts` pass-through on three models, three views, `WorkspaceDestinationHost`, `WorkspaceContent`, and `WorkspaceView`'s `.environment`; `WorkspaceNavCounts.sourceTypes` / `.metadataFields`; `WorkspaceNavOriginCounts`; the `GoStore` origin-counts mapper; `FakeStore.originCounts(from:)`; proto `VocabularyOriginCounts` and two nav-count fields; `typeOriginProto`, `fieldOriginProto`; `CountByOrigin`, `OriginCounts`, `sqlCountByOrigin` in `sourcetypes`, `metadatafields`, `properties`; eight RPCs (handlers, dispatch, messages, tests); `subjectvocab.PlaceableTypes`; `subjectpositions.Clear` / `sqlClear`; Swift `listSubjectTypes`, `clearSubjectPosition`, `listSubjectTypesCalls`, two `CoreMethod` constants; the `user_id` on two requests; the `userID:` parameter on two store methods; the "until S14-17" doc lines.
**Next adds.** S14-27 moves `.vocabulary`'s thirteen keys onto descriptor `follows` slots: each key named in D-5 declares the list it follows, and the switch goes. A new vocabulary table is `Lists: []List{ListVocabulary}` on its registry entry and gets every vocabulary reload with no Swift change. A new vocabulary-label key adds itself to D-5's set (S14-27: its descriptor).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Source type, metadata field or Property create, rename or delete now reaches the Conclusion keys | `CatalogQueryRegistry` Conclusion descriptors (`whenHidden: .markStale` lists, `.evict` details) | Hidden on a Configure page, so lists mark stale and details evict, as `.updatedProperty` did; a deleted Property's page rows were already refused while in use (`properties.Delete` refuses `ErrInUse`) |
| A Property or binding write now reaches `sourceTypesList`, `metadataFieldsList`, `typeSuggestions`, `sourceWorkspace`, `propertyTerms` | their descriptors | Each reload reads the catalog afresh; `sourceWorkspace` is `.markStale` when hidden; a deleted type's `typeSuggestions` key reloads to the engine's refusal as today on `.deletedSourceType` (`allCached`) |
| A composer custom term now reaches `propertiesWorkspace`, `sourceTypesList` and `sourceWorkspace`, which the composer place holds on screen | `CitationComposerView` | They reload in place behind the open composer; the composer's draft lives in `CitationComposerModel`, not in those keys, so the open term sheet and the draft stay |

### Scale

At the map's Scale (3,000 Sources; 37,000 Subjects; 92,000 Observations; 200,000 audit rows):

- **A rename on a Configure page.** On screen: that page's one or two keys reload (`sourceTypesList` ~20 rows, `metadataFieldsList` ~40, `propertiesWorkspace` ~60 Properties and 10 types). Hidden: `sourceTypesList`, `metadataFieldsList`, warmed `typeSuggestions` and `propertyTerms` revalidate (default `whenHidden`), each a list of tens of rows; `sourceWorkspace` keys and the Conclusion lists mark stale and load only when shown; Conclusion details and Person reads evict. No Persons list (3,000+ rows) loads at write time. This is what `.updatedProperty` already did for the Conclusion keys.
- **A composer custom term** (the one frequent vocabulary write). Today it reloads warmed `propertyTerms`. After: also the on-screen `propertiesWorkspace` (the same read every Observation save in the composer already reloads through `property_usage`), `sourceTypesList` and the composer's `sourceWorkspace` (one Source's page read); hidden Conclusion keys stale or evict. Three small reads more, serialized by `catalogsession`, once per new term.
- **Go.** Each handler adds one `writeEffectsProto` copy of a set `writes.Run` already resolved; no new query. `GetWorkspaceNavCounts` runs two `GROUP BY origin` queries fewer per recount.

### Edges

- **Assumptions.** `invalidations(for:project:)` dedupes its output, so `.vocabulary` with `.propertyUsage` or `structure` names `propertiesWorkspace` once (T-5). `WorkspaceSession.apply` revalidates each key once even when `refreshes` and invalidations both name it (`aKeyInvalidatedAndRefreshedRevalidatesOnce`, kept with effects in place of `.updatedProperty`).
- **Neighbors.** A write that changes nothing (`UpdateSourceType` with the same label, description and icon) returns empty effects, revision 0, and reloads nothing: today it reloaded `sourceTypesList`. The model's `patchTypesList` still writes the returned row, so the page shows the same row (T-7). A cardinality change on a user Property reports `handles` as well as `vocabulary` (AC-1.5, T-1); a label change reports no handles. The type–field join reports nothing and keeps its mutation (AC-1.6, T-14).
- **Removed inputs.** A deleted Property releases its bindings (`subject_type_properties`, `vocabulary`), so the Properties workspace reloads once. A deleted Source type cascades its suggestion joins without audit; its `typeSuggestions` key is in `.allCached(.typeSuggestions)`, as on `.deletedSourceType` today.
- **Real data.** None — no table-driven rule changes; the registry entries are unchanged.
- **Untrusted size.** None — no new list read from data the app did not write.
- **Labels the effects do not reach.** `EvidenceGraphRows` observations and `citationsByArtifact` carry `propertyLabel`; a Property rename leaves a visited canvas showing the old label until its graph next reloads. Today's `.updatedProperty` does not reach them either. Out (structural-debt line), so no reload a researcher sees changes beyond D-5's union.

### Schema

None — no catalog change. The effects registry is unchanged.

### Go core

| Package | Change |
| --- | --- |
| `core/database/sourcetypes` | Delete `CountByOrigin`, `OriginCounts`, `sqlCountByOrigin` and the "CountByOrigin splits seeded user and plugin" test case |
| `core/database/metadatafields` | Same three and the same test case |
| `core/database/properties` | Delete `CountByOrigin`, `OriginCounts`, `sqlCountByOrigin` (no caller today) and any test case naming them |
| `core/database/subjectvocab` | Delete `PlaceableTypes` and its assertions in `subjectvocab_test.go` |
| `core/database/subjectpositions` | Delete `Clear` and `sqlClear`; the package comment's "Set and Clear" reads "Set" |
| `core/database/connect` | `connect_test.go` "missing position invalid" removes the position row with SQL (D-11) |

Errors: none new.

### FFI

`api/proto/engine.proto`:

```proto
// Method
reserved 5, 41, 45, 51, 53, 57, 60, 61;
reserved "METHOD_REMOVE_INSTALL_IDENTITY", "METHOD_ENSURE_FILE_THUMBNAIL",
  "METHOD_LIST_SUBJECT_TYPES", "METHOD_CLEAR_SUBJECT_POSITION",
  "METHOD_LIST_PROPERTIES", "METHOD_LIST_SUBJECT_TYPE_PROPERTIES",
  "METHOD_LIST_PLACEABLE_SUBJECT_TYPES", "METHOD_GET_SUBJECT_TYPE_PRESENTATION";

message AssignTypeFieldRequest {
  string project_dir = 1;
  reserved 2;
  reserved "user_id"; // the join is not audited (catalog-deletes.md)
  string type_id = 3;
  string field_id = 4;
}
// RemoveTypeFieldRequest: the same.

message GetWorkspaceNavCountsResponse {
  int32 sources = 1;
  reserved 2, 3, 4;
  reserved "source_types", "metadata_fields", "files";
  int32 persons = 5;
  int32 events = 6;
  int32 places = 7;
}

message CreateSourceTypeResponse { SourceType type = 1; WriteEffects effects = 2; }
message DeleteSourceTypeResponse { WriteEffects effects = 1; }
// … D-1's other twelve.
```

Deleted messages: `RemoveInstallIdentityRequest/Response`, `EnsureFileThumbnailRequest/Response`, `ListSubjectTypesRequest/Response`, `ClearSubjectPositionRequest/Response`, `ListPropertiesRequest/Response`, `ListSubjectTypePropertiesRequest/Response`, `ListPlaceableSubjectTypesRequest/Response`, `GetSubjectTypePresentationRequest/Response`, `VocabularyOriginCounts`. Handlers: `source_defs.go`, `subject_defs.go`, `property_terms.go`, `workspace_counts.go`, `thumbnails.go` (delete `EnsureFileThumbnail`; the `enrichSourceProto` comment names `thumbnailPass` callers `SetSourceCover` and the Artifact reads), `identity.go`, `subjects.go`; `api/ffi/dispatch.go`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

None — no chrome change. Store, registry and model changes are in D-4 to D-8 and D-10. The header lines draw the same text from the same rows they fall back to today.

### Strings

None — no key added or removed. The vocabulary toasts name the same rows.

## 7. Steps

**STEP-1 — Vocabulary writes return their effects in Go; the allow-list goes**
Files: `api/proto/engine.proto`, `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift` (regenerated), `api/ffi/handlers/source_defs.go`, `subject_defs.go`, `property_terms.go`, `write_effects_test.go`, `vocabulary_write_effects_test.go` *(new)*, `source_defs_test.go`; Swift for D-2 only: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Features/SourceTypes/SourceTypesModel.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `SourceTypesModelTests.swift`. Skill: `add-ffi-handler`.
Do: write T-1 first (it fails: no `effects`). D-1's proto fields and handler changes; D-2's reservation, `parseSuggestionJoin(typeID, fieldID string)`, and the Swift `userID:` removal on `assignTypeField` / `removeTypeField` so the app builds against the regenerated messages; D-3 in `write_effects_test.go`; T-3's `user_id` lines in `TestRetiredFieldsStayReserved`; drop `UserId` from the assign / remove requests in `source_defs_test.go`.
Check: `./scripts/generate-proto.sh`; `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; `rg -n effectsPending api` prints nothing; `xcodebuild test` (as `scripts/check-pr.sh` runs it) passes.

**STEP-2 — `vocabulary` reaches every key that holds a label**
Files: `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `CatalogMutation.swift` (`CatalogList.vocabulary` doc), `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`. Skill: `add-workspace-place`.
Do: T-5 first. D-5's switch split. The mutations still exist in this step; `invalidations(for:by:project:)` dedupes the union.
Check: T-5 and the updated `listCases` pass under `xcodebuild test`.

**STEP-3 — The Swift store returns vocabulary effects; callers apply them; the mutations go**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Features/SourceTypes/SourceTypesModel.swift`, `macos/App/Features/Metadata/MetadataModel.swift`, `macos/App/Features/Properties/PropertiesModel.swift`, `macos/App/Features/CitationComposer/CitationComposerModel.swift`, `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`; tests `OnboardingModelTests.swift`, `SourceTypesModelTests.swift`, `MetadataModelTests.swift`, `PropertiesModelTests.swift`, `CitationComposerModelTests.swift`, `CatalogQueryRegistryTests.swift`, `WorkspaceSessionTests.swift`. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-4, D-7, then D-6. Rewrite `applyFieldCRUDInvalidatesMetadataList` and `applyTypeCRUDInvalidatesTypesList` to apply `CatalogWriteEffects(lists: [.vocabulary])`; replace the `.updatedProperty` / `conclusionTriggers` assertions with T-6; in `aKeyInvalidatedAndRefreshedRevalidatesOnce` use `CatalogWriteEffects(handles: ["e1"], evidenceGraphs: ["g1"], lists: [.vocabulary])` with no mutation. Model tests that relied on a mutation's reload seed `store.nextWriteEffects = CatalogWriteEffects(lists: [.vocabulary])`.
Check: `rg -n 'createdSourceType|updatedSourceType|deletedSourceType|createdMetadataField|updatedMetadataField|deletedMetadataField|createdProperty\b|updatedProperty|deletedProperty|mutatedSubjectTypeProperties|createdPropertyTerm|conclusionTriggers' macos` prints nothing; T-6 to T-11 and T-14 pass; `xcodebuild test` passes with no Swift warning.

**STEP-4 — The vocabulary counts nothing draws go**
Files: `macos/App/Features/Catalog/CatalogCounts.swift`, `macos/App/Features/SourceTypes/SourceTypesModel.swift`, `macos/App/Features/SourceTypes/SourceTypesView.swift`, `macos/App/Features/Metadata/MetadataModel.swift`, `macos/App/Features/Metadata/MetadataView.swift`, `macos/App/Features/Properties/PropertiesModel.swift`, `macos/App/Features/Properties/PropertiesView.swift`, `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `macos/App/Features/Workspace/WorkspaceContent.swift`, `macos/App/Features/Workspace/WorkspaceView.swift`, `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`; `api/proto/engine.proto` (+ regenerated), `api/ffi/handlers/workspace_counts.go`, `workspace_counts_test.go`, `write_effects_test.go` (T-3 nav-count lines); `core/database/sourcetypes/sourcetypes.go` and `_test.go`, `core/database/metadatafields/metadatafields.go` and `_test.go`, `core/database/properties/properties.go` and `_test.go`; tests `CatalogCountsTests.swift`, `SourceTypesModelTests.swift`, `MetadataModelTests.swift`, `PropertiesModelTests.swift`, `WorkspaceDestinationHostTests.swift`, `WorkspaceSidebarSectionsTests.swift`. Skills: `add-workspace-place`, `add-ffi-handler`, `add-swift-test`.
Do: D-8. Delete `publishReplacesVocabularySummariesWithoutStoreRoundTrip`, `createPublishesTheCountOnceTheReloadSettles`, `deletePublishesTheCountOnceTheReloadSettles` and the `counts.metadata` / `counts.sourceTypes` assertions; T-12, T-13.
Check: `rg -n 'publishSourceTypes|publishMetadata|publishProperties|syncCatalogCounts|WorkspaceNavOriginCounts|VocabularyOriginCounts|CountByOrigin' api core macos --glob '!*.pb.*'` prints nothing; `rg -n catalogCounts macos/App/Features/{SourceTypes,Metadata,Properties}` prints nothing; Go and Xcode tests pass.

**STEP-5 — RPCs with no Swift caller go**
Files: `api/proto/engine.proto` (+ regenerated), `api/ffi/dispatch.go`, `api/ffi/handlers/thumbnails.go`, `subject_defs.go`, `identity.go`, `subjects.go`, their tests (`thumbnails_test.go`, `subject_defs_test.go`, `identity_test.go`, `subjects_test.go`), the fixture files of D-11, `write_effects_test.go` (T-3 method lines); `core/database/subjectvocab/subjectvocab.go` and `_test.go`, `core/database/subjectpositions/subjectpositions.go` and `_test.go`, `core/database/connect/connect_test.go`; Swift `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`, `OnboardingModelTests.swift`, `SubjectStoreTests.swift`. Skill: `add-ffi-handler`.
Do: D-9, D-10, D-11.
Check: `rg -n 'EnsureFileThumbnail|GetSubjectTypePresentation|ListPlaceableSubjectTypes|ListPropertiesRequest|ListSubjectTypeProperties|RemoveInstallIdentity|ListSubjectTypesRequest|ClearSubjectPosition|PlaceableTypes\(|subjectpositions\.Clear' api core macos/App --glob '!*.pb.*'` prints only the `reserved` names in `engine.proto` and T-3's lines; `rg -n 'listSubjectTypes|clearSubjectPosition' macos` prints nothing; Go and Xcode tests pass.

**STEP-6 — Guidance says vocabulary writes report effects**
Files: `docs/macos-client-patterns.md`, `.agents/skills/add-workspace-place/SKILL.md`, `.agents/skills/add-ffi-handler/SKILL.md`.
Do: D-13.
Check: `rg -n 'until S14-17|conclusionTriggers|syncCatalogCounts|publish\*' docs/*.md .agents .cursor` prints nothing; `python3 scripts/check-guidance-names.py` passes.

**STEP-7 — Version and docs**
Files: `VERSION`, `core/version.go`, `MARKETING_VERSION` in `macos/Provenencia.xcodeproj/project.pbxproj`; `docs/deployment-plan/spike-14/completed.md`, `deployment-plan.md` (S14-17 row status, scope entry: D-8 replaces the "report `counts`" line and the two extra RPCs of D-9 are named), this plan (status, Landed note), `docs/ideas/structural-debt.md` (one line: a Property rename leaves visited canvases and `citationsByArtifact` on the old label; their keys follow no vocabulary list).
Do: per the Definition of done; the patch bump per `docs/versioning.md`.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-17-vocabulary-effects.md` passes; links resolve; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `api/ffi/handlers/vocabulary_write_effects_test.go` `TestVocabularyWritesReturnEffects` *(new)* | One row per D-1 write on a fresh project (`newTestProject`-style fixture used by `source_defs_test.go`): create, update with a new label, delete → `effects.lists == ["vocabulary"]`, `revision > 0`. Row "a cardinality change names the handles that observe the Property": from `citationFixture`, `CreateProperty` a user text Property, `CreateCitationWithObservations` one Observation on the fixture's Subject with it, `promoteInDir` the Subject, `UpdateProperty` cardinality `single` → `multiple` → `effects.handles` holds the handle and `lists == ["vocabulary"]`. Row "an unchanged Source type update reports nothing": `UpdateSourceType` with the same label, description, icon → `lists` empty, `revision == 0` | AC-2.2, AC-1.5, D-1 |
| T-2 | Go | `api/ffi/handlers/write_effects_test.go` `TestEveryAuditedWriteReturnsEffects` (changed) | `missingEffects(engine.File_engine_proto)` → none; the two fixture subtests pass | AC-2.1, D-3 |
| T-3 | Go | `write_effects_test.go` `TestRetiredFieldsStayReserved` (extended) | `AssignTypeFieldRequest` and `RemoveTypeFieldRequest` reserve 2 / `user_id`; `GetWorkspaceNavCountsResponse` reserves 2, 3 / `source_types`, `metadata_fields`; `Method` reserves D-9's eight numbers and names and declares none of them | AC-2.3, AC-3.1, D-8, D-9 |
| T-4 | Go | `source_defs_test.go` assign / remove cases; `workspace_counts_test.go` `TestGetWorkspaceNavCounts` (changed) | Assign and remove with no `UserId` → the suggestion list as today; nav counts → `sources`, `persons`, `events`, `places` only | D-2, D-8 |
| T-5 | Swift | `CatalogQueryRegistryTests.swift` `vocabularyReachesEveryLabelKey` *(new)* and `listCases` (changed) | `invalidations(for: CatalogWriteEffects(lists: [.vocabulary]), project:)` == D-5's thirteen, in order; `[.propertyUsage]` → `[.key(.propertiesWorkspace)]`; `[.vocabulary, .propertyUsage]` with `structure: true` → thirteen, `propertiesWorkspace` once | AC-1.1, AC-1.2, AC-1.4, D-5 |
| T-6 | Swift | `CatalogQueryRegistryTests.swift` `noMutationReachesAConclusionKey` *(new, replaces the `conclusionTriggers` assertions)* | For every remaining `CatalogMutation` (one sample each), `invalidations(by:)` names no Conclusion kind | D-6 |
| T-7 | Swift | `SourceTypesModelTests.swift` `addingATypeReloadsFromItsEffects` *(new)* | `nextWriteEffects = [.vocabulary]`, warmed `sourceTypesList` and `metadataFieldsList`; add "Diary" → both reload once; selection is the new type. A second case: `nextWriteEffects = .empty`, edit with the same label → no reload, row unchanged | AC-1.3, D-7 |
| T-8 | Swift | `MetadataModelTests.swift` `renamingAFieldReloadsItsDependents` *(new)* | `nextWriteEffects = [.vocabulary]`, warmed `typeSuggestions(t1)`, `sourceWorkspace(s1)` hidden; rename "Volume" → `metadataFieldsList` and `typeSuggestions(t1)` reload, `sourceWorkspace(s1)` is stale | AC-1.2 |
| T-9 | Swift | `PropertiesModelTests.swift` `renamingAPropertyStalesConclusionPages` *(new)* | `nextWriteEffects = [.vocabulary]`, warmed hidden `conclusionDetail(e1)` and `personsList`; `submitEdit(label: "Trade", …)` → `conclusionDetail(e1)` handle is nil (evicted), `personsList` stale, `propertiesWorkspace` reloaded | AC-1.1 |
| T-10 | Swift | `CitationComposerModelTests.swift` `customTermReloadsWarmedTerms` *(new)* | `nextWriteEffects = [.vocabulary]`, warmed `propertyTerms(p1)`; `confirmCustomTerm` "Cooper" → `propertyTerms(p1)` reloaded once and the draft holds the new term id | AC-1.4 |
| T-11 | Swift | `WorkspaceSessionTests.swift` `aKeyInvalidatedAndRefreshedRevalidatesOnce` (changed) | Effects with `handles`, `evidenceGraphs` and `vocabulary`, no mutation → the `personClaims` key is evicted once and not reloaded | AC-1.1, Edges |
| T-12 | Swift | `SourceTypesModelTests.swift` `loadPopulatesTypesPool` and `MetadataModelTests.swift` `loadPopulatesFields` (renamed from `…AndCounts`), plus `countLineCountsTheRowsShown` *(new, each file)* | 3 seeded + 2 user types → `countLine` the "5 · 3 seeded · 2 yours" string from `L10n.SourceTypes.countLine`; 4 seeded + 1 plugin field → `L10n.Metadata` plugin variant | AC-4.1, AC-4.2, D-8 |
| T-13 | Swift | `CatalogCountsTests.swift` `refreshAllPopulatesEveryDestination` (changed) | Recount → `sources`, `persons`, `events`, `places` set; `badge(for: .sourceTypes / .metadata / .properties)` nil | AC-4.3, D-8 |
| T-14 | Swift | `SourceTypesModelTests.swift` `assigningAppendsTheFieldAndUpdatesTheListCount`, `removingASuggestionNoSourceUsesOmitsTheValueCount` (kept, calls lose `userID:`) | Assign / remove → suggestion list and type row count as today; `typeSuggestions(t1)` reloads through `.assignedTypeSuggestion` | AC-1.6, D-2 |
| T-15 | Go / Swift | Deleted with their code: `TestEnsureFileThumbnail`, `TestRemoveInstallIdentity`, `TestListSubjectTypes`, `TestClearSubjectPositionBadProto`, `TestListProperties`, `SubjectStoreTests.listSubjectTypesReturnsSeededRows`; changed: `SubjectStoreTests.setPositionPersistsAcrossSessionClose`, `subjectpositions_test.go` "get missing is tray", `connect_test.go` "missing position invalid", `subjectvocab_test.go` "registry placeable presentation connect" | Pass with the fixtures of D-11 | AC-3.1, AC-3.2, D-9–D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CatalogQueryRegistryTests`, `WorkspaceSessionTests`, the three vocabulary model test files, the `api/ffi/handlers` fixtures that read seeded ids, `TestSwiftVocabularyMatchesGo` (`api/ffi/swift_vocabulary_test.go`), `core/database/effects` lists tests (unchanged, must stay green).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md rule 2 and "Conclusion keys" (D-13)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with one promoted Person whose Citation records an Occupation (create the user Property "Occupation" on Properties first if the project has none). Open the Person page, then go to Properties and rename "Occupation" to "Trade". Go Back to the Person: the row reads "Trade" (AC-1.1).
2. Open a Source page, then go to Metadata and rename a user field shown on that page. Go Back: the Source page shows the new label (AC-1.2).
3. On Source types add "Diary": it is selected and the header count line goes up by one; the sidebar shows no badge on Configure rows (AC-1.3, AC-4.1, AC-4.3).
4. On Source types assign a field to "Diary", then remove it: the suggestion list and the row's count follow (AC-1.6).
5. In the citation composer add a custom term to a term-valued Property: the combo offers it and the draft keeps its values (AC-1.4).
6. Each view at Regular and Wide: nothing is drawn differently.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A test or view reads the vocabulary counts through a path §4 missed | STEP-4's `rg` check; the build fails on a deleted member |
| A Go fixture needs an id only a deleted RPC returned | D-11 names the replacement read; `GetPropertiesWorkspaceResponse` carries both lists |
| The composer's on-screen keys reload behind an open term sheet and reset it | Reach row 3; T-10 asserts the draft keeps the term |

**Stop and report if:**

- A vocabulary write's `writes.Result.Effects.Lists` lacks `vocabulary` for a create, a labelled update or a delete (T-1 fails on the registry, not the handler): the registry entry would have to change, which D-1 rules out.
- `S14-16`'s merged code differs from §4 (its PR [#686](https://github.com/mendahu/provenencia/pull/686) changes `missingEffects`, `effectsPending`, `GoStore.mapWriteEffects` or the Swift store's effect shape before it lands).
- Any non-test file outside the handlers this plan deletes calls `ListProperties`, `ListSubjectTypes`, `ClearSubjectPosition`, `subjectpositions.Clear` or `subjectvocab.PlaceableTypes`.
- A view reads `CatalogCounts` from the SwiftUI environment (`@Environment(CatalogCounts.self)`) after all.
- A remaining `CatalogMutation` case is applied by a caller this plan does not name and reaches a key only through a deleted case.
- An idea file or open spike, found at implementation time, plans a use of `RemoveInstallIdentity` or `EnsureFileThumbnail`.

## 12. Open questions

None.

## 13. PR

**Title:** S14-17 — Vocabulary writes return their effects; dead RPCs go
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-17-vocabulary-effects.md`.
