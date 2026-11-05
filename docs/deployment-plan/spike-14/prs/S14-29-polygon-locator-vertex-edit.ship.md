# Ship log — S14-29 Polygon locator vertex edit

**Plan:** [`S14-29-polygon-locator-vertex-edit.md`](S14-29-polygon-locator-vertex-edit.md) · **Branch:** `s14-29-polygon-locator-vertex-edit` · **PR:** #715
**State:** ready

## Round 0 — implement
Head `cdd98d6b`. Gate PASS (orchestrator run: xcode-test 50s, all nine checks). Report deviations (recorded in the plan's Landed note): `ArtifactRegionTarget` is a `typealias` for the planned tuple; `ArtifactFileFixtures.pdfHost` gives `onPageChange` a no-op default; Manual check not run (needs a GUI session).

## Round 1 — review
Base `5d53a09c`, head `cdd98d6b`. Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`. SEC+PRF: locator code, and hit-testing that loops over stored polygon points.

CON and SEC+PRF: clean. SEC+PRF confirms `decodeRegion` clamps points and caps them at 256 before the overlay sees them.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1, UX-4 | MNT, UX | major (MNT rated blocker) | Overlay draws a region the model refused | Fix now → F-1 | Verified: `onCommit` returns `Void`, `commitViewerRegion` returns false while transcribing and on `.needsPage` / `.wrongPage`; overlay stays live. Re-rated major: self-heals on the next model-driven `updateNSView`; Fix now either way. Not contradicting D-5/D-6: no model change, only the discarded `Bool` is read |
| UX-2 | UX | major | Pointer jitter edits the region; grabbed vertex jumps to the pointer | Fix now → F-2 | Verified: `mouseDragged` `.resize` ignores `edit.origin`; `draft == start` is the only guard, `ArtifactRegionDraft` compares exact points. AC-4.4 intent "a stray click must not touch a saved Citation". Adds to D-5, does not contradict it |
| MNT-2 | MNT | major | Cursor mapping has two switches and two inputs | Fix now → F-3 | Verified: `applyCursor` passes the pointer as `handle:`, `resetCursorRects` passes `target.normalized`. Outline stays a `hitHandle` branch (D-2); that part of the finding is not taken |
| MNT-3 | MNT | major | T-8 adds a third inline region locator JSON | Fix now → F-4 | Same-pass cleanup. Lines 29, 718, 738 of `CitationComposerModelTests.swift` |
| MNT-4 | MNT | major | `near` copies `nearlyEqual` | Fix now → F-4 | Same-pass cleanup. `ArtifactRegionOverlayTests.swift:81` |
| UX-3 | UX, MNT outside | minor | Open-hand cursor sticks after leaving the outline | Fix now → F-5 | Verified: `case nil: break`, no outline cursor rect. Defect in code this PR adds |
| MNT-6 | MNT | minor | Tuple typealias built positionally | Fix now → F-6 | Struct contradicts D-2's tuple. Labelled construction closes the swap risk and keeps D-2 |
| MNT-7 | MNT | minor | `hostInBorderlessWindow` overloads repeat window construction | Fix now → F-6 | Same-pass cleanup; `HostedWindow.swift` |
| MNT-8 | MNT | minor | `add-swift-test` omits the new host helpers | Fix now → F-7 | Helpers this PR adds |
| MNT-9, UX-5 | MNT, UX | minor | Out items sent to the dogfood log have no record there | Fix now → F-7 | Verified: `docs/dogfood/ux/` holds only its README |
| MNT-10 | MNT | minor | Patterns doc says GraphCanvas is evidence-graph only | Fix now → F-7 | `docs/macos-client-patterns.md:27`, `GraphCanvas/README.md:137` |
| UX-1 | UX | major | A refused edit beeps with no reason and no announcement | To user | Contradicts D-5's rejected "a new L10n announcement" and §6 Strings "None". Cap refusal pairs the beep with an announcement |
| UX-6 | UX | minor | Tooltip implies a delete always works above three points | To user | Same string decision as UX-1 |
| MNT-5, UX-7 | MNT, UX | minor | Anchor `radius + 1` is a document point, not a screen point | To user | Contradicts D-2's `radius + 1`; conflicts with D-4's "sizes in screen points" |

Nits (logged): `CON-1 · ArtifactRegionOverlayTests.swift vertexDragCommitsOnce · no outline click-without-drag case` · `CON-2 · ArtifactMediaViewport.swift:289 · magnification write skipped during layout; next pass corrects it` · `UX-8 · ArtifactRegionOverlay.swift:227 · delete refused at three points stays silent; §1 pins this` · `UX-9 · ArtifactRegionOverlay.swift:222 · right-click near an outline swallows PDFKit's menu` · `MNT-11 · ArtifactRegionOverlayTests.swift:16 · unused kind: parameter` · `MNT-12 · ArtifactRegionOverlayTests.swift:2 · unneeded import PDFKit` · `MNT-13 · ArtifactPDFViewportPageTests.swift:36 · { _ in } is now the default` · `MNT-14 · ArtifactFileFixtures.swift:150 · pdfHost fixed size` · `MNT-15 · ArtifactRegionOverlay.swift:393 · screenPoint reads like a CGPoint`

**To user, answered 2026-11-05.** UX-1 and UX-6: "Announce and fix tooltip". Add the announcement and reword the tooltip; this supersedes D-5's rejected "a new L10n announcement" and §6 Strings "None". MNT-5, UX-7: "Scale it"; the anchor's extra reach is one screen point, amending D-2. All three become Fix now → F-1, F-8.

### Fix plan 1

Record each user amendment (D-2, D-5, §6 Strings) in the plan's Landed note, marked "user, 2026-11-05". Do not rewrite the Decision rows.

**F-1 — Every refused region edit reverts, beeps and announces** (from MNT-1, UX-4, UX-1)
Files: `ArtifactRegionOverlay.swift`, `ArtifactMediaViewport.swift`, `ArtifactPDFViewport.swift`, `ArtifactViewer.swift`, `L10n.swift`, `Localizable.xcstrings`, callers in tests. Skill: `add-localized-string`.
Failing test first: `ArtifactRegionOverlayTests/refusedCommitRevertsEdit` — `onCommit` returns false; a valid vertex drag and a valid right-click delete each leave `committed` equal to the start draft, with one `onCommit` call each.
Do: `onCommit` and the `onCommitRegion` closures on both viewports, their coordinators and `ArtifactViewer` become `(ArtifactRegionDraft) -> Bool`. `commitViewerRegion` already returns the `Bool`, so `CitationComposerViewerPane` does not change and no model code changes (D-6 stands). Coordinators return `coordinator?.onCommitRegion?(draft) ?? false`. `finishEdit` becomes: `draft == start` → nothing; `draft.isValid` and `onCommit?(draft) == true` → `committed = draft`; otherwise `committed = start`, then a private `refuseEdit()` that beeps and posts `AccessibilityAnnouncement.post(L10n.ArtifactViewer.regionEditRefused)` the way the cap refusal does (`:277-282`). Copy: "The region can't be changed that way, so it was put back." `commitAndDisarm` keeps its behaviour and discards the result with `_ =`: `onDisarm` drives the model update that corrects it. The silent nil from `deleteFreeformVertex` at three points stays silent (§1).
Check: the test passes; `invalidEditRevertsAndDoesNotCommit` still passes; `scripts/check-pr.sh` PASS.

**F-2 — A press moves nothing until it is a drag, and a vertex keeps its grab offset** (from UX-2, CON-1)
Files: `ArtifactRegionOverlay.swift`, `ArtifactRegionOverlayTests.swift`.
Failing test first: `ArtifactRegionOverlayTests/subThresholdDragDoesNotEdit` — a press on a vertex and one on the outline, each moved under the threshold and released: no commit, `committed` unchanged. `vertexKeepsGrabOffset` — press 5 points from a vertex, drag 20 points: the vertex moves by exactly the 20-point delta.
Do: the `edit` tuple gains `moved: Bool`, false at `mouseDown`. In `mouseDragged`, while `moved` is false, return until `GraphCanvasPointerHitTesting.isDrag(from: pressDocument, to: document, threshold: GraphCanvasPointerHitTesting.dragThreshold * screenPoint)`, where `pressDocument` is `edit.origin` through `documentPoint(normalized:imageRect:)`; then latch `moved = true`. `.resize(index)` moves to the start handle plus the pointer delta: `handlePoints(for: edit.start, imageRect:)[index]` (range-guarded) + (normalized − `edit.origin`), clamped into the unit square. `.move` is unchanged.
Check: both tests pass; `vertexDragCommitsOnce` still passes (AC-1.3 presses exactly on the vertex); `scripts/check-pr.sh` PASS.

**F-3 — One handle-to-cursor mapping** (from MNT-2)
Files: `ArtifactRegionOverlay.swift`.
Failing test first: none (no cursor seam); the plain refactor is held by `hitHandle*` and the overlay tests.
Do: add one private `cursor(for target: ArtifactRegionTarget, centroid: CGPoint) -> NSCursor` holding the only `.move` / `.resize` switch and reading `target.normalized`. `resetCursorRects` calls it per target. `applyCursor` takes the hit handle, looks up `targets.first { $0.handle == handle }` and calls it; an outline hit with no target is `.openHand`. One centroid expression, not two. `drawCommitted` iterates `handleTargets`: `.resize` handles first, the anchor last, so the drawing order stays. The outline stays a branch of `hitHandle` (D-2).
Check: `rg 'moveAnchor' macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` prints one line; `scripts/check-pr.sh` PASS.

**F-4 — Test copies** (from MNT-3, MNT-4)
Files: `RegionFixtures.swift`, `CitationComposerModelTests.swift`, `ArtifactRegionOverlayTests.swift`.
Do: add `RegionFixtures.locatorJSON(points:page:)`; `locatorJSON(ringCount:)` calls it. Move lines 29, 718 and 738 of `CitationComposerModelTests.swift` onto it and delete the `regionLocatorJSON` constant. Delete `near`; call `ArtifactRegionGeometry.nearlyEqual(a, b, epsilon: 1e-4)`.
Check: `rg '"type": ?"region"' macos/ProvenenciaTests` prints only `RegionFixtures.swift`; `rg 'func near\(' macos/ProvenenciaTests/ArtifactRegionOverlayTests.swift` prints nothing; `scripts/check-pr.sh` PASS.

**F-5 — The overlay restores the cursor it set** (from UX-3)
Files: `ArtifactRegionOverlay.swift`.
Failing test first: none (no cursor seam); covered by Manual check on an image at 100%.
Do: a private `didSetCursor: Bool`, true when `applyCursor` sets a cursor for a hit. In `case nil`, when it is true, call `NSCursor.arrow.set()` and clear it. When it is false, do nothing, so the document's pan hand and PDFView's I-beam stay untouched.
Check: `scripts/check-pr.sh` PASS.

**F-6 — Small code copies** (from MNT-6, MNT-7)
Files: `ArtifactRegionGeometry.swift`, `HostedWindow.swift`.
Do: keep D-2's tuple; build it with labels in `handleTargets` (`(handle: .move, normalized: anchor, document: …)`). Add a private `borderlessWindow(size:)` that both `hostInBorderlessWindow` overloads call.
Check: `scripts/check-pr.sh` PASS.

**F-7 — Docs and guidance** (from MNT-8, MNT-9, UX-5, MNT-10)
Files: `.agents/skills/add-swift-test/SKILL.md`, `docs/dogfood/ux/2026-11-05-region-vertex-editing.md` *(new)*, `docs/macos-client-patterns.md`, `macos/App/Features/GraphCanvas/README.md`.
Do: name `sizedPDFDocument`, `pdfHost` and `hostInBorderlessWindow(nsView:size:)` in the skill's fixtures line. Add one dogfood note in the README's format covering: add a vertex by clicking an edge, keyboard nudging of vertices, VoiceOver per-vertex elements. Update the GraphCanvas row ("used by the evidence graph") to include the Artifact overlay's use of `GraphCanvasCoordinates`, and add the Artifact overlay to the stroke-width note at README `:137`.
Check: `python3 scripts/check-plan-docs.py` passes through `scripts/check-pr.sh`.

**F-8 — Tooltip and anchor reach** (from UX-6, MNT-5, UX-7)
Files: `L10n.swift`, `Localizable.xcstrings`, `ArtifactRegionGeometry.swift`, `ArtifactRegionOverlay.swift`, `ArtifactRegionGeometryTests.swift`, `ArtifactRegionOverlayTests.swift`. Skill: `add-localized-string`.
Failing test first: `ArtifactRegionOverlayTests/handleSizesAreInScreenPoints` extended — `magnification = 0.5` → `moveAnchorHitRadius == 16`; `magnification = 1` → 8.
Do: `citationComposer.freeformDeleteVertex` default value becomes "Drag a vertex to reshape, or the outline to move. Right-click a vertex to delete it. At least three points remain, and a shape that crosses itself is put back." `hitHandle` takes `anchorRadius:` in place of `radius + 1`; the overlay passes a new internal `moveAnchorHitRadius = 8 * screenPoint`; geometry tests pass `radius + 1` explicitly, so their cases keep their meaning.
Check: the test passes; `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `3f544d48`. Gate PASS (orchestrator run). F-1 to F-8 done. Most of the fixes share one commit, `87fc1ac6`, because they share `ArtifactRegionOverlay.swift` and the catalog. F-3's `rg moveAnchor` prints three lines, not one: one `moveAnchor(` call in `cursorCentroid(for:)` and two for the F-8 property `moveAnchorHitRadius`; accepted. Manual check not run (needs a GUI session).

## Round 2 — review
Scope `cdd98d6b...3f544d48`. Reviewers: MNT, UX (each had a Fix-now major). CON and SEC+PRF had none.

| ID | Round-1 finding | Round 2 |
| --- | --- | --- |
| MNT-1, UX-4 | Overlay draws a region the model refused | Resolved (MNT, UX) |
| UX-1 | Refusal has no announcement | Resolved (UX) |
| UX-2 | Jitter edits; vertex jumps to the pointer | Resolved (UX) |
| MNT-2 | Two cursor mappings | Resolved (MNT) |
| MNT-3, MNT-4 | Test copies | Resolved (MNT) |
| UX-3, UX-6, MNT-5, MNT-6, MNT-7, MNT-8, MNT-9, UX-5, MNT-10 | Minors | Resolved (orchestrator read the fix commits) |

No new blockers. No Fix plan 2.

Nits (logged): `MNT-16 · ArtifactRegionOverlay.swift:205 · inline min(max()) repeats clampNormalized; the move functions already clamp` · `UX note · ArtifactRegionOverlay.swift mouseDragged · threshold is measured from a clamped origin, so a press outside the media rect on an edge vertex can pass it early; read from code, not reproduced`

## Result
Gate PASS · CI green on `16056f18` · open: none · follow-ups: add a vertex by clicking an edge, keyboard nudging, VoiceOver per-vertex elements (`docs/dogfood/ux/2026-11-05-region-vertex-editing.md`) · to user: UX-1, UX-6, MNT-5 answered 2026-11-05 · CON miss: none · Manual check §10 not run (needs a GUI session)

## User-reported defect, 2026-11-05
From the user's testing on a very high-resolution image, zoomed in: a polygon closed as a triangle on the fourth click. User confirmed the cause and asked for the fix on this PR. `closeEnoughToOrigin` uses 3% of the page in normalized coordinates, so the close zone scales with image size and zoom. No review round: this is one item the user asked for, outside the two-round loop.

### Fix plan 3

**F-9 — The polygon close distance is in screen points** (user, 2026-11-05)
Files: `ArtifactRegionGeometry.swift`, `ArtifactRegionOverlay.swift`, `ArtifactRegionGeometryTests.swift`, `ArtifactRegionOverlayTests.swift`.
Failing test first: `ArtifactRegionOverlayTests/freeformFourthClickNearFirstAddsVertex` — an armed freeform overlay with a 3000×4000 image and `magnification = 1`; three clicks, then a fourth 60 points from the first: four in-progress points, no commit (fails today, closes as a triangle). A fifth click within `handleHitRadius` of the first closes with one commit of the four points; a click on the first point at `magnification = 4` with a pointer 6 screen points away also closes and one 8 screen points away adds a vertex. Add `closeEnoughToOriginUsesDocumentRadius` to the geometry tests.
Do: `closeEnoughToOrigin` becomes `closeEnoughToOrigin(_ point: CGPoint, origin: CGPoint, imageRect: CGRect, radius: CGFloat) -> Bool`, both points normalized, the distance measured between their `documentPoint(normalized:imageRect:)` values; delete the `threshold` default. `handleFreeformClick` passes `imageRect` and `handleHitRadius`, so the close zone equals a handle's grab radius at any zoom, on images and on a PDF (where magnification stays 1 and the overlay is already at screen scale). Nothing else about closing changes: three-point minimum, `isValid`, the armed-tool drawing paths.
Check: the tests pass; `rg closeEnoughToOrigin macos` shows only the new signature; add one sentence to the plan's Landed note, marked "user, 2026-11-05"; `scripts/check-pr.sh` PASS.

## Round 3 — fix (F-9)
Head `c5a68fb1`. Gate PASS (orchestrator run). `freeformFourthClickNearFirstAddsVertex`, `freeformCloseZoneIsHandleRadiusAtAnyZoom` and `closeEnoughToOriginUsesDocumentRadius` added; the two overlay tests failed before the fix. The planned overlay test was split into those two. No reviewer dispatched: one user-requested item, outside the review loop.
