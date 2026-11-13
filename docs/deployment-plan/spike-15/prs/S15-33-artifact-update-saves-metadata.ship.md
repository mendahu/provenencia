# Ship log — S15-33 Editing an Artifact can save its metadata

**Plan:** [`S15-33-artifact-update-saves-metadata.md`](S15-33-artifact-update-saves-metadata.md) · **Branch:** `s15-33-artifact-update-saves-metadata` · **PR:** #806
**State:** ready

## Round 0 — implement
Head `b243ef591`. Gate PASS (orchestrator rerun). Report deviations: T-4 runs on `page2` (page1's layout already holds `provider` after the "empty clears" subtest; noted in the plan's Landed note); an `errParams` test helper in `sourcemetadata_test.go` and `handlers/metadata_test.go` instead of inline `errors.As`; large regenerated `engine.pb.go` descriptor diff. The implementer's first gate run failed `promotealign` `TestLargeSourceBudgets` under CPU contention from a concurrent run; clean rerun passed.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. UX routed by `core/apperr` (param on an existing error code) and `macos/App/Platform/Generated/engine.pb.swift`; SEC by `api/ffi/**`; PRF by `core/database/**`.

UX: no findings. SEC+PRF: no findings (traced `parseDraft`/`parseID`, `normalizeFor` after `loadFields`, param is a field-row UUID never shown; one board read per Artifact write, independent of the Source's Artifact count).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | minor | T-6 sends the unchanged label "P6", so AC-1.4's label rollback and "no revision" are untested with a draft | Fix now → F-1 | verified: `artifacts_test.go:504` sends `Label: "P6"`, the Artifact's existing label; only metadata is checked after |
| MNT-1 | MNT | minor | Three copies of the Artifact label lookup in `artifacts_test.go` | Fix now → F-2 (same-pass cleanup) | verified: `range ws.GetArtifacts()` at :294, :360, :469; this PR adds two |
| MNT-2 | MNT | minor | Board fetched three new ways plus older inline copies | Fix now → F-3 (same-pass cleanup) | verified: `GetMetadataBoardRequest{` at `artifacts_test.go:386,511`, `evidence_graphs_test.go:217,276`, `metadata_test.go:171,225,277` |
| MNT-3 | MNT | minor | T-5 reads params by hand beside `errParams` | Fix now → F-4 | verified: `artifacts_test.go:491` |

Nits (logged): `MNT-4 · artifacts_test.go:376 · update closure uses parent t` · `MNT-5 · artifacts.go:60,70,128,142 · owner built twice per handler` · `MNT-6 · sourcemetadata.go:955 · normalizeValue could return (string, bool) or fold into normalizeFor` · `MNT-7 · errParams copied in two packages` · `CON-2 · plan header: Landed 2026-11-12 vs Planned 2026-11-13 (plan commits carry 2026-11-13 author dates; Landed uses the session's date)`

### Fix plan 1

**F-1 — T-6 proves the label and revision roll back with the draft** (from CON-1)
Files: `api/ffi/handlers/artifacts_test.go`.
Failing test first: none (strengthens a passing test; the transaction already rolls back). Run it once with the new assertions to show it passes.
Do: in `TestUpdateArtifactMetadata` "a refused unlink drops the draft", send `Label: "P6 renamed"`. Record `latestRevisionTypes` before the call; after the refusal, assert the Artifact's label is still "P6" (via the F-2 helper) and `latestRevisionTypes` is unchanged.
Check: subtest passes.

**F-2 — One Artifact-label test helper** (from MNT-1)
Files: `api/ffi/handlers/artifacts_test.go` (or `metadata_test.go` beside `latestRevisionTypes`).
Do: add package-level `artifactLabel(t *testing.T, dir, sourceID, artifactID string) string` (GetSourceWorkspace, match id, return label; fail if absent). Replace `labelOf` in `TestUpdateArtifactGraphLinks`, `labelOf` in `TestUpdateArtifactMetadata`, and the inline loop in "nothing changed writes no revision".
Check: `grep -n "range ws.GetArtifacts()" api/ffi/handlers/*_test.go` prints only the helper's line (plus any loop that does something other than a label lookup — name it in the report).

**F-3 — One board-fetch test helper** (from MNT-2)
Files: `api/ffi/handlers/metadata_test.go`, `artifacts_test.go`, `evidence_graphs_test.go`.
Do: add `boardEntries(t *testing.T, dir string, owner *engine.MetadataOwner) []*engine.MetadataWorkspaceEntry` next to `assertReplyBoard`; `assertReplyBoard` calls it. Move the `board` closure (`artifacts_test.go:386`), the T-6 inline fetch (`:511`), `evidence_graphs_test.go:217,276` and `metadata_test.go:171,277` onto it. Leave `metadata_test.go:187` (expects an error, no owner).
Check: `grep -n "GetMetadataBoardRequest{" api/ffi/handlers/*_test.go` prints only the helper's line and `metadata_test.go`'s error-path call.

**F-4 — T-5 reads params through `errParams`** (from MNT-3)
Files: `api/ffi/handlers/artifacts_test.go`.
Do: `if !errors.Is(err, sourcemetadata.ErrInvalid) || !slices.Equal(errParams(t, err), []string{link})`, as T-7 does; drop the `apperr` import if unused.
Check: `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `456ef4fbb`. Gate PASS (orchestrator rerun). F-1 – F-4 done. Deviation: `TestGetMetadataBoard` (`metadata_test.go:171`) keeps its inline fetch because it also asserts `board.GetOwnerId()`, which `boardEntries` does not return; accepted (different shape, not a copy of the entries read).

## Round 2 — review
Not dispatched: round 1 had no Fix-now blocker or major. Orchestrator read the fix commit: CON-1 **Resolved** (T-6 sends "P6 renamed", asserts label "P6" and unchanged `latestRevisionTypes`); MNT-1 **Resolved** (`artifactLabel`, one `range ws.GetArtifacts()`); MNT-2 **Resolved** (`boardEntries`; remaining inline calls are the owner-id test and the error path); MNT-3 **Resolved** (`errors.Is` + `errParams`, `apperr` import gone).

## Result
`origin/main` unchanged since the branch was cut. Gate PASS · CI green (after the budget change) · open: none · follow-ups: none · to user: CI budget below (answered)

**CI.** `test` (app) and `localizable` pass. `Go tests` fails twice (run 43349656962, first attempt and rerun) on `core/database/promotealign` `TestLargeSourceBudgets/compose_every_summary`: 517ms then 507ms against a 500ms budget. Not this PR: the diff touches nothing in `promotealign`, and its only shared dependency is `core/apperr`, where this PR changed doc comments only (`go list -deps`). Locally on this head the subtest takes 176–181ms. The CI runner sits right at the budget. A fix round would mean changing an unrelated perf budget or test, which is outside this plan; taken to the user.

**User (2026-11-12):** "increase the budget to 600ms on this pr". `budgetComposeAll` in `core/database/promotealign/perf_test.go` raised from 500 to 600 ms with the CI measurements in its comment; 600 ms still fails a per-row place graph (about four times the 220 ms baseline). Gate PASS.
