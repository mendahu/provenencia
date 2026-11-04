# Review — S14-26 A folded nickname shows as a nickname

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (Grok 4.7), executability (Composer)

## To user
None.

## Writer's own calls
- Corrected map Check / bar examples from James+Jake to Jim+James and Bill+William because the product seed folds `jake` into `jacob`, not `james` (probe at planning base).
- Chose path-local `nickEdge` / `chainNick` (D-7) over alias-index BFS after M-1 showed Al→Alexandra would be retagged.
- `fold` returns `(folds, nickClass)` so hop class is not classified twice (M-5).
- Distinct nicks sort by folded `sources` then `rank` (D-11), not `total`.
- Minor version bump (visible feature).
- No `ReasonRetyped`; no Swift change.

## Round 1
### M-1 · blocker · A nickname walk over the alias index retags prefix folds the hop did not take
**Answer:** Fixed — D-1/D-7 classify hops with Compare+Expands via `nickEdge` beside `foldValues`; alias-index token BFS removed. Edges/T-3 pin Al+Alexandra / Ann+Annette formal-only.

### M-2 · blocker · Nick order reads `total`, which a folded value never has
**Answer:** Fixed — D-11: `foldAttach` carries `sources`/`rank`; sort matches `foldValues` sink tie-break.

### M-3 · major · The shipped chain rows are the neighbors
**Answer:** Fixed — STEP-2 / T-5 update `TestValueRulesFoldChains` with `nameSpec`; Edges lists Caroline/Chick/Charles, Mo/Molly/Mary, Fred+Bill+William, Bill beside Robert+William.

### M-4 · major · The 25% gate never loads the alias index
**Answer:** Fixed — Scale / T-7 `BenchmarkReconcileNamesAliases` under `aliasRules` with Bill/William/Jim/James/Chris/Christopher corpus.

### X-1 · blocker · `foldAttach` cannot carry support order
**Answer:** Fixed — same as M-2 (D-11).

### X-2 · major · Cache integration goldens are not in Steps or Tests
**Answer:** Fixed — STEP-2 / T-5 name `TestReconciledNames` Chris+Christopher and Ann+Anna+Hannah `nameSpec` updates.

### X-3 · major · `TestReconcileNamesAliases` update scope is under-specified
**Answer:** Fixed — STEP-2 updates every nickname-class fold row; abbreviation/initial/cycle/`sameResult` stay formal-only.

### X-4 · major · `T-3` does not pin how nick parts are asserted
**Answer:** Fixed — T-3 requires `nameSpec`, not `givenOf`.

### X-5 · major · Go checks could not be run in the reviewer environment
**Answer:** Declined for the plan — implementer uses Go 1.27 from the module cache. Planner verified probes with `go1.27.3`.

### X-6 · minor · STEP-4 partly duplicates map edits
**Answer:** Fixed — STEP-4 confirms map; required edit is `conclusion-reconciliation.md` §7.2.

### X-7 · minor · STEP-2 regression check is not self-contained
**Answer:** Fixed — T-7 relative gate vs stubbed no-op.

### X-8 · nit · Current state rows verified
**Answer:** Logged.

## Round 2
### M-5 · blocker · Nickname hops are classified a second time beside `nameModule.fold`
**Answer:** Fixed — D-1: `module.fold` returns `(folds, nickClass bool)`; `foldValues` sets `edge` and `nickEdge` from one call; `nicknameHop` deleted from the plan.

### X-9 · major · STEP-2 check never runs the DB integration tests
**Answer:** Fixed — STEP-2 Check and T-5 use `TestReconciledNames`.

### X-10 · minor · How to extend `TestValueRulesFoldChains` for nicks
**Answer:** Fixed — STEP-2 / T-5: assert full `nameSpec`, not `givenOf`.

### X-11 · minor · `BenchmarkReconcileNamesAliases` corpus shape
**Answer:** Fixed — Scale names the 200-candidate aliasRules corpus with Bill/William/Jim/James/Chris/Christopher.
