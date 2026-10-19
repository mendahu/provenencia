# Ship log — S12-16 Add Source reports its effects

**Plan:** [`S12-16-add-source-effects.md`](S12-16-add-source-effects.md) · **Branch:** `s12-16-add-source-effects` · **PR:** #492
**State:** ready

Worktree: `/Users/jakerobins/Repos/provenencia-s12-16`, cut from `main` @ `2b14ad10`.

## Round 0 — implement
Head `adadc2a3`. Gate PASS (orchestrator run). PR #492 (draft). Report deviations: `onNavCountsMoved` marked `@ObservationIgnored` (noted in plan's Landed note); Status line uses `Landed in [#492](url) (date)` form for the plan checker; `SourcesModel.create()` change made in the STEP-3 commit group; manual check not run. The implementer committed this ship log in `d16582c7` (said it left it untouched); harmless, it is committed at finish anyway.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (`macos/App/**` → UX; `api/ffi/**` → SEC; `macos/App/Features/**` models → PRF).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, MNT-1 (+UX outside-lens) | CON, MNT | major | `insertSource` inserted between `bumpSource` and its `@discardableResult` and doc block | Fix now → F-1 | verified: `FakeStore.swift:3315-3336` at head; `bumpSource` has no attribute, six call sites discard its result |
| UX-1 | UX | major | Opening the Sources list after Copy (from a page) or Delete republishes the stale cached count; the reload's `.ready → .ready` does not fire `onChange(of: status)` | Fix now → F-2 | read: `SourcesListView.swift:116-126`, `QueryHandle.beginLoading(stale:)` keeps `.ready`, `applySuccess` bumps `revision`. Keeps D-7 (`syncCatalogCounts()` still runs on appear and on load); the PR's own skill line ("Publish from a list only on load") names this case. Failing test first verifies it |
| MNT-2 | MNT | major | Three more private `waitUntil` copies (six in the target) | Fix now → F-3 | verified: `grep "func waitUntil" macos/ProvenenciaTests` → 6; the three new ones are byte-identical to `PromoteModelTests:11`. The plan's "copy" is test-step wording, not a D-# Decision, Scope or AC |
| MNT-3 | MNT | minor | `apply`'s doc no longer names its last step | Fix now → F-4 (rides along; F-2 is in the same area) | read: `WorkspaceSession.swift:164-176` |
| UX-2 | UX | minor | A failed recount after Add, Copy or Delete is silent | Rejected | D-8 decided this ("Failing the write, or a toast — the write succeeded" rejected) |
| PRF-note | SEC+PRF (outside lens → CON) | minor | Two recount tasks close together could finish out of order | Follow-up | needs two Source writes within one FFI round trip; Go `catalogsession` serializes the ops. Not reproduced; list in PR follow-ups |

Nits (logged): `MNT-4 · SourcesModelTests.swift:815-819,844-848 · repeated list-handle wait block; a file-local helper` · `UX-3 · SourcesModel.swift:420 · syncCatalogCounts doc should say when it is safe to call` (folded into F-2)

### Fix plan 1

**F-1 — Put `@discardableResult` and its doc back on `bumpSource`** (from CON-1, MNT-1)
Files: `macos/App/Platform/FakeStore.swift`.
Failing test first: none (attribute placement; no test can catch it). Check by build warnings instead.
Do: move `insertSource` and its two-line doc comment ("Appends a new Source…") above the `/// Runs body under stateLock…` doc block, so that block and `@discardableResult` sit directly above `private func bumpSource`. Change no code inside either function.
Check: `xcodebuild` shows no "result of call to 'bumpSource(_:evidence:)' is unused" warning; scripts/check-pr.sh PASS.

**F-2 — Publish the Sources count from the list only when a load lands** (from UX-1, UX-3)
Files: `macos/App/Features/Sources/SourcesListView.swift`, `macos/App/Features/Sources/SourcesModel.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift` (or a session-level test in the same target). Skill: `add-workspace-place`.
Failing test first: `SourcesModelTests.hiddenListRepublishesCountAfterReload` (new). Warm the Sources list with one Source and a `CatalogCounts` on the model; move the session to another location so the list is hidden; append a Source to the FakeStore and `session.apply(.createdSource)`; then go back to `.sectionRoot(.sources)` and drive the same publish path the view uses (factor the view's two publish triggers into one `SourcesModel` method the test can call, e.g. `syncCatalogCountsIfLoaded(handle)` that publishes only when `status == .ready && !isFetching`). After `waitForQuery`, assert `counts.sources == model.sources.count == 2`. Run it red against today's trigger (it publishes 1) before the fix.
Do: in `SourcesListView`, replace `.onChange(of: sourcesHandle.status)` with `.onChange(of: sourcesHandle.revision)`, and publish only when `status == .ready && !isFetching`. Guard the `onAppear` publish with `!sourcesHandle.isFetching` as well. Reword the `syncCatalogCounts()` doc to say it publishes the count of a settled list: call it only when the list handle is `.ready` and not refetching.
Check: the new test passes; existing `SourcesModelTests` pass; scripts/check-pr.sh PASS.

**F-3 — One shared `waitUntil` test helper** (from MNT-2)
Files: `macos/ProvenenciaTests/QueryHandleWaiting.swift`, `WorkspaceSessionTests.swift`, `SourceCopyModelTests.swift`, `SourcePageModelTests.swift`, `PromoteModelTests.swift`.
Failing test first: none (test refactor). The existing tests are the check.
Do: add a module-level `@MainActor func waitUntil(_ condition: () -> Bool) async -> Bool` to `QueryHandleWaiting.swift` with `PromoteModelTests.waitUntil`'s body, with a one-line doc. Delete the private copies in the four files above. Leave `EvidenceGraphModelTests` and `CitationAutoTranscribeTests` alone: their signatures differ (logged follow-up).
Check: `grep -rn "func waitUntil" macos/ProvenenciaTests` → the shared one, EvidenceGraphModelTests and CitationAutoTranscribeTests only; all Swift tests pass; scripts/check-pr.sh PASS.

**F-4 — Name the recount in `apply`'s doc** (from MNT-3)
Files: `macos/App/Features/Workspace/Session/WorkspaceSession.swift`.
Do: extend the last sentence of the `apply(_:_:)` doc: "…graph progress for the Sources the effects and the mutation moved, and runs `onNavCountsMoved` for a mutation in `CatalogQueryRegistry.navCountTriggers`."
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `ffa3cdef`. Gate PASS (orchestrator run). F-1 `dd0cdc65`, F-2 `c9b323a9`, F-3 `3ca8fbda`, F-4 `ffa3cdef` done. F-2 deviation: a new `SourcesModel.syncCatalogCountsIfLoaded(_:)` is the view's one publish path (`onAppear` and `onChange(of: revision)`); its test `hiddenListRepublishesCountAfterReload` covers the guard (red against a `status == .ready`-only guard), while the view's `status` → `revision` trigger switch has no unit test (SwiftUI `onChange`).

## Round 2 — review
Reviewers: CON (Sonnet), MNT, UX. Scope `adadc2a3...ffa3cdef`. Not re-sent: SEC+PRF (no findings in round 1).

Round-1 majors: CON-1/MNT-1 Resolved (CON, MNT), UX-1 Resolved (CON, UX), MNT-2 Resolved (CON, MNT). Minor MNT-3: Resolved (orchestrator read `WorkspaceSession.swift:174-177`; CON and MNT agree). No new blockers.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-2 (= MNT and UX notes) | CON | minor | F-2's new test went in under `writeActivityMovesTheRowInTheUpdatedSort`'s doc comment, so that comment now documents the new test | Follow-up | round 2 fixes only blockers; one-line comment move in `SourcesModelTests.swift:365` |
| MNT-note | MNT | minor | `PropertiesView` and `MetadataView` still publish counts on `onChange(of: status) == .ready` with no `isFetching` guard: the shape F-2 replaced for Sources | Follow-up | not code this PR touches; candidate for one settled-list publish rule across lists |
| UX-note | UX | — | `onAppear` may run before `.task { warmListQueries() }` starts the refetch, briefly publishing the cached count until the reload's `revision` bump corrects it | Logged | low confidence; self-correcting; the manual check (step 5–6) will show it |

Nits (logged): `CON-3 · SourcesModelTests.swift:380-395 · add #expect(handle.isFetching) before the sentinel check`

## Result
Gate PASS · CI green (round 2 head `f2aa0ebb`; re-runs on round 3) · open: none · follow-ups: PropertiesView/MetadataView settled-list publish (in its own session), recount task ordering (round 1 PRF note), MNT-4 (repeated list-handle wait block) · done in round 3: CON-2, CON-3, last `waitUntil` copies · to user: — · CON miss: none

## Round 3 — user-requested follow-ups
After the PR was marked ready, the user asked (2026-10-19) to fold two follow-ups into this PR: CON-2, the misplaced test doc, and the leftover `waitUntil` copies in `EvidenceGraphModelTests` and `CitationAutoTranscribeTests`. This is a fix round only, with no review round, since both changes are test-only.

### Fix plan 3

**F-5 — Put the Updated-sort doc back on its own test** (from CON-2)
Files: `macos/ProvenenciaTests/SourcesModelTests.swift`.
Do: move `/// A write's activity moves the row in the Updated sort without a reload.` from above `hiddenListRepublishesCountAfterReload` to directly above `@Test func writeActivityMovesTheRowInTheUpdatedSort`. Give `hiddenListRepublishesCountAfterReload` its own one-line doc: `/// A hidden list publishes its count only once its reload lands, not from the stale cached rows.` While in the test, add `#expect(handle.isFetching)` before the sentinel check if the handle is in scope there (CON-3 nit).
Check: scripts/check-pr.sh PASS.

**F-6 — Move the last two `waitUntil` copies onto the shared helper** (from MNT-2 follow-up)
Files: `macos/ProvenenciaTests/QueryHandleWaiting.swift`, `EvidenceGraphModelTests.swift`, `CitationAutoTranscribeTests.swift`.
Do: change the shared helper to `func waitUntil(timeout: Duration = .seconds(2), _ condition: () -> Bool) async -> Bool`. Poll every 10 ms against a `ContinuousClock` deadline, as `CitationAutoTranscribeTests.waitUntil` does, and update its doc. Existing trailing-closure callers stay unchanged. Delete the private copies in both files. `EvidenceGraphModelTests`'s copy defaulted to 5 s, so its call sites pass `timeout: .seconds(5)` to keep their budget. `CitationAutoTranscribeTests`'s copy matched the 2 s default, so its call sites stay as they are.
Check: `grep -rn "func waitUntil" macos/ProvenenciaTests` → only `QueryHandleWaiting.swift`; scripts/check-pr.sh PASS.

## Round 3 — fix
Head `a1621338`. Gate PASS (orchestrator run). F-5 `2bb7e41e`: the Updated-sort doc is back on its own test, and the hidden-list test has its own doc plus `#expect(handle.isFetching)` (CON-3). F-6 `a1621338`: one shared `waitUntil(timeout:_:)`; `EvidenceGraphModelTests` passes `timeout: .seconds(5)` at its 9 call sites; `grep "func waitUntil"` → `QueryHandleWaiting.swift` only.
