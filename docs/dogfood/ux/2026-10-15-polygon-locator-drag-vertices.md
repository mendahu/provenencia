# Drag polygon locator vertices instead of redrawing

- **Date:** 2026-10-15
- **Type:** Feature request
- **Where:** Citation composer / polygon locator
- **Annoyance:** After drawing a freeform polygon locator, a slightly misplaced corner means clearing the region and redrawing the whole shape. There is no apparent way to grab a vertex and nudge it.
- **Wanted:** After a polygon is drawn, drag its individual vertices to adjust the shape (and ideally move the whole polygon), with the citation's region updating in place.

## Code notes (observed, not implemented)

- `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` — `ArtifactRegionOverlayView` already has most of the plumbing for this: `handlePoints(for:)` returns every vertex for `.freeform`, `drawCommitted` draws a handle on each, `hitHandle(at:)` / `mouseDown` / `mouseDragged` route a `.resize(index)` drag through `moveHandle` → `ArtifactRegionGeometry.moveFreeformVertex`, and `rightMouseDown` deletes a vertex via `deleteFreeformVertex`. `mouseUp` then calls `onCommit` with the edited draft. So the request is partly about why this path is not reachable or discoverable in practice rather than a feature from scratch.
- Same file — candidates for why it does not work as a user expects: `handleHitSize` is a fixed 7 pt in document space (tiny on a zoomed-out PDF page); `hitTest` only claims the event when `armedTool == nil` and a handle or the stroke is hit, so any still-armed tool swallows the click into `handleFreeformClick`; and only the resize handles exist for freeform — `moveAnchor` is circle-only, so there is no way to translate the whole polygon.
- `macos/App/Features/ArtifactViewer/ArtifactViewer.swift` — `regionOverlayInput` only passes `committedRegion` through when `ArtifactRegionOverlayInput.committedOnCurrentPage` agrees the locator page matches `model.page`. Combined with the page indicator bug (`2026-10-15-pdf-page-indicator-stuck-on-page-1.md`), a polygon drawn while scrolled to a later page can be stamped page 1 and then hidden from the overlay, so its handles never show.
- `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` and `CitationComposerModel.setRegion` — `committedRegion` is `model.locator.region`, i.e. the pending citation's draft. A region on an already-saved citation is not fed back as `committed`, so editing after save would need the open citation's locator to round-trip into the overlay.
- `macos/App/Features/ArtifactViewer/ArtifactPDFViewport.swift` — `overlayView` is layered over `PDFView` and hidden unless `overlayEnabled`; worth confirming PDFKit's scroll view does not take the mouse down before the overlay's handle hit-test on the PDF path.
- `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift` — existing geometry tests to extend if `moveFreeformVertex` / a freeform translate gain behaviour.
