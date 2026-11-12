# S15-26 — The metadata board takes an owner

**Status:** Draft
**Spike / slice:** Spike 15 · Slice 2 — Metadata at three levels
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Source page's Metadata board is drawn by the same views with the same copy, spacing and identifiers; only the types behind it change)
**Board:** —
**Depends on:** — (S15-08, which gave the four board RPCs an owner on the wire, landed in [#791](https://github.com/mendahu/provenencia/pull/791))
**Blocks:** S15-09, S15-10, S15-11
**Branch:** `s15-26-metadata-board-takes-an-owner`
**Version bump:** none — a preparatory refactor with no change a researcher can see, no wire and no catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (the board stays a feature snowflake), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "One metadata board", "Metadata lives at three levels with one vocabulary", "One level type" and "Metadata saves with its owner's write" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); S15-08's [plan](S15-08-artifacts-and-graphs-hold-metadata.md) D-11, D-16 and **Next adds**; S15-07's [plan](S15-07-metadata-operations-take-an-owner.md) D-7 (a draft board sends its whole order); [S15-D1](../design/S15-D1-source-page.md) SD-9 and its inventory row "Metadata board" (what S15-09 builds on this)
**Planned:** 2026-11-12 against `main` @ `7d2f3e7`
**Review:** — (not reviewed yet)
**Runs on:** macOS. `scripts/check-pr.sh` and every Swift check need Xcode ([`README.md`](../../../../README.md)). Without it, stop.

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

The Source page's Metadata board becomes one board model and one board view that work on any metadata owner (a Source, an Artifact, an Evidence graph), so S15-09 (a board on each Artifact card), S15-10 (the add flow) and S15-11 (the graph form) instance the same board instead of copying it. The Swift store's four metadata writes take a `MetadataOwner` (level and id), matching the owner RPCs S15-08 put on the wire. The Source page moves onto the new board and looks and behaves exactly as today.

**Why now.** Working rule "One metadata board" (map review K-2, kept by the user 2026-11-09): S15-09, S15-10 and S15-11 each need the board, and all three depend on this PR. S15-08 left the Swift store's Source-named methods for this PR (its D-16).

**What stays.** Every pixel, string, accessibility identifier, keyboard path and error placement of the Source page's Metadata section, the Add metadata dialog and the clear confirm. The `sourceWorkspace` query key and its mutations (`.mutatedSourceMetadata`, `.mutatedSourceWorkspace`) and what each write sends to the session. The wire (`engine.proto`) and Go are untouched.

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

**US-2.** As a contributor building S15-09 – S15-11, I instance the board on any owner by giving it an object that holds the owner's rows and runs its writes, so that every board in the app behaves the same.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A `MetadataBoardModel` on a test owner with rows `[Author (suggested, empty)]`; save "Mary Robins" on Author | The test owner received `set(fieldID: "f-author", valueText: "Mary Robins")`, its `entries` hold Author saved, and `didWrite(…, valuesChanged: true)` was called once after `entries` changed |
| AC-2.2 | Same owner; reorder writes fail | The owner's `entries` are back in their earlier order and the owner's `showFailure` got the message |
| AC-2.3 | Same owner; `failureMessage(for:)` returns nil (the owner is gone) | No field error, no banner, `entries` unchanged |
| AC-2.4 | Dismiss the Author suggestion | The owner received `dismissSuggestion(fieldID: "f-author")`, `entries` are what it returned, `didWrite(…, valuesChanged: false)` was called |
| AC-2.5 | `GenealogyStore.setMetadata` with `MetadataOwner(level: .artifact, id: "a1")` on `FakeStore` | The Artifact's list holds the value; the Source `s1`'s list is unchanged |

## 3. Scope

| In | Covered by |
| --- | --- |
| `MetadataOwner` *(new)* value type; the four store methods renamed and taking an owner (`GenealogyStore`, `GoStore`, `FakeStore`, `ThrowingStore`) | STEP-1, AC-2.5 |
| `MetadataBoardOwner` *(new)* protocol: the owner's rows, its writes, its field vocabulary and where its failures show | STEP-2, AC-2.1 – AC-2.4 |
| `SourceMetadataSection` becomes `MetadataBoardModel` *(renamed)*, driven by a `MetadataBoardOwner` | STEP-3, AC-1.1 – AC-1.8 |
| `SourceMetadataBoardOwner` *(new)*: the Source page's owner over `SourcePageContext` | STEP-3, AC-1.1 – AC-1.8 |
| `MetadataBoardView` *(new)* and `MetadataBoardAddForm` *(new)* split out of `SourcePageMetadataView`; the Source page instances them | STEP-4, manual check |
| Board tests against a test owner | STEP-5, T-1 – T-6 |

| Out | Home |
| --- | --- |
| A board on each Artifact card; mapping `GetSourceWorkspaceResponse.artifact_metadata` into Swift; an Artifact owner | S15-09 |
| A draft owner (the board on a form, collecting a `sourcemetadata.Draft` with its whole order, no store writes) and the Swift `Draft` type sent on `CreateArtifactRequest.metadata` | S15-10 (S15-11 reuses it) |
| `GenealogyStore.getMetadataBoard` over `METHOD_GET_METADATA_BOARD` | S15-11 (its first reader) |
| Renaming `SourcePageMetadataLabel`, `SourcePageMetadataValueChrome`, `SourcePageLayout.metadata*` and the `L10n.Sources.metadata*` keys to board names | S15-09 decides when a second view first draws them; not planned before that |
| Any change to the Add metadata dialog or the clear confirm being hosted once by the page shell | not planned (S15-09 keeps one dialog per page; see D-7) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Board model | `macos/App/Features/Sources/SourceMetadataSection.swift` `SourceMetadataSection` | `@Observable` class over `SourcePageContext`. Reads rows from `context.workspace?.metadata`, writes through `context.store.setSourceMetadata` / `clearSourceMetadata` / `dismissSourceMetadataSuggestion` / `reorderSourceMetadata` with `context.sourceID`, patches `context.workspace?.metadata`, then calls `context.applyMetadataWrite(effects)` (set, clear) or `context.applyWrite(effects)` (dismiss, reorder). Failures: `context.clearPageError()` before each write; `applySetError` sends `sourcemetadata.invalid` to `fieldError` / `addValueError` and anything else to `context.showPageError`; clear, dismiss and reorder use `context.showFailure`; `context.fail` returns nil for a deleted Source. Field options come from `context.fields`. Bound to the Source: it cannot serve an Artifact or a form |
| Delete item | same file `SourceMetadataDeleteItem` | Snapshot for the clear confirm; used by `SourcePageView`'s `pvConfirm(item:)` |
| Board view | `macos/App/Features/Sources/SourcePageMetadataView.swift` `SourcePageMetadataView` | Takes `SourcePageModel`; draws the section header (title, count, Add), the intro line, the empty state, the `PVReorderableList` of saved rows and the suggestion rows, all through `model.metadata`. `addForm` is the Add dialog's body. Private `SourcePageMetadataSuggestionRow` and `SourcePageMetadataTextEditor` draw one row each. Identifiers are literal `"sources.page.metadata…"` / `"sources.page.addMetadata…"` strings |
| Page wiring | `macos/App/Features/Sources/SourcePageModel.swift` `SourcePageModel.metadata`, `init`, `apply(_:)`, `acceptsArtifactPageDrop` | Builds `SourceMetadataSection(context:)`; calls `metadata.resetDrafts()` when the workspace's metadata changed; reads `metadata.isAdding` and `metadata.pendingDelete` |
| Page shell | `macos/App/Features/Sources/SourcePageView.swift` `pvFormDialog` (Add metadata, around line 164), `pvConfirm` (clear, around line 183), `addMetadataPresented`, `SourcePageOverview` | Hosts the Add dialog with `SourcePageMetadataView(model: model).addForm` and the clear confirm, both reading `model.metadata` |
| Page context | `macos/App/Features/Sources/SourcePageContext.swift` `SourcePageContext` | Owns `workspace`, `fields`, `clearPageError`, `fail`, `showPageError`, `showFailure`, `applyWrite`, `applyMetadataWrite`. Unchanged by this PR |
| Row chrome | `macos/App/Features/Sources/SourcePageChrome.swift` `SourcePageLayout`, `SourcePageMetadataLabel`, `sourcePageMetadataValueChrome()` | Shared label and value chrome for the rows. Unchanged |
| Store protocol | `macos/App/Platform/GenealogyStore.swift` `setSourceMetadata`, `clearSourceMetadata`, `dismissSourceMetadataSuggestion`, `reorderSourceMetadata` (around line 1380) | Take `sourceID: String` |
| Go store | `macos/App/Platform/GoStore.swift` same four methods (around line 336) | Build the owner requests with `req.owner.level = Level.source.rawValue`, `req.owner.id = sourceID` (S15-08 D-16) |
| Fake store | `macos/App/Platform/FakeStore.swift` same four methods (around line 724), `metadataBySource`, `setSourceMetadataError`, `reorderSourceMetadataError` | Keyed by Source id only |
| Throwing stub | `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore` (around line 424) | Implements the four methods with `throw StoreBoom.boom` |
| Level | `macos/App/Platform/Level.swift` `Level`, `Level.metadataLevels` | Mirrors `vocabkey.Level` (S15-06). Unchanged |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` `loadMetadataSuggestionsAndValues` … `metadataSetIOFailureLandsOnPageError` (lines 698 – 1105) and the banner test around line 2230 | Drive the board through `model.metadata` and `FakeStore`; reference `store.setSourceMetadataError` and `store.reorderSourceMetadataError` |

**Precedent.** No earlier Swift section model was moved behind a protocol; the closest shape is `WorkspaceLeaveGuard` (`macos/App/Features/Workspace/WorkspaceNavigation.swift`): an `AnyObject` protocol a feature model conforms to so shared code drives it. Copy its form (a small `@MainActor` protocol, a doc comment per requirement). S15-08 (#791) is the precedent for the store side: copy how `GoStore` fills `req.owner`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `struct MetadataOwner: Sendable, Hashable { var level: Level; var id: String }` *(new)* in `macos/App/Platform/GenealogyStore.swift`, beside `CatalogSourceWorkspace`. Swift does not check `level.holdsMetadata`; Go answers `sourcemetadata.invalid` | Mirrors the proto `MetadataOwner`; one place decides validity (Go) | A Swift enum with associated ids (`.source(id)`) — a second spelling of `Level`; a Swift-side precondition — duplicates Go's rule |
| D-2 | The four store methods become `setMetadata(projectDir:userID:owner:fieldID:valueText:)`, `clearMetadata(projectDir:userID:owner:fieldID:)`, `dismissMetadataSuggestion(projectDir:userID:owner:fieldID:)`, `reorderMetadata(projectDir:userID:owner:fieldIDs:)`, with today's return types. The Source-named methods are deleted, not kept beside them | The names match the RPCs (S15-08 D-11); one path per operation | Overloads keeping `sourceID:` — two paths for one write |
| D-3 | `FakeStore` keeps `metadataBySource` (seeded by many tests) for `.source` owners and adds `metadataByOwner: [MetadataOwner: [CatalogMetadataEntry]]` *(new)* for `.artifact` and `.graph` owners; one private helper `metadataRows(for:)` / `setMetadataRows(_:for:)` picks the map. `setSourceMetadataError` → `setMetadataError`, `reorderSourceMetadataError` → `reorderMetadataError`. The fake adds no level validation (`swift-tests.mdc`: no vocabulary `CoreInvokeError`s in the fake) | Existing seeds keep working; an Artifact write in a later test never lands on a Source; S15-09 seeds `metadataByOwner` | Re-keying every seed to `metadataByOwner` — churn in four test files for no behavior; throwing for non-Source levels — the rule forbids it |
| D-4 | `@MainActor protocol MetadataBoardOwner: AnyObject` *(new)* in `macos/App/Features/Sources/MetadataBoardOwner.swift` *(new)* with exactly: `var entries: [CatalogMetadataEntry] { get set }`; `var fields: [CatalogMetadataField] { get }`; `func set(fieldID: String, valueText: String) async throws -> (CatalogMetadataEntry, CatalogWriteEffects)`; `func clear(fieldID: String) async throws -> CatalogWriteEffects`; `func dismissSuggestion(fieldID: String) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects)`; `func reorder(fieldIDs: [String]) async throws -> ([CatalogMetadataEntry], CatalogWriteEffects)`; `func didWrite(_ effects: CatalogWriteEffects, valuesChanged: Bool)`; `func clearFailure()`; `func failureMessage(for error: Error) -> String?` (nil: the owner is gone and its host is leaving; write no state); `func showFailure(_ message: String)` | The rule names the owner as "its entries and its write operations". The board patches `entries` before `didWrite`, keeping today's order (patch, then `session.apply`). A draft owner (S15-10) conforms by returning `.empty` effects and doing nothing in `didWrite`; a form shows failures in its own error line | A struct of closures — ten closures with no name for the bundle; the board calling `SourcePageContext` — the coupling this PR removes; the owner applying effects inside `set` — the session would see effects before the board patched rows |
| D-5 | `SourceMetadataSection` is renamed `MetadataBoardModel` and moves to `macos/App/Features/Sources/MetadataBoardModel.swift` *(new)* (`git mv`). Its init is `init(owner: any MetadataBoardOwner)`. Its public surface (`drafts`, `editingFieldID`, `savingFieldID`, `fieldError`, `fieldErrorID`, `pendingDelete`, the add-dialog state, `entries`, `saved`, `suggested`, `fieldComboOptions`, `addSheetIsDirty`, `canSubmitAdd`, `resetDrafts`, `beginEdit`, `cancelEdit`, `save`, `askClear`, `confirmClear`, `dismissSuggestion`, `moveSaved`, `openAdd`, `cancelAdd`, `createFromAdd`, `applySetError`) keeps its names and behavior. Each `context.*` call maps one to one: `context.workspace?.metadata` → `owner.entries`; `context.fields` → `owner.fields`; `context.store.<x>SourceMetadata…` → `owner.<x>`; `context.applyMetadataWrite(e)` → `owner.didWrite(e, valuesChanged: true)`; `context.applyWrite(e)` → `owner.didWrite(e, valuesChanged: false)`; `context.clearPageError()` → `owner.clearFailure()`; `context.fail(e)` → `owner.failureMessage(for: e)`; `context.showPageError(m)` → `owner.showFailure(m)`; `context.showFailure(e)` → `if let m = owner.failureMessage(for: e) { owner.showFailure(m) }`. `SourceMetadataDeleteItem` is renamed `MetadataBoardDeleteItem` | A mechanical move keeps every behavior the page tests pin; the names say what the type is now | A fresh model — risks the focus and lag fixes the current one carries (`SourcePageMetadataTextEditor`'s local draft) |
| D-6 | `final class SourceMetadataBoardOwner: MetadataBoardOwner` *(new)*, in `MetadataBoardOwner.swift`, holds `SourcePageContext` (unowned is not used: `let context`; the page model owns both) and `MetadataOwner(level: .source, id: context.sourceID)`. `entries` reads and writes `context.workspace?.metadata` (a write when `workspace` is nil does nothing); `fields` is `context.fields`; writes call the D-2 store methods with `context.projectDir`, `context.userID` and its owner; `didWrite(e, valuesChanged: true)` → `context.applyMetadataWrite(e)`, `false` → `context.applyWrite(e)`; the three failure requirements forward to `context.clearPageError`, `context.fail`, `context.showPageError`. `SourcePageModel.init` builds `metadata = MetadataBoardModel(owner: SourceMetadataBoardOwner(context: context))` | One owner per host keeps the page's session wiring where it is; `SourcePageContext` is unchanged | `extension SourcePageContext: MetadataBoardOwner` — the page can then hold only one board, and S15-09 puts up to 60 on it |
| D-7 | `MetadataBoardView` *(new)*, `macos/App/Features/Sources/MetadataBoardView.swift` *(new)*: `struct MetadataBoardView: View { @Bindable var board: MetadataBoardModel; let accessibilityIdentifierPrefix: String }`. It draws what `SourcePageMetadataView.body` draws below the intro line: the empty state, the saved list and the suggestions, with `SourcePageMetadataSuggestionRow` and `SourcePageMetadataTextEditor` moved into this file (still `private`). `MetadataBoardAddForm` *(new)*, same file: `struct MetadataBoardAddForm: View { @Bindable var board: MetadataBoardModel; let accessibilityIdentifierPrefix: String }`, today's `addForm`. Every literal identifier becomes `"\(accessibilityIdentifierPrefix)…"` with the same suffix; the Source page passes `"sources.page.metadata"` to the board and `"sources.page.addMetadata"` to the form, so every identifier string is unchanged. `SourcePageMetadataView` keeps the section header (its Add button keeps `"sources.page.addMetadata"`) and the intro line, then `MetadataBoardView(board: model.metadata, accessibilityIdentifierPrefix: "sources.page.metadata")`; its `addForm` is deleted and `SourcePageView`'s Add dialog body becomes `MetadataBoardAddForm(board: model.metadata, accessibilityIdentifierPrefix: "sources.page.addMetadata")`. The Add dialog and the clear confirm stay hosted once by the page shell | S15-D1 SD-9 puts "the same rows, suggestions and Add" on an Artifact card under a Disclosure, without the Source's header and intro, so the header stays with the page. One dialog per page, not 60 | Moving the header into the board — the Artifact card draws a Disclosure instead; one dialog per board — 60 dialog modifiers on a Source page |
| D-8 | Board copy stays `L10n.Sources.metadata*`; no string, key or `.xcstrings` entry changes | No chrome change; S15-09's brief decides any Artifact wording | New `L10n.MetadataBoard` namespace now — a rename with no second caller |
| D-9 | New tests in `macos/ProvenenciaTests/MetadataBoardModelTests.swift` *(new)* drive `MetadataBoardModel` through `TestMetadataBoardOwner` *(new, private to that file)*: records each call and its arguments in order, returns scripted results or throws a scripted error, and returns a scripted `failureMessage`. The existing `SourcePageModelTests` metadata tests stay as the Source-owner tests, edited only for D-3's renamed fake properties | The map's Check: "Swift tests drive the board against a fake owner"; the page tests prove AC-1.x unchanged | Rewriting the page tests on the test owner — loses the end-to-end Source check |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `MetadataOwner` *(new)* | type | `Provenencia_Engine_V1_MetadataOwner` (generated) | Generated types stay inside `GoStore` (`macos-client.mdc`); the app speaks `Level` |
| `MetadataBoardOwner` *(new)* | protocol | `SourcePageContext` | The context is one Source page's dependencies; the board needs one owner's rows and writes (searched: `rg "protocol " macos/App/Features`) |
| `SourceMetadataBoardOwner` *(new)* | type | `SourcePageContext` | Extending the context allows one owner per page (D-6) |
| `MetadataBoardModel` *(renamed from `SourceMetadataSection`)* | type | itself | Rename and re-home, not new |
| `MetadataBoardView`, `MetadataBoardAddForm` *(new)* | view | `SourcePageMetadataView` | Split from it; the page keeps its header (D-7) |
| `metadataByOwner`, `metadataRows(for:)`, `setMetadataRows(_:for:)` *(new)* | fake state, helper | `FakeStore.metadataBySource` | Kept for Source owners (D-3) |
| `TestMetadataBoardOwner` *(new)* | test helper | `FakeStore` | The board's contract is the protocol, not the store; a store-backed owner would retest `FakeStore` |

**Builds on.** `Level` (S15-06), the owner RPCs (S15-08), `SourcePageContext`'s failure door (`fail`, `showPageError`) and session apply calls, the `WorkspaceLeaveGuard` protocol shape.
**Deletes.** `SourceMetadataSection` (renamed), `SourceMetadataDeleteItem` (renamed), `SourcePageMetadataView.addForm`, the four `…SourceMetadata…` store methods in `GenealogyStore`, `GoStore`, `FakeStore` and `ThrowingStore`, and the fake's `setSourceMetadataError` and `reorderSourceMetadataError`. No string or test is deleted.
**Next adds.** S15-09 maps `artifact_metadata`, adds an Artifact owner (a second conformer, or `SourceMetadataBoardOwner` generalized over its `MetadataOwner`; S15-09 decides) and puts `MetadataBoardView` in each card's Disclosure. S15-10 adds a draft owner and the Swift `Draft`; S15-11 reuses it and adds `getMetadataBoard`.

### Reach

None — no new path into existing code, no new state reachable: the only owner this PR builds is the Source page's, over the same workspace.

### Scale

None — no write, list or loop changes: each board operation makes the same one store call and the same one `session.apply` as today. The protocol adds one dynamic dispatch per operation.

### Edges

None — no rule, comparison or algorithm changes; the board's rules (saved vs suggested, reorder of saved rows only, optimistic move and revert) move unchanged and are pinned by the existing page tests and T-1 – T-6.

### Schema

None — no catalog change.

### Go core

None — Go is untouched.

### FFI

None — the owner RPCs exist (S15-08). `GoStore` only renames its four methods and reads `owner.level.rawValue` and `owner.id` instead of `Level.source.rawValue` and `sourceID`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Section header, intro | `PVSectionHeader`, `Text` | `SourcePageMetadataView` | Unchanged; reads `model.metadata.saved.count`, calls `model.metadata.openAdd()` |
| Board rows | `PVEmptyState`, `PVReorderableList`, `PVInlineEdit`, `PVTextArea`, `PVCard(border: .dashed)`, `PVInput`, `PVButton`, `PVIconButton` | `MetadataBoardView` *(new)* | Moved code; same modifiers and identifiers |
| Add dialog body | `PVField`, `PVComboBox`, `PVInput` | `MetadataBoardAddForm` *(new)* | Hosted by `SourcePageView`'s `pvFormDialog` as today |
| Clear confirm | `pvConfirm(item:)` | `SourcePageView` | Reads `model.metadata.pendingDelete` (`MetadataBoardDeleteItem?`) |

The board is a feature snowflake (S15-D1 inventory row "Metadata board": Snowflake) in `Features/Sources/`; no kit component changes. States: empty, saved only, suggestions only, both, editing, saving, field error — all as today. Accessibility: identifiers and labels unchanged.

### Strings

None — no key added, changed or removed (D-8).

## 7. Steps

**STEP-1 — Store methods take an owner**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift` (`ThrowingStore`), `macos/ProvenenciaTests/SourcePageModelTests.swift`, `macos/App/Features/Sources/SourceMetadataSection.swift` (call sites only). Skill: `add-swift-test`.
Do: add `MetadataOwner` (D-1); rename the four methods (D-2) in the protocol and all three conformers; `GoStore` fills `req.owner` from the parameter; `FakeStore` per D-3; `SourceMetadataSection` calls the new methods with `MetadataOwner(level: .source, id: context.sourceID)` (temporary, replaced in STEP-3); rename `setSourceMetadataError` / `reorderSourceMetadataError` uses in `SourcePageModelTests`. Add T-6 and T-8.
Check: the Xcode test run (`scripts/check-pr.sh`'s Xcode step, or `xcodebuild test` as the README gives it) passes with no Swift warning; `rg -n "SourceMetadata\(|setSourceMetadata|clearSourceMetadata|dismissSourceMetadataSuggestion|reorderSourceMetadata" macos` prints nothing.

**STEP-2 — The owner protocol and the Source owner**
Files: `macos/App/Features/Sources/MetadataBoardOwner.swift` *(new)*.
Do: `MetadataBoardOwner` (D-4) and `SourceMetadataBoardOwner` (D-6), each requirement with a one-line doc comment; the `failureMessage` comment says nil means the owner is gone and no state is written.
Check: the app builds.

**STEP-3 — The board model takes an owner**
Files: `git mv` of `macos/App/Features/Sources/SourceMetadataSection.swift` to `MetadataBoardModel.swift` *(new)* in the same folder; `macos/App/Features/Sources/SourcePageModel.swift`; `macos/App/Features/Sources/SourcePageView.swift` (the `pvConfirm` item type only, if spelled).
Do: rename per D-5 and replace every `context` use with the D-5 mapping; drop the `context` property; `SourcePageModel.metadata` is `MetadataBoardModel`, built per D-6.
Check: Xcode tests pass, every `SourcePageModelTests` metadata test unchanged except STEP-1's renames; `rg -n "context" macos/App/Features/Sources/MetadataBoardModel.swift` prints nothing.

**STEP-4 — The board view takes the model**
Files: `macos/App/Features/Sources/MetadataBoardView.swift` *(new)*, `macos/App/Features/Sources/SourcePageMetadataView.swift`, `macos/App/Features/Sources/SourcePageView.swift`. Skill: `add-ui-component`.
Do: per D-7: move the empty state, saved list and suggestions into `MetadataBoardView`, the two private row views with them, and `addForm` into `MetadataBoardAddForm`; prefix every identifier; `SourcePageMetadataView` keeps header and intro and instances the board; `SourcePageView`'s Add dialog body is `MetadataBoardAddForm`.
Check: the app builds; `rg -n '"sources\.page\.(metadata|addMetadata)' macos/App/Features/Sources/MetadataBoardView.swift` prints nothing (all prefixed); every identifier listed in §10 step 3 still exists when the page runs (Accessibility Inspector or the view debugger).

**STEP-5 — Board tests on a test owner**
Files: `macos/ProvenenciaTests/MetadataBoardModelTests.swift` *(new)*. Skill: `add-swift-test`.
Do: T-1 – T-5 with `TestMetadataBoardOwner` (D-9).
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row status), this plan (status, Landed note).
Do: per the Definition of done. No domain doc names the Swift board types (`rg` at planning found none in `docs/*.md`).
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-26-metadata-board-takes-an-owner.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Let the Swift store's metadata writes take an owner."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/MetadataBoardModelTests.swift` `saveFillsSuggestionThenReportsWrite` *(new)* | Owner rows `[Author suggested, empty]`; `drafts["f-author"] = "Mary Robins"`; `save` → calls `[set(f-author, "Mary Robins"), didWrite(valuesChanged: true)]` in that order; `owner.entries` Author `hasValue`; `board.suggested` empty | AC-2.1, D-4, D-5 |
| T-2 | Swift | same · `reorderFailureRevertsAndShowsFailure` *(new)* | Saved `[A, B]`, scripted reorder error, `failureMessage` "Boom" → `moveSaved` leaves `owner.entries` `[A, B]`; `showFailure("Boom")` called once; no `didWrite` | AC-2.2 |
| T-3 | Swift | same · `goneOwnerWritesNoState` *(new)* | `failureMessage` scripted nil; `set` throws `sourcemetadata.invalid` → `fieldError` nil, no `showFailure`, `owner.entries` unchanged; same for `confirmClear` (`pendingDelete` stays set) | AC-2.3, AC-1.8 |
| T-4 | Swift | same · `dismissSuggestionTakesOwnersRows` *(new)* | Rows `[Author suggested]`; dismiss returns `[]` → `owner.entries == []`; `didWrite(valuesChanged: false)` called | AC-2.4 |
| T-5 | Swift | same · `invalidValueStaysOnTheFieldOrDialog` *(new)* | `set` throws `sourcemetadata.invalid`, `failureMessage` "Bad" → editing a saved row: `fieldError == "Bad"`, `fieldErrorID` the row; from the Add dialog: `addValueError == "Bad"`; no `showFailure` in either | AC-1.3, D-5 |
| T-6 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `artifactOwnerMetadataLeavesTheSourceAlone` *(new)* | `FakeStore` with `metadataBySource["s1"] = [Repository]`; `setMetadata(owner: MetadataOwner(level: .artifact, id: "a1"), fieldID: "f-author", valueText: "Ancestry")` → `metadataByOwner[(.artifact, "a1")]` holds Author; `metadataBySource["s1"]` still `[Repository]` | AC-2.5, D-3 |
| T-7 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `loadMetadataSuggestionsAndValues` (existing, unchanged, now through `SourceMetadataBoardOwner`; so are T-7a – T-7f) | Saved and suggested rows split | AC-1.1 |
| T-7a | Swift | same · `saveMetadataValueFillsSuggestion` (existing) | Suggestion filled → saved | AC-1.2 |
| T-7b | Swift | same · `invalidMetadataSetLandsOnFieldError` (existing) | Invalid value → row error | AC-1.3 |
| T-7c | Swift | same · `metadataSetIOFailureLandsOnPageError` (existing) | I/O failure → banner | AC-1.4 |
| T-7d | Swift | same · `reorderSavedMetadataFailureRevertsAndSetsPageError` (existing) | Failed reorder → revert and banner | AC-1.5 |
| T-7e | Swift | same · `clearSuggestedMetadataReturnsToDashedList` (existing) | Cleared suggestion → dashed row | AC-1.6 |
| T-7f | Swift | same · `addMetadataRequiresFieldAndValue` (existing) | No field → field error, no write | AC-1.7 |
| T-8 | Swift | same · `metadataWriteOnDeletedSourceLeavesThePage` *(new)* | Saved Repository; `store.setMetadataError = FakeStore.sourceNotFound`; edit and save → `model.gone != nil`, `model.pageError == nil`, `model.metadata.fieldError == nil`, the workspace row unchanged (the pattern of the `FakeStore.sourceNotFound` tests near line 2256) | AC-1.8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests`, `SourceCopyModelTests` and `CatalogQueryRegistryTests` (they seed `metadataBySource`, which stays).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected: D-8)
- [ ] Sweep: `SourceMetadataSection`, `SourceMetadataDeleteItem`, `SourcePageMetadataView.addForm`, `setSourceMetadata`, `clearSourceMetadata`, `dismissSourceMetadataSuggestion`, `reorderSourceMetadata`, `setSourceMetadataError`, `reorderSourceMetadataError` `rg`'d across `macos/`, `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches today's Source page (no brief: the manual check compares before and after); no new kit primitive
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated — not gated, nothing to move
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Before the change (on `main`), open a project with a Census Source that has two saved metadata values, one URL value, and one unfilled type suggestion. Screenshot the page at Regular and at Wide (pane wider than 1600pt).
2. On the branch, open the same Source. The Metadata section matches both screenshots: header with count and Add, intro, saved rows, "Suggested by this type" with the dashed row.
3. Fill the suggestion and press Return; edit a saved value and save; open the URL value's link; drag a row to a new place and reopen the page (the order holds); clear a value that was a suggestion (it returns as a dashed row); dismiss a suggestion; Add metadata with no field (the field error shows), then add a field and value. With Accessibility Inspector, the identifiers `sources.page.addMetadata`, `sources.page.metadata.list`, `sources.page.metadata.suggestions`, `sources.page.metadata.<fieldID>.value`, `sources.page.addMetadata.field` and `sources.page.addMetadata.value` are present.
4. Open the Metadata fields page: the field you just used shows its Used by count raised by one (the `valuesChanged: true` path still applies `.mutatedSourceMetadata`).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The page's `@Bindable` bindings (`$model.metadata.addFieldID`) stop compiling through `@Bindable var board` | `MetadataBoardModel` stays `@Observable` `final class`; the form binds `$board.addFieldID` |
| `didWrite` called before rows are patched changes refresh order | D-4/D-5 fix the order: patch `entries`, then `didWrite`; T-1 pins it |
| A moved identifier string changes | STEP-4's check and §10 step 3 |

**Stop and report if:**

- `SourceMetadataSection` or `SourcePageMetadataView` on `main` no longer matches §4 (S16 or S15 work moved it).
- A `SourcePageModelTests` metadata test needs a change other than D-3's renamed fake properties to pass — that is a behavior change.
- The Swift compiler rejects `any MetadataBoardOwner` as an `@Observable` stored property with `entries { get set }` in a way that needs the protocol shape changed (D-4).
- A conformer of `GenealogyStore` other than `GoStore`, `FakeStore` and `ThrowingStore` exists.

## 12. Open questions

None.

## 13. PR

**Title:** S15-26 — The metadata board takes an owner
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-26-metadata-board-takes-an-owner.md`.
