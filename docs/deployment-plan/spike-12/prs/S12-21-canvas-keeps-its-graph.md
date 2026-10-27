# S12-21 — The canvas keeps its graph and shows its failures

**Status:** Draft
**Spike / slice:** Spike 12 · Slice 5 — Audit fixes: High
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the two new states are `PVCallout`s in the canvas's existing callout slot, the shape of its No-artifact callout; audit A-11 "Reuses the canvas's existing callout; no new chrome")
**Board:** —
**Depends on:** —
**Blocks:** S12-37 (same files: the canvas's state becomes enums after this PR)
**Branch:** `s12-21-canvas-keeps-its-graph`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-20 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (this PR edits it), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [app-health audit](../app-health-audit.md) A-2, A-11, A-19; working rules "A destination declares its identity", "FakeStore is a seeded fake, not a second engine", "Fix the High and Medium items, grouped by shared files or behavior"; map review K-14, K-19; bar outcome 9
**Planned:** 2026-10-27 against `main` @ `e741182`
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

Going from one Evidence graph's canvas straight to another's (omnibar graph hit, Back / Forward) builds a fresh canvas for the second graph, so placing a Person puts it on the graph on screen. Today the canvas keeps the first graph's model, and a Subject placed "on Smith" lands on Jones, where it can only be deleted (A-2). The fix is structural: each workspace place declares the id its view resets on, and `WorkspaceDestinationHost` applies `.id(…)` once from that declaration, for every place. This replaces three hand-written `.id`s and fills in the canvas's missing one (working rule "A destination declares its identity").

The canvas also stops failing silently (A-11). A failed graph read shows a danger callout with **Read again**. A graph the loaded graph list no longer has (Back to a deleted graph, or a relaunch onto one) says the graph no longer exists and offers **Go to Source page**, instead of "No artifacts in this evidence graph". The composer stops swallowing a failed Citation load on an Artifact switch or a line refresh (A-19): it shows the error the way `loadCitationPreservingAuthoring` already does.

**Why now.** S12-15's graph hits made canvas → canvas a main path, so A-2 is now a common way to put a Subject in the wrong graph. S12-37 turns the canvas's state into enums and waits on this PR's files (map review K-19).

**What stays.** The canvas header, palette, cards and callout slot keep their layout. No Go, proto or catalog change. History entries of a deleted graph stay in history. The canvas's VoiceOver label and history rows still read "Evidence graph" (S12-29).

## 2. User stories

**US-1.** As a researcher on the Jones canvas, when I open the Smith graph through the omnibar or Back / Forward and place a Person, the Person is on Smith, so that I never have to delete a Subject placed in the wrong graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The host shows graph `g-a` of Source `src-1`, then navigation goes to graph `g-b` of Source `src-2` | A new `EvidenceGraphModel` is built for `g-b`: the session's `.evidenceGraph(graphId: "g-b")` and `.sourceWorkspace(sourceId: "src-2")` keys are loaded by the canvas (T-1) |
| AC-1.2 | `PlaceRegistry.standard.resolve` on two graph locations of different graph ids | Their `viewIdentity` values differ: `"g-a"` and `"g-b"` (T-2) |
| AC-1.3 | Two Source page locations `src-1` and `src-2` | `viewIdentity` is `"src-1"` and `"src-2"`; the Source page resets by remounting, and `SourcePageModel.prepare(for:)` no longer exists (T-2, D-4) |
| AC-1.4 | A composer location and a Promote location | `viewIdentity` equals `CitationComposerEntry(location:)?.identityKey` and `PromoteEntry(location:)?.identityKey`, the values the host used before (T-2) |
| AC-1.5 | A Metadata row location (`fieldId: "fld-1"`) or a Source type row location (`typeId: "typ-1"`) | `viewIdentity` is `nil`: selecting a row in a master–detail place does not remount its list (T-2) |
| AC-1.6 | Manual: census Source with graphs Jones and Smith, both linked to a page. Open Jones; omnibar "Smith", Return on the graph hit; arm Person; place "Test" | The "Test" card appears on Smith; Jones does not hold "Test" (§10 step 1) |

**US-2.** As a researcher whose graph read failed, I see that it failed and can read it again, so that an empty grid never passes for an empty graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The graph key's handle has `status == .error` (listSubjects throws) | `EvidenceGraphModel.graphReadError` is the `L10n.Errors.message(for:)` text, and the canvas shows the danger callout "This evidence graph could not be read" with that text and **Read again** (T-3) |
| AC-2.2 | The read error is showing and the store now succeeds; **Read again** (`readGraphAgain()`) | The graph key reloads, `graphReadError` is `nil`, the rows show (T-3) |
| AC-2.3 | The read failed and the handle has no value | The empty overlay "No subjects yet" does not show (it shows only on `.ready`, unchanged) (T-3 asserts `status == .error`) |

**US-3.** As a researcher who goes Back to a graph that was deleted, I am told it no longer exists and can go to its Source page, so that I do not place Subjects that every write then refuses.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The graph list handle is settled (`isSettled`) and has no row with the canvas's graph id | `session.evidenceGraphPresence(id:)` is `.gone`; `EvidenceGraphModel.graphIsGone` is `true`; `canCite` is `false`, so the palette's Place and Connect stay disabled as today (T-4) |
| AC-3.2 | The graph list has no value yet, is refetching after a write (`isFetching`), or failed | `evidenceGraphPresence(id:)` is `.unknown`; `graphIsGone` is `false`; `canCite` is `true`, so neither the gone callout nor the No-artifact callout flashes while the list refreshes (T-4) |
| AC-3.3 | The graph list has the graph | `evidenceGraphPresence(id:)` is `.present(graph)`; `canCite` is `graph.hasArtifact`, as today (T-4) |
| AC-3.4 | `graphIsGone` | `EvidenceGraphModel.canvasNotice` is `.gone`, ahead of a read error and of the No-artifact notice; the callout slot shows the warning callout "This evidence graph no longer exists" with **Go to Source page**; the empty overlay does not show (T-7, §10 step 3) |
| AC-3.5 | **Go to Source page** | `navigation.go(to: model.sourcePageLocation(…))`, the same target as the No-artifact callout (§10 step 3) |

**US-4.** As a researcher in the composer, when a Citation fails to load on an Artifact switch or a line refresh, I see the error, so that I do not type a duplicate into what looks like a blank new Citation.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Switching to Artifact `art-1`, whose first Citation `cit-2` cannot be fetched | `fields.error` is the `L10n.Errors.message(for:)` text; `activeCitationID` is `nil`; the Artifact switch completes (`selectedArtifactID == "art-1"`, `isSwitchingCitation == false`) (T-5) |
| AC-4.2 | `refreshCitationLines()` on open Citation `cit-1` when its fetch throws | `fields.error` is set; the Observation rows and connections stay as they were (T-6) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ResolvedPlace.viewIdentity` replaces `deepId`; each `PlaceRegistry` spec declares it (D-1, D-2) | STEP-1, AC-1.2 – AC-1.5 |
| `WorkspaceDestinationHost` resolves the place once and applies `.id(place.viewIdentity)` once; the three hand-written `.id`s go (D-3) | STEP-2, AC-1.1 |
| `SourcePageModel.prepare(for:)` and `resetSections()` deleted; `SourcePageContext.sourceID` becomes `let`; `sync(from:)` drops its `sourceID` guard (D-4) | STEP-2, AC-1.3 |
| `ConclusionDetailPage`'s own `.id(entityId)` deleted; the four views the host now resets drop `.task(id:)` for `.task` (D-11) | STEP-2, AC-1.1 |
| Hosting test: graph of Source 1 → graph of Source 2 (D-9) | STEP-3, T-1 |
| `WorkspaceSession.evidenceGraphPresence(id:)` derived from the graph list; `canCite` and `graphIsGone` both read it (D-5) | STEP-4, AC-3.1 – AC-3.3 |
| Canvas read-error callout with Read again; missing-graph callout with Go to Source page; the callout slot's order as one model-owned `canvasNotice` (D-6, D-7) | STEP-5, AC-2.1 – AC-2.3, AC-3.4, AC-3.5 |
| Composer: one failure helper; a failed Citation load on Artifact switch and on line refresh sets `fields.error` (D-8) | STEP-6, AC-4.1, AC-4.2 |
| `add-workspace-place` teaches the declared identity (working rule) | STEP-7 |

| Out | Home |
| --- | --- |
| The graph's name in the canvas's VoiceOver label and in history rows | S12-29 |
| Pruning history entries of a deleted graph | not planned (S12-11 Out row, audit A-11 "Known") |
| The canvas's tool and sheet flags as enums | S12-37 |
| A gone-graph state in the composer and Promote (they open on a deleted graph only from history) | not planned |
| `PVCallout` announcing itself to VoiceOver | S12-30 |
| Removing FakeStore's `citationIsInGraph` membership rule | S12-35 |
| A shared test helper for hosting a view in a window (map review M-9, declined) | S12-31 makes a new test file free of pbxproj edits; not planned before |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Place declaration | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` `PlaceRegistry.Spec.deepId`, `resolve(_:project:)` | Each spec has `deepId: (WorkspaceLocation) -> String?`: `subjectId` (Promote), `subjectId ?? citationId` (composer), `graphId` (graph), `sourceId` (Source page), `typeId`, `fieldId`, `entityId` (three conclusion details), `nil` (lists, Properties). No app code reads it |
| Resolved place | `macos/App/Features/Workspace/Session/ResolvedPlace.swift` `ResolvedPlace.deepId` | "Active deep id … nil at list roots". Read only by `PlaceRegistryTests` and `PersonsListTests` |
| Host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `body`, `presentation(for:projectDir:registry:)` | Resolves the presentation through `presentation(for:)` and switches on it; `presentation(for:)` is called by the body and by `WorkspaceDestinationHostTests`' local helper only. `.sourceGraph` arm mounts `EvidenceGraphView` with no `.id` (A-2). Hand-written `.id(entry.identityKey)` on the composer and Promote arms, `.id(navigation.currentLocation.entityId)` on `.personDetail`; none on `.eventDetail` / `.placeDetail` / `.sourcePage` |
| Conclusion detail | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (lines 143–147) | A fourth hand-written reset: `// A new handle starts with every disclosure closed.` `.id(entityId)` and `.task(id: entityId)`. Person, Event and Place detail all render through it, so Event and Place already reset on `entityId` |
| Id-keyed tasks | `EvidenceGraphView.swift` `.task(id: graphID)`, `SourcePageView.swift` `.task(id: sourceID)`, `macos/App/Features/CitationComposer/CitationComposerView.swift` `.task(id: entry.identityKey)`, `ConclusionDetailPage.swift` `.task(id: entityId)` | Each re-runs on an id change inside one view, the path A-2 shows is wrong |
| Canvas view | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `EvidenceGraphView.init`, `.task(id: graphID)` | Builds `@State model` from `graphID` and `sourceID` once; `.task(id:)` re-runs `prepare()` on the old model when the id changes |
| Canvas callout slot | `EvidenceGraphView.swift` `EvidenceGraphContent.canvasPane` (lines 293–314) | Top-right `HStack`, an `if`/`else` chain in the view: `deleteError` callout (danger, when `pendingImpact == nil`) › `fileOpenError` callout (danger, when `pendingImpact == nil`, `deleteError == nil`, `resolvedFileArtifactID == selectedFileArtifactID`) › `noArtifactCallout` (warning, when `!model.canCite`) › placing banner (`inputMode == .placing(kind)`, `armedHint(for:)`) › connecting banner (`connectArmedHint`). `graphHandle.status == .error` is never read. No test covers the order |
| Canvas empty overlay | `EvidenceGraphView.swift` `EvidenceGraphDocumentBody` (`if subjects.isEmpty, bridges.isEmpty, graphHandle.status == .ready, inputMode == .idle`) | Shows "No subjects yet" on a ready, empty graph, including a deleted one (its rows read empty) |
| Canvas model | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `canCite`, `connectRulesError`, `prepare()`, `sourcePageLocation(title:ref:)`, `private var graphKey` | `canCite` decides "missing from the list" inline: `false` for a graph missing from a list with a value (even one still refetching), `true` while the list has no value. `connectRulesError` maps a handle's error to copy (`status == .error`, `L10n.Errors.message(for:)`). `prepare()` queries the graph, fields, rules, sources, graph list, workspace and citation-count keys |
| Graph list reads | `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` `evidenceGraph(id:)`, `evidenceGraphsHandle`, `evidenceGraphsRead(sourceID:)` | The one place the graph list is read. `evidenceGraph(id:)` is `nil` both while the list has no value and when it has no such graph |
| Query handle | `macos/App/Features/Workspace/Session/QueryHandle.swift` `QueryHandle.isSettled`, `status`, `error`, `read`, `resetForInvalidation()` | `isSettled` is `status == .ready && !isFetching` |
| Re-read precedent | `macos/App/Features/Sources/SourceCopyModel.swift` `reloadWorkspace()` | `session.invalidate(key)` then `session.query(key)` |
| Source page | `macos/App/Features/Sources/SourcePageView.swift` `.task(id: sourceID)`; `macos/App/Features/Sources/SourcePageModel.swift` `prepare(for:)`, `resetSections()`; `macos/App/Features/Sources/SourcePageContext.swift` `var sourceID` | `prepare(for:)` rebinds the model in place when history switches Source (the hand-made reset the declared identity replaces). `resetSections()` has one caller, `prepare(for:)`. `context.sourceID` is assigned only by `init` and `prepare(for:)`. `sync(from:sourceID:)` guards `context.sourceID == sourceID` for that rebind; callers `SourcePageView.swift` lines 49, 247, 254, 257 and `SourcePageModel.warmFromSession()` line 120 |
| Composer switch | `macos/App/Features/CitationComposer/CitationComposerModel.swift` `selectArtifactAndLoad(_:)` | Resets fields, then `try? await loadCitation(first.citation.id)`: a failure leaves a blank Citation with no message (A-19) |
| Composer refresh | `CitationComposerModel.swift` `refreshCitationLines()` | `if let loaded = try? await fetchOpenCitation(id)`: a failure is silent |
| Composer precedent | `CitationComposerModel.swift` `loadCitationPreservingAuthoring(_:)` (lines 1012–1016), `selectCitationAndLoad(_:)` (lines 1053–1056) | Two copies of `do { try await loadCitation(id) } catch { if Task.isCancelled { return }; fields.error = L10n.Errors.message(for: error) … }` |
| Composer fetch | `CitationComposerModel.swift` `fetchOpenCitation(_:)` | Cached path when the Citation is on a cached artifact list **and** the graph handle has a value; otherwise `store.getCitation`, which `FakeStore` throws for an id not in `citationsByID` |
| Strings | `macos/App/Platform/L10n.swift` `L10n.EvidenceGraph.noArtifactTitle` / `noArtifactMessage` / `noArtifactAction`; `macos/App/Resources/Localizable.xcstrings` | The No-artifact callout's copy; the new keys sit beside it |
| Skill | `.agents/skills/add-workspace-place/SKILL.md` step 2 (line 91 "`deepId`: optional id string for history restore / selection"), step 4 (line 124), step 5 (line 163 "`SourcePageView` pattern — `.task(id: sourceID)` warms workspace key") | Says nothing about view identity; line 163 teaches the id-keyed task |
| Tests | `macos/ProvenenciaTests/PlaceRegistryTests.swift` (`deepId` expectations at lines 19, 53, 76, 93, 109, 110, 121, 277, 289, 306), `PersonsListTests.swift` line 101, `WorkspaceDestinationHostTests.swift` (presentation only), `EvidenceGraphModelTests.swift` `canCiteNeedsALinkedArtifact`, `CitationComposerModelTests.swift` `steppingWithDirtyCitationFieldsAsks` | No test mounts the host or checks a reset; no test for a failed graph read or a failed composer load |
| Test helpers | `macos/ProvenenciaTests/QueryHandleWaiting.swift` `waitUntil(timeout:_:)`, `waitForQuery`; `EvidenceGraphFixtures.swift` `evidenceGraph(_:…)`; `PVSelectPlacementTests.swift` (an `NSHostingView` in a borderless `NSWindow`) | Reused by T-1 – T-6 |

**Precedent.** The composer and Promote arms already reset on an identity key (`CitationComposerEntry.identityKey`, `PromoteEntry.identityKey`); this PR moves that declaration into the registry, where the place is defined. `SourceCopyModel.reloadWorkspace()` is the re-read; copy its two lines. `loadCitationPreservingAuthoring(_:)` is the composer's error handling; copy its `catch`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ResolvedPlace.deepId` is renamed `viewIdentity: String?`, and `PlaceRegistry.Spec.deepId` is renamed `viewIdentity`. Doc comment: "The id this place's view resets on: a different value remounts the destination. `nil` keeps one view across locations (list roots, master–detail places that follow `currentLocation`)." | Working rule: the declared identity replaces `deepId` (map review K-14), which no app code reads | Keeping `deepId` beside a new field — two ids per place, one unused |
| D-2 | Values: `.sourcePromote` → `PromoteEntry(location:)?.identityKey`; `.sourceCitationComposer` → `CitationComposerEntry(location:)?.identityKey`; `.sourceGraph` → `graphId`; `.sourceDetail` → `sourceId`; `.personDetail`, `.eventDetail`, `.placeDetail` → `entityId`; every other spec (`.sourceTypesDetail`, `.metadata`, `.properties`, all lists and section roots) → `nil` | Composer, Promote and Person keep the exact `.id` the host used, so their behavior does not change. Source types and Metadata follow `currentLocation` in their own `onChange`; remounting them on a row click would drop list scroll and the model | `typeId` / `fieldId` (today's `deepId`) — remounts the master list on every selection |
| D-3 | The host resolves the place once through `static func place(for:projectDir:registry:) -> ResolvedPlace?` *(new)*, switches on `place?.presentation ?? .sourcesList`, and applies `.id(place?.viewIdentity)` once to a `Group` around the `switch`. The static `presentation(for:projectDir:registry:)` is deleted; `WorkspaceDestinationHostTests`' private `presentation(for:)` helper returns `WorkspaceDestinationHost.place(for: location, projectDir: projectDir)?.presentation ?? .sourcesList`. The `.id` modifiers on the composer, Promote and Person arms are deleted | One place applies identity; a new place gets it by declaring a value (working rule) | `.id` per arm — the shape that missed the canvas |
| D-4 | The Source page resets through the host's `.id(sourceId)`. `SourcePageModel.prepare(for:)` and its only callee `resetSections()` are deleted, the `model.prepare(for: sourceID)` line in `SourcePageView`'s task goes, and `SourcePageContext.sourceID` becomes `let`. `sync(from:sourceID:)` becomes `sync(from:)` with no guard; its five call sites drop the argument | The working rule asks to name `prepare(for:)` as the exception or move it; with the host's `.id` a new model is built per Source, so `prepare(for:)`'s guard always returns. One reset mechanism instead of two | Keeping `prepare(for:)` as an exception — it would be dead code that looks live |
| D-5 | `enum EvidenceGraphPresence { case unknown, present(CatalogEvidenceGraph), gone }` *(new)* and `WorkspaceSession.evidenceGraphPresence(id:) -> EvidenceGraphPresence` *(new)*, both in `WorkspaceSession+EvidenceGraphs.swift`: `.present(g)` when `evidenceGraph(id:)` returns `g`; otherwise `.gone` only when `evidenceGraphsHandle` exists, `isSettled`, and has a value; otherwise `.unknown`. `EvidenceGraphModel.canCite` becomes `switch session.evidenceGraphPresence(id: graphID) { case .present(let g): g.hasArtifact; case .unknown: true; case .gone: false }` with its doc comment updated ("True while the graph list cannot say: no value yet or refetching"). `EvidenceGraphModel.graphIsGone` *(new)* is `if case .gone = session.evidenceGraphPresence(id: graphID)` (computed). No stored flag | Map review K-19: derived from the list, not a new model flag. One rule for "this graph is gone", decided once where graph-list reads live (`evidenceGraphsRead(sourceID:)` is the three-state precedent), so `canCite` and the gone notice cannot disagree while the list refetches after a write. The composer and Promote can reuse it | A `Bool` `evidenceGraphIsGone(id:)` beside `canCite`'s inline rule — two rules that disagree during a refetch (review M-1); a stored `isGone` set in `prepare()` — goes stale; reading the graph rows' failure — a deleted graph's rows can read as empty |
| D-6 | A private `EvidenceGraphModel.errorMessage<Value>(_ handle: QueryHandle<Value>?) -> String?` *(new)*: `handle.status == .error` → `handle.error.map(L10n.Errors.message(for:))`, else `nil`. `connectRulesError` returns `errorMessage(session.queryHandle(connectRulesKey) as QueryHandle<[CatalogConnectRule]>?)`. `graphReadError: String?` *(new)* returns it for the graph key's `QueryHandle<EvidenceGraphRows>` (shown whether or not a stale value is kept). `readGraphAgain()` *(new)*: `session.invalidate(graphKey)` then `session.query(graphKey)` as `QueryHandle<EvidenceGraphRows>` | The `SourceCopyModel.reloadWorkspace()` shape; only the graph rows key is re-read, because it is the one the canvas draws from | Re-running all of `prepare()` — re-queries six keys that did not fail |
| D-7 | The callout slot's order moves into the model: `enum EvidenceGraphCanvasNotice: Equatable { case gone, readError(String), deleteError(String), fileNotFound(String), noArtifact, placing(EvidencePrimaryKind), connecting }` *(new)* in `EvidenceGraphModel.swift`, and `EvidenceGraphModel.canvasNotice: EvidenceGraphCanvasNotice?` *(new, computed)*, first match wins: `graphIsGone` → `.gone`; `graphReadError` → `.readError`; then today's conditions in today's order, unchanged (`pendingImpact == nil` with `deleteError` → `.deleteError`; `pendingImpact == nil`, `deleteError == nil`, `fileOpenError`, `resolvedFileArtifactID == selectedFileArtifactID` → `.fileNotFound`; `!canCite` → `.noArtifact`; `inputMode == .placing(kind)` → `.placing(kind)`; `inputMode == .connecting` → `.connecting`). `canvasPane` replaces its `if`/`else` chain with `if let notice = model.canvasNotice { switch notice { … } }`, each arm drawing today's view (the `placing` kind type is the one `inputMode`'s `.placing` case carries). New arms: `.gone` → `PVCallout(tone: .warning, title:, message:)` with `PVButton(…, variant: .secondary, size: .sm)` **Go to Source page** calling `navigation.go(to: model.sourcePageLocation(title: sourceTitle, ref: sourceRef))`, id `evidenceGraph.gone.callout`; `.readError(message)` → `PVCallout(tone: .danger, title:, message:)` with **Read again** calling `model.readGraphAgain()`, ids `evidenceGraph.readError.callout` and `evidenceGraph.readError.readAgain`; both `.frame(width: 360, alignment: .trailing)`. The empty overlay condition in `EvidenceGraphDocumentBody` gains `!model.graphIsGone`. Both new callouts copy `noArtifactCallout`'s call shape: `title:` a `LocalizedStringResource`, `message:` a `String` (`L10n.string(L10n.EvidenceGraph.goneMessage)` for gone, the error text for read error); the callout id goes on the `PVCallout`, the `.readAgain` id on the `PVButton` | A gone graph's rows may also fail to read; "no longer exists" is the true message. The read error outranks the rest because nothing on the canvas is current without the rows. A model-owned order is unit-testable (T-7) and the next notice is one case (review M-4; S12-37 covers only the tool and sheet flags). Same component, slot and width as `noArtifactCallout` — no new chrome | Two more arms in the view's `if`/`else` — seven untested arms; a full-canvas empty state — new chrome the audit ruled out |
| D-8 | Composer: one private helper `showingLoadFailure<T>(_ work: () async throws -> T) async -> T?` *(new)* in `CitationComposerModel.swift`: returns `work()`'s value; on a thrown error returns `nil`, and sets `fields.error = L10n.Errors.message(for: error)` unless `Task.isCancelled`. Its four callers: `loadCitationPreservingAuthoring(_:)` and `selectCitationAndLoad(_:)` (each `guard await showingLoadFailure({ try await loadCitation(id) }) != nil else { return }`, keeping today's early return and behavior); `selectArtifactAndLoad(_:)` (`_ = await showingLoadFailure { try await loadCitation(first.citation.id) }`, then falls through to the existing `Task.isCancelled` check and `presentAfterIdentityChange()`); `refreshCitationLines()` (`if let loaded = await showingLoadFailure({ try await fetchOpenCitation(id) }) ?? nil { … }`: on failure the lines stay as they are, and it still `await reloadListedCitations()`) | The sibling's handling (A-19 fix), written once instead of four copies (review M-5). The Artifact switch has already happened, so the viewer must still show the new Artifact | Two more inline `catch` blocks — four copies; returning early in the switch — the viewer would keep the old Artifact's file |
| D-9 | T-1 is a hosting test in `WorkspaceDestinationHostTests`: an `NSHostingView` of `WorkspaceDestinationHost` (with `.environment(navigation)` and `.environment(session)`) in a borderless `NSWindow` (the `PVSelectPlacementTests` shape, plus `window.orderFrontRegardless()`), `FakeStore` seeded with Source `src-1` / graph `g-a` and Source `src-2` / graph `g-b`. Harness: `userID: "user-1"`, `sessionDisplayName: ""`, `store` the test's `FakeStore`, `catalogCounts: CatalogCounts(projectDir:store:)`; `navigation.projectUuid` is not set. Window `NSRect(x: 0, y: 0, width: 1200, height: 800)`, the host as its `contentView` with the same frame; both held in local `let`s to the end of the test. The store is a private `makeCanvasStore()` *(new)* in this suite that copies the subject types, properties and `subjectTypePropertiesByType` lines of `EvidenceGraphModelTests.makeStore`, then sets `sourcesByProject[projectDir]` to two `CatalogSource(id:ref:sourceTypeID:title:description:)` rows (`src-1` "Jones census", `src-2` "Smith census", `sourceTypeID: ""`) and the two graph rows. Its observable is which cache keys the canvas's model loads, not on-screen text | The model's `prepare()` queries its own graph and Source keys; the view body only reads handles. With the old shared model, `g-b`'s key and `src-2`'s workspace never load, so the assertion fails before the fix and passes after. No accessibility-tree scraping | Reading the header text from the accessibility tree — depends on offscreen accessibility; a registry-only test — does not prove the host applies the id |
| D-10 | Copy (`L10n.EvidenceGraph`, beside `noArtifact*`), key = value — comment: `goneTitle` `evidenceGraph.gone.title` = "This evidence graph no longer exists" — "Title on the canvas callout when the open Evidence graph was deleted"; `goneMessage` `evidenceGraph.gone.message` = "It was deleted. The Source page lists the evidence graphs this Source still has." — "Body on the canvas callout when the open Evidence graph was deleted"; `goneAction` `evidenceGraph.gone.action` = "Go to Source page" — "Callout action that leaves a deleted Evidence graph for its Source page"; `readErrorTitle` `evidenceGraph.readError.title` = "This evidence graph could not be read" — "Title on the canvas callout when the Evidence graph's rows failed to load; the body is the error text"; `readAgain` `evidenceGraph.readError.readAgain` = "Read again" — "Button that reads a canvas's Evidence graph again after it failed to load". Titles have no period, as `noArtifactTitle`. Each `Localizable.xcstrings` entry copies the shape of `evidenceGraph.noArtifact.action`: `comment`, `"extractionState": "manual"`, an `en` `stringUnit` with `"state": "translated"` | One key per surface with its own translator comment, as `noArtifactAction` and `L10n.Sources.formReadAgain` are | Reusing `L10n.Sources.formReadAgain` / `noArtifactAction` — their keys and comments name another surface |

| D-11 | Every reset goes through the host's `.id`: `ConclusionDetailPage`'s `.id(entityId)` and its comment line are deleted (the host's remount also starts the disclosures closed); `.task(id: graphID)` in `EvidenceGraphView`, `.task(id: sourceID)` in `SourcePageView`, `.task(id: entry.identityKey)` in `CitationComposerView` and `.task(id: entityId)` in `ConclusionDetailPage` become `.task`, bodies unchanged | With one view per id, each `id:` can never change during the view's life; the id-keyed task is what made A-2 look handled. One reset mechanism, as the working rule says (review M-2, M-3) | Keeping them — a second, dead reset path the skill keeps teaching |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ResolvedPlace.viewIdentity` *(renamed from `deepId`)* | field | `ResolvedPlace.deepId` | It is the rename (D-1) |
| `WorkspaceDestinationHost.place(for:projectDir:registry:)` *(new)* | func | `WorkspaceDestinationHost.presentation(for:projectDir:registry:)` | Replaces it: one resolve entry point (D-3) |
| `EvidenceGraphPresence`, `WorkspaceSession.evidenceGraphPresence(id:)` *(new)* | enum / func | `WorkspaceSession.evidenceGraph(id:)`, `evidenceGraphsRead(sourceID:)`, `EvidenceGraphModel.canCite` | `evidenceGraph(id:)` is `nil` for both "no list yet" and "no such graph"; this tells them apart once, beside `evidenceGraphsRead`'s three-state read, and `canCite` moves onto it |
| `EvidenceGraphCanvasNotice`, `EvidenceGraphModel.canvasNotice` *(new)* | enum / computed | `EvidenceGraphContent.canvasPane`'s `if`/`else`; `CitationComposerModel.Phase` (model-owned state precedent) | Moves the existing order into the model; S12-37's enums cover tools and sheets, not notices |
| `EvidenceGraphModel.graphIsGone`, `graphReadError`, `readGraphAgain()`, private `errorMessage(_:)` *(new)* | computed / func | `EvidenceGraphModel.connectRulesError`; `SourceCopyModel.reloadWorkspace()` | `errorMessage(_:)` is `connectRulesError`'s body made shared, and `connectRulesError` moves onto it |
| `CitationComposerModel.showingLoadFailure(_:)` *(new, private)* | func | The two `catch` blocks in `loadCitationPreservingAuthoring(_:)`, `selectCitationAndLoad(_:)` | Extracts them; four callers |
| Gone and read-error callouts *(new)* | view code | `EvidenceGraphContent.noArtifactCallout` | Same `PVCallout` + `PVButton` composition in the same slot; no new component |
| `L10n.EvidenceGraph.gone*`, `readErrorTitle`, `readAgain` *(new)* | L10n keys | `L10n.EvidenceGraph.noArtifact*`, `L10n.Sources.formReadAgain` | D-10 |

Searched for: `deepId`, `identityKey`, `.id(` in `WorkspaceDestinationHost.swift`; `isSettled`, `evidenceGraph(id:`; `Read again`, `readAgain`, `formReadAgain`; `try? await` in `CitationComposerModel.swift`; `NSHostingView` in tests.

**Builds on.** `PlaceRegistry` (the place declaration), the `WorkspaceSession+EvidenceGraphs` extension (graph-list reads), `PVCallout` in the canvas callout slot, `L10n.Errors.message(for:)`, the composer's `loadCitationPreservingAuthoring` error handling, `waitUntil` from `QueryHandleWaiting.swift`.
**Deletes.** `ResolvedPlace.deepId` and `Spec.deepId` (renamed), and the scattered `deepId` assertions in `PlaceRegistryTests` and `PersonsListTests` (T-2 is the one identity test); `WorkspaceDestinationHost.presentation(for:projectDir:registry:)`; the hand-written `.id(entry.identityKey)` on the composer and Promote arms, `.id(navigation.currentLocation.entityId)` on the Person arm, and `.id(entityId)` in `ConclusionDetailPage`; the `id:` of four `.task(id:)`s (D-11); `SourcePageModel.prepare(for:)`, `SourcePageModel.resetSections()`, the `model.prepare(for: sourceID)` call, `var` on `SourcePageContext.sourceID`, and `sync`'s `sourceID:` parameter and guard; `canCite`'s inline list rule; `connectRulesError`'s inline handle read; the canvas's callout `if`/`else` chain; the two inline composer `catch` blocks and the two `try?`; the `deepId` bullet, the "deep id (Fields/Types)" phrase and the `.task(id: sourceID)` line in `add-workspace-place`.
**Next adds.** S12-29 reads the graph name for the canvas; S12-37's state enum can carry `graphIsGone` and `graphReadError` as states. A new workspace place declares `viewIdentity` in its spec and the host resets it with no host edit. The composer and Promote can show a gone state from `evidenceGraphIsGone(id:)`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Canvas → canvas (omnibar, Back / Forward) | `EvidenceGraphView` | Resets on `graphId` (D-2, T-1) · gone callout (D-7) · read-error callout (D-7) |
| Source page → Source page (history) | `SourcePageView` | Resets on `sourceId` (D-4) · its existing `fallbackToSectionRoot()` on a failed read · unchanged |
| Person / Event / Place → another of the same kind (history) | `ConclusionDetailPage` | Already remounted on `entityId` by its own `.id`; now by the host's instead (D-11), same behavior · unchanged · unchanged |
| **Go to Source page** from a gone graph whose Source was also deleted | `SourcePageView` | Its failed workspace read already calls `navigation.fallbackToSectionRoot()` (`SourcePageView.swift`) |

### Scale

None — no write, list or loop changes. **Read again** re-reads one key; `evidenceGraphIsGone(id:)` is a dictionary lookup on the session's existing `EvidenceGraphIndex`.

### Schema

None — no catalog change.

### Go core

None — Swift only (map row).

### FFI

None —

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Gone callout | `PVCallout(tone: .warning, title:, message:)` + `PVButton(variant: .secondary, size: .sm)` | `EvidenceGraphView.swift` `EvidenceGraphContent` (private `goneCallout` *(new)*) | D-7; id `evidenceGraph.gone.callout` |
| Read-error callout | `PVCallout(tone: .danger, title:, message:)` + `PVButton(variant: .secondary, size: .sm)` | `EvidenceGraphView.swift` `EvidenceGraphContent` (private `readErrorCallout` *(new)*) | D-7; ids `evidenceGraph.readError.callout`, `evidenceGraph.readError.readAgain` |
| Identity | `.id(place?.viewIdentity)` | `WorkspaceDestinationHost.body` | D-3 |

States: gone — gone callout, palette disabled through `canCite`, no empty overlay. Read failed with no value — read-error callout over a blank grid, no empty overlay. Read failed with a stale value — read-error callout over the stale cards. Accessibility: the callouts are reachable in the canvas group's children; the button labels are their visible text.

### Strings

L10n keys *(new)*, D-10: `L10n.EvidenceGraph.goneTitle`, `goneMessage`, `goneAction`, `readErrorTitle`, `readAgain`, each with a `Localizable.xcstrings` entry. Removed: none. No action changes what it covers.

## 7. Steps

**STEP-1 — Places declare their view identity**
Files: `macos/App/Features/Workspace/Session/ResolvedPlace.swift`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `macos/ProvenenciaTests/PlaceRegistryTests.swift`, `macos/ProvenenciaTests/PersonsListTests.swift`. Skill: `add-workspace-place`.
Do: rename `deepId` to `viewIdentity` on `ResolvedPlace` and `Spec` with D-1's doc comment; set each spec's closure per D-2; delete every `#expect(… .deepId …)` line in the two test files (the lines listed in §4) and add T-2, the one table test for every place's identity.
Check: `rg -n 'deepId' macos` → no hit; the Xcode project builds; `PlaceRegistryTests` and `PersonsListTests` pass.

**STEP-2 — The host applies the identity once; the Source page drops its hand-made reset**
Files: `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift` (its private helper), `macos/App/Features/Sources/SourcePageView.swift`, `macos/App/Features/Sources/SourcePageModel.swift`, `macos/App/Features/Sources/SourcePageContext.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Features/CitationComposer/CitationComposerView.swift`. Skill: `add-workspace-place`.
Do: D-3, D-4 and D-11.
Check: `rg -n '\.id\(' macos/App/Features/Workspace/WorkspaceDestinationHost.swift` → one hit; `rg -n '\.id\(entityId\)' macos/App` → no hit; `rg -n '\.task\(id: (graphID|sourceID|entityId|entry\.identityKey)\)' macos/App` → no hit; `rg -n 'prepare\(for|resetSections|sync\(from: [a-zA-Z]+, sourceID:|presentation\(for: location, projectDir' macos/App` → no hit; `WorkspaceDestinationHostTests`, `SourcePageModelTests`, `PersonDetailTests` and `PromoteModelTests` pass.

**STEP-3 — Hosting test for canvas → canvas**
Files: `macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift`. Skill: `add-swift-test`.
Do: T-1 per D-9.
Check: T-1 passes. Then temporarily remove the `.id(…)` from STEP-2 and confirm T-1 fails; restore it (report both runs).

**STEP-4 — A graph missing from the list reads as gone**
Files: `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`.
Do: D-5 (`canCite` moves onto `evidenceGraphPresence(id:)`); add T-4.
Check: T-4 and `canCiteNeedsALinkedArtifact` pass.

**STEP-5 — The canvas shows a failed read and a gone graph**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skills: `add-ui-component`, `add-localized-string`.
Do: D-6, D-7, D-10; add T-3 and T-7.
Check: T-3, T-7 and the rest of `EvidenceGraphModelTests` pass; `python3 scripts/check-localizable-xcstrings.py` passes; `rg -n 'if model.pendingImpact == nil, let deleteError' macos/App` → no hit.

**STEP-6 — The composer shows a failed Citation load**
Files: `macos/App/Features/CitationComposer/CitationComposerModel.swift`, `macos/ProvenenciaTests/CitationComposerModelTests.swift`.
Do: D-8; add T-5 and T-6.
Check: T-5, T-6 and the rest of `CitationComposerModelTests` pass; `rg -n 'try\? await (loadCitation|fetchOpenCitation)' macos/App` → no hit; `rg -c 'fields.error = L10n.Errors.message' macos/App/Features/CitationComposer/CitationComposerModel.swift` → `1`.

**STEP-7 — Skill**
Files: `.agents/skills/add-workspace-place/SKILL.md`.
Do: step 2: replace the `deepId` bullet with "`viewIdentity`: the id the place's view resets on. `WorkspaceDestinationHost` applies `.id(place.viewIdentity)` once for every place; a different value remounts the destination and its `@State` model. Declare the deep id a view is built from (graph, Source, entity, composer or Promote entry key); declare `nil` for list roots and for master–detail places that follow `currentLocation` themselves. Do not add `.id` in a host arm and do not rebind a model in place on an id change." Step 4: add one line, "The host applies identity from the spec; an arm adds no `.id`.", and reword "selection from `currentLocation` deep id (Fields/Types)" to "selection from `currentLocation`'s row id (Fields/Types); their `viewIdentity` is `nil`". Step 5: replace "`SourcePageView` pattern — `.task(id: sourceID)` warms workspace key;" with "`SourcePageView` pattern — the host remounts it per Source through `viewIdentity`; a plain `.task` warms the workspace key;".
Check: `rg -n 'deepId|task\(id: sourceID' .agents .cursor docs/*.md` → no hit; `rg -n viewIdentity .agents/skills/add-workspace-place/SKILL.md` → at least two hits; `python3 scripts/check-guidance-names.py` passes.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-21-canvas-keeps-its-graph.md` passes; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift` `canvasToCanvasBuildsTheSecondGraphsModel` *(new)* | `FakeStore`: sources `src-1`, `src-2`; graph list `[evidenceGraph("g-a", sourceID: "src-1", artifactIDs: ["art-1"]), evidenceGraph("g-b", sourceID: "src-2", artifactIDs: ["art-2"])]`. `WorkspaceNavigation` goes to `.evidenceGraph(sourceId: "src-1", graphId: "g-a", title: "Jones")`; host mounted per D-9; `waitUntil` the session's `.evidenceGraph(graphId: "g-a")` handle `isSettled`. Navigation goes to `.evidenceGraph(sourceId: "src-2", graphId: "g-b", title: "Smith")` → `waitUntil` (2 s) the `.evidenceGraph(graphId: "g-b")` handle and the `.sourceWorkspace(sourceId: "src-2")` handle are both `isSettled` → `true` | AC-1.1, D-3, D-9 |
| T-2 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `viewIdentityResetsOnTheViewsOwnId` *(new)*, a table over every spec; the composer location is the one in `citationJumpResolvesComposerWithoutSubjectId`, the Promote one `WorkspaceLocation.promote(sourceId: "src-1", graphId: "g-1", subjectId: "sub-1", kind: .event, ref: "CEV-1", title: nil, sourceTitle: nil)` as in `WorkspaceDestinationHostTests.presentationPromote` | Resolve: graph `g-a` vs `g-b` → `"g-a"`, `"g-b"`; Source page `src-1` vs `src-2` → `"src-1"`, `"src-2"`; composer location → its `CitationComposerEntry.identityKey`; Promote location → its `PromoteEntry.identityKey`; Person / Event / Place detail `e1` → `"e1"`; Metadata `fld-1`, Source type `typ-1`, Sources root, Properties → `nil` | AC-1.2, AC-1.3, AC-1.4, AC-1.5, D-2 |
| T-3 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `graphReadFailureShowsAndReadsAgain` *(new)* | `store.listSubjectsError` set; `await model.prepare()` → graph handle `status == .error`, `model.graphReadError` non-nil. Clear `listSubjectsError`; `model.readGraphAgain()`; `waitUntil` the handle `isSettled` → `graphReadError == nil` | AC-2.1, AC-2.2, AC-2.3, D-6 |
| T-4 | Swift | `EvidenceGraphModelTests.swift` `graphPresenceFollowsTheList` *(new)* | (a) No graph list handle → `.unknown`, `graphIsGone == false`, `canCite == true`. (b) List settled with the graph → `.present`, `canCite == graph.hasArtifact`. (c) List settled without it (`store.evidenceGraphsByProject` without `g-1`, list loaded) → `.gone`, `graphIsGone`, `canCite == false`. (d) List with an error and no value (`listEvidenceGraphsError`) → `.unknown`. (e) Refetching: load the list without `g-1`, then `session.invalidate(session.evidenceGraphsKey)` and `session.query` it again without awaiting (handle `isFetching`) → `.unknown`, `canCite == true` | AC-3.1, AC-3.2, AC-3.3, D-5 |
| T-5 | Swift | `macos/ProvenenciaTests/CitationComposerModelTests.swift` `artifactSwitchShowsAFailedCitationLoad` *(new)* | `seedArtifact` plus the `art-1` `CatalogArtifact` literal of `steppingWithDirtyCitationFieldsAsks`; `seedCitation(store)` (`cit-1` on `art-0`); `store.citationsByID["cit-2"] = CatalogCitation(id: "cit-2", ref: "CIT-2", artifactID: "art-1", locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"}]}"#, transcription: "other", description: "", transcriptionUncertain: false, transcriptionNote: "")`; `makeModel(store: store, citationID: "cit-1")` (an edit entry: no dialog opens, so the switch does not stop at `pendingLeave`); `await model.prepare()`. Warm `art-1`'s list: `model.session.query(.citationsByArtifact(project: model.session.projectKey, graphId: graphID, artifactId: "art-1"))` and `await model.session.readyValue(…)`. Then remove `cit-2` from `store.citationsByID`, set `store.listSubjectsError`, and call `resetForInvalidation()` on `model.session.queryHandle(model.graphKey)` so the graph re-read fails with no value. `model.requestSelectArtifact("art-1")`; `waitUntil { !model.isSwitchingCitation && model.selectedArtifactID == "art-1" }` → `fields.error` non-nil, `activeCitationID == nil`, `selectedArtifactID == "art-1"` | AC-4.1, D-8 |
| T-6 | Swift | `CitationComposerModelTests.swift` `lineRefreshShowsAFailedCitationLoad` *(new)* | `seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-1", citationID: "cit-1")])`; `makeModel(store: store, citationID: "cit-1")`; `await model.prepare()`; note `observationRows.rows.count` (1). Remove `cit-1` from `store.citationsByID`; `model.session.invalidate(.citationsByArtifact(project: model.session.projectKey, graphId: graphID, artifactId: "art-0"))`; `await model.refreshCitationLines()` → `fields.error` non-nil, `observationRows.rows.count` unchanged | AC-4.2, D-8 |
| T-7 | Swift | `EvidenceGraphModelTests.swift` `canvasNoticeOrder` *(new)* | (a) Gone graph with a failed graph read (`listSubjectsError`) → `.gone`. (b) Graph present, failed read → `.readError(_)`. (c) Graph present with no Artifact, read ok → `.noArtifact`. (d) Graph with an Artifact, read ok, idle → `nil`. (e) As (d), `model.armedKind = .person` → `.placing(.person)`. (f) As (c), `model.armedKind = .person` → `.noArtifact` (today's order: the No-artifact callout outranks the placing banner) | AC-3.4, D-7 |

The tests script failures through FakeStore's existing `…Error` hooks and its missing-id throw only; no FakeStore change (working rule "FakeStore is a seeded fake"). A hook takes a local `struct TestFailure: Error {}` declared in the test file.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PlaceRegistryTests` (identity values), `CitationComposerModelTests` (the switch path), `PromoteModelTests` (`identityKey`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes; T-1 was seen failing without the host's .id (STEP-3)
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `rg -n 'deepId|prepare\(for:|resetSections' .agents/ .cursor/rules/ docs/*.md macos/` has no hit; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Rules this plan bends: none
- [ ] UI: only PVCallout and PVButton in the canvas callout slot; no new kit primitive
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S12-21 map row status; completed.md: one short paragraph of what shipped
- [ ] Brief: none (not gated)
- [ ] Domain doc: none changes; add-workspace-place updated (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. (AC-1.6) Open a project with a census Source that has graphs Jones and Smith, both linked to a page. Open Jones from the Sources list. Press ⌘K, type "Smith", Return on the graph hit. Arm Person and place "Test". The "Test" card appears on Smith. Back: Jones has no "Test". Forward and Back between the two canvases; each shows its own cards and the header reads the census title.
2. Repeat with Jones on one Source and Smith on another Source: after the omnibar jump the header reads the second Source's title.
3. (AC-3.4, AC-3.5) On a Source with two graphs, open graph Smith, go Back to the Source page, delete Smith (it has no Subjects), then Forward. The canvas shows "This evidence graph no longer exists" with **Go to Source page**; the palette is disabled; "No subjects yet" does not show. **Go to Source page** opens the Source page.
4. In the composer, open a Citation and switch to another Artifact: the first Citation on it loads as before (no error).
5. Go between two Source pages through history: each shows its own title, metadata and notes.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Remounting the Source page on `sourceId` loses something `prepare(for:)` kept | `prepare(for:)` cleared every section and the workspace; a remount clears the same. §10 step 5 |
| A just-created graph's canvas flashes the gone callout while the list refetches | D-5 requires `isSettled`; the place warms the list key on navigation, so a refetch is in flight |
| The hosting test cannot run `.task` offscreen | D-9 orders the window front; Stop condition below |
| Back onto a deleted graph shows the read-error callout for a moment, then the gone callout, when the rows' read fails before the graph list settles | Accepted: the final state is right, and D-7's order is deliberate. Do not reorder or delay the notices to hide it |

**Stop and report if:**

- After STEP-3's mount, the `g-a` graph handle is not settled within 2 s (the hosted canvas did not run its `.task`), or the host needs an environment object other than `WorkspaceNavigation` and `WorkspaceSession` that the test cannot build from `FakeStore`.
- A view other than those in §4 reads `ResolvedPlace.deepId` or calls `SourcePageModel.prepare(for:)`.
- With `.id(place?.viewIdentity)` applied, any existing test of the composer, Promote, Source types or Metadata fails because the view remounted.
- T-5 cannot reach `store.getCitation` without a FakeStore change (the cached path still returns `cit-2`).
- `EvidenceGraphView` reads the graph handle's error anywhere already, so D-6 would add a second reader.

## 12. Open questions

None.

## 13. PR

**Title:** S12-21 — The canvas keeps its graph and shows its failures
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence (T-1 – T-6, T-1 failing without the `.id`); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-21-canvas-keeps-its-graph.md`.
