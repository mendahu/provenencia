# S10-20 — Open a file and box-select cards

**Status:** Landed in [#396](https://github.com/mendahu/provenencia/pull/396) (2026-10-14)
**Spike / slice:** Spike 10 · Slice 5 — Artifacts and multi-select
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-20--open-file-and-box-select)
**Brief:** [S10-D6](../design/S10-D6-evidence-graph.md) (EG-1…EG-7), board agreed 2026-10-14
**Board:** The agreed frames are the S10-D6 pass in the Claude Design export *Evidence graph canvas.dc.html*. §6 macOS describes every frame this PR ships. The implementer does not need the export.
**Depends on:** S10-01 (landed [#361](https://github.com/mendahu/provenencia/pull/361)), S10-03 (landed [#363](https://github.com/mendahu/provenencia/pull/363)), S10-D6
**Blocks:** —
**Branch:** `s10-20-open-file-box-select`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md)
**Rules that bind:** `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/workspace-navigation.mdc`
**Context:** working rules "Box-select is a marquee of subject cards" and "A drop on the Source page opens Add Artifact" (the graph header sentence), [S10-D6](../design/S10-D6-evidence-graph.md) EG-1…EG-7, dogfood bar items 2 and 3 in [`../deployment-plan.md`](../deployment-plan.md)
**Planned:** 2026-10-14 against `main` @ `48d4b41e`
**Review:** [`S10-20-open-file-box-select.review.md`](S10-20-open-file-box-select.review.md) — Fits with changes, Ready with changes (applied); 2 rounds

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

A researcher on the Evidence graph can open one of this Source's files in the system app, and can box person, event, and place cards and move that group together. The header gains a file select and Open. The palette gains Select. Shift-drag on empty canvas draws the same rectangle. This is dogfood bar items 2 and 3, and slice 5's graph half of the Check.

**Why now.** S10-01 and S10-03 have landed, and the S10-D6 board is agreed. Unpromote is drawn on that board and stays S10-44.

**What stays.** Promote all, the jump to the Source page, Place, Connect, the no-artifact callout, the cards, the edges, two-finger pan, and pinch zoom. Empty-canvas drag still pans when Select is disarmed and Shift is up. A drag that starts on a card still moves that card. Card delete stays the impact confirm. The full card stays at every zoom. The Source page Artifacts section is unchanged.

## 2. User stories

**US-1.** As a researcher, when a Source has files, I pick one in the graph header and Open it in its own app, so I can read the scan without leaving for the Source page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Artifacts in order: cover "Page 6" with a file, "Page 7" with a file, fileless "Loose note", unlabeled file `ART-7Q2KD` | The select lists Page 6, Page 7, and ART-7Q2KD, in that order. Loose note is absent. The initial selection is Page 6. If the cover has no file, the initial selection is the first row that has a file. |
| AC-1.2 | The select is showing Page 6, and the researcher chooses Page 7 | The closed label becomes Page 7. No file opens. |
| AC-1.3 | Page 7 is selected and the researcher presses Open | `ProjectFiles.openObject` runs with that artifact's `relPath`. The graph's card selection and camera stay as they were. |
| AC-1.4 | The Source has no artifact whose file has a non-empty `relPath` | The select and Open are absent. Promote all and Jump to Source page stay. |
| AC-1.5 | Open is pressed and `ProjectFiles.openObject` returns false for "Page 7" / `1881-peel-n-p7.jpg` | A danger callout reads title "File not found" and message "Page 7 points at 1881-peel-n-p7.jpg, which is no longer where Provenencia stored it." The select still shows Page 7. Choosing another file clears the callout. |
| AC-1.6 | A label longer than the closed control, two artifacts that share the label "Page 6" | The closed control is 208pt wide, shows the label (or the ref when the label is empty), and truncates the tail. Its help is that row's filename. The menu row shows the filename, so the two "Page 6" rows are distinct. |

**US-2.** As a researcher, when I draw a rectangle on empty canvas, the person, event, and place cards inside it are selected, so I can grab a household without selecting the relationship card.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Select is disarmed. Shift is down. The pointer goes down on empty canvas and drags past 4pt | A rectangle is drawn. The canvas does not pan. |
| AC-2.2 | The rectangle covers three people and one relationship, then the pointer goes up | The three people show the selection ring. The relationship does not. The rectangle is gone. A rectangle that hits no person, event, or place clears the selection. |
| AC-2.3 | Select is armed. Shift is up. The pointer drags on empty canvas | The same rectangle is drawn. The canvas does not pan. |
| AC-2.4 | Select is disarmed. Shift is up. The pointer drags on empty canvas | The canvas pans. No rectangle. The selection does not change. |
| AC-2.5 | Shift is down, or Select is armed, and the drag starts on a card | No rectangle. That drag is a card drag. |
| AC-2.6 | The Source has no artifact (`canCite` is false) | Place and Connect are dimmed and disabled. Select is not dimmed and still arms. Its help is "Hold Shift to select". It sits after Connect, past a hairline, with `PVSymbol.regionRectangle`. Arming Select disarms Place and Connect. Arming Place or Connect disarms Select. Escape, or pressing Select again, disarms Select and leaves the selection. |

**US-3.** As a researcher, when several subject cards are selected, I drag one or press an arrow, and the group moves together, so I can shift a household without restacking each card.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Three people are selected and a relationship sits between them. The researcher drags one selected person by the same document delta | All three people snap and are stored by that delta, each from its own cell, the way one card is stored today. The relationship stays on its cell. Edges meet the people in their new cells. |
| AC-3.2 | Those three are selected, and the researcher drags a fourth person who is not selected | The selection becomes that fourth person only, and only that card moves. |
| AC-3.3 | Three people are selected. The researcher presses the right arrow, including while Select is armed | Each selected person moves one cell right. A relationship is not moved. Return and Space do nothing while Select is armed. |
| AC-3.4 | Three people are selected, and the researcher clicks one person | Only that person is selected. A click on empty canvas clears the selection and does not disarm Select. |

## 3. Scope

| In | Covered by |
| --- | --- |
| EG-1, EG-3. Header file select and Open. Menu rows, 208pt closed control, filename help, cover as the initial selection, missing-file callout. | STEP-1, STEP-2, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6 |
| EG-2. Promote all, the Source-page jump, Place, Connect, the no-artifact callout, pan, and zoom stay. The file select does not replace that callout. | STEP-2, STEP-5, AC-1.4, AC-2.6 |
| EG-4. Palette Select after Connect, rectangle icon, help "Hold Shift to select", available when Place and Connect are dimmed. | STEP-5, AC-2.6 |
| EG-5, EG-6. Document-space rectangle from Shift-drag or from Select. Intersecting person, event, and place cards. Bridges excluded. Pan when the tool is disarmed and Shift is up. | STEP-3, STEP-4, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| EG-7. Group drag and arrow keys. Dragging an unselected subject card replaces the selection. Escape or a second press of Select disarms and keeps the selection. A click selects one card. | STEP-4, STEP-5, AC-2.6, AC-3.1, AC-3.2, AC-3.3, AC-3.4 |

| Out | Home |
| --- | --- |
| Unpromote, its confirm, and the promoted footer (EG-8, EG-9, EG-10). The board draws a ghost unlink `PVIconButton`. | S10-44 |
| An in-app viewer window. | not planned — working rule "No in-app viewer window" |
| Dropping a file on the graph or on a card. | S10-18 (Source page drop) |
| Selecting or moving bridge cards with the rectangle. A bridge still moves when it is the single clicked card, as today. | not planned — S10-D6 §2.1 |
| Shift-click to add one card. An undo stack. A group box that stays after the pointer goes up. | not planned — S10-D6 §2.1 |
| A key that arms Select. Shift is the gesture. The palette button is the tool. | not planned |
| Changing the citation composer's artifact row, or adding a plural to `citationComposer.artifactCitationCount`. | not planned — D-4 |
| A new `PVSymbol` for a missing-file icon. | not planned — D-6 |
| Arming Place still clears the selection. That is the code today. This PR does not change it. Arming Connect still does not clear the selection. | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Header | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `header` | Title, source title, count, then `promoteAllButton` and `jumpToSourcePageButton` (`PVButton` secondary, small). No file select and no Open. |
| Missing action | `EvidenceGraphView.canvasPane` | `deleteError` is a danger `PVCallout` at width 360, top right. Otherwise `noArtifactCallout` when `!model.canCite`. No file-open failure. |
| Open on the Source page | `macos/App/Features/Sources/SourceArtifactsSection.swift` `open` | `ProjectFiles.openObject(projectDir:relPath:)`. On false, `L10n.Sources.fileOpenMissing` ("That file is missing from the project folder."). The graph does not call this. |
| Artifacts | `CatalogQueryKey.sourceWorkspace` → `CatalogSourceWorkspace.artifacts` | The graph place does not load it. `PlaceRegistry` `sourceGraph` keys are graph, properties, sources list, connect rules (`macos/App/Features/Workspace/Session/PlaceRegistry.swift`). |
| Counts | `CatalogQueryKey.citationCounts` | `[String: Int]` per artifact. `CitationComposerModel.listedCount` reads it. The graph does not. |
| Cover | `CatalogSource.coverMode`, `primaryArtifactID` | `SourceArtifactsSection.isCover` is `coverMode == "artifact" && primaryArtifactID == art.id`. The graph already loads the sources list (`sourceRow`). |
| Composer row | `CitationComposerEvidenceColumn.artifactMenuRow` | Private. 48pt thumb of the live viewer or a glyph, title is label or filename, meta is kind · size-or-pixels, then the citation count. No Cover mark. No filename on the meta line when the label is set. |
| Palette | `EvidenceGraphPalette` | Person, Event, Place, a hairline, Connect. The whole stack uses `.opacity(0.55)` when `!canCite`. Person, Event, Place, and Connect are `.disabled(!model.canCite)`. |
| Mode | `EvidenceCanvasInputMode` | `.idle`, `.placing`, `.connecting`. `disarm()` clears place and connect. `prepare` and `refreshCanCite` call `disarm()` when `!canCite`. |
| Pointer | `GraphCanvasPointerController` | Empty-canvas drag in `.idle` pans. A drag that starts on a card sets `draggingItemID` and calls `onSelect`, then `onDragEnded(id, delta)`. `mouseDown` does not receive Shift. `GraphCanvasPointerToolMode` has no selecting case. |
| Selection | `EvidenceGraphModel.selectedSubjectID` | One id. `selectSubject(id:)` replaces it. `moveSubject` and arrow keys require `inputMode == .idle` and that one id. `commitDrag` takes the id as a parameter and requires `.idle`. `KeyEffect` carries one `moveSubjectID`. |
| Selection paint | `EvidenceGraphDocumentBody.cards` | `isSelected: model.selectedSubjectID == id`. The accent border and `PVColor.graphRing` already paint a selected card. |
| Edges while dragging | `EvidenceGraphEdgesHost` | Reads `pointer.offsets` per endpoint id. A bridge with no offset stays put. |
| Canvas growth | `EvidenceGraphDragCanvasSize` | Grows from `pointer.offsets` already. Peer offsets will grow the canvas with the group. |
| Tests | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `arrowRequestsMoveOfSelectedSubject` | Expects `effect.moveSubjectID`. `macos/ProvenenciaTests/GraphCanvasPointerHitTestingTests.swift` drives `mouseDown` / `mouseDragged`. `macos/ProvenenciaTests/PlaceRegistryTests.swift` expects the four graph query keys. |

**Precedent.** Source page Open is `SourceArtifactsSection.open` plus `ProjectFiles.openObject`. Copy that call. Do not copy `L10n.Sources.fileOpenMissing`. The menu row uses the composer's spacing (HStack spacing 10, then two mono lines) and adds Cover plus the filename. The thumbnail is `CachedThumbnail`. The composer row uses `PVThumbnail` of the live viewer. Do not extract `artifactMenuRow`. Group persist shares `commitDrag`'s snap through `snapAndPatch` (D-11).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The file list is `EvidenceGraphFiles` *(new)*, signatures in §6. A row is kept when `CatalogArtifact.openableRelPath` *(new)* is non-nil: `file != nil`, `fileID` non-empty, trimmed `relPath` non-empty. Order is the workspace array order. Title is the trimmed label, or the ref when the label is empty (`titleIsRef`). `initialID` uses `CatalogSource.coverArtifactID` *(new)* when that id is in the kept rows, otherwise the first kept id. | The graph's keep rule is stricter than `SourceArtifactsSection.open` (no `fileID` check, no trim) and than `SourcePageArtifactsView.hasFile` (no `relPath` check). Those two stay as they are. | Reusing `artifactMenuRow`. Extracting a shared `ArtifactMenuRow`: the graph row's title font, Cover mark, and meta line are caller-specific, and §3 Out leaves the composer row alone. |
| D-2 | The graph place loads the existing `sourceWorkspace` and `citationCounts` keys. Append them to `PlaceRegistry` `sourceGraph` after `.connectRules`, in that order, and await both in `prepare`. No new `CatalogQueryKey` and no new FFI method. The select is shown only when the workspace handle `status == .ready` and the kept list is non-empty. While loading, or on `.error`, the select and Open are absent. Citation counts use the handle value or `[:]`. A missing count, including while the counts handle is still loading, is 0, so the row shows the existing "No citations". | The queries already exist. `PlaceRegistryTests` compares key arrays with `==`. An error callout for a failed workspace load is chrome the board does not draw. | A slimmer artifact query. Waiting to show the select until counts load. |
| D-3 | Closed `PVSelect` is size `.sm`, in a 208pt frame, `menuWidth: 360`, `rowHeight: 64`, `maxVisibleRows: 6`. `PVSelect` gains `closedHelp: String? = nil`. When set, the closed label's help is that string. When nil, help stays the closed label. The graph passes `originalFilename` when it is non-empty, and nil when it is empty. Add the parameter to both `PVSelect` inits (the generic one and the `Row == PVSelectPlainRow` extension) and forward it. | Header buttons are already `.sm` (22pt). The board's 28pt hint is the design canvas measuring its own small control. 208, 360, and 64 are the board. Help must be the filename, and today `closedLabel` hard-codes help to `triggerText`. | A 28pt one-off height. Putting the filename in the closed label. |
| D-4 | The meta line is filename, kind, and byte size: new key `evidenceGraph.header.fileMeta` = "%1$@ · %2$@ · %3$@". Kind is a verbatim token for `image/jpeg` and `image/jpg` → JPEG, `image/png` → PNG, `image/tiff` and `image/tif` → TIFF, `image/gif` → GIF, `image/webp` → WEBP. `application/pdf` uses `L10n.CitationComposer.artifactKindPDF`. Any other image media type uses `artifactKindImage`. Anything else uses `artifactKindUnknown`. Size is `IngestMediaPolicy.formatByteSize`. `displayFilename` is the trimmed `originalFilename`, otherwise the last path component of `relPath`, otherwise the ref. The meta line always uses `fileMeta` with that string. The citation line is `L10n.CitationComposer.artifactCitationCount`. Cover is `L10n.Sources.coverBadge` ("Cover") with `.pvMicroCaps()` and `PVColor.textMuted`, shown only when `isCover`. The title is `PVFont.mono(size: PVTypeScale.caption)` when `titleIsRef`, and `PVFont.body(size: PVTypeScale.caption)` with `PVColor.textPrimary` otherwise. The thumbnail is `CachedThumbnail` at size 48, layout `.tile`, `relPath` `thumbnailRelPath` when that is non-empty, otherwise the openable `relPath`, plus `mediaType` and `originalFilename` so a missing raster falls through to the MIME glyph. `closedHelp`, the meta line, Open's accessibility label, and the missing-file message all use `displayFilename`. | The board's row is filename · JPEG · 4.2 MB, not the composer's "Image · 4.2 MB", and this screen does not open `ArtifactViewer` to learn a page count. The citation string already says "No citations" at 0 and "%lld citations" otherwise. The board's Cover is micro type beside the title, not the Source page's `PVBadge`. | Opening each file for a page count. A new plural on the shared citation key, which would change the composer. A `PVBadge` for Cover. A new "Cover" key beside `coverBadge`. Moving `kindToken` onto `ArtifactViewerKind` or `PVFileTypeGlyph`, which return a viewer kind and a symbol. |
| D-5 | Open is `PVButton` secondary, small, icon `.externalLink`, title `L10n.EvidenceGraph.openFile` ("Open"), `.help` "Open in its app", accessibility label `L10n.EvidenceGraph.openFileAccessibility(filename:)` ("Open {filename} in its app"), identifier `evidenceGraph.header.openFile`. It calls `ProjectFiles.openObject(projectDir: session.projectKey.projectDir, relPath:)`. It does not change `selectedSubjectIDs` or the camera. The select's accessibility label is "Artifact file". Identifier `evidenceGraph.header.fileSelect`. | The board's button. The same open function as the Source page. | `L10n.Sources.fileOpenMissing` as the graph copy. An in-app viewer. |
| D-6 | The model owns the file list, the same way it owns `sourceRow`. `fileRows` reads the `sourceWorkspace` handle and returns `[]` unless `status == .ready`. It calls `fileBearing` with `sourceRow?.coverArtifactID`. `citationCount(for:)` reads the `citationCounts` handle and returns 0 when the id is missing. `selectedFileArtifactID: String` starts empty. `resolvedFileArtifactID` is the stored id when it is in `fileRows`, otherwise `initialID`, otherwise `""`. `selectFile(_ id: String)` stores the id and clears `fileOpenError`. `openSelectedFile()` uses the resolved row's `relPath`. On false, `fileOpenError` is `fileNotFoundMessage(name: title, filename: displayFilename)`. On true, clear `fileOpenError` before the call. The callout is the next `else if` after `deleteError`: `pendingImpact == nil`, `deleteError == nil`, `fileOpenError != nil`, and `resolvedFileArtifactID == selectedFileArtifactID`. Width 360, identifier `evidenceGraph.fileOpen.callout`, icon omitted. An armed Place or Connect banner stays hidden while it shows, the same as `deleteError`. | `sourceRow` is private, so the view cannot compute the cover. `CitationComposerModel` keeps the selected artifact id on the model. `.cursor/rules/macos-l10n.mdc` puts argument copy on an `L10n` format function. | A view-owned list plus `syncFileSelection`. A new symbol. A dialog. Reusing `deleteError`. An injected `fileOpener`. AC-1.3 and AC-1.5 are manual. |
| D-7 | `selectedSubjectIDs: Set<String>` is the stored selection. `selectedSubjectID` is get-only: the only member when the set has one id, otherwise nil. There is no setter. `prepare`'s write and `activateSubject` write the set (`activateSubject` sets the set to `[id]` and then `activatedSubjectID`). `selectSubject(id:)` replaces the set with that one id, or empties it when id is nil, and does not change `armedSelect`. A bridge id is allowed there. `replacePrimarySelection(_ ids: Set<String>)` keeps only ids the current graph snapshot lists as subjects, and clears `activatedSubjectID` when it is no longer in the set. Deleting a subject does `selectedSubjectIDs.remove(id)` and clears `activatedSubjectID` when it was that id. `isSelected` reads the set. Escape order: the existing cite-tool branch (which calls `disarm()`), then `armedSelect` (clear only that flag), then activation, then `selectSubject(id: nil)` when the set is non-empty. | One id cannot represent the marquee. Return, Space, and the bridge highlight still read `selectedSubjectID`. The two writes today are `prepare` and `activateSubject`. | A setter on `selectedSubjectID`. A second selection model. Shift-click add. |
| D-8 | Marquee geometry lives on `GraphCanvasPointerController`. `GraphCanvasPointerToolMode.selecting` is the armed tool. In `.selecting` the pointer matches `.idle` for cards: click, card drag, nested actions, and hover. The only difference is that an empty-canvas drag marquees instead of panning. `mouseDown` gains `shiftDown: Bool = false`. `pendingClick` becomes `pendingClick(start:targetID:shiftDown:)`. `mouseUp` on a card drag clears the anchor offset, every peer offset, `draggingItemID`, and `draggingPeerIDs`. Empty-canvas drag also marquees when mode is `.idle` and that flag is true. A drag that starts on a hit target never marquees. Every `switch mode` (`mouseDown`, `mouseUp`, `mouseMoved`) and the `mouseDragged` `mode == .idle` guards gain `.selecting` in the same step as the case. Gesture case `marquee(origin:current:)`. `shiftActiveDocumentOrigin` shifts both points. `marqueeRect` is the normalized document rect, nil when the gesture is not a marquee. On mouse up the pointer calls `onMarqueeEnded` with `idsIntersecting` of that rect against `hitTargets`, then clears `marqueeRect`. `GraphCanvasDocumentView.mouseDown` passes `event.modifierFlags.contains(.shift)`. | The gesture conflict already lives in this controller. A defaulted parameter keeps the existing press tests compiling. Select is a thin variant of idle. | A SwiftUI `DragGesture` on top of the AppKit document. A marquee in screen space. A fourth gesture family for `.selecting`. |
| D-9 | The rectangle is a `Rectangle` in a child view that is the only reader of `marqueeRect`. Fill `PVColor.accent.opacity(0.09)`, stroke `PVColor.accent` at line width 1.5, `allowsHitTesting(false)`. It is not rounded and it is not divided by magnification. `EvidenceGraphDocumentBody.body` does not read `marqueeRect`. | The board: accent stroke, 9% accent fill, drawn in document space so it scales with zoom. The document body must not follow a property that changes every pointer event. | `screenPixelDocumentWidth`, which would hold the stroke at 1 screen pixel. A kit marquee primitive. |
| D-10 | `GraphCanvasPointerHitTesting.idsIntersecting(_ rect: CGRect, in targets: [GraphCanvasHitTarget]) -> Set<String>` returns ids whose `frame.intersects(rect)`. A shared edge is not a hit: `CGRect.intersects` returns false for a zero-area overlap. The pointer calls it on mouse up against its own `hitTargets`. `onMarqueeEnded` delivers that set. `replacePrimarySelection` then drops bridge ids. A miss delivers an empty set and clears the selection. | `publishHitTargets` already stores the document-space `edgeFrame`s. A second frame loop in the view would have to track that function. | Rebuilding subject frames at release. Treating a shared edge as a hit. |
| D-11 | Signatures in §6. Card-drag start calls `onDragBegan` and does not call `onSelect`. `beginGroupDrag(anchor:) -> Set<String>` keeps `selectedSubjectIDs` when the anchor is a subject already in the set, and otherwise `selectSubject(id: anchor)`. The return value is the subjects that move, including the anchor. The model stores that as `groupDragIDs: [String]` until commit: the anchor first, then the other selected subject ids in snapshot order. The pointer keeps `draggingItemID` as the anchor and `draggingPeerIDs` as the rest. `Gesture.draggingCard` stores those peers. The same offset is written for the anchor and each peer. `onDragEnded` still passes the anchor and the delta. The view calls `commitGroupDrag(anchor:documentDelta:)`, which snaps each id in `groupDragIDs` through `snapAndPatch` and one `lastPositionWrite` task that persists them in order, then clears `groupDragIDs`. `commitDrag` stays the one-card API and calls the same helper. A card's `isDragging` is `draggingItemID == id || draggingPeerIDs.contains(id)`. | EG-7. Edges already follow `pointer.offsets`. Calling `onSelect` at drag start would collapse the set before the peers are read. `draggingItemID` stays the anchor so `EvidenceGraphDragCanvasSize` and the existing `onChange`s keep compiling. | A second `draggingItemIDs` that duplicates the anchor. Moving bridges by the group delta. One store transaction for the group. |
| D-12 | `commitDrag`, `commitGroupDrag`, and `moveSubject` run when `inputMode` is `.idle` or `.selecting`. Arrow keys do too. `KeyEffect.SubjectMove` is `Equatable`. It replaces `moveSubjectID`, `moveFromX`, `moveFromY`, `moveDeltaX`, and `moveDeltaY`. Fields: `subjectID: String`, `fromX`, `fromY`, `deltaX`, `deltaY`, all `Int64`. `moves: [SubjectMove]` defaults to `[]`, in snapshot subject order. `moveSubjects(_ moves: [KeyEffect.SubjectMove])` patches every cell, then assigns one `lastPositionWrite` that persists them in that order. The view calls it once. `moveSubject` stays the one-card function and is not what the arrow path calls. Return and Space still require `.idle` and a single `selectedSubjectID`. | A group move while Select is armed is EG-7. One task matches `commitGroupDrag`. A `Set` has no order, so snapshot order keeps T-9 stable. | A view loop of `moveSubject`. Disarming Select before an arrow works. |
| D-13 | `armedSelect: Bool` sits beside the stored `armedKind` and `armedConnect`. `disarm()` clears place, connect, and `armedSelect`. Callers that already call `disarm()` (edit, create cancel, create confirm, palette Escape, the armed-banner cancel) also leave Select off. `disarmCiteTools()` *(new)* clears place and connect and leaves `armedSelect`. `prepare` and `refreshCanCite` call `disarmCiteTools()` when `!canCite`. `toggleArm` and `toggleConnect` set `armedSelect = false` only when they arm, not when they toggle themselves off. `toggleSelect` does not read `canCite`. Escape while `armedSelect` clears only that flag and does not clear `selectedSubjectIDs`. `selectSubject(id: nil)` does not clear `armedSelect`. No crosshair: the two `onChange`s and `onDisappear` stay on `armedKind` and `armedConnect`. `accessibilityGraphLabel` gains no Select branch. `emptyOverlay` stays `inputMode == .idle`, so it hides while Select is armed. | Select has to stay available on a Source with nothing to cite. This PR does not rewrite the place and connect arm path. `inputMode` is already the exclusive enum. | One `ArmedTool` enum in this PR. That fold is a later cleanup, named under Next adds. A banner for the armed Select tool. |
| D-14 | The palette is two groups in one card. The first group is Person, Event, Place, the existing hairline, and Connect, and it keeps `.opacity(model.canCite ? 1 : 0.55)`. A second hairline and Select sit outside that opacity. Person, Event, Place, Connect, and Select all go through one private `paletteTool` in `EvidenceGraphPalette`. Parameters: title, `armed`, `enabled`, `help: LocalizedStringResource?`, focus case, identifier, action, and `icon: (Bool) -> some View` (the Bool is `armed`, so a kind button can tint). `paletteTool` calls `toolAccessibilityLabel(name:armed:)` itself. Delete `toolAccessibilityLabel(for:armed:)` and `connectToolAccessibilityLabel(armed:)`. Kind buttons and Connect pass `enabled: model.canCite`. Select passes `enabled: true`, help `toolSelectHelp`, icon `PVIcon(.regionRectangle, size: 17)`, focus `EvidenceGraphFocus.toolSelect`, identifier `evidenceGraph.palette.select`. | EG-4. `toolButton` and `connectButton` already share the armed chrome. The two label functions only forward. | A new palette component. Keeping the two forwarding wrappers. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `EvidenceGraphFiles` *(new)* | type | `CitationComposerEvidenceColumn.artifactMenuRow` | That row is private, uses the live viewer, and has no Cover mark. One graph call site, so it stays a snowflake in the feature folder. |
| `PVSelect.closedHelp` | parameter | `PVSelect.closedLabel` help, which is `triggerText` | The closed label and the help are different strings on this one control. Default nil keeps every current caller. |
| `GraphCanvasPointerToolMode.selecting`, `marquee` gesture, `marqueeRect`, `onMarqueeEnded`, `onDragBegan`, `draggingPeerIDs` | cases on the canvas pointer | `GraphCanvasPointerController` pan and card drag | The gesture already lives there. A second drag recognizer would fight it. `draggingItemID` stays the anchor. |
| `GraphCanvasPointerHitTesting.idsIntersecting` *(new)* | func | `topmostTarget` | Hit testing of a rect against `hitTargets` is the same helper. The model drops bridge ids after. |
| `selectedSubjectIDs`, `armedSelect`, `beginGroupDrag`, `commitGroupDrag`, `disarmCiteTools` | fields and funcs | `selectedSubjectID`, `commitDrag`, `disarm` | The one-id API stays for click, Return, and the bridge highlight. |
| `L10n.EvidenceGraph` keys in §6 Strings | keys | `L10n.EvidenceGraph`, `L10n.CitationComposer.artifactCitationCount` | The citation count and the PDF / Image / File words already exist and are reused. |

**Builds on.** `PVSelect`'s `row:` builder, `CachedThumbnail`, `ProjectFiles.openObject`, `PlaceRegistry` query keys, `GraphCanvasPointerController`, `commitDrag` / `persistPositionOrRevert`, `EvidenceGraphEdgesHost` offsets, and the existing selection ring.
**Deletes.** The single-move fields on `KeyEffect` (`moveSubjectID`, `moveFromX`, `moveFromY`, `moveDeltaX`, `moveDeltaY`). The palette's container opacity moves onto the cite-tool group. Nothing else becomes dead.
**Next adds.** S10-44 fills the unlink button on the promoted footer. This PR does not change that footer. The selection set is what a later group command would read. No later PR in the map adds another marquee. `openableRelPath` and `coverArtifactID` are the graph's predicates. Adopting them at `SourceArtifactsSection.open`, `SourcePageArtifactsView.hasFile`, and `CitationComposerModel.defaultPendingArtifactID` is a later change, because those three predicates are not the same rule. A shared artifact menu row waits for a third caller. Folding `armedKind`, `armedConnect`, and `armedSelect` into one `armedTool` waits until a fourth tool needs the same slot.

### Schema

None — no catalog change.

### Go core

None — open uses the file already stored for the artifact. Positions use `setSubjectPosition`, which already exists.

### FFI

None — `sourceWorkspace` and `citationCounts` are already registered. `CatalogQueryRegistry` is not edited.

### macOS

The header's trailing cluster, leading with the new controls, then the two buttons that exist. Spacing stays `PVSpacing.space3`.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| File select | `PVSelect` | `header` in `EvidenceGraphView` | D-1, D-3, D-4. Shown only when the kept list is non-empty. |
| Menu row | snowflake | `EvidenceGraphFiles` plus a private row in `EvidenceGraphView` | 48pt `CachedThumbnail`, title, Cover, meta, citation count. |
| Open | `PVButton` | `header` | D-5. |
| Missing-file callout | `PVCallout` danger | `canvasPane`, the `deleteError` slot | D-6. Width 360. |
| Select | palette button | `EvidenceGraphPalette` | D-13, D-14. |
| Rectangle | snowflake | child of `EvidenceGraphDocumentBody` | D-9. |
| Selection ring | existing card chrome | `isSelected` reads `selectedSubjectIDs` | No new ring. |

Frames this PR ships, in the board's words:

1. **One file.** The 208pt select shows that artifact's label. Open sits beside it, then Promote all, then Jump to Source page. All four are small.
2. **Several files.** The menu is 360pt wide and each row is 64pt: 48pt thumbnail, label (mono when it is the ref), Cover on the cover row, then `filename · KIND · size`, then the citation count. The cover is selected when it has a file. The closed control shows the label only.
3. **Another file chosen.** The closed label changes. Nothing opens.
4. **Open.** The button's help is "Open in its app". The graph does not gain a sheet or a toast.
5. **No file.** The select and Open are not in the header. The no-artifact callout stays when the Source cannot be cited. Place and Connect are dimmed. Select is not.
6. **Missing file.** Danger callout, top right, title "File not found", message naming the label and the filename. The select still shows that artifact.
7. **Select armed.** Palette order: Person, Event, Place, hairline, Connect, hairline, Select. Select uses the rectangle icon at 17pt. Armed is accent fill and the graph ring, the same as Connect. Help is "Hold Shift to select".
8. **The rectangle.** Accent stroke and 9% accent fill, in document space. On release the people inside have the ring. The relationship does not. The rectangle is gone.
9. **The move.** Dragging one selected person moves every selected person by that delta. The relationship stays. Edges follow the people.
10. **Pan still pans.** Select disarmed, Shift up, empty-canvas drag pans. No rectangle and no new selection.

States: no file-bearing artifact hides both controls. Workspace still loading hides both controls. A citation-count query that is still loading shows "No citations" until the numbers arrive. A rectangle that misses clears the selection. One selected card keeps today's Return and Space behavior. Two or more selected cards ignore Return and Space. Select armed ignores Return and Space. A Source with `canCite == false` still arms Select.

Accessibility: the select's label is "Artifact file". Open's label is "Open {filename} in its app". Select's label follows the place-tool pattern (`name, tool, on/off`). The rectangle is not an accessibility element. VoiceOver selection of one card stays the rotor's `onRotorSelect`, which calls `selectSubject`. While Select is armed the graph's spoken label stays `L10n.Workspace.evidenceGraphTitle`. The empty-graph overlay is hidden while Select is armed, because it already requires `inputMode == .idle`.

### Signatures

`CatalogArtifact.openableRelPath: String?` — the trimmed `file.relPath` when `file != nil`, `fileID` is non-empty, and the trimmed path is non-empty. Otherwise nil.

`CatalogSource.coverArtifactID: String?` — `primaryArtifactID` when `coverMode == "artifact"` and that id is non-empty. Otherwise nil.

`EvidenceGraphFiles` is an enum with no cases.

```swift
struct Row: Equatable {
    var id: String
    var ref: String
    var title: String
    var titleIsRef: Bool
    var filename: String
    var relPath: String
    var mediaType: String
    var byteSize: Int64
    var thumbnailRelPath: String
    var isCover: Bool
}

static func fileBearing(_ artifacts: [CatalogArtifact], coverArtifactID: String?) -> [Row]
static func initialID(in rows: [Row], coverArtifactID: String?) -> String?
static func kindToken(mediaType: String) -> String
```

`fileBearing` keeps workspace order. `isCover` is `id == coverArtifactID`. `initialID` returns `coverArtifactID` when that id is in `rows`, otherwise `rows.first?.id`.

`L10n.EvidenceGraph.fileMeta(filename:kind:size:) -> String`, `openFileAccessibility(filename:) -> String`, `fileNotFoundMessage(name:filename:) -> String`. The other new keys are `LocalizedStringResource` properties named in the Strings table. Cover reuses `L10n.Sources.coverBadge`. There is no `fileCover` key.

`beginGroupDrag(anchor:) -> Set<String>` returns the moving ids. The model also stores `private(set) var groupDragIDs: [String]`: anchor first, then the other selected subject ids in the graph snapshot's subject order. Tests read it. `GraphCanvasPointerController.onDragBegan: ((String) -> Set<String>)?`. `onMarqueeEnded: ((Set<String>) -> Void)?`. `draggingPeerIDs: Set<String>` is empty for a one-card drag. `groupDragIDs` is cleared at the end of `commitGroupDrag`.

`@discardableResult func commitGroupDrag(anchor: String, documentDelta: CGSize) -> [(subjectID: String, gridX: Int64, gridY: Int64)]`, in `groupDragIDs` order. `snapAndPatch` is private. It bumps `positionGeneration` the way `commitDrag` does today (`bumpPositionGeneration` returns `Int`), patches the cell, and returns `(gridX: Int64, gridY: Int64, generation: Int)?`. The persist task captures that `generation` from the return value and passes it to `persistPositionOrRevert`. `commitDrag` calls it once, then assigns one `lastPositionWrite`. `commitGroupDrag` calls it once per id in `groupDragIDs` order, then assigns one `lastPositionWrite` that persists those cells in that same order. An id whose snap returns nil is omitted from the array.

`EvidenceGraphFiles.Row.displayFilename` is the trimmed `filename`, otherwise the last path component of `relPath`, otherwise `ref`.

`func openSelectedFile()`. `func selectFile(_ id: String)`. `func citationCount(for artifactID: String) -> Int`. `var fileRows: [EvidenceGraphFiles.Row]`. `var resolvedFileArtifactID: String`.

`func moveSubjects(_ moves: [KeyEffect.SubjectMove])`. The view's arrow branch calls it once inside the existing key `Task`.

`KeyEffect.SubjectMove` fields are `subjectID: String`, `fromX: Int64`, `fromY: Int64`, `deltaX: Int64`, `deltaY: Int64`.

### Strings

New keys under `L10n.EvidenceGraph`. Add each to `macos/App/Resources/Localizable.xcstrings` per `add-localized-string`.

| Key | English |
| --- | --- |
| `evidenceGraph.header.artifactFile` | Artifact file |
| `evidenceGraph.header.openFile` | Open |
| `evidenceGraph.header.openFileHelp` | Open in its app |
| `evidenceGraph.header.openFileAccessibility` | Open %1$@ in its app |
| `evidenceGraph.header.fileMeta` | %1$@ · %2$@ · %3$@ |
| `evidenceGraph.fileNotFound.title` | File not found |
| `evidenceGraph.fileNotFound.message` | %1$@ points at %2$@, which is no longer where Provenencia stored it. |
| `evidenceGraph.palette.select` | Select |
| `evidenceGraph.palette.selectHelp` | Hold Shift to select |

Removed: none. Reused, not duplicated: `L10n.CitationComposer.artifactCitationCount`, `artifactKindPDF`, `artifactKindImage`, `artifactKindUnknown`, `artifactKindDetail`, `L10n.Sources.coverBadge`. Property names: `artifactFile`, `openFile`, `openFileHelp`, `openFileAccessibility(filename:)`, `fileMeta(filename:kind:size:)`, `fileNotFoundTitle`, `fileNotFoundMessage(name:filename:)`, `toolSelect`, `toolSelectHelp`.

## 7. Steps

**STEP-1 — File list**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphFiles.swift` *(new)*, `macos/App/Platform/GenealogyStore.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skill: `add-swift-test`.
Do: Add `openableRelPath` and `coverArtifactID` and `EvidenceGraphFiles` as D-1 and the §6 signatures. `kindToken` follows D-4. Do not add a message function here. T-2 in this step asserts the kind tokens only. Register the new file in the app target the way `EvidenceGraphPalette.swift` is registered: one `PBXBuildFile`, one `PBXFileReference`, one child of the EvidenceGraph group, one Sources build-phase entry. Use two fresh 24-hex ids that `rg` shows are unused in the pbxproj. Tests T-1 and T-2.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/EvidenceGraphModelTests CODE_SIGNING_ALLOWED=NO` passes T-1 and T-2.

**STEP-2 — Header select, Open, and the callout**
Files: `macos/App/DesignSystem/Components/Select/PVSelect.swift`, `macos/App/DesignSystem/README.md`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PlaceRegistryTests.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skills: `add-localized-string`, `add-ui-component`, `add-workspace-place`.
Do: Add `closedHelp` (D-3) and the strings, including `fileNotFoundMessage`. In the DesignSystem README Select contract, the help sentence becomes: help shows the full closed label, unless `closedHelp` is set. Append `.sourceWorkspace` then `.citationCounts` after `.connectRules` on the `sourceGraph` place, and await both in `prepare` (D-2). Add `fileRows`, `resolvedFileArtifactID`, `selectFile`, `citationCount(for:)`, `fileOpenError`, and `openSelectedFile` (D-5, D-6). The header select binds to `resolvedFileArtifactID` / `selectFile`. Put the select and Open in `header` ahead of Promote all. Menu row as D-4. Callout as D-6, the next `else if` after `deleteError` and before `noArtifactCallout`. Update `resolvesSourceGraphDistinctFromPage` and `queryKeysMatchDesignTable`. Extend T-2 with the missing-file sentence. Add T-13.
Check: `python3 scripts/check-localizable-xcstrings.py` passes. `-only-testing:ProvenenciaTests/PlaceRegistryTests` passes. `-only-testing:ProvenenciaTests/EvidenceGraphModelTests` passes.

**STEP-3 — Rectangle**
Files: `macos/App/Features/GraphCanvas/GraphCanvasHitTarget.swift`, `macos/App/Features/GraphCanvas/GraphCanvasPointerController.swift`, `macos/App/Features/GraphCanvas/GraphCanvasDocumentView.swift`, `macos/App/Features/GraphCanvas/README.md`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/ProvenenciaTests/GraphCanvasPointerHitTestingTests.swift`.
Do: Add `.selecting`, `shiftDown`, the `marquee` gesture, `marqueeRect`, `onMarqueeEnded`, and `idsIntersecting` (D-8, D-9, D-10). In `.selecting`, card click, card drag, nested actions, and hover match `.idle`. Update every `switch mode` and the `mouseDragged` idle guards in this step. Shift both marquee points in `shiftActiveDocumentOrigin`. Pass Shift from `mouseDown`. Draw the rectangle from a child view. Leave `onMarqueeEnded` unset until STEP-4. Update the GraphCanvas README for the marquee: the pointer row gains marquee, the host-callback row gains marquee-ended, and Input mapping gains the Shift-drag and selecting-mode marquee. Tests T-3, T-4, T-5, T-6.
Check: `-only-testing:ProvenenciaTests/GraphCanvasPointerHitTestingTests` passes. `-only-testing:ProvenenciaTests/GraphCanvasPointerPressTests` passes.

**STEP-4 — Group move**
Files: `macos/App/Features/EvidenceGraph/EvidenceCanvasInputMode.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Features/GraphCanvas/GraphCanvasPointerController.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`.
Do: Add `case selecting` on `EvidenceCanvasInputMode` and `armedSelect` on the model. `inputMode` returns `.selecting` when `armedSelect` is set and no sheet is up, ahead of `.idle`. Nothing arms it until STEP-5. In this same step, `syncPointerMode` maps `.selecting` to `pointer.mode = .selecting` and does not clear the selection, and the `onHover` switch's `.selecting` case matches `.idle` (break). Add the selection set, `beginGroupDrag`, `onDragBegan`, `draggingPeerIDs`, `commitGroupDrag`, `snapAndPatch`, and `KeyEffect.SubjectMove` (D-7, D-11, D-12). `beginGroupDrag` reads subjects from the graph query the model already holds (`graphKey`). If that handle has no value, `selectSubject(id: anchor)` and return `[anchor]`. Card `isSelected` reads the set. `isDragging` reads `draggingItemID` and `draggingPeerIDs`. Click still uses `onSelect` → `selectSubject`. Wire `onMarqueeEnded` to `replacePrimarySelection`. Arrow handling calls `moveSubjects(effect.moves)` once. Update `arrowRequestsMoveOfSelectedSubject` to read `moves`. Deleting a subject removes that id from the set. Allow `.idle` and `.selecting` in `commitDrag`, `commitGroupDrag`, and `moveSubject`. The view's drag-end calls `commitGroupDrag`. In the GraphCanvas README, the pointer row gains group-drag peers, the host-callback row gains drag-began, and the dragging paragraph names `draggingItemID` as the anchor and `draggingPeerIDs` as the rest. Tests T-7, T-8, T-9, T-11, T-12.
Check: `-only-testing:ProvenenciaTests/EvidenceGraphModelTests` passes. `-only-testing:ProvenenciaTests/GraphCanvasPointerPressTests` passes.

**STEP-5 — Palette Select**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphPalette.swift`, `macos/App/Features/EvidenceGraph/EvidenceCanvasInputMode.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skills: `add-ui-component`, `add-localized-string`.
Do: `armedSelect`, `EvidenceCanvasInputMode.selecting`, and the `syncPointerMode` / `onHover` cases already exist from STEP-4. Add `toggleSelect` and `disarmCiteTools` (D-13, D-14). Point `prepare` and `refreshCanCite` at `disarmCiteTools`. `disarm()` also clears `armedSelect`. `toggleArm` and `toggleConnect` clear `armedSelect` only when they arm. Escape in `handleKey` clears `armedSelect` and keeps the set. Palette layout and `paletteTool` as D-14. Test T-10.
Check: `-only-testing:ProvenenciaTests/EvidenceGraphModelTests` passes.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do: Per the Definition of done. Leave [S10-D6](../design/S10-D6-evidence-graph.md) in `design/`. S10-44 still uses it. Do not archive it in this PR.
Check: links in the map row and `completed.md` resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits. Suggested split: the file select and Open; the rectangle and group move; the palette Select. Docs ride with the last commit.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `fileBearingListKeepsCoverAndDropsFileless` *(new)* | `a`: id `a`, file non-nil, `fileID` `file-a`, `relPath` `objects/a.jpg`, label `Page 1`, ref `ART-AAAA`. `b`: id `b`, the same shape, label `Page 2`. `c`: id `c`, `file` nil. `d`: id `d`, file present, label empty, ref `ART-7Q2KD`. Also drop an artifact whose `fileID` is empty, and one whose `relPath` is only spaces. `coverArtifactID` `a` → ids `a`, `b`, `d`, title of `d` is `ART-7Q2KD`, `titleIsRef` only on `d`, initial `a`. `coverArtifactID` `c` → initial `a`. Empty input → initial nil. | AC-1.1, AC-1.4, D-1 |
| T-2 | Swift | `EvidenceGraphModelTests.swift` `fileKindTokenAndMissingCopy` *(new)* | `image/jpeg` → JPEG. `application/pdf` → the PDF string. `text/plain` → the File string. `L10n.EvidenceGraph.fileNotFoundMessage(name: "Page 7", filename: "1881-peel-n-p7.jpg")` → the AC-1.5 sentence. The message lines are added in STEP-2, when the key exists. | AC-1.5, D-4, D-6 |
| T-3 | Swift | `macos/ProvenenciaTests/GraphCanvasPointerHitTestingTests.swift` `GraphCanvasPointerPressTests` `shiftDragOnEmptyCanvasMarquees` *(new)* | One hit target, frame `CGRect(x: 200, y: 200, width: 100, height: 80)`. Do not use the shared `controller()` helper. Idle, `shiftDown: true`, down at (10, 10), drag to (80, 90) → `marqueeRect` is that rect, `draggingItemID` is nil. Same drag with `shiftDown: false` → `marqueeRect` is nil. Shift-drag that starts inside the card frame → `draggingItemID` is set, `marqueeRect` is nil. | AC-2.1, AC-2.4, AC-2.5, D-8 |
| T-4 | Swift | `GraphCanvasPointerHitTestingTests.swift` `idsIntersectingUsesCardFrames` *(new)* | Targets are `GraphCanvasHitTarget` with `acceptsConnect: true` and no actions. Rect `(0, 0, 100, 100)`. `p1` at `(10, 10, 20, 20)` and `p2` at `(50, 50, 20, 20)` are included. `p3` at `(200, 0, 20, 20)` is excluded. A frame at `(100, 0, 20, 20)` shares the rect's right edge and is excluded. | AC-2.2, D-10 |
| T-5 | Swift | `GraphCanvasPointerPressTests` `selectingModeMarqueesWithoutShift` *(new)* | Mode `.selecting`, `shiftDown: false`, empty drag from (10, 10) to (80, 90) with no target there → `marqueeRect` is set. | AC-2.3, D-8 |
| T-6 | Swift | `GraphCanvasPointerPressTests` `marqueeReleaseClearsTheRect` *(new)* | After the T-3 drag, mouse up → `onMarqueeEnded` receives the ids whose frames intersect that rect, and `marqueeRect` is nil. | AC-2.2, D-8 |
| T-7 | Swift | `EvidenceGraphModelTests.swift` `primarySelectionDropsBridges` *(new)* | Snapshot with persons `p1`, `p2` and bridge `b1`. `replacePrimarySelection(["p1", "b1"])` → set `p1`. `selectSubject(id: "b1")` → `selectedSubjectID` is `b1`. `armedSelect = true` then `selectSubject(id: nil)` leaves `armedSelect` true and the set empty. | AC-2.2, AC-3.4, D-7 |
| T-8 | Swift | `EvidenceGraphModelTests.swift` `groupDragKeepsTheSet` *(new)* | Set `p1`, `p2`, snapshot order `p1` then `p2`. `beginGroupDrag(anchor: "p1")` returns both, and `groupDragIDs` is `["p1", "p2"]`. `beginGroupDrag(anchor: "p3")` returns `["p3"]` and the set is `p3`. | AC-3.1, AC-3.2, D-11 |
| T-9 | Swift | `EvidenceGraphModelTests.swift` `arrowMovesEachSelectedSubject` *(new)* | Persons `p1` at (2, 3) and `p2` at (4, 3), both selected, `armedSelect` true. Right arrow → two `SubjectMove` values, `deltaX` 1, `fromX` 2 and 4. `arrowRequestsMoveOfSelectedSubject` reads `moves` for the one-card case. | AC-3.3, D-12 |
| T-10 | Swift | `EvidenceGraphModelTests.swift` `selectStaysAvailableWithoutAnArtifact` *(new)* | Sources list row for this Source with `hasArtifact: false`. `toggleSelect` sets `armedSelect`. `toggleArm(.person)` does not set `armedKind`. Escape clears `armedSelect` and keeps `selectedSubjectIDs`. `L10n.string(L10n.EvidenceGraph.toolSelectHelp)` is "Hold Shift to select". | AC-2.6, D-13 |
| T-11 | Swift | `EvidenceGraphModelTests.swift` `commitGroupDragSnapsEachCard` *(new)* | Copy `arrowRequestsMoveOfSelectedSubject`'s `setQueryValue` setup with `p1` at (2, 3) and `p2` at (4, 3). `replacePrimarySelection` of both, then `beginGroupDrag(anchor: "p1")`. `commitGroupDrag(anchor: "p1", documentDelta: CGSize(width: 40, height: 0))` returns `(p1, 3, 3)` then `(p2, 5, 3)`. `lastPositionWrite` is non-nil. | AC-3.1, D-11 |
| T-12 | Swift | `GraphCanvasPointerPressTests` `dragPeersShareTheAnchorOffset` *(new)* | `onDragBegan` returns the anchor plus `peer`. Drag the anchor past 4pt. `offsets[peer]` equals `offsets[anchor]`. `draggingItemID` is the anchor. `draggingPeerIDs` is `peer`. After mouse up both offsets are nil and `draggingPeerIDs` is empty. | AC-3.1, D-11 |
| T-13 | Swift | `EvidenceGraphModelTests.swift` `resolvedFileFallsBackAndSelectFileClearsTheError` *(new)* | After `prepare`, set the `sourceWorkspace` value's `artifacts` to the T-1 kept rows and `source.coverMode` `artifact` with `primaryArtifactID` `a`. `selectedFileArtifactID = "missing"` → `resolvedFileArtifactID` is `a`. Set `fileOpenError` to a non-nil string, then `selectFile("b")` clears it. | D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. The suites this PR is most likely to break: `EvidenceGraphModelTests`, `GraphCanvasPointerHitTestingTests`, `PlaceRegistryTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each. AC-1.3 and AC-1.5 are manual check steps 1 and 2. There is no open-file seam.
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief stays in design/ until S10-44 lands. design/README.md is not archived in this PR
- [ ] Domain doc updated if a rule in docs/*.md changed: none
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

This is the slice 5 graph half. Use a Source that has two artifacts with files, one of them the cover, one artifact with no file, three person cards, and one relationship card between them.

1. Open the graph. The select shows the cover's label. Open the menu: both files are listed, the fileless artifact is not, and each row shows a thumbnail, the filename, and a citation count. Choose the other file. It does not open. Press Open. That file opens in the system app. The cards and the camera have not moved. (AC-1.1, AC-1.2, AC-1.3, AC-1.6)
2. Move that artifact's object file out of the project objects directory, press Open, and read the danger callout. The select still shows that artifact. Choose the other file. The callout goes away. Put the file back when you are done. (AC-1.5)
3. On a Source whose artifacts have no files, the select and Open are absent. Promote all and Jump to Source page remain. On a Source with no artifacts, the no-artifact callout remains, Place and Connect are dimmed, and Select still arms. Its help reads "Hold Shift to select". (AC-1.4, AC-2.6)
4. With Select disarmed, hold Shift and drag on empty canvas across the three people and the relationship. The people are selected. The relationship is not. The rectangle is gone. Drag one selected person. All three land together. Reload the graph. They are still on those cells. The relationship did not move. (AC-2.1, AC-2.2, AC-3.1)
5. Arm Select. Without Shift, drag on empty canvas. The rectangle appears. Press Escape. Select disarms and the people stay selected. Click one person. Only that person is selected. With Select disarmed and Shift up, drag empty canvas. The view pans and the selection does not change. (AC-2.3, AC-2.4, AC-2.6, AC-3.4)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A group drag near the document edge has to grow the canvas | `EvidenceGraphDragCanvasSize` already reads every `pointer.offsets` entry. Setting peer offsets uses that path. |
| `CGRect.intersects` excludes a shared edge | T-4 locks the exclusion. A card that only touches the rectangle is not selected. |
| `lastPositionWrite` is one task | `commitGroupDrag` assigns one task that persists every id, so a test can await it the way `commitDrag` tests do. |

**Stop and report if:**

- `ProjectFiles.openObject` no longer returns false when the object is missing.
- `sourceWorkspace` or `citationCounts` is not a `CatalogQueryKey`, or the workspace value is not `CatalogSourceWorkspace`.
- `commitDrag` no longer snaps with `GraphCanvasGridMapping` and `persistPositionOrRevert`.
- `EvidenceGraphDocumentBody` cannot host a child that is the only reader of a pointer property without the parent body reading it.
- Showing the select requires a new FFI method or a new query key.
- The brief's EG-1…EG-7 and an AC in §2 disagree.
- A group move cannot persist without a new transaction type or a change to bridge positions.

## 12. Open questions

None.

## 13. PR

**Title:** S10-20 — Open a file from the graph and box-select subject cards
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-20-open-file-box-select.md`.

## Landed

Differences from this plan: `PVSelect`'s `closedHelp` is the row's `displayFilename` (D-4) in every case, not the raw original filename D-3 names, so a row with no stored name still helps with the path's last component or the ref. `openSelectedFile()` also stores the resolved id in `selectedFileArtifactID`; without that, a researcher who never changed the select would not see the missing-file callout, whose guard compares the two. The arrow branch calls `moveSubjects` directly, not inside a `Task`, because it is not async and the patch has to land before the next key press reads the cells. `snapAndPatch` returns a non-optional tuple, since no case returned nil. `paletteTool`'s `icon` is a non-escaping `@ViewBuilder (Bool) -> some View`. The new pbxproj ids are `AA...0FE001` and `AA...0FE002`. The rectangle is a child of the base layer, so cards paint above it. The `evidenceGraph.palette.select` and `selectHelp` strings landed with STEP-2's other strings. The manual check was not run by the implementer; it is for the reviewer. The menu does not open the original object when `thumbnailRelPath` is empty; the user chose the MIME glyph over D-4's openable-path fallback. Arming Place, Connect, or Select clears the missing-file callout.
