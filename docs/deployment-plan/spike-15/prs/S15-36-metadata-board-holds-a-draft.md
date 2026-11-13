# S15-36 — A metadata board can hold a draft

**Status:** Planned
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
**Review:** [`S15-36-metadata-board-holds-a-draft.review.md`](S15-36-metadata-board-holds-a-draft.review.md) — Fits with changes — Ready with changes, 2 rounds; every finding Fixed or Logged; nothing To user
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
| AC-1.3 | The same draft; dismiss Provider while it is empty | Provider leaves the rows; the draft's layout ends with `Provider`, `sortOrder` 2, `dismissed: true` |
| AC-1.4 | The same draft; dismiss URL while it holds a value, or dismiss Roll | Nothing changes (the engine's no-op) |
| AC-1.5 | The same draft; drag the saved rows to Roll, URL | The rows read Roll, URL, Provider; the layout lists them 0, 1, 2 |
| AC-1.6 | The same draft; add Film number `F-7` | Film number is the last row, saved, not suggested |
| AC-1.7 | Dismiss Provider, then add Provider `Ancestry` | Provider is a saved row; the layout lists Provider once, `dismissed: false` |
| AC-1.8 | Fill Provider, then clear it; or drag Roll above URL, then back | `isChanged` is false and `draft` is nil |
| AC-1.9 | AC-1.1, AC-1.6, then `updateArtifact(…, metadata: host.draft)` on `FakeStore` | The store records the draft: values Provider `Ancestry`, URL `https://example.org/a`, Roll `12`, Film number `F-7` in row order; layout the four rows 0 – 3, none dismissed |
| AC-1.10 | A type that suggests Provider (order 0) then URL (order 1) at the `artifact` level | `MetadataDraftHost.unsavedRows(from:)` answers Provider, URL: suggested, empty, in that order |

**US-2.** As a contributor, I open an add row on any board, pick a field and type a value, and its confirm adds to a draft or saves on a live board, so that S15-34 draws one inline row for both. Until then the Add metadata dialog presents the same row.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Open the add row; pick no field, or type only spaces | `canConfirmAddRow` is false; confirming writes nothing |
| AC-2.2 | On a draft board, pick Film number, type ` F-7 `, confirm | The draft gains Film number `F-7`; the row closes |
| AC-2.3 | On a live board, confirm a value the engine refuses (`sourcemetadata.invalid`) | The message is the row's `error`; the row stays open with its field and value |
| AC-2.4 | On a live board, confirm and the write fails with an I/O error | The host shows the failure; the row stays open with no error of its own |
| AC-2.5 | The add row is open with Film number and `F-7` typed; the board's rows are reset (another window wrote) and Film number is not on the board | The row stays open with Film number and `F-7` |
| AC-2.6 | As AC-2.5, but the reset rows now hold Film number | The row stays open; its field is cleared; `F-7` stays |
| AC-2.7 | The Source page's Add metadata dialog: submit with no field; with a field and only spaces; with a value the engine refuses; with a good value | "Choose a field."; the value-required message; the refusal under the value; the dialog closes and the value is saved — as today |

**US-3.** As a contributor, when the sheet's Save is refused for one copy-details value, the board shows the message under that row, so that SA-9 needs no lookup in the sheet.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `sourcemetadata.invalid` with params `[URL's field id]`, while Roll is in edit | `showRefusal` answers true; Roll's edit is cancelled; URL's row is in edit with its value; "This metadata change can't be saved." is its error |
| AC-3.2 | `sourcemetadata.invalid` with no params, or `[""]`, or a field id not on the board, or any other code | `showRefusal` answers false and changes nothing |

**US-4.** As a contributor, a draft board clears a row with no confirm, and a live board still asks.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `requestClear` on a draft board's saved Roll | Answers false (no confirm); Roll is cleared |
| AC-4.2 | `requestClear` on the Source page's board | Answers true and writes nothing; the page then opens its confirm, as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `CatalogMetadataDraft` *(new)*, the Swift mirror of the wire `MetadataDraft` | STEP-1, AC-1.9 |
| `metadata: CatalogMetadataDraft?` on `GenealogyStore`'s `createArtifact`, `updateArtifact`, `createEvidenceGraph`, `updateEvidenceGraph` in `GoStore`, `FakeStore` and `ThrowingStore`; every existing caller passes `nil` | STEP-1 |
| `createArtifact` and `updateArtifact` answer the Artifact's board after the write (the reply's `metadata`) | STEP-1 |
| `CoreInvokeError.params` *(new)*; `SourceArtifactsSection`'s two ingest refusals read it | STEP-1 |
| One client reorder rule, `[CatalogMetadataEntry].reordered(by:)` *(new)*, used by the draft host and `FakeStore.reorderMetadata` | STEP-2, AC-1.5 |
| `MetadataDraftHost` *(new)*: a host whose writes change only its own rows, its `draft` and `isChanged`, and `unsavedRows(from:)` | STEP-2, AC-1.1 – AC-1.10 |
| `confirmsClear` on `MetadataBoardHost`; `MetadataBoardModel.requestClear` *(new)* clears a draft at once and tells a live board's page to confirm | STEP-2, AC-4.1, AC-4.2 |
| The add row's state on `MetadataBoardModel`, for every host, kept across `resetDrafts`; the Add metadata dialog presents it | STEP-3, AC-2.1 – AC-2.7 |
| `MetadataBoardModel.showRefusal` *(new)*, sharing the row-refusal check and placement with `save` | STEP-4, AC-3.1, AC-3.2 |
| The map row's Ships line and `add-workspace-place` § 6 say what a draft board does | STEP-5 |

| Out | Home |
| --- | --- |
| The Artifact sheet, its model, the inline Add field row's view (and its clearing of the row's error on edit), Escape and Return in the row, removing the Add metadata dialog's view and `MetadataBoardAddForm` | S15-34 |
| Patching the page's cached workspace from `updateArtifact`'s returned rows | S15-34 |
| Add mode: reading the type's suggestions, a new Artifact's draft on `createArtifact`, the retry on `updateArtifact`, reading `createArtifact`'s rows | S15-10 |
| The graph form's board, `getMetadataBoard` on the Swift store | S15-11 |
| Rows on the graph write replies | not planned (S15-11's scope: a graph Save reloads the workspace) |
| A Swift test of `GoStore`'s draft mapping | not planned: `swift-tests.mdc` rules out generated protobuf tests; S15-33's handler tests cover the wire, and S15-34's manual check runs the real write |
| Showing a URL in its canonical form before Save | not planned: the engine canonicalizes on Save, and the reply's rows replace the draft's (S15-34) |
| `OnboardingView`'s `case .coded(_, _, let kind, _)` read of the error kind | not planned: it reads the kind, which has no accessor; this PR adds only the one it needs |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Host protocol | `macos/App/Features/Sources/MetadataBoardHost.swift` `MetadataBoardHost` | `entries`, `fields`, four throwing writes that answer the owner's rows and effects, `didWrite`, `clearFailure`, `failureMessage(for:)`, `showFailure`. No way to say a clear needs no confirm |
| Live host | same file, `SourcePageMetadataHost` | The Source's rows in the page's workspace; writes through the store; `didWrite` applies `.wroteMetadataBoard`. The only host in the app |
| Board model | `macos/App/Features/Sources/MetadataBoardModel.swift` `MetadataBoardModel` | `drafts`, `editingFieldID`, `savingFieldID`, `fieldError`/`fieldErrorID`; `beginEdit` (cancels another row's edit, seeds the buffer, clears that row's error), `cancelEdit`, `save` (trims), `clear`, `add(fieldID:valueText:)` (does not trim; answers `MetadataBoardAddResult`), `dismissSuggestion`, `moveSaved` (sets the rows to saved rows then suggestions, then `host.reorder` with that order), `resetDrafts`, `resetIfChanged`. `classifySetFailure` decides a row refusal (host message nil → owner gone; `.sourceMetadataInvalid` → row; else host failure). No add-row state |
| Dialogs | `macos/App/Features/Sources/MetadataBoardDialogs.swift` `MetadataBoardDialogs` | The page's one Add dialog with its own state (`addTarget`, `addFieldID`, `addValue`, `addFieldError`, `addValueError`, `isSavingAdd`, `canSubmitAdd`, `addSheetIsDirty`; `createFromAdd` checks field then value, trims, calls `target.add`) and one clear confirm (`askClear`, sync, sets `pendingDelete`) |
| View call sites | `macos/App/Features/Sources/SourcePageMetadataView.swift` `onClear`; `MetadataBoardView.swift` `MetadataBoardAddForm` (binds `$dialogs.addFieldID`, `$dialogs.addValue`, reads the two errors and `fieldComboOptions`); `SourcePageView.swift` (reads `isSavingAdd`, `addSheetIsDirty`, `canSubmitAdd`, `isAdding`); `SourcePageModel.swift` (`metadataDialogs.isAdding`) | Unchanged by this PR: the dialog keeps every member they read |
| Board view | `macos/App/Features/Sources/MetadataBoardView.swift` `MetadataBoardView` | Draws `board.saved`, `board.suggested`, the empty state; reads `editingFieldID`, `fieldErrorID`, `fieldError`. Does not read `sortOrder`. Unchanged |
| Errors | `macos/App/Platform/CoreInvoke.swift` `CoreInvokeError` `code`, `engineCode` | "The one read of a coded error"; no params accessor. `SourceArtifactsSection.swift` matches `case .coded(_, let code, _, let params)` twice (create's ingest refusal, attach's ingest refusal) |
| Store types | `macos/App/Platform/GenealogyStore.swift` `CatalogMetadataEntry`, `MetadataOwner`, `CatalogTypeSuggestion` | No Swift draft type, no shared reorder |
| Store writes | same file, protocol `GenealogyStore` `createArtifact`, `updateArtifact` (answer `(CatalogArtifact, CatalogWriteEffects)`), `createEvidenceGraph` (answers `(CatalogEvidenceGraph, CatalogWriteEffects)`), `updateEvidenceGraph` (answers `CatalogWriteEffects`) | No `metadata` parameter; the Artifact writes drop the reply's rows |
| Go store | `macos/App/Platform/GoStore.swift` `createArtifact`, `updateArtifact`, `createEvidenceGraph`, `updateEvidenceGraph`, `mapMetadataEntry` (private static) | Build the four requests without `metadata`; the Artifact replies' `metadata` is not read |
| Wire | `api/proto/engine.proto` `MetadataDraft` (`values`, `layout`), `MetadataValue` (`field_id`, `value_text`), `MetadataLayoutRow` (`field_id`, `sort_order`, `dismissed`); `metadata` on `CreateArtifactRequest` (9), `UpdateArtifactRequest` (7), `CreateEvidenceGraphRequest` (7), `UpdateEvidenceGraphRequest` (7); `metadata` rows on both Artifact replies (3) | Present since S15-08 and S15-33. Generated Swift: `Provenencia_Engine_V1_MetadataDraft`, `…_MetadataValue` (`fieldID`, `valueText`), `…_MetadataLayoutRow` (`fieldID`, `sortOrder`, `dismissed`) |
| Engine rule | `core/database/sourcemetadata/sourcemetadata.go` `Apply` | The stored values become `values` (a value left out is deleted); each listed layout row is written as given; an unlisted layout row stays. A field id listed twice in `values` or `layout` is `ErrInvalid`. Appends no layout row for a new value: the caller sends the whole order |
| Engine order | same file, `listWorkspaces` | No layout: suggestions in suggestion order, then extras; with layout: layout order, unlisted fields after. A dismissed suggestion with no value is hidden |
| Fake | `macos/App/Platform/FakeStore.swift` `createArtifact`, `updateArtifact` (records, awaits `updateArtifactHold`, throws `StoreBoom.boom` for an id in no `artifactsBySource`), `createEvidenceGraph`, `updateEvidenceGraph`, `reorderMetadata` (appends the rows the ids leave out by iterating a `Dictionary`: undefined order), `metadataWriteResults`, `scriptedMetadataWrite`, `createArtifactGraphIDs`, `updateArtifactGraphIDs`, `suggestionsByLevel` | Runs no metadata rule (S15-26 Landed, D-4) |
| Test conformer | `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore` | Throws from every method; the four signatures as above |
| App callers | `macos/App/Features/Sources/SourceArtifactsSection.swift` (`updateArtifact` ×2, `createArtifact` ×1), `macos/App/Features/Sources/SourceGraphFormModel.swift` (`createEvidenceGraph`, `updateEvidenceGraph`) | Destructure two-element tuples |
| Tests | `macos/ProvenenciaTests/MetadataBoardModelTests.swift` (`TestMetadataBoardHost`, `addDialogWritesToItsTargetBoard`, `clearConfirmTargetsItsBoard`, `invalidValueStaysOnTheFieldOrDialog`), `SourcePageModelTests.swift` (dialog tests near lines 878, 931, 1114; `createEvidenceGraph` once), `EvidenceGraphStoreTests.swift` (`createArtifact` ×4, `createEvidenceGraph` ×5, `updateEvidenceGraph` ×2) | Drive the dialog through its own state and the two-element writes |
| Fixtures | `macos/ProvenenciaTests/CatalogFixtures.swift` `catalogArtifact`, `catalogMetadataField`, `catalogMetadataEntry` | Shared builders (S15-09) |
| Docs | `.agents/skills/add-workspace-place/SKILL.md` § 6 "Metadata values"; map S15-36 **Ships** | The skill says every board's host applies `.wroteMetadataBoard`; the map says the draft host "trims values, empty `openHref`" and that its starting rows come "from `listTypeSuggestions`" |

**Precedent.** S15-26 ([#798](https://github.com/mendahu/provenencia/pull/798), commit `cef790e`) added `MetadataBoardHost` and its first host, and the S15-26 plan's **Next adds** wrote this host's rules: copy that list. `ingestArtifactFile` answers a named tuple (`artifact:`, `file:`, `reused:`, `effects:`): copy that shape for the Artifact writes. S15-08's `graphIDs: [String]?` on `createArtifact` is the precedent for an optional request part that every caller states. `GoStore.mapObservationDraft` / `mapDepictionInput` are the shape for a Swift-to-proto map. `FakeStore.createArtifactGraphIDs` is the shape for recording what a write received. `CoreInvokeError.code` / `engineCode` are the shape for `params`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `CatalogMetadataDraft` *(new)* in `GenealogyStore.swift`, beside `MetadataOwner`: `struct CatalogMetadataDraft: Sendable, Equatable { var values: [Value] = []; var layout: [LayoutRow] = [] }` with nested `struct Value: Sendable, Equatable { var fieldID: String; var valueText: String }` and `struct LayoutRow: Sendable, Equatable { var fieldID: String; var sortOrder: Int32; var dismissed: Bool }`. Doc: "What a form holds for an owner's metadata until its one Save: one value per field, and layout rows written as given. Mirrors the wire `MetadataDraft` (`sourcemetadata.Draft`)." | One wire mirror, named like `CatalogObservationDraft` (the store's other draft) | `MetadataDraft` (the map's working name; the store's draft types carry `Catalog`); reusing `[CatalogMetadataEntry]` (display fields the wire does not take, and no dismissed rows) |
| D-2 | The four store writes gain a last parameter `metadata: CatalogMetadataDraft?` with no default. Doc: "`metadata` nil sends none (the write leaves metadata as it is, or a new owner holds none); otherwise the owner's stored values and listed layout rows become it, in the same write." `createArtifact` and `updateArtifact` answer `(artifact: CatalogArtifact, metadata: [CatalogMetadataEntry], effects: CatalogWriteEffects)`: the Artifact's board after the write, in display order. The graph writes keep their answers. Every existing caller passes `metadata: nil` and binds the rows to `_` | The rule "Metadata saves with its owner's write" names these four writes; every caller says whether it sends a draft, as `graphIDs` does; the Artifact replies already carry the board (S15-33 D-2) | A protocol extension with the old signatures (two ways to call one write); a result struct (`ingestArtifactFile` already answers a named tuple) |
| D-3 | `GoStore`: `private static func mapMetadataDraft(_ d: CatalogMetadataDraft) -> Provenencia_Engine_V1_MetadataDraft` *(new)*, beside `mapMetadataEntry`, field by field. Each of the four writes sets `req.metadata = Self.mapMetadataDraft(metadata)` only when `metadata` is non-nil (unset is "absent"). The Artifact writes answer `resp.metadata.map(Self.mapMetadataEntry)` | Presence is what the engine reads; one map for four writes | Always setting it (an empty draft clears every value) |
| D-4 | `FakeStore` records each call's draft in `createArtifactMetadata`, `updateArtifactMetadata`, `createEvidenceGraphMetadata`, `updateEvidenceGraphMetadata` *(new, each `[CatalogMetadataDraft?]`)*, appended where the method appends its graph ids or counts its call (before any hold or scripted error). It applies no draft. `scriptedMetadataWrite(_:)` splits: `scriptedMetadataRows(_ owner:) -> [CatalogMetadataEntry]` *(new)* stores `metadataWriteResults[owner]` when scripted and answers `metadataByOwner[owner] ?? []`; `scriptedMetadataWrite` calls it and keeps its activity bump and effects. `createArtifact` and `updateArtifact` answer `scriptedMetadataRows(MetadataOwner(level: .artifact, id:))`. `reorderMetadata` answers `current.reordered(by: fieldIDs)` (D-6) | `swift-tests.mdc`: the fake seeds and scripts, and runs no engine rule (S15-26 D-4 as landed) | Applying the draft in the fake (a second copy of `Apply`) |
| D-5 | `CoreInvokeError.params: [String]` *(new)*, beside `code`: the coded error's params, `[]` for `.failed`; its doc says a caller reads params through it, not the case. `SourceArtifactsSection`'s two `case .coded(_, let code, _, let params)` matches become `if let coded = error as? CoreInvokeError, let code = coded.code, let callout = L10n.Errors.ingestCallout(code: code, params: coded.params)` | `code`'s doc: "the one read of a coded error"; D-13 is the third reader of params (review M-2) | A third case match |
| D-6 | `extension Array where Element == CatalogMetadataEntry { func reordered(by fieldIDs: [String]) -> [CatalogMetadataEntry] }` *(new)* in `GenealogyStore.swift`, beside `CatalogMetadataEntry`: the rows in `fieldIDs` order (ids with no row, and repeats of an id already placed, skipped), then the rows the ids left out in their current order; every row's `sortOrder` becomes its index. `FakeStore.reorderMetadata` and the draft host both call it | One client reorder rule, with a stable left-out order (the fake's today is undefined; review M-4) | Two copies |
| D-7 | `MetadataDraftHost` *(new)*, `@MainActor final class` in `macos/App/Features/Sources/MetadataDraftHost.swift` *(new)*, conforms to `MetadataBoardHost`. It holds no store and no context: `init(entries: [CatalogMetadataEntry], fields: [CatalogMetadataField])` keeps `entries` as the rows it opened with. Its four writes are declared without `throws` and answer `(entries, .empty)` after changing `entries`. **set**(f, v): a row for f takes `valueText = v`, `hasValue = true`, and `openHref` back from the opened row when that row held exactly `v`, else `""`; with no row for f, a field from `fields` with id f is appended as `CatalogMetadataEntry(field:, valueText: v, hasValue: true, suggested: false, sortOrder: Int32(entries.count))`; an id in neither is a no-op; f leaves the dismissed list. The host does not trim: every board path trims before `set` (`save`, `confirmAddRow`). **clear**(f): a `suggested` row empties (`valueText ""`, `hasValue false`, `openHref ""`); any other row leaves; an id with no row is a no-op. **dismissSuggestion**(f): only an empty suggested row leaves, and f is appended to the dismissed list; any other call is a no-op, as `sourcemetadata.DismissSuggestion` is. **reorder**(ids): `entries.reordered(by: ids)`. `didWrite`, `clearFailure`, `showFailure` do nothing; `failureMessage(for:)` answers `L10n.Errors.message(for: error)` | A draft writes nothing until its owner's Save (rule); a host with no store cannot write by mistake; the engine's row rules live in the one client place with nothing stored to ask (S15-26 plan **Next adds**, reviewed) | A mode flag on `SourcePageMetadataHost` (one class, two write paths); asking the engine per action |
| D-8 | `MetadataDraftHost.isChanged` *(new)* compares what the board shows, not the raw row order (`moveSaved` rebuilds the rows as saved rows then suggestions; review X-1): true when the saved rows' `(field.id, valueText)` in row order differ from the opened rows' saved rows, or the empty suggestions' ids in row order differ from the opened rows' empty suggestions', or the dismissed list is not empty. `MetadataDraftHost.draft: CatalogMetadataDraft?` *(new)*: nil when not `isChanged`; otherwise `values` = each row with `hasValue`, in row order; `layout` = every row in row order with `sortOrder` its index and `dismissed: false`, then each dismissed id with `sortOrder` `entries.count + k` (`k` from 0, in dismissed order) and `dismissed: true` | SA-8: a value typed and put back, or a drag put back, is unchanged; `Apply` appends no layout row for a new value, so the draft sends the whole order (S15-07 D-7); a nil draft leaves a never-touched owner with no layout | Sending a draft on every Save (layout rows the researcher never set, a revision for nothing); comparing whole entries or raw order |
| D-9 | `static func unsavedRows(from suggestions: [CatalogTypeSuggestion]) -> [CatalogMetadataEntry]` *(new)* on `MetadataDraftHost`, pure: each suggestion in its order as `CatalogMetadataEntry(field: s.field, valueText: "", hasValue: false, suggested: true, sortOrder: Int32(s.sortOrder))`. The form that opens a draft for a new owner reads the suggestions (S15-10, S15-11) | What `ListWorkspace` answers for an owner with no values and no layout; the host stays a pure in-memory model (review M-6) | A store read on the host; a session key (a form's rows live as long as the form) |
| D-10 | `MetadataBoardHost` gains `var confirmsClear: Bool { get }` *(new)*: `SourcePageMetadataHost` true, `MetadataDraftHost` false. `MetadataBoardModel.requestClear(fieldID:) async -> Bool` *(new)*: with no row holding a value for `fieldID`, answers false; with a host that confirms, answers true and writes nothing (the caller opens its confirm); otherwise runs `clear(fieldID:)` and answers false. `SourcePageMetadataView`'s `onClear` becomes `{ id in Task { if await model.metadata.requestClear(fieldID: id) { model.metadataDialogs.askClear(on: model.metadata, fieldID: id) } } }`. `askClear` stays synchronous and is only the confirm | Working rule "A board adds a field in an inline row": a draft clears with no confirm. The rule sits on the board every form has, and the page-level dialog stays the confirm only (review M-5) | `askClear` going `async` (the sheet and the graph form would hold a dialog object just to reach a clear that never confirms); the view choosing per host |
| D-11 | The add row lives on `MetadataBoardModel`: `struct MetadataBoardAddRow: Equatable { var fieldID = ""; var value = ""; var error: String? }` *(new)*, `var addRow: MetadataBoardAddRow?` (nil is closed), `private(set) var isSavingAddRow = false`, `var canConfirmAddRow: Bool` (open, not saving, a field, a value with non-whitespace text), `openAddRow()` (a closed row opens empty; an open row keeps its field, value and error), `closeAddRow()` (does nothing while `isSavingAddRow`), `confirmAddRow() async` (returns unless `canConfirmAddRow`; clears the row's `error`; sets `isSavingAddRow`, reset by `defer` on every outcome; calls `add(fieldID:valueText:)` with the trimmed value; then `.saved` sets `addRow = nil`, `.invalid(m)` sets `addRow?.error = m`, `.failed` leaves the row as it is). `resetDrafts` keeps `addRow`; when its `fieldID` is now a row on the board, it sets `fieldID = ""` and keeps `value` and `error`. The picker reads `fieldComboOptions`. The row's `error` clears only on a new row and at the next confirm; the view that clears it on edit is S15-34's | SA-13's row on every board, one write path (`add`) | Closing the row on reset (the researcher loses typing to another window's write) |
| D-12 | `MetadataBoardDialogs`' Add becomes a presenter of its target's add row (review M-3). It keeps `addTarget`, `isAdding`, `addFieldError` (its "Choose a field." is the dialog's own) and `openAdd(on:)` (now: when `addTarget` is another board, `addTarget?.closeAddRow()` first; then `board.closeAddRow()` and `board.openAddRow()`, so the dialog always opens on an empty row as today; then `addTarget = board`, `addFieldError = nil`). A reload that puts the dialog's picked field on the board clears the dialog's field, as AC-2.6 does for the row; this is the one dialog behavior that changes, and only while another window adds that field. These become computed over `addTarget?.addRow` and `addTarget`: `addFieldID` and `addValue` (get from the row, `""` when closed; set writes the row when open), `addValueError` (the row's `error`; set writes it), `isSavingAdd` (`addTarget?.isSavingAddRow ?? false`). `canSubmitAdd` becomes `addTarget?.canConfirmAddRow ?? false` (one confirm gate; review M-9); `addSheetIsDirty` keeps its body over the computed members. `cancelAdd()`: does nothing while saving; else `addTarget?.closeAddRow()`, `addTarget = nil`. `createFromAdd()`: keeps its two required checks and messages (field, then trimmed value) as today, then `await target.confirmAddRow()`, then `addTarget = nil` when `target.addRow == nil`. Its stored `addFieldID`, `addValue`, `addValueError`, `isSavingAdd` go. Every view and test reading these names compiles unchanged | One add mechanism from now on, exercised by the live page before S15-34 draws the row; S15-34 only deletes the dialog's view | Two copies of the add rule until S15-34 |
| D-13 | Row refusals share one check and one placement (review M-1). `MetadataBoardModel` gains `private func rowRefusalMessage(_ error: Error) -> String?` *(new)*: `host.failureMessage(for:)` when `(error as? CoreInvokeError)?.engineCode == .sourceMetadataInvalid`, else nil; and `private func placeRowError(_ message: String, fieldID: String)` *(new)*: sets `fieldError` and `fieldErrorID`. `classifySetFailure` keeps its three outcomes and calls `host.failureMessage` once: nil is `.ownerGone`; with the code `.sourceMetadataInvalid` the message is `.invalid`; else it shows the message. `rowRefusalMessage` serves `showRefusal` only. `save`'s `.invalid` branch calls `placeRowError`. `showRefusal(_ error: Error) -> Bool` *(new)*: `fieldID = (error as? CoreInvokeError)?.params.first`; answers false unless it is a row with a value and `rowRefusalMessage(error)` is non-nil; then `beginEdit(fieldID:)` (which cancels another row's edit and seeds the buffer), `placeRowError(message, fieldID:)`, true. An empty id matches no row | One place decides "a row refusal and its words"; `beginEdit` already owns opening a row | `showRefusal` reading `L10n` and setting the fields itself (a second path that skips the host's message and owner-gone rule) |
| D-14 | Tests: `macos/ProvenenciaTests/MetadataDraftHostTests.swift` *(new)* drives `MetadataBoardModel` over `MetadataDraftHost`, and a `FakeStore` for AC-1.9. `MetadataBoardModelTests.swift` gains the live add-row, `requestClear` and refusal-sharing tests on `TestMetadataBoardHost` (gains `var confirmsClear = true`). `SourcePageModelTests`' dialog tests and `MetadataBoardModelTests`' dialog tests pass unchanged (AC-2.7). `EvidenceGraphStoreTests` change only for the signatures | The map's Check: "Swift tests drive a draft board against FakeStore"; the board's contract against a test host (S15-26 D-12) | A test of `FakeStore`'s own recording (S15-26 review M-10) |

**Rules this plan bends.** None. `macos-client.mdc`'s fixture-table line covers Swift mirrors of a Go text rule pinned against one Go function (`FieldSlug`, `TextMatch`); D-7 mirrors where the engine places a row after a write, which has no single Go function to pin, and the S15-26 plan (reviewed) put it in the draft host as "the one client place this rule lives". The values a draft sends are checked by `Apply` on Save.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CatalogMetadataDraft` *(new)* | type | `CatalogObservationDraft`; `[CatalogMetadataEntry]` | No Swift type carries values plus dismissed layout rows |
| `metadata:` on four store writes; rows on two answers | param / answer | `graphIDs: [String]?`; `ingestArtifactFile`'s named tuple | Extends the existing writes; no new method |
| `GoStore.mapMetadataDraft` *(new)* | func | `mapObservationDraft`, `mapDepictionInput` | Same shape, one more wire type |
| `CoreInvokeError.params` *(new)* | accessor | `CoreInvokeError.code`, `engineCode` | Same shape; replaces two case matches |
| `reordered(by:)` *(new)* | func | `FakeStore.reorderMetadata`'s body | Moved out of the fake so the draft host shares it |
| `MetadataDraftHost` *(new)* | type / file | `SourcePageMetadataHost`; `TestMetadataBoardHost` (tests) | The live host writes through the store; a draft must not. The protocol is the shared part |
| `MetadataDraftHost.unsavedRows(from:)` *(new)* | func | `CatalogQueryRegistry` `.typeSuggestions` (a read, Source level) | A pure mapping; the read stays with the caller |
| `confirmsClear`, `requestClear` *(new)* | property / func | `MetadataBoardDialogs.askClear` | `askClear` stays the confirm; the switch lives on the board every form has |
| `MetadataBoardAddRow`, `addRow`, `openAddRow`, `closeAddRow`, `confirmAddRow`, `canConfirmAddRow`, `isSavingAddRow` *(new)* | type / state | `MetadataBoardDialogs`' add state | Replaces it: the dialog presents the row (D-12) |
| `rowRefusalMessage`, `placeRowError`, `showRefusal` *(new)* | func | `classifySetFailure`, `save`'s error lines | Split out of them and shared |
| `scriptedMetadataRows` *(new)* | fake helper | `FakeStore.scriptedMetadataWrite` | Split out so the Artifact writes answer rows the same way |
| `createArtifactMetadata`, `updateArtifactMetadata`, `createEvidenceGraphMetadata`, `updateEvidenceGraphMetadata` *(new)* | fake state | `createArtifactGraphIDs`, `updateArtifactGraphIDs` | Same recording shape |
| `MetadataDraftHostTests` *(new)* | test file | `MetadataBoardModelTests` | That file tests the board against a test host; this one a real host |

Searched: `MetadataDraft` in `macos/` (only generated), `Draft` types in `Platform/`, `askClear`, `confirm` in `Features/Sources`, `listTypeSuggestions(` callers, `metadataWriteResults`, `GraphIDs` in `FakeStore`, `case .coded(` in `macos/App`, `reorder` in `Platform/`. No new L10n key, no kit component.

**Builds on.** S15-26's `MetadataBoardHost` (a second conformer, as its plan named), S15-33's wire fields, `MetadataBoardModel.add` and `beginEdit`, `L10n.Errors.message(for:)`.
**Deletes.** `MetadataBoardDialogs`' stored `addFieldID`, `addValue`, `addValueError`, `isSavingAdd` and its own trim-and-`add` call (D-12); `FakeStore.reorderMetadata`'s inline reorder loop (D-6); `SourceArtifactsSection`'s two `case .coded` matches (D-5). Doc comments made stale are reworded: `MetadataBoardAddResult` ("what a board's Add reports to the dialog that asked" → "to the add row that asked"), `MetadataBoardModel.host` ("Read by `MetadataBoardDialogs` for the Add dialog and clear confirm" → "Read by `MetadataBoardDialogs` for the clear confirm and the Add dialog over the add row"), `add` ("the Add dialog's write" → "the add row's write"), `fieldError` ("after `sourcemetadata.invalid`, from a row's own save or the owner's write (`showRefusal`); cleared on next edit").
**Next adds.** S15-34: the sheet's model holds a `MetadataDraftHost` over `CatalogSourceWorkspace.metadataRows(for:)` and a `MetadataBoardModel` on it; Save sends `updateArtifact(…, metadata: host.draft)`, applies `.wroteMetadataBoard` from the answer's rows, and on a failure calls `board.showRefusal` and shows its top Callout when false; its dirty check reads `host.isChanged`; the inline row binds `addRow`; `onClear` calls `requestClear`; it deletes the dialog's view, `MetadataBoardAddForm` and the dialog's add members. S15-10: add mode reads `listTypeSuggestions(…, level: .artifact)`, builds the host from `unsavedRows(from:)` and sends `createArtifact(…, metadata: host.draft)`. S15-11: the graph form builds it from `getMetadataBoard` (edit) or `unsavedRows(from:)` (create).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A board whose writes never throw and whose `didWrite` applies nothing | `MetadataBoardModel` | `save`, `clear`, `dismissSuggestion`, `moveSaved` set `host.entries` then `didWrite`; no failure branch runs; `moveSaved`'s optimistic set is kept (T-1, T-3, T-6) |
| The Source page's Add dialog over the board's add row | `MetadataBoardDialogs`, `MetadataBoardAddForm`, `SourcePageView` | Every member they read keeps its name and meaning (D-12); the existing dialog tests pass unchanged (T-12) |
| An open add row across `resetDrafts` | `MetadataBoardModel.resetDrafts` | Today it touches only edit state and buffers; D-11 keeps the row and clears a picked field the reset put on the board (T-9) |
| A refusal for a field not on the board (deleted meanwhile, or never on it) | `showRefusal` | Answers false; the caller shows its general failure (T-10) |
| `requestClear` on a draft | `MetadataBoardModel` | No confirm; the board's `clear` runs (T-5) |

### Scale

At the map's Scale (a Source with 60 Artifacts, 63 seeded fields): the draft host's actions are in memory over one owner's rows (at most the 63 fields plus extras), each linear in the rows. No action writes or reloads. No caller sends a draft or reads the answered rows in this PR, so no write changes its cost; S15-34 makes one Save one write (S15-33). The Source page's Add dialog makes the same one `setMetadata` write as today.

### Edges

- **Assumptions.** `Apply` refuses a field id listed twice in `values` or `layout`: D-7 removes a re-added field from the dismissed list, `set` never appends a field already on the board, and `reordered(by:)` skips repeats (T-4, T-6). `listWorkspaces` sorts by layout and puts unlisted fields after; the draft lists every visible row, so after Save the engine's order is the draft's (T-1). The board never reads `sortOrder` (`MetadataBoardView`, `MetadataBoardModel`), so renumbering changes no drawing. `moveSaved` always rebuilds rows as saved then suggestions, so `isChanged` compares the two groups apart (T-6).
- **Neighbors.** A suggested row cleared stays empty; a non-suggested row cleared leaves (T-2). Dismiss on a valued row, or a non-suggested row, is a no-op, as the engine's (T-3). A value set back to its opened text is unchanged and gets its link back (T-6). An add of a field a stored layout had dismissed sends `dismissed: false` for it: after Save the dismiss is lifted, so a later clear leaves it as an empty suggestion. This is the researcher choosing the field again; a live board's `Set` keeps the stored flag, so the two differ only after a later clear (T-4 pins the draft's layout).
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
| Draft type, reorder rule | — | `macos/App/Platform/GenealogyStore.swift` `CatalogMetadataDraft`, `reordered(by:)` | D-1, D-6 |
| Error params | — | `macos/App/Platform/CoreInvoke.swift` `CoreInvokeError.params` | D-5 |
| Store writes | — | `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `OnboardingModelTests.swift` `ThrowingStore` | D-2 – D-4 |
| Draft host | Snowflake model | `macos/App/Features/Sources/MetadataDraftHost.swift` *(new)* | D-7 – D-9 |
| Clear switch | — | `MetadataBoardHost.swift` (`confirmsClear`), `MetadataBoardModel.swift` (`requestClear`), `SourcePageMetadataView.swift` | D-10 |
| Add row, dialog over it | — | `MetadataBoardModel.swift`, `MetadataBoardDialogs.swift` | D-11, D-12 |
| Row refusals | — | `MetadataBoardModel.swift` | D-13 |

Sketch of the new host:

```swift
/// A board over rows no write has saved: a form's copy details, sent with its
/// owner's one write as `draft`. Its writes change only its own rows.
@MainActor
final class MetadataDraftHost: MetadataBoardHost {
    var entries: [CatalogMetadataEntry]
    let fields: [CatalogMetadataField]
    var confirmsClear: Bool { false }
    var isChanged: Bool { … }                 // D-8
    var draft: CatalogMetadataDraft? { … }    // D-8
    init(entries: [CatalogMetadataEntry], fields: [CatalogMetadataField])
    func set(fieldID: String, valueText: String) -> ([CatalogMetadataEntry], CatalogWriteEffects)
    // clear, dismissSuggestion, reorder: D-7; didWrite, clearFailure, showFailure: no-ops
    static func unsavedRows(from suggestions: [CatalogTypeSuggestion]) -> [CatalogMetadataEntry]
}
```

States: no view changes. Accessibility: none.

### Strings

None — no new copy. The refusal reuses `error.sourcemetadata.invalid` ("This metadata change can't be saved."); the dialog keeps `metadataFieldRequired` and `metadataValueRequired`.

## 7. Steps

**STEP-1 — The store writes carry a draft and answer the Artifact's board**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/App/Platform/CoreInvoke.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourceGraphFormModel.swift`, `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skill: `add-swift-test`.
Do: D-1 – D-6. Each existing caller passes `metadata: nil`; the three `SourceArtifactsSection` sites and the test sites bind the rows to `_` (`let (updated, _, effects) = …`). Rewrite the doc comments of the four protocol methods per D-2.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes with no Swift warning; `rg -n "metadata: nil" macos/App` lists the five app call sites; `rg -n "case \.coded\(" macos/App/Features/Sources` prints nothing.

**STEP-2 — The draft host and the clear switch**
Files: `macos/App/Features/Sources/MetadataDraftHost.swift` *(new)*, `macos/App/Features/Sources/MetadataBoardHost.swift`, `macos/App/Features/Sources/MetadataBoardModel.swift`, `macos/App/Features/Sources/SourcePageMetadataView.swift`, `macos/ProvenenciaTests/MetadataDraftHostTests.swift` *(new)*, `macos/ProvenenciaTests/MetadataBoardModelTests.swift`. Skill: `add-swift-test`.
Do: D-7 – D-10. `TestMetadataBoardHost` gains `var confirmsClear = true`. Add T-1 – T-7, T-11.
Check: Xcode tests pass; `clearConfirmTargetsItsBoard` passes unchanged.

**STEP-3 — The add row, and the dialog over it**
Files: `macos/App/Features/Sources/MetadataBoardModel.swift`, `macos/App/Features/Sources/MetadataBoardDialogs.swift`, `macos/ProvenenciaTests/MetadataBoardModelTests.swift`, `macos/ProvenenciaTests/MetadataDraftHostTests.swift` *(new)*, created in STEP-2.
Do: D-11, D-12, and the four doc comments under **Deletes**. Add T-8, T-9.
Check: Xcode tests pass; `addDialogWritesToItsTargetBoard`, `invalidValueStaysOnTheFieldOrDialog` and the `SourcePageModelTests` dialog tests pass with no edit (`git diff cd64512 -- macos/ProvenenciaTests/SourcePageModelTests.swift` shows only STEP-1's signature edit).

**STEP-4 — A refused value lands on its row**
Files: `macos/App/Features/Sources/MetadataBoardModel.swift`, `macos/ProvenenciaTests/MetadataDraftHostTests.swift` *(new)*, created in STEP-2.
Do: D-13. Add T-10.
Check: Xcode tests pass; the existing `invalidValueStaysOnTheFieldOrDialog` and `goneOwnerWritesNoState` pass unchanged.

**STEP-5 — Docs**
Files: `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row; S15-36 **Ships**), `.agents/skills/add-workspace-place/SKILL.md` (§ 6 "Metadata values"), this plan (status, Landed note).
Do: per the Definition of done. The map's S15-36 **Ships** replaces "trims values, empty `openHref`" with "answers rows with the typed value (the board trims) and an empty `openHref` until Save", and "the starting rows of a draft with no owner yet (the type's `artifact` suggestions from `listTypeSuggestions`)" with "the starting rows of a draft with no owner yet, mapped from the type's suggestions the form reads (`unsavedRows(from:)`)". The skill line becomes: "Metadata values: a live board's host applies the write's effects with `.wroteMetadataBoard(sourceId:owner:entries:)`, which patches the owner's rows into the cached Source page with no reload (`SourcePageMetadataHost.didWrite`); a draft board (`MetadataDraftHost`) writes nothing, and its owner's Save applies `.wroteMetadataBoard` from the write's reply rows", followed by the line's current text from "; a value created or released" on.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Let a metadata board hold a draft that its owner's write saves.").

## 8. Tests

Rows in T-1 – T-7 and T-10 start from the AC-1.1 board: `provider` (Provider, suggested, empty), `url` (URL, suggested, `https://example.org/a`, `openHref` `https://example.org/a`), `roll` (Roll, not suggested, `12`), in that order; the host's `fields` are `provider`, `url`, `roll` and `film` (Film number). Build with `catalogMetadataField` / `catalogMetadataEntry`; set `openHref` on the URL row after building it. `invalid` is `CoreInvokeError.coded(status: 1, code: "sourcemetadata.invalid", kind: .user, params: …)`.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `MetadataDraftHostTests.swift` `boardActionsBuildTheDraft` *(new)* | Board over the host; `drafts["provider"] = "Ancestry"`, `await board.save(fieldID: "provider")`; `await board.add(fieldID: "film", valueText: "F-7")`; `FakeStore` seeded `artifactsBySource["s1"] = [catalogArtifact("a1", label: "Page 1")]`, `updateArtifactCalls == 0`; `_ = try await store.updateArtifact(projectDir: "/tmp/p", userID: "u1", artifactID: "a1", label: "Page 1", description: "", graphIDs: nil, metadata: host.draft)` → `updateArtifactMetadata == [draft]`, values `provider` Ancestry, `url`, `roll`, `film` F-7, layout 0 – 3 undismissed | AC-1.1, AC-1.6, AC-1.9 |
| T-2 | Swift | `MetadataDraftHostTests` `clearKeepsSuggestionsAndDropsExtras` *(new)* | `board.clear("url")`, `board.clear("roll")`, `board.clear("film")` → `url` suggested and empty with `openHref ""`; `roll` gone; `film` a no-op; `draft.values` holds no `url` or `roll` | AC-1.2 |
| T-3 | Swift | `MetadataDraftHostTests` `dismissOnlyEmptySuggestions` *(new)* | `dismissSuggestion("url")`, `("roll")` → rows unchanged, `isChanged` false; `dismissSuggestion("provider")` → `provider` gone; layout last row `provider`, `sortOrder` 2, dismissed | AC-1.3, AC-1.4 |
| T-4 | Swift | `MetadataDraftHostTests` `readdingADismissedFieldListsItOnce` *(new)* | Dismiss `provider`, then `await board.add(fieldID: "provider", valueText: "Ancestry")` → `provider` a saved row; layout lists `provider` once, `dismissed: false` | AC-1.7 |
| T-5 | Swift | `MetadataDraftHostTests` `draftClearsWithoutConfirm` *(new)* | `await board.requestClear(fieldID: "roll")` → false; `roll` gone | AC-4.1, D-10 |
| T-6 | Swift | `MetadataDraftHostTests` `revertedDraftIsUnchanged` *(new)* | Fill `provider` then clear it → `isChanged` false, `draft` nil; set `url` to `https://example.org/b` then back → unchanged, `openHref` restored; `moveSaved(from: [1], to: 0)` → rows Roll, URL, Provider, `draft` layout 0, 1, 2; `moveSaved(from: [1], to: 0)` again → rows URL, Roll, Provider, `isChanged` false. `[a, b].reordered(by: ["b", "x", "b"])` → b, a, sortOrders 0, 1 | AC-1.5, AC-1.8, D-6 |
| T-7 | Swift | `MetadataDraftHostTests` `unsavedRowsAreTheTypesSuggestions` *(new)* | `unsavedRows(from: [CatalogTypeSuggestion(field: provider, sortOrder: 0), (url, 1)])` → two rows, suggested, empty, in that order; `[]` → `[]` | AC-1.10, D-9 |
| T-8 | Swift | `MetadataDraftHostTests` `addRowConfirmsOnlyWhenFilled` *(new)*; `MetadataBoardModelTests` `addRowOnLiveBoard` *(new)* | Draft host: closed → `canConfirmAddRow` false; open, no field → false; field, value `"  "` → false; `film`, ` F-7 ` → confirm, rows end with `film` `F-7`, `addRow` nil. `TestMetadataBoardHost` (`fields = [film]`), `board.openAddRow()`, then `board.addRow?.fieldID = "film"`, `board.addRow?.value = "F-7"`: `setError = invalid` (params `[]`), `failureText = "Bad"` → `addRow?.error == "Bad"`, row open, `isSavingAddRow` false; `setError = boom` → `.showFailure("Bad")` recorded, `addRow?.error` nil, row open | AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| T-9 | Swift | `MetadataBoardModelTests` `addRowSurvivesReset` *(new)* | `TestMetadataBoardHost` (`fields = [film]`, no entries); row open with `film`, `F-7`; `resetDrafts()` → unchanged; `host.entries = [film row "F-1"]`, `resetDrafts()` → `fieldID ""`, `value "F-7"`, open | AC-2.5, AC-2.6 |
| T-10 | Swift | `MetadataDraftHostTests` `refusalLandsOnItsRow` *(new)* | `board.beginEdit(fieldID: "roll")`; `invalid` with params `["url"]` → true; `editingFieldID == "url"`, `drafts["roll"] == "12"`, `drafts["url"] == "https://example.org/a"`, `fieldErrorID == "url"`, `fieldError == L10n.Errors.message(for: error)`. Params `[]`, `[""]`, `["gone"]`, code `artifacts.invalid` → false, nothing changed | AC-3.1, AC-3.2, D-13 |
| T-11 | Swift | `MetadataBoardModelTests` `liveBoardAsksBeforeClearing` *(new)* | `TestMetadataBoardHost` with a saved row: `await board.requestClear(fieldID:)` → true, no `.clear` call; with `confirmsClear = false` → false, `.clear` recorded | AC-4.2, D-10 |
| T-12 | Swift | `MetadataBoardModelTests` `addDialogWritesToItsTargetBoard`, `invalidValueStaysOnTheFieldOrDialog`; `SourcePageModelTests` dialog tests near lines 878, 931, 1114 *(existing, unchanged)* | As today | AC-2.7, D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests` (artifact create and edit, the metadata dialog), `MetadataBoardModelTests`, `EvidenceGraphStoreTests`, `SourceGraphFormModelTests`.

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
| `requestClear` delays the Source page's confirm by one `Task` hop | The view starts it at the click; no state is read between |
| The dialog over the add row changes what the Source page's Add does | D-12 keeps every member name; T-12 runs the existing dialog tests unchanged |
| A draft listing a field twice is refused on Save | D-7 drops a re-added field from the dismissed list; `reordered(by:)` skips repeats; T-4, T-6 |

**Stop and report if:**

- `CreateArtifactResponse.metadata`, `UpdateArtifactResponse.metadata` or the `metadata` request fields of §4 are missing or typed otherwise in `engine.pb.swift` at your base.
- `MetadataBoardView` or `MetadataBoardModel` reads `sortOrder` (D-6's renumbering would then change a drawing).
- A `MetadataBoardHost` conformer exists besides `SourcePageMetadataHost` and `TestMetadataBoardHost`.
- `sourcemetadata.ErrInvalid`'s refused value no longer carries the field id as `params[0]` (S15-33 D-6).
- Any test outside §8 or STEP-1's signature changes needs editing to pass (T-12 must pass unchanged).

## 12. Open questions

None.

## 13. PR

**Title:** S15-36 — A metadata board can hold a draft
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-36-metadata-board-holds-a-draft.md`.
