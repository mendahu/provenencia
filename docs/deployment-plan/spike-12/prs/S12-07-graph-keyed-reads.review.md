# Review — S12-07 Graph-keyed reads

**Verdict:** round 1: Fits with changes — Not ready; answered below, round 2 pending
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None.

## Writer's own calls
- D-2: the element types are renamed too (`EvidenceSubject`, `EvidenceObservation`, `EvidenceMember`, `EvidenceEdge`) and the handler helpers become `listedProtosFromGraph` / `membershipsFromGraph`; the RPC and handler function names wait for S12-08.
- D-5: `evidence_graphs` rows now set `Effect.Source` (S12-03 D-5 left it off "until S12-07"), so a graph create, rename or delete reports its Source in `WriteEffects.sources`; link rows still do not.
- D-6 / D-9 / D-10: packages renamed `graphfill`, `graphprogress`, `grapheventtitles`.
- D-7: an unknown Source still reads empty through one handler helper; `SoleGraph` adds one indexed read per Promote and handler call until S12-08 / S12-09.
- D-9: `graphprogress.List` returns one row per graph; the handler writes one wire row per graph with its Source id, which is one per Source while each Source has one graph.
- Map: the S12-07 scope **Check** now reads "load and count independently; a write under the Source drops both" — the old "invalidate independently" contradicted the working rule "Effects are graph-keyed, in one step".

## Round 1

### M-1 · major · Claim evidence keeps three maps of one fact; keep one map of members
**Plan:** D-8 added `graphOf` and `graphOfSource` beside `sourceOf`.
**Codebase:** `core/database/promotealign/evidence.go` `evidenceReader.sourceOf`, `accepted`; `claimevidence.Member` and `Bridge` carry `Subject`.
**Better shape:** one `memberOf map[string]graphcache.Member`; `titles` resolves each Subject through it.
**Answer:** Fixed — D-8 replaces `sourceOf` with `memberOf`; `titles` resolves `p.Subject` / `b.Subject`; `graphOfSource` and its "first graph seen" rule are gone.

### M-2 · major · STEP-9 misses a rule, an idea file and code comments that the renames make false
**Plan:** the docs step and its `rg` scope.
**Codebase:** `.cursor/rules/catalog-queries.mdc:17`, `docs/ideas/shared-list-helpers.md:7,23`, `core/database/effects/registry.go` `Set.Sources` / `Effect.Source` docs, `graphcache` element doc comments, `promotealign/layer.go:42`.
**Answer:** Fixed — §4 gains "Comments" and "Docs naming the old symbols" rows; STEP-3 rewrites the code comments, STEP-6 the registry docs, STEP-10 the rule and idea file; the STEP-10 check uses `rg --hidden` over `.cursor` too; D-12 names the exempt paths.

### M-3 · major · Three handler sites each map "unknown Source → empty"
**Answer:** Fixed — D-7: `sourceGraphID` (the one `ErrInvalid` check) and `sourceEvidenceGraph` on top of it; positions and progress branch on `ok`.

### M-4 · minor · `graphprogress` reads `SourceID` with an extra join and query
**Answer:** Fixed — D-9 groups by `s.graph_id, s.source_id` (the composite FK guarantees it); `Get` returns `GraphID` only and the handler writes the request's Source id on both paths.

### M-5 · minor · Source-named helpers and element types on a graph-keyed type
**Answer:** Fixed — D-2 renames the four element types and the two handler helpers; the cost line is corrected.

### M-6 · minor · `promotealign.Querier` is a copy the idea file retires when touched
**Answer:** Fixed for `promotealign` (D-11). `graphcache.Querier` stays listed in the idea file (Out); STEP-10 drops `promotealign` and `sourceeventtitles` from that list.

### M-7 · nit · Two `SoleGraph` functions with different shapes
**Answer:** Fixed — D-7 says the fixture keeps its name and row return and S12-08 deletes both; Out lists it.

### X-1 · blocker · STEP-2 is one step with a tree that cannot build partway
**Answer:** Fixed — STEP-2 now does only what builds alone (element renames, `Member.GraphID`, package move); STEP-3 is the re-key, with the file order written out and a note that it is the one step that commits once.

### X-2 · blocker · The STEP-2 file list misses files that must change to compile
**Answer:** Fixed — `profiles.go`, `terms.go`, `sourceeventtitles_test.go`, `promotealign_test.go` added to STEP-2 / STEP-3; positions and progress tests are in STEP-7 / STEP-8 where those packages change.

### X-3 · blocker · The STEP-2 and STEP-9 `rg` checks cannot print nothing as written
**Answer:** Fixed — STEP-3's pattern names exact symbols (`Graph\(\)\.Source\(|g\.Source\(` …); `canonicalgraph.go` and the walk test comment are rewritten in STEP-3; STEP-10's check passes when only D-12's exempt paths match.

### X-4 · major · `Found` and the `!sg.Found` check in `loadInputs`
**Answer:** Fixed — D-8: `ErrInvalid` → `promote.ErrInvalid`; more than one graph returned as is; `!Found` → `promote.ErrInvalid` kept.

### X-5 · major · D-8 leaves the signatures and the `titles` fallback to the implementer
**Answer:** Fixed — D-8 lists every signature; `titles` skips a Subject not in `memberOf` (none exists: every evidence Subject came through `accepted`). The row Source's own title is not read today and is not added.

### X-6 · major · D-7 return value and handler detail
**Answer:** Fixed — D-7 notes no handler reads the empty graph's `SourceID`; D-9 has `GetSourceGraphProgress` write the request's Source id on both paths (T-10).

### X-7 · major · Checks that cannot be run or cannot tell the step is right
**Answer:** Partly fixed.
- Xcode tests: Declined. The template's DoD line requires `check-pr: PASS`, every Spike 12 plan keeps it (`S12-06-graph-isolation.md` §9), and `ship-pr` runs on the product owner's Mac. This PR has no Swift change, so the Xcode gate shows it touched no Swift.
- Pragma stop condition: Fixed — removed; no check reads those query plans.
- Stop condition 2's test name: Fixed — it names the `main` test it is compared with.
- `SoleGraph` outside the query counter: Fixed — D-7 accepts the one indexed read per call and says who removes it; §11 lists it.

### X-8 · minor · Unplanned decisions in tests
**Answer:** Fixed — §8 opens with the two-graph fixture recipe; T-13 sorts by `GraphID` in the test; D-9 and D-10 say the test package clauses change; STEP-7 names the moved test file; T-14 follows T-13's order in §8 only as a row (IDs are references, not order); D-6 has `checkEvidenceGraph` set `GraphID`; D-6 names `loadGraphSource` as one function.

### X-9 · minor · Temptation
**Answer:** Fixed — helpers renamed (D-2); the `graph.go` comments are in STEP-3.
