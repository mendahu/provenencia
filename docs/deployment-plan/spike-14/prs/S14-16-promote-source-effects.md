# S14-16 — Promote and Source writes return their effects

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes (batch 5; was Spike 12's S12-32)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: every view draws the same rows; only what reloads after a Promote Done or a Source write changes, and the Promote header reads the same title from the session)
**Board:** —
**Depends on:** S13-24 (landed [#639](https://github.com/mendahu/provenencia/pull/639))
**Blocks:** S14-17
**Branch:** `s14-16-promote-source-effects`
**Version bump:** patch — a refresh fix and dead-code removal, no feature ([`docs/versioning.md`](../../../versioning.md)); the next patch after `VERSION` at merge (0.2.8 → 0.2.9 at planning)
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [`app-health-audit.md`](../app-health-audit.md) A-29, A-35; working rules "Every audited write returns its effects, and a test holds them to it", "FakeStore is a seeded fake, not a second engine"; [`macos-client-patterns.md`](../../../macos-client-patterns.md) cache contract rule 2 and "Conclusion keys"; [`core/database/effects/README.md`](../../../../core/database/effects/README.md); S14-11's plan **Next adds** ([`S14-11-reload-only-what-moved.md`](S14-11-reload-only-what-moved.md)); S13-01 ship log (the `conclusionTriggers` leftover)
**Planned:** 2026-11-02 against `main` @ `25ed4bd`
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

A Promote Done, a Source rename or type change, a Source cover change and a Source delete tell the app what they changed, the way every other write under a Source already does. The app reloads exactly what moved. A Promote Done reloads the canvas, its progress, the people it filed and the sidebar counts from its effects. It no longer reloads every cached Person, Event and Place page. The sidebar recounts when a write reports that a Source or a handle was created or deleted, through a new effects list, `counts`, and the newest recount wins. A Source rename shows its new title on the Promote page header. The dead single-Subject `PromoteSubject` path goes from Go, the wire and Swift. A Go test holds every audited write's response to a `WriteEffects` field, with an allow-list of the vocabulary writes S14-17 converts.

**Why now.** Audit A-29 (audited writes drop their effects, and Swift restates them by hand with `.promotedSubject(graphId:)`) and A-35 (the `promoteSubject` path has no production caller). The S13-01 ship log measured the cost: `.promotedSubject` in `conclusionTriggers` drops every cached person's claims on any Promote. S13-24 removed Apply's pin inputs, so the handler and its Swift mapping are small enough to rewrite now. S14-11 left two shims (`CatalogMutation.readIDs(for:)`, `evidenceGraphId`) for this PR to delete.

**What stays.** What every view draws. `.createdSource(source)`, `.updatedSource(source)`, `.changedSourceType(source)` and `.deletedSource(id:)` stay as the mutations that carry a Source row or name a Source-page key, applied beside the new effects. `.updatedProperty` stays in `conclusionTriggers` until S14-17. The vocabulary writes return nothing new (S14-17). `promote.Save` and `promotetest.Save` stay: about 48 core tests seed promotions through them. `promote.BatchResult` stays as the package's tested result. The sidebar's vocabulary badges keep their `publish*` path.

**Already on `main`.** The map row's "one shared proto header message" landed in [#662](https://github.com/mendahu/provenencia/pull/662) (`3a1984b5`). `engine.proto` declares `message ConclusionSummary`, the four messages carry `ConclusionSummary summary` with their old fields reserved, and Swift maps it once in `GoStore.mapConclusionSummary`. This PR does nothing for that item. STEP-10 drops it from the map's row and scope entry.

## 2. User stories

**US-1.** As a researcher, when I press Done on Promote, the canvas, the graph's progress, the people I filed and the sidebar counts update, and Person pages I visited for other people are not reloaded.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Done that mints one new Person from graph G | `ApplyPromoteGraphAlignmentResponse.effects` has the new handle in `handles`, G in `evidence_graphs`, `evidence_graph_progress` with G in `reads`, `counts` in `lists`, and `revision > 0` |
| AC-1.2 | A Done that files one Subject onto an existing handle H | `effects.handles` has H; `lists` has no `counts` |
| AC-1.3 | The session holds `.personClaims` and `.conclusionDetail` for persons A and B, A's on screen; a Done reports `handles: [A]` | A's keys revalidate; B's keys are neither reloaded nor evicted |
| AC-1.4 | A Done whose effects carry `lists: [.counts]` | The sidebar recount hook runs once; with no `counts` it does not run |

**US-2.** As a researcher, when I rename a Source, change its type or cover, or delete it, every view that shows it updates through the write's own effects.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `UpdateSource` changes the title of a Source with graph G | `effects.source_scopes` has the Source, `evidence_graphs` has G, no `counts` |
| AC-2.2 | `SetSourceCover` to `type_icon` | `effects.source_scopes` has the Source; `evidence_graphs` is empty; no `counts` |
| AC-2.3 | `DeleteSource` on a Source with one empty graph G | `effects.evidence_graphs` has G; `reads` has `evidence_graphs` with the Source id; `lists` has `counts` |
| AC-2.4 | The Source page deletes its Source | The session applies the effects and `.deletedSource(id:)`; the recount hook runs once (from `counts`) |
| AC-2.5 | Promote is open on a graph of Source S; S is renamed to "Census 1851 (Ward 3)" in another pane | The Promote header reads "Census 1851 (Ward 3)" |

**US-3.** As a contributor, I cannot add an audited write that drops its effects.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A request message with a `user_id` field whose response has no `WriteEffects` field, not on the allow-list | `TestEveryAuditedWriteReturnsEffects` fails naming the request |
| AC-3.2 | An allow-list entry whose response now has `WriteEffects`, or whose request no longer exists | The same test fails naming the stale entry |
| AC-3.3 | `METHOD_PROMOTE_SUBJECT` (85) is called | `ffi.Call` returns the unknown-method error |

**US-4.** As a researcher, when two writes recount the sidebar close together, the badges show the newer count.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Recount 1 starts, the store gains a Source, recount 2 starts and finishes, then recount 1 finishes | `CatalogCounts.sources` is recount 2's value |

## 3. Scope

| In | Covered by |
| --- | --- |
| `effects.ListCounts` *(new)* on `sources` (creates and deletes) and `canonical_entities` (creates, deletes, and a write of `merged_into_id`) | STEP-1, AC-1.1, AC-1.2, AC-2.3 |
| `ApplyPromoteGraphAlignment`, `UpdateSource`, `SetSourceCover`, `DeleteSource` return `WriteEffects`; Apply's `revision` and `written` reserved | STEP-2, AC-1.1 – AC-2.3 |
| `PromoteSubject` RPC deleted, method 85 and its messages reserved or removed | STEP-3, AC-3.3 |
| `TestEveryAuditedWriteReturnsEffects` *(new)* with the S14-17 allow-list | STEP-4, AC-3.1, AC-3.2 |
| Swift `CatalogList.counts`; the session recounts on it; `navCountTriggers` goes; newest recount wins | STEP-5, AC-1.4, AC-4.1 |
| The four store methods return effects and their callers apply them | STEP-6, AC-1.4, AC-2.4 |
| `.promotedSubject`, `evidenceGraphId`, `readIDs(for:)` and the `mutation` parameter of `CatalogQueryFollows.ids` go; `conclusionTriggers` is `[.updatedProperty]` | STEP-7, AC-1.3 |
| Swift `promoteSubject` path deleted; tests seed through `FakeStore.seedPromotion` *(new)* | STEP-8 |
| The Promote header reads the Source title from the session | STEP-9, AC-2.5 |
| Docs: cache contract rule 2 and "Conclusion keys", `add-workspace-place`, effects README, map | STEP-10 |

| Out | Home |
| --- | --- |
| Vocabulary writes returning effects; emptying the allow-list | S14-17 |
| Removing `.updatedProperty` from `conclusionTriggers` | S14-17 |
| A descriptor slot for `lists` / `structure`; `CatalogQueryRegistry.invalidations(for list:)` | S14-27 |
| The shared proto header message | Done in #662 (see **Already on `main`**) |
| Folding `conclusion_detail.go`'s inline `ConclusionSummary` builders into `conclusionSummaryProto` | not planned (a survey find; no map row asks for it) |
| Removing `promote.BatchResult.Written` | not planned: `core/database/promote/batch_test.go` asserts on it as the package's result |
| Recounting the vocabulary badges from effects (source types, metadata fields) | not planned: they publish from their lists (`publishSourceTypes`, `publishMetadata`) |
| Stale doc line "Promote target suggestions" in "Conclusion keys" beyond what STEP-10 rewrites | STEP-10 rewrites the whole bullet; nothing left over |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Lists | `core/database/effects/registry.go` `List`, `ListVocabulary`, `ListPropertyUsage`, `Lists` | Two catalog-wide lists. No counts list |
| List gating | `registry.go` `Effect.Lists`, `Effect.ListsOnField`, `Effect.movesLists`, `Effect.validate` | A create or delete always moves `Lists`. An update moves them unless `ListsOnField` names a field the update did not write. Nothing says "never on an update" |
| Registry entries | `registry.go` `registry["sources"]`, `registry["canonical_entities"]` | No `Lists` on either. `canonical_entities` is `{Entity, Handles: self(), HandleScope: self()}` |
| Handle updates | `core/database/canonicalentities/canonicalentities.go` `sqlSetPreferred`, `sqlCountByTypeKey` | The only production update of `canonical_entities` writes `preferred_likeness_observation_id`. The count filters `merged_into_id IS NULL`; only `graphcachetest` writes `merged_into_id` today |
| List tests | `core/database/effects/lists_test.go` `TestResolveLists`; `registry_completeness_test.go` `TestPropertyUsageList`, `TestListsValidate` | Hand-built `rowchange.Change`s compared with `DeepEqual` |
| Apply handler | `api/ffi/handlers/promote_batch.go` `ApplyPromoteGraphAlignment` | Calls `writes.Run`, overwrites `res.Revision` with the recorded revision, and returns `revision` and `written`. Drops the `writes.Result` |
| Source handlers | `api/ffi/handlers/sources.go` `UpdateSource`, `SetSourceCover`, `DeleteSource` | Each calls `writes.Run` and discards the `writes.Result`. `UpdateSourceResponse{source=1}`, `SetSourceCoverResponse{source=1}`, `DeleteSourceResponse{}` |
| Precedent handlers | `sources.go` `CreateSource` (`Effects: writeEffectsProto(res)`); `promote.go` `UnpromoteSubject` | Return `WriteEffects effects` at the next free field number |
| Effects helper | `api/ffi/handlers/write_effects.go` `writeEffectsProto`, `listStrings` | Maps a `writes.Result` to the wire, `lists` from `Set.Lists` |
| Source delete | `core/database/sources/sources.go` `Delete`; `core/database/deleteimpact/register.go` `inboundEdgeList`; `facets.go` Source facets | Refused while the Source has Artifacts or Subjects. Releases notes, graphs (`evidence_graph` changes carrying `source_id`), metadata, credibility and layout, then emits the `source` delete. So a Source delete touches no handle |
| Graph rows | `registry.go` `registry["evidence_graphs"]` | A graph row names itself in `EvidenceGraph` and moves `ReadEvidenceGraphs` keyed by `source_id` |
| Single promote | `api/ffi/handlers/promote.go` `PromoteSubject`; `api/ffi/dispatch.go` `MethodPromoteSubject`, `route`; `engine.proto` `METHOD_PROMOTE_SUBJECT = 85`, `PromoteSubjectRequest`, `PromoteSubjectResponse` | No Swift production caller. Handler tests use it as a seed: `promote_test.go` (`promotableSubject`, `promoteOK`, `TestPromoteSubject`, `TestPromoteSubjectJoin`, `personBeside`), `conclusion_detail_test.go`, `conclusion_summaries_test.go` (direct call at :74), `promote_align_test.go`, `search_catalog_test.go`, `source_write_effects_test.go` `filedSubject`, `person_detail_test.go` (`mustRPC(t, PromoteSubject, …)`) |
| Seed helper | `core/database/promote/promotetest/promotetest.go` `Save` | Runs `promote.Save` in one audited write; fails the test on error |
| Reserved methods | `engine.proto` `enum Method` top block | `reserved 27; reserved "METHOD_COUNT_FILES";` and four more, each with a comment |
| Router test | `api/ffi/dispatch_test.go` `TestCallRouter` | Its "unknown method" case uses 100, which is `METHOD_UNPROMOTE_SUBJECT` and passes only because a nil request fails |
| Reserved-field test | `api/ffi/handlers/write_effects_test.go` `TestRetiredFieldsStayReserved` | Checks retired fields by descriptor |
| Swift lists | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogList` | `vocabulary`, `propertyUsage`. `api/ffi/swift_vocabulary_test.go` `listStrings` holds it to `effects.Lists` |
| Mutations | `CatalogMutation.swift` `CatalogMutation.promotedSubject(graphId:)`, `CatalogMutationKind.promotedSubject`, `invalidationKind`, `evidenceGraphId`, `readIDs(for:)` | `.promotedSubject` is built only in `PromoteModel.promote()`. `evidenceGraphId` and `readIDs` are S14-11's shims, commented "S14-16 removes this" |
| Follows | `CatalogLastUpdated.swift` `CatalogQueryFollows.ids(in:mutation:)` | `.reads` appends `mutation?.readIDs(for:)`; no other case reads `mutation` |
| Triggers | `CatalogQueryRegistry.swift` `conclusionTriggers` = `[.promotedSubject, .deletedSource, .updatedProperty]`; `navCountTriggers` = `[.createdSource, .deletedSource, .promotedSubject]` | Every Conclusion key (`personsList`, `eventsList`, `placesList`, `conclusionDetail`, `personClaims`, `personFamily`, `personEvents`) uses `conclusionTriggers`; each also follows `.touchedHandles` through `refreshes` |
| Graph key | `CatalogQueryRegistry.swift` `.evidenceGraph` | `invalidateOn: [.deletedSource, .promotedSubject]`, invalidation from `mutation.evidenceGraphId` else `.allCached(.evidenceGraph)`; `refreshes` follows `.evidenceGraphs`. Comment: "Promote reports none" |
| Graph list | `CatalogQueryRegistry.swift` `.evidenceGraphsList` | Comment: "`deleteSource` reports none, so `CatalogMutation.readIDs(for:)` names the deleted Source" |
| List switch | `CatalogQueryRegistry.swift` `invalidations(for list:project:)` | `case .vocabulary, .propertyUsage:` → the Properties workspace |
| Recount | `WorkspaceSession.swift` `onNavCountsMoved`, `apply(_:_:)` tail | Recounts when `mutation?.invalidationKind` is in `navCountTriggers` |
| Counts | `macos/App/Features/Catalog/CatalogCounts.swift` `refreshAll()` | No generation: the last recount to finish wins, even when it started first |
| Counts wiring | `macos/App/Features/Workspace/WorkspaceView.swift` | `session.onNavCountsMoved = { await counts.refreshAll() }` |
| Store protocol | `macos/App/Platform/GenealogyStore.swift` `updateSource -> CatalogSource`, `setSourceCover -> CatalogSource`, `deleteSource` (no result), `applyPromoteGraphAlignment` (no result), `promoteSubject(projectDir:userID:subjectID:entityID:) -> CatalogPromoteResult`, extension `promoteSubject(projectDir:userID:subjectID:)`, `struct CatalogPromoteResult` | — |
| GoStore | `macos/App/Platform/GoStore.swift` `updateSource`, `setSourceCover`, `deleteSource`, `promoteSubject`, `applyPromoteGraphAlignment`, `mapWriteEffects`; `CoreInvoke.swift` `CoreMethod.promoteSubject` | Apply and delete discard their responses |
| FakeStore | `macos/App/Platform/FakeStore.swift` `updateSource`, `setSourceCover`, `deleteSource`, `promoteSubject` (with refusals: `promote.invalid`, `promote.unsupported_type`, `identityclaims.already_member`, `identityclaims.type_mismatch`), `promoteSubjectDelayNanoseconds`, `applyPromoteGraphAlignment`, `workspaceNavCounts`, `nextWriteEffects` | `promoteSubject` re-implements engine refusals, against the working rule |
| Test stub | `macos/ProvenenciaTests/OnboardingModelTests.swift` (a `GenealogyStore` stub) | Implements all five methods |
| Callers | `macos/App/Features/Promote/PromoteModel.swift` `promote()`; `macos/App/Features/Sources/SourcePageModel.swift` `confirmResourceDelete` (`"source"`); `SourceIdentitySection.swift` (rename / type); `SourcePageContext.swift` `applySource(_:typeChanged:)`, `setSourceCover(mode:primaryArtifactID:)`, `refreshCoverFromStore` | Promote applies `.promotedSubject(graphId:)`; delete applies `.deletedSource(id:)`; rename, type and cover apply `.updatedSource` / `.changedSourceType` with no effects |
| Promote header | `macos/App/Features/Promote/PromoteView.swift` header; `PromoteModel.entry.sourceTitle` | Reads the title frozen into the location when Promote opened |
| Session names | `macos/App/Features/Workspace/Session/WorkspaceSession+Sources.swift` `source(id:)`; `macos/App/Features/Workspace/WorkspaceLocationLabels.swift` | The breadcrumbs read `session.source(id:)?.title`, which `.updatedSource` patches |
| Swift tests | `WorkspaceSessionTests` (`applyPromoteMarksHiddenPersonsListStale`, `applyTriggerEvictsHiddenDetailPagesAndRevalidatesTheVisibleOne`, `aKeyInvalidatedAndRefreshedRevalidatesOnce`, `applyRunsNavCountsHookForTriggers`, `applySkipsNavCountsHookForOtherWrites`, `deletedSourceDropsItsGraphs`); `CatalogQueryRegistryTests` (`navCountTriggersAreSourceAndPromoteWrites`, the matrix asserting `conclusionTriggers`, `personClaimsInvalidateOnPromote`, `personFamilyInvalidateOnPromote`, `personEventsInvalidateOnPromote`, `readIDs` through `.ids(in:mutation:)`); `SubjectStoreTests` promote tests; `CatalogCountsTests.refreshAllCountsPromotedPersons`; `SourcePageModelTests.confirmDeleteSourceRecountsThroughSession`; `SourceCopyModelTests.copyRecountsThroughSession`; `PromoteModelTests` (`promoteSendsItsGraph`, `donePromotesEveryRowWithATargetOnly`, `aStaleDoneReproposes`); `EvidenceGraphModelTests`, `SearchCatalogStoreTests`, `WorkspaceSidebarSectionsTests` (seed through `promoteSubject`) | — |

**Precedent.** `CreateSource` and `UnpromoteSubject` (`api/ffi/handlers/sources.go`, `promote.go`) are the handler shape: `Effects: writeEffectsProto(res)` at the next free field. S12-25's `de0b4411` ("Report every catalog-wide list a write changed as one declared set on the effects, the wire and Swift") is how a list is added: a `List` constant, an entry in `Lists`, `Lists:` on the registry entries, Swift `CatalogList` with its vocabulary guard. On Swift, `SourcesModel` creating a Source applies `session.apply(effects, .createdSource(created))`: copy it for the update, cover and delete callers. S14-11 (#657) is the precedent for deleting a mutation shim once effects carry the ids.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Go adds `ListCounts List = "counts"` *(new)*, appended to `Lists` (`[ListVocabulary, ListPropertyUsage, ListCounts]`). Its comment: "A Source or a Person, Event or Place handle was created, deleted or merged, so a sidebar count changed." | Working rule "Every audited write returns its effects": sidebar counts become an effects list. A catalog-wide fact is a list, as `vocabulary` is | A `counts bool` on `WriteEffects` — `lists` already carries catalog-wide facts and the vocabulary guard test |
| D-2 | `Effect` gains `ListsOnCreateDelete bool` *(new)*: when set, an update never moves `Lists`. `movesLists` returns false for an update when it is set. `validate` rejects it without `Lists`, and rejects it together with `ListsOnField` | A Source rename or cover change does not change a count, and no column of `sources` does. The flag states that directly | `ListsOnField: "id"` (an id is never updated) — a trick the next reader must decode; moving counts on every Source update — a recount query per rename |
| D-3 | `registry["sources"]` gains `Lists: []List{ListCounts}, ListsOnCreateDelete: true`. `registry["canonical_entities"]` gains `Lists: []List{ListCounts}, ListsOnField: "merged_into_id"` | The Persons, Events and Places counts filter `merged_into_id IS NULL` (`sqlCountByTypeKey`), so a merge moves them; a preferred-picture update does not. `Handles` and `HandleScope` on both entries are unchanged | `ListsOnCreateDelete` on handles too — a future merge writer would silently skip the recount |
| D-4 | `ApplyPromoteGraphAlignmentResponse` becomes `{ reserved 1, 2; reserved "revision", "written"; WriteEffects effects = 3; }`. `message ApplyPromoteGraphAlignmentWritten` is deleted. The handler keeps `writes.Run`, drops the revision overwrite and the `Written` loop, and returns `Effects: writeEffectsProto(wrote)` | Nothing in Swift reads either field (`GoStore` discards the response). The effects carry the revision (`WriteEffects.revision`) and the handles. An all-skip Done that writes nothing returns zero effects, which applies as a no-op | Keeping `written` — a second report of the same handles that nothing reads |
| D-5 | `UpdateSourceResponse` and `SetSourceCoverResponse` gain `WriteEffects effects = 2`; `DeleteSourceResponse` gains `WriteEffects effects = 1`. Each handler keeps its `writes.Run` call and passes its `writes.Result` to `writeEffectsProto`. `SetSourceCover` returns the result of the `writes.Run` inside its `set` closure the same way, on both the direct path and the `thumbnailPass.run` path | `CreateSource` precedent; next free field number | — |
| D-6 | `PromoteSubject` goes: handler (`promote.go` `PromoteSubject`), `dispatch.go` `MethodPromoteSubject` and its `route` case, `METHOD_PROMOTE_SUBJECT = 85` (becomes `reserved 85; reserved "METHOD_PROMOTE_SUBJECT";` in the reserved block, with a comment "S14-16: Promote files through ApplyPromoteGraphAlignment"), and `PromoteSubjectRequest` / `PromoteSubjectResponse` (deleted; no message references them). `canonicalEntityProto` and `identityClaimProto` stay if a remaining handler calls them, and are deleted with it if not (`rg 'canonicalEntityProto\|identityClaimProto' api` decides) | A-35. Method numbers are reserved, as for `METHOD_COUNT_FILES` | Keeping the RPC for tests — tests seed through `promotetest.Save` (D-7) |
| D-7 | Handler tests seed promotions through one helper in `api/ffi/handlers/promote_test.go`: `promoteInDir(t, dir, userID, subjectID, entityID string) promote.Result` *(new)*, which runs `withProjectCatalog(dir, …)` and `promotetest.Save` (an empty `entityID` mints). It replaces `promoteOK`; `promotableSubject` returns `(dir, userID, subjectID string)` instead of a request. `TestPromoteSubject` and `TestPromoteSubjectJoin` are deleted: `core/database/promote/promote_test.go` covers `Save` | One seed path for core and handler tests | A handler-level seed through Apply — it needs a proposal revision per seed |
| D-8 | `TestEveryAuditedWriteReturnsEffects` *(new)* in `api/ffi/write_effects_pin_test.go` *(new)*, package `ffi`. It walks `engine.File_engine_proto.Messages()`; for each message named `…Request` with a field `user_id`, it looks up the message `…Response` and requires a field whose message type is `provenencia.engine.v1.WriteEffects` (by `FullName` of `engine.WriteEffects` descriptor). Requests in `effectsPending` *(new)*, a `map[string]string` of request name → `"S14-17"`, are skipped. The test also fails when an `effectsPending` entry names no request or names one whose response already has the field. The 16 entries: `CreateSourceTypeRequest`, `UpdateSourceTypeRequest`, `DeleteSourceTypeRequest`, `CreateMetadataFieldRequest`, `UpdateMetadataFieldRequest`, `DeleteMetadataFieldRequest`, `AssignTypeFieldRequest`, `RemoveTypeFieldRequest`, `CreatePropertyRequest`, `UpdatePropertyRequest`, `DeletePropertyRequest`, `CreatePropertyTermRequest`, `UpdatePropertyTermRequest`, `DeletePropertyTermRequest`, `AssignSubjectTypePropertyRequest`, `RemoveSubjectTypePropertyRequest` | Working rule (map review K-10). Descriptors, not regex over the `.proto` text, so a nested or one-line message is seen | Checking handler code — the wire is the contract S14-17 empties |
| D-9 | Swift `CatalogList` gains `case counts = "counts"`. `CatalogQueryRegistry.invalidations(for list:project:)` returns `[]` for `.counts`. `WorkspaceSession.apply(_:_:)` recounts when `effects.lists.contains(.counts)` and `onNavCountsMoved` is set (one `trackWrite`), in place of the `navCountTriggers` check. `navCountTriggers` is deleted | The recount is not a cache key (S14-27's scope says so). Every write that changes a count now reports it: `CreateSource` and Copy Source already return effects; this PR adds Apply and `DeleteSource` | Keeping `navCountTriggers` beside the list — two mechanisms for one job |
| D-10 | `CatalogCounts.refreshAll()` takes a generation: `private var refreshGeneration = 0` *(new)*; each call increments it and captures its value, and after the store returns it assigns the counts (or `lastRefreshError`) only when the captured value still equals `refreshGeneration` | The map's leftover: the newest recount wins. Go serializes catalog ops in call order (`catalogsession`), so the call started last reads the newest catalog | Serializing recounts in Swift — slower, and a queued recount is redundant |
| D-11 | Store signatures: `updateSource(…) -> (CatalogSource, CatalogWriteEffects)`, `setSourceCover(…) -> (CatalogSource, CatalogWriteEffects)`, `deleteSource(…) -> CatalogWriteEffects`, `applyPromoteGraphAlignment(…) -> CatalogWriteEffects`. `GoStore` maps `resp.hasEffects ? Self.mapWriteEffects(resp.effects) : .empty`, as its other effect-returning methods do. `FakeStore` returns `nextWriteEffects` and keeps its current state changes. The `OnboardingModelTests` stub returns `.empty` | `createSource` / `addSourceNote` tuple shape; FakeStore working rule (scripted effects) | — |
| D-12 | Callers: `PromoteModel.promote()` applies `session.apply(effects)`. `SourcePageModel.confirmResourceDelete` (`"source"`) applies `session.apply(effects, .deletedSource(id: target.id))`. `SourcePageContext.applySource(_:effects:typeChanged:)` gains an `effects: CatalogWriteEffects` parameter and applies `session.apply(effects, typeChanged ? .changedSourceType(source) : .updatedSource(source))`; the rename path in `SourceIdentitySection` and `setSourceCover(mode:primaryArtifactID:)` pass the returned effects. `refreshCoverFromStore` (a read, no write) keeps `session.apply(.updatedSource(ws.source))` | The mutation still carries the enriched row (`patchedSource`) and names the Source-page key; the effects name the graphs, scopes and counts | Dropping `.updatedSource` — the effects cannot carry the row |
| D-13 | `CatalogMutation.promotedSubject(graphId:)`, `CatalogMutationKind.promotedSubject`, its `invalidationKind` arm, `evidenceGraphId` and `readIDs(for:)` are deleted. `CatalogQueryFollows.ids(in:mutation:)` becomes `ids(in effects: CatalogWriteEffects) -> [String]` and its `.reads` arm returns the effects' ids only; every caller drops the argument. `conclusionTriggers` becomes `[.updatedProperty]` with its doc saying "Property updates report no effects until S14-17". `.evidenceGraph` gets `invalidateOn: []` and keeps its required `invalidation:` closure as `{ _, _ in .allCached(.evidenceGraph) }`; its comment says every graph write, Promote and Source delete included, names its graphs in `evidenceGraphs` | Source delete reports its graphs and the graph-list read (AC-2.3); Promote reports its graph, progress and handles (AC-1.1). A Source delete touches no handle (it is refused while the Source has Subjects), so `.deletedSource` leaves `conclusionTriggers` with no loss | Keeping `.deletedSource` in `.evidenceGraph`'s `invalidateOn` — it would mark every cached canvas stale on a delete that names its own graphs |
| D-14 | A deleted Source's graph keys follow the effects like a `DeleteEvidenceGraph`: `.applyHiddenPolicy(follows: .evidenceGraphs)` marks a hidden one stale. No eviction is added | Same path as the existing graph delete; no navigation reaches a deleted Source's graph | Evicting deleted graphs — a new case for one write |
| D-15 | The Swift `promoteSubject` path goes: both protocol forms, `GoStore.promoteSubject`, `CoreMethod.promoteSubject`, `FakeStore.promoteSubject` and `promoteSubjectDelayNanoseconds`, the `OnboardingModelTests` stub method, `CatalogPromoteResult`, and the stale `CatalogMutation` doc line naming it. Tests that seed a promotion call `FakeStore.seedPromotion(projectDir:subjectID:entityID:) -> (entity: CatalogCanonicalEntity, claim: CatalogIdentityClaim)` *(new, not on the protocol)*: it writes `membershipBySubject` and `claimBySubject` as `promoteSubject` did, records no call, and throws nothing (a Subject it does not hold is a test bug: `preconditionFailure`). `SubjectStoreTests`' promote tests, which test the fake's refusals, are deleted | Working rule "FakeStore is a seeded fake, not a second engine"; A-35 | Keeping `promoteSubject` on the protocol for tests — a production surface with no production caller |
| D-16 | `PromoteModel` gains `var sourceTitle: String? { session.source(id: entry.sourceID)?.title.trimmedNonEmpty ?? entry.sourceTitle }` *(new)*, and the header in `PromoteView` reads `model.sourceTitle` | The breadcrumbs' rule (`LocationNames.sourceTitle`): the session's patched Source row first, the carried title as fallback. Map Check: "a Source rename refreshes Promote's Source titles" | Re-reading the Source in Promote — the session already holds it |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `effects.ListCounts` *(new)* | const | `ListVocabulary`, `ListPropertyUsage` | Extends the same vocabulary (D-1) |
| `Effect.ListsOnCreateDelete` *(new)* | field | `Effect.ListsOnField` | `ListsOnField` cannot say "no update moves it" (D-2) |
| `WriteEffects effects` on four responses | proto field | `CreateSourceResponse.effects` | Same field on the remaining writes |
| `TestEveryAuditedWriteReturnsEffects`, `effectsPending` *(new)* | test, map | `swift_vocabulary_test.go` (descriptor-free regex), `TestRetiredFieldsStayReserved` (descriptor checks) | No test walks every request today; searched `File_engine_proto` and `Messages()` in `api/` |
| `promoteInDir` *(new)* | test helper | `promoteOK` (replaced), `promotetest.Save` (wrapped) | Handler tests hold a dir, not a `*database.Catalog` |
| `CatalogList.counts` *(new)* | enum case | `CatalogList.vocabulary` | Mirrors D-1 |
| `CatalogCounts.refreshGeneration` *(new)* | property | `WorkspaceSession.generations` (per key) | Counts are not a cache key |
| `FakeStore.seedPromotion` *(new)* | test seed | `FakeStore.promoteSubject` (deleted) | Same state writes, no refusals, off the protocol |
| `FakeStore.workspaceNavCountsDelays` *(new)*: `[UInt64]`, popped per `workspaceNavCounts` call | test hook | `createCitationDelayNanoseconds` | One delay per call is needed to finish two recounts out of order (T-12) |
| `PromoteModel.sourceTitle` *(new)* | computed property | `LocationNames.sourceTitle` | Promote has the session, not `LocationNames` |

**Builds on.** The effects registry and its list vocabulary (`de0b4411`), `writeEffectsProto`, descriptor `refreshes` (S13-01, S14-11), FakeStore's `nextWriteEffects` scripting.
**Deletes.** Go: `PromoteSubject` handler and route, `MethodPromoteSubject`, `PromoteSubjectRequest` / `Response`, `ApplyPromoteGraphAlignmentWritten`, Apply's revision overwrite and `Written` loop, `promoteOK`, `TestPromoteSubject`, `TestPromoteSubjectJoin`; `canonicalEntityProto` / `identityClaimProto` if no caller remains (D-6). Swift: `navCountTriggers`, `CatalogMutation.promotedSubject`, `CatalogMutationKind.promotedSubject`, `evidenceGraphId`, `readIDs(for:)`, the `mutation` parameter of `CatalogQueryFollows.ids`, `.deletedSource` and `.promotedSubject` in `conclusionTriggers` and in `.evidenceGraph`'s `invalidateOn`, both `promoteSubject` protocol forms and their implementations, `CoreMethod.promoteSubject`, `promoteSubjectDelayNanoseconds`, `CatalogPromoteResult`, `SubjectStoreTests`' promote tests, `CatalogQueryRegistryTests.navCountTriggersAreSourceAndPromoteWrites`. Docs: the "report none" sentence of rule 2, the `navCountTriggers` line of `add-workspace-place`, the `conclusionTriggers` bullet's claim that every Conclusion key busts together, every "S14-16 removes this" comment.
**Next adds.** S14-17 converts each vocabulary write and deletes its `effectsPending` entry; when the map is empty it deletes `effectsPending` and `.updatedProperty` leaves `conclusionTriggers` (then the set is empty and goes). S14-27 moves `.vocabulary` / `.propertyUsage` onto descriptors; `.counts` stays out of that switch as a recount. A new count on the sidebar is `Lists: []List{ListCounts}` on its table's registry entry.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Source delete's effects name deleted graphs | `.evidenceGraph` refresh (`applyHiddenPolicy`), `.evidenceGraphsList` `.merged` read | Same as `DeleteEvidenceGraph` today: a hidden canvas goes stale; the graph list re-reads the deleted Source's graphs (empty) and merges, as the `readIDs` shim did |
| A handle-deleting or merging write (Unpromote, Subject delete) that reports `counts` now recounts | `CatalogCounts.refreshAll` | It already ran on Promote; the count read is the same query |
| The Promote header reads the session's Source row | `PromoteModel.sourceTitle` | Falls back to the carried title when the Sources list is not warm or the Source is gone |

### Scale

At the map's Scale (3,000 Sources; 37,000 Subjects; 200,000 audit rows; 60 graphs on one census):

- **Promote Done.** Today `.promotedSubject` invalidates every Conclusion key (`allCached(.conclusionDetail)` evicts every hidden detail page, every warmed `personClaims` / `personFamily` / `personEvents`, and the three lists) and the canvas key. After: only keys whose handles are in `touchedHandles`, the one graph and its progress read, the three lists (which follow any touched handle: unchanged), and a recount only when a handle was minted. The recount is `GetWorkspaceNavCounts` (six `COUNT` queries) that already ran on every Done.
- **Source rename.** Moves the Source's graphs (60 on the census) through `EvidenceGraph` and their search documents, as the Go side already does on every rename; Swift now marks those 60 graph keys stale when hidden instead of leaving them showing the old title. No Conclusion key reloads.
- **Source delete.** A Source with Subjects or Artifacts is refused, so a delete moves at most its empty graphs. It no longer reloads every Conclusion key or every cached canvas.
- **Pin test.** One walk of about 300 messages, once per `go test`.

### Edges

- **Assumptions.** `movesLists` assumes a list is moved by every create and delete of its table; D-2 keeps that and only narrows updates. `Resolve` dedupes lists in `Lists` order (T-1 checks `counts` lands last). The session's recount assumes one hook call per apply; D-9 calls it at most once however many changes reported `counts`.
- **Neighbors.** A `canonical_entities` update of `preferred_likeness_observation_id` → no `counts` (T-1). A write of `merged_into_id` → `counts` (T-1). A `sources` update of `title`, `source_type_id` or cover columns → no `counts` (T-1, T-3). A Done that only joins → no `counts` (T-4). A Copy Source (already returns effects) → `counts` from its `sources` create, so `copyRecountsThroughSession` keeps passing with scripted effects (T-11).
- **Removed inputs.** `.deletedSource` leaving `conclusionTriggers`: a Source delete is refused while it has Subjects (`deleteimpact` inbound `subjects.source_id`), so it touches no handle and no Conclusion key loses a reload. `.deletedSource` leaving `.evidenceGraph` and `readIDs`: the delete's effects name the same graphs and the same Source read (T-3, T-9).
- **Real data.** None — no table, seed or configuration drives a rule here.
- **Untrusted size.** None — no list read from data the app did not write is added.

### Schema

None — no catalog change. `canonical_entities.merged_into_id` already exists.

### Go core

| Package | Change |
| --- | --- |
| `core/database/effects` | `ListCounts` *(new)*, `Lists` gains it (D-1); `Effect.ListsOnCreateDelete` *(new)*, `movesLists` and `validate` read it (D-2); `registry["sources"]` and `registry["canonical_entities"]` gain `Lists` (D-3). `README.md` names the `counts` list and the new flag |

Errors: none new.

### FFI

`api/proto/engine.proto`:

```proto
message UpdateSourceResponse { Source source = 1; WriteEffects effects = 2; }
message SetSourceCoverResponse { Source source = 1; WriteEffects effects = 2; }
message DeleteSourceResponse { WriteEffects effects = 1; }
message ApplyPromoteGraphAlignmentResponse {
  reserved 1, 2;
  reserved "revision", "written";
  WriteEffects effects = 3;
}
// enum Method: reserved 85; reserved "METHOD_PROMOTE_SUBJECT";
```

`ApplyPromoteGraphAlignmentWritten`, `PromoteSubjectRequest` and `PromoteSubjectResponse` are deleted. Handlers: `api/ffi/handlers/promote_batch.go`, `sources.go`, `promote.go`; `api/ffi/dispatch.go`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Recount | — | `WorkspaceSession.apply(_:_:)`, `CatalogCounts.refreshAll()` | D-9, D-10 |
| Effects callers | — | `PromoteModel`, `SourcePageModel`, `SourcePageContext`, `SourceIdentitySection` | D-12 |
| Promote header title | existing `Text(verbatim:)` | `PromoteView` | D-16; same font and place |

States: unchanged. Accessibility: unchanged.

### Strings

None — no copy changes. No action changes what it covers.

## 7. Steps

**STEP-1 — A `counts` effects list.**
Files: `core/database/effects/registry.go`, `lists_test.go`, `registry_completeness_test.go`, `README.md`. Skill: `add-audit-scope` (read only: scopes are unchanged).
Do: D-1, D-2, D-3. Add T-1 and T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/...` passes. `go test -tags fts5 ./api/ffi/ -run TestSwiftVocabulary` fails on the missing Swift `counts` case (fixed in STEP-5; note it and continue — STEP-5 is the step that makes it pass).

**STEP-2 — Promote and Source handlers return their effects.**
Files: `api/proto/engine.proto`, regenerated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/promote_batch.go`, `sources.go`, `promote_batch_test.go`, `source_write_effects_test.go`, `write_effects_test.go`. Skill: `add-ffi-handler`.
Do: D-4, D-5. Rewrite `promote_batch_test.go`'s assertions on `Written` / `Revision` to assert on `GetEffects()` (T-4). Add T-3 cases. Extend `TestRetiredFieldsStayReserved` for Apply's fields 1 and 2 (T-6). The Swift build still compiles: the Swift store does not read the new fields yet.
Check: `./scripts/generate-proto.sh` runs clean; `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/...` passes.

**STEP-3 — The single-Subject `PromoteSubject` RPC goes.**
Files: `api/proto/engine.proto` (+ regenerate), `api/ffi/handlers/promote.go`, `promote_test.go`, `conclusion_detail_test.go`, `conclusion_summaries_test.go`, `promote_align_test.go`, `search_catalog_test.go`, `source_write_effects_test.go`, `person_detail_test.go`, `api/ffi/dispatch.go`, `dispatch_test.go`. Skill: `add-ffi-handler`.
Do: D-6, D-7. In `TestCallRouter`, change the "unknown method" case to an unused number above the enum's highest value, and add `{name: "reserved promote subject", method: 85}` (T-7). Leave the Swift `promoteSubject` in place until STEP-8, but it now calls a reserved method: STEP-8 removes it before any Swift test runs it.
Check: `CGO_ENABLED=1 go test -tags fts5 ./...` passes; `rg 'PromoteSubject\b|PromoteSubjectRe(quest|sponse)|MethodPromoteSubject' api core --type go` finds nothing outside generated code's reserved comment.

**STEP-4 — Every audited write is pinned to its effects.**
Files: `api/ffi/write_effects_pin_test.go` *(new)*.
Do: D-8 (T-5).
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestEveryAuditedWriteReturnsEffects` passes; removing `effects = 1` from `DeleteSourceResponse` locally makes it fail naming `DeleteSourceRequest` (revert).

**STEP-5 — The sidebar recounts from `counts`; the newest recount wins.**
Files: `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift`, `macos/App/Features/Catalog/CatalogCounts.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `CatalogCountsTests.swift`, `CatalogQueryRegistryTests.swift`. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-9, D-10. Add `FakeStore.workspaceNavCountsDelays`. Replace `applyRunsNavCountsHookForTriggers` and `applySkipsNavCountsHookForOtherWrites` with T-10; delete `navCountTriggersAreSourceAndPromoteWrites`; add T-12, T-14's `.counts` case. Update the `onNavCountsMoved` and `refreshAll` doc comments to name `WriteEffects.lists` `counts`. `.createdSource` keeps its other uses.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabulary` passes; `xcodebuild test` passes for `WorkspaceSessionTests`, `CatalogCountsTests`, `CatalogQueryRegistryTests`.

**STEP-6 — The four store methods return effects and their callers apply them.**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/App/Features/Sources/SourcePageModel.swift`, `SourcePageContext.swift`, `SourceIdentitySection.swift`, `SourceArtifactsSection.swift` (only if it calls `applySource` or the cover method directly), and the tests that call these methods. Skill: `add-swift-test`.
Do: D-11, D-12. `PromoteModel.promote()` calls `session.apply(effects)` in this step; the `.promotedSubject` case is left unused until STEP-7 deletes it. Update `SourcePageModelTests.confirmDeleteSourceRecountsThroughSession` and `SourceCopyModelTests.copyRecountsThroughSession` to script `nextWriteEffects = CatalogWriteEffects(lists: [.counts])` (T-11). Add T-8, T-13.
Check: `xcodebuild test` passes for `PromoteModelTests`, `SourcePageModelTests`, `SourceCopyModelTests`.

**STEP-7 — The Promote and Source-delete shims go.**
Files: `CatalogMutation.swift`, `CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `CatalogQueryRegistryTests.swift`, `EvidenceGraphModelTests.swift`.
Do: D-13, D-14. Tests that applied `.promotedSubject(graphId:)` apply `CatalogWriteEffects` naming the same graph (`evidenceGraphs`, `reads: [.init(read: .evidenceGraphProgress, ids: [g])]`) and handles. `personClaimsInvalidateOnPromote`, `personFamilyInvalidateOnPromote` and `personEventsInvalidateOnPromote` are renamed `personClaimsRevalidateOnTouchedHandle`, `personFamilyRevalidateOnTouchedHandle` and `personEventsRevalidateOnTouchedHandle` and apply `CatalogWriteEffects(handles: [h])` in place of the mutation. The matrix asserts `conclusionTriggers == [.updatedProperty]`. `deletedSourceDropsItsGraphs` scripts the delete's effects (`reads: [.init(read: .evidenceGraphs, ids: [sourceID])]`, `evidenceGraphs`) instead of relying on `readIDs` (T-9). Add T-15. Update the comments at `.evidenceGraph` and `.evidenceGraphsList` and in `CatalogQueryFollows`.
Check: `rg 'promotedSubject|evidenceGraphId|readIDs\(|navCountTriggers' macos/App macos/ProvenenciaTests` finds nothing; `xcodebuild test` passes.

**STEP-8 — The Swift `promoteSubject` path goes.**
Files: `GenealogyStore.swift`, `GoStore.swift`, `CoreInvoke.swift`, `FakeStore.swift`, `OnboardingModelTests.swift`, `SubjectStoreTests.swift`, `CatalogCountsTests.swift`, `CatalogQueryRegistryTests.swift`, `WorkspaceSessionTests.swift`, `EvidenceGraphModelTests.swift`, `SearchCatalogStoreTests.swift`, `WorkspaceSidebarSectionsTests.swift`.
Do: D-15. Each seeding call becomes `store.seedPromotion(projectDir:subjectID:entityID:)` (no `try await`). `EvidenceGraphModelTests`' `#expect(!store.recordedCalls.contains { $0.hasPrefix("promoteSubject") })` checks are deleted with the method.
Check: `rg 'promoteSubject\b|CatalogPromoteResult|promoteSubjectDelayNanoseconds' macos` finds nothing; `xcodebuild test` passes.

**STEP-9 — The Promote header follows a Source rename.**
Files: `PromoteModel.swift`, `PromoteView.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`.
Do: D-16 (T-16).
Check: `xcodebuild test` passes for `PromoteModelTests`.

**STEP-10 — Docs.**
Files: `docs/macos-client-patterns.md` (rule 2's last sentence: Source row writes and promote now report effects, and apply them beside `.createdSource` / `.updatedSource` / `.changedSourceType` / `.deletedSource`; the sidebar recounts on `WriteEffects.lists` `counts`; the "Conclusion keys" first bullet: each Conclusion key follows `touchedHandles` through its descriptor, and `conclusionTriggers` holds only `.updatedProperty` until S14-17), `.agents/skills/add-workspace-place/SKILL.md` (the `navCountTriggers` line → "A write that changes a sidebar count reports `counts` in `WriteEffects.lists` (Go `effects.ListCounts` on its table's registry entry); `WorkspaceSession.apply` recounts. Features never call `CatalogCounts.refreshAll()`."), `core/database/effects/README.md` (done in STEP-1; check), `docs/deployment-plan/spike-14/deployment-plan.md` (S14-16 row and scope entry: drop the shared-header item, "landed in #662"; status), `completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-16-promote-source-effects.md` passes; `scripts/check-guidance-names.py` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/effects/lists_test.go` `TestResolveLists` | New rows: `source` create → `[counts]`; `source` update of `title` → no `counts`; `source` delete → `[counts]`; `canonical_entity` create → `[counts]`; `canonical_entity` update of `preferred_likeness_observation_id` → none; update of `merged_into_id` → `[counts]`; a batch with a `source_type` create and a `source` create → `[vocabulary, counts]` (order) | D-1 – D-3 |
| T-2 | Go | `core/database/effects/registry_completeness_test.go` `TestListsValidate` | `Effect{ListsOnCreateDelete: true}` → error; `Effect{Lists: [counts], ListsOnCreateDelete: true, ListsOnField: "x"}` → error; `Effect{Lists: [counts], ListsOnCreateDelete: true}` → nil | D-2 |
| T-3 | Go | `api/ffi/handlers/source_write_effects_test.go` `TestSourceWritesReturnEffects` | New cases: rename a filed Source → `source_scopes` has it, `evidence_graphs` has its graph, no `counts`; `SetSourceCover` `type_icon` → `source_scopes` has it, no graph, no `counts`; a new `TestDeleteSourceReturnsEffects` *(new)* in the same file: create a Source (it has one empty graph), delete it → `evidence_graphs` has the graph, `reads` has `evidence_graphs` with the Source id, `lists` has `counts` | AC-2.1, AC-2.2, AC-2.3 |
| T-4 | Go | `api/ffi/handlers/promote_batch_test.go` `TestApplyPromoteGraphAlignment` | "mints a person in one revision" → `effects.revision > proposal.revision`, one `handles` id whose `GetConclusionDetail` ref starts `PER-`, graph in `evidence_graphs`, progress read for the graph, `counts` in `lists`; the join case → `handles` has the existing handle, no `counts` | AC-1.1, AC-1.2, D-4 |
| T-5 | Go | `api/ffi/write_effects_pin_test.go` `TestEveryAuditedWriteReturnsEffects` *(new)* | The real descriptor set → passes; subtests on a fixture file descriptor built in the test (`protodesc.NewFile` from a `descriptorpb.FileDescriptorProto` with `XRequest{user_id}` / `XResponse{}`) → fails naming `XRequest`; a stale `effectsPending` entry → fails naming it | AC-3.1, AC-3.2, D-8 |
| T-6 | Go | `api/ffi/handlers/write_effects_test.go` `TestRetiredFieldsStayReserved` | `ApplyPromoteGraphAlignmentResponse` reserves 1 and 2 and the names `revision`, `written`; the `Method` enum reserves 85 and `METHOD_PROMOTE_SUBJECT` | D-4, D-6 |
| T-7 | Go | `api/ffi/dispatch_test.go` `TestCallRouter` | method 85 → unknown-method error; "unknown method" uses an unused number | AC-3.3 |
| T-8 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `doneAppliesItsEffects` *(new)* | `nextWriteEffects = (evidenceGraphs: [g], handles: [h], lists: [.counts])`; Done → the canvas key for g revalidates, the recount hook ran once | AC-1.4, D-12 |
| T-9 | Swift | `WorkspaceSessionTests.deletedSourceDropsItsGraphs` | Delete effects with the `evidence_graphs` read for S → the graph list drops S's graphs | AC-2.4, D-13 |
| T-10 | Swift | `WorkspaceSessionTests.countsListRecounts` *(new, replaces the two nav-count tests)* | `apply(CatalogWriteEffects(lists: [.counts]))` → hook ran once; `apply(.empty, .createdSource(s))` → hook did not run; effects with `[.vocabulary]` → did not run | AC-1.4, D-9 |
| T-11 | Swift | `SourcePageModelTests.confirmDeleteSourceRecountsThroughSession`, `SourceCopyModelTests.copyRecountsThroughSession` | Scripted `[.counts]` → recount; `sources` badge as before | AC-2.4 |
| T-12 | Swift | `CatalogCountsTests.newestRecountWins` *(new)* | `workspaceNavCountsDelays = [200ms, 0]`; start recount 1, add a Source to the fake, start and await recount 2, await recount 1 → `sources` is recount 2's value | AC-4.1, D-10 |
| T-13 | Swift | `SourcePageModelTests.renameAppliesItsEffects` *(new)* | `nextWriteEffects = (evidenceGraphs: [g])`, a warmed visible canvas for g; rename → canvas key revalidated; the Sources list row shows the new title | AC-2.1, D-12 |
| T-14 | Swift | `CatalogQueryRegistryTests` | `invalidations(for: .counts, project:) == []`; matrix asserts `conclusionTriggers == [.updatedProperty]`; `.evidenceGraph`'s `invalidateOn` is empty | D-9, D-13 |
| T-15 | Swift | `WorkspaceSessionTests.promoteRevalidatesOnlyItsPeople` *(new)* | Warm `.personClaims` and `.conclusionDetail` for A (visible) and B (hidden); apply `CatalogWriteEffects(handles: [A])` → A's keys reload once; B's handles keep their values and are not stale | AC-1.3 |
| T-16 | Swift | `PromoteModelTests.headerFollowsSourceRename` *(new)* | Warm the Sources list with S "Census 1851"; open Promote with carried title "Census 1851"; apply `.updatedSource(S renamed "Census 1851 (Ward 3)")` → `model.sourceTitle == "Census 1851 (Ward 3)"`; with the list not warm → the carried title | AC-2.5, D-16 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `api/ffi/swift_vocabulary_test.go` (the `CatalogList` guard), `core/database/effects` `TestResolveLists` order, `WorkspaceSessionTests` and `CatalogQueryRegistryTests` (trigger sets), every handler test that seeded through `PromoteSubject`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `rg` with no hit across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/App`, `api`, `core` (generated reserved comments excepted) for: `navCountTriggers`, `promotedSubject`, `evidenceGraphId`, `readIDs(for`, `CatalogPromoteResult`, `promoteSubject(`, `PromoteSubjectRequest`, `MethodPromoteSubject`, `ApplyPromoteGraphAlignmentWritten`, "S14-16 removes"; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI unchanged except the Promote header's title source (D-16)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated: no brief)
- [ ] Domain doc updated: macos-client-patterns.md rule 2 and "Conclusion keys"; add-workspace-place; core/database/effects/README.md
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a census Source that has at least two graphs and promoted people. Open Alice's Person page, then Bob's, then go to graph G. Open Promote on G, promote one new Person, press Done. See: the canvas shows the filed badge, G's progress updates, the Persons sidebar count goes up by one. Back to Bob's page: it shows without a reload spinner; Alice's (if she was filed) reloads.
2. On the same Source, rename it to "Census 1851 (Ward 3)". See: the canvas header breadcrumb and, after opening Promote on G, the Promote header read the new title.
3. Make a new Source with no Artifacts. Delete it from its Source page. See: the Sources count goes down by one; the Evidence graphs list no longer shows its graph.
4. Change a Source's cover to the type icon. See: the Source row's cover changes; nothing else reloads.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Conclusion key relied on `.promotedSubject` for something `touchedHandles` does not cover (a Promote target list, a family read keyed by another handle) | T-15 and the existing `personFamily` / `personEvents` refresh tests are rewritten against effects; the manual check walks two pages |
| `SetSourceCover`'s thumbnail path returns before `writes.Run` | D-5 names both paths; T-3 covers `type_icon` and an `artifact` cover |
| Swift build breaks between STEP-3 and STEP-8 (the Swift `promoteSubject` calls a reserved method) | No Swift test calls `GoStore.promoteSubject`; FakeStore's version runs until STEP-8 |

**Stop and report if:**

- A `sources` update writes any column whose change moves a sidebar count (D-2's premise; check `sources.Update` and `sources.SetCover` change fields).
- Production code writes `canonical_entities.subject_type_id` (D-3's premise; `rg 'subject_type_id' core/database/canonicalentities`).
- `sources.Delete` succeeds on a Source with Subjects (D-13's premise that a delete touches no handle).
- A Conclusion key's test cannot pass from effects alone without adding a mutation to `conclusionTriggers`.
- `TestEveryAuditedWriteReturnsEffects` finds a request with `user_id` that is not in §5 D-8's list and not in this PR's four writes.
- `.evidenceGraph`'s `invalidation:` closure cannot stay with `invalidateOn: []` (the descriptor initializer refuses it).

## 12. Open questions

None.

## 13. PR

**Title:** S14-16 — Promote and Source writes return their effects
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-16-promote-source-effects.md`.
