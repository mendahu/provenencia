# Ship log — S14-18 Name folds never crash

**Plan:** [`S14-18-name-folds-never-crash.md`](S14-18-name-folds-never-crash.md) · **Branch:** `s14-18-name-folds-never-crash` · **PR:** #677
**State:** ready

## Round 0 — implement
Head `31b35ba2`. Gate PASS (orchestrator run). PR #677 (draft). Report deviations: version bump 0.2.6 → 0.2.7 (VERSION already 0.2.6 at base; plan allows); §10 manual check not run (user's); T-2 reversed rows generated in a loop.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes; `project.pbxproj` changes only MARKETING_VERSION).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | T-7 adds a third copy of the name-parts-to-spec loop in `TestReconciledNames` | Fix now → F-1 | verified: `p.Type+"="+p.Value` at `autoreconciler_test.go:600, 624, 660`; same-pass cleanup |
| MNT-2 | MNT | minor | New `givenOf` sits beside the same inline loop in `TestValueRulesFoldNicknamesOnly` | Fix now → F-2 | verified: loop at `:1629`, helper at `:1690`; same-pass cleanup. AC-2.4's "unchanged" reads as rows and results: both stay as they are |
| MNT-3 | MNT | minor | `indexOf` rewrites `slices.Index` | Fix now → F-3 | verified: `pipeline_test.go:427`, new file code |
| MNT-4 | MNT | nit→minor | `into` comment says "the fuller value it folds into"; now a reached sink | Fix now → F-4 | stale comment in a changed file (`pipeline.go:64`) |
| SEC-1 | SEC | minor | Rewritten fold has no fuzz target | Fix now → F-5 | verified: no `func Fuzz` in `core/autoreconcile`; minor in code this PR changes; rubric asks for one |

Nits (logged): `CON-1 / MNT-5 · core/autoreconcile/names_test.go:417-425 · anonymous row struct written three times; a named local type says it once`
PRF: no findings (probe n=696 dense fold graph 3.84 s vs 4.83 s no-fold; cost still set by n² compares). CON: clean.

### Fix plan 1

**F-1 — One name-spec helper for `TestReconciledNames`** (from MNT-1)
Files: `core/database/autoreconciler/autoreconciler_test.go`.
Do: add `nameSpec(t *testing.T, raw …) string` (unmarshal with `valuecodec.UnmarshalName`, join parts `type=value` with `|`, the shape of `spec` in `core/autoreconcile/names_test.go:43`). Move the three subtests at ~597, ~621, ~657 onto it; expected strings unchanged.
Check: `grep -n 'p.Type+"="+p.Value' core/database/autoreconciler/` prints only the line inside `nameSpec`; `TestReconciledNames` passes; scripts/check-pr.sh PASS.

**F-2 — `TestValueRulesFoldNicknamesOnly` uses `givenOf`** (from MNT-2)
Files: same file.
Do: replace the inline `var given []string` loop (~1629) with `givenOf(res.Values[0].Value)`. Rows and assertions unchanged.
Check: `grep -n 'var given \[\]string' core/database/autoreconciler/autoreconciler_test.go` prints only the line inside `givenOf`; test passes.

**F-3 — Use `slices.Index`** (from MNT-3)
Files: `core/autoreconcile/pipeline_test.go`.
Do: replace `indexOf` calls with `slices.Index`; delete `indexOf`.
Check: `grep -rn "func indexOf" core` prints nothing; tests pass.

**F-4 — Fix the `into` comment** (from MNT-4)
Files: `core/autoreconcile/pipeline.go:64`.
Do: comment reads: the value it reaches through folds (a sink); nil on a root.
Check: comment updated; no code change.

**F-5 — `FuzzReconcileNames`** (from SEC-1)
Files: `core/autoreconcile/names_test.go` (or a new `fuzz_test.go` in the package).
Do: fuzz target building candidates from fuzzed bytes (candidate count capped small, word count, words drawn from `aliasRules` tokens plus free text, part types), reconciled under `aliasRules`. Reuse the existing invariant checks (`checkNameInvariants`) rather than rewriting them: no panic, each candidate in exactly one value, result unchanged by input order. Seed the corpus with the B., Bill, William and cycle shapes.
Check: `go test -run FuzzReconcileNames ./core/autoreconcile/` passes (seed corpus); `go test -fuzz FuzzReconcileNames -fuzztime 30s ./core/autoreconcile/` finds nothing (report output); scripts/check-pr.sh PASS.

## Round 1 — fix
Head `3e95a8bd`. Implementer gate PASS. F-1 to F-5 done. F-5 deviation: the fuzz run broke the rename invariant under `aliasRules` (Rick, Ricky, Ricky as `nick` renamed to `suffix` leaves the alias role, so the fold stops: correct behaviour). The implementer skipped the rename check whenever alias tables differ from `testRules`. Orchestrator decision: the skip contradicts T-4 ("every invariant included" under both rule sets), so it is replaced, not accepted. The rename check stays and, under alias tables, keeps the alias roles (given, nick, untyped) fixed and permutes the rest.

**F-6 — Role-preserving rename under alias tables** (from F-5 deviation; T-4)
Files: `core/autoreconcile/names_test.go` `checkNameList`.
Do: remove the early `return`. When `rules.Names.Aliases` applies (differs from `testRules`), use a rename map that keeps `given`, `nick` and `""` as they are and permutes only the roles alias tables do not key on (e.g. `surname`→`suffix`, `suffix`→`surname`, `prefix`→`x`); under `testRules` keep today's full map. Update the comment to say why.
Check: `TestReconcileNamesInvariants` (both rule sets) and the seed corpus pass; `go test -tags fts5 -fuzz FuzzReconcileNames -fuzztime 30s ./core/autoreconcile/` finds nothing; scripts/check-pr.sh PASS.

F-6 head `90a1d33b`. Gate PASS (orchestrator run). Deviation: `prefix`→`x` failed (unknown types read as untyped, an alias role), so `prefix` stays fixed and only surname↔suffix is permuted under alias tables; comment says so. Accepted: still a one-to-one rename over the roles alias tables ignore. 30 s fuzz: 461,959 execs, nothing found.

SEC-1: **Resolved** (orchestrator read `FuzzReconcileNames`, 3e95a8bd + 90a1d33b: seeds cover chain, cycle and free-text shapes; it reuses `checkNameList`). PRF: no findings, no re-review.

## Round 2 — review
Reviewers: MNT (had major MNT-1). Scope: `31b35ba2...90a1d33b`. CON and SEC+PRF not sent: no Fix-now blocker or major.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) third copy of name-spec loop | Resolved | `nameSpec` at `autoreconciler_test.go:1669`; grep leaves only its line |

Nits (logged): `MNT-6 · core/autoreconcile/names_test.go:542 · FuzzReconcileNames doc lists fewer invariants than checkNameList checks` · `MNT-7 · core/database/autoreconciler/autoreconciler_test.go:1669 · nameSpec comment could point to spec`

No Fix plan 2 (no round-2 blocker). Plan Landed note extended with the fuzz target and the alias-role rename (orchestrator, docs only). `origin/main` unchanged since the branch was cut (`f4a1c940`); no merge needed.

## Result
Gate PASS · CI pending · open: none · follow-ups: none · to user: — · CON miss: none
