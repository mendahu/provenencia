# Ship log — S10-65 Render the likeness preview off the main thread

**Plan:** [`S10-65-depiction-preview-off-main.md`](S10-65-depiction-preview-off-main.md) · **Branch:** `s10-65-depiction-preview-off-main` · **PR:** #472
**State:** fixing 2

Base: `main` @ `dc67a099` (worktree `.claude/worktrees/s10-65`).

## Round 0 — implement
Head `703cc2c7`. Gate PASS (orchestrator run). PR #472 (draft).
Report deviations:
- **D-6 cut key.** `depictionCutKey(for:)` keys on artifact id, viewer kind, loading flag, page and region points instead of `locator.encodeJSON()`. Verified: `CitationLocatorDraft.encodeJSON()` (line 68) serializes `[String: Any]` without `.sortedKeys`, so the string is not stable and the memo never hit. Intent of D-6 (one key for trigger, read and save) kept. → To user (accept the deviation from D-6's text).
- `main` had moved since the plan (S10-63, S10-62); no Current state row invalidated.
- D-12 `SendableImage` not needed.
- Manual check (§10) not run; needs the user.

## Round 1 — review
Base `dc67a099` · Head `703cc2c7`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Cut key is a hand-joined string; draft types are already `Equatable` | Fix now → F-1 | verified `+Depiction.swift:170-180`; `ArtifactRegionDraft`, `ArtifactViewerKind` are `Equatable, Sendable`. Departs from D-6 text: to user 2026-10-18, user chose the typed struct key |
| CON-1 | CON | minor | PDF cut via `drawPDFPage` may drop annotations the viewer shows | Fix now → F-2 (test first) | low confidence; carries a failing test, so the fix round verifies it. A red test stops the item and comes back to the orchestrator; the rendering is not changed in this round |
| UX-1 | UX | minor | VoiceOver hears the same label while the cut runs | Fix now → F-3 | `PVThumbnail.swift:140` returns the caller's label first; view file already touched by F-1; one modifier; D-9 keeps the label, which a value does not change |
| SEC-1 | SEC | minor | PDF cut reopens the object path without re-running `ProjectFiles.objectURL` | Follow-up | D-3 binds the snapshot to `documentURL` + `CGPDFDocument(url)`; the fix changes D-3's shape. Needs a live local writer. Follow-up: carry `projectDir`/`relPath` in `Snapshot.pdf` and resolve through `objectURL` in `pageBox` |
| UX-2 | UX | minor | Saved tile says "Image to save" on an edit that saves none | Follow-up | needs a new string; plan Strings: None. With UX-6 |
| MNT-2 | MNT | minor | 2,048 px cap applied twice on the PDF path | Rejected | D-4 step 6 asks for the inner `scaledToFit`; it is a no-op, so no harm |
| MNT-3 | MNT | nit | `nonisolated` on only some statics | Logged | — |
| CON-2 | CON | nit | Header comment line 13 is 126 columns | Logged | (also noted by PRF, outside lens) |
| SEC-o1 / PRF note | PRF | — | Detached cut not cancelled on key change | Rejected | D-8/D-10 accept it; bounded by user actions (PRF checked) |
| PRF | PRF | — | No findings | — | — |

### Fix plan 1

**F-1 — Make the cut key a typed `Equatable` value** (from MNT-1; D-6 deviation, user-approved 2026-10-18)
Files: `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift`, `macos/App/Features/CitationComposer/CitationComposerModel.swift` (memo), `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` (`.task(id:)`), the plan's Landed note.
Failing test first: none (a refactor); the existing `previewIsCutAfterTheRegionCommits`, `staleCutIsNotStored`, `editWithoutRedrawShowsTheSavedImage` and `savedPixelsAreThePreviewPixels` must stay green.
Do: add `struct DepictionCutKey: Equatable, Sendable { let artifactID: String?; let kind: ArtifactViewerKind; let isLoading: Bool; let page: Int?; let region: ArtifactRegionDraft }` beside `DepictionPreviewMemo`. `depictionCutKey(for:)` returns `DepictionCutKey?`; `DepictionPreviewMemo.key` and `depictionPreviewKey` take that type; compare with `==`; pass `model.depictionPreviewKey` to `.task(id:)` unchanged. Delete the string join and its "not `encodeJSON()`" comment. Rewrite the Landed note's cut-key bullet: the key is a typed `Equatable` value of artifact id, viewer kind, loading flag, page and region draft, not `locator.encodeJSON()` (unstable key order, no `.sortedKeys`); approved by the user against D-6's text. Do not touch `encodeJSON()`.
Check: tests above pass; scripts/check-pr.sh PASS.

**F-2 — Test that the PDF cut keeps annotations** (from CON-1)
Files: `macos/ProvenenciaTests/…/ArtifactRegionCropTests.swift`, `ArtifactFileFixtures` if a helper is needed. Skill: `add-swift-test`.
Failing test first: `ArtifactRegionCropTests/pdfCutKeepsAnnotations` — write the marked 100×200 pt PDF; through PDFKit add a black-filled `.square` `PDFAnnotation(bounds: CGRect(x: 0, y: 0, width: 100, height: 20))` (interior and border colour black) over a white area at the page bottom; save it; `await render(viewer, page: 1, region: rectangle(0, 0.9, 1, 1))`; expect it to contain black.
Do: if the test passes on the current code, keep it as a guard and report "passes on head". If it fails, do **not** change `pageBox` or the rendering: remove the test, report it **Reproduced** with the output, and continue with F-3. The orchestrator takes it to the user.
Check: test suite green; scripts/check-pr.sh PASS.

**F-3 — Tell VoiceOver the preview is loading while the cut runs** (from UX-1)
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` (`CitationDepictionValueArea.preview`, `.cutting` branch).
Failing test first: none (view modifier).
Do: on the `.cutting` branch only, add `.accessibilityValue(Text(L10n.DesignSystem.thumbnailLoading))`. Keep the label as D-9 says. No new strings.
Check: build; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `704a71cc`. Gate PASS (orchestrator run). F-1 done (`704a71cc`), F-3 done (`113c7640`).
F-2 **reproduced**: `pdfCutKeepsAnnotations` failed (`containsBlack(cut)`, `ArtifactRegionCropTests.swift:138`); test removed, rendering unchanged. CON-1 re-rated **major**: a regression against base (`PDFPage.thumbnail` drew annotations) that breaks US-2 on annotated PDFs. The fix changes D-3/D-4's `CGPDFDocument` + `drawPDFPage` → **To user**.
User 2026-10-18: amend D-3/D-4 — draw the box through PDFKit inside the detached task. CON-1 → Fix now → F-4.

### Fix plan 1 — amendment (user decision on CON-1)

**F-4 — Draw the PDF box through PDFKit so annotations are in the cut** (from CON-1; amends D-3/D-4 on the user's call, 2026-10-18)
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift` (`pageBox`), `ArtifactRegionCropTests.swift`, the plan's Landed note. Skill: `add-swift-test`.
Failing test first: `ArtifactRegionCropTests/pdfCutKeepsAnnotations` — the same test as F-2 (marked PDF, black `.square` annotation at the page bottom, `rectangle(0, 0.9, 1, 1)`, expect black). It must fail on `704a71cc`.
Do: in `pageBox`, open the document inside the task with `PDFDocument(url:)` and `page(at: pageIndex)` instead of `CGPDFDocument`, and draw with `page.draw(with: .mediaBox, to: context)` instead of `context.drawPDFPage`. Keep everything else in D-4: `outputContext`, the y-down box from the visual size, the 4 px/pt and 2,048 px caps, and only the box drawn. Read the media box and rotation from the `PDFPage` (`bounds(for: .mediaBox)`, `rotation`). `PDFPage.draw(with:to:)` may apply the page rotation itself: if so, adjust only the rotation step of the transform so that `pdfRegionFollowsTheDisplayedPage`, `pdfCutDrawsOnlyTheBox` and `savedPixelsAreThePreviewPixels` stay green unchanged. The `PDFDocument` is created and used only inside the task; nothing from the viewer's document crosses. Keep the snapshot `.pdf(url:pageIndex:)` as it is. Add a Landed-note bullet: the PDF cut draws through `PDFPage.draw(with:to:)` on a task-owned `PDFDocument`, not `CGContext.drawPDFPage`, so annotations the viewer shows are in the likeness (CON-1; user's call against D-3/D-4's text).
Stop if: the three existing PDF tests cannot all stay green with the new draw. Report what differs; do not change their assertions.
Check: `pdfCutKeepsAnnotations` red then green; all `ArtifactRegionCropTests` and `CitationComposerDepictionTests` green; scripts/check-pr.sh PASS.

## Round 1 — fix (F-4)
Head `644f743b`. Gate PASS (orchestrator run). F-4 done: `pdfCutKeepsAnnotations` red on `704a71cc`, green on `644f743b`; existing PDF tests unchanged and green. Orchestrator note: STEP-3's idea-doc line now says "Core Graphics" while the cut draws with PDFKit again — for round 2.

## Round 2 — review
Lenses: CON, MNT, UX (had Fix-now findings). Scope `703cc2c7...644f743b`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | — | (round 1) typed cut key | Resolved | `+Depiction.swift:12-18`; no `String` key left |
| CON-1 | CON | — | (round 1) annotations | Resolved | `pdfCutKeepsAnnotations` red → green |
| UX-1 | UX | — | (round 1) VoiceOver loading value | Resolved | `CitationComposerAuthoringBox.swift:974` |
| CON-3, MNT-4 | CON, MNT | major | Idea doc says Core Graphics; the cut draws with PDFKit after F-4 | To user → Fix now → F-5 | `docs/ideas/depictions-and-likenesses.md:13` vs `ArtifactRegionCrop.swift:137`. Not a blocker, so outside the round-2 cap; user 2026-10-18: fix in this PR, no round 3 |

### Fix plan 2 (user-approved; not a round-2 blocker)

**F-5 — Put the idea doc back to PDFKit** (from CON-3, MNT-4)
Files: `docs/ideas/depictions-and-likenesses.md`, the plan's Landed note.
Failing test first: none (docs).
Do: in the "Windows" bullet, change "Spike 10 draws it with Core Graphics on the Mac." back to "Spike 10 draws it with PDFKit on the Mac." so the file has no net diff against `dc67a099`. Add to the Landed note: STEP-3's idea-doc edit was undone by F-4, since the cut draws with PDFKit. Do not edit the plan's STEP-3 or D-10 text.
Check: `git diff dc67a099 -- docs/ideas/depictions-and-likenesses.md` is empty; scripts/check-pr.sh PASS.
