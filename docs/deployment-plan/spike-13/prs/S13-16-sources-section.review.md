# Review — S13-16 Sources section

**Verdict:** Fits with changes — Not ready (round 1, on the draft; every finding Fixed)
**Rounds:** 1 (round 2 runs when board S13-D6 is agreed and the plan is refreshed). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- Should a graph opened from a card also be scrolled to the Subject? The maintainability reviewer showed this needs no canvas rework (M-12). **Answered 2026-11-05: yes.** It is now D-13 and T-8.

## Writer's own calls

- D-1: `ConclusionPageSection.queryKey` becomes `queryKeys`, a list headed by the section's own read. The place warms every key in it.
- D-5: a claim whose Subject type cannot be resolved is drawn as a person card with an empty type label. It is never hidden.
- D-7: the card's VoiceOver label is the graph card's own label plus ", on <graph>".
- D-13: the canvas centers the opening Subject through a new `openingFocus` on its opening camera. The unused pending-selection hook is deleted.
- The SS-1, SS-3, SS-4, SS-6, SS-8 and SS-W1 answers stay open until the board is agreed (Open questions 1).

## Round 1 — maintainability

### M-1 · blocker · D-13 adds a second "select this Subject when the graph opens" path
`WorkspaceSession.pendingGraphSelections`, `setPendingGraphSelection` and `consumePendingGraphSelection` already exist, and the setter has no caller.
**Answer:** Fixed. D-13 now drives the selection from the location's `subjectId` at the point where the model consumes the pending selection today. The hook is deleted (Deletes, STEP-8).

### M-2 · blocker · `claimCard` is a second placed-Subject builder and hard-codes `.person`
**Answer:** Fixed. D-4 extracts one per-Subject initializer from `build`'s loop, and the sort stays private. D-5 resolves the kind from the type. `.person` is only the fallback for an unresolved type, and that Stop condition is gone.

### M-3 · blocker · `CatalogSourceTypePill` already shows a Source's type
**Answer:** Fixed. D-10 uses it, and so does the brief's inventory.

### M-4 · major · The Timeline's title link already exists
**Answer:** Fixed. D-10 copies it, and the brief names that one control.

### M-5 · major · Graph locations carry the Source's ref and title
**Answer:** Fixed. D-6 builds the location through a `SourcesListNavigation.canvasLocation(sourceId:sourceRef:sourceTitle:graphId:subjectId:)` overload.

### M-6 · major · A second warm path in the view
**Answer:** Fixed. D-1 adds `queryKeys`, and the place warms the supporting lists.

### M-7 · major · A second way to vary the card's paint
**Answer:** Fixed. D-2 adds `Chrome { canvas, ghost, page }`, folding `isGhost` into it. T-4 now checks the paint and the regions, not header hit ids that were never regions.

### M-8 · major · The footer leaves no room for S13-17
**Answer:** Fixed. D-2 records the `handle` and `slot` parts of `promotedFooterParts`.

### M-9 · major · A second card VoiceOver label; the graph does not read rows
**Answer:** Fixed. D-7 changes `EvidenceCardText.accessibilityLabel`. The brief's SS-9 drops the per-row reading, and that is now an Out line.

### M-10 · major · A third copy of the source type label lookup
**Answer:** Fixed. D-8 adds `label(forTypeID:)` and moves both copies onto it.

### M-11 · minor · `EvidenceSubjectCardImage` is not an image
**Answer:** Fixed. Renamed `EvidenceCardViewHost`.

### M-12 · minor · "No reveal API" is not accurate
**Answer:** Fixed. The question went to the user, and the reveal is in scope (D-13).

## Round 1 — executability

### X-1 · blocker · The title cannot be a `PVButton` inside `PVSectionHeader`
**Answer:** Fixed. D-10 drops `PVSectionHeader` and uses the Timeline title link with the heading trait.

### X-2 · blocker · The pending-selection hook is ignored
**Answer:** Fixed, the same fix as M-1.

### X-3 · major · `PVChip` is a button
**Answer:** Fixed. The type is shown with `CatalogSourceTypePill`, a label (M-3).

### X-4 · major · How the host gets the layout and footer frames; footer paint
**Answer:** Fixed. D-3 adds a static `EvidenceCardView.layout(for:)` and overlays controls at the region rects. D-2 says `.page` paints the band and the hairline, with no handle text and no icon.

### X-5 · major · The lists in `make`, nil snapshot, `cardinalityByKey`, `EvidenceCardText(subject:orders:)`
**Answer:** Fixed in D-5 (nil snapshot falls back and redraws; `PropertiesSnapshot.cardinalityByKey`), D-7 (`EvidenceCardText(subject:orders:)`) and D-9 (`make` reads the handles).

### X-6 · major · `observationDisplayOrder` is already static
**Answer:** Fixed. D-4 keeps it private in its file.

### X-7 · major · `kind` and the type field
**Answer:** Fixed. D-5 names `subject.subjectTypeID` and resolves the kind.

### X-8 · major · Location fields when the ref or name is empty
**Answer:** Fixed in D-6 and D-7. The Source's ref and title fill the location. A graph with no name shows its ref (AC-2.3).

### X-9 · minor · Heading accessibility key, plural forms, footer height
**Answer:** Fixed. `sourcesGroupAccessibility` and its untyped variant are added, the plural `one` and `other` forms are given, and the footer is the graph's 36pt band (D-2).

### X-10 · minor · Review link and the map's Depends
**Answer:** Fixed. This log now exists, and the map row lists S13-13.

### X-11 · minor · Check quality and test order
**Answer:** Fixed. T-4 compares the painted output, T-8 covers the opening focus, and the test IDs are in order.

## Round 2 — the agreed board (Source rows), 2026-11-05

The plan was rewritten after the board was agreed, and fresh reviewers checked all of it. **Verdict:** Fits with changes / Ready with changes. Every finding is Fixed, except X-26, which is Declined with evidence. No blocker is open, and the plan is `Planned`.

### Maintainability

- **M-21 · blocker · The link style stays buried in a recipe's row file.** Fixed. The provenance row's link moves into a kit component, `PVInkLink` (`DesignSystem/Components/InkLink/`). The record row and the Sources title both use it (D-7, STEP-5).
- **M-22 · blocker · The title cell re-derives `PVTitleText`'s face.** Fixed. A titled Source is `.name`, an untitled one `.ref`, drawn with `PVTitleText(.row)` (D-3, D-7). A `PVTitleText` that fixes its own ink is a Stop condition.
- **M-23 · major · A fake placed Subject and a warmed key, only to compute a constant.** Fixed. A static `displayName(subject:kind:observations:eventTitle:orders:)` replaces it (D-2). `.propertiesWorkspace` and the `properties:` input are gone.
- **M-24 · major · A third hand-written fold.** Fixed. One `ConclusionFold.shown`; Family and Timeline move onto it (D-10, T-9).
- **M-25 · major · `queryKeys` with an unwritten "first is own" rule.** Fixed. `queryKey` stays as it is, and `supportingQueryKeys` is added (D-1).
- **M-26 · minor · The closest column precedent is missing.** Fixed. D-4 copies `SourcesListColumns`' per-column cell modifiers.
- **M-27 · minor · The page location is built inline.** Fixed. A `pageLocation(sourceId:ref:title:)` overload is added (D-5).
- **M-28 · minor · The `subjectId` doc comment becomes untrue.** Fixed (D-11, STEP-7).

### Executability

- **X-21 · major · A test file name is wrong.** Fixed: it is `ProvenanceSheetTests.swift`.
- **X-22 · major · Where the init goes.** Moot: there is no new initializer (M-23).
- **X-23 · major · Who owns `openingFocus`, and in which space.** Fixed. The model's grid-space value uses the constant frame. The view converts it with `documentPoint(_:origin:)` (D-11).
- **X-24 · major · Call sites and model reuse.** Fixed. Both `applyOpeningCamera` call sites pass the focus, and the model is per mount (D-11, Reach).
- **X-25 · major · T-8 has no harness.** Fixed. The offset is now the pure function `GraphCanvasCoordinates.contentOffset(centering:…)`, tested in `GraphCanvasCoordinatesTests`.
- **X-26 · major · `PVButton(.link, .sm)` breaks caption size.** Declined. `PVButton.swift`'s `.link` branch drops the padding, the fixed height and the chrome, and draws `size.buttonFont` (caption medium at `.sm`) underlined in link ink. That is the board's graph link.
- **X-27 · major · Nil lists, duplicates, `Equatable`.** Fixed in D-3.
- **X-28 · major · Strings and ids.** Fixed. The catalog key names are given (D-12, Strings), and the accessibility label uses `sourcesDateNotRecorded`.
- **X-29 · major · Checks that cannot run here.** Fixed where the plan can fix it. The FakeStore error case is named (manual item 3). `xcodebuild` and `check-pr.sh` run on a Mac, as for every Swift PR.
