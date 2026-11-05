# Review — S14-29 Polygon locator vertex edit

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
None.

## Writer's own calls
- The PDF root cause is a reading of AppKit's `hitTest` contract (superview coordinates) against a non-flipped host and a flipped overlay; it is not reproduced in the app from this session. STEP-1 makes T-4 fail first, and a stop condition covers the case where it does not.
- The UX note's "moveAnchor is circle-only" is false at the base; the plan keeps the existing centroid diamond and adds the outline as a second move target (user picked the outline, 2026-11-05).
- "Feed a saved Citation's locator back as `committedRegion`" is already true; the plan pins it with a test (D-6) and corrects the map scope entry instead of adding code.
- ArtifactViewer calls `GraphCanvasCoordinates.screenPixelDocumentWidth` directly (M-1) rather than adding a copy in `ArtifactRegionGeometry`; it is a pure function in the same app target.
- Rectangle, L and circle outlines also move their shape, because the hit rule is shared (D-3).
- The UX note is deleted on landing, as S14-30 did in `7e69154a`.
- A click with no drag no longer commits (round 2 X-9): today a handle click re-commits the unchanged draft; with the outline now a large target, skipping unchanged commits keeps a stray click from dirtying a saved Citation.
- The map's S14-29 row and Ships line were corrected at planning time, not on landing.

## Round 1

### M-1 · blocker · Screen-point sizing adds a second, implicit "document width of one screen point" mechanism
**Plan:** D-4 `screenScale` via `convert(_:to: nil)`; `layoutOverlay` reaching into the overlay.
**Codebase:** `GraphCanvasCoordinates.screenPixelDocumentWidth(magnification:)`; overlay inputs with `didSet`; `ArtifactMediaNSScrollView.magnification` `didSet`.
**Answer:** Fixed — D-4 is now a declared `magnification` input set from the scroll view's `didSet` and `configureOverlay`, sizes from `screenPixelDocumentWidth`; T-7 is a pure test. Also answers X-4, X-7, X-12.

### M-2 · blocker · `resetCursorRects` keeps its own handle loop beside the "one hit rule"
**Answer:** Fixed — D-2 adds `handleTargets`; D-7 makes `resetCursorRects` iterate it with `handleHitRadius` as pad; `pad` 8 deleted.

### M-3 · major · `editStartDraft` sits beside `moveDragPoints` in a drag-state pile
**Answer:** Fixed — D-5 is one `edit` tuple and one `finishEdit` used by `mouseUp` and `rightMouseDown`; the four fields are in Deletes and STEP-4's check. Also answers X-3.

### M-4 · major · The test window bootstrap would be a third hand-rolled copy
**Answer:** Fixed — D-8: `hostInBorderlessWindow(nsView:size:)` and `ArtifactFileFixtures.pdfHost`; `ArtifactPDFViewportPageTests` switches to them (STEP-1).

### M-5 · minor · `ArtifactRegionHandle` belongs beside the geometry
**Answer:** Fixed — D-2 moves it; STEP-2 check.

### M-6 · minor · Right-click delete inherits anchor and stroke precedence
**Answer:** Fixed — D-2 splits out `nearestHandle`; D-7 has `freeformVertexIndex` call it; T-3b.

### M-7 · minor · The map scope entry stays false after this PR
**Answer:** Fixed — STEP-7 rewrites the row scope and Ships line.

### M-8 · nit · Test-ID drift and a helper name
**Answer:** Fixed — tests renumbered; helper named `distanceToSegment`.

### X-1 · blocker · Stop condition "T-4 passes before STEP-1" cannot be run in order
**Answer:** Fixed — STEP-1 writes T-4 and runs it red before D-1.

### X-2 · blocker · T-4 negative point undefined; obvious point hits the anchor
**Answer:** Fixed — AC-1.2 and T-4 name normalized (0.5, 0.02), above the top edge and far from the centroid, converted with `host.convert(_:from: overlayView)` like the positive points.

### X-3 · major · Right-click invalid result underspecified
**Answer:** Fixed — D-5 `finishEdit`: keep the old shape, beep, no `onCommit`; AC-4.3, T-6.

### X-4 · major · STEP-3 redraw hook unverified
**Answer:** Fixed — by M-1: the `magnification` input's `didSet` redraws and invalidates cursor rects.

### X-5 · major · No test covers the drawn handle size
**Answer:** Fixed — internal `handleDrawSize`; T-7 asserts it.

### X-6 · major · Circle, L and rectangle outline behaviour not stated
**Answer:** Fixed — D-3 states the circle ring and handle-beats-outline; AC-2.4, T-1b; Out and Neighbors updated.

### X-7 · major · `screenScale` fallback
**Answer:** Fixed — moot after M-1.

### X-8 · minor · `editStartDraft` lifecycle
**Answer:** Fixed — §1 What stays: a click with no drag still commits the unchanged draft.

### X-9 · minor · Sweep scope and the UX note
**Answer:** Fixed — DoD sweep covers `docs/` including spike folders; STEP-7 deletes the note as `7e69154a` did.

### X-10 · minor · `MARKETING_VERSION` count
**Answer:** Fixed — STEP-6 says all four lines.

### X-11 · minor · T-5 event synthesis underspecified
**Answer:** Fixed — T-5 and T-6 use T-4's host, call the overlay's handlers directly with named event types and points, and name the bowtie; the beep is manual only.

### X-12 · minor · T-6 host
**Answer:** Fixed — moot after M-1 (T-7 is a bare overlay).

### X-13 · minor · Helper reuse for T-4; `layout()` overwrites `mediaRectOverride`
**Answer:** Fixed — T-4 mounts a one-page PDF through `pdfHost`, lays out, and reads `mediaRectOverride` instead of setting it; a stop condition covers an empty media box.

## Round 2

### M-9 · blocker · D-4 has two writers for the overlay's zoom, neither on every zoom path
**Plan:** D-4 set `overlayView.magnification` from the scroll view's `didSet` and from `configureOverlay`.
**Codebase:** `ArtifactMediaNSScrollView.layoutDocumentCentered` is where the setter, wheel zoom, pinch (`clipBoundsDidChange`) and `updateNSView` all end; `configureOverlay` runs before the magnification is applied.
**Answer:** Fixed — D-4 has one writer in `layoutDocumentCentered` after `guard let document`; STEP-3 checks exactly one hit; manual check 7 zooms by chrome, pinch and wheel. Also answers round-2 X-3.

### X-1 · major · AC-4.3 has no automated test; beep contradiction
**Answer:** Fixed — AC-4.3 names a five-point shape whose delete crosses; T-6 covers it; the nil path stays silent (§1, D-5); manual check 5 adds the right-click beep.

### X-2 · major · `handleTargets` returns document points; cursor kind needs normalized
**Answer:** Fixed — targets carry `normalized` and `document` (D-2); D-7 names `target.normalized`.

### X-3 · major · D-4 access path
**Answer:** Fixed — by M-9.

### X-4 · minor · T-8 dirty from the page stamp alone; tuple vs CGPoint
**Answer:** Fixed — T-8 seeds a page-1 selector, asserts the page is unchanged and not dirty after load, and compares `CGPoint`s.

### X-5 · minor · T-5 / T-6 event construction
**Answer:** Fixed — §8 names one `event(_:at:in:window:)` helper with every argument and the down / dragged / up sequence.

### X-6 · minor · DoD sweep hits the review log
**Answer:** Fixed — "outside this plan and its review log".

### X-7 · minor · STEP-2 / STEP-4 checks are greps only
**Answer:** Fixed — tests move into the steps they cover (T-1 – T-3b in STEP-2, T-7 in STEP-3, T-5 – T-6 in STEP-4).

### X-8 · minor · T-4 depends on an unverified PDF fixture
**Answer:** Fixed — D-8 adds `sizedPDFDocument(size:)` (300×400 media box); the stop condition checks the media rect.

### X-9 · minor · An outline click now dirties a Citation
**Answer:** Fixed — D-5 `finishEdit` skips an unchanged draft; AC-4.4, T-5, manual check 8.

### X-10 · nit · Smaller choices
**Answer:** Fixed — T-6 coordinates; `finishEdit(committed, start: edit.start)`; `drawHandle` uses `handleDrawSize` for committed and in-progress handles.
