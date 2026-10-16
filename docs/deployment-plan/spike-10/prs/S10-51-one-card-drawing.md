# S10-51 — Draw Evidence cards in one place

**Status:** Planned
**Spike / slice:** Spike 10 · Slice 11 — One card drawing
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-51--draw-evidence-cards-in-one-place)
**Brief:** — (no new chrome: the ghost and every card look as they do today)
**Board:** —
**Depends on:** S10-44 ([#418](https://github.com/mendahu/provenencia/pull/418), ready, not merged)
**Blocks:** —
**Branch:** `s10-51-one-card-drawing`, cut from `s10-44-unpromote` @ `1c489b80`; the PR's base is `s10-44-unpromote` until #418 merges
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`macos/App/Features/GraphCanvas/README.md`](../../../../macos/App/Features/GraphCanvas/README.md), working rules "The full card stays at every zoom." and "Every card is mounted.", [#367](https://github.com/mendahu/provenencia/pull/367) (cards drawn in AppKit), S10-44 ship log follow-ups MNT-3, MNT-4, UX-2, CON-4 ([`S10-44-unpromote.ship.md`](S10-44-unpromote.ship.md))
**Planned:** 2026-10-16 against `s10-44-unpromote` @ `2768698` (`main` @ `98819f8` plus #418)
**Review:** [`S10-51-one-card-drawing.review.md`](S10-51-one-card-drawing.review.md) — Fits, Ready with changes (applied); 2 rounds

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

Since [#367](https://github.com/mendahu/provenencia/pull/367) the Evidence graph draws every subject and bridge card with `EvidenceAppKitCard`. The SwiftUI card views it replaced are still in the tree. They draw the placement ghost and feed the card layout tests, and their accessibility action list has already drifted from the live card's. This PR draws the ghost with the live card, points the layout tests at the live card, deletes the SwiftUI card views, and names the live card `EvidenceCardView`, so a card change is made once and the tests check the card researchers see.

**Why now.** S10-44 had to add the Unpromote footer to both cards, and its review found the two VoiceOver action lists already disagree (MNT-4, UX-2). `EvidenceCardLayoutTests.measuredLayoutMatchesThePaint` measures the SwiftUI card, and no test builds `EvidenceAppKitCard` at all.

**What stays.** Every card's paint, layout, hit targets, edges, VoiceOver label, actions and rotor entries. The ghost's look, opacity and position. The constant-based first-frame geometry (`contentHeight`, the `layout: nil` branches of `actionTargets`, `edgeFrame`, `contentFrame`). `EvidenceCardHeaderActionHits`. The SwiftUI base layer (grid, edges, rubber band, marquee, empty state) and its `GraphCanvasDocumentView.hostingView`.

## 2. User stories

**US-1.** As a contributor, when I change how an Evidence card looks or what it does, I change `EvidenceCardView` and the shared geometry only, so the change cannot miss a second copy.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | T-8's searches | No match. `EvidenceSubjectCard` and `EvidenceBridgeCard` are `enum`s with static members only |
| AC-1.2 | `GraphCanvasItem` | Has one view builder, `makeView`, and no SwiftUI `content` closure. `GraphCanvasItemLayer` creates no `NSHostingView` and does not import SwiftUI |
| AC-1.3 | Run `EvidenceCardLayoutTests` | Every case builds `EvidenceCardView`, reads its `layout`, renders it to a bitmap, and the painted card height equals `layout.size.height` within 0.5pt. Same nine subject cases and the bridge case as today |
| AC-1.4 | A filed card and an unfiled card | The only VoiceOver action list is `EvidenceCanvasCardContent.namedActions(for:)`. Unpromote is listed only on the filed card (MNT-4, UX-2) |

**US-2.** As a researcher, when I place, select, promote or open cards on the Evidence graph, nothing looks or behaves differently.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Arm the Person tool and hover the canvas | The ghost is an unfiled Person card at 62% opacity, centred on the hovered grid point as today, with no VoiceOver element and no effect on clicks |
| AC-2.2 | `EvidenceCardView(ghostOf: EvidenceCanvasCard.ghost(p, presentation: nil))` for an uncited Person `p` | Its `layout` equals `EvidenceCardView(card: EvidenceCanvasCard(placed: .subject(p, nil), text: EvidenceCardText(subject: p), state: .resting), actions: [])`'s; it is not an accessibility element and has no custom actions |
| AC-2.3 | A promoted card's Unpromote square, hovered | The fill is a `PVControlSize.sm.height` square and the glyph is `PVControlSize.sm.iconGlyphSize`, centred in the 36pt slot: today's 22pt and 13pt (MNT-3) |
| AC-2.4 | A filed, cited bridge card | Its action targets have no `unpromote` target (CON-4) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Rename `EvidenceAppKitCard` → `EvidenceCardView` and `appKitView(for:)` → `cardView(for:)` | STEP-1 |
| `EvidenceCanvasCard.State.resting`, `EvidenceCanvasCard.ghost(_:presentation:)`, `EvidenceCardView(ghostOf:)`, and a SwiftUI host for the ghost | STEP-2, AC-2.1, AC-2.2 |
| Layout tests measure and render `EvidenceCardView` | STEP-3, AC-1.3 |
| `GraphCanvasItem` builds one `NSView`; the item layer drops the SwiftUI path | STEP-4, AC-1.2 |
| Delete the SwiftUI card views, their layout reporting, the SwiftUI accessibility chain, the unused coordinate-space name and two dead metrics; correct the comments and README paragraphs that describe them | STEP-5, AC-1.1, AC-1.4 |
| Unpromote fill and glyph from `PVControlSize.sm` (S10-44 MNT-3); bridge-card case in the footer test (S10-44 CON-4) | STEP-6, AC-2.3, AC-2.4 |
| Map, completed, plan status | STEP-7 |

| Out | Home |
| --- | --- |
| Removing the constant-based first-frame geometry (`contentHeight`, `layout: nil` branches). The edge layer and `GraphCanvasEdgeGeometryTests` still read it | not planned |
| Moving the ghost so its top sits where the placed card's top lands (today it is centred on the grid point) | not planned |
| The other S10-44 follow-ups: CON-1, CON-2, CON-3, UX-1, CON-OUT-1, the 11 hand copies of `pvMicroCaps()` | [`S10-44-unpromote.ship.md`](S10-44-unpromote.ship.md) follow-ups |
| Drawing the base layer (grid, edges, rubber band, marquee) in AppKit | not planned |
| Renaming `EvidenceSubjectCard` / `EvidenceBridgeCard` to layout-only names | not planned (D-2) |
| Replacing the hand-written resting `State` in `EvidenceGraphModelTests.unpromoteVoiceOverActionFollowsFooterAction` with `.resting` | not planned (file not otherwise touched) |
| Removing `import SwiftUI` from files other than `GraphCanvasItemLayer.swift` | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Live card | `EvidenceAppKitCard.swift` (now `EvidenceCardView.swift` in `macos/App/Features/EvidenceGraph/`) `EvidenceAppKitCard` | `init(card:actions:)` measures the card into `layout`, makes itself a VoiceOver group with `actions`, label and identifier, and sizes to `layout.size` plus `GraphCanvasItemLayer.paintOutset` on each side; `intrinsicContentSize` returns that size. `draw(_:)` sets alpha 0.92 dragging, 0.76 uncited subject, else 1, then paints through the same `measure`. Header comment says "The placement ghost stays SwiftUI." `paintFooter` paints the Unpromote fill as a literal `22×22` rect and the glyph as `13×13`. Referenced by name in `EvidenceCanvasCard.swift` and 4 `project.pbxproj` lines |
| Card value | `EvidenceCanvasCard.swift` `EvidenceCanvasCard`, `EvidenceCanvasCard.State` | `State` holds `isSelected`, `isActivated`, `isConnectingFrom`, `isDragging`, `canCite`, `hoveredActionID`, `pressedActionID`. Built with live values in `EvidenceGraphView.cards` (`state(_:)`, line 1032) and with a hand-written all-false value in `EvidenceGraphModelTests.swift` lines 1615–1621. Type comment line 7 says AppKit "rebuilds a card's hosting view" |
| Card builder | `EvidenceCanvasCard.swift` `EvidenceCanvasCardContent` | Comment lines 57–60 say "Builds a card's hosted view … item content closures". `appKitView(for:)` builds the live card and records its layout. `view(for:)`, `subject(_:presentation:card:)` and `bridge(_:card:)` build the SwiftUI cards with a hand-written `.accessibilityAction` chain (Unpromote on every subject card). `namedActions(for:)` is the live list; `accessibilityActions(for:)` maps it |
| Subject geometry and SwiftUI view | `EvidenceSubjectCard.swift` `EvidenceSubjectCard` (`struct … : View`) | Header lines 3–7 call it a paint-only card. Lines 9–287 and 355–412: static metrics, ids, `promotedFooterParts`, `origin`, `edgeFrame`, `actionTargets`, `footerFrame`, `contentHeight`, `contentCenter`, `topLeadingOffset`, `accessibilityLabel`, `footerAccessibilityActionName`, `membershipName`. Among them, `shellPaddingBottom` (line 22) is read by nothing, and `promoteInset` (line 34) only by the chrome (line 621). Lines 289–353 and 414–417: instance properties, `body`, `ghost(placed:presentation:)`, `ghostOpacity`. Lines 420–780: `private struct EvidenceSubjectCardChrome`. Lines 782–835: `enum EvidenceCardHeaderActionHits`, read by both `actionTargets`. `documentCoordinateSpace` (line 10) is read only by `EvidenceGraphView.swift` line 874 `.coordinateSpace(name:)`, and nothing resolves that name |
| Bridge geometry and SwiftUI view | `EvidenceBridgeCard.swift` `EvidenceBridgeCard` (`struct … : View`) | Header lines 4–9 say "Paint-only". Lines 11–26 and 58–305: static metrics, ids, `contentFrame`, `contentHeight`, `extraObservations`, `bodyTextHeight`, `actionTargets`, `citationHitFrame`, `accessibilityLabel`. Lines 28–56: instance properties and `body`. Line 307 to the end: `private struct EvidenceBridgeCardChrome` |
| Row views | `EvidenceCitedPropertyRow.swift` | `EvidenceCitedPropertyRowStyle`, `EvidenceCitedPropertyRow`, `EvidenceCitedPropertyStack`, `EvidenceConflictBracket`. Used only by the two chrome structs. The live card paints rows itself from `EvidenceCitedPropertyMarks` |
| Layout | `EvidenceCardLayout.swift` | `EvidenceCardLayout` and `EvidenceCardLayoutStore` (live path). `EvidenceCardHitRegionKey`, `View.evidenceCardHitRegion(_:)` and `View.reportsEvidenceCardLayout(_:)` serve only the SwiftUI cards. Type comment says "measured from SwiftUI layout"; the `regions` comment (line 15) links ``View/evidenceCardHitRegion(_:)`` |
| Ghost | `EvidenceGraphView.swift` `EvidenceGraphDocumentBody.body` lines 856–863 | `EvidenceSubjectCard.ghost(placed:presentation:)` `.position(center − canvas.origin)`, centre from `EvidenceSubjectCard.contentCenter`. Below the item layer's cards. The ghost subject is `EvidenceGraphModel.ghostPlacedSubject()`: id `"ghost"`, `isCited: false`, no membership |
| Items | `EvidenceGraphView.swift` `pushItems(_:)` | `GraphCanvasItem(…, content: { content.view(for: card) }, appKitView: { content.appKitView(for: card) })` |
| Canvas items | `macos/App/Features/GraphCanvas/GraphCanvasItemLayer.swift` | Imports SwiftUI only for `AnyView`/`NSHostingView`. Comments at lines 6–9, 18–22, 41 and 120 and the `makeView` comment describe SwiftUI hosting views. `GraphCanvasItem.content: () -> AnyView` (required), `appKitView: (() -> NSView)?`. `update(_:)` reuses an `NSHostingView<AnyView>` when `appKitView` is nil; `makeView(_:)` prefers `appKitView`, sets `frame.size = fittingSize`, else hosts `padded(content())`. `place(_:)` offsets every view by `paintOutset` |
| Canvas document | `GraphCanvasDocumentView.swift` | Line 10 comment: "``items`` — one hosting view per card". `hostingView` (line 15) is the base layer and stays |
| Canvas docs | `macos/App/Features/GraphCanvas/README.md` | Rows 18–19 and 34–35, paragraphs from line 46 and line 63, and the "Derive hit rects from measured paint" paragraph (lines 74–79, naming `.evidenceCardHitRegion(_:)`) describe SwiftUI-hosted items |
| Layout tests | `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift` `placed(rows:…)`, `measure`, `measureSubject`, `measuredLayoutMatchesThePaint` (9 cases), `bridgeTargetsFollowThePaint` | Hosts the SwiftUI cards in an `NSHostingView`, waits for `onLayout` to settle, bitmaps it and measures the painted height down the centre column |
| Item tests | `macos/ProvenenciaTests/GraphCanvasItemLayerTests.swift` `item(_:at:model:raised:rotor:counter:)` | `content:` returns `AnyView(Color.red.frame(width: 100, height: 40))`. `placesItemsAtTheirOriginOutsetForPaint` expects `view.frame.width >= 100 + 2 * outset` |
| Footer tests | `macos/ProvenenciaTests/EvidenceSubjectCardFooterTests.swift` `membership`, `promotedFooterSplitsHandleAndUnpromote` | Checks two unpromoted subject cards for no `unpromote` target; no bridge |
| VoiceOver test | `EvidenceGraphModelTests.swift` `unpromoteVoiceOverActionFollowsFooterAction` | Reads `namedActions(for:)`. Unaffected |
| Bridge value | `macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift` `SourceGraphPlacedBridge` | `membership: CatalogSubjectMembership?` marks a filed bridge |
| Project | `macos/Provenencia.xcodeproj/project.pbxproj` | `EvidenceCitedPropertyRow.swift` and `EvidenceAppKitCard.swift` each have four entries (PBXBuildFile, PBXFileReference, group child, Sources phase) |

**Precedent.** [#367](https://github.com/mendahu/provenencia/pull/367) (`c6b2ef92`) moved the live card to AppKit and added `appKitView` beside `content`. This PR finishes that move. `1411c8e1` added the measured-layout tests; keep their cases and assertions, change only what they build.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The ghost is a construction mode of the live card. Today's `init(card:actions:)` becomes `private init(card:actions:isGhost:)` *(designated)*, which stores `private let isGhost` and runs today's accessibility setup (element, `.group` role, label, identifier, actions, selected) only when `isGhost` is false; when true it calls `setAccessibilityElement(false)` and nothing else. `convenience init(card:actions:)` passes `isGhost: false`; `convenience init(ghostOf card: EvidenceCanvasCard)` *(new)* passes `actions: []`, `isGhost: true`. `draw(_:)` uses alpha 0.62 when `isGhost`, before the dragging and uncited cases | One paint path; a ghost with VoiceOver actions or a dragging ghost cannot be built. 0.62 is today's ghost opacity (`EvidenceSubjectCard.ghost`), and today's chrome under it has no 0.76 | `State.isGhost`: an interaction field the item path never sets (M-3). `alphaValue` on the view: multiplies with 0.76. A cached bitmap `Image`: a second render path |
| D-2 | `EvidenceSubjectCard` and `EvidenceBridgeCard` become `@MainActor enum`s holding today's static members, except `shellPaddingBottom`, `promoteInset` and `documentCoordinateSpace` (deleted). Same files, same names | Dozens of call sites in the app and tests read these names; the statics are the shared geometry. `@MainActor` keeps the isolation the `View` conformance gave them, so `nonisolated` members and `L10n` calls compile as today | Renaming them to layout names: ~258 references for no behavior. Moving statics into `EvidenceCardView`: the placement, edge layer and model read them without a view |
| D-3 | `GraphCanvasItem` replaces `content` and `appKitView` with `var makeView: @MainActor () -> NSView` *(new, required)*. The returned view sizes itself through `intrinsicContentSize` (so `fittingSize` works) and already includes the `paintOutset` margin. `GraphCanvasItemLayer.makeView(_:)` keeps `view.frame.size = view.fittingSize`; it drops `padded(_:)` and the `NSHostingView` branches, and a changed model always rebuilds the view, as it does today for AppKit cards. `GraphCanvasItemLayer.swift` drops `import SwiftUI` | The Evidence graph is the only host, and it always passed `appKitView` | Keeping SwiftUI `content` as optional generic tooling: a path nothing uses |
| D-4 | `EvidenceGhostCard` *(new)* is a private `NSViewRepresentable` in `EvidenceGraphView.swift` that makes `EvidenceCardView(ghostOf: card)`; `updateNSView` does nothing. The base layer places it where the SwiftUI ghost is today with `.id(ghost.kind.rawValue)`, `.fixedSize()`, `.position(center − canvas.origin)`, `.allowsHitTesting(false)` and `.accessibilityHidden(true)`; `center` is `EvidenceSubjectCard.contentCenter(gridX:gridY:)`. A presentation change while one kind stays armed keeps the old ghost colours until the kind changes; that is accepted | The base layer keeps today's z-order (the ghost under the cards) and centring; the card's outset is symmetric, so centring the view centres the card. The ghost's text and colours depend only on the kind and its presentation | Pushing the ghost as a `GraphCanvasItem`: items sit above the base layer, so the ghost would cover cards, and `cards`/`pushItems` would rebuild on every hover-cell change (M-4) |
| D-5 | `EvidenceCanvasCard.State` gains `static let resting` *(new)*: every flag false, `canCite: true`, both ids nil. `EvidenceCanvasCard` gains `static func ghost(_ placed: SourceGraphPlacedSubject, presentation: CatalogSubjectTypePresentation?) -> EvidenceCanvasCard` *(new)*: `.subject(placed, presentation)`, `EvidenceCardText(subject: placed)`, `.resting`. The base layer and T-1/T-2 use the factory; STEP-3's tests use `.resting` | One home for the ghost recipe and the resting state (M-1) | Building them inline at each site: five copies to update when `State` grows |
| D-6 | Layout tests build `EvidenceCardView(card:actions: [])`, read `view.layout`, make an offscreen `NSWindow` whose `contentRect` is `CGRect(origin: .zero, size: view.intrinsicContentSize)`, set the view as its content view with `view.frame = window.contentView!.bounds` before the swap, and bitmap `view.bounds` with `bitmapImageRepForCachingDisplay(in:)` / `cacheDisplay(in:to:)`. The painted height is the alpha > 0.5 run down the centre column (`pixelsWide / 2`, inside the card because the outset is symmetric), scaled by `rep.pixelsHigh / view.bounds.height`, as today. No settle loop: the layout exists after `init` | Same assertion, now on the live card | Comparing `layout` against itself without pixels: proves nothing about paint |
| D-7 | Delete `documentCoordinateSpace` and the `.coordinateSpace(name:)` line that uses it | Nothing resolves the name (`rg "evidenceGraphDocument"` finds only the definition) | Keeping it: dead code |
| D-8 | Unpromote fill rect: a square of side `PVControlSize.sm.height`, centred in `parts.slot`. Glyph rect: a square of side `PVControlSize.sm.iconGlyphSize`, centred in `parts.slot` | The values the deleted `PVIconButton(size: .sm)` read; one token change now moves both | Literal 22 and 13: drift (MNT-3) |
| D-9 | Rename `EvidenceAppKitCard` → `EvidenceCardView` (file `EvidenceAppKitCard.swift` → `EvidenceCardView.swift`, with its four pbxproj entries) and `EvidenceCanvasCardContent.appKitView(for:)` → `cardView(for:)`, before any new call site | After this PR there is no non-AppKit card; the name should say what it is. 10 references today | Leaving "AppKit" in the name: marks a distinction that no longer exists (M-5) |
| D-10 | Ghost and layout tests stay in `EvidenceCardLayoutTests`; no new test file | It already builds cards from `placed(rows:…)` fixtures | A new suite for two tests |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `EvidenceCardView(ghostOf:)` *(new)* | init | `EvidenceCardView(card:actions:)` | It is a convenience init on that view |
| `EvidenceCanvasCard.ghost(_:presentation:)` *(new)* | func | `EvidenceSubjectCard.ghost(placed:presentation:)` (deleted) | It replaces that one for one on the value type |
| `EvidenceCanvasCard.State.resting` *(new)* | static | Hand-written all-false `State` in `EvidenceGraphModelTests.swift:1615` | That is a test literal; this is the shared value |
| `EvidenceGhostCard` *(new)* | `NSViewRepresentable` | `EvidenceCanvasCardContent.cardView(for:)` builds the item view | That path records a layout and adds actions, for the item layer. Searched `NSViewRepresentable` in `Features/EvidenceGraph` and `Features/GraphCanvas`: none |
| `GraphCanvasItem.makeView` *(renamed)* | property | `GraphCanvasItem.appKitView` | It is `appKitView`, made required |
| `EvidenceCardView` *(renamed)* | class | `EvidenceAppKitCard` | Rename (D-9) |

**Builds on.** [#367](https://github.com/mendahu/provenencia/pull/367)'s live card and `EvidenceCanvasCard` value; `namedActions(for:)` from S10-44 as the one action list.
**Deletes.** `EvidenceSubjectCard` instance properties, `body`, `ghost(placed:presentation:)`, `ghostOpacity`, `EvidenceSubjectCardChrome`, `documentCoordinateSpace`, `shellPaddingBottom`, `promoteInset`. `EvidenceBridgeCard` instance properties, `body`, `EvidenceBridgeCardChrome`. The file `EvidenceCitedPropertyRow.swift` (four types) and its four pbxproj entries. `EvidenceCardHitRegionKey`, `View.evidenceCardHitRegion(_:)`, `View.reportsEvidenceCardLayout(_:)`. `EvidenceCanvasCardContent.view(for:)`, `subject(_:presentation:card:)`, `bridge(_:card:)`. `GraphCanvasItem.content`, `GraphCanvasItemLayer.padded(_:)` and its `NSHostingView` branches. The layout test helper's settle loop.
**Next adds.** A new card control is added in `EvidenceCardView` (paint and `regions`), the static `actionTargets`, and `namedActions(for:)` — three places, no SwiftUI copy.

### Schema

None — no catalog change.

### Go core

None — macOS only.

### FFI

None — macOS only.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Ghost card | `EvidenceCardView(ghostOf:)` | `EvidenceGraphView.swift` `EvidenceGhostCard` *(new, private)* | D-1, D-4, D-5 |
| Card item | `EvidenceCardView` | `EvidenceGraphView.pushItems` → `GraphCanvasItem(makeView: { content.cardView(for: card) })` | D-3, D-9 |

```swift
private struct EvidenceGhostCard: NSViewRepresentable {
    let card: EvidenceCanvasCard
    func makeNSView(context: Context) -> EvidenceCardView { EvidenceCardView(ghostOf: card) }
    func updateNSView(_ view: EvidenceCardView, context: Context) {}
}
// base layer:
EvidenceGhostCard(card: .ghost(ghost, presentation: model.presentation(for: ghost.kind.rawValue)))
    .id(ghost.kind.rawValue)
    .fixedSize()
    .allowsHitTesting(false)
    .accessibilityHidden(true)
    .position(CGPoint(x: center.x - canvas.origin.x, y: center.y - canvas.origin.y))
```

States: no change. Accessibility: the ghost is not an element (today `.accessibilityHidden(true)`); cards keep their custom actions from `namedActions(for:)`.

**Comment and README text after STEP-5** (exact meaning; wording may follow the file's voice):

| File | Text |
| --- | --- |
| `EvidenceSubjectCard.swift` header | "Geometry, action ids and VoiceOver copy for a primary subject card. ``EvidenceCardView`` paints the card; hit targets and edges read these." |
| `EvidenceBridgeCard.swift` header | Same, for a bridge card; keep the board layout sentence |
| `EvidenceCardLayout.swift` type comment | "Painted geometry of one graph card, measured by ``EvidenceCardView``." The rest of the paragraph stays |
| `EvidenceCardLayout.swift` `regions` | "Action id → painted rect of that piece." |
| `EvidenceCardView.swift` header | Replace "The placement ghost stays SwiftUI." with "The placement ghost is this view built with ``init(ghostOf:)``." |
| `EvidenceCanvasCard.swift` line 7 | "rebuilds a card's view" for "rebuilds a card's hosting view" |
| `EvidenceCanvasCard.swift` `EvidenceCanvasCardContent` | "Builds a card's view and its VoiceOver actions. Holds references only (the model, navigation, and layout store)." |
| `GraphCanvasItemLayer.swift` | Each item is an `NSView` the host builds with `makeView`, sized by its `intrinsicContentSize` and including `paintOutset`; AppKit owns where items are and rebuilds a view only when its model changes. Lines 18–22, 41, 120 and the `makeView` comment say "view", not "hosting view" or SwiftUI |
| `GraphCanvasDocumentView.swift` line 10 | "one view per card" |
| `macos/App/Features/GraphCanvas/README.md` | Row 19: "One `NSView` per item, built by the host (`makeView`), positioned by AppKit; rebuilt only when the item's `Equatable` model changes". Row 35: "(id, origin, model, `makeView`, rotor entry)". Line 46 paragraph: each item is its own `NSView` above the base layer. Line 63 paragraph (through line 66): items set their own AppKit accessibility (label, custom actions); the document view serves the rotors because items are separate views. Lines 74–79: "The Evidence graph's `EvidenceCardView` measures its pieces into an `EvidenceCardLayout` and paints from the same measurement; hit targets and edge anchors are built from it (constants only before a card is built)." Keep the sentences about drawn pixels vs `fittingSize` and about SwiftUI gestures on the base layer |

### Strings

None — no copy added or removed. Every `L10n.EvidenceGraph` key the deleted chrome used is also used by `EvidenceCardView`, `namedActions` or the static geometry; STEP-5's check confirms none goes unused.

## 7. Steps

**STEP-1 — Rename the live card**
Files: `EvidenceAppKitCard.swift` → `EvidenceCardView.swift` (`git mv`), `EvidenceCanvasCard.swift`, `EvidenceGraphView.swift`, `project.pbxproj`.
Do: D-9. Update the four pbxproj entries for the file (path and comments).
Check: `rg -n "EvidenceAppKitCard|appKitView\(for" macos` prints nothing; the app builds; `scripts/check-pr.sh` PASS.

**STEP-2 — Ghost on the live card**
Files: `EvidenceCanvasCard.swift`, `EvidenceCardView.swift`, `EvidenceGraphView.swift`, `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift`. Skill: `add-swift-test`.
Do: write T-1 and T-2 first, with `placed(rows: 0)` as the subject. Add D-5's `resting` and `ghost(_:presentation:)`, D-1's `init(ghostOf:)` and alpha, and D-4's `EvidenceGhostCard`; replace the `EvidenceSubjectCard.ghost(…)` block in `EvidenceGraphDocumentBody.body` with the §6 sketch.
Check: T-1, T-2 pass.

**STEP-3 — Layout tests measure the live card**
Files: `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift`.
Do: rewrite `measure` per D-6 to take an `EvidenceCanvasCard`. `measureSubject` builds `EvidenceCanvasCard(placed: .subject(p, nil), text: EvidenceCardText(subject: p), state: .resting)`. `bridgeTargetsFollowThePaint` builds `EvidenceCanvasCard(placed: .bridge(bridge), text: EvidenceCardText(bridge: bridge, in: snapshot), state: .resting)`. The bitmap includes the `paintOutset` margin; the centre-column run measures the card body, so the assertion stays `abs(layout.size.height - drawn) < 0.5`. Keep every case and every other assertion.
Check: all nine `measuredLayoutMatchesThePaint` cases and `bridgeTargetsFollowThePaint` pass, with the SwiftUI views still present.

**STEP-4 — One view builder per canvas item**
Files: `GraphCanvasItemLayer.swift`, `EvidenceGraphView.swift` (`pushItems`), `macos/ProvenenciaTests/GraphCanvasItemLayerTests.swift`.
Do: D-3. In the test file add a private `final class SizedView: NSView` whose `intrinsicContentSize` is `NSSize(width: 100 + 2 * GraphCanvasItemLayer.paintOutset, height: 40 + 2 * GraphCanvasItemLayer.paintOutset)`; the `item(…)` helper's `makeView` bumps `counter` and returns a `SizedView`. Every existing assertion stays.
Check: `GraphCanvasItemLayerTests` pass; the app builds.

**STEP-5 — Delete the SwiftUI cards**
Files: `EvidenceSubjectCard.swift`, `EvidenceBridgeCard.swift`, `EvidenceCitedPropertyRow.swift` (delete), `EvidenceCardLayout.swift`, `EvidenceCanvasCard.swift`, `EvidenceCardView.swift`, `EvidenceGraphView.swift`, `GraphCanvasItemLayer.swift`, `GraphCanvasDocumentView.swift`, `macos/App/Features/GraphCanvas/README.md`, `project.pbxproj`.
Do: everything under **Deletes**. Make both card types `@MainActor enum` (D-2). Delete D-7's lines. Write the §6 comment and README text. Every other static on the two enums stays, even one you find unread: the review confirmed each has a live reader.
Check: T-8. `python3 scripts/check-localizable-xcstrings.py` is OK. `scripts/check-pr.sh` PASS.

**STEP-6 — S10-44 follow-ups**
Files: `EvidenceCardView.swift` (`paintFooter`), `macos/ProvenenciaTests/EvidenceSubjectCardFooterTests.swift`.
Do: D-8. In `promotedFooterSplitsHandleAndUnpromote`, add a filed cited bridge: `SourceGraphPlacedBridge(subject: CatalogSubject(id: "b1", ref: "REL-3RN7K", sourceID: "x", subjectTypeID: "t", label: "", description: ""), kind: .relationship, typeLabel: "Relationship", gridX: 5, gridY: 5, isCited: true, observations: [<one observation>], membership: CatalogSubjectMembership(subjectID: "b1", claimID: "cb1", entity: CatalogCanonicalEntity(id: "eb1", ref: "REL-7KD45", subjectTypeID: "t", label: ""), kind: "relationship"))` in a `SourceGraphSnapshot(sourceId: "x", bridges: [bridge])`, and expect `EvidenceBridgeCard.actionTargets(for:in:canCite: true)` to have no `EvidenceSubjectCard.unpromoteActionID` (T-5). Copy the observation shape from `EvidenceCardLayoutTests.observation(_:value:)`.
Check: T-4, T-5 pass.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-51-one-card-drawing.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `EvidenceCardLayoutTests` `ghostCardMatchesTheLiveLayout` *(new)* | `p = placed(rows: 0)`: `EvidenceCardView(ghostOf: .ghost(p, presentation: nil)).layout` == `EvidenceCardView(card: EvidenceCanvasCard(placed: .subject(p, nil), text: EvidenceCardText(subject: p), state: .resting), actions: []).layout` | AC-2.2, D-1, D-5 |
| T-2 | Swift | `EvidenceCardLayoutTests` `ghostCardIsNotAnAccessibilityElement` *(new)* | Ghost of `p` → `isAccessibilityElement() == false`, `accessibilityLabel()` nil or empty, `accessibilityIdentifier()` empty, `accessibilityCustomActions()` empty or nil; the non-ghost card → `isAccessibilityElement() == true`, identifier `evidenceGraph.subject.s` | AC-2.1, AC-2.2, D-1 |
| T-3 | Swift | `EvidenceCardLayoutTests` `measuredLayoutMatchesThePaint` (9 cases), `bridgeTargetsFollowThePaint` *(changed)* | Each fixture on `EvidenceCardView` → painted height = `layout.size.height` ± 0.5; every existing region and target assertion holds | AC-1.3, D-6 |
| T-4 | Swift | `EvidenceCardLayoutTests` `measuredLayoutMatchesThePaint` `cited3+promoted`, `longHandleName+promoted` | Still pass after D-8 (regions do not change). The drawn sizes are manual check 3 | AC-2.3 |
| T-5 | Swift | `EvidenceSubjectCardFooterTests` `promotedFooterSplitsHandleAndUnpromote` *(changed)* | Filed cited bridge → no `unpromote` target | AC-2.4 |
| T-6 | Swift | `GraphCanvasItemLayerTests` (all) *(changed helper)* | `makeView` items → placement, rebuild counts, raise order, rotors as today | AC-1.2, D-3 |
| T-7 | Swift | `EvidenceGraphModelTests` `unpromoteVoiceOverActionFollowsFooterAction` | Unchanged; passes | AC-1.4 |
| T-8 | Search | STEP-5 check | `rg -n "EvidenceSubjectCardChrome\|EvidenceBridgeCardChrome\|EvidenceCitedPropertyRow\|EvidenceCitedPropertyStack\|reportsEvidenceCardLayout\|evidenceCardHitRegion\|func view\(for card\|struct EvidenceSubjectCard\b\|struct EvidenceBridgeCard\b\|promoteInset\|shellPaddingBottom\|documentCoordinateSpace" macos` → no output. `rg -n "NSHostingView\|import SwiftUI\|hosting view per card" macos/App/Features/GraphCanvas/GraphCanvasItemLayer.swift macos/App/Features/GraphCanvas/GraphCanvasDocumentView.swift` → only `GraphCanvasDocumentView`'s base-layer `hostingView` lines | AC-1.1, AC-1.2, D-2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `GraphCanvasEdgeGeometryTests`, `GraphCanvasPointerHitTestingTests`, `EvidenceGraphModelTests` (they read the statics D-2 keeps).

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
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: none; GraphCanvas README per STEP-5
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a Source with a cited person, an uncited person, a promoted person, a cited relationship and an uncited participation. Every card looks as it did before this PR, at 1× and at 0.25×.
2. Arm the Person tool and move over the canvas. The faded Person card follows the pointer, centred on the grid point, sharp on a Retina display, under any card it passes. Click: the card is placed and the ghost disappears. Arm Event: the ghost shows an Event card.
3. Hover the Unpromote square on the promoted card: the same small fill and glyph as before. Hover and click the handle row, Promote, Add property and a property row: each works as before.
4. With VoiceOver on, move through the Subjects rotor. The ghost is never read. A filed card's actions list Unpromote; an unfiled card's do not.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A retargeted layout case fails: the live paint and its own `layout` disagree | That is a real bug the old tests hid. Stop (below); do not loosen the tolerance |
| The ghost renders blurry inside the SwiftUI base layer | `syncBackingScale` runs on `viewDidMoveToWindow`; manual step 2 checks it |
| The representable sizes to zero | `intrinsicContentSize` is set; `.fixedSize()` makes SwiftUI use it |

**Stop and report if:**

- `s10-44-unpromote` (or `main`, once #418 merges) has changes to `EvidenceSubjectCard.swift`, `EvidenceAppKitCard.swift` or `EvidenceCanvasCard.swift` after `1c489b80` that make a §4 row untrue.
- After STEP-3, any `measuredLayoutMatchesThePaint` case or `bridgeTargetsFollowThePaint` fails on `EvidenceCardView`. Report the case, the layout and drawn heights, and the region that disagrees.
- Making `EvidenceSubjectCard` or `EvidenceBridgeCard` an `@MainActor enum` breaks a caller that needs a non-isolated member beyond those already `nonisolated`.
- A `L10n.EvidenceGraph` key becomes unused after STEP-5.
- `GraphCanvasItem(` is built anywhere besides `EvidenceGraphView.pushItems` and `GraphCanvasItemLayerTests`.

## 12. Open questions

None.

## 13. PR

**Title:** S10-51 — Draw Evidence cards in one place
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-51-one-card-drawing.md`.
