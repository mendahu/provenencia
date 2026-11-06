# S13-16 — Sources section

**Status:** Landed in #722 (2026-11-05)
**Spike / slice:** Spike 13 · Slice 6 — Sources
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D6](../design/archive/S13-D6-sources-section.md) (revision 2), board agreed 2026-11-05
**Board:** Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Sources.dc.html*, turn "S13-D6 rev 2 · Sources as Source rows, Remove and the last-record warning", frames 14a – 14y (this PR: 14a – 14m, 14v – 14x). The implementer may not be able to open it: **Design → macOS** describes every frame in words with its measures
**Depends on:** S13-01 (landed [#655](https://github.com/mendahu/provenencia/pull/655)), S13-10 (landed [#634](https://github.com/mendahu/provenencia/pull/634)), S13-13 (landed [#691](https://github.com/mendahu/provenencia/pull/691): `ConclusionPageSectionHost`), S13-D6 (board agreed 2026-11-05)
**Blocks:** S13-17, S13-28
**Branch:** `s13-16-sources-section`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-14 and S13-15 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (the section key the place warms), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`../person-detail.md` § 5 Sources](../person-detail.md#5-sources) (R-33 – R-37, R-2, R-13b); working rules "Each section reads its own query", "Claims are read per person from the graph cache, grouped by Source in Go", "Sections are places", "A full-width section keeps its states in the band", "Canonical only, filtered in Go"; brief [S13-D6](../design/archive/S13-D6-sources-section.md); the S13-01 plan (D-4, D-14: the type id and the Swift types), the S13-13 plan (D-10: the section host); the S13-36 plan (the provenance record row and table this section borrows)
**Planned:** 2026-11-05 against `main` @ `124d492`; refreshed 2026-11-05 against `4c642497` after the implementer's stop on D-11 (ship log, round 0)
**Landed:** the code follows the plan; four notes. (1) The refresh's T-11 had a marker inside the Timeline's hidden run fold away with its neighbors, against D-10's "no behavior change"; review round 1 reversed it, so the marker stays after "Show N more" as at `4c642497` (T-11 now says so). (2) Source rows fold at 50 (round 2, SEC-2): a throwaway measure showed 15 – 21 ms a row in Debug and a SwiftUI crash at 1,200 Sources. (3) Every `.link` button gains a focus ring, outset by the ring's width so it clears the glyphs, by the user's choice (round 2, MNT-17); the graph link no longer draws its own. (4) The app was not run: manual check items 2 (cold and warm graph open), 3 and 4 are unverified; only the pure offset, the viewport controller in a hosted scroll view and the model's selection are tested. (5) The compact header was removed at the user's call during the manual check, reversing S13-10 D-17: the header and switcher stay pinned in the full form at every scroll position.
**Review:** [`S13-16-sources-section.review.md`](S13-16-sources-section.review.md) — round 1 (card draft) Fits with changes / Not ready, every finding Fixed; round 2 (Source rows, board agreed) and the D-11 refresh in the log

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

A researcher opens a Person's Sources section and sees **the Sources behind the person**: one row per Source, laid out like a provenance sheet row (its title as the link to the Source page; its type, its publication date and its ref in their own columns), and under each Source, small and secondary, one line per record of the person on that Source's Evidence graphs: the graph's name (opening the graph with that Subject selected and in view), the graph's ref, the record's Subject name and ref. Today the section is an EmptyState ("Sources arrives in a later version of Provenencia."). The date column shows "—" until Spike 15 adds the Source date (user, 2026-11-05).

**Why now.** Slice 6's chrome, from the agreed board (S13-D6 revision 2; the Subject-card draft was rejected on 2026-11-05). The read (`ListPersonClaims`, S13-01) and its cache key (`.personClaims`, with its invalidation) have been in the app, unread, since #655; the section host (S13-13) makes the section a content view only. S13-17 adds Unpromote at the end of each graph line.

**What stays.** The engine: no Go, proto or store change. The Evidence graph canvas looks and behaves as today, except that a graph opened from a location naming a Subject now selects it and opens on it (STEP-7). The header, the switcher and the other four sections. The provenance sheet's look; its citation link style is shared, not changed.

## 2. User stories

**US-1.** As a researcher, when I open a Person's Sources section, I see the Sources the person's records come from, each with its key data, so that I can tell where everything I know about them was recorded.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann Weber with records from "Hamburg census 1867" (graphs Household 214 and Household 388), "Hamburger Adressbuch 1861" and an untitled Source SRC-4KQ7M (a letter) | A column header line "SOURCE · TYPE · DATE · REF", then three Source rows in the read's order: "Hamburg census 1867" with two graph lines (Household 214, then Household 388), "Hamburger Adressbuch 1861" with one, then SRC-4KQ7M (title in mono) with one (R-33, R-34, R-36; SS-1, SS-5) |
| AC-1.2 | The census row, its Source type id naming the type "Census", ref SRC-9TQ2M | Type column: the Source type pill "Census" with the type's icon; Date column: "—" in muted ink; Ref column: "SRC-9TQ2M" in mono (R-35; SS-3) |
| AC-1.3 | A Source whose type id is not in the source types list (not loaded yet, or deleted meanwhile) | The Type cell is empty; the row keeps its height and columns (D-6) |
| AC-1.4 | The Household 388 record: graph "Household 388", ref EVG-5MD8R, Subject "J. Weber", CPR-2M8XF | Its line reads "Household 388 · EVG-5MD8R · J. Weber · CPR-2M8XF" at caption size, refs in mono, indented 20pt under the title; a row too narrow cuts the Subject name first (SS-5) |
| AC-1.5 | A Source with 9 records (Hamburger Adressbuch 1855 – 1880) | Five graph lines, then "Show 4 more"; a click shows all nine. A Source with 6 records shows all six (SS-9; D-10) |
| AC-1.6 | No accepted record in the read (a delete raced the read) | The EmptyState "No records" · "This person has no promoted records.", id `persons.detail.sources.empty` (SS-7) |
| AC-1.7 | The read fails | The danger Callout "Sources could not be read" with the error, id `persons.detail.sources.error` (SS-8) |

**US-2.** As a researcher, from a Source row or a graph line, I open the Source page or the Evidence graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Click (or Return on) "Hamburg census 1867" | `navigation.go(to: .sourcePage(sourceId:ref:title:))` with its id, ref and title; Back returns to Sources (R-37; SS-2) |
| AC-2.2 | Click "Household 388" | `navigation.go(to:)` the location `SourcesListNavigation.canvasLocation(sourceId:sourceRef:sourceTitle:graphId:subjectId:)` builds (the Source's id, ref and title, the graph id, the Subject's id); the graph opens with that card selected and centered; Back returns to Sources (SS-5; D-5, D-11) |
| AC-2.3 | A graph with no ref | The line drops the graph ref and its separator; a graph with no name shows its ref as the link (D-7) |
| AC-2.4 | A graph opened from the Sources list, the Promote page or the provenance sheet (no Subject) | Opens as today: its saved camera, no selection change (D-11) |

**US-3.** As a researcher, the section loads and refreshes on its own, at both widths, and keeps room for How the records connect.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Open the Person page on Sources | The place warms `.personClaims` (the section's `queryKey`) and `.sourceTypesList` (its `supportingQueryKeys`); the section shows its spinner until the claims' first value (SS-8; D-1) |
| AC-3.2 | Promote another Subject into the person on a graph, then return | The new line is there; a reload keeps the rows drawn with "Refreshing" in the switcher row (S13-01's `refreshes`, unchanged) |
| AC-3.3 | The pane at 2296pt (Wide) | The section is primary-only and centered at the 1180pt primary measure, columns at their Regular widths; crossing 1600 keeps the scroll position and the focused link (SS-W1, SS-W2; D-9) |
| AC-3.4 | Regular or Wide | No How the records connect band or column is drawn: it is reserved on the board only (SS-11; D-9) |

## 3. Scope

| In | Covered by |
| --- | --- |
| A section declares its supporting keys; `.sources` reads `.personClaims` and warms `.sourceTypesList` | STEP-1, AC-3.1 |
| One name rule for a Subject, shared with the graph card | STEP-2, AC-1.4 |
| One label lookup for a source type id | STEP-3, AC-1.2, AC-1.3 |
| The content: Source rows, graph lines, folds, locations (SS-1 – SS-5, SS-9; R-33 – R-37) | STEP-4, AC-1.1 – AC-1.5, AC-2.1 – AC-2.3 |
| One ink link in the kit (`PVInkLink`), used by the provenance record row and the Sources title; one fold rule (`ConclusionFold`) for Family, Timeline and Sources | STEP-5, AC-2.1, AC-1.5 |
| The Sources section on the section host: the header line, rows, lines, fold, states, keyboard, VoiceOver, dark mode, Regular and Wide (SS-1 – SS-10, SS-W1 – SS-W3) | STEP-6, AC-1.6, AC-1.7, AC-3.3, AC-3.4 |
| A graph location's Subject: selected and centered on open; the unused pending-selection hook goes (user, 2026-11-05) | STEP-7, AC-2.2, AC-2.4 |
| Strings | STEP-8 |
| Docs | STEP-9 |

| Out | Home |
| --- | --- |
| Unpromote on a line, its confirm, landing on the Persons list (RC-1 – RC-7) | S13-17 |
| The Source date's value | [Spike 15's Source date](../../spike-15/source-provenance-date.md) |
| How the records connect | S13-28 (board S13-D10); its place is SS-11 |
| A record count per Source | not planned (the board's SS-4 answer, user 2026-11-05) |
| A copy button beside the title | not planned until Spike 15 formats a citation (board SS-2) |
| A provenance control on a row | not planned this spike (R-13b) |
| Subject cards, property rows, an Observation's citation | not planned (user, 2026-11-05) |
| A filter, search or sort | not planned (brief § 2.1) |
| Any Go, proto or store change | not planned: S13-01's read carries every field |
| Event and Place pages with a Sources section | their redesign (R-1) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| The Sources place | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` | `case .sources:` draws `PVEmptyState` with `L10n.Conclusions.pageSectionLater`, its last caller |
| Section keys | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection.queryKey(project:entityId:)` | One optional key: `.family → .personFamily`, `.timeline → .personEvents`, others nil. `fillsWidth`: `.timeline, .gallery` |
| Its readers | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` (`.personDetail`, `location.pageSection.flatMap { $0.queryKey(...) }`); `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `sectionRefreshing`; `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` `key` | The place warms the one key, the page's refresh follows it, the host reads it. Other places list supporting keys in `queryKeys` (`.sourceGraph` lists `.propertiesWorkspace`, `.sourceTypesList`) |
| Section host | `ConclusionPageSectionHost` | Spinner, error Callout (`<root>.error`), EmptyState (`<root>.empty`) around `body(content)`; `make` runs in the body, so a closure that reads other handles redraws when they change |
| Claims read | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonClaimGroup{sourceID, sourceRef, title, sourceTypeID, claims}`, `CatalogPersonClaim{claimID, graphID, graphRef, graphName, subject, observations}` | Groups in Source title order, claims in graph name order (`core/database/persondetail/claims.go`); `graphRef` empty for a graph with no ref |
| Claims key | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `case .personClaims` | `.sessionFresh`, evicted when hidden; refreshes on the person's touched handles, the graphs it holds, property labels |
| A Subject's name | `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift` `EvidenceGraphPlacedSubject.displayName(orders:)` | The graph card's name: the Event title for an event, else its first naming Observation (in `observationDisplayOrder`, private), else its label, else its ref; it reads only `kind`, `eventTitle`, `observations` and `subject` |
| Source types | `.sourceTypesList`; `macos/App/Platform/GenealogyStore.swift` `[CatalogSourceType].iconKey(forTypeID:)`; label copies in `macos/App/Features/Sources/SourcesModel.swift` `typeLabel(for:)` and `macos/App/Features/Sources/SourceIdentitySection.swift` `typeLabel` | Two copies of `types.first { $0.id == … }?.label ?? ""` |
| Source type look | `macos/App/Features/Catalog/SourceTypeMarkers.swift` `CatalogSourceTypePill(label:iconKey:)` | The Source page header's pill |
| Provenance row parts | `macos/App/DesignSystem/Recipes/ProvenanceSheet/PVProvenanceRecordRow.swift` `CitationLineButtonStyle` (private, its one input `counted`): body-sm, primary (else secondary) ink, underline at 35%, link ink on hover; the private `CitationLineButton` adds `@FocusState`, `.pvFocusRing`, wrapping and the AX label; `PVProvenanceColumns.spacing` = 14; `PVProvenanceParts.swift` `PVProvenanceColumnHeadings` (micro-caps column names); `PVProvenanceRecordTable` (`PVDivider()` between rows, `PVDivider(color: PVColor.borderStrong)` over a group) | The look the board borrows; the style is private, so a second caller would copy it |
| Kit table | `macos/App/DesignSystem/Components/Table/PVTable.swift` | A single-selection browse table (sort, filter, one line a row): not this section's shape |
| Feature columns | `macos/App/Features/Sources/SourcesListColumns.swift` | The Sources list's column widths (Type 170) and one cell modifier per column shared by its caption band and rows |
| Folds | `macos/App/Features/Conclusions/PersonFamilyContent.swift` `folded(_:_:)` (`foldLimit` 30); `PersonTimelineContent.swift` (its own `prefix` folds and `firstRevealed(by:)`) | Two hand-written copies of the fold rule |
| Title faces | `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` `PVTitleText(_:style:lineLimit:)`, `PVTitleSource` | `.name` plain, `.label` italic (a working label), `.ref` mono, per style |
| Location doc | `macos/App/Features/Workspace/WorkspaceLocation.swift` `subjectId` | Its comment names only the composer |
| Fold | `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` `ConclusionShowMoreRow(count:expand:)` | Family and Timeline's "Show N more" |
| Graph locations | `macos/App/Features/Sources/SourcesListNavigation.swift` `canvasLocation(for:graph:)` | Source's ref and title in `ref` / `title` / `sourceTitle`; no `subjectId` |
| Opening a graph | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `.sourceGraph` → `EvidenceGraphView(graphID:sourceID:session:store:userID:)`; `EvidenceGraphContent.canvasPane` passes `restoredCamera: model.session.graphCamera(for: graphID)`; `macos/App/Features/GraphCanvas/GraphCanvasScrollView.swift` `applyOpeningCamera(_:contentSize:)` on mount and on a `contentID` change | A location's `subjectId` is not read by the canvas; `EvidenceGraphView.body` shows `ProgressView` only while a query handle is missing. A warmed key's handle exists at once, in the loading state, so the canvas mounts with an empty snapshot. `model.prepare()` runs in the view's `.task`, after that mount, and its reads finish later. `makeNSView` defers `applyOpeningCamera` one main-queue turn, because the view has no frame until then. `updateNSView` applies a camera only on a `contentID` change, so nothing set in `prepare()` reaches the camera |
| Pending selection | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `pendingGraphSelections`, `setPendingGraphSelection(subjectID:graphID:)`, `consumePendingGraphSelection(graphID:)`; consumed in `EvidenceGraphModel` after `readyValue(graphKey)` | No caller of the setter: half dead |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys`; `EvidenceGraphModelTests.swift`; `ProvenanceSheetTests.swift`, `PVProvenanceColumnsTests.swift`; `GraphCanvasCoordinatesTests.swift`; `PersonFamilyContentTests.swift`, `PersonTimelineContentTests.swift` (their folds) | `sectionQueryKeys` expects `.sources` → nil |

**Precedent.** S13-13 ([#691](https://github.com/mendahu/provenencia/pull/691)) added the Timeline on `ConclusionPageSectionHost` with one key and a `…Content` built in `make`; copy that shape. S13-36 ([#702](https://github.com/mendahu/provenencia/pull/702)) moved the provenance row into the kit; its column constants are the precedent for this section's fixed columns.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ConclusionPageSection` keeps `queryKey(project:entityId:)` (the section's own read; `.sources → .personClaims`) and gains `supportingQueryKeys(project:entityId:) -> [CatalogQueryKey]` *(new)*: `.sources → [.sourceTypesList]`, every other section `[]`. Only `PlaceRegistry`'s `.personDetail` spec appends the supporting keys; the host and `sectionRefreshing` are unchanged. The section reads `.sourceTypesList` with `session.queryHandle` | The type says which key is the section's own; two of the three readers do not move; S13-28 appends to the supporting list (round 2 M-25) | Renaming to a list headed by the own key (an unwritten convention); a `session.query` in the section's `.task` |
| D-2 | A record's name: `EvidenceGraphPlacedSubject.displayName(subject:kind:observations:eventTitle:orders:) -> String` *(new, static)* in `EvidenceGraphSnapshot.swift`, which sorts the Observations with the private `observationDisplayOrder` and applies today's rule; the instance `displayName(orders:)` calls it. The Sources content calls it with `kind: .person` (a Person handle's members are person Subjects: `identityclaims.Membership.Kind` is the Subject type key) and `eventTitle: nil` | The line names the record exactly as its graph card does, with no made-up placed Subject and no Subject types read (round 2 M-23) | A placed Subject built with grid 0; `subject.trimmedLabel` (differs from the card and the S13-17 confirm) |
| D-3 | `PersonSourcesContent` *(new)* in `macos/App/Features/Conclusions/PersonSourcesContent.swift` *(new)*: `init(groups: [CatalogPersonClaimGroup], sourceTypes: [CatalogSourceType], orders: NameFormatOrders, expanded: Set<String>)`, `Equatable` (it holds formatted strings, so a name-format change is a content change), `isEmpty` when no row. `rows: [Row]`, `Row{id (sourceID), title: PVTitleSource (.name(title), else .ref(sourceRef)), helpTitle, typeLabel: String?, typeIconKey: String?, ref: String, accessibilityLabel, location: WorkspaceLocation, lines: [Line] (folded), hiddenCount: Int, firstRevealedLineID: String?}`; `Line{id (claimID), subjectID, graphID, graphTitle (graphName, else graphRef), graphRef: String? (nil when empty or used as the title), subjectName, subjectRef, accessibilityLabel, location: WorkspaceLocation}`. Ids are unique: Go groups by Source id and a claim id is one claim. `make` re-runs when `.sourceTypesList` loads (the host's body reads the handle), so a Type cell fills in when the list arrives | The Timeline and Family pattern; `graphID` for S13-17 (round 2 R-1) | Mapping in the view |
| D-4 | The Source row's columns: `PersonSourcesColumns` *(new, in the section file)* with `citation` flexible (`.infinity`, min 0), `type` 176, `date` 120, `ref` 104, spacing `PVProvenanceColumns.spacing` (14), and one `View` modifier per column (`sourcesCitationCell()`, `sourcesTypeCell()`, `sourcesDateCell()`, `sourcesRefCell()` *(new)*) that the header and every row apply, copying `macos/App/Features/Sources/SourcesListColumns.swift`'s shape. The header line is micro-caps (`pvMicroCaps()`, muted): `L10n.Conclusions.sourcesColumnSource`, `…Type`, `…Date`, `…Ref` *(new keys)*, followed by `PVDivider()`. Rows are 8pt top and bottom padding; between Source rows `PVDivider()` (the subtle hairline); the header's rule `PVDivider(color: PVColor.borderStrong)`. A row is one `HStack(alignment: .firstTextBaseline)` of the four cells, its graph lines under the citation cell's column, indented 20pt | The board's measures (14a: "Citation 1fr · Type 176pt · Date 120pt · Ref 104pt (14pt gaps)", 8pt padding, hairlines); the provenance table's tokens; one frame per column shared by the header and the rows, as `SourcesListColumns`' cell modifiers do (round 2 M-26) | `PVTable` (a selection browse table with one-line rows); `PVProvenanceColumnHeadings` (its columns are Read as · Citation · Outcome · Record) |
| D-5 | Locations through `SourcesListNavigation`: the title opens `pageLocation(sourceId:ref:title:)` *(new overload)*, which `pageLocation(for:)` now calls; a line opens `canvasLocation(sourceId:sourceRef:sourceTitle:graphId:subjectId:)` *(new overload)*, which `canvasLocation(for:graph:)` now calls with `subjectId: nil` | One place builds each Source location (round 1 M-5, round 2 M-27) | Building `.sourcePage` inline |
| D-6 | `[CatalogSourceType].label(forTypeID:) -> String?` *(new)* beside `iconKey(forTypeID:)`; `SourcesModel.typeLabel(for:)` and `SourceIdentitySection.typeLabel` move onto it (`?? ""`). Nil → an empty Type cell | One lookup (round 1 M-10) | A third copy |
| D-7 | The cells. **Citation:** `PVInkLink(ink: .primary, identifier:, action:) { PVTitleText(row.title, style: .rowLink, lineLimit: 2) }` with `.help(row.helpTitle)`: a titled Source in the row face, an untitled one's ref in mono (the title source's own mapping); no copy button. `PVTitleText.Style.rowLink` *(new)* is the `.row` face for a title inside a link: its `.ref` case keeps the mono font and drops `textSecondary`, so `PVInkLink`'s ink and hover reach an untitled Source's ref. `.name` and `.label` set no ink in either style. **Type:** `CatalogSourceTypePill(label:iconKey:)` when `typeLabel` is non-nil. **Date:** `L10n.Conclusions.sourcesDateNone` ("—") in caption, `PVColor.textMuted`. **Ref:** mono caption. **Line:** `PVButton(line.graphTitle, variant: .link, size: .sm)` (no padding, auto height, caption medium, link ink, underlined: `PVButton`'s `.link` branch), then " · " + `graphRef` (mono), " · " + `subjectName` (`.lineLimit(1)`, `.truncationMode(.tail)`, `.layoutPriority(-1)`, `.help` with the full name), " · " + `subjectRef` (mono), in caption `PVColor.textSecondary`; a missing part drops its separator | Board 14a, 14d (SS-2, SS-3, SS-5); the title source decides the face (round 2 M-22) | Hand-mapping `.ref` to mono in the cell; a `.label` title (a Source title is not a working label) |
| D-8 | `PersonSourcesSection(entityId:)` *(new)* in `macos/App/Features/Conclusions/PersonSourcesSection.swift` *(new)* on `ConclusionPageSectionHost<[CatalogPersonClaimGroup], PersonSourcesContent, PersonSourcesList>` with `errorTitle: L10n.Conclusions.sourcesErrorTitle`, `emptyTitle: L10n.Conclusions.sourcesEmptyTitle`, `emptyMessage: L10n.string(L10n.Conclusions.sourcesEmptyMessage)`; `make` reads the two lists' handle values, `session.nameFormat` and the section's `@State expanded: Set<String>` (reset on `entityId` change). `PersonSourcesList` *(new, same file)* takes `content`, `open: (WorkspaceLocation) -> Void` and `expand: (String) -> Void`. Ids: `persons.detail.sources.row.<sourceID>`, `.title.<sourceID>`, `.line.<claimID>`, `.graph.<claimID>`, `.more.<sourceID>` | The host owns the root, `.empty`, `.error` | A view that reads the claims handle itself |
| D-9 | Wide: `ConclusionPageSection.sources.fillsWidth` stays false: primary-only, centered at the primary measure (1180pt), columns at their Regular widths. Nothing is drawn for How the records connect at either width; S13-28 adds it in the secondary column at Wide and a band above the rows at Regular, which moves the rows down and nothing else | Board 14b, 14c, SS-11 (user, 2026-11-05) | A placeholder region in the app |
| D-10 | Fold: `ConclusionFold.shown(_ list:, limit:, minimumHidden:, expanded:) -> (shown: [T], hidden: Int)` and `ConclusionFold.firstRevealed(_ list:, limit:, minimumHidden:) -> T?` *(new, `list[limit]` when `list.count - limit >= minimumHidden`, else nil, expanded or not; the Timeline's two folds and Sources all read it)* in `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift`, beside the row: all of `list` when `expanded` or when `list.count - limit < minimumHidden`, else its first `limit`. Family (`limit: 30, minimumHidden: 1`) and Timeline (their limits, `minimumHidden: 1`) move onto it with no behavior change (the Timeline's dated fold threads markers between event rows, so it folds one list of the event rows, limit `eventFoldLimit`, and rebuilds its list from the result; only event rows count toward the limit; the first hidden event row becomes the one "Show N more", the other hidden event rows drop, and a marker among them stays in place, as today); Sources passes `limit: 5, minimumHidden: 2` to a Source's lines (more than six lines → five and "Show N more") and `limit: 50, minimumHidden: 1` (`PersonSourcesContent.rowFoldLimit`) to its rows, so 51 Sources show fifty and "Show 1 more"; the row fold is `@State` in the section, so it resets on each mount and a restored section opens folded, and the header line counts every Source. The fold key is the Source id. After unfolding, focus goes to the first revealed line's graph link (`firstRevealedLineID`), as the Timeline's `firstRevealed(by:)` does | One fold rule instead of three copies (round 2 M-24); board 14e (SS-9) | A third hand-written `prefix` |
| D-11 | A graph location's `subjectId` selects and centers that Subject or bridge on open, replacing the pending-selection hook. `WorkspaceDestinationHost` passes `navigation.currentLocation.subjectId` to `EvidenceGraphView(…, openingSubjectID: String? = nil)` *(new parameter)* → `EvidenceGraphModel(…, openingSubjectID: String? = nil)` *(new)*. The `.sourceGraph` place's `viewIdentity` becomes the graph id plus the `subjectId` when there is one, because the view is now built from it (`add-workspace-place`: declare the deep id a view is built from), so every location with a different Subject is a fresh mount. At the end of `prepare()`, where it consumes the pending selection today, the model takes `openingSubjectID` once (it sets the stored value to nil, so a re-run `.task` does not reselect over the user), calls `replacePrimarySelection([id])`, and looks the id up in `currentSnapshot()`: `subjects`, then `bridges`. It sets `openingFocus: CGPoint?` *(new, `private(set)`, never cleared)* to the grid-space center of `EvidenceSubjectCard.edgeFrame(for:)` or `EvidenceBridgeCard.contentFrame(for:in:)`, with no measured layout. An id in neither list (or a failed read, whose snapshot holds nothing): no selection, nil. The canvas mounts as it does today, at the saved camera, with `applyOpeningCamera`, `makeNSView`, `updateNSView` and the representable's parameters unchanged. `EvidenceGraphDocumentBody` adds `@Environment(\.graphCanvasViewport)` and `.onChange(of: model.openingFocus, initial: true)`, which calls `GraphCanvasViewportController.center(onGridPoint:)` *(new)* when the value is non-nil; `.onChange` fires once, because the value is set once. `center(onGridPoint:)` defers one `DispatchQueue.main.async`, so it runs after the mount's deferred camera, in queue order, once the view has a frame. It then reads the `GraphCanvasDocumentView`'s `contentOrigin` and `frame.size` and the live magnification, converts with `GraphCanvasCoordinates.documentPoint(_:origin:)`, and scrolls as `panByViewDelta` does (`clip.scroll(to:)`, `reflectScrolledClipView`). It keeps the magnification, which is the saved camera's, and the camera report follows as for a pan. The offset is a pure function, `GraphCanvasCoordinates.contentOffset(centering:visibleSize:contentSize:magnification:) -> CGPoint` *(new)*: `visibleSize` is the clip view's `frame.size` in view points, the function divides it by `magnification`, and it clamps with `clampedContentOrigin`. The caller passes `contentView.frame.size`, never `bounds.size`. `pendingGraphSelections`, `setPendingGraphSelection` and `consumePendingGraphSelection` are deleted, and `WorkspaceLocation.subjectId`'s doc comment says it names the Subject a graph or composer opens on | One way to open a graph on a Subject or bridge, for the Sources line and for the delete-impact rows that already emit this location (round 1 M-1, M-12; round 2 X-23 – X-25, M-28; user, 2026-11-05). Apply after mount: user, 2026-11-05, after the implementer's stop; through the viewport bridge: refresh M-32 | A second selection path; a camera set before navigating; holding the canvas mount until the focus resolves (user chose to apply after mount); a value-diff latch and a second opening-camera path in the representable (refresh M-32) |
| D-12 | VoiceOver: a Source row is one element, `L10n.Conclusions.sourcesRowAccessibility(title:type:date:ref:)` *(new)* with `date` = `sourcesDateNotRecorded` ("date not recorded"; never the "—" glyph): "Hamburg census 1867, Census, date not recorded, SRC-9TQ2M" (no type: `sourcesRowAccessibilityUntyped(title:date:ref:)`); its title link reads `sourcesOpenSource(_:)` ("Open Hamburg census 1867"); a line's link reads `sourcesOpenGraph(graph:subject:ref:)` ("Open Household 214, record Johann Weber, CPR-7Q2KD"), the line's other text hidden from VoiceOver. Tab: after the switcher, each row's title, then its lines' links (and S13-17's Unpromote), then "Show N more". Ids: the host's root `persons.detail.sources`, then D-8's | Board 14v (SS-10; round 2 X-28) | Reading every cell |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionPageSection.supportingQueryKeys` *(new)* | func | `queryKey`; the other places' `queryKeys` lists | Only the place reads it |
| `EvidenceGraphPlacedSubject.displayName(subject:kind:observations:eventTitle:orders:)` *(new)* | func | `displayName(orders:)` | Its body, lifted to a static the instance calls |
| `[CatalogSourceType].label(forTypeID:)` *(new)* | func | `iconKey(forTypeID:)`, two `typeLabel` copies | Replaces the copies |
| `SourcesListNavigation.canvasLocation(sourceId:…subjectId:)`, `pageLocation(sourceId:ref:title:)` *(new)* | func | `canvasLocation(for:graph:)`, `pageLocation(for:)` | Overloads; the old ones call them |
| `PVTitleText.Style.rowLink` *(new)* | enum case | `PVTitleText.Style` | The title owns its face; the style says the enclosing link owns the ink. No flag that means nothing at `.heading` |
| `PVInkLink` *(new)* in `macos/App/DesignSystem/Components/InkLink/PVInkLink.swift` *(new)* | component | the provenance row's private `CitationLineButton` + `CitationLineButtonStyle` | Promoted on its second caller (`.cursor/rules/design-system-layers.mdc`); the record row moves onto it |
| `ConclusionFold.shown` *(new)* | func | Family's `folded`, Timeline's folds | Replaces both copies |
| `PersonSourcesColumns` and its cell modifiers *(new)* | type | `SourcesListColumns`, `PVProvenanceColumns` | Different columns; copies `SourcesListColumns`' shape and the provenance spacing token |
| `PersonSourcesContent`, `PersonSourcesSection`, `PersonSourcesList` *(new)* | type, views | `PersonTimelineContent`, `PersonTimelineSection` | Same pattern |
| `GraphCanvasViewportController.center(onGridPoint:)`, `GraphCanvasCoordinates.contentOffset(centering:…)`, `EvidenceGraphView`/`EvidenceGraphModel` `openingSubjectID`, `EvidenceGraphModel.openingFocus` *(new)* | parameters, func | `panByViewDelta`, `replacePrimarySelection`, `pendingGraphSelections` | One opening path; replaces the hook. The viewport controller is already the canvas's imperative bridge |
| `L10n.Conclusions.sources*` *(new)* | keys | `timelineErrorTitle`, `ProvenanceSheet.column*` | One set per section |
| `PersonSourcesContentTests`, `ConclusionFoldTests` *(new)* | test files | `PersonTimelineContentTests` | One file per type |

**Builds on.** `ConclusionPageSectionHost`; `.personClaims` (S13-01); the provenance row's tokens and its link, promoted; `CatalogSourceTypePill`; `ConclusionShowMoreRow`; `SourcesListNavigation`; the canvas's opening camera.
**Deletes.** The `.sources` EmptyState; `L10n.Conclusions.pageSectionLater` and its `Localizable.xcstrings` entry; `WorkspaceSession.pendingGraphSelections`, `setPendingGraphSelection(subjectID:graphID:)`, `consumePendingGraphSelection(graphID:)` and the model's consume; the bodies of the two `typeLabel` copies; `CitationLineButton` and `CitationLineButtonStyle` (moved into `PVInkLink`); Family's `folded` and the Timeline's hand-written folds (moved onto `ConclusionFold`).
**Next adds.** S13-17 appends Unpromote to `Line`'s view and takes `Line.subjectName` and `Line.graphID` for its confirm; S13-28 appends its keys to `.sources`' `supportingQueryKeys` and fills the secondary column; Spike 15 replaces `sourcesDateNone` with the Source date in the Date cell; any surface that opens a graph on a Subject passes `subjectId`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Title → Source page | `macos/App/Features/Sources/SourcePageView.swift` | Existing; a Source deleted meanwhile falls back to the Sources list (today) |
| Line → graph with `subjectId` | `EvidenceGraphView`, `EvidenceGraphModel`, `GraphCanvasViewportController`, `PlaceRegistry` `.sourceGraph` | New (D-11). A Subject gone from the graph: no selection, today's camera. The place's `viewIdentity` carries the `subjectId`, so every arrival with one is a fresh mount and selects and centers, Back and restore included (T-7, T-12). The canvas shows the saved camera, then scrolls to the focus once the reads finish (T-10) |
| Delete-impact row → graph with `subjectId` | `PVDeleteImpact.activate`, `core/database/deleteimpact/projectors.go` `projectSubject` (no change to either) | Already emits this location, for Subjects and bridges, and `prepare()`'s selection ignored it. It now selects and centers (T-7) |
| Another person's page on Sources | `PersonSourcesSection` | The host's key carries the entity id; `expanded` resets |

### Scale

Brief § 2: typical 3 – 20 records from 2 – 12 Sources, large 50 from 14 (9 under one). The warm read issues no SQL (S13-01). `PersonSourcesContent` is rebuilt when the claims, either list, the name orders or `expanded` change: O(claims × observations) for the names (one sort per Subject, as the graph's snapshot does); a Type cell fills in when the list loads, without a spinner. Each row is a few text views; 50 lines is far under the Timeline's 60-row budget. Layout costs 15 – 21 ms a Source in a Debug build (300 Sources: 4.6 – 6.3 s; 1,200 crashed SwiftUI), so Source rows fold at 50 (D-10). No new write, key or invalidation; one extra warmed key, a catalog-wide list warm on any Sources visit.

### Edges

- **Assumptions.** The read's order is kept (rows by title, lines by graph name); the content does not re-sort (T-1). A Source's fold counts its lines (T-3); the section's fold counts Sources (T-13).
- **Neighbors.** 6 lines → no fold; 7 → five and "Show 2 more" (T-3). A graph opened with no `subjectId` → today's camera (T-7, T-8). A bridge named → selected and centered (T-7). A pan during the reads → the focus still wins, once. The canvas origin moves after the focus lands → `shiftVisibleOrigin` keeps the same grid point in view. A marker inside the Timeline's hidden run → it stays after "Show N more"; the fold counts event rows only (T-11).
- **Removed inputs.** Source title empty → ref in mono (T-1); type missing → empty cell (T-2); graph ref or name empty (T-2); the name always read from the Subject's Observations (T-4); opening Subject missing from the graph → no selection, no focus (T-7).
- **Many Sources.** 50 → all fifty; 51 → fifty and "Show 1 more"; 60 → fifty and "Show 10 more" (T-13). Unfolding focuses the 51st Source's first graph link.
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** The app's own catalog; 50 records and long titles are the stated large case (two lines then cut).

### Schema

None — no catalog change.

### Go core

None — S13-01's read carries every field.

### FFI

None —

### macOS

The board, in words (frames 14a – 14m, 14v – 14x; the implementer may not open it):

- **14a Regular · Sources.** Under the switcher, the micro-caps header line SOURCE · TYPE · DATE · REF (muted), a strong rule, then one row per Source: the title (body-sm, primary ink, underlined at 35%) in the flexible column, the type pill (icon and label) at 176pt, "—" muted at 120pt, the ref in mono at 104pt, 14pt gaps, 8pt row padding, a subtle hairline between Sources. Under the title, indented 20pt, one caption line per record: the graph name as a link (link ink, underlined), then "· graph ref · Subject name · Subject ref" in secondary ink, refs mono. The long Verzeichniß title wraps to two lines. (The dashed How the records connect band above the rows is frames only: D-9.)
- **14d One row's states.** Resting; hover on the title (link ink); keyboard focus on the title (the kit ring); hover and focus on a graph link; a narrow row cutting the Subject name ("Weber, Joh., Tisch" and an ellipsis) with the graph name and both refs at full width. A row never lifts or tints as a whole.
- **14e Large.** 50 records from 14 Sources; the Adressbuch shows five lines and "Show 4 more"; the untitled letter's title in mono.
- **14g – 14j States.** EmptyState with the library glyph; the section spinner; the danger Callout; rows kept with "Refreshing" in the switcher row.
- **14b, 14c, 14k – 14m Wide.** The 1180pt primary centered, columns at Regular widths; states centered in the same column. 14c shows the future secondary column for S13-28 (not built).
- **14v Keyboard and VoiceOver** (D-12). **14w, 14x Dark** — the same tokens in dark.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Section | `ConclusionPageSectionHost` | `macos/App/Features/Conclusions/PersonSourcesSection.swift` *(new)* | D-8 |
| Header line, rows, rules | `pvMicroCaps()`, `PVDivider`, `PersonSourcesColumns` | same file | D-4 |
| Title link | `PVInkLink` + `PVTitleText(.rowLink)` | `macos/App/DesignSystem/Components/InkLink/PVInkLink.swift` *(new)* | D-7 |
| Type | `CatalogSourceTypePill` | — | D-7 |
| Graph line | `PVButton(.link, .sm)` + caption text | same file | D-7 |
| Fold | `ConclusionFold`, `ConclusionShowMoreRow` | `ConclusionShowMoreRow.swift` | D-10 |
| States | the host's spinner, `PVCallout`, `PVEmptyState` (glyph `library`) | host | SS-7, SS-8, SS-W3 |

### Strings

L10n keys *(new)*, in `L10n.Conclusions`, catalog keys `conclusions.sources.<name>` (`sourcesErrorTitle` → `conclusions.sources.errorTitle`), each with a comment saying where it shows: `sourcesErrorTitle` = "Sources could not be read"; `sourcesEmptyTitle` = "No records"; `sourcesEmptyMessage` = "This person has no promoted records."; `sourcesColumnSource` = "Source"; `sourcesColumnType` = "Type"; `sourcesColumnDate` = "Date"; `sourcesColumnRef` = "Ref"; `sourcesDateNone` = "—"; `sourcesRowAccessibility(title:type:date:ref:)` = "%1$@, %2$@, %3$@, %4$@" and `sourcesRowAccessibilityUntyped(title:date:ref:)` = "%1$@, %2$@, %3$@"; `sourcesDateNotRecorded` = "date not recorded"; `sourcesOpenSource(_:)` = "Open %@"; `sourcesOpenGraph(graph:subject:ref:)` = "Open %1$@, record %2$@, %3$@". Removed: `conclusions.pageSection.later` (`pageSectionLater`).

## 7. Steps

**STEP-1 — A section declares its keys**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-workspace-place`.
Do: D-1.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PersonDetailTests` passes (T-6); Family and Timeline load as before.

**STEP-2 — One name rule for a Subject**
Files: `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift`, `macos/ProvenenciaTests/PersonSourcesContentTests.swift` *(new)*.
Do: D-2.
Check: T-4 passes; `EvidenceGraphModelTests` pass unchanged.

**STEP-3 — One source type label lookup**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Features/Sources/SourcesModel.swift`, `macos/App/Features/Sources/SourceIdentitySection.swift`.
Do: D-6.
Check: the Sources tests pass unchanged.

**STEP-4 — The section's content and locations**
Files: `macos/App/Features/Conclusions/PersonSourcesContent.swift` *(new)*, `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift`, `macos/App/Features/Sources/SourcesListNavigation.swift`, `macos/ProvenenciaTests/PersonSourcesContentTests.swift`, `macos/ProvenenciaTests/ConclusionFoldTests.swift` *(new)*.
Do: D-3, D-5, and D-10's `ConclusionFold.shown` (added here, in `ConclusionShowMoreRow.swift`, with T-9).
Check: T-1, T-2, T-3, T-9 pass.

**STEP-5 — One ink link, one fold rule**
Files: `macos/App/DesignSystem/Components/InkLink/PVInkLink.swift` *(new)*, `macos/App/DesignSystem/Recipes/ProvenanceSheet/PVProvenanceRecordRow.swift`, `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift`, `macos/App/DesignSystem/README.md`, `macos/App/Features/Conclusions/PersonFamilyContent.swift`, `PersonTimelineContent.swift`, `macos/ProvenenciaTests/PersonTimelineContentTests.swift`. Skill: `add-ui-component`.
Do: `PVTitleText.Style.rowLink` (D-7). `PVInkLink(ink: Ink (.primary, .secondary), accessibilityLabel: String? = nil, identifier:, action:, label:)` owns the record row's style (body-sm, ink, 35% underline, link ink on hover), `@FocusState`, `.pvFocusRing(cornerRadius: PVRadius.xs)`, wrapping and `pvActivatesOnReturn` (new on the record row, intended: a focused kit link activates on Return as a Timeline row and a Show more row do); the record row's citation line becomes `PVInkLink(ink: record.counted ? .primary : .secondary, …)`. Family and Timeline move onto STEP-4's `ConclusionFold`, `shown` and `firstRevealed`; T-11 is added here.
Check: T-5 passes; `ProvenanceSheetTests`, `PersonFamilyContentTests`, `PersonTimelineContentTests` pass unchanged; `scripts/check-guidance-names.py` passes.

**STEP-6 — The Sources section**
Files: `macos/App/Features/Conclusions/PersonSourcesSection.swift` *(new)*, `PersonDetailView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-4, D-7 – D-10, D-12, with their keys; replace the `.sources` case.
Check: manual check items 1 – 4; the build has no warning.

**STEP-7 — A graph opens on the Subject its location names**
Files: `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `EvidenceGraphModel.swift`, `macos/App/Features/GraphCanvas/GraphCanvasScrollView.swift`, `GraphCanvasCoordinates.swift`, `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `macos/App/Features/Workspace/WorkspaceLocation.swift` (`subjectId` comment), `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `macos/App/Features/GraphCanvas/README.md`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`, `GraphCanvasCoordinatesTests.swift`, `GraphCanvasViewportControllerTests.swift` *(new)*, `PlaceRegistryTests.swift`.
Do: D-11, deleting the pending-selection hook. The GraphCanvas README gains `contentOffset(centering:…)` in its coordinates list and `center(onGridPoint:)` beside the camera contract.
Check: T-7, T-8, T-10, T-12 pass; `rg 'PendingGraphSelection|pendingGraphSelections' macos` finds nothing; `rg -n 'graph or composer opens on' macos/App/Features/Workspace/WorkspaceLocation.swift` finds the comment.

**STEP-8 — Strings check**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: delete `pageSectionLater` once `rg pageSectionLater macos/App` finds only its definition; every new key has a catalog entry and comment.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `.agents/skills/add-workspace-place/SKILL.md` (a section's supporting keys), `.agents/skills/add-ui-component/SKILL.md` if it lists kit components.
Do: per the Definition of done. The brief stays open for S13-17.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-16-sources-section.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PersonSourcesContentTests.swift` `rowsKeepTheReadsOrder` *(new)* | The AC-1.1 groups → three rows in order, the census row's lines Household 214 then 388, the letter's title `.ref("SRC-4KQ7M")`; `[]` → `isEmpty` | AC-1.1, AC-1.6, D-3 |
| T-2 | Swift | `macos/ProvenenciaTests/PersonSourcesContentTests.swift` `cellsAndFallbacks` *(new)* | Type id in the list → `typeLabel` "Census" and icon key; unknown → nil; `graphRef` "" → nil and the graph name as title; graph name "" → the ref as title and `graphRef` nil; a titled Source → `.name`, untitled → `.ref`; `location` = `pageLocation(sourceId:ref:title:)`; line `location` = `canvasLocation(sourceId:sourceRef:sourceTitle:graphId:subjectId:)` with the Subject id, `graphID` the claim's; row accessibility label "Hamburg census 1867, Census, date not recorded, SRC-9TQ2M" | AC-1.2, AC-1.3, AC-1.4, AC-2.1, AC-2.2, AC-2.3, D-3, D-5, D-6, D-7, D-12 |
| T-3 | Swift | same · `foldPastSixLines` *(new)* | 9 lines → 5 shown, `hiddenCount` 4, `firstRevealedLineID` the 6th; 6 → 6, 0; 7 → 5, 2; Source id in `expanded` → all 9, `firstRevealedLineID` still the 6th | AC-1.5, D-10 |
| T-13 | Swift | same · `foldPastFiftySources` *(new)* | 60 Source groups → 50 rows, `hiddenRowCount` 10, `firstRevealedRowLineID` the first line of the 51st; 51 → 50, 1; 50 → 50, 0; `rowsExpanded: true` → 60, 0 | D-10 |
| T-4 | Swift | same · `lineNamesTheRecordAsItsCard` *(new)* | A claim with a `name` Observation "J. Weber" and label "jw" → `subjectName` "J. Weber", equal to `displayName(orders:)` of the same Subject placed by `EvidenceGraphSnapshot.build`; no naming Observation → the label; no label → the ref | AC-1.4, D-2 |
| T-5 | Swift | `macos/ProvenenciaTests/ProvenanceSheetTests.swift`, `PVProvenanceColumnsTests.swift` (existing) | Unchanged with the record row on `PVInkLink` | D-7 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys` (changed) | `queryKey` unchanged for Family and Timeline; `.sources` → `.personClaims`; `supportingQueryKeys`: `.sources` → `[.sourceTypesList]`, others `[]`; `fillsWidth` true only for Timeline and Gallery | AC-3.1, AC-3.3, D-1, D-9 |
| T-7 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `openingSubjectIsSelected` *(new)* | `openingSubjectID: "s2"` on a graph with s1, s2 → `selectedSubjectID == "s2"`, `openingFocus` = the center of `EvidenceSubjectCard.edgeFrame(for: s2)`; a bridge id → selected, `openingFocus` the center of `EvidenceBridgeCard.contentFrame(for:in:)`; `"gone"` → no selection, nil; nil → today; `prepare()` run a second time after the user selects s1 → s1 stays | AC-2.2, AC-2.4, D-11 |
| T-8 | Swift | `macos/ProvenenciaTests/GraphCanvasCoordinatesTests.swift` `contentOffsetCentersAFocus` *(new)* | `contentOffset(centering: (3000, 1200), visibleSize: 1000×800, contentSize: 4000×4000, magnification: 1)` → (2500, 800); magnification 2 → the visible rect (500×400 in content space) centered on the point; a focus at (100, 100) → clamped to (0, 0); at (3950, 3950) → clamped to the far edge | AC-2.2, AC-2.4, D-11 |
| T-9 | Swift | `macos/ProvenenciaTests/ConclusionFoldTests.swift` *(new)* `foldRule` | `shown([1…31], limit: 30, minimumHidden: 1, expanded: false)` → 30 shown, 1 hidden; 7 items, limit 5, minimum 2 → 5, 2; 6 items → 6, 0; expanded → all; `firstRevealed` on 7 items, limit 5, minimum 2 → the 6th; 6 items → nil | D-10 |
| T-10 | Swift | `macos/ProvenenciaTests/GraphCanvasViewportControllerTests.swift` *(new)* `centerOnGridPoint` | An `NSScrollView` hosted with `hostInBorderlessWindow(nsView:size:)` at 1000×800, `scrollerStyle = .overlay`, a `GraphCanvasDocumentView(pointer:frame:)` of 4000×4000, and a `GraphCanvasViewportController` holding it. After each call the test awaits one main-queue turn (`withCheckedContinuation` resumed from `DispatchQueue.main.async`). `center(onGridPoint: (3000, 1200))` → clip bounds origin (2500, 800). `contentOrigin` (-500, -500), grid (2500, 700) → the same. Magnification 2, grid (3000, 1200) → (2750, 1000), magnification still 2. A `scroll(to: .zero)` queued before the call → the focus wins | AC-2.2, D-11 |
| T-11 | Swift | `macos/ProvenenciaTests/PersonTimelineContentTests.swift` `markerInsideHiddenRunStaysShown` *(new)* | 60 dated events with a birth marker placed past the 50th → 50 event rows shown, `.more` with 10, the marker after `.more`, as before the move; expanded → the marker in place; `firstRevealed(by: .dated)` the 51st event | D-10 |
| T-12 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `viewIdentityResetsOnTheViewsOwnId` (extended) | The same graph with `subjectId` s1, s2 and none → three identities; a graph identity with no `subjectId` stays the graph id | D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: the Family and Timeline fold tests, `EvidenceGraphModelTests`, the provenance row tests.

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
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not this PR: S13-17 archives S13-D6
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected; `add-workspace-place` if it names `queryKey`; `macos/App/Features/GraphCanvas/README.md` per STEP-7
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Dogfood project, a Person with records from at least three Sources, one Source with two Evidence graphs. Sources: the header line, one row per Source in title order with its type pill, "—" and ref; under the two-graph Source, two lines (AC-1.1, AC-1.2, AC-1.4). Compare with board frame 14a.
2. Click a Source title: its Source page; Back: Sources. Click a graph line for a graph not opened since launch (the cold open): the graph centered on that Subject, its card selected; Back: Sources; click it again (the warm open): the same. Open a graph from the Sources list: it opens where it was (AC-2.1, AC-2.2, AC-2.4).
3. An untitled Source: its ref in mono, turning link ink on hover. A Source with seven or more records shows five and "Show N more"; click it (AC-1.5). A Person with more than 50 Sources shows fifty and "Show N more"; click it and focus lands on the 51st Source's first graph link; leave and return: folded again. Promote a new Subject into the person on a graph and return: the line is there (AC-3.2). In a preview whose FakeStore has no `seededPersonClaims` entry for the person (`listPersonClaims` throws `conclusiondetails.not_found`): "Sources could not be read" (AC-1.7).
4. Stretch across 1600pt and back: centered at 1180pt, scroll and focus kept, no How the records connect region (AC-3.3, AC-3.4). Tab order and VoiceOver as frame 14v. Tab to a graph link: it shows the focus ring and Return opens the graph. Tab to a `.link` button on the provenance sheet: it shows the same ring, and the button's layout does not move. A narrow pane: each line stays on one line, the Subject's name cut first. Unfold a Source with the keyboard: focus lands on the first line it added. Dark mode at both widths.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Renaming `queryKey` misses a caller | The compiler; T-6 |
| The renamed link style changes the provenance sheet | T-5; the style's body is unchanged |
| A line overflows on long names | D-7's cut order; manual item 1 with the Verzeichniß title and "Weber, Joh., Tischler" |

**Stop and report if:**

- Manual item 2 shows the saved camera after a cold or warm open: `center(onGridPoint:)` ran with `viewport.scrollView` nil, or the mount's deferred camera ran after it (D-11).
- The section host cannot take a `make` that reads two other handles without a redraw loop (D-8).
- The board's frames and this plan's measures disagree (14a's columns, 14e's fold).

## 12. Open questions

None.

## 13. PR

**Title:** S13-16 — Sources section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-16-sources-section.md`.
