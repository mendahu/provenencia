# Ship log — S13-24 The pin engine goes

**Plan:** [`S13-24-the-pin-engine-goes.md`](S13-24-the-pin-engine-goes.md) · **Branch:** `s13-24-the-pin-engine-goes` · **PR:** #639
**State:** ready

## Round 0 — implement
Head `cc3bc843`. Gate PASS (orchestrator run against `705cd98c`). Report deviations:
- T-8's "negative, fixed" case expects `target ""`, not `handle`: the pair scores 1.5, under the weak band, and D-8 starts a band-less placed row unset. Plan T-8 and D-8 disagreed; D-8 followed.
- `TestMemoFollowsCommit/a Source retitled` removed: only the evidence query saw a retitle.
- Tests on deleted `Comparisons` assert `Row.Score` instead (`TestProposeJudgesAgainstTheMajority`, `TestProposeMultipleToponyms`, `TestProposeComparesTheLifespan`, `TestEventPeriodStandsForTheDate`); `TestGenderPoolsWithSex` keeps band and score.
- Renamed `TestContestedRow` → `TestAConflictCountsAgainstAPlacedPair`, `TestProposeMarksAConflictingRowContested` → `TestProposeFilesAConflictingRowOnTheLinks` (AC-4.1 `rg`).
- `MemoSnapshotForTest` lost `memberGraphID`; `BridgedOnly` export went with `bridgedOnly`; a names case uses an uncertain transcription in place of a low-confidence claim; `TestForEntityEvidence` gained a fourth person for T-12.
- Docs outside the list: `conclusion-reconciliation.md` Promote row lost its pin mention; `promote-graph-alignment.md` line 51 no longer says Contested and row reason are computed.
- Manual check not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (diff touches `macos/App/**`, `Localizable.xcstrings`, `api/ffi/**`, `core/database/**`, `core/claimevidence`, `core/graphalign`, `core/subjectcompare`, `core/autoreconcile`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `segmentcompare.ReadFrom` left with no caller | Fix now → F-1 | verified: `rg 'ReadFrom\('` at head finds only `key.go:105` and its test; base caller was `promotealign/evidence.go:227`; no doc names a later caller |
| MNT-2 | MNT | major | `cand.result`, `cand.links` (and `cand.edge`) set, never read | Fix now → F-2 | verified: `rg '\.result\b\|\.links\b\|\.edge\b' core/graphalign` finds only the assignment `align.go:257`; readers were D-9's deleted `contested`, `comparisonsFrom`, `alternative` |
| MNT-3 | MNT | minor | `init?(move:)` keeps `default` over a closed enum | Fix now → F-3 | verified `PromoteFlow.swift:62`; code this PR changes |
| MNT-5 | MNT | nit→minor | `ProvenanceColumns` / `ProvenanceJoins` comments name different aliases | Fix now → F-4 | contradictory comment in code this PR adds |
| UX-1 | UX | nit→minor | `CatalogReconcilerOutcome` comment still says "moderate confidence" | Fix now → F-4 | stale comment made false by this PR, `GenealogyStore.swift:587` |
| CON-1 | CON | minor | Placed band-less row: T-8 said `handle`, D-8 says unset | Rejected | D-8 binds and names this branch; a decided row keeps its earlier menu (`PromoteFlow.swift:517`), `load()` sends `fixed: []`, so no fresh row reaches it (UX traced the same). T-8's wording is a plan erratum, noted in the Landed note |
| SEC/PRF | SEC+PRF | — | No findings | — | — |

Nits (logged): `MNT-4 · api/ffi/handlers/promote_align.go:83,94 · r.Start() computed twice per row`

Outside-lens notes: UX → CON "manual check not run" — goes to the user in the report (plan §10).

### Fix plan 1
**F-1 — Delete `segmentcompare.ReadFrom`** (from MNT-1)
Files: `core/segmentcompare/key.go`, `core/segmentcompare/segmentcompare_test.go`. Skill: `add-match-rule`.
Failing test first: none (deletion).
Do: delete `ReadFrom` with its doc comment and `TestReadFrom`; drop any import or helper only they used. If `docs/matching-model.md` or a skill names `ReadFrom`, drop that line.
Check: `rg -n 'ReadFrom' core api docs .agents .cursor` finds nothing; `go test ./core/segmentcompare/...` passes; scripts/check-pr.sh PASS.

**F-2 — `cand` keeps only what is read** (from MNT-2)
Files: `core/graphalign/align.go`.
Failing test first: none (dead fields).
Do: remove `edge`, `result`, `links` (and the `links` comment) from `cand`; the assignment becomes `c.score = sc.Score`. Keep `ScoreCandidate` and `Scored` as they are (`align_test.go` reads `Scored.Result`). If `st.segments` is then needed only for `edge`, keep the `edge` sum feeding `ScoreCandidate`.
Check: `rg -n '\.result\b|\.links\b|c\.edge\b' core/graphalign` finds nothing outside tests; `go test ./core/graphalign/... ./core/database/promotealign/...` (with `-tags fts5`) passes with goldens unchanged; scripts/check-pr.sh PASS.

**F-3 — Move-reason switch is exhaustive** (from MNT-3)
Files: `macos/App/Features/Promote/PromoteFlow.swift`.
Do: replace `default: return nil` in `init?(move:)` with `case .unrecognized: return nil`.
Check: scripts/check-pr.sh PASS (xcode-test, swift-warnings).

**F-4 — Comments say what the code does** (from MNT-5, UX-1)
Files: `core/database/autoreconciler/autoreconciler.go`, `macos/App/Platform/GenealogyStore.swift`.
Do: write the alias contract once, on `ProvenanceColumns`: the query binds `c` (`citations`) and `a` (`artifacts`) and includes `ProvenanceJoins`, after which `sg` is in scope for a caller's own column; `ProvenanceJoins`' comment points at it. `CatalogReconcilerOutcome`'s comment ends "(standard credibility)."
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `a2c55088`. Gate PASS. F-1 (`7a48cd4e`), F-2 (`02f864f7`), F-3 (`f6ec8bd3`), F-4 (`a2c55088`) done. F-3 and F-4 read by the orchestrator: **Resolved** (`PromoteFlow.swift:62` `case .unrecognized`; alias contract once on `ProvenanceColumns`; `GenealogyStore.swift:587` ends "(standard credibility).").

## Round 2 — review
Reviewers: MNT (only lens with a Fix-now major). MNT-1 Resolved, MNT-2 Resolved. New findings: none. No Fix plan 2.

## Finish
Merged `origin/main` @ `cc870c65` (S13-30 descope: docs and one `PromoteView.swift` comment). Gate FAIL once in `xcode-test` (`SourceGraphFormModelTests`, load-state timing, code this PR does not touch; the implementer saw the same flake in round 0), rerun PASS with no change.

## Result
Gate PASS · CI green · open: none · follow-ups: none · to user: — · CON miss: none
