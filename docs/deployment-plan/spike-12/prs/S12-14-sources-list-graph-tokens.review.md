# Review — S12-14 Graph counts in the Sources list and on the Source page

**Reshape after the board (2026-10-26).** Board S12-D5 drew a line of graph tokens instead of the accordion; the user adopted it (compact counts with a hover help) and chose to remove `primary_count` end to end (M-5 answered). The plan was rewritten as `S12-14-sources-list-graph-tokens.md`; Decisions D-1 to D-19 replace the accordion's, keeping the round 1 and 2 fixes that still apply (D-3, D-5 – D-9, D-13, D-14, D-17). Rounds 1 and 2 below reviewed the accordion plan; the token plan's review is under "Token plan".

**Verdict:** Fits with changes — Ready with changes (round 1: Not ready; every finding answered)
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)
**Covers:** the plan and brief [S12-D5](../design/S12-D5-sources-list-graph-accordion.md)

## To user

- Answered 2026-10-26: remove end to end (plan D-18, STEP-3).
- (was) **M-5 — `primaryCount` has no reader after the new status rule.** Recommended: drop it end to end in this PR (Go `graphprogress.Progress.PrimaryCount`, `engine.proto` field 5 reserved, the handler, `GoStore`, `FakeStore`, `EvidenceGraphProgress`, tests), as one added step, per "High tolerance for churn". The PR stops being Swift only. Alternative: Swift side only, Go and proto logged. Plan Open question 2.

## Writer's own calls

- D-1: a Source with no Artifact keeps "Needs an artifact" and does not expand, though the user chose "every row expands".
- D-5: Empty is `subjectCount == 0` (not `primaryCount == 0`), so a graph with only value Subjects reads In progress, and can never be Promoted.
- D-7: the user's any-graph rule for In progress is applied to Empty and Promoted too, so a Source can show under several statuses.
- D-6: the stored status value moves from `"filed"` to `"promoted"`, with `"filed"` read once as Promoted.
- Status label "Fully promoted" → "Promoted" (follows from the new meaning).
- D-12: a loaded graph list with no graph for the Source shows the loading state, not an error.

## Round 1

### M-1 · blocker · Second copy of the progress-handle reads
**Plan:** D-8 had `SourceGraphsSection.progress(for:)` / `countsLoading` copy `SourcesModel`'s reads. **Codebase:** `SourcesModel` `evidenceGraphProgressKey(for:)`, `progressHandle`, `graphProgress(for:)`, `graphCountsLoading`; `WorkspaceSession.refreshEvidenceGraphProgress` / `patchEvidenceGraphProgress` build the key inline; `WorkspaceSession+EvidenceGraphs.swift` is "the one place" for the graph list's key and handle. **Better shape:** progress reads in that session extension; callers forward.
**Answer:** Fixed — D-13 (new), STEP-1 (new), T-10; D-8 forwards; Codebase fit and Deletes updated.

### M-2 · major · Count line drawn by one view, spoken by a copy
**Codebase:** `L10n.Sources.graphZoneAccessibility` repeats `graphCountLine`'s three branches; the page card would combine raw pieces and never say "counts loading". **Better shape:** one spoken formatter; the view labels itself.
**Answer:** Fixed — D-3 adds `graphCountLineSpoken` and the view's own label; T-11.

### M-3 · major · Two row call sites, both growing
**Codebase:** `SourcesListView.listBody` builds `SourcesSplitRow` twice (`SourcesSplitRowProgress` and inline). **Better shape:** one row wrapper reading through the session.
**Answer:** Fixed — D-15, `SourcesListRow` (D-2), STEP-4.

### M-4 · major · More hand copies of "no value yet"
**Codebase:** `SourcesModel.graphCountsLoading` (twice), `isAwaitingFilterData`; `QueryHandle` has only `isSettled`. **Better shape:** `QueryHandle.isAwaitingFirstValue`.
**Answer:** Fixed — D-14, STEP-1, T-10. Checked `QueryHandle.applySuccess`: `.ready` always carries a value, so the types branch's `value == nil && status != .error` gives the same answer.

### M-5 · major · `primaryCount` loses its only reader
**Codebase:** only `SourcesModel.EvidenceStatus.init(_:)` reads it; it is carried by `graphprogress.Progress.PrimaryCount`, `engine.proto` `primary_count = 5`, `handlers/evidence_graph_progress.go`, `GoStore`, `FakeStore`, `EvidenceGraphProgress`.
**Answer:** To user — against the map's "Swift only"; recommendation above. The status rule staying in Swift until a second reader is logged in Out.

### M-6 · minor · Two names for one status
**Codebase:** `statusMenu` identifiers use the raw value; `SourcesListSettings.init(from:)` already handles stale values per field.
**Answer:** Fixed — D-6: raw `"promoted"`, legacy `"filed"` mapped in `SourcesListSettings`; T-3.

### M-7 · minor · Public observed `Set` on the session
**Codebase:** `graphCameras` is private behind `graphCamera(for:)` / `setGraphCamera(_:for:)`.
**Answer:** Fixed — D-4: private storage, `isSourceRowExpanded` / `toggleSourceRowExpanded`, read only in `SourcesListRow`.

### M-8 · minor · Brief inventory disagrees with the plan
**Answer:** Fixed — brief §6: Open graph zone's home is `SourcesGraphSubRow.swift`; Count line names two call sites and one spoken rule.

### M-9 · nit · `sources.page.graphsCount` now used by the list
**Answer:** Logged; key kept.

**Log it (reviewer):** `SourcesModel.warmListQueries()` warms the keys the `.sourcesList` place warms. Logged in plan Out.

### X-1 · blocker · `PlaceRegistryTests` not listed
**Codebase:** `PlaceRegistryTests.swift` `resolvesSourceDetail` and the place table assert the `.sourceDetail` keys.
**Answer:** Fixed — STEP-6 Files and Do; D-8 says "last of its query keys"; §4 row.

### X-2 · blocker · `rg "Fully promoted" docs` cannot pass
**Answer:** Fixed — DoD checks `rg -n 'fullyPromoted|Fully promoted' macos`.

### X-3 · blocker · STEP-3 check matches kept code
**Codebase:** `SourcePageModelTests.entryGraphLocation(_:)`.
**Answer:** Fixed — STEP-4 check pattern names the deleted symbols only and says the helper stays.

### X-4 · major · STEP-3 / STEP-4 order left open
**Answer:** Fixed — expanded state, expander and the D-9 deletions are one step (STEP-4).

### X-5 · major · Existing status tests need fixtures
**Answer:** Fixed — STEP-2 names each test and its new seed or case.

### X-6 · major · How T-7 / T-8 warm the progress key
**Codebase:** `SourcePageModel.warmFromSession()` has its own key list.
**Answer:** Fixed — D-8: `warmFromSession` unchanged; T-7 uses `session.apply(location:)` and `waitForQuery`.

### X-7 · major · Sub-row has no Source id for identifiers
**Answer:** Fixed — D-11 uses `graph.sourceID`.

### X-8 · major · Expander states and model API underspecified
**Answer:** Fixed — D-12: `SourcesModel.graphZone(for:)` / `GraphZone`, the zero-graph case, accessibility labels, the loading rule (D-14); T-9.

### X-9 · major · Two row branches
**Answer:** Fixed — D-15 (with M-3).

### X-10 · major · Sub-row layout and the magic 44
**Answer:** Fixed — D-11 paddings and spacing; D-16 `SourcesSplitLayout.thumbnailSize` / `subRowLeadingInset`.

### X-11 · minor · Mutator and animation
**Answer:** Fixed — D-4: session methods, no animation.

### X-12 · minor · ACs without tests or manual steps
**Answer:** Fixed — T-9 (expander states), §10 step 6 (page restored at relaunch), §10 step 2 (a graph with no Artifact, reachable by deleting its only Artifact).

### X-13 · minor · Brief inconsistencies
**Answer:** Fixed — brief Count line call sites (with M-8), frame 5 (row expanded earlier in the session), frame 7 (all three statuses).

**Checks the dry run could not run:** Swift builds and tests (no Xcode in the sandbox). Not a plan defect.

## Round 2

Reviewers checked every round-1 answer: all hold (maintainability notes M-4 left the "error with no value" check in two features, raised as M-10).

### M-10 · blocker · "One Source's graphs: loading, failed or rows" written a second time
**Plan:** D-12 `SourcesModel.graphZone(for:)` / `GraphZone` and D-7 `statuses(for:)` each read `evidenceGraphsHandle`, and disagreed on a loaded list with no graph for the Source. **Codebase:** `SourceGraphsSection.rows` / `loadError` already answer the same question for the page; `WorkspaceSession+EvidenceGraphs.swift` is the list's one reader. **Better shape:** one session read, `evidenceGraphsRead(sourceID:)`.
**Answer:** Fixed — D-17 (new) `EvidenceGraphsRead` / `evidenceGraphsRead(sourceID:)`; D-7, D-8 (`rows`, `loadError` forward), D-12 (`.loaded([])` → `.loading` for the row only) read it; STEP-1, T-9.

Below blocker (reviewer, not raised): brief §1 said "under its ref and description" — Fixed to "under its ref, above its description". §6 Next adds had a canvas reading `SourcesModel.EvidenceStatus` — Fixed: a reader outside Sources moves `EvidenceStatus` first.

### X-14 · blocker · A third navigation test calls a deleted helper
**Codebase:** `SourcesListNavigationTests.graphLocationBlockedWithoutArtifact` calls `graphLocation(for:graphs:)`; the private `graphsBySource(_:)` is used only by deleted tests; the STEP-4 pattern missed both.
**Answer:** Fixed — D-9, §4 and T-5 delete the test and the helper (the per-graph blocked case stays in `graphLocationOpensTheGivenGraph`); the STEP-4 `rg` pattern matches `graphsBySource(`.

**Verdict:** Fits with changes — Ready with changes. No blocker open. Open: board S12-D5 (plan Blocked) and M-5 (To user).


## Token plan — round 1

**Reviewers:** maintainability (Opus) — Fits with changes; executability (Sonnet) — Ready with changes.

### M-11 · blocker · `EvidenceGraphsRead` is a third copy of the loading / failed / loaded fold
**Codebase:** `ConclusionListPresentation.init(value:isFetching:status:error:)`, `ConclusionDetailPresentation`; `QueryHandle` has only `isSettled`.
**Answer:** Fixed — D-14 `QueryRead<Value>` / `QueryHandle.read` replaces `isAwaitingFirstValue` and the graph-only enum; D-17 returns `QueryRead<[CatalogEvidenceGraph]>`; `macos-client-patterns.md` row in STEP-1 and the DoD. Moving the two Conclusion presentations onto it: Out, logged here for a later Conclusions change (keeps this PR off Conclusions code).

### M-12 · blocker · The count rule written three times, and again as Empty
**Answer:** Fixed — D-3 `SourcesGraphCounts` and `L10n.Sources.graphCounts(_:style:)` (compact / full / spoken); D-5 Empty is `.notStarted`; the token and count line switch only for styling; T-11.

### M-13 · major · `openGraphNamed` duplicates `openEvidenceGraphLabel(name:)`
**Answer:** Fixed — the token reuses `openEvidenceGraphLabel(name:)`, whose default becomes "Open graph %@" (Source page Open reads the same); T-12.

### M-14 · major · Button styles do not name `PVHoverEffect` / `pvFocusRing`
**Answer:** Fixed — D-10 `SourcesLinkButtonStyle` and D-11 `SourcesGraphTokenButtonStyle` wrap `PVHoverEffect` and apply `pvFocusRing`, with no hand hover state or reduce-motion check; Codebase fit rows.

### M-15 · major · Fifth hand copy of the dashed stroke
**Answer:** Fixed — D-20 `PVStroke` token; the four `StrokeStyle` copies and `PVChip`'s pattern move onto it (STEP-4); brief inventory row added.

### M-16 · major · `PVFlowLayout` folder, preview and README row
**Answer:** Fixed — D-19: `Components/FlowLayout/`, `#Preview`, README row; STEP-4 Files and the pbxproj group.

### M-17 · minor · Two readings of "loaded, no graphs for this Source"
**Answer:** Fixed — D-17 reads it as `.loading` once; D-7 and D-12 use the session read directly; `tokenLine(for:)` is gone.

### M-18 · minor · `SourcesListNavigation` header goes stale
**Answer:** Fixed — D-9 rewrites the header; STEP-7's `rg` matches `split-row`.

### X-15 · blocker · STEP-3 `rg` matches the `reserved "primary_count"` line
**Answer:** Fixed — the pattern matches identifiers and `primary_count = ` only.

### X-16 · major · Proto regeneration needs tools a sandbox lacks
**Answer:** Fixed — §7 preamble, D-18 and a stop condition: stop and report, never hand-edit generated files.

### X-17 · major · A ButtonStyle cannot recolor a title that sets its own color
**Answer:** Fixed — D-10: the style publishes a private environment value; `SourcesLinkLabel` reads it for the title color, underline and cover outline. The board's 2 px underline offset is not available in SwiftUI's `underline`; the default underline is used (stated in D-10).

### X-18 · major · The 56 indent does not line up with the title
**Answer:** Fixed — D-10 link inner spacing `space4`; D-16 `tokenIndent` = 4 + 44 + 8 and the row's leading padding absorbs the link's 4 pt pad.

### X-19 · major · Token identifiers and labels missing for four states
**Answer:** Fixed — the token takes `sourceID`; the token states table lists visible text, help, label, hint, identifier and focus per state.

### X-20 · major · FakeStore's `primaryTypeIDs` would be dead
**Answer:** Fixed — D-18 deletes it and keeps `primaryTypeKeys`; STEP-3's `rg` matches `primaryTypeIDs`.

### X-21 · major · `listCaption` asserted in STEP-6 but added in STEP-7
**Answer:** Fixed — §6 Strings names the step for each key; the caption assertion is T-13 in STEP-7.

### X-22 · minor · `statuses(for:)` while loading or failed
**Answer:** Fixed — D-7 returns `[]`; T-2.

### X-23 · minor · STEP-4 README and pbxproj group
**Answer:** Fixed (with M-16).

### X-24 · minor · STEP-8 omits `design/README.md`
**Answer:** Fixed.

### X-25 · minor · Handler test edit unspecified
**Answer:** Fixed — D-18 and T-4: drop the primary clause, keep `FiledCount` 0; the `FiledCount` doc comment wording given.

### X-26 · minor · How tests warm, unwrap and apply
**Answer:** Fixed — §8 preamble: warm, `try #require` the handle, `waitForQuery`, `CatalogWriteEffects(evidenceGraphs:)`.

### X-27 · minor · Small visual choices open
**Answer:** Fixed — D-11 gives colors, the italic font, hover, focus, and blocked tokens as an empty-action `Button`.

### X-28 · nit · Stale "split row" comments
**Answer:** `SourcesListNavigation` fixed (M-18); `PVList.swift` is Out (a file this PR does not touch).

**Logged:** Shift-F10 for `pvRightClickMenu` (board frame 1) — a kit PR; the Conclusion presentations onto `QueryHandle.read`.
