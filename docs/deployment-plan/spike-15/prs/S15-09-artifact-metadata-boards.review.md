# Review — S15-09 Artifacts show their metadata on the Source page

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet, the implementer's model). No blocker left open after round 1, so no round 2 (review-plan step 5).

## To user
- None.

## Writer's own calls
- The board on a card keeps dismiss on suggestions and drag-to-reorder on saved rows, though the board frame draws neither (rule "One metadata board"; brief SD-9 "the same rows, suggestions and Add").
- Compact density: a 76pt label (as drawn); a suggestion row stacks its label above the input, Save and dismiss (the input would be about 30pt wide at the card's minimum otherwise).
- A board drops in-progress edits only when the incoming rows differ from what it shows (D-7, D-9), not when they differ from the last applied workspace.
- The board's chrome moves out of `SourcePageChrome.swift` under board names (S15-26 left that call to this PR); the `L10n.Sources.metadata*` keys keep their names.
- Out as "not planned": a metadata write on an owner deleted in another window keeps showing `sourcemetadata.invalid`.

## Round 1

### X-1 · blocker · STEP-6's check can never print nothing
`SourcePageMetadataView` and `SourcePageMetadataHost` stay and match the bare `SourcePageMetadata`.
**Answer:** Fixed — STEP-6's `rg` names the exact renamed symbols.

### X-2 · major · AC-3.1 says two layout changes; STEP-1's fixture makes one
**Answer:** Fixed — AC-3.1 lists the delete's changes without a count; T-1 says each case has exactly one value and one layout change (one `Set`).

### X-3 · minor · Removing the owner's delete shifts the index
**Answer:** Fixed — T-1 filters, then loops over the filtered slice's metadata changes and expects an empty result and no error.

### X-4 · minor · D-5's guard ignores the optional mutation
**Answer:** Fixed — the pattern is `…? = mutation`.

### X-5 · minor · Card view choices (binding, style plumbing, suggestion row lines, Add alignment)
**Answer:** Fixed — D-12 spells the binding; D-10 says the input, Save and dismiss share one line under the label, the Add is leading-aligned, and `style` reaches the suggestion row and `style.labelWidth` the label.

### X-6 · minor · T-6 and T-10 do not name the reloading write
**Answer:** Fixed — `await model.notes.add()`, then `settleWrites()` and `sync(from:)`.

### X-7 · minor · Missing key vs `[]`; `refreshCoverFromStore` sets the workspace before `apply`
**Answer:** Fixed — §6 Edges "Paths that set the workspace first" accepts both, with the reason.

### X-8 · minor · T-8 and the delete's one extra workspace read
**Answer:** Fixed — T-8 keeps that count unchanged. (D-9's hand edit is gone, M-3.)

### M-1 · blocker · Two reset rules that bypass D-2's owner→rows mapping
**Answer:** Fixed — D-9 is now `MetadataBoardModel.resetIfChanged(from:to:)` (precedent `SourceIdentitySection.resetIfChanged`); D-7 and D-8 call it with `metadataRows(for:)`, so D-2 is the only reader of the row fields and **Next adds**' "a `.graph` case to D-2 only" holds.

### M-2 · major · A bare `page / card` enum branched on in the view
**Answer:** Fixed — D-10's `MetadataBoardStyle { regular, compact }` carries its measures as properties (as `PVListDensity`); D-11's per-style label widths fold into it.

### M-3 · minor · D-9 hand-edited the cached workspace on an Artifact delete
**Answer:** Fixed — dropped; the `refreshCoverFromStore` read that follows the delete replaces the workspace, and `syncBoards` drops the board by Artifact membership.

### M-4 · minor · `copyDetailsTitle` puts copy choice in the section model
**Answer:** Fixed — D-13's `L10n.Sources.artifactCopyDetails(count:)`; the card passes `board.saved.count`.

### M-5 · minor · The hook is a third copy of the one-change walk; the skill does not learn it
**Answer:** Fixed — STEP-1 moves `TestScopeThroughDeletedParent` onto the hook (which takes a `querier`) and adds one sentence to `add-audit-scope`.

### M-6 · minor · `add-workspace-place` line 175 lists the riding mutations
**Answer:** Fixed — STEP-9 updates line 175; the DoD sweep checks every line that lists page mutations.

### M-7 · minor · `reorderHandleWidth` copies a kit constant
**Answer:** Fixed — D-11 adds `PVReorderHandle.width` and deletes the copy.

### M-8 · nit · `CatalogQueryPatch` lives in `CatalogLastUpdated.swift`
Logged; not answered.
