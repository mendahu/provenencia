# Review — S15-36 A metadata board can hold a draft

**Verdict:** round 1: Fits with changes — Ready with changes. Round 2: pending.
**Rounds:** 1 (round 2 due: round 1 had blockers). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- The Swift draft type is `CatalogMetadataDraft`, not the map's `MetadataDraft`: the store's other draft is `CatalogObservationDraft`. The map's S15-36 scope entry now uses the new name.
- The four store writes take a required `metadata:` parameter, and every caller states `nil`. There are no overloads.
- A draft sends nothing when it is unchanged (`draft` is nil). When it has changed, it sends every visible row's layout plus the rows dismissed in the draft.
- A draft that re-adds a field dismissed before the form opened lifts that dismissal on Save. A live board keeps the dismissal.
- Confirming the add row closes it, as the Add dialog closes on save today.
- `showRefusal` opens the refused row's edit, as a live board leaves a refused row in edit.
- No Swift test of `GoStore.mapMetadataDraft`: `swift-tests.mdc` rules out generated protobuf tests.

## Round 1

### X-1 · blocker · "Reorder, then reorder back, is unchanged" cannot hold with `moveSaved` and D-7's `isChanged`
`moveSaved` always rebuilds the rows as saved rows then suggestions. The opened order puts the empty suggestion Provider first, so dragging back never restores it, and `isChanged` stays true.
**Answer:** Fixed. D-8 (was D-7) now compares the saved rows' `(id, value)` sequence and the empty suggestions' id sequence separately, plus the dismissed list. T-6 states both moves and the expected rows; the Edges section notes the assumption.

### X-2 · major · T-1 calls `FakeStore.updateArtifact` with no Artifact seeded
**Answer:** Fixed. T-1 seeds `artifactsBySource["s1"]` with `catalogArtifact("a1", label: "Page 1")` and names every argument. D-4 records the draft before the hold or any error.

### X-3 · major · D-10 leaves add-row behaviours open
These were: error clearing, the row closing during a confirm, `isSavingAddRow` on each outcome, re-opening an open row, and a stale picked field.
**Answer:** Fixed in D-11 (was D-10):
- `isSavingAddRow` is reset by `defer` on every outcome.
- `closeAddRow` does nothing while saving.
- The outcome writes use `addRow?`.
- An open row keeps its field, value and error when `openAddRow` is called again.
- The row's error clears only on a new row and at the next confirm; clearing it on edit is S15-34's view (an Out line).
- A picked field that the reset put on the board is cleared, and `canConfirmAddRow` then gates.

### X-4 · major · `showRefusal` and another row in edit
**Answer:** Fixed. D-13 runs `beginEdit(fieldID:)`, which cancels another row's edit and resets its buffer. `placeRowError` replaces any earlier row error. An empty id matches no row. T-10 starts with Roll in edit.

### X-5 · major · D-6 silent on three draft-host cases
**Answer:** Fixed in D-7 / D-6:
- `clear` of an id with no row is a no-op, and a dismissed field has no row.
- `reordered(by:)` skips repeats.
- `k` counts from 0, in dismissed order.

T-2 and T-6 cover them.

### X-6 · minor · Live-board tests need `TestMetadataBoardHost.fields`
**Answer:** Fixed. T-8 and T-9 set `fields = [film]` and give the scripted errors and entries.

### X-7 · minor · Which T-8 half goes in which file
**Answer:** Fixed. The draft half is in `MetadataDraftHostTests`, the live half in `MetadataBoardModelTests`.

### X-8 · minor · The Board model row reads as if `add` trims
**Answer:** Fixed. §4 now says `add` does not trim and `createFromAdd` does.

### X-9 · minor · Test-site counts
**Answer:** Confirmed by the reviewer; no change.

### X-10 · check · No `xcodebuild` in the review environment
**Answer:** Noted. The implementer runs on macOS (the **Runs on** line).

### M-1 · blocker · `showRefusal` adds a second way to turn a metadata refusal into a row error
**Answer:** Fixed. In D-13, `rowRefusalMessage` (the code check plus `host.failureMessage`) and `placeRowError` are shared by `classifySetFailure`/`save` and `showRefusal`. `showRefusal` adds only the param lookup and `beginEdit`.

### M-2 · major · A case match on `.coded` against `CoreInvokeError`'s one-read rule
**Answer:** Fixed. D-5 adds `CoreInvokeError.params` and moves `SourceArtifactsSection`'s two ingest matches onto it. `OnboardingView`'s match reads the kind and stays (an Out line).

### M-3 · major · The add row duplicates the Add dialog's rule
**Answer:** Fixed. In D-12 the dialog presents its target's add row: computed `addFieldID`, `addValue`, `addValueError` and `isSavingAdd`, and `createFromAdd` keeps its two required messages and then calls `confirmAddRow`. Every member name stays, so the views and the existing dialog tests are unchanged (T-12). S15-34 only deletes.

### M-4 · major · The draft host's reorder copies `FakeStore.reorderMetadata`, which leaves rows in undefined order
**Answer:** Fixed. D-6 adds `[CatalogMetadataEntry].reordered(by:)`, used by both, with a stable order for the rows the ids leave out. D-8's layout uses the same index rule.

### M-5 · minor · The no-confirm clear lives on the page-level dialog object
**Answer:** Fixed. D-10 adds `MetadataBoardModel.requestClear(fieldID:) async -> Bool`, and `askClear` stays synchronous as the confirm only.

### M-6 · minor · `unsavedRows` puts a store read on a host that holds no store
**Answer:** Fixed. D-9 makes it the pure `unsavedRows(from:)`, and the form makes the read (S15-10, S15-11). T-7 needs no `FakeStore`.

### M-7 · minor · The map row and an `add-workspace-place` line no longer hold
**Answer:** Fixed. STEP-5 rewrites the map's S15-36 Ships text and the skill's "Metadata values" line.

### M-8 · nit · Doc comments made stale
**Answer:** Fixed. **Deletes** lists the four rewordings, and STEP-3 makes them.

## Round 2

The maintainability reviewer (Opus, fresh context) found that every round-1 answer holds against the code, with the notes below.

### M-9 · blocker · The dialog's `canSubmitAdd` keeps its own copy of the add row's confirm gate
**Answer:** Fixed. In D-12, `canSubmitAdd` is now `addTarget?.canConfirmAddRow ?? false`, and `addSheetIsDirty` keeps its body.

### Notes on round-1 answers
- M-2: D-5's "Why" pointed at D-12 instead of D-13. **Fixed.** The `.coded` matches inside `CoreInvokeError.errorDescription` and `L10n.Errors.message(for:)` predate this PR and are no worse after it. **Logged.**
- M-4: `FakeStore.reorderMetadata` now uses the production `reordered(by:)` instead of its own local sort. The reviewer judged this no worse than today. **Logged.**
- M-6: the map's Ships line still said the starting rows come "from `listTypeSuggestions`". **Fixed:** STEP-5 rewords it too.
