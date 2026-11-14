# S15-10 — Adding an Artifact asks for its copy details

**Status:** Draft
**Spike / slice:** Spike 15 · Slice 2 — Metadata at three levels
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S15-D14](../design/S15-D14-source-page-artifacts.md) SA-7 (and frame 13's empty board, SA-13), board agreed 2026-11-13
**Board:** [Claude Design — Source Detail](https://claude.ai/artifact/443NPNYmGErAmxgV1fBsFK), page *Source page*: row "Add an Artifact", frames `sheet-add` (a file chosen, the four suggestions empty) and `sheet-add-refused` (the refused-file Callout); row "Artifact sheet", frame `sheet-no-sugg` (a type with no Artifact suggestions: only the Add field row). The implementer reads it with the Artifact tool's `read` on that link, `path` "project/Source.dc.html": the sheet is the `PVFormDialog open="{{ sheetOpen }}"` block; add mode is its `sheetIsAdd` branches (the File section above Label, the graph group `nlRows`), and `sheetSaveDisabled`'s add arm. The board's other "Add an Artifact" frames (`add-one` to `add-attach`) are the old Add dialog's states and are not built (see Out)
**Depends on:** S15-34 (landed [#819](https://github.com/mendahu/provenencia/pull/819)), S15-36 (landed [#810](https://github.com/mendahu/provenencia/pull/810)), S15-33 (landed [#806](https://github.com/mendahu/provenencia/pull/806)), S15-35 (landed [#824](https://github.com/mendahu/provenencia/pull/824)), S15-D14 (board agreed 2026-11-13)
**Blocks:** —
**Branch:** `s15-10-artifact-sheet-add-mode`
**Version bump:** minor, 0.13.1 → **0.14.0** (or the next minor after `main`'s `VERSION` when you rebase): `VERSION`, `core/version.go`, and every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj` — a backward-compatible feature on a screen (Add artifact opens the Artifact sheet and saves copy details with the new Artifact); no wire or catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md) (two keys removed, comments updated), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (read only: no kit change; the sheet stays a Source page snowflake on `PVFormDialog`), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (read only: the create applies through `SourceArtifactsSection.show`, as § 6 "Metadata values" says)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "An Artifact is added and edited in one sheet", "Metadata saves with its owner's write", "One metadata board", "Suggestions have a level", "One brief, one PR" (S15-D14 predates it and is exempt) ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); the map's S15-10 scope entry; [S15-D14](../design/S15-D14-source-page-artifacts.md) § 2, § 2.1, SA-7, § 6 and its **Board** line; S15-34's [plan](S15-34-artifact-edit-sheet.md) D-4, D-7 and **Next adds** ("S15-10: `ArtifactSheetModel.Mode.create`, the File row first, `MetadataDraftHost.unsavedRows(from:)`, `createArtifact(…, metadata: host.draft)` applied through `show(_:metadata:effects:)`; it removes the Add artifact dialog") and its **Landed** note; S15-36's [plan](S15-36-metadata-board-holds-a-draft.md) D-9 and **Next adds**, and its [review](S15-36-metadata-board-holds-a-draft.review.md) X-15 (logged for this PR)
**Planned:** 2026-11-14 against `main` @ `c09b83b`
**Review:** — (not reviewed yet)
**Runs on:** macOS. Every Swift check and `scripts/check-pr.sh` need Xcode ([`README.md`](../../../../README.md)). No proto or catalog change: never run `./scripts/generate-proto.sh` in this PR

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

Add artifact opens the Artifact sheet in add mode, so a researcher records a new Artifact's copy details (provider, URL, accessed date on a census image) at the moment they add it, not afterwards in a second edit. The sheet is the one S15-34 built for editing: File first, then label, description, Evidence graphs and the copy-details board, which starts as the Source type's `artifact`-level suggestions, empty. The Artifact and its copy details save in one `CreateArtifact` write; the file then ingests as its own write, as today. The separate Add artifact dialog and its draft go.

**Why now.** S15-33, S15-36 and S15-34 built the engine write, the draft board and the sheet; S15-34's **Next adds** names this PR as the sheet's add mode. Working rule "An Artifact is added and edited in one sheet" binds it.

**What stays.** The edit sheet and everything it does. The Add file dialog for an Artifact with no file (first attach): same chrome, same behavior; only its file state moves onto the shared `IngestFileChoice` (D-11). The page-level file drop, the created toast, the cover refresh after an add, the Artifact list and its row. The engine: no Go, proto or catalog change.

## 2. User stories

**US-1.** As a researcher, when I add an Artifact, I fill its copy details in the same sheet, so they save with it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source "1880 census, Ottawa" (`SRC-4KD2P`) whose type suggests Provider, URL, Accessed at the `artifact` level; click **Add artifact** | The sheet opens titled "Add artifact", subtitle "SRC-4KD2P · 1880 census, Ottawa", File section first, then Label, Description, Evidence graphs, and the copy-details board with Provider, URL, Accessed as empty suggested rows in that order. The primary button reads "Add artifact" and is disabled |
| AC-1.2 | AC-1.1's sheet; type Label "Sheet 26 (image 13)", fill Provider "FamilySearch" and URL "https://example.org/ark/1"; click **Add artifact** | One `createArtifact` call with that label and a draft holding the two values; no `updateArtifact` call. The sheet closes, the list shows the new Artifact with Provider and URL in its copy details, and the "ART-… added" toast shows |
| AC-1.3 | AC-1.1's sheet; type a label only; Add | `createArtifact` is sent with `metadata: nil`. The reply's rows (the three empty suggestions) are put on the page for the new Artifact, so its edit sheet opened next shows the three suggestions |
| AC-1.4 | A type with no `artifact` suggestions; Add artifact | The board shows only the Add field button (frame `sheet-no-sugg`). Add field "Film number" = "F-7", label "Roll 12", Add → the draft carries Film number |
| AC-1.5 | AC-1.2, but the engine refuses the URL value (`sourcemetadata.invalid`, params `[url field id]`) | The sheet stays open; the URL row is in edit with "This metadata change can't be saved." under it; no Artifact is on the page |

**US-2.** As a researcher, I attach the file in the same sheet, as the Add dialog let me.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Drop one `kansas-1880-sheet-26.jpg` on the Source page | The sheet opens in add mode with the file named in the File row. Label "Sheet 26", Add → `createArtifact`, then `ingestArtifactFile` on the new Artifact; the sheet closes |
| AC-2.2 | AC-2.1, but the engine refuses the file (`ingest.invalid`) | The Artifact is on the page with no file; the sheet stays open with the refused-file Callout under the File row; the graph boxes are disabled; Add artifact is disabled until the file is cleared or another one is chosen |
| AC-2.3 | AC-2.2; choose a good file, change Provider to "Ancestry", Add | No second `createArtifact`. One `updateArtifact` on the made Artifact with `graphIDs: nil` and the draft, then `ingestArtifactFile`; the sheet closes |
| AC-2.4 | AC-2.2; choose a good file and change nothing else; Add | No `updateArtifact` call; one `ingestArtifactFile`; the sheet closes |
| AC-2.5 | Drop two files on the page | The sheet opens with the multi-file refusal Callout and no file; Add artifact is disabled |
| AC-2.6 | Choose `notes.xlsx` in the File row | The refusal Callout shows with no engine call; Add artifact is disabled |

**US-3.** As a researcher, I choose the Evidence graphs the new Artifact is in.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Graphs B, J, S; uncheck B and S; Add | `createArtifact` sends `graphIDs: ["g-j"]` |
| AC-3.2 | Uncheck all three | "This artifact won't be in an evidence graph…" shows under the group; Add sends `graphIDs: []` |
| AC-3.3 | The session has no graph list for the Source | No graph group shows; Add sends `graphIDs: nil` (the engine links every graph) |
| AC-3.4 | Graphs B, J, S shown; another window makes graph N; Add with B, J, S checked | `createArtifact` sends `["g-b", "g-j", "g-s", "g-n"]`: a graph made while the sheet is open starts linked, as the Add dialog's did |

**US-4.** As a researcher, I can leave the sheet without surprises.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Add mode with a label typed; press Escape | The kit's discard confirm shows. Cancel closes at once with nothing written |
| AC-4.2 | Add mode with nothing typed, no file, boxes untouched; Escape | The sheet closes with no confirm |
| AC-4.3 | The `artifact`-level suggestion read fails | No sheet opens; the page banner shows the error message |
| AC-4.4 | AC-2.2's sheet (Artifact made, file refused); another window deletes that Artifact | The sheet closes and the "Artifact no longer here" toast names "Sheet 26" |
| AC-4.5 | Add mode before any Save; another window adds a note to the Source (the workspace changes) | The sheet stays open with everything typed |

## 3. Scope

| In | Covered by |
| --- | --- |
| SA-7: Add artifact opens the Artifact sheet in add mode: title "Add artifact", File (today's drop row and refused-file Callout) first, then the edit mode's fields; graphs start checked and lock after a refused file; copy details start as the type's `artifact` suggestions, empty; button "Add artifact", enabled once Label has text | STEP-3, STEP-5; AC-1.1, AC-2.2, AC-3.1 |
| The suggestions are read with `listTypeSuggestions(level: .artifact)` when the sheet opens (D-2) | STEP-4; AC-1.1, AC-4.3 |
| One `createArtifact` with the label, description, graph ids and the draft; the reply's rows applied through `show(_:metadata:effects:)` | STEP-3; AC-1.2, AC-1.3 |
| A refused value lands on its row; nothing is created | STEP-3; AC-1.5 |
| The file ingests as its own write after the create; a refused file keeps the made Artifact and a retry saves onto it, sending the draft on `updateArtifact` only when something changed | STEP-3; AC-2.1 – AC-2.4 |
| The page drop opens the sheet in add mode with the file staged | STEP-4, STEP-5; AC-2.1, AC-2.5 |
| One `IngestFileChoice` value for a chosen file, used by the sheet and the Add file dialog | STEP-2; AC-2.5, AC-2.6 |
| The Add artifact dialog, `SourceArtifactsSection.Draft` and its members, `addForm`, and the two dead strings go | STEP-4, STEP-5; AC-1.1 |
| FakeStore hooks for the create's error and id and the suggestion read (S15-36 review X-15) | STEP-1 |

| Out | Home |
| --- | --- |
| Filling copy details automatically from a URL or a provider | [`docs/ideas/registry-source-import.md`](../../../ideas/registry-source-import.md) |
| Editing a type's `artifact`-level suggestions on the type page | S15-12 |
| Seeded `artifact`-level suggestions on the shipped types | S15-19 |
| Move on the Artifact's menu | S15-25 |
| The graph form's board | S15-11 |
| An ingest progress bar and the old dialog's other frames on the board (`add-one`, `add-some`, `add-none`, `add-loading`, `add-copying`, `add-failed`, `add-zip`, `add-two`, `add-attach`) | Not planned: SA-7 and the § 6 inventory ship today's drop row ("File drop — Ship"); the app has no ingest progress today |
| Dropping the cover reload (`refreshCoverFromStore`) after an add | Not planned: an ingest can set the Source's cover, and the Add file path reloads the same way; this PR keeps the write path as it is |
| Any change to the edit sheet's behavior | Not planned |

## 4. Current state

Verified at `c09b83b`.

| Area | Where | Today |
| --- | --- | --- |
| Section model | `macos/App/Features/Sources/SourceArtifactsSection.swift` `SourceArtifactsSection` | Holds the Add dialog: `struct Draft` (label, description, `filePath`, `fileName`, `reject`, `createdID`, `excludedGraphIDs`), `isAdding`, `draft`, `draftLabelError`, `isSavingDraft`, `draftGraphIDs`, `draftInNoGraph`, `draftGraphsLocked`, `canSubmitDraft`, `addSheetIsDirty`, `draftGraphsChanged`, `openAdd()`, `cancelAdd()`, `pickFile()`, `clearFile()`, `create()` (create with `metadata: nil`, `show` without rows, then ingest; on an ingest callout keeps the dialog open with `createdID`; a retry calls `updateArtifact` only when label or description changed), and the shared file code `FilePickTarget { create, attach }`, `applyDroppedURLs(_:into:)`, `applyPickedFile(url:into:)`, `setReject(_:fileName:into:)`. `stagePageDrop(_:)` calls `openAdd()` then `applyDroppedURLs(_, into: .create)`. `openSheet(artifactID:at:)` builds `ArtifactSheetModel(mode: .edit(art))`. `workspaceChanged()` reads `sheet.artifact.id` and `sheet.openedLabel`. `show(_:metadata:effects:)` is the one apply path (S15-34 D-7). Add file: `struct AttachDraft` (`artifactID`, `artifactRef`, `filePath`, `fileName`, `reject`), `openAttach`, `cancelAttach`, `pickAttachFile`, `clearAttachFile`, `confirmAttach`, `canSubmitAttach`, `attachSheetIsDirty` |
| Sheet model | `macos/App/Features/Sources/ArtifactSheetModel.swift` `ArtifactSheetModel` | `enum Mode { case edit(CatalogArtifact) }`; `openedLabel` is `let`; `prefill`, `draft`, `labelError`, `saveError`, `isSaving`, `scrollTarget`/`scrollRequest`; `private let locks = UnlinkLocks()` (state `.loading`); `graphIDs`, `excludedGraphIDs`, `graphsChanged()` (sets choices from `links(artifact.id)` and starts `loadLinkLocks`), `linkedGraphIDs`, `liveGraphIDs`, `lockedGraphs`, `checkedGraphIDs`, `graphs`, `inNoGraph`, `linksChanged`, `savedGraphIDs`, `loadLinkLocks()`; `host`/`board` from the workspace rows; `var artifact: CatalogArtifact` (non-optional); `isDirty`, `canSave`, `save()` (one `updateArtifact`), `scrollToTop()`, `saveErrorBody(_:)`; `subtitle` |
| Sheet view | `macos/App/Features/Sources/ArtifactSheet.swift` `ArtifactSheetModifier`, `ArtifactSheetForm` | Copy: `verbatimTitle: L10n.Sources.editArtifactLabel(ref: model?.artifact.ref ?? "")`, `confirm: L10n.Sources.saveArtifact`. Body: save error and `fields` in the `.top` group, `graphLinks` (loading, `SourceFormLoadArea` + `SourceArtifactGraphLinks(locked:isDisabled: model.isSaving)`, the no-graph line), `copyDetails` (`MetadataBoardView(placement: .form)`) |
| Add dialog view | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `SourcePageArtifactsView.addForm`; the header's and empty state's Add buttons call `model.artifacts.openAdd()` | `addForm`: Label (with `artifactLabelHint`), Description (`PVInput`, `formDescriptionHint`), graphs (`isDisabled: draftGraphsLocked`), File (`PVField(label: optionalFile)` + `IngestFileDropRow` + reject `PVCallout`, id `sources.page.addArtifact.ingestReject`). `addFileForm` reads `attachDraft.fileName`, `.reject` and calls `applyDroppedURLs($0, into: .attach)` |
| Page view | `macos/App/Features/Sources/SourcePageView.swift` | `.pvFormDialog(isPresented: addArtifactPresented, copy: addArtifactDialogTitle / addArtifactDialogSubtitle / addArtifactConfirm, …) { SourcePageArtifactsView(model:).addForm }`; `addArtifactPresented` binding; `.dropDestination` calls `model.artifacts.stagePageDrop(urls)` synchronously; `.artifactSheet($model.artifacts.sheet)` |
| Drop gate | `macos/App/Features/Sources/SourcePageModel.swift` `acceptsArtifactPageDrop` | `!artifacts.isAdding && !artifacts.isAttaching && … && artifacts.sheet == nil` |
| Draft host | `macos/App/Features/Sources/MetadataDraftHost.swift` `MetadataDraftHost.unsavedRows(from:)` | Pure; each suggestion empty, `suggested: true`, in order. No caller yet (S15-36 D-9) |
| Store | `macos/App/Platform/GenealogyStore.swift` `createArtifact(…, metadata:)` → `(artifact, metadata, effects)`; `listTypeSuggestions(projectDir:typeID:level:)`; `GoStore.swift` maps both | `CreateArtifactResponse.metadata` is always the new Artifact's board, suggestions included, with or without a draft (`api/ffi/handlers/artifacts.go` `CreateArtifact` → `ownerBoardProto`; Go test `TestCreateArtifactMetadata`). A refused draft value fails the whole create (`sourcemetadata.invalid`, params `[field id]`); no Artifact is made |
| Session key | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `.typeSuggestions` | Reads the `source` level only, for the type page. Not used here (D-2) |
| Fake | `macos/App/Platform/FakeStore.swift` `createArtifact`, `listTypeSuggestions`, `metadataWriteResults`, `scriptedMetadataRows` | `createArtifact` never throws and mints a random id, so a test cannot script its reply rows (S15-36 review X-15); `listTypeSuggestions` throws only for a level that holds no metadata and records nothing |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Sources` | `addArtifactDialogTitle` "Add artifact", `addArtifactDialogSubtitle`, `addArtifactConfirm` "Add artifact", `optionalFile` "File", `artifactLabelHint`, `ingestDropIdleCreate`, `toastArtifactCreatedTitle/Fileless/WithFile`; `formDescriptionHint` (also used by `SourcesListView`). Comments on `artifactGraphsLabel`, `artifactGraphsHint`, the no-graph line and `addArtifact` name "the Add artifact dialog" |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` | The Add dialog's tests: `createFilelessArtifact`, `createArtifactRequiresLabel`, `createArtifactWithPDFLeavesTypeIconCover`, `creatingAnArtifactWithAFileOpensTheGraphJump`, `useAsThumbnailAndRevertToTypeIcon`, `filelessArtifactCannotBeCover`, `createFilelessArtifactClearsMIMECoverForTypeIcon`, `aRefusedFileKeepsTheCreatedArtifactAndARetryAttachesToIt`, `ingestMissingFileSurfacesIngestInvalid`, `applyPickedFileRejectsSpreadsheetWithoutFFI`, `ingestThenRejectSecondAttach`, `pageDropStagesFileWithoutCreating`, `pageDropOfTwoFilesOpensRejectCallout`, `addArtifactOnOneGraphShowsTheGroup`, `addArtifactLinksOnlyTheCheckedGraphs`, `addArtifactWithNoGraphChecked`, `addArtifactExcludedGraphIsDirty`, `addArtifactRetryIgnoresGraphChoice`, and the sheet tests that read `sheet.artifact` or `isAdding` (`pencilOpensTheSheet`, `linkToAGraphOpensAtGraphs`, `artifactGoneClosesTheSheet`, and every other `sheet.artifact.id` read) |

**Precedent.** S15-34 ([#819](https://github.com/mendahu/provenencia/pull/819), merge `64a91ae`) built this sheet in edit mode: copy its `save()` shape (unconfirmed-input stop, label check, one write, `onSaved`, `board.showRefusal`, `saveErrorBody`, `scrollToTop`). The two-mode model is `SourceGraphFormModel` (`Mode { create, edit }`; create sets `UnlinkLocks(state: .ready)` and reads no locks) and `SourceGraphsSection.openCreate`/`openEdit`: copy that split. The retry-after-refused-file rule is today's `SourceArtifactsSection.create()`: move its logic, do not rewrite it.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **One model, a second mode.** `ArtifactSheetModel.Mode` gains `case create(suggestions: [CatalogMetadataEntry])`. Create mode: `prefill = ArtifactSheetDraft()`, `host = MetadataDraftHost(entries: suggestions, fields: context.fields)`, `locks = UnlinkLocks(state: .ready)` (`locks` becomes an init-set `private let`), `openedLabel = ""`. `opensAt` is `.top` | Working rule "An Artifact is added and edited in one sheet"; S15-34 D-4 and **Next adds**; the graph form's precedent | A second `ArtifactAddModel` (a copy of the sheet's graphs, board, save and errors) |
| D-2 | **The suggestion read.** `SourceArtifactsSection.openAdd() async -> ArtifactSheetModel?` *(was sync `openAdd()`, `@discardableResult`)* reads `context.store.listTypeSuggestions(projectDir:, typeID: context.source.sourceTypeID, level: .artifact)` before the sheet opens, maps it with `MetadataDraftHost.unsavedRows(from:)`, and builds the sheet in create mode. An empty `sourceTypeID` (or no `context.source`) skips the read and opens with no suggestions. On a failure no sheet opens and the page banner shows the message (`context.showFailure(error)`). After the read it checks `sheet == nil` again and opens nothing when a sheet opened meanwhile | S15-36 D-9: "the form makes the read", no session key (a form's rows live as long as the form); the read is one local call, so the sheet opens with its board complete and the host's `opened` rows are the suggestions, which `isChanged` and `draft` compare against | A `CatalogQueryKey` with a level (a cache for a value read once per open, and its invalidation); opening the sheet first and filling the board when the read lands (`host` and `board` are `let`; replacing them under an open edit loses typed text); opening without suggestions on a failure (the researcher would not know the type suggests fields) |
| D-3 | **No second open while one is pending.** `SourceArtifactsSection.isOpeningAdd` *(new, `private(set) var Bool`)* is true while `openAdd` reads; `openAdd` and `stagePageDrop` return at once while it is true or `sheet != nil`. `SourcePageModel.acceptsArtifactPageDrop` replaces `!artifacts.isAdding` with `!artifacts.isOpeningAdd` | A double click or a drop during the read must not open two sheets | No guard (two reads race, the second overwrites a sheet the researcher typed in) |
| D-4 | **The page drop.** `stagePageDrop(_ urls: [URL]) async`: guard as D-3, `guard let sheet = await openAdd() else { return }`, then `sheet.file.choose(urls)`. `SourcePageView`'s `.dropDestination` keeps its guard, starts `Task { await model.artifacts.stagePageDrop(urls) }` and returns true | The sheet that opened is the one that takes the file; an edit sheet opened meanwhile never gets it | Writing into `self.sheet` after the await |
| D-5 | **Which Artifact the sheet holds.** `var artifact: CatalogArtifact` is replaced by `var artifactID: String?` *(new)*: the edited Artifact's id, or in create mode the Artifact the first Save made (`created?.id`), nil before it. `private(set) var created: CatalogArtifact?` *(new)*. `openedLabel` becomes `private(set) var`; create mode sets it to the label the create sent. Edit-only code (`graphsChanged`'s exclusion, `linkedGraphIDs`, `savedGraphIDs`' filter, `loadLinkLocks`) reads the edited Artifact through a `private var editedArtifact: CatalogArtifact?` *(new)*. `SourceArtifactsSection.workspaceChanged()` closes the sheet only when `sheet.artifactID` is non-nil and that id is not on the page | One field answers "which Artifact" in both modes; an add sheet with nothing made has nothing to lose | An optional `artifact` that edit code force-unwraps |
| D-6 | **Graphs in create mode.** `graphsChanged()` sets `graphIDs` once, as today; create mode sets `excludedGraphIDs = []` (every box starts checked) and starts no lock read. `linkedGraphIDs` is `liveGraphIDs` in create mode (a new Artifact's untouched boxes are all checked), so `linksChanged` and `isDirty` follow. `savedGraphIDs` in create mode is `checkedGraphIDs` plus every session graph of the Source not in `graphIDs` (made since the sheet opened: starts linked, as the dialog's live list did). `var graphsLocked: Bool { created != nil }` *(new)*; the view passes `isDisabled: model.isSaving \|\| model.graphsLocked` | Today's dialog behavior (state is the unchecked set; the retry keeps the made Artifact's links); SA-7 "after a refused file the graph boxes are locked" | Locks with hints (there is no engine block to show); hiding the group after a refused file |
| D-7 | **The create Save.** `save()` dispatches on `mode`: edit runs today's body unchanged; create runs `saveCreate()` *(new, private)*: (1) clear `saveError`, `labelError`; (2) `board.hasUnconfirmedInput` → `artifactSheetFinishOpenField`, scroll to top, false; (3) empty label → `artifactLabelRequired`, scroll to top, false; (4) `file.reject != nil` → false; (5) `context.clearPageError()`, `isSaving` for the rest; (6) if `created == nil`: `createArtifact(sourceID:, fileID: "", label:, description:, graphIDs: graphIDs == nil ? nil : savedGraphIDs, metadata: host.draft)`, then `onSaved(art, rows, effects)` **always with the reply's rows**, `created = art`, `openedLabel = label`, `sent = Sent(label, description, metadata)`; (7) else (a retry): when `label`, `description` or `host.draft` differ from `sent`, one `updateArtifact(artifactID: created.id, label:, description:, graphIDs: nil, metadata: host.draft)`, `onSaved(updated, metadata == nil ? nil : rows, effects)`, `created = updated`, `sent` updated; (8) if `file.filePath` is set: `ingestArtifactFile`, `onSaved(ingested.artifact, nil, ingested.effects)`; (9) `await context.refreshCoverFromStore()`, the created toast exactly as today (`toastArtifactCreatedTitle(ref:)`, fileless or with-file body, `.success`), return true. `private struct Sent: Equatable { label, description, metadata: CatalogMetadataDraft? }` *(new)*; `private var sent: Sent?` | Working rule "Metadata saves with its owner's write"; S15-33's reply always carries the board, so the new Artifact's rows (its suggestions) reach the page and the cached key through `show`'s `.wroteMetadataBoard` with no reload; the retry keeps today's "write only what changed" so a retry adds no empty revision | Sending `metadata` on every retry (an update and a revision for nothing); the edit mode's `metadata == nil ? nil : rows` on the create (the new owner would have no rows on the page until the cover reload) |
| D-8 | **Create failures.** In `saveCreate`'s `catch`, in this order: (a) a `CoreInvokeError` whose code maps through `L10n.Errors.ingestCallout(code:params:)` → `file.refuse(callout)`, `scrollToTop()`, false (the sheet stays; `created` is set when the create had run); (b) `guard let message = context.fail(error) else { return true }`; (c) `board.showRefusal(error)` → false; (d) `saveError = saveErrorBody(message)`, `scrollToTop()`, false | Today's order in `create()` for (a); the edit Save's order for (b)–(d); a refused value fails the whole create, so (c) leaves `created` nil and the page untouched | A new error copy for add mode (the board's "Couldn't save the artifact" title and "Your changes are kept." body fit both) |
| D-9 | **Enablement and dirt in create mode.** `canSave`: create → `!isSaving && file.reject == nil && !draft.label.trimmed.isEmpty`; edit unchanged. `isDirty`: create → `textChanged \|\| linksChanged \|\| board.hasUnsavedChanges \|\| file.filePath != nil`; edit unchanged | SA-7 "enabled once Label has text"; today's `canSubmitDraft` and `addSheetIsDirty`; the board's `sheetSaveDisabled` add arm | Checking the label only in `save()` as edit does (SA-7 says the button waits for it) |
| D-10 | **Copy on the model.** `var title: String` *(new)*: edit → `L10n.Sources.editArtifactLabel(ref:)`, create → `L10n.string(L10n.Sources.addArtifactDialogTitle)`. `var confirmLabel: LocalizedStringResource` *(new)*: edit → `saveArtifact`, create → `addArtifactConfirm`. `subtitle` is unchanged ("SRC-4KD2P · 1880 census, Ottawa" in both modes, as the board draws). `ArtifactSheetModifier` reads `model?.title ?? ""` and `model?.confirmLabel ?? L10n.Sources.saveArtifact`; the prefix stays `sources.page.artifactSheet` | One sheet, one identifier prefix; the board's add subtitle is the edit one | Keeping `addArtifactDialogSubtitle` (the board drops it) |
| D-11 | **One file choice.** `struct IngestFileChoice: Equatable` *(new)* in `macos/App/Features/Sources/IngestFileChoice.swift` *(new)*: `filePath: String?` (absolute, set only when the app accepts the file), `fileName: String?`, `reject: L10n.Errors.IngestCallout?`; `mutating func choose(_ urls: [URL])` (more than one → `ingestCallout(reason: .multiFile)` naming the first; one → `choose(_ url:)`); `mutating func choose(_ url: URL)` (today's `applyPickedFile` body: `IngestMediaPolicy.validate`, the too-large size label); `@MainActor mutating func pick()` (today's `ProjectFiles.pickFileForIngest` prompt and message, then `choose(url)`); `mutating func refuse(_ callout:)` (`filePath = nil`, name kept, `reject = callout`); `mutating func clear()`. The sheet holds `var file = IngestFileChoice()`. `AttachDraft`'s `filePath`/`fileName`/`reject` become `var file = IngestFileChoice()`; `pickAttachFile()` → `attachDraft.file.pick()`, `clearAttachFile()` → `attachDraft.file.clear()`, `attachDroppedURLs(_ urls: [URL])` *(new, replaces `applyDroppedURLs(_:into: .attach)`)*, `confirmAttach` uses `attachDraft.file.refuse(callout)`. `FilePickTarget`, `applyDroppedURLs(_:into:)`, `applyPickedFile(url:into:)` and `setReject` go | The create half of the file code moves into the sheet; one value keeps the two file rows (add, attach) on one validation instead of a copy | Copying `applyPickedFile` into the sheet model; the sheet calling back into the section for file state |
| D-12 | **What goes from the section and page.** Deleted: `SourceArtifactsSection.Draft`, `isAdding`, `draft`, `draftLabelError`, `isSavingDraft`, `draftGraphIDs`, `draftInNoGraph`, `draftGraphsLocked`, `canSubmitDraft`, `addSheetIsDirty`, `draftGraphsChanged`, `cancelAdd()`, `pickFile()`, `clearFile()`, `create()`; `SourcePageArtifactsView.addForm`; `SourcePageView`'s Add artifact `.pvFormDialog` and `addArtifactPresented`. The header's and empty state's Add buttons call `Task { await model.artifacts.openAdd() }` | SA-7 replaces the dialog; S15-34 **Next adds** | Keeping the dialog behind a flag |
| D-13 | **The File section's place and look.** In create mode `ArtifactSheetForm` draws, inside the `.top` group, above `fields`: `PVField(label: L10n.Sources.optionalFile) { IngestFileDropRow(fileName: model.file.fileName, isRejected: model.file.reject != nil, isBusy: model.isSaving, idleHint: L10n.Sources.ingestDropIdleCreate, onChoose: { model.file.pick() }, onClear: { model.file.clear() }, onDropURLs: { model.file.choose($0) }); if let reject = model.file.reject { PVCallout(tone: .danger, title: reject.title, message: reject.message) } }`, the Callout id `sources.page.artifactSheet.ingestReject` *(new)*. The save-error Callout stays first. Label and Description are the edit mode's (`PVTextArea`, `artifactDescriptionHint`, Label takes focus on open) | SA-7 "File (today's drop row, optional, with its refused-file Callout under it), then the edit mode's fields"; frame `sheet-add`; the § 6 inventory "File drop — Ship, add mode only" | A new file component; the board's uppercase caption label (the kit's `PVField` label is the app's field label) |
| D-14 | **Strings.** Reused: `addArtifactDialogTitle` (title), `addArtifactConfirm` (button), `optionalFile`, `ingestDropIdleCreate`, `toastArtifactCreated*`, every edit sheet key. Removed (no caller left): `addArtifactDialogSubtitle`, `artifactLabelHint`, from `L10n.swift` and `Localizable.xcstrings`. Comments that name "the Add artifact dialog" (`addArtifact`, `addArtifactDialogTitle`, `addArtifactConfirm`, `optionalFile`, `artifactGraphsLabel`, `artifactGraphsHint`, the no-graph line) say "the Artifact sheet in add mode" instead. No new key | One brief: the board's add copy is the app's existing copy | Renaming the two reused keys (churn in the catalog for no copy change) |
| D-15 | **FakeStore.** `createArtifactError: Error?` *(new)*: `createArtifact` records the call (`createArtifactGraphIDs`, `createArtifactMetadata`) and then throws it. `nextArtifactID: String?` *(new)*: the next `createArtifact` uses it as the id and clears it, so a test scripts `metadataWriteResults[.artifact owner]` for the reply (S15-36 review X-15). `listTypeSuggestionsError: Error?` *(new)* and `listTypeSuggestionsLevels: [Level]` *(new, each call's level in order)*. No engine rule moves into the fake | The tests in § 8 need each; `.cursor/rules/swift-tests.mdc` (the fake scripts answers, it does not judge) | A fake that applies the draft |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ArtifactSheetModel.Mode.create(suggestions:)` *(new)* | enum case | `SourceGraphFormModel.Mode.create` | Extends the sheet S15-34 built for this (its **Next adds**) |
| `ArtifactSheetModel.artifactID`, `created`, `editedArtifact`, `graphsLocked`, `title`, `confirmLabel`, `file`, `saveCreate()`, `Sent`, `sent` *(new)* | properties / func / type | `SourceArtifactsSection.Draft.createdID`, `create()`, `draftGraphsLocked` | Moved from the section's dialog into the sheet; the section's copies are deleted (D-12) |
| `IngestFileChoice` *(new)* | type / file | `SourceArtifactsSection.applyPickedFile(url:into:)`, `Draft`/`AttachDraft` file fields | Extracted from them so the sheet and the Add file dialog share one; searched `IngestMediaPolicy.validate` callers: only `applyPickedFile` |
| `SourceArtifactsSection.isOpeningAdd`, `attachDroppedURLs(_:)` *(new)*; `openAdd()` and `stagePageDrop(_:)` become `async` | property / func | `isAdding`, `applyDroppedURLs(_:into:)` | Replace them (D-3, D-11) |
| `FakeStore.createArtifactError`, `nextArtifactID`, `listTypeSuggestionsError`, `listTypeSuggestionsLevels` *(new)* | test hooks | `updateArtifactError`, `createArtifactGraphIDs` | Same shape as the update hooks |
| `macos/ProvenenciaTests/IngestFileChoiceTests.swift` *(new)* | test file | `SourcePageModelTests.applyPickedFileRejectsSpreadsheetWithoutFFI` | The value is pure; its tests need no page |
| `sources.page.artifactSheet.ingestReject` *(new)* | a11y id | `sources.page.addArtifact.ingestReject` | Replaces it under the sheet's prefix |

**Builds on.** `ArtifactSheetModel` and `ArtifactSheet` (S15-34), `MetadataDraftHost.unsavedRows(from:)` (S15-36 D-9), `SourceArtifactsSection.show(_:metadata:effects:)` (S15-34 D-7, the one apply path), `UnlinkLocks(state: .ready)` (the graph form's create), the `createArtifact` reply's rows (S15-33).
**Deletes.** The Add artifact dialog: `SourcePageArtifactsView.addForm`, `SourcePageView`'s `.pvFormDialog(isPresented: addArtifactPresented, …)` and `addArtifactPresented`; `SourceArtifactsSection.Draft` and every member in D-12; `FilePickTarget`, `applyDroppedURLs(_:into:)`, `applyPickedFile(url:into:)`, `setReject`; `ArtifactSheetModel.artifact`; strings `addArtifactDialogSubtitle`, `artifactLabelHint`; a11y ids `sources.page.addArtifact.label`, `.description`, `.graphs`, `.graphs.none`, `.ingestReject` and the `sources.page.addArtifact` dialog prefix (the header button's own id `sources.page.addArtifact` stays); the tests that drive `model.artifacts.draft` (ported, § 8).
**Next adds.** S15-11's graph form makes the same suggestion read for `level: .graph` in its create mode (D-2's shape: read, `unsavedRows`, open). S15-12 and S15-19 make the suggestions real for the shipped types; nothing here changes when they land.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A sheet with no Artifact (create mode before the first Save) | `SourceArtifactsSection.workspaceChanged()`; `ArtifactSheetModel.graphsChanged()`, `loadLinkLocks()`, `savedGraphIDs`, `linkedGraphIDs` | `workspaceChanged` reads `artifactID` and leaves a nil-id sheet open (AC-4.5, T-11); `loadLinkLocks` returns at once in create mode; `graphsChanged` starts no read (T-8) |
| A create sheet whose Artifact exists (after a refused file) | `workspaceChanged()` | Closes with the toast when the made Artifact leaves the page (AC-4.4, T-11) |
| The page drop opens the sheet after an `await` | `stagePageDrop`, `acceptsArtifactPageDrop` | `isOpeningAdd` refuses a second drop and the highlight (T-10); a sheet that opened meanwhile is never given the file (D-4) |
| The board in a sheet on rows no owner holds | `MetadataBoardModel`, `MetadataDraftHost` | Already the draft host's case (S15-36); the add row and refusal mapping work on suggested rows (T-4, T-6) |

### Scale

At the map's Scale (a Source with up to 60 Artifacts): an Add costs one suggestion read (one type's `artifact` rows, a handful), one `createArtifact`, at most one `updateArtifact` on a retry, at most one ingest, and today's one workspace reload in `refreshCoverFromStore`. `show` appends one Artifact and writes its rows into the cached key once through `.wroteMetadataBoard`; no per-row work is added to the list. No budget is set or leaned on.

### Edges

- **Assumptions.** `MetadataDraftHost.draft` is nil while the board equals what it opened with; in create mode it opened with the suggestions, so an untouched board sends no draft and the engine writes no layout (Go `TestCreateArtifactMetadata`'s plain create). T-5 holds it. A dismissed suggestion is a change, so it sends a draft with the dismissed layout row. T-5 holds that too.
- **Neighbors.** A type with no `artifact` suggestions opens an empty board (T-2's second case). A Source whose type has suggestions only at the `source` level shows none (the read is `.artifact`; T-2 asserts the level). A retry with nothing changed sends no update; a retry with one changed value sends one (T-7).
- **Removed inputs.** A suggested field deleted from the project between the read and Save: the engine refuses it naming the field, and the row shows the refusal (T-6's path). A graph deleted while the sheet is open drops out of `liveGraphIDs`, so it is neither shown nor sent (today's rule, unchanged). The made Artifact deleted elsewhere closes the sheet (T-11).
- **Real data.** None — no table-driven rule changes; the seed has no `artifact` suggestions yet (S15-19).
- **Untrusted size.** None — the suggestion list is the project's own catalog rows for one type, read once.

### Schema

None — no catalog change.

### Go core

None — `CreateArtifact` already takes the draft and answers the board (S15-08, S15-33).

### FFI

None — no proto change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Artifact sheet, add mode | `PVFormDialog` (S15-34's scrolling body) | `macos/App/Features/Sources/ArtifactSheet.swift` | Title "Add artifact", subtitle "<Source ref> · <Source title>", confirm "Add artifact", width 560, `isDirty`/`isRunning`/`confirmDisabled` from the model |
| File | `IngestFileDropRow` (snowflake, Ship) + `PVCallout` | same file, `ArtifactSheetForm` | D-13; create mode only, first in the `.top` group |
| Label, Description | `PVField` + `PVInput` / `PVTextArea` | same | The edit mode's, unchanged |
| Evidence graphs | `SourceFormLoadArea` + `SourceArtifactGraphLinks` | same | `isDisabled: model.isSaving \|\| model.graphsLocked`; `locked:` is empty in create mode |
| Copy details | `MetadataBoardView(placement: .form)` | same | The edit mode's; starts with the suggestions empty; frame `sheet-no-sugg` when none |
| Model | `ArtifactSheetModel` | `macos/App/Features/Sources/ArtifactSheetModel.swift` | D-1, D-5 – D-10 |
| File choice | `IngestFileChoice` | `macos/App/Features/Sources/IngestFileChoice.swift` *(new)* | D-11 |
| Opening | `SourceArtifactsSection.openAdd()`, `stagePageDrop(_:)` | `macos/App/Features/Sources/SourceArtifactsSection.swift` | D-2 – D-4 |

States: empty board (only Add field), saving (`isRunning`, fields and board disabled), a refused value (its row), a failed save (top Callout), a refused file (File Callout, boxes disabled), the graph list loading (today's loading line). The board is readable at Regular and Wide alike: the sheet is the same at both widths; the page behind it is the Wide Source page (the map's **Wide** line). Accessibility: the File row's buttons and the Callout keep today's labels; Label takes focus on open; the kit's Escape order (SA-15) needs no code.

### Strings

L10n keys *(new)*: none. Removed: `L10n.Sources.addArtifactDialogSubtitle` (`sources.page.addArtifactDialogSubtitle`), `L10n.Sources.artifactLabelHint` (`sources.page.artifactLabelHint`). Comments updated per D-14. Old-scope copy: the comments naming "the Add artifact dialog" are the only copy that named the old flow; no visible string named it.

## 7. Steps

**STEP-1 — FakeStore hooks**
Files: `macos/App/Platform/FakeStore.swift`. Skill: `add-swift-test`.
Do: D-15.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/SourcePageModelTests` passes (nothing reads the hooks yet).

**STEP-2 — One file choice**
Files: `macos/App/Features/Sources/IngestFileChoice.swift` *(new)*, `macos/App/Features/Sources/SourceArtifactsSection.swift` (`AttachDraft`, the attach methods), `macos/App/Features/Sources/SourcePageArtifactsView.swift` (`addFileForm`), `macos/ProvenenciaTests/IngestFileChoiceTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-11 for the attach path only. The create path still uses `Draft`'s fields and a `case create` arm until STEP-4; keep `applyPickedFile(url:into:)` working for `.create` by having it call `IngestFileChoice` on a local value and copy the three fields into `draft` (STEP-4 deletes it). Write T-1.
Check: T-1 and the Add file tests in `SourcePageModelTests` pass.

**STEP-3 — The sheet's add mode**
Files: `macos/App/Features/Sources/ArtifactSheetModel.swift`. Skill: `add-swift-test`.
Do: D-1, D-5, D-6, D-7, D-8, D-9, D-10 in the model. Edit mode's behavior is unchanged: its `save()` body moves into `saveEdit()` *(new, private)* untouched.
Check: the edit sheet's tests in `SourcePageModelTests` pass after their `sheet.artifact.id` reads become `sheet.artifactID` (the only test edit in this step).

**STEP-4 — Open the sheet in add mode; delete the dialog's model**
Files: `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourcePageModel.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skill: `add-swift-test`.
Do: D-2, D-3, D-4, D-5's `workspaceChanged`, D-12's section deletions, `FilePickTarget`/`applyDroppedURLs(_:into:)`/`applyPickedFile(url:into:)`/`setReject`. Port every test in § 4's Tests row onto the sheet with a private helper `openAddSheet(_ model: SourcePageModel) async throws -> ArtifactSheetModel` *(new)* (`await model.artifacts.openAdd()`, `#require(model.artifacts.sheet)`): `draft.label` → `sheet.draft.label`, `draft.filePath = path` → `sheet.file.choose(URL(fileURLWithPath: path))` for a file that exists, or `sheet.file.filePath = path` where the test needs a path the policy has not seen (today's tests set it directly), `create()` → `_ = await sheet.save()`, `isAdding` → `model.artifacts.sheet != nil`, `draftGraphIDs` → the recorded `store.createArtifactGraphIDs`. Each ported test keeps its name and its assertions on the page, the cover and the store. Write T-2 – T-11.
Check: `SourcePageModelTests` and `IngestFileChoiceTests` pass. `rg -n "artifacts\.(draft|isAdding|create\(|cancelAdd|canSubmitDraft|addSheetIsDirty|isSavingDraft|draftLabelError)" macos` has no hit outside `SourcePageArtifactsView.swift` and `SourcePageView.swift` (STEP-5 removes those).

**STEP-5 — The view**
Files: `macos/App/Features/Sources/ArtifactSheet.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Features/Sources/SourcePageView.swift`. Skill: `add-ui-component` (read only).
Do: D-10's modifier copy; D-13's File section; D-6's `isDisabled`; D-12's view deletions; the Add buttons and the drop call D-4's async paths.
Check: the app builds; the same `rg` as STEP-4 has no hit anywhere in `macos`. In the app, Add artifact opens the sheet with the File row first (frame `sheet-add`).

**STEP-6 — Strings**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-14.
Check: `rg -n "addArtifactDialogSubtitle|artifactLabelHint" macos .agents .cursor docs/*.md` has no hit; the build has no missing-key warning.

**STEP-7 — Version**
Files: `VERSION`, `core/version.go`, every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj`.
Do: the header's bump.
Check: the three agree (`cat VERSION`; `rg -n "MARKETING_VERSION|const Version" macos/Provenencia.xcodeproj/project.pbxproj core/version.go`).

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row), this plan (status, Landed note), `docs/deployment-plan/spike-15/design/README.md` (S15-D14 stays in `design/` until S15-25 lands, its handoff; note S15-10 landed).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-10-artifact-sheet-add-mode.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Ask for an Artifact's copy details when it is added."). No WIP commits.

## 8. Tests

Swift Testing, `@MainActor`, `FakeStore` and `SourcePageModel` as the existing `SourcePageModelTests` build them. A source type id for the suggestion fixture is the one `makeStore()` gives the Source; seed `store.suggestionsByLevel[.artifact, default: [:]][typeID]` with three `CatalogTypeSuggestion`s (Provider, URL, Accessed) over fields in `store.fieldsByProject[projectDir]`.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/IngestFileChoiceTests.swift` *(new)* `chooseAcceptsAnImage`, `chooseRefusesASpreadsheetWithoutTheEngine` (moved from `applyPickedFileRejectsSpreadsheetWithoutFFI`), `twoURLsAreRefused`, `refuseKeepsTheNameAndDropsThePath`, `clearEmptiesTheChoice` | a temp `.png` → `filePath` set, `reject` nil; `notes.xlsx` → `filePath` nil, `fileName` "notes.xlsx", `reject` set; two URLs → multi-file `reject`, `fileName` the first's; `refuse(callout)` after a good choice → `filePath` nil, `fileName` kept; `clear()` → all nil | AC-2.5, AC-2.6, D-11 |
| T-2 | Swift | `SourcePageModelTests.addOpensTheSheetWithTheTypesArtifactSuggestions` *(new)* | three `artifact` suggestions → `sheet.board` rows are Provider, URL, Accessed, empty, `suggested`; `store.listTypeSuggestionsLevels == [.artifact]`; `sheet.title == "Add artifact"`; `sheet.canSave == false`; with no suggestions scripted → no rows | AC-1.1, AC-1.4, D-2 |
| T-3 | Swift | `SourcePageModelTests.aFailedSuggestionReadOpensNoSheet` *(new)* | `listTypeSuggestionsError` set → `openAdd()` returns nil, `model.artifacts.sheet == nil`, the page error is set | AC-4.3, D-2 |
| T-4 | Swift | `SourcePageModelTests.addSavesTheCopyDetailsWithTheArtifact` *(new)* | `nextArtifactID = "a-new"`, `metadataWriteResults[artifact a-new]` scripted; label "Sheet 26 (image 13)", Provider set through the board → `save()` true; `createArtifactMetadata.last` holds one value for Provider; `updateArtifactCalls == 0`; the page's `copyDetails(for: "a-new")` equals the scripted rows; the cached `.sourceWorkspace` value holds the same rows | AC-1.2, D-7 |
| T-5 | Swift | `SourcePageModelTests.anUntouchedBoardSendsNoDraftButKeepsTheRows` *(new)* | label only → `createArtifactMetadata == [nil]`; the page holds the reply rows for the new id; a second case dismisses URL first → the draft's layout has URL `dismissed: true` | AC-1.3, Edges |
| T-6 | Swift | `SourcePageModelTests.aRefusedValueCreatesNothing` *(new)* | `createArtifactError = catalogMetadataRefusal(params: [urlFieldID])`, URL filled → `save()` false; `sheet.board.editingFieldID == urlFieldID` and its row error set; `model.artifacts.items` unchanged; `sheet.artifactID == nil` | AC-1.5, D-8 |
| T-7 | Swift | ported `aRefusedFileKeepsTheCreatedArtifactAndARetryAttachesToIt`, `addArtifactRetryIgnoresGraphChoice`; `aRetryWithChangesUpdatesOnce` *(new)*, `aRetryWithNoChangeOnlyIngests` *(new)* | `ingestArtifactFileError` = `ingest.invalid` → `save()` false, the Artifact on the page, `sheet.file.reject` set, `sheet.graphsLocked`, `canSave` false; clear the error, choose a good file and change Provider → `save()` true, `createArtifactGraphIDs.count == 1`, `updateArtifactGraphIDs == [nil]`, `updateArtifactMetadata.last` non-nil; with no change → `updateArtifactCalls == 0` | AC-2.2, AC-2.3, AC-2.4, D-6, D-7 |
| T-8 | Swift | ported `addArtifactOnOneGraphShowsTheGroup`, `addArtifactLinksOnlyTheCheckedGraphs`, `addArtifactWithNoGraphChecked`; `aGraphMadeWhileAddingIsLinked` *(new)* | `["g-1"]` → sent `["g-1"]`; no graph list → sent `nil`; exclude g-b and g-s → `["g-j"]`; exclude all → `inNoGraph`, sent `[]`; add graph g-n to the session after open → sent `["g-b", "g-j", "g-s", "g-n"]`; `linkLoadState == .ready` throughout | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-6 |
| T-9 | Swift | ported `addArtifactExcludedGraphIsDirty`, `createArtifactRequiresLabel`; `addModeDirtAndEnablement` *(new)* | fresh → `!isDirty`, `!canSave`; label typed → both true; label cleared, a file chosen → `isDirty`; a refused file → `!canSave`; a filled suggestion → `isDirty`; an excluded graph → `isDirty`; `save()` with a blank label → `labelError` set, no store call | AC-4.1, AC-4.2, D-9 |
| T-10 | Swift | ported `pageDropStagesFileWithoutCreating`, `pageDropOfTwoFilesOpensRejectCallout`; the drop half of the test that asserts `!model.acceptsArtifactPageDrop` with an edit sheet open | one file → a create sheet with `file.fileName` and `filePath`, no store write; two files → `file.reject` set, `!canSave`; with an edit sheet open → the drop changes nothing and the sheet is still the edit sheet | AC-2.1, AC-2.5, D-3, D-4 |
| T-11 | Swift | `SourcePageModelTests.anAddSheetCloseRule` *(new)* | open add, change the workspace (`context.setWorkspace` with a note added) → the sheet stays; make the Artifact with a refused file, then remove it from the workspace → the sheet closes and the toast body names its label | AC-4.4, AC-4.5, D-5 |
| T-12 | Swift | ported `createFilelessArtifact`, `createArtifactWithPDFLeavesTypeIconCover`, `creatingAnArtifactWithAFileOpensTheGraphJump`, `useAsThumbnailAndRevertToTypeIcon`, `filelessArtifactCannotBeCover`, `createFilelessArtifactClearsMIMECoverForTypeIcon`, `ingestMissingFileSurfacesIngestInvalid`, `ingestThenRejectSecondAttach`, `aBoardWriteKeepsAnArtifactShownOnlyOnThePage` (where it adds through the dialog) | each keeps today's assertions on the page, the cover, the toast and the store, driven through the sheet | AC-1.2, AC-2.1, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests` (the sheet and Add file tests), `MetadataDraftHostTests`, `MetadataDraftHostParityTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (addArtifactDialogSubtitle, artifactLabelHint gone)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`SourceArtifactsSection.Draft`, `isAdding`, `addForm`, `applyPickedFile`, `applyDroppedURLs`, `FilePickTarget`, `ArtifactSheetModel.artifact`, `sources.page.addArtifact.ingestReject`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] design/README.md: S15-D14's row notes S15-10 landed (the brief stays until S15-25)
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The seed has no `artifact`-level suggestions yet (S15-19), and the type page cannot set them (S15-12), so step 1 adds three by hand. The slice Check ("Adding a downloaded census image offers provider, URL and accessed date, and the values show on the Artifact afterwards") is true after step 4.

1. Open a project with a Source of the type "Census" and at least two Evidence graphs. On the Metadata page, add fields "Provider" (text), "URL" (URL) and "Accessed" (text) if they are not there. Quit the app.
2. In Terminal, with the project's `provenencia.sqlite`:
   ```bash
   sqlite3 "<project folder>/provenencia.sqlite" "INSERT INTO source_type_metadata_fields (source_type_id, field_id, level, sort_order) SELECT t.id, f.id, 'artifact', CASE f.label WHEN 'Provider' THEN 0 WHEN 'URL' THEN 1 ELSE 2 END FROM source_types t, source_metadata_fields f WHERE t.label = 'Census' AND f.label IN ('Provider', 'URL', 'Accessed');"
   ```
3. Reopen the project and the Census Source. Click **Add artifact**: the sheet is titled "Add artifact", File first, then Label, Description, Evidence graphs (all checked) and the board with Provider, URL, Accessed empty. The button is disabled.
4. Choose a downloaded `.jpg`. Label "Sheet 26 (image 13)", Provider "FamilySearch", URL "https://example.org/ark/1", Accessed "14 November 2026". Add artifact: the sheet closes, the toast names the new ref, and the Artifact's row shows the three values and its thumbnail. Open its edit sheet: the three values are there.
5. Add artifact again; type a label, uncheck one graph, press Escape: the discard confirm shows; keep editing, then Cancel: nothing is added.
6. Drop two files on the page: the sheet opens with the "one file at a time" Callout and Add disabled. Clear it, choose an `.xlsx`: the refusal Callout shows.
7. Widen the window to Wide: the sheet is the same; the page behind it is the Wide Source page.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A test ported from the dialog loses an assertion | Each keeps its name and its assertions (STEP-4); T-12 lists them |
| The create sheet reaches edit-only code with no Artifact | D-5 routes edit-only code through `editedArtifact`; Reach lists each path; T-8, T-11 |
| A retry writes an empty revision | D-7 compares with `sent`; T-7's `aRetryWithNoChangeOnlyIngests` |

**Stop and report if:**

- `CreateArtifactResponse.metadata` is empty for a plain create on a type with `artifact` suggestions (it must carry them; Go `TestCreateArtifactMetadata` asserts it).
- `MetadataDraftHost.draft` is non-nil for a board opened from `unsavedRows(from:)` and left untouched.
- `PVFormDialog`'s `onConfirm` closes the sheet on a `false` answer, so a refused file or value would close it.
- A test needs `FakeStore` to apply a draft or judge a value to pass.
- The board's `sheet-add` frame and SA-7 disagree on the order of File and Label.

## 12. Open questions

None.

## 13. PR

**Title:** S15-10 — Adding an Artifact asks for its copy details
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence (test names); deviations from the plan; the manual check, including the SQL step; `Plan: docs/deployment-plan/spike-15/prs/S15-10-artifact-sheet-add-mode.md`.
