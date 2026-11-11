# Ship log — S16-09 Conclusion pages refresh from the Sources they cite and show their last value while they refresh

**Plan:** [`S16-09-source-follow-and-stale-pages.md`](S16-09-source-follow-and-stale-pages.md) · **Branch:** `s16-09-source-follow-and-stale-pages` · **PR:** [#780](https://github.com/mendahu/provenencia/pull/780)
**State:** ready

## Before round 0

- Branch base: `main` @ `43b7247f` (plan planned against `cd03582`; the planning PR #779 merged into `43b7247f`).

## Round 0 — implement
Head `33bd9403` (code head `50f993fb`). Gate PASS (orchestrator run on `33bd9403`). PR [#780](https://github.com/mendahu/provenencia/pull/780) (draft). Report deviations: version bump 0.8.0 → 0.8.1 (main was already 0.8.0, plan said 0.7.1); STEP-2–4 in one commit `9ac956f8` (same descriptors and tests); two test expectations changed beyond the plan's list (`invalidationsComeFromRegistrySpecs` refreshes gain `.source`; `applyTriggerStalesHiddenDetailPagesAndRevalidatesTheVisibleOne` expects stale, not nil); manual check not run. Landing commit `33bd9403` not pushed: `git push` fails from the sandbox (SSH host not resolvable).

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. UX: diff touches `macos/App/**`. SEC+PRF: cache code in `macos/App/Features/Workspace/Session/**` (hidden-entity cap, eviction loops).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `sourceScopes` and `CatalogWriteEffects` comments say a Source write reloads nothing / drops reads | Fix now → F-1 | verified: `CatalogMutation.swift:91-92` ("drops those cached reads") and `:111-112` ("it reloads nothing"); `.conclusionDetail` now follows `.source`; rule 2 in `macos-client-patterns.md` still says activity only |
| MNT-2 | MNT | minor | `deletedWith == .handle` spelled twice in the session to mean "page the cap counts" | Fix now → F-2 | verified: `WorkspaceSession.swift:236`, `:251`; code this PR adds |
| MNT-3 | MNT | minor | `trimHiddenEntities` and `evict(_:ids:)` repeat one filter over handles with two meanings of on screen | Fix now → F-2 | verified: `WorkspaceSession.swift:249-257`, `:507-512`; code this PR adds |
| CON-1 | CON | minor | AC-1.3's "holds the new value" passes without a reload (seed equals cached value) | Fix now → F-3 | test this PR adds; reseeding a changed value is how `backToAStalePageShowsItsValueWhileItRefreshes` already proves it |
| CON-2 | CON | minor | AC-3.2's "its reload reads gone" is not exercised | Fix now → F-3 | verified: e3 stays seeded, test ends at `isFetching` |
| UX-1 (+CON-3) | UX | minor | Manual check §10 not run | Rejected | ship-pr step 8: the user runs the manual check; the report and PR body carry §10's steps |
| SEC/PRF | SEC+PRF | — | none | — | clean |

Nits (logged): `MNT-4 · WorkspaceSessionTests.swift:1248 · T-2 builds its detail inline beside citedDetail(_:)` · `MNT-5 · WorkspaceSession.swift:512 · drop ids from shownOrder only when entity == .handle` (rides on F-2) · `CON-4 · UnpromoteFlowTests.swift:218 · comment wording differs from plan` · `UX-2 · ConclusionDetailPage.swift:312 · a merged entity's stale page stays clickable for one reload (plan accepted)`

### Fix plan 1

**F-1 — Say what a Source write and a write's effects now do** (from MNT-1)
Files: `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `docs/macos-client-patterns.md`.
Failing test first: — (comments and docs)
Do: `CatalogWriteEffects` struct comment: the session applies each cached key's hidden policy to the reads a write moved (a visible key reloads, a hidden one is marked stale) and evicts the hidden keys of what `deleted` names; it patches rows whose last activity moved. `sourceScopes` comment: the session patches those rows in the cached Sources list with `activity`, and the detail keys whose records cite one of those Sources follow it through `.source`. Extend rule 2's `sourceScopes` parenthetical in `macos-client-patterns.md` the same way.
Check: `rg "reloads nothing|drops those cached reads" macos/App` prints nothing; scripts/check-pr.sh PASS.

**F-2 — One named membership and one walk for per-entity keys** (from MNT-2, MNT-3, MNT-5)
Files: `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` (descriptor), `macos/App/Features/Workspace/Session/WorkspaceSession.swift`.
Failing test first: — (refactor; `hiddenConclusionPagesAreCappedAtFifty`, `deletedEntitiesLoseTheirHiddenKeys` stay green unchanged)
Do: add a documented computed property on `CatalogQueryDescriptor` (e.g. `isEntityPage`: `deletedWith == .handle`, the keys `shownOrder` and the hidden-entity cap count) and use it at both session sites (`apply(location:)` and `trimHiddenEntities`). Add one private helper on the session that yields `(key, ownId)` for this project's warmed keys whose descriptor `deletedWith` is a given entity; use it from `trimHiddenEntities` and `evict(_:ids:)`. In `evict(_:ids:)`'s comment, say it works key by key (a visible key stays) while the cap works entity by entity. Remove ids from `shownOrder` only when `entity == .handle`.
Check: `rg "deletedWith == \.handle" macos/App/Features/Workspace/Session/WorkspaceSession.swift` prints nothing; `rg "for key in .*handles.keys" macos/App/Features/Workspace/Session/WorkspaceSession.swift` no longer shows the two filters in `trimHiddenEntities` and `evict(_:ids:)`; scripts/check-pr.sh PASS.

**F-3 — Make the two AC tests prove their AC** (from CON-1, CON-2)
Files: `macos/ProvenenciaTests/WorkspaceSessionTests.swift`.
Failing test first: `detailFollowsTheSourcesItsRecordsCite` — before the `sourceScopes: ["s-pic"]` apply, reseed the store with a changed detail (e.g. `var renamed = citedDetail("e1"); renamed.memberCount = 2`) and expect `handle.value == renamed` after `settleWrites`. `deletedEntitiesLoseTheirHiddenKeys` — before the second apply set `store.seededConclusionDetails["e3"] = nil`; after it `await session.settleWrites()` and expect `ConclusionDetailPresentation(visible!.reading, isFetching: false) { $0.entity.id } == .gone`. (Each should pass on the existing code; show it would fail if the reload did not run, e.g. by confirming the new expectation differs from the pre-apply value.)
Do: the test edits above only.
Check: both tests pass; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `baa1b9b4` (commits `1d0c4288` F-1, `0445969c` F-2, `baa1b9b4` F-3; push now works and carried `33bd9403`). F-1, F-2, F-3 done.
- Implementer deviation: F-1's `rg` still hits `engine.proto:768` / generated `engine.pb.swift`. That comment is on `handle_scopes` ("it reloads nothing and leaves detail pages alone"), which stays true: `handle_scopes` only patches list rows. The `source_scopes` proto comment does not claim "reloads nothing". No change.
- CON-1, CON-2 (minors, orchestrator read): **Resolved** — `detailFollowsTheSourcesItsRecordsCite` reseeds `renamed` and expects it (and `!= citedDetail("e1")`); `deletedEntitiesLoseTheirHiddenKeys` clears e3, settles and expects `.gone`. The implementer reports the reseed first failed the later hidden-stage expectation, showing the reload is real.

## Round 2 — review
Reviewers: MNT (had Fix-now major MNT-1). Not sent: CON (minors only, resolved above), UX (no Fix-now), SEC+PRF (clean).

MNT-1: **Resolved** (round-2 reviewer: comments at `CatalogMutation.swift:92-95`, `:114-115` and rule 2 now match `.applyHiddenPolicy(followsValue: .source)`). MNT-2, MNT-3, MNT-5 (minors, orchestrator read): **Resolved** — `isEntityPage` on the descriptor, one `warmedKeys(deletedWith:)` walk serving `trimHiddenEntities` and `evict(_:ids:)`, `shownOrder` trimmed only for `.handle`. No new blocker.

Nits (logged): `MNT-6 · WorkspaceSession.swift:250 · trimHiddenEntities counts pages through warmedKeys(deletedWith: .handle), not isEntityPage` · `MNT-7 · engine.proto:774-775 · source_scopes comment could add that citing detail keys follow it`

Gate PASS on `baa1b9b4` (orchestrator run).

## Finish
- `origin/main` moved to `84ed2753` (S15-06 plan docs only; no file this PR touches). `git` could not reach github.com from the sandbox for a while; the user pushed the ship log, then fetch worked and `origin/main` was merged (`36c2b0e4`). Gate PASS on the merged head.
- `gh pr ready 780`; CI green (localizable, test ×2).

## Result
Gate PASS (`baa1b9b4`) · CI green · open: none · follow-ups: none · to user: — · nits logged: MNT-4, MNT-5 (fixed in F-2), MNT-6, MNT-7, CON-4, UX-2
