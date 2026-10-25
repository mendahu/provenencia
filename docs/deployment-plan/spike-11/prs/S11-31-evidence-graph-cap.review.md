# Review — S11-31 Bound the loaded Evidence graphs

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

Round 1 raised three executability blockers (X-1 – X-3), each a missing decision with a one-to-six-line fix, and no maintainability blocker or Reshape. All are Fixed below, so no round 2 was run.

## To user

- None.

## Writer's own calls

- The plan is written against [#536](https://github.com/mendahu/provenencia/pull/536)'s head before it merges (user, 2026-10-24); the implementer branches from `main` only after #536 merges.
- The cap is 64 Evidence graphs. It is sized from a measurement on S11-21's fixture: a full evidence pass over Source A reads 21 graphs, and one request reads at most 2.
- A hit on a memo counts as a use, so Promote's own graph is not evicted while it reads only memos.
- The PR is renamed "Bound the loaded Evidence graphs", and the map row and scope entry are reworded to match (M-2).
- The many-Sources page check runs on Source A, not the one-place Source (M-3).

## Round 1

### X-1 · blocker · The base branch is not stated, and the first stop condition fires on `main`
**Plan:** Header; Implementer step 2; Stop condition 1.
**Codebase:** `main` @ `6152d4b` lacks `perf_test.go` and S11-21's graphcache changes.
**Answer:** Fixed. **Branch** now says "from `main` after #536 has merged". **Planned** says `main` does not yet hold the §4 files. The §4 intro says to compare `main` with `74f64d0` once #536 has merged. Stop condition 1 already stops work while #536 is open. The reviewer read S11-21's row as "Landed #536" on #536's own branch; on `main` it is still Planned.

### X-2 · blocker · T-1 needs a Source type id that `newWalkFixture` does not expose
**Answer:** Fixed. D-8 gets the id the way `TestWalkEvidenceGraphStaysOnItsGraph` does: `sourcetypes.Upsert` with the "book" input. `walkFixture` is not changed. The §4 Memo tests row records this.

### X-3 · major · The T-1 table shape is under-specified
**Answer:** Fixed. D-8 adds a `capStep` table sketch for the order cases (T-1a, b, c, e). T-1d, T-1f and T-1g are plain subtests. T-1g gives the exact `OnCommit` call with revision 10. The §8 rows state each sequence in the table's own op words.

### X-4 · major · The map row status and its target text are not given
**Answer:** Fixed. STEP-3 now gives the `completed.md` heading and its first line (at the top, above S11-21), the map Status `Landed #N`, and the plan's status text. The Plan link and the Planned status are written by this planning commit.

### X-5 · minor · Several Current state rows are inaccurate or imprecise
**Answer:**
- `dropDerived`: Fixed. The row now says it also resets the match index.
- `evidenceRequests` and `pass`: Fixed. Both are now named as closures inside `TestLargeSourceBudgets`.
- `canon.go`: Fixed. It moved to a new "Memo-only readers" row, with `MemoSnapshotForTest`.
- Handlers: Declined. `api/ffi/handlers/subjects.go:160`, `citations.go:122`, `observations.go:28`, `promote.go:133` and `event_titles.go:55` each call `c.Graph().EvidenceGraph(graphID)`. The reviewer opened `evidence_graphs.go`, which is a different file. The row now names each file and the call.
- Working set: Fixed. The row says how it was measured. Stop condition 3 covers a pass that exceeds the cap.

### X-6 · minor · Small decisions left to the implementer
**Answer:** Fixed.
- D-5: the hooks are nil-guarded like `SetQueryCounterForTest`.
- §6 adds a one-line doc comment per function, and gives the const comment and the `EvidenceGraphMemo` comment text.
- D-6: `Drop` comes before recording; STEP-2 notes that `f.evidence` fails through the fixture's `t`.
- D-4: the order on a miss is store, touch, evict.
- STEP-3 gives the exact doc sentence.

### M-1 · major · Store one entry per loaded graph instead of adding a third parallel map
**Answer:** Fixed. D-1 and D-3 are rewritten. `loadedEvidenceGraph{graph, memos, used}` replaces the `memos` field and the planned `evidenceUsed` map. `forgetEvidenceGraph` is dropped, since removing a graph is now one `delete`. `dropDerived` clears memos on each entry. `Verify` and `LoadedEvidenceGraphForTest` read `e.graph`. §6 **Deletes** lists `Graph.memos`.

### M-2 · major · The title and map row say "Source" for what is an Evidence graph
**Answer:** Fixed. The plan is retitled "Bound the loaded Evidence graphs". The file, slug and branch are `s11-31-evidence-graph-cap`. This commit rewords the map row, the slice listing, the scope entry's `Verify` sentence and the Check to say Evidence graphs. S11-21's plan ("sizes its Source cap") sits on #536's branch and is left alone.

### M-3 · minor · T-2 is called the "hub-page check" but does not run on the hub Source
**Answer:** Fixed with option (b); option (a) is declined. The one-place Source's evidence reaches few graphs: its city handle's other events are bulk rows with no members (§4 Working set). So running it there would exercise no eviction. The plan and the map Check now both name "a page whose records hold members from many Sources", which is Source A (D-6).

### M-4 · nit · A test-only memo reader now changes eviction order
**Answer:** Fixed. D-2 and §6 add "A hit counts as a use of the graph" to the `EvidenceGraphMemo` doc comment. The §11 Risks table records it.

### M-5 · nit · An unbacked "Next adds" line and the name `useEvidenceGraph`
**Answer:** Fixed. The displays and details line is dropped, so **Next adds** is "None named". `useEvidenceGraph` is renamed `touchEvidenceGraph`.

## Refresh 2026-10-25 — S11-21 landed

S11-21 merged as #536 (`933cc3c`). Between the planning base `74f64d0` and `933cc3c`, no Go file changed: the later commits are the evidence graph canvas (macOS) and docs. Every §4 row was re-read on `main`, and the working set was measured again with the same throwaway test. Source A gives 23 graphs, 21 for a full pass and at most 2 per request, as before. The one-place Source gives 1 for a full evidence pass over all 33 handle rows.

Changes to the plan:

- The header, Planned and §4 now point at `main` @ `933cc3c`. Stop condition 1, which waited on #536, now checks the shape of `TestLargeSourceBudgets` on `main`.
- §4 Memo-only readers: the candidates memo lives in `loadCandidatesMemo`, not `loadCanon`. S11-21's Landed note records that move; the earlier row named the wrong function.
- §4 Working set and D-6 cite the one-place measurement.
- §3 Out: card image memory (`docs/ideas/evidence-graph-card-memory.md`, parked by S11-21) gets its own budget. This commit answers that idea's open question.
- Map conflict: S11-21's PR reworded S11-31's Check to "a Source whose events are all located at one place with many filed events". This plan keeps "a page whose records hold members from many Sources", because a full evidence pass on the one-place Source loads 1 graph and cannot reach any cap. Flagged to the user.
