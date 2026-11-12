# S15-26 — The metadata board takes an owner

**Status:** Planned
**Spike / slice:** Spike 15 · Slice 2 — Metadata at three levels
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Source page's Metadata board is drawn by the same views with the same copy, spacing and identifiers; only the types behind it change)
**Board:** —
**Depends on:** — (S15-08, which gave the four board RPCs an owner on the wire, landed in [#791](https://github.com/mendahu/provenencia/pull/791))
**Blocks:** S15-09, S15-10, S15-11
**Branch:** `s15-26-metadata-board-takes-an-owner`
**Version bump:** patch, 0.10.0 → **0.10.1** (or the next patch after `main`'s `VERSION` when you rebase): `VERSION`, `core/version.go`, and every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj` — two response fields old clients ignore (`SetMetadataResponse.metadata`, `ClearMetadataResponse.metadata`; `SetMetadataResponse.entry` reserved, its only reader `GoStore` ships in the same bundle), no screen or catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (its Mutations section changes), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (the board stays a feature snowflake), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "One metadata board", "Metadata lives at three levels with one vocabulary", "One level type" and "Metadata saves with its owner's write" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); S15-08's [plan](S15-08-artifacts-and-graphs-hold-metadata.md) D-11, D-16 and **Next adds**, and its [ship log](S15-08-artifacts-and-graphs-hold-metadata.ship.md) nit on `.mutatedSourceMetadata`; S15-07's [plan](S15-07-metadata-operations-take-an-owner.md) D-7 (a draft board sends its whole order); [S15-D1](../design/S15-D1-source-page.md) SD-9 and its inventory row "Metadata board" (what S15-09 builds on this); [`docs/macos-client-patterns.md`](../../../macos-client-patterns.md) § mutations
**Planned:** 2026-11-12 against `main` @ `7d2f3e7`
**Review:** [`S15-26-metadata-board-takes-an-owner.review.md`](S15-26-metadata-board-takes-an-owner.review.md) — Fits with changes — Ready with changes, 2 rounds; every finding Fixed or answered; nothing To user
**Runs on:** macOS. STEP-1's `./scripts/generate-proto.sh` needs `protoc`, `protoc-gen-go` and `protoc-gen-swift`; every Swift check and `scripts/check-pr.sh` need Xcode ([`README.md`](../../../../README.md)). Without them, stop: never hand-edit the generated `engine.pb.go` or `engine.pb.swift`

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

The Source page's Metadata board becomes one board model and one board view that work on any metadata owner (a Source, an Artifact, an Evidence graph), so S15-09 (a board on each Artifact card), S15-10 (the add flow) and S15-11 (the graph form) instance the same board instead of copying it. The Swift store's four metadata writes take a `MetadataOwner` (level and id), matching the owner RPCs S15-08 put on the wire. The Add metadata dialog and the clear confirm become one page-level model that targets a board, so a page with many boards still has one of each. Set and Clear return the owner's rows like dismiss and reorder do, so Swift stops re-deriving where the engine puts a row and what a clear leaves. The Source page moves onto the new board and looks and behaves exactly as today.

**Why now.** Working rule "One metadata board" (map review K-2, kept by the user 2026-11-09): S15-09, S15-10 and S15-11 each need the board, and all three depend on this PR. S15-08 left the Swift store's Source-named methods for this PR (its D-16).

**What stays.** Every pixel, string, accessibility identifier, keyboard path and error placement of the Source page's Metadata section, the Add metadata dialog and the clear confirm. Each write still makes one store call and one `session.apply`, which invalidates the page's `sourceWorkspace` key as today (§6 Scale). Go's metadata rules are untouched; the Set and Clear handlers only answer the list they read (Set already reads it).

## 2. User stories

**US-1.** As a researcher, when I use the Metadata board on a Source page after this PR, it behaves exactly as before, so that nothing I rely on moved.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source with saved "Repository: NRO" and an unfilled suggestion "Author" | The board shows Repository in the reorderable list and Author under "Suggested by this type", as today |
| AC-1.2 | Type "Mary Robins" in Author's suggestion row, press Return | Author moves to the saved list with "Mary Robins"; no suggestion is left |
| AC-1.3 | Edit a saved URL value to `file:/tmp` and save; the engine answers `sourcemetadata.invalid` | The error shows under that row (not in the page banner); the row stays in edit mode |
| AC-1.4 | A save fails with an I/O error | The page banner shows the message with no title |
| AC-1.5 | Drag a saved row; the reorder write fails | The page banner shows the failure and the rows return to their earlier order |
| AC-1.6 | Clear a saved value that was a type suggestion, confirm | The row returns to the suggestions as an empty row |
| AC-1.7 | Add metadata from the dialog with no field chosen | The field shows "Choose a field." (`metadataFieldRequired`); nothing is written |
| AC-1.8 | A write finds the Source deleted in another window | The page leaves with the deleted notice and the board writes no state |
| AC-1.9 | Add metadata from the dialog; the engine answers `sourcemetadata.invalid` | The error shows under the dialog's value field; the dialog stays open |

**US-2.** As a contributor building S15-09 – S15-11, I instance the board on any owner by giving it a host that holds the owner's rows and runs its writes, so that every board in the app behaves the same and a page keeps one Add dialog and one clear confirm.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A `MetadataBoardModel` on a test host with rows `[Author (suggested, empty)]`; save "Mary Robins" on Author | The host received `set(fieldID: "f-author", valueText: "Mary Robins")`, its `entries` hold Author saved, and `didWrite` was called once after `entries` changed |
| AC-2.2 | Same host; the reorder write fails | The host's `entries` are back in their earlier order and the host's `showFailure` got the message |
| AC-2.3 | Same host; `failureMessage(for:)` returns nil (the owner is gone) | No field error, no banner, `entries` unchanged |
| AC-2.4 | Clear Author on a host whose `clear` returns `[Author suggested, empty]` | The host's `entries` are that list; `didWrite` was called |
| AC-2.5 | One `MetadataBoardDialogs` and two boards A and B; open Add on B and save | B's host received the write; A's did not; the dialog closes |
| AC-2.6 | Ask to clear on A, confirm | A's host received `clear`; `pendingDelete` is nil afterwards |

**US-3.** As the engine, Set and Clear answer with the owner's rows, so that a client never re-derives where a row goes or what a clear leaves.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ClearMetadata` on a saved value of a suggested field | `ClearMetadataResponse.metadata` holds that field with no value, still suggested |
| AC-3.2 | `ClearMetadata` on a value of a field that is not suggested | `metadata` no longer lists the field |
| AC-3.3 | `SetMetadata` on a field with no row yet, on an owner with one saved value | `SetMetadataResponse.metadata` lists the new value after the saved one, in the owner's layout order |

## 3. Scope

| In | Covered by |
| --- | --- |
| `SetMetadataResponse.metadata` and `ClearMetadataResponse.metadata` *(new fields)*, filled from `sourcemetadata.ListWorkspace` as `DismissMetadataSuggestion` does | STEP-1, AC-3.1 – AC-3.3 |
| `MetadataOwner` *(new)* value type; the four store methods renamed and taking an owner (`GenealogyStore`, `GoStore`, `FakeStore`, `ThrowingStore`); `setMetadata` and `clearMetadata` return the list | STEP-2, AC-1.2, AC-1.6 |
| `FakeStore` holds metadata rows in one map keyed by owner | STEP-2 |
| `MetadataBoardHost` *(new)* protocol: the owner's rows, its writes, its field vocabulary and where its failures show | STEP-3, AC-2.1 – AC-2.4 |
| `SourcePageMetadataHost` *(new)*: a host over `SourcePageContext` for one owner on the Source page | STEP-3, AC-1.1 – AC-1.8 |
| `SourceMetadataSection` becomes `MetadataBoardModel` *(renamed)*, driven by a host, without dialog state | STEP-4, AC-1.1 – AC-1.8 |
| `MetadataBoardDialogs` *(new)*: the page's one Add dialog and one clear confirm, each targeting a board | STEP-4, AC-1.7, AC-1.9, AC-2.5, AC-2.6 |
| `.mutatedSourceMetadata` and `SourcePageContext.applyMetadataWrite` deleted; every board write applies `.mutatedSourceWorkspace` | STEP-5 |
| `MetadataBoardView` *(new)* and `MetadataBoardAddForm` *(new)* split out of `SourcePageMetadataView`; the Source page instances them | STEP-6, manual check |
| Board and dialog tests against a test host | STEP-7, T-2 – T-8 |

| Out | Home |
| --- | --- |
| A board on each Artifact card; mapping `GetSourceWorkspaceResponse.artifact_metadata` into Swift; a host per Artifact | S15-09 |
| Patching the cached `sourceWorkspace` from a board write instead of reloading it (review M-2; the Go read is about 2 ms at Scale, §6) | S15-09, only if its own measurement of 60 drawn boards shows the reload is slow; not planned otherwise |
| A draft host (the board on a form, collecting a `sourcemetadata.Draft` with its whole order, no store writes) and the Swift `Draft` type sent on `CreateArtifactRequest.metadata` | S15-10 (S15-11 reuses it) |
| `GenealogyStore.getMetadataBoard` over `METHOD_GET_METADATA_BOARD` | S15-11 (its first reader) |
| Renaming `SourcePageMetadataLabel`, `SourcePageMetadataValueChrome`, `SourcePageLayout.metadata*`, the private row views and the `L10n.Sources.metadata*` keys to board names | S15-09 decides when a second view first draws them; not planned before that |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Set and Clear handlers | `api/ffi/handlers/metadata.go` `SetMetadata`, `ClearMetadata` | `SetMetadata` reads `sourcemetadata.ListWorkspace(c, owner)` after the write, then answers only the one entry (or `internal.unknown` if missing). `ClearMetadata` runs `sourcemetadata.Clear` in `writes.Run`; answers `ClearMetadataResponse{Effects}` only. `DismissMetadataSuggestion` and `ReorderMetadata` in the same file also answer `Metadata: metadataEntriesProto(entries)` from `sourcemetadata.ListWorkspace(c, owner)` after the write |
| Set and Clear proto | `api/proto/engine.proto` `message SetMetadataResponse`, `message ClearMetadataResponse` | `MetadataWorkspaceEntry entry = 1; WriteEffects effects = 2;` and `WriteEffects effects = 1;` |
| Handler test | `api/ffi/handlers/metadata_test.go` `TestMetadataOwners` | Reads `set.GetEntry()`; clears `m.suggested` and `m.extra`, checks only `source_scopes`. No other Go code reads `SetMetadataResponse.entry` (`rg GetEntry api`) |
| Board model | `macos/App/Features/Sources/SourceMetadataSection.swift` `SourceMetadataSection` | `@Observable` class over `SourcePageContext`. Rows from `context.workspace?.metadata`; writes through `context.store.setSourceMetadata` / `clearSourceMetadata` / `dismissSourceMetadataSuggestion` / `reorderSourceMetadata` with `context.sourceID`; patches `context.workspace?.metadata` (`replaceEntry` puts a set's one row at its index or appends it: a copy of the engine's layout placement in `sourcemetadata.Set` / `ListWorkspace`), then `context.applyMetadataWrite(effects)` (set, clear) or `context.applyWrite(effects)` (dismiss, reorder). `applyCleared` re-derives a clear's result (a suggested row empties, any other row goes), the engine's `sourcemetadata.Clear` rule. Holds the Add dialog state (`isAdding`, `addFieldID`, `addValue`, `addFieldError`, `addValueError`, `isSavingAdd`) and the clear confirm (`pendingDelete`, `isClearing`). Failures: `context.clearPageError()` before each write; `applySetError` sends `sourcemetadata.invalid` to `fieldError` or `addValueError`, anything else to `context.showPageError`; clear, dismiss, reorder use `context.showFailure`; `context.fail` returns nil for a deleted Source. Field options from `context.fields` |
| Delete item | same file `SourceMetadataDeleteItem` | `fieldID`, `label`; used by `SourcePageView`'s `pvConfirm(item:)` |
| Board view | `macos/App/Features/Sources/SourcePageMetadataView.swift` `SourcePageMetadataView` | Takes `SourcePageModel`; draws the section header (title, count, Add), the intro, the empty state, the `PVReorderableList` of saved rows and the suggestion rows through `model.metadata`. `addForm` is the Add dialog's body. Private `SourcePageMetadataSuggestionRow` and `SourcePageMetadataTextEditor` (a comment names `SourceMetadataSection.drafts`). Identifiers are literal `"sources.page.metadata…"` / `"sources.page.addMetadata…"` strings |
| Page wiring | `macos/App/Features/Sources/SourcePageModel.swift` `metadata`, `init`, `apply(_:)`, `acceptsArtifactPageDrop` | Builds `SourceMetadataSection(context:)`; `metadata.resetDrafts()` when the workspace's metadata changed; reads `metadata.isAdding` and `metadata.pendingDelete` |
| Page shell | `macos/App/Features/Sources/SourcePageView.swift` `pvFormDialog` (Add metadata, around line 164), `pvConfirm` (clear, around line 183), `addMetadataPresented` (around line 310) | Read and write `model.metadata`'s dialog and confirm state |
| Page context | `macos/App/Features/Sources/SourcePageContext.swift` `applyWrite`, `applyMetadataWrite` | `.mutatedSourceWorkspace(sourceId:)` and `.mutatedSourceMetadata(sourceId:)` |
| Mutation | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogMutation.mutatedSourceMetadata`, `CatalogMutationKind.mutatedSourceMetadata` | Its comment says it "also moves each field's `usedBy`"; since S15-08 the field list follows the effects' `metadata_field_usage` instead, so the comment is false |
| Registry | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `.sourceWorkspace` (around line 240) | The only kind that lists `.mutatedSourceMetadata`, in the same branch as `.mutatedSourceWorkspace`: both invalidate the one `sourceWorkspace` key |
| Store protocol | `macos/App/Platform/GenealogyStore.swift` the four `…SourceMetadata…` methods (around line 1380) | Take `sourceID: String`; `clearSourceMetadata` returns `CatalogWriteEffects` |
| Go store | `macos/App/Platform/GoStore.swift` same four (around line 336) | Fill `req.owner` with `Level.source.rawValue` and `sourceID` (S15-08 D-16) |
| Fake store | `macos/App/Platform/FakeStore.swift` `metadataBySource` (stored), the four methods (around line 724), `setSourceMetadataError`, `reorderSourceMetadataError`, `sourcesHoldingField` | One stored map by Source id; `clearSourceMetadata` holds a third copy of the Clear rule |
| Throwing stub | `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore` (around line 424) | `throw StoreBoom.boom` for the four |
| Guidance | `.agents/skills/add-workspace-place/SKILL.md` lines 172 and 175; `docs/macos-client-patterns.md` line 89 | Teach `context.applyMetadataWrite` and list `.mutatedSourceMetadata` among the kept mutations |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` metadata tests (lines 698 – 1105, and around 2230); `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` lines 723, 740, 937, 949, 1398 | Drive the board and dialog through `model.metadata`; list `.mutatedSourceMetadata` among mutations |

**Precedent.** `WorkspaceLeaveGuard` (`macos/App/Features/Workspace/WorkspaceNavigation.swift`) for the shape of a small `@MainActor` `AnyObject` protocol a feature model conforms to. S15-08 (#791) for `GoStore` filling `req.owner`. `DismissMetadataSuggestion` in `api/ffi/handlers/metadata.go` for a write answering the owner's list.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ClearMetadataResponse` gains `repeated MetadataWorkspaceEntry metadata = 2;`; `ClearMetadata` reads `sourcemetadata.ListWorkspace(c, owner)` after `writes.Run`, exactly as `DismissMetadataSuggestion` does. `SetMetadataResponse` becomes `reserved 1; reserved "entry"; WriteEffects effects = 2; repeated MetadataWorkspaceEntry metadata = 3;`; `SetMetadata` answers the `entries` it already reads and drops its find-the-row loop and the `internal.unknown` fallback | All four writes answer the owner's whole row set: one return shape, one board path, and the client stops copying `sourcemetadata.Clear`'s rule and the layout placement of `sourcemetadata.Set` (reviews M-5, M-11) | Keeping `entry` beside `metadata` — a second answer nothing reads; keeping Swift's `applyCleared` and `replaceEntry` — engine rules in the client, reached by three levels |
| D-2 | `struct MetadataOwner: Sendable, Hashable { var level: Level; var id: String }` *(new)* in `macos/App/Platform/GenealogyStore.swift`, beside `CatalogSourceWorkspace`. Swift does not check `level.holdsMetadata`; Go answers `sourcemetadata.invalid` | Mirrors the proto `MetadataOwner` and Go's `sourcemetadata.Owner`: "owner" is the identity everywhere | A Swift enum with associated ids — a second spelling of `Level` |
| D-3 | The store methods become `setMetadata(projectDir:userID:owner:fieldID:valueText:) -> ([CatalogMetadataEntry], CatalogWriteEffects)`, `clearMetadata(projectDir:userID:owner:fieldID:) -> ([CatalogMetadataEntry], CatalogWriteEffects)`, `dismissMetadataSuggestion(projectDir:userID:owner:fieldID:) -> ([CatalogMetadataEntry], CatalogWriteEffects)`, `reorderMetadata(projectDir:userID:owner:fieldIDs:) -> ([CatalogMetadataEntry], CatalogWriteEffects)`. The Source-named methods are deleted | Names match the RPCs (S15-08 D-11); one path per operation | Overloads keeping `sourceID:` — two paths for one write |
| D-4 | `FakeStore` stores `metadataByOwner: [MetadataOwner: [CatalogMetadataEntry]]` *(new)*. `metadataBySource` becomes a computed `[String: [CatalogMetadataEntry]]` get/set view over the `.source` owners, so every existing seed and read compiles unchanged: the getter builds the dictionary from the `.source` keys; the setter removes every `.source` key, then stores each entry of the new dictionary under `MetadataOwner(level: .source, id:)`; neither takes `stateLock` (callers are already inside `withState`, as for the computed `suggestionsByType`). The four methods read and write `metadataByOwner[owner]`; `bumpActivity` runs only for a `.source` owner; `setMetadata` keeps the fake's scripted result (replace at the row's index, else append) and returns the list; `clearMetadata` keeps its scripted result (a suggested row empties, any other row goes) and returns the list; both as `dismissMetadataSuggestion` returns its list (`swift-tests.mdc`: the fake seeds and scripts). `setSourceMetadataError` → `setMetadataError`, `reorderSourceMetadataError` → `reorderMetadataError`. No level validation | One stored map, so a Source delete or copy cannot clear one map and miss another (review M-6) | Two stored maps and a picker — a split brain |
| D-5 | `@MainActor protocol MetadataBoardHost: AnyObject` *(new)* in `macos/App/Features/Sources/MetadataBoardHost.swift` *(new)*, exactly: `var entries: [CatalogMetadataEntry] { get set }`; `var fields: [CatalogMetadataField] { get }`; `func set(fieldID: String, valueText: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects)`; `func clear(fieldID: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects)`; `func dismissSuggestion(fieldID: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects)`; `func reorder(fieldIDs: [String]) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects)`; `func didWrite(_ effects: CatalogWriteEffects)`; `func clearFailure()`; `func failureMessage(for error: Error) -> String?` (nil: the owner is gone and its host is leaving; write no state); `func showFailure(_ message: String)`. The board patches `entries` before calling `didWrite` | "Host" names the role (rows, writes, where failures show); "owner" stays the identity (review M-8). One `didWrite` since both mutations now do the same thing (D-8; review M-3). Patch, then apply, is today's order | A struct of closures — ten unnamed closures; a `valuesChanged` flag — a dead distinction every host must take |
| D-6 | `final class SourcePageMetadataHost: MetadataBoardHost` *(new)*, same file: `init(context: SourcePageContext, owner: MetadataOwner, rows: @escaping (CatalogSourceWorkspace) -> [CatalogMetadataEntry], setRows: @escaping (inout CatalogSourceWorkspace, [CatalogMetadataEntry]) -> Void)`, plus `static func source(context: SourcePageContext) -> SourcePageMetadataHost` with `owner: MetadataOwner(level: .source, id: context.sourceID)`, `rows: { $0.metadata }`, `setRows: { $0.metadata = $1 }`. `entries` get is `context.workspace.map(rows) ?? []`; set applies `setRows` to `context.workspace` when it is not nil. `fields` is `context.fields`. Writes call the D-3 methods with `context.projectDir`, `context.userID`, `owner`. `didWrite` → `context.applyWrite(effects)`. `clearFailure` → `context.clearPageError()`; `failureMessage(for:)` → `context.fail(_:)`; `showFailure` → `context.showPageError(_:)` | S15-09 instances the same type per Artifact with its own owner and rows (review M-9) | One class per level — a near-copy in S15-09; `extension SourcePageContext: MetadataBoardHost` — one board per page |
| D-7 | `SourceMetadataSection` is renamed `MetadataBoardModel` and moves to `macos/App/Features/Sources/MetadataBoardModel.swift` *(new)* (`git mv`); `init(host: any MetadataBoardHost)`, stored as `let host: any MetadataBoardHost` (read by `MetadataBoardDialogs`). It keeps the row state and operations with their names and behavior: `drafts`, `editingFieldID`, `savingFieldID`, `fieldError`, `fieldErrorID`, `entries`, `saved`, `suggested`, `fieldComboOptions`, `resetDrafts`, `beginEdit`, `cancelEdit`, `save`, `dismissSuggestion`, `moveSaved`. The dialog state and its operations move to D-9. Mapping of each `context.*` use: `workspace?.metadata` → `host.entries`; `fields` → `host.fields`; store calls → `host.<op>`; `applyMetadataWrite` and `applyWrite` → `host.didWrite`; `clearPageError` → `host.clearFailure`; `fail` → `host.failureMessage(for:)`; `showPageError` → `host.showFailure`; `context.showFailure(e)` → one private `report(_ error: Error)` *(new)*: `if let m = host.failureMessage(for: e) { host.showFailure(m) }` (review M-7). `replaceEntry` and `applyCleared` are deleted: a set sets `host.entries` to the returned list, sets `drafts[fieldID]` to that field's `valueText` in it, and calls `didWrite` (`save` then exits edit mode as today); a clear sets `host.entries` to the returned list, sets `drafts[fieldID] = ""`, exits edit mode on that row, calls `didWrite`, then `syncDrafts()`. `moveSaved` still restores `previous` after a failure whether or not the owner is gone, as today. Two methods *(new)* serve D-9: `add(fieldID: String, valueText: String) async -> MetadataBoardAddResult` and `clear(fieldID: String) async -> Bool` (true when cleared) | A mechanical move of the row behavior the page tests pin; the dialog leaves the board so many boards share one dialog (review M-1) | A fresh model — risks the focus and lag fixes the current one carries |
| D-8 | Delete `CatalogMutation.mutatedSourceMetadata`, `CatalogMutationKind.mutatedSourceMetadata`, its `kind` mapping, its entry in `.sourceWorkspace`'s `invalidateOn` and switch, and `SourcePageContext.applyMetadataWrite`. Update `CatalogQueryRegistryTests` at the five lines in §4 (remove the case from each list; delete the `invalidations(by: .mutatedSourceMetadata…)` expectation). `add-workspace-place` line 172 reads: "Metadata values: a board's host applies the write's effects with the page's own mutation (`SourcePageMetadataHost.didWrite` → `context.applyWrite(effects)`); a value created or released on a Source, Artifact or Evidence graph reports `metadata_field_usage`, which the field list's `usedBy` follows." Line 175 and `docs/macos-client-patterns.md` line 89 drop `mutatedSourceMetadata` from their lists | Since S15-08 the two mutations reach the one same key; the field list follows `metadata_field_usage` (review M-3) | Keeping both — a dead mutation and a false comment |
| D-9 | `@MainActor @Observable final class MetadataBoardDialogs` *(new)* in `macos/App/Features/Sources/MetadataBoardDialogs.swift` *(new)*: the page's one Add dialog and one clear confirm. State: `private(set) var addTarget: MetadataBoardModel?`, `var isAdding: Bool { addTarget != nil }`, `addFieldID`, `addValue`, `addFieldError`, `addValueError`, `private(set) var isSavingAdd`, `var pendingDelete: MetadataBoardDeleteItem?`, `private(set) var isClearing`. Operations: `openAdd(on board:)`, `cancelAdd()` (no-op while saving), `fieldComboOptions` (the target's, or `[]`), `addSheetIsDirty`, `canSubmitAdd`, `createFromAdd()` (today's validation with today's strings, then `await target.add(…)`: `.saved` closes, `.invalid(message)` sets `addValueError`, `.failed` keeps the dialog open), `askClear(on board:, fieldID:)` (only for a saved row, as today), `confirmClear()` (`if await item.board.clear(fieldID:) { pendingDelete = nil }`). Guards and flags exactly as today: `confirmClear` returns when `pendingDelete` is nil or `isClearing`, and wraps `isClearing` in a `defer`; `createFromAdd` returns when `addTarget` is nil and wraps `isSavingAdd` in a `defer`. `fieldComboOptions` is `addTarget?.fieldComboOptions ?? []`; the board keeps its own `fieldComboOptions` (the tests near `SourcePageModelTests` lines 882 and 902 read it on the board). `MetadataBoardModel.add` calls `host.clearFailure()` first, as `createFromAdd` called `clearPageError` today. `enum MetadataBoardAddResult { case saved, invalid(String), failed }` *(new)*, same file. `struct MetadataBoardDeleteItem: Identifiable, Equatable` *(renamed from `SourceMetadataDeleteItem`)* adds `let board: MetadataBoardModel`; `==` compares `fieldID`, `label` and `ObjectIdentifier(board)`. `MetadataBoardModel.add` returns `.invalid(message)` for `sourcemetadata.invalid`, shows any other failure through the host and returns `.failed`, and returns `.failed` with no state written when `failureMessage` is nil. `SourcePageModel` gains `let metadataDialogs = MetadataBoardDialogs()` (a `var` if `@Bindable` needs a writable path, as the comment on the section properties says); `acceptsArtifactPageDrop` reads `metadataDialogs.isAdding` and `metadataDialogs.pendingDelete`; `SourcePageView`'s dialog, confirm and `addMetadataPresented` read `model.metadataDialogs`, and the header and empty-state Add buttons call `model.metadataDialogs.openAdd(on: model.metadata)` | One owner of "which board's Add or clear is open", so S15-09 adds no dialog state and the drop check stays one read (review M-1) | Dialog state per board with a page pointer to the active one — two owners of one fact |
| D-10 | `MetadataBoardView` *(new)*, `macos/App/Features/Sources/MetadataBoardView.swift` *(new)*: `@Bindable var board: MetadataBoardModel; let accessibilityIdentifierPrefix: String; let onAdd: () -> Void`. It draws what `SourcePageMetadataView.body` draws below the intro: the empty state (its Add calls `onAdd`), the saved list, the suggestions; `onClear` on a saved row calls a `let onClear: (String) -> Void` parameter. `SourcePageMetadataSuggestionRow` and `SourcePageMetadataTextEditor` move into this file, still `private`, names unchanged. `MetadataBoardAddForm` *(new)*, same file: `@Bindable var dialogs: MetadataBoardDialogs; let accessibilityIdentifierPrefix: String`, today's `addForm`. Every literal identifier becomes `"\(accessibilityIdentifierPrefix)…"` with the same suffix; the page passes `"sources.page.metadata"` and `"sources.page.addMetadata"`, so every identifier string is unchanged. `SourcePageMetadataView` keeps the section header (Add keeps `"sources.page.addMetadata"`) and the intro, then `MetadataBoardView(board: model.metadata, accessibilityIdentifierPrefix: "sources.page.metadata", onAdd: { model.metadataDialogs.openAdd(on: model.metadata) }, onClear: { model.metadataDialogs.askClear(on: model.metadata, fieldID: $0) })`; its `addForm` is deleted | S15-D1 SD-9 draws the Artifact board without the Source's header and intro; dialog actions go to the page's one `MetadataBoardDialogs` | The header in the board — the Artifact card draws a Disclosure instead |
| D-11 | Board copy stays `L10n.Sources.metadata*`; no string, key or `.xcstrings` entry changes | No chrome change; S15-09's brief decides any Artifact wording | A new namespace now — a rename with no second caller |
| D-12 | New Swift tests in `macos/ProvenenciaTests/MetadataBoardModelTests.swift` *(new)* drive `MetadataBoardModel` and `MetadataBoardDialogs` through `TestMetadataBoardHost` *(new, private to that file)*: records each call with its arguments in order (`didWrite` included), returns scripted results or throws a scripted error, returns a scripted `failureMessage`. Existing `SourcePageModelTests` metadata tests stay as the Source-host tests; their edits are limited to the D-4 fake renames and moving dialog calls from `model.metadata.<x>` to `model.metadataDialogs.<x>` (`openAdd()` becomes `openAdd(on: model.metadata)`, `askClear(fieldID:)` becomes `askClear(on: model.metadata, fieldID:)`) | The map's Check: "Swift tests drive the board against a fake owner" | A test that pins `FakeStore`'s own maps — tests test infrastructure (review M-10) |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ClearMetadataResponse.metadata` *(new field)* | proto field | `DismissMetadataSuggestionResponse.metadata` | Same shape on the sibling (D-1) |
| `MetadataOwner` *(new)* | type | `Provenencia_Engine_V1_MetadataOwner` (generated) | Generated types stay inside `GoStore` (`macos-client.mdc`) |
| `MetadataBoardHost` *(new)* | protocol | `SourcePageContext` | The context is one page's dependencies; a board needs one owner's rows and writes (searched `rg "protocol " macos/App/Features`) |
| `SourcePageMetadataHost` *(new)* | type | `SourcePageContext` | Extending the context allows one board per page (D-6) |
| `MetadataBoardModel` *(renamed)* | type | `SourceMetadataSection` | Rename and re-home |
| `MetadataBoardDialogs`, `MetadataBoardAddResult` *(new)* | type | the dialog half of `SourceMetadataSection` | Moved out so one dialog serves many boards (D-9) |
| `MetadataBoardView`, `MetadataBoardAddForm` *(new)* | view | `SourcePageMetadataView` | Split from it (D-10) |
| `metadataByOwner` *(new)* | fake state | `FakeStore.metadataBySource` | Becomes the stored map; the old name a computed view (D-4) |
| `TestMetadataBoardHost` *(new)* | test helper | `FakeStore` | The board's contract is the protocol, not the store |

**Builds on.** `Level` (S15-06), the owner RPCs (S15-08), `DismissMetadataSuggestion`'s list answer, `SourcePageContext`'s failure door and `applyWrite`, the `WorkspaceLeaveGuard` protocol shape.
**Deletes.** `SourceMetadataSection` and `SourceMetadataDeleteItem` (renamed); `SourceMetadataSection.applyCleared`; `SourcePageMetadataView.addForm`; the four `…SourceMetadata…` store methods in `GenealogyStore`, `GoStore`, `FakeStore` and `ThrowingStore`; the fake's `setSourceMetadataError` and `reorderSourceMetadataError`; `CatalogMutation.mutatedSourceMetadata`, `CatalogMutationKind.mutatedSourceMetadata` and their registry entries; `SourcePageContext.applyMetadataWrite`; the guidance lines in D-8. No string is deleted.
**Next adds.** S15-09 maps `artifact_metadata`, builds a `SourcePageMetadataHost` per Artifact with `owner: MetadataOwner(level: .artifact, id:)` and rows from its board, puts `MetadataBoardView` in each card's Disclosure, and wires its Add and clear to the page's one `MetadataBoardDialogs`. S15-10 adds a draft host. A draft host computes no engine-built part of a row: its `set` returns its rows with that row holding the typed value trimmed, `hasValue: true`, an empty `openHref` and its `suggested` flag from the rows it was given (a new field appended at the end); its `dismissSuggestion` removes that suggestion from its own rows and records it in the draft; its `clear` empties a row that came as a suggestion and removes any other (nothing is stored to ask the engine, so the draft host is the one client place this rule lives); its `reorder` returns its rows in the given order; `sourcemetadata.invalid` arrives at the form's one save, not at a field, so the board's field-error path does not run for a draft (review M-4). S15-11 reuses the draft host and adds `getMetadataBoard`.

### Reach

None — no new path into existing code, no new state reachable: the only host this PR builds is the Source page's, over the same workspace, and the one `MetadataBoardDialogs` only ever targets that board.

### Scale

Each board write makes one store call and one `session.apply(effects, .mutatedSourceWorkspace(sourceId:))`, as today (set and clear used `.mutatedSourceMetadata`, which reached the same key). That mutation invalidates and, while the page is shown, reloads the whole `sourceWorkspace` (`WorkspaceSession.applyHiddenPolicy` → `invalidate` + `revalidate`): one `GetSourceWorkspace`, which since S15-08 carries every Artifact's board. At the map's Scale (60 Artifacts) one field save on the Source's board reloads 60 Artifact boards. Measured at planning (a throwaway handler test on `TestMetadataOwners`' fixture, 60 Artifacts with 3 values each, 4 Source values, 30 runs, Linux): `GetSourceWorkspace` averages **2.3 ms** and a Source-level `SetMetadata` **1.8 ms**. The Go side of the reload is not worth a patch path; the Swift decode and redraw of 60 boards is S15-09's to measure when it draws them, and `didWrite` is the one place a host would change to patch instead. Set's response adds no query (it already read the list); Clear's adds one `ListWorkspace(c, owner)` for one owner, the read dismiss and reorder already make.

### Edges

- **Assumptions.** The board assumes a write's returned list is the owner's whole row set in display order; dismiss and reorder already return that from `ListWorkspace`, and set and clear now return it from the same call (T-1). The page assumes one Add dialog and one confirm; `MetadataBoardDialogs` holds one target (T-6, T-7).
- **Neighbors.** Clearing a suggested row keeps it as an empty suggestion; clearing a non-suggested row removes it (AC-3.1, AC-3.2, T-1); dismiss and reorder unchanged.
- **Removed inputs.** A board whose host's Source is deleted mid-write writes no state (T-4, T-10). An Add dialog opened on a board stays bound to that board's host until it closes.
- **Real data.** None — no table-driven rule.
- **Untrusted size.** None — no list from data the app did not write is walked by new code.

### Schema

None — no catalog change.

### Go core

None — `sourcemetadata` is unchanged; the handler calls the existing `ListWorkspace`.

### FFI

`api/proto/engine.proto`: `message ClearMetadataResponse { WriteEffects effects = 1; repeated MetadataWorkspaceEntry metadata = 2; }` and `message SetMetadataResponse { reserved 1; reserved "entry"; WriteEffects effects = 2; repeated MetadataWorkspaceEntry metadata = 3; }`. Handlers in `api/ffi/handlers/metadata.go`: `ClearMetadata`, after `writes.Run`, `entries, err := sourcemetadata.ListWorkspace(c, owner)` and answers `Metadata: metadataEntriesProto(entries)`; `SetMetadata` answers `Metadata: metadataEntriesProto(entries)` from the list it already reads. Regenerate with `./scripts/generate-proto.sh`. `GoStore.setMetadata` and `GoStore.clearMetadata` map `resp.metadata` with `Self.mapMetadataEntry`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Section header, intro | `PVSectionHeader`, `Text` | `SourcePageMetadataView` | Add calls `model.metadataDialogs.openAdd(on: model.metadata)` |
| Board rows | `PVEmptyState`, `PVReorderableList`, `PVInlineEdit`, `PVTextArea`, `PVCard(border: .dashed)`, `PVInput`, `PVButton`, `PVIconButton` | `MetadataBoardView` *(new)* | Moved code; same modifiers and identifiers |
| Add dialog body | `PVField`, `PVComboBox`, `PVInput` | `MetadataBoardAddForm` *(new)* | Hosted by `SourcePageView`'s `pvFormDialog`, reading `model.metadataDialogs` |
| Clear confirm | `pvConfirm(item:)` | `SourcePageView` | Reads `model.metadataDialogs.pendingDelete` |

The board is a feature snowflake (S15-D1 inventory: Snowflake) in `Features/Sources/`; no kit component changes. States as today. Accessibility: identifiers and labels unchanged.

### Strings

None — no key added, changed or removed (D-11).

## 7. Steps

**STEP-1 — Clear answers the owner's rows**
Files: `api/proto/engine.proto`, `api/ffi/handlers/metadata.go`, `api/ffi/handlers/metadata_test.go`, generated `engine.pb.go` / `engine.pb.swift`. Skill: `add-ffi-handler`.
Do: D-1 and §6 FFI; update `TestMetadataOwners`' Set assertion from `GetEntry()` to the list, and extend it (T-1).
Check: `./scripts/generate-proto.sh` leaves no diff after commit; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-2 — Store methods take an owner**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift` (`ThrowingStore`), `macos/ProvenenciaTests/SourcePageModelTests.swift`, `macos/App/Features/Sources/SourceMetadataSection.swift` (call sites only). Skill: `add-swift-test`.
Do: D-2, D-3, D-4. `SourceMetadataSection` calls the new methods with `MetadataOwner(level: .source, id: context.sourceID)` (temporary; STEP-4 replaces the class): `replaceEntry` and `applyCleared` take the returned list and set `context.workspace?.metadata` to it in place of their index-or-append and empty-or-remove lines, and keep their other lines (`drafts`, edit mode, `applyMetadataWrite`, `syncDrafts`) as they are. Rewrite each doc comment on the four methods and the two fake error properties to say owner where it says Source ("Permanently dismiss an unfilled type suggestion for this owner."), and the `replaceEntry` comment to name `setMetadata`. Rename the fake error properties in the tests. Add T-10.
Check: Xcode tests pass with no Swift warning; `rg -n "\b(set|clear|reorder)SourceMetadata\b|dismissSourceMetadataSuggestion|SourceMetadataError" macos --glob '!*.pb.swift'` prints nothing (`.mutatedSourceMetadata` does not match; STEP-5 removes it).

**STEP-3 — The host protocol and the Source page host**
Files: `macos/App/Features/Sources/MetadataBoardHost.swift` *(new)*.
Do: D-5 and D-6, a one-line doc comment per requirement; `failureMessage`'s says nil means the owner is gone and no state is written.
Check: the app builds.

**STEP-4 — The board model takes a host; the page keeps one dialog**
Files: `git mv` of `macos/App/Features/Sources/SourceMetadataSection.swift` to `MetadataBoardModel.swift` *(new)* in the same folder; `macos/App/Features/Sources/MetadataBoardDialogs.swift` *(new)*; `macos/App/Features/Sources/SourcePageModel.swift`; `macos/App/Features/Sources/SourcePageView.swift`; `macos/App/Features/Sources/SourcePageMetadataView.swift` (dialog calls only); `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Do: D-7 and D-9; `SourcePageModel.metadata = MetadataBoardModel(host: SourcePageMetadataHost.source(context: context))`; move the tests' dialog calls per D-12.
Check: Xcode tests pass; `rg -n "context|applyCleared|replaceEntry" macos/App/Features/Sources/MetadataBoardModel.swift` prints nothing; `git diff 7d2f3e7 -- macos/ProvenenciaTests/SourcePageModelTests.swift` shows only D-12's edits and T-10.

**STEP-5 — One mutation for a board write**
Files: `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/App/Features/Sources/SourcePageContext.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `.agents/skills/add-workspace-place/SKILL.md`, `docs/macos-client-patterns.md`. Skill: `add-workspace-place`.
Do: D-8.
Check: Xcode tests pass; `rg -n "mutatedSourceMetadata|applyMetadataWrite" macos .agents .cursor docs/*.md` prints nothing.

**STEP-6 — The board view takes the model**
Files: `macos/App/Features/Sources/MetadataBoardView.swift` *(new)*, `macos/App/Features/Sources/SourcePageMetadataView.swift`, `macos/App/Features/Sources/SourcePageView.swift`. Skill: `add-ui-component`.
Do: D-10; the text editor's comment that names `SourceMetadataSection.drafts` names `MetadataBoardModel.drafts`.
Check: the app builds; `rg -n '"sources\.page\.(metadata|addMetadata)' macos/App/Features/Sources/MetadataBoardView.swift` prints nothing; the identifier suffixes match today's: `git show 7d2f3e7:macos/App/Features/Sources/SourcePageMetadataView.swift | rg -o 'sources\.page\.(metadata|addMetadata)[^"]*' | sort -u` is compared by eye with the identifiers the new `SourcePageMetadataView.swift` spells plus those `MetadataBoardView.swift` builds from its two prefixes (`PVInlineEdit`'s prefix stays `"\(accessibilityIdentifierPrefix).\(fieldID)"`); this half of the check, and seeing them in the running app (§10 step 3), is manual.

**STEP-7 — Board and dialog tests on a test host**
Files: `macos/ProvenenciaTests/MetadataBoardModelTests.swift` *(new)*. Skill: `add-swift-test`.
Do: T-2 – T-8 with `TestMetadataBoardHost` (D-12).
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-8 — Version and docs**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION`), `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row status), this plan (status, Landed note).
Do: the header's bump; the Definition of done's docs lines.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-26-metadata-board-takes-an-owner.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence ending in a period, like the rest of `git log` ("Let Clear answer with the owner's metadata rows."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `api/ffi/handlers/metadata_test.go` `TestMetadataOwners` (extended; artifact and graph sub-tests) | Set `m.suggested` "FamilySearch" → `set.GetMetadata()` lists it with that value, `suggested` true; set `m.extra` "14" → its list has `m.extra` after `m.suggested`; capture the response of the existing first clear (`m.suggested`, today discarded into `&engine.ClearMetadataResponse{}`) → `GetMetadata()` lists it with no value, `suggested` true; clear `m.extra` → not listed | AC-3.1, AC-3.2, AC-3.3, D-1 |
| T-2 | Swift | `macos/ProvenenciaTests/MetadataBoardModelTests.swift` `saveFillsSuggestionThenReportsWrite` *(new)* | Rows `[Author suggested, empty]`; `drafts["f-author"] = "Mary Robins"`; `set` scripted to return `[Author saved "Mary Robins", suggested]`; `save` → calls `[clearFailure, set(f-author, "Mary Robins"), didWrite]` in that order; `host.entries` is the returned list; `drafts["f-author"] == "Mary Robins"`; `board.suggested` empty | AC-2.1, D-5, D-7 |
| T-3 | Swift | same · `reorderFailureRevertsAndShowsFailure` *(new)* | Saved `[A, B]`, reorder throws, `failureMessage` "Boom" → `moveSaved` leaves `host.entries` `[A, B]`; `showFailure("Boom")` once; no `didWrite` | AC-2.2 |
| T-4 | Swift | same · `goneOwnerWritesNoState` *(new)* | `failureMessage` nil: `set` throws `sourcemetadata.invalid` → `fieldError` nil, no `showFailure`, `entries` unchanged; `clear` throws → returns false, no `showFailure`; reorder throws → `entries` back to the earlier order, no `showFailure` | AC-2.3, AC-1.8 |
| T-5 | Swift | same · `clearTakesTheHostsRows` *(new)* | Rows `[Author saved, suggested]`; `clear` returns `[Author suggested, empty]` → `host.entries` that list, `drafts["f-author"] == ""`, `didWrite` once, returns true | AC-2.4, D-7 |
| T-6 | Swift | same · `addDialogWritesToItsTargetBoard` *(new)* | Boards A and B on two hosts; `dialogs.openAdd(on: B)`, field `f-author`, value "Ancestry", `createFromAdd` → B's host got `set`, A's got none; `dialogs.isAdding` false | AC-2.5, D-9 |
| T-7 | Swift | same · `clearConfirmTargetsItsBoard` *(new)* | A has saved Author; `dialogs.askClear(on: A, fieldID: "f-author")`, `confirmClear` → A's host got `clear`; `pendingDelete` nil; on a throwing `clear`, `pendingDelete` stays | AC-2.6, D-9 |
| T-8 | Swift | same · `invalidValueStaysOnTheFieldOrDialog` *(new)* | `set` throws `sourcemetadata.invalid`, `failureMessage` "Bad" → editing a saved row: `fieldError == "Bad"`, `fieldErrorID` the row; from the dialog: `dialogs.addValueError == "Bad"`, dialog open; no `showFailure` either way | AC-1.3, AC-1.9 |
| T-9 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `loadMetadataSuggestionsAndValues` (existing; so are T-9a – T-9f, edited only per D-12) | Saved and suggested rows split | AC-1.1 |
| T-9a | Swift | same · `saveMetadataValueFillsSuggestion` (existing) | Suggestion filled → saved | AC-1.2 |
| T-9b | Swift | same · `invalidMetadataSetLandsOnFieldError` (existing) | Invalid value → row error | AC-1.3 |
| T-9c | Swift | same · `metadataSetIOFailureLandsOnPageError` (existing) | I/O failure → banner | AC-1.4 |
| T-9d | Swift | same · `reorderSavedMetadataFailureRevertsAndSetsPageError` (existing) | Failed reorder → revert and banner | AC-1.5 |
| T-9e | Swift | same · `clearSuggestedMetadataReturnsToDashedList` (existing) | Cleared suggestion → dashed row | AC-1.6 |
| T-9f | Swift | same · `addMetadataRequiresFieldAndValue` (existing) | No field → field error, no write | AC-1.7 |
| T-9g | Swift | same · `invalidMetadataAddLandsOnValueField` (existing) | Invalid add → dialog value error | AC-1.9 |
| T-10 | Swift | same · `metadataWriteOnDeletedSourceLeavesThePage` *(new)* | Saved Repository; `store.setMetadataError = FakeStore.sourceNotFound`; edit and save → `model.gone != nil`, `model.pageError == nil`, `model.metadata.fieldError == nil`, the row unchanged (the pattern of the `FakeStore.sourceNotFound` tests near line 2256) | AC-1.8 |
| T-11 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` (existing, edited per D-8) | `.mutatedSourceWorkspace(sourceId: "s1")` still invalidates `sourceWorkspace(s1)` | D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests`, `SourceCopyModelTests`, `MetadataModelTests` and `CatalogQueryRegistryTests` (they seed or read `metadataBySource`, which stays as a computed view).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected: D-11)
- [ ] Sweep: `SourceMetadataSection`, `SourceMetadataDeleteItem`, `applyCleared`, `SourcePageMetadataView.addForm`, `setSourceMetadata`, `clearSourceMetadata`, `dismissSourceMetadataSuggestion`, `reorderSourceMetadata`, `setSourceMetadataError`, `reorderSourceMetadataError`, `mutatedSourceMetadata`, `applyMetadataWrite` `rg`'d across `macos/`, `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches today's Source page (no brief: the manual check compares before and after); no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated — not gated, nothing to move
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/macos-client-patterns.md (D-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Before the change (on `main`), open a project with a Census Source that has two saved metadata values, one URL value, and one unfilled type suggestion. Screenshot the page at Regular and at Wide (pane wider than 1600pt).
2. On the branch, open the same Source. The Metadata section matches both screenshots: header with count and Add, intro, saved rows, "Suggested by this type" with the dashed row.
3. Fill the suggestion and press Return; edit a saved value and save; open the URL value's link; drag a row and reopen the page (the order holds); clear a value that was a suggestion (it returns as a dashed row); clear a value that was not (it goes); dismiss a suggestion; Add metadata with no field (the field error shows), then add a field and value. With Accessibility Inspector, `sources.page.addMetadata`, `sources.page.metadata.list`, `sources.page.metadata.suggestions`, `sources.page.metadata.<fieldID>.value`, `sources.page.addMetadata.field` and `sources.page.addMetadata.value` are present.
4. Open the Metadata fields page: the field you just used shows its Used by count raised by one (it follows the effects' `metadata_field_usage`).
5. With the Add dialog open, drop a file on the page: nothing is added (the drop check reads `metadataDialogs`).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `@Bindable` bindings (`$dialogs.addFieldID`) do not compile through `SourcePageModel.metadataDialogs` declared `let` | D-9 allows `var`, as the section properties' comment explains |
| `didWrite` before rows are patched changes refresh order | D-5/D-7 fix the order; T-2 pins it |
| A moved identifier changes | STEP-6's check and §10 step 3 |
| `MetadataBoardDeleteItem` holding a class breaks `pvConfirm`'s `Equatable` use | D-9's `==` over `ObjectIdentifier` |

**Stop and report if:**

- `SourceMetadataSection`, `SourcePageMetadataView` or the Clear handler on `main` no longer match §4.
- A `SourcePageModelTests` metadata test needs an edit beyond D-12 to pass — that is a behavior change.
- `.mutatedSourceMetadata` turns out to reach a key other than `sourceWorkspace` (D-8's premise).
- The Swift compiler rejects `any MetadataBoardHost` with `entries { get set }` as a stored property of an `@Observable` class in a way that needs the protocol shape changed (D-5).
- A conformer of `GenealogyStore` other than `GoStore`, `FakeStore` and `ThrowingStore` exists.

## 12. Open questions

None. (M-2 was answered by measurement: §6 Scale.)

## 13. PR

**Title:** S15-26 — The metadata board takes an owner
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-26-metadata-board-takes-an-owner.md`.
