# Review — S12-13 Artifact graph checkboxes

**Verdict:** Fits with changes — Ready with changes (round 1; round 2 pending)
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet). Both read the plan; the maintainability reviewer also read the S12-D4 brief.

## To user

- None. The board's save-refused copy was settled with the user before review (2026-10-25): the body reads "Some changes may not be saved. Every field and box keeps its value — save again to finish."

## Writer's own calls

- D-12: the graph list's scroll cap is the kit checklist's 264pt, not the board's 280pt, so the two checklists on the Source page are one size (M-7).
- D-8: each successful link or unlink patches the card's snapshot in place, and after a refusal the locks are read again over it, keeping the researcher's checkboxes (X-4, X-12).
- D-9: the canvas waits for the graph list in `prepare()` so a graph with no links disarms the tools on open (X-5).
- §10: the manual check uses a project built from scratch, and the composer's LK-12 copy is checked by test, because the composer cannot be opened on a graph with no links.

## Round 1

### M-1 · blocker · Link and unlink writes go through `applyWrite`, which reloads the Source page payload on every write
**Plan:** D-8 applies every link and unlink through `context.applyWrite`.
**Codebase:** `SourcePageContext.applyWrite(_:)` = `session.apply(effects, .mutatedSourceWorkspace(sourceId:))`, which reloads `GetSourceWorkspace`; `SourceGraphFormModel.save()` uses `context.session.apply(effects)`; `add-workspace-place` §6.
**Better shape:** `session.apply(effects)` for link and unlink; `applyWrite` only for `updateArtifact`; one line in `add-workspace-place` §6.
**Answer:** Fixed — D-8 and Builds on; STEP-8 adds the skill line.

### M-2 · major · Two representations of one checklist, and hand-written arithmetic beside `PVArtifactChecklistState`
**Plan:** D-2 (excluded set, hand filter), D-5 (`checked` / `prefill`), D-12 (binds `checked`).
**Codebase:** `PVArtifactChecklist.swift` `PVArtifactChecklistState`; callers `SourceGraphFormModel.checkedArtifactIDs`, `SourceCopyModel`.
**Better shape:** One excluded-set representation on both surfaces through the helper, renamed `PVChecklistState`.
**Answer:** Fixed — D-2, D-5, D-7, D-12; Codebase fit rename row; STEP-3 files.

### M-3 · major · "A graph links an Artifact" becomes its third hand-written copy
**Codebase:** `SourcesListNavigation.graphBlockedReason(for:)`, `graphLocation(for:graph:)`.
**Answer:** Fixed — `CatalogEvidenceGraph.hasArtifact` in D-9, read by both sites and `canCite`; STEP-1; Next adds names S12-14.

### M-4 · major · New tuple-array call hooks beside FakeStore's ordered `recordedCalls`
**Codebase:** `FakeStore.recordedCalls: [String]`.
**Answer:** Fixed — link and unlink append to `recordedCalls`; T-8 asserts the order; the preflight keeps a graph-ids array (reads stay out of `recordedCalls`).

### M-5 · major · The session index depends on a Sources-feature helper, and leaves a wrapper and a dead helper
**Codebase:** `SourcesListNavigation.graphsBySource(_:)`; `SourcesModel.graphsBySource`.
**Answer:** Fixed — D-10 groups in the builder and deletes the helper and the wrapper; `graphsBySourceKeepsListOrder` moves into T-1; STEP-1 files and check.

### M-6 · minor · Error-code string matches spread to a second model
**Codebase:** `CoreInvokeError.coded`; `SourceGraphFormModel.save()`.
**Answer:** Fixed — `CoreInvokeError.code` (D-8, STEP-2); the graph form moves onto it. Named constants for the two codes are not added: two call sites, and the codes are already the `L10n.Errors` keys.

### M-7 · minor · `SourceArtifactGraphLinks` re-solves the checklist frame, and adds a one-caller Bool
**Codebase:** `PVArtifactChecklist` frame, `maxRowsHeight = 264`, locked `accessibilityValue`.
**Answer:** Fixed — D-12 copies the frame and the 264 cap, names the near-copy in Codebase fit, drops `showsNeedsGraph` (D-3: `addForm` places the caption).

### M-8 · minor · Codebase fit and Deletes miss some adds and stale text
**Answer:** Fixed — `visibleGraphs`, `cardSaveError`, `linkLoadTask` rows; `noArtifactAction` and the `SourceFormLoadArea` doc comment in Deletes; STEP-5.

### M-9 · minor · The Swift "only artifact" lock is a second client-side copy of the engine's last-link invariant
**Answer:** Logged — §3 Out row. When the unlink preflight next changes, it can report the last-link block so the graph form and the card read one engine answer.

### X-1 · blocker · T-7 cannot be written: the fake store's unlink preflight ignores the graph
**Codebase:** `FakeStore.getEvidenceGraphUnlinkImpact` returns `unlinkImpactBlocks.first { $0.artifactID == id }` for any graph.
**Answer:** Fixed — `unlinkImpactBlocksByGraph` replaces `unlinkImpactBlocks`; the five existing uses move (Codebase fit, STEP-2, §4 row).

### X-2 · blocker · How tests wait for the preflight that `beginEdit` starts in a `Task`
**Answer:** Fixed — D-6: `linkLoadTask` (tests await its value), Read again calls `loadLinkLocks()` directly, and a result for a cancelled or switched card is dropped by an artifact-id check.

### X-3 · major · `saveFields` early return and ordering are not specified
**Codebase:** `saveFields` returns when label and description are unchanged.
**Answer:** Fixed — D-8: the early return covers fields and links; `clearPageError()` stays; `savingID` covers every write; `replace(_:effects:)` for `updateArtifact` only.

### X-4 · major · After a partial failure, `graphs`, `prefill` and `locked` are not said to refresh
**Answer:** Fixed — D-8 patches the snapshot after each successful write and re-reads the locks after any error; D-6 says `loadLinkLocks()` reads that snapshot; T-10 asserts the patched prefill. (Round 2 X-12 replaced the first answer, a rebuild from the live session.)

### X-5 · major · `fileRows` guard and `canCite` timing in `prepare()`
**Codebase:** `EvidenceGraphModel.prepare()`; `EvidenceGraphView` `.onChange(of: model.canCite)` only republishes hit targets.
**Answer:** Fixed — D-9 keeps the workspace `.ready` guard and makes `prepare()` wait for the graph list; T-13 asserts the disarm.

### X-6 · major · Manual check §10 is not executable as written
**Answer:** Fixed — §10 rewritten from a new project with named graphs, links and steps, including the only-artifact frame and a reachable no-links graph (delete its only Artifact, then Back).

### X-7 · minor · Copy mismatch risk in the lock hint
**Answer:** Fixed — T-7 names the block's `via`, `kind` and `total`, as `SourceGraphFormModelTests.block` builds it.

### X-8 · minor · L10n details left to the implementer
**Answer:** Fixed — §6 Strings names the catalog keys, the translator-comment rule and the two `(String) -> String` messages; `noArtifactAction` is in Deletes. The trailing space before an empty error is accepted, as in `graphFormSaveErrorMessage`.

### X-9 · minor · Scrolling and layout
**Codebase:** `PVFormDialog` does not scroll its body.
**Answer:** Fixed — D-12 says so and relies on the 264pt list cap; a Stop condition covers a cut-off dialog.

### X-10 · minor · Weak checks
**Answer:** Fixed in part — STEP-2's check now runs the moved preflight tests; STEP-3 names the order (`visibleGraphs` is the session's order). Declined: a VoiceOver snapshot test — the repo has no UI test target (`add-swift-test` covers unit tests only); LK-10 stays a manual check line.

### X-11 · nit · Plan prose slips
Logged.

## Round 2

### X-12 · blocker · D-8's "live session" reads are stale right after a link or unlink
**Plan:** D-8 read `context.session.evidenceGraph(id:)` for "currently linked" and rebuilt `prefillExcluded` from it after an error.
**Codebase:** `WorkspaceSession.apply(_:)` invalidates the graph list and starts `revalidate`, then returns; `evidenceGraph(id:)` reads the old value until the reload lands (`WorkspaceSession.swift` ~178–235).
**Better shape:** Await `readyValue` after each write, or track successful writes locally.
**Answer:** Fixed — D-8 tracks them locally: each successful write patches `linkEdit.prefillExcluded` and the snapshot graph's `artifactIDs`; the diff and the locks read the snapshot; nothing waits on the reload. T-10 and the Risks row updated.
