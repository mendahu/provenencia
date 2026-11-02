# Ship log — S14-16 Promote and Source writes return their effects

**Plan:** [`S14-16-promote-source-effects.md`](S14-16-promote-source-effects.md) · **Branch:** `s14-16-promote-source-effects` · **PR:** #686
**State:** ready

## Round 0 — implement
Base `04a2c99b`.
Head `7fb7bb5e` (PR #686, draft). Report: gate PASS (implementer's run). Deviations: T-4 association case built as a Done minting both ends plus the association; T-3 cover case reverts artifact → type_icon; dead code left (`personBeside` uncalled, `SourcesModel.catalogCounts` stored unread, proto `IdentityClaim` unreferenced); several tests rewritten beyond those step 8 named (assertions on invalidations that no longer exist); `SourcesModelTests`/`CatalogCountsTests` publish assertions deleted (D-18). Manual check not run.

**Gate 1 (orchestrator) FAIL** — xcode-test: `CatalogCountsTests.newestRecountWins()` — `Expectation failed: started` at `CatalogCountsTests.swift:127` (the 5 s `waitUntil` for the first recount to claim its delay timed out under the parallel suite; the test ran 8.8 s). Sent back as gate-fix round 0.1.

## Round 0.1 — gate fix
Head `20dde76e`. G-1: `FakeStore.workspaceNavCountsDelays: [Duration]` replaced by `workspaceNavCountsHold = FakeHold()` (the existing hold pattern); test orders the recounts through the hold. **Gate 2 (orchestrator) PASS.** Deviation: the plan's D-10, hook table, step and T-12 still name the delay list; the hook is now a hold (test-only; shipped behavior unchanged). Noted for the PR body.

## Round 1 — review
Base `04a2c99b` · head `20dde76e`.
Reviewers: CON (Sonnet), MNT, UX (`macos/App/**`), SEC+PRF (`api/ffi/**`, `core/database/**`, `macos/App/Features/**` models). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, MNT-1a | CON, MNT | major | `SourcesModel.catalogCounts` stored, never read after D-18 | Fix now → F-1 | verified: `rg catalogCounts macos/App/Features/Sources/` finds only declaration, init, assignment, `SourcesListView` pass-through. D-18 orphaned it; dead code goes with the PR that made it dead |
| PRF-1 | PRF | major | Promote header runs three Sources-list scans per body pass; page now observes the session map and Sources list | Fix now → F-2 | verified: `PromoteView.header` (`PromoteView.swift:104`) is a computed property of the body and reads `sourceMark`, `sourceTitle`, `sourceRef`, each a `session.source(id:)` scan; body also reads `flow`. New code, per-edit cost linear in the catalog |
| CON-3, MNT-1b | CON, MNT | minor | `personBeside` uncalled; proto `IdentityClaim` unreferenced | Fix now → F-3 | verified by `rg`. D-7 re-signed `personBeside` assuming a caller that D-7 itself deleted; deleting it keeps D-7's aim (one seed path). `IdentityClaim` lost both users in D-4/D-6; deleting a message frees no field number |
| CON-2, MNT-2 | CON, MNT | minor | Stale delay-list doc comment above `workspaceNavCountsHold`; Landed note silent on the hold | Fix now → F-4 | verified `FakeStore.swift:182-186` |
| CON-4 | CON | minor (was nit) | `TestUpdateAndCoverReturnEffects` doc names a delete it does not cover | Fix now → F-4 | stale comment in a file this PR adds |
| MNT-4 | MNT | minor (was nit) | Effects mapped two ways in `GoStore` (`hasEffects ? … : .empty` ×11 vs direct ×33); PR adds four guarded | Fix now → F-5 | same-pass cleanup: this PR adds copies of the guarded form |
| CON-5, MNT-OOL, SEC-OOL | CON, MNT, SEC | minor | `fileSubject` moved the primary-type check after the already-member check: an association Subject with an accepted claim now gets `ErrAlreadyMember` / no-op instead of `ErrUnsupportedType` | To user | verified at base `04a2c99b`: `Save` and `applyRow` checked `IsPrimaryType` before `AcceptedEntityForSubjectTx`. D-17 (and Assumptions line 222) put the type check inside `fileSubject` after the callers' member check, which contradicts the plan's own stop condition "without changing a refusal one of them returns today". Nothing written either way; not reachable from Promote rows. Fixing needs a D-17 change (split the type check out, or change `fileSubject`'s signature) → user's call |
| MNT-3 | MNT | minor | `promote.BatchResult.Written` built only for `batch_test.go` now the handler discards it | To user | contradicts the plan's Out line / "What stays" (`BatchResult.Written` stays as the package's tested result) |
| UX-1 | UX | minor | Sidebar recount failing after a write is silent; Sources badge lost its publish fallback | Follow-up | silent-after-write predates this PR (Promote already recounted); D-18 is a Decision. Home: idea file, `WorkspaceView` observes `lastRefreshError` via `.onChange` |

Nits (logged): `UX-2 · PromoteModel.swift:41 · backDestination carries the title frozen at open; could pass sourceTitle`

To-user items do not block Fix plan 1 (no F-# depends on them); they go to the user with the report.

### Fix plan 1

**F-1 — Remove the dead `catalogCounts` dependency from the Sources model** (from CON-1, MNT-1)
Files: `macos/App/Features/Sources/SourcesModel.swift`, `SourcesListView.swift`, `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`, any other constructor call site. Skill: `add-workspace-place`.
Failing test first: none (deletion; the build proves no reader).
Do: delete `SourcesModel.catalogCounts`, its init parameter and assignment, the `SourcesListView` parameter and pass-through, the `WorkspaceDestinationHost` argument, and the `catalogCounts:` parameter of the `SourcesModelTests` factory. Leave the vocabulary models' (`SourceTypesModel`, `MetadataModel`, `PropertiesModel`) `catalogCounts` alone; they still publish (D-18).
Check: `rg -n catalogCounts macos/App/Features/Sources macos/ProvenenciaTests/SourcesModelTests.swift` prints nothing; scripts/check-pr.sh PASS.

**F-2 — Promote header reads the Source row once, in its own view** (from PRF-1)
Files: `macos/App/Features/Promote/PromoteView.swift`, `PromoteModel.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`. Skill: `add-ui-component` (feature-private snowflake).
Failing test first: none required for the view split; keep `headerFollowsSourceRow` green. If the model's title/ref/mark are reshaped (below), adjust that test to the new API, still asserting AC-2.5/AC-2.6.
Do: move `header` into `private struct PromoteHeader: View` taking the model, so its observation covers only the Source row and types, not `flow`. Read the row once per evaluation: give `PromoteModel` one accessor that returns the header's title, ref and mark from a single `session.source(id:)` lookup (e.g. a small `PromoteSourceHeader` value or a method taking the row), with the title from that same row and the carried title as fallback, not a second lookup through `LocationNames`. Keep the fallbacks D-16 names (carried title, empty ref, default mark). Leave `WorkspaceLocationLabels.sourceTitle` and the toolbar as they are.
Check: `grep -n "session.source(id:" macos/App/Features/Promote` shows one call site; `headerFollowsSourceRow` passes; scripts/check-pr.sh PASS.

**F-3 — Delete `personBeside` and the proto `IdentityClaim` message** (from CON-3, MNT-1)
Files: `api/ffi/handlers/promote_test.go`, `api/proto/engine.proto`, regenerated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: delete `personBeside` and its comment. Delete `message IdentityClaim` (no field number to reserve; no message references it) and regenerate with the repo's usual proto generation command.
Check: `rg -n 'personBeside' api` and `rg -n '\bIdentityClaim\b' api/proto/engine.proto macos/App --glob '!*.pb.*'` print nothing; `rg -n 'V1_IdentityClaim\b|IdentityClaim struct' api macos/App` prints nothing; scripts/check-pr.sh PASS.

**F-4 — Stale comments** (from CON-2, MNT-2, CON-4)
Files: `macos/App/Platform/FakeStore.swift`, `api/ffi/handlers/source_write_effects_test.go`, the plan's Landed note.
Do: delete the three delay-list doc lines above `workspaceNavCountsHold` (`FakeStore.swift:182-184`). Fix `TestUpdateAndCoverReturnEffects`'s doc comment so it names only what it covers (the delete lives in `TestDeleteSourceReturnsEffects`). In the plan's **Landed** note only (no Decision, Step or Test edits): replace differences (3), (4) and (5) with one line that they were removed in review (F-1, F-3), and add one difference: T-12 holds the first recount through `FakeStore.workspaceNavCountsHold` (`FakeHold`) instead of `workspaceNavCountsDelays`, and the Promote header is its own view (F-2).
Check: `sed -n 180,190p macos/App/Platform/FakeStore.swift` shows one doc comment; scripts/check-pr.sh PASS.

**F-5 — Map write effects one way in `GoStore`** (from MNT-4)
Files: `macos/App/Platform/GoStore.swift`, a Swift test (`macos/ProvenenciaTests/`, wherever `GoStore` mapping is tested, else the closest store-mapping test file). Skill: `add-swift-test`.
Failing test first: a test that a default (unset) `Provenencia_Engine_V1_WriteEffects()` maps to `CatalogWriteEffects.empty` — only if `mapWriteEffects` is reachable from tests; if it is `private`, decide by reading `CatalogWriteEffects.empty` and `CatalogActivity` instead.
Do: if an unset message already maps equal to `.empty`, drop the `resp.hasEffects ? … : .empty` guard at all 11 sites and call `Self.mapWriteEffects(resp.effects)` directly. If it does not (e.g. `CatalogActivity(revision: 0, at: "")` differs from `.empty`'s activity), add one `mapWriteEffects(ofResponse:)`-style helper taking `hasEffects` and the message, and move all 44 sites onto whichever single form is correct. Do not change what any call returns.
Check: `grep -c "hasEffects ? Self.mapWriteEffects" macos/App/Platform/GoStore.swift` prints 0 (or every site uses the one helper); scripts/check-pr.sh PASS.

## Round 1 — fix
Head `3b7ff678`. **Gate (orchestrator) PASS.** F-1 to F-5 done. F-3's regenerated Swift landed in a second commit (`3b7ff678`) after `794f797c`. F-5: no test (`mapWriteEffects` is private); orchestrator confirmed `CatalogWriteEffects.empty == CatalogWriteEffects()` with `CatalogActivity()` defaults 0/"", which an unset message maps to.
Minors read by the orchestrator: CON-2/MNT-2 **Resolved** (one doc comment above the hold; Landed note updated); CON-4 **Resolved**; CON-3/MNT-1b **Resolved** (`rg personBeside api`, `rg IdentityClaim engine.proto` empty); MNT-4 **Resolved** (`grep -c hasEffects GoStore.swift` = 0).

## Round 2 — review
Scope `20dde76e...3b7ff678`. Reviewers: CON (CON-1), MNT (MNT-1a), SEC+PRF (PRF-1). Not sent: UX (no Fix-now blocker or major).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | (round 1) | Resolved | round-2 CON: `rg catalogCounts` over Sources and its tests empty |
| MNT-1a | MNT | major | (round 1) | Resolved | round-2 MNT, same evidence |
| PRF-1 | PRF | major | (round 1) header re-runs on every flow change | Not resolved → F-6 | verified `PromoteView.swift:723-738`: `PromoteHeader`'s `promoting` / `countLine` read `model.flow`, so the header (and its `sourceHeader` scan) re-runs on every row decision; scans cut 3 → 1 only. Doc comment at `:681` claims otherwise. Finishing a round-1 Fix-now, not a new finding |
| MNT-5 | MNT | blocker | F-2 inlined a copy of `WorkspaceLocationLabels.sourceTitle`'s rule in `PromoteModel.sourceHeader` | Fix now → F-7 | verified: `PromoteModel.swift:114` `row?.title.trimmedNonEmpty ?? entry.sourceTitle` beside `WorkspaceLocationLabels.swift:58`. Caused by Fix plan 1's wording ("not a second lookup through `LocationNames`") |

### Fix plan 2

**F-6 — The Promote header observes only the Source row and types** (finishes PRF-1)
Files: `macos/App/Features/Promote/PromoteView.swift`.
Do: move the right-hand counts (`promoting`, `countLine` and the `VStack` that shows them) out of `PromoteHeader`. Either `PromoteView` (which already reads `flow`) computes the two strings and passes them to `PromoteHeader` as plain `let` values, or the counts become their own small private view taking the model. After the change `PromoteHeader.body` reads `model.sourceHeader` and nothing from `model.flow`. Make the doc comment at `PromoteHeader` true for the result. Layout and copy unchanged.
Check: `sed -n '/private struct PromoteHeader/,/^}/p' macos/App/Features/Promote/PromoteView.swift | grep -c 'flow'` prints 0; scripts/check-pr.sh PASS.

**F-7 — One rule for a Source's displayed title** (from MNT-5)
Files: `macos/App/Features/Workspace/WorkspaceLocationLabels.swift`, `macos/App/Features/Promote/PromoteModel.swift`.
Do: add `static func sourceTitle(live: String?, carried: String?) -> String?` to `WorkspaceLocationLabels` returning `live?.trimmedNonEmpty ?? carried?.trimmedNonEmpty`; make `sourceTitle(sourceID:carried:names:)` call it with `names.sourceTitle(sourceID)`; make `PromoteModel.sourceHeader` call `WorkspaceLocationLabels.sourceTitle(live: row?.title, carried: entry.sourceTitle)`. Keep the single `session.source(id:)` lookup.
Check: `grep -n "trimmedNonEmpty ??" macos/App/Features/Promote/PromoteModel.swift` prints nothing; `headerFollowsSourceRow` and `WorkspaceToolbarBreadcrumbTests` pass; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `33642a78`. Implementer gate PASS. F-7 done and read: **MNT-5 Resolved** (`WorkspaceLocationLabels.sourceTitle(live:carried:)` is the one rule; both callers use it). F-6 read: `PromoteHeader<Counts: View>` takes a `counts` view that `PromoteView` builds from `flow`. `PromoteHeader.body` no longer reads `flow`, but the parent rebuilds `counts` (a `VStack` of `Text`) on every flow change, and SwiftUI cannot prove that value equal, so the header body (and its `sourceHeader` scan) still re-runs per row decision. **PRF-1 not resolved** → F-6b, the reviewer's other option.

**F-6b — The counts observe `flow` themselves, in a child view**
Files: `macos/App/Features/Promote/PromoteView.swift`.
Do: replace the generic `counts` parameter with `private struct PromoteHeaderCounts: View { let model: PromoteModel }` whose body holds the `promoting`/`countLine` strings and the trailing `VStack` (moved from `promotingCounts`). `PromoteHeader` goes back to `let model: PromoteModel` only and places `PromoteHeaderCounts(model: model)` after the `Spacer`. Delete `promotingCounts`. Then a flow change invalidates only `PromoteHeaderCounts`. `PromoteView` re-creates `PromoteHeader(model:)` with the same class reference, so SwiftUI skips the header body. Doc comment says so. Layout and copy unchanged.
Check: `PromoteHeader`'s body and its stored properties read no `flow` and take no view parameter; `grep -n promotingCounts PromoteView.swift` empty; scripts/check-pr.sh PASS.
F-6b head `01e9a3dd` (`4def3e07` code, `01e9a3dd` Landed note). Orchestrator read: `PromoteHeader` holds only `let model`, its body reads `sourceHeader` and places `PromoteHeaderCounts(model:)`, which alone reads `flow`. **PRF-1 Resolved.** **Gate (orchestrator) PASS** on `01e9a3dd`; `origin/main` unmoved since base `04a2c99b`.

## Result
CI: first run red — Go `TestLargeSourceBudgets/warm` 373 ms vs 350 ms budget (`promotealign` proposal, untouched by this PR; local gate green) and macOS job never acquired a runner (capacity). Rerun of failed jobs green.

Gate PASS · CI green · open: none (blocker/major) · follow-ups: UX-1 · to user: CON-5 (`fileSubject` refusal order), MNT-3 (`BatchResult.Written`) · CON miss: none

## User decision — 2026-11-02
On the report, the user answered "Yes, fix both": CON-5 (`fileSubject` refusal order) and MNT-3 (`BatchResult.Written`) are **Fix now**. This supersedes D-17's placement of the primary-type check inside `fileSubject` (and Assumptions' "callers check membership first" for that check) and the Out line "Removing `promote.BatchResult.Written` — not planned" / "What stays". Recorded in the plan's Landed note, not in its Decisions.

### Fix plan 3 (user-approved)

**F-8 — The primary-type refusal comes before the membership refusal again** (from CON-5)
Files: `core/database/promote/promote.go`, `batch.go`, `promote_test.go`, `batch_test.go`.
Failing test first: `TestSave` (or the existing refusal table in `promote_test.go`) case "an association Subject already filed by its bridge is unsupported": file both ends of a participation through `promotetest.Save` (which files the bridge, `AppendFiling`), then `Save` the association Subject → want `ErrUnsupportedType` (today: `identityclaims.ErrAlreadyMember`). And a `batch_test.go` case: a `.handle` row naming that association Subject on its own handle → want `ErrUnsupportedType` (today: a silent no-op). Run both red first.
Do: add one unexported `requirePrimary(tx *sql.Tx, subject subjects.Subject) error` in `promote.go` (the `subjecttypes.GetByIDTx` + `vocabkey.IsPrimaryType` → `ErrUnsupportedType` block, moved out of `fileSubject`). `Save` calls it right after loading the Subject, before its already-member check; `applyRow` calls it right after its graph check, before its already-member lookup — the order both had at base `04a2c99b`. `fileSubject` drops the type check and its doc says callers check the type and membership first.
Check: both new cases pass; `go test -tags fts5 ./core/database/promote/...` green; scripts/check-pr.sh PASS.

**F-9 — `SaveBatch` returns what its caller uses** (from MNT-3)
Files: `core/database/promote/batch.go`, `batch_test.go`, `api/ffi/handlers/promote_batch.go`, any other `SaveBatch` caller (`rg -n 'SaveBatch|BatchResult|promote\.Written' core api`).
Do: delete `Written` and `BatchResult`. `SaveBatch` returns `(struct{}, []rowchange.Change, error)` so it still fits `writes.Run[T]` / `writes.Call[T]` directly (follow an existing `struct{}` write in core if one exists; else that shape). `applyRow` returns `([]rowchange.Change, []byte, error)` (changes, filed entity id or nil). Rewrite the three `batch_test.go` assertions on `res.Written` (around lines 214, 321–324, 354) to read the filed claims back (`identityclaims.AcceptedEntityForSubjectTx` or the package's existing read helper) or to assert on the returned `rowchange.Change`s, proving the same facts: two claims filed; each on the expected handle; an anchor files nothing.
Check: `rg -n 'BatchResult|promote\.Written|\bWritten\b' core/database/promote api/ffi/handlers` prints nothing (except unrelated words); go tests green; scripts/check-pr.sh PASS.

**F-10 — Landed note**
Files: the plan.
Do: in the **Landed** note only, add one difference: "In review, by the user's decision: the primary-type refusal runs before the already-member refusal again (`requirePrimary`, outside `fileSubject`), and `BatchResult` / `Written` are removed; `SaveBatch` returns only its changes." No Decision, Out or Test edits.

## Round 3 — fix (user-approved)
Head `09b9b7ed`. **Gate (orchestrator) PASS.** F-8: `requirePrimary` runs before the member check in `Save` and `applyRow`; `TestSaveRefusesAnAssociationAlreadyFiledByItsBridge` and `TestSaveBatchRefusesAnAssociationRowAlreadyFiled` ran red first (`already_member`; nil). **CON-5 Resolved.** F-9: `BatchResult` / `Written` gone, `SaveBatch` returns `(struct{}, changes, error)`; `batch_test.go` reads claims back. **MNT-3 Resolved.** F-10 Landed note (7).
Orchestrator read: `applyRow`'s second return (filed entity id) has no reader (`batch.go:68` binds `_`), in code this PR rewrote; F-9 named that shape. → F-11.

**F-11 — `applyRow` returns only its changes**
Files: `core/database/promote/batch.go`.
Do: `applyRow` returns `([]rowchange.Change, error)`; drop the entity id (bind `entity` to `_` from `fileSubject` if unused) and fix its doc comment; `SaveBatch` binds `rowChanges, err := applyRow(…)`.
Check: `grep -n "applyRow(" core/database/promote/batch.go` shows two returns in the signature; scripts/check-pr.sh PASS.
F-11 head `13eccf4f` (only `batch.go`). **Gate (orchestrator) PASS.**

## Result (after the user's decision)
Gate PASS · CI pending on `13eccf4f`+log · open: none · follow-ups: UX-1 · to user: none (CON-5, MNT-3 answered "fix both", fixed in round 3) · CON miss: none
