# S13-23 — Promote stops asking for pins

**Status:** Draft
**Spike / slice:** Spike 13 · Slice 0 — Removing Observation pins
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: working rule "No brief for the removal"; the page loses its sheet and assessment column and draws nothing new)
**Board:** —
**Depends on:** —
**Blocks:** S13-24, S14-02, S14-22
**Branch:** `s13-23-promote-stops-asking-for-pins`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0 until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (removing a `CatalogQueryKey` and a place warm)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [conclusion §5.3](../../../conclusion-layer-data-model.md#53-promote), [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3, §3.1, §8; working rules: "Pins leave in three PRs, the client first", "No brief for the removal", "Promote files accepted claims and nothing else", "The score ranks, unseen", "Spike 14's Promote PRs build on the removal", "Each removal PR updates what it makes false"
**Planned:** 2026-10-29 against `main` @ `577c9df1`
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

Promote becomes a quick filing action in the Mac app. A researcher picks a target per row and presses Done: no row opens an evidence sheet, no row shows a strong, medium or weak band, a Contested mark or a reason, and Done sends no pins, confidence or argument. This is the client half of the pin removal decided in [conclusion §5](../../../conclusion-layer-data-model.md#5-membership-evidence) (docs PR #623) and the first of the three slice 0 PRs (working rule "Pins leave in three PRs, the client first"). After it, no Swift outside the generated proto file names a Method or field that S13-24 removes, apart from the two proposal fields the draft rule still reads (D-4).

**Why now.** Slice 0 runs first in Spike 13. S13-24 (the Go engine) and Spike 14's S14-02 and S14-22 wait on this PR, because it deletes the Promote files and Swift mappings they would otherwise have to edit.

**What stays.** The proposal and its ranking (Go is untouched apart from one Go test file); the target menu, Find another record…, New and Skip; the connection subrows; the conflict and duplicate notes; the Updated and Decided badges and the Updated caption for a moved or via-reached row; the leave guard; Done's one transaction and the stale re-propose; `promoteSubject` (the graph card's single-Subject Promote, which S14-16 deletes); the Identity Claim's `confidenceGradeID` and `argument` on `CatalogIdentityClaim`.

## 2. User stories

**US-1.** As a researcher, when I promote a Source, I pick a target per row and press Done, and nothing asks me to judge evidence, pin observations, grade confidence or write an argument.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Promote page with a row on a suggested record (strong) and a row drafted onto a weak record | Neither row shows a band, a Contested badge, a reason or a chevron; nothing on the row opens a sheet |
| AC-1.2 | The page's column headings | They read Source and Target; the Assessment heading is gone |
| AC-1.3 | Done with one row on a handle, one New, one Skip | `applyPromoteGraphAlignment` receives two rows on handle and New with no `confidenceGradeID`, `argument`, `defaultPins` or `pinIDs` field to send, and a Skip row; the Go request carries `default_pins` false, no `pin_ids`, empty `confidence_grade_id` and `argument` |
| AC-1.4 | The page opens | No `listClaimConfidenceGrades` call is made, and the Promote place warms no confidence grade key |
| AC-1.5 | A row the researcher has not touched | Opening the page and pressing Done is the whole flow: no row's state changes the leave guard except a target change |

**US-2.** As a researcher, when a proposal moves a suggested row, I still see why it moved.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A re-proposal moves row B because settling moved neighbor Gracie to PER-Z (`move_reason` `neighbor_moved`) | Row B shows the Updated badge and "Updated · after Gracie moved to PER-Z" |
| AC-2.2 | A re-proposal moves row B onto a record reached from row A (`via_neighbor_subject_id` = A, A on PER-Z, no move reason) | Row B shows "Updated · via Gracie → PER-Z (spouse)" when the bridge between them has the term spouse, "Updated · via Gracie → PER-Z" when it has none |
| AC-2.3 | A re-proposal moves row B with neither a move reason nor a via neighbor (an unplaced row now drafted onto a weak record) | Row B shows the Updated badge and no caption line |
| AC-2.4 | A via neighbor or a move neighbor that is not a row on the page | The caption reads "Updated · Reached from a matched neighbor" (via) or "Updated · after a connected row changed" (move), as today |

**US-3.** As a contributor, I can regenerate the proto after S13-24 without a Swift edit beyond the draft rule.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `rg 'pinIds\|pinIDs\|defaultPins\|getPromoteClaimEvidence\|ClaimEvidence\|confidenceGradesList\|listClaimConfidenceGrades\|ConfidenceGrade\b' macos/App macos/ProvenenciaTests -g '!**/Generated/**'` | No hit |
| AC-3.2 | `rg '\.score\b\|\.contested\b\|reasonProperty\|\.pins\b' macos/App/Platform/GoStore.swift macos/App/Features/Promote` | No hit; `r.assessment` and `r.reason` are read only where `GoStore.proposePromoteGraphAlignment` maps the row (D-4) |
| AC-3.3 | `go test -tags fts5 ./api/ffi/...` | Passes, with the evidence enum rows and the share-phrase subtest gone from `TestSwiftVocabularyMatchesGo` |

## 3. Scope

| In | Covered by |
| --- | --- |
| Delete the evidence sheet and everything only it uses: `PromoteEvidenceSheet`, `PromoteEvidenceBoard`, `PromoteEvidenceCards`, `PromoteEvidenceDisplay`, their tests, the sheet state in `PromoteFlow` and `PromoteModel` | STEP-1, AC-1.1, AC-1.5 |
| Delete the assessment column: `PromoteReadoutBadges`, the readout button, the band and Contested display, `PromoteFlow.SelectionReadout`, the row and menu items' band and reason fields | STEP-2, AC-1.1, AC-1.2 |
| The Updated caption reads the move reason, else the via neighbor (D-3; user, 2026-10-29) | STEP-2, AC-2.1 – AC-2.4 |
| Delete the grade picker's read: `PromoteModel.grades` and `confidenceKey`, `PromoteView`'s query, the Promote place's warm, `CatalogQueryKey.confidenceGradesList` and its registry entry, `listClaimConfidenceGrades` in `GenealogyStore`, `GoStore`, `FakeStore` and the `OnboardingModelTests` conformer, `CatalogClaimConfidenceGrade` | STEP-3, AC-1.4, AC-3.1 |
| Delete the claim evidence read and wire mapping: `getPromoteClaimEvidence` in `GenealogyStore`, `GoStore` (with every evidence mapper), `FakeStore` and the conformer; `CoreMethod.getPromoteClaimEvidence` and `CoreMethod.listClaimConfidenceGrades`; the `CatalogClaimEvidence*` types and `CatalogEvidenceTerm` | STEP-3, AC-3.1 |
| Done sends no pins, confidence or argument: `CatalogPromoteBatchRow` loses `confidenceGradeID`, `argument`, `defaultPins`, `pinIDs`; `CatalogPromoteBatchWritten` loses `pins`; `target` becomes `PromoteTargetCode` (Spike 11 leftover) | STEP-4, AC-1.3 |
| Swift's proposal row and alternative lose the fields nothing reads after this PR (D-4), and carry one `header: CatalogConclusionHeader?` instead of three optionals (Spike 11 leftover) | STEP-4, AC-3.2 |
| Delete the share wording: `ConnectionSentenceRegistry`'s `KindEntry.share`, `ShareInput`, `share(_:similar:)`, `seededShare`, `namedShare`, `termName`; `L10n.EvidenceGraph.share*`; `L10n.PropertyTerm.Wording`'s `share`, `shareSimilar`, `shareOtherEnd`, `shareOtherEndSimilar`, `recordedDirect` and every entry's values for them; the share tests in `ConnectionSentenceRegistryTests` | STEP-5, AC-3.1 |
| Delete the wire enums `EvidenceOutcome`, `EvidencePropertyCause`, `KinshipPattern`, `ValueRelation`, `ValueCause` and `L10n.Promote.outcome(…)`, `lineSentence(…)` and the sheet's other `L10n.Promote` members | STEP-1, STEP-5, AC-3.3 |
| `api/ffi/swift_vocabulary_test.go`: the five evidence enum rows, their names in the "tolerate a value this build doesn't know" list, the "every term a sentence shares has its share phrases" subtest, and the helpers and imports only they use | STEP-5, AC-3.3 |
| `Localizable.xcstrings` loses every key no Swift names after this PR in `promote.*` and `evidenceGraph.share*`, the stale `promote.toast.filed` included | STEP-1, STEP-2, STEP-5 |
| `project.pbxproj` loses the entries for the deleted Swift files (S14-15 has not landed) | STEP-1, STEP-2 |
| The docs and skills the removal makes false, listed in the working rule "Each removal PR updates what it makes false" for S13-23, plus the Swift-side sentences of `add-match-rule` (D-9) | STEP-6 |
| Map, completed.md and plan status | STEP-7 |

| Out | Home |
| --- | --- |
| The Go evidence engine, `GetPromoteClaimEvidence` and `ListClaimConfidenceGrades` RPCs, Apply's pin and claim-field inputs, the proposal's display-only fields on the wire | S13-24 |
| `PromoteFlow.draftTarget` and the `assessment` and `reason` fields it reads; `PromoteFlow.Assessment` | S13-24 (what a row starts on becomes Go's) |
| The `identity_claim_evidence` table, `identityclaims` pins, the pin delete facet, `PVDeleteImpact`'s pin line, `FakeStore.pinsByObservation`, the pin cases in `PromoteModelTests` (`aPinnedObservationDeleteNamesThePerson`) and `PVDeleteImpactTests` | S13-25 |
| `promoteSubject` and its `confidenceGradeID` / `argument` parameters | S14-16 |
| `viaBridgeType` and `viaRole` on `CatalogPromoteGraphAlignmentRow`: mapped and unread today, not on S13-24's reserve list | not planned (dead mapping, one line each; the next PR that edits the proposal mapping drops them) |
| Any change to proposals, the ranking or the band in Go | not planned (working rule "The score ranks, unseen") |
| A claim editor that sets confidence and argument after filing | [identity-claim-review idea](../../../ideas/identity-claim-review.md) |
| Synchronized Xcode folders | S14-15 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Sheet | `macos/App/Features/Promote/PromoteEvidenceSheet.swift`, `PromoteEvidenceBoard.swift`, `PromoteEvidenceCards.swift`, `PromoteEvidenceDisplay.swift` | The row's evidence sheet: board of lines and relationship cards, pin checkboxes, status, confidence picker, argument field. Only callers: each other, `PromoteView` (`.pvSheet`), `PromoteModel.board(of:)`, `PromoteFlow.Row.board` |
| Sheet tests | `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift`, `PromoteEvidenceDisplayTests.swift` | Test only the files above |
| Page model | `macos/App/Features/Promote/PromoteModel.swift` `PromoteModel` | Holds `SheetEvidence`, `sheetEvidence`, `evidenceGeneration`, `grades`, `confidenceKey`, `propertyLabels`, `propertyLabelsByKey`; `togglePin`, `setArgument`, `setConfidence`, `openSheet`, `closeSheet`, `reloadEvidence`, `setSheetSubject`, `loadEvidence`, `board(of:)`, `assessmentText`, `propertyLabel(key:origin:)`, `propertyLabel(key:)`; `reasonText(_:on:)` words every `PromoteFlow.Reason` case; `load()` reads `confidenceKey`; `graphSnapshot()` fills the label maps from `propertiesKey` and reads its `types` |
| Row list | `macos/App/Features/Promote/PromoteFlow.swift` `PromoteFlow` | `Assessment`; `Reason` (12 cases, three `init`s from row, alternative and move); `SelectionReadout`; `Alternative.readingTitle`, `.assessment`, `.reason`; `Row.assessment`, `.reason`, `.board`, `.pins`, `.confidenceGradeID`, `.argument`, `.contested`, `.move`, `.readingTitle`, `.selectionReadout`, `.captionReason`; `sheetSubjectID`; `evidenceNeighbors(for:)`, `setEvidence`, `togglePin`, `setArgument`, `setConfidence`; `batchRow(_:)` builds `defaultPins`, `pinIDs`, grade and argument and a string target; `item(id:ref:kind:person:event:place:…)` rebuilds a `CatalogConclusionHeader` from three optionals; `draftTarget` reads `Assessment(wire: proposal.assessment)` and `Reason(proposal) == .taken` |
| Page view | `macos/App/Features/Promote/PromoteView.swift` `PromoteView`, `PromoteAlignmentRow`, `PromoteAnchorRow`, `PromoteLayout` | `.task` queries `confidenceKey`; `.pvSheet(item: sheetBinding)`; `columnHeaders` shows Source, Target, Assessment; the row's assessment cell is a button with `PromoteReadoutBadges`, the reason and a chevron, identifier `promote.row.<ref>.evidence`; the anchor row puts "Filed earlier" in that column; caption `if row.updated` with `row.captionReason` |
| Readout | `macos/App/Features/Promote/PromoteReadout.swift` `PromoteReadoutBadges`, `PromoteUpdatedCaption` | Badges for the row and sheet title; the caption for the row and sheet |
| Store protocol | `macos/App/Platform/GenealogyStore.swift` | `CatalogClaimConfidenceGrade`; `CatalogClaimEvidencePin`, `…Line`, `…Section`, `…Bundle`, `CatalogEvidenceTerm`, `CatalogClaimEvidenceShape`, `…Relationship`, `…Offer`, `CatalogClaimEvidence`; `CatalogPromoteGraphAlignmentAlternative` (`score`, `person`/`event`/`place`, `assessment`, `reason`, `reasonPropertyKey`, `reasonPropertyOrigin`, `viaNeighborSubjectID`); `CatalogPromoteGraphAlignmentRow` (`score`, `assessment`, `reason`, `reasonPropertyKey`, `reasonPropertyOrigin`, `contested`, `person`/`event`/`place`, …); `CatalogPromoteBatchRow` (`target: String`, `confidenceGradeID`, `argument`, `defaultPins`, `pinIDs`); `CatalogPromoteBatchWritten.pins`; protocol `getPromoteClaimEvidence`, `listClaimConfidenceGrades` |
| Go store | `macos/App/Platform/GoStore.swift` `proposePromoteGraphAlignment`, `getPromoteClaimEvidence`, `mapEvidenceTerm`, `mapEvidenceShape`, `mapEvidenceSection`, `applyPromoteGraphAlignment`, `listClaimConfidenceGrades`, `mapConclusionHeader` | Maps every proposal field; maps the evidence; sends grade, argument, `defaultPins`, `pinIds`; reads `w.pins` |
| Method ids | `macos/App/Platform/CoreInvoke.swift` `CoreMethod.listClaimConfidenceGrades`, `.getPromoteClaimEvidence` | Used only by the two `GoStore` calls |
| Fake store | `macos/App/Platform/FakeStore.swift` `promoteEvidence`, `promoteEvidenceError`, `proposePromoteGraphAlignment`, `getPromoteClaimEvidence`, `applyPromoteGraphAlignment`, `listClaimConfidenceGrades` | Builds proposal rows with `score`/`assessment`/`reason`; checks Skip carries no pins, grade or argument and New no pins; writes `pins: row.pinIDs.count` |
| Conformer | `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore` | Implements `getPromoteClaimEvidence`, `listClaimConfidenceGrades` |
| Query key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.confidenceGradesList`, `Kind.confidenceGradesList`; `CatalogQueryRegistry.swift` its spec; `PlaceRegistry.swift` the Promote place's `queryKeys` | Session-fresh grades list warmed on the Promote place |
| Key tests | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `confidenceGradesLoadInOrder`, `invalidationsComeFromRegistrySpecs`; `PlaceRegistryTests.swift` `queryKeysMatchDesignTable`; `SubjectStoreTests.swift` `promoteJoinsExistingHandleWithGradeAndArgument` (last two lines read grades) | Name the key or the store call |
| Share wording | `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift` `KindEntry.share`, `ShareInput`, `termName`, `seededShare`, `namedShare`, `share(_:similar:)`, every entry's `share:` closure | Only caller: `PromoteEvidenceDisplay` |
| Share tests | `macos/ProvenenciaTests/ConnectionSentenceRegistryTests.swift` from `// MARK: - What a row shares with a connected record` (helpers `term`, `share`; tests `aNeighborEventIsNamedByItsType` … `nothingNamedIsAConnection`) | Test only the share wording |
| Wire enums | `macos/App/Platform/ConclusionVocabulary.swift` `EvidenceOutcome`, `EvidencePropertyCause`, `KinshipPattern`, `ValueRelation`, `ValueCause` | Read only by the evidence types, `L10n.Promote.outcome`, `lineSentence` and the sheet |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Promote` (lines 1316–1792), `L10n.EvidenceGraph` share section (from line 2417), `L10n.PropertyTerm.Wording` (line 2637) and its 77 share / `recordedDirect` arguments | Sheet, column and share copy beside page copy that stays |
| Catalog | `macos/App/Resources/Localizable.xcstrings` | 172 `evidenceGraph.share*`, `promote.sheet.*`, `promote.assessment.*` keys, plus the column, readout and reason keys; `promote.toast.filed` is already named by no Swift |
| Parity test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` rows "evidence outcomes", "evidence property causes", "kinship patterns", "value relations", "value causes"; the tolerate list; subtest "every term a sentence shares has its share phrases"; helpers `patternCodes`, `sortedRelations` | Hold the five enums and the share fields to Go |
| Flow tests | `macos/ProvenenciaTests/PromoteFlowTests.swift`, `PromoteModelTests.swift` | Cover drafting, readout, contested, caption, pins, argument, evidence, sheet and the rest of the page (test names in §8) |
| Project | `macos/Provenencia.xcodeproj/project.pbxproj` | 24 `PromoteEvidence*` entries (4 per file: build file, file reference, group child, sources phase) and 4 for `PromoteReadout.swift`; not on synchronized folders (`grep -c fileSystemSynchronized` is 0) |
| Docs | [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) line 7 ("Until the code catches up …"), principle 5 ("every conflict … shows as *contested*"), §3.1, §4 `Proposal` list, §8; [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) line 693; [`matching-model.md`](../../../matching-model.md) lines 5, 122 | Describe the sheet or the band as current, or as going |
| Skills | `.agents/skills/add-localized-string/SKILL.md` lines 147, 151; `.agents/skills/add-seeded-vocabulary/SKILL.md` line 80; `.agents/skills/add-match-rule/SKILL.md` lines 76, 91 | Cite `PromoteEvidenceDisplay.Role`, `promote.sheet.subtitle.*`, share phrases, `recordedDirect`, the five Swift enums and `PromoteEvidenceDisplay.wording`; `scripts/check-guidance-names.py` fails once those names are gone |
| Ideas | `docs/ideas/term-management.md` line 28, `docs/ideas/international-names.md` line 43, `docs/ideas/binding-order-editing.md` line 7 | Describe the sheet as current |

**Precedent.** `35145e44` ("Send a decided row's New or Skip target as a typed PromoteTargetCode instead of a raw string") typed `CatalogPromoteGraphAlignmentFixed.target`: copy it for `CatalogPromoteBatchRow.target` (D-6). `mapConclusionHeader` in `GoStore.swift` is the shape for the row and alternative header mappers (D-7). `1d220bb6` ("Remove the per-graph link and unlink calls …") is a Swift-and-store removal of the same kind: protocol, `GoStore`, `FakeStore`, conformer and tests in one change.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Done's batch row is `CatalogPromoteBatchRow(subjectID:target:entityID:)` and nothing else. `GoStore.applyPromoteGraphAlignment` sets `subjectID`, `target`, `entityID` only; Go reads unset `pin_ids`, `default_pins` false, empty grade and argument as an accepted claim with no pins (`ResolvePins` skips a row with neither) | Working rule "Promote files accepted claims and nothing else" | Sending `default_pins: false` explicitly — the proto default already is false, and the field goes in S13-24 |
| D-2 | `CatalogPromoteBatchResult.written` keeps `entity` and `claim`; `CatalogPromoteBatchWritten.pins` goes. Nothing reads `written` today beyond tests; it stays because S14-16 rewrites the Apply mapping | Smallest change that names no pin | Deleting `written` — S14-16 owns the Apply shape |
| D-3 | The Updated caption: `Row.caption: Reason?` *(new field, replaces `move` and `reason`)* is the move reason, else `.via(neighborID:)` when the proposal's `viaNeighborSubjectID` is not empty, else nil. The view draws `PromoteUpdatedCaption` only when `row.updated` and `row.caption != nil`. A row moved with neither shows the Updated badge alone (user, 2026-10-29: "Move, else via") | Keeps §3.1's "via Gracie → PER-Z" caption without reading the `reason` field S13-24 reserves; `via_neighbor_subject_id` is set only when the walk reached the row from a neighbor (`api/ffi/handlers/promote_align.go` `proposalProto`, `r.Via != nil`) and S13-24 keeps it | Move reasons only (the map's first wording) — the user chose to keep via; showing "Updated · Nothing to match on" for the rest — it words a reason the page no longer shows |
| D-4 | `PromoteFlow.Reason` keeps four cases: `via(neighborID:)`, `neighborPlaced`, `neighborMoved`, `neighborReleased`. `init?(move:)` returns nil for a code that is not a move code; `init?(via:)` *(new)* reads `viaNeighborSubjectID`. `Swift`'s proposal row keeps `assessment` and `reason` only because `draftTarget` reads them until S13-24; `draftTarget` tests `proposal.reason == .taken` on `PromoteReasonCode` instead of building a `Reason`. `PromoteFlow.Assessment` stays for `draftTarget` alone. `PromoteReasonCode` keeps every case (it mirrors `graphalign.Reason`, which S13-24 shrinks) | Every other case was worded only by the readout; the draft rule leaves in S13-24 with its fields | Removing `assessment` and `reason` now — `draftTarget` is S13-24's (map scope) |
| D-5 | Swift's proposal types lose every field no Swift reads after this PR: `CatalogPromoteGraphAlignmentRow` loses `score`, `reasonPropertyKey`, `reasonPropertyOrigin`, `contested`; `CatalogPromoteGraphAlignmentAlternative` loses `score`, `assessment`, `reason`, `reasonPropertyKey`, `reasonPropertyOrigin`, `viaNeighborSubjectID`. `viaBridgeType` and `viaRole` stay (Out) | S13-24 reserves these proto fields and its scope allows only the draft rule as a Swift edit | Keeping the mapping until S13-24 — its regeneration would then break the Swift build |
| D-6 | `CatalogPromoteBatchRow.target` is `PromoteTargetCode`; `PromoteFlow.batchRow` sets `.handle`, `.new` or `.skip`; `GoStore` sends `row.target.wire`; `FakeStore` switches on the code. An unset row never reaches `batchRow` (`filedRows` filters it) | Spike 11 leftover; `PromoteTargetCode.wire` already exists for `Fixed` (`35145e44`) | A second enum for the batch — one wire vocabulary for both |
| D-7 | `CatalogPromoteGraphAlignmentRow` and `…Alternative` carry `header: CatalogConclusionHeader?` *(new field)* instead of `person`, `event`, `place`. `GoStore` maps each oneof with a private static mapper beside `mapConclusionHeader`: `mapPromoteRowHeader` *(new)* and `mapPromoteAlternativeHeader` *(new)*. `PromoteFlow.item(id:ref:header:orders:)` takes the header; `hold` passes its header straight through. The `kind` filter in `item` goes: the header's case is its kind | Spike 11 leftover; one header type through the store | Keeping three optionals and a builder in the flow — the store split what the flow rebuilds |
| D-8 | The assessment slot stays as empty space: the working row's assessment `HStack` becomes `Color.clear.frame(maxWidth: .infinity, maxHeight: 1)`, the column header's third label becomes `Color.clear.frame(maxWidth: .infinity, maxHeight: 1)`, and `PromoteAnchorRow` keeps "Filed earlier" in that slot. `PromoteLayout` is unchanged | Working rule "No brief for the removal": the target menu, subrows and notes stay where they are and nothing new is drawn | Collapsing the column — the target column and anchors would move, a layout change with no board |
| D-9 | `PromoteUpdatedCaption` moves into `PromoteView.swift` as a `private struct`, and `PromoteReadout.swift` is deleted with its preview | Its only caller after the sheet goes is the row | Keeping a file named for a readout that no longer exists |
| D-10 | `PromoteModel` keeps the `propertiesKey` read for the snapshot's `types`; `propertyLabels`, `propertyLabelsByKey` and both `propertyLabel` functions go. `syncNameOrders` stays (load and re-propose title rows); its comment drops the sheet | The labels served the readout's `agrees` reason and the sheet | Dropping the query — `EvidenceGraphSnapshot.build` needs the types |
| D-11 | Every `L10n.Promote` member and catalog key named only by the deleted files or the deleted column goes; every share member of `L10n.EvidenceGraph` goes; `L10n.PropertyTerm.Wording` keeps its other fields. A key stays only if Swift in `macos/App` still names it. STEP-5's check command proves it, and `promote.toast.filed` (named by nothing today) goes with them | One sweep, checked by a command, not a hand list of 170 keys | A hand-written key list in this plan — it drifts the moment a key is renamed |
| D-12 | S13-23 edits `add-match-rule` lines 76 and 91 to drop the Swift enums and `PromoteEvidenceDisplay` (the Swift step for a wire code); the rest of `add-match-rule` stays S13-24's. The map's working rule "Each removal PR updates what it makes false" gains this line in STEP-6 | `scripts/check-guidance-names.py` (run by `check-pr.sh`) fails once those names are gone | Leaving it to S13-24 — this PR's gate would fail |
| D-13 | Tests: a test whose subject is deleted goes whole; a test that covers kept behavior and also asserts a deleted field drops only that assertion. The kept and changed tests are named in §8 | Keeps the coverage of what stays | Rewriting the suites |

**Rules this plan bends.** None. D-12 moves two sentences of `add-match-rule` from S13-24 to S13-23 inside the working rule's own list; STEP-6 records it in the map.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PromoteFlow.Row.caption` *(new)* | field | `PromoteFlow.Row.move`, `captionReason` | Replaces both: one stored value the view reads |
| `PromoteFlow.Reason.init?(via:)` *(new)* | func | `PromoteFlow.Reason.init?(move:)` | Same shape for the other caption source |
| `CatalogPromoteGraphAlignmentRow.header`, `CatalogPromoteGraphAlignmentAlternative.header` *(new)* | field | `CatalogSearchHit`'s header, `CatalogConclusionDetail.header` (`GenealogyStore.swift` lines 242, 820) | Reuses `CatalogConclusionHeader`; replaces three optionals |
| `GoStore.mapPromoteRowHeader`, `GoStore.mapPromoteAlternativeHeader` *(new)* | func | `GoStore.mapConclusionHeader` | Each proto message has its own `OneOf_Header` type; the bodies copy `mapConclusionHeader` |

Searched for: an existing caption source type (none beyond `Reason`), a typed batch target (`PromoteTargetCode`, reused), a header enum (`CatalogConclusionHeader`, reused).

**Builds on.** `PromoteTargetCode.wire` (`35145e44`); `CatalogConclusionHeader` and `mapConclusionHeader`; the place registry's per-place `queryKeys` (`add-workspace-place`).
**Deletes.** Files: `PromoteEvidenceSheet.swift`, `PromoteEvidenceBoard.swift`, `PromoteEvidenceCards.swift`, `PromoteEvidenceDisplay.swift`, `PromoteReadout.swift`, `PromoteEvidenceBoardTests.swift`, `PromoteEvidenceDisplayTests.swift` and their `project.pbxproj` entries. Types: `PromoteModel.SheetEvidence`, `PromoteFlow.SelectionReadout`, `EvidenceBoard` (in `PromoteEvidenceBoard.swift`), `PromoteReadoutBadges`, `CatalogClaimConfidenceGrade`, every `CatalogClaimEvidence*` type, `CatalogEvidenceTerm`, `ConnectionSentenceRegistry.ShareInput`, the five wire enums. Members: listed in §3 In and §4. Mechanisms: the sheet's evidence read per row; the grade warm; the share wording table; default pins on Done. Callees whose last caller goes: `PromoteFlow.evidenceNeighbors(for:)`, `PromoteFlow.Row.readingTitle`, `PromoteFlow.Alternative.readingTitle`, `PromoteFlow.Row.shownRef` stays (via wording reads it), `PromoteFlow.bridge(between:and:)` stays (via wording reads it), `GoStore.mapEvidenceTerm`, `mapEvidenceShape`, `mapEvidenceSection`, `ConnectionSentenceRegistry.termName`, `seededShare`, `namedShare`, `L10n.EvidenceGraph.share` (private helper), `patternCodes`, `sortedRelations` and the Go imports only they used.
**Next adds.** S13-24 deletes `draftTarget`, `PromoteFlow.Assessment` and the row's `assessment` and `reason` with one Swift edit. S14-22 changes the Promote rows on a page with no column to work around. S14-02 is re-planned without renaming the deleted Promote files.

### Reach

None — no new path into existing code. The page loses a sheet and a column; the caption shows for fewer rows, never for a new state.

### Scale

At the map's Scale (a Source of 100 new people on one Promote page): the page made one `getPromoteClaimEvidence` call per opened sheet and warmed one grades read; after this PR it makes none. Done sends one Apply with 100 rows of three fields. No new loop; `makeRow` and `merge` do less work per row. No budget changes.

### Edges

- **Assumptions.** The caption assumes `viaNeighborSubjectID` is set only with `reason` `via` (`proposalProto` sets it under `r.Via != nil`; `core/graphalign/rows.go` sets `row.Via` only with `ReasonVia`). T-3 holds it with a row that has both and one that has neither.
- **Neighbors.** A row with a move and a via neighbor shows the move (AC-2.1, T-3); one with only a via neighbor shows via (AC-2.2, T-3); one with neither shows no caption (AC-2.3, T-3); a decided row shows no caption whatever it carries (`updated` is false; T-4).
- **Removed inputs.** A retarget clears `caption` as it cleared `move` (T-4). Nothing else on the row is derived from a deleted field.
- **Real data.** None — no table- or seed-driven rule changes.
- **Untrusted size.** None — no new list read from data the app did not write.

### Schema

None — no catalog change.

### Go core

None — Go code is untouched; the only Go edit is `api/ffi/swift_vocabulary_test.go` (STEP-5).

### FFI

None — no proto or handler change. Swift stops calling two Methods and sending four fields; S13-24 reserves them.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Row grid without the assessment cell | `PromoteAlignmentRow` (existing) | `macos/App/Features/Promote/PromoteView.swift` | D-8: the slot is empty space; the Updated and Decided badges, target menu and subrows stay |
| Column headers | `PromoteView.columnHeaders` (existing) | same | Source, Target, then empty space and the badge column |
| Updated caption | `PromoteUpdatedCaption` (moved, private) | same | D-3, D-9 |

States: a row with nothing to say shows the badge only. Accessibility: the evidence button and its label (`L10n.Promote.evidenceFor`) go; the target menu's spoken label (`L10n.Promote.filesOn`) and the connection labels stay. Identifier `promote.row.<ref>.evidence` goes; no UI test names it (`rg 'row\..*\.evidence' macos` finds only `PromoteView.swift`).

### Strings

Removed: every `promote.sheet.*` key; `promote.assessment.*`; `promote.assessment.chosenNone`; `promote.row.contested`, `promote.row.assessmentContested`, `promote.row.evidenceFor`; `promote.page.column.assessment`; the reason keys off the caption path (`promote.reason.filed`, `.agrees`, `.decided`, `.weak`, `.taken`, `.unplaced`, `.empty`); every `evidenceGraph.share*` key; `promote.toast.filed`. Kept on the caption path: `promote.row.updatedDetail`, `promote.reason.via`, `.viaNeighbor`, `.viaUnknown`, `.neighborPlaced`, `.neighborMoved`, `.neighborReleased`, `.moveUnknown`. No key is added.

Old-scope copy: none left. The Done summary, toast and leave confirm do not mention evidence, pins or confidence (`rg -i 'pin|evidence|confidence' ` over the kept `L10n.Promote` values finds only removed members; STEP-2 check).

## 7. Steps

**STEP-1 — Delete the evidence sheet**
Files: delete `macos/App/Features/Promote/PromoteEvidenceSheet.swift`, `PromoteEvidenceBoard.swift`, `PromoteEvidenceCards.swift`, `PromoteEvidenceDisplay.swift`, `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift`, `PromoteEvidenceDisplayTests.swift`; edit `PromoteModel.swift`, `PromoteFlow.swift`, `PromoteView.swift`, `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`, `PromoteFlowTests.swift`, `PromoteModelTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: remove the 24 `PromoteEvidence*` lines from `project.pbxproj` (`grep -n PromoteEvidence` lists them). In `PromoteModel` remove `SheetEvidence`, `sheetEvidence`, `evidenceGeneration`, `togglePin`, `setArgument`, `setConfidence`, `openSheet`, `closeSheet`, `reloadEvidence`, `setSheetSubject`, `loadEvidence`, `board(of:)`, `propertyLabel(key:)` and `propertyLabelsByKey` (D-10). In `PromoteFlow` remove `sheetSubjectID`, `evidenceNeighbors(for:)`, `setEvidence`, `togglePin`, `setArgument`, `setConfidence`, `Row.board`, `Row.pins`, `Row.confidenceGradeID`, `Row.argument`, `Row.readingTitle`, `Alternative.readingTitle`, and their uses in `select`, `makeRow`, `item`. `batchRow` sends no grade, argument or pins yet still compiles against today's `CatalogPromoteBatchRow` by passing `confidenceGradeID: nil, argument: ""` and default `defaultPins` / `pinIDs` (STEP-4 removes the fields). In `PromoteView` remove `.pvSheet` and `sheetBinding`; the row button stays until STEP-2 but its action is removed: STEP-1 replaces `Button { model.openSheet(row.subjectID) } label: { … }` with its label content alone. Remove the `L10n.Promote` members only the sheet used (the sheet section from `backToRows` to `evidenceRetry`, `sheetTitleNew`, `evidenceUnavailable`, `statusAccepted`, `whySuggested`, `whySuggestedSpoken`, `outcome`, `notCompared`, `lineSentence` and the `line*`, `spoken*`, `recorded*`, `subtitle*`, `evidenceHeading*`, `share*` members) and their catalog keys. Drop the tests of deleted behavior listed for STEP-1 in §8.
Check: `rg -n 'PromoteEvidence|EvidenceBoard|sheetSubjectID|openSheet|togglePin|setArgument|setConfidence|promote\.sheet' macos -g '!**/Generated/**'` finds nothing; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-2 — Remove the assessment column and narrow the caption**
Files: `PromoteFlow.swift`, `PromoteModel.swift`, `PromoteView.swift`, delete `PromoteReadout.swift`, `L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`, `PromoteFlowTests.swift`, `PromoteModelTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: per D-3, D-4, D-8, D-9, D-10. `Reason` keeps `via`, `neighborPlaced`, `neighborMoved`, `neighborReleased`; add `init?(via:)`; `init?(move:)` returns nil on a non-move code. Remove `SelectionReadout`, `Row.assessment`, `Row.reason`, `Row.contested`, `Row.selectionReadout`, `Row.captionReason`, `Row.move` (replaced by `Row.caption`), `Alternative.assessment`, `Alternative.reason`. `makeRow` sets `caption` from the proposal for a suggested row and for a kept decided row; `select` clears `caption` where it cleared `move` and `contested`. `draftTarget` reads `Assessment(wire: proposal.assessment)` and `proposal.reason == .taken`. In `PromoteModel` remove `assessmentText`, `propertyLabels`, `propertyLabel(key:origin:)`; `reasonText` keeps the `.via` and move arms. In `PromoteView` the assessment cell and the header label become empty space (D-8); the caption reads `if row.updated, let caption = row.caption`. Move `PromoteUpdatedCaption` into `PromoteView.swift` as `private struct`; delete `PromoteReadout.swift` and its 4 `project.pbxproj` lines. Remove `L10n.Promote` `assessmentStrong`, `assessmentMedium`, `assessmentWeak`, `assessmentNone`, `assessmentChosenNone`, `assessment(_:)` (both), `contested`, `assessmentContested`, `evidenceFor`, `columnAssessment`, `alreadyFiled`, `agreesOn`, `reasonDecided`, `reasonWeak`, `reasonTaken`, `reasonUnplaced`, `reasonEmpty` and their keys.
Check: `rg -n 'SelectionReadout|PromoteReadout|captionReason|assessmentText|\.contested|columnAssessment' macos -g '!**/Generated/**'` finds nothing; `rg -n 'PromoteReadout' macos/Provenencia.xcodeproj/project.pbxproj` finds nothing; Xcode tests pass with T-1 – T-4.

**STEP-3 — Remove the grade read and the evidence read from the store**
Files: `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`, `CatalogQueryKey.swift`, `CatalogQueryRegistry.swift`, `PlaceRegistry.swift`, `PromoteModel.swift`, `PromoteView.swift`, `OnboardingModelTests.swift`, `CatalogQueryRegistryTests.swift`, `PlaceRegistryTests.swift`, `SubjectStoreTests.swift`. Skill: `add-workspace-place`, `add-swift-test`.
Do: remove `grades`, `confidenceKey`, the `load()` read and `PromoteView`'s `QueryHandle<[CatalogClaimConfidenceGrade]>` line; remove `.confidenceGradesList` from `CatalogQueryKey` (case, `Kind`, `kind` and `project` arms), its `CatalogQueryRegistry` spec and the Promote place's `queryKeys`. Remove `getPromoteClaimEvidence` and `listClaimConfidenceGrades` from `GenealogyStore`, `GoStore`, `FakeStore` (with `promoteEvidence`, `promoteEvidenceError`) and `ThrowingStore`; `CoreMethod.getPromoteClaimEvidence`, `.listClaimConfidenceGrades`; `mapEvidenceTerm`, `mapEvidenceShape`, `mapEvidenceSection`; `CatalogClaimConfidenceGrade`, `CatalogClaimEvidencePin`, `…Line`, `…Section`, `…Bundle`, `CatalogClaimEvidenceShape`, `…Relationship`, `…Offer`, `CatalogClaimEvidence`. `CatalogEvidenceTerm` stays until STEP-5, because `ConnectionSentenceRegistry` still names it. Tests per §8.
Check: `rg -n 'getPromoteClaimEvidence|listClaimConfidenceGrades|confidenceGradesList|CatalogClaimEvidence|ConfidenceGrade\b' macos -g '!**/Generated/**'` finds nothing; Xcode tests pass.

**STEP-4 — Done sends targets only; the proposal row carries what the page reads**
Files: `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `PromoteFlow.swift`, `PromoteFlowTests.swift`, `PromoteModelTests.swift`, any test that builds `CatalogPromoteGraphAlignmentRow` or `…Alternative`. Skill: `add-swift-test`.
Do: per D-1, D-2, D-5, D-6, D-7. `CatalogPromoteBatchRow` is `subjectID`, `target: PromoteTargetCode`, `entityID`; `CatalogPromoteBatchWritten` loses `pins`; `GoStore.applyPromoteGraphAlignment` sends `proto.target = row.target.wire` and sets no grade, argument or pin field. `FakeStore.applyPromoteGraphAlignment` switches on `.skip`, `.new`, `.handle` (anything else throws `promote.invalid`) and drops the pin, grade and argument guards; its claim has `confidenceGradeID: nil` and `argument: ""`; its comment drops `PinOffers`. Remove the D-5 fields from the two proposal types, `GoStore`'s mapping and `FakeStore`'s proposal rows. Replace `person` / `event` / `place` with `header` (D-7); `PromoteFlow.item(id:ref:header:orders:)`.
Check: AC-3.2's `rg` holds; `rg -n 'person: proposal\.person|alt\.person|altPerson' macos/App` finds nothing; Xcode tests pass with T-5, T-6.

**STEP-5 — Delete the share wording and the wire enums**
Files: `ConnectionSentenceRegistry.swift`, `ConnectionSentenceRegistryTests.swift`, `L10n.swift`, `Localizable.xcstrings`, `ConclusionVocabulary.swift`, `GenealogyStore.swift`, `api/ffi/swift_vocabulary_test.go`. Skill: `add-localized-string`; `.cursor/rules/go-tests.mdc`.
Do: remove `KindEntry.share` and every entry's `share:` closure, `ShareInput`, `termName`, `seededShare`, `namedShare`, `share(_:similar:)`; `CatalogEvidenceTerm`; the `// MARK: What a row shares` section of `L10n.EvidenceGraph`; `Wording`'s `share`, `shareSimilar`, `shareOtherEnd`, `shareOtherEndSimilar`, `recordedDirect`, every entry's arguments for them and the three private `recordedDirect*` constants; the five enums in `ConclusionVocabulary.swift`; the share tests and helpers from `// MARK: - What a row shares with a connected record` to the end of that section. Remove the stale catalog keys (D-11). In `swift_vocabulary_test.go` remove the five rows, their names from the tolerate list (it keeps `ReconciledState`, `ReconcilerReason`, `PromoteReasonCode`), the share subtest, `patternCodes`, `sortedRelations` and the imports `go vet` reports unused.
Check: this prints `[]`:
```bash
python3 - <<'EOF'
import json, pathlib
cat = json.load(open('macos/App/Resources/Localizable.xcstrings'))['strings']
src = ''.join(p.read_text() for p in pathlib.Path('macos/App').rglob('*.swift'))
print([k for k in cat if k.startswith(('promote.', 'evidenceGraph.share')) and f'"{k}"' not in src])
EOF
```
`rg -n 'EvidenceOutcome|EvidencePropertyCause|KinshipPattern|ValueRelation|ValueCause|ShareInput|recordedDirect|shareSimilar|shareOtherEnd' macos -g '!**/Generated/**'` finds nothing; `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes; Xcode tests pass.

**STEP-6 — Docs and skills**
Files: `docs/promote-graph-alignment.md`, `docs/interpretation-layer-data-model.md`, `docs/matching-model.md`, `.agents/skills/add-localized-string/SKILL.md`, `.agents/skills/add-seeded-vocabulary/SKILL.md`, `.agents/skills/add-match-rule/SKILL.md`, `docs/ideas/term-management.md`, `docs/ideas/international-names.md`, `docs/ideas/binding-order-editing.md`, `docs/deployment-plan/spike-13/deployment-plan.md`. Skill: —.
Do: `promote-graph-alignment.md`: delete line 7's last sentence ("Until the code catches up, the sheet and the pin writes described in older PR plans still run."); principle 5 ends at "is surfaced" (the Contested clause goes); §3.1's caption bullet says the mark carries the move, else the via neighbor ("via Gracie → PER-Z"), else nothing; the contested bullet stays; §4's `Proposal` list and §8 need no change (verify; §8 already says decisions live in the page's draft). `interpretation-layer-data-model.md` line 693: `sort_order` orders Add-property menus, Properties lists and the conclusion page's fields; drop the sheet and "The sheet is to be removed". `matching-model.md` lines 5 and 122: the sheet is gone; the `evidence` section is removed by S13-24. `add-localized-string` line 147: cite a kept role-attribute helper or drop the parenthesis; line 151: the per-kind sentence example becomes `promote.page.kind.person` style or another kept per-kind key (`L10n.Promote.newOption(_:)`'s keys), and "outside the evidence sheet" goes. `add-seeded-vocabulary` line 80: drop the share phrase, similar phrase and `recordedDirect` clause. `add-match-rule` lines 76 and 91 (D-12): a wire code needs no Swift enum now; drop the enum list, the `KinshipPattern` step's Swift half and `PromoteEvidenceDisplay.wording`. Ideas: say the sheet was removed (S13-23). In the map's working rule "Each removal PR updates what it makes false", add to the S13-23 line: "`add-match-rule` lines on the Swift enums and `PromoteEvidenceDisplay` (the guidance-name check fails otherwise)".
Check: `python3 scripts/check-guidance-names.py` passes; `rg -n 'evidence sheet|PromoteEvidence|promote\.sheet|recordedDirect|share phrase' .agents .cursor docs/*.md docs/ideas/*.md` finds only past-tense mentions that name S13-23.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (map row), this plan (status). Skill: —.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-23-promote-stops-asking-for-pins.md` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aSkipRowFilesNoArgumentConfidenceOrPins` → renamed `doneSendsATargetPerFiledRow` | A handle row, a New row, a Skip row → `batchRows()` is `[(a, .handle, "per-1"), (b, .new, nil), (c, .skip, nil)]` | AC-1.3, D-1, D-6 |
| T-2 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `doneSendsDefaultOrChosenPins` → renamed `doneFilesEveryRowWithNoClaimFields` | Done on a FakeStore page → `lastPromoteBatch` rows carry only subject, target and entity; `recordedCalls` holds no `getPromoteClaimEvidence` and no grades call | AC-1.3, AC-1.4 |
| T-3 | Swift | `PromoteFlowTests.swift` `theCaptionReasonIsTheMoveWhenThereIsOne` → renamed `theCaptionIsTheMoveElseTheViaNeighbor` | Proposal row with `moveReason` `.neighborMoved` and `viaNeighborSubjectID` "a" → `caption == .neighborMoved(neighborID:)`; only `viaNeighborSubjectID` "a" → `.via(neighborID: "a")`; neither → nil; `moveReason` `.taken` (not a move code) and no via → nil | AC-2.1 – AC-2.3, D-3, D-4 |
| T-4 | Swift | `PromoteFlowTests.swift` `retargetingClearsContestedAndMoveUntilTheNextProposal` → renamed `retargetingClearsTheCaptionUntilTheNextProposal` | A row with a move caption is retargeted → `caption` nil, `updated` false; the next proposal sets it again | Edges, D-3 |
| T-5 | Swift | `PromoteFlowTests.swift` `matchMenusUseConclusionTitles` | Rows and alternatives built with `header: .person(…)` → menu titles as today | D-7 |
| T-6 | Swift | `PromoteModelTests.swift` `movesAreWordedNamingTheNeighbor`, `aViaReasonNamesTheTermInLowercase`, `reasonsAreWordedWithLabelsAndRefs` (narrowed to the via arms: via with a role, via without, via to a row not on the page) | → "after Gracie moved to PER-Z", "via Gracie → PER-Z (spouse)", "via Gracie → PER-Z", "Reached from a matched neighbor" | AC-2.1, AC-2.2, AC-2.4 |
| T-7 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `queryKeysMatchDesignTable` | The Promote place's keys no longer list `.confidenceGradesList` | AC-1.4 |
| T-8 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `invalidationsComeFromRegistrySpecs` | The expectation list drops `.confidenceGradesList`; `confidenceGradesLoadInOrder` is deleted | AC-1.4 |
| T-9 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` | The remaining rows and subtests pass with the five enums gone | AC-3.3 |
| T-10 | Swift | `PromoteModelTests.swift` `leaveGuardAsksOnlyAfterAManualChange`, `PromoteFlowTests.swift` `leaveAsksOnlyAfterAManualChange` | Unchanged: only a target change arms the guard | AC-1.5 |

Deleted with their subject (STEP-1, STEP-2): `PromoteEvidenceBoardTests.swift`, `PromoteEvidenceDisplayTests.swift`; in `PromoteFlowTests`: `aPersonRecordReadsNaturallyInTheSheetAndSortedInTheMenu` (keep its menu half in `matchMenusUseConclusionTitles` if not already asserted there), `anUnplacedRowDraftedOntoAWeakRecordCaptionsItsReason`, `theAssessmentFollowsTheSelectedRecord`, `aContestedRowAndItsMoveReachTheFlow`, `aContestedRowShowsContestedInItsReadout`, `aContestedRowWithoutAHandleShowsNoReadout`, `aDecidedRowTakesTheLatestContestedFlag`, `anAnchorRowIsNeverContested`, `doneSendsDefaultPinsAndConnectionsFollowEnds` (keep its connections half as `connectionsFollowTheirEnds`), `aNeverOpenedRowSendsDefaultPins`, `openingDraftsTheCheckedPins`, `reopeningKeepsOnlyOfferedPins`, `aLineTogglesEveryMember`, `evidenceForAnOldTargetIsIgnored`, `evidenceNeighborsAreTheRowsDoneFiles`, `aRetargetClearsEvidenceAndPins`, `everyRowStartsWithAnEmptyArgument`, `aTypedArgumentSurvivesARetargetAndAReproposal`, `aManualPinOnTheSameTargetSurvivesAReproposal`, `anUnsettledRowKeepsItsEvidenceOnlyOnTheSameTarget`, `holdingAFoundRecordResetsTheOldTargetsEvidence`; in `PromoteModelTests`: `openingASheetLoadsItsEvidence`, `aSheetOpenedAfterANameOrderChangeWordsNamesInTheNewOrder`, `aFailedEvidenceReadShowsAnError`, `researcherPropertyLabel`, `reloadEvidenceRetries`, `aLateAnswerForAnotherSheetIsDropped`, `aChosenRecordBelowWeakReadsLittleEvidence`, `anUnplacedRowIsWordedNoRecordSuggested`, `theContestedBandIsSpoken`, `theSheetMarksSpeakTheContestedBand`, `theSheetCarriesTheRowsCaption`; in `ConnectionSentenceRegistryTests`: `aNeighborEventIsNamedByItsType`, `aVowelInitialEventTypeTakesAn`, `aFewSharesReadWhole`, `aNeighborPersonIsNamedByRoleAndName`, `aRelationshipIsTheTermTheNeighborCarries`, `aResearchersTermReadsAsAConnectionNamingIt`, `aPlaceRelationshipIsParentWithinSuccessorOrPredecessor`, `aLocationNamesThePlace`, `nothingNamedIsAConnection`; `CatalogQueryRegistryTests.confidenceGradesLoadInOrder`. Narrowed (D-13): `aFoundRecordSurvivesAReproposalWithItsBandRefreshed` (keeps "the found record stays held", loses the band), `reasonCodesParseThroughTheWireVocabulary` (keeps `PromoteReasonCode` parsing; drops `PromoteFlow.Reason` cases that are gone), `SubjectStoreTests.promoteJoinsExistingHandleWithGradeAndArgument` (loses its last two lines), the draft tests (`aMediumMatchStartsOnItsHandle`, `aWeakMatchStartsOnItsHandle`, `aRowWhoseRecordWasTakenStartsUnset`, `aWeakCandidateBelowTheBarStartsSelected`, `anUnmatchedRowStartsUnselected`, `anUnplacedRowNeverDraftsNewOrSkip`) keep their target assertions and drop band or reason assertions. `aPinnedObservationDeleteNamesThePerson` stays (S13-25).

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteFlowTests`, `PromoteModelTests`, `EvidenceGraphModelTests` (builds proposals through `FakeStore`), `CatalogQueryRegistryTests`, `TestSwiftVocabularyMatchesGo` ("every seeded term has wording" reads `Wording` entries whose arguments STEP-5 edits).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (STEP-5's command prints [])
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief; nothing new drawn, D-8)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: promote-graph-alignment.md (status, principle 5, §3.1), interpretation-layer-data-model.md line 693, matching-model.md lines 5 and 122 (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source whose Evidence graph has unpromoted people, one of whom has an earlier filed record that matches (the product owner's test census). Open Promote on one person, then Map the rest. See: no row has a band, Contested badge, reason or chevron; the headings read Source and Target; the target menus, connection subrows and notes sit where they did before (AC-1.1, AC-1.2).
2. Click where a row's band used to be. Nothing opens (AC-1.1).
3. Change one person's target to the matching record. A spouse row that moves shows Updated and "Updated · via <name> → PER-… (spouse)" or "Updated · after …" (AC-2.1, AC-2.2); a row that moves for no named reason shows only Updated (AC-2.3).
4. Press Done. The toast says how many filed. Run `sqlite3 <project>/provenencia.sqlite "SELECT count(*) FROM identity_claim_evidence;"` before and after: the count does not change (bar 9, app half).
5. Reopen the same graph's Promote page: the filed rows show under Already filed.
6. On the branch, run AC-3.1's and AC-3.2's `rg` commands from the repo root. AC-3.1's finds nothing; AC-3.2's finds nothing (AC-3.1, AC-3.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A kept view reads a deleted `L10n.Promote` member and the build breaks late | STEP-1 and STEP-2 each build and test before the next step |
| `EvidenceGraphModelTests` or another suite builds `CatalogPromoteGraphAlignmentRow` with `score:` or `person:` | STEP-4's check runs the whole Xcode suite; fix the initializer calls in that step |
| `Wording` argument removal leaves a trailing comma or an empty entry | `TestSwiftVocabularyMatchesGo` "every seeded term has wording" and the Xcode build catch it in STEP-5 |
| S14-15 lands first and the `project.pbxproj` entries are gone | Stop condition below |

**Stop and report if:**

- `grep -c fileSystemSynchronized macos/Provenencia.xcodeproj/project.pbxproj` is not 0 (S14-15 landed): skip the `project.pbxproj` edits and report it.
- A file outside §4 names any type or member §3 deletes (`rg` it before deleting); report the path.
- `api/ffi/handlers/promote_align.go` sets `ViaNeighborSubjectId` other than under `r.Via != nil`, or `graphalign` sets `Row.Via` with a reason other than `via` (D-3's assumption).
- Go Apply rejects a row with no `pin_ids` and `default_pins` false (`promote.invalid` or any error from `TestApplyPromoteGraphAlignment…` in `api/ffi/handlers/promote_batch_test.go` when run against a request with only subject, target and entity).
- Removing the assessment cell moves the target column or the anchor row's text in the running app (D-8 failed).

## 12. Open questions

None.

## 13. PR

**Title:** S13-23 — Promote stops asking for pins
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-23-promote-stops-asking-for-pins.md`.
