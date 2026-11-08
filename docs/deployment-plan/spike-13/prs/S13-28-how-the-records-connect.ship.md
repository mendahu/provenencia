# Ship log — S13-28 How the records connect

**Plan:** [`S13-28-how-the-records-connect.md`](S13-28-how-the-records-connect.md) · **Branch:** `s13-28-how-the-records-connect` (stacked on `s13-27-plan`, #743) · **PR:** #744
**State:** fixing 3 (user-requested refactor)

Stacked PR: the base branch is `s13-27-plan`, not `main` (the user, 2026-11-08). The gate runs `scripts/check-pr.sh --base origin/s13-27-plan`, and reviewers diff against the `s13-27-plan` head.

## Round 0 — implement
Head `d788a00e`. Gate PASS (`--base origin/s13-27-plan`; xcode-test 70s). Report deviations (all in the plan's Landed note):
- `previewSession(entityId:recordLinks:failure:)` instead of a `delay:`; nil links answer an hour late so existing previews show the spinner.
- New file `PersonRecordLinksColumn.swift` (title, caption, width and fitting logic).
- `ConclusionPageSectionHost.emptyTitle` is `LocalizedStringResource?` (nil default) so no empty literal becomes a catalog key.
- Nodes placed with `.position`; the first frame before measuring is off by the dot's inset.
- `PVPageColumnsLayout.widths` is `@Sendable`; `PVPageColumns.columns` captures its inputs as locals (Swift warnings).
- Unguarded hover ordering: a late `onHover(nil)` leave after the next node's enter clears the highlight.
- §10 manual check not run (pin while scrolling, resize focus, short-window shrink, click-through, VoiceOver, Tab).

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Skipped: none (UX: `macos/App/**` and strings; SEC+PRF: `macos/App/Features/**` models loop over the record-links read's nodes and arcs).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | Every hover or focus change re-strokes every arc on the main thread | Fix now → F-1 | verified: the plan's Scale line ("about 120 arcs") is wrong. S13-27's own Scale fixture `householdGraphs(50)` is 1,225 pairs × 3 lines = 3,675 arcs (`completed.md`). `RecordLinksDiagram.draw(_:highlighted:into:)` strokes all of them each time `highlighted` changes. Reviewer's CG proxy: 0.1–0.4 s per change. Splitting the layers keeps D-5 (arcs on a `Canvas`, hit testing off) |
| PRF-2 | PRF | minor | Content and layout rebuilt on every hover | Fix now → F-1 | the host's `body:` closure captures `highlighted`, so `make` runs on each hover; F-1's split removes it |
| CON-4 | CON (+PRF outside lens) | minor | `zip(arcs, layout.curves)` misaligns when the layout drops an arc | Fix now → F-1 | `RecordLinksLayout.curves` is a `compactMap`, and `arcsWithUnknownNodesAreSkipped` makes the drop the contract |
| SEC-1 | SEC | blocker | No bound on arcs or nodes drawn; a crafted catalog sends up to ~1.1M lines to a main-thread draw | Fix now → F-2 | verified: `recordlinks.Pairs` past `MaxReadWork` still lists every pair's lines at 0 facts (`core/recordlinks/recordlinks.go:183-193`), and `PersonRecordLinksContent.init` appends one arc per line with no count check. Working rule "caps for adversarial-only costs": one high total cap |
| MNT-1 / CON-1 | MNT, CON | blocker | The neighbour set is computed three times, and the tested copy is called only from tests | Fix now → F-3 | `rg "neighbors\(of" macos` finds only the definition and the test |
| MNT-2 | MNT | major | The diagram places a single node itself, beside the layout's tested one-node case | Fix now → F-4 | `RecordLinksDiagram.swift:53,66` branches on `nodes.count == 1` with its own constants |
| MNT-3 | MNT | major | The DesignSystem README doesn't list `PVPageVisibleRect`, `PVPageColumnsLayout`, `sideColumnWidths`, `pinOffset` or `widthSideColumnMax` | Fix now → F-5 | the README PageColumns row (line 192) is unchanged |
| MNT-4 | MNT | minor | The stacked decision and the width fallback are each worked out twice | Fix now → F-6 | `ConclusionPinnedSideColumn` falls back to `widthContentMax` beside Metrics' own fallback |
| PRF-3 | PRF | minor | The pinned column's layout may re-measure the rows on every scroll tick | Fix now → F-6 | low confidence; F-6 confirms it with `_printChanges` first and reports Not reproduced if the layout does not rerun |
| MNT-5 | MNT | minor | The shared host hard-codes the diagram's height (232) | Fix now → F-7 | the literal is `minimumDiameter + 2 × verticalRoom` |
| MNT-6 | MNT | minor | The Person-free diagram sets `persons.detail.sources.node.*` ids | Fix now → F-7 | `RecordLinksDiagram.swift:140` |
| UX-1 / CON-2 | UX, CON | minor | A node with no location is an enabled button that does nothing | Fix now → F-8 | D-4 says "not clickable"; there is no `.disabled` |
| UX-2 / CON-3 | UX, CON | minor | Past 30 nodes, the focused node's ref is hidden while another graph is hovered | Fix now → F-8 | D-5 says "highlighted or focused"; `showsRef` reads only `state` |
| UX-3 | UX | minor | A hover on a removed or moved line is never cleared and hides the focus highlight | Fix now → F-8 | low confidence; clearing on content change costs one line and is safe either way |

Nits (logged): `MNT-7 · PVPageColumns.swift PVPageSideColumnWidths.init · drop the memberwise copy` · `MNT-8 · RecordLinksDiagram.swift:283 · inert PVFieldFocus static for previews` · `CON-5 · PVPageColumnsTests.swift:76 · test name says 1014, asserts 642` · `CON-6 · ship log header stale (fixed by the orchestrator)` · `UX-4 · PersonRecordLinksContent.swift · empty title fallback formats "Household 214 — "` · `UX-5 · PersonRecordLinksColumn.swift:240 · title lacks .isHeader`

Process note: the MNT reviewer ran `git checkout` on the head commit during its read-only review and put the branch back; the tree was clean afterwards.

### Fix plan 1

**F-1 — Draw the arcs at rest once; redraw only the highlighted node's arcs on hover** (from PRF-1, PRF-2, CON-4)
Files: `macos/App/Features/Conclusions/RecordLinksDiagram.swift`, `RecordLinksLayout.swift`, `PersonSourcesSection.swift`, `PersonSourcesPreviewData.swift`, `macos/ProvenenciaTests/RecordLinksLayoutTests.swift`, `PersonRecordLinksContentTests.swift`; the plan's §6 Scale line and Landed note. Skill: `add-swift-test`, `add-ui-component`.
Failing test first: `RecordLinksLayoutTests.eachCurveKeepsItsArc`. Use 3 nodes with arcs `[arc(0, 5), arc(1, 2, strong: true)]`. The one curve that comes back carries arc index 1 (or `strong == true` and endpoints 1–2). Also add `PersonRecordLinksContentTests.householdScale`: content for `PersonSourcesPreviewData.householdRecordLinks(graphs: 50)` *(new, every pair shares three people)* has 50 nodes and 3,675 arcs.
Do:
- Give `RecordLinksLayout.Curve` the arc it draws (its index into `arcs`, or `a`, `b` and `strong`). Draw from `layout.curves` alone, with no `zip`.
- Split the drawing into two layers:
  - An at-rest `Canvas` whose inputs are only the layout (nodes, arcs, size). It strokes every arc, thin then thick, `borderStrong`. It sits in a child view that is `Equatable` on those inputs (or is otherwise kept out of the highlight's dependency), so a hover does not redraw it. While something is highlighted, the layer takes `.opacity(0.18)`, a composite rather than a redraw.
  - A second `Canvas` above it that strokes only the highlighted node's arcs, in `accentLine`.
  The look in D-5 is unchanged.
- Build `PersonRecordLinksContent` and `RecordLinksLayout` once per read value and size, not on each hover. Pass the highlight separately: the host's `body:` closure hands the built content to a child view that takes the content and `highlightedGraphID`. Nothing that changes on hover may call the content's init.
- Add a 50-graph household preview using `householdRecordLinks(graphs: 50)`, beside the existing "Many" preview.
- Correct the plan's §6 Scale line from "about 120 arcs" to S13-27's 3,675-arc household fixture, and note the change in the Landed note.
Check: the new tests pass. In the household preview, wrap the at-rest layer's body in `let _ = Self._printChanges()` temporarily and confirm it does not print on hover, then remove the line. `scripts/check-pr.sh --base origin/s13-27-plan` PASS.

**F-2 — One high total cap on arcs and nodes drawn** (from SEC-1)
Files: `macos/App/Features/Conclusions/PersonRecordLinksContent.swift`, `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift`; the plan's Edges and Landed note.
Failing test first: `PersonRecordLinksContentTests.arcsAndNodesStopAtTheCap`. Build a read with more lines than `maxArcs`, some strong and some thin. Content has exactly `maxArcs` arcs and keeps every strong one before any thin one. Build a read with `maxNodes + 5` graphs. Content has `maxNodes` nodes, and no arc names a dropped node.
Do:
- Add `static let maxArcs = 75_000` (20 × the 3,675 household Scale) and `static let maxNodes = 1_000` (20 × the 50-graph Scale) on `PersonRecordLinksContent`. Each gets a comment stating that rule, in the form of `core/matchregistry/recordlinks.go:18`.
- In `init`, keep the first `maxNodes` graphs in read order. Count lines before appending. Past `maxArcs`, keep strong lines first, then thin ones in read order, until the cap.
- No new UI and no copy: only a crafted catalog reaches either cap.
Check: the test passes; `check-pr.sh` PASS.

**F-3 — One neighbour function, used by the content and the diagram** (from MNT-1, CON-1)
Files: `PersonRecordLinksContent.swift`, `RecordLinksDiagram.swift`, `PersonRecordLinksContentTests.swift`.
Failing test first: `RecordLinksArcTests` or a case in `PersonRecordLinksContentTests`, `#expect(RecordLinksArc.neighbors(of: 0, in: content.arcs) == [1, 3, 4])` on the Johann fixture. It does not compile until the function exists.
Do:
- Add `static func neighbors(of: Int, in: [RecordLinksArc]) -> Set<Int>` next to `RecordLinksArc`.
- `init`'s `connected` count and `RecordLinksDiagram` (or F-1's highlight layer) both call it.
- Delete the instance `neighbors(of:)` and its private copy in `init`, and point T-4 at the shared function.
Check: `rg -n "arc\.a == |arc\.b == " macos/App` prints only the shared function's lines. `check-pr.sh` PASS.

**F-4 — The layout owns the one-node geometry** (from MNT-2)
Files: `RecordLinksLayout.swift`, `RecordLinksDiagram.swift`, `RecordLinksLayoutTests.swift`.
Failing test first: change T-5's one-node row in `zeroOneAndTwoNodes` to expect the point (width / 2, 24) and size (width × 64).
Do:
- Move `singleDotY` and `singleHeight` into `RecordLinksLayout` and return them from its one-node branch.
- The diagram always reads `layout.nodes[i]` and `layout.size`. It draws the circle only when there is more than one node, and has no other `nodes.count == 1` branch for geometry.
Check: `rg -n "nodes.count == 1" macos/App/Features/Conclusions/RecordLinksDiagram.swift` prints at most the circle and one-record-line conditions. `check-pr.sh` PASS.

**F-5 — List the kit's new side-column pieces in the DesignSystem README** (from MNT-3)
Files: `macos/App/DesignSystem/README.md` (PageColumns row, line 192).
Do: extend the row with:
- `PVPageVisibleRect`: the one reader of the visible rectangle in a section's own coordinates.
- `PVPageColumnsLayout`: places a section's main and side column from widths it is given; `PVPageColumnsMetrics.sideColumnWidths`, `pinOffset` and `PVSpacing.widthSideColumnMax` belong with it.
Check: `rg -n "PVPageVisibleRect|PVPageColumnsLayout" macos/App/DesignSystem/README.md` hits; `check-pr.sh` PASS.

**F-6 — One width fallback; the scroll moves only the side column's offset** (from MNT-4, PRF-3)
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/App/Features/Conclusions/ConclusionPinnedSideColumn.swift`, `macos/ProvenenciaTests/PVPageColumnsTests.swift`.
Failing test first: `PVPageColumnsTests.sideColumnWidthsFromAProposal`. An unspecified or infinite proposal gives the same result as `columnWidths(proposal:)`'s fallback pane.
Do:
- Add `PVPageColumnsMetrics.sideColumnWidths(proposal:mainMin:)` next to `columnWidths(proposal:)`. `ConclusionPinnedSideColumn`'s layout closure and its stacked check both use it, with no local fallback.
- For PRF-3: first check with `_printChanges` (or a counter in `placeSubviews`) whether scrolling the "Many" preview reruns the layout. If it does, restructure so that only the side column's offset reads `PVPageVisibleRect`'s `visible`, and the layout is built outside that read. If it does not, report PRF-3 Not reproduced with what you saw.
Check: the test passes; `check-pr.sh` PASS.

**F-7 — The column placement carries its loading height; the diagram takes an id root** (from MNT-5, MNT-6)
Files: `ConclusionPageSectionHost.swift`, `PersonSourcesSection.swift`, `RecordLinksDiagram.swift`.
Do:
- `placement: .column(loadingHeight: CGFloat)`, with the Sources caller passing `RecordLinksLayout.minimumDiameter + 2 * RecordLinksLayout.verticalRoom` (expose `verticalRoom` if it is private). Delete `columnLoadingHeight`.
- `RecordLinksDiagram` takes `idRoot: String`, and the Person caller passes `persons.detail.sources`. Node ids stay `persons.detail.sources.node.<id>`.
Check: `rg -n "232|persons\.detail" macos/App/Features/Conclusions/RecordLinksDiagram.swift macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` prints no literal 232 and no Person id in the diagram. `check-pr.sh` PASS.

**F-8 — Node and hover details** (from UX-1/CON-2, UX-2/CON-3, UX-3)
Files: `RecordLinksDiagram.swift`, `PersonSourcesSection.swift`.
Do:
- A node whose `location == nil` is `.disabled` and gets `pvActivatesOnReturn(when: false)` (D-4).
- `showsRef` is true also when the node's own button has focus (D-5 "highlighted or focused"), read in the button through the focus handle or `isFocused`.
- In `PersonSourcesList`'s `.onChange(of: content)`, clear `hoveredGraphID` when the new content's `claimGraphs` no longer holds it.
Check: `check-pr.sh` PASS. In the 50-graph preview, Tab to a node while the pointer rests on another line: the focused node's ref shows.

## Round 1 — fix
Head `f3e2e17a` (commits `af360965`, `4813b1c6`, `c439fd3e`, `f3e2e17a`). Gate PASS (orchestrator run, xcode-test 57s).
- F-1 to F-8 are done.
- Only F-4's test was seen failing first. The F-1, F-2, F-3 and F-6 tests use new symbols, so they could not compile before the fix.
- **F-1:** hover moved into `@Observable RecordLinksHover`. The list, the host and `make` don't read it, so a hover no longer rebuilds content or layout. The rest arcs are an `Equatable` `RecordLinksRestArcs` Canvas, dimmed with `.opacity(0.18)`. `RecordLinksAccentArcs` draws only the highlighted node's arcs. The `_printChanges` check was not run, because the implementer can't drive the app.
- **F-6, PRF-3:** not checked. It needs the running app, so it moves to the manual check and is not marked Not reproduced.
- **Other changes:** `ConclusionPageSectionHost.Placement` is no longer `Equatable`. `RecordLinksArc.touches(_:)` was added for the accent layer.
- **UX minors (F-8):** resolved. I read them in the diff:
  - `.disabled(node.location == nil)` and `pvActivatesOnReturn(when: node.location != nil)`.
  - `showsRef … || handle.isActive()`.
  - `onChange(of: content)` clears `hover.graphID`.
  UX does not go to round 2.

## Round 2 — review
Reviewers: CON (CON-1 major), MNT (MNT-1 blocker, MNT-2 and MNT-3 majors), SEC+PRF (PRF-1 and SEC-1 blockers). Scope `d788a00e...f3e2e17a`. Skipped: UX (minors only, resolved above).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 / MNT-1 | CON, MNT | blocker | Neighbour set computed three times | Resolved | one `RecordLinksArc.neighbors(of:in:)`, called from `init` and `RecordLinksHighlightedLayers`; T-4 points at it |
| MNT-2 | MNT | major | Diagram placed one node itself | Resolved | `RecordLinksLayout` returns (w/2, 24) and height 64; no `nodes.count == 1` in the diagram |
| MNT-3 | MNT | major | README missing the side-column pieces | Resolved | README line 192 |
| PRF-1 | PRF | blocker | Every hover re-strokes every arc | Resolved (by reading) | `RecordLinksRestArcs(layout:).equatable()`, hover in `RecordLinksHover`; 147 strokes per hover at Scale, not 3,675. `_printChanges` not run; moved to the manual check |
| SEC-1 | SEC | blocker | No bound on arcs or nodes drawn | Resolved | `maxArcs = 75_000`, `maxNodes = 1_000`; `arcsAndNodesStopAtTheCap` |
| PRF-4 (CON, MNT, PRF outside lens) | PRF | minor | `PersonRecordLinksContent.init` calls `neighbors(of:in:)` once per node, so O(nodes × arcs) | Follow-up | measured by PRF: 0.36 ms at Scale (50 × 3,675), 70 ms `-O` at the caps; only a crafted catalog reaches the caps. Round 2 fixes blockers only |

Nits (logged): `CON-7 · S13-28 plan T-4 and D-4/D-5/D-6 still name neighbors(of: Int), highlightedGraphID, onHover; the Landed note records the change` · `CON-8 · docs/macos-client-patterns.md:104 · says placement: .column, now .column(loadingHeight:)`

No Fix plan 2: round 2 raised no blocker.

## Result
Gate PASS (`f3e2e17a`) · CI not run (workflows trigger on PRs into `main`; runs once #743 merges and #744 retargets) · open: none · to user: —
Follow-ups, placed by the user (2026-11-08):
- PRF-4 → `docs/ideas/structural-debt.md` (the S13-28 ship log PRF-4 row).
- PRF-3 → checked in the app by the user; the pinned column does not re-measure the rows on scroll. No change.
- CON-8 → fixed on this PR (`docs/macos-client-patterns.md:104`).

## Round 3 — user-requested refactor (2026-11-08)

The user asked for this round, after the PR was ready: "I want the graph component built completely agnostic to the data. It should be a generic thing that just takes in nodes and the amounts and qualities of arcs between them … similar to what you'd expect from a graphing chart library, where you can give it the props for colors, thickness, labels". It stays in `Features/Conclusions`; it does not move into the kit. Nothing a researcher sees changes: same geometry, colours, thicknesses, labels, hover, focus and spoken text. This round also takes PRF-4 (O(nodes × arcs) neighbour counting), which the user had parked in `docs/ideas/structural-debt.md`, because the adapter is being rewritten anyway; its row there is removed.

The term is the map's own: a **circular link diagram** (working rule "The view is a circular link diagram"), with **nodes** and **arcs**.

### Fix plan 3

**F-9 — A generic circular link diagram, its layout and its style** (user)
Files *(new)*:
- `macos/App/Features/Conclusions/CircularLinkDiagram/CircularLinkDiagram.swift`
- `CircularLinkDiagram/CircularLinkLayout.swift`
- `CircularLinkDiagram/CircularLinkDiagramStyle.swift`
- `CircularLinkDiagram/CircularLinkData.swift`

Deleted: `RecordLinksDiagram.swift`, `RecordLinksLayout.swift`. The project uses synchronized folders, so no `project.pbxproj` edit is needed. Skills: `add-ui-component`, `add-swift-test`.

Nothing under `CircularLinkDiagram/` may name a Person, a graph, a Source, `WorkspaceLocation`, `Catalog*`, `L10n` or a `persons.*` id. The Check greps for it.

Data, in `CircularLinkData.swift`:
```swift
struct CircularLinkNode<ID: Hashable>: Identifiable, Equatable {
    var id: ID
    var label: String                  // printed beside the dot
    var help: String? = nil            // the native .help tooltip; nil draws none
    var accessibilityLabel: String? = nil   // nil reads `label`
    var isEnabled: Bool = true         // false: drawn, not clickable, no Return (D-4)
}
struct CircularLinkArc<ID: Hashable>: Equatable {
    var from: ID
    var to: ID
    var lineWidth: CGFloat
    var color: Color? = nil            // nil: style.arcColor
}
@MainActor @Observable final class CircularLinkHover<ID: Hashable> { var id: ID? }
```
The caller passes arcs one at a time. Several arcs on one unordered pair are the "amount": the layout groups them by pair, in input order, and fans them. Callers no longer pass `slot` or `slots`. An arc whose end is not a node, or whose ends are the same node, is skipped.

Neighbours (PRF-4 and MNT-1's single function), static on `CircularLinkArc`:
- `static func neighbors(of: ID, in: [Self]) -> Set<ID>`, used by the diagram for the one highlighted node.
- `static func neighborSets(in: [Self]) -> [ID: Set<ID>]`, one pass over the arcs, used by adapters that need every node's count.

Both are tested.

Style, in `CircularLinkDiagramStyle.swift`: `struct CircularLinkDiagramStyle: Equatable` holds every number and colour the diagram draws with today, under the names below. Its `static let standard` gives exactly today's values (D-5 and the board):

| Prop | Today's value |
| --- | --- |
| `circleColor` | `PVColor.borderSubtle` |
| `circleLineWidth` | 1 |
| `arcColor` | `PVColor.borderStrong` |
| `highlightedArcColor` | `PVColor.accentLine` |
| `dimmedArcOpacity` | 0.18 |
| `arcCurvature` | 0.3, the control point's fraction from the centre toward the chord's midpoint |
| `arcSpacingFraction` | 0.06 |
| `arcSpacingRange` | 7...14 |
| `dotSize` | 8 |
| `dotRingWidth` | 1.5 |
| `nodeColor` | `PVColor.textSecondary` |
| `highlightedNodeColor` | `PVColor.accent` |
| `neighborNodeFill` | `PVColor.accentSoft` |
| `neighborNodeRing` | `PVColor.accentLine` |
| `recededNodeOpacity` | 0.35 |
| `labelFont` | `PVFont.mono(size: PVTypeScale.caption, weight: .regular)` |
| `highlightedLabelFont` | the semibold variant of `labelFont` |
| `labelColor` | `PVColor.textSecondary` |
| `highlightedLabelColor` | `PVColor.accent` |
| `neighborLabelColor` | `PVColor.textPrimary` |
| `labelGap` | `PVSpacing.space3` |
| `labelsShownUpTo` | 30, past which only the highlighted or focused node prints its label |
| `overlayLabelBackground` | `PVColor.surfacePage` |
| `nodePadding` | 4 × 2 |
| `nodeHoverFill` | `PVColor.surfaceHover` |
| `sideRoom` | 84 |
| `verticalRoom` | 30 |
| `minimumDiameter` | 172 |
| `labelSideThreshold` | 0.35 |
| `singleNodeY` | 24 |
| `singleNodeHeight` | 64 |

The kit tokens are only the standard style's defaults.

Layout, in `CircularLinkLayout.swift`: `struct CircularLinkLayout<ID: Hashable>: Equatable`.
- It is today's `RecordLinksLayout` with these changes:
  - It is built from `nodes: [ID]`, `arcs: [CircularLinkArc<ID>]`, `width`, `diameter: CGFloat?` and `style`.
  - It computes slots itself.
  - Each `Curve` carries its arc's index into the input `arcs`, plus `lineWidth` and `color`.
  - `curves` are ordered thinnest first, then by input order. This generalises "thick above thin".
  - `static func diameter(width:fitting:style:)` replaces the two-argument one.
- The geometry for the standard style is unchanged to the point: today's layout tests keep their numbers.

View, in `CircularLinkDiagram.swift`: `struct CircularLinkDiagram<ID: Hashable>: View`. Its props are:
- `nodes`, `arcs`
- `style: CircularLinkDiagramStyle = .standard`
- `width: CGFloat`, `diameter: CGFloat? = nil`
- `hover: CircularLinkHover<ID>`
- `focusedID: ID? = nil`
- `accessibilityLabel: String`, the group's spoken label, passed in
- `accessibilityIdentifierRoot: String? = nil`, giving node ids `<root>.node.<String(describing: id)>`; nil sets none
- `onOpen: (ID) -> Void = { _ in }`
- `focus: ((ID) -> PVFieldFocus)? = nil`, where nil gives the nodes no focus binding

The body is today's diagram with every literal and token read from `style`, and with:
- `node.isEnabled` in place of `location != nil`
- `node.help` and `node.accessibilityLabel ?? node.label`
- `style.labelsShownUpTo` in place of `refHidingThreshold`

These keep round 1's performance shape, so the PRF-1 reasoning still holds:
- The at-rest arcs are an `Equatable` Canvas on the layout alone (and the style), dimmed by opacity.
- An accent Canvas draws only the highlighted node's arcs, each in `highlightedArcColor`.
- Only the inner layers read `hover.id`.

Each arc at rest strokes `curve.color ?? style.arcColor` at `curve.lineWidth`, with round caps.

Previews in this file use synthetic data only:
- "Six nodes": letters A–F, mixed widths, one pair with three arcs, and one node with no arcs.
- "One node".
- "Custom style": other colours, a thicker arc and a larger dot, showing the props work.
- "300 arcs, 40 nodes".

Failing test first, in `macos/ProvenenciaTests/CircularLinkLayoutTests.swift` *(new; replaces `RecordLinksLayoutTests.swift`, whose cases move here rewritten for ids)*:
- `arcsOnOnePairFanInInputOrder`: three arcs A–B, B–A, A–B get slots 0, 1 and 2 of 3, and pairs are unordered.
- `curvesDrawThinnestFirst`.
- `arcsNamingNoNodeOrOneNodeAreSkipped`.
- `eachCurveKeepsItsArcIndex`.
- `aStyleMovesTheGeometry`: `sideRoom: 50` at width 340 gives diameter 240.
- `neighborSetsMatchNeighborsOfEachNode`.

Plus every existing layout case, ported, with the same numbers.

Check:
- `rg -n "Person|Graph|Source|WorkspaceLocation|Catalog|L10n|persons\." macos/App/Features/Conclusions/CircularLinkDiagram` prints nothing.
- `rg -n "RecordLinksDiagram|RecordLinksLayout|RecordLinksArc|RecordLinksNode|RecordLinksHover" macos` prints nothing.
- The tests pass.

**F-10 — The Person adapter and column on the generic diagram** (user)
Files: `PersonRecordLinksContent.swift`, `PersonRecordLinksColumn.swift`, `PersonSourcesSection.swift`, `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift`; `docs/ideas/structural-debt.md` (remove the S13-28 PRF-4 row).
Do:
- `PersonRecordLinksContent` becomes the only Person-aware piece. It produces:
  - `nodes: [CircularLinkNode<String>]`, keyed by graph id. Each has the label (ref, else name), the help string, the spoken label as today, and `isEnabled = location != nil`.
  - `arcs: [CircularLinkArc<String>]`, one per line. `lineWidth` is `Self.strongLineWidth` (3) or `Self.weakLineWidth` (1.25), both new static constants on the adapter and both the board's values.
  - `locations: [String: WorkspaceLocation]`, for `open`.
- It keeps the caps (`maxArcs`, `maxNodes`, strong lines first) and the duplicate-id and unknown-pair rules.
- Each node's connected count comes from one `CircularLinkArc.neighborSets(in:)` pass.
- `RecordLinksNode.connected` goes away; only the spoken label used it.
- `PersonRecordLinksColumn` renders `CircularLinkDiagram` with:
  - `style: .standard`
  - `accessibilityLabel: L10n.Conclusions.recordLinksGroupAccessibility(count:)`
  - `accessibilityIdentifierRoot: "persons.detail.sources"`, so node ids are unchanged
  - `diameter: CircularLinkLayout<String>.diameter(width:fitting:style:)`
- The Person previews ("Johann", "One graph", "50 household graphs, Wide", "50 graphs, Wide") move from the old diagram file to `PersonRecordLinksColumn.swift`.
- `PersonSourcesSection`:
  - `hover` becomes `CircularLinkHover<String>`.
  - `open` looks up `links.locations[id]`.
  - The column's `loadingHeight` is `CircularLinkDiagramStyle.standard.minimumDiameter + 2 * CircularLinkDiagramStyle.standard.verticalRoom`, still 232.
- Port the content tests to the new types, keeping every case and number.
- Add `strongAndWeakLinesMapToTheBoardsWidths`.
- `aGraphWithNoLocationIsDrawnAndNotClickable` asserts `isEnabled == false`.

Check:
- The tests pass.
- `rg -n "slot" macos/App/Features/Conclusions/Person*` prints nothing.
- `scripts/check-pr.sh --base origin/s13-27-plan` PASS.

**F-11 — Docs** (user)
Files:
- the plan's Landed note: one paragraph saying the diagram is a generic `CircularLinkDiagram` with a style, under the Person adapter, at the user's request (2026-11-08), and that D-4, D-5 and D-6 name the old types
- the `completed.md` S13-28 paragraph: the type names
- `docs/macos-client-patterns.md`, if it names `RecordLinksDiagram` or `RecordLinksLayout` (`rg` first)

Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-28-how-the-records-connect.md` reports no errors.

### Round 3 — fix
Head `09fa41ee` (commits `96904754`, `09fa41ee`). Gate PASS (orchestrator run, xcode-test 73s). F-9, F-10 and F-11 are done.

Deviations the implementer reported:
- **`Curve` fields:** `Curve` also carries `fromNode` and `toNode`.
- **`LabelSide`:** it is now the top-level `CircularLinkLabelSide`.
- **Empty values:** help and the identifier are set to `""` when nil.
- **Slots:** the layout now numbers a pair's arcs from the arcs it receives. Two cases differ, and only a crafted catalog reaches either: past the cap, and a pair listed twice.

### Round 3 — review
Reviewers: CON, MNT, SEC+PRF. Skipped: UX (no copy and no visible change; CON checked the spoken labels, tooltips and identifiers line by line).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Reversed arcs on one pair draw on top of each other | Fix now → F-12 | verified by arithmetic: at width 340 with 4 nodes, A→B slot 0/2 and B→A slot 1/2 both give control (185.37, 100.63). The offset's normal follows each arc's own direction |
| MNT-3 | MNT | major | Two neighbour functions whose rule differs from the layout's | Fix now → F-12 | `CircularLinkData.swift:27-43` keeps arcs to ids that are not nodes; `CircularLinkLayout.swift:101` skips them |
| MNT-4 | MNT | minor | Duplicate node ids are not handled by the generic diagram | Fix now → F-12 | the layout places every entry, but `ForEach` is keyed by id |
| PRF-5 | PRF | minor | Arc grouping moved into the layout and reruns, string-keyed, on every focus change and resize frame | Fix now → F-12 | measured by PRF: layout init went from 0.027 ms to 1.19 ms at Scale. F-12 resolves the ids once, in the data, and groups on ints. The host still rebuilds the adapter when the Sources list's body reruns; that cost is round 1's, and F-12 shrinks it |
| MNT-1 | MNT | major | Three drawing values are still literals | Fix now → F-13 | `CircularLinkDiagram.swift:221-222, 302, 304` |
| MNT-2 | MNT | major | Callers compute the diagram's size and pass the style twice | Fix now → F-14 | `PersonRecordLinksColumn.swift:46-50`, `PersonSourcesSection.swift:204` |
| CON-3 / MNT-5 | CON, MNT | nit | Empty help and identifier are set where the docs say "none" | Fix now → F-15 | one-line guards in a file this round adds |
| MNT-7 | MNT | nit | A highlighted arc ignores its per-arc colour, and nothing says so | Fix now → F-15 | doc comment |
| CON-4 | CON | nit | No end-to-end check that the Person geometry is unchanged | Fix now → F-15 | Johann control-point test |
| CON-2 | CON | minor | Past `maxArcs`, a pair's slots are the kept arcs' order | Rejected (accepted drift) | only a crafted catalog reaches it; reintroducing caller slots would undo the round's goal. The Landed note records it (F-15) |

Nits (logged): `MNT-6 · PersonRecordLinksContentTests · strength recovered from lineWidth == strongLineWidth`

### Fix plan 4

**F-12 — A dataset value that resolves the arcs once** (from CON-1, MNT-3, MNT-4, PRF-5)
Files:
- `CircularLinkDiagram/CircularLinkData.swift`
- `CircularLinkDiagram/CircularLinkLayout.swift`
- `CircularLinkDiagram/CircularLinkDiagram.swift`
- `PersonRecordLinksContent.swift`
- `PersonRecordLinksColumn.swift`
- `macos/ProvenenciaTests/CircularLinkLayoutTests.swift`
- `PersonRecordLinksContentTests.swift`

Failing tests first:
- `CircularLinkLayoutTests.reversedArcsOnOnePairFanApart`: arcs A→B and B→A on 4 nodes at width 340 have control points 7pt apart.
- `CircularLinkDataTests.duplicateNodeIDsKeepTheFirst`.
- `CircularLinkDataTests.neighborsSkipArcsToNoNode`: an arc A→Z, where Z is not a node, gives A no neighbours.
- `CircularLinkDataTests.slotsAreAssignedPerUnorderedPairInInputOrder`, moved from the layout tests.
- Fix the comment in `arcsOnOnePairFanInInputOrder`.

Do:
- Add `struct CircularLinkDataset<ID: Hashable>: Equatable` in `CircularLinkData.swift`, built with `init(nodes: [CircularLinkNode<ID>], arcs: [CircularLinkArc<ID>])`. It is the one place the arcs are resolved:
  - It drops duplicate node ids, keeping the first.
  - It resolves each arc to node indexes, skipping an arc whose end is not a node or whose ends are the same node. This is the one skip rule.
  - It groups arcs by unordered index pair, keyed on ints, giving each its slot and slot count in input order.
  - It records each pair's reference direction, the first arc's from → to. Each resolved arc stores `flipped: Bool` so the layout takes the offset's normal from the reference direction.
  - It builds `neighbors: [Set<Int>]` (index adjacency) in the same pass.
- It exposes:
  - `nodes`
  - `resolvedArcs`, each with `from`, `to`, `slot`, `slots`, `flipped`, the source arc's `lineWidth` and `color`, and its index into the input arcs
  - `neighbors(ofIndex:)`
  - `neighborCount(of id: ID)`
- Delete `CircularLinkArc.neighbors(of:in:)` and `neighborSets(in:)`.
- `CircularLinkLayout` becomes geometry only: `init(dataset:width:diameter:style:)`, with no dictionaries and no grouping.
- `CircularLinkDiagram` takes `dataset: CircularLinkDataset<ID>` in place of `nodes` and `arcs`. Its highlight reads `dataset.neighbors(ofIndex:)`.
- `PersonRecordLinksContent` builds one dataset in its init (caps first, as today) and stores it as `dataset`. Its connected count reads `dataset.neighborCount(of:)`. The column passes `content.dataset`.

Check:
- The tests pass, and every old layout and Johann number is unchanged.
- `rg -n "neighborSets|neighbors\(of:" macos/App` prints nothing.

**F-13 — The last literals become style props** (from MNT-1)
Files: `CircularLinkDiagramStyle.swift`, `CircularLinkDiagram.swift`.
Do:
- Add three props, with defaults equal to today's values:
  - `overlayLabelPadding: CGFloat = 3`
  - `overlayLabelCornerRadius: CGFloat = PVRadius.sm`
  - `nodeCornerRadius: CGFloat = PVRadius.md`
- The view reads them. `nodeCornerRadius` goes into the node button style for both the hover fill and `pvFocusRing`.

Check:
- `rg -n "PVRadius|PVSpacing|PVColor|PVFont|PVTypeScale" macos/App/Features/Conclusions/CircularLinkDiagram/CircularLinkDiagram.swift` prints nothing. The tokens live only in the style's defaults.
- A scan for bare numeric literals in the diagram file prints only 0, 1, 2 and the opacity arithmetic.

**F-14 — The diagram sizes itself** (from MNT-2)
Files: `CircularLinkDiagram.swift`, `CircularLinkDiagramStyle.swift`, `PersonRecordLinksColumn.swift`, `PersonSourcesSection.swift`.
Do:
- Replace the `diameter: CGFloat?` prop with `fittingHeight: CGFloat? = nil`. The diagram calls `CircularLinkLayout.diameter(width:fitting:style:)` with its own style, and handles a single node itself.
- Add a computed `var minimumHeight: CGFloat { minimumDiameter + 2 * verticalRoom }` on the style.
- The column passes `fittingHeight: fitting` and drops `single ? nil :` and its second `.standard`.
- The section's `loadingHeight` reads `CircularLinkDiagramStyle.standard.minimumHeight`, which is 232.

Check:
- `rg -n "CircularLinkLayout" macos/App/Features/Conclusions/Person*` prints nothing.
- `check-pr.sh` PASS.

**F-15 — Nits and notes** (from CON-3, MNT-5, MNT-7, CON-4, CON-2)
Do:
- Apply `.help` only when `help` is non-nil, and `.accessibilityIdentifier` only when the root is non-nil. Use a small `@ViewBuilder` or modifier; do not pass empty strings.
- Doc comment on `CircularLinkArc.color`: a highlighted arc draws in `style.highlightedArcColor`.
- Add `CircularLinkLayoutTests.johannGeometryIsUnchanged`. Take the Johann dataset at width 340 and assert the two-arc pairs' control points. Compute them from the f4e8589a formula with slot 0/1 and 1/2, and write them down as literals.
- The plan's Landed note gets two lines:
  - The dataset type.
  - Past the 75,000-arc cap, a pair's slots count only the kept arcs (CON-2, accepted).

Check: the tests pass; `scripts/check-pr.sh --base origin/s13-27-plan` PASS.
