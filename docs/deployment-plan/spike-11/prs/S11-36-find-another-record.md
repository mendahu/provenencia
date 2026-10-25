# S11-36 — Find another record in the target menu

**Status:** Blocked — on board S11-D3
**Spike / slice:** Spike 11 · Slice 8 — Find another record
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S11-D3](../design/S11-D3-promote-target-menu.md), board not agreed yet
**Board:** [Claude Design — Promote flow](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share), S11-D3 frames (not drawn yet). Until they are, §6 macOS describes each frame in words from the brief
**Depends on:** S11-35 (landed [#543](https://github.com/mendahu/provenencia/pull/543)), S11-37 (landed [#551](https://github.com/mendahu/provenencia/pull/551)), S11-D3
**Blocks:** —
**Branch:** `s11-36-find-another-record`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [Find another record § Shape](../promote-find-another-record.md#shape), [§ Where the Swift side changes](../promote-find-another-record.md#where-the-swift-side-changes), [§ Decided](../promote-find-another-record.md#decided-user-2026-10-24); working rules "A found record is a decided row, nothing new", "No new search RPC, and one search model", "An inline picker, every record of the kind, no warning", "The menu entry is a sentinel option", "The picked record joins the row's menu"; the map's S11-36 scope entry (picker on `PVComboBox`, the hit row's VoiceOver label, `selected`)
**Planned:** 2026-10-25 against `main` @ `056b6ed`
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

A researcher on the Promote page who knows a Subject is a record Promote did not offer can find that record from the row and file the Subject on it. Every working row's target menu ends with "Find another record…". Choosing it turns the row's target cell into a search field over every record of the row's kind; picking a hit holds it as the row's decided target through S11-35's `hold`, so the row shows the record's own band, Contested and evidence, and Done files it. A chosen record the evidence rates below Weak reads "Little evidence" instead of "No match".

**Why now.** Since S11-05 a record whose first given name or surname differs from the Subject's is never a candidate, and the only route left is New and a merge the app does not have ([the problem](../promote-find-another-record.md#the-problem)). S11-35 built the call and S11-37 the shared search; this PR is the surface that uses them. It makes the spike's dogfood bar 11 true.

**What stays.** Go, the FFI and Promote's proposals are unchanged. The menu's other items and their order, how a record item is drawn, the band, Contested, Decided and Updated Badges, the row's notes and subrows, the evidence sheet's content, and the toolbar omnibar's look and behavior.

## 2. User stories

**US-1.** As a researcher, when Promote offers nothing (or the wrong records) for a Subject, I choose "Find another record…" and search the row's kind, so that I can reach a differently named record.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A person row Promote left unset (menu: New Person, Skip); open its menu | The last item is "Find another record…" with a search icon (FR-1) |
| AC-1.2 | Choose "Find another record…" on that row | The target cell shows a focused search field "Search people" and a Cancel button; the row is still not Decided; no proposal is sent (FR-2) |
| AC-1.3 | Type "Polly" in a catalog holding person PER-FAR "Polly Jones" and place PLC-1 "Polly Street" | The list shows PER-FAR only, drawn as the omnibar's hit row (FR-3) |
| AC-1.4 | Type "P" | The list line reads "Type at least two letters of a name or a ref"; no search runs (FR-4) |
| AC-1.5 | Type "Pollx" | The list line reads "No people match “Pollx”" (FR-4) |
| AC-1.6 | The store's search fails | The list line shows the failure's message (FR-4) |

**US-2.** As a researcher, when I pick a hit, it becomes the row's target like a menu pick, so that I see its evidence and Done files it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Pick PER-FAR from the list | The search closes; the menu shows "Jones, Polly" selected; the row is Decided; the page proposes again with `{subjectID: a, handleID: e-far}` fixed (FR-5) |
| AC-2.2 | Open the menu after the pick | PER-FAR is an item after the offered records and before New Person, checked (FR-5) |
| AC-2.3 | Pick a hit already in the row's menu | That item is selected; the menu has no second copy (FR-5) |
| AC-2.4 | Press Done | The batch row for the Subject files on `e-far` |

**US-3.** As a researcher, opening and cancelling the search changes nothing, so that I can look without deciding.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Open the search on a row, then press Cancel search | The menu comes back with the same target; the row is not Decided; no proposal was sent (FR-6) |
| AC-3.2 | Open the search on row A, then choose "Find another record…" on row B | Row A shows its menu again, unchanged; row B shows the search (FR-6) |
| AC-3.3 | In the search, press Escape until the list is closed and the field is empty, then press Escape again | The search closes as in AC-3.1 (FR-6) |

**US-4.** As a researcher, a record I chose that the evidence rates below Weak does not read as "No match", so that the row does not contradict my choice.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A row holds a found record and Go answers band None, not contested | The assessment reads "Little evidence · Your choice", on the row and on the sheet title (FR-7) |
| AC-4.2 | The same, contested | "Little evidence", then Contested, then "Your choice" (FR-7) |
| AC-4.3 | A row whose offered record Promote rates None (reason not `decided`) | The band still reads "No match" (FR-7) |
| AC-4.4 | VoiceOver on the AC-4.1 row's evidence button | "Evidence for Ada: Little evidence, Your choice" (FR-8) |

**US-5.** As a VoiceOver user, the entry, the field, each hit and Cancel are read in words.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | VoiceOver on the entry, the field, a hit, the Cancel button | "Find another record…"; "Find another record for Ada"; the hit's `CatalogSearchHitPresentation.accessibilityLabel(for:)` ("Jones, Polly, Person, PER-FAR"); "Cancel search" (FR-8) |
| AC-5.2 | VoiceOver on an omnibar hit after this PR | Reads as before this PR: title, kind, subtitle, ref, match context |

## 3. Scope

| In | Covered by |
| --- | --- |
| One sentinel token and one handle-token spelling on `PromoteFlow.Target`; `PromoteFlow.setTarget` ignores the sentinel (D-1, D-2) | STEP-1, AC-3.1 |
| `PromoteModel`'s one open search: open by the sentinel, query, pick through `hold`, close (D-3 – D-6) | STEP-2, AC-1.2, AC-1.3, AC-2.1, AC-3.1, AC-3.2 |
| "Little evidence" for a chosen record at band None, on the row, the sheet title and the spoken label (FR-7, D-9) | STEP-3, AC-4.1 – AC-4.4 |
| `CatalogSearchHitRow` carries the hit's VoiceOver label and loses `selected` (D-10, D-11) | STEP-4, AC-5.1, AC-5.2 |
| The "Find another record…" menu entry (FR-1) | STEP-5, AC-1.1 |
| The search in the target cell: `PVComboBox` with remote results, its list lines, Cancel, Escape (FR-2 – FR-6, FR-8, FR-9; D-7, D-8, D-12) | STEP-6, AC-1.2 – AC-1.6, AC-2.1 – AC-2.3, AC-3.3, AC-5.1 |
| Docs: feature doc, map, completed, brief archive | STEP-7 |

| Out | Home |
| --- | --- |
| Searching from the evidence sheet; comparing two records before picking; records of another kind; merging records | not planned in this spike ([what stays later](../promote-find-another-record.md#what-stays-later)) |
| A Callout or mark for a `Different` pick | not planned (refused, user 2026-10-24) |
| A band, score or Contested mark per hit in the list | not planned (S11-D3 §2.1: the search does not compare) |
| An action item, divider or section header in `PVSelect` | not planned (working rule "The menu entry is a sentinel option") |
| Hiding the kind Badge in `PVOmnibarHitRow` for a one-kind list | not planned (S11-D3 non-goals: no change to the hit row recipe) |
| A shared popup presenter for the omnibar's Escape and outside-click monitor | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) |
| Removing `PVOmnibarHitRow`'s own `selected` parameter | not planned: it is recipe API and its previews draw it |
| Any change to `SearchCatalog`, its ranking, or Go | not planned (working rule "No new search RPC, and one search model") |
| Closing the search when the field loses focus | not planned (D-8: the search stays until a pick or a cancel, as the Source page's type editor does) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Target tokens | `macos/App/Features/Promote/PromoteFlow.swift` `PromoteFlow.Target.token` (line 16), `target(for:on:)` (line 638) | `"handle:\(id)"`, `"new"`, `"skip"`, `""`. `target(for:on:)` parses the `"handle:"` prefix; an unknown token returns the row's current target |
| setTarget | `PromoteFlow.setTarget(subjectID:token:)` (line 432) → `select(_:at:)` (line 462) | Any token on a working row marks the row decided, sets `manual`, clears `updated` and returns true, even when the target does not change. A sentinel reaching it would decide the row |
| hold | `PromoteFlow.hold(subjectID:id:ref:header:)` (line 442); `PromoteModel.hold(subjectID:id:ref:header:)` (line 179) | Appends an unoffered record to `row.menu` with reason `decided` and band None, or reuses the offered item, then selects it; the model re-proposes. No caller outside tests |
| Model target call | `macos/App/Features/Promote/PromoteModel.swift` `setTarget(subjectID:token:)` (line 173) | Forwards to the flow; re-proposes when it returns true |
| Model state | `PromoteModel` `private(set) var flow`, `session`, `store`, `isSaving` (lines 93 – 126) | `@Observable @MainActor` class; `session.projectKey.projectDir` and `session.nameFormat` are what search calls need |
| Target cell | `macos/App/Features/Promote/PromoteView.swift` `PromoteAlignmentRow` (line 309): `PVSelect` (line 331), `targetOptions` (line 427), `targetBinding` (line 436) | `PVSelect` of `row.menu` items (`"handle:\(alt.id)"`), New, Skip; the binding's setter calls `model.setTarget`. The cell is `PromoteLayout.targetColumn` (330 pt) wide. The whole row is `.disabled(model.isSaving)` in `sectionBlock` |
| Menu rows | `PromoteView.swift` `PromoteTargetMenuRow` (line 505) | Plain text for `new` / `skip`; mark, title and "ref · subtitle" for a `row.menu` item matched by `"handle:\($0.id)"`; draws nothing for any other option id |
| PVSelect state | `macos/App/DesignSystem/Components/Select/PVSelect.swift` `PVSelect` `session` (`PVSelectSession`), `.onChange(of: selection)` (line 172), `.onChange(of: session.selection)` (line 183) | Keeps its own `session.selection`. It writes a committed option to the binding, and copies the binding back only when the binding's value changes. A setter that ignores a token leaves `session.selection` on that token, so the closed field would show the sentinel's label |
| Readout | `macos/App/Features/Promote/PromoteReadout.swift` `PromoteReadoutBadges` | Band Badge text `L10n.Promote.assessment(readout.assessment)`; None reads "No match" (`promote.assessment.none`). Drawn on the row (`PromoteView.swift` line 354) and the sheet title (`PromoteEvidenceSheet.swift` line 90) |
| Spoken band | `PromoteModel.assessmentText(_:)` (line 318) | `L10n.Promote.assessment(readout.assessment)`, plus ", contested" |
| Search model | `macos/App/Features/Catalog/Search/CatalogSearchModel.swift` `CatalogSearchModel` | `init(kinds:)`, `query`, `hits`, `isLoading`, `searchError`, `trimmedQuery`, `minQueryLength` (2), `scheduleSearch(projectDir:location:store:)` (a query under two characters calls `dismiss()`), `cancel()` |
| Hit row | `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift` `CatalogSearchHitRow` | `hit`, `projectDir`, `selected` (forwarded to `PVOmnibarHitRow`), `@Environment(WorkspaceSession.self)` for `nameFormat`. No accessibility label: `PVOmnibarHitRow` combines its children, which reads in layout order |
| Omnibar call site | `macos/App/Features/Workspace/OmnibarResultsHost.swift` `OmnibarResultsPanel` (line 92) | Passes `selected: false` (line 117; `PVMenuRowButtonStyle` draws selection) and sets `.accessibilityLabel(CatalogSearchHitPresentation.accessibilityLabel(for: hit))` on the `Button` |
| Combo box | `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift` `PVComboBox`, the `Empty == Text` and general inits | Remote results: `onQueryChange` stops local filtering; `isLoading` shows the searching line (`PVComboBoxMatch.listLine`, #554); `empty(query)` is the caller's line; `onCommit(value, via)`; `focus: PVComboFocus`; Escape two-stage then `.ignored`; outside click closes the list and blurs. No feature caller uses `onQueryChange` yet |
| Popup environment | `PVComboBox.swift` `PVComboBoxPopupWindow` (line 818), `PVComboBoxHostingView` (line 1082) | The list is an `NSHostingView` in a child window. Environment values set above the box do not reach it; only modifiers inside the `row` view value do |
| Precedent | `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `metaRow` (line 185) | Swaps an inline `PVComboBox` (`focus: PVComboFocus($typeFieldFocused, equals: .combo)`, `.onAppear { DispatchQueue.main.async { typeFieldFocused = .combo } }`) and a `PVIconButton(.dismiss, …)` cancel in place of the type pill while `model.identity.editingType` is true |
| Strings | `macos/App/Platform/L10n.swift` `enum Promote` (line 1305), `newOption(_:)` (line 1708); `macos/App/Resources/Localizable.xcstrings` | Kind-switched copy follows `newOption`'s shape |
| Tests | `macos/ProvenenciaTests/PromoteFlowTests.swift` (`holdingAFoundRecordResetsTheOldTargetsEvidence`, `anAnchorOrMissingRowOrEmptyIDHoldsNothing`); `macos/ProvenenciaTests/PromoteModelTests.swift` (`aFoundRecordIsProposedHeldAndFiled`, `store(subjects:)`, `model(store:subjectID:)`); `macos/ProvenenciaTests/CatalogSearchTests.swift`; `macos/ProvenenciaTests/QueryHandleWaiting.swift` `waitUntil`, `waitForSearch` | `FakeStore.searchCatalog` searches `membershipBySubject` entities by kind; `searchCatalogError` makes it throw |

**Precedent.** The Source page's type editor (`SourcePageIdentityHeader.metaRow`, `3e9a0712` "Sync the Source page to explicit edit flows and DateValue editing.", reshaped in `2823c737`) swaps a cell for an inline `PVComboBox` plus a dismiss `PVIconButton`, focuses it on appear, and keeps the editing flag on the model. Copy the swap, the focus-on-appear, and the model-owned flag. Do not copy its `onChange(of: typeDraftID)` save: this picker commits through `onCommit`. S11-35 (#543) added `hold` and S11-37 (#551) the search model; this PR calls both unchanged.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PromoteFlow.Target` gains `static let findToken = "find"` *(new)* and `static func handleToken(_ id: String) -> String` *(new)* returning `"handle:\(id)"`. `Target.token`, `target(for:on:)` (prefix parse through `handleToken("")`), `PromoteView.targetOptions` and `PromoteTargetMenuRow` all use them; no other `"handle:"` literal stays in `macos/App` | One spelling of each token, the one S11-35 left to this PR (S11-35 Out line); the menu and the parser cannot drift | A token enum — because `PVSelect` values are strings and the four call sites are in two files |
| D-2 | `PromoteFlow.setTarget` returns false and changes nothing when `token == Target.findToken` | The rule "opening or cancelling the search never marks the row decided" holds in the type that owns decisions, whatever the caller | Relying on callers alone — because `select` decides on any token |
| D-3 | `PromoteModel.setTarget(subjectID:token:)` sends `Target.findToken` to `openFind(subjectID:)` *(new)* and returns before calling the flow. `PromoteView`'s binding keeps calling `model.setTarget` | This is the "row's binding catches the token before `setTarget`" of the working rule, in a place a model test drives | Catching it in the view's binding setter — because no test can drive a view binding; a separate model call from the binding — because the binding would branch on a token the model already knows |
| D-4 | `PromoteModel` holds one open search: `struct Find { let subjectID: String; let search: CatalogSearchModel }` *(new, nested)* and `private(set) var find: Find?` *(new)*. `openFind(subjectID:)` returns without change for an anchor or a missing row; otherwise it cancels any open search and sets `find` with `CatalogSearchModel(kinds: [row.kind.rawValue])` | One search at a time (FR-6), kind from the row (working rule "every record of the kind"); one optional value cannot leave a subject id without its search | Per-row `@State` in the view — because opening on another row must close the first and tests cannot reach view state; two optionals — because they can disagree |
| D-5 | `findQueryChanged(_ text: String)` *(new)* sets `find.search.query = text` and calls `find.search.scheduleSearch(projectDir: session.projectKey.projectDir, location: nil, store: store)` | The S11-37 model owns debounce, generation, the two-character floor and errors; Promote has no location (S11-37 D-3) | Calling `store.searchCatalog` from Promote — because it rebuilds the debounce |
| D-6 | `pickFound(hitID: String)` *(new)* finds the hit in `find.search.hits`, calls `closeFind()`, then `hold(subjectID: find.subjectID, id: hit.id, ref: hit.ref, header: hit.header)`. An id not in `hits` does nothing. `closeFind()` *(new)* calls `find?.search.cancel()` and sets `find = nil` | The pick goes through S11-35's call, which titles the item and handles a hit already in the menu (AC-2.3); closing first means the menu is back when the re-proposal lands | Passing a `CatalogSearchHit` — because the combo box commits a value string, and a test then needs a hand-built hit |
| D-7 | While `model.find?.subjectID == row.subjectID`, the target cell shows `PromoteRecordFinder` *(new, private, in `PromoteView.swift`)* instead of the `PVSelect`, in an `if` / `else` so the select leaves the hierarchy | Dropping the `PVSelect` drops its `session.selection` copy (Current state "PVSelect state"); a fresh select is built from the binding when the search closes, so the sentinel never shows as the selection. Same swap as the Source type editor | Keeping the select mounted with the picker beside it — because the select would keep showing the sentinel's label; a kit change to `PVSelect` — because the working rule gives it no action item |
| D-8 | `PromoteRecordFinder` is an `HStack` of `PVComboBox` (`size: .sm`, remote results) and `PVIconButton(.dismiss, label: L10n.Promote.findCancel, size: .sm)`, filling `PromoteLayout.targetColumn`. Focus: `@FocusState private var focus: Field?` with `enum Field { case query }`, passed as `PVComboFocus($focus, equals: .query)`, set in `.onAppear { DispatchQueue.main.async { focus = .query } }`. It closes on Cancel, on a pick, and on `.onExitCommand { model.closeFind() }`, which reaches it only when the combo box returns `.ignored` for Escape (nothing left to close or clear). Losing focus does not close it | Reuses the box's own Escape and outside click (map scope: no new hand-rolled monitor) and the Source page's cancel; `onExitCommand` is the standard Escape route, not an event monitor | Closing on focus loss — because a click on the scroll view or the assessment column would throw the search away mid-look, and the Source editor does not; a `NSEvent` monitor — refused by the map scope |
| D-9 | `L10n.Promote.band(_ readout: PromoteFlow.SelectionReadout) -> LocalizedStringResource` *(new)* returns `assessmentChosenNone` *(new)* when `readout.assessment == .none && readout.reason == .decided`, else `assessment(readout.assessment)`. `PromoteReadoutBadges` and `PromoteModel.assessmentText` call it. The tone stays `.neutral` | One place words the band for the row, the sheet and VoiceOver. `reason == .decided` is exactly a record the researcher chose that Promote never offered for that row (a menu pick keeps its offered reason, [why](../promote-find-another-record.md#where-the-swift-side-changes)) | Testing "found" with a new flag on `Alternative` — because the reason already says it (working rule "nothing new"); changing `assessmentNone`'s copy for every row — because an offered None is still no match (FR-7) |
| D-10 | `CatalogSearchHitRow` sets `.accessibilityLabel(CatalogSearchHitPresentation.accessibilityLabel(for: hit))` on its `PVOmnibarHitRow`. `OmnibarResultsPanel` drops its own `.accessibilityLabel` on the `Button` | Both callers get the label from the one row; the combo box's row shell has no label of its own (S11-37 ship log UX-1; the map lets the plan move it) | Applying it at the picker's call site too — because two call sites would each have to remember it |
| D-11 | `CatalogSearchHitRow` loses `selected`; both callers stop passing it, and `PVOmnibarHitRow` gets its default `false` | Both callers draw selection outside the row (`PVMenuRowButtonStyle`, the combo box's row shell) (S11-37 ship log MNT-3) | Keeping it — because a parameter no caller sets is dead |
| D-12 | The combo box's options are `find.search.hits.map { PVComboBoxOption(value: $0.id, label: CatalogSearchHitPresentation.title(for: $0, orders: model.session.nameFormat)) }`; `selection` is `.constant("")`; `isLoading: find.search.isLoading`; the row builder draws `CatalogSearchHitRow(hit:projectDir:)` for the option's hit with `.environment(model.session)` applied inside the builder; `onCommit` calls `model.pickFound(hitID:)` and returns false | The list lives in its own `NSHostingView` (Current state "Popup environment"), so the hit row's `@Environment(WorkspaceSession.self)` must travel in the row value. The box mirrors the label into the field only for a committed option; the search closes on commit, so nothing is mirrored | Passing `nameFormat` into `CatalogSearchHitRow` — because S11-37 D-8 kept the environment read and the omnibar call site would change for nothing |
| D-13 | The combo box's `empty` line, in this order: `find.search.searchError` (verbatim); `find.search.trimmedQuery.count < CatalogSearchModel.minQueryLength` → `L10n.Promote.findHint`; else `L10n.Promote.findNoMatches(row.kind, query: find.search.trimmedQuery)` | FR-4's lines; the searching line is the box's own while `isLoading` | Reusing the omnibar's `omnibarNoMatchesTitle` — because FR-4 words it by kind |
| D-14 | Accessibility identifiers: the entry option `promote.row.<ref>.target.find`; the combo box prefix `promote.row.<ref>.find` (the box adds `.list`, `.option.<id>`, `.loading`, `.empty`); Cancel `promote.row.<ref>.find.cancel` | Call-site ids under the row's existing `promote.row.<ref>.` prefix (`add-ui-component`) | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PromoteFlow.Target.findToken`, `handleToken(_:)` *(new)* | constants | the `"handle:"` literals in `PromoteFlow.swift` and `PromoteView.swift` | They replace those literals (D-1) |
| `PromoteModel.Find`, `find`, `openFind`, `findQueryChanged`, `pickFound`, `closeFind` *(new)* | type / funcs | `macos/App/Features/Sources/SourceIdentitySection.swift` `SourceIdentitySection.editingType`, `beginEditType`, `cancelEditType` (searched `beginEdit`, `editing`) | Same shape on this page's model; the Source editor's state is that page's |
| `PromoteRecordFinder` *(new, private)* | view | `SourcePageIdentityHeader.metaRow`'s combo + cancel; `OmnibarResultsPanel` (searched `PVComboBox(`, `CatalogSearchHitRow(`) | One call site, a feature snowflake that composes `PVComboBox`, `CatalogSearchHitRow` and `PVIconButton` (S11-D3 inventory). The omnibar's panel is an overlay with its own monitor (Out) |
| `L10n.Promote.band(_:)` *(new)* | func | `L10n.Promote.assessment(_:)` | Wraps it with the one chosen-None case |
| `L10n.Promote.findOption`, `findPlaceholder(_:)`, `findLabel(name:)`, `findCancel`, `findHint`, `findNoMatches(_:query:)`, `assessmentChosenNone` *(new)* | keys | `L10n.Promote.newOption(_:)` (kind switch); `L10n.Workspace.omnibarNoMatchesTitle` | Promote copy lives in `L10n.Promote`; the omnibar's line is not by kind |

**Builds on.** S11-35's `PromoteFlow.hold` / `PromoteModel.hold`; S11-37's `CatalogSearchModel` and `CatalogSearchHitRow`; `PVComboBox` remote results and its searching line (#554); the Source type editor's swap.
**Deletes.** Every `"handle:"` literal outside `PromoteFlow.Target` (D-1); `CatalogSearchHitRow.selected` and both `selected: false` call-site arguments (D-11); the `.accessibilityLabel` on `OmnibarResultsPanel`'s hit `Button` (D-10).
**Next adds.** A later "pick a record" surface (merge, link a record from the evidence sheet) composes `PVComboBox` + `CatalogSearchModel` + `CatalogSearchHitRow` as `PromoteRecordFinder` does, and gets the VoiceOver label free (D-10). When a second one lands, `PromoteRecordFinder`'s body is the candidate to promote to a recipe.

### Schema

None — no catalog change.

### Go core

None — Go already holds, scores, explains and files a found record (S11-35).

### FFI

None — `SearchCatalog` with `kinds` and `ProposePromoteGraphAlignment` with fixed pairs are used as they are.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Target cell | `PVSelect` or `PromoteRecordFinder` | `macos/App/Features/Promote/PromoteView.swift` `PromoteAlignmentRow` | `if model.find?.subjectID == row.subjectID { PromoteRecordFinder(…) } else { PVSelect(…) }`, both `.frame(width: PromoteLayout.targetColumn)` (D-7) |
| Menu entry (FR-1) | `PVSelectOption` | `targetOptions` | Appended after Skip: `PVSelectOption(value: PromoteFlow.Target.findToken, label: L10n.string(L10n.Promote.findOption), accessibilityIdentifier: "promote.row.\(row.ref).target.find")` |
| Entry row | `PVIcon(.search, size: 14)` + text | `PromoteTargetMenuRow` | New branch for `option.id == PromoteFlow.Target.findToken`: `HStack(spacing: PVSpacing.space3)`, icon and `Text(verbatim: option.label)` in `PVFont.body(size: PVTypeScale.bodySmall)`, both `PVColor.textSecondary` |
| Search field + list | `PVComboBox` (`Empty` general init with `empty:` builder) | `PromoteRecordFinder` | D-8, D-12, D-13. `placeholder: L10n.Promote.findPlaceholder(row.kind)`, `label: L10n.Promote.findLabel(name: row.name)`, `onQueryChange: model.findQueryChanged` |
| Hit row | `CatalogSearchHitRow` → `PVOmnibarHitRow` | `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift` | D-10, D-11, D-12 |
| Cancel | `PVIconButton(.dismiss, label: L10n.Promote.findCancel, size: .sm)` | `PromoteRecordFinder` | `model.closeFind()` |
| Band wording (FR-7) | `PVBadge` | `macos/App/Features/Promote/PromoteReadout.swift` `PromoteReadoutBadges` | `L10n.Promote.band(readout)`; tone unchanged (D-9) |

Frames in words, until the board is agreed (S11-D3 §5): (1) the menu ends with the search icon and "Find another record…" after Skip; (2) the target cell is a small search field "Search people" with a × button after it, the list open on the hint line, the assessment column as before; (3) four person hit rows, the first highlighted; (4) the no-match and failure lines; (5) after a pick, the record selected, Decided, its band, Contested, "Your choice"; (6) "Little evidence · Your choice" on the row and the sheet title.

States: empty field and one character → hint line; searching with no hits → the box's searching line; no hits → no-match line; failure → the error message; filing → the row is already `.disabled(model.isSaving)`, which disables the entry, the field and Cancel (FR-9). Accessibility: D-10, D-14, the strings below.

### Strings

L10n keys *(new)*, in `L10n.Promote` and `Localizable.xcstrings`:

- `findOption` — `promote.target.find` = "Find another record…"
- `findPlaceholder(_ kind:) -> LocalizedStringResource` — `promote.find.placeholder.person` = "Search people", `.event` = "Search events", `.place` = "Search places"
- `findLabel(name:) -> String` — `promote.find.label` = "Find another record for %@"
- `findCancel` — `promote.find.cancel` = "Cancel search"
- `findHint` — `promote.find.hint` = "Type at least two letters of a name or a ref"
- `findNoMatches(_ kind:, query:) -> String` — `promote.find.noMatches.person` = "No people match “%@”", `.event` = "No events match “%@”", `.place` = "No places match “%@”"
- `assessmentChosenNone` — `promote.assessment.chosenNone` = "Little evidence"

Removed: none. The board's copy wins over these values where it differs; the refresh at board agreement updates this list.

## 7. Steps

**STEP-1 — One spelling for target tokens; the sentinel decides nothing**
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `macos/App/Features/Promote/PromoteView.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2. Add T-1 and T-2.
Check: `rg '"handle:' macos/App` prints only `PromoteFlow.swift`'s `handleToken`; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PromoteFlowTests` passes.

**STEP-2 — The model's one open search**
Files: `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`.
Do: D-3 – D-6. Add T-3 – T-6.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PromoteModelTests` passes.

**STEP-3 — "Little evidence" for a chosen record at band None**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Promote/PromoteReadout.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`. Skill: `add-localized-string`.
Do: D-9; add `assessmentChosenNone`. Add T-7.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PromoteModelTests` passes.

**STEP-4 — The hit row carries its VoiceOver label and loses `selected`**
Files: `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift`, `macos/App/Features/Workspace/OmnibarResultsHost.swift`. Skill: `add-ui-component`.
Do: D-10, D-11.
Check: the app builds; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CatalogSearchTests` passes; VoiceOver on an omnibar hit reads as before (AC-5.2, manual check 6).

**STEP-5 — The menu entry**
Files: `macos/App/Features/Promote/PromoteView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: the "Menu entry" and "Entry row" rows of §6 macOS; add `findOption`.
Check: in the app, every working row's menu ends with the entry (AC-1.1); choosing it does nothing visible yet and the row stays undecided (STEP-6 draws the search).

**STEP-6 — The search in the target cell**
Files: `macos/App/Features/Promote/PromoteView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skills: `add-ui-component`, `add-localized-string`.
Do: D-7, D-8, D-12 – D-14; `PromoteRecordFinder`; the remaining keys under §6 Strings.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`; manual check 1 – 5.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row; design track row), this plan (status, Landed note), `docs/deployment-plan/spike-11/design/S11-D3-promote-target-menu.md` → `docs/deployment-plan/spike-11/design/archive/`, `docs/deployment-plan/spike-11/design/README.md`, `docs/deployment-plan/spike-11/promote-find-another-record.md`.
Do: per the Definition of done. In the feature doc, rewrite "The menu entry:" under § Where the Swift side changes in the past tense: the sentinel `PromoteFlow.Target.findToken`, `PromoteModel.find`, `PromoteRecordFinder`; add the "Little evidence" wording to § Shape's "Held, not moved".
Check: `python3 scripts/check-plan-docs.py` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `theFindTokenDecidesNothing` *(new)* | An unset row and a row on `handle:e1`; `setTarget(token: Target.findToken)` on each → returns false; targets, `decided`, `updated` and `manual` unchanged | AC-3.1, D-2 |
| T-2 | Swift | `PromoteFlowTests.swift` `handleTokensAreSpelledOnce` *(new)* | `Target.handle(id: "e1", ref: "PER-1", title: "A").token == Target.handleToken("e1")`; `setTarget(token: Target.handleToken("e1"))` on a row whose menu has `e1` → target `e1` | D-1 |
| T-3 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `findingOpensASearchOfTheRowsKindAndDecidesNothing` *(new)* | Loaded page, row `a` unset; record `store.lastPromoteFixed`; `setTarget(subjectID: "a", token: Target.findToken)` → `find?.subjectID == "a"`, `find?.search.kinds == ["person"]`, row `a` not decided, target unset, `isProposing` false, `lastPromoteFixed` unchanged | AC-1.2, D-3, D-4 |
| T-4 | Swift | `PromoteModelTests.swift` `cancellingOrMovingAFindChangesNoRow` *(new)* | Ada and Bea; open find on `a`, open find on `b` → `find?.subjectID == "b"`; `closeFind()` → `find == nil`; both rows' targets and `decided` unchanged; `lastPromoteFixed` unchanged | AC-3.1, AC-3.2, D-4, D-6 |
| T-5 | Swift | `PromoteModelTests.swift` `aFoundRecordIsSearchedPickedAndHeld` *(new)* | `membershipBySubject` holds person `e-far` "Polly Jones" (PER-FAR) and place `p-1` "Polly Street" (PLC-1, kind `place`); open find on `a`; `findQueryChanged("Polly")`; `waitForSearch(find.search)` → hits `["e-far"]`; `pickFound(hitID: "e-far")` → `find == nil`; after `waitUntil { !model.isProposing }`, `lastPromoteFixed == [{a, handleID: e-far}]`, row `a` decided on `e-far`; `done()` files `e-far` | AC-1.3, AC-2.1, AC-2.4, D-5, D-6 |
| T-6 | Swift | `PromoteModelTests.swift` `aFindShowsFailuresAndIgnoresUnknownHits` *(new)* | `store.searchCatalogError` set; open find, `findQueryChanged("Polly")`, `waitForSearch` → `find?.search.searchError != nil`; `pickFound(hitID: "nope")` → `find` still open, row `a` unchanged | AC-1.6, D-6 |
| T-7 | Swift | `PromoteModelTests.swift` `aChosenRecordBelowWeakReadsLittleEvidence` *(new)* | `L10n.Promote.band` of (`.none`, `.decided`) → `assessmentChosenNone`; (`.none`, `.noMatch`) → `assessmentNone`; (`.weak`, `.decided`) → `assessmentWeak`; `assessmentText` of (`.none`, `.decided`, contested) → "Little evidence, contested" | AC-4.1, AC-4.2, AC-4.3, AC-4.4, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteFlowTests` (token literals), `PromoteModelTests` (`assessmentText` expectations), `CatalogSearchTests`, and `scripts/check-localizable-xcstrings.py`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: promote-find-another-record.md § Where the Swift side changes and § Shape (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a project with a Source that has a person Subject "Ada" Promote offers nothing for, and a filed person record with a different first name (create "Polly Jones" from another Source and promote it New), plus a filed place whose name contains "Polly".

1. Open Promote on Ada. Her row is unset. Open the target menu: the last item is "Find another record…" with a search icon (AC-1.1). Choose it: a focused "Search people" field and × replace the menu; the row has no Decided Badge (AC-1.2).
2. Type "P": the hint line (AC-1.4). Type "olly": the searching line, then Polly Jones only, no place (AC-1.3). Type "x": "No people match “Pollx”" (AC-1.5).
3. Press Escape three times: the list closes, the text clears, the menu comes back unchanged and undecided (AC-3.3). Open the search on Ada again, then choose the entry on another row: Ada's menu comes back (AC-3.2). Press × on the open one (AC-3.1).
4. Search "Polly" on Ada and pick Polly Jones: the menu shows "Jones, Polly", Decided; in under a second the band, Contested when it conflicts, and "Your choice" (AC-2.1). Open the menu: Polly Jones sits before New Person, checked (AC-2.2). If the band is None it reads "Little evidence" (AC-4.1). Open the evidence sheet: its title reads the same; the evidence lists Polly Jones's values. Search Polly again and pick her: no second menu item (AC-2.3).
5. Press Done: Ada files on Polly Jones's record (AC-2.4). The slice Check is true.
6. VoiceOver: the entry, the field ("Find another record for Ada"), a hit, ×, and the evidence button (AC-4.4, AC-5.1). Open the toolbar omnibar, search "Polly", and VoiceOver a hit: it reads title, kind, subtitle, ref as before (AC-5.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The swapped-out `PVSelect` comes back still showing the sentinel | D-7 removes it from the hierarchy; manual check 3 |
| The hit row crashes in the popup for want of a `WorkspaceSession` | D-12 applies `.environment(model.session)` inside the row builder |
| The omnibar's VoiceOver changes when the label moves onto the row | Manual check 6 (AC-5.2) |
| The board draws something these Decisions do not allow | The plan is refreshed with `plan-pr` when the board is agreed, before it is Planned |

**Stop and report if:**

- After the search closes, the `PVSelect` field shows "Find another record…" as its value (D-7 is wrong).
- `PVComboBox` returns `.handled` for Escape with an empty field and a closed list, so `.onExitCommand` never fires (D-8).
- VoiceOver on a picker hit reads the row in layout order instead of `accessibilityLabel(for:)`, or an omnibar hit reads differently from before (D-10).
- The agreed S11-D3 board draws an action item, divider or section header in the menu, or a picker outside the target cell (working rule "The menu entry is a sentinel option").
- `PromoteFlow.hold` or `CatalogSearchModel` no longer has the signature in §4.

## 12. Open questions

1. **Board.** S11-D3 has to be drawn and agreed on the Promote flow board. Recommended: paste the brief into the same Claude Design board S11-D1 and S11-D2 used, and record the frame numbers in the brief's Board notes. This plan is refreshed then.
2. **"Little evidence" (FR-7).** Recommended: the neutral band Badge reads "Little evidence" for a chosen record at band None, on the row and the sheet, and an offered None keeps "No match" (D-9). The board may word it differently.
3. **Closing on focus loss (D-8).** Recommended: the search stays open until a pick, Cancel or Escape, as the Source page's type editor does, rather than closing when the field loses focus.

## 13. PR

**Title:** S11-36 — Find another record in the target menu
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-36-find-another-record.md`.
