# Ship log — S10-21 Handle last-activity scope

**Plan:** [`S10-21-handle-last-activity.md`](S10-21-handle-last-activity.md) · **Branch:** `s10-21-handle-last-activity` · **PR:** #398
**State:** ready

Base: `main` @ `7d77b657` (plan merged in #397; the plan's planning base `a8d7971a` differs only by the plan docs).

## Round 0 — implement
Head `b3b1f104`. Gate PASS (orchestrator run). Report: no deviations from Decisions. The implementer's first gate hit a one-off failure in `WorkspaceSessionTests.invalidateAllMatching()` (no Swift source changed); it passed on rerun and in the orchestrator's run. AC-4.3 manual check not run (user). The implementer already set the plan Status to Landed #398 and the map "done, #398" ahead of merge.

## Round 1 — review
Reviewers on `7d77b657...b3b1f104`. UX, Security, Performance: clean, no findings.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `audit.Scope` is dead and its comment is false | Fix now → F-1 | verified: `git grep -nw Scope` finds only the type and its comment in `scopes.go` |
| MNT-2 | MNT | major | `.cursor/rules/catalog-audit.mdc` still points at resolvers in `scopes.go` and omits the handle scope | Fix now → F-2 | verified: lines 12, 15, 16 as quoted; small docs addition in the STEP-8 concern |
| CON-OUT | MNT (routed from CON) | minor | Audit doc §4.1 sentence "through an effect path, which walks each change through its effect path" is garbled | Fix now → F-2 | verified: `docs/audit-revision-history.md:330`; F-2 already edits the same concern |
| CON-1 | CON | major | Plan has no Landed note listing test deviations; nil `q` is untested | Fix now → F-3 | verified: only line 3 changed in the plan; deviations at `scopes_test.go` `TestLatestByScope` and `conclusionheaders_test.go` `TestListHeadersCarryUpdated` |
| CON-2 | CON | minor | No test shows an evidence pin moves its handle | Fix now → F-4 | D-2 puts pins in scope and nothing tests it; a few lines in an existing fixture |
| MNT-3 | MNT | minor | `TestSourceScopes` inlines the writes the new seed helpers wrap | Fix now → F-5 | F-3 already touches `scopes_test.go`; mechanical |
| MNT-4 | MNT | minor | Three identical fill blocks in `List*` | Rejected | D-6 chose this shape; reviewer marks leaving it acceptable; `compose.go` is not in the fix plan |

### Fix plan 1

**F-1 — Delete the dead `audit.Scope` type** (from MNT-1)
Files: `core/database/audit/scopes.go`.
Do: delete `type Scope` and its comment. Nothing else references it.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/audit/...` passes; scripts/check-pr.sh PASS.

**F-2 — Correct the catalog-audit rule and the §4.1 sentence** (from MNT-2, CON-OUT)
Files: `.cursor/rules/catalog-audit.mdc`, `docs/audit-revision-history.md`.
Do: in the rule, line 12 says a new `EntityType` (or `rowFacet` entity type) needs a non-`None` entry in `core/database/effects/registry.go` (its `Source` / `HandleScope` paths), and `TestResolversCoverEveryEntityType` fails CI without one. Line 15 names both readers: the Sources list "Updated" (`scope_type = 'source'`) and Person / Event / Place last activity (`canonical_entity`, read with `audit.LatestByScope`). Line 16 says conclusion-layer work carries no Source scope; it stores a `canonical_entity` scope instead. In §4.1 line 330, rewrite the second sentence as: "`Record` stores one scope row per id each `scopeResolvers` entry resolves. Each resolver walks a change through its effect path using the change's own fields, the live row, or — for parents deleted in the same revision — the old values of that parent's delete change."
Check: scripts/check-pr.sh PASS.

**F-3 — Landed note and the nil-querier case** (from CON-1)
Files: `docs/deployment-plan/spike-10/prs/S10-21-handle-last-activity.md`, `core/database/audit/scopes_test.go`.
Failing test first: `TestLatestByScope` — `audit.LatestByScope(nil, audit.ScopeCanonicalEntity)` → `ErrInvalid` (expected to pass already; if it fails, D-5's nil check is missing and this step adds it).
Do: add the case. Add a `## Landed` section at the end of the plan, in the S10-20 plan's shape, listing every difference from the plan: T-3 promotes and edits a place Subject's toponym, not a person's (a toponym is not valid on a person); T-4 builds the Event from a single "Fire" `event_name`, not `TestListEvents`' `fire` inputs; anything else the implementer knows differs. Plus the fix-round additions (F-4, F-5).
Check: the test passes; scripts/check-pr.sh PASS.

**F-4 — Test that an evidence pin moves its handle** (from CON-2)
Files: `core/database/effects/handle_paths_test.go`.
Do: new `TestHandlePaths` subtest "evidence pin reaches its handle": `rowchange.Change{EntityType: "identity_claim_evidence", EntityID: <alice's claim id>, Action: rowchange.ActionCreate, Fields: map[string]rowchange.FieldDiff{"identity_claim_id": {New: dashed(<alice's claim id>)}}}` → `mustHandleScopes` = `{aliceRes.Entity.ID}`. Alice's claim id is `aliceRes.Claim.ID`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/...` passes.

**F-5 — `TestSourceScopes` uses the seed's write helpers** (from MNT-3)
Files: `core/database/audit/scopes_test.go`.
Do: replace the inline `"update_observation"`, `"delete_observation"`, and `"promote_subject"` `writes.Call` blocks in `TestSourceScopes` with `s.updateObservation`, `s.deleteObservation`, and `s.promote`; move those three methods up beside the other `scopeSeed` methods. Every `TestSourceScopes` case keeps its assertions.
Check: `go test ... -run 'TestSourceScopes|TestHandleScopes'` passes with the same subtests; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `0e801854`. Gate PASS (orchestrator run). F-1, F-2, F-4, F-5 done. F-3: the nil-querier case passed on first run (D-5's check was already there); the Landed note is added.

## Round 2 — review
Lenses with Fix-now findings: CON, MNT. Scope `b3b1f104...0e801854`.
CON: clean — CON-1 and CON-2 resolved (`TestLatestByScope` nil case, `TestHandlePaths/evidence pin reaches its handle`, Landed note). No regressions from F-1, F-2, F-5.
MNT: clean — MNT-1, MNT-2, MNT-3, and the routed §4.1 sentence resolved.

## Result
Gate PASS · CI green (Go tests, macOS tests, localizable) · open: none · follow-ups: none (MNT-4 rejected per D-6) · to user: —
