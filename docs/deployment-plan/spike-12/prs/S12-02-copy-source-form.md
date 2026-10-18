# S12-02 — Copy Source form

**Status:** Blocked — on board S12-D1
**Spike / slice:** Spike 12 · Slice 1 — Copy Source
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D1](../design/S12-D1-copy-source-form.md), board not agreed yet
**Board:** — (to be set when the board is agreed: exported frames in the repo, or the frames described in words in §6 macOS)
**Depends on:** S12-01 (landed [#478](https://github.com/mendahu/provenencia/pull/478)), S12-D1
**Blocks:** S12-12, S12-16
**Branch:** `s12-02-copy-source-form`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** —
**Context:** [Copy Source](../source-structure.md#copy-source), working rules: "The form", "Entry points", "One title and type form, one Artifact checklist", "Two tools, one rule for choosing"; S12-01 ship log follow-up UX-1 ([`S12-01-copy-source-engine.ship.md`](S12-01-copy-source-engine.ship.md))
**Planned:** 2026-10-18 against `main` @ `817bad1` (S12-01 merged in #478; §4 rechecked: no file under §4 changed between the S12-01 branch head `77724d8` and `817bad1`)
**Review:** [`S12-02-copy-source-form.review.md`](S12-02-copy-source-form.review.md) — Fits with changes / Not ready → fixed, 1 round; plan stays Blocked on the board

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

A researcher can copy a Source from its page or from a Sources list row's right-click menu. A form with the Add Source title and type fields, prefilled `Copy of <title>` and the original's type, lets them keep or drop the description, metadata, credibility and each Artifact. Saving creates the copy through S12-01's `copySource` and opens its page. This is the dogfood bar's Copy Source outcome: an obituary starts from a sibling's publication, issue, page and URL instead of retyping them.

**Why now.** S12-01 put the engine and the store method in place so this PR is Swift only. S12-12 (Graph form) reuses the Artifact checklist built here, and S12-16 moves Add Source onto the effects path this PR uses.

**What stays.** Add Source looks and behaves as it does today (same fields, ids, validation copy, toast, write-through). The Source page header keeps its cover menu, Delete and Evidence graph jump. A Sources list row keeps its page and graph click zones.

## 2. User stories

**US-1.** As a researcher on a Source page, when I choose Copy…, I get a form prefilled from this Source so that I can start a sibling Source in one step.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source "Obituary — Mary Alderwick", type Obituary, with a description, two metadata values, a credibility grade and two Artifacts; Copy… on its page | The form shows title `Copy of Obituary — Mary Alderwick`, type Obituary, Description, Metadata and Credibility checked, both Artifacts checked, All Artifacts checked |
| AC-1.2 | Same; Copy with nothing changed | A new Source opens with that title and type, the description, both metadata values, the credibility grade and two Artifacts with the same labels; the original is unchanged |
| AC-1.3 | Same; uncheck Credibility and the second Artifact; Copy | The copy has no credibility and one Artifact (the first one's label) |
| AC-1.4 | An original with an empty description and no credibility | Description and Credibility are disabled and unchecked with the hint "None on this Source"; the copy has an empty description and no credibility |
| AC-1.5 | An original with no Artifacts | The Artifacts group is replaced by "This Source has no Artifacts."; Copy sends no Artifact ids |
| AC-1.6 | Title cleared, or type cleared; Copy | No write; the Add Source title or type error shows under that field |

**US-2.** As a researcher on the Sources list, when I right-click a row and choose Copy…, I get the same form for that Source.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Right-click a row | A menu with one item, Copy…, opens at the click point; left-click on the row still opens the page or the graph as today |
| AC-2.2 | Choose Copy… on a Source whose workspace is not cached | Title and type show at once; the parts area shows a loading state and Copy is disabled; once loaded, the parts and Artifacts show as in AC-1.1 |
| AC-2.3 | Copy | The new Source's page opens; going back to the Sources list shows the copy as a row without a manual refresh |

**US-3.** As a researcher, when the copy fails, I keep my choices so that I can retry.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The store throws on `copySource` | The dialog stays open with a danger callout at the top (`L10n.Errors.message(for:)`), and every field and checkbox keeps its value |
| AC-3.2 | An Artifact is deleted while the form is open (another window) | Its row disappears from the checklist when the workspace reloads, and its id is not sent |
| AC-3.3 | Cancel after changing a checkbox | The discard confirm shows, as Add Source's does; Cancel with the prefill untouched closes at once |

## 3. Scope

| In | Covered by |
| --- | --- |
| One Artifact display title and one fileless-icon rule on `CatalogArtifact`, used by every caller | STEP-1, D-1, D-2 |
| Shared title-and-type fields, one type combo row and one validation for Add Source, Copy Source and the Source page type editor (working rule "One title and type form"; S12-D1 inventory "Title and type fields") | STEP-2, AC-1.6 |
| `SourceCopyModel`: prefill, part availability, Artifact exclusions, dirty tracking, submit through `store.copySource` + `session.apply(effects, .createdSource)` (CS-4, CS-5, CS-9, CS-10) | STEP-3, AC-1.1 – AC-1.5, AC-2.3, AC-3.1 – AC-3.3 |
| `SourceArtifactChecklist` snowflake with All Artifacts (CS-6, CS-7) | STEP-4, AC-1.1, AC-1.3, AC-1.5 |
| The Copy Source form dialog, its loading and error states, strings (CS-3, CS-8, CS-10) | STEP-5, AC-2.2, AC-3.1, AC-3.3 |
| The Copy action on the Source page header (CS-1) | STEP-6, AC-1.2 |
| The Copy action in a Sources list row's context menu, with an accessibility action (CS-2) | STEP-7, AC-2.1, AC-2.3 |
| One shared query-wait test helper for the two Sources suites | STEP-3, D-19 |

| Out | Home |
| --- | --- |
| Graph choices in the form (a copy starts with one Evidence graph) | not planned (map scope "Not this PR") |
| Add Source returning effects and dropping its write-through | S12-16 |
| Publishing the sidebar Sources count from the Source page host after a copy (the count catches up when the Sources list next loads) | S12-16 (one create path for both forms) |
| Promoting the Artifact checklist to a kit component | S12-12 (its second caller) |
| A `pvActionMenu(items:)` recipe for the four hand-wired right-click menus | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) (STEP-8 adds the line) |
| Moving the other five test suites onto the shared query-wait helper | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) (STEP-8 adds the line) |
| Any other Sources list row menu item (Delete, Open graph) | not planned |
| A toast after Copy | not planned (D-11) |
| Copying notes, Subjects, Citations, Observations, the cover | not planned (working rule "The form") |
| Editing the description, metadata, credibility or an Artifact inside the form | not planned (S12-D1 §2.1) |
| Selecting the prefilled title text on open | not planned (`PVInput` has no select-all; the field is focused, D-15) |
| A friendlier engine error for a stale form (S12-01 UX-1) | not planned — D-9 keeps the form current instead |
| `FakeStore` error injection | not planned — T-7 uses an unknown original (D-18) |
| Any change to the Add Source description field, toast or write-through | not planned (S12-16 owns the write-through) |

## 4. Current state

Line numbers are at the planning base.

| Area | Where | Today |
| --- | --- | --- |
| Add Source form | `macos/App/Features/Sources/SourcesListView.swift` `SourcesListContent.addForm(typeOptions:)` (L389–443), `sourceTypeComboRow(option:query:)` (L445–456, reads `model.types`) | Private to `SourcesListContent`. Error callout `sources.add.error`, load-error callout, Title `PVField`+`PVInput(activateOnAppear: true)` (`sources.add.title`), Type `PVField`+`PVComboBox` (`sources.add.type`, clears `typeError` on change), Description `PVInput` |
| Add Source dialog | same file, L88–124, `addPresented` binding L136–147 | `.pvFormDialog(...accessibilityIdentifierPrefix: "sources.add"...)`; confirm runs `model.create()` then `navigation.go(to: WorkspaceLocation(section: .sources, sourceId:, ref:, title:))` built by hand (L100–106) |
| Add Source model | `macos/App/Features/Sources/SourcesModel.swift` `Draft` (L54–58: `sourceTypeID`, `title`, `description`, memberwise init), stored `typeError` / `titleError` / `createError` (L69–72), `typeComboOptions` (L313), `canSubmitAdd` (L317), `addSheetIsDirty` (L325), `openAdd()`, `cancelAdd()`, `create()` (L368–416) | Validation is inline in `create()` on a local trimmed title, writing `L10n.string(L10n.Sources.typeRequired)` / `titleRequired` into the stored errors. After the write: `session.apply(.createdSource)`, a `session.setQueryValue` insert into `.sourcesList`, `syncCatalogCounts()`, toast |
| Source page type editor | `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `metaRow` type combo (around L227–243, inline `PVMark` + `PVComboBoxPlainRow`); `SourceIdentitySection.typeComboOptions` (`SourceIdentitySection.swift` L45) | Second copy of the type combo row and of the options mapping |
| Sources list rows | `macos/App/Features/Sources/SourcesSplitRow.swift` `SourcesSplitRow` (L29), `pageZone` `Button(action: onOpenPage)` (L55); call sites `SourcesListView.swift` L313–334 and `SourcesSplitRowProgress` (L460–484) | Two zones (`onOpenPage`, `onOpenGraph`); `.accessibilityElement(children: .contain)` on the row. No context menu. Row data is `CatalogSource` only |
| Source page header | `SourcePageIdentityHeader.swift` actions `HStack` (L26–29), `deleteSourceButton` (L176–187), `coverMenu` / `coverKeyboard` / `coverMenuPanel` (L10–11, L33–39, L64–90) | `deleteSourceButton`, `evidenceGraphJump`. The cover menu is the right-click precedent |
| Source page shell | `macos/App/Features/Sources/SourcePageView.swift` `SourcePageContent` (L76), dialog chain from L142 | Toast overlay, three `.pvFormDialog`s, `.pvConfirm`s, `.pvDeleteImpact` |
| Source page model | `macos/App/Features/Sources/SourcePageModel.swift` `source`, `workspace`, `askDeleteSource()` (L164), `askDeleteArtifact` (L169–172, `art.label.isEmpty ? art.ref : art.label`); private `context: SourcePageContext` with `session`, `store`, `userID`, `projectDir`, `types` | No copy state |
| Artifact naming and icon | `SourcePageArtifactsView.swift` L243 (`art.label.isEmpty ? art.ref : art.label`), private `hasFile(_:)` (L360–362), private `artifactTypeIconKey(for:)` (L364–369, reads `model.identity.types`); `macos/App/Features/EvidenceGraph/EvidenceGraphFiles.swift` `fileBearing` (L32–40, trims the label first); `CatalogArtifact.openableRelPath` (`macos/App/Platform/GenealogyStore.swift` L84) | The label-else-ref rule is written three times and has drifted (one trims) |
| Store | `GenealogyStore.swift` `copySource(projectDir:userID:sourceID:title:sourceTypeID:copyDescription:copyMetadata:copyCredibility:artifactIDs:) async throws -> (CatalogSource, CatalogWriteEffects)` (from L988) | Added by S12-01. No caller |
| Fake store | `macos/App/Platform/FakeStore.swift` `copySource` (L367–416), `getSourceWorkspace` (L328) | Copies chosen parts; throws `StoreBoom.boom` when the original is not in `sourcesByProject`; silently drops unknown Artifact ids |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `apply(_:_:)` (L174), `query(_:)` (L87), `queryHandle(_:)` (L69), `invalidate(_:)` (L329); `CatalogMutation.createdSource`; `CatalogQueryRegistry.swift` `.sourcesList` / `.sourceTypesList` invalidate on `.createdSource` (L54–72) | `sourceScopes` do not insert rows; the mutation reloads or marks stale |
| Navigation | `macos/App/Features/Sources/SourcesListNavigation.swift` `pageLocation(for:)` (L6–14); `WorkspaceNavigation.go(to:)` | Used by list rows |
| Kit | `PVCheckbox` (`isIndeterminate:`, `isDisabled:`, `description:`), `PVContextMenu`, `PVFormDialog` (`width:`, `confirmDisabled:`), `PVField`, `PVIcon.duplicate` (`PVIcon.swift` L123), `CachedThumbnail` (`Platform/CachedThumbnail.swift`) | All exist. No checklist component |
| Strings | `macos/App/Platform/L10n.swift` `enum L10n.Sources` (from L4815), `sources.add.*` (L5053–5145); `macos/App/Resources/Localizable.xcstrings` | No `sources.copy.*` keys |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` (private `makeSession`, `makeModel`, `waitForQuery`, `warmLists`; `createRequiresTypeAndTitle`, `createReturnsSourceAndPublishesCount`), `SourcePageModelTests.swift` | Seven suites keep a private query-wait helper. The Xcode project has no synchronized folders: every new Swift file is added to `macos/Provenencia.xcodeproj/project.pbxproj` (`add-ui-component`, `add-swift-test`) |

**Precedent.** Effects then mutation after a write: `EvidenceGraphModel.createSubject` (`macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` L1044–1053), `let (created, effects) = try await store.createSubject(...)` then `session.apply(effects, .mutatedSourceGraph(sourceId:))`. Copy that call shape. Open a Source's page: `SourcesListNavigation.pageLocation(for:)`. A baseline-compared dirty flag: `NameValueFormDialogModifier` (`macos/App/Features/Names/NameValueFormDialog.swift`, `isDirty: draft != baseline`). A right-click menu: `SourcePageIdentityHeader` `coverMenu`. A model reading the type pool: `SourcePageContext.types`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `extension CatalogArtifact` in `GenealogyStore.swift`, beside `openableRelPath`: `var displayTitle: String` *(new)* (trimmed label, else `ref`) and `var hasDisplayLabel: Bool` *(new)* (trimmed label non-empty). Callers move onto it: `SourcePageArtifactsView` L243, `SourcePageModel.askDeleteArtifact`, `EvidenceGraphFiles.fileBearing` (`title: artifact.displayTitle, titleIsRef: !artifact.hasDisplayLabel`) | One naming rule; the copies drifted (review M-1) | A fourth inline copy |
| D-2 | Same extension: `var hasFile: Bool` *(new)* (`file != nil && !fileID.isEmpty`) and `func thumbnailTypeIconKey(sourceTypeIconKey: String?) -> String?` *(new)* (nil when `hasFile`, else the given key). `SourcePageArtifactsView` deletes its private `hasFile(_:)` and `artifactTypeIconKey(for:)`; its call sites pass `model.identity.types.first { $0.id == model.source?.sourceTypeID }?.iconKey` | Per-Artifact rules live on the type (review M-2) | Putting the helper in the checklist file |
| D-3 | `SourcesModel.Draft` gains `var description: String = ""` (default) and `var validation: SourceDraftValidation` *(new)*, where `struct SourceDraftValidation: Equatable { var titleError: String?; var typeError: String?; var isValid: Bool }` (nested in `SourcesModel`) uses today's `typeRequired` / `titleRequired` strings, and `var trimmedTitle: String` *(new)*. Stored model errors stay; `create()` and `copy()` copy `validation` into them on submit. `canSubmitAdd` uses `!isSaving && validation.isValid` | One rule, named apart from the shown state (review M-6) | Computed `Draft.typeError` beside stored `typeError` |
| D-4 | `SourceTypeComboRow` *(new)* view in `macos/App/Features/Sources/SourceTypeComboRow.swift` (`type: CatalogSourceType?`, `option:`, `query:`) and `static func options(for types: [CatalogSourceType]) -> [PVComboBoxOption]` on it. `SourcesModel.typeComboOptions`, `SourceIdentitySection.typeComboOptions` and the Source page header's type combo use them | One look for a type in a picker (review M-3) | Moving only the list's copy |
| D-5 | `SourceTitleTypeFields` *(new)* in `macos/App/Features/Sources/SourceTitleTypeFields.swift`: `@Binding title: String`, `@Binding sourceTypeID: String`, `@Binding titleError: String?`, `@Binding typeError: String?`, `types: [CatalogSourceType]`, `accessibilityIdentifierPrefix: String`. Body = today's Title (`PVInput(activateOnAppear: true)`) and Type fields, options from D-4, `typeError = nil` on type change. `addForm` uses it with prefix `"sources.add"` (ids unchanged); Description and the callouts stay in `addForm` | Working rule; `addForm` is private (map scope) | Making `SourcesListContent` internal |
| D-6 | `SourceCopyModel` *(new)*, `@MainActor @Observable final class`, `macos/App/Features/Sources/SourceCopyModel.swift`. `init(original: CatalogSource, session: WorkspaceSession, userID: String, store: any GenealogyStore)` warms `.sourceWorkspace(project: session.projectKey, sourceId: original.id)` and `.sourceTypesList(project: session.projectKey)` with `session.query` and keeps both handles. Members: `let original`; `var choices: SourceCopyChoices`; `let prefill: SourceCopyChoices`; `var titleError: String?`, `var typeError: String?`, `var copyError: String?`; `private(set) var isSaving = false`; computed `workspace: CatalogSourceWorkspace?` (handle value), `loadError: String?` (handle error through `L10n.Errors.message(for:)`), `isReady` (`workspace != nil && !isSaving`), `types: [CatalogSourceType]`, `sourceTypeIconKey: String?` (the original's type), `descriptionAvailable`, `metadataAvailable`, `credibilityAvailable` (D-8), `artifactIDsToCopy: [String]` (D-9), `isDirty` (`choices != prefill`) | The form needs its own state from two hosts; the view never calls `session.query` (review M-3) | Copy state on `SourcesModel` |
| D-7 | `struct SourceCopyChoices: Equatable` *(new)* in the same file: `draft: SourcesModel.Draft`, `includeDescription`, `includeMetadata`, `includeCredibility` (all `true`), `excludedArtifactIDs: Set<String>` (empty). `prefill` = `draft` with `title: L10n.Sources.copySourceTitlePrefill(original.title)`, `sourceTypeID: original.sourceTypeID`, `description: ""` | Dirty is one comparison; a new part is one field (review M-7) | Four hand comparisons |
| D-8 | Availability from `workspace`: description when `workspace.source.description` trimmed is non-empty; metadata when `workspace.metadata` is non-empty; credibility when `workspace.credibility != nil`. The value sent is `include… && …Available`. In the view an unavailable part is `isDisabled: true` and reads unchecked; its description line is `copySourcePartAbsent` instead of its hint | CS-5 | Hiding absent parts |
| D-9 | `artifactIDsToCopy` = `workspace.artifacts` ids in workspace order, minus `excludedArtifactIDs`. No seeding: a deleted Artifact drops out and a new one starts on (AC-3.2) | Exclusions need no view hook (review M-4); the live handle keeps the form current (S12-01 UX-1) | A seeded selection set |
| D-10 | `func copy() async -> CatalogSource?`: copy `choices.draft.validation` into `titleError` / `typeError`; return nil if invalid. Set `isSaving` (cleared by `defer`), clear `copyError`, call `store.copySource(projectDir: session.projectKey.projectDir, userID:, sourceID: original.id, title: choices.draft.trimmedTitle, sourceTypeID: choices.draft.sourceTypeID, copyDescription:, copyMetadata:, copyCredibility:, artifactIDs: artifactIDsToCopy)`, then `session.apply(effects, .createdSource)`, return the copy. No `setQueryValue`, no `syncCatalogCounts`. On throw: `copyError = L10n.Errors.message(for: error)`, return nil, `choices` untouched | Working rule forbids the write-through | Copying `create()`'s write-through |
| D-11 | No toast. Opening the copy is the confirmation | CS-9 | A toast as Add Source shows |
| D-12 | Hosts: `SourcesModel.copying: SourceCopyModel?` + `func openCopy(_ source: CatalogSource)` *(new)*; `SourcePageModel.copying: SourceCopyModel?` + `func openCopy()` *(new)* (no-op when `source` is nil; passes `context.session`, `context.userID`, `context.store`) | Testable presentation state, as `isAdding` | `@State` in the views |
| D-13 | `View.sourceCopyDialog(_ model: Binding<SourceCopyModel?>)` *(new)* in `macos/App/Features/Sources/SourceCopyDialog.swift`, a `ViewModifier` that reads `@Environment(WorkspaceNavigation.self)`. `isPresented` getter = `model != nil`; setter `false` sets `model = nil` unless `isSaving`. `.pvFormDialog(isPresented:, copy: PVFormDialogCopy(title: copySourceDialogTitle, subtitle: copySourceDialogSubtitle(original.title), confirm: copySourceConfirm, cancel: L10n.Sources.cancelAction), width: 520, isRunning: isSaving, isDirty: isDirty, confirmDisabled: !isReady, accessibilityIdentifierPrefix: "sources.copy")`. Confirm: `Task { if let copy = await m.copy() { model = nil; navigation.go(to: SourcesListNavigation.pageLocation(for: copy)) } }` | One post-copy path for both hosts (review M-8) | An `onCopied` closure per host |
| D-14 | Form body, top to bottom: `PVCallout(tone: .danger, message: copyError)` (`sources.copy.error`); `SourceTitleTypeFields` (prefix `sources.copy`); `PVField(label: copySourcePartsHeading)` holding three `PVCheckbox(_:description:isChecked:isDisabled:)` (L10n init; ids `sources.copy.description` / `.metadata` / `.credibility`; Description has no hint, Metadata `copySourceMetadataHint`, Credibility `copySourceCredibilityHint`); `PVField(label: copySourceArtifactsHeading)` holding the checklist (`sources.copy.artifacts`). While `workspace == nil && loadError == nil`: `ProgressView().controlSize(.small)` in place of both groups. On `loadError`: `PVCallout(tone: .danger, message: loadError)` (`sources.copy.loadError`) in place of both groups | CS-3 – CS-8; `PVSectionHeader` is a page-level h3, too large for a dialog | `PVSectionHeader` |
| D-15 | Title focus: `activateOnAppear: true` (as Add Source); the prefill is not selected | `PVInput` has no select-all | A new `PVInput` option |
| D-16 | `SourceArtifactChecklist` *(new)*, `macos/App/Features/Sources/SourceArtifactChecklist.swift`: `artifacts: [CatalogArtifact]`, `@Binding excluded: Set<String>`, `projectDir: String`, `sourceTypeIconKey: String?`. An All Artifacts `PVCheckbox(copySourceAllArtifacts, isChecked:, isIndeterminate:)` (checked when nothing excluded; indeterminate when some; setting it true clears `excluded`, false sets every id), then per Artifact an `HStack`: `CachedThumbnail(projectDir:, relPath: art.thumbnailRelPath, mediaType: art.file?.mediaType ?? "", originalFilename: art.file?.originalFilename ?? "", typeIconKey: art.thumbnailTypeIconKey(sourceTypeIconKey:), size: 32)` (default `.tile` layout) + `PVCheckbox(verbatim: art.displayTitle, isChecked:)` (binding inserts / removes the id in `excluded`). Rows in a `ScrollView` with `.frame(maxHeight: 240)`. Empty `artifacts` → `Text(copySourceNoArtifacts)` in `PVColor.textMuted`, `PVFont.body(size: PVTypeScale.bodySmall)` | Brief inventory "Artifact checklist" (snowflake, New; S12-12 promotes it) | A kit component now |
| D-17 | Page entry: `copySourceButton` *(new)* in `SourcePageIdentityHeader`, `PVIconButton(.duplicate, label: L10n.Sources.copySourceEntry, size: .sm) { model.openCopy() }`, placed before `deleteSourceButton`, disabled when `model.source == nil`, id `sources.page.copySource`. `SourcePageContent` adds `.sourceCopyDialog($model.copying)` to its dialog chain. List entry: `SourcesSplitRow` gains `var onCopy: () -> Void`, `@State rowMenu = PVContextMenuState()`, `@State rowKeyboard = PVContextMenuKeyboard.inactive`; on `pageZone` (not the whole row, so S12-14's graph sub-rows do not inherit it): `.pvContextMenuTrigger($rowMenu)`, `.pvContextMenu($rowMenu, keyboard: $rowKeyboard, keyboardTitles: [L10n.string(L10n.Sources.copySourceEntry)]) { PVContextMenuPanel(accessibilityIdentifier: "sources.row.menu") { PVContextMenuItem(L10n.Sources.copySourceEntry, index: 0, accessibilityIdentifier: "sources.row.copy") { onCopy() } } }`, and `.accessibilityAction(named: Text(L10n.Sources.copySourceEntry), onCopy)` on the `pageZone` button. Both call sites pass `onCopy: { model.openCopy(source) }`; `SourcesListContent` adds `.sourceCopyDialog($model.copying)` | CS-1, CS-2; the cover menu pattern | A native `.contextMenu`; a shared row builder for two one-line closures (review M-9, declined) |
| D-18 | Copy tests are Swift Testing model tests over `FakeStore` in `macos/ProvenenciaTests/SourceCopyModelTests.swift` *(new)*. The failure case copies an original that is not in `sourcesByProject` (the fake throws `StoreBoom.boom`). No UI tests | The model holds every rule | A throw hook in `FakeStore` |
| D-19 | `macos/ProvenenciaTests/QueryHandleWaiting.swift` *(new)*: `@MainActor func waitForQuery<Value>(_ handle: QueryHandle<Value>) async` with the body of `SourcesModelTests`' private `waitForQuery`. `SourcesModelTests` drops its private copy; `SourceCopyModelTests` uses it. The other five suites are logged in `docs/ideas/shared-list-helpers.md` | Stops the eighth copy (review M-5) | Copying the helper again |
| D-20 | Add Source's confirm builds its location with `SourcesListNavigation.pageLocation(for: created)` instead of the hand-built `WorkspaceLocation` | One way to open a Source's page (review M-8); same file as D-5 | Leaving two |
| D-21 | String symbols carry a `copySource` prefix (§6 Strings) | "copy" also means UI text in the kit (`PVFormDialogCopy`) (review M-11) | `copyDialogTitle` |
| D-22 | The Xcode build, Swift tests and the manual check run on the macOS dev machine, as for every PR (`scripts/check-pr.sh` `xcode-test` is macOS only; S12-01 D-16). An implementer on Linux stops after writing code and reports | The checks this PR depends on need Xcode | Skipping them |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CatalogArtifact.displayTitle`, `hasDisplayLabel`, `hasFile`, `thumbnailTypeIconKey` *(new)* | members | three inline label rules; private `hasFile`, `artifactTypeIconKey` | They replace them (D-1, D-2) |
| `SourceDraftValidation`, `Draft.validation`, `Draft.trimmedTitle` *(new)* | type / members | inline checks in `SourcesModel.create()` | Moved (D-3) |
| `SourceTypeComboRow` *(new)* | view | `SourcesListContent.sourceTypeComboRow`, `SourcePageIdentityHeader` inline row, two `typeComboOptions` | Replaces all of them (D-4) |
| `SourceTitleTypeFields` *(new)* | view | `SourcesListContent.addForm` | Extracted (D-5) |
| `SourceCopyModel`, `SourceCopyChoices` *(new)* | types | `SourcesModel` add state | Searched `Draft`, `isAdding`, `create(`: the add state lives on the list model, which the page lacks (D-6, D-7) |
| `SourceArtifactChecklist` *(new)* | view | none — searched `PVCheckbox` callers (`PromoteView`, `CitationComposerEvidenceColumn`) and Artifact lists (`SourcePageArtifactsView`) | Snowflake per brief (D-16) |
| `View.sourceCopyDialog` *(new)* | modifier | `NameValueFormDialogModifier` | Same shape, one presenter for two hosts (D-13) |
| `copying` / `openCopy` on two models *(new)* | members | `isAdding` / `openAdd()` | Same shape (D-12) |
| `SourcesSplitRow.onCopy` + row menu *(new)* | member | `SourcePageIdentityHeader.coverMenu` | Same kit menu (D-17) |
| `L10n.Sources.copySource*` *(new)* | keys | `sources.add.*` | New copy |
| `waitForQuery` *(new, shared)* | helper | seven private copies | Shared (D-19) |
| `SourceCopyModelTests` *(new)* | test file | `SourcesModelTests` | New model |

**Builds on.** `GenealogyStore.copySource` (S12-01), `session.apply(_:_:)` with `.createdSource`, `PVFormDialog`, `PVCheckbox`, `PVContextMenu`, `CachedThumbnail`, `SourcesListNavigation.pageLocation(for:)`.
**Deletes.** Inline validation in `SourcesModel.create()`; `sourceTypeComboRow` and the title/type block in `SourcesListView.swift`; the inline type row in `SourcePageIdentityHeader`; the two `typeComboOptions` mapping bodies; the three label-else-ref expressions; private `hasFile(_:)` and `artifactTypeIconKey(for:)`; the hand-built Add Source `WorkspaceLocation`; `SourcesModelTests`' private `waitForQuery`.
**Next adds.** S12-12 promotes `SourceArtifactChecklist` with its exclusion binding. S12-16 moves Add Source onto `apply(effects, .createdSource)`, as `SourceCopyModel.copy()` does, and decides the sidebar count for both. S12-14 puts its accordion under the row's page zone, which keeps the menu. More row menu items go in the `SourcesSplitRow` panel.

### Schema

None — no catalog change.

### Go core

None — S12-01 shipped the engine.

### FFI

None — `CopySource` and `GenealogyStore.copySource` exist (S12-01).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Copy Source dialog | `PVFormDialog` 520pt | `Features/Sources/SourceCopyDialog.swift` *(new)* | D-13, D-14 |
| Title and type | `PVField` + `PVInput`, `PVField` + `PVComboBox` + `SourceTypeComboRow` | `Features/Sources/SourceTitleTypeFields.swift` *(new)* | D-4, D-5 |
| Part checkboxes | `PVCheckbox` in a `PVField` group | `SourceCopyDialog.swift` | D-8, D-14 |
| Artifact checklist | `PVCheckbox` + `CachedThumbnail` | `Features/Sources/SourceArtifactChecklist.swift` *(new)* | D-16 |
| Save / load errors | `PVCallout` danger | `SourceCopyDialog.swift` | D-10, D-14 |
| Page Copy action | `PVIconButton(.duplicate)` | `SourcePageIdentityHeader.swift` | D-17 |
| Row Copy action | `PVContextMenu` on `pageZone` | `SourcesSplitRow.swift` | D-17 |

States: loading (list host) → spinner in the parts area, Copy disabled; load error → callout, Copy disabled; no Artifacts → one muted line; 40 Artifacts → rows scroll in 240pt; save error → callout, choices kept. Accessibility: every control has the id above; the page zone has a named accessibility action; checkboxes carry their labels.

When the board is agreed, replace this paragraph with the frames' layout and copy where they differ from D-13 – D-17 (handoff in [S12-D1](../design/S12-D1-copy-source-form.md) §7).

### Strings

L10n keys *(new)* in `L10n.Sources`, each with a `Localizable.xcstrings` entry:

| Symbol | Key | Value |
| --- | --- | --- |
| `copySourceEntry` | `sources.copy.entry` | "Copy…" |
| `copySourceDialogTitle` | `sources.copy.dialogTitle` | "Copy Source" |
| `copySourceDialogSubtitle(_ title:)` | `sources.copy.dialogSubtitle` | "Start a new Source from %@. Notes and the Evidence graph are not copied." |
| `copySourceConfirm` | `sources.copy.confirm` | "Copy" |
| `copySourceTitlePrefill(_ title:)` | `sources.copy.titlePrefill` | "Copy of %@" |
| `copySourcePartsHeading` | `sources.copy.partsHeading` | "Copy from the original" |
| `copySourceDescription` | `sources.copy.description` | "Description" |
| `copySourceMetadata` | `sources.copy.metadata` | "Metadata" |
| `copySourceMetadataHint` | `sources.copy.metadataHint` | "Values and field layout" |
| `copySourceCredibility` | `sources.copy.credibility` | "Credibility" |
| `copySourceCredibilityHint` | `sources.copy.credibilityHint` | "Grade and argument" |
| `copySourcePartAbsent` | `sources.copy.partAbsent` | "None on this Source" |
| `copySourceArtifactsHeading` | `sources.copy.artifactsHeading` | "Artifacts" |
| `copySourceAllArtifacts` | `sources.copy.allArtifacts` | "All Artifacts" |
| `copySourceNoArtifacts` | `sources.copy.noArtifacts` | "This Source has no Artifacts." |

Cancel reuses `L10n.Sources.cancelAction`. Formatted keys follow `toastCreatedTitle(ref:)` (`L10n.format`). The board's final copy replaces these values before status `Planned`.

## 7. Steps

Xcode check, used below as **XT**: `NSUnbufferedIO=YES xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/<Suite> CODE_SIGNING_ALLOWED=NO` (from `scripts/check-pr.sh`). Every new Swift file is added to `macos/Provenencia.xcodeproj/project.pbxproj` in its step (app target for `macos/App/…`, `ProvenenciaTests` target for tests).

**STEP-1 — One Artifact title and icon rule**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Features/Sources/SourcePageModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphFiles.swift`.
Do: D-1, D-2. Add T-1.
Check: XT with `SourcePageModelTests`, `CatalogArtifactTests` *(new suite in T-1)* pass; the Evidence graph file menu and Source page cards show the same titles and thumbnails.

**STEP-2 — Shared title and type fields**
Files: `SourcesModel.swift`, `SourcesListView.swift`, `SourceIdentitySection.swift`, `SourcePageIdentityHeader.swift` (type row only), `SourceTypeComboRow.swift` *(new)*, `SourceTitleTypeFields.swift` *(new)*. Skill: `add-ui-component`.
Do: D-3, D-4, D-5, D-20. Add T-2.
Check: XT with `SourcesModelTests` passes with its existing tests unchanged; Add Source and the Source page type editor look the same.

**STEP-3 — Copy model and shared test helper**
Files: `SourceCopyModel.swift` *(new)*, `SourcesModel.swift`, `SourcePageModel.swift`, `L10n.swift` + `Localizable.xcstrings` (`copySourceTitlePrefill` only), `macos/ProvenenciaTests/QueryHandleWaiting.swift` *(new)*, `macos/ProvenenciaTests/SourcesModelTests.swift`, `macos/ProvenenciaTests/SourceCopyModelTests.swift` *(new)*, `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skills: `add-swift-test`, `add-workspace-place`, `add-localized-string`.
Do: D-6 – D-12, D-18, D-19. Add T-3 – T-12.
Check: XT with `SourceCopyModelTests`, `SourcesModelTests`, `SourcePageModelTests` passes.

**STEP-4 — Artifact checklist**
Files: `SourceArtifactChecklist.swift` *(new)*, `L10n.swift` + `Localizable.xcstrings` (`copySourceAllArtifacts`, `copySourceNoArtifacts`). Skills: `add-ui-component`, `add-localized-string`.
Do: D-16.
Check: build passes (XT any suite).

**STEP-5 — Copy dialog and the rest of the strings**
Files: `SourceCopyDialog.swift` *(new)*, `L10n.swift`, `Localizable.xcstrings`. Skills: `add-localized-string`, `add-ui-component`.
Do: D-13 – D-15 and every §6 Strings key not added in STEP-3 or STEP-4.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; build passes.

**STEP-6 — Source page entry**
Files: `SourcePageIdentityHeader.swift`, `SourcePageView.swift`.
Do: D-17 (page half).
Check: manual check steps 1–4.

**STEP-7 — Sources list entry**
Files: `SourcesSplitRow.swift`, `SourcesListView.swift`.
Do: D-17 (list half).
Check: manual check steps 5–7.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status), `docs/deployment-plan/spike-12/design/` (S12-D1 → `docs/deployment-plan/spike-12/design/archive/` *(new)*; the README's Scheduled row moves to an Archived line linking it), `docs/ideas/shared-list-helpers.md` (two lines: the `pvActionMenu(items:)` recipe over the four right-click menus; the five suites left on private query-wait helpers).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-02-copy-source-form.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Copy a Source from its page or the Sources list."). No WIP commits.

## 8. Tests

Fixture for T-4 – T-11 (`SourceCopyModelTests`): project `projectDir`; type `t1` "Obituary"; Source `s1` "Obit", type `t1`, description "From the Herald"; `metadataBySource["s1"]` two entries; `credibilityBySource["s1"]` one; `artifactsBySource["s1"]` `a1` (label "Front page") and `a2` (label ""). Build the model, then `await waitForQuery(handle)` on the workspace handle (`session.query(.sourceWorkspace(project: session.projectKey, sourceId: "s1"))` returns the same handle the model holds).

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/CatalogArtifactTests.swift` *(new)* `displayTitleAndIcon` | label " Scan " → `displayTitle == "Scan"`, `hasDisplayLabel`; label "  " → `displayTitle == ref`, `!hasDisplayLabel`; fileless → `thumbnailTypeIconKey(sourceTypeIconKey: "k") == "k"`; with file + fileID → nil | D-1, D-2 |
| T-2 | Swift | `SourcesModelTests` `draftValidation` *(new)* | `Draft(sourceTypeID: "", title: "  ")` → `validation.typeError` / `titleError` equal `L10n.string(L10n.Sources.typeRequired)` / `titleRequired`, `!isValid`; `Draft(sourceTypeID: "t1", title: " Album ")` → both nil, `isValid`, `trimmedTitle == "Album"` | AC-1.6, D-3 |
| T-3 | Swift | `SourceCopyModelTests` `prefillsTitleAndType` *(new)* | fixture, before load → `choices.draft.title == "Copy of Obit"`, `sourceTypeID == "t1"`, `!isReady`, `!isDirty` | AC-1.1, AC-2.2, D-6, D-7 |
| T-4 | Swift | `allPartsOnWhenLoaded` *(new)* | after load → three parts available and included, `artifactIDsToCopy == ["a1", "a2"]`, `isReady` | AC-1.1, D-8, D-9 |
| T-5 | Swift | `copyAllCreatesCopyAndReturnsIt` *(new)* | `copy()` → a Source titled "Copy of Obit", type `t1`; `store.sourcesByProject` count 2; the copy's metadata count 2, credibility set, Artifact labels ["Front page", ""]; `s1`'s entries unchanged; `copying` untouched by the model | AC-1.2, D-10 |
| T-6 | Swift | `uncheckedPartsAreNotCopied` *(new)* | `includeCredibility = false`, `excludedArtifactIDs = ["a2"]` → copy has no credibility and one Artifact "Front page" | AC-1.3 |
| T-7 | Swift | `copyFailureKeepsChoices` *(new)* | original `s9` not seeded → `copy()` returns nil, `copyError != nil`, `choices` equal to before | AC-3.1, D-10, D-18 |
| T-8 | Swift | `absentPartsAreUnavailableAndNotSent` *(new)* | `s1` with description "", no credibility, no Artifacts → `!descriptionAvailable`, `!credibilityAvailable`, `artifactIDsToCopy == []`; copy has description "", no credibility, no Artifacts | AC-1.4, AC-1.5, D-8 |
| T-9 | Swift | `copyRefreshesSourcesList` *(new)* | `let list: QueryHandle<[CatalogSource]> = session.query(.sourcesList(project:))`, wait; `copy()`; wait → `list.value` contains the copy's id | AC-2.3, D-10 |
| T-10 | Swift | `validationBlocksCopy`, `isDirtyTracksChanges` *(new)* | title " " → nil, `titleError` set, store count 1; `includeMetadata = false` → `isDirty`; back to `true` → `!isDirty` | AC-1.6, AC-3.3, D-7 |
| T-11 | Swift | `artifactDeletedElsewhereIsDropped` *(new)* | after load, remove `a2` from `store.artifactsBySource["s1"]`, `session.invalidate(.sourceWorkspace(project: session.projectKey, sourceId: "s1"))`, re-query and wait → `artifactIDsToCopy == ["a1"]` | AC-3.2, D-9 |
| T-12 | Swift | `SourcesModelTests` `openCopySetsCopying` *(new)*; `SourcePageModelTests` `openCopyUsesThePageSource` *(new)* | `openCopy(source)` → `copying?.original.id == source.id`; page model after `sync` with a loaded workspace (the suite's existing setup) → `openCopy()` sets `copying`; with `source == nil` → stays nil | D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcesModelTests` (validation and helper moved), `SourcePageModelTests`, Evidence graph file menu tests that read `EvidenceGraphFiles.fileBearing` titles.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected (source-layer-data-model.md §4.2 Copy Source already describes the behavior, from S12-01)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a project with an Obituary Source "Obituary — Mary Alderwick" that has a description, two metadata values (Publication, Page), a credibility grade and two Artifacts (one with a file, one fileless), and a second Source with no description, no credibility and no Artifacts.

1. Open the obituary's page. Copy… sits before Delete in the header. Choose it: the form shows `Copy of Obituary — Mary Alderwick`, type Obituary, three parts and both Artifacts checked, All Artifacts checked. (AC-1.1)
2. Uncheck Credibility and the second Artifact; All Artifacts shows mixed. Copy. The new Source's page opens: same type, description, Publication and Page values, one Artifact, no credibility, the type icon as cover. The original's page is unchanged. (AC-1.2, AC-1.3)
3. Copy the obituary again, clear the title, Copy: Add Source's title error shows; nothing is created. Cancel: the discard confirm shows. (AC-1.6, AC-3.3)
4. Copy the second Source: Description and Credibility are disabled with "None on this Source"; the Artifacts line reads "This Source has no Artifacts." (AC-1.4, AC-1.5)
5. On the Sources list, left-click a row: its page opens as before. Right-click a row: a menu with Copy… opens at the pointer. (AC-2.1)
6. Restart the app, open the Sources list, right-click the obituary, Copy…: the parts area loads, then fills. Copy. The copy's page opens; Back to the list shows the copy as a row. (AC-2.2, AC-2.3)
7. With VoiceOver, the row's actions rotor offers Copy…. (CS-2)

This makes the slice 1 Check true: the new Source has "Copy of …", the same type, description, metadata in the same layout, credibility and the checked Artifacts, and no cover; unchecked parts are absent.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Extracting the fields changes Add Source's ids or behavior | D-5 keeps the `sources.add` prefix; `SourcesModelTests` must stay green unchanged |
| The row context menu catches left clicks | `pvContextMenuTrigger` only catches right / Ctrl-click (the cover menu works this way); manual check step 5 |
| A 40-Artifact Source makes the dialog taller than the window | D-16 caps the rows at 240pt |
| The fake drops unknown Artifact ids, so a test cannot prove an id is not sent | T-11 asserts `artifactIDsToCopy`, not the fake's result |
| Moving the label rule changes Evidence graph file-menu titles for whitespace-only labels | Intended (the trimmed rule wins); `EvidenceGraphModelTests` `fileBearingListKeepsCoverAndDropsFileless` must stay green |

**Stop and report if:**

- `GenealogyStore.copySource` or `FakeStore.copySource` on `main` no longer matches §4.
- The agreed S12-D1 board differs from D-13 – D-17 or §6 Strings (another layout, a toast, more menu items). The plan is updated before code.
- `.createdSource` no longer invalidates `.sourcesList` in `CatalogQueryRegistry.swift`, so T-9 cannot pass without a write-through.
- `SourceTitleTypeFields` cannot keep `sources.add.title` / `sources.add.type` without changing `SourcesListContent` beyond `addForm`.
- You are on a machine without Xcode (D-22): write the code, run the Python checks, and report that XT, `check-pr.sh` and the manual check are not run.
- `pvContextMenuTrigger` on `pageZone` swallows the page or graph button's left click.

## 12. Open questions

1. **How does the agreed S12-D1 board reach the implementer?** Recommendation: export the frames as PNGs into `docs/deployment-plan/spike-12/design/S12-D1/` *(new)* and link them on the **Board** line.

## 13. PR

**Title:** S12-02 — Copy Source form
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-02-copy-source-form.md`.
