# Review — S12-08 Graph id end to end

**Verdict:** Round 1: Fits with changes — Ready with changes. Round 2: pending
**Rounds:** 1 of 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- None.

## Writer's own calls

- Planning split S12-18 off and moved Promote's location and key into this PR. Entry points open the first graph in the list. All three were the user's calls (2026-10-22).
- D-1: old proto field numbers are reserved, not reused. The four RPCs named for a Source are renamed for the Evidence graph. `GetCitation` takes the graph id.
- D-5: Artifact and Citation writes still name every graph of their Source, as `main` does. Subject, Observation, value, claim and graph writes name their own graph.
- D-11: a history entry from before this PR that names a graph, composer or Promote surface opens the Source page.
- D-15: the test fixture `evidencegraphstest.SoleGraph` stays. The S12-07 plan had said S12-08 would delete it.
- M-1 (accepted): `.mutatedSourceGraph` is deleted rather than renamed. Graph writes refresh the Swift cache through their effects alone.

## Round 1

### M-1 · blocker · Rename keeps `.mutatedSourceGraph` alive when the write effects already cover it

Every caller already applies effects that name the Subject's graph (`subjects` → `field("graph_id")`). Delete the mutation and keep `.deletedSubject(sourceId:)` only for citation counts. `.promotedSubject(graphId:)` stays, because Promote's apply returns no effects.

**Answer:** Fixed.
- D-8 deletes `.mutatedSourceGraph` and its kind. `.deletedSubject(sourceId:)` is unchanged and leaves the graph key's `invalidateOn`. `.promotedSubject(graphId:)` stays.
- AC-2.6 is rewritten and AC-2.7 is added. T-14, T-15 and T-17 assert the effects path.
- A new Stop condition covers a write whose effects stop naming the graph.
- STEP-10 changes the doc examples.

### M-2 · blocker · The legacy-decode rule ends up in three places

**Answer:** Fixed.
- D-11 normalizes once in the memberwise `init`, and `init(from:)` shares it.
- D-10 drops the place-registry guard.
- T-13's guard case moves to T-23, an init test covering the proto mapper too.

### M-3 · major · Build `.graph` and `.citationComposer` locations through factories

**Answer:** Fixed.
- D-10 adds `WorkspaceLocation.evidenceGraph(…)` and `.citationComposer(…)` with a non-optional `graphId`. Every literal moves onto them.
- The STEP-8 check confirms no `sourceSurface: .graph` or `.citationComposer` literal is left outside them.
- Next adds names the factory.

### M-4 · major · The mutation-to-graph mapping is still hand-written in two places

**Answer:** Fixed.
- D-8 adds `CatalogMutation.evidenceGraphId`, which the registry and the session both read, and deletes `graphProgressSource(of:)`.
- STEP-10 deletes the idea section "Source graph mutations".

### M-5 · major · `CreateSubject` derives the Source in the handler, and S12-18 has to move it

**Answer:** Fixed.
- D-3: `subjects.InsertTx` reads the graph's Source when `SourceID` is empty, and `CreateSubject` sends only `GraphID`.
- An unknown graph is now `subjects.invalid` (AC-3.5, T-8). T-22 is new.
- The S12-18 scope already says "`CreateInput.SourceID` goes; `InsertTx` reads the graph's Source".

### M-6 · major · A skill this PR makes false is missing from STEP-10

**Answer:** Fixed. `add-audit-scope` is in STEP-10 and the DoD, and `Effect\.Source\b` is in the STEP-2 and STEP-10 checks.

### M-7 · minor · The graph delete-impact projector names S12-08 and is left alone

**Answer:** Fixed. D-10 makes `projectEvidenceGraph` return a `graph` surface with `GraphID`, and T-10 gains the case.

### M-8 · minor · The "one catalog querier" idea says to move packages this PR touches

**Answer:** Declined, with an Out line.
- This PR changes one SQL column in `conclusiondetails` and the registry entries in `effects`, not their read plumbing.
- Moving both queriers adds churn to a PR already at ten steps.
- The Out row points at the idea file, which keeps both open.

### M-9 · minor · `SourceTx` becomes a fourth graph-to-Source read

**Answer:** Fixed.
- D-3 makes it `SourceOf(q database.Querier, …)`, and `graphprogress.Get` uses it, deleting `sqlGraphSource`.
- `deleteimpact` keeps its two reads because `evidencegraphs` imports `deleteimpact`.

### M-10 · minor · Two registry paths are spelled out by hand next to the helpers that cover them

**Answer:** Fixed. D-5 adds `observationGraph`, and `sources` uses `graphsOfSource(self())`.

### M-11 · minor · No decision on the `.sourceGraph` place and presentation ids

**Answer:** Fixed. D-8 keeps `PlaceID.sourceGraph` and `WorkspacePresentationID.sourceGraph`. The STEP-6 check names them as the only allowed hits.

### X-1 · blocker · The AC-3.8 `rg` never matches

**Answer:** Fixed. STEP-3 writes the command with a bare `|`. T-21 cites it and has moved after T-20.

### X-2 · major · STEP-3's `GetSourceId()` check names the wrong expected output

**Answer:** Fixed. The file list drops `connect.go`, and the expected hit is `citations.go` `CitationCountsBySource` only.

### X-3 · major · The STEP-2 check cannot print nothing

**Answer:** Fixed.
- The pattern now matches only effects getters, and the check names the generated `GetSources()` that stay.
- STEP-2 rewrites the `evidence_graph.go` line 126 comment.

### X-4 · blocker · The FakeStore rule for writes that land on a graph is not decided

**Answer:** Fixed.
- D-14 says which graph each fake write names: Subject, Observation, bridge, Artifact and Citation writes.
- `bumpSource(_:graphs:)` replaces `bumpSource(_:evidence:)`, and `sourceOfGraph(_:)` is added.
- The fake's delete-impact and search locations use the D-10 factories.

### X-5 · major · D-14's `createSource` change may break existing Swift tests

**Answer:** Fixed.
- D-14 names the tests whose graph counts may change by the Default graph.
- STEP-8 and the Stop condition allow exactly that.

### X-6 · major · Reserved clauses would collide

**Answer:** Fixed. D-1 extends the existing `reserved` statements of `WriteEffects` and `CreateCitedBridgeRequest`.

### X-7 · major · No way to regenerate here

**Answer:** Fixed. There is a new Stop condition: without `protoc`, `protoc-gen-go`, `protoc-gen-swift` or `xcodebuild`, stop before STEP-2. The ship-pr implementer runs on macOS, where `scripts/check-pr.sh` runs.

### X-8 · major · Unresolved choices in the Swift registry and session

**Answer:** Fixed.
- **`.deletedSource`:** it is asserted in T-14.
- **`apply`:** progress graphs are `effects.evidenceGraphs` plus `mutation?.evidenceGraphId` (D-8).
- **Warm-up:** `.sourcesList` warms `.evidenceGraphsList` (D-10). The graph list invalidates itself on the `evidence_graphs` list key (S12-05), so the entry graph follows it.
- **Open button:** it is disabled while the list has no value (D-12).
- **File lists:** `SourcesSplitRow.swift` and the listed tests are added to STEP-8.

### X-9 · minor · Current state counts and notes that are not exact

**Answer:** Fixed.
- §4 now says 371 lines.
- STEP-6 renames the `L10n.swift` type reference and tightens its check.
- D-8 states the comment change plainly.
- STEP-10 now verifies the working rule rather than rewording it.
- D-6 relies on `graphprogress.Get` as it is.

### X-10 · minor · §8 and DoD wording

**Answer:** Fixed. T-21 cites its STEP-3 command, and the manual check states what each Back shows.
