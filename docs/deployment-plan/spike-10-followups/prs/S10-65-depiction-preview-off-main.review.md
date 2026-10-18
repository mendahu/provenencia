# Review — S10-65 Render the likeness preview off the main thread

**Verdict:** Fits — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (opus, pr-reviewer), executability (sonnet, Explore)

## To user
- None.

## Writer's own calls
- D-1: `ArtifactRegionCrop` splits into a main-actor `snapshot` and a nonisolated `cut` over a `Sendable` `Snapshot`; `render` keeps its name and becomes `async`.
- D-2: the photo snapshot is the viewer's decoded `CGImage`; no ImageIO re-decode.
- D-3: the PDF cut opens its own `CGPDFDocument` from `pdfDocument?.documentURL`.
- D-4: the PDF cut draws only the box at the old scale rule, white-filled, through `getDrawingTransform`, into one shared `outputContext`.
- D-6: an unchanged edit never cuts, with or without a readable saved image, because `saveEdit` sends no image unless the region is redrawn.
- D-8: a Save before the cut lands runs the same deterministic cut; no stored task handle.
- D-9: the cut starts from the value area's `.task(id:)`; an unchanged edit shows `CachedThumbnail` wrapped in the preview label.
- D-12: a private `@unchecked Sendable` wrapper for `CGImage`, only if the build warns at the two crossings.
- D-13: the preview key is computed twice per `body` pass and not cached.
- STEP-1 and STEP-2 are one commit; no version bump; no domain-doc change.

## Round 1

### M-1 · blocker · `ArtifactViewerModel.fileURL` stores a fact `PDFDocument.documentURL` already holds
The plan added `fileURL` with three hand-placed resets and missed `installImageRasterForTesting`; `PDFDocument(url:)` already carries `documentURL`, which `ArtifactViewerContent` reads as identity, and `pdfDocument` derives from `content`.
**Answer:** Fixed — D-3 reads `viewer.pdfDocument?.documentURL`; `fileURL` and its resets are gone; `ArtifactViewerModel.swift` drops out of the Files list; T-3 checks only `snapshot` (nil before load, `.pdf` with `pageIndex == 0` and the document's URL after, nil after `unload()`); §4 viewer row records the fact.

### M-2 · blocker · The PDF box draw copies `scaledToFit`'s bitmap context
**Answer:** Fixed — D-4 adds `private static func outputContext(width:height:)` (sRGB, 8-bit, premultipliedLast, `.high`); `scaledToFit` and the PDF cut both use it; Codebase fit and Deletes updated; Next adds names it for "Transparent polygon".

### M-3 · major · D-6 and D-7 each write out the "unchanged saved likeness" rule
**Answer:** Fixed — D-6 adds one `private var unchangedDepictionImagePath: String?`; `depictionPreviewKey` and `depictionPreviewState` (D-7) both read it. See the outside-lens answer below for why it does not require a non-empty path.

### M-4 · minor · `depictionPreviewKey` property and `depictionPreviewKey(for:)` func share a name with different nil rules
**Answer:** Fixed — the raw func is `depictionCutKey(for:)` (D-6, D-8, Codebase fit).

### M-5 · minor · `ArtifactRegionCrop.Source` reuses the domain's central noun
**Answer:** Fixed — `enum Snapshot: Sendable` (D-1, Codebase fit, stop condition).

### M-6 · minor · The memo encodes "failed" as two optionals both nil
**Answer:** Fixed — D-5: `cut: (cgImage: CGImage, image: NSImage)?`, nil = failed; T-5 reads `depictionPreviewMemo?.cut?.cgImage`.

### M-7 · minor · A doc comment and the map row become false
**Answer:** Fixed — STEP-1 rewrites the `ArtifactRegionCrop` header and the `render` / `cut` doc comments (exact header text given). The map row text goes to the writer of the map: "The composer cuts a likeness preview once per region, off the main actor, drawing only the face box of a PDF page; an edit with no redraw shows the saved image (audit PRF-1; Medium item 7)".

### M-outside-1 · The `.saved` state reads `thumbnailLoading` and shows a file mark when missing
`CachedThumbnail` sets `Content(loading: true)` with no label and falls back to the PNG file-type mark; `imageLabel` labels only the image state.
**Answer:** Fixed — D-9 wraps the `.saved` `CachedThumbnail` in `.accessibilityElement(children: .ignore)` + `.accessibilityLabel(depictionPreviewLabel)`, so the label holds in every state; the missing-file mark is accepted and stated (§6 States), matching the saved line.

### M-outside-2 · An unchanged edit with an empty `imageRelPath` cuts a preview that Save never sends
Checked `saveEdit` (`CitationComposerModel+Authoring.swift`): with `isDepictionRedrawn` false it commits with no depiction input, so Go keeps whatever it has and no new image is sent.
**Answer:** Fixed — the rule is "an unchanged edit never cuts": `unchangedDepictionImagePath` returns the path even when empty; a non-empty path shows `.saved`, an empty one shows `.none` (D-6, D-7). New AC-3.3 and T-10 pin it. Requiring a non-empty path (the suggested shape) was rejected because it would make the preview and Save disagree.

### X-1 · blocker · STEP-1 and STEP-2 checks cannot pass before STEP-3 and STEP-4
**Answer:** Fixed — the app edits (crop, model, value area) are one STEP-1 with one build check at its end; all test edits are STEP-2 with the test run; docs are STEP-3. The commit paragraph says STEP-1 and STEP-2 are one commit because the test bundle reads the old API until STEP-2.

### X-2 · major · D-4's PDF box draw leaves geometry choices open
**Answer:** Fixed — D-4 lists the draw as six code-shaped steps in call order (and says the transform maps a point last-call-first), gives the x and y translation, uses `abs(Int(page.rotationAngle)) % 180 == 90`, uses `max(1, …)` for a sub-pixel box (as `scaledToFit`), sets no clip, and puts `guard region.isValid` at the top of `cut` (D-1). Interpolation comes from `outputContext`.

### X-3 · major · The manual check and D-9 disagree on the saved-face flash
**Answer:** Fixed — manual check step 4 now says "from the thumbnail cache, with no cut (no loading state after the first frame)"; §6 States names the missing-image fallback (the PNG file-type mark).

### X-4 · major · Sendable and warning checks
**Answer:** Fixed — STEP-1 names all four files; it builds `main` first and tees the log, then compares the count of `Sendable` / isolation warnings in those files against the base. D-12 decides the fallback if `CGImage` is reported non-`Sendable`: a private `@unchecked Sendable` wrapper at the two crossings, no other suppression. The stop condition fires only on an error or a warning D-12 does not remove.

### X-5 · minor · D-8 leaves small behaviors open
**Answer:** Fixed — D-8 says it returns nil when the cut key is nil, writes `cut: nil` on a failed cut, and compares `depictionCutKey(for:)` after the await; D-6 says the cut key is nil without a valid region; D-7 says `DepictionPreviewState` lives in `CitationComposerModel+Depiction.swift`.

### X-6 · minor · T-7 does not match AC-3.2
**Answer:** Fixed — T-7 redraws on the `editingSavedLikeness()` model; the hidden-region model stays as an extra `.none` assertion.

### X-7 · minor · The Overlap paragraph names only S10-63
**Answer:** Fixed — a second overlap paragraph names S10-62 (`load`'s guard, `linkToOutsidePNG`, the reopen after `objectURL`, accepted per S10-62 D-4) and S10-64 (`writeImage`), with the landing-order rule.

### X-8 · minor · Manual check step 3 cannot be done as written
**Answer:** Fixed — step 3 screenshots the preview before Save and compares it with the edit's saved image.

### X-9 · minor · Fixture and test details are open
**Answer:** Fixed — D-11 moves the `pixels` body unchanged (DeviceRGB, premultipliedLast, `bytesPerRow: width * 4`, `[]` on failure); STEP-2 adds `import ImageIO` to the depiction tests; `fileURL` is gone (M-1); T-3 pattern-matches with `guard case`.

### X-10 · minor · The header status contradicts the Planned line
**Answer:** Declined — the plan stays `Draft` until review-plan passes (plan-pr step 8); the Planned line records the base commit the plan was written against, as the other S10-6x plans do.

### X-temptations
**Answer:** Fixed — Out lines for the OCR crop, clearing the memo on close or Citation switch, and locked or encrypted PDFs (accepted as `.failed`); the failed-cut message stays with UX-6. The double key computation is a Decision (D-13: not cached).

## Round 2

Maintainability: every round-1 answer holds (M-1 – M-7 and both outside-lens notes resolved). No new blocker. Verdict: Fits. Two non-blocking notes, answered below.
Executability: every round-1 answer holds; every Current state row true at `e97cf54`. Verdict: Ready with changes.

### Maintainability note · `docs/ideas/depictions-and-likenesses.md` line 11 says the Mac draws the PDF region with PDFKit
**Answer:** Fixed — same as X-4.

### Maintainability note · The accessibility label is added by hand on `CachedThumbnail` at two call sites
**Answer:** Logged — Out line "not planned — move the label into `CachedThumbnail` when a third caller adds one".

### X-1 · blocker (environment) · Every Swift check needs a Mac
**Answer:** Declined — host-only; Stop condition 1 already stops without a macOS host with Xcode. Not a plan defect.

### X-2 · minor · D-12 does not say where `SendableImage` goes
**Answer:** Fixed — D-12: `case image(SendableImage)`, a `SendableImage?` task result, `render` unwraps it; `cut` still takes a `Snapshot`; T-3 unchanged.

### X-3 · minor · D-8 does not say what `depictionImage(for:)` returns when the memo write is skipped
**Answer:** Fixed — D-8: it returns the cut either way.

### X-4 · minor · STEP-3 says no doc changes, but the ideas file's Windows bullet becomes false
**Answer:** Fixed — STEP-3 edits the bullet to "Spike 10 draws it with Core Graphics on the Mac."

### X-5 · minor · The saved-tile loading wording in §6 States and manual check 4
**Answer:** Fixed — §6 States: first frame is the empty dash, a hit shows the image at once, a miss shows the loading bar until the decode ends; manual check 4: no cut loading bar, the saved face shows when the cache returns it.

### X-6 · nit · "Draws only the face box" may not avoid decoding the whole scan
**Answer:** Logged — the cut is off the main thread either way; the PR body makes no promise about decode cost.

