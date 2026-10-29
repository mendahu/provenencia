# Review — S12-52 A Citation belongs to one Evidence graph

**Verdict:** Fits with changes — Ready with changes → Fixed (no open blocker)
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
None. Product calls were asked before planning (2026-10-29): a Citation belongs to one graph and never moves; a graph delete deletes its Citations; an Observation outside its Citation's graph is refused; no split in the backfill (the user's three dogfood catalogs have no Citation used by two graphs; one has two empty Citations on a page linked to two graphs).

## Writer's own calls
- Backfill order: earliest Observation's graph, else the Artifact's first linked graph in list order, else the Source's first graph (plus the missing link row).
- A create with an unlinked graph is `citations.invalid`, not a new user-kind code.
- One `ReleaseCitation` shared by graph delete and `citations.Delete`; `releaseEach` shared with `releaseSubjectClaims`.
- `graph_id` stays off the `Citation` proto and Swift type (no reader).
- New confirm line "The graph’s citations are deleted with it: %@." instead of the generic "This also changes".
- `error.evidencegraphs.in_use` reworded to "Subjects still need this evidence graph."

## Round 1

### M-1 · blocker · The delete-impact projector still derives a Citation's graph (same as X-1)
`projectCitation` falls back to `sqlFirstArtifactGraph`.
**Answer:** Fixed — D-6 has `projectCitation` read `c.graph_id` and deletes `sqlFirstArtifactGraph`; §4 Projector row; Deletes; T-7 asserts the listed row's `Location.GraphID`; STEP-2 sweeps the `impact_test.go` row.

### M-2 · blocker · Rules this plan bends omits the "An Artifact with Citations keeps at least one graph" rule
**Answer:** Fixed — listed in **Rules this plan bends** under the user's 2026-10-29 answer; D-11 / STEP-8; the map rule is marked superseded now.

### M-3 · major · `add-catalog-delete` teaches the removal shape D-6 replaces
**Answer:** Fixed — D-11 adds the shared `releaseX` / `ReleaseX` bullet and repoints the § Domain `Delete` example.

### M-4 · major · Backfill rule (3) breaks the linked-graph invariant (same as X-2)
**Answer:** Fixed — D-1 inserts the missing `evidence_graph_artifacts` rows after the rebuild; AC-3.1 asserts Z's link.

### M-5 · major · `releaseGraphCitations` copies `releaseSubjectClaims`' loop
**Answer:** Fixed — D-6 adds `releaseEach`; `releaseSubjectClaims` moves onto it (same-pass cleanup).

### M-6 · minor · Partial delete change
**Answer:** Fixed — D-6 writes `rowchange.DeletedRow` over every column.

### M-7 · minor · FakeStore counts keep the link filter
**Answer:** Fixed — D-10 drops it.

### M-8 · minor · Backfill order differs from the graph list
**Answer:** Fixed — D-1 orders by `name COLLATE NOCASE, ref, id`.

### M-9 · minor · Sweep misses test and vocabulary sites
**Answer:** Fixed — DoD sweep includes `macos/ProvenenciaTests`; STEP-5/6/7 name the vocabulary row, `SaveCitation` and the Swift test files.

### M-10 · nit · Stale comments
**Answer:** Fixed for the `sqlPersonNamesByArtifact` comment (D-2). Unexporting `CitationsUsingArtifactTx` declined: a rename outside the plan's need.

### X-1 · blocker · `projectCitation` fallback
**Answer:** Fixed — see M-1.

### X-2 · blocker · Backfill (3) and the invariant
**Answer:** Fixed — see M-4.

### X-3 · major · Step order breaks tests before the sweep
**Answer:** Fixed — shared helpers (`observationstest.SaveCitation`, `graphcachetest`) move into STEP-2; STEP-2/3/4 sweep their own packages; STEP-5 the rest.

### X-4 · major · Old-behavior tests not enumerated
**Answer:** Fixed — §4 "Old-behavior tests" row and STEP-2/3/6/7 give each file's new expectation.

### X-5 · major · Swift files missing from STEP-7
**Answer:** Fixed — `OnboardingModelTests.swift` (`ThrowingStore`), `L10nResolutionTests.swift`, `PVDeleteImpactTests.swift` added.

### X-6 · major · How FakeStore records `graphID`
**Answer:** Fixed — D-10 names the recorded string and the two asserts that change; T-10.

### X-7 · minor · Where the `GraphID` check runs
**Answer:** Fixed — D-2: in `normalizeCreateInput`.

### X-8 · minor · Docs rows incomplete
**Answer:** Fixed — D-11 / §4: `source-layer-data-model.md` 420–425, `deleteimpact/README.md`, the `evidencegraphs.Delete` comment; DoD sweeps the phrases.

### X-9 · minor · STEP-5 grep cannot prove completeness
**Answer:** Fixed — the test run is the check; the grep is gone.

### X-10 · minor · Tests with an unlinked Artifact
**Answer:** Fixed — D-8 rule.

### X-11 · minor · `TestReadsDeclare` change
**Answer:** Fixed — STEP-4.

## Round 2

Maintainability: every round-1 answer holds; no new blocker. Verdict Fits. Executability: one new blocker; verdict Ready with changes.

### X-12 · blocker · STEP-2's build check fails: the `deleteimpact` D-5 deletions break `evidencegraphs` until STEP-3
**Answer:** Fixed — all D-5 deletions move to STEP-3 with `register.go`; STEP-3's check adds `deleteimpact` and `go build ./...`.

### X-13 · major · `SaveCitation` signature
**Answer:** Fixed — D-8: `SaveCitation(c, userID, graphID, art, inputs...)`.

### X-14 · minor · Delete change described two ways (also M-6 leftover)
**Answer:** Fixed — D-6 keeps only `rowchange.DeletedRow` over every column, NULL as `nil`.

### X-15 · minor · Facet release Count / List / Remaining SQL
**Answer:** Fixed — D-6 gives the three queries; the stray `releaseGraphCitations` name is gone.

### X-16 · nit · Cascade line join
**Answer:** Fixed — AC-2.3 reads "CIT-A and CIT-B."

### Outside-lens notes from MNT round 2
- D-10 FakeStore counts named a nonexistent per-graph key → Fixed: counts every seeded Citation per Artifact.
- A migrated multi-graph Citation would make its graph's delete fail on a raw FK → Logged in §11 as accepted (no data has the shape).
- M-11 nit: "if callerless" hedge → Fixed.

**Final verdict:** Fits with changes — Ready with changes → Fixed. No open blocker; To user: none.
