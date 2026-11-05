# Ship log — S14-28 PDF page tracks scroll

**Plan:** [`S14-28-pdf-page-tracks-scroll.md`](S14-28-pdf-page-tracks-scroll.md) · **Branch:** `s14-28-pdf-page-tracks-scroll` · **PR:** #711
**State:** ready

## Round 0 — implement
Head `2e30fb914afbe9c0c0f42ea58453134a2d8676df`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). Report: patch bump is 0.3.1 (from 0.3.0 on current main), not 0.2.14 from planning base; `pdfPageMediaSize` / `contentIdentity` landed with helper/viewer commits rather than a separate commit; manual check §10 not run in implementer pass.

## Round 1 — review
Reviewers: CON (Composer), MNT (Grok 4.7), UX (Grok 4.7), SEC+PRF (Grok 4.7).
Skipped: none — UX matches `macos/App/**`; SEC+PRF matches `macos/App/Features/**` models (`ArtifactViewerModel`) and live PDFKit notification → model feedback.
Base `4cf944ca5077fb118c6976c52a56e542a1a95b1a`. Head `2e30fb914afbe9c0c0f42ea58453134a2d8676df`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Manual check §10 not run; scroll/swipe AC paths unverified | Rejected | ship-pr §8 Finish assigns §10 to the user after ready-for-review; T-1–T-4 + CitationLocatorDraftTests are implementer AC evidence; §10 stays on the user report / Manual check |
| CON-2 | CON | minor | T-2 does not wait for page navigation before `sync` | Fix now → F-2 | verified: `syncZoomOnCurrentPageDoesNotSnapToPageOne` called `go(to:)` then `sync` with no `waitUntil`; T-1/T-3 wait |
| MNT-1 | MNT | minor | Host tests add a third blank-PDF builder | Fix now → F-1 | same-pass cleanup: this PR added `blankPDFDocument` beside `writeBlankPDF` and `minimalPDFData` loops; D-6 named the fixture |

Nits (logged): `CON-3 · ArtifactPDFViewportPageTests.swift · D-6 names writeBlankPDF; suite uses local blankPDFDocument — folded into F-1` · `MNT-2 · ArtifactPDFViewportPageTests.swift:107 · bootstrapHost is marked throws and never throws; drop throws and callers' try`

UX: Clean · SEC+PRF: Clean

### Fix plan 1

**F-1 — One blank-PDF document builder on ArtifactFileFixtures** (from MNT-1; supersedes D-6's "use writeBlankPDF" when the host needs a live `PDFDocument`)
Files: `macos/ProvenenciaTests/ArtifactFileFixtures.swift`, `macos/ProvenenciaTests/ArtifactPDFViewportPageTests.swift`, `macos/ProvenenciaTests/ArtifactViewerModelTests.swift`. Skill: `add-swift-test`.
Failing test first: none (structure cleanup) — prove with `rg` that no private insert loop remains.
Do: Add `ArtifactFileFixtures.blankPDFDocument(pageCount:)` that builds a `PDFDocument` by inserting empty `PDFPage`s. Point `writeBlankPDF` at it (then `dataRepresentation` + write). Replace `ArtifactPDFViewportPageTests.blankPDFDocument` with the fixture. Replace `ArtifactViewerModelTests.minimalPDFData`'s insert loop with the fixture (return `document.dataRepresentation()`). Drop `throws` from `bootstrapHost` if it still never throws (MNT-2 nit riding along).
Check: `rg -nE 'insert\(PDFPage\(\)' macos/ProvenenciaTests` shows only the shared fixture body; `scripts/check-pr.sh` PASS.

**F-2 — Wait for page 3 before zoom sync in T-2** (from CON-2)
Files: `macos/ProvenenciaTests/ArtifactPDFViewportPageTests.swift`. Skill: `add-swift-test`.
Failing test first: make `syncZoomOnCurrentPageDoesNotSnapToPageOne` `async` and assert `#expect(await waitUntil { host.pdfView.currentPage.map { document.index(for: $0) == 2 } ?? false })` after `go(to: pageThree)` before `sync`.
Do: Mirror T-3's wait pattern before the zoom `sync`.
Check: test passes; `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `daee1c69c3551ce0aae5d361aaeea5783bdd15ab`. Gate PASS (orchestrator). F-1 done (`77a6d50c`: sole `insert(PDFPage())` in `ArtifactFileFixtures.blankPDFDocument`; host + model tests call it; `bootstrapHost` no longer `throws`). F-2 done (`daee1c69`: T-2 waits for index 2 before zoom `sync`).

## Round 2 — review
Skipped — no Fix-now blocker or major from round 1 (CON-1 Rejected; CON-2 and MNT-1 were minors; orchestrator verified fix commits).

## Result
Gate PASS · CI pending · open: none · follow-ups: none · to user: — · CON miss: none
