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
**Review:** [`S13-16-sources-section.review.md`](S13-16-sources-section.review.md) — round 1 on the draft

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

A researcher opens a Person's Sources section and sees every record that makes up the person: one card per accepted identity claim, drawn as the Evidence graph draws that Subject, grouped under its Source (title and type) in Source title order. The Source's title opens its Source page; each card names its Evidence graph and opens it with the Subject selected (user, 2026-11-05). Today the section is an EmptyState ("Sources arrives in a later version of Provenencia.").

**Why now.** Slice 6's chrome. The read (`ListPersonClaims`, S13-01) and its cache key (`.personClaims`, with its invalidation) have been in the app, unread, since #655; the section host (S13-13) makes the section a content view only. S13-17 puts Remove on these cards.

**What stays.** The engine: no Go, proto or store change. The Evidence graph canvas and its card look and behave as today. The header, the switcher and the other four sections. The Remove control is S13-17's: this PR draws the card's footer without it.

## 2. User stories

**US-1.** As a researcher, when I open a Person's Sources section, I see each of the person's records under the Source it came from, so that I can tell where everything I know about them was recorded.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann Weber with claims from "Hamburg census 1867" (graphs Household 214 and Household 388), "Hamburger Adressbuch 1861" and an untitled Source SRC-4KQ7M | Three groups, in the read's order: "Hamburg census 1867" with two cards (Household 214, then Household 388), "Hamburger Adressbuch 1861" with one, then SRC-4KQ7M (in mono) with one (R-33, R-34, R-36; SS-1) |
| AC-1.2 | The census group, its Source type id naming the seeded type "Census" | Its heading reads the title, a Chip "Census" with the type's icon, and "2 records" (R-35; SS-2) |
| AC-1.3 | A group whose type id is not in the source types list (the list not loaded yet, or the type deleted meanwhile) | The heading shows no Chip; the title and count stay (D-7) |
| AC-1.4 | A claim on a Subject with five Observations (name, sex, age, occupation, birthplace) | Its card paints the Evidence graph card's header and the five property rows in the graph's order (property label A→Z), with no edit or delete icon and no Add property row (SS-3; D-2, D-3) |
| AC-1.5 | A person with no accepted claim in the read (a delete raced the read) | The EmptyState "No records" · "This person has no promoted records.", id `persons.detail.sources.empty` (SS-6) |
| AC-1.6 | The read fails | The danger Callout "Sources could not be read" with the error, id `persons.detail.sources.error` (SS-7) |

**US-2.** As a researcher, from a Source group or a claim card, I can open the Source page or the Evidence graph, so that I can go to the evidence itself.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Click the "Hamburg census 1867" title | `navigation.go(to: .sourcePage(sourceId:ref:title:))` with that Source's id, ref and title; Back returns to the Sources section (R-37; SS-2) |
| AC-2.2 | Click the Household 388 card's graph control | `navigation.go(to: .evidenceGraph(sourceId:graphId:subjectId:ref:title:sourceTitle:))` with the claim's graph, its Subject id, the graph's ref and name and the Source title; the graph opens with that Subject's card selected; Back returns to Sources (SS-4; D-6, D-13) |
| AC-2.3 | A claim whose graph has no ref (`graphRef` empty) | The footer shows the graph name only (D-6) |

**US-3.** As a researcher, the section loads and refreshes on its own, at both widths.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Open the Person page on Sources | The place warms `.personClaims` for the person (it is the section's `queryKey`), and the section shows its spinner until the first value (SS-7; D-1) |
| AC-3.2 | Promote another Subject into the person on an Evidence graph, then return | The section shows the new claim; the reload keeps the cards drawn with "Refreshing" in the switcher row (S13-01's `refreshes`, unchanged) |
| AC-3.3 | The pane crosses 1600pt | The section's Wide arrangement is the board's SS-W1 answer; the scroll position and focus survive (SS-W1, SS-W2) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `.sources` reads `.personClaims` (`ConclusionPageSection.queryKey`) | STEP-1, AC-3.1 |
| The display content: groups, headings, the placed Subjects, the type label and icon, the locations (SS-1, SS-2, SS-4; R-33 – R-37) | STEP-2, STEP-3, AC-1.1 – AC-1.3, AC-2.1 – AC-2.3 |
| The Evidence graph card read-only: a page chrome on the one paint, and a SwiftUI host (SS-3) | STEP-4, STEP-5, AC-1.4 |
| The Sources section view on the section host, its footer (graph control), states, keyboard and VoiceOver, dark mode (SS-3 – SS-10) | STEP-6, AC-1.5, AC-1.6, AC-2.2 |
| The Wide arrangement (SS-W1 – SS-W3) | STEP-6, AC-3.3 |
| The Evidence graph selects the Subject a location names (user, 2026-11-05) | STEP-7, AC-2.2 |
| Strings | STEP-8 |
| Docs | STEP-9 |

| Out | Home |
| --- | --- |
| Remove on a card, its confirm, landing on the Persons list (RC-1 – RC-7) | S13-17 |
| The Source date and sorting by it | [Spike 15's Source date](../../spike-15/source-provenance-date.md) |
| How the records connect (links between cards) | S13-28 (board S13-D10) |
| A provenance control on a card | not planned this spike (R-13b) |
| Scrolling the Evidence graph to the selected card on open | not planned (D-13; the canvas has no reveal API) |
| Edit, delete, Add property or Promote from the page's card | not planned (they stay on the Evidence graph; brief § 2.1) |
| A fold, filter, search or sort | not planned (brief SS-8, § 2.1), unless the board's SS-8 answer folds |
| Any Go, proto or store change | not planned: S13-01's read carries everything (D-4) |
| Event and Place pages with a Sources section | their redesign (R-1) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| The Sources place | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` | `case .sources:` draws `PVEmptyState` with `L10n.Conclusions.pageSectionLater`. Family and Timeline cases call `PersonFamilySection(entityId:)` and `PersonTimelineSection(entityId:)` |
| Section query key | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection.queryKey(project:entityId:)` | `.family → .personFamily`, `.timeline → .personEvents`; `.vitals, .gallery, .sources` → nil. `fillsWidth`: `.timeline, .gallery` true |
| Place warm | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` (`.personDetail` spec) | Appends `location.pageSection.flatMap { $0.queryKey(...) }` to the warmed keys (S13-12 D-8): a non-nil `.sources` key is warmed with no other change |
| Section host | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` `ConclusionPageSectionHost` | Gets the section's handle, starts the read, draws the spinner, error Callout (`<root>.error`) and EmptyState (`<root>.empty`) around `body(content)`; preconditions a non-nil key. Family and Timeline use it |
| Claims read | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonClaimGroup{sourceID, sourceRef, title, sourceTypeID, claims}`, `CatalogPersonClaim{claimID, graphID, graphRef, graphName, subject, observations}`; `GoStore.listPersonClaims` | Groups in Source title order, claims in graph name order (Go, `core/database/persondetail/claims.go` `SortByTitle`); observations in wire order (S13-01 D-14: "S13-16 sorts for display") |
| Claims key | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `case .personClaims` | `.sessionFresh`, evicted when hidden; refreshes on the person's touched handles, the graphs it holds, and property labels. No view reads it |
| FakeStore | `macos/App/Platform/FakeStore.swift` `seededPersonClaims`, `listPersonClaims` | Scripted groups by entity id |
| Placed Subject | `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift` `EvidenceGraphPlacedSubject`; `EvidenceGraphSnapshot.build` | `build` maps a Subject's type to `EvidencePrimaryKind(rawValue: type.key)`, `typeLabel: type.label`, sorts its Observations with the private `observationDisplayOrder` (label, key, ref) and sets `isCited`. The sort is private, so a second builder would copy it |
| The card paint | `macos/App/Features/EvidenceGraph/EvidenceCardView.swift` `EvidenceCardView` (a `GraphCanvasItemView`), `static func subject(...)`, `paintHeader`, `recordFooter`, `paintFooter` | One measure-and-paint for layout and paint. The header always paints the pen and trash icons; a cited card paints an Add property row; the footer paints Promote or the handle row with the Unpromote square. Hit testing is the canvas's (`EvidenceSubjectCard.actionTargets`), not the view's |
| The card's model | `macos/App/Features/EvidenceGraph/EvidenceCanvasCard.swift` `EvidenceCanvasCard{placed, text, state, nameOrders, projectDir, cardinalityByKey}` | `State.resting` exists; the ghost builds a card with `EvidenceCardText(subject:)` |
| The canvas item view | `macos/App/Features/GraphCanvas/GraphCanvasItemView.swift` | Its frame is the content plus `paintOutset` (24pt) on every side; `draw(_:)` paints, so it draws in any view hierarchy; `intrinsicContentSize` is the outset frame |
| Subject types | `macos/App/Features/Properties/PropertiesSnapshot.swift` `PropertiesSnapshot{types, presentationsByKey, …}`, key `.propertiesWorkspace` | The Evidence graph reads its card types, presentations and `cardinalityByKey` here |
| Source types | `.sourceTypesList`; `macos/App/Platform/GenealogyStore.swift` `[CatalogSourceType].iconKey(forTypeID:)`; `macos/App/Features/Sources/SourcesModel.swift` `typeLabel(for:)` | The Sources list labels a type from the cached list (S13-01 D-4) |
| Locations | `macos/App/Features/Workspace/WorkspaceLocation.swift` `.sourcePage(sourceId:ref:title:)`, `.evidenceGraph(sourceId:graphId:subjectId:ref:title:sourceTitle:)` | Both exist. A graph location's `subjectId` is carried and persisted, but the canvas does not read it: `WorkspaceDestinationHost` passes only `graphID` and `sourceID` to `EvidenceGraphView`, and only Promote and the composer read `subjectId` |
| Graph selection | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` (`.sourceGraph`); `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `init(graphID:sourceID:session:store:userID:)`; `EvidenceGraphModel.selectSubject(id:)` | `selectSubject(id:)` replaces the selection with one card. No API scrolls the canvas to a card |
| Kit | `PVSectionHeader(title:meta:aside:actions:)`, `PVChip(text:isSelected:glyph:)`, `PVFlowLayout`, `PVButton`, `PVEmptyState`, `PVCallout` | All exist |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys`; `EvidenceCardLayoutTests.swift`; `EvidenceSubjectCardFooterTests.swift`; `CatalogQueryRegistryTests.swift` `personClaimsLoadThroughTheStore` | `sectionQueryKeys` expects `.sources` → nil and `fillsWidth` only for Timeline and Gallery |

**Precedent.** S13-13 ([#691](https://github.com/mendahu/provenencia/pull/691)) added the Timeline on `ConclusionPageSectionHost` with one `queryKey` case and a `…Content` value built in the section's `make` closure from captured inputs (`orders`, `widthClass`); copy that shape for `PersonSourcesSection` and `PersonSourcesContent`. S13-12 ([#676](https://github.com/mendahu/provenencia/pull/676)) wired `queryKey` into the place. Do not copy Family's hand-built host; it is gone.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ConclusionPageSection.sources.queryKey` returns `.personClaims(project:entityId:)` | One source of truth for the section's read (S13-12 D-8, S13-13 D-10); the place warms it and the switcher's refresh follows it | A second warm path in `PlaceRegistry` |
| D-2 | The card is the Evidence graph's paint (`EvidenceCardView`), not a second card. `EvidenceCanvasCard` gains `chrome: Chrome` *(new)*, `enum Chrome { case canvas, page }`, default `.canvas`. `.page` paints the header without the pen and trash icons, skips the Add property row (cited and uncited), and lays out the footer slot (same height, same hairline above it) with no paint inside and its region recorded as `EvidenceSubjectCard.pageFooterID` *(new)*. Every `.canvas` layout and paint is unchanged | R-35 and the brief's non-goal "a second Subject card"; one measure keeps the two hosts' rows identical | A SwiftUI look-alike card (two paints drift); a `State` flag (state is interaction, chrome is the host) |
| D-3 | The host is `EvidenceSubjectCardImage` *(new)*, an `NSViewRepresentable` in `Features/EvidenceGraph/EvidenceSubjectCardImage.swift` *(new)* that makes `EvidenceCardView(card:actions: [])` and replaces it when the card value changes (`updateNSView` compares the stored card). Its SwiftUI frame is `layout.size`; the view is offset by `-GraphCanvasItemView.paintOutset` on both axes so the shadow paints outside the frame. It exposes `layout` to its caller | The paint is AppKit; `draw(_:)` already paints outside the canvas (`renderImage` uses it) | Rendering to an image (`renderImage`) per card: a likeness loads late and would need a re-render path |
| D-4 | No engine change. Kind is `.person` for every claim (a Person's claims are person Subjects); the type label, the presentation and `cardinalityByKey` come from `PropertiesSnapshot` (`.propertiesWorkspace`, started by the section); Observations sort with `EvidenceGraphSnapshot.observationDisplayOrder`, made `static` internal (not private) and used by both builders | S13-01 left placement to Swift (working rule "Each read returns data, not layout"); the card's order must match the graph's | Sorting in Go; a second sort |
| D-5 | `EvidenceGraphPlacedSubject.claimCard(_ claim: CatalogPersonClaim, types: [CatalogSubjectType]) -> EvidenceGraphPlacedSubject` *(new)* in `EvidenceGraphSnapshot.swift`: gridX/Y 0, `isCited` from the Observations, `membership` nil (the footer is the page's), `eventTitle` nil, `typeLabel` from the Subject's type id (empty when not found) | The card needs a placed Subject; one builder beside `build` | Building it in the view |
| D-6 | `PersonSourcesContent` *(new)* in `Features/Conclusions/PersonSourcesContent.swift` *(new)*: `init(groups: [CatalogPersonClaimGroup], sourceTypes: [CatalogSourceType], properties: PropertiesSnapshot?, orders: NameFormatOrders)`, `Equatable`, with `groups: [Group]` where `Group{id (sourceID), title: PVTitleSource (.label(title), else .ref(sourceRef)), typeLabel: String?, typeIconKey: String?, countText, location: WorkspaceLocation, cards: [Card]}` and `Card{id (claimID), card: EvidenceCanvasCard (chrome .page, state .resting, the session's name orders and project folder), graphTitle (graphName, else graphRef), graphRef: String?, graphLocation: WorkspaceLocation, accessibilityLabel}`; `isEmpty` when no group | The Timeline and Family pattern (`…Content` from the read, built in `make`); testable without a view | Mapping in the view body |
| D-7 | The type label is the cached source type's label; a type id missing from `.sourceTypesList` gives no Chip (`typeLabel` nil). The section starts `.sourceTypesList` and `.propertiesWorkspace` with `session.query` in its `.task` and reads their handles in `make` | S13-01 D-4 (user, 2026-10-31); a missing label is not an error | Blocking the section on the lists |
| D-8 | The view is `PersonSourcesSection(entityId:)` *(new)* in `Features/Conclusions/PersonSourcesSection.swift` *(new)* on `ConclusionPageSectionHost<[CatalogPersonClaimGroup], PersonSourcesContent, PersonSourcesList>` with `errorTitle: L10n.Conclusions.sourcesErrorTitle`, `emptyTitle: L10n.Conclusions.sourcesEmptyTitle`, `emptyMessage: L10n.Conclusions.sourcesEmptyMessage` *(new keys)*; `PersonSourcesList` *(new, same file)* draws the groups. Ids: `persons.detail.sources.group.<sourceID>`, `.title.<sourceID>`, `.card.<claimID>`, `.graph.<claimID>` | The host owns the root, `.empty` and `.error` ids | A view that reads the handle itself |
| D-9 | The group heading is `PVSectionHeader(title:meta:aside:)`: the title a `PVButton` (ghost) with `PVTitleText`, opening `group.location`; the type a `PVChip(text:isSelected: false, size: .sm, glyph:)` from `typeIconKey`; the meta `L10n.Conclusions.sourcesRecordCount(_:)` *(new, plural)*. **Refreshed from the board's SS-1 and SS-2 answers** | Kit first (brief inventory) | A custom heading row |
| D-10 | Cards wrap with `PVFlowLayout` at their 264pt width, top-aligned; the footer's graph control is a `PVButton` (ghost, `.sm`) overlaid on the card's `pageFooterID` region, its title the graph's name and its ref in mono. **The gap, and whether the whole footer is the control, follow the board's SS-3 and SS-4 answers** | The paint owns the card; SwiftUI owns the controls, so hover, focus ring and Return are the kit's (SS-5) | Hit testing inside the AppKit view |
| D-11 | Wide (SS-W1): **the board's answer, recorded here on agreement.** (A) sets `ConclusionPageSection.sources.fillsWidth` true, and the host's states keep the band (`pvKeepsPageBand`, already in the host); (B) leaves it false; (C) adds a secondary fill and is a scope change that goes back to the map | The user left it to the board (2026-11-05) | — |
| D-13 | `EvidenceGraphView` takes `selectingSubjectID: String?` *(new)*, which `WorkspaceDestinationHost` passes from `navigation.currentLocation.subjectId`; the view calls `model.selectSubject(id:)` in `.task(id: selectingSubjectID)` once the graph's snapshot holds that Subject (the model's `selectOnOpen(subjectID:) async` *(new)*: it waits with `session.readyValue` on the graph key, then selects only when the snapshot contains the id). The canvas is not scrolled to the card | The user asked for the graph opened at the Subject (2026-11-05); selection is the existing state. Scrolling needs a viewport API the canvas does not have | A scroll-into-view API in this PR (a canvas change of its own) |
| D-12 | No fold (brief SS-8 recommendation), **unless the board folds**; then `ConclusionShowMoreRow` and the count the board names | One card is one record; Family and Timeline fold with the shared row | A new fold control |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `EvidenceCanvasCard.Chrome` *(new)* | type | `EvidenceCanvasCard.State`, the ghost initializer | Extends the one card model; state is interaction, chrome is the host |
| `EvidenceSubjectCard.pageFooterID` *(new)* | constant | `openHandleActionID`, `promoteActionID` | One more region id on the same layout |
| `EvidenceSubjectCardImage` *(new)* | component (feature) | `GraphCanvasItemLayer` (the canvas host) | The canvas host pans, zooms and hit-tests a whole graph; a page needs one card at rest. Searched `NSViewRepresentable` over `GraphCanvasItemView`: none |
| `EvidenceGraphPlacedSubject.claimCard(_:types:)` *(new)* | func | `EvidenceGraphSnapshot.build` | Shares `observationDisplayOrder`; `build` needs positions and a whole graph |
| `PersonSourcesContent` *(new)* | type | `PersonTimelineContent`, `PersonFamilyContent` | Same pattern, different content |
| `PersonSourcesSection`, `PersonSourcesList` *(new)* | views | `PersonTimelineSection` | Same host, different body |
| `L10n.Conclusions.sources*` *(new)* | keys | `timelineErrorTitle`, `timelineEmptyTitle` | Same namespace, one set per section |
| `PersonSourcesContentTests` *(new)* | test file | `PersonTimelineContentTests` | One file per content type |

**Builds on.** `ConclusionPageSectionHost` and `ConclusionPageSection.queryKey` (S13-13 D-10, S13-12 D-8); `.personClaims` and its `refreshes` (S13-01); the Evidence graph paint and `EvidenceCardText`; `PropertiesSnapshot`; the kit's `PVSectionHeader`, `PVChip`, `PVFlowLayout`, `PVButton`.
**Deletes.** The `.sources` EmptyState in `PersonDetailSections`. `L10n.Conclusions.pageSectionLater` loses its last caller once Sources is built (Family, Timeline and Gallery already replaced theirs): delete the key and its `Localizable.xcstrings` entry (check with `rg pageSectionLater macos/App`).
**Next adds.** S13-17 adds Remove to `Card` and an `IconButton` at the end of the footer; S13-28 places its view in or beside `PersonSourcesList` (board S13-D10); an Event or Place page with sections reuses `EvidenceSubjectCardImage` with its own kind; a future read-only card anywhere (a Promote preview) uses `.page` chrome.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Sources title → Source page | `macos/App/Features/Sources/SourcePageView.swift` | An existing location; a Source deleted meanwhile falls back to the Sources list (`fallbackToSectionRoot`, today's behavior) |
| Card → Evidence graph with `subjectId` | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` | New: the canvas now reads `subjectId` (D-13). A Subject no longer on the graph selects nothing; a graph deleted meanwhile: the graph's own missing-id handling (unchanged). Back and restore to a graph entry with a `subjectId` select it again, which is the intent |
| A Person page opened on Sources for another person | `PersonSourcesSection` | The host's key carries the entity id, so the handle changes with it |
| `EvidenceCardView` in a SwiftUI page | `GraphCanvasItemView.draw(_:)` | Paints in any hierarchy; its likeness loads run in `viewDidMoveToWindow` with `projectDir` set |

### Scale

At the map's Scale for the section (brief § 2: typical 3 – 20 claims, large 50, up to 30 rows a card): the warm read issues no SQL (S13-01, T-6 there). Per card, one `EvidenceCardView` measures once per value change (`updateNSView` replaces the view only when the card value differs); 50 cards × ≤ 30 rows is the measure the graph already does for a 50-card graph. The content is rebuilt in `make` when the claims, the source types, the properties snapshot or the name orders change; each rebuild is O(claims × observations log observations). No new write, key or invalidation.

### Edges

- **Assumptions.** `observationDisplayOrder` keeps same-key rows adjacent (the conflict bracket needs it, `EvidenceCitedPropertyMarks.conflictRuns`): `claimCard` uses the same function, T-3 holds the order. The paint's region ids are unique per card (`.page` adds one id and drops edit, delete and addProperty; T-4).
- **Neighbors.** `.canvas` chrome: layout and regions unchanged (T-5 runs the existing `EvidenceCardLayoutTests` cases unchanged). An uncited claim (a Subject with no Observation): `.page` draws the header and the footer slot with no quiet Add property row (T-4).
- **Removed inputs.** A claim's Subject type missing from `PropertiesSnapshot`: `typeLabel` empty, presentation nil, the card paints the default person style (T-2). A Source type missing: no Chip (AC-1.3, T-2). Graph ref empty: name only (AC-2.3, T-2). Source title empty: the ref in mono (AC-1.1, T-1).
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** The read is the app's own catalog; 50 claims and 30 rows a card are the stated large case (no fold, D-12). No bound beyond the read's.

### Schema

None — no catalog change.

### Go core

None — S13-01's read carries every field (D-4).

### FFI

None —

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Section | `ConclusionPageSectionHost` | `macos/App/Features/Conclusions/PersonSourcesSection.swift` *(new)* | D-8; replaces the `.sources` case in `PersonDetailView.swift` |
| Group heading | `PVSectionHeader`, `PVButton` (ghost) + `PVTitleText`, `PVChip` (`.sm`, type glyph) | same file | D-9, SS-2 |
| Card | `EvidenceSubjectCardImage` (`.page` chrome) | `macos/App/Features/EvidenceGraph/EvidenceSubjectCardImage.swift` *(new)* | D-2, D-3, SS-3 |
| Card footer | `PVButton` (ghost, `.sm`) over `pageFooterID` | `PersonSourcesSection.swift` | D-10, SS-4; S13-17 adds the Remove `PVIconButton` at its end |
| Card wrap | `PVFlowLayout` | same file | D-10 |
| States | the host's spinner, `PVCallout`, `PVEmptyState` (glyph `library`) | host | SS-6, SS-7, SS-W3 |

States: empty (AC-1.5), error (AC-1.6), loading (AC-3.1), refreshing (AC-3.2), large (50 cards, no fold). Accessibility: the group title is a heading button ("Hamburg census 1867, Census, 2 records"); each card is an accessibility group labelled `Card.accessibilityLabel` ("Person, CPR-7Q2KD, Anna Weber, on Household 214"), its rows read by the card view as on the graph; the graph control reads "Open Household 214" (`L10n.Conclusions.sourcesOpenGraph(_:)` *(new)*). Tab order: title, then each card's control (SS-9). The board's frames are described here in words on agreement.

### Strings

L10n keys *(new)*, in `L10n.Conclusions`: `sourcesErrorTitle` = "Sources could not be read"; `sourcesEmptyTitle` = "No records"; `sourcesEmptyMessage` = "This person has no promoted records."; `sourcesRecordCount(_:)` = "%lld records" (plural: "1 record"); `sourcesOpenGraph(_:)` = "Open %@"; `sourcesCardAccessibility(type:ref:name:graph:)` = "%1$@, %2$@, %3$@, on %4$@". The empty copy and any heading wording follow the board (SS-6). Removed: `conclusions.pageSection.later` (`pageSectionLater`) once its last caller goes.

## 7. Steps

**STEP-1 — The section reads the claims**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-workspace-place`.
Do: `.sources` → `.personClaims(project:entityId:)` in `queryKey` (D-1). Update `sectionQueryKeys` for `.sources`. Leave `fillsWidth` until STEP-6 (D-11).
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PersonDetailTests` passes.

**STEP-2 — One builder for a placed Subject**
Files: `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift`, `macos/ProvenenciaTests/PersonSourcesContentTests.swift` *(new)*.
Do: make `observationDisplayOrder` internal static; add `EvidenceGraphPlacedSubject.claimCard(_:types:)` (D-4, D-5).
Check: T-3 passes; `EvidenceGraphModelTests` pass unchanged.

**STEP-3 — The section's content**
Files: `macos/App/Features/Conclusions/PersonSourcesContent.swift` *(new)*, `macos/ProvenenciaTests/PersonSourcesContentTests.swift`.
Do: `PersonSourcesContent` (D-6, D-7) with the locations (AC-2.1, AC-2.2).
Check: T-1, T-2 pass.

**STEP-4 — The page chrome on the card paint**
Files: `macos/App/Features/EvidenceGraph/EvidenceCanvasCard.swift`, `EvidenceCardView.swift`, `EvidenceSubjectCard.swift`, `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift`. Skill: `add-ui-component` (snowflake extension).
Do: `Chrome` and `pageFooterID` (D-2). `paintHeader` takes whether to paint the icons; `subject(...)` skips the Add property rows and records `pageFooterID` for `.page`.
Check: T-4, T-5 pass.

**STEP-5 — The card on a SwiftUI page**
Files: `macos/App/Features/EvidenceGraph/EvidenceSubjectCardImage.swift` *(new)*.
Do: D-3. A `#Preview` with one cited and one uncited card.
Check: the preview renders both cards at 264pt with their shadow; the build has no warning.

**STEP-6 — The Sources section**
Files: `macos/App/Features/Conclusions/PersonSourcesSection.swift` *(new)*, `PersonDetailView.swift`, `WorkspaceLocation.swift` (`fillsWidth` per D-11), `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-ui-component`.
Do: D-8 – D-12; replace the `.sources` case.
Check: T-6 passes; the manual check items 1 – 4 at Regular and Wide.

**STEP-7 — The graph selects the Subject it was opened on**
Files: `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `EvidenceGraphModel.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`.
Do: D-13.
Check: T-7 passes; manual check item 3.

**STEP-8 — Strings**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: the keys in **Strings**; delete `pageSectionLater` and its catalog entry when `rg pageSectionLater macos/App` finds only its definition.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note).
Do: per the Definition of done. The brief stays open for S13-17.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-16-sources-section.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PersonSourcesContentTests.swift` `groupsKeepTheReadsOrder` *(new)* | The AC-1.1 groups → three groups in order, two cards in the census group (Household 214, 388), the untitled Source's title `.ref("SRC-4KQ7M")`; `isEmpty` false; `[]` → `isEmpty` | AC-1.1, AC-1.5, D-6 |
| T-2 | Swift | same · `headingAndFooterFromTheCachedLists` *(new)* | Type id "st-census" in the list → `typeLabel` "Census" and its icon key; an unknown id → nil; count text "2 records" / "1 record"; a claim with `graphRef` "" → `graphRef` nil; a Subject type missing from the snapshot → `typeLabel` ""; the locations equal `.sourcePage(...)` and `.evidenceGraph(... subjectId: claim.subject.id ...)` | AC-1.2, AC-1.3, AC-2.1 – AC-2.3, D-7 |
| T-3 | Swift | same · `claimCardSortsRowsAsTheGraph` *(new)* | Observations in wire order (occupation, name, age) → the card's rows by label (age, name, occupation), as `EvidenceGraphSnapshot.build` orders the same Subject; `isCited` false with none | AC-1.4, D-4, D-5 |
| T-4 | Swift | `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift` `pageChromeDropsTheEditingControls` *(new)* | A cited card with three rows in `.page` → regions hold the three rows and `pageFooterID`, no `edit`, `delete`, `addProperty`, `openHandle`, `unpromote`; height = the canvas height minus the Add property row; an uncited card in `.page` → no `addProperty` | AC-1.4, D-2 |
| T-5 | Swift | `EvidenceCardLayoutTests.swift` existing cases | `.canvas` (default) layouts unchanged | D-2 |
| T-7 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `selectOnOpenSelectsTheSubject` *(new)* | A graph with Subjects s1, s2; `selectOnOpen(subjectID: "s2")` → `selectedSubjectID == "s2"`; `"gone"` → selection unchanged | AC-2.2, D-13 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys` (changed) | `.sources` → `.personClaims(project, "e1")`; `fillsWidth` per D-11 | AC-3.1, D-1, D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `EvidenceCardLayoutTests`, `EvidenceSubjectCardFooterTests`, `EvidenceGraphModelTests` (the paint and the snapshot builder).

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
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open the dogfood project, a Person with records from at least three Sources, one Source with two Evidence graphs. Open the Sources section. See one group per Source in title order, the two-graph Source with two cards, each heading with the type Chip and the record count (AC-1.1, AC-1.2).
2. Compare one card with the same Subject on its Evidence graph: same header, same rows in the same order, no pen, trash or Add property (AC-1.4).
3. Click a Source title: its Source page; Back: Sources. Click a card's graph control: the graph with that Subject's card selected; Back: Sources (AC-2.1, AC-2.2).
4. Promote a new Subject into the person on a graph; return: the card is there (AC-3.2). Make the read fail (FakeStore's scripted error in a preview, or a project folder made unreadable): the danger Callout "Sources could not be read" (AC-1.6). Stretch the window across 1600pt and back: the arrangement is the board's, the scroll stays (AC-3.3). Dark mode at both widths.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The AppKit card in a SwiftUI flow lays out at the wrong size (the paint outset) | D-3 sizes to `layout.size` and offsets by the outset; STEP-5's preview checks it |
| A `.page` change moves the canvas card | T-5 keeps every canvas layout test unchanged |
| The card's VoiceOver element and the SwiftUI footer button read twice | The card view is the group's element; the footer button is a separate element after it (SS-9); checked in item 4 with VoiceOver |

**Stop and report if:**

- The board is not agreed: this plan stays `Blocked — on board S13-D6`.
- The board's SS-W1 answer is (C), a secondary column: that is a scope change for the map (D-11).
- The board draws a card that is not the Evidence graph's paint (a different card): D-2 is wrong.
- A claim's Subject type key is not `person` in the dogfood project (D-4's kind).
- `EvidenceCardView` does not paint, or paints at the wrong scale, inside an `NSViewRepresentable` (D-3).
- `EvidenceGraphSnapshot.build`'s order cannot be shared without changing a canvas test.

## 12. Open questions

1. **The board's answers** to SS-1 (group separation), SS-2 (title control), SS-3 (gap), SS-4 (footer control), SS-6 (empty copy), SS-8 (fold) and SS-W1 (Wide). Recommended: the brief's recommendations; D-9 – D-12 are written for them and are refreshed on agreement.

## 13. PR

**Title:** S13-16 — Sources section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-16-sources-section.md`.
