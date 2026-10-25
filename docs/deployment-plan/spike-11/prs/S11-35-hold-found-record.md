# S11-35 — Hold a record Promote did not offer

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 8 — Find another record
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the call has no surface until S11-36 adds the menu entry and the picker; a held record reads through the existing row readout, which already draws band None with `decided` and Contested)
**Board:** —
**Depends on:** —
**Blocks:** S11-36
**Branch:** `s11-35-hold-found-record`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [Find another record § Shape](../promote-find-another-record.md#shape), [§ What the engine already does](../promote-find-another-record.md#what-the-engine-already-does), [§ Where the Swift side changes](../promote-find-another-record.md#where-the-swift-side-changes); working rules "A found record is a decided row, nothing new", "The picked record joins the row's menu", "Engine first, proven by tests"
**Planned:** 2026-10-25 against `main` @ `22e3878`
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

The Promote page can hold a row on a record that Promote never offered: a call on `PromoteFlow` takes the record, puts it in the row's target menu, selects it and marks the row decided, and the page sends it to Go as a fixed pair on every re-proposal. Go already holds, scores, explains and files such a pair; this PR pins that with tests instead of assuming it. It is the engine half of [Find another record](../promote-find-another-record.md); the menu entry and the search that call it are S11-36.

**Why now.** Since S11-05, a record whose first given name or surname differs from the Subject's is never a candidate, so a researcher who knows they are one person cannot file the Subject on it from Promote (the feature doc's "The problem"). S11-36 needs a call to hand its pick to, and S11-35 and S11-37 run in parallel ahead of it.

**What stays.** The target menu, its tokens and `setTarget`; the proposal wire (no new field, flag or reason code); every Go package's code; `FakeStore`'s proposal logic.

## 2. User stories

**US-1.** As a researcher, when I choose a record for a row that Promote did not offer, the row holds it, decided, and the menu shows it as the selection, so that I can go back to an offered record, New or Skip.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Row `a` opened unset with reason `noMatch` and an empty menu; `hold(subjectID: "a", record:)` with id `e7`, ref `PER-7`, title `Jones, Polly`, reading title `Polly Jones`, subtitle `1817 –` | Returns true. Target `.handle(id: "e7", ref: "PER-7", title: "Jones, Polly")`; the menu's last item is `e7` with assessment `.none` and reason `.decided`; `decided` and `manual` are true; `readingTitle` is `Polly Jones`; `selectionReadout` is `.none` / `.decided` |
| AC-1.2 | AC-1.1, then `fixed(for:)` over the decided rows and `batchRows()` (rest mapped) | `fixed` holds `(a, handle e7)`; the batch row for `a` is target `handle`, entity `e7`, `defaultPins` true |
| AC-1.3 | AC-1.1, then a proposal whose row `a` is target `handle`, handle `e7`, assessment `weak`, reason `decided`, contested | The menu still holds `e7`; its assessment is `.weak`; its reason stays `.decided`; the row is contested and its target is `e7` |
| AC-1.4 | AC-1.3, then row `b` is decided with `setTarget` and a proposal arrives whose row `a` is target `new`, assessment `none` | Row `a` keeps target `e7` and `e7` in its menu |
| AC-1.5 | Row `a` with menu `[e1, e2]` on target `e1`; `hold` with id `e2` | The menu is still `[e1, e2]`; `e2` keeps the assessment and reason it was offered with; the target is `e2` |
| AC-1.6 | Row `a` was holding `e1` with evidence and pins; `hold` with id `e7` | `board` is nil, `pins` empty, `contested` false, `move` nil (the `setTarget` reset) |
| AC-1.7 | Row `b` targets `e1`; `hold(subjectID: "a", …)` with id `e1` that `a`'s menu did not hold | Row `a`'s duplicate note is `.sharesHandle(otherName: "b", ref: "PER-1")` |
| AC-1.8 | `hold` on an anchor row, on a subject id that is not a row, or with an empty record id | Returns false; rows unchanged |

**US-2.** As the engine, when the page sends a fixed pair on a handle that no shortlist key reaches, I load it, score it, mark it contested on a conflict, explain it and file it like any decided handle.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Filed `Robins`; Subject `Robnis` on another Source, unfixed | The row neither targets the handle nor lists it among its alternatives |
| AC-2.2 | AC-2.1 with `Fixed{Robnis, handle}`; sex at birth absent / same on both / female against male | Target `handle` on that handle, reason `decided`, `ConflictWithFixed` false; a `name` comparison exists; the `sex_at_birth` comparison is `unknown` / `agree` / `conflict`; `Contested` false / false / true |
| AC-2.3 | Over FFI: filed `Mary Smith` (male); `Polly Jones` (male) on another Source, fixed on that handle | Row target `handle`, the handle's id, reason `decided`, assessment `none`, contested true |
| AC-2.4 | AC-2.3's pair through `GetPromoteClaimEvidence` | No error; a `name` section and a `sex_at_birth` section; at least one offer |
| AC-2.5 | AC-2.3's pair through `ApplyPromoteGraphAlignment` with `DefaultPins` | One written row on the handle's entity id, with pins > 0 |

**US-3.** As a contributor, the fake store holds a found record the way Go does, so the Swift tests and S11-36's previews can drive it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `FakeStore` with no scripted proposals; Ada (`a`) unfiled on graph `g-1`; a person entity `e-far` filed through Subject `x` on another graph; `PromoteModel.hold(subjectID: "a", record:)` with `e-far` | `store.lastPromoteFixed == [(a, handle e-far)]`; after the proposal the row targets `e-far`, the menu holds it with reason `.decided`; `done()` sends a batch row `a` → `handle`, entity `e-far` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PromoteFlow.hold(subjectID:record:)` *(new)*: adds the record to `row.menu` with reason `decided` and band None, then runs `setTarget`'s path | STEP-3, AC-1.1 – AC-1.8 |
| `PromoteModel.hold(subjectID:record:)` *(new)*, with `setTarget`'s re-propose follow-up through one private helper | STEP-4, AC-3.1 |
| Go tests: an off-shortlist fixed handle is loaded, banded, contested on a conflict (`promotealign`), and the same over the FFI for propose, evidence and Apply (`handlers`) | STEP-1, STEP-2, AC-2.1 – AC-2.5 |
| `FakeStore` holds a fixed handle outside its candidates: pinned by a model test, no code change (D-7) | STEP-4, AC-3.1 |
| Feature doc and map scope entry say the engine claims are pinned, and where | STEP-5 |

| Out | Home |
| --- | --- |
| The "Find another record" entry in the target menu, the inline picker, building an `Alternative` from a `SearchHit` | S11-36 |
| The shared search model and hit row | S11-37 |
| How a held record with band None reads (copy, styling) | S11-D3, S11-36 |
| Any new proposal field, flag or reason code | not planned (working rule "A found record is a decided row, nothing new") |
| Go code changes | not planned unless a test in STEP-1 or STEP-2 fails (Stop conditions) |
| A kind check in `hold` or in `FakeStore`'s held row | not planned: S11-36 searches the row's kind only, and Go's `checkFixed` refuses another kind with `promote.invalid` |
| `FakeStore` scripted proposals (`promoteProposals`) honouring `fixed` | not planned: tests script the row they expect |
| Searching from the evidence sheet; records of another kind; merging records | [what stays later](../promote-find-another-record.md#what-stays-later) |

## 4. Current state

Verified at `22e3878`. The Go rows were also checked by running throwaway tests against the planning base (the results are in AC-2.1 – AC-2.5).

| Area | Where | Today |
| --- | --- | --- |
| Menu pick | `macos/App/Features/Promote/PromoteFlow.swift` `PromoteFlow.setTarget(subjectID:token:)` (line 432) | Resolves the token through `target(for:on:)`; on a change clears `board`, `pins`, `contested`, `move`; then sets `decided`, clears `updated`, sets `manual`, runs `noteDuplicates()`, returns true. Refuses anchors and unknown rows with false |
| Token resolution | same file, `target(for:on:)` (line 610) | `handle:<id>` resolves only when `row.menu` holds `<id>`; any other token leaves the target as it was. A record outside the menu has no way in. This is the gap |
| Re-proposal of a decided row | same file, `makeRow` decided branch (lines 639 – 660) | `kept = previous`, so `menu` and `target` survive. When the proposal's row is target `handle` on the held id, the menu item's `assessment` is refreshed, and its `reason` only when the new one is not `.decided` |
| Menu item | same file, `PromoteFlow.Alternative` (line 140) | `id`, `ref`, `title` (sorted), `readingTitle`, `subtitle`, `assessment` (default `.none`), `reason` (default `.empty`) |
| Readout | same file, `Row.selectionReadout` (line 192); `PromoteReadout.swift` preview "Readout badges" (line 42) | Reads the selected menu item's band and reason; the preview already draws `.none` + `.decided` + contested |
| Model | `macos/App/Features/Promote/PromoteModel.swift` `setTarget(subjectID:token:)` (line 173) | Calls the flow; on true clears `staleNote`, takes `nextProposal()` and starts `repropose(generation:)`, which sends `flow.fixed(for: flow.rows.filter(\.decided))` |
| Fake propose | `macos/App/Platform/FakeStore.swift` `proposePromoteGraphAlignment` (line 1640), `buildPromoteProposal` (line 1662) | Scripted `promoteProposals` win and ignore `fixed`. Unscripted, a fixed handle resolves among `membershipBySubject` entities of any graph and comes back target `handle`, `strong`, reason `decided`. The fake has no candidates and every fake entity is a membership's, so any record the fake's search (`handleSearchHits`, also over memberships) can find, it can hold. No code gap |
| Fake Apply | same file, `applyPromoteGraphAlignment` (line 1746) | A `handle` row files onto any membership entity of the Subject's type |
| Fixed check | `core/database/promotealign/promotealign.go` `checkFixed` (line 163) | A handle must be 16 bytes, a known unmerged node, of the Subject's kind. It never checks the handle was a candidate |
| Canon | `core/database/promotealign/canon.go` `loadCanon` (line 36) | Adds every fixed handle before the candidates; fixed Subjects get no shortlist |
| Scoring | `core/graphalign/align.go` `(*state).anchor` (line 184); `core/graphalign/rows.go` `proposal` (line 21) | Scores each fixed handle as the Subject's first candidate; the row reports `ReasonDecided`, `st.band(score)`, comparisons and `Contested` |
| Evidence | `core/database/promotealign/evidence.go` `Evidence` (line 40) | Validates the pair with `checkFixed`; any handle of the right kind |
| Apply | `api/ffi/handlers/promote_batch.go` `ApplyPromoteGraphAlignment`; `promotealign.ResolvePins`; `core/database/promote/batch.go` `SaveBatch` | A `handle` row files on any 16-byte entity of the Subject's type |
| Go tests | `core/database/promotealign/promotealign_test.go` `TestProposeDoesNotCompareAnUnshortlistedHandle` (line 1252), `TestProposeMarksAConflictingRowContested` (line 1222), `TestProposeWithOnlyFixedSubjectsBuildsNoIndex` (line 279), helper `comparisonOutcome` (line 836) | Robins / Robnis is not shortlisted; nothing fixes such a handle |
| FFI tests | `api/ffi/handlers/promote_align_test.go` `personOnNewSource` (line 30, name fixed to `Mary Smith`), `TestProposeSendsContested` (line 86); `promote_evidence_test.go` `newEvidenceFixture`; `promote_batch_test.go` `TestApplyPromoteGraphAlignment` | Every fixed handle in these tests is a look-alike the shortlist also finds |
| Swift tests | `macos/ProvenenciaTests/PromoteFlowTests.swift` helpers `proposal`, `row`, `fact`; `anUnmatchedRowStartsUnselected` (line 242), `duplicateNotesNameTheOtherRowAndItsRef` (line 853); `macos/ProvenenciaTests/PromoteModelTests.swift` helpers `store`, `model`, `subject`; `newAndSkipDecisionsAreHeldOnTheNextProposal` (line 399), `doneSendsDefaultOrChosenPins` (line 310) | Menu picks only |

**Precedent.** `22600f2a` ("promote: hold New and Skip decisions, and warn when rows collide") made New and Skip fixed pairs: copy the Go test shape of `TestProposeHoldsNewAndSkipDecisions` (one table, a fixture per case) and the model test `newAndSkipDecisionsAreHeldOnTheNextProposal` (read `store.lastPromoteFixed` after `waitUntil { !model.isProposing }`). `TestProposeDoesNotCompareAnUnshortlistedHandle` gives the off-shortlist pair (`Robins` / `Robnis`). Do not copy that commit's engine changes; nothing in Go changes here.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The flow call is `mutating func hold(subjectID: String, record: Alternative) -> Bool`, `@discardableResult`, on `PromoteFlow`, beside `setTarget` | The map's call takes id, ref, sorted and reading titles and subtitle: exactly `Alternative`'s fields. "Held, not moved" is the feature doc's word | A parameter list of five strings — because `Alternative` already names them; a new `FoundRecord` type — because it would copy `Alternative` |
| D-2 | `hold` returns false and changes nothing when the record id is empty, the row is missing, or the row is an anchor | Same refusals as `setTarget`, plus an empty id, which would select nothing | Asserting — because a stale pick after a reload must not crash the page |
| D-3 | When `row.menu` has no item with the record's id, `hold` appends a copy of the record with `assessment = .none` and `reason = .decided`, whatever the caller set; when it has one, the menu is unchanged and the existing item keeps its band and reason | The map's reason is `decided` (K-2): a found record was never offered. Band None until Go answers, then `makeRow` refreshes it. An offered record found by search is the same pick as from the menu (feature doc decided 3: the search does not leave menu records out) | Inserting first — because the menu's order is Promote's offer, best first; replacing the offered item — because it would lose the sentence it was offered with |
| D-4 | After the menu step, `hold` returns `setTarget(subjectID:token: "handle:" + record.id)` | One path for a decision: the reset, `decided`, `manual`, `updated`, duplicate notes. `target(for:on:)` now resolves the id from the menu | Copying `setTarget`'s body — because two copies drift |
| D-5 | `PromoteModel.hold(subjectID:record:)` calls the flow and, on true, runs the same follow-up as `setTarget`. That follow-up moves into one private helper, `proposeAfterDecision()` *(new)*, which both call | The map: "with `setTarget`'s re-propose follow-up" | Calling `setTarget` from `hold` in the model — because the flow already ran it; a second inline copy of three lines — because the next decision call would make a third |
| D-6 | No Go code. The engine claims are pinned by four Go tests: one in `promotealign` (no FFI) and three over the FFI handlers for propose, evidence and Apply | Working rule "Engine first, proven by tests"; map review K-7 put filing in the handler tests. The run at the planning base passed all four claims | A `promote` package test for filing — because `SaveBatch` never sees whether a handle was a candidate; the FFI test covers `ResolvePins` too |
| D-7 | `FakeStore` code does not change. AC-3.1 pins that its unscripted proposal holds a membership entity from another graph and its Apply files it | §4: the fake has no candidates, so every held handle is "outside its candidates", and its search hits come from the same memberships. The map scope entry's "(today it finds entities only among memberships)" is corrected in STEP-5 | Adding a standalone entity store to the fake — because nothing can search or show such an entity |
| D-8 | The off-shortlist proof in AC-2.1 is "the unfixed row neither targets the handle nor lists it", not "New with `no_match`" | S11-47 removes New and `no_match` from Go's proposals; the test must outlive it | Asserting `TargetNew` / `ReasonNoMatch` — because S11-47 would have to rewrite it |
| D-9 | Go tests pin target, handle, reason, `Contested`, `ConflictWithFixed` and comparison outcomes; only the FFI test pins a band (`none` for a different first name), which the feature doc states | Bands move with registry knobs (S11-43, S11-45); a different given name and surname read `none` under any of them | Pinning `weak` for Robnis — because a knob change would break a test that is about holding |
| D-10 | The FFI tests get a named person helper: `namedPersonOnNewSource(t, dir, userID, typeID, name, sexKey, polarity)` *(new)*; `personOnNewSource` keeps its signature and calls it with `"Mary Smith"`. One fixture, `newFoundRecordFixture(t)` *(new)*, serves AC-2.3 – AC-2.5 | The three FFI tests need the same differently named pair; the existing helper hard-codes the name | A copy of `personOnNewSource` — because it is 50 lines; changing every caller's arguments — because the Implementer block forbids touching unrelated callers |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PromoteFlow.hold(subjectID:record:)` *(new)* | func | `PromoteFlow.setTarget(subjectID:token:)` | Reused: `hold` adds the menu item, then calls it (D-4). A token cannot carry the record's titles |
| `PromoteModel.hold(subjectID:record:)` *(new)* | func | `PromoteModel.setTarget(subjectID:token:)` | Same follow-up, shared through `proposeAfterDecision()` (D-5) |
| `PromoteModel.proposeAfterDecision()` *(new, private)* | func | the three lines in `setTarget` after the flow call | Extracted, not copied |
| `TestProposeHoldsAnUnshortlistedHandle` *(new)* | test | `TestProposeDoesNotCompareAnUnshortlistedHandle`, `TestProposeMarksAConflictingRowContested` | Uses their fixture and pair; searched `rg "Fixed\{" core/database/promotealign/*_test.go`: no test fixes an unshortlisted handle |
| `namedPersonOnNewSource` *(new)*, `newFoundRecordFixture` *(new)* | test helpers | `personOnNewSource`, `newEvidenceFixture` | `personOnNewSource` becomes a call to the named helper; `newEvidenceFixture` is three look-alike members, the opposite of what these tests need |
| `TestProposeHoldsAFoundRecord`, `TestGetPromoteClaimEvidenceForAFoundRecord` *(new)*; subtest "files a found record" | tests | `TestProposeSendsContested`, `TestGetPromoteClaimEvidence`, `TestApplyPromoteGraphAlignment` "pins the evidence's offers" | New cases; the existing ones fix only look-alikes |
| Five `PromoteFlowTests` and one `PromoteModelTests` case *(new)* | tests | `anUnmatchedRowStartsUnselected`, `newAndSkipDecisionsAreHeldOnTheNextProposal` | Existing helpers; no new helper |

**Builds on.** The fixed-pair path from `22600f2a`; `setTarget` as the one decision path; `makeRow`'s decided branch, unchanged.
**Deletes.** None — nothing becomes dead. The three follow-up lines in `PromoteModel.setTarget` move into `proposeAfterDecision()`.
**Next adds.** S11-36 builds a `PromoteFlow.Alternative` from a `SearchHit` (its `header` through the same title and subtitle helpers `PromoteFlow.menu(for:kind:orders:)` uses) and calls `PromoteModel.hold`. S11-47 changes what an unfixed row carries; D-8 keeps these tests out of its way.

### Schema

None — no catalog change.

### Go core

None — no production code. Tests only, in `core/database/promotealign` (D-6).

### FFI

None — no proto or handler change. Tests only, in `api/ffi/handlers`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Hold a found record | `PromoteFlow.hold(subjectID:record:)` *(new)* | `macos/App/Features/Promote/PromoteFlow.swift`, after `setTarget` | Sketch below |
| Model entry | `PromoteModel.hold(subjectID:record:)` *(new)*, `proposeAfterDecision()` *(new, private)* | `macos/App/Features/Promote/PromoteModel.swift`, after `setTarget` | — |

```swift
/// Holds a record Promote did not offer: it joins the row's menu as the
/// researcher's choice, then is selected like a menu pick. Returns true when
/// suggested rows should be proposed again.
@discardableResult
mutating func hold(subjectID: String, record: Alternative) -> Bool

// PromoteModel
func setTarget(subjectID: String, token: String) {
    guard flow.setTarget(subjectID: subjectID, token: token) else { return }
    proposeAfterDecision()
}

func hold(subjectID: String, record: PromoteFlow.Alternative) {
    guard flow.hold(subjectID: subjectID, record: record) else { return }
    proposeAfterDecision()
}

/// A decision changed the fixed pairs: propose again.
private func proposeAfterDecision()
```

No view changes. States: the row reads its band and reason through `selectionReadout` as today; band None with `.decided` is already drawn (PromoteReadout preview "Readout badges").

### Strings

None — no copy. `.decided` is worded by the existing `L10n.Promote.reasonDecided`.

## 7. Steps

Xcode check, used below as **XT**: `NSUnbufferedIO=YES xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/<Suite> CODE_SIGNING_ALLOWED=NO` (from `scripts/check-pr.sh`).

**STEP-1 — Pin the engine in `promotealign`**
Files: `core/database/promotealign/promotealign_test.go`. Rule: `.cursor/rules/go-tests.mdc`.
Do: add `TestProposeHoldsAnUnshortlistedHandle` (T-1) after `TestProposeDoesNotCompareAnUnshortlistedHandle`. One table over `sex` in `{"", "same", "conflict"}`; each case builds its own `newFixture(t)`. Filed `f.person(f.source, f.artifact, "Robins")`, given `female` sex at birth unless the case is `""`; promote it. Subject `Robnis` on `f.newSource("Register")`, given `female` for `same` and `male` for `conflict` (cite the Observation as `TestProposeMarksAConflictingRowContested` does). Propose unfixed, then with `[]graphalign.Fixed{{SubjectID: robnis.ID, HandleID: h}}`, and assert AC-2.1 and AC-2.2 with `rowFor` and `comparisonOutcome`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestProposeHoldsAnUnshortlistedHandle -v` passes all three cases.

**STEP-2 — Pin propose, evidence and Apply over the FFI**
Files: `api/ffi/handlers/promote_align_test.go`, `promote_evidence_test.go`, `promote_batch_test.go`. Rule: `.cursor/rules/go-tests.mdc`.
Do: in `promote_align_test.go`, move `personOnNewSource`'s body into `namedPersonOnNewSource` *(new)* with a `name string` parameter after `typeID`, passed to `namevaluetest.Western(name)`, and the Source title still `"Register " + polarity`; `personOnNewSource` returns `namedPersonOnNewSource(t, dir, userID, typeID, "Mary Smith", sexKey, polarity)`. Add `type foundRecordFixture struct { dir, userID, graphID, incoming, handle string }` and `newFoundRecordFixture(t)` *(new)*: `subjectFixture`, a `Mary Smith` (male, positive) promoted with `promoteOK`, then `Polly Jones` (male, positive) on a new Source; `graphID` is `soleGraphID` of that Source. Add `TestProposeHoldsAFoundRecord` (T-2), `TestGetPromoteClaimEvidenceForAFoundRecord` (T-3) in `promote_evidence_test.go`, and the subtest `files a found record` (T-4) at the end of `TestApplyPromoteGraphAlignment`, which proposes first to read the revision as "pins the evidence's offers" does.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/ -run 'TestProposeHoldsAFoundRecord|TestGetPromoteClaimEvidence|TestApplyPromoteGraphAlignment|TestProposeSendsContested|TestProposePromoteGraphAlignment' -v` passes.

**STEP-3 — `PromoteFlow.hold`**
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`. Skill: `add-swift-test`.
Do: add `hold(subjectID:record:)` after `setTarget` per D-1 – D-4 and the sketch. Add T-5 – T-9.
Check: XT with `PromoteFlowTests` passes, old and new.

**STEP-4 — `PromoteModel.hold` and the fake**
Files: `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`. Skill: `add-swift-test`.
Do: add `proposeAfterDecision()` and `hold(subjectID:record:)` per D-5; `setTarget` calls the helper. Add T-10. Do not edit `FakeStore.swift` (D-7).
Check: XT with `PromoteModelTests` passes, old and new; `git diff --stat -- macos/App/Platform/FakeStore.swift` is empty.

**STEP-5 — Docs**
Files: `docs/deployment-plan/spike-11/promote-find-another-record.md`, `deployment-plan.md` (map row and S11-35 scope entry), `completed.md`, this plan.
Do: in the feature doc, after the "What the engine already does" table, replace the paragraph starting "S11-35 pins this with tests" with one that names T-1 – T-4 by test name and says no Go code changed. In the map's S11-35 scope entry, replace "`FakeStore`'s propose holds a fixed handle outside its candidates (today it finds entities only among memberships)" with "`FakeStore`'s unscripted propose already holds a fixed handle of any filed record (it has no candidates); a model test pins it". Then the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-35-hold-found-record.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Hold a record Promote did not offer as a decided row."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeHoldsAnUnshortlistedHandle` *(new)* | Robins filed, Robnis incoming; sex `""` / `same` / `conflict` → unfixed row neither targets nor lists the handle; fixed row `TargetHandle` on it, `ReasonDecided`, `ConflictWithFixed` false, a `name` comparison, `sex_at_birth` `OutcomeUnknown` / `OutcomeAgree` / `OutcomeConflict`, `Contested` false / false / true | AC-2.1, AC-2.2, D-8, D-9 |
| T-2 | Go | `api/ffi/handlers/promote_align_test.go` `TestProposeHoldsAFoundRecord` *(new)* | `newFoundRecordFixture`; unfixed → the incoming row's `handle_id` is not the handle and no alternative is; fixed → `target` `handle`, `handle_id` the handle, `reason` `decided`, `assessment` `none`, `contested` true | AC-2.1, AC-2.3 |
| T-3 | Go | `api/ffi/handlers/promote_evidence_test.go` `TestGetPromoteClaimEvidenceForAFoundRecord` *(new)* | Same fixture, `GetPromoteClaimEvidence` on the pair → no error; `sectionOf(resp, "name")` and `sectionOf(resp, "sex_at_birth")` non-nil; `len(resp.GetOffers()) > 0` | AC-2.4 |
| T-4 | Go | `api/ffi/handlers/promote_batch_test.go` `TestApplyPromoteGraphAlignment` / `files a found record` *(new)* | Same fixture; propose fixed, Apply `{incoming, "handle", handle, DefaultPins}` at its revision → one written row, entity id the handle, `GetPins() > 0` | AC-2.5 |
| T-5 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aFoundRecordIsHeldAndJoinsTheMenu` *(new)* | Row `a`: `row(id: "a", target: "new", handleID: "", assessment: "none")` with `reason = .noMatch`, `alternatives = []`; `mapRest()`; `hold` AC-1.1's record → AC-1.1 and AC-1.2 | AC-1.1, AC-1.2, D-3 |
| T-6 | Swift | same file, `aFoundRecordSurvivesAReproposalWithItsBandRefreshed` *(new)* | T-5's start plus row `b`; merge a proposal with `a` on `e7` `weak` `decided` contested → AC-1.3; `setTarget(b, "handle:e2")`, merge with `a` as `new` / `none` → AC-1.4 | AC-1.3, AC-1.4 |
| T-7 | Swift | same file, `aFoundRecordAlreadyInTheMenuIsPickedNotAddedAgain` *(new)* | `row(id: "a")` (menu `e1`, `e2`); `hold` id `e2` → AC-1.5 | AC-1.5, D-3 |
| T-8 | Swift | same file, `holdingAFoundRecordResetsTheOldTargetsEvidence` and `aFoundRecordSharingAHandleNamesTheOtherRow` *(new)* | `setEvidence` on `a`'s `e1`, then `hold` `e7` → AC-1.6. Rows `a` (unset) and `b` on `e1`; `hold(a, e1)` → AC-1.7 | AC-1.6, AC-1.7, D-4 |
| T-9 | Swift | same file, `anAnchorOrMissingRowOrEmptyRecordHoldsNothing` *(new)* | Anchor row, unknown id, record id `""` → false; `rows` equal before and after | AC-1.8, D-2 |
| T-10 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `aFoundRecordIsProposedHeldAndFiled` *(new)* | `store(subjects: [Ada])`, no `promoteProposals`; `store.membershipBySubject["x"]` on entity `e-far` (`PER-FAR`, `type-person`, kind `person`); `load`, `mapRest`, `hold(a, e-far)`, `waitUntil { !model.isProposing }` → AC-3.1's fixed pair, row and menu; `done()`, `waitUntil { store.lastPromoteBatch != nil }` → `a` / `handle` / `e-far` | AC-3.1, D-5, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteFlowTests`, `PromoteModelTests` (the `setTarget` follow-up moved), `TestProposeSendsContested` and every user of `personOnNewSource` (its body moved).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored; no Go production file and no FakeStore.swift line changed
- [ ] No new user-facing copy
- [ ] Version bump — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the S11-35 scope entry's FakeStore sentence (STEP-5); completed.md: one short paragraph of what shipped
- [ ] promote-find-another-record.md names the engine tests (STEP-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None — nothing on the page calls `hold` until S11-36 adds the menu entry. The slice Check for this PR is T-5 – T-10 and the Go tests T-1 – T-4.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S11-47 or S11-34 lands first and moves `makeRow`, `draftTarget` or the model's `setTarget` | Step 2 of the Implementer block: compare the §4 rows, stop on a change |
| A later registry change moves Robnis's band | D-9: T-1 pins no band |
| `personOnNewSource`'s move changes a fixture's Source title or name | STEP-2's check reruns its callers |

**Stop and report if:**

- T-1 – T-4 fail at the planning base for a reason other than a typo in the test: a fixed off-shortlist handle is refused, not loaded (no `name` comparison), not contested on a sex conflict, has no evidence, or does not file. That is the engine gap the working rule says S11-35 fixes; report the failing assertion and the function before writing Go code.
- The unfixed Robnis or Polly Jones row targets or lists the filed handle (the pair is no longer off-shortlist).
- `makeRow`'s decided branch no longer keeps `previous.menu`, or no longer refreshes the held item's band.
- `target(for:on:)` no longer resolves `handle:<id>` from `row.menu`.
- T-10 needs a `FakeStore.swift` change to pass.

## 12. Open questions

None.

## 13. PR

**Title:** S11-35 — Hold a record Promote did not offer
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (test names); deviations from the plan; manual check (none, per §10); `Plan: docs/deployment-plan/spike-11/prs/S11-35-hold-found-record.md`.
