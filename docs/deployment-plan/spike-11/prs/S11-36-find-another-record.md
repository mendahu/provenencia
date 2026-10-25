# S11-36 — Find another record in the target menu

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 8 — Find another record
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S11-D3](../design/S11-D3-promote-target-menu.md), board agreed 2026-10-25
**Board:** [Claude Design — Promote flow](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share), frames 20, 21, 22, 22b, 22c, 23, 23b, 24, 25 (and its sheet 25s), 26, 26b. The brief's **Board notes** list where the board differs from the brief; the board wins, and this plan already follows it
**Depends on:** S11-35 (landed [#543](https://github.com/mendahu/provenencia/pull/543)), S11-37 (landed [#551](https://github.com/mendahu/provenencia/pull/551)), S11-D3 (agreed)
**Blocks:** —
**Branch:** `s11-36-find-another-record`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-searchable-kind`](../../../../.agents/skills/add-searchable-kind/SKILL.md) (STEP-7 edits it)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [Find another record § Shape](../promote-find-another-record.md#shape), [§ Where the Swift side changes](../promote-find-another-record.md#where-the-swift-side-changes), [§ Decided](../promote-find-another-record.md#decided-user-2026-10-24); working rules "A found record is a decided row, nothing new", "No new search RPC, and one search model", "An inline picker, every record of the kind, no warning", "The menu entry is a sentinel option", "The picked record joins the row's menu"; the map's S11-36 scope entry (picker on `PVComboBox`, the hit row's VoiceOver label, `selected`); [`macos/App/DesignSystem/README.md`](../../../../macos/App/DesignSystem/README.md) § "The combo box subset"
**Planned:** 2026-10-25 against `main` @ `056b6ed` (code unchanged at `bfb852a9`)
**Review:** [`S11-36-find-another-record.review.md`](S11-36-find-another-record.review.md) — round 1 answered; round 2 running

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

A researcher on the Promote page who knows a Subject is a record Promote did not offer can find that record from the row and file the Subject on it. Every working row's target menu ends with "Find another record…". Choosing it turns the row's target cell into a search field over every record of the row's kind; picking a hit holds it as the row's decided target through S11-35's `hold`, so the row shows the record's own band, Contested and evidence, and Done files it. A record the researcher chose that the evidence rates below Weak reads "Little evidence" instead of "No match".

**Why now.** Since S11-05 a record whose first given name or surname differs from the Subject's is never a candidate, and the only route left is New and a merge the app does not have ([the problem](../promote-find-another-record.md#the-problem)). S11-35 built the call and S11-37 the shared search; this PR is the surface that uses them. It makes the spike's dogfood bar 11 true. Building it on the kit's `PVComboBox` needs four small kit fixes, made here (user, 2026-10-25): a custom empty line, reporting the box's own query resets, the host environment in the popup, and the hit row's dead `selected`.

**What stays.** Go, the FFI and Promote's proposals are unchanged. The menu's other items and their order, how a record item is drawn, the band, Contested, Decided and Updated Badges, the row's notes and subrows, the evidence sheet's content, the toolbar omnibar's look and behavior, and every existing `PVComboBox` caller's behavior.

## 2. User stories

**US-1.** As a researcher, when Promote offers nothing (or the wrong records) for a Subject, I choose "Find another record…" and search the row's kind, so that I can reach a differently named record.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A person row Promote left unset (menu: New Person, Skip); open its menu | The last item is "Find another record…" with a search icon, never checked (FR-1, frame 20) |
| AC-1.2 | Choose "Find another record…" on that row | The target cell shows a focused search field "Search people" and a Cancel search button; the row is still not Decided; no proposal is sent (FR-2, frame 21) |
| AC-1.3 | Type "Polly" in a catalog holding person PER-FAR "Polly Jones" and place PLC-1 "Polly Street" | The list shows PER-FAR only, drawn as the omnibar's hit row (FR-3, frame 22) |
| AC-1.4 | Type "P" | The list line reads "Type at least two letters of a name or a ref"; no search runs (FR-4, frame 21) |
| AC-1.5 | Type "Pollx" | The list line reads "No people match “Pollx”" (FR-4, frame 23) |
| AC-1.6 | The store's search fails | The list line shows the failure's own message (FR-4, frame 23b) |

**US-2.** As a researcher, when I pick a hit, it becomes the row's target like a menu pick, so that I see its evidence and Done files it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Pick PER-FAR from the list | The search closes; the menu shows the record selected; the row is Decided; the page proposes again with `{subjectID: a, handleID: e-far}` fixed (FR-5, frame 24) |
| AC-2.2 | Open the menu after the pick | PER-FAR is an item after the offered records and before New Person, checked (FR-5, frame 24) |
| AC-2.3 | Pick a hit already in the row's menu | That item is selected; the menu has no second copy (FR-5) |
| AC-2.4 | Press Done | The batch row for the Subject files on `e-far` |

**US-3.** As a researcher, opening and cancelling the search changes nothing, so that I can look without deciding.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Open the search on a row, then press Cancel search | The menu comes back with the same target; the row is not Decided; no proposal was sent (FR-6) |
| AC-3.2 | Open the search on row A, then choose "Find another record…" on row B | Row A shows its menu again, unchanged; row B shows the search (FR-6, frame 26) |
| AC-3.3 | In the search, type "Pol", press Escape (list closes), Escape (text clears, hint line on reopen), Escape | The search closes as in AC-3.1 (FR-6, frame 26) |
| AC-3.4 | Type "Pol", then click the page outside the field | The list closes and the typed text is dropped; the search field and Cancel stay in the target cell; clicking the field reopens the list on the hint line, not on old hits (FR-6, frame 26) |
| AC-3.5 | Press Done while a search is open | The search closes, and every menu is disabled while filing (FR-9, frame 26b) |

**US-4.** As a researcher, a record I chose that the evidence rates below Weak does not read as "No match", so that the row does not contradict my choice.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A row decided on a found record; Go answers band None, not contested | The assessment reads "Little evidence · Your choice", on the row and the sheet title (FR-7, frames 25, 25s) |
| AC-4.2 | The same, contested | "Little evidence", then Contested, then "Your choice" (FR-7, frame 25) |
| AC-4.3 | A row the researcher decided on an offered alternative Promote rates None | The band reads "Little evidence" (board: S11-D2 E03 reads it too) |
| AC-4.4 | A suggested row (not decided) whose proposed record Promote rates None | The band still reads "No match" (frame 25: Ken's PER-7KG40) |
| AC-4.5 | VoiceOver on the AC-4.1 row's evidence button | "Evidence for Ada: Little evidence, Your choice" (FR-8) |

**US-5.** As a VoiceOver user, the entry, the field, each hit and Cancel are read in words.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | VoiceOver on the entry, the field, a hit, the Cancel button | "Find another record…"; "Find another record for Ada"; the hit's `CatalogSearchHitPresentation.accessibilityLabel(for:orders:)` with the session's name order ("Polly Jones, Person, PER-FAR" with the fake store's data); "Cancel search" (FR-8) |
| AC-5.2 | VoiceOver on an omnibar hit after this PR | Reads title, kind, subtitle, ref, match context, as before this PR |

## 3. Scope

| In | Covered by |
| --- | --- |
| One spelling and one decoder for every target-menu token; `setTarget` refuses a token it cannot resolve (D-1, D-2) | STEP-1, AC-3.1 |
| `PromoteModel`'s one open record search: open by the sentinel, query, pick through `hold`, close, closed by Done (D-3 – D-6) | STEP-2, AC-1.2, AC-1.3, AC-2.1, AC-3.1, AC-3.2, AC-3.5 |
| `CatalogSearchModel.isBelowMinimumQuery` for the list's lines (D-13) | STEP-2, AC-1.4 |
| "Little evidence" for a chosen record at band None, on the row, the sheet title and the spoken label (FR-7, D-9) | STEP-3, AC-4.1 – AC-4.5 |
| Kit: `PVComboBox` custom-empty init, reported query resets, host environment in the popup; `PVOmnibarHitRow` and `CatalogSearchHitRow` lose `selected`; the hit row carries its VoiceOver label (D-10, D-11, D-15 – D-17) | STEP-4, STEP-5, AC-3.3, AC-3.4, AC-5.1, AC-5.2 |
| The menu entry and the search in the target cell (FR-1 – FR-6, FR-8, FR-9; D-7, D-8, D-12, D-14) | STEP-6, AC-1.1 – AC-1.6, AC-2.1 – AC-2.3, AC-3.3, AC-3.4 |
| Docs: feature doc, skill, ideas note, map, completed, brief archive | STEP-7 |

| Out | Home |
| --- | --- |
| Searching from the evidence sheet; comparing two records before picking; records of another kind; merging records | not planned in this spike ([what stays later](../promote-find-another-record.md#what-stays-later)) |
| A Callout or mark for a `Different` pick | not planned (refused, user 2026-10-24) |
| A band, score or Contested mark per hit in the list | not planned (S11-D3 §2.1) |
| An action item, divider or section header in `PVSelect` | not planned (working rule "The menu entry is a sentinel option") |
| Hiding the kind Badge in `PVOmnibarHitRow` for a one-kind list | not planned (S11-D3 non-goals) |
| A shared popup presenter; one popup list under the combo box and the omnibar | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) "One popup dismissal", "One popup list under the combo box and the omnibar" |
| Moving `CatalogSearchModel.selectedIndex`, `moveSelection`, `selectedHit` onto `OmnibarResultsModel` | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) (STEP-7 adds the line) |
| Any change to `SearchCatalog`, its ranking, or Go | not planned (working rule "No new search RPC, and one search model") |
| Closing the search when the field loses focus | not planned (board frame 26: the search stays until a pick or Cancel) |
| The proposal wire strings `"new"`, `"skip"`, `"handle"` in `fixed(for:)` and `batchRow` | not planned: they are proposal vocabulary, not menu tokens |
| Rewriting the `"handle:…"` literals in `macos/ProvenenciaTests` | not planned: tests keep spelling the token they assert |
| S11-D4 (the kit project's ComboBox and omnibar parity) | the design-system project; no PR |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Target tokens | `macos/App/Features/Promote/PromoteFlow.swift` `PromoteFlow.Target.token` (line 16), `target(for:on:)` (line 638) | `"handle:\(id)"`, `"new"`, `"skip"`, `""`. `target(for:on:)` parses `"skip"`, `"new"` and the `"handle:"` prefix; an unknown token, or a handle not in `row.menu`, returns the row's current target |
| setTarget | `PromoteFlow.setTarget(subjectID:token:)` (line 432) → `select(_:at:)` (line 462) | Any token on a working row marks the row decided, sets `manual`, clears `updated` and returns true, even when the target does not change |
| hold | `PromoteFlow.hold(subjectID:id:ref:header:)` (line 442); `PromoteModel.hold(subjectID:id:ref:header:)` (line 179) | Appends an unoffered record to `row.menu` with reason `decided` and band None, or reuses the offered item, then selects it; the model re-proposes. No caller outside tests |
| Readout | `PromoteFlow.SelectionReadout` (line 133), `Row.selectionReadout` (line 192) | Carries `assessment`, `reason`, `contested`; not whether the row is decided. A menu pick keeps its offered reason; only a found record has reason `decided` |
| Model target call | `macos/App/Features/Promote/PromoteModel.swift` `setTarget(subjectID:token:)` (line 173); `done()` (line 260), `leave()` (line 268) | Forwards to the flow; re-proposes when it returns true |
| Model state | `PromoteModel` `private(set) var flow`, `session`, `store`, `isSaving` (lines 93 – 126) | `@Observable @MainActor` class; `session.projectKey.projectDir` and `session.nameFormat` are what search calls need |
| Target cell | `macos/App/Features/Promote/PromoteView.swift` `PromoteAlignmentRow` (line 309): `PVSelect` (line 331), `targetOptions` (line 427), `targetBinding` (line 436) | `PVSelect` of `row.menu` items, New, Skip; the binding's setter calls `model.setTarget`. The cell is `PromoteLayout.targetColumn` (330 pt). The whole row is `.disabled(model.isSaving)` in `sectionBlock` |
| Menu rows | `PromoteView.swift` `PromoteTargetMenuRow` (line 505) | Literal `"new"` / `"skip"` checks; decodes a record item with `row.menu.first { "handle:\($0.id)" == option.id }`; draws nothing for any other option |
| PVSelect state | `macos/App/DesignSystem/Components/Select/PVSelect.swift` `session` (`PVSelectSession`), `.onChange(of: selection)` (line 172), `.onChange(of: session.selection)` (line 183) | Keeps its own `session.selection`, copied back from the binding only when the binding's value changes. A setter that ignores a token leaves the field showing that token's label |
| Band copy | `macos/App/Platform/L10n.swift` `L10n.Promote.assessment(_ assessment:)` (line 1355); callers `PromoteReadout.swift` `PromoteReadoutBadges` (line 11), `PromoteModel.assessmentText(_:)` (line 318); `PromoteModelTests.swift` lines 566 – 585 | Maps the band alone; None reads "No match". Drawn on the row (`PromoteView.swift` line 354) and the sheet title (`PromoteEvidenceSheet.swift` line 90) |
| Search model | `macos/App/Features/Catalog/Search/CatalogSearchModel.swift` | `init(kinds:)`, `query`, `hits`, `isLoading`, `searchError`, `trimmedQuery`, `minQueryLength` (2), `showsResults`, `showsErrorState`, `showsEmptyState`, `scheduleSearch(projectDir:location:store:)` (under two characters it calls `dismiss()`), `cancel()` |
| Hit row | `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift`; `macos/App/DesignSystem/Recipes/OmnibarHitRow/PVOmnibarHitRow.swift` | `CatalogSearchHitRow` forwards `selected` to `PVOmnibarHitRow`, which fills `surfaceHover` when true (line 87); only its own preview passes `true`. No accessibility label: the combined element reads in layout order |
| Spoken hit | `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` `accessibilityLabel(for:)` | Titles with `title(for:)`'s default `NameFormatOrders()`, not the session's order the row shows |
| Omnibar call site | `macos/App/Features/Workspace/OmnibarResultsHost.swift` `OmnibarResultsPanel` (line 92) | Passes `selected: false` (line 117) and sets `.accessibilityLabel(…accessibilityLabel(for: hit))` on the `Button` (line 127) |
| Combo box inits | `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift` `PVComboBox`; extensions `where Empty == Text` (line 742) and `where Row == PVComboBoxPlainRow, Empty == Text` (line 779) | The struct's `@State private` properties make its memberwise init private, so a feature cannot pass `empty:`; both public inits fix a static `emptyLabel` |
| Combo box resets | `PVComboBox.closeAndRestore()` (line 662), the Escape clear in `handleKey` (line 712), `typedText` | Only the `typedText` setter calls `onQueryChange`. Click-away and the second Escape set `query = selectedLabel` without telling a remote caller, which keeps the old query and hits |
| Popup environment | `PVComboBox.swift` `PVComboBoxPopupWindow.updateNSView` (line 837), `Coordinator.setContent` (line 887) | Hands `content` to an `NSHostingView` in a child window. The host's environment (a `WorkspaceSession`, `isEnabled`) does not reach the rows |
| Precedent | `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `metaRow` (line 185); `macos/App/Features/Sources/SourceIdentitySection.swift` `editingType`, `beginEditType()`, `cancelEditType()` | Swaps an inline `PVComboBox` (`focus: PVComboFocus($typeFieldFocused, equals: .combo)`, `.onAppear { DispatchQueue.main.async { typeFieldFocused = .combo } }`) and a `PVIconButton(.dismiss, …)` in place of the type pill while the model flag is true |
| Strings | `macos/App/Platform/L10n.swift` `enum Promote` (line 1305), `newOption(_:)` (line 1708); `macos/App/Resources/Localizable.xcstrings` | Kind-switched copy follows `newOption`'s shape |
| Tests | `macos/ProvenenciaTests/PromoteFlowTests.swift`; `macos/ProvenenciaTests/PromoteModelTests.swift` (`aFoundRecordIsProposedHeldAndFiled`, `store(subjects:)`, `model(store:subjectID:)`); `macos/ProvenenciaTests/CatalogSearchTests.swift`; `macos/ProvenenciaTests/PVComboBoxTests.swift`; `macos/ProvenenciaTests/QueryHandleWaiting.swift` `waitUntil`, `waitForSearch` | `FakeStore.searchCatalog` searches `membershipBySubject` entities by kind, titling a hit from its entity label; `searchCatalogError` makes it throw |

**Precedent.** The Source page's type editor (`3e9a0712` "Sync the Source page to explicit edit flows and DateValue editing.", reshaped in `2823c737`) swaps a cell for an inline `PVComboBox` plus a dismiss `PVIconButton`, focuses it on appear, and keeps the editing flag on the model. Copy the swap, the focus-on-appear and the model-owned flag. Do not copy its `onChange(of: typeDraftID)` save: this picker commits through `onCommit`. #554 added `isLoading` to `PVComboBox`; copy its shape (one documented parameter, a `PVComboBoxMatch` helper, a `PVComboBoxTests` case) for the kit changes.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PromoteFlow.Target` gains `static let findToken = "find"` *(new)* and `static func handleToken(_ id: String) -> String` *(new)*. `Row.menuItem(forToken:) -> Alternative?` *(new)* is the one decoder of a record token. Every option value in `PromoteView` is spelled from `Target` (`Target.newKind.token`, `Target.skip.token`, `handleToken`, `findToken`); `PromoteTargetMenuRow` decodes with `menuItem(forToken:)` and compares the rest with `Target` tokens | One spelling and one decoder for the four tokens (review M-5); the menu and the parser cannot drift | A token enum — because `PVSelect` values are strings |
| D-2 | `target(for:on:)` returns `Target?`: `nil` for a token it cannot resolve (unknown, `findToken`, a handle not in the menu). `setTarget` returns false and changes nothing on `nil` | Any unresolved token decides nothing, so the flow never needs to know the sentinel (review M-4) | A `findToken` special case in `setTarget` — because every other unresolved token would still decide the row |
| D-3 | `PromoteModel.setTarget(subjectID:token:)` sends `Target.findToken` to `openRecordSearch(subjectID:)` *(new)* and returns before calling the flow. `PromoteView`'s binding keeps calling `model.setTarget` | The working rule's "the row's binding catches the token before `setTarget`", in a place a model test drives | Catching it in the view's binding — because no test can drive a view binding |
| D-4 | `PromoteModel` holds one open search: `struct RecordSearch { let subjectID: String; let search: CatalogSearchModel }` *(new, nested)* and `private(set) var recordSearch: RecordSearch?` *(new)*. `openRecordSearch(subjectID:)` does nothing for an anchor or a missing row; otherwise it cancels any open search and sets one with `CatalogSearchModel(kinds: [row.kind.rawValue])` | One search at a time (FR-6), kind from the row; one optional cannot leave a subject id without its search. One noun across the model and view (review M-10) | Per-row `@State` — because opening on another row must close the first and tests cannot reach view state |
| D-5 | `recordSearchQueryChanged(_ text: String)` *(new)* sets `recordSearch.search.query = text` and calls `scheduleSearch(projectDir: session.projectKey.projectDir, location: nil, store: store)` | The S11-37 model owns debounce, generation, the two-character floor and errors; Promote has no location | Calling `store.searchCatalog` from Promote — because it rebuilds the debounce |
| D-6 | `pickSearchHit(id: String)` *(new)* finds the hit in `recordSearch.search.hits`, calls `closeRecordSearch()`, then `hold(subjectID:id:ref:header:)` with the hit's `subjectID`, `id`, `ref`, `header`. An id not in `hits` does nothing. `closeRecordSearch()` *(new)* calls `search.cancel()` and sets `recordSearch = nil`. `done()` and `leave()` call `closeRecordSearch()` first | The pick goes through S11-35's call, which titles the item and reuses an item already in the menu (AC-2.3); a search never outlives the page's work (AC-3.5) | Passing a `CatalogSearchHit` — because the combo box commits a value string |
| D-7 | While `model.recordSearch?.subjectID == row.subjectID`, the target cell shows `PromoteRecordSearch` *(new, private, in `PromoteView.swift`)* instead of the `PVSelect`, in an `if` / `else`, both `.frame(width: PromoteLayout.targetColumn)` | Dropping the `PVSelect` drops its `session.selection` copy (Current state "PVSelect state"), so the sentinel never shows as the selection. Same swap as the Source type editor | Keeping the select mounted — because it would keep showing the sentinel's label; a `PVSelect` change — the working rule gives it no action item |
| D-8 | `PromoteRecordSearch(row: PromoteFlow.Row, model: PromoteModel)` is an `HStack(spacing: PVSpacing.space2)` of `PVComboBox` (`size: .sm`, remote results, default `maxListHeight` 288; the board draws 290) and `PVIconButton(.dismiss, label: L10n.Promote.searchCancel, size: .sm)` calling `model.closeRecordSearch()`. Focus: `@FocusState private var focus: Field?`, `enum Field { case query }`, `PVComboFocus($focus, equals: .query)`, set in `.onAppear { DispatchQueue.main.async { focus = .query } }`. Escape: the box's own stages; when the box returns `.ignored` (list closed, field empty), `.onExitCommand { model.closeRecordSearch() }` on the `HStack` closes the search. Losing focus does not close it | Board frame 26 and the Source page's cancel; `onExitCommand` is the standard Escape route, not an event monitor (map scope) | Closing on focus loss — the board keeps the search; an `NSEvent` monitor — refused by the map scope |
| D-9 | `SelectionReadout` gains `var chosen = false` *(new)*; `Row.selectionReadout` sets it from `decided`. `L10n.Promote.assessment(_ readout: PromoteFlow.SelectionReadout) -> LocalizedStringResource` replaces the band-only function: `assessmentChosenNone` *(new)* when `readout.assessment == .none && readout.chosen`, else the band's resource. The band-only mapping becomes `private`. `PromoteReadoutBadges` and `PromoteModel.assessmentText` call the readout form; tone unchanged | One function words the band for the row, the sheet and VoiceOver (review M-1). The board applies it to every researcher choice, a menu pick at None included (S11-D2 E03), and leaves a suggested None as "No match" (frame 25) | Keying on `reason == .decided` — because a menu pick keeps its offered reason and the board reads it "Little evidence" too; a flag on `Alternative` — the row already knows it is decided |
| D-10 | `CatalogSearchHitRow` sets `.accessibilityLabel(CatalogSearchHitPresentation.accessibilityLabel(for: hit, orders: session.nameFormat))` on its `PVOmnibarHitRow`. `accessibilityLabel(for:)` gains `orders: NameFormatOrders = NameFormatOrders()` and passes it to `title(for:orders:)`. `OmnibarResultsPanel` drops its own label | Both callers get the label from the one row, and the spoken title matches the shown one (review X-7) | Labelling at each call site — two places to remember |
| D-11 | `PVOmnibarHitRow` loses `selected` and its `.background` fill, and its preview drops `selected: true`; `CatalogSearchHitRow` loses `selected`; both callers stop passing it | Both hosts draw selection outside the row (`PVMenuRowButtonStyle`, the combo box's row shell); a parameter no caller sets is dead (review M-6; user, 2026-10-25) | Keeping it as recipe API — nothing real sets it |
| D-12 | Combo box wiring: options `recordSearch.search.hits.map { PVComboBoxOption(value: $0.id, label: CatalogSearchHitPresentation.title(for: $0, orders: model.session.nameFormat)) }`; `selection: .constant("")`; `isLoading: recordSearch.search.isLoading`; `onQueryChange: model.recordSearchQueryChanged`; `onCommit: { id, _ in model.pickSearchHit(id: id); return false }`; the row builder draws `CatalogSearchHitRow(hit:projectDir: model.session.projectKey.projectDir)` for `hits.first { $0.id == option.value }`, and `EmptyView()` when there is none | Board frames 21 – 24; with D-17 the popup carries the session, so the builder needs no environment of its own | `.environment(model.session)` in the builder — fixed once in the kit instead (review M-3) |
| D-13 | `CatalogSearchModel` gains `var isBelowMinimumQuery: Bool { trimmedQuery.count < Self.minQueryLength }` *(new)*, and `showsResults` and `scheduleSearch`'s guard read it. The picker's empty line: `showsErrorState` → the error (verbatim); `isBelowMinimumQuery` → `L10n.Promote.searchHint`; else `L10n.Promote.searchNoMatches(row.kind, query: trimmedQuery)` | The search model owns "what the list says" for both search surfaces (review M-2); board frames 21, 23, 23b | Re-deriving the order in Promote — a second reading of the same state |
| D-14 | Accessibility identifiers: the entry option `promote.row.<ref>.target.find` (on its `PVSelectOption`); the combo box `accessibilityIdentifierPrefix: "promote.row.<ref>.search"` (the box adds `.list`, `.option.<id>`, `.loading`, `.empty`); Cancel `.accessibilityIdentifier("promote.row.<ref>.search.cancel")` on the `PVIconButton` | Call-site ids under the row's `promote.row.<ref>.` prefix (`add-ui-component`) | — |
| D-15 | `PVComboBox` gains a public init in an extension in `PVComboBox.swift` (no constraint on `Row` or `Empty`) with the same parameters as the `Empty == Text` init minus `emptyLabel`, plus `@ViewBuilder row:` and `@ViewBuilder empty: @escaping (String) -> Empty`; the `Empty == Text` init calls it. The type comment's "Content is the caller's" paragraph names it | The memberwise init is private (Current state "Combo box inits"); the board's empty line changes with the search's state (review X-1) | A static `emptyLabel` — cannot carry an error message |
| D-16 | In remote mode (`onQueryChange != nil`), when the box itself puts the field back (`closeAndRestore()` and the Escape clear), it calls `onQueryChange(query)` after assigning, when the text changed | A remote caller's query and hits follow the field (review X-2), so a click away or Escape never leaves stale hits behind (AC-3.3, AC-3.4). Local-filter callers do not pass `onQueryChange` and see no change | Resetting from Promote — the box is the only one that knows it restored the text |
| D-17 | `PVComboBoxPopupWindow.updateNSView` passes `content.environment(\.self, context.environment)` to `setContent` | Rows in the popup see what the field sees: the `WorkspaceSession` `CatalogSearchHitRow` reads, and `isEnabled` (FR-9). Fixed once for every combo box (review M-3) | Applying `.environment` in each caller's row builder |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PromoteFlow.Target.findToken`, `handleToken(_:)`, `Row.menuItem(forToken:)` *(new)* | constants / func | the `"handle:"`, `"new"`, `"skip"` literals in `PromoteFlow.swift` and `PromoteView.swift` | They replace those literals and the second decoder (D-1) |
| `PromoteModel.RecordSearch`, `recordSearch`, `openRecordSearch`, `recordSearchQueryChanged`, `pickSearchHit`, `closeRecordSearch` *(new)* | type / funcs | `SourceIdentitySection.editingType`, `beginEditType`, `cancelEditType` | Same shape on this page's model |
| `PromoteRecordSearch` *(new, private)* | view | `SourcePageIdentityHeader.metaRow`'s combo + cancel; `OmnibarResultsPanel` | One call site, a feature snowflake composing `PVComboBox`, `CatalogSearchHitRow`, `PVIconButton` (S11-D3 inventory). The second combo-plus-cancel swap after the Source editor; no third, so no recipe yet |
| `SelectionReadout.chosen` *(new)* | field | `Row.decided` | The readout is what the badges and the spoken label get; it carries the one fact they need |
| `CatalogSearchModel.isBelowMinimumQuery` *(new)* | property | `showsResults`, `scheduleSearch`'s guard | Names the check both already make |
| `PVComboBox` general init *(new)* | init | the two `Empty == Text` inits | They fix a static empty label; they now call it |
| L10n `findOption`, `searchPlaceholder(_:)`, `searchLabel(name:)`, `searchCancel`, `searchHint`, `searchNoMatches(_:query:)`, `assessmentChosenNone` *(new)* | keys | `L10n.Promote.newOption(_:)`; `L10n.Workspace.omnibarNoMatchesTitle` | Promote copy lives in `L10n.Promote`; the omnibar's line is not by kind |

**Builds on.** S11-35's `hold`; S11-37's `CatalogSearchModel` and `CatalogSearchHitRow`; `PVComboBox` remote results (#554); the Source type editor's swap.
**Deletes.** Every `"handle:"`, `"new"` and `"skip"` menu-token literal outside `PromoteFlow.Target` in `macos/App/Features/Promote/PromoteView.swift` (D-1); `PromoteTargetMenuRow`'s own decoder (D-1); the band-only public `L10n.Promote.assessment(_:)` (D-9); `PVOmnibarHitRow.selected`, its fill and its `selected: true` preview, `CatalogSearchHitRow.selected` and both `selected: false` arguments (D-11); the `.accessibilityLabel` on `OmnibarResultsPanel`'s hit `Button` (D-10).
**Next adds.** A later "pick a record" surface (merge, link from the evidence sheet) composes `PVComboBox` (D-15 init) + `CatalogSearchModel` + `CatalogSearchHitRow` as `PromoteRecordSearch` does, and gets the label, the environment and the resets free. At a third combo-plus-cancel swap, `PromoteRecordSearch`'s body is the candidate recipe. S11-45 and S11-47 add band rules to the one `L10n.Promote.assessment(_:)`.

### Schema

None — no catalog change.

### Go core

None — Go already holds, scores, explains and files a found record (S11-35).

### FFI

None — `SearchCatalog` with `kinds` and `ProposePromoteGraphAlignment` with fixed pairs are used as they are.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Target cell | `PVSelect` or `PromoteRecordSearch` | `macos/App/Features/Promote/PromoteView.swift` `PromoteAlignmentRow` | D-7 |
| Menu entry (FR-1, frame 20) | `PVSelectOption` | `targetOptions` | Appended after Skip: `PVSelectOption(value: PromoteFlow.Target.findToken, label: L10n.string(L10n.Promote.findOption), accessibilityIdentifier: "promote.row.\(row.ref).target.find")` |
| Entry row (frame 20) | `PVIcon(.search, size: 14)` + text | `PromoteTargetMenuRow` | `HStack(spacing: PVSpacing.space3)`; `Text(verbatim: option.label)` in `PVFont.body(size: PVTypeScale.bodySmall)`; icon and text `PVColor.textSecondary` |
| Search field + list (frames 21 – 23b) | `PVComboBox` (D-15 init) | `PromoteRecordSearch` | D-8, D-12, D-13. `placeholder: L10n.Promote.searchPlaceholder(row.kind)`, `label: L10n.Promote.searchLabel(name: row.name)` |
| Hit row (frames 22, 22b) | `CatalogSearchHitRow` → `PVOmnibarHitRow` | `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift` | D-10, D-11; the ref is accented on a ref match by the existing `refAccent` |
| Cancel (frame 21) | `PVIconButton(.dismiss, label: L10n.Promote.searchCancel, size: .sm)` | `PromoteRecordSearch` | D-8, D-14 |
| Picked (frame 24) | `PVSelect` | `PromoteAlignmentRow` | The select comes back with the held record; the assessment column is empty until Promote answers, as after any menu change |
| Band wording (frames 25, 25s) | `PVBadge` | `macos/App/Features/Promote/PromoteReadout.swift` `PromoteReadoutBadges` | D-9 |

The board frames, in words, for an implementer without the board: (20) the menu ends with the search icon and "Find another record…" after Skip, never checked; (21) the target cell is a small "Search people" field and an × button filling the 330 pt column, the list on the two-letter line, assessment empty, no Decided; (22) four person hit rows, the first highlighted, one with a "Name: …" match line; (22b) a typed ref, the matched refs in accent; (22c) the box's searching line; (23) "No people match “Mairon”"; (23b) the search's own error message; (24) the record selected, Decided, Weak, Contested, "Your choice", and in the open menu after the offered record and before New, checked; (25) "Little evidence · Your choice", one with Contested, a suggested None still "No match", and the sheet title (25s) reading "Little evidence"; (26) a search open on one row, the event rows' menus untouched; (26b) while filing, menus, the field and × disabled.

States: empty field and one character → hint line; searching with no hits → the box's searching line; no hits → no-match line; failure → the error message; filing → `.disabled(model.isSaving)` on the row, now reaching the popup rows through D-17 (FR-9).

### Strings

L10n keys *(new)*, in `L10n.Promote` and `Localizable.xcstrings` (board copy):

- `findOption` — `promote.target.find` = "Find another record…"
- `searchPlaceholder(_ kind:) -> LocalizedStringResource` — `promote.search.placeholder.person` = "Search people", `.event` = "Search events", `.place` = "Search places"
- `searchLabel(name:) -> String` — `promote.search.label` = "Find another record for %@"
- `searchCancel` — `promote.search.cancel` = "Cancel search"
- `searchHint` — `promote.search.hint` = "Type at least two letters of a name or a ref"
- `searchNoMatches(_ kind:, query:) -> String` — `promote.search.noMatches.person` = "No people match “%@”", `.event` = "No events match “%@”", `.place` = "No places match “%@”"
- `assessmentChosenNone` — `promote.assessment.chosenNone` = "Little evidence"

Removed: none (the band-only function becomes `private`; its keys stay).

## 7. Steps

**STEP-1 — One spelling and one decoder for target tokens**
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `macos/App/Features/Promote/PromoteView.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2. Add T-1, T-2. Existing tests keep their token literals.
Check: `rg '"(handle:|new|skip|find)' macos/App/Features/Promote/PromoteView.swift` prints nothing; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PromoteFlowTests` passes.

**STEP-2 — The model's one record search**
Files: `macos/App/Features/Promote/PromoteModel.swift`, `macos/App/Features/Catalog/Search/CatalogSearchModel.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`, `macos/ProvenenciaTests/CatalogSearchTests.swift`.
Do: D-3 – D-6, D-13's `isBelowMinimumQuery`. Add T-3 – T-6, T-8.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PromoteModelTests -only-testing:ProvenenciaTests/CatalogSearchTests` passes.

**STEP-3 — "Little evidence" for a chosen record at band None**
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Promote/PromoteReadout.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`. Skill: `add-localized-string`.
Do: D-9; `assessmentChosenNone`; move `PromoteModelTests` lines 566 – 585 onto the readout form; set the readout preview's `.none, .decided` case to `chosen: true`. Add T-7.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PromoteModelTests` passes.

**STEP-4 — Combo box: a custom empty line, reported resets, the host environment**
Files: `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift`, `macos/ProvenenciaTests/PVComboBoxTests.swift`, `macos/App/DesignSystem/README.md` (§ "The combo box subset"). Skill: `add-ui-component`.
Do: D-15, D-16, D-17. Put D-16's rule in a pure helper `PVComboBoxMatch.reportsRestore(isRemote:old:new:) -> Bool` *(new)* so T-9 tests it; the README's subset section gains one sentence each for the custom empty line, the reported resets and the environment.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PVComboBoxTests` passes; the Source type editor, the Add Source type field and the Citation composer combo boxes open, filter and commit as before (manual check 7).

**STEP-5 — The hit row: no `selected`, its own VoiceOver label**
Files: `macos/App/DesignSystem/Recipes/OmnibarHitRow/PVOmnibarHitRow.swift`, `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift`, `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift`, `macos/App/Features/Workspace/OmnibarResultsHost.swift`, `macos/ProvenenciaTests/CatalogSearchTests.swift`. Skill: `add-ui-component`.
Do: D-10, D-11. Add T-10.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CatalogSearchTests` passes; VoiceOver on an omnibar hit reads as before (manual check 6).

**STEP-6 — The menu entry and the search in the target cell**
Files: `macos/App/Features/Promote/PromoteView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skills: `add-ui-component`, `add-localized-string`.
Do: §6 macOS rows "Target cell" through "Picked"; D-7, D-8, D-12, D-14; every key under §6 Strings but `assessmentChosenNone`.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`; manual checks 1 – 5.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row; design track row), this plan (status, Landed note), `docs/deployment-plan/spike-11/design/S11-D3-promote-target-menu.md` → `docs/deployment-plan/spike-11/design/archive/`, `docs/deployment-plan/spike-11/design/README.md`, `docs/deployment-plan/spike-11/promote-find-another-record.md`, `.agents/skills/add-searchable-kind/SKILL.md`, `docs/ideas/shared-list-helpers.md`.
Do: per the Definition of done. Feature doc: rewrite "The menu entry:" under § Where the Swift side changes in the past tense (`Target.findToken`, `PromoteModel.recordSearch`, `PromoteRecordSearch`) and add "Little evidence" to § Shape's "Held, not moved". Skill: the `CatalogSearchModel` bullet adds that `CatalogSearchHitRow` carries its own VoiceOver label, and that a picker is `PVComboBox(onQueryChange:isLoading:empty:)` over `CatalogSearchModel(kinds:)`, as `PromoteRecordSearch`. Ideas: under "One popup list under the combo box and the omnibar", one sentence that `CatalogSearchModel.selectedIndex`, `moveSelection` and `selectedHit` are the omnibar's and move with that list.
Check: `python3 scripts/check-plan-docs.py` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `anUnresolvedTokenDecidesNothing` *(new)* | An unset row and a row on `handleToken("e1")`; `setTarget` with `Target.findToken`, `"bogus"`, and `handleToken("not-in-menu")` → each returns false; targets, `decided`, `updated`, `manual` unchanged | AC-3.1, D-2 |
| T-2 | Swift | `PromoteFlowTests.swift` `menuTokensAreSpelledAndDecodedOnce` *(new)* | `Target.handle(id: "e1", ref: "PER-1", title: "A").token == Target.handleToken("e1")`; `row.menuItem(forToken: handleToken("e1"))?.id == "e1"`; `menuItem(forToken: Target.newKind.token) == nil`; `setTarget(token: handleToken("e1"))` → target `e1` | D-1 |
| T-3 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `aRecordSearchOpensOnTheRowsKindAndDecidesNothing` *(new)* | Loaded page, row `a` unset; note `store.lastPromoteFixed`; `setTarget(subjectID: "a", token: Target.findToken)` → `recordSearch?.subjectID == "a"`, `recordSearch?.search.kinds == ["person"]`, row `a` undecided and unset, `isProposing` false, `lastPromoteFixed` unchanged | AC-1.2, D-3, D-4 |
| T-4 | Swift | `PromoteModelTests.swift` `cancellingOrMovingARecordSearchChangesNoRow` *(new)* | Ada and Bea; open on `a`, open on `b` → `recordSearch?.subjectID == "b"`; `closeRecordSearch()` → nil; both rows unchanged; `lastPromoteFixed` unchanged; open on `a`, `done()` → `recordSearch == nil` | AC-3.1, AC-3.2, AC-3.5, D-4, D-6 |
| T-5 | Swift | `PromoteModelTests.swift` `aFoundRecordIsSearchedPickedAndHeld` *(new)* | `membershipBySubject` holds person `e-far` label "Polly Jones" (PER-FAR) and place `p-1` label "Polly Street" (PLC-1, kind `place`); open on `a`; `recordSearchQueryChanged("Polly")`; `await waitForSearch(model.recordSearch!.search)` → hit ids `["e-far"]`; `pickSearchHit(id: "e-far")` → `recordSearch == nil`; after `waitUntil { !model.isProposing }`, `lastPromoteFixed == [{a, handleID: e-far}]`, row `a` decided on `e-far`; `done()` files `e-far` | AC-1.3, AC-2.1, AC-2.4, D-5, D-6 |
| T-6 | Swift | `PromoteModelTests.swift` `aRecordSearchShowsFailuresAndIgnoresUnknownHits` *(new)* | `store.searchCatalogError` set; open, `recordSearchQueryChanged("Polly")`, `waitForSearch` → `showsErrorState`; `pickSearchHit(id: "nope")` → search still open, row `a` unchanged | AC-1.6, D-6 |
| T-7 | Swift | `PromoteModelTests.swift` `aChosenRecordBelowWeakReadsLittleEvidence` *(new)* | `L10n.Promote.assessment` of readout (`.none`, chosen) → `assessmentChosenNone`; (`.none`, not chosen) → `assessmentNone`; (`.weak`, chosen) → `assessmentWeak`; `assessmentText` of (`.none`, chosen, contested) → "Little evidence, contested"; a row decided on an offered alternative at None has `selectionReadout?.chosen == true` | AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5, D-9 |
| T-8 | Swift | `macos/ProvenenciaTests/CatalogSearchTests.swift` `belowMinimumQueryIsTheTwoLetterFloor` *(new)* | query `""`, `"P"`, `" P "` → `isBelowMinimumQuery`; `"Po"` → not | AC-1.4, D-13 |
| T-9 | Swift | `macos/ProvenenciaTests/PVComboBoxTests.swift` `aRestoreIsReportedOnlyToRemoteCallers` *(new)* | `reportsRestore(isRemote: true, old: "Pol", new: "")` → true; `(true, "", "")` → false; `(false, "Pol", "")` → false | AC-3.3, AC-3.4, D-16 |
| T-10 | Swift | `CatalogSearchTests.swift` `accessibilityLabelUsesTheGivenNameOrder` *(new)* | A person hit with a header "John Henry Smith"; `accessibilityLabel(for:orders:)` with a surname-first order starts with the same title `title(for:orders:)` returns for that order | AC-5.1, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteFlowTests` (tokens), `PromoteModelTests` (band expectations), `CatalogSearchTests`, `PVComboBoxTests`, and `scripts/check-localizable-xcstrings.py`.

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
- [ ] Domain doc updated if a rule in docs/*.md changed: promote-find-another-record.md, add-searchable-kind skill, shared-list-helpers idea (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a project with a Source that has a person Subject "Ada" Promote offers nothing for, a filed person record with a different first name (create "Polly Jones" from another Source and promote it New), and a filed place whose name contains "Polly".

1. Open Promote on Ada. Her row is unset. Open the target menu: the last item is "Find another record…" with a search icon, unchecked (AC-1.1). Choose it: a focused "Search people" field and × replace the menu; no Decided Badge (AC-1.2).
2. Type "P": the hint line (AC-1.4). Type "olly": the searching line, then Polly Jones only, no place (AC-1.3). Type "x": "No people match “Pollx”" (AC-1.5).
3. Clear the field, type "Pol", click the page outside: the list closes, the text is gone, the field and × stay; click the field: the hint line, no old hits (AC-3.4). Type "Pol", press Escape three times: the list closes, the text clears, the menu comes back unchanged and undecided (AC-3.3). Open the search on Ada, then choose the entry on another row: Ada's menu comes back (AC-3.2). Press × (AC-3.1).
4. Search "Polly" on Ada and pick Polly Jones: the menu shows her, Decided; in under a second the band, Contested when it conflicts, and "Your choice" (AC-2.1). Open the menu: she sits before New Person, checked (AC-2.2). If the band is None it reads "Little evidence", and the evidence sheet's title too (AC-4.1). Search Polly again and pick her: no second item (AC-2.3). On another row Promote proposed at None, the band reads "No match" until you choose that record from the menu, then "Little evidence" (AC-4.3, AC-4.4).
5. Open a search on a row and press Done: the search closes and the page files; Ada files on Polly Jones's record (AC-2.4, AC-3.5). The slice Check is true.
6. VoiceOver: the entry, the field ("Find another record for Ada"), a hit, ×, and the evidence button (AC-4.5, AC-5.1). In the toolbar omnibar, search "Polly" and VoiceOver a hit: title, kind, subtitle, ref as before (AC-5.2).
7. Regression: the Source page's type editor, the Add Source type field, and a Citation composer combo box each open, filter, commit and dismiss as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The swapped-out `PVSelect` comes back showing the sentinel | D-7 removes it from the hierarchy; manual check 3 |
| Forwarding the environment into the popup changes an existing combo box's look | D-17 passes the field's own environment; manual check 7 |
| Reporting resets to remote callers changes a local caller | D-16 applies only when `onQueryChange` is set; no existing feature caller sets it; T-9 |
| The omnibar's VoiceOver changes when the label moves onto the row | Manual check 6 (AC-5.2) |

**Stop and report if:**

- After the search closes, the `PVSelect` field shows "Find another record…" as its value (D-7 is wrong).
- `PVComboBox` returns `.handled` for Escape with an empty field and a closed list, so `.onExitCommand` never fires (D-8).
- `content.environment(\.self, context.environment)` does not compile or does not carry the `WorkspaceSession` into the popup (D-17).
- An existing `PVComboBox` caller draws or behaves differently after STEP-4 (manual check 7).
- VoiceOver on a picker hit reads the row in layout order, or an omnibar hit reads differently from before (D-10).
- An existing test relies on an unresolved token deciding a row (D-2).
- `PromoteFlow.hold` or `CatalogSearchModel` no longer has the signature in §4.

## 12. Open questions

None.

## 13. PR

**Title:** S11-36 — Find another record in the target menu
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-36-find-another-record.md`.
