# Review — S12-09 Promote per graph

**Verdict:** Fits with changes — Ready with changes (round 1)
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

None.

## Writer's own calls

- **D-2 (from review M-2).** `identityclaims.UnfileBridges` takes only the Subject. It is not re-keyed to a graph id. Isolation keeps every bridge on its ends' graph, so unfiling is per graph by construction, and the working rule "filing and unfiling bridges is per graph" holds without a parameter. A caller can no longer pass a graph that disagrees with the Subject. Veto it, and D-2 goes back to `UnfileBridges(tx, graphID, subjectID)`.
- **Waiting bridges on another graph.** Done files only the waiting bridges on its own graph (D-1, T-4). A waiting bridge on another graph files on that graph's Done, or when a claim completes it.
- **No graph on `UnpromoteSubjectRequest`** (D-4). The engine reads the graph from the Subject.
- **Renames.** `FileSourceBridgesTx` becomes `FileEvidenceGraphBridgesTx`. The "graph alignment" RPC and type names stay: "graph" there means `graphalign`.
- **Test helper.** `(*fixture).graph(src)` in `promotealign` tests (D-9). The package's six spelled-out sole-graph reads move onto it.
- **Manual check.** The app can't make a second graph before S12-12, so the two-graph behavior is checked in Go tests only.

## Round 1

### M-1 · major · Re-keying keeps two copies of the bridge-head query and its row scan
`FileSourceBridgesTx` and `bridgesTouching` in `core/database/identityclaims/bridgefile.go` each build the same key list, join, select and scan loop. **Better shape:** one private `bridgeHeads(tx, cond, args...)`.
**Answer:** Fixed. D-1 adds `bridgeHeads`. `FileEvidenceGraphBridgesTx` and `FileBridgesTx` call it, and `bridgesTouching` is deleted. Codebase fit and Deletes are updated. `UnfileBridges` keeps its own query, because it selects claim ids, not heads.

### M-2 · major · `UnfileBridges(tx, graphID, subjectID)` takes two ids for one fact
**Better shape:** `UnfileBridges(tx, subjectID)`, the mirror of `FileBridgesTx`. Per graph holds through isolation.
**Answer:** Fixed, as a writer's call listed above. D-2, D-4, AC-1.6 and T-5 are rewritten. D-10 deletes the subtest "a bridge on another Source stays filed", because no call can name a scope now.

### M-3 · minor · The new `f.graph` helper leaves the same read spelled out in the same fixture
**Answer:** Fixed. D-9 moves `bare`, `memo_test.go` and `source_store_test.go` onto `f.graph`. STEP-3 files and the Codebase fit row are updated.

### M-4 · nit · `SaveBatch` checks the graph id's length twice
**Answer:** Fixed. D-3 drops the length check, and `SourceOf` owns it.

### X-1 · blocker · Handler test mutators that set `SourceId` to something other than `f.sourceID`
`promote_evidence_test.go` has the mutators "an unknown Source" and "a Subject not on the Source".
**Answer:** Fixed. STEP-5 names both, their new names, and their values (`uuid.NewV7`, and `soleGraphID(t, f.dir, otherSource)` read after `otherSource` is made).

### X-2 · major · How two-graph tests cite without `f.bare` / `f.person`; positions
**Answer:** Fixed. D-12 says:
- Subjects come from `subjectstest.On`, with no canvas position, as `TestEvidenceReadsMemberGraph` does.
- `promotealign` cites with `f.cite(<artifact>, …)`.
- In `batch_test.go` the Jones Subject is not cited, and the Default person and event are `f.bare` made before the split.

### X-3 · major · T-2 does not name the Census Source
**Answer:** Fixed. T-1 and T-2 use `census, censusArt := f.newSource("Census")`. T-2's handle comes from the fixture's own Source.

### X-4 · major · AC-1.5 claims `ResolvePins` with no test, and `ResolvePins` validates lazily
**Answer:** Fixed. AC-1.5 says `ResolvePins` refuses a bad graph for a handle row that asks for default pins. T-6 adds the subtest "an unknown graph is invalid", and covers AC-1.5.

### X-5 · minor · Steps leave the tree red between commits
**Answer:** Fixed. The commits line now says three commits: STEP-1 – STEP-5, STEP-6 – STEP-7, then STEP-8.

### X-6 · minor · `generate-proto.sh` and `xcodebuild` are not in the review sandbox
**Answer:** Declined. These are environment limits, not plan defects. §11 already stops before STEP-1 commits when `protoc` or `xcodebuild` is missing, and the generated files must not be hand-edited.

### X-7 · minor · Some call sites covered only by "every call"
**Answer:** Fixed. D-9 names `memo_test.go`'s direct `Propose` calls and `pins_test.go`'s `EvidenceRequest` literals and direct `ResolvePins` calls.

### X-8 · nit · Handler-test details left implicit
**Answer:** Fixed. STEP-5 says to read `soleGraphID` on the line before a request built after `personOnNewSource`.

### X-9 · nit · Adjacent edits covered
**Answer:** Logged.
