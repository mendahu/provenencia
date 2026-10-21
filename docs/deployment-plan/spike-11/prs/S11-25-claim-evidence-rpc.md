# S11-25 — Claim evidence RPC and pin policy

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 5 — Claim evidence
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the shipped sheet reads the new lines in today's columns and cards; S11-D2 gates the new layout in S11-20)
**Board:** —
**Depends on:** S11-19 (landed [#513](https://github.com/mendahu/provenencia/pull/513), with its §14 amendment)
**Blocks:** S11-20
**Branch:** `s11-25-claim-evidence-rpc`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (`core/claimevidence` is a guarded layer package), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md) (one changed helper signature, no new key), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (kit components only)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/catalog-session.mdc`
**Context:** [`claim-evidence.md`](../claim-evidence.md) (Lines, Writing pins, Evidence through relationships, Retires); [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3 and §9; [S11-19 plan](S11-19-claim-evidence-query.md) (D-3, D-12, D-13, §14) and its [ship log](S11-19-claim-evidence-query.ship.md) ("Apply memo check"); working rules "The pin policy has one owner, in Go", "One wire vocabulary", "Build, then migrate", "High tolerance for churn", "A separate, lazy query per row"; S11-13 ship log MNT-5, PRF-2
**Planned:** 2026-10-21 against `main` @ `10d3752`
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

The Promote sheet reads a row's claim evidence from S11-19's query, through a new lazy RPC, when the researcher opens it. Done pins what the evidence offers: for a row whose sheet was never opened, Go pins the lines and relationships that start checked; for an opened row, it pins the ones the researcher left checked. Each pin lands on the new claim and backfills every member's claim it names, so joining an entity with five named members pins all five. The proposal stops carrying sheet lines, and the old exhibit code (`attachExhibits`, `pairExhibits`, `withRowLines`, `refLine`, proposal field 9) is deleted.

**Why now.** S11-19 built the query with no caller ("Build, then migrate"). This PR is the migrate half. S11-20 rebuilds the sheet on the RPC this PR adds.

**What stays.** The sheet's layout: the columns (Pin, Compared, Here, the target, Result, Weight), the "This subject" card and one "Through <name>" card per neighbor, the conflict callout, and the claim fields. All existing strings. The proposal's rows, bands, reasons, alternatives, contested and move flags. Bridge filing at Done. The pin tables (`identity_claim_evidence`, primary key `(identity_claim_id, observation_id)`). The rule that a pin's Observations land on the new claim and on the member's claim ([`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §9).

## 2. User stories

Conventions. The Go fixture is `promotealign_test.go`'s (`newFixture`, `newSource`, `person`, `datedBirth`, `participation`, `filedBirth`, `join`, `promote`) and `memo_test.go`'s `newWarmFixture`. "Checked ids" is `claimEvidenceOffers(e)` filtered to `Checked` (D-3). Line pin id `l:<in>:<mem>` and bundle pin id `b:<neighbor>:<key>:<bridge>` use lowercase hex of the 16-byte ids (D-3).

**US-1.** As the researcher, when I open a row's sheet, I see every member's evidence for that row's target, with the lines that match already checked.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A handle with five accepted members, each `name` "John Smith" on Sources "Census 1851" … "Census 1891"; a Source person "John Smith" on that handle; `GetPromoteClaimEvidence` for the row | One section `name` with one line: relation `same`, outcome `agree`, checked, pinnable, five pins in member order, each with its member Subject, Observation, Source id, display "John Smith" and pin id `l:<in>:<mem>`; `sources` lists the five titles |
| AC-1.2 | `newWarmFixture`: Source person and birth 1817 with the birth row on the filed birth's handle; request neighbors hold the birth | One relationship: neighbor the Source birth, `bridge_type` `participation`, `term` `subject`, outcome `agree`, checked, one bundle with id `b:<birth>:<key>:<bridge>`, and one section `date` under it with one checked line |
| AC-1.3 | AC-1.2 with the birth row's target New, Skip, or left out of `neighbors` | No relationship |
| AC-1.4 | A request with an unknown Source, a Subject not on it, an empty handle, or a merged handle | Error code `promote.invalid` |
| AC-1.5 | Opening a sheet for the first time (Swift) | The row's pins are the evidence's checked pin ids |
| AC-1.6 | Opening it again after the researcher unchecked a line, then set a neighbor to Skip | The row keeps its unchecked state; pins the new evidence no longer offers are dropped |
| AC-1.7 | The sheet while the RPC is in flight; after it fails | Skeleton bars in the lines area; then a danger callout with the error message, and the claim fields still editable |

**US-2.** As the researcher, when I press Done without opening a row's sheet, the row is pinned exactly as the sheet would have drafted it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.1's page; Done with the row never opened | The new claim pins the incoming name Observation and all five member name Observations; each of the five members' accepted claims gains both the incoming Observation and its own |
| AC-2.2 | AC-1.2's page; Done with the person row never opened | The person's new claim pins the Source participation bridge's Observations, the Source birth's `event_type` Observation, the filed bridge's Observations and the filed birth's `event_type` Observation (the bundle), and the two date Observations; the filed person member's claim gains every Observation of the bundle pin (both sides); the filed birth member's claim gains both date Observations |
| AC-2.3 | AC-1.1's page; open the sheet, change nothing, Done | The same pins as AC-2.1 |
| AC-2.4 | A row on a handle whose member's `name` conflicts ("John Smith" against "Peter Brown") | Done pins nothing for that line, opened or not |
| AC-2.5 | The draft argument | Every row starts with an empty argument; Done files what the researcher typed, or nothing |

**US-3.** As the engine, Apply pins only what the evidence offers, from the page's final targets.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ResolvePins` with an opened row whose ids include one id the evidence no longer offers, and one non-pinnable id (a conflict line's) | Both are dropped; the rest pin; no error |
| AC-3.2 | A New or Skip row that sends pin ids, or a row with `DefaultPins` and pin ids | `promote.ErrInvalid` |
| AC-3.3 | A row whose Subject is already an accepted member of its target | `ResolvePins` gives it no pins; `SaveBatch` writes nothing for it, as today |
| AC-3.4 | `ResolvePins` over a page with three handle rows on a cold catalog | `loadLayer` runs once; the match index keeps the handle profiles |
| AC-3.5 | `promote.SaveBatch` with a `Pin` whose `Member` has no accepted claim, or an id that is not 16 bytes | `promote.ErrInvalid`, nothing written |
| AC-3.6 | Two rows' pins name one Observation on one claim | It is pinned once |

**US-4.** As a contributor, the evidence's codes cross the wire once and Swift parses them in one place.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `TestSwiftVocabularyMatchesGo` | Swift's `EvidenceOutcome`, `EvidencePropertyCause`, `ValueRelation`, `ValueCause`, `SegmentCause` list exactly Go's `propertycompare.Outcome*`, `propertycompare.Cause*`, `valuecompare.Relations()`, `valuecompare.Cause*`, `segmentcompare.Cause*` values, each with an `unrecognized` case |
| AC-4.2 | `rg -n "attachExhibits\|pairExhibits\|withRowLines\|refLine\|bridgeNeighbors\|loadExhibitObservations\|exhibitObs\|graphalign.Exhibit\|pairClaim\|draftedPins\|draftedArgument\|CatalogPromoteGraphAlignmentComparison\|CatalogPromoteGraphAlignmentPair\|ApplyPromoteGraphAlignmentPair\|PromoteGraphAlignmentComparison" core api macos --glob '!**/Generated/**' --glob '!**/engine.pb.go'` | Prints nothing |

## 3. Scope

| In | Covered by |
| --- | --- |
| `claimevidence`: `Pin.Value`, `Bundle.Bridge`, `Relationship.Kind`, `Evidence.Kind` | STEP-1, D-2 |
| Pin ids and offers, `ResolvePins`, the sheet's value text (`promotealign`) | STEP-2, STEP-3, D-3, D-4, D-6, AC-3.1–AC-3.4 |
| `promote.Pin` replaces `promote.Pair`; `pairClaim`, `pinTargets` and the one-hop SQL deleted | STEP-3, D-5, AC-3.5, AC-3.6 |
| RPC `GetPromoteClaimEvidence` *(new)*; Apply row takes `pin_ids` and `default_pins`; proposal field 9 and the comparison and pair messages removed | STEP-4, D-1, D-7, D-8, AC-1.1–AC-1.4, AC-2.1–AC-2.4 |
| Go deletions: `exhibits.go` (but `acceptedClaims`), `exhibits_pin_test.go`, `graphalign.Exhibit`, `Row.Exhibits`, `Comparison.ValueType`, the exhibit assertions in `promotealign_test.go` | STEP-5, D-9, AC-4.2 |
| Wire vocabulary in `ConclusionVocabulary.swift`, under `swift_vocabulary_test.go` | STEP-6, D-10, AC-4.1 |
| Swift store: types, `getPromoteClaimEvidence`, `GoStore`, `FakeStore`; the batch row shape | STEP-7, D-11 |
| `PromoteFlow` and `PromoteModel` on the evidence; no drafted argument | STEP-8, D-12, D-13, D-15, AC-1.5, AC-1.6, AC-2.5 |
| The sheet reads the evidence in today's shape | STEP-9, D-14, AC-1.7 |
| Docs | STEP-10 |

| Out | Home |
| --- | --- |
| The sheet's new layout, copy built from relation and cause codes, "why suggested", expanding a collapsed line to uncheck one member, the relationship's own sentence | S11-20, S11-D2 |
| Lifespan and likeness lines | S11-20 (user, 2026-10-21) |
| Timing `Evidence` and Apply on a 100-Subject Source | S11-21 |
| Bounding the Source cache | S11-31 |
| Promote per graph | S12-09 |
| Pins on a row the researcher files New | not planned (a New handle has no members to pin against; today's rule) |
| Drafting an argument, in Go or Swift | not planned (user, 2026-10-21) |
| A Swift test of `GoStore`'s proto mapping | not planned (no `GoStore` mapping test exists; the Go handler test and Swift flow tests cover each side) |

## 4. Current state

Verified at `10d3752`.

| Area | Where | Today |
| --- | --- | --- |
| Query | `core/database/promotealign/evidence.go` `Evidence`, `EvidenceRequest`, `evidenceReader`, `acceptedClaims` (called from `exhibits.go`) | S11-19's query. No caller outside tests. Returns `claimevidence.Evidence`; an empty handle node gives `Evidence{Sources: {}}` |
| Evidence types | `core/claimevidence/claimevidence.go` `Pin{Subject, Observation, Source}`, `Line`, `Section`, `Bundle{Member, Neighbor, Source, Observations}`, `Relationship`, `Evidence{Sections, Relationships, Sources}`; `sections.go` `line` builds the one `Pin` from the speaking pair (`sp.Detail.RefB`); `build.go` builds `Bundle` from `Bridge{Subject, …}` | A pin has no member value, a bundle no bridge id, a relationship no neighbor kind, the evidence no row kind |
| Setup | `core/database/promotealign/promotealign.go` `propose`, `loadInputs`, `mergeFixed`, `checkFixed` | `propose` calls `attachExhibits` after `Align` (line 74). `mergeFixed` adds `g.Source(id).Members` as anchors |
| Layer memo | `core/database/promotealign/layer.go` `loadLayerMemo`, `loadLayer` | Memoized on the Source graph under `layerMemoKey + namePattern` |
| Exhibits | `core/database/promotealign/exhibits.go` `exhibitObs`, `attachExhibits`, `withRowLines`, `refLine`, `rowLabeler`, `valueLabel`, `pairExhibits`, `cloneDate`, `acceptedMembers`, `acceptedClaims`, `bridgeNeighbors`, `loadExhibitObservations` | Fills `Row.Exhibits` on every proposal: one pair per incoming value (first `Compatible`, else closest), a second SQL read of Observations, and bridges of any type (`bridgeNeighbors`). `valueLabel` and `rowLabeler` word a value (a term by its catalog label, a lead's pool tried in order) |
| Exhibit tests | `core/database/promotealign/exhibits_pin_test.go` (`obs`, `dateObs`, `befDate`, three tests); `promotealign_test.go` tests that read `row.Exhibits` (lines 323–1240: `TestProposeOneHopDateExhibit`, `TestProposeOneHopExhibitsPairOnlyTheMappedNeighbor`, `TestExhibitNameWithoutForm`, `TestProposeRowWithoutExhibitsShowsValues`, `TestProposeEventWithoutExhibitsShowsDates`, `TestProposeNegativeDateConflictLine`, and assertions inside others) | S11-13 MNT-5: `obs` copies production's `exhibitObs` construction |
| Proposal types | `core/graphalign/types.go` `Exhibit`, `Row.Exhibits`, `Comparison.ValueType` (doc: "Pins come from Exhibit"); `core/graphalign/rows.go` `comparisonsFrom` sets `ValueType` (line 240) | `Comparison.ValueType` is read only by `refLine` |
| Pin write | `core/database/promote/promote.go` `Input.Pairs`, `Pair`, `pinTargets`, `handleOf`, `sqlOneHop`, `sqlNearHandle`, `bridgeRow`, `pairClaim`, `pinPairs`; `batch.go` `BatchRow.Pairs`, `SaveBatch` builds `pinTargets`, `applyRow` | `pinPairs` pins both Observations of a pair on the new claim and on the member's claim (`identityclaims.PinTx`, `ON CONFLICT DO NOTHING`); `pairClaim` checks same Property and the one-hop shape by SQL. `Save`'s `Input.Pairs` has no wire caller (`PromoteSubjectRequest` has no pairs) |
| Pin tests | `core/database/promote/pins_test.go` `TestSavePins`; `batch_test.go` `TestSaveBatchOneHopPin`; `unpromote_test.go` line 214; `core/database/deleteimpact/conclusion_release_test.go` line 104; `core/database/deleteimpact/pins_e2e_test.go` line 96 | Build pins with `promote.Pair` |
| Proto | `api/proto/engine.proto` `Method` (last `METHOD_GET_EVIDENCE_GRAPH_UNLINK_IMPACT = 109`); `PromoteGraphAlignmentComparison`; `PromoteGraphAlignmentRow.comparisons = 9`; `ApplyPromoteGraphAlignmentPair`; `ApplyPromoteGraphAlignmentRow.pairs = 6` | Lines ride on every proposal |
| Handlers | `api/ffi/handlers/promote_align.go` `ProposePromoteGraphAlignment`, `proposalProto` (maps `Exhibits`); `promote_batch.go` `ApplyPromoteGraphAlignment` (parses pairs, `writes.Run` → `promote.SaveBatch`); `api/ffi/dispatch.go` method constants and `Call` cases | — |
| Handler tests | `api/ffi/handlers/promote_align_test.go` `TestProposePromoteGraphAlignment` (asserts comparisons, line 91), `TestProposeRowWithoutExhibitsSendsValues`; `promote_batch_test.go` `TestApplyPromoteGraphAlignment`; `api/ffi/dispatch_test.go` | — |
| Vocabulary | `macos/App/Platform/ConclusionVocabulary.swift` (`ReconciledState`, `ReconcilerReason`, `PromoteReasonCode`, each with `init(wire:)` and `unrecognized`); `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` (table rows, `wireCase` regex, the "unrecognized" subtest listing enums) | No relation, outcome or cause enums |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `CatalogPromoteGraphAlignmentComparison`, `CatalogPromoteGraphAlignmentRow.comparisons`, `CatalogPromoteGraphAlignmentPair`, `CatalogPromoteBatchRow.pairs`, protocol `proposePromoteGraphAlignment`, `applyPromoteGraphAlignment`; `GoStore.swift` maps comparisons (lines 999–1024, dates through `ComparisonValueDisplay.string`) and pairs (1086–1090); `FakeStore.swift` `proposePromoteGraphAlignment`, `applyPromoteGraphAlignment` (checks `row.pairs`); `CoreInvoke.swift` `CoreMethod` | — |
| Flow | `macos/App/Features/Promote/PromoteFlow.swift` `Row.comparisons`, `Row.pins`, `Row.argumentEdited`, `setTarget` (clears comparisons, pins, argument), `togglePin(subjectID:comparisonID:)`, `batchRow` (pairs from pinned comparisons), `makeRow` (`kept.pins`, `draftedPins`, `draftedArgument`), `resolve(_:)` | Pins and the argument are drafted from proposal field 9 |
| Model | `macos/App/Features/Promote/PromoteModel.swift` `openSheet`, `closeSheet`, `setSheetSubject`, `togglePin` | Opening a sheet only sets `sheetSubjectID` |
| Sheet | `macos/App/Features/Promote/PromoteView.swift` `PromoteEvidenceSheet` (`comparisonLine`, `groups`, `ExhibitGroup`, `incomingText`, `compared`, `outcomeTone`, `weightText`, `pin`) | Reads `row.comparisons`; one card per `groupSubjectID` |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Promote.outcome(_ outcome: String)` (line 1374), `through`, `ownRecords`, `sheetNeighbor`, `pinLine`, `negatedValue`, `conflictCalloutTitle`, `conflictCalloutBody`, `argumentHint` | `outcome` switches on strings |
| Swift tests | `macos/ProvenenciaTests/PromoteFlowTests.swift` (pins and argument tests, lines 500–630); `PromoteModelTests.swift` `row(_:handleID:assessment:)` helper with comparisons, `doneSendsThePinnedPair` | — |
| Kit | `macos/App/DesignSystem/Components/SkeletonBar` `PVSkeletonBar(width:height:)`; `Callout` `PVCallout(tone:title:message:compact:)`, tone `.danger`; `Checkbox` `PVCheckbox` | — |

**Precedent.** [#511](https://github.com/mendahu/provenencia/pull/511) (S12-05) added RPCs end to end: `f82d8e44` (proto methods and messages, regenerated), `15287b1e` (handlers, dispatch, handler tests), `8802e0e0` (`GenealogyStore`, `GoStore`, `FakeStore`, `CoreMethod`). Copy that split and the handler shape (`proto.Unmarshal`, `parseID`, `withProjectCatalog`, `proto.Marshal`). [#504](https://github.com/mendahu/provenencia/pull/504) (S11-17) added `PromoteReasonCode` and its `TestSwiftVocabularyMatchesGo` row: copy that for D-10.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The RPC.** `METHOD_GET_PROMOTE_CLAIM_EVIDENCE = 110`, handler `GetPromoteClaimEvidence` in `api/ffi/handlers/promote_evidence.go` *(new)*. It parses the request into `promotealign.EvidenceRequest` (neighbors as `graphalign.Fixed`, as `ProposePromoteGraphAlignment` parses `fixed`), calls `promotealign.Evidence` inside `withProjectCatalog`, and maps the result with `claimEvidenceOffers` ids (D-3) and `EvidenceText` (D-4). Messages in §6 FFI. Errors pass through (`promote.invalid`) | One lazy read per row ([`claim-evidence.md`](../claim-evidence.md#separate-from-matching)); the handler only maps | Lines on the proposal — what this PR removes |
| D-2 | **Four fields on the pure types.** `Pin.Value propertycompare.Value` (the member's value, `sp.B` in `sections.go` `line`); `Bundle.Bridge []byte` (the filed bridge Subject, from `Bridge.Subject` in `build.go`); `Relationship.Kind vocabkey.SubjectType` (the neighbor's, `Neighbor.Kind`); `Evidence.Kind vocabkey.SubjectType` (`Build`'s `kind`). Nothing else in `claimevidence` changes | The sheet shows each member's value; a bundle needs a stable id; the value text needs each section's Subject kind (D-4). All four are already in hand where the type is built | A second read in the handler — a copy of what `Build` held |
| D-3 | **Pin ids and offers, one owner.** `core/database/promotealign/pins.go` *(new)*: `type offer struct{ id string; checked bool; pin promote.Pin }` and `func claimEvidenceOffers(e claimevidence.Evidence) []offer`. Each pinnable line gives one offer per `Pin`: id `"l:" + hex(line.Incoming.Ref) + ":" + hex(pin.Observation)`, checked = `line.Checked`, `promote.Pin{Member: pin.Subject, Observations: {line.Incoming.Ref, pin.Observation}}`. Each pinnable relationship gives one offer per `Bundle`: id `"b:" + hex(rel.Neighbor) + ":" + rel.Key.String() + ":" + hex(bundle.Bridge)`, checked = `rel.Checked`, `promote.Pin{Member: bundle.Member, Observations: rel.Incoming ++ bundle.Observations}`. Own sections first, then relationships in order, each followed by its sections. Exported for the handler as `func PinOffers(e claimevidence.Evidence) []PinOffer` with `PinOffer{ID string; Checked bool}` | "The pin policy has one owner, in Go": the RPC's checked set and Apply's default set are one function, so "opening the sheet shows the same defaults" holds by construction | Ids built in Swift — a second owner. Sending Observation pairs back — cannot say "this bundle" |
| D-4 | **Value text.** `core/database/promotealign/sheettext.go` *(new)*: `type EvidenceText` and `func NewEvidenceText(g *graphcache.Graph) (EvidenceText, error)`, holding term labels by `property|term` and each product property's value type, from `g.Vocabulary()`. `Incoming(kind, s, l) string` words `l.Incoming` with value type of `l.Property` and term labels tried on `l.Property`, `s.Key`, then `matchregistry.Default().Subject.Pools[kind][s.Key].With` in order. `Member(kind, s, l, p) string` words `p.Value` with the same value type and the order `s.Key`, the pool, `l.Property`. Wording is `valueLabel`'s, moved unchanged. `kind` is `Evidence.Kind` for own sections and `Relationship.Kind` under a relationship | Moves `valueLabel` and `rowLabeler`'s pool rule out of the deleted `exhibits.go`; values carry no `Type` (`subjectprofile.Value`) | Display strings in `claimevidence` — copy is not layer policy. In the handler — needs the registry's pools and the vocabulary in `api` |
| D-5 | **One pin shape.** `promote.Pin{Member []byte; Observations [][]byte}` *(new)* replaces `promote.Pair`; `Input.Pins` and `BatchRow.Pins` replace `Pairs`. `pinPins(tx, claim, pins)` *(renamed from `pinPairs`)*: for each pin, `Member` and every Observation id must be 16 bytes, and `Member` must have an accepted claim (`identityclaims.AcceptedEntityForSubjectTx`), else `ErrInvalid`; every Observation is pinned on the new claim and on that member claim (`PinTx`), counting new pins on the new claim. `pairClaim`, `pinTargets`, `handleOf`, `sqlOneHop`, `sqlNearHandle`, `bridgeRow` and `SaveBatch`'s `targets` map are deleted. `applyRow`'s "same handle and no pairs → nothing" and "New with pairs → invalid" rules read `Pins` | The evidence decided which Observations go together (D-6); the write only lands them. §9's rule (both claims get the pin's Observations) is kept. One incoming Observation against many members is many `Pin`s sharing an id, written once | Keeping `Pair` and adding a bundle type — two shapes for one write. Keeping the SQL checks — a second, weaker copy of the evidence's policy |
| D-6 | **Apply resolves pins in Go.** `core/database/promotealign/pins.go`: `type ApplyRow struct{ promote.BatchRow; DefaultPins bool; PinIDs []string }` *(new)*, `func ResolvePins(c *database.Catalog, sourceID []byte, rows []ApplyRow) ([]promote.BatchRow, error)` *(new)*. Neighbors for every row: each other row as `graphalign.Fixed` (handle rows with their `EntityID`, New and Skip by target), then every `g.Source(sourceID).Members` Subject no row names, on its `EntityID`. For each row with target `handle`: a row whose Subject is in `Members` gets no pins; otherwise one `Evidence` call, then `DefaultPins` → every checked offer, else every offer whose id is in `PinIDs` (an id not offered is dropped). A New or Skip row with `PinIDs`, or any row with `DefaultPins` and `PinIDs`, is `promote.ErrInvalid`. Rows keep their order; `Pins` keep offer order with duplicates removed by id | Working rule "The pin policy has one owner, in Go": Go applies default pins at Apply for rows whose sheet was never opened, and the evidence RPC drafts them for rows that were. Neighbors match what Swift sends when the sheet opens (D-12): the page's rows plus already filed Subjects. Dropping a vanished id is the user's call (2026-10-21) | Validating in the write transaction — needs the graph cache inside `promote`, a cycle (`promotealign` imports `promote`). Refusing Done on a vanished id — user chose drop |
| D-7 | **Apply handler.** `ApplyPromoteGraphAlignment` parses each row into `promotealign.ApplyRow` (`pin_ids` as strings, `default_pins`), calls `promotealign.ResolvePins` inside `withProjectCatalog` before `writes.Run`, and passes the resolved rows to `promote.SaveBatch` unchanged otherwise. A catalog write between the two is caught by `SaveBatch`'s `SeenRevision` check (`promote.stale`) | The graph read happens before the write, as `Propose` does; the revision stamp already guards the gap | Resolving inside `writes.Run` — the graph cache is read outside write transactions everywhere else |
| D-8 | **Proto changes.** `PromoteGraphAlignmentRow`: `reserved 9; // comparisons: the claim evidence RPC`. `PromoteGraphAlignmentComparison` and `ApplyPromoteGraphAlignmentPair` deleted. `ApplyPromoteGraphAlignmentRow`: `reserved 6; // pairs: pin_ids`, `repeated string pin_ids = 7;`, `bool default_pins = 8;`. Regenerate with `./scripts/generate-proto.sh` | "High tolerance for churn": no adapter. `reserved` follows the file's own precedent (`reserved 8; // reasons`) | Keeping field 9 empty — a dead field |
| D-9 | **Go deletions.** `exhibits.go` goes; `acceptedClaims` moves to `evidence.go` unchanged. `propose` no longer calls `attachExhibits`. `inputs.q` is deleted (its only reader was `attachExhibits`); `loadInputs` keeps its local `Querier` for `valuerules.Load`. `graphalign.Exhibit`, `Row.Exhibits` and `Comparison.ValueType` (and its write in `comparisonsFrom`) go; the `Comparison` doc loses "Pins come from Exhibit". `exhibits_pin_test.go` goes (S11-13 MNT-5, PRF-2). In `promotealign_test.go`, a test whose subject is exhibit lines is deleted, and an assertion on `row.Exhibits` inside another test is removed; the display cases they held (a name with no form reads its parts, a date reads `CompactDisplay`, a term reads its label, a negative value) move to T-6 | "Build, then migrate": the query replaced them in S11-19 and its tests cover lines | Porting exhibit tests to `Evidence` — S11-19's T-3 to T-16 already cover them |
| D-10 | **Wire vocabulary.** `ConclusionVocabulary.swift` gains `EvidenceOutcome` (`propertycompare.Outcome`), `EvidencePropertyCause` (`propertycompare.Cause`), `ValueRelation` (`valuecompare.Relation`), `ValueCause` (`valuecompare.Cause`) and `SegmentCause` (`segmentcompare.Cause`), each `init(wire:)` with an `unrecognized(String)` case; `""` parses to `.none` in the three cause enums. `TestSwiftVocabularyMatchesGo` gets one row per enum (relations from `valuecompare.Relations()`, the rest listed by constant as the reasons row does) and the "unrecognized" subtest lists the five. `L10n.Promote.outcome` takes `EvidenceOutcome` | Working rule "One wire vocabulary"; the map moved cause codes here from S11-17 | Strings in views — the file's rule "views and models use the typed cases" |
| D-11 | **Swift store.** In `GenealogyStore.swift`: `CatalogClaimEvidence{sections, relationships, sourceTitles: [String: String]}`, `CatalogClaimEvidenceSection{propertyKey, lines}`, `CatalogClaimEvidenceLine{propertyKey, outcome, propertyCause, relation, cause, points, pinnable, checked, incomingObservationID, incomingDisplay, incomingNegative, members}`, `CatalogClaimEvidencePin{id, memberSubjectID, observationID, sourceID, display}`, `CatalogClaimEvidenceRelationship{neighborSubjectID, key, bridgeType, term, end, eventType, outcome, cause, points, pinnable, checked, bundles, sections}`, `CatalogClaimEvidenceBundle{id, memberSubjectID, sourceID}` *(all new)*. `CatalogClaimEvidence` has `checkedPinIDs: Set<String>` and `offeredPinIDs: Set<String>` (pinnable lines' and relationships' ids). Protocol `getPromoteClaimEvidence(projectDir:sourceID:subjectID:handleID:neighbors:) async throws -> CatalogClaimEvidence` *(new)*. `CatalogPromoteBatchRow.pairs` becomes `defaultPins: Bool` and `pinIDs: [String]`; `CatalogPromoteGraphAlignmentComparison`, `CatalogPromoteGraphAlignmentPair` and `CatalogPromoteGraphAlignmentRow.comparisons` go. `GoStore` formats dates with `ComparisonValueDisplay.string` as today. `FakeStore` gains `var promoteEvidence: [String: CatalogClaimEvidence]` by Subject id, `var promoteEvidenceError: Error?`, and records `getPromoteClaimEvidence subject=<id> handle=<id> neighbors=<n>` | Mirrors the wire one to one; the two id sets are what the flow needs (D-12) | Reusing the comparison struct — it holds one member |
| D-12 | **Flow.** `PromoteFlow.Row` loses `comparisons` and `argumentEdited`, and gains `evidence: CatalogClaimEvidence?` (nil: never opened). `setEvidence(subjectID:handleID:evidence:)` *(new)*: ignored unless the row's target is still `.handle(handleID)`; when `evidence` was nil, `pins = evidence.checkedPinIDs`, else `pins = pins ∩ evidence.offeredPinIDs`; then stores it. `togglePin(subjectID:pinIDs:)` replaces `togglePin(subjectID:comparisonID:)`: all ids in `pins` → remove all, else insert all; marks the row decided. `setTarget` to a new target sets `evidence = nil`, `pins = []`. `makeRow` keeps `previous.evidence` and `previous.pins` when the previous row is decided, or when its target equals the new drafted target; otherwise both start empty. `evidenceNeighbors(for:) -> [CatalogPromoteGraphAlignmentFixed]` *(new)*: every other row whose `resolve` gives a handle id that does not start with `new:`, as `CatalogPromoteGraphAlignmentFixed(subjectID:handleID:)`. `batchRow`: a handle row sends `defaultPins: evidence == nil`, `pinIDs: evidence == nil ? [] : pins.sorted()`; New and Skip send `false`, `[]`. `draftedPins`, `draftedArgument` and `kept.pins` go | The page's targets are the query's inputs; Go's neighbor rule (D-6) reads the same rows and anchors. Opening alone does not mark a row decided, so an untouched opened row equals its defaults | Sending every row as a neighbor — New rows and hidden rows would differ from Apply |
| D-13 | **Model.** `openSheet` and `setSheetSubject` set the subject, then, for a row whose target is `.handle`, start `loadEvidence(subjectID:handleID:)` *(new)*: `sheetLoading = true`, `sheetError = nil`, call `store.getPromoteClaimEvidence` with `flow.evidenceNeighbors(for:)`, then `flow.setEvidence` or `sheetError = L10n.Errors.message(for:)`. A generation counter drops an answer for a sheet that is no longer the open one. `togglePin(subjectID:pinIDs:)` forwards | Lazy per row; refetching on every open is the user's answer for reopening (2026-10-21) | Caching evidence across opens — misses neighbor changes |
| D-14 | **The sheet in today's shape** (user, 2026-10-21). The "This subject" card lists every line of `evidence.sections` in section order. Each relationship is a card titled `L10n.Promote.through(neighbor name)` (the flow row's name, else `sheetNeighbor`), with the neighbor row's `ref`; its header carries a `PVCheckbox` bound to the relationship's bundle ids, disabled when it is not pinnable or has no bundle; its sections' lines follow. A line is one row: the checkbox is bound to its pins' ids and disabled when not pinnable; Compared is `propertyLabel(key: line.propertyKey, origin: provenencia)`; Here is the incoming display (`negatedValue` when negative); the target column lists each pin's display with its Source title (`sourceTitles[sourceID]`) under it; Result is `outcome(line.outcome)`; Weight is `line.points`. The conflict callout shows when any line in any section has outcome `.conflict`. While `evidence` is nil and loading: three `PVSkeletonBar`s in the lines area. On error: `PVCallout(tone: .danger, title: L10n.Promote.evidenceUnavailable, message: error)`. The relationship checkbox's spoken label is `pinLine(property: through(name), here: "", there: "")` | No new chrome, no new string beyond the error title; S11-20 redraws it | Hiding relationships until S11-20 — Done would pin bundles the researcher cannot see |
| D-15 | **No drafted argument** (user, 2026-10-21). A row's argument starts empty and is only what the researcher types. Go drafts none | User decision | Drafting in Go at Apply — declined by the user |
| D-16 | **The layer-once test seam.** `promotealign` gets `var layerBuildHook func()` (nil in production), called first in `loadLayer`; `export_test.go` gets `func CountLayerBuildsForTest(f func()) (restore func())`. No other seam | AC-3.4 needs to count builds; a warm build reads the cached Source graph and issues no query, so `SetQueryCounterForTest` cannot tell | A query counter — blind to a rebuild from the cache |

**Strings.** One new key, for D-14's error title: `L10n.Promote.evidenceUnavailable` = "The evidence could not be loaded" (`promote.sheet.evidenceUnavailable`, comment "Evidence sheet callout title when the evidence read fails").

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `GetPromoteClaimEvidence` handler, `METHOD_GET_PROMOTE_CLAIM_EVIDENCE` and its messages *(new)* | FFI | `ProposePromoteGraphAlignment` | A lazy per-row read with its own inputs; the proposal stops carrying lines |
| `promotealign.PinOffers`, `PinOffer`, `claimEvidenceOffers`, `offer` *(new)* | funcs, types | `PromoteFlow.draftedPins`; `pairExhibits`' `Pinned` | Moves the default-pin rule into Go, beside the query |
| `promotealign.ApplyRow`, `ResolvePins` *(new)* | type, func | `promote.pairClaim` | Replaces the SQL checks with the evidence |
| `promotealign.EvidenceText`, `NewEvidenceText` *(new)* | type, func | `exhibits.go` `valueLabel`, `rowLabeler` | Moved; keyed by the evidence's types instead of `graphalign.Comparison` |
| `promote.Pin` *(new)*, `pinPins` *(renamed)* | type, func | `promote.Pair`, `pinPairs` | Replaced, not beside |
| `claimevidence` `Pin.Value`, `Bundle.Bridge`, `Relationship.Kind`, `Evidence.Kind` *(new fields)* | fields | — | Data `Build` already holds |
| `EvidenceOutcome`, `EvidencePropertyCause`, `ValueRelation`, `ValueCause`, `SegmentCause` *(new)* | Swift enums | `PromoteReasonCode` | Same pattern, same file, same test |
| `CatalogClaimEvidence*` *(new)*, `getPromoteClaimEvidence` *(new)* | Swift types, store method | `CatalogPromoteGraphAlignmentComparison` | That type holds one member; deleted |
| `PromoteFlow.setEvidence`, `evidenceNeighbors`, `Row.evidence` *(new)* | funcs, field | `merge`, `repropose`'s fixed list | Per-row lazy state |
| `PromoteModel.loadEvidence`, `sheetLoading`, `sheetError` *(new)* | func, state | `repropose`, `proposeError` | Same generation pattern |
| `L10n.Promote.evidenceUnavailable` *(new)* | key | `graphUnavailable` | Says what failed on the sheet |
| `layerBuildHook`, `CountLayerBuildsForTest` *(new)* | test seam | `graphcache.Graph.SetQueryCounterForTest` | That counts queries, not builds (D-16) |
| Tests: `promotealign/pins_test.go`, `promotealign/sheettext_test.go`, `handlers/promote_evidence_test.go` *(new)* | tests | `memo_test.go` `newWarmFixture`; `promote_align_test.go` | Reuse those fixtures |

**Builds on.** S11-19's `Evidence`, `loadInputs`, `loadLayerMemo`, the profiler memo (D-22) and `newWarmFixture`; `identityclaims.PinTx`; `withProjectCatalog`; the S12-05 RPC split; `PromoteReasonCode`'s vocabulary pattern; the kit's `PVCheckbox`, `PVSkeletonBar`, `PVCallout`.
**Deletes.** `exhibits.go` (all but `acceptedClaims`), `exhibits_pin_test.go`, `graphalign.Exhibit`, `Row.Exhibits`, `Comparison.ValueType`; `promote.Pair`, `pairClaim`, `pinTargets`, `handleOf`, `sqlOneHop`, `sqlNearHandle`, `bridgeRow`; proto `PromoteGraphAlignmentComparison`, `ApplyPromoteGraphAlignmentPair`, field 9, field 6; `proposalProto`'s exhibit loop; `TestProposeRowWithoutExhibitsSendsValues`; Swift `CatalogPromoteGraphAlignmentComparison`, `CatalogPromoteGraphAlignmentPair`, `Row.comparisons`, `Row.argumentEdited`, `draftedPins`, `draftedArgument`, `ExhibitGroup`. With them goes every `autoreconcile.Compatible` call in Promote.
**Next adds.** S11-20 words lines from the D-10 enums and redraws the sheet on `CatalogClaimEvidence` with no wire change; unchecking one member of a collapsed line is a per-pin toggle, because ids are already per pin. Lifespan and likeness lines (S11-20) are new `claimevidence` lines, and get offers from D-3 with no change to Apply.

### Schema

None — `identity_claim_evidence` already holds one incoming Observation against many member Observations.

### Go core

| Package | Change |
| --- | --- |
| `core/claimevidence` | D-2's four fields, set in `sections.go` `line` and `build.go` `Build` |
| `core/graphalign` | Delete `Exhibit`, `Row.Exhibits`, `Comparison.ValueType` (D-9) |
| `core/database/promote` | `Pin` *(new)*; `Input.Pins`, `BatchRow.Pins`; `pinPins`; deletions (D-5) |
| `core/database/promotealign` | `pins.go` *(new)* (D-3, D-6), `sheettext.go` *(new)* (D-4); `promotealign.go` drops `attachExhibits`; `exhibits.go` deleted, `acceptedClaims` to `evidence.go`; `layer.go` hook (D-16) |

```go
// package promote
// Pin is one piece of evidence the researcher confirmed: Observations pinned
// on the new claim and backfilled onto Member's accepted claim.
type Pin struct {
	Member       []byte
	Observations [][]byte
}

// package promotealign
type PinOffer struct {
	ID      string
	Checked bool
}

func PinOffers(e claimevidence.Evidence) []PinOffer

type ApplyRow struct {
	promote.BatchRow
	DefaultPins bool     // the sheet was never opened: pin what the evidence checks
	PinIDs      []string // the pins the researcher left checked
}

func ResolvePins(c *database.Catalog, sourceID []byte, rows []ApplyRow) ([]promote.BatchRow, error)

type EvidenceText struct{ /* term labels, value types */ }

func NewEvidenceText(g *graphcache.Graph) (EvidenceText, error)
func (t EvidenceText) Incoming(kind vocabkey.SubjectType, s claimevidence.Section, l claimevidence.Line) string
func (t EvidenceText) Member(kind vocabkey.SubjectType, s claimevidence.Section, l claimevidence.Line, p claimevidence.Pin) string
```

Errors: `promote.invalid` (D-5, D-6) as today.

### FFI

`api/proto/engine.proto`, regenerated with `./scripts/generate-proto.sh`:

```proto
// GetPromoteClaimEvidence is one Promote row's claim evidence against its
// target, from the targets the page holds for its neighbors (S11-25).
message GetPromoteClaimEvidenceRequest {
  string project_dir = 1;
  string source_id = 2;
  string subject_id = 3;
  string handle_id = 4;
  repeated PromoteGraphAlignmentFixed neighbors = 5;
}

// One member Observation a line names. id is what Apply takes back.
message PromoteClaimEvidencePin {
  string id = 1;
  string member_subject_id = 2;
  string member_observation_id = 3;
  string source_id = 4;
  string display = 5;
  DateValueInput date = 6;
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
  string incoming_display = 10;
  DateValueInput incoming_date = 11;
  bool incoming_negative = 12;
  repeated PromoteClaimEvidencePin members = 13;
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

message PromoteClaimEvidenceSource {
  string id = 1;
  string title = 2;
}

message GetPromoteClaimEvidenceResponse {
  repeated PromoteClaimEvidenceSection sections = 1;
  repeated PromoteClaimEvidenceRelationship relationships = 2;
  repeated PromoteClaimEvidenceSource sources = 3; // sorted by id
}
```

The handler sets a line's `checked` from `Line.Checked` and a relationship's from `Relationship.Checked`; pin and bundle ids come from `PinOffers`' id rule (D-3), exposed for the handler as `promotealign.LinePinID(incoming, observation []byte) string` and `promotealign.BundlePinID(neighbor []byte, key vocabkey.BridgeKey, bridge []byte) string` *(new)*, which `claimEvidenceOffers` also calls. A date value sets `date` / `incoming_date` with `valuecodec.DateToProto`. Dispatch: `MethodGetPromoteClaimEvidence` constant and `Call` case in `api/ffi/dispatch.go`; a row in `dispatch_test.go`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Line checkbox | `PVCheckbox` | `PromoteView.swift` `PromoteEvidenceSheet.evidenceLine` *(renamed from `comparisonLine`)* | Bound to the line's pin ids (D-14) |
| Relationship card header checkbox | `PVCheckbox` | `PromoteEvidenceSheet` relationship card | Bound to the bundle ids |
| Loading | `PVSkeletonBar` ×3, height 12 | lines area | While `evidence == nil && sheetLoading` |
| Error | `PVCallout(tone: .danger, …, compact: true)` | above the lines | `sheetError` |
| Member column | `VStack` of display + italic muted Source title per pin | existing column | Today's one-member block, repeated |

States: no target handle (New, Skip, unset) → no lines and no fetch, as today's empty comparisons; loading; error; huge (50 members on one line) → the member column grows, the sheet scrolls (today's `ScrollView`). Accessibility: `pinLine` labels; `there` is the pins' displays joined with ", ". Identifiers: `promote.sheet.pin.<first pin id>` for a line, `promote.sheet.relationship.<key>` for a relationship.

### Strings

`L10n.Promote.evidenceUnavailable` *(new)*. Changed signature: `L10n.Promote.outcome(_ outcome: EvidenceOutcome)`; `.unrecognized` reads "Unknown". Removed: none.

## 7. Steps

**STEP-1 — The four fields**
Files: `core/claimevidence/claimevidence.go`, `sections.go`, `build.go`, `sections_test.go`, `build_test.go`. Skill: `add-match-rule`.
Do: D-2. Add the expected `Value`, `Bridge`, `Kind` to existing test expectations; no other expectation changes.
Check: `go test ./core/claimevidence/... ./core/matchregistry/...` passes.

**STEP-2 — Value text**
Files: `core/database/promotealign/sheettext.go` *(new)*, `sheettext_test.go` *(new)*.
Do: D-4. Move `valueLabel` from `exhibits.go` to this file unchanged (one package, so `exhibits.go` still compiles until STEP-5).
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestEvidenceText ./core/database/promotealign/` passes (T-6).

**STEP-3 — One pin shape and Apply's resolver**
Files: `core/database/promote/promote.go`, `batch.go`, `pins_test.go`, `batch_test.go`, `unpromote_test.go`, `core/database/deleteimpact/conclusion_release_test.go`, `core/database/deleteimpact/pins_e2e_test.go`, `core/database/promotealign/pins.go` *(new)*, `pins_test.go` *(new)*, `layer.go`, `export_test.go`. Skill: `use-catalog-session`.
Do: D-5 (each `promote.Pair{IncomingObservationID: a, MemberObservationID: b}` in a test becomes `promote.Pin{Member: <b's Subject>, Observations: {a, b}}`; a `pins_test.go` case that tested a `pairClaim` refusal (outsider, other property, provisional) is deleted, its rule now T-3's); D-3; D-6; D-16.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promote/... ./core/database/promotealign/... ./core/database/deleteimpact/...` passes; T-2, T-3, T-4, T-5 pass.

**STEP-4 — The wire**
Files: `api/proto/engine.proto`, regenerated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/dispatch.go`, `dispatch_test.go`, `api/ffi/handlers/promote_evidence.go` *(new)*, `promote_evidence_test.go` *(new)*, `promote_batch.go`, `promote_batch_test.go`, `promote_align.go`, `promote_align_test.go`. Skill: `add-ffi-handler`.
Do: D-1, D-7, D-8. Delete `proposalProto`'s exhibit loop and `TestProposeRowWithoutExhibitsSendsValues`; drop the comparison assertions in `TestProposePromoteGraphAlignment`. The Go side passes on its own; Swift compiles against the regenerated file only after STEP-8 (see the commit line).
Check: `./scripts/generate-proto.sh` leaves no diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; T-7, T-8 pass.

**STEP-5 — Delete the exhibits**
Files: `core/database/promotealign/exhibits.go` (deleted), `exhibits_pin_test.go` (deleted), `evidence.go`, `promotealign.go`, `promotealign_test.go`, `core/graphalign/types.go`, `rows.go`.
Do: D-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; the AC-4.2 `rg` prints nothing for its Go names.

**STEP-6 — Wire vocabulary**
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/App/Platform/L10n.swift` (`outcome`).
Do: D-10.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestSwiftVocabularyMatchesGo ./api/ffi/` passes (T-9).

**STEP-7 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`. Skill: `add-swift-test`.
Do: D-11. `FakeStore.applyPromoteGraphAlignment`'s skip and new checks read `row.pinIDs.isEmpty` instead of `row.pairs.isEmpty`.
Check: `rg -n "pairs|CatalogPromoteGraphAlignmentPair" macos/App/Platform --glob '!Generated/**'` prints nothing. The app builds only once STEP-8 lands (the flow reads the old types); STEP-8's check covers both.

**STEP-8 — Flow and model**
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `PromoteModel.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`, `PromoteModelTests.swift`. Skill: `add-swift-test`.
Do: D-12, D-13, D-15. Rewrite the pin and argument tests listed in §8; delete the comparison fixtures in both test files' `row(...)` helpers.
Check: `xcodebuild test` (ProvenenciaTests) passes; T-10, T-11 pass.

**STEP-9 — The sheet**
Files: `macos/App/Features/Promote/PromoteView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-14 and the new key.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test` passes; §10 manual check.

**STEP-10 — Docs**
Files: `docs/promote-graph-alignment.md` (§3: the sheet opens lazily and the argument is the researcher's own, not drafted; §9: Go pins what the evidence offers, checked lines and relationships for a sheet never opened, and a pin's Observations land on the new claim and on each member's claim it names; §12: delete the "Exhibit values" bullet and "and exhibits" from the query-count bullet), `docs/conclusion-layer-data-model.md` (the `argument` sentence after the Done list: the researcher writes it; Promote drafts none), `docs/deployment-plan/spike-11/claim-evidence.md` (Status: RPC and pins built in S11-25; Retires: done in S11-25), `docs/deployment-plan/spike-11/comparison-layers.md` (reuse table, layer 2 row: "Promote's claim evidence (`core/claimevidence`)"; duplicates table: `pairExhibits` row "— done in S11-25"), `docs/deployment-plan/spike-11/deployment-plan.md` (map row status), `docs/deployment-plan/spike-11/completed.md`, this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-25-claim-evidence-rpc.md` reports no errors.

Commits: one per step, except that STEP-4 (its Swift part, the regenerated `engine.pb.swift`), STEP-7 and STEP-8 land as one commit so every commit builds in Xcode; say so in the PR. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Load a Promote row's evidence when its sheet opens."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/claimevidence/sections_test.go` `TestSections`, `build_test.go` `TestBuild` | Existing rows with D-2's fields in the expectations: each pin's `Value` is the member's value; each bundle's `Bridge` its bridge Subject; `Relationship.Kind` event; `Evidence.Kind` the row kind | D-2 |
| T-2 | Go | `core/database/promote/pins_test.go` `TestSavePins` (rewritten) | A `Pin` with the incoming and member name Observations → both on the new claim and on the member's claim, `Pins` 2; the same `Pin` twice → written once; `Member` with no accepted claim → `ErrInvalid`; a 4-byte id → `ErrInvalid`; a New row with a pin → `ErrInvalid` | AC-3.5, AC-3.6, D-5 |
| T-3 | Go | `core/database/promotealign/pins_test.go` `TestResolvePins` *(new)*, table-driven | AC-1.1's page: default row → five pins, one per member, each `{Member, [in, mem]}`; opened row with one line id → that pin only; an id not offered → dropped; a conflict line's id → dropped; New row with ids → `ErrInvalid`; default with ids → `ErrInvalid`; a row already on its handle → no pins | AC-3.1, AC-3.2, AC-3.3, D-3, D-6 |
| T-4 | Go | same file `TestResolvePinsBundlesARelationship` *(new)* | `newWarmFixture`, person row on the person handle, birth row on the birth handle, both default → the person row's pins hold the bundle `{member person, Source bridge Observations + Source birth event_type + filed bridge Observations + filed birth event_type}` and the date line pin; `PinOffers` for the RPC's evidence lists the same ids, checked | AC-2.2, AC-2.3, D-3, D-6 |
| T-5 | Go | same file `TestResolvePinsBuildsTheLayerOnce` *(new)* | Cold catalog, three handle rows on one Source, `CountLayerBuildsForTest` → 1; `ProfilesKeptForTest` → kept | AC-3.4, D-16 |
| T-6 | Go | `core/database/promotealign/sheettext_test.go` `TestEvidenceText` *(new)*, table-driven | A name with a form → the form; a name with no form → its parts joined by spaces; a date BEF 2001-03-31 → `datevalue.CompactDisplay`; a `gender` term `man` under section `sex_at_birth` → its catalog label; a term with no label → its key; an integer → digits | D-4, D-9 |
| T-7 | Go | `api/ffi/handlers/promote_evidence_test.go` `TestGetPromoteClaimEvidence` *(new)* | AC-1.1 → one section, one line, five pins with ids, displays, Source ids, five sources; AC-1.2 → one relationship with one bundle id and one date section; AC-1.3 subtests → no relationship; AC-1.4 subtests → error code `promote.invalid` | AC-1.1–AC-1.4, D-1 |
| T-8 | Go | `api/ffi/handlers/promote_batch_test.go` `TestApplyPromoteGraphAlignment` (extended) | AC-1.1's page, row `default_pins` → five member claims each gain the incoming and their own name Observation; the same page with `pin_ids` from T-7's response → the same pins; a row with `default_pins` and `pin_ids` → `promote.invalid`; a conflict-only row default → no pins | AC-2.1, AC-2.3, AC-2.4, D-7 |
| T-9 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` | Five new rows and the "unrecognized" subtest | AC-4.1, D-10 |
| T-10 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` | `aNeverOpenedRowSendsDefaultPins` *(new)*: handle row, no evidence → `defaultPins` true, `pinIDs` empty. `openingDraftsTheCheckedPins` *(new)*: `setEvidence` → `pins == checkedPinIDs`, batch sends them sorted, `defaultPins` false. `reopeningKeepsOnlyOfferedPins` *(new)*: uncheck a line, `setEvidence` with evidence lacking one id → pins lose it and keep the unchecked state. `aLineTogglesEveryMember` *(new)*: `togglePin(pinIDs: [a, b])` twice → both in, then both out. `evidenceForAnOldTargetIsIgnored` *(new)*. `evidenceNeighborsAreHandleRowsAndAnchors` *(new)*: a handle row, a New row, a Skip row, an anchor, a hidden row → only the handle row and the anchor. `aRetargetClearsEvidenceAndPins` (rewrites `aRetargetRedraftsPinsAndArgumentForTheNewTarget`). `everyRowStartsWithAnEmptyArgument` *(new)*. `aSkipRowFilesNoArgumentConfidenceOrPins` reads `pinIDs`. Deleted: `aManualPinOnTheSameTargetSurvivesAReproposal` is rewritten on `setEvidence`; `aConflictingComparisonIsNotDraftedAsAPin` and `aNegatedConflictLineIsNotDraftedAsAPin` are deleted (Go owns drafting: T-3) | AC-1.5, AC-1.6, AC-2.5, D-12, D-15 |
| T-11 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` | `openingASheetLoadsItsEvidence` *(new)*: `FakeStore.promoteEvidence` set → the call is recorded with the handle and the neighbor count, the row's pins are the checked ids. `aFailedEvidenceReadShowsAnError` *(new)*: `promoteEvidenceError` → `sheetError` set, the row keeps no evidence. `doneSendsDefaultOrChosenPins` (rewrites `doneSendsThePinnedPair`): one row never opened, one opened → `lastPromoteBatch` rows carry `defaultPins` true / false with the ids | AC-1.5, AC-1.7, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `promotealign_test.go` (D-9), `core/database/promote` and `deleteimpact` pin tests (D-5), `api/ffi/handlers` Promote tests (D-8), `PromoteFlowTests` (D-12), `core/graphalign` tests that set `ValueType`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings: promote.sheet.evidenceUnavailable; no stale strings left
- [ ] UI uses kit components only (PVCheckbox, PVSkeletonBar, PVCallout); no new kit primitive
- [ ] The AC-4.2 rg prints nothing
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — none: not gated
- [ ] Domain docs: promote-graph-alignment.md §3, §9, §12; conclusion-layer-data-model.md argument sentence; claim-evidence.md status and Retires; comparison-layers.md two rows
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a filed person "John Smith" (two Sources, each with John Smith and a birth in 1817, both promoted onto one person and one birth). Add a third Source with "John Smith", a birth in 1817 and a subject participation. Open Promote on it. Open the person row's sheet: skeleton bars, then one "This subject" card with a checked name line listing two members with their Source titles, and a "Through <birth name>" card with a checked header and a checked date line.
2. Close the sheet without changes. Press Done. Open the filed person's conclusion page: the new claim's pins include the name, the bundle and the date Observations; each earlier member's claim gained the incoming name Observation.
3. Repeat on a fresh copy, but uncheck the date line before Done: the date Observations are not pinned on the new claim.
4. Repeat, set the birth row to Skip, reopen the person sheet: no "Through" card; Done pins only the name.
5. With the catalog closed under the page (or the FakeStore error), the sheet shows the danger callout and the claim fields still work.
6. AC-4.2: run the `rg` command in AC-4.2 from the repo root; it prints nothing.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Done now runs the query once per handle row | Memos (S11-19 §14): T-5; S11-21 times it |
| A neighbor change after opening drops pins quietly | User decision (2026-10-21); reopening shows what is left |
| Go's evidence at Apply differs from the sheet's because of a catalog write between them | `SaveBatch`'s `SeenRevision` refuses the batch (`promote.stale`), and the page re-proposes |

**Stop and report if:**

- `Evidence`'s inputs at Apply (the rows plus `g.Source(id).Members`) cannot reproduce what the sheet was given for the same page, shown by T-4's `PinOffers` ids differing from `ResolvePins`'.
- A filed person member's claim cannot take a bridge Subject's Observation as a pin (a `PinTx` or foreign key error in T-4).
- `identity_claim_evidence`'s primary key is not `(identity_claim_id, observation_id)` at your base.
- A test in `promotealign_test.go` asserts something other than exhibits that changes value after STEP-5.
- `claimevidence` needs an import under `core/database` or of `core/matchregistry` in a non-test file.
- The Xcode build cannot pass with STEP-4 alone and the commits cannot be grouped as the commit line says.

## 12. Open questions

None. Four product calls were answered by the user on 2026-10-21: relationships keep their card with a header checkbox (D-14); a collapsed line is one row listing every member (D-14); no argument is drafted (D-15); pins the evidence no longer offers are dropped (D-6).

## 13. PR

**Title:** S11-25 — Claim evidence RPC and pin policy
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-25-claim-evidence-rpc.md`.
