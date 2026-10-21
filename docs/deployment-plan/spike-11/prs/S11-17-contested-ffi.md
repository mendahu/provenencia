# S11-17 — Contested flag through FFI

**Status:** Landed in [#504](https://github.com/mendahu/provenencia/pull/504) (2026-10-19)

**Landed 2026-10-19.** Differences from the plan: the branch is `s11-17-contested-ffi-impl`, because `s11-17-contested-ffi` is the stale branch of the merged plan PR #502 (user decision). Nothing else.
**Spike / slice:** Spike 11 · Slice 4 — Segments and graph alignment
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the wire and Swift's parsing only; the contested marker and the move-reason copy are S11-18 behind S11-D1)
**Board:** —
**Depends on:** S11-16 (landed [#501](https://github.com/mendahu/provenencia/pull/501); [plan](S11-16-graph-alignment.md)). Its last fix round removed a dead nil guard in `core/graphalign/rows.go` and changed nothing this plan reads. Branch from `origin/main` (fetch first)
**Blocks:** S11-18
**Branch:** `s11-17-contested-ffi`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [Bands](../comparison-layers.md#bands) ("A contested marker, not a cap"), [Layer 6 procedure](../comparison-layers.md#layer-6-procedure) step 5 ("Each move has a reason"), [`docs/matching-model.md`](../../../matching-model.md) (glossary: contested); working rules "One wire vocabulary", "High tolerance for churn"; [S11-16 plan](S11-16-graph-alignment.md) D-9 (move reasons) and its **Next adds**
**Planned:** 2026-10-19 against `main` @ `ed45eba` (S11-16 merged); drafted the same day against S11-16's branch tip `920fa2a` and refreshed
**Review:** [`S11-17-contested-ffi.review.md`](S11-17-contested-ffi.review.md) — Fits / Ready with changes, 2 rounds (findings fixed); refreshed against `main` @ `ed45eba` after S11-16 merged

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

S11-16 gave every Promote row a contested flag (a comparison behind the reported pair conflicts, whatever the band) and, when settling released an earlier pair, a move with its reason. The Promote handler drops both, so the page cannot show either. This PR sends `Flags.Contested` and `Row.Move` over `engine.proto` and parses them into Swift's proposal row and `PromoteFlow.Row`. It also moves Promote's reason codes out of `PromoteFlow.swift` into `Platform/ConclusionVocabulary.swift`, where `api/ffi/swift_vocabulary_test.go` fails on a drift between Go and Swift (working rule "One wire vocabulary").

**Why now.** S11-16 built the flag and the move. S11-18 draws the contested marker and words the move reasons in the existing "Updated · <reason>" line, and it needs both on `PromoteFlow.Row`. Dogfood bars 3 ("updated because…") and 4 ("a conflicting row shows contested") depend on that pair of PRs.

**What stays.** The page looks and reads exactly as before: no view, string or L10n change. `Row.reason` keeps its meaning and wording. Cause codes stay off the wire (S11-25).

## 2. User stories

**US-1.** As the app, a proposal row says whether it is contested, so S11-18 can mark it at any band.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A filed handle "Mary Smith", sex at birth `male` (positive). A new Source Subject "Mary Smith" with a **negative** `male` sex-at-birth Observation. `ProposePromoteGraphAlignment` on that Source, no fixed rows | The Subject's row: `target` `skip`, `reason` `weak`, `contested` true |
| AC-1.2 | AC-1.1, with the incoming Subject fixed on the handle | `target` `handle`, `reason` `decided`, `contested` true |
| AC-1.3 | AC-1.1 with a **positive** `male` Observation | `target` `handle`, `reason` `agrees`, `contested` false |

**US-2.** As the app, a row whose earlier pair was released carries the move's reason and the neighbor that caused it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `proposalProto` on a `graphalign.Row` with `Move{Reason: ReasonNeighborPlaced, NeighborSubjectID: <uuid N>}` | `move_reason` `neighbor_placed`, `move_neighbor_subject_id` the canonical string of N |
| AC-2.2 | `proposalProto` on a row with `Move` nil | `move_reason` `""`, `move_neighbor_subject_id` `""` |

**US-3.** As the app, `PromoteFlow` keeps the contested flag and the move on each row, fresh after every proposal.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `load` a proposal with row `a` (`contested` true, `moveReason` `.neighborMoved`, `moveNeighborSubjectID` `b`) and row `b` (clean) | `a.contested` true, `a.move` `.neighborMoved(neighborID: "b")`; `b.contested` false, `b.move` nil |
| AC-3.2 | Row `a` decided on `e1`; `merge` a proposal where `a` is `contested` true | `a` keeps target `e1` and `decided`; `a.contested` true |
| AC-3.3 | A filed (anchor) Subject whose proposal row is `contested` true | The anchor row: `contested` false, `move` nil |
| AC-3.4 | `moveReason` nil with `moveNeighborSubjectID` `b`; then `moveReason` `.neighborReleased` with `moveNeighborSubjectID` `""`; then `moveReason` `.unrecognized("later_move")` with `moveNeighborSubjectID` `b` | `move` nil; `move` `.empty`; `move` `.empty` |

**US-4.** As a contributor, a reason code renamed on one side fails a Go test.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `TestSwiftVocabularyMatchesGo` reads `enum PromoteReasonCode` in `ConclusionVocabulary.swift` | Its wire cases equal the ten `graphalign.Reason` constants: `via`, `decided`, `agrees`, `weak`, `taken`, `no_match`, `empty`, `neighbor_placed`, `neighbor_moved`, `neighbor_released` |
| AC-4.2 | `PromoteReasonCode(wire: "later_code")`; `PromoteReasonCode(wire: "no_match")` | `.unrecognized("later_code")`; `.noMatch` |
| AC-4.3 | `PromoteFlow.Reason(code: c, …)` for each code the page already knew, for `.neighborPlaced` with neighbor `n` and with neighbor `""`, and for `.unrecognized("later_code")` | The same case as today (`.via` with an empty neighbor → `.empty`); `.neighborPlaced(neighborID: "n")`, `.empty`; `.empty` |
| AC-4.4 | The two searches in STEP-3's check | Both print nothing |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PromoteGraphAlignmentRow` fields `contested`, `move_reason`, `move_neighbor_subject_id`; regenerated Go and Swift (D-1) | STEP-1 |
| `proposalProto` sends them (D-2) | STEP-2, AC-1.1, AC-1.2, AC-1.3, AC-2.1, AC-2.2 |
| `PromoteReasonCode` *(new)* in `ConclusionVocabulary.swift` (D-3); `PromoteFlow.Reason` gains three move cases and takes a typed code (D-4) | STEP-3, STEP-6, AC-4.2, AC-4.3, AC-4.4 |
| Go drift guard row for `PromoteReasonCode` (D-5) | STEP-4, AC-4.1 |
| `CatalogPromoteGraphAlignmentRow` and `CatalogPromoteGraphAlignmentAlternative` reason fields typed, new row fields, the `GoStore` mapping, `FakeStore` and test literals typed (D-6, D-8) | STEP-3, AC-4.4 |
| `PromoteFlow.Row.contested`, `PromoteFlow.Row.move` *(new)* (D-7) | STEP-5, STEP-6, AC-3.1, AC-3.2, AC-3.3, AC-3.4 |
| Docs the PR makes false (STEP-7) | STEP-7 |

| Out | Home |
| --- | --- |
| The contested marker beside the band, its copy and accessibility label | S11-18 (behind S11-D1) |
| Wording move reasons (`PromoteModel.reasonText` cases beyond the interim of D-4), and deciding where a move shows: a move is settled within one proposal, while `Row.updated` is a change between proposals | S11-18 with S11-D1 (user decision 2026-10-19; its scope entry says so) |
| Value and segment cause codes (`valuecompare.Cause`, `segmentcompare.Cause`) on the wire and in Swift | S11-25, with the evidence lines that word them (user decision 2026-10-19; its scope entry says so). The S11-14 plan's **Next adds** line naming S11-17 is superseded |
| A contested flag on alternatives (`graphalign.Alternative` has none) | Not planned — the marker is per row (comparison-layers §Bands) |
| `FakeStore` producing contested rows or moves | Not planned — S11-18 sets the fields on fake rows in its own tests when it needs them |
| Moving Promote's `target`, `assessment` and comparison `outcome` codes into `ConclusionVocabulary.swift` (`PromoteFlow.swift` literals `"handle"`, `"new"`, `"skip"`, `"conflict"`; `PromoteFlow.Assessment`) | S12-09 (Promote per graph rewrites the same Promote code; maintainability review M-5) |
| `ConflictWithFixed` changes | Not this PR — S11-16 kept it |

## 4. Current state

Verified at `ed45eba`.

| Area | Where | Today |
| --- | --- | --- |
| Flag | `core/graphalign/types.go` `RowFlags.Contested` | Set by `rows.go` `(*state).proposal` from `(*cand).contested()` on the reported candidate: the placed pair, or the best non-assigned candidate of a Skip row. Never sent |
| Move | `core/graphalign/types.go` `Move{Reason Reason; NeighborSubjectID []byte}`, `Row.Move *Move` | Set when `releasedFrom[s]` exists and the final target is not that handle. Never sent |
| Reason codes | `core/graphalign/types.go` `type Reason string`; `ReasonVia` … `ReasonEmpty` (7 row reasons); `ReasonNeighborPlaced`, `ReasonNeighborMoved`, `ReasonNeighborReleased` (move reasons, "appear only in Move.Reason") | No list function; tests name the constants |
| Proto | `api/proto/engine.proto` `PromoteGraphAlignmentRow` (fields 1–22, `reserved 8`) | No contested or move field. Highest field number 22 |
| Handler | `api/ffi/handlers/promote_align.go` `proposalProto(c, prop)` | Builds each `engine.PromoteGraphAlignmentRow` from `graphalign.Row`; copies `Flags.ConflictWithFixed`, `PossibleDuplicate`, `DuplicateOf`, `Via`; drops `Flags.Contested` and `Move` |
| Handler tests | `api/ffi/handlers/promote_align_test.go` `personOnNewSource`, `TestProposeNegativeConflictLine`, `TestProposePromoteGraphAlignment`; `subjects_test.go` `subjectFixture`; `promote_test.go` `promoteOK`; `harness_test.go` `runRPC`, `marshalProto` | `personOnNewSource(…, "male", PolarityNegative)` against a filed positive `male` handle yields a conflicting sex-at-birth comparison. Probe at planning: unfixed → skip / weak / Contested true; fixed → handle / decided / Contested true; positive → handle / agrees / Contested false. A `Move` cannot be produced from a handler fixture cheaply; `core/graphalign/align_test.go` `TestSettlingMovesAPairAndSaysWhy` covers it in the engine |
| Swift wire row | `macos/App/Platform/GenealogyStore.swift` `CatalogPromoteGraphAlignmentRow`, `CatalogPromoteGraphAlignmentAlternative` | `reason`, `viaNeighborSubjectID`, … as `String` with `""` defaults; no contested or move |
| Swift mapping | `macos/App/Platform/GoStore.swift` `proposePromoteGraphAlignment` (the `resp.rows.map` at about L853) | Maps every row field by name; nothing to map for contested or move |
| Swift fake | `macos/App/Platform/FakeStore.swift` `proposePromoteGraphAlignment` (about L1515–L1575) | Builds rows with `reason: "decided"` / `"empty"`; relies on default arguments for fields it does not set |
| Reason parsing | `macos/App/Features/Promote/PromoteFlow.swift` `PromoteFlow.Reason.init(code:propertyKey:propertyOrigin:viaNeighborID:)` | `switch code` on string literals `"via"` (only with a neighbor), `"decided"`, `"agrees"`, `"weak"`, `"taken"`, `"no_match"`; `default: .empty`. `.filed` is Swift-only (anchor rows), never on the wire |
| Reason wording | `macos/App/Features/Promote/PromoteModel.swift` `reasonText(_:on:)` | Exhaustive `switch` over `PromoteFlow.Reason`; `.via` resolves the neighbor's name, ref and bridge phrase. `PromoteView.swift` shows `L10n.Promote.updatedDetail(model.reasonText(for: row))` only when `row.updated` |
| Test literals | `PromoteFlowTests.swift` (the `row(…)` helper and about six tests), `PromoteModelTests.swift` (helper, `reasonsAreWordedWithLabelsAndRefs`) | Set `reason` on wire rows and alternatives as strings (`"agrees"`, `"taken"`, `"weak"`, `"decided"`, `"via"`) |
| Flow row | `PromoteFlow.swift` `PromoteFlow.Row`; `makeRow(_:fact:previous:orders:)` | Three construction paths: anchor (`fact.anchor`), decided keep (`previous.decided`, refreshes `assessment`, `reason`, `conflictNote`, `proposedDuplicateOf`), fresh. No contested or move |
| Vocabulary | `macos/App/Platform/ConclusionVocabulary.swift` | `ReconciledState`, `ReconcilerReason` (with `case unrecognized(String)` and `init(wire:)`), seeded keys, `DateQualifier`. Header: "Spell these values only here" |
| Drift guard | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo`, regex `wireCase` (`case "([a-z_]*)": self = `), subtest "states and reasons tolerate a value this build doesn't know" | Rows for field states and outcome reasons; none for Promote reasons |
| Swift tests | `macos/ProvenenciaTests/PromoteFlowTests.swift` helpers `proposal(rows:revision:)`, `row(id:target:handleID:assessment:pinned:)`, `fact(_:anchor:handleID:)`; `decidedRowsSurviveAReproposal`; `PromoteModelTests.swift` `reasonsAreWordedWithLabelsAndRefs` | Pass reasons as wire strings into `CatalogPromoteGraphAlignmentRow` (correct: that struct is the wire shape) |
| Docs | `docs/deployment-plan/spike-11/comparison-layers.md` §Layer 6 "Built in S11-16" paragraph | "…the wire carries neither until S11-17." |

**Precedent.** `edb419f1` ("promote: typed reasons and labels instead of English and keys") put `reason` on `PromoteGraphAlignmentRow` and carried it to Swift: proto field, `proposalProto` line, `CatalogPromoteGraphAlignmentRow` field, `GoStore` mapping, `PromoteFlow.Reason`. Copy that path field by field, except where the code is parsed. For that, copy the reconciler: `ReconcilerReason` in `ConclusionVocabulary.swift`, typed fields `CatalogReconciledValue.reason` and `CatalogReconcilerOutcome.reason` in `GenealogyStore.swift`, and `ReconcilerReason(wire:)` called in `GoStore.swift`. For the guard row, copy the "outcome reasons" row of `TestSwiftVocabularyMatchesGo`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Wire.** `PromoteGraphAlignmentRow` gains `bool contested = 23;` ("a comparison behind the pair this row reports conflicts, whatever its band; the page marks it"), `string move_reason = 24;` ("a graphalign.Reason move code, set when settling released an earlier pair of this row; empty when no move"), `string move_neighbor_subject_id = 25;` ("with move_reason: the layer neighbor that caused it"). No new message | Matches the row's flat style (`via_neighbor_subject_id`, `via_bridge_type`, `duplicate_of_subject_id`); an empty reason is "no move", as an empty `via_neighbor_subject_id` is "not via" | A `PromoteGraphAlignmentMove` message — a second shape for the same kind of fact on one row; reusing `reason` (field 20) for a move — loses `via` / `agrees`, which the page words (S11-16 D-9) |
| D-2 | **Handler.** In `proposalProto`, the row literal gains `Contested: r.Flags.Contested`; after the `r.Via` block, `if r.Move != nil { row.MoveReason = string(r.Move.Reason); row.MoveNeighborSubjectId = database.UUIDString(r.Move.NeighborSubjectID) }` | One place builds rows; same shape as the `Via` copy, which uses `database.UUIDString` (`core/database/uuid.go`; it replaced the handlers' `uuidString` on `main`) | A helper function — two lines |
| D-3 | **Swift wire enum.** `enum PromoteReasonCode: Sendable, Equatable` *(new)* in `ConclusionVocabulary.swift`, after `ReconcilerReason`, doc "Why a Promote row reads as it does, and why settling moved it (`graphalign.Reason`). The page words the typed cases." Cases in Go's order: `via`, `decided`, `agrees`, `weak`, `taken`, `noMatch`, `empty`, `neighborPlaced`, `neighborMoved`, `neighborReleased`, `unrecognized(String)` (doc "A code this build doesn't know"). `init(wire:)` with one `case "<code>": self = .<case>` line per code and `default: self = .unrecognized(wire)`. One enum for both row and move reasons | Go has one `Reason` type whose move reasons "extend Reason" (S11-16 D-9, working rule "One wire vocabulary"); one Swift enum mirrors one Go type and one guard row checks it. The `case "x": self = ` shape is what `wireCase` reads | Two Swift enums (row reasons, move reasons) — two guard rows for one Go type; a `String` raw-value enum read by `listCase`, the shape `CatalogList` uses (`macos/App/Features/Workspace/Session/CatalogMutation.swift`) — `init?(rawValue:)` drops a code this build does not know instead of keeping it as `.unrecognized`, which the "tolerate a value this build doesn't know" subtest requires of reasons; `ReconcilerReason` is the shape for reasons |
| D-4 | **PromoteFlow.Reason takes the typed code and gains move cases.** `PromoteFlow.Reason` gains `.neighborPlaced(neighborID: String)`, `.neighborMoved(neighborID: String)`, `.neighborReleased(neighborID: String)` after `.via`. `init(code:propertyKey:propertyOrigin:viaNeighborID:)` becomes `init(code: PromoteReasonCode, propertyKey: String, propertyOrigin: String, neighborID: String)`: `.via` → `.via(neighborID:)` when `neighborID` is not empty, else `.empty`; `.decided`, `.agrees`, `.weak`, `.taken`, `.noMatch` → their cases as today; the three move codes → the matching move case with `neighborID`. One rule for all four neighbor cases: an empty `neighborID` gives `.empty`, as `.via` does today; `.empty` and `.unrecognized` → `.empty`. The row and alternative convenience `init`s pass `viaNeighborSubjectID` as `neighborID`. `PromoteModel.reasonText(_:on:)` gets one interim arm, `case .neighborPlaced, .neighborMoved, .neighborReleased: return L10n.string(L10n.Promote.reasonEmpty)`, with the comment "Worded in S11-18; no row reason is a move reason." `.filed` stays Swift-only | One reason type and one wording switch on the page: S11-18 replaces one interim arm, the compiler lists every move case, and the move cases reuse `.via`'s neighbor naming (maintainability review M-1). The interim arm is unreachable: Go sends move codes only in `move_reason` | A separate `PromoteFlow.Move` carrying `PromoteReasonCode` — S11-18 would need a second switch with a `default` and a copy of the neighbor lookup; new L10n strings now — copy is S11-18 (Out) |
| D-5 | **Guard.** `TestSwiftVocabularyMatchesGo` gains a row `name: "promote reasons", path: swiftVocabulary, enum: "PromoteReasonCode", re: wireCase` whose `want` is `sorted(...)` of the ten `string(graphalign.Reason…)` constants, written out like the "outcome reasons" row. The "tolerate a value this build doesn't know" subtest's list gains `"PromoteReasonCode"` | Same guard as `ReconcilerReason`; `graphalign` has no list function and adding one only for a test is new API | An `AllReasons()` in `graphalign` — production code for a test; reading the Go constants by reflection — the house test lists constants |
| D-6 | **Swift wire row parses in the store.** On `CatalogPromoteGraphAlignmentRow`, `reason` becomes `PromoteReasonCode = .empty`; new after `viaRole`: `var contested: Bool = false` (doc "A comparison behind the reported pair conflicts, whatever the band"), `var moveReason: PromoteReasonCode? = nil` (doc "A move reason when settling released an earlier pair of this row in this proposal; nil otherwise"), `var moveNeighborSubjectID: String = ""`. On `CatalogPromoteGraphAlignmentAlternative`, `reason` becomes `PromoteReasonCode = .empty`, and its doc comment "via, agrees, or weak. The page words it." becomes "A row reason, worded by the page." The doc comment on `reason` becomes "A row reason, worded by the page." `GoStore` maps `reason: PromoteReasonCode(wire: r.reason)`, `contested: r.contested`, `moveReason: r.moveReason.isEmpty ? nil : PromoteReasonCode(wire: r.moveReason)`, `moveNeighborSubjectID: r.moveNeighborSubjectID` (in struct order: after `viaRole:`, before `person:`), and the alternative's `reason: PromoteReasonCode(wire: a.reason)` | Copies where `ReconcilerReason` is parsed: `CatalogReconciledValue.reason: ReconcilerReason` and `GoStore`'s `ReconcilerReason(wire:)` calls (maintainability review M-2). Wire codes become types in one layer, so no feature or test spells them. Doc comments name the type, not the codes, so a new code leaves no stale comment (M-4) | Strings on the wire row parsed in `PromoteFlow` — a second place codes are parsed, and `FakeStore` and tests keep the literals |
| D-7 | **Flow row.** `PromoteFlow.Row` gains `var contested = false` (doc "A comparison behind this row's pair conflicts, whatever its band; the page marks it") and `var move: Reason? = nil` (doc "Why settling released an earlier pair of this row in this proposal. Not `updated`, which compares with the previous proposal"), declared after `duplicateNote` so the memberwise init keeps working where they are omitted. `PromoteFlow.Reason` gains `init?(move proposal: CatalogPromoteGraphAlignmentRow)`, beside its row and alternative `init`s: nil when `moveReason` is nil, else `self.init(code: moveReason, propertyKey: "", propertyOrigin: "", neighborID: proposal.moveNeighborSubjectID)`. In `makeRow`: the anchor path leaves both at their defaults; the decided-keep path sets `kept.contested = proposal.contested` and `kept.move = Reason(move: proposal)` next to `kept.conflictNote`; the fresh path passes `contested: proposal.contested, move: Reason(move: proposal)` | S11-18 reads `row.contested` and `row.move` and needs nothing else from the wire. Refreshing on the decided path keeps contested true to the latest proposal, as `assessment` is. An anchor row is filed and not proposed. An unrecognized move code becomes `.empty`, as an unrecognized row reason does | Leaving `PromoteFlow` alone until S11-18 — S11-18 would then also change the wire path, which the map gives this PR; tying `move` to `updated` — they measure different things (maintainability review M-3) |
| D-8 | **Literals become typed cases.** `FakeStore.proposePromoteGraphAlignment` writes `reason: .decided` (twice) and `reason: .empty`. `PromoteFlowTests` and `PromoteModelTests` write typed cases where they set `reason` (`.agrees`, `.via`, `.taken`, `.weak`, `.decided`) on wire rows and alternatives | Follows from D-6; afterwards no Promote reason code is spelled outside `ConclusionVocabulary.swift` (AC-4.4) | — |
| D-9 | **Tests.** Go: a new table-driven `TestProposeSendsContested` through `runRPC` (`reqFn` builds the AC-1 fixtures with `personOnNewSource`; `after` finds the incoming Subject's row by id and checks `target`, `reason`, `contested`), and `TestProposalProtoSendsMove` calling `proposalProto` directly inside `withProjectCatalog` on a `subjectFixture` project with a literal `graphalign.Proposal` (rows with `Target` skip and no handle, so no header lookup hits). Swift: new `@Test`s in `PromoteFlowTests.swift`; `PromoteReasonCode` cases tested there too | The handler can produce contested from real data (probe in §4); a move needs a settle that a handler fixture cannot reach cheaply, and the engine already tests it (`TestSettlingMovesAPairAndSaysWhy`), so the handler test covers only the copy | A handler fixture that forces a release — needs several people and links for a two-line copy; a `GoStore` test — no Swift test calls `GoStore` (it needs the dylib) |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PromoteGraphAlignmentRow.contested`, `.move_reason`, `.move_neighbor_subject_id` *(new)* | proto fields | `PromoteGraphAlignmentRow` `conflict_with_fixed`, `via_neighbor_subject_id` | Extends that message (D-1) |
| `PromoteReasonCode` *(new)* | Swift enum | `ConclusionVocabulary.swift` `ReconcilerReason`; `PromoteFlow.Reason` | `ReconcilerReason` is the reconciler's vocabulary; `PromoteFlow.Reason` carries payloads and a Swift-only case, so it stays the page's type and `PromoteReasonCode` is parse-only, as `ReconcilerReason` is (D-3, D-4). Searched `ConclusionVocabulary.swift`, `rg 'enum .*Reason' macos/App` |
| `PromoteFlow.Row.contested`, `.move` *(new)* | Swift fields | `Row.conflictNote`, `Row.updated` | Different facts: `conflictNote` is `ConflictWithFixed`, `updated` is Swift's own target diff |
| `PromoteFlow.Reason.neighborPlaced`, `.neighborMoved`, `.neighborReleased` *(new)* | Swift enum cases | `PromoteFlow.Reason.via(neighborID:)` | Extends the page's one reason type (D-4) |
| `PromoteFlow.Reason.init?(move:)` *(new)* | Swift init | `PromoteFlow.Reason.init(_ row:)`, `init(_ alt:)` | Every wire-row-to-`Reason` conversion lives on `Reason` (maintainability review round 2, M-7) |
| `TestProposeSendsContested`, `TestProposalProtoSendsMove` *(new)* | Go tests | `TestProposeNegativeConflictLine`, `TestProposePromoteGraphAlignment` | Reuse their helpers; new assertions |

**Builds on.** `add-ffi-handler` (proto → regenerate → handler → `runRPC` test); the wire-vocabulary guard in `swift_vocabulary_test.go`; precedent `edb419f1`.
**Deletes.** Every Promote reason string literal in `macos/App` and `macos/ProvenenciaTests` outside `ConclusionVocabulary.swift`: the six in `PromoteFlow.Reason.init`, three in `FakeStore`, about ten in `PromoteFlowTests` and `PromoteModelTests`. The sentence "the wire carries neither until S11-17" in `comparison-layers.md`.
**Next adds.** S11-18 reads `row.contested` for the marker and replaces D-4's interim `reasonText` arm with the three move sentences, naming the neighbor as `.via` does. A new graphalign reason is one Go constant, one `PromoteReasonCode` case and line, one entry in the D-5 row, and one `PromoteFlow.Reason` case the compiler then asks `reasonText` for. S11-25 adds cause codes to `ConclusionVocabulary.swift` and parses them in `GoStore` the same way.

### Schema

None — no catalog change.

### Go core

None — `core/graphalign` already sets both fields (S11-16).

### FFI

`api/proto/engine.proto`, message `PromoteGraphAlignmentRow`, after `reason_property_origin = 22`:

```proto
  // A comparison behind the pair this row reports conflicts, whatever its
  // band. The page marks it.
  bool contested = 23;
  // A graphalign.Reason move code, set when settling released an earlier
  // pair of this row. Empty when no move.
  string move_reason = 24;
  // With move_reason: the layer neighbor that caused the move.
  string move_neighbor_subject_id = 25;
```

Regenerate with `./scripts/generate-proto.sh` (writes `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`; commit both). Handler `api/ffi/handlers/promote_align.go` `proposalProto` per D-2. No new method; `dispatch.go` unchanged.

### macOS

No view change. Model and platform changes only:

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Wire vocabulary | `PromoteReasonCode` *(new)* | `macos/App/Platform/ConclusionVocabulary.swift` | D-3 |
| Wire row | `CatalogPromoteGraphAlignmentRow`, `CatalogPromoteGraphAlignmentAlternative` | `macos/App/Platform/GenealogyStore.swift` | D-6 |
| Mapping | `GoStore.proposePromoteGraphAlignment`; `FakeStore.proposePromoteGraphAlignment` | `macos/App/Platform/GoStore.swift`, `FakeStore.swift` | D-6, D-8 |
| Flow | `PromoteFlow.Reason`, `PromoteFlow.Row`; `PromoteModel.reasonText` interim arm | `macos/App/Features/Promote/PromoteFlow.swift`, `PromoteModel.swift` | D-4, D-7 |

States: not applicable (no UI). Accessibility: none (S11-18).

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Proto fields**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: add the three fields per §6 FFI. Run `./scripts/generate-proto.sh`.
Check: `git diff --stat` shows the proto and both generated files; `CGO_ENABLED=1 go build -tags fts5 ./api/...` passes.

**STEP-2 — Handler sends contested and move**
Files: `api/ffi/handlers/promote_align.go`, `api/ffi/handlers/promote_align_test.go`. Skill: `add-ffi-handler`; rule `.cursor/rules/go-tests.mdc`.
Do: D-2. Add T-1 and T-2 (D-9). T-1: copy `TestProposeNegativeConflictLine`'s setup (`personOnNewSource` for the handle Subject, `promoteOK`, `personOnNewSource` for the incoming Subject; both are "Mary Smith" because the helper names them so). Each table row's `reqFn` assigns the incoming Subject id to a variable declared in that row's scope, and its `after` reads it. The fixed row sends `Fixed: []*engine.PromoteGraphAlignmentFixed{{SubjectId: incoming, HandleId: handle.Entity.GetId()}}`. T-2: `a, b, n := uuid.New(), uuid.New(), uuid.New()`; `graphalign.Proposal{Rows: []graphalign.Row{{SubjectID: a[:], Target: graphalign.TargetSkip, Move: &graphalign.Move{Reason: graphalign.ReasonNeighborPlaced, NeighborSubjectID: n[:]}}, {SubjectID: b[:], Target: graphalign.TargetSkip}}}`; expect `n.String()`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/ -run 'TestPropose'` passes.

**STEP-3 — Swift reason codes, typed in the store**
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/App/Features/Promote/PromoteFlow.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`. Skill: `add-swift-test`.
Do: D-3; D-6 (all of it: typed `reason` fields, the three new row fields and the `GoStore` mapping); D-4; D-8. These land together because the typed fields and the typed `Reason.init` do not build apart.
Check: `rg -n 'reason: "|\.reason = "' macos/ProvenenciaTests/Promote*Tests.swift macos/App/Platform/FakeStore.swift` and `rg -n 'case "(via|decided|agrees|weak|taken|no_match)"' macos/App/Features/Promote` both print nothing (AC-4.4; at planning they print exactly D-8's literals and `PromoteFlow.Reason.init`'s six cases); `xcodebuild test` passes `PromoteFlowTests` and `PromoteModelTests` with only D-8's literal changes.

**STEP-4 — Go guard over the Swift enum**
Files: `api/ffi/swift_vocabulary_test.go`.
Do: D-5. Import `github.com/mendahu/provenencia/core/graphalign`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes. Then rename `case "taken"` to `case "took"` in the Swift file, rerun, see it fail with the "promote reasons" subtest, and revert.

**STEP-5 — Flow row fields**
Files: `macos/App/Features/Promote/PromoteFlow.swift`.
Do: D-7.
Check: the app target builds; `xcodebuild test` passes `PromoteFlowTests` and `PromoteModelTests`.

**STEP-6 — Swift tests**
Files: `macos/ProvenenciaTests/PromoteFlowTests.swift`. Skill: `add-swift-test`.
Do: add T-4, T-5 and T-6. For AC-3.2, decide row `a` as `decidedRowsSurviveAReproposal` does: `load`, `mapRest()`, `setTarget(subjectID: "a", token: "handle:e1")`, then `merge`. The `row(…)` helper's signature stays (STEP-3 already typed its `reason`); set `contested`, `moveReason`, `moveNeighborSubjectID` on the returned value in each test.
Check: `xcodebuild test` (the command in `scripts/check-pr.sh`) passes `PromoteFlowTests` and `PromoteModelTests`.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-11/comparison-layers.md`, `completed.md`, `deployment-plan.md` (map row status), this plan (status).
Do: in `comparison-layers.md`, replace "the wire carries neither until S11-17." with "S11-17 sends both on `PromoteGraphAlignmentRow` (`contested`, `move_reason`, `move_neighbor_subject_id`); S11-18 shows them." The S11-17, S11-18 and S11-25 scope entries were updated at planning; do not edit them. The rest per the Definition of done.
Order: open the PR first, then write commit 4 with its number. `completed.md`: a new `## S11-17 — Contested flag through FFI` section at the top, shaped like the S11-16 entry ("**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N). <what shipped>"). Map row status: `Landed #N`. This plan: `Landed in [#N](<PR url>) (YYYY-MM-DD)` plus a **Landed** note.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-17-contested-ffi.md` reports no errors; links resolve.

Commits: four. (1) STEP-1 + STEP-2, "Send each Promote row's contested flag and move reason to the app." (2) STEP-3 + STEP-4, "Parse Promote's reason codes once, in the wire vocabulary, under the Go drift guard." (3) STEP-5 + STEP-6, "Keep each Promote row's contested flag and move in the flow." (4) STEP-7, "Record S11-17 in the docs." The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Send each Promote row's contested flag and move reason to the app.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `api/ffi/handlers/promote_align_test.go` `TestProposeSendsContested` *(new)* | Rows: "negative, unfixed" → skip / weak / contested true; "negative, fixed" → handle / decided / true; "positive" → handle / agrees / false (fixtures per AC-1) | AC-1.1, AC-1.2, AC-1.3, D-2 |
| T-2 | Go | same file, `TestProposalProtoSendsMove` *(new)* | Proposal with row A (`Move{ReasonNeighborPlaced, N}`) and row B (no move) → A `move_reason` `neighbor_placed`, `move_neighbor_subject_id` N's string; B both `""` | AC-2.1, AC-2.2, D-2 |
| T-3 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` (new row "promote reasons"; `PromoteReasonCode` in the unrecognized subtest) | Swift cases = the ten constants; `case unrecognized(String)` present | AC-4.1, D-5 |
| T-4 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aContestedRowAndItsMoveReachTheFlow` *(new)* | AC-3.1 and AC-3.4 inputs → those rows' `contested` and `move` | AC-3.1, AC-3.4, D-7 |
| T-5 | Swift | same, `aDecidedRowTakesTheLatestContestedFlag` *(new)*; `anAnchorRowIsNeverContested` *(new)* | AC-3.2; AC-3.3 | AC-3.2, AC-3.3, D-7 |
| T-6 | Swift | same, `reasonCodesParseThroughTheWireVocabulary` *(new)* | `PromoteReasonCode(wire:)` for all ten codes and `later_code`; `PromoteFlow.Reason(code:…)` per AC-4.3 | AC-4.2, AC-4.3, D-3, D-4 |

AC-4.4 is the `rg` in STEP-3. Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteFlowTests`, `PromoteModelTests.reasonsAreWordedWithLabelsAndRefs` (reason parsing), `TestProposePromoteGraphAlignment` (`want` compares populated fields only, so new fields do not break it).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] No user-facing copy added; Localizable.xcstrings unchanged
- [ ] No UI change
- [ ] Version bump: none (header)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] comparison-layers.md "Built in S11-16" sentence per STEP-7
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Nothing changes on the page; the check is that nothing broke.

1. Open a project with a filed person "Mary Smith" (sex at birth Male). Add a Source with a person "Mary Smith" whose sex at birth is recorded as **not** Male. Open Promote on that Subject. The row reads as before this PR (same band, reason and menu). No new mark appears; S11-18 adds it.
2. In Xcode, set a breakpoint in `PromoteFlow.makeRow` and confirm the row's `proposal.contested` is true.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `main` moves under §4 before the PR starts | Implementer step 2 compares `main` with `ed45eba` |
| `generate-proto.sh` tool versions differ and rewrite unrelated generated code | Diff the generated files; only the three fields' code changes. If more changes, stop |
| A Swift test builds `PromoteFlow.Row` by memberwise init with positional arguments after `duplicateNote` | The new fields are declared last with defaults (D-7) |

**Stop and report if:**

- `protoc`, `protoc-gen-go`, `protoc-gen-swift` or `xcodebuild` is missing. This PR is built on macOS, where `scripts/check-pr.sh` runs; do not hand-edit generated files, and do not report Swift tests as passing without running them.
- `graphalign.RowFlags.Contested`, `graphalign.Row.Move` or the three move reason constants are missing or named differently on `main`.
- `PromoteGraphAlignmentRow` already uses field number 23, 24 or 25.
- T-1's negative fixture does not produce `Contested` true at the handler (the §4 probe no longer holds).
- `TestSwiftVocabularyMatchesGo` cannot read `PromoteReasonCode` with `wireCase` without changing the regex.
- Any existing Swift or Go test needs a behavior change (not a new field) to pass.
- Making a test pass needs a view or L10n change beyond D-4's interim arm (Out).

## 12. Open questions

None.

## 13. PR

**Title:** S11-17 — Contested flag through FFI
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-17-contested-ffi.md`.
