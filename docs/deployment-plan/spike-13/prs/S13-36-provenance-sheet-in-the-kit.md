# S13-36 — The provenance sheet joins the design system

**Status:** Planned
**Spike / slice:** Spike 13 · Slice 5 — Gallery (preparatory)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D13](../design/S13-D13-provenance-sheet-in-the-kit.md), board drawn in the kit and confirmed by the user 2026-11-03
**Board:** Claude Design project [*Provenencia Design System*](https://claude.ai/design/p/0f6c1f29-8408-4735-b114-5ae4c4c18445?via=share): `recipes/provenance-sheet.card.html`, `provenance-sheet-variants.card.html`, `provenance-sheet-levels.card.html`, `provenance-sheet-states.card.html`, and `recipes/PVProvenance*`. The board's names and slots bind. Its measures and marks do not: where it differs from the shipping sheet, the shipping sheet wins (D-11). The implementer may not be able to open it: **Design → macOS** gives every measure
**Depends on:** S13-11 (landed [#650](https://github.com/mendahu/provenencia/pull/650)), S13-D13 (board)
**Blocks:** S13-15
**Branch:** `s13-36-provenance-sheet-in-the-kit`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go or proto change
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-14 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`../provenance-sheet.md`](../provenance-sheet.md) (the boundary; this plan follows it except where D-8, D-9 and D-10 say); [person-detail.md](../person-detail.md) R-8 – R-15; brief S13-D13 PSK-1 – PSK-9, PSK-W1; working rules "The provenance sheet is a kit recipe", "The provenance sheet replaces the Why on the Person page only", "The person card lives in the design system; display is decoupled from content"; the [S13-34](S13-34-display-pieces-move-down.md) and [S13-35](S13-35-person-card-joins-the-design-system.md) plans (the precedent); [S13-15](S13-15-likeness-viewer.md) D-9 (the next caller)
**Planned:** 2026-11-03 against `main` @ `f2df2b70`
**Review:** [`S13-36-provenance-sheet-in-the-kit.review.md`](S13-36-provenance-sheet-in-the-kit.review.md) — rounds 1 and 2 Fits with changes / Ready with changes; every finding Fixed, Declined with evidence, or answered by the user

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

The provenance sheet looks and behaves exactly as it does today, but it now lives in the design system: one record row, a record table and the sheet, each placeable alone, filled by an assembler that builds them from the wire records every caller already holds. A contributor adding a provenance view makes one assembler call and maps one link callback, and the wording of provenance lives in one place.

**Why now.** S13-15 places one record row under the viewer's image and cannot do that while the row is private to `Features/Conclusions/`. The user asked for the sheet in the kit (2026-11-03), and a family tree in another feature is a known caller.

**What stays.** Every pixel, word, keyboard path and VoiceOver reading of the Vitals sheet and of the Event and Place Why accordion. `ProvenanceSheetItem` and its follow-a-refresh behavior. The Vitals provenance control. The Why's table layout, which stays in `Features/` until those pages are redesigned (working rule "The provenance sheet replaces the Why on the Person page only").

## 2. User stories

**US-1.** As a researcher, I notice nothing: the Vitals sheet is the sheet I had.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Birth date with two owners, five records, one outvoted, Regular pane | Title "Birth date", the value in mono, summary, column headings, two owner lines with "· shown" on the first, 880pt wide, identical to `main` |
| AC-1.2 | A mixed Occupation | Two value headings and "Other records", each record's ink as on `main` |
| AC-1.3 | The same sheets at a Wide pane | 1280pt; Read as 220pt, Outcome 280pt, Citation takes the rest |
| AC-1.4 | A row with 30 records | The records scroll past 560pt; title, header, headings and Done stay |
| AC-1.5 | Each link of a record | The citation line opens the Citation composer with no Observation in focus; Source opens the Source page; Observation opens the composer with the Observation in focus; each closes the sheet first; Back returns to the page |
| AC-1.6 | Keyboard and VoiceOver | Tab order citation line → Copy citation → Source → Observation → next record → Done; each record reads as one line; the title and group lines are headings |
| AC-1.7 | A page refresh while the sheet is open | The sheet follows its row; a row that leaves closes it |

**US-2.** As a researcher on an Event or Place page, the Why accordion is unchanged.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A mixed Event date → Why | Read as · Source · Outcome · reserved column, owner lines as today; a Source opens the Citation with the Observation in focus |

**US-3.** As a contributor, I can place provenance anywhere.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | One `CatalogReconcilerOutcome` and its `CatalogConclusionField` | `PVProvenanceSheetContent.Record(outcome:in:)` gives the record `ReconciliationRecord.init` gave, with `citation: CitationLineDisplay.Parts` |
| AC-3.2 | The recipe folder *(new)* | Names no `Features/` type: `scripts/check-guidance-names.py` passes |
| AC-3.3 | The recipe's previews | One record with and without the Read as column; a lone row above a table built from the same `PVProvenanceColumns`, columns aligned (PSK-4, PSK-6) |
| AC-3.4 | A page load | `ReconciledValueRowModel` builds no records; content is built when a sheet or a Why opens |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ReconciledValueDisplay`, `CitationLineDisplay`, `EventTitleDisplay`, `PlaceTitleDisplay` move to `Platform/` | STEP-1 |
| Content, its groups and records, and the assembler in the recipe (§ 6 **Assembler**) | STEP-2, STEP-3, STEP-4, AC-3.1, AC-3.4 |
| `PVProvenanceColumns`, `PVProvenanceRecordRow`, `PVProvenanceRecordTable`, `PVProvenanceSheet` and their parts (PSK-1 – PSK-9, PSK-W1) | STEP-5, STEP-6, AC-1.1 – AC-1.6 |
| Vitals opens the recipe; links map to places in the caller | STEP-6, AC-1.5, AC-1.7 |
| The Why reads the recipe's content and parts | STEP-4, STEP-6, AC-2.1 |
| `L10n.ProvenanceSheet` | STEP-7 |
| Kit README row, docs | STEP-8 |

| Out | Home |
| --- | --- |
| The one-record variant's product caller; the lone row's caller | S13-15 |
| `PVProvenanceSheetContent.narrowed(to:title:)` | not planned: it has no caller since tiles lost their sheet (user, 2026-11-03); S13-15 uses `Record(outcome:in:)` |
| A `Record` that holds no `CatalogReconcilerOutcome` | the family tree's relationship provenance, when it is designed |
| The Why's table moving onto the kit table | the Event and Place redesign (R-1) |
| Changing the kit specimens to match the app (marks, Record width, Regular columns) | the orchestrator, in Claude Design, after this PR; no repo change |
| Refreshing the S13-15 plan | S13-15's own refresh after this lands |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Sheet | `macos/App/Features/Conclusions/ProvenanceSheet.swift` `ProvenanceSheet`, `ProvenanceSheetItem` | `PVPanel` with header, zero-height reserved block, `SheetHeadings`, `PVDivider`, a `ScrollView` capped at `recordsMaxHeight` 560; width 880 or 1280; `open` closes then navigates; ids `conclusion.detail.provenance.…` |
| Table | `Features/Conclusions/ReconciliationRecordRows.swift` `ReconciliationRecordRows`, `SheetHeadings`, private `CitationLineButton`, `CitationLineButtonStyle`, `citationFirstLineCenter` | One SwiftUI `Grid` with `.why` and `.sheet` modes; fixed widths Read as 140/220, Outcome 190/280, Record 112, reserved 88, spacing 14, copy track 26 + 6 |
| Record | `Features/Conclusions/ReconciliationRecord.swift` `ReconciliationRecord`, `ReconciliationRecordGroup` | Holds three `WorkspaceLocation`s, so it cannot enter the kit; `sourceTitle` and `citationLine` hold the citation twice; `composerLocation(for:focusingObservation:)` is also called by `ConclusionDetailPage.swift:195` |
| Row model | `Features/Conclusions/ReconciledValueRowModel.swift` `ReconciledValueRowModel` | Builds every record in `init`, `spanning`, `derived`; `valueGroups`, `pooledValueGroups`, `ownerLine`, `joinSummary`, `listedValues`, `otherValue` are the provenance rules; `ValueStyle`, `OtherValue` are its types; `whyTitle` is set by `PlaceDetailContent.swift:70` |
| Row view | `Features/Conclusions/ReconciledValueRow.swift` | `valueContent`, `valueLine`, `font(for:)`, `badge(_:)` are static and read by the sheet |
| Outcome | `Features/Conclusions/ReconciliationOutcome.swift` `ReconciliationOutcome` | Mark and phrase, 14pt mark |
| Why | `Features/Conclusions/ReconciliationReasoningView.swift` | Sunken card around `ReconciliationRecordRows` in `.why` mode |
| Presenter | `Features/Conclusions/ConclusionDetailPage.swift:410-430` `ConclusionDetailRowsSection` | `pvSheet(item:)` with `ProvenanceSheetItem`; `onChange(of: rows)` closes a gone row |
| Wording | `Features/Conclusions/ReconciledValueDisplay.swift`, `CitationLineDisplay.swift`, `EventTitleDisplay.swift`, `PlaceTitleDisplay.swift` | Pure; read `Catalog*`, `L10n`, `Platform/` formatters; the only Features name among them is `ReconciliationRecord.sourceTitle` at `CitationLineDisplay.swift:44` |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Conclusions.provenance*`, `whyReadAs`, `whyOutcome`, `badgeMerged`, `badgeMixed`, `badgeConcluded` | Eighteen `provenance*` keys under `conclusions.provenance.*` |
| Kit | `macos/App/DesignSystem/Components/ClipboardButton/PVClipboardButton.swift`, `Panel/PVPanel`, `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` | The clipboard button ships; `PVPersonCardContent` is the content-with-init precedent |
| Sheet closing | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` `PVSheetClose`, `\.pvSheetClose` | `callAsFunction(then:)` closes, then runs the next step; `add-ui-component` makes it the rule, and `ProvenanceSheet.close` is the exception today |
| Tests | `macos/ProvenenciaTests/ProvenanceSheetTests.swift` (23 lines touch these types), `PersonDetailTests.swift` (35), `EventDetailTests.swift` (7), `PlaceDetailTests.swift` (2), `ReconciledValueDisplayTests.swift` (1) | Pin grouping, summaries, ink, wording |

**Precedent.** [#690](https://github.com/mendahu/provenencia/pull/690) (S13-35) moved the person card: content in the recipe with an `init` from a `Catalog*` type, its copy to `L10n.PersonCard`. Copy that. S13-34 moved `PersonTitleDisplay` and friends to `Platform/` by `git mv` with no code change: copy that for STEP-1.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One PR, not two | Six code steps on one view, under `plan-pr`'s size rule; STEP-1 is a `git mv` a reviewer reads commit by commit | A separate Platform PR — a full ship loop for four file moves with no behavior |
| D-2 | STEP-1 moves the four files whole, by `git mv`, in a rename-only commit. A second commit gives `CitationLineDisplay` `sourceTitle(for:)` (moved from `ReconciliationRecord`) and points `ReconciledValueDisplay.outcomePhrase`'s inlined copy of that rule at it | The rest of each file has no reason to stay behind; `provenance-sheet.md` § 5 named the record and state functions, and the file is one type. Two commits keep `git log --follow` clean | Splitting `ReconciledValueDisplay` in two — a seam nobody needs |
| D-3 | Types nest under `PVProvenanceSheetContent` *(new)* in `Recipes/ProvenanceSheet/PVProvenanceSheetContent.swift` *(new)*: `Shown`, `Value`, `Record`, `Group`, `Group.OwnerLine`, `Group.ValueGroup`, `ValueGroup.Heading`, and `PVProvenanceFace`, `PVProvenanceLink` beside them. No view code in an extension of the content type: its nested `Group` shadows SwiftUI's | Matches `provenance-sheet.md` § 5 | Top-level `PVProvenanceRecord` as the web kit names it — the Swift doc already names `PVProvenanceSheetContent.Record` and S13-15 D-9 cites it |
| D-4 | `Record` holds `outcome: CatalogReconcilerOutcome`, `citation: CitationLineDisplay.Parts`, `id`, `readAs`, `readAsIsPhrase`, `counted`, `phrase`, `mark: PVSymbol`, `observationRef`, `subjectRef`, and the computed `spoken`. `sourceTitle`, `citationLine` and `groupRank` go: read `citation.title`, `citation.line`, `outcome.groupRank`. It holds no `WorkspaceLocation` | The map's scope entry; `CitationLineDisplay.line(_:)` then has no app caller and is deleted | A `citation: String` as the web kit has — it would hold the citation a second time |
| D-5 | `ReconciledValueRowModel` holds what to build from, not what was built: `shown: PVProvenanceSheetContent.Shown` (D-16), a private `Equatable` enum `RecordsSource` *(new)* with `.field(CatalogConclusionField)`, `.spanning([CatalogConclusionField])`, `.derived(owners:shownOwnerID:kind:)`, plus `locale` and `orders`. `recordCount` is computed from the source. `func provenanceContent() -> PVProvenanceSheetContent` *(new)* builds on call, passing the row's `label` and `shown`, so a relabelled row and its sheet cannot disagree. `.empty` holds `.spanning([])`. `records`, `recordGroups`, `provenanceSummary`, `style`, `lead`, `emptyText`, `badge`, `listedValues` leave the model; callers and tests read `row.shown.…`, and `PlaceDetailContent.namesRow` sets `row.shown.emptyText`. `otherValues` stays, as `[PVProvenanceSheetContent.Value]` | `provenance-sheet.md` § 5 "Build on open": the sheet follows a refresh for free and nothing carries a precomputed sheet | Keeping eager content on the row — it costs every row on every load for a sheet most rows never open |
| D-6 | `PVProvenanceFace` *(new)*, `.text`, `.date`, `.place`, replaces `ReconciledValueRowModel.ValueStyle` everywhere, and every `style` property or parameter of that type becomes `face`; `PVProvenanceSheetContent.Value` replaces `OtherValue` with the same fields. `PVProvenanceFace.valueFont` and `readAsFont` *(new)* replace `ReconciledValueRow.font(for:)` and the private `readAsFont`, with today's sizes | One type for one idea, as the brief's Value faces row moves it | A mapping between two enums |
| D-7 | `PVProvenanceRecordRow` is an `HStack(alignment: .firstTextBaseline, spacing: columns.spacing)` of fixed-width cells with Citation at `maxWidth: .infinity`; `PVProvenanceRecordTable` is a `VStack(spacing: 0)` of rows, rules and group lines; no `Grid` | A row must stand alone, and `GridRow` needs a `Grid`; `SheetHeadings` already lines up with the table as an `HStack`, so every column but Citation is fixed and nothing moves | A `Grid` inside each row — rows would not line up with one another |
| D-8 | The link callback is `open: (PVProvenanceSheetContent.Record, PVProvenanceLink) -> Void`; `PVProvenanceLink` *(new)* is `.source`, `.citation`, `.observation`. `WorkspaceLocation.provenance(_ record:, _ link:)` *(new)* and `WorkspaceLocation.composer(for:focusingObservation:)` *(moved from `ReconciliationRecord.composerLocation`)* go in `Features/Workspace/WorkspaceLocation.swift`. `Features/Conclusions/ProvenanceSheet.swift` is renamed `ProvenanceSheetItem.swift`. The table takes this callback from STEP-3 on; the Why's Source cell reports `.observation`, today's destination for it | The kit names no navigation type; one mapping for every caller, including a tree in another feature. Handing back the record lets a later non-outcome `Record` change one type and one mapping, not three kit levels | An outcome-typed callback, as `provenance-sheet.md` § 5.1 drew it — every kit signature changes for the named relationship caller; a fourth link case for the Why — it is the Observation destination |
| D-9 | The Why keeps `ReconciliationRecordRows` as its own `Grid` with `.why` the only mode (the `provenance` parameter goes), reads `PVProvenanceColumns.spacing` and `.actionWidth`, and draws the recipe's parts `PVProvenanceOwnerLine`, `PVProvenanceOutcome` and `PVProvenanceReadAs`. `ReconciliationOutcome.swift` is deleted | Only the table layout is the Why's. The brief left it the Features outcome view; sharing the part avoids a second copy of one 10-line view with no visible change | Copying the owner line and outcome into Features — copies a later review removes |
| D-10 | No `.narrowed`; no `tableMaxHeight` parameter: `PVProvenanceSheet.recordsMaxHeight` stays 560; no `showsHeadings`: a caller places `PVProvenanceColumnHeadings` above a table. § 6 **Assembler** is the whole assembler | Nothing calls the others; the optional heading row of PSK-6 stays a composition | Building for callers that do not exist |
| D-11 | Where the board and the app differ, the app wins and nothing changes: outcome marks (`ReconciledValueDisplay.outcomeMark`), Record column 112pt, Regular columns 140 and 190pt, the Read as cell carrying the record's one spoken line | The map's Check: "looks and behaves as before"; S13-36 is preparatory | Adopting the board's marks — a visible change this PR has no story for |
| D-12 | `L10n.ProvenanceSheet` *(new)* takes the sheet's words: `otherRecords`, `ownerShown`, `ownerShownMark`, `columnReadAs`, `columnCitation`, `columnOutcome`, `columnRecord`, `sourceLink`, `sourceLinkSpoken`, `observationLink`, `copyCitation`, `citationCopied`, `done`, `recordCount`, `recordsFromOwners`, `stateSingle`, `summaryJoin`, `badgeMerged`, `badgeMixed`, `badgeConcluded`. Each catalog key is today's with `conclusions.provenance.` replaced by `provenanceSheet.`, so `provenanceSheet.column.citation`, and the three `recordsFrom…` variants keep their suffixes; `conclusions.why.readAs` → `provenanceSheet.column.readAs`, `conclusions.why.outcome` → `provenanceSheet.column.outcome`, `conclusions.badge.merged` → `provenanceSheet.badge.merged`, and likewise `.mixed`, `.concluded`. Values and comments carry over. `whyReadAs`, `whyOutcome` are deleted: the Why reads the new columns. Stay in `L10n.Conclusions`: `provenanceOpen` (the caller's control), `whySource`, `sourceCount`, `citation*`, `a11y*`, `against*`, `outcome*`, read through `Platform/` | S13-35's `L10n.PersonCard`; one key per word | Leaving recipe copy in `L10n.Conclusions` |
| D-13 | Row, table and sheet take `idPrefix: String`; `ConclusionDetailRowsSection` passes `"conclusion.detail.provenance"`, so every id is as today: `\(idPrefix).\(record.id).citation`, `.copy`, `.source`, `.observation`, and `\(idPrefix).done`. The recipe applies `.accessibilityElement(children: .contain)`; the caller applies the sheet's own id `conclusion.detail.provenance.\(row.id)` outside it | The kit names no page | A hard-coded prefix |
| D-14 | `PVProvenanceStateBadge(_ state: ReconciledValueDisplay.StateBadge)` *(new)* replaces `ReconciledValueRow.badge(_:)`, and the Vitals row draws it | The board's part; recipes may read `Platform/` | A second badge mapping |
| D-15 | `PVProvenanceColumns` *(new)* is `Equatable` with `showsReadAs: Bool = true`, `widthClass: PVWidthClass = .regular`, computed `readAsWidth`, `outcomeWidth`, `panelWidth` (880, 1280), and statics `spacing` 14, `recordWidth` 112, `actionWidth` 88, `copyTrackWidth` 26, `copyTrackSpacing` 6. With `showsReadAs` false the row drops that cell and the record's `spoken` line moves to its Outcome cell, so the record still reads as one element. The reserved column always shows | PSK-4, PSK-6, D-11; S13-15's Panel reads `panelWidth` | Passing widths loose |
| D-16 | `PVProvenanceSheetContent.Shown` *(new)* holds `face`, `lead: String?`, `listedValues: [Value]`, `emptyText`, `badge: ReconciledValueDisplay.StateBadge?`. Its `init(field:locale:orders:)` is the one derivation of what a value line shows; a derived or spanning row builds one from its given lead (`mixed` sets `badge` as today). `PVProvenanceValueList(shown)` *(new)* draws it for `ReconciledValueRow` and the sheet header, and `PVProvenanceValueLine(_ value:, face:)` *(new)* draws one `Value` with its Source count for that list and for a value heading. `PVProvenanceSheetContent` holds `title`, `shown`, `summary`, `groups` | Today's `valueContent` and `valueLine`, moved, with one derivation for row and sheet | A `values` array as the web kit has — it cannot say lead versus listed; four loose fields — row and sheet could disagree after a relabel |

**Rules this plan bends.** D-8 bends the working rule "The provenance sheet is a kit recipe" and `provenance-sheet.md` § 5.1, both dated user 2026-11-03, which give the callback the outcome: it hands back the `Record`, which carries the outcome. The user's answer: the `Record` (2026-11-03). STEP-8 updates both. D-9 narrows the brief's inventory row "Outcome mark and phrase", and D-10 drops `.narrowed` from § 5; STEP-8 updates those documents, the brief's row before it is archived. The Why keeps its look and its table, as the working rule says.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `Recipes/ProvenanceSheet/PVProvenanceSheetContent.swift` | type, assembler | `ReconciliationRecord`, `ReconciledValueRowModel`'s statics | Moved from them; those names hold `WorkspaceLocation` and Vitals fields |
| `PVProvenanceColumns` | type | `ReconciliationRecordRows` statics | Moved |
| `PVProvenanceRecordRow.swift` | view | `sheetRow`, `CitationLineButton` | Moved |
| `PVProvenanceRecordTable.swift` (with `PVProvenanceOwnerLine`, the value heading, `PVProvenanceColumnHeadings`) | view | `ReconciliationRecordRows` sheet mode, `SheetHeadings` | Moved |
| `PVProvenanceSheet.swift` | view | `ProvenanceSheet` | Moved |
| `PVProvenanceParts.swift` (`PVProvenanceOutcome`, `PVProvenanceReadAs`, `PVProvenanceStateBadge`, `PVProvenanceValueList`, `PVProvenanceValueLine`) | views, type | `ReconciliationOutcome`, `ReconciledValueRow` statics | Moved |
| `RecordsSource`, `provenanceContent()` | type, func | `CatalogQueryKey` holds no content either | Private to the row model |
| `Shown` | type | The row's loose `lead`, `listedValues`, `emptyText`, `badge` | Replaces them (D-16) |
| `runs(of:)` | private func | The two run-splitting loops in `valueGroups`, `pooledValueGroups` | One loop for both |
| `WorkspaceLocation.provenance(_:_:)` | func | `ReconciliationRecord.composerLocation` | Same code, in the file that owns the type, which the kit does not see |
| `L10n.ProvenanceSheet` | namespace | `L10n.PersonCard` | Same shape |

**Builds on.** `add-ui-component` (recipe folder, README row), `\.pvSheetClose`, `PVPanel`, `PVClipboardButton`, `PVButton`, `PVBadge`, `pvSheet(item:)`, S13-35's content shape.
**Deletes.** `ReconciliationRecord.swift` (both types), `ReconciliationOutcome.swift`, `ProvenanceSheet` the struct, `ReconciliationRecordRows.SheetHeadings`, its `.sheet` mode, `sheetRow`, `valueHeading`, `segments(of:)`, `ownerLine`, `CitationLineButton`, `CitationLineButtonStyle`, `citationFirstLineCenter`, `actionColumnWidth`, `columnSpacing`, `copyTrackWidth`, `copyTrackSpacing`, `readAsWidth`, `outcomeWidth`, `recordWidth`, `widthClass`, `isSheet`, `columns`; `CitationLineDisplay.line(_:)` (its tests move to `.parts(…).line`); `ProvenanceSheet.width(_:)`; `ReconciledValueRowModel.ValueStyle`, `OtherValue`, the fields D-5 removes, and every private static it moves; `ReconciledValueRow.valueContent`, `valueLine`, `font(for:)`, `badge(_:)`; the `L10n.Conclusions` keys of D-12.
**Next adds.** S13-15: one `PVProvenanceRecordRow` from `Record(outcome:in:)`. Event and Place redesign: `PVProvenanceSheet` with `provenanceContent()`, then `ReconciliationRecordRows` and `ReconciliationReasoningView` go. A relationship's provenance changes `Record.outcome` and the one mapping, `WorkspaceLocation.provenance`.

### Reach

None — no new path into existing code: the sheet opens from the same control, and each link reaches the same place.

### Scale

At the map's Scale (a person with 41-Subject-sized Sources, a row with 30+ records): page load builds fewer records than today (D-5). A sheet builds its content on each body evaluation while open, in work linear in the row's outcomes, plus one `outcomePhrase` lookup per record over `field.outcomes`, as `init` does today. The table is an eager `VStack`, as the `Grid` was. A row's `==` compares its held fields, about the size of the records it compared before; `onChange(of: rows)` keeps that cost. No write or cache changes.

### Edges

None — no rule changes; every rule moves with its tests. **Assumptions:** `groupRank` runs and the shown-owner-first order move unchanged; `ProvenanceSheetTests` holds them. **Untrusted size:** the record count comes from the wire, unbounded as today.

### Schema

None — no catalog change.

### Go core

None — no Go change.

### FFI

None — no RPC change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Sheet | `PVProvenanceSheet(content:columns:idPrefix:open:)` | `Recipes/ProvenanceSheet/` | Body of today's `ProvenanceSheet`; title `.accessibilityHeading(.h1)`; `columns.panelWidth`; reserved block; `PVProvenanceColumnHeadings`, `PVDivider`, `ScrollView` capped 560. Reads `@Environment(\.pvSheetClose)`: Done calls `sheetClose()`, a link runs `sheetClose { open(record, link) }`. Done `.keyboardShortcut(.defaultAction)` |
| Table | `PVProvenanceRecordTable(groups:face:columns:idPrefix:open:)` | same | Owner rule `borderStrong`, others `borderSubtle`; first record under a line draws no rule. `PVProvenanceOwnerLine` and the value heading keep today's padding; `gridCellColumns` stays with the Why |
| Row | `PVProvenanceRecordRow(record:face:columns:idPrefix:open:)` | same | Today's `sheetRow`; vertical padding `space3`; ids per D-13 |
| Citation link | private in the row file | same | `CitationLineButton`, its style, the alignment guide, moved unchanged |
| Caller | `ConclusionDetailRowsSection` | `ConclusionDetailPage.swift` | Builds `row.provenanceContent()` inside the `pvSheet` closure; `open: { navigation.go(to: .provenance($0, $1)) }`; `.environment(navigation)` goes |

**Assembler.** The whole API, in `PVProvenanceSheetContent.swift`; each defaults `locale: .autoupdatingCurrent`, `orders: NameFormatOrders()` as the row model does:

```swift
extension PVProvenanceSheetContent.Record { init(outcome:, in field:, locale:, orders:) }
extension PVProvenanceSheetContent.Shown { init(field:, locale:, orders:) }
extension PVProvenanceSheetContent {
    init(field:, title: String, shown: Shown? = nil, locale:, orders:)              // nil: Shown(field:…)
    static func spanning(fields:, title: String, shown: Shown, locale:, orders:) -> Self
    static func derived(owners:, shownOwnerID: String?, kind: CatalogDerivedRowKind?,
                        title: String, shown: Shown, locale:, orders:) -> Self
    static func value(_: CatalogReconciledValue, propertyKey:, locale:, orders:) -> Value
}
```

Summaries as today: `init` the state line, against line and record count; `.spanning` the record count, nil with no records; `.derived` "records from owners" with two or more groups and a `kind`, else the count. `value(…)` is the one builder of a `Value`: `otherValues`, `Shown` and the value headings all call it.

States: as today. One record: no group line, summary "1 record". Accessibility: as AC-1.6.

### Strings

L10n keys *(new)*: the `L10n.ProvenanceSheet` list of D-12, keys by its rule, values unchanged. Removed: the matching `conclusions.*` entries. No visible word changes.

## 7. Steps

**STEP-1 — The wording and titles move to Platform**
Files: `ReconciledValueDisplay.swift`, `CitationLineDisplay.swift`, `EventTitleDisplay.swift`, `PlaceTitleDisplay.swift` → `macos/App/Platform/`; `ReconciliationRecord.swift`.
Do: D-2, in its two commits. The Xcode project uses synchronized folders: no `project.pbxproj` edit.
Check: `scripts/check-pr.sh` PASS; `git log --follow` traces each file.

**STEP-2 — The face and the value types**
Files: `macos/App/DesignSystem/Recipes/ProvenanceSheet/PVProvenanceSheetContent.swift` *(new)*, these types only, `ReconciledValueRowModel.swift`, `ReconciledValueRow.swift`, `ReconciliationRecord.swift`, `ReconciliationRecordRows.swift`, `ReconciliationReasoningView.swift`, `PersonDetailContent.swift`, tests.
Do: D-6, and `value(…)`. Rename only.
Check: `xcodebuild test`, as `scripts/check-pr.sh` runs it, passes; `rg '\b(ValueStyle|OtherValue)\b' macos` prints nothing.

**STEP-3 — Records and groups are the recipe's, built by the assembler**
Files: `PVProvenanceSheetContent.swift`, `ReconciledValueRowModel.swift`, `ReconciliationRecordRows.swift`, `ProvenanceSheet.swift`, `WorkspaceLocation.swift`, `ConclusionDetailPage.swift`, tests; delete `ReconciliationRecord.swift`. Skill: `add-swift-test`.
Do: D-3, D-4, D-8, and the **Assembler** block without `Shown`: at the end of this step the content holds `title`, `summary`, `groups`, and no assembler entry takes `shown:`. The sheet's header still reads the row's fields. Move `valueGroups`, `pooledValueGroups`, `ownerLine`, `joinSummary` with one `runs(of:)`; `derived` keeps forcing `counted = false` on other owners. `ReconciliationRecordRows` takes `open: (Record, PVProvenanceLink) -> Void`, for both modes, and its callers map it through `WorkspaceLocation.provenance`. The row still builds eagerly here. Add T-1, T-3.
Check: `xcodebuild test` passes.

**STEP-4 — A row builds its content on call**
Files: `ReconciledValueRowModel.swift`, `ReconciledValueRow.swift`, `ProvenanceSheet.swift`, `ConclusionDetailPage.swift`, `PersonDetailContent.swift`, `EventDetailContent.swift`, `PlaceDetailContent.swift`, tests.
Do: D-5, `Shown` and D-16's data; the sheet reads `provenanceContent()`, and the Why reads its `groups`. `ReconciledValueRow.valueContent` and `valueLine` stay until STEP-6, reading `row.shown`. Add T-2; update T-5, T-6.
Check: `xcodebuild test` passes; `rg 'recordGroups|provenanceSummary' macos/App` prints nothing.

**STEP-5 — Columns, row, table and parts**
Files: `macos/App/DesignSystem/Recipes/ProvenanceSheet/PVProvenanceColumns.swift` *(new)*, `PVProvenanceRecordRow.swift` *(new)*, `PVProvenanceRecordTable.swift` *(new)*, `PVProvenanceParts.swift` *(new)*. Skill: `add-ui-component`.
Do: D-7, D-14, D-15, `PVProvenanceValueList`, § 6 macOS. Previews: PSK-2, PSK-3, PSK-4, PSK-6 (a lone row over a table, a row with no Read as), PSK-7, PSK-9. Add T-4.
Check: the app builds; the previews render; T-4 passes.

**STEP-6 — Vitals and the Why move onto the recipe**
Files: `PVProvenanceSheet.swift` *(new)*, `ProvenanceSheet.swift` → `ProvenanceSheetItem.swift`, `ConclusionDetailPage.swift`, `ReconciledValueRow.swift`, `ReconciliationRecordRows.swift`, `ReconciliationReasoningView.swift`; delete `ReconciliationOutcome.swift`.
Do: D-9, D-13, § 6 macOS, the **Deletes** list. `ReconciledValueRow` draws `PVProvenanceValueList`. `ReconciledValueRow.Provenance` stays with `.why` and `.sheet`: it is how a page chooses the Why or the provenance control. `ReconciliationReasoningView` keeps its `WorkspaceNavigation` and passes `open: { navigation.go(to: .provenance($0, $1)) }`; the Why reports only `.observation`, having no citation link. Recipe comments do not name `Features/` types.
Check: `scripts/check-pr.sh` PASS, including `guidance-names`; `rg 'WorkspaceLocation|WorkspaceNavigation|ReconciledValueRow' macos/App/DesignSystem/Recipes/ProvenanceSheet` prints nothing.

**STEP-7 — The sheet's copy gets its namespace**
Files: `L10n.swift`, `Localizable.xcstrings`, the recipe, `ReconciliationRecordRows.swift`, tests. Skill: `add-localized-string`.
Do: D-12.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `rg 'Conclusions\.(provenance|whyReadAs|whyOutcome|badge(Merged|Mixed|Concluded))' macos` prints only `provenanceOpen`.

**STEP-8 — Docs**
Files: `macos/App/DesignSystem/README.md` (a ProvenanceSheet row), `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row, S13-15's row, the working rule "The provenance sheet is a kit recipe"), `provenance-sheet.md` (§§ 3, 5, 5.1: no `.narrowed`, the callback hands back a `Record`, `PVProvenanceColumns` has no action-column toggle, `composer(for:)` moved, the Why shares parts), this plan (status, Landed note), `design/S13-D13-provenance-sheet-in-the-kit.md` → `design/archive/` with its Board line filled in and its Outcome row saying the Why shares the part, `design/README.md`, `docs/macos-client-patterns.md:106`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/ProvenanceSheetTests.swift` `aRecordComesFromOneOutcome` *(new)* | An outvoted outcome in a merged field → `readAs`, `counted` false, `phrase`, `mark`, `citation == CitationLineDisplay.parts(outcome)`, `outcome` kept | AC-3.1, D-4 |
| T-2 | Swift | same, `aRowBuildsItsContentOnCall` *(new)* | `ReconciledValueRowModel(field:)` with 3 outcomes → `recordCount == 3`; `provenanceContent()` gives the 3 records, title, summary; `.empty` → 0 and no records | AC-3.4, D-5 |
| T-3 | Swift | same, `linksReachTheirPlaces` *(new)* | A record through `.source` → `.sourcePage(sourceId: outcome.sourceID, ref: nil, title: outcome.sourceTitle)`; `.citation` → `.citationComposer` with `observationId` nil; `.observation` → with the id | AC-1.5, D-8 |
| T-4 | Swift | `macos/ProvenenciaTests/PVProvenanceColumnsTests.swift` *(new)* `columnsFollowTheWidthClass` | `.regular` → 140, 190; `.wide` → 220, 280; `recordWidth` 112, `actionWidth` 88 | AC-1.3, D-11, D-15 |
| T-5 | Swift | `ProvenanceSheetTests` every existing test, updated to `provenanceContent()` | Same inputs, same expected values | AC-1.1, AC-1.2, AC-1.7 |
| T-6 | Swift | `EventDetailTests`, `PlaceDetailTests`, `PersonDetailTests`, updated the same way | Same expected values | AC-2.1 |

Must stay green: everything `scripts/check-pr.sh` runs; `ProvenanceSheetTests` is the suite most likely to break.

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
- [ ] Version bump as the header says: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc updated: provenance-sheet.md § 5, macos-client-patterns.md line 106
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Compare against a build of `main` @ `f2df2b70` where a step says so.

1. Open a Person with a merged name and a birth date on two events. Open each row's sheet at Regular: title, value, Badge, summary, owner lines, "· shown", ink and marks match `main` (AC-1.1).
2. Open a mixed value: the value headings and "Other records" match (AC-1.2).
3. Stretch the pane past 1600 with a sheet open: 1280pt, wider Citation, scroll and focus kept (AC-1.3).
4. Open a row with many records: Done and the headings stay while the records scroll (AC-1.4).
5. Follow a citation line, a Source link and an Observation link: each closes the sheet and lands as on `main`; Back returns (AC-1.5).
6. Tab through a sheet, then with VoiceOver: order and readings as AC-1.6; copy a citation and hear "Citation copied".
7. Open an Event and a Place, expand a Why: unchanged, and a Source opens the Citation (AC-2.1).
8. Dark mode: steps 1 and 7 (PSK-9).
9. In Xcode, open the previews in `macos/App/DesignSystem/Recipes/ProvenanceSheet/` *(new)*: one record with and without Read as, and a lone row above a table with its columns aligned (AC-3.3). Run `python3 scripts/check-guidance-names.py`: it passes (AC-3.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `HStack` rows shift a baseline the `Grid` gave | D-7 keeps the `firstTextBaseline` alignment and fixed widths; manual check 1 compares with `main` |
| Tests assert on removed fields | T-5 keeps inputs and expected values; only the access path changes |
| A sheet's content rebuilt on redraw changes `ForEach` ids | Ids are observation ids, as today |

**Stop and report if:**

- A row cannot match `main` with `HStack` columns (D-7): give the screenshot pair.
- `check-guidance-names.py` fails on a name the recipe needs from `Features/`.
- An owner line, value heading or outcome differs between sheet and Why in a way one part cannot express (D-9).
- `ReconciledValueDisplay`, `EventTitleDisplay` or `PlaceTitleDisplay` turns out to need a `Features/` type in `Platform/`.

## 12. Open questions

None.

## 13. PR

**Title:** S13-36 — The provenance sheet joins the design system
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-36-provenance-sheet-in-the-kit.md`.
