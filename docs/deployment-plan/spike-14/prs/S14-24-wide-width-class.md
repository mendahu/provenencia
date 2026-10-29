# S14-24 — Give every workspace page its width class

**Status:** Planned
**Spike / slice:** Spike 14 · Slice 6 — Wide layout
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a token, an environment value and a measurement at the workspace host; nothing on screen changes)
**Board:** —
**Depends on:** —
**Blocks:** S14-25
**Branch:** `s14-24-wide-width-class`
**Version bump:** none — no researcher-visible change; the first page that reads the width class (S14-25 and later) takes the bump ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`
**Context:** [`docs/macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide), [`add-design-brief` Wide layout](../../../../.agents/skills/add-design-brief/SKILL.md#wide-layout-binding), [S14-D1](../design/S14-D1-wide-layout.md) (the kit side of the same token), working rule "(Batch 6) One breakpoint, 1600pt of pane width"
**Planned:** 2026-10-29 against `main` @ `b8ca4cfc`
**Review:** [`S14-24-wide-width-class.review.md`](S14-24-wide-width-class.review.md) — Fits with changes — Ready with changes, 2 rounds; To user answered 2026-10-29

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

Every workspace page can tell whether it has a Regular pane (below 1600pt) or a Wide one (1600pt and up), so a page designed for a large monitor reads one value instead of measuring itself. The user set the breakpoint on 2026-10-29 because pages built for the 1366pt default window sit jammed to one side of a large monitor with empty space beside them ([§ 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide)). This PR adds the token, the width class and the one measurement at the workspace host. No page reads it yet.

**Why now.** The user asked for the utility to be in place before any Wide frame is designed, so S14-D1's board and every later view brief build on a value that already exists (2026-10-29).

**What stays.** Every page's layout and look at every window size. `widthContentMax`, the gutters, the window's default and minimum sizes. `SourcesListColumns`, which picks its own columns from its own width.

## 2. User stories

**US-1.** As a contributor building a Wide frame, when I write a workspace page, I read `@Environment(\.pvWidthClass)` and get `.regular` or `.wide` for the pane the page is in, so I do not measure it myself.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Pane width 1102 (default window, sidebar open), 1464 (16" MacBook Pro full screen), 1599 | `PVWidthClass(paneWidth:)` is `.regular` |
| AC-1.2 | Pane width 1600, 1656 (1920 monitor), 2296 (2560 monitor) | `PVWidthClass(paneWidth:)` is `.wide` |
| AC-1.3 | Pane width 0 (first layout pass) | `.regular` |
| AC-1.4 | A view with `.pvMeasuresWidthClass()` hosted 2000pt wide | A child reading `\.pvWidthClass` sees `.wide`; the same view hosted 1200pt wide sees `.regular` |
| AC-1.5 | A child of `WorkspaceDestinationHost` hosted at 2000pt wide | Reads `.wide` without measuring anything itself |
| AC-1.6 | A destination whose root hugs its content (an empty `else` branch, a short `Text`) in a 2000pt host | The host still reads `.wide`: it measures the space it is given, not its child |

**US-2.** As a researcher, nothing I see changes.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The app at the default window and stretched to 2560pt, on the Sources list, a Source page and a Person page | Each page looks as it does on `main` |
| AC-2.2 | Resizing the window across a 1600pt pane on a Source page | The scroll position and focus stay where they were; the Source page does not reload |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PVSpacing.widthWideBreakpoint` = 1600 | STEP-1, AC-1.2, D-1 |
| `PVWidthClass` with `init(paneWidth:)`, the `pvWidthClass` environment value and the `pvMeasuresWidthClass()` modifier, in a new kit folder | STEP-2, AC-1.1 – AC-1.4, D-2, D-3, D-4 |
| `WorkspaceDestinationHost` fills the space it is given and measures once, outside the place's `.id` | STEP-3, AC-1.5, AC-1.6, AC-2.2, D-5 |
| One test helper hosts a view in a borderless window; the three existing copies move onto it | STEP-4, D-9 |
| Kit README row and token note; § 9 of `macos-client-patterns.md` names the real symbols | STEP-5, D-7 |
| Swift tests for the class, the modifier and a filling measurement | STEP-4, T-1 – T-3 |

| Out | Home |
| --- | --- |
| The Page columns component and any page reading the width class | S14-25, gated by [S14-D1](../design/S14-D1-wide-layout.md) |
| A Wide frame for any page (Source page, Person page, lists, Promote, composer) | Each view's next brief, per `add-design-brief` "Wide layout (binding)" |
| Moving `SourcesListColumns` onto the width class | not planned: it picks columns at its own widths, not at the breakpoint |
| The CSS token in the Claude Design kit | S14-D1 WL-1 |
| A compact class or a third breakpoint | not planned |
| `CitationComposerView.composeSplit`'s own 1400pt `threeUp` breakpoint moving onto `pvWidthClass` | The composer's next brief ([`ideas/wide-layout.md`](../../../ideas/wide-layout.md)); the user chose to leave it, 2026-10-29, because moving it here changes the composer between 1400 and 1599pt |
| An animation or transaction on a class flip | not planned |
| A Wide arrangement for the pages that already have a side pane (Properties, Metadata, Source types) | Each page's Wide brief; listed in [`ideas/wide-layout.md`](../../../ideas/wide-layout.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Layout tokens | `macos/App/DesignSystem/Tokens/PVSpacing.swift` `PVSpacing` | `widthContentMax` 1180, `widthInspector` 340, `widthSidebar` 264, `gutterPage` 32, `gutterPageWide` 56 (unused), `widthWorkspaceDefault` 1366. No breakpoint |
| Host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `WorkspaceDestinationHost.body` | `Group { destination(for:) }.id(place?.viewIdentity)`: every destination mounts here and resets on its `viewIdentity` |
| Host's parent | `macos/App/Features/Workspace/WorkspaceContent.swift` | Places the host in a `VStack` under the toolbar in the detail area beside the sidebar. The host has no frame of its own, so its size is whatever the destination's root asks for: several branches render nothing (`if let sourceID …` with no else) |
| Width measured in a feature | `macos/App/Features/Sources/SourcesListView.swift` `onGeometryChange(for: SourcesListColumns.self)` | The idiom to copy: an `Equatable` value computed from `size.width`, so the action runs only when the value changes |
| Environment keys in the kit | `macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift` `PVContextMenuDismissKey` | `private struct …: EnvironmentKey` plus an `EnvironmentValues` extension. No `@Entry` in the repo |
| Kit README | `macos/App/DesignSystem/README.md` component table | One row per `Components/` folder; `scripts/check-guidance-names.py` fails on a folder with no row |
| Project file | `macos/Provenencia.xcodeproj/project.pbxproj` | Lists every Swift file by hand (S14-15 has not landed). An app file is five entries: `PBXBuildFile`, `PBXFileReference`, its folder group with the file as child, the group in the `Components` group's children, and the build file in the app's Sources phase (`PVFlowLayout.swift` and the `FlowLayout` group are the shape to copy). A test file is four, per `add-swift-test` |
| Pages with a side pane | `macos/App/Features/Properties/PropertiesView.swift` `PropertiesContent.body` (340pt inspector); `macos/App/Features/Metadata/MetadataView.swift` `MetadataContent.body` and `macos/App/Features/SourceTypes/SourceTypesView.swift` `SourceTypesContent.body` (380pt detail pane) | Each lays out a list and a fixed side pane in an `HStack`. D-6: they read the class of the whole page |
| Second breakpoint | `macos/App/Features/CitationComposer/CitationComposerView.swift` `composeSplit` | `GeometryReader` with `threeUp = geo.size.width >= 1400`: a private width switch at another number (Out) |
| Window mounts in tests | `macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift` `mountHost`; `macos/ProvenenciaTests/PVSelectPlacementTests.swift` `aLongClosedLabelStaysInsideTheCallerWidth`, `aShortClosedLabelHugsItsText` | Three hand copies of `NSHostingView` → frame → borderless `NSWindow` → `contentView` |
| Host tests | `macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift` `mountHost` | Mounts the host in a borderless `NSWindow` at 1200 × 800 with a `FakeStore`; `waitUntil` in `macos/ProvenenciaTests/QueryHandleWaiting.swift` |

**Precedent.** `e6b8b9d5` ("Share one flow layout and one dashed outline across the design system") added `macos/App/DesignSystem/Components/FlowLayout/PVFlowLayout.swift`, its README row and its five pbxproj entries in one commit. Copy its folder, README row and pbxproj shape. Do not copy its call-site rewrites: this PR adds no call site but the host.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PVSpacing.widthWideBreakpoint: CGFloat = 1600`, beside the other layout widths, with a doc comment that it is compared with the pane width, not the window or screen | The user's value (2026-10-29); § 9 has the reasoning | A window-width breakpoint — a collapsed sidebar changes the space a page has |
| D-2 | `enum PVWidthClass: Equatable, Sendable { case regular, wide }` with `init(paneWidth: CGFloat)`: `.wide` when `paneWidth >= PVSpacing.widthWideBreakpoint`, else `.regular` | One comparison, owned by the kit. `Equatable` so `onGeometryChange` fires only when the class changes | `UserInterfaceSizeClass` / `horizontalSizeClass` — macOS does not set it, and its meaning is iOS's |
| D-3 | The environment value is `EnvironmentValues.pvWidthClass` through a `private struct PVWidthClassKey: EnvironmentKey`, default `.regular` | Matches the kit's only environment-key idiom (`PVContextMenu`). `.regular` is what every page is designed for today, so a view outside the host is correct by default | `@Entry` — not used in the repo |
| D-4 | `View.pvMeasuresWidthClass()` is a `ViewModifier` with `@State var widthClass: PVWidthClass = .regular`; it applies `.environment(\.pvWidthClass, widthClass)` to its content and `onGeometryChange(for: PVWidthClass.self)` with `PVWidthClass(paneWidth: $0.size.width)` | One measurement serves every child. The geometry action runs only on a class change, so a resize within a class costs nothing | A `GeometryReader` wrapper — it takes all offered space and changes the host's layout |
| D-5 | `WorkspaceDestinationHost.body` applies `.frame(maxWidth: .infinity, maxHeight: .infinity)` and then `.pvMeasuresWidthClass()` after `.id(place?.viewIdentity)`. The frame keeps the default center alignment, so a destination root that fills is unchanged; manual checks 2 – 4 and a Stop condition catch one that relied on the host hugging it | Outside the `.id`, the measured class survives navigation, so a new page starts with the right class instead of `.regular` for one frame. The frame makes the host measure the pane, not its child (review round 2 X-9) | Inside the `.id`, or per destination — re-measures on every navigation and repeats the call 15 times. A `.background` probe — a second view for what a frame does |
| D-6 | The width class is the whole page's. A page with a fixed side pane (Properties, Metadata, Source types) reads the same class as every page and plans its whole arrangement, side pane included, in its own Wide frame. No page re-measures | One measurement and one definition: the CSS container query (S14-D1 WL-10), § 9 and the code agree on the pane. Round 1 re-measured Properties only; round 2 found two more side-pane pages and that a re-measured column also drops the gap and gutters, so it drifts from the rule (review M-5, M-6). **Changes the working rule; the user confirmed 2026-10-29: the whole inside view is what a page designs for, and an inspector is one page's own arrangement** | Re-measuring the column beside each side pane — three call sites, a fourth definition of pane. Subtracting pane widths at the host — the host does not know which destination has one |
| D-7 | New folder `macos/App/DesignSystem/Components/WidthClass/` holding `PVWidthClass.swift` (new) | A kit-wide, domain-free layout value, read by Page columns (S14-25) and by any page; layer **component** per `design-system-layers.md` | `Tokens/` — tokens are values transcribed from CSS, not types with behavior. `Platform/` — this is design-system, not process plumbing |
| D-9 | `func hostInBorderlessWindow<V: View>(_ view: V, windowSize: CGSize, hostSize: CGSize? = nil, orderFront: Bool = true) -> (NSWindow, NSHostingView<V>)` (new) in `macos/ProvenenciaTests/HostedWindow.swift` (new): it makes the borderless window at `windowSize`, sets the hosting view as its content view, sets the hosting view's frame to `hostSize ?? windowSize` after that, and orders the window front when asked. Callers hold the returned window for the test's length. `mountHost` calls it with 1200 × 800; the two `PVSelectPlacementTests` cases with `windowSize` 900 × 200, `hostSize` 900 × 80 and `orderFront: false`, and keep their own `layoutSubtreeIfNeeded`, so they lay out as today | T-2 and T-3 would be the fourth and fifth copy, and S14-25 needs it again (review M-3) | Copying the mount into T-2 |
| D-10 | `PVWidthClass` is `Sendable`; `PVWidthClassKey.defaultValue` is a plain `static let` | A `Sendable` enum needs no `nonisolated(unsafe)` under complete concurrency checking | Copying `PVContextMenu`'s key annotations |
| D-8 | The CSS token follows from S14-D1 WL-1; the README's token note says so | See **Rules this plan bends** | Waiting for the board before adding the Swift value — the user asked for the utility first |

**Rules this plan bends.**

- The map's working rule "(Batch 6) One breakpoint, 1600pt of pane width" measured the pane "minus any inspector". D-6 measures the whole page. The rule, § 9, the `add-design-brief` Paste block and S14-D1 were reworded on the planning branch to match. The user confirmed on 2026-10-29: "the whole inside view is what we're talking about here, and the inspector is just a specific implementation that those particular views have."
- `macos/App/DesignSystem/README.md` "Every value in these files is transcribed from the CSS, not invented": `widthWideBreakpoint` lands in Swift before `--width-wide-breakpoint` exists in the Claude Design kit. The user asked for the utility first and the brief alongside it (2026-10-29). STEP-5 adds a note to the README's token section that this one value was set in Swift first and S14-D1 WL-1 adds the CSS, and S14-25 checks the two agree.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVSpacing.widthWideBreakpoint` (new) | token | `PVSpacing.widthContentMax` | A different number; it sits beside it |
| `PVWidthClass` (new) | type | `SourcesListColumns`; `CitationComposerView` `composeSplit` (searched: `onGeometryChange`, `GeometryReader`, `ViewThatFits`, `horizontalSizeClass`) | `SourcesListColumns` is the Sources list's own column widths, not a kit-wide class. The composer's 1400pt switch is a second breakpoint; it moves with the composer's brief (Out, user 2026-10-29) |
| `pvWidthClass` environment value, `PVWidthClassKey` (new) | key | `PVContextMenuDismissKey` | Same idiom, different value |
| `pvMeasuresWidthClass()` (new) | modifier | `SourcesListView`'s `onGeometryChange` | That one sets feature state; this one sets the environment for every child |
| `Components/WidthClass/` (new) | folder | `Components/FlowLayout/` | Same shape: a domain-free layout piece |
| `PVWidthClassTests.swift` (new) | test | `macos/ProvenenciaTests/PVTableTests.swift` (a kit piece tested on its own) | — |
| `hostInBorderlessWindow` (new) | test helper | `mountHost` and the two `PVSelectPlacementTests` mounts | Replaces all three (D-9) |

**Builds on.** The kit's component folders and README table; `onGeometryChange` as `SourcesListView` uses it; the host's single mount point.
**Deletes.** The three hand-written window mounts in tests (D-9). `gutterPageWide` stays unused (S14-D1 keeps one gutter rule).
**Next adds.** S14-25's Page columns reads `pvWidthClass` and is the first reader. Every later Wide frame reads it; [`ideas/wide-layout.md`](../../../ideas/wide-layout.md) lists the pages waiting for one.

### Reach

None — no new path into existing code. Every destination gets an environment value it does not read.

### Scale

None — no write, list or loop changes. The geometry action runs only when the class flips, at most once per resize across 1600.

### Edges

- **Assumptions.** `>=` puts exactly 1600 at Wide (T-1). Width 0 and negative widths are Regular (T-1).
- **Neighbors.** 1599 and 1599.5 Regular, 1600 Wide (T-1).
- **Removed inputs.** None — the class has one input.
- **Real data.** None — no table or seed.
- **Untrusted size.** None — the width comes from layout, not from data.

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None — no wire change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Breakpoint | `PVSpacing.widthWideBreakpoint` (new) | `macos/App/DesignSystem/Tokens/PVSpacing.swift` | After `widthContentMax` |
| Width class | `PVWidthClass` (new) | `macos/App/DesignSystem/Components/WidthClass/PVWidthClass.swift` (new) | The enum, the key, the environment value and the modifier in one file |
| Measurement | `pvMeasuresWidthClass()` (new) | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` | After `.id` and a filling frame (D-5) |

```swift
enum PVWidthClass: Equatable, Sendable {
    case regular
    case wide
    init(paneWidth: CGFloat)
}

extension EnvironmentValues {
    var pvWidthClass: PVWidthClass { get set }
}

extension View {
    func pvMeasuresWidthClass() -> some View
}
```

States: none on screen. Accessibility: none — nothing visible changes.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Breakpoint token**
Files: `macos/App/DesignSystem/Tokens/PVSpacing.swift`.
Do: add `widthWideBreakpoint` = 1600 after `widthContentMax` with the D-1 doc comment.
Check: the app builds.

**STEP-2 — Width class**
Files: `macos/App/DesignSystem/Components/WidthClass/PVWidthClass.swift` (new), `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-ui-component`.
Do: write the enum, key, environment value and modifier per D-2 – D-4 and D-10. Doc comments: the enum says Regular is below `widthWideBreakpoint` and Wide at or above it, measured by `pvMeasuresWidthClass()`, with a link to § 9 of `docs/macos-client-patterns.md`; the modifier says it measures the view it is applied to, so that view must fill the space to be measured, and gives that class to the view's content. Register the file with the five entries in § 4 (Project file row), with new 24-hex IDs that appear nowhere else in the file (`rg` each before use).
Check: `xcodebuild -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' build` succeeds with no warning from the new file.

**STEP-3 — Measure at the host**
Files: `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`.
Do: after `.id(place?.viewIdentity)` in `body`, add `.frame(maxWidth: .infinity, maxHeight: .infinity)` and then `.pvMeasuresWidthClass()` (D-5), with the comment `// Fill the detail area, then measure it once: every destination reads \.pvWidthClass from here.`
Check: the app builds; `WorkspaceDestinationHostTests` pass unchanged.

**STEP-4 — Tests**
Files: `macos/ProvenenciaTests/PVWidthClassTests.swift` (new), `macos/ProvenenciaTests/HostedWindow.swift` (new), `macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift`, `macos/ProvenenciaTests/PVSelectPlacementTests.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-swift-test`.
Do: add `hostInBorderlessWindow` (D-9) and move the three existing mounts onto it; their tests pass unchanged. Register the two new test files with four pbxproj entries each, with new 24-hex IDs checked with `rg` before use. Write T-1 – T-3 with `@MainActor` on the suite and `@Test(arguments:)` for T-1. The probe is a `View` that writes `@Environment(\.pvWidthClass)` into a reference box in `onAppear` and `onChange(of:)`. In T-2 the probe has `.frame(maxWidth: .infinity, maxHeight: .infinity)` inside the modifier's content; T-2 resizes with `window.setContentSize(NSSize(width: 1200, height: 400))`. Each T-3 case is its own mount. Wait with `waitUntil` at its default timeout.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes.

**STEP-5 — Docs**
Files: `macos/App/DesignSystem/README.md`, `docs/macos-client-patterns.md`, `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/deployment-plan.md` (map row), this plan (status).
Do: add a `WidthClass` row (new) to the README component table, after the `FlowLayout` row (`Components/WidthClass/PVWidthClass.swift`: the Regular / Wide class of the pane, measured once by `WorkspaceDestinationHost`; read `\.pvWidthClass`) and the D-8 note as one sentence under the Token file map table. In § 9 replace the paragraph that starts "The breakpoint becomes a token" and the sentence "Measure it with `onGeometryChange` on the destination's root" with: read `@Environment(\.pvWidthClass)`; `WorkspaceDestinationHost` fills the detail area and measures it once. Then do the Definition of done's docs lines.
Check: `python3 scripts/check-guidance-names.py` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PVWidthClassTests.swift` `classFollowsPaneWidth` (new) | Parameterized: −1, 0, 1102, 1464, 1599, 1599.5 → `.regular`; `PVSpacing.widthWideBreakpoint`, 1600, 1656, 2296 → `.wide` | AC-1.1, AC-1.2, AC-1.3, D-1, D-2 |
| T-2 | Swift | `macos/ProvenenciaTests/PVWidthClassTests.swift` `modifierFollowsTheWindow` (new) | Probe under `.pvMeasuresWidthClass()` in a 2000 × 400 window → waits for `.wide`; the same window resized to 1200 × 400 → waits for `.regular` (so `.regular` is a measurement, not the default) | AC-1.4, D-4 |
| T-3 | Swift | `macos/ProvenenciaTests/PVWidthClassTests.swift` `aFilledHostMeasuresTheSpaceNotTheChild` (new) | A probe with no frame of its own (a short `Text`), under `.frame(maxWidth: .infinity, maxHeight: .infinity).pvMeasuresWidthClass()` as D-5 applies them, in a 2000 × 400 window → waits for `.wide`. Separate mount: a probe with no modifier anywhere reads the default `.regular` | AC-1.6, D-3, D-5 |

AC-1.5 is checked by manual check 1 (the host applies the frame and modifier T-3 tests in that order); a probe cannot be placed inside a destination without changing it.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `WorkspaceDestinationHostTests` and `PVSelectPlacementTests` (their mounts move onto the helper).

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
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md § 9 names the real symbols
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Read `WorkspaceDestinationHost.body`: `.id(place?.viewIdentity)`, then `.frame(maxWidth: .infinity, maxHeight: .infinity)`, then `.pvMeasuresWidthClass()`, so every destination is inside its content (AC-1.5 with T-3).
2. Open a project at the default window. Visit the Sources list, a Source page and a Person page. Each looks as on `main` (AC-2.1).
3. Stretch the window across a 2560pt-wide monitor, or full screen on any display wider than 1864pt. The same three pages look as on `main` (AC-2.1).
4. On the Properties, Metadata and Source types pages at the same width, the list and side pane look as on `main`.
5. On a Source page scrolled halfway, drag the window edge back and forth across about 1864pt of window width (1600 pane plus the 264 sidebar). The page keeps its scroll position and does not show its loading state (AC-2.2).

Without a display wider than about 1864pt, steps 3 and 5 are reported as not run, and T-2 and T-3 are the evidence.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `onGeometryChange` on the host changes the host's layout | It reads geometry without proposing a size; the manual check compares pages with `main` |
| The class starts `.regular` for one frame on a Wide window at launch | Accepted: no page reads it yet. S14-25 checks whether its first frame flickers |

**Stop and report if:**

- `WorkspaceDestinationHost.body` no longer ends with `.id(place?.viewIdentity)`.
- The filling frame moves any destination at either width (a root that relied on the host hugging it).
- Any page looks different from `main` at either width in the manual check.
- T-2 or T-3 cannot observe the class through a hosted window without a `sleep` or a timeout over `waitUntil`'s default.
- `onGeometryChange` is unavailable at the macOS 14 deployment target.

## 12. Open questions

None.

## 13. PR

**Title:** S14-24 — Give every workspace page its width class
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-24-wide-width-class.md`.
