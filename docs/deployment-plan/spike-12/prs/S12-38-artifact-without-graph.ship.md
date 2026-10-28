# Ship log — S12-38 Add an Artifact without a graph; the card saves links in one write

**Plan:** [`S12-38-artifact-without-graph.md`](S12-38-artifact-without-graph.md) · **Branch:** `s12-38-artifact-without-graph` · **PR:** #605
**State:** fixing 1

## Round 0 — implement
Head `972b0de1`. Gate PASS (orchestrator, on the pushed head). Report: deviations: commits are coarser than one per behavior (STEPs 1–4 in one commit, STEPs 5–6 in another); the review log's link to the brief now points at the archive; `design/README.md` "Scheduled" reads "None open."; `sources.page.artifact.graphs.none` is at the end of `Localizable.xcstrings`; the manual check was not run (noted in the Landed note).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX ran for `macos/App/**` and `Localizable.xcstrings`, SEC for `api/ffi/**` (deleted handlers), PRF for the `macos/App/Features/**` models.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `FakeOCREngine` hand-rolls the hold that `FakeHold` now owns | Fix now → F-1 | verified: `CitationAutoTranscribeTests.swift:431-466` holds a flag under an `NSLock`, stores a `CheckedContinuation` and resumes it in `release()`, as `FakeHold` (`FakeStore.swift:8-42`) does. Same-pass cleanup: this PR adds `FakeHold`. Supersedes the Codebase-fit search limit in D-10 |
| UX-1 | UX | minor | A just-linked Artifact briefly shows "Not in an evidence graph" during the list refetch | Fix now → F-2 | verified: `artifactIDsInNoGraph` guards only `case .loaded = evidenceGraphsRead`, and `handle.read` stays `.loaded` on a stale refetch, while `evidenceGraphPresence` already checks `isSettled`. Minor in code this PR adds |
| UX-2 | UX | minor | A card opened in edit before the graph list loads never gets its group | Fix now → F-3 | verified: `beginEdit` builds `LinkEdit` only when `visibleGraphs` is non-nil, and nothing builds it later. Brief ND-1 (agreed board) says that while the list loads, the card shows the group's loading state (LK-8). Code this PR changes |
| CON-1 | CON | minor | `CatalogWriteEffects.merging` and `CoreInvokeError.code` lost their last callers | Fix now → F-4 | verified: `rg "\.merging\("` hits only `CatalogMutation.swift:96` and `CatalogQueryRegistryTests.swift:1242-1247`, and no `CoreInvokeError.code` reader remains (`OnboardingView`/`DeleteImpactFlow` pattern-match `.coded`). MNT's "merging still has callers" is wrong |
| MNT-2 | MNT | minor | The first-block-per-key lock reduction is written twice | Fix now → F-5 | verified: `SourceArtifactsSection.swift:218-223` and `SourceGraphFormModel.swift:85-89` differ only in key. Same-pass cleanup |
| UX-3 | UX | minor | Link to a graph, like the pencil, drops another card's unsaved edit without asking | Follow-up | behavior from before this PR (`beginEdit` → `cancelFields`). Needs the S10-12 dirty-confirm rule for both entry points; a product call for an idea file |

Nits (logged): `CON-2 · plan :4 · Landed note's commit grouping does not match the log` · `MNT-3 · SourceArtifactsSection.swift:327 · savedGraphIDs can be private` · `MNT-4 · SourceArtifactGraphLinks.swift:38 · point SourceGraphFormModel.locked's doc at lockHint` · `UX-4 · L10n evidenceGraphsArtifactInUse · card body repeats "an artifact you removed" (pre-existing)`. SEC+PRF: no findings. MNT's Outside-my-lens PRF note (index built every revision) is logged; it is linear and in scale.

### Fix plan 1

**F-1 — Move `FakeOCREngine` onto `FakeHold`; name `FakeHold` in `add-swift-test`** (from MNT-1)
Files: `macos/ProvenenciaTests/CitationAutoTranscribeTests.swift`, `.agents/skills/add-swift-test/SKILL.md`. Skill: `add-swift-test`.
Failing test first: none (refactor). The existing hold tests in `CitationAutoTranscribeTests` (around lines 312 and 325) must pass unchanged in what they assert.
Do: In `FakeOCREngine`, delete `shouldHold`, `hold` and `release()`. Add `let hold = FakeHold()` and call `await hold.wait()` in `recognizeText` after the locked read. The tests set `engine.hold.isOn = true` and call `engine.hold.release()`. Add one bullet to the FakeStore bullets in `add-swift-test/SKILL.md`: "To pause a fake call mid-write, give it a `FakeHold` (`isOn`, `count`, `release()`); do not hand-roll continuations." Leave `Gate` in `WorkspaceSessionTests` alone (plan Out, M-5).
Check: `rg -n "CheckedContinuation" macos/ProvenenciaTests` prints only `WorkspaceSessionTests.swift` (`Gate`); scripts/check-pr.sh PASS.

**F-2 — Mark nothing "in no graph" while the graph list refetches** (from UX-1)
Files: `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift` (and `SourcePageModelTests.swift` if the Add path test lives there).
Failing test first: extend `artifactIDsInNoGraphReadsTheIndex` (or add `artifactIDsInNoGraphIsNilWhileTheListRefetches`): with a loaded list, start a refetch that FakeStore holds (use a `FakeHold` on `listEvidenceGraphs` if one is needed, which F-1's rule makes the way), then `artifactIDsInNoGraph` → nil; after release → the set again.
Do: in `artifactIDsInNoGraph`, also guard `evidenceGraphsHandle?.isSettled == true`, the check `evidenceGraphPresence` uses. Update its doc comment ("nil unless the Source's graphs are loaded and settled"). Do not change `evidenceGraphsRead` (D-1). Confirm the Add path (`show`) starts the refetch before or in the same main-actor turn as it appends the card, so no frame shows the badge on the new card. If it appends first and an await falls between them, move `context.session.apply(effects)` before the append.
Check: the test passes; scripts/check-pr.sh PASS.

**F-3 — A card in edit picks up the group when the graph list arrives** (from UX-2; brief ND-1)
Files: `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift` (or wherever `graphLinks` renders), `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Failing test first: `cardEditBeforeTheListLoadsGetsTheGroup` (new): open `beginEdit("P3")` with no graph list value; `linkEdit == nil` and the card shows the group's loading state; then load the list; `linkEdit` exists for P3, with `excluded` from the loaded graphs, and its lock read starts.
Do: per ND-1, while the card is in edit and `visibleGraphs` is nil, the group area shows the group's existing loading state (the `SourceFormLoadArea` / LK-8 loading the group already uses when `loadState` is loading). Do not add a second loading view. When `visibleGraphs` becomes non-nil while `editingArtifactID` is set and `linkEdit` is nil, build `LinkEdit` and start `loadLinkLocks` the way `beginEdit` does. Extract the build from `beginEdit` into one private function both paths call. Trigger it from the view's observation of the graph read (e.g. `.onChange` of `visibleGraphs?.map(\.id)`) or the section's existing reload hook; pick the one the section already uses for list changes. A read error still shows S12-D4's callout as today.
Check: the test passes; scripts/check-pr.sh PASS.

**F-4 — Delete the two helpers this PR left without callers** (from CON-1)
Files: `macos/App/Features/Workspace/Session/CatalogMutation.swift` (`merging`, line 96), `macos/App/Platform/CoreInvoke.swift` (`CoreInvokeError.code`, line 32), `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` (the `merging` test, about lines 1235-1248).
Failing test first: none (deletion).
Do: delete both and the `merging` test.
Check: `rg -n "func merging|\.merging\(" macos` prints nothing; `rg -n "var code: String\?" macos/App/Platform/CoreInvoke.swift` prints nothing; scripts/check-pr.sh PASS.

**F-5 — One lock-hint reduction for both models** (from MNT-2; nits MNT-3, MNT-4 ride along)
Files: `macos/App/Features/Sources/SourceArtifactGraphLinks.swift`, `SourceArtifactsSection.swift`, `SourceGraphFormModel.swift`.
Failing test first: none (refactor); `cardEditLocksEveryBlock` and `editLocksEveryBlock` pass unchanged.
Do: next to `lockHint`, add `extension Sequence where Element == CatalogEvidenceGraphUnlinkBlock { func lockHints(by key: KeyPath<Element, String>) -> [String: String] }` holding the one first-block-per-key loop. `lockedGraphs = (linkEdit?.blocks ?? []).lockHints(by: \.graphID)`; the graph form's `locked` uses `.lockHints(by: \.artifactID)`. Make `savedGraphIDs` `private` if nothing outside the section reads it. Add a doc comment on `SourceGraphFormModel.locked` naming `lockHints` in `SourceArtifactGraphLinks.swift`.
Check: `rg -n "where hints\[" macos/App` prints nothing; scripts/check-pr.sh PASS.

Also: fix the plan's Landed note (CON-2) so it lists the commits as the log has them (STEPs 1–4, STEPs 5–6, STEP-7, docs), and add any fix commits.
