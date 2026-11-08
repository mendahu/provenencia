# Ship log — S13-28 How the records connect

**Plan:** [`S13-28-how-the-records-connect.md`](S13-28-how-the-records-connect.md) · **Branch:** `s13-28-how-the-records-connect` (stacked on `s13-27-plan`, #743) · **PR:** #744
**State:** ready

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
Gate PASS (`f3e2e17a`) · CI see PR · open: none · follow-ups: PRF-4, PRF-3 runtime check (manual), CON-8 doc line · to user: —
