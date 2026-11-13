# S15-34 — Artifacts are edited in a sheet

**Status:** Draft
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
**Review:** — (not reviewed yet)
**Size:** 11 steps across the Artifacts section, the Metadata section and two kit components, past the plan-pr limit of about 10. The user chose to keep it one PR over a three-way split (2026-11-13). Work the Steps in order and run each check; the commits are one per step
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
| AC-2.2 | Change anything, then press Cancel, Escape, or click outside | The kit's standard "Discard changes?" confirm ("Your changes will be lost.", Discard / Keep editing) |
| AC-2.3 | Type in a row and put the original text back, then Cancel | Closes at once (unchanged) |
| AC-2.4 | A row is in edit, or the add row is open, and press Escape | Only that row closes; the sheet stays. A second Escape (nothing open, nothing changed) closes the sheet |
| AC-2.5 | A row is in edit and press Return | The row's value is confirmed into the draft; the sheet does not save |

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
| One metadata value view for board rows and read-only copy details (inventory "Metadata value") | STEP-2, AC-4.1 |
| The inline Add field row on every board; the Add metadata dialog removed; the "field taken" notice (SA-13; S15-36 review UX-2) | STEP-3, AC-5.2 – AC-5.7 |
| The Metadata header: help tooltip, no count, ghost Add field (user, 2026-11-13); Wide | STEP-4, AC-5.1, AC-5.6 |
| The Artifact sheet's model in edit mode: draft, link locks, draft board, dirty, one-write Save, refusal mapping (SA-6, SA-8, SA-9) | STEP-5, AC-1.2 – AC-1.6, AC-2.1 – AC-2.3, AC-3.1 – AC-3.4 |
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
| Draft host | `macos/App/Features/Sources/MetadataDraftHost.swift` `MetadataDraftHost` (`isChanged`, `draft`, `confirmsClear == false`) | Ready for the sheet (S15-36). Unchanged |
| Live host | `macos/App/Features/Sources/MetadataBoardHost.swift` `SourcePageMetadataHost` | Unchanged |
| Metadata section | `macos/App/Features/Sources/SourcePageMetadataView.swift` `SourcePageMetadataView` | Header meta `metadataFieldCount`, primary "Add field" (`L10n.Sources.addMetadata`, comment says "dialog") opening the dialog, `metadataIntro` paragraph under the header |
| Page | `macos/App/Features/Sources/SourcePageView.swift` `.pvFormDialog(isPresented: addMetadataPresented …)` with `MetadataBoardAddForm`, `addMetadataPresented`, `.pvConfirm(item: … pendingEditSwitch …)` | Both go |
| Page model | `macos/App/Features/Sources/SourcePageModel.swift` `acceptsArtifactPageDrop` (reads `metadataDialogs.isAdding`, `artifacts.pendingEditSwitch`), `apply(_:)` (calls `artifacts.seedDrafts()`) | Updated |
| Graph form precedent | `macos/App/Features/Sources/SourceGraphFormModel.swift` `SourceGraphFormModel` (`Mode`, `draft`, `prefill`, `isDirty`, `loadKeptLinks`, `save() -> Bool`), `SourceGraphFormDialog.swift` `View.sourceGraphFormDialog(_:)` | The shape the sheet copies |
| Kit form dialog | `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift` `PVFormDialogContent`, `pvFormDialog(item:…)` | No scrolling body; Confirm has `.keyboardShortcut(.defaultAction)` always |
| Kit sheet | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` `PVSheetChrome` (a zero-size Escape button, `.keyboardShortcut(.cancelAction)`; discard copy `.discardChanges`) | Escape always goes to the sheet, ahead of an inline edit's own `.cancelAction` button (`PVInlineEdit`) |
| Kit panel | `macos/App/DesignSystem/Components/Panel/PVPanel.swift` `PVPanel.headerAndBody` | Title, subtitle and body in one stack |
| Section header | `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift` `PVSectionHeader(title:meta:aside:actions:)` | Has an aside slot |
| Icons | `macos/App/DesignSystem/Components/Icon/PVIcon.swift` `PVSymbol` | No help symbol (`info` = "info.circle" exists) |
| Store | `macos/App/Platform/GenealogyStore.swift` `updateArtifact(…, graphIDs:, metadata:) -> (artifact:, metadata:, effects:)`; `CatalogSourceWorkspace.metadataRows(for:)`, `setMetadataRows(_:for:)` | Ready (S15-33, S15-09) |
| Cache patch | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `.wroteMetadataBoard(sourceId:owner:entries:)`; `CatalogLastUpdated.swift` `metadataBoard()` | Patches the owner's rows into the cached Source workspace; no invalidation |
| Fake | `macos/App/Platform/FakeStore.swift` `updateArtifactCalls`, `updateArtifactGraphIDs`, `updateArtifactMetadata`, `updateArtifactError`, `updateArtifactHold`, `scriptedMetadataRows`, `getEvidenceGraphUnlinkImpactCalls`, `getEvidenceGraphUnlinkImpactPairs`, `getEvidenceGraphUnlinkImpactError`, `unlinkImpactBlocks`, `metadataWriteResults` | Enough for every test here |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` card tests (`cancelArtifactFieldsRestoresDrafts`, `editingArtifactIsOneCardAtATime`, `cardSwitch*`, `cardEdit*`, `cardSave*`, `cardCitationsLockHoldsUntilTheLinkIsSaved`, `cardLinkReadFailsAndReadAgain`, `cardCancelDropsLinkChanges`, `cardFieldsOnlySaveErrorSaysChangesKept`, `beginEditOnTheSameCardKeepsItsEdits`, `restingCardMarksAnArtifactInNoGraph`) and dialog tests (`addMetadataFromDialog`, `addMetadataDropdownSeesFieldCreatedElsewhere`, `addMetadataRequiresFieldAndValue`, `invalidMetadataAddLandsOnValueField`); `MetadataBoardModelTests.swift`; `MetadataDraftHostTests.swift`; `SourceGraphFormModelTests.swift` | Card and dialog tests are ported (§ 8) |

**Precedent.** S15-12's graph form (`SourceGraphFormModel` + `View.sourceGraphFormDialog`, `pvFormDialog(item:)` with `onConfirm: { await $0.save() }`): copy the model shape (a `Mode`, one `Draft` value, `prefill`, `isDirty = draft != prefill`, a load state for the engine's locks, `save() -> Bool`) and the presentation modifier. S15-36's **Next adds** is the recipe for the board inside it: a `MetadataDraftHost` over `CatalogSourceWorkspace.metadataRows(for:)` and a `MetadataBoardModel` on it; Save sends `metadata: host.draft`, applies `.wroteMetadataBoard` from the reply's rows, and on a failure calls `board.showRefusal` and shows the top Callout when it answers false; dirty reads `host.isChanged`. `PVConfirm`'s `.keyboardShortcut(tone.returnConfirms ? nil : .defaultAction)` is the precedent for a conditional shortcut. `PVTextArea`'s `PVTextAreaContentHeightKey` is the precedent for measuring content height with a `PreferenceKey`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Kit scrolling body.** `PVFormDialogContent` and both `pvFormDialog` overloads gain `scrollsBody: Bool = false` and `scrollTarget: AnyHashable? = nil`. When `scrollsBody`, `PVPanel` gets the form wrapped in `ScrollViewReader { ScrollView(.vertical) { form() } }` with the body's height set to the form's measured height (a private `PVFormDialogBodyHeightKey` PreferenceKey on a background `GeometryReader`) capped at `PVFormDialogContent.maxScrollingBodyHeight = 560` *(new)*; the title, subtitle and footer stay fixed. On appear, a non-nil `scrollTarget` scrolls to that id with `anchor: .top`. Every existing caller passes neither and draws as today | SA-6 (body scrolls between fixed title and buttons); S15-11 is the second caller; inventory "Artifact sheet: the scrolling body is opt-in and the kit's, not a local ScrollView" | A `ScrollView` inside the sheet's form — the inventory forbids it, and S15-11 would copy it; a fixed height — a short sheet would show empty space |
| D-2 | **Inner edit.** A content view reports an open inner edit with `View.pvSheetInnerEdit(_ open: Bool)` *(new, `PVSheet.swift`)*, a PreferenceKey reduced with `||`. `PVSheetChrome` reads it (`onPreferenceChange` into `@State innerEditOpen`), disables its zero-size Escape button while true (so the inner editor's own `.cancelAction` button takes Escape), and puts `\.pvSheetInnerEditOpen` *(new environment value)* on its content. `PVFormDialogContent` reads that value and sets its Confirm's shortcut to `innerEditOpen ? nil : .defaultAction`. A pure `PVSheetKeys` *(new enum)* holds the two rules (`escapeClosesSheet(innerEditOpen:)`, `confirmShortcut(innerEditOpen:)`) so T-1 tests them | SA-15: the kit owns the order, so S15-11's graph form gets it; one mechanism, read in the two places that own a key | A per-feature `onExitCommand` — each sheet would redo it; disabling Confirm itself — Save must stay clickable |
| D-3 | **One value view.** `MetadataValueText` *(new, `MetadataBoardView.swift`, internal)* takes an entry, an accessibility identifier and `lineLimit: Int?`; it draws today's `displayValue` (mono caption; a URL with `openURL` is the plain underlined link button in `PVColor.textLink` with the open-URL help). `lineLimit: nil` keeps today's wrapping on board rows; the card's copy details pass `1` with `.truncationMode(.tail)`. `MetadataBoardTextEditor.displayValue` calls it | Inventory "Metadata value" (brief review M-6): one view for both; not `PVButton .link`, whose font shifts the row | Two copies of the URL drawing |
| D-4 | **The sheet's model.** `ArtifactSheetModel` *(new, `macos/App/Features/Sources/ArtifactSheetModel.swift`)*, `@MainActor @Observable final class`, `enum Mode { case edit(CatalogArtifact) }` (S15-10 adds `create`), `enum Section: Hashable { case top, graphs, copyDetails }` and `let opensAt: Section`. `ArtifactSheetDraft` *(new)* `{ label, description }`; `prefill`; `labelError`; `saveError: String?`; `isSaving`. Links: the card's `LinkEdit` logic moves here unchanged in behavior as `graphIDs` (captured at open), `excludedGraphIDs`, `blocks`, `linkLoadState` (`SourceFormLoadState`, or nil when the session has no graph list) and the computed `lockedGraphs`, `checkedGraphIDs`, `graphs`, `inNoGraph`, `linksChanged`, `loadLinkLocks()`, `graphsChanged()`, `savedGraphIDs`. Board: `let host: MetadataDraftHost` built from `context.workspace?.metadataRows(for: .artifact owner) ?? []` and `context.fields`, and `let board: MetadataBoardModel(host:)` | The graph form's shape (precedent); S15-36 **Next adds**; the scope moves `LinkEdit`'s preflight, lock load and graph-keeps message into the sheet | Keeping link state on `SourceArtifactsSection` — two owners for one sheet's state |
| D-5 | **Dirty.** `isDirty = draft != prefill \|\| linksChanged \|\| host.isChanged \|\| board.hasUnconfirmedInput`. `MetadataBoardModel.hasUnconfirmedInput` *(new, public)*: an open add row with a picked field or a value with non-whitespace text, or `hasTypedEdit`. Label and description compare trimmed text | SA-8: a value typed and put back is unchanged (`isChanged` compares values; `hasTypedEdit` compares the buffer) | Counting an open but empty add row as a change |
| D-6 | **Save.** `save() async -> Bool` (true closes). Trims; empty label → `labelError = L10n.string(L10n.Sources.artifactLabelRequired)` ("Give the artifact a label."), false. A link read not `.ready` → false. `graphIDs = linksChanged ? savedGraphIDs : nil`; `metadata = host.draft`. Nothing changed (text equal, `graphIDs == nil`, `metadata == nil`) → true with no write. Else one `updateArtifact`; on success `SourceArtifactsSection.applySaved(artifact:metadata:effects:)` *(new, D-7)*, true. On failure: `context.fail(error)` nil → true (the page is leaving); else `board.showRefusal(error)` true → false (the row shows it); else `saveError` = today's `artifactCardSaveErrorGraphKeeps` / `artifactCardSaveErrorKept` body (the `saveErrorBody` rule, moved), and when `graphIDs != nil` the locks are read again first; false | SA-9; S15-33 D-6 (the refused value's field id is `params[0]`); one write per Save at Scale | A second write for the metadata — the working rule "Metadata saves with its owner's write" |
| D-7 | **Applying a Save.** `SourceArtifactsSection.applySaved(artifact:metadata:effects:)` *(new)*: `patchArtifacts` with the artifact replaced **and** `setMetadataRows(metadata, for: .artifact owner)` in the same workspace value (one `setWorkspace`), then `context.session.apply(effects, .wroteMetadataBoard(sourceId:, owner: .artifact, entries: metadata))`. `replace(_:effects:)` and its `applyWrite` call go (their last caller was `saveFields`; `confirmAttach` switches to `show`, D-12) | Map review K-1: never `.mutatedSourceWorkspace`; the page and the cache read the same rows | `context.applyWrite` — it reloads the page |
| D-8 | **Presentation.** `View.artifactSheet(_ model: Binding<ArtifactSheetModel?>)` *(new, `macos/App/Features/Sources/ArtifactSheet.swift`)* wraps `pvFormDialog(item:)` with `PVFormDialogCopy` with `verbatimTitle: L10n.Sources.artifactSheetEditTitle(ref: artifact.ref)`: title "Edit ART-7QK2M" (a `verbatim` title: `PVFormDialogCopy` gains `verbatimTitle: String?` *(new)*, used when set, as `verbatimSubtitle` is), subtitle verbatim "<Source ref> · <Source title>" ("SRC-4KD2P · 1880 U.S. census, Ottawa County, Kansas, population schedule"), confirm `L10n.Sources.saveArtifact` ("Save artifact"), cancel `L10n.Sources.cancelAction`; width 560; `isRunning: isSaving`; `isDirty`; `confirmDisabled: !canSave` (`canSave` = not saving, label has text, link read ready or absent); `scrollsBody: true`; `scrollTarget: opensAt` (nil for `.top`); prefix `sources.page.artifactSheet` | SA-6 copy (§ 2.1 "Edit ART-7QK2M"); the graph form's modifier | A `LocalizedStringResource` title with a format argument — `PVFormDialogCopy.title` is a resource; a verbatim title mirrors the existing verbatim subtitle |
| D-9 | **The sheet's body**, in order (each section carries `.id(Section)`): the top danger `PVCallout` (title `artifactCardSaveErrorTitle`, message `saveError`, `announces: true`) when `saveError` is set; `PVTabPair` Label (`PVField` required, `PVInput`, `activateOnAppear` when `opensAt == .top`) and Description (`PVField` with `artifactDescriptionHint`, `PVTextArea` `lineLimit: 3...8`); Evidence graphs: today's `graphLinksGroup` body (`SourceFormLoadArea` + `SourceArtifactGraphLinks` with `locked`, and the none line) moved from the card, shown only when the session has the Source's graphs; Copy details: `PVField(label: L10n.Sources.artifactCopyDetailsLabel)` *(new key, "Copy details")* around `MetadataBoardView(board:, placement: .form, …)`. Fields and board are `.disabled(isSaving)`. The body reports `.pvSheetInnerEdit(board.editingFieldID != nil \|\| board.addRow != nil)` | SA-6 order; SA-9 states; the graph group's states are today's | Drawing the graph group anew |
| D-10 | **Board placement.** `MetadataBoardView` replaces `onAdd` with `placement: MetadataBoardPlacement` *(new enum: `page`, `form`)*. `.page`: the empty state (when no saved and no suggested rows) keeps title and message with no button; no foot button; the add row draws at the foot when `board.addRow != nil`. `.form`: no empty state; the foot shows the ghost "Add field" button (`L10n.Sources.addMetadata`) when the row is closed, the row when open. The add row is `MetadataBoardAddRowView` *(new, private in `MetadataBoardView.swift`)*: a bordered row (`PVCard` default border, `PVRadius.sm`) holding `PVComboBox` (field; options `board.fieldComboOptions`; width 150 on `.page`, 168 on `.form`; placeholder and label `L10n.Sources.metadataField`, empty label `L10n.Sources.typeNoMatch`), `PVInput(size: .sm, mono: true, prompt: L10n.Sources.metadataValue)` (`onSubmit` confirms), `PVIconButton(.check)` disabled unless `canConfirmAddRow` (shows a `ProgressView` while `isSavingAddRow`), `PVIconButton(.dismiss)` with `.keyboardShortcut(.cancelAction)` that calls `closeAddRow`; under it the row's `error` (`PVErrorText`, caption) or the taken notice (D-13). Picking a field or typing clears the row's error | SA-13; the user's call that the Source's section has no foot button (2026-11-13); board decision (bordered row, 168 / 150pt) | Two booleans (`showsEmptyState`, `showsFootButton`) — one enum names the two real hosts; S15-11 is `.form` |
| D-11 | **Opening and closing the sheet.** `SourceArtifactsSection.sheet: ArtifactSheetModel?` *(new)*; `openSheet(artifactID:at:)` *(new)* builds it unless one is open or a save is running; the card's pencil opens `.top`, "Link to a graph" `.graphs`, "+n more" `.copyDetails`. `workspaceChanged()` *(new, replaces `seedDrafts`)*: when `sheet`'s Artifact is no longer in `items`, `sheet = nil` and `context.toast = VocabularyToast(title: L10n.string(L10n.Sources.toastArtifactGoneTitle), body: L10n.Sources.toastArtifactGoneBody(label:), tone: .info)` *(new keys)*, with the label the sheet opened with. The sheet's `graphsChanged()` runs from the sheet view's `onChange(of: visibleGraphs ids)`, as the card's did | SA-12, SA-14; one sheet at a time replaces "one card at a time" and its switch question | Keeping `pendingEditSwitch` — a sheet is modal, so no second card can ask |
| D-12 | **What goes from the Artifacts section.** `editingArtifactID`, `pendingEditSwitch`, `PendingEditSwitch`, `editGate`, `labels`, `descriptions`, `fieldErrors`, `savingID`, `cardSaveError`, `linkEdit`, `LinkEdit`, `linkLoadTask`, `seedDrafts`, `beginEdit`, `confirmEditSwitch`, `cancelEditSwitch`, `openEdit`, `startLinkEdit`, `awaitingGraphs`, `graphsChanged`, `dropLinkEdit`, `linkedGraphIDs`, `liveGraphIDs`, `lockedGraphs`, `checkedGraphIDs`, `linkEditGraphs`, `cardInNoGraph`, `linksDirty`, `loadLinkLocks`, `fieldsDirty`, `textChanged`, `canSaveFields`, `cardLabelError`, `cancelFields`, `savedGraphIDs`, `saveFields`, `saveErrorBody`, `replace`. `stagePageDrop` drops its `savingID` and `cancelFields` lines and returns when `sheet != nil`. `show` drops its `labels`/`descriptions` lines. `confirmAttach` calls `show(ingested.artifact, effects: ingested.effects)` instead of `replace`, as `create` already does, and keeps its `refreshCoverFromStore()` | The sheet owns edit; `replace` had two callers and its reload is the thing K-1 forbids | Keeping `replace` for `confirmAttach` — its `.mutatedSourceWorkspace` reload is unneeded once the list is patched (`create` already relies on `show`) |
| D-13 | **The taken notice.** `MetadataBoardAddRow` gains `takenFieldLabel: String?` *(new)*. `resetDrafts` sets it to the cleared field's label when it clears the row's `fieldID`; picking a field (`fieldID` set non-empty through the view) clears it. The add row shows `L10n.Sources.metadataAddRowFieldTaken(field:)` *(new)* ("%@ was added in another window. Pick another field.") as a muted caption and posts it with `.pvAnnounce` | S15-36 review UX-2 (the scope's line) | A toast — the cause is the row, and a toast would outlive it |
| D-14 | **Dialogs.** `MetadataBoardDialogs` keeps only `pendingDelete`, `isClearing`, `askClear`, `confirmClear` and `MetadataBoardDeleteItem` / `MetadataBoardAddResult`; every add member goes, and `MetadataBoardAddForm`. `SourcePageView` drops the Add metadata `pvFormDialog`, `addMetadataPresented`, and the `pendingEditSwitch` `pvConfirm`; it adds `.artifactSheet($model.artifacts.sheet)`. `acceptsArtifactPageDrop` drops `metadataDialogs.isAdding` and `artifacts.pendingEditSwitch == nil` and adds `artifacts.sheet == nil` | SA-13 "The Add metadata dialog goes"; S15-36 D-12 made the dialog a presenter of the row, so nothing else reads its add members | Keeping the dialog for the Source's board |
| D-15 | **Metadata header.** `SourcePageMetadataView`: `PVSectionHeader(title:, aside:, actions:)` with no `meta`; the aside is `PVIcon(.circleHelp, size: 15)` *(new `PVSymbol.circleHelp = "questionmark.circle"`)* in `PVColor.textMuted`, `.help(L10n.string(L10n.Sources.metadataIntro))`, `.accessibilityElement()`, `.accessibilityLabel(L10n.Sources.metadataHelpLabel)` *(new, "About metadata")*, `.accessibilityHint(L10n.Sources.metadataIntro)`; the action is `PVButton(L10n.Sources.addMetadata, variant: .ghost, size: .sm, icon: .plus)` calling `model.metadata.openAddRow()`, id `sources.page.addMetadata` kept; the intro `Text` goes; the board is `.page`. `metadataFieldCount` goes | User 2026-11-13 (board "Metadata" rows, `metaHelp`); the DesignSystem README says a macOS tooltip is `.help()` | Porting the web `PVTooltip` — README: port only when a call site needs richer content |
| D-16 | **Read-only copy details on today's card** (SA-4). Under the card's description: when the Artifact has saved values (`metadataRows(for:)` entries with `hasValue`, in saved order), a two-column grid of the first four (`MetadataBoardLabel`-free: field name in `PVTypeScale.caption`, `PVColor.textMuted`, one line; then `MetadataValueText(lineLimit: 1)`), then `PVButton("+n more", variant: .link, size: .sm)` (`L10n.Sources.artifactCopyDetailsMore(count:)` *(new)*) opening `.copyDetails`, its accessibility label `L10n.Sources.artifactCopyDetailsMoreLabel(count:ref:)` *(new)* ("4 more copy details of ART-7QK2M"). The block is one accessibility group whose children read "Field, value". Nothing when there are none. Ids `sources.page.artifact.<id>.copyDetails` and `sources.page.artifact.<id>.copyDetails.more` | SA-4, SA-11, inventory "Copy details, read-only": not a Table, not the board | Drawing the board read-only |
| D-17 | **Strings.** Kept: `addMetadata` (comment updated to "Button that opens the inline Add field row"), `metadataField`, `metadataValue`, `typeNoMatch`, `artifactLabelRequired`, `artifactCardSaveErrorTitle`, `artifactCardSaveErrorKept`, `artifactCardSaveErrorGraphKeeps` (comments updated from "card" to "Artifact sheet"), `saveArtifact`, `artifactDescriptionHint`. Changed: `editArtifactLabel(ref:)` → "Edit %@". New: `artifactSheetEditTitle(ref:)` ("Edit %@"), `artifactCopyDetailsLabel`, `artifactCopyDetailsMore(count:)`, `artifactCopyDetailsMoreLabel(count:ref:)`, `toastArtifactGoneTitle`, `toastArtifactGoneBody(label:)`, `metadataAddRowFieldTaken(field:)`, `metadataHelpLabel`, `metadataAddRowConfirmLabel` ("Add this field"), `metadataAddRowCancelLabel` ("Stop adding a field"). Removed: `addMetadataDialogTitle`, `addMetadataDialogSubtitle`, `addMetadataConfirm`, `metadataFieldHint`, `metadataValueHint`, `metadataFieldRequired`, `metadataValueRequired`, `metadataFieldCount`, `artifactCardLabelRequired`, and from `Localizable.xcstrings` | `add-localized-string` | — |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `scrollsBody`, `scrollTarget`, `maxScrollingBodyHeight`, `PVFormDialogBodyHeightKey` | kit params / key | `PVFormDialogContent`; `PVTextAreaContentHeightKey` | Extends the form dialog (opt-in); the key copies `PVTextArea`'s pattern |
| `PVFormDialogCopy.verbatimTitle` | field | `verbatimSubtitle` | Same shape for the title |
| `pvSheetInnerEdit(_:)`, `\.pvSheetInnerEditOpen`, `PVSheetKeys` | kit modifier / env / enum | `PVSheetChrome`, `PVSheetDismiss` | Extends the one sheet presenter; searched for an existing inner-edit signal (`onExitCommand`, `PreferenceKey` in `DesignSystem/`): none |
| `PVSymbol.circleHelp` | enum case | `PVSymbol.info` | A help glyph is not an info glyph (the board's `circle-help`) |
| `MetadataValueText` | view | `MetadataBoardTextEditor.displayValue` | Pulled out of it (D-3) |
| `MetadataBoardPlacement`, `MetadataBoardAddRowView` | enum / view | `MetadataBoardAddForm` (removed), the board's empty state | The row replaces the dialog's form |
| `MetadataBoardAddRow.takenFieldLabel`, `MetadataBoardModel.hasUnconfirmedInput` | field / computed | `resetDrafts`, `hasTypedEdit` | Extends the existing add row and edit state |
| `ArtifactSheetModel`, `ArtifactSheetDraft` | type | `SourceGraphFormModel`, `SourceArtifactsSection.LinkEdit` | Copies the graph form's shape; absorbs `LinkEdit` |
| `View.artifactSheet(_:)`, `ArtifactSheetForm` | modifier / view | `View.sourceGraphFormDialog(_:)`, `SourcePageArtifactsView.editFields` | Same shape as the graph form's modifier; the card's fields move here |
| `SourceArtifactsSection.sheet`, `openSheet(artifactID:at:)`, `applySaved(…)`, `workspaceChanged()` | members | `graphs.form` on `SourceGraphsSection`; `show`, `replace`, `seedDrafts` | Same host pattern as the graph form; `applySaved` replaces `replace` |
| L10n keys (D-17) | key | `L10n.Sources` | — |
| `ArtifactSheetModelTests` | test file | `SourceGraphFormModelTests.swift`, the card tests in `SourcePageModelTests.swift` | The card tests move here (§ 8) |

**Builds on.** `pvFormDialog(item:)` and `PVSheet`'s one close rule; S15-36's `MetadataDraftHost`, add row and `showRefusal`; S15-33's `updateArtifact(metadata:)` and its returned rows; S15-09's `patchArtifacts`, `setWorkspace`, `.wroteMetadataBoard` and `metadataRows(for:)`; the graph form's model shape; `SourceFormLoadArea` and `SourceArtifactGraphLinks`.

**Deletes.** Card edit (D-12's list) and its view (`editFields`, `graphLinks`, `graphLinksGroup`, `labelBinding`, `descriptionBinding`, `commitEdit`, the `editing` branches of `artifactCard` and `metaRow`); `SourceArtifactsSection.replace`; the Add metadata dialog (D-14) and `MetadataBoardAddForm`; the `pendingEditSwitch` confirm; the `metadataIntro` paragraph and `metadataFieldCount`; the strings in D-17's Removed list; the ported card and dialog tests. `PVDiscardGate` stays (the sheet chrome uses it).

**Next adds.** S15-35: the list redraws each entry and reuses D-16's copy-details block and `openSheet(…, at:)`. S15-10: `ArtifactSheetModel.Mode.create`, the File row first, `MetadataDraftHost.unsavedRows(from:)`, `createArtifact(…, metadata: host.draft)`; it removes the Add artifact dialog. S15-11: the graph form passes `scrollsBody: true`, draws `MetadataBoardView(placement: .form)` and reports `pvSheetInnerEdit`. S15-25: Move on the card menu.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Pencil, "Link to a graph", "+n more" open the sheet | `ArtifactSheetModel.init` | A new model per open (no reset needed); `openSheet` returns when the id is not in `items`; the graph list missing → no graph group (today's rule) |
| A page reload while the sheet is open | `SourceArtifactsSection.workspaceChanged()` | Artifact gone → closes with the toast (AC-3.5); Artifact still there → the sheet keeps its draft (the draft host does not read the workspace after open) |
| The Source deleted elsewhere during Save | `ArtifactSheetModel.save` | `context.fail` answers nil → true, the page leaves (today's rule) |
| A graph made or deleted while the sheet is open | `graphsChanged()`, `savedGraphIDs`, `graphs` | Moved unchanged from the card (T-5 ports `cardSaveKeepsAGraphMadeDuringEdit`, `cardSaveSkipsAGraphDeletedDuringEdit`) |
| The add row on the Source's board while a reload adds its field | `resetDrafts` | The taken notice (AC-5.5) |

### Scale

At the map's Scale (a Source with 60 Artifacts, 63 seeded fields): one Save is one `updateArtifact` write, and the page patches its workspace from the reply (one `setWorkspace`, one `.wroteMetadataBoard` patch of the cached key); no page reload and no list reload unless the links changed (the effects then reload the graph list, as today). Opening the sheet builds the draft host from one owner's rows (at most 63 plus extras) and runs one unlink preflight for the Artifact's linked graphs (as the card did). The cards' copy details read `metadataRows(for:)` per card from the workspace already in memory: 60 dictionary lookups per render, no store call.

### Edges

- **Assumptions.** `MetadataDraftHost.isChanged` compares saved values and empty suggestions, not open buffers; D-5 adds the buffers through `hasUnconfirmedInput`. `showRefusal` answers false when another row holds typed text (S15-36 Landed): the top Callout then shows (T-6). `savedGraphIDs` sends graphs made since open that link the Artifact (moved unchanged).
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
| Copy details, read-only | `Text` + `MetadataValueText` + `PVButton .link` | `SourcePageArtifactsView.swift` | D-16 |
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

**STEP-1 — Kit: a scrolling form body and inner edits that keep Escape and Return.**
Files: `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift`, `macos/App/DesignSystem/Components/Panel/PVSheet.swift`, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/PVSheetKeysTests.swift` *(new)*. Skill: `add-ui-component`.
Do: D-1, D-2, and `PVFormDialogCopy.verbatimTitle` (D-8). Add a `#Preview` of a scrolling form dialog. README: the FormDialog and Panel rows name the opt-in scrolling body and `pvSheetInnerEdit`.
Check: T-1 passes; `xcodebuild test` passes with every existing dialog unchanged (no caller passes the new parameters).

**STEP-2 — One metadata value view.**
Files: `macos/App/Features/Sources/MetadataBoardView.swift`.
Do: D-3.
Check: the build passes; the Source's board draws values as before (manual: a URL value still opens).

**STEP-3 — Every board adds a field in an inline row; the Add metadata dialog goes.**
Files: `MetadataBoardView.swift`, `MetadataBoardModel.swift`, `MetadataBoardDialogs.swift`, `SourcePageView.swift`, `SourcePageModel.swift` (`acceptsArtifactPageDrop`), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/MetadataBoardModelTests.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-10, D-13, D-14, D-5's `hasUnconfirmedInput`, the add-row strings of D-17, and remove the dialog strings. Port the dialog tests (§ 8 T-2).
Check: T-2 passes; `rg -n "MetadataBoardAddForm|addMetadataPresented|createFromAdd|addMetadataDialog" macos` finds nothing.

**STEP-4 — The Metadata header: help tooltip, no count.**
Files: `SourcePageMetadataView.swift`, `macos/App/DesignSystem/Components/Icon/PVIcon.swift`, `L10n.swift`, `Localizable.xcstrings`.
Do: D-15; the board's placement `.page`.
Check: build passes; manual: AC-5.1, AC-5.4, AC-5.6.

**STEP-5 — The Artifact sheet's model.**
Files: `macos/App/Features/Sources/ArtifactSheetModel.swift` *(new)*, `SourceArtifactsSection.swift` (`applySaved` only), `macos/ProvenenciaTests/ArtifactSheetModelTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-4, D-5, D-6, D-7.
Check: T-3 – T-6 pass.

**STEP-6 — The Artifact sheet's view.**
Files: `macos/App/Features/Sources/ArtifactSheet.swift` *(new)*, `L10n.swift`, `Localizable.xcstrings`.
Do: D-8, D-9 and the sheet strings of D-17.
Check: build passes; a `#Preview` of the sheet over a fake model renders the `sheet-edit` frame's content.

**STEP-7 — The card opens the sheet; card edit goes.**
Files: `SourceArtifactsSection.swift`, `SourcePageArtifactsView.swift`, `SourcePageView.swift`, `SourcePageModel.swift`, `SourcePageModelTests.swift`, `ArtifactSheetModelTests.swift`.
Do: D-11, D-12, D-14's sheet lines; the card pencil and "Link to a graph" call `openSheet`; port the card tests (§ 8 T-5, T-7).
Check: T-5, T-7 pass; `rg -n "editingArtifactID|saveFields|pendingEditSwitch|linkEdit|cardSaveError|cardLabelError|seedDrafts" macos` finds nothing.

**STEP-8 — Read-only copy details on the card.**
Files: `SourcePageArtifactsView.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcePageModelTests.swift`.
Do: D-16.
Check: T-8 passes; manual AC-4.1 – AC-4.4.

**STEP-9 — Strings sweep.**
Files: `L10n.swift`, `Localizable.xcstrings`.
Do: D-17's changed and removed keys and updated comments.
Check: `rg -n "artifactCardLabelRequired|metadataFieldCount|addMetadataConfirm|metadataFieldHint|metadataValueHint|metadataFieldRequired|metadataValueRequired" macos .agents docs` finds nothing outside archived plans.

**STEP-10 — Version bump.**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION`). Skill: `.cursor/rules/versioning.mdc`.
Do: 0.12.0 (or the next minor after `main`).
Check: the three agree.

**STEP-11 — Docs.**
Files: `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row), this plan (status, Landed note), `docs/deployment-plan/spike-15/design/README.md` (S15-D14 stays in `design/` until S15-25 lands, per its handoff), `docs/deployment-plan/spike-15/prs/S15-26-metadata-board-takes-an-owner.md` and `S15-09-artifact-metadata-boards.md` (**Next adds**: a line "Superseded by S15-34"), `.agents/skills/add-workspace-place/SKILL.md` § 6 "Metadata values" (name the Artifact sheet as the draft board's first owner Save).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` passes; links resolve.

Commits: one per step. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Edit an Artifact's label, graphs and copy details in one sheet.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PVSheetKeysTests.swift` *(new)* `escapeAndReturnYieldToAnInnerEdit` | `escapeClosesSheet(innerEditOpen: true)` → false, `false` → true; `confirmShortcut(innerEditOpen: true)` → nil, `false` → `.defaultAction` | D-2, AC-2.4, AC-2.5 |
| T-2 | Swift | `MetadataBoardModelTests.swift` `addRowSavesOnALiveBoard` *(ported from `addMetadataFromDialog`)*, `addRowNeedsFieldAndValue` *(from `addMetadataRequiresFieldAndValue`)*, `addRowRefusalStaysOnTheRow` *(from `invalidMetadataAddLandsOnValueField`)*, `addRowTakenFieldSaysSo` *(new)*, `hasUnconfirmedInputReadsOpenRows` *(new)*; `SourcePageModelTests.swift` `addRowPickerSeesFieldCreatedElsewhere` *(from `addMetadataDropdownSeesFieldCreatedElsewhere`)* | Live board: open, pick Author, "Eliza", confirm → row saved, `addRow == nil`; empty field or blank value → `canConfirmAddRow` false, no write; `sourcemetadata.invalid` → `addRow.error` set, row open; picked Collection then `resetDrafts` with Collection on the board → `fieldID == ""`, `takenFieldLabel == "Collection"`, value kept; an open empty row → `hasUnconfirmedInput` false, with a value → true | AC-5.2, AC-5.3, AC-5.5, D-5, D-13 |
| T-3 | Swift | `ArtifactSheetModelTests.swift` *(new)* `opensFromTheWorkspace`, `dirtyIgnoresPutBackText`, `labelRequired` | Opens with the Artifact's label, description, three copy-detail rows; type then put back → `isDirty` false; empty label → `save()` false, `labelError == "Give the artifact a label."`, no write | AC-1.1, AC-1.4, AC-2.3, D-5 |
| T-4 | Swift | same, `saveIsOneWriteWithTheDraft`, `saveWithNothingChangedWritesNothing`, `saveSendsNoLinksWhenOnlyTextChanged` *(from `cardSaveSendsNoLinksWhenOnlyTextChanged`)*, `savePatchesWithoutReload` | AC-1.2's edits → `updateArtifactCalls == 1`, the draft's values in that order, `graphIDs == nil`; nothing changed → 0 calls, `save()` true; the reply rows land in `workspace.metadataRows(for: artifact)` and the cached key, `getSourceWorkspace` not called | AC-1.2, AC-1.3, D-6, D-7 |
| T-5 | Swift | same, ported from the card tests: `oneGraphShowsTheGroup`, `locksEveryBlock`, `beforeTheListLoadsGetsTheGroup`, `citationsLockHoldsUntilTheLinkIsSaved`, `keepsAGraphMadeDuringEdit`, `linkReadFailsAndReadAgain`, `skipsAGraphDeletedDuringEdit`, `refusedSaveNamesTheGraphAndLocksIt`, `fieldsOnlySaveErrorSaysChangesKept`, `reloadFailureOffersReadAgain` | The same inputs and expectations, on `ArtifactSheetModel` instead of `model.artifacts` | AC-1.5, AC-1.6, AC-3.3, AC-3.4, D-4 |
| T-6 | Swift | same, `refusedValueLandsOnItsRow`, `refusedValueWithTypedTextElsewhereShowsTheCallout` | `sourcemetadata.invalid` with `params[0]` = URL's id → the URL row in edit with the message, `saveError == nil`; the same while the Provider row holds typed text → `saveError` set | AC-3.2, D-6 |
| T-7 | Swift | `SourcePageModelTests.swift` `pencilOpensTheSheet`, `linkToAGraphOpensAtGraphs`, `artifactGoneClosesTheSheet`, `restingCardMarksAnArtifactInNoGraph` *(kept)* | Pencil → `sheet` set, `opensAt == .top`; Link to a graph → `.graphs`; a workspace without the Artifact → `sheet == nil`, toast title "Artifact no longer here" | AC-3.5, AC-4.3, D-11 |
| T-8 | Swift | `SourcePageModelTests.swift` `cardCopyDetailsShowFourThenMore` | A pure helper `SourcePageArtifactsView.copyDetailsShown(_:)` *(new, static)*: 0 rows → ([], 0); 3 → (3, 0); 8 → (first 4 in saved order, 4) | AC-4.1, D-16 |

Removed with their behavior: `editingArtifactIsOneCardAtATime`, `cardSwitchAsksWhenDirty`, `cardSwitchConfirmMovesEdit`, `cardSwitchCancelKeepsEdit`, `cardSwitchCleanMovesAtOnce`, `beginEditOnTheSameCardKeepsItsEdits`, `cancelArtifactFieldsRestoresDrafts` (a modal sheet replaces "one card at a time"; Cancel discards the model).

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests`, `MetadataBoardModelTests`, `MetadataDraftHostTests`, `MetadataDraftHostParityTests`, `SourceGraphFormModelTests`.

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
4. Open it again, change something, press Escape with nothing open: "Discard changes?" with "Your changes will be lost." (AC-2.2).
5. "+4 more" on the eight-detail card opens the sheet at Copy details; the body scrolls between a fixed title and footer (AC-4.2). "Link to a graph" on an Artifact in no graph opens at Evidence graphs (AC-4.3).
6. In a second window, delete the Artifact whose sheet is open in the first: the sheet closes and the toast "Artifact no longer here" shows (AC-3.5).
7. Stretch the window to Wide (pane ≥ 1600pt): the Metadata section is in the side column; open its add row: it fits (AC-5.6). Open the sheet: it is the same as at Regular.
8. VoiceOver: a card's copy details read "Provider, Ancestry" and so on; the pencil reads "Edit ART-7QK2M" (AC-4.4); the help glyph reads "About metadata".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A `ScrollView` inside a sheet collapses to zero height or grows to the screen | D-1 measures the form and caps the body; the manual check opens short and long sheets |
| Escape still reaches the sheet while a row is in edit (shortcut order) | D-2 disables the sheet's button while an inner edit is open; T-1 pins the rule, step 3 of the manual check the behavior |
| `showRefusal` cannot place a refused value (no `params`) | D-6 falls back to the top Callout |
| Porting the card tests changes what they assert | T-5 keeps each test's inputs and expectations; only the receiver changes |

**Stop and report if:**

- `updateArtifact` no longer returns the Artifact's rows, or `MetadataDraftHost.draft` no longer sends the whole order (S15-33, S15-36 changed since the planning base).
- A `PVInlineEdit` cancel button does not take Escape once the sheet's Escape button is disabled (D-2 cannot work as written).
- The scrolling body cannot be measured without changing `PVPanel`'s layout for every caller.
- Removing `replace` breaks `confirmAttach`'s cover or graph refresh in a way `show` plus `refreshCoverFromStore` does not cover.
- The board's frames and an AC disagree.

## 12. Open questions

None.

## 13. PR

**Title:** S15-34 — Artifacts are edited in a sheet
**Body:** objective in one sentence; what shipped against US-1 – US-5; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-34-artifact-edit-sheet.md`.
