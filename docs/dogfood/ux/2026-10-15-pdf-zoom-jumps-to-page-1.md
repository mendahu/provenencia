# Zooming the PDF viewer jumps back to page 1

- **Date:** 2026-10-15
- **Type:** Bug
- **Where:** Citation composer / PDF viewer zoom
- **Annoyance:** After scrolling a multi-page PDF to a later page, changing the zoom level throws the view back to page 1.
- **Wanted:** Zooming keeps the current page in view, and ideally the scroll position within it.

## Related

Almost certainly the same root cause as [PDF page indicator stays on page 1 while scrolling](2026-10-15-pdf-page-indicator-stuck-on-page-1.md): `ArtifactViewerModel.page` never learns about scroll-driven page changes, so anything that re-syncs the viewer from the model snaps back to the stale page. Fixing the indicator bug (push `pdfView.currentPage` back into the model on `.PDFViewPageChanged`) should resolve this one as well; it is worth re-testing zoom when that lands.

## Code notes (observed, not fixed)

- `macos/App/Features/ArtifactViewer/ArtifactPDFViewport.swift` — `ArtifactPDFHostView.sync(document:page:zoom:resetPage:…)` runs on every SwiftUI update (`updateNSView` → `apply` → `sync`). It first resolves `page` (the model's page, still 1 after a scroll) and calls `pdfView.go(to:)` whenever `pdfView.currentPage != pdfPage`, and only then applies `scaleFactor`. A zoom change triggers an update, the page comparison sees "viewer on page 5, model says 1", and the view is sent to page 1 before the zoom is applied. Nothing in `sync` preserves the visible rect across the scale change either.
- Same file — `pdfLayoutChanged(_:)` observes `.PDFViewPageChanged` / `.PDFViewVisiblePagesChanged` but only reports zoom (`onZoomChange`) and re-lays out the overlay; there is no `onPageChange` to keep the model's `page` current, which is the shared root cause.
- `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` — `page` changes only via `setPage(_:)` from the chrome; `setZoom(_:)` / `zoomIn()` / `zoomOut()` update `zoom`, which is what drives the re-sync.
- `macos/App/Features/ArtifactViewer/ArtifactViewer.swift` — `contentIdentity` includes `model.page`; if the model's page starts tracking scroll, check that this does not flip `resetPage` and cause a `go(to:)` of its own on every scroll.
