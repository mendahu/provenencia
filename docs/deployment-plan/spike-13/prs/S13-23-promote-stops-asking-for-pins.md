# S13-23 — Promote stops asking for pins

**Status:** Landed in #633 (2026-10-29)
**Spike / slice:** Spike 13 · Slice 0 — Removing Observation pins
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D11](../design/S13-D11-promote-without-assessment.md), board agreed 2026-10-29 (§8 Board notes there)
**Board:** [Claude Design — Promote flow](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share), the S13-D11 group: frames **27a** (rows, option (a)), **27W** (Wide), **28** (captions), **29** (busy rows) gate this PR; 30, 30b, 30W and 31 are S13-30's. Every frame on the board below them (S11-D3, S11-D1, S9-D16) is redrawn without the column and in the promote wording. Open the link and read each frame's note; §6 macOS repeats what they decide
**Depends on:** S13-D11 (board agreed 2026-10-29)
**Blocks:** S13-24, S14-02, S14-22
**Branch:** `s13-23-promote-stops-asking-for-pins`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0 until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (removing a `CatalogQueryKey` and a place warm)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** [conclusion §5.3](../../../conclusion-layer-data-model.md#53-promote), [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3, §3.1, §8; working rules: "Pins leave in three PRs, the client first", "Promote's removal has a brief", "Promote files claims with a status, Accepted by default or Provisional", "Identity Claims carry no confidence and no argument", "The score ranks, unseen", "Spike 14's Promote PRs build on the removal", "Each removal PR updates what it makes false"
**Planned:** 2026-10-29 against `main` @ `51cf4aef` (refreshed from `577c9df1` against the agreed board)
**Review:** [`S13-23-promote-stops-asking-for-pins.review.md`](S13-23-promote-stops-asking-for-pins.review.md) — Fits with changes / Ready, 2 rounds; every blocker Fixed

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

Promote becomes a quick filing action in the Mac app. A researcher picks a target per row and presses Done: no row opens an evidence sheet, no row shows a strong, medium or weak band, a Contested mark or a reason, and Done sends no pins, confidence or argument. This is the client half of the pin removal decided in [conclusion §5](../../../conclusion-layer-data-model.md#5-membership-evidence) (docs PR #623) and the first of the three slice 0 PRs (working rule "Pins leave in three PRs, the client first"). The page is laid out again without the column as board S13-D11 draws it. After this PR, no Swift outside the generated proto file names a Method or field that S13-24 removes, apart from the two proposal fields the draft rule still reads (D-4).

**Why now.** Slice 0 runs first in Spike 13. S13-24 (the Go engine) and Spike 14's S14-02 and S14-22 wait on this PR, because it deletes the Promote files and Swift mappings they would otherwise have to edit.

**What stays.** The proposal and its ranking (Go is untouched apart from one Go test file); the target menu, Find another record…, New and Skip; the connection subrows; the conflict and duplicate notes; the Updated and Decided badges and the Updated caption for a moved or via-reached row; the leave guard; Done's one transaction and the stale re-propose; every filed row is Accepted, as today (the row's status control is S13-30's, D-18); `promoteSubject` (the graph card's single-Subject Promote, which S14-16 deletes), without its grade and argument parameters (D-17).

## 2. User stories

**US-1.** As a researcher, when I promote a Source, I pick a target per row and press Done, and nothing asks me to judge evidence, pin observations, grade confidence or write an argument.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Promote page with a row on a suggested record (strong) and a row drafted onto a weak record | Neither row shows a band, a Contested badge, a reason or a chevron; nothing on the row opens a sheet (S13-D11 PA-1, PA-2) |
| AC-1.2 | The page's column headings and Already promoted rows | The headings read **On this Source** and **Promotes to**, the Assessment heading is gone, the freed slot stays empty between the target menu and the badges, and an Already promoted row carries "Promoted from this Source earlier" in that slot, as frame 27a draws (PA-2 – PA-4) |
| AC-1.3 | Done with one row on handle `e1`, one New, one Skip | `applyPromoteGraphAlignment` receives `[(a, "e1", .handle), (b, "", .new), (c, "", .skip)] (`subjectID`, `handleID`, `target`)` as `CatalogPromoteGraphAlignmentFixed`; the Go request rows carry subject, target and entity only, so `default_pins` is false, `pin_ids` empty, and `confidence_grade_id` and `argument` empty |
| AC-1.4 | The page opens | No `listClaimConfidenceGrades` call is made, and the Promote place warms no confidence grade key |
| AC-1.5 | A row the researcher has not touched | Opening the page and pressing Done is the whole flow: only a target change arms the leave guard |
| AC-1.7 | The Promote page in a 2560 pt window, sidebar open (Wide), and in the default 1366 pt window (Regular) | At Wide the row list widens to at most 1600 pt and centers, the target menu and its open menu are 460 pt, and the header and footer bars span the pane; at Regular the list is at most 1200 pt and the target menu 330 pt; a resize across 1600 keeps each menu's selection, an open search and the scroll position (frame 27W; S13-D11 PA-W1 as the board revised it) |
| AC-1.8 | Any Promote copy, and the Evidence graph's Unpromote confirm | No user-facing string says file, filed, filing or unfiled for this process (the confirm's list header reads "Also unpromoted"): the placeholder reads "Select where this promotes", Skip "Skip — not promoted yet", the footer "Promote 12 · Skip 3 · 22 connections", the busy Done "Promoting", the toast "Promoted 12 subjects and 22 connections", a waiting subrow's neighbor "…is already promoted", the Unpromote confirm's cancel "Keep promotion" (user, 2026-10-29: promote everywhere) |
| AC-1.9 | The mapped page's intro line, and a duplicate note on two person, event or place rows | The intro reads "Opened from Gracie Gray Gates (Frickleton)" with nothing after it; the duplicate note's body names the kind: "These may be the same person." / "…same event." / "…same place.", then "Combine them on the Evidence graph, or pick another target for one row." (user, 2026-10-29) |
| AC-1.6 | Done files a row, and the graph card promotes a Subject | Both claims come back with status accepted; no Swift type carries a claim confidence or argument, and neither call sends one (D-17, D-18) |

**US-2.** As a researcher, when a proposal moves a suggested row, I still see why it moved.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A re-proposal moves row B because settling moved neighbor Gracie to PER-Z (`moveReason` `.neighborMoved`, `moveNeighborSubjectID` Gracie's row) | Row B shows the Updated badge and "Updated · after Gracie moved to PER-Z" |
| AC-2.2 | A re-proposal moves row B onto a record reached from row A (`viaNeighborSubjectID` = A, A on PER-Z, no move reason) | Row B shows "Updated · via Gracie → PER-Z (spouse)" when the bridge between them has the term spouse, "Updated · via Gracie → PER-Z" when it has none |
| AC-2.3 | A re-proposal moves row B with neither a move reason nor a via neighbor (an unplaced row now drafted onto a weak record) | Row B shows the Updated badge and no caption line (PA-6) |
| AC-2.4 | A via neighbor or a move neighbor that is not a row on the page | The caption reads "Updated · Reached from a matched neighbor" (via) or "Updated · after a connected row changed" (move), as today |

**US-3.** As a contributor, I can regenerate the proto after S13-24 without a Swift edit beyond the draft rule.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `rg 'pinIds\|pinIDs\|defaultPins\|getPromoteClaimEvidence\|ClaimEvidence\|confidenceGradesList\|listClaimConfidenceGrades\|ConfidenceGrade\b\|CatalogPromoteBatch\|confidenceGradeID' macos/App macos/ProvenenciaTests -g '!**/Generated/**'` | No hit |
| AC-3.4 | `rg -n 'argument' macos/App/Platform/GoStore.swift macos/App/Platform/GenealogyStore.swift macos/App/Platform/FakeStore.swift` | Only Source credibility lines (`upsertSourceCredibilityAssessment`, `CatalogCredibilityAssessment`, `mapCredibility`); no Identity Claim or `promoteSubject` line, and no comment about a claim's argument |
| AC-3.2 | `rg '\.score\b\|\.contested\b\|reasonProperty\|\.pins\b\|viaBridgeType\|viaRole\|\.written\b' macos/App/Platform/GoStore.swift macos/App/Features/Promote` | No hit; `r.assessment` and `r.reason` are read only where `GoStore.proposePromoteGraphAlignment` maps the row (D-4) |
| AC-3.3 | `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` | Passes, with the evidence enum rows and the share-phrase subtest gone from `TestSwiftVocabularyMatchesGo` |

## 3. Scope

| In | Covered by |
| --- | --- |
| Delete the evidence sheet and everything only it uses: `PromoteEvidenceSheet`, `PromoteEvidenceBoard`, `PromoteEvidenceCards`, `PromoteEvidenceDisplay`, their tests, the sheet state in `PromoteFlow` and `PromoteModel` (S13-D11 PA-1) | STEP-1, AC-1.1, AC-1.5 |
| Delete the assessment column and lay the row, headings and Already promoted rows out as the board draws them: `PromoteReadoutBadges`, the readout button, `PromoteFlow.SelectionReadout`, the row and menu items' band and reason fields (PA-2 – PA-4, PA-8) | STEP-2, AC-1.1, AC-1.2 |
| The Updated caption reads the move reason, else the via neighbor (D-3; user, 2026-10-29; PA-5 – PA-7) | STEP-2, AC-2.1 – AC-2.4 |
| At Wide the row list widens to 1600 pt and centers, and the target column is 460 pt (frame 27W, D-19) | STEP-2, AC-1.7 |
| Promote, never file: every user-facing string that names the promote process says promote (D-20, user 2026-10-29); the intro line drops its second half and the duplicate note names the kind (D-21, user 2026-10-29) | STEP-6, AC-1.8, AC-1.9 |
| Delete the grade picker's read: `PromoteModel.grades` and `confidenceKey`, `PromoteView`'s query, the Promote place's warm, `CatalogQueryKey.confidenceGradesList` and its registry entry, `listClaimConfidenceGrades` in `GenealogyStore`, `GoStore`, `FakeStore` and the `OnboardingModelTests` conformer, `CatalogClaimConfidenceGrade` | STEP-3, AC-1.4, AC-3.1 |
| Delete the claim evidence read and wire mapping: `getPromoteClaimEvidence` in `GenealogyStore`, `GoStore` (with every evidence mapper), `FakeStore` and the conformer; `CoreMethod.getPromoteClaimEvidence` and `CoreMethod.listClaimConfidenceGrades`; the `CatalogClaimEvidence*` types | STEP-3, AC-3.1 |
| Done sends a target per row and nothing else: `CatalogPromoteBatchRow`, `CatalogPromoteBatchWritten` and `CatalogPromoteBatchResult` go; Done sends `[CatalogPromoteGraphAlignmentFixed]` from `PromoteFlow.fixed(for:)` and the store call returns nothing (D-1, D-2, D-6) | STEP-4, AC-1.3 |
| `FakeStore.applyPromoteGraphAlignment` files what it is sent and refuses nothing but a stale revision (D-14) | STEP-4, AC-1.3 |
| The app stops carrying claim confidence and argument: `CatalogIdentityClaim.confidenceGradeID` and `.argument` go; `promoteSubject` loses its `confidenceGradeID` and `argument` parameters in `GenealogyStore`, `GoStore`, `FakeStore` and `ThrowingStore`, and `GoStore` stops sending and mapping them (D-17) | STEP-4, AC-1.6, AC-3.1, AC-3.4 |
| Swift's proposal row and alternative lose every field nothing reads after this PR (D-5), and carry one `header: CatalogConclusionHeader?` instead of three optionals (D-7, Spike 11 leftover) | STEP-4, AC-3.2 |
| Delete the share wording: `ConnectionSentenceRegistry`'s `KindEntry.share`, `ShareInput`, `share(_:similar:)`, `seededShare`, `namedShare`, `termName`; `CatalogEvidenceTerm`; `L10n.EvidenceGraph.share*`; `L10n.PropertyTerm.Wording` and `wording(propertyKey:termKey:)`, whose term labels become plain resources behind `resource(propertyKey:termKey:)` (D-15) | STEP-5, AC-3.1 |
| Delete the wire enums `EvidenceOutcome`, `EvidencePropertyCause`, `KinshipPattern`, `ValueRelation`, `ValueCause` and the sheet's `L10n.Promote` members | STEP-1, STEP-5, AC-3.3 |
| `api/ffi/swift_vocabulary_test.go`: the five evidence enum rows, their names in the "tolerate a value this build doesn't know" list, the "every term a sentence shares has its share phrases" subtest, the helpers and imports only they use; "every seeded term has wording" reads `resource(propertyKey:termKey:)` | STEP-5, AC-3.3 |
| `Localizable.xcstrings` loses every key no Swift names after this PR in `promote.*` and `evidenceGraph.share*`, the stale `promote.toast.filed` included | STEP-1, STEP-2, STEP-5 |
| `project.pbxproj` loses the entries for the deleted Swift files (S14-15 has not landed) | STEP-1, STEP-2 |
| The docs and skills the removal makes false (working rule "Each removal PR updates what it makes false", S13-23 line), and Promote's row in `docs/ideas/wide-layout.md` | STEP-7 |
| The brief's S13-23 rows marked shipped, the map, completed.md and plan status | STEP-8 |

| Out | Home |
| --- | --- |
| The Go evidence engine, `GetPromoteClaimEvidence` and `ListClaimConfidenceGrades` RPCs, Apply's pin and claim-field inputs, the proposal's display-only fields on the wire | S13-24 |
| `PromoteFlow.draftTarget` and the `assessment` and `reason` fields it reads; `PromoteFlow.Assessment` | S13-24 (what a row starts on becomes Go's) |
| The row's status control (Accepted or Provisional), `status` on the Apply row, provisional claims and their upgrade | S13-30 (after S13-24; S13-D11 PS-1 – PS-6) |
| Reserving `confidence_grade_id` and `argument` on the wire; the auto-reconciler's claim-confidence tiebreak | S13-24 |
| Dropping `identity_claims.confidence_grade_id`, `.argument`, `claim_confidence_grades` and its seed | S13-25 |
| A home for the reasoning that records belong together | not decided (working rule "Identity Claims carry no confidence and no argument") |
| One shared proto header message for the four `oneof header` copies, so Swift has one mapper | S13-24's planner (it regenerates the proto); noted in the review log |
| The `identity_claim_evidence` table, `identityclaims` pins, the pin delete facet, `PVDeleteImpact`'s pin line, `FakeStore.pinsByObservation`, `PromoteModelTests.aPinnedObservationDeleteNamesThePerson`, the pin cases in `PVDeleteImpactTests` | S13-25 |
| `promoteSubject` and its `confidenceGradeID` / `argument` parameters | S14-16 |
| Any change to proposals, the ranking or the band in Go | not planned (working rule "The score ranks, unseen") |
| A claim editor that sets confidence and argument after filing | [identity-claim-review idea](../../../ideas/identity-claim-review.md) |
| The via caption's term comes from the first page bridge between the two rows (`PromoteFlow.bridge(between:and:)`), not from the edge the walk used (`via_bridge_type`, `via_role`, which D-5 stops mapping) | S14-22 (Promote's wire says what Go decided) |
| Synchronized Xcode folders | S14-15 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Sheet | `macos/App/Features/Promote/PromoteEvidenceSheet.swift`, `PromoteEvidenceBoard.swift`, `PromoteEvidenceCards.swift`, `PromoteEvidenceDisplay.swift` | The row's evidence sheet: board of lines and relationship cards, pin checkboxes, status, confidence picker, argument field. Only callers: each other, `PromoteView` (`.pvSheet`), `PromoteModel.board(of:)`, `PromoteFlow.Row.board` |
| Sheet tests | `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift`, `PromoteEvidenceDisplayTests.swift` | Test only the files above |
| Page model | `macos/App/Features/Promote/PromoteModel.swift` `PromoteModel` | Holds `SheetEvidence`, `sheetEvidence`, `evidenceGeneration`, `grades`, `confidenceKey`, `propertyLabels`, `propertyLabelsByKey`; `togglePin`, `setArgument`, `setConfidence`, `openSheet`, `closeSheet`, `reloadEvidence`, `setSheetSubject`, `loadEvidence`, `board(of:)` (its `labels:` closure is the only caller of `propertyLabel(key:)`), `assessmentText`, `propertyLabel(key:origin:)`, `propertyLabel(key:)`; `reasonText(_:on:)` words every `PromoteFlow.Reason` case; `load()` reads `confidenceKey`; `graphSnapshot()` fills the two label maps (lines 450–459) from `propertiesKey` and reads its `types`; `file()` ignores the Apply result |
| Row list | `macos/App/Features/Promote/PromoteFlow.swift` `PromoteFlow` | `Assessment`; `Reason` (12 cases; `init(code:propertyKey:propertyOrigin:neighborID:)`, `init(_ row:)`, `init?(move:)`, `init(_ alt:)`); `SelectionReadout`; `Alternative.readingTitle`, `.assessment`, `.reason`; `Row.assessment`, `.reason`, `.board`, `.pins`, `.confidenceGradeID`, `.argument`, `.contested`, `.move`, `.readingTitle`, `.selectionReadout`, `.captionReason`; `sheetSubjectID`; `evidenceNeighbors(for:)`, `setEvidence`, `togglePin`, `setArgument`, `setConfidence`; `fixed(for:)` maps a row's `Target` to `CatalogPromoteGraphAlignmentFixed` for a proposal; `batchRows()` / `batchRow(_:)` map the same `Target` again to a string target plus grade, argument and pins; `item(id:ref:kind:person:event:place:…)` rebuilds a `CatalogConclusionHeader` from three optionals; `draftTarget` reads `Assessment(wire: proposal.assessment)` and `Reason(proposal) == .taken` |
| Page view | `macos/App/Features/Promote/PromoteView.swift` `PromoteView`, `PromoteAlignmentRow`, `PromoteAnchorRow`, `PromoteLayout` | `.task` queries `confidenceKey`; `.pvSheet(item: sheetBinding)`; `columnHeaders`, `PromoteAlignmentRow` and `PromoteAnchorRow` each spell the same grid (kind tile, flexible name, a literal `14` pt arrow, `targetColumn`, a flexible assessment slot, `badgeColumn`); the row's assessment cell is a button with `PromoteReadoutBadges`, the reason and a chevron, identifier `promote.row.<ref>.evidence`; the anchor row puts "Filed earlier" in the assessment slot; caption `if row.updated` with `row.captionReason` |
| Readout | `macos/App/Features/Promote/PromoteReadout.swift` `PromoteReadoutBadges`, `PromoteUpdatedCaption` | Badges for the row and sheet title; the caption for the row and sheet |
| Store protocol | `macos/App/Platform/GenealogyStore.swift` | `CatalogClaimConfidenceGrade`; `CatalogClaimEvidencePin`, `…Line`, `…Section`, `…Bundle`, `CatalogEvidenceTerm`, `CatalogClaimEvidenceShape`, `…Relationship`, `…Offer`, `CatalogClaimEvidence`; `CatalogPromoteGraphAlignmentFixed` (`subjectID`, `handleID`, `target: PromoteTargetCode`); `CatalogPromoteGraphAlignmentAlternative` (`score`, `person`/`event`/`place`, `assessment`, `reason`, `reasonPropertyKey`, `reasonPropertyOrigin`, `viaNeighborSubjectID`); `CatalogPromoteGraphAlignmentRow` (`score`, `assessment`, `reason`, `reasonPropertyKey`, `reasonPropertyOrigin`, `viaBridgeType`, `viaRole`, `contested`, `person`/`event`/`place`, …); `CatalogPromoteBatchRow`, `CatalogPromoteBatchWritten`, `CatalogPromoteBatchResult`; protocol `getPromoteClaimEvidence`, `applyPromoteGraphAlignment(…) -> CatalogPromoteBatchResult`, `listClaimConfidenceGrades` |
| Go store | `macos/App/Platform/GoStore.swift` `proposePromoteGraphAlignment`, `getPromoteClaimEvidence`, `mapEvidenceTerm`, `mapEvidenceShape`, `mapEvidenceSection`, `mapFixed`, `applyPromoteGraphAlignment`, `listClaimConfidenceGrades`, `mapConclusionHeader`, `mapSearchHeader` | Maps every proposal field; maps the evidence; Apply sends grade, argument, `defaultPins`, `pinIds` and maps `written`, which nothing reads |
| Method ids | `macos/App/Platform/CoreInvoke.swift` `CoreMethod.listClaimConfidenceGrades`, `.getPromoteClaimEvidence` | Used only by the two `GoStore` calls |
| Fake store | `macos/App/Platform/FakeStore.swift` `promoteEvidence`, `promoteEvidenceError`, `lastPromoteBatch`, `proposePromoteGraphAlignment`, `getPromoteClaimEvidence`, `applyPromoteGraphAlignment`, `listClaimConfidenceGrades` | Builds proposal rows with `score`/`assessment`/`reason`; Apply copies Go's refusals (lines 1860–1887: "Go's applyRow", `promote.invalid`, `identityclaims.already_member`) and writes `pins: row.pinIDs.count`; no test expects those refusals (`rg 'already_member\|promote.invalid' macos/ProvenenciaTests` finds nothing) |
| Conformer | `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore` | Implements `getPromoteClaimEvidence`, `applyPromoteGraphAlignment`, `listClaimConfidenceGrades`, `promoteSubject(…confidenceGradeID:argument:)` |
| Claim fields | `macos/App/Platform/GenealogyStore.swift` `CatalogIdentityClaim` (`confidenceGradeID`, `argument`), protocol `promoteSubject(projectDir:userID:subjectID:entityID:confidenceGradeID:argument:)` and the extension's `promoteSubject(projectDir:userID:subjectID:)`; `GoStore.promoteSubject` (sets `req.confidenceGradeID`, `req.argument`, maps `resp.claim.confidenceGradeID`, `.argument`); `FakeStore.promoteSubject` (stores both on the claim); callers passing `confidenceGradeID: nil, argument: ""`: `EvidenceGraphModelTests.swift` (2), `SearchCatalogStoreTests.swift` (1), `SubjectStoreTests.swift` (2, one with "cg-high") | Carry a grade and an argument that Promote never sets; the Go side stops reading them in S13-24 |
| Query key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.confidenceGradesList`, `Kind.confidenceGradesList`; `CatalogQueryRegistry.swift` its spec; `PlaceRegistry.swift` the Promote place's `queryKeys` | Session-fresh grades list warmed on the Promote place |
| Key tests | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `confidenceGradesLoadInOrder`, `invalidationsComeFromRegistrySpecs`; `PlaceRegistryTests.swift` `queryKeysMatchDesignTable`; `SubjectStoreTests.swift` `promoteJoinsExistingHandleWithGradeAndArgument` (last two lines read grades) | Name the key or the store call |
| Share wording | `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift` `KindEntry.share`, `ShareInput`, `termName`, `seededShare`, `namedShare`, `share(_:similar:)`, every entry's `share:` closure | Only caller: `PromoteEvidenceDisplay` |
| Share tests | `macos/ProvenenciaTests/ConnectionSentenceRegistryTests.swift` from `// MARK: - What a row shares with a connected record` (line 239: helpers `term`, `share`; tests `aNeighborEventIsNamedByItsType` … `nothingNamedIsAConnection`) | Test only the share wording |
| Wire enums | `macos/App/Platform/ConclusionVocabulary.swift` `EvidenceOutcome`, `EvidencePropertyCause`, `KinshipPattern`, `ValueRelation`, `ValueCause` | Read only by the evidence types, `L10n.Promote.outcome`, `lineSentence` and the sheet |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Promote` (lines 1316–1792), `L10n.EvidenceGraph` share section (from line 2417), `L10n.PropertyTerm` (line 2635): `struct Wording` (label plus 77 share / `recordedDirect` arguments across 39 entries), `wording(propertyKey:termKey:)` (the switch), `resource(propertyKey:termKey:)` (returns `wording(…)?.label`) | Sheet, column and share copy beside page copy that stays; once the share fields go, `Wording` holds one field |
| Term label tests | `macos/ProvenenciaTests/EventTitleDisplayTests.swift`, `PropertyTermDisplayTests.swift` | Read `L10n.PropertyTerm.<entry>.label` |
| Catalog | `macos/App/Resources/Localizable.xcstrings` | 172 `evidenceGraph.share*`, `promote.sheet.*`, `promote.assessment.*` keys, plus the column, readout and reason keys; `promote.toast.filed` is already named by no Swift |
| Parity test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` rows "evidence outcomes", "evidence property causes", "kinship patterns", "value relations", "value causes"; the tolerate list (line 530); subtests "every seeded term has wording" (reads `static func wording(propertyKey`) and "every term a sentence shares has its share phrases"; helpers `patternCodes`, `sortedRelations` | Hold the five enums and the share fields to Go |
| Flow tests | `macos/ProvenenciaTests/PromoteFlowTests.swift` (fixture `row(id:target:handleID:assessment:)` builds `score`, `reason: .agrees`, `reasonPropertyKey`, alternatives with `score`), `PromoteModelTests.swift` (fixture `row(_:handleID:assessment:)`, same fields) | Cover drafting, readout, contested, caption, pins, argument, evidence, sheet and the rest of the page (test names in §8) |
| Project | `macos/Provenencia.xcodeproj/project.pbxproj` | 24 `PromoteEvidence*` entries (4 per file: build file, file reference, group child, sources phase) and 4 for `PromoteReadout.swift`; not on synchronized folders (`grep -c fileSystemSynchronized` is 0) |
| Docs | [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) line 7 (last sentence, "Until the code catches up …"), line 35 (principle 5, "…is surfaced, and every conflict the engine found shows as *contested*."), line 62 (§3.1, "Rows whose suggestion changed …"); [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) line 693; [`matching-model.md`](../../../matching-model.md) line 5 ("The registry's `evidence` section goes with the claim evidence sheet."), line 122 ("It is deprecated and goes with the sheet and the pins") | Describe the sheet or the Contested mark as current, or as going |
| Skills | `.agents/skills/add-localized-string/SKILL.md` lines 147, 151; `.agents/skills/add-seeded-vocabulary/SKILL.md` line 80; `.agents/skills/add-match-rule/SKILL.md` lines 76, 91 | Cite `PromoteEvidenceDisplay.Role`, `promote.sheet.subtitle.*`, share phrases, `recordedDirect`, `L10n.PropertyTerm.wording`, the five Swift enums and `PromoteEvidenceDisplay.wording`; `scripts/check-guidance-names.py` fails once those names are gone |
| Ideas | `docs/ideas/term-management.md` line 28, `docs/ideas/international-names.md` line 43, `docs/ideas/binding-order-editing.md` line 7 | Describe the sheet as current; `international-names.md` names `L10n.PropertyTerm.wording` |

**Precedent.** `35145e44` ("Send a decided row's New or Skip target as a typed PromoteTargetCode instead of a raw string") typed `CatalogPromoteGraphAlignmentFixed.target`, which Done now reuses (D-6). `mapConclusionHeader` and `mapSearchHeader` in `GoStore.swift` are the shape for the row and alternative header mappers (D-7). `1d220bb6` ("Remove the per-graph link and unlink calls …") is a Swift-and-store removal of the same kind: protocol, `GoStore`, `FakeStore`, conformer and tests in one change.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Done sends, per filed row, the subject, the target code and the record id. `GoStore.applyPromoteGraphAlignment` maps each `CatalogPromoteGraphAlignmentFixed` to an `ApplyPromoteGraphAlignmentRow` with `subjectID`, `target = fixed.target.wire` and `entityID = fixed.handleID` when it is not empty, and sets nothing else. Go reads unset `pin_ids`, `default_pins` false and an empty grade and argument as an accepted claim with no pins (`promotealign.ResolvePins` skips a row with neither) | Working rule "Promote files accepted claims and nothing else" | Sending `default_pins: false` explicitly — the proto default already is false, and the field goes in S13-24 |
| D-2 | `applyPromoteGraphAlignment` returns nothing (`async throws`). `CatalogPromoteBatchResult` and `CatalogPromoteBatchWritten` go with the `written` mapping in `GoStore` and `FakeStore`. `PromoteModel.file()` already ignores the result | `rg '\.written\b' macos` finds only the `GoStore` line that builds it; S14-16 gives Apply its `WriteEffects` return | Keeping `written` for S14-16 — nothing reads it, and S14-16's row names `WriteEffects`, not `written` |
| D-3 | The Updated caption: `Row.caption: Reason?` *(new field, replaces `move` and `reason`)* is `Reason(move:)`, else `Reason(via:)`, else nil, computed in `makeRow` for every non-anchor row (a kept decided row too, so it is current if the row is later retargeted and re-proposed). The view draws `PromoteUpdatedCaption` only when `row.updated` and `row.caption != nil`. A row moved with neither shows the Updated badge alone (user, 2026-10-29: "Move, else via"; S13-D11 PA-5, PA-6) | Keeps §3.1's "via Gracie → PER-Z" caption without reading the `reason` field S13-24 reserves; `via_neighbor_subject_id` is set only when the walk reached the row from a neighbor (`api/ffi/handlers/promote_align.go` `proposalProto`, under `r.Via != nil`) and S13-24 keeps it | Move reasons only (the map's first wording) — the user chose to keep via; "Updated · Nothing to match on" for the rest — it words a reason the page no longer shows |
| D-4 | `PromoteFlow.Reason` has four cases: `via(neighborID:)`, `neighborPlaced(neighborID:)`, `neighborMoved(neighborID:)`, `neighborReleased(neighborID:)`, and two initializers: `init?(move proposal: CatalogPromoteGraphAlignmentRow)` returns the move case for `moveReason` `.neighborPlaced`, `.neighborMoved` or `.neighborReleased` with a non-empty `moveNeighborSubjectID`, else nil (`.via`, `.taken`, an unrecognized code and an empty neighbor all give nil); `init?(via proposal:)` *(new)* returns `.via(neighborID:)` when `viaNeighborSubjectID` is not empty, else nil. `init(code:propertyKey:propertyOrigin:neighborID:)`, `init(_ row:)` and `init(_ alt:)` go. The Swift proposal row keeps `assessment` and `reason` only because `draftTarget` reads them until S13-24: `draftTarget` tests `proposal.reason == .taken` on `PromoteReasonCode`. `PromoteFlow.Assessment` stays for `draftTarget` alone. `PromoteReasonCode` keeps every case (it mirrors `graphalign.Reason`, which S13-24 shrinks) | Every other case was worded only by the readout; the draft rule leaves in S13-24 with its fields | Removing `assessment` and `reason` now — `draftTarget` is S13-24's (map scope) |
| D-5 | Swift's proposal types lose every field no Swift reads after this PR: `CatalogPromoteGraphAlignmentRow` loses `score`, `reasonPropertyKey`, `reasonPropertyOrigin`, `contested`, `viaBridgeType`, `viaRole`; `CatalogPromoteGraphAlignmentAlternative` loses `score`, `assessment`, `reason`, `reasonPropertyKey`, `reasonPropertyOrigin`, `viaNeighborSubjectID`, and keeps `handleID`, `handleRef`, `header` | S13-24 reserves most of these proto fields and its scope allows only the draft rule as a Swift edit; `viaBridgeType` and `viaRole` have no reader and this PR rewrites their mapping | Keeping the mapping until S13-24 — its regeneration would then break the Swift build |
| D-6 | Done builds on the proposal's target mapping: `PromoteFlow.batchRows()` returns `fixed(for: filedRows)`, and `batchRow(_:)` goes. `CatalogPromoteBatchRow` is deleted; `applyPromoteGraphAlignment(…, rows: [CatalogPromoteGraphAlignmentFixed])` and `FakeStore.lastPromoteBatch: [CatalogPromoteGraphAlignmentFixed]?`. `CatalogPromoteGraphAlignmentFixed`'s doc comment says it is a row's target as the page holds it, for a proposal or a Done. `fixed(for:)`'s comment drops "or a claim evidence read" | One place turns a page `Target` into a wire target (Spike 11 leftover: Done's target becomes typed); `filedRows` already drops unset rows | A typed `CatalogPromoteBatchRow` beside `Fixed` — the same three fields under two names, two switches to keep in step |
| D-7 | `CatalogPromoteGraphAlignmentRow` and `…Alternative` carry `header: CatalogConclusionHeader?` *(new field)* instead of `person`, `event`, `place`. `GoStore` maps each oneof with `mapPromoteRowHeader` *(new)* and `mapPromoteAlternativeHeader` *(new)*, private static functions placed directly after `mapConclusionHeader`, with `mapSearchHeader` moved up beside them so the four sit together. `PromoteFlow.item(id:ref:header:orders:)` takes the header; `hold` passes its header straight through; the `kind` switch in `item` goes (the header's case is its kind) | Spike 11 leftover; one header type through the store. Swift cannot share one function across the generated `OneOf_Header` enums | Keeping three optionals and a builder in the flow — the store split what the flow rebuilds |
| D-8 | The board agreed PA-2 option (a) (frame 27a): the freed slot stays, empty on a working row and holding "Promoted from this Source earlier" on an Already promoted row, so nothing moves when S13-30 puts the status control in it. The three hand-kept grids become one: `PromoteGridRow` *(new)*, a private generic view in `PromoteView.swift` with five slots (kind tile, Source, target, the freed slot, badges) and the arrow between Source and target, filled by `columnHeaders`, `PromoteAlignmentRow` and `PromoteAnchorRow`. Its column widths are frame 27a's grid: 28 pt tile, flexible Source, 14 pt arrow, the target width (D-19), a flexible freed slot, 112 pt badges. `PromoteLayout` gains `arrowColumn: CGFloat = 14` *(new)* replacing the literal `14`s and holds every width the grid row uses. The freed slot carries a comment naming S13-30. The Source column and the freed slot are both flexible and share the width left after the fixed columns equally, as the board's grid draws them (`minmax(0,1fr)` each) and as today's two `maxWidth: .infinity` cells do; at Wide each takes half of the extra. An Already promoted row whose target is not a record draws an empty target slot. Notes, callouts and subrows keep `PromoteLayout.noteIndent` (kind tile plus `columnGap`, 40 pt): the board's 42 pt comes from its 14 pt web gap, and the app keeps its `PVSpacing.space5` gap | Working rule "Promote's removal has a brief"; S13-D11 PA-2 – PA-4 and its Column grid row; this PR rewrites all three grids, and S14-22 edits the rows next | Option (b) — deleted on the board; keeping three grids — the third copy, which S14-22 would edit three times |
| D-9 | `PromoteUpdatedCaption` moves into `PromoteView.swift` as a `private struct`, and `PromoteReadout.swift` is deleted with its preview | Its only caller after the sheet goes is the row | Keeping a file named for a readout that no longer exists |
| D-10 | `PromoteModel` keeps the `propertiesKey` read for the snapshot's `types`; `propertyLabels`, `propertyLabelsByKey`, both `propertyLabel` functions and the label-fill block in `graphSnapshot()` go (STEP-1 removes `propertyLabelsByKey`, `propertyLabel(key:)` and their lines in the block; STEP-2 the rest). `syncNameOrders` stays (load and re-propose title rows); its comment drops the sheet | The labels served the readout's `agrees` reason and the sheet | Dropping the query — `EvidenceGraphSnapshot.build` needs the types |
| D-11 | Every `L10n.Promote` member and catalog key named only by the deleted files or the deleted column goes; every share member of `L10n.EvidenceGraph` goes. A key stays only if Swift in `macos/App` still names it. STEP-5's check command proves it, and `promote.toast.filed` (named by nothing today) goes with them | One sweep, checked by a command, not a hand list of 170 keys | A hand-written key list in this plan — it drifts the moment a key is renamed |
| D-12 | S13-23 edits `add-match-rule` lines 76 and 91 to drop the Swift enums and `PromoteEvidenceDisplay` (the Swift step for a wire code); the rest of `add-match-rule` stays S13-24's. The map's working rule "Each removal PR updates what it makes false" already lists it for S13-23 (planning commit) | `scripts/check-guidance-names.py` (run by `check-pr.sh`) fails once those names are gone | Leaving it to S13-24 — this PR's gate would fail |
| D-13 | Tests: delete a test in the step that removes the last member it names; a test that covers kept behavior and also asserts a deleted field drops only that assertion in that step. The shared fixtures `PromoteFlowTests.row(id:target:handleID:assessment:)` and `PromoteModelTests.row(_:handleID:assessment:)` keep their signatures and stop passing `score`, `reasonPropertyKey`, `reasonPropertyOrigin` (STEP-4) and build alternatives as `CatalogPromoteGraphAlignmentAlternative(handleID:handleRef:)`; their `reason:` becomes `.unplaced` for a placed row's default where `.agrees` was only there for the readout | Keeps the coverage of what stays and each step building | Rewriting the suites |
| D-14 | `FakeStore.applyPromoteGraphAlignment` records the rows, keeps its stale-revision check (the fake's audit clock that `proposePromoteGraphAlignment` reads, which `aStaleDoneReproposes` drives through a real write), and then files every `.handle` and `.new` row: `.skip` and any other code write nothing; a Subject the fake does not hold, a Subject already filed, or a `.handle` whose record the fake does not hold is left as it is. No `promote.invalid` or `identityclaims.already_member` arm, no "Go's applyRow" comment; a row whose Subject type is not a primary kind, or whose `.handle` record is of another type, is left as it is; the audit revision advances once when the Done wrote at least one claim, as today (count the claims written in a local, since `written` goes); claims are written with status accepted (D-17 removes their grade and argument) | `.cursor/rules/swift-tests.mdc`: FakeStore does not re-implement refusals; test engine rules in Go | Keeping the copied refusals — they would drift from Go's Apply when S13-24 and S14-16 change it |
| D-15 | `L10n.PropertyTerm.Wording` and `wording(propertyKey:termKey:)` go. Each of the 39 entries becomes `static let <name> = LocalizedStringResource("propertyTerm.…", defaultValue: …, comment: …)` with today's label key, value and comment; `resource(propertyKey:termKey:)` holds the switch (`case (SeededPropertyKey.x, SeededYTerm.z): return <name>`). Tests reading `L10n.PropertyTerm.<name>.label` read `L10n.PropertyTerm.<name>` (`EventTitleDisplayTests.swift` lines 59, 86, 103; `PropertyTermDisplayTests.swift` lines 10, 21; the other `.label` uses there are not term wording and stay). The switch keeps `default: return nil`: `PropertyTermDisplay` and the Go subtest rely on it. The Go subtest "every seeded term has wording" reads `static func resource(propertyKey` with the same `wordingCase` regex | After the share fields go, `Wording` wraps one resource behind two lookups | Keeping `Wording` with one field — a wrapper and a second lookup for one value |
| D-17 | The app carries no claim confidence or argument: `CatalogIdentityClaim` is `id`, `subjectID`, `entityID`, `status`; the protocol's `promoteSubject` is `promoteSubject(projectDir:userID:subjectID:entityID:)` (the extension's three-argument form calls it with `entityID: nil`); `GoStore` sends neither field and maps neither back; `FakeStore` writes neither. `SubjectStoreTests.promoteJoinsExistingHandleWithGradeAndArgument` becomes `promoteJoinsExistingHandle` and asserts the joined entity and status only | Working rule "Identity Claims carry no confidence and no argument"; S13-24 reserves both fields on `IdentityClaim` and `PromoteSubjectRequest`, and this keeps its regeneration free of a Swift edit | Leaving them for S14-16, which deletes `promoteSubject` — S13-24 lands first, and its regeneration would break the Swift build |
| D-18 | This PR adds no status control and sends no status: the Apply row has no `status` field until S13-30, and Go writes accepted. The brief draws the control (PS-1 – PS-6) for S13-30 | User, 2026-10-29: the status needs the engine first, and S13-23 stays the removal PR | Adding the control now — it could only send Accepted |
| D-19 | Wide, as frame 27W draws it: `PromoteView` reads `@Environment(\.pvWidthClass)`. At `.regular` the row list's maximum width is 1200 pt and the target column 330 pt (today); at `.wide` they are 1600 pt and 460 pt, and `PVSelect`'s `menuWidth` follows the target column. The list centers at both widths (its outer frame's alignment becomes `.center`; at Regular the pane is narrower than 1200 pt, so nothing moves). `PromoteLayout` holds both pairs as named constants *(new)*: `listMaxRegular` 1200, `listMaxWide` 1600, `targetColumnRegular` 330, `targetColumnWide` 460, and one function from `PVWidthClass` to each. The header and footer bars span the pane. The record search in the target cell takes the target column's width. One owner: `PromoteGridRow` reads `pvWidthClass` once and hands the resolved target width to its target slot, and the row's `PVSelect` `menuWidth` and the record search read that value, so no other view reads the width class. `PVPageColumns` is not adopted: the page has one column | Board frame 27W ("a departure from PA-W1, which kept 1200"), agreed; `macos-client-patterns.md` § 9 rules 2, 3 and 5 (a table fills the width; state survives the resize: nothing here is keyed by width) | Centering at 1200 — the board widened it; stretching the target menu without a maximum — rule 2 |
| D-20 | Promote, never file (user, 2026-10-29: "Promote is the term we are using to describe this process, so it should be changed to Promote everywhere"; and, the same day, rename the code names too): STEP-6 changes every user-facing value that names the process (§6 Strings, first table), renames the catalog keys and `L10n` members that say file (§6 Strings, second table), renames the Promote feature's identifiers that say file (§6 Strings, third table), and rewords comments that call the process filing in every file this PR edits. File, files and filing for a stored file (an Artifact's file, the ingest errors, `deleteImpact.kind.file`) stay | The board's rule "User-facing copy says promote, never file"; `add-localized-string`'s semantic keys; one word for one process in copy and code | Values only — the code and the copy would disagree (user chose the rename) |
| D-21 | `L10n.Promote.openedFrom(name:)`'s value becomes "Opened from %@" (the board's and today's second half are both untrue: a weak row starts drafted onto its record). `L10n.Promote.duplicateBody` becomes `duplicateBody(_ kind: EvidencePrimaryKind)` *(new signature)* returning one of three whole sentences, keys `promote.warning.duplicateBody.person`, `.event`, `.place` *(new)*: "These may be the same person. Combine them on the Evidence graph, or pick another target for one row." and the same with event and place; `promote.warning.duplicateBody` goes. `PromoteAlignmentRow` passes `row.kind` | User, 2026-10-29 (both); `add-localized-string` "Whole phrases" (a key per sentence, not a kind argument) | One format with a kind word — the whole-phrase rule forbids it |
| D-16 | The PR is implemented and checked on macOS: `scripts/check-pr.sh` runs `xcodebuild test`, which exists only there | Every Swift step's check is the Xcode suite | Implementing on Linux — STEP-1 to STEP-5 would have no check |

**Rules this plan bends.** None. The working rule that said "No brief for the removal" was replaced by "Promote's removal has a brief" (user, 2026-10-29), and the brief gate in `plan-pr` and `plan-spike` now names removals (same commit as this plan's refresh).

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PromoteFlow.Row.caption` *(new)* | field | `PromoteFlow.Row.move`, `captionReason` | Replaces both: one stored value the view reads |
| `PromoteFlow.Reason.init?(via:)` *(new)* | func | `PromoteFlow.Reason.init?(move:)` | Same shape for the other caption source |
| `CatalogPromoteGraphAlignmentRow.header`, `CatalogPromoteGraphAlignmentAlternative.header` *(new)* | field | `CatalogSearchHit`'s header, `CatalogConclusionDetail.header` (`GenealogyStore.swift` lines 242, 820) | Reuses `CatalogConclusionHeader`; replaces three optionals |
| `GoStore.mapPromoteRowHeader`, `GoStore.mapPromoteAlternativeHeader` *(new)* | func | `GoStore.mapConclusionHeader`, `GoStore.mapSearchHeader` | Each proto message has its own `OneOf_Header` type; the four mappers sit together (D-7); one shared proto message is S13-24's call (Out) |
| `PromoteLayout.arrowColumn` *(new)* | constant | `PromoteLayout.kindTile`, `targetColumn`, `badgeColumn` | Names the literal `14` the three grids repeat |
| `PromoteGridRow` *(new)* | view | The grids in `PromoteView.columnHeaders`, `PromoteAlignmentRow.body`, `PromoteAnchorRow.body`; `PVTable` (`macos/App/DesignSystem/Components/Table/PVTable.swift`) | Replaces the three copies; `PVTable` is a single-select browse table with no inline controls, so it does not fit a row of menus |

Searched for: an existing caption source type (none beyond `Reason`), a typed Done target (`CatalogPromoteGraphAlignmentFixed` and `PromoteTargetCode`, reused), a header enum (`CatalogConclusionHeader`, reused), a single term-label lookup (`L10n.PropertyTerm.resource`, kept as the one lookup).

**Builds on.** `PromoteFlow.fixed(for:)` and `PromoteTargetCode.wire` (`35145e44`); `CatalogConclusionHeader`, `mapConclusionHeader`, `mapSearchHeader`; `PromoteLayout`; the place registry's per-place `queryKeys` (`add-workspace-place`).
**Deletes.** Files: `PromoteEvidenceSheet.swift`, `PromoteEvidenceBoard.swift`, `PromoteEvidenceCards.swift`, `PromoteEvidenceDisplay.swift`, `PromoteReadout.swift`, `PromoteEvidenceBoardTests.swift`, `PromoteEvidenceDisplayTests.swift` and their `project.pbxproj` entries. Types: `PromoteModel.SheetEvidence`, `PromoteFlow.SelectionReadout`, `EvidenceBoard` (in `PromoteEvidenceBoard.swift`), `PromoteReadoutBadges`, `CatalogClaimConfidenceGrade`, every `CatalogClaimEvidence*` type, `CatalogEvidenceTerm`, `CatalogPromoteBatchRow`, `CatalogPromoteBatchWritten`, `CatalogPromoteBatchResult`, `ConnectionSentenceRegistry.ShareInput`, `L10n.PropertyTerm.Wording`, the five wire enums. Mechanisms: the sheet's evidence read per row; the grade warm; the share wording table; default pins on Done; Done's second target mapping (`batchRow`); the fake's copied Apply refusals; the two-step term label lookup (`wording` then `.label`). Callees whose last caller goes: `PromoteFlow.evidenceNeighbors(for:)`, `PromoteFlow.Row.readingTitle`, `PromoteFlow.Alternative.readingTitle`, `GoStore.mapEvidenceTerm`, `mapEvidenceShape`, `mapEvidenceSection`, `ConnectionSentenceRegistry.termName`, `seededShare`, `namedShare`, `L10n.EvidenceGraph.share` (private helper), `L10n.PropertyTerm.wording`, `patternCodes`, `sortedRelations` and the Go imports only they used. Kept because the via wording reads them: `PromoteFlow.Row.shownRef`, `PromoteFlow.bridge(between:and:)`.
**Next adds.** S13-24 deletes `draftTarget`, `PromoteFlow.Assessment` and the row's `assessment` and `reason` with one Swift edit, and can give the four header oneofs one proto message. S14-22 changes the Promote rows in one grid view (`PromoteGridRow`) instead of three. S14-16 rewrites Apply on one target mapping and no result type. S14-02 is re-planned without renaming the deleted Promote files.

### Reach

None — no new path into existing code. The page loses a sheet and a column; the caption shows for fewer rows, never for a new state.

### Scale

At the map's Scale (a Source of 100 new people on one Promote page): the page made one `getPromoteClaimEvidence` call per opened sheet and warmed one grades read; after this PR it makes none. Done sends one Apply with 100 rows of three fields. No new loop; `makeRow` and `merge` do less work per row. No budget changes.

### Edges

- **Assumptions.** The caption assumes `viaNeighborSubjectID` is set only with `reason` `via` (`proposalProto` sets it under `r.Via != nil`; `core/graphalign/rows.go` sets `row.Via` only with `ReasonVia`). T-3 holds it with a row that has both and one that has neither.
- **Neighbors.** A row with a move and a via neighbor shows the move (AC-2.1, T-3); one with only a via neighbor shows via (AC-2.2, T-3); one with neither shows no caption (AC-2.3, T-3); a move code with an empty neighbor id shows no caption (T-3); a decided row shows no caption whatever it carries (`updated` is false; T-4).
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

From the S13-D11 inventory (binding):

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Promote row | `PromoteAlignmentRow` (snowflake, Extend) | `macos/App/Features/Promote/PromoteView.swift` | Loses the assessment cell; laid out per PA-2 and D-8 |
| Column headings | `PromoteView.columnHeaders` (snowflake, Extend) | same | On this Source and Promotes to (frame 27a; D-20) |
| Already promoted row | `PromoteAnchorRow` (snowflake, Extend) | same | Re-aligned per PA-4 |
| Column grid | `PromoteGridRow` *(new)* and `PromoteLayout` (snowflake, Extend) | same | D-8 |
| Band and Contested marks, reason line, chevron | `PromoteReadoutBadges`, the evidence button (Remove) | `PromoteReadout.swift`, `PromoteView.swift` | `PVBadge` stays in the kit |
| Updated caption | `PromoteUpdatedCaption` (Ship, moved private) | `PromoteView.swift` | D-3, D-9 |
| Updated, Decided badges; target menu; record search; notes; subrows | `PVBadge`, `PVSelect`, `PVComboBox`, `PVCallout`, `PromoteConnectionSubrows` (Ship) | as today | Unchanged |
| Evidence sheet | Remove | the four `PromoteEvidence*` files | PA-1 |

States: a row with nothing to say shows the badge only. Accessibility: the evidence button and its label (`L10n.Promote.evidenceFor`) go; the target menu's spoken label (`L10n.Promote.filesOn`) and the connection labels stay (PA-8). Identifier `promote.row.<ref>.evidence` goes; no UI test names it (`rg 'row\..*\.evidence' macos` finds only `PromoteView.swift`).

### Strings

Removed: every `promote.sheet.*` key; `promote.assessment.*`; `promote.assessment.chosenNone`; `promote.row.contested`, `promote.row.assessmentContested`, `promote.row.evidenceFor`; `promote.page.column.assessment`; the reason keys off the caption path (`promote.reason.filed`, `.agrees`, `.decided`, `.weak`, `.taken`, `.unplaced`, `.empty`); every `evidenceGraph.share*` key; `promote.toast.filed`. Kept on the caption path: `promote.row.updatedDetail`, `promote.reason.via`, `.viaNeighbor`, `.viaUnknown`, `.neighborPlaced`, `.neighborMoved`, `.neighborReleased`, `.moveUnknown`. The `propertyTerm.*` label keys stay. Added: `promote.warning.duplicateBody.person`, `.event`, `.place` (D-21); `promote.warning.duplicateBody` goes.

Changed values (D-20, D-21; keys stay), in `L10n.Promote` unless named:

| Key | Today | After |
| --- | --- | --- |
| `promote.page.column.target` | Files on | Promotes to |
| `promote.target.placeholder` | Select where this files | Select where this promotes |
| `promote.target.filesOn` | Files %@ on | Promotes %@ to |
| `promote.target.skip` | Skip — not filed yet | Skip — not promoted yet |
| `promote.row.alreadyFiledHeading` | Already filed | Already promoted |
| `promote.row.filedEarlier` | Filed from this Source earlier | Promoted from this Source earlier (comment: "On a read-only already-promoted row, in the slot after the record it was promoted to") |
| `promote.page.rowSummary` | %1$lld rows · %2$lld already filed | %1$lld rows · %2$lld already promoted |
| `promote.section.countFiled` | %1$lld · %2$lld already filed | %1$lld · %2$lld already promoted |
| `promote.page.openedFrom` | Opened from %@ · weak and unmatched rows start as Skip | Opened from %@ |
| `promote.connection.files` | Files on Done | Promotes on Done |
| `promote.connection.spoken.files` | Files on Done. | Promotes on Done. |
| `promote.connection.selfLink` | Not filed · both ends are %@ | Not promoted · both ends are %@ |
| `promote.connection.spoken.selfLink` | Not filed, both ends are %@. | Not promoted, both ends are %@. |
| `promote.connection.otherEnd.filed` | %@ is already filed | %@ is already promoted |
| `promote.done.file` | File %lld | Promote %lld |
| `promote.done.filing` | Filing | Promoting |
| `promote.stale.title` | Nothing was filed — the catalog changed while this page was open | Nothing was promoted — the catalog changed while this page was open |
| `promote.toast.filedTitle` | Filed %1$lld subjects and %2$lld connections | Promoted %1$lld subjects and %2$lld connections |
| `promote.toast.filedBody` | %lld subjects are still on the graph, not filed yet | %lld subjects are still on the graph, not promoted yet |
| `promote.leave.title` | Leave Promote without filing %@? | Leave Promote without promoting %@? |
| `promote.leave.pageMessage` | You chose targets on %lld rows. Nothing has been filed; opening Promote again proposes the graph afresh. | You chose targets on %lld rows. Nothing has been promoted; opening Promote again proposes the graph afresh. (each plural variant) |
| `promote.warning.duplicateTitle` | %1$@ is also filed on %2$@ | %1$@ is also promoted to %2$@ |
| `promote.warning.duplicateNewTitle` | %@ is also filed New and looks the same | %@ is also promoted New and looks the same |
| `evidenceGraph.subject.unpromoteConfirmMessage` (`L10n.EvidenceGraph`) | This filing is removed. The cards stay on the graph, and the handle stays. | This promotion is removed. The cards stay on the graph, and the handle stays. |
| `evidenceGraph.subject.unpromoteKeep` | Keep filing | Keep promotion |
| `evidenceGraph.subject.unpromoteConfirmBridgesNote` | The other end of each stays filed. | The other end of each stays promoted. |
| `evidenceGraph.subject.unpromoteConfirmBridgesHeader` | Also unfiled | Also unpromoted (user, 2026-10-29) |

Renamed keys and `L10n` members (D-20; values as in the table above). A member's parameter that says file is renamed with it:

| Today | After |
| --- | --- |
| `promote.done.file`, `L10n.Promote.doneSummary(file:skip:unselected:connections:)` | `promote.done.promote`, `doneSummary(promote:skip:unselected:connections:)` |
| `promote.done.filing`, `L10n.Promote.filing` | `promote.done.promoting`, `L10n.Promote.promoting` |
| `promote.toast.filedTitle`, `filedTitle(filed:connections:)` | `promote.toast.promotedTitle`, `promotedTitle(promoted:connections:)` |
| `promote.toast.filedBody`, `filedBody(skipped:)` | `promote.toast.promotedBody`, `promotedBody(skipped:)` |
| `promote.page.rowSummary`, `rowSummary(rows:filed:)` | key stays, `rowSummary(rows:promoted:)` |
| `promote.section.countFiled`, `sectionCountFiled(rows:filed:)` | `promote.section.countPromoted`, `sectionCountPromoted(rows:promoted:)` |
| `promote.row.alreadyFiledHeading`, `alreadyFiledHeading` | `promote.row.alreadyPromotedHeading`, `alreadyPromotedHeading` |
| `promote.row.filedEarlier`, `filedEarlier` | `promote.row.promotedEarlier`, `promotedEarlier` |
| `promote.target.filesOn`, `filesOn(name:)` | `promote.target.promotesTo`, `promotesTo(name:)` |
| `promote.connection.files`, `promote.connection.spoken.files` | `promote.connection.promotes`, `promote.connection.spoken.promotes` |
| `promote.connection.otherEnd.filed`, `connectionOtherEnd(name:filed:)` | `promote.connection.otherEnd.promoted`, `connectionOtherEnd(name:promoted:)` |

`promote.reason.filed` / `alreadyFiled` and `promote.toast.filed` are deleted (STEP-2, STEP-5), not renamed.

Renamed Promote feature identifiers (D-20), with every call site and test:

| Today | After |
| --- | --- |
| `PromoteFlow.fileCount` | `promoteCount` |
| `PromoteFlow.filedRows` (private) | `promotedRows` |
| `PromoteFlow.ConnectionState.files`, `PromoteFlow.Bridge.files` | `.promotes`, `promotes` |
| `PromoteFlow.BridgeFact.alreadyFiled` | `alreadyPromoted` |
| `PromoteFlow.OtherEnd.filed` | `promoted` |
| `PromoteModel.file()` (private) | `promote()` |
| local `filed` in `PromoteView.kindSection`, local `files` in `PromoteConnectionSubrows.subrow` | `promoted`, `promotes` |

Each value changes in `L10n.swift`'s `defaultValue` and in `Localizable.xcstrings`' English value, with comments that said "file" reworded the same way. Promote's other strings already say promote or name no process.

Old-scope copy: none left. The Done summary, toast and leave confirm do not mention evidence, pins or confidence (STEP-2 check).

## 7. Steps

**STEP-1 — Delete the evidence sheet**
Files: delete `macos/App/Features/Promote/PromoteEvidenceSheet.swift`, `PromoteEvidenceBoard.swift`, `PromoteEvidenceCards.swift`, `PromoteEvidenceDisplay.swift`, `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift`, `PromoteEvidenceDisplayTests.swift`; edit `PromoteModel.swift`, `PromoteFlow.swift`, `PromoteView.swift`, `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`, `PromoteFlowTests.swift`, `PromoteModelTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: remove the 24 `PromoteEvidence*` lines from `project.pbxproj` (`grep -n PromoteEvidence` lists them). In `PromoteModel` remove `SheetEvidence`, `sheetEvidence`, `evidenceGeneration`, `togglePin`, `setArgument`, `setConfidence`, `openSheet`, `closeSheet`, `reloadEvidence`, `setSheetSubject`, `loadEvidence`, `board(of:)`, `propertyLabel(key:)`, `propertyLabelsByKey` and its lines in `graphSnapshot()` (D-10). In `PromoteFlow` remove `sheetSubjectID`, `evidenceNeighbors(for:)`, `setEvidence`, `togglePin`, `setArgument`, `setConfidence`, `Row.board`, `Row.pins`, `Row.confidenceGradeID`, `Row.argument`, `Row.readingTitle`, `Alternative.readingTitle`, and their uses in `select`, `makeRow`, `item`. Until STEP-4, `batchRow` passes `confidenceGradeID: nil, argument: ""` and leaves `defaultPins` and `pinIDs` at their defaults. In `PromoteView` remove `.pvSheet` and `sheetBinding`, and replace the assessment cell's `Button { model.openSheet(row.subjectID) } label: { … }` with its label content alone (STEP-2 removes the content). Remove every `L10n.Promote` member whose only callers are in the deleted files, and its catalog key: in lines 1351–1624 that is every member except these, which stay for STEP-2 or for good: `evidenceFor`, `contested`, `assessmentContested`, `assessmentChosenNone`, both `assessment(_:)` (STEP-2 removes them), and `neighborPlaced`, `neighborMoved`, `neighborReleased`, `reasonMoveUnknown` (kept, caption path). Outside that range: `sheetTitleNew`, `evidenceUnavailable`, `statusAccepted`. Before deleting a member, `rg` its name across `macos/App` and keep it if a file that stays calls it; the build then proves the rest. Delete each test tagged STEP-1 in §8.
Check: `rg -n 'PromoteEvidence|EvidenceBoard|sheetSubjectID|openSheet|togglePin|setArgument|setConfidence|promote\.sheet' macos -g '!**/Generated/**'` finds nothing; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-2 — Remove the assessment column, lay out the row from the board, narrow the caption**
Files: `PromoteFlow.swift`, `PromoteModel.swift`, `PromoteView.swift`, delete `PromoteReadout.swift`, `L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`, `PromoteFlowTests.swift`, `PromoteModelTests.swift`. Skill: `add-ui-component`, `add-localized-string`, `add-swift-test`.
Do: per D-3, D-4, D-8, D-9, D-10, D-19. `Reason` becomes the four cases and two initializers of D-4. Remove `SelectionReadout`, `Row.assessment`, `Row.reason`, `Row.contested`, `Row.selectionReadout`, `Row.captionReason`, `Row.move` (replaced by `Row.caption`), `Alternative.assessment`, `Alternative.reason`. `makeRow` sets `caption` per D-3; `select` clears `caption` where it cleared `move` and `contested`. `draftTarget` reads `Assessment(wire: proposal.assessment)` and `proposal.reason == .taken`. In `PromoteModel` remove `assessmentText`, `propertyLabels`, `propertyLabel(key:origin:)` and the rest of the label-fill block; `reasonText` keeps the `.via` and three move arms. In `PromoteView` add `PromoteGridRow` and build `columnHeaders`, `PromoteAlignmentRow` and `PromoteAnchorRow` on it per the board's PA-2 option and `PromoteLayout` (D-8); (`L10n.Promote.filedEarlier`'s value and comment change in STEP-6); the caption reads `if row.updated, let caption = row.caption`. Move `PromoteUpdatedCaption` into `PromoteView.swift` as `private struct`; delete `PromoteReadout.swift` and its 4 `project.pbxproj` lines. Remove `L10n.Promote` `assessmentStrong`, `assessmentMedium`, `assessmentWeak`, `assessmentNone`, `assessmentChosenNone`, `assessment(_:)` (both), `contested`, `assessmentContested`, `evidenceFor`, `columnAssessment`, `alreadyFiled`, `agreesOn`, `reasonDecided`, `reasonWeak`, `reasonTaken`, `reasonUnplaced`, `reasonEmpty` and their keys. Delete or narrow each test tagged STEP-2 in §8.
Check: `rg -n 'SelectionReadout|PromoteReadout|captionReason|assessmentText|selectionReadout|columnAssessment|row\.contested|\.move\b' macos/App/Features/Promote macos/ProvenenciaTests -g '!**/Generated/**'` finds nothing; `rg -n 'PromoteReadout' macos/Provenencia.xcodeproj/project.pbxproj` finds nothing; `rg -n -i '\bpins?\b|evidence|confidence' macos/App/Platform/L10n.swift` shows no line inside `enum Promote`; `rg -n 'frame\(width: 14' macos/App/Features/Promote/PromoteView.swift` finds nothing (icon sizes of 14 stay); Xcode tests pass with T-3, T-4, T-6.

**STEP-3 — Remove the grade read and the evidence read from the store**
Files: `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`, `CatalogQueryKey.swift`, `CatalogQueryRegistry.swift`, `PlaceRegistry.swift`, `PromoteModel.swift`, `PromoteView.swift`, `OnboardingModelTests.swift`, `CatalogQueryRegistryTests.swift`, `PlaceRegistryTests.swift`, `SubjectStoreTests.swift`. Skill: `add-workspace-place`, `add-swift-test`.
Do: remove `grades`, `confidenceKey`, the `load()` read and `PromoteView`'s `QueryHandle<[CatalogClaimConfidenceGrade]>` line; remove `.confidenceGradesList` from `CatalogQueryKey` (case, `Kind`, `kind` and `project` arms), its `CatalogQueryRegistry` spec and the Promote place's `queryKeys`. Remove `getPromoteClaimEvidence` and `listClaimConfidenceGrades` from `GenealogyStore`, `GoStore`, `FakeStore` (with `promoteEvidence`, `promoteEvidenceError`) and `ThrowingStore`; `CoreMethod.getPromoteClaimEvidence`, `.listClaimConfidenceGrades`; `mapEvidenceTerm`, `mapEvidenceShape`, `mapEvidenceSection`; `CatalogClaimConfidenceGrade`, `CatalogClaimEvidencePin`, `…Line`, `…Section`, `…Bundle`, `CatalogClaimEvidenceShape`, `…Relationship`, `…Offer`, `CatalogClaimEvidence`. `CatalogEvidenceTerm` stays until STEP-5, because `ConnectionSentenceRegistry` still names it. Apply the STEP-3 test changes in §8.
Check: `rg -n 'getPromoteClaimEvidence|listClaimConfidenceGrades|confidenceGradesList|CatalogClaimEvidence|ConfidenceGrade\b' macos -g '!**/Generated/**'` finds nothing; Xcode tests pass with T-7, T-8.

**STEP-4 — Done sends targets only; the proposal row carries what the page reads**
Files: `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `PromoteFlow.swift`, `OnboardingModelTests.swift`, `PromoteFlowTests.swift`, `PromoteModelTests.swift`, any test that builds `CatalogPromoteGraphAlignmentRow` or `…Alternative` (`rg -l 'CatalogPromoteGraphAlignment(Row|Alternative)\(' macos/ProvenenciaTests`). Skill: `add-swift-test`; `.cursor/rules/swift-tests.mdc`.
Do: per D-1, D-2, D-5, D-6, D-7, D-13, D-14. Delete `CatalogPromoteBatchRow`, `CatalogPromoteBatchWritten`, `CatalogPromoteBatchResult`; the protocol's `applyPromoteGraphAlignment` takes `rows: [CatalogPromoteGraphAlignmentFixed]` and returns nothing, in `GenealogyStore`, `GoStore`, `FakeStore` and `ThrowingStore`; `PromoteModel.file()` drops `_ =`. `PromoteFlow.batchRows()` returns `fixed(for: filedRows)`; `batchRow(_:)` goes; fix the two comments D-6 names. Rewrite `FakeStore.applyPromoteGraphAlignment` per D-14. Remove the D-5 fields from the two proposal types, `GoStore`'s mapping and `FakeStore`'s proposal rows. Replace `person` / `event` / `place` with `header` (D-7). Update the fixtures per D-13 and tests that read `lastPromoteBatch` (`.entityID` becomes `.handleID`). Per D-17 remove `CatalogIdentityClaim.confidenceGradeID` and `.argument` and `promoteSubject`'s two parameters everywhere §4's Claim fields row lists, and their callers' arguments. Reword two comments: `FakeStore.claimBySubject` (line 100) becomes "The Identity Claim Promote wrote per Subject id."; the extension's three-argument `promoteSubject` (`GenealogyStore.swift` line 1671) becomes "Promote onto a new handle (the graph card's Confirm)."
Check: AC-3.1's, AC-3.2's and AC-3.4's `rg` hold; `rg -n 'person: proposal\.person|alt\.person|altPerson|Go.s applyRow' macos/App` finds nothing; Xcode tests pass with T-1, T-2, T-5.

**STEP-5 — Delete the share wording, the term wording wrapper and the wire enums**
Files: `ConnectionSentenceRegistry.swift`, `ConnectionSentenceRegistryTests.swift`, `L10n.swift`, `Localizable.xcstrings`, `ConclusionVocabulary.swift`, `GenealogyStore.swift`, `EventTitleDisplayTests.swift`, `PropertyTermDisplayTests.swift`, `api/ffi/swift_vocabulary_test.go`. Skill: `add-localized-string`; `.cursor/rules/go-tests.mdc`.
Do: remove `KindEntry.share` and every entry's `share:` closure, `ShareInput`, `termName`, `seededShare`, `namedShare`, `share(_:similar:)`; `CatalogEvidenceTerm`; the `// MARK: What a row shares` section of `L10n.EvidenceGraph`; the three private `recordedDirect*` constants. Per D-15 turn every `Wording` entry into a plain `LocalizedStringResource` constant, move the switch into `resource(propertyKey:termKey:)`, delete `Wording` and `wording`, and drop `.label` in the two test files. Remove the five enums in `ConclusionVocabulary.swift`; the share tests and helpers from `// MARK: - What a row shares with a connected record` to the end of that section. Remove the stale catalog keys (D-11). In `swift_vocabulary_test.go` remove the five rows, their names from the tolerate list (it keeps `ReconciledState`, `ReconcilerReason`, `PromoteReasonCode`), the share subtest, `patternCodes`, `sortedRelations` and the imports `go vet` reports unused; point "every seeded term has wording" at `static func resource(propertyKey`.
Check: this prints `[]`:
```bash
python3 - <<'EOF'
import json, pathlib
cat = json.load(open('macos/App/Resources/Localizable.xcstrings'))['strings']
src = ''.join(p.read_text() for p in pathlib.Path('macos/App').rglob('*.swift'))
print([k for k in cat if k.startswith(('promote.', 'evidenceGraph.share')) and f'"{k}"' not in src])
EOF
```
`rg -n 'EvidenceOutcome|EvidencePropertyCause|KinshipPattern|ValueRelation|ValueCause|ShareInput|recordedDirect|shareSimilar|shareOtherEnd|struct Wording|PropertyTerm\.wording|\.label\)' macos/App macos/ProvenenciaTests -g '!**/Generated/**' | rg 'PropertyTerm|Evidence|Kinship|Value|Share|recorded|Wording'` finds nothing; `CGO_ENABLED=1 go vet -tags fts5 ./api/...` and `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` pass; Xcode tests pass with T-9.

**STEP-6 — Promote, never file**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Promote/PromoteView.swift`, `PromoteFlow.swift`, `PromoteModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` (call sites only if a member name changes there; none do), `macos/ProvenenciaTests/L10nResolutionTests.swift`, `PromoteFlowTests.swift`, `PromoteModelTests.swift`. Skill: `add-localized-string`.
Do: per D-20 and D-21 and the §6 Strings table. Rename the keys, members and identifiers in the second and third tables and every call site and test that names them; reword comments that call the process filing. Update `L10nResolutionTests`' four Evidence graph Unpromote expectations to the new values, and add the two `leaveDetail(rows:)` expectations that `PromoteEvidenceDisplayTests` held (deleted in STEP-1) to `PromoteModelTests` as T-12 with the new wording.
Check: the STEP-5 catalog loop, changed to print every key or English value under `promote.` or `evidenceGraph.subject.unpromote` that matches `(un)?fil(e|ed|es|ing)` (no leading ``, so "unfiled" is caught), prints nothing; `rg -n -i '(un)?fil(e|ed|es|ing)|fileCount|filedRows|alreadyFiled|filesOn|filedEarlier|filedTitle|filedBody|countFiled|alreadyFiledHeading' macos/App/Features/Promote macos/ProvenenciaTests/Promote*.swift` finds nothing; Xcode tests pass with T-12, T-13.

**STEP-7 — Docs and skills the removal makes false**
Files: `docs/promote-graph-alignment.md`, `docs/ideas/wide-layout.md`, `docs/interpretation-layer-data-model.md`, `docs/matching-model.md`, `.agents/skills/add-localized-string/SKILL.md`, `.agents/skills/add-seeded-vocabulary/SKILL.md`, `.agents/skills/add-match-rule/SKILL.md`, `docs/ideas/term-management.md`, `docs/ideas/international-names.md`, `docs/ideas/binding-order-editing.md`. Skill: —.
Do, with these replacements:
- `promote-graph-alignment.md` line 7: delete the sentence "Until the code catches up, the sheet and the pin writes described in older PR plans still run." Line 35: the sentence ends at "is surfaced." (delete ", and every conflict the engine found shows as *contested*") Line 64 (§3.1, the bullet that starts "Rows whose suggestion changed"): append "A row whose suggestion changed for any other reason shows the mark alone." §4 and §8 need no change.
- `docs/ideas/wide-layout.md` line 15: the row "| Promote | To be decided in its brief | |" leaves the table of pages still on their Regular layout; add one sentence after the table: "Promote widened in S13-23: its list goes to 1600 pt and its target menu to 460 pt at Wide (S13-D11 frame 27W)."
- `interpretation-layer-data-model.md` line 693: the first two sentences ("`sort_order` orders Add-property menus and Properties lists, the conclusion page's fields and the claim evidence sheet's sections (one order, cached with the vocabulary as `graphcache.Vocabulary.BindingOrder`)." and "The sheet is to be removed" with its §5.3 link) become "`sort_order` orders Add-property menus and Properties lists and the conclusion page's fields (one order, cached with the vocabulary as `graphcache.Vocabulary.BindingOrder`)."
- `matching-model.md` line 5: "The registry's `evidence` section goes with the claim evidence sheet." becomes "The claim evidence sheet is gone (S13-23); the registry's `evidence` section goes with the pin engine (S13-24)." Line 122: "It is deprecated and goes with the sheet and the pins" becomes "It is deprecated: the sheet it served is gone (S13-23), and it goes with the pins (S13-24)".
- `add-localized-string` line 147: replace "(`PromoteEvidenceDisplay.Role`)" with nothing. Line 151: replace the example "the evidence sheet's subtitle is `promote.sheet.subtitle.person`, `.event` and `.place`, each a whole sentence, not one format with a "person" or "event" argument" with "the Promote menu's new-record option is `promote.new.person`, `.event` and `.place` (`L10n.Promote.newOption(_:)`), each a whole phrase, not one format with a "person" or "event" argument", and "Existing splices outside the evidence sheet" becomes "Existing splices".
- `add-seeded-vocabulary` line 80: "its entry in `L10n.PropertyTerm.wording` (`macos/App/Platform/L10n.swift`): the term's label, and for a role, event type or relationship type its share phrase and similar phrase, each whole, and for a relationship type that an implied link can carry its `recordedDirect` phrase ("recorded as spouses", which the evidence sheet's caption uses), with the catalog keys" becomes "its label constant and its case in `L10n.PropertyTerm.resource` (`macos/App/Platform/L10n.swift`), with the catalog key".
- `add-match-rule` line 76 (D-12): "A code that reaches the wire (a relation, a value cause, an outcome or an outcome cause) also gets a case in its enum in `macos/App/Platform/ConclusionVocabulary.swift` (`ValueRelation`, `ValueCause`, `EvidenceOutcome`, `EvidencePropertyCause`), and stays in its row of `TestSwiftVocabularyMatchesGo` (`api/ffi/swift_vocabulary_test.go`), which fails when the two sides drift. UI copy built from it follows." becomes "No comparison code reaches the Mac app since the evidence sheet went (S13-23), so a code has no Swift enum." Also "the evidence sheet does not read it, so it has no Swift enum" becomes "it has no Swift enum". Line 91: delete the sentence "Then add its arm to `PromoteEvidenceDisplay.wording` and its two `promote.sheet.recorded.*` keys; the `switch` does not compile without it." and replace "Add the pattern's `KinshipPattern` case in `macos/App/Platform/ConclusionVocabulary.swift`: `TestSwiftVocabularyMatchesGo` reads `kinship.Patterns()` and fails until it is there." with "A pattern has no Swift case since the evidence sheet went (S13-23)."
- Ideas: `term-management.md` line 28, "The evidence sheet already reads the inverse through `neighbor_term` (the engine reads each relationship from the neighbor's side, S11-33); only the Citation Composer still reads the Swift map." becomes "The Citation Composer is the only reader of the Swift map (the evidence sheet, which read the inverse through `neighbor_term`, went in S13-23)." `international-names.md` line 43: "S11-33 made every sentence on the Promote evidence sheet one whole translatable phrase" becomes "S11-33 made every sentence on the Promote evidence sheet (removed in S13-23) one whole translatable phrase", and "`L10n.PropertyTerm.wording`" becomes "`L10n.PropertyTerm.resource`". `binding-order-editing.md` line 7: "sets the conclusion page's fields and the claim evidence sheet's sections" becomes "sets the conclusion page's fields".
Check: `python3 scripts/check-guidance-names.py` passes; `rg -n 'evidence sheet|PromoteEvidence|promote\.sheet|recordedDirect|share phrase|PropertyTerm\.wording' .agents .cursor docs/*.md docs/ideas/*.md` finds only lines that say the sheet went in S13-23.

**STEP-8 — Plan, map and brief status**
Files: `docs/deployment-plan/spike-13/completed.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (map row), `docs/deployment-plan/spike-13/design/S13-D11-promote-without-assessment.md` (its gate table marks S13-23 shipped; it moves to the design archive when S13-30 lands), `docs/deployment-plan/spike-13/design/README.md`, this plan (status). Skill: —.
Do: per the Definition of done. The map row's Plan column already links this plan; its Status becomes `Landed in #N (date)`.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-23-promote-stops-asking-for-pins.md` passes; links resolve. S13-D11 stays in `design/` (it also gates S13-30) with its S13-23 rows marked shipped.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aSkipRowFilesNoArgumentConfidenceOrPins` → renamed `doneSendsATargetPerFiledRow` (STEP-4) | A row on `e1`, a New row, a Skip row and an unset row → `batchRows()` is `[(a, "e1", .handle), (b, "", .new), (c, "", .skip)]` as `CatalogPromoteGraphAlignmentFixed` (`subjectID`, `handleID`, `target`); the unset row is absent | AC-1.3, D-6 |
| T-2 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `doneSendsDefaultOrChosenPins` → renamed `doneFilesEveryRowWithATargetOnly` (STEP-4) | Done on a FakeStore page → `lastPromoteBatch` holds each filed row's subject, handle and target; `recordedCalls` holds no `getPromoteClaimEvidence` call and no grades call; the filed claims have status accepted | AC-1.3, AC-1.4, D-14 |
| T-3 | Swift | `PromoteFlowTests.swift` `theCaptionReasonIsTheMoveWhenThereIsOne` → renamed `theCaptionIsTheMoveElseTheViaNeighbor` (STEP-2) | Proposal row with `moveReason` `.neighborMoved`, `moveNeighborSubjectID` "n" and `viaNeighborSubjectID` "a" → `caption == .neighborMoved(neighborID: "n")`; only `viaNeighborSubjectID` "a" → `.via(neighborID: "a")`; neither → nil; `moveReason` `.taken` and no via → nil; `moveReason` `.neighborPlaced` with an empty neighbor → nil | AC-2.1 – AC-2.3, D-3, D-4 |
| T-4 | Swift | `PromoteFlowTests.swift` `retargetingClearsContestedAndMoveUntilTheNextProposal` → renamed `retargetingClearsTheCaptionUntilTheNextProposal` (STEP-2) | A row with a move caption is retargeted → `caption` nil, `updated` false; the next proposal sets it again | Edges, D-3 |
| T-5 | Swift | `PromoteFlowTests.swift` `matchMenusUseConclusionTitles` (STEP-4) | Rows and alternatives built with `header: .person(…)` → menu titles as today, sorted form | D-7 |
| T-6 | Swift | `PromoteModelTests.swift` `movesAreWordedNamingTheNeighbor`, `aViaReasonNamesTheTermInLowercase`, `reasonsAreWordedWithLabelsAndRefs` (narrowed in STEP-2 to the via arms: a row "b" with `viaNeighborSubjectID` "a", where row "a" (name Gracie) is on PER-Z and the page's bridge between them has term "spouse"; the same with a bridge with no term; and `viaNeighborSubjectID` "x" naming no row; the `agreesOn` and `captionReason` parts go) | → "after Gracie moved to PER-Z", "via Gracie → PER-Z (spouse)", "via Gracie → PER-Z", "Reached from a matched neighbor" | AC-2.1, AC-2.2, AC-2.4 |
| T-7 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `queryKeysMatchDesignTable` (STEP-3) | The Promote place's keys no longer list `.confidenceGradesList` | AC-1.4 |
| T-8 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `invalidationsComeFromRegistrySpecs` (STEP-3) | The expectation list drops `.confidenceGradesList`; `confidenceGradesLoadInOrder` is deleted | AC-1.4 |
| T-9 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` (STEP-5) | The remaining rows pass; "every seeded term has wording" reads `resource(propertyKey:termKey:)` and still names every seeded term | AC-3.3, D-15 |
| T-11 | Swift | `macos/ProvenenciaTests/SubjectStoreTests.swift` `promoteJoinsExistingHandleWithGradeAndArgument` → renamed `promoteJoinsExistingHandle` (STEP-4) | Promote a second Subject onto the first one's handle → the same entity, status accepted, membership recorded; an Event handle for a person still throws | AC-1.6, D-17 |
| T-12 | Swift | `PromoteModelTests.swift` `leaveDetailCountsRows` *(new, STEP-6)* | `L10n.Promote.leaveDetail(rows: 1)` → "You chose a target on 1 row. Nothing has been promoted; opening Promote again proposes the graph afresh."; `rows: 3` → "You chose targets on 3 rows. Nothing has been promoted; …" | AC-1.8, D-20 |
| T-13 | Swift | `macos/ProvenenciaTests/L10nResolutionTests.swift` (the test holding the `unpromoteKeep` expectation, STEP-6) | `unpromoteKeep` → "Keep promotion"; `unpromoteConfirmMessage` → "This promotion is removed. The cards stay on the graph, and the handle stays."; `unpromoteConfirmBridgesNote` → "The other end of each stays promoted."; `unpromoteConfirmBridgesHeader` → "Also unpromoted"; `L10n.Promote.duplicateBody(.event)` → "These may be the same event. Combine them on the Evidence graph, or pick another target for one row."; `L10n.Promote.openedFrom(name: "Gracie")` → "Opened from Gracie" | AC-1.8, AC-1.9, D-21 |
| T-10 | Swift | `PromoteModelTests.swift` `leaveGuardAsksOnlyAfterAManualChange`, `PromoteFlowTests.swift` `leaveAsksOnlyAfterAManualChange` | Unchanged: only a target change arms the guard | AC-1.5 |

**Deleted or narrowed, by step** (D-13).

- **STEP-1, deleted:** `PromoteEvidenceBoardTests.swift`, `PromoteEvidenceDisplayTests.swift`; in `PromoteFlowTests`: `aNeverOpenedRowSendsDefaultPins`, `openingDraftsTheCheckedPins`, `reopeningKeepsOnlyOfferedPins`, `aLineTogglesEveryMember`, `evidenceForAnOldTargetIsIgnored`, `evidenceNeighborsAreTheRowsDoneFiles`, `aRetargetClearsEvidenceAndPins`, `everyRowStartsWithAnEmptyArgument`, `aTypedArgumentSurvivesARetargetAndAReproposal`, `aManualPinOnTheSameTargetSurvivesAReproposal`, `anUnsettledRowKeepsItsEvidenceOnlyOnTheSameTarget`, `holdingAFoundRecordResetsTheOldTargetsEvidence`; in `PromoteModelTests`: `openingASheetLoadsItsEvidence`, `aSheetOpenedAfterANameOrderChangeWordsNamesInTheNewOrder`, `aFailedEvidenceReadShowsAnError`, `reloadEvidenceRetries`, `aLateAnswerForAnotherSheetIsDropped`, `theSheetMarksSpeakTheContestedBand`, `theSheetCarriesTheRowsCaption`. **Narrowed:** `aPersonRecordReadsNaturallyInTheSheetAndSortedInTheMenu` keeps its menu assertions and is renamed `aPersonRecordReadsSortedInTheMenu`; `aSkipRowFilesNoArgumentConfidenceOrPins` and `doneSendsDefaultOrChosenPins` drop their pin, grade and argument assertions (STEP-4 rewrites them as T-1, T-2).
- **STEP-2, deleted:** in `PromoteFlowTests`: `anUnplacedRowDraftedOntoAWeakRecordCaptionsItsReason`, `theAssessmentFollowsTheSelectedRecord`, `aContestedRowAndItsMoveReachTheFlow`, `aContestedRowShowsContestedInItsReadout`, `aContestedRowWithoutAHandleShowsNoReadout`, `aDecidedRowTakesTheLatestContestedFlag`, `anAnchorRowIsNeverContested`; in `PromoteModelTests`: `researcherPropertyLabel`, `aChosenRecordBelowWeakReadsLittleEvidence`, `anUnplacedRowIsWordedNoRecordSuggested`, `theContestedBandIsSpoken`. **Narrowed:** the draft tests `aMediumMatchStartsOnItsHandle`, `aWeakMatchStartsOnItsHandle`, `aRowWhoseRecordWasTakenStartsUnset`, `aWeakCandidateBelowTheBarStartsSelected`, `anUnmatchedRowStartsUnselected`, `anUnplacedRowNeverDraftsNewOrSkip`, `aRowThatLosesItsRecordSaysNoRecordIsSuggested` keep their target assertions and drop band, readout and reason assertions; `aFoundRecordSurvivesAReproposalWithItsBandRefreshed` keeps "the found record stays held" and is renamed `aFoundRecordSurvivesAReproposal`; `reasonCodesParseThroughTheWireVocabulary` keeps its `PromoteTargetCode` and `PromoteReasonCode` parsing and drops the `PromoteFlow.Reason(code:…)` block (T-3 covers the initializers); `doneSendsDefaultPinsAndConnectionsFollowEnds` keeps its connection assertions and is renamed `connectionsFollowTheirEnds`.
- **STEP-3, deleted:** `CatalogQueryRegistryTests.confidenceGradesLoadInOrder`. **Narrowed:** `SubjectStoreTests.promoteJoinsExistingHandleWithGradeAndArgument` loses its last two lines (STEP-4 then renames it, T-11).
- **STEP-5, deleted:** in `ConnectionSentenceRegistryTests`: `aNeighborEventIsNamedByItsType`, `aVowelInitialEventTypeTakesAn`, `aFewSharesReadWhole`, `aNeighborPersonIsNamedByRoleAndName`, `aRelationshipIsTheTermTheNeighborCarries`, `aResearchersTermReadsAsAConnectionNamingIt`, `aPlaceRelationshipIsParentWithinSuccessorOrPredecessor`, `aLocationNamesThePlace`, `nothingNamedIsAConnection`, with the helpers `term` and `share`.
- **Kept:** `aPinnedObservationDeleteNamesThePerson` (S13-25).

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteFlowTests`, `PromoteModelTests`, `EvidenceGraphModelTests` (builds proposals through `FakeStore`), `CatalogQueryRegistryTests`, `EventTitleDisplayTests`, `PropertyTermDisplayTests`, `TestSwiftVocabularyMatchesGo`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS on macOS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (STEP-5's command prints [])
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the S13-D11 board and inventory; no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief S13-D11's gate table marks S13-23 shipped; the brief stays in design/ until S13-30 lands (it gates S13-30)
- [ ] Domain doc updated if a rule in docs/*.md changed: promote-graph-alignment.md (status, principle 5, §3.1), interpretation-layer-data-model.md line 693, matching-model.md lines 5 and 122, ideas/wide-layout.md's Promote row (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source whose Evidence graph has unpromoted people, one of whom matches a record promoted earlier (the product owner's test census). Open Promote on one person, then Map the rest. See: no row has a band, Contested badge, reason or chevron; the headings read On this Source and Promotes to; the slot after the target menu is empty; rows, Already promoted rows, subrows and notes line up as frames 27a and 29 draw them; the intro line reads "Opened from <name>" (AC-1.1, AC-1.2, AC-1.9).
2. Click where a row's band used to be. Nothing opens (AC-1.1).
3. Change one person's target to the matching record. A spouse row that moves shows Updated and "Updated · via <name> → PER-… (spouse)" or "Updated · after …" (AC-2.1, AC-2.2); a row that moves for no named reason shows only Updated (AC-2.3).
4. Press Done. The toast says how many filed. Run `sqlite3 <project>/provenencia.sqlite "SELECT count(*) FROM identity_claim_evidence;"` before and after: the count does not change (bar 9, app half).
5. Reopen the same graph's Promote page: the promoted rows show under Already promoted.
6. Widen the window to 2560 × 1440 with the sidebar open: the list widens to 1600 pt and centers, the target menus are 460 pt, and the header and footer bars span the pane (frame 27W). Read the footer, a waiting subrow, Skip in a menu, and the toast after Done: each says promote, never file; on the Evidence graph, Unpromote's confirm says "Keep promotion" (AC-1.8). Narrow it back to 1366 × 900: the layout is as in step 1, and an open record search, each menu's selection and the scroll position survive both resizes (AC-1.7).
7. On the branch, run AC-3.1's, AC-3.2's and AC-3.4's `rg` commands from the repo root. The first two find nothing; the third finds only Source credibility lines (AC-3.1, AC-3.2, AC-3.4).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A kept view reads a deleted `L10n.Promote` member and the build breaks late | STEP-1 and STEP-2 each build and test before the next step |
| `EvidenceGraphModelTests` or another suite builds `CatalogPromoteGraphAlignmentRow` with `score:` or `person:` | STEP-4's `rg -l` lists them; fix the initializer calls in that step |
| Turning `Wording` entries into constants leaves a trailing comma or drops a comment | The Xcode build and `TestSwiftVocabularyMatchesGo` "every seeded term has wording" catch it in STEP-5 |
| S14-15 lands first and the `project.pbxproj` entries are gone | Stop condition below |

**Stop and report if:**

- `which xcodebuild` finds nothing: this PR needs macOS (D-16).
- The plan's Status is not `Planned` (the board is not agreed, or the plan has not been refreshed against it).
- `grep -c fileSystemSynchronized macos/Provenencia.xcodeproj/project.pbxproj` is not 0 (S14-15 landed): skip the `project.pbxproj` edits and report it.
- A file outside §4 names any type or member §3 deletes (`rg` it before deleting); report the path.
- `api/ffi/handlers/promote_align.go` sets `ViaNeighborSubjectId` other than under `r.Via != nil`, or `core/graphalign/rows.go` sets `Row.Via` with a reason other than `via` (D-3's assumption).
- Go Apply rejects a row with no `pin_ids` and `default_pins` false (any error from `TestApplyPromoteGraphAlignment` in `api/ffi/handlers/promote_batch_test.go`, or from Done in the manual check).
- A test outside §8 relies on `FakeStore.applyPromoteGraphAlignment` throwing `promote.invalid` or `identityclaims.already_member` (D-14).
- The agreed board draws something on the row that the S13-D11 inventory does not list, or a frame gating this PR (27a, 27W, 28, 29) contradicts a Decision here.

## 12. Open questions

None.

## 13. PR

**Title:** S13-23 — Promote stops asking for pins
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-23-promote-stops-asking-for-pins.md`.

## Landed note

The code differs from the plan in these ways:

- STEP-1 and STEP-2 are one commit, and STEP-3 and STEP-4 are one commit: the sheet and the column share `PromoteFlow` and `PromoteModel`, and the store calls and Done's mapping share `GenealogyStore`, `GoStore` and `FakeStore`, so neither pair built apart. STEP-3 also deleted `ConnectionSentenceRegistry.ShareInput.init?(_:neighborName:)` early, because it named `CatalogClaimEvidenceRelationship`.
- STEP-6's `rg` also matches "profile", "Filed" and "Files" in test names: `doneFilesEveryRowWithATargetOnly` is `donePromotesEveryRowWithATargetOnly`, `doneSendsATargetPerFiledRow` is `doneSendsATargetPerPromotedRow`, `promoteAllStartsExpandedAndHidesFiledBridges` is `promoteAllStartsExpandedAndHidesPromotedBridges`, `aFoundRecordIsProposedHeldAndFiled` is `aFoundRecordIsProposedHeldAndPromoted`, and the `PromoteFlow` comment "Default profile orders" reads "Default name orders".
- T-2 drops the two assertions that named `getPromoteClaimEvidence` and `listClaimConfidenceGrades` calls (AC-3.1 forbids the names; the calls no longer exist). It sends one New and one Skip row, because the fake holds no record to promote onto.
- T-10's two tests arm the guard with a target change, not `setArgument` (deleted).
- T-6's `reasonsAreWordedWithLabelsAndRefs` became `aViaCaptionNamesTheNeighborByItsRecord`.
- `CatalogQueryRegistryTests.everyKindHasOneDescriptor` counts 18 kinds, not 19.
- `GenealogyStore` and `FakeStore` doc comments that called the process filing say promote or apply; `add-localized-string`'s "a share phrase with its article" example reads "a phrase with its article".
- The manual check (section 10) was not run in the app.
- The S13-D11 brief's Paste block was refreshed to the current `add-design-brief` wording (pane-only frames), because `check-plan-docs.py` fails a brief it checks that differs from the skill, and this PR changes the brief's gate table.
