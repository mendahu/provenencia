# Ship log — S12-22 Artifacts can stand outside every graph

**Plan:** [`S12-22-artifacts-outside-graphs.md`](S12-22-artifacts-outside-graphs.md) · **Branch:** `s12-22-artifacts-outside-graphs` · **PR:** #600
**State:** ready

## Round 0 — implement
Head `6169d7b8`. Gate PASS (orchestrator, on the pushed head with the landing edits stashed). Report: deviations — `wireCase` regexp in `api/ffi/swift_vocabulary_test.go` accepts `.` (T-17 needs dotted codes); STEP-1/STEP-2 stop-gaps never written (core, proto and handlers changed together); three commits (engine, Swift, docs); `apperr.go` diff is gofmt realignment; FakeStore gains four scripted fields per D-14; manual check not run. **Stopped on:** the final "Record S12-22 as landed in pull request 600." commit was denied by the auto-mode permission classifier; the three landing edits (`completed.md`, map row, plan Status and Landed note) sit uncommitted in the working tree. Taken to the user; review proceeds meanwhile since it is read-only. User said yes (2026-10-27); orchestrator committed it as `583d4208`, with UX-2's S12-38 Check line in the same map edit.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none — UX for `macos/App/**`, `Localizable.xcstrings`, `core/apperr`; SEC for `api/ffi/**`; PRF for `core/database/**`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | `checkUnlink` hand-writes a second copy of the Citation row scan | Fix now → F-1 | verified: `evidencegraphs.go` `CitationsUsingArtifactTx` and `checkUnlink` rule (3) scan `deleteimpact.Listed` line for line; same-pass cleanup |
| PRF-1 | PRF | minor | `Unlink` lists every Citation row only to test whether any exist | Fix now → F-2 | verified: `Unlink` discards `checkUnlink`'s rows; minor in a file this PR changes |
| MNT-2 | MNT | major | Via constants are second literals of their codes; `UnlinkImpact` keeps a reason→via switch | Fix now → F-3 | verified: `register.go:15` and `UnlinkVia` repeat `apperr` codes; switch in `UnlinkImpact` drops a block silently for an unlisted reason |
| UX-1 | UX | minor | `citations_need_graph` copy says "this artifact" in the graph form too | To user → Fix now → F-7 | verified: `SourceGraphFormModel.save` shows `L10n.Errors.message`; user (2026-10-27) chose "An artifact with citations must stay in at least one evidence graph." for AC-4.1 |
| MNT-3 | MNT | minor | `ErrInUse` and `Delete` comments describe only the Subjects rule | Fix now → F-4 | verified: `evidencegraphs.go:28`, `:232` untouched by the diff |
| MNT-4 | MNT | minor | `Create` calls `requireDistinctPairs` before insert, against D-5 | Fix now → F-5 | verified: calls at `Create`, `applyLinks`, `UnlinkImpact` |
| MNT-5 | MNT | minor | Test helper `graphsOf` copies `linkedGraphs`; `bareCitation` leaves an inline copy | Fix now → F-6 | same-pass cleanup; both copies in test files this PR changes |
| UX-2 | UX | minor | Card's partial-save copy says "save again to finish" for a refusal saving cannot clear | Later PR (S12-38) | written into S12-38's map scope **Check** (`deployment-plan.md`): the box ends checked and locked, no "save again" text |

Nits (logged): `CON-1 · core/database/artifacts/artifacts.go:150 · reflow the Update doc comment` · `UX-3 · SourceArtifactsSection.swift:234 · card now shows loading for one round trip; look during manual check` · `MNT-6 · evidencegraphs.go:723 · Unlink and checkUnlink both run IsLinkedTx` · `MNT-7 · evidencegraphs.go:763 · redundant ids = nil` · `MNT-8 · catalog-deletes.md:18, source-layer-data-model.md:425 · "a graph that holds alone the only graph of an Artifact" does not parse`

### Fix plan 1

**F-1 — One Citation row scan in `evidencegraphs`** (from MNT-1)
Files: `core/database/evidencegraphs/evidencegraphs.go`.
Failing test first: none (refactor); existing `TestUnlink`, `TestUnlinkImpact`, `TestCitationsUsingArtifact*` stay green.
Do: add unexported `queryCitationRows(tx *sql.Tx, q string, args ...any) ([]deleteimpact.Listed, error)` holding the Query/Scan/Err loop. `CitationsUsingArtifactTx` keeps its id check and calls it with `sqlCitationsUsing`; `checkUnlink` rule (3) calls it with `sqlCitationsNeedGraph`.
Check: `grep -n 'rows.Scan(&row.ID, &row.Ref)' core/database/evidencegraphs` prints exactly one line (inside `queryCitationRows`); scripts/check-pr.sh PASS.

**F-2 — `Unlink` probes, `UnlinkImpact` lists** (from PRF-1)
Files: `core/database/evidencegraphs/evidencegraphs.go`.
Failing test first: none (no behavior change); `TestUnlinkImpact`'s Unlink/UnlinkImpact agreement case stays green.
Do: give `checkUnlink` a `list bool` parameter. When false (from `Unlink`), rules (1) and (3) test existence with one helper beside `queryCitationRows`, `citationRowsExist(tx, q, args...)`, running `SELECT EXISTS(` + q + `)`, and return a nil row slice. When true (from `UnlinkImpact`), they list through `queryCitationRows`. The reasons and their order do not change; this stays the one unlink check (D-6).
Check: both callers pass the flag; scripts/check-pr.sh PASS.

**F-3 — A via is its code** (from MNT-2)
Files: `core/database/deleteimpact/register.go`, `core/database/evidencegraphs/evidencegraphs.go`, tests that compare the constants.
Failing test first: none (refactor).
Do: `ViaCitationsNeedGraph = apperr.CodeEvidenceGraphsCitationsNeedGraph`; `UnlinkVia = apperr.CodeEvidenceGraphsArtifactInUse`. In `UnlinkImpact` replace the reason switch with `if len(rows) > 0 { block.Citations, err = deleteimpact.ListedGroup(tx, apperr.From(block.Reason).Code(), catalogmodel.KindCitation, rows) }`.
Check: `grep -n '"evidencegraphs\.' core/database/deleteimpact/register.go core/database/evidencegraphs/evidencegraphs.go` prints no via literal; `TestUnlinkImpact` green; scripts/check-pr.sh PASS.

**F-4 — Comments match `Delete`'s rules** (from MNT-3; nits CON-1, MNT-8 ride along)
Files: `core/database/evidencegraphs/evidencegraphs.go:28`, `:232`; `core/database/artifacts/artifacts.go:150`; `docs/catalog-deletes.md:18`; `docs/source-layer-data-model.md:425`.
Do: `ErrInUse` and `Delete` comments add "or while it is the only graph of an Artifact with Citations". Reflow the `artifacts.Update` doc comment. Docs: "a graph that is the only graph of an Artifact with Citations".
Check: scripts/check-pr.sh PASS.

**F-5 — `Create` leaves the pair check to `applyLinks`** (from MNT-4)
Files: `core/database/evidencegraphs/evidencegraphs.go` `Create`.
Failing test first: none; the existing repeated-Artifact `Create` case still returns `ErrInvalid` and writes nothing.
Do: drop the pre-insert `requireDistinctPairs`; build `want` after `insert` with `g.ID`.
Check: `grep -n requireDistinctPairs core/database/evidencegraphs/evidencegraphs.go` prints the definition and two calls (`applyLinks`, `UnlinkImpact`); scripts/check-pr.sh PASS.

**F-6 — One test helper for an Artifact's graphs** (from MNT-5)
Files: `core/database/evidencegraphs/evidencegraphstest/`, `core/database/evidencegraphs/evidencegraphs_test.go`, `core/database/artifacts/artifacts_test.go`.
Do: add `evidencegraphstest.ArtifactGraphs(t, c, artifactID) [][]byte` ordered by `graph_id`; replace `graphsOf` and `linkedGraphs` with it; use `sameIDs` and delete `sameSet`; `TestMembership` calls `f.bareCitation`.
Check: `grep -rn 'SELECT graph_id FROM evidence_graph_artifacts WHERE artifact_id' core --include=*_test.go core/database/evidencegraphs/evidencegraphstest` prints only the helper; scripts/check-pr.sh PASS.

**F-7 — The refusal reads right in the graph form too** (from UX-1; user answer 2026-10-27)
Files: `macos/App/Platform/L10n.swift` (`evidenceGraphsCitationsNeedGraph`), `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/L10nResolutionTests.swift` (T-15), the plan's Landed note. Skill: `add-localized-string`.
Failing test first: T-15's `citations_need_graph` row expects "An artifact with citations must stay in at least one evidence graph." — red against the current copy.
Do: change the defaultValue and the catalog's English value to that text; keep the key and comment (update the comment if it names the card). In the plan's Landed note, add one line: AC-4.1's copy changed to this text by the user's answer to review UX-1 (2026-10-27). Do not edit the AC table.
Check: T-15 passes; `python3 scripts/check-localizable-xcstrings.py` OK; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `af36ad2e` (commits `754d3d6c`…`af36ad2e` on top of the landing commit `583d4208`). Gate PASS (orchestrator). F-1 to F-7 done. Deviations: F-2 lists rule (1) through `CitationsUsingArtifactTx` (now on `queryCitationRows`) and probes only on the `Unlink` path; F-6 sorts `tt.want` before `sameIDs`; the plan's §6 Strings line and the review log keep the old copy, the Landed note records the change.
Minors read by the orchestrator, not re-reviewed: **PRF-1 Resolved** — `Unlink` calls `checkUnlink(…, false)`, rules (1) and (3) run `SELECT EXISTS(…)` through `citationRowsExist`; `UnlinkImpact` still lists. **UX-1 Resolved** — `L10n.swift` defaultValue and the xcstrings English value read "An artifact with citations must stay in at least one evidence graph."; T-15 went red first. MNT-3 to MNT-5 go back with MNT's round 2.

## Round 2 — review
Reviewers: MNT (round 1 had MNT-1 blocker, MNT-2 major). Not sent: CON (clean), UX (minor only), SEC+PRF (minor only). Scope `583d4208...af36ad2e`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | (round 1) one Citation row scan | Resolved | one `rows.Scan(&row.ID, &row.Ref)` left, in `queryCitationRows` |
| MNT-2 | MNT | major | (round 1) a via is its code | Resolved | both vias are `apperr.Code…`; reason switch gone |
| MNT-3 – MNT-5 | MNT | minor | (round 1) comments, `Create` pair check, test helper | Resolved | greps in Fix plan 1 print what they should |

No new findings. Observation (not a finding): `UnlinkVia` now has only test callers; it names the wire via those tests pin. No Fix plan 2.

## Finish
Merged `origin/main` (`e9fd92d2`, S11-46 plan docs only). Gate PASS against `e9fd92d2`.

## Result
Gate PASS · CI pending · open: none · follow-ups: none (UX-2 written into S12-38's Check) · to user: UX-1 (answered 2026-10-27: new copy) and the denied landing commit (answered: orchestrator committed `583d4208`) · CON miss: none
