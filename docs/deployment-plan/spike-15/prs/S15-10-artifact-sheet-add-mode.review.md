# Review — S15-10 Adding an Artifact asks for its copy details

**Verdict:** (round 2 pending)
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- If the `artifact`-level suggestion read fails, no sheet opens and the page banner shows the error (D-2). The alternative, opening with no suggestions, would hide that the type suggests fields.
- A new Artifact's reply rows always go on the page, even when the draft was nil (D-7), so its edit sheet shows the suggestions at once.
- After a create whose file was refused, the sheet's baselines move to what the create wrote (D-7, D-16): a retry writes only what changed since, and Escape on an unchanged sheet closes it without a confirm.
- The file code for Add artifact and Add file becomes one `IngestFileChoice` value with one refusal mapping, and `IngestFileDropRow` binds it (D-11, D-13). The Add file dialog's code changes, not its behavior.
- Out, not planned: the board's ingest progress bar and the old Add dialog's frames (SA-7 and the inventory ship today's drop row). The cover reload after an add goes to `docs/ideas/structural-debt.md` (M-9).

## Round 1

### X-1 · blocker · STEP-3 deletes `artifact` while its readers still compile against it
**Answer:** Fixed. STEP-3 now edits every reader of `artifact` (`workspaceChanged`, the modifier copy, `sheetOpensFromTheWorkspace`), so the app target builds at the step's check.

### X-2 · major · The § 4 Tests row overstates the `sheet.artifact` reads
**Answer:** Fixed. § 4 names the one read (`sheetOpensFromTheWorkspace`, line 1701).

### X-3 · major · `pencilOpensTheSheet`'s drop half is not in the ported list
**Answer:** Fixed. T-10 names it and gives its new assertions.

### X-4 · blocker · STEP-2 leaves the intermediate shape of `applyDroppedURLs`/`setReject` undecided
**Answer:** Fixed. In STEP-2 the `.create` arms of all three stay as they are and the `.attach` arms go; STEP-4 deletes all four.

### X-5 · major · STEP-2 and T-1 disagree about the spreadsheet test
**Answer:** Fixed. It stays in STEP-2, STEP-4 deletes it, and T-1 replaces it.

### X-6 · major · The list of Add button callers is incomplete
**Answer:** Fixed. D-12 names the binding's setter and says `model.graphs.openAdd()` is untouched.

### X-7 · major · Choices about the File row left open
**Answer:** Fixed. D-13 says the File row sits outside `PVTabPair` and above it, is disabled while saving, uses the system Tab order, and stays enabled after a refused file.

### X-8 · major · The Description control and the Label hint change silently
**Answer:** Fixed. D-13 records both changes as deliberate (SA-7, "then the edit mode's fields").

### X-9 · major · Edge cases in the create Save
**Answer:** Fixed. D-7, D-8 and D-9 now state:
- Save is allowed while the graph list loads.
- The label is compared trimmed.
- Assignments follow each `try`.
- (b) returns true when the page is leaving.
- `isDirty` after a refused file follows D-7's moved baselines.

### X-10 · major · Some tests cannot be written from their row
**Answer:** Fixed. Each row now points to an existing test or helper to copy:
- T-4 reads the cache as `sheetSavePatchesTheCachedWorkspaceRows` does.
- T-1 uses `writeTempPNG()`.
- T-3 reads `model.pageError`.
- T-6 fills its value as `sheetRefusedValueLandsOnItsRow` does.
- T-11 follows `artifactGoneClosesTheSheet`.

### X-11 · minor · The brief draws four suggestions; the plan uses three
**Answer:** Fixed. § 8 says the fixture uses three; the board's fourth (Image number) is not needed.

### X-12 · minor · Smaller choices
**Answer:** Fixed:
- `sheetSaved()` builds the shared `onSaved` closure (D-12).
- `isOpeningAdd` is reset with `defer`.
- The fake's `createArtifact` becomes `try withState`.
- STEP-5's in-app check says how to report it when it was not run.
- The a11y id `rg` is in STEP-5.

### M-1 · blocker · `saveCreate()` copies the edit Save
**Answer:** Fixed. D-7 is one `save()`: the opening checks once, a switch around the write only, as `SourceGraphFormModel.save()` does, and one catch (D-8). The update call is `writeUpdate(…)`, shared by edit and the retry.

### M-2 · blocker · `Sent` is a third baseline beside `prefill` and the host's `opened`
**Answer:** Fixed, with one change to the suggested shape. `Sent` is gone. After the create, `prefill` and the host's baseline move forward: `MetadataDraftHost.markSaved()` (D-16, T-13).

The reviewer's version kept re-sending the dismissed ids. That is not needed: `sourcemetadata.apply` writes layout inserts and updates only, never deletes (`core/database/sourcemetadata/sourcemetadata.go`, the `apply` doc). So `markSaved()` clears `dismissed`, and the dismissed row the create wrote stays dismissed.

### M-3 · major · Put the suggestion read in one helper, since S15-11 needs it for `.graph`
**Answer:** Fixed. D-2 adds `SourcePageContext.unsavedMetadataRows(level:)`.

### M-4 · major · Both file rows wire `IngestFileChoice` into `IngestFileDropRow` by hand
**Answer:** Fixed. D-13 gives the row a `Binding<IngestFileChoice>` init that draws the refusal Callout; it replaces the six-parameter init.

### M-5 · major · The engine refusal → Callout mapping stays copied twice
**Answer:** Fixed. D-11 adds `IngestFileChoice.refuse(ifIngestRefusal:)`; both callers use it, and T-1 tests it.

### M-6 · minor · Per-mode copy belongs in the modifier
**Answer:** Fixed. D-10 switches on `model?.mode` in `ArtifactSheetModifier.copy`, as the graph form does. The model gains only `editedRef` and `isCreate`.

### M-7 · minor · One per-mode "links" test instead of branches in each graph property
**Answer:** Fixed. D-6 adds `links(_:)`.

### M-8 · minor · Comments describing the old dialog are missing from D-14
**Answer:** Fixed. D-14 lists them, and STEP-6 adds the `rg`.

### M-9 · minor · The add still reloads the whole workspace for the cover
**Answer:** Fixed by logging it. The Out row's home is now a `docs/ideas/structural-debt.md` row, added in STEP-8.

### M-10 · nit · `addArtifactDialogTitle` keeps "Dialog" in its name
**Answer:** Logged. The name stays: a rename would churn the string catalog for no change in copy.
