# S14-16 — Promote and Source writes return their effects

**Status:** Landed in [#686](https://github.com/mendahu/provenencia/pull/686) (2026-11-02)
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
**Landed:** [#686](https://github.com/mendahu/provenencia/pull/686), 2026-11-02. Differences from the plan: (1) T-4's association case (`TestApplyPromoteGraphAlignmentReportsABridgeMint`) asserts a Done that files both ends of a participation and so mints the association: three handles and `counts`. Seeding the ends first through `promotetest.Save` already files the bridge (`AppendFiling`), so a Done that mints only the association cannot be built from a handler test; T-1 holds the association's `canonical_entity` create to `counts`. (2) T-3's cover case sets an `artifact` cover and reverts to `type_icon` on one Source, because a new Source already has the type icon and an unchanged cover writes nothing. (3) The dead `personBeside` helper, the unread `SourcesModel.catalogCounts` and the unreferenced proto `IdentityClaim` message were removed in review (F-1, F-3). (4) T-12 holds the first recount through `FakeStore.workspaceNavCountsHold` (`FakeHold`) instead of `workspaceNavCountsDelays`, and the Promote header is its own view reading the Source row once, with its counts built by the Promote page (F-2, F-6). (5) T-8 asserts the visible canvas reloads once and the recount runs once; T-13 does the same through a rename. (6) The Swift `OnboardingModelTests` stub returns `.empty` for the two `CatalogWriteEffects` methods and keeps throwing for the two that return a Source.
**Planned:** 2026-11-02 against `main` @ `25ed4bd`
**Review:** [`S14-16-promote-source-effects.review.md`](S14-16-promote-source-effects.review.md) — round 1: Fits with changes / Not ready → Fixed; round 2: Fits with changes / Ready with changes → Fixed

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

A Promote Done, a Source rename or type change, a Source cover change and a Source delete tell the app what they changed, the way every other write under a Source already does. The app reloads exactly what moved. A Promote Done reloads the canvas, its progress, the people it filed and, when it minted a handle, the sidebar counts, from its effects. It no longer reloads every cached Person, Event and Place page. The sidebar recounts when a write reports that a Source or a handle was created, deleted or merged, through a new effects list, `counts`, and the newest recount wins. A Source rename or type change shows on the Promote page header, which reads its Source row from the session. The dead single-Subject `PromoteSubject` path goes from Go, the wire and Swift. A Go test holds every audited write's response to a `WriteEffects` field, with an allow-list of the vocabulary writes S14-17 converts.

**Why now.** Audit A-29 (audited writes drop their effects, and Swift restates them by hand with `.promotedSubject(graphId:)`) and A-35 (the `promoteSubject` path has no production caller). The S13-01 ship log measured the cost: `.promotedSubject` in `conclusionTriggers` drops every cached person's claims on any Promote. S13-24 removed Apply's pin inputs, so the handler and its Swift mapping are small enough to rewrite now. S14-11 left two shims (`CatalogMutation.readIDs(for:)`, `evidenceGraphId`) for this PR to delete.

**What stays.** What every view draws. `.createdSource(source)`, `.updatedSource(source)`, `.changedSourceType(source)` and `.deletedSource(id:)` stay as the mutations that carry a Source row or name a Source-page key, applied beside the new effects. `.updatedProperty` stays in `conclusionTriggers` until S14-17. The vocabulary writes return nothing new (S14-17). `promote.Save` and `promotetest.Save` stay: about 48 core tests seed promotions through them, and after this PR `Save` and Apply's rows file through one function (D-17). `promote.BatchResult.Written` stays as the package's tested result. The sidebar's vocabulary badges keep their `publish*` path.

**Already on `main`.** The map row's "one shared proto header message" landed in [#662](https://github.com/mendahu/provenencia/pull/662) (`3a1984b5`). `engine.proto` declares `message ConclusionSummary`, the four messages carry `ConclusionSummary summary` with their old fields reserved, and Swift maps it once in `GoStore.mapConclusionSummary`. This PR does nothing for that item; the map's row and scope entry already say so.

## 2. User stories

**US-1.** As a researcher, when I press Done on Promote, the canvas, the graph's progress, the people I filed and the sidebar counts update, and Person pages I visited for other people are not reloaded.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Done that mints one new Person from graph G | `ApplyPromoteGraphAlignmentResponse.effects` has the new handle in `handles`, G in `evidence_graphs`, `evidence_graph_progress` with G in `reads`, `counts` in `lists`, and `revision > 0` |
| AC-1.2 | A Done that files one Subject onto an existing handle H and completes no connection | `effects.handles` has H; `lists` has no `counts`. A Done whose connection mints an association handle (`identityclaims.FileEvidenceGraphBridgesTx`) reports `counts` (D-1) |
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
| AC-2.6 | Promote is open on S (a Census, census mark); S's type changes to Register | The header's mark is the Register type's icon; the ref is S's ref |

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
| Swift `promoteSubject` path deleted; tests seed through `FakeStore.seedPromotion` *(new)*, which the fake's Apply also writes through | STEP-3 |
| `PromoteSubject` RPC deleted, method 85 and its messages reserved or removed; `promote.Save` and Apply's rows file through one function | STEP-4, AC-3.3 |
| `TestEveryAuditedWriteReturnsEffects` *(new)* with the S14-17 allow-list | STEP-5, AC-3.1, AC-3.2 |
| Swift `CatalogList.counts`; the session recounts on it; `navCountTriggers` goes; newest recount wins; the Sources badge's list-load publish goes | STEP-6, AC-1.4, AC-4.1 |
| The four store methods return effects and their callers apply them | STEP-7, AC-1.4, AC-2.4 |
| `.promotedSubject`, `evidenceGraphId`, `readIDs(for:)` and the `mutation` parameters of `CatalogQueryFollows.ids` and `WorkspaceSession.revalidatedKeys` go; `conclusionTriggers` is `[.updatedProperty]` | STEP-8, AC-1.3 |
| The Promote header reads its Source row from the session; one helper names a location's Source for the header and the breadcrumbs | STEP-9, AC-2.5, AC-2.6 |
| Docs: cache contract rule 2 and "Conclusion keys", `add-workspace-place`, `add-ffi-handler`, effects README, map | STEP-10 |

| Out | Home |
| --- | --- |
| Vocabulary writes returning effects; emptying the allow-list | S14-17 |
| Removing `.updatedProperty` from `conclusionTriggers` | S14-17 |
| A descriptor slot for `lists` / `structure`; `CatalogQueryRegistry.invalidations(for list:)` | S14-27 |
| The shared proto header message | Done in #662 (see **Already on `main`**) |
| Folding `conclusion_detail.go`'s inline `ConclusionSummary` builders into `conclusionSummaryProto` | not planned (a survey find; no map row asks for it) |
| Removing `promote.BatchResult.Written` | not planned: `core/database/promote/batch_test.go` asserts on it as the package's result (`BatchResult.Revision` goes, D-4) |
| The vocabulary badges' list-load writers (`publishSourceTypes`, `publishMetadata`, `publishProperties`) | S14-17, whose vocabulary writes then report `counts` (map scope entry, review M-13) |
| An optional `invalidation:` closure on `CatalogQueryDescriptor.make` when `invalidateOn` is empty | S14-27, which reshapes descriptors (review M-11) |
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
| Effects docs | `core/database/effects/README.md` (the `Set.Lists` paragraph); `registry.go` `List` doc | The README lists an `evidence_graphs` list that `Lists` no longer has (S14-11 moved it to `reads`); the `List` doc says "one the app caches whole" |
| Skill docs | `.agents/skills/add-workspace-place/SKILL.md` § 6 ("Identity/cover/title: `session.apply(.updatedSource(source))`"; "the writes that still report none (`updateSource`, `setSourceCover`, `deleteSource`, promote, Property writes) keep their mutation kinds") and the `navCountTriggers` line; `.agents/skills/add-ffi-handler/SKILL.md` § Catalog RPCs (`Effects: writeEffectsProto(wrote)`) | — |
| Source delete | `core/database/sources/sources.go` `Delete`; `core/database/deleteimpact/register.go` `inboundEdgeList`; `facets.go` Source facets | Refused while the Source has Artifacts or Subjects. Releases notes, graphs (`evidence_graph` changes carrying `source_id`), metadata, credibility and layout, then emits the `source` delete. So a Source delete touches no handle |
| Graph rows | `registry.go` `registry["evidence_graphs"]` | A graph row names itself in `EvidenceGraph` and moves `ReadEvidenceGraphs` keyed by `source_id` |
| Single promote | `api/ffi/handlers/promote.go` `PromoteSubject`; `api/ffi/dispatch.go` `MethodPromoteSubject`, `route`; `engine.proto` `METHOD_PROMOTE_SUBJECT = 85`, `PromoteSubjectRequest`, `PromoteSubjectResponse` | No Swift production caller. Handler tests use it as a seed: `promote_test.go` (`promotableSubject`, `promoteOK`, `TestPromoteSubject`, `TestPromoteSubjectJoin`, `personBeside`), `conclusion_detail_test.go`, `conclusion_summaries_test.go` (direct call at :74), `promote_align_test.go`, `search_catalog_test.go`, `source_write_effects_test.go` `filedSubject`, `person_detail_test.go` (`mustRPC(t, PromoteSubject, …)`) |
| Seed helper | `core/database/promote/promotetest/promotetest.go` `Save` | Runs `promote.Save` in one audited write; fails the test on error |
| Filing rule | `core/database/promote/promote.go` `Save`; `batch.go` `applyRow`, `BatchResult`, `SaveBatch` | `Save` and `applyRow` repeat the same steps: primary-type check (`ErrUnsupportedType`), mint through `canonicalentities.InsertTx` or join through `joinTarget`, claim through `identityclaims.InsertTx`. They differ in the already-member case (`applyRow` treats a `.handle` row already on that handle as a no-op) and in what follows (`Save` calls `identityclaims.AppendFiling`; `SaveBatch` calls `FileEvidenceGraphBridgesTx` once). `BatchResult.Revision` exists only for the handler's overwrite. The package doc and `Save`'s doc call it the promote_subject write |
| Bridge mints | `core/database/identityclaims/bridgefile.go` (association mint through `canonicalentities.InsertTx`) | A Done can mint association handles, which are `canonical_entities` rows of a non-primary type |
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
| Sources badge publish | `CatalogCounts.publishSources(_:)`; `macos/App/Features/Sources/SourcesModel.swift` `syncCatalogCounts()`; `SourcesListView.swift` `.onSettled(sourcesHandle) { model.syncCatalogCounts() }`; tests `SourcesModelTests` (two `syncCatalogCounts()` calls), `CatalogCountsTests` (`publishSources(7)`) | A second writer of `CatalogCounts.sources` beside `refreshAll`, outside any ordering. `refreshAll` runs on workspace open and on every recount |
| Recount | `WorkspaceSession.swift` `onNavCountsMoved`, `apply(_:_:)` tail, `revalidatedKeys(for:_:)` | Recounts when `mutation?.invalidationKind` is in `navCountTriggers`. `revalidatedKeys` takes the mutation only to pass it to `follows.ids(in:mutation:)` |
| Counts | `macos/App/Features/Catalog/CatalogCounts.swift` `refreshAll()` | No generation: the last recount to finish wins, even when it started first |
| Counts wiring | `macos/App/Features/Workspace/WorkspaceView.swift` | `session.onNavCountsMoved = { await counts.refreshAll() }` |
| Store protocol | `macos/App/Platform/GenealogyStore.swift` `updateSource -> CatalogSource`, `setSourceCover -> CatalogSource`, `deleteSource` (no result), `applyPromoteGraphAlignment` (no result), `promoteSubject(projectDir:userID:subjectID:entityID:) -> CatalogPromoteResult`, extension `promoteSubject(projectDir:userID:subjectID:)`, `struct CatalogPromoteResult` | — |
| GoStore | `macos/App/Platform/GoStore.swift` `updateSource`, `setSourceCover`, `deleteSource`, `promoteSubject`, `applyPromoteGraphAlignment`, `mapWriteEffects`; `CoreInvoke.swift` `CoreMethod.promoteSubject` | Apply and delete discard their responses |
| FakeStore | `macos/App/Platform/FakeStore.swift` `updateSource`, `setSourceCover`, `deleteSource`, `promoteSubject` (with refusals: `promote.invalid`, `promote.unsupported_type`, `identityclaims.already_member`, `identityclaims.type_mismatch`), `promoteSubjectDelayNanoseconds`, `applyPromoteGraphAlignment`, `workspaceNavCounts`, `nextWriteEffects` | `promoteSubject` re-implements engine refusals, against the working rule |
| Test stub | `macos/ProvenenciaTests/OnboardingModelTests.swift` (a `GenealogyStore` stub) | Implements all five methods |
| Callers | `macos/App/Features/Promote/PromoteModel.swift` `promote()`; `macos/App/Features/Sources/SourcePageModel.swift` `confirmResourceDelete` (`"source"`); `SourceIdentitySection.swift` (rename / type); `SourcePageContext.swift` `applySource(_:typeChanged:)`, `setSourceCover(mode:primaryArtifactID:)`, `refreshCoverFromStore` | Promote applies `.promotedSubject(graphId:)`; delete applies `.deletedSource(id:)`; rename, type and cover apply `.updatedSource` / `.changedSourceType` with no effects |
| Promote header | `macos/App/Features/Promote/PromoteView.swift` header; `PromoteModel.entry.sourceTitle`, `sourceRef`, `sourceMark`, `attachSource()` | The title is frozen into the location when Promote opened; `attachSource()` awaits the Sources and source-types keys, scans for the Source itself, and copies its ref and the type's icon into stored `sourceRef` / `sourceMark`. A rename or type change after load leaves all three stale |
| Breadcrumb title | `macos/App/Features/Workspace/WorkspaceToolbar.swift` (source crumb) | Writes inline `names.sourceTitle(sourceID).flatMap { $0.trimmedNonEmpty } ?? carriedTitle.flatMap { $0.trimmedNonEmpty }`; `WorkspaceLocationLabels` owns the graph half of this job (`graphName`, `graphLabel`) |
| Session names | `macos/App/Features/Workspace/Session/WorkspaceSession+Sources.swift` `source(id:)` ("the one place a Source is read by id"); `macos/App/Features/Workspace/WorkspaceLocationLabels.swift` `LocationNames(session:)` | The breadcrumbs read `session.source(id:)?.title`, which `.updatedSource` patches. `EvidenceGraphModel.canvasAccessibilityLabel` already calls `WorkspaceLocationLabels` from a model through `LocationNames(session: session)` |
| FakeStore membership | `FakeStore.swift` `promoteSubject` and `applyPromoteGraphAlignment` | Both build the mint-or-join entity (`"\(type.refPrefix)-FAKE\(n)"`), claim and `CatalogSubjectMembership`: two copies |
| Delay hook | `FakeStore.swift` `promoteProposeDelays: [Duration]` | Per-call delays consumed in call order inside `withState`; empty means no delay |
| Swift tests | `WorkspaceSessionTests` (`applyPromoteMarksHiddenPersonsListStale`, `applyTriggerEvictsHiddenDetailPagesAndRevalidatesTheVisibleOne`, `aKeyInvalidatedAndRefreshedRevalidatesOnce`, `applyRunsNavCountsHookForTriggers`, `applySkipsNavCountsHookForOtherWrites`, `deletedSourceDropsItsGraphs`); `CatalogQueryRegistryTests` (`navCountTriggersAreSourceAndPromoteWrites`, the matrix asserting `conclusionTriggers`, `personClaimsInvalidateOnPromote`, `personFamilyInvalidateOnPromote`, `personEventsInvalidateOnPromote`, `readIDs` through `.ids(in:mutation:)`); `SubjectStoreTests` promote tests; `CatalogCountsTests.refreshAllCountsPromotedPersons`; `SourcePageModelTests.confirmDeleteSourceRecountsThroughSession`; `SourceCopyModelTests.copyRecountsThroughSession`; `PromoteModelTests` (`promoteSendsItsGraph`, `donePromotesEveryRowWithATargetOnly`, `aStaleDoneReproposes`); `EvidenceGraphModelTests`, `SearchCatalogStoreTests`, `WorkspaceSidebarSectionsTests` (seed through `promoteSubject`) | — |

**Precedent.** `CreateSource` and `UnpromoteSubject` (`api/ffi/handlers/sources.go`, `promote.go`) are the handler shape: `Effects: writeEffectsProto(res)` at the next free field. S12-25's `de0b4411` ("Report every catalog-wide list a write changed as one declared set on the effects, the wire and Swift") is how a list is added: a `List` constant, an entry in `Lists`, `Lists:` on the registry entries, Swift `CatalogList` with its vocabulary guard. On Swift, `SourcesModel` creating a Source applies `session.apply(effects, .createdSource(created))`: copy it for the update, cover and delete callers. S14-11 (#657) is the precedent for deleting a mutation shim once effects carry the ids.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Go adds `ListCounts List = "counts"` *(new)*, appended to `Lists` (`[ListVocabulary, ListPropertyUsage, ListCounts]`). Its comment: "A Source or a handle (of any type) was created, deleted or merged, so a sidebar count may have changed." The `List` type doc becomes "a catalog-wide fact a write changed that no row id names: a list the app caches whole, or the sidebar counts". An association handle minted by a bridge reports `counts` too, though the sidebar counts only Person, Event and Place: one extra recount, accepted | Working rule "Every audited write returns its effects": sidebar counts become an effects list. A catalog-wide fact is a list, as `vocabulary` is | A `counts bool` on `WriteEffects` — `lists` already carries catalog-wide facts and the vocabulary guard test |
| D-2 | `Effect` gains `ListsOnCreateDelete bool` *(new)*: when set, an update never moves `Lists`. `movesLists` returns false for an update when it is set. `validate` rejects it without `Lists`, and rejects it together with `ListsOnField`. The doc comments on both fields say they are one choice: which updates move `Lists` (all, those writing one field, none) | A Source rename or cover change does not change a count, and no column of `sources` does. The flag states that directly | `ListsOnField: "id"` (an id is never updated) — a trick the next reader must decode; moving counts on every Source update — a recount query per rename |
| D-3 | `registry["sources"]` gains `Lists: []List{ListCounts}, ListsOnCreateDelete: true`. `registry["canonical_entities"]` gains `Lists: []List{ListCounts}, ListsOnField: "merged_into_id"` | The Persons, Events and Places counts filter `merged_into_id IS NULL` (`sqlCountByTypeKey`), so a merge moves them; a preferred-picture update does not. `Handles` and `HandleScope` on both entries are unchanged | `ListsOnCreateDelete` on handles too — a future merge writer would silently skip the recount |
| D-4 | `ApplyPromoteGraphAlignmentResponse` becomes `{ reserved 1, 2; reserved "revision", "written"; WriteEffects effects = 3; }`. `message ApplyPromoteGraphAlignmentWritten` is deleted. The handler binds `_, wrote, err := writes.Run(…)`, drops the revision overwrite and the `Written` loop, and returns `Effects: writeEffectsProto(wrote)`. `promote.BatchResult.Revision` and its doc go; `SaveBatch` returns `BatchResult{Written: written}` | Nothing in Swift reads either field (`GoStore` discards the response). The effects carry the revision (`WriteEffects.revision`) and the handles. A Done that writes nothing (all skips, no bridge) returns empty effects and reloads nothing; the page navigates to the canvas, which nothing changed, as before | Keeping `written` — a second report of the same handles that nothing reads |
| D-5 | `UpdateSourceResponse` and `SetSourceCoverResponse` gain `WriteEffects effects = 2`; `DeleteSourceResponse` gains `WriteEffects effects = 1`. Each handler keeps its `writes.Run` call and passes its `writes.Result` to `writeEffectsProto`. `SetSourceCover` returns the result of the `writes.Run` inside its `set` closure the same way, on both the direct path and the `thumbnailPass.run` path | `CreateSource` precedent; next free field number | — |
| D-6 | `PromoteSubject` goes: handler (`promote.go` `PromoteSubject`), `dispatch.go` `MethodPromoteSubject` and its `route` case, `METHOD_PROMOTE_SUBJECT = 85` (becomes `reserved 85; reserved "METHOD_PROMOTE_SUBJECT";` in the reserved block, with a comment "S14-16: Promote files through ApplyPromoteGraphAlignment"), and `PromoteSubjectRequest` / `PromoteSubjectResponse` (deleted; no message references them). `identityClaimProto` is deleted (its only callers are `PromoteSubject` and Apply's `Written` loop); `canonicalEntityProto` stays (other handlers call it) | A-35. Method numbers are reserved, as for `METHOD_COUNT_FILES` | Keeping the RPC for tests — tests seed through `promotetest.Save` (D-7) |
| D-7 | Handler tests seed promotions through one helper in `api/ffi/handlers/promote_test.go`: `promoteInDir(t, dir, userID, subjectID, entityID string) promote.Result` *(new)*, which parses the ids (`uuid.MustParse`), runs `withProjectCatalog(dir, …)` and `promotetest.Save` (an empty `entityID` mints). Callers that read `.GetEntity().GetId()` read `database.UUIDString(res.Entity.ID)`; `.GetEntity().GetRef()` reads `res.Entity.Ref`; `.GetClaim().GetId()` reads `database.UUIDString(res.Claim.ID)`. It replaces `promoteOK`. `promotableSubject` returns `(dir, userID, subjectID string)`; `personBeside` becomes `personBeside(t, dir, userID, subjectID string, forms ...string) string`. `TestPromoteSubject` and `TestPromoteSubjectJoin` are deleted with every case in them (bad proto, bad ids, already member, type mismatch, join): `core/database/promote/promote_test.go` covers each refusal (`ErrAlreadyMember`, `ErrTypeMismatch`, `ErrInvalid`). No other handler test asserts a `PromoteSubject` error | One seed path for core and handler tests | A handler-level seed through Apply — it needs a proposal revision per seed |
| D-8 | `TestEveryAuditedWriteReturnsEffects` *(new)* in `api/ffi/handlers/write_effects_test.go`, beside `TestRetiredFieldsStayReserved`. It walks `engine.File_engine_proto.Messages()`; for each message named `…Request` with a field `user_id`, it looks up the message `…Response` and requires a field whose message type is `provenencia.engine.v1.WriteEffects` (by `FullName` of `engine.WriteEffects` descriptor). Requests in `effectsPending` *(new)*, a `map[string]string` of request name → `"S14-17"`, are skipped. The test also fails when an `effectsPending` entry names no request or names one whose response already has the field. The walk is `func missingEffects(in file: protoreflect.FileDescriptor, pending map[string]string) []string` *(new, test file)*, which returns one message per failure; a `…Request` with `user_id` and no `…Response` message is a failure naming it. The real test calls it on `engine.File_engine_proto` with `effectsPending` and expects none. The 16 entries: `CreateSourceTypeRequest`, `UpdateSourceTypeRequest`, `DeleteSourceTypeRequest`, `CreateMetadataFieldRequest`, `UpdateMetadataFieldRequest`, `DeleteMetadataFieldRequest`, `AssignTypeFieldRequest`, `RemoveTypeFieldRequest`, `CreatePropertyRequest`, `UpdatePropertyRequest`, `DeletePropertyRequest`, `CreatePropertyTermRequest`, `UpdatePropertyTermRequest`, `DeletePropertyTermRequest`, `AssignSubjectTypePropertyRequest`, `RemoveSubjectTypePropertyRequest` | Working rule (map review K-10). Descriptors, not regex over the `.proto` text, so a nested or one-line message is seen | Checking handler code — the wire is the contract S14-17 empties |
| D-9 | Swift `CatalogList` gains `case counts = "counts"`. `CatalogQueryRegistry.invalidations(for list:project:)` returns `[]` for `.counts`. `WorkspaceSession.apply(_:_:)` recounts when `effects.lists.contains(.counts)` and `onNavCountsMoved` is set (one `trackWrite`), in place of the `navCountTriggers` check. `navCountTriggers` is deleted | The recount is not a cache key (S14-27's scope says so). Every write that changes a count now reports it: `CreateSource` and Copy Source already return effects; this PR adds Apply and `DeleteSource` | Keeping `navCountTriggers` beside the list — two mechanisms for one job |
| D-10 | `CatalogCounts.refreshAll()` takes a generation (T-12 drives it with `FakeStore.workspaceNavCountsDelays: [Duration]` *(new)*, declared as `promoteProposeDelays` is; empty means no delay. `FakeStore.workspaceNavCounts` builds its `WorkspaceNavCounts` and claims the call's delay inside one `withState`, then sleeps, then returns that snapshot, so a delayed recount returns the counts from when it was called): `private var refreshGeneration = 0` *(new)*; each call increments it and captures its value, and after the store returns it assigns the counts (or `lastRefreshError`) only when the captured value still equals `refreshGeneration` | The map's leftover: the newest recount wins. Go serializes catalog ops in call order (`catalogsession`), so the call started last reads the newest catalog | Serializing recounts in Swift — slower, and a queued recount is redundant |
| D-18 | The Sources badge has one writer, the recount: `CatalogCounts.publishSources(_:)`, `SourcesModel.syncCatalogCounts()` and the `.onSettled(sourcesHandle)` modifier in `SourcesListView` are deleted, with the `SourcesModelTests` cases that call `syncCatalogCounts()` and the `CatalogCountsTests` case that calls `publishSources` (each is deleted, or rewritten to assert the recount when it asserts more than the publish). The vocabulary badges keep `publish*` until S14-17 | Review M-13: after D-3 and D-9 every Source create, copy and delete reports `counts`, so the publish only re-sets what the recount wrote, outside D-10's ordering | Keeping the publish — two writers of one badge, and a skill that teaches both |
| D-11 | Store signatures: `updateSource(…) -> (CatalogSource, CatalogWriteEffects)`, `setSourceCover(…) -> (CatalogSource, CatalogWriteEffects)`, `deleteSource(…) -> CatalogWriteEffects`, `applyPromoteGraphAlignment(…) -> CatalogWriteEffects`. `GoStore` maps `resp.hasEffects ? Self.mapWriteEffects(resp.effects) : .empty`, as its other effect-returning methods do. `FakeStore` returns `nextWriteEffects` and keeps its current state changes. The `OnboardingModelTests` stub returns `.empty` | `createSource` / `addSourceNote` tuple shape; FakeStore working rule (scripted effects) | — |
| D-12 | Callers: `PromoteModel.promote()` applies `session.apply(effects)`. `SourcePageModel.confirmResourceDelete` (`"source"`) applies `session.apply(effects, .deletedSource(id: target.id))`. `SourcePageContext.applySource(_:effects:typeChanged:)` gains an `effects: CatalogWriteEffects` parameter (no default) and applies `session.apply(effects, typeChanged ? .changedSourceType(source) : .updatedSource(source))`; its two callers, the rename path in `SourceIdentitySection` and `SourcePageContext.setSourceCover(mode:primaryArtifactID:)`, pass the returned effects. `refreshCoverFromStore` (a read, no write) keeps `session.apply(.updatedSource(ws.source))` | The mutation still carries the enriched row (`patchedSource`) and names the Source-page key; the effects name the graphs, scopes and counts | Dropping `.updatedSource` — the effects cannot carry the row |
| D-13 | `CatalogMutation.promotedSubject(graphId:)`, `CatalogMutationKind.promotedSubject`, its `invalidationKind` arm, `evidenceGraphId` and `readIDs(for:)` are deleted. `CatalogQueryFollows.ids(in:mutation:)` becomes `ids(in effects: CatalogWriteEffects) -> [String]` and its `.reads` arm returns the effects' ids only; every caller drops the argument. `WorkspaceSession.revalidatedKeys(for:_:)` loses its `mutation` parameter (its only use was that call); `apply` and `WorkspaceSessionTests` stop passing it. `conclusionTriggers` becomes `[.updatedProperty]` with its doc saying "Property updates report no effects until S14-17". `.evidenceGraph` gets `invalidateOn: []` and keeps its required `invalidation:` closure as `{ _, _ in .allCached(.evidenceGraph) }`; its comment says every graph write, Promote and Source delete included, names its graphs in `evidenceGraphs` | Source delete reports its graphs and the graph-list read (AC-2.3); Promote reports its graph, progress and handles (AC-1.1). A Source delete touches no handle (it is refused while the Source has Subjects), so `.deletedSource` leaves `conclusionTriggers` with no loss | Keeping `.deletedSource` in `.evidenceGraph`'s `invalidateOn` — it would mark every cached canvas stale on a delete that names its own graphs |
| D-14 | A deleted Source's graph keys follow the effects like a `DeleteEvidenceGraph`: `.applyHiddenPolicy(follows: .evidenceGraphs)` marks a hidden one stale. No eviction is added | Same path as the existing graph delete; no navigation reaches a deleted Source's graph | Evicting deleted graphs — a new case for one write |
| D-15 | The Swift `promoteSubject` path goes: both protocol forms, `GoStore.promoteSubject`, `CoreMethod.promoteSubject`, `FakeStore.promoteSubject` and `promoteSubjectDelayNanoseconds`, the `OnboardingModelTests` stub method, `CatalogPromoteResult`, and the stale `CatalogMutation` doc line naming it. Tests that seed a promotion call `FakeStore.seedPromotion(projectDir:subjectID:entityID:) -> (entity: CatalogCanonicalEntity, claim: CatalogIdentityClaim)` *(new, not on the protocol)*: it writes `membershipBySubject` and `claimBySubject` as `promoteSubject` did, records no call, and throws nothing (a Subject it does not hold is a test bug: `preconditionFailure`). It is the fake's one membership writer. The write itself is `private func seedPromotionLocked(subjectID: String, type: CatalogSubjectType, entity existing: CatalogCanonicalEntity?) -> (entity: CatalogCanonicalEntity, claim: CatalogIdentityClaim)` *(new)*, called inside `withState`: it mints (`"\(type.refPrefix)-FAKE\(membershipBySubject.count + 1)"`) when `existing` is nil, else joins `existing`, then writes the claim (`accepted`) and the membership, and returns the pair (on a join, the existing entity). `FakeStore.applyPromoteGraphAlignment` keeps every guard it has today (stale revision; the row's Subject in the graph, not already a member, of a primary type; a `.handle` target that exists and has the Subject's type, else `continue`) and calls `seedPromotionLocked` in place of its mint-or-join block. `seedPromotion` looks up the Subject in any graph, its type, and, for a non-nil `entityID`, the existing entity; a missing Subject, type or entity is `preconditionFailure`; it checks nothing else (no primary-type, already-member or type-match rule: the test asked for that state). `SubjectStoreTests`' promote tests, which test the fake's refusals, are deleted | Working rule "FakeStore is a seeded fake, not a second engine"; A-35 | Keeping `promoteSubject` on the protocol for tests — a production surface with no production caller |
| D-16 | `WorkspaceLocationLabels.sourceTitle(sourceID: String, carried: String?, names: LocationNames) -> String?` *(new)* returns `names.sourceTitle(sourceID)?.trimmedNonEmpty ?? carried?.trimmedNonEmpty`; the toolbar's source crumb calls it in place of its inline rule. `PromoteModel` reads its Source row in one place: `var source: CatalogSource? { session.source(id: entry.sourceID) }` *(new)*, and `sourceTitle` (through the helper with `LocationNames(session: session)` and `entry.sourceTitle`), `sourceRef` (`source?.ref ?? ""`) and `sourceMark` (`let types: [CatalogSourceType]? = session.queryHandle(.sourceTypesList(project: session.projectKey))?.value`, as `WorkspaceSession.source(id:)` reads its key; then `types?.iconKey(forTypeID:)` for `source?.sourceTypeID`, else `.defaultTypeMark`) become computed. `attachSource()` becomes `warmSourceKeys()`, which only awaits `session.readyValue` on the two keys so the header has values when the page settles | One read of the Source row, one home for the rule that names a location's Source (as `graphName` is for graphs). Map Check: "a Source rename refreshes Promote's Source titles"; a type change moves the mark the same way | An inline rule in `PromoteModel` — a second copy of the toolbar's; keeping stored `sourceRef` / `sourceMark` — a frozen copy beside a live title |
| D-17 | `core/database/promote` files one Subject through one unexported function, `fileSubject(tx *sql.Tx, subject subjects.Subject, entityID []byte) (canonicalentities.Entity, identityclaims.Claim, []rowchange.Change, error)` *(new)*: the primary-type check, mint (`entityID == nil`) or join (`joinTarget`), and the accepted claim. `Save` keeps its argument checks, its already-member refusal and `AppendFiling`; `applyRow` keeps its target checks, its graph check and its already-on-this-handle no-op. The package doc and `Save`'s doc stop calling it the promote_subject RPC's write and say it is the single-Subject form used to seed tests and by `SaveBatch`'s rule | Review M-3: once the RPC goes, every seed would run a copy production no longer runs | Deleting `Save` — about 48 core tests seed through `promotetest.Save` |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `effects.ListCounts` *(new)* | const | `ListVocabulary`, `ListPropertyUsage` | Extends the same vocabulary (D-1) |
| `Effect.ListsOnCreateDelete` *(new)* | field | `Effect.ListsOnField` | `ListsOnField` cannot say "no update moves it" (D-2) |
| `WriteEffects effects` on four responses | proto field | `CreateSourceResponse.effects` | Same field on the remaining writes |
| `TestEveryAuditedWriteReturnsEffects`, `effectsPending` *(new)* | test, map | `TestRetiredFieldsStayReserved` (descriptor checks, same file) | No test walks every request today; searched `File_engine_proto` and `Messages()` in `api/` |
| `promote.fileSubject` *(new)* | func | `promote.Save`, `applyRow` (both keep their own checks) | The shared half of two copies (D-17) |
| `promoteInDir` *(new)* | test helper | `promoteOK` (replaced), `promotetest.Save` (wrapped) | Handler tests hold a dir, not a `*database.Catalog` |
| `CatalogList.counts` *(new)* | enum case | `CatalogList.vocabulary` | Mirrors D-1 |
| `CatalogCounts.refreshGeneration` *(new)* | property | `WorkspaceSession.generations` (per key) | Counts are not a cache key |
| `FakeStore.seedPromotion` *(new)* | test seed | `FakeStore.promoteSubject` (deleted), the mint block in `FakeStore.applyPromoteGraphAlignment` (now calls it) | One membership writer in the fake, no refusals, off the protocol |
| `FakeStore.workspaceNavCountsDelays` *(new)*: `[Duration]` | test hook | `promoteProposeDelays` (same shape) | Per call, so two recounts can finish out of order (T-12) |
| `WorkspaceLocationLabels.sourceTitle(sourceID:carried:names:)` *(new)* | func | `WorkspaceLocationLabels.graphName`; the toolbar's inline rule (moves onto it) | The Source half of the job `graphName` does |
| `PromoteModel.source` *(new)* | computed property | `WorkspaceSession.source(id:)` | The page's one read of its Source row; `sourceTitle`, `sourceRef`, `sourceMark` derive from it |

**Builds on.** The effects registry and its list vocabulary (`de0b4411`), `writeEffectsProto`, descriptor `refreshes` (S13-01, S14-11), FakeStore's `nextWriteEffects` scripting.
**Deletes.** Go: `PromoteSubject` handler and route, `MethodPromoteSubject`, `PromoteSubjectRequest` / `Response`, `ApplyPromoteGraphAlignmentWritten`, Apply's revision overwrite and `Written` loop, `promote.BatchResult.Revision`, the duplicated filing steps in `Save` and `applyRow` (D-17), `promoteOK`, `TestPromoteSubject`, `TestPromoteSubjectJoin`; `identityClaimProto` (D-6). Stale docs: the effects README's `evidence_graphs` list. Swift: `navCountTriggers`, `CatalogMutation.promotedSubject`, `CatalogMutationKind.promotedSubject`, `evidenceGraphId`, `readIDs(for:)`, the `mutation` parameters of `CatalogQueryFollows.ids` and `WorkspaceSession.revalidatedKeys`, `PromoteModel.attachSource()` and its stored `sourceRef` / `sourceMark`, the toolbar's inline Source-title rule, `CatalogCounts.publishSources`, `SourcesModel.syncCatalogCounts` and its `.onSettled` hook (D-18), the mint block in `FakeStore.applyPromoteGraphAlignment`, `.deletedSource` and `.promotedSubject` in `conclusionTriggers` and in `.evidenceGraph`'s `invalidateOn`, both `promoteSubject` protocol forms and their implementations, `CoreMethod.promoteSubject`, `promoteSubjectDelayNanoseconds`, `CatalogPromoteResult`, `SubjectStoreTests`' promote tests, `CatalogQueryRegistryTests.navCountTriggersAreSourceAndPromoteWrites`. Docs: the "report none" sentence of rule 2, the `navCountTriggers` line and both § 6 lines of `add-workspace-place` that name `updateSource` / `setSourceCover` / `deleteSource` / promote as writes that report none or apply only `.updatedSource`, the `conclusionTriggers` bullet's claim that every Conclusion key busts together, every "S14-16 removes this" comment.
**Next adds.** S14-17 converts each vocabulary write and deletes its `effectsPending` entry; when the map is empty it deletes `effectsPending` and `.updatedProperty` leaves `conclusionTriggers` (then the set is empty and goes). S14-27 moves `.vocabulary` / `.propertyUsage` onto descriptors; `.counts` stays out of that switch as a recount. A new count on the sidebar is `Lists: []List{ListCounts}` on its table's registry entry.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Source delete's effects name deleted graphs | `.evidenceGraph` refresh (`applyHiddenPolicy`), `.evidenceGraphsList` `.merged` read | Same as `DeleteEvidenceGraph` today: a hidden canvas goes stale; the graph list re-reads the deleted Source's graphs (empty) and merges, as the `readIDs` shim did |
| A handle-deleting or merging write (Unpromote, Subject delete) that reports `counts` now recounts | `CatalogCounts.refreshAll` | It already ran on Promote; the count read is the same query |
| The Promote header reads the session's Source row | `PromoteModel.source`, `sourceTitle`, `sourceRef`, `sourceMark` | Falls back to the carried title, an empty ref and the default mark when the Sources list is not warm or the Source is gone, as before `attachSource` ran today |

### Scale

At the map's Scale (3,000 Sources; 37,000 Subjects; 200,000 audit rows; 60 graphs on one census):

- **Promote Done.** Today `.promotedSubject` invalidates every Conclusion key (`allCached(.conclusionDetail)` evicts every hidden detail page, every warmed `personClaims` / `personFamily` / `personEvents`, and the three lists) and the canvas key. After: only keys whose handles are in `touchedHandles`, the one graph and its progress read, the three lists (which follow any touched handle: unchanged), and a recount only when a handle was minted. The recount is `GetWorkspaceNavCounts` (six `COUNT` queries) that already ran on every Done.
- **Source rename.** Moves the Source's graphs (60 on the census) through `EvidenceGraph` and their search documents, as the Go side already does on every rename; Swift now marks those 60 graph keys stale when hidden instead of leaving them showing the old title. No Conclusion key reloads.
- **Source delete.** A Source with Subjects or Artifacts is refused, so a delete moves at most its empty graphs. It no longer reloads every Conclusion key or every cached canvas.
- **Pin test.** One walk of about 300 messages, once per `go test`.

### Edges

- **Assumptions.** `movesLists` assumes a list is moved by every create and delete of its table; D-2 keeps that and only narrows updates. `Resolve` dedupes lists in `Lists` order (T-1 checks `counts` lands last). The session's recount assumes one hook call per apply; D-9 calls it at most once however many changes reported `counts`. `fileSubject` (D-17) assumes its caller already refused a Subject with an accepted claim; both callers keep that check before it.
- **Neighbors.** A `canonical_entities` update of `preferred_likeness_observation_id` → no `counts` (T-1). A write of `merged_into_id` → `counts` (T-1). A `sources` update of `title`, `source_type_id` or cover columns → no `counts` (T-1, T-3). A Done that only joins → no `counts` (T-4). A Done whose connection mints an association handle → `counts` and one recount the sidebar does not need (D-1, T-4). A Done that writes nothing → empty effects, nothing reloads (D-4). A Copy Source (already returns effects) → `counts` from its `sources` create, so `copyRecountsThroughSession` keeps passing with scripted effects (T-11).
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

Toolchain: STEP-2 and STEP-4 need `protoc`, `protoc-gen-go` and `protoc-gen-swift` (`./scripts/generate-proto.sh`); every Swift check needs Xcode. Generated files are never edited by hand (stop condition).

**STEP-1 — A `counts` effects list.**
Files: `core/database/effects/registry.go`, `lists_test.go`, `registry_completeness_test.go`, `README.md`. Skill: `add-audit-scope` (read only: scopes are unchanged).
Do: D-1, D-2, D-3. In `README.md`'s `Set.Lists` paragraph: list `vocabulary`, `property_usage` and `counts` (and drop the `evidence_graphs` list, which is a keyed read since S14-11), and describe `ListsOnCreateDelete` beside `ListsOnField`. Add T-1 and T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/...` passes. `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` now fails on the missing Swift `counts` case; STEP-6 makes it pass. Note it and continue.

**STEP-2 — Promote and Source handlers return their effects.**
Files: `api/proto/engine.proto`, regenerated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `core/database/promote/batch.go` (`BatchResult.Revision`), `api/ffi/handlers/promote_batch.go`, `sources.go`, `promote_batch_test.go`, `source_write_effects_test.go`, `write_effects_test.go`. Skill: `add-ffi-handler`.
Do: D-4, D-5. Rewrite `promote_batch_test.go`'s assertions on `Written` / `Revision` to assert on `GetEffects()` (T-4). Add T-3 cases. Extend `TestRetiredFieldsStayReserved` for Apply's fields 1 and 2 (T-6). Swift still compiles: `GoStore` discards both responses, and no Swift code names `ApplyPromoteGraphAlignmentWritten`.
Check: `./scripts/generate-proto.sh` runs clean; `CGO_ENABLED=1 go test -tags fts5 ./core/database/promote/... ./api/ffi/handlers/...` passes; the Xcode build succeeds.

**STEP-3 — The Swift `promoteSubject` path goes.**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `CoreInvoke.swift`, `FakeStore.swift`, `CatalogMutation.swift` (the doc line naming `promoteSubject`), `macos/ProvenenciaTests/OnboardingModelTests.swift`, `SubjectStoreTests.swift`, `CatalogCountsTests.swift`, `CatalogQueryRegistryTests.swift`, `WorkspaceSessionTests.swift`, `EvidenceGraphModelTests.swift`, `SearchCatalogStoreTests.swift`, `WorkspaceSidebarSectionsTests.swift`, `PromoteModelTests.swift` (if it relies on the fake's Apply block). Skill: `add-swift-test`.
Do: D-15. Each seeding call becomes `store.seedPromotion(projectDir:subjectID:entityID:)` (no `try await`). `EvidenceGraphModelTests`' `#expect(!store.recordedCalls.contains { $0.hasPrefix("promoteSubject") })` checks are deleted with the method. This step comes before STEP-4, whose regeneration removes `Provenencia_Engine_V1_PromoteSubjectRequest` and `.promoteSubject` from `engine.pb.swift`.
Check: `rg '\bpromoteSubject\b|CatalogPromoteResult|promoteSubjectDelayNanoseconds' macos` finds nothing (`unpromoteSubject` does not match `\bpromoteSubject\b`); `xcodebuild test` passes for `PromoteModelTests`, `SubjectStoreTests`, `WorkspaceSessionTests`, `CatalogCountsTests`.

**STEP-4 — The `PromoteSubject` RPC goes, and Promote files a Subject one way.**
Files: `api/proto/engine.proto` (+ regenerate), `core/database/promote/promote.go`, `batch.go`, `api/ffi/handlers/promote.go`, `promote_test.go`, `conclusion_detail_test.go`, `conclusion_summaries_test.go`, `promote_align_test.go`, `search_catalog_test.go`, `source_write_effects_test.go`, `person_detail_test.go`, `api/ffi/dispatch.go`, `dispatch_test.go`. Skill: `add-ffi-handler`.
Do: D-17 first (`go test ./core/database/promote/...` green), then D-6 and D-7. In `TestCallRouter`, change the "unknown method" case's number from 100 to 1000 and add `{name: "reserved promote subject", method: 85}` (T-7).
Check: `./scripts/generate-proto.sh` runs clean; `CGO_ENABLED=1 go test -tags fts5 ./...` passes; the Xcode build succeeds; `rg '\bPromoteSubject(Request|Response)?\b|MethodPromoteSubject' api core --type go -g '!*.pb.go'` finds nothing.

**STEP-5 — Every audited write is pinned to its effects.**
Files: `api/ffi/handlers/write_effects_test.go`.
Do: D-8 (T-5).
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/ -run TestEveryAuditedWriteReturnsEffects` passes; removing `effects = 1` from `DeleteSourceResponse` locally and regenerating makes it fail naming `DeleteSourceRequest` (revert).

**STEP-6 — The sidebar recounts from `counts`; the newest recount wins.**
Files: `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift`, `macos/App/Features/Catalog/CatalogCounts.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `CatalogCountsTests.swift`, `CatalogQueryRegistryTests.swift`. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-9, D-10, D-18 (files add `macos/App/Features/Sources/SourcesModel.swift`, `SourcesListView.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`). Add `FakeStore.workspaceNavCountsDelays`. Replace `applyRunsNavCountsHookForTriggers` and `applySkipsNavCountsHookForOtherWrites` with T-10; delete `navCountTriggersAreSourceAndPromoteWrites`; add T-12 and T-14's `.counts` case. Update the `onNavCountsMoved` and `refreshAll` doc comments to name `WriteEffects.lists` `counts`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; `xcodebuild test` passes for `WorkspaceSessionTests`, `CatalogCountsTests`, `CatalogQueryRegistryTests`.

**STEP-7 — The four store methods return effects and their callers apply them.**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/App/Features/Sources/SourcePageModel.swift`, `SourcePageContext.swift`, `SourceIdentitySection.swift`, and the tests that call these methods (`rg 'updateSource\(|setSourceCover\(|deleteSource\(|applyPromoteGraphAlignment\(' macos/ProvenenciaTests`; `EvidenceGraphStoreTests` calls `deleteSource`). Skill: `add-swift-test`.
Do: D-11, D-12. `PromoteModel.promote()` calls `session.apply(effects)`; the `.promotedSubject` case is left unused until STEP-8 deletes it. Update `SourcePageModelTests.confirmDeleteSourceRecountsThroughSession` and `SourceCopyModelTests.copyRecountsThroughSession` to script `nextWriteEffects = CatalogWriteEffects(lists: [.counts])` (T-11). Add T-8, T-13.
Check: `xcodebuild test` passes for `PromoteModelTests`, `SourcePageModelTests`, `SourceCopyModelTests`.

**STEP-8 — The Promote and Source-delete shims go.**
Files: `CatalogMutation.swift`, `CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `CatalogQueryRegistryTests.swift`, `EvidenceGraphModelTests.swift`.
Do: D-13, D-14. Tests that applied `.promotedSubject(graphId:)` apply `CatalogWriteEffects` naming the same graph (`evidenceGraphs`, `reads: [.init(read: .evidenceGraphProgress, ids: [g])]`) and handles. `personClaimsInvalidateOnPromote`, `personFamilyInvalidateOnPromote` and `personEventsInvalidateOnPromote` are renamed `personClaimsRevalidateOnTouchedHandle`, `personFamilyRevalidateOnTouchedHandle` and `personEventsRevalidateOnTouchedHandle` and apply `CatalogWriteEffects(handles: [h])` in place of the mutation. The matrix asserts `conclusionTriggers == [.updatedProperty]`. `deletedSourceDropsItsGraphs` scripts the delete's effects (`reads: [.init(read: .evidenceGraphs, ids: [sourceID])]`, `evidenceGraphs`) instead of relying on `readIDs` (T-9). Add T-15. Update the comments at `.evidenceGraph`, `.evidenceGraphsList` and in `CatalogQueryFollows`.
Check: `rg 'CatalogMutation\.promotedSubject|\.promotedSubject\(graphId|case promotedSubject|evidenceGraphId|readIDs\(|navCountTriggers' macos/App macos/ProvenenciaTests` finds nothing (`PVDeleteImpactPreviewData.promotedSubject` in `PVDeleteImpact.swift` and `PVDeleteImpactTests.swift` is unrelated preview data and does not match); `xcodebuild test` passes.

**STEP-9 — The Promote header reads its Source row from the session.**
Files: `macos/App/Features/Workspace/WorkspaceLocationLabels.swift`, `WorkspaceToolbar.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `PromoteView.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`, `WorkspaceToolbarBreadcrumbTests.swift`.
Do: D-16 (T-16, T-17).
Check: `xcodebuild test` passes for `PromoteModelTests` and `WorkspaceToolbarBreadcrumbTests`.

**STEP-10 — Docs.**
Files: `docs/macos-client-patterns.md`, `.agents/skills/add-workspace-place/SKILL.md`, `.agents/skills/add-ffi-handler/SKILL.md`, `docs/deployment-plan/spike-14/deployment-plan.md`, `docs/deployment-plan/spike-14/completed.md`, this plan.
Do:
- `macos-client-patterns.md` rule 2's last sentence: Source row writes and Promote report their effects and apply them beside `.createdSource` / `.updatedSource` / `.changedSourceType` / `.deletedSource`, which carry the row or name the Source page; the sidebar recounts on `WriteEffects.lists` `counts`. The "Conclusion keys" first bullet: each Conclusion key follows `touchedHandles` through its descriptor `refreshes`, and `conclusionTriggers` holds only `.updatedProperty` until S14-17.
- `add-workspace-place` § 6: "Identity/cover/title: `context.applySource(updated, effects:, typeChanged:)` → `session.apply(effects, .updatedSource(source))`"; "a Source delete applies `apply(effects, .deletedSource(id:))`; Promote applies `apply(effects)`"; the "writes that still report none" clause names only Property and other vocabulary writes (S14-17). The `navCountTriggers` line → "A write that changes a sidebar count reports `counts` in `WriteEffects.lists` (Go `effects.ListCounts` on its table's registry entry); `WorkspaceSession.apply` recounts. Features never call `CatalogCounts.refreshAll()`." The "Sidebar badges" line and the checklist's `syncCatalogCounts` item: a badge whose writes report `counts` (Sources, Persons, Events, Places) comes only from the recount; `.onSettled(handle) { model.syncCatalogCounts() }` with `publish*` is only for the vocabulary badges, until S14-17.
- `add-ffi-handler` § Catalog RPCs: add "`TestEveryAuditedWriteReturnsEffects` (`api/ffi/handlers/write_effects_test.go`) fails on a request with `user_id` whose response has no `WriteEffects`."
- The map's S14-16 row status and scope entry (the shared-header item already says "landed in #662"); `completed.md` paragraph; this plan's status and Landed note.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-16-promote-source-effects.md` passes; `python3 scripts/check-guidance-names.py` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/effects/lists_test.go` `TestResolveLists` | New rows: `source` create → `[counts]`; `source` update of `title` → no `counts`; `source` delete → `[counts]`; `canonical_entity` create → `[counts]`; `canonical_entity` update of `preferred_likeness_observation_id` → none; update of `merged_into_id` → `[counts]`; a batch with a `source_type` create and a `source` create → `[vocabulary, counts]` (order) | D-1 – D-3 |
| T-2 | Go | `core/database/effects/registry_completeness_test.go` `TestListsValidate` | `Effect{ListsOnCreateDelete: true}` → error; `Effect{Lists: [counts], ListsOnCreateDelete: true, ListsOnField: "x"}` → error; `Effect{Lists: [counts], ListsOnCreateDelete: true}` → nil | D-2 |
| T-3 | Go | `api/ffi/handlers/source_write_effects_test.go` `TestSourceWritesReturnEffects` | New cases: rename a filed Source → `source_scopes` has it, `evidence_graphs` has its graph, no `counts`; `SetSourceCover` `type_icon` → `source_scopes` has it, no graph, no `counts`; a new `TestDeleteSourceReturnsEffects` *(new)* in the same file: create a Source (it has one empty graph), delete it → `evidence_graphs` has the graph, `reads` has `evidence_graphs` with the Source id, `lists` has `counts` | AC-2.1, AC-2.2, AC-2.3 |
| T-4 | Go | `api/ffi/handlers/promote_batch_test.go` `TestApplyPromoteGraphAlignment` | "mints a person in one revision" → `effects.revision > proposal.revision`, one `handles` id whose `GetConclusionDetail` ref starts `PER-`, graph in `evidence_graphs`, progress read for the graph, `counts` in `lists`; the join case (a Subject with no connection) → `handles` has the existing handle, no `counts`; a Done whose `part_of` or kinship connection mints an association handle → `counts` in `lists` | AC-1.1, AC-1.2, D-1, D-4 |
| T-4b | Go | `core/database/promote/promote_test.go`, `batch_test.go` (existing) | Unchanged cases pass after D-17: mint, join, already member, type mismatch, unsupported type, `.handle` row already on its handle is a no-op | D-17 |
| T-5 | Go | `api/ffi/handlers/write_effects_test.go` `TestEveryAuditedWriteReturnsEffects` *(new)* | The real descriptor set → passes; subtests on a fixture file descriptor built in the test (`protodesc.NewFile` from a `descriptorpb.FileDescriptorProto` with `XRequest{user_id}` / `XResponse{}`) → fails naming `XRequest`; a stale `effectsPending` entry → fails naming it | AC-3.1, AC-3.2, D-8 |
| T-6 | Go | `api/ffi/handlers/write_effects_test.go` `TestRetiredFieldsStayReserved` | `ApplyPromoteGraphAlignmentResponse` reserves 1 and 2 and the names `revision`, `written`; the `Method` enum reserves 85 and `METHOD_PROMOTE_SUBJECT` | D-4, D-6 |
| T-7 | Go | `api/ffi/dispatch_test.go` `TestCallRouter` | method 85 → unknown-method error; "unknown method" uses 1000 | AC-3.3 |
| T-8 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `doneAppliesItsEffects` *(new)* | `store.nextWriteEffects = CatalogWriteEffects(handles: [h], evidenceGraphs: [g], lists: [.counts])`; Done → the canvas key for g revalidates, the recount hook ran once | AC-1.4, D-12 |
| T-9 | Swift | `WorkspaceSessionTests.deletedSourceDropsItsGraphs` | Delete effects with the `evidence_graphs` read for S → the graph list drops S's graphs | AC-2.4, D-13 |
| T-10 | Swift | `WorkspaceSessionTests.countsListRecounts` *(new, replaces the two nav-count tests)* | `apply(CatalogWriteEffects(lists: [.counts]))` → hook ran once; `apply(.empty, .createdSource(s))` → hook did not run; effects with `[.vocabulary]` → did not run | AC-1.4, D-9 |
| T-11 | Swift | `SourcePageModelTests.confirmDeleteSourceRecountsThroughSession`, `SourceCopyModelTests.copyRecountsThroughSession` | Scripted `[.counts]` → recount; `sources` badge as before | AC-2.4 |
| T-12 | Swift | `CatalogCountsTests.newestRecountWins` *(new)* | `workspaceNavCountsDelays = [.milliseconds(200), .zero]`; start recount 1 in a `Task`, then `_ = try await store.createSource(…)` on the fake, start and await recount 2, await recount 1 → `sources` is recount 2's value (fails with the guard removed, since recount 1 returns its earlier snapshot) | AC-4.1, D-10 |
| T-13 | Swift | `SourcePageModelTests.renameAppliesItsEffects` *(new)* | `store.nextWriteEffects = CatalogWriteEffects(evidenceGraphs: [g])`, a warmed visible canvas for g; rename → canvas key revalidated; the Sources list row shows the new title | AC-2.1, D-12 |
| T-14 | Swift | `CatalogQueryRegistryTests` | `invalidations(for: .counts, project:) == []`; matrix asserts `conclusionTriggers == [.updatedProperty]`; `.evidenceGraph`'s `invalidateOn` is empty | D-9, D-13 |
| T-15 | Swift | `WorkspaceSessionTests.promoteRevalidatesOnlyItsPeople` *(new)* | Warm `.personClaims` and `.conclusionDetail` for A (visible) and B (hidden); apply `CatalogWriteEffects(handles: [A])` → A's keys reload once; B's handles keep their values and are not stale | AC-1.3 |
| T-16 | Swift | `PromoteModelTests.headerFollowsSourceRow` *(new)* | Warm the Sources list with S "Census 1851" (type Census) and the source-types list; open Promote with carried title "Census 1851"; apply `.updatedSource(S renamed "Census 1851 (Ward 3)")` → `model.sourceTitle == "Census 1851 (Ward 3)"`; apply `.changedSourceType(S as Register)` → `model.sourceMark` is the Register icon and `model.sourceRef` is S's ref; with the Sources list not warm → the carried title, `""` and `.defaultTypeMark` | AC-2.5, AC-2.6, D-16 |
| T-17 | Swift | `WorkspaceToolbarBreadcrumbTests` (existing source-crumb cases) | Still pass with the toolbar on `WorkspaceLocationLabels.sourceTitle`; a new case: session title "  " and carried "Census" → "Census" | D-16 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `api/ffi/swift_vocabulary_test.go` (the `CatalogList` guard), `core/database/effects` `TestResolveLists` order, `WorkspaceSessionTests` and `CatalogQueryRegistryTests` (trigger sets), every handler test that seeded through `PromoteSubject`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `rg` with no hit across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/App`, `macos/ProvenenciaTests`, `api`, `core` (excluding `*.pb.go`, `engine.pb.swift` and `docs/deployment-plan/`) for: `navCountTriggers`, `CatalogMutation\.promotedSubject|\.promotedSubject\(graphId`, `evidenceGraphId`, `readIDs\(`, `CatalogPromoteResult`, `\bpromoteSubject\b` (expected unrelated hit: `core/database/conclusionsummaries/placechains_test.go`'s local `f.promoteSubject` helper, which stays), `PromoteSubjectRequest`, `MethodPromoteSubject`, `ApplyPromoteGraphAlignmentWritten`, `attachSource`, `publishSources`, "S14-16 removes"; and across `.agents/` and `docs/*.md` for "report none" / "reports none" naming a Source or Promote write; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI unchanged except the Promote header's title source (D-16)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] docs/deployment-plan/spike-14/deployment-plan.md: map row status; docs/deployment-plan/spike-14/completed.md: one short paragraph of what shipped
- [ ] Brief: N/A (no brief gates this PR)
- [ ] Domain doc updated: macos-client-patterns.md rule 2 and "Conclusion keys"; add-workspace-place § 6; add-ffi-handler § Catalog RPCs; core/database/effects/README.md
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a census Source that has at least two graphs and promoted people. Open Alice's Person page, then Bob's, then go to graph G. Open Promote on G, promote one new Person, press Done. See: the canvas shows the filed badge, G's progress updates, the Persons sidebar count goes up by one. Back to Bob's page: it shows without a reload spinner; Alice's (if she was filed) reloads.
2. Open Promote on G. In a second window on the same project, rename the Source to "Census 1851 (Ward 3)" and change its type. See: the Promote header's title and mark follow, and the breadcrumb reads the new title.
3. Make a new Source with no Artifacts. Delete it from its Source page. See: the Sources count goes down by one; the Evidence graphs list no longer shows its graph.
4. Change a Source's cover to the type icon. See: the Source row's cover changes; nothing else reloads.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Conclusion key relied on `.promotedSubject` for something `touchedHandles` does not cover (a Promote target list, a family read keyed by another handle) | T-15 and the existing `personFamily` / `personEvents` refresh tests are rewritten against effects; the manual check walks two pages |
| `SetSourceCover`'s thumbnail path returns before `writes.Run` | D-5 names both paths; T-3 covers `type_icon` and an `artifact` cover |
| The Swift build breaks when STEP-4 regenerates `engine.pb.swift` without `PromoteSubjectRequest` and `Method.promoteSubject` | STEP-3 deletes every Swift use first; STEP-4's check includes the Xcode build |
| `fileSubject` (D-17) changes what a seed or a Done files | T-4b: the existing promote and batch tests run unchanged |

**Stop and report if:**

- `protoc`, `protoc-gen-go` or `protoc-gen-swift` is missing (STEP-2, STEP-4), or Xcode is missing for a Swift check: do not hand-edit `engine.pb.go` or `engine.pb.swift`, and do not skip a Swift check.

- A `sources` update writes any column whose change moves a sidebar count (D-2's premise; check `sources.Update` and `sources.SetCover` change fields).
- Production code writes `canonical_entities.subject_type_id` (D-3's premise; `rg 'subject_type_id' core/database/canonicalentities`).
- `sources.Delete` succeeds on a Source with Subjects (D-13's premise that a delete touches no handle).
- A Conclusion key's test cannot pass from effects alone without adding a mutation to `conclusionTriggers`.
- `TestEveryAuditedWriteReturnsEffects` finds a request with `user_id` that is not in §5 D-8's list and not in this PR's four writes.
- `fileSubject` cannot serve both `Save` and `applyRow` without changing a refusal one of them returns today (T-4b).
- `.evidenceGraph`'s `invalidation:` closure cannot stay with `invalidateOn: []` (the descriptor initializer refuses it).

## 12. Open questions

None.

## 13. PR

**Title:** S14-16 — Promote and Source writes return their effects
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-16-promote-source-effects.md`.
