# Review — S12-38 Add an Artifact without a graph; the card saves links in one write

**Verdict:** Fits with changes — Not ready (gated on board S12-D7; round 2 runs on the agreed board)
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet). Both reviewed the plan and brief [S12-D7](../design/S12-D7-artifact-in-no-graph.md).

## To user

- **M-1** (answered 2026-10-28): should the card lift the `citations_need_graph` lock in Swift when another box is checked, so a cited Artifact moves graphs in one Save? The user's answer: "Let's do it in two steps. No logic in Swift." The plan has no lift. Every preflight block locks its box on the card and in the graph form (D-3, D-7). A cited Artifact moves in two Saves (AC-3.7). The map's working rule "An Artifact with Citations keeps at least one graph" now records it.
- **X-1 / Open question 1:** the board S12-D7 is not agreed. The plan stays `Blocked — on board S12-D7` until it is, and round 2 runs on the final Strings.

## Writer's own calls

- D-7: the graph form now also locks a graph's only Artifact (`last_artifact`). Before, it filtered that reason out. This follows from "No logic in Swift". A swap inside the form now takes two Saves, as on the card. Brief ND-12 draws it.
- D-4: a card Save sends links only when they changed. It keeps any graph made during edit that links the Artifact.
- D-1: "the session has graphs for this Source" means non-nil and non-empty, as `evidenceGraphsRead` already treats it.
- Brief ND-3 / ND-4 copy, and the Badge with a ghost Button under the description, are recommendations for the board.

## Round 1

### M-1 · blocker · `effectiveLocks` puts a Swift copy of the "keeps at least one graph" rule back in the client
The Swift lift of the Citations lock re-derives the engine's `checkUnlink` third reason against the proposed set. Working rules "One unlink check, one link writer" and "An Artifact card's Save is one write" forbid that. The reviewer offered two shapes: the engine answers for the proposed set, or keep the lift and record the bent rule.
**Answer:** To user → answered (2026-10-28): no lift, two Saves, no logic in Swift. D-2 / D-3 rewritten: `needsGraph` and `effectiveLocks` are gone, and every block locks. **Rules this plan bends:** none.

### M-2 · blocker · "In no graph" is computed twice, outside the session's one place for graph link reads
`isInNoGraph` (D-6) and the new-graph prefill (D-8) each walked the list, per card. `WorkspaceSession+EvidenceGraphs.swift` owns graph link reads. The plan's **Next adds** line about the idea file was wrong: that idea is an engine `deleteimpact` group.
**Answer:** Fixed. D-6 adds `EvidenceGraphIndex.linkedArtifactIDsBySource` and `WorkspaceSession.artifactIDsInNoGraph(sourceID:artifactIDs:)`. D-6 and D-8 both read it. STEP-1 and T-0 are new. **Next adds** is corrected.

### M-3 · major · The card and the graph form handle the same unlink blocks in two shapes
**Answer:** Fixed. D-3: `LinkEdit.blocks` replaces `locked`. One `CatalogEvidenceGraphUnlinkBlock.lockHint` serves both models, and no reason `switch` is left. D-7: the form keeps every block and re-reads on any known `EvidenceGraphUnlinkReason`, with no code literal.

### M-4 · major · **Deletes** misses code and comments the one write makes dead or false
**Answer:** Fixed. **Deletes** lists `listEvidenceGraphsCalls` and every stale comment by file and line, and each Step rewrites the comments on the code it changes. Both are added to the DoD sweep.

### M-5 · major · D-10 hand-rolls a third FakeStore hold
**Answer:** Fixed. D-10 adds `FakeHold`, used by `updateArtifact` and `setSubjectPosition`, and `EvidenceGraphModelTests` moves onto it. Moving the test-side `Gate` in `WorkspaceSessionTests.swift` is logged in Out (not planned).

### M-6 · major · A card Save reloads the whole graph list to move links it just wrote
**Answer:** Declined for this PR. The map already schedules the fix: S12-26 ("Reload only what moved") replaces the catalog-wide `evidence_graphs` reload with a per-Source re-read merged through the cache layer, for every link write, the graph form's included. Doing it here would build a second patch path beside S12-26's. Out row added.

### M-7 · minor · The no-graph line is two copies of one caption
**Answer:** Fixed: `noGraphLine(_:)` in §6 macOS.

### M-8 · minor · Brief: the graph row's owner is the checklist recipe, not `PVCheckbox`
**Answer:** Fixed in the brief inventory.

### M-9 · minor · Brief: name the wrapping layout, and say why the mark is not in `metaRow`
**Answer:** Fixed. The inventory names `PVFlowLayout`, and ND-4 says why the row sits under the description.

### X-1 · blocker · Not startable: board not agreed, Strings not final
**Answer:** Agreed. This is the plan's status, not a gap in its text: it stays Blocked, and round 2 runs after the board is agreed (Open question 1).

### X-2 · blocker · "The session list has a value" is not one read
`evidenceGraphs(sourceID:)` returns `[]` for a loaded list with no graph for the Source.
**Answer:** Fixed. D-1 defines the one condition (non-nil and non-empty) and how tests produce it (`listEvidenceGraphsError`). D-2, D-6 and D-8 read it.

### X-3 · major · Unlisted existing tests break
**Answer:** Fixed. STEP-2, STEP-3 and STEP-4 give a verdict for each test in the S12-13 section, including `cardCancelDropsLinkChanges` and `cardFieldsOnlySaveErrorSaysChangesKept` (which now awaits `linkLoadTask`).

### X-4 · major · T-6's "no recorded call" clause can never fail
**Answer:** Fixed. The clause is dropped.

### X-5 · major · The resting row's layout is left to the implementer
**Answer:** Fixed. §6 macOS gives the spacing tokens, `.fixedSize()` children and "cards are not levelled", and brief ND-4 matches.

### X-6 · major · Names and ids not defined
**Answer:** Fixed. `cardInNoGraph(id:)` is defined; the line ids are `sources.page.addArtifact.graphs.none` and `sources.page.artifact.<id>.graphs.none`. Every new string has its comment, and `artifactLinkToGraphLabel(ref:)` follows `editArtifactLabel(ref:)`.

### X-7 · major · `artifactGraphOnlyArtifact` deleted by no Step
**Answer:** Fixed. STEP-3 deletes it, and D-11 names a Step for each deleted string.

### X-8 · major · The link hold's deletion step and the new hold's position are open
**Answer:** Fixed. STEP-4 removes the link hold; STEP-7 removes the link methods and their error vars. D-10 gives `updateArtifact`'s order: record, wait, then throw.

### X-9 · minor · A failed re-read's callout body is unspecified
**Answer:** Fixed. D-5 uses `artifactCardSaveErrorKept`. T-9b sets the error after the open read.

### X-10 · minor · A nil refusal code in the graph form
**Answer:** Fixed. D-7: a nil code does not re-read, and T-13 covers it.

### X-11 · minor · The manual check's data does not exist, and step 7 is impossible
**Answer:** Fixed. §10 has a seed, step 4 is the two-Save move, step 5 reaches `citations_need_graph` through a second window, and step 7 makes the no-graph Artifact through Add. AC-3.3's fixture has two Artifacts on Default.

### X-12 · minor · Docs step: no domain-doc change; archive links break
**Answer:** Fixed. The DoD's domain-doc line reads "none" with the reason, STEP-8 gives the `add-workspace-place` text, and the brief's links are rewritten when it is archived.

### X-13 · minor · Brief gaps: loading frames, card heights, ND-6 placement
**Answer:** Fixed. Brief frame 11 (loading), ND-4 (heights) and ND-6 (below the group, never over the load states).
