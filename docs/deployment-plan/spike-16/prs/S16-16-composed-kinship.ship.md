# Ship log — S16-16 Kinship answers are composed from the pivots, and the Family read is one pass

**Plan:** [`S16-16-composed-kinship.md`](S16-16-composed-kinship.md) · **Branch:** `s16-16-composed-kinship` · **PR:** #814
**State:** fixed 1, awaiting review

## Round 0 — implement
Base `main` @ `aa150de0f`. Dispatched `pr-implementer`, round 0.
Local head `e883317d4` (2 commits, not pushed). Gate not run.

**Stopped on** §11: `warm-ms` at 16,000 is 231.6 vs base 137.9 (1.68×, limit 1.25×). Same M1, load 7–9; a re-run still slow.

| Measure | Base | After | Ratio |
| --- | --- | --- | --- |
| warm-ms 4,000 | 115.9 | 180.3 | 1.56 |
| warm-ms 16,000 | 137.9 | 231.6 | 1.68 |
| after-commit-ms 4,000 | 621.5 | 825.9 | 1.33 |
| after-commit-ms 16,000 | 1399 | 1289 | 0.92 |
| TestLargeSourceBudgets warm log | 131.8 ms | 258 / 199 ms | 1.5–1.96 |
| kinship-B 4,000 / 16,000 | 1,345,832 / 4,534,264 | 0 / 0 (HeapDelta artefact, needs re-run) | — |

Budgets still met. Cause as §6 Scale predicted: each warm proposal recomposes `For` per walked person. Done: STEP-1–3, part of STEP-5. Not done: docs rows, map, completed.md, gate, push, PR.
Deviation: `warmProposal` uses `databasetest.Warm` (main already had it), same behaviour.

**State:** stopped — memoising `For` is the user's call (plan §11, rule "No stored answer without a measurement").

**User, 2026-11-13:** accept the warm slowdown as the new normal. Plan §3, STEP-4 Check and §11 amended; memoising `For` parked in `docs/ideas/promote-ranking.md` § "Memoise a person's composed kinship". Raise `budgetWarm` to match: CI measured 383 ms before (S16-13) × the worst ratio seen (1.96) = 751 ms, × 1.25 headroom → 950 ms. Other budgets unchanged; a red CI on one of them is a gate failure with the CI figure as evidence.

## Round 0 — resume
Dispatched `pr-implementer` from STEP-4 (memory re-measure) onward.

Machine: `Jakes-M1-Macbook-Pro.local` (Apple M1), `-benchtime 1x`, one run each.

| Measure | Base (`42b032773`) | After | Ratio |
| --- | --- | --- | --- |
| warm-ms 4,000 | 115.9 | 201.2 | 1.74 |
| warm-ms 16,000 | 137.9 | 312.8 | 2.27 |
| after-commit-ms 4,000 | 621.5 | 919.5 | 1.48 |
| after-commit-ms 16,000 | 1399 | 1817 | 1.30 |
| TestLargeSourceBudgets warm | 131.8 ms | 141.8 ms (budget 950) | 1.08 |
| kinship-B 4,000 / 16,000 | 1,345,832 / 4,534,264 | 0 / 0 | — |
| evidence-B 4,000 / 16,000 | 34,491,240 / 44,955,880 | 28,430,544 / 38,494,000 | 0.82 / 0.86 |
| canonlinks.kinship-keys | 14,162 / 14,162 | 14,162 / 14,162 | 1.00 |
| canonlinks.for-keys, siblings-keys | 8,115 / 20,115 each | entries gone | — |

The machine was loaded (the earlier run's 180.3 / 231.6 warm figures are the same code), so timings vary run to run: the budget test's warm run read 141.8 ms.

**kinship-B 0 is real, not a HeapDelta artefact.** The `evidence` fill (a proposal of Source A) runs before the `kinship` fill and holds all 14,162 pivots; a temporary log showed `Relationships` over all 8,115 / 20,115 persons leaving the key count at 14,162. The base's 1.3 / 4.5 MB were the stored `for` and `siblings` answers, which no longer exist, and `evidence-B` fell by 6.1 / 6.5 MB for the same reason (its proposal stored `for` answers). The 1.1x memory check on the sum of the two fills: 35.8 / 49.5 MB before, 28.4 / 38.5 MB after (0.79 / 0.78). Passes.

Budget raised to 950 ms (`budgetWarm`). `PROVENENCIA_GRAPH_VERIFY=1` suites of canonlinks, persondetail and promotealign pass. check-pr: PASS.

Orchestrator gate on `4a5e60b38`: check-pr PASS (incremental).

## Round 1 — review
Base `aa150de0f`, head `4a5e60b38`. Reviewers: CON, MNT, SEC+PRF (`core/database/**`, `canonlinks` kinship over catalog data). Skipped: UX (no `macos/App/**`, strings, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | AC-3.3 sweep hits `canonlinks.for` / `canonlinks.siblings` in the new idea section | Fix now → F-1 | verified: the AC's `rg` prints `docs/ideas/promote-ranking.md:21` |
| CON-2 | CON | major | No `## Landed` note listing differences from the plan | Fix now → F-2 | verified: DoD line; S16-10/13/30 carry one |
| CON-3, SEC outside-lens | CON, PRF | minor | Idea file, plan and `completed.md` quote one run's figures; runs disagree (1.68× vs 2.27× warm; after-commit 0.92× vs 1.30×) | Fix now → F-1 | verified: two ship-log tables; PRF re-measured after-commit at 4,000 alternating base/head: best 625 vs 630 ms, unchanged |
| CON-4 | CON | minor | T-6 "a commit elsewhere" passes on base; does not prove the catch-all drop is gone | Fix now → F-3 | reviewer ran head test on base: PASS |
| MNT-1 | MNT | minor | `Families` adds two first-seen dedupes its callees already do; `want` loop is `database.UniqueBlobIDs` | Fix now → F-4 | verified: `graphcache/load.go:27`, `canonlinks/reader.go` `loadSteps` dedupes; `sqlin.go:10` |
| MNT-2 | MNT | minor | `matching-model.md:31` still says `For` "answers from the graph cache's results" | Fix now → F-5 | verified by `rg` |

SEC+PRF: no findings. `edgeKey` probed with 2M random edges, no collision. Family read warm on a 4,000-child hub 15.7 → 28.7 ms (one child 21.6 → 44.7 ms); at 20,000 children 100 → 113 ms: linear in a fan-out already capped by S16-01 (`MaxKinPerRole`), noted, not a finding.

Nits (logged): `MNT-3 · reader_test.go TestComposedAnswersReusePivots · build returns []*[]byte, [][]byte enough` · `MNT-4/CON-5 · TestComposedAnswersReusePivots · T-2 asks a two-row table, code has two subtests` · `MNT-5 · perf_test.go:190 · two Source A graph handles in one test, say why` · `MNT-6 · promotealign_test.go TestProposeIgnoresPageSiblings · comment could say siblings are no longer stored` · `CON-6 · TestComposedAnswersReusePivots "commit ann" · explicit dropped ⊆ expected loop`

### Fix plan 1

**F-1 — Idea file and figures say what the runs say** (from CON-1, CON-3, PRF outside-lens)
Files: `docs/ideas/promote-ranking.md` § "Memoise a person's composed kinship", plan §3 Out row and §11 parenthesis, `docs/deployment-plan/spike-16/completed.md` S16-16 paragraph.
Do: Name no deleted cache entry: say "the stored per-person implied links and siblings answers", not `canonlinks.for` / `canonlinks.siblings`. Give warm as the range the ship log records on a loaded M1: 1.56–1.74× at 4,000, 1.68–2.27× at 16,000; point to the ship log for the tables. Say the proposal after a commit is unchanged within noise (best of 3/5 alternating base and head at 4,000: 625 vs 630 ms). Plan and `completed.md` quote the same range, not one run.
Check: AC-3.3's `rg -n "LoadRelationships|LoadSiblings|forAnswers|siblingAnswers|canonlinks\.for\b|canonlinks\.siblings|memo slot" core docs .agents .cursor -g '!docs/deployment-plan/**'` prints nothing; check-pr PASS.

**F-2 — Landed note** (from CON-2)
Files: the plan. Do: add `## Landed` after the header in the S16-30 style, listing every difference: `budgetWarm` 500 → 950 ms and the lifted warm-ms limit (user, 2026-11-13, memoise parked in the idea file); `warmProposal` uses `databasetest.Warm`; `TestComposedAnswersReusePivots` is two subtests, not a two-row table; the `canonlinks.kinship` size cell gives keys and "no heap figure of its own" because the evidence fill already holds every pivot (kinship-B 0, explained in the ship log); and the F-3 / F-4 changes below if they differ from the plan's text.
Check: plan-docs PASS.

**F-3 — T-6 proves the catch-all drop is gone** (from CON-4)
Files: `core/database/persondetail/family_query_test.go` "a commit elsewhere" (or a `canonlinks` test if `Peek` is only reachable there).
Failing test first: on the base this subtest passes; add the assertion that every `canonlinks.kinship` pivot key held before `w.k.Commit(stranger)` is still held after (D-7's `Peek`). Show it would fail with a catch-all drop (state how you checked: e.g. temporarily clearing the entry in Apply, reverted).
Check: test passes at head; check-pr PASS.

**F-4 — `Families` drops the redundant dedupes** (from MNT-1)
Files: `core/database/persondetail/family.go`.
Do: drop `isNamed` and append every named relative (`g.Nodes` dedupes, `graphcache/load.go:27`); drop `isAround`, pass every `Around` id, and fill `aroundKin` by ranging over the `Relationships` map (`loadSteps` dedupes). Replace the `want` first-seen loop with `database.UniqueBlobIDs(ids)`. No behaviour change: `TestPersonFamilyQueryCount`, `TestCraftedFamily` and the persondetail family suites stay green unchanged; run them with `PROVENENCIA_GRAPH_VERIFY=1` too.
Check: `rg -n "isNamed|isAround" core` prints nothing; check-pr PASS.

**F-5 — `matching-model.md:31`** (from MNT-2)
Do: reword to match `canon.go:56`: `For` composes the links from the graph cache's kinship pivots.
Check: `rg -n "answers from the graph cache's results" docs core` prints nothing.

Nits: fix MNT-6 and CON-6 (lines in tests this PR adds or changes) if they are one line each; leave the rest logged.

## Round 1 — fix
Dispatched `pr-implementer`, fix round 1. All five items applied as new commits.

- F-1: idea file, plan §3 Out row and §11, and `completed.md` quote the ship-log range (warm 1.56–1.74× at 4,000, 1.68–2.27× at 16,000; after a commit unchanged within noise) and name no deleted entry. AC-3.3's `rg` prints nothing.
- F-2: plan has a `## Landed` section listing the differences.
- F-3: `TestComposedAnswersReusePivots/commit a stranger` (in `canonlinks`, where `Peek` is reachable) asserts every held pivot survives a commit on an unlinked person. Checked by temporarily replacing the pivot Apply with `k.Clear(g)` (reverted): the subtest failed ("dropped the pivot of Bob"); it passes at head.
- F-4: `Families` drops `isNamed` and `isAround`, uses `database.UniqueBlobIDs`, fills `aroundKin` from the `Relationships` map. `persondetail` suites pass with and without `PROVENENCIA_GRAPH_VERIFY=1` (also `canonlinks`, `promotealign`).
- F-5: `matching-model.md` reworded.
- Nits: MNT-6 done (comment). CON-6 not done: the dropped ⊆ expected check is a loop, not one line; left logged.

Gate: check-pr PASS.
