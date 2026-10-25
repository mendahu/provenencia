# S12-02 — Copy Source form

**Status:** Landed in [#485](https://github.com/mendahu/provenencia/pull/485) (2026-10-18)
**Landed:** Two differences. `sources.copy.artifactCount` is `%1$lld of %2$lld` (positional), because `scripts/check-localizable-xcstrings.py` requires numbered placeholders when a string has more than one argument. T-9 (`copyRefreshesSourcesList`) re-queries the list after the copy: from the Source page the list is not on screen, so `.createdSource` marks it stale (`whenHidden: .markStale`) and the next `session.query` reloads it; no write-through was added. Round 1 (user decisions 2026-10-18) changed two more things: `PVSkeletonBar` lives in `Components/SkeletonBar/PVSkeletonBar.swift` with its own group and README row, not `Components/List/` as D-23 named; and D-5's display rule is now that after the first edit to either Source field both field errors show (nothing on open), for Add Source and Copy Source alike. Round 3 (user-requested) made that rule follow only the researcher's own edits: `SourceTitleTypeFields` raises errors from wrapping bindings, so the app's changes to the draft (sole-type fill, Copy prefill) show none.
**Spike / slice:** Spike 12 · Slice 1 — Copy Source
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D1](../design/archive/S12-D1-copy-source-form.md), board agreed 2026-10-18
**Board:** [Claude Design — Source page, S12-D1 frames](https://claude.ai/design/p/de1e1ccc-aa35-455f-9b9c-3ae269593dc7?via=share) (the S12-D1 strip: From this page, Obituary, Some unchecked, Parts absent, Census 40, Loading, Load error, Validation, Save failed). If you cannot open it, §6 macOS describes every frame in words; build from §6
**Depends on:** S12-01 (landed [#478](https://github.com/mendahu/provenencia/pull/478)), S12-D1
**Blocks:** S12-12, S12-16, S12-17
**Branch:** `s12-02-copy-source-form`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** —
**Context:** [Copy Source](../source-structure.md#copy-source), working rules: "The form", "Entry points", "One title and type form, one Artifact checklist", "Two tools, one rule for choosing"; S12-01 ship log follow-up UX-1 ([`S12-01-copy-source-engine.ship.md`](S12-01-copy-source-engine.ship.md))
**Planned:** 2026-10-18 against `main` @ `817bad1` (S12-01 merged in #478; §4 rechecked: no file under §4 changed between the S12-01 branch head `77724d8` and `817bad1`)
**Review:** [`S12-02-copy-source-form.review.md`](S12-02-copy-source-form.review.md) — Fits with changes / Ready with changes → fixed, 2 rounds

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

A researcher can copy a Source from its page. A form with the Add Source title and type fields, prefilled `Copy of <title>` and the original's type, lets them keep or drop the description, metadata, credibility and each Artifact. Saving creates the copy through S12-01's `copySource` and opens its page. With S12-17 (the same form from a Sources list row's right-click menu) this is the dogfood bar's Copy Source outcome: an obituary starts from a sibling's publication, issue, page and URL instead of retyping them.

**Why now.** S12-01 put the engine and the store method in place so this PR is Swift only. S12-12 (Graph form) reuses the Artifact checklist built here, and S12-16 moves Add Source onto the effects path this PR uses.

**What stays.** Add Source keeps its fields, ids, validation copy, toast and write-through. One thing changes for it (user, 2026-10-18): the shared fields show the title and type errors as you edit, and Create is disabled while the form is invalid, the same rule Copy uses (D-5). The Source page header keeps its cover menu, Delete and Evidence graph jump; the jump moves to the start of the action row (board). The Sources list does not change (S12-17).

## 2. User stories

**US-1.** As a researcher on a Source page, when I choose Copy…, I get a form prefilled from this Source so that I can start a sibling Source in one step.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source "Obituary — Mary Alderwick", type Obituary, with a description, two metadata values, a credibility grade and two Artifacts; Copy… on its page | The form shows title `Copy of Obituary — Mary Alderwick`, type Obituary, Description, Metadata and Credibility checked, both Artifacts checked, All Artifacts checked |
| AC-1.2 | Same; Copy with nothing changed | A new Source opens with that title and type, the description, both metadata values, the credibility grade and two Artifacts with the same labels; the original is unchanged |
| AC-1.3 | Same; uncheck Credibility and the second Artifact; Copy | The copy has no credibility and one Artifact (the first one's label) |
| AC-1.4 | An original with an empty description and no credibility | Description and Credibility are disabled and unchecked with the hint "None on this Source"; the copy has an empty description and no credibility |
| AC-1.5 | An original with no Artifacts | The Artifacts group is replaced by "This Source has no Artifacts."; Copy sends no Artifact ids |
| AC-1.6 | Title cleared | The Add Source title error shows under the field and Copy is disabled; no write. A type left empty shows the Add Source type error |
| AC-1.7 | The form opens before the original's workspace is loaded (S12-17's list host; on the page it is cached) | The title shows at once; the type shows its label once the type list is loaded (the combo shows its placeholder until then); the parts area shows the loading state and Copy is disabled; once loaded, the parts and Artifacts show as in AC-1.1 |
| AC-1.9 | Add Source: open it, type a title, clear it | The title error shows under the field and Create is disabled; type a title and pick a type: the errors clear and Create is enabled |
| AC-1.8 | After a copy, open the Sources list | The copy is a row without a manual refresh |

**US-3.** As a researcher, when the copy fails, I keep my choices so that I can retry.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The store throws on `copySource` | The dialog stays open with a danger callout at the top titled "The copy could not be saved.", and every field and checkbox keeps its value |
| AC-3.4 | The original's workspace fails to load | A compact danger callout "The original's parts could not be read." with Read again replaces the parts area; Copy is disabled; Read again reloads it |
| AC-3.2 | An Artifact is deleted while the form is open (another window) | Its row disappears from the checklist when the workspace reloads, and its id is not sent |
| AC-3.3 | Change a checkbox, then click outside the sheet; or press Cancel | Outside click: the kit's "Discard changes?" confirm, as in every form dialog. Cancel closes at once (kit behavior; user, 2026-10-18) |

## 3. Scope

| In | Covered by |
| --- | --- |
| One Artifact display title and one fileless-icon rule on `CatalogArtifact`, used by every caller | STEP-1, D-1, D-2 |
| Shared title-and-type fields, one type combo row and one validation for Add Source, Copy Source and the Source page type editor (working rule "One title and type form"; S12-D1 inventory "Title and type fields") | STEP-2, AC-1.6 |
| `SourceCopyModel`: prefill, part availability, Artifact exclusions, dirty tracking, submit through `store.copySource` + `session.apply(effects, .createdSource)` (CS-4, CS-5, CS-9, CS-10) | STEP-3, AC-1.1 – AC-1.5, AC-1.7, AC-1.8, AC-3.1 – AC-3.3 |
| `SourceArtifactChecklist` snowflake with All Artifacts (CS-6, CS-7) | STEP-4, AC-1.1, AC-1.3, AC-1.5 |
| The Copy Source form dialog, its loading and error states, strings (CS-3, CS-8, CS-10) | STEP-5, AC-1.7, AC-3.1, AC-3.3 |
| The Copy action on the Source page header (CS-1) | STEP-6, AC-1.2 |
| One shared query-wait test helper for the two Sources suites | STEP-3, D-19 |

| Out | Home |
| --- | --- |
| Graph choices in the form (a copy starts with one Evidence graph) | not planned (map scope "Not this PR") |
| Add Source returning effects and dropping its write-through | S12-16 |
| Publishing the sidebar Sources count from the Source page host after a copy (the count catches up when the Sources list next loads) | S12-16 (one create path for both forms) |
| Promoting the Artifact checklist to a kit component | S12-12 (its second caller) |
| Moving `MetadataModelTests`, `PropertiesModelTests` and `SourceTypesModelTests` onto the shared query-wait helper | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) (STEP-7 adds the line) |
| The Copy action in a Sources list row's right-click menu, `SourcesModel.copying` / `openCopy` | S12-17 (brief S12-D6) |
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
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` (private `makeSession`, `makeModel`, `waitForQuery`, `warmLists`; `createRequiresTypeAndTitle`, `createReturnsSourceAndPublishesCount`), `SourcePageModelTests.swift` | Four suites keep a private `waitForQuery` (`SourcesModelTests`, `MetadataModelTests`, `PropertiesModelTests`, `SourceTypesModelTests`); it returns when `status` is `.ready` or `.error`, so it does not wait out a refetch (`QueryHandle.isFetching`). The Xcode project has no synchronized folders: every new Swift file is added to `macos/Provenencia.xcodeproj/project.pbxproj` (`add-ui-component`, `add-swift-test`) |

**Precedent.** Effects then mutation after a write: `EvidenceGraphModel.createSubject` (`macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` L1044–1053), `let (created, effects) = try await store.createSubject(...)` then `session.apply(effects, .mutatedSourceGraph(sourceId:))`. Copy that call shape. Open a Source's page: `SourcesListNavigation.pageLocation(for:)`. A baseline-compared dirty flag: `NameValueFormDialogModifier` (`macos/App/Features/Names/NameValueFormDialog.swift`, `isDirty: draft != baseline`). A right-click menu: `SourcePageIdentityHeader` `coverMenu`. A model reading the type pool: `SourcePageContext.types`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `extension CatalogArtifact` in `GenealogyStore.swift`, beside `openableRelPath`: `var displayTitle: String` *(new)* (trimmed label, else `ref`) and `var hasDisplayLabel: Bool` *(new)* (trimmed label non-empty). Callers move onto it: `SourcePageArtifactsView` L243, `SourcePageModel.askDeleteArtifact`, `EvidenceGraphFiles.fileBearing` (`title: artifact.displayTitle, titleIsRef: !artifact.hasDisplayLabel`) | One naming rule; the copies drifted (review M-1) | A fourth inline copy |
| D-2 | Same extension: `var hasFile: Bool` *(new)* (`file != nil && !fileID.isEmpty`) and `func thumbnailTypeIconKey(sourceTypeIconKey: String?) -> String?` *(new)* (nil when `hasFile`, else the given key). `SourcePageArtifactsView` deletes its private `hasFile(_:)` and `artifactTypeIconKey(for:)`; its call sites pass `model.identity.types.first { $0.id == model.source?.sourceTypeID }?.iconKey` | Per-Artifact rules live on the type (review M-2) | Putting the helper in the checklist file |
| D-3 | `SourcesModel.Draft` gains `var description: String = ""` (default), `var validation: SourceDraftValidation` *(new)* and `var trimmedTitle: String` *(new)*. `struct SourceDraftValidation: Equatable { var titleError: String?; var typeError: String?; var isValid: Bool }` (nested in `SourcesModel`) uses today's `typeRequired` / `titleRequired` strings. Stored model errors stay; `create()` and `copy()` still copy `validation` into them on submit (a guard: the button is disabled while invalid). `canSubmitAdd` becomes `!isSaving && draft.validation.isValid` and is wired: the Add Source `.pvFormDialog` passes `confirmDisabled: !model.canSubmitAdd` (user, 2026-10-18: one rule for both forms) | One rule, named apart from the shown state (review M-6, M-12) | Computed `Draft.typeError` beside stored `typeError`; a Copy-only rule |
| D-4 | `SourceTypeComboRow` *(new)* view in `macos/App/Features/Sources/SourceTypeComboRow.swift` (`type: CatalogSourceType?`, `option:`, `query:`) and `static func options(for types: [CatalogSourceType]) -> [PVComboBoxOption]` on it. `SourcesModel.typeComboOptions`, `SourceIdentitySection.typeComboOptions` and the Source page header's type combo use them | One look for a type in a picker (review M-3) | Moving only the list's copy |
| D-5 | `SourceTitleTypeFields` *(new)* in `macos/App/Features/Sources/SourceTitleTypeFields.swift`: `@Binding draft: SourcesModel.Draft`, `@Binding titleError: String?`, `@Binding typeError: String?`, `types: [CatalogSourceType]`, `accessibilityIdentifierPrefix: String`. Body = today's Title (`PVInput(text: $draft.title, activateOnAppear: true)`) and Type fields, options from D-4. It owns the one display rule for both forms: `.onChange(of: draft.title) { titleError = draft.validation.titleError }` and `.onChange(of: draft.sourceTypeID) { typeError = draft.validation.typeError }`, replacing today's type-only `typeError = nil`. Errors therefore show once a field is edited, not on open. `addForm` uses it with prefix `"sources.add"` (ids unchanged); Description and the callouts stay in `addForm` | Working rule; `addForm` is private (map scope); one rule owned by the shared view (review M-12) | Making `SourcesListContent` internal; a host-side `onChange` per form |
| D-6 | `SourceCopyModel` *(new)*, `@MainActor @Observable final class`, `macos/App/Features/Sources/SourceCopyModel.swift`. `init(original: CatalogSource, session: WorkspaceSession, userID: String, store: any GenealogyStore)` warms `.sourceWorkspace(project: session.projectKey, sourceId: original.id)` and `.sourceTypesList(project: session.projectKey)` with `session.query` and keeps both handles. Members: `let original`; `var choices: SourceCopyChoices`; `let prefill: SourceCopyChoices`; `var titleError: String?`, `var typeError: String?`, `var copyError: String?`; `private(set) var isSaving = false`; computed `workspace: CatalogSourceWorkspace?` (handle value), `canCopy` (`loadState == .ready && !isSaving && choices.draft.validation.isValid`), `loadState: SourceCopyLoadState` *(new enum: `.loading`, `.failed(String)`, `.ready`)* derived from the workspace handle with this precedence: a value present → `.ready` (a failed refetch keeps the stale value and stays `.ready`); else `status == .error` → `.failed(L10n.Errors.message(for: error))`; else `.loading` (covers `.idle`, `.loading`, and Read again after a failure, which sets `.loading`), `types: [CatalogSourceType]`, `sourceTypeIconKey: String?` (the original's type), `descriptionAvailable`, `metadataAvailable`, `credibilityAvailable` (D-8), `artifactIDsToCopy: [String]` (D-9), `isDirty` (`choices != prefill`). `func reloadWorkspace()` *(new)*: `session.invalidate` the workspace key, then `session.query` it (same handle). After a failure the query alone would reload (`WorkspaceSession.ensureQuery` reloads a handle that is not `.ready`); the invalidate makes Read again reload in every state, which T-13 relies on | The form needs its own state from two hosts; the view never calls `session.query` (review M-3) | Copy state on `SourcesModel` |
| D-7 | `struct SourceCopyChoices: Equatable` *(new)* in the same file: `draft: SourcesModel.Draft`, `includeDescription`, `includeMetadata`, `includeCredibility` (all `true`), `excludedArtifactIDs: Set<String>` (empty). `prefill` = `draft` with `title: L10n.Sources.copySourceTitlePrefill(original.title)`, `sourceTypeID: original.sourceTypeID`, `description: ""` | Dirty is one comparison; a new part is one field (review M-7) | Four hand comparisons |
| D-8 | Availability from `workspace`: description when `workspace.source.description` trimmed is non-empty; metadata when `workspace.metadata` is non-empty; credibility when `workspace.credibility != nil`. The value sent is `include… && …Available`. In the view an unavailable part is `isDisabled: true` and reads unchecked; its description line is `copySourcePartAbsent` instead of its hint | CS-5 | Hiding absent parts |
| D-9 | `artifactIDsToCopy` = `workspace.artifacts` ids in workspace order, minus `excludedArtifactIDs`. No seeding: a deleted Artifact drops out and a new one starts on (AC-3.2) | Exclusions need no view hook (review M-4); the live handle keeps the form current (S12-01 UX-1) | A seeded selection set |
| D-10 | `func copy() async -> CatalogSource?`: copy `choices.draft.validation` into `titleError` / `typeError`; return nil if invalid. Set `isSaving` (cleared by `defer`), clear `copyError`, call `store.copySource(projectDir: session.projectKey.projectDir, userID:, sourceID: original.id, title: choices.draft.trimmedTitle, sourceTypeID: choices.draft.sourceTypeID, copyDescription:, copyMetadata:, copyCredibility:, artifactIDs: artifactIDsToCopy)`, then `session.apply(effects, .createdSource)`, return the copy. No `setQueryValue`, no `syncCatalogCounts`. On throw: `copyError = L10n.Errors.message(for: error)`, return nil, `choices` untouched. The checkboxes and fields are disabled while `isSaving` (board) | Working rule forbids the write-through | Copying `create()`'s write-through |
| D-11 | No toast. Opening the copy is the confirmation | CS-9 | A toast as Add Source shows |
| D-12 | Host: `SourcePageModel.copying: SourceCopyModel?` + `func openCopy()` *(new)* (no-op when `source` is nil; passes `context.session`, `context.userID`, `context.store`). S12-17 adds the same pair to `SourcesModel` | Testable presentation state, as `isAdding` | `@State` in the view |
| D-13 | `View.sourceCopyDialog(_ model: Binding<SourceCopyModel?>)` *(new)* in `macos/App/Features/Sources/SourceCopyDialog.swift`, a `ViewModifier` that reads `@Environment(WorkspaceNavigation.self)`. `isPresented` getter = `model != nil`; setter `false` sets `model = nil` unless `isSaving`. `.pvFormDialog(isPresented:, copy: PVFormDialogCopy(title: copySourceDialogTitle, subtitle: copySourceDialogSubtitle(original.title), confirm: copySourceConfirm, cancel: L10n.Sources.cancelAction), width: 520, isRunning: isSaving, isDirty: isDirty, confirmDisabled: !canCopy, accessibilityIdentifierPrefix: "sources.copy")`. Discard follows the kit: Cancel closes; an outside click on a dirty form asks the kit's "Discard changes?" (user, 2026-10-18; the board's custom "Discard this copy?" is not built). Confirm: `Task { if let copy = await m.copy() { model = nil; navigation.go(to: SourcesListNavigation.pageLocation(for: copy)) } }` | One post-copy path for both hosts (review M-8) | An `onCopied` closure per host |
| D-14 | Form body, top to bottom (board frames, §6 macOS): (1) on `copyError`, `PVCallout(tone: .danger, title: copySourceSaveErrorTitle, message: copySourceSaveErrorMessage(copyError))` (`sources.copy.error`). (2) `SourceTitleTypeFields(draft: $model.choices.draft, ...)` (prefix `sources.copy`). (3) A `PVDivider()`, then `Text(copySourcePartsHeading).pvMicroCaps().foregroundStyle(PVColor.textFaint)` (review M-13). (4) `loadState == .ready`: three `PVCheckbox(_:description:isChecked:isDisabled:)` (L10n init; ids `sources.copy.description` / `.metadata` / `.credibility`; descriptions `copySourceDescriptionHint`, `copySourceMetadataHint`, `copySourceCredibilityHint`, or `copySourcePartAbsent` when unavailable), then `SourceArtifactChecklist` (`sources.copy.artifacts`). (5) `.loading`: three `PVSkeletonBar` *(new, D-23)* 12pt tall, widths 42 %, 36 %, 48 % of the column, with a `PVSpacing.space5` (12pt) gap between bars, then `Text(copySourceLoading)` italic `PVTypeScale.caption` `PVColor.textMuted`; `.accessibilityElement(children: .ignore)` labelled `copySourceLoading`. (6) `.failed(message)`: `PVCallout(tone: .danger, title: copySourceLoadErrorTitle, message: copySourceLoadErrorMessage(message), compact: true) { PVButton(copySourceReadAgain, variant: .secondary, size: .sm) { model.reloadWorkspace() } }` (`sources.copy.loadError`). The kit callout has no `detail` slot, so the raw error goes in the message, as Add Artifact's reject callout does. Body spacing: a `VStack(alignment: .leading, spacing: PVSpacing.space6)` (16pt) for (1)–(3) and between the parts group and the checklist; the three checkboxes in a `VStack(spacing: PVSpacing.space5)` (12pt); heading to first checkbox `PVSpacing.space5` | CS-3 – CS-9; board frames; `PVSectionHeader` is a page-level h3 | `PVSectionHeader`; `ProgressView`; local bars; adding `detail` to `PVCallout` for one caller |
| D-15 | Title focus: `activateOnAppear: true` (as Add Source); the prefill is not selected | `PVInput` has no select-all | A new `PVInput` option |
| D-16 | `SourceArtifactChecklist` *(new)*, `macos/App/Features/Sources/SourceArtifactChecklist.swift`: `artifacts: [CatalogArtifact]`, `@Binding excluded: Set<String>`, `projectDir: String`, `sourceTypeIconKey: String?`, `isDisabled: Bool`. Layout (board): `VStack(alignment: .leading, spacing: PVSpacing.space4)`; a header row, `Text(copySourceArtifactsHeading)` (`PVTypeScale.bodySmall`, medium, `PVColor.textPrimary`) and trailing `Text(copySourceArtifactCount(selected, total))` (`PVFont.mono(size: PVTypeScale.micro)`, `PVColor.textMuted`); under it one bordered box (1pt `PVColor.borderSubtle`, `PVRadius.sm`, clipped): an All Artifacts row on `PVColor.surfaceSunken`, its checkbox aligned with the row checkboxes (leading inset = row padding + 32pt thumbnail + `PVSpacing.space5`), with `PVCheckbox(copySourceAllArtifacts, isChecked:, isIndeterminate:)` (checked when nothing excluded; indeterminate when some; setting it true clears `excluded`, false sets every id), then a `ScrollView` capped at 264pt of Artifact rows, each separated by a `PVDivider()`, an `HStack(spacing: PVSpacing.space5)` of: `CachedThumbnail(projectDir:, relPath: art.thumbnailRelPath, mediaType: art.file?.mediaType ?? "", originalFilename: art.file?.originalFilename ?? "", typeIconKey: art.thumbnailTypeIconKey(sourceTypeIconKey:), size: 32)`, `PVCheckbox(verbatim: art.hasDisplayLabel ? art.displayTitle : "", isChecked:)` with `.accessibilityLabel(art.displayTitle)`, and a trailing `Text(art.ref)` (`PVFont.mono(size: PVTypeScale.micro)`, `PVColor.textMuted`). The ref is always shown, so an unlabeled Artifact shows only its ref (this plan's rule; the board's rows are all labelled). Row bindings: `Binding(get: { !excluded.contains(id) }, set: { on in if on { excluded.remove(id) } else { excluded.insert(id) } })`; All Artifacts: `Binding(get: { excluded.isEmpty }, set: { on in excluded = on ? [] : Set(artifacts.map(\.id)) })`, `isIndeterminate: !excluded.isEmpty && excluded.count < artifacts.count`. Row padding `PVSpacing.space2` vertical, `PVSpacing.space5` horizontal. Empty `artifacts` → only `Text(copySourceNoArtifacts)` italic `PVTypeScale.caption` `PVColor.textMuted`: no header row, no count, no box (board, Parts absent) | Brief inventory "Artifact checklist" (snowflake, New; S12-12 promotes it); board frames | A kit component now |
| D-17 | Page entry: `copySourceButton` *(new)* in `SourcePageIdentityHeader`, `PVIconButton(.duplicate, label: L10n.Sources.copySourceEntry, size: .sm) { model.openCopy() }`, disabled when `model.source == nil`, id `sources.page.copySource`. The actions `HStack` becomes `evidenceGraphJump`, `copySourceButton`, `deleteSourceButton` in that order (board frame 1: Jump to graph, Copy, Delete; today it is Delete, Jump). `SourcePageContent` adds `.sourceCopyDialog($model.copying)` to its dialog chain. The modifier takes only a `Binding<SourceCopyModel?>`, so S12-17's list host adds one line | CS-1; board; one presenter for both hosts | A text button; a toolbar; keeping today's order |
| D-18 | Copy tests are Swift Testing model tests over `FakeStore` in `macos/ProvenenciaTests/SourceCopyModelTests.swift` *(new)*. The failure case copies an original that is not in `sourcesByProject` (the fake throws `StoreBoom.boom`). No UI tests | The model holds every rule | A throw hook in `FakeStore` |
| D-19 | `macos/ProvenenciaTests/QueryHandleWaiting.swift` *(new)*: `@MainActor func waitForQuery<Value>(_ handle: QueryHandle<Value>) async` with the body of `SourcesModelTests`' private `waitForQuery`, changed to return only when `(status == .ready || status == .error) && !handle.isFetching`, so it also waits out a refetch of a stale value (T-9, T-11, T-13). `SourcesModelTests` drops its private copy; `SourceCopyModelTests` uses it. `MetadataModelTests`, `PropertiesModelTests` and `SourceTypesModelTests` are logged in `docs/ideas/shared-list-helpers.md` | Stops a fifth copy (review M-5); a refetch keeps `status == .ready` (`QueryHandle.beginLoading(stale:)`) | Copying the helper again; a second refetch helper |
| D-23 | `PVSkeletonBar` *(new)* in `macos/App/DesignSystem/Components/List/PVSkeletonBar.swift`: `struct PVSkeletonBar: View { var width: CGFloat?; var height: CGFloat }`, the body of `PVListSkeleton`'s private `bar(width:height:)` (`RoundedRectangle(cornerRadius: 2, style: .continuous).fill(PVColor.borderSubtle)`), `width` nil → fills the proposed width. `PVListSkeleton` uses it; its private `bar` goes. The copy form's loading state is its second caller | One loading-placeholder look (review M-14); content-agnostic, so a Component | Local bars with other tokens |
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
| `SourcePageModel.copying` / `openCopy` *(new)* | members | `SourcesModel.isAdding` / `openAdd()` | Same shape (D-12) |
| `L10n.Sources.copySource*` *(new)* | keys | `sources.add.*` | New copy |
| `waitForQuery` *(new, shared)* | helper | four private copies | Shared, and waits out a refetch (D-19) |
| `PVSkeletonBar` *(new)* | component | `PVListSkeleton`'s private `bar(width:height:)` | Lifted for its second caller (D-23) |
| `SourceCopyLoadState` *(new)* | enum | `QueryHandle.status` | Names the dialog's three bodies and their precedence (D-6) |
| `SourceCopyModelTests` *(new)* | test file | `SourcesModelTests` | New model |

**Builds on.** `GenealogyStore.copySource` (S12-01), `session.apply(_:_:)` with `.createdSource`, `PVFormDialog`, `PVCheckbox`, `PVContextMenu`, `CachedThumbnail`, `SourcesListNavigation.pageLocation(for:)`.
**Deletes.** Inline validation in `SourcesModel.create()`; `sourceTypeComboRow` and the title/type block in `SourcesListView.swift`; the inline type row in `SourcePageIdentityHeader`; the two `typeComboOptions` mapping bodies; the three label-else-ref expressions; private `hasFile(_:)` and `artifactTypeIconKey(for:)`; the hand-built Add Source `WorkspaceLocation`; `SourcesModelTests`' private `waitForQuery`; `PVListSkeleton`'s private `bar(width:height:)`; the unused-by-UI state of `SourcesModel.canSubmitAdd` (now wired, D-3).
**Next adds.** S12-12 promotes `SourceArtifactChecklist` with its exclusion binding. S12-16 moves Add Source onto `apply(effects, .createdSource)`, as `SourceCopyModel.copy()` does, and decides the sidebar count for both. S12-17 adds `SourcesModel.copying` / `openCopy`, the row menu and one `.sourceCopyDialog` line; the model, dialog, fields and checklist are reused unchanged.

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

The agreed board's frames, in words (dialog 520pt, kit `PVFormDialog` chrome):

1. **Header.** The Source page header action row reads, left to right: Jump to graph (secondary button), Copy (icon button), Delete (icon button). Copy is the same size and neutral tone as Delete.
2. **Obituary.** Title "Copy Source"; subtitle "Start a new Source from Obituary — Mary Alderwick. Notes and the Evidence graph are not copied." Title field (focused, prefilled "Copy of Obituary — Mary Alderwick"), Source type combo (Obituary, type icon). Divider. "COPY FROM THE ORIGINAL". Description / "The original's description, as written"; Metadata / "Values and field layout"; Credibility / "Grade and argument"; all checked. "Artifacts" with "2 of 2" at the right. Box: sunken All Artifacts row, checked; rows "Clipping, Yarmouth Mercury, 18 Jan 1902, p. 5" ART-0118 and "Funeral notice, same issue" ART-0119, both checked, 32pt thumbnails. Footer: Cancel, Copy (primary).
3. **Some unchecked.** Credibility unchecked; ART-0119 unchecked; All Artifacts shows the mixed mark; "1 of 2".
4. **Parts absent.** Description and Credibility disabled, unchecked, "None on this Source"; Metadata checked; "This Source has no Artifacts." in muted italics; no box.
5. **Census, 40.** "40 of 40"; about six rows visible, the rest scroll inside the box; the footer stays put. Unlabeled Artifacts show only their ref.
6. **Loading.** Title and type shown; under the heading three grey bars and "Reading the original's parts and Artifacts"; Copy disabled.
7. **Load error.** Under the heading a compact danger callout: "The original's parts could not be read." / "Title and type are kept. Try reading them again." plus the error, and a Read again button; Copy disabled.
8. **Validation.** Title empty and outlined in danger; the Add Source title error under it; Copy disabled.
9. **Save failed.** A danger callout above the Title field: "The copy could not be saved." / "Nothing was written. Your choices are kept." plus the error; choices as before; Copy enabled to retry.

Not built from the board: its custom "Discard this copy?" confirm on Cancel (D-13); its shorter validation wording (D-3 keeps Add Source's strings; user, 2026-10-18); its error-specific hints ("free some disk space, then copy again") and the boxed monospace error detail under a callout's message (the kit `PVCallout` has no `detail` slot; the error text ends the message instead, D-14). The board's after-copy note ("the app opens its page now") is a board annotation, not a toast (D-11).

Accessibility: every control has the id above; checkboxes carry their labels; unlabeled Artifact rows are labelled with their ref; the loading block reads as one label.

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
| `copySourceDescriptionHint` | `sources.copy.descriptionHint` | "The original's description, as written" |
| `copySourceArtifactCount(_ selected:, _ total:)` | `sources.copy.artifactCount` | "%lld of %lld" |
| `copySourceLoading` | `sources.copy.loading` | "Reading the original's parts and Artifacts" |
| `copySourceLoadErrorTitle` | `sources.copy.loadErrorTitle` | "The original's parts could not be read." |
| `copySourceLoadErrorMessage(_ error:)` | `sources.copy.loadErrorMessage` | "Title and type are kept. Try reading them again. %@" |
| `copySourceReadAgain` | `sources.copy.readAgain` | "Read again" |
| `copySourceSaveErrorTitle` | `sources.copy.saveErrorTitle` | "The copy could not be saved." |
| `copySourceSaveErrorMessage(_ error:)` | `sources.copy.saveErrorMessage` | "Nothing was written. Your choices are kept. %@" |

Cancel reuses `L10n.Sources.cancelAction`. Formatted keys follow `toastCreatedTitle(ref:)` (`L10n.format`). These follow the board's wording except where §6 macOS lists "Not built from the board"; apostrophes are typographic, as on the board.

## 7. Steps

Xcode check, used below as **XT**: `NSUnbufferedIO=YES xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/<Suite> CODE_SIGNING_ALLOWED=NO` (from `scripts/check-pr.sh`). Every new Swift file is added to `macos/Provenencia.xcodeproj/project.pbxproj` in its step (app target for `macos/App/…`, `ProvenenciaTests` target for tests).

**STEP-1 — One Artifact title and icon rule**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Features/Sources/SourcePageModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphFiles.swift`.
Do: D-1, D-2. Add T-1.
Check: XT with `SourcePageModelTests`, `CatalogArtifactTests` *(new suite in T-1)* pass; the Evidence graph file menu and Source page cards show the same titles and thumbnails.

**STEP-2 — Shared title and type fields**
Files: `SourcesModel.swift`, `SourcesListView.swift` (`addForm`, and `confirmDisabled: !model.canSubmitAdd` on the Add Source dialog), `SourceIdentitySection.swift`, `SourcePageIdentityHeader.swift` (type row only), `SourceTypeComboRow.swift` *(new)*, `SourceTitleTypeFields.swift` *(new)*. Skill: `add-ui-component`.
Do: D-3, D-4, D-5, D-20. Add T-2, T-14.
Check: XT with `SourcesModelTests` passes with its existing tests unchanged; Add Source and the Source page type editor look the same; manual check step 6 (AC-1.9).

**STEP-3 — Copy model and shared test helper**
Files: `SourceCopyModel.swift` *(new)*, `SourcePageModel.swift`, `L10n.swift` + `Localizable.xcstrings` (`copySourceTitlePrefill` only), `macos/ProvenenciaTests/QueryHandleWaiting.swift` *(new)*, `macos/ProvenenciaTests/SourcesModelTests.swift`, `macos/ProvenenciaTests/SourceCopyModelTests.swift` *(new)*, `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skills: `add-swift-test`, `add-workspace-place`, `add-localized-string`.
Do: D-6 – D-12, D-18, D-19. Add T-3 – T-13, T-15.
Check: XT with `SourceCopyModelTests`, `SourcesModelTests`, `SourcePageModelTests` passes.

**STEP-4 — Artifact checklist**
Files: `SourceArtifactChecklist.swift` *(new)*, `L10n.swift` + `Localizable.xcstrings` (`copySourceAllArtifacts`, `copySourceNoArtifacts`). Skills: `add-ui-component`, `add-localized-string`.
Do: D-16.
Check: build passes (XT any suite).

**STEP-5 — Copy dialog and the rest of the strings**
Files: `SourceCopyDialog.swift` *(new)*, `macos/App/DesignSystem/Components/List/PVSkeletonBar.swift` *(new)*, `macos/App/DesignSystem/Components/List/PVList.swift`, `L10n.swift`, `Localizable.xcstrings`. Skills: `add-localized-string`, `add-ui-component`.
Do: D-13 – D-15, D-23 and every §6 Strings key not added in STEP-3 or STEP-4.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; build passes.

**STEP-6 — Source page entry**
Files: `SourcePageIdentityHeader.swift`, `SourcePageView.swift`.
Do: D-17.
Check: manual check steps 1–4.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status), `docs/deployment-plan/spike-12/design/` (S12-D1 → `docs/deployment-plan/spike-12/design/archive/` *(new)*; the README's Scheduled row moves to an Archived line linking it), `docs/ideas/shared-list-helpers.md` (one line: `MetadataModelTests`, `PropertiesModelTests` and `SourceTypesModelTests` still keep a private `waitForQuery`; move them onto `QueryHandleWaiting.swift`).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-02-copy-source-form.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Copy a Source from its page or the Sources list."). No WIP commits.

## 8. Tests

Fixture for T-4 – T-11 (`SourceCopyModelTests`): project `projectDir`; type `t1` "Obituary"; Source `s1` "Obit", type `t1`, description "From the Herald"; `metadataBySource["s1"]` two entries; `credibilityBySource["s1"]` one; `artifactsBySource["s1"]` `a1` (label "Front page") and `a2` (label ""). Build the model, then `await waitForQuery(handle)` on the workspace handle (`session.query(.sourceWorkspace(project: session.projectKey, sourceId: "s1"))` returns the same handle the model holds).

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/CatalogArtifactTests.swift` *(new)* `displayTitleAndIcon` | label " Scan " → `displayTitle == "Scan"`, `hasDisplayLabel`; label "  " → `displayTitle == ref`, `!hasDisplayLabel`; fileless → `thumbnailTypeIconKey(sourceTypeIconKey: "k") == "k"`; with file + fileID → nil | D-1, D-2 |
| T-2 | Swift | `SourcesModelTests` `draftValidation` *(new)* | `Draft(sourceTypeID: "", title: "  ")` → `validation.typeError` / `titleError` equal `L10n.string(L10n.Sources.typeRequired)` / `titleRequired`, `!isValid`; `Draft(sourceTypeID: "t1", title: " Album ")` → both nil, `isValid`, `trimmedTitle == "Album"` | AC-1.6, D-3 |
| T-3 | Swift | `SourceCopyModelTests` `prefillsTitleAndType` *(new)* | fixture, before load → `choices.draft.title == "Copy of Obit"`, `sourceTypeID == "t1"`, `loadState == .loading`, `!canCopy`, `!isDirty` | AC-1.1, AC-1.7, D-6, D-7 |
| T-4 | Swift | `allPartsOnWhenLoaded` *(new)* | after load → three parts available and included, `artifactIDsToCopy == ["a1", "a2"]`, `loadState == .ready` | AC-1.1, D-8, D-9 |
| T-5 | Swift | `copyAllCreatesCopyAndReturnsIt` *(new)* | `copy()` → a Source titled "Copy of Obit", type `t1`; `store.sourcesByProject` count 2; the copy's metadata count 2, credibility set, Artifact labels ["Front page", ""]; `s1`'s entries unchanged | AC-1.2, D-10 |
| T-6 | Swift | `uncheckedPartsAreNotCopied` *(new)* | `includeCredibility = false`, `excludedArtifactIDs = ["a2"]` → copy has no credibility and one Artifact "Front page" | AC-1.3 |
| T-7 | Swift | `copyFailureKeepsChoices` *(new)* | original `s9` not seeded → `copy()` returns nil, `copyError != nil`, `choices` equal to before | AC-3.1, D-10, D-18 |
| T-8 | Swift | `absentPartsAreUnavailableAndNotSent` *(new)* | `s1` with description "", no credibility, no Artifacts → `!descriptionAvailable`, `!credibilityAvailable`, `artifactIDsToCopy == []`; copy has description "", no credibility, no Artifacts | AC-1.4, AC-1.5, D-8 |
| T-9 | Swift | `copyRefreshesSourcesList` *(new)* | `let list: QueryHandle<[CatalogSource]> = session.query(.sourcesList(project:))`, wait; `copy()`; wait → `list.value` contains the copy's id | AC-1.8, D-10 |
| T-10 | Swift | `validationBlocksCopy`, `isDirtyTracksChanges` *(new)* | title " " → `!canCopy`, `copy()` returns nil and sets `titleError`, store count 1; `includeMetadata = false` → `isDirty`; back to `true` → `!isDirty` | AC-1.6, AC-3.3, D-6, D-7 |
| T-11 | Swift | `artifactDeletedElsewhereIsDropped` *(new)* | after load, remove `a2` from `store.artifactsBySource["s1"]`, `session.invalidate(.sourceWorkspace(project: session.projectKey, sourceId: "s1"))`, re-query and wait → `artifactIDsToCopy == ["a1"]` | AC-3.2, D-9 |
| T-13 | Swift | `SourceCopyModelTests` `readAgainReloadsWorkspace` *(new)* | after load, note `store.getSourceWorkspaceCalls`; `reloadWorkspace()`, `waitForQuery` (D-19) → calls +1, `loadState == .ready` | AC-3.4, D-6 |
| T-14 | Swift | `SourcesModelTests` `canSubmitAddFollowsValidation` *(new)* | `openAdd()` with two types → `!canSubmitAdd`; set type `t1` and title "Album" → `canSubmitAdd`; title " " → `!canSubmitAdd` | AC-1.9, D-3 |
| T-15 | Swift | `SourceCopyModelTests` `loadStatePrecedence` *(new)* | before load → `.loading`; after load → `.ready`; an original id whose load fails is not reachable with `FakeStore` (no error hook, D-18), so `.failed` is covered by manual check step 7 | AC-1.7, AC-3.4, D-6 |
| T-12 | Swift | `SourcePageModelTests` `openCopyUsesThePageSource` *(new)* | page model after `sync` with a loaded workspace (the suite's existing setup) → `openCopy()` sets `copying`; with `source == nil` → stays nil | D-12 |

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

1. Open the obituary's page. The header's actions read Jump to graph, Copy, Delete. Choose Copy: the form shows `Copy of Obituary — Mary Alderwick`, type Obituary, three parts and both Artifacts checked, All Artifacts checked. (AC-1.1)
2. Uncheck Credibility and the second Artifact; All Artifacts shows mixed. Copy. The new Source's page opens: same type, description, Publication and Page values, one Artifact, no credibility, the type icon as cover. The original's page is unchanged. (AC-1.2, AC-1.3)
3. Copy the obituary again and clear the title: Add Source's title error shows under it and Copy is disabled; nothing is created. Uncheck Metadata and click outside the sheet: the kit's discard confirm shows; Keep editing, then Cancel: the sheet closes at once. (AC-1.6, AC-3.3)
4. Copy the second Source: Description and Credibility are disabled with "None on this Source"; the Artifacts line reads "This Source has no Artifacts." (AC-1.4, AC-1.5)
5. From the copy's page, open the Sources list: the copy is a row. (AC-1.8)
6. Open Add Source. Type a title, then clear it: the title error shows and Create is disabled. Type a title and pick a type: Create is enabled. (AC-1.9)

AC-1.7 (the loading state) is covered by T-3; on the page the workspace is already cached. S12-17's manual check exercises it from the list. Not checked by hand: the load-error state (AC-3.4) has no on-demand trigger in the app; T-13 covers Read again's reload and T-15 the state precedence.

With S12-17, this makes the slice 1 Check true: the new Source has "Copy of …", the same type, description, metadata in the same layout, credibility and the checked Artifacts, and no cover; unchecked parts are absent.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Extracting the fields changes Add Source's ids or behavior | D-5 keeps the `sources.add` prefix; `SourcesModelTests` must stay green unchanged |
| A 40-Artifact Source makes the dialog taller than the window | D-16 caps the rows at 264pt |
| The fake drops unknown Artifact ids, so a test cannot prove an id is not sent | T-11 asserts `artifactIDsToCopy`, not the fake's result |
| Moving the label rule changes Evidence graph file-menu titles for whitespace-only labels | Intended (the trimmed rule wins); `EvidenceGraphModelTests` `fileBearingListKeepsCoverAndDropsFileless` must stay green |

**Stop and report if:**

- `GenealogyStore.copySource` or `FakeStore.copySource` on `main` no longer matches §4.
- The agreed S12-D1 board differs from D-13 – D-17 or §6 Strings (another layout, a toast). The plan is updated before code.
- `.createdSource` no longer invalidates `.sourcesList` in `CatalogQueryRegistry.swift`, so T-9 cannot pass without a write-through.
- `SourceTitleTypeFields` cannot keep `sources.add.title` / `sources.add.type` without changing `SourcesListContent` beyond `addForm`.
- You are on a machine without Xcode (D-22): write the code, run the Python checks, and report that XT, `check-pr.sh` and the manual check are not run.

## 12. Open questions

None. The board link is on the **Board** line; Cancel and the title wording were answered by the user on 2026-10-18.

## 13. PR

**Title:** S12-02 — Copy Source form
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-02-copy-source-form.md`.
