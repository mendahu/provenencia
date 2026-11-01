# Ship log — S13-01 Person claims read

**Plan:** [`S13-01-person-claims-read.md`](S13-01-person-claims-read.md) · **Branch:** `s13-01-person-claims-read` · **PR:** #655
**State:** ready

## Round 0 — implement
Head `38e66e9e`. Gate PASS (orchestrator run). Draft PR #655. Report deviations: (1) T-8 warms a hidden `.personClaims` key — no place shows it until S13-16, so a visible reload-once is not observable; (2) T-6 asserts 0 graph queries on the warm call; the counter does not count `files.RelPaths`, and "at most one query" rests on one `listedProtos` call; T-6 assigns the `person` Property to the Person type; (3) `CatalogQueryRefresh` has two initializers; `revalidatedKeys(for:_:)` takes an optional mutation; (4) the working rule "One invalidation rule for person-section keys" was already right, left unchanged; (5) `persondetail` deps include `api/proto/engine` through `apperr`, as `conclusiondetails` does.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (`macos/App/**` routes UX; `api/ffi/**` and `core/database/**` route SEC+PRF).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, MNT-4 | CON, MNT | blocker | `add-workspace-place` SKILL.md:27 and `docs/macos-client-patterns.md:87` still name the descriptor's old `refresh` field and the fan-outs in `invalidations` | Fix now → F-4 | verified: both lines say `` `refresh` ``; the field is `refreshes:` at head |
| MNT-1 (+ SEC outside-lens note) | MNT | blocker | `ListPersonClaims` filters each claim's Observations again instead of reading `Claim.Observations` | Fix now → F-2 | verified: `person_detail.go:41-50` passes a Subject-id `keep` to `lc.rows`; `Claim.Observations` is read only in `claims_test.go`. Same rule in two layers; same-pass cleanup. Supersedes D-10's "keep = Subject id match" detail; D-10's one mapping path and one `listedProtos` call stay |
| MNT-2 | MNT | blocker | The page-handle guard (nil catalog/graph, id length, `Node`, merged → `ErrNotFound`) is copied from `conclusiondetails.ForEntity` into `persondetail.Claims` | Fix now → F-1 | verified: `conclusiondetails.go:169-178` and `claims.go:41-52` match. Same-pass cleanup; S13-04/05/27 read through it next |
| MNT-3 | MNT | major | The `EvidenceSubject` → `subjects.Subject` literal is in `ListSubjects` and `ListPersonClaims` | Fix now → F-3 | verified: `subjects.go:166`, `person_detail.go:74`. Same-pass cleanup; supersedes D-11's "repeated, not extracted" |
| MNT-5 | MNT | minor | `memberGraph.member` is set and never read; D-2 names the `EvidenceMember` | Fix now → F-5 | verified: only the literal at `claims.go:133` sets it. Minor in a file this PR adds |
| CON-2 (+ UX, SEC outside-lens notes) | CON | minor | `.personClaims`'s `.evidenceGraphs` refresh names no graph while the key's first load is in flight, so a graph write in that window leaves it stale | To user → Fix now → F-6 | verified: the closure returns `[]` for a nil value (`CatalogQueryRegistry.swift` `.personClaims` arm); `WorkspaceSession.revalidate` would cancel and restart the in-flight load if named. D-13 says "empty when it holds no value", so the fix contradicts a Decision. **User, 2026-10-31:** a second window can write to the open catalog, so the gap is real; fix it in the refresh helper → F-6 |
| PRF-1 | PRF | minor | A person with more than 64 distinct Evidence graphs cycles the shared graph LRU on every read | Follow-up | past Scale: the map's Slice 7 Scale is 50 graphs (`deployment-plan.md:34`) < `maxEvidenceGraphs = 64`. Revisit if the person budget rises |

Nits (logged): `MNT-6 · macos/App/Platform/GoStore.swift:2120 · mapSubject can be private static again` · `MNT-7 · core/database/graphcache/evidence_graph.go:92 · rewrap the long doc-comment line` · `MNT-8 · api/ffi/handlers/person_detail.go:107-108 · one index over counts instead of counts + next`

### Fix plan 1
Supersedes D-10's `keep` detail (F-2) and D-11's "repeated, not extracted" (F-3) under the same-pass cleanup rule; record both in the plan's Landed note.

**F-1 — One page-handle lookup** (from MNT-2)
Files: `core/database/conclusiondetails/conclusiondetails.go`, `core/database/persondetail/claims.go`.
Failing test first: none (refactor); `TestClaimsMissingAndEmpty` and the `ForEntity` not-found tests stay green.
Do: export `func Handle(c *database.Catalog, entityID []byte) (*graphcache.Node, error)` in `conclusiondetails` holding the guard at `conclusiondetails.go:169-178`: nil catalog or graph, `len(entityID) != 16`, unknown or merged handle → `ErrNotFound`; a `Node` error passes through. `ForEntity` and `persondetail.Claims` call it.
Check: `grep -rnE 'n == nil \|\| n\.Merged' core/database/conclusiondetails core/database/persondetail | grep -v _test.go` prints only `Handle`'s line and the loop at `ForEntity`'s merged-skip (`seen[...]`); scripts/check-pr.sh PASS. Other `n.Merged` sites (`conclusionheaders/compose.go`, `promotealign`) return empty or `ErrInvalid`, not `ErrNotFound`; leave them.

**F-2 — The handler reads `Claim.Observations`** (from MNT-1)
Files: `api/ffi/handlers/source_read.go`, `api/ffi/handlers/person_detail.go`.
Failing test first: none (refactor); `TestListPersonClaims` and the `listedProtosFromGraph` callers' tests stay green.
Do: add `(lc listedContext) rowsOf(eg graphcache.EvidenceGraph, obs []graphcache.EvidenceObservation) ([]observations.Listed, error)` (Subject labels from `eg.Subjects`, `listedFromSource` per row). `rows(eg, keep)` filters then calls `rowsOf`, so there is one mapping path. `ListPersonClaims` passes `cl.Graph, cl.Observations`; it keeps one `listedProtos` call per response.
Check: `grep -n "bytes.Equal" api/ffi/handlers/person_detail.go` prints nothing; scripts/check-pr.sh PASS.

**F-3 — One `EvidenceSubject` → proto mapping** (from MNT-3)
Files: `api/ffi/handlers/subjects.go`, `api/ffi/handlers/person_detail.go`.
Failing test first: none (refactor); `ListSubjects` and `TestListPersonClaims` tests stay green.
Do: add `func evidenceSubjectProto(s graphcache.EvidenceSubject) *engine.Subject` beside `subjectProto` in `subjects.go`; use it in `ListSubjects` and `ListPersonClaims`.
Check: `grep -rn "subjectProto(subjects.Subject{" api` prints only the line inside `evidenceSubjectProto`; scripts/check-pr.sh PASS.

**F-4 — Guidance names `refreshes`** (from CON-1, MNT-4)
Files: `.agents/skills/add-workspace-place/SKILL.md`, `docs/macos-client-patterns.md`.
Do: SKILL.md:27 `patches` / `refresh` → `patches` / `refreshes`. `macos-client-patterns.md:87`: each kind's `refreshes` follow its read; the Evidence graph, list, detail and person-claims keys follow `evidenceGraphs` and `touchedHandles` through their descriptor `refreshes`, revalidated once alongside the registry's invalidations in `apply`; point to `add-workspace-place` Refresh.
Check: ``rg '`refresh`' .agents .cursor docs/*.md`` prints nothing; scripts/check-pr.sh PASS.

**F-5 — `memberGraph` keeps one member field** (from MNT-5)
Files: `core/database/persondetail/claims.go`.
Do: keep the `EvidenceMember` as D-2 says (it carries `ClaimID`) and drop the separate `claimID` and the unused `graphcache.Member`; read `ClaimID` from it.
Check: `go test ./core/database/persondetail/...` passes; scripts/check-pr.sh PASS.

**F-6 — A refresh that follows its cached value follows everything until the value arrives** (from CON-2; user, 2026-10-31)
Files: `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` (`CatalogQueryRefresh` and the `.personClaims` arm), `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` if the helper needs a registry case, `.agents/skills/add-workspace-place/SKILL.md` (Refresh row).
Failing test first: `WorkspaceSessionTests.personClaimsFollowEveryGraphUntilLoaded` — a warmed `.personClaims` e1 with no value yet (first load still in flight); effects `evidenceGraphs` [g9] → `revalidatedKeys` names it. Once it holds a value with g1, `evidenceGraphs` [g9] → not named, [g1] → named (T-9's case keeps passing).
Do: add a `CatalogQueryRefresh` form for ids read from the key's cached value, e.g. `applyHiddenPolicy(followsValue: slot) { (value: Value) -> Set<ID> }`. The helper reads `session.queryHandle(key)?.value`; when it is nil it matches any id in the slot, as `applyHiddenPolicyOnAny` does; otherwise it matches the closure's ids. The `.personClaims` `.evidenceGraphs` refresh uses it; its closure takes the groups and returns their claims' graph ids. Reuse the existing match path in `revalidatedKeys`, and add no second dedupe. Amend D-13's "empty when it holds no value" in the plan's Landed note (user, 2026-10-31). Add one line to the `add-workspace-place` Refresh row: a refresh whose ids come from the cached value uses `followsValue`.
Check: the new test and T-9 pass; `grep -n "queryHandle(key)?.value" macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` prints only the helper's line; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `9f9d32af`. Gate PASS (orchestrator run). F-1 – F-6 done. Report deviations: `5b3aee05` (F-2, F-3) did not build on its own; `40dcec6c` and `da7a1f8c` fix imports and gofmt. F-6's test was written after the fix, so it was not seen red first; the old closure returned `[]` for a nil value, so its g9 case would have failed. T-9's trailing "no value yet holds no graph" assertion moved into the new test, because F-6 reverses it. The plan's Landed note gains items for D-10, D-11 and D-13.

Minors read by the orchestrator:
- MNT-5 → **Resolved.** `memberGraph` holds one `graphcache.EvidenceMember`, and `Claim.ClaimID` reads `m.member.ClaimID`.
- CON-2 → **Resolved.** `CatalogQueryRefresh.applyHiddenPolicy(followsValue:ids:)` in `CatalogLastUpdated.swift` returns `nil` (any id) until the key holds a value; the `.personClaims` arm uses it. `personClaimsFollowEveryGraphUntilLoaded` covers it.

## Round 2 — review
Reviewers: CON (Sonnet; CON-1 blocker), MNT (MNT-1, MNT-2 blockers, MNT-3 major). Not sent back: UX (no findings), SEC+PRF (PRF-1 minor, Follow-up). Scope `38e66e9e...9f9d32af`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| — | CON | — | CON-1 resolved; no new findings | — | `rg '`refresh`' .agents .cursor docs/*.md` prints nothing; the F-1/F-2/F-3/F-5 refactors keep behavior |
| — | MNT | — | MNT-1, MNT-2, MNT-3, MNT-4 resolved; no new findings | — | each F-# grep prints only the expected line; F-6 reuses the `.revalidate` nil-means-any path |

Nits (logged): none.

## Result
Gate PASS (on `9f9d32af`; `origin/main` unchanged since the branch was cut, `79630be9`) · CI <pending> · open: none · follow-ups: PRF-1 (more than 64 Evidence graphs per person cycles the graph LRU; revisit if the Slice 7 budget of 50 rises) · to user: CON-2 (answered 2026-10-31: fix in the refresh helper, F-6) · CON miss: none
