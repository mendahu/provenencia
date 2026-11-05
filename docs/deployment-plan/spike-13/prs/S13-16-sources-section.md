# S13-16 — Sources section

**Status:** Blocked — on board S13-D6
**Spike / slice:** Spike 13 · Slice 6 — Sources
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D6](../design/S13-D6-sources-section.md), board not drawn yet
**Board:** — (not drawn yet; on agreement this line names the Claude Design turn and frames, and **Design → macOS** describes every frame in words)
**Depends on:** S13-01 (landed [#655](https://github.com/mendahu/provenencia/pull/655)), S13-10 (landed [#634](https://github.com/mendahu/provenencia/pull/634)), S13-13 (landed [#691](https://github.com/mendahu/provenencia/pull/691): `ConclusionPageSectionHost`), S13-D6 (board)
**Blocks:** S13-17, S13-28
**Branch:** `s13-16-sources-section`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-14 and S13-15 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (the section key the place warms), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`../person-detail.md` § 5 Sources](../person-detail.md#5-sources) (R-33 – R-37, R-2, R-13b); working rules "Each section reads its own query", "Claims are read per person from the graph cache, grouped by Source in Go", "Sections are places", "A full-width section keeps its states in the band", "Canonical only, filtered in Go"; brief [S13-D6](../design/S13-D6-sources-section.md); the S13-01 plan (D-4, D-14: the type id and the Swift types), the S13-13 plan (D-10: the section host)
**Planned:** 2026-11-05 against `main` @ `124d492` (draft: refreshed on board agreement)
**Review:** [`S13-16-sources-section.review.md`](S13-16-sources-section.review.md) — round 1: Fits with changes / Not ready; every finding Fixed on the draft; round 2 runs when the board is agreed

**Re-plan pending (2026-11-05).** The user rejected the board's first draft and the brief was rewritten (revision 2): Source rows like the provenance sheet's, with secondary Evidence graph links, a reserved date column and a reserved place for how the records connect; no Subject cards. Every Decision about the card (D-2 – D-5, D-7's `Card`, D-10, D-11, T-3 – T-5, STEP-2, STEP-5, STEP-6) is void and is rewritten when the new board is agreed. D-1, D-6, D-8, D-9 (host), D-13 (a graph opens on the Subject) and their tests carry over.

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

A researcher opens a Person's Sources section and sees every record that makes up the person: one card per accepted identity claim, drawn as the Evidence graph draws that Subject, grouped under its Source (title and type) in Source title order. The Source's title opens its Source page; each card names its Evidence graph and opens it with the Subject selected and scrolled into view (user, 2026-11-05). Today the section is an EmptyState ("Sources arrives in a later version of Provenencia.").

**Why now.** Slice 6's chrome. The read (`ListPersonClaims`, S13-01) and its cache key (`.personClaims`, with its invalidation) have been in the app, unread, since #655; the section host (S13-13) makes the section a content view only. S13-17 puts Remove on these cards.

**What stays.** The engine: no Go, proto or store change. The Evidence graph canvas and its card look and behave as today, except that a graph opened from a location naming a Subject now selects it and opens on it (STEP-7). The header, the switcher and the other four sections. Remove is S13-17's: this PR leaves the footer's end square empty.

## 2. User stories

**US-1.** As a researcher, when I open a Person's Sources section, I see each of the person's records under the Source it came from, so that I can tell where everything I know about them was recorded.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann Weber with claims from "Hamburg census 1867" (graphs Household 214 and Household 388), "Hamburger Adressbuch 1861" and an untitled Source SRC-4KQ7M | Three groups, in the read's order: "Hamburg census 1867" with two cards (Household 214, then Household 388), "Hamburger Adressbuch 1861" with one, then SRC-4KQ7M (in mono) with one (R-33, R-34, R-36; SS-1) |
| AC-1.2 | The census group, its Source type id naming the seeded type "Census" | Its heading reads the title, the Source type pill "Census" with the type's icon, and "2 records" (R-35; SS-2) |
| AC-1.3 | A group whose type id is not in the source types list (the list not loaded yet, or the type deleted meanwhile) | The heading shows no pill; the title and count stay (D-8) |
| AC-1.4 | A claim on a Subject with five Observations (name, sex, age, occupation, birthplace) | Its card paints the Evidence graph card's header and the five property rows in the graph's order (property label A→Z), with no pen or trash icon and no Add property row (SS-3; D-2, D-4) |
| AC-1.5 | A person with no accepted claim in the read (a delete raced the read) | The EmptyState "No records" · "This person has no promoted records.", id `persons.detail.sources.empty` (SS-6) |
| AC-1.6 | The read fails | The danger Callout "Sources could not be read" with the error, id `persons.detail.sources.error` (SS-7) |

**US-2.** As a researcher, from a Source group or a claim card, I can open the Source page or the Evidence graph, so that I can go to the evidence itself.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Click (or Return on) the "Hamburg census 1867" title | `navigation.go(to: .sourcePage(sourceId:ref:title:))` with that Source's id, ref and title; Back returns to the Sources section (R-37; SS-2) |
| AC-2.2 | Click the Household 388 card's graph control | `navigation.go(to:)` the location `SourcesListNavigation.canvasLocation(sourceId:sourceRef:sourceTitle:graphId:subjectId:)` builds: the Source's id, ref and title (as every graph location carries them), the claim's graph id and its Subject's id; the graph opens with that card selected and centered; Back returns to Sources (SS-4; D-6, D-13) |
| AC-2.3 | A claim whose graph has no ref (`graphRef` empty) | The footer shows the graph name only; a graph with no name shows its ref (D-7) |
| AC-2.4 | The Source list's Evidence graph row, the Promote page and the provenance sheet open graphs with no Subject | The graph opens as today: its saved camera, no selection change (D-13) |

**US-3.** As a researcher, the section loads and refreshes on its own, at both widths.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Open the Person page on Sources | The place warms `.personClaims`, `.propertiesWorkspace` and `.sourceTypesList` (the section's `queryKeys`), and the section shows its spinner until the claims' first value (SS-7; D-1) |
| AC-3.2 | Promote another Subject into the person on an Evidence graph, then return | The section shows the new claim; the reload keeps the cards drawn with "Refreshing" in the switcher row (S13-01's `refreshes`, unchanged) |
| AC-3.3 | The pane crosses 1600pt | The section's Wide arrangement is the board's SS-W1 answer; the scroll position and focus survive (SS-W1, SS-W2) |

## 3. Scope

| In | Covered by |
| --- | --- |
| A section declares every key it reads; `.sources` reads `.personClaims` and its two lists (`ConclusionPageSection.queryKeys`) | STEP-1, AC-3.1 |
| One per-Subject builder for a placed Subject, shared with `EvidenceGraphSnapshot.build` | STEP-2, AC-1.4 |
| One label lookup for a source type id | STEP-3, AC-1.2, AC-1.3 |
| The display content: groups, headings, cards, locations (SS-1, SS-2, SS-4; R-33 – R-37) | STEP-4, AC-1.1 – AC-1.3, AC-2.1 – AC-2.3 |
| The Evidence graph card in a page chrome, and a SwiftUI host (SS-3) | STEP-5, STEP-6, AC-1.4 |
| The Sources section on the section host: heading row, cards, footer graph control, states, keyboard and VoiceOver, dark mode (SS-3 – SS-10) | STEP-7, AC-1.5, AC-1.6, AC-2.2 |
| The Wide arrangement (SS-W1 – SS-W3) | STEP-7, AC-3.3 |
| A graph location's Subject: selected and centered on open; the unused pending-selection hook goes (user, 2026-11-05) | STEP-8, AC-2.2, AC-2.4 |
| Strings | STEP-9 |
| Docs | STEP-10 |

| Out | Home |
| --- | --- |
| Remove on a card, its confirm, landing on the Persons list (RC-1 – RC-7) | S13-17 |
| The Source date and sorting by it | [Spike 15's Source date](../../spike-15/source-provenance-date.md) |
| How the records connect (links between cards) | S13-28 (board S13-D10) |
| A provenance control on a card | not planned this spike (R-13b) |
| Edit, delete, Add property or Promote from the page's card | not planned (they stay on the Evidence graph; brief § 2.1) |
| A card that reads its property rows to VoiceOver one by one | not planned (the graph card does not either; brief SS-9) |
| A fold, filter, search or sort | not planned (brief SS-8, § 2.1), unless the board's SS-8 answer folds |
| Any Go, proto or store change | not planned: S13-01's read carries everything |
| Event and Place pages with a Sources section | their redesign (R-1) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| The Sources place | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` | `case .sources:` draws `PVEmptyState` with `L10n.Conclusions.pageSectionLater`, its last caller. Family and Timeline call `PersonFamilySection(entityId:)` and `PersonTimelineSection(entityId:)` |
| Section keys | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection.queryKey(project:entityId:)` | One optional key: `.family → .personFamily`, `.timeline → .personEvents`, others nil. `fillsWidth`: `.timeline, .gallery` |
| Its readers | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` (`.personDetail` spec, `location.pageSection.flatMap { $0.queryKey(...) }`); `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `sectionRefreshing`; `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` `key` | The place warms the one key, the page's refresh follows it, the host reads it. Every other place lists its supporting keys in `queryKeys` (`.sourceGraph` lists `.propertiesWorkspace`, `.sourceTypesList`) |
| Section host | `ConclusionPageSectionHost` | Gets the section's handle, starts the read, draws the spinner, error Callout (`<root>.error`) and EmptyState (`<root>.empty`) around `body(content)`; `make: (Value) -> Content` runs in the body, so a closure that reads other handles redraws when they change |
| Claims read | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonClaimGroup{sourceID, sourceRef, title, sourceTypeID, claims}`, `CatalogPersonClaim{claimID, graphID, graphRef, graphName, subject, observations}` | Groups in Source title order, claims in graph name order (`core/database/persondetail/claims.go`); observations in wire order |
| Claims key | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `case .personClaims` | `.sessionFresh`, evicted when hidden; refreshes on the person's touched handles, the graphs it holds, property labels |
| Placed Subject | `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift` `EvidenceGraphSnapshot.build` | Its loop body, per Subject: the kind from `EvidencePrimaryKind(rawValue: type.key)`, `typeLabel: type.label`, Observations sorted by the `private static` `observationDisplayOrder`, `isCited`, membership, event title |
| The card | `macos/App/Features/EvidenceGraph/EvidenceCardView.swift` `EvidenceCardView`: `isGhost` (private), `init(ghostOf:)`, `private static func measure`, `subject(...)`, `paintHeader` (always paints `PVSymbol.penLine` and `PVSymbol.trash`), `recordFooter`, `paintFooter` | One measure for layout and paint; a cited card adds an Add property row; the footer is Promote or the handle row with the Unpromote square, split by `EvidenceSubjectCard.promotedFooterParts(footer:)` into `handle`, `hairline`, `slot` |
| Card hits | `macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift` `actionTargets(for:canCite:…)`, `EvidenceCardHeaderActionHits.frames(…, showDelete:)` | The canvas's hit targets; the pen and trash hits come from `EvidenceCardHeaderActionHits`, not from layout `regions` |
| The card's model | `macos/App/Features/EvidenceGraph/EvidenceCanvasCard.swift` `EvidenceCanvasCard{placed, text, state, nameOrders, projectDir, cardinalityByKey}`, `static func ghost(_:presentation:)` | `EvidenceCardText(subject:orders:)` carries `var accessibilityLabel` ("Person, CPR-…, name, cited") |
| Canvas item view | `macos/App/Features/GraphCanvas/GraphCanvasItemView.swift` | Frame = content + `paintOutset` (24pt) each side; `draw(_:)` paints in any hierarchy |
| Subject types | `macos/App/Features/Properties/PropertiesSnapshot.swift` `PropertiesSnapshot{types, presentationsByKey, …}`, `cardinalityByKey` | Key `.propertiesWorkspace` |
| Source types | `.sourceTypesList`; `macos/App/Platform/GenealogyStore.swift` `[CatalogSourceType].iconKey(forTypeID:)`; label lookups copied in `macos/App/Features/Sources/SourcesModel.swift` `typeLabel(for:)` and `macos/App/Features/Sources/SourceIdentitySection.swift` `typeLabel` | Two copies of `types.first { $0.id == … }?.label ?? ""` |
| Source type look | `macos/App/Features/Catalog/SourceTypeMarkers.swift` `CatalogSourceTypePill(label:iconKey:)` | The Source page header's pill |
| Title link | `macos/App/Features/Conclusions/PersonTimelineSection.swift` (event title) | `Button { open } label: { PVTitleText(title, style: .heading, lineLimit: 2) }` with `PVFocusRingButtonStyle`, `.focused`, `pvActivatesOnReturn(when:)` |
| Graph locations | `macos/App/Features/Sources/SourcesListNavigation.swift` `canvasLocation(for:graph:)` | Puts the Source's ref and title in `ref` / `title` and `sourceTitle`; no `subjectId`. `.sourcePage(sourceId:ref:title:)` in `WorkspaceLocation.swift` |
| Opening a graph | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `.sourceGraph` → `EvidenceGraphView(graphID:sourceID:session:store:userID:)`; `EvidenceGraphView` passes `restoredCamera: model.session.graphCamera(for: graphID)` to the canvas; `macos/App/Features/GraphCanvas/GraphCanvasScrollView.swift` `applyOpeningCamera(_:contentSize:)` on mount and on a `contentID` change | A graph location's `subjectId` is not read by the canvas. The canvas mounts only once the graph's rows are loaded (`EvidenceGraphView` shows `ProgressView` before) |
| Pending selection | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `pendingGraphSelections`, `setPendingGraphSelection(subjectID:graphID:)`, `consumePendingGraphSelection(graphID:)`; consumed in `EvidenceGraphModel` after its `readyValue(graphKey)` wait | `rg setPendingGraphSelection macos` finds no caller: half dead |
| Kit | `PVFlowLayout`, `PVButton`, `PVEmptyState`, `PVCallout`, `PVFocusRingButtonStyle`, `pvActivatesOnReturn` | All exist |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys`; `EvidenceCardLayoutTests.swift` (incl. `ghostCardMatchesTheLiveLayout`, `ghostPaintsAtGhostOpacity`); `EvidenceSubjectCardFooterTests.swift`; `EvidenceGraphModelTests.swift` | `sectionQueryKeys` expects `.sources` → nil |

**Precedent.** S13-13 ([#691](https://github.com/mendahu/provenencia/pull/691)) added the Timeline on `ConclusionPageSectionHost` with one `queryKey` case and a `…Content` built in `make` from captured inputs (`orders`, `widthClass`); copy that shape for `PersonSourcesSection` and `PersonSourcesContent`, and copy its title link. The ghost card (`EvidenceCanvasCard.ghost`, `EvidenceCardView(ghostOf:)`) is the precedent for a second presentation of the one paint.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ConclusionPageSection.queryKey(project:entityId:)` becomes `queryKeys(project:entityId:) -> [CatalogQueryKey]` *(rename)*, the section's own read first: `.family → [.personFamily]`, `.timeline → [.personEvents]`, `.sources → [.personClaims, .propertiesWorkspace, .sourceTypesList]`, others `[]`. `PlaceRegistry` appends them all; the host and `sectionRefreshing` read `.first` (the host's precondition becomes "non-empty"). The section reads its two lists with `session.queryHandle` | The place owns warming (every other place lists its supporting keys); one declaration; S13-28 gets a slot (round 1 M-6) | A `session.query` in the section's `.task` (a second warm path) |
| D-2 | One presentation enum for the card: `EvidenceCanvasCard.Chrome { case canvas, ghost, page }` *(new)*, field `chrome` default `.canvas`. `EvidenceCardView`'s private `isGhost` goes: `.ghost` is the ghost (opacity 0.62, no accessibility element), and `EvidenceCanvasCard.ghost(_:presentation:)` sets it. `EvidenceCanvasCard.page(_:presentation:text:orders:projectDir:cardinalityByKey:)` *(new)* builds a resting `.page` card. `.page` paints the header without the pen and trash icons (`paintHeader(…, icons: Bool)` *(new parameter)*, true for `.canvas` and `.ghost`), skips the Add property row (cited and uncited), and lays out and paints the footer band and hairline of a promoted card with no handle text and no Unpromote icon, recording `EvidenceSubjectCard.pageGraphActionID` *(new)* = the `handle` part and `EvidenceSubjectCard.pageSlotActionID` *(new)* = the `slot` part of `promotedFooterParts`. `EvidenceCardHeaderActionHits` and `actionTargets` are not called for a page card (no canvas hosts it) | One way to vary the paint (round 1 M-7); the footer keeps the graph's geometry so S13-17 fills the slot without re-laying it out (round 1 M-8); R-35 one card | A `Chrome { canvas, page }` beside `isGhost`; one region over the whole footer |
| D-3 | `EvidenceCardView.layout(for card: EvidenceCanvasCard) -> EvidenceCardLayout` *(new, static)* returns `measure(card)` without painting. The host is `EvidenceCardViewHost` *(new)*, an `NSViewRepresentable` in `macos/App/Features/EvidenceGraph/EvidenceCardViewHost.swift` *(new)*: `init(card:)`; `makeNSView` builds `EvidenceCardView(card:actions: [])`; `updateNSView` replaces the view's content only when `card` differs (it rebuilds through a container `NSView` holding one `EvidenceCardView` subview at origin `-paintOutset`); its SwiftUI frame is `EvidenceCardView.layout(for:).size`. Callers place SwiftUI controls over `layout.regions[pageGraphActionID]` and `[pageSlotActionID]` with `.overlay(alignment: .topLeading)` and `.offset` | The paint is AppKit and already paints outside the canvas; the layout is one static call, so SwiftUI knows the footer before the view exists (round 1 X-4); named for what it is (round 1 M-11) | Rendering an image per card (a likeness loads late) |
| D-4 | One per-Subject builder: `EvidenceGraphPlacedSubject.init(subject:kind:typeLabel:observations:gridX:gridY:membership:eventTitle:)` *(new)* sorts the Observations with `EvidenceGraphSnapshot.observationDisplayOrder` (stays private to the file: the initializer lives in the same file) and sets `isCited`. `build`'s loop resolves the type, calls it for a primary kind, and keeps its bridge branch | One place for a Subject's card rules; Event and Place callers become a call (round 1 M-2) | A second builder sharing only the sort |
| D-5 | The Sources content resolves each claim's type from `PropertiesSnapshot.types` by `subject.subjectTypeID`: kind `EvidencePrimaryKind(rawValue: type.key)`, `typeLabel` `type.label`, presentation `presentationsByKey[type.key]`. When the snapshot has no value yet, or the type is missing or not a primary kind, the card is drawn as kind `.person` with an empty type label and no presentation (Promote files only person Subjects into a person), and redraws when the snapshot arrives. `cardinalityByKey` is `PropertiesSnapshot.cardinalityByKey`, empty when nil | A claim is never hidden for want of a label (R-33); the fallback is the one kind a Person's claim has, not a rule for every claim | Hiding the section until the types load |
| D-6 | Graph locations go through one builder: `SourcesListNavigation.canvasLocation(sourceId:sourceRef:sourceTitle:graphId:subjectId:)` *(new overload)*; `canvasLocation(for:graph:)` calls it with `subjectId: nil`. The Source title location is `.sourcePage(sourceId:ref:title:)` with the group's id, ref and title | One place builds a graph location; history and breadcrumbs keep the Source's ref and title (round 1 M-5) | Putting the graph's ref and name in the location |
| D-7 | `PersonSourcesContent` *(new)* in `macos/App/Features/Conclusions/PersonSourcesContent.swift` *(new)*: `init(groups: [CatalogPersonClaimGroup], sourceTypes: [CatalogSourceType], properties: PropertiesSnapshot?, orders: NameFormatOrders, projectDir: String)`, `Equatable`; `groups: [Group]`, `Group{id (sourceID), title: PVTitleSource (.label(title), else .ref(sourceRef)), typeLabel: String?, typeIconKey: String?, countText: String, accessibilityLabel: String, location: WorkspaceLocation, cards: [Card]}`; `Card{id (claimID), subjectID, card: EvidenceCanvasCard (D-2 `.page`), layout: EvidenceCardLayout, graphTitle: String (graphName, else graphRef), graphRef: String? (nil when empty or equal to `graphTitle`), graphLocation: WorkspaceLocation}`; `isEmpty` when no group. The card's `text` is `EvidenceCardText(subject:orders:)` with `accessibilityLabel` replaced by `L10n.Conclusions.sourcesCardOnGraph(label:graph:)` *(new)* | The Timeline and Family pattern; one VoiceOver element per card, the graph card's own label plus its graph (round 1 M-9) | A second card label on a SwiftUI group around the view |
| D-8 | `[CatalogSourceType].label(forTypeID:) -> String?` *(new)* beside `iconKey(forTypeID:)` in `GenealogyStore.swift`, nil for a missing id; `SourcesModel.typeLabel(for:)` and `SourceIdentitySection.typeLabel` move onto it (`?? ""` at those two callers). Nil → no pill | One lookup (round 1 M-10); S13-01 D-4 | A third copy |
| D-9 | `PersonSourcesSection(entityId:)` *(new)* in `macos/App/Features/Conclusions/PersonSourcesSection.swift` *(new)* on `ConclusionPageSectionHost<[CatalogPersonClaimGroup], PersonSourcesContent, PersonSourcesList>` with `errorTitle: L10n.Conclusions.sourcesErrorTitle`, `emptyTitle: L10n.Conclusions.sourcesEmptyTitle`, `emptyMessage: L10n.string(L10n.Conclusions.sourcesEmptyMessage)`; `make` reads the `.propertiesWorkspace` and `.sourceTypesList` handles' values, `session.nameFormat` and `session.projectKey.projectDir`. `PersonSourcesList` *(new, same file)* takes `content` and `open: (WorkspaceLocation) -> Void`. Ids: `persons.detail.sources.group.<sourceID>`, `.title.<sourceID>`, `.card.<claimID>`, `.graph.<claimID>` | The host owns the root, `.empty` and `.error` ids | A view that reads the claims handle itself |
| D-10 | The group heading row: the Timeline's title link (`Button` + `PVTitleText(group.title, style: .heading, lineLimit: 2)`, `PVFocusRingButtonStyle`, `.focused`, `pvActivatesOnReturn(when:)`, `.help` with the full title, `.accessibilityAddTraits(.isHeader)`, label `group.accessibilityLabel` = `L10n.Conclusions.sourcesGroupAccessibility(title:type:count:)` *(new)*), then `CatalogSourceTypePill(label:iconKey:)` when `typeLabel` is non-nil, then `countText` in caption muted, on one `firstTextBaseline` row that wraps (`PVFlowLayout`). **Spacing between groups and any rule between them: from the board's SS-1** | The sibling section's title link and the Source page's type pill (round 1 M-3, M-4); `PVSectionHeader` takes copy, not a view (round 1 X-1); a Chip is a button (round 1 X-3) | `PVSectionHeader` with a button in `aside`; `PVChip` as a label |
| D-11 | Cards wrap with `PVFlowLayout` at their 264pt width, top-aligned. Over `pageGraphActionID` sits a `PVButton` (ghost, `.sm`) titled `graphTitle` with `graphRef` in mono after it, label `L10n.Conclusions.sourcesOpenGraph(_:)` *(new)*, opening `graphLocation`. The `pageSlotActionID` square stays empty (S13-17). **The gap, and whether the button fills the whole handle part, follow the board's SS-3 and SS-4** | The paint owns the card, SwiftUI owns the controls (hover, focus ring, Return are the kit's: SS-5) | Hit testing inside the AppKit view |
| D-12 | Wide (SS-W1): **the board's answer, recorded on agreement.** (A) sets `ConclusionPageSection.sources.fillsWidth` true (the host's states already keep the band); (B) leaves it false. The brief offers no secondary column | The user left it to the board (2026-11-05) | — |
| D-13 | A graph location's `subjectId` selects and centers that Subject on open, replacing the pending-selection hook. `WorkspaceDestinationHost` passes `navigation.currentLocation.subjectId` to `EvidenceGraphView(…, openingSubjectID:)` *(new parameter)*, which passes it to `EvidenceGraphModel(…, openingSubjectID:)` *(new)*. The model selects it where it consumes the pending selection today, after its `readyValue(graphKey)` wait, when the snapshot holds the id. The canvas gets `openingFocus: CGPoint?` *(new, on `GraphCanvasScrollView`)* = the Subject card's `EvidenceSubjectCard.edgeFrame(for:)` center when the model has an opening Subject the snapshot holds; `applyOpeningCamera(_:contentSize:focus:)` *(new parameter)* keeps the restored magnification (else 1) and sets the content offset that centers `focus` in the visible rect, clamped as today. `pendingGraphSelections`, `setPendingGraphSelection` and `consumePendingGraphSelection` are deleted | One way to open a graph on a Subject, which history and restore pick up; the canvas already applies its camera at mount, after the rows load (round 1 M-1, M-12; user, 2026-11-05: scroll into view) | A second `selectOnOpen` path beside the pending hook; a camera set before navigating (the Sources page has no positions) |
| D-14 | No fold (brief SS-8), **unless the board folds**; then `ConclusionShowMoreRow` at the count the board names | One card is one record | A new fold control |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionPageSection.queryKeys` *(rename of `queryKey`)* | func | `queryKey` | Extended to a list; its three readers move |
| `EvidenceCanvasCard.Chrome`, `.page(...)` *(new)* | type, func | `isGhost`, `EvidenceCanvasCard.ghost` | Folds the ghost in; one enum |
| `EvidenceSubjectCard.pageGraphActionID`, `pageSlotActionID` *(new)* | constants | `openHandleActionID`, `unpromoteActionID` | The page's controls are not the canvas's actions |
| `EvidenceCardView.layout(for:)` *(new)* | func | the private `measure` | Exposes it, no new measure |
| `EvidenceCardViewHost` *(new)* | view | `GraphCanvasItemLayer` | The canvas hosts a whole graph; a page needs one card at rest |
| `EvidenceGraphPlacedSubject.init(subject:kind:…)` *(new)* | init | `EvidenceGraphSnapshot.build`'s loop body | Extracted from it |
| `[CatalogSourceType].label(forTypeID:)` *(new)* | func | `iconKey(forTypeID:)`, `SourcesModel.typeLabel`, `SourceIdentitySection.typeLabel` | Replaces two copies |
| `SourcesListNavigation.canvasLocation(sourceId:…subjectId:)` *(new)* | func | `canvasLocation(for:graph:)` | Overload; the old one calls it |
| `GraphCanvasScrollView.openingFocus`, `applyOpeningCamera(…focus:)` *(new)* | parameter | `restoredCamera` | The opening camera's one path |
| `EvidenceGraphModel(…openingSubjectID:)` *(new)* | parameter | `pendingGraphSelections` | Replaces it |
| `PersonSourcesContent`, `PersonSourcesSection`, `PersonSourcesList` *(new)* | type, views | `PersonTimelineContent`, `PersonTimelineSection` | Same pattern |
| `L10n.Conclusions.sources*` *(new)* | keys | `timelineErrorTitle`, `timelineEmptyTitle` | One set per section |
| `PersonSourcesContentTests` *(new)* | test file | `PersonTimelineContentTests` | One file per content type |

**Builds on.** `ConclusionPageSectionHost`; `.personClaims` and its `refreshes` (S13-01); the Evidence graph paint, `EvidenceCardText`, `promotedFooterParts`; `PropertiesSnapshot`; `CatalogSourceTypePill`; the Timeline's title link; `SourcesListNavigation`; the canvas's opening camera.
**Deletes.** The `.sources` EmptyState; `L10n.Conclusions.pageSectionLater` and its `Localizable.xcstrings` entry (its last caller); `EvidenceCardView.isGhost`; `WorkspaceSession.pendingGraphSelections`, `setPendingGraphSelection(subjectID:graphID:)`, `consumePendingGraphSelection(graphID:)` and the model's consume; the two `typeLabel` copies' bodies; the name `queryKey` (renamed).
**Next adds.** S13-17 puts Remove in `pageSlotActionID`; S13-28 appends its keys to `.sources`' `queryKeys`; an Event or Place page with sections uses `EvidenceCanvasCard.page` with the kind its type gives; any surface that opens a graph on a Subject passes `subjectId`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Sources title → Source page | `macos/App/Features/Sources/SourcePageView.swift` | An existing location; a Source deleted meanwhile falls back to the Sources list (`fallbackToSectionRoot`, today) |
| Card → graph with `subjectId` | `EvidenceGraphView`, `EvidenceGraphModel`, `GraphCanvasScrollView` | New: the canvas reads `subjectId` (D-13). A Subject no longer on the graph: no selection, today's camera. Back and restore to that entry select and center it again (intended). Opening the same graph from the canvas's own history with another `subjectId`: `PlaceRegistry` `.sourceGraph` keeps the model by `graphId`, so the opening Subject applies only on a fresh mount; T-8 pins it |
| A Person page opened on Sources for another person | `PersonSourcesSection` | The host's key carries the entity id |
| `EvidenceCardView` in a SwiftUI page | `GraphCanvasItemView.draw(_:)` | Paints in any hierarchy; likeness loads run in `viewDidMoveToWindow` with `projectDir` set |

### Scale

Brief § 2: typical 3 – 20 claims, large 50, up to 30 rows a card. The warm read issues no SQL (S13-01). Per card: one static `layout(for:)` in `PersonSourcesContent` and one `EvidenceCardView` measure when the host builds the view; the host replaces the view only when the card differs, so a reload with unchanged cards measures nothing new. 50 cards × 30 rows is the graph's own measure for a 50-card graph. `PersonSourcesContent` is rebuilt when the claims, either list, the name orders or the project change: O(claims × observations log observations). No new write, key or invalidation; two extra warmed keys, both catalog-wide lists already warm on any graph visit.

### Edges

- **Assumptions.** Same-key rows stay adjacent (the conflict bracket): the one initializer sorts for both callers (T-3). A page card is never handed to `actionTargets` (D-2); T-4 checks paint and regions, not header hit ids.
- **Neighbors.** `.canvas` and `.ghost` layouts and paint unchanged (T-5: `EvidenceCardLayoutTests` unchanged, `ghostCardMatchesTheLiveLayout`, `ghostPaintsAtGhostOpacity`). An uncited claim in `.page`: header, then the footer, no quiet Add property row (T-4). A graph opened with no `subjectId`: today's camera (AC-2.4, T-8).
- **Removed inputs.** Type missing or snapshot nil: kind `.person`, empty label, default style (D-5, T-2). Source type missing: no pill (T-2). Graph ref or name empty (T-2). Source title empty: ref in mono (T-1). Opening Subject missing from the graph: no selection, no focus (T-8).
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** The read is the app's own catalog; 50 claims and 30 rows a card is the stated large case (D-14).

### Schema

None — no catalog change.

### Go core

None — S13-01's read carries every field.

### FFI

None —

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Section | `ConclusionPageSectionHost` | `macos/App/Features/Conclusions/PersonSourcesSection.swift` *(new)* | D-9 |
| Group heading | Timeline title link + `CatalogSourceTypePill` + caption | same file | D-10, SS-2 |
| Card | `EvidenceCardViewHost` (`.page`) | `macos/App/Features/EvidenceGraph/EvidenceCardViewHost.swift` *(new)* | D-2, D-3, SS-3 |
| Card footer | `PVButton` (ghost, `.sm`) over `pageGraphActionID` | `PersonSourcesSection.swift` *(new)* | D-11, SS-4 |
| Card wrap | `PVFlowLayout` | same file | D-11 |
| States | the host's spinner, `PVCallout`, `PVEmptyState` (glyph `library`) | host | SS-6, SS-7, SS-W3 |

States: empty (AC-1.5), error (AC-1.6), loading (AC-3.1), refreshing (AC-3.2), large (50 cards, no fold). Accessibility: the group title is a heading button reading `group.accessibilityLabel` ("Hamburg census 1867, Census, 2 records"); each card is the card view's one element ("Person, CPR-7Q2KD, Anna Weber, cited, on Household 214"); the graph button reads "Open Household 214". Tab order: title, then each card's graph button, in reading order (SS-9). The board's frames are described here in words on agreement.

### Strings

L10n keys *(new)*, in `L10n.Conclusions`: `sourcesErrorTitle` = "Sources could not be read"; `sourcesEmptyTitle` = "No records"; `sourcesEmptyMessage` = "This person has no promoted records."; `sourcesRecordCount(_:)` = "%lld records", a plural in `Localizable.xcstrings` with `one` "%lld record" and `other` "%lld records"; `sourcesOpenGraph(_:)` = "Open %@"; `sourcesGroupAccessibility(title:type:count:)` = "%1$@, %2$@, %3$@" (with no type: `sourcesGroupAccessibilityUntyped(title:count:)` = "%1$@, %2$@"); `sourcesCardOnGraph(label:graph:)` = "%1$@, on %2$@". The empty copy follows the board (SS-6). Removed: `conclusions.pageSection.later` (`pageSectionLater`).

## 7. Steps

**STEP-1 — A section declares its keys**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `ConclusionPageSectionHost.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-workspace-place`.
Do: D-1. `fillsWidth` waits for STEP-7 (D-12).
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PersonDetailTests` passes (T-6); Family and Timeline load as before.

**STEP-2 — One builder for a placed Subject**
Files: `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift`, `macos/ProvenenciaTests/PersonSourcesContentTests.swift` *(new)*.
Do: D-4.
Check: T-3 passes; `EvidenceGraphModelTests` pass unchanged.

**STEP-3 — One source type label lookup**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Features/Sources/SourcesModel.swift`, `macos/App/Features/Sources/SourceIdentitySection.swift`.
Do: D-8.
Check: the Sources list and the Source page's identity section show the same type labels (existing Sources tests pass).

**STEP-4 — The section's content and locations**
Files: `macos/App/Features/Conclusions/PersonSourcesContent.swift` *(new)*, `macos/App/Features/Sources/SourcesListNavigation.swift`, `macos/ProvenenciaTests/PersonSourcesContentTests.swift`.
Do: D-5 – D-7.
Check: T-1, T-2 pass.

**STEP-5 — The page chrome on the card paint**
Files: `macos/App/Features/EvidenceGraph/EvidenceCanvasCard.swift`, `EvidenceCardView.swift`, `EvidenceSubjectCard.swift`, `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift`. Skill: `add-ui-component`.
Do: D-2, and `layout(for:)` from D-3.
Check: T-4, T-5 pass.

**STEP-6 — The card on a SwiftUI page**
Files: `macos/App/Features/EvidenceGraph/EvidenceCardViewHost.swift` *(new)*.
Do: D-3, with a `#Preview` of one cited and one uncited page card.
Check: the build has no warning; the preview shows both cards at 264pt with their shadow unclipped.

**STEP-7 — The Sources section**
Files: `macos/App/Features/Conclusions/PersonSourcesSection.swift` *(new)*, `PersonDetailView.swift`, `macos/App/Features/Workspace/WorkspaceLocation.swift` (`fillsWidth`, D-12), `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-ui-component`.
Do: D-9 – D-12, D-14; replace the `.sources` case.
Check: T-6 passes; manual check items 1 – 4.

**STEP-8 — A graph opens on the Subject its location names**
Files: `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `EvidenceGraphModel.swift`, `macos/App/Features/GraphCanvas/GraphCanvasScrollView.swift`, `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`.
Do: D-13, deleting the pending-selection hook.
Check: T-7, T-8 pass; `rg 'PendingGraphSelection|pendingGraphSelections' macos` finds nothing; manual check item 3.

**STEP-9 — Strings**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: the keys in **Strings** (STEP-4 and STEP-7 use them: add each key in the step that first uses it, and finish the catalog entries here); delete `pageSectionLater` once `rg pageSectionLater macos/App` finds only its definition.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-10 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `.agents/skills/add-workspace-place/SKILL.md` if it names `queryKey`.
Do: per the Definition of done. The brief stays open for S13-17.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-16-sources-section.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PersonSourcesContentTests.swift` `groupsKeepTheReadsOrder` *(new)* | The AC-1.1 groups → three groups in order, two cards in the census group (Household 214, 388), the untitled Source's title `.ref("SRC-4KQ7M")`; `[]` → `isEmpty` | AC-1.1, AC-1.5, D-7 |
| T-2 | Swift | same · `headingFooterAndFallbacks` *(new)* | Type id "st-census" in the list → `typeLabel` "Census" and its icon key; unknown id → nil; count "2 records" / "1 record"; `graphRef` "" → nil, graph name "" → `graphTitle` the ref; Subject type missing, or `properties` nil → kind `.person`, empty label; `location` = `.sourcePage(...)`, `graphLocation` = `canvasLocation(sourceId:sourceRef:sourceTitle:graphId:subjectId:)` with the Source's ref and title and the Subject id; the card's accessibility label ends ", on Household 388" | AC-1.2, AC-1.3, AC-2.1 – AC-2.3, D-5, D-6, D-8 |
| T-3 | Swift | same · `cardRowsMatchTheGraph` *(new)* | Observations in wire order (occupation, name, age) → the card's rows age, name, occupation, equal to `EvidenceGraphSnapshot.build`'s rows for the same Subject; none → `isCited` false | AC-1.4, D-4 |
| T-4 | Swift | `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift` `pageChromeDropsTheEditingControls` *(new)* | A cited card with three rows in `.page` → regions hold the three rows, `pageGraphActionID` and `pageSlotActionID` (equal to `promotedFooterParts`' handle and slot), and no `addProperty`; height = the canvas card's height minus the Add property row; its `renderImage` differs from the canvas card's only in the header icon area and the footer content (pixels compared outside those rects); an uncited `.page` card has no `addProperty` | AC-1.4, D-2, D-3 |
| T-5 | Swift | `EvidenceCardLayoutTests.swift` existing cases, `ghostCardMatchesTheLiveLayout`, `ghostPaintsAtGhostOpacity` | Unchanged with `Chrome.ghost` | D-2 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys` (changed) | `.family` → `[.personFamily]`, `.timeline` → `[.personEvents]`, `.sources` → `[.personClaims, .propertiesWorkspace, .sourceTypesList]`, others `[]`; `fillsWidth` per D-12 | AC-3.1, D-1, D-12 |
| T-7 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `openingSubjectIsSelected` *(new)* | Model with `openingSubjectID: "s2"` on a graph with s1, s2 → after prepare `selectedSubjectID == "s2"` and `openingFocus` = s2's card center; `"gone"` → no selection, `openingFocus` nil; nil → as today | AC-2.2, AC-2.4, D-13 |
| T-8 | Swift | `macos/ProvenenciaTests/GraphCanvasScrollViewTests.swift` *(new)* `openingFocusCentersThePoint` | `applyOpeningCamera(nil, contentSize: 4000×4000, focus: (3000, 1200))` in a 1000×800 view → the visible rect's center is (3000, 1200) at magnification 1; focus near an edge → clamped as without focus; focus nil → today's centering | AC-2.2, AC-2.4, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `EvidenceCardLayoutTests`, `EvidenceSubjectCardFooterTests`, `EvidenceGraphModelTests`, the Family and Timeline tests (the key rename).

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

1. Open the dogfood project, a Person with records from at least three Sources, one Source with two Evidence graphs. Open the Sources section. See one group per Source in title order, the two-graph Source with two cards, each heading with the type pill and the record count (AC-1.1, AC-1.2).
2. Compare one card with the same Subject on its Evidence graph: same header, same rows in the same order, no pen, trash or Add property (AC-1.4).
3. Click a Source title: its Source page; Back: Sources. Click a card's graph button: the graph opens centered on that Subject, its card selected; Back: Sources. Open a graph from the Sources list: it opens where it was (AC-2.1, AC-2.2, AC-2.4).
4. Promote a new Subject into the person on a graph; return: the card is there (AC-3.2). In a preview with FakeStore's scripted claims error, see "Sources could not be read" (AC-1.6). Stretch the window across 1600pt and back: the board's arrangement, scroll kept (AC-3.3). VoiceOver on one card: one reading with its graph. Dark mode at both widths.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The AppKit card in a SwiftUI flow lays out at the wrong size | D-3 sizes from `layout(for:)` and offsets the view by the outset; STEP-6's preview |
| Folding `isGhost` into `Chrome` moves the ghost | T-5 runs the ghost tests unchanged |
| Renaming `queryKey` misses a caller | The compiler; T-6 |

**Stop and report if:**

- The board is not agreed: this plan stays `Blocked — on board S13-D6`.
- The board draws a card that is not the Evidence graph's paint (D-2), or moves Remove out of the footer's slot (D-2's regions).
- `EvidenceCardView` does not paint, or paints at the wrong scale, inside `EvidenceCardViewHost` (D-3).
- `applyOpeningCamera` runs before the canvas's content size is known, so a focus cannot be centered (D-13).
- `EvidenceCardText.accessibilityLabel` is not the card view's VoiceOver label (D-7).

## 12. Open questions

1. **The board's answers** to SS-1 (group separation), SS-3 (gap), SS-4 (footer control), SS-6 (empty copy), SS-8 (fold) and SS-W1 (Wide). Recommended: the brief's recommendations; D-10 – D-12 and D-14 are written for them and are refreshed on agreement.

## 13. PR

**Title:** S13-16 — Sources section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-16-sources-section.md`.
