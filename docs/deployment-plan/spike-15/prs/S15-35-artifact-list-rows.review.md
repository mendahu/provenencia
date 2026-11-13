# Review — S15-35 The Artifact list shows each Artifact's copy details

**Verdict:** round 1: Fits with changes (maintainability) — Not ready (executability); every round-1 finding answered; round 2 pending.
**Rounds:** 1 so far. **Reviewers:** maintainability (fresh, read-only), executability (fresh, read-only)

## To user

- **Person page pill (asked while planning).** The type pill also shows on the Person page's Sources rows, which no board draws. **Answer (user, 2026-11-13):** both change; one pill everywhere (D-2, AC-5.2).

## Writer's own calls

- The "+n" after the graph names is the kit's `PVBadge.moreCount` (a subtle outlined "+1"), not the board's plain text: the kit already draws every "+N" after a line and keeps it whole while the text truncates (M-2). A small visual difference from the board.
- A PDF or fileless Artifact's mark draws at the kit tile's size, not the board's 30pt glyph (D-11).
- The header keeps its own type icon lookup (`model.identity.types.iconKey(forTypeID: model.identity.sourceTypeID)`): it follows the identity's type while the type is being changed, which `SourcePageContext.sourceTypeIconKey` does not (M-4, the header half).
- The row has no `#Preview` (X-5): it needs a live `SourcePageModel`, and no Sources preview builds one.

## Round 1

### M-1 · blocker · `graphNames(for:)` is a second derivation of graph membership beside `inNoGraph`
**Answer:** Fixed — D-6: one `graphLines: [String: ArtifactGraphLine]` built once per render behind `inNoGraph`'s readiness (`artifactIDsInNoGraph`), using `CatalogEvidenceGraph.links`; the row takes one `ArtifactGraphLine?`. T-3 asserts the enum and that it agrees with `inNoGraph`. Extending `EvidenceGraphIndex` was not needed: the index holds ids per Source, and the graph form already reads `artifactIDsInNoGraph`.

### M-2 · blocker · "+N" drawn as text instead of `PVBadge.moreCount`
**Answer:** Fixed — D-7 uses `PVBadge.moreCount(more).layoutPriority(1)`.

### M-3 · blocker · Menu title as a localization key
**Answer:** Fixed — D-13: `pvRightClickMenu(verbatimTitle:)` and `PVContextMenuPanel.verbatimTitle`, drawn with `Text(verbatim:)`, as `PVFormDialogCopy` and `PVEmptyState` do; STEP-4.

### M-4 · major · A private copy of `sourceTypeIconKey`
**Answer:** Fixed for the Artifacts view — `SourceArtifactsSection.thumbnailTypeIconKey(for:)` over `SourcePageContext.sourceTypeIconKey`; the view's private copy goes (D-8, T-6). The header half: Declined — see Writer's own calls.

### M-5 · major · `.fixedSize()` inside `PVTag`
**Answer:** Fixed — D-1 drops it; D-4 fixes the Tag at the header call site; AC-5.2 and Edges say the Person page truncates, AC-5.3 that the header never cuts the type.

### M-6 · major · Hard-coded ", " in the spoken graph list
**Answer:** Fixed — D-7 uses `L10n.DesignSystem.spokenJoin(all)`.

### M-7 · minor · A closure per menu action
**Answer:** Fixed — `ArtifactRowAction` enum with `menuActions(canSetCover:)`; S15-25 adds `.move` (D-6, T-4).

### M-8 · minor · "First N, then a count" a third time
**Answer:** Declined — the split is `prefix(2)` inside `graphLines`; `ConclusionFold` is a Conclusions paging fold (pages of 50, a minimum hidden count), and `ArtifactCopyDetailsView.shown` also filters by value. A shared helper would be larger than the two lines it saves.

### M-9 · minor · "Card" wording left behind
**Answer:** Fixed — D-12 and STEP-6 list every comment named, the `CachedThumbnail` `.tile` asides (D-10), and the test rename; STEP-6's check `rg`s "artifact card".

### M-10 · nit · `@Bindable` and the unread constant
**Answer:** Fixed — `let model`; `size: Self.thumbnailSize` (D-7).

### X-1 · blocker · `EvidenceGraphView` passes `layout: .tile`
**Answer:** Fixed — § 4 Thumbnail row corrected; D-10 and STEP-5 drop the argument; the STEP-5 `rg` catches `layout: .tile`.

### X-2 · blocker · `arrange` with a nil, 0 or infinite width
**Answer:** Fixed — D-3 re-measures only when the width is finite and positive; nil reads as 0 as today; `count == 0` gives no frames; T-1 `idealProbeReMeasuresNothing`.

### X-3 · blocker · `.fixedSize()` on the Tag against the truncation claims
**Answer:** Fixed — with M-5.

### X-4 · major · `focusedPreviewID`
**Answer:** Fixed — it goes (no reader); D-7, D-8, the STEP-5 `rg`.

### X-5 · major · The row's `#Preview`
**Answer:** Fixed — no row preview (D-7, STEP-5); see Writer's own calls.

### X-6 · major · What feeds `canSetCover`; the PDF claim
**Answer:** Fixed — D-7 passes `model.artifacts.canUseAsThumbnail(art)`. The PDF claim holds: `core/derivatives` makes no thumbnail for a PDF (`thumbnail_test.go`, "a PDF is skipped"); § 4 says so.

### X-7 · major · T-3's expected values
**Answer:** Fixed — T-3 writes the exact arrays (FakeStore keeps input order) and builds the not-loaded case with no page location applied.

### X-8 · minor · T-4 and `LocalizedStringResource` titles
**Answer:** Fixed — T-4 asserts the enum and compares `L10n.string(action.title)`.

### X-9 · minor · STEP-4's file kind
**Answer:** Fixed — STEP-4 creates the `View` struct with an `EmptyView()` body.

### X-10 · minor · Doc steps
**Answer:** Fixed — the map row links the plan now; STEP-8 says which plan-doc warnings are expected.

### X-11 · minor · Not yet Planned
**Answer:** Fixed by this loop — status moves to Planned after round 2.
