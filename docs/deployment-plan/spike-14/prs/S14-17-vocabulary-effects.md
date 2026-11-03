# S14-17 — Vocabulary writes return their effects; dead RPCs go

**Status:** Landed in #694 (2026-11-03)
**Landed:** Base was `s14-27-lists-through-descriptors` @ `f1b3fe65`, as planned; every §4 row held. Differences: T-1 puts the term writes on the seeded `event_type` Property, because a researcher cannot create a term-valued Property; `graphcache.OnCommit` drops labels on `property_vocabulary` alone, as D-5 and T-3 say (the Go core table says either family); `FakeStore.assignSubjectTypePropertyErrors` was added for T-9; the `subjectvocab` case keeps its presentation, locked-binding and connect-rule assertions, renamed "registry presentation locked connect rules"; T-9 and T-10 assert that the reload happened, not an exact fetch count, and T-9's create case binds one failing type, because `Draft.bindTypeIDs` is a set with no order for "the second assign"; `QueryHandleSettled.onSettled` and `artifacts.HasPrimaryFile` lost their last callers and were deleted in review round 1; D-6's literal `propertyTerms` `invalidation: { _, _ in .allCached(.propertyTerms) }` is superseded: `CatalogQueryDescriptor.make` now takes `invalidateOn` and `invalidation` together or neither, so the 16 descriptors that name no mutation (`propertyTerms` among them) carry no closure; `CatalogQueryRefresh.followsPropertyLabels` is the one shared `.list(.propertyVocabulary)` entry; `SourceTypesModel.countLine` and `MetadataModel.countLine` are `nil` until the list has a value; `MARKETING_VERSION` bumped in the macOS commit; the review log's trailing blank line was trimmed so `git diff --check` passes.
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes (batch 5; was Spike 12's S12-33)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: every view draws the same rows and copy; only what reloads after a vocabulary write changes, and the Source types and Metadata header lines count the rows the page already shows)
**Board:** —
**Depends on:** S14-16 (PR [#686](https://github.com/mendahu/provenencia/pull/686), merged 2026-11-02), S14-27 (PR [#693](https://github.com/mendahu/provenencia/pull/693), open; this PR is stacked on its branch, user 2026-11-03)
**Blocks:** —
**Branch:** `s14-17-vocabulary-effects`
**Version bump:** patch — a refresh fix and dead-code removal, no feature ([`docs/versioning.md`](../../../versioning.md)); the next patch after `VERSION` at merge (0.2.10 at planning)
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [`app-health-audit.md`](../app-health-audit.md) A-29, A-35; working rules "Every audited write returns its effects, and a test holds them to it", "FakeStore is a seeded fake, not a second engine"; [`macos-client-patterns.md`](../../../macos-client-patterns.md) cache contract rule 2, "Conclusion keys" and "One cache owns each list"; [`core/database/effects/README.md`](../../../../core/database/effects/README.md) (`Set.Lists`); [`catalog-deletes.md`](../../../catalog-deletes.md) (the type–field join is never audited); S14-16's plan **Next adds** ([`S14-16-promote-source-effects.md`](S14-16-promote-source-effects.md))
**Planned:** 2026-11-02 against `s14-16-promote-source-effects` @ `a5e057b`. **Refreshed** 2026-11-03 against `s14-27-lists-through-descriptors` @ `f1b3fe65` (`main` @ `42eb321c`, S14-16 merged, + S14-27). The PR's base is S14-27's branch until #693 merges, then `main`
**Review:** [`S14-17-vocabulary-effects.review.md`](S14-17-vocabulary-effects.review.md) — round 1: Fits with changes / Ready with changes → Fixed, two To user answered (2026-11-02); round 2 (refresh, 2026-11-03): Fits with changes / Ready with changes → Fixed

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

Every vocabulary write tells the app what it changed, the way every other audited write does after S14-16. Creating, renaming or deleting a Source type, a metadata field, a Property or a Property term, and binding or unbinding a Property on a Subject type, return `WriteEffects`. Go reports which vocabulary family moved: `source_vocabulary` (Source types, metadata fields) or `property_vocabulary` (Properties, terms, Subject types, bindings), in place of the one `vocabulary` list. Each cached key follows the family whose labels it holds, through the descriptor slot S14-27 adds. Nine `CatalogMutation` cases and `CatalogQueryRegistry.conclusionTriggers` go. `TestEveryAuditedWriteReturnsEffects` runs with no allow-list. The two unaudited type–field join writes stop sending a `user_id`. The vocabulary counts nothing draws go. Eight RPCs with no Swift caller go, with their handlers, tests and Swift store methods, and their method numbers are reserved.

**Why now.** S14-16 added the pin test with sixteen vocabulary writes on its allow-list (`effectsPending`) and left `.updatedProperty` in `conclusionTriggers`; this PR empties both (working rule, map review K-10). Audit A-35 lists the unclaimed RPCs. User answers, 2026-11-02: one PR; the vocabulary counts nobody draws are removed; the type–field join stays unaudited and drops its `user_id`; vocabulary splits into two lists so "One cache owns each list" stands; S14-27's descriptor slot lands first.

**What stays.** What every view draws. `.assignedTypeSuggestion(typeId:)`, `.removedTypeSuggestion(typeId:)` (the join has no effects, D-2), `.updatedMetadataField(id:)` and `.deletedMetadataField(id:)` (they name the Source pages a field relabel or delete restates, which a list cannot tell from a create, D-6) stay. `.createdSource`, `.updatedSource`, `.changedSourceType`, `.deletedSource`, `.mutatedSourceWorkspace` and `.mutatedSourceMetadata` stay (S14-27 retires `.deletedSubject`). The models' list patches after a write stay. `UpdatePropertyTerm`, `DeletePropertyTerm` and `ReorderSubjectTypeProperties` stay (named by [`term-management.md`](../../../ideas/term-management.md) and [`binding-order-editing.md`](../../../ideas/binding-order-editing.md)). The sidebar recount and its `counts` trigger are unchanged.

## 2. User stories

**US-1.** As a researcher, when I rename a Property, a Source type or a metadata field, the pages that show the label show the new one the next time I look, and a vocabulary write never reloads a page that shows none of its labels.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A warmed Person page (`conclusionDetail` for `e1`) and the warmed Persons list; on Properties the researcher renames the user Property "Occupation" to "Trade" | The effects carry `lists: ["property_vocabulary"]`; the hidden `conclusionDetail` key is evicted and the Persons list is marked stale, so opening the Person shows "Trade" |
| AC-1.2 | A visited Source page (`sourceWorkspace` for `s1`) and a warmed `typeSuggestions` for type `t1`; on Metadata the researcher renames the field "Volume" to "Book" | The effects carry `source_vocabulary`; `metadataFieldsList` and `typeSuggestions(t1)` reload; `sourceWorkspace(s1)` is marked stale through `.updatedMetadataField(id:)` |
| AC-1.3 | The researcher adds a Source type "Diary" on Source types, with a Source page and a Person page visited earlier | `sourceTypesList`, `metadataFieldsList` and warmed `typeSuggestions` reload; the Source page and the Person page are untouched; the new row shows, selected, as today |
| AC-1.4 | In the citation composer the researcher adds a custom term "Cooper" to Occupation | The warmed `propertyTerms(Occupation)` key and the on-screen `propertiesWorkspace` reload, and no other `propertyTerms` key does; the composer's `sourceWorkspace`, `sourceTypesList` and `evidenceGraph` do not; the combo offers "Cooper" and the draft keeps its values |
| AC-1.5 | The researcher sets a user Property's Holds from single to multiple, and one Person observes it | The effects name that Person's handle in `handles` and carry `property_vocabulary`; the Person's page reloads its values |
| AC-1.6 | The researcher assigns the field "Volume" to type `t1` | `typeSuggestions(t1)` reloads and every visited Source page is marked stale, as today (`.assignedTypeSuggestion(typeId:)`) |

**US-2.** As the engine, every request with a `user_id` returns `WriteEffects`, and no write is excused.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `TestEveryAuditedWriteReturnsEffects` on `engine.File_engine_proto` | Passes; the file has no `effectsPending` map |
| AC-2.2 | `CreateSourceType`, `UpdateSourceType` (label changed), `DeleteSourceType`, `CreateMetadataField`, `UpdateMetadataField` (label changed), `DeleteMetadataField` on user rows | `effects.lists == ["source_vocabulary"]`, `effects.revision` > 0 |
| AC-2.3 | `CreateProperty`, `UpdateProperty` (label changed), `DeleteProperty`, `CreatePropertyTerm`, `UpdatePropertyTerm` (label changed), `DeletePropertyTerm`, `AssignSubjectTypeProperty`, `RemoveSubjectTypeProperty`, `ReorderSubjectTypeProperties` on user rows | `effects.lists == ["property_vocabulary"]`, `effects.revision` > 0 |
| AC-2.4 | `AssignTypeFieldRequest`, `RemoveTypeFieldRequest` | No `user_id` field; number 2 and the name are reserved |

**US-3.** As a contributor, every RPC in `engine.proto` has a Swift caller, except the three an idea file names.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `rg` for `EnsureFileThumbnail`, `GetSubjectTypePresentation`, `ListPlaceableSubjectTypes`, `ListProperties`, `ListSubjectTypeProperties`, `RemoveInstallIdentity`, `ListSubjectTypes`, `ClearSubjectPosition` | No handler, dispatch case, message or Swift method; each `METHOD_…` number and name is reserved |
| AC-3.2 | `rg -n 'listSubjectTypes|clearSubjectPosition' macos` | No hit |

**US-4.** As a researcher, the Source types and Metadata headers count the rows they show.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | The Source types page lists 3 seeded and 2 user types | The header reads `L10n.SourceTypes.countLine(total: 5, seeded: 3, user: 2)` from those rows |
| AC-4.2 | The Metadata page lists 4 seeded fields and 1 plugin field | The header reads `L10n.Metadata.countLineWithPlugin(total: 5, seeded: 4, user: 0, plugin: 1)` |
| AC-4.3 | The sidebar after any write | Badges on Sources, Persons, Events and Places as today; none on Configure rows |

## 3. Scope

| In | Covered by |
| --- | --- |
| `effects.ListVocabulary` splits into `ListSourceVocabulary` and `ListPropertyVocabulary` (D-5) | STEP-1, AC-2.2, AC-2.3 |
| Fourteen vocabulary handlers return `WriteEffects` (D-1) | STEP-2, AC-2.2, AC-2.3 |
| `effectsPending` and `missingEffects`'s `pending` parameter go (D-3) | STEP-2, AC-2.1 |
| `AssignTypeField` / `RemoveTypeField` drop `user_id` (D-2) | STEP-2, AC-2.4, AC-1.6 |
| Each label-holding key follows its family through S14-27's slot (D-6) | STEP-3, AC-1.1–AC-1.4 |
| A term write reports a keyed `property_terms` read that `propertyTerms` follows (D-14) | STEP-1, STEP-2, STEP-3, AC-1.4 |
| Swift store vocabulary writes return `CatalogWriteEffects`; callers apply them; nine mutation cases and `conclusionTriggers` go (D-4, D-6, D-7) | STEP-4, STEP-5, AC-1.1–AC-1.5 |
| Vocabulary counts removed (D-8) | STEP-6, AC-4.1–AC-4.3 |
| Eight RPCs with no Swift caller deleted, numbers reserved; Swift `listSubjectTypes` / `clearSubjectPosition` deleted (D-9–D-11) | STEP-7, AC-3.1, AC-3.2 |
| Guidance (D-13) | STEP-8 |

| Out | Home |
| --- | --- |
| Deleting `UpdatePropertyTerm`, `DeletePropertyTerm`, `ReorderSubjectTypeProperties` | not planned ([`term-management.md`](../../../ideas/term-management.md), [`binding-order-editing.md`](../../../ideas/binding-order-editing.md)) |
| The descriptor list slot and moving today's `vocabulary` / `property_usage` / `structure` reloads onto it | S14-27 (lands first) |
| Lists finer than two families (one per vocabulary table) | not planned (user, 2026-11-02: two lists) |
| A visited canvas (`evidenceGraph`) and `citationsByArtifact` keep a Property's old label after a rename (their rows carry `propertyLabel`; following `property_vocabulary` would reload the on-screen graph on every composer custom term) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md), the line STEP-8 adds |
| Auditing the type–field join | not planned (user, 2026-11-02) |
| `CreateSourceType`, `UpdateSourceType`, `CreateMetadataField`, `UpdateMetadataField` record their audit revision with no `UserID` in `writes.Op` (`api/ffi/handlers/source_defs.go`) | follow-up task suggested at planning; not this PR |
| The audit's test-net items (L-16) | not planned (map scope entry) |
| A sidebar recount that fails after a write is silent (S14-16 UX-1) | S14-16's follow-up |

## 4. Current state

Verified on `s14-27-lists-through-descriptors` @ `f1b3fe65` (refresh, 2026-11-03). S14-27's rows are in their post-S14-27 state.

| Area | Where | Today |
| --- | --- | --- |
| Pin test | `api/ffi/handlers/write_effects_test.go` `effectsPending`, `missingEffects`, `TestEveryAuditedWriteReturnsEffects` | Sixteen `…Request` names map to `"S14-17"`; `missingEffects(file, pending)` skips them and reports a stale entry; five fixture subtests, three of which test `pending` |
| Reserved fields test | `write_effects_test.go` `TestRetiredFieldsStayReserved` | Pins S14-16's reserved fields and `METHOD_PROMOTE_SUBJECT` (85) |
| Effects mapping | `api/ffi/handlers/write_effects.go` `writeEffectsProto` | The one mapper every handler uses |
| Lists | `core/database/effects/registry.go` `ListVocabulary`, `Lists`; entries `source_types`, `source_metadata_fields`, `subject_types`, `properties`, `property_terms`, `subject_type_properties` | All six set `Lists: []List{ListVocabulary}`; `properties` adds `Handles: onField("cardinality", …)`; `property_terms` adds `Handles` and `Structure`; `source_type_metadata_fields` is `None` |
| Go display cache | `core/database/graphcache/commit.go` `OnCommit` | `set.Has(effects.ListVocabulary)` drops every cached label and vocabulary memo |
| List tests | `core/database/effects/lists_test.go` (lines 61, 63, 81), `registry_completeness_test.go` (line 185), `core/database/graphcache/graph_test.go` (lines 456, 626), `evidence_graph_memo_test.go` (line 77), `core/database/graphcache/graphcachetest/graphcachetest.go` (line 322), `core/writes/run_test.go` (line 350), `core/catalogsession/warm_test.go` (line 97), `api/ffi/handlers/write_effects_test.go` (line 77) | Name `ListVocabulary` |
| Swift parity | `api/ffi/swift_vocabulary_test.go` "catalog-wide lists" (`listStrings`, `CatalogList` in `swiftCatalogMut`) | Every `effects.Lists` value has a `CatalogList` case |
| Source type / field handlers | `api/ffi/handlers/source_defs.go` `CreateSourceType`, `UpdateSourceType`, `DeleteSourceType`, `CreateMetadataField`, `UpdateMetadataField`, `DeleteMetadataField` | Discard the `writes.Result` (`got, _, err :=`) |
| Type–field join | `source_defs.go` `AssignTypeField`, `RemoveTypeField`, `parseSuggestionJoin(userID, typeID, fieldID)`; `core/database/sourcevocab/sourcevocab.go` `AppendSuggestion`, `DeleteSuggestion` | Validate `user_id`, never use it; plain `db.Exec` outside `writes.Run`, unaudited |
| Property / binding handlers | `api/ffi/handlers/subject_defs.go` `CreateProperty`, `UpdateProperty`, `DeleteProperty`, `AssignSubjectTypeProperty`, `RemoveSubjectTypeProperty` | Discard the `writes.Result`; `ReorderSubjectTypeProperties` already returns `Effects: writeEffectsProto(wrote)`, asserted non-nil in `subject_defs_test.go` (around line 259) |
| Term handlers | `api/ffi/handlers/property_terms.go` `CreatePropertyTerm`, `UpdatePropertyTerm`, `DeletePropertyTerm` | Discard the `writes.Result` |
| Dead RPC handlers | `thumbnails.go` `EnsureFileThumbnail` (its name is also in the `enrichSourceProto` comment); `subject_defs.go` `ListProperties`, `ListSubjectTypeProperties`, `ListPlaceableSubjectTypes`, `GetSubjectTypePresentation`; `identity.go` `RemoveInstallIdentity`; `subjects.go` `ListSubjectTypes`, `ClearSubjectPosition`; `api/ffi/dispatch.go` constants and cases; `METHOD_…` 5, 41, 45, 51, 53, 57, 60, 61 | No Swift caller, except `ListSubjectTypes` and `ClearSubjectPosition`, whose Swift store methods only `macos/ProvenenciaTests/SubjectStoreTests.swift` calls |
| Test fixtures on dead RPCs | `property_terms_test.go` `propertyIDByKey` (unmarshals by hand); inline `ListProperties` / `ListSubjectTypes` lookups in `citations_test.go` (43, 75, 115, 633, 927), `connect_test.go` (32, 79, 186, 215), `conclusion_detail_test.go` (441), `person_detail_test.go` (37), `promote_test.go` (135), `subjects_test.go` (34); `subject_defs_test.go` `TestReorderSubjectTypeProperties`'s `list` helper (219, `ListSubjectTypeProperties`); `subjects_test.go` `TestSubjectPositions` "set list clear persist across session close" (357; its clear half calls `ClearSubjectPosition`, ~437); `subjects_test.go` `subjectFixture` (18, `ListSubjectTypes`); `harness_test.go` `mustRPC` (118) is the reader D-11's helpers use | Read seeded ids through RPCs this PR deletes; `GetPropertiesWorkspaceResponse` carries `properties` (1), `types` (2) and `groups` (3, `SubjectTypePropertiesGroup` per type) |
| Tests on dead RPCs that also cover kept ones | `subject_defs_test.go` `TestListProperties` (:13, "lists seeded properties": `name`, `toponym`, `gender`, `interests`, likeness keys) and `TestSubjectTypePropertiesAndRegistry` (:111, driven by `runRPC` on `ListSubjectTypes`; "bindings locked placeable connect" :114: `participation` locked bindings via `ListSubjectTypeProperties` :135, placeable count via `ListPlaceableSubjectTypes` :154, `ListConnectRules` ≥ 3 rules :165–173); dedicated: `subjects_test.go` `TestListSubjectTypes` (:58), `TestClearSubjectPositionBadProto` (:515), `identity_test.go` `TestRemoveInstallIdentity` (:44), `thumbnails_test.go` `TestEnsureFileThumbnail` (:60) | Their assertions on seeded Properties, locked bindings and connect rules outlive the RPCs |
| Callees of dead RPCs | `core/database/subjectvocab/subjectvocab.go` `PlaceableTypes` (only the deleted `ListPlaceableSubjectTypes` calls it), `PresentationFor` (no non-test caller); `core/database/subjectpositions/subjectpositions.go` `Clear`, `sqlClear` | Called outside their package only by the deleted handlers and by tests: `subjectvocab_test.go` "registry placeable presentation connect", `subjectpositions_test.go` "get missing is tray; clear returns to tray", `core/database/connect/connect_test.go` "missing position invalid" |
| Kept callees | `identity.Remove` (`core/identity/identity.go`), `thumbnailPass.need` | Still called by `api/ffi/handlers/session.go` (sign-out), `artifacts.go`, `sources.go` |
| Nav counts | `api/ffi/handlers/workspace_counts.go` `GetWorkspaceNavCounts`, `typeOriginProto`, `fieldOriginProto`; proto `GetWorkspaceNavCountsResponse` fields 2, 3, `VocabularyOriginCounts`; `sourcetypes.CountByOrigin`, `metadatafields.CountByOrigin`, `properties.CountByOrigin` (no caller) and their `OriginCounts`, `sqlCountByOrigin` | Two origin counts per recount that nothing draws |
| Swift store | `macos/App/Platform/GenealogyStore.swift` (the twelve vocabulary writes, `assignTypeField` / `removeTypeField`, `listSubjectTypes`, `clearSubjectPosition`, `workspaceNavCounts` doc, `WorkspaceNavCounts`, `WorkspaceNavOriginCounts`); `GoStore.swift` (same, `mapOriginCounts`); `FakeStore.swift` (same, `listSubjectTypesCalls`, `originCounts(from:)`); `CoreInvoke.swift` (`listSubjectTypes`, `clearSubjectPosition`) | Vocabulary writes return a row, a list or nothing. S14-16's Source writes return `(Row, CatalogWriteEffects)` / `CatalogWriteEffects`, `FakeStore` returns `nextWriteEffects` |
| Mutations | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogMutation`, `CatalogMutationKind`, `invalidationKind`, `CatalogList` | Thirteen vocabulary cases; `CatalogList.vocabulary = "vocabulary"` |
| Registry | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `conclusionTriggers` (:91, doc "until S14-17"), the descriptors of `sourceTypesList`, `metadataFieldsList`, `sourceWorkspace` (comment "plain create/relabel of a type or field leaves the page alone"), `typeSuggestions` (comment "relabelling or deleting a field reaches types the edit never named"), `propertiesWorkspace` (`refreshes` :346–348: `.list(.vocabulary)`, `.list(.propertyUsage)`, `.structure`, from S14-27), `propertyTerms`, and the seven Conclusion kinds (`personsList`, `eventsList`, `placesList`, `conclusionDetail`, `personClaims`, `personFamily`, `personEvents`; :494–614) | Vocabulary keys list their mutation kinds in `invalidateOn`; Conclusion kinds use `conclusionTriggers == [.updatedProperty]`; S14-27 deleted `invalidations(for list:project:)` and `.deletedSubject` |
| Follow slots | `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` `CatalogQueryFollows`, `CatalogQueryRefresh.applyHiddenPolicyOnAny(follows:)` | Follows `handle`, `source`, `reads`, `touchedHandles`, `evidenceGraphs`, and (S14-27) `.list(CatalogList)` and `.structure`, both `isCatalogWide`; the id-reading constructors trap on a catalog-wide slot (`requireRowIDs`) |
| Callers | `macos/App/Features/SourceTypes/SourceTypesModel.swift` (`submit`, `assignPickedField`, `removeSuggestion`, `confirmPendingImpact`); `macos/App/Features/Metadata/MetadataModel.swift` (delete, create, update); `macos/App/Features/Properties/PropertiesModel.swift` (`submitEdit`, `setCardinality`, delete, the create flow — `createProperty`, then `assignSubjectTypeProperty` per `draft.bindTypeIDs`, then one `apply(.createdProperty)` or `apply(.mutatedSubjectTypeProperties)` after the loop — and `toggleBinding`); `macos/App/Features/CitationComposer/CitationComposerModel.swift` `confirmCustomTerm` | Apply mutations |
| Vocabulary counts | `macos/App/Features/Catalog/CatalogCounts.swift` (`CatalogCountSummary`, `CatalogCounts.metadata` / `.sourceTypes` / `.properties`, `publish*`, `badge(for:)`, class and summary doc comments); `SourceTypesModel.publishCounts`, `countLine`; `MetadataModel.syncCatalogCounts`, `countLine`; `PropertiesModel.syncCatalogCounts`; `.onSettled(…) { model.syncCatalogCounts() }` in `macos/App/Features/Metadata/MetadataView.swift` and `macos/App/Features/Properties/PropertiesView.swift`; `catalogCounts:` on the three models and on `macos/App/Features/SourceTypes/SourceTypesView.swift`, `MetadataView`, `PropertiesView`, `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `macos/App/Features/Workspace/WorkspaceContent.swift`, and `macos/App/Features/Workspace/WorkspaceView.swift` (`.environment(catalogCounts)` :70 and the `catalogCounts:` argument it passes to `WorkspaceContent` :62; it also passes `catalogCounts:` to `WorkspaceSidebar` :56, which `WorkspaceSidebar.swift:18` holds for `WorkspaceSidebarSections.make(counts:)` and which stays) | Configure rows draw no badge (`WorkspaceSidebarSections.swift:46` `count: nil`); `properties` is read only by `badge(for:)`, which no row calls, and by tests; the two `countLine`s fall back to their rows; no view reads `CatalogCounts` from the environment. `macos/App/Features/CatalogVocabulary/VocabularyPane.swift` owns `seededCount` / `userCount` / `pluginCount` on `CatalogVocabularyRow` arrays |
| Docs | `docs/macos-client-patterns.md` rule 2 (:88–89, S14-27's `.list(…)` / `.structure` entries), "Conclusion keys" (:99, "`.updatedProperty`, until S14-17" :101), "One cache owns each list" (:117); `.agents/skills/add-workspace-place/SKILL.md` lines 41, 156, 174, 176 ("Vocabulary CRUD: registry `invalidateOn` + list patch where cheap"); `.agents/skills/add-ffi-handler/SKILL.md` line 57 (the effects rule; names no vocabulary); `core/database/effects/README.md` :15 `Set.Lists` paragraph | Name `vocabulary`, `.updatedProperty` / `conclusionTriggers` "until S14-17", `publish*` |
| Swift tests | `CatalogQueryRegistryTests.swift` (`blanketTriggersAreGone`, `applyFieldCRUDInvalidatesMetadataList`, `applyTypeCRUDInvalidatesTypesList`, `applySuggestionMutationsInvalidateTypeKey`, the `.updatedProperty` assertions at 894, 923, 959; S14-27's `.vocabulary` uses in `catalogWideSlotsReportOneToken` (1189), `propertiesWorkspaceFollowsListsAndStructure` (1201), `listsAndStructureCases` (1231) and `hiddenPropertiesWorkspaceIsMarkedStaleByAList` (1258)); `WorkspaceSessionTests.swift` `aKeyInvalidatedAndRefreshedRevalidatesOnce` (1001) and the `[.vocabulary]` apply at 1478 in `countsListRecounts`; `EvidenceGraphStoreTests.swift` (lines 51, 52); `SourceTypesModelTests.swift` (`loadPopulatesTypesPoolAndCounts`, `assigningAppendsTheFieldAndUpdatesTheListCount`, `removingASuggestionNoSourceUsesOmitsTheValueCount`, and the count assertions in `addingMintsTheKeyFromTheLabelAndLandsOnTheNewType` (246–247) and `confirmingDeleteDropsTheTypeAndItsSuggestions` (487–488)); `MetadataModelTests.swift` (`loadPopulatesFieldsAndCounts`, count assertions); `PropertiesModelTests.swift` (`createPublishesTheCountOnceTheReloadSettles`, `deletePublishesTheCountOnceTheReloadSettles`); `CatalogCountsTests.swift` (`refreshAllPopulatesEveryDestination`, `publishReplacesVocabularySummariesWithoutStoreRoundTrip`, `summaryFromRowsSplitsByOrigin`); `SubjectStoreTests.swift` (`setPositionPersistsAcrossSessionCloseThenClear`, `listSubjectTypesReturnsSeededRows`); `OnboardingModelTests.swift` (its store stub); `WorkspaceDestinationHostTests.swift`; `CitationComposerModelTests.swift` | Cover today's mutation-driven behavior |

**Precedent.** S14-16 (PR [#686](https://github.com/mendahu/provenencia/pull/686)): `823f243` (`WriteEffects effects` on responses, `Effects: writeEffectsProto(wrote)`), `37fb20f` (Swift store `(Row, CatalogWriteEffects)`, callers `session.apply(effects, …)`), `794f797` (`Self.mapWriteEffects(resp.effects)`, no `hasEffects` guard), and `4f47e5a` (a new `effects.List`, `ListCounts`, with its registry entries, `Lists` order, Swift `CatalogList` case and parity row: copy it for the two new lists). `ReorderSubjectTypeProperties` in `subject_defs.go` is the in-file handler shape. Reserve methods the way S14-16 reserved 85. Do not copy `effectsPending`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The fourteen audited vocabulary handlers keep the `writes.Result` and set `Effects: writeEffectsProto(wrote)`. Proto: `WriteEffects effects = 2;` on `CreateSourceTypeResponse`, `UpdateSourceTypeResponse`, `CreateMetadataFieldResponse`, `UpdateMetadataFieldResponse`, `CreatePropertyResponse`, `UpdatePropertyResponse`, `CreatePropertyTermResponse`, `UpdatePropertyTermResponse`; `WriteEffects effects = 1;` on `DeleteSourceTypeResponse`, `DeleteMetadataFieldResponse`, `DeletePropertyResponse`, `DeletePropertyTermResponse`, `AssignSubjectTypePropertyResponse`, `RemoveSubjectTypePropertyResponse` | Working rule; `writes.Run` already resolves each write's effects | A hand-built `WriteEffects` in handlers — a second source beside the registry |
| D-2 | `AssignTypeFieldRequest` and `RemoveTypeFieldRequest` get `reserved 2; reserved "user_id";` in place of `user_id`. `parseSuggestionJoin(typeID, fieldID string)` loses its user parameter. Swift `assignTypeField` / `removeTypeField` lose `userID:` in `GenealogyStore`, `GoStore`, `FakeStore`, the `OnboardingModelTests` stub and `SourceTypesModel`. The writes stay outside `writes.Run`; `SourceTypesModel` keeps `.assignedTypeSuggestion(typeId:)` / `.removedTypeSuggestion(typeId:)` | User, 2026-11-02; `catalog-deletes.md`: the join is never audited; `add-ffi-handler`: an unaudited write has no effects | Auditing the join (refused); keeping `user_id` with a test exception — an allow-list under another name |
| D-3 | `effectsPending` is deleted. `missingEffects(file protoreflect.FileDescriptor) []string` drops the `pending` parameter and both pending branches. `TestEveryAuditedWriteReturnsEffects` keeps the subtests "a response without effects fails naming the request" and "a response with effects passes" and deletes the three pending subtests | S14-16 **Next adds** | An empty map kept "for later" |
| D-4 | Swift store: `createSourceType`, `updateSourceType` → `(CatalogSourceType, CatalogWriteEffects)`; `deleteSourceType` → `CatalogWriteEffects`; `createMetadataField`, `updateMetadataField` → `(CatalogMetadataField, CatalogWriteEffects)`; `deleteMetadataField` → `CatalogWriteEffects`; `createProperty`, `updateProperty` → `(CatalogProperty, CatalogWriteEffects)`; `deleteProperty` → `CatalogWriteEffects`; `createPropertyTerm` → `(CatalogPropertyTerm, CatalogWriteEffects)`; `assignSubjectTypeProperty`, `removeSubjectTypeProperty` → `CatalogWriteEffects`. `GoStore` maps with `Self.mapWriteEffects(resp.effects)`. `FakeStore` returns `nextWriteEffects` from each; the `OnboardingModelTests` stub returns `.empty` for effects-only methods and keeps throwing for the rest. No Swift method for `UpdatePropertyTerm` / `DeletePropertyTerm` | S14-16 shape; "FakeStore is a seeded fake" | A `FakeStore` that builds the family effects itself |
| D-5 | Go: `ListVocabulary` is replaced by `ListSourceVocabulary List = "source_vocabulary"` (on `source_types`, `source_metadata_fields`) and `ListPropertyVocabulary List = "property_vocabulary"` (on `subject_types`, `properties`, `property_terms`, `subject_type_properties`). `Lists` becomes `[ListSourceVocabulary, ListPropertyVocabulary, ListPropertyUsage, ListCounts]`. `graphcache.OnCommit` drops labels, the vocabulary and the derived memos only when the set has `ListPropertyVocabulary`: nothing it holds comes from Source types or metadata fields (`graphcache.Vocabulary` is Properties, terms, types and bindings). The wire key `vocabulary` is retired (one build ships Go and Swift together). Swift `CatalogList.vocabulary` becomes `.sourceVocabulary = "source_vocabulary"` and `.propertyVocabulary = "property_vocabulary"`; the parity test covers both. The test files of §4 "List tests" use the family each case writes (`registry_completeness_test.go` line 185 uses `ListPropertyVocabulary`; `run_test.go` uses the family of the table its case writes; the commits with no table (`graph_test.go` 456 and 626, `evidence_graph_memo_test.go` 77, `warm_test.go` 97, `graphcachetest.go` 322) use `ListPropertyVocabulary`, and `Canon.CommitVocabulary` keeps its name with its doc reading "a Property-side vocabulary label was written"; `write_effects_test.go` `TestWriteEffectsProtoLists` (77–78) uses `ListPropertyVocabulary` and wants `"property_vocabulary"`). Only the `CatalogList` case is renamed: `CitationComposerVocabulary` and `model.vocabulary` / `context.vocabulary` in the composer are a different symbol and stay | User, 2026-11-02: two lists, so a Source-side write reaches no Property-side key and the reverse; "One cache owns each list" stands for pages. Copies `4f47e5a` | One `vocabulary` list reaching every label key (bends "One cache owns each list"); one list per table (user chose two) |
| D-6 | Swift keys follow the family through S14-27's slot, named here `CatalogQueryFollows.list(CatalogList)`, as a `refreshes` entry `.applyHiddenPolicyOnAny(follows: .list(…))` (every warmed key of the kind takes its hidden policy). `source_vocabulary`: `sourceTypesList`, `metadataFieldsList`, `typeSuggestions`. `property_vocabulary`: `propertiesWorkspace`, `personsList`, `eventsList`, `placesList`, `conclusionDetail`, `personClaims`, `personFamily`, `personEvents`. S14-27's `propertiesWorkspace` entry `.applyHiddenPolicyOnAny(follows: .list(.vocabulary))` (`CatalogQueryRegistry.swift:346`) becomes `.list(.propertyVocabulary)`. Deleted: `CatalogMutation` cases `createdSourceType`, `updatedSourceType(id:)`, `deletedSourceType(id:)`, `createdMetadataField`, `createdProperty`, `updatedProperty`, `deletedProperty`, `mutatedSubjectTypeProperties`, `createdPropertyTerm(propertyId:)`, their kinds and `invalidationKind` arms; `conclusionTriggers`. Kept with a narrower reach: `.updatedMetadataField(id:)` and `.deletedMetadataField(id:)` invalidate only `.allCached(.sourceWorkspace)`; `metadataFieldsList` and `typeSuggestions` drop them from `invalidateOn`. New `invalidateOn`: `sourceTypesList` `[.createdSource, .deletedSource, .changedSourceType]`; `metadataFieldsList` `[.mutatedSourceMetadata]`; `typeSuggestions` `[.assignedTypeSuggestion, .removedTypeSuggestion]`; `propertiesWorkspace` `[]` (S14-27 already removed `.deletedSubject`); `propertyTerms` `[]` with `invalidation: { _, _ in .allCached(.propertyTerms) }` and one `refreshes` entry `.applyHiddenPolicy(follows: .reads(.propertyTerms)) { _, key in [propertyId] }` (its rows hold no Property label, binding or type, so it follows the keyed read D-14 adds, not `property_vocabulary`); `typeSuggestions`' `invalidation` closure stays as it is; the seven Conclusion kinds `[]`. `sourceWorkspace`'s set is unchanged | Each key follows the labels it holds; a field relabel or delete restates every Source page's derived rows, which a list cannot tell from a field create, so the mutation that names them stays (rule 2's "keys the effects cannot name") | A Source page following `source_vocabulary` — a type create would mark every visited Source page stale |
| D-7 | Callers apply each write's effects when that write returns: `SourceTypesModel` add, edit, delete → `session.apply(effects)`; `MetadataModel` add → `apply(effects)`, edit → `apply(effects, .updatedMetadataField(id:))`, delete → `apply(effects, .deletedMetadataField(id:))`; `PropertiesModel.submitEdit`, `setCardinality`, delete, `toggleBinding` → `apply(effects)`; the Properties create flow applies the create's effects right after `createProperty` returns, then each assign's effects after it returns; a thrown assign leaves the create applied and shows `formError` as today; `CitationComposerModel.confirmCustomTerm` → `apply(effects)`, then `await warmTerms(for:)` as today. Patches after `apply` stay | Rule 2; a write's effects describe that write, so they apply when it lands | One `apply` after the loop — a failed assign would leave a created Property unshown, as today |
| D-8 | Vocabulary counts go (user, 2026-11-02: "We don't use those counts, so if they're not in use, let's remove them"). `CatalogCountSummary` is deleted; `SourceTypesModel.countLine` and `MetadataModel.countLine` read `types.count` / `fields.count` and `.seededCount`, `.userCount`, `.pluginCount` (`VocabularyPane.swift`). `CatalogCounts` loses `metadata`, `sourceTypes`, `properties`, `publish*`; `badge(for:)` answers `.sources`, `.persons`, `.events`, `.places` and returns `nil` in one `default` arm; `refreshAll` assigns only sources and handles; the class doc says it holds the sidebar's Research badges. `SourceTypesModel.publishCounts()`, `MetadataModel.syncCatalogCounts()`, `PropertiesModel.syncCatalogCounts()`, both `.onSettled(…) { model.syncCatalogCounts() }`, the `catalogCounts` property and parameter of the three models and views, `WorkspaceDestinationHost.catalogCounts`, `WorkspaceContent.catalogCounts`, the argument `WorkspaceView` passes to `WorkspaceContent`, and `WorkspaceView`'s `.environment(catalogCounts)` go; `WorkspaceView` keeps its `@State` for `WorkspaceSidebar` and the recount. Go: `GetWorkspaceNavCountsResponse` becomes `reserved 2, 3, 4; reserved "source_types", "metadata_fields", "files";`; `VocabularyOriginCounts` is deleted; `typeOriginProto`, `fieldOriginProto`, and `CountByOrigin` / `OriginCounts` / `sqlCountByOrigin` in `sourcetypes`, `metadatafields`, `properties` with their test cases go. Swift: `WorkspaceNavCounts.sourceTypes` / `.metadataFields`, `WorkspaceNavOriginCounts`, `GoStore.mapOriginCounts`, `FakeStore.originCounts(from:)` go; the `workspaceNavCounts` doc says "sidebar totals for Sources and unmerged handles" | No view draws a Configure badge; the headers already count their rows; one place owns the origin split (`VocabularyPane.swift`) | Reporting `counts` from vocabulary writes; keeping `CatalogCountSummary` as a forwarder |
| D-9 | Eight RPCs go: `RemoveInstallIdentity` (5), `EnsureFileThumbnail` (41), `ListSubjectTypes` (45), `ClearSubjectPosition` (51), `ListProperties` (53), `ListSubjectTypeProperties` (57), `ListPlaceableSubjectTypes` (60), `GetSubjectTypePresentation` (61). One new pair after S14-16's in the `Method` enum: `reserved 5, 41, 45, 51, 53, 57, 60, 61;` and `reserved` with the eight `METHOD_…` names, under the comment `// S14-17: RPCs with no Swift caller.`. Their request and response messages, handlers, `dispatch.go` constants and cases go. `SubjectTypePresentation`, `SubjectType`, `Property`, `SubjectTypeProperty` stay. `RemoveInstallIdentity` is not planned: nothing in `docs/ideas/`, `docs/*.md` or the open spikes names it, and sign-out removes the identity through `identity.Remove` in `session.go`. The `enrichSourceProto` comment names `SetSourceCover` and the Artifact reads as `thumbnailPass` callers | Audit A-35; the map's Check: no RPC without a Swift caller except three — `ListSubjectTypes` and `ClearSubjectPosition` lose their last with D-10 | Leaving those two |
| D-10 | Swift: `GenealogyStore.listSubjectTypes`, `clearSubjectPosition`, their `GoStore` / `FakeStore` bodies, `FakeStore.listSubjectTypesCalls`, `CoreMethod.listSubjectTypes` / `.clearSubjectPosition`, the `OnboardingModelTests` stubs go. `SubjectStoreTests.listSubjectTypesReturnsSeededRows` is deleted; `setPositionPersistsAcrossSessionCloseThenClear` keeps its set-and-reopen half as `setPositionPersistsAcrossSessionClose` | Map scope entry | — |
| D-11 | Go fixtures: `propertyIDByKey(t, dir, key) string` moves to `harness_test.go` and reads through `mustRPC(t, GetPropertiesWorkspace, …)`; `subjectTypeIDByKey(t, dir, key) string` *(new)* sits beside it with the same read. Every by-key lookup of §4 uses one of the two; a test reads `GetPropertiesWorkspaceResponse.properties` / `.types` directly only where it walks the whole list (`connect_test.go` 79 and 215). `TestListProperties`'s case and `TestSubjectTypePropertiesAndRegistry`'s case move into `TestGetPropertiesWorkspace` (`subject_defs_test.go:191`), and both old tests are deleted: "lists seeded properties" reads `resp.Properties` with the same key checks; "bindings locked placeable connect" becomes "locked bindings and connect rules": the `participation` group's locked bindings via `resp.Groups` (same count as today), the three palette types via `group.Presentation.Placeable`, and `ListConnectRules` ≥ 3 rules unchanged. Dead callees go: `subjectvocab.PlaceableTypes`, `subjectvocab.PresentationFor` and the "registry placeable presentation connect" case except its connect-rule assertions, which stay as "connect rules"; `subjectpositions.Clear` and `sqlClear` with the clear half of "get missing is tray; clear returns to tray" (renamed "get missing is tray"); the package comment "Set and Clear" reads "Set". `TestReorderSubjectTypeProperties`'s `list` helper reads the `person` type's bindings from `GetPropertiesWorkspace`'s `groups` (the group whose type id is `personID`), in the same order. `TestSubjectPositions` "set list clear persist across session close" drops its clear half and is renamed "set list persist across session close". `connect_test.go` "missing position invalid" removes the row with `db, err := s.c.DB()` then `db.Exec("DELETE FROM subject_positions WHERE subject_id = ?", s.place.ID)` outside `writes.Call` (`writes.Run` rolls back a closure that returns no changes) | One helper for one lookup (`add-ffi-handler`: shared helpers in `harness_test.go`); a callee whose last caller goes is deleted | A test-only RPC; inline lookups per test |
| D-12 | No new copy, no new error codes, no migration | Behavior and wire only | — |
| D-13 | Guidance: rule 2 lists as exceptions only the Source-row mutations (`.createdSource`, `.updatedSource`, `.changedSourceType`, `.deletedSource`), `.mutatedSourceWorkspace`, `.mutatedSourceMetadata`, `.updatedMetadataField` / `.deletedMetadataField` (the Source pages a field relabel or delete restates) and the type–field join (unaudited, no effects), and says a vocabulary write applies its effects, whose `source_vocabulary` or `property_vocabulary` each label key follows. "Conclusion keys" first bullet: they follow `touchedHandles`, and `property_vocabulary` for a label change; no `conclusionTriggers`. "One cache owns each list": a Source-vocabulary write reloads the Source vocabulary lists and no page; a field relabel or delete marks the Source pages stale. `add-workspace-place` line 41 drops `syncCatalogCounts`; line 156 "Sidebar badges": they come only from the recount, and Configure rows have none; line 174 "writes that still report none" names only the type–field join; line 176 reads "Vocabulary CRUD: apply the write's effects; a key that holds a vocabulary label follows its family list". `add-ffi-handler` line 57 adds "A write that is not audited takes no `user_id`." `core/database/effects/README.md` names the two families and the `property_terms` read. The `sourceWorkspace` and `typeSuggestions` descriptor comments say what they follow | Sweep rule (review M-4) | — |
| D-14 | Go: `ReadPropertyTerms Read = "property_terms"` *(new)* in `core/database/effects/registry.go`; `KeyTable()` returns `"properties"` for it; `Reads` appends it; the `property_terms` entry gets `Reads: []ReadPath{{ReadPropertyTerms, field("property_id")}}`. Swift: `CatalogRead.propertyTerms = "property_terms"` *(new)*; the "catalog reads" parity row covers it. A term create, update or delete reports `reads: {property_terms: [propertyId]}`, so only that Property's `propertyTerms` key reloads, as `.createdPropertyTerm(propertyId:)` did | The deleted mutation named one key; a list would reload every warmed term list on any Property-family write. Copies `ReadEvidenceGraphs` (`field("source_id")`) and the README recipe for a keyed read | `propertyTerms` following `property_vocabulary` (plan review M-10) |

**Rules this plan bends.** None. The map's scope line "The source type and metadata field writes report `counts` (Go `effects.ListCounts` on their registry entries), so their sidebar badges come from the recount" is replaced by D-8 on the user's answer (2026-11-02); the scope entry was updated with this plan.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ListSourceVocabulary`, `ListPropertyVocabulary` *(new)*; Swift `.sourceVocabulary`, `.propertyVocabulary` *(new)* | effects list | `ListVocabulary`, which they replace; `ListCounts` (`4f47e5a`) | Replace, not beside: `ListVocabulary` goes |
| `effects` field on 14 responses *(new)* | proto field | `ReorderSubjectTypePropertiesResponse.effects` | Same shape |
| `TestVocabularyWritesReturnEffects` in `api/ffi/handlers/vocabulary_write_effects_test.go` *(new)* | Go test file | `api/ffi/handlers/source_write_effects_test.go` | One table for every vocabulary write, `ReorderSubjectTypeProperties` included (its non-nil assertion in `subject_defs_test.go` moves here) |
| `subjectTypeIDByKey` *(new)* | test helper | `propertyIDByKey`, moved beside it into `harness_test.go` | Its sibling |
| `ReadPropertyTerms` / `CatalogRead.propertyTerms` *(new)* | keyed read | `ReadEvidenceGraphs` / `CatalogRead.evidenceGraphs` | Same mechanism, one more read |
| `refreshes` entries on twelve descriptors | descriptor slot use | S14-27's list slot | Uses it as designed |

Searched: `rg 'Effects: writeEffectsProto' api/ffi/handlers`, `rg 'IDByKey|mustRPC' api/ffi/handlers`, `rg 'CountByOrigin|OriginCounts|CatalogCountSummary'`, `rg 'seededCount'`, `rg 'catalogCounts'`, `rg 'ListVocabulary|\.vocabulary\b'`.

**Builds on.** `writeEffectsProto` and `writes.Run`; the effects registry's `Lists` (S14-16's `ListCounts` precedent); S14-27's descriptor list slot; S14-16's Swift store shape and `GoStore.mapWriteEffects`; `VocabularyPane.swift`'s origin counts; `add-ffi-handler` reservation.
**Deletes.** `ListVocabulary` and Swift `.vocabulary`; `effectsPending`, `missingEffects`'s `pending` and three subtests; nine `CatalogMutation` cases, kinds, arms; `conclusionTriggers`; `CatalogCountSummary`; `CatalogCounts.metadata` / `.sourceTypes` / `.properties` / `publish*`; `SourceTypesModel.publishCounts`; two `syncCatalogCounts`; two `.onSettled` hooks; the `catalogCounts` pass-through (three models, three views, host, content, the `WorkspaceView` argument and `.environment`); `WorkspaceNavCounts.sourceTypes` / `.metadataFields`; `WorkspaceNavOriginCounts`; `GoStore.mapOriginCounts`; `FakeStore.originCounts(from:)`; proto `VocabularyOriginCounts`, nav fields 2 and 3; `typeOriginProto`, `fieldOriginProto`; `CountByOrigin` / `OriginCounts` / `sqlCountByOrigin` ×3; eight RPCs; `subjectvocab.PlaceableTypes`, `subjectvocab.PresentationFor`; `subjectpositions.Clear` / `sqlClear`; Swift `listSubjectTypes`, `clearSubjectPosition`, `listSubjectTypesCalls`, two `CoreMethod` constants; `user_id` on two requests and `userID:` on two store methods; the inline id lookups; the "until S14-17" doc lines.
**Next adds.** A new vocabulary table sets `Lists: []List{ListSourceVocabulary}` or `ListPropertyVocabulary` on its registry entry and reaches every key of its family with no Swift change. A new key that holds a label adds one `refreshes` entry on its descriptor.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Property-family write now reaches the Conclusion keys (create and delete too, beside today's update) | Conclusion descriptors (`.markStale` lists, `.evict` details) | Hidden on Properties and in the composer: lists mark stale, details evict; a Property delete is refused while in use (`properties.Delete`) |
| A Source type or field create now reaches `metadataFieldsList` / `sourceTypesList` / `typeSuggestions` beside its own list | those descriptors | Small lists reload from the catalog; a deleted type's `typeSuggestions` key reloads to the engine's answer, as on `.deletedSourceType` today |
| A composer custom term reaches the on-screen `propertiesWorkspace` | `CitationComposerView` | Reloads in place; the composer's draft lives in `CitationComposerModel`, so the open term sheet and draft stay (T-10) |
| The Properties create flow applies the create's effects before the binds | `PropertiesView` | The new row shows while binds run; a failed bind shows `formError` with the row present (T-9) |

### Scale

At the map's Scale (3,000 Sources; 37,000 Subjects; 92,000 Observations; 200,000 audit rows):

- **Source-family write** (Source types, Metadata pages). Reloads `sourceTypesList` (~20 rows), `metadataFieldsList` (~40) and warmed `typeSuggestions` keys (tens of rows each). A field relabel or delete marks visited Source pages stale (`.markStale`, no fetch now), as today. No Conclusion or Property key moves.
- **Property-family write** (Properties page). Reloads `propertiesWorkspace` (~60 Properties, 10 types); a term write reloads only its Property's `propertyTerms` key; Conclusion lists (3,000+ Person rows) mark stale and load only when shown; details and Person reads evict. Today `.updatedProperty` did the same.
- **Composer custom term** (the frequent one). Today: the Property's `propertyTerms` key. After: the same key (D-14). After: also the on-screen `propertiesWorkspace`, the same read every Observation save in the composer already reloads through `property_usage`; hidden Conclusion keys stale or evict. No canvas, Source page or Source vocabulary list reloads.
- **Go.** One `writeEffectsProto` copy per handler of a set `writes.Run` already resolved; two `GROUP BY origin` queries fewer per recount.

### Edges

- **Assumptions.** `graphcache.OnCommit` drops labels on `property_vocabulary` only, because it holds no Source vocabulary (T-3 holds both sides). `WorkspaceSession.apply` revalidates a key once when `refreshes` and invalidations both name it (`aKeyInvalidatedAndRefreshedRevalidatesOnce`, kept with effects in place of `.updatedProperty`). `effects.Set.Lists` dedupes in `Lists` order, so a Property delete that releases bindings reports `property_vocabulary` once.
- **Neighbors.** An update that changes nothing (`UpdateSourceType` with the same label, description, icon) returns empty effects and reloads nothing; today it reloaded `sourceTypesList`; the model's `patchTypesList` still writes the returned row (T-1, T-7). A cardinality change reports `handles` and `property_vocabulary`; a label change reports no handles (T-1). `ReorderSubjectTypeProperties` now reports `property_vocabulary` (T-1). The type–field join reports nothing and keeps its mutation (T-14).
- **Removed inputs.** A deleted Property releases its bindings (audited, `property_vocabulary`). A deleted Source type cascades its suggestion joins unaudited; its `typeSuggestions` key follows `source_vocabulary`.
- **Real data.** None — no table-driven rule; the six registry entries only rename their list.
- **Untrusted size.** None — no new list read from data the app did not write.
- **Labels the effects do not reach.** `EvidenceGraphRows` observations and `citationsByArtifact` carry `propertyLabel`; a rename leaves a visited canvas on the old label, as today. Out (structural-debt line).

### Schema

None — no catalog change. The effects registry renames one list into two (D-5); no migration.

### Go core

| Package | Change |
| --- | --- |
| `core/database/effects` | `ListSourceVocabulary`, `ListPropertyVocabulary` *(new)* replace `ListVocabulary`; `Lists` order; six entries; README `Set.Lists`; tests per §4 |
| `core/database/graphcache` | `OnCommit` drops labels when the set has `ListSourceVocabulary` or `ListPropertyVocabulary`; test fixtures per §4 |
| `core/database/sourcetypes`, `metadatafields`, `properties` | Delete `CountByOrigin`, `OriginCounts`, `sqlCountByOrigin` and their test cases |
| `core/database/subjectvocab` | Delete `PlaceableTypes`, `PresentationFor`; test case per D-11 |
| `core/database/subjectpositions` | Delete `Clear`, `sqlClear`; test case and package comment per D-11 |
| `core/database/connect` | `connect_test.go` "missing position invalid" per D-11 |

Errors: none new.

### FFI

`api/proto/engine.proto`:

```proto
// Method, after S14-16's pair:
// S14-17: RPCs with no Swift caller.
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
// D-1 lists the other twelve.
```

Deleted messages: the request and response of each D-9 method, and `VocabularyOriginCounts`. Handlers: `source_defs.go`, `subject_defs.go`, `property_terms.go`, `workspace_counts.go`, `thumbnails.go`, `identity.go`, `subjects.go`; `api/ffi/dispatch.go`. Regenerate with `./scripts/generate-proto.sh` (needs `protoc` and the Go and Swift plugins: a macOS machine with the repo's toolchain).

### macOS

None — no chrome change. Store, registry and model changes are D-4, D-6 to D-8 and D-10.

### Strings

None — no key added or removed.

## 7. Steps

Steps 2 to 8 run `./scripts/generate-proto.sh` or `xcodebuild`; both need the macOS toolchain (see Stop conditions).

**STEP-1 — Go reports the vocabulary family**
Files: `core/database/effects/registry.go`, `core/database/effects/README.md`, `core/database/effects/lists_test.go`, `core/database/effects/registry_completeness_test.go`; `core/database/graphcache/commit.go`, `core/database/graphcache/graph_test.go`, `core/database/graphcache/evidence_graph_memo_test.go`, `core/database/graphcache/graphcachetest/graphcachetest.go`, `core/writes/run_test.go`, `core/catalogsession/warm_test.go`, `api/ffi/handlers/write_effects_test.go` (line 77).
Do: D-5's Go half and D-14's Go half. T-2 first.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/ffi/handlers/...` passes; `rg -n ListVocabulary core api` prints nothing. The Swift parity test (`api/ffi` "catalog-wide lists", "catalog reads") fails until STEP-2 renames `CatalogList` and adds `CatalogRead.propertyTerms`; STEP-2's check runs it.

**STEP-2 — Vocabulary handlers return their effects; the allow-list goes**
Files: `api/proto/engine.proto` (+ regenerated `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`); `api/ffi/handlers/source_defs.go`, `subject_defs.go`, `property_terms.go`, `write_effects_test.go`, `vocabulary_write_effects_test.go` *(new)*, `source_defs_test.go`, `subject_defs_test.go`; `api/ffi/swift_vocabulary_test.go`; Swift for D-2 and D-5's Swift half: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Features/SourceTypes/SourceTypesModel.swift`, `macos/App/Features/Workspace/Session/CatalogMutation.swift` (`CatalogList`; `CatalogRead.propertyTerms`, D-14), and every Swift use of `.vocabulary` (S14-27's descriptor entry, `WorkspaceSessionTests`, `EvidenceGraphStoreTests`, `CatalogQueryRegistryTests`) renamed to `.propertyVocabulary`; `macos/ProvenenciaTests/OnboardingModelTests.swift`, `SourceTypesModelTests.swift`. Skill: `add-ffi-handler`.
Do: T-1 first. D-1, D-2, D-3; T-3's `user_id` lines; drop `UserId` from the assign and remove requests in `source_defs_test.go`; move the reorder effects assertion into T-1.
Check: `./scripts/generate-proto.sh`; `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; `rg -n effectsPending api` prints nothing; `xcodebuild test` (as `scripts/check-pr.sh` runs it) passes.

**STEP-3 — Each label key follows its family**
Files: `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`. Skill: `add-workspace-place`.
Do: T-5 first. D-6's `refreshes` entries and D-14's `propertyTerms` entry; the mutations still exist in this step.
Check: T-5 passes under `xcodebuild test`.

**STEP-4 — The Swift store returns vocabulary effects; callers apply them**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Features/SourceTypes/SourceTypesModel.swift`, `macos/App/Features/Metadata/MetadataModel.swift`, `macos/App/Features/Properties/PropertiesModel.swift`, `macos/App/Features/CitationComposer/CitationComposerModel.swift`; tests `OnboardingModelTests.swift`, `SourceTypesModelTests.swift`, `MetadataModelTests.swift`, `PropertiesModelTests.swift`, `CitationComposerModelTests.swift`. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-4, D-7. In this step each caller applies `apply(effects, .old)` with its current mutation, so nothing it reloads today is lost before STEP-5.
Check: T-7 to T-10 pass; `xcodebuild test` passes.

**STEP-5 — The mutations the effects cover go**
Files: `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, the four models of STEP-4; tests `CatalogQueryRegistryTests.swift`, `WorkspaceSessionTests.swift`. Skill: `add-workspace-place`.
Do: D-6's deletions and `invalidateOn` sets; callers drop the deleted mutations. Rewrite `applyFieldCRUDInvalidatesMetadataList` and `applyTypeCRUDInvalidatesTypesList` to apply `CatalogWriteEffects(lists: [.sourceVocabulary])`; T-6; T-11.
Check: `rg -n 'createdSourceType|updatedSourceType|deletedSourceType|createdMetadataField|createdProperty\b|updatedProperty|deletedProperty|mutatedSubjectTypeProperties|createdPropertyTerm|conclusionTriggers' macos` prints nothing; T-6, T-11, T-14 pass; `xcodebuild test` passes with no Swift warning.

**STEP-6 — The vocabulary counts nothing draws go**
Files: `macos/App/Features/Catalog/CatalogCounts.swift`, `macos/App/Features/SourceTypes/SourceTypesModel.swift`, `macos/App/Features/SourceTypes/SourceTypesView.swift`, `macos/App/Features/Metadata/MetadataModel.swift`, `macos/App/Features/Metadata/MetadataView.swift`, `macos/App/Features/Properties/PropertiesModel.swift`, `macos/App/Features/Properties/PropertiesView.swift`, `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `macos/App/Features/Workspace/WorkspaceContent.swift`, `macos/App/Features/Workspace/WorkspaceView.swift`, `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`; `api/proto/engine.proto` (+ regenerated), `api/ffi/handlers/workspace_counts.go`, `workspace_counts_test.go`, `write_effects_test.go`; `core/database/sourcetypes/sourcetypes.go` and `_test.go`, `core/database/metadatafields/metadatafields.go` and `_test.go`, `core/database/properties/properties.go` and `_test.go`; tests `CatalogCountsTests.swift`, `SourceTypesModelTests.swift`, `MetadataModelTests.swift`, `PropertiesModelTests.swift`, `WorkspaceDestinationHostTests.swift`, `WorkspaceSidebarSectionsTests.swift`. Skills: `add-workspace-place`, `add-ffi-handler`, `add-swift-test`.
Do: D-8. Delete `publishReplacesVocabularySummariesWithoutStoreRoundTrip`, `summaryFromRowsSplitsByOrigin`, `createPublishesTheCountOnceTheReloadSettles`, `deletePublishesTheCountOnceTheReloadSettles` and the vocabulary count assertions; T-3's nav-count lines; T-12, T-13.
Check: `rg -n 'publishSourceTypes|publishMetadata|publishProperties|syncCatalogCounts|CatalogCountSummary|WorkspaceNavOriginCounts|VocabularyOriginCounts|CountByOrigin|mapOriginCounts' api core macos --glob '!*.pb.*'` prints nothing; `rg -n catalogCounts macos/App/Features/SourceTypes macos/App/Features/Metadata macos/App/Features/Properties macos/App/Features/Workspace/WorkspaceDestinationHost.swift macos/App/Features/Workspace/WorkspaceContent.swift` prints nothing; Go and Xcode tests pass.

**STEP-7 — RPCs with no Swift caller go**
Files: `api/proto/engine.proto` (+ regenerated), `api/ffi/dispatch.go`, `api/ffi/handlers/thumbnails.go`, `subject_defs.go`, `identity.go`, `subjects.go`, `harness_test.go`, `property_terms_test.go`, the test files of §4 "Test fixtures" and "Tests on dead RPCs", `thumbnails_test.go`, `identity_test.go`, `write_effects_test.go`; `core/database/subjectvocab/subjectvocab.go` and `_test.go`, `core/database/subjectpositions/subjectpositions.go` and `_test.go`, `core/database/connect/connect_test.go`; Swift `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`, `OnboardingModelTests.swift`, `SubjectStoreTests.swift`. Skill: `add-ffi-handler`.
Do: D-9, D-10, D-11; T-3's method lines; T-15.
Check: `rg -n 'EnsureFileThumbnail|GetSubjectTypePresentation|ListPlaceableSubjectTypes|ListPropertiesRequest|ListSubjectTypeProperties|RemoveInstallIdentity|ListSubjectTypesRequest|ClearSubjectPosition|PlaceableTypes\(|PresentationFor\(|subjectpositions\.Clear' api core macos/App --glob '!*.pb.*'` prints only the reserved names in `engine.proto` and T-3's lines; `rg -n 'listSubjectTypes|clearSubjectPosition' macos` prints nothing; `rg -l 'GetPropertiesWorkspaceRequest' api/ffi/handlers/*_test.go` lists only `harness_test.go`, `subject_defs_test.go` and `connect_test.go`; Go and Xcode tests pass.

**STEP-8 — Guidance, version and docs**
Files: `docs/macos-client-patterns.md`, `.agents/skills/add-workspace-place/SKILL.md`, `.agents/skills/add-ffi-handler/SKILL.md`, `core/database/effects/README.md`; `VERSION`, `core/version.go`, `MARKETING_VERSION` in `macos/Provenencia.xcodeproj/project.pbxproj`; `docs/deployment-plan/spike-14/completed.md`, `deployment-plan.md` (S14-17 row status; the scope entry already carries D-5, D-8 and D-9's two extra RPCs), this plan (status, Landed note), `docs/ideas/structural-debt.md` (line: "A Property rename leaves a visited Evidence graph canvas and its citations by Artifact on the old label: `evidenceGraph` and `citationsByArtifact` rows carry `propertyLabel` and follow no vocabulary list, because following `property_vocabulary` would reload the on-screen graph on every composer custom term (S14-17).").
Do: D-13; the patch bump per `docs/versioning.md`; the Definition of done's docs lines.
Check: `rg -n 'until S14-17|conclusionTriggers|syncCatalogCounts|publish\*|Vocabulary CRUD: registry|vocabulary-header|vocabulary destination headers' docs/*.md .agents .cursor core/database/effects/README.md macos/App` prints nothing; `python3 scripts/check-guidance-names.py` passes; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-17-vocabulary-effects.md` passes; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `api/ffi/handlers/vocabulary_write_effects_test.go` `TestVocabularyWritesReturnEffects` *(new)* | Each row starts from `sourceFixture(t)` and works on rows it creates (seeded types, fields, terms and locked bindings refuse edits). Source family: create a user type, update its label, delete it; the same for a user field → `lists == ["source_vocabulary"]`, `revision > 0`. Property family: create a user text Property, update its label, create a term on a user term-valued Property, update the term's label, delete the term (each term row also reports `reads == {"property_terms": [propertyId]}`), assign the user Property to the `person` type, reorder `person`'s bindings, remove the binding, delete the Property → `lists == ["property_vocabulary"]`, `revision > 0`. "A cardinality change names the handles that observe it": `citationFixture`; `CreateProperty` a user text Property; `AssignSubjectTypeProperty` it to the `place` type; `CreateCitationWithObservations` one Observation on the fixture's place Subject with it (`ValueText: "x"`, built the way `filedSubject` builds one); `promoteInDir`; `UpdateProperty` cardinality `single` → `multiple` → `handles` holds the handle, `lists == ["property_vocabulary"]`. "An unchanged update reports nothing": `UpdateSourceType` on a user type with the same label, description, icon → `lists` empty, `revision == 0` | AC-2.2, AC-2.3, AC-1.5, D-1 |
| T-2 | Go | `core/database/effects/lists_test.go` (changed) | A `source_types` and a `source_metadata_fields` change → `[ListSourceVocabulary]`; a `properties`, `property_terms` and `subject_type_properties` change → `[ListPropertyVocabulary]`; both in one batch → both, in `Lists` order | D-5 |
| T-3 | Go | `api/ffi/handlers/write_effects_test.go` `TestRetiredFieldsStayReserved` (extended); `core/database/graphcache/graph_test.go` (changed rows) | `AssignTypeFieldRequest` / `RemoveTypeFieldRequest` reserve 2 / `user_id`; `GetWorkspaceNavCountsResponse` reserves 2, 3 / `source_types`, `metadata_fields`; `Method` reserves D-9's numbers and names; `OnCommit` with `ListPropertyVocabulary` drops labels and sets the warm flag; with `ListSourceVocabulary` alone it keeps the labels, the match index and the warm flag (`warm_test.go`, `graph_test.go`) | AC-2.4, AC-3.1, D-5, D-8, D-9 |
| T-4 | Go | `write_effects_test.go` `TestEveryAuditedWriteReturnsEffects` (changed); `source_defs_test.go` assign / remove; `workspace_counts_test.go` `TestGetWorkspaceNavCounts` | No pending; assign / remove with no `UserId` → today's lists; nav counts → `sources`, `persons`, `events`, `places` | AC-2.1, D-2, D-3, D-8 |
| T-5 | Swift | `CatalogQueryRegistryTests.swift` `listFamiliesReachTheirKeys` *(new)*, parameterized | Warm one key of each D-6 kind; apply `CatalogWriteEffects(lists: [.sourceVocabulary])` → exactly `sourceTypesList`, `metadataFieldsList`, `typeSuggestions` revalidate or take their hidden policy; `[.propertyVocabulary]` → exactly D-6's eight Property-family kinds; `CatalogWriteEffects(reads: [.propertyTerms: ["p1"]])` → only `propertyTerms(p1)` of two warmed term keys; neither touches `sourceWorkspace`, `evidenceGraph`, `citationsByArtifact` | AC-1.1–AC-1.4, D-6, D-14 |
| T-6 | Swift | `CatalogQueryRegistryTests.swift` `noMutationReachesAConclusionKey` *(new)* | For each remaining `CatalogMutation` case, built with `CatalogSource(id: "s1", ref: "SRC-1", sourceTypeID: "", title: "T", description: "")` where it carries a Source, `invalidations(by:)` names no Conclusion kind | D-6 |
| T-7 | Swift | `SourceTypesModelTests.swift` `addingATypeReloadsFromItsEffects` *(new)* | `store.nextWriteEffects = CatalogWriteEffects(lists: [.sourceVocabulary])`, warmed `sourceTypesList`, `metadataFieldsList` and `sourceWorkspace(s1)`; add "Diary" → both lists reload once, `sourceWorkspace(s1)` untouched, the new type selected. Second case: `nextWriteEffects = .empty`, edit with the same label → no reload, row unchanged | AC-1.3, D-7 |
| T-8 | Swift | `MetadataModelTests.swift` `renamingAFieldReloadsItsDependents` *(new)* | `nextWriteEffects = CatalogWriteEffects(lists: [.sourceVocabulary])`, warmed `typeSuggestions(t1)`, hidden `sourceWorkspace(s1)`; rename "Volume" → `metadataFieldsList` and `typeSuggestions(t1)` reload, `sourceWorkspace(s1)` stale | AC-1.2 |
| T-9 | Swift | `PropertiesModelTests.swift` `renamingAPropertyStalesConclusionPages` *(new)*; `createAppliesEachWriteAsItLands` *(new)* | `nextWriteEffects = CatalogWriteEffects(lists: [.propertyVocabulary])`, hidden `conclusionDetail(e1)` and `personsList`; `submitEdit(label: "Trade", …)` → `conclusionDetail(e1)` handle nil, `personsList` stale, `propertiesWorkspace` reloaded. Create with two bind types where the second assign throws (a `FakeStore` assign error) → `propertiesWorkspace` revalidated after the create and after the first assign, `formError` set | AC-1.1, D-7 |
| T-10 | Swift | `CitationComposerModelTests.swift` `customTermReloadsTermsAndKeepsTheDraft` *(new)* | `nextWriteEffects = CatalogWriteEffects(lists: [.propertyVocabulary], reads: [.propertyTerms: ["p1"]])`, warmed `propertyTerms(p1)`, `propertyTerms(p2)`, `sourceWorkspace(s1)`, `evidenceGraph(g1)`; `confirmCustomTerm` "Cooper" → `propertyTerms(p1)` and `propertiesWorkspace` reload once, `propertyTerms(p2)`, `sourceWorkspace(s1)` and `evidenceGraph(g1)` untouched, the draft holds the new term id | AC-1.4 |
| T-11 | Swift | `WorkspaceSessionTests.swift` `aKeyInvalidatedAndRefreshedRevalidatesOnce` (changed) | `CatalogWriteEffects(handles: ["e1"], evidenceGraphs: ["g1"], lists: [.propertyVocabulary])`, no mutation → the hidden `personClaims` key is evicted once and not reloaded | AC-1.1, D-6 |
| T-12 | Swift | `SourceTypesModelTests.swift` `loadPopulatesTypesPool`, `MetadataModelTests.swift` `loadPopulatesFields` (renamed from `…AndCounts`); `countLineCountsTheRowsShown` *(new, each file)* | 3 seeded + 2 user types → `L10n.SourceTypes.countLine(total: 5, seeded: 3, user: 2)`; 4 seeded + 1 plugin field → `L10n.Metadata.countLineWithPlugin(total: 5, seeded: 4, user: 0, plugin: 1)` | AC-4.1, AC-4.2 |
| T-13 | Swift | `CatalogCountsTests.swift` `refreshAllPopulatesEveryDestination` (changed) | Recount → `sources`, `persons`, `events`, `places`; `badge(for:)` nil for the three Configure sections | AC-4.3 |
| T-14 | Swift | `SourceTypesModelTests.swift` `assigningAppendsTheFieldAndUpdatesTheListCount`, `removingASuggestionNoSourceUsesOmitsTheValueCount`; `CatalogQueryRegistryTests.swift` `applySuggestionMutationsInvalidateTypeKey` (kept; no `userID:`) | Assign / remove → suggestions and row count as today; `typeSuggestions(t1)` and `sourceWorkspace` follow `.assignedTypeSuggestion` | AC-1.6, D-2 |
| T-15 | Go / Swift | Deleted: `TestEnsureFileThumbnail`, `TestRemoveInstallIdentity`, `TestListSubjectTypes`, `TestClearSubjectPositionBadProto`, `TestListProperties`, `TestSubjectTypePropertiesAndRegistry`, `SubjectStoreTests.listSubjectTypesReturnsSeededRows`. Changed: `TestGetPropertiesWorkspace` (two moved cases, D-11), `TestReorderSubjectTypeProperties` (bindings through `groups`), `TestSubjectPositions` "set list persist across session close", `SubjectStoreTests.setPositionPersistsAcrossSessionClose`, `subjectpositions_test.go` "get missing is tray", `connect_test.go` "missing position invalid", `subjectvocab_test.go` "connect rules" | Pass with D-11's fixtures; the moved cases assert today's seeded Properties, `participation` locked bindings, three palette types and ≥ 3 connect rules | AC-3.1, AC-3.2, D-9–D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CatalogQueryRegistryTests`, `WorkspaceSessionTests`, the three vocabulary model test files, `api/ffi/handlers` fixtures, `TestSwiftVocabularyMatchesGo`, `core/database/graphcache`.

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
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md rule 2, "Conclusion keys", "One cache owns each list" (D-13)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with one promoted Person whose Citation records a user Property "Occupation". Open the Person page, go to Properties, rename "Occupation" to "Trade", go Back: the row reads "Trade" (AC-1.1).
2. Open a Source page, go to Metadata, rename a user field the page shows, go Back: the page shows the new label (AC-1.2).
3. On Source types add "Diary": it is selected and the count line goes up by one; Configure rows show no badge (AC-1.3, AC-4.1, AC-4.3).
4. Assign a field to "Diary", then remove it: the suggestion list and the row's count follow (AC-1.6).
5. In the citation composer add a custom term to a term-valued Property: the combo offers it, the draft keeps its values, the canvas behind does not reload (AC-1.4).
6. Each view at Regular and Wide: nothing is drawn differently.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A test or view reads the vocabulary counts through a path §4 missed | STEP-6's `rg`; the build fails on a deleted member |
| A Go fixture needs an id only a deleted RPC returned | D-11's helpers; `GetPropertiesWorkspaceResponse` carries both lists |
| A composer reload resets the open term sheet | T-10 |

To compare the base: `git diff f1b3fe65 HEAD --stat -- api core macos/App` (and, once #693 has merged, `git fetch origin main && git diff f1b3fe65 origin/main --stat -- api core macos/App`), then read each §4 path that shows.

**Stop and report if:**

- S14-27's branch (or `main`, once #693 has merged) changes its list slot from one `CatalogQueryFollows` case taking a `CatalogList` used with `applyHiddenPolicyOnAny(follows:)`: the plan is refreshed first.
- S14-16's merged code differs from §4 (`missingEffects`, `effectsPending`, `GoStore.mapWriteEffects`, the Swift store's effect shape).
- `protoc`, its Go and Swift plugins, or `xcodebuild` are missing: finish STEP-1, report, and stop; Steps 2–8 need the macOS toolchain.
- A vocabulary write's resolved `Lists` lacks its family for a create, a labelled update or a delete (T-1 fails in the registry, not the handler).
- Any non-test file outside the deleted handlers calls `ListProperties`, `ListSubjectTypes`, `ClearSubjectPosition`, `subjectpositions.Clear`, `subjectvocab.PlaceableTypes` or `subjectvocab.PresentationFor`.
- A view reads `CatalogCounts` from the environment (`@Environment(CatalogCounts.self)`).
- A remaining `CatalogMutation` caller this plan does not name reaches a key only through a deleted case.
- An idea file or open spike, found at implementation time, plans a use of `RemoveInstallIdentity` or `EnsureFileThumbnail`.

## 12. Open questions

None.

## 13. PR

**Title:** S14-17 — Vocabulary writes return their effects; dead RPCs go
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-17-vocabulary-effects.md`.
