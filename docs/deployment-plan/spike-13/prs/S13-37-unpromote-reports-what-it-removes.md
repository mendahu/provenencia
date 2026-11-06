# S13-37 — Unpromote reports what it removes

**Status:** Planned
**Spike / slice:** Spike 13 · Slice 6 — Sources
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: two wire fields and their Swift mapping; nothing draws them until S13-17)
**Board:** —
**Depends on:** S13-08 (landed [#707](https://github.com/mendahu/provenencia/pull/707))
**Blocks:** S13-17
**Branch:** `s13-37-unpromote-reports-what-it-removes`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, and this PR adds no catalog format step ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md) (read: `ReleaseClaim` and the handle release), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`../person-detail.md`](../person-detail.md) R-38, R-39; working rule "One Unpromote, deletes through the registry"; the S13-17 plan (D-3, D-6: its consumers) and its review log (round 1 M-2, M-3, the reason for this PR); the S13-08 plan
**Planned:** 2026-11-06 against `main` @ `d5cefe5`
**Review:** [`S13-37-unpromote-reports-what-it-removes.review.md`](S13-37-unpromote-reports-what-it-removes.review.md) — round 1 Fits with changes / Not ready; every finding Fixed

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

The Unpromote confirm and the page that runs it should be told by the engine what an Unpromote takes away, not work it out in Swift. After this PR, the Unpromote preflight (`GetUnpromoteImpact`) also returns the bridge Subjects whose accepted claims the write will take back, chosen by the same rule the write runs; and every write's effects (`WriteEffects`) name the canonical entities it deleted. S13-17 then words the confirm's "Also unpromoted" list from the engine's ids and leaves the Person page for the Persons list when the write's effects name the person (user, 2026-11-05).

**Why now.** S13-17's review (round 1, M-2 and M-3) found that its draft copied the engine's bridge rule in Swift and inferred the person's deletion after the write; the user chose this preparatory PR (2026-11-05).

**What stays.** Unpromote and Subject delete do exactly what they do now; the audit rows, the recompute and the refresh are unchanged. The Evidence graph's Unpromote confirm looks and behaves as today: its list is still built from its snapshot until S13-17 moves it onto the preflight's ids. No Swift view reads either new field in this PR.

## 2. User stories

**US-1.** As the Unpromote confirm, I learn from the engine which connections an Unpromote takes back.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A promoted person Subject P with two filed bridges (a relationship R and a participation Q, both with accepted claims) and one unfiled bridge U on its graph | `GetUnpromoteImpact(P)` returns `unpromoted_bridge_subject_ids` = [R, Q] in bridge Subject id order, without U |
| AC-1.2 | Then `UnpromoteSubject(P)` | The write deletes exactly the accepted claims of R and Q (the ids AC-1.1 returned) |
| AC-1.3 | A promoted Subject with no filed bridges | The preflight returns no bridge ids; `deletes` as today |
| AC-1.4 | A Subject with no accepted claim (a stale second click) | The preflight returns no bridge ids and no deletes, as the write changes nothing |

**US-2.** As a page that ran a write, I learn from its effects which handles it deleted.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Unpromote of a person's last accepted member | `WriteEffects.deleted_handles` = [that person's entity id] |
| AC-2.2 | Unpromote that also empties a relationship handle (its last bridge member) | `deleted_handles` holds the person and the relationship handle |
| AC-2.3 | Unpromote of a member when the person keeps another | `deleted_handles` is empty |
| AC-2.4 | Subject delete of a handle's last member (S13-08) | `deleted_handles` names that handle |
| AC-2.5 | Any write that deletes no handle (a Source edit, a Promote, an Observation delete) | `deleted_handles` is empty |
| AC-2.6 | The Swift store reads an Unpromote | `CatalogWriteEffects.deletedHandles` holds the ids; `getUnpromoteImpact` returns `CatalogUnpromoteImpact{deletes, unpromotedBridgeSubjectIDs}`; the Evidence graph's confirm reads `deletes` as before |

## 3. Scope

| In | Covered by |
| --- | --- |
| One rule for "the filed bridges an Unpromote takes back", shared by the write, the preflight's deleted-handles list and the new id list | STEP-1, AC-1.1 – AC-1.4 |
| The preflight returns the bridge Subject ids | STEP-2, STEP-4, AC-1.1 – AC-1.4 |
| A write's effects name the handles it deleted, through the effects registry | STEP-3, STEP-4, AC-2.1 – AC-2.5 |
| Proto, regenerated; Swift store mapping and types; FakeStore | STEP-4, STEP-5, AC-2.6 |
| Docs | STEP-6 |

| Out | Home |
| --- | --- |
| The confirm's list from the engine's ids; leaving the Person page on `deletedHandles` | S13-17 |
| The session acting on `deletedHandles` (dropping keys, leaving pages) | S13-17 for the Person page; other pages with their redesign (R-1) |
| Deleted entities other than handles (Sources, Subjects, Citations) | not planned: no consumer |
| Any change to what Unpromote or Subject delete removes | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Unpromote | `core/database/promote/unpromote.go` `Unpromote(tx, userID, subjectID)` | Releases the Subject's accepted claim (`deleteimpact.ReleaseClaim`), then `identityclaims.UnfileBridges`; a Subject with no accepted claim returns no changes |
| Preflight | `core/database/promote/unpromote_impact.go` `UnpromoteImpact(c, subjectID) (deleteimpact.Group, error)`, `unpromoteDeletedHandleRows` | A read snapshot: the primary handle when last member (`deleteimpact.LastMemberGroup`) plus `identityclaims.UnpromoteBridgeHandlesDeleted`, deduped, sorted by ref, as a `ListedGroup` (`listedCap`) |
| Bridge rule | `core/database/identityclaims/bridgefile.go` `UnfileBridges(tx, subjectID)`, `UnpromoteBridgeHandlesDeleted(tx, subjectID)`, `bridgeHeads`, `bridgeHeadsSQL` | Both call `bridgeHeads(tx, touchesEnd, subjectID)` (ordered by bridge Subject id; the SQL LEFT JOINs the accepted claim) and skip a head with `claimID == nil`: the same filter twice; `UnpromoteBridgeHandlesDeleted` then queries each head's `entity_id`. `UnfileBridges`' doc comment sits above `UnpromoteBridgeHandlesDeleted` |
| Handle release | `core/database/deleteimpact/handle_gone.go` `releaseHandleIfGone` | The only place that emits `rowchange.Change{EntityType: "canonical_entity", Action: rowchange.ActionDelete}`; reached from `ReleaseClaim` by Unpromote, `UnfileBridges` and Subject delete |
| Effects | `core/database/effects/registry.go` `Set` (`Handles`, `EvidenceGraphs`, `Reads`, `SourceScopes`, `HandleScopes`, `Lists`, `Structure`, …), `Effect` (`Handles`, `HandleScope`, `SourceScope`, `EvidenceGraph`, `Reads[].Path`, `Structure`), `paths()`, `validate`, registry entry `"canonical_entities"`; `core/database/effects/path.go` `onField`, `onStatus`; `core/database/effects/resolve.go` `Resolve(tx, changes)` (`appendPath` per id set) | Every id set is a `Path`; guards narrow by a change's shape; `validate` checks where a path ends. No set names deletes |
| Writes | `core/writes/run.go` `Run` | Calls `effects.Resolve` once and returns the set in its `Result` |
| Wire | `api/proto/engine.proto` `WriteEffects` (fields 1, 5 – 10, 13 – 15: 14 is `evidence_graphs`, 15 `reads`; reserved 2, 3, 4, 11, 12), `GetUnpromoteImpactResponse{deletes = 1}`; `api/ffi/handlers/write_effects.go` `writeEffectsProto`; `api/ffi/handlers/promote.go` `GetUnpromoteImpact`, `UnpromoteSubject` | `writeEffectsProto` maps every `Set` field the app reads |
| Client doc | `docs/macos-client-patterns.md` (the `apply(effects)` paragraph) | Lists the `WriteEffects` fields the session applies |
| Swift | `macos/App/Platform/GoStore.swift` `mapWriteEffects`, `getUnpromoteImpact -> CatalogDeleteImpactGroup?`; `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogWriteEffects`; `macos/App/Platform/GenealogyStore.swift` (protocol); `macos/App/Platform/FakeStore.swift` `unpromoteImpactBySubject`, `getUnpromoteImpact`, `nextWriteEffects`; `macos/ProvenenciaTests/OnboardingModelTests.swift` (`ThrowingStore`) | One caller of `getUnpromoteImpact`: `EvidenceGraphModel.beginUnpromote` |
| Tests | `core/database/promote/unpromote_impact_test.go` `TestUnpromoteImpactListsBridgeHandles`, `TestUnpromoteImpactEmptyWhenHandleStays`; `core/database/promote/unpromote_test.go` `TestUnpromote`; `api/ffi/handlers/promote_test.go` `TestUnpromoteSubject`, `TestGetUnpromoteImpact`; `api/ffi/handlers/subjects_test.go` `TestDeleteSubject`; `api/ffi/handlers/write_effects_test.go` `TestWriteEffectsProto`; `core/database/effects/registry_completeness_test.go`; `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` Unpromote cases | — |

**Precedent.** S13-08 ([#707](https://github.com/mendahu/provenencia/pull/707)) added `GetUnpromoteImpact` and its handler tests; copy its handler shape and its test helpers (`promotableSubject`, `promoteInDir`). `Effect.Lists` with `ListsOnField` is the precedent for a registry flag that `Resolve` reads per change.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One read of the filed bridges: `identityclaims.FiledBridgesOf(tx *sql.Tx, subjectID []byte) ([]FiledBridge, error)` *(new)*, `FiledBridge{SubjectID, ClaimID, EntityID []byte}` *(new; `ClaimID` and `EntityID` non-nil copies)*: `bridgeHeads(tx, touchesEnd, subjectID)` keeping the heads with an accepted claim, in bridge Subject id order, with `bridgeHeadsSQL` also selecting `ic.entity_id` from the claim it already LEFT JOINs. `UnpromoteBridgeHandlesDeleted(tx, filed []FiledBridge)` *(signature change)* takes that list and drops its per-head `SELECT entity_id` query; `UnfileBridges` calls `FiledBridgesOf` itself. STEP-1 moves `UnfileBridges`' detached doc comment back above it | One read feeds the write, the deletes forecast and the new list, so they cannot disagree; one query fewer per bridge (round 1 M-3) | A third copy of the filter; two reads in the preflight |
| D-2 | The preflight returns both: `promote.UnpromoteImpact` returns `UnpromoteReport{Deletes deleteimpact.Group, UnpromotedBridgeSubjects [][]byte}` *(new type; the function keeps its name)*, from one read transaction. After `loadPrimary`, it checks `identityclaims.AcceptedEntityForSubjectTx` as `Unpromote` does: `sql.ErrNoRows` → an empty report (the write changes nothing, AC-1.4). Otherwise it calls `FiledBridgesOf` once, passes the list to `UnpromoteBridgeHandlesDeleted`, and returns its Subject ids | The same guard as the write, in the same order; one snapshot | A second RPC; a report that lists bridges the write would not touch (round 1 X-3); `promote.Impact` (reads as a Promote's impact beside `deleteimpact.Impact`, round 1 M-5) |
| D-3 | Wire: `GetUnpromoteImpactResponse` gains `repeated string unpromoted_bridge_subject_ids = 2` *(new)*; the handler fills it with `idStrings` | Field 2 is free | — |
| D-4 | Deleted handles are a registry path: `onDelete(inner Path) Path` *(new guard in `core/database/effects/path.go`, beside `onStatus`)* runs `inner` only when `w.ch.Action == rowchange.ActionDelete`; `Effect.DeletedHandles Path` *(new)*; the `"canonical_entities"` entry sets `DeletedHandles: onDelete(self())`; `Resolve` fills `Set.DeletedHandles [][]byte` *(new)* with one more `appendPath` (deduped, in change order); `paths()` includes it, and `validate` checks that a non-zero `DeletedHandles` path ends at `canonical_entities`, as it checks a read's `KeyTable()` | One grammar for "which ids a change yields": every id set is a `Path` (round 1 M-2) | A `ReportsDelete bool` flag with an Entity-name rule; collecting deletes in `releaseHandleIfGone` |
| D-5 | Wire: `WriteEffects` gains `repeated string deleted_handles = 16` *(new; 14 and 15 are `evidence_graphs` and `reads`)*, mapped in `writeEffectsProto` with `idStrings(set.DeletedHandles)`. A merge updates `merged_into_id` and is not a delete, so a merged handle is not listed | The next free number | — |
| D-6 | Swift: `CatalogWriteEffects.deletedHandles: [String] = []` *(new)*, mapped in `GoStore.mapWriteEffects`. `CatalogUnpromoteImpact{deletes: CatalogDeleteImpactGroup?, unpromotedBridgeSubjectIDs: [String]}` *(new, `Sendable, Equatable`)* in `GenealogyStore.swift`; `getUnpromoteImpact(projectDir:subjectID:) -> CatalogUnpromoteImpact` (non-optional; `deletes` nil when the response has none, as today's `nil`). `EvidenceGraphModel.beginUnpromote` reads `impact.deletes` with today's `total > 0` rule. `session.apply(effects)` ignores `deletedHandles` in this PR | The store keeps one call per RPC; S13-17 reads both new fields | A second store method |
| D-7 | FakeStore stays scripted (`swift-tests.mdc`), one dict for the whole return: `unpromoteImpactBySubject: [String: CatalogUnpromoteImpact]` *(retyped)*; `getUnpromoteImpact` returns `unpromoteImpactBySubject[id] ?? CatalogUnpromoteImpact(deletes: nil, unpromotedBridgeSubjectIDs: [])`. Its one test use (`EvidenceGraphModelTests` `unpromoteLastRecordNamesTheDeletedHandle`) wraps its group in `CatalogUnpromoteImpact`. `unpromoteSubject` keeps returning `nextWriteEffects`. `ThrowingStore` throws, like its neighbours | `deleteImpactByID` scripts the whole report the same way (round 1 M-4) | Two dicts keyed by the same Subject id |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `identityclaims.FiledBridgesOf`, `FiledBridge` *(new)* | func, type | `bridgeHeads` + the two copies of its filter, the per-head `entity_id` query | Replaces the copies and the query; `bridgeHeads` stays private |
| `promote.UnpromoteReport` *(new)* | type | `deleteimpact.Group` (today's return), `deleteimpact.Report` | Holds the group and the ids |
| `onDelete`, `Effect.DeletedHandles`, `Set.DeletedHandles` *(new)* | guard, fields | `onStatus`, `Effect.HandleScope` / `Set.HandleScopes` | One more id set in the registry's own grammar; `Handles` is the recompute set, not the deleted ones |
| `GetUnpromoteImpactResponse.unpromoted_bridge_subject_ids`, `WriteEffects.deleted_handles` *(new)* | proto fields | — | New facts |
| `CatalogUnpromoteImpact`, `CatalogWriteEffects.deletedHandles` *(new)* | Swift type, field | `CatalogDeleteImpactGroup` | Mirror the wire; FakeStore's dict is retyped, not doubled |
| `effects` test `deleted_handles_test.go` *(new)* | test file | `reads_resolve_test.go` | One file per resolved field |

**Builds on.** The effects registry and `Resolve`; `deleteimpact.ReleaseClaim` (the one handle release); S13-08's preflight and handler.
**Deletes.** The duplicated `claimID == nil` filters in `UnfileBridges` and `UnpromoteBridgeHandlesDeleted`; `UnpromoteBridgeHandlesDeleted`'s per-head `entity_id` query.
**Next adds.** S13-17 words the list from `unpromotedBridgeSubjectIDs` and leaves on `deletedHandles`; a later "the page's entity was deleted" rule for Event and Place reads `deletedHandles`; a researcher-facing handle delete reports through the path with no new code; a later deleted-Sources set is one more `onDelete` path (and replaces the client-asserted `.deletedSource(id:)` mutation then: logged in `docs/ideas/structural-debt.md`).

### Reach

None — no new navigation path or view state; the two fields have no Swift reader beyond the store until S13-17.

### Scale

Map Scale (slice 0: a Source of 100 people on one Promote page; the 41-Subject obituary). The preflight runs `bridgeHeads` once, as today, and drops one `entity_id` query per filed bridge. `Resolve` runs one more path per change, which returns at once unless the change deletes a handle; a Promote Done of 100 rows with their bridges (hundreds of changes) gains one comparison each and no query. `deleted_handles` is empty for every write but an Unpromote or a Subject delete.

### Edges

- **Assumptions.** `releaseHandleIfGone` is the only emitter of a handle delete (`rg -U 'EntityType:\s+"canonical_entity",\n[^\n]*\n\s*Action:\s+rowchange.ActionDelete' core` finds one); a later emitter is covered by the registry flag with no change. The preflight and the write read the same rule (D-1); T-2 compares the preflight's ids with the write's released claims.
- **Neighbors.** A bridge whose other end is unpromoted has no accepted claim (it waits), so it is not listed and not touched (AC-1.1's U). A merge updates a handle and is not listed (D-5).
- **Removed inputs.** A bridge deleted between the preflight and the write: the write's own set wins; the confirm's list was a forecast (S13-17 skips an id it cannot word).
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** The bridge list is the Subject's filed bridges on one graph, already read by the preflight today; `deleted_handles` holds at most the handles the write released.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/identityclaims` | `func FiledBridgesOf(tx *sql.Tx, subjectID []byte) ([]FiledBridge, error)` *(new)*; `UnpromoteBridgeHandlesDeleted(tx *sql.Tx, filed []FiledBridge)`; `UnfileBridges` calls `FiledBridgesOf` |
| `core/database/promote` | `type UnpromoteReport struct { Deletes deleteimpact.Group; UnpromotedBridgeSubjects [][]byte }` *(new)*; `UnpromoteImpact(c, subjectID) (UnpromoteReport, error)` |
| `core/database/effects` | `func onDelete(inner Path) Path`, `Effect.DeletedHandles Path`, `Set.DeletedHandles [][]byte` *(new)*; `Resolve` fills it; `validate` checks its end; the README's `Set` paragraph names it |

### FFI

`api/proto/engine.proto`: `GetUnpromoteImpactResponse.unpromoted_bridge_subject_ids = 2`, `WriteEffects.deleted_handles = 16` *(new)*. Handlers: `api/ffi/handlers/promote.go` `GetUnpromoteImpact`; `api/ffi/handlers/write_effects.go` `writeEffectsProto`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

None — no view changes. Store types and mapping only (D-6, D-7).

### Strings

None — no copy.

## 7. Steps

**STEP-1 — One rule for the filed bridges**
Files: `core/database/identityclaims/bridgefile.go`.
Do: D-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/identityclaims/... ./core/database/promote/...` passes (the existing Unpromote tests unchanged).

**STEP-2 — The preflight returns the bridges**
Files: `core/database/promote/unpromote_impact.go`, `core/database/promote/unpromote_impact_test.go`, `api/ffi/handlers/promote.go` (read `.Deletes`, so the tree builds). Skill: `add-catalog-delete` (read).
Do: D-2.
Check: `go build ./...` succeeds; T-1, T-2 pass; `TestUnpromoteImpactListsBridgeHandles` and `TestUnpromoteImpactEmptyWhenHandleStays` pass on `impact.Deletes`.

**STEP-3 — Effects name deleted handles**
Files: `core/database/effects/path.go`, `core/database/effects/registry.go`, `core/database/effects/resolve.go`, `core/database/effects/deleted_handles_test.go` *(new)*, `core/database/effects/README.md`, `core/database/promote/unpromote_test.go`.
Do: D-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/... ./core/writes/... ./core/database/promote/...` passes (T-3, T-4, T-6).

**STEP-4 — The wire**
Files: `api/proto/engine.proto`, `api/ffi/handlers/promote.go`, `api/ffi/handlers/write_effects.go`, `api/ffi/handlers/promote_test.go`, `api/ffi/handlers/subjects_test.go`, `api/ffi/handlers/write_effects_test.go`, generated files. Skill: `add-ffi-handler`, `use-catalog-session`.
Do: D-3, D-5; `./scripts/generate-proto.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes (T-5, T-7, T-8, T-11). `./scripts/generate-proto.sh` needs `protoc`, `protoc-gen-go` and `protoc-gen-swift` (a Mac with the repo's toolchain); on a host without them, commit the `.proto` change and leave regeneration to that host before STEP-5.

**STEP-5 — Swift store**
Files: `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skill: `add-swift-test`.
Do: D-6, D-7.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes (T-9). On a host without Xcode, STEP-4's proto generation and this step's checks are run by `scripts/check-pr.sh` on a Mac, as S13-01 did.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `core/database/effects/README.md` (STEP-3), `docs/macos-client-patterns.md` (add `deleted_handles` to the applied fields: "`apply` ignores it until S13-17"), `docs/catalog-deletes.md` if it lists what a delete reports, `docs/ideas/structural-debt.md` (one line: the client-asserted `.deletedSource(id:)` and the engine-reported `deleted_handles` are two ways to learn a delete; the Event and Place "entity was deleted" rule picks the engine's).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-37-unpromote-reports-what-it-removes.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/promote/unpromote_impact_test.go` `TestFiledBridgesOf` *(new)*, table-driven, on `newBridgeWorld` / `newHousehold` | P with filed R then Q (created in that order, so R's UUIDv7 sorts first) and unfiled U → [R, Q] with their claim and entity ids; no bridges → empty; a bridge Subject's own id → empty (`touchesEnd` matches bridges with the id as a value end) | AC-1.1, AC-1.3, D-1 |
| T-2 | Go | same · `TestUnpromoteImpactListsUnpromotedBridges` *(new)* | P as T-1 → `UnpromotedBridgeSubjects` [R, Q]; then `Unpromote` → its changes delete exactly R's and Q's accepted claims (and P's); a Subject with no accepted claim but filed bridges on its graph → empty report | AC-1.1, AC-1.2, AC-1.4, D-1, D-2 |
| T-3 | Go | `core/database/effects/deleted_handles_test.go` `TestResolveReportsDeletedHandles` *(new)*, table-driven | changes [canonical_entity delete h1, identity_claim delete, canonical_entity update h2] → `DeletedHandles` [h1]; two deletes h1, h3 → [h1, h3]; no handle delete → empty | AC-2.1, AC-2.5, D-4 |
| T-4 | Go | same · `TestDeletedHandlesPathEndsAtHandles` *(new)* | `validate` of an entry whose `DeletedHandles` path ends at `sources` → error; the shipped registry → only `"canonical_entities"` sets it | D-4 |
| T-5 | Go | `api/ffi/handlers/promote_test.go` `TestGetUnpromoteImpact` (extended) | A promoted Subject with no bridges (`promotableSubject`, `promoteInDir`) → `unpromoted_bridge_subject_ids` empty and `deletes` as today; the bridge contents are T-2's (the handler only maps) | AC-1.3, D-3 |
| T-6 | Go | `core/database/promote/unpromote_test.go` `TestUnpromoteReportsDeletedHandles` *(new)*: `writes.Run` over `Unpromote` on `newBridgeWorld` | Last member → `Result.Effects.DeletedHandles` [person id]; a household whose spouse relationship's last bridge member goes with it → [person, relationship handle]; the person keeps a member → empty | AC-2.1, AC-2.2, AC-2.3, D-4 |
| T-7 | Go | `api/ffi/handlers/subjects_test.go` `TestDeleteSubjectReportsDeletedHandle` *(new case, beside `TestDeleteSubject`)* | A Subject promoted with `promoteInDir` as its handle's only member, deleted → `effects.deleted_handles` [that handle] | AC-2.4 |
| T-8 | Go | `api/ffi/handlers/write_effects_test.go` `TestWriteEffectsProto` (extended) | A `Set` with `DeletedHandles` [h1] → proto `deleted_handles` (field 16) ["<h1 uuid>"]; empty → nil | AC-2.5, D-5 |
| T-9 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` Unpromote cases | `unpromoteLastRecordNamesTheDeletedHandle` scripts `CatalogUnpromoteImpact(deletes: group, unpromotedBridgeSubjectIDs: [])`; it and `unpromotePreflightFailureStaysInTheConfirm` pass with today's behavior | AC-2.6, D-6, D-7 |
| T-11 | Go | `api/ffi/handlers/promote_test.go` `TestUnpromoteSubject` (extended) | Last member → `effects.deleted_handles` [that handle] through the handler | AC-2.1, D-5 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/effects` (the registry's validation), `core/database/promote` (the shared bridge read), `EvidenceGraphModelTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — none: no brief
- [ ] Domain doc updated if a rule in docs/*.md changed: `core/database/effects/README.md` (the `Set` paragraph); `docs/catalog-deletes.md` if it lists what a delete reports
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open the dogfood project. On an Evidence graph, Unpromote a person's last member: the confirm looks and reads as before, and the person is deleted (AC-2.6). On a graph, Unpromote a member whose person keeps another: as before.
2. The reviewer runs `CGO_ENABLED=1 go test -tags fts5 ./core/database/identityclaims/... ./core/database/promote/... ./core/database/effects/... ./api/ffi/handlers/ -run 'FiledBridgesOf|UnpromoteImpact|DeletedHandles|Unpromote|DeleteSubject|WriteEffectsProto'` and sees them pass.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The path reports a change the app did not mean (a merge) | A merge is an update (D-5); T-3's update case |
| The preflight and the write drift | One function (D-1); T-2 compares them |

**Stop and report if:**

- A second emitter of a `canonical_entity` delete exists at the base (the `rg` in Edges finds more than `handle_gone.go`).
- `UnfileBridges` and `UnpromoteBridgeHandlesDeleted` cannot share `FiledBridgesOf` without changing which claims `UnfileBridges` releases (T-2 or `TestUnpromote` fails).
- `effects.Resolve`'s ghost walk drops a deleted handle's change before the flag is read.
- `core/database/promote`'s tests cannot import `core/writes` (an import cycle) for T-6: move T-6 to `core/writes` with the same fixture.

## 12. Open questions

None.

## 13. PR

**Title:** S13-37 — Unpromote reports what it removes
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-37-unpromote-reports-what-it-removes.md`.
