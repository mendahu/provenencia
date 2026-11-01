# Review — S14-18 Name folds never crash

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1 (no blocker, so no round 2). **Reviewers:** maintainability (opus), executability (sonnet)

## To user
None. The B-3 question was put to the user while planning (2026-11-01): when one person's records hold two names of a nickname set, folding into the set's head is intended. B-3 left the PR; the map's working rule "A nickname folds into its set's formal name, as today" replaces audit 2's option a.

## Writer's own calls
- D-1: the target is the best sink a value **reaches** (most Sources, then lowest rank), not the first one found, and `into` is flattened to one hop.
- D-2: a fold cycle with no way out keeps each member as its own value (Dick, Ricky, Rick shows all three) rather than collapsing into one member.
- D-3: `CacheVersion` 22, although a crashed project would rebuild anyway, because a value with a direct target can now reach a better-supported one through a chain.
- Patch version bump (a crash fix).
- AC-1.5 accepts Caroline, Chick, Charles → **Charles**, which is what the unchanged fold rules give once the chain resolves.

## Round 1

### X-1 · major · Stop condition names a benchmark no step runs
§11 names `BenchmarkReconcileNames` +25%, but no step runs it and the quoted ms are not portable.
**Answer:** Fixed. STEP-1's check runs the benchmark with and without the step on the same machine (median within 25%); the §11 stop condition points at that check.

### X-2 · major · T-3 leaves the `foldValues` fixture to the implementer
**Answer:** Fixed, with M-4. T-3 builds fixed shapes by hand: unit from `split`, `rank` = index, `sources = {"s<index>"}`, named shapes and expected roots.

### X-3 · major · T-6 and AC-2.1 do not say how the generator builds a unit of one to three words
**Answer:** Fixed. T-6 says one to three separate `given=<word>` parts, then `surname=Robins`, ids `{0,0,0,i+1}`, no SourceID, default Provenance, never negative or provisional.

### X-4 · major · T-1, T-5 and AC-1.8 give expected parts but not ids, reasons or part order
**Answer:** Fixed. §2 states ids, input order, one `kept` value holding every candidate, strict part order and each Outcome's reason (`folded` / `kept`; all `kept` in the cycle). AC-1.8's third candidate is one candidate with two given parts. Every value was confirmed against the prototype at the planning base.

### X-5 · minor · `chainRules` is under-specified
**Answer:** Fixed, superseded by M-2: there is no second table.

### X-6 · minor · T-4 reuses the rename invariant under the alias table without saying it holds
**Answer:** Fixed. Probed at the planning base with the fix and the widened pools: every invariant, renaming included, passes under the chain sets. T-4 says so.

### X-7 · minor · Struct and function comments become false
**Answer:** Fixed, with M-1. STEP-1 rewords the `unitValue` comments and `elect`'s comment, and checks `rg -n maximal core/autoreconcile/*.go` has no non-test hit.

### X-8 · minor · T-7 does not say which Sources the three Subjects sit on
**Answer:** Fixed. Ann and Hannah on `f.subject`, Anna on `f.subjectOn(f.other, …)`.

### X-9 · nit · STEP-6 does not name the new version
**Answer:** Fixed. `0.2.5` → `0.2.6`, or the next patch at the implementer's base.

### X-10 · nit · Checks not runnable in the reviewer's environment
**Answer:** Logged. `check-pr.sh` needs Xcode; the Go checks in STEP-1 – STEP-4 run anywhere.

### M-1 · major · "Maximal", "sink" and "root" become three names for one idea
**Answer:** Fixed. **Deletes** retires "maximal" in `pipeline.go`'s comments; `into == nil` is a root everywhere; "sink" stays local to `foldValues`. The STEP-1 rewording and `rg` check are under X-7.

### M-2 · major · `chainRules` copies `aliasRules`, on a reason that is not true
**Answer:** Fixed. Checked: adding the six sets to `aliasRules` leaves every `TestReconcileNamesAliases` and `TestPrefix…` row passing. The chain sets go into `aliasRules`, T-1 / T-2 become rows in its `folds` table, and T-4 runs under `testRules` and `aliasRules`. D-4, Codebase fit, STEP-2 and §8 updated.

### M-3 · major · T-6 copies the name-invariants generator into a second package
**Answer:** Declined, with evidence. A shared `core/autoreconcile/autoreconciletest` package imports `core/autoreconcile`, so `names_test.go` (package `autoreconcile`, internal) cannot import it: Go refuses the cycle. `TestReconcileNamesInvariants` reads unexported `testRules`, `aliasRules`, `signatureOf`, `shape` and `moduleFor`. Sharing would need that test moved to an external test package and the test tables exported, which rewrites an existing test this PR does not own (Implementer rule 4). T-6 keeps three checks through the exported API only (no panic, one value per candidate, order-free); the internal test keeps the full set. The cost the reviewer names (an invariant added to one and not the other) is real but bounded: T-6 exists to run the shipped table, which the internal test cannot reach.

### M-4 · minor · T-3 rebuilds `reconcile`'s grouping step inside the test
**Answer:** Fixed. T-3 builds fixed shapes by hand; random coverage stays with T-4 and T-6, which would panic on a nil `total`.

### M-5 · minor · **Next adds** omits what flattening discards for S14-26
**Answer:** Fixed. **Next adds** says flattening keeps no path or fold kind, and that S14-26 asks the module about the (value, root) pair, or adds the path in its own plan.

### M-6 · nit · The planning probe is still in the working tree
**Answer:** Declined. The reviewer saw a transient probe. `git status` was clean before and after it, and the planning branch commits only docs. The header still says to start from a clean checkout.
