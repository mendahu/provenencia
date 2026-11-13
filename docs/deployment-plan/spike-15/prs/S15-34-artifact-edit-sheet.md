# S15-34 — Artifacts are edited in a sheet

**Status:** Landed in [#819](https://github.com/mendahu/provenencia/pull/819) (2026-11-13)
**Spike / slice:** Spike 15 · Slice 2 — Metadata at three levels
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S15-D14](../design/S15-D14-source-page-artifacts.md) SA-4, SA-6, SA-8, SA-9, SA-12 – SA-15 and the Wide rows for the Metadata section and the sheet, board agreed 2026-11-13; plus the user's 2026-11-13 additions on the same board (the Metadata section header: help tooltip, no field count)
**Board:** [Claude Design — Source Detail](https://claude.ai/artifact/443NPNYmGErAmxgV1fBsFK), page *Source page*: rows "Artifact sheet" (`sheet-edit`, `sheet-editing`, `sheet-discard`, `sheet-saving`, `sheet-refused`, `sheet-failed`, `sheet-eight`, `sheet-no-sugg`, `art-loading`, `art-only`, `art-twelve`, `art-error`, `art-refused`), "Artifacts" (`art-gone`) and "Metadata" (`meta-add-row`, `meta-add-row-wide`); decisions panel entry "S15-D14". The implementer reads it with the Artifact tool's `read` on that link, `path` "project/Source.dc.html" (the sheet at `PVFormDialog open="{{ sheetOpen }}"`, `sheetBoardEl`, `addRowEl`, `valueEl`, `srcAddVals`, `metaHelp`) and "project/Source page.dc.html" (the frame list). §6 macOS describes each frame in words, so the plan stands without the board
**Depends on:** S15-09 (landed [#802](https://github.com/mendahu/provenencia/pull/802), trimmed), S15-33 (landed [#806](https://github.com/mendahu/provenencia/pull/806)), S15-36 (landed [#810](https://github.com/mendahu/provenencia/pull/810)), S15-D14 (board agreed 2026-11-13)
**Blocks:** S15-35, S15-10, S15-11
**Branch:** `s15-34-artifact-edit-sheet`
**Version bump:** minor, 0.11.1 → **0.12.0** (or the next minor after `main`'s `VERSION` when you rebase): `VERSION`, `core/version.go`, and every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj` — a backward-compatible feature on a screen (an Artifact is edited in a sheet with its copy details; every board adds a field inline); no wire or catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (the kit changes in STEP-1; the sheet, the add row and the copy details stay Source page snowflakes), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (read only: the Save patches the page's cached workspace through `setWorkspace` and `.wroteMetadataBoard`, as § 6 "Metadata values" already says)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "An Artifact is added and edited in one sheet", "A board adds a field in an inline row", "One metadata board", "Metadata saves with its owner's write", "One brief, one PR" (S15-D14 predates it and is exempt) ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); the map's S15-34 scope entry and its **Also ships** line; [S15-D14](../design/S15-D14-source-page-artifacts.md) § 4, § 4.1, § 6 and its **Board** line (the user's calls of 2026-11-13); S15-36's [plan](S15-36-metadata-board-holds-a-draft.md) **Next adds** (the recipe this plan follows) and its **Landed** note; S15-33's [plan](S15-33-artifact-update-saves-metadata.md) D-2 and D-6; S15-09's [plan](S15-09-artifact-metadata-boards.md) **Landed** note (the trim: `.wroteMetadataBoard`, `metadataRows(for:)`, `setWorkspace`, `patchArtifacts` stayed)
**Planned:** 2026-11-13 against `main` @ `b9a881b`
**Review:** [`S15-34-artifact-edit-sheet.review.md`](S15-34-artifact-edit-sheet.review.md) — Fits with changes — Ready with changes, 2 rounds; every blocker Fixed; Cancel and unconfirmed-text Save answered by the user
**Size:** 11 steps across the Artifacts section, the Metadata section and the kit (form dialog, sheet, dismiss rule, inline edit, icons), past the plan-pr limit of about 10. The user chose to keep it one PR over a three-way split (2026-11-13). Work the Steps in order and run each check; the commits are one per step
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

An Artifact is edited in one sheet instead of on its card: label, description, the Evidence graphs it is in, and its copy details (provider, URL, accessed date on a census), with one Save that writes everything and a Cancel that discards everything. Each card shows its saved copy details read-only, so a researcher sees where each copy came from without opening anything. Every metadata board, the Source's own included, adds a field in an inline row instead of the Add metadata dialog. The Metadata section's header loses its field count and moves its long intro into a help tooltip beside the title.

**Why now.** The user rejected S15-09's Copy details on each card (cards grew too tall at 60 Artifacts); S15-D14 re-cut the work, and S15-33 and S15-36 landed the engine write and the draft board this PR draws (map review, user 2026-11-12). S15-35 (the list layout), S15-10 (add mode) and S15-11 (the graph form) all build on the sheet, the inline add row and the kit's scrolling body.

**What stays.** The card grid, its preview well, its size and its order (S15-35 replaces them). The Add artifact dialog and the Add file dialog (S15-10 replaces the first). Set cover / Use as thumbnail, Delete and its DeleteImpact confirm, the file open, the page file drop. The Source's own board rows: they still save the moment a row is confirmed, and a clear still asks first. The Evidence graphs section and the graph form.

## 2. User stories

**US-1.** As a researcher, when I press an Artifact's pencil, I edit its label, description, graphs and copy details in one sheet and save them together, so that nothing is half saved.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Census Source, Artifact `ART-7QK2M` "Sheet 14 (image 1)" with Provider "Ancestry", URL, Accessed saved; the type suggests Image number at the `artifact` level. Press its pencil | A sheet titled "Edit ART-7QK2M", 560pt wide, shows in order: Label "Sheet 14 (image 1)" (required), Description, Evidence graphs (its checkboxes), Copy details (the three saved rows, "Suggested by this type" with Image number, the "Add field" button). No file row |
| AC-1.2 | In the sheet, change the label to "Sheet 14 (recto)", fill Image number "473 of 642", add Film number "1254392" through the add row, drag URL above Provider, press Save artifact | One `updateArtifact` call carries label "Sheet 14 (recto)", the unchanged description, `graphIDs: nil`, and a `metadata` draft whose values are URL, Provider, Accessed, Image number, Film number in that order; the sheet closes; the card shows "Sheet 14 (recto)" and its copy details; no Source workspace reload (`getSourceWorkspace` not called) |
| AC-1.3 | Open the sheet and press Save artifact with nothing changed | No write; the sheet closes |
| AC-1.4 | Clear the label and press Save artifact | No write; "Give the artifact a label." under Label; the sheet stays open |
| AC-1.5 | The Robins graph has a Citation drawn on the Artifact | Robins is checked and locked with "1 citation in this graph uses this artifact" (today's hint); unchecking other graphs works; while the lock read runs, "Reading the evidence graphs…" shows and Save is disabled; a failed read shows today's compact danger Callout with Read again |
| AC-1.6 | Uncheck Hale and save | `updateArtifact` carries `graphIDs` with every checked graph and every graph made since the sheet opened that links the Artifact; the graph list reloads through the effects |

**US-2.** As a researcher, nothing in the sheet saves until I press Save artifact, and leaving with changes asks first.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Fill a suggestion, clear a saved row, dismiss a suggestion, reorder, add a field | No store write happens until Save artifact; clearing a row asks nothing |
| AC-2.2 | Change anything, then press Escape or click outside | The kit's standard "Discard changes?" confirm ("Your changes will be lost.", Discard / Keep editing). The visible Cancel closes at once, as on every dialog (user, 2026-11-13) |
| AC-2.3 | Type in a row and put the original text back, then Cancel | Closes at once (unchanged) |
| AC-2.4 | A row is in edit, or the add row is open, or a suggestion holds typed text, and press Escape | Only that field reacts: the row edit cancels, the add row closes, the suggestion's text clears; the sheet stays. With nothing open and nothing changed, Escape closes the sheet |
| AC-2.5 | A row is in edit and press Return | The row's value is confirmed into the draft; the sheet does not save |
| AC-2.6 | Type "473 of 642" into the Image number suggestion without pressing its Save, then press Save artifact. Then press the suggestion's ✕ and Save artifact again | First: no write; the top Callout says "Finish the open field first: save its value or clear it."; the typed text stays (user, 2026-11-13). Second: Image number is dismissed, nothing blocks, one write |

**US-3.** As a researcher, a failed save keeps everything I typed and tells me where it failed.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Save while the store holds the write | Save artifact shows its loading state; Label, Description, the graph boxes and the board are disabled |
| AC-3.2 | The engine refuses the URL value (`sourcemetadata.invalid`, `params[0]` = the URL field's id) | "This metadata change can't be saved." under the URL row (open for edit, typed text kept); the sheet stays open |
| AC-3.3 | The engine refuses for a graph lock (Hale was unchecked and a Citation took the Artifact since the sheet opened) | Top danger Callout "Couldn't save the artifact" with "Hale household, dwelling 113 keeps this artifact. Your changes are kept. …"; the locks are read again and Hale shows locked |
| AC-3.4 | Any other failure (I/O) | Top danger Callout "Couldn't save the artifact", "Your changes are kept. …"; the sheet stays open |
| AC-3.5 | Another window deletes the Artifact while its sheet is open | When the page's workspace drops it, the sheet closes with its draft and the page shows the info Toast "Artifact no longer here" / "“Sheet 14 (image 1)” was deleted or moved in another window." |

**US-4.** As a researcher, each card shows its saved copy details at a glance, and its shortcuts open the sheet where I need it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Cards for Artifacts with 0, 3 and 8 copy details | 0: nothing; 3: three lines, field name then value; 8: four lines and "+4 more". A URL value is a link; each value is one line, truncated |
| AC-4.2 | Press "+4 more" on Sheet 17 | The sheet opens scrolled to Copy details |
| AC-4.3 | Press "Link to a graph" on Sheet 20 (in no graph) | The sheet opens scrolled to Evidence graphs |
| AC-4.4 | VoiceOver on a card with copy details | The copy details read as "Provider, Ancestry" and so on; the pencil reads "Edit ART-7QK2M" |

**US-5.** As a researcher, every board adds a field in an inline row, and the Metadata header is quieter.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | Source page, Metadata header | Title "Metadata", a help icon beside it whose tooltip is the intro text, no "n fields" count, a ghost "Add field" button on the right; no intro paragraph under the header |
| AC-5.2 | Press the header's Add field | An inline row opens at the foot of the Source's board: a field picker (fields not on the board), a value input, ✓ disabled until both are filled, ✕. ✓ saves at once and closes the row; the new row shows on the board |
| AC-5.3 | The engine refuses the value | The refusal shows under the add row; the row stays open with its text |
| AC-5.4 | The Source has no metadata and no suggestions | The empty state shows its title and message and no button; the header's Add field opens the row below it |
| AC-5.5 | The add row has "Collection" picked; another window adds Collection to this Source | The picker clears, the row says "Collection was added in another window. Pick another field.", and VoiceOver announces it; the typed value stays |
| AC-5.6 | Wide (pane 2296): the Metadata section in the 480pt side column with the row open | The row fits the column: picker 150pt, value fills, ✓ ✕ |
| AC-5.7 | The Add metadata dialog | Gone: no `sources.page.addMetadata` sheet, no `MetadataBoardAddForm` |

## 3. Scope

| In | Covered by |
| --- | --- |
| Kit: `PVFormDialog` opt-in scrolling body (fixed title and footer) with an initial scroll target; `PVSheet` stands down Escape and the form's Return while content reports an open inner edit (SA-6, SA-15) | STEP-1, AC-2.4, AC-2.5 |
| One metadata value view for board rows and read-only copy details (inventory "Metadata value"); the draft host made observable | STEP-2, AC-4.1, AC-2.1 |
| The inline Add field row on every board; the Add metadata dialog removed; the "field taken" notice (SA-13; S15-36 review UX-2) | STEP-3, AC-5.2 – AC-5.7 |
| The Metadata header: help tooltip, no count, ghost Add field (user, 2026-11-13); Wide | STEP-4, AC-5.1, AC-5.6 |
| The Artifact sheet's model in edit mode: draft, link locks, draft board, dirty, one-write Save, refusal mapping (SA-6, SA-8, SA-9) | STEP-5, AC-1.2 – AC-1.6, AC-2.1 – AC-2.3, AC-2.6, AC-3.1 – AC-3.4 |
| The Artifact sheet's view and presentation (SA-6, SA-9, SA-15) | STEP-6, AC-1.1, AC-1.5, AC-2.4, AC-2.5 |
| The card opens the sheet; card edit removed; the Artifact gone closes the sheet (SA-12, SA-14) | STEP-7, AC-3.5, AC-4.2, AC-4.3 |
| Read-only copy details on today's cards; "+n more" (SA-4, SA-11) | STEP-8, AC-4.1, AC-4.4 |
| Strings swept, tests ported, version bump, docs | STEP-9 – STEP-11 |

| Out | Home |
| --- | --- |
| The list's new layout (rows, 64pt thumbnail, copy-details column widths, Wide copy-details column), the row menu, the type pill, the header row clip, the description size | S15-35 |
| Add mode: the sheet with File first; the Add artifact dialog removed | S15-10 |
| The graph form's board (it reuses `.form` placement and the scrolling body) | S15-11 |
| Move to another Source… on an Artifact | S15-25 |
| Flagging fields a citation needs on a board | S15-39 |
| Unknown and Supplied on a row | S15-41 |
| Reordering saved rows by keyboard | [`docs/ideas/keyboard-reorder.md`](../../../ideas/keyboard-reorder.md) |
| Filling copy details from a URL | [`docs/ideas/registry-source-import.md`](../../../ideas/registry-source-import.md) |
| A Swift test of the sheet's SwiftUI layout or key routing beyond the pure rules in T-1 | not planned: `swift-tests.mdc` tests models and pure rules; the manual check covers keys |

## 4. Current state

Verified at `b9a881b`.

| Area | Where | Today |
| --- | --- | --- |
| Card edit | `macos/App/Features/Sources/SourceArtifactsSection.swift` `editingArtifactID`, `pendingEditSwitch`, `editGate`, `labels`, `descriptions`, `fieldErrors`, `savingID`, `cardSaveError`, `beginEdit`, `confirmEditSwitch`, `cancelEditSwitch`, `openEdit`, `cancelFields`, `fieldsDirty`, `textChanged`, `canSaveFields`, `cardLabelError`, `saveFields`, `saveErrorBody`, `replace` | One card in edit at a time; `saveFields` sends `updateArtifact(… metadata: nil)` and applies `context.applyWrite(effects)` (`.mutatedSourceWorkspace`, which reloads the page). Goes |
| Link locks | same file, `LinkEdit`, `linkEdit`, `linkLoadTask`, `startLinkEdit`, `awaitingGraphs`, `graphsChanged`, `dropLinkEdit`, `linkedGraphIDs`, `liveGraphIDs`, `lockedGraphs`, `checkedGraphIDs`, `linkEditGraphs`, `cardInNoGraph`, `linksDirty`, `loadLinkLocks`, `savedGraphIDs` | The engine's unlink preflight (`getEvidenceGraphUnlinkImpact`) for the card in edit; locks, loading, failure, graphs made or deleted during edit. Moves into the sheet's model (D-4) |
| Add dialog | same file, `Draft`, `isAdding`, `draft`, `draftLabelError`, `isSavingDraft`, `openAdd`, `create`, `stagePageDrop` (calls `cancelFields`) | Stays (S15-10); `stagePageDrop` stops calling `cancelFields` |
| Seeding | same file, `seedDrafts()` (called from `SourcePageModel.apply` when `artifacts` changes) | Seeds `labels`/`descriptions` and drops a gone card's edit. Replaced by `workspaceChanged()` (D-11) |
| Patch helpers | same file, `patchArtifacts(_:)`, `show(_:effects:)` | `patchArtifacts` sets the page workspace with `context.setWorkspace`. Kept; `show` loses its `labels`/`descriptions` lines |
| Card view | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `artifactCard`, `metaRow`, `restingText`, `noGraphRow`, `descriptionLine`, `graphLinks`, `graphLinksGroup`, `editFields`, `labelBinding`, `descriptionBinding`, `commitEdit` | The card switches to fields in edit. The edit branch goes; `noGraphRow`'s button calls `beginEdit` |
| Card pencil label | `macos/App/Platform/L10n.swift` `L10n.Sources.editArtifactLabel(ref:)` | "Edit label and description of %@". Becomes "Edit %@" (SA-11) |
| Board view | `macos/App/Features/Sources/MetadataBoardView.swift` `MetadataBoardView` (`onAdd`, empty state with an Add button), `MetadataBoardAddForm`, `MetadataBoardTextEditor.displayValue`, `MetadataBoardValueChrome`, `MetadataBoardLabel` | The empty state's button calls `onAdd` (the dialog). `displayValue` draws a value (mono caption, URL as an underlined link button) |
| Board model | `macos/App/Features/Sources/MetadataBoardModel.swift` `MetadataBoardAddRow`, `addRow`, `isSavingAddRow`, `canConfirmAddRow`, `openAddRow`, `closeAddRow`, `confirmAddRow`, `resetDrafts` (clears a picked field the rows now hold, silently), `showRefusal`, `hasTypedEdit` (private), `fieldComboOptions` | The add row's state exists (S15-36); no view draws it inline |
| Dialogs | `macos/App/Features/Sources/MetadataBoardDialogs.swift` `MetadataBoardDialogs` (`addTarget`, `isAdding`, `addFieldError`, `addFieldID`, `addValue`, `addValueError`, `isSavingAdd`, `addSheetIsDirty`, `canSubmitAdd`, `fieldComboOptions`, `openAdd`, `cancelAdd`, `createFromAdd`, `askClear`, `confirmClear`, `pendingDelete`, `isClearing`) | Presents the add row in a dialog; keeps the clear confirm |
| Draft host | `macos/App/Features/Sources/MetadataDraftHost.swift` `MetadataDraftHost` (`isChanged`, `draft`, `confirmsClear == false`) | A plain `final class`, not `@Observable`: a view over a draft board does not see its rows change. Made observable (D-18) |
| Live host | `macos/App/Features/Sources/MetadataBoardHost.swift` `SourcePageMetadataHost` | Unchanged |
| Metadata section | `macos/App/Features/Sources/SourcePageMetadataView.swift` `SourcePageMetadataView` | Header meta `metadataFieldCount`, primary "Add field" (`L10n.Sources.addMetadata`, comment says "dialog") opening the dialog, `metadataIntro` paragraph under the header |
| Page | `macos/App/Features/Sources/SourcePageView.swift` `.pvFormDialog(isPresented: addMetadataPresented …)` with `MetadataBoardAddForm`, `addMetadataPresented`, `.pvConfirm(item: … pendingEditSwitch …)` | Both go |
| Page model | `macos/App/Features/Sources/SourcePageModel.swift` `acceptsArtifactPageDrop` (reads `metadataDialogs.isAdding`, `artifacts.pendingEditSwitch`), `apply(_:)` (calls `artifacts.seedDrafts()`) | Updated |
| Graph form precedent | `macos/App/Features/Sources/SourceGraphFormModel.swift` `SourceGraphFormModel` (`Mode`, `draft`, `prefill`, `isDirty`, `loadKeptLinks`, `save() -> Bool`), `SourceGraphFormDialog.swift` `View.sourceGraphFormDialog(_:)` | The shape the sheet copies |
| Kit form dialog | `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift` `PVFormDialogContent`, `pvFormDialog(item:…)` | No scrolling body; Confirm has `.keyboardShortcut(.defaultAction)` always |
| Kit sheet | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` `PVSheetChrome` (a zero-size Escape button, `.keyboardShortcut(.cancelAction)`; discard copy `.discardChanges`); `PVSheetDismiss.swift` `PVSheetDismiss.action(isRunning:isDirty:isConfirmingDiscard:)`, the one rule for Escape and an outside click | Escape always goes to the sheet, ahead of an inline edit's own `.cancelAction` button (`PVInlineEdit.swift`, lines 282, 304, 326). The visible Cancel closes at once (`PVFormDialog.swift` `onCancel`) |
| Kit scroll-to | `macos/App/DesignSystem/Components/ScrollRowIntoView/PVScrollRowIntoView.swift` `.pvScrollsRowIntoView(_:anchor:animated:scrollsOnAppear:…)` | The kit's one scroll-to-id; `DesignSystem/README.md` says not to add another `ScrollViewReader` |
| Kit panel | `macos/App/DesignSystem/Components/Panel/PVPanel.swift` `PVPanel.headerAndBody` | Title, subtitle and body in one stack; D-1 wraps only the body slot's content |
| Kit measuring | `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` (`.onGeometryChange(for:)`) | The kit's idiom for reading a size into state |
| Section header | `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift` `PVSectionHeader(title:meta:aside:actions:)` | Has an aside slot |
| Icons | `macos/App/DesignSystem/Components/Icon/PVIcon.swift` `PVSymbol` | No help symbol (`info` = "info.circle" exists) |
| Store | `macos/App/Platform/GenealogyStore.swift` `updateArtifact(…, graphIDs:, metadata:) -> (artifact:, metadata:, effects:)`; `CatalogSourceWorkspace.metadataRows(for:)`, `setMetadataRows(_:for:)` | Ready (S15-33, S15-09) |
| Cache patch | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `.wroteMetadataBoard(sourceId:owner:entries:)`; `CatalogLastUpdated.swift` `metadataBoard()` | Patches the owner's rows into the cached Source workspace; no invalidation |
| Fake | `macos/App/Platform/FakeStore.swift` `updateArtifactCalls`, `updateArtifactGraphIDs`, `updateArtifactMetadata`, `updateArtifactError`, `updateArtifactHold`, `scriptedMetadataRows`, `getEvidenceGraphUnlinkImpactCalls`, `getEvidenceGraphUnlinkImpactPairs`, `getEvidenceGraphUnlinkImpactError`, `unlinkImpactBlocks`, `metadataWriteResults` | Enough for every test here |
| Tests | `macos/ProvenenciaTests/PVPanelFormDialogTests.swift` `PVSheetDismissTests`, `PVFormDialogControlsTests`; `macos/ProvenenciaTests/SourcePageModelTests.swift` card tests (`cardSaveIsOneWrite`, `cardSaveErrorStaysOnTheCardThatSaved`, `cancelArtifactFieldsRestoresDrafts`, `editingArtifactIsOneCardAtATime`, `cardSwitch*`, `cardEdit*`, `cardSave*`, `cardCitationsLockHoldsUntilTheLinkIsSaved`, `cardLinkReadFailsAndReadAgain`, `cardCancelDropsLinkChanges`, `cardFieldsOnlySaveErrorSaysChangesKept`, `beginEditOnTheSameCardKeepsItsEdits`, `restingCardMarksAnArtifactInNoGraph`) and dialog tests (`addMetadataFromDialog`, `addMetadataDropdownSeesFieldCreatedElsewhere`, `addMetadataRequiresFieldAndValue`, `invalidMetadataAddLandsOnValueField`); `MetadataBoardModelTests.swift` (`addDialogWritesToItsTargetBoard`, `addRowOnLiveBoard`, `addRowSurvivesReset`, `invalidValueStaysOnTheFieldOrDialog`, on a private `TestMetadataBoardHost`); `MetadataDraftHostTests.swift`; `SourceGraphFormModelTests.swift` | Each is ported, updated or removed as § 8 lists |

**Precedent.** S15-12's graph form (`SourceGraphFormModel` + `View.sourceGraphFormDialog`, `pvFormDialog(item:)` with `onConfirm: { await $0.save() }`): copy the model shape (a `Mode`, one `Draft` value, `prefill`, `isDirty = draft != prefill`, a load state for the engine's locks, `save() -> Bool`) and the presentation modifier. S15-36's **Next adds** is the recipe for the board inside it: a `MetadataDraftHost` over `CatalogSourceWorkspace.metadataRows(for:)` and a `MetadataBoardModel` on it; Save sends `metadata: host.draft`, applies `.wroteMetadataBoard` from the reply's rows, and on a failure calls `board.showRefusal` and shows the top Callout when it answers false; dirty reads `host.isChanged`. `PVConfirm`'s `.keyboardShortcut(tone.returnConfirms ? nil : .defaultAction)` is the precedent for a conditional shortcut. `PVTextArea`'s `PVTextAreaContentHeightKey` is the precedent for measuring content height with a `PreferenceKey`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Kit scrolling body.** `PVFormDialogContent` and both `pvFormDialog` overloads gain `scrollsBody: Bool = false` and `scrollTarget: AnyHashable? = nil`. When `scrollsBody`, the form is wrapped in `ScrollView(.vertical)`; the form's height is read with `.onGeometryChange(for: CGFloat.self) { $0.size.height }` into `@State bodyHeight` (the `PVPageColumns` idiom), and the scroll view is framed at `min(bodyHeight, PVFormDialogContent.maxScrollingBodyHeight)` with `maxScrollingBodyHeight = 560` *(new)*; the title, subtitle and footer stay fixed. The scroll view carries `.pvScrollsRowIntoView(scrollTarget, anchor: .top, animated: false, scrollsOnAppear: true)` so a section with that `.id` opens at the top. Every existing caller passes neither parameter and draws as today | SA-6 (the body scrolls between a fixed title and buttons); S15-11 is the second caller; the inventory says the scrolling body is opt-in and the kit's; `DesignSystem/README.md` § ScrollRowIntoView: no other `ScrollViewReader` (review M-1, M-7) | A `ScrollViewReader` of its own (README forbids it); a new PreferenceKey (the kit measures with `onGeometryChange`); `.frame(maxHeight:)` alone (a ScrollView in a sheet does not size to its content) |
| D-2 | **Inner edit.** `View.pvSheetInnerEdit(_ open: Bool)` *(new, `PVSheet.swift`)* sets a PreferenceKey reduced with `\|\|`. `PVSheetChrome` reads it with `onPreferenceChange` into `@State innerEditOpen` and: (1) gives its zero-size Escape button `.keyboardShortcut(innerEditOpen ? nil : .cancelAction)` (the `PVConfirm` precedent), so Escape reaches the inner edit's own handler; (2) passes it to the outside-click path only — `dismiss()` calls `PVSheetCloseRule.outcome(…, innerEditOpen:)` *(new input, default false)*, which passes it to `PVSheetDismiss.action(isRunning:isDirty:isConfirmingDiscard:innerEditOpen:)` *(new input, default false)*, answering `.ignore` when true; a content-initiated close through `\.pvSheetClose` passes false and follows the dirty rule as today; (3) puts `\.pvSheetInnerEditOpen` *(new environment value)* on its content, which `PVFormDialogContent` reads for its Confirm's shortcut, `PVFormDialogControls.confirmShortcut(innerEditOpen:)` *(new: nil when open, `.defaultAction` otherwise)*. The components that own an open edit raise the signal: `PVInlineEdit` applies `.pvSheetInnerEdit(isEditing)` (its own `.cancelAction` Cancel then takes Escape); the board's add row (D-10) applies `.pvSheetInnerEdit(true)` while drawn and closes on `.onExitCommand`; a suggestion row applies `.pvSheetInnerEdit(hasTypedText)` and on `.onExitCommand` clears its text. Outside a sheet nothing reads it | SA-15: the kit owns the order, so S15-10 and S15-11 get it with no code; one dismiss rule, as `PVSheet`'s doc promises; footer rules live in `PVFormDialogControls` (review M-2, M-4, X2-1) | A new `PVSheetKeys` enum; an `.ignore` answer alone (the hidden button would still take Escape and hand it to nobody) |
| D-3 | **One value view.** `MetadataValueText` *(new, internal, `MetadataBoardView.swift`)* takes the entry, an accessibility identifier, `lineLimit: Int?` and `chrome: Bool`. It draws today's `displayValue` (mono caption; a URL with `openURL` is the plain underlined link button in `PVColor.textLink` with `L10n.Sources.openMetadataURL` as help). `chrome: true` applies `metadataBoardValueChrome()` (board rows, so entering edit does not shift glyphs); `chrome: false` draws bare text (the card). `lineLimit: nil` wraps as today; `1` truncates the tail | Inventory "Metadata value": one view for both; not `PVButton .link` | Two copies of the URL drawing |
| D-4 | **The sheet's model.** `ArtifactSheetModel` *(new, `macos/App/Features/Sources/ArtifactSheetModel.swift`)*, `@MainActor @Observable final class`, `init(context: SourcePageContext, mode: Mode, opensAt: Section, onSaved: @escaping @MainActor (CatalogArtifact, [CatalogMetadataEntry], CatalogWriteEffects) -> Void)`; `enum Mode { case edit(CatalogArtifact) }` (S15-10 adds `create`); `enum Section: Hashable { case top, graphs, copyDetails }`. State: `ArtifactSheetDraft` *(new)* `{ label, description }`, `prefill`, `labelError`, `saveError`, `isSaving`. Links, moved from the card's `LinkEdit` with unchanged behavior: `graphIDs` (set when the session's graph list is first loaded: at init if it is, else by `graphsChanged()`), `excludedGraphIDs`, `blocks`, `linkLoadState: SourceFormLoadState`, `linkLoadTask: Task<Void, Never>?` (what tests await), `awaitingGraphs` (true while the list is `.loading` and `graphIDs` is not set), `lockedGraphs`, `checkedGraphIDs`, `graphs`, `inNoGraph`, `linksChanged`, `savedGraphIDs`, `loadLinkLocks()`, `graphsChanged()` (sets `graphIDs` and starts `linkLoadTask` once). Board: `let host: MetadataDraftHost` from `context.workspace?.metadataRows(for: MetadataOwner(level: .artifact, id:)) ?? []` and `context.fields`, and `let board: MetadataBoardModel(host:)` | The graph form's shape (precedent); S15-36 **Next adds**; the `onSaved` closure keeps the apply on the section (D-7) so S15-10 reuses it (review X-5, X-7) | A back reference to `SourceArtifactsSection`; capturing `graphIDs` only at open (a sheet opened before the list loads would never get its group) |
| D-5 | **Dirty and unconfirmed input: one typed-text signal.** `MetadataBoardHost` gains `var isChanged: Bool { get }` (`SourcePageMetadataHost`: false; `MetadataDraftHost`: today's). `MetadataBoardModel` gains `typedFieldIDs: Set<String>` *(new)* and `setTyped(_ typed: Bool, fieldID:)` *(new)*: `MetadataBoardTextEditor` and `MetadataBoardSuggestionRow` call it from `onChange` of their local text only when "differs from the starting text" flips, never per keystroke (the local buffers stay, per the editor's comment). An id is cleared where its buffer is: the editor calls `setTyped(false, …)` when it leaves edit; a suggestion row when it fills or its Escape clears its text; the model removes the id in `dismissSuggestion`, `clear` and `resetDrafts` (only for fields no longer on the board). `hasTypedEdit` is deleted; `showRefusal`'s "another row holds typed text" check reads `!typedFieldIDs.subtracting([fieldID]).isEmpty`. `hasUnconfirmedInput` *(new)* = `!typedFieldIDs.isEmpty \|\| ` an open add row with a picked field or a value with non-whitespace text. `hasUnsavedChanges` *(new)* = `host.isChanged \|\| hasUnconfirmedInput`. The sheet: `isDirty = draft (trimmed) != prefill \|\| linksChanged \|\| board.hasUnsavedChanges` | SA-8; the board's buffers are view-local by design, so the views report flips (review X-2); one typed-text rule in the board (review M2-1); a dismissed or cleared row cannot leave a stale flag that would block Save (review X2-3) | Keeping `hasTypedEdit` beside it (it reads `drafts`, which the views never write, so the two disagree) |
| D-6 | **Save.** `save() async -> Bool` (true closes). In order: `board.hasUnconfirmedInput` → `saveError = L10n.string(L10n.Sources.artifactSheetFinishOpenField)` *(new: "Finish the open field first: save its value or clear it." — worded for every field kind, since a suggestion row keeps with "Save" and its ✕ dismisses)*, false, no write (user, 2026-11-13). Trimmed label empty → `labelError = L10n.string(L10n.Sources.artifactLabelRequired)`, false. The link read present and not `.ready` → false. `graphIDs = linksChanged ? savedGraphIDs : nil`; `metadata = host.draft`. Nothing changed (text equal, `graphIDs == nil`, `metadata == nil`) → true, no write. Else `isSaving`, `saveError = nil`, one `updateArtifact`; on success `onSaved(artifact, metadata rows, effects)`, true. On failure: `context.fail(error)` nil → true (the page is leaving); else `board.showRefusal(error)` true → false; else, when `graphIDs != nil`, `await loadLinkLocks()`, then `saveError` = the card's `saveErrorBody` rule (moved): `artifactCardSaveErrorGraphKeeps(graph:error:)` when an unchecked graph is now locked, else `artifactCardSaveErrorKept(_:)`; false. `canSave` = not saving and (no link read, or `.ready`); the label is checked in `save()`, as the graph form does | SA-9; S15-33 D-6 (`params[0]`); the user's call on unconfirmed input; AC-1.4 needs an enabled Save (review X-4) | Saving unconfirmed text silently, or dropping it |
| D-7 | **Applying a write: one write per fact, as the live board does.** `SourceArtifactsSection.show(_:effects:)` becomes `show(_ art: CatalogArtifact, metadata: [CatalogMetadataEntry]? = nil, effects: CatalogWriteEffects)`: (1) the Artifact replaced or appended and `context.setWorkspace(workspace)` (the page and the cached key; no slot patches the Artifact list); (2) when `metadata` is set, `context.workspace?.setMetadataRows(metadata, for: .artifact owner)` on the page only, as `SourcePageMetadataHost.entries`' setter does; (3) `context.session.apply(effects, .wroteMetadataBoard(sourceId:, owner:, entries: metadata))` when `metadata` is set, else `context.session.apply(effects)`, so the `.metadataBoard()` slot alone writes the rows into the cached key. `patchArtifacts` is folded into it. The sheet's `onSaved` calls `show` | Map review K-1 (never `.mutatedSourceWorkspace`); one apply path for every Artifact write, so S15-10's create (an append) calls the same function (review M-3); the rows reach the cache through the slot that owns that job, once (review M2-2), and `add-workspace-place` § 6's line stays true | `setMetadataRows` before `setWorkspace` (the rows written into the cache twice); a separate `applySaved` |
| D-8 | **Presentation.** `View.artifactSheet(_ model: Binding<ArtifactSheetModel?>)` *(new)*, in `macos/App/Features/Sources/ArtifactSheet.swift` (new), wraps `pvFormDialog(item:)`. `PVFormDialogCopy` gains `verbatimTitle: String?` and an init `init(verbatimTitle: String, verbatimSubtitle: String?, confirm:, cancel:)` that also stores `title = LocalizedStringResource(stringLiteral: verbatimTitle)` (never drawn when `verbatimTitle` is set); `PVFormDialogContent` draws `verbatimTitle` when set, as it does the subtitle. Title `L10n.Sources.editArtifactLabel(ref:)` ("Edit ART-7QK2M"; the same copy and argument as the pencil, review M-11); subtitle "<Source ref> · <Source title>"; confirm `L10n.Sources.saveArtifact`; cancel `L10n.Sources.cancelAction`; width 560; `isRunning: isSaving`; `isDirty`; `confirmDisabled: !canSave`; `scrollsBody: true`; `scrollTarget: opensAt == .top ? nil : opensAt`; prefix `sources.page.artifactSheet`. **Cancel closes at once**, the kit's rule for every dialog; Escape and an outside click ask (user, 2026-11-13, over SA-8, which had Cancel ask too) | The graph form's modifier; review X-3 answered by the user | A per-sheet Cancel that asks |
| D-9 | **The sheet's body**, top to bottom, each section with `.id(Section)`: the danger `PVCallout` (title `artifactCardSaveErrorTitle`, message `saveError`, not compact, `announces: true`) when set; `PVTabPair` Label (`PVField` required, error `labelError`, `PVInput`, `activateOnAppear` when `opensAt == .top`; editing clears `labelError`) and Description (`PVField` hint `artifactDescriptionHint`, `PVTextArea` `lineLimit: 3...8`); Evidence graphs (id `.graphs`): when `awaitingGraphs`, today's loading `SourceFormLoadArea`; when `graphIDs` is set, `SourceFormLoadArea(state: linkLoadState, …, onReadAgain: loadLinkLocks)` around `SourceArtifactGraphLinks(graphs:, excluded:, locked:, isDisabled: isSaving)` and the none line when `inNoGraph`; nothing when the session has no list; the view's `onChange(of: visibleGraphs ids)` calls `graphsChanged()`; Copy details (id `.copyDetails`): `PVField(label: L10n.Sources.artifactCopyDetailsLabel)` around `MetadataBoardView(board:, placement: .form, accessibilityIdentifierPrefix: "sources.page.artifactSheet.copyDetails", onClear: { id in Task { _ = await board.requestClear(fieldID: id) } })` (a draft host clears with no confirm). Label, Description and the board are `.disabled(isSaving)` | SA-6, SA-9; the card's graph group moved with its states (review X-7) | Drawing the graph group anew |
| D-10 | **Board placement and the add row.** `MetadataBoardView` replaces `onAdd` with `placement: MetadataBoardPlacement` *(new enum: `page`, `form`)*. `.page`: the empty state (no saved, no suggested rows) keeps title and message with no button; no foot button; the add row at the foot while `board.addRow != nil`. `.form`: no empty state; at the foot the ghost "Add field" button (`L10n.Sources.addMetadata`, `PVButton(variant: .ghost, size: .sm, icon: .plus)`) while the row is closed, the row while open. `MetadataBoardAddRowView` *(new, private)*: a `PVCard` (default border, `PVRadius.sm`) row of `PVComboBox(size: .sm)` (options `board.fieldComboOptions`; width 150 on `.page`, 168 on `.form`; placeholder and label `L10n.Sources.metadataField`; empty label `L10n.Sources.typeNoMatch`), `PVInput(size: .sm, mono: true, prompt: L10n.Sources.metadataValue)` (`onSubmit` confirms), then ✓: `PVIconButton(.check, label: L10n.Sources.metadataAddRowConfirmLabel, tone: .accent)` disabled unless `canConfirmAddRow`, replaced by a `ProgressView().controlSize(.small)` framed at `PVControlSize.sm.height` while `isSavingAddRow` (as `PVInlineEdit`'s icon stack does), and ✕: `PVIconButton(.dismiss, label: L10n.Sources.metadataAddRowCancelLabel)` calling `closeAddRow()`. Escape inside the row closes it through `.onExitCommand { board.closeAddRow() }` on the row, not a `.cancelAction` shortcut (the sheet's Escape button stands down while the row raises `pvSheetInnerEdit`, D-2). Under the row: the row's `error` (`PVErrorText`, caption) or the taken notice (D-13). The row binds through `MetadataBoardModel.pickAddRowField(_:)` and `setAddRowValue(_:)` *(new)*, which clear `error` and `takenFieldLabel`. One inner edit at a time: `openAddRow()` cancels an open row edit; `beginEdit(fieldID:)` closes the add row (its content goes, as ✕ would) | SA-13; the user's call that the Source's section has no foot button; board decision (bordered row, 168 / 150pt); `add-ui-component` "Do not add `.keyboardShortcut(.cancelAction)` to a sheet yourself" (review X-13); two open inner edits would fight over Escape | Two booleans for the placement; a `.cancelAction` ✕ |
| D-11 | **Opening and closing the sheet.** `SourceArtifactsSection.sheet: ArtifactSheetModel?` *(new)*; `openSheet(artifactID:at:)` *(new)* does nothing when `sheet != nil` or the id is not in `items`; it builds the model with `onSaved: { [weak self] in self?.show($0, metadata: $1, effects: $2) }` and starts `linkLoadTask` when the graph list is loaded. The pencil opens `.top`, "Link to a graph" `.graphs`, "+n more" `.copyDetails`. `workspaceChanged()` *(new, replaces `seedDrafts`, called where `SourcePageModel.apply` called it)*: when `sheet`'s Artifact is not in `items`, `sheet = nil` and `context.toast = VocabularyToast(title: L10n.string(L10n.Sources.toastArtifactGoneTitle), body: L10n.Sources.toastArtifactGoneBody(label:), tone: .info)` *(new keys)*, with the label the sheet opened with | SA-12, SA-14; one modal sheet replaces "one card at a time" | Keeping `pendingEditSwitch` |
| D-12 | **What goes from the Artifacts section.** `editingArtifactID`, `pendingEditSwitch`, `PendingEditSwitch`, `editGate`, `labels`, `descriptions`, `fieldErrors`, `savingID`, `cardSaveError`, `LinkEdit`, `linkEdit`, `linkLoadTask`, `seedDrafts`, `beginEdit`, `confirmEditSwitch`, `cancelEditSwitch`, `openEdit`, `startLinkEdit`, `awaitingGraphs`, `graphsChanged`, `dropLinkEdit`, `linkedGraphIDs`, `liveGraphIDs`, `lockedGraphs`, `checkedGraphIDs`, `linkEditGraphs`, `cardInNoGraph`, `linksDirty`, `loadLinkLocks`, `fieldsDirty`, `textChanged`, `canSaveFields`, `cardLabelError`, `cancelFields`, `savedGraphIDs`, `saveFields`, `saveErrorBody`, `replace`, `patchArtifacts`. `stagePageDrop` drops its `savingID` guard and `cancelFields` line and returns when `sheet != nil`. `confirmAttach` calls `show(ingested.artifact, effects: ingested.effects)` instead of `replace` (as `create` does) and keeps `refreshCoverFromStore()` | The sheet owns edit; `replace`'s `.mutatedSourceWorkspace` reload is what K-1 forbids | Keeping `replace` for `confirmAttach` |
| D-13 | **The taken notice.** `MetadataBoardAddRow` gains `takenFieldLabel: String?` *(new)*; `resetDrafts` sets it to the cleared field's label when it clears the row's `fieldID`. The add row shows `L10n.Sources.metadataAddRowFieldTaken(field:)` *(new: "%@ was added in another window. Pick another field.")* as a muted caption and posts it with `.pvAnnounce` | S15-36 review UX-2 | A toast |
| D-14 | **Dialogs.** `MetadataBoardDialogs` keeps only `MetadataBoardDeleteItem`, `pendingDelete`, `isClearing`, `askClear`, `confirmClear`; every add member and `MetadataBoardAddForm` go; `MetadataBoardAddResult` moves to `MetadataBoardModel.swift`. The doc comments that name the dialog are rewritten: `MetadataBoardModel.host`, the `MetadataBoardView` header, the `MetadataBoardDialogs` class doc (review M-10). `SourcePageView` drops the Add metadata `pvFormDialog`, `addMetadataPresented` and the `pendingEditSwitch` `pvConfirm`, and adds `.artifactSheet($model.artifacts.sheet)`. `acceptsArtifactPageDrop` drops `metadataDialogs.isAdding` (STEP-3) and swaps `artifacts.pendingEditSwitch == nil` for `artifacts.sheet == nil` (STEP-7) | SA-13 "The Add metadata dialog goes" | Keeping the dialog for the Source's board |
| D-15 | **Metadata header.** `SourcePageMetadataView`: `PVSectionHeader(title:, aside:, actions:)` with no `meta`; the aside is `PVIcon(.circleHelp, size: 15)` *(new `PVSymbol.circleHelp = "questionmark.circle"`)* in `PVColor.textMuted` with `.help(L10n.string(L10n.Sources.metadataIntro))`, `.accessibilityElement()`, `.accessibilityLabel(L10n.Sources.metadataHelpLabel)` *(new: "About metadata")*, `.accessibilityHint(L10n.Sources.metadataIntro)`; the action is `PVButton(L10n.Sources.addMetadata, variant: .ghost, size: .sm, icon: .plus)` calling `model.metadata.openAddRow()`, id `sources.page.addMetadata`; the intro `Text` and `metadataFieldCount` go; the board is `.page` | User 2026-11-13 (board `metaHelp`); README: a macOS tooltip is `.help()` | Porting the web `PVTooltip` |
| D-16 | **Read-only copy details** (SA-4). `ArtifactCopyDetailsView` *(new)*, in `macos/App/Features/Sources/ArtifactCopyDetailsView.swift` (new), takes the Artifact's rows, its ref and `onMore`. `static func shown(_ rows: [CatalogMetadataEntry]) -> (shown: [CatalogMetadataEntry], more: Int)`: the rows with `hasValue`, in saved order, the first four and the count past them. It draws a two-column `Grid`: field name (`PVTypeScale.caption`, `PVColor.textMuted`, one line) and `MetadataValueText(lineLimit: 1, chrome: false)`; then, when `more > 0`, `PVButton(L10n.Sources.artifactCopyDetailsMore(count:), variant: .link, size: .sm)` *(new key: "+%lld more")* with accessibility label `L10n.Sources.artifactCopyDetailsMoreLabel(count:ref:)` *(new: "%lld more copy details of %@")*. The grid is one accessibility group whose children read "Field, value". Nothing when `shown` is empty. Ids `sources.page.artifact.<id>.copyDetails`, `sources.page.artifact.<id>.copyDetails.more`. The card places it under the description and passes `onMore` = `openSheet(…, at: .copyDetails)` | SA-4, SA-11; inventory "Copy details, read-only" is its own snowflake, and S15-35 moves this view into the new list (review M-5) | A block inside `artifactCard` |
| D-17 | **Strings.** Kept: `addMetadata` (comment: "Button that opens the inline Add field row"), `metadataField`, `metadataValue`, `typeNoMatch`, `artifactLabelRequired`, `artifactCardSaveErrorTitle`, `artifactCardSaveErrorKept`, `artifactCardSaveErrorGraphKeeps` (comments: "Artifact sheet" for "card"), `saveArtifact`, `artifactDescriptionHint`, `metadataIntro` (comment: "Tooltip on the Metadata section's help icon"). Changed: `editArtifactLabel(ref:)` → "Edit %@" (the pencil's label and the sheet's title). New: `artifactCopyDetailsLabel` ("Copy details"), `artifactCopyDetailsMore(count:)`, `artifactCopyDetailsMoreLabel(count:ref:)`, `artifactSheetFinishOpenField` ("Finish the open field first: save its value or clear it."), `toastArtifactGoneTitle` ("Artifact no longer here"), `toastArtifactGoneBody(label:)` ("“%@” was deleted or moved in another window."), `metadataAddRowFieldTaken(field:)`, `metadataHelpLabel`, `metadataAddRowConfirmLabel` ("Add this field"), `metadataAddRowCancelLabel` ("Stop adding a field"). Removed: `addMetadataDialogTitle`, `addMetadataDialogSubtitle`, `addMetadataConfirm`, `metadataFieldHint`, `metadataValueHint`, `metadataFieldRequired`, `metadataValueRequired`, `metadataFieldCount`, `artifactCardLabelRequired`, from `L10n.swift` and `Localizable.xcstrings` | `add-localized-string` | — |
| D-18 | **The draft host is observable.** `MetadataDraftHost` becomes `@Observable` (`entries` and `dismissed` tracked; `fields` and `opened` are `let` or `@ObservationIgnored`), so the board redraws after a clear, dismiss, reorder or fill on a draft, and `isDirty` is re-read | Today it is a plain class: a draft board's view and the sheet's `isDirty` would not see changes (review X-1) | Copying rows into the observable board model |

**Rules this plan bends.** None. The sheet keeps the kit's Cancel rule (D-8, the user's call over the brief's SA-8 wording); the add row uses no `.cancelAction` shortcut (D-10); no second `ScrollViewReader` (D-1).

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `scrollsBody`, `scrollTarget`, `maxScrollingBodyHeight` | kit params | `PVFormDialogContent`; `.pvScrollsRowIntoView` (reused for the scroll); `PVPageColumns`' `onGeometryChange` (reused idiom); the capped `ScrollView`s in `PVChecklistFrame`, `PVDeleteImpactGroupStack`, `PVProvenanceSheet` | Those cap a list inside a form; a sheet's whole body must size to its content up to a cap, so it reads its height |
| `PVFormDialogCopy.verbatimTitle` | field | `verbatimSubtitle` | Same shape for the title |
| `pvSheetInnerEdit(_:)`, `\.pvSheetInnerEditOpen`, `innerEditOpen:` on `PVSheetDismiss.action`, `PVFormDialogControls.confirmShortcut(innerEditOpen:)` | kit modifier / env / rule inputs | `PVSheetDismiss`, `PVFormDialogControls` | Extends the one dismiss rule and the footer rules; searched `DesignSystem/` for an inner-edit signal (`onExitCommand`, `PreferenceKey`): none |
| `PVSymbol.circleHelp` | enum case | `PVSymbol.info` | A help glyph is not an info glyph (the board's `circle-help`) |
| `MetadataValueText` | view | `MetadataBoardTextEditor.displayValue` | Pulled out of it (D-3) |
| `MetadataBoardPlacement`, `MetadataBoardAddRowView` | enum / view | `MetadataBoardAddForm` (removed), the board's empty state | The row replaces the dialog's form |
| `MetadataBoardHost.isChanged`, `MetadataBoardModel.typedFieldIDs`, `setTyped`, `hasUnconfirmedInput`, `hasUnsavedChanges`, `pickAddRowField`, `setAddRowValue`, `MetadataBoardAddRow.takenFieldLabel` | protocol member / fields / funcs | `MetadataDraftHost.isChanged`, `hasTypedEdit`, `resetDrafts` | Extends the board's existing state; one read for every draft form |
| `ArtifactSheetModel`, `ArtifactSheetDraft` | type | `SourceGraphFormModel`, `SourceArtifactsSection.LinkEdit` | Copies the graph form's shape; absorbs `LinkEdit` |
| `View.artifactSheet(_:)` | modifier | `View.sourceGraphFormDialog(_:)` | Same shape |
| `ArtifactCopyDetailsView` | view | — (the brief's own snowflake) | S15-35 moves it into the list |
| `SourceArtifactsSection.sheet`, `openSheet(artifactID:at:)`, `workspaceChanged()`; `show(_:metadata:effects:)` | members | `SourceGraphsSection.form`; `show`, `seedDrafts` | Same host pattern as the graph form; `show` extended, not forked |
| L10n keys (D-17) | key | `L10n.Sources` | — |

**Builds on.** `pvFormDialog(item:)` and `PVSheet`'s one close rule; S15-36's `MetadataDraftHost`, add row and `showRefusal`; S15-33's `updateArtifact(metadata:)` and its returned rows; S15-09's `patchArtifacts`, `setWorkspace`, `.wroteMetadataBoard` and `metadataRows(for:)`; the graph form's model shape; `SourceFormLoadArea` and `SourceArtifactGraphLinks`.

**Deletes.** Card edit (D-12's list) and its view (`editFields`, `graphLinks`, `graphLinksGroup`, `labelBinding`, `descriptionBinding`, `commitEdit`, the `editing` branches of `artifactCard` and `metaRow`); `SourceArtifactsSection.replace` and `patchArtifacts` (folded into `show`); the Add metadata dialog (D-14), `MetadataBoardAddForm` and the dialog's add members; the doc comments that describe the dialog (D-14); the `pendingEditSwitch` confirm; the `metadataIntro` paragraph and `metadataFieldCount`; the strings in D-17's Removed list; the ported card and dialog tests. `PVDiscardGate` stays (the sheet chrome uses it).

**Next adds.** S15-35: the list redraws each entry, places `ArtifactCopyDetailsView` and calls `openSheet(…, at:)`. S15-10: `ArtifactSheetModel.Mode.create`, the File row first, `MetadataDraftHost.unsavedRows(from:)`, `createArtifact(…, metadata: host.draft)` applied through `show(_:metadata:effects:)` (an append); it removes the Add artifact dialog. S15-11: the graph form passes `scrollsBody: true` and draws `MetadataBoardView(placement: .form)`; Escape and Return need no code (the kit's components raise the signal). S15-25: Move on the card menu.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Pencil, "Link to a graph", "+n more" open the sheet | `ArtifactSheetModel.init` | A new model per open (no reset needed); `openSheet` returns when the id is not in `items`; the graph list loading → `awaitingGraphs`, then `graphsChanged()`; no list → no group (today's rule) |
| A page reload while the sheet is open | `SourceArtifactsSection.workspaceChanged()` | Artifact gone → closes with the toast (AC-3.5); Artifact still there → the sheet keeps its draft (the draft host does not read the workspace after open) |
| The Source deleted elsewhere during Save | `ArtifactSheetModel.save` | `context.fail` answers nil → true, the page leaves (today's rule) |
| A graph made or deleted while the sheet is open | `graphsChanged()`, `savedGraphIDs`, `graphs` | Moved unchanged from the card (T-5 ports `cardSaveKeepsAGraphMadeDuringEdit`, `cardSaveSkipsAGraphDeletedDuringEdit`) |
| The add row on the Source's board while a reload adds its field | `resetDrafts` | The taken notice (AC-5.5) |

### Scale

At the map's Scale (a Source with 60 Artifacts, 63 seeded fields): one Save is one `updateArtifact` write, and the page patches its workspace from the reply (one `setWorkspace`, then the same rows through `.wroteMetadataBoard`, idempotent); no page reload and no list reload unless the links changed (the effects then reload the graph list, as today). Opening the sheet builds the draft host from one owner's rows (at most 63 plus extras) and runs one unlink preflight for the Artifact's linked graphs (as the card did). The cards' copy details read `metadataRows(for:)` per card from the workspace already in memory; it checks `artifacts.contains` first, so about 3,600 id comparisons per render at 60 cards, and no store call.

### Edges

- **Assumptions.** `MetadataDraftHost.isChanged` compares saved values and empty suggestions, not open buffers; D-5 adds the buffers through `typedFieldIDs`, the board's one typed-text signal, which `showRefusal` also reads. Because D-6 stops Save on unconfirmed input, `showRefusal` meets typed text in another row only from a direct call (T-2's `refusalWithTypedTextElsewhereAnswersFalse`). `savedGraphIDs` sends graphs made since open that link the Artifact (moved unchanged).
- **Neighbors.** A Save with only text changed sends `graphIDs: nil` and `metadata: nil` (T-4). A Save with only a dismissal sends a draft with the dismissed layout row (T-4). A Save with nothing changed writes nothing (AC-1.3). An add row opened and closed empty is not dirty (T-3).
- **Removed inputs.** An Artifact deleted while its sheet is open: the draft is lost, by SA-12 (T-7). A field the add row picked that another window added: cleared with a notice, the value kept (T-2).
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** The copy-details block shows at most four values per card whatever the row count; the sheet's board is the existing board over one owner's rows.

### Schema

None — no catalog change.

### Go core

None — no engine change.

### FFI

None — S15-33 shipped the request field and reply rows.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Artifact sheet | `PVFormDialog` (`pvFormDialog(item:)`, `scrollsBody`, 560pt) | `macos/App/Features/Sources/ArtifactSheet.swift` *(new)* | D-8, D-9 |
| Label / Description | `PVField` + `PVInput` / `PVTextArea` in `PVTabPair` | same | Today's card fields, moved |
| Evidence graphs | `SourceFormLoadArea` + `SourceArtifactGraphLinks` | same | Today's group, moved |
| Copy details board | `MetadataBoardView(placement: .form)` | `MetadataBoardView.swift` | The Source's board, in the sheet |
| Inline Add field row | `PVCard` + `PVComboBox` + `PVInput` + `PVIconButton` ✓ ✕ | `MetadataBoardView.swift` | D-10 |
| Metadata value | `MetadataValueText` | `MetadataBoardView.swift` | D-3 |
| Copy details, read-only | `Grid` of `Text` + `MetadataValueText` + `PVButton .link` | `macos/App/Features/Sources/ArtifactCopyDetailsView.swift` *(new)* | D-16 |
| Errors | `PVCallout` (danger, top), row error via `PVInlineEdit`, `PVErrorText` | — | SA-9 |
| Discard | the kit's `.discardChanges` (`PVSheet`) | — | AC-2.2 |
| Toast | `VocabularyToast`, info | `SourcePageContext.toast` | AC-3.5 |
| Help tooltip | `PVIcon(.circleHelp)` + `.help()` | `SourcePageMetadataView.swift` | D-15 |

**Frames in words.** *Sheet, edit* (`sheet-edit`): title "Edit ART-7QK2M", subtitle "SRC-4KD2P · 1880 U.S. census, Ottawa County, Kansas, population schedule"; Label input; Description text area; Evidence graphs: Robins checked and locked with its hint, Hale checked, Ostrander unchecked; Copy details: three saved rows (drag handle, field name 132pt semibold, value mono), "SUGGESTED BY THIS TYPE" with a dashed Image number row, then the ghost "+ Add field"; footer Cancel / Save artifact. *Editing inside* (`sheet-editing`): the URL row's editor open, Image number typed in its suggestion, the add row open on "Film number" with value "1254392". *Saving*: Save shows its spinner; everything above is disabled. *Value refused*: under the URL row "This metadata change can't be saved." in the danger tone. *Save failed*: a danger Callout "Couldn't save the artifact" at the top of the body. *Eight copy details* (`sheet-eight`): the body scrolls between the fixed title and footer, opened at Copy details. *Nothing suggested* (`sheet-no-sugg`): Copy details shows only "+ Add field". *Artifact gone* (`art-gone`): the page with the info toast. *Add field* (`meta-add-row`): at the foot of the Source's Metadata board, a bordered row: field picker, mono value, ✓ ✕. *Add field, Wide* (`meta-add-row-wide`): the same row in the 480pt side column. *Header*: "Metadata", a 15pt help glyph in muted ink, then on the right a ghost "+ Add field".

States: empty (D-10), loading and failed graph reads (today's), saving, refused row, top failure, gone. Large: eight copy details scroll inside the body; twelve graphs scroll inside their field (today's `SourceArtifactGraphLinks`). Accessibility: the sheet's fields keep their labels; the add row's controls are labelled ("Field", "Value", "Add this field", "Stop adding a field"); the taken notice is announced; the help glyph reads "About metadata" with the intro as its hint; the card's pencil reads "Edit ART-7QK2M"; the copy details read as one group of "Field, value".

**Wide.** The sheet is the same at both widths (a sheet does not follow the window). The Metadata section sits in the 480pt side column at Wide (S15-03's layout, unchanged); the add row's picker is 150pt there and the value input fills the rest. The cards keep today's grid at both widths (S15-35 changes them).

### Strings

D-17. The removed dialog strings named the old scope ("Add metadata" dialog); no other copy names it.

## 7. Steps

**STEP-1 — Kit: a scrolling form body, and inner edits that keep Escape and Return.**
Files: `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift`, `macos/App/DesignSystem/Components/Panel/PVSheet.swift`, `macos/App/DesignSystem/Components/Panel/PVSheetDismiss.swift`, `macos/App/DesignSystem/Components/InlineEdit/PVInlineEdit.swift`, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/PVPanelFormDialogTests.swift`, `macos/ProvenenciaTests/PVDiscardGateTests.swift`. Skill: `add-ui-component`.
Do: D-1, D-2 (with `PVInlineEdit` raising `pvSheetInnerEdit(isEditing)`), and D-8's `PVFormDialogCopy.verbatimTitle`. A `#Preview` of a scrolling form dialog with a scroll target. README: the FormDialog, Panel and InlineEdit rows name the scrolling body, `pvSheetInnerEdit` and the dismiss rule's new input.
Check: T-1 passes; `xcodebuild test` passes and every existing dialog draws as before (no caller passes the new parameters).

**STEP-2 — One metadata value view; the draft host is observable.**
Files: `macos/App/Features/Sources/MetadataBoardView.swift`, `macos/App/Features/Sources/MetadataDraftHost.swift`, `macos/ProvenenciaTests/MetadataDraftHostTests.swift`.
Do: D-3, D-18.
Check: T-9 passes; `MetadataDraftHostTests` and `MetadataDraftHostParityTests` pass; manual: a URL value on the Source's board still opens.

**STEP-3 — Every board adds a field in an inline row; the Add metadata dialog goes.**
Files: `MetadataBoardView.swift`, `MetadataBoardModel.swift`, `MetadataBoardHost.swift` (`isChanged`), `MetadataDraftHost.swift`, `MetadataBoardDialogs.swift`, `SourcePageMetadataView.swift` (the header button calls `model.metadata.openAddRow()`; the board passes `placement: .page`), `SourcePageView.swift` (the dialog goes), `SourcePageModel.swift` (`acceptsArtifactPageDrop` drops `metadataDialogs.isAdding`), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/MetadataBoardModelTests.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-5 (the board members and the views' `setTyped` calls), D-10, D-13, D-14 (the dialog half), the add-row strings of D-17; remove the dialog strings.
Check: T-2 passes; `rg -n "MetadataBoardAddForm|addMetadataPresented|createFromAdd|addMetadataDialog" macos` finds nothing.

**STEP-4 — The Metadata header: help tooltip, no count.**
Files: `SourcePageMetadataView.swift`, `macos/App/DesignSystem/Components/Icon/PVIcon.swift`, `L10n.swift`, `Localizable.xcstrings`.
Do: D-15.
Check: build passes; manual AC-5.1, AC-5.4, AC-5.6.

**STEP-5 — The Artifact sheet's model, and how it opens.**
Files: `macos/App/Features/Sources/ArtifactSheetModel.swift` *(new)*, `SourceArtifactsSection.swift` (`sheet`, `openSheet`, `show(_:metadata:effects:)`), `SourcePageModelTests.swift` (a `// MARK: Artifact sheet (S15-34)` section, using the file's `makeCensus`, `makeStore` and `graph` helpers through `model.artifacts.openSheet`). Skill: `add-swift-test`.
Do: D-4, D-5 (the sheet's `isDirty`), D-6, D-7, D-11's `sheet` and `openSheet`.
Check: T-3 – T-6 pass.

**STEP-6 — The Artifact sheet's view.**
Files: `macos/App/Features/Sources/ArtifactSheet.swift` *(new)*, `L10n.swift`, `Localizable.xcstrings`.
Do: D-8, D-9 and the sheet's strings of D-17.
Check: build passes; a `#Preview` of the sheet's form over a model built from fixtures shows the `sheet-edit` frame's content.

**STEP-7 — The card opens the sheet; card edit goes.**
Files: `SourceArtifactsSection.swift`, `SourcePageArtifactsView.swift`, `SourcePageView.swift` (`.artifactSheet`, the `pendingEditSwitch` confirm goes), `SourcePageModel.swift` (`workspaceChanged()`; `acceptsArtifactPageDrop` uses `artifacts.sheet == nil`), `SourcePageModelTests.swift`.
Do: D-11's `workspaceChanged`, D-12; the pencil and "Link to a graph" call `openSheet`; the card tests per § 8.
Check: T-5, T-7 pass; `rg -n "editingArtifactID|saveFields|pendingEditSwitch|linkEdit|cardSaveError|cardLabelError|artifacts\.seedDrafts|func seedDrafts" macos/App/Features/Sources macos/ProvenenciaTests/SourcePageModelTests.swift` finds nothing.

**STEP-8 — Read-only copy details on the card.**
Files: `macos/App/Features/Sources/ArtifactCopyDetailsView.swift` *(new)*, `SourcePageArtifactsView.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcePageModelTests.swift`.
Do: D-16.
Check: T-8 passes; manual AC-4.1 – AC-4.4.

**STEP-9 — Strings sweep.**
Files: `L10n.swift`, `Localizable.xcstrings`.
Do: D-17's changed and removed keys and the updated comments.
Check: `rg -n "artifactCardLabelRequired|metadataFieldCount|addMetadataConfirm|metadataFieldHint|metadataValueHint|metadataFieldRequired|metadataValueRequired" macos .agents .cursor docs/*.md` finds nothing. (Spike 15's plan docs that name them as history are expected: `deployment-plan.md`, `prs/S15-09-*`, `prs/S15-26-*`, `prs/S15-36-*`, this plan.)

**STEP-10 — Version bump.**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (every `MARKETING_VERSION`). Skill: `.cursor/rules/versioning.mdc`.
Do: 0.12.0 (or the next minor after `main`).
Check: the three agree.

**STEP-11 — Docs.**
Files: `docs/deployment-plan/spike-15/completed.md`; `docs/deployment-plan/spike-15/deployment-plan.md` (map row); this plan (status, Landed note); `docs/deployment-plan/spike-15/design/README.md` (S15-D14 stays in `design/` until S15-25 lands, its handoff; note S15-34 landed); `docs/deployment-plan/spike-15/prs/S15-26-metadata-board-takes-an-owner.md` and `S15-09-artifact-metadata-boards.md` (**Next adds**: "Superseded by S15-34"); `.agents/skills/add-workspace-place/SKILL.md` § 6 — "Metadata values" names the Artifact sheet as the first owner Save that applies `.wroteMetadataBoard`, and "Other page edits (notes, artifacts, credibility): `context.applyWrite`" becomes notes and credibility only, with Artifact writes patched through `SourceArtifactsSection.show` (review M-6, X-12); `docs/ideas/structural-debt.md` (a row: the sheet's and the graph form's unlink-lock reads are two copies; one `UnlinkLocks` value both hold; touch with S15-11, review M-8).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` passes; `rg -n "artifacts, credibility" .agents/skills/add-workspace-place/SKILL.md` finds nothing.

Commits: one per step. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Edit an Artifact's label, graphs and copy details in one sheet.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PVPanelFormDialogTests.swift` `PVSheetDismissTests` `innerEditStandsDown` *(new)*, `PVFormDialogControlsTests` `confirmShortcutYieldsToAnInnerEdit` *(new)*; `PVDiscardGateTests.swift` `sheetCloseOutcomes` *(extended)* | `action(isRunning: false, isDirty: true, isConfirmingDiscard: false, innerEditOpen: true)` → `.ignore`, with `false` → `.confirmDiscard`; `confirmShortcut(innerEditOpen: true)` → nil, `false` → `.defaultAction`; `PVSheetCloseRule.outcome(…, innerEditOpen: true)` → `.nothing`, and with the default `false` every existing case unchanged | D-2, AC-2.4, AC-2.5 |
| T-2 | Swift | `MetadataBoardModelTests.swift` (its `TestMetadataBoardHost`): `addRowOnLiveBoard` *(kept)*, `addRowSurvivesReset` *(updated: expects `takenFieldLabel` set)*, `addRowTakenFieldSaysSo` *(new)*, `pickingAFieldClearsTheNotice` *(new)*, `unconfirmedInputReadsOpenRows` *(new)*, `oneInnerEditAtATime` *(new)*, `dismissingATypedSuggestionClearsItsFlag` *(new)*, `refusalWithTypedTextElsewhereAnswersFalse` *(new, replaces the `hasTypedEdit` path)*, `invalidValueStaysOnTheField` *(was `invalidValueStaysOnTheFieldOrDialog`, dialog half removed)*. `SourcePageModelTests.swift` (its `makeStore` / `makeModel`): `addRowSavesTheSourceValue` *(was `addMetadataFromDialog`)*, `addRowNeedsFieldAndValue` *(was `addMetadataRequiresFieldAndValue`)*, `addRowRefusalStaysOnTheRow` *(was `invalidMetadataAddLandsOnValueField`)*, `addRowPickerSeesFieldCreatedElsewhere` *(was `addMetadataDropdownSeesFieldCreatedElsewhere`)* | Live board: open, pick Author, "Eliza", confirm → saved, `addRow == nil`; no field or a blank value → `canConfirmAddRow` false, no write; `sourcemetadata.invalid` → `addRow.error`, row open; Collection picked, then `resetDrafts` with Collection on the board → `fieldID == ""`, `takenFieldLabel == "Collection"`, value kept; `pickAddRowField` clears it; an open empty row → `hasUnconfirmedInput` false, a value → true, `setTyped(true, fieldID:)` → true; `beginEdit` closes the add row, `openAddRow` ends the row edit; `setTyped(true, fieldID: image)` then `dismissSuggestion(image)` → `hasUnconfirmedInput` false; `setTyped(true, fieldID: provider)` then `showRefusal` naming URL → false | AC-5.2, AC-5.3, AC-5.5, D-5, D-10, D-13 |
| T-3 | Swift | `SourcePageModelTests.swift` § Artifact sheet: `sheetOpensFromTheWorkspace`, `sheetDirtyIgnoresPutBackText`, `sheetLabelRequired`, `sheetStopsOnUnconfirmedInput` *(all new)* | Opens with the Artifact's label, description and its copy-detail rows; a label typed then put back → `isDirty` false; empty label → `save()` false, `labelError == "Give the artifact a label."`, no write; `board.setTyped(true, fieldID:)` then Save → false, `saveError` is the finish-the-field copy, no write | AC-1.1, AC-1.4, AC-2.3, AC-2.6, D-5, D-6 |
| T-4 | Swift | same: `sheetSaveIsOneWriteWithTheDraft` *(was `cardSaveIsOneWrite`, extended)*, `sheetSaveWithNothingChangedWritesNothing`, `sheetSaveSendsNoLinksWhenOnlyTextChanged` *(was `cardSaveSendsNoLinksWhenOnlyTextChanged`)*, `sheetSavePatchesWithoutReload` | AC-1.2's edits → `updateArtifactCalls == 1`, the draft's values in that order, `graphIDs == nil`; nothing changed → 0 calls, `save()` true; the reply rows land in `workspace.metadataRows(for: artifact)` and in the cached key; `getSourceWorkspace` not called | AC-1.2, AC-1.3, D-6, D-7 |
| T-5 | Swift | same, ported (same inputs and expectations, on `model.artifacts.sheet`): `sheetOnOneGraphShowsTheGroup` (`cardEditOnOneGraphShowsTheGroup`), `sheetLocksEveryBlock` (`cardEditLocksEveryBlock`), `sheetBeforeTheListLoadsGetsTheGroup` (`cardEditBeforeTheListLoadsGetsTheGroup`, asserting `awaitingGraphs`), `sheetCitationsLockHoldsUntilTheLinkIsSaved` (`cardCitationsLockHoldsUntilTheLinkIsSaved`), `sheetSaveKeepsAGraphMadeDuringEdit` (`cardSaveKeepsAGraphMadeDuringEdit`), `sheetLinkReadFailsAndReadAgain` (`cardLinkReadFailsAndReadAgain`), `sheetSaveSkipsAGraphDeletedDuringEdit` (`cardSaveSkipsAGraphDeletedDuringEdit`), `sheetSaveRefusedNamesTheGraphAndLocksIt` (`cardSaveRefusedNamesTheGraphAndLocksIt`), `sheetFieldsOnlySaveErrorSaysChangesKept` (`cardFieldsOnlySaveErrorSaysChangesKept`), `sheetSaveReloadFailureOffersReadAgain` (`cardSaveReloadFailureOffersReadAgain`), `sheetMarksAnArtifactInNoGraph` (the card half of `restingCardMarksAnArtifactInNoGraph`: `openSheet(artifactID: "P6", at: .graphs)`, await `linkLoadTask`, uncheck every graph → `checkedGraphIDs` empty and `inNoGraph` true, and `save()` sends `graphIDs == []`; its section half, `artifacts.inNoGraph`, stays in `restingCardMarksAnArtifactInNoGraph`) | As today, with `saveError` for `cardSaveError` and `sheet.save()` for `saveFields(id:)` | AC-1.5, AC-1.6, AC-3.3, AC-3.4, D-4 |
| T-6 | Swift | same: `sheetRefusedValueLandsOnItsRow` *(new)* | `sourcemetadata.invalid`, `params[0]` = the URL field's id → the URL row in edit with "This metadata change can't be saved.", `saveError == nil` (the typed-elsewhere case is T-2's, since D-6 stops Save first) | AC-3.2, D-6 |
| T-7 | Swift | same: `pencilOpensTheSheet`, `linkToAGraphOpensAtGraphs`, `artifactGoneClosesTheSheet` *(new)*; `restingCardMarksAnArtifactInNoGraph` *(kept for its section half only: `artifacts.inNoGraph`; the card-edit half moves to T-5)*, `pageDropStagesFileWithoutCreating` *(updated: with no sheet open a drop stages Add as today; with a sheet open a drop stages nothing, `isAdding == false`)* | Pencil → `sheet` set, `opensAt == .top`; Link to a graph → `.graphs`; a workspace without the Artifact → `sheet == nil`, toast title "Artifact no longer here"; the drop cases as listed | AC-3.5, AC-4.3, D-11 |
| T-8 | Swift | same: `copyDetailsShowFourThenMore` *(new)* | `ArtifactCopyDetailsView.shown`: 0 rows → ([], 0); 3 → (3, 0); 8 → (first 4 in saved order, 4); empty suggestions are not shown | AC-4.1, D-16 |
| T-9 | Swift | `MetadataDraftHostTests.swift` `draftHostIsObservable` *(new)* | `withObservationTracking { _ = host.entries }` fires after `clear`, `dismissSuggestion`, `reorder` and `set` | D-18 |

**Removed with the behavior they pin:** `editingArtifactIsOneCardAtATime`, `cardSwitchAsksWhenDirty`, `cardSwitchConfirmMovesEdit`, `cardSwitchCancelKeepsEdit`, `cardSwitchCleanMovesAtOnce`, `beginEditOnTheSameCardKeepsItsEdits`, `cancelArtifactFieldsRestoresDrafts`, `cardCancelDropsLinkChanges`, `cardSaveErrorStaysOnTheCardThatSaved` (a modal sheet replaces "one card at a time"; Cancel discards the model with its links), and `MetadataBoardModelTests.addDialogWritesToItsTargetBoard` (the dialog goes).

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests`, `MetadataBoardModelTests`, `MetadataDraftHostTests`, `MetadataDraftHostParityTests`, `SourceGraphFormModelTests`, `PVPanelFormDialogTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — S15-D14 stays until S15-25 lands (its handoff); design/README.md notes S15-34 landed
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected; add-workspace-place § 6 per STEP-11
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a project with a census Source (type Census, which suggests Provider, URL, Accessed and Image number at the `artifact` level), three Evidence graphs, twelve Artifacts, one of them with three copy details and one with eight, and a Citation drawn on one Artifact in the first graph.

1. On the Source page at Regular (window 1366 × 900): the cards show their copy details (0, 3, 8 → four and "+4 more") (AC-4.1). The Metadata header shows no count, a help glyph whose tooltip is the intro, and a ghost "Add field" (AC-5.1). On a Source with no metadata and no suggestions, the empty state has no button and the header's Add field opens the row below it (AC-5.4).
2. Press "Add field" in the header: the row opens at the foot of the board; pick a field, type a value, press Return: it saves and the row closes (AC-5.2). Open it again and press Escape: it closes. No Add metadata dialog appears anywhere (AC-5.7).
3. Press an Artifact's pencil: the sheet "Edit ART-…" opens with Label, Description, Evidence graphs (the cited graph locked with its hint) and Copy details (AC-1.1, AC-1.5). Edit a URL row, press Escape: only the row closes (AC-2.4). Press Return in a row: the row confirms, the sheet stays (AC-2.5). Fill a suggestion, clear a row (no confirm), dismiss a suggestion, add a field inline, drag a row: nothing is written yet (AC-2.1). Change the label, press Save artifact: Save shows its spinner and the fields are disabled while it runs (AC-3.1); the sheet closes and the card shows the new values; the page does not flash a reload (AC-1.2).
4. Open it again, change something, press Escape with nothing open: "Discard changes?" with "Your changes will be lost."; Keep editing, then press Cancel: it closes at once (AC-2.2). Open it again, type into a suggestion without its Save and press Save artifact: the Callout asks to finish the field; press the suggestion's ✕, then Save artifact: it saves (AC-2.6). Type into a suggestion and press Escape: the text clears and the sheet stays (AC-2.4).
5. "+4 more" on the eight-detail card opens the sheet at Copy details; the body scrolls between a fixed title and footer (AC-4.2). "Link to a graph" on an Artifact in no graph opens at Evidence graphs (AC-4.3).
6. In a second window, delete the Artifact whose sheet is open in the first: the sheet closes and the toast "Artifact no longer here" shows (AC-3.5).
7. Stretch the window to Wide (pane ≥ 1600pt): the Metadata section is in the side column; open its add row: it fits (AC-5.6). Open the sheet: it is the same as at Regular.
8. VoiceOver: a card's copy details read "Provider, Ancestry" and so on; the pencil reads "Edit ART-7QK2M" (AC-4.4); the help glyph reads "About metadata".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A `ScrollView` inside a sheet collapses to zero height or grows to the screen | D-1 measures the form and caps the body; the manual check opens short and long sheets |
| Escape still reaches the sheet while a row is in edit (shortcut order) | D-2 stands the sheet's rule down while an inner edit is open; T-1 pins the rule, step 3 of the manual check the behavior |
| A local editor reports `setTyped` per keystroke and brings back the lag the editor's comment records | D-5 reports only when "differs from the starting text" flips |
| `showRefusal` cannot place a refused value (no `params`) | D-6 falls back to the top Callout |
| Porting the card tests changes what they assert | T-5 keeps each test's inputs and expectations; only the receiver changes |

**Stop and report if:**

- `updateArtifact` no longer returns the Artifact's rows, or `MetadataDraftHost.draft` no longer sends the whole order (S15-33, S15-36 changed since the planning base).
- A `PVInlineEdit` cancel button does not take Escape once the sheet's dismiss rule answers `.ignore` (D-2 cannot work as written).
- `.onExitCommand` on the add row does not receive Escape inside a sheet whose Escape button stands down (D-10).
- The scrolling body cannot be measured without changing `PVPanel`'s layout for every caller.
- Removing `replace` breaks `confirmAttach`'s cover or graph refresh in a way `show` plus `refreshCoverFromStore` does not cover.
- The scrolling body sizes wrong in a sheet with `onGeometryChange` (empty space or a collapsed body in the STEP-1 preview).
- The board's frames and an AC disagree.

## 12. Open questions

None.

## 13. PR

**Title:** S15-34 — Artifacts are edited in a sheet
**Body:** objective in one sentence; what shipped against US-1 – US-5; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-34-artifact-edit-sheet.md`.

## Landed

- `ArtifactSheetModel`'s `onSaved` carries the reply rows as optional (`[CatalogMetadataEntry]?`, nil when the draft was nil) instead of D-4's non-optional array, so a text-only Save leaves the page's rows and the cached key alone, as D-7 says "when `metadata` is set".
- The sheet model sets its graph choices and starts its lock read in `init` when the session's graph list is loaded (`graphsChanged()` is the one place both happen), instead of `openSheet` starting `linkLoadTask`.
- `ArtifactSheetModel.subtitle` ("<Source ref> · <Source title>") is a model property the view's copy reads; `SourceArtifactsSection.copyDetails(for:)` reads a card's rows from the page's workspace.
- `MetadataBoardModel` also clears a field's typed flag on a successful set (`applySet`) and on `cancelEdit`, not only in `dismissSuggestion`, `clear` and `resetDrafts`, so a filled suggestion cannot leave a stale flag that would block Save.
- `MetadataBoardDialogs.openAdd`'s callers and the `MetadataBoardView` `onAdd` parameter are gone with the dialog; the `.page` placement's empty state has no button.
- STEP-3 and STEP-4 are one commit (the header's Add field opens the inline row, so the two share code and strings).
- The manual check (section 10) was not run by the implementer; the Swift checks and `scripts/check-pr.sh` pass.
- Review round 1, user's call of 2026-11-13: the add row uses the kit's `PVInlineEditActions(.iconRow)` ✓/✕ (ids `.addRow.save` / `.addRow.cancel`), takes Escape through the ✕'s `.cancelAction` and has no `.onExitCommand`, and its picker takes focus on open; this supersedes D-10's rejected ".cancelAction ✕". A suggestion row raises `.pvSheetInnerEdit` only while its input is focused with text, which supersedes D-2's `hasTypedText` signal.
