# Review — S13-15 Likeness viewer

**Verdict:** Fits with changes — Not ready (rounds 1 and 2), every finding Fixed; the next review runs on the refresh after the board and S13-14. The plan stays `Blocked — on board S13-D5 and on S13-14`; it is refreshed and reviewed again after both.
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet). One reviewer per lens covered S13-D5, S13-14 and S13-15 together.

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

Findings are abridged; the reviewers' full text was returned to the writer.

### X2-1 · blocker · D-11 names a symbol and a host that do not exist
**Answer:** Fixed. D-11 uses `.pvScrollsRowIntoView(focusedID, whenRequested: AnyHashable(closeRequest))` on the `LazyVGrid`, with `.id(tile.id)` on each tile, because `PVPageColumns` owns the scroll view.

### M2-4 · major · Escape and a click outside skip the focus return
**Answer:** Fixed. D-11 runs the focus return in `.onChange(of: viewing)` when the item goes from set to nil, so every close route reaches it. D-1 and D-10 close through `PVSheetClose`; the `close()` closure is gone.

### X2-2, M2-6 · blocker · `ViewerCopy.label` breaks `Equatable`
**Answer:** Fixed. D-12: `ViewerCopy` holds only `LocalizedStringResource`s, and `label(index:count:line:)` is a method over `labelFormat`. The new key is a resource.

### X2-3, M2-5 · major · The frame is fixed too late; a second header path; rotation; "one decode"
**Answer:** Fixed. D-5 adds `ArtifactImageDecode.declaredSize(url:typeIdentifiers:)`, split from `decode`'s own sniff, header and size check, with the edges turned for orientations 5 – 8. The frame is published between two detached tasks (D-6), and T-5 *(new)* pins the turned size. D-5 now says "at most one result is shown"; § Scale gives the bound for a crafted file (64 MP).

### X2-4 · major · The item has no home; T-3 has no step and passes trivially
**Answer:** Fixed. The item lives in `ConclusionPictureViewer.swift` with `resolved(in:)`. STEP-1 builds it and runs T-3, which now tests resolution and the constant id.

### X2-8 · minor · The `macos-client.mdc` edit is underspecified
**Answer:** Fixed. STEP-2 gives the exact clause to add to line 21.

## Change after review (user, 2026-11-03)
The viewer shows the likeness's provenance record right under the image, drawn with the provenance sheet's `ReconciliationRecordRows` (D-9). Its Source, Citation and Observation links close the viewer and navigate (D-10). There is no Open citation button and no separate caption.

These Decisions have not been through `review-plan`. They are reviewed on the refresh after the board and S13-14.

## Change after review (user, 2026-11-03, later)
Tiles get no provenance control, no one-record sheet and no context menu. A likeness's provenance is its one record row under the viewer's image (the kit's Record row, S13-D13 PSK-6). Until S13-15 ships, a tile's click opens its Citation, as today. S13-14 no longer depends on S13-36, but S13-15 does. R-13b and R-32 are updated in `person-detail.md`. This is reviewed again on the board refresh.

## Board refresh (2026-11-03)
Board turn 13's viewer answers are in D-4, D-6, D-9 and D-11, and § 12 is empty. The button wording is the app's "Use as portrait" (user, 2026-11-03). The plan stays Blocked on S13-14 and S13-36 and is reviewed again on its refresh.

## Refresh after S13-14 and S13-36 landed (2026-11-03)
Every § 4 row was re-read at `4c245957`. Writer's own calls, for the user to veto:
- **A likeness has no field.** The plan assumed `detail.fields` held one, and it does not: `loadProperties` skips depiction Properties. `Record(outcome:in:)` takes an optional field (D-9), and the tile carries its outcome. No Go change.
- **One focus restore.** D-11 reuses S13-14's `keepFocus()` as `restoreFocus(to:)`; no second scroll token.
- **One hint field.** The Person's `openTip` becomes `likenessViewHint`; `likenessOpenTip` goes (D-12).
- **AC-4.3.** The section leaves the tree when the gallery empties, so the plan no longer promises where focus lands.

### Refresh round 1: maintainability Fits with changes, executability Ready with changes
- **M-1, M-2 · blockers:** Fixed. D-5: `admit` and `uprightSize` are one admission rule. D-12: `viewerLabel` is a per-tile string, through a typed `L10n` function.
- **X-1, X-2 · blockers:** Fixed. STEP-1 adds `outcome`, and the tests build through `tiles(from:…)`. D-6 names every state.
- **M-3, X-8 · major:** Fixed. D-10: `toastInViewer`, one host.
- **M-4 · major:** Fixed. D-2, D-3: no `lastIndex`.
- **M-5, X-9, X-10 · major:** Fixed. `ConclusionPictureImage`, `frame`, T-7, AC-1.5's numbers, a Reach row.
- **X-3 · major:** Fixed. STEP-5 to STEP-8.
- **X-4, X-5, X-6, X-7 · major:** Fixed. D-12, D-9, D-9, D-1.
- **M-6, M-7, M-8, M-9, X-11, X-12 · minor:** Fixed. X-13 Fixed by M-1.
- **Nits:** M-10, M-12, M-14 Fixed. M-11, M-13, M-15 Logged.

### Refresh round 2: maintainability Fits with changes, executability Ready with changes
Rounds are capped at two, so these answers have not been re-reviewed.
- **M-16 · blocker:** Fixed. D-10: the section's `.onDisappear` clears `toastInViewer`, and a close clears the Toast rather than remounting it.
- **M-17 · blocker:** Fixed. The one-line title was the writer's addition, not the board's: D-9 lets a long title wrap, with no kit change.
- **M-18 · blocker:** Fixed. D-5: a `maxEdge` parameter; the viewer passes `ThumbnailDecodePolicy.maxDeclaredEdge`; T-5, Scale, Edges.
- **M-19 · blocker:** Fixed. D-8: `FocusClaim` on appear, no `.defaultFocus`.
- **X-1 · major:** Fixed. D-6 gives the arithmetic; T-7 pins the width bound.
- **X-2 to X-5:** Fixed. D-10, D-6, D-9, D-12.
- **Outside the lens, the band on a short pane:** Fixed. `bandHeight` reads `\.pvPageViewportSize`; T-7, manual check 6.
- **Outside the lens, one decode at a time:** Fixed. The plan now says a superseded decode finishes and is dropped; `ThumbnailCache.withDecodeSlot` is private.
- **Outside the lens, the Toast over the close button:** Accepted in § 11.
