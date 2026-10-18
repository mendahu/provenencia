# S12-02 — Copy Source form

**Status:** Blocked — on board S12-D1, and on S12-01 merging ([#478](https://github.com/mendahu/provenencia/pull/478) open)
**Spike / slice:** Spike 12 · Slice 1 — Copy Source
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D1](../design/S12-D1-copy-source-form.md), board not agreed yet
**Board:** — (to be set when the board is agreed: exported frames in the repo, or the frames described in words in §6 macOS)
**Depends on:** S12-01 ([#478](https://github.com/mendahu/provenencia/pull/478), open), S12-D1
**Blocks:** S12-12, S12-16
**Branch:** `s12-02-copy-source-form`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** —
**Context:** [Copy Source](../source-structure.md#copy-source), working rules: "The form", "Entry points", "One title and type form, one Artifact checklist", "Two tools, one rule for choosing"; S12-01 ship log follow-up UX-1 ([`S12-01-copy-source-engine.ship.md`](S12-01-copy-source-engine.ship.md))
**Planned:** 2026-10-18 against `s12-01-copy-source-engine` @ `77724d8` (= `main` @ `c01c4e0` + S12-01)
**Review:** — (not reviewed yet)

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
| Shared title-and-type fields and validation for Add Source and Copy Source (working rule "One title and type form"; S12-D1 inventory "Title and type fields") | STEP-1, AC-1.6 |
| `SourceCopyModel`: prefill, part availability, selection, submit through `store.copySource` + `session.apply(effects, .createdSource)` (CS-4, CS-5, CS-9) | STEP-2, AC-1.1 – AC-1.5, AC-2.3, AC-3.1, AC-3.2 |
| `SourceArtifactChecklist` snowflake with All Artifacts (CS-6, CS-7) | STEP-3, AC-1.1, AC-1.3, AC-1.5 |
| The Copy Source form dialog, its loading and error states, strings (CS-3, CS-8, CS-10) | STEP-4, AC-2.2, AC-3.1, AC-3.3 |
| The Copy action on the Source page header; opening the copy (CS-1) | STEP-5, AC-1.2 |
| The Copy action in a Sources list row's context menu, with an accessibility action (CS-2) | STEP-6, AC-2.1, AC-2.3 |

| Out | Home |
| --- | --- |
| Graph choices in the form (a copy starts with one Evidence graph) | not planned (map scope "Not this PR") |
| Add Source returning effects and dropping its write-through | S12-16 |
| Promoting the Artifact checklist to a kit component | S12-12 (its second caller) |
| Any other Sources list row menu item (Delete, Open graph) | not planned |
| A toast after Copy | not planned (D-9) |
| Copying notes, Subjects, Citations, Observations, the cover | not planned (working rule "The form") |
| Editing the description, metadata, credibility or an Artifact inside the form | not planned (S12-D1 §2.1) |
| A friendlier engine error for a stale form (S12-01 UX-1) | not planned — D-10 keeps the form current instead |
| `FakeStore` error injection | not planned — T-7 uses an unknown original (D-14) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Add Source form | `macos/App/Features/Sources/SourcesListView.swift` `SourcesListContent.addForm(typeOptions:)` (L389–443), `sourceTypeComboRow(option:query:)` (L445–456) | Private to `SourcesListContent`. Error callout `sources.add.error`, load-error callout, Title `PVField`+`PVInput` (`sources.add.title`), Type `PVField`+`PVComboBox` (`sources.add.type`, clears `typeError` on change), Description `PVInput`. Cannot be reused by a second form |
| Add Source dialog | same file, L88–124, `addPresented` binding L136–147 | `.pvFormDialog(isPresented:copy:isRunning:isDirty:accessibilityIdentifierPrefix:"sources.add":onConfirm:)`; confirm runs `model.create()` then `navigation.go(to: WorkspaceLocation(...))` |
| Add Source model | `macos/App/Features/Sources/SourcesModel.swift` `Draft` (L54–58), `create()` (L366–416), `canSubmitAdd` (L317), `addSheetIsDirty` (L325), `openAdd()`, `cancelAdd()` | Validation is inline in `create()`: `typeError = L10n.string(L10n.Sources.typeRequired)`, `titleError = L10n.string(L10n.Sources.titleRequired)` on the trimmed title. After the write: `session.apply(.createdSource)` plus a `session.setQueryValue` insert into `.sourcesList` |
| Sources list rows | `macos/App/Features/Sources/SourcesSplitRow.swift` `SourcesSplitRow` (L29); call sites `SourcesListView.swift` L313–334 and `SourcesSplitRowProgress` (L460–484) | Two zones (`onOpenPage`, `onOpenGraph`). No context menu. Row data is `CatalogSource` only (no Artifacts) |
| Source page header | `macos/App/Features/Sources/SourcePageIdentityHeader.swift` actions `HStack` (L26–29), `deleteSourceButton` (L176–187) | `deleteSourceButton`, `evidenceGraphJump`. Cover context menu on the header row (`coverMenu`, L33–39) is the precedent for a `PVContextMenu` |
| Source page shell | `macos/App/Features/Sources/SourcePageView.swift` `SourcePageContent` (L76), dialogs from L142 | Hosts the toast overlay, three `.pvFormDialog`s, `.pvConfirm`s and `.pvDeleteImpact`. The page's model is `SourcePageModel` |
| Source page model | `macos/App/Features/Sources/SourcePageModel.swift` `source`, `workspace`, `askDeleteSource()` (L164); `SourcePageContext` (`SourcePageContext.swift`) `session`, `store`, `userID`, `projectDir` | No copy state |
| Artifact thumbnail icon | `macos/App/Features/Sources/SourcePageArtifactsView.swift` private `artifactTypeIconKey(for:)` (L364–369) | Returns the Source type's icon key for a fileless Artifact. Private to the card view |
| Store | `macos/App/Platform/GenealogyStore.swift` `copySource(projectDir:userID:sourceID:title:sourceTypeID:copyDescription:copyMetadata:copyCredibility:artifactIDs:) async throws -> (CatalogSource, CatalogWriteEffects)` (L984–998) | Added by S12-01. No caller |
| Fake store | `macos/App/Platform/FakeStore.swift` `copySource` (L367–416), `getSourceWorkspace` (L328) | Copies chosen parts; throws `StoreBoom.boom` when the original is not in `sourcesByProject`; silently drops unknown Artifact ids; returns `bumpSource(copyID, evidence: !chosen.isEmpty)` |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `apply(_:_:)` (L174), `query(_:)` (L87); `CatalogMutation.createdSource` (`CatalogMutation.swift`); `CatalogQueryRegistry.swift` `.sourcesList` / `.sourceTypesList` invalidate on `.createdSource` (L54–72) | `apply(effects, .createdSource)` reloads or marks stale the Sources list and the type pool. `sourceScopes` do not insert rows |
| Navigation | `macos/App/Features/Sources/SourcesListNavigation.swift` `pageLocation(for:)`; `WorkspaceNavigation.go(to:)` | Used by list rows to open a page |
| Kit | `PVCheckbox` (`DesignSystem/Components/Checkbox/PVCheckbox.swift`, `isIndeterminate:`, `isDisabled:`, `description:`), `PVContextMenu` (`DesignSystem/Components/ContextMenu/PVContextMenu.swift`), `PVFormDialog` (`width:`, `confirmDisabled:`), `PVIcon.duplicate` (`DesignSystem/Components/Icon/PVIcon.swift` L123), `CachedThumbnail` (`Platform/CachedThumbnail.swift`) | All exist. No checklist component |
| Strings | `macos/App/Platform/L10n.swift` `enum L10n.Sources` (from L4815), `sources.add.*` (L5053–5145), `sources.page.*`; `macos/App/Resources/Localizable.xcstrings` | No `sources.copy.*` keys |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` (`makeModel`, `warmLists`, `createRequiresTypeAndTitle`, `createReturnsSourceAndPublishesCount`), `SourcePageModelTests.swift` | Add Source covered; no copy tests. `OnboardingModelTests.swift` `ThrowingStore.copySource` stub exists |

**Precedent.** Effects then mutation after a write: `EvidenceGraphModel.createSubject` (`macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` L1044–1053), `let (created, effects) = try await store.createSubject(...)` then `session.apply(effects, .mutatedSourceGraph(sourceId:))`. Copy that call shape. Create then open the page: `SourcesListView.swift` L100–110. Copy the navigation, not the `setQueryValue` insert in `SourcesModel.create()`. A right-click menu: `SourcePageIdentityHeader` `coverMenu` / `coverKeyboard` / `coverMenuPanel`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Title-and-type validation moves onto `SourcesModel.Draft` as `var trimmedTitle: String`, `var typeError: String?`, `var titleError: String?` *(new)* (the current Add Source strings, unchanged). `SourcesModel.create()` and `canSubmitAdd` use them; behavior and copy are unchanged | Working rule: one validation for both forms | A second validator in the copy model |
| D-2 | `SourceTitleTypeFields` *(new)* in `macos/App/Features/Sources/SourceTitleTypeFields.swift`: `@Binding draft: SourcesModel.Draft`, `@Binding titleError: String?`, `@Binding typeError: String?`, `typeOptions: [PVComboBoxOption]`, `types: [CatalogSourceType]`, `accessibilityIdentifierPrefix: String`. Body = today's Title and Type fields from `addForm`, plus `sourceTypeComboRow` moved here. `addForm` calls it with prefix `"sources.add"`, so ids stay `sources.add.title` / `sources.add.type`. Description and the callouts stay in `addForm` | Working rule; `addForm` is private today (map scope) | Making `SourcesListContent` internal; copying the fields |
| D-3 | The copy model holds `draft: SourcesModel.Draft` with `description` always `""`. The description part is a checkbox, so the field is not shown and not sent | Working rule names `SourcesModel.Draft`; one draft type for title + type | A new `CopyDraft` type |
| D-4 | `SourceCopyModel` *(new)*, `@MainActor @Observable final class`, `macos/App/Features/Sources/SourceCopyModel.swift`. `init(original: CatalogSource, session: WorkspaceSession, userID: String, store: any GenealogyStore)`. On init it sets `draft = Draft(sourceTypeID: original.sourceTypeID, title: L10n.Sources.copyTitlePrefill(original.title), description: "")` and `workspaceHandle = session.query(.sourceWorkspace(project:sourceId:))` (the key the Source page uses). One instance per opened form | The form needs its own state from two hosts; the page and the list share it | Putting copy state on `SourcesModel` (the page has none) |
| D-5 | Hosts own `var copying: SourceCopyModel?` *(new)* and `func openCopy(...)` *(new)*: `SourcesModel.openCopy(_ source: CatalogSource)` and `SourcePageModel.openCopy()` (uses `source`, no-op when nil). Both pass their own `session`, `userID`, `store` | Testable presentation state, as `isAdding` is | `@State` in the views |
| D-6 | Part availability from the loaded workspace: description when `workspace.source.description` trimmed is non-empty; metadata when `workspace.metadata` is non-empty; credibility when `workspace.credibility != nil`. Each `include…` flag starts `true`; the value sent is `include && available`. Unavailable parts show disabled and unchecked | CS-5; the engine copies what it is told | Hiding absent parts (the brief draws them disabled) |
| D-7 | Artifact selection: `selectedArtifactIDs: Set<String>`, seeded with every Artifact id the first time the workspace is ready (`seedIfNeeded()`, called from the dialog's `.task` and `.onChange(of: workspaceHandle.status)`). `artifactIDsToCopy` = workspace Artifacts in workspace order filtered by the selection | CS-6; order matches the page; ids deleted elsewhere drop out (AC-3.2) | Seeding at init (the list host has no workspace yet) |
| D-8 | `func copy() async -> CatalogSource?` validates with D-1 (sets `titleError` / `typeError`, returns nil), sets `isSaving`, calls `store.copySource(... title: draft.trimmedTitle ...)`, then `session.apply(effects, .createdSource)`, then returns the copy. No `setQueryValue`. On throw: `copyError = L10n.Errors.message(for: error)`, returns nil, keeps every choice | Working rule forbids the write-through; `.createdSource` refreshes `.sourcesList` and `.sourceTypesList` | Copying `create()`'s write-through |
| D-9 | After a successful copy the host sets `copying = nil` and calls `navigation.go(to: SourcesListNavigation.pageLocation(for: copy))`. No toast | CS-9: opening the copy is the confirmation; a toast on the page being left is not seen | A toast as Add Source shows |
| D-10 | Stale form (S12-01 UX-1): the checklist reads the live `.sourceWorkspace` handle, so a delete elsewhere reloads it and D-7 drops the id. No extra refresh or error mapping | The session already invalidates the key on Artifact deletes | Mapping `sources.invalid` to a new message |
| D-11 | `SourceArtifactChecklist` *(new)*, `macos/App/Features/Sources/SourceArtifactChecklist.swift`: `artifacts: [CatalogArtifact]`, `@Binding selection: Set<String>`, `projectDir: String`, `sourceTypeIconKey: String?`. An All Artifacts `PVCheckbox` (checked when all selected, indeterminate when some, toggling selects all or none), then one row per Artifact: `CachedThumbnail` size 32 + `PVCheckbox(verbatim: label.isEmpty ? ref : label, isChecked:)`. Rows sit in a `ScrollView` capped at 240pt. Empty `artifacts` → one muted `Text(L10n.Sources.copyNoArtifacts)` | Brief inventory "Artifact checklist" (snowflake, New, promoted by S12-12) | A kit component now (one caller) |
| D-12 | Fileless Artifact icon: the private `artifactTypeIconKey(for:)` body becomes `extension CatalogArtifact { func thumbnailTypeIconKey(sourceTypeIconKey: String?) -> String? }` *(new)* in `SourceArtifactChecklist.swift`; `SourcePageArtifactsView` calls it | One rule for both thumbnails | A second private copy |
| D-13 | `View.sourceCopyDialog(_ model: Binding<SourceCopyModel?>, onCopied: @escaping (CatalogSource) -> Void)` *(new)* in `macos/App/Features/Sources/SourceCopyDialog.swift`: `.pvFormDialog(isPresented: model != nil, width: 520, isRunning:, isDirty:, confirmDisabled: !model.isReady, accessibilityIdentifierPrefix: "sources.copy")`. Form: `copyError` callout (`sources.copy.error`), `SourceTitleTypeFields` (prefix `sources.copy`), section heading, three part `PVCheckbox`es (ids `sources.copy.description` / `.metadata` / `.credibility`), then the checklist (`sources.copy.artifacts`). While the workspace loads: `ProgressView().controlSize(.small)` in the parts area; on load failure: danger `PVCallout` (`sources.copy.loadError`). Type options and icons come from the `.sourceTypesList` handle (`session.query`) | One presenter for both hosts | Two inline dialogs |
| D-14 | `isDirty` = title ≠ the prefill, type ≠ the original's, any `include…` false, or the selection ≠ all Artifacts | CS-10 | Always dirty |
| D-15 | Source page entry: `copySourceButton` *(new)* in `SourcePageIdentityHeader`, `PVIconButton(.duplicate, label: L10n.Sources.copySource, size: .sm)` placed before `deleteSourceButton`, disabled when `model.source == nil`, id `sources.page.copySource`, action `model.openCopy()`. `SourcePageContent` adds `.sourceCopyDialog($model.copying) { navigation.go(...) }` | CS-1 | A text button; a toolbar |
| D-16 | List entry: `SourcesSplitRow` gains `var onCopy: () -> Void` and a `PVContextMenuState` + `PVContextMenuKeyboard` pair; `.pvContextMenuTrigger` on the row, `.pvContextMenu` with `PVContextMenuPanel(title: nil, accessibilityIdentifier: "sources.row.menu")` and one `PVContextMenuItem(L10n.Sources.copySource, index: 0, accessibilityIdentifier: "sources.row.copy")`, keyboard titles `[L10n.string(L10n.Sources.copySource)]`; `.accessibilityAction(named: Text(L10n.Sources.copySource), onCopy)`. Both call sites pass `onCopy: { model.openCopy(source) }`. `SourcesListContent` adds `.sourceCopyDialog($model.copying) { ... }` | CS-2; the cover menu pattern | A native SwiftUI `.contextMenu` (the kit's menu is the rule) |
| D-17 | Tests are Swift Testing model tests over `FakeStore` in a new `macos/ProvenenciaTests/SourceCopyModelTests.swift` *(new)*, with its own `makeSession` / seeding helpers copied from `SourcesModelTests`. The failure case copies an original that is not in `sourcesByProject` (the fake throws `StoreBoom.boom`). No UI tests | The model holds every rule; the fake has no error injection | Adding a throw hook to `FakeStore` |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourcesModel.Draft.trimmedTitle` / `typeError` / `titleError` *(new)* | members | inline checks in `SourcesModel.create()` | They are those checks, moved (D-1) |
| `SourceTitleTypeFields` *(new)* | view | `SourcesListContent.addForm` | Extracted from it; `addForm` uses it (D-2) |
| `SourceCopyModel` *(new)* | type | `SourcesModel` add state | Searched `Draft`, `isAdding`, `create(`: the add state lives on the list model, which the page does not have (D-4) |
| `SourceArtifactChecklist` *(new)* | view | none — searched `PVCheckbox` callers (`PromoteView`, `CitationComposerEvidenceColumn`) and Artifact lists (`SourcePageArtifactsView` cards) | No checklist of Artifacts exists; snowflake per brief (D-11) |
| `CatalogArtifact.thumbnailTypeIconKey(sourceTypeIconKey:)` *(new)* | func | `SourcePageArtifactsView.artifactTypeIconKey(for:)` | Lifted from it; both callers use it (D-12) |
| `View.sourceCopyDialog` *(new)* | modifier | `.pvFormDialog` call in `SourcesListView` | Wraps `pvFormDialog` once for two hosts (D-13) |
| `SourcesModel.copying` / `openCopy`, `SourcePageModel.copying` / `openCopy` *(new)* | members | `isAdding` / `openAdd()` | Same shape (D-5) |
| `SourcesSplitRow.onCopy` + row menu *(new)* | member | `SourcePageIdentityHeader.coverMenu` | Same kit menu (D-16) |
| `L10n.Sources.copy*` *(new)* | keys | `sources.add.*` | New copy |
| `SourceCopyModelTests` *(new)* | test file | `SourcesModelTests` | New model |

**Builds on.** `GenealogyStore.copySource` (S12-01), `session.apply(_:_:)` with `.createdSource` and the registry's invalidations, `PVFormDialog`, `PVCheckbox`, `PVContextMenu`, `SourcesListNavigation.pageLocation(for:)`.
**Deletes.** The inline validation in `SourcesModel.create()` (moved to `Draft`), the title/type block and `sourceTypeComboRow` in `SourcesListView.swift` (moved to `SourceTitleTypeFields`), and the private `artifactTypeIconKey(for:)` body.
**Next adds.** S12-12 promotes `SourceArtifactChecklist` to its second caller (the Graph form). S12-16 removes `create()`'s write-through so Add Source uses the same `apply(effects, .createdSource)` path as `SourceCopyModel.copy()`. More row menu items go in the `SourcesSplitRow` panel.

### Schema

None — no catalog change.

### Go core

None — S12-01 shipped the engine.

### FFI

None — `CopySource` and `GenealogyStore.copySource` exist (S12-01).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Copy Source dialog | `PVFormDialog` 520pt | `Features/Sources/SourceCopyDialog.swift` *(new)* | D-13 |
| Title and type | `PVField` + `PVInput`, `PVField` + `PVComboBox` | `Features/Sources/SourceTitleTypeFields.swift` *(new)* | D-2; Add Source uses it too |
| Part checkboxes | `PVCheckbox` (label, description hint, disabled) | `SourceCopyDialog.swift` | D-6 |
| Artifact checklist | `PVCheckbox` + `CachedThumbnail` | `Features/Sources/SourceArtifactChecklist.swift` *(new)* | D-11 |
| Save / load errors | `PVCallout` danger | `SourceCopyDialog.swift` | D-8, D-13 |
| Page Copy action | `PVIconButton(.duplicate)` | `SourcePageIdentityHeader.swift` | D-15 |
| Row Copy action | `PVContextMenu` | `SourcesSplitRow.swift` | D-16 |

States: loading (list host) → spinner in the parts area, Copy disabled; load error → callout, Copy disabled; no Artifacts → one muted line; 40 Artifacts → rows scroll in 240pt; save error → callout, choices kept. Accessibility: every control has the id above; the row has a named accessibility action; checkboxes carry their labels.

When the board is agreed, replace this paragraph with the frames' layout and copy if they differ from D-11 – D-16 (handoff in [S12-D1](../design/S12-D1-copy-source-form.md) §7).

### Strings

L10n keys *(new)* in `L10n.Sources`, with `Localizable.xcstrings` entries:

- `sources.copy.action` `copySource` = "Copy…"
- `sources.copy.title` `copyDialogTitle` = "Copy Source"
- `sources.copy.subtitle` `copyDialogSubtitle(_ title:)` = "Start a new Source from %@. Notes and the Evidence graph are not copied."
- `sources.copy.confirm` `copyAction` = "Copy"; Cancel reuses `L10n.Sources.cancelAction`
- `sources.copy.titlePrefill` `copyTitlePrefill(_ title:)` = "Copy of %@"
- `sources.copy.partsHeading` `copyPartsHeading` = "Copy from the original"
- `sources.copy.description` = "Description"; `sources.copy.metadata` = "Metadata", `sources.copy.metadataHint` = "Values and field layout"; `sources.copy.credibility` = "Credibility", `sources.copy.credibilityHint` = "Grade and argument"; `sources.copy.partAbsent` = "None on this Source"
- `sources.copy.artifactsHeading` = "Artifacts"; `sources.copy.allArtifacts` = "All Artifacts"; `sources.copy.noArtifacts` = "This Source has no Artifacts."

The board's final copy replaces these values before status `Planned`.

## 7. Steps

**STEP-1 — Share the title and type fields**
Files: `macos/App/Features/Sources/SourcesModel.swift`, `macos/App/Features/Sources/SourcesListView.swift`, `macos/App/Features/Sources/SourceTitleTypeFields.swift` *(new)*. Skill: `add-ui-component`.
Do: D-1, D-2. Add T-1.
Check: `xcodebuild test … -only-testing:ProvenenciaTests/SourcesModelTests` passes unchanged plus T-1; Add Source looks the same.

**STEP-2 — Copy model**
Files: `macos/App/Features/Sources/SourceCopyModel.swift` *(new)*, `SourcesModel.swift`, `SourcePageModel.swift`, `macos/ProvenenciaTests/SourceCopyModelTests.swift` *(new)*. Skill: `add-swift-test`, `add-workspace-place` (session query and apply).
Do: D-3 – D-10, D-14, D-17. Add T-2 – T-9. Add only the strings this step reads (`copyTitlePrefill`) per `add-localized-string`.
Check: `-only-testing:ProvenenciaTests/SourceCopyModelTests` passes.

**STEP-3 — Artifact checklist**
Files: `macos/App/Features/Sources/SourceArtifactChecklist.swift` *(new)*, `macos/App/Features/Sources/SourcePageArtifactsView.swift`. Skill: `add-ui-component`.
Do: D-11, D-12. Add T-10.
Check: build passes; Source page Artifact cards show the same thumbnails.

**STEP-4 — Copy dialog and strings**
Files: `macos/App/Features/Sources/SourceCopyDialog.swift` *(new)*, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`, `add-ui-component`.
Do: D-13 and every key in §6 Strings.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; build passes.

**STEP-5 — Source page entry**
Files: `macos/App/Features/Sources/SourcePageIdentityHeader.swift`, `macos/App/Features/Sources/SourcePageView.swift`.
Do: D-9, D-15.
Check: manual check step 1–4.

**STEP-6 — Sources list entry**
Files: `macos/App/Features/Sources/SourcesSplitRow.swift`, `macos/App/Features/Sources/SourcesListView.swift`.
Do: D-9, D-16.
Check: manual check step 5–7.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status), `docs/deployment-plan/spike-12/design/` (S12-D1 → `docs/deployment-plan/spike-12/design/archive/` *(new)*, `design/README.md`).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-02-copy-source-form.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Copy a Source from its page or the Sources list."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `SourcesModelTests` `draftValidationErrors` *(new)* | `Draft(sourceTypeID: "", title: "  ")` → `typeError` and `titleError` equal today's strings; `"t1"`, `" Album "` → both nil, `trimmedTitle == "Album"` | AC-1.6, D-1 |
| T-2 | Swift | `SourceCopyModelTests` `prefillsTitleAndType` *(new)* | original "Obit", type `t1` → `draft.title == "Copy of Obit"`, `draft.sourceTypeID == "t1"` | AC-1.1, D-4 |
| T-3 | Swift | `seedsAllPartsAndArtifactsWhenLoaded` *(new)* | seeded description, 2 metadata, credibility, 2 Artifacts; wait for the workspace; `seedIfNeeded()` → three parts available and included, `artifactIDsToCopy` = both ids in order | AC-1.1, AC-2.2, D-6, D-7 |
| T-4 | Swift | `copyAllCreatesCopyAndReturnsIt` *(new)* | T-3 fixture; `copy()` → returns a Source titled "Copy of Obit"; `store.sourcesByProject` has 2; the copy's `metadataBySource` count 2, `credibilityBySource` set, `artifactsBySource` labels equal; original's entries unchanged | AC-1.2, D-8 |
| T-5 | Swift | `uncheckedPartsAreNotCopied` *(new)* | T-3 fixture; `includeCredibility = false`, remove the second id → copy has no credibility, one Artifact with the first label | AC-1.3 |
| T-6 | Swift | `absentPartsAreUnavailableAndNotSent` *(new)* | empty description, no credibility, no Artifacts → `descriptionAvailable == false`, `credibilityAvailable == false`, `artifactIDsToCopy == []`; copy has `description == ""`, no credibility | AC-1.4, AC-1.5, D-6 |
| T-7 | Swift | `copyFailureKeepsChoices` *(new)* | original not in `sourcesByProject` → `copy()` returns nil, `copyError != nil`, `draft` and `include…` flags unchanged | AC-3.1, D-8, D-17 |
| T-8 | Swift | `copyRefreshesSourcesListThroughSession` *(new)* | warm `.sourcesList` via `session.query`, `copy()`, wait → the list handle's value contains the copy's id; no `setQueryValue` call path (the list reloads from the fake) | AC-2.3, D-8 |
| T-9 | Swift | `validationBlocksCopy`, `isDirtyTracksChanges` *(new)* | title `" "` → nil, `titleError` set, store has 1 Source; untouched → `isDirty == false`; one checkbox off → `true` | AC-1.6, AC-3.3, D-14 |
| T-10 | Swift | `SourceCopyModelTests` `artifactDeletedElsewhereIsDropped` *(new)* | T-3 fixture; remove Artifact 2 from `store.artifactsBySource`, `session.invalidate` the workspace key, wait → `artifactIDsToCopy` = Artifact 1 only | AC-3.2, D-7, D-10 |
| T-11 | Swift | `SourcesModelTests` `openCopySetsCopyingForThatSource`, `SourcePageModelTests` `openCopyUsesThePageSource` *(new)* | `openCopy(source)` → `copying?.draft.title == "Copy of …"` | D-5 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcesModelTests` (Add Source validation moved), `SourcePageModelTests` (new model member), the Xcode build of `SourcePageArtifactsView` (D-12).

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
| Extracting the fields changes Add Source's ids or behavior | D-2 keeps the `sources.add` prefix; `SourcesModelTests` must stay green unchanged |
| The row context menu catches left clicks | `pvContextMenuTrigger` only catches right / Ctrl-click (the cover menu works this way); manual check step 5 |
| A 40-Artifact Source makes the dialog taller than the window | D-11 caps the rows at 240pt |
| The fake drops unknown Artifact ids, so a test cannot prove an id is not sent | T-10 asserts `artifactIDsToCopy`, not the fake's result |

**Stop and report if:**

- S12-01 lands with a `copySource` signature or effects different from §4 (check its **Landed** note).
- The agreed S12-D1 board differs from D-11 – D-16 or §6 Strings (another layout, a toast, more menu items). The plan is updated before code.
- `.createdSource` no longer invalidates `.sourcesList` in `CatalogQueryRegistry.swift`, so T-8 cannot pass without a write-through.
- `SourceTitleTypeFields` cannot keep `sources.add.title` / `sources.add.type` without changing `SourcesListContent` beyond `addForm`.
- `pvContextMenuTrigger` on `SourcesSplitRow` swallows the page or graph button's left click.

## 12. Open questions

1. **How does the agreed S12-D1 board reach the implementer?** Recommendation: export the frames as PNGs into `docs/deployment-plan/spike-12/design/S12-D1/` *(new)* and link them on the **Board** line.

## 13. PR

**Title:** S12-02 — Copy Source form
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-02-copy-source-form.md`.
