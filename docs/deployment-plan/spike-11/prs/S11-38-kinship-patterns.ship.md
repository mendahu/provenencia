# Ship log — S11-38 Kinship patterns

**Plan:** [`S11-38-kinship-patterns.md`](S11-38-kinship-patterns.md) · **Branch:** `s11-38-kinship-patterns` · **PR:** #564
**State:** fixing 1 (SEC-1 to user)

Branch cut from `origin/main` @ `e3b85f97` with the two plan commits from `s11-38-plan` cherry-picked on top (the plan never landed on `main`).

## Round 0 — implement
Head `0581b563` · PR #564 (draft). Gate PASS (diff-check, gofmt, go-vet, go-test, localizable, plan-docs, xcode-test, swift-warnings). Report: deviations none; every Current state row held against `846fa65`.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes). SEC+PRF ran because `core/kinship` is a new matching layer package and the diff touches `core/matchregistry` and graph-shaped code.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | `chains` compares every pair of parent steps and allocates a 3-string key per pair | Fix now → F-1 | verified: `kinship.go:177-179` builds `triple{…}` before the `bytes.Equal(up.m, down.g)` filter, and lines 186/191 rescan all steps per triple; reviewer measured 4× per doubling (S≈8k → 4.2 s, 3.1 GB). `loadCanon` feeds a whole family in S11-39 |
| MNT-1 | MNT | major | Two copies of "match an edge to a key pair from either end" | Fix now → F-2 | verified: `kinship.go:64-69` and `:126-131` are the same D-3 rule as two switches |
| SEC-1 | SEC (+PRF) | major | Event patterns emit k(k-1)/2 links per event from a catalog-sized participant count | To user | verified: `kinship.go:146-166` nested loops over one event's group. Output size is quadratic by D-5's definition ("every two different people who are each `spouse`"), so any bound changes D-5 and needs a number the "No new knob" rule and the `numbers` guard refuse in this package |
| MNT-2 | MNT | minor | Skill marks `core/kinship/patterns.go` "(new)" | Fix now → F-3 | verified: `.agents/skills/add-match-rule/SKILL.md:87`; file this PR changes |
| MNT-3 | MNT | minor | Second copy of the `LookupBridge`+`Fatalf` test helper | Fix now → F-4 | same-pass cleanup: `kinship_test.go:20-27` copies `segmentcompare_test.go:293-300` (`bridgeOf`) |
| CON-1 | CON | minor | Landed note omits the `docs/ideas/kinship-view.md` line | Rejected | that line came from the plan's own review commit (`90bd3126`, review M-3), not from implementation; it is no difference from the plan |

Nits (logged): `CON-2 · core/kinship/kinship_test.go:196 · no non-mutation check in TestDeriveInvariants (covered per row)` · `MNT-4 · core/kinship/kinship.go:58,73 · drop out, return append(implied, chains(...)...)` · `MNT-5 · core/kinship/kinship.go:113 · Concat makes duplicate spouse roleKeys (harmless: the loop breaks on first match)` · `MNT-6 · core/kinship/kinship.go:99 · participant at package level, roleKeys local`

### Fix plan 1

**F-1 — Index parent steps in `chains`** (from PRF-1)
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`.
Failing test first: none (behavior-preserving); `TestDerive` and `TestDeriveInvariants` must stay green unchanged.
Do: in one pass over `steps`, build `byG map[string][]int` (step indices by `g`) and `byPair map[[2]string][]int` (indices by `(g, m)`), indices in step order. For each `up` in `steps`, iterate only `byG[string(up.m)]`; skip `G == C`; build the `made` key only after the match. Take `from` from `byPair[{up.g, up.m}]` then `byPair[{up.m, down.m}]` instead of rescanning `steps`. Output order and contents must stay exactly as D-6/D-11 define. No numeric literals (the `numbers` guard), so no `make` capacity hints. Add `BenchmarkDerive` in `kinship_test.go` over a synthetic parent tree (each child two parents) so S11-39 has a baseline; size it with existing test values, not a new package constant.
Check: `go test ./core/kinship/... ./core/matchregistry/...`; `go test -run x -bench Derive ./core/kinship/` shows roughly linear growth when the tree doubles (report two sizes); scripts/check-pr.sh PASS.

**F-2 — One helper for "read an edge by a key pair from either end"** (from MNT-1)
Files: `core/kinship/kinship.go`.
Do: add an unexported `func (e Edge) reads(first, second vocabkey.BridgeKey) (x, y []byte, ok bool)` returning the ends in key-pair order (D-3). Use it in the parent-step loop in `Derive` and the participant loop in `eventLinks`. Leave the implied-link loop as is unless the helper reads it equally plainly. Afterwards `rg -n 'KeyA == .*KeyB ==' core/kinship` prints only the helper's own lines.
Check: `TestDerive`, `TestDeriveInvariants` pass; scripts/check-pr.sh PASS.

**F-3 — Drop "(new)" from the skill** (from MNT-2)
Files: `.agents/skills/add-match-rule/SKILL.md:87`.
Do: remove " (new)" after `core/kinship/patterns.go`.
Check: `rg -n '\(new\)' .agents/skills/*/SKILL.md` prints only `plan-pr` lines.

**F-4 — Share the bridge-lookup test helper** (from MNT-3)
Files: `core/segmentcompare/segmentcomparetest/terms.go` (or a sibling file in that package), `core/kinship/kinship_test.go`, `core/segmentcompare/segmentcompare_test.go`.
Do: add `func Bridge(t testing.TB, key string) connectrules.Bridge` beside `Terms` (`t.Helper()`, `connectrules.LookupBridge`, `t.Fatalf` on miss). Delete `bridge` in `kinship_test.go` and `bridgeOf` in `segmentcompare_test.go`; point their callers at `segmentcomparetest.Bridge`.
Check: `rg -n 'LookupBridge\(' core -g '*_test.go'` prints nothing; `rg -n 'func Bridge' core/segmentcompare/segmentcomparetest` prints one line; scripts/check-pr.sh PASS.
