# Ship log — S14-40 Build the input index

**Plan:** [`S14-40-build-the-input-index.md`](S14-40-build-the-input-index.md) · **Branch:** `s14-40-build-the-input-index` · **PR:** —
**State:** implementing (resumed at STEP-7 after the D-13 amendment)

## Round 0 — implement
Base `main` @ `78f26812` (plan was cut against `26ada55`). Local head `f6fc77d6`, unpushed, plus uncommitted T-9/T-10 in `core/database/promotealign/perf_test.go` and the single-graph isolator in `conclusionsummaries/conclusionsummariestest/inputs.go`. Gate not run. STEP-1 through STEP-6 committed; STEP-7 stopped; STEP-8 and STEP-9 not started.

**Stopped on.** In `TestLargeSourceBudgets`, the "inputs cover Dependents" subtest (T-10) fails with 9,987 misses. Every miss is a person that `Dependents([place])` returns, but whose recorded inputs do not include that place.

Orchestrator check:
- `conclusionsummaries/dependents.go:191` `sqlBirthOrDeath` matches an event on its kept `event_type` term (birth or death) only. It does not check whether the event has an accepted member.
- The person row's `lifePicks` keeps only events where `Node.Canonical` is true (`graphcache/graph.go:170` requires `len(AcceptedMembers()) > 0`).
- `bulkCanon` in `promotealign/bench_test.go` builds the fixture's events with no accepted `identity_claims`. The implementer spot-checked two events (`accepted=0`), not all 9,987.

So `Dependents` reports more dependents than the composers actually read. The index looks more precise, not short. T-2, T-3 and T-9 (the isolated-composition oracle on all 8,117 persons, 10,070 events and 52 places) pass.

Deviations so far:
1. Base drift: perf helpers moved to `databasetest.Fastest`, `Since` and `WithinBudget`.
2. New package `conclusionsummaries/conclusionsummariestest`, shared by T-2/T-3 and T-9/T-10.
3. `InputScope.Node` and `Follow` call the exported graph methods, so the recorder also hears reads made through `NoInputs()`.
4. `writeInputs` uses `INSERT OR IGNORE` in 200-row statements.
5. `lifeOwners` takes an `*InputScope`.

Provisional timings (noisy machine):
- First part_of from the city: 2.25 s → 1.22 s.
- Compose all rows: 146 → 188 ms (budget 500 ms).

**To user.** How should D-13 treat birth and death events that have no accepted member?

**User (2026-11-08):** "Exclude them in D-13". The plan's D-13, AC-1.3, T-10 and Stop condition 2 are amended. A person missing for X is excused when every birth or death subject event of theirs that `Dependents([X])` returns has no accepted member. Excused persons are counted in the log. Any other miss is still a stop. Committed with this log as "Amend S14-40's D-13 to excuse non-canonical birth and death events."

## Round 0 — resume
