# S14-29 — Polygon locator vertex edit

**Status:** Draft
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
**Context:** working rule "(Batch 7) Polygon edit includes a saved Citation"; bar outcome 24; [`ux/2026-10-15-polygon-locator-drag-vertices.md`](../ux/2026-10-15-polygon-locator-drag-vertices.md); S14-28's Landed note
**Planned:** 2026-11-05 against `main` @ `124d492`
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

After a freeform polygon Citation region is drawn, the researcher can grab a vertex and nudge it, or grab the outline and move the whole shape. This works on a new Citation and on a saved Citation reopened in the composer, on a PDF and on an image, at any zoom. Today the plumbing exists (`moveFreeformVertex`, `translate`, the `.resize` / `.move` drag in `ArtifactRegionOverlayView`), but on a PDF the overlay's `hitTest` reads its point in the wrong coordinate space, so a click on a handle usually misses, and on an image the hit radius shrinks with zoom.

**Why now.** Batch 7 dogfood note (2026-10-15). S14-28 made `model.page` honest, so a region drawn on a later PDF page is no longer stamped page 1 and hidden.

**What stays.** The look of handles, the move diamond and the stroke at 1×, and on every PDF. Right-click deletes a freeform vertex. The armed-tool drawing paths. The locator model (`CitationLocatorDraft.setRegion` restamps the viewer page). Depiction regions stay rectangle-only (S14-03). Go accepts the same regions.

## 2. User stories

**US-1.** As a researcher with a committed freeform region on a PDF page, when I drag a vertex, that vertex follows the pointer and the Citation's region updates on release.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `ArtifactPDFHostView` 400×500 in a window; overlay `committed` = freeform triangle; mouse-down point at a vertex's position, given in host (superview) coordinates | `overlayView.hitTest(point)` returns the overlay (today it returns nil for a vertex in the top or bottom part of the page, because y is mirrored) |
| AC-1.2 | Same, point 30 pt away from every vertex and the stroke | `hitTest` returns nil, so PDFKit keeps selection and scroll |
| AC-1.3 | Vertex 2 of a triangle dragged to (0.6, 0.7) and released, giving a valid polygon | `onCommit` fires once with points where index 2 == (0.6, 0.7) and the other two are unchanged |

**US-2.** As a researcher, when I drag the outline of a committed region (away from any vertex) or the centre diamond, the whole shape moves.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Freeform square (0.2…0.4); document point on the top edge midpoint, more than the hit radius from either corner | `ArtifactRegionGeometry.hitHandle` returns `.move` |
| AC-2.2 | Same point at the centroid | `.move` (diamond, as today) |
| AC-2.3 | Drag from the edge by (+0.1, +0.05) | every point moves by (+0.1, +0.05), clamped into the unit square by `translate` |

**US-3.** As a researcher zoomed out on an image Artifact, handles stay large enough to grab.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Overlay inside an `NSScrollView` document view at `magnification` 0.5 | `handleHitRadius` == 14 (7 screen pt ÷ 0.5) and the drawn handle is 14 document pt wide (7 screen pt) |
| AC-3.2 | Same at magnification 1 and in `ArtifactPDFHostView` | `handleHitRadius` == 7 (unchanged today) |
| AC-3.3 | A 256-point ring whose vertices sit closer than the hit radius; point exactly on vertex 10 | `hitHandle` returns `.resize(10)` (nearest wins, not the first in range) |

**US-4.** As a researcher reopening a saved Citation, I can nudge its polygon and save the change.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Composer opened on `cit-1` with a 4-point region locator; `commitViewerRegion` with the same region with one vertex moved | returns true; `fields.locator.region` equals the moved region; `fields.isDirty`; `canSaveCitation` |
| AC-4.2 | A drag that leaves a self-intersecting (bowtie) shape is released | `onCommit` is not called; the overlay's `committed` goes back to the shape from before the drag; the system beep plays |

## 3. Scope

| In | Covered by |
| --- | --- |
| `hitTest` converts its superview point into overlay coordinates | STEP-1, AC-1.1, AC-1.2, D-1 |
| One pure hit function in `ArtifactRegionGeometry` (anchor, nearest handle, then stroke → `.move`) used by `hitTest`, `mouseDown`, cursor and right-click | STEP-2, AC-2.1, AC-2.2, AC-3.3, D-2, D-3 |
| Outline drag translates the shape | STEP-2, AC-2.3, D-3 |
| Hit radius, cursor pad, handle and diamond draw size in screen points | STEP-3, AC-3.1, AC-3.2, D-4 |
| Invalid drag reverts and beeps | STEP-4, AC-4.2, D-5 |
| Saved-Citation edit pinned by a model test | STEP-5, AC-4.1, D-6 |
| Tests | STEP-5 |
| Patch bump | STEP-6 |
| Docs | STEP-7 |

| Out | Home |
| --- | --- |
| Editing rectangles, L-shapes and circles beyond what the shared hit fix gives them | not planned (map scope entry) |
| Adding a vertex by clicking an edge | dogfood log |
| Depiction regions (rectangle-only) | not planned (S14-03 rule) |
| Keyboard nudging of vertices; VoiceOver per-vertex elements | dogfood log |
| Undo of a region edit | not planned |
| New handle look, hover state or colours | not planned (no brief) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Hit claim | `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` `ArtifactRegionOverlayView.hitTest(_:)` | Passes `point` straight to `hitHandle(at:)` and `isNearCommittedStroke(_:)`. AppKit gives `hitTest` a point in the **superview's** coordinates. In `ArtifactPDFHostView` (not flipped) the overlay is flipped and fills the host, so y is mirrored and handle hits miss. In `ArtifactMediaDocumentView` (flipped, overlay frame = bounds) the two spaces match, so images work |
| Handle hit | same file `hitHandle(at:)` | Move anchor first (radius `handleHitSize + 1`), then the first resize handle within `handleHitSize`. First-in-range, not nearest |
| Vertex hit for delete | same file `freeformVertexIndex(at:)` | Same loop as `hitHandle` over `committed.points` |
| Stroke | same file `isNearCommittedStroke(_:)` | Claims the event in `hitTest`, but `mouseDown` only acts on `hitHandle`, so an outline click is swallowed and does nothing |
| Hit size | same file `handleHitSize` | Constant 7, in overlay coordinates. Under `NSScrollView` magnification (images) that is 7 × magnification screen pt; minimum zoom is `ArtifactViewerModel.minZoom` 0.5 |
| Draw size | same file `drawHandle(at:)` 7, `drawMoveAnchor(at:)` 9, cursor `pad` 8 in `resetCursorRects` | Same space as the hit size |
| Drag | same file `mouseDown` / `mouseDragged` / `mouseUp` | `.resize(i)` → `moveHandle` → `ArtifactRegionGeometry.moveFreeformVertex`; `.move` → `translate`. `mouseUp` calls `onCommit` only if the draft `isValid`; an invalid draft stays drawn in `committed` though the model still holds the old one |
| Move anchor | `ArtifactRegionGeometry.swift` `moveAnchor(for:)` | `circleCenter(points)`, the centroid, for **every** kind. Freeform already has the diamond (the UX note says "circle-only"; that is not true at this base) |
| Geometry | same file `moveFreeformVertex`, `translate`, `deleteFreeformVertex`, `handlePoints(for:imageRect:)` | Pure; `handlePoints` for `.freeform` is the points themselves |
| PDF host | `ArtifactPDFViewport.swift` `ArtifactPDFHostView` | `isFlipped == false`; `overlayView` pinned to all four edges; `layoutPageOverlay` sets `mediaRectOverride` in overlay (flipped) space |
| Image host | `ArtifactMediaViewport.swift` `ArtifactMediaDocumentView` | `isFlipped == true`; overlay `frame = bounds`; lives under `ArtifactMediaNSScrollView.magnification` |
| Saved Citation | `CitationComposerModel.swift` `loadCitation` → `fields.replace(from:)`; `CitationComposerModel+Depiction.swift` `viewerRegion`, `commitViewerRegion` | A loaded Citation's locator goes into `fields.locator`, `viewerRegion` returns `locator.region`, and `commitViewerRegion` → `setRegion` writes it back. So the saved locator **is** already fed as `committedRegion`; no test pins edit-after-load |
| Tests | `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift` `translateClampsToUnitSquare`, `freeformKeepsAtLeastThreeVertices`; `ArtifactPDFViewportPageTests.swift` (host-in-window pattern); `CitationComposerModelTests.swift` `editPrepareKeepsRegionLocatorAndIsNotDirty` | No hit-test, screen-size or edit-after-load test |

**Precedent.** `af433b89` ("Fix region editing and match the locator bar to the board") last changed the edit path in this overlay; copy its shape (geometry helper in `ArtifactRegionGeometry`, overlay calls it). Host tests copy `ArtifactPDFViewportPageTests`' `bootstrapHost` / `hostWindow` from S14-28 (#711).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `hitTest(_ point:)` first does `let local = convert(point, from: superview)` and uses `local` for every check | AppKit's contract; fixes the PDF host and is a no-op for the image host | Flipping the PDF host (changes PDFKit layout and `layoutPageOverlay`); special-casing the PDF path |
| D-2 | Add `ArtifactRegionGeometry.hitHandle(at documentPoint: CGPoint, draft: ArtifactRegionDraft, imageRect: CGRect, radius: CGFloat) -> ArtifactRegionHandle?` *(new)*. Order: move anchor within `radius + 1` → `.move`; else the **nearest** handle within `radius` → `.resize(i)`; else the nearest stroke segment within `radius` → `.move`; else nil. The overlay's `hitHandle(at:)` calls it; `isNearCommittedStroke` and the private `distance(_:segmentFrom:to:)` move into it; `freeformVertexIndex(at:)` returns the `.resize(i)` index when `committed.kind == .freeform` | One testable rule for hitTest, mouseDown, cursor and right-click; nearest-wins is needed for dense freehand rings | Keeping three loops in the view; first-in-range |
| D-3 | An outline hit is `.move` (user, 2026-11-05). It reuses the existing `.move` drag (`moveDragOrigin` / `translate`) and `applyCursor`'s open hand; `resetCursorRects` is unchanged (cursor on the outline comes from `mouseMoved`) | The outline is a larger target than the diamond; nothing new is drawn | A new `.translate` case; diamond only |
| D-4 | Sizes are screen points. Add `private var screenScale: CGFloat` = `convert(CGSize(width: 1, height: 1), to: nil).width`, falling back to 1 when ≤ 0 or not finite; `handleHitRadius` *(rename of `handleHitSize`, internal for tests)* = `7 / screenScale`; `drawHandle` size `7 / screenScale`, diamond `9 / screenScale`, line widths `1.5 / screenScale` and `1.25 / screenScale`, cursor `pad` `8 / screenScale`. The stroke width (2) is unchanged | On a PDF `screenScale` is 1, so nothing changes there; on images it matches the PDF look at every zoom (user, 2026-11-05) | Hit radius only (handles stay invisible when zoomed out); the 44 pt `PVSpacing.hitMin` (swallows PDF clicks over a small region) |
| D-5 | `mouseDown` on a handle stores `editStartDraft = committed` *(new private)*. `mouseUp` while editing: valid → `onCommit` (as today); invalid → `committed = editStartDraft`, `NSSound.beep()`, no `onCommit`. `editStartDraft` clears with the other drag state. Same for the right-click delete: an invalid result is not applied | The view and the model must agree; beep is the overlay's existing refusal sound (cap reached) | Committing invalid shapes; leaving the bowtie drawn; a new L10n announcement (no copy in this PR) |
| D-6 | No composer model change. Edit-after-save is pinned by a test only (T-6) | §4: `viewerRegion` already returns the loaded locator; the map's "feed a saved Citation's locator back" is true at this base | Adding a parallel `committedRegion` path in the model |

**Rules this plan bends.** None. The working rule says "an open Citation whose locator is fed back as `committedRegion`"; D-6 records that it already is, and T-6 holds it.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ArtifactRegionGeometry.hitHandle(at:draft:imageRect:radius:)` *(new)* | func | `ArtifactRegionOverlayView.hitHandle(at:)`, `freeformVertexIndex(at:)`, `isNearCommittedStroke(_:)` (private) | Moves those three into one pure function beside `handlePoints` and `moveAnchor`; the view keeps thin wrappers |
| `screenScale`, `editStartDraft` *(new private)* | property | `handleHitSize` | Extends the same view |
| `ArtifactRegionOverlayTests.swift` *(new)* | file | `ArtifactPDFViewportPageTests.swift` | That file is about page feedback; overlay hits and scale are a separate suite. It copies the window bootstrap shape |

**Builds on.** `ArtifactRegionGeometry` helpers; `ArtifactRegionHandle`; `af433b89`'s edit path; S14-28's host test pattern; `RegionFixtures.ring(count:)`.
**Deletes.** `ArtifactRegionOverlayView.isNearCommittedStroke(_:)` and the private `distance(_:segmentFrom:to:)` (moved into geometry); the loop bodies of `hitHandle(at:)` and `freeformVertexIndex(at:)`; the name `handleHitSize` (renamed `handleHitRadius`).
**Next adds.** Insert-vertex-on-edge would be one more case in `hitHandle` (an edge index) if the dogfood log asks for it.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Clicks on a PDF region now reach the overlay | `mouseDown` edit path | Only within `radius` of a handle or the stroke; elsewhere PDFKit still gets the click (AC-1.2) |
| Edit of a loaded Citation's region | `commitViewerRegion` → `setRegion` → `fields` dirty | Existing Save / discard prompts apply (`hasUnsavedDocumentWork`); restamps the viewer page, which equals the locator page because the overlay only shows on it (`committedOnCurrentPage`) |

### Scale

None — no catalog write. `hitHandle` is O(points) per mouse event, with points ≤ `maxRegionPoints` (256).

### Edges

- **Assumptions.** `handlePoints(.freeform) == points`, so `.resize(i)` is a vertex index (held by T-3). `screenScale` is uniform in x and y (`NSScrollView` magnification is).
- **Neighbors.** Anchor vs. a vertex near the centroid of a tiny shape: anchor wins, as today (T-2). Rectangle edges now also translate the rectangle (shared rule): accepted, since the stroke claim already swallowed those clicks. Armed tool: `hitTest` still returns self first, unchanged.
- **Removed inputs.** A drag to an invalid shape restores the start shape (T-5).
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
| Region overlay | `ArtifactRegionOverlayView` (feature snowflake) | `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` | D-1, D-3 – D-5 |
| Hit rule | `ArtifactRegionGeometry.hitHandle` | `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift` | D-2 |

States: no region → `hitTest` nil (unchanged). Accessibility: unchanged; the overlay has no per-vertex elements (Out).

### Strings

None — no copy changes. The tooltip `L10n.CitationComposer.freeformDeleteVertex` still describes right-click delete.

## 7. Steps

**STEP-1 — Convert the hit-test point**
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift`.
Do: D-1 in `hitTest(_:)`.
Check: builds; `rg 'convert\(point, from: superview\)' macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` hits once.

**STEP-2 — One hit rule in geometry, outline moves the shape**
Files: `ArtifactRegionGeometry.swift`, `ArtifactRegionOverlay.swift`.
Do: add `hitHandle(at:draft:imageRect:radius:)` per D-2 (move the segment-distance helper into `ArtifactRegionGeometry` as `private static`). Make the overlay's `hitHandle(at:)` call it; `hitTest` returns self when it is non-nil (drop the separate stroke check); `freeformVertexIndex(at:)` maps `.resize(i)` for freeform. Delete `isNearCommittedStroke` and the view's `distance`.
Check: builds; `rg isNearCommittedStroke macos/` has no hits.

**STEP-3 — Screen-point sizes**
Files: `ArtifactRegionOverlay.swift`.
Do: D-4. Rename `handleHitSize` → `handleHitRadius` (internal). Scale the draw sizes and line widths of `drawHandle`, `drawMoveAnchor` and the cursor pad. Invalidate cursor rects and redraw when magnification changes: call `window?.invalidateCursorRects(for: overlayView)` and `overlayView.needsDisplay = true` from `ArtifactMediaDocumentView.layoutOverlay()` (already called on zoom).
Check: `rg handleHitSize macos/` has no hits.

**STEP-4 — Revert an invalid edit**
Files: `ArtifactRegionOverlay.swift`.
Do: D-5 in `mouseDown`, `mouseUp`, `rightMouseDown`.
Check: builds.

**STEP-5 — Tests**
Files: `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift`, `macos/ProvenenciaTests/ArtifactRegionOverlayTests.swift` *(new)*, `macos/ProvenenciaTests/CitationComposerModelTests.swift`. Skill: `add-swift-test`.
Do: T-1 – T-6 per §8.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/ArtifactRegionGeometryTests -only-testing:ProvenenciaTests/ArtifactRegionOverlayTests -only-testing:ProvenenciaTests/CitationComposerModelTests` passes.

**STEP-6 — Version bump**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION` Debug and Release).
Do: patch bump together (`0.3.2` → `0.3.3` at the planning base, or the next patch of what `VERSION` reads).
Check: the three places match.

**STEP-7 — Docs**
Files: this plan (status, Landed note), `docs/deployment-plan/spike-14/deployment-plan.md` (map row), `docs/deployment-plan/spike-14/completed.md`.
Do: per the Definition of done.
Check: links resolve.

Commits: one per behavior. Suggested: "Let clicks on a PDF region reach its handles."; "Move a region by dragging its outline and pick the nearest vertex."; "Keep region handles the same size on screen at any zoom."; "Put a region back when a drag would make it invalid."; "Cover region hits, handle size and editing a saved Citation's region."; "Bump version to 0.3.3."

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `ArtifactRegionGeometryTests.swift` `hitHandleOutlineMovesShape` *(new)* | square (0.2,0.2)…(0.4,0.4) in a 100×100 `imageRect`, radius 3: (30, 20) → `.move`; (30, 30) → `.move`; (20, 20) → `.resize(0)`; (60, 60) → nil | AC-2.1, AC-2.2, D-2, D-3 |
| T-2 | Swift | same `hitHandleAnchorWinsOverNearVertex` *(new)* | tiny triangle whose centroid is within radius of vertex 0; point at the centroid → `.move` | D-2 |
| T-3 | Swift | same `hitHandlePicksNearestVertex` *(new)* | `RegionFixtures.ring(count: 256)` in a 500×500 rect, radius 7; point on vertex 10 → `.resize(10)` | AC-3.3 |
| T-4 | Swift | `ArtifactRegionOverlayTests.swift` `pdfHostHitTestFindsVertex` *(new, @MainActor)* | `ArtifactPDFHostView` 400×500 in a borderless window (copy S14-28's `hostWindow`), `mediaRectOverride` (50,50,300,400), committed triangle (0.1,0.1),(0.9,0.1),(0.5,0.9); vertex 2 converted to host coordinates with `host.convert(_:from: overlayView)` → `hitTest` returns `overlayView`; a point 30 pt from every vertex and edge → nil | AC-1.1, AC-1.2, D-1 |
| T-5 | Swift | same `invalidDragRevertsAndDoesNotCommit` *(new)* and `vertexDragCommitsOnce` *(new)* | drive `mouseDown`/`mouseDragged`/`mouseUp` with `NSEvent.mouseEvent` at window points on the overlay: drag a square's vertex across the opposite edge (bowtie) → `committed` equals the start square and `onCommit` count 0; drag a triangle vertex 2 to (0.6,0.7) → `onCommit` once with that point | AC-1.3, AC-4.2, D-5 |
| T-6 | Swift | same `handleRadiusIsInScreenPoints` *(new)* | overlay as document subview of an `NSScrollView` with `allowsMagnification`, in a window; magnification 0.5 → `handleHitRadius == 14`; 1 → 7; in an `ArtifactPDFHostView` → 7 | AC-3.1, AC-3.2, D-4 |
| T-7 | Swift | `CitationComposerModelTests.swift` `editSavedCitationRegionMarksDirty` *(new)* | `seedArtifact`; `seedCitation(locatorJSON: regionLocatorJSON)`; `prepare()`; `commitViewerRegion(.init(kind: .freeform, points: [(0.1,0.1),(0.45,0.1),(0.4,0.4),(0.1,0.4)]))` → true; `fields.locator.region?.points[1] == (0.45,0.1)`; `fields.isDirty`; `canSaveCitation` | AC-4.1, D-6 |

Must stay green: everything `scripts/check-pr.sh` runs; `ArtifactRegionGeometryTests`, `ArtifactPDFViewportPageTests`, `CitationLocatorDraftTests`, `CitationComposerDepictionTests` (rectangle-only depiction drawing shares the overlay).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `handleHitSize` and `isNearCommittedStroke` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive (no brief: handles look as on a PDF today)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: none owed
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open the Citation composer on a multi-page PDF Artifact. Scroll to page 3, arm the polygon tool, draw a five-point shape, close it.
2. Drag a vertex near the top of the page and one near the bottom (AC-1.1). Each follows the pointer; the locator chip still reads "Polygon · 5 points".
3. Drag the outline between two vertices (AC-2.3): the whole shape moves. Drag the centre diamond: it moves too.
4. Click well away from the shape: PDF text selection still works (AC-1.2).
5. Drag a vertex across the opposite edge so the shape crosses itself; release (AC-4.2): beep, and the shape returns.
6. Save. Leave and reopen that Citation; nudge a vertex; Save is enabled; save, reopen, the change is kept (AC-4.1).
7. Open an image Artifact, zoom to the minimum (50%), and grab a vertex (AC-3.1): handles look the same size as at 100% and are easy to hit.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| PDFKit's `PDFView` takes the mouse-down before the overlay even after D-1 | The overlay is the later sibling, so the host's `hitTest` asks it first; T-4 and manual check 2 confirm |
| Synthesized `NSEvent`s in T-5 do not reach `mouseDragged` | Call the overlay's `mouseDown(with:)` / `mouseDragged(with:)` / `mouseUp(with:)` directly with events built at the window location |
| `convert(_:to: nil)` returns 0 for a view with no window | D-4's fallback to 1; T-6 runs in a window |

**Stop and report if:**

- T-4 passes **before** STEP-1 (then the y-mirror is not the PDF cause; report what does block handle clicks).
- `viewerRegion` does not return a loaded Citation's region (T-7 fails before any model change): D-6 is wrong.
- `commitViewerRegion` returns false in T-7 because the fake PDF Artifact refuses a region (`locatorCapabilities`); report the capability values.
- Keeping PDF clicks working needs the stroke claim dropped (AC-1.2 vs D-3 conflict).

## 12. Open questions

None.

## 13. PR

**Title:** S14-29 — Polygon locator vertex edit
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-29-polygon-locator-vertex-edit.md`.
