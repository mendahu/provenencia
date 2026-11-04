# Review — S14-28 PDF page tracks scroll

**Verdict:** Fits — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (Grok 4.7), executability (Composer)

## To user
None.

## Writer's own calls
- D-4 deletes `pdfPageMediaSize` once `contentIdentity` stops reading it (last production caller).
- D-8 promotes a shared `pageNumber(for page:in:)` instead of a host-local clamp.
- Host tests use a plain `NSWindow` content view, not `hostInBorderlessWindow` / SwiftUI representable.
- AC-3 relies on existing `CitationLocatorDraftTests` plus manual check 4 once T-1 keeps `model.page` honest; no new stamp test.
- STEP-1 wires the sole `ArtifactPDFViewport` call site with the required `onPageChange` so the step still builds (round 2 X-9).

## Round 1

### M-1 · blocker · The visible page is computed a third time, then clamped against a second count
**Answer:** Fixed — D-2 / D-8 / STEP-1: one `ArtifactViewerModel.pageNumber(for page:in:)`; selection overload and `reportUserSelection` call it; host reports through it with no `clampPage`; `setPage` remains the only clamp.

### M-2 · major · The new content key adds an image dimension and leaves the page-size property unread
**Answer:** Fixed — D-4 / Deletes / STEP-2: identity is kind + pageCount + displayImage width only; `pdfPageMediaSize` and its two expects are deleted.

### M-3 · major · T-4 and T-5 open a second home for model and locator tests
**Answer:** Fixed — T-4 moves to `ArtifactViewerModelTests`; former T-5 dropped; AC-3 cites existing `CitationLocatorDraftTests` + manual check 4.

### M-4 · minor · The viewport header still says only chrome moves the page
**Answer:** Fixed — STEP-1 updates the `ArtifactPDFViewport` file comment.

### X-1 · major · Xcode step checks and DoD tests not runnable here
**Answer:** Declined (environment) — STEP-3 Check now names the Linux/`rg` partial gate and macOS CI; same as other macOS viewer PRs. No plan shape change.

### X-2 · minor · Host test `sync(…)` trailing parameters left open
**Answer:** Fixed — D-6 and T-2/T-3 pin `selectionEnabled: true`, `findSelection: nil`, `findActivationID: 0`.

### X-3 · minor · T-1 host bootstrap before `go(to:)` not spelled out
**Answer:** Fixed — D-6 requires bootstrap `sync(…, page: 1, resetPage: true, …)` then `startObserving()`.

### X-4 · minor · T-3 async gap between `go(to:)` and `sync`
**Answer:** Fixed — T-3 row requires `waitUntil { page == 3 }` before `sync`.

### X-5 · minor · T-4 “loaded with 3-page PDF” without a cited harness
**Answer:** Fixed — T-4 cites `minimalPDFData(pageCount: 3)` + `writeFile` + `load` in `ArtifactViewerModelTests`.

### X-6 · minor · Find `go(to: selection)` vs `suppressPageCallback` scope
**Answer:** Fixed — D-3 and Edges Neighbors: do not wrap Find's `go(to:)`.

### X-7 · nit · T-5 largely duplicates existing locator tests
**Answer:** Fixed — folded into M-3 (T-5 dropped).

### X-8 · nit · Manual check (§10) not runnable in this agent
**Answer:** Logged — expected; implementer/reviewer runs §10 on Mac.

## Round 2

### Maintainability
No new blockers. Verdict: Fits.

### X-9 · blocker · STEP-1 requires a build that STEP-1 cannot produce
**Answer:** Fixed — STEP-1 Files/Do include wiring `onPageChange: { model.setPage($0) }` at the sole call site; STEP-2 is only identity + `pdfPageMediaSize` deletion.

### X-10 · minor · Plan doc fails `check-plan-docs.py` today
**Answer:** Fixed — D-8 drops "if needed"; visibility is `internal`, no direct test call.

### X-11 · minor · Host test suite isolation not named
**Answer:** Fixed — D-6 and STEP-3 require `@MainActor` `@Suite`.
