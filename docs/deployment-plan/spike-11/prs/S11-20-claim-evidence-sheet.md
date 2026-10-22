# S11-20 — Claim evidence sheet

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 5 — Claim evidence
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S11-D2](../design/S11-D2-evidence-sheet.md), board agreed 2026-10-22
**Board:** [Claude Design — Promote flow](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share), frames E01 – E14, agreed 2026-10-22. The implementer cannot open the link from every host: S11-D2 §5.3 gives every frame's content and [§9 Board notes](../design/S11-D2-evidence-sheet.md#9-board-notes-agreed-2026-10-22) its layout where it differs from §4; together they are the board in words. Frame E*nn* is §5.3 Frame *nn*
**Depends on:** S11-25 (landed [#517](https://github.com/mendahu/provenencia/pull/517); its Landed note is folded into §4), S11-32 (landed [#523](https://github.com/mendahu/provenencia/pull/523); its Landed note is folded into §4), S11-D2 (agreed)
**Blocks:** —
**Branch:** `s11-20-claim-evidence-sheet`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/design-system-layers.mdc`
**Context:** [S11-D2](../design/S11-D2-evidence-sheet.md) (requirements ES-1 – ES-22, copy table, inventory); [`claim-evidence.md`](../claim-evidence.md) (Lines, Wording, Evidence through relationships); [S11-D1](../design/S11-D1-promote-rows.md) (band, Contested, captions, the "·" connection form); [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3; working rules "The pin policy has one owner, in Go", "One wire vocabulary"; S11-18 review M-2, M-6 (subtitle and neighbor naming); S11-12 D-12 (a side that contradicts itself)
**Planned:** 2026-10-22 against `main` @ `6504003`; refreshed for the board, S11-25's and S11-32's landings against `main` @ `05db499`
**Review:** [`S11-20-claim-evidence-sheet.review.md`](S11-20-claim-evidence-sheet.review.md) — round 1 on the draft: Fits with changes / Not ready; round 2 at the STEP-0 refresh (see the log); every finding fixed or declined

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

The Promote evidence sheet answers two questions and keeps them apart. The title carries why the row was suggested: its band and Contested badges, with the row's reason as their tooltip, and under the title the Updated caption and a contested line when they apply. The claim fields follow. Then "What's the evidence that this is <record>?" lists every member's evidence as sentences the researcher checks or unchecks: property sections, collapsed lines that expand to each record, relationships that name the connection that matched, a lifespan line that is never pinned, and likenesses as two crops side by side. The table with its Weight column and the "Through <name>" cards go.

**Why now.** S11-25 moved the sheet onto the evidence RPC in today's shape; S11-32 added the lifespan and likeness lines and ordered sections by the catalog's binding order. This PR is the redraw S11-D2 specifies, and closes the dogfood bar item "the evidence sheet shows all the evidence".

**What stays.** The data: `CatalogClaimEvidence` and every Go path (no wire change). Pins as S11-25 left them: `PromoteFlow.Row.pins`, `togglePin(subjectID:pinIDs:)`, `setEvidence`, `defaultPins`, the lazy load and `SheetEvidence` states. The panel, its title, the claim fields, the footer. The Promote rows, which only change by sharing their readout badges with the summary.

## 2. User stories

Conventions. "Evidence E1" is a `CatalogClaimEvidence` with one `name` section holding one line: relation `same`, outcome `agree`, pinnable, checked, `incoming` the name "John Smith", five members (pin ids `p1`…`p5`) on Sources "Census 1851" … "Census 1891", each value "John Smith". "Row R" is a `PromoteFlow.Row` named "John Smith" with target `.handle(id: "h1", ref: "PER-7QH3D", title: "John Smith")`, readout assessment `.strong`, readout reason `.via` a neighbor "Mary Smith" whose `shownRef` is "PER-4KX2M".

**US-1.** As a researcher, when I open a row's sheet, I first read why it was suggested, then the evidence, so that I never mistake matching's reasons for the pins I file.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Row R, contested | No subtitle. The Strong and Contested Badges sit at the right end of the title; their tooltip is "via Mary Smith → PER-4KX2M" and they read as one element, "Why this was suggested: Strong, contested, via Mary Smith → PER-4KX2M". Under the title: "Something below disagrees with this record." (ES-1, ES-2, board E03) |
| AC-1.2 | Row R, `updated`, move `.neighborMoved` | The Badges' tooltip is the readout reason "via Mary Smith → PER-4KX2M" (as the row button); under the title, "Updated · after Mary Smith moved to PER-4KX2M" in italics. Row R `updated` with no move: "Updated · via Mary Smith → PER-4KX2M". Not `updated` and not contested: nothing under the title (ES-2, board E12) |
| AC-1.3 | Row R with E1, all five pins checked | The claim fields first; then the heading "What's the evidence that this is PER-7QH3D John Smith?" with meta "5 checked" and the sentence "Check every line that supports this match. Checked lines are pinned to the claim when you press Done."; while loading, no meta (ES-3, ES-21, board E01) |
| AC-1.4 | A row set to New, Skip or unset | No Badges; the EmptyState "Pick a record to see its evidence"; claim fields shown for New, hidden for Skip and unset (ES-19, board E13c, E13d) |
| AC-1.5 | Any row with claim fields | Status over Confidence in a 200 pt column, beside the argument field labelled "Additional comments or arguments about this claim (optional)", "(optional)" muted; no hint under any field (ES-21, board) |

**US-2.** As a researcher, I see every member's evidence as a sentence, and I can uncheck one record of a collapsed line.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Row R with E1 | A "Name" Card with one line "Same name · 5 records", Badge Agrees, box checked, "Show 5 records" (ES-5, ES-6, ES-9) |
| AC-2.2 | AC-2.1, expand, uncheck Census 1861 | Five member rows; Census 1861 unchecked; the line's box partly checked; pins `p1, p3, p4, p5`; heading meta "4 checked" (ES-9, ES-10) |
| AC-2.3 | AC-2.2, press the line's box | All five checked (ES-10) |
| AC-2.4 | A `date` line `contains` / `precision` | "Same date, recorded less precisely", Badge Similar, box unchecked and enabled (ES-8) |
| AC-2.5 | A `sex_at_birth` line whose incoming is negative "male", outcome `conflict` | "Disagrees", Badge Conflicts, box disabled; "This source: not male"; the conflict Callout under the heading (ES-4, ES-7) |
| AC-2.6 | Every relation × cause in the ES-8 table, for a name value, a date value, and a term value | The table's sentence (T-2) |

**US-3.** As a researcher, a relationship names the connection that matched, and checking it pins what is behind it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A relationship `participation`, outcome `agree`, checked, one bundle, one `date` section; on the page, the bridge between John and the birth row reads "John Smith · subject of · Birth of John Smith" | A Card headed "They share a matching connection", then "John Smith · subject of · Birth of John Smith", the birth row's `shownRef`, Badge Agrees, box checked; the date line under it, indented (ES-12, ES-13) |
| AC-3.2 | The same on the birth (event) row, whose neighbor is John | "They share a matching connection" — the same bridge's words (ES-13) |
| AC-3.3 | Outcome `partial` (S11-D2 Frame 4: the marriage dated 1840 against 12 Oct 1843) | "They share a similar connection", Badge Similar, box unchecked and enabled; the "Different dates" line under it (ES-13). No relationship ever arrives with outcome `conflict` (`claimevidence` keeps agree and partial only) |
| AC-3.4 | A relationship with no sections; a relationship whose neighbor has no bridge in `flow.bridges` | "Nothing else to compare."; pressing its box toggles its bundle ids. No bridge: the second line reads "with <neighbor name>" (ES-13, ES-14) |
| AC-3.5 | No sheet string reads "Through" | `rg -n 'promote.sheet.through\|Through %@' macos/App` prints nothing |

**US-4.** As a researcher, I see a person's lifespan check and their likenesses.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Lifespan agree, this side's span 14 May 1817 – 2 Jun 1890, the record's 14 May 1817 – 2 Apr 1871 (S11-D2 Frame 1) | A "Lifespan" Card, last among own Cards: "The lifespans fit: 14 May 1817 – 2 Jun 1890 here, 14 May 1817 – 2 Apr 1871 on the record", Badge Agrees, no checkbox, the caption "Worked out from dated events. A guide for matching, never pinned." (ES-16) |
| AC-4.2 | Lifespan conflict, the record's span 3 Mar 1741 – 9 Jan 1810 (S11-D2 Frame 10: the two lives do not overlap) | "The lifespans don't fit: 14 May 1817 – 2 Jun 1890 here, 3 Mar 1741 – 9 Jan 1810 on the record", Badge Conflicts; the conflict Callout shows. A span of one date prints it once: "1843" (ES-16) |
| AC-4.3 | This side contradicts itself; the record has no dated events; neither side has a dated event | "No lifespan for this source: its own dates contradict each other"; "No lifespan to compare: the record has no dated events"; Badge Not compared; neither: no Lifespan Card (ES-16) |
| AC-4.4 | A `likeness` line, same image, two members | A "Likeness" Card after Name; two 96 pt crops (this Source's, then the first member's), each over its caption; "Same likeness · 2 records", Agrees, checked (ES-15) |
| AC-4.5 | A `likeness` line, two images | "Compare these likenesses", Badge "Judge by eye", unchecked, enabled (ES-15) |

**US-5.** As a VoiceOver user, each line reads as one sentence, and the checkbox speaks its own state.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | AC-2.1's line | Label "Name. Same name, 5 records. This source: John Smith."; the checkbox value "checked"; a partly checked box speaks "partly checked" (ES-22) |
| AC-5.2 | The lifespan line | Its sentence only; no checkbox element (ES-22) |
| AC-5.3 | Two likeness lines, one same image (agree) and one different image (unjudged), for two members of one handle | Their labels and checkbox labels differ: each names the member's Source and the line's outcome, so two likeness pins never read alike (S11-32 review UX-1: today's sheet words both "Pin Likeness: Likeness and Likeness") |

## 3. Scope

| In | Covered by |
| --- | --- |
| Kit: `PVSectionHeader` title as `PVCopy`; `PVCheckbox` speaks "partly checked"; `PVPanel` title accessory; `PVField` "(optional)" marker (S11-D2 inventory, Extend) | STEP-1, D-11, D-15, D-18, D-19, AC-1.5, AC-5.1 |
| Shared pieces: the readout badges, the likeness tile, the life span text, the neighbor naming, the bridge kind on `BridgeFact` | STEP-2, D-8, D-9, D-13, D-14, D-16 |
| Pure display helpers (D-2) | STEP-3, AC-2.2, AC-2.3 |
| Copy: ES-8 table, relationship and lifespan sentences, summary, heading, states, spoken labels; changed and removed keys (D-4, D-5, D-12) | STEP-4, AC-2.6 |
| Sheet shell, title marks, the lines under the title, claim fields, heading, states (ES-1 – ES-4, ES-17 – ES-21 as S11-D2 §9 changes them) | STEP-5, US-1 |
| Property Cards, lines, collapse (ES-5 – ES-11) | STEP-6, US-2 |
| Relationship Cards (ES-12 – ES-14) | STEP-7, US-3 |
| Lifespan and likeness Cards (ES-15, ES-16) | STEP-8, US-4 |
| Spoken labels, identifiers, removal of the old table | STEP-9, US-5, AC-3.5 |
| Docs | STEP-10 |

| Out | Home |
| --- | --- |
| Any Go, proto or store change, including section order | S11-25, S11-32 (a missing field is a Stop condition) |
| Opening the sheet from a connection subrow | not planned (decided here for S11-D1's note) |
| Select-all / pin-nothing | not planned (S11-D2 §2.1) |
| Face comparison of two likenesses | not planned |
| Statuses beyond Accepted | a later spike (today's hint) |
| Moving `PlacePeriodDisplay.span` onto `PersonLifeDisplay.dateSpan` | not planned (its open and until copy differ: `placePeriodOpen`, `placePeriodUntil`) |
| The same indeterminate gap in `Features/Sources/SourceArtifactChecklist.swift` | fixed by D-15 for every caller; no change there |

## 4. Current state

Verified at `6504003`, then read against S11-25's plan for the rows it changes. Rows marked **after S11-25** or **after S11-32** describe code those PRs write; recheck them at your base.

| Area | Where | Today |
| --- | --- | --- |
| Sheet | `macos/App/Features/Promote/PromoteView.swift` `private struct PromoteEvidenceSheet` (lines 664 to about 950, after S11-25), `PromoteView.body` `.sheet(item: sheetBinding)` (lines 68–70), `private enum PromoteLayout` (line 952); the file is 979 lines | A `PVPanel(width: 880)` with a conflict callout, `sheetColumns` (Pin, Compared, This source, target, Result, Weight), one `PVCard` per `ExhibitGroup` ("This subject", "Through <name>"), `comparisonLine`, claim fields, "Back to all rows". `sheetSubtitle` uses `L10n.Promote.assessment(readout.assessment)`, not `assessmentText`; group refs use `row.ref`, not `shownRef` |
| Row readout | `PromoteView.swift` `PromoteAlignmentRow` (band and Contested `PVBadge`s, lines 490–498; `if row.updated` caption with `captionReason`, lines 525–529), `private func badgeTone(_:)` (lines 566–573) | The band, Contested and reason the summary repeats |
| Sheet after S11-25 | same file, `PromoteEvidenceSheet` from line 664 at `376eb9a` | `PVPanel(title:subtitle: sheetSubtitle, width: 880)`; claim fields after the evidence (`PVField(label: status, hint: sheetStatusHint)`, confidence, `PVField(label: argument)` with no hint, identifier `promote.sheet.argument`); `sheetColumns`; a danger `PVCallout(title: evidenceUnavailable)` on failure; three `PVSkeletonBar`s (`promote.sheet.loading`); relationship cards titled `through(name)` with a header `PVCheckbox` spoken `pinRelationship(name)`, identifier `promote.sheet.relationship.<neighborSubjectID>.<key>`; `evidenceLine` (from line 849) binds a `PVCheckbox` to a line's pin ids; `weightText` (line 919) |
| Evidence types after S11-25 | `macos/App/Platform/GenealogyStore.swift` `CatalogClaimEvidence{sections, relationships, offers, sourceTitles}`, `checkedPinIDs`, `offeredPinIDs`, `CatalogClaimEvidenceSection{propertyKey, lines}`, `CatalogClaimEvidenceLine{propertyKey, outcome, propertyCause, relation, cause, points, pinnable, checked, incomingObservationID, incoming, incomingNegative, members}`, `CatalogClaimEvidencePin{id, memberSubjectID, observationID, sourceID, value}`, `CatalogClaimEvidenceRelationship{neighborSubjectID, key, bridgeType, term, end, eventType, outcome, cause, points, pinnable, checked, bundles, sections}`, `CatalogClaimEvidenceBundle{id, memberSubjectID, sourceID}` | S11-25 D-11. `relationship.outcome` is assumed to be `EvidenceOutcome` and `cause` `SegmentCause` |
| Vocabulary after S11-25 | `macos/App/Platform/ConclusionVocabulary.swift` `EvidenceOutcome`, `EvidencePropertyCause`, `ValueRelation`, `ValueCause`, `SegmentCause` | S11-25 D-10; `L10n.Promote.outcome(_: EvidenceOutcome)` |
| Added by S11-32 | `macos/App/Platform/GenealogyStore.swift` `CatalogClaimEvidence.lifespan: CatalogClaimEvidenceLifespan?` (nil for a kind with no lifespan, or when neither side has dates and neither contradicts itself), `CatalogClaimEvidenceLifespan{outcome, propertyCause, relation, cause, row, record}`, `CatalogClaimEvidenceLifeSide{first: CatalogDateValueInput?, last: CatalogDateValueInput?, missing: LifespanMissing}` (L407–431), `CatalogConclusionValue.depiction(CatalogDepictionValue)` (L667); `ConclusionVocabulary.swift` `ValueCause.unjudged`, `ValueRelation.none`, `LifespanMissing{none, undated, contradicts, unrecognized}` (L227); `ReconciledValueDisplay.string(for:)` words `.depiction` as `L10n.Conclusions.likenessValue` ("Likeness"); sections arrive in the catalog's binding order (`graphcache.Vocabulary.BindingOrder()`), person seeded `name`, `likeness`, `sex_at_birth`, `gender`, `interests` | S11-32 Landed note; verified at `05db499` |
| Flow | `macos/App/Features/Promote/PromoteFlow.swift` `Row` (`name`, `ref`, `shownRef`, `target`, `selectionReadout`, `captionReason` = `move ?? reason`, `updated`, `move`, `contested`, `pins`), `row(_:)`, `bridges: [BridgeFact]`, `struct BridgeFact{id, sentence, phrase, endAName, endBName, mark, endA, endB, alreadyFiled}` (line 218) | `BridgeFact` has no bridge kind |
| Model | `macos/App/Features/Promote/PromoteModel.swift` `reasonText(_:on:)` (from line 241; `.via` and the move cases look up `flow.row(id)`, `.name`, `.shownRef`), `assessmentText(_:)` (line 279), `propertyLabel(key:origin:)` (line 285), `facts()` (line 358; `BridgeFact` built from `snapshot.bridges` from line 392, `bridge.kind` in hand), `session.projectKey.projectDir` | `assessmentText` gives "Strong, contested" |
| Values to text | `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `string(for:locale:orders:)` (words `.depiction` as "Likeness", S11-32); `macos/App/Features/Conclusions/PersonLifeDisplay.swift` `private static func dateSpan(birth:death:locale:)` (line 37: "1817–1880", open and death-only forms through `L10n.Conclusions.lifeSpan*`, each end through `DateRowDisplay.line`) | S11-25 words line values with the first |
| Likeness tile | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailLikenessSection` `tileView` (a 96 pt `CachedThumbnail(fallbackMark: .subjectPerson)` over a one-line muted Source caption, around line 270) | The one crop-and-caption today |
| Kit | `macos/App/DesignSystem/Components/Panel/PVPanel.swift` `PVPanel<Body, Footer>` (`title: Text`, `subtitle: Text?`; no title accessory); `macos/App/DesignSystem/Components/Field/PVField.swift` (`label`, `hint`, `error`, `required` drawn with `L10n.DesignSystem.requiredMarker`; no optional marker); SwiftUI `.help` is the kit's tooltip (`macos/App/DesignSystem/README.md`, Tooltip row); `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift` (stores `title: LocalizedStringResource`, renders `Text(title)`); `macos/App/DesignSystem/Components/Copy/PVCopy.swift` `protocol PVCopy` (`pvText`), used generically by `PVDisclosureButton<Title: PVCopy>`; `macos/App/DesignSystem/Components/Checkbox/PVCheckbox.swift` (`isIndeterminate` drawn by `PVCheckboxMark`, spoken state from `isChecked` only); `PVBadge`, `PVCallout`, `PVCard(tone: .sunken)`, `PVSkeletonBar`, `PVEmptyState` | — |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Promote` (from line 1297): `sheetSubtitle`, `through`, `ownRecords`, `sheetNeighbor`, `pinLine`, `pinRelationship` (S11-25), `negatedValue`, `conflictCalloutTitle`, `conflictCalloutBody`, `argument` ("Argument"), `sheetStatusHint`, `evidenceUnavailable` (S11-25), `sheetHere`, `sheetPin`, `sheetCompared`, `sheetResult`, `sheetWeight`, `outcome(_: EvidenceOutcome)`; `argumentHint` is gone (S11-25); `macos/App/Resources/Localizable.xcstrings` | Today's sheet strings |
| Project | `macos/Provenencia.xcodeproj/project.pbxproj` | Lists every Swift file; new files are registered there |
| Tests | `macos/ProvenenciaTests/PromoteModelTests.swift` `theContestedBandIsSpoken`, `reasonsAreWordedWithLabelsAndRefs`; `macos/ProvenenciaTests/PromoteFlowTests.swift` (S11-25's T-10 tests); `macos/ProvenenciaTests/L10nResolutionTests.swift` | No test renders the sheet; there is no UI-test target |

**Precedent.** S11-18 ([#507](https://github.com/mendahu/provenencia/pull/507)) built S11-D1's frames on the same page: kit Badges and Callouts composed in `PromoteView.swift`, copy through `L10n.Promote`, wording tested on `PromoteModel` (`theContestedBandIsSpoken`). Copy that: words in `L10n` functions that switch on wire enums (`L10n.Promote.outcome`, `assessment`), choices in a pure `*Display` enum tested without a view (`ReconciledValueDisplay`, `PersonLifeDisplay`).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The sheet moves to its own file.** `macos/App/Features/Promote/PromoteEvidenceSheet.swift` *(new)* holds `struct PromoteEvidenceSheet` (internal) and its private subviews: `EvidenceTitleLines` (D-8's lines under the title), `EvidenceSectionCard`, `EvidenceLineRow`, `EvidenceMemberRow`, `EvidenceRelationshipCard`, `EvidenceLikenessLine`, `EvidenceLifespanCard` *(all new)*. `PromoteView.swift` keeps only the `.sheet` call. Every new file is registered in `project.pbxproj` | The rethink replaces the sheet wholesale; `PromoteView.swift` is 979 lines | Keeping it in `PromoteView.swift` |
| D-2 | **Pure display helpers.** `macos/App/Features/Promote/PromoteEvidenceDisplay.swift` *(new)*: `enum PromoteEvidenceDisplay` with `enum CheckState { case checked, unchecked, partial }`; `checkState(pinIDs: [String], pins: Set<String>) -> CheckState` (all in → checked, none → unchecked, else partial; empty → unchecked); `checkedCount(_ e: CatalogClaimEvidence, pins: Set<String>) -> Int` (`pins ∩ e.offeredPinIDs`); `hasConflict(_ e:) -> Bool` (any line in any section or relationship section with outcome `.conflict`, or a lifespan with outcome `.conflict`; a relationship never has outcome `.conflict`); `bridge(for r: CatalogClaimEvidenceRelationship, row: String, in bridges: [PromoteFlow.BridgeFact]) -> PromoteFlow.BridgeFact?` (the first bridge whose `kind == EvidenceBridgeKind(rawValue: r.bridgeType)` and whose `{endA, endB}` is `{row, r.neighborSubjectID}`, in `bridges` order) | One testable home for every choice the view makes; the `*Display` naming of `ReconciledValueDisplay` | Logic inside SwiftUI bodies — untestable (no UI target) |
| D-3 | **Order comes from Go.** Sections and relationships render in the order `CatalogClaimEvidence` gives. Sections follow the catalog's binding order for the kind (S11-32; person seeded `name`, `likeness`, `sex_at_birth`, `gender`, `interests` on a new catalog); relationships keep S11-19's order (one-cardinality keys, birth and death, first). The lifespan Card is last among own Cards. Section Card headers are `model.propertyLabel(key:origin: "provenencia")`, likeness included | One owner of order (review M-2) | A Swift rank or a `likeness` special case |
| D-4 | **Words through `L10n.Promote`, switching on wire enums.** `lineSentence(relation: ValueRelation, cause: ValueCause, propertyCause: EvidencePropertyCause, value: CatalogConclusionValue) -> LocalizedStringResource` *(new)* implements the ES-8 table: the `value` case picks the name, date or generic variant; `propertyCause == .negative` or `relation == .different` → the "Disagrees" family; `ValueCause.unjudged` → "Compare these likenesses"; `.same` on a depiction → "Same likeness"; `.unknown` (other causes) and `.unrecognized` → the generic "Similar". `outcome(_ outcome: EvidenceOutcome, cause: ValueCause = .none) -> LocalizedStringResource` (signature changed, key kept): Agrees, Similar, Conflicts, "Judge by eye" for `.unjudged`, "Not compared" for `.unknown` and `.unrecognized`; a relationship and the lifespan call it with their outcome only. `ValueRelation.none` (an empty relation, the lifespan with a side missing) words as "Not compared" | The precedent (`outcome`, `assessment`); "One wire vocabulary": codes parsed once, worded once | A `Wording` type in the feature folder |
| D-5 | **Counts are plural keys.** "· %lld records", "%lld checked", "Show %lld records", ", %lld records" (spoken): one key each with `plural` variations in `Localizable.xcstrings` | `add-localized-string` §4 | `count == 1` branches |
| D-6 | **Expansion is view state.** `@State private var expanded: Set<String>` on `PromoteEvidenceSheet`, keyed by a line's first pin id; reset when the sheet closes | Not part of the claim | Storing it on `PromoteFlow.Row` |
| D-7 | **Toggles reuse S11-25.** A line's box calls `model.togglePin(subjectID:pinIDs:)` with every pin id (partial → all in, ES-10, which S11-25's "else insert all" does); a member row calls it with its one id; a relationship's box with its bundle ids. A non-pinnable line has a disabled box; the lifespan line has none | One pin owner | A per-member toggle function |
| D-8 | **The title says what the row says (board).** `PromoteReadoutBadges(readout:)` *(new)* in `macos/App/Features/Promote/PromoteReadoutBadges.swift` draws the band `PVBadge` (with the tone map that was `badgeTone`) and the Contested `PVBadge`; `PromoteAlignmentRow` and the sheet both instance it, and `badgeTone` is deleted. The sheet passes it as the panel's title accessory (D-18), with `.help(model.reasonText(readout.reason, on: row))`, focusable, combined into one accessibility element labelled `L10n.Promote.whySuggestedSpoken(assessment: model.assessmentText(readout), reason: model.reasonText(readout.reason, on: row))` *(new)* "Why this was suggested: %1$@, %2$@". No subtitle: `sheetSubtitle` is removed. Under the title, only when one applies, a group labelled `L10n.Promote.whySuggested` holds, when `row.updated`, `L10n.Promote.updatedDetail(model.reasonText(row.captionReason, on: row))` (italic, `PVColor.infoForeground`, as the row's caption) and, when `readout.contested`, `somethingDisagrees` (primary text). A row with no handle target has no Badges | S11-D2 §9 (E01 – E13); S11-18 review M-2; review M-5, X-6 | A subtitle plus a summary Card — the board dropped both as repeating each other |
| D-9 | **Neighbors are found once; each surface keeps its words.** `PromoteModel.neighbor(_ subjectID: String) -> PromoteFlow.Row?` *(new)* = `flow.row(subjectID)`. `reasonText` uses it in its `.via` and move cases and keeps everything else as today: its own fallbacks (`reasonVia`, `reasonMoveUnknown`) and the move cases' `target.handleRef`, so `reasonsAreWordedWithLabelsAndRefs` and `movesAreWordedNamingTheNeighbor` pass unchanged. The sheet uses it for a relationship's ref (`shownRef`) and the "with <name>" fallback, with `sheetNeighbor` when it is nil | One lookup (review M-10); a missing neighbor reads differently in a reason and on the sheet (review X-14) | A tuple that cannot say "missing" |
| D-10 | **Values.** A line's and a member's value text is `ReconciledValueDisplay.string(for:orders: model.session.nameFormat)` as S11-25 does; a negative incoming is `negatedValue`. A depiction value draws D-16's tile. This Source's crop caption is "This source" (`L10n.Promote.thisSourceCaption`); a member's is `sourceTitles[pin.sourceID]`. A collapsed same-image line shows the first member's crop | One value-to-text path | A local image loader |
| D-11 | **Section header takes `PVCopy`.** `PVSectionHeader<Title: PVCopy, Aside, Actions>` stores `title: Title` and renders `title.pvText`, as `PVDisclosureButton` does. Existing callers pass a `LocalizedStringResource` and compile unchanged; `Text(resource)` and `resource.pvText` render the same. The heading is `PVSectionHeader(title: L10n.Promote.evidenceHeading(ref:title:), meta: …)` | `PVCopy` is the kit's one way to take copy (review M-1, X-5) | A `verbatimTitle:` init — a second pattern per component |
| D-12 | **Keys.** Kept with new text or signature: `conflictCalloutTitle` ("Some evidence disagrees"), `conflictCalloutBody` ("Disagreeing lines can't be pinned. The claim can still be filed."), `sheetNeighbor` ("a connected record"), `argument` ("Additional comments or arguments about this claim"), `outcome` (D-4), `pinRelationship` (kept; the relationship checkbox's spoken label adds the D-14 link after it). Removed with their `Localizable.xcstrings` entries: `sheetSubtitle`, `sheetStatusHint`, `claimHeading`, `through`, `ownRecords`, `pinLine`, `sheetHere`, `sheetPin`, `sheetCompared`, `sheetResult`, `sheetWeight` | One name per concept (review M-8); "High tolerance for churn"; the board has no subtitle and no field hints | Delete and re-add under new names |
| D-13 | **Life span text: one formatter, one chooser.** `PersonLifeDisplay` gains `static func span(first: CatalogDateValueInput?, last: CatalogDateValueInput?, locale: Locale = .autoupdatingCurrent) -> String` *(new)*: `dateText(first, locale:)` when `first == last`, else `dateSpan(birth: first, death: last, locale:)` ("14 May 1817 – 2 Jun 1890", "abt 1815 – 1890"; one event prints "1843"). `PromoteEvidenceDisplay.lifespanSentence(_ l: CatalogClaimEvidenceLifespan, locale: Locale = .autoupdatingCurrent) -> String` *(new)* picks the sentence: `row.missing == .contradicts` → `lifespanContradictsThisSource`; else `record.missing == .contradicts` → `lifespanContradictsRecord`; else `row.missing != .none` → `lifespanNoneThisSource`; else `record.missing != .none` → `lifespanNoneRecord`; else outcome `.conflict` → `lifespanConflict(here:there:)`; else `lifespanFits(here:there:)`, with each side's `span(first:last:)`. `.unrecognized` reads as its `.undated` sentence. The Badge is `outcome(l.outcome)`. The Lifespan Card shows exactly when `evidence.lifespan` is not nil (Go leaves it out when neither side has dates and neither contradicts itself) | The Person page and the sheet share the span and its one-date rule; the wording is one tested function, not view code (review M-4, M-11, X-13) | Calling `dateSpan` with the start only — gives "1843 –" |
| D-14 | **A relationship is named by its bridge on the page.** `PromoteFlow.BridgeFact` gains `kind: EvidenceBridgeKind`, set in `PromoteModel.facts()` from `bridge.kind`. The header's first line is the outcome sentence (`relationshipAgrees` "They share a matching connection" for `.agree`, `relationshipSimilar` "They share a similar connection" for `.partial`; `claimevidence` sends no other relationship outcome); the second is `L10n.Promote.connectionLink(first: bridge.endAName, phrase: bridge.phrase, second: bridge.endBName)` *(new)* "%1$@ · %2$@ · %3$@" from D-2's `bridge(for:…)`, else `relationshipWith(name)` "with %@" (D-9). S11-27 calls `connectionLink` for its subrows | One owner of link wording and end order, shared with S11-D1's subrows; every row kind and outcome is covered; no term or `End` reading in Swift (review M-3, X-2, X-3) | Wording from S11-32 term labels and `End` — the inverse-term and symmetric-link cases cannot be read correctly from the key |
| D-15 | **`PVCheckbox` speaks its mixed state.** With `isIndeterminate`, its accessibility value is "partly checked" from a new key `checkboxPartlyChecked` in `L10n.DesignSystem` (`macos/App/Platform/L10n.swift` line 93, beside `requiredMarker`). The sheet's spoken labels carry no state words | The control owns its state (review M-6) | State words in each sheet label |
| D-16 | **One likeness tile.** `LikenessTile<Overlay: View>(projectDir:relPath:caption:captionWidth:overlay:)` *(new)* in `macos/App/Features/Conclusions/LikenessTile.swift`: the 96 pt `CachedThumbnail(fallbackMark: .subjectPerson)`, `.accessibilityHidden(true)`, with `overlay` drawn over the crop only (default `EmptyView`), then the one-line muted caption with its `.help` and `.accessibilityHidden(true)`, `maxWidth: captionWidth`. Lifted from `ConclusionDetailLikenessSection.tileView`, which passes its open-Citation `Button` (focus ring, `.help`, label, identifier `persons.detail.likeness.<id>`) as `overlay` and `Self.column` as `captionWidth`; its portrait badge stays after the tile in its `VStack`. The sheet passes no overlay and 96 as the width. The board's italic captions are not copied: the shared tile keeps the gallery's upright caption. `projectDir` is `model.session.projectKey.projectDir` | Second call site (review M-7); the gallery's hit area, focus and VoiceOver stay as they are (review X-16); it depends on `CachedThumbnail` (Platform), so it is a feature view, not a `DesignSystem/Recipes` type | A fixed init wrapped in a button — moves the gallery's hit area and focus ring |
| D-17 | **States.** Loading: the title, its Badges and the heading without meta; three `PVSkeletonBar(width:height: 12)` as S11-25 drew. Failed: `PVCallout(tone: .danger, title: evidenceUnavailable, message: <the read's error>, compact: true)` in the evidence area; the claim fields stay. No record: `PVEmptyState(icon: .searchEmpty, title: pickRecordTitle, message: pickRecordBody, compact: true)`; claim fields shown for New, hidden for Skip and unset. Nothing to compare (loaded; no section, no relationship, no Lifespan Card): `PVEmptyState(icon: .searchEmpty, title: nothingToCompareTitle, message: nothingToCompareBody, compact: true)`. `.searchEmpty` (`text.magnifyingglass`) is the kit's file-search mark the board draws. Identifiers: `promote.sheet.marks`, `promote.sheet.heading`, `promote.sheet.section.<propertyKey>`, `promote.sheet.pin.<first pin id>`, `promote.sheet.member.<pin id>`, `promote.sheet.expand.<first pin id>`, `promote.sheet.relationship.<neighborSubjectID>.<key>` (S11-25's), `promote.sheet.lifespan`, `promote.sheet.likeness.<first pin id>`; `promote.sheet.loading`, `promote.sheet.argument` and `promote.sheet.close` stay | Board E13a – E13e; review X-12; S11-25's relationship identifier is already unique | Keys that repeat for two children |
| D-18 | **`PVPanel` gains a title accessory.** `PVPanel<Body: View, Footer: View, TitleAccessory: View>` with `@ViewBuilder titleAccessory: () -> TitleAccessory`, drawn in the title row after the title, top-aligned, with `Spacer(minLength: PVSpacing.space6)` between. Both existing inits move to an extension `where TitleAccessory == EmptyView`, so every caller compiles unchanged; a new init takes `titleAccessory:`. The subtitle parameter stays for other callers | Board E01: the Badges sit at the right end of the title. The panel owns its title row | Badges in the body under the title — the board put them in the title; an overlay — breaks with a two-line title |
| D-19 | **Claim fields first, as the board draws them.** In the sheet body, after the title's lines (D-8) and before the evidence: `HStack(alignment: .top, spacing: PVSpacing.space6)` of a 200 pt `VStack` (Status `PVField` with no hint over Confidence `PVField`) and the argument `PVField(label: L10n.Promote.argument, marker: .optional)` around `PVTextArea`. `PVField`'s `required: Bool` becomes `marker: PVField.Marker = .none` with `enum Marker { case none, required, optional }` *(new)*: `.required` draws `L10n.DesignSystem.requiredMarker` in `PVColor.danger` as today, `.optional` draws `L10n.DesignSystem.optionalMarker` *(new)* "(optional)" in `PVColor.textMuted`; every `PVField(…, required: true)` caller becomes `marker: .required` (`rg -n 'required: true' macos/App`, about a dozen, compile-checked). The "Claim" heading (`L10n.Promote.claimHeading`) goes: the board draws the fields with no heading. Hidden for Skip and unset rows | Board, all frames; S11-25 already removed the argument hint; one input with no impossible state (review M-12) | A second `optional: Bool` beside `required` |
| D-20 | **Line layout (board §9).** A line is an `HStack(alignment: .top)`: the checkbox (or, on the lifespan line, `PVIcon(.info)` in `PVColor.textFaint`, hidden from VoiceOver), the text column, the disclosure, then the Badge in a fixed 116 pt trailing column (`EvidenceLayout.badgeColumn` *(new)*, private to the sheet file). The text column: the sentence (+ muted "· N records"), then the lifespan caption in italics, then one muted detail line `L10n.Promote.lineDetail(parts:)` *(new)* joining "This source: <value>" and, for one member, "<Source>: <value>" with " · " (a collapsed line shows only "This source: <value>"; a likeness and the lifespan line have none). Members: indented to the text column, each row a `PVDivider` then checkbox, value, Source title in muted caption. Under a relationship: each section a micro-caps label (`.pvMicroCaps()`) with the property's name, indented 28 pt, with a `PVDivider` between sections; "Nothing else to compare." italic muted. The evidence intro `evidenceIntro` sits under the heading, before the conflict Callout | Board E01, E02, E05, E08; S11-D2 §9 | Two detail lines (the brief's ES-6) — the board joins them |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PromoteEvidenceSheet.swift` *(new)* and its private subviews | file, views | `PromoteView.swift` `PromoteEvidenceSheet` | Moved and rewritten (D-1) |
| `PromoteEvidenceDisplay.swift` `PromoteEvidenceDisplay` *(new)* | enum, funcs | `ReconciledValueDisplay` | Display only; pins stay in `PromoteFlow` (D-2) |
| `PromoteReadoutBadges.swift` `PromoteReadoutBadges` *(new)* | view | `PromoteAlignmentRow`'s badges, `badgeTone` | Extracted; both call sites use it (D-8) |
| `LikenessTile.swift` `LikenessTile` *(new)* | view | `ConclusionDetailLikenessSection.tileView` | Extracted; both call sites use it (D-16) |
| `PromoteModel.neighborName` *(new)* | func | `reasonText`'s lookups | Extracted (D-9) |
| `BridgeFact.kind` *(new)* | field | `SourceGraphPlacedBridge.kind` | Carried through (D-14) |
| `PVSectionHeader` generic title; `PVCheckbox` mixed value | kit extend | `PVDisclosureButton<Title: PVCopy>`; `CheckboxAccessibility` | D-11, D-15 |
| `L10n.Promote` sheet functions and keys *(new)*, `connectionLink` *(new)* | keys | `L10n.Promote.outcome`, `assessment` | Same pattern (D-4, D-12, D-14) |
| `PromoteEvidenceDisplayTests.swift` *(new)* | tests | `PromoteModelTests.swift` | Tests the display enum and L10n functions without a view |

**Builds on.** S11-25's `CatalogClaimEvidence`, `togglePin(subjectID:pinIDs:)`, `SheetEvidence`; S11-32's `lifespan`, `.depiction`, `ValueCause.unjudged`, `ValueRelation.none` and binding-order sections; `ReconciledValueDisplay`; `PersonLifeDisplay.dateSpan`; `CachedThumbnail`; `EvidenceBridgeEdgeSummary` (through `BridgeFact.phrase`); `PromoteModel.assessmentText`, `reasonText`; `PVCopy`; the kit.
**Deletes.** `sheetColumns`, `weightText`, `badgeTone`, `ExhibitGroup` (if S11-25 left it), the "Through" relationship card, `PromoteLayout.pinColumn` and `weightColumn` when `rg` finds no other reader, the D-12 removed keys.
**Next adds.** S11-27's subrows call `connectionLink`. A new relation or cause is one case in `lineSentence`. A new section kind needs no view change. A likeness on another surface is one `LikenessTile`.

### Schema

None — no catalog change.

### Go core

None — the lines and their order come from S11-25 and S11-32.

### FFI

None — no wire change.

### macOS

From the S11-D2 inventory. Paths under `macos/App/`.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Shell | `PVPanel(title:width: 880) { … } footer: { … } titleAccessory: { PromoteReadoutBadges(readout:) }` (D-18) | `Features/Promote/PromoteEvidenceSheet.swift` *(new)* | Title as today (`sheetTitleHandle` / `sheetTitleNew`); no subtitle; the Badges' tooltip is the reason (D-8) |
| Lines under the title | `EvidenceTitleLines`: text only (the Updated caption, the contested line), in a group named `whySuggested` | same | ES-2 as §9 changes it; shown only when one applies; no checkbox |
| Heading | `PVSectionHeader(title:meta:)` | same | ES-3, D-11 |
| Conflict | `PVCallout(tone: .warning, compact: true)` | same | ES-4, when `PromoteEvidenceDisplay.hasConflict` |
| Section / relationship Card | `PVCard`, header on `PVColor.surfaceSunken` as today's group header | same | ES-5, ES-12 |
| Line | `PVCheckbox(isChecked:isIndeterminate:isDisabled:accessibilitySpokenLabel:)`, `Text`, `PVDisclosureButton`, `PVBadge(subtle: true)`, `PVDivider` | same | ES-6 – ES-11 |
| Likeness line | two `LikenessTile` | same | ES-15, D-16 |
| Lifespan Card | `PVCard`, `Text`, `PVBadge` | same | ES-16, D-13, no checkbox |
| Loading / failed / empty | `PVSkeletonBar` ×3 / `PVCallout(tone: .danger)` / `PVEmptyState` | same | ES-17 – ES-20, D-17 |

States: loading, failed, no record, nothing to compare, 50 members (the member list grows; `ScrollView`).

### Strings

New, under `promote.sheet.*` (English as in S11-D2 §4; the board's copy wins where it differs):

| Function / key | English |
| --- | --- |
| `whySuggested`, `whySuggestedSpoken(assessment:reason:)` | "Why this was suggested" / "Why this was suggested: %1$@, %2$@" |
| `evidenceIntro` | "Check every line that supports this match. Checked lines are pinned to the claim when you press Done." |
| `lineDetail(parts:)` | the parts joined with " · " (one key, "%1$@ · %2$@", folded over the parts) |
| `somethingDisagrees` | "Something below disagrees with this record." |
| `evidenceHeading(ref:title:)` | "What's the evidence that this is %1$@ %2$@?" |
| `checkedCount(_:)` (plural) | "%lld checked" |
| `lineSentence(…)` | the ES-8 table (one key per cell) |
| `recordCount(_:)` (plural) | "· %lld records" |
| `showRecords(_:)` (plural), `hideRecords` | "Show %lld records" / "Hide records" |
| `thisSource(_:)`, `onSource(source:value:)`, `thisSourceCaption` | "This source: %@" / "%1$@: %2$@" / "This source" |
| `relationshipAgrees`, `relationshipSimilar`, `relationshipWith(_:)`, `connectionLink(first:phrase:second:)` | D-14 |
| `nothingUnder` | "Nothing else to compare." |
| `lifespanCaption`, `lifespanFits(here:there:)`, `lifespanConflict(here:there:)`, `lifespanNoneThisSource`, `lifespanNoneRecord`, `lifespanContradictsThisSource`, `lifespanContradictsRecord` | "Worked out from dated events. A guide for matching, never pinned." / "The lifespans fit: %1$@ here, %2$@ on the record" / "The lifespans don't fit: %1$@ here, %2$@ on the record" / "No lifespan to compare: this source has no dated events" / "No lifespan to compare: the record has no dated events" / "No lifespan for this source: its own dates contradict each other" / "No lifespan for the record: its own dates contradict each other" (D-13 picks one) |
| `pickRecordTitle`, `pickRecordBody`, `nothingToCompareTitle`, `nothingToCompareBody` | ES-19, ES-20 |
| `spokenLine(property:sentence:here:)`, `spokenLineCount(property:sentence:count:here:)` (plural: "%1$@. %2$@, %3$lld records. This source: %4$@."), `spokenMember(source:value:)`, `spokenRelationship(sentence:link:)`, `spokenLikeness(source:sentence:)` | ES-22 |

Changed and removed: D-12. `checkboxPartlyChecked` "partly checked" (D-15); `L10n.DesignSystem.optionalMarker` "(optional)" (D-19).

## 7. Steps

**STEP-0 — Refresh on landing**
Files: this plan. Skill: `plan-pr` ("Refresh an existing plan").
Do: done (2026-10-22): S11-25's, S11-32's and the board's differences are in §4 and the Decisions, and `review-plan` round 2 ran (see the log). This step is the planner's: the implementer starts at STEP-1 on a plan whose status is `Planned`.
Check: status `Planned`; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-20-claim-evidence-sheet.md` reports no errors.

**STEP-1 — Kit**
Files: `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift`, `macos/App/DesignSystem/Components/Checkbox/PVCheckbox.swift`, `macos/App/DesignSystem/Components/Panel/PVPanel.swift`, `macos/App/DesignSystem/Components/Field/PVField.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, and every `PVField(…, required: true)` caller (`rg -n 'required: true' macos/App`). Skill: `add-ui-component`.
Do: D-11, D-15, D-18, D-19's `PVField.Marker` and its callers. Each component's `#Preview` gains one case for its new input.
Check: `xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` succeeds with no change to any caller of `PVSectionHeader` or `PVPanel`, and `PVField`'s callers changed only from `required: true` to `marker: .required`; T-6 passes.

**STEP-2 — Shared pieces**
Files: `macos/App/Features/Promote/PromoteReadoutBadges.swift` *(new)*, `macos/App/Features/Promote/PromoteView.swift`, `macos/App/Features/Conclusions/LikenessTile.swift` *(new)*, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Features/Conclusions/PersonLifeDisplay.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/App/Features/Promote/PromoteFlow.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`.
Do: D-8's `PromoteReadoutBadges` (the row uses it), D-16, D-13's `PersonLifeDisplay.span`, D-9, D-14's `BridgeFact.kind`.
Check: `xcodebuild test` passes; `PromoteModelTests` `reasonsAreWordedWithLabelsAndRefs` and `movesAreWordedNamingTheNeighbor` pass unchanged; the Person gallery and the Promote rows look as before.

**STEP-3 — Display helpers**
Files: `macos/App/Features/Promote/PromoteEvidenceDisplay.swift` *(new)*, `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift` *(new)*, `project.pbxproj`. Skill: `add-swift-test`.
Do: D-2.
Check: T-1 passes.

**STEP-4 — Copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift` (new) from STEP-3. Skill: `add-localized-string`.
Do: D-4, D-5, D-12's changed keys, §6 Strings; keep the removed keys until STEP-9 so the old sheet still builds.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; T-2, T-3 pass.

**STEP-5 — Shell, title marks, claim fields, heading, states**
Files: `macos/App/Features/Promote/PromoteEvidenceSheet.swift` *(new)*, `PromoteView.swift`, `project.pbxproj`. Skill: `add-ui-component`.
Do: D-1, D-8, D-17, D-19, D-20's intro; ES-1 – ES-4, ES-17 – ES-21 as S11-D2 §9 changes them. Lines render as S11-25 left them until STEP-6.
Check: the app builds; T-4 passes; §10 steps 1 and 7.

**STEP-6 — Property Cards and lines**
Files: `PromoteEvidenceSheet.swift`.
Do: ES-5 – ES-11, D-3, D-6, D-7, D-10, D-20.
Check: T-1 and T-2 pass; §10 steps 2–3.

**STEP-7 — Relationships**
Files: `PromoteEvidenceSheet.swift`.
Do: ES-12 – ES-14, D-14.
Check: T-1's `bridgeForARelationship` and T-3 pass; §10 step 4.

**STEP-8 — Lifespan and likeness**
Files: `PromoteEvidenceSheet.swift`.
Do: ES-15, ES-16, D-10, D-13, D-16.
Check: T-3 passes; §10 steps 5–6.

**STEP-9 — Spoken labels and removal**
Files: `PromoteEvidenceSheet.swift`, `PromoteView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: ES-22 spoken labels and D-17's identifiers; delete D-12's removed keys and the code §6 Deletes names.
Check: `rg -n 'sheetColumns|weightText|badgeTone|through\(|ownRecords|sheetWeight|ExhibitGroup|promote\.sheet\.through' macos/App` prints nothing; `python3 scripts/check-localizable-xcstrings.py` passes; T-5 passes; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes.

**STEP-10 — Docs**
Files: `docs/promote-graph-alignment.md` (§3 "Assessment" bullet: the sheet opens with why the row was suggested, then the evidence as sentences; no weights), `docs/deployment-plan/spike-11/claim-evidence.md` (Status: sheet built in S11-20; Wording: settled in S11-D2, relationships named by their connection), `docs/deployment-plan/spike-11/deployment-plan.md` (map row status), `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/design/README.md` and the brief (to the spike's design archive folder), this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py` reports no errors.

Commits: one per step from STEP-1 (STEP-6 to STEP-8 may be one commit if the sheet does not build between them; say so in the PR). The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Show every record's evidence on the Promote sheet as sentences."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift` *(new)* `checkStateFollowsThePins`, `checkedCountIsOfferedPinsOnly`, `aConflictAnywhereShowsTheCallout`, `bridgeForARelationship` | `["p1"…"p5"]` with all / none / four pins → checked / unchecked / partial; `[]` → unchecked. Pins `{p1, x}` with `x` not offered → 1. A conflict only inside a relationship section → true; only a lifespan conflict → true; a partial relationship with no conflicting line → false. Two bridges between the row and the neighbor, one `participation`, one `relationship`, and a relationship of type `participation` → the participation bridge; no bridge between them → nil | AC-2.2, AC-2.3, AC-3.1, AC-3.4, D-2 |
| T-2 | Swift | same file `lineSentencesFollowTheTable` | Table-driven over every ES-8 row × name / date / term value → the row's resource key; `.negative` property cause with relation `.same` → "Disagrees" family; `.unrecognized` relation → "Similar" | AC-2.4, AC-2.5, AC-2.6, D-4 |
| T-3 | Swift | same file `relationshipAndLifespanSentences`, `countsArePlural`, `aSpanOfOneDatePrintsIt` | Outcomes agree / partial → the two D-14 sentences; `connectionLink("John Smith", "subject of", "Birth of John Smith")` → "John Smith · subject of · Birth of John Smith"; `lifespanSentence` for each case of AC-4.1 – AC-4.3, both sides missing (contradicts wins over undated, this source before the record), and `.unrecognized` → its English; `PersonLifeDisplay.span(first:last:)` with one date twice → "1843"; `recordCount(1)` / `(5)` → "· 1 record" / "· 5 records" | AC-3.1, AC-3.2, AC-3.3, AC-4.1, AC-4.2, AC-4.3, D-5, D-13, D-14 |
| T-4 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `theSheetMarksSpeakTheContestedBand` *(new)*, `theSheetCarriesTheRowsCaption` *(new)* | AC-1.1, AC-1.2's rows → the spoken "Why this was suggested: …" label, the tooltip reason and the Updated string, built through D-8's calls | AC-1.1, AC-1.2, D-8 |
| T-7 | Swift | `PromoteEvidenceDisplayTests.swift` `aLineDetailJoinsItsParts`, `anOptionalFieldSaysSo` | `lineDetail(["This source: John Smith", "1871 England Census — Leeds, Briggate: J. Smith"])` → the two joined with " · "; one part → itself; `L10n.string(L10n.DesignSystem.optionalMarker)` → "(optional)" | AC-1.5, D-19, D-20 |
| T-5 | Swift | `PromoteEvidenceDisplayTests.swift` `spokenLabelsReadTheLine` | AC-5.1's line → "Name. Same name, 5 records. This source: John Smith." (`spokenLineCount`); the lifespan line → its sentence; AC-5.3's two likeness lines → labels and checkbox labels that differ, each naming its Source and outcome | AC-5.1, AC-5.2, AC-5.3 |
| T-6 | Swift | `PromoteEvidenceDisplayTests.swift` `aPartlyCheckedBoxSaysSo` | `L10n.string` of D-15's key → "partly checked" | AC-5.1, D-15 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteFlowTests` and `PromoteModelTests` (strings, `BridgeFact`), `L10nResolutionTests` (removed keys).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without Xcode, report the Go and plan-doc steps and say Xcode was not run
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; D-12's removed keys are gone; no stale strings left
- [ ] UI matches the S11-D2 board (frames E01 – E14 as §5.3 and §9 describe them); the kit changes are D-11, D-15, D-18 and D-19 (with its callers) only
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to the design archive and design/README.md updated
- [ ] Domain doc updated: promote-graph-alignment.md §3; claim-evidence.md Status and Wording
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Data: build the project in [S11-D2 §5.2](../design/S11-D2-evidence-sheet.md#52-the-project) (the obituary of John Smith and the records it is promoted against). Each step compares the sheet with the brief's frame of the same number; every value, sentence, outcome and count must match.

1. Open Promote on the obituary; open John Smith's sheet. Skeleton bars, then Frame 1. No "Through". (AC-1.1 – AC-1.3, AC-2.1, AC-3.1, AC-4.1, AC-4.4, AC-4.5, ES-17)
2. Expand the name line; uncheck the 1861 Census; then press the line's box: Frame 2, then all four checked. (AC-2.2, AC-2.3)
3. Open Mary Smith's sheet: Frame 5 (a disagreeing member, not contested). Open the birth's sheet: Frame 6. Open Leeds's: Frame 7. (AC-2.4, AC-2.5, AC-3.2)
4. Edit a fresh copy so the obituary's marriage reads "1840"; in the marriage row's menu pick EVT-2W8NC: Frame 3; reopen John's sheet: Frame 4. Press Done with the marriage relationship left unchecked; on the person's conclusion page the marriage bundle is not pinned. (AC-1.1, AC-3.3, AC-3.4)
5. In John's menu pick PER-3X9QV: Frame 9; pick PER-5J2NB: Frame 10. (AC-4.2, AC-4.3, AC-4.5)
6. On a fresh copy set the obituary's death to "2 Jun 1809": Frame 11. (AC-4.3)
7. Set the death row to New, then Skip, then EVT-0P7XR: Frame 13's three states. (AC-1.4)
8. With VoiceOver, move through a line, a partly checked line, a member, a relationship, the lifespan line and two likeness lines. (AC-5.1, AC-5.2, AC-5.3)
9. Run `rg -n 'promote.sheet.through|Through %@' macos/App` from the repo root: it prints nothing. (AC-3.5)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A dependency lands differently from its plan | STEP-0 refresh against both Landed notes (done) |
| The board changes copy or layout | The board wins; STEP-0 copies its notes into §6 |
| Two bridges of one kind join the row and its neighbor (a person witness and subject of one event) | D-2 takes the first in `flow.bridges` order; the sentence is still true for the pinned bundle's kind |
| A record with many members makes a tall sheet | Collapsed lines and the existing `ScrollView`; S11-21 times the read |

**Stop and report if:**

- `CatalogClaimEvidence` lacks a field this plan reads (a line's `relation`, `cause`, `propertyCause`; a relationship's `bridgeType`, `outcome`; the lifespan result with each side's window and no-lifespan reason; the depiction case).
- Sections do not arrive in the kind's binding order (on a new catalog: person `name`, `likeness`, `sex_at_birth`, …).
- Showing a line needs a Go or wire change.
- `togglePin(subjectID:pinIDs:)` does not insert every id when some are already in `pins`.
- The board draws a control the S11-D2 inventory does not list.
- Making `PVSectionHeader` generic changes how an existing caller renders, or does not compile without editing a caller.

## 12. Open questions

None. The board reaches the implementer as S11-D2 §5.3 (every frame's content) and §9 (its layout), with the link on **Board** (user, 2026-10-22).

## 13. PR

**Title:** S11-20 — Claim evidence sheet
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-20-claim-evidence-sheet.md`.
