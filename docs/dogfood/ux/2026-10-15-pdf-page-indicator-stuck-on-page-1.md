# PDF page indicator stays on page 1 while scrolling

- **Date:** 2026-10-15
- **Type:** Bug
- **Where:** Citation composer / PDF viewer
- **Annoyance:** Viewing a multi-page PDF in the citation composer and scrolling through the document, the page number indicator in the top-left stays on page 1 instead of updating to the page currently in view.
- **Consequence:** Using the polygon selector tool to create a citation region while looking at a later page attaches the citation to page 1 rather than the page actually being viewed.
- **Wanted:** The page indicator tracks the visible page while scrolling, and polygon selections are attached to the page they were drawn on.

## Code notes (observed, not fixed)

- `macos/App/Features/ArtifactViewer/ArtifactPDFViewport.swift` — the coordinator observes `.PDFViewPageChanged` / `.PDFViewVisiblePagesChanged`, but `pdfLayoutChanged(_:)` only re-lays out the region overlay and reports zoom via `onZoomChange`. There is no equivalent page callback, so `pdfView.currentPage` never flows back to the model when the user scrolls.
- `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` — `page` only changes through `setPage(_:)` from the chrome (prev/next/page field), so the indicator and `model.page` stay at 1 after a scroll.
- `macos/App/Features/ArtifactViewer/ArtifactViewer.swift` and `macos/App/Features/CitationComposer/CitationLocatorDraft.swift` — the region locator is restamped with `viewerPage: model.page`, which is why a polygon drawn on a later page lands on page 1.
