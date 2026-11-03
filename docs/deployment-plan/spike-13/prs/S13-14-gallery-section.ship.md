# Ship log — S13-14 Gallery section

**Plan:** [`S13-14-gallery-section.md`](S13-14-gallery-section.md) · **Branch:** `s13-14-gallery-section` · **PR:** #701
**State:** ready

## Round 0 — implement
Head `564a389c`. Gate PASS (orchestrator re-ran). Report: deviations: `GalleryTileLook` is top-level, not nested in the generic `GalleryTileView` (nested forced `Overlay == EmptyView`); `EventDetailTests` and `PlaceDetailTests` updated to AC-1.4; old section gained only the image hint (D-11), not Return; T-2 loop change landed in the STEP-3 commit. Manual check, frames, VoiceOver not run (no GUI).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. SEC+PRF routed by `macos/App/Features/**` models, grid and `Platform/CachedThumbnail.swift`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | The new section copies the old section's tile and body | Fix now → F-1 | Verified: `tileView` and `body` differ only by look, grid gaps, `.help`, Return, `.contain`. Same-pass cleanup: this PR added the copy. Supersedes D-2's "stays unchanged" and its rejected `style:` (its reason, "every line would branch", stopped holding once D-9's look held the measures) and the map rule's deferral to the Event and Place redesign. No researcher-visible change: Event and Place stay `.small` |
| MNT-5 | MNT | minor | `GalleryTileView` lives in the page file | Fix now → F-1 | Shared type, moves with the section. Supersedes § 6's Home for the tile |
| MNT-2 | MNT | major | Tile label keys now equal `likenessLabel` and `pictureFrom` | Fix now → F-2 | Verified: L10n.swift:9231/9241, 9604/9621. Same-pass cleanup: this PR made them equal. Supersedes § Strings' "Changed" |
| UX-1 | UX | major | Focus does not survive crossing 1600 in the lazy grid | Fix now → F-3 | AC-4.2 unmet by construction: no width handling; the Timeline has `onChange(of: widthClass)` and S13-13's ship log records the drop. The drop itself needs the app (§ 10 step 6) |
| UX-O1 | CON (from UX's Outside my lens) | major | Every image tile wears the gender line, not only the fallback | Fix now → F-4 | Verified: `PVThumbnail.lineColor` takes `colors?.line` for `.image`; D-7 "the tile's fallback wears it", D-9 "the fallback's tint", frame 13a tints tile 7 only |
| UX-2 | UX | minor | A failed Use as portrait leaves focus nowhere | Fix now → F-5 | Verified: focus returns only on `true`; the button disables itself. Copy added by this PR |
| MNT-3 + UX-3 | MNT, UX | minor | `openTip` comments are false | Fix now → F-6 | Verified: L10n.swift:9224, 9597 |
| MNT-4 + CON-1 | MNT, CON | minor | Two stale doc comments | Fix now → F-7 | Verified: `CachedThumbnail.swift:15`, `ConclusionDetailPage.swift:84` |
| MNT-6 | MNT | minor | Kit README omits `pvKeepsPageBand` | Fix now → F-8 | Verified: README.md:191 row |
| MNT-7 | MNT | minor | S13-15 status still says S13-14 unlanded | Fix now → F-8 | Verified: map row 148, S13-15 lines 3, 8, 108 |
| SEC-1 | SEC | minor | The tile count has no bound, contrary to § 6 Edges | Follow-up | `gallery.ForEntities` and the `body`-time `tiles(from:)` predate this PR, which adds a constant per tile (no measurement at a large count). Code unchanged here; the false Edges line is corrected in the Landed note (F-8) |

Nits (logged): `CON-1` merged into F-7 · `MNT-8` and `MNT-9` and `MNT-13` fall out of F-1, F-4 · `MNT-10 · CitationLineDisplay.swift:48 · build line with joined([title, detail])` · `MNT-11 · CitationLineDisplay.swift:47 · titleIsRef restates ReconciliationRecord.swift:62` · `MNT-12 · ConclusionDetailPage.swift:870 · caption mono ternary repeats EvidenceGraphView.swift:1144` · `MNT-14 · PVPageColumns.swift:430 · extension sits after previews` · `MNT-15 · ConclusionDetailPage.swift:855 · caption branches repeat modifiers` · `MNT-16 · PersonDetailTests.swift:180 · eventPicturesKeepTheirWording sits in the Person tests` · `UX-4 · ConclusionDetailPage.swift:197 · useAsLabel names the Source only` · `UX-5 · ConclusionGallerySection.swift:70 · portrait spoken twice`

### Fix plan 1

**F-1 — One gallery section with a look; delete the old section** (from MNT-1, MNT-5; closes nit MNT-13)
Files: `macos/App/Features/Conclusions/ConclusionGallerySection.swift`, `ConclusionDetailPage.swift`, `PersonDetailView.swift`, `macos/App/Platform/CachedThumbnail.swift` (doc only), `docs/deployment-plan/spike-13/deployment-plan.md`, `docs/deployment-plan/spike-13/prs/S13-15-likeness-viewer.md`, the plan's Landed note. Skill: `add-ui-component`.
Failing test first: none, view code; no behavior changes for any caller.
Do: `ConclusionGallerySection` takes `look: GalleryTileLook`. Delete `ConclusionDetailGallerySection`; `ConclusionDetailDefaultPageSections` draws `ConclusionGallerySection(look: .small)`, `PersonDetailSections` draws `.large`. `GalleryTileLook` gains the differences so no view branches on a caller: `columnGap` (`.small` `space6`, `.large` `space8`), `rowGap` (`.small` `space7`, `.large` `space8`), and the image button's `.help(gallery.openTip)` and Return for one look each: `.small` keeps `.help` and has no Return, `.large` has Return and no `.help`; the tile's `.accessibilityElement(children: .contain)` stays `.large` only (AC-1.4: Event and Place behave as before). One `open` closure in the section for the click and Return. Move `GalleryTileLook` and `GalleryTileView` into `ConclusionGallerySection.swift`. Update the section's doc comment.
Docs: in the map, rewrite the working rule "The Person's gallery is redrawn on the Person page only": Event and Place keep the `.small` look on the one section until their redesign, and the header menu sentence stays. In S13-15's plan, rewrite every line that names `ConclusionDetailGallerySection` or says the new section has one caller: S13-15 changes only the `.large` look's open action. The plan's Landed note says F-1 supersedes D-2 and § 6 Home. Leave the landed plans (S13-10, S13-12, S14-02) and the brief alone.
Check: `grep -rn 'ConclusionDetailGallerySection' macos` prints nothing; `grep -n 'tileView\|LazyVGrid' macos/App/Features/Conclusions/ConclusionDetailPage.swift` prints no gallery line; scripts/check-pr.sh PASS.

**F-2 — One label string per kind** (from MNT-2)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `PersonDetailContent.swift`, `ConclusionDetailPage.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`, `EventDetailTests.swift`, `PlaceDetailTests.swift`. Skill: `add-localized-string`.
Failing test first: none, wording is unchanged; T-4 keeps pinning "Likeness from X".
Do: pass `L10n.Conclusions.likenessLabel` and `pictureFrom` as `tileLabel:`. Delete `likenessTileLabel`, `pictureTileLabel` and their catalog entries. Widen the kept comments to "the Source title or the citation line". Keep the `Lead` variants. The Landed note says F-2 supersedes § Strings "Changed".
Check: `grep -n '"Likeness from %@"\|"Picture from %@"' macos/App/Platform/L10n.swift` prints exactly two lines; `grep -rn 'likenessTileLabel(\|pictureTileLabel(' macos` prints nothing (the `Lead` names do not match); `python3 scripts/check-localizable-xcstrings.py` passes; scripts/check-pr.sh PASS.

**F-3 — A focused tile survives crossing 1600** (from UX-1)
Files: `ConclusionGallerySection.swift`.
Failing test first: none, view focus; § 10 step 6 is the check by hand.
Do: give each tile `.id(tile.id)`. Add `@Environment(\.pvWidthClass)`, as `PersonTimelineSection`. Hold `@State lastFocused`, set whenever `focusedID` is non-nil and cleared on the next main-actor turn after it goes nil, so a width change in the same update still sees it while a click away clears it. On `.onChange(of: widthClass)` with `focusedID ?? lastFocused` set: set a `@State scrollTarget` to that id, bump a `@State scrollRequest`, and reassign `focusedID` after the scroll's turn. Put `.pvScrollsRowIntoView(scrollTarget, whenRequested: scrollRequest)` on the grid (no new `ScrollViewReader`). No other scroll behavior changes.
Check: `.pvScrollsRowIntoView` and `.onChange(of: widthClass)` present; scripts/check-pr.sh PASS. Report in the Landed note that it is unverified in the app.

**F-4 — The gender colors reach the fallback only** (from UX-O1; closes nits MNT-8, MNT-9)
Files: `macos/App/Platform/CachedThumbnail.swift`, `ConclusionGallerySection.swift`.
Failing test first: none, view code.
Do: add `var fallbackColors: PVThumbnail.Colors? = nil` to `CachedThumbnail`, documented as tint and line only while no raster shows; `tileBody` passes `colors ?? (shown.image == nil ? fallbackColors : nil)`. Do not touch `PVThumbnail` or the header thumb (its `colors` stay). `GalleryTileView` takes `fallbackColors` as a `let` and passes it straight, with no `look` gate; the section passes `gallery.fallbackColors` for both looks (nil on Event and Place).
Check: `grep -n 'look == .large ? colors' -r macos` prints nothing; scripts/check-pr.sh PASS.

**F-5 — Focus returns to the tile whether the save worked or not** (from UX-2)
Files: `ConclusionGallerySection.swift`.
Failing test first: none, view code.
Do: `_ = await lead.choose(tile, copy: gallery.choose); focusedID = tile.id`, keeping `choose`'s `@discardableResult` state as it is.
Check: scripts/check-pr.sh PASS.

**F-6 — `openTip` comments say what the string does** (from MNT-3, UX-3)
Files: `L10n.swift`, `Localizable.xcstrings`.
Do: both comments say it is the spoken hint of the tile's image button, and its tooltip on Event and Place.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**F-7 — Stale doc comments** (from MNT-4, CON-1)
Files: `CachedThumbnail.swift:15`, `ConclusionDetailPicture`'s doc.
Do: name `GalleryTileLook.small`; say "its citation", not "its Source's title".
Check: `grep -rn 'GalleryTileView.Look' macos` prints nothing.

**F-8 — Docs** (from MNT-6, MNT-7, SEC-1)
Files: `macos/App/DesignSystem/README.md`, `docs/deployment-plan/spike-13/deployment-plan.md`, `docs/deployment-plan/spike-13/prs/S13-15-likeness-viewer.md`, the plan's Landed note.
Do: a clause in the README's PageColumns row for `pvKeepsPageBand()`. S13-15's map row and plan Status and Depends lines: blocked on S13-36 only, S13-14 landed. S13-15's tile-data row lists `detail` in `Parts`. The Landed note says § 6 Edges "bounded by the detail read" was wrong: `gallery.ForEntities` has no limit, a follow-up.
Check: `python3 scripts/check-plan-docs.py` reports no errors; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `c4bef1cf`. Gate PASS (orchestrator re-ran). F-1 to F-8 done; none needed a failing test. Deviations: `completed.md` updated; likeness `openTip` comment omits "tooltip" (`.large` has none). F-3 unverified in the app. Minors read by the orchestrator: UX-2 (F-5), MNT-3 + UX-3 (F-6), MNT-4 + CON-1 (F-7), MNT-6, MNT-7 (F-8) **Resolved**. MNT-5 is covered by MNT's re-review of F-1.

## Round 2 — review
Reviewers: CON (Sonnet, for UX-O1), MNT, UX. Skipped: SEC+PRF (no Fix-now blocker or major). Scope: `564a389c...c4bef1cf`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Round 1 | Resolved | One section, `GalleryTileLook` holds the differences; `git grep ConclusionDetailGallerySection -- macos` empty |
| MNT-2 | MNT | major | Round 1 | Resolved | Keys gone; one stale line in S13-15's plan fixed by the orchestrator |
| MNT-5 | MNT | minor | Round 1 | Resolved | Types moved to `ConclusionGallerySection.swift` |
| UX-O1 | CON | major | Round 1 | Resolved | `CachedThumbnail.fallbackColors`; no raster tile takes the tint |
| UX-1 | UX | major | Focus across 1600 | **Not resolved → To user** | F-3 covers a tile that stays realized (AC-4.2's tile 12). Off the realized set (40 likenesses, tile 30, Wide to Regular): the focus write can precede the animated scroll, a drop after the width update is not restored, and focus on Use as portrait is not tracked at all. Code reading, medium confidence; needs § 10 step 6. Not a blocker, so the cap allows no Fix plan 2 without the user |

Outside-lens note, MNT: F-3 runs on both looks, so a focused Event or Place tile also scrolls and refocuses on a width change. Logged: no visible change unless a tile is focused and the pane crosses the line; those pages are not `fillsWidth`.

No new blockers from CON, MNT or UX. SEC+PRF not re-run.

### Fix plan 2
The user answered UX-1: fix it now (2026-11-03), a fix plan outside the cap for a major. No third review round: one verifier on UX-1 after the gate.

**F-9 — A tile's keyboard stop survives leaving the lazy grid** (from UX-1)
Files: `macos/App/Features/Conclusions/ConclusionGallerySection.swift`, the plan's Landed note.
Failing test first: none, view focus; § 10 step 6 is the check by hand.
Do:
- A private `TileFocus: Hashable` with `.image(String)` and `.useAs(String)` and an `id`. `@FocusState` and `lastFocused` hold it. Bind `.focused` on the image button and on Use as portrait, as the Timeline binds every stop in its rows; Return keeps to `.image`.
- `@State pendingFocus: TileFocus?`. `keepFocus` sets it from `focused ?? lastFocused` and requests the scroll. Pass `animated: false` to `.pvScrollsRowIntoView`: this restores a position, it reveals nothing new.
- One `applyPending()` writes the stop, mapped to `.image` when Use as portrait is not there to take it (the tile is now the lead, or `lead.isSaving`). Call it from the tile's `.onAppear` when `pendingFocus?.id == tile.id`, and after the scroll's turn for a tile already realized.
- In `.onChange(of: focused)`: a stop other than the pending one, or the pending one landing, clears `pendingFocus`; nil while pending is set means the tile dropped after the width update, so request the scroll and `applyPending()` again.
- `pendingFocus` also ends a bounded moment after `keepFocus` (one short `Task.sleep`), so a tile that never dropped does not take focus back after a later click elsewhere. Replace `Task.yield()` timing that the steps above cover.
- F-5's focus return after Use as portrait stays `.image(tile.id)`.
If `PVButton` cannot take `.focused`, stop and report rather than wrapping it.
Check: `.focused` appears on both stops; `animated: false` on the scroll; scripts/check-pr.sh PASS. The Landed note still says F-3 and F-9 are unverified in the app.

## Round 2 — fix
Head `ab673c1a`. Gate PASS (orchestrator re-ran). F-9 done, unverified in the app.

Verifier on UX-1 (one finding, not a review round): Use as portrait stop Refuted, no loop. Partly stands: `focused = pending` is a bare write, and `FocusClaim` records why that is unreliable for a tile inserted by the same commit; any visible non-nil clears `pendingFocus`, so a write with no taker can end the retry; `lastFocused` survives the expiry and can pull focus on a later crossing. A click within the 400ms window also takes focus back: Logged, bounded by the window. Beyond the finding: the `.adaptive` grid reflows away from the 1600 line too, Logged, unconfirmed and outside AC-4.2.

### Fix plan 2, addendum
Inside the user's "fix UX-1 now"; the verifier's first and third items are plain from the code. No further verification: § 10 step 6 decides.

**F-10 — The restore claims focus the way the app does, and leaves nothing behind** (from UX-1, verifier)
Files: `macos/App/Features/Conclusions/ConclusionGallerySection.swift`.
Failing test first: none, view focus.
Do: `applyPending()` goes through `FocusClaim.claim($focused, stop)` instead of a bare write. In `.onChange(of: focused)` a non-nil stop clears `pendingFocus` only when it differs from the pending stop; the pending stop itself stays until the expiry, so the tile's `.onAppear` can still claim it. At the expiry, also clear `lastFocused` when `focused` is nil. Fix the comment that says a landing ends it.
Check: `grep -n 'focused = ' ConclusionGallerySection.swift` prints only the F-5 return; scripts/check-pr.sh PASS.

Head `79a7c522`. Gate PASS (orchestrator re-ran). F-10 done. UX-1: fixed in code; whether a tile off the realized set keeps the keyboard is unverified until § 10 step 6. The implementer saw one failure of `PVWidthClassTests.aFilledHostMeasuresTheSpaceNotTheChild()` that passed unchanged on the next run; the test covers nothing here. Logged as a flake.

## Result
Gate PASS · CI pending · open: none (UX-1 fixed, unverified in the app; § 10 step 6) · follow-ups: SEC-1, the `.adaptive` reflow away from 1600, the `PVWidthClassTests` flake, the 400ms click window · to user: UX-1 answered (fix now, 2026-11-03) · CON miss: UX-O1
