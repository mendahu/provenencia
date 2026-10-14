# Review — S10-20 Open a file and box-select cards

**Verdict:** Fits with changes — Ready with changes (applied)
**Rounds:** 2. **Reviewers:** maintainability (claude-fable-5-1-thinking-high), executability (claude-sonnet-5-5-high)

## To user

None.

## Writer's own calls

- A shared edge is not a selection. `CGRect.intersects` returns false for it, and T-4 locks that.
- Select stays a stored `armedSelect` beside `armedKind` and `armedConnect`. `inputMode` stays the exclusive enum.
- The graph menu row stays a snowflake. The composer row is Out.
- `openableRelPath` and `coverArtifactID` are new helpers. The Source page and the composer keep their looser predicates.
- Cover reuses `L10n.Sources.coverBadge` as micro type. It is not a `PVBadge`.
- Open stays a direct `ProjectFiles.openObject` call. AC-1.3 and AC-1.5 are manual.

## Round 1

### M-1 · major · A third sibling armed flag where one `armedTool` enum is due
**Plan:** D-13 added `armedSelect` beside `armedKind` and `armedConnect`.
**Codebase:** `EvidenceGraphView` `onChange(of: model.armedKind)` and `onChange(of: model.armedConnect)` push and pop the crosshair. `inputMode` is already the exclusive enum.
**Better shape:** one stored `armedTool` enum, with computed accessors.
**Answer:** Declined — those two `onChange`s subscribe to the stored properties. Replacing them with computed accessors rewrites the place and connect cursor path in this PR. D-13 states the rejection. `disarmCiteTools` is the one new function, and it exists so `!canCite` does not turn Select off.

### M-2 · major · The Select button is the third copy of the palette button
**Plan:** D-14 said Select copies `connectButton`.
**Codebase:** `EvidenceGraphPalette.toolButton` and `connectButton` already share the armed chrome. `toolAccessibilityLabel(for:)` and `connectToolAccessibilityLabel` share one sentence.
**Better shape:** one private `paletteTool` and one `toolAccessibilityLabel(name:armed:)`.
**Answer:** Fixed — D-14 and STEP-5. Person, Event, Place, Connect, and Select go through `paletteTool`. The two label functions call `toolAccessibilityLabel(name:armed:)`.

### M-3 · major · The GraphCanvas README is made false and STEP-6 does not touch it
**Plan:** STEP-6 listed only the spike docs.
**Codebase:** `macos/App/Features/GraphCanvas/README.md` describes pan, select, and `draggingItemID` only.
**Better shape:** update that README with the pointer changes.
**Answer:** Fixed — STEP-3 updates the marquee rows. STEP-4 updates the peer-drag rows.

### M-4 · major · Marquee hit-testing rebuilds subject frames the pointer already holds
**Plan:** D-10 rebuilt `edgeFrame`s at release.
**Codebase:** `publishHitTargets` already stores those rects on `pointer.hitTargets`.
**Better shape:** `idsIntersecting` over `[GraphCanvasHitTarget]`, called by the pointer.
**Answer:** Fixed — D-10 and §6 Signatures. `onMarqueeEnded` delivers the id set. `replacePrimarySelection` still drops bridges.

### M-5 · major · The file menu row is a second copy of the composer's artifact row
**Plan:** D-1 rejected extracting `artifactMenuRow`.
**Codebase:** `CitationComposerEvidenceColumn.artifactMenuRow` has no Cover mark, titles fall back to the filename, and the thumb is the live viewer.
**Better shape:** a shared `ArtifactMenuRow`.
**Answer:** Declined — §3 Out says the composer row stays. The graph row's title font, Cover mark, and meta line are not that row. A parameter bag for two different callers is the extraction D-1 already rejected. Next adds names a third caller as the time to extract.

### M-6 · major · Fourth copy of the cover rule, third variant of "has an openable file"
**Plan:** D-1 spelled both predicates inline.
**Codebase:** `isCover`, `defaultPendingArtifactID`, and `SourcePageIdentityHeader` each spell the cover test. `open` and `hasFile` are different file tests.
**Better shape:** `coverArtifactID` and `openableRelPath`, and migrate the old sites.
**Answer:** Fixed for the helpers — D-1 and §6 Signatures. Declined for the migrations — `open` does not check `fileID` or trim, and `hasFile` does not check `relPath`. Unifying them would change the Source page. Next adds records the later adoption.

### M-7 · minor · Half of `EvidenceGraphFiles` belongs elsewhere
**Plan:** `missingFileMessage` lived on `EvidenceGraphFiles`. `kindToken` lived there too.
**Codebase:** `.cursor/rules/macos-l10n.mdc` puts argument copy on an `L10n` format function. `ArtifactViewerKind.classify` and `PVFileTypeGlyph.fromMediaType` return a viewer kind and a symbol.
**Better shape:** the message moves to `L10n`. `kindToken` moves beside the classifier.
**Answer:** Fixed for the message — `L10n.EvidenceGraph.fileNotFoundMessage(name:filename:)` in D-6. Declined for `kindToken` — those two functions do not return a short format label, and the composer keeps "Image". D-4 names the rejection.

### M-8 · minor · Drag-start grows two callbacks and two "dragging" properties
**Plan:** `onDragBegan` plus `dragPeers`, and `draggingItemIDs` beside `draggingItemID`.
**Codebase:** `draggingItemID` is what `EvidenceGraphDragCanvasSize` and the README read.
**Better shape:** one callback that returns the moving ids. Anchor stays `draggingItemID`. Peers are a separate set.
**Answer:** Fixed — D-11 and §6 Signatures. `onDragBegan: ((String) -> Set<String>)?`. `draggingPeerIDs` is the rest. `draggingItemID` stays the anchor.

### M-9 · minor · `commitGroupDrag` copies `commitDrag`'s snap instead of sharing it
**Plan:** group persist copied the snap inline.
**Codebase:** `commitDrag` is origin, dropped point, `gridCell`, generation, patch, one task.
**Better shape:** one private snap helper.
**Answer:** Fixed — `snapAndPatch` in D-11 and §6 Signatures. Both commits call it.

### M-10 · minor · Near-copies of existing strings and of the Cover mark
**Plan:** a new "Cover" key and micro type. A new Open accessibility string.
**Codebase:** `L10n.Sources.coverBadge` is "Cover". `openArtifactFile` is "Open %@ in its default app". The Source page Cover is a `PVBadge`.
**Better shape:** reuse both strings and the badge.
**Answer:** Fixed for the word — D-4 uses `L10n.Sources.coverBadge`. Declined for the badge and for `openArtifactFile` — the board's row is micro type beside the title, and the Open label is "in its app", which is not "in its default app".

### X-1 · blocker · T-4 and D-10 claim `CGRect.intersects` includes a shared edge, and it does not
**Plan:** D-10 and T-4 said a shared edge is included.
**Codebase:** `CGRect.intersects` returns false for a zero-area overlap.
**Better shape:** match the function, and rewrite the test and the risk row.
**Answer:** Fixed — D-10, T-4, and the risk row. A shared edge is not a hit.

### X-2 · blocker · The group-move API the tests depend on is not defined
**Plan:** D-11 and D-12 named behaviors and not signatures.
**Codebase:** `KeyEffect` has one move. `Gesture.draggingCard` stores one id.
**Better shape:** a signatures block.
**Answer:** Fixed — §6 Signatures. `SubjectMove`, `beginGroupDrag`, `groupDragIDs`, `onDragBegan`, `draggingPeerIDs`, `commitGroupDrag`, and `snapAndPatch` are named there. T-8, T-9, T-11, and T-12 call those names.

### X-3 · major · `.selecting` pointer behavior beyond marquee and card drag is undecided
**Plan:** D-8 did not say what card click, actions, and hover do in `.selecting`.
**Codebase:** `mouseDown`, `mouseUp`, and `mouseMoved` switch exhaustively. `mouseDragged` guards on `.idle`.
**Better shape:** `.selecting` matches `.idle` for cards. Empty-canvas drag marquees.
**Answer:** Fixed — D-8 and STEP-3. Every `switch mode` and the idle drag guards gain `.selecting` in that step.

### X-4 · major · Selection model edges are open
**Plan:** `selectedSubjectID` became derived, with no setter rule, no `replacePrimarySelection` signature, and no multi-select Escape rule.
**Codebase:** `prepare` and `activateSubject` assign `selectedSubjectID`. Escape clears when that id is non-nil.
**Better shape:** state the setter, the signature, and Escape.
**Answer:** Fixed — D-7. Get-only `selectedSubjectID`. `replacePrimarySelection(_ ids: Set<String>)`. Escape with Select disarmed clears a non-empty set. `disarm()` also clears `armedSelect`, including the callers that already call it. `selectSubject(id: nil)` does not.

### X-5 · major · STEP-4 will not compile as ordered
**Plan:** STEP-4 added `EvidenceCanvasInputMode.selecting` and left the exhaustive switches to STEP-5. The file was not in the step's file list.
**Codebase:** `syncPointerMode` and the `onHover` switch are exhaustive.
**Better shape:** update those switches in the step that adds the case.
**Answer:** Fixed — STEP-4 lists `EvidenceCanvasInputMode.swift` and updates `syncPointerMode` and `onHover` in that step.

### X-6 · major · Header select, Open, and callout details are open
**Plan:** D-1 through D-6 named behaviors and not types.
**Codebase:** `PVSelect` has two inits. `PVCallout` shows beside `pendingImpact == nil`.
**Better shape:** a signatures table and L10n property names.
**Answer:** Fixed — §6 Signatures, D-3 (both inits, empty filename passes nil help), D-4 (fonts, thumbnail path, mono title), D-6 (`selectedFileArtifactID` `didSet`, `syncFileSelection`, callout gate). Counts still loading show "No citations" (D-2).

### X-7 · minor · D-1 says "same openable-file rule" as the Source page, but it is not
**Answer:** Fixed — D-1 now says the graph rule is stricter, and the Source page predicates stay.

### X-8 · minor · Tests T-1, T-3, and T-9 cannot be written from their rows alone
**Answer:** Fixed — T-1 names the four artifacts. T-3 names its own hit target and says not to use `controller()`. T-4 names the rects. T-9 names `SubjectMove` fields.

### X-9 · major · The riskiest behavior has no automated test
**Plan:** group persist and peer offsets were manual only.
**Better shape:** one `commitGroupDrag` test and one pointer peer test. Keep Open manual.
**Answer:** Fixed — T-11 and T-12. AC-1.3 and AC-1.5 stay manual, named in the Definition of done. No `fileOpener` seam.

### X-10 · minor · PlaceRegistry key order is not stated
**Answer:** Fixed — D-2 appends `.sourceWorkspace` then `.citationCounts` after `.connectRules`.

### X-11 · minor · Armed-Select side effects are not covered
**Answer:** Fixed — D-13. No crosshair. Spoken label stays the graph title. `emptyOverlay` stays `inputMode == .idle`.

### X-12 · nit · A Current state row and the callout slot are slightly off
**Answer:** Fixed — the Precedent names `PVThumbnail` on the composer row and `CachedThumbnail` on the graph. D-6 gates the file callout on `pendingImpact == nil`.

## Round 2

Round-1 answers held. No new blocker. The majors and the cheap minors below are applied. Round 2 is the cap.

### M-12 · major · File rows and file selection are split between view and model
**Answer:** Fixed — D-6. The model owns `fileRows`, `resolvedFileArtifactID`, `selectFile`, and `openSelectedFile`. There is no `syncFileSelection`.

### M-13 · minor · Third direct `ByteCountFormatter` call
**Answer:** Fixed — D-4 calls `IngestMediaPolicy.formatByteSize`.

### M-14 · minor · Two accessibility-label wrappers that only forward
**Answer:** Fixed — D-14. `paletteTool` calls `toolAccessibilityLabel(name:armed:)` and the two wrappers are deleted.

### M-15 · minor · Arrow group move persists through a view loop
**Answer:** Fixed — D-12. `moveSubjects` assigns one `lastPositionWrite`. The view calls it once.

### M-16 · minor · D-13's "Why" teaches a false fact about `onChange`
**Answer:** Fixed — D-13 now says this PR does not rewrite the place and connect arm path. Next adds names the `armedTool` fold for a later tool.

### M-17 · minor · Menu-row `CachedThumbnail` drops the MIME-glyph fallback
**Answer:** Fixed — D-4 passes `mediaType` and `originalFilename`.

### X-13 · major · `snapAndPatch` cannot hand each card's generation to the persist task
**Answer:** Fixed — §6 Signatures. `snapAndPatch` returns `(gridX, gridY, generation: Int)`. The task captures that generation.

### X-14 · major · The view cannot reach the cover or the file rows
**Answer:** Fixed — same change as M-12. `sourceRow` stays private. `openSelectedFile()` takes no rows.

### X-15 · minor · An empty `originalFilename` breaks three strings
**Answer:** Fixed — `displayFilename` on the row. Help, meta, the accessibility label, and the missing-file message use it.

### X-16 · minor · Arrow `moves` order is undefined
**Answer:** Fixed — D-12. `moves` follow snapshot subject order. `SubjectMove` is `Equatable`. The view calls `moveSubjects` once.

### X-17 · minor · Selection edge cases around activation, deletion, and Escape order
**Answer:** Fixed — D-7. `replacePrimarySelection` clears activation when the id leaves the set. Delete removes the id and clears activation when it matches. Escape handles cite tools, then Select, then activation, then a non-empty set.

### X-18 · minor · `paletteTool` leaves out Select's help and the icon's armed colour
**Answer:** Fixed — D-14. `help` is a parameter. The icon closure receives `armed`.

### X-19 · minor · Pointer details and test placement
**Answer:** Fixed — D-8. `pendingClick(start:targetID:shiftDown:)`. Mouse up clears peer offsets. T-3, T-5, T-6, and T-12 live on `GraphCanvasPointerPressTests`.

### X-20 · minor · The stricter file predicate has no test
**Answer:** Fixed — T-1 drops an empty `fileID` and a blank `relPath`, and names id `d`. T-13 covers `resolvedFileArtifactID` and `selectFile`.

### X-21 · minor · STEP-2 is large, the callout hides banners, and two doc gaps
**Answer:** Fixed for the callout, the Select contract line, and manual step 2. Declined for splitting STEP-2 — it is one header behavior, and the step already names every file.

### X-22 · nit · Two Current state rows are loose
**Answer:** Fixed — `commitDrag` takes the id as a parameter. The palette row names Person, Event, Place, and Connect. The brief and map wording travel with this plan.
