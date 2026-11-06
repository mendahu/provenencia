# Ship log — S13-37 Unpromote reports what it removes

**Plan:** [`S13-37-unpromote-reports-what-it-removes.md`](S13-37-unpromote-reports-what-it-removes.md) · **Branch:** `s13-37-unpromote-reports-what-it-removes` · **PR:** #725
**State:** ready

## Round 0 — implement
Head `6464aeb7`. Gate PASS (run by the orchestrator on the pushed head, base `91e58c90`). Report: T-4 lives in its own file, `deleted_handles_path_test.go` (package `effects`), because T-3's file is `effects_test`; both proto fields landed in the first commit; `newBridgeFixture` is new test scaffolding; the Current state rows held after S14-33 moved the base; `docs/catalog-deletes.md` unchanged (it does not list what a delete reports); manual check not run.

## Round 1 — review
Reviewers: CON, MNT, UX (`macos/App/**`), SEC+PRF (`api/ffi/**`, `core/database/**`). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, MNT-1, UX-1 | CON, MNT, UX | blocker | The deletes forecast misses a handle whose members are all bridges this Unpromote takes back | Fix now → F-1 | verified: `UnpromoteBridgeHandlesDeleted` excludes one bridge per call (`bridgefile.go:128`), the write releases all; CON and MNT probes gave forecast 1, write deleted 2. Predates the PR, but in a function it rewrote and under D-1's "cannot disagree" |
| CON-3, MNT-3, UX-2 | CON, MNT, UX | major | The forecast swallows a `HandleGoesWithout` error | Fix now → F-1 | verified: `bridgefile.go:129`; twin `lastMemberRows` returns it |
| MNT-2 | MNT | blocker | `DeletedHandles` is a per-kind field through nine sites, and this PR's docs name the next kind | To user | verified: the sites are real and `Set.Reads` is the keyed precedent; the shape is D-4 and D-5's, so it contradicts a Decision |
| CON-2 | CON | major | AC-1.4's guard has no test | Fix now → F-2 | carries a test; the fix round runs it red with the guard removed |
| MNT-4 | MNT | major | The preflight re-derives the write's guard | Fix now → F-3 | verified: `unpromote.go:31-41` and `unpromote_impact.go:41-48`; same-pass cleanup, this PR adds the copy |
| MNT-5 | MNT | minor | The `GetUnpromoteImpactResponse` comment no longer describes the message | Fix now → F-4 | verified: minor in a message this PR changes |
| CON-4 | CON | minor | The docs do not record that §10 step 1 was not run | Fix now → F-5 | verified: the user runs the manual check; the Landed note and `completed.md` say so meanwhile |

SEC+PRF: no findings in either lens.

Nits (logged): `MNT-6 · core/database/effects/deleted_handles_path_test.go · T-4 could sit in registry_completeness_test.go` · `MNT-7 · core/database/effects/deleted_handles_test.go:64-71 · could use sameIDs` · `MNT-8 · core/database/promote/unpromote_test.go:212 · near-copy of w.unpromote` · `CON-5 · core/database/promote/unpromote_test.go:212 · closure calls the outer t.Fatal from subtests` · `CON-6 · three new tests are not table-driven`

MNT-2 answered by the user (2026-11-06): keyed list, superseding D-4 and D-5's per-kind field → F-6.

### Fix plan 1

**F-1 — The forecast decides a handle from every filed bridge on it** (from CON-1, MNT-1, UX-1, CON-3, MNT-3, UX-2)
Files: `core/database/identityclaims/bridgefile.go`, `core/database/deleteimpact/handle_gone.go`, `core/database/promote/unpromote_impact_test.go`. Skill: `add-catalog-delete` (read).
Failing test first: `TestUnpromoteImpactListsSharedBridgeHandle` in `unpromote_impact_test.go`: James with spouse bridges to John and Johnny, Johnny joined to John's handle, James promoted, on `newBridgeWorld`. `report.Deletes.Total` is 2, and the listed ids equal the handles the same Unpromote's `writes.Run` result reports as deleted (the person and the shared relationship handle).
Do: `handleGoesWithout` and `HandleGoesWithout` take the Subjects leaving, `leaving [][]byte`, counting accepted claims with `subject_id NOT IN (...)`. Keep one function: `lastMemberRows` and `releaseHandleIfGone` pass their one id. The `Impact` check stays. `UnpromoteBridgeHandlesDeleted` groups `filed` by `EntityID`, in first-seen order, and asks once per handle with every filed bridge on it, so `seen` goes. Split `if err != nil || !goes`: return the error, `continue` on `!goes`. Update the function's doc comment.
Check: the test passes; the existing Unpromote tests pass; scripts/check-pr.sh PASS.

**F-2 — T-2's no-claim case reaches the guard** (from CON-2)
Files: `core/database/promote/unpromote_impact_test.go`.
Failing test first: replace the no-accepted-claim subtest's body so P's claim alone is released through `deleteimpact.ReleaseClaim` inside `writes.Call`, leaving R and Q filed; `UnpromoteImpact` then returns no bridges and no deletes. Run it red once with the `AcceptedEntityForSubjectTx` block removed, then restore the block.
Do: the test only.
Check: the test passes on the restored code; report the red output. scripts/check-pr.sh PASS.

**F-3 — One guard for the write and the preflight** (from MNT-4)
Files: `core/database/promote/unpromote.go`, `core/database/promote/unpromote_impact.go`.
Failing test first: none; a refactor covered by `TestUnpromote` and F-2's test.
Do: one private helper in `promote`, called by both, holding `loadPrimary`, `AcceptedEntityForSubjectTx` and `sql.ErrNoRows` as "nothing to do". Delete both branches.
Check: `grep -n "AcceptedEntityForSubjectTx" core/database/promote/*.go | grep -v _test` prints one line, or only lines outside these two functions; scripts/check-pr.sh PASS.

**F-4 — The response comment describes the response** (from MNT-5)
Files: `api/proto/engine.proto`, generated files.
Failing test first: none; a comment.
Do: move the sentence onto `deletes`, saying every handle the Unpromote deletes; regenerate.
Check: scripts/check-pr.sh PASS.

**F-5 — The docs say the app check is outstanding** (from CON-4)
Files: the plan's Landed note, `docs/deployment-plan/spike-13/completed.md`.
Failing test first: none; docs.
Do: add "§10 step 1 not yet run in the app", as S13-16's entry words it.
Check: `check-plan-docs.py` reports no errors.

**F-6 — Effects report deleted ids as one keyed list** (from MNT-2; the user chose this, superseding D-4 and D-5)
Files: `core/database/effects/registry.go`, `resolve.go`, `README.md`, `deleted_handles_test.go`, `deleted_handles_path_test.go`, `core/database/promote/unpromote_test.go`, `api/proto/engine.proto`, generated files, `api/ffi/handlers/write_effects.go` and its tests, `promote_test.go`, `subjects_test.go`, `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/App/Platform/GoStore.swift`, `docs/macos-client-patterns.md`, `docs/ideas/structural-debt.md`, `docs/deployment-plan/spike-13/deployment-plan.md`, the plan's Landed note.
Failing test first: T-3 asserts `Set.Deleted` is `[{Entity: "canonical_entity", IDs: [h1]}]`, and T-4 refuses an entry whose `Deleted` path ends at a table other than its own.
Do: `Effect.DeletedHandles` becomes `Effect.Deleted Path`, still `onDelete(self())` on `"canonical_entities"`. `validate` checks that a non-zero `Deleted` ends at the entry's own table, with no hardcoded table. `Set.DeletedHandles` becomes `Set.Deleted []DeletedIDs` with `DeletedIDs{Entity string, IDs [][]byte}`, where `Entity` is the entry's `Effect.Entity`; `Resolve` collects by entity as it collects `reads` and reports in a stable order, ids deduped in change order. Wire: `WriteEffects.deleted_handles = 16` becomes `repeated WriteEffectsDeleted deleted = 16`, with `message WriteEffectsDeleted { string entity = 1; repeated string ids = 2; }`, mapped by a `deletedProtos` beside `readProtos`. Swift: `CatalogWriteEffects.deleted: [CatalogWriteEffectsDeleted] = []`, shaped like `CatalogWriteEffectsRead`, plus a computed `deletedHandles: [String]` reading the `canonical_entity` entry, so S13-17's D-6 reads as planned; `mapWriteEffects` maps it. Update T-3, T-4, T-6, T-7, T-8 and T-11. Update the README's `Set` paragraph, the map row, `macos-client-patterns.md`, and the structural-debt row (a later deleted-Sources set is one registry line). The Landed note records that D-4 and D-5 are superseded, by the user, in round 1.
Check: `rg -n 'deleted_handles|DeletedHandles' core api` prints only `onDelete` comments, test names, or nothing else that names a field; `session.apply(effects)` still ignores it; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `2c403323`. Gate PASS (run by the orchestrator on the pushed head). F-1 to F-6 done. F-1 ran red first (forecast 1, write deleted 2); F-2 ran red with the guard removed; F-6's tests were written with the code, as the new names would not compile before it. Deviations: F-4 and F-6 share `297ea33d`; `validate` also refuses a `Deleted` path on an entry with no `Entity`; `Set.Deleted` is ordered by each entity's first appearance.

## Round 2 — review
Reviewers: CON, MNT, UX (each had a Fix-now blocker or major). Not sent: SEC+PRF (no findings in round 1). Scope: `6464aeb7...2c403323`.

Round-1 status: CON-1, MNT-1, UX-1, CON-2, CON-3, MNT-2, MNT-3, UX-2, MNT-4 Resolved (per the reviewers; CON's seeded probe matched forecast and write over 447 Unpromotes). MNT-5 Resolved (read at `engine.proto:1121`). CON-4 Resolved in the Landed note.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-9 | MNT | blocker | `WriteEffects.deleted`'s entity key reaches Swift as a raw string with a silent fallback and no parity row | Fix now → F-7 | verified: `CatalogMutation.swift` compares the literal `"canonical_entity"`, and `swift_vocabulary_test.go` has a "catalog reads" row but none for this. F-6 said "shaped like `CatalogWriteEffectsRead`", whose key is a typed enum |
| MNT-10 | MNT | blocker | `handleGoesWithout` hand-rolls an `IN` list beside `database.SQLInPlaceholders` | Fix now → F-8 | verified: `handle_gone.go:150-160`, `core/database/sqlin.go:28,36`; `leaving` is never empty |

CON and UX: clean. No new blockers from either.

### Fix plan 2

**F-7 — The deleted key is typed vocabulary with a parity row** (from MNT-9)
Files: `core/database/effects/registry.go`, `api/ffi/swift_vocabulary_test.go`, `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/App/Platform/GoStore.swift`, `macos/ProvenenciaTests/` (a new case beside the existing `CatalogWriteEffects` tests, or a new file if none). Skill: `add-swift-test`.
Failing test first: a "deleted entities" row in `swift_vocabulary_test.go`, beside "catalog reads": `enum: "CatalogDeletedEntity"`, `path: swiftCatalogMut`, `want` from `effects.DeletedEntities()`. Fails until the enum exists.
Do: Go: `effects.DeletedEntities() []string`, every registry entry with a non-zero `Deleted`, sorted. Swift: `enum CatalogDeletedEntity: String, Sendable, CaseIterable { case handle = "canonical_entity" }` beside `CatalogRead`; `CatalogWriteEffectsDeleted.entity: CatalogDeletedEntity`; `mapWriteEffects` decodes with `compactMap` as `reads` does (an unknown key is dropped); `deletedHandles` compares `.handle`. Swift test: `CatalogWriteEffects(deleted: [..handle..]).deletedHandles` returns the ids; with no entry it is empty.
Check: the vocabulary test passes; `grep -n '"canonical_entity"' macos/App/Features/Workspace/Session/CatalogMutation.swift` prints only the enum case; scripts/check-pr.sh PASS.

**F-8 — The `NOT IN` list uses the shared helpers** (from MNT-10)
Files: `core/database/deleteimpact/handle_gone.go`.
Failing test first: none; a refactor covered by `TestUnpromoteImpactListsSharedBridgeHandle`.
Do: build the query with `database.SQLInPlaceholders(len(leaving))` and the args with `database.BlobArgs(leaving)`. Drop `notIn` and its branch.
Check: `grep -n 'notIn' core/database/deleteimpact/handle_gone.go` prints nothing; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `371d7e52`. F-7 and F-8 done, in one commit. F-7's vocabulary row was written with the code, as it would not compile before `DeletedEntities`; the Swift tests are in `CatalogQueryRegistryTests.swift`. Read by the orchestrator: both match the fix plan. Final gate result below.

## Result
Gate PASS · CI green on `f266cfbb` (Go tests, macOS tests) · open: none · follow-ups: MNT-10's private copies of the `IN` placeholder helper (`core/database/graphcache/sql.go:13`, `core/search/fts.go:292`), untouched by this PR, home `docs/ideas/structural-debt.md` (user, 2026-11-06) · to user: MNT-2, answered (keyed list)
