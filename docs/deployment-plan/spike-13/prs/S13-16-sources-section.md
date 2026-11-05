# S13-16 — Sources section

**Status:** Planned
**Spike / slice:** Spike 13 · Slice 6 — Sources
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D6](../design/S13-D6-sources-section.md) (revision 2), board agreed 2026-11-05
**Board:** Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Sources.dc.html*, turn "S13-D6 rev 2 · Sources as Source rows, Remove and the last-record warning", frames 14a – 14y (this PR: 14a – 14m, 14v – 14x). The implementer may not be able to open it: **Design → macOS** describes every frame in words with its measures
**Depends on:** S13-01 (landed [#655](https://github.com/mendahu/provenencia/pull/655)), S13-10 (landed [#634](https://github.com/mendahu/provenencia/pull/634)), S13-13 (landed [#691](https://github.com/mendahu/provenencia/pull/691): `ConclusionPageSectionHost`), S13-D6 (board agreed 2026-11-05)
**Blocks:** S13-17, S13-28
**Branch:** `s13-16-sources-section`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-14 and S13-15 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (the section key the place warms), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`../person-detail.md` § 5 Sources](../person-detail.md#5-sources) (R-33 – R-37, R-2, R-13b); working rules "Each section reads its own query", "Claims are read per person from the graph cache, grouped by Source in Go", "Sections are places", "A full-width section keeps its states in the band", "Canonical only, filtered in Go"; brief [S13-D6](../design/S13-D6-sources-section.md); the S13-01 plan (D-4, D-14: the type id and the Swift types), the S13-13 plan (D-10: the section host); the S13-36 plan (the provenance record row and table this section borrows)
**Planned:** 2026-11-05 against `main` @ `124d492`
**Review:** [`S13-16-sources-section.review.md`](S13-16-sources-section.review.md) — round 1 (card draft) Fits with changes / Not ready, every finding Fixed; round 2 (Source rows, board agreed) in the log

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
| AC-3.1 | Open the Person page on Sources | The place warms `.personClaims`, `.propertiesWorkspace` and `.sourceTypesList` (the section's `queryKeys`); the section shows its spinner until the claims' first value (SS-8; D-1) |
| AC-3.2 | Promote another Subject into the person on a graph, then return | The new line is there; a reload keeps the rows drawn with "Refreshing" in the switcher row (S13-01's `refreshes`, unchanged) |
| AC-3.3 | The pane at 2296pt (Wide) | The section is primary-only and centered at the 1180pt primary measure, columns at their Regular widths; crossing 1600 keeps the scroll position and the focused link (SS-W1, SS-W2; D-9) |
| AC-3.4 | Regular or Wide | No How the records connect band or column is drawn: it is reserved on the board only (SS-11; D-9) |

## 3. Scope

| In | Covered by |
| --- | --- |
| A section declares every key it reads; `.sources` reads `.personClaims` and two lists | STEP-1, AC-3.1 |
| One per-Subject builder for a placed Subject (its name rule), shared with `EvidenceGraphSnapshot.build` | STEP-2, AC-1.4 |
| One label lookup for a source type id | STEP-3, AC-1.2, AC-1.3 |
| The content: Source rows, graph lines, folds, locations (SS-1 – SS-5, SS-9; R-33 – R-37) | STEP-4, AC-1.1 – AC-1.5, AC-2.1 – AC-2.3 |
| The provenance citation link style, shared | STEP-5, AC-2.1 |
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
| A Subject's name | `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift` `EvidenceGraphPlacedSubject.displayName(orders:)`; `EvidenceGraphSnapshot.build` | The graph card's name: its first naming Observation, else its label, else its ref. `build`'s loop body maps a Subject (kind from `EvidencePrimaryKind(rawValue: type.key)`, `typeLabel`, Observations sorted by the `private static` `observationDisplayOrder`, `isCited`, membership, event title) |
| Subject types | `macos/App/Features/Properties/PropertiesSnapshot.swift` `PropertiesSnapshot.types`, key `.propertiesWorkspace` | — |
| Source types | `.sourceTypesList`; `macos/App/Platform/GenealogyStore.swift` `[CatalogSourceType].iconKey(forTypeID:)`; label copies in `macos/App/Features/Sources/SourcesModel.swift` `typeLabel(for:)` and `macos/App/Features/Sources/SourceIdentitySection.swift` `typeLabel` | Two copies of `types.first { $0.id == … }?.label ?? ""` |
| Source type look | `macos/App/Features/Catalog/SourceTypeMarkers.swift` `CatalogSourceTypePill(label:iconKey:)` | The Source page header's pill |
| Provenance row parts | `macos/App/DesignSystem/Recipes/ProvenanceSheet/PVProvenanceRecordRow.swift` `CitationLineButtonStyle` (private): body-sm, primary ink, underline at 35%, link ink on hover; `CitationLineButton` adds `.pvFocusRing`; `PVProvenanceColumns.spacing` = 14; `PVProvenanceParts.swift` `PVProvenanceColumnHeadings` (micro-caps column names); `PVProvenanceRecordTable` (`PVDivider()` between rows, `PVDivider(color: PVColor.borderStrong)` over a group) | The look the board borrows; the style is private, so a second caller would copy it |
| Kit table | `macos/App/DesignSystem/Components/Table/PVTable.swift` | A single-selection browse table (sort, filter, one line a row): not this section's shape |
| Fold | `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` `ConclusionShowMoreRow(count:expand:)` | Family and Timeline's "Show N more" |
| Graph locations | `macos/App/Features/Sources/SourcesListNavigation.swift` `canvasLocation(for:graph:)` | Source's ref and title in `ref` / `title` / `sourceTitle`; no `subjectId` |
| Opening a graph | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `.sourceGraph` → `EvidenceGraphView(graphID:sourceID:session:store:userID:)`; `EvidenceGraphView` passes `restoredCamera: model.session.graphCamera(for: graphID)`; `macos/App/Features/GraphCanvas/GraphCanvasScrollView.swift` `applyOpeningCamera(_:contentSize:)` on mount and on a `contentID` change | A location's `subjectId` is not read by the canvas; the canvas mounts once the rows are loaded (`ProgressView` before) |
| Pending selection | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `pendingGraphSelections`, `setPendingGraphSelection(subjectID:graphID:)`, `consumePendingGraphSelection(graphID:)`; consumed in `EvidenceGraphModel` after `readyValue(graphKey)` | No caller of the setter: half dead |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys`; `EvidenceGraphModelTests.swift`; `PVProvenanceSheetTests.swift` and the record row tests | `sectionQueryKeys` expects `.sources` → nil |

**Precedent.** S13-13 ([#691](https://github.com/mendahu/provenencia/pull/691)) added the Timeline on `ConclusionPageSectionHost` with one key and a `…Content` built in `make`; copy that shape. S13-36 ([#702](https://github.com/mendahu/provenencia/pull/702)) moved the provenance row into the kit; its column constants are the precedent for this section's fixed columns.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ConclusionPageSection.queryKey(project:entityId:)` becomes `queryKeys(project:entityId:) -> [CatalogQueryKey]`, the section's own read first: `.family → [.personFamily]`, `.timeline → [.personEvents]`, `.sources → [.personClaims, .propertiesWorkspace, .sourceTypesList]`, others `[]`. `PlaceRegistry` appends them all; the host and `sectionRefreshing` read `.first` (the host's precondition: non-empty). The section reads the two lists with `session.queryHandle` | The place owns warming; one declaration; S13-28 appends its keys (round 1 M-6) | A `session.query` in the section's `.task` |
| D-2 | One per-Subject builder: `EvidenceGraphPlacedSubject.init(subject:kind:typeLabel:observations:gridX:gridY:membership:eventTitle:)` *(new)* in `EvidenceGraphSnapshot.swift` sorts with `observationDisplayOrder` (stays private to the file) and sets `isCited`; `build`'s loop calls it for a primary kind. The Sources content builds one per claim (kind from `PropertiesSnapshot.types` by `subject.subjectTypeID`, else `.person`; grid 0) only to read `displayName(orders:)` | The line names the record as its graph card does, one naming rule; the kind fallback is the one kind a Person's claim has (round 1 M-2) | `subject.trimmedLabel` (differs from the graph card and the S13-17 confirm) |
| D-3 | `PersonSourcesContent` *(new)* in `macos/App/Features/Conclusions/PersonSourcesContent.swift` *(new)*: `init(groups: [CatalogPersonClaimGroup], sourceTypes: [CatalogSourceType], properties: PropertiesSnapshot?, orders: NameFormatOrders, expanded: Set<String>)`, `Equatable`, `isEmpty` when no row. `rows: [Row]`, `Row{id (sourceID), title: PVTitleSource (.label(title), else .ref(sourceRef)), helpTitle, typeLabel: String?, typeIconKey: String?, ref: String, accessibilityLabel, location: WorkspaceLocation, lines: [Line] (folded), hiddenCount: Int}`; `Line{id (claimID), subjectID, graphTitle (graphName, else graphRef), graphRef: String? (nil when empty or used as the title), subjectName, subjectRef, accessibilityLabel, location: WorkspaceLocation}` | The Timeline and Family pattern; testable without a view | Mapping in the view |
| D-4 | The Source row's columns: `PersonSourcesColumns` *(new, in the section file)* with `citation` flexible (`.infinity`, min 0), `type` 176, `date` 120, `ref` 104, spacing `PVProvenanceColumns.spacing` (14). The header line is micro-caps (`pvMicroCaps()`, muted): `L10n.Conclusions.sourcesColumnSource`, `…Type`, `…Date`, `…Ref` *(new keys)*, followed by `PVDivider()`. Rows are 8pt top and bottom padding; between Source rows `PVDivider()` (the subtle hairline); the header's rule `PVDivider(color: PVColor.borderStrong)`. A row is one `HStack(alignment: .firstTextBaseline)` of the four cells, its graph lines under the citation cell's column, indented 20pt | The board's measures (14a: "Citation 1fr · Type 176pt · Date 120pt · Ref 104pt (14pt gaps)", 8pt padding, hairlines); the provenance table's tokens | `PVTable` (a selection browse table with one-line rows); `PVProvenanceColumnHeadings` (its columns are Read as · Citation · Outcome · Record) |
| D-5 | Locations: the title opens `.sourcePage(sourceId:ref:title:)`; a line opens `SourcesListNavigation.canvasLocation(sourceId:sourceRef:sourceTitle:graphId:subjectId:)` *(new overload)*, which `canvasLocation(for:graph:)` now calls with `subjectId: nil` | One place builds a graph location (round 1 M-5) | The graph's ref and name in the location |
| D-6 | `[CatalogSourceType].label(forTypeID:) -> String?` *(new)* beside `iconKey(forTypeID:)`; `SourcesModel.typeLabel(for:)` and `SourceIdentitySection.typeLabel` move onto it (`?? ""`). Nil → an empty Type cell | One lookup (round 1 M-10) | A third copy |
| D-7 | The cells. **Citation:** the title as a link, `Button` with `Text(verbatim: row.title.text)` (`.lineLimit(2)`, `.truncationMode(.tail)`; a `.ref` title in `PVFont.mono(size: PVTypeScale.bodySmall)`) styled by `PVCitationLinkButtonStyle(counted: true)` *(the record row's `CitationLineButtonStyle`, made internal and renamed in its file)*, `.pvFocusRing`, `pvActivatesOnReturn(when:)`, `.help(row.helpTitle)`; no copy button. **Type:** `CatalogSourceTypePill(label:iconKey:)` when `typeLabel` is non-nil. **Date:** "—" (`L10n.Conclusions.sourcesDateNone` *(new)*) in caption, `PVColor.textMuted`. **Ref:** mono caption. **Line:** caption size, `PVButton(variant: .link, size: .sm)` titled `graphTitle`, then " · " + `graphRef` (mono), " · " + `subjectName` (`.lineLimit(1)`, `.truncationMode(.tail)`, `.layoutPriority(-1)` so it is cut first, `.help` with the full name), " · " + `subjectRef` (mono), all `PVColor.textSecondary`; the separators drop with a missing part | Board 14a, 14d (SS-2, SS-3, SS-5): the provenance citation link, the type pill, the Subject cut first | A local link style (a copy of the record row's) |
| D-8 | `PersonSourcesSection(entityId:)` *(new)* in `macos/App/Features/Conclusions/PersonSourcesSection.swift` *(new)* on `ConclusionPageSectionHost<[CatalogPersonClaimGroup], PersonSourcesContent, PersonSourcesList>` with `errorTitle: L10n.Conclusions.sourcesErrorTitle`, `emptyTitle: L10n.Conclusions.sourcesEmptyTitle`, `emptyMessage: L10n.string(L10n.Conclusions.sourcesEmptyMessage)`; `make` reads the two lists' handle values, `session.nameFormat` and the section's `@State expanded: Set<String>` (reset on `entityId` change). `PersonSourcesList` *(new, same file)* takes `content`, `open: (WorkspaceLocation) -> Void` and `expand: (String) -> Void`. Ids: `persons.detail.sources.row.<sourceID>`, `.title.<sourceID>`, `.line.<claimID>`, `.graph.<claimID>`, `.more.<sourceID>` | The host owns the root, `.empty`, `.error` | A view that reads the claims handle itself |
| D-9 | Wide: `ConclusionPageSection.sources.fillsWidth` stays false: primary-only, centered at the primary measure (1180pt), columns at their Regular widths. Nothing is drawn for How the records connect at either width; S13-28 adds it in the secondary column at Wide and a band above the rows at Regular, which moves the rows down and nothing else | Board 14b, 14c, SS-11 (user, 2026-11-05) | A placeholder region in the app |
| D-10 | Fold: a Source with more than 6 lines shows 5 and `ConclusionShowMoreRow(count: hidden, expand:)`; 6 or fewer show all (no fold hides one line). The fold key is the Source id; unfolding keeps the scroll position | Board 14e (SS-9) | A new fold control |
| D-11 | A graph location's `subjectId` selects and centers that Subject on open, replacing the pending-selection hook. `WorkspaceDestinationHost` passes `navigation.currentLocation.subjectId` to `EvidenceGraphView(…, openingSubjectID:)` *(new parameter)* → `EvidenceGraphModel(…, openingSubjectID:)` *(new)*. The model selects it where it consumes the pending selection today (after its `readyValue(graphKey)` wait) when the snapshot holds the id. The canvas gets `openingFocus: CGPoint?` *(new, on `GraphCanvasScrollView`)* = that card's `EvidenceSubjectCard.edgeFrame(for:)` center; `applyOpeningCamera(_:contentSize:focus:)` *(new parameter)* keeps the restored magnification (else 1) and centers `focus` in the visible rect, clamped as today. `pendingGraphSelections`, `setPendingGraphSelection`, `consumePendingGraphSelection` are deleted | One way to open a graph on a Subject; history and restore pick it up; the canvas applies its camera at mount, after the rows load (round 1 M-1, M-12; user, 2026-11-05) | A second selection path; a camera set before navigating (the page has no positions) |
| D-12 | VoiceOver: a Source row is one element, `L10n.Conclusions.sourcesRowAccessibility(title:type:date:ref:)` *(new)* ("Hamburg census 1867, Census, date not recorded, SRC-9TQ2M"; with no type, `sourcesRowAccessibilityUntyped(title:date:ref:)`), its title link a separate button reading `sourcesOpenSource(_:)` ("Open Hamburg census 1867"); a line's link reads `sourcesOpenGraph(graph:subject:ref:)` ("Open Household 214, record Johann Weber, CPR-7Q2KD"), the line's text hidden from VoiceOver. Tab: after the switcher, each row's title, then its lines' links (and S13-17's Unpromote), then "Show N more" | Board 14v (SS-10) | Reading every cell |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionPageSection.queryKeys` *(rename of `queryKey`)* | func | `queryKey` | Extended to a list; three readers move |
| `EvidenceGraphPlacedSubject.init(subject:kind:…)` *(new)* | init | `EvidenceGraphSnapshot.build`'s loop body | Extracted from it |
| `[CatalogSourceType].label(forTypeID:)` *(new)* | func | `iconKey(forTypeID:)`, two `typeLabel` copies | Replaces the copies |
| `SourcesListNavigation.canvasLocation(sourceId:…subjectId:)` *(new)* | func | `canvasLocation(for:graph:)` | Overload; the old one calls it |
| `PVCitationLinkButtonStyle` *(rename of the private `CitationLineButtonStyle`)* | style | — | Made shareable, not copied |
| `PersonSourcesColumns` *(new)* | type | `PVProvenanceColumns` | Different columns; shares its spacing token |
| `PersonSourcesContent`, `PersonSourcesSection`, `PersonSourcesList` *(new)* | type, views | `PersonTimelineContent`, `PersonTimelineSection` | Same pattern |
| `GraphCanvasScrollView.openingFocus`, `applyOpeningCamera(…focus:)`, `EvidenceGraphView`/`EvidenceGraphModel` `openingSubjectID` *(new)* | parameters | `restoredCamera`, `pendingGraphSelections` | One opening path; replaces the hook |
| `L10n.Conclusions.sources*` *(new)* | keys | `timelineErrorTitle`, `ProvenanceSheet.column*` | One set per section |
| `PersonSourcesContentTests` *(new)* | test file | `PersonTimelineContentTests` | One file per content type |

**Builds on.** `ConclusionPageSectionHost`; `.personClaims` (S13-01); the provenance row's tokens and link style; `CatalogSourceTypePill`; `ConclusionShowMoreRow`; `SourcesListNavigation`; the canvas's opening camera.
**Deletes.** The `.sources` EmptyState; `L10n.Conclusions.pageSectionLater` and its `Localizable.xcstrings` entry; `WorkspaceSession.pendingGraphSelections`, `setPendingGraphSelection(subjectID:graphID:)`, `consumePendingGraphSelection(graphID:)` and the model's consume; the bodies of the two `typeLabel` copies; the name `queryKey`; the name `CitationLineButtonStyle`.
**Next adds.** S13-17 appends Unpromote to `Line`'s view; S13-28 appends its keys to `.sources`' `queryKeys` and fills the secondary column; Spike 15 replaces `sourcesDateNone` with the Source date in the Date cell; any surface that opens a graph on a Subject passes `subjectId`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Title → Source page | `macos/App/Features/Sources/SourcePageView.swift` | Existing; a Source deleted meanwhile falls back to the Sources list (today) |
| Line → graph with `subjectId` | `EvidenceGraphView`, `EvidenceGraphModel`, `GraphCanvasScrollView` | New (D-11). A Subject gone from the graph: no selection, today's camera. Back and restore re-select and re-center (intended). `PlaceRegistry` `.sourceGraph` keeps the model by `graphId`, so the opening Subject applies on a fresh mount only (T-7) |
| Another person's page on Sources | `PersonSourcesSection` | The host's key carries the entity id; `expanded` resets |

### Scale

Brief § 2: typical 3 – 20 records from 2 – 12 Sources, large 50 from 14 (9 under one). The warm read issues no SQL (S13-01). `PersonSourcesContent` is rebuilt when the claims, either list, the name orders or `expanded` change: O(claims × observations) for the names (one sort per Subject, as the graph's snapshot does). Each row is a few text views; 50 lines is far under the Timeline's 60-row budget. No new write, key or invalidation; two extra warmed keys, both catalog-wide lists warm on any graph visit.

### Edges

- **Assumptions.** The read's order is kept (rows by title, lines by graph name); the content does not re-sort (T-1). The fold counts lines, not Sources (T-3).
- **Neighbors.** 6 lines → no fold; 7 → five and "Show 2 more" (T-3). A graph opened with no `subjectId` → today's camera (T-7, T-8).
- **Removed inputs.** Source title empty → ref in mono (T-1); type missing → empty cell (T-2); graph ref or name empty (T-2); Subject type missing or snapshot nil → kind `.person`, the name still read from Observations (T-2); opening Subject missing from the graph → no selection, no focus (T-7).
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
| Title link | `Text` + `PVCitationLinkButtonStyle`, `.pvFocusRing`, `pvActivatesOnReturn` | same file | D-7 |
| Type | `CatalogSourceTypePill` | — | D-7 |
| Graph line | `PVButton(.link, .sm)` + caption text | same file | D-7 |
| Fold | `ConclusionShowMoreRow` | — | D-10 |
| States | the host's spinner, `PVCallout`, `PVEmptyState` (glyph `library`) | host | SS-7, SS-8, SS-W3 |

### Strings

L10n keys *(new)*, in `L10n.Conclusions`: `sourcesErrorTitle` = "Sources could not be read"; `sourcesEmptyTitle` = "No records"; `sourcesEmptyMessage` = "This person has no promoted records."; `sourcesColumnSource` = "Source"; `sourcesColumnType` = "Type"; `sourcesColumnDate` = "Date"; `sourcesColumnRef` = "Ref"; `sourcesDateNone` = "—"; `sourcesRowAccessibility(title:type:date:ref:)` = "%1$@, %2$@, %3$@, %4$@" and `sourcesRowAccessibilityUntyped(title:date:ref:)` = "%1$@, %2$@, %3$@"; `sourcesDateNotRecorded` = "date not recorded"; `sourcesOpenSource(_:)` = "Open %@"; `sourcesOpenGraph(graph:subject:ref:)` = "Open %1$@, record %2$@, %3$@". Removed: `conclusions.pageSection.later` (`pageSectionLater`).

## 7. Steps

**STEP-1 — A section declares its keys**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `ConclusionPageSectionHost.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-workspace-place`.
Do: D-1.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PersonDetailTests` passes (T-6); Family and Timeline load as before.

**STEP-2 — One builder for a placed Subject**
Files: `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift`, `macos/ProvenenciaTests/PersonSourcesContentTests.swift` *(new)*.
Do: D-2.
Check: T-4 passes; `EvidenceGraphModelTests` pass unchanged.

**STEP-3 — One source type label lookup**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Features/Sources/SourcesModel.swift`, `macos/App/Features/Sources/SourceIdentitySection.swift`.
Do: D-6.
Check: the Sources tests pass unchanged.

**STEP-4 — The section's content and locations**
Files: `macos/App/Features/Conclusions/PersonSourcesContent.swift` *(new)*, `macos/App/Features/Sources/SourcesListNavigation.swift`, `macos/ProvenenciaTests/PersonSourcesContentTests.swift`.
Do: D-3, D-5, D-10's fold in the content.
Check: T-1 – T-3 pass.

**STEP-5 — Share the citation link style**
Files: `macos/App/DesignSystem/Recipes/ProvenanceSheet/PVProvenanceRecordRow.swift`. Skill: `add-ui-component`.
Do: rename `CitationLineButtonStyle` to `PVCitationLinkButtonStyle`, internal; the record row uses it unchanged.
Check: the provenance sheet tests pass unchanged; `scripts/check-guidance-names.py` passes.

**STEP-6 — The Sources section**
Files: `macos/App/Features/Conclusions/PersonSourcesSection.swift` *(new)*, `PersonDetailView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-4, D-7 – D-10, D-12, with their keys; replace the `.sources` case.
Check: manual check items 1 – 4; the build has no warning.

**STEP-7 — A graph opens on the Subject its location names**
Files: `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `EvidenceGraphModel.swift`, `macos/App/Features/GraphCanvas/GraphCanvasScrollView.swift`, `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`, `macos/ProvenenciaTests/GraphCanvasScrollViewTests.swift` *(new)*.
Do: D-11, deleting the pending-selection hook.
Check: T-7, T-8 pass; `rg 'PendingGraphSelection|pendingGraphSelections' macos` finds nothing.

**STEP-8 — Strings check**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: delete `pageSectionLater` once `rg pageSectionLater macos/App` finds only its definition; every new key has a catalog entry and comment.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `.agents/skills/add-workspace-place/SKILL.md` if it names `queryKey`.
Do: per the Definition of done. The brief stays open for S13-17.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-16-sources-section.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PersonSourcesContentTests.swift` `rowsKeepTheReadsOrder` *(new)* | The AC-1.1 groups → three rows in order, the census row's lines Household 214 then 388, the letter's title `.ref("SRC-4KQ7M")`; `[]` → `isEmpty` | AC-1.1, AC-1.6, D-3 |
| T-2 | Swift | same · `cellsAndFallbacks` *(new)* | Type id in the list → `typeLabel` "Census" and icon key; unknown → nil; `graphRef` "" → nil and the graph name as title; graph name "" → the ref as title and `graphRef` nil; Subject type missing or `properties` nil → `subjectName` from its naming Observation; `location` = `.sourcePage(...)`; line `location` = `canvasLocation(sourceId:sourceRef:sourceTitle:graphId:subjectId:)` with the Subject id; row accessibility label "Hamburg census 1867, Census, date not recorded, SRC-9TQ2M" | AC-1.2, AC-1.3, AC-1.4, AC-2.1, AC-2.2, AC-2.3, D-5, D-6, D-7, D-12 |
| T-3 | Swift | same · `foldPastSixLines` *(new)* | 9 lines → 5 shown, `hiddenCount` 4; 6 → 6, 0; 7 → 5, 2; Source id in `expanded` → all 9 | AC-1.5, D-10 |
| T-4 | Swift | same · `lineNamesTheRecordAsItsCard` *(new)* | A claim with a `name` Observation "J. Weber" and label "jw" → `subjectName` "J. Weber", equal to `EvidenceGraphSnapshot.build`'s `displayName` for the same Subject | AC-1.4, D-2 |
| T-5 | Swift | `macos/ProvenenciaTests/ProvenanceSheetTests.swift`, `PVProvenanceColumnsTests.swift` (existing) | Unchanged with the renamed style | D-7 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys` (changed) | `.family` → `[.personFamily]`, `.timeline` → `[.personEvents]`, `.sources` → `[.personClaims, .propertiesWorkspace, .sourceTypesList]`, others `[]`; `fillsWidth` true only for Timeline and Gallery | AC-3.1, AC-3.3, D-1, D-9 |
| T-7 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `openingSubjectIsSelected` *(new)* | `openingSubjectID: "s2"` on a graph with s1, s2 → `selectedSubjectID == "s2"`, `openingFocus` = s2's card center; `"gone"` → no selection, nil; nil → today | AC-2.2, AC-2.4, D-11 |
| T-8 | Swift | `macos/ProvenenciaTests/GraphCanvasScrollViewTests.swift` *(new)* `openingFocusCentersThePoint` | `applyOpeningCamera(nil, contentSize: 4000×4000, focus: (3000, 1200))` in a 1000×800 view → visible center (3000, 1200) at magnification 1; near an edge → clamped; nil → today | AC-2.2, AC-2.4, D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: the Family and Timeline tests (the key rename), `EvidenceGraphModelTests`, the provenance row tests.

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
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected; `add-workspace-place` if it names `queryKey`
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Dogfood project, a Person with records from at least three Sources, one Source with two Evidence graphs. Sources: the header line, one row per Source in title order with its type pill, "—" and ref; under the two-graph Source, two lines (AC-1.1, AC-1.2, AC-1.4). Compare with board frame 14a.
2. Click a Source title: its Source page; Back: Sources. Click a graph line: the graph centered on that Subject, its card selected; Back: Sources. Open a graph from the Sources list: it opens where it was (AC-2.1, AC-2.2, AC-2.4).
3. A Source with seven or more records shows five and "Show N more"; click it (AC-1.5). Promote a new Subject into the person on a graph and return: the line is there (AC-3.2). In a preview with FakeStore's scripted claims error: "Sources could not be read" (AC-1.7).
4. Stretch across 1600pt and back: centered at 1180pt, scroll and focus kept, no How the records connect region (AC-3.3, AC-3.4). Tab order and VoiceOver as frame 14v. Dark mode at both widths.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Renaming `queryKey` misses a caller | The compiler; T-6 |
| The renamed link style changes the provenance sheet | T-5; the style's body is unchanged |
| A line overflows on long names | D-7's cut order; manual item 1 with the Verzeichniß title and "Weber, Joh., Tischler" |

**Stop and report if:**

- `applyOpeningCamera` runs before the canvas's content size is known, so a focus cannot be centered (D-11).
- `PVButton(.link, .sm)` does not render at caption size inside the line (D-7).
- The section host cannot take a `make` that reads two other handles without a redraw loop (D-8).
- The board's frames and this plan's measures disagree (14a's columns, 14e's fold).

## 12. Open questions

None.

## 13. PR

**Title:** S13-16 — Sources section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-16-sources-section.md`.
