# S14-29 — Polygon locator vertex edit

**Status:** Landed in #715 (2026-11-05)
**Landed:** Product version is 0.3.3. `ArtifactRegionTarget` is a `typealias` for the planned tuple. `ArtifactFileFixtures.pdfHost` defaults `onPageChange` to a no-op. Manual check §10 not run in implementer pass.
**Spike / slice:** Spike 14 · Slice 7b — Region edit
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: handles, the move diamond and the stroke already draw today; on a PDF they already draw at a fixed screen size, and this PR gives image Artifacts the same size — user, 2026-11-05, "Fixed size, no brief")
**Board:** —
**Depends on:** S14-28 (landed [#711](https://github.com/mendahu/provenencia/pull/711))
**Blocks:** —
**Branch:** `s14-29-polygon-locator-vertex-edit`
**Version bump:** patch — researcher-visible fix to region editing ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (classification only: `ArtifactRegionOverlayView` stays a feature snowflake)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 7) Polygon edit includes a saved Citation"; bar outcome 24; dogfood note `2026-10-15-polygon-locator-drag-vertices` (deleted on land); S14-28's Landed note
**Planned:** 2026-11-05 against `main` @ `124d492`
**Review:** [`S14-29-polygon-locator-vertex-edit.review.md`](S14-29-polygon-locator-vertex-edit.review.md) — Fits with changes — Ready with changes, 2 rounds; no open blockers; To user none

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

After a freeform polygon Citation region is drawn, the researcher can grab a vertex and nudge it, or grab the outline and move the whole shape. This works on a new Citation and on a saved Citation reopened in the composer, on a PDF and on an image, at any zoom. Today the plumbing exists (`moveFreeformVertex`, `translate`, the `.resize` / `.move` drag in `ArtifactRegionOverlayView`), but on a PDF the overlay's `hitTest` reads its point in the wrong coordinate space, so a click on a handle usually misses, and on an image the hit radius shrinks with zoom.

**Why now.** Batch 7 dogfood note (2026-10-15). S14-28 made `model.page` honest, so a region drawn on a later PDF page is no longer stamped page 1 and hidden.

**What stays.** The look of handles, the move diamond and the stroke at 1×, and on every PDF. Right-click deletes a freeform vertex; a delete refused by `deleteFreeformVertex` (3 points left) stays silent, as today. The armed-tool drawing paths. The locator model (`CitationLocatorDraft.setRegion` restamps the viewer page). Depiction regions stay rectangle-only (S14-03). Go accepts the same regions.

## 2. User stories

**US-1.** As a researcher with a committed freeform region on a PDF page, when I drag a vertex, that vertex follows the pointer and the Citation's region updates on release.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | One-page PDF in `ArtifactPDFHostView` 400×500 in a window; overlay `committed` = triangle (0.1,0.1),(0.9,0.1),(0.5,0.9); point at vertex 0 and at vertex 2, each converted from overlay to host coordinates | `overlayView.hitTest(point)` returns `overlayView` for both (today the top or bottom vertex misses, because y is mirrored) |
| AC-1.2 | Same; the point at normalized (0.5, 0.02), above the top edge, converted the same way | `hitTest` returns nil, so PDFKit keeps selection and scroll |
| AC-1.3 | Vertex 2 dragged to (0.6, 0.7) and released, giving a valid polygon | `onCommit` fires once with points where index 2 == (0.6, 0.7) and the other two are unchanged |

**US-2.** As a researcher, when I drag the outline of a committed region (away from any handle) or the centre diamond, the whole shape moves.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Freeform square (0.2,0.2),(0.4,0.2),(0.4,0.4),(0.2,0.4) in a 100×100 `imageRect`, radius 3; document point (30, 20) (top edge midpoint) | `ArtifactRegionGeometry.hitHandle` returns `.move` |
| AC-2.2 | Same, point (30, 30) (centroid) | `.move` (diamond, as today) |
| AC-2.3 | Drag from the edge by (+0.1, +0.05) | every point moves by (+0.1, +0.05), clamped into the unit square by `translate` |
| AC-2.4 | Circle region; point on its radius handle | `.resize(0)`: a handle beats the stroke it sits on |

**US-3.** As a researcher zoomed out on an image Artifact, handles stay large enough to grab and see.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Overlay `magnification = 0.5` | `handleHitRadius == 14` and `handleDrawSize == 14` (7 screen pt each) |
| AC-3.2 | `magnification = 1` (the default, and always on the PDF host) | `handleHitRadius == 7`, `handleDrawSize == 7` (today's values) |
| AC-3.3 | `RegionFixtures.ring(count: 256)` in a 500×500 rect, radius 7, point on vertex 10 | `hitHandle` returns `.resize(10)` (nearest wins, not the first in range) |

**US-4.** As a researcher reopening a saved Citation, I can nudge its polygon and save the change; a drag that would break the shape is refused.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Composer opened on `cit-1` with a 4-point region locator; `commitViewerRegion` with one vertex moved | returns true; `fields.locator.region` equals the moved region; `fields.isDirty`; `canSaveCitation` |
| AC-4.2 | Square (0.2,0.2),(0.4,0.2),(0.4,0.4),(0.2,0.4); vertex 0 dragged to (0.5, 0.3) (a self-crossing shape) and released | `onCommit` is not called; `committed` equals the start square; the system beep plays (manual check 5) |
| AC-4.3 | Right-click on vertex 0 of (0.1,0.1),(0.9,0.1),(0.9,0.9),(0.5,0.2),(0.1,0.9); the remaining four points cross (edge (0.1,0.9)→(0.9,0.1) meets (0.9,0.9)→(0.5,0.2)) | `committed` keeps the five points; beep; no `onCommit` |
| AC-4.4 | A click on a handle or the outline with no drag | no `onCommit` (today a handle click commits the unchanged draft; with the outline now a target, a stray click must not touch a saved Citation) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `hitTest` converts its superview point into overlay coordinates | STEP-1, AC-1.1, AC-1.2, D-1 |
| One target list and one hit rule in `ArtifactRegionGeometry`, used by `hitTest`, `mouseDown`, cursor rects, `applyCursor` and right-click | STEP-2, AC-2.1, AC-2.2, AC-2.4, AC-3.3, D-2, D-3, D-7 |
| Outline drag translates the shape | STEP-2, AC-2.3, D-3 |
| Overlay `magnification` input (one writer); sizes in screen points | STEP-3, AC-3.1, AC-3.2, D-4 |
| One edit state; invalid edit reverts and beeps (drag and right-click); unchanged edit does not commit | STEP-4, AC-4.2, AC-4.3, AC-4.4, D-5 |
| Shared AppKit-host test helpers | STEP-1, D-8 |
| Saved-Citation edit pinned by a model test | STEP-5, AC-4.1, D-6 |
| Patch bump | STEP-6 |
| Docs, map scope entry, UX note | STEP-7 |

| Out | Home |
| --- | --- |
| New handle kinds or editing behaviour for rectangles, L-shapes and circles beyond the shared hit rule (which also lets their outline move them) | not planned (map scope entry) |
| Adding a vertex by clicking an edge | dogfood log |
| Depiction regions (rectangle-only) | not planned (S14-03 rule) |
| Keyboard nudging of vertices; VoiceOver per-vertex elements | dogfood log |
| Undo of a region edit | not planned |
| New handle look, hover state or colours | not planned (no brief) |
| Moving `GraphCanvasCoordinates.screenPixelDocumentWidth` to a shared module | not planned (one caller outside GraphCanvas) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Hit claim | `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` `ArtifactRegionOverlayView.hitTest(_:)` | Passes `point` straight to `hitHandle(at:)` and `isNearCommittedStroke(_:)`. AppKit gives `hitTest` a point in the **superview's** coordinates. In `ArtifactPDFHostView` (not flipped) the overlay is flipped and fills the host, so y is mirrored and handle hits miss. In `ArtifactMediaDocumentView` (flipped, overlay frame = bounds) the two spaces match, so images work |
| Handle hit | same file `hitHandle(at:)` | Move anchor first (radius `handleHitSize + 1`), then the first resize handle within `handleHitSize`. First-in-range, not nearest |
| Vertex hit for delete | same file `freeformVertexIndex(at:)` | Same loop over `committed.points`; ignores the anchor |
| Stroke | same file `isNearCommittedStroke(_:)`, private `distance(_:segmentFrom:to:)` | Claims the event in `hitTest`, but `mouseDown` only acts on `hitHandle`, so an outline click is swallowed and does nothing. Uses `committed.points` (a 32-point ring for a circle) |
| Cursor | same file `resetCursorRects()` (own loop over `moveAnchor` + `handlePoints`, `pad` 8) and `applyCursor(at:)` (uses `hitHandle`) | Two lists of targets |
| Sizes | same file `handleHitSize` 7; `drawHandle(at:)` 7, line 1.5; `drawMoveAnchor(at:)` 9, line 1.25 | In overlay coordinates. Under `NSScrollView` magnification (images) that is size × magnification on screen; minimum zoom is `ArtifactViewerModel.minZoom` 0.5 |
| Drag state | same file `activeHandle`, `isEditingCommitted`, `moveDragOrigin`, `moveDragPoints` | Four fields set together in `mouseDown`, cleared in `mouseUp`. `mouseUp` calls `onCommit` only if the draft `isValid`; an invalid draft stays drawn in `committed` though the model still holds the old one. `rightMouseDown` assigns `committed` before checking `isValid` |
| Handle enum | same file `enum ArtifactRegionHandle` | Declared in the view file; geometry value types (`ArtifactResizeCursorKind`, `LParams`) live in `ArtifactRegionGeometry.swift` |
| Move anchor | `ArtifactRegionGeometry.swift` `moveAnchor(for:)` | `circleCenter(points)`, the centroid, for **every** kind. Freeform already has the diamond (the UX note says "circle-only"; that is not true at this base) |
| Geometry | same file `moveFreeformVertex`, `translate`, `deleteFreeformVertex`, `handlePoints(for:imageRect:)`; private `distance(_ a: [CGPoint], _ b: [CGPoint])` | Pure; `handlePoints` for `.freeform` is the points themselves |
| Screen-point math | `macos/App/Features/GraphCanvas/GraphCanvasCoordinates.swift` `screenPixelDocumentWidth(magnification:)` | `1 / max(magnification, minimumMagnification)`; the Evidence graph's one-screen-point rule (`EvidenceGraphEdgeLayer`, `GraphCanvasGridView`) |
| PDF host | `ArtifactPDFViewport.swift` `ArtifactPDFHostView` | `isFlipped == false`; `overlayView` pinned to all four edges; `layoutPageOverlay` sets `mediaRectOverride` from the current page (`.zero` with no page). Never magnifies the overlay |
| Image host | `ArtifactMediaViewport.swift` `ArtifactMediaDocumentView`, `ArtifactMediaNSScrollView.magnification` `didSet`, `ArtifactMediaViewport.configureOverlay` | Document flipped; overlay `frame = bounds`; zoom is scroll-view magnification |
| Saved Citation | `CitationComposerModel.swift` `loadCitation` → `fields.replace(from:)`; `CitationComposerModel+Depiction.swift` `viewerRegion`, `commitViewerRegion` | A loaded Citation's locator goes into `fields.locator`, `viewerRegion` returns `locator.region`, and `commitViewerRegion` → `setRegion` writes it back. So the saved locator **is** already fed as `committedRegion`; no test pins edit-after-load |
| Test helpers | `macos/ProvenenciaTests/HostedWindow.swift` `hostInBorderlessWindow(_:windowSize:hostSize:orderFront:)` (SwiftUI only); `ArtifactPDFViewportPageTests.swift` private `bootstrapHost(document:onPageChange:)`, `hostWindow(for:)`; `ArtifactFileFixtures.blankPDFDocument(pageCount:)` | The PDF host bootstrap is private to one file |
| Tests | `ArtifactRegionGeometryTests.swift` `translateClampsToUnitSquare`, `freeformKeepsAtLeastThreeVertices`; `CitationComposerModelTests.swift` `editPrepareKeepsRegionLocatorAndIsNotDirty`, private `regionLocatorJSON`, `seedArtifact`, `seedCitation` | No hit-test, screen-size or edit-after-load test |

**Precedent.** `af433b89` ("Fix region editing and match the locator bar to the board") last changed the edit path in this overlay; copy its shape (geometry helper in `ArtifactRegionGeometry`, overlay calls it). The screen-size rule copies the Evidence graph's use of `GraphCanvasCoordinates.screenPixelDocumentWidth` (`EvidenceGraphEdgeLayer`). Host tests build on S14-28's (#711) bootstrap, moved to a shared helper.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `hitTest(_ point:)` first does `let local = convert(point, from: superview)` and uses `local` | AppKit's contract; fixes the PDF host, no-op for the image host | Flipping the PDF host (changes PDFKit layout and `layoutPageOverlay`); special-casing the PDF path |
| D-2 | Move `enum ArtifactRegionHandle` into `ArtifactRegionGeometry.swift`. Add to `ArtifactRegionGeometry` *(new)*: `handleTargets(for draft:, imageRect:) -> [(handle: ArtifactRegionHandle, normalized: CGPoint, document: CGPoint)]` (the move anchor first as `.move`, then `handlePoints` as `.resize(i)`; hit math uses `document`, the cursor kind uses `normalized`); `nearestHandle(at:in targets:, radius:) -> ArtifactRegionHandle?` (the nearest `.resize` target within `radius`); `hitHandle(at documentPoint:, draft:, imageRect:, radius:) -> ArtifactRegionHandle?`: anchor within `radius + 1` → `.move`; else `nearestHandle`; else the polygon of `documentPoints(normalized: draft.points…)` within `radius` of any edge → `.move`; else nil. The segment helper moves in as `private static func distanceToSegment(_:from:to:)` (geometry already has a `distance` with another meaning) | One list of targets and one rule; nearest-wins for dense freehand rings; each caller takes the part it means | Three loops in the view; first-in-range |
| D-3 | An outline hit is `.move` for every kind (user, 2026-11-05). It reuses the `.move` drag. For a circle the outline is its stored 32-point ring, which lies on the drawn oval; a handle beats the outline it sits on (D-2 order) | The outline is a larger target than the diamond; nothing new is drawn; the stroke claim already swallowed those clicks | A new `.translate` case; restricting it to freeform |
| D-4 | `ArtifactRegionOverlayView` gets a declared input `var magnification: CGFloat = 1` *(new)* with the same `didSet` as `imageSize` (`needsDisplay`, `invalidateCursorRects`). One writer: `ArtifactMediaNSScrollView.layoutDocumentCentered(resetScroll:allowScrollAdjust:)` sets `document.overlayView.magnification = magnification` right after its `guard let document` (every zoom path ends there: the setter's `didSet`, wheel zoom, pinch through `clipBoundsDidChange`, `updateNSView`); the PDF host leaves it at 1. Let `pixel = GraphCanvasCoordinates.screenPixelDocumentWidth(magnification: magnification)`; internal `handleHitRadius = 7 * pixel` (replaces `handleHitSize`), internal `handleDrawSize = 7 * pixel` (used by `drawHandle`, so committed and in-progress handles both follow), diamond `9 * pixel`, handle line `1.5 * pixel`, diamond line `1.25 * pixel`. The stroke width (2) is unchanged | Explicit and testable like the overlay's other inputs; one screen-point formula in the app; on a PDF nothing changes; on images it matches the PDF look (user, 2026-11-05) | Reading scale from `convert(_:to: nil)` (implicit, needs a window, second formula); hit radius only; 44 pt `PVSpacing.hitMin` (swallows PDF clicks) |
| D-5 | One private edit state replaces `activeHandle`, `isEditingCommitted`, `moveDragOrigin`, `moveDragPoints`: `private var edit: (handle: ArtifactRegionHandle, origin: CGPoint, start: ArtifactRegionDraft)?`. `.move` translates `edit.start.points` by (pointer − `origin`). Add `private func finishEdit(_ draft: ArtifactRegionDraft, start: ArtifactRegionDraft)`: `draft == start` → nothing; valid → `committed = draft`, `onCommit(draft)`; invalid → `committed = start`, `NSSound.beep()`, no `onCommit`. `mouseUp` calls `finishEdit(committed, start: edit.start)` when `edit != nil` and `committed` is non-nil, then sets `edit = nil`; `rightMouseDown` computes the next points; nil from `deleteFreeformVertex` returns silently (as today); otherwise it calls `finishEdit` with the current draft as `start`. `draw` uses `edit != nil` where it read `isEditingCommitted` | One revert rule, no impossible combinations; beep is the overlay's existing refusal (cap reached); skipping an unchanged commit keeps a stray click from dirtying a saved Citation | A fifth `editStartDraft` field; committing invalid shapes; a new L10n announcement |
| D-6 | No composer model change. Edit-after-save is pinned by T-8 only | §4: `viewerRegion` already returns the loaded locator | A parallel `committedRegion` path in the model |
| D-7 | `resetCursorRects` iterates `handleTargets` with pad `handleHitRadius` (`.move` → open hand; `.resize` → `resizeCursorKind(handle: target.normalized, centroid:)` as today) and drops the `pad` 8 constant. `freeformVertexIndex(at:)` calls `nearestHandle` over the `.resize` targets (no anchor, no outline, as today). `applyCursor` keeps calling `hitHandle`, so the outline shows the open hand on mouse-move | Cursor and hit read one list; right-click can still delete a vertex near the centroid of a small shape | Keeping a second loop |
| D-8 | Test helpers: add `hostInBorderlessWindow(nsView: NSView, size: CGSize) -> NSWindow` *(new overload)* to `HostedWindow.swift`; move `bootstrapHost` out of `ArtifactPDFViewportPageTests` to `ArtifactFileFixtures.pdfHost(document:onPageChange:) -> ArtifactPDFHostView` *(new, @MainActor)*; add `ArtifactFileFixtures.sizedPDFDocument(size:) -> PDFDocument` *(new)* (one `PDFPage()` with `setBounds(CGRect(origin: .zero, size: size), for: .mediaBox)`); `ArtifactPDFViewportPageTests` uses both and deletes its private copies | A third hand-rolled window bootstrap is due for extraction | Copying `hostWindow` into the new file |

**Rules this plan bends.** None. The working rule says "an open Citation whose locator is fed back as `committedRegion`"; D-6 records that it already is, T-8 holds it, and the map's scope entry says so.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ArtifactRegionGeometry.handleTargets`, `nearestHandle`, `hitHandle` *(new)* | func | `ArtifactRegionOverlayView.hitHandle(at:)`, `freeformVertexIndex(at:)`, `isNearCommittedStroke(_:)`, the loop in `resetCursorRects` | Replaces those four loops with one pure list and rule beside `handlePoints` and `moveAnchor` |
| `ArtifactRegionOverlayView.magnification` *(new input)* | property | `imageSize`, `mediaRectOverride` inputs; `GraphCanvasCoordinates.screenPixelDocumentWidth` | Same input pattern; sizes reuse the graph's formula |
| `edit` state, `finishEdit` *(new private)* | property / func | `activeHandle`, `isEditingCommitted`, `moveDragOrigin`, `moveDragPoints` | Replaces the four fields |
| `hostInBorderlessWindow(nsView:size:)`, `ArtifactFileFixtures.pdfHost`, `ArtifactFileFixtures.sizedPDFDocument` *(new)* | helper | `hostInBorderlessWindow` (SwiftUI), private `bootstrapHost` / `hostWindow` | Extends the shared helper file; moves the private copy |
| `ArtifactRegionOverlayTests.swift` *(new)* | file | `ArtifactPDFViewportPageTests.swift` | That file is about page feedback; overlay hits are a separate suite |

**Builds on.** `ArtifactRegionGeometry`; `GraphCanvasCoordinates.screenPixelDocumentWidth`; `af433b89`'s edit path; S14-28's host bootstrap; `RegionFixtures.ring(count:)`.
**Deletes.** `isNearCommittedStroke(_:)`, the view's `distance(_:segmentFrom:to:)`, the loop bodies of `hitHandle(at:)` and `freeformVertexIndex(at:)`, the `pad` 8 constant and the separate loop in `resetCursorRects`, `handleHitSize`, `activeHandle`, `isEditingCommitted`, `moveDragOrigin`, `moveDragPoints`; in tests, the private `bootstrapHost` and `hostWindow` in `ArtifactPDFViewportPageTests`. The UX note `ux/2026-10-15-polygon-locator-drag-vertices.md` on landing.
**Next adds.** Insert-vertex-on-edge is one more target kind in `handleTargets` and one case in `hitHandle`, and the cursor follows.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Clicks on a PDF region now reach the overlay | `mouseDown` edit path | Only within the radius of a target or the outline; elsewhere PDFKit still gets the click (AC-1.2) |
| Edit of a loaded Citation's region | `commitViewerRegion` → `setRegion` → `fields` dirty | Existing Save and discard prompts apply (`hasUnsavedDocumentWork`); restamps the viewer page, which equals the locator page because the overlay only shows on it (`committedOnCurrentPage`) |

### Scale

None — no catalog write. `hitHandle` and `handleTargets` are O(points) per mouse event, points ≤ `maxRegionPoints` (256).

### Edges

- **Assumptions.** `handlePoints(.freeform) == points`, so `.resize(i)` is a vertex index (T-3). Magnification is uniform in x and y (`NSScrollView` magnification is).
- **Neighbors.** Anchor vs. a vertex near the centroid of a small shape: drag takes the anchor (T-2), right-click takes the vertex (T-3b). Circle radius handle on the outline: the handle wins (AC-2.4). Rectangle, L and circle outlines now move their shape (D-3). Armed tool: `hitTest` still returns self first.
- **Removed inputs.** A drag or delete that gives an invalid shape restores the start shape (T-6).
- **Real data.** None — no table or seed.
- **Untrusted size.** A loaded locator is bounded by `maxRegionPoints` through `isValidPolygon`; the hit loop is linear in it.

### Schema

None — no catalog change.

### Go core

None — macOS overlay only.

### FFI

None — no RPC.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Region overlay | `ArtifactRegionOverlayView` (feature snowflake) | `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` | D-1, D-3 – D-5, D-7 |
| Targets and hit rule | `ArtifactRegionGeometry` | `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift` | D-2 |
| Zoom input | `ArtifactMediaNSScrollView.layoutDocumentCentered` | `macos/App/Features/ArtifactViewer/ArtifactMediaViewport.swift` | D-4 |

States: no region → `hitTest` nil (unchanged). Accessibility: unchanged; the overlay has no per-vertex elements (Out).

### Strings

None — no copy changes. The tooltip `L10n.CitationComposer.freeformDeleteVertex` still describes right-click delete.

## 7. Steps

**STEP-1 — Test helpers, a red PDF hit test, then convert the point**
Files: `macos/ProvenenciaTests/HostedWindow.swift`, `macos/ProvenenciaTests/ArtifactFileFixtures.swift`, `macos/ProvenenciaTests/ArtifactPDFViewportPageTests.swift`, `macos/ProvenenciaTests/ArtifactRegionOverlayTests.swift` *(new)*, `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift`. Skill: `add-swift-test`.
Do: D-8 helpers; switch `ArtifactPDFViewportPageTests` to them. Write T-4 and run it: it must fail on the vertex that sits far from the vertical middle. Then apply D-1.
Check: T-4 fails before D-1 and passes after; `ArtifactPDFViewportPageTests` still passes.

**STEP-2 — One target list and hit rule in geometry; outline moves the shape**
Files: `ArtifactRegionGeometry.swift`, `ArtifactRegionOverlay.swift`, `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift`. Skill: `add-swift-test`.
Do: D-2, D-3, D-7. `hitTest` returns self when `hitHandle` is non-nil (drop the separate stroke check). Delete `isNearCommittedStroke` and the view's `distance`. Write T-1 – T-3b.
Check: T-1 – T-3b and T-4 pass; `rg 'isNearCommittedStroke|segmentFrom' macos/` has no hits; `rg 'enum ArtifactRegionHandle' macos/App` hits only `ArtifactRegionGeometry.swift`.

**STEP-3 — Magnification input and screen-point sizes**
Files: `ArtifactRegionOverlay.swift`, `ArtifactMediaViewport.swift`, `macos/ProvenenciaTests/ArtifactRegionOverlayTests.swift` *(new)*.
Do: D-4. Write T-7.
Check: T-7 passes; `rg 'handleHitSize' macos/` has no hits; `rg 'overlayView.magnification' macos/App` hits exactly once, in `layoutDocumentCentered`.

**STEP-4 — One edit state; revert an invalid edit**
Files: `ArtifactRegionOverlay.swift`, `macos/ProvenenciaTests/ArtifactRegionOverlayTests.swift` *(new)*.
Do: D-5 in `mouseDown`, `mouseDragged`, `mouseUp`, `rightMouseDown`, `draw`. Write T-5 and T-6.
Check: T-5 and T-6 pass; `rg 'moveDragPoints|moveDragOrigin|isEditingCommitted|activeHandle' macos/App` has no hits.

**STEP-5 — Saved-Citation test and the full suites**
Files: `macos/ProvenenciaTests/CitationComposerModelTests.swift`. Skill: `add-swift-test`.
Do: T-8.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/ArtifactRegionGeometryTests -only-testing:ProvenenciaTests/ArtifactRegionOverlayTests -only-testing:ProvenenciaTests/ArtifactPDFViewportPageTests -only-testing:ProvenenciaTests/CitationComposerModelTests` passes.

**STEP-6 — Version bump**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (all four `MARKETING_VERSION` lines: two targets × Debug and Release).
Do: patch bump together (`0.3.2` → `0.3.3` at the planning base, or the next patch of what `VERSION` reads).
Check: `rg 'MARKETING_VERSION' macos/Provenencia.xcodeproj/project.pbxproj` shows the new value four times; `VERSION` and `core/version.go` match.

**STEP-7 — Docs**
Files: this plan (status, Landed note), `docs/deployment-plan/spike-14/deployment-plan.md`, `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/ux/2026-10-15-polygon-locator-drag-vertices.md`.
Do: per the Definition of done. In the map: the row status (the scope entry was corrected when this plan was written). Delete the UX note and replace its links as `7e69154a` did for S14-30 (the map's dogfood table row becomes plain text with "note deleted YYYY-MM-DD; landed #N"; the scope **Context** and this plan's **Context** drop the link).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-29-polygon-locator-vertex-edit.md` passes; `rg 'polygon-locator-drag-vertices' docs/` hits nothing but this plan's plain-text mention and its review log.

Commits: one per behavior. Suggested: "Let clicks on a PDF region reach its handles."; "Move a region by dragging its outline and pick the nearest vertex."; "Keep region handles the same size on screen at any zoom."; "Put a region back when an edit would make it invalid."; "Cover editing a saved Citation's region."; "Bump version to 0.3.3."

## 8. Tests

Overlay tests drive the handlers through one private helper in `ArtifactRegionOverlayTests.swift`: `event(_ type: NSEvent.EventType, at normalized: CGPoint, in host: ArtifactPDFHostView, window: NSWindow) -> NSEvent` = `NSEvent.mouseEvent(with: type, location: host.overlayView.convert(ArtifactRegionGeometry.documentPoint(normalized: normalized, imageRect: media), to: nil), modifierFlags: [], timestamp: 0, windowNumber: window.windowNumber, context: nil, eventNumber: 0, clickCount: 1, pressure: 1)!`, where `media` is `host.overlayView.mediaRectOverride`. A drag is one `.leftMouseDown` at the start, one `.leftMouseDragged` at the end, one `.leftMouseUp` at the end, each passed to the overlay's `mouseDown(with:)` / `mouseDragged(with:)` / `mouseUp(with:)` directly. A right-click is one `.rightMouseDown` passed to `rightMouseDown(with:)`. The host is `pdfHost(document: sizedPDFDocument(size: CGSize(width: 300, height: 400)))` in `hostInBorderlessWindow(nsView:size: CGSize(width: 400, height: 500))`, then `layoutSubtreeIfNeeded()`.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `ArtifactRegionGeometryTests.swift` `hitHandleOutlineMovesShape` *(new)* | square (0.2,0.2),(0.4,0.2),(0.4,0.4),(0.2,0.4) in a 100×100 `imageRect`, radius 3: (30, 20) → `.move`; (30, 30) → `.move`; (20, 20) → `.resize(0)`; (60, 60) → nil; `handleTargets` lists `.move` first, then `.resize(0)…` | AC-2.1, AC-2.2, D-2, D-3 |
| T-1b | Swift | same `hitHandleCircleHandleBeatsOutline` *(new)* | `circleRing(center: (0.5,0.5), radius: 0.2)` in a 100×100 rect, radius 3; point on `handlePoints(…)[0]` → `.resize(0)`; point on the ring a quarter turn away → `.move` | AC-2.4, D-3 |
| T-2 | Swift | same `hitHandleAnchorWinsOverNearVertex` *(new)* | triangle (0.50,0.50),(0.52,0.50),(0.51,0.52) in a 100×100 rect, radius 3; point at the centroid → `.move` | D-2 |
| T-3 | Swift | same `hitHandlePicksNearestVertex` *(new)* | `RegionFixtures.ring(count: 256)` in a 500×500 rect, radius 7; point on vertex 10 → `.resize(10)` | AC-3.3 |
| T-3b | Swift | same `nearestHandleIgnoresAnchor` *(new)* | T-2's triangle; `nearestHandle` over its `.resize` targets at vertex 0 → `.resize(0)` | D-7 |
| T-4 | Swift | `ArtifactRegionOverlayTests.swift` `pdfHostHitTestFindsVertex` *(new, @MainActor)* | the host above; `committed` = AC-1.1 triangle; for vertex 0 and vertex 2, `p = host.convert(ArtifactRegionGeometry.documentPoint(normalized: v, imageRect: media), from: host.overlayView)` → `host.overlayView.hitTest(p) === host.overlayView`; normalized (0.5, 0.02) converted the same way → nil | AC-1.1, AC-1.2, D-1 |
| T-5 | Swift | same `vertexDragCommitsOnce` *(new)* | `onCommit` appends to a local array. Triangle; drag (0.5,0.9) → (0.6,0.7) → one commit, index 2 == (0.6,0.7). Square (0.2,0.2)…(0.4,0.4); drag (0.3,0.2) → (0.4,0.25) → one commit, every point moved by (+0.1,+0.05). Click with no drag at (0.2,0.2) (down and up at the same point) → no commit | AC-1.3, AC-2.3, AC-4.4 |
| T-6 | Swift | same `invalidEditRevertsAndDoesNotCommit` *(new)* | AC-4.2 square, drag (0.2,0.2) → (0.5,0.3) → `committed` equals the start square, zero commits. AC-4.3 five-point shape, right-click at (0.1,0.1) → `committed` keeps the five points, zero commits. (0.2,0.2),(0.4,0.2),(0.4,0.4),(0.2,0.4),(0.1,0.3), right-click at (0.1,0.3) → one commit with the square | AC-4.2, AC-4.3, D-5 |
| T-7 | Swift | same `handleSizesAreInScreenPoints` *(new)* | bare `ArtifactRegionOverlayView`; `magnification = 0.5` → `handleHitRadius == 14`, `handleDrawSize == 14`; `magnification = 1` → 7, 7 | AC-3.1, AC-3.2, D-4 |
| T-8 | Swift | `CitationComposerModelTests.swift` `editSavedCitationRegionMarksDirty` *(new)* | `seedArtifact`; `seedCitation(locatorJSON:)` with `regionLocatorJSON`'s selectors plus `{"type":"page","artifact_page":1}`; `prepare()`; `fields.isDirty == false`; `commitViewerRegion(ArtifactRegionDraft(kind: .freeform, points: [CGPoint(x: 0.1, y: 0.1), CGPoint(x: 0.45, y: 0.1), CGPoint(x: 0.4, y: 0.4), CGPoint(x: 0.1, y: 0.4)]))` → true; `fields.locator.page == 1` (the page did not change); `fields.locator.region?.points[1] == CGPoint(x: 0.45, y: 0.1)`; `fields.isDirty`; `canSaveCitation` | AC-4.1, D-6 |

The beep (AC-4.2, AC-4.3) is checked by hand (manual check 5). Must stay green: everything `scripts/check-pr.sh` runs; `ArtifactPDFViewportPageTests` (helpers moved), `CitationLocatorDraftTests`, `CitationComposerDepictionTests` (rectangle-only depiction drawing shares the overlay).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `handleHitSize`, `isNearCommittedStroke`, `moveDragPoints`, `isEditingCommitted` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/` (spike folders included) and `L10n.swift` with no hit left outside this plan and its review log; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive (no brief: handles look as on a PDF today)
- [ ] Version bump as the header says (VERSION, core/version.go, all four MARKETING_VERSION lines together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped; UX note deleted and its links replaced (STEP-7)
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: none owed
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open the Citation composer on a multi-page PDF Artifact. Scroll to page 3, arm the polygon tool, draw a five-point shape, close it.
2. Drag a vertex near the top of the page and one near the bottom (AC-1.1). Each follows the pointer; the locator chip still reads five points.
3. Drag the outline between two vertices (AC-2.3): the whole shape moves. Drag the centre diamond: it moves too.
4. Click well away from the shape: PDF text selection still works (AC-1.2).
5. Drag a vertex across the opposite edge so the shape crosses itself; release (AC-4.2): beep, and the shape returns. Draw a concave shape and right-click the vertex whose removal would make it cross itself (AC-4.3): beep, and the shape stays.
6. Save. Leave and reopen that Citation; nudge a vertex; Save is enabled; save, reopen, the change is kept (AC-4.1).
7. Open an image Artifact, zoom to the minimum (50%) with the chrome, then by pinch, then by mouse wheel, and grab a vertex each time (AC-3.1): handles look the same size as at 100% and are easy to hit.
8. Click once on a saved Citation's outline without dragging (AC-4.4): Save stays disabled.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| PDFKit's `PDFView` takes the mouse-down before the overlay even after D-1 | The overlay is the later sibling, so the host's `hitTest` asks it first; T-4 and manual check 2 confirm |
| Synthesized `NSEvent`s in T-5 do not reach the overlay | T-5 calls the overlay's `mouseDown(with:)` / `mouseDragged(with:)` / `mouseUp(with:)` directly |
| A blank `PDFPage()` has no useful media box | D-8's `sizedPDFDocument` sets a 300×400 media box |

**Stop and report if:**

- T-4 passes before D-1 in STEP-1 (then the y-mirror is not the PDF cause; report what does block handle clicks).
- `mediaRectOverride` is not 300×400 (± 1) after layout in T-4 at zoom 1 (report its value).
- `viewerRegion` does not return a loaded Citation's region (T-8 fails with no model change): D-6 is wrong.
- `commitViewerRegion` returns false, or `fields.locator.page` is not 1, in T-8 because the fake PDF Artifact refuses a region (`locatorCapabilities`); report the capability values.
- Keeping PDF clicks working needs the outline claim dropped (AC-1.2 against D-3).

## 12. Open questions

None.

## 13. PR

**Title:** S14-29 — Polygon locator vertex edit
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-29-polygon-locator-vertex-edit.md`.
