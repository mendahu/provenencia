# Review — S16-16 Kinship answers are composed from the pivots, and the Family read is one pass

Plan: [`S16-16-composed-kinship.md`](S16-16-composed-kinship.md). Lint: `scripts/check-plan-docs.py` clean before round 1.

## Round 1 (2026-11-13)

Maintainability (fresh, strongest model): **Fits with changes**. Executability (fresh, implementing model): **Not ready**.

### Maintainability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | `forLinks` derives the steps again beside `loadSteps`; the new `Relationships` would run both per id | **Fixed** — D-11 (new): `stepped.atEnd`; `forLinks(s stepped, chain)`; `For` = `loadSteps([id])` + `forLinks`; `around` goes. D-1, D-2, D-3, Builds on, Deletes, Scale updated |
| M-2 | major | `BenchmarkLargeSourcePropose` would be a third copy of the after-commit timing | **Fixed** — D-8: `largeFixture.warmProposal` and `proposalAfterCommit` in `perf_test.go`, called by `TestLargeSourceBudgets` and the benchmark; Codebase fit rows |
| M-3 | minor | Next adds for S16-24 should name the `nodeSource` seam | **Fixed** — Next adds names an `InputScope`-backed `nodeSource` |
| M-4 | minor | Single-id `Relationships` callers not listed | **Fixed** — §4 rows, Reach, STEP-2 Files, D-10 (`promotealign_test.go`, the direct test calls) |
| M-5 | nit | Map row 179 names `DropOnHandleCommit`, `dropHandleMemos`, already gone | **Fixed** — STEP-5 drops them from the row's scope cell |

### Executability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | STEP-2 cannot compile: the For and siblings entries share helpers | **Fixed** — the reader, its callers and the Family rewrite are one step (STEP-2), one commit, as S16-10 landed |
| X-2 | blocker | `TestRelationshipsOptionsSiblings` AC-1.9 and `fromNames` peek the deleted entry | **Fixed** — D-10: AC-1.9 asserts an empty map; `fromNames` reads `siblingsOf` on `loadSteps` (a merged `kinship.Link` carries no `From`, checked in `core/kinship/kinship.go` `Shape`) |
| X-3 | blocker | `promotealign_test.go` `TestProposeIgnoresStoredSiblings` calls `LoadSiblings` | **Fixed** — D-10 renames it `TestProposeIgnoresPageSiblings` and rewrites the setup; T-10 |
| X-4 | major | Interim `family.go` shape at STEP-3 undefined | **Fixed** — no interim: STEP-2 lands D-5 directly |
| X-5 | major | The `TestLargeSourceBudgets` stop has no step that checks it | **Fixed** — STEP-1 runs it on the base, STEP-2's Check runs it |
| X-6 | major | Benchmark loop and machine name unspecified | **Fixed** — D-8: ignore `b.N` as `BenchmarkCacheSizes`, warm then after-commit order, `hostname`, ship log path |
| X-7 | minor | T-2's dropped set loose | **Fixed** — AC-1.1 and T-2 name the set (`{Ann}` ∪ her node's neighbours: M1, B1) and assert equality |
| X-8 | minor | AC labels collide with existing subtest names | **Fixed** — the Stop condition names "the existing subtests named AC-2.2 / AC-2.3 (S16-01's labels)" |
| X-9 | minor | Notes, none blocking | Logged |

## Round 2 (2026-11-13)

Maintainability: **Fits with changes**, two new blockers in round-1 fix text. Executability: **Ready with changes**, no blocker; all round-1 answers confirmed against the code (Go 1.27 toolchain not available to the reviewer, so no check was run).

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-6 | blocker | Next adds pointed S16-22 at unexported `loadSteps`/`forLinks` and S16-24 at an `InputScope` `nodeSource` that records nothing warm | **Fixed** — S16-22 reads the batched `Relationships`; S16-24 follows the map (handles through `InputScope`, pivots from the loaded steps) |
| M-7 | blocker | D-11 had `loadSteps` inline `eventLinksAt` + `kinEdges` beside `parentEdges` | **Fixed** — D-11: `parentEdges` returns `(atEnd, edges, err)`, the one builder; Codebase fit row |
| X-10 | major | Merged query-count test shape open | **Fixed** — D-10 names the helper, the two subtests and the ids |
| X-11 | major | Helpers' graph and `propose` unspecified | **Fixed** — D-8 gives the helper's first line |
| X-12 | minor | T-2 one flow or two rows | **Fixed** — two rows, fresh world each |
| X-13 | minor | `filed` helper checks `hasSiblings` | **Fixed** — D-10 |
| X-14 | minor | Batched edge cases | **Fixed** — D-2 states empty, repeated and short ids |
| X-15 | minor | Base-fails case and heap noise | **Fixed** — STEP-4 Check and §11: base-failing rule; `kinship-B` ≤ 1.1 × base |
| X-16 | nit | Map status value | **Fixed** — `Landed #N` |

## Verdict

Planned. Two rounds; every finding Fixed; no To user item.
