# Review — S12-02 Copy Source form

**Verdict:** Fits with changes — Not ready → fixed (round 1); round 2 after the board (below).
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet). The maintainability reviewer also read brief S12-D1.

## To user
None open.

## Answered by the user
- **Board delivery (2026-10-18).** The board is the Claude Design link on the plan's **Board** line; §6 macOS describes every frame in words for an implementer who cannot open it.
- **Sidebar count after a copy from the Source page.** Out line to S12-16 stands; S12-17's list host republishes it through its existing status hook.
- **Cancel on a changed form (2026-10-18).** Match the kit: Cancel closes; an outside click asks the kit's "Discard changes?". The board's "Discard this copy?" is not built; brief CS-10 corrected (it wrongly said Add Source asks on Cancel).
- **Title error wording (2026-10-18).** Keep Add Source's strings (one validation, working rule). The board's "Enter a title for the source" is not used. Copy is disabled while the title is empty, as the board shows.
- **One brief per view, one view per PR (2026-10-18).** The user pointed out that the form touched two views (Source page, Sources list), so it needs two briefs. Asked whether the PR should split too (the `plan-pr` size rule), the user chose to split. S12-02 keeps the form and the Source page entry (brief S12-D1). The Sources list row menu moves to the new S12-17 (brief S12-D6), which depends on S12-02. The form's loading and load-error states stay in S12-D1 / S12-02 because they belong to the dialog. Review findings M-9 and the row-menu parts of D-17 / X-6 now land in S12-17's plan.

## Writer's own calls
- No toast after Copy; opening the copy is the confirmation (D-11).
- Absent parts (no description, metadata or credibility) show disabled with "None on this Source" rather than hidden (D-8, brief CS-5).
- An All Artifacts tri-state checkbox above the per-Artifact rows (D-16, brief CS-6).
- The Copy action on the Source page is an icon button before Delete; the list row's menu has only Copy (D-17).
- The title prefill is focused, not selected (D-15).
- The label rule moves to the trimmed form, so a whitespace-only label shows the ref everywhere (D-1).
- The plan was written against the S12-01 branch head, then refreshed on `main` @ `817bad1` once #478 merged; no §4 file changed.

## Round 1

### M-1 · blocker · A fourth "Artifact label, else ref" rule
Three copies exist (`SourcePageArtifactsView` L243, `SourcePageModel.askDeleteArtifact`, `EvidenceGraphFiles.fileBearing`, the last trimmed). Better shape: `CatalogArtifact.displayTitle` beside `openableRelPath`.
**Answer:** Fixed — D-1 adds `displayTitle` and `hasDisplayLabel` and moves the three callers; STEP-1, T-1.

### M-2 · major · Fileless-icon helper in the wrong file
**Answer:** Fixed — D-2 puts `hasFile` and `thumbnailTypeIconKey(sourceTypeIconKey:)` on `CatalogArtifact` in `GenealogyStore.swift`; the private copies in `SourcePageArtifactsView` go.

### M-3 · major · Source type plumbing moved once, copies left
**Answer:** Fixed — D-4 adds `SourceTypeComboRow` with `options(for:)`, used by both `typeComboOptions` and the Source page header's type row; D-5 takes only `types`; D-6 warms `.sourceTypesList` in the model. Brief inventory names both rows.

### M-4 · major · View-seeded selection
**Answer:** Fixed — D-7 / D-9 store `excludedArtifactIDs`; `artifactIDsToCopy` is the live Artifacts minus exclusions. No seeding hook.

### M-5 · major · An eighth copy of the query-wait helper
**Answer:** Fixed — D-19 adds `macos/ProvenenciaTests/QueryHandleWaiting.swift`; `SourcesModelTests` and `SourceCopyModelTests` use it; the other five suites are logged in `docs/ideas/shared-list-helpers.md` (STEP-8).

### M-6 · minor · Validation named after the field it fills
**Answer:** Fixed — D-3 `Draft.validation: SourceDraftValidation` and `trimmedTitle`; `description` gets a default.

### M-7 · minor · `isDirty` by hand
**Answer:** Fixed — D-7 `SourceCopyChoices: Equatable`; `isDirty = choices != prefill`.

### M-8 · minor · After-copy closure per host; hand-built Add Source location
**Answer:** Fixed — D-13 reads `WorkspaceNavigation` from the environment and owns close + open; D-20 moves Add Source onto `pageLocation(for:)`.

### M-9 · minor · Fourth hand-wired right-click menu
**Answer:** Partly fixed — D-17 puts the trigger on `pageZone` so S12-14's graph sub-rows do not inherit it; the `pvActionMenu(items:)` recipe is logged (Out, STEP-8). Declined: a shared row-actions builder for the two call sites — each `onCopy` is one line, and `SourcesSplitRowProgress` already duplicates `onOpenPage` / `onOpenGraph` outside this PR's scope.

### M-10 · minor · Brief names two thumbnail pieces
**Answer:** Fixed — brief inventory names only `CachedThumbnail` at 32pt.

### M-11 · nit · String key and symbol names
**Answer:** Fixed — D-21, §6 Strings table with a `copySource…` symbol prefix.

### X-1 · blocker · Not Planned: board absent, strings provisional
**Answer:** Declined as a plan defect — this is the intended `Blocked — on board S12-D1` state (`plan-pr` §3). It goes to the user as the board-delivery question.

### X-2 · blocker · Xcode checks cannot run on Linux
**Answer:** Fixed — D-22 and a Stop condition: the Xcode build, tests and manual check run on the macOS dev machine, as for S12-01 (D-16 there; its ship log worktree is on macOS).

### X-3 · major · Current state rows not exactly true
**Answer:** Fixed — line numbers corrected (`create()` L368, `copySource` from L988), `Draft` and the stored errors described as they are, `sourceTypeComboRow` noted as reading `model.types`, the base line is now `main` @ `817bad1` after #478 merged.

### X-4 · blocker · Undefined members and names
**Answer:** Fixed — D-3 (validation shape, `description` default), D-6 (every `SourceCopyModel` member), D-6 / D-4 (types handle in the model), D-13 (who clears `copying`, cancel while saving), D-15 (title focus), Out line for the sidebar count.

### X-5 · major · Step and test mapping
**Answer:** Fixed — T-12 covers `openCopy` in STEP-3; the deleted-Artifact test is T-11 in STEP-3; strings split by step; XT command quoted; T-9 asserts the list contents only; T-11 names the key and the wait.

### X-6 · major · UI choices left open
**Answer:** Fixed — D-14 (checkbox init, hints, absent wording, `PVField` group headings), D-16 (`CachedThumbnail` arguments, `displayTitle`), D-2 (icon key at the page call site), D-17 (menu item action, modifier on `pageZone`, accessibility action on the page zone button).

### X-7 · minor · Smaller gaps
**Answer:** Fixed — dirty uses `choices != prefill` (an excluded id that is later deleted still counts as a change, which is acceptable); `openCopy()` uses `context`; STEP-8 says how the README lists the archived brief; the manual check's cover line relies on the engine (S12-01 D-10), not the fake.
