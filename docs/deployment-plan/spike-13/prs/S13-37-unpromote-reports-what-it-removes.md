# S13-37 — Unpromote reports what it removes

**Status:** Draft
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
**Review:** [`S13-37-unpromote-reports-what-it-removes.review.md`](S13-37-unpromote-reports-what-it-removes.review.md) — round 1 in the log

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
| Bridge rule | `core/database/identityclaims/bridgefile.go` `UnfileBridges(tx, subjectID)` and `UnpromoteBridgeHandlesDeleted(tx, subjectID)` | Both call `bridgeHeads(tx, touchesEnd, subjectID)` (ordered by bridge Subject id) and skip a head with `claimID == nil`: the same filter written twice |
| Handle release | `core/database/deleteimpact/handle_gone.go` `releaseHandleIfGone` | The only place that emits `rowchange.Change{EntityType: "canonical_entity", Action: rowchange.ActionDelete}`; reached from `ReleaseClaim` by Unpromote, `UnfileBridges` and Subject delete |
| Effects | `core/database/effects/registry.go` `Set`, `Effect`, registry entry `"canonical_entities"`; `core/database/effects/resolve.go` `Resolve(tx, changes)` | `Resolve` walks each change's registry entry into the set's fields; no field names deletes |
| Writes | `core/writes/run.go` `Run` | Calls `effects.Resolve` once and returns the set in its `Result` |
| Wire | `api/proto/engine.proto` `WriteEffects` (fields 1, 5 – 10, 13; reserved 2, 3, 4, 11, 12), `GetUnpromoteImpactResponse{deletes = 1}`; `api/ffi/handlers/write_effects.go` `writeEffectsProto`; `api/ffi/handlers/promote.go` `GetUnpromoteImpact`, `UnpromoteSubject` | `writeEffectsProto` maps every `Set` field the app reads |
| Swift | `macos/App/Platform/GoStore.swift` `mapWriteEffects`, `getUnpromoteImpact -> CatalogDeleteImpactGroup?`; `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogWriteEffects`; `macos/App/Platform/GenealogyStore.swift` (protocol); `macos/App/Platform/FakeStore.swift` `unpromoteImpactBySubject`, `getUnpromoteImpact`, `nextWriteEffects`; `macos/ProvenenciaTests/OnboardingModelTests.swift` (`ThrowingStore`) | One caller of `getUnpromoteImpact`: `EvidenceGraphModel.beginUnpromote` |
| Tests | `core/database/promote/unpromote_impact_test.go` `TestUnpromoteImpactListsBridgeHandles`, `TestUnpromoteImpactEmptyWhenHandleStays`; `core/database/promote/unpromote_test.go` `TestUnpromote`; `api/ffi/handlers/promote_test.go` `TestUnpromoteSubject`, `TestGetUnpromoteImpact`; `api/ffi/handlers/subjects_test.go` `TestDeleteSubject`; `api/ffi/handlers/write_effects_test.go` `TestWriteEffectsProto`; `core/database/effects/registry_completeness_test.go`; `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` Unpromote cases | — |

**Precedent.** S13-08 ([#707](https://github.com/mendahu/provenencia/pull/707)) added `GetUnpromoteImpact` and its handler tests; copy its handler shape and its test helpers (`promotableSubject`, `promoteInDir`). `Effect.Lists` with `ListsOnField` is the precedent for a registry flag that `Resolve` reads per change.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One rule: `identityclaims.FiledBridgesOf(tx *sql.Tx, subjectID []byte) ([]FiledBridge, error)` *(new)*, `FiledBridge{SubjectID, ClaimID []byte}` *(new)*: `bridgeHeads(tx, touchesEnd, subjectID)` keeping heads with a claim, in bridge Subject id order. `UnfileBridges` and `UnpromoteBridgeHandlesDeleted` call it instead of filtering themselves | The write, the deletes forecast and the new list read one filter, so they cannot disagree | A third copy of the filter in the preflight |
| D-2 | The preflight returns both: `promote.UnpromoteImpact` returns `Impact{Deletes deleteimpact.Group, UnpromotedBridgeSubjects [][]byte}` *(new type; the function keeps its name)*, from one read transaction; `UnpromotedBridgeSubjects` are `FiledBridgesOf`'s Subject ids. A Subject with no accepted claim returns an empty `Impact` (AC-1.4) | One snapshot for the confirm; the write also does nothing in that case | A second RPC for the bridges |
| D-3 | Wire: `GetUnpromoteImpactResponse` gains `repeated string unpromoted_bridge_subject_ids = 2` *(new)*; the handler fills it with `idStrings` | Field 2 is free | — |
| D-4 | Deleted handles through the registry: `Effect.ReportsDelete bool` *(new)*, true only on the `"canonical_entities"` entry; `Resolve` appends `ch.EntityID` to `Set.DeletedHandles [][]byte` *(new)* for a change whose entry has `ReportsDelete` and whose `Action == rowchange.ActionDelete`, once per id in change order. The registry's `validate` refuses `ReportsDelete` on an entry whose `Entity` is not `canonical_entity` (the field's name says handles) | Every write that deletes a handle reports it, with no per-write code (Subject delete gets it free); the registry already decides what each table's change means | Collecting deletes in `releaseHandleIfGone` and threading them out (a side channel beside the effects); a field per write |
| D-5 | Wire: `WriteEffects` gains `repeated string deleted_handles = 14` *(new)*, mapped in `writeEffectsProto` with `idStrings(set.DeletedHandles)`. A merge is not a delete (it updates `merged_into_id`), so a merged handle is not listed | The next free field; `idStrings` drops an empty id | Reusing `handles` with a flag |
| D-6 | Swift: `CatalogWriteEffects.deletedHandles: [String] = []` *(new)*, mapped in `GoStore.mapWriteEffects`. `CatalogUnpromoteImpact{deletes: CatalogDeleteImpactGroup?, unpromotedBridgeSubjectIDs: [String]}` *(new, `Sendable, Equatable`)* in `GenealogyStore.swift`; `getUnpromoteImpact(projectDir:subjectID:) -> CatalogUnpromoteImpact` (non-optional; `deletes` nil when the response has none, as today's `nil`). `EvidenceGraphModel.beginUnpromote` reads `impact.deletes` with today's `total > 0` rule. `session.apply(effects)` ignores `deletedHandles` in this PR | The store keeps one call per RPC; S13-17 reads both new fields | A second store method |
| D-7 | FakeStore stays scripted (`swift-tests.mdc`): `unpromoteImpactBySubject` keeps its `CatalogDeleteImpactGroup` value; `unpromotedBridgesBySubject: [String: [String]]` *(new)* scripts the ids; `getUnpromoteImpact` returns `CatalogUnpromoteImpact(deletes: unpromoteImpactBySubject[id], unpromotedBridgeSubjectIDs: unpromotedBridgesBySubject[id] ?? [])`. Its `unpromoteSubject` keeps returning `nextWriteEffects`, so a test scripts `deletedHandles` there. `ThrowingStore` throws, like its neighbours | No derived engine rule in Swift | FakeStore computing the bridges from its graph |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `identityclaims.FiledBridgesOf`, `FiledBridge` *(new)* | func, type | `bridgeHeads` + the two copies of its filter | Replaces the copies; `bridgeHeads` stays private |
| `promote.Impact` *(new)* | type | `deleteimpact.Group` (today's return) | Holds the group and the ids |
| `Effect.ReportsDelete`, `Set.DeletedHandles` *(new)* | fields | `Effect.Lists`, `Set.Handles` | One more registry fact; `Handles` is the recompute set and survives the delete's ghost walk differently |
| `GetUnpromoteImpactResponse.unpromoted_bridge_subject_ids`, `WriteEffects.deleted_handles` *(new)* | proto fields | — | New facts |
| `CatalogUnpromoteImpact`, `CatalogWriteEffects.deletedHandles`, `FakeStore.unpromotedBridgesBySubject` *(new)* | Swift type, fields | `CatalogDeleteImpactGroup`, `unpromoteImpactBySubject` | Mirror the wire |
| `effects` test `deleted_handles_test.go` *(new)* | test file | `reads_resolve_test.go` | One file per resolved field |

**Builds on.** The effects registry and `Resolve`; `deleteimpact.ReleaseClaim` (the one handle release); S13-08's preflight and handler.
**Deletes.** The duplicated `claimID == nil` filters in `UnfileBridges` and `UnpromoteBridgeHandlesDeleted`.
**Next adds.** S13-17 words the list from `unpromotedBridgeSubjectIDs` and leaves on `deletedHandles`; a later "the page's entity was deleted" rule for Event and Place reads `deletedHandles`; a researcher-facing handle delete reports through `ReportsDelete` with no new code.

### Reach

None — no new navigation path or view state; the two fields have no Swift reader beyond the store until S13-17.

### Scale

Map Scale (slice 0: a Source of 100 people on one Promote page; the 41-Subject obituary). The preflight's extra work is `FiledBridgesOf`, the query `UnpromoteBridgeHandlesDeleted` already runs, now shared: no extra query. `Resolve` checks one bool per change; a Promote Done of 100 rows with their bridges (hundreds of changes) gains one comparison each and no query. `deleted_handles` is empty for every write but an Unpromote or a Subject delete.

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
| `core/database/identityclaims` | `func FiledBridgesOf(tx *sql.Tx, subjectID []byte) ([]FiledBridge, error)` *(new)*; `UnfileBridges`, `UnpromoteBridgeHandlesDeleted` call it |
| `core/database/promote` | `type Impact struct { Deletes deleteimpact.Group; UnpromotedBridgeSubjects [][]byte }` *(new)*; `UnpromoteImpact(c, subjectID) (Impact, error)` |
| `core/database/effects` | `Effect.ReportsDelete bool`, `Set.DeletedHandles [][]byte` *(new)*; `Resolve` fills it; `validate` refuses the flag off handles; the README's `Set` paragraph names it |

### FFI

`api/proto/engine.proto`: `GetUnpromoteImpactResponse.unpromoted_bridge_subject_ids = 2`, `WriteEffects.deleted_handles = 14` *(new)*. Handlers: `api/ffi/handlers/promote.go` `GetUnpromoteImpact`; `api/ffi/handlers/write_effects.go` `writeEffectsProto`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

None — no view changes. Store types and mapping only (D-6, D-7).

### Strings

None — no copy.

## 7. Steps

**STEP-1 — One rule for the filed bridges**
Files: `core/database/identityclaims/bridgefile.go`, `core/database/identityclaims/identityclaims_test.go`.
Do: D-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/identityclaims/... ./core/database/promote/...` passes (T-1, and the existing Unpromote tests unchanged).

**STEP-2 — The preflight returns the bridges**
Files: `core/database/promote/unpromote_impact.go`, `core/database/promote/unpromote_impact_test.go`. Skill: `add-catalog-delete` (read).
Do: D-2.
Check: T-2 passes; `TestUnpromoteImpactListsBridgeHandles` and `TestUnpromoteImpactEmptyWhenHandleStays` pass on `impact.Deletes`.

**STEP-3 — Effects name deleted handles**
Files: `core/database/effects/registry.go`, `core/database/effects/resolve.go`, `core/database/effects/deleted_handles_test.go` *(new)*, `core/database/effects/README.md`.
Do: D-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/... ./core/writes/...` passes (T-3, T-4).

**STEP-4 — The wire**
Files: `api/proto/engine.proto`, `api/ffi/handlers/promote.go`, `api/ffi/handlers/write_effects.go`, `api/ffi/handlers/promote_test.go`, `api/ffi/handlers/subjects_test.go`, `api/ffi/handlers/write_effects_test.go`, generated files. Skill: `add-ffi-handler`, `use-catalog-session`.
Do: D-3, D-5; `./scripts/generate-proto.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes (T-5 – T-8).

**STEP-5 — Swift store**
Files: `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`, `macos/ProvenenciaTests/FakeStoreUnpromoteTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-6, D-7.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes (T-9, T-10). On a host without Xcode, STEP-4's proto generation and this step's checks are run by `scripts/check-pr.sh` on a Mac, as S13-01 did.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `core/database/effects/README.md` (STEP-3), `docs/catalog-deletes.md` if it lists what a delete reports.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-37-unpromote-reports-what-it-removes.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/identityclaims/identityclaims_test.go` `TestFiledBridgesOf` *(new)*, table-driven | P with filed R, Q and unfiled U → [R, Q] with their claim ids, in Subject id order; no bridges → empty; a bridge Subject itself as `subjectID` → its own filed bridges only | AC-1.1, AC-1.3, D-1 |
| T-2 | Go | `core/database/promote/unpromote_impact_test.go` `TestUnpromoteImpactListsUnpromotedBridges` *(new)* | P as T-1 → `UnpromotedBridgeSubjects` [R, Q]; then `Unpromote` → its changes delete exactly R's and Q's accepted claims (and P's); a Subject with no claim → empty `Impact` | AC-1.1, AC-1.2, AC-1.4, D-2 |
| T-3 | Go | `core/database/effects/deleted_handles_test.go` `TestResolveReportsDeletedHandles` *(new)*, table-driven | changes [canonical_entity delete h1, identity_claim delete, canonical_entity update h2] → `DeletedHandles` [h1]; two deletes h1, h3 → [h1, h3]; no handle delete → empty | AC-2.1, AC-2.5, D-4 |
| T-4 | Go | same · `TestReportsDeleteOnlyOnHandles` *(new)* | `validate` of an entry with `ReportsDelete` and `Entity` "source" → error; the shipped registry → `"canonical_entities"` is the only entry with it | D-4 |
| T-5 | Go | `api/ffi/handlers/promote_test.go` `TestGetUnpromoteImpact` (extended) | The AC-1.1 graph → `unpromoted_bridge_subject_ids` [R, Q]; no bridges → empty | AC-1.1, AC-1.3, D-3 |
| T-6 | Go | same · `TestUnpromoteReportsDeletedHandles` *(new)* | Last member → `effects.deleted_handles` [person id]; the AC-2.2 case → [person, relationship handle]; the person keeps a member → empty | AC-2.1 – AC-2.3, D-5 |
| T-7 | Go | `api/ffi/handlers/subjects_test.go` `TestDeleteSubject` (extended) | Deleting a handle's last member → `deleted_handles` [that handle] | AC-2.4 |
| T-8 | Go | `api/ffi/handlers/write_effects_test.go` `TestWriteEffectsProto` (extended) | A `Set` with `DeletedHandles` [h1] → proto `deleted_handles` ["<h1 uuid>"]; empty → nil | AC-2.5, D-5 |
| T-9 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` Unpromote cases (moved onto `CatalogUnpromoteImpact`) | Unchanged behavior: `unpromoteLastRecordNamesTheDeletedHandle` and `unpromotePreflightFailureStaysInTheConfirm` pass with the scripted `deletes` | AC-2.6, D-6, D-7 |
| T-10 | Swift | `macos/ProvenenciaTests/FakeStoreUnpromoteTests.swift` *(new)* `fakeStoreScriptsUnpromoteImpact` | `unpromotedBridgesBySubject["s1"] = ["b1"]` → `getUnpromoteImpact` returns them with the scripted deletes; `nextWriteEffects.deletedHandles` round-trips through `unpromoteSubject` | AC-2.6, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/effects` (the registry's validation), `api/ffi/swift_vocabulary_test.go` if it checks `WriteEffects` field parity, `EvidenceGraphModelTests`.

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
2. The reviewer runs `CGO_ENABLED=1 go test -tags fts5 ./core/database/identityclaims/... ./core/database/promote/... ./core/database/effects/... ./api/ffi/handlers/ -run 'FiledBridgesOf|UnpromoteImpact|ReportsDelete|DeletedHandles|Unpromote|DeleteSubject|WriteEffectsProto'` and sees them pass.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The registry flag reports a delete the app did not mean (a merge) | A merge is an update (D-5); T-3's update case |
| The preflight and the write drift | One function (D-1); T-2 compares them |

**Stop and report if:**

- A second emitter of a `canonical_entity` delete exists at the base (the `rg` in Edges finds more than `handle_gone.go`).
- `UnfileBridges` and `UnpromoteBridgeHandlesDeleted` cannot share `FiledBridgesOf` without changing which claims `UnfileBridges` releases (T-2 or `TestUnpromote` fails).
- `effects.Resolve`'s ghost walk drops a deleted handle's change before the flag is read.
- `api/ffi/swift_vocabulary_test.go` requires a Swift reader for every `WriteEffects` field and `deletedHandles`' mapping does not satisfy it.

## 12. Open questions

None.

## 13. PR

**Title:** S13-37 — Unpromote reports what it removes
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-37-unpromote-reports-what-it-removes.md`.
