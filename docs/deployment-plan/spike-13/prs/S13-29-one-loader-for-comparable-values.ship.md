# Ship log — S13-29 One loader for a Subject's comparable values

**Plan:** [`S13-29-one-loader-for-comparable-values.md`](S13-29-one-loader-for-comparable-values.md) · **Branch:** `s13-29-one-loader-for-comparable-values` · **PR:** #741
**State:** ready

## Round 0 — implement
Head `c9462e27`. Gate PASS (run by the orchestrator on the pushed head). Report: STEP-7's `main` budget run used `git checkout main` in this checkout, not a second worktree (the user's standing rule; orchestrator's instruction); D-13's GNU sed replaced by an equivalent perl rename (BSD sed has no `\u`); STEP-3's `rg` still prints two `connectrules.Bridge{}` lines, a different type; `promotealign_test.go` dropped its redundant `graphfill` blank import; `docs/ideas/structural-debt.md` S14-36 entry's `promotealign.decode` became `evidencelayer.Decode` (DoD sweep); §10 manual check not run in the app.

## Round 1 — review
Base `986d304f`, head `c9462e27`. Reviewers: CON, MNT, SEC+PRF (security: `core/subjectcompare`, `core/subjectprofile`, `core/recordlinks`, and the catalog value loader; performance: `core/database/**`). Skipped: UX (no `macos/App/**`, `Localizable.xcstrings`, `core/apperr`, or new error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC+PRF | blocker | `WalkEvidenceGraph` has no bound on one bridge's ends; 8,000 crafted rows cost 38 s and 18.8 GB in `Load` | Fix now → F-1 | Read: `evidence_graph.go` loops `reaches` × `c.tos` with no cap; `kinship`'s caps run later, in `Derive`. The reviewer's probe measured the product. The plan's Untrusted size line is wrong (not a Decision; the Landed note says so). One high total cap, per the user's S13-26 SEC-1 answer |
| CON-1 | CON | major | `TestLoadSharesPromotesMemo` passes when Promote never touches the memo | Fix now → F-2 | Mutant run by the reviewer: unmemoized `loadInputs`, all four memo tests pass |
| MNT-3 | MNT | minor | `MemoSnapshot` is a one-field wrapper with a false doc; the Load lookup is written three times | Fix now → F-2 | Same-pass cleanup; supersedes D-12's `MemoSnapshot` line. CON-6 (nit) is the same stale doc |
| MNT-2 | MNT | major | The Source-building helpers stay copied in other test packages | Fix now → F-3 | Same-pass cleanup: this PR moved one copy. Supersedes the plan's Out row (user's rule, S12-18: a planning deferral does not stop a test-only dedupe). The copies are external test packages (`head` of each), so the cycle does not apply. A body grep finds more than the five named: `promote/bridges_test.go`, `autoreconciler`, `conclusiondetails`, `search/handles_test.go`, `persondetail/claims_test.go` |
| MNT-1, CON-3 | MNT, CON | major | `docs/matching.md:11` says `promotealign` builds both sides | Fix now → F-4 | Read the line; same sentence as `add-match-rule` line 15 |
| MNT-4 | MNT | minor | `checkForbids` lacks `checkPure`'s `IgnoredGoFiles` check | Fix now → F-5 | Reviewer's probe: a tagged file escapes; `evidencelayer` has no `layerPackages` row |
| CON-5 | CON | minor | `TestLifeEventOf` drops T-1's mixed-polarity row | Fix now → F-6 | Read the table |
| PRF-1 | SEC+PRF | minor | Past 64 graphs a person cycles the LRU; the Out line's debt entry is missing | Fix now (entry) → F-7; Later PR (behavior) → S13-27's map scope entry | `grep` finds no entry; the caller does not exist yet |
| CON-2 | CON | major | The plan has no Landed note | Fix now → F-8 | `git diff` of the plan is the Status line only |
| CON-4 | CON | minor | A graph deleted between Promote's two reads returns `ErrNoGraph` | Rejected | Not reachable: `Propose` runs inside `catalogsession.Do`, and "Operations on the same project serialize" (`core/catalogsession/session.go:186`), so no commit lands between the reads |

Nits (logged): `CON-7 · evidencelayer_test.go · no case for Load(nil, ...)` · `MNT-5 · evidencelayer.go:154 · ps.Values = …; append(subjects, ps)` · `MNT-6 · evidencelayer.go:116 · ids rebuilds primary's keys` · `MNT-7 · evidencelayer.go:234 · pass rules.Kinship from build` · `MNT-8 · promotealign.go · name the evidencelayer.Layer param evidence` · `MNT-9 · promotealign_test.go:45-51 · fixture keeps duplicate fields; declined at plan review` · `MNT-10 · subjectcompare.go:74 · "or handle" ahead of the structural-debt fold` · `MNT-11 · add-match-rule/SKILL.md:96 · move the readerPackages sentence to the paragraph's end` · `MNT-12 · conclusion-reconciliation.md:290 · could list evidencelayer`

### Fix plan 1

**F-1 — Bound one Evidence graph walk before it builds an edge** (from SEC-1)
Files: `core/database/graphcache/evidence_graph.go`, `core/database/graphcache/evidence_graph_walk_test.go`, `core/database/graphcache/export_test.go` if a hook is needed. Rule: `.cursor/rules/catalog-queries.mdc`.
Failing test first: `TestWalkEvidenceGraphEdgeCap` — an in-memory `EvidenceGraph` with one bridge citing 224 Subjects on each endpoint (50,176 edges) → today it returns the edges; after the fix it returns no edges and a nil error. A second case at just under the cap (223 × 224) still returns every edge.
Do: `const maxWalkEdges = 50_000`, unexported, doc: the most edges one walk builds; a graph over it is malformed and the walk reaches nothing; real graphs hold about one edge per bridge Subject, and a 100-person Source about 200 (state the figure you measure). After the `eg.Observations` loop and before `seen` is made, sum `len(c.tos)` over `reaches`, skipping `filtered && !c.filtered`; over the cap, `return nil, nil`. The count runs before any edge, so the result does not depend on order. Extend `WalkEvidenceGraph`'s doc with one sentence. This is the one high total cap the user asked for on S13-26 SEC-1: no tuning, no second axis. It covers `graphfill` too, which calls the same walk. Log the largest edge count one walk sees in `TestLargeSourceBudgets`' fixture in your report (a throwaway run, not committed).
Check: the test passes; `TestLargeSourceBudgets` and `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/...` pass; scripts/check-pr.sh PASS.

**F-2 — Prove a Propose stores the layer, and read the layer through one test helper** (from CON-1, MNT-3, CON-6)
Files: `core/database/evidencelayer/evidencelayertest/evidencelayertest.go`, `core/database/promotealign/memo_test.go`, `core/database/promotealign/export_test.go`, `core/database/evidencelayer/evidencelayer_test.go`.
Failing test first: `TestLoadSharesPromotesMemo` — after `propose`, read `c.Graph().EvidenceGraphMemo(graph, "evidencelayer\x00"+namePattern)` directly (the literal key, as `TestLoadUnknownGraph` does; no production symbol, D-4). Fail with "Propose memoized no layer" when absent. Then, counting queries, `evidencelayer.Load` → 0 queries and `&got.Subjects[0] == &stored.Subjects[0]`. Prove it by pointing `loadInputs` at an unmemoized build locally: the test must fail. Revert that mutant.
Do: add `func (r *Records) Layer(graph []byte) evidencelayer.Layer` to `evidencelayertest`: open the DB, `valuerules.Load`, `evidencelayer.Load`, `r.T.Fatal` on error. Delete `MemoSnapshot` and `MemoSnapshotForTest` from `export_test.go` (supersedes D-12's `MemoSnapshot` line; the Landed note says so). Use `Layer` in `memo_test.go`'s `snapshot` readers, `memoJSON`, and `evidencelayer_test.go`'s `links`. `TestMemoIsNotMutated` keeps its assertions on the JSON; fix the `snapshot` comment. No `promotealign` assertion changes except dropping the check the deleted hook carried.
Check: the three call sites share the helper; `grep -rn MemoSnapshot core` prints nothing; scripts/check-pr.sh PASS.

**F-3 — Move every Source-building test helper copy onto the shared records** (from MNT-2)
Files: `core/database/grapheventtitles/grapheventtitles_test.go`, `core/database/canonicalgraph/canonicalgraph_test.go`, `core/database/conclusionsummaries/walks_test.go`, `core/database/conclusionsummaries/conclusionsummaries_test.go`, `core/database/promote/batch_test.go`, `core/database/promote/bridges_test.go`, `core/database/graphcache/evidence_graph_walk_test.go`, and any further hit of the Check grep; `core/database/evidencelayer/evidencelayertest/evidencelayertest.go` only if a copy's shape needs a new builder. `docs/ideas/structural-debt.md` (drop the entry this replaces).
Failing test first: none (test-only refactor); every existing test keeps passing with its assertions unchanged.
Do: each fixture embeds `*evidencelayertest.Records` or calls its methods, and its local copies of `prop`, `term`, `cite`, `participation`, `relationship`, `newSource` and the like are deleted. Find copies by body, not name: a helper that looks up a Property or term by key, or calls `create_cited_bridge` through the `connect` path, is a copy whatever it is called. Where a copy's signature differs (no `src, art`, or a role fixed), rewrite the call sites to pass `f.Source, f.Artifact` or keep a one-line local adapter. Where a helper truly differs in what it does, keep it and name why in your report. An in-package test (`package connect`, `package citations`, `package subjects`) cannot import `evidencelayertest` and stays. Stop and report if a move would create an import cycle in an external test package. Do not change `promotealign`'s `fixture` field duplicates (MNT-9, declined).
Check: `grep -rnE 'func \(\w+ \*?\w+\) (prop|term|cite|participation|relationship|newSource)\(' core --include='*_test.go'` prints only what your report lists, each with its reason; `grep -rln 'create_cited_bridge' core --include='*_test.go'` likewise; scripts/check-pr.sh PASS.

**F-4 — Say which package builds each side in the matching doc** (from MNT-1, CON-3)
Files: `docs/matching.md`.
Failing test first: none (doc).
Do: line 11, to match `add-match-rule` line 15: `core/database/evidencelayer` builds an Evidence graph's side with `subjectprofile.ForSubject`, memoized on the graph, and `core/database/promotealign` builds the canon's with `subjectprofile.Build`. Drop "once per proposal".
Check: `rg -n 'each side' docs/*.md` shows only the corrected line; scripts/check-pr.sh PASS.

**F-5 — One import reader for both guard checks** (from MNT-4)
Files: `core/matchregistry/guard_test.go`.
Failing test first: none practical without a tagged file in the tree; prove it by adding a `//go:build fts5` file to `core/database/evidencelayer` importing `core/graphalign`, seeing `TestReaderPackages` fail, and deleting it.
Do: extract `importPkg(t, root, dir)` holding the `go/build` context, `ImportDir`, and the `GoFiles` and `IgnoredGoFiles` failures; `checkPure` and `checkForbids` both call it.
Check: `TestReaderPackages` and `TestLayerPackages` pass; scripts/check-pr.sh PASS.

**F-6 — Test both polarities on one event** (from CON-5)
Files: `core/subjectprofile/lifeevent_test.go`.
Failing test first: add the row `[negative death, positive birth]` → `EventBirth`. It passes on today's body; that is expected, since the row pins moved code.
Do: add the row.
Check: `TestLifeEventOf` passes; scripts/check-pr.sh PASS.

**F-7 — Record the multi-graph load and the cache cap against S13-27** (from PRF-1)
Files: `docs/ideas/structural-debt.md`, `docs/deployment-plan/spike-13/deployment-plan.md`.
Failing test first: none (doc).
Do: add the debt entry the plan's Out line names: a multi-graph load, and `maxEvidenceGraphs` (64) against a person in more Evidence graphs than that, whose per-graph `Load`s would evict the page's own graphs and memos. In S13-27's map scope entry add one sentence: its budget subtest grows past 64 graphs and its read is checked against `maxEvidenceGraphs`.
Check: `grep -n 'maxEvidenceGraphs' docs/ideas/structural-debt.md docs/deployment-plan/spike-13/deployment-plan.md` prints both; scripts/check-pr.sh PASS.

**F-8 — Add the plan's Landed note** (from CON-2)
Files: `docs/deployment-plan/spike-13/prs/S13-29-one-loader-for-comparable-values.md`.
Failing test first: none (doc).
Do: add a **Landed note** under the header, in the shape of `S13-26-scoring-two-evidence-graphs.md:18`. List every difference: no second worktree for the `main` budget run; perl in place of GNU sed; STEP-3's `rg` printing `connectrules.Bridge{}`; the dropped `graphfill` blank import; the S14-36 line in `structural-debt.md`; the walk cap, which corrects the plan's Untrusted size line; `MemoSnapshot` deleted, superseding D-12; the other fixtures moved, superseding the Out row. Do not edit the Decisions or the Out row. Update the PR body's deviations to match.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `921f8f13`. F-1 to F-8 done. F-1's largest walk in `TestLargeSourceBudgets`' fixture is 60 edges, so `maxWalkEdges` (50,000) sits about 800× above it. F-2's mutant fails the rewritten test. F-3 moved `grapheventtitles`, `canonicalgraph`, `conclusionsummaries`, `promote`, `graphcache`, `autoreconciler`, `conclusiondetails`, `persondetail` and `search` onto `evidencelayertest.Records`, added `Records.Labelled`, and made `Locate` return the bridge; helpers kept with a reason: `promote/bridges_test.go`'s `connect.Result` family, `conclusionsummaries` `placeRel`, the `cite`s and thin `source` wrappers that pick the Artifact from the Subject's Source, `observations_test.go`'s `booleanFixture.cite`, `canonlinks`' `relationship`. Deviations: F-7 extended the existing S13-16 PRF-2 debt entry rather than adding a second; `completed.md` gained one sentence on the cap.

## Round 2 — review
Base `986d304f`, head `20f4ad41`, scope `c9462e27...20f4ad41`. Reviewers: CON, MNT, SEC+PRF (each had a Fix-now blocker or major). Skipped: UX (still no matching path).

Round 1 follow-through: CON-1, CON-2, CON-3 / MNT-1, SEC-1 **Resolved**. MNT-3, MNT-4, CON-5, PRF-1 **Resolved** (read in the fix commits; CON checked them by mutant or test too). MNT-2 **Not resolved**: exact copies remain in files F-3 moved.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-2 | SEC+PRF | blocker | Associations per link × children per parent run `Load` to 21 GB on 32,002 rows, under the walk cap | Fix now → F-9 | Read `kinship.chains`: each grandchild's chain copies the up link's associations; `set.Add` merged them first, so `MaxPairSteps` sees one step, and `MaxPivotEnd` limits parents only. The memo holds the result. Canon callers run `Derive` per pivot (`canonlinks/reader.go`), so a per-`Derive` total cap is safe there. One high total cap, per the user's S13-26 SEC-1 answer |
| MNT-2 (remainder) | MNT | major | `search/handles_test.go` `locatedBirth`, `autoreconciler` `mustProp`, `conclusiondetails` `termID` and hand-written `propertyterms.Lookup`s are copies F-3 left, with no debt entry | Fix now → F-10 | Reviewer proved the `search` swap passes under an overlay. Rides on Fix plan 2 because round 1's major must close before the PR is ready; it is not a new round-2 finding |

Nits (logged): `CON-8 · evidence_graph_walk_test.go · TestWalkEvidenceGraphEdgeCap has no filtered case and does not pin the exact cap` · `CON-9 · plan Landed note · kept list omits canonlinks' relationship and search's bridge calls (F-10 settles it)`

Note: CON saw `TestLargeSourceBudgets/warm` at 470 ms against 375 ms in one run that overlapped its mutant builds; three runs alone passed, and so did the verify run. The fix commits add one O(reaches) count to the walk.

### Fix plan 2

**F-9 — Bound what one Derive's chains copy before it builds one** (from SEC-2)
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`, `core/matchregistry/kinship.go`, `docs/matching-model.md`. Skill: `add-match-rule`.
Failing test first: `TestDeriveChainWorkCap` — one grandparent-to-middle link holding A associations and one middle with C children, A × C over the cap → no chain implied; just under → every chain implied. Today the over-cap case implies C chains.
Do: add `MaxChainWork int` to `kinship.Rules`, doc in the shape of `recordlinks.Rules.MaxPairWork`: the most association copies one Derive's chains may make before they read as malformed and imply nothing; counted before any chain. In `chains`, first run the same loops as the build (same `maxPivotEnd` and `bytes.Equal` skips) and sum, per chain, `len(Associations)` of every kept step of both pairs, read from `idx.byPair` without building `from`. Over the cap, return nil. Then build. Pass the new rule from `Derive`. Set `MaxChainWork: 100_000` in `defaultKinship`, with a `decided: user, 2026-11-07 (S13-26 SEC-1 rule, S13-29 SEC-2)` comment: one high total cap for a cost only a crafted catalog reaches. Add its row to the knob inventory in `docs/matching-model.md` beside `kinship.MaxPairSteps`, and run the matching skill's guard and any registry test that walks the knobs. State the largest count `TestLargeSourceBudgets`' fixture reaches in your report (a throwaway run, not committed).
Check: the test passes; `TestDeriveBound`, `core/matchregistry` and `core/database/canonlinks` tests pass; `TestLargeSourceBudgets` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes; scripts/check-pr.sh PASS.

**F-10 — Move the last builder copies onto the shared records** (from MNT-2)
Files: `core/search/handles_test.go`, `core/database/autoreconciler/autoreconciler_test.go`, `core/database/conclusiondetails/conclusiondetails_test.go`, `core/database/conclusiondetails/placerecords_test.go`, `core/database/promote/promote_test.go`, the plan's Landed note.
Failing test first: none (test-only refactor); assertions unchanged.
Do: `locatedBirth` calls `f.Participation(f.Source, f.Artifact, person, event, "subject")` and `f.Locate(f.Source, f.Artifact, event, place)`; drop imports left unused. `mustProp` goes (use `f.Prop`). `termID` goes (use `f.Term("sex_at_birth", key).ID`). `propertyterms.Lookup` at `handles_test.go:200,218,347` and `placerecords_test.go:41` become `Term`. Try `promote_test.go:260` `newCatalog` on `evidencelayertest.New`; if it cannot move, name why. Extend the plan's Landed note: the `MaxChainWork` cap, which corrects the Untrusted size line again; the kept list gains `canonlinks`' `relationship` and anything F-10 keeps. Update the PR body's deviations to match.
Check: `grep -rn 'mustProp\|termID(' core --include='*_test.go'` and `grep -rn 'propertyterms.Lookup' core --include='*_test.go'` print only what your report lists, each with a reason; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `38da3443`. F-9 and F-10 done. F-9: `kinship.MaxChainWork` 100,000; `chains` and its count share one walk, `eachChain`; the largest count in `TestLargeSourceBudgets`' fixture is 4; `TestDeriveManyAssociations` dropped from n = 20,000 to 4,000 to stay under the shipped cap, assertions unchanged; `TestDeriveChainWorkCap` was written with the field, so it had no red run. Read by the orchestrator: the count runs before any chain, and canon callers run `Derive` per pivot. F-10: the named copies are gone and `promote`'s `newCatalog` moved onto `evidencelayertest.New`; 19 one-off `propertyterms.Lookup` calls in `citations`, `connect`, `deleteimpact`, `effects`, `observations`, `subjects` and `subjectvocab` tests stay, named in the Landed note. No round 3: the fix commits for F-9 and F-10 had no fresh review.

## Result
Gate PASS on `38da3443` (orchestrator) · budgets pass on the final head (warm proposal 96 ms of 375 ms) · CI green (Go tests, macOS tests on `f9209ccd`) · open: none · follow-ups: none (debt went where the plan named: `structural-debt.md` S13-16 PRF-2 entry, S13-27 Check) · to user: —
