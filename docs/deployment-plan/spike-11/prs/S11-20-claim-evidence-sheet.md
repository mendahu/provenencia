# S11-20 — Claim evidence sheet

**Status:** Blocked — on S11-25, S11-32, board S11-D2
**Spike / slice:** Spike 11 · Slice 5 — Claim evidence
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S11-D2](../design/S11-D2-evidence-sheet.md), board not drawn
**Board:** — (to be recorded when the board is agreed: link, frame numbers, and S11-D2 §8 board notes)
**Depends on:** S11-25 (in progress; this draft follows [its plan](S11-25-claim-evidence-rpc.md)), S11-32 (scoped; this draft follows its [scope entry](../deployment-plan.md#s11-32--lifespan-and-likeness-evidence-lines)), S11-D2
**Blocks:** —
**Branch:** `s11-20-claim-evidence-sheet`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/design-system-layers.mdc`
**Context:** [S11-D2](../design/S11-D2-evidence-sheet.md) (requirements ES-1 – ES-22, copy table, inventory); [`claim-evidence.md`](../claim-evidence.md) (Lines, Wording, Evidence through relationships); [S11-D1](../design/S11-D1-promote-rows.md) (band, Contested, move captions, the "·" connection form); [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3; working rules "The pin policy has one owner, in Go", "One wire vocabulary"; S11-18 review M-2, M-6 (subtitle and neighbor naming); S11-12 D-12 (a side that contradicts itself)
**Planned:** 2026-10-22 against `main` @ `6504003` (draft: S11-25 and S11-32 not landed)
**Review:** — (not reviewed yet: review-plan runs when the plan is refreshed on S11-25 and S11-32's landing)

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

The Promote evidence sheet answers two questions and keeps them apart. A summary at the top says why the row was suggested: its band, Contested when it applies, and the row's reason. Under it, "What's the evidence that this is <record>?" lists every member's evidence as sentences the researcher checks or unchecks: property sections, collapsed lines that expand to each record, relationships that say what matched, a lifespan line that is never pinned, and likenesses as two crops side by side. The table with its Weight column and the "Through <name>" cards go.

**Why now.** S11-25 moved the sheet onto the evidence RPC in today's shape; S11-32 adds the lifespan and likeness lines. This PR is the redraw S11-D2 specifies, and closes the dogfood bar item "the evidence sheet shows all the evidence".

**What stays.** The data: `CatalogClaimEvidence` and every Go path (no wire change). Pins as S11-25 left them: `PromoteFlow.Row.pins`, `togglePin(subjectID:pinIDs:)`, `setEvidence`, `defaultPins`, the sheet's lazy load and `SheetEvidence` states. The panel, its title, the claim fields, the footer. The Promote rows.

## 2. User stories

Conventions. "Evidence E1" is a `CatalogClaimEvidence` with one `name` section holding one line: relation `same`, outcome `agree`, pinnable, checked, `incoming` the name "John Smith", five members (pin ids `p1`…`p5`) on Sources "Census 1851" … "Census 1891", each value "John Smith". "Row R" is a `PromoteFlow.Row` named "John Smith" with target `.handle(id: "h1", ref: "PER-7QH3D", title: "John Smith")`, assessment `.strong`, reason `.via` a neighbor "Mary Smith" whose `shownRef` is "PER-4KX2M".

**US-1.** As a researcher, when I open a row's sheet, I first read why it was suggested, then the evidence, so that I never mistake matching's reasons for the pins I file.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Row R, contested | Subtitle "Strong, contested · via Mary Smith → PER-4KX2M"; the summary Card holds the Strong and Contested Badges, the reason, and "Something below disagrees with this record." (ES-1, ES-2) |
| AC-1.2 | Row R, not contested, with move `.neighborMoved` | Subtitle "Strong · after Mary Smith moved to PER-4KX2M"; no "Something below disagrees" line; the summary shows "Updated · after Mary Smith moved to PER-4KX2M" only when the row is `updated` (ES-2) |
| AC-1.3 | Row R with E1, all five pins checked | Heading "What's the evidence that this is PER-7QH3D John Smith?" with meta "5 checked" (ES-3) |
| AC-1.4 | A row set to New, Skip or unset | No subtitle; the EmptyState "Pick a record to see its evidence"; claim fields shown for New (ES-19) |

**US-2.** As a researcher, I see every member's evidence as a sentence, and I can uncheck one record of a collapsed line.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Row R with E1 | A "Name" Card with one line "Same name · 5 records", Badge Agrees, box checked, "Show 5 records" (ES-5, ES-6, ES-9) |
| AC-2.2 | AC-2.1, expand, uncheck Census 1861 | Five member rows; Census 1861 unchecked; the line's box partly checked; pins `p1, p3, p4, p5`; heading meta "4 checked" (ES-9, ES-10) |
| AC-2.3 | AC-2.2, press the line's box | All five checked (ES-10) |
| AC-2.4 | A `date` line `contains` / `precision` | "Same date, recorded less precisely", Badge Similar, box unchecked and enabled (ES-8) |
| AC-2.5 | A `sex_at_birth` line whose incoming is negative "male", outcome `conflict` | "Disagrees", Badge Conflicts, box disabled; "This source: not male"; the conflict Callout under the heading (ES-4, ES-7) |
| AC-2.6 | Every relation × cause in the ES-8 table, for a name value, a date value, and a term value | The table's sentence (T-2) |

**US-3.** As a researcher, a relationship says what matched, and checking it pins what is behind it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A relationship `participation` / term `subject` / event type `birth`, outcome `agree`, checked, one bundle, one `date` section | A Card headed "They have a matching birth event" with the neighbor's `shownRef`, Badge Agrees, box checked; the date line under it, indented (ES-12, ES-13) |
| AC-3.2 | A `relationship` key, term label "parent of", the row Subject at end A, neighbor "Ann Gates" | "They have a matching relationship: John Smith · parent of · Ann Gates"; at end B: "…: Ann Gates · parent of · John Smith" (ES-13) |
| AC-3.3 | A partial relationship (outcome `partial`) | "They have a similar birth event", box unchecked (ES-13) |
| AC-3.4 | A `location` relationship with no sections | "It happened at a matching place: Toronto" and "Nothing else to compare on this event."; pressing its box toggles its bundle ids (ES-14) |
| AC-3.5 | No sheet string reads "Through" | `rg -n 'promote.sheet.through\|Through %@' macos/App` prints nothing |

**US-4.** As a researcher, I see a person's lifespan check and their likenesses.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Lifespan agree, here 1817–1880, there about 1815–1890 | A "Lifespan" Card, last among own Cards: "The lifespans fit: 1817–1880 here, about 1815–1890 on the record", no checkbox, the caption "Worked out from dated events. A guide for matching, never pinned." (ES-16) |
| AC-4.2 | Lifespan conflict | "The lifespans don't fit", Badge Conflicts (ES-16) |
| AC-4.3 | This side contradicts itself | "No lifespan for this source: its own dates contradict each other"; the record side: "No lifespan for the record: its own dates contradict each other" (ES-16) |
| AC-4.4 | A `likeness` line, same image | A "Likeness" Card after Name; two 96 pt crops with Source titles; "Same likeness", Agrees, checked (ES-15) |
| AC-4.5 | A `likeness` line, two images | "Compare these likenesses", Badge "Judge by eye", unchecked, enabled (ES-15) |

**US-5.** As a VoiceOver user, each line reads as one sentence with its state.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | AC-2.1's line | "Name. Same name, 5 records. This source: John Smith. Checked." (ES-22) |
| AC-5.2 | The lifespan line | Its sentence only; no checkbox element (ES-22) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PVSectionHeader` `verbatimTitle:` init (S11-D2 inventory, Extend) | STEP-1 |
| Pure presentation: check state, sections order, sentences' inputs, counts (D-2, D-3) | STEP-2, AC-2.2, AC-2.3 |
| Copy: ES-8 table, relationship and lifespan sentences, summary, heading, states, spoken labels; removed keys (D-4, D-5) | STEP-3, AC-2.6 |
| Sheet shell, subtitle, summary, heading, states (ES-1 – ES-4, ES-17 – ES-21) | STEP-4, US-1 |
| Property Cards, lines, collapse (ES-5 – ES-11) | STEP-5, US-2 |
| Relationship Cards (ES-12 – ES-14) | STEP-6, US-3 |
| Lifespan and likeness Cards (ES-15, ES-16) | STEP-7, US-4 |
| Accessibility (ES-22), removal of the old table | STEP-8, US-5, AC-3.5 |
| Docs | STEP-9 |

| Out | Home |
| --- | --- |
| Any Go, proto or store change | S11-25, S11-32 (a missing field is a Stop condition) |
| Opening the sheet from a connection subrow | not planned (decided here for S11-D1's note) |
| Select-all / pin-nothing | not planned (S11-D2 §2.1) |
| Face comparison of two likenesses | not planned |
| Statuses beyond Accepted | a later spike (today's hint) |
| Ordering relationships beyond what Go sends | not planned (D-3: Go's order puts birth and death first) |
| Wording the old reconciler `outcome*` keys in `L10n.Conclusions` | not planned (other view) |

## 4. Current state

Verified at `6504003`, then read against S11-25's plan for the rows it changes. Rows marked **after S11-25** or **after S11-32** describe code those PRs write; recheck them at your base.

| Area | Where | Today |
| --- | --- | --- |
| Sheet | `macos/App/Features/Promote/PromoteView.swift` `PromoteEvidenceSheet` (lines 664–896), `PromoteView.body` `.sheet(item: sheetBinding)` (lines 68–70), `PromoteLayout` (lines 900–910) | A `PVPanel(width: 880)` with a conflict callout, `sheetColumns` (Pin, Compared, This source, target, Result, Weight), one `PVCard` per `ExhibitGroup` ("This subject", "Through <name>"), `comparisonLine`, claim fields, "Back to all rows". `sheetSubtitle` uses `L10n.Promote.assessment(readout.assessment)`, not `assessmentText`; group refs use `row.ref`, not `shownRef` |
| Sheet after S11-25 | same file | Reads `row.evidence: CatalogClaimEvidence?` and `model.sheetEvidence`; `evidenceLine` (renamed from `comparisonLine`) binds a `PVCheckbox` to a line's pin ids; relationship cards titled `through(name)` with a header checkbox; skeleton bars while loading; a danger callout on failure (S11-25 D-14) |
| Evidence types after S11-25 | `macos/App/Platform/GenealogyStore.swift` `CatalogClaimEvidence{sections, relationships, offers, sourceTitles}`, `checkedPinIDs`, `offeredPinIDs`, `CatalogClaimEvidenceSection{propertyKey, lines}`, `CatalogClaimEvidenceLine{propertyKey, outcome, propertyCause, relation, cause, points, pinnable, checked, incomingObservationID, incoming, incomingNegative, members}`, `CatalogClaimEvidencePin{id, memberSubjectID, observationID, sourceID, value}`, `CatalogClaimEvidenceRelationship{neighborSubjectID, key, bridgeType, term, end, eventType, outcome, cause, points, pinnable, checked, bundles, sections}`, `CatalogClaimEvidenceBundle{id, memberSubjectID, sourceID}` | S11-25 D-11 |
| Vocabulary after S11-25 | `macos/App/Platform/ConclusionVocabulary.swift` `EvidenceOutcome`, `EvidencePropertyCause`, `ValueRelation`, `ValueCause`, `SegmentCause` | S11-25 D-10; `L10n.Promote.outcome(_: EvidenceOutcome)` |
| Added by S11-32 | `CatalogClaimEvidence` lifespan field; relationship `eventTypeLabel`, `termLabel`; `CatalogConclusionValue.depiction(CatalogDepictionValue)`; a `ValueCause` case for "cannot judge two images" | Per the S11-32 scope entry. Exact names come from S11-32's plan; STEP-0 records them here |
| Flow | `macos/App/Features/Promote/PromoteFlow.swift` `Row` (`name`, `ref`, `shownRef`, `target`, `selectionReadout`, `captionReason`, `updated`, `move`, `contested`, `pins`), `row(_:)`, `sheetSubjectID` | `shownRef` is `target.handleRef ?? ref` |
| Model | `macos/App/Features/Promote/PromoteModel.swift` `reasonText(_:on:)` (lines 200–235), `assessmentText(_:)` (lines 238–241), `propertyLabel(key:origin:)` (lines 244–246), `openSheet`, `closeSheet`, `togglePin` | `assessmentText` gives "Strong, contested" |
| Values to text | `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `string(for:locale:orders:)` | S11-25 words line values with it |
| Crops | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail(projectDir:relPath:size:fallbackMark:imageLabel:)`; `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailLikenessSection` (96 pt crops, line 270) | The Person gallery's crop; the precedent for ES-15 |
| Kit | `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift` `init(title: LocalizedStringResource, meta:aside:actions:)`; `macos/App/DesignSystem/Components/EmptyState/PVEmptyState.swift` (`init(icon:verbatimTitle:message:compact:)` at line 116 is the verbatim precedent); `PVCheckbox` (`isIndeterminate:`, `accessibilitySpokenLabel:`); `PVDisclosureButton(_:isExpanded:)`; `PVBadge`; `PVCallout`; `PVCard(tone: .sunken)`; `PVSkeletonBar` | `PVSectionHeader` has no verbatim title |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Promote` (lines 1297–1499): `sheetSubtitle`, `through`, `ownRecords`, `sheetNeighbor`, `pinLine`, `negatedValue`, `conflictCalloutTitle`, `conflictCalloutBody`, `argumentHint`, `sheetHere`, `sheetPin`, `sheetCompared`, `sheetResult`, `sheetWeight`, `outcome`; `macos/App/Resources/Localizable.xcstrings` | Today's sheet strings |
| Tests | `macos/ProvenenciaTests/PromoteModelTests.swift` `theContestedBandIsSpoken`, `reasonsAreWordedWithLabelsAndRefs`; `macos/ProvenenciaTests/PromoteFlowTests.swift` (S11-25's T-10 tests); `macos/ProvenenciaTests/L10nResolutionTests.swift` | No test renders the sheet; there is no UI-test target |

**Precedent.** S11-18 ([#507](https://github.com/mendahu/provenencia/pull/507)) built S11-D1's frames on the same page: kit Badges and Callouts composed in `PromoteView.swift`, copy through `L10n.Promote`, wording tested on `PromoteModel` (`theContestedBandIsSpoken`). Copy that: words in `L10n` functions that switch on wire enums (`L10n.Promote.outcome`, `assessment`), choices in pure Swift tested without a view.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The sheet moves to its own file.** `macos/App/Features/Promote/PromoteEvidenceSheet.swift` *(new)* holds `PromoteEvidenceSheet` and its private subviews: `EvidenceSummary`, `EvidenceSectionCard`, `EvidenceLineRow`, `EvidenceMemberRow`, `EvidenceRelationshipCard`, `EvidenceLikenessLine`, `EvidenceLifespanCard` *(all new, private)*. `PromoteView.swift` keeps only the `.sheet` call | The rethink replaces the sheet wholesale; `PromoteView.swift` is 927 lines | Keeping it in `PromoteView.swift` — the file mixes rows and sheet |
| D-2 | **Pure presentation in one place.** `macos/App/Features/Promote/PromoteEvidencePresentation.swift` *(new)*: `enum PromoteEvidence` with `enum CheckState { case checked, unchecked, partial }`; `static func checkState(pinIDs: [String], pins: Set<String>) -> CheckState` (all in → checked, none → unchecked, else partial; empty `pinIDs` → unchecked); `static func checkedCount(_ e: CatalogClaimEvidence, pins: Set<String>) -> Int` (`pins ∩ e.offeredPinIDs`); `static func ownSections(_ e:) -> [CatalogClaimEvidenceSection]` (every section but `likeness`, in Go order, then `likeness` moved to index 1 when a `name` section is first, else to index 0); `static func hasConflict(_ e:) -> Bool` (any line in any section, or any relationship section, with outcome `.conflict`); `static func relationshipEnds(_ r:, rowName:, neighborName:) -> (String, String)` (row first when `r.end` is the A end, else neighbor first) | One testable home for every choice the view makes; the view only lays out | Logic inside SwiftUI bodies — untestable (no UI target) |
| D-3 | **Order comes from Go.** Sections and relationships render in the order `CatalogClaimEvidence` gives (S11-19 D-11: one-cardinality keys, birth and death, first). The lifespan Card is last among own Cards; likeness follows D-2 | S11-D2's "birth and death first" is Go's order already | A Swift rank table — a second owner of order |
| D-4 | **Words through `L10n.Promote`, switching on wire enums.** `lineSentence(relation: ValueRelation, cause: ValueCause, propertyCause: EvidencePropertyCause, outcome: EvidenceOutcome, value: CatalogConclusionValue) -> LocalizedStringResource` *(new)* implements the ES-8 table: the `value` case picks the name, date or generic variant; `propertyCause == .negative` or `relation == .different` → the "Disagrees" family; S11-32's cannot-judge cause → "Compare these likenesses"; `.same` on a depiction → "Same likeness". `outcomeBadge(_ outcome: EvidenceOutcome, cause: ValueCause) -> LocalizedStringResource` *(new, replaces `outcome(_:)`)*: Agrees, Similar, Conflicts, "Judge by eye" for the cannot-judge cause, else "Not compared". `relationshipSentence(...)`, `lifespanSentence(...)`, the summary, heading and state strings are format functions per §6 Strings | The precedent (`outcome`, `assessment`); "One wire vocabulary": codes parsed once, worded once | A `Wording` type in the feature folder — a second home for copy |
| D-5 | **Counts are plural keys.** "· %lld records", "%lld checked", "Show %lld records", the argument hint: one key each with `plural` variations in `Localizable.xcstrings` | `add-localized-string` §4 | `count == 1` branches |
| D-6 | **Expansion is view state.** `@State private var expanded: Set<String>` on `PromoteEvidenceSheet`, keyed by a line's first pin id; reset when the sheet closes | Not part of the claim; S11-D2 does not keep it | Storing it on `PromoteFlow.Row` |
| D-7 | **Toggles reuse S11-25.** A line's box calls `model.togglePin(subjectID:pinIDs:)` with every pin id (partial → all in, ES-10, which S11-25's "else insert all" already does); a member row calls it with its one id; a relationship's box with its bundle ids. A non-pinnable line has a disabled box; the lifespan line has none | One pin owner; no new flow API | A per-member toggle function |
| D-8 | **Subtitle and summary use the row's words.** Subtitle: `L10n.Promote.sheetSubtitle(assessment: model.assessmentText(readout), reason: model.reasonText(row.captionReason, on: row))`. The summary Card shows `readout.assessment` and `readout.contested` Badges as `PromoteAlignmentRow` does (`badgeTone`, moved from `PromoteView` to `PromoteLayout` as a `static func` so both files call it), `reasonText(readout.reason, on: row)`, and `L10n.Promote.updatedDetail(reasonText(move, on: row))` when `row.updated && row.move != nil` | S11-18 review M-2 (`assessmentText`); the row and the sheet say the same thing | Re-deriving the band text |
| D-9 | **Neighbors are named as rows are.** A relationship's neighbor name is `model.flow.row(neighborSubjectID)?.name`, else `L10n.Promote.sheetNeighborFallback` ("a connected record"); its ref is that row's `shownRef`, else none | S11-18 review M-6 | `row.ref` |
| D-10 | **Values.** A line's and a member's value text is `ReconciledValueDisplay.string(for:orders: model.session.nameFormat)` as S11-25 does; a negative incoming is `negatedValue`. A depiction value draws `CachedThumbnail(projectDir: model.projectDir, relPath: depiction.imageRelPath, size: 96, fallbackMark: .subjectPerson, imageLabel: …)`; an empty `imageRelPath` draws the mark | One value-to-text path; the gallery's crop | A local image loader |
| D-11 | **The heading uses the kit.** `PVSectionHeader` gains `init(verbatimTitle: String, meta: String? = nil, aside:actions:)` *(new)*, the same shape as `PVEmptyState`'s verbatim init; the existing init calls through it. The heading is `PVSectionHeader(verbatimTitle: L10n.Promote.evidenceHeading(ref:title:), meta: L10n.Promote.checkedCount(n))` | S11-D2 inventory (Extend); `add-localized-string` forbids interpolating into a resource | A local `Text` heading |
| D-12 | **Removed keys.** `through`, `ownRecords`, `sheetNeighbor`, `pinLine`, `sheetHere`, `sheetPin`, `sheetCompared`, `sheetResult`, `sheetWeight`, `conflictCalloutTitle`, `conflictCalloutBody`, `outcome(_:)` and their `Localizable.xcstrings` entries go. `sheetSubtitle` and `argumentHint` keep their keys with new text | "High tolerance for churn"; no dead strings | Leaving them for a later cleanup |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PromoteEvidenceSheet.swift` *(new)* and its private subviews | file, views | `PromoteView.swift` `PromoteEvidenceSheet` | Moved and rewritten (D-1) |
| `PromoteEvidencePresentation.swift` `PromoteEvidence` *(new)* | enum, funcs | `PromoteFlow` drafting helpers | Pins stay in `PromoteFlow`; this is display only (D-2) |
| `PVSectionHeader.init(verbatimTitle:…)` *(new)* | component init | `PVEmptyState.init(icon:verbatimTitle:…)` | Extends the kit piece the brief names (D-11) |
| `L10n.Promote` sheet functions and keys *(new)* | keys | `L10n.Promote.outcome`, `assessment` | Same pattern; old keys deleted (D-4, D-12) |
| `PromoteEvidenceSheetTests.swift` *(new)* | tests | `PromoteModelTests.swift` | Tests `PromoteEvidence` and the L10n functions without a view |

**Builds on.** S11-25's `CatalogClaimEvidence`, `togglePin(subjectID:pinIDs:)`, `SheetEvidence`; S11-32's lifespan, labels and depiction case; `ReconciledValueDisplay`; `CachedThumbnail`; `PromoteModel.assessmentText`, `reasonText`; the kit.
**Deletes.** `sheetColumns`, `weightText`, `ExhibitGroup` (if S11-25 left it), the "Through" relationship card, `PromoteLayout.pinColumn` and `weightColumn` if no row code reads them (check with `rg`), the D-12 keys.
**Next adds.** A new relation or cause is one case in `lineSentence`; a new section kind needs no view change; a likeness on another kind is the same `CachedThumbnail` line.

### Schema

None — no catalog change.

### Go core

None — the lines come from S11-25 and S11-32.

### FFI

None — no wire change.

### macOS

From the S11-D2 inventory. Paths under `macos/App/`.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Shell | `PVPanel(title:subtitle:width: 880)` | `Features/Promote/PromoteEvidenceSheet.swift` *(new)* | Title as today (`sheetTitleHandle` / `sheetTitleNew`); subtitle D-8 |
| Summary | `PVCard(tone: .sunken)` + `PVBadge` | same | ES-2; no checkbox |
| Heading | `PVSectionHeader(verbatimTitle:meta:)` | same | ES-3, D-11 |
| Conflict | `PVCallout(tone: .warning, compact: true)` | same | ES-4, when `PromoteEvidence.hasConflict` |
| Section / relationship Card | `PVCard`, header on `PVColor.surfaceSunken` as today's group header | same | ES-5, ES-12 |
| Line | `PVCheckbox(isChecked:isIndeterminate:isDisabled:accessibilitySpokenLabel:)`, `Text`, `PVDisclosureButton`, `PVBadge(subtle: true)`, `PVDivider` | same | ES-6 – ES-11 |
| Likeness line | two `CachedThumbnail(size: 96)` | same | ES-15 |
| Lifespan Card | `PVCard`, `Text`, `PVBadge` | same | ES-16, no checkbox |
| Loading / failed | `PVSkeletonBar` ×3 / `PVCallout(tone: .danger)` | same | ES-17, ES-18, from `model.sheetEvidence` |
| Empty | `PVEmptyState(icon:title:message:compact: true)` | same | ES-19 (target not a handle), ES-20 (loaded, no sections, no relationships, no lifespan) |

States: loading, failed, no record, nothing to compare, 50 members (the member list grows; `ScrollView`). Identifiers: `promote.sheet.summary`, `promote.sheet.heading`, `promote.sheet.section.<propertyKey>`, `promote.sheet.pin.<first pin id>`, `promote.sheet.member.<pin id>`, `promote.sheet.expand.<first pin id>`, `promote.sheet.relationship.<key>`, `promote.sheet.lifespan`, `promote.sheet.likeness.<pin id>`.

### Strings

New, under `promote.sheet.*` (English as in S11-D2 §4; the board's copy wins where it differs):

| Function / key | English |
| --- | --- |
| `whySuggested` | "Why this was suggested" |
| `somethingDisagrees` | "Something below disagrees with this record." |
| `evidenceHeading(ref:title:)` | "What's the evidence that this is %1$@ %2$@?" |
| `checkedCount(_:)` (plural) | "%lld checked" |
| `disagreeTitle`, `disagreeBody` | "Some evidence disagrees" / "Disagreeing lines can't be pinned. The claim can still be filed." |
| `lineSentence(…)` | the ES-8 table (one key per cell) |
| `recordCount(_:)` (plural) | "· %lld records" |
| `showRecords(_:)` (plural), `hideRecords` | "Show %lld records" / "Hide records" |
| `thisSource(_:)`, `onSource(source:value:)` | "This source: %@" / "%1$@: %2$@" |
| `outcomeBadge(…)` | "Agrees", "Similar", "Conflicts", "Judge by eye", "Not compared" |
| `relationshipEvent(type:)`, `relationshipEventRole(type:role:)`, `relationshipLink(first:term:second:)`, `relationshipPlace(_:)` and their "similar" forms | ES-13 |
| `nothingUnder` | "Nothing else to compare on this event." |
| `sheetNeighborFallback` | "a connected record" |
| `likenessSection` | "Likeness" |
| `lifespanSection`, `lifespanCaption`, `lifespanFits(here:there:)`, `lifespanConflict(here:there:)`, `lifespanNone(side:)`, `lifespanContradicts(side:)`, `sideThisSource`, `sideRecord` | ES-16 |
| `pickRecordTitle`, `pickRecordBody`, `nothingToCompareTitle`, `nothingToCompareBody` | ES-19, ES-20 |
| `spokenLine(…)`, `spokenMember(…)`, `spokenRelationship(…)`, `spokenLikeness(…)` | ES-22 |
| `sheetSubtitle` (text changed) | "%1$@ · %2$@" |
| `argumentHint(pinned:)` (text changed, plural) | "%lld pieces of evidence checked · say why these are the same" |

Removed: D-12.

## 7. Steps

**STEP-0 — Refresh on landing**
Files: this plan. Skill: `plan-pr` ("Refresh an existing plan").
Do: when S11-25 and S11-32 have landed and the board is agreed, record their **Landed** notes' differences in §4, S11-32's exact names in the "Added by S11-32" row, the board's frames on **Board**, and run `review-plan`. This step is the planner's, not the implementer's: the implementer starts at STEP-1 on a plan whose status is `Planned`.
Check: status `Planned`; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-20-claim-evidence-sheet.md` reports no errors.

**STEP-1 — Kit: verbatim section header**
Files: `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift`. Skill: `add-ui-component`.
Do: D-11's init.
Check: `xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` succeeds; existing callers unchanged.

**STEP-2 — Presentation helpers**
Files: `macos/App/Features/Promote/PromoteEvidencePresentation.swift` *(new)*, `macos/ProvenenciaTests/PromoteEvidenceSheetTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-2.
Check: T-1 passes.

**STEP-3 — Copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PromoteEvidenceSheetTests.swift` (new) from STEP-2. Skill: `add-localized-string`.
Do: D-4, D-5 and §6 Strings; keep the D-12 keys until STEP-8 so the old sheet still builds.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; T-2, T-3 pass.

**STEP-4 — Shell, summary, heading, states**
Files: `macos/App/Features/Promote/PromoteEvidenceSheet.swift` *(new)*, `macos/App/Features/Promote/PromoteView.swift`. Skill: `add-ui-component`.
Do: D-1, D-8, D-11; ES-1 – ES-4, ES-17 – ES-21; move `badgeTone` to `PromoteLayout`. Lines render as S11-25 left them until STEP-5.
Check: the app builds; T-4 passes; §10 step 1.

**STEP-5 — Property Cards and lines**
Files: `PromoteEvidenceSheet.swift`.
Do: ES-5 – ES-11, D-6, D-7, D-10.
Check: §10 steps 2–3.

**STEP-6 — Relationships**
Files: `PromoteEvidenceSheet.swift`.
Do: ES-12 – ES-14, D-9.
Check: §10 step 4.

**STEP-7 — Lifespan and likeness**
Files: `PromoteEvidenceSheet.swift`.
Do: ES-15, ES-16, D-10.
Check: §10 steps 5–6.

**STEP-8 — Accessibility and removal**
Files: `PromoteEvidenceSheet.swift`, `PromoteView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: ES-22 spoken labels and the §6 identifiers; delete the D-12 keys and the code §6 Deletes names.
Check: `rg -n 'sheetColumns|weightText|through\(|ownRecords|sheetWeight|ExhibitGroup|promote\.sheet\.through' macos/App` prints nothing; `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes.

**STEP-9 — Docs**
Files: `docs/promote-graph-alignment.md` (§3 "Assessment" bullet: the sheet opens with why the row was suggested, then the evidence as sentences; no weights), `docs/deployment-plan/spike-11/claim-evidence.md` (Status: sheet built in S11-20; Wording: "settled in S11-D2"), `docs/deployment-plan/spike-11/deployment-plan.md` (map row status), `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/design/README.md` and the brief (to the spike's design archive folder), this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py` reports no errors.

Commits: one per step from STEP-1 (STEP-5 to STEP-7 may be one commit if the sheet does not build between them; say so in the PR). The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Show every record's evidence on the Promote sheet as sentences."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PromoteEvidenceSheetTests.swift` *(new)* `checkStateFollowsThePins`, `checkedCountIsOfferedPinsOnly`, `likenessFollowsName`, `aConflictAnywhereShowsTheCallout`, `relationshipEndsReadAsTheLinkReads` | `["p1"…"p5"]` with all / none / four pins → checked / unchecked / partial; `[]` → unchecked. Pins `{p1, x}` with `x` not offered → 1. Sections `[date, likeness, name]` from Go order `name` first → `[name, likeness, date]`. A conflict only inside a relationship section → true. End A → (row, neighbor); end B → (neighbor, row) | AC-2.2, AC-2.3, AC-3.2, D-2, D-3 |
| T-2 | Swift | same file `lineSentencesFollowTheTable` | Table-driven over every ES-8 row × name / date / term value → the row's resource key; `.negative` property cause with relation `.same` → "Disagrees" family; an unrecognized relation → the generic "Similar" family | AC-2.4, AC-2.5, AC-2.6, D-4 |
| T-3 | Swift | same file `relationshipAndLifespanSentences`, `countsArePlural` | AC-3.1 – AC-3.3's keys and labels → their English; each lifespan case of AC-4.1 – AC-4.3 → its English; `recordCount(1)` / `(5)` → "· 1 record" / "· 5 records" | AC-3.1, AC-3.2, AC-3.3, AC-4.1, AC-4.2, AC-4.3, D-5 |
| T-4 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `theSheetSubtitleSpeaksTheContestedBand` *(new)*, `aMovedRowsSubtitleIsTheMove` *(new)* | AC-1.1, AC-1.2's rows → the subtitle strings, built through the functions D-8 names | AC-1.1, AC-1.2, D-8 |
| T-5 | Swift | `PromoteEvidenceSheetTests.swift` `spokenLabelsReadTheLine` | AC-5.1's line → its spoken label; the lifespan line → its sentence | AC-5.1, AC-5.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteFlowTests` and `PromoteModelTests` (strings), `L10nResolutionTests` (removed keys).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; the D-12 keys are gone; no stale strings left
- [ ] UI matches the S11-D2 inventory and board; the only kit change is PVSectionHeader's verbatim init
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc updated: promote-graph-alignment.md §3; claim-evidence.md Status and Wording
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Data: S11-25's manual-check project (a filed person "John Smith" on two Sources with a birth in 1817, a third Source with "John Smith", a birth in 1817 and a subject participation), plus on the third Source a likeness on John drawn from a new image, and on one earlier Source a likeness drawn from the same image.

1. Open Promote on the third Source; open John's sheet. Skeleton bars, then the summary (band, reason), the heading with "N checked", no "Through". Set a sibling row to Skip: the subtitle and summary still match the row's caption. (AC-1.1 – AC-1.3, ES-17)
2. The Name Card: "Same name · 2 records", checked. Expand; uncheck one record: the line is partly checked and the heading count drops by one. Press the line's box: both checked. (AC-2.1 – AC-2.3)
3. Find or make a date line recorded less precisely ("ABT 1817" against "1817"): "Same date, recorded less precisely", unchecked. (AC-2.4)
4. The birth Card: "They have a matching birth event" with the ref of the birth row, checked, the date line under it. Uncheck it, press Done; on the person's conclusion page the bundle is not pinned. (AC-3.1, AC-3.4)
5. The Likeness Card after Name: two crops per line, "Same likeness" checked for the shared image, "Compare these likenesses" unchecked for the other. (AC-4.4, AC-4.5)
6. The Lifespan Card last: "The lifespans fit …", no checkbox. Edit a fresh copy so the third Source's John dies in 1810: reopen; "No lifespan for this source: its own dates contradict each other". (AC-4.1, AC-4.3)
7. Set John's row to New: the sheet shows "Pick a record to see its evidence". (AC-1.4)
8. With VoiceOver, move through a line, a member, a relationship and the lifespan line. (AC-5.1, AC-5.2)
9. Run `rg -n 'promote.sheet.through|Through %@' macos/App` from the repo root: it prints nothing. (AC-3.5)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The draft is written before S11-25 and S11-32 land | STEP-0 refresh; Current state rows marked; status Blocked |
| The board changes copy or layout | The board wins; STEP-0 copies its notes into §6 |
| A record with many members makes a tall sheet | Collapsed lines and the existing `ScrollView`; S11-21 times the read |

**Stop and report if:**

- `CatalogClaimEvidence` lacks a field this plan reads (a line's `relation`, `cause`, `propertyCause`, a relationship's `end`, labels, the lifespan result, the depiction case).
- Showing a line needs a Go or wire change.
- `togglePin(subjectID:pinIDs:)` does not insert every id when some are already in `pins`.
- The board draws a control the S11-D2 inventory does not list.
- `PVSectionHeader`'s verbatim init would change how its existing callers render.

## 12. Open questions

- **How the agreed board reaches the implementer** (the S11-D1 board link, exported frames in the repo, or screenshots). Recommended: the Claude Design link plus a §8 "Board notes" table in S11-D2, as S11-D1 did. Answer when the board is agreed.

## 13. PR

**Title:** S11-20 — Claim evidence sheet
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-20-claim-evidence-sheet.md`.
