# S13-24 — The pin engine goes

**Status:** Planned
**Spike / slice:** Spike 13 · Slice 0 — Removing Observation pins
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Promote page already shows no band or pins since S13-23, and Vitals loses one cause word, "low-confidence claim", which only a claim graded before S13-23 could show; user, 2026-10-30)
**Board:** —
**Depends on:** S13-23 (landed [#633](https://github.com/mendahu/provenencia/pull/633))
**Blocks:** S13-25, S13-26, S13-29, S14-16, S14-20
**Branch:** `s13-24-the-pin-engine-goes`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0 until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md) (deleting Methods, reserving fields), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md) (removing one key), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/catalog-audit.mdc`
**Context:** [conclusion §5, §5.3](../../../conclusion-layer-data-model.md#53-promote), [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §4, [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §2, §3.1, §6, §8, §9, [`matching-model.md`](../../../matching-model.md#knob-inventory); working rules: "Pins leave in three PRs, the client first", "Promote files accepted claims and nothing else", "Identity Claims carry no confidence and no argument", "The score ranks, unseen", "Spike 14's Promote PRs build on the removal", "Slice 7 captures requirements, not an implementation", "One person's links, direct and implied, from one function", "Each removal PR updates what it makes false"
**Planned:** 2026-10-30 against `main` @ `1253dfaa`
**Review:** [`S13-24-the-pin-engine-goes.review.md`](S13-24-the-pin-engine-goes.review.md) — Fits with changes / Ready with changes, 2 rounds; every blocker Fixed; To user answered 2026-10-30

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

The engine half of the pin removal. After S13-23 the Mac app stopped asking for pins, confidence and argument, so the engine that offers and writes them is dead weight: the claim evidence RPC, the evidence query in `promotealign`, `core/claimevidence`, Apply's pin resolution and its writes onto members' claims, the confidence grade RPC, and the auto-reconciler's claim-confidence weighing. This PR deletes them and reserves their wire. It also moves the last Promote rule the app still holds, what a row starts on, into Go, so the proposal's display-only fields (score, band, Contested, reason) can be reserved too. Done files accepted claims and their connections and nothing else, as [conclusion §5.3](../../../conclusion-layer-data-model.md#53-promote) states.

**Why now.** Slice 0 runs first; S13-25 drops the tables only once nothing reads or writes them, and Spike 14's S14-16 and S14-20 rewrite the Apply handler and `promotealign` after it.

**What stays.** Proposals rank and choose exactly as today (`graphalign` scoring, the accept bar, settling, Move reasons, conflict and duplicate flags); every row the page started on before starts on the same target after (D-8); `UnpromoteSubject`; `PromoteSubject` without its two claim fields; bridge filing; the `identity_claim_evidence` table, `identityclaims.PinTx` and `PinnedObservations`, the delete facet that releases pins, and the `claim_confidence_grades` table and seed (all S13-25's); `canonlinks.Reader.Relationships` (S13-04 calls it, D-12).

## 2. User stories

**US-1.** As a researcher, when I press Done on a Promote page, the rows I chose are promoted with their connections, and no other record's claim changes.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Done with row A on handle `e1` (whose member B records the same birth date as A), row C New and row D Skip | A's claim and C's minted record and claim are written, accepted; bridges whose ends are both promoted are filed; `identity_claim_evidence` gains no row, and B's claim is unchanged |
| AC-1.2 | Done with a row on the handle its Subject already belongs to | The row writes nothing and the Done succeeds (an anchor), as for any row with no pins today |
| AC-1.3 | The Apply response | Each written row carries its record and claim only; the Apply request carries subject, target and record per row |

**US-2.** As a researcher, a value's strength on a Person page comes from the Source and the transcription, not from a claim grade.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Two members of one person record different given names; one member's claim carries the seeded grade `low_confidence` (set in the catalog before this PR), both cite `standard` Sources with certain transcriptions | Neither name is weak; the reconciler resolves them as it would two ungraded claims (both kept as distinct values, or decided by support, per the pipeline); no outcome says why from a claim |
| AC-2.2 | A weak value from a `low_trust` Source on the Person page | The outcome reads "weak · low-trust Source"; no outcome anywhere reads "low-confidence claim" |
| AC-2.3 | A project opened after this PR whose reconciler cache was written at version 20 | The cache rebuilds on open (version 21) |
| AC-2.4 | A person whose name is recorded once under a `low_trust` Source and once, differently, under a `standard` Source with an uncertain transcription | The Person page's outcome for each record carries the same `Provenance` the reconciler weighed (`{Credibility: -1, Uncertain: false}` and `{Credibility: 0, Uncertain: true}`), both read through `autoreconciler.ProvenanceColumns`; `rg -n "key = 'standard'" core/database -g '*.go' -g '!*_test.go'` finds only `autoreconciler.go` |

**US-3.** As a researcher, a Promote page starts each row where it started before this PR.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A row the walk placed on `e1` | It starts on `e1` |
| AC-3.2 | An unplaced row whose best candidate `e9` scores in the weak band, not taken by another row | It starts on `e9`, the menu lists `e9` once |
| AC-3.3 | An unplaced row whose best candidate scores medium or strong (lost one-to-one), or below the weak band, or that has no candidate | It starts unset |
| AC-3.4 | A row whose best record another row took (`taken`), whatever its band | It starts unset, with the duplicate note naming the other row |

**US-4.** As a contributor, the wire, the registry and the docs name nothing the pin engine used.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `rg -n 'claimevidence|ResolvePins|PinOffers|acceptedClaims|ClaimEvidence|ListClaimConfidenceGrades|ClaimConfidence\b|\.BindingOrder\(|\) Shown\(|Flags\.Contested' core api -g '!api/proto/**'` | No hit (`core/onboarding`'s `TestCreateCatalogSeedsClaimConfidenceGrades` stays for S13-25 and does not match; `subjectvocab`'s `sqlBindingOrder` does not match) |
| AC-4.2 | `api/proto/engine.proto` | Methods 89 and 110 and every field D-2 lists are reserved by number and name; the deleted messages are gone |
| AC-4.3 | `TestKnobInventoryMatchesRegistry` | Passes with no `evidence.*` row and no `Evidence` registry field |

## 3. Scope

| In | Covered by |
| --- | --- |
| Apply files claims and bridges only: `promote.BatchRow`, `Written`, `PinSet`, `pinSets`; `promotealign.ApplyRow`, `ResolvePins`, `PinOffers`, `LinePinID`, `BundlePinID`; the handler's mapping | STEP-2, AC-1.1, AC-1.2, AC-1.3, D-3, D-4 |
| The evidence query and RPC: `GetPromoteClaimEvidence` handler and dispatch, `promotealign` `evidence.go`, `values.go`, their tests and the evidence parts of `perf_test.go`, `memo_test.go`, `export_test.go` | STEP-1, STEP-2, AC-4.1, D-10 |
| `core/claimevidence` goes; its router's meta rule moves to `subjectcompare` as `Rules.Meta` with a test; the registry's `evidence` section, its guard entry and knob rows; `kinship.Recording.Shown`; `graphcache.Vocabulary.BindingOrder` | STEP-3, AC-4.1, AC-4.3, D-11, D-12 |
| Claims carry no confidence or argument in Go: `identityclaims.CreateInput` and `Claim`, `promote.Input`, `PromoteSubject`'s handler and `identityClaimProto`; `ListClaimConfidenceGrades` handler and dispatch | STEP-4, D-5, D-6 |
| The reconciler stops reading claim confidence: `autoreconcile.Provenance`, `autoreconciler`'s load, `CacheVersion` 21, `conclusiondetails`' outcome, the outcome handler, and Swift's weak cause (user, 2026-10-30) | STEP-5, AC-2.1, AC-2.2, AC-2.3, D-7 |
| The provenance columns have one owner: `autoreconciler.ProvenanceColumns`, `ProvenanceJoins`, `ProvenanceDest`, read by both queries (user, 2026-10-30) | STEP-5, AC-2.4, D-13 |
| What a row starts on is Go's (audit 2 B-26): `graphalign.Row.Start`, the proposal handler sends it, Swift's `draftTarget` reads it; `graphalign` drops the display-only fields | STEP-6, AC-3.1 – AC-3.4, D-8, D-9 |
| The proto: Methods and messages deleted, fields reserved, regenerated (Go and Swift) | STEP-7, AC-4.2, D-1, D-2 |
| Docs and skills this makes false (working rule "Each removal PR updates what it makes false") | STEP-8 |

| Out | Home |
| --- | --- |
| The `identity_claim_evidence` table, `core/database/identityclaims/pins.go` (`PinTx`, `PinnedObservations`), the delete facet and `ReleaseClaim`'s pin step, `TestBridgeSubjectDeleteReleasesPins`, the effects entry | S13-25 |
| The `identity_claims.confidence_grade_id` and `.argument` columns, `claim_confidence_grades` with its seed, `claimconfidencegrades`, its `Install` callers in tests, the onboarding readiness check, `deleteimpact`'s grade edge and `facets.go`'s read of the two columns | S13-25 |
| `status` on the Apply row, provisional claims | not planned (S13-30 descoped 2026-10-30, parked in [`ideas/provisional-identity-claims.md`](../../../ideas/provisional-identity-claims.md); working rule "Promote files accepted claims and nothing else") |
| A Subject-pair compare over every Property (`CompareEvery`), and exporting `lifespan` | Slice 7 (S13-26), with open question 11 (user, 2026-10-30) |
| `graphalign.Subject.Provenance` (always 1, `core/database/promotealign/layer.go`), the rest of audit 2 B-37 | S14-20 (user, 2026-10-30) |
| One shared proto header message for the four `oneof header` copies (S13-23 review log) | S14-16, which rewrites the Promote mapping (user, 2026-10-30) |
| Renaming Swift's `PromoteFlow.Reason` to say "move" | not planned (`PromoteReasonCode` is renamed here, D-9) |
| `canonical_entities.argument` | stays (working rule "Identity Claims carry no confidence and no argument") |
| `graphalign.Row.Score`, `Row.Assessment`, `Alternative.Score`, `Row.Reason` codes other than `weak`, `Via.Key` | stay: tests read the ranking through them, and `Start` reads the band and `taken` (D-9) |

## 4. Current state

Verified at `1253dfaa`.

| Area | Where | Today |
| --- | --- | --- |
| Apply handler | `api/ffi/handlers/promote_batch.go` `ApplyPromoteGraphAlignment` | Maps each row to `promotealign.ApplyRow{BatchRow, DefaultPins, PinIDs}` with `ConfidenceGradeID` and `Argument`; calls `promotealign.ResolvePins` outside the write, then `promote.SaveBatch`; sends `Written.Pins` |
| Pin resolution | `core/database/promotealign/pins.go` `LinePinID`, `BundlePinID`, `PinOffer`, `PinOffers`, `ApplyRow`, `ResolvePins` | Builds each handle row's evidence and turns chosen or default offers into `promote.PinSet`s. S13-23's app sends neither pins nor `default_pins`, so every row skips |
| Batch write | `core/database/promote/batch.go` `BatchRow` (`ConfidenceGradeID`, `Argument`, `PinSets`), `Written.Pins`, `SaveBatch`, `applyRow` | `applyRow` validates the grade, refuses pin sets on New or Skip, treats a row already on its own handle as an anchor only without pin sets (else `ErrAlreadyMember`), writes the claim with grade and argument, then `pinSets` |
| Promote step | `core/database/promote/promote.go` package doc, `Input` (`ConfidenceGradeID`, `Argument`), `PinSet`, `Save`, `pinSets` | `pinSets` is the only non-test caller of `identityclaims.PinTx`; the package doc describes pin sets and backfill |
| Claim write | `core/database/identityclaims/identityclaims.go` `Claim`, `CreateInput`, `InsertTx`, `sqlInsert`, `sqlColumns`, `scanClaim` | Writes and reads `confidence_grade_id` and `argument`; validates the grade's length; the audit create row lists both |
| Evidence RPC | `api/ffi/handlers/promote_evidence.go` `GetPromoteClaimEvidence`; `api/ffi/dispatch.go` `MethodGetPromoteClaimEvidence` | No Swift caller since S13-23 |
| Evidence query | `core/database/promotealign/evidence.go` (`EvidenceRequest`, `Evidence`, `acceptedClaims`, `bindingOrder`, `evidenceReader`, `membersMemoKey`), `values.go` (`EvidenceValues`) | Only the RPC and `ResolvePins` call it. `inputs` in `promotealign.go` is shared with `Propose`; its comment names "a claim evidence query" |
| Evidence tests | `core/database/promotealign/evidence_test.go`, `pins_test.go`, `values_test.go`; `perf_test.go` `TestLargeSourceBudgets` subtests "evidence warm", "warm queries" (its first half), "evidence on a place row", "evidence under a small cap", `budgetEvidence`, `budgetPlaceEvidence`, `evidenceRequests`; `memo_test.go` `TestEvidenceWarmIssuesNoQuery`, the `promotealign.Evidence` calls in `TestMemoFollowsCommit` and `TestMemoIsNotMutated`; `export_test.go` `MemoSnapshot.Members` read of `membersMemoKey` | Test the evidence query. The graph cap is also covered by `core/database/graphcache/evidence_graph_cap_test.go` |
| Handler tests | `api/ffi/handlers/promote_evidence_test.go` (`evidenceFixture`, `newEvidenceFixture`, `newFoundRecordFixture`, `sectionOf`, `memberObservation`, `offerOf`, `pinnedOn`, six tests; `newFoundRecordFixture` is also used by `promote_align_test.go` `TestProposeHoldsAFoundRecord`); `api/ffi/dispatch_test.go` row "claim evidence with no request"; `core/database/deleteimpact/conclusion_release_test.go` and `pins_e2e_test.go` fixtures (pin through `promote.BatchRow.PinSets`); `core/database/deleteimpact/pins_e2e_test.go` (passes `promote.Input.Argument` and reads `Claim.Argument`); `promote_batch_test.go` `TestApplyPromoteGraphAlignment` subtest "pins the evidence's offers" (uses `evidenceFixture`) | |
| Claim evidence policy | `core/claimevidence` (`claimevidence.go`, `build.go`, `sections.go` with `route`, `routeOf`, `Sections`; four test files) | Only `promotealign` and `matchregistry` import it |
| Registry | `core/matchregistry/registry.go` `Registry.Evidence`; `evidence.go` `defaultEvidence`; `evidence_test.go`; `guard_test.go` `layerPackages` entry `core/claimevidence` | Three knob rows `evidence.Checked`, `evidence.ByEye`, `evidence.Defining` in `docs/matching-model.md` |
| Helpers | `core/kinship/kinship.go` `Recording.Shown` (test `TestRecordingShown`); `core/database/graphcache/vocab.go` `Vocabulary.BindingOrder` (read in `vocab_test.go` `TestVocabularyBindings`, `TestVocabularyBindingsPluginTypeSeededKey`, `TestVocabularyDropsOnBindingReorder`); `core/database/canonlinks/reader.go` `Reader.Relationships` | `Shown`'s only caller is `core/claimevidence/build.go`; `BindingOrder`'s is `core/database/promotealign/evidence.go`; `Relationships`'s is `evidence.go`, and the working rule "One person's links, direct and implied, from one function" has S13-04 call it |
| Grades RPC | `api/ffi/handlers/promote.go` `ListClaimConfidenceGrades`; `dispatch.go` `MethodListClaimConfidenceGrades`; test `promote_test.go` (the `ListClaimConfidenceGrades` call near line 380) | No Swift caller since S13-23 |
| PromoteSubject | `api/ffi/handlers/promote.go` `PromoteSubject` (reads `confidence_grade_id`, `argument`), `identityClaimProto` (sends both); test `promote_test.go` `TestPromoteSubjectJoin` (grade and argument, and a "bad grade id" case) | |
| Reconciler | `core/autoreconcile/autoreconcile.go` `Provenance.ClaimConfidence`, `Weak`, `Stronger`; `core/database/autoreconciler/autoreconciler.go` `sqlLoadCandidates` (joins `claim_confidence_grades`), the scan into `confidence`, `CacheVersion = 20` | A low-confidence claim makes its values weak and loses the tiebreak |
| Reconciler tests | `core/autoreconcile/pipeline_test.go` (`lowClaim`, "a low-confidence claim is weak", the `Stronger` and `Weak` tables, the random `Provenance`), `names_test.go` ("a low-confidence claim drops", the random `Provenance`); `core/database/autoreconciler/autoreconciler_test.go` subtest "a low-confidence claim's different given name is weak" (passes `promote.Input.ConfidenceGradeID`), `f.grade`, the `grades` loop near line 1475 | |
| Outcome detail | `core/database/conclusiondetails/conclusiondetails.go` `Outcome.ClaimConfidenceKey`, `sqlOutcomes` (joins `identity_claims` and `claim_confidence_grades` for the key and offset), the scan; `api/ffi/handlers/conclusion_detail.go` maps `ClaimConfidenceKey`, `ClaimConfidenceOffset`; test `conclusiondetails_test.go` reads `o.Provenance.ClaimConfidence` | |
| Swift weak cause | `macos/App/Platform/GenealogyStore.swift` `CatalogReconcilerOutcome.claimConfidenceKey`, `.claimConfidenceOffset`, `.isLowConfidenceClaim`; `GoStore.swift` (two mapped lines near 1284); `FakeStore.swift` (`claimConfidenceKey: ""` near 2032); `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `weakCauses`; `L10n.swift` `L10n.Conclusions.weakLowConfidenceClaim` and `outcomeWeakBecause`'s comment; `Localizable.xcstrings` key `conclusions.weak.lowConfidenceClaim`; test `ReconciledValueDisplayTests.swift` `weakSaysWhichEvidenceIsWeak` and the `outcome(…)` helper's `claimConfidenceOffset` parameter | The only Swift reader of claim confidence left after S13-23 |
| Prose | `core/database/canonlinks/canonlinks.go` package doc ("claim evidence"); `core/graphalign/types.go` `Bridge` comment ("the claim evidence query reads them back"); `core/database/autoreconciler/autoreconciler.go` package doc (claim confidence) and `sqlLoadCandidates`' comment; `core/database/promote/batch.go` `BatchRow`, `SaveBatch` comments | Describe the removed engine |
| Proposal | `core/graphalign/rows.go` `proposal`, `cand.contested`, `alternative`, `strongestAgreement`, `comparisonsFrom`; `core/graphalign/types.go` `Comparison`, `ReasonWeak`, `Alternative` (`Score`, `Assessment`, `Reason`, `ReasonProperty`, `ViaNeighbor`), `RowFlags.Contested`, `Row.Comparisons`, `Row.ReasonProperty` | An unplaced row keeps `Target` empty and reports its best candidate's band; nothing outside `graphalign`, its tests and the handler reads the display fields |
| Proposal handler | `api/ffi/handlers/promote_align.go` `proposalProto`, `productOrigin` | Sends `target` and `handle_id` as `graphalign` placed them, plus score, assessment, reason, reason property, contested and each alternative's display fields; a row header only for a placed handle |
| Proposal tests | `api/ffi/handlers/promote_align_test.go` `TestProposeSendsContested`, `TestProposeHoldsAFoundRecord`, `TestProposePromoteGraphAlignment`; `core/graphalign/align_test.go` (reads `Contested`, `Comparisons`, `ReasonProperty`, alternatives' `Assessment` and `Reason`); `core/database/promotealign/promotealign_test.go` (reads `Comparisons`, `Contested`, alternatives' `Score`) | |
| Swift draft rule | `macos/App/Features/Promote/PromoteFlow.swift` `Assessment`, `draftTarget`; `GenealogyStore.swift` `CatalogPromoteGraphAlignmentRow.assessment`, `.reason`; `GoStore.swift` `proposePromoteGraphAlignment` maps both; `FakeStore.swift` `proposePromoteGraphAlignment` sets both; `macos/App/Platform/ConclusionVocabulary.swift` `PromoteReasonCode` (ten cases) | `draftTarget`: band none → unset; target handle → its handle; `taken` → unset; band weak → first menu item; else unset |
| Swift draft tests | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aPersonRecordReadsSortedInTheMenu`, `aMediumMatchStartsOnItsHandle`, `aWeakMatchStartsOnItsHandle`, `anUnplacedRowNeverDraftsNewOrSkip`, `aRowThatLosesItsRecordSaysNoRecordIsSuggested`, `aRowWhoseRecordWasTakenStartsUnset`, `aWeakCandidateBelowTheBarStartsSelected`, `anUnmatchedRowStartsUnselected`, `reasonCodesParseThroughTheWireVocabulary`, the `row(…)` helper's `assessment` parameter; `PromoteModelTests.swift` `row(_:handleID:assessment:)` | |
| Parity | `api/ffi/swift_vocabulary_test.go` row "promote reasons" (ten codes) | |
| Proto | `api/proto/engine.proto` `Method` (`METHOD_LIST_CLAIM_CONFIDENCE_GRADES = 89`, `METHOD_GET_PROMOTE_CLAIM_EVIDENCE = 110`), `IdentityClaim` 5–6, `ClaimConfidenceGrade`, `ListClaimConfidenceGrades*`, `PromoteSubjectRequest` 5–6, `ReconcilerOutcomeDetail` 14 and 16, `GetPromoteClaimEvidence*` and every `PromoteClaimEvidence*` message, `PromoteGraphAlignmentAlternative` 3 and 7–11, `PromoteGraphAlignmentRow` 6, 7, 17, 18, 20–23, `ApplyPromoteGraphAlignmentRow` 4, 5, 7, 8, `ApplyPromoteGraphAlignmentWritten` 3 | Generated into `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift` by `scripts/generate-proto.sh` |

**Precedent.** S13-23 ([#633](https://github.com/mendahu/provenencia/pull/633)) is the client half of this same removal: copy its order (stop reading a field, then delete it) and its sweep. The proto's reserved blocks show the house form for a deleted Method and field: `reserved 88, 92, 94, 95;` plus `reserved "METHOD_…";` in `Method`, and `reserved 3; reserved "source_id";` in a message. A moved pure function with its tests: S11-51 moved `linkSet` into `kinship.LinkSet` (working rule "One person's links, direct and implied, from one function").

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Methods `METHOD_LIST_CLAIM_CONFIDENCE_GRADES = 89` and `METHOD_GET_PROMOTE_CLAIM_EVIDENCE = 110` are removed and reserved: add `89` and `110` to a new `reserved 89, 110;` line and their names to a `reserved "METHOD_LIST_CLAIM_CONFIDENCE_GRADES", "METHOD_GET_PROMOTE_CLAIM_EVIDENCE";` line in `enum Method`. Messages `ClaimConfidenceGrade`, `ListClaimConfidenceGradesRequest`, `ListClaimConfidenceGradesResponse`, `GetPromoteClaimEvidenceRequest`, `GetPromoteClaimEvidenceResponse` and every message named `PromoteClaimEvidence*` are deleted | Neither RPC comes back (map scope); `add-ffi-handler` reserves a removed Method | Keeping the Methods unreserved — a later Method could reuse the number |
| D-2 | Fields reserved by number and name, each message's comment updated: `IdentityClaim` 5 `confidence_grade_id`, 6 `argument`; `PromoteSubjectRequest` 5, 6 (same names); `ReconcilerOutcomeDetail` 14 `claim_confidence_key`, 16 `claim_confidence_offset`; `PromoteGraphAlignmentRow` 6 `score`, 7 `assessment`, 17 `via_bridge_type`, 18 `via_role`, 20 `reason`, 21 `reason_property_key`, 22 `reason_property_origin`, 23 `contested`; `PromoteGraphAlignmentAlternative` 3 `score`, 7 `assessment`, 8 `reason`, 9 `reason_property_key`, 10 `reason_property_origin`, 11 `via_neighbor_subject_id`; `ApplyPromoteGraphAlignmentRow` 4 `confidence_grade_id`, 5 `argument`, 7 `pin_ids`, 8 `default_pins` (6 is already reserved; its comment `// pairs: pin_ids` becomes `// pairs`); `ApplyPromoteGraphAlignmentWritten` 3 `pins`. The comments of `ApplyPromoteGraphAlignmentRequest` ("claims, pins, and bridges") and `PromoteGraphAlignmentRow.target` say what D-8 sends | Map scope plus D-7 and D-9; the alternative's `via_neighbor_subject_id` has no Go source once D-9 drops `Alternative.ViaNeighbor` and no Swift reader since S13-23 (D-5 there) | Leaving the alternative's `via_neighbor_subject_id` — a field Go can no longer fill |
| D-3 | Apply maps each wire row straight to `promote.BatchRow{SubjectID, Target, EntityID}` inside `withProjectCatalog` and calls `promote.SaveBatch`; no read runs before the write. `promotealign.ApplyRow`, `ResolvePins`, `PinOffer`, `PinOffers`, `LinePinID`, `BundlePinID` and `pins.go` go. The response's `Written` carries `entity` and `claim` | Map scope; Done "files accepted claims and bridges only" | A no-op `ResolvePins` kept for S14-16 — the working rule says no step keeps a shim |
| D-4 | `promote.BatchRow` is `SubjectID`, `Target`, `EntityID`; `Written` is `Entity`, `Claim`; `PinSet` and `pinSets` go. `applyRow`: Skip with a non-empty `EntityID` is `ErrInvalid`; New with a non-empty `EntityID` is `ErrInvalid`; a row whose Subject already has an accepted claim on the row's handle returns nil (an anchor), on any other handle or target `identityclaims.ErrAlreadyMember`. `BatchRow`'s comment loses its PinSets sentence; `SaveBatch`'s first line reads "files one Done in a single transaction: claims and bridges whose ends are both handles." The package doc loses its pin sentence and reads: "A step either mints a new handle of the Subject's type or joins an existing one, and files an accepted claim." `Save`'s comment loses "It pins nothing: SaveBatch is the one pin write." | The anchor rule today differs only for a row with pin sets, which no longer exist; the rest is the map scope | Keeping the `len(row.PinSets) == 0` term — it is always true |
| D-5 | `identityclaims.CreateInput` is `SubjectID`, `EntityID`, `Status`; `Claim` loses `ConfidenceGradeID` and `Argument`. `sqlInsert` names `(id, subject_id, entity_id, subject_type_id, status)`; `sqlColumns` is `id, subject_id, entity_id, subject_type_id, status`; `scanClaim` scans five columns; the audit create row loses the two keys; `InsertTx` loses the argument trim and the grade-length check; `mapInsert`'s comment drops "or grade". The two columns stay in the schema, nullable, and new claims leave them NULL | `CreateInput` stops taking them (map scope); `Claim`'s two fields would have no reader once `identityClaimProto` stops sending them, and a struct field filled from a column nobody reads is the shim the working rule forbids. This takes "identityclaims stops reading and writing the two columns" from S13-25's scope; the map moved it in the planning commit | Leaving the read and write to S13-25 — S13-24 would keep a field no code reads |
| D-6 | `promote.Input` is `SubjectID`, `EntityID`; `PromoteSubject`'s handler reads neither field and `identityClaimProto` sends `id`, `subject_id`, `entity_id`, `status`; `ListClaimConfidenceGrades` and its dispatch case go | Map scope | — |
| D-7 | `autoreconcile.Provenance` is `Credibility`, `Uncertain`. `Weak` is `p.Credibility < 0 \|\| p.Uncertain`; `Stronger` compares credibility, then certainty, and two equal ones are not stronger. `sqlLoadCandidates` drops the `claim_confidence_grades` join and its column; the scan drops `confidence`. The package doc's "Claim confidence and status have no edit path yet" reads "Claim status has no edit path yet" and its "and claim confidence and status" reads "and claim status"; `sqlLoadCandidates`' comment says "(standard)" for the default grade. `CacheVersion` is 21 with a comment line naming S13-24; `TestEnsureCatalogRebuildsStaleVersion` adds 20 to its stale list. `conclusiondetails.Outcome` loses `ClaimConfidenceKey`; `sqlOutcomes` drops `COALESCE(cg.key, '')`, the claim offset column, and the `identity_claims ic` and `claim_confidence_grades cg` joins (`ic` serves only `cg`). The handler maps neither field. Swift: `CatalogReconcilerOutcome` loses `claimConfidenceKey`, `claimConfidenceOffset` and `isLowConfidenceClaim`; `weakCauses` lists the Source then the transcription (its doc comment: "the Source, then the transcription"); `L10n.Conclusions.weakLowConfidenceClaim` and key `conclusions.weak.lowConfidenceClaim` go; `outcomeWeakBecause`'s comment lists "low-trust Source, uncertain transcription" (user, 2026-10-30: remove, no brief) | Map scope (tiebreak, `Weak`, `CacheVersion`); the outcome fields are the other reader of the same grade, missed by the map and decided by the user | Leaving the outcome fields to S13-25 — they would always read 0 in between |
| D-8 | What a row starts on is Go's. `graphalign` gains `type RowStart struct { Target Target; HandleID []byte; Ref string }` *(new)*, the wire's `target`, `handle_id` and `handle_ref`, and `func (r Row) Start() RowStart` *(new)*, what the page starts the row on: a placed row (`Target == TargetHandle`) whose `Assessment` is not `AssessNone` → `{TargetHandle, HandleID, HandleRef}`; a held New or Skip (`Target` is `TargetNew` or `TargetSkip`; only a held row has either) → `{Target: that target}`; an unplaced row (`Target == ""`) with `Reason != ReasonTaken`, `Assessment == AssessWeak` and at least one alternative → `{TargetHandle, Alternatives[0].HandleID, Alternatives[0].Ref}`; any other row → the zero `RowStart`. `proposalProto` sends `Start()`'s three fields as `target`, `handle_id`, `handle_ref`, and builds the row header for the started handle; `addHeaderID` adds the started id in place of `r.HandleID` (for a placed row they are the same; a weak start's id is also an alternative's). Swift's `draftTarget(proposal, menu:)` becomes: target `.handle` and `menu.first?.id == proposal.handleID` → `.handle(id:ref:title:)` of that item, else `.unset`; `PromoteFlow.Assessment` goes | Audit 2 B-26 (map scope). Each branch is today's Swift rule: a held row is kept from the page's previous decision, so its started target is never read (`makeRow` keeps `previous` when it was decided), and sending it as given keeps `PromoteTargetCode`'s wire unchanged. `graphalign`'s own `Row.Target` is unchanged, so `flagSharedHandles`, `wouldBeNew`, settling and the existing goldens see no change | Changing `proposal()` to set `Target` on a weak row — it would move the duplicate flags and the alike-New warning; a separate wire field — the page reads one target |
| D-9 | `graphalign` drops what only the reserved wire read: `RowFlags.Contested` and `cand.contested`; `Row.Comparisons`, `Comparison`, `comparisonsFrom`; `Row.ReasonProperty` and `strongestAgreement`; `Alternative.Assessment`, `.Reason`, `.ReasonProperty`, `.ViaNeighbor` and `ReasonWeak` (`alternative` returns `HandleID`, `Ref` and `Score`). `Alternative.Score`, `Row.Score`, `Row.Assessment`, `Row.Reason` (`via`, `decided`, `agrees`, `taken`, `empty`, unplaced) and the move codes stay. `productOrigin` in the handler goes, and so do the handler's `ViaBridgeType` and `ViaRole` lines (`graphalign.Via.Key` stays as the walk's edge that the `promotealign_test.go` goldens observe). This closes the wire half of audit 2 B-37 (`via_bridge_type`, `via_role`, both `score` fields). Swift's `PromoteReasonCode` is renamed `PromoteMoveReason` and keeps `neighborPlaced`, `neighborMoved`, `neighborReleased` and `unrecognized(String)`; the parity row's `enum:` names `PromoteMoveReason`, and so does the "states and reasons tolerate a value this build doesn't know" subtest's list in `api/ffi/swift_vocabulary_test.go`; `CatalogPromoteGraphAlignmentRow` loses `assessment` and `reason`; the "promote reasons" parity row lists the three move codes | Map scope ("drops `Flags.Contested` and the row-level comparisons"); settling and `Move` read none of them (`rg Comparisons core/graphalign` finds only `rows.go`); the alternatives' display fields lose their only reader with the wire | Keeping the alternatives' band and reason — computed per alternative per proposal for nobody. `Alternative.Score` stays: the alignment goldens read the ranking through it, as they do `Row.Score` |
| D-10 | `promotealign` loses `evidence.go`, `pins.go`, `values.go` and `evidence_test.go`, `pins_test.go`, `values_test.go`. `TestLargeSourceBudgets` loses the subtests "evidence warm", "evidence on a place row", "evidence under a small cap", `budgetEvidence`, `budgetPlaceEvidence` and `evidenceRequests`; "warm queries" keeps its Propose half (zero graph queries, one SELECT) and drops the evidence pass. `TestEvidenceWarmIssuesNoQuery` goes; `TestMemoFollowsCommit` and `TestMemoIsNotMutated` keep their Propose assertions and drop the `promotealign.Evidence` calls and `w.f.evidence(w.request())` lines, and `TestMemoIsNotMutated` drops its `len(first.Members) == 0` term. `TestProfilesLiveWithTheMatchIndex` reads the match index after `w.f.propose(w.f.graph(w.srcB), nil)` where it called `w.f.evidence(w.request())`, with "Evidence" in its two failure messages reading "Propose". `warmFixture.request` goes. `fixture.filedBirth` (with its comment) moves from `evidence_test.go` into `promotealign_test.go` before `evidence_test.go` is deleted: `newWarmFixture`, `TestCandidatesMemoFollowsCommits` and `perf_test.go`'s setup call it. `fixture.evidence` goes with `evidence_test.go`. `MemoSnapshot` loses `Members`. `inputs`' comment reads "what a proposal reads first". `api/ffi/handlers/promote_evidence.go` and `promote_evidence_test.go` go, except `evidenceFixture` and `newFoundRecordFixture`, which `TestProposeHoldsAFoundRecord` and "files a found record" use: they move to `promote_align_test.go`, the type renamed `foundRecordFixture` (STEP-1) | Map scope; the cap is still tested in `core/database/graphcache/evidence_graph_cap_test.go` | — |
| D-11 | The router's one rule `Rules.Role` does not already give moves to `subjectcompare` as `func (r Rules) Meta(kind vocabkey.SubjectType, p vocabkey.PropertyKey) vocabkey.PropertyKey` *(new)*, beside `Lead` and `Derived` in `core/subjectcompare/subjectcompare.go`: the meta a property's values compare under, built on `Role` the way `Lead` is: `lead, role, ok := r.Role(kind, p)`; `p` when `!ok` or `role == RolePeriodEnd`, else `lead`. Its doc comment says how the old router's other answers read from `Role`: a property counts under `lead` when `ok` (else under its own key, unweighted), its weights are `Weights[kind][lead]`, and its values compare through `Side` when `ok` and `role` is not `RolePeriodEnd`. No `Route` type and no compare over every Property are added, and `lifespan` stays unexported (user, 2026-10-30: move the router only). `core/claimevidence` is then deleted whole | Slice 7's working rule keeps the router as the starting point; `Role` already returns lead, role and found, so only the meta is new (round 2 M-8: a struct carrying `Role`'s zero value would read `RoleLead` for a property with no lead) | `CompareEvery` now — a function with no caller whose shape slice 7 has not decided; a `Route` struct — it repeats `Role` |
| D-12 | The registry loses `Registry.Evidence`, `defaultEvidence`, `evidence.go`, `evidence_test.go`, the `claimevidence` import and the `core/claimevidence` `layerPackages` entry; `matching-model.md` loses the three `evidence.*` rows. `kinship.Recording.Shown` and `TestRecordingShown` go. `core/database/canonlinks/canonlinks.go`'s package doc drops "claim evidence" from its readers; `graphalign.Bridge`'s comment on `Associations` reads "The layers never read them; the loader sets them for a later reader of the bridge Subjects (S13-29's shared loader)." `graphcache.Vocabulary.BindingOrder` goes; the three `vocab_test.go` tests read the provenencia person's order through `v.TypeOrder(id)` with the id from `subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)`, and keep their assertions. `canonlinks.Reader.Relationships` stays: S13-04's working rule calls it, and deleting it now means S13-04 re-adds it; STEP-8 drops it from the map's S13-24 list | Map scope, except `Relationships` | Deleting `Relationships` — the map's own working rule "One person's links, direct and implied, from one function" binds S13-04 to it |
| D-13 | The provenance columns have one owner, `core/database/autoreconciler` (user, 2026-10-30). It exports `ProvenanceColumns` *(new)*, the select-list fragment `COALESCE(sg.sort_order - (SELECT sort_order FROM source_credibility_grades WHERE key = 'standard' AND origin = 'provenencia'), 0), c.transcription_uncertain`; `ProvenanceJoins` *(new)*, `LEFT JOIN source_credibility_assessments sca ON sca.source_id = a.source_id LEFT JOIN source_credibility_grades sg ON sg.id = sca.credibility_grade_id`; and `func ProvenanceDest(p *autoreconcile.Provenance) []any` *(new)*, the scan targets `&p.Credibility, &p.Uncertain` in that order. Their doc comment states the aliases a query must bind: `c` for `citations` and `a` for `artifacts`; `sg` is then in scope for a caller's own extra column. `sqlLoadCandidates` and `conclusiondetails.sqlOutcomes` build their text from the two fragments (string concatenation in the `const` block) and scan through `ProvenanceDest`; `conclusiondetails` keeps its own `COALESCE(sg.key, '')` for `CredibilityKey`, and its join moves from `sca.source_id = src.id` to the shared `a.source_id` (the same Source: `src` joins on `a.source_id`). `conclusiondetails` imports `autoreconciler` (no cycle: `autoreconciler` imports neither `conclusiondetails` nor a package that does; `conclusionheaders_test.go`'s import of it is an external test package) | Map review M-5: two copies of the credibility offset drifted on claim confidence once; the next provenance input (Source quality, [`research-judgment-model.md`](../../../research-judgment-model.md) §5) is added in one place | A `core/database/provenance` package — a package for two constants and a scan helper whose rule is the reconciler's; leaving both copies — the user chose one owner |

**Rules this plan bends.** `plan-pr` **Size**: about 10 steps is the limit; this plan has 9 steps because STEP-5 and STEP-6 each carry their Swift half, and the work is about 13 steps' worth of change. Answer: the user kept one PR when offered a split in three (2026-10-30). No rule update follows.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `subjectcompare.Rules.Meta` *(new)* | func | `claimevidence.routeOf` (its meta rule, moved); `subjectcompare.Rules.Lead` | Built on `Role` the way `Lead` is; the rest of the old router is `Role` itself |
| `graphalign.Row.Start`, `graphalign.RowStart` *(new)* | func, type | Swift `PromoteFlow.draftTarget` (moved, not copied) | The rule reads `graphalign`'s band and reason, which stop crossing the wire |
| `autoreconciler.ProvenanceColumns`, `ProvenanceJoins`, `ProvenanceDest` *(new)* | SQL fragments, func | `conclusiondetails.sqlOutcomes`' copy of the same columns | Replaces that copy (D-13); the reconciler's rule owns how provenance is read |
| `TestRowStart` *(new)* | test | Swift draft tests in `PromoteFlowTests.swift` | They move with the rule |

**Builds on.** The proto's reserved blocks; `subjectcompare.Rules` methods (`Role`, `Lead`, `Derived`); `graphalign`'s pure row assembly.
**Deletes.** Everything in D-1 – D-12 and §4's rows for them: `core/claimevidence` (6 files); `promotealign` `evidence.go`, `pins.go`, `values.go` and their tests; `api/ffi/handlers/promote_evidence.go` and its test; `ListClaimConfidenceGrades`; `promote.PinSet`, `pinSets`; `core/matchregistry/evidence.go`, `evidence_test.go`; `Recording.Shown`; `Vocabulary.BindingOrder`; `graphalign`'s display-only fields and `ReasonWeak`; `productOrigin`; `conclusiondetails.sqlOutcomes`' own credibility subquery and `sca`/`sg` join text (D-13); Swift's `PromoteFlow.Assessment`, five `PromoteReasonCode` row-reason cases, `CatalogPromoteGraphAlignmentRow.assessment` and `.reason`, the claim confidence outcome fields, `L10n.Conclusions.weakLowConfidenceClaim` and its key. `identityclaims.PinTx` loses its last non-test caller and stays for S13-25 (Out). `claimconfidencegrades` has no non-test caller outside `Install` and the readiness check, which S13-25 removes.
**Next adds.** S13-25's migration drops columns nothing reads; S14-16 rewrites an Apply handler with no pre-write read; S13-26 starts from `subjectcompare.Rules.Role`, `Rules.Meta` and `Compare`.

### Reach

None — no new path into existing code. The started handle of a weak row reaches `proposalProto`'s header lookup, which already fetched headers for every alternative.

### Scale

At the map's Scale (a Source of 100 new people, the 41-Subject obituary): Done does one transaction with no evidence read before it (today `ResolvePins` reads only when a row asks for pins, which S13-23 stopped). `Row.Start` is O(1) per row. The reconciler's candidate load loses one join. `CacheVersion` 21 rebuilds every handle once on first open, the same cost as any version bump (`autoreconciler.Rebuild`).

### Edges

- **Assumptions.** `Start` relies on `proposal()` filling `Alternatives` best first (`sortedCands`) and on a placed row's band coming from its own score; T-1 holds both with the registry's bands. `draftTarget` relies on `menu` listing `proposal.handleID` first when the target is `.handle`, which `menu(for:orders:)` does; T-9 holds it for a weak start.
- **Neighbors.** A placed row with `AssessNone` (only a registry whose accept bar sits under the weak band; the shipped bar and weak band are both 2) → unset, as Swift did (T-1 case). An unplaced weak row that is also `PossibleDuplicate` through `flagAlikeNewRows` but not `taken` → starts on its candidate, as before (T-1 case). A held New → wire `new`, Swift ignores it for a decided row (T-9).
- **Removed inputs.** A claim's grade no longer weakens or orders a value: a value carried only by a low-confidence claim, which was weak and dropped against a stronger value, now competes as standard evidence (T-5, AC-2.1). This is the user's call (working rule "Identity Claims carry no confidence and no argument"); `CacheVersion` makes every cached outcome recompute.
- **Real data.** None — no table or seed drives a changed rule; the bands are the registry's three numbers above.
- **Untrusted size.** None — no new list or map read from catalog data.

### Schema

None — no catalog change. The two claim columns and the two tables stay until S13-25; new claims write NULL to both columns (D-5).

### Go core

| Package | Change |
| --- | --- |
| `core/database/promote` | D-4: `BatchRow`, `Written`, `Input` lose fields; `PinSet`, `pinSets` go; `applyRow`'s anchor rule |
| `core/database/identityclaims` | D-5 |
| `core/database/promotealign` | D-3, D-10: `pins.go`, `evidence.go`, `values.go` go |
| `core/claimevidence` | Deleted (D-11) |
| `core/subjectcompare` | `Rules.Meta` *(new)* (D-11) |
| `core/matchregistry` | D-12 |
| `core/kinship`, `core/database/graphcache` | D-12 |
| `core/autoreconcile`, `core/database/autoreconciler`, `core/database/conclusiondetails` | D-7; D-13 `ProvenanceColumns`, `ProvenanceJoins`, `ProvenanceDest` *(new)* |
| `core/graphalign` | D-8 `Row.Start` *(new)*; D-9 |

```go
// RowStart is what the page starts a row on: the wire's target, handle_id
// and handle_ref.
type RowStart struct {
	Target   Target
	HandleID []byte
	Ref      string
}

// Start is what the page starts the row on: a placed row's handle when it
// has a band; a held New or Skip as given; an unplaced row's best candidate
// when it is in the weak band and no other row took it; otherwise nothing.
func (r Row) Start() RowStart
```

Errors: unchanged codes. A Skip or New row with a record id stays `promote.invalid`.

### FFI

`api/proto/engine.proto` per D-1 and D-2. Handlers: `promote_batch.go` (D-3), `promote.go` (D-6), `conclusion_detail.go` (D-7), `promote_align.go` (D-8, D-9); `promote_evidence.go` deleted; `api/ffi/dispatch.go` loses `MethodGetPromoteClaimEvidence`, `MethodListClaimConfidenceGrades` and their two `case` arms. Regenerate with `./scripts/generate-proto.sh`.

### macOS

None — no view changes. Model and mapping only:

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Weak cause list | `ReconciledValueDisplay.weakCauses` | `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` | Two causes (D-7) |
| Draft rule | `PromoteFlow.draftTarget` | `macos/App/Features/Promote/PromoteFlow.swift` | Reads the started target (D-8) |
| Proposal row | `CatalogPromoteGraphAlignmentRow` | `macos/App/Platform/GenealogyStore.swift` | Loses `assessment`, `reason` (D-9) |
| Reason codes | `PromoteReasonCode` | `macos/App/Platform/ConclusionVocabulary.swift` | Three move codes (D-9) |

States: unchanged.

### Strings

Removed: `L10n.Conclusions.weakLowConfidenceClaim` = "low-confidence claim" (key `conclusions.weak.lowConfidenceClaim`). Changed comment only: `conclusions.outcome.weakBecause` lists "low-trust Source, uncertain transcription". No action changes what it covers.

## 7. Steps

**STEP-1 — The claim evidence RPC goes**
Files: `api/ffi/handlers/promote_evidence.go`, `api/ffi/handlers/promote_evidence_test.go`, `api/ffi/dispatch.go`, `api/ffi/dispatch_test.go`, `api/ffi/handlers/promote_batch_test.go`, `api/ffi/handlers/promote_align_test.go`. Skill: `add-ffi-handler`.
Do: move `evidenceFixture` (renamed `foundRecordFixture`, same fields; its comment reads "foundRecordFixture is a handle and, on a Source of its own, a person the shortlist does not reach from it.") and `newFoundRecordFixture` from `promote_evidence_test.go` into `promote_align_test.go`, where `TestProposeHoldsAFoundRecord` uses it. Then delete `promote_evidence.go` and `promote_evidence_test.go` (with `newEvidenceFixture`, `request`, `evidence`, `sectionOf`, `memberObservation`, `offerOf`, `pinnedOn`), the `case MethodGetPromoteClaimEvidence:` arm in `dispatch.go`, the `dispatch_test.go` row "claim evidence with no request", and the subtest "pins the evidence's offers" in `TestApplyPromoteGraphAlignment`. The subtest "files a found record" sends its row with subject, target and entity only (no `DefaultPins`) and asserts one written row on `f.handle`, without the `GetPins()` term. The `MethodGetPromoteClaimEvidence` constant goes in STEP-7 with the enum value.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes; `rg -n 'GetPromoteClaimEvidence\(|evidenceFixture|pinnedOn|DefaultPins: true' api` finds nothing.

**STEP-2 — Done files claims and bridges only; the evidence query goes**
Files: `api/ffi/handlers/promote_batch.go`, `core/database/promote/batch.go`, `core/database/promote/promote.go`, `core/database/promote/batch_test.go`, `core/database/promote/pins_test.go`, `core/database/promote/unpromote_test.go`, `core/database/promotealign/pins.go`, `promotealign_test.go`, `pins_test.go`, `evidence.go`, `evidence_test.go`, `values.go`, `values_test.go`, `perf_test.go`, `memo_test.go`, `export_test.go`, `promotealign.go`, `core/database/deleteimpact/pins_e2e_test.go`, `core/database/deleteimpact/conclusion_release_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: per D-3, D-4, D-10. The handler stops calling `ResolvePins` and stops reading `GetConfidenceGradeId`, `GetArgument`, `GetPinIds`, `GetDefaultPins` and setting `Pins` (the fields exist until STEP-7). Delete `core/database/promote/pins_test.go`, `TestSaveBatchPinsANeighborRecord`, `TestUnpromoteKeepsOtherMembersPins` in `unpromote_test.go` (it builds pins through `pins_test.go`'s `newPinFixture` and `f.join`; Unpromote releases pins through `deleteimpact.ReleaseClaim`, whose pin release `TestSubjectDeleteReleasesOwnClaimPins` and the `pins_e2e_test.go` tests keep covering until S13-25), and D-10's files and subtests. `promote.Input`'s two fields stay until STEP-4. In the two `deleteimpact` fixtures (S13-25's pin tests, which stay), the `promote.SaveBatch` row loses `PinSets`, and the same `writes.Call` then pins each Observation the set named on both claims, the grounding member's and the joined one, with `identityclaims.PinTx`, appending each returned change, so the fixture's pins and audit rows are what the backfill wrote before. `TestSaveBatchPinsANeighborRecord` is the only `batch_test.go` test that reads `PinSets` or `Pins`. Add T-2 and T-3. Update `inputs`' comment.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promote/... ./core/database/promotealign/... ./core/database/deleteimpact/... ./api/ffi/...` passes with T-2, T-3 and T-10, and with `TestSubjectDeleteReleasesOwnClaimPins`, `TestPinnedObservationDeleteEndToEnd` and `TestPinnedMemberDeleteEndToEnd` unchanged; `TestLargeSourceBudgets` runs "warm queries" with the Propose assertions only; `rg -n 'PinSet|ResolvePins|PinOffers|ApplyRow|DefaultPins|PinIDs|promotealign\.Evidence|EvidenceRequest|EvidenceValues|membersMemoKey|acceptedClaims' core api -g '!api/proto/**'` finds nothing.

**STEP-3 — `claimevidence` goes; its router moves to `subjectcompare`**
Files: `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/subjectcompare_test.go`, `core/claimevidence/` (deleted), `core/matchregistry/registry.go`, `evidence.go`, `evidence_test.go`, `guard_test.go`, `docs/matching-model.md` (the three knob rows), `core/kinship/kinship.go`, `core/kinship/kinship_test.go`, `core/database/graphcache/vocab.go`, `core/database/graphcache/vocab_test.go`. Skill: `add-match-rule`.
Do: per D-11, D-12. Add `Rules.Meta` and T-4, then delete the package. Remove only the three `evidence.*` knob rows here; the prose in `matching-model.md` is STEP-8's.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/subjectcompare/... ./core/matchregistry/... ./core/kinship/... ./core/database/graphcache/...` passes, with T-4, `TestKnobInventoryMatchesRegistry` and `TestLayerPackages`; `rg -n 'claimevidence|\.BindingOrder\(|\) Shown\(' core api` finds nothing.

**STEP-4 — Claims carry no confidence or argument in Go**
Files: `core/database/identityclaims/identityclaims.go`, `identityclaims_test.go`, `core/database/deleteimpact/pins_e2e_test.go`, `core/database/autoreconciler/autoreconciler_test.go`, `core/database/promote/promote.go`, `core/database/promote/batch.go`, `promote_test.go`, `api/ffi/handlers/promote.go`, `api/ffi/handlers/promote_test.go`, `api/ffi/dispatch.go`. Skill: `add-ffi-handler`, `add-audit-scope` (the create row's keys).
Do: per D-5, D-6. In `core/database/deleteimpact/pins_e2e_test.go` the grounding `promote.Input` loses `Argument` and the `member.Argument` check becomes a check that the member's claim is still accepted. In `autoreconciler_test.go` the subtest "a low-confidence claim's different given name is weak" stops passing the grade through `promote.Input` and sets it with `UPDATE identity_claims SET confidence_grade_id = ? WHERE subject_id = ?` (its expectation changes in STEP-5); the random test's `grades` slice and its `in.ConfidenceGradeID = f.grade(g)` lines go. In `identityclaims_test.go` the "create" case loses its `cl.ConfidenceGradeID != nil` term, and the `claimconfidencegrades` import goes with the deleted subtest, its only user. Delete the subtests "confidence grade and argument stored" (`identityclaims_test.go`) and "confidence grade and argument land on the claim on both paths" (`promote_test.go`); `TestPromoteSubjectJoin` sends no grade or argument and loses its "bad grade id" case; the `ListClaimConfidenceGrades` test call goes; the `case MethodListClaimConfidenceGrades:` arm goes. The handler stops reading and sending the two fields while they still exist.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/identityclaims/... ./core/database/promote/... ./core/database/deleteimpact/... ./core/database/autoreconciler/... ./api/ffi/...` passes; `rg -n 'ConfidenceGradeID|ConfidenceGradeId|GetArgument\(\)' core/database/identityclaims core/database/promote api/ffi/handlers/promote.go api/ffi/handlers/promote_batch.go` finds nothing.

**STEP-5 — The reconciler stops reading claim confidence, and provenance has one reader**
Files: `core/autoreconcile/autoreconcile.go`, `pipeline_test.go`, `names_test.go`, `core/database/autoreconciler/autoreconciler.go`, `autoreconciler_test.go`, `core/database/conclusiondetails/conclusiondetails.go`, `conclusiondetails_test.go`, `api/ffi/handlers/conclusion_detail.go`, `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Features/Conclusions/ReconciledValueDisplay.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: per D-7, then D-13: add the three exports to `autoreconciler.go` beside `sqlLoadCandidates`, build `sqlLoadCandidates` and `sqlOutcomes` from them, and scan through `ProvenanceDest`; add T-12. In `pipeline_test.go` delete `lowClaim`, the "a low-confidence claim is weak" row, the two `Stronger` rows and the `Weak` map entries that name `ClaimConfidence`, and drop `ClaimConfidence` from the random `Provenance`; in `names_test.go` delete "a low-confidence claim drops" and drop it from the random `Provenance`. In `autoreconciler_test.go` change the subtest's expectation per T-5, and the comment above the provenance table near line 918 drops "claim confidence" from its list. The handler stops mapping `ClaimConfidenceKey` and `ClaimConfidenceOffset`; GoStore stops mapping them.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/... ./core/database/autoreconciler/... ./core/database/conclusiondetails/... ./api/ffi/...` passes with T-5, T-6, T-12; `rg -n "key = 'standard'" core/database -g '*.go' -g '!*_test.go'` finds only `core/database/autoreconciler/autoreconciler.go`; `rg -n 'ClaimConfidenceKey|ClaimConfidenceOffset|\.ClaimConfidence\b|ClaimConfidence:|claimConfidence|isLowConfidenceClaim|lowConfidenceClaim|LowConfidenceClaim' core api macos/App macos/ProvenenciaTests -g '!**/Generated/**' -g '!api/proto/**'` finds nothing (the `claimconfidencegrades` package, its seed check and `CodeClaimConfidenceInvalid` are S13-25's and do not match); `python3 scripts/check-localizable-xcstrings.py` passes; Xcode tests pass with T-7.

**STEP-6 — Go says what a row starts on**
Files: `core/graphalign/types.go`, `rows.go`, `align_test.go`, `api/ffi/handlers/promote_align.go`, `promote_align_test.go`, `core/database/promotealign/promotealign_test.go`, `macos/App/Features/Promote/PromoteFlow.swift`, `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Platform/ConclusionVocabulary.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`, `PromoteModelTests.swift`, `api/ffi/swift_vocabulary_test.go`. Skills: `add-match-rule`, `add-swift-test`.
Do: per D-8, D-9. Add `Row.Start` and T-1 first, then point `proposalProto` at it (T-8), then delete D-9's fields. In `align_test.go` and `promotealign_test.go` only the term on a deleted field goes and the rest of each condition stays: `Alternatives[0].Assessment != graphalign.AssessNone` becomes `Alternatives[0].Score >= cfg.Alignment.Bands.Weak` (the test's own config, or `matchregistry.Default().AlignConfig()` where it has none), so a below-bar alternative is still checked to sit below the bar; a `.ViaNeighbor` term goes; a `Flags.Contested` term and the table column that feeds it go; a `Comparisons` or `ReasonProperty` assertion goes; `.Score` terms stay. `TestProposeSendsContested` becomes T-8. Swift: `draftTarget` per D-8; the draft tests per T-9, and every other `row(…)` call in `PromoteFlowTests.swift` and `PromoteModelTests.swift` drops its `assessment:` argument with no other change (the new `draftTarget` reads only `target` and `handleID`, so a row's start is unchanged); `FakeStore.proposePromoteGraphAlignment` stops setting `assessment` and `reason`; `PromoteReasonCode` per D-9 (renamed `PromoteMoveReason`; `rg -n PromoteReasonCode macos api` finds nothing after) and `reasonCodesParseThroughTheWireVocabulary` lists the three move codes; the parity row per D-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/... ./core/database/promotealign/... ./api/ffi/...` passes with T-1, T-8 and the "promote reasons" parity row; `rg -n 'Contested|Comparisons|ReasonProperty|ReasonWeak\b|strongestAgreement|productOrigin' core/graphalign api/ffi/handlers` finds nothing; `rg -n 'draftTarget|Assessment' macos/App/Features/Promote` shows `draftTarget` only; Xcode tests pass with T-9.

**STEP-7 — The wire: Methods deleted, fields reserved, proto regenerated**
Files: `api/proto/engine.proto`, `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/dispatch.go`. Skill: `add-ffi-handler`.
Do: per D-1, D-2. Remove the two `Method…` constants from `dispatch.go`. Run `./scripts/generate-proto.sh`.
Check: `CGO_ENABLED=1 go vet -tags fts5 ./api/... ./core/...` and `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` pass; Xcode tests pass with no Swift edit in this step outside `macos/App/Platform/Generated/`; `git diff --stat -- macos` lists only `engine.pb.swift` for this step; AC-4.1's `rg` finds nothing.

**STEP-8 — Docs and skills**
Files: `docs/research-judgment-model.md`, `docs/interpretation-layer-data-model.md`, `docs/conclusion-reconciliation.md`, `docs/matching-model.md`, `docs/promote-graph-alignment.md`, `docs/deployment-plan/archive/spike-11/comparison-layers.md`, `.agents/skills/add-match-rule/SKILL.md`, `.agents/skills/add-seeded-vocabulary/SKILL.md`, `.cursor/rules/match-rules.mdc`, `.agents/skills/ship-pr/SKILL.md`, `.agents/skills/review-pr/lenses/security.md`.
Do:
- `research-judgment-model.md` §4: Identity Claims carry no confidence (working rule, 2026-10-29); Reconciliation Claim confidence is open, not decided; line 5's and line 25's claim editor sentence say so; line 35's table row reads "Columns on Reconciliation Claims (open)"; the `identity_claims` DDL lines in §4 are marked "dropped by S13-25".
- `conclusion-reconciliation.md` §4: the "Identity Claim confidence" input row (line 69) goes; line 72 reads "A candidate is **weak** if either is below default. One candidate is **stronger** than another by comparing credibility, then certainty." and its first sentence names credibility only; line 40's "credibility, certainty and confidence" reads "credibility and certainty"; line 214's "status or confidence" reads "status". The Promote package row (line 287) names no removed package and stays.
- `matching-model.md`: line 5's last sentence reads "The claim evidence sheet and the registry's `evidence` section are gone (S13-23, S13-24)."; line 122's `evidence` sentences go.
- `promote-graph-alignment.md`: §2 principle 5 says the core returns where each row starts (`Row.Start`), the page does not draft; §6 (line 137) drops claim confidence from provenance (graph alignment's `Subject.Provenance` is 1, `core/database/promotealign/layer.go`); §8's "One Evidence request reads each graph once" goes; §9 says Done files claims and bridges and reads nothing first; §12's "a warm Evidence request reads nothing" goes.
- `interpretation-layer-data-model.md` line 693: "cached with the vocabulary as `graphcache.Vocabulary.BindingOrder`" reads "cached with the vocabulary (`graphcache.Vocabulary.TypeOrder`)".
- `comparison-layers.md` line 471: "Promote's claim evidence (`core/claimevidence`)" reads "the claim evidence sheet (removed in S13-24)".
- `add-match-rule`: the description's package list without `core/claimevidence`; line 30's "for the score and the sheet alike" reads "for the score"; line 32's "and in claim evidence" goes; line 33 keeps `Reader.Relationships` (S13-04 calls it) and drops the sentence that begins "Claim evidence's canon side reads" and `kinship.Recording.Shown`; line 35's bullet becomes one sentence: "`subjectcompare.Rules.Role` says what a Property counts under and how, and `Rules.Meta` the meta its values compare under; `Meta` moved from the claim evidence policy (S13-24), and the two are the starting point for slice 7's comparison."; line 78 reads "A move reason is layer 6's: a `Reason` constant in `core/graphalign/types.go` used as `Move.Reason`, its `PromoteMoveReason` wire case with its row in `TestSwiftVocabularyMatchesGo`, its code in the `PromoteGraphAlignmentRow.move_reason` comment in `api/proto/engine.proto` (regenerate), its `PromoteFlow.Reason` case, and its `PromoteModel.reasonText` arm with an L10n line. A row reason stays in Go."; line 100's test command drops `./core/claimevidence/...`.
- `add-seeded-vocabulary` line 58: the `claimevidence` `seedOrder` sentence goes.
- `.cursor/rules/match-rules.mdc`, `ship-pr`, `.agents/skills/review-pr/lenses/security.md`: drop `core/claimevidence` from the path lists.
Check: `rg -n -i 'claim evidence|pin set|claim-confidence' core api -g '!api/proto/**' -g '!core/database/migrations/**' -g '!core/database/claimconfidencegrades/**' -g '!core/onboarding/**' -g '!core/database/deleteimpact/**' -g '!core/database/identityclaims/pins.go' -g '!core/database/effects/**'` finds nothing (migrations are history; the others are S13-25's: `effects`' `claimEvidenceGraph` is the pin table's audit scope, which S13-25 removes with the table); `rg -n -i 'claim confidence' core api -g '!api/proto/**' -g '!core/database/migrations/**' -g '!core/database/claimconfidencegrades/**' -g '!core/onboarding/**' -g '!core/database/effects/**' -g '!core/database/autoreconciler/autoreconciler_test.go'` finds nothing (the test's `f.grade` comment stays true for T-5; its line-918 list comment drops "claim confidence" in STEP-5); `rg -n 'claimevidence|ListClaimConfidenceGrades|Vocabulary\.BindingOrder|BindingOrder\(\)|Recording\.Shown|ResolvePins|PinOffers|evidence\.(Checked|ByEye|Defining)' .agents .cursor docs -g '!docs/deployment-plan/**'` finds nothing; `python3 scripts/check-guidance-names.py` passes.

**STEP-9 — Plan, map and completed**
Files: `docs/deployment-plan/spike-13/completed.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (map row), this plan (status, Landed note).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-24-the-pin-engine-goes.md` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/graphalign/align_test.go` `TestRowStart` *(new)*, table-driven | placed `e1`, strong → handle `e1`; placed `e1` with `AssessNone` → ""; held New → `new`, nil; held Skip → `skip`, nil; unplaced weak, alternatives `[e9, e2]` → handle `e9` with `e9`'s ref; unplaced weak, `PossibleDuplicate` from alike-New, not taken → handle `e9`; unplaced weak, `ReasonTaken` → ""; unplaced medium → ""; unplaced strong → ""; unplaced `AssessNone` with alternatives → ""; unplaced weak, no alternatives → ""; `ReasonEmpty` → "" | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-8 |
| T-2 | Go | `core/database/promote/batch_test.go` `TestSaveBatchFilesClaimsAndBridgesOnly` *(new)* | Handle `e1` whose member B records A's birth date; Done with A → `e1`, C New, D Skip → two claims, accepted; one bridge filed; `SELECT COUNT(*) FROM identity_claim_evidence` is 0; B's claim row unchanged | AC-1.1, D-4 |
| T-3 | Go | `core/database/promote/batch_test.go` `TestSaveBatchAnchorsARowOnItsOwnHandle` *(new)* | A already accepted on `e1`; Done with A → `e1` → no write, no error; A → `e2` → `identityclaims.ErrAlreadyMember`; Skip with an entity id → `promote.ErrInvalid` | AC-1.2, D-4 |
| T-4 | Go | `core/subjectcompare/subjectcompare_test.go` `TestMeta` *(new)*, table-driven, with `matchregistry.Default().Subject` | person `name` → `name`; person `gender` → `sex_at_birth`; event `end_date` → `end_date`; event `start_date` → `start_date`; event `date` → `date`; person `religion` → `religion`; person `lifespan_span` → `lifespan_span` | D-11 |
| T-5 | Go | `core/database/autoreconciler/autoreconciler_test.go` subtest "a claim's grade does not weaken its values" (replaces "a low-confidence claim's different given name is weak") | Two members with different given names; one claim's `confidence_grade_id` set to `low_confidence` by `UPDATE identity_claims` → the outcomes equal those of the same fixture with no grade, and neither outcome's reason is `weak` | AC-2.1, D-7 |
| T-6 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestEnsureCatalogRebuildsStaleVersion` | Its `stale` list gains 20 → `NeedsRebuild` true, and after `EnsureCatalog` `StoredVersion` is `CacheVersion` (21) | AC-2.3, D-7 |
| T-7 | Swift | `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift` `weakSaysWhichEvidenceIsWeak` | A weak outcome with credibility −2 and uncertain → causes `["low-trust Source", "uncertain transcription"]`; a weak outcome with neither → "weak" | AC-2.2, D-7 |
| T-8 | Go | `api/ffi/handlers/promote_align_test.go` `TestProposeSendsWhereARowStarts` (replaces `TestProposeSendsContested`) | Negative sex Observation against a promoted record, unfixed → `target` ""; fixed → `handle` on the record; positive → `handle`; and `TestProposalProtoStartsAWeakRow` *(new)*, built like `TestProposalProtoSendsMove`: `subjectFixture`, one person promoted with `promoteOK` (record `h`, ref `r`), then `proposalProto` on a `graphalign.Proposal` with one row `{SubjectID: s, Kind: vocabkey.Person, Assessment: graphalign.AssessWeak, Reason: graphalign.ReasonUnplaced, Alternatives: [{HandleID: h, Ref: r}]}` → `target` `handle`, `handle_id` `h`, `handle_ref` `r`, a person header for `h`; the same row with `Reason: graphalign.ReasonTaken` → `target` "", no row header | AC-3.1, AC-3.2, AC-3.4, D-8 |
| T-9 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aRowStartsWhereTheProposalSays` (replaces `aMediumMatchStartsOnItsHandle`, `aWeakMatchStartsOnItsHandle`, `anUnplacedRowNeverDraftsNewOrSkip`, `aRowWhoseRecordWasTakenStartsUnset`, `aWeakCandidateBelowTheBarStartsSelected`) | Target `.handle`, `handleID` `e9`, alternatives `[e9, e2]` → row on `e9`, menu `[e9, e2]`; target `.unplaced` with alternatives → `.unset`; target `.new` → `.unset`; and `aRowThatLosesItsRecordSaysNoRecordIsSuggested`, `anUnmatchedRowStartsUnselected` and `decidedRowsSurviveAReproposal` pass with the `assessment` parameter removed from `row(…)` | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-8 |
| T-10 | Go | `api/ffi/handlers/promote_batch_test.go` `TestApplyPromoteGraphAlignment` | Existing cases pass with rows of subject, target and entity only; each `Written` has `entity` and `claim` | AC-1.3, D-3 |
| T-12 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityEvidence` | The weak `low_trust` outcome's `Provenance` equals `autoreconcile.Provenance{Credibility: -1, Uncertain: false}` (`low_trust` sort order 1 minus `standard` 2) exactly, and a record cited with `transcription_uncertain` under a `standard` Source has `Provenance{Credibility: 0, Uncertain: true}` | AC-2.4, D-13 |
| T-11 | Go | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry`; `core/matchregistry/guard_test.go` `TestLayerPackages` | No `Evidence` registry field, no `evidence.*` row, no `core/claimevidence` entry → both pass | AC-4.3, D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestLargeSourceBudgets`, `TestMemoFollowsCommit`, `TestKnobInventoryMatchesRegistry`, `TestLayerPackages`, the S11-39 alignment goldens in `align_test.go`, `swift_vocabulary_test.go`, `PromoteFlowTests`.

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
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: research-judgment-model.md §4, conclusion-reconciliation.md §4, promote-graph-alignment.md §2, §6, §8, §9, §12, matching-model.md (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source of several people, some matching existing records. Open Promote: rows start where they started before this PR (a suggested record, a weak candidate drafted, others unset). Press Done: the toast counts the promoted subjects and connections; `sqlite3 <project>/catalog.sqlite 'SELECT COUNT(*) FROM identity_claim_evidence'` is unchanged.
2. Run AC-4.1's `rg`: no hit (AC-4.1). In `api/proto/engine.proto`, `enum Method` reserves 89 and 110 with their names, and each message in D-2 reserves its numbers and names (AC-4.2).
3. Open a Person page whose birth date has a weak value from a `low_trust` Source: the outcome reads "weak · low-trust Source". No outcome reads "low-confidence claim".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A row starts somewhere different after the move | T-1 mirrors every branch of the old Swift rule; T-9 keeps the Swift outcomes |
| A deleted test was the only cover for something that stays | D-10 names the graph cap's other test; STEP-2 keeps the Propose half of "warm queries" |
| The reconciler outcome changes on real projects | Only for values from graded claims, graded before S13-23 (user's call); `CacheVersion` recomputes them |

**Stop and report if:**

- `go build ./core/...` reports an import cycle once `conclusiondetails` imports `autoreconciler` (D-13's premise).

- `core/graphalign` reads `Comparisons`, `Contested`, `ReasonProperty` or an alternative's display field outside `rows.go`'s assembly (D-9's premise).
- An S11-39 golden in `align_test.go` changes a row's `Target`, `HandleID` or ranking after STEP-6 (D-8 says `graphalign`'s rows are unchanged).
- `rg -n 'claimevidence' core` finds an importer other than `core/database/promotealign` and `core/matchregistry`.
- A non-test caller of `canonlinks.Reader.Relationships` other than `core/database/promotealign/evidence.go` exists, or none is planned in S13-04's scope entry (D-12).
- `identityclaims.Claim`'s `ConfidenceGradeID` or `Argument` is read outside `identityclaims` and `api/ffi/handlers/promote.go` (D-5).
- Xcode needs a Swift edit outside `macos/App/Platform/Generated/` in STEP-7.
- `TestPragmaHonesty` or an effects completeness test fails because the audit create row lacks the two column keys (D-5).

## 12. Open questions

None.

## 13. PR

**Title:** S13-24 — The pin engine goes
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-24-the-pin-engine-goes.md`.
