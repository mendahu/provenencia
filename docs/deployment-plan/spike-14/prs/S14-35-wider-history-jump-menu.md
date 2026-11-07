# S14-35 — Wider history jump menu

**Status:** Landed in #730 (2026-11-06)
**Spike / slice:** Spike 14 · Slice 7d — List and nav polish
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the map scopes Batch 7's other PRs as ungated; the panel keeps its rows, components and copy, and only its width changes)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-35-wider-history-jump-menu`
**Version bump:** patch — a researcher-visible layout bug fix in the Mac app ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md) (read only: the jump menu is part of navigation history; nothing in it changes)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 7) History jump menu is wider first"; bar outcome 30; [`ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) § History menu row labels (the dogfood note's "How it works today" predated S14-14: row leaves now come from `WorkspaceLocationLabels`, see §4)
**Planned:** 2026-11-06 against `main` @ `f48846c`
**Review:** [`S14-35-wider-history-jump-menu.review.md`](S14-35-wider-history-jump-menu.review.md) — Fits with changes — Ready with changes, 2 rounds; no open blocker

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

Long-pressing or right-clicking Back or Forward opens a menu of the places in history. The menu is 360pt wide, so a deep place's row ("Sources › Smith family 1871 census · Citation for Smith, p. 4") is cut at the end, and a stack of composer or graph places reads as identical rows. After this PR the menu is 600pt wide when the content column has room, so typical Source, graph and composer titles show in full, and it narrows to stay inside the content column on a small window.

**Why now.** Dogfood note 2026-10-29: "they're truncating the last part of the breadcrumb and so all the entries look the same." Batch 7 working rule "History jump menu is wider first."

**What stays.** The row's labels (`HistoryJumpMenuHost.rows`, `WorkspaceLocationLabels`), tail truncation, one line per row, the ref column, keyboard and dismiss behavior, and every other `PVContextMenuPanel` caller.

## 2. User stories

**US-1.** As a researcher with several Citation composers and Evidence graphs in my history, when I open the Back or Forward menu, I read each row's graph name and Citation in full so that I jump to the place I mean.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Content column 1102pt wide (default 1366pt window minus the 264pt sidebar); Back opened from a button at x = 12 | The panel is 600pt wide |
| AC-1.2 | A history row whose leaf is "Smith family 1871 census · Citation for Smith, p. 4" (51 characters) with ref `CIT-F4N2P`, at the default window size | The row shows the whole leaf with no ellipsis |
| AC-1.3 | Content column 436pt wide (700pt minimum window minus the sidebar); Back opened from x = 12 | The panel is 408pt wide (436 − 12 − 16) and its right edge stays inside the content column |
| AC-1.4 | Content column not measured yet (width 0) | The panel is 600pt wide |
| AC-1.5 | Any other `PVContextMenuPanel` (Select, action menus, composer viewer menu) | Its width is unchanged |

## 3. Scope

| In | Covered by |
| --- | --- |
| The history jump panel's preferred width becomes 600pt (`HistoryJumpMenuModel.panelWidth`) | STEP-1, STEP-3, AC-1.1, AC-1.2 |
| The panel's width is capped to the content column's width minus its anchor's x and a 16pt margin | STEP-1, STEP-2, STEP-3, AC-1.3, AC-1.4 |
| `WorkspaceContent` measures its width into its own `contentWidth` state and passes it to `HistoryJumpMenuHost` | STEP-2, STEP-3, AC-1.3 |
| Swift tests of the width rule | STEP-1, AC-1.1, AC-1.3, AC-1.4 |
| Version bump, docs | STEP-4, STEP-5 |

| Out | Home |
| --- | --- |
| Head or middle truncation of the leaf | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) § "History menu row labels", added in STEP-5 (working rule "History jump menu is wider first") |
| Multi-line rows | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) § "History menu row labels", added in STEP-5 (working rule "History jump menu is wider first") |
| Building row labels from the full `breadcrumbItems` (Source title before graph name) | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) § "History menu row labels", added in STEP-5 (working rule "History jump menu is wider first") |
| Changing `PVContextMenuPanel`'s default width or adding a `minWidth` to the kit | Not planned: the history menu is the only caller that needs it (D-4) |
| Capping the omnibar results panel to the content column (it already runs past the column's left edge at the minimum width) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) row "S14-35 plan"; `OmnibarResultsHost` is not touched |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Panel width | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `HistoryJumpMenuPanel.body` | `PVContextMenuPanel(width: 360, accessibilityIdentifier: "workspace.toolbar.jumpMenu")`. A literal, with no cap to the window. This is the bug |
| Row layout | same file, `HistoryJumpMenuPanel.body` row `HStack` | Glyph 15pt, section label (caption 13pt, muted, `lineLimit(1)`), "›", leaf (bodySmall 14pt medium, `lineLimit(1)`, `.truncationMode(.tail)`, `.layoutPriority(1)`), then `refLabel` (mono 11pt, `.fixedSize`). Row padding 9pt horizontal; panel padding `PVSpacing.space2` (4pt). About 200pt of a row is not leaf, so 360pt leaves about 160pt (roughly 22 characters) for the leaf |
| Row labels | same file, `HistoryJumpMenuHost.rows(for:navigation:names:)` | Leaf is `WorkspaceLocationLabels.labels(for:names:)`'s leaf, joined to its section with `L10n.Conclusions.pageSectionLeaf` ("%1$@ · %2$@"). On a graph surface the leaf is the graph's name and the section is the composer's "Citation for" label or the Promote label. Unchanged by this PR |
| Placement | same file, `HistoryJumpMenuHost.body` | `.offset(x: anchor.minX, y: anchor.maxY + 4)`, where `anchor` is the Back or Forward button's frame in `HistoryJumpMenuModel.contentCoordinateSpace`. Nothing keeps the panel's right edge inside the column |
| Model | same file, `HistoryJumpMenuModel` | `@Observable`; holds `side` and `anchors`. No width rule |
| Host | `macos/App/Features/Workspace/WorkspaceContent.swift` `WorkspaceContent` | Owns `@State` `jumpMenu` and `omnibarResults`; `body` names the coordinate space and overlays `HistoryJumpMenuHost(jumpMenu:)` and `OmnibarResultsHost` top-leading over the toolbar and destination. Does not measure its width |
| Kit panel | `macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift` `PVContextMenuPanel` | `var width: CGFloat = 180`, applied as `.frame(width: width, alignment: .leading)`. Unchanged |
| Window floor | `macos/App/DesignSystem/Tokens/PVSpacing.swift` `widthWindowMin` (700), `widthSidebar` (264), `widthWorkspaceDefault` (1366) | Smallest content column with the sidebar shown is 436pt |
| Tests | `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift` | Covers crumb and location labels; nothing covers the jump panel's width |

**Precedent.** `OmnibarResultsModel.panelWidth` (`macos/App/Features/Workspace/OmnibarResultsModel.swift`, 640pt) is the overlay panel beside this one in the same content column: a `static let` on the panel's model, read by its host. Copy that shape. Width measurement copies `PersonFamilySection` (`macos/App/Features/Conclusions/PersonFamilySection.swift`): `.onGeometryChange(for: CGFloat.self) { $0.size.width } action: { … }`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Preferred width is 600pt. | 600 − about 200pt of chrome leaves about 400pt, roughly 55 characters of 14pt medium text: a graph name plus its "Citation for" label fits (AC-1.2's 51-character leaf). It stays below the omnibar results panel (640pt), the other overlay in this column. | 480pt — leaves about 280pt (≈38 characters), still cuts a composer row; 640pt — no gain on typical titles, and the two overlays would read as the same element |
| D-2 | The width is `min(600, contentWidth − anchorMinX − 16)` when `contentWidth > 0`, else 600. Computed by `HistoryJumpMenuModel.fittedPanelWidth(contentWidth:anchorMinX:)` *(new)*, a `static` pure function. | At the 700pt minimum window a fixed 600pt panel opened from Forward would run off the window's right edge. A pure function is testable without a view. 16pt is `PVSpacing.space6`. | A fixed 600pt — overflows on small windows; `GeometryReader` inside `HistoryJumpMenuHost` — the overlay fills the column and a reader there is a second layout pass for one number |
| D-3 | `WorkspaceContent` owns the column width: `@State private var contentWidth: CGFloat = 0` *(new)*, set with `.onGeometryChange`, passed as `HistoryJumpMenuHost(jumpMenu:contentWidth:)` *(new argument)*. | The column's width is a fact about the column, and `WorkspaceContent` owns the column, its coordinate space and both overlay models. The omnibar cap (structural-debt row "S14-35 plan") reads the same value as one more argument. | A `contentWidth` on `HistoryJumpMenuModel` — the omnibar host would read another feature's model, or the value would have to move (round 2, M-7) |
| D-4 | `PVContextMenuPanel` is not changed. `HistoryJumpMenuPanel` takes `width: CGFloat` and passes it on. | The kit already takes a width; only this caller needs more. | A kit `minWidth` or intrinsic sizing — touches every menu for one caller |
| D-5 | No floor below the cap, and no `max(0, …)` either. | The narrowest column with the sidebar is 436pt, so the cap gives at least 408pt from Back, already wider than today's 360pt; from Forward (about 32pt further right) it gives about 376pt. With the sidebar hidden the column is wider. | A floor of 360 — would bring back the overflow it exists to avoid |
| D-6 | Truncation stays `.tail`, one line. | Working rule "History jump menu is wider first": truncation changes wait for dogfood. | `.middle` / `.head` — Out |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `HistoryJumpMenuModel.panelWidth` *(new)* `static let = 600` | constant | `OmnibarResultsModel.panelWidth` | Same shape, own value: the two panels are different widths |
| `HistoryJumpMenuModel.fittedPanelWidth(contentWidth:anchorMinX:)` *(new)* | func | searched `clamp`, `min(` in `macos/App/Features/Workspace` and `macos/App/DesignSystem/Components/ContextMenu`: `PVContextMenu.swift` clamps its `NSPanel` frame to the screen, not to a SwiftUI column | That clamp is for the NSPanel-hosted menus; this panel is a SwiftUI overlay |
| `WorkspaceContent.contentWidth` *(new)* `@State`, and `HistoryJumpMenuHost.contentWidth` *(new)* `let` | state + init argument | `WorkspaceContent.jumpMenu` / `omnibarResults` `@State`; `pvMeasuresWidthClass()` in `macos/App/DesignSystem/Components/WidthClass/PVWidthClass.swift` | The view that owns the column holds it, beside its two overlay models. `pvMeasuresWidthClass()` measures the pane at `WorkspaceDestinationHost`, below the toolbar, and publishes a class, not points; its environment value cannot reach `WorkspaceContent`'s overlays. The `@State` carries a doc comment: overlay placement only, not a width-class source (`docs/macos-client-patterns.md` §9: pages read `\.pvWidthClass`) |
| `HistoryJumpMenuModelTests` *(new)* in `macos/ProvenenciaTests/HistoryJumpMenuModelTests.swift` *(new)* | test file | `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift` | That suite is about labels; a width suite keeps names honest. Synchronized folders (S14-15) add the file to the target with no project edit |

**Builds on.** The overlay-panel pattern in `WorkspaceContent` (`HistoryJumpMenuHost`, `OmnibarResultsHost`), `OmnibarResultsModel.panelWidth`, `PVContextMenuPanel(width:)`.
**Deletes.** The literal `360` in `HistoryJumpMenuPanel`. The dogfood note, once its directions 2–4 are in `docs/ideas/shared-list-helpers.md` (direction 1 shipped). Nothing else becomes dead.
**Next adds.** A later truncation or label change (Out) edits `HistoryJumpMenuPanel`'s row only; the width rule stays.

### Reach

None — no new path into existing code: the menu opens from the same controls and lists the same rows.

### Scale

None — no write, list or loop changes. Rows are capped at 15 (`NavigationHistoryStore.backJumpItems(limit:)`); one extra width read per layout.

### Edges

- **Assumptions.** `anchor.minX` and `contentWidth` are in the same space: both come from `WorkspaceContent`'s frame (`contentCoordinateSpace` is named on that view, and `.onGeometryChange` reads its size). T-1 holds the arithmetic.
- **Neighbors.** Column width exactly `600 + anchorMinX + 16` → 600 (T-1 row "exact fit"); one point less → 599.
- **Removed inputs.** `contentWidth == 0` before first layout → 600 (AC-1.4, T-1).
- **Real data.** None — no table or seed drives the rule.
- **Untrusted size.** None — titles are already capped to one line by `lineLimit(1)`; a very long title still truncates at the tail.

### Schema

None — no catalog change.

### Go core

None — Swift-only layout fix.

### FFI

None — no RPC changes.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| History jump panel | `PVContextMenuPanel` (unchanged) | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `HistoryJumpMenuPanel` | Takes `width: CGFloat` *(new parameter)*; passes it as `PVContextMenuPanel(width:)` |
| Width rule | — | `WorkspaceToolbar.swift` `HistoryJumpMenuModel` | `static let panelWidth: CGFloat = 600` (the preferred width); `static func fittedPanelWidth(contentWidth: CGFloat, anchorMinX: CGFloat) -> CGFloat` (the width drawn) |
| Host | — | `WorkspaceToolbar.swift` `HistoryJumpMenuHost` | New stored `let contentWidth: CGFloat` after `jumpMenu`; `body` passes `HistoryJumpMenuModel.fittedPanelWidth(contentWidth: contentWidth, anchorMinX: anchor.minX)` |
| Measure | — | `macos/App/Features/Workspace/WorkspaceContent.swift` `WorkspaceContent.body` | `@State private var contentWidth: CGFloat = 0` beside `jumpMenu`, with the doc comment `/// The content column's width, for placing overlays only. Pages read \.pvWidthClass.`; `.onGeometryChange(for: CGFloat.self) { $0.size.width } action: { contentWidth = $0 }` after `.frame(maxWidth: .infinity, maxHeight: .infinity)` and `.background(PVColor.surfacePage)`, immediately before `.coordinateSpace(name:)`, so it reads the full column width; the overlay becomes `HistoryJumpMenuHost(jumpMenu: jumpMenu, contentWidth: contentWidth)` |

Sketch:

```swift
static let panelWidth: CGFloat = 600

static func fittedPanelWidth(contentWidth: CGFloat, anchorMinX: CGFloat) -> CGFloat {
    guard contentWidth > 0 else { return Self.panelWidth }
    return min(Self.panelWidth, contentWidth - anchorMinX - PVSpacing.space6)
}
```

States: empty — the panel is not shown (unchanged); a column narrower than the panel — the panel narrows (D-2). Accessibility: unchanged (identifier `workspace.toolbar.jumpMenu`, label `L10n.Workspace.jumpMenuAccessibilityLabel`).

### Strings

None — no copy changes; the menu still covers the same history entries.

## 7. Steps

**STEP-1 — Width rule and its tests**
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/ProvenenciaTests/HistoryJumpMenuModelTests.swift` *(new)*. Skill: `add-swift-test`.
Do: add `panelWidth` and `fittedPanelWidth(contentWidth:anchorMinX:)` to `HistoryJumpMenuModel` per §6. Write T-1.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -only-testing:ProvenenciaTests/HistoryJumpMenuModelTests` passes (the scheme and flags `scripts/check-pr.sh` uses).

**STEP-2 — Measure the content column and pass it to the host**
Files: `macos/App/Features/Workspace/WorkspaceContent.swift`, `macos/App/Features/Workspace/WorkspaceToolbar.swift` (`HistoryJumpMenuHost` gains `let contentWidth: CGFloat`).
Do: add the `contentWidth` `@State`, the `.onGeometryChange` and the new host argument per §6 (after `.frame(maxWidth:maxHeight:)` and `.background`, immediately before `.coordinateSpace(name:)`; not on the bare `VStack`, which can be narrower than the column).
Check: the app builds with no new warning.

**STEP-3 — Panel reads the width**
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift`.
Do: give `HistoryJumpMenuPanel` a `width: CGFloat` property and pass it to `PVContextMenuPanel(width:)` in place of `360`; in `HistoryJumpMenuHost.body` pass `HistoryJumpMenuModel.fittedPanelWidth(contentWidth: contentWidth, anchorMinX: anchor.minX)`.
Check: `rg -n "width: 360" macos/App/Features/Workspace` prints nothing; manual check §10 steps 1–3.

**STEP-4 — Version bump**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (all four `MARKETING_VERSION` lines).
Do: read `VERSION` on `main` when the branch is cut and add one to its patch number (`.cursor/rules/versioning.mdc`): 0.3.7 → 0.3.8 at the planning base.
Check: `VERSION`, `core.Version` and all four `MARKETING_VERSION` lines show the same string.

**STEP-5 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/deployment-plan.md` (map row), this plan (status), `docs/deployment-plan/spike-14/ux/2026-10-29-nav-history-menu-truncates-breadcrumb.md` (deleted), `docs/ideas/shared-list-helpers.md`.
Do: per the Definition of done. This step's commit comes after the PR exists, because it needs `#N` and the date. Map PR table row: Plan column links this file, Status `Landed #N`. `completed.md`: a new section at the top, `## S14-35 — Wider history jump menu`, then `**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N).` and two or three sentences on what shipped, ending with the version, in the shape of the S14-34 entry. Scope entry `#### S14-35`: after "without tail truncation" add ", capped to the content column on a small window", and replace the clause about a test or snapshot with "Swift test of the width rule." In `docs/ideas/shared-list-helpers.md`, add a section `## History menu row labels` directly after `## The history menu on the shared rows`: one sentence of source ("Dogfood 2026-10-29; S14-35 widened the panel to 600pt and left these."), the verbatim annoyance quote from the ux note, and the note's fix directions 2–4 as three bullets, with every code pointer in them checked against `main` (the row leaf is `WorkspaceLocationLabels.labels(for:names:)` joined by `L10n.Conclusions.pageSectionLeaf`, at a width from `HistoryJumpMenuModel.fittedPanelWidth`). Its `docs/ideas/README.md` line already says "history menu rows"; leave it. Then delete the ux note. In the map's "Where the notes went", the S14-35 row's first cell becomes "Dogfood 2026-10-29: history jump menu truncates breadcrumb tails (note deleted YYYY-MM-DD; landed #N; leftovers in shared-list-helpers.md)", with that file name linked to `../../ideas/shared-list-helpers.md`. Point the S14-35 scope entry's **Context.** and this plan's **Context:** header at that section in place of the ux note.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-35-wider-history-jump-menu.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/HistoryJumpMenuModelTests.swift` *(new)* `panelWidthFitsTheColumn` *(new, parameterized)* | `@Suite struct HistoryJumpMenuModelTests` (no `@MainActor`: the function is a static pure function); `@Test(arguments: [(CGFloat, CGFloat, CGFloat)])` over (contentWidth, anchorMinX, expected); `#expect(HistoryJumpMenuModel.fittedPanelWidth(contentWidth:anchorMinX:) == expected)`. Cases: (1102, 12) → 600; (436, 12) → 408; (0, 12) → 600; exact fit (628, 12) → 600; (627, 12) → 599 | AC-1.1, AC-1.3, AC-1.4, D-1, D-2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `WorkspaceToolbarBreadcrumbTests` (same file under test).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: T-1 for AC-1.1, AC-1.3, AC-1.4; manual check for AC-1.2, AC-1.5 (run by the user or orchestrator, §10)
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief; PVContextMenuPanel unchanged)
- [ ] Version bump: patch (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Fix directions 2–4 in docs/ideas/shared-list-helpers.md § History menu row labels; ux note deleted; no link to the ux path left (rg)
- [ ] Domain doc updated if a rule in docs/*.md changed: none
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The user or the `ship-pr` orchestrator runs these steps on a Mac; the implementing agent does not, and reports them as not run. The PR description carries two screenshots of the open Back menu: one at 1366pt (step 1) and one at the minimum width (step 3).

1. Open a project with one Source that has two Evidence graphs, each with a long name ("Smith family 1871 census", "Smith family 1881 census"). In each graph, open the Citation composer on a Citation titled "Smith, p. 4", then return to the canvas. Window at its default size (1366pt). Long-press Back. Each composer row reads its graph name and "Citation for Smith, p. 4" in full; the two rows differ (AC-1.2, bar outcome 30).
2. Right-click Forward after going Back twice. The panel opens under Forward, 600pt wide, and its rows read in full.
3. Shrink the window to its minimum width with the sidebar shown. Long-press Back, then Forward. The panel's right edge stays inside the window; long rows end in an ellipsis (AC-1.3).
4. Open a `PVSelect` (a Source's type) and a Source row's action menu. They are as wide as before (AC-1.5).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `.onGeometryChange` on `WorkspaceContent` reports a width in a different space from `anchors` | Both come from the same view's frame; §10 step 3 checks the right edge |
| 600pt still cuts a typical row | §10 step 1 is the check; see Stop |

**Stop and report if:**

- `HistoryJumpMenuPanel` no longer passes `width: 360`, or `HistoryJumpMenuHost` no longer offsets by `anchor.minX` (§4 has drifted).
- §10 step 1's rows still end in an ellipsis at 600pt (D-1 is wrong; report the row text and the leaf's rendered width). Do not pick another width: there is no fallback width in this plan.
- The panel's right edge leaves the window at the minimum size in §10 step 3 (D-2 or D-3 is wrong).
- Passing the new width needs a change to `PVContextMenuPanel` (D-4).

## 12. Open questions

None.

## 13. PR

**Title:** S14-35 — Wider history jump menu
**Body:** objective in one sentence; what shipped against US-1; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-35-wider-history-jump-menu.md`.
