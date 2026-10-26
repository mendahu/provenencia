# Ship log — S11-39 Matching reads kinship shapes

**Plan:** [`S11-39-matching-reads-kinship.md`](S11-39-matching-reads-kinship.md) · **Branch:** `s11-39-matching-reads-kinship` · **PR:** #567
**State:** ready

## Round 0 — implement
Head `f631a272`. Gate PASS (orchestrator run). Report deviations: STEP-7 baseline taken from a `git archive` export of `7efa02f` rather than a worktree (export deleted); extra `cloneShapes` helper in `promotealign/terms.go`; T-6 goldens use `kinTree`/`kinSide`/`kinShape` helpers; manual check not run in the app (Go goldens cover its steps). The landing commit also picked up this ship log stub.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF (`core/database/promotealign`, graph code, loads stored links and trusts their count). Skipped: UX (no `macos/App/**`, strings, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC/PRF | major | `kinship.Derive` pairs every participant of a pattern event (and parents × children through a middle person), unbounded; now reached from both loaders on every proposal | To user → Fix now → F-1 | verified: `eventLinks` nested `for x in group { for y in group }` (`kinship.go:189-191`); reviewer's scratch run, 2,000 spouses on one marriage → 1,999,000 links in 4.5 s. Defect in callers this PR adds, so Fix now in shape, but the fix puts a bound in a package documented "no numbers" under the working rule "No new knob" |
| PRF-1 | PRF | major | Canon re-derives implied links on every Promote decision, no memo | Rejected (memo) · Fix now → F-3 (budget case) | D-8 and the Out line decline the memo on `BenchmarkDerive` (16 ms at 8,190 links); the reviewer confirms `TestLargeSourceBudgets` has no chains or marriages in the canon, so D-8's "budgets moved within noise" was not measured. A budget case that exercises the path tests D-8's premise; the Out line's "revisit if a budget fails" then applies |
| MNT-1 | MNT | minor | `sideLink` repeats `graphalign.Bridge` field by field | Rejected | D-5 names `sideLink` embedding `kinship.Edge` so `withImplied` reads `l.Edge`; holding `Bridge` moves the conversion into `withImplied` instead of removing it |
| MNT-2 | MNT | minor | Loader doc comments omit the merge and the implied links | Fix now → F-2 | read `layer.go:155-157`, `canon.go:32-36`: unchanged by the diff |
| SEC-2 | SEC | minor | `linkSet.add` rescans associations on each merge, O(M²) for M marriages of one couple | Fix now → F-4 | `terms.go` is added by this PR; `slices.ContainsFunc` over `have.Associations` per added association |

Nits (logged): `CON-1 · plan:17 · Review line says 2 rounds` (misread: those are the plan review's rounds) · `CON-2 · plan:145 · Codebase fit names TestWithImplied, which does not exist` · `MNT-3 · core/graphalign/align.go:326 · pass st.placedOf instead of a closure` · `MNT-4 · promotealign/terms.go · linkSet would read better in links.go`

**SEC-1, user's answer (2026-10-26):** "the cheap structural guard is fine, but I want it to be more than 2, just to account for strange familial situations. If we cap it at something like 10". Bound: 10 per end.

### Fix plan 1

**F-1 — Bound what one pivot can imply in `core/kinship`** (from SEC-1; user-approved)
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`. Skill: `add-match-rule` (read it; this is a structural guard, not a registry knob, so no `matchregistry` entry).
Failing test first: add `TestDerive` rows: (a) a marriage with 11 spouse participations implies nothing for that marriage; (b) a marriage with 10 spouses implies 45 spouse links; (c) a birth with 11 parent participations (father/mother roles) implies no parent links for that birth; (d) a middle person with 11 parents implies no grandparent link through them, while one with 10 does. Also keep every existing row passing.
Do: add one unexported constant beside the patterns, `maxPivotEnd = 10`, with a comment: a pivot (an event, or the parent between in a chain) with more than 10 participants at either end reads as malformed data and implies nothing; it bounds the links one pivot makes so `Derive` stays linear in its input. In `eventLinks`, per event group, count participants with `atA` and with `atB`; if either count exceeds `maxPivotEnd`, skip the group. In `chains`, skip a middle person whose distinct parents exceed `maxPivotEnd` (count distinct `g` over steps whose child is that person). Update the package doc comment, which says "no numbers": say it holds no weights or scores, only this bound on malformed data.
Check: the new rows pass; `go test ./core/kinship/ -bench Derive` still runs; scripts/check-pr.sh PASS.

**F-2 — Say what the loaders return** (from MNT-2)
Files: `core/database/promotealign/layer.go` (`loadLayerBridges` comment), `core/database/promotealign/canon.go` (`loadCanon` comment).
Do: add one sentence to each: links go through one `linkSet`, so links with the same ends and keys merge, and the links the kinship patterns imply follow the recorded ones (with no association of their own beyond those they were made from).
Check: scripts/check-pr.sh PASS.

**F-3 — Make the large-source budgets exercise implied links** (from PRF-1; the memo stays out per D-8)
Files: `core/database/promotealign` large-source budget test (`TestLargeSourceBudgets` and its fixture).
Do: add one budget case (or extend the fixture) whose canon holds a multi-generation parent tree with marriages, at a size comparable to `BenchmarkDerive/generations=12` (about 8,000 parent links), reached from the proposal's seeds, so a warm proposal runs `withImplied` over it. Reuse the existing budget limits and fixture builders; do not add a new budget number unless no existing one fits, and if so name it in the report. If the case fails its budget, stop and report the numbers (the plan's Out line says to revisit the memo then); do not add a memo.
Check: the case passes; report its time against the budget; scripts/check-pr.sh PASS.

**F-4 — Merge associations without rescanning** (from SEC-2)
Files: `core/database/promotealign/terms.go`.
Failing test first: none needed (behavior unchanged); `TestLinkSet` must stay green, including association order.
Do: in `linkSet`, keep per link a set of its associations (for example `seen []map[string]struct{}` parallel to `links`, or a `map[seenEdge]map[string]struct{}`), filled on a miss and checked on a hit instead of `slices.ContainsFunc`. Keep association order as first added.
Check: `TestLinkSet`, `TestLayerBridgesMergeKinshipShapes` pass; scripts/check-pr.sh PASS.

Also record in the plan's Landed note: F-1 adds the bound of 10 per pivot end (user decision on SEC-1, in this ship log); F-3 adds the budget case.

## Round 1 — fix
Head `823856b0`. Implementer's gate PASS without F-1. F-2 done (`350f777f`), F-4 done (`5a29a0a6`), F-3 done (`823856b0`): the shared large-source fixture now holds a 12-generation tree (8,190 parent links, 4,095 marriages); warm 123 ms vs 73 ms without it (budget 250 ms), cold 1.11 s (2.5 s), after a commit 478 ms (1.5 s). The memo stays out per D-8. Deviation: the tree lookups run before `bulkCanon`'s transaction (one connection; a query inside it deadlocked).

**Stopped on F-1.** `core/matchregistry/guard_test.go` marks `core/kinship` `numbers: true`: "the registry owns every tunable". `const maxPivotEnd = 10` fails `TestLayerPackages/core/kinship/numbers`. Fix plan 1 was wrong to call the bound "not a registry knob". The patch (code + `TestDeriveBound`, 10 vs 11 at each pivot) is saved, unapplied. **To user:** where the bound lives.

**F-1, user's answer (2026-10-26):** "Put the cap in core/matchregistry as a registry value (10), read in promotealign and passed into kinship.Derive."

### Fix plan 1, F-1 revised

**F-1′ — Bound what one pivot can imply, through a registry value** (from SEC-1; user-approved twice)
Files: `core/matchregistry` (the registry value and its doc), `core/kinship/kinship.go`, `core/kinship/kinship_test.go`, `core/database/promotealign/terms.go` (`withImplied`) and its callers in `layer.go`/`canon.go`, `docs/matching-model.md` knob inventory. Skill: `add-match-rule` (follow its steps for a registry knob, including the knob inventory row).
Start from the saved patch `/private/tmp/claude-501/-Users-jakerobins-Repos-provenencia/e7ccccd4-c472-4924-8362-33987dd34c10/scratchpad/f1.patch` (logic and `TestDeriveBound`), then:
Do: add the value to the registry where values of its kind live (as `PrefixMinLength` does), named for what it bounds (for example `Kinship.MaxPivotEnd`), value 10, with a comment: a pivot (an event, or the parent between in a chain) with more participants than this at either end reads as malformed data and implies nothing; it keeps derivation linear in its input. `kinship.Derive` takes the bound as an argument (no numeric literal in `core/kinship`; it must not import `matchregistry`). `promotealign` reads it from `matchregistry.Default()` and passes it through `withImplied`. `TestDeriveBound` passes the bound in and pins 10 vs 11 at each end; `BenchmarkDerive` passes the registry's value or a test-local one. Update `core/kinship`'s package doc ("no numbers" stays true: the bound comes in as an argument).
Check: `TestDeriveBound` rows pass (fail without the skip); `TestLayerPackages` passes unchanged; scripts/check-pr.sh PASS. Add an F-1 line to the plan's Landed note.
F-1′ head `16ff9ef2`. Gate PASS (orchestrator run). `kinship.Rules{MaxPivotEnd}` from `matchregistry.Default().Kinship` (10), passed through `withImplied`; knob inventory row added; `TestLayerPackages` unchanged.

## Round 2 — review
Reviewers: SEC+PRF (round-1 majors SEC-1, PRF-1). Not sent back: CON (clean), MNT (minors only). MNT-2 read by the orchestrator in `layer.go`/`canon.go`: **Resolved**. SEC-2 (F-4): resolved, `linkSet` keeps a per-link association set.

SEC-1: Resolved for events and distinct parents. PRF-1: Resolved (budget case: warm 140 ms / 250 ms, cold 1.31 s / 2.5 s, after a commit 609 ms / 1.5 s).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-3 | SEC/PRF | blocker | `chains` dedupes grandparent associations with `slices.ContainsFunc` over a growing slice: quadratic in a parent link's associations, which `MaxPivotEnd` does not bound; `Rules`'s comment says Derive is linear | Fix now → F-5 | verified: `kinship.go:280-286`; reviewer's probe, 16,000 associations on one g→m link × 10 grandchildren → 4.1 s |

### Fix plan 2

**F-5 — Dedupe chain associations through a set** (from SEC-3)
Files: `core/kinship/kinship.go` (`chains`), `core/kinship/kinship_test.go`.
Failing test first: add a `BenchmarkDerive` case (or a `TestDeriveBound` row with a time-independent check) for one parent link g→m carrying several thousand associations plus m→c1..c10; and a `TestDerive` row asserting the implied grandparent link's associations are deduped and in first-added order (duplicates across the two steps appear once).
Do: in `chains`, replace the `slices.ContainsFunc(assoc, …)` scan with a `map[string]struct{}` keyed by `string(a)`, keeping first-added order, the same shape F-4 used in `linkSet.add` (`core/database/promotealign/terms.go`). Optionally collect `from` and the deduped associations once per `(g,m)` pair and reuse it across that pair's grandchildren. Keep the `Rules` comment's "linear in its input" true.
Check: the new rows pass; `go test ./core/kinship/ -bench Derive` shows the many-association case roughly linear (report the timings at two sizes); scripts/check-pr.sh PASS.

## Round 2 — fix
Head `9c29cbaf`. Gate PASS (orchestrator run; the implementer's first run hit an unrelated Swift debounce flake, rerun passed). F-5 done: `chains` dedupes through a set; `TestDeriveManyAssociations` (20,000 associations, each twice, × 10 grandchildren) 11.8 s before → 0.03 s after. Merged `origin/main` (S12-14 plan docs only) as `d2194f2e`; gate PASS.

## Result
Gate PASS · CI pending · open: none · follow-ups: none · to user: SEC-1 (answered: bound of 10, then as a `matchregistry` value) · CON miss: none
