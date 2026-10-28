# Review — S12-22 Artifacts can stand outside every graph

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **M-6: retire the card's Swift last-link rule in S12-22 instead of S12-38?** Recommendation: yes. The card would look the same; `isOnlyArtifact` and the per-graph lock loop go one PR earlier. Map row S12-38 and this plan's Out change with the answer. The plan stays Draft until this is answered (§12).

## Writer's own calls

- The new code is `evidencegraphs.citations_need_graph`, and its copy is "Citations on this artifact need it to stay in an evidence graph."
- `evidencegraphs.in_use` is reworded to "Subjects or citations still need this evidence graph."
- One via and one heading serve the unlink block and the graph delete preview: "%lld citations would be left in no evidence graph".
- An unlink pair reports only its first reason, in this order: in use, last artifact, citations need graph.
- `artifacts.Update` takes the links as a fourth parameter, not an input struct.
- No migration repairs a Citation that A-3 already stranded in an existing project.

## Round 1

### M-1 · blocker · "This graph is the Artifact's only graph" spelled twice
Maintainability. D-6 (3) and D-10 each wrote the rule in their own SQL.
**Answer:** Fixed. D-10 adds `deleteimpact.SQLCitationsOnlyInGraph`. The probe and `checkUnlink` (3) both select from it, and `sqlArtifactCitations`, `sqlSoleGraphArtifacts` and the link count are gone (D-6, D-10, STEP-2, STEP-3, Scale).

### M-2 · blocker · A second vocabulary for the unlink reason
Maintainability. The proto enum, the Swift enum and two mappers duplicated the apperr codes.
**Answer:** Fixed. `EvidenceGraphUnlinkBlock.reason_code` carries `apperr.From(b.Reason).Code()`, and Swift reads `reasonCode` (D-12, §6 FFI, D-13, Next adds, T-11).

### M-3 · blocker · FakeStore gains a membership rule, and T-12 tested it
Maintainability.
**Answer:** Fixed. The fake records `createArtifactGraphIDs` and `updateArtifactGraphIDs` and keeps today's expression with the optional unwrapped, adding no branch (D-14). T-12 now tests the Swift side (what the Add dialog and the card send), in `SourcePageModelTests`.

### M-4 · major · `applyLinks` should own the diff for every link loop
Maintainability.
**Answer:** Fixed. `applyLinks(tx, userID, stored, want []Pair)` owns the distinct check and the diff. Graph `Create`, `Update`, `LinkArtifact` and `SetArtifactGraphs` call it (D-5, STEP-1).

### M-5 · major · Fake unlink blocks keyed by graph while blocks carry their graph
Maintainability.
**Answer:** Fixed. `unlinkImpactBlocks` and `getEvidenceGraphUnlinkImpactPairs` replace the graph-keyed hooks. `unlinkBlock` takes its graph with no default, and STEP-6 lists the call sites (D-14, D-15, T-13).

### M-6 · major · Retire `isOnlyArtifact` now through the pair-shaped preflight
Maintainability.
**Answer:** To user. It moves scope from S12-38 (see To user above).

### M-7 · minor · `parseGraphLinks` beside `parseIDs`
Maintainability.
**Answer:** Fixed. The helper goes in `api/ffi/handlers/catalog.go` (D-2).

### X-1 · major · Proto and Swift checks need the Mac toolchain
Executability.
**Answer:** Fixed. §7 says where those checks run and that generated code is never hand-edited.

### X-2 · major · Existing Swift tests that pass `graphIDs: []` meaning every graph
Executability.
**Answer:** Fixed. STEP-6 names the three tests and the replaced `unlinkImpactFiltersToAskedArtifacts`.

### X-3 · major · The STEP-6 check is vacuous
Executability.
**Answer:** Fixed. The check now runs `rg` over `macos` for the removed names and for `graphIDs: []`.

### X-4 · minor · The T-5 fixture contradicts itself
Executability.
**Answer:** Fixed. Default lists Y.

### X-5 · minor · STEP-4 and STEP-5 overlap
Executability.
**Answer:** Fixed. STEP-4 names `audit/scopes_test.go`, the handler call sites (zero value only) and the `TestArtifacts` rows; its check builds `./api/...` too.

### X-6 · minor · Two handler test helpers
Executability.
**Answer:** Fixed. D-16 covers `createArtifact` as well.

### X-7 · minor · Small open choices
Executability.
**Answer:** Fixed.
- The empty group for `last_artifact` is now specified (D-6, D-15).
- The reason is a code, so there is no unmapped enum.
- `SetArtifactGraphs` order is now specified (D-5).
- The `mapDeleteImpactGate` remark is gone with the enum.

### X-8 · nit · Risk row 1 is overstated
Executability.
**Answer:** Fixed. The row now notes that `Split` is safe.

### X-9 · nit · Housekeeping
Executability.
**Answer:** Fixed.
- The map row links the plan.
- The test line reference now reads ~153.
