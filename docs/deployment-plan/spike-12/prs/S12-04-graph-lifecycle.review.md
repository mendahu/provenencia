# Review — S12-04 Graph lifecycle

**Verdict:** Fits with changes — Ready with changes. No executability blocker; maintainability blocker M-1 and major M-2 are to the user.
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- **M-1** (answered 2026-10-19: shared check). The last graph is a `deleteimpact` extra gate `last_graph`; the map's working rule "Refusals are of two kinds" was reworded.
- **M-2** (answered 2026-10-19: one write). `Update` carries the full link set (empty = unchanged).

## Writer's own calls
- Graph names are not unique per Source and have no length cap (D-6).
- Graph list order is name (case-insensitive), then ref, then id, and is project-wide when no Source is given; linked Artifact ids are ordered by Artifact ref (D-2).
- `evidencegraphs.in_use` is conflict kind like every `*.in_use`; `evidencegraphs.last_graph` is user kind (D-9).
- The last-graph refusal is checked before the Subjects probe (D-10).
- The graph's link release on delete is not `Named`; naming linked Artifacts in the confirm waits for brief S12-D2 (D-12).
- The graph projector opens the Source page until S12-08 gives a graph a place (D-12).
- The S12-05 map scope entry gains `L10n.Errors` copy for the two new codes.
- No version bump.

## Round 1

### M-1 · blocker · The last-graph rule bypasses `deleteimpact`'s extra-gate mechanism, so preview and write disagree and S12-11 must rebuild the rule in Swift
**Plan:** D-10 private gate; Out row to S12-11; Risk row.
**Codebase:** `deleteimpact` `GateEdgeLocked` / `GateOriginLocked`, `Impact`'s `isEdgeObservation` check, `Codes.EdgeLocked`; `add-catalog-delete`: "Preview and write share one function".
**Better shape:** `GateLastGraph` checked by `Impact` before inbound collection, through a per-kind extra-gate list; `Codes.LastGraph`; proto value and copy in S12-05.
**Answer:** To user → user chose the shared check (2026-10-19). Fixed — D-15 (one per-kind `extraGates` list; `GateLastGraph`; `Codes.LastGraph`), D-10 has no private gate, AC-4.2 and T-5 check the gate, T-8 checks `Refuse`; the S12-11 Out and Risk rows are gone; S12-05's scope gains the proto value and recipe copy; working rule reworded.

### M-2 · major · `Update` gives S12-12's "edit its name and links" no slot
**Better shape:** `UpdateInput.ArtifactIDs`, nil = unchanged; link additions first, then unlink removals, in one write.
**Answer:** To user → user chose one write (2026-10-19). Fixed — D-7 diffs `UpdateInput.ArtifactIDs` (empty = unchanged, since proto3 cannot tell nil from empty) against the stored links, links additions then unlinks removals; AC-3.5 – AC-3.8 and T-3 cover it; S12-12's map scope says it saves through one `Update`.

### M-3 · major · The graph row's audit field map would be built twice, with different formatting
**Answer:** Fixed — D-10 adds `graphFields(g Graph)` beside `linkFields`; `insert` and `Delete` both use it. Codebase fit and Deletes updated.

### M-4 · major · `evidencegraphstest` would hold two ways to read a Source's graph
**Answer:** Fixed — D-4 is `evidencegraphstest.SoleGraph` returning the `Graph`; `Ref` goes and its two callers (`catalogsession`, `onboarding` tests) move; package doc rewritten. STEP-2 files and check updated.

### M-5 · major · `projectEvidenceGraph` adds another inline "trimmed name, else ref"
**Answer:** Fixed — D-12 renames `vocabTitle` to `titleOrRef` and moves the inline copies in `projectSubject`, `projectArtifact` and `projectSource` onto it. `projectCitation` keeps its own rule: it prefers a transcription snippet, so it is not the same body.

### M-6 · minor · `sqlSourceExists` repeats a check the schema already makes
**Answer:** Fixed — D-5 drops the pre-check; an unknown Source fails the `source_id` foreign key and `insert`'s `mapConstraint` returns `ErrInvalid` (`database.IsConstraintViolation` includes foreign-key failures).

### M-7 · minor · `ListTx` has no caller outside the package
**Answer:** Fixed — D-1 has `list(tx *sql.Tx, …)` and no exported `ListTx`; AC-1.6 removed.

### M-8 · minor · `graphRows` stays as a raw-SQL copy of `List`'s row read
**Answer:** Declined — `graphRows` is the independent table oracle for `TestCreateDefault` and `TestEnsureRefs` (S12-03 tests this plan does not otherwise touch); reading them through `List` would make those tests depend on the function under test elsewhere.

### X-1 · major · AC-5.1's list key has no test that asserts it
**Answer:** Fixed — T-7 adds `Resolve` cases over `Update` and `Delete` changes in `TestEvidenceGraphsResolve`; STEP-6 covers it; DoD line updated.

### X-2 · major · STEP-7 and the DoD cannot be completed on a host that is not macOS
**Answer:** Fixed — STEP-7 says what to report when the Xcode stage cannot run; `gofmt -l` added to its check.

### X-3 · minor · The map row does not link the plan
**Answer:** Fixed — the map row links the plan now; STEP-8 only sets its status.

### X-4 · minor · T-1's Artifact order fixture is underspecified
**Answer:** Fixed — AC-1.2 no longer names literal refs; T-1 reads refs by SQL and sorts its expectation.

### X-5 · minor · nil, empty and 15-byte `sourceID`
**Answer:** Fixed — D-1: `len(sourceID) == 0` means every graph; T-1 checks `[]byte{}` equals `nil`.

### X-6 · minor · T-1's raw insert inside `run` returns no changes
**Answer:** Fixed by M-7 — the `ListTx` case is gone. (`writes.Run` accepts an empty change list and rolls back: `core/writes/run.go`.)

### X-7 · minor · Delete's scope resolution relies on ghost rows; the Stop list covers updates only
**Answer:** Fixed — the stop condition covers any change from `Update` or `Delete`, including link deletes resolved after the graph row is gone; T-7 exercises it.

### X-8 · minor · Commit subjects missing for STEP-1, STEP-6, STEP-8
**Answer:** Fixed — each step names its subject.

### X-9 · minor · Branch base not stated
**Answer:** Fixed — the Branch line says to branch from the commit that adds this plan.

Smaller points: D-2 now says `list` closes the first query's rows before the second. "A copy of `in.ArtifactIDs`" is a new outer slice holding the caller's ids, as `Graph` copies elsewhere; left as written.

## Round 2

Maintainability (opus): round-1 answers hold; D-7 sound. One new blocker.

### M-9 · blocker · `extraGates` becomes a second per-kind gate list beside `originRules`
**Plan:** D-15 added a gate list beside `originRules` and kept `originGate` as a separate step; STEP-8's docs would call it "one per-kind list".
**Codebase:** `deleteimpact/register.go` `originRules` / `originRuleFor`; `impact.go` `originGate`; `docs/catalog-deletes.md` § Policy item 3 already names `origin_locked` an extra gate.
**Answer:** Fixed — D-15's `extraGateList` (`sync.OnceValue`) holds edge-locked, last-graph and one `originHolds(r)` entry per `originRules` row; `originGate` and `originRuleFor` go; `Impact` has one loop. STEP-5 runs `TestImpactGates` green on the edge and origin entries first. STEP-8 also fixes the `add-catalog-delete` checklist and `Delete` sketch lines that describe extra gates as domain-side checks.

Minor (not numbered): §6's `UpdateInput` sketch lacked `ArtifactIDs` — Fixed. D-7 reads one graph's links through the Source's `list` — accepted at today's sizes.
