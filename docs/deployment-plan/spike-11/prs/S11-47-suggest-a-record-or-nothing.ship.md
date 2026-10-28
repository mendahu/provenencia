# Ship log — S11-47 Promote suggests a record or nothing

**Plan:** [`S11-47-suggest-a-record-or-nothing.md`](S11-47-suggest-a-record-or-nothing.md) · **Branch:** `s11-47-suggest-a-record-or-nothing` · **PR:** #598
**State:** ready

## Round 0 — implement
Head `b7cae938`. Gate PASS (orchestrator run against `c409b9d4`). Report: no deviations. Base moved from Planned `a70932e` to `c409b9d4`; no Current-state file changed. AC-1.8 and AC-2.5 are the manual check (plan §10), not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`, `Localizable.xcstrings`. SEC+PRF: `api/ffi/**` (security); `macos/App/Features/Promote/PromoteModel.swift` and `core/graphalign` graph code (performance).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | A weak-drafted row's Updated caption reads "No record suggested" while a record is selected | Fix now → F-1 | verified: `makeRow` gives `reason: Reason(proposal)` = `.unplaced` for a weak unplaced row, and `draftTarget`'s weak arm drafts `.handle(menu.first)`. `captionReason` is `move ?? reason`, which gives `.unplaced`. Regression from `weak` → `unplaced` (D-3) |
| MNT-1 | MNT | major | Send side spells `graphalign.Target` as raw strings beside the new `PromoteTargetCode` | Fix now → F-2 (Fixed only); batch row → Follow-up | Same-pass cleanup: this PR adds the typed copy (`PromoteTargetCode`) of the vocabulary `CatalogPromoteGraphAlignmentFixed.target` carries (`promotealign.checkFixed` switches on `graphalign.Target*`). No behavior a researcher sees changes, so it supersedes the Out line "Typing the send side" for `Fixed`, not To user. The batch row is `promote.Target` (`core/database/promote/batch.go:21`), a separate Go type whose Swift spelling this PR does not touch → Follow-up |
| CON-2 / MNT-2 | CON, MNT | minor | `possible_duplicate` comments (proto field 19, `RowFlags.PossibleDuplicate`) still say "both are New" | Fix now → F-3 | `wouldBeNew` (`rows.go:126`) admits unplaced rows; both files are in this diff |
| MNT-3 | MNT | minor | name-candidates.md says the removed Skip fallback still survives | Fix now → F-3 | `draftTarget` now ends `return .unset` |
| MNT-4 | MNT | minor | add-match-rule's row-reason recipe omits the proto comment listing the codes | Fix now → F-3 | Plan §4 precedent includes "proto comment"; D-6 had to edit it |
| CON-1 | CON | minor | Plan has no Landed note | Fix now → F-4 | DoD box "a Landed note lists every difference"; this round adds differences (F-1, F-2) |

Nits (logged): `CON-3 / MNT-5 · core/graphalign/rows.go:132-136 · rewrap the flagAlikeNewRows doc comment` (rides with F-3) · `MNT-6 · core/graphalign/types.go:76 · Target doc could say what "" means on Fixed vs Row`

### Fix plan 1

**F-1 — An unplaced row drafted onto a weak record captions the record's reason** (from UX-1)
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`. Skill: `add-swift-test`.
Failing test first: `PromoteFlowTests/anUnplacedRowDraftedOntoAWeakRecordCaptionsItsReason`. Merge a re-proposal where row `a` is assessment `weak`, target unplaced (`""`), reason unplaced, alternative `e9` with reason `weak`, over a page whose row `a` was on `e1`. Expect `updated == true`, `target` on `e9`, and `captionReason == .weak`. It fails today with `.unplaced`.
Do: `var captionReason: Reason { move ?? (reason == .unplaced ? selectionReadout?.reason : nil) ?? reason }`. Update its doc comment to say "else, for a row Promote suggested no record for but the page drafted onto one, that record's reason". The existing `aRowThatLosesItsRecordSaysNoRecordIsSuggested` must still pass (it drafts `.unset`, so `selectionReadout` is nil).
Check: both tests pass; scripts/check-pr.sh PASS.

**F-2 — Type the Fixed send side as `PromoteTargetCode`** (from MNT-1; supersedes the plan's Out line "Typing the send side" for `CatalogPromoteGraphAlignmentFixed` only)
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Features/Promote/PromoteFlow.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`, any FakeStore read of `Fixed.target`.
Failing test first: extend `PromoteFlowTests/reasonCodesParseThroughTheWireVocabulary` with a round trip: for every `PromoteTargetCode` case except `.unrecognized`, `PromoteTargetCode(wire: c.wire) == c`. It fails to compile until `wire` exists.
Do: add `var wire: String` to `PromoteTargetCode` (`.unplaced` → `""`, `.handle` → `"handle"`, `.new` → `"new"`, `.skip` → `"skip"`, `.unrecognized(let s)` → `s`). Make `CatalogPromoteGraphAlignmentFixed.target: PromoteTargetCode = .handle`. `fixed(for:)` passes `.new` / `.skip`. `GoStore.mapFixed` sends `f.target.wire`. Update the test fixtures at `PromoteModelTests.swift:412-413` to `.new` / `.skip`. Leave `CatalogPromoteBatchRow.target`, `batchRow`, the FakeStore batch switch, and `Target.token` (menu tokens) as they are; they are Out.
Check: `rg -n 'CatalogPromoteGraphAlignmentFixed\([^)]*target: "' macos` prints nothing. `rg -n 'var target: String = "handle"' macos/App` prints nothing. scripts/check-pr.sh PASS.

**F-3 — Comments and docs the change made false** (from CON-2, MNT-2, MNT-3, MNT-4; nits CON-3/MNT-5)
Files: `api/proto/engine.proto` (field 19 comment), the generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift` (regenerate with `./scripts/generate-proto.sh` as round 0 did), `core/graphalign/types.go:160-162`, `core/graphalign/rows.go:130-137`, `docs/deployment-plan/spike-11/name-candidates.md:69`, `.agents/skills/add-match-rule/SKILL.md:78`.
Do:
- Reword field 19 and `RowFlags.PossibleDuplicate` to "or both would be new records (unplaced or decided New) and score alike".
- Rewrap the `flagAlikeNewRows` doc comment to the neighbors' width.
- name-candidates.md: put the fallback sentence in the past tense ("the Skip / New it received survived only in a rare fallback (an unplaced row with a band that was not taken fell through to Skip)").
- add-match-rule: after "its `PromoteReasonCode` wire case with its row in `TestSwiftVocabularyMatchesGo`", add ", its code in the `PromoteGraphAlignmentRow.reason` comment in `api/proto/engine.proto` (regenerate)".
Check: `rg -n "both are New|both rows are New" api core macos/App` prints nothing. `gofmt -l core` prints nothing. scripts/check-pr.sh PASS.

**F-4 — Landed note** (from CON-1)
Files: `docs/deployment-plan/spike-11/prs/S11-47-suggest-a-record-or-nothing.md`.
Do: add a Landed note under the Status line. Say the proto output was regenerated with `./scripts/generate-proto.sh`. List F-1 (caption for a weak-drafted unplaced row), and F-2, which supersedes the Out line "Typing the send side" for `CatalogPromoteGraphAlignmentFixed` (same-pass cleanup; the batch row stays Out as a follow-up). Also list F-3's comment edits beyond D-6, D-9 and D-15.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-47-suggest-a-record-or-nothing.md` reports 0 errors; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `a57caf98`. Gate PASS (orchestrator run). F-1, F-2, F-3, F-4 done. Deviations: `PromoteFlowTests.swift:911-912` also read `Fixed.target` as a string and moved to `.new` / `.skip`; one sentence added to `completed.md` for F-1.
CON minors read by the orchestrator: CON-1 **Resolved** (Landed note under Status), CON-2 **Resolved** (proto field 19 reworded, regenerated), CON-3 nit **Resolved** (rewrapped).

## Round 2 — review
Reviewers: UX (UX-1 major), MNT (MNT-1 major). Not sent: CON (minors only), SEC+PRF (no findings). Scope `b7cae938...a57caf98`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | (round 1) weak-drafted caption | **Resolved** | `captionReason` reads `selectionReadout?.reason` for an unplaced row; `anUnplacedRowDraftedOntoAWeakRecordCaptionsItsReason` |
| MNT-1 | MNT | major | (round 1) Fixed send side typed | **Resolved** | `Fixed.target: PromoteTargetCode`, `wire` round trip in `reasonCodesParseThroughTheWireVocabulary`; batch row stays a follow-up |

No new findings. MNT-2, MNT-3, MNT-4 **Resolved** (orchestrator read of `9847e04d`; MNT reviewer saw the same).

## Result
Gate PASS · CI green · open: none · follow-ups: MNT-1 (batch row: type `CatalogPromoteBatchRow.target` / `promote.Target` in Swift, reusing `PromoteTargetCode.wire` or its own enum), MNT-6 (nit) · to user: — · CON miss: none
