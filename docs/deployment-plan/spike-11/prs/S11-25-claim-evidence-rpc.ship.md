# Ship log — S11-25 Claim evidence RPC and pin policy

**Plan:** [`S11-25-claim-evidence-rpc.md`](S11-25-claim-evidence-rpc.md) · **Branch:** `s11-25-claim-evidence-rpc` · **PR:** #517
**State:** ready

## Round 0 — implement
Base `6504003`.
Head `ce5d342`. Gate PASS (run by orchestrator). Report: 6 commits; deviations — T-7 asserts the sex_at_birth pins only (fixture also yields a name line), merged handle not tested at handler level; `DateValueDisplayTests.aComparisonFormatsABoundInTheLocale` deleted with `ComparisonValueDisplay.swift`; dead code removed (`comparisonsFrom`'s `metas`, `provisionalOnly`, `FakeStore.recordPin`); tests renamed (`TestSaveBatchPinsANeighborRecord`, `TestProposeComparesTheLifespan`); `TestProposeNegativeDateConflictLine` deleted; Swift `ValueCause.letterCase` for wire `case`; `loadEvidence` synchronous with its own task; `ResolvePins` loads Source inputs lazily. Manual check not walked.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (`macos/App/**` and strings route UX; `api/ffi/**`, `core/claimevidence`, `core/database/**`, and `macos/App/Features/**` models route SEC+PRF).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major (reviewer: blocker) | `evidenceValueProto` is a second `ConclusionValue` encoder beside `conclusionValueProto` | Fix now → F-1 | verified `promote_evidence.go:149` vs `conclusion_detail.go:140`: same five builders. Same-pass cleanup: supersedes D-1's "`evidenceValueProto` … as `conclusionValueProto` does". Rated major: one package, two copies |
| UX-1 | UX | major | Argument hint says "Drafted from the N pinned comparisons" though D-15 drafts nothing and N counts pins | To user → **Remove the hint** (user, 2026-10-21) → F-2 | verified `PromoteView.swift:701`, `L10n.swift:1466` |
| MNT-2 | MNT | major | `promote.Result.Pins` is dead now that `Save` writes no pins | Fix now → F-3 | verified `promote.go:51-53`; only `Written.Pins` is read |
| SEC-1 | SEC | minor | `ResolvePins` passes a caller's `PinSets` through on rows it does not resolve | Fix now → F-4 | verified: `out[i] = row.BatchRow` then `continue` before the reset; no wire path today. Code this PR adds |
| UX-2 | UX | minor | Relationship checkbox spoken label ends in a dangling "and" | To user → **Add new key** (user, 2026-10-21) → F-2 | verified `PromoteView.swift:806` |
| MNT-3 | MNT | minor | Pin id field choice lives at two call sites | Fix now → F-5 | `pins.go:49,62` and `promote_evidence.go:82,133`; code this PR adds |
| MNT-4 | MNT | minor | GoStore copies the `Fixed` proto mapping into `getPromoteClaimEvidence` | Fix now → F-6 | same-pass cleanup; `GoStore.swift:967-973` vs `1052-1058` |
| MNT-5 | MNT | minor | `compareStrings` re-implements `strings.Compare` | Fix now → F-7 | `promote_evidence.go:101` |
| MNT-6 | MNT | minor | `promote-graph-alignment.md:87` still says the proposal carries drafted pins (also UX outside-lens note) | Fix now → F-8 | verified line 87 |
| UX-4 | UX | minor | Relationship accessibility identifier is not unique per neighbor | Fix now → F-2 | `PromoteView.swift:808` uses only `relationship.key` |
| UX-3 | UX | minor | Loaded evidence with no lines and no relationships shows bare column headers | Later PR → S11-20 | needs new copy, and S11-20 redraws the sheet; written into the S11-20 map scope entry by F-8 |

Nits (logged): `CON-1 · completed.md · "In review" wording vs "Landed"` · `CON-2 · promote_evidence_test.go · merged handle not exercised at handler level` · `UX-5 · PromoteView.swift:860 · 50-member spoken label` · `UX-6 · PromoteView.swift:774 · stale evidence on failed refetch` · `MNT-7 · values.go:27,53 · pools copied` · `MNT-8 · pins.go:112 · Source read per row; single PinSets reset` · `MNT-9 · pins.go:40-68 · walk repeated` · `MNT-10 · PromoteFlow.swift:362 · fixed(for:) could be static`

Outside-lens notes: CON's PRF pointer (one `evidence` per handle row at Done) is S11-19's accepted cost, timed in S11-21; Rejected.

### Fix plan 1

**F-1 — One `ConclusionValue` encoder** (from MNT-1)
Files: `api/ffi/handlers/conclusion_detail.go`, `api/ffi/handlers/promote_evidence.go`, `api/ffi/handlers/conclusion_detail_test.go`.
Failing test first: a `conclusionValueProto` table case for a term with `TermKey` + `TermLabel` and no `TermID` → `ConclusionTerm{Key, Label}` with empty `Id`.
Do: widen `conclusionValueProto`'s term case to `len(v.TermID) > 0 || v.TermKey != ""`, setting `Id` only when `TermID` is set. Rewrite `evidenceValueProto` to build a `conclusiondetails.Value` from `EvidenceValue` (Name, Date, TermKey = `Value.Term`, TermLabel, HasInteger/Integer, HasText/Text, chosen by `v.Type`) and return `conclusionValueProto(...)`; it holds no `engine.ConclusionValue_*` literals. Supersedes D-1's "as `conclusionValueProto` does" — note it in the plan's Landed note.
Check: `grep -n "ConclusionValue_" api/ffi/handlers/*.go | grep -v _test` prints only lines in `conclusion_detail.go`; `TestGetPromoteClaimEvidence` passes; scripts/check-pr.sh PASS.

**F-2 — Sheet copy and accessibility** (from UX-1, UX-2, UX-4)
Files: `macos/App/Features/Promote/PromoteView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: (UX-1, user decision) drop the Argument field's hint: `PVField(label: L10n.Promote.argument)` with no hint; delete `L10n.Promote.argumentHint(pinned:)` and the `promote.sheet.argument.hint` key. (UX-2, user decision) add `L10n.Promote.pinRelationship(_ name: String)` = "Pin the records through %@" (`promote.sheet.pinRelationship`, comment "Spoken label of the checkbox that pins every record reached through a related person; %@ is that person's name"), and use it as the relationship header checkbox's `accessibilitySpokenLabel` in place of `pinLine(property: through(name), here: "", there: "")`. (UX-4) identifier `promote.sheet.relationship.<neighborSubjectID>.<key>`.
Check: `rg "argumentHint|argument\.hint" macos` prints nothing; `python3 scripts/check-localizable-xcstrings.py` passes; scripts/check-pr.sh PASS.

**F-3 — Delete `promote.Result.Pins`** (from MNT-2)
Files: `core/database/promote/promote.go`.
Do: delete the field and its comment. `Written.Pins` stays.
Check: `go build ./...`; scripts/check-pr.sh PASS.

**F-4 — `ResolvePins` is the only source of pin sets** (from SEC-1, MNT-8 rides along)
Files: `core/database/promotealign/pins.go`, `core/database/promotealign/pins_test.go`.
Failing test first: `TestResolvePins/a caller's pin sets are dropped` — a handle row with `BatchRow.PinSets` set and no `DefaultPins`/`PinIDs` comes back with no pin sets.
Do: set `out[i].PinSets = nil` right after `out[i] = row.BatchRow`, before any `continue`; remove the two later resets. Read `c.Graph().Source(sourceID)` once outside the loop.
Check: the test passes; scripts/check-pr.sh PASS.

**F-5 — Pin ids take the evidence types** (from MNT-3)
Files: `core/database/promotealign/pins.go`, `api/ffi/handlers/promote_evidence.go`, their tests.
Do: `LinePinID(l claimevidence.Line, p claimevidence.Pin)` and `BundlePinID(rel claimevidence.Relationship, b claimevidence.Bundle)`; both call sites pass the evidence values, so the field choice lives in `pins.go` only.
Check: `rg "LinePinID\(|BundlePinID\(" core api` shows only calls passing the evidence types; scripts/check-pr.sh PASS.

**F-6 — One Swift `Fixed` mapping** (from MNT-4)
Files: `macos/App/Platform/GoStore.swift`.
Do: `private static func mapFixed(_:) -> Provenencia_Engine_V1_PromoteGraphAlignmentFixed` beside `mapEvidenceSection`; use it in both the propose and the claim-evidence requests.
Check: `rg -n "Provenencia_Engine_V1_PromoteGraphAlignmentFixed\(" macos/App` prints only `mapFixed`'s line; scripts/check-pr.sh PASS.

**F-7 — `strings.Compare`** (from MNT-5)
Files: `api/ffi/handlers/promote_evidence.go`.
Do: sort with `strings.Compare(a.GetId(), b.GetId())`; delete `compareStrings`.
Check: `rg "func compareStrings" api` prints nothing; scripts/check-pr.sh PASS.

**F-8 — Docs** (from MNT-6, UX-3 Later PR)
Files: `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/deployment-plan.md`, `docs/deployment-plan/spike-11/prs/S11-25-claim-evidence-rpc.md`.
Do: line 87 → "the comparisons (the score's inputs); pins come from the claim evidence (§9)". In the S11-20 scope entry, add after the **Lifespan** paragraph: `**Empty evidence (from S11-25 review UX-3).** When a row's evidence has no lines and no relationships, say there is nothing to compare instead of drawing the column headers alone.` In the plan's Landed note, record: F-1 supersedes D-1's separate encoder; the argument hint is removed and `promote.sheet.pinRelationship` added (user, 2026-10-21), superseding the Strings section's "Removed: none".
Check: `scripts/check-plan-docs.py` 0 errors; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `1f026ea`. Gate PASS (run by orchestrator). F-1 to F-8 done. Deviations: F-1/F-4/F-5/F-7 share commit `1dcaef5`; F-4's Source read is lazy inside the existing `!ready` block; F-5's "unoffered id" test value is `offers[0].ID + "x"`.

**Orchestrator finding ORC-1 · major · `autoreconcile.Compatible` is dead.** The deleted `exhibits.go` was its only production caller (`rg '\bCompatible\b' core api` outside tests: only its definition). It also leaves S11-28's plan naming `exhibits.go`, so a full `check-plan-docs.py` errors (`S11-28-alias-fold.md:91`). Decision: **Fix now** → F-9; S11-28's plan amended in the same fix (user, 2026-10-21, "Delete it here"). Goes to round 2 with the MNT reviewer.

### Fix plan 1b

**F-9 — Delete `autoreconcile.Compatible`; amend S11-28** (from ORC-1)
Files: `core/autoreconcile/modules.go`, `core/autoreconcile/modules_test.go`, `docs/deployment-plan/spike-11/prs/S11-28-alias-fold.md`, `docs/deployment-plan/spike-11/prs/S11-28-alias-fold.review.md`, and any doc or skill that names `Compatible` (`rg -n '\bCompatible\b' core api docs .agents .cursor`).
Do: delete `Compatible`, its doc comment, `TestCompatible` and `TestCompatibleIgnoresAliases`, and anything only they used. In S11-28's plan: delete the Scope row "`Compatible` follows the new fold…", the Codebase-fit row "Compatible" (line 91), D-7, T-3, and the "Rename `TestCompatibleIgnoresAliases`…" sentence in STEP-1; reword the `aliasRules` row's "Replaces" cell to "—" and its reason to "Shared alias table for T-1"; drop "`core/database/promotealign` tests that count exhibits through `Compatible`" and the following `promotealign` expectation sentence from §9's "Most likely to break". Leave a dated line under S11-28's review log: "2026-10-21 — S11-25 deleted `autoreconcile.Compatible` with its last caller (`promotealign/exhibits.go`); D-7, T-3 and their rows removed. Other numbering unchanged." Do not renumber other Decisions or Tests. In S11-25's plan Landed note, add that it deletes `Compatible`.
Check: `rg -n '\bCompatible\b' core api` prints nothing; `python3 scripts/check-plan-docs.py` reports 0 errors; scripts/check-pr.sh PASS.

F-9: head `8f008d7`. Gate PASS (run by orchestrator). Deviations: also reworded live `Compatible` mentions in `add-match-rule` SKILL.md, `conclusion-reconciliation.md:146`, `promote-graph-alignment.md:135`; history lines left.

Minor read by the orchestrator: SEC-1 **Resolved** (`pins.go`: `out[i].PinSets = nil` follows `out[i] = row.BatchRow` before any `continue`; `TestResolvePins/a caller's pin sets are dropped`).

## Round 2 — review
Scope `ce5d342..8f008d7`. Reviewers: MNT (MNT-1, MNT-2, ORC-1), UX (UX-1). Not sent: CON (no findings), SEC+PRF (SEC-1 minor, resolved above).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | second `ConclusionValue` encoder | Resolved | `promote_evidence.go:140-155` calls `conclusionValueProto`; `ConclusionValue_` literals only in `conclusion_detail.go` |
| MNT-2 | MNT | major | `promote.Result.Pins` dead | Resolved | field deleted |
| ORC-1 | MNT | major | `autoreconcile.Compatible` dead | Resolved | `rg '\bCompatible\b' core api` empty; `check-plan-docs.py` 0 errors |
| UX-1 | UX | major | argument hint claims a draft | Resolved | hint, `argumentHint(pinned:)` and key removed |

No new findings. Minors UX-2, UX-4 confirmed in code by the UX reviewer; MNT-3 to MNT-6 confirmed by the MNT reviewer.
Doc leftovers fixed by the orchestrator in the ship-log commit: S11-28 STEP-1 Files no longer lists `modules_test.go` (MNT round 2); S11-25 Landed note records the relationship identifier change from D-14 (UX round 2 outside-lens note).

## Result
Gate PASS · CI pending · open: none · follow-ups: UX-3 (S11-20 scope entry) · to user: UX-1 (removed hint), UX-2 (new key), ORC-1 (delete `Compatible`, amend S11-28) — all answered 2026-10-21 · CON miss: none
