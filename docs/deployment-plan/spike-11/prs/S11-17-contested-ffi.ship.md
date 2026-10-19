# Ship log — S11-17 Contested flag through FFI

**Plan:** [`S11-17-contested-ffi.md`](S11-17-contested-ffi.md) · **Branch:** `s11-17-contested-ffi-impl` (plan header names `s11-17-contested-ffi`; that name is the merged plan PR #502's stale branch, so the user chose a new name) · **PR:** #504
**State:** ready

## Round 0 — implement
Head `07d84a8b`. Gate PASS (orchestrator ran `scripts/check-pr.sh` against `3db19994`). Report deviations: branch renamed per the user; commits 2 and 3 split as planned by temporarily holding back the D-7 hunks; commit 4 subject per the implement-pr skill; manual check (§10) not run (no visible change).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (`macos/App/**` routes UX; `api/ffi/**` routes SEC; `macos/App/Features/Promote/PromoteModel.swift` routes PRF).

SEC+PRF: no findings. UX: no findings. CON: no findings. MNT: four minors.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | minor | `macos-client.mdc` wire-vocabulary list omits Promote reason codes | Fix now → F-3 | verified: `.cursor/rules/macos-client.mdc:20` lists five categories; this PR adds a sixth to `ConclusionVocabulary.swift` |
| MNT-2 | MNT | minor | `TestProposeSendsContested` copies `TestProposeNegativeConflictLine`'s setup | Fix now → F-1 | same-pass cleanup: this PR adds the copy; verified `promote_align_test.go:79-89` vs the "negative, fixed" reqFn |
| MNT-3 | MNT | minor | `TestProposalProtoSendsMove` skips `catalogsession.CloseAll` cleanup | Fix now → F-2 | verified: calls `withProjectCatalog` directly; `runRPC` registers the cleanup at `harness_test.go:38`, this test does not |
| MNT-4 | MNT | minor | Vocabulary guard's `want` is hand-written, so a new Go reason passes unnoticed | Follow-up | pattern shared by every guard row; a source-level check covers all rows at once, not this PR's job |

Nits (logged): `UX-1 · PromoteModel.swift:225 · say the interim move arm must not be used for row.move` · `CON-1 · S11-17-contested-ffi.md:3 · Landed note paragraph sits above the header lines` · `MNT-5 · GenealogyStore.swift:327 · moveNeighborSubjectID lacks a doc comment` · `MNT-6 · promote_align_test.go:121 · nested fixture struct for two fields` (rides along in F-1, which rewrites that table)

### Fix plan 1

**F-1 — One fixture for the contested and conflict-line cases** (from MNT-2, MNT-6)
Files: `api/ffi/handlers/promote_align_test.go`.
Failing test first: none (test-only refactor); the folded assertion must pass.
Do: Fold `TestProposeNegativeConflictLine`'s assertion into `TestProposeSendsContested`'s "negative, fixed" row: in `after`, when the case is the fixed negative one, also require a `sex_at_birth` comparison with outcome `conflict` that is `IncomingNegative`, not `Pinned`, has empty `IncomingObservationId`, and `IncomingDisplay == "Male"` (same checks, same search over all rows as the old test). Then delete `TestProposeNegativeConflictLine` and update the contested test's doc comment to say it also covers the conflict line. Flatten the nested `fixture` struct: put `polarity` and `fix` directly on the case struct.
Check: `grep -n "TestProposeNegativeConflictLine" -r api` prints nothing; `grep -c "personOnNewSource(t, dir, userID, typeID, \"male\", observations.PolarityPositive)" api/ffi/handlers/promote_align_test.go` counts only the contested reqFn and the "reason properties" case; `go test -tags fts5 ./api/ffi/handlers/ -run TestPropose` passes; scripts/check-pr.sh PASS.

**F-2 — Close the catalog session in the direct-call test** (from MNT-3)
Files: `api/ffi/handlers/promote_align_test.go`.
Do: Make `t.Cleanup(func() { _ = catalogsession.CloseAll() })` the first line of `TestProposalProtoSendsMove`, as `conclusion_detail_test.go:43` does.
Check: `rg -n 'CloseAll' api/ffi/handlers/promote_align_test.go` prints one line; tests pass.

**F-3 — Name Promote reason codes in the macOS rule** (from MNT-1)
Files: `.cursor/rules/macos-client.mdc`.
Do: Add "Promote reason codes" to the parenthesized list on line 20, after "reconciler reasons".
Check: `rg -n 'Promote reason codes' .cursor/rules/macos-client.mdc` prints line 20; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `962b9037`. Gate PASS (orchestrator ran it). F-1 done: `TestProposeNegativeConflictLine` folded into `TestProposeSendsContested` through `requireNegativeConflictLine`; nested fixture flattened. F-2 done. F-3 done. The F-1 grep counts 3, not 2, because the "reason properties" case already calls `personOnNewSource` twice; the conflict-line copy is gone. All three read by the orchestrator: **Resolved**.

## Round 2 — review
Not run: round 1 had no Fix-now blocker or major.

## Result
Gate PASS · CI pending · open: none · follow-ups: MNT-4 · to user: — · CON miss: none
