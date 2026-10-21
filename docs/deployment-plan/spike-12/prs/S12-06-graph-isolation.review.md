# Review — S12-06 Graph isolation

**Verdict:** Fits — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None open. Two product calls were asked during planning and answered 2026-10-21: a `source`-typed Subject value is exempt from the graph check (D-3), and isolation gets its own code `observations.other_graph` with copy "That subject or artifact is in another evidence graph." (D-1).

## Writer's own calls
- D-3 narrowed after X-8: the exemption never applies to edge rows (bridge endpoints), and it reads the registry role `RoleReification`, not the key `source` (M-1).
- D-5: `subjects.CreateInput.GraphID` lands here, not in S12-08. It is optional, falling back to `SoleGraphTx`; S12-08 makes it required. The map's S12-08 entry is updated in STEP-8.
- D-6: `connect.CreateInput.SourceID` is removed, and the handler stops reading `CreateCitedBridgeRequest.source_id`. The wire field stays, unread, until S12-08 removes it. S12-08 adds no graph id to that RPC (M-3).
- D-4: the check runs after `resolveValue`, so an Observation that is both cross-graph and invalid in some other way reports the earlier error.

## Round 1

### M-1 · major · Exemption hard-codes `t.key = 'source'` instead of the registry's reification role
**Answer:** Fixed — D-3 uses `subjectvocab.TypeByKey(key).Role == RoleReification` with origin provenencia. The `vocabkey` constant is Out (not planned): there is no new literal now.

### M-2 · major · `sqlValueSubjectGraph` is a second graph-membership read
**Answer:** Fixed — D-2 reads both graphs through `evidencegraphs.SubjectGraphTx`. `sqlSubjectTypeKey` gains `t.origin`. `sqlValueSubjectGraph` is gone. `sqlSubjectExists` is deleted (D-4). The `evidencegraphs.ErrInvalid` → `observations.ErrInvalid` mapping is stated in D-2.

### M-3 · major · The S12-08 scope still says cited-bridge RPCs take a graph id
**Answer:** Fixed — the Out row and STEP-8 now rewrite the S12-08 entry: `source_id` is dropped and no graph id is added.

### M-4 · major · `source-layer-data-model.md` § 7.1 becomes false
**Answer:** Fixed — STEP-8, the DoD and the Context line now include it, and the working rule's S12-06 line names both docs.

### M-5 · major · D-9 defers a shared second-graph fixture although four packages need one
**Answer:** Fixed — `evidencegraphstest.Create` (D-9, STEP-2). The `deleteimpact` inline create moves onto it.

### M-6 · minor · A third per-code copy test
**Answer:** Fixed — `evidenceGraphErrorsHaveCopy` is renamed `errorCodesHaveCopy` and gains a row. The depiction test stays (Out).

### M-7 · minor · The unread `source_id` has no marker
**Answer:** Fixed (partly) — a Go comment goes in the handler (D-6). A proto comment is declined: it needs a regenerate for a field S12-08 deletes.

### M-8 · nit · The test is named after a private function
**Answer:** Fixed — `TestObservationsStayInOneGraph`.

### M-9 · nit · Two codes for "a graph of another Source"
**Answer:** Logged — the reason is written in D-5.

### X-1 · blocker · STEP-2 cannot pass its own check
**Answer:** Fixed — the steps are reordered. STEP-2 is `GraphID` plus the fixture helper, and STEP-3 is `requireSameGraph` plus T-1.

### X-2 · blocker · The T-1 and T-2 fixtures are not specified
**Answer:** Fixed — §8 has a T-1 fixture recipe (helpers, order, ids, one write call per case, what "no row left" means). T-2's setup is in its row. Book's graph is named "Book Default"; the Subjects Part and Rec are named.

### X-3 · major · The `SubjectGraphTx` error is not mapped
**Answer:** Fixed — D-2.

### X-4 · major · The early-failure order is not stated
**Answer:** Fixed — D-4 states it. §4 Tests says no test expects `ErrInvalid` for a cross-Source write, and a Stop condition covers it.

### X-5 · major · Invariant numbering
**Answer:** Fixed — append invariant 24, no renumbering.

### X-6 · minor · Literal counts are wrong
**Answer:** Fixed — the counts are dropped. §4 and STEP-4 rely on `rg`. `core/search` is added to the likely-to-break list.

### X-7 · minor · R-1's caller list is incomplete
**Answer:** Fixed — §4 has an "Observation writers" row from `rg`. Promote is not among them.

### X-8 · minor · A `source` endpoint could slip through Connect
**Answer:** Fixed — D-3 excludes edge rows. AC-1.11 covers it. The product rules have no `source` endpoint (`core/connectrules/connectrules.go` `productBridges`).

### X-9 · minor · The index check passes by default
**Answer:** Noted in AC-3.2. Accepted.

### X-10 · minor · The Swift checks cannot run on Linux
**Answer:** Fixed — STEP-7 names the fallback checks and says the Xcode boxes are reported for the user to run.

## Round 2

The maintainability reviewer confirmed M-1 to M-9: **Fits**, no new blockers. Its one wording note on M-3 is Fixed: STEP-8 now rewrites the whole S12-08 line, keeping List and create-Subject.

The executability reviewer confirmed X-1 to X-10: **Ready with changes**, no blockers.

### X-11 · major · T-3 mutates the shared walk fixture
**Answer:** Fixed — STEP-5 appends the subtest after the index subtest and restores York's `graph_id` with `t.Cleanup`.

### X-12 · minor · D-3's key read on the value Subject is not stated
**Answer:** Fixed — D-3 names the third call and the discarded `origin`, and maps `ErrNoRows` to `ErrInvalid`.

### X-13 · minor · T-4's audit read is not named
**Answer:** Fixed — the case uses the create change's `Fields["graph_id"]`, as `TestSubjects` does.

### X-14 · minor · Lint warnings
**Answer:** Fixed — the map row links the plan (status Planned), and the path is full.

### X-15 · minor · Two `source_id` paragraphs
**Answer:** Fixed — STEP-8 names the first and keeps the second.

### X-16 · minor · Length check in two places
**Answer:** Fixed — `InsertTx` only.

### X-17 · nit · Copy placement
**Answer:** Fixed — the new copy goes after its `observations.invalid` neighbours.

### X-18 · nit · STEP-3's check is narrow
**Answer:** Logged — R-3 covers it.

## Map changes
- The S12-06 row links the plan and is Planned.
- The S12-06 scope entry records what planning added.
- The working rule "Graphs are fully isolated" records the reification exemption and the code.
- STEP-8 updates the S12-08 entry when the PR lands.

## Refresh — S11-19 landed (2026-10-21)

S11-19 ([#513](https://github.com/mendahu/provenencia/pull/513)) deleted `canonicalgraph.WalkSource` and its SQL. One Source's walk is now `graphcache.Graph.WalkSource` over a loaded `SourceGraph`, and it reaches only Subjects on that `SourceGraph`.

- **Walk change dropped.** The plan's walk change is gone: the old D-8, STEP-5, T-3, US-3 and AC-3.1–3.2. The Out row sends any walk change to S12-07.
- **IDs renumbered:**
  - US-4 → US-3 and AC-4.1 → AC-3.1.
  - STEP-6–8 → STEP-5–7.
  - T-4–5 → T-3–4.
  - D-9 → D-8.
- **Heads-up suggestion declined:** keeping a `graph_id` join on `canonicalgraph.Walk` alone. `Walk` reads only `auto_reconciler_values` and `canonical_entities`, which have no graph (`core/database/canonicalgraph/canonicalgraph.go` `sqlWalkSelect`).
- **Touchpoints unchanged.** `git diff 9ca7472 10d3752` touches none of the plan's files.
- **One new likely-to-break package.** `core/database/graphcache` `source_walk_test.go` builds same-Source bridges, so it is added to the likely-to-break list.
- **Map conflict resolved.** The S12-06 scope entry conflicted with S11-19's edit and was rewritten: no walk change.
- **No new decision.** No re-review was needed, because the change only removes scope.
