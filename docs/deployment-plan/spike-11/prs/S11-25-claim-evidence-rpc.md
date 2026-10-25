# S11-25 — Claim evidence RPC and pin policy

**Status:** Landed in [#517](https://github.com/mendahu/provenencia/pull/517) (2026-10-21)

**Landed 2026-10-21.** Differences from the plan: STEP-4 to STEP-8 are one commit, as the plan says. T-7's fixture people also share a name, so its response has a name line beside the `sex_at_birth` line and the test asserts the `sex_at_birth` pins and offers; a merged handle is not exercised at handler level. `DateValueDisplayTests.aComparisonFormatsABoundInTheLocale` went with `ComparisonValueDisplay.swift` (and its `project.pbxproj` lines). Dead after the deletions and removed: `graphalign.comparisonsFrom`'s `metas` parameter, the `provisionalOnly` test helper and `FakeStore.recordPin` (the fake pins nothing; Go owns which Observations a pin names). `TestSaveBatchOneHopPin` became `TestSaveBatchPinsANeighborRecord` with its refusal cases removed, `TestProposeKeepsTheLifespanOffTheSheet` became `TestProposeComparesTheLifespan`, and `TestProposeNegativeDateConflictLine` was deleted with its only assertion. `ValueCause`'s Swift case for the wire code `case` is `letterCase`. `PromoteModel.loadEvidence(subjectID:handleID:)` is synchronous and starts its own task so it captures the generation first. `ResolvePins` loads the Source inputs lazily, when a row first needs pins. It also deletes `autoreconcile.Compatible` with its tests, its last caller being gone. Review round 1: `evidenceValueProto` no longer builds its own `ConclusionValue` shapes; it fills a `conclusiondetails.Value` and calls `conclusionValueProto`, which now takes a term by key without an id (supersedes D-1's separate encoder). The Argument field's hint is removed and `promote.sheet.pinRelationship` is added for the relationship checkbox's spoken label (user, 2026-10-21), and the checkbox's identifier is `promote.sheet.relationship.<neighborSubjectID>.<key>` (D-14 named the key alone, which two neighbors can share), superseding the Strings section's "Removed: none". `promote.Result.Pins` is deleted; `ResolvePins` drops every caller-supplied pin set; `LinePinID` and `BundlePinID` take the evidence types. No stop condition fired.
**Spike / slice:** Spike 11 · Slice 5 — Claim evidence
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the shipped sheet reads the new lines in today's columns and cards; S11-D2 gates the new layout in S11-20)
**Board:** —
**Depends on:** S11-19 (landed [#513](https://github.com/mendahu/provenencia/pull/513), with its §14 amendment)
**Blocks:** S11-20
**Branch:** `s11-25-claim-evidence-rpc`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (`core/claimevidence` is a guarded layer package; this PR extends the skill), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (kit components only)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/catalog-session.mdc`
**Context:** [`claim-evidence.md`](../claim-evidence.md) (Lines, Writing pins, Evidence through relationships, Retires); [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3 and §9; [S11-19 plan](S11-19-claim-evidence-query.md) (D-3, D-12, D-13, §14) and its [ship log](S11-19-claim-evidence-query.ship.md) ("Apply memo check"); working rules "The pin policy has one owner, in Go", "One wire vocabulary", "Build, then migrate", "High tolerance for churn", "A separate, lazy query per row"; S11-13 ship log MNT-5, PRF-2
**Planned:** 2026-10-21 against `main` @ `10d3752`
**Review:** [`S11-25-claim-evidence-rpc.review.md`](S11-25-claim-evidence-rpc.review.md) — 2 rounds: Fits with changes / Ready with changes; every finding fixed or declined with evidence

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

The Promote sheet reads a row's claim evidence from S11-19's query, through a new lazy RPC, when the researcher opens it. Done pins what the evidence offers: for a row whose sheet was never opened, Go pins the lines and relationships that start checked; for an opened row, it pins the ones the researcher left checked. Each pin lands on the new claim and backfills the claim of every member it names, so joining an entity with five named members pins all five. The proposal stops carrying sheet lines, and the old exhibit code (`attachExhibits`, `pairExhibits`, `withRowLines`, `refLine`, proposal field 9) is deleted.

**Why now.** S11-19 built the query with no caller ("Build, then migrate"). This PR is the migrate half. S11-20 rebuilds the sheet on the RPC this PR adds.

**What stays.** The sheet's layout: the columns (Pin, Compared, Here, the target, Result, Weight), the "This subject" card and one "Through <name>" card per neighbor, the conflict callout, and the claim fields. All existing strings. The proposal's rows, bands, reasons, alternatives, contested and move flags. Bridge filing at Done. The pin tables (`identity_claim_evidence`, primary key `(identity_claim_id, observation_id)`). The rule that a pin's Observations land on the new claim and on the member's claim ([`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §9).

## 2. User stories

Conventions. Go fixtures: `promotealign_test.go`'s `newFixture`, `newSource`, `person`, `datedBirth`, `participation`, `promote`, `join`; `evidence_test.go`'s `filedBirth`; `memo_test.go`'s `newWarmFixture`; and `namedMembers` *(new, T-3)*. A pin id is `l:<in>:<mem>` for a line member or `b:<neighbor>:<key>:<bridge>` for a relationship bundle, each id written with `database.UUIDString` and the key with `vocabkey.BridgeKey.String()` (D-3). "Offers" are `promotealign.PinOffers(e)`.

**US-1.** As the researcher, when I open a row's sheet, I see every member's evidence for that row's target, with the lines that match already checked.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `namedMembers(5, "John Smith")`: a handle with five accepted members, each `name` "John Smith", on Sources "Census 1851", "Census 1861", "Census 1871", "Census 1881", "Census 1891"; a Source person "John Smith"; the evidence for that row | One section `name` with one line: relation `same`, outcome `agree`, checked, pinnable, five pins in member order, each with its member Subject, Observation, Source id, value (a `ConclusionValue` name "John Smith") and id `l:<in>:<mem>`; five offers, all checked; five Source titles |
| AC-1.2 | `newWarmFixture` with the birth row on the filed birth's handle in the request | One relationship: neighbor the Source birth, `bridge_type` `participation`, `term` `subject`, outcome `agree`, checked, one bundle `b:<birth>:<key>:<bridge>`, and one section `date` with one checked line; the bundle's offer and the date line's offer are checked |
| AC-1.3 | AC-1.2 with the birth row's target New, Skip, or left out of the request (the birth is not promoted) | No relationship |
| AC-1.4 | AC-1.2 with the birth already promoted onto the birth handle, and the request naming no neighbor | One relationship: an already filed Subject is a neighbor without being named (D-6) |
| AC-1.5 | A request with an unknown Source, a Subject not on it, an empty handle, or a merged handle | Error code `promote.invalid` |
| AC-1.6 | Opening a sheet for the first time (Swift) | The row's pins are the response's checked offer ids |
| AC-1.7 | Opening it again after the researcher unchecked a line, then set a neighbor to Skip | The line stays unchecked; pins the new response no longer offers are dropped |
| AC-1.8 | The sheet while the RPC is in flight; after it fails | Skeleton bars in the lines area; then a danger callout with the error message, and the claim fields still editable |

**US-2.** As the researcher, when I press Done without opening a row's sheet, the row is pinned exactly as the sheet would have drafted it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.1's page; Done with the row never opened | The new claim pins the incoming name Observation and all five member name Observations; each of the five members' accepted claims gains the incoming Observation and its own |
| AC-2.2 | AC-1.2's page; Done with the person row never opened | The person's new claim pins the bundle (the Source participation bridge's Observations, the Source birth's `event_type` Observation, the filed bridge's Observations, the filed birth's `event_type` Observation) and both date Observations; the filed person member's claim gains every Observation of the bundle; the filed birth member's claim gains both date Observations |
| AC-2.3 | AC-1.1's page; open the sheet, change nothing, Done | The same pins as AC-2.1 |
| AC-2.4 | A Source person whose `sex_at_birth` is a negative "male", against a member who is male | The line is a conflict, not pinnable, and has no offer; Done pins nothing for it, opened or not |
| AC-2.5 | Any row | Its argument starts empty; Done files what the researcher typed, or nothing |

**US-3.** As the engine, Apply pins only what the evidence offers, from the page's final targets.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ResolvePins` with an opened row whose ids include one id the evidence no longer offers, and one conflict line's id | Both are dropped; the rest pin; no error |
| AC-3.2 | A New or Skip row that sends pin ids, or a row with `DefaultPins` and pin ids | `promote.ErrInvalid` |
| AC-3.3 | A row whose Subject is already an accepted member of the Source graph | `ResolvePins` gives it no pins; `SaveBatch` writes nothing for it, as today |
| AC-3.4 | `ResolvePins` over a page with three handle rows, after a `Propose` of that page | `loadInputs` runs once (one call in `ResolvePins`); no graph query is issued |
| AC-3.5 | `promote.SaveBatch` with a `PinSet` whose `Member` has no accepted claim, or an id that is not 16 bytes | `promote.ErrInvalid`, nothing written |
| AC-3.6 | Two pin sets name one Observation on one claim | It is pinned once |

**US-4.** As a contributor, the evidence's codes cross the wire once and Swift parses them in one place.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `TestSwiftVocabularyMatchesGo` | Swift's `EvidenceOutcome`, `EvidencePropertyCause`, `ValueRelation`, `ValueCause`, `SegmentCause` list exactly Go's codes (D-10), each with an `unrecognized` case |
| AC-4.2 | `rg -n 'attachExhibits|pairExhibits|withRowLines|refLine|bridgeNeighbors|loadExhibitObservations|exhibitObs|graphalign\.Exhibit|pairClaim|draftedPins|draftedArgument|CatalogPromoteGraphAlignmentComparison|CatalogPromoteGraphAlignmentPair|ApplyPromoteGraphAlignmentPair|PromoteGraphAlignmentComparison|promote\.Pair\b' core api macos --glob '!**/Generated/**' --glob '!**/engine.pb.go'`, and `rg -n '"(agree|partial|conflict)"' macos/App/Features/Promote` | Both print nothing. Before this PR the first prints hits in `exhibits.go`, `promote.go`, `engine.proto`, `GoStore.swift` and others |

## 3. Scope

| In | Covered by |
| --- | --- |
| `claimevidence`: `Pin.Value`, `Bundle.Bridge`, `Relationship.Kind`, `Evidence.Kind` | STEP-1, D-2 |
| Typed evidence values (`promotealign`) | STEP-2, D-4 |
| `Evidence` split into `loadInputs` plus `evidence`; neighbors through `mergeFixed`; pin ids and offers; `ResolvePins` | STEP-3, D-3, D-6, AC-1.4, AC-3.1–AC-3.4 |
| `promote.PinSet` replaces `promote.Pair`; `Save` loses pins; `pairClaim`, `pinTargets` and the one-hop SQL deleted | STEP-3, D-5, AC-3.5, AC-3.6 |
| RPC `GetPromoteClaimEvidence` *(new)*; `fixedFromProto`; Apply row takes `pin_ids` and `default_pins`; proposal field 9 and the comparison and pair messages removed | STEP-4, D-1, D-7, D-8, AC-1.1–AC-1.5, AC-2.1–AC-2.4 |
| Go deletions: `exhibits.go` (but `acceptedClaims`), `exhibits_pin_test.go`, `graphalign.Exhibit`, `Row.Exhibits`, `Comparison.ValueType`, `inputs.q`, the exhibit assertions in `promotealign_test.go` | STEP-5, D-9, AC-4.2 |
| Wire vocabulary in `ConclusionVocabulary.swift`, under `swift_vocabulary_test.go`; the `add-match-rule` step for a code | STEP-6, STEP-10, D-10, AC-4.1 |
| Swift store: types, `getPromoteClaimEvidence`, `GoStore`, `FakeStore`; the batch row shape | STEP-7, D-11 |
| `PromoteFlow` and `PromoteModel` on the evidence; no drafted argument | STEP-8, D-12, D-13, D-15, AC-1.6, AC-1.7, AC-2.5 |
| The sheet reads the evidence in today's shape | STEP-9, D-14, AC-1.8 |
| Docs | STEP-10 |

| Out | Home |
| --- | --- |
| The sheet's new layout, copy built from relation and cause codes, "why suggested", expanding a collapsed line to uncheck one member, the relationship's own sentence | S11-20, S11-D2 |
| A depiction value on `ConclusionValue` for likeness lines | S11-20 (one `oneof` case, review M-12) |
| Renaming `ConclusionValue`, which now carries Observation values on two messages | not planned (it already did: `ReconcilerOutcomeDetail.recorded`) |
| Lifespan and likeness lines | S11-20 (user, 2026-10-21) |
| `Causes()` enumerators in `valuecompare`, `propertycompare`, `segmentcompare` | not planned (D-10: the test lists codes by constant, as the reasons row does) |
| Timing `Evidence` and Apply on a 100-Subject Source | S11-21 |
| Bounding the Source cache | S11-31 |
| Promote per graph | S12-09 |
| Pins on a row the researcher files New | not planned (a New handle has no members; today's rule) |
| Drafting an argument, in Go or Swift | not planned (user, 2026-10-21) |
| A Swift test of `GoStore`'s proto mapping | not planned (none exists; the Go handler test and the Swift flow tests cover each side) |

## 4. Current state

Verified at `10d3752`.

| Area | Where | Today |
| --- | --- | --- |
| Query | `core/database/promotealign/evidence.go` `Evidence`, `EvidenceRequest`, `evidenceReader`, `newEvidenceReader` | S11-19's query; no caller outside tests. `Evidence` calls `loadInputs`, then `checkFixed` on the row and each `req.Neighbors` entry (no anchors), then `newEvidenceReader` (a new profiler). An empty handle node gives `Evidence{Sources: {}}` |
| Evidence types | `core/claimevidence/claimevidence.go` `Pin{Subject, Observation, Source}`, `Line`, `Section`, `Bundle{Member, Neighbor, Source, Observations}`, `Relationship`, `Evidence{Sections, Relationships, Sources}`; `sections.go` `line` builds the one `Pin` from the speaking pair (`sp.Detail.RefB`, value `sp.B`); `build.go` builds each `Bundle` from a `Bridge{Subject, …}` | No member value on a pin, no bridge id on a bundle, no neighbor kind on a relationship, no row kind on the evidence |
| Setup | `core/database/promotealign/promotealign.go` `propose` (line 54 destructures `q, cfg, namePattern, terms, layer, primary := in.q, …`; line 74 calls `attachExhibits`), `inputs`, `loadInputs`, `mergeFixed`, `checkFixed` | `mergeFixed` checks each caller `Fixed` with `checkFixed`, adds `g.Source(id).Members` that the caller did not name as anchors, and sorts by Subject id |
| Layer memo | `core/database/promotealign/layer.go` `loadLayerMemo`, `loadLayer` | Memoized on the Source graph |
| Exhibits | `core/database/promotealign/exhibits.go` `exhibitObs`, `attachExhibits`, `withRowLines`, `refLine`, `rowLabeler`, `valueLabel`, `pairExhibits`, `cloneDate`, `acceptedMembers`, `acceptedClaims`, `bridgeNeighbors`, `loadExhibitObservations` | Fills `Row.Exhibits` on every proposal. `valueLabel` words a value (name form or parts, `datevalue.CompactDisplay`, term label or key, integer digits, text); `rowLabeler` tries a term's label on the lead, then its pool's `With` |
| Exhibit tests | `core/database/promotealign/exhibits_pin_test.go`; in `promotealign_test.go`, helpers `conflictExhibit` and `dateLine`, and `row.Exhibits` reads in `TestProposeSeedFromPromotedNeighbor`, `TestProposeOverlappingBoundShowsTheDateAndResembles`, `TestProposeOneHopDateExhibit`, `TestProposeOneHopExhibitsPairOnlyTheMappedNeighbor`, `TestShorterStructuredNameIsAWeakMatch`, `TestProposeExtraGivenNamesAreAWeakMatch`, `TestProposeJudgesAgainstTheMajority` (through `conflictExhibit`), `TestExhibitNameWithoutForm`, `TestProposeRowWithoutExhibitsShowsValues`, `TestProposeEventWithoutExhibitsShowsDates`, `TestProposeNegativeDateConflictLine` (through `dateLine`), `TestProposeKeepsTheLifespanOffTheSheet` | S11-13 MNT-5: `obs` copies production's `exhibitObs` construction |
| Typed values | `api/proto/engine.proto` `ConclusionValue` (oneof text, integer, `ConclusionTerm{id, key, label}`, `DateValueInput`, `NameValueInput`); `api/ffi/handlers/conclusion_detail.go` `conclusionValueProto(conclusiondetails.Value)`; `macos/App/Platform/GoStore.swift` `mapConclusionValue`; `GenealogyStore.swift` `CatalogConclusionValue`; `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `string(for:locale:orders:)` | The Conclusion page's one value-to-text path, in the researcher's locale and name order |
| Comparison text | `macos/App/Features/Promote/ComparisonValueDisplay.swift` | Called only from `GoStore`'s comparison mapping (lines 1009, 1016), which this PR deletes |
| Proposal types | `core/graphalign/types.go` `Exhibit`, `Row.Exhibits`, `Comparison.ValueType` (doc: "Pins come from Exhibit"); `core/graphalign/rows.go` `comparisonsFrom` sets `ValueType` (line 240) | `Comparison.ValueType` is read only by `refLine` |
| Pin write | `core/database/promote/promote.go` `Input.Pairs`, `Pair`, `pinTargets`, `handleOf`, `sqlOneHop`, `sqlNearHandle`, `bridgeRow`, `pairClaim`, `pinPairs`; `batch.go` `BatchRow.Pairs`, `SaveBatch` (builds `targets`), `applyRow` | `pinPairs` pins both Observations of a pair on the new claim and on the member's claim (`identityclaims.PinTx`, `ON CONFLICT DO NOTHING`). `api/ffi/handlers/promote.go:40` calls `Save` with no pairs; only tests set `Input.Pairs` |
| Pin tests | `core/database/promote/pins_test.go` `TestSavePins`; `batch_test.go` `TestSaveBatchOneHopPin`; `unpromote_test.go` line 214; `core/database/deleteimpact/conclusion_release_test.go` line 104; `core/database/deleteimpact/pins_e2e_test.go` line 96 | Pin through `promote.Pair`, four of them through `Save` |
| Proto | `api/proto/engine.proto` `Method` (last `METHOD_GET_EVIDENCE_GRAPH_UNLINK_IMPACT = 109`); `PromoteGraphAlignmentComparison`; `PromoteGraphAlignmentRow.comparisons = 9`; `ApplyPromoteGraphAlignmentPair`; `ApplyPromoteGraphAlignmentRow.pairs = 6`; `reserved 8; // reasons` precedent | — |
| Handlers | `api/ffi/handlers/promote_align.go` `ProposePromoteGraphAlignment` (inline `fixed` parse, lines 25–36), `proposalProto` (exhibit loop); `promote_batch.go` `ApplyPromoteGraphAlignment`; `api/ffi/dispatch.go` method constants and `Call` cases | — |
| Handler tests | `api/ffi/handlers/promote_align_test.go` `personOnNewSource(t, dir, userID, typeID, sexKey, polarity)`, `requireNegativeConflictLine`, `TestProposePromoteGraphAlignment`, `TestProposeRowWithoutExhibitsSendsValues`; `promote_test.go` `promoteOK`; `promote_batch_test.go` `TestApplyPromoteGraphAlignment`; `api/ffi/dispatch_test.go` | `requireNegativeConflictLine` reads comparisons |
| Vocabulary | `macos/App/Platform/ConclusionVocabulary.swift` (`PromoteReasonCode` with `init(wire:)` and `unrecognized`); `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` (rows; `wireCase` reads `case "([a-z_]*)": self = `; the "unrecognized" subtest lists enums) | No relation, outcome or cause enums |
| Skill | `.agents/skills/add-match-rule/SKILL.md` "Add a relation, cause or unit code" | Says codes "reach the wire and UI copy later"; no Swift step |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `CatalogPromoteGraphAlignmentComparison`, `CatalogPromoteGraphAlignmentRow.comparisons`, `CatalogPromoteGraphAlignmentPair`, `CatalogPromoteBatchRow.pairs`, protocol methods; `GoStore.swift` comparison mapping (lines 999–1024, `ComparisonValueDisplay.string` for dates) and pair mapping (1086–1090); `FakeStore.swift` `proposePromoteGraphAlignment`, `applyPromoteGraphAlignment` (checks `row.pairs`); `CoreInvoke.swift` `CoreMethod` | — |
| Flow | `macos/App/Features/Promote/PromoteFlow.swift` `Row.comparisons`, `Row.pins`, `Row.argumentEdited`, `setTarget`, `togglePin(subjectID:comparisonID:)`, `batchRows`, `batchRow`, `makeRow` (`kept.pins`, `draftedPins`, `draftedArgument`), `scope` | — |
| Model | `macos/App/Features/Promote/PromoteModel.swift` `openSheet`, `setSheetSubject`, `togglePin`, `repropose` (builds `fixed` from decided rows, line 260) | — |
| Sheet | `macos/App/Features/Promote/PromoteView.swift` `PromoteEvidenceSheet` (`row.comparisons.contains { $0.outcome == "conflict" }` line 673, `comparisonLine`, `groups`, `ExhibitGroup`, `incomingText`, `compared`, `outcomeTone(_: String)` line 859, `weightText`, `pin`) | — |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Promote.outcome(_ outcome: String)` (line 1374), `through`, `ownRecords`, `sheetNeighbor`, `pinLine`, `negatedValue`, `conflictCalloutTitle`, `conflictCalloutBody`, `argumentHint`, `graphUnavailable` | — |
| Swift tests | `macos/ProvenenciaTests/PromoteFlowTests.swift` (pins and argument tests, lines 500–630); `PromoteModelTests.swift` `row(_:handleID:assessment:)` (comparisons), `doneSendsThePinnedPair` | — |
| Kit | `PVSkeletonBar(width:height:)`, `PVCallout(tone:title:message:compact:)` (`.danger`), `PVCheckbox` | — |

**Precedent.** [#511](https://github.com/mendahu/provenencia/pull/511) (S12-05) added RPCs end to end: `f82d8e44` (proto and regenerated code), `15287b1e` (handlers, dispatch, handler tests), `8802e0e0` (`GenealogyStore`, `GoStore`, `FakeStore`, `CoreMethod`). Copy that split and its handler shape. [#504](https://github.com/mendahu/provenencia/pull/504) (S11-17) added `PromoteReasonCode` and its vocabulary row: copy that for D-10.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The RPC.** `METHOD_GET_PROMOTE_CLAIM_EVIDENCE = 110`, handler `GetPromoteClaimEvidence` in `api/ffi/handlers/promote_evidence.go` *(new)*. It parses neighbors with `fixedFromProto([]*engine.PromoteGraphAlignmentFixed) ([]graphalign.Fixed, error)` *(new, in `promote_align.go`, extracted from `ProposePromoteGraphAlignment`'s loop, which then calls it)*. Inside `withProjectCatalog`: `promotealign.Evidence`, `promotealign.PinOffers`, one `promotealign.NewEvidenceValues(c.Graph())`, then the mapping table in §6 FFI. `evidenceValueProto(promotealign.EvidenceValue) *engine.ConclusionValue` *(new, in `promote_evidence.go`)* sets the `oneof` case for `Type` as `conclusionValueProto` does (name `valuecodec.NameToProto`, date `valuecodec.DateToProto`, term `ConclusionTerm{Key, Label}` with an empty `id`, integer, text). Errors pass through | One lazy read per row; one parse of `Fixed` (review M-6) | Lines on the proposal — what this PR removes |
| D-2 | **Four fields on the pure types.** `Pin.Value propertycompare.Value` (`sp.B` in `sections.go` `line`); `Bundle.Bridge []byte` (`Bridge.Subject` in `build.go`); `Relationship.Kind vocabkey.SubjectType` (`Neighbor.Kind`); `Evidence.Kind vocabkey.SubjectType` (`Build`'s `kind`). Nothing else in `claimevidence` changes | The sheet shows each member's value; a bundle needs a stable id; the text needs each section's kind (D-4). Each is in hand where the type is built | A second read in the handler |
| D-3 | **Pin ids and offers, one owner.** `core/database/promotealign/pins.go` *(new)*: `LinePinID(incoming, observation []byte) string` = `"l:" + database.UUIDString(incoming) + ":" + database.UUIDString(observation)`; `BundlePinID(neighbor []byte, key vocabkey.BridgeKey, bridge []byte) string` = `"b:" + database.UUIDString(neighbor) + ":" + key.String() + ":" + database.UUIDString(bridge)`. `type PinOffer struct{ ID string; Checked bool; Set promote.PinSet }`, `func PinOffers(e claimevidence.Evidence) []PinOffer`: each pinnable line gives one offer per `Pin` (`LinePinID(line.Incoming.Ref, pin.Observation)`, `line.Checked`, `PinSet{Member: pin.Subject, Observations: {line.Incoming.Ref, pin.Observation}}`); each pinnable relationship gives one offer per `Bundle` (`BundlePinID(rel.Neighbor, rel.Key, bundle.Bridge)`, `rel.Checked`, `PinSet{Member: bundle.Member, Observations: rel.Incoming ++ bundle.Observations}`). Order: own sections' lines, then each relationship's bundles followed by its sections' lines. The RPC sends the offers (id, checked); Apply uses the same function | "The pin policy has one owner, in Go": the sheet's checked set and Apply's default set are one function, and Swift derives nothing (review M-3). Ids use the wire's id encoding (review M-8) | Ids or offer rules in Swift — a second owner. Observation pairs on the wire — cannot say "this bundle" |
| D-4 | **Typed values, worded in Swift** (review M-12). `core/database/promotealign/values.go` *(new)*: `type EvidenceValue struct{ Type string; Value propertycompare.Value; TermLabel string }`, `type EvidenceValues`, `func NewEvidenceValues(g *graphcache.Graph) (EvidenceValues, error)`, holding term labels by `property|term` and each property's value type, both read from `g.Vocabulary()` for origin `provenencia` only (a narrowing of `attachExhibits`' `termLabels`, which had no origin filter; lines compare product properties only). `Incoming(kind, s, l) EvidenceValue`: `Type` is `l.Property`'s value type, `Value` is `l.Incoming`, `TermLabel` the first label found on `l.Property`, `s.Key`, then `matchregistry.Default().Subject.Pools[kind][s.Key].With` in order (empty when none or not a term). `Member(kind, s, l, p) EvidenceValue`: the same type, `Value` is `p.Value`, labels tried on `s.Key`, the pool, then `l.Property`. `kind` is `Evidence.Kind` for own sections and `Relationship.Kind` under a relationship. `valueLabel`, `rowLabeler` and `cloneDate` are deleted with `exhibits.go`, not moved | One value-to-text path: the wire carries `ConclusionValue`, and Swift words it with `ReconciledValueDisplay` in the researcher's locale and name order, as the Conclusion page does. S11-20's likeness value is one `oneof` case | Go display strings beside per-type date fields — Swift rewrote every date, and each new value type would add a field to two messages |
| D-5 | **One pin write.** `promote.PinSet{Member []byte; Observations [][]byte}` *(new)*: Observations pinned on the new claim and backfilled onto `Member`'s accepted claim. `BatchRow.PinSets` replaces `Pairs`. `promote.Input.Pairs` is deleted, and `Save` writes no pins: `SaveBatch` is the one pin write (review M-4). `pinSets(tx, claim, sets)` *(renamed from `pinPairs`)*: `Member` and every Observation id must be 16 bytes and `Member` must have an accepted claim (`identityclaims.AcceptedEntityForSubjectTx`), else `ErrInvalid`; every Observation is pinned on the new claim and on that member claim with `PinTx`, counting new pins on the new claim. `pairClaim`, `pinTargets`, `handleOf`, `sqlOneHop`, `sqlNearHandle`, `bridgeRow` and `SaveBatch`'s `targets` map are deleted. `applyRow` reads `PinSets` where it read `Pairs` (Skip and New with pin sets → `ErrInvalid`; a Subject already on the same handle with none → nothing) | The evidence decided which Observations go together (D-6); the write only lands them. §9's rule is kept. One incoming Observation against many members is many sets sharing it, written once | Keeping `Pair` and adding a bundle type — two shapes for one write. Keeping the SQL checks — a weaker copy of the evidence's policy. The name `Pin` — `claimevidence.Pin` is one member Observation (review M-7) |
| D-6 | **One query body, one neighbor rule, one load per page.** `evidence.go`: `Evidence(c, req)` becomes `loadInputs` + `newEvidenceReader` + `evidence(g, in, r, req)` *(new, private)*, which holds today's body. `evidence` keeps today's checks on the request: more `req.Neighbors` than primary Subjects is `promote.ErrInvalid`, and a Subject named twice keeps its last target at its first position; the de-duplicated list then goes to `mergeFixed`. In `evidence`, the neighbors are `mergeFixed(g, req.SourceID, in.primary, req.Neighbors)` minus the row's own Subject, so an already filed Subject is a neighbor on its handle whether the request names it or not (AC-1.4); the request still wins when it names one. `pins.go`: `type ApplyRow struct{ promote.BatchRow; DefaultPins bool; PinIDs []string }` *(new)*, `func ResolvePins(c *database.Catalog, sourceID []byte, rows []ApplyRow) ([]promote.BatchRow, error)` *(new)*: one `loadInputs` and one `newEvidenceReader` for the page; for each handle row whose Subject is not in `g.Source(sourceID).Members`, one `evidence` call with the other rows as `graphalign.Fixed` (handle rows with `EntityID`, New and Skip by target); `DefaultPins` → every checked offer's set, else every offer whose id is in `PinIDs` (ids not offered are dropped). A New or Skip row with `PinIDs`, or any row with `DefaultPins` and `PinIDs`, is `promote.ErrInvalid`. Rows keep their order; sets keep offer order, one per id | Working rule "The pin policy has one owner, in Go": Go applies default pins at Apply for rows whose sheet was never opened, and the RPC drafts them for rows that were. `mergeFixed` already owns "the page's rows plus filed Subjects" for `Propose` (review M-1). One load per page makes "builds the layer once" true by construction (review M-2). Dropping a vanished id is the user's call (2026-10-21) | A neighbor rule in `ResolvePins` and another in Swift. A test hook in `loadLayer`. Refusing Done on a vanished id (user chose drop) |
| D-7 | **Apply handler.** `ApplyPromoteGraphAlignment` parses each row into `promotealign.ApplyRow` (`pin_ids`, `default_pins`), calls `promotealign.ResolvePins` inside `withProjectCatalog` before `writes.Run`, and passes the resolved rows to `promote.SaveBatch`. A write between the two is refused by `SaveBatch`'s `SeenRevision` check (`promote.stale`) | The graph read happens outside the write transaction, as `Propose` does | Resolving in `promote` — needs the graph cache there, a cycle |
| D-8 | **Proto changes.** `PromoteGraphAlignmentRow`: `reserved 9; // comparisons: GetPromoteClaimEvidence`. `PromoteGraphAlignmentComparison` and `ApplyPromoteGraphAlignmentPair` deleted. `ApplyPromoteGraphAlignmentRow`: `reserved 6; // pairs: pin_ids`, `repeated string pin_ids = 7;`, `bool default_pins = 8;`. Regenerate with `./scripts/generate-proto.sh` | No adapter ("High tolerance for churn"); `reserved` as the file already does | An empty field 9 |
| D-9 | **Go deletions.** `exhibits.go` goes; `acceptedClaims` moves to `evidence.go` unchanged. `propose` drops the `attachExhibits` call and `q` from its destructure at line 54; `inputs.q` is deleted; `loadInputs` keeps its local `Querier` for `valuerules.Load`. `graphalign.Exhibit`, `Row.Exhibits`, `Comparison.ValueType` and its write in `comparisonsFrom` go; the `Comparison` doc loses "Pins come from Exhibit". `exhibits_pin_test.go` goes (S11-13 MNT-5, PRF-2). In `promotealign_test.go`: delete `TestProposeOneHopDateExhibit`, `TestProposeOneHopExhibitsPairOnlyTheMappedNeighbor`, `TestExhibitNameWithoutForm`, `TestProposeRowWithoutExhibitsShowsValues`, `TestProposeEventWithoutExhibitsShowsDates` and the helpers `conflictExhibit`, `dateLine`; in `TestProposeSeedFromPromotedNeighbor`, `TestProposeOverlappingBoundShowsTheDateAndResembles`, `TestShorterStructuredNameIsAWeakMatch`, `TestProposeExtraGivenNamesAreAWeakMatch`, `TestProposeJudgesAgainstTheMajority`, `TestProposeNegativeDateConflictLine` and `TestProposeKeepsTheLifespanOffTheSheet`, remove only the `row.Exhibits`, `conflictExhibit` and `dateLine` assertions, and delete a test left with no assertion. `ComparisonValueDisplay.swift` is deleted in STEP-8 (its one caller goes). `requireNegativeConflictLine` and `TestProposeRowWithoutExhibitsSendsValues` go; the negative case is T-7's | "Build, then migrate"; S11-19's tests cover lines | Porting exhibit tests to `Evidence` — S11-19's T-3 to T-16 cover them |
| D-10 | **Wire vocabulary.** `ConclusionVocabulary.swift` gains `EvidenceOutcome` (`propertycompare.Outcome*`), `EvidencePropertyCause` (`propertycompare.Cause*`), `ValueRelation` (`valuecompare.Relation`), `ValueCause` (`valuecompare.Cause*`) and `SegmentCause` (`segmentcompare.Cause*`). Each has `init(wire:)` with a `switch`, one `case "<code>": self = .<case>` line per Go code, and `default: self = .unrecognized(wire)`. The three cause enums also have a `none` case, listed first, matched by `case "": self = .none` (an agreeing segment and a line with no cause send `""`). `TestSwiftVocabularyMatchesGo` gets one row per enum: relations from `valuecompare.Relations()`, every other enum by its constants, as the reasons row does, and `""` in the `want` list of the three cause rows; the "unrecognized" subtest lists the five. `L10n.Promote.outcome` and `PromoteView`'s `outcomeTone` take `EvidenceOutcome`. `add-match-rule`'s "Add a relation, cause or unit code" gains the Swift case and the test row (review M-5) | Working rule "One wire vocabulary"; the map moved cause codes here from S11-17. `wireCase` matches `""`, so `""` is in `want` | Strings in views. `Causes()` functions in three packages — hand-listing is the reasons row's precedent (Out) |
| D-11 | **Swift store.** In `GenealogyStore.swift` *(all new)*: `CatalogClaimEvidence{sections, relationships, offers: [CatalogClaimEvidenceOffer], sourceTitles: [String: String]}` with `checkedPinIDs` and `offeredPinIDs` (sets read straight from `offers`), `CatalogClaimEvidenceOffer{id, checked}`, `CatalogClaimEvidenceSection{propertyKey, lines}`, `CatalogClaimEvidenceLine{propertyKey, outcome: EvidenceOutcome, propertyCause: EvidencePropertyCause, relation: ValueRelation, cause: ValueCause, points, pinnable, checked, incomingObservationID, incoming: CatalogConclusionValue, incomingNegative, members}`, `CatalogClaimEvidencePin{id, memberSubjectID, observationID, sourceID, value: CatalogConclusionValue}`, `CatalogClaimEvidenceRelationship{neighborSubjectID, key, bridgeType, term, end, eventType, outcome, cause: SegmentCause, points, pinnable, checked, bundles, sections}`, `CatalogClaimEvidenceBundle{id, memberSubjectID, sourceID}`. Protocol `getPromoteClaimEvidence(projectDir:sourceID:subjectID:handleID:neighbors:) async throws -> CatalogClaimEvidence`. `GoStore` maps `incoming` and each pin's `value` with the existing `mapConclusionValue`. In Swift the Go field `Relationship.Members` is `bundles` and `Line.Members` is `members`, as on the wire. `CatalogPromoteBatchRow.pairs` becomes `defaultPins: Bool` and `pinIDs: [String]`; `CatalogPromoteGraphAlignmentComparison`, `CatalogPromoteGraphAlignmentPair` and `CatalogPromoteGraphAlignmentRow.comparisons` go. `FakeStore`: `var promoteEvidence: [String: CatalogClaimEvidence]` by Subject id (missing → empty evidence), `var promoteEvidenceError: Error?`, and a recorded call `getPromoteClaimEvidence subject=<id> handle=<id> neighbors=<count>`; its apply checks read `pinIDs.isEmpty` | Mirrors the wire; Swift copies offers and derives nothing | Reusing the comparison struct — it holds one member |
| D-12 | **Flow.** `PromoteFlow.Row` loses `comparisons` and `argumentEdited` and gains `evidence: CatalogClaimEvidence?` (nil: never opened). `fixed(for rows: [Row]) -> [CatalogPromoteGraphAlignmentFixed]` *(new)*: handle → `(subjectID, handleID)`, New → target `new`, Skip → target `skip`, unset → none; `PromoteModel.repropose` calls it with the decided rows instead of its own loop. `evidenceNeighbors(for subjectID:)` *(new)* = `fixed(for:)` over the rows `batchRows()` files (`scope`, not anchor, not unset) other than that row; Go adds filed Subjects (D-6). `setEvidence(subjectID:handleID:evidence:)` *(new)*: ignored unless the row's target is still `.handle(handleID)`; when `evidence` was nil, `pins = evidence.checkedPinIDs`, else `pins = pins ∩ evidence.offeredPinIDs`; then stores it. `togglePin(subjectID:pinIDs:)` replaces `togglePin(subjectID:comparisonID:)`: all ids in `pins` → remove all, else insert all; marks the row decided. `setTarget` to a new target sets `evidence = nil` and `pins = []`. `makeRow` keeps `previous.evidence` and `previous.pins` when the previous row is decided, or when its target equals the new drafted target; otherwise both start empty. `setTarget`, `makeRow` and `merge` never change `argument`: it is what the researcher typed, kept across a retarget and a re-proposal. `batchRow`: a handle row sends `defaultPins: evidence == nil`, `pinIDs: evidence == nil ? [] : pins.sorted()`; New and Skip send `false`, `[]`. `draftedPins`, `draftedArgument` and `kept.pins` go | The page's rows are the query's inputs, through the same mapping `repropose` uses; Go owns the filed-Subject rule (review M-1). Opening alone does not mark a row decided, so an untouched opened row equals its defaults | Swift sending anchors — a second copy of `mergeFixed` |
| D-13 | **Model.** `enum SheetEvidence: Equatable { case idle, loading, failed(String) }` *(new)* and `private(set) var sheetEvidence = SheetEvidence.idle` on `PromoteModel`; the loaded value lives in `Row.evidence` (review M-10). Every `openSheet` and `setSheetSubject` call (closing included) bumps the generation and sets `sheetEvidence = .idle` first; then it sets the subject and, for a row whose target is `.handle`, calls `loadEvidence(subjectID:handleID:)` *(new)*: `sheetEvidence = .loading`, `store.getPromoteClaimEvidence` with `flow.evidenceNeighbors(for:)`, then `flow.setEvidence` and `.idle`, or `.failed(L10n.Errors.message(for:))`. A generation counter (as `proposeGeneration`) drops an answer for a sheet that is no longer the open one. Every open refetches. `togglePin(subjectID:pinIDs:)` forwards | Lazy per row; refetching on open is the user's answer for reopening (2026-10-21) | Two flags that allow "loading with an error" |
| D-14 | **The sheet in today's shape** (user, 2026-10-21). The "This subject" card lists every line of `evidence.sections` in section order. Each relationship is a card titled `L10n.Promote.through(neighbor name)` (the flow row's name, else `sheetNeighbor`), with the neighbor row's `ref`; its header has a `PVCheckbox` bound to its bundles' ids, disabled when not pinnable or with no bundle; its sections' lines follow. A line is one row: the checkbox is bound to its pins' ids and disabled when not pinnable; Compared is `model.propertyLabel(key: line.propertyKey, origin: "provenencia")`; Here is `ReconciledValueDisplay.string(for: line.incoming, orders: model.session.nameFormat)` (`negatedValue` when negative); the target column lists each pin's value, worded the same way, with its Source title (`sourceTitles[pin.sourceID]`) under it; Result is `outcome(line.outcome)` with `outcomeTone(line.outcome)`; Weight is `weightText(line.points)`. The conflict callout shows when a line in any section has outcome `.conflict`. With `sheetEvidence == .loading` and no `evidence`: three `PVSkeletonBar(width:height: 12)` in the lines area. With `.failed(message)`: `PVCallout(tone: .danger, title: L10n.Promote.evidenceUnavailable, message: message, compact: true)`. Spoken labels: a line's is `pinLine(property:here:there:)` with `there` the pins' worded values joined by ", "; a relationship's is `pinLine(property: through(name), here: "", there: "")`. Identifiers: `promote.sheet.pin.<first pin id>`, `promote.sheet.relationship.<key>` | No new chrome; S11-20 redraws it | Hiding relationships — Done would pin bundles the researcher cannot see |
| D-15 | **No drafted argument** (user, 2026-10-21). A row's argument starts empty and is only what the researcher types | User decision | Drafting in Go at Apply — declined |

**Strings.** One new key: `L10n.Promote.evidenceUnavailable` = "The evidence could not be loaded" (`promote.sheet.evidenceUnavailable`, comment "Evidence sheet callout title when the evidence read fails").

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `GetPromoteClaimEvidence` handler, `METHOD_GET_PROMOTE_CLAIM_EVIDENCE`, its messages *(new)* | FFI | `ProposePromoteGraphAlignment` | A lazy per-row read; the proposal stops carrying lines |
| `fixedFromProto` *(new)* | func | `ProposePromoteGraphAlignment`'s loop | Extracted; both handlers call it |
| `promotealign.evidence` *(new)* | func | `Evidence`'s body | Moved, so Apply loads once |
| `promotealign.PinOffers`, `PinOffer`, `LinePinID`, `BundlePinID` *(new)* | funcs, type | `PromoteFlow.draftedPins`; `pairExhibits`' `Pinned` | Moves the default-pin rule into Go beside the query |
| `promotealign.ApplyRow`, `ResolvePins` *(new)* | type, func | `promote.pairClaim` | Replaces the SQL checks with the evidence |
| `promotealign.EvidenceValue`, `EvidenceValues`, `NewEvidenceValues` *(new)*; handler `evidenceValueProto` *(new)* | types, funcs | `conclusionValueProto`; `exhibits.go` `rowLabeler` | Builds a `ConclusionValue` from a `propertycompare.Value`, which `conclusionValueProto`'s `conclusiondetails.Value` is not; keeps `rowLabeler`'s pool rule for term labels |
| `promote.PinSet` *(new)*, `pinSets` *(renamed)* | type, func | `promote.Pair`, `pinPairs` | Replaced, not beside |
| `claimevidence` `Pin.Value`, `Bundle.Bridge`, `Relationship.Kind`, `Evidence.Kind` | fields | — | Data `Build` already holds |
| `EvidenceOutcome`, `EvidencePropertyCause`, `ValueRelation`, `ValueCause`, `SegmentCause` *(new)* | Swift enums | `PromoteReasonCode` | Same pattern, file and test |
| `CatalogClaimEvidence*`, `getPromoteClaimEvidence` *(new)* | Swift types, store method | `CatalogPromoteGraphAlignmentComparison` | That holds one member; deleted |
| `PromoteFlow.fixed(for:)`, `evidenceNeighbors`, `setEvidence`, `Row.evidence` *(new)* | funcs, field | `PromoteModel.repropose`'s loop | The loop moves into `fixed(for:)` |
| `PromoteModel.SheetEvidence`, `sheetEvidence`, `loadEvidence` *(new)* | enum, state, func | `repropose`, `proposeError` | Same generation pattern, one state value |
| `L10n.Promote.evidenceUnavailable` *(new)* | key | `graphUnavailable` | Names the sheet's failure |
| Tests: `promotealign/pins_test.go`, `promotealign/values_test.go`, `handlers/promote_evidence_test.go`, fixture `namedMembers` *(new)* | tests | `memo_test.go`, `promote_align_test.go` `personOnNewSource` | Reuse those fixtures |

**Builds on.** `ConclusionValue`, `mapConclusionValue`, `ReconciledValueDisplay`; S11-19's `Evidence`, `loadInputs`, `mergeFixed`, the profiler memo (D-22) and `newWarmFixture`; `identityclaims.PinTx`; `withProjectCatalog`; the S12-05 RPC split; `PromoteReasonCode`'s vocabulary pattern; `PVCheckbox`, `PVSkeletonBar`, `PVCallout`.
**Deletes.** `exhibits.go` (all but `acceptedClaims`), `exhibits_pin_test.go`, `graphalign.Exhibit`, `Row.Exhibits`, `Comparison.ValueType`, `inputs.q`; `promote.Pair`, `Input.Pairs`, `pairClaim`, `pinTargets`, `handleOf`, `sqlOneHop`, `sqlNearHandle`, `bridgeRow`; proto `PromoteGraphAlignmentComparison`, `ApplyPromoteGraphAlignmentPair`, fields 9 and 6; `proposalProto`'s exhibit loop; `requireNegativeConflictLine`, `TestProposeRowWithoutExhibitsSendsValues`; `valueLabel`, `rowLabeler`, `cloneDate`; Swift `ComparisonValueDisplay.swift`, `CatalogPromoteGraphAlignmentComparison`, `CatalogPromoteGraphAlignmentPair`, `Row.comparisons`, `Row.argumentEdited`, `draftedPins`, `draftedArgument`, `ExhibitGroup`, `repropose`'s fixed loop. With them goes every `autoreconcile.Compatible` call in Promote.
**Next adds.** S11-20 words lines from the D-10 enums and redraws the sheet on `CatalogClaimEvidence` with no wire change; unchecking one member of a collapsed line is a per-pin toggle, as ids are per pin. S11-20's likeness value is a `DepictionValue` case on `ConclusionValue`; its lifespan (not pinnable) and likeness lines get offers from `PinOffers` with no change to Apply or Swift. S12-09 re-keys the filed-Subject rule in `mergeFixed` only.

### Schema

None — `identity_claim_evidence` already holds one incoming Observation against many member Observations.

### Go core

| Package | Change |
| --- | --- |
| `core/claimevidence` | D-2's four fields, set in `sections.go` `line` and `build.go` `Build` |
| `core/graphalign` | Delete `Exhibit`, `Row.Exhibits`, `Comparison.ValueType` (D-9) |
| `core/database/promote` | `PinSet` *(new)*; `BatchRow.PinSets`; `Input.Pairs` deleted; `pinSets`; deletions (D-5) |
| `core/database/promotealign` | `evidence.go`: `evidence` *(new)*, neighbors through `mergeFixed`, `acceptedClaims` moved in (D-6); `pins.go` *(new)* (D-3, D-6); `values.go` *(new)* (D-4); `promotealign.go`: no `attachExhibits`, no `inputs.q`; `exhibits.go` deleted |

```go
// package promote
// PinSet is evidence the researcher confirmed: Observations pinned on the new
// claim and backfilled onto Member's accepted claim.
type PinSet struct {
	Member       []byte
	Observations [][]byte
}

// package promotealign
func LinePinID(incoming, observation []byte) string
func BundlePinID(neighbor []byte, key vocabkey.BridgeKey, bridge []byte) string

type PinOffer struct {
	ID      string
	Checked bool
	Set     promote.PinSet
}

func PinOffers(e claimevidence.Evidence) []PinOffer

type ApplyRow struct {
	promote.BatchRow
	DefaultPins bool     // the sheet was never opened: pin what the evidence checks
	PinIDs      []string // the pins the researcher left checked
}

func ResolvePins(c *database.Catalog, sourceID []byte, rows []ApplyRow) ([]promote.BatchRow, error)

type EvidenceValue struct {
	Type      string // propertymeta.ValueType*
	Value     propertycompare.Value
	TermLabel string
}

func NewEvidenceValues(g *graphcache.Graph) (EvidenceValues, error)
func (v EvidenceValues) Incoming(kind vocabkey.SubjectType, s claimevidence.Section, l claimevidence.Line) EvidenceValue
func (v EvidenceValues) Member(kind vocabkey.SubjectType, s claimevidence.Section, l claimevidence.Line, p claimevidence.Pin) EvidenceValue
```

Errors: `promote.invalid` (D-5, D-6) as today.

### FFI

`api/proto/engine.proto`, regenerated with `./scripts/generate-proto.sh`:

```proto
// GetPromoteClaimEvidence is one Promote row's claim evidence against its
// target, from the targets the page holds for its rows (S11-25).
message GetPromoteClaimEvidenceRequest {
  string project_dir = 1;
  string source_id = 2;
  string subject_id = 3;
  string handle_id = 4;
  repeated PromoteGraphAlignmentFixed neighbors = 5;
}

message PromoteClaimEvidencePin {
  string id = 1;
  string member_subject_id = 2;
  string member_observation_id = 3;
  string source_id = 4;
  ConclusionValue value = 5;
}

message PromoteClaimEvidenceLine {
  string property_key = 1;
  string outcome = 2;        // propertycompare.Outcome
  string property_cause = 3; // propertycompare.Cause
  string relation = 4;       // valuecompare.Relation
  string cause = 5;          // valuecompare.Cause
  double points = 6;
  bool pinnable = 7;
  bool checked = 8;
  string incoming_observation_id = 9;
  ConclusionValue incoming = 10;
  bool incoming_negative = 11;
  repeated PromoteClaimEvidencePin members = 12;
}

message PromoteClaimEvidenceSection {
  string property_key = 1;
  repeated PromoteClaimEvidenceLine lines = 2;
}

message PromoteClaimEvidenceBundle {
  string id = 1;
  string member_subject_id = 2;
  string source_id = 3;
}

message PromoteClaimEvidenceRelationship {
  string neighbor_subject_id = 1;
  string key = 2; // BridgeKey.String(): identity only
  string bridge_type = 3;
  string term = 4;
  string end = 5;
  string event_type = 6;
  string outcome = 7;
  string cause = 8; // segmentcompare.Cause
  double points = 9;
  bool pinnable = 10;
  bool checked = 11;
  repeated PromoteClaimEvidenceBundle bundles = 12;
  repeated PromoteClaimEvidenceSection sections = 13;
}

// A pin the evidence offers; checked ones are the defaults.
message PromoteClaimEvidenceOffer {
  string id = 1;
  bool checked = 2;
}

message PromoteClaimEvidenceSource {
  string id = 1;
  string title = 2;
}

message GetPromoteClaimEvidenceResponse {
  repeated PromoteClaimEvidenceSection sections = 1;
  repeated PromoteClaimEvidenceRelationship relationships = 2;
  repeated PromoteClaimEvidenceOffer offers = 3;
  repeated PromoteClaimEvidenceSource sources = 4;
}
```

Handler mapping (`kind` is `e.Kind` for top-level sections and `rel.Kind` inside a relationship; every id is `database.UUIDString`):

| Field | Source |
| --- | --- |
| section `property_key` | `Section.Key` |
| line `property_key`, `outcome`, `property_cause`, `relation`, `cause`, `points`, `pinnable`, `checked` | `Line.Property`, `Outcome`, `PropertyCause`, `Relation`, `Cause`, `Points`, `Pinnable`, `Checked` |
| line `incoming_observation_id`, `incoming_negative` | `Line.Incoming.Ref`, `Line.Incoming.Negative` |
| line `incoming` | `evidenceValueProto(values.Incoming(kind, section, line))` |
| pin `id` | `LinePinID(line.Incoming.Ref, pin.Observation)` |
| pin `member_subject_id`, `member_observation_id`, `source_id` | `Pin.Subject`, `Pin.Observation`, `Pin.Source` |
| pin `value` | `evidenceValueProto(values.Member(kind, section, line, pin))` |
| relationship `neighbor_subject_id`, `key` | `Relationship.Neighbor`, `Relationship.Key.String()` |
| relationship `bridge_type`, `term`, `end`, `event_type` | `Key.Type`, `Key.Term`, `Key.End`, `Key.EventType` |
| relationship `outcome`, `cause`, `points`, `pinnable`, `checked` | `Outcome`, `Cause`, `Points`, `Pinnable`, `Checked` |
| `bundles` (Go `Relationship.Members`): `id`, `member_subject_id`, `source_id` | `BundlePinID(rel.Neighbor, rel.Key, bundle.Bridge)`, `Bundle.Member`, `Bundle.Source` |
| `offers` | `PinOffers(e)`: `ID`, `Checked`, in order |
| `sources` | `Evidence.Sources` (keyed by raw id bytes): `id` = `database.UUIDString([]byte(key))`, `title`, sorted by `id` |

Dispatch: `MethodGetPromoteClaimEvidence` constant and `Call` case in `api/ffi/dispatch.go`; a row in `dispatch_test.go`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Line checkbox | `PVCheckbox` | `PromoteView.swift` `PromoteEvidenceSheet.evidenceLine` *(renamed from `comparisonLine`)* | Bound to the line's pin ids (D-14) |
| Relationship card header checkbox | `PVCheckbox` | `PromoteEvidenceSheet` relationship card | Bound to the bundle ids |
| Loading | `PVSkeletonBar` ×3 | lines area | `sheetEvidence == .loading` and no evidence |
| Error | `PVCallout(tone: .danger, …, compact: true)` | above the lines | `.failed(message)` |
| Member column | a `VStack` of worded value + italic muted Source title per pin | existing column | Today's one-member block, repeated |

States: a row whose target is not a handle → no lines and no fetch, as today; loading; error; a line with 50 members → the column grows and the sheet scrolls (today's `ScrollView`).

### Strings

`L10n.Promote.evidenceUnavailable` *(new)*. Changed signature: `L10n.Promote.outcome(_ outcome: EvidenceOutcome)`; `.unknown` and `.unrecognized` read "Unknown". Removed: none.

## 7. Steps

**STEP-1 — The four fields**
Files: `core/claimevidence/claimevidence.go`, `sections.go`, `build.go`, `sections_test.go`, `build_test.go`. Skill: `add-match-rule`.
Do: D-2. Add the new fields to existing expectations; no other expectation changes.
Check: `go test ./core/claimevidence/... ./core/matchregistry/...` passes (T-1).

**STEP-2 — Typed evidence values**
Files: `core/database/promotealign/values.go` *(new)*, `values_test.go` *(new)*.
Do: D-4.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestEvidenceValues ./core/database/promotealign/` passes (T-6).

**STEP-3 — One query body, one pin write, Apply's resolver**
Files: `core/database/promotealign/evidence.go`, `evidence_test.go`, `pins.go` *(new)*, `pins_test.go` *(new)*, `promotealign_test.go` (`namedMembers`), `core/database/promote/promote.go`, `batch.go`, `pins_test.go`, `batch_test.go`, `unpromote_test.go`, `core/database/deleteimpact/conclusion_release_test.go`, `core/database/deleteimpact/pins_e2e_test.go`. Skill: `use-catalog-session`.
Do: D-6's split and `mergeFixed` neighbors, then D-5, then D-3 and D-6's `ResolvePins`. A test that pinned through `Save` with `Pairs` pins through `SaveBatch` with one `BatchRow{Target: TargetHandle, PinSets: …}` (as `TestSaveBatchOneHopPin` builds a batch); `promote.Pair{IncomingObservationID: a, MemberObservationID: b}` becomes `PinSet{Member: <b's Subject>, Observations: {a, b}}`. A `pins_test.go` case that tested a `pairClaim` refusal (outsider, other property, provisional, one-hop) is deleted; its rule is now T-3's.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promote/... ./core/database/promotealign/... ./core/database/deleteimpact/...` passes; S11-19's `TestEvidence*` pass unchanged; T-2, T-3, T-4, T-5 pass.

**STEP-4 — The wire**
Files: `api/proto/engine.proto`, regenerated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/dispatch.go`, `dispatch_test.go`, `api/ffi/handlers/promote_evidence.go` *(new)*, `promote_evidence_test.go` *(new)*, `promote_batch.go`, `promote_batch_test.go`, `promote_align.go`, `promote_align_test.go`. Skill: `add-ffi-handler`.
Do: D-1, D-7, D-8 and the §6 mapping. Delete `proposalProto`'s exhibit loop, `requireNegativeConflictLine` and `TestProposeRowWithoutExhibitsSendsValues`; drop the comparison assertions in `TestProposePromoteGraphAlignment`. The Go side passes alone; Swift builds again after STEP-8 (commit line).
Check: a second `./scripts/generate-proto.sh` run leaves no diff; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; T-7, T-8 pass.

**STEP-5 — Delete the exhibits**
Files: `core/database/promotealign/exhibits.go` (deleted), `exhibits_pin_test.go` (deleted), `evidence.go`, `promotealign.go`, `promotealign_test.go`, `core/graphalign/types.go`, `rows.go`.
Do: D-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; the first AC-4.2 `rg` prints no Go file.

**STEP-6 — Wire vocabulary**
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `api/ffi/swift_vocabulary_test.go`.
Do: D-10's enums and test rows.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestSwiftVocabularyMatchesGo ./api/ffi/` passes (T-9).

**STEP-7 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`. Skill: `add-swift-test`.
Do: D-11.
Check: `rg -n 'pairs|CatalogPromoteGraphAlignmentPair' macos/App/Platform --glob '!Generated/**'` prints nothing. The app builds after STEP-8, whose check covers both.

**STEP-8 — Flow and model**
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `PromoteModel.swift`, `ComparisonValueDisplay.swift` (deleted), `macos/App/Platform/L10n.swift` (`outcome`), `macos/ProvenenciaTests/PromoteFlowTests.swift`, `PromoteModelTests.swift`. Skill: `add-swift-test`.
Do: D-12, D-13, D-15, D-10's `L10n.Promote.outcome`. Rewrite the tests listed in T-10 and T-11; delete the comparison fixtures in both test files' `row(...)` helpers. Keep `PromoteView.swift` compiling with the smallest edits (the sheet is STEP-9).
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes; T-10, T-11 pass.

**STEP-9 — The sheet**
Files: `macos/App/Features/Promote/PromoteView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-14 and the new key.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test` passes; both AC-4.2 `rg` commands print nothing; §10.

**STEP-10 — Docs**
Files: `.agents/skills/add-match-rule/SKILL.md` ("Add a relation, cause or unit code": a code that reaches the wire gets a case in its `ConclusionVocabulary.swift` enum and stays in its `TestSwiftVocabularyMatchesGo` row), `docs/promote-graph-alignment.md` (§3: the sheet loads its evidence when opened, and the argument is the researcher's own; §9: Go pins what the evidence offers, the checked lines and relationships for a sheet never opened, and a pin's Observations land on the new claim and on the claim of each member it names; §12: delete the "Exhibit values" bullet and "and exhibits" from the query-count bullet), `docs/conclusion-layer-data-model.md` (the `argument` sentence after the Done list: the researcher writes it; Promote drafts none), `docs/deployment-plan/spike-11/claim-evidence.md` (Status: RPC and pins built in S11-25; Retires: done in S11-25), `docs/deployment-plan/spike-11/comparison-layers.md` (reuse table layer 2 row: "Promote's claim evidence (`core/claimevidence`)"; duplicates table `pairExhibits` row: "— done in S11-25"), `docs/deployment-plan/spike-11/deployment-plan.md` (map row status), `docs/deployment-plan/spike-11/completed.md`, this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-25-claim-evidence-rpc.md` reports no errors.

Commits: one per step, except STEP-4 to STEP-8, which land as one commit: `generate-proto.sh` writes the Go and Swift code together, and Swift builds only once STEP-8 is done. Inside that commit, run each step's Go check as you go. Say so in the PR. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Load a Promote row's evidence when its sheet opens."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/claimevidence/sections_test.go` `TestSections`, `build_test.go` `TestBuild` | Existing rows with D-2's fields expected: each pin's `Value` is the member value; each bundle's `Bridge` its bridge Subject; `Relationship.Kind` event; `Evidence.Kind` the row kind | D-2 |
| T-2 | Go | `core/database/promote/pins_test.go` `TestSavePins` (rewritten on `SaveBatch`) | One `PinSet` with the incoming and member name Observations → both on the new claim and on the member's claim, `Pins` 2; the same set twice → written once; `Member` with no accepted claim → `ErrInvalid`; a 4-byte id → `ErrInvalid`; a New row with a set → `ErrInvalid` | AC-3.5, AC-3.6, D-5 |
| T-3 | Go | `core/database/promotealign/pins_test.go` `TestResolvePins` *(new)*, table-driven; fixture `(f *fixture) namedMembers(n int, name string) (handle []byte, members []subjects.Subject)` *(new, `promotealign_test.go`)*: `n` Sources titled "Census 1851", "Census 1861", …, each with a person `name`, the first `promote`d, the rest `join`ed to its handle | `namedMembers(5, "John Smith")` and a new Source person "John Smith": default row → five sets, one per member, each `{member, [in, mem]}`; opened row with one id → that set only; an id not offered → dropped; a conflict line's id → dropped; New row with ids → `ErrInvalid`; default with ids → `ErrInvalid`; a row already a member of the Source graph → no sets | AC-1.1, AC-2.1, AC-3.1, AC-3.2, AC-3.3, D-3, D-6 |
| T-4 | Go | same file `TestResolvePinsBundlesARelationship` *(new)* | `newWarmFixture`, person row on the person handle, birth row on the birth handle, both default → the person row's sets hold the bundle set `{filed person, Source bridge Observations + Source birth event_type + filed bridge Observations + filed birth event_type}` and the date line set; `PinOffers(Evidence(request))` gives the same ids, checked. Subtest AC-1.4: the birth promoted first, the request naming no neighbor → the relationship is there | AC-1.2, AC-1.3, AC-1.4, AC-2.2, AC-2.3, D-3, D-6 |
| T-5 | Go | same file `TestResolvePinsReadsThePageOnce` *(new)* | Three handle rows; `Propose` the page, then `SetQueryCounterForTest`; `ResolvePins` → no graph query, and the sets equal three single-row `Evidence` calls' default offers | AC-3.4, D-6 |
| T-6 | Go | `core/database/promotealign/values_test.go` `TestEvidenceValues` *(new)*, table-driven | A `name` line → `Type` name; a `gender` term `man` in section `sex_at_birth` → `Type` term, `TermLabel` its catalog label; a term with no label → `TermLabel` empty; a member value under a pooled lead → the lead's label; a `date` → `Type` date; a `start_date` line in section `date` → `Type` date | D-4 |
| T-7 | Go | `api/ffi/handlers/promote_evidence_test.go` `TestGetPromoteClaimEvidence` *(new)*, `runRPC` | Fixture: three `personOnNewSource(…, "male", positive)` Subjects, the first promoted with `promoteOK`, the rest with `promoteOK` and `EntityId` set to its handle; an incoming `personOnNewSource(…, "male", positive)`. → one `sex_at_birth` line, three pins with ids and Source ids, three checked offers, three sources. The incoming negative (`PolarityNegative`) → a conflict line, not pinnable, no offers. A bad source id, a Subject not on the Source, an empty handle → error | AC-1.1, AC-1.5, AC-2.4, D-1 |
| T-8 | Go | `api/ffi/handlers/promote_batch_test.go` `TestApplyPromoteGraphAlignment` (extended) | T-7's fixture, row `default_pins` → each member's claim gains the incoming and its own `sex_at_birth` Observation; the same with `pin_ids` from T-7's offers → the same pins; `default_pins` and `pin_ids` together → error; the negative incoming with `default_pins` → no pins | AC-2.1, AC-2.3, AC-2.4, D-7 |
| T-9 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` | Five new rows (with `""` in the three cause rows) and the "unrecognized" subtest | AC-4.1, D-10 |
| T-10 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` | `aNeverOpenedRowSendsDefaultPins` *(new)*: handle row, no evidence → `defaultPins` true, `pinIDs` empty. `openingDraftsTheCheckedPins` *(new)*: `setEvidence` → `pins == checkedPinIDs`, batch sends them sorted, `defaultPins` false. `reopeningKeepsOnlyOfferedPins` *(new)*: uncheck a line, `setEvidence` with offers lacking one id → pins lose it and keep the unchecked line unchecked. `aLineTogglesEveryMember` *(new)*: `togglePin(pinIDs: [a, b])` twice → both in, then both out. `evidenceForAnOldTargetIsIgnored` *(new)*. `evidenceNeighborsAreTheRowsDoneFiles` *(new)*: a handle row, a New row, a Skip row, an unset row, an anchor, a hidden row → the handle, New and Skip rows only. `aRetargetClearsEvidenceAndPins` (rewrites `aRetargetRedraftsPinsAndArgumentForTheNewTarget`). `everyRowStartsWithAnEmptyArgument` *(new)*. `aTypedArgumentSurvivesARetargetAndAReproposal` *(new)*. `aManualPinOnTheSameTargetSurvivesAReproposal` (rewritten on `setEvidence` and `merge`). `aSkipRowFilesNoArgumentConfidenceOrPins` reads `pinIDs`. Deleted: `aConflictingComparisonIsNotDraftedAsAPin`, `aNegatedConflictLineIsNotDraftedAsAPin` (Go owns drafting: T-3, T-7) | AC-1.6, AC-1.7, AC-2.5, D-12, D-15 |
| T-11 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` | `openingASheetLoadsItsEvidence` *(new)*: `FakeStore.promoteEvidence` set → the call is recorded with the handle and the neighbor count, the row's pins are the checked ids, `sheetEvidence == .idle`. `aFailedEvidenceReadShowsAnError` *(new)*: `promoteEvidenceError` → `sheetEvidence == .failed(…)`, the row keeps no evidence; closing the sheet → `.idle`. `aLateAnswerForAnotherSheetIsDropped` *(new)*: open A, open B before A's answer → A's evidence is not stored. `doneSendsDefaultOrChosenPins` (rewrites `doneSendsThePinnedPair`): one row never opened, one opened → `lastPromoteBatch` rows carry `defaultPins` true / false with the ids | AC-1.6, AC-1.8, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `promotealign_test.go` (D-9), S11-19's `evidence_test.go` (D-6's `mergeFixed`), `core/database/promote` and `deleteimpact` pin tests (D-5), `api/ffi/handlers` Promote tests (D-8), `PromoteFlowTests` (D-12).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without Xcode, report the Go and plan-doc steps and say Xcode was not run
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings: promote.sheet.evidenceUnavailable; no stale strings left
- [ ] UI uses kit components only (PVCheckbox, PVSkeletonBar, PVCallout); no new kit primitive
- [ ] Both AC-4.2 rg commands print nothing
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — none: not gated
- [ ] Docs: add-match-rule skill; promote-graph-alignment.md §3, §9, §12; conclusion-layer-data-model.md argument sentence; claim-evidence.md; comparison-layers.md two rows
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a filed person "John Smith" (two Sources, each with John Smith and a birth in 1817, both promoted onto one person and one birth). Add a third Source with "John Smith", a birth in 1817 and a subject participation. Open Promote on it. Open the person row's sheet: skeleton bars, then a "This subject" card with a checked name line listing two members with their Source titles, and a "Through <birth name>" card with a checked header and a checked date line.
2. Close the sheet without changes. Press Done. On the filed person's conclusion page, the new claim's pins include the name, the bundle and the date Observations, and each earlier member's claim gained the incoming name Observation.
3. On a fresh copy, uncheck the date line before Done: the date Observations are not pinned.
4. On a fresh copy, set the birth row to Skip and reopen the person sheet: no "Through" card; Done pins only the name.
5. With `FakeStore.promoteEvidenceError` set in a preview or test, the sheet shows the danger callout and the claim fields still work.
6. Run both AC-4.2 `rg` commands from the repo root: each prints nothing.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Done now runs the query once per handle row | One load per page (D-6) and the S11-19 memos; T-5; S11-21 times it |
| A neighbor change after opening drops pins quietly | User decision (2026-10-21); reopening shows what is left |
| A catalog write between Apply's read and its write | `SaveBatch`'s `SeenRevision` refuses the batch (`promote.stale`); the page re-proposes |

**Stop and report if:**

- An S11-19 `TestEvidence*` test changes value after D-6's `mergeFixed` change, other than a filed neighbor now counting.
- A filed person member's claim cannot take a bridge Subject's Observation as a pin (a `PinTx` or foreign key error in T-4).
- `identity_claim_evidence`'s primary key is not `(identity_claim_id, observation_id)` at your base.
- A `promotealign_test.go` assertion on something other than exhibits changes value after STEP-5.
- `claimevidence` needs an import under `core/database` or of `core/matchregistry` in a non-test file.
- A Go cause constant sends a value `wireCase` cannot read (anything outside `[a-z_]*`).
- `./scripts/generate-proto.sh` cannot run on your host (no `protoc`, `protoc-gen-go` or `protoc-gen-swift`): regenerated code is never written by hand.

## 12. Open questions

None. Four product calls were answered by the user on 2026-10-21: relationships keep their card with a header checkbox (D-14); a collapsed line is one row listing every member (D-14); no argument is drafted (D-15); pins the evidence no longer offers are dropped (D-6).

## 13. PR

**Title:** S11-25 — Claim evidence RPC and pin policy
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-25-claim-evidence-rpc.md`.
