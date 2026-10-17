# Review — S10-34 Composer depiction row

**Verdict:** round 1: Fits with changes — Not ready; round 2: Fits with changes — Ready with changes → all applied. Planned 2026-10-17.
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None open. The four product calls were asked before writing and are recorded in D-8, D-9, D-10 and the **Board** line: redraw on edit, the PDF page rule, Duplicate copies the region, and the board link comes from the user.

## Writer's own calls
- CDR-6 / D-11: after a depiction saves, the dialog keeps the property and clears the person (a group photo is a run of different people). Every other value type keeps the person, as today.
- D-11: a subject change keeps the property and its value when the new subject's type also offers it, instead of clearing on any type change. This applies to every value type.
- D-12: choosing another Citation in the dialog clears a drawn region.
- D-1, D-5, D-17: this PR deletes the dead property-menu cache and `CitationObservationRows.applySubject`, moves the crop functions out of `OCRImage`, and moves the file fixtures out of `CitationAutoTranscribeTests`.
- Copy in §6 Strings is the brief's until the board is agreed. The saved-line thumbnail is 32pt and the preview 96pt (S10-D15 CDR-4, CDR-7).

## Round 1 — maintainability

### M-1 · blocker · The image crop becomes a second copy of Auto Transcribe's crop
**Answer:** Fixed — D-5: `ArtifactRegionCrop` takes `cgImage(from:)` and both `crop` functions from `OCRImage`, adds `boundingBox(of:region:)`, and `runAutoTranscribe` calls it. STEP-2 lists the moves and their tests.

### M-2 · blocker · Four loose depiction fields and a second page-stamp rule beside `CitationLocatorDraft`
**Answer:** Fixed — D-2: `DepictionValueDraft` wraps `CitationLocatorDraft`. D-4 encodes through it, and D-7 commits through `locator.setRegion(_:capabilities:viewerPage:)`.

### M-3 · blocker · `likenessPreview` is stored state kept in sync by hand
**Answer:** Fixed — D-14: `depictionPreview` is computed over `currentDepictionDraft`, with a one-entry memo.

### M-4 · major · Generic machinery named after the `likeness` key
**Answer:** Fixed — every new symbol, file, and L10n key says `depiction`. The plan file and branch are renamed to `composer-depiction-row`.

### M-5 · major · The drawing state should take `over: ComposerBase`
**Answer:** Fixed — D-7 adds `drawingDepiction(over: ComposerBase)`, which reuses the `customTerm`/`newSubject` arms.

### M-6 · major · `allowsDepiction: Bool` with a hard-coded `true` caller
**Answer:** Fixed — D-1 adds `propertyOptions(forSubjectID:capabilities:)` with one `isOffered(valueType:capabilities:)` rule. The `true` caller, `applySubject`, has no callers and is deleted.

### M-7 · major · D-1 edits a dead path instead of deleting it
**Answer:** Fixed — D-1 deletes the chain; the Out line is gone; §6 Deletes lists it.

### M-8 · major · Two value clearers already drift
**Answer:** Fixed — D-2: `ObservationRowValues.clearValue()` is the one clearer.

### M-9 · minor · `depictionWriteError` does not follow the `<call>Error` shape
**Answer:** Fixed — D-16: `addObservationsToCitationError`, `updateObservationError`.

### M-10 · minor · Fixture file named for one feature
**Answer:** Fixed — D-17: `ArtifactFileFixtures` without `waitUntil`, plus one line in `add-swift-test`. Folding the other copies is an Out line (dogfood log).

### M-11 · minor · Render/write/call/delete appears at two call sites
**Answer:** Fixed — D-6: `withDepictionInput`, used by D-5's save and D-9's edit.

### M-12 · nit · The 2048 cap repeats a Go constant
**Answer:** Fixed — D-5: the doc comment cites `stageMaxEdge`.

## Round 1 — executability

### X-1 · minor · `customTerm`/`newSubject` use `over:`, not `returnTo:`
**Answer:** Fixed — §4 row and D-7 (see M-5).

### X-2 · blocker · The drawing state machine is undecided
**Answer:** Fixed — D-7 covers the enum shape, the `suspendedDialog`/`restorePresented` arms, leave while drawing (T-20), the actions disabled while drawing (AC-1.10), the tool being disarmed on commit, and the transcription guard.

### X-3 · blocker · D-7 contradicts CDR-8 about what the viewer shows during edit
**Answer:** Fixed — D-7: `viewerRegion` is the depiction's region whenever a depiction draft is open, not only while drawing.

### X-4 · blocker · L10n keys are used before STEP-8 adds them
**Answer:** Fixed — STEP-1 adds every key.

### X-5 · blocker · The focus function cannot tell a depiction apart
**Answer:** Fixed — D-11: `initialPropertyFocus(subjectID:propertyID:isDepiction:)` with all four results, and §4/STEP-1 list every call site.

### X-6 · major · STEP-1 omits `propertyValuesComplete`, `fields(from:)`, `valueFields()`
**Answer:** Fixed — D-2 and STEP-1 name them.

### X-7 · major · The kept-property subject rule leaves stale values
**Answer:** Fixed — D-11: a kept property keeps its value; §11 risk row states the change for other value types.

### X-8 · major · Which draft `commitViewerRegion` writes, and when the preview clears
**Answer:** Fixed — D-7 `currentDepictionDraft` covers both the authoring draft and the edit session. The preview is derived (D-14), so it needs no clearing.

### X-9 · major · Render source can differ from the draft's Artifact
**Answer:** Fixed — D-12: a Citation change clears the region; an artifact switch already resets `authoring`. Drawing disables both (D-7).

### X-10 · major · Loaded depiction whose region Go hid
**Answer:** Fixed — D-2 `regionHidden`, D-15, CDR-14, AC-2.4, T-21. The line keeps "—" when the image path is empty (D-13).

### X-11 · major · `performCommit` add branch; `saveEdit` with a new Citation
**Answer:** Fixed — D-9: both branches carry the input. An edit exists only for a saved Observation, so its Citation exists. `withDepictionInput` wraps the commit.

### X-12 · major · Fixtures the tests need are missing
**Answer:** Fixed — D-17 names `writePNG(… topHalfRed:)`, `writeMarkedPDF(rotation:)`, the `prop-likeness` fixture, and the private `makeStore`/`makeModel`. §8 names the viewer load path and the `CoreInvokeError.coded` shape.

### X-13 · major · Accessibility labels contradict each other
**Answer:** Fixed — D-13: the line label goes on the row, the thumb label on the thumbnail, and the person comes from `vocabulary.subjectsByID`.

### X-14 · major · FakeStore format and edit-path details
**Answer:** Fixed — D-16: appended tokens, a `depictionSequence` counter, lookup of the stored depiction, and mapping in both builders.

### X-15 · minor · Render math loose
**Answer:** Fixed — D-5: points from the normalized bbox × visual size, a 4096 page cap, crop in `CGImage` pixels, the `CGContext` spec, and the encode error type (D-6).

### X-16 · minor · D-8 details
**Answer:** Fixed — D-8 defines an existing Citation as `context.citationID?.isEmpty == false`. A new Citation with no page has no region (`setRegion` stamps a page). D-7 adds the transcription guard.

### X-17 · minor · `PVThumbnail.Content` shape
**Answer:** Fixed — D-14: `PVThumbnail.Content(image: Image(nsImage:), label:)`. D-13 gives the `CachedThumbnail` arguments.

### X-18 · minor · Docs step detail
**Answer:** Fixed — STEP-10 gives the section and the bullet text.

## Round 2 — after the board was agreed (2026-10-17)

Board: Claude Design *Citation Composer*, section "Depiction row", frames D1a–D10 and D5c, checked frame by frame against an export. §6 "Frames" added; D-19 (caption), D-20 (save-error Callout), D-21 (depiction copy) added. Verdicts: maintainability Fits with changes; executability Ready with changes. Every round-1 answer was confirmed applied.

### M-13 · blocker · `regionHidden` is a stored flag that goes stale after a redraw
**Answer:** Fixed — D-2: computed `!valueDepictionID.isEmpty && !hasRegion`; D-15, D-19, and the Save gate read it; T-21 checks it turns false after a commit.

### M-14 · major · D-20 makes two error styles and "Not saved" twice
**Answer:** Fixed — D-20: the bridge Callouts get the same title, `performCommit`'s fallbacks use `observationSaveFailed`, and the `rowStateError` comment is updated.

### M-15 · major · Depiction copy beside the functions that own it
**Answer:** Fixed — D-21: `duplicateOrigin(line:)` owns the note, `unsavedSummary`'s holds-leave arm owns the leave message, and `editSheetSubtitle` owns the subtitle.

### M-16 · minor · "Redrawn" defined twice
**Answer:** Fixed — `isDepictionRedrawn` (D-9) is used by D-9 and D-19.

### X-19 · major · Drawing mode leaves most of the screen undefined
**Answer:** Fixed — D-7 lists every disabled control and the guarded model actions; STEP-4 names the files; T-20 checks the guards.

### X-20 · major · The needs-page Callout does not unblock Draw
**Answer:** Fixed — D-8 and §6 Strings: the copy says to set and save the page. This is a deviation from the board's copy and is listed to the user.

### X-21 · major · D9a copy is not what the app shows
**Answer:** Fixed — D-20 keeps the mapped error string and states that the board's frame is shorter.

### X-22 · major · A 32pt tile does not keep the line height
**Answer:** Fixed — D-13: depiction lines use `space1` vertical padding, about 36pt, as drawn.

### X-23 · major · Render failure and order in the two save paths
**Answer:** Fixed — D-6: render first, a named catch, and the edit-session error path.

### X-24 · major · FakeStore cannot prove T-16
**Answer:** Fixed — D-16 records ` valueDepictionID=<sent id>`; T-16 checks it.

### X-25 · minor · `kind: .validation` does not exist
**Answer:** Fixed — `.user`.

### X-26 · minor · The new focus signature breaks a test
**Answer:** Fixed — D-11: default parameters, `isDepiction(propertyID:)`, and the call-site count corrected in §4.

### X-27 · minor · `base(of:)` needs an arm
**Answer:** Fixed — D-7.

### X-28 · minor · `regionHidden` after a redraw
**Answer:** Fixed — see M-13.

### X-29 · minor · Escape wiring
**Answer:** Fixed — D-7: Cancel carries `.cancelAction` while drawing; drawing closes the find row.

### X-30 · minor · Smaller gaps
**Answer:** Fixed — D-14 gives the NSImage and the 96pt size. D-1 sets `depictionCapabilities` from the selected Artifact's media type, so loading does not change the list. D-13 gives the label sources and the nil fallback. The D8c fixture has no region. D4 was already handled.

**Status:** no open blocker after round 2; the plan is `Planned`.
