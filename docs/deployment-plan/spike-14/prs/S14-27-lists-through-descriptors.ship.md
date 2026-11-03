# Ship log — S14-27 Every write reaches its keys through their descriptors

**Plan:** [`S14-27-lists-through-descriptors.md`](S14-27-lists-through-descriptors.md) · **Branch:** `s14-27-lists-through-descriptors` · **PR:** #693
**State:** ready

## Round 0 — implement
Base `main` @ `42eb321c` (S14-16, PR #686, merged 2026-11-02; Planned against its branch head `a5e057b`). Head `52a9d0d6`. Gate PASS (orchestrator run; xcode-test 47s). Report deviations: base check run against `42eb321c`, not `a5e057b`; no `Workspace/Session` changes between them, and every §4 row held. AC-2.5's test does not count graph loads because `FakeStore` has no graph-load counter: it asserts that the graph handle is fetching, and the Landed note says so. T-3 reuses `seedStore`'s `g1`. Unused `graph2`/`registry` locals were removed after swift-warnings flagged them. Manual check §10 not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX (`macos/App/**`), SEC+PRF (`api/ffi/handlers/connect_test.go`; cache/registry code; `Features/**` models). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | minor | T-7 marshals/unmarshals by hand instead of `mustRPC` | Fix now → F-1 | verified: `connect_test.go:252-284` hand-rolls two RPC round trips; `harness_test.go:118` `mustRPC`; `promote_batch_test.go:117` already calls it for `CreateCitedBridge` |
| MNT-2 | MNT | nit → minor | Two stale comments over the mutation-only assertion | Fix now → F-2 | verified: `CatalogQueryRegistryTests.swift:891-898`; first comment describes effects over a mutation-only `#expect`, second (edited by this PR) has no assertion under it. Stale comment in a changed file |
| MNT-3 | MNT | nit → minor | Same `precondition` message written four times | Fix now → F-3 | verified: `CatalogLastUpdated.swift:92,187,208,219`, all four added by this PR. Same-pass cleanup |

CON, UX, SEC+PRF: no findings, no nits.

### Fix plan 1
**F-1 — T-7 through `mustRPC`** (from MNT-1)
Files: `api/ffi/handlers/connect_test.go`.
Do: in `TestDeleteBridgeReportsPropertyUsage`, replace both marshal/call/unmarshal blocks with `mustRPC(t, CreateCitedBridge, &engine.CreateCitedBridgeRequest{…}, &bridge)` and `mustRPC(t, DeleteSubject, &engine.DeleteSubjectRequest{…}, &resp)`, as `promote_batch_test.go:117` does. Keep the `contains(…, "property_usage")` assertion and the test's doc comment.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/ -run TestDeleteBridgeReportsPropertyUsage` passes; scripts/check-pr.sh PASS.

**F-2 — Comments that match their assertions** (from MNT-2)
Files: `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`.
Do: around lines 891-898, rewrite the comment over `let updated = registry.invalidations(by: .updatedProperty, …)` so it says only what the assertion checks (a mutation names no graph key; the graph is reached through `refreshes`, pinned by `propertyUsageAndGraphEffectsReloadEachOnce`). Delete the dangling "A Subject delete reports its handles, its graph and `property_usage` … names no mutation." lines, or fold their point into that one comment. No assertion changes.
Check: scripts/check-pr.sh PASS.

**F-3 — One `precondition` for the four id-slot constructors** (from MNT-3)
Files: `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift`.
Do: add one `fileprivate` helper (for example `requireRowIDs(_ follows: CatalogQueryFollows)`) that holds the `precondition(!follows.isCatalogWide, "a catalog-wide slot names no ids; use applyHiddenPolicyOnAny(follows:)")`, and call it from the four sites (lines 92, 187, 208, 219). Behavior unchanged; `noDescriptorFollowsCounts` still builds every descriptor.
Check: `grep -c 'a catalog-wide slot names no ids' macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` prints `1`; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `4b67f026`. Gate PASS (orchestrator run; xcode-test 51s). F-1 done (`02b1488c`), F-2 done (`fb517a7c`), F-3 done (`4b67f026`; `grep -c` prints `1`). MNT-1, MNT-2, MNT-3: **Resolved**, read by the orchestrator in `52a9d0d6..4b67f026`.

## Round 2 — review
Not run: round 1 had no Fix-now blocker or major.

## Result
Gate PASS · `main` unchanged since branch cut (`42eb321c`) · CI green (test 3m27s, test 8m38s, localizable) · open: none · follow-ups: none · to user: — · CON miss: none
