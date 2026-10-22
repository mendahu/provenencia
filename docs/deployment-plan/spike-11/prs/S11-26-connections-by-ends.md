# S11-26 — Connections file by their ends

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 6 — Connection subrows
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the connections panel loses its per-row switch and shows the state caption it already shows for unfiled rows; the panel itself is replaced by S11-27 under S11-D1)
**Board:** —
**Depends on:** —
**Blocks:** S11-27, S12-09
**Branch:** `s11-26-connections-by-ends`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md) (proto field removal and regeneration), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [Connection subrows](../promote-connection-subrows.md) "The problem" §2 and "What it would change"; [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3 ("Bridges aren't rows"), §9 ("Bridges" bullet) and §9.1; [`docs/conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §5.4; working rules "High tolerance for churn", "One wire vocabulary"
**Planned:** 2026-10-22 against `main` @ `1e8a580`
**Review:** — (not reviewed yet)

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

A connection (participation, location, relationship, place relationship) files on Done exactly when both its ends are handles and they are two different handles. The researcher can no longer switch one off for this Done only: that switch hid a connection that then filed itself silently on the next Done or the next claim (bridges file automatically, §9.1), so what the page showed was not what the catalog ended up with. A researcher who disputes a connection fixes it on the Evidence graph.

**Why now.** S11-27 builds connection subrows whose state comes only from their ends ([connection subrows](../promote-connection-subrows.md) "The idea"). Removing the switch first keeps S11-27 a view change, and S12-09 rewrites the same Promote filing code per graph.

**What stays.** Filing itself: `fileBridge`, the association keys, self-link refusal, `UnfileBridges`, and every Promote row control. The connections panel, its header count and the footer count stay (they go or change in S11-27). `ConnectionState.files`, `.endSkipped` and `.selfLink` and their captions stay.

## 2. User stories

**US-1.** As a researcher, when I press Done, every connection whose two ends are filed (on this Done or earlier) as two different handles is filed, so that the page never shows a connection as left out that will file later anyway.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source with person John and event Birth, a participation between them; both rows New; Done | One participation claim is filed |
| AC-1.2 | Same Source; John New, Birth Skip; Done | No participation claim is filed |
| AC-1.3 | The Promote page's connections panel with a connection whose ends both file | The row shows the caption "Filed" and no switch; no control on the row changes its state |
| AC-1.4 | The connections panel header | Reads "N filed · M not filed" (no "switched off" count) |

**US-2.** As a contributor, the Apply request, `promote.Batch` and the bridge filing functions have no way to leave a filable connection out, so that no caller can bring the hidden third state back.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `rg -n "SkipBridge\|skip_bridge\|skipBridge\|bridgeOff\|toggleBridge"` over `api/`, `core/`, `macos/App/`, `macos/ProvenenciaTests/` | Only the `reserved "skip_bridge_ids"` line in `api/proto/engine.proto` matches |
| AC-2.2 | `identityclaims.FileBridgesTx` and `identityclaims.FileSourceBridgesTx` | Take `(tx, id)` only |
| AC-2.3 | `ApplyPromoteGraphAlignmentRequest` field 6 | Is `reserved 6; reserved "skip_bridge_ids";` |

**US-3.** As a researcher, when I leave Promote with unsaved choices, the leave guard counts only the rows I changed.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Change the target on 2 rows, then navigate away | The guard reads "You chose targets on 2 rows. Nothing has been filed; opening Promote again proposes the graph afresh." |

## 3. Scope

| In | Covered by |
| --- | --- |
| Remove `promote.Batch.SkipBridgeIDs` and its validation in `SaveBatch` | STEP-1, AC-2.1 |
| Remove the `skip` parameter of `FileBridgesTx`, `FileSourceBridgesTx` and `fileHeads` | STEP-1, AC-2.2 |
| Replace `TestSaveBatchSkipsSwitchedOffBridge` with a skipped-end test | STEP-1, AC-1.2 |
| Reserve `ApplyPromoteGraphAlignmentRequest` field 6; regenerate Go and Swift; handler stops parsing it | STEP-2, AC-2.3 |
| Remove `skipBridgeIDs:` from `GenealogyStore.applyPromoteGraphAlignment` and every conformer | STEP-3, AC-2.1 |
| Remove `PromoteFlow.bridgeOff`, `toggleBridge`, `skipBridgeIDs()`, `ConnectionState.off`; `PromoteModel.toggleBridge` | STEP-4, AC-2.1 |
| Connections panel row shows the state caption for every state; switch and `bridgeOn` removed | STEP-5, AC-1.3 |
| Copy: header meta, hint, leave guard, `.files` caption comment; remove `promote.connections.off` | STEP-5, AC-1.4, AC-3.1 |
| Docs: `promote-graph-alignment.md` §3 and §9, `conclusion-layer-data-model.md` §5.4 | STEP-6 |

| Out | Home |
| --- | --- |
| Connection subrows, `connections(hostedBy:)`, removing the panel and `bridgesOpen` | S11-27 |
| New wording for the `.files` caption, or a VoiceOver phrase for a waiting connection | S11-27 (S11-D1) |
| A "open on Evidence graph" action on a connection | S11-27 decides with S11-D1; otherwise not planned |
| Filing per graph | S12-09 |
| Any change to `fileBridge`, `UnfileBridges`, or association keys | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Filing | `core/database/identityclaims/bridgefile.go` `FileBridgesTx(tx, subjectID, skip)`, `FileSourceBridgesTx(tx, sourceID, skip)`, `fileHeads(tx, bridges, skip)` | `fileHeads` skips every bridge id in `skip`; doc comments describe the switch ("skip names bridge Subjects the caller switched off") |
| Filing callers | `core/database/identityclaims/identityclaims.go` (`FileBridgesTx(tx, subjectID, nil)`); `core/database/effects/conclusion_handles_test.go` (`FileSourceBridgesTx(tx.Tx, src.ID, nil)`); `core/database/promote/batch.go` `SaveBatch` (`FileSourceBridgesTx(q, in.SourceID, skip)`) | Two pass nil; `SaveBatch` passes the switch-off set |
| Batch | `core/database/promote/batch.go` `Batch.SkipBridgeIDs`, `SaveBatch` | Field documented as "Evidence bridges the researcher switched off"; `SaveBatch` builds `skip` and returns `ErrInvalid` for an id that is not 16 bytes; its doc says "bridges (except switched-off ones)" |
| Handler | `api/ffi/handlers/promote_batch.go` | `parseIDs(req.GetSkipBridgeIds())` → `promote.Batch.SkipBridgeIDs` |
| Wire | `api/proto/engine.proto` `ApplyPromoteGraphAlignmentRequest` | `repeated string skip_bridge_ids = 6;`; generated in `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift` (`skipBridgeIds`) |
| Store | `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift` `applyPromoteGraphAlignment(…, skipBridgeIDs:)`; `FakeStore.lastPromoteBatch: (rows:, skipBridgeIDs:)?`; `macos/ProvenenciaTests/OnboardingModelTests.swift` (a throwing conformer) | Every conformer takes and forwards the list |
| Flow | `macos/App/Features/Promote/PromoteFlow.swift` `ConnectionState.off`, `bridgeOff`, `filedBridges`, `toggleBridge(_:)`, `skipBridgeIDs()` | `filedBridges` maps an id in `bridgeOff` to `.off` before checking its ends; `toggleBridge` sets `manual` |
| Model | `macos/App/Features/Promote/PromoteModel.swift` `toggleBridge(_:)`, Done (`skipBridgeIDs: flow.skipBridgeIDs()`) | Forwards the switch |
| View | `macos/App/Features/Promote/PromoteView.swift` `connectionRow(_:)`, `bridgeOn(_:)`, `connectionMeta`, `connectionsPanel`, the `.pvConfirm` leave copy | `.files` / `.off` rows render a `PVSwitch`; the others a muted italic caption; the leave guard passes `bridgeOff.count` |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Promote.connectionState`, `connectionFiled`, `connectionsMeta(filed:off:unfiled:)`, `connectionsHint`, `leaveDetail(rows:connections:)`; `macos/App/Resources/Localizable.xcstrings` (English only) | `promote.connections.off` = "Not filed this time"; `promote.connections.meta` = "%1$lld filed · %2$lld switched off · %3$lld not filed"; `promote.connections.hint` = "Switch off a relationship you don’t accept from this Source; it stays on the graph"; `promote.leave.pageMessage` names "switched off %2$lld connections" |
| Tests | `core/database/promote/batch_test.go` `TestSaveBatchSkipsSwitchedOffBridge`, `TestSaveBatchFilesBridgeWhoseEndsWereAlreadyPromoted` (ends with a comment and `_ = bridge` about the switch) | No test covers a bridge with a Skipped end |
| Tests | `macos/ProvenenciaTests/PromoteFlowTests.swift` `doneSendsPinsAndASwitchedOffBridge`, `promoteAllStartsExpandedAndHidesFiledBridges` (asserts `skipBridgeIDs().isEmpty`); `macos/ProvenenciaTests/PromoteModelTests.swift` (reads `store.lastPromoteBatch?.rows` at two sites) | |
| Docs | `docs/promote-graph-alignment.md` §3 bullet "Bridges aren't rows", §9 bullet "Bridges"; `docs/conclusion-layer-data-model.md` §5.4 bullet "On the Promote page, each bridge can be switched off" | State the switch as a rule |

**Precedent.** Removing a proto field: `MetadataWorkspaceEntry` in `api/proto/engine.proto` keeps `reserved 3, 8; reserved "date_value_id", "date";` for its removed fields. Copy that shape. Removing a wire field together with every caller in one PR follows S11-25 ([plan](S11-25-claim-evidence-rpc.md), proposal field 9 removed).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Field 6 of `ApplyPromoteGraphAlignmentRequest` is deleted and reserved by number and name | The repo's removal convention (Precedent); a reserved number stops a later field from reusing it | Leaving it unused — dead wire, against "High tolerance for churn" |
| D-2 | `FileBridgesTx(tx *sql.Tx, subjectID []byte)` and `FileSourceBridgesTx(tx *sql.Tx, sourceID []byte)`, each returning `(assocIDs [][]byte, changes []rowchange.Change, err error)` as today; `fileHeads(tx, bridges)` | The map row; nil is the only value left | A variadic or options struct — nothing needs it |
| D-3 | `ConnectionState.off` is deleted. `filedBridges` maps by ends only: `.endSkipped`, `.selfLink`, `.files` in that order, as today without the `bridgeOff` branch | No caller can produce it | Keeping it for S11-27 — S11-27 has no switched-off state |
| D-4 | Every connection row in the panel renders the existing caption branch (`Text(L10n.Promote.connectionState(bridge.state))`, caption, italic, `PVColor.textMuted`). The `PVSwitch` branch and `bridgeOn(_:)` are deleted. The accessibility identifier `promote.connection.<id>` moves to the row's `HStack` | One rendering for every state; the panel is replaced in S11-27, so no new styling is added | Keeping a disabled switch — a control that does nothing |
| D-5 | The `.files` caption keeps key `promote.connections.filed` and value "Filed"; its comment becomes "A connection that files on Done" | Shortest correct change; S11-27 rewrites the wording under S11-D1 | New wording now — it would be redone in S11-27 |
| D-6 | `connectionsMeta(filed:unfiled:)`; key `promote.connections.meta` keeps its name, value "%1$lld filed · %2$lld not filed", comment "Connections list summary; 1 = filed, 2 = not filed". `unfiled` counts `.endSkipped` and `.selfLink` as today | Drops the dead count | A new key — English only, `extractionState: manual`, so the key can change in place |
| D-7 | `connectionsHint` keeps key `promote.connections.hint`, value "A connection files once both its ends are filed; to dispute one, change it on the Evidence graph" | The hint described the switch; this says the rule and the way back ([connection subrows](../promote-connection-subrows.md) "The idea", last paragraph) | Removing the hint — the panel would then not say why a row is not filed |
| D-8 | `leaveDetail(rows:)`; key `promote.leave.pageMessage`, value "You chose targets on %lld rows. Nothing has been filed; opening Promote again proposes the graph afresh.", comment "Leave guard body; argument is how many rows changed" | The connections count is always 0 now | Keeping a second argument |
| D-9 | `FakeStore.lastPromoteBatch` becomes `[CatalogPromoteBatchRow]?`; `PromoteModelTests` reads `store.lastPromoteBatch ?? []` and `store.lastPromoteBatch == nil` | A one-field tuple is not valid Swift | A new name — the two test sites change either way |
| D-10 | `TestSaveBatchSkipsSwitchedOffBridge` is replaced by `TestSaveBatchLeavesBridgeWithASkippedEndUnfiled` (person New, event Skip, participation between them → 0 participation claims). In `TestSaveBatchFilesBridgeWhoseEndsWereAlreadyPromoted` the trailing switch comment and `_ = bridge` are deleted, and `bridge :=` becomes a bare `f.participation(person, event)` call | Keeps the one negative case Done still has (AC-1.2) | Deleting the test only — no test would cover a skipped end |
| D-11 | `doneSendsPinsAndASwitchedOffBridge` becomes `doneSendsDefaultPinsAndConnectionsFollowEnds`: rows `a` and `b`, bridge `br` between them; `a` New, `b` Skip → `connectionLines().first?.state == .endSkipped`; `b` set to New → `.files`; the existing pin assertions stay on row `a`. `promoteAllStartsExpandedAndHidesFiledBridges` drops its `skipBridgeIDs()` line | Swift proof of AC-1.1 / AC-1.2 at the flow layer | A view test — there is no UI-test target |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `TestSaveBatchLeavesBridgeWithASkippedEndUnfiled` *(new)* | test | `core/database/promote/batch_test.go` `TestSaveBatchSkipsSwitchedOffBridge` | Replaces it; reuses `newBatchFixture`, `bare`, `participation`, `save`, `bridgeClaims` |
| `doneSendsDefaultPinsAndConnectionsFollowEnds` *(new)* | test | `PromoteFlowTests.swift` `doneSendsPinsAndASwitchedOffBridge` | Renamed and rewritten in place; reuses `proposal`, `row`, `fact`, `setTarget` |

**Builds on.** Proto `reserved` convention; S11-25's one-PR wire removal.
**Deletes.** `Batch.SkipBridgeIDs`, the `skip` parameters and loop, proto field 6, `skipBridgeIds` in generated code, the `skipBridgeIDs:` store parameter in four conformers, `ConnectionState.off`, `bridgeOff`, `toggleBridge` (flow and model), `skipBridgeIDs()`, `bridgeOn(_:)`, the `PVSwitch` branch, `L10n.Promote.connectionState` `.off` case and `promote.connections.off`, the "switched off" doc rules, `TestSaveBatchSkipsSwitchedOffBridge`.
**Next adds.** S11-27 adds `connections(hostedBy:)` over `filedBridges` with three states; S12-09 makes `SaveBatch` per graph with one fewer field to carry.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/identityclaims` | D-2 signatures; doc comment of `FileBridgesTx` loses its sentence "skip names bridge Subjects the caller switched off; nil files every bridge."; `FileSourceBridgesTx` loses "skip names bridge Subjects to leave unfiled." |
| `core/database/identityclaims` (`identityclaims.go`), `core/database/effects` (`conclusion_handles_test.go`) | Drop the `nil` argument |
| `core/database/promote` | `Batch` loses `SkipBridgeIDs` and its doc sentence; `SaveBatch` loses the `skip` map and loop; its doc reads "claims, pin sets, and bridges whose ends are both handles" |

Errors: unchanged. The `ErrInvalid` for a malformed skip id goes with the field.

### FFI

`api/proto/engine.proto` `ApplyPromoteGraphAlignmentRequest`:

```proto
message ApplyPromoteGraphAlignmentRequest {
  string project_dir = 1;
  string user_id = 2;
  string source_id = 3;
  int64 seen_revision = 4;
  repeated ApplyPromoteGraphAlignmentRow rows = 5;
  reserved 6;
  reserved "skip_bridge_ids";
}
```

Handler `api/ffi/handlers/promote_batch.go`: the `parseIDs(req.GetSkipBridgeIds())` block and the `SkipBridgeIDs:` field are deleted (STEP-1). Regenerate with `./scripts/generate-proto.sh` (needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`); commit both generated files.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Connection row state | existing caption `Text` | `PromoteView.connectionRow(_:)` | D-4; the combine comment reads "One element: the connection's sentence, then whether it files." |
| Panel meta | existing `Text(verbatim:)` | `PromoteView.connectionMeta` | D-6 |
| Leave guard | `.pvConfirm` | `PromoteView` body | `L10n.Promote.leaveDetail(rows:)` (D-8) |

States: unchanged. Accessibility: the row stays one combined element; it reads the sentence then the caption ("Filed", "Not filed · an end is skipped", "Not filed · both ends are the same record").

### Strings

Changed values (keys kept): `promote.connections.meta` (D-6), `promote.connections.hint` (D-7), `promote.leave.pageMessage` (D-8); comment only: `promote.connections.filed` (D-5). Removed: `promote.connections.off`. Each change is made in `L10n.swift` and `Localizable.xcstrings` together.

## 7. Steps

**STEP-1 — Filing takes no skip list**
Files: `core/database/identityclaims/bridgefile.go`, `core/database/identityclaims/identityclaims.go`, `core/database/effects/conclusion_handles_test.go`, `core/database/promote/batch.go`, `core/database/promote/batch_test.go`.
Do: D-2; the Go core table; D-10 (T-1). In `api/ffi/handlers/promote_batch.go`, delete the `parseIDs(req.GetSkipBridgeIds())` block and the `SkipBridgeIDs:` field in this step so the tree builds; the proto field is unread until STEP-2 removes it.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promote/... ./core/database/identityclaims/... ./core/database/effects/... ./api/...` passes.

**STEP-2 — Reserve the wire field**
Files: `api/proto/engine.proto`, `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: the FFI sketch; `./scripts/generate-proto.sh`.
Check: `rg -n "skip_bridge_ids|SkipBridgeIds|skipBridgeIds" api macos/App` matches only the `reserved` line; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-3 — Store takes no skip list**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`, `macos/App/Features/Promote/PromoteModel.swift` (Done call).
Do: remove the `skipBridgeIDs:` parameter and `req.skipBridgeIds` line; D-9.
Check: `xcodebuild test` (the `scripts/check-pr.sh` command) passes; `PromoteFlow.skipBridgeIDs()` still exists and is now unused until STEP-4.

**STEP-4 — Flow states come from ends only**
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `PromoteModel.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`. Skill: `add-swift-test`.
Do: D-3; delete `bridgeOff`, `toggleBridge` (flow and model), `skipBridgeIDs()`; D-11 (T-2, T-3).
Check: `rg -n "bridgeOff|toggleBridge|skipBridgeIDs|\.off\b" macos/App/Features/Promote macos/ProvenenciaTests` is empty.

**STEP-5 — Panel and copy**
Files: `macos/App/Features/Promote/PromoteView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-4 – D-8.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test` (the `scripts/check-pr.sh` command) passes.

**STEP-6 — Docs**
Files: `docs/promote-graph-alignment.md`, `docs/conclusion-layer-data-model.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: the §3 bullet "Bridges aren't rows" reads "**Bridges aren't rows** (participation, relationship, location). A summary line reads \"22 connections will be filed\". It expands to the list of connections, each saying whether it files. Filing rules: §9.1." §9 bullet reads "**Bridges** are filed by the rules in §9.1." §5.4: replace the switched-off bullet with "**The Promote page has no per-bridge switch.** A connection the researcher disputes is fixed on the Evidence graph (delete it, or change its type or role)." Then the Definition of done lines.
Check: `rg -n -i "switched off|switched-off" docs/*.md` is empty; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. Suggested: "File every connection whose ends are both handles on Done." (STEPS 1–4), "Show each connection's state instead of a switch on Promote." (STEP-5), "Drop the connection switch-off rule from the Promote docs." (STEP-6). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/promote/batch_test.go` `TestSaveBatchLeavesBridgeWithASkippedEndUnfiled` *(new)* | person New, event Skip, one participation → `bridgeClaims("participation") == 0` | AC-1.2, D-10 |
| T-2 | Go | `core/database/promote/batch_test.go` `TestSaveBatchFilesBothEndsOnce` (existing) | person and event New → one participation claim | AC-1.1 |
| T-3 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `doneSendsDefaultPinsAndConnectionsFollowEnds` *(new name)* | `a` New, `b` Skip → `.endSkipped`; `b` New → `.files`; `batchRows()` row `a` has `defaultPins` and no `pinIDs` | AC-1.1, AC-1.2, D-3, D-11 |
| T-4 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `promoteAllStartsExpandedAndHidesFiledBridges` (existing, one line removed) | already-filed bridge hidden; new one `.files` | D-11 |
| T-5 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` (existing sites using `lastPromoteBatch`) | Done records the rows | D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/effects` (`conclusion_handles_test.go` signature), `api/ffi/handlers` `promote_batch_test.go`, `PromoteFlowTests`, `PromoteModelTests`, `L10nResolutionTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: AC-1.1 T-2/T-3; AC-1.2 T-1/T-3; AC-1.3, AC-1.4, AC-3.1 manual check; AC-2.1 rg output; AC-2.2, AC-2.3 diff
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; promote.connections.off removed from both
- [ ] UI: no new component; the switch is gone from the connections panel
- [ ] Version bump — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief — none (not gated)
- [ ] Domain docs: promote-graph-alignment.md §3 and §9, conclusion-layer-data-model.md §5.4 (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source holding person John, event Birth of John, and a participation joining them, none promoted. Open Promote on John and choose "Map the rest of this graph".
2. Set both rows to New. Expand the connections panel. The John–Birth row shows "Filed" and no switch (AC-1.3); the header reads "1 filed · 0 not filed" (AC-1.4).
3. Set Birth to Skip. The row shows "Not filed · an end is skipped"; the header reads "0 filed · 1 not filed".
4. Set Birth back to New, change one more target, and navigate away. The leave guard reads "You chose targets on 2 rows. Nothing has been filed; opening Promote again proposes the graph afresh." (AC-3.1). Cancel, press Done. The canonical John is linked to the Birth.
5. Run `rg -n "SkipBridge|skip_bridge|skipBridge|bridgeOff|toggleBridge" api core macos/App macos/ProvenenciaTests`: only the `reserved "skip_bridge_ids";` line in `api/proto/engine.proto` matches (AC-2.1, AC-2.3).
6. Run `rg -n "func FileBridgesTx|func FileSourceBridgesTx" core`: each takes two parameters, `tx *sql.Tx` and one id `[]byte` (AC-2.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `protoc` plug-in versions differ from those that generated the committed files, so regeneration rewrites unrelated lines | Stop condition below |
| A caller of `FileBridgesTx` / `FileSourceBridgesTx` exists that §4 does not list | `go build ./...` fails at STEP-1; stop condition below |

**Stop and report if:**

- `rg -n "FileBridgesTx|FileSourceBridgesTx" --glob '*.go'` finds a caller not listed in §4.
- `./scripts/generate-proto.sh` changes lines in `engine.pb.go` or `engine.pb.swift` other than the `skip_bridge_ids` field and the file's header version comment.
- A Go or Swift test other than those named in §8 fails because a bridge is now filed (that test relied on the switch).
- `scripts/check-localizable-xcstrings.py` rejects changing a key's value or placeholders in place (D-6, D-8).

## 12. Open questions

None.

## 13. PR

**Title:** S11-26 — Connections file by their ends
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-26-connections-by-ends.md`.
