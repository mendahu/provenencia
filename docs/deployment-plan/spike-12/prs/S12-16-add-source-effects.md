# S12-16 — Add Source reports its effects

**Status:** Planned
**Spike / slice:** Spike 12 · Slice 1 — Copy Source
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the Add Source form, toast and list rows look and behave as they do today)
**Board:** —
**Depends on:** S12-02 (landed [#485](https://github.com/mendahu/provenencia/pull/485))
**Blocks:** —
**Branch:** `s12-16-add-source-effects`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 and S12-02 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Runs on:** macOS with `protoc`, `protoc-gen-go`, `protoc-gen-swift` and Xcode (D-10)
**Context:** [S12-01 plan review M-9](S12-01-copy-source-engine.review.md), working rule "One title and type form, one Artifact checklist"; [`docs/macos-client-patterns.md`](../../../macos-client-patterns.md) cache contract items 2 and 4
**Planned:** 2026-10-19 against `main` @ `6191fc8` (refreshed after S12-17 landed in [#488](https://github.com/mendahu/provenencia/pull/488); first planned against `6ecfd88`. Of the §4 files, S12-17 changed only `SourcesModel.swift`, `SourcesListView.swift` and `SourcesModelTests.swift`, none of them in the code this PR edits)
**Review:** [`S12-16-add-source-effects.review.md`](S12-16-add-source-effects.review.md) — Fits with changes / Ready with changes → fixed, 1 round; round 2 after S12-17 refresh: fixed

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

Add Source and Copy Source refresh the Sources list the same way. Today Copy Source returns `WriteEffects` and calls `session.apply(effects, .createdSource)`, while Add Source returns no effects, applies `.createdSource` bare, and then writes the new row into the cached Sources list by hand with `session.setQueryValue`. After this PR, `CreateSourceResponse` carries `WriteEffects`, Add Source applies them with `.createdSource` exactly as Copy does, and the hand-written insert is gone. The researcher sees no change: the new Source appears in the list, its type's `usedBy` moves, the sidebar's Sources count goes up by one, and the toast and the jump to the new Source's page are as before.

**Why now.** The S12-01 plan review (M-9) found the two creates reporting a new Source two ways, and the user put this PR in the map after S12-02 (2026-10-18) so the two PRs did not both edit `SourcesModel` at once. S12-02 has landed.

**What stays.** The Add Source form, its validation, its toast and the navigation to the new page. `.createdSource` and what it invalidates in `CatalogQueryRegistry` (the Sources list and the Source types list). Copy Source's code.

## 2. User stories

**US-1.** As a researcher, when I add a Source from the Sources list, I see it in the list and in the sidebar count, as before, so that nothing about adding a Source changes for me.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The Sources list is on screen with one Source "Deed". I add "Family album" (type Photograph) | `create()` returns "Family album"; the cached Sources list holds "Deed" and "Family album"; the sidebar's Sources count is 2; the success toast is set; the form closes |
| AC-1.2 | Same, and the Photograph type had `usedBy` 0 | After the add, a reload of the Source types list shows Photograph `usedBy` 1 (the existing `.createdSource` invalidation, unchanged) |

**US-2.** As a contributor, when Add Source succeeds, the Sources list is refreshed through `WorkspaceSession.apply` from the write's effects, the way Copy Source refreshes it, so that there is one cache-update path for a new Source.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `CreateSource` RPC with a valid type and title "Passport" | The response's `effects.revision` > 0 and `effects.source_scopes` contains the new Source's id |
| AC-2.2 | The Sources list is on screen with "Deed" cached. A second Source "Will" is put into the store behind the session's back. I add "Family album" | The cached list holds all three, "Will" included. Only a reload from the store can bring "Will" in, so the list was refreshed by the session, not by a hand insert |
| AC-2.3 | Read `SourcesModel.create()` after the PR | It calls `session.apply(effects, .createdSource)` then `await catalogCounts?.refreshAll()`, and contains no `setQueryValue` and no `queryHandle(listKey)` patch |

## 3. Scope

| In | Covered by |
| --- | --- |
| `CreateSourceResponse` gains `WriteEffects effects = 2`; `handlers.CreateSource` fills it with `writeEffectsProto(res)` | STEP-1, STEP-2, AC-2.1 |
| `GenealogyStore.createSource` returns `(CatalogSource, CatalogWriteEffects)`, in the protocol, `GoStore`, `FakeStore` and the test `ThrowingStore` | STEP-3, AC-2.2 |
| `SourcesModel.create()` applies the effects with `.createdSource`, drops the `setQueryValue` insert, and recounts the sidebar with `catalogCounts?.refreshAll()` | STEP-4, AC-1.1, AC-2.2, AC-2.3 |
| The two docs that list `createSource` as a write that reports no effects | STEP-5 |

| Out | Home |
| --- | --- |
| Any change to the Add Source form or toast | not planned (map scope entry) |
| `updateSource`, `setSourceCover`, `deleteSource` reporting effects | not planned (they keep their mutation kinds per `docs/macos-client-patterns.md` cache contract item 2) |
| The sidebar Sources count after Copy Source. From the Source page (`SourcePageModel` builds the `SourceCopyModel`) it stays at the old number until the next workspace open, because `SourcesListView` publishes the count on `sourcesHandle.status` changes and a stale reload keeps status `.ready`. From a list row (`SourcesModel.openCopy`, S12-17) it lags until the researcher returns to the list. Deleting a Source (`SourcePageModel`, `.deletedSource(id:)`) has the same lag. Found while planning, not run | raised to the user 2026-10-19 as a separate fix: pass `CatalogCounts` into `SourceCopyModel` from both builders and call `await catalogCounts?.refreshAll()` after its `apply`, the D-7 line, and the same line after the Source delete's `apply`; not this PR |
| Copy Source from a Sources list row | landed in #488 (S12-17) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Proto | `api/proto/engine.proto` `CreateSourceResponse` | `{ Source source = 1; }` — no effects. `CopySourceResponse` is `{ Source source = 1; WriteEffects effects = 2; }` |
| Handler | `api/ffi/handlers/sources.go` `CreateSource` | `writes.Run(c, writes.Op{Action: "create_source", UserID: userID})` returns `s, res, err`; sets `s.Updated = res.Activity`; responds `&engine.CreateSourceResponse{Source: sp}`. `res` is used only for its activity |
| Handler (precedent) | `api/ffi/handlers/sources.go` `CopySource` | Responds `&engine.CopySourceResponse{Source: sp, Effects: writeEffectsProto(res)}` |
| Effects | `api/ffi/handlers/write_effects.go` `writeEffectsProto`; `core/database/effects/registry.go` `"sources"` | A `sources` insert yields `SourceScopes` = the new id (`SourceScope: self()`), no `Sources` (evidence) entry. `TestSourceEvidenceIsNotActivity` pins `{"sources", false, true, false}` |
| Generated | `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift` | Committed; produced by `./scripts/generate-proto.sh` (needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`) |
| Store protocol | `macos/App/Platform/GenealogyStore.swift` `createSource(projectDir:userID:sourceTypeID:title:description:)` | `async throws -> CatalogSource`. `copySource` returns `(CatalogSource, CatalogWriteEffects)` |
| Go store | `macos/App/Platform/GoStore.swift` `createSource` | Returns `Self.mapSource(resp.source)`. `copySource` returns `(Self.mapSource(resp.source), Self.mapWriteEffects(resp.effects))` |
| Fake store | `macos/App/Platform/FakeStore.swift` `createSource` | Mints a `CatalogSource` with `updated: CatalogActivity(revision: nextAuditRevision)`, bumps `nextAuditRevision`, appends. `copySource` appends then calls `bumpSource(copyID, evidence: …)`, which sets `updated` and returns `CatalogWriteEffects(sources:citationSources:sourceScopes: [id], activity:)` |
| Test store | `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore.createSource` | Throws `StoreBoom.boom`; returns `CatalogSource` |
| Model | `macos/App/Features/Sources/SourcesModel.swift` `create()` | After `store.createSource`: `session.apply(.createdSource)`; then `session.queryHandle(listKey)` → `rows.insert(created, at: 0)` → `session.setQueryValue(listKey, value: rows)`; then `syncCatalogCounts()`, closes the form, sets the toast, returns `created`. The only caller of `store.createSource` in the app. Since S12-17 it also holds `copying: SourceCopyModel?` and `openCopy(_:)`, which builds `SourceCopyModel(original:session:userID:store:)` for Copy… on a list row; this PR does not touch them |
| Model (precedent) | `macos/App/Features/Sources/SourceCopyModel.swift` `copy()` | `let (copy, effects) = try await store.copySource(…)`; `session.apply(effects, .createdSource)`; no list write |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `apply(_:_:)`, `apply(invalidations:)`, `readyValue(_:)` | `apply` patches last-updated rows, then invalidates. A visible key is invalidated and reloaded at once (the old value stays, `isFetching` true, `status` stays `.ready`); a hidden `sourcesList` is marked stale (`whenHidden: .markStale`). `readyValue` waits out an in-flight load and returns the value, or nil when the key is stale or never warmed |
| Registry | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `specs` | `sourcesList` invalidates on `.createdSource`, `.deletedSource`; `sourceTypesList` also invalidates on `.createdSource` (`usedBy`) |
| Place | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` `.sourcesList` spec | Query keys: `sourcesList`, `sourceTypesList`, `sourceGraphProgress`. Add Source opens only from `SourcesListView`, so the list key is visible when `create()` runs |
| View | `macos/App/Features/Sources/SourcesListView.swift` | `onConfirm` awaits `model.create()` then `navigation.go(to: SourcesListNavigation.pageLocation(for: created))`, which unmounts the list. Publishes the count `onAppear` and on `sourcesHandle.status` → `.ready`. Since S12-17 it also hosts `.sourceCopyDialog($model.copying)`; `SourceCopyDialog` awaits `copy()` and then goes to the copy's page |
| Counts | `macos/App/Features/Catalog/CatalogCounts.swift` `publishSources`, `refreshAll`; `macos/App/Features/Promote/PromoteModel.swift` | `publishSources` writes a count the model already holds; `refreshAll` recounts through `workspaceNavCounts` for a write with no loaded list to publish from. Promote calls `session.apply(…)` then `await catalogCounts?.refreshAll()`. `SourcesModel` holds `catalogCounts: CatalogCounts?`. `FakeStore.workspaceNavCounts` counts `sourcesByProject` |
| Docs | `docs/macos-client-patterns.md` cache contract item 2, second paragraph (line 87); `.agents/skills/add-workspace-place/SKILL.md` (line 172) | Both name `createSource` among the writes that report no `WriteEffects` |
| Tests | `api/ffi/handlers/sources_test.go` `TestCreateSourceAndWorkspace` case "creates source" | Checks title, description, audit action `create_source`, workspace read. No effects check. `TestCopySource` case "copies a source with its Artifact" checks `resp.Effects.GetRevision() <= 0 \|\| !slices.Contains(resp.Effects.GetSourceScopes(), …)` |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` `createReturnsSourceAndPublishesCount` | Warms lists with no visible location, creates, expects `sources.count == 1`, `counts.sources == 1`, toast set. Passes today only because of the hand insert: with no visible key `.createdSource` just marks the list stale |
| Tests | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `createdSourceInvalidatesTypesList` | Pins `.createdSource` → Source types list reload (AC-1.2). Unchanged |

**Precedent.** S12-01 ([#478](https://github.com/mendahu/provenencia/pull/478)) added `CopySourceResponse.effects` and S12-02 ([#485](https://github.com/mendahu/provenencia/pull/485)) applied them in `SourceCopyModel.copy()`. Copy the handler line, the `GoStore` mapping and the `apply(effects, .createdSource)` call exactly. Do not copy S12-02's test `copyRefreshesSourcesList`'s re-query: it models a hidden list, and Add Source runs with the list on screen.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Proto: `message CreateSourceResponse { Source source = 1; WriteEffects effects = 2; }`. No other proto change | Same shape and field number as `CopySourceResponse` | A new response type — the existing one gains one field, wire-compatible |
| D-2 | Handler: `out = &engine.CreateSourceResponse{Source: sp, Effects: writeEffectsProto(res)}`. Nothing else in `CreateSource` changes | `CopySource` precedent | Building effects by hand — `writeEffectsProto` is the one mapper |
| D-3 | `GenealogyStore.createSource` keeps its parameters and returns `(CatalogSource, CatalogWriteEffects)`. Doc comment: "Creates a Source in one audited write and returns it with the write's effects." | Matches `copySource`; a tuple is the store's convention for writes with effects | A new struct — no other store write uses one |
| D-4 | `GoStore.createSource` returns `(Self.mapSource(resp.source), Self.mapWriteEffects(resp.effects))` | `copySource` precedent | — |
| D-5 | `FakeStore` gains one private `insertSource(_ source: CatalogSource, projectDir: String, evidence: Bool) -> (CatalogSource, CatalogWriteEffects)` *(new)*: append to `sourcesByProject[projectDir, default: []]`, `let effects = bumpSource(source.id, evidence: evidence)`, return `(sourcesByProject[projectDir]?.first { $0.id == source.id } ?? source, effects)`. Callers run it inside their existing `withState { … }` block. `createSource` mints the row with no `updated` argument and returns `insertSource(source, projectDir: projectDir, evidence: false)`; its manual `nextAuditRevision` lines go. `copySource` replaces its own append / `bumpSource` / `stored` lookup (its last three statements) with `insertSource(copy, projectDir: projectDir, evidence: !chosen.isEmpty)`. `bumpSource` cannot return `.empty` there, because the row was just appended; `?? source` only satisfies the optional | `bumpSource` already sets `updated` and returns `sourceScopes: [id]` with a revision, as the engine does for a `sources` insert (`SourceScope: self()`, no evidence). One insert path in the fake mirrors the engine, where Copy goes through `sources.Create`; S12-03's Default graph then has one place to go in the fake | Two near-copies of append + bump + lookup in `createSource` and `copySource` (review M-2) |
| D-6 | `ThrowingStore.createSource` in `OnboardingModelTests.swift` changes only its return type | Conformance | — |
| D-7 | `SourcesModel.create()` after the store call: `let (created, effects) = try await store.createSource(…)`; `session.apply(effects, .createdSource)`; `await catalogCounts?.refreshAll()`; then the form state, toast and return as today. The `listKey` / `queryHandle` / `rows.insert` / `setQueryValue` block and the `syncCatalogCounts()` call inside `create()` are deleted (`syncCatalogCounts()` itself stays: `SourcesListView` calls it on appear and on load) | `apply` refreshes the list from the registry; the sidebar count is recounted the way Promote recounts after its write (`PromoteModel`, `CatalogCounts.refreshAll` doc: writes with no loaded list to publish from). It works whether or not the list is on screen, so every Source create (Add, Copy, S12-17) can use the same line | Waiting on `session.readyValue(sourcesListKey)` then `syncCatalogCounts()` — works only while the list is on screen, a third count path (review M-1). Publishing `sources.count + 1` — a guess. Keeping `syncCatalogCounts()` in `create()` — it reads the list before the reload lands |
| D-8 | A failed `refreshAll()` sets `catalogCounts.lastRefreshError` and leaves the badges as they were; `create()` still succeeds, closes the form, sets the toast and returns the Source. No extra handling | Same as Promote; the create already committed | Failing the create, or a second toast — the write succeeded |
| D-9 | Tests: Go extends the existing "creates source" case; Swift changes `createReturnsSourceAndPublishesCount` to put the list on screen first and adds one test, `createRefreshesListThroughSession` *(new)*. No `FakeStore` call counter | The planted-row check (AC-2.2) proves a reload without a new test hook | A `listSources` call counter on `FakeStore` — new fake state for one test |
| D-10 | The PR is built on the macOS dev machine `scripts/check-pr.sh` targets, with `protoc`, `protoc-gen-go`, `protoc-gen-swift` and Xcode installed. STEP-1 regenerates `engine.pb.go` before STEP-2, because T-1 reads `created.Effects`, which does not compile until then. Generated files come only from `./scripts/generate-proto.sh` | Generated code is committed; a hand edit drifts from the generator (S12-01 D-16) | Hand-editing `engine.pb.go` / `engine.pb.swift`; running STEP-2 before STEP-1 |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CreateSourceResponse.effects` *(new field)* | proto field | `CopySourceResponse.effects`, `IngestArtifactFileResponse.effects` | Reuses `WriteEffects`; only the field is new |
| `FakeStore.insertSource` *(new)* | helper | `FakeStore.copySource` tail, `bumpSource` | It is that tail, lifted so `createSource` and `copySource` share it (D-5) |
| `SourcesModelTests.createRefreshesListThroughSession` *(new)* | test | `SourceCopyModelTests.copyRefreshesSourcesList` | That one tests a hidden list; this one a visible list. Reuses `makeModel`, `warmLists`, `waitForQuery` |

Searched: `writeEffectsProto` callers, `mapWriteEffects`, `bumpSource`, `refreshAll` and `publishSources` callers, `setQueryValue` callers in `Features/Sources`.

**Builds on.** `writeEffectsProto` (Go), `GoStore.mapWriteEffects`, `FakeStore.bumpSource`, `WorkspaceSession.apply(_:_:)`, the registry's `.createdSource` invalidations, `CatalogCounts.refreshAll` (the Promote recount).
**Deletes.** The "or create" in the `syncCatalogCounts()` and `publishSources` doc comments. The hand insert into the Sources list in `SourcesModel.create()` (seven lines); the manual `nextAuditRevision` bump in `FakeStore.createSource` and the duplicated insert tail in `FakeStore.copySource` (both go through `insertSource`); the `syncCatalogCounts()` call in `create()`; `createSource` from the "report no effects" lists in `docs/macos-client-patterns.md` and `add-workspace-place`.
**Next adds.** S12-03 makes `sources.Create` mint a Default graph in the same write. The new `evidence_graphs` row's effect (`Set.EvidenceGraphs`, S12-03 / S12-05) then reaches Add Source through these effects with no Add Source code, as working rule "The graph list has its own effect" expects for Copy Source.

### Schema

None — no catalog change.

### Go core

None — `core/` does not change. The effects are already computed by `writes.Run`.

### FFI

`api/proto/engine.proto`: `CreateSourceResponse` gains `WriteEffects effects = 2;` (D-1). Handler `api/ffi/handlers/sources.go` `CreateSource` per D-2. No new method; `api/ffi/dispatch.go` and `CoreMethod` do not change. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Store method | `GenealogyStore.createSource` | `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`; `macos/ProvenenciaTests/OnboardingModelTests.swift` | D-3 – D-6 |
| Add Source save | `SourcesModel.create()` | `macos/App/Features/Sources/SourcesModel.swift` | D-7, D-8 |

```swift
// SourcesModel.create(), after validation and isSaving
let (created, effects) = try await store.createSource(
    projectDir: session.projectKey.projectDir,
    userID: userID,
    sourceTypeID: draft.sourceTypeID,
    title: draft.trimmedTitle,
    description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
)
session.apply(effects, .createdSource)
await catalogCounts?.refreshAll()
// isAdding / errors / toast / return as today
```

States: unchanged. A failed list reload after a successful create leaves the old rows (the session keeps a stale value on failure); a failed recount leaves the badge (D-8). The create still succeeds and navigates.

### Strings

None — no copy changes.

## 7. Steps

**STEP-1 — Proto field**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: D-1, D-10. Run `./scripts/generate-proto.sh`. Never hand-edit the generated files.
Check: `git diff --stat` shows the proto and both generated files; `CGO_ENABLED=1 go build -tags fts5 ./api/...` passes.

**STEP-2 — Handler returns effects**
Files: `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`.
Do: D-2. Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/ -run 'TestCreateSourceAndWorkspace|TestCopySource'` passes.

**STEP-3 — Store returns effects**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/App/Features/Sources/SourcesModel.swift` (only the destructuring `let (created, _) =`, so the tree builds; STEP-4 finishes it).
Do: D-3 – D-6.
Check: the Xcode build of the `Provenencia` scheme succeeds (`xcodebuild build` as `scripts/check-pr.sh` runs it, or a full `scripts/check-pr.sh`).

**STEP-4 — Add Source applies the effects**
Files: `macos/App/Features/Sources/SourcesModel.swift`, `macos/App/Features/Catalog/CatalogCounts.swift` (doc comments only), `macos/ProvenenciaTests/SourcesModelTests.swift`. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-7, D-8. Change T-2; add T-3. Fix three doc comments the change makes false: `SourcesModel.syncCatalogCounts()` becomes "Refreshes sidebar totals from the cached list (after a warm load)."; `CatalogCounts.publishSources` drops "or create" ("Writes the Sources total the feature model already knows from its in-memory list (after load)."); `CatalogCounts.refreshAll` ends "Conclusion handle counts have no loaded list to publish from, so a Promote and Add Source recount through here."
Check: `rg -n "setQueryValue|rows.insert" macos/App/Features/Sources/SourcesModel.swift` prints nothing; `rg -n "refreshAll" macos/App/Features/Sources/SourcesModel.swift` prints the one D-7 line; `xcodebuild test` (ProvenenciaTests) passes `SourcesModelTests`, `SourceCopyModelTests`, `CatalogQueryRegistryTests`, `OnboardingModelTests`.

**STEP-5 — Docs**
Files: `docs/macos-client-patterns.md`, `.agents/skills/add-workspace-place/SKILL.md`, `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status).
Do: in `macos-client-patterns.md` cache contract item 2 (second paragraph, line 87), change "The exceptions are the Source row's own writes (`createSource`, `updateSource`, `setSourceCover`, `deleteSource`) and promote, which keep their mutation kinds." to "The exceptions are the Source row's own writes and promote, which keep their mutation kinds: `createSource` reports its effects and keeps `.createdSource` (a new row the effects cannot describe); `updateSource`, `setSourceCover`, `deleteSource` and promote report none." In `add-workspace-place` line 172, drop `createSource` from "the writes that still report none" and add: "`createSource` returns `WriteEffects` and is applied with `apply(effects, .createdSource)`, as Copy Source is." After line 156 ("`syncCatalogCounts()` when list handle reaches `.ready`"), add: "A write that changes a sidebar count is applied first (`session.apply`), then `await catalogCounts?.refreshAll()`. Publish from the list only on load: the list may be hidden or still reloading when the write returns." Then the Definition of done's docs lines.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-16-add-source-effects.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits. Suggested: "Report Add Source's write effects over FFI." (STEP-1 – STEP-2) and "Refresh the Sources list after Add Source the way Copy Source does." (STEP-3 – STEP-5).

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `api/ffi/handlers/sources_test.go` `TestCreateSourceAndWorkspace` case "creates source" (extended) | Create "Passport" → `created.Effects.GetRevision() > 0` and `slices.Contains(created.Effects.GetSourceScopes(), created.Source.GetId())`; `t.Fatalf("effects %+v", created.Effects)` otherwise, as `TestCopySource` | AC-2.1, D-2 |
| T-2 | Swift | `macos/ProvenenciaTests/SourcesModelTests.swift` `createReturnsSourceAndPublishesCount` (changed) | Before `openAdd()`, add `session.apply(location: .sectionRoot(.sources))`. Create "Family album", then `await waitForQuery(listHandle)` on the Sources list handle (`session.queryHandle(SourcesModel.sourcesListKey(for: session))`, from `macos/ProvenenciaTests/QueryHandleWaiting.swift`): `create()` does not wait for the reload `apply` starts → `created?.title == "Family album"`, `model.sources.count == 1`, `counts.sources == 1`, toast set, `!model.isAdding` | AC-1.1, D-7 |
| T-3 | Swift | `macos/ProvenenciaTests/SourcesModelTests.swift` `createRefreshesListThroughSession` *(new)* | `makeModel` with sources `[source(id: "s1", title: "Deed", typeID: "t1")]`, types `[photoType()]`, a `CatalogCounts`; `warmLists`; `session.apply(location: .sectionRoot(.sources))`; then `store.sourcesByProject[projectDir]?.append(source(id: "s2", title: "Will", typeID: "t1"))`; create "Family album", then `await waitForQuery(listHandle)` on the Sources list handle as in T-2 → `Set(model.sources.map(\.title)) == ["Deed", "Will", "Family album"]` and `counts.sources == 3` | AC-2.2, AC-1.1, D-7 |
| — | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `createdSourceInvalidatesTypesList` (unchanged) | Already pins `.createdSource` → types list reload | AC-1.2 |
| — | Review | STEP-4 check `rg` | No `setQueryValue` / `rows.insert` in `SourcesModel.swift` | AC-2.3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcesModelTests` (create tests relied on the hand insert; `openCopySetsCopyingForThatSource` from S12-17 must stay green), `OnboardingModelTests` (`ThrowingStore` conformance), every Go handler test that unmarshals `CreateSourceResponse` (they ignore the new field and stay green).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI change
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/macos-client-patterns.md cache contract item 2 and .agents/skills/add-workspace-place/SKILL.md (STEP-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with at least one Source, on the Sources list. Note the sidebar Sources count (N) and a type's `usedBy` on the Source types page.
2. Add Source: type Photograph, title "Family album". See the success toast and the new Source's page open; the sidebar count reads N + 1.
3. Go back to the Sources list. "Family album" is in it with no reload flicker.
4. Open Source types. Photograph's `usedBy` is one higher than in step 1.
5. Copy "Family album" from its page (S12-02), and copy it again from its Sources list row's right-click menu (S12-17). Each copy opens; going back to the list shows both. (The sidebar count after a copy is the known Out defect.)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The recount holds the form a moment longer | One local `workspaceNavCounts` read, as after Promote; `isSaving` keeps Create disabled, so no double submit |
| The new row's place in the list changes (the hand insert put it first) | The list view sorts by `settings.sort` (`visibleSources`), so the insert position never reached the screen |
| A Go test elsewhere compares a whole `CreateSourceResponse` with `proto.Equal` | Stop condition below |

**Stop and report if:**

- `SourcesModel.create()` is no longer the only caller of `store.createSource` (`rg -n "\.createSource\(" macos/App`).
- T-3 fails because "Will" is missing after the create and after waiting for the list handle: the list did not reload on `.createdSource`, so the registry row in §4 is not true.
- A Go test fails because it compares a full `CreateSourceResponse` (`proto.Equal` or a `want` bytes match) and would need a changed expectation outside `sources_test.go`.
- `./scripts/generate-proto.sh` cannot run where you are (no `protoc`, `protoc-gen-go` or `protoc-gen-swift`), or Xcode is missing for STEP-3 / STEP-4. Do not hand-edit generated files; report the Go evidence you have.

## 12. Open questions

None.

## 13. PR

**Title:** S12-16 — Add Source reports its effects
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-16-add-source-effects.md`.
