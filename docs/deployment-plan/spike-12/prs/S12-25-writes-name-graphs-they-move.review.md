# Review — S12-25 Writes name only the graphs they move

**Verdict:** Fits with changes — Ready → Fixed (no open blockers)
**Rounds:** 2. **Reviewers:** maintainability (grok-4.7-xhigh), executability (composer-2.5)

## To user
None.

## Writer's own calls
- Closed `effects.Read` / `CatalogRead` vocabulary (mirror of `List` / `CatalogList`), not open strings.
- Citation Artifact graphs via `across`, not `sqlPath`.
- Citation refresh reuses extracted `applyHiddenPolicy` (same path as invalidations); progress keeps Get-one merge.
- Delete `CatalogQueryFollows.evidenceGraphs` when progress moves to reads.
- Composer Artifact visibility: `selectedArtifactID` setter copies `currentLocation` and `replaceCurrent`s (no rebuilt `composerLocation`).

## Round 1

### M-1 · blocker · Artifact-linked graphs already have an `across` walk
**Answer:** Fixed — D-3 / Codebase fit use `across("evidence_graph_artifacts", "artifact_id", "graph_id")`; stop condition updated.

### M-2 · blocker · Read names are an open string beside the closed list vocabulary
**Answer:** Fixed — D-1 adds `type Read` + `var Reads`, validate like Lists, Swift `CatalogRead`, vocabulary-test row, `CatalogQueryFollows.reads(CatalogRead)`.

### M-3 · blocker · Citation refresh reloads with a second copy of the descriptor loader
**Answer:** Fixed — superseded by round-2 M-6: no second store call; policy lives in one session function.

### M-4 · major · `CatalogQueryFollows.evidenceGraphs` has no caller after progress moves
**Answer:** Fixed — D-6 and Deletes remove the case.

### M-5 · minor · Audit-scope skill still says `Set Source too`
**Answer:** Fixed — D-10 includes rewriting that line onto `EvidenceGraph` / `Reads`.

### X-1 · major · Go step checks not runnable in this agent environment
**Answer:** Declined — environment limit; plan checks stay the repo commands for the implementer on macOS/CI.

### X-2 · major · Full Definition of done cannot be shown here
**Answer:** Declined — same; DoD remains `scripts/check-pr.sh` on macOS.

### X-3 · major · STEP-4 does not pick Model vs View for `replaceCurrent`
**Answer:** Fixed — D-8 / STEP-4 name `CitationComposerModel.selectedArtifactID` setter.

### X-4 · minor · STEP-1 file list vs Do disagree on README
**Answer:** Fixed — README removed from STEP-1 Files (STEP-6 / D-10 owns it).

### X-5 · minor · D-8 does not name a composer location builder
**Answer:** Fixed then superseded by M-7 — copy `currentLocation` instead of a builder.

### X-6 · minor · D-6 names `CatalogWriteEffectsReadName` but not its home
**Answer:** Fixed — folded into D-1's `CatalogRead` in `CatalogMutation.swift`.

### X-7 · nit · STEP-1 / T-2 allow two test file targets
**Answer:** Fixed — `reads_test.go` only.

### X-8 · nit · D-3 “both check to the same end” vs `sqlPath.check`
**Answer:** Fixed — M-1's `across` makes both arms end at `evidence_graphs`.

## Round 2

### M-6 · blocker · Citation refresh copies the hidden-key policy into a session helper
**Answer:** Fixed — D-6 extracts `applyHiddenPolicy(to:)` from `apply(invalidations:)` and citation refresh calls it; no `refreshCitationRead`.

### M-7 · blocker · `composerLocation` rebuilds the history entry from a second field list
**Answer:** Fixed — D-8 copies `navigation.currentLocation`, sets `artifactId`, `replaceCurrent`; no `composerLocation`.

Executability round 2: Ready (no new blockers).
