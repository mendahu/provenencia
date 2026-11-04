# Ship log — S13-15 Likeness viewer

**Plan:** [`S13-15-likeness-viewer.md`](S13-15-likeness-viewer.md) · **Branch:** `s13-15-likeness-viewer` · **PR:** #703
**State:** stopped — UX-5 and MNT-2 open after round 2; To user

## Round 0 — implement
Head `c440b2cb`. Gate PASS (orchestrator re-ran). Report: deviations: `ConclusionPictureImage.sized` and `.decoded` hold the file's pixel size and the view computes the frame on each draw; every close clears the Toast and `toastInViewer`; focus restore skipped for an empty gallery; plan § 4 row moved to the new decode path because lint failed. Manual check, frames, keys, VoiceOver not run (no GUI).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. SEC+PRF routed by `macos/App/Features/**` models, an image decode of a catalog file, and `Platform/ImageHeaderDecode.swift`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | Use as portrait drops the keyboard; arrows stop | Fix now → F-1 | Verified: `imageFocused` is written only in `.onAppear`; the tile restores focus for the same case |
| UX-2 | UX | major | Stepping to an end with a button drops the keyboard | Fix now → F-1 | Verified: nothing reclaims focus when a focused control is disabled or removed. Runtime result needs the app |
| SEC-1 + MNT-2 | SEC, MNT | major | No byte bound; a second admission rule beside `ThumbnailDecodePolicy.accepts` | Fix now → F-2 | Verified: `admit` checks no length; `accepts` and `admit` are two sniff, header and fit gates. Same-pass: this PR added `admit` |
| SEC-2 + PRF-1 | SEC, PRF | major | Decodes in flight have no bound; default priority | Fix now → F-3 | Verified: `Task.detached` per step, no slot. Supersedes D-5's "steps more than 120 ms apart can overlap" and the matching Scale line, which recorded the gap: the plan's own Check is one decode shown, and the kit's rule is one in flight |
| MNT-1 + UX-5 | MNT, UX | major | `toastInViewer` mirrors `viewing`; opening or emptying speaks a Toast twice | Fix now → F-4 | Verified: the writers miss both paths, so AC-2.1's "spoken once" fails. Moves D-10's and D-11's wiring to the viewer's own mount; D-10's rule stands |
| MNT-3 | MNT | major | Third "drop empties, then spokenJoin" | Fix now → F-5 | Verified: three call sites. Same-pass cleanup |
| CON-2 | CON | minor | One frame of the old image under the new counter | Fix now → F-6 | Verified by reading: `image` resets inside the task, after the render |
| CON-3 | CON | minor | A blank sheet for one draw when the shown likeness goes | Fix now → F-6 | Verified: the closure has no `else` |
| UX-3, UX-4 | UX | minor | Arrows do nothing from close or the footer; an arrow at an end may beep | Fix now → F-7 | Verified by reading. Supersedes D-8's "on the panel body" and `.ignored` |
| UX-6, UX-7, UX-8 | UX | minor, nit | A missing image spoken as present; the portrait not spoken on a step; the footer button's label | Fix now → F-8 | Verified: `viewerLabel` ignores `leads`, and the image's label ignores `image` |
| MNT-5, MNT-7 | MNT | minor | `citation` stored beside `outcome`; hard-coded panel padding and 320 | Fix now → F-9 | Verified |
| MNT-6 | MNT | minor | The decode keeps its Artifact name in `Platform/` | Fix now → F-2 | Its callers are this PR's |
| MNT-4 | MNT | minor | `viewerLabel` defaults | Rejected | D-12 put the label on the tile so no `L10n.format` sits outside `L10n` (`macos-l10n.mdc`); `ViewerCopy` carrying a format is what plan review M-2 removed |
| CON-1 | CON | minor | R-32 still says a tile click opens the Citation | Fix now → F-10 | Verified: `person-detail.md:75` |

Nits (logged): `CON-4 · ConclusionGallerySection.swift:98 · rewrap` · `CON-5 · ReconciledValueDisplay.swift:136-139 · doc line mid-block` · `MNT-8 · restoreFocus and GalleryTileView docs · rewrap` · `MNT-9 · ConclusionPictureViewer.swift:85 · overlapping inputs` · `MNT-10 · buttons could use move(to:)` · `MNT-11 · declaredSize's default has no production caller` · `MNT-12 · ConclusionPictureStep Equatable unused` · `MNT-13 · stepColumn names two things` · `MNT-14 · :240 · 256 from ThumbnailCache.maxPixelSize` · `UX-9 · :194 · bare focusable, no kit ring` · `UX-10 · :197 · a gallery of one speaks 1 of 1` · `UX-11 · :240 · .waiting draws above the cap for 120 ms` · `PRF · each likeness runs admit twice`

### Fix plan 1

**F-1 — The viewer keeps the keyboard** (from UX-1, UX-2)
Files: `macos/App/Features/Conclusions/ConclusionPictureViewer.swift`.
Failing test first: none, view focus; manual check 3 is extended by hand.
Do: replace `imageFocused` with `@FocusState focus: ViewerStop?` (`.image`, `.previous`, `.next`) bound on the band and both step buttons. After `lead.choose`, whatever the result, `FocusClaim.claim($focus, .image)`. After a step, and on `.onChange(of: step.count)`, when the focused button is now disabled or gone, claim `.image`.
Check: scripts/check-pr.sh PASS. The Landed note says this is unverified in the app.

**F-2 — One admission for a picture file, with a byte bound** (from SEC-1, MNT-2, MNT-6)
Files: `macos/App/Platform/ArtifactImageDecode.swift` → `macos/App/Platform/BoundedImageDecode.swift`, `macos/App/Platform/ImageHeaderDecode.swift`, `macos/App/Platform/ThumbnailCache.swift`, `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift`, `macos/App/Features/ArtifactViewer/ArtifactViewerContent.swift`, `ConclusionPictureViewer.swift`, `api/ffi/swift_vocabulary_test.go`, `.cursor/rules/macos-client.mdc`, tests.
Failing test first: `declaredSizeRefusesAnOversizedFile` — a sparse file over the limit with a valid PNG header → `.tooLarge` from `maxBytes`.
Do: `ImageDecodeLimits` *(new)*: `typeIdentifiers`, `maxEdge`, `maxBytes: Int?`. `ImageHeaderDecode.admit(_ source:, limits:)` holds the type, header and fit check, and both `ThumbnailDecodePolicy.accepts` and the decode's `admit` call it, on `conforms(to:)`. `BoundedImageDecode.decode(url:limits:)` and `declaredSize(url:limits:)` read the file's size before `CGImageSourceCreateWithURL` when `maxBytes` is set. `ThumbnailDecodePolicy.limits` is the one value the viewer passes, with `maxDecodeBytes = 134217728`, held equal to `raster.MaxDecodeBytes` by a new `TestSwiftVocabularyMatchesGo` row. The Artifact viewer passes `ImageDecodeLimits(typeIdentifiers:)`, no byte bound, as ingest accepts 512 MiB. Update `ThumbnailDecodePolicy`'s doc comment.
Check: `rg 'ArtifactImageDecode' macos api .cursor .agents` prints nothing; the test passes; `go test -tags fts5 ./api/ffi/...` passes; scripts/check-pr.sh PASS.

**F-3 — One decode in flight** (from SEC-2, PRF-1)
Files: `ConclusionPictureViewer.swift`, the plan's Landed note.
Failing test first: none, view task; `ConclusionPictureImage` has no scheduler to test.
Do: hold the running decode as `@State decodeTask: Task<Void, Never>?`. In `load()`, after the sleep and before opening a file, `await decodeTask?.value` and recheck `Task.isCancelled`, so only the latest wanted likeness decodes next. Both detached tasks take `priority: .userInitiated`. The Landed note says this supersedes D-5's overlap sentence and the Scale line.
Check: scripts/check-pr.sh PASS.

**F-4 — The viewer owns its Toast** (from MNT-1, UX-5)
Files: `ConclusionPictureViewer.swift`, `ConclusionGallerySection.swift`, `ConclusionLeadModel.swift`, the plan's Landed note.
Do: `ConclusionLeadModel.viewerOpened()` and `viewerClosed()` *(new)* each set `toastInViewer` and clear `toast`. The viewer calls them in `.onAppear` and `.onDisappear`. The section keeps `.onDisappear { lead.viewerClosed() }`, since a removed sheet's `.onDisappear` is unproven, and its `onChange(of: viewing)` does focus only. Landed note: this moves D-10's and D-11's writer; the rule is unchanged.
Check: `rg 'toastInViewer =' macos/App` prints only `ConclusionLeadModel.swift`; scripts/check-pr.sh PASS.

**F-5 — `spokenJoin` drops empty parts** (from MNT-3)
Files: `L10n.swift`, `PVProvenanceSheetContent.swift`, `EventsListView.swift`, `AccessibilityAnnouncement.swift`, tests.
Failing test first: `spokenJoinDropsEmptyParts` — `["a", "", "b"]` → "a, b"; `["", ""]` → nil.
Do: `spokenJoin(_ parts:)` filters empties and returns nil when none remain. Remove the three filters.
Check: `grep -rn -B2 'spokenJoin(' macos/App | grep 'isEmpty'` prints nothing; scripts/check-pr.sh PASS.

**F-6 — No stale frame, no blank sheet** (from CON-2, CON-3)
Files: `ConclusionPictureViewer.swift`, `ConclusionGallerySection.swift`.
Do: `.sized` and `.decoded` carry the tile id, and `imageContent` treats a mismatch with `step.tile.id` as `.waiting`. The sheet closure falls back to the first tile's step when `shownID` has just left.
Check: scripts/check-pr.sh PASS.

**F-7 — Arrows work anywhere in the viewer** (from UX-3, UX-4)
Files: `ConclusionPictureViewer.swift`, the plan's Landed note.
Do: move both `.onKeyPress` calls to the `PVPanel` expression. `move(to:)` returns `.handled` at the ends. Landed note: this supersedes D-8's two phrases.
Check: scripts/check-pr.sh PASS.

**F-8 — What VoiceOver hears** (from UX-6, UX-7, UX-8)
Files: `ConclusionPictureViewer.swift`, `ConclusionDetailPage.swift`, `PersonDetailContent.swift`, `L10n.swift`, `Localizable.xcstrings`, tests. Skill: `add-localized-string`.
Failing test first: `personGalleryOffersTheViewer` — the lead tile's `viewerLabel` ends ", portrait".
Do: `viewerLabel` also takes the lead flag; `likenessViewerLabelLead(index:count:line:)` = "Likeness %1$lld of %2$lld, from %3$@, portrait". `.refused` adds `.accessibilityValue` from `likenessViewerImageUnavailable` = "Image unavailable". The footer button takes `.accessibilityLabel(step.tile.useAsLabel)`.
Check: the test passes; `python3 scripts/check-localizable-xcstrings.py` passes.

**F-9 — Small facts held once** (from MNT-5, MNT-7)
Files: `ConclusionDetailPage.swift`, `ConclusionPictureViewer.swift`, `macos/App/DesignSystem/Components/Panel/PVPanel.swift`.
Do: `ConclusionDetailPicture.citation` is computed from `outcome`. `PVPanel.bodyInset` names the padding, and the frame reads it. `minBandHeight` 320 and `chromeAllowance` 320 are named, and the comment matches.
Check: scripts/check-pr.sh PASS.

**F-10 — R-32** (from CON-1)
Files: `docs/deployment-plan/spike-13/person-detail.md:75`.
Do: the Person's tile opens the viewer; Event and Place tiles still open the Citation.
Check: `python3 scripts/check-plan-docs.py` reports 0 errors.

## Round 1 — fix
Head `d3ef0f1a`. Gate PASS (orchestrator re-ran); full `check-plan-docs.py` 0 errors. F-1 to F-10 done. The implementer wrote the F-2, F-5 and F-8 tests beside their fixes, not red first, so the three pin behavior but did not fail first. Deviations: `ImageHeaderDecode.Refusal` and `maxDeclaredPixels` moved with `admit`; `allowedTypes` is gone for `limits`; `accepts` has no byte bound, having no file URL; `decodeTask` holds the decode's result; `.refused` carries the tile id too; `PVPanelMetrics.bodyInset`, since a generic type holds no stored static; the viewer label closure takes a fourth `leads`. Two `PVWidthClassTests` failed once and passed unchanged, the flake S13-14 logged. Minors read by the orchestrator: CON-1 (F-10), CON-2, CON-3 (F-6) **Resolved**.

## Round 2 — review
Reviewers: MNT, UX, SEC+PRF. Skipped: CON (no Fix-now blocker or major). Scope: `c440b2cb...d3ef0f1a`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | major | No byte bound | Resolved | The viewer's `declaredSize` and `decode` read the file's size first; `TestSwiftVocabularyMatchesGo` holds the limit |
| SEC-2 + PRF-1 | SEC, PRF | major | Decodes in flight | Resolved | `decodeTask` awaited after the sleep; `.userInitiated` |
| UX-1, UX-2 | UX | major | The keyboard drops | Resolved | `FocusClaim` after `choose`; `reclaimFocusIfStranded()`. Unverified in the app |
| MNT-1 | MNT | major | `toastInViewer` mirrored | Resolved | Only `viewerOpened()` and `viewerClosed()` write it |
| MNT-3 | MNT | major | spokenJoin | Resolved | Three filters gone |
| MNT-5, MNT-6, MNT-7 | MNT | minor | Round 1 | Resolved | Read by round 2's reviewer |
| UX-3, UX-4, UX-6, UX-7, UX-8 | UX | minor, nit | Round 1 | Resolved | Read by round 2's reviewer |
| UX-5 | UX | major | Opening the viewer during a Toast still speaks it twice | **Not resolved → To user** | The viewer's overlay is ungated, and `viewerOpened()` runs in `.onAppear`, after the first draw mounts it. AC-2.1 fails on Use as portrait then Return. Medium confidence |
| MNT-2 | MNT | major | The byte limit applies on one of two URL paths | **Not resolved → To user** | `ThumbnailCache.swift:79` opens by URL and decodes with no byte check, though it shares `limits`; the doc comment and the F-2 deviation are wrong that it has no URL. A file over 128 MiB shows on the tile and refuses in the viewer |
| UX-13 (with MNT's outside note) | UX, CON | major | `viewerClosed()` clears the Toast when the Gallery section leaves, with no viewer open | **To user** | Re-rated from a nit: `ConclusionGallerySection.swift:100` runs on every section switch and on Event and Place, so a tile's or the header menu's Toast, a danger Toast included, vanishes. Regression from F-4. Not a blocker, so the cap allows no Fix plan 2 without the user |

No new blockers. CON not re-run.

Open minors and nits, no fix plan without the user: `UX-12 · :329 · focus nil also true on close, Use as portrait, a record link, so a step moves the keyboard to the image` · `UX-14 · :230 · "Image unavailable" reaches VoiceOver only when the element is read` · `UX-15 · ConclusionGallerySection.swift:70-73 · the fallback can announce "Likeness 1 of 6" for one draw` · `UX-16 · :230 · ", portrait" changes the label after Use as portrait` · `EventsListView.swift:69 · a filter now redundant`

## Result
Gate PASS (`d3ef0f1a`) · CI not run (draft) · open: UX-5, MNT-2, UX-13 · follow-ups: — · to user: UX-5, MNT-2, UX-13 · CON miss: UX-13
