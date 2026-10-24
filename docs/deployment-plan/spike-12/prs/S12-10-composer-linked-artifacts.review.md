# Review — S12-10 Composer offers linked Artifacts

**Verdict:** round 1: Fits with changes — Not ready. Round 2: pending
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None open. Product calls answered at planning (user, 2026-10-24): the empty callout always names the graph; the Citation stepper and counts are filtered by graph in this PR; the canvas's cite gate and Files list move to S12-13.

## Writer's own calls
- D-1: a graph's Citations include Citations with no Observation yet, shown in every graph linked to their Artifact.
- D-2: observation counts and person names count only the graph's Observations.
- D-3: an Observation-less Citation does not block an unlink.
- D-6: `WriteEffects.citation_sources` stays; the app drops all cached citation reads on it.
- M-7 accepted: citation counts are over the graph's linked Artifacts; until S12-13 the canvas Files menu shows no count for an Artifact unlinked from its graph (only on multi-graph Sources).

## Round 1

### X-1 · blocker · STEP-8 asks for an S12-13 edit that is already there
The map's S12-13 entry already carries the canvas sentence (added when the map was wired at planning).
**Answer:** Fixed — STEP-8 edits only the S12-10 row status; the Out row says the S12-13 entry already says so.

### X-2 · major · The `completed.md` path is wrong
**Answer:** Fixed — STEP-8 names `docs/deployment-plan/spike-12/completed.md` and § 7.1 (line 420).

### X-3 · blocker · STEP-3's build check cannot pass before STEP-4
**Answer:** Fixed — D-6's code moves into STEP-3; STEP-4 is tests only.

### X-4 · major · Existing Swift tests break and no fixing rule is given
**Answer:** Fixed — STEP-5: the fixture graph links `["art-0", "art-1"]`; T-6 – T-9 set their own links.

### X-5 · major · Entry case for AC-1.5 / T-9 not named
**Answer:** Fixed — AC-1.5 and T-9 use the `.addProperty` entry with `artifactID: "art-0"` and a graph linking art-1 only.

### X-6 · major · FakeStore changes not fully specified
**Answer:** Fixed — D-10 names `citationIsInGraph`, the `personNames(for:graphID:projectDir:)` signature, and that counts use the same helper over the graph's linked Artifacts (after M-7).

### X-7 · major · T-10 and T-11 data and second-model setup unspecified
**Answer:** Fixed — STEP-7 adds `makeModel(graphID:)` and `seedSecondGraph` (graph `g-2`, person `sub-person-2` Clara); T-10, T-11 name ids and data.

### X-8 · minor · Go SQL details left to guess
**Answer:** Fixed — STEP-1 states how the fragments are concatenated and each query's bind order.

### X-9 · minor · Existing Go tests and the handler test need more than "move onto graph"
**Answer:** Fixed — STEP-1 names the cases and the `SoleGraph` call; T-1, T-2 name Subjects (`s.person` Alice, Ben, Clara via `subjectstest.On`); STEP-2 reads the graph id at line 245 with `ListEvidenceGraphs`.

### X-10 · minor · Order of warm in `warmCatalogHandles`
**Answer:** Fixed — D-8: queried after `graphKey`, added to the `guard`.

### X-11 · nit · Temptation items not covered
Logged. `cachedCitation(id:)`'s loop is named in STEP-3; the skill lines are in STEP-8 (M-3).

### M-1 · major · `.deletedSubject(sourceId:)` is a second, redundant path to the citation keys
**Answer:** Fixed — D-6: `.deletedSubject` loses its payload and leaves both citation specs; the effects are the one path. STEP-4 updates the five test calls; a Stop condition covers erase effects without `citation_sources`.

### M-2 · major · `.allCached(.citationCounts)` reloads every cached graph's counts
**Answer:** Fixed — D-6: the `.citationCounts` spec gains `whenHidden: .markStale`; T-5 asserts it.

### M-3 · major · Skills and comments this PR makes false are not in STEP-8
**Answer:** Fixed — STEP-8 edits `add-workspace-place` (lines 172 – 173), `add-catalog-delete` (lines 57, 111) and the `effects/registry.go` comments.

### M-4 · minor · D-7 gives S12-13 only the filter, not the graph-by-id lookup
**Answer:** Fixed — D-7 adds `SourcesListNavigation.linkedArtifacts(graphID:artifacts:in:)` beside `graphsBySource(in:)`; D-8 calls it.

### M-5 · minor · The `.citationsByArtifact` key is built at three sites
**Answer:** Fixed — D-11 `CitationComposerContext.citationsKey(artifactID:)`. The context's other duplicate keys and its unread `sourceID` are logged here, not planned.

### M-6 · minor · The "Observation on a Subject of the graph" rule is written in two packages
**Answer:** Fixed — D-1 names `sqlUsedByGraph` with a comment pointing at `evidencegraphs.sqlCitationsUsing`; the map's `evidencegraphs` rule says `citations` owns the graph's Citation reads; the § 7.1 bullet refers to the Unlink rule.

### M-7 · minor · "The graph's Artifacts" means two things
**Answer:** Fixed — D-4 counts over `evidence_graph_artifacts`; T-3 adds the unlinked-Artifact row. The canvas Files menu effect is under Writer's own calls.
