# Review — S13-15 Likeness viewer

**Verdict:** Fits with changes — Not ready (round 1), every finding Fixed. The plan stays `Blocked — on board S13-D5 and on S13-14`; it is refreshed and reviewed again after both.
**Rounds:** 1 (round 2 below). **Reviewers:** maintainability (Opus), executability (Sonnet). One reviewer per lens covered S13-D5, S13-14 and S13-15 together.

## To user
None.

## Writer's own calls
- The viewer is the kit sheet, form (A) (D-4). If the board picks an overlay, the implementer stops and asks.
- A likeness is never drawn above its pixel size ÷ 2 in points, so a 300-pixel face shows at most 150pt wide (D-6, from brief LV-2). The board may revise this.
- The panel's title is the Source title, its subtitle the counter, and close sits in the title row (D-9).
- A 120 ms pause before each full decode, so a held arrow key decodes nothing until it stops (D-5).
- `ArtifactImageDecode` moves to `Platform/` so the viewer and the Artifact viewer share one full-size decode (D-5, STEP-2). The move updates one Go test's path constant.
- The viewer carries its own Toast overlay while open (D-10).

## Round 1

Findings are abridged; the reviewers' full text was returned to the writer.

### M-1 · blocker · The viewer speaks Person-only strings
**Answer:** Fixed. D-12: `ConclusionDetailGallery.ViewerCopy` (`previous`, `next`, `close`, `viewHint`, `label`) sits on the slot, with Person's `.likeness`. Event and Place leave it nil, so the section opens a viewer only when the slot offers one.

### M-2 · blocker · A third way to decode a project image
**Answer:** Fixed. D-5 reuses `ArtifactImageDecode.decode(url:typeIdentifiers:)`, moved to `Platform/`. The `ThumbnailCache` extraction is gone. STEP-2 also updates `api/ffi/swift_vocabulary_test.go` (`swiftArtifactImageDecode`) and `.cursor/rules/macos-client.mdc` line 21.

### M-7 · major · Form (B) would be a new presenter
**Answer:** Fixed. D-4 makes form (A) binding and form (B) a stop-and-ask. The brief now recommends (A) and marks the cost of (B).

### M-8 · major · A third "N of M" key
**Answer:** Fixed. D-9 reuses `L10n.DesignSystem.selectOptionPosition(current:count:)`. The new label key uses `%lld`.

### M-9 · major · Stale copy on the tile image
**Answer:** Fixed. D-12: with a viewer, the image's hint becomes `viewer.viewHint` and its help drops "Open citation".

### M-10 · minor / X-1 · blocker · The item's id changes on every step
**Answer:** Fixed. D-3: the item's `id` is the constant "viewer", while `shownID` and `lastIndex` are separate fields that `show(_:)` sets together. T-3 *(new)* pins the identity.

### X-2 · blocker · The image's size while loading is undefined
**Answer:** Fixed. D-6 reads the header size first, fixes the frame, then draws the cached tile image and then the decoded image inside it.

### X-3, X-4 · major · Extraction contradictions; the decode is not cancellable
**Answer:** Fixed. See M-2. D-5 orders the steps: a cancellable sleep, `objectURL` on the main actor, one detached header read and decode, and the result discarded if cancelled. At most one decode runs at a time.

### X-5 · major · `PVPanel` needs a title
**Answer:** Fixed. D-9 sets the title, subtitle, title accessory and footer. `PVSheetChrome` draws no close control, so the close button is not a duplicate.

### X-9 · major · Focus return and scroll; the opening index
**Answer:** Fixed. D-11 uses a `closeRequest` token with `pvScrollRowIntoView(row:request:)`. D-3 iterates the tiles enumerated, so the button knows its index.

### X-10 · major · The Toast sits under the sheet
**Answer:** Fixed. D-10: the panel carries its own `vocabularyToastOverlay`.

### X-11 · major · Accessibility focus and keys
**Answer:** Fixed. D-8 sets `@FocusState` with `.defaultFocus`, keeps the image area's identity across steps, and speaks each step with `.pvAnnounce(label)`. The label changes with the index, so no repeat is lost.

### X-12 · major · Caption edge cases
**Answer:** Fixed. D-9: an empty `detail` is omitted; the ref line is omitted when `titleIsRef`; the label's third argument is `citation.line`.

### X-13 · minor · D-1 omits `pageIdentifier`; who builds the step
**Answer:** Fixed. D-1 takes `pageIdentifier` and a `step` that the section builds (D-3).

### X-18 · minor · AC-3.3 cannot be run with a modal open
**Answer:** Fixed. AC-3.3 now closes the viewer, then presses Back and restores the project.

## Round 2
Pending: see the end of this file after it runs.
