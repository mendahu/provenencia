# Ship log — S16-13 Per-graph records in the cache registry, and the Sources section reads them

**Plan:** [`S16-13-records-tier.md`](S16-13-records-tier.md) · **Branch:** `s16-13-records-tier` · **PR:** #809
**State:** ready

## Round 0 — implement
Base `cd64512ce` (main).
Head `3d38d885f` (two commits; STEP-2..7 uncommitted in the tree, compiling, Go tests green). Gate not run.

**Stopped on** §11: "`BenchmarkCacheSizes`' kinship figure grows by more than 20%". D-1's `container/list` LRU breaks it:

| Variant | kinship-B @4,000 | kinship-B @16,000 |
| --- | --- | --- |
| Planning base | 1,441,264 | 4,394,656 |
| D-1 as written (`container/list`) | +20.4% | +34.6% |
| `container/list`, shared key string (in tree) | +11.4% | +24.1% |
| Intrusive doubly linked list in `keyedValues[T]` | −6.6% | +3.2% |

Deterministic across three runs. Intrusive variant saved by the implementer in the scratchpad (`cache.go.intrusive`).

Other deviations to confirm with the plan fix:
1. `promotealign/export_test.go` `EmptyShortlistForTest` helper now also mints a generation (`indexGen.Add(1)`), so the new `candidates` Check does not fail before `matchindex` in `TestVerifyChecksMatchIndex`. Test outside §8's list.
2. Planning base measured on `main` @ `cd64512ce` in the tree, not a worktree at `aeb2773` (no worktree unasked).
3. T-14 lives in `promotealign/memo_test.go` (needs unexported `candidates`).
4. AC-1.6: no `value_subject_label` on the wire; label arrives as `value_text`, covered by existing `TestListPersonClaims`.
5. `graphfill.EvidenceGraph` invalid-id error text prefix `graphfill:`.
6. `TestLargeSourceBudgets` missed by 3–8% only in a full parallel run; passes alone.

Not done: STEP-8 docs (`docs/catalog-caches.md` rows; `TestCacheDocListsEveryEntry` fails until then), landing records, commits/push/draft PR, final `rg` sweep (`docs/promote-graph-alignment.md:165`, `docs/ideas/structural-debt.md:19`, `docs/ideas/person-page-performance.md:21`).

**Plan amended** (user, 2026-11-12: "Amend D-1, resume"): D-1 now holds the use-order links in the item; §11's 20% bound stays. Recorded in the review log. Resuming round 0.

### Round 0 resumed
Head `21b2e183e`. Gate PASS (orchestrator, against `cd64512ce`). Kinship with in-item links: −6.6% / +3.2%. Claims cold 13.2 ms / warm 0.18 ms (base 14.7 / 13.9); 80-graph record links 93 ms (base 113). Deviations still open: `EmptyShortlistForTest` mints a generation (needs user OK, §11); T-14 location; base measured in place; `TestLargeSourceBudgets` 3–8% miss in one parallel run only; manual check not run.

## Round 1 — review
Reviewers: CON, MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, `core/apperr`, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, MNT-2 | CON, MNT | major | Docs still describe the deleted memo API / store | Fix now → F-2 | verified: `rg` hits `promote-graph-alignment.md:150,168,169`, `matching.md:11`, `cache-architecture.md:54`, `ideas/person-page-performance.md:21`, `ideas/structural-debt.md:19` |
| MNT-1 | MNT | major | Layer rebuilt three ways; exported `Build` re-adds D-8's pattern param | Fix now → F-1 | verified: `evidencelayer.go` `Load`/`checkLayer` duplicate; `canon.go` `checkCandidates` calls `evidencelayer.Build` with `shortlistInputs`' pattern; two `graphalign.Layer{…}` literals |
| MNT-3 | MNT | minor | `catalog-queries.mdc` says reads only via `Get`/`GetMany`; PR adds `Peek` readers | Fix now → F-3 | verified: rule line 18; `PeekEvidenceGraph`, `canon.go` Peek |
| PRF-1 | PRF | minor | `budgetLinks80` = 1.5 s can't catch growth | To user → answered: set 200 ms, drop D-15 floor → F-4 | user 2026-11-12: "Set 200 ms" |
| CON-2 | CON | minor | `EmptyShortlistForTest` mints a generation (§11) | To user → answered: keep → F-5 (record only) | user 2026-11-12: "Keep it" |
| CON-3 | CON | minor | "Check is not a use" tests catch a regression ~50% | Fix now → F-6 | verified by reading the two-key test shape |
| CON-4 | CON | minor | T-1b doesn't prove cap override survives Drop | Fix now → F-6 | verified: steps read only two graphs after Drop |
| PRF-2 | PRF | minor | >64 graphs: record links reload every graph | Follow-up (already homed) | Plan Out line; `docs/ideas/structural-debt.md` S13-16 PRF-2 row (narrowed to record links + layers by F-2) |
| CON (outside lens → PRF) | PRF | minor | Cold claims read builds `evidencegraphs.IndexOf` | Rejected | `BenchmarkGraphIndex` at Scale: 60.8 ms build, once per catalog open; cold budget 1 s |

Nits (logged): `MNT-4 · graphfill/records.go RecordsOf · redundant !eg.Found branch` · `MNT-5 · persondetail/claims.go · ClaimID == nil implied` · `MNT-6 · graphcache/graph.go:28-31 · capsForTest comment placement` · `MNT-7 · promotealign/export_test.go · one CandidatesForTest accessor` · `MNT-8 · graphs_test.go, memo_test.go · needless SetCapForTest cleanup` · `CON-5 · records_test, TestLayersEntry, TestEntriesFollowSignals · not table tests` · `CON-6 · memo_test.go · T-6 cases in a new test; note in Landed`

### Fix plan 1

**F-1 — One fresh layer build; unexport `Build`** (from MNT-1)
Files: `core/database/evidencelayer/evidencelayer.go`, `core/database/promotealign/canon.go`, `core/database/promotealign/promotealign.go`.
Failing test first: none (refactor; existing `TestLayersEntry`, `TestVerifyLeavesEntriesAlone`, `TestCandidatesMemoFollowsCommits` and the Promote goldens must stay green, also with `PROVENENCIA_GRAPH_VERIFY=1`).
Do: add unexported `layerOf(g, eg)` (not `Found` → `ErrNoGraph`, `nameformats.LoadFor(g).Key`, build) used by `Load`'s fill; add `evidencelayer.Fresh(g, graphID) (Layer, error)` = `graphfill.LoadEvidenceGraph` + `layerOf`, storing nothing; `checkLayer` = `Fresh` then `DeepEqual`; `checkCandidates` calls `evidencelayer.Fresh` instead of `LoadEvidenceGraph` + `Build` (and takes no name pattern from `shortlistInputs` for the layer); unexport `Build`. Add one `promotealign` helper turning an `evidencelayer.Layer` + metas into `graphalign.Layer`, used by `loadInputs` and `checkCandidates`.
Check: `rg -n 'evidencelayer\.Build|LoadEvidenceGraph\(' --type go core api` outside tests prints only `evidencelayer.go` (inside `Fresh`); `rg -n 'graphalign\.Layer\{' --type go core` outside tests prints one line (the helper). scripts/check-pr.sh PASS.

**F-2 — Finish the docs sweep** (from CON-1, MNT-2)
Files: `docs/promote-graph-alignment.md` (§7 item 1 line ~150; §8 bullets ~168–169), `docs/matching.md:11`, `docs/deployment-plan/spike-16/cache-architecture.md:54`, `docs/ideas/person-page-performance.md:21`, `docs/ideas/structural-debt.md:19`.
Do: rewrite to the entries: loaded graphs are `graphfill.EvidenceGraph`; `evidencelayer.Layers` and `promotealign.candidates` are keyed by graph, derived from the vocabulary and `nameformats.Default`, dropped on `Set.EvidenceGraphs`, LRU at 64; candidates used only at the current index generation. `matching.md`: "held in `evidencelayer.Layers`". `cache-architecture.md`: replace `memberGraphs` with "the claims read". `person-page-performance.md`: a keyed registry entry with `DropNamedGraphs` (as `evidencelayer.Layers`). `structural-debt.md:19`: new names; the claims read no longer loads graphs; only record links and layers past 64 remain.
Check: `rg -n 'EvidenceGraphMemo|SetEvidenceGraphMemo|memberGraphs|maxEvidenceGraphs|g\.EvidenceGraph\(|memoized on the graph|Evidence graph memo' docs --glob '!docs/deployment-plan/archive/**' --glob '!docs/deployment-plan/spike-16/prs/**'` prints only `deployment-plan.md:334,426` and `person-page-audit.md:113` (history: map scope entries and the audit). scripts/check-pr.sh PASS.

**F-3 — Rule names `Peek`** (from MNT-3)
Files: `.cursor/rules/catalog-queries.mdc:18`.
Do: "reads only through its `Get`, `GetMany` or `Peek` (each counts as a use; a Check reads none of them)".
Check: scripts/check-pr.sh PASS (guidance-names).

**F-4 — `budgetLinks80` = 200 ms; D-15 amended** (from PRF-1, user 2026-11-12)
Files: `core/database/persondetail/recordlinks_perf_test.go`; plan D-15 and Landed note; review log amendment entry.
Do: `budgetLinks80 = 200 * time.Millisecond // measured: S16-13 planning base 113 ms × 1.25, rounded up`. Drop "no lower than `budgetLinksPast`" from D-15; add an amendment line to the review log citing the user's answer ("Set 200 ms").
Check: `TestRecordLinksBudgets/80 graphs` passes alone and in scripts/check-pr.sh.

**F-5 — Record the user's OK on `EmptyShortlistForTest`** (from CON-2, user 2026-11-12: "Keep it")
Files: plan Landed note, review log.
Do: no code; note the answer in the Landed note's deviation and in the review log amendment.

**F-6 — Tighten the LRU tests** (from CON-3, CON-4)
Files: `core/database/graphcache/cache_test.go`, `core/database/promotealign/memo_test.go`, `core/database/graphfill/graphs_test.go`.
Failing test first: none (strengthening; mutation check below).
Do: "a Check and KeyCounts are no use" and `TestVerifyLeavesEntriesAlone`: override cap to 64, Set keys 1..64, `g.Verify()` + `KeyCounts`, Set key 65, assert key 1 gone and 2..65 held. `TestEvidenceGraphCap/"a drop clears the loaded set"`: after the Drop, read graphs 1, 2, 3 at cap 2 and expect graph 1 not loaded.
Check: temporarily make Check touch keys (local edit, not committed) and see the T-1 case fail; revert. scripts/check-pr.sh PASS.

## Round 1 — fix
Head `387105af5`. Gate PASS (orchestrator; Go results cached; implementer's first gate run failed `TestRecordLinksBudgets/80_graphs` at 204.7 ms vs 200 ms, rerun PASS). F-1..F-6 done.
Deviations: F-1 `rg` also prints `graphfill/load.go` (the definition); `TestVerifyLeavesEntriesAlone` uses six real graphs at cap 6 (a touching Check is missed 1 in 6), no mutation check on it.
PRF-1 (F-4): read by orchestrator — `budgetLinks80 = 200 ms`, D-15 floor dropped, review-log amendment present. **Resolved**, but the budget's margin under parallel load is being measured (see below).

## Round 2 — review
Reviewers: CON, MNT (both had Fix-now majors). SEC+PRF not sent back (minors only; PRF-1 resolved above).

**MNT round 2:** Clean. MNT-1 and MNT-2 Resolved (one build path via `Fresh`/`layerOf`, `build` unexported, `alignLayer` the one `graphalign.Layer` literal; F-2 sweep prints only allowed history lines). MNT-3 Resolved. No new findings.
Nits (logged): `MNT-9 · promotealign/memo_test.go:259 · wrap the long comment line`

**CON round 2:** Clean. CON-1 Resolved. F-1 keeps the Check's name pattern (both via `nameformats.LoadFor`); F-6 tests verified; targeted packages green with and without `PROVENENCIA_GRAPH_VERIFY=1`. CON-3, CON-4 Resolved.
Nits (logged): `CON-7 · plan D-9 (line 170) · still names evidencelayer.Build; note F-1's Fresh in the Landed note`

No Fix plan 2 (no round-2 blockers).

### Budget check (orchestrator)
Uncached `go test -tags fts5 ./api/... ./core/...` ×3: `TestRecordLinksBudgets/80_graphs` passed every run; alone 93 / 109 / 95 ms against 200 ms. `TestLargeSourceBudgets/warm` missed once under parallel load (406 ms vs 375 ms); alone ×4, head 120–134 ms vs base `cd64512ce` 120–126 ms, so not a regression from this PR.

## Finish
Merged `origin/main` (`71773c1be`, docs only). Gate PASS against `71773c1be`.

## Result
Gate PASS · CI green (runs 43415894164, 43415894149 on `a6d87c341`, after F-7) · open: none · follow-ups: PRF-2 (already homed in `docs/ideas/structural-debt.md`) · to user: PRF-1 (answered: 200 ms), CON-2 (answered: keep)

## CI — round 1 (red)
Run 43364250762 (head `add73b9c4`): Go tests failed on three timing budgets only; macOS tests and localizable green.

| Test | Budget | CI | Local head | Local base |
| --- | --- | --- | --- | --- |
| `TestRecordLinksBudgets/80_graphs` | 200 ms | 414 ms | 93–109 ms | 113 ms |
| `TestLargeSourceBudgets/warm` | 375 ms | 383 ms | 120–134 ms | 120–126 ms |
| `TestLargeSourceBudgets/compose_every_summary` | 600 ms | 679 ms | — | — |

User (2026-11-13): "we can increase the budget on all three of those." New budgets = CI measurement × 1.25, rounded up to the next 100 ms.

### CI fix plan
**F-7 — Raise three budgets to the CI runner** (CI run 43364250762; user 2026-11-13)
Files: `core/database/persondetail/recordlinks_perf_test.go`, `core/database/promotealign/perf_test.go`; plan D-15 and Landed note; review log amendment.
Do: `budgetLinks80 = 600 ms` (CI 414 ms × 1.25 = 518, rounded up; the planning base on a Mac was 113 ms). `budgetWarm = 500 ms` (CI 383 ms × 1.25 = 479). `budgetComposeAll = 900 ms` (CI 679 ms × 1.25 = 849; still fails a per-row place graph at about four times the cost). Extend each budget's existing comment with the S16-13 CI figure and "raised by the user on 2026-11-13", in the style of the comment above `budgetWarm`. Update D-15's figure and the Landed note's `budgetLinks80` line; add a review-log amendment quoting the user.
Check: the three tests pass alone; scripts/check-pr.sh PASS; push.
F-7 addition (user, 2026-11-13): each budget comment records its raise history and says that creeping toward the budget again is a slowdown to investigate, not a reason to raise it.

## CI fix — done
Head `7d9d36d1d`. Gate PASS (orchestrator). F-7 done: `budgetLinks80` 600 ms, `budgetWarm` 500 ms, `budgetComposeAll` 900 ms; each comment records its history and that creep toward it is a slowdown to investigate. Deviation: commit `c6e334ec6`'s review-log amendment misquoted the user (the F-7 title); reworded in `7d9d36d1d`.
