# Review — S12-12 Graph form

**Verdict:** Fits with changes — Ready with changes (all applied)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet). Both read the plan and the S12-D3 brief.

## To user

- None. The three product calls were answered before the plan was written (2026-10-25): lock in-use Artifacts when Edit opens, stay on the Source page after Add, one `last_graph` body.

## Writer's own calls

- D-1, D-2: the checklist becomes `PVArtifactChecklist` under `DesignSystem/Recipes/ArtifactChecklist/`, with its strings in a new `L10n.ArtifactChecklist` namespace.
- D-3 (after M-3): a locked row uses the kit's existing lock mark (`PVCheckboxMark.isLocked`), passed through a new `isLocked` parameter on the labelled `PVCheckbox`, and VoiceOver reads "locked".
- D-8: no toast after an edit; the after-add toast is "Evidence graph added" / "{ref} is ready to open."
- D-10: following a Citation in Kept linked closes the form without asking "Discard changes?", as the delete notice's rows do.
- D-11: Add evidence graph is a primary small button like Add artifact, and stays visible but disabled while the Source has no Artifact.
- D-13: Description is a three-line `PVTextArea`, where Add artifact uses a one-line `PVInput`.
- Copy in §6 Strings (titles, subtitle, hints, errors), to be settled by the board.

## Round 1

### M-1 · blocker · The Kept linked rows get their own copy of the close-then-navigate rule
**Plan:** D-10's closure set the binding to nil and then navigated "as `PVDeleteImpactControls.activate` does"; D-4 gave the list a bare `onActivate`.
**Codebase:** `PVDeleteImpact.swift` `PVDeleteImpactControls.activate(_:onNavigate:dismiss:then:)`.
**Better shape:** the list takes `onNavigate` and `dismiss` and calls `activate` itself.
**Answer:** Fixed. D-4 now has `PVDeleteImpactGroupList(title:group:onNavigate:dismiss:)` calling `activate`; D-10 passes `navigation.go` and `{ model = nil }`; the §6 macOS Kept linked row is updated.

### M-2 · blocker · A third place computes "page Artifacts minus unchecked"
**Plan:** D-6 computed `checkedArtifactIDs` apart from D-3's helper; `SourceCopyModel.artifactIDsToCopy` is a second copy.
**Better shape:** `PVArtifactChecklistState.checkedIDs(ids:excluded:locked:)`, called by both models.
**Answer:** Fixed. D-3 adds `checkedIDs`, and `checkedCount` derives from it. D-6 and `SourceCopyModel.artifactIDsToCopy` call it (STEP-1 files). T-1 gained a case. Codebase fit has a row for it.

### M-3 · major · Locked rows draw a dimmed check while the kit already has a lock state
**Codebase:** `PVCheckbox.swift` `PVCheckboxMark.isLocked`, used by `PropertiesView.swift:365` and `:603`.
**Answer:** Fixed. D-3 now uses `PVCheckbox(verbatim:description:isChecked:isLocked:)` with a new `isLocked` parameter that forwards to the mark and sets a new spoken value, `L10n.DesignSystem.checkboxLocked`. The brief's Row checkbox inventory row is now **Extend** and names the lock mark; GF-7, GF-15 and frame 7 are updated. A new stop condition covers a board that draws something else.

### M-4 · major · The edit form's loading and read-error area copies the Copy form's private one
**Codebase:** `SourceCopyDialog.swift` `SourceCopyForm.loading` and the `.failed` branch; `SourceCopyModel.swift` `SourceCopyLoadState`.
**Answer:** Fixed. New D-14: `SourceFormLoadState` (renamed) and `SourceFormLoadArea` *(new)* in `Features/Sources/`, called by both forms. STEP-2 moves the Copy form onto it with unchanged ids. The §6 macOS loading and error rows are merged. The brief's inventory row is now Extend.

### M-5 · minor · The "optional model, refuse to close while saving" presenter is copied
**Answer:** Logged. The S12-13 scope entry in `deployment-plan.md` now says to add `pvFormDialog(item:)` when its dialogs make the third such form. The plan's Out table points there. A kit change is not worth it for the second copy.

### M-6 · nit · "Read again" repeats `copySourceReadAgain`
**Answer:** Fixed with M-4. One key, `sources.form.readAgain` (`L10n.Sources.formReadAgain`), and `graphFormReadAgain` is dropped.

### M-7 · nit · `openEdit` repeats `askDeleteGraph`'s ownership check
**Answer:** Fixed. D-9 adds `SourceGraphsSection.ownGraph(id:)`, and `askDeleteGraph` calls it (STEP-4).

### X-1 · blocker · The plan is Blocked and §6 Frames is empty
**Answer:** Declined as a plan defect. `plan-pr` step 3 sets status `Blocked — on board S#-D#` with the brief written and the board not drawn. The brief's Handoff fills Frames and Strings and moves the plan to Planned. It is not handed to an implementer before then, as S12-11 was not.

### X-2 · major · The DoD contradicts the brief's state
**Answer:** Fixed. The DoD line now says the planner archives the brief at board agreement, before the PR starts, and the implementer only checks that it is there.

### X-3 · major · No source for `sourceTypeIconKey` or `projectDir`
**Answer:** Fixed. D-6 reads both from `SourcePageContext` (`projectDir`, `types` for the Source's type), as `SourceCopyModel.sourceTypeIconKey` does.

### X-4 · major · A locked box-only row shows no visible hint
**Answer:** Fixed. D-3 says a locked row always uses the labelled form, so `displayTitle` falls back to the ref and the hint is visible. Locked rows ignore `isDisabled`.

### X-5 · major · `PVArtifactChecklistState` signatures underspecified
**Answer:** Fixed. D-3 gives every signature and its formula, `liveExcluded` subtracts `locked`, and T-1 is "parameterized over the five cases in this row".

### X-6 · minor · Choices left to the implementer
**Answer:** Fixed. D-11 names the actions (`openAdd()`, `openEdit(id:)`). D-13 uses `activateOnAppear: true`, as `SourceTitleTypeFields`. D-7 says `init` sets `.ready` for a graph with no links. D-9 makes `openEdit` a no-op without a workspace. D-1 names the group id `D5C000000000000000000023`. D-2 places the enum before `enum LastUpdated`. STEP-3 says `try withState`.

### X-7 · minor · Checks that prove little
**Answer:** Fixed in part. STEP-3 now runs `SourcePageModelTests`, and STEP-6's app check opens both modes of the form. STEP-5 stays build-only because its view is first reachable in STEP-6, and the plan says so. The xcodebuild checks need macOS, as for every Swift PR.

## Round 2

Executability (sonnet) accepted the answers to X-2 – X-6 and the Out home for M-5, found every Current state row true, and raised:

### X-8 · blocker · T-6 and T-8 double-start the preflight
**Answer:** Fixed. §8 fixtures say tests never call `loadKeptLinks()` after `openEdit` starts it and wait with `waitUntil { … loadState != .loading }`. T-6 waits, and T-8 waits for `.failed`, clears the hook, then calls `loadKeptLinks()` as Read again would (2 calls).

### X-9 · major · Test fixtures are private to `SourcePageModelTests`
**Answer:** Fixed. The new test file copies `makeStore` / `makeModel` as private helpers, the repo's per-file convention. Artifacts are built as `SourceCopyModelTests.artifact(_:label:)` builds them, and §8 gives the unlink-block fixture through `PVDeleteImpactPreviewData.listed(ref:title:)`.

### X-10 · major · A locked row's VoiceOver value is overwritten by the ref value
**Answer:** Fixed. D-3 sets a locked row's value to `L10n.ArtifactChecklist.lockedValue(ref:)` ("Locked, %@"), and unlocked rows keep the ref. The key is listed in §6 Strings.

### X-11 · minor · STEP-2 lists a rename in `SourceCopyModelTests.swift` that does not exist
**Answer:** Fixed. The file was removed from STEP-2's list.

### X-12 · minor · `SourceGraphForm` has no wiring for the closures
**Answer:** Fixed. D-10: the modifier passes `onNavigate` and `dismiss` into `SourceGraphForm`'s init.

### X-13 · minor · Add button help and the loading-time reason
**Answer:** Fixed. D-11 uses `.help(reason ?? title)` as Open does. D-9 makes `addBlockedReason` nil while the workspace is loading.

### X-14 · nit · T-5's edit case setup
**Answer:** Fixed. "edit set up as T-6, with no blocks".

Maintainability (opus) confirmed that M-1, M-2, M-4 – M-7 landed correctly. On M-3 it noted two stale docs and the same value-override risk as X-10 (fixed above), and raised:

### M-8 · blocker · A second spoken word for the kit's lock mark beside the one Properties already uses
**Codebase:** `L10n.Properties.bindingLocked` ("Locked"), read only at `PropertiesView.swift:385` and `:625` to voice `PVCheckboxMark(isLocked: true)`.
**Answer:** Fixed. D-3: the kit key `designSystem.checkbox.locked` is "Locked" (the same word); both Properties readers move onto it, and `bindingLocked` and its xcstrings entry are deleted. STEP-1 lists `PropertiesView.swift`, and its `rg` check includes `Properties.bindingLocked`.

### M-3 follow-ups (Q10 docs)
**Answer:** Fixed. D-3 rewords the `PVCheckbox.swift` header comment and the README Checkbox row. STEP-2 adds `PVDeleteImpactGroupList` to the README DeleteImpact row.

## Verdict

No blocker is open after two rounds; the cap is reached. **Fits with changes — Ready with changes**, all applied. The plan stays `Blocked — on board S12-D3`. At board agreement the planner fills §6 Frames and Strings from the board, archives the brief and sets the plan to Planned.

## Board agreement (2026-10-25)

The writer rendered all thirteen S12-D3 frames from the exported board and checked them against the plan. Changes folded into the plan (§6 Frames and Strings, D-1, D-8, D-13, D-14, Out):
- The Add subtitle is the board's "An evidence graph is a discrete group of subjects and their relationships. …".
- A caption under the Artifacts heading ("Select all the artifacts that are relevant to this evidence graph.") becomes an optional `description` on `PVArtifactChecklist`. The Copy form passes none.
- Description carries `PVField`'s `.optional` marker.
- The Kept linked callout has a lock icon and no title, so `graphFormKeptLinkedTitle` is dropped.
- The read-error and save-error bodies follow the Copy form's pattern ("… are kept. … %@").
- The board's mono error "detail" box and its web lock stand-in are web-kit gaps, listed in Out.

No structural change, so the review is not rerun.
