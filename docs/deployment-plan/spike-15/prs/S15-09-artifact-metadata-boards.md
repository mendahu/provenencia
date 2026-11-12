# S15-09 — Artifacts show their metadata on the Source page

**Status:** Draft
**Spike / slice:** Spike 15 · Slice 2 — Metadata at three levels
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S15-D1](../design/S15-D1-source-page.md) SD-9, SD-10, board drawn and validated 2026-11-10
**Board:** [Claude Design — Source Detail](https://claude.ai/artifact/443NPNYmGErAmxgV1fBsFK), page *Source page*, row "Artifact cards", frame "Copy details" (`art-details`, the board's frames 9a / 9b: twelve Artifacts, ART-0615's Copy details open), and the decisions panel entry "S15-D1 · SD-9 – SD-14". The implementer reads it with the Artifact tool's `read` on that link, `path` "project/Source.dc.html" (the card at its `metaShow` / `metaBoard` lines, `artMetaBoard(a)`, and the panel) and "project/Source page.dc.html" (frame list). §6 macOS describes the frame in words
**Depends on:** S15-08 (landed [#791](https://github.com/mendahu/provenencia/pull/791)), S15-26 (landed [#798](https://github.com/mendahu/provenencia/pull/798)), S15-D1 (board validated 2026-11-10)
**Blocks:** — (S15-10 and S15-11 reuse the board without this PR; S15-25 adds Move to the same cards)
**Branch:** `s15-09-artifact-metadata-boards`
**Version bump:** minor, 0.10.3 → **0.11.0** (or the next minor after `main`'s `VERSION` when you rebase): `VERSION`, `core/version.go`, and every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj` — a backward-compatible feature on a screen (each Artifact card gains its metadata board); no wire or catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (a descriptor patch and its Patch row), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (the board stays a feature snowflake), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md) (STEP-1 pins a scope path)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-audit.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "One metadata board", "Metadata lives at three levels with one vocabulary" and "Suggestions have a level" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); the map's S15-09 scope entry and its two **Also** lines; S15-26's [plan](S15-26-metadata-board-takes-an-owner.md) (D-5, D-6, D-10, **Next adds**, its Out rows and **Landed** note) and [ship log](S15-26-metadata-board-takes-an-owner.ship.md) MNT-1; S15-08's [plan](S15-08-artifacts-and-graphs-hold-metadata.md) and [ship log](S15-08-artifacts-and-graphs-hold-metadata.ship.md) CON-2 / F-8; [S15-D1](../design/S15-D1-source-page.md) SD-9, SD-10, SD-W2 and its inventory rows "Metadata board", "Artifact card", "Copy details disclosure"; [`docs/macos-client-patterns.md`](../../../macos-client-patterns.md) § mutations
**Planned:** 2026-11-12 against `main` @ `83895e4`
**Review:** — (not reviewed yet)
**Runs on:** macOS. Every Swift check and `scripts/check-pr.sh` need Xcode ([`README.md`](../../../../README.md)). STEP-1 runs anywhere Go and cgo SQLite build. No proto change: do not run `./scripts/generate-proto.sh`

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

Each Artifact card on the Source page gets its own metadata board: the copy details of that Artifact (provider, URL, accessed date on a census), with the same saved rows, type suggestions, edit, clear, reorder and Add as the Source's board. The board sits behind a closed **Copy details (n)** disclosure under the card's description, as S15-D1 draws it, because a Source can hold 60 Artifacts. It is the S15-26 board with the Artifact as owner; the engine has stored and served these values since S15-08.

A board write stops reloading the whole page: the page's cached workspace takes the owner's new rows in place (S15-26 review MNT-1, required by the user 2026-11-12). Today every save reloads `GetSourceWorkspace` with every Artifact's board, and that reload can end a row edit the researcher began after the save. The PR also adds the test S15-08's review left open (CON-2): a released Artifact or graph value resolves, on its own, to its Source's scope through its deleted owner.

**Why now.** Bar 3's Artifact half ("the Artifact holds provider, URL and accessed date"). S15-08 put the values on the wire and S15-26 made the board take any owner; nothing draws them yet.

**What stays.** The Source's own Metadata section looks and behaves as today (same header, intro, rows, identifiers and copy). The Artifact card's preview, meta row, edit mode, no-graph row and delete are unchanged. The Add metadata dialog and the clear confirm stay the page's one of each (`MetadataBoardDialogs`). No Go behavior changes.

## 2. User stories

**US-1.** As a researcher on a Source page, I open an Artifact's copy details and see and edit its own metadata, so that where I got this copy is recorded on the copy, not on the Source.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A census Source with Artifacts ART-0615 (saved Provider "Ancestry", URL, Accessed "12 March 2025") and ART-0611 (nothing saved) | ART-0615's card shows a closed disclosure "Copy details (3)" under its description; ART-0611's shows "Copy details". No board is drawn until one is opened |
| AC-1.2 | Open ART-0615's Copy details | Its board shows the three saved rows in their order, then "Suggested by this type" with the type's unfilled Artifact suggestions, then an "Add field" button. Only that card grows |
| AC-1.3 | In ART-0615's board, fill the suggestion "Image" with "0473" and press Return | Image moves to the saved rows; the title reads "Copy details (4)"; the Source's own board and ART-0611's board are unchanged |
| AC-1.4 | Edit ART-0615's URL value, clear Provider (confirm), drag Accessed above URL, dismiss a suggestion | Each is written to ART-0615 only and shows at once, as on the Source's board |
| AC-1.5 | Press "Add field" on ART-0615's board, pick a field, type a value, Add | The page's Add metadata dialog opens; the value saves on ART-0615 |
| AC-1.6 | A value the engine refuses (`sourcemetadata.invalid`, a URL `file:/tmp`) on an Artifact's row | The error shows under that row on that card; the row stays in edit |
| AC-1.7 | An Artifact whose type suggests nothing and that has no value | Opened, its board shows only "Add field" (no empty-state block) |
| AC-1.8 | Put ART-0615's card into edit (pencil) | The disclosure and board hide while the card is in edit; after Save or Cancel the disclosure shows again, still open if it was |
| AC-1.9 | Open ART-0615's Copy details, resize the window across 1600pt | The disclosure stays open (SD-W2) |
| AC-1.10 | Delete ART-0615 from its card | The card, its board and its open state go; other cards keep theirs |

**US-2.** As a researcher, a metadata save does not reload the page, so that a row I start editing right after a save stays in edit.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Save a value on any board (Source or Artifact) | `GetSourceWorkspace` is not called again; the page's cached workspace holds the owner's new rows |
| AC-2.2 | Save row A on the Source's board, then begin editing row B before the page re-syncs | Row B stays in edit after the page applies the cached workspace |
| AC-2.3 | Begin editing a row on ART-0611's board, then save a value on the Source's board | ART-0611's row stays in edit |
| AC-2.4 | A note save (any other page write) after a board write | The page reloads as today; a board whose rows did not change keeps its edit and drafts |

**US-3.** As the engine, a value released by an Artifact or graph delete names its Source's scope on its own, through its deleted owner, so that the scope does not ride only on the owner's own delete.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The changes of a real `artifacts.Delete` of an Artifact with one value (its value change, its layout change, its link deletes and the Artifact's delete) | Each `artifact_metadata` and `artifact_metadata_layout` change, resolved alone with the whole revision as ghosts, scopes the Source; resolved without the Artifact's delete among the ghosts it scopes nothing |
| AC-3.2 | The same for a real `evidencegraphs.Delete` of a graph with one value | Each `evidence_graph_metadata` and `evidence_graph_metadata_layout` change scopes the Source alone; without the graph's delete it scopes nothing |

## 3. Scope

| In | Covered by |
| --- | --- |
| A test hook resolving one change's Source scope inside a revision, and a test over the real `artifacts.Delete` and `evidencegraphs.Delete` changes (map **Also**, S15-08 CON-2) | STEP-1, AC-3.1, AC-3.2 |
| `CatalogSourceWorkspace.artifactMetadata` *(new)*, mapped from `GetSourceWorkspaceResponse.artifact_metadata`; one owner→rows accessor on the workspace | STEP-2, AC-1.1 |
| `SourcePageMetadataHost` reads and writes its rows by owner through that accessor | STEP-3, AC-1.3, AC-1.4 |
| `CatalogMutation.wroteMetadataBoard` *(new)* and `CatalogQueryPatch.metadataBoard()` *(new)* on `.sourceWorkspace`; the host's `didWrite` applies it in place of `.mutatedSourceWorkspace` (map **Also**, S15-26 MNT-1) | STEP-4, AC-2.1 |
| The page resets a board only when its rows differ from what the board shows; one board model per Artifact, created, kept and dropped with the workspace; open state per Artifact | STEP-5, AC-1.8 – AC-1.10, AC-2.2 – AC-2.4 |
| `MetadataBoardView` takes a style (page or card); the board's chrome moves out of `SourcePageChrome.swift` into the board's file under board names | STEP-6, AC-1.2, AC-1.7 |
| Each Artifact card draws "Copy details (n)" and the board (SD-9, SD-10) | STEP-7, AC-1.1, AC-1.2, AC-1.5, AC-1.6 |
| Swift tests for the boards, the patch and the reset rule | STEP-8 |
| Guidance, version and plan docs | STEP-9 |

| Out | Home |
| --- | --- |
| The add flow's copy fields (a draft host) | S15-10 |
| The Evidence graph form's board; `getMetadataBoard` in Swift | S15-11 |
| The Move menu item on the card (SD-14) | S15-25 |
| The citation preview (SD-11 – SD-13) | S15-23 |
| A metadata write on an Artifact or Source deleted in another window says the owner is gone (today: `sourcemetadata.invalid`, shown as a field or dialog error) | not planned (S15-03 left not-found handling for metadata writes out) |
| Renaming the `L10n.Sources.metadata*` keys and their `.xcstrings` ids to board names | not planned (keys are stable string ids; `Sources` is the feature's namespace) |
| Moving the cached workspace's `source.updated` after a board write | not planned (nothing on the Source page reads it; the Sources list row follows the effects' `sourceScopes`) |
| Naming the target Artifact in the Add metadata dialog | not planned (the board draws the same dialog "as on the Source") |
| `FakeStore.reorderMetadata` still renumbers rows itself (S15-26 round-2 note) | not planned in this PR (no test here needs another reorder answer) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Workspace read (Go) | `api/ffi/handlers/sources.go` `GetSourceWorkspace`; `api/ffi/handlers/metadata.go` `artifactBoardsProto` | Answers `artifact_metadata`: one `MetadataBoard { owner_id, entries }` per Artifact, in the order of `artifacts`, `owner_id` from `database.UUIDString` (the same formatter as `Artifact.id`) |
| Workspace type | `macos/App/Platform/GenealogyStore.swift` `CatalogSourceWorkspace` (around line 1260) | `source`, `notes`, `metadata`, `artifacts`, `credibility`; no Artifact boards. Ten `CatalogSourceWorkspace(` initializer calls in `macos/` |
| Go store | `macos/App/Platform/GoStore.swift` `getSourceWorkspace` (around line 137) | Maps everything but `resp.artifactMetadata`; `Self.mapMetadataEntry` maps one entry |
| Fake store | `macos/App/Platform/FakeStore.swift` `getSourceWorkspace` (around line 468), `deleteSource` (around line 711), `deleteArtifact` (around line 848), `metadataByOwner`, `metadataWriteResults`, `scriptedMetadataWrite` | Workspace reads `.source` owner rows only; `deleteSource` clears the Source owner's rows; `deleteArtifact` does not touch `metadataByOwner`; writes return scripted rows per owner |
| Host | `macos/App/Features/Sources/MetadataBoardHost.swift` `MetadataBoardHost`, `SourcePageMetadataHost` | `SourcePageMetadataHost(context:owner:rows:setRows:)` reads and writes rows through two closures; `.source(context:)` passes `{ $0.metadata }` / `{ $0.metadata = $1 }`. `didWrite` → `context.applyWrite(effects)` → `.mutatedSourceWorkspace(sourceId:)`, which reloads the whole workspace |
| Board model | `macos/App/Features/Sources/MetadataBoardModel.swift` `MetadataBoardModel` | Patches `host.entries` then calls `host.didWrite(effects)` on set, clear, dismiss and reorder; `resetDrafts()` exits edit and clears field errors |
| Board dialogs | `macos/App/Features/Sources/MetadataBoardDialogs.swift` `MetadataBoardDialogs` | One Add dialog and one clear confirm, each targeting a `MetadataBoardModel` (`openAdd(on:)`, `askClear(on:fieldID:)`) |
| Board view | `macos/App/Features/Sources/MetadataBoardView.swift` `MetadataBoardView`, `MetadataBoardAddForm`, private `SourcePageMetadataSuggestionRow`, `SourcePageMetadataTextEditor` | Draws the empty state (with a primary Add), the reorderable saved rows, the suggestions; label column `SourcePageLayout.metadataLabelWidth` (132) |
| Board chrome | `macos/App/Features/Sources/SourcePageChrome.swift` `SourcePageLayout.metadataLabelWidth`, `.reorderHandleWidth`, `.metadataColumnSpacing`, `.metadataSuggestionLabelInset`, `SourcePageMetadataValueChrome`, `sourcePageMetadataValueChrome()`, `SourcePageMetadataLabel` | Used only by `MetadataBoardView.swift` (`rg` at the base). `SourcePageLayout.notesBylineWidth` and `SourcePageSectionTopPadding` are used by other Source page views |
| Source section | `macos/App/Features/Sources/SourcePageMetadataView.swift` `SourcePageMetadataView` | Header, intro, `MetadataBoardView(board: model.metadata, accessibilityIdentifierPrefix: "sources.page.metadata", …)` |
| Page model | `macos/App/Features/Sources/SourcePageModel.swift` `apply(_:)`, `confirmResourceDelete(leaving:)` case `"artifact"` | `apply` resets the Source board when `appliedWorkspace?.metadata != workspace.metadata`; Artifact delete removes the Artifact from `workspace.artifacts`, `apply`s, `setQueryValue`s the key |
| Page context | `macos/App/Features/Sources/SourcePageContext.swift` `applyWrite`, `workspace` | `applyWrite` → `.mutatedSourceWorkspace(sourceId:)` (notes, credibility, Artifacts and, today, the board) |
| Artifacts section | `macos/App/Features/Sources/SourceArtifactsSection.swift` `SourceArtifactsSection` (`items`, `seedDrafts`, `editingArtifactID`, `labels`, `descriptions`) | Per-card state keyed by Artifact id; no boards |
| Artifact card | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `artifactCard`, `restingText`, `noGraphRow` | `LazyVGrid` of cards (min 264pt); resting text is title, description, then the no-graph row; edit fields replace it |
| Page sync | `macos/App/Features/Sources/SourcePageView.swift` `SourcePageContent` `.onChange(of: workspaceHandle.value)` | Calls `model.sync(from:)` → `apply` on every handle value change |
| Mutations | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogMutation`, `invalidationKind`, `patchedSource` | `.updatedSource` is the patch-only precedent (`invalidationKind` nil) |
| Patches | `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` `CatalogQueryPatch.updatedSourceShell()` | Reads the mutation, guards the key's Source id, `session.patchCached(key, as: CatalogSourceWorkspace.self)` |
| Registry | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `.sourceWorkspace` (around line 230) | `patches: [.updatedSourceShell()]`; no `refreshes` |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `apply(_:_:)`, `applyPatches`, `patchCached` | Patches run before invalidation; a patch-only mutation invalidates nothing; `patchCached` keeps the stale flag and restarts an in-flight load |
| Effects scope | `core/database/effects/resolve.go` `Resolve`, `ghostsFrom`, `appendPath`; `core/database/effects/path.go` `walk` | `Resolve` walks every change with ghosts from the whole batch; no exported way to resolve one change alone |
| Scope tests | `core/database/effects/path_test.go` `TestScopeThroughDeletedParent` (stub tables and hand-built changes); `core/database/artifacts/artifacts_test.go` (around line 625) and `core/database/evidencegraphs/evidencegraphs_test.go` (around line 1455) | The real deletes resolve as a whole batch, where the owner's own delete already scopes the Source (S15-08 ship log, CON-2 "Not resolved") |
| Fixture | `core/database/effects/reads_resolve_test.go` `readsFixture` (package `effects_test`) | A catalog, a Source, its Default graph, a Jones graph, an Artifact linked to both, and an open write transaction rolled back at cleanup |
| Guidance | `.agents/skills/add-workspace-place/SKILL.md` lines 62 and 172; `docs/macos-client-patterns.md` line 89 | Line 172 says a board's host applies `context.applyWrite(effects)`; line 62 names `.updatedSourceShell` as the page-shell patch; line 89 lists the mutations rule 2 keeps |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` metadata tests, `makeStore`, `makeModel`; `macos/ProvenenciaTests/MetadataBoardModelTests.swift`; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `applyUpdatedSourcePatchesListAndWorkspace`, the `everyMutation` list (around line 934) and the `mutations` list (around line 1392) | Drive the Source board through `model.metadata`; no Artifact board |

**Precedent.** S15-26 (#798) for the board, the host, the dialogs and `MetadataBoardModelTests`' test host; copy its identifier-prefix pattern. `CatalogQueryPatch.updatedSourceShell()` with `.updatedSource` for a patch-only mutation on the page key; copy its guard and `patchCached` shape. `TestScopeThroughDeletedParent` for resolving one change with ghosts; `readsFixture` for a real catalog in `effects_test`. `ReconciledValueRow` (`macos/App/Features/Conclusions/ReconciledValueRow.swift`) for a `PVDisclosureButton` on a feature row.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `core/database/effects/export_test.go` *(new)* (`package effects`) adds `func SourceScopeOfChange(tx *sql.Tx, changes []rowchange.Change, i int) ([][]byte, error)`: `table := entityTable[changes[i].EntityType]`, `w := &walk{q: tx, ch: changes[i], ghosts: ghostsFrom(changes), table: table}`, `return appendPath(w, table, registry[table].SourceScope, nil)`; an unknown entity returns `errf("effects: no entry for %s", …)`. The test is `core/database/effects/released_metadata_scope_test.go` *(new)* (`package effects_test`), on `readsFixture` | `artifacts` and `evidencegraphs` import `effects` (`go list -deps`), so only the external test package can run their real deletes, and an `export_test.go` hook is visible to it and to no production code. Same walk as `TestScopeThroughDeletedParent` | An exported production function — API for a test; the test in `artifacts_test` — it cannot see an unexported walk |
| D-2 | `CatalogSourceWorkspace` gains `var artifactMetadata: [String: [CatalogMetadataEntry]] = [:]` *(new)*, keyed by Artifact id, and two members *(new)* in the same file: `func metadataRows(for owner: MetadataOwner) -> [CatalogMetadataEntry]?` (`.source` with `owner.id == source.id` → `metadata`; `.artifact` with an Artifact of that id in `artifacts` → `artifactMetadata[owner.id] ?? []`; anything else → nil) and `mutating func setMetadataRows(_ rows: [CatalogMetadataEntry], for owner: MetadataOwner) -> Bool` (the same cases; writes and returns true, or returns false and writes nothing). The default keeps every initializer call compiling | One place says where an owner's rows live in the workspace, read by the host (D-4) and the patch (D-5). A dictionary because cards read by id | An array parallel to `artifacts` — order coupling; rows on `CatalogArtifact` — the Artifact type is shared with lists that carry no board |
| D-3 | `GoStore.getSourceWorkspace` sets `artifactMetadata: Dictionary(resp.artifactMetadata.map { ($0.ownerID, $0.entries.map(Self.mapMetadataEntry)) }, uniquingKeysWith: { _, last in last })`. `FakeStore.getSourceWorkspace` sets `artifactMetadata` to `metadataByOwner[MetadataOwner(level: .artifact, id: a.id)] ?? []` for each Artifact of the Source. `FakeStore.deleteArtifact` sets `metadataByOwner[MetadataOwner(level: .artifact, id: artifactID)] = nil`, as `deleteSource` does for the Source owner | A duplicate owner id on the wire never traps. The fake keeps state, not rules (`swift-tests.mdc`); releasing an owner's rows with the owner is bookkeeping the Source delete already does | `Dictionary(uniqueKeysWithValues:)` — traps on a duplicate |
| D-4 | `SourcePageMetadataHost` becomes `init(context: SourcePageContext, owner: MetadataOwner)`; `entries` get is `context.workspace?.metadataRows(for: owner) ?? []`; set is `context.workspace?.setMetadataRows(newValue, for: owner)` (result discarded). The `rows` / `setRows` closures go; `static func source(context:)` stays as `SourcePageMetadataHost(context: context, owner: MetadataOwner(level: .source, id: context.sourceID))`. An Artifact's host is `SourcePageMetadataHost(context: context, owner: MetadataOwner(level: .artifact, id: artifactID))` | D-2 owns the owner→rows mapping; the closures would be a second copy (S15-26 nit MNT-6) | Closures per Artifact — two copies of D-2 |
| D-5 | `CatalogMutation.wroteMetadataBoard(sourceId: String, owner: MetadataOwner, entries: [CatalogMetadataEntry])` *(new)*: patch-only (`invalidationKind` nil, `patchedSource` nil). `CatalogQueryPatch.metadataBoard()` *(new)* in `CatalogLastUpdated.swift`, beside `updatedSourceShell()`: `follows: nil`; `guard case .wroteMetadataBoard(let sourceId, let owner, let entries)? = mutation` (the patch receives `CatalogMutation?`), `guard case .sourceWorkspace(_, let keySource) = key, keySource == sourceId` (else `.unchanged`), then `session.patchCached(key, as: CatalogSourceWorkspace.self) { $0.setMetadataRows(entries, for: owner) }`. `.sourceWorkspace`'s `patches` becomes `[.updatedSourceShell(), .metadataBoard()]` | The write already describes what changed for the page (the owner's whole board, S15-26 D-1); a reload of 60 Artifacts' boards adds nothing (map **Also**; user 2026-11-12). Patch-only like `.updatedSource`, so the effects still move every other key (Sources row, field list `usedBy`) | Keeping `.mutatedSourceWorkspace` beside the patch — the reload MNT-1 removes; a patch keyed by effects ids — the effects carry no rows |
| D-6 | `SourcePageMetadataHost.didWrite(_ effects:)` becomes `context.session.apply(effects, .wroteMetadataBoard(sourceId: context.sourceID, owner: owner, entries: entries))`. `MetadataBoardModel` is unchanged: it sets `host.entries` before `didWrite` (S15-26 D-5), so `entries` is the engine's answer | One line; every board write (set, clear, dismiss, reorder) takes the patch | A new `didWrite(entries:effects:)` protocol requirement — the host already holds them |
| D-7 | `SourcePageModel.apply(_:)` reads `let shown = context.workspace` before it assigns the new value, and resets the Source board when `shown?.metadata != workspace.metadata` (in place of `previous?.metadata`). It then calls `artifacts.syncBoards(shown: shown, to: workspace)` *(new)*. Other sections keep comparing with `appliedWorkspace` | The board writes its rows into `context.workspace` itself, so "rows differ from what the board shows" is the test for a change made elsewhere; a patch or reload carrying the rows the board already shows keeps any edit begun meanwhile (AC-2.2 – AC-2.4) | Comparing with `appliedWorkspace` — resets after every own write; updating `appliedWorkspace` from the host — a second writer of the page's applied state |
| D-8 | `SourceArtifactsSection` owns the Artifact boards: `private(set) var boards: [String: MetadataBoardModel] = [:]` *(new)*, `var openBoardIDs: Set<String> = []` *(new)*, `func board(for id: String) -> MetadataBoardModel?` *(new)*, `func syncBoards(shown: CatalogSourceWorkspace?, to workspace: CatalogSourceWorkspace)` *(new)*: for each Artifact in `workspace.artifacts`, a missing board is created with D-4's Artifact host and `resetDrafts()`; an existing board calls `resetDrafts()` when `shown?.artifactMetadata[id] != workspace.artifactMetadata[id]`; boards and `openBoardIDs` entries whose id is no longer an Artifact are removed. `func copyDetailsTitle(id: String) -> String` *(new)*: `board(for: id)?.saved.count ?? 0`; 0 → `L10n.string(L10n.Sources.artifactCopyDetails)`, otherwise `L10n.Sources.artifactCopyDetailsCount(n)` | The section already holds per-card state keyed by id. Boards are made in `apply` (from `sync`, never during a view body), so no state changes during a view update. Open state in the model survives the lazy grid and a resize across 1600 (SD-W2) | Boards on `SourcePageModel` — a second owner of per-card state; `@State` open flags in the card — lost when `LazyVGrid` drops a card |
| D-9 | `SourcePageModel.confirmResourceDelete` case `"artifact"` also sets `workspace.artifactMetadata[target.id] = nil` before `apply(workspace)` | The removed Artifact's rows leave with it; `syncBoards` drops its board and open flag | Leaving the entry — a stale board key in the cache |
| D-10 | `enum MetadataBoardStyle { case page, card }` *(new)* in `MetadataBoardView.swift`; `MetadataBoardView` gains `var style: MetadataBoardStyle = .page`. `.page` draws exactly what it draws today. `.card`: label column 76pt; saved rows have no horizontal padding (vertical `PVSpacing.space3` as today); a suggestion row puts its label on its own line, and under it one line holds the input, Save and dismiss (dashed card padding vertical `space3`, horizontal `space4`; no reorder-handle inset on the label); groups are spaced `space4` (heading top padding `space4` when saved rows exist); there is no `PVEmptyState`; after the suggestions (or alone, when there are none) a `PVButton(L10n.Sources.addMetadata, variant: .ghost, size: .sm, icon: .plus)` calls `onAdd`, leading-aligned, identifier `"\(accessibilityIdentifierPrefix).add"`. `MetadataBoardView` passes `style` to `MetadataBoardSuggestionRow` and passes the label width (`style` → `pageLabelWidth` / `cardLabelWidth`) to `MetadataBoardLabel`; `MetadataBoardTextEditor` is unchanged | SD-9: "the same rows, suggestions and Add as the Source's board, at the card's width"; the board frame draws a 76pt label and a trailing ghost Add, no empty-state block. At the card's minimum (264pt, inner 232pt) a one-line suggestion row leaves its input about 30pt; stacked, about 120pt. The saved row keeps its handle, so reorder works on a card as on the Source | A second card board view — a copy of the board; dropping dismiss or reorder on the card — two behaviors for one board (rule "One metadata board") |
| D-11 | The board's chrome leaves `SourcePageChrome.swift` for `MetadataBoardView.swift`, `private` there and renamed: `SourcePageLayout.metadataLabelWidth`, `.reorderHandleWidth`, `.metadataColumnSpacing`, `.metadataSuggestionLabelInset` → `private enum MetadataBoardLayout` *(new)* `pageLabelWidth` (132), `cardLabelWidth` (76), `reorderHandleWidth` (20), `columnSpacing` (`PVSpacing.space5`), `suggestionLabelInset`; `SourcePageMetadataLabel` → `private struct MetadataBoardLabel` with a `width: CGFloat` parameter; `SourcePageMetadataValueChrome` / `sourcePageMetadataValueChrome()` → `private struct MetadataBoardValueChrome` / `private extension View { func metadataBoardValueChrome() }`; `SourcePageMetadataSuggestionRow` → `MetadataBoardSuggestionRow`, `SourcePageMetadataTextEditor` → `MetadataBoardTextEditor`. `SourcePageLayout.notesBylineWidth`, `SourcePageSectionTopPadding` and `sourcePageSectionTopPadding()` stay | S15-26 left the rename to the PR where a second place draws the board (its Out row). Only the board uses them, and the label width becomes a style choice | Leaving `SourcePage*` names on a board drawn on cards — names that say the wrong owner |
| D-12 | In `SourcePageArtifactsView.restingText`, after the description and the no-graph row, when `model.artifacts.board(for: art.id)` exists: a `VStack(alignment: .leading, spacing: PVSpacing.space4)` with `.padding(.top, PVSpacing.space4)` and a `PVDivider()` above it, holding `PVDisclosureButton(model.artifacts.copyDetailsTitle(id: art.id), isExpanded: Binding(get: { model.artifacts.openBoardIDs.contains(art.id) }, set: { if $0 { model.artifacts.openBoardIDs.insert(art.id) } else { model.artifacts.openBoardIDs.remove(art.id) } }), flushLeading: true)` (identifier `"sources.page.artifact.\(art.id).copyDetails"`, `.accessibilityLabel(L10n.Sources.artifactCopyDetailsLabel(ref: art.ref))`), and, when open, `MetadataBoardView(board:, accessibilityIdentifierPrefix: "sources.page.artifact.\(art.id).metadata", onAdd: { model.metadataDialogs.openAdd(on: board) }, onClear: { model.metadataDialogs.askClear(on: board, fieldID: $0) }, style: .card)`. Nothing shows while the card is in edit (the resting text is not drawn) | The board frame: a kit Disclosure under the description with a subtle rule above, closed by default, label without a count at 0; the page's one dialog and confirm (S15-26 D-9). `restingText` is already the part edit replaces, so AC-1.8 needs no flag | The board always open — 60 boards drawn; a local chevron row — the kit has `PVDisclosureButton` (inventory: Ship) |
| D-13 | Copy: `L10n.Sources.artifactCopyDetails` = "Copy details" (`sources.page.artifactCopyDetails`), `L10n.Sources.artifactCopyDetailsCount(_ count: Int)` = "Copy details (%lld)" (`sources.page.artifactCopyDetailsCount`), `L10n.Sources.artifactCopyDetailsLabel(ref: String)` = "Copy details of %@" (`sources.page.artifactCopyDetailsLabel`, VoiceOver label, argument the ART- ref, as `editArtifactLabel(ref:)`). The board's rows, suggestions heading, Add and dialog keep their `L10n.Sources.metadata*` / `addMetadata` copy | The board's decision names "Copy details (n)"; every other string is the Source board's ("one metadata board") | New Artifact copies of the board strings — two keys for one text |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourceScopeOfChange` *(new, test-only)* | func | `Resolve` | Resolve folds the whole batch; the test needs one change's path (D-1) |
| `released_metadata_scope_test.go` *(new)* | test file | `path_test.go` `TestScopeThroughDeletedParent` | That one uses stub tables; this one the real deletes (it stays) |
| `CatalogSourceWorkspace.artifactMetadata`, `metadataRows(for:)`, `setMetadataRows(_:for:)` *(new)* | field, funcs | `CatalogSourceWorkspace.metadata` | Extends the same type (D-2) |
| `CatalogMutation.wroteMetadataBoard` *(new)* | mutation case | `.updatedSource` | A different payload on the same patch-only path (D-5) |
| `CatalogQueryPatch.metadataBoard()` *(new)* | patch | `.updatedSourceShell()` | Writes the owner's rows instead of the Source (D-5) |
| `SourceArtifactsSection.boards`, `openBoardIDs`, `board(for:)`, `syncBoards`, `copyDetailsTitle` *(new)* | state, funcs | `SourceArtifactsSection.labels` / `seedDrafts` | Same per-card, id-keyed state (D-8) |
| `MetadataBoardStyle` *(new)* | enum | `MetadataBoardView.accessibilityIdentifierPrefix` | One parameter for the card's layout differences (D-10) |
| `MetadataBoardLayout`, `MetadataBoardLabel`, `MetadataBoardValueChrome` *(renamed)* | private types | `SourcePageLayout.metadata*`, `SourcePageMetadataLabel`, `SourcePageMetadataValueChrome` | Moved and renamed (D-11) |
| `L10n.Sources.artifactCopyDetails`, `artifactCopyDetailsCount`, `artifactCopyDetailsLabel` *(new)* | keys | `L10n.Sources.metadataFieldCount` | Different text (D-13) |

**Builds on.** S15-26's board, host and dialogs; S15-08's `artifact_metadata` on the workspace read; the `.sourceWorkspace` descriptor's `patches` slot and `patchCached`; `PVDisclosureButton`; `readsFixture`.
**Deletes.** `SourcePageMetadataHost`'s `rows` / `setRows` init parameters; the board's use of `context.applyWrite` (notes, credibility and Artifacts keep it); `SourcePageLayout.metadataLabelWidth`, `.reorderHandleWidth`, `.metadataColumnSpacing`, `.metadataSuggestionLabelInset`, `SourcePageMetadataValueChrome`, `sourcePageMetadataValueChrome()`, `SourcePageMetadataLabel` (moved and renamed, D-11); the private names `SourcePageMetadataSuggestionRow`, `SourcePageMetadataTextEditor`; the Source board's reset against `appliedWorkspace` (D-7); `add-workspace-place` line 172's `context.applyWrite(effects)` sentence.
**Next adds.** S15-10's draft host conforms to `MetadataBoardHost` and draws `MetadataBoardView(style: .card)` in the add flow if its brief draws card width; S15-11 does the same in the graph form, and its live graph host on a Source page would add a `.graph` case to D-2's two members (nothing else); S15-25 adds a menu item to the same card.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A board per Artifact on the page | `MetadataBoardModel`, `MetadataBoardView`, `MetadataBoardDialogs` | Each board has its own host and owner (T-4); the dialog and confirm target the board that asked (S15-26 T-6, T-7, and T-7 here); an Artifact gone from the workspace drops its board (T-8) |
| A write's rows reaching the cache without a reload | `.sourceWorkspace` key, `SourcePageModel.apply` | Only the matching Source's key and owner change (T-2); a key whose first load is in flight falls back to a reload (`patchCached` → `.awaitingFirstValue`, as today); `apply` resets only a board whose rows differ from what it shows (T-6) |
| An Artifact id with no board in `artifactMetadata` (older engine or fake) | D-2 | Reads `[]`; a write stores the answer under that id (T-3) |
| An Artifact deleted elsewhere while its board writes | `sourcemetadata.Set` → `requireOwner` | The engine answers `sourcemetadata.invalid`; the board shows it on the row or in the dialog, as it does for the Source today (Out: not-found handling) |

### Scale

At the map's Scale (a Source with **60 Artifacts**, three values each): the workspace read and decode are unchanged (S15-08 measured `GetSourceWorkspace` at 2.3 ms in Go). The page holds 60 `MetadataBoardModel`s, made once and kept across loads; a closed disclosure draws no board, so a card's cost is one `saved.count` over its rows. A board write now makes one store call and one in-place patch of one owner's rows: no `GetSourceWorkspace`, where today every save re-reads and re-decodes all 60 boards (S15-26 §6 Scale). `apply` compares 61 owners' row arrays per sync (each a handful of entries). The effects still patch the Sources row (`sourceScopes`) and move the field list (`metadata_field_usage`) as today. No budget is set or leaned on.

### Edges

- **Assumptions.** The patch assumes the mutation's rows are the owner's whole board in display order; the host sends the rows the board just set from the engine's answer (D-6, T-2). `setMetadataRows` assumes an Artifact owner names an Artifact of this Source; any other owner, a `.graph` owner included, writes nothing and the patch reports `.unchanged` (T-3).
- **Neighbors.** A write on ART-A changes ART-A's rows only: the Source's rows and ART-B's are equal before and after (T-2, T-5). A Source-board write does not reset an Artifact board's edit, nor the reverse (T-6). A reload with unchanged rows resets no board (T-6).
- **Removed inputs.** An Artifact deleted on this page loses its board and open flag; the dialog cannot be open then (the delete confirm is the open sheet) (T-8). A reload that drops an Artifact (deleted elsewhere) drops its board the same way (`syncBoards`).
- **Real data.** None — no table-driven rule.
- **Paths that set the workspace first.** `SourcePageContext.refreshCoverFromStore` (after an Artifact create, ingest or delete) sets `context.workspace` before `sync` runs `apply`, so D-7 sees no row change there and resets no board; accepted: rows show from `host.entries` either way, a new Artifact still gets its board (a missing board is always created), and drafts are seeded by `beginEdit` and the suggestion row before a save. A missing `artifactMetadata` key against `[]` counts as a change and resets that board once (harmless: it has no edit open).
- **Untrusted size.** `artifact_metadata` comes from the engine, one board per Artifact; a duplicate `owner_id` keeps the last (D-3); a board for an id not in `artifacts` is held in the dictionary and never read (D-2 reads only listed Artifacts).

### Schema

None — no catalog change.

### Go core

None in production code. Test-only: `core/database/effects/export_test.go` *(new)* (D-1).

### FFI

None — `GetSourceWorkspaceResponse.artifact_metadata` exists since S15-08; no proto or handler change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Copy details disclosure | `PVDisclosureButton` | `SourcePageArtifactsView.restingText` | D-12; flush leading; title from `copyDetailsTitle(id:)` |
| Rule above it | `PVDivider` | same | As the board's `border-top: 1px solid var(--border-subtle)` |
| Artifact board | `MetadataBoardView(style: .card)` | `macos/App/Features/Sources/MetadataBoardView.swift` | Feature snowflake (S15-D1 inventory "Metadata board": Ship / Extend) |
| Saved rows | `PVReorderableList`, `PVReorderHandle`, `PVInlineEdit`, `PVTextArea`, `PVIconButton` | same | As on the Source board, label 76pt |
| Suggestions | `PVCard(border: .dashed)`, `PVInput`, `PVButton`, `PVIconButton` | same | Label above the controls (D-10) |
| Add | `PVButton` ghost sm, plus | same | Opens the page's Add metadata dialog |
| Add dialog, clear confirm | `pvFormDialog`, `pvConfirm(item:)` | `SourcePageView` | Unchanged; they read `model.metadataDialogs` |

**The frame in words** ("Copy details", 9a / 9b). Twelve Artifact cards, four across at Regular (1102 pane) and at Wide (the main column, 2296 pane); cards keep their size. Under ART-0615's description a thin rule, then the ghost "Copy details (3)" with a chevron, open: three rows, each a semibold caption label (Provider, URL, Accessed) in a 76pt column and the mono value beside it (the URL a link, wrapping), with the row's edit control; then the faint uppercase suggestions heading and one dashed suggestion row ("Image", an input, Save); then a ghost "+ Add". Only that card is taller; its row neighbours keep their height. Other cards show "Copy details (1)", "Copy details (2)" or "Copy details", closed. While a card is in edit (its label and description fields), no disclosure shows.

States: closed (default); open with rows; open with only Add (AC-1.7); a row error on the card (AC-1.6). Accessibility: the disclosure speaks "Copy details of ART-0615" and its expanded state (`pvExpandedState`); identifiers `sources.page.artifact.<id>.copyDetails`, `sources.page.artifact.<id>.metadata.list`, `.metadata.suggestions`, `.metadata.<fieldID>.value`, `.metadata.add`. Wide: the cards stay in the main column; nothing moves (SD-W1); open state survives the resize (SD-W2, D-8).

### Strings

L10n keys *(new)*: `L10n.Sources.artifactCopyDetails` = "Copy details"; `L10n.Sources.artifactCopyDetailsCount(_:)` = "Copy details (%lld)"; `L10n.Sources.artifactCopyDetailsLabel(ref:)` = "Copy details of %@". None removed. The board's copy that names a Source does not change scope: "Suggested by this type", "Add field", "Add metadata", "Pick any field from this project's vocabulary, then enter the value as written on the record.", "Search the whole source metadata vocabulary" (the field vocabulary keeps its name, working rule "Metadata lives at three levels"), and the clear confirm "This removes the saved value. Suggested fields return to the dashed list; extra fields disappear." all read correctly on an Artifact. The Source section's intro (`L10n.Sources.metadataIntro`) is drawn only by `SourcePageMetadataView`, not on cards.

## 7. Steps

**STEP-1 — A released value scopes its Source through its deleted owner**
Files: `core/database/effects/export_test.go` *(new)*, `core/database/effects/released_metadata_scope_test.go` *(new)*. Skill: `add-audit-scope`; rule `go-tests.mdc`.
Do: D-1 and T-1. In the test, on `readsFixture`: `metadatafields.Create(tx, "Provider", metadatafields.DataTypeText, "")`; `sourcemetadata.Set(tx, userID, owner, sourcemetadata.Input{FieldID: field.ID, ValueText: "Ancestry"})` once with `owner := sourcemetadata.Owner{Level: vocabkey.LevelArtifact, ID: art.ID}` and once with `sourcemetadata.Owner{Level: vocabkey.LevelGraph, ID: jones.ID}`; run `artifacts.Delete(tx, userID, art.ID)`, then, on a fresh fixture, `evidencegraphs.Delete(tx, userID, jones.ID)`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/...` passes; `go vet -tags fts5 ./core/database/effects/...` is clean.

**STEP-2 — The workspace carries each Artifact's board**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`. Skill: `add-swift-test`.
Do: D-2, D-3.
Check: Xcode tests pass with no Swift warning.

**STEP-3 — The host reads its rows by owner**
Files: `macos/App/Features/Sources/MetadataBoardHost.swift`.
Do: D-4. Update the doc comments (the host is "one owner's rows inside the page's workspace").
Check: Xcode tests pass; `rg -n "rows:|setRows" macos/App/Features/Sources/MetadataBoardHost.swift` prints nothing.

**STEP-4 — A board write patches the page instead of reloading it**
Files: `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/App/Features/Sources/MetadataBoardHost.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`. Skill: `add-workspace-place`.
Do: D-5, D-6; T-2, T-3; add `.wroteMetadataBoard(sourceId: "s1", owner: MetadataOwner(level: .source, id: "s1"), entries: [])` to the `everyMutation` and `mutations` lists.
Check: Xcode tests pass; `rg -n "applyWrite" macos/App/Features/Sources/MetadataBoardHost.swift` prints nothing.

**STEP-5 — The page keeps one board per Artifact**
Files: `macos/App/Features/Sources/SourcePageModel.swift`, `macos/App/Features/Sources/SourceArtifactsSection.swift`.
Do: D-7, D-8, D-9.
Check: Xcode tests pass (the existing `SourcePageModelTests` metadata tests unchanged).

**STEP-6 — The board takes a card style; its chrome moves to its file**
Files: `macos/App/Features/Sources/MetadataBoardView.swift`, `macos/App/Features/Sources/SourcePageChrome.swift`. Skill: `add-ui-component`.
Do: D-10, D-11. `.page` must draw exactly today's view.
Check: the app builds; `rg -n "SourcePageMetadataLabel|SourcePageMetadataValueChrome|sourcePageMetadataValueChrome|SourcePageMetadataSuggestionRow|SourcePageMetadataTextEditor|metadataLabelWidth|metadataColumnSpacing|metadataSuggestionLabelInset|SourcePageLayout\.reorderHandleWidth" macos` prints nothing (`SourcePageMetadataView` and `SourcePageMetadataHost` stay); the Source page's Metadata section looks as on `main` (§10 step 1).

**STEP-7 — Each Artifact card shows Copy details**
Files: `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skills: `add-ui-component`, `add-localized-string`.
Do: D-12, D-13.
Check: the app builds; the localizable check in `scripts/check-pr.sh` passes; §10 steps 2 – 6.

**STEP-8 — Swift tests**
Files: `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skill: `add-swift-test`.
Do: T-4 – T-10. Seed an Artifact's rows with `store.metadataByOwner[MetadataOwner(level: .artifact, id: "a1")]` and its write answers with `store.metadataWriteResults[…]`; drive the page's re-sync with `model.sync(from:)` on the session's `sourceWorkspace` handle.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-9 — Guidance, version and docs**
Files: `.agents/skills/add-workspace-place/SKILL.md` (lines 62, 172), `docs/macos-client-patterns.md` (line 89), `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION`), `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row), `docs/deployment-plan/spike-15/design/README.md`, `docs/deployment-plan/spike-15/design/S15-D1-source-page.md`, this plan (status, Landed note).
Do: line 62 names `.metadataBoard` beside `.updatedSourceShell` ("a board write's `.wroteMetadataBoard` puts the owner's rows on the cached Source page"); line 172 reads "Metadata values: a board's host applies the write's effects with `.wroteMetadataBoard(sourceId:owner:entries:)`, which patches the owner's rows into the cached Source page with no reload (`SourcePageMetadataHost.didWrite`); a value created or released on a Source, Artifact or Evidence graph reports `metadata_field_usage`, which the field list's `usedBy` follows."; line 89 adds, after the `.updatedSource(source)` clause, "a metadata board write applies `.wroteMetadataBoard(sourceId:owner:entries:)`, a patch-only mutation that writes the owner's rows into the cached Source page". The version bump. S15-D1's **Implements later as** line and `design/README.md` note S15-09 landed; the brief stays in `design/` (it gates S15-23 and S15-25). The Definition of done's docs lines.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-09-artifact-metadata-boards.md` reports no errors; `rg -n "context.applyWrite\(effects\)\)" .agents/skills/add-workspace-place/SKILL.md` prints nothing.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show each Artifact's copy details on its card."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/effects/released_metadata_scope_test.go` `TestReleasedMetadataScopesThroughDeletedOwner` *(new)*, table of two cases (Artifact, graph) | The real delete's changes; for each `*_metadata` / `*_metadata_layout` change `i`: `SourceScopeOfChange(tx, changes, i)` → exactly the Source id; then filter the owner's delete (`artifact` / `evidence_graph` entity) out of `changes` and, for each metadata change `j` of the filtered slice, `SourceScopeOfChange(tx, filtered, j)` → an empty result and no error. Each case has exactly one value change and one layout change (one `Set`, which appends one layout row) | AC-3.1, AC-3.2, D-1 |
| T-2 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `metadataBoardWritePatchesTheWorkspace` *(new)* | Source s1 with Artifacts a1, a2 (a1 one saved row, a2 none) and s2 warmed; `apply(effects, .wroteMetadataBoard(sourceId: "s1", owner: .artifact a1, entries: X))` → s1's `artifactMetadata["a1"] == X`, `metadata` and `artifactMetadata["a2"]` unchanged, s2 unchanged; after `settleWrites()` `getSourceWorkspaceCalls` unchanged; the same with owner `.source s1` → `metadata == X` | AC-2.1, D-5 |
| T-3 | Swift | same · `metadataBoardPatchIgnoresOwnersItCannotPlace` *(new)* | `.graph` owner, and `.artifact` "a9" not in `artifacts` → workspace unchanged; `registry.invalidations(by: .wroteMetadataBoard(…))` is empty | D-2, D-5 |
| T-4 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `artifactBoardsFollowTheWorkspace` *(new)* | Artifacts a1 (saved Provider "Ancestry") and a2 (no rows) → `model.artifacts.board(for: "a1")?.saved.map(\.valueText) == ["Ancestry"]`, `board(for: "a2")?.entries.isEmpty == true`, `copyDetailsTitle(id: "a1") == "Copy details (1)"`, `copyDetailsTitle(id: "a2") == "Copy details"` | AC-1.1, D-8, D-13 |
| T-5 | Swift | same · `artifactBoardSaveWritesItsOwnerWithoutReload` *(new)* | a1 with suggestion Image; scripted answer for a1 `[Image "0473"]`; `board(for: "a1").save` → `metadataByOwner[.artifact a1]` is the answer, the Source owner's rows unchanged, the session's `sourceWorkspace` value has `artifactMetadata["a1"]` == answer; after `settleWrites()` `getSourceWorkspaceCalls` unchanged | AC-1.3, AC-2.1, D-6 |
| T-6 | Swift | same · `boardWriteKeepsAnEditBegunElsewhere` *(new)* | Source rows A and B saved, a1 row P saved. Save A (scripted answer), `beginEdit(B)`, `model.sync(from: handle)` → `metadata.editingFieldID == "B"`. Then `board(for: "a1").beginEdit(P)`, save on the Source board, `sync` → a1's `editingFieldID == "P"`. Then a note save (`model.notes` draft set, `await model.notes.add()`, which applies `.mutatedSourceWorkspace`) with rows unchanged, `await session.settleWrites()`, `sync` → both edits still open | AC-2.2, AC-2.3, AC-2.4, D-7 |
| T-7 | Swift | same · `addDialogOnAnArtifactBoardWritesTheArtifact` *(new)* | `metadataDialogs.openAdd(on: board a1)`, field and value, scripted answer for a1 → `createFromAdd` writes a1 (its rows are the answer), Source rows unchanged, dialog closed | AC-1.5 |
| T-8 | Swift | same · `deletingAnArtifactDropsItsBoard` *(new)* | a1 and a2, `openBoardIDs = ["a1", "a2"]`; `askDeleteArtifact("a1")`, `confirmResourceDelete` → `board(for: "a1") == nil`, `openBoardIDs == ["a2"]`, `board(for: "a2")` is the same object as before; the existing one-extra-`getSourceWorkspace` count of the Artifact delete is unchanged | AC-1.10, D-8, D-9 |
| T-9 | Swift | same · `invalidValueOnAnArtifactRowStaysOnTheRow` *(new)* | `store.setMetadataError = CoreInvokeError.coded(…)` with `sourcemetadata.invalid`, as `invalidMetadataSetLandsOnFieldError` builds it; save on a1's saved row → a1's `fieldError` set and `fieldErrorID` the row; the Source board's `fieldError` nil; `model.pageError` nil | AC-1.6 |
| T-10 | Swift | same · `boardsSurviveAReload` *(new)* | a1's board object and `openBoardIDs` before and after a note save (`await model.notes.add()`, then `settleWrites()` and `sync(from:)`) with a1's rows unchanged → same object, still open | AC-1.9 (open state lives in the model), D-8 |
| T-11 | Swift | `macos/ProvenenciaTests/MetadataBoardModelTests.swift` (existing, unchanged) and the `SourcePageModelTests` metadata tests (existing, unchanged) | The Source board behaves as before | D-7, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests` (the reset rule, D-7), `CatalogQueryRegistryTests` (mutation lists), `WorkspaceSessionTests` (patches on `.sourceWorkspace`), `SourceCopyModelTests` (workspace initializers).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings (three keys, D-13); no stale strings left
- [ ] Sweep: `SourcePageMetadataLabel`, `SourcePageMetadataValueChrome`, `sourcePageMetadataValueChrome`, `SourcePageMetadataSuggestionRow`, `SourcePageMetadataTextEditor`, `metadataLabelWidth`, `metadataColumnSpacing`, `metadataSuggestionLabelInset`, `SourcePageLayout.reorderHandleWidth`, `setRows:` `rg`'d across `macos/`, `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory (Copy details disclosure, Artifact card, Metadata board); no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief stays in design/ (it gates S15-23, S15-25); design/README.md and S15-D1's header note S15-09 landed
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/macos-client-patterns.md line 89 (STEP-9)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Before the change (on `main`), open a Census Source with saved Source metadata and an unfilled suggestion; screenshot its Metadata section at Regular and at Wide. On the branch the section matches both (D-10 `.page`, D-11).
2. (AC-1.1) On the branch, open a Source with 12 Artifacts whose type suggests provider, URL and accessed date at the Artifact level (set the suggestions on the type page if none, or through `AddTypeSuggestion` with level `artifact`). Every card shows "Copy details", closed.
3. (AC-1.2, AC-1.3, AC-1.5, AC-1.7) Open one card's Copy details first on an Artifact of a type with no Artifact suggestions: only "Add field" shows. Then on the census card: the suggestions show under "Suggested by this type" with "Add field" at the end; only that card grows (SD-10). Fill Provider "Ancestry" and press Return: the title reads "Copy details (1)". Add a URL through "Add field"; click it to open the browser. Accessed date "12 March 2025".
4. (AC-1.4, AC-2.2) Edit the URL, then drag Accessed to the top, then clear Provider (confirm), then dismiss a suggestion. Each shows at once, and a row you start editing right after a save stays in edit. Reopen the page: the board holds what you left.
5. (AC-1.8, AC-1.9) Resize the window across 1600pt with the disclosure open: it stays open. Put the card in edit (pencil): the disclosure hides; Cancel: it is back, open.
6. With VoiceOver, the disclosure reads "Copy details of ART-…" and collapsed / expanded. With Accessibility Inspector, `sources.page.artifact.<id>.copyDetails` and `sources.page.artifact.<id>.metadata.add` are present.
7. (AC-1.10) Delete an Artifact that has values: its card goes; the Metadata fields page's Used by for that field drops by one.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `PVReorderableList` (a `List`) inside a `LazyVGrid` card measures wrongly | It already sits inside the page's scroll view; §10 step 3 – 4; Stop condition below |
| A board model made in `apply` while a view reads `boards` | `apply` runs from `sync` (`onChange`, `task`, tests), never inside a body |
| The patch lands while the key's first load is in flight | `patchCached` returns `.awaitingFirstValue` and `apply` reloads the key, as for `.updatedSourceShell` |
| D-7 changes when the Source board resets after a reload | T-6 and the unchanged Source board tests pin it |

**Stop and report if:**

- `GetSourceWorkspaceResponse.artifact_metadata`, `SourcePageMetadataHost`, `MetadataBoardModel`'s patch-then-`didWrite` order, or `.sourceWorkspace`'s `patches` on `main` no longer match §4.
- An existing `SourcePageModelTests` or `MetadataBoardModelTests` metadata test needs an edit to pass: D-7 or D-6 changed behavior the plan says stays.
- `PVReorderableList` inside a card does not show its rows or does not reorder at the card's width.
- The generated Swift field for `MetadataBoard.owner_id` is not `ownerID`, or an Artifact's `id` and its board's `owner_id` differ in format for the same Artifact.
- `SourceScopeOfChange` resolves a released change to the Source without the owner's delete among the ghosts (the scope then does not come through the owner, and CON-2's premise is wrong).
- `artifacts` or `evidencegraphs` cannot be imported from `package effects_test` (an import cycle).

## 12. Open questions

None.

## 13. PR

**Title:** S15-09 — Artifacts show their metadata on the Source page
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-09-artifact-metadata-boards.md`.
