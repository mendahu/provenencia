# S12-17 — Copy Source from the Sources list

**Status:** Blocked — on S12-02 (not landed), and on board S12-D6
**Spike / slice:** Spike 12 · Slice 1 — Copy Source
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D6](../design/S12-D6-sources-list-copy.md), board not agreed yet
**Board:** — (set when the board is agreed, the same way as S12-D1's)
**Depends on:** S12-02 ([plan](S12-02-copy-source-form.md), not landed), S12-D6
**Blocks:** —
**Branch:** `s12-17-sources-list-copy`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** —
**Context:** [Copy Source](../source-structure.md#copy-source), working rule "Entry points"; [S12-02 plan](S12-02-copy-source-form.md) D-6, D-12, D-13, D-17 (the form this PR opens); S12-02 review M-9 (row menu shape)
**Planned:** 2026-10-18 against `main` @ `817bad1`, with S12-02 not yet built. Rows marked *(from S12-02)* describe what the S12-02 plan adds; when S12-02 lands, read its **Landed** note and refresh this plan before implementing
**Review:** [`S12-17-sources-list-copy.review.md`](S12-17-sources-list-copy.review.md) — Fits with changes / Ready with changes → fixed, 1 round; one item To user

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

**Why now.** S12-02 builds the form, its model and its presenter for any host. This PR is the second host: one model pair, one row menu and one modifier line.

**What stays.** The row's page and graph zones, its thumbnail, title, type, progress, Updated column and ref. The Copy form is unchanged. Add Source is unchanged.

## 2. User stories

**US-1.** As a researcher on the Sources list, when I right-click a row and choose Copy…, I get the Copy Source form for that Source so that I can start a sibling Source without opening the original.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Right-click (or Ctrl-click) the page zone of "Obituary — Mary Alderwick" | A kit menu with one item, Copy…, opens at the pointer |
| AC-1.2 | Choose Copy… | The Copy Source form opens with `Copy of Obituary — Mary Alderwick` and type Obituary; the parts area loads (S12-02 AC-1.7), then shows the parts and Artifacts as on the Source page |
| AC-1.3 | Copy | The copy's page opens; Back to the Sources list shows the copy as a row, and the sidebar Sources count includes it (S12-02 T-9 and the existing status hook, D-7; manual check step 5) |
| AC-1.4 | Left-click the page zone; left-click the graph zone | The Source page opens; the Evidence graph opens — as today |
| AC-1.5 | Right-click the graph zone | No menu opens |
| AC-1.6 | VoiceOver on the row's page zone, actions rotor | Copy… is listed and opens the same form |
| AC-1.7 | Menu open; arrow Down, Return; or Escape | Return chooses Copy…; Escape closes the menu |

## 3. Scope

| In | Covered by |
| --- | --- |
| `SourcesModel.copying` / `openCopy(_:)` (D-1) | STEP-1, AC-1.2, AC-1.3 |
| Row menu and accessibility action on `SourcesSplitRow.pageZone` (S12-D6 CL-1, CL-3 – CL-5) | STEP-2, AC-1.1, AC-1.4 – AC-1.7 |
| Both row call sites pass `onCopy`; the list hosts `.sourceCopyDialog` (CL-2) | STEP-3, AC-1.2, AC-1.3 |
| The kit menu modifier idea logged (review M-1, To user) | STEP-4 |

| Out | Home |
| --- | --- |
| Other row menu items (Delete, Open graph) | not planned (S12-D6 §2.1) |
| A row hover or overflow (⋯) button, multi-select copy | not planned (S12-D6 §2.1) |
| A kit `pvActionMenu(items:)` modifier in `DesignSystem/Components/ContextMenu/` over the hand-wired right-click menus, which also gives the cover menu its missing VoiceOver action | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) (STEP-4 adds the line) — or a follow-up PR if the user asks (review M-1) |
| One row call site instead of `SourcesSplitRowProgress` | not planned — the struct exists so one Source's progress merge repaints only its row (its doc comment, `SourcesListView.swift` L459); review M-2 declined |
| Any change to the Copy form, its model or its strings | not planned (S12-02 owns them) |
| The graph accordion's sub-rows | S12-14 (they do not get this menu, D-2) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Row | `macos/App/Features/Sources/SourcesSplitRow.swift` `SourcesSplitRow` (L29): `source`, `typeLabel`, `typeIconKey`, `projectDir`, `progress`, `countsLoading`, `onOpenPage`, `onOpenGraph`; `body` = `SourcesSplitLayout.columns { pageZone } graph: { graphZone }`, `.accessibilityElement(children: .contain)`; `pageZone` = `Button(action: onOpenPage)` with `.accessibilityLabel(L10n.Sources.openSourcePage)`, id `sources.row.<id>.page` (L54–100); `graphZone` (L102–) | No menu, no `onCopy` |
| Row call sites | `macos/App/Features/Sources/SourcesListView.swift` `SourcesListContent` `ForEach(rows)` (L313–334, no progress handle) and `private struct SourcesSplitRowProgress` (L460–484) | Each builds `onOpenPage` / `onOpenGraph` with `navigation.go(to:)` |
| List dialog chain | `SourcesListView.swift` L84–130: `.vocabularyToastOverlay`, `.pvFormDialog` (Add Source), `.task`, `.onChange(of: sourcesHandle.status)` → `model.syncCatalogCounts()` on ready | The status hook republishes the sidebar count whenever the list reloads |
| List model | `macos/App/Features/Sources/SourcesModel.swift` `SourcesModel` init `(session:userID:store:catalogCounts:settingsStore:)`, private `userID`, `store`, `session`; `isAdding` / `openAdd()` | No copy state |
| Copy model *(from S12-02)* | `macos/App/Features/Sources/SourceCopyModel.swift` *(new)* in S12-02 `SourceCopyModel(original:session:userID:store:)` (S12-02 D-6) | Loads the original's workspace itself; host-agnostic |
| Copy presenter *(from S12-02)* | `macos/App/Features/Sources/SourceCopyDialog.swift` *(new)* in S12-02 `View.sourceCopyDialog(_: Binding<SourceCopyModel?>)` (S12-02 D-13) | Closes and opens the copy's page itself |
| Page host *(from S12-02)* | `SourcePageModel.copying` / `openCopy()` (S12-02 D-12); `SourcePageIdentityHeader.copySourceButton` (S12-02 D-17) | The shape to copy |
| Entry string *(from S12-02)* | `L10n.Sources.copySourceEntry` = "Copy…" (`sources.copy.entry`) | Reused here |
| Kit menu | `macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift` `pvContextMenuTrigger` (L260), `pvContextMenu(_:keyboard:keyboardTitles:...)` (L280), `PVContextMenuPanel` (L83, `title: LocalizedStringResource?`), `PVContextMenuItem` (L121); `PVRightClickCatcherView.hitTest` (L742) returns itself only for right-click and Ctrl-click | Left clicks pass through to the button underneath |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` `makeModel`, `waitForQuery` *(shared from S12-02 D-19)* | No copy tests on the list model |

**Precedent.** The `ConclusionDetailPage` portrait menu, the one menu with a VoiceOver action (`macos/App/Features/Conclusions/ConclusionDetailPage.swift` L397–437): `.pvContextMenuTrigger` on the target, `.pvContextMenu(..., keyboard:, keyboardTitles:) { PVContextMenuPanel { PVContextMenuItem(_, index: 0, accessibilityIdentifier:) { action } } }`, and `.accessibilityAction(named:)` with the same action. Copy its menu, item and accessibility action; this row has no title (D-3) and no hint (D-6). The host side copies S12-02's `SourcePageModel.openCopy()` and `.sourceCopyDialog($model.copying)`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `SourcesModel.copying: SourceCopyModel?` *(new)* and `func openCopy(_ source: CatalogSource)` *(new)*: `copying = SourceCopyModel(original: source, session: session, userID: userID, store: store)` | Same pair as S12-02 D-12 on the page model | `@State` in the view |
| D-2 | The menu lives on `pageZone` only: `SourcesSplitRow` gains `var onCopy: () -> Void`, `@State private var rowMenu = PVContextMenuState()`, `@State private var rowKeyboard = PVContextMenuKeyboard.inactive`. `pageZone` gets, appended after its last modifier `.accessibilityIdentifier("sources.row.\(source.id).page")`, in this order: `.accessibilityAction(named:)` (D-4), `.pvContextMenuTrigger($rowMenu)`, then `.pvContextMenu($rowMenu, keyboard: $rowKeyboard, keyboardTitles: [L10n.string(L10n.Sources.copySourceEntry)]) { PVContextMenuPanel(accessibilityIdentifier: "sources.row.\(source.id).menu") { PVContextMenuItem(L10n.Sources.copySourceEntry, index: 0, accessibilityIdentifier: "sources.row.\(source.id).copy") { onCopy() } } }` | CL-1, CL-3; trigger and panel on one view share a coordinate space; S12-14's graph sub-rows sit outside `pageZone` (S12-02 review M-9) | The whole row; a native `.contextMenu` |
| D-3 | The panel has no title (`title` nil): one item needs no heading | S12-D6 §2: title optional | A "Source" heading |
| D-4 | `.accessibilityAction(named: Text(L10n.Sources.copySourceEntry)) { onCopy() }` on the `pageZone` button, first in D-2's order | CL-4; portrait menu precedent | On the row container |
| D-5 | Both call sites pass `onCopy: { model.openCopy(source) }`. `SourcesSplitRowProgress` already holds `model`. `SourcesListContent` adds `.sourceCopyDialog($model.copying)` after the Add Source `.pvFormDialog` | CL-2; one presenter (S12-02 D-13) | A per-host dialog |
| D-6 | No new strings: the menu item and the accessibility action reuse `copySourceEntry`. No `.help` or `.accessibilityHint` on the page zone: it already has a label and value, and a hover tooltip on every row is noise | Same action, same words | A `sources.row.menuHint` string as the cover and portrait carry |
| D-7 | The sidebar count needs no code: after a copy the list reloads (`.createdSource` from S12-02 D-10) and the existing `.onChange(of: sourcesHandle.status)` calls `syncCatalogCounts()` | Existing hook | Calling `syncCatalogCounts()` from the copy path |
| D-8 | Tests: one Swift Testing model test on `SourcesModel` (T-1). The prefill and the list refresh after a copy are S12-02's (its T-3, T-9); the count publish is `createReturnsSourceAndPublishesCount`'s. The menu and accessibility action are checked by hand (§10); no UI tests, as for the portrait menu | One owner per rule (review M-3) | Re-testing S12-02's model here; XCUITest for a right-click |
| D-9 | The Xcode build, tests and manual check run on the macOS dev machine (`scripts/check-pr.sh` `xcode-test` is macOS only). On Linux, stop after writing code and report | As S12-02 D-22 | Skipping them |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourcesModel.copying` / `openCopy(_:)` *(new)* | members | `SourcePageModel.copying` / `openCopy()` *(from S12-02)* | Same shape on the second host (D-1) |
| `SourcesSplitRow.onCopy`, `rowMenu`, `rowKeyboard` *(new)* | members | `ConclusionDetailPage.portraitMenu` / `portraitKeyboard`, `SourcePageIdentityHeader.coverMenu` | The kit's menu, wired the same way (D-2). A kit modifier over the three is the user's call (review M-1) |
| T-1 *(new)* | test | S12-02 T-12 `openCopyUsesThePageSource` | Same shape on the list model |

**Builds on.** S12-02's `SourceCopyModel` and `sourceCopyDialog`, the kit `PVContextMenu`, `copySourceEntry`.
**Deletes.** None — nothing becomes dead.
**Next adds.** More row menu items are one `PVContextMenuItem` each, with the next `index` and title. S12-14's accordion goes below `pageZone` and keeps the menu off its sub-rows. A kit `pvActionMenu(items:)` modifier would replace the three hand-wired right-click menus.

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None — `CopySource` exists (S12-01).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Row Copy menu | `PVContextMenu` (`pvContextMenuTrigger`, `PVContextMenuPanel`, `PVContextMenuItem`) | `Features/Sources/SourcesSplitRow.swift` `pageZone` | D-2, D-3 |
| Row accessibility action | `.accessibilityAction(named:)` | `pageZone` | D-4 |
| Copy Source form | `sourceCopyDialog` *(from S12-02)* | `Features/Sources/SourcesListView.swift` | D-5 |

States: the form's loading and errors are S12-02's. Accessibility: D-4. When the board is agreed, replace this paragraph with the frames' differences from D-2 – D-5, if any.

### Strings

None — `L10n.Sources.copySourceEntry` from S12-02 (D-6).

## 7. Steps

Xcode check, used below as **XT**: `NSUnbufferedIO=YES xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/<Suite> CODE_SIGNING_ALLOWED=NO` (from `scripts/check-pr.sh`).

**STEP-1 — List model copy state**
Start only when S12-02's plan says `Landed` and the S12-D6 board is agreed (this plan's status is then `Planned`).
Files: `macos/App/Features/Sources/SourcesModel.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-8. Add T-1.
Check: XT with `SourcesModelTests` passes.

**STEP-2 — Row menu**
Files: `macos/App/Features/Sources/SourcesSplitRow.swift`. Skill: `add-ui-component`.
Do: D-2, D-3, D-4. Until STEP-3, both call sites pass `onCopy: {}` so the build stays green.
Check: the app and tests build (XT `SourcesModelTests`). The menu itself is checked in STEP-3's manual check.

**STEP-3 — List host**
Files: `macos/App/Features/Sources/SourcesListView.swift`.
Do: D-5, D-7. Replace STEP-2's `{}` placeholders.
Check: manual check steps 1–7.

**STEP-4 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row status `Landed #N`), `completed.md` (a `## S12-17 — Copy Source from the Sources list` paragraph at the top, in the S12-01 style), this plan (status), `docs/deployment-plan/spike-12/design/` (S12-D6 → `docs/deployment-plan/spike-12/design/archive/`, *(new)* in S12-02 if absent; the README's Scheduled row moves to an Archived line), `docs/ideas/shared-list-helpers.md` (one line, unless the user chose a follow-up PR: a kit `pvActionMenu(items:)` modifier in `DesignSystem/Components/ContextMenu/` that owns the menu and keyboard state, derives `keyboardTitles` and `index`, and adds one `.accessibilityAction` per item, over the right-click menus `SourcePageIdentityHeader.coverMenu` (which lacks a VoiceOver action today), `ConclusionDetailPage.portraitMenu` and the new `SourcesSplitRow.rowMenu`; it needs a `trigger` placement option, since the cover and portrait menus put the trigger and the panel on different views).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-17-sources-list-copy.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Copy a Source from a Sources list row's menu."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/SourcesModelTests.swift` `openCopySetsCopyingForThatSource` *(new)* | `let s1 = source(id: "s1", title: "Obit", typeID: "t1")`; `makeModel(sources: [s1], types: [photoType()])` (the suite's helpers); `model.openCopy(s1)` → `model.copying?.original.id == "s1"`; before, `model.copying == nil` | AC-1.2, D-1 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcesModelTests`, the Sources list build (the new `onCopy` parameter at both call sites).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none expected (D-6)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use the S12-02 manual-check project (an Obituary Source "Obituary — Mary Alderwick" with a description, two metadata values, credibility and two Artifacts). Restart the app first so the obituary's workspace is not cached.

1. Open the Sources list. Left-click the obituary's page zone: its page opens. Back. Left-click its graph zone: the Evidence graph opens. Back. (AC-1.4)
2. Right-click the graph zone: nothing opens. (AC-1.5)
3. Right-click the page zone: a menu with Copy… opens at the pointer. Press Escape: it closes. Right-click again, arrow Down, Return: the form opens. (AC-1.1, AC-1.7)
4. The form shows `Copy of Obituary — Mary Alderwick`, type Obituary; the parts area shows its loading state, then the three parts and both Artifacts checked. (AC-1.2)
5. Copy. The copy's page opens. Back: the copy is a row in the list, and the sidebar Sources count went up by one. (AC-1.3)
6. Ctrl-click a page zone: the same menu opens. (AC-1.1)
7. With VoiceOver on a row's page zone, the actions rotor offers Copy…; choosing it opens the form. (AC-1.6)

This and S12-02 make the slice 1 Check true from both entry points.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The right-click catcher blocks left clicks on the page button | `PVRightClickCatcherView.hitTest` returns nil for plain left clicks; manual check step 1 |
| The menu panel opens at the wrong point inside the scrolling list | Trigger and panel are on the same view (D-2), as the portrait and cover menus; manual check step 3 |
| A row's `@State` menu resets when the lazy list recycles the row | Acceptable: a recycled row is off screen, so its menu is closed |

**Stop and report if:**

- S12-02 landed with a different `SourceCopyModel` init, `sourceCopyDialog` signature or `copySourceEntry` name than §4 says (check its **Landed** note).
- The agreed S12-D6 board differs from D-2 – D-5 (more items, a title, a hover button). The plan is updated before code.
- `pvContextMenuTrigger` on `pageZone` stops the page button opening on a left click.
- You are on a machine without Xcode (D-9).

## 12. Open questions

None. The board-delivery question is shared with S12-02 and answered there.

## 13. PR

**Title:** S12-17 — Copy Source from the Sources list
**Body:** objective in one sentence; what shipped against US-1; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-17-sources-list-copy.md`.
