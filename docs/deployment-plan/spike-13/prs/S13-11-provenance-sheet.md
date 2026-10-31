# S13-11 — Provenance sheet

**Status:** Blocked — on board S13-D2
**Spike / slice:** Spike 13 · Slice 2 — Provenance sheet
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D2](../design/S13-D2-provenance-sheet.md), board agreed 2026-10-31
**Board:** agreed 2026-10-31 — [Claude Design, *Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail.dc.html*, turn 6 "S13-D2 · Provenance sheet", frames 6a – 6n. §6 **macOS** describes each frame in words; the brief's §8 records where the board and the plan differ
**Depends on:** S13-06 (landed [#640](https://github.com/mendahu/provenencia/pull/640)), S13-10 (landed [#634](https://github.com/mendahu/provenencia/pull/634)), S13-D2 (board agreed 2026-10-31)
**Blocks:** —
**Branch:** `s13-11-provenance-sheet`
**Version bump:** none — Spike 11 – 14 PRs leave `VERSION` alone until a release is cut, as S13-06 and S13-10 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`person-detail.md`](../person-detail.md) R-8 – R-13b, R-14, R-15; working rules: "The provenance sheet replaces the Why on the Person page only", "A derived row carries its owning entity's records", "Engine PRs run ahead of their slice", "Redesigns adopt Page columns"; [`../deployment-plan.md`](../deployment-plan.md) scope S13-11; brief S13-D12 ([archived](../design/archive/S13-D12-life-rows-why.md)) for the group line per owner; S13-06 review UX-3 (VoiceOver names the shown owner); [`docs/ideas/citation-formats.md`](../../../ideas/citation-formats.md) (the citation line, user 2026-10-31)
**Planned:** 2026-10-30 against `main` @ `c54f34b`
**Review:** — (the plan's review runs once the board is agreed and D-10 is filled; the brief's review, whose fixes this plan carries in D-11 – D-15, is [`../design/S13-D2-provenance-sheet.review.md`](../design/S13-D2-provenance-sheet.review.md))

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

On the Person page, every Vitals value explains itself in one click: a provenance control beside the value opens a sheet listing every record behind it, what the auto-reconciler did with each and why, and links that open the record's Source, its Citation and its Observation (dogfood bar 2). The sheet replaces the Why accordion on the Person page.

**Why now.** S13-06 gave the four life rows their records, and S13-10 reserved a 28pt slot at the end of every Vitals row for this control. The slice's Check is bar 2.

**What stays.** Each row's value, state Badge, Source count, disagree note, other-values disclosure and open arrow. The Event and Place pages keep the Why accordion (working rule "The provenance sheet replaces the Why on the Person page only"). The read: `GetConclusionDetail` stays the one read; its outcomes gain the Artifact label, Citation label and page the citation line needs (D-8).

## 2. User stories

**US-1.** As a researcher on a Person's Vitals, when I click the control beside a value, I see every record behind it and what happened to each, so that I can judge the value without leaving the page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The name row has three records (two Kept into "James Robins", one Folded) | The row shows the control and no Why button; the control opens a sheet titled "Name" listing the three records, Kept ones in primary ink |
| AC-1.2 | A life row with no event (birth date stated empty, no records) | The slot is empty; no control |
| AC-1.3 | A birth date stated empty whose only record is Denied | The control shows; the sheet lists the Denied record |
| AC-1.4 | The Event page's date row | Shows the Why button and the accordion, no control (unchanged) |

**US-2.** As a researcher, from a record in the sheet I open its Source, its Citation or its Observation, so that I can check the evidence itself.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Click the record's Source | The sheet closes; the Source page of `source_id` opens; Back returns to the Person on Vitals |
| AC-2.2 | Click the record's Citation | The sheet closes; the citation composer opens on `citation_id` with no Observation in focus |
| AC-2.4 | A record on Citation "Baptism of Ada", Artifact "Vol. 3", Source "Register of Baptisms", locator page 37 with no printed label | Its citation line reads "Register of Baptisms, Vol. 3, Baptism of Ada, image 37"; with `page_label` "23" it reads "…, p. 23"; a Source with no Artifact label, Citation label or page reads "Register of Baptisms" |
| AC-2.3 | Click the record's Observation | The sheet closes; the composer opens on `citation_id` with `observation_id` in focus (today's Why Source link) |

**US-3.** As a researcher on a person with two birth events, the sheet tells me which records belong to which event and which event the row shows.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Birth date with records from `EVT-4KX2M` (shown) and `EVT-9PQ2R` | Two groups, `EVT-4KX2M`'s first, each under its group line; the second group's records in secondary ink |
| AC-3.2 | VoiceOver on the first group line | Reads "Birth of Ada Lovelace, EVT-4KX2M, shown value" as a heading |
| AC-3.3 | One owner | No group line |

**US-4.** As a researcher, the sheet stays true while the page refreshes.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | The sheet is open on Birth date; the page reloads with a fourth record | The sheet lists four records |
| AC-4.2 | The sheet is open on an Occupation row; the reload drops that row | The sheet closes |

## 3. Scope

| In | Covered by |
| --- | --- |
| The control on every Vitals row with records (PRV-1) | STEP-4, AC-1.1 – AC-1.3 |
| No Why on the Person page's rows (PRV-2) | STEP-4, AC-1.1, AC-1.4 |
| The sheet: header, summary, records, refs (PRV-3 – PRV-5) | STEP-5, STEP-6, STEP-7, AC-1.1 |
| Value groups (PRV-6) | STEP-7, T-5 |
| Owner groups and the shown owner (PRV-7, PRV-8) | STEP-3, STEP-7, AC-3.1 – AC-3.3 |
| The citation line (PRV-15) | STEP-1, STEP-2, AC-2.4 |
| Three links per record (PRV-9) | STEP-2, STEP-7, AC-2.1 – AC-2.3 |
| Room for Reconciliation Claims (PRV-10) | STEP-6, STEP-7 |
| Long and refreshing (PRV-11) | STEP-5, AC-4.1, AC-4.2 |
| Keyboard, VoiceOver, dark mode (PRV-12 – PRV-14) | STEP-4, STEP-7 |
| Wide (PRV-W1, PRV-W2) | STEP-5 |

| Out | Home |
| --- | --- |
| Reconciliation Claims and a "conclude this value" action (R-10) | not planned this spike ([`person-detail.md`](../person-detail.md#not-this-spike)) |
| The control in Family, Timeline, Gallery, Sources (R-13b) | not planned this spike |
| The sheet on the Event and Place pages; deleting the Why accordion | their redesign ([`person-detail.md`](../person-detail.md#not-this-spike)) |
| A formatted standard citation (*Evidence Explained* templates per Source type, short form, Copy citation) and the Source fields it needs | [`docs/ideas/citation-formats.md`](../../../ideas/citation-formats.md) |
| The Source date (descoped 2026-10-31: [`docs/ideas/source-provenance-date.md`](../../../ideas/source-provenance-date.md), Spike 15) |
| The Source type, or any other new wire field beyond the citation line's parts | a follow-up engine PR if the board asks (working rule "Engine PRs run ahead of their slice") |
| New outcome words or causes | S11-28, S11-44, S11-45 (Spike 14) change the reconciler's causes; the sheet words what it is given |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Row | `macos/App/Features/Conclusions/ReconciledValueRow.swift` `ReconciledValueRow` | `Provenance` is `.why` or `.sheet`. In `.sheet` mode a `Color.clear` 28 × 28 slot sits at the row's end, hidden from VoiceOver, and the Why disclosure and its `ReconciliationReasoningView` still show. `disclosures` shows the Why button when `model.records` is not empty |
| Rows section | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailRowsSection` | Draws a heading (when `title` is given) and one `ReconciledValueRow` per row, passing `provenance` |
| Person Vitals | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` | Vitals is `ConclusionDetailRowsSection(rows:refreshing:title: nil, provenance: .sheet)` |
| Why table | `macos/App/Features/Conclusions/ReconciliationReasoningView.swift` `ReconciliationReasoningView` | Sunken Card, title, Grid Read as · Source · Outcome · 88pt reserved column. Takes `groups: [ReconciliationRecordGroup]`; draws `ownerLine` before a group whose `owner` is set (S13-06). The Source link calls `navigation.go(to: record.location)` |
| Record | `macos/App/Features/Conclusions/ReconciliationRecord.swift` `ReconciliationRecord` | `id`, `readAs`, `readAsIsPhrase`, `counted`, `sourceTitle`, `phrase`, `mark`, `location` (`composerLocation(for:)`: the composer on the Citation with the Observation in focus). No Source page or Citation-only location, no refs |
| Groups | same file `ReconciliationRecordGroup`, `ReconciliationRecordGroup.OwnerLine` | `owner` nil for one owner; `OwnerLine` has `id`, `title`, `ref`, `spoken` ("title, ref") |
| Row model | `macos/App/Features/Conclusions/ReconciledValueRowModel.swift` `ReconciledValueRowModel` | `recordGroups`, `records` (flat), `whyTitle`, `badge`, `count`, `against`, `listedValues`. `derived(…)` builds owner groups, the shown owner first, `ownerLine(_:style:locale:orders:)` titles them |
| Only caller with owner groups | `macos/App/Features/Conclusions/PersonDetailContent.swift` `PersonDetailContent.rows` | Passes `owners:` to `.derived`. `EventDetailContent.placeRow` calls `.derived` with no owners, so after this PR no `.why` row has an owner group |
| Outcome text | `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `outcomePhrase`, `outcomeMark`, `weakCauses`, `spokenState`, `countLine`, `againstLine` | The reconciler's words, reused as is |
| Outcome cell | `macos/App/Features/Conclusions/ReconciliationOutcome.swift` `ReconciliationOutcome` | Mark + phrase |
| Outcome read | `core/database/conclusiondetails/conclusiondetails.go` `sqlOutcomes`, `loadOutcomes`, `Outcome`, `finish` | Joins `citations c`, `artifacts a` and `sources src` already, and selects `c.id`, `a.id`, `src.id`, `src.title`; not `a.label`, `c.label` or `c.locator_json`. `finish` orders a field's outcomes by `ValueRank` (0 last), then Observation id; an outvoted outcome carries its losing value's rank (`conclusiondetails_test.go` "outvoted", `ValueRank` 2) |
| Outcome wire | `api/proto/engine.proto` `ReconcilerOutcomeDetail` (fields 1 – 20; 14, 16 reserved); `api/ffi/handlers/conclusion_detail.go` `reconcilerOutcomeProto` | No Artifact label, Citation label or page |
| Locator | `core/locator/validate.go` `decode`, `validateDoc`, `document`, `selectorRaw` (`ArtifactPage *int`, `PageLabel *string`), `TypePage`, `MaxDocumentBytes`; `core/locator/depiction.go` `ValidateDepictionRegion` | `ValidateDepictionRegion` already reads a Citation locator's pages (decode, validate, collect page selectors, at most one); nothing exported returns the page |
| Swift wire | `macos/App/Platform/GenealogyStore.swift` `CatalogReconcilerOutcome`; `macos/App/Platform/GoStore.swift` (the `CatalogReconcilerOutcome(` mapping) | Carries `observationRef`, `subjectID`, `subjectRef`, `graphID`, `citationID`, `artifactID`, `sourceID`, `sourceTitle`. No Artifact label, Citation label or page |
| Go tests | `core/database/conclusiondetails/conclusiondetails_test.go` (`outcomeOf`, the `fixture` with its `source` helper filing Artifact "Scan" and the package `const locator` page 1), `api/ffi/handlers/conclusion_detail_test.go` `TestConclusionDetailProtoCarriesVoteArtifactAndMembers` (tests `reconcilerOutcomeProto` directly), `core/locator/depiction_test.go`, `core/locator/validate_test.go` | Cover outcomes' Source title and vote, the proto mapping, and the locator's validation; other tests share the fixture's "Scan" label and page 1 |
| Locations | `macos/App/Features/Workspace/WorkspaceLocation.swift` `sourcePage(sourceId:ref:title:)`, `citationComposer(…)` | Both exist. A `citationComposer` with a `citationId` and no `observationId` opens the Citation for edit (`CitationComposerEntry.init?(location:)`, `.edit`) |
| Sheet kit | `macos/App/DesignSystem/Components/Panel/PVSheet.swift` `.pvSheet(item:…)`; `PVPanel.swift` `PVPanel` | One presenter, Escape and outside click; `PVPanel` title, subtitle, body, optional footer |
| Page columns | `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` `PVPageColumns` | Has a secondary slot; the Person page passes none |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` `aSourceOpensTheRecordsCitationInTheComposer`, `lifeRowsCarryTheirOwnersRecords`, `anOwnerThatIsNotTheShownOneIsNamed`, `competingOwnersAreGroupedAndSecondary`; `ReconciledValueDisplayTests.swift` | Cover records, owner groups and the composer location |
| Xcode project | `macos/Provenencia.xcodeproj/project.pbxproj` | Lists each Swift file; no synchronized folders. A new file needs its entries |

**Precedent.** S13-06 ([#640](https://github.com/mendahu/provenencia/pull/640)) built the owner group on the same row model and table; copy how it threads a display-only flag through `ReconciliationRecordGroup` and tests it in `PersonDetailTests`. For a feature sheet on the kit presenter, copy `DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` (`.pvSheet(item:)` with a `PVPanel` body).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The sheet body is one new snowflake, `ProvenanceSheet` *(new)* in `macos/App/Features/Conclusions/ProvenanceSheet.swift`, taking a `ReconciledValueRowModel` and a close action. Nothing Person-only in it | R-13: Event and Place adopt it later; the row model is already kind-neutral | A Person-only view — fails R-13; a kit component — one caller today ([`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)) |
| D-2 | `ConclusionDetailRowsSection` owns the presentation in `.sheet` mode: `@State private var provenanceItem: ProvenanceSheetItem?` (`ProvenanceSheetItem` *(new)*, `Identifiable` on the row id, with `func row(in rows: [ReconciledValueRowModel]) -> ReconciledValueRowModel?` that returns the row with that id, or nil when there is none or it has no records) and one `.pvSheet(item:)`. The sheet content shows `item.row(in: rows)`; `.onChange(of: rows)` sets the item to nil when `row(in:)` returns nil (PRV-11), never during render. The section reads `@Environment(\.pvWidthClass)` and passes the sheet width in (D-10), and passes `navigation` into the sheet content with `.environment(navigation)` | One sheet per section, not 20; the live lookup keeps the sheet true across a refresh; explicit environment so the sheet's links navigate whatever SwiftUI forwards to sheet content | A snapshot `item:` of the row model — goes stale on refresh (AC-4.1); a `.sheet` per row — 20 presenters |
| D-3 | `ReconciledValueRow` gains `var openProvenance: () -> Void = {}`. In `.sheet` mode the slot holds `PVIconButton(.info, label: L10n.Conclusions.provenanceOpen(label), size: .sm, action: openProvenance)` when `!model.records.isEmpty`, else stays `Color.clear`; the Why disclosure button and `ReconciliationReasoningView` are not drawn. `.why` mode is unchanged. The one `.sheet` caller, `ConclusionDetailRowsSection`, always passes the closure (D-2) | Keeps the `Provenance` enum a plain value that `ConclusionDetailRowsSection` and `ReconciliationRecordRows` also take (D-12) | An enum case carrying a closure (review M-10) — the section and the rows view take `Provenance` too and have no closure to give |
| D-4 | `ReconciliationRecord` gains `sourceLocation` (`.sourcePage(sourceId: outcome.sourceID, ref: nil, title: outcome.sourceTitle)`: the outcome carries no Source ref), `citationLocation` (`composerLocation(for: outcome, focusingObservation: false)`), `observationRef`, `subjectRef`, `groupRank: Int` (D-13) and `citationLine` (D-17). `composerLocation(for:)` gains `focusingObservation: Bool = true`, so one builder makes both composer places and the gallery tiles keep calling it unchanged; `location` keeps its meaning (the Observation in focus) and its name | All from fields the outcome carries; one composer-location builder (review M-8, X-6) | A second composer builder beside `composerLocation(for:)` |
| D-5 | A link closes the sheet first, then calls `navigation.go(to:)`. The sheet is not a place: Back lands on the Person page on Vitals with the sheet closed | Sheets are not in `WorkspaceLocation` anywhere in the app; a modal over a navigation would leave it orphaned | Reopening the sheet on Back — needs a new location field for one view |
| D-6 | The owner group line moves with the grid into `ReconciliationRecordRows` (D-12) and stays driven by data: a line where `group.owner` is set. `ReconciliationReasoningView` keeps its `groups:` input and passes it on; it loses its own grid and `ownerLine` | No `.why` caller builds owners today (`EventDetailContent.placeRow` passes none), so the accordion draws none; one table for both callers (review M-2) | Switching the accordion to a flat `records:` input — a signature change the Event and Place redesign would undo |
| D-7 | `ReconciliationRecordGroup.OwnerLine` gains `shown: Bool`; `derived(…)` sets it on the shown owner's line when there are several groups. `spoken` appends the new key `L10n.Conclusions.provenanceOwnerShown` ("%@, shown value") when `shown` | S13-06 UX-3: VoiceOver names which group is behind the shown value; the model already knows the shown owner | A view-side check of "first group" — the first group is not always the shown one (no shown owner when its records are empty) |
| D-8 | No new RPC. `GetConclusionDetail`'s outcomes gain the citation line's parts and their value group: `conclusiondetails.Outcome` gains `ArtifactLabel`, `CitationLabel`, `ArtifactPage int` (0 for none), `PageLabel` and `GroupRank int`. The parts come from adding `COALESCE(a.label, '')`, `COALESCE(c.label, '')` and `c.locator_json` to `sqlOutcomes` and reading the page with `locator.Page` *(new)*. `ReconcilerOutcomeDetail` gains `artifact_label = 21`, `citation_label = 22`, `artifact_page = 23`, `page_label = 24`, `group_rank = 25` | Working rules "A derived row carries its owning entity's records" (no new provenance RPC) and "Engine PRs run ahead of their slice" (grouping in Go, placement in Swift; user, 2026-10-31: Go groups the records per value). The joins are already in the query | A composed string from Go — the page words ("p.", "image") are localized copy; grouping in Swift — the working rule puts grouping in Go (user, 2026-10-31) |
| D-16 | `locator.Page(locatorJSON string) (artifactPage int, pageLabel string)` *(new)* in `core/locator/depiction.go`, beside `ValidateDepictionRegion`. The Citation-page step of `ValidateDepictionRegion` moves into one unexported `citationPage(doc document) (page int, label string, ok bool, err error)` that runs after `validateDoc` and keeps the rule "the Citation locator has at most one page selector"; `ValidateDepictionRegion` and `Page` both call it. `Page` returns the page's `artifact_page` and trimmed `page_label`, and `0, ""` on any error (a document `decode` or `validateDoc` refuses, longer than `MaxDocumentBytes`, two page selectors, or a page selector with no `artifact_page`) or when there is no page selector. A bad locator leaves the line without a page and the read goes on | One reader of a Citation's page, with the rules the package already states (review M-1) | A second reader that skips `validateDoc` and takes the first page; parsing in SQL with `json_extract`; failing the read on a bad locator — one bad Citation would blank the Person page |
| D-17 | The citation line is `CitationLineDisplay.line(_ outcome:)` *(new)* in `macos/App/Features/Conclusions/CitationLineDisplay.swift`: the parts Source title (else the Observation ref, as `ReconciliationRecord.sourceTitle(for:)` does), Artifact label, Citation label, page; each trimmed and dropped when empty; joined with `L10n.Conclusions.citationLineSeparator` (", "). The page is `L10n.Conclusions.citationPagePrinted(label)` ("p. %@") when `pageLabel` is not empty, else `citationPageImage(n)` ("image %lld") when `artifactPage` > 0, else nothing. The composer's viewer keeps "Page %lld" (`L10n.CitationComposer.locatorPage`): it names which page of the file is open, not a citation (user, 2026-10-31) | One place to replace when the formatter of [`docs/ideas/citation-formats.md`](../../../ideas/citation-formats.md) lands; "image N" is the *Evidence Explained* word for a page of a digital image | "page N" in the line — reads as a printed page; "Image N" in the composer (user, 2026-10-31) |
| D-9 | Tests are Swift Testing unit tests on the models (`PersonDetailTests`, a new `ProvenanceSheetTests` *(new)*); the view is checked by hand | Repo practice: no snapshot or UI tests for feature views ([`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)) | XCUITest — not used in the repo |
| D-10 | From the board (turn 6): the control is `PVIconButton(.info, label:, size: .sm)` with the default neutral tone (the board's "ghost"; PRV-1). The sheet closes with a primary `PVButton` "Done" in `PVPanel`'s sunken footer, Escape, or a click outside, which `.pvSheet` already does (PRV-4). The shown owner's line ends "· shown" in italic, muted (PRV-8). The citation line opens the Citation; under it two `PVButton` `.link` `.sm` links, "Source" and "Observation OBS-7QH3D"; the Record column holds the Observation and Subject refs as plain mono caption (PRV-9). The sheet is 880pt wide at Regular and 1280pt at Wide; at Wide Read as is capped at 220pt and Outcome at 280pt and the Citation column takes the rest (PRV-W1) | The agreed board | — |
| D-11 | The form is `.pvSheet` at both widths; no docked panel at Wide (brief domain fact "Form") | Event and Place, the next callers, have no secondary column; one presenter, one focus path (S13-D2 review M-5, X-12) | A docked panel at Wide — a width-switching presenter with one caller |
| D-18 | The citation line is a local cell inside `ReconciliationRecordRows`, drawn in `.sheet` mode in the Citation column: a `Button` with `.buttonStyle(.plain)` around `Text(record.citationLine)` in body small, its ink by PRV-5's rule (primary when counted, else secondary), underlined in that ink at 35% opacity, turning `PVColor.textLinkHover` on hover; it wraps (`fixedSize(horizontal: false, vertical: true)`) and is never truncated, and the text stays selectable (PRV-15, board turn 6 intro). In `.why` mode the column keeps today's Source title `PVButton` `.link` | A `PVButton` `.link` paints `PVColor.textLink` and would hide the counted / not-counted ink the table reads by; the cell lives with the grid, so the table stays one table (review M-3) | `PVButton` `.link` — loses the ink rule; a cell builder passed from `ProvenanceSheet` — a slot with one caller; a new kit variant — one caller |
| D-19 | Focus inside the sheet runs in reading order, record by record: the citation line, then "Source", then "Observation", then the next record; "Done" last (PRV-12). The board's note on 6l lists Source → Observation → Citation; reading order wins, as PRV-12 says | PRV-12 | The board note's order — Tab would jump back up to the line |
| D-12 | One record rows view, `ReconciliationRecordRows` *(new)* in `macos/App/Features/Conclusions/ReconciliationRecordRows.swift`, extracted from `ReconciliationReasoningView`'s grid. Inputs: `groups: [ReconciliationRecordGroup]`, `style`, `provenance: ReconciledValueRow.Provenance`. Under `.why`: Read as · Source · Outcome · 88pt reserved, owner lines where set (today's table). Under `.sheet`: Read as · Citation (D-18) with "Source" and "Observation" links under it · Outcome · Record (the Observation and Subject refs) · 88pt reserved, owner lines, and value headings (D-13). `ReconciliationReasoningView` wraps it in its sunken Card; `ProvenanceSheet` wraps it in a `ScrollView` capped at `ProvenanceSheet.recordsMaxHeight` *(new, 560pt)* as `PVDeleteImpactGroupStack` caps its list, with the column headings above the scroll | One table for both callers until the accordion is deleted; the split is the existing enum, so Event and Place delete one case when they adopt the sheet (review M-1 of the brief, M-2, M-3, X-4) | Two Bool switches on the rows view; a second grid in the sheet |
| D-13 | Go decides which value each record counts under: in `conclusiondetails` `finish`, each outcome's `GroupRank` is its `ValueRank` when its reason is `kept` or `folded` and that value is displayed (`ReconciledValue.Displayed()`), else 0; `finish` then orders outcomes by `GroupRank` (0 last), then `ValueRank`, then Observation id. Swift places headings over those runs: `ReconciliationRecordGroup` gains `values: [ValueGroup]` *(new)*, one per run of equal `groupRank` in wire order, each with `heading: Heading?`, where `Heading` is `.value(ReconciledValueRowModel.OtherValue)` (the value and its Source count, built by the code that builds `otherValues`) or `.otherRecords`. Headings show only when the group has two or more runs; a group whose records all have `groupRank` 0 still shows "Other records" (frame 6i); a group with one value run shows none (frames 6d, 6h) | User, 2026-10-31: grouping in Go, per the working rule; one wording for a value line (review M-5); outvoted records carry their losing value's rank, so Go keys on the displayed value they went into (review X-1) | Grouping by `ValueRank` in Swift — puts outvoted records under a losing value's heading, and the working rule puts grouping in Go |
| D-14 | The summary line (PRV-4) is `ReconciledValueRowModel.provenanceSummary` *(new)*, computed once in each factory as `count` and `against` are: `init(field:)` joins `spokenState(field)`, `againstLine(field)` and the record count with `L10n.Conclusions.provenanceSummaryJoin`; `derived(…)` gives the record count, or with two or more owner groups `provenanceRecordsFromOwners(records:owners:kind:)`, the owner kind passed in by a new `derived` parameter `ownerKind: ProvenanceOwnerKind` *(new: `.birthEvent`, `.deathEvent`, `.place`)* that `PersonDetailContent.rows` sets per life row and `EventDetailContent.placeRow` sets to `.place`; `spanning` and `empty` give the record count, or nothing when there are no records. The row model stores no field (review M-4) | Only what the row knows; no new reconciler wording (S13-D2 review X-1, M-2) | Storing the field on the row model — a second copy of the outcomes, and meaningless on derived and span rows |
| D-15 | The owner line's visibility stays as S13-06 built it: a line when there are several owners or the one owner is not the shown one | Matches the code and the corrected PRV-7 (review X-4) | — |

**Rules this plan bends.** None. (The working rule "Engine PRs run ahead of their slice" (grouping happens in Go, placement in Swift) is followed: D-13, user 2026-10-31.)

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ProvenanceSheet` *(new)* | view | `ReconciliationReasoningView` | The accordion stays for Event and Place in its flat form; the sheet has a header, owner and value groups and three links. Shares `ReconciliationOutcome` and `ReconciliationRecord`; does not copy the grid code beyond the columns it adds |
| `locator.Page`, `citationPage` *(new)* | func | `core/locator/depiction.go` `ValidateDepictionRegion`'s page step | Extracted from it; both call `citationPage` (D-16) |
| `ReconcilerOutcomeDetail.artifact_label`, `citation_label`, `artifact_page`, `page_label`, `group_rank` *(new)*; `conclusiondetails.Outcome` fields, `CatalogReconcilerOutcome` fields *(new)* | proto fields, fields | `ReconcilerOutcomeDetail.source_title`, `value_rank`; `conclusiondetails.finish` | Extended, same message; the grouping joins the existing sort |
| `CitationLineDisplay` *(new)* | enum (static display helpers) | `ReconciliationRecord.sourceTitle(for:)`, `EventTitleDisplay`, `PlaceTitleDisplay` | The same display-helper shape as the title displays; `sourceTitle(for:)` becomes its first part |
| `ReconciliationRecordRows` *(new)* | view | `ReconciliationReasoningView`'s grid | Extracted from it (D-12), not beside it |
| `ReconciliationRecordGroup.ValueGroup`, `.Heading`, `.values` *(new)*; `ReconciledValueRowModel.provenanceSummary`, `ProvenanceOwnerKind` *(new)* | types, field | `ReconciliationRecordGroup`, `ReconciledValueRowModel.OtherValue`, `ReconciledValueDisplay.spokenState`, `againstLine` | Extended; the heading reuses `OtherValue`; the summary composes the existing words (D-13, D-14) |
| `ProvenanceSheetItem` *(new)* | type | `PVDeleteImpact`'s `item:` snapshot | The item is an id, not a snapshot (D-2) |
| `ReconciliationRecord.sourceLocation`, `.citationLocation`, `.observationRef`, `.subjectRef`, `.groupRank`, `.citationLine` *(new)*; `composerLocation(for:focusingObservation:)` | fields, parameter | `ReconciliationRecord.location`, `composerLocation(for:)` | Extended, not duplicated |
| `OwnerLine.shown` *(new)* | field | `OwnerLine` | Extended |
| `ReconciledValueRow.openProvenance` *(new)* | property | `ReconciledValueRow.Provenance` | Extended |
| `ReconciliationRecordRows` input `provenance`, `ProvenanceSheet.recordsMaxHeight` *(new)* | parameter, constant | `ReconciledValueRow.Provenance`, `PVDeleteImpactLayout.listMaxHeight` | Reuses the enum; the cap mirrors DeleteImpact's |
| `L10n.Conclusions.provenance*` *(new)* | keys | `L10n.Conclusions.why*` | The `why*` keys stay for the accordion; the sheet's copy is its own |

**Builds on.** `.pvSheet` and `PVPanel` (the one sheet presenter), `PVIconButton`, `PVButton` `.link`, `ReconciledValueDisplay` for every outcome word, S13-06's owner groups.
**Deletes.** `ReconciliationReasoningView`'s own grid and `ownerLine` (moved into `ReconciliationRecordRows`, D-6, D-12); the Citation-page scan inside `ValidateDepictionRegion` (moved into `citationPage`, D-16); the `Color.clear` placeholder in `.sheet` mode when the row has records. The "S13-11 puts the provenance control in the slot" comment on `ReconciledValueRow.Provenance` is rewritten to say what the slot holds.
**Next adds.** The Event and Place redesign passes `.sheet` and deletes `ReconciliationReasoningView`; a Reconciliation Claims PR fills the reserved block and column (PRV-10).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Sheet → Source page | `WorkspaceLocation.sourcePage` | Existing place; a deleted Source falls back to the Sources list (`SourcePageView` calls `navigation.fallbackToSectionRoot()`) |
| Sheet → composer on a Citation, no Observation | `CitationComposerEntry.init?(location:)` `.edit` | Existing route (the Why's Source link opens it today, with an Observation); a deleted Citation is the composer's existing case, unchanged by this PR |
| Sheet open across a refresh that drops the row | `ConclusionDetailRowsSection` | Closes (D-2, AC-4.2) |

### Scale

At a person with 20 Vitals rows and 30 records on one row: the sheet renders one row's records; no cache key or invalidation changes. The one read gains two text columns and one locator decode per outcome row; a person's detail has tens to a few hundred outcomes, and `decode` is bounded at `MaxDocumentBytes` (64 KB), already decoded on every Citation read. The row adds one IconButton.

### Edges

- **Assumptions.** `finish`'s new order (by `GroupRank`, 0 last, then `ValueRank`) also reaches the Event and Place accordion: an outvoted record whose losing value had a rank now sorts after the counted ones instead of among them; T-11 pins it. `derived(…)` puts the shown owner first only when `shownOwnerID` matches an owner; `shown` (D-7) is set by id, not by position. Test T-3 covers a shown owner with no records.
- **Neighbors.** A row with records but an empty value (all Denied) shows the control (AC-1.3); a row with no records does not (AC-1.2); one owner shows no group line (AC-3.3).
- **Removed inputs.** A row that leaves `rows` on refresh closes the sheet (AC-4.2).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** The record count per row is bounded by the reconciler's outcomes for one field, already drawn by the accordion today. A locator is bounded by `MaxDocumentBytes` before it is unmarshalled (D-16); a locator with several page selectors gives the first; labels are shown as written and wrap (PRV-15).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/locator` | `func Page(locatorJSON string) (int, string)` *(new)* and unexported `citationPage(doc document) (int, string, bool, error)` *(new)* in `depiction.go`; `ValidateDepictionRegion` calls `citationPage` — D-16 |
| `core/database/conclusiondetails` | `Outcome` gains `ArtifactLabel`, `CitationLabel`, `ArtifactPage`, `PageLabel`, `GroupRank`; `sqlOutcomes` selects `COALESCE(a.label, '')`, `COALESCE(c.label, '')`, `c.locator_json`; `loadOutcomes` scans them and calls `locator.Page`; `finish` sets `GroupRank` and orders by it — D-8, D-13 |

Errors: none new; a bad locator yields no page (D-16).

### FFI

`api/proto/engine.proto`: `ReconcilerOutcomeDetail` gains `string artifact_label = 21; string citation_label = 22; int32 artifact_page = 23; string page_label = 24; int32 group_rank = 25;` (`group_rank`: the displayed value the record went into, 0 for none). `api/ffi/handlers/conclusion_detail.go` `reconcilerOutcomeProto` fills them. Regenerate with `./scripts/generate-proto.sh`. Swift: `CatalogReconcilerOutcome` gains `artifactLabel`, `citationLabel`, `artifactPage: Int`, `pageLabel`, `groupRank: Int` (defaults `""` / 0 so `FakeStore` and test fixtures keep compiling), mapped in `GoStore`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Vitals row | snowflake, extended | `ReconciledValueRow` | D-3 |
| Provenance control | `PVIconButton(.info, …, size: .sm)` | the row's slot | D-10 |
| Sheet | `.pvSheet(item:)` + `PVPanel` (sunken footer, "Done") | `ConclusionDetailRowsSection`, `ProvenanceSheet` *(new)* | D-1, D-2, D-10, D-11 |
| Record rows | snowflake, extracted | `ReconciliationRecordRows` *(new)* | D-12 |
| Citation line | local `Button(.plain)` | `ProvenanceSheet` | D-17, D-18 |
| Links, outcome, rules | `PVButton` `.link` `.sm`, `ReconciliationOutcome`, `PVDivider` | `ReconciliationRecordRows` | D-4 – D-7, D-10 |

**The frames in words** (board turn 6; the person is Margaret Ellen Robins, 20 Vitals rows):

- **6a, 6b (rows with the control, Regular and Wide).** Every Vitals row with records ends with the `info` IconButton in the 28pt slot; no row shows a Why button. Death place has no Place and no records: its slot is empty and the values still line up. The other-values disclosure, the Mixed Badge and the open arrow are where S13-D1 and S13-D12 put them. At Wide, Vitals is primary-only in the centred 1180pt column; nothing in the rows moves.
- **6c (birth date, two values, one owner).** An 880pt `PVPanel` over the page, title "Birth date". Header: the shown value in mono, the Mixed Badge, then the summary "4 records" (a life row knows only its record count, D-14). The table: headings Read as · Citation · Outcome · Record · (88pt reserved). A value line before each value's records ("3 Feb 1822 · 2 Sources" in the value's face, then the count), then "Other records" with the outvoted census record. One owner, so no owner lines. Footer: primary "Done".
- **6d (merged name).** Summary "merged from 3 Sources · 3 records". All records went into one value: no value lines; the ink carries it. Three citation lines: a will with a title only; a census with Source, Artifact and page; a baptism register with Source, Artifact, Citation label and page that wraps to two lines. Under each line, "Source" and "Observation OBS-7QH3D" as small links. Record column: the Observation ref above the Subject ref, mono caption, muted. The empty place for "Concluded by you" sits between the summary and the table at zero height (PRV-10).
- **6e (Mixed occupation).** Two values, each under its value line; the outvoted record and the weak record (its causes as the engine words them) under "Other records", last, secondary ink.
- **6f, 6g (two birth events, Regular and Wide).** Owner lines are the outer groups with a stronger rule between owners; value lines sit inside. The shown event first, its line ending "· shown" (italic, muted), spoken "Birth of Margaret Ellen Robins, EVT-4KX2M, shown value". The second event's line says nothing extra; its record is secondary ink. Summary "4 records from 2 birth events". At Wide the same sheet is 1280pt; Read as is capped at 220pt and Outcome at 280pt; the Citation column takes the rest. Resizing across 1600 keeps the sheet open on the row, its scroll and focus.
- **6h (death place, two Places).** Toronto's names first (shown), then Yorkville's, each under its Place line. The board draws a value line inside each Place; the build drops it, because each Place holds one value (D-13).
- **6i (birth date, all lost).** The header shows "No birth date recorded" in italic and no Badge. The baptism "denied by" the deposition and the deposition's negative record that eliminated it ("kept", primary ink) both sit under "Other records".
- **6j (30 records, scrolled).** The records scroll inside the sheet, capped at the window's height; the title, header, column headings and "Done" stay.
- **6k, 6l (focus).** Tab reaches the control with the kit focus ring; VoiceOver reads "Provenance of Birth date, button"; Space or Return opens the sheet. In the sheet, focus moves in reading order (D-19); each record reads as one element ("3 Feb 1822, Family Bible, Robins family, kept, PER-3HW8C") and its links as separate controls. Escape closes and focus returns to the control.
- **6m, 6n (dark).** 6a and 6f in dark mode, from the kit's tokens; no local colours.

States: a row with no records has no control; a row with records and no value shows the control and the sheet shows the empty text; 30 records scroll inside the sheet. Accessibility: the control's label names the row; the sheet title and each owner and value line are headings; each record is one element with its links separate; Escape closes; focus returns to the control.

### Strings

L10n keys *(new)* under `L10n.Conclusions`, with plural variations in `Localizable.xcstrings` as `sourceCount` has:

| Key | English | Where |
| --- | --- | --- |
| `provenanceOpen(_ label:)` | "Provenance of %@" | The control's tooltip and label (PRV-1) |
| `provenanceRecordCount(_ n:)` | "%lld record" / "%lld records" | Summary (D-14) |
| `provenanceRecordsFromOwners(records:owners:kind:)` | "%1$lld records from %2$lld birth events", "%1$lld records from %2$lld death events", "%1$lld records from %2$lld Places" (one key per `ProvenanceOwnerKind`) | Summary of a life row with several owners (D-14) |
| `provenanceSummaryJoin(_:_:)` | "%1$@ · %2$@" | Joins the summary's parts (D-14) |
| `provenanceOtherRecords` | "Other records" | Value heading (D-13) |
| `provenanceOwnerShown(_ line:)` | "%@, shown value" | Spoken owner line (D-7) |
| `provenanceOwnerShownMark` | "· shown" | Visible mark after the owner's ref (D-10) |
| `provenanceColumnCitation` | "Citation" | Column heading (D-12) |
| `provenanceColumnRecord` | "Record" | Column heading (D-12) |
| `provenanceSourceLink` | "Source" | Link (D-10) |
| `provenanceObservationLink(_ ref:)` | "Observation %@" | Link and its spoken label (D-10) |
| `provenanceDone` | "Done" | Footer Button (D-10) |
| `citationLineSeparator` | ", " | Citation line (D-17) |
| `citationPagePrinted(_ label:)` | "p. %@" | Citation line (D-17) |
| `citationPageImage(_ page:)` | "image %lld" | Citation line (D-17) |

Reused: `whyReadAs`, `whySource` and `whyOutcome` for the shared headings. Removed: none (`why*` keys stay for the accordion).

## 7. Steps

**STEP-1 — The engine sends each record's citation parts**
Files: `core/locator/depiction.go`, `core/locator/validate_test.go`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/conclusiondetails/conclusiondetails_test.go`, `api/proto/engine.proto` (+ generated code), `api/ffi/handlers/conclusion_detail.go`, `api/ffi/handlers/conclusion_detail_test.go`, `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`. Skill: `add-ffi-handler`.
Do: D-8, D-13 (the Go half: `GroupRank` and the order in `finish`), D-16; `./scripts/generate-proto.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/locator/... ./core/database/conclusiondetails/... ./api/ffi/handlers/...` passes, with T-8, T-9 and T-11, and `depiction_test.go` unchanged and green.

**STEP-2 — Records carry their citation line, three places and refs**
Files: `macos/App/Features/Conclusions/ReconciliationRecord.swift`, `macos/App/Features/Conclusions/CitationLineDisplay.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-swift-test`, `add-localized-string`.
Do: D-4, D-17.
Check: T-1 and T-10 pass.

**STEP-3 — The shown owner is marked**
Files: `ReconciliationRecord.swift` (`OwnerLine`), `ReconciledValueRowModel.swift` (`derived`, `ownerLine`), `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-7.
Check: T-2, T-3 pass.

**STEP-4 — The row shows the control and no Why in `.sheet` mode**
Files: `ReconciledValueRow.swift`. Skill: `add-ui-component`.
Do: D-3; rewrite the `Provenance` comment.
Check: T-6 passes; the Event page row still shows its Why (manual 4).

**STEP-5 — The section presents the sheet**
Files: `ConclusionDetailPage.swift` (`ConclusionDetailRowsSection`), `ProvenanceSheet.swift` *(new)*, `macos/ProvenenciaTests/ProvenanceSheetTests.swift` *(new)*, `project.pbxproj` (the app file in the App target, the test file in the ProvenenciaTests target). Skill: `add-swift-test`.
Do: D-2, D-5; the form and widths per D-10.
Check: T-4 passes; manual 1 and 5.

**STEP-6 — One record rows view, the accordion on it**
Files: `ReconciliationRecordRows.swift` *(new)*, `ReconciliationReasoningView.swift`, `ReconciledValueRow.swift`, `project.pbxproj`. Skill: `add-ui-component`.
Do: D-12, D-6.
Check: `rg -n "ownerLine|groups:" macos/App/Features/Conclusions/ReconciliationReasoningView.swift` prints nothing; manual 4 (the Event page's Why looks as before).

**STEP-7 — The sheet body**
Files: `ProvenanceSheet.swift`, `ReconciliationRecordRows.swift`, `ReconciliationRecord.swift` (`ValueGroup`), `ReconciledValueRowModel.swift` (D-13, D-14), `PersonDetailContent.swift` and `EventDetailContent.swift` (`ownerKind:`), `L10n.swift`, `Localizable.xcstrings`, `ProvenanceSheetTests.swift`. Skill: `add-localized-string`, `add-ui-component`.
Do: header and summary (D-14), value and owner groups (D-13) through `ReconciliationRecordRows`, the citation line (D-18), links and widths per D-10, focus order (D-19), the reserved block and column. Match frames 6c – 6n.
Check: T-5, T-7 pass; manual 1 – 3.

**STEP-8 — Docs**
Files: `docs/macos-client-patterns.md` (the row disclosure paragraph: the Person page's sheet, `ReconciliationRecordRows` shared by the accordion and the sheet, a record's three places, one sheet per section that follows its row by id), `completed.md`, `deployment-plan.md` (map row), this plan (status), `design/S13-D2-provenance-sheet.md` → `design/archive/`, `design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` on the touched docs reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `aRecordOpensItsSourceCitationAndObservation` *(new)* | An outcome with source `S1`, citation `C1`, observation `O1` → `sourceLocation` is `.sourcePage(sourceId: "S1", …)`; `citationLocation` has `citationId` `C1` and `observationId` nil; `location` has `observationId` `O1`; refs copied | AC-2.1 – AC-2.3, D-4 |
| T-2 | Swift | `PersonDetailTests.swift` `theShownOwnerIsSpokenAsShown` *(new)* | Two birth events, `EVT-4KX2M` shown → first group's `owner.shown` true and `spoken` "Birth of Ada Lovelace, EVT-4KX2M, shown value"; second group's `shown` false | AC-3.1, AC-3.2, D-7 |
| T-3 | Swift | `PersonDetailTests.swift` `oneOwnerHasNoGroupLine` *(new)*; extend `anOwnerThatIsNotTheShownOneIsNamed` | One owner → `owner` nil; a shown owner with no records and one other owner → the other's line has `shown` false | AC-3.3, D-7 |
| T-4 | Swift | `macos/ProvenenciaTests/ProvenanceSheetTests.swift` *(new)* `theSheetFollowsItsRowAcrossARefresh` | `ProvenanceSheetItem(id: "life.birthDate")` against rows with 3 then 4 records → resolves to the 4-record row; against rows without it → nil | AC-4.1, AC-4.2, D-2 |
| T-5 | Swift | `ProvenanceSheetTests.swift` `valuesGroupTheirRecordsAndTheRestComeLast` *(new)* | Records in wire order with `groupRank` 1, 1, 2, 0 (a Mixed name) → three value groups with headings `.value` (rank 1, its `OtherValue`), `.value` (rank 2), `.otherRecords`; all `groupRank` 1 (merged name) → one group, heading nil; all `groupRank` 0 (all-lost birth date) → one group, heading `.otherRecords`; two Place owners each all `groupRank` 1 → each owner one group, heading nil | D-13 |
| T-6 | Swift | `PersonDetailTests.swift` `onlyRowsWithRecordsOfferProvenance` *(new)* | A person with no birth event → the birth date row's `records` is empty; a birth date whose only record is Denied → `lead` nil and `records.count` 1 | AC-1.2, AC-1.3, D-3 |
| T-7 | Swift | `ProvenanceSheetTests.swift` `theSummarySaysOnlyWhatTheRowKnows` *(new)* | A merged name, 3 Sources, 1 against, 4 records → "merged from 3 Sources · 1 record disagrees · 4 records"; a life row with 2 owners and 4 records → "4 records from 2 birth events" | D-14, AC-1.1 |
| T-8 | Go | `core/locator/validate_test.go` `TestPage` *(new)*, table-driven | `{"version":1,"selectors":[{"type":"artifact"},{"type":"page","artifact_page":37,"page_label":" 23 "}]}` → 37, "23"; artifact selector only → 0, ""; two page selectors → 0, "" (at most one); `{"type":"page"}` with no `artifact_page` → 0, ""; `""`, `"{"`, version 2, and a 65,537-byte document → 0, "" | D-16 |
| T-9 | Go | `conclusiondetails_test.go` `TestForEntityCitationParts` *(new)*: its own Source with `artifactstest.Create` labelled "Vol. 3" and a Citation through `citations.CreateWithObservations` with `Label: "Baptism of Ada"` and locator page 37 (not the shared `fixture.source` and `const locator`, which other tests read); `conclusion_detail_test.go` `TestConclusionDetailProtoCarriesVoteArtifactAndMembers` (extend) | The outcome carries `ArtifactLabel` "Vol. 3", `CitationLabel` "Baptism of Ada", `ArtifactPage` 37, `PageLabel` ""; `reconcilerOutcomeProto` copies all five new fields into 21 – 25 | AC-2.4, D-8 |
| T-10 | Swift | `PersonDetailTests.swift` `aRecordNamesItsCitation` *(new)* | Source "Register of Baptisms", Artifact "Vol. 3", Citation "Baptism of Ada", page 37 → "Register of Baptisms, Vol. 3, Baptism of Ada, image 37"; page label " 23 " → "…, p. 23"; only a title → "Register of Baptisms"; empty title → the Observation ref first | AC-2.4, D-17 |
| T-11 | Go | `conclusiondetails_test.go` `TestForEntityGroupsOutcomesByShownValue` *(new)* (or extend `TestForEntity`) | The name fixture in `TestForEntity` (kept and folded into rank 1, outvoted with `ValueRank` 2, no-evidence) → kept and folded `GroupRank` 1; outvoted and no-evidence `GroupRank` 0; order: the `GroupRank` 1 outcomes first, then the rest | D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonDetailTests` (owner groups), `ReconciledValueDisplayTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`ReconciliationReasoningView.ownerLine`, `ReconciliationReasoningView(… groups:`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated: `docs/macos-client-patterns.md` (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. (AC-1.1, AC-1.2, AC-1.3) Open a person with a merged name (three records), a Mixed birth date and no death event. On Vitals, each row with records shows the control and no Why button; a row with nothing recorded shows an empty slot. Click the name's control: the sheet lists three records with outcomes and refs. Press Escape: focus is back on the control.
2. (AC-2.1, AC-2.2, AC-2.3) From a record, click Source, then Back; click Citation, then Back; click Observation, then Back. Each opens its place and Back lands on Vitals with the sheet closed.
3. (AC-3.1, AC-3.2) Open a person with two birth events (S13-D12's fixture): the birth date's sheet shows two groups, the shown one first and marked; VoiceOver reads it "…, shown value".
4. (AC-1.4) Open an Event page: its rows show the Why disclosure and accordion as before.
5. (AC-2.4) On a person whose birth record sits on a Citation with a label and a page, the birth date's sheet shows the citation line "Source title, Artifact label, Citation label, image N"; a long line wraps and is never cut.
6. Resize the window across 1600pt with the sheet open: it stays open on the same row (PRV-W2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The board asks for data the wire lacks (the Source type) | Out: the board note becomes a follow-up engine PR; this PR ships the sheet without it |
| A researcher expects a formatted standard citation | The line is the first step; the rest is in [`docs/ideas/citation-formats.md`](../../../ideas/citation-formats.md) |
| Focus does not return to the control after Escape | Manual 1; stop condition below |

**Stop and report if:**

- `ReconciliationReasoningView`, `ReconciledValueRow` or `ReconciledValueRowModel.derived` no longer match §4.
- Another caller of `ReconciliationReasoningView` passes owner groups (D-6 would drop them).
- `.pvSheet(item:)` cannot follow a row id across a refresh without a stale snapshot (D-2).
- The agreed board needs a field `CatalogReconcilerOutcome` does not carry beyond the citation line's parts.
- `sqlOutcomes` no longer joins `citations` and `artifacts`, so the parts need a second query (D-8).
- After Escape, focus does not return to the control, and no `PVSheet` change inside this PR's files fixes it.

## 12. Open questions

None.

## 13. PR

**Title:** S13-11 — Provenance sheet
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-11-provenance-sheet.md`.
