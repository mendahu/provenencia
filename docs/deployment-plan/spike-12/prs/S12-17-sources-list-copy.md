# S12-17 — Copy Source from the Sources list

**Status:** Planned
**Spike / slice:** Spike 12 · Slice 1 — Copy Source
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D6](../design/S12-D6-sources-list-copy.md), board agreed 2026-10-18
**Board:** [Claude Design — Sources List](https://claude.ai/design/p/118d2eb7-3b46-4912-a688-09ec7b0070fa?via=share), file `Sources List Frame.dc.html`, presets `rowmenu` (the one-item menu open on a row) and `copied` (the copy as the top row). If you cannot open it, §6 macOS describes the frames in words; build from §6
**Depends on:** S12-02 ([plan](S12-02-copy-source-form.md), landed in #485), S12-D6 (agreed)
**Blocks:** —
**Branch:** `s12-17-sources-list-copy`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** —
**Context:** [Copy Source](../source-structure.md#copy-source), working rule "Entry points"; [S12-02 plan](S12-02-copy-source-form.md) D-6, D-12, D-13, D-17 (the form this PR opens); S12-02 review M-9 (row menu shape); this plan's review M-1 (the kit menu modifier, user answer 2026-10-18: build it in this PR)
**Planned:** 2026-10-18 against `main` @ `817bad1`; refreshed 2026-10-18 against `main` @ `6ecfd88` after S12-02 landed (#485) and the user moved the kit menu modifier into this PR. Rows marked *(from S12-02)* name code S12-02 added; its **Landed** note changes nothing this plan uses
**Review:** [`S12-17-sources-list-copy.review.md`](S12-17-sources-list-copy.review.md) — round 1 (first plan): Fits with changes / Ready with changes → fixed; M-1 answered by the user (build the kit modifier here). Refresh 2026-10-18: Fits with changes / Ready with changes → fixed, no blocker

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

A researcher can copy a Source straight from the Sources list: right-click a row's page zone, choose Copy…, and the S12-02 Copy Source form opens for that Source. Saving opens the copy's page, as from the Source page. With S12-02 this completes the dogfood bar's Copy Source outcome ("Copy… on the Source page or a Sources list row's context menu").

The row menu is the app's third right-click menu. This PR adds two kit layers in one file. Menu items (`PVActionMenuItem`, drawn by `PVActionMenuRows`) are written once and give both the rows and the keyboard titles, with indexes derived from position. A right-click modifier, `pvRightClickMenu`, adds the menu's state, the right-click trigger and one VoiceOver action per item on top. The two existing right-click menus (the Source page cover, the Person page portrait) move onto `pvRightClickMenu`, and the Citation composer's button-opened Clear menu moves onto the items layer. The cover menu gains the VoiceOver action it lacks today.

**Why now.** S12-02 builds the form, its model and its presenter for any host. This PR is the second host. Each hand-wired menu names an item in parallel places (keyboard title, item and index, and for right-click menus the accessibility action), and the copies have already drifted (review M-1); a third copy would add a third drift.

**What stays.** The row's page and graph zones, its thumbnail, title, type, progress, Updated column and ref. The Copy form is unchanged. Add Source is unchanged. The cover and portrait menus keep their titles, items, hints, tooltips and accessibility identifiers. `PVSelect` stays as it is (it seeds its own highlight). The Clear menu keeps its button, its own state and `dismissOnClickAway`.

## 2. User stories

**US-1.** As a researcher on the Sources list, when I right-click a row and choose Copy…, I get the Copy Source form for that Source so that I can start a sibling Source without opening the original.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Right-click (or Ctrl-click) the page zone of "Obituary — Mary Alderwick" | A kit menu with one item, Copy…, opens at the pointer |
| AC-1.2 | Choose Copy… | The Copy Source form opens with `Copy of Obituary — Mary Alderwick` and type Obituary; the parts area loads (S12-02 AC-1.7), then shows the parts and Artifacts as on the Source page |
| AC-1.3 | Copy | The copy's page opens; Back to the Sources list shows the copy as a row, and the sidebar Sources count includes it (S12-02 T-9 and the existing status hook, D-9; manual check step 5) |
| AC-1.4 | Left-click the page zone; left-click the graph zone | The Source page opens; the Evidence graph opens — as today |
| AC-1.5 | Right-click the graph zone | No menu opens |
| AC-1.6 | VoiceOver on the row's page zone, actions rotor | Copy… is listed and opens the same form |
| AC-1.7 | Menu open; arrow Down, Return; or Escape | Return chooses Copy…; Escape closes the menu |

**US-2.** As a researcher using the Source page cover, the Person page portrait or the Citation composer Clear menu, the menus work as before, and with VoiceOver I can reach the cover's action too.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Source page whose cover is an Artifact; right-click the cover thumbnail | The menu titled "Source thumbnail" (`L10n.Sources.thumbnailMenuTitle`) opens at the pointer with Revert to default; choosing it sets the cover back to the type icon — as today |
| AC-2.2 | Source page whose cover is the type icon; right-click the cover | The menu opens with the disabled "Default icon is in use" row (`thumbnailDefaultInUse`); arrow keys highlight nothing — as today |
| AC-2.3 | VoiceOver on the cover with an Artifact cover, actions rotor | Revert to default is listed and does the same as AC-2.1 (new). With the type-icon cover, no action is listed |
| AC-2.4 | Person page whose lead portrait is not the earliest; right-click the portrait; then VoiceOver actions rotor | The menu opens at the pointer with Use earliest likeness; the rotor lists the same action — as today |
| AC-2.5 | Either menu open; arrow Down, Return; or Escape | Return chooses the highlighted item; Escape closes — as today |
| AC-2.6 | Citation composer with a drawn region; click Clear; arrow Down twice, Return | The menu shows Clear region and Reset to entire artifact; Return on the second resets the locator; with no region, Clear region is disabled — as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| Kit `PVActionMenuItem`, `PVActionMenuRows` and `View.pvRightClickMenu(title:accessibilityIdentifier:items:)` (D-2, D-3); the kit doc comments (D-15) | STEP-2, T-2, T-3 |
| The Citation composer Clear menu moves onto `PVActionMenuItem` / `PVActionMenuRows` (D-14) | STEP-3, AC-2.6 |
| The cover menu and the portrait menu move onto `pvRightClickMenu` (D-4, D-5) | STEP-3, AC-2.1 – AC-2.5 |
| `SourcesModel.copying` / `openCopy(_:)` (D-1) | STEP-1, AC-1.2, AC-1.3 |
| Row menu on `SourcesSplitRow.pageZone` through `pvRightClickMenu` (S12-D6 CL-1, CL-3 – CL-5; D-6) | STEP-4, AC-1.1, AC-1.4 – AC-1.7 |
| Both row call sites pass `onCopy`; the list hosts `.sourceCopyDialog` (CL-2; D-7) | STEP-5, AC-1.2, AC-1.3 |
| DesignSystem README row for ContextMenu and its "action menus" sentence (D-13) | STEP-6 |
| Map row and S12-17 scope entry name the kit layers | STEP-6 |

| Out | Home |
| --- | --- |
| Other row menu items (Delete, Open graph) | not planned (S12-D6 §2.1) |
| A row hover or overflow (⋯) button, multi-select copy | not planned (S12-D6 §2.1) |
| Opening a right-click menu from the keyboard (Shift-F10 or the Menu key) and an "Actions for <title>" accessibility label on the panel, both drawn in the web board (D-10) | not planned — user answer 2026-10-18 ("leave both out") |
| Moving `PVSelect` onto `PVActionMenuItem` | not planned — `PVSelect` seeds its own highlight from the committed value and draws checkmark rows (`pvContextMenu` doc comment: it leaves `keyboardTitles` nil) |
| Moving the Clear menu onto `pvRightClickMenu` | not planned — it opens from a button, so it needs neither the trigger nor VoiceOver actions (D-14) |
| One row call site instead of `SourcesSplitRowProgress` | not planned — the struct exists so one Source's progress merge repaints only its row (its doc comment, `SourcesListView.swift` L415); review M-2 declined |
| Any change to the Copy form, its model or its strings | not planned (S12-02 owns them) |
| The graph accordion's sub-rows | S12-14 (they do not get this menu, D-6) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Kit menu | `macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift`: `PVContextMenuState` (L7), `PVContextMenuKeyboard` (L32, `.inactive`, `.opened(titles:)`), `PVContextMenuPanel` (L83: `title`, `width` 180, `accessibilityIdentifier`), `PVContextMenuItem` (L121: `init(_ title: LocalizedStringResource, index:isEnabled:…accessibilityIdentifier:action:)` L136), `pvContextMenuTrigger(_:)` (L260, an overlay `PVRightClickCatcher`), `pvContextMenu(_:keyboard:keyboardTitles:…content:)` (L280; presents the panel in a child `NSPanel` positioned from the host view's own screen frame, `PVContextMenuPopupWindow`), `PVRightClickCatcherView.hitTest` (L742) returns itself only for right-click and Ctrl-click | Every right-click caller wires state, keyboard, titles, indexes and the VoiceOver action by hand |
| Cover menu | `macos/App/Features/Sources/SourcePageIdentityHeader.swift`: `@State coverMenu`, `coverKeyboard` (L10–11), `coverMenuItemCount` (L15–17); `.pvContextMenu($coverMenu, keyboard: count > 0 ? … : nil, keyboardTitles: …) { coverMenuPanel }` on the header `HStack` (L32–39, with the comment "Menu on the row (not the 72pt thumb)…"); `coverThumbnail` (L55–68: `.accessibilityHint`, `.accessibilityIdentifier("sources.page.cover")`, `.pvContextMenuTrigger($coverMenu)`, `.help`); `coverMenuPanel` (L70–86: title `thumbnailMenuTitle`, id `sources.page.cover.menu`; `thumbnailDefaultInUse` disabled when `coverMode == "type_icon"`, else `thumbnailRevertToDefault`, index 0, id `sources.page.cover.revert`, `Task { await model.artifacts.revertCoverToTypeIcon() }`) | No VoiceOver action for Revert to default |
| Portrait menu | `macos/App/Features/Conclusions/ConclusionDetailPage.swift`: `@State portraitMenu`, `portraitKeyboard` (L365–366), `showsPortraitMenu` (L369); `.pvContextMenu($portraitMenu, keyboard:, keyboardTitles: [useEarliestLikeness]) { PVContextMenuPanel(title: portraitMenuTitle, id "persons.detail.portrait.menu") { PVContextMenuItem(useEarliestLikeness, index: 0, id "persons.detail.portrait.useEarliest") { Task { await portrait?.useEarliest() } } } }` on the header row (L397–415); `thumbnail` (L417–440): when `showsPortraitMenu`, `thumb.pvContextMenuTrigger(…).help(…).accessibilityHint(…).accessibilityAction(named: useEarliestLikeness) { … }.accessibilityIdentifier("persons.detail.portrait")` | The one menu with a VoiceOver action |
| Clear menu | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` `@State clearMenu`, `clearKeyboard` (L6–7); `clearMenuButton` (L217–252): a `PVButton` with `.pvContextMenu($clearMenu, keyboard: $clearKeyboard, keyboardTitles: [clearRegion, resetToEntireArtifact], dismissOnClickAway: true)`, panel id `citationComposer.clearMenu.menu`, items `clearRegion` (index 0, `isEnabled: model.locator.hasRegion && model.locatorCapabilities.supportsRegionLocator`, id `citationComposer.clearRegion`, `model.clearRegion()`) and `resetToEntireArtifact` (index 1, id `citationComposer.resetToEntireArtifact`, `model.resetToEntireArtifact()`); `toggleClearMenu()` (L255) | Titles and indexes kept in parallel by hand |
| Kit doc comments | `PVContextMenu.swift` `PVContextMenuInteraction` (L21, "Action-menu defaults…"); `pvContextMenu` doc (L267–274: "Attach this to the view whose local coordinates match `state.origin` (the trigger for `PVSelect`, or a larger ancestor for right-click menus…)"); `pvContextMenuTrigger` doc (L258–259); `macos/App/DesignSystem/README.md` L363–364 ("not `PVContextMenu` action menus (click-to-open overflow / right-click)") | The ancestor advice is the overlay-era rule D-3 retires |
| Why the menus sit on the row | Commit `71f32c35` "Present PVContextMenu in a popUpMenu NSPanel like ComboBox." changed the comment from "so overflow hits register" (an overlay-era reason) to "so click origin shares the panel's coordinate space" | Since `71f32c35` the panel is a child window placed from the host's screen frame, so a trigger and panel on the same view share coordinates (D-3) |
| Row | `macos/App/Features/Sources/SourcesSplitRow.swift` `SourcesSplitRow` (L29): `source`, `typeLabel`, `typeIconKey`, `projectDir`, `progress`, `countsLoading`, `onOpenPage` (L36), `onOpenGraph` (L37); `body` = `SourcesSplitLayout.columns { pageZone } graph: { graphZone }`, `.accessibilityElement(children: .contain)`; `pageZone` (L54–100) = `Button(action: onOpenPage)` with `.accessibilityLabel(L10n.Sources.openSourcePage)` (L93), `.accessibilityValue` (L94), id `sources.row.<id>.page` (L99); `graphZone` (L103–) | No menu, no `onCopy` |
| Row call sites | `macos/App/Features/Sources/SourcesListView.swift` `SourcesListContent` (L56) `ForEach(rows)` (L309–330; the `SourcesSplitRow(` call without a progress handle at L314) and `private struct SourcesSplitRowProgress` (L416–440) | Each builds `onOpenPage` / `onOpenGraph` with `navigation.go(to:)` |
| List dialog chain | `SourcesListView.swift` L87–131: `.vocabularyToastOverlay` (L87), `.pvFormDialog` (Add Source, L88–110), `.task(id: model.isAdding)`, `.onChange(of: sourcesHandle.status)` → `model.syncCatalogCounts()` on ready (L115–117) | The status hook republishes the sidebar count whenever the list reloads |
| List model | `macos/App/Features/Sources/SourcesModel.swift` `SourcesModel` init `(session:userID:store:catalogCounts:settingsStore:)`, private `userID`, `store`, `session` (L96–98); `isAdding` (L85) / `openAdd()` (L351) | No copy state |
| Copy model *(from S12-02)* | `macos/App/Features/Sources/SourceCopyModel.swift` `SourceCopyModel.init(original:session:userID:store:)` (L43) | Loads the original's workspace itself; host-agnostic |
| Copy presenter *(from S12-02)* | `macos/App/Features/Sources/SourceCopyDialog.swift` `View.sourceCopyDialog(_ model: Binding<SourceCopyModel?>)` (L7); reads `WorkspaceNavigation` from the environment | Closes and opens the copy's page itself |
| Page host *(from S12-02)* | `macos/App/Features/Sources/SourcePageModel.swift` `copying` (L19) / `openCopy()` (L168) | The shape to copy |
| Entry string *(from S12-02)* | `macos/App/Platform/L10n.swift` `L10n.Sources.copySourceEntry` (L5181) = "Copy…" (`sources.copy.entry`) | Reused here |
| String resolve | `macos/App/Platform/L10n.swift` `L10n.string(_ resource: LocalizedStringResource) -> String` (L20); kit files already call `L10n` (`PVFormDialog.swift`, `PVToast.swift`) | How the modifier turns item titles into keyboard titles |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` `photoType(id:)` (L11), `source(id:title:typeID:ref:updatedRevision:)` (L25), `makeModel(store:sources:types:catalogCounts:settingsStore:) -> (SourcesModel, WorkspaceSession)` (L46); `macos/ProvenenciaTests/SourcePageModelTests.swift` `openCopyUsesThePageSource` (L74); `macos/ProvenenciaTests/PVContextMenuTests.swift` suites `PVContextMenuStateTests` (L7), `PVContextMenuKeyboardTests` (L39) | No copy tests on the list model; no action-menu tests |
| Xcode project | `macos/Provenencia.xcodeproj/project.pbxproj` lists every Swift file (`PVContextMenu.swift`: build file `AA00000000000000000003C1`, file ref `AA00000000000000000003C0`, in group `D5C000000000000000000012 /* ContextMenu */` (L1516, entry L1519), sources entry L2198); `AA00000000000000000FE005` and `AA00000000000000000FE006` are unused | A new file needs a build file, a file ref, a group entry and a sources entry |
| DesignSystem README | `macos/App/DesignSystem/README.md` "What's built vs. not yet" table (L146–) | No ContextMenu row |

**Precedent.** The portrait menu (`ConclusionDetailPage.swift` L397–440) is the shape `pvRightClickMenu` packages: trigger, panel with a title, indexed items, keyboard titles, and an `.accessibilityAction(named:)` per item. The host side copies S12-02's `SourcePageModel.openCopy()` and `.sourceCopyDialog($model.copying)`. A new kit file registered in the Xcode project: `PVSkeletonBar.swift` in S12-02 (#485, commit `86a47614` fixed a group id that collided with another file; pick ids no other line uses).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `SourcesModel.copying: SourceCopyModel?` *(new)* and `func openCopy(_ source: CatalogSource)` *(new)*: `copying = SourceCopyModel(original: source, session: session, userID: userID, store: store)` | Same pair as `SourcePageModel.copying` / `openCopy()` | `@State` in the view |
| D-2 | New file `macos/App/DesignSystem/Components/ContextMenu/PVActionMenu.swift` *(new)*, registered in the Xcode project. It holds two layers. **Items:** `struct PVActionMenuItem` *(new)* with `let title: LocalizedStringResource`, `var isEnabled: Bool = true`, `var accessibilityIdentifier: String? = nil`, `var action: () -> Void = {}` (synthesized memberwise init, no custom init), and `struct PVActionMenuRows: View` *(new)* with `let items: [PVActionMenuItem]`, whose body is `ForEach(Array(items.enumerated()), id: \.offset) { i, item in PVContextMenuItem(item.title, index: item.isEnabled ? i : nil, isEnabled: item.isEnabled, accessibilityIdentifier: item.accessibilityIdentifier, action: item.action) }`. **Right-click:** `extension View { func pvRightClickMenu(title: LocalizedStringResource? = nil, accessibilityIdentifier: String?, items: [PVActionMenuItem]) -> some View }` *(new)*, backed by `private struct PVRightClickMenuModifier: ViewModifier` that owns `@State menu = PVContextMenuState()` and `@State keyboard = PVContextMenuKeyboard.inactive`. Its body, in this order: `content.accessibilityActions { ForEach(Array(items.enumerated()).filter { $0.element.isEnabled }, id: \.offset) { _, item in Button(action: item.action) { Text(item.title) } } }`, `.pvContextMenuTrigger($menu)`, `.pvContextMenu($menu, keyboard: titles == nil ? nil : $keyboard, keyboardTitles: titles) { PVContextMenuPanel(title: title, accessibilityIdentifier: accessibilityIdentifier) { PVActionMenuRows(items: items) } }`, where `titles = PVActionMenuItem.keyboardTitles(items)` | One rule turns items into rows, titles and indexes for every action menu (review M-1, user answer; refresh review M-1); the right-click layer adds only what right-click menus need. Component layer in the ContextMenu family, as `.pvFormDialog` is for sheets (`add-ui-component` step 5). Named for its trigger, beside `PVRightClickCatcher`, because "action menu" already means every non-Select menu (README L363, refresh review M-3) | One `pvActionMenu` for right-click only; a recipe under `Recipes/`; a `PVContextMenuPanel(items:)` init (the panel is generic over `Content`, so it needs a constrained extension for no gain over a rows view) |
| D-3 | `static func keyboardTitles(_ items: [PVActionMenuItem]) -> [String]?` *(new)* on `PVActionMenuItem`: `nil` when no item is enabled, else every item's title through `L10n.string(_:)`, in order (disabled rows keep their slot so indexes match positions). The modifier puts the trigger and the panel on the view it is applied to; there is no placement option | `nil` keeps today's cover behavior (no keyboard when only "Default icon is in use" shows). Since `71f32c35` the panel is a child window placed from its host's screen frame, so the click point and panel share one view's coordinates; the row-level placement in the cover and portrait is the overlay-era workaround (§4) | A `trigger:` placement parameter; titles of enabled items only (indexes would stop matching rows) |
| D-4 | Cover: `SourcePageIdentityHeader` drops `coverMenu`, `coverKeyboard`, `coverMenuItemCount`, `coverMenuPanel`, and the header `HStack`'s `.pvContextMenu` with its comment. `coverThumbnail` replaces `.pvContextMenuTrigger($coverMenu)` with `.pvRightClickMenu(title: L10n.Sources.thumbnailMenuTitle, accessibilityIdentifier: "sources.page.cover.menu", items: coverMenuItems)`; `private var coverMenuItems: [PVActionMenuItem]` *(new)* is `[PVActionMenuItem(title: L10n.Sources.thumbnailDefaultInUse, isEnabled: false)]` when `model.source?.coverMode == "type_icon"`, else `[PVActionMenuItem(title: L10n.Sources.thumbnailRevertToDefault, accessibilityIdentifier: "sources.page.cover.revert") { Task { await model.artifacts.revertCoverToTypeIcon() } }]`. The hint, help and `sources.page.cover` id stay where they are | Same items and ids; gains the VoiceOver action (AC-2.3) | Leaving the cover hand-wired |
| D-5 | Portrait: `ConclusionDetailPage` drops `portraitMenu`, `portraitKeyboard`, and the header row's `.pvContextMenu` with its comment. In `thumbnail`'s `showsPortraitMenu` branch, `.pvContextMenuTrigger($portraitMenu)` and `.accessibilityAction(named:)` are replaced by one `.pvRightClickMenu(title: L10n.Conclusions.portraitMenuTitle, accessibilityIdentifier: "persons.detail.portrait.menu", items: [PVActionMenuItem(title: L10n.Conclusions.useEarliestLikeness, accessibilityIdentifier: "persons.detail.portrait.useEarliest") { Task { await portrait?.useEarliest() } }])` in the trigger's place; `.help`, `.accessibilityHint` and `persons.detail.portrait` stay | Same items and ids | Leaving the portrait hand-wired |
| D-6 | Row: `SourcesSplitRow` gains `var onCopy: () -> Void` *(new)*, after `onOpenGraph`. `pageZone` gets, after its `.accessibilityIdentifier("sources.row.\(source.id).page")`, `.pvRightClickMenu(accessibilityIdentifier: "sources.row.\(source.id).menu", items: [PVActionMenuItem(title: L10n.Sources.copySourceEntry, accessibilityIdentifier: "sources.row.\(source.id).copy", action: onCopy)])`. No title (one item needs no heading; S12-D6 §2: title optional). The menu is on `pageZone` only, so the graph zone and S12-14's sub-rows get none | CL-1, CL-3 – CL-5; S12-02 review M-9 | The whole row; a native `.contextMenu`; a "Source" heading |
| D-7 | Both call sites pass `onCopy: { model.openCopy(source) }`. `SourcesSplitRowProgress` already holds `model`. `SourcesListContent` adds `.sourceCopyDialog($model.copying)` right after the Add Source `.pvFormDialog` | CL-2; one presenter (S12-02 D-13) | A per-host dialog |
| D-8 | No new strings: the row item and its VoiceOver action reuse `copySourceEntry`; the cover and portrait reuse theirs. No `.help` or `.accessibilityHint` on the page zone: it already has a label and value, and a hover tooltip on every row is noise | Same action, same words | A `sources.row.menuHint` string as the cover and portrait carry |
| D-9 | The sidebar count needs no code: after a copy the list reloads (`.createdSource`, S12-02 Landed note) and the existing `.onChange(of: sourcesHandle.status)` calls `syncCatalogCounts()` | Existing hook | Calling `syncCatalogCounts()` from the copy path |
| D-10 | The agreed board is a web prototype. It draws opening the menu with Shift-F10 or the Menu key and an `aria-label` "Actions for <title>" on the panel; neither is built (user answer 2026-10-18). Keyboard and VoiceOver users reach Copy… from the actions rotor (D-2). Everything else in the `rowmenu` and `copied` frames matches D-6, D-7: one item, Copy…, 180 pt wide (the panel's default `width`), at the pointer; Ctrl-click opens the menu and not the page; right-click on the graph zone opens nothing; the copy appears as a new row | The macOS kit has neither; the brief's CL-1 – CL-5 do not ask for them | A key monitor on the row; a panel accessibility label |
| D-11 | Tests: T-1 on `SourcesModel`; T-2, T-3 pure tests of `PVActionMenuItem.keyboardTitles` in `PVContextMenuTests.swift` (new suite `PVActionMenuItemTests`, no new test file). The prefill and the list refresh after a copy are S12-02's (its T-3, T-9). The menus, their placement and their VoiceOver actions are checked by hand (§10); no UI tests, as for the portrait menu today | One owner per rule (review M-3); a SwiftUI modifier's view tree is not unit-testable here | XCUITest for a right-click |
| D-12 | The Xcode build, tests and manual check run on the macOS dev machine (`scripts/check-pr.sh` `xcode-test` is macOS only). On Linux, stop after writing code and report | As S12-02 D-22 | Skipping them |
| D-13 | `macos/App/DesignSystem/README.md` "What's built" table gains one row after Select: `\| ContextMenu \| Components/ContextMenu/PVContextMenu.swift, PVActionMenu.swift (action menus are PVActionMenuItem lists: a right-click menu is one .pvRightClickMenu, which adds the state, the trigger and one VoiceOver action per item; a button-opened menu draws PVActionMenuRows inside .pvContextMenu, with keyboardTitles from PVActionMenuItem.keyboardTitles; PVSelect uses .pvContextMenu directly) \|`. L363–364's parenthesis becomes "(`PVActionMenuItem` lists, button-opened or `.pvRightClickMenu`)" | The kit's own index names the new layers, so the next menu finds them | A README in the folder |
| D-14 | Clear menu: `CitationComposerViewerPane` gains `private var clearMenuItems: [PVActionMenuItem]` *(new)*: `[PVActionMenuItem(title: L10n.CitationComposer.clearRegion, isEnabled: model.locator.hasRegion && model.locatorCapabilities.supportsRegionLocator, accessibilityIdentifier: "citationComposer.clearRegion") { model.clearRegion() }, PVActionMenuItem(title: L10n.CitationComposer.resetToEntireArtifact, accessibilityIdentifier: "citationComposer.resetToEntireArtifact") { model.resetToEntireArtifact() }]`. `clearMenuButton` keeps its `PVButton`, `$clearMenu`, `$clearKeyboard` and `dismissOnClickAway: true`; `keyboardTitles:` becomes `PVActionMenuItem.keyboardTitles(clearMenuItems)`, and the panel's content becomes `PVActionMenuRows(items: clearMenuItems)`. `toggleClearMenu()` is unchanged | No menu keeps parallel arrays (refresh review M-1); Reset is always enabled, so the titles are never nil | Moving it to `pvRightClickMenu` (Out) |
| D-15 | Kit comments: the `pvContextMenu` doc's "Attach this…" sentence becomes "Attach this to the view whose local coordinates match `state.origin`: the trigger view. Right-click menus use `.pvRightClickMenu`, which does this." The `pvContextMenuTrigger` doc gains "Feature code uses `.pvRightClickMenu`." `PVContextMenuInteraction`'s "Action-menu defaults" stays (it is right for both layers) | The kit's comments agree with D-3 (refresh review M-2) | Leaving the comments |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVActionMenu.swift`: `PVActionMenuItem`, `PVActionMenuRows`, `pvRightClickMenu(title:accessibilityIdentifier:items:)`, `PVRightClickMenuModifier` *(new)* | kit items + modifier | `pvContextMenu` + `pvContextMenuTrigger` (composed, not replaced); the portrait menu's hand wiring | It composes the two existing modifiers; extending `pvContextMenu` itself would push right-click-only concerns (trigger, VoiceOver actions) onto `PVSelect` and the Clear menu (D-2, Out) |
| `PVActionMenuItem.keyboardTitles(_:)` *(new)* | function | `PVContextMenuKeyboard.opened(titles:)` | It produces that function's input from items (D-3) |
| `SourcesModel.copying` / `openCopy(_:)` *(new)* | members | `SourcePageModel.copying` / `openCopy()` | Same shape on the second host (D-1) |
| `SourcesSplitRow.onCopy` *(new)* | member | `onOpenPage`, `onOpenGraph` | Same callback shape (D-6) |
| `SourcePageIdentityHeader.coverMenuItems` *(new)* | member | `coverMenuPanel` (deleted) | Replaces it (D-4) |
| `CitationComposerViewerPane.clearMenuItems` *(new)* | member | the inline items and `keyboardTitles` array | Replaces them (D-14) |
| T-1, T-2, T-3 *(new)* | tests | `openCopyUsesThePageSource`; `PVContextMenuKeyboardTests` | Same shapes |

**Builds on.** S12-02's `SourceCopyModel` and `sourceCopyDialog`, the kit `PVContextMenu`, `copySourceEntry`.
**Deletes.** The Clear menu's hand-kept `keyboardTitles` array and item indexes; `SourcePageIdentityHeader.coverMenu`, `coverKeyboard`, `coverMenuItemCount`, `coverMenuPanel` and its row-level `.pvContextMenu`; `ConclusionDetailPage.portraitMenu`, `portraitKeyboard`, its row-level `.pvContextMenu` and the thumb's `.accessibilityAction`.
**Next adds.** Another item in any action menu is one more `PVActionMenuItem`. Another right-click menu anywhere is one `.pvRightClickMenu` line; another button-opened menu is `PVActionMenuRows` inside `.pvContextMenu`. S12-14's accordion goes below `pageZone` and keeps the menu off its sub-rows.

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None — `CopySource` exists (S12-01).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Action menu items and right-click modifier | `PVActionMenuItem`, `PVActionMenuRows`, `pvRightClickMenu` *(new)* over `PVContextMenu` | `DesignSystem/Components/ContextMenu/PVActionMenu.swift` | D-2, D-3 |
| Clear menu | `PVActionMenuRows` in `pvContextMenu` | `Features/CitationComposer/CitationComposerViewerPane.swift` | D-14 |
| Cover menu | `pvRightClickMenu` | `Features/Sources/SourcePageIdentityHeader.swift` `coverThumbnail` | D-4 |
| Portrait menu | `pvRightClickMenu` | `Features/Conclusions/ConclusionDetailPage.swift` `thumbnail` | D-5 |
| Row Copy menu | `pvRightClickMenu` | `Features/Sources/SourcesSplitRow.swift` `pageZone` | D-6 |
| Copy Source form | `sourceCopyDialog` *(from S12-02)* | `Features/Sources/SourcesListView.swift` | D-7 |

Board frames, in words. `rowmenu`: the Sources list with the Suffolk Chronicle obituary row right-clicked on its page zone; a 180 pt kit panel at the pointer, no title row, one item "Copy…"; the row keeps its hover fill; nothing else changes. `copied`: after Copy, the list (sorted by Date added) shows the copy as the top row, with its own ref, "Today" in the Updated column, and the copied title. The form between them is S12-D1's (S12-02). States: the form's loading and errors are S12-02's. Accessibility: D-2's per-item actions. Differences from the board: D-10.

### Strings

None — every title exists (D-8).

## 7. Steps

The implementer evidences every step with XT and `scripts/check-pr.sh`. §10's manual checks need a person at the app (right-click, Ctrl-click, VoiceOver, where the panel opens); the implementer lists them as "for the user" and does not tick them.

Xcode check, used below as **XT**: `NSUnbufferedIO=YES xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/<Suite> CODE_SIGNING_ALLOWED=NO` (from `scripts/check-pr.sh`).

**STEP-1 — List model copy state**
Files: `macos/App/Features/Sources/SourcesModel.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`. Skill: `add-swift-test`.
Do: D-1. Add T-1.
Check: XT with `SourcesModelTests` passes.

**STEP-2 — Kit action menu**
Files: `macos/App/DesignSystem/Components/ContextMenu/PVActionMenu.swift` *(new)*, `macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift` (comments), `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/ProvenenciaTests/PVContextMenuTests.swift`. Skills: `add-ui-component`, `add-swift-test`.
Do: D-2, D-3, D-15 (D-15 edits `PVContextMenu.swift` comments only). Register the file next to `PVContextMenu.swift`: file ref `AA00000000000000000FE005`, build file `AA00000000000000000FE006`, in group `D5C000000000000000000012 /* ContextMenu */` and in the app target's Sources phase beside `AA00000000000000000003C1`; `grep -c` each id first and expect 0. Header comment: the two layers and when to use each. Add T-2, T-3. No call site yet.
Check: XT with `PVActionMenuItemTests` passes.

**STEP-3 — Move the cover, portrait and Clear menus**
Files: `macos/App/Features/Sources/SourcePageIdentityHeader.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift`. Skill: `add-ui-component`.
Do: D-4, D-5, D-14.
Check: XT with `SourcePageModelTests` passes (the app builds); `rg -n 'coverMenu\b|coverKeyboard|portraitMenu\b|portraitKeyboard|pvContextMenuTrigger' macos/App/Features` prints nothing. Manual check steps 8–11 are the user's.

**STEP-4 — Row menu**
Files: `macos/App/Features/Sources/SourcesSplitRow.swift`. Skill: `add-ui-component`.
Do: D-6, D-8. Until STEP-5, both call sites pass `onCopy: {}` so the build stays green.
Check: XT `SourcesModelTests` passes (the app builds).

**STEP-5 — List host**
Files: `macos/App/Features/Sources/SourcesListView.swift`.
Do: D-7, D-9. Replace STEP-4's `{}` placeholders.
Check: XT `SourcesModelTests` passes (the app builds). Manual check steps 1–7 are the user's.

**STEP-6 — Docs**
Files: `macos/App/DesignSystem/README.md` (D-13), `docs/deployment-plan/spike-12/completed.md` (a `## S12-17 — Copy Source from the Sources list` paragraph at the top, in the S12-02 style, naming `pvRightClickMenu`), `docs/deployment-plan/spike-12/deployment-plan.md` (map row L113: scope cell "Copy… in a Sources list row's context menu opens the S12-02 form; kit action menu items and `pvRightClickMenu`, with the cover, portrait and Clear menus moved onto them", status `Landed #N` with the open PR's number; the S12-17 scope entry's **Ships** (L231) gains "and the kit's `PVActionMenuItem` / `pvRightClickMenu`, with the Source cover, Person portrait and Citation composer Clear menus moved onto them"; every link to the brief, here and in this plan, points at `design/archive/`, and the brief's own relative links gain one `../`), this plan (status, Landed note), `docs/deployment-plan/spike-12/design/S12-D6-sources-list-copy.md` → `docs/deployment-plan/spike-12/design/archive/` (the README's Scheduled row moves to an Archived line, as S12-D1's did).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-17-sources-list-copy.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Copy a Source from a Sources list row's menu."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/SourcesModelTests.swift` `openCopySetsCopyingForThatSource` *(new)* | `let s1 = source(id: "s1", title: "Obit", typeID: "t1")`; `let (model, _) = makeModel(sources: [s1], types: [photoType()])`; before, `model.copying == nil`; `model.openCopy(s1)` → `model.copying?.original.id == "s1"` | AC-1.2, D-1 |
| T-2 | Swift | `macos/ProvenenciaTests/PVContextMenuTests.swift` `PVActionMenuItemTests.keyboardTitlesKeepEverySlot` *(new)* | `[PVActionMenuItem(title: "Clear region", isEnabled: false), PVActionMenuItem(title: "Reset")]` → `["Clear region", "Reset"]` (string-literal `LocalizedStringResource`s resolve to themselves in the test bundle) | D-3, AC-2.5 |
| T-3 | Swift | same file · `PVActionMenuItemTests.keyboardTitlesNilWhenNothingIsEnabled` *(new)* | `[PVActionMenuItem(title: "Default icon is in use", isEnabled: false)]` → `nil`; `[]` → `nil` | D-3, AC-2.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: the app build (the Clear menu's items, the new `onCopy` parameter at both row call sites, the removed `@State` in the cover and portrait), `SourcesModelTests`, `PVContextMenuTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none expected (D-8)
- [ ] UI matches the brief inventory; the new kit pieces are PVActionMenuItem, PVActionMenuRows and pvRightClickMenu (D-2), listed in the DesignSystem README (D-13)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Setup, in any dogfood project: Add Source "Obituary — Mary Alderwick", type Obituary; on its page add a description, two metadata values and a credibility grade, attach two image Artifacts, and pin one as the cover. For the portrait: a Person with two likenesses (depictions drawn on two Citations' pages) and the later one chosen as the portrait in the likeness gallery, so the header offers Use earliest likeness. For the Clear menu: a Citation on an image Artifact with a drawn region. Restart the app so the obituary's workspace is not cached.

1. Open the Sources list. Left-click the obituary's page zone: its page opens. Back. Left-click its graph zone: the Evidence graph opens. Back. (AC-1.4)
2. Right-click the graph zone: nothing opens. (AC-1.5)
3. Right-click the page zone: a menu with Copy… opens at the pointer. Press Escape: it closes. Right-click again, arrow Down, Return: the form opens. (AC-1.1, AC-1.7)
4. The form shows `Copy of Obituary — Mary Alderwick`, type Obituary; the parts area shows its loading state, then the three parts and both Artifacts checked. (AC-1.2)
5. Copy. The copy's page opens. Back: the copy is a row in the list, and the sidebar Sources count went up by one. (AC-1.3)
6. Ctrl-click a page zone: the same menu opens, and the page does not open. (AC-1.1)
7. With VoiceOver on a row's page zone, the actions rotor offers Copy…; choosing it opens the form. (AC-1.6)
8. Open the obituary's page. Right-click the cover thumbnail: the "Source thumbnail" menu opens at the pointer with Revert to default; arrow Down, Return reverts the cover to the type icon. Right-click again: the disabled "Default icon is in use" row shows; arrow Down highlights nothing; Escape closes. (AC-2.1, AC-2.2, AC-2.5)
9. Pin an Artifact as the cover again. With VoiceOver on the cover, the actions rotor offers Revert to default and it works; after it, the rotor offers no action. (AC-2.3)
10. Open the Person. Right-click the portrait: its menu opens at the pointer; Use earliest likeness works from the menu. Choose the later likeness again in the gallery; with VoiceOver on the portrait, the actions rotor offers Use earliest likeness and it works. (AC-2.4, AC-2.5)
11. Open the Citation's composer. Click Clear: the menu shows Clear region and Reset to entire artifact; arrow Down twice, Return resets the locator. Click Clear again: Clear region is disabled. (AC-2.6)

This and S12-02 make the slice 1 Check true from both entry points.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The right-click catcher blocks left clicks on the page button | `PVRightClickCatcherView.hitTest` returns nil for plain left clicks; manual check step 1 |
| A menu panel opens away from the pointer now that the cover and portrait panels sit on the thumbnail | D-3's reasoning (`71f32c35`); manual check steps 3, 8, 10 |
| `.accessibilityActions` on the portrait thumbnail stops listing the action while `.accessibilityHint` and the id come after it | Same order as today's single `.accessibilityAction`; manual check step 10 |
| A row's `@State` menu resets when the lazy list recycles the row | Acceptable: a recycled row is off screen, so its menu is closed |

**Stop and report if:**

- `SourceCopyModel.init`, `sourceCopyDialog` or `L10n.Sources.copySourceEntry` differs from §4 at your base.
- `.accessibilityActions { … }` with a `ForEach` does not compile against the macOS 14 deployment target, or lists no action in the rotor (D-2).
- A menu on the cover or portrait thumbnail opens visibly away from the pointer (D-3 is then wrong; report the offset you see).
- `pvRightClickMenu` on `pageZone` stops the page button opening on a left click.
- You are on a machine without Xcode (D-12).

## 12. Open questions

None. M-1 and the board's two web-only details were answered by the user on 2026-10-18 (D-2, D-10). D-14 follows from that answer: the user asked for the hand-wired menus to be refactored in this PR.

## 13. PR

**Title:** S12-17 — Copy Source from the Sources list
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-17-sources-list-copy.md`.
