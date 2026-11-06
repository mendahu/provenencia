# S14-33 — A bridge between two handles files when it is made

**Status:** Landed in #723 (2026-11-06)
**Spike / slice:** Spike 14 · Slice 7c — Evidence graph polish
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the bridge card already shows membership when its bridge is filed; this PR only makes the engine file it sooner)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-33-auto-file-bridges`
**Version bump:** patch — a researcher-visible behavior fix in the engine; no schema, no new RPC ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md) (read only: the filed claim's audit scope is already registered), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md) (handler test only)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-audit.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 7) Creating a bridge files it when both ends are already handles"; bar outcome 28; dogfood note 2026-11-02 (deleted on land); [`docs/promote-graph-alignment.md` §9.1](../../../promote-graph-alignment.md#91-bridge-filing)
**Planned:** 2026-11-05 against `main` @ `c8a8bdb`
**Review:** [`S14-33-auto-file-bridges.review.md`](S14-33-auto-file-bridges.review.md) — Fits — Ready with changes, 2 rounds; no open blocker; nothing to user

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

A researcher who connects two people who are already promoted (spouse, parent, sibling) sees the relationship filed onto its canonical association in the same save, with no second Promote. The same holds for every bridge kind: participation, location and place relationship. The request came from dogfood (2026-11-02, Jake): filing already happens when the *second end* is promoted after the bridge exists, but not when the bridge is drawn *after* both ends are handles, because `connect.CreateCitedBridge` creates no claim and so nothing triggers filing.

**Why now.** §9.1 of `promote-graph-alignment.md` says filing is automatic everywhere; this is the one write path that breaks that rule, and the Conclusion layer (adjacency, search, the Person page's relationships) stays missing work the researcher already finished.

**What stays.** Filing rules (`connectrules.Decide`, `fileBridge`, `findAssociation`) are unchanged. Promote, Promote batch Done, Unpromote and `UnfileBridges` are unchanged. A refused filing (self-link, place loop) still leaves the bridge unfiled without failing the write. No app chrome changes and no wire changes.

## 2. User stories

**US-1.** As a researcher, when I connect two Subjects that are both already promoted, the bridge files onto its association in the same write, so that the Conclusion layer shows the link at once.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Alice and Bob are promoted Persons; connect Alice → Bob as `relationship` with `relationship_type` = spouse | The bridge Subject has an accepted claim onto a new `relationship` handle; the handle has 1 member; its cached `person` / `related_to` ends are Alice's and Bob's handles; the auto-reconciler cache matches a rebuild |
| AC-1.2 | James (Person) and Birth (Event) are promoted; connect them as `participation` with role subject | The bridge has an accepted claim onto a `participation` handle with James's and Birth's handles as ends |
| AC-1.3 | Alice–Bob spouse is filed (association A, 1 member). Alice2 is joined to Alice's handle and Bob2 to Bob's on a second record; connect Alice2 → Bob2 spouse | The new bridge joins A (A has 2 members); no second `relationship` handle is minted |
| AC-1.4 | Same as AC-1.2, through the FFI `CreateCitedBridge` handler | The response's `effects.handles` contains the new association handle's id |

**US-2.** As a researcher, when the bridge cannot file yet or must not file, the connect still saves and nothing else changes.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Only Alice is promoted; connect Alice → Bob spouse | Write succeeds; the bridge has no accepted claim; no `relationship` handle exists. Promoting Bob afterwards files it (1 handle, 1 member) |
| AC-2.2 | Alice and "Alice again" are both on one Person handle; connect them spouse | Write succeeds; no `relationship` handle; both endpoint claims still stand |
| AC-2.3 | Places York, Upper Canada, Canada all promoted; York part_of Upper Canada and Upper Canada part_of Canada are filed; connect Canada → York part_of | Write succeeds; that bridge has no accepted claim; `place_relationship` handles stay at 2 |
| AC-2.4 | After a participation drawn between a promoted Person and a promoted Event has filed, run a Promote batch Done on the graph (`promote.SaveBatch` with no rows) | The bridge still has exactly one accepted claim |

## 3. Scope

| In | Covered by |
| --- | --- |
| `connect.CreateCitedBridge` files the new bridge in the same transaction through `identityclaims.FileBridgesTx`, every bridge kind | STEP-1, AC-1.1 – AC-1.3 |
| A refused or waiting filing does not fail the connect write | STEP-1, AC-2.1 – AC-2.3 |
| The filed claim and any minted handle reach the write's effects (handles, evidence graph) through the existing `effects` registry | STEP-1, AC-1.4 |
| The two batch tests that relied on "a bridge drawn between handles stays unfiled" seed that state another way | STEP-2 |
| New Go tests in `promote` and the FFI handler | STEP-2, STEP-3 |
| `promote-graph-alignment.md` §9.1 "When" names bridge creation; the Promoted/Promoted row in §5.1 says when that state still occurs; `conclusion-layer-data-model.md` §5.4 and the `create_cited_bridge` row of `audit-revision-history.md` name the new filing | STEP-5 |

| Out | Home |
| --- | --- |
| Per-card Promote footer on bridge cards | not planned (map scope entry "Not this PR") |
| Changing Unpromote / `UnfileBridges` (including warning when unpromoting one end of an auto-filed bridge — note open question 4) | not planned; behaviour stays as §9.1 "An end that leaves later" |
| Back-filling unfiled bridges between handles in catalogs written before this PR | not planned: Promote batch Done on the graph already files them (`FileEvidenceGraphBridgesTx`) |
| Filing when a bridge's Observations are edited later (a changed `relationship_type`) | not planned; no dogfood report |
| A toast or other notice that the bridge filed (note open question 6) | not planned: bar outcome 28 asks only that membership shows; a notice is chrome and needs a brief |
| Any Swift change (FakeStore, `EvidenceGraphModel`, `PromoteModel`) | not planned: see D-5 |
| An in-transaction mint-and-claim step shared by `promotetest.ClaimWithoutFiling` and the copies in `core/database/effects/conclusion_handles_test.go` (which run inside an open `*sql.Tx` and so cannot call the `writes.Call` helper) | not planned: a later test tidy; this PR does not touch that file |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Connect write | `core/database/connect/connect.go` `CreateCitedBridge` | Validates the rule, inserts the bridge Subject (`subjects.InsertTx`), its position, then the Observations onto a new or existing Citation, and returns `Result{Subject, Citation, Observations}` with the row changes. It never calls `identityclaims`, so a bridge between two handles stays unfiled |
| Filing | `core/database/identityclaims/bridgefile.go` `FileBridgesTx(tx *sql.Tx, subjectID []byte)` | Lists bridge Subjects matching `fileCond` (the id itself **or** a bridge touching it), and for each runs `filePass.fileBridge`: skips a bridge that already has an accepted claim, waits when an end is unpromoted, returns `filed=false` when `connectrules.Decide` refuses (self-link, loop), otherwise joins or mints the association and inserts the accepted claim. Returns association ids and the `identity_claim` / `canonical_entity` changes. Does not commit |
| Existing callers | `core/database/identityclaims/identityclaims.go` `AppendFiling` (used by `Create` and `promote.Save` at `core/database/promote/promote.go`), `core/database/promote/batch.go` `SaveBatch` → `FileEvidenceGraphBridgesTx` | Every caller discards the association ids; the handles are recomputed by `writes.Run` from the changes |
| Effects | `core/database/effects/registry.go` `"identity_claims"` and `"canonical_entities"` entries | An `identity_claim` change resolves to its handle (`field("entity_id")`), its observers, the claim's Evidence graph (`claimGraph`) and `ReadEvidenceGraphProgress`; `writes.Run` (`core/writes/run.go`) then runs `autoreconciler.RecomputeReported` on the handles. Nothing to add |
| FFI handler | `api/ffi/handlers/connect.go` `CreateCitedBridge` | Runs `connect.CreateCitedBridge` inside `writes.Run` (action `create_cited_bridge`) and returns `Effects: writeEffectsProto(wrote)`. Unchanged by this PR |
| Import graph | `go list -deps ./core/database/identityclaims` | Does not include `core/database/connect`, so `connect` may import `identityclaims` without a cycle |
| Swift | `macos/App/Platform/FakeStore.swift` `createCitedBridge`; `macos/App/Platform/GoStore.swift` `createCitedBridge` | FakeStore does not file bridges on any path (no filing code in it; working rule "FakeStore re-implements no engine rule"). GoStore applies the returned `CatalogWriteEffects`, which already reload the graph because the write creates a Subject on it |
| Tests — filing | `core/database/promote/bridges_test.go` `TestBridgeFiling`, `bridgeWorld` helpers (`node`, `relationship`, `participation`, `location`, `placeRelationship`, `promote`, `join`, `handles`, `members`, `cachedEntity`, `relationships`) | Cover "bridge first, then promote". Every subtest creates its bridges before the ends are promoted, so they stay green with this PR |
| Tests — batch | `core/database/promote/batch_test.go` `TestSaveBatchFilesBridgeWhoseEndsWereAlreadyPromoted` (line 233), `TestSaveBatchFilesOnlyItsGraph` (line 378), helpers `newBatchFixture`, `bare`, `participation`, `bridgeClaims`, `save`, `revision` | Both promote the two ends **first**, then draw the participation, and assert it is unfiled before the batch. With this PR the connect files it, so both fail. Verified at the planning base by applying the STEP-1 hook and running `go test -tags fts5 ./core/... ./api/...`: these two are the only failures (`batch_test.go:241 bridge filed before the batch`, `batch_test.go:386 participation claims 1 before any Done, want 0`) |
| Tests — connect | `core/database/connect/connect_test.go` `TestCreateCitedBridge` | Package `connect` (internal test). Builds bridges between unpromoted Subjects only; no fixture promotes. The filing tests go beside `TestBridgeFiling` in `promote` (D-7) |
| Tests — handler | `api/ffi/handlers/connect_test.go` `TestCreateCitedBridge`, `connectFixture`; `api/ffi/handlers/promote_test.go` `promoteInDir` | Table of `rpcTest` cases through `runRPC`; `promoteInDir(t, dir, userID, subjectID, "")` promotes a Subject in a fixture project |
| Domain doc | `docs/promote-graph-alignment.md` §9.1 "When:" bullet; §5.1 table row "Promoted / Promoted" | "after **any** claim create" — does not name bridge creation. The row calls a bridge between two handles a "missing / unfiled relationship" the Promote page files |

**Precedent.** `identityclaims.Create` (`identityclaims.go`) and `promote.Save` (`promote.go`, line ~75) both finish their write with `identityclaims.AppendFiling(q, subjectID, changes)` and return the merged changes; the filing rides the same revision. Copy that shape: one call at the end of the write, filed changes appended, association ids discarded. Do not copy `SaveBatch`'s graph-wide `FileEvidenceGraphBridgesTx`: it would file unrelated old bridges inside a connect write.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Filing runs inside `connect.CreateCitedBridge`, as its last step before `return`, on `tx.Tx` | The working rule names this function; one owning write means every caller (the FFI handler and the many Go tests that build bridges through it) gets the rule, and the filing lands on the connect's own revision | A follow-up in the FFI handler — because Go callers would skip it; a Swift follow-up RPC — because it re-implements the engine rule in the app (note fix direction 3) |
| D-2 | Call `identityclaims.AppendFiling(q, bridge.ID, changes)` and keep its returned changes; discard the association ids | `AppendFiling` is the existing "file what this Subject completes and append the changes" helper; with the bridge's own id, `fileCond` selects the new bridge itself (`SELECT ?`) and bridges whose Observations name it as a value (none: it was minted in this transaction). Same shape as `promote.Save` | `FileBridgesTx` on both endpoint ids (note fix direction 2) — because it also files older unfiled bridges touching either end, which is batch Done's job (Out); a new `FileBridgeTx(one id)` — because `AppendFiling` already does exactly this with no new code |
| D-3 | Every bridge kind files: `participation`, `relationship`, `location`, `place_relationship`, and any later kind in `connectrules.BridgeTypeKeys()` | The working rule says every bridge kind; `fileBridge` already handles all of them through `connectrules.LookupBridge` | Relationship-only — because it would need a kind check the filer does not have and contradicts §9.1 |
| D-4 | A wait or refusal (`filed=false`) is silent: the connect returns success and `Result` is unchanged. A database error from filing fails the whole write (rolls back the bridge) | Matches the "Refused filings don't fail the claim" rule in §9.1; a real SQL error is not a refusal | Returning the filing outcome on `connect.Result` or the proto — because no caller reads it: the app reads membership from the reloaded graph (D-5) |
| D-5 | No Swift and no proto change. FakeStore does not file on connect | The response's effects already carry the new handle and the graph; the graph reload shows the bridge's membership (`PromoteModel.bridgeFacts` reads `bridge.membership`). FakeStore re-implements no engine rule, and it does not file on promote either | Teaching FakeStore to file — because that is the engine rule the working rule forbids copying |
| D-6 | The two batch tests seed "an unfiled bridge between two handles" by drawing the bridge **first**, then claiming both ends with `promotetest.ClaimWithoutFiling` *(new)*, which calls `canonicalentities.InsertTx` + `identityclaims.InsertTx` (accepted) in one `writes.Call` and returns the handle id. This stands for a catalog written before this PR | The state is still real (older catalogs, a filing that was refused when made), and batch Done must still file it; `InsertTx` is the claim write without `AppendFiling`. `promotetest` is the package that exists so tests do not each spell a claim write | Deleting the two tests — because batch Done's graph-wide filing is still the back-fill path (Out line); unfiling with `deleteimpact.ReleaseClaim` after the connect — because it also deletes the minted handle and so tests a different state; a private `batchFixture` method — because `core/database/effects/conclusion_handles_test.go` already spells the same mint-and-claim pair twice, and a third private copy would follow |
| D-7 | New behaviour tests T-1 – T-5 go in `core/database/promote/bridges_test.go` as `TestBridgeFilesWhenDrawnBetweenHandles` *(new)*, using `bridgeWorld`; T-6 goes in `core/database/promote/batch_test.go` as `TestSaveBatchKeepsABridgeFiledOnCreate` *(new)*, using `batchFixture` | `bridgeWorld` already builds every bridge kind through `connect.CreateCitedBridge` and holds `TestBridgeFiling`; `batchFixture` already has `save`, `revision`, `bridgeClaims` and `participation` for a batch test | A new fixture in `connect_test.go` — because it would duplicate `bridgeWorld` |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| Call to `identityclaims.AppendFiling` in `connect.CreateCitedBridge` | call | `promote.Save`'s identical call | Reused as is |
| `TestBridgeFilesWhenDrawnBetweenHandles` *(new)* | test | `TestBridgeFiling` in the same file | A separate top-level test keeps "drawn after the ends" apart from `TestBridgeFiling`'s "drawn before"; it reuses every `bridgeWorld` helper |
| `promotetest.ClaimWithoutFiling(t testing.TB, c *database.Catalog, userID []byte, s subjects.Subject) []byte` *(new)* | test helper | `promotetest.Save` (files bridges); `promote.fileSubject` (unexported); the `claim` closure and inline copy in `core/database/effects/conclusion_handles_test.go` | `Save` files bridges, which is the state the batch tests must avoid; `fileSubject` is unexported; the effects test copies are private to that test. Moving those copies onto the helper is Out |
| Handler case "files a participation between two promoted subjects" *(new)* in `TestCreateCitedBridge` | test case | the "creates participation" case; `promoteInDir` | Reuses `connectFixture` and `promoteInDir` |

**Builds on.** `identityclaims.AppendFiling` / `FileBridgesTx`, `connectrules.Decide`, the `effects` registry entries for `identity_claims` and `canonical_entities`, `writes.Run`'s handle recompute.
**Deletes.** None — nothing becomes dead. `FileEvidenceGraphBridgesTx` stays: batch Done still back-fills older bridges.
**Next adds.** A later bridge kind (organization relationship, [`docs/ideas/organizations.md`](../../../ideas/organizations.md)) files on create with no change here once it is in `connectrules`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A `create_cited_bridge` revision now can carry `identity_claim` and `canonical_entity` creates | `effects.Resolve`, `audit.Record`, `autoreconciler.RecomputeReported` (`core/writes/run.go`) | Same change types `promote_subject` already records; the effects registry resolves them; the prototype run passed every `core/...` and `api/...` test apart from the two in §4, including the effects and audit-scope tests |
| A bridge card is filed the moment it appears on the graph | `PromoteModel.bridgeFacts` (`alreadyPromoted: bridge.membership != nil`), `PromoteFlow` connections | Already handled: these read a filed bridge the same as one filed by a later promote |
| Undo of a filed bridge: deleting the bridge Subject | `subjects.Delete` → `deleteimpact` | `TestBridgeSubjectDeleteKeepsTheEndClaims` already covers deleting a filed bridge |

### Scale

At the map's Scale (the 60-graph census): one connect adds one `bridgeHeads` query seeking by the new bridge's id, one `bridgeObservations` read, one accepted-claim lookup per end value, and for `place_relationship` one `filedLinks` read of that term's filed links (the same reads `promote.Save` makes for each bridge it files). No per-graph or per-catalog loop is added. The write's effects grow by at most one handle and its observers, which `RecomputeReported` already recomputes for promote. No budget is set or leaned on.

### Edges

- **Assumptions.** `fileBridge` assumes the bridge's Observations are written before it runs: the call is after `InsertManyTx` / `InsertWithObservationsTx` (STEP-1 places it there). It assumes a bridge with an accepted claim is filed and skips it, so a second filing pass (batch Done) is a no-op — AC-2.4, T-6.
- **Neighbors.** One end unpromoted → waits, files when the second end is promoted (AC-2.1, T-4); both ends on one handle → refused, write succeeds (AC-2.2, T-5); loop among filed links → refused (AC-2.3, T-5); an existing association between the same handles → joined, not minted (AC-1.3, T-3); a bridge drawn before either end is promoted → unchanged (`TestBridgeFiling` stays green).
- **Removed inputs.** Unpromoting one end after an auto-file runs `UnfileBridges`, as for a bridge filed by promote; unchanged and Out.
- **Real data.** None — no table- or seed-driven rule changes; the bridge kinds come from `connectrules.BridgeTypeKeys()` as today.
- **Untrusted size.** None — the call files one bridge whose Observations this write just validated (`exactObservationInputs` requires exactly the rule's edges plus its disambiguation).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/connect` | `CreateCitedBridge`: after the Citation/Observations branch and before building `Result`, `_, changes, err = identityclaims.AppendFiling(q, bridge.ID, changes)`; on `err` return `Result{}, nil, err`. Import `github.com/mendahu/provenencia/core/database/identityclaims`. Update the package doc to "Package connect creates cited Interpretation bridges atomically and files a bridge whose ends are already handles." Update the function doc comment's first sentence to: "CreateCitedBridge writes a bridge Subject, position, Citation, and Observations on tx, and files the bridge when both ends are already handles (§9.1). A wait or a refused filing is not an error." |

| `core/database/identityclaims` | Doc comments only: `AppendFiling` (`identityclaims.go`) and `FileBridgesTx` (`bridgefile.go`) each gain "subjectID is a new member (the bridges touching it file) or a new bridge (it files itself)." |

Errors: none new.

### FFI

None — the handler and `engine.proto` are unchanged; the effects already returned carry the filed handle.

### macOS

None — see D-5.

### Strings

None — no copy; the write's scope (what one connect covers) gains the filing, and no existing string names the connect's scope.

## 7. Steps

**STEP-1 — File the bridge on create**
Files: `core/database/connect/connect.go`, `core/database/identityclaims/identityclaims.go`, `core/database/identityclaims/bridgefile.go` (doc comments).
Do: the Go core changes above (D-1, D-2, D-4).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` fails only `TestSaveBatchFilesBridgeWhoseEndsWereAlreadyPromoted` and `TestSaveBatchFilesOnlyItsGraph` (the result at the planning base).

**STEP-2 — Re-seed the two batch tests**
Files: `core/database/promote/promotetest/promotetest.go`, `core/database/promote/batch_test.go`.
Do: add `func ClaimWithoutFiling(t testing.TB, c *database.Catalog, userID []byte, s subjects.Subject) []byte` *(new)* to `promotetest` beside `Save`: one `writes.Call(c, writes.Op{Action: "create_identity_claim", UserID: userID}, …)` returning `[]byte`, which calls `canonicalentities.InsertTx(tx.Tx, canonicalentities.CreateInput{SubjectTypeID: s.SubjectTypeID})`, then `identityclaims.InsertTx(tx.Tx, identityclaims.CreateInput{SubjectID: s.ID, EntityID: entity.ID, Status: identityclaims.StatusAccepted})`, and returns both changes; the helper returns the entity id and fails the test on error. Its doc comment: it claims s onto a new handle without filing bridges, as a catalog written before S14-33 left a bridge drawn between two handles. Update the package doc to say it holds the claim helpers. In both batch tests, move `f.participation(person, event)` before the ends are claimed and replace the two `promotetest.Save` calls with `promotetest.ClaimWithoutFiling(t, f.c, userID, …)` (D-6). In `TestSaveBatchFilesBridgeWhoseEndsWereAlreadyPromoted`, `per` and `evt` go; the batch rows' `EntityID` take the helper's return values. Keep every assertion.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promote/...` passes.

**STEP-3 — Filing-on-create tests**
Files: `core/database/promote/bridges_test.go`.
Do: add `TestBridgeFilesWhenDrawnBetweenHandles` *(new)* with subtests T-1 – T-5 from §8 (each `w.node` takes any distinct grid position, with the 0 / 2 / 4 spacing `TestBridgeFiling` uses; "the bridge" is the `connect.Result.Subject` the helper returns), using `bridgeWorld` (`node`, `promote`, `join`, `relationship`, `participation`, `location`, `placeRelationship`, `term`, `prop`, `handles`, `members`, `cachedEntity`, `relationships`) and `w.filed(s)` (`core/database/promote/unpromote_test.go`) for "has a claim". End T-1 and T-3 with `autoreconcilertest.AssertCacheMatchesRebuild(t, w.c)`. Add `TestSaveBatchKeepsABridgeFiledOnCreate` *(new)* (T-6) to `batch_test.go`.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestBridgeFilesWhenDrawnBetweenHandles ./core/database/promote/` passes.

**STEP-4 — Handler test**
Files: `api/ffi/handlers/connect_test.go`. Skill: `add-ffi-handler`.
Do: add T-7 as a `rpcTest` case in `TestCreateCitedBridge`, the same participation request as the "creates participation" case (`connectFixture`'s Person, Event, properties and role term; both already have positions). Declare `var dir string` in the test function above the table; `reqFn` sets it from `connectFixture`, calls `promoteInDir(t, dir, userID, personID, "")` and `promoteInDir(t, dir, userID, eventID, "")`, then returns the request. `after` unmarshals the response, reads the bridge's accepted claim inside `withProjectCatalog(dir, …)` with `identityclaims.AcceptedEntityForSubject(c, <bridge id bytes from uuid.MustParse(created.GetSubject().GetId())>)`, and asserts `created.GetEffects().GetHandles()` contains `database.UUIDString(claim.EntityID)` (use `slices.Contains`). Add the imports `slices`, `github.com/google/uuid`, `github.com/mendahu/provenencia/core/database` and `github.com/mendahu/provenencia/core/database/identityclaims` to `connect_test.go` where they are missing.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestCreateCitedBridge ./api/ffi/handlers/` passes.

**STEP-5 — Domain doc**
Files: `docs/promote-graph-alignment.md`, `docs/conclusion-layer-data-model.md`, `docs/audit-revision-history.md`.
Do: in §9.1 "When:", the first sentence becomes "after **any** claim create, every bridge Subject whose ends are now both handles is filed in the same transaction; and when a bridge is drawn between two Subjects that are already handles (`connect.CreateCitedBridge`), that bridge is filed in the same transaction." The bullet's other sentences stay. The bullet title "**Refused filings don't fail the claim.**" becomes "**Refused filings don't fail the claim or the connect.**" The §5.1 "Promoted / Promoted" row keeps its meaning and adds: "Since S14-33 a bridge drawn between two handles files when it is made, so this row appears for a bridge from an older catalog or one whose filing was refused when it was made." In `conclusion-layer-data-model.md` §5.4 (the paragraph that starts "**Bridge Subjects**"), "That happens on Done for the bridge's Evidence graph, or later, when a claim makes its second end a member." becomes "That happens when the bridge is drawn between two handles, on Done for the bridge's Evidence graph, or later, when a claim makes its second end a member."; in the refusal bullet, "never fails the claim that triggered it" becomes "never fails the write that triggered it". In `audit-revision-history.md`'s action table, the `create_cited_bridge` row becomes "Atomic Connect: bridge subject + Citation + edge Observations, and, when both ends are handles, the bridge's identity claim (and a minted association)".
Check: `rg -n "CreateCitedBridge" docs/promote-graph-alignment.md` prints the §9.1 line; `rg -n "drawn between two handles" docs/conclusion-layer-data-model.md` and `rg -n "identity claim" docs/audit-revision-history.md` each print the edited line.

**STEP-6 — Version bump**
Files: `VERSION`, `core/version.go`, every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj`. Rule: `.cursor/rules/versioning.mdc`.
Do: raise PATCH by one from what `main` holds when you branch (0.3.4 → 0.3.5 at the planning base; S14-32 may land first and take 0.3.5, then this PR takes 0.3.6). All three places agree.
Check: `rg -n "<old version>" VERSION core/version.go macos/Provenencia.xcodeproj/project.pbxproj` prints nothing.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-14/deployment-plan.md` (map row status → `Landed in #N`, Plan column keeps this file's link; "Where the notes went" row for the dogfood note becomes "Dogfood 2026-11-02: relationship between promoted subjects does not file (note deleted YYYY-MM-DD; landed #N)"), `docs/deployment-plan/spike-14/completed.md`, the dogfood note `ux/2026-11-02-relationship-auto-promote-between-promoted-subjects.md` (deleted), this plan.
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-33-auto-file-bridges.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("File a bridge when it is drawn between two promoted Subjects."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/promote/bridges_test.go` `TestBridgeFilesWhenDrawnBetweenHandles` / "relationship between two handles files" *(new)* | promote Alice, promote Bob, `relationship(alice, bob, spouse)` → `handles("relationship")` has 1; `members` = 1; bridge has an accepted claim onto it; `cachedEntity(assoc, "person")` = Alice's handle, `cachedEntity(assoc, "related_to")` = Bob's; cache matches rebuild | AC-1.1, D-1, D-2 |
| T-2 | Go | same / "participation, location and place relationship file" *(new)* | promote James and Birth, `participation(james, birth, subjectRole)` → 1 `participation` handle; promote Birth2 and York, `location(birth2, york)` → 1 `location` handle; promote two places, `placeRelationship(a, b, partOf)` with `partOf := w.term(w.prop("place_relationship_type"), "part_of")` → 1 `place_relationship` handle | AC-1.2, D-3 |
| T-3 | Go | same / "a second record joins the association" *(new)* | after T-1's setup, `alice2` joined to Alice's handle, `bob2` to Bob's, `relationship(alice2, bob2, spouse)` → `handles("relationship")` still 1, `members` = 2; cache matches rebuild | AC-1.3 |
| T-4 | Go | same / "one end unpromoted waits" *(new)* | promote Alice only, `relationship(alice, bob, spouse)` → no error, bridge has no claim, 0 handles; then promote Bob → 1 handle, 1 member | AC-2.1 |
| T-5 | Go | same / "a refused filing does not fail the connect" *(new)* | (a) promote Alice, join "Alice again" to her handle, `relationship(alice, dup, spouse)` → no error, 0 `relationship` handles, both end claims stand. (b) promote York, Upper Canada, Canada; `placeRelationship(york, uc, part_of)`, `placeRelationship(uc, canada, part_of)` file; `placeRelationship(canada, york, part_of)` → no error, that bridge has no claim, `handles("place_relationship")` = 2 | AC-2.2, AC-2.3, D-4 |
| T-6 | Go | `core/database/promote/batch_test.go` `TestSaveBatchKeepsABridgeFiledOnCreate` *(new)* | `promotetest.Save` on a bare person and event, `f.participation(person, event)` → `f.bridgeClaims("participation")` = 1; `f.save(promote.Batch{GraphID: f.graph, SeenRevision: f.revision()})` → still 1 | AC-2.4 |
| T-7 | Go (FFI) | `api/ffi/handlers/connect_test.go` `TestCreateCitedBridge` / "files a participation between two promoted subjects" *(new)* | Person and Event promoted with `promoteInDir`; participation request → the bridge has an accepted claim, and `effects.handles` contains that claim's entity id | AC-1.4, D-5 |
| T-8 | Go | `core/database/promote/batch_test.go` `TestSaveBatchFilesBridgeWhoseEndsWereAlreadyPromoted`, `TestSaveBatchFilesOnlyItsGraph` (re-seeded) | ends claimed with `promotetest.ClaimWithoutFiling` after the bridge → unfiled before the batch; filed by the batch as today | D-6 |

T-6's empty batch is the shape `TestSaveBatchFilesOnlyItsGraph` already uses.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/promote` (batch tests, `TestBridgeFilingReadsDoNotGrowWithTheBridges`), `core/database/effects` (audit scopes and handles per action), `core/database/graphcache`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks); off macOS the script skips `xcode-test`: report it as not run (this PR changes no Swift)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it) — nothing renamed
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump: patch (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the notes row; completed.md: one short paragraph of what shipped; the dogfood note deleted
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain docs updated: docs/promote-graph-alignment.md §9.1 "When" and the §5.1 Promoted/Promoted row; docs/conclusion-layer-data-model.md §5.4; docs/audit-revision-history.md `create_cited_bridge` row
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source that has an Evidence graph. Add Persons Alice and Bob; promote both (each card's Promote → Done).
2. Connect Alice → Bob, choose Spouse, cite and save.
3. The relationship mid-card shows its membership (filed) at once; open Promote from the graph: the connection reads as already filed and the footer counts no connection to file. Alice's Person page lists Bob as spouse.
4. Repeat with a promoted Event and Place connected as a location; the location shows filed.

Run on a Mac; off macOS, report the manual check as **not run** in the PR description.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Another test or path relies on "drawn between handles stays unfiled" | The prototype run at the planning base found only the two batch tests; STEP-1's check re-runs it |
| Filing reads an Observation that is not written yet | STEP-1 places the call after both Observation branches; T-1 and T-2 fail otherwise |

**Stop and report if:**

- STEP-1's check shows any failing test other than the two named. Do not re-seed it yourself: report the test and its failure line.
- `go build ./...` reports an import cycle after `connect` imports `identityclaims`.
- `AppendFiling` or `FileBridgesTx` no longer has the signature in §4.
- A test in `core/database/effects` fails because `create_cited_bridge` is pinned to a fixed set of entity types (that would need a registry change this plan does not name).
- T-7 needs a proto or handler change to see the handle in `effects.handles`.

## 12. Open questions

None.

## 13. PR

**Title:** S14-33 — A bridge between two handles files when it is made
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence (T-1 – T-8); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-33-auto-file-bridges.md`.
