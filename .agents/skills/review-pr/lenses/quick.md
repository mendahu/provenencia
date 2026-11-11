# Quick lens

The question: **is this small fix right for every input that reaches it, proven
by its test, and no bigger than it says?** One reviewer runs this lens on a
[`quick-pr`](../../quick-pr/SKILL.md), in place of the five lenses. It
borrows the checks from the others that a small diff can still get wrong, and
leaves out what only a feature can get wrong.

There is no plan. The PR body is the plan: its **Problem**, **Cause**, **Fix**,
**Test** and **Not in this PR** lines are what you check the code against.

## Read first

- The PR body (the orchestrator gives the file or the PR number).
- The diff, `git diff <base>...<head>`, and `git log <base>..<head>`.
- Each changed function in full, its **callers** (`rg -n "<Name>\("`), and its
  existing tests.

Keep to that. A quick review reads the change and its neighbors, not the
package's history or the spike's docs. Open a domain doc only when the fix
changes a rule it states.

## How to work

1. Restate the bug as an input and a wrong result. Find the line in the old
   code that produced it. If the PR body's **Cause** is not that line, the fix
   may treat a symptom: that is a finding.
2. Read the test. Does it assert the reported wrong result, so it fails on the
   old code? Check the assertion against the old code by reading it, or run the
   test after checking out the old version of the changed non-test files in a
   scratch worktree (`git worktree add <scratch> <base>`, copy the test in) and
   remove the worktree after. Do not edit the PR's checkout.
3. For each changed function: list its callers and say whether the change holds
   for each. A changed default, ordering, nil or empty handling, or error is
   where a small fix breaks a caller.
4. Probe the inputs next to the reported one: empty, one, many, the boundary
   the fix moves, a value the app did not write, a catalog made before this
   change. A claim that holds for all inputs gets a short seeded probe or
   table-driven test run, not only the reported case.
5. Run the **path checks** below for the paths the diff touches.
6. You may run targeted Go tests (`CGO_ENABLED=1 go test -tags fts5 ./core/<pkg>/...`).
   Do not run `xcodebuild`. Do not edit files.

## Look for

### Correct and complete

- The fix covers the reported case and not the rule: a second caller, a sibling
  type, or the other branch still computes the wrong result.
- A fix in a caller when the function that owns the behavior is wrong (a patch
  where the layer that owns it should change).
- An off-by-one, a flipped comparison, or a boundary moved too far.
- A new early return or guard that skips work the old path did (an audit
  effect, a cache invalidation, a cleanup).

### Proven

- No test, or a test that passes on the old code.
- A test that checks that the code runs, not the value the researcher saw.
- A test that only covers the reported example when the bug is a rule.

### Scope

- A hunk that traces to nothing in **Problem** or **Fix**: a rename, a tidy, a
  reformatted block. Scope creep is a minor; scope creep that changes behavior
  is a major.
- A change that breaks the quick-pr fit list (a migration, an FFI change, a new
  package, a view that looks different). Rate it **major** with
  `Fit: escalate`, so the orchestrator takes it to `plan-pr`.

### Reuse

- A private copy of a helper, query, formatter or component that already
  exists ([`maintainability.md`](maintainability.md) Q1 and Q5): search by body
  shape as well as by name. Name the existing code.
- A second spelling of a concept that has one name in the code.

### Path checks

Run the matching block when the diff touches these paths. The paths are the
routing table in [`ship-pr`](../../ship-pr/SKILL.md) §3.

| Diff touches | Read and apply |
| --- | --- |
| `macos/App/**`, `Localizable.xcstrings`, `core/apperr`, an error code | [`ux.md`](ux.md) **Copy**, **Accessibility**, **Errors** |
| A security path in that table, or code that trusts the size or range of something it loaded | [`security.md`](security.md) **Look for** |
| A performance path in that table, a loop, a query, or main-thread work | [`performance.md`](performance.md) **Look for** |

Number every finding `QCK-`, whatever block it came from, and say the block
in the title when it is not this file's ("QCK-2 · major · Security: …").

## Severity in this lens

| Severity | Here |
| --- | --- |
| **blocker** | The reported bug is not fixed; a caller now breaks; data written wrong or lost; a security hole; a hang |
| **major** | The fix holds for the example but not the rule; no test, or a test that passes on the old code; a private copy of existing code; a change outside the quick-pr fit list |
| **minor** | A neighbor input untested where the code is right; scope creep with no behavior change |
| **nit** | Taste |

Rate the code as merged ([`SKILL.md`](../SKILL.md) "Rate the code as merged").
"It is a quick fix" never lowers a severity.

## Fit

Use the `Fit:` line from the finding format, with one more value:
`escalate`, for a finding whose right fix breaks the quick-pr fit list.

## Good finding shape

For a fix that added a missing-qualifier guard to `datevalues.InsertTx` only
(an illustration, not a known bug):

```markdown
### QCK-1 · major · The guard covers new dates but not edited ones
**Where:** `core/database/datevalues/datevalues.go:65`
**Plan:** PR body, Fix
**Problem:** The qualifier check runs on insert, but `UpdateTx` writes through `update` without it, so editing a Source date to `ABT` with no year still saves.
**Evidence:** `UpdateTx` (line 114) → `update` (line 126) never calls the new check; probe: update a stored date to qualifier `ABT`, empty year, read it back, row saved.
**Fix:** Move the check into `normalize`, which both paths already call, and test both.
**Fit:** this PR
**Confidence:** high
```
