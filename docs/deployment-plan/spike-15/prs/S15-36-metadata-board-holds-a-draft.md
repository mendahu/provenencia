# S15-36 — A metadata board can hold a draft

**Status:** Draft
**Spike / slice:** Spike 15 · Slice 2 — Metadata at three levels
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the draft host, the add row's state and the refusal mapping have no view until S15-34; the Source page's board draws and behaves as today)
**Board:** —
**Depends on:** S15-26 (landed [#798](https://github.com/mendahu/provenencia/pull/798)), S15-33 (landed [#806](https://github.com/mendahu/provenencia/pull/806))
**Blocks:** S15-34, S15-10, S15-11
**Branch:** `s15-36-metadata-board-holds-a-draft`
**Version bump:** none — a preparatory change with no change a researcher can see, no wire and no catalog change: the app sends no draft until S15-34 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (read only: the board stays a feature snowflake; nothing is added to the kit)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** working rules "One metadata board", "A board adds a field in an inline row", "Metadata saves with its owner's write", "An Artifact is added and edited in one sheet" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); S15-26's [plan](S15-26-metadata-board-takes-an-owner.md) **Next adds** (what a draft host does) and its Landed note (D-4: `FakeStore` runs no engine rule); S15-33's [plan](S15-33-artifact-update-saves-metadata.md) D-2 (both Artifact replies carry the board) and D-6 (a refused value's `params[0]` is its field's wire id); S15-07's [plan](S15-07-metadata-operations-take-an-owner.md) D-7 (a draft board sends its whole order); [S15-D14](../design/S15-D14-source-page-artifacts.md) SA-8, SA-9, SA-13 (what S15-34 draws on this)
**Planned:** 2026-11-13 against `main` @ `cd64512`
**Review:** — (not reviewed yet)
**Runs on:** macOS. Every check needs Xcode ([`README.md`](../../../../README.md)). No proto change: never run `./scripts/generate-proto.sh` in this PR

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

The metadata board can work on rows that nothing has saved yet, so the Artifact sheet (S15-34, edit; S15-10, add) and the Evidence graph form (S15-11) put the same board in a form whose one Save carries the copy details in its owner's write. This PR adds the draft host, the Swift metadata draft and its wire mapping, `metadata:` on the four Artifact and graph store writes, the rows an Artifact write answers, the add row's state on every board, a clear with no confirm on a draft, and the mapping of a refused value to its row. Nothing draws any of it yet.

**Why now.** S15-34, S15-10 and S15-11 each need the draft board; the map makes this their shared preparatory PR (map review, user 2026-11-12). S15-33 put the draft on `UpdateArtifactRequest` and the board on both Artifact replies; no Swift reads them yet.

**What stays.** Every view, string and identifier. The Source page's board writes at once, through `SourcePageMetadataHost`, and its clear still asks first. The Add metadata dialog (`MetadataBoardDialogs`, `MetadataBoardAddForm`) stays until S15-34. Every existing caller of the four store writes sends no draft and ignores the returned rows, so each write sends what it sends today.

## 2. User stories

**US-1.** As a contributor building the Artifact sheet, I put a `MetadataBoardModel` over a `MetadataDraftHost`, and every board action changes the draft and writes nothing, so that the sheet's one Save carries the whole draft.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A draft over Provider (suggested, empty), URL `https://example.org/a` (suggested, saved) and Roll `12` (not suggested, saved); fill Provider with `Ancestry` | Provider holds `Ancestry`, `hasValue`; no store call is made |
| AC-1.2 | The same draft; clear URL, then clear Roll | URL stays as an empty suggestion; Roll leaves the rows |
| AC-1.3 | The same draft; dismiss Provider while it is empty | Provider leaves the rows; the draft's layout ends with `Provider`, `dismissed: true` |
| AC-1.4 | The same draft; dismiss URL while it holds a value, or dismiss Roll | Nothing changes (the engine's no-op) |
| AC-1.5 | The same draft; reorder the saved rows to Roll, URL | The rows read Roll, URL, Provider; the layout lists them 0, 1, 2 |
| AC-1.6 | The same draft; add Film number `F-7` with the add row | Film number is the last row, saved, not suggested |
| AC-1.7 | Dismiss Provider, then add Provider `Ancestry` with the add row | Provider is a saved row; the layout lists Provider once, `dismissed: false` |
| AC-1.8 | Fill Provider, then clear it; or reorder and reorder back | `isChanged` is false and `draft` is nil |
| AC-1.9 | AC-1.1, AC-1.6, then `updateArtifact(…, metadata: host.draft)` on `FakeStore` | The store records the draft: values Provider `Ancestry`, URL `https://example.org/a`, Roll `12`, Film number `F-7` in row order; layout the four rows 0 – 3, none dismissed |
| AC-1.10 | A new Artifact on a Source whose type suggests Provider then URL at the `artifact` level | `MetadataDraftHost.unsavedRows` answers Provider, URL: suggested, empty, in that order |

**US-2.** As a contributor, I open an add row on any board, pick a field and type a value, and its confirm adds to a draft or saves on a live board, so that S15-34 draws one inline row for both.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Open the add row; pick no field, or type only spaces | `canConfirmAddRow` is false; confirming writes nothing |
| AC-2.2 | On a draft board, pick Film number, type ` F-7 `, confirm | The draft gains Film number `F-7`; the row closes |
| AC-2.3 | On a live board, confirm a value the engine refuses (`sourcemetadata.invalid`) | The message shows on the add row; the row stays open with its field and value |
| AC-2.4 | On a live board, confirm and the write fails with an I/O error | The host shows the failure; the row stays open with no error of its own |
| AC-2.5 | The add row is open with Film number and `F-7` typed; the board's rows are reset (another window wrote) and Film number is not on the board | The row stays open with Film number and `F-7` |
| AC-2.6 | As AC-2.5, but the reset rows now hold Film number | The row stays open; its field is cleared; `F-7` stays |

**US-3.** As a contributor, when the sheet's Save is refused for one copy-details value, the board shows the message under that row, so that SA-9 needs no lookup in the sheet.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `sourcemetadata.invalid` with params `[URL's field id]` | `showRefusal` answers true; URL's row is in edit with its value; "This metadata change can't be saved." is its error |
| AC-3.2 | `sourcemetadata.invalid` with no params, or a field id not on the board, or any other code | `showRefusal` answers false and changes nothing |

**US-4.** As a contributor, a draft board clears a row with no confirm, and a live board still asks.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `askClear` on a draft board's saved Roll | No confirm is pending; Roll is cleared |
| AC-4.2 | `askClear` on the Source page's board | The confirm is pending, as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `CatalogMetadataDraft` *(new)*, the Swift mirror of the wire `MetadataDraft` | STEP-1, AC-1.9 |
| `metadata: CatalogMetadataDraft?` on `GenealogyStore`'s `createArtifact`, `updateArtifact`, `createEvidenceGraph`, `updateEvidenceGraph` in `GoStore`, `FakeStore` and `ThrowingStore`; every existing caller passes `nil` | STEP-1 |
| `createArtifact` and `updateArtifact` answer the Artifact's board after the write (the reply's `metadata`) | STEP-1 |
| `MetadataDraftHost` *(new)*: a host whose writes change only its own rows, its `draft` and `isChanged`, and `unsavedRows` for an owner nothing has saved | STEP-2, AC-1.1 – AC-1.10 |
| `confirmsClear` on `MetadataBoardHost`; `MetadataBoardDialogs.askClear` clears a draft at once | STEP-2, AC-4.1, AC-4.2 |
| The add row's state on `MetadataBoardModel`, for every host, kept across `resetDrafts` | STEP-3, AC-2.1 – AC-2.6 |
| `MetadataBoardModel.showRefusal` *(new)*: a refused value's field param placed on its row | STEP-4, AC-3.1, AC-3.2 |

| Out | Home |
| --- | --- |
| The Artifact sheet, its model, the inline Add field row's view, Escape and Return in the row, removing the Add metadata dialog and `MetadataBoardAddForm` | S15-34 |
| Patching the page's cached workspace from `updateArtifact`'s returned rows | S15-34 |
| Add mode: a new Artifact's draft on `createArtifact`, the retry on `updateArtifact`, reading `createArtifact`'s rows | S15-10 |
| The graph form's board, `getMetadataBoard` on the Swift store | S15-11 |
| Rows on the graph write replies | not planned (S15-11's scope: a graph Save reloads the workspace) |
| A Swift test of `GoStore`'s draft mapping | not planned: `swift-tests.mdc` rules out generated protobuf tests; S15-33's handler tests cover the wire, and S15-34's manual check runs the real write |
| Showing a URL in its canonical form before Save | not planned: the engine canonicalizes on Save, and the reply's rows replace the draft's (S15-34) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Host protocol | `macos/App/Features/Sources/MetadataBoardHost.swift` `MetadataBoardHost` | `entries`, `fields`, four throwing writes that answer the owner's rows and effects, `didWrite`, `clearFailure`, `failureMessage(for:)`, `showFailure`. No way to say a clear needs no confirm |
| Live host | same file, `SourcePageMetadataHost` | The Source's rows in the page's workspace; writes through the store; `didWrite` applies `.wroteMetadataBoard`. The only host in the app |
| Board model | `macos/App/Features/Sources/MetadataBoardModel.swift` `MetadataBoardModel` | `drafts`, `editingFieldID`, `savingFieldID`, `fieldError`/`fieldErrorID`; `save`, `clear`, `add(fieldID:valueText:)` (answers `MetadataBoardAddResult`), `dismissSuggestion`, `moveSaved`, `resetDrafts`, `resetIfChanged`. Trims a value before every `set` (`save`, and the dialog before `add`). No add-row state; a refusal is placed only by its own `set` (`classifySetFailure`) |
| Dialogs | `macos/App/Features/Sources/MetadataBoardDialogs.swift` `MetadataBoardDialogs` | The page's one Add dialog (`addTarget`, `addFieldID`, `addValue`, its two errors, `createFromAdd` calls `target.add`) and one clear confirm: `askClear(on:fieldID:)` (sync) always sets `pendingDelete` |
| View call site | `macos/App/Features/Sources/SourcePageMetadataView.swift` `SourcePageMetadataView` | `onClear: { model.metadataDialogs.askClear(on: model.metadata, fieldID: $0) }` |
| Board view | `macos/App/Features/Sources/MetadataBoardView.swift` `MetadataBoardView` | Draws saved rows (`board.saved`), suggestions (`board.suggested`), the empty state; reads `editingFieldID`, `fieldErrorID`, `fieldError`. Does not read `sortOrder`. Unchanged by this PR |
| Store types | `macos/App/Platform/GenealogyStore.swift` `CatalogMetadataEntry`, `MetadataOwner`, `CatalogTypeSuggestion` | No Swift draft type |
| Store writes | same file, protocol `GenealogyStore` `createArtifact`, `updateArtifact` (answer `(CatalogArtifact, CatalogWriteEffects)`), `createEvidenceGraph` (answers `(CatalogEvidenceGraph, CatalogWriteEffects)`), `updateEvidenceGraph` (answers `CatalogWriteEffects`) | No `metadata` parameter; the Artifact writes drop the reply's rows |
| Go store | `macos/App/Platform/GoStore.swift` `createArtifact`, `updateArtifact`, `createEvidenceGraph`, `updateEvidenceGraph`, `mapMetadataEntry` (private static) | Build the four requests without `metadata`; `CreateArtifactResponse.metadata` and `UpdateArtifactResponse.metadata` are not read |
| Wire | `api/proto/engine.proto` `MetadataDraft` (`values`, `layout`), `MetadataValue` (`field_id`, `value_text`), `MetadataLayoutRow` (`field_id`, `sort_order`, `dismissed`); `metadata` on `CreateArtifactRequest` (9), `UpdateArtifactRequest` (7), `CreateEvidenceGraphRequest` (7), `UpdateEvidenceGraphRequest` (7); `metadata` rows on both Artifact replies (3) | All present since S15-08 and S15-33. Generated Swift: `Provenencia_Engine_V1_MetadataDraft`, `…_MetadataValue` (`fieldID`, `valueText`), `…_MetadataLayoutRow` (`fieldID`, `sortOrder`, `dismissed`) |
| Engine rule | `core/database/sourcemetadata/sourcemetadata.go` `Apply` | The stored values become `values` (a value left out is deleted); each listed layout row is written as given; an unlisted layout row stays. A field id listed twice in `values` or in `layout` is `ErrInvalid`. Appends no layout row for a new value: the caller sends the whole order |
| Engine order | same file, `listWorkspaces` | An owner with no layout lists suggestions in suggestion order, then extras; with layout, layout order, unlisted fields after. A dismissed suggestion with no value is hidden |
| Fake | `macos/App/Platform/FakeStore.swift` `createArtifact`, `updateArtifact`, `createEvidenceGraph`, `updateEvidenceGraph`, `metadataWriteResults`, `scriptedMetadataWrite`, `createArtifactGraphIDs`, `updateArtifactGraphIDs`, `updateArtifactError`, `listTypeSuggestions`, `suggestionsByLevel` | Records graph ids per call; runs no metadata rule (S15-26 Landed, D-4) |
| Test conformer | `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore` | Throws from every method; the four signatures as above |
| App callers | `macos/App/Features/Sources/SourceArtifactsSection.swift` (`updateArtifact` ×2, `createArtifact` ×1), `macos/App/Features/Sources/SourceGraphFormModel.swift` (`createEvidenceGraph`, `updateEvidenceGraph`) | Destructure two-element tuples |
| Tests | `macos/ProvenenciaTests/MetadataBoardModelTests.swift` (`TestMetadataBoardHost`, `clearConfirmTargetsItsBoard`), `SourcePageModelTests.swift` (`askClear` at two sites; `createEvidenceGraph` once), `EvidenceGraphStoreTests.swift` (`createArtifact` ×4, `createEvidenceGraph` ×5, `updateEvidenceGraph` ×2) | Call the sync `askClear` and the two-element writes |
| Fixtures | `macos/ProvenenciaTests/CatalogFixtures.swift` `catalogMetadataField`, `catalogMetadataEntry` | Shared builders (S15-09) |

**Precedent.** S15-26 ([#798](https://github.com/mendahu/provenencia/pull/798), commit `cef790e`) added `MetadataBoardHost` and its first host, and the S15-26 plan's **Next adds** wrote this host's rules: copy that list. `ingestArtifactFile` answers a named tuple (`artifact:`, `file:`, `reused:`, `effects:`): copy that shape for the Artifact writes. S15-08's `graphIDs: [String]?` on `createArtifact` is the precedent for an optional request part that every caller states. `GoStore.mapObservationDraft` / `mapDepictionInput` are the shape for a Swift-to-proto map. `FakeStore.createArtifactGraphIDs` is the shape for recording what a write received.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `CatalogMetadataDraft` *(new)* in `GenealogyStore.swift`, beside `MetadataOwner`: `struct CatalogMetadataDraft: Sendable, Equatable { var values: [Value] = []; var layout: [LayoutRow] = [] }` with nested `struct Value: Sendable, Equatable { var fieldID: String; var valueText: String }` and `struct LayoutRow: Sendable, Equatable { var fieldID: String; var sortOrder: Int32; var dismissed: Bool }`. Doc comment: "What a form holds for an owner's metadata until its one Save: one value per field, and layout rows written as given. Mirrors the wire `MetadataDraft` (`sourcemetadata.Draft`)." | One wire mirror, named like `CatalogObservationDraft` (the store's other draft) | `MetadataDraft` (the map's working name; the store's draft types carry `Catalog`); reusing `[CatalogMetadataEntry]` (it carries display fields the wire does not take, and no dismissed rows) |
| D-2 | The four store writes gain a last parameter `metadata: CatalogMetadataDraft?` with no default. Doc: "`metadata` nil sends none (the write leaves metadata as it is, or a new owner holds none); otherwise the owner's stored values and listed layout rows become it, in the same write." `createArtifact` and `updateArtifact` answer `(artifact: CatalogArtifact, metadata: [CatalogMetadataEntry], effects: CatalogWriteEffects)`: the Artifact's board after the write, in display order. The graph writes keep their answers. Every existing caller passes `metadata: nil` and binds the rows to `_` | The rule "Metadata saves with its owner's write" names these four writes; every caller says whether it sends a draft, as `graphIDs` does; the Artifact replies already carry the board (S15-33 D-2) | A protocol extension with the old signatures (two ways to call one write); a result struct (a fourth type for one tuple; `ingestArtifactFile` already answers a named tuple) |
| D-3 | `GoStore`: `private static func mapMetadataDraft(_ d: CatalogMetadataDraft) -> Provenencia_Engine_V1_MetadataDraft` *(new)*, beside `mapMetadataEntry`, mapping each value and layout row field by field. Each of the four writes sets `req.metadata = Self.mapMetadataDraft(metadata)` only when `metadata` is non-nil (a message field has presence: unset is "absent"). The Artifact writes answer `resp.metadata.map(Self.mapMetadataEntry)` | Presence is what the engine reads; one map for four writes | Always setting it (an empty draft clears every value: S15-07's `Apply`) |
| D-4 | `FakeStore` records each call's draft in `createArtifactMetadata`, `updateArtifactMetadata`, `createEvidenceGraphMetadata`, `updateEvidenceGraphMetadata` *(new, each `[CatalogMetadataDraft?]`)*, appended where the method appends its graph ids or counts its call (before any scripted error). It applies no draft. `scriptedMetadataWrite(_:)` splits: `scriptedMetadataRows(_ owner:) -> [CatalogMetadataEntry]` *(new)* stores `metadataWriteResults[owner]` when scripted and answers `metadataByOwner[owner] ?? []`; `scriptedMetadataWrite` calls it and keeps its activity bump and effects. `createArtifact` and `updateArtifact` answer `scriptedMetadataRows(MetadataOwner(level: .artifact, id:))` | `swift-tests.mdc`: the fake seeds and scripts, and runs no engine rule (S15-26 D-4 as landed) | Applying the draft in the fake (a second copy of `Apply`) |
| D-5 | `MetadataDraftHost` *(new)*, `@MainActor final class` in `macos/App/Features/Sources/MetadataDraftHost.swift` *(new)*, conforms to `MetadataBoardHost`. It holds no store and no context: `init(entries: [CatalogMetadataEntry], fields: [CatalogMetadataField])` keeps `entries` as the rows it opened with. Its four writes are declared without `throws` and answer `(entries, .empty)` after changing `entries` per D-6. `didWrite`, `clearFailure` and `showFailure` do nothing; `failureMessage(for:)` answers `L10n.Errors.message(for: error)` (no write of its throws, so the board never calls them) | A draft writes nothing until its owner's Save (rule); a host with no store cannot write by mistake | A `SourcePageMetadataHost` mode flag (one class with two write paths); a host per owner kind (the rules are the same for an Artifact and a graph) |
| D-6 | The draft host's writes, each on its own `entries` (S15-26 **Next adds**). **set**(f, v): a row for f takes `valueText = v`, `hasValue = true`, and `openHref` back from the row it opened with when that row held exactly `v`, else `""`; with no row for f, a field from `fields` with id f is appended as `CatalogMetadataEntry(field:, valueText: v, hasValue: true, suggested: false, sortOrder: Int32(entries.count))`; an id in neither is a no-op. f leaves the dismissed list. The host does not trim: every board path trims before `set` (`save`, `confirmAddRow`, the dialog). **clear**(f): a row that is `suggested` empties (`valueText ""`, `hasValue false`, `openHref ""`); any other row leaves. **dismissSuggestion**(f): only an empty suggested row leaves and f is appended to the dismissed list; any other call is a no-op, as `sourcemetadata.DismissSuggestion` is. **reorder**(ids): the rows in `ids` order, ids not on the board skipped, then the rows `ids` left out in their order; every row's `sortOrder` becomes its index, as `FakeStore.reorderMetadata` does | The engine's rules for the rows it answers, kept in the one client place that has nothing stored to ask (S15-26 plan, reviewed) | Asking the engine per action (a write per action is what a draft avoids) |
| D-7 | `MetadataDraftHost.isChanged` *(new)*: the rows' `(field.id, hasValue, valueText)` in order differ from the opened rows', or the dismissed list is not empty. `MetadataDraftHost.draft: CatalogMetadataDraft?` *(new)*: nil when not `isChanged`; otherwise `values` = each row with `hasValue`, in row order; `layout` = every row in order with `sortOrder` its index and `dismissed: false`, then each dismissed id with `sortOrder` `entries.count + k` and `dismissed: true` | SA-8: a value typed and put back is unchanged; `Apply` appends no layout row for a new value, so the draft sends the whole order (S15-07 D-7); a nil draft leaves a never-touched owner with no layout | Sending a draft on every Save (writes layout rows the researcher never set, and a revision for nothing); comparing whole entries (`sortOrder` and `openHref` change without the researcher changing anything) |
| D-8 | `static func unsavedRows(store: any GenealogyStore, projectDir: String, typeID: String, level: Level) async throws -> [CatalogMetadataEntry]` *(new)* on `MetadataDraftHost`: an empty `typeID` answers `[]` with no call; otherwise `listTypeSuggestions` mapped in its order to `CatalogMetadataEntry(field: s.field, valueText: "", hasValue: false, suggested: true, sortOrder: Int32(s.sortOrder))`. A one-shot read when a form opens, not a `CatalogQueryKey` | What `ListWorkspace` answers for an owner with no values and no layout; the `typeSuggestions` key is the Source level's, kept warm for the type page | A new RPC for an owner that does not exist yet; a session key (a form's rows live as long as the form) |
| D-9 | `MetadataBoardHost` gains `var confirmsClear: Bool { get }` *(new)*: `SourcePageMetadataHost` true, `MetadataDraftHost` false. `MetadataBoardDialogs.askClear(on:fieldID:)` becomes `async`: with the same guard as today (a row with a value), a board whose host confirms sets `pendingDelete`; any other clears at once with `_ = await board.clear(fieldID:)`. `SourcePageMetadataView`'s `onClear` wraps it in `Task { await … }` | Working rule "A board adds a field in an inline row": a draft clears with no confirm; one entry point for every board's clear keeps the rule in one place | The sheet's view choosing per board (the rule would live in a view); a second `askClear` |
| D-10 | The add row lives on `MetadataBoardModel`: `struct MetadataBoardAddRow: Equatable { var fieldID = ""; var value = ""; var error: String? }` *(new)*, `var addRow: MetadataBoardAddRow?` (nil is closed), `private(set) var isSavingAddRow = false`, `var canConfirmAddRow: Bool` (open, not saving, a field, a value with non-whitespace text), `openAddRow()` (opens an empty row; keeps a row already open), `closeAddRow()` (does nothing while saving), `confirmAddRow() async` (returns unless `canConfirmAddRow`; clears the row's error; calls `add(fieldID:valueText:)` with the trimmed value; `.saved` closes the row, `.invalid(m)` sets its error, `.failed` leaves it open). `resetDrafts` keeps `addRow`; when its `fieldID` is now a row on the board, it sets `fieldID` to `""` and keeps `value`. The field picker reads `fieldComboOptions`, as the dialog does | SA-13's row on every board, with one write path (`add`) shared with the dialog until S15-34 removes the dialog; ✓ closes the row, as the dialog closes on save | Moving the dialog onto the row's state now (S15-34 deletes the dialog; two edits of the same code); closing the row on reset (the researcher loses typing to another window's write) |
| D-11 | `MetadataBoardModel.showRefusal(_ error: Error) -> Bool` *(new)*: when `error` is a `CoreInvokeError` with `engineCode == .sourceMetadataInvalid` and its first param is the id of a row with a value, it sets `editingFieldID` to it, `drafts[id]` to the row's `valueText`, `fieldErrorID` to it and `fieldError` to `L10n.Errors.message(for: error)`, and answers true; otherwise it changes nothing and answers false. Params are read with `if case .coded(_, _, _, let params) = error` | S15-33 D-6: the param is the field's wire id, never copy; the board owns where a row's error shows; the sheet calls it once on a failed Save and shows its top Callout when false | The sheet matching ids (a second place that knows row errors); opening no edit (the saved row's editor draws the error, as AC-1.3 of S15-26 does on a live board) |
| D-12 | Tests: `macos/ProvenenciaTests/MetadataDraftHostTests.swift` *(new)* drives `MetadataBoardModel` over `MetadataDraftHost`, and a `FakeStore` for D-8 and AC-1.9. `MetadataBoardModelTests.swift` gains the add-row and refusal tests on `TestMetadataBoardHost` (gains `var confirmsClear = true`) and awaits `askClear`. `SourcePageModelTests` and `EvidenceGraphStoreTests` change only for the new signatures | The map's Check: "Swift tests drive a draft board against FakeStore"; the board's contract is tested against a test host (S15-26 D-12) | A test of `FakeStore`'s own recording (tests test infrastructure, S15-26 review M-10) |

**Rules this plan bends.** None. `macos-client.mdc`'s fixture-table line covers Swift mirrors of a Go text rule pinned against one Go function (`FieldSlug`, `TextMatch`); D-6 mirrors where the engine places a row after a write, which has no single Go function to pin, and the S15-26 plan (reviewed) put it in the draft host as "the one client place this rule lives". The values a draft sends are checked by `Apply` on Save.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CatalogMetadataDraft` *(new)* | type | `CatalogObservationDraft`; `[CatalogMetadataEntry]` | No Swift type carries values plus dismissed layout rows; named after the store's other draft |
| `metadata:` on four store writes; rows on two answers | param / answer | `graphIDs: [String]?`; `ingestArtifactFile`'s named tuple | Extends the existing writes; no new method |
| `GoStore.mapMetadataDraft` *(new)* | func | `mapObservationDraft`, `mapDepictionInput` | Same shape, one more wire type |
| `MetadataDraftHost` *(new)* | type / file | `SourcePageMetadataHost`; `TestMetadataBoardHost` (tests) | The live host writes through the store; a draft must not. The protocol is the shared part |
| `MetadataDraftHost.unsavedRows` *(new)* | func | `CatalogQueryRegistry` `.typeSuggestions` (Source level, a session key) | That key is the type page's, at one level; a form needs one read at another level |
| `confirmsClear` *(new)* | protocol property | — | Searched `confirm` in `Features/Sources`: no per-board switch exists |
| `MetadataBoardAddRow`, `addRow`, `openAddRow`, `closeAddRow`, `confirmAddRow`, `canConfirmAddRow`, `isSavingAddRow` *(new)* | type / state | `MetadataBoardDialogs`' add state (`addFieldID`, `addValue`, `addValueError`, `createFromAdd`) | The dialog's state is page-level and leaves with the dialog in S15-34; the row is per board (SA-13). Both write through `MetadataBoardModel.add` |
| `showRefusal` *(new)* | func | `MetadataBoardModel.classifySetFailure` (a refusal of the board's own `set`) | That path places a refusal of a write the board made; this places one from the owner's write, by the param |
| `scriptedMetadataRows` *(new)* | fake helper | `FakeStore.scriptedMetadataWrite` | Split out of it so the Artifact writes answer rows the same way |
| `createArtifactMetadata`, `updateArtifactMetadata`, `createEvidenceGraphMetadata`, `updateEvidenceGraphMetadata` *(new)* | fake state | `createArtifactGraphIDs`, `updateArtifactGraphIDs` | Same recording shape |
| `MetadataDraftHostTests` *(new)* | test file | `MetadataBoardModelTests` | That file tests the board against a test host; this one tests a real host |

Searched: `MetadataDraft` in `macos/` (only generated), `Draft` types in `Platform/`, `askClear`, `confirm` in `Features/Sources`, `listTypeSuggestions(` callers, `metadataWriteResults`, `GraphIDs` in `FakeStore`. No new L10n key, no kit component.

**Builds on.** S15-26's `MetadataBoardHost` (a second conformer, as its plan named), S15-33's wire fields, `MetadataBoardModel.add` (the add row and the dialog share it), `L10n.Errors.message(for:)` for the refusal copy.
**Deletes.** None — nothing becomes dead: the dialog stays until S15-34, and every changed signature keeps its callers. `FakeStore.scriptedMetadataWrite` keeps its callers and delegates its row step.
**Next adds.** S15-34: the sheet's model holds a `MetadataDraftHost` over `CatalogSourceWorkspace.metadataRows(for:)` and a `MetadataBoardModel` on it; Save sends `updateArtifact(…, metadata: host.draft)`, applies `.wroteMetadataBoard` from the answer's rows, and on `sourcemetadata.invalid` calls `board.showRefusal`; its dirty check reads `host.isChanged`; the inline row binds `addRow`; the sheet's board passes `onClear` through `askClear`. S15-10: add mode builds the host from `unsavedRows(…, level: .artifact)` and sends `createArtifact(…, metadata: host.draft)`. S15-11: the graph form builds it from `getMetadataBoard` (edit) or `unsavedRows(…, level: .graph)` (create).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A board whose writes never throw and whose `didWrite` applies nothing | `MetadataBoardModel` | `save`, `clear`, `dismissSuggestion`, `moveSaved` reach `host.entries = …` then `didWrite`; no failure branch runs; `moveSaved`'s optimistic set is kept (T-1, T-3) |
| An open add row across `resetDrafts` | `MetadataBoardModel.resetDrafts` | Today it touches only edit state and buffers; D-10 keeps the row and clears a picked field the reset put on the board (T-9) |
| A picked field that left the picker | `confirmAddRow` | Cleared on reset (D-10); with none picked, `canConfirmAddRow` is false (T-8) |
| A refusal for a field not on the board (deleted meanwhile, or a create's id the board never had) | `showRefusal` | Answers false; the caller shows its general failure (T-10) |
| `askClear` on a draft | `MetadataBoardDialogs` | No confirm, no `pendingDelete`; the board's `clear` runs (T-5) |

### Scale

At the map's Scale (a Source with 60 Artifacts, 63 seeded fields): the draft host's actions are in memory over one owner's rows (at most the 63 fields), each linear in the rows. No action writes or reloads. `unsavedRows` is one `listTypeSuggestions` call per form opened in add mode, not per Artifact. No caller sends a draft or reads the answered rows in this PR, so no write changes its cost; S15-34 makes one Save one write (S15-33).

### Edges

- **Assumptions.** `Apply` refuses a field id listed twice in `values` or `layout`: D-6 removes a re-added field from the dismissed list, and `set` never appends a field already on the board (T-4). `listWorkspaces` sorts by layout and puts unlisted fields after; the draft lists every visible row, so after Save the engine's order is the draft's (T-1). The board never reads `sortOrder` (`MetadataBoardView`, `MetadataBoardModel`), so D-6's renumbering changes no drawing.
- **Neighbors.** A suggested row cleared stays empty; a non-suggested row cleared leaves (T-2). Dismiss on a valued row, or on a non-suggested row, is a no-op, as the engine's (T-3). A value set back to its opened text is unchanged and gets its link back (T-6). An add of a field a stored layout had dismissed sends `dismissed: false` for it: after Save the dismiss is lifted, so a later clear leaves it as an empty suggestion. This is the researcher choosing the field again; a live board's `Set` keeps the stored flag, so the two differ only after a later clear (T-4 pins the draft's layout).
- **Removed inputs.** A layout row the draft does not list stays stored (`Apply`): a suggestion dismissed before the form opened stays dismissed; it is not among the rows and not in the draft. A field deleted from the vocabulary while the form is open is refused on Save as a bare `ErrInvalid` (S15-33 D-3): `showRefusal` answers false (T-10).
- **Real data.** None — no rule driven by a table.
- **Untrusted size.** None — the rows come from the engine's board for one owner (bounded by the field vocabulary) and from the researcher's actions.

### Schema

None — no catalog change.

### Go core

None — S15-33 shipped the engine side.

### FFI

None — the wire is as S15-33 left it; `GoStore` sets and reads existing fields.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Draft type | — | `macos/App/Platform/GenealogyStore.swift` `CatalogMetadataDraft` | D-1 |
| Store writes | — | `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `OnboardingModelTests.swift` `ThrowingStore` | D-2 – D-4 |
| Draft host | Snowflake model | `macos/App/Features/Sources/MetadataDraftHost.swift` *(new)* | D-5 – D-8 |
| Clear switch | — | `MetadataBoardHost.swift` (`confirmsClear` on the protocol and `SourcePageMetadataHost`), `MetadataBoardDialogs.swift`, `SourcePageMetadataView.swift` | D-9 |
| Add row state, refusal | — | `MetadataBoardModel.swift` | D-10, D-11 |

Sketch of the new host:

```swift
/// A board over rows no write has saved: a form's copy details, sent with its
/// owner's one write as `draft`. Its writes change only its own rows.
@MainActor
final class MetadataDraftHost: MetadataBoardHost {
    var entries: [CatalogMetadataEntry]
    let fields: [CatalogMetadataField]
    var confirmsClear: Bool { false }
    var isChanged: Bool { … }                 // D-7
    var draft: CatalogMetadataDraft? { … }    // D-7
    init(entries: [CatalogMetadataEntry], fields: [CatalogMetadataField])
    func set(fieldID: String, valueText: String) -> ([CatalogMetadataEntry], CatalogWriteEffects)
    // clear, dismissSuggestion, reorder: D-6; didWrite, clearFailure, showFailure: no-ops
    static func unsavedRows(store: any GenealogyStore, projectDir: String, typeID: String, level: Level) async throws -> [CatalogMetadataEntry]
}
```

States: no view changes. Accessibility: none.

### Strings

None — no new copy. The refusal reuses `error.sourcemetadata.invalid` ("This metadata change can't be saved.").

## 7. Steps

**STEP-1 — The store writes carry a draft and answer the Artifact's board**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourceGraphFormModel.swift`, `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skill: `add-swift-test`.
Do: D-1 – D-4. Each existing caller passes `metadata: nil`; the three `SourceArtifactsSection` sites and the test sites bind the rows to `_` (`let (updated, _, effects) = …`). Rewrite the doc comments of the four protocol methods per D-2.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes with no Swift warning; `rg -n "metadata: nil" macos/App` lists the five app call sites.

**STEP-2 — The draft host and the clear switch**
Files: `macos/App/Features/Sources/MetadataDraftHost.swift` *(new)*, `macos/App/Features/Sources/MetadataBoardHost.swift`, `macos/App/Features/Sources/MetadataBoardDialogs.swift`, `macos/App/Features/Sources/SourcePageMetadataView.swift`, `macos/ProvenenciaTests/MetadataDraftHostTests.swift` *(new)*, `macos/ProvenenciaTests/MetadataBoardModelTests.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skill: `add-swift-test`.
Do: D-5 – D-9. `TestMetadataBoardHost` gains `var confirmsClear = true`; the four existing `askClear` calls gain `await`. Add T-1 – T-7.
Check: Xcode tests pass; `clearConfirmTargetsItsBoard` and the two `SourcePageModelTests` clear tests pass with only `await` added.

**STEP-3 — The add row**
Files: `macos/App/Features/Sources/MetadataBoardModel.swift`, `macos/ProvenenciaTests/MetadataBoardModelTests.swift`, `macos/ProvenenciaTests/MetadataDraftHostTests.swift` *(new)*, created in STEP-2.
Do: D-10. Add T-8, T-9.
Check: Xcode tests pass; `git diff cd64512 -- macos/App/Features/Sources/MetadataBoardDialogs.swift` shows only STEP-2's `askClear` change.

**STEP-4 — A refused value lands on its row**
Files: `macos/App/Features/Sources/MetadataBoardModel.swift`, `macos/ProvenenciaTests/MetadataDraftHostTests.swift` *(new)*, created in STEP-2.
Do: D-11. Add T-10.
Check: Xcode tests pass.

**STEP-5 — Docs**
Files: `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row), this plan (status, Landed note).
Do: per the Definition of done.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Let a metadata board hold a draft that its owner's write saves.").

## 8. Tests

Rows in T-1 – T-7 start from the AC-1.1 board: `provider` (Provider, suggested, empty), `url` (URL, suggested, `https://example.org/a`, `openHref` `https://example.org/a`), `roll` (Roll, not suggested, `12`); `fields` also holds `film` (Film number). Build with `catalogMetadataField` / `catalogMetadataEntry`; set `openHref` on the URL row after building it.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `MetadataDraftHostTests.swift` `boardActionsBuildTheDraft` *(new)* | Board over the host; `drafts["provider"] = "Ancestry"`, `save`; `await board.add(fieldID: "film", valueText: "F-7")`; `FakeStore().updateArtifactCalls == 0`; then `updateArtifact(…, graphIDs: nil, metadata: host.draft)` on that store → `updateArtifactMetadata == [draft]` with values `provider` Ancestry, `url`, `roll`, `film` F-7 and layout 0 – 3 undismissed | AC-1.1, AC-1.6, AC-1.9 |
| T-2 | Swift | `MetadataDraftHostTests` `clearKeepsSuggestionsAndDropsExtras` *(new)* | `board.clear("url")`, `board.clear("roll")` → `url` is suggested and empty with `openHref ""`; `roll` is gone; `draft.values` holds no `url` or `roll` | AC-1.2 |
| T-3 | Swift | `MetadataDraftHostTests` `dismissOnlyEmptySuggestions` *(new)* | `dismissSuggestion("url")` and `("roll")` → rows unchanged, `isChanged` false; `dismissSuggestion("provider")` → `provider` gone; layout last row `provider`, `sortOrder` 2, dismissed | AC-1.3, AC-1.4 |
| T-4 | Swift | `MetadataDraftHostTests` `readdingADismissedFieldListsItOnce` *(new)* | Dismiss `provider`, then `await board.add(fieldID: "provider", valueText: "Ancestry")` → `provider` a saved row; layout lists `provider` once, `dismissed: false` | AC-1.7 |
| T-5 | Swift | `MetadataDraftHostTests` `draftClearsWithoutConfirm` *(new)* | `await MetadataBoardDialogs().askClear(on: board, fieldID: "roll")` → `pendingDelete` nil; `roll` gone | AC-4.1, D-9 |
| T-6 | Swift | `MetadataDraftHostTests` `revertedDraftIsUnchanged` *(new)* | Fill `provider` then clear it → `isChanged` false, `draft` nil; set `url` to `https://example.org/b` then back to `https://example.org/a` → unchanged and `openHref` restored; `moveSaved` to Roll, URL → rows Roll, URL, Provider, layout 0, 1, 2; move back → unchanged | AC-1.5, AC-1.8 |
| T-7 | Swift | `MetadataDraftHostTests` `unsavedRowsAreTheTypesSuggestions` *(new)* | `FakeStore` with `suggestionsByLevel[.artifact]["t1"]` = Provider (0), URL (1) → two rows, suggested, empty, in that order; `typeID ""` with level `.citation` → `[]`, no throw | AC-1.10, D-8 |
| T-8 | Swift | `MetadataBoardModelTests` `addRowConfirmsOnlyWhenFilled` *(new)* and `addRowOnLiveBoard` *(new)* | Draft host: closed → `canConfirmAddRow` false; open, no field → false; field, value `"  "` → false; field `film`, ` F-7 ` → confirm, host rows end with `film` `F-7`, `addRow` nil. `TestMetadataBoardHost` with `setError = invalid`, `failureText = "Bad"` → `addRow?.error == "Bad"`, row open; `setError = boom` → `.showFailure` recorded, `addRow?.error` nil, row open | AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| T-9 | Swift | `MetadataBoardModelTests` `addRowSurvivesReset` *(new)* | Row open with `film`, `F-7`; `resetDrafts()` → unchanged; host entries gain a `film` row, `resetDrafts()` → `fieldID ""`, `value "F-7"`, row open | AC-2.5, AC-2.6 |
| T-10 | Swift | `MetadataDraftHostTests` `refusalLandsOnItsRow` *(new)* | `.coded(status: 1, code: "sourcemetadata.invalid", kind: .user, params: ["url"])` → true; `editingFieldID == "url"`, `drafts["url"] == "https://example.org/a"`, `fieldErrorID == "url"`, `fieldError == L10n.Errors.message(for: error)`. Params `[]`, params `["gone"]`, code `artifacts.invalid` → false, `editingFieldID` nil, `fieldError` nil | AC-3.1, AC-3.2 |
| T-11 | Swift | `MetadataBoardModelTests` `clearConfirmTargetsItsBoard` *(existing, `await` added)* | As today: the confirm is pending for a live board | AC-4.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests` (artifact create and edit, metadata clear), `EvidenceGraphStoreTests`, `SourceGraphFormModelTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: no name renamed or deleted; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Rules this plan bends — none
- [ ] UI matches the brief inventory — no UI change
- [ ] Version bump — none, per the header
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — n/a, not gated (S15-D14 moves with S15-25)
- [ ] Domain doc — none states this rule
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source that has metadata. Add a value with the Add metadata dialog, edit it, drag it, and clear it (the confirm appears). Each behaves as before this PR.
2. Edit an Artifact's label on its card and Save; add an Evidence graph and edit it. Each saves as before (the writes send no draft).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A three-element answer breaks a destructuring site the plan missed | STEP-1's check builds every target; the compiler lists each |
| `askClear` becoming `async` shows the Source page's confirm a turn later | The view starts it in a `Task` at the click; no state is read between |
| A draft listing a field twice is refused on Save | D-6 drops a re-added field from the dismissed list; T-4 |

**Stop and report if:**

- `CreateArtifactResponse.metadata`, `UpdateArtifactResponse.metadata` or the `metadata` request fields of §4 are missing or typed otherwise in `engine.pb.swift` at your base.
- `MetadataBoardView` or `MetadataBoardModel` reads `sortOrder` (D-6's renumbering would then change a drawing).
- A `MetadataBoardHost` conformer exists besides `SourcePageMetadataHost` and `TestMetadataBoardHost`.
- `sourcemetadata.ErrInvalid`'s refused value no longer carries the field id as `params[0]` (S15-33 D-6).
- Any test outside §8 or the signature changes of STEP-1 and STEP-2 needs editing to pass.

## 12. Open questions

None.

## 13. PR

**Title:** S15-36 — A metadata board can hold a draft
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-36-metadata-board-holds-a-draft.md`.
