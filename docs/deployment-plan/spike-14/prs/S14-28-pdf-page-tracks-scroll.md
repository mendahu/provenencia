# S14-28 — PDF page tracks scroll

**Status:** Landed in #711 (2026-11-04)
**Landed:** Product version is 0.3.1 (patch on 0.3.0 at implement time, not 0.2.14 from the planning base). Manual check §10 not run in implementer pass.
**Spike / slice:** Spike 14 · Slice 7a — PDF page sync
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the page field and zoom chrome already exist; this PR only keeps `model.page` honest when PDFKit changes the visible page)
**Board:** —
**Depends on:** —
**Blocks:** S14-29
**Branch:** `s14-28-pdf-page-tracks-scroll`
**Version bump:** patch — researcher-visible bugfix for the page indicator, zoom snap, and region page stamps ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 7) PDF page sync is one PR"; bar outcome 23; [`ux/2026-10-15-pdf-page-indicator-stuck-on-page-1.md`](../ux/2026-10-15-pdf-page-indicator-stuck-on-page-1.md), [`ux/2026-10-15-pdf-zoom-jumps-to-page-1.md`](../ux/2026-10-15-pdf-zoom-jumps-to-page-1.md); S14-29 depends on this so a region drawn while scrolled is not stamped page 1
**Planned:** 2026-11-04 against `main` @ `d56de8d3`
**Review:** [`S14-28-pdf-page-tracks-scroll.review.md`](S14-28-pdf-page-tracks-scroll.review.md) — Fits — Ready with changes, 2 rounds; no open blockers; To user none

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

Scrolling a multi-page PDF in the Artifact viewer (Citation composer) updates the page indicator to the page in view, zooming keeps that page, and a region drawn there is stamped with that page — not page 1. Both dogfood notes share one root cause: `ArtifactViewerModel.page` never learns scroll-driven page changes, so chrome and `sync` keep acting on a stale page.

**Why now.** Batch 7 pulls these UX notes while Spike 13 continues; S14-29's polygon edit depends on a region drawn while scrolled not being stamped page 1 and then hidden.

**What stays.** Single-page `PDFView` display mode; chrome prev/next/page field and Set page; Find's page jumps through `setPage`; image Artifacts and `ArtifactMediaViewport`; exact scroll-rect preservation within a page across zoom (Out).

## 2. User stories

**US-1.** As a researcher viewing a multi-page PDF, when I scroll or swipe to another page, the page field shows that page so I know where I am and Set page / region stamps use it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | 3-page PDF loaded; `PDFView` navigates to page 3 (scroll / swipe / `go(to:)` that fires `.PDFViewPageChanged`) | `ArtifactViewerModel.page == 3`; the page field text is `"3"` |
| AC-1.2 | Same, then navigate back to page 1 | `model.page == 1` |
| AC-1.3 | Model page is already 2; PDFKit reports page 2 again | `setPage(2)` is a no-op (page stays 2; no churn beyond the existing `guard clamped != page`) |

**US-2.** As a researcher zoomed in on a later page, when I change zoom, the viewer stays on that page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | 3-page PDF; viewer and model on page 3; `setZoom(2.0)` then a SwiftUI-driven `sync` with `page: model.page`, `resetPage: false` | `pdfView.currentPage` is still document page index 2 (1-based page 3); scale is ~2.0 |
| AC-2.2 | After AC-1.1 (model page 3 from a PDFKit report), chrome zoom-in | Page field still shows 3; view does not jump to page 1 |

**US-3.** As a researcher drawing a Citation region on a later page, the locator page is the page I drew on.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Model page is 5 after a PDFKit report; `CitationLocatorDraft.setRegion(…, viewerPage: model.page)` on a PDF capability set | Draft `page == 5` |
| AC-3.2 | Committed region with locator page 5; viewer `model.page == 5` | `ArtifactRegionOverlayInput.committedOnCurrentPage` returns the region (not nil) |

## 3. Scope

| In | Covered by |
| --- | --- |
| One `pageNumber(for page:in:)` helper shared by Find, selection report, and the new page callback | STEP-1, D-8 |
| `ArtifactPDFViewport` / `ArtifactPDFHostView` report the visible 1-based page on `.PDFViewPageChanged` and `.PDFViewVisiblePagesChanged` into the model (beside zoom) | STEP-1, AC-1.1, AC-1.2, D-1, D-2 |
| `suppressPageCallback` around intentional `go(to:)` in `sync` only, mirroring zoom | STEP-1, D-3 |
| `ArtifactViewer` wires `onPageChange: { model.setPage($0) }` (same step as the required parameter) | STEP-1, AC-1.1, D-5 |
| `contentIdentity` drops `model.page` and `pdfPageMediaSize`; delete `pdfPageMediaSize` | STEP-2, D-4, AC-2.1 |
| Swift tests: scroll/report → model page; zoom-while-on-page-N does not snap; duplicate `setPage` no-op | STEP-3, T-1 – T-4, AC-1.*, AC-2.* |
| Patch version bump | STEP-4 |
| Docs: plan Landed, map row, `completed.md` | STEP-5 |

| Out | Home |
| --- | --- |
| Polygon vertex edit / feeding a saved Citation as `committedRegion` | S14-29 |
| Preserving the exact scroll rect within a page across zoom | not planned (map scope: nice-to-have if cheap; this PR does not add it) |
| Changing `PDFView.displayMode` away from `.singlePage` | not planned |
| Image / `ArtifactMediaViewport` page concepts | not planned (images have no pages) |
| New L10n / chrome redesign of the page field | not planned |
| Refactoring Find's `go(to: selection)` suppress path | not planned — Find already calls `setPage` in `activateHit`; a PageChanged callback that re-sets the same page is a no-op |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| PDF viewport | `macos/App/Features/ArtifactViewer/ArtifactPDFViewport.swift` `ArtifactPDFViewport` | `onZoomChange` only. `Coordinator` / `apply` wire zoom into `ArtifactPDFHostView`. No `onPageChange` |
| PDF host observe | same file `ArtifactPDFHostView.startObserving` / `pdfLayoutChanged(_:)` | Observes `.PDFViewPageChanged`, `.PDFViewScaleChanged`, `.PDFViewVisiblePagesChanged`, and scroll-view `boundsDidChange`. `pdfLayoutChanged` always re-lays out the overlay; reports zoom only when `notification.name == .PDFViewScaleChanged` and `!suppressZoomCallback`. Never reports page |
| PDF host sync | same file `ArtifactPDFHostView.sync(document:page:zoom:resetPage:…)` | On every SwiftUI update: if `pdfView.currentPage !=` model page **or** `resetPage`, calls `pdfView.go(to:)`, then applies `scaleFactor` with `suppressZoomCallback`. A zoom-only update with stale `page: 1` while the view is on page 5 snaps to page 1 |
| Viewer wire | `macos/App/Features/ArtifactViewer/ArtifactViewer.swift` `canvas` PDF branch | Passes `page: model.page`, `zoom: model.zoom`, `onZoomChange: { model.setZoom($0) }`. No page callback |
| Content identity | same file `contentIdentity` | Exact: `"\(model.kind.rawValue)-\(model.page)-\(model.pageCount)-\(model.displayImage?.size.width ?? model.pdfPageMediaSize?.width ?? 0)"`. Includes `model.page`, so a scroll-driven `setPage` flips `contentID` → `updateNSView` passes `resetPage: true` → forced `go(to:)` on every scroll report (dogfood zoom note warns about this) |
| PDF media size | `ArtifactViewerModel.swift` `pdfPageMediaSize` | Computed from `document.page(at: page - 1)` media box. Production readers: only `contentIdentity`'s width fallback. Tests: `loadsPDFAndPages` expects 612×792 |
| Page number helper | same file `pageNumber(for:in:)` (private static) | Takes a `PDFSelection`, returns 1-based page or nil (`NSNotFound`). Find's `activateHit` uses it. `ArtifactPDFHostView.reportUserSelection` inlines the same `document.index(for:)` + 1 formula for a selection's first page |
| Model page | `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` `page`, `setPage(_:)`, `setZoom(_:)` | `page` is private(set); chrome and Find call `setPage`. `setPage` clamps against `model.pageCount`. `setZoom` only changes `zoom`. No viewport-owned page API needed |
| Page chrome | same file `ArtifactViewerToolChrome` | `onChange(of: model.page)` refreshes the page field; prev/next/field call `setPage` |
| Find highlight go | `ArtifactPDFViewport.swift` `applyFindHighlight` | Calls `pdfView.go(to: selection)` without suppress; Find already called `setPage` in `activateHit` |
| Locator tests | `macos/ProvenenciaTests/CitationLocatorDraftTests.swift` `pdfRegionStampsPageWhenMissing`, `pdfRegionRestampsPageAfterClearingRegion`, `overlayHidesCommittedRegionOffLocatorPage` | Already pin `setRegion(…, viewerPage:)` stamping and `committedOnCurrentPage` show/hide |
| Region stamp | `macos/App/Features/CitationComposer/CitationComposerModel.swift` `setRegion` / `viewerPage`; `CitationLocatorDraft.setRegion(…, viewerPage:)` | Stamps `viewerPage: artifactViewer.page`. Wrong when the model is stale |
| Overlay hide | `ArtifactRegionOverlay.swift` `committedOnCurrentPage` | Hides committed ink when `locatorPage != viewerPage`. A page-1 stamp on a later drawn page hides handles (feeds S14-29) |
| Model tests | `macos/ProvenenciaTests/ArtifactViewerModelTests.swift` | Covers load, chrome `setPage` / next / clamp, Find page jumps, zoom clamp. No PDFKit → model page feedback; no host `sync` zoom-while-on-N |
| PDF fixtures | `macos/ProvenenciaTests/ArtifactFileFixtures.swift` `writeBlankPDF(pageCount:)` | Multi-page blank PDFs for host tests |
| Window host | `macos/ProvenenciaTests/HostedWindow.swift` `hostInBorderlessWindow` | Borderless `NSWindow` + `NSHostingView` for layout; usable to size an `ArtifactPDFHostView` |
| Wait helper | `macos/ProvenenciaTests/QueryHandleWaiting.swift` `waitUntil` | Polls a condition up to 2s |

**Precedent.** `2169a09e` ("Remount composer PDFs onto a live PDFKit page…") added `ArtifactPDFViewport` and the zoom callback with `suppressZoomCallback` around intentional `scaleFactor` writes. Copy that callback shape for page (`onPageChange` + suppress around `go(to:)`). Do not copy putting `model.page` into `contentIdentity` — that line predates live PDFKit page feedback and fights this PR (dogfood zoom note).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Add `onPageChange: (Int) -> Void` on `ArtifactPDFViewport`, stored on `Coordinator` and `ArtifactPDFHostView`, wired like `onZoomChange` | One mirror of the existing zoom feedback path; model already owns clamping via `setPage` | A Combine publisher, NotificationCenter to the model, or polling `currentPage` on a timer |
| D-2 | Report page only for `.PDFViewPageChanged` and `.PDFViewVisiblePagesChanged`. Resolve the page with `ArtifactViewerModel.pageNumber(for:in:)` on `pdfView.currentPage` (D-8); if non-nil, call `onPageChange?(page)` with no host-side `clampPage`. Do not report from `.PDFViewScaleChanged` or `boundsDidChange` | Matches the map scope; one numbering rule with Find and selection; `setPage` is the only clamp (against `model.pageCount`) | Inlining `document.index` + host `clampPage(document.pageCount)`; reporting on every `pdfLayoutChanged`; picking a "most visible" page other than `currentPage` |
| D-3 | `suppressPageCallback` (private on `ArtifactPDFHostView`), set true only around `pdfView.go(to:)` inside `sync`, mirroring `suppressZoomCallback` around `scaleFactor`. Do not wrap `applyFindHighlight`'s `go(to: selection)` — Find already `setPage`s in `activateHit`; a duplicate PageChanged is a no-op via `setPage`'s equality guard | Chrome-driven `sync` echoes would otherwise re-enter the callback path; Find must keep reporting so a stray desync can heal | Suppressing every `go(to:)` in the file; relying on `setPage`'s guard with no sync suppress |
| D-4 | Change `contentIdentity` to `"\(model.kind.rawValue)-\(model.pageCount)-\(model.displayImage?.size.width ?? 0)"` — drop `model.page` and the `pdfPageMediaSize` fallback. Delete `pdfPageMediaSize` and the two expects in `loadsPDFAndPages` | Scroll-driven `setPage` must not flip `resetPage`. Differently sized PDF pages must not flip identity either (`layoutPageOverlay` already follows the media box). Image remount still keys off kind + width (same width term as today). Height is not added | Keeping `model.page` / `pdfPageMediaSize` in the key; adding image height; leaving `pdfPageMediaSize` with no production caller |
| D-5 | `ArtifactViewer` PDF branch: `onPageChange: { model.setPage($0) }`. Reuse `setPage`; no new model method | One writer for page from chrome, Find, and the viewport | `reportVisiblePage` / `applyViewportPage` parallel API |
| D-6 | Host tests (T-1 – T-3) live in a `@MainActor` `@Suite` (same as `ArtifactViewerModelTests`). Drive `ArtifactPDFHostView` directly in a sized `NSWindow` content view (~400×500, `orderFrontRegardless`); bootstrap with `sync(document:page:1,zoom:1,resetPage:true,selectionEnabled:true,findSelection:nil,findActivationID:0)` then `startObserving()`; use `ArtifactFileFixtures.writeBlankPDF` + `waitUntil`. Do not mount the SwiftUI representable | Host owns observe/sync; trailing `sync` args are pinned; `@MainActor` matches `waitUntil` | XCUITest; testing only `setPage` without the host; leaving `sync(…)` trailing args to inference; a non-isolated suite |
| D-7 | Patch bump `VERSION` / `core.Version` / both Xcode `MARKETING_VERSION` together on the implementing PR | Researcher-visible fix in a release ([`docs/versioning.md`](../../../versioning.md)) | No bump (docs-only) — this PR ships code |
| D-8 | Promote page numbering to one internal helper: `ArtifactViewerModel.pageNumber(for page: PDFPage, in document: PDFDocument) -> Int?` *(new overload or rewrite)*. The existing `pageNumber(for selection:in:)` calls it. `reportUserSelection` uses it instead of its inline formula. The host's page report uses it. Visibility is `internal` (same module as the host; no test calls the helper directly) | Three call sites already encode the same `index` / `NSNotFound` / `+ 1` rule | A fourth copy in the host; making the helper `public` |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `onPageChange: (Int) -> Void` on `ArtifactPDFViewport` / host *(new property)* | property | `onZoomChange: (CGFloat) -> Void` on the same types | Extend the same pair: add the sibling callback; do not invent a second channel |
| `suppressPageCallback` *(new flag)* | flag | `suppressZoomCallback` on `ArtifactPDFHostView` | Same suppress-around-write pattern for the new callback |
| `ArtifactViewerModel.pageNumber(for page:in:)` *(new overload)* | func | private `pageNumber(for selection:in:)` and the inline formula in `reportUserSelection` | Extend: one `PDFPage` helper; selection and selection-report call it; do not add a host-local copy |
| `ArtifactPDFViewportPageTests.swift` *(new)* | file | `ArtifactViewerModelTests.swift` (page/zoom model) | Host notification + `sync` only; model `setPage` no-op stays in `ArtifactViewerModelTests` (T-4) |

**Builds on.** Zoom feedback in `ArtifactPDFViewport` (`2169a09e`); `ArtifactViewerModel.setPage` / `clampPage` / `pageNumber`; `ArtifactFileFixtures.writeBlankPDF`; `waitUntil`; prefer `ArtifactPDFHostView` as an `NSWindow` `contentView` (D-6).
**Deletes.** `ArtifactViewerModel.pdfPageMediaSize` and the two `loadsPDFAndPages` expects that read it (last production caller was `contentIdentity`). The stale "page only from chrome" behavior goes away with no other symbol to remove.
**Next adds.** S14-29 polygon edit can rely on `model.page` matching the drawn page so stamps and `committedOnCurrentPage` agree.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| PDFKit page change while a PDF Artifact is shown | `ArtifactViewerModel.setPage` via `onPageChange` | `setPage` no-ops when unsupported or unchanged; unload resets `page` to 1; missing `currentPage` / `NSNotFound` → no callback |
| Scroll-driven page with a committed region on another locator page | `committedOnCurrentPage` | Already hides when `locatorPage != viewerPage`; once viewer page is honest, hide/show tracks scroll (existing behavior, now correct) |

### Scale

None — no catalog write, list, or loop. Page reports are O(1) per PDFKit notification.

### Edges

- **Assumptions.** `PDFView.displayMode == .singlePage` (unchanged); `currentPage` is the authoritative visible page for locators; 1-based page indices match chrome and `CitationLocatorDraft`. Held by existing Find/`setPage` tests plus T-1.
- **Neighbors.** Chrome `setPage(3)` then zoom: still works (`page:` drives `sync`). Find `activateHit` then viewport PageChanged for the same page: `setPage` no-op (Find's `go(to: selection)` stays without suppress — D-3). Image Artifact: no `onPageChange` wire. Empty PDF / nil `currentPage`: no report.
- **Removed inputs.** None — this PR does not decide, exclude, or delete catalog inputs.
- **Real data.** None — not table/seed driven.
- **Untrusted size.** None — page index is a single Int from PDFKit for a document the viewer already loaded through `ProjectFiles`.

### Schema

None — no catalog change.

### Go core

None — macOS-only viewer sync.

### FFI

None — no RPC.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| PDF page feedback | `ArtifactPDFViewport` / `ArtifactPDFHostView` (existing snowflake) | `macos/App/Features/ArtifactViewer/` | Add `onPageChange` + suppress; no new kit primitive |
| Wire to model | `ArtifactViewer` | same | `onPageChange: { model.setPage($0) }` |
| Page chrome | `ArtifactViewerToolChrome` | same | Unchanged; follows `model.page` via existing `onChange` |

States: loading / empty / image paths unchanged. Accessibility: page field already labelled; value updates with `model.page`.

### Strings

None — no new or changed user-facing copy. Existing page field / Set page strings stay.

## 7. Steps

**STEP-1 — Page-number helper, host report, and viewer wire**
Files: `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift`, `macos/App/Features/ArtifactViewer/ArtifactPDFViewport.swift`, `macos/App/Features/ArtifactViewer/ArtifactViewer.swift`. Skill: —.
Do: Add `ArtifactViewerModel.pageNumber(for page: PDFPage, in document: PDFDocument) -> Int?` (D-8); route the selection overload and `reportUserSelection` through it. Update the file header comment on `ArtifactPDFViewport` so it says chrome and Find still drive `go(to:)`, and scroll/swipe report the visible page through `onPageChange`. Add required `onPageChange` through `ArtifactPDFViewport` → `Coordinator` → `host.onPageChange` exactly as `onZoomChange` is wired (no default closure). On `ArtifactPDFHostView`, add `onPageChange` and `suppressPageCallback`. In `pdfLayoutChanged`, when the notification is `.PDFViewPageChanged` or `.PDFViewVisiblePagesChanged` and `!suppressPageCallback`, if `let page = pdfView.currentPage`, `let document = pdfView.document`, `let number = ArtifactViewerModel.pageNumber(for: page, in: document)`, call `onPageChange?(number)` — no host clamp. In `sync`, wrap only that method's `pdfView.go(to:)` with `suppressPageCallback = true/false` (not Find's `applyFindHighlight`). In the same step, on the PDF `ArtifactPDFViewport(` call site in `ArtifactViewer`, add `onPageChange: { model.setPage($0) }` beside `onZoomChange` so the scheme still builds (required parameter, sole call site).
Check: on macOS, Provenencia scheme builds; on any host, `rg onPageChange macos/App/Features/ArtifactViewer/ArtifactPDFViewport.swift` shows the property and the two notification branches; `rg 'onPageChange: \\{ model.setPage' macos/App/Features/ArtifactViewer/ArtifactViewer.swift` hits; `rg suppressPageCallback` shows the sync wrap only; `rg 'pageNumber\\(for' macos/App/Features/ArtifactViewer/` hits the helper, selection overload, `reportUserSelection`, and the page report.

**STEP-2 — Fix content identity, delete pdfPageMediaSize**
Files: `macos/App/Features/ArtifactViewer/ArtifactViewer.swift`, `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift`, `macos/ProvenenciaTests/ArtifactViewerModelTests.swift`. Skill: —.
Do: Replace `contentIdentity` with D-4's string. Delete `pdfPageMediaSize` and the two expects in `loadsPDFAndPages`.
Check: `rg pdfPageMediaSize macos/` has no hits; `contentIdentity` has no `model.page`.

**STEP-3 — Tests**
Files: `macos/ProvenenciaTests/ArtifactPDFViewportPageTests.swift` *(new)*, `macos/ProvenenciaTests/ArtifactViewerModelTests.swift`. Skill: `add-swift-test`.
Do: Implement T-1 – T-3 in a new `@MainActor` `@Suite` (same as `ArtifactViewerModelTests`) per §8 and D-6 (bootstrap `sync` args, `waitUntil` before assertions that need a callback). Implement T-4 in `ArtifactViewerModelTests` beside `loadsPDFAndPages`, loading via that suite's existing `minimalPDFData(pageCount: 3)` + `writeFile` + `load` path. Do not add a new locator stamp test (AC-3 covered by existing `CitationLocatorDraftTests` once T-1 keeps `model.page` honest — manual check 4).
Check: on macOS, `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/ArtifactPDFViewportPageTests` and the T-4 case in `ArtifactViewerModelTests` pass. On Linux agents, the `rg` checks from STEP-1/2 plus reviewer's macOS CI are the gate (`scripts/check-pr.sh` skips Xcode off macOS).

**STEP-4 — Version bump**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION` Debug and Release). Skill: —.
Do: Bump patch together (at planning base: `0.2.13` → `0.2.14`, or whatever `VERSION` reads on the implementer's `main`).
Check: the three places match; `rg MARKETING_VERSION macos/Provenencia.xcodeproj/project.pbxproj` shows the new value twice.

**STEP-5 — Docs**
Files: this plan (status + Landed note), `docs/deployment-plan/spike-14/deployment-plan.md` (map row), `docs/deployment-plan/spike-14/completed.md`.
Do: per Definition of done.
Check: links resolve; map Plan column still points here; Status `Landed in #N`.

Commits: one per behavior. Suggested subjects: "Share one PDF page-number helper across Find, selection, and scroll report."; "Report the visible PDF page back into the Artifact viewer model."; "Drop page from PDF content identity and delete pdfPageMediaSize."; "Cover PDF page feedback and zoom-without-snap in tests."; "Bump version to 0.2.14." Docs commit as needed. Full imperative sentences ending in a period.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/ArtifactPDFViewportPageTests.swift` `pageChangedNotificationUpdatesCallback` *(new)* | Bootstrap per D-6; `onPageChange` records Ints; `go(to:)` index 2 → `waitUntil` callback `3`; then `go(to:)` index 0 → callback `1` | AC-1.1, AC-1.2, D-1, D-2, D-8 |
| T-2 | Swift | same file `syncZoomOnCurrentPageDoesNotSnapToPageOne` *(new)* | Host on page 3; `sync(document:page:3,zoom:2.0,resetPage:false,selectionEnabled:true,findSelection:nil,findActivationID:0)`; assert `document.index(for: currentPage!) == 2` and `abs(scaleFactor - 2) < 0.01` | AC-2.1, D-3, D-6 |
| T-3 | Swift | same file `pageReportThenSyncZoomKeepsPage` *(new)* | Local `var page = 1` from `onPageChange`; `go(to:)` page 3; `waitUntil { page == 3 }`; then `sync(…, page: page, zoom: 1.5, resetPage: false, selectionEnabled: true, findSelection: nil, findActivationID: 0)` → still page index 2 | AC-1.1, AC-2.1, AC-2.2, D-5 |
| T-4 | Swift | `ArtifactViewerModelTests.swift` `setPageNoOpWhenUnchanged` *(new)* | Load 3-page PDF via that file's `minimalPDFData(pageCount: 3)` + `writeFile` + `load`; `setPage(2)`; `setPage(2)` again → `page == 2` | AC-1.3, D-5 |

Must stay green: everything `scripts/check-pr.sh` runs; `CitationLocatorDraftTests.pdfRegionStampsPageWhenMissing`, `pdfRegionRestampsPageAfterClearingRegion`, `overlayHidesCommittedRegionOffLocatorPage` (AC-3.1, AC-3.2 once T-1 keeps `model.page` honest — manual check 4). Suites most likely to break: `ArtifactViewerModelTests` (`pdfPageMediaSize` expects removed in STEP-2).

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
- [ ] Domain doc updated if a rule in docs/*.md changed: none owed
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a multi-page PDF Artifact (3+ pages). Open the Citation composer on that Artifact.
2. Scroll or swipe until a later page is in view. See the top-left page field show that page (not stuck on 1).
3. Zoom in and out (AC-2.2). Stay on that page; do not jump to page 1.
4. Arm a region tool, draw a small rectangle on page 5 (or the later page), commit (AC-3.1, AC-3.2). See the locator page chip / Set page state reflect page 5; the committed overlay stays visible on that page; save and confirm the Citation's page is 5.
5. Use chrome next/previous and the page field: they still change the page and the indicator stays in sync when you scroll afterward.
6. Scroll to page 2, pause, scroll slightly within the same page without changing page: the page field stays on 2 (AC-1.3 in the app).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `.PDFViewPageChanged` does not fire for the researcher's scroll gesture on this macOS / PDFKit version | Also listen to `.PDFViewVisiblePagesChanged` (D-2); manual check 2; Stop if neither fires when the visible page changes |
| Removing page from `contentIdentity` breaks an image remount that relied on page in the key | Images keep `page == 1`; identity still includes image pixel size (D-4); image pan/zoom tests stay green |
| Host test flaky without a real window | Size the host in an `NSWindow`, `orderFrontRegardless`, `waitUntil` up to 2s; Stop if T-1 cannot observe a callback after `go(to:)` with observing on |
| `go(to:)` inside T-1 does not post the notification when suppress is wrong | T-1 does not call `sync`'s suppressed path; it calls `pdfView.go(to:)` directly with observing on |

**Stop and report if:**

- `ArtifactPDFHostView.sync` or `pdfLayoutChanged` no longer match §4 (API moved).
- Neither `.PDFViewPageChanged` nor `.PDFViewVisiblePagesChanged` fires when `pdfView.go(to:)` changes the page in T-1 on the CI macOS image.
- Keeping zoom on page N requires preserving the visible rect (Out) or changing `displayMode` (Out).
- A chrome-only page change stops updating the PDF view after D-4 (evidence: `sync` no longer sees `page:` differ from `currentPage`).

## 12. Open questions

None.

## 13. PR

**Title:** S14-28 — PDF page tracks scroll
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-28-pdf-page-tracks-scroll.md`.
