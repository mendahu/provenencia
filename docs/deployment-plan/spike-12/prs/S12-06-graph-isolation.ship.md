# Ship log — S12-06 Graph isolation

**Plan:** [`S12-06-graph-isolation.md`](S12-06-graph-isolation.md) · **Branch:** `s12-06-graph-isolation` · **PR:** #515
**State:** ready

## Round 0 — implement
Head `0f6f480c`. Gate PASS (orchestrator run). Report: no deviations from Decisions; T-1 finds Census's Default graph with `evidencegraphs.List` instead of `evidencegraphstest.SoleGraph` (Source has two graphs); manual check not run (no two-graph path reachable in the app until slice 4).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`, `Localizable.xcstrings`, `core/apperr`. SEC: `api/ffi/handlers/connect.go`. PRF: `core/database/**`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Domain doc still allows cross-Source `remark` on a `source` Subject; contradicts the new rule beside it | Fix now → F-1 | verified: Letter citation's Artifact cannot be linked to the cert Source's graph (`Link` refuses other-Source Artifacts), so `requireSameGraph` returns `ErrOtherGraph`; already refused by `requireSubjectSource` before. User, 2026-10-21: make the docs match the code; source-to-source is a preliminary idea, not supported yet. Supersedes STEP-7 "the next paragraph stays" |
| MNT-2 | MNT | major | Third copy of the `sqlSubjectTypeKey` query/scan/ErrNoRows mapping; two callers discard `origin` | Fix now → F-2 | same-pass cleanup: this PR adds the third copy (D-3 named the shape, not a deferral) |
| CON-1 | CON | major→minor | Plan has no Landed note | Fix now → F-3 | DoD line 268 requires it; implementer's report lists a difference (T-1 finds Default with `List`, not `SoleGraph`). Docs-only, re-rated minor |
| MNT-3 | MNT | minor | "New graph, then unlink from Default" fixture written twice | Fix now → F-4 | same-pass cleanup: both copies added by this PR |
| MNT-out | MNT→CON | — | Plan Status says Landed before merge | Rejected | convention: S12-05 plan did the same (`S12-05-graph-rpcs.md:3`) |

UX and SEC+PRF: clean, no findings.

Nits (logged): `UX-1 · macos/App/Platform/L10n.swift:9694 · message gives the reason, not the next step; revisit when slice 4 makes a second graph reachable` · `MNT-4 · evidencegraphstest.go:74 · Create takes testing.TB, siblings take *testing.T` · `MNT-5 · interpretation-layer-data-model.md:939,960 · invariants 3 and 24 state the rule twice (plan STEP-7 asked for both)`

### Fix plan 1

**F-1 — Docs match the isolation rule for `source` Subjects** (from MNT-1)
Files: `docs/interpretation-layer-data-model.md` (the "For Subjects of type `source`" paragraph after the `source_id` paragraph, the §5.1 credibility/commentary text around "Cited commentary about a Source … continues to use Observations on the Source subject", the `remark` line and the Book/letter sentence after it, and the Letter L1 example with `SC1 -- remark --> …`), `docs/ideas/source-to-source-relationships.md` (lines "Do not enforce same-Source Observations in SQL" and "Cross-source Observations stay legal"), `docs/research-judgment-model.md:89`.
Failing test first: none (docs only).
Do: say what the code does. Another Source refers to a `source` Subject only as the value of a non-edge `mentions` Observation (D-3). Observations whose Subject is another Source's `source` Subject — including `remark` — are refused with `observations.other_graph`. A `remark` on a `source` Subject is written from that Source's own graph. Rewrite the Letter L1 example so it no longer writes `remark` on SC1 from the letter; keep a `mentions` from the letter's own `source` Subject to SC1 if the example needs the link. In the idea file, mark the cross-Source commentary lines as not supported under graph isolation (S12-06): a preliminary idea, not current behavior. In `research-judgment-model.md`, narrow "cited commentary about a Source … continues to use Observations on the Source Node" to the same rule. Do not change code.
Check: `rg -n "remark" docs/interpretation-layer-data-model.md` shows no cross-Source remark; scripts/check-pr.sh PASS.

**F-2 — One subject type-key read** (from MNT-2)
Files: `core/database/observations/observations.go`, `core/database/observations/write.go`.
Failing test first: none (refactor; existing T-1 and observation tests cover it).
Do: add `subjectTypeKeyTx(tx *sql.Tx, subjectID []byte) (key, origin string, err error)` in `observations.go`: runs `sqlSubjectTypeKey`, maps `sql.ErrNoRows` to `ErrInvalid`. Move `insertOne`, `isEdgeTx` and `requireSameGraph` onto it, deleting the discarded `typeOrigin` variables (callers use `_`). Keep `mapMissing` only if `requireSameGraph` still needs it for the `evidencegraphs` and citation reads.
Check: `rg -n "sqlSubjectTypeKey" core/database/observations` prints only the const and the one line in `subjectTypeKeyTx`; `rg -n "typeOrigin" core/database/observations` prints nothing; scripts/check-pr.sh PASS.

**F-3 — Landed note** (from CON-1)
Files: `docs/deployment-plan/spike-12/prs/S12-06-graph-isolation.md`.
Do: add a `## Landed note` (placed as earlier Spike 12 plans place it) listing every difference from the plan: T-1 finds Census's Default graph with `evidencegraphs.List` because `SoleGraph` needs a one-graph Source (or the F-4 helper, if it replaces that); F-1 edits the `source`-Subject paragraph STEP-7 said stays (user, 2026-10-21); F-2's `subjectTypeKeyTx` helper beyond D-3's shape; F-4's test helper; the manual check was not run because no two-graph path is reachable in the app until slice 4.
Check: scripts/check-pr.sh PASS (plan-docs).

**F-4 — Split-graph test helper** (from MNT-3)
Files: `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go`, `core/database/connect/connect_test.go`, `core/database/observations/observations_test.go`; `core/database/audit/scopes_test.go` if its inline `Unlink` (line ~336) is the same "move an Artifact out of Default" shape.
Do: add one helper beside `Create` (e.g. `Split(t, c, userID, sourceID, name, artifactIDs...) (def, graph evidencegraphs.Graph)`) that finds the Source's Default graph, creates the named graph with the Artifacts, and unlinks them from Default, failing the test on error. Move both new tests onto it, which also gives them one way to find Default. Move the audit site only if it is the same shape; otherwise leave it and say why in the report. The `Unlink` calls in `evidencegraphs_test.go` test `Unlink` itself and stay.
Check: `rg -n "evidencegraphs.Unlink\(" core -g '*_test.go'` prints only `evidencegraphs_test.go` lines (and the audit line if left, with the reason); scripts/check-pr.sh PASS.

## Round 1 — fix
Head `7154f5b3`. Gate PASS (orchestrator run). F-1 `f0a7c82f`, F-2 `13e4840c`, F-4 `061bea47`, F-3 `7154f5b3` done. Deviations: plan heading is `## Landed` (S11-22 precedent); `Split` takes `*testing.T`; `audit/scopes_test.go:336` left (it tests Link/Unlink audit scopes, not the fixture). CON-1 **Resolved** (orchestrator read the Landed section: lists T-1/Split, the STEP-7 doc edit, `subjectTypeKeyTx`, manual check not run). MNT-3 rides MNT round 2.

## Round 2 — review
Reviewers: MNT (MNT-1, MNT-2 majors). Not sent: CON (CON-1 re-rated minor, resolved above), UX and SEC+PRF (clean). Scope: `0f6f480c...7154f5b3`.
Head `7154f5b3`. MNT round 2: clean. MNT-1, MNT-2, MNT-3 **Resolved**. No new blockers.
Outside-lens note (MNT→CON, minor, logged): `interpretation-layer-data-model.md:581` and the idea file say another Source refers to a `source` Subject "only" through `mentions`; the code allows any non-edge Subject-valued Property (line 579 states the general rule). Wording only; round 2 fixes blockers only → Follow-up.

## Result
Gate PASS · CI pending · open: none · follow-ups: MNT-out-2 (`mentions`-only wording at `interpretation-layer-data-model.md:581`), UX-1 (next step in `observations.other_graph` copy when slice 4 makes a second graph reachable) · to user: MNT-1 (answered 2026-10-21: docs match code) · CON miss: none
