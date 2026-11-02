# S13-13 — Timeline section

**Status:** Blocked — on board S13-D4
**Spike / slice:** Spike 13 · Slice 4 — Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D4](../design/S13-D4-timeline-section.md), board not agreed yet
**Board:** — (not drawn yet). When it is agreed, the brief's **Board** line names the Claude Design turn and frames; this plan then describes every frame in words under **Design → macOS**, so an implementer who cannot open the board can still build it
**Depends on:** S13-05 (landed [#680](https://github.com/mendahu/provenencia/pull/680)), S13-12 (landed [#676](https://github.com/mendahu/provenencia/pull/676)), S13-D4 (board)
**Blocks:** —
**Branch:** `s13-13-timeline-section`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go or proto change unless the board asks for another participant order (§ 12, question 6)
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-05 and S13-12 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 3. Timeline](../person-detail.md#3-timeline) (R-2, R-3, R-23 – R-28); brief [S13-D4](../design/S13-D4-timeline-section.md); the S13-05 plan [`S13-05-person-timeline-read.md`](S13-05-person-timeline-read.md) and the S13-12 plan [`S13-12-family-section.md`](S13-12-family-section.md), with their **Landed** notes; working rules: "Engine PRs run ahead of their slice" (data, not layout), "Each section reads its own query", "Timeline is direct participation, any role", "One person tile, built once", "Colors reuse `PersonGenderTreatment`", "A timeline tile opens the participant on Timeline", "Timeline lists fold", "Sections are places"
**Planned:** 2026-11-02 against `main` @ `51dccd4` (draft; frame-dependent Decisions wait on the board)
**Review:** [`S13-13-timeline-section.review.md`](S13-13-timeline-section.review.md)

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

A researcher opens a Person's Timeline section and reads the person's life as a list of events, earliest first: each event's picture, title, date and place, the person's role in it, and the other people who took part as the same person tiles as on Family; undated events sit in an Undated group at the end. A click on an event opens its Event page, and a click on a participant opens that participant's Timeline. It fills the empty place S13-10 left (S13-D1 PPF-10) with the data S13-05 reads, and makes the S13-12 person tile a recipe, since this is its second caller.

**Why now.** Slice 4's Check is bar 4 ("A life in order"); S13-05 has the read, S13-12 has the tile, and this is the last PR of the slice.

**What stays.** The header, the switcher, Vitals, Family, Gallery and the Sources empty place; the person tile's size, content and colors (Family looks the same); `ListPersonEvents`, its order, its roles and its filtering (S13-05); the Events list and Event page; every other link into a Person page.

## 2. User stories

**US-1.** As a researcher, when I open a Person's Timeline, I see every event they took part in, in date order, so that I read their life from birth to death.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann (brief § 5 sample, events 1 – 13) | Events 1 – 11 in that order, then an "Undated" group with 12 (Residence at Altona) and 13 (Migration), in the order the read returns them (R-23, R-24; D-6) |
| AC-1.2 | Luise's birth names no father | It is not listed (R-23; S13-05 filters it) |
| AC-1.3 | Event 4 | Title "Marriage of Johann Weber and Maria Schulz" (`EventTitleDisplay`), "4 Jun 1844 · Bremen" (`EventSecondaryDisplay`), Johann's role "Subject", tiles Maria Schulz and Friedrich Weber (R-25, R-27; D-5, D-6) |
| AC-1.4 | Event 8: Johann is Witness and Informant | Both roles shown, in the read's order, joined as the board words it (D-5; § 12 question 3) |
| AC-1.5 | Event 1 has a lead picture; event 11 has none | Event 1 shows its picture; event 11 shows the Event mark on the sunken tone (R-26 as updated 2026-11-02; D-7) |
| AC-1.6 | Event 13: no date, no place | The date and place line is not drawn |
| AC-1.7 | A person with no promoted events (`[]`) | The EmptyState of TL-10 |
| AC-1.8 | The read fails with `conclusiondetails.not_found` | A danger Callout in the section titled "Timeline could not be read" with the mapped message; the header and switcher still show (TL-11) |
| AC-1.9 | Only undated events | The Undated group is the whole list; its caption shows or not as the board answers TL-3 |

**US-2.** As a researcher, I can move from an event to its page and from a participant to their life.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | On Johann's Timeline, click Anna Weber's tile on event 1 | Anna's Person page opens on Timeline; breadcrumbs *Persons › Weber, Anna › Timeline* (`PersonsList.location` gives the sorted title, D-1) |
| AC-2.2 | Then Back | Johann's Timeline again |
| AC-2.3 | Click event 4's title | The Event page of that marriage opens, titled as the Events list titles it (D-2) |
| AC-2.4 | VoiceOver on Anna's tile under event 1 | "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851", a button (the tile's spoken label with the role term's label, D-5) |
| AC-2.5 | Tab from event 1's title | Focus moves to its tiles in order, then to event 2's title (TL-12) |

**US-3.** As a researcher, long lives and crowded events stay readable without hiding anything.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | An event with 41 other participants | 15 tiles, then "Show 26 more"; pressing it shows all 41 (D-8) |
| AC-3.2 | A person with 120 dated events | 50 events, then "Show 70 more"; pressing it shows all 120; the Undated group folds the same way on its own (D-8) |
| AC-3.3 | Unfold a list, then open another person and come Back | The list is folded again (D-8, as Family) |

**US-4.** As a researcher, the section stays current while I work.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Johann's Timeline is showing; a write touches Maria Schulz's handle | The timeline reloads; the list stays drawn and the switcher row shows "Refreshing" until it lands (D-3; the descriptor S13-05 added) |
| AC-4.2 | Opening a Person straight on Timeline (from a tile, or Back) | The timeline read starts with the page, not after the detail read (D-3) |

## 3. Scope

| In | Covered by |
| --- | --- |
| The person tile becomes the `PVPersonTile` recipe, Family unchanged (TL-6 inventory, D-4) | STEP-1 |
| One "Show N more" row for Family and Timeline (TL-7, D-9) | STEP-2 |
| The Timeline section's query key: warm, read, refresh (D-3) | STEP-3, AC-4.1, AC-4.2 |
| The section's content model: rows, roles, participants, locations, folds, accessibility (TL-1 – TL-9, TL-12; D-1, D-2, D-5 – D-8) | STEP-4, AC-1.1 – AC-1.6, AC-2.1, AC-2.3, AC-2.4, AC-3.1 – AC-3.3 |
| The section view and its states (TL-1 – TL-13, TL-W1 – TL-W3; D-10, D-11) | STEP-5, STEP-6, AC-1.7 – AC-1.9 |
| R-26 now shows the event's lead picture (user, 2026-11-02) | planning commit; STEP-4 (D-7) |

| Out | Home |
| --- | --- |
| A scaled time axis, chart, zoom or swimlane | not planned (brief § 2.1) |
| Age at each event, gaps between events | not planned this spike (brief § 2.1) |
| Filtering or searching the timeline | not planned |
| Events reached only through a relative or a place | not planned (R-23) |
| The provenance control on a row, role or tile | not planned this spike (R-13b) |
| Editing events or participations from the section | not planned this spike |
| Renaming the tile's "this person" L10n keys (`familyThisPerson`, `familyThisPersonSpoken`) that the recipe reads | not planned; the Timeline has no focus tile, and the keys keep their copy |
| A participant order other than the read's (by role or title) | a follow-up engine change in this PR's STEP-0, only if the board asks (§ 12 question 6); otherwise not planned |
| `EventSubject` folding into {role, `PersonSummary`} | logged by S13-05 in [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Timeline section | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` | `.timeline, .sources` draw one `PVEmptyState` with `L10n.Conclusions.pageSectionLater`. `.family` draws `PersonFamilySection(entityId:)` |
| Section key | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection.queryKey(project:entityId:)` | `.family → .personFamily`; every other section nil. Read by `PlaceRegistry` Spec `.personDetail` (warms it) and `ConclusionDetailLoaded.sectionRefreshing` (`ConclusionDetailPage.swift`) |
| Timeline key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.personEvents(project:entityId:)`; `CatalogQueryRegistry.swift` its descriptor | `.sessionFresh`, `.evict` when hidden, `.allCached(.personEvents)`, refreshes on any touched handle; loads `store.listPersonEvents`. No view queries it |
| Timeline data | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonEvent` (`event: CatalogEventSummary`, `roles: [CatalogConclusionValue]`, `participants: [CatalogEventParticipant]`, `dated`), `CatalogEventParticipant` (`person: CatalogPersonSummary`, `roles`) | Rows in Go's date order, undated last and flagged; roles are `.term(id:key:label:origin:)` with labels filled in Go; participants in participation-ref order, never the person |
| Person tile | `macos/App/Features/Conclusions/PersonTile.swift` `PersonTile`, `PersonTileContent`, private `PersonTileButtonStyle`, `PersonTileButtonBody` | A snowflake. `PersonTileContent.init(summary:role:isFocus:locale:orders:)`; the role is spoken only (`spokenLabel`), never drawn. `PersonTile.height` = 92. Callers: `PersonFamilySection.swift` (`PersonFamilyDiagram`), `PersonFamilyContent.swift` (`Tile.content`), `PersonFamilyLayout.swift` (`PersonTile.height`); tests `macos/ProvenenciaTests/PersonTileTests.swift` |
| Show more | `macos/App/Features/Conclusions/PersonFamilySection.swift` private `PersonFamilyMoreRow`; `macos/App/Platform/L10n.swift` `L10n.Conclusions.familyShowMore(_:)` (key `conclusions.family.showMore`, "Show %lld more") | A ghost `PVButton` (`.pv(.ghost, size: .sm)`) with Return handling; Family only |
| Folds | `macos/App/Features/Conclusions/PersonFamilyContent.swift` `foldLimit` (30), `Fold`, `expanded`; `PersonFamilySection` keeps `@State expanded`, reset `.onChange(of: entityId)` | The pattern to copy for the Timeline's folds |
| Read states | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPresentation` | Generic `init<Value>(_ read: QueryRead<Value>, isFetching:content:)` (S13-12 D-10) |
| Event title, line, location | `macos/App/Features/Conclusions/EventTitleDisplay.swift` `titleSource(_:locale:orders:)`; `EventsListView.swift` `EventsList.location(_:orders:)`, `EventSecondaryDisplay.content(_:locale:)`, `EventSecondaryLine` | The Events row's title, date · place · "+N", and the Event page location |
| Role labels | `macos/App/Platform/PropertyTermDisplay.swift` `PropertyTermDisplay.name(key:origin:propertyKey:catalogLabel:)`; `SeededPropertyKey.role`; `L10n.PropertyTerm.resource` role cases ("Subject", "Father", "Mother", "Child", "Witness", "Informant", "Pallbearer") | The product name for a seeded role, else the catalog label |
| Person link | `macos/App/Features/Conclusions/PersonsListView.swift` `PersonsList.location(_:orders:)`; `WorkspaceLocation.withPageSection(_:)` | As S13-12 D-1 |
| Thumbnail | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail`; mark `PVMarkKey.subjectEvent` | The Events list's thumb and fallback |
| Wrapping grid | `macos/App/DesignSystem/Components/FlowLayout/PVFlowLayout.swift` `PVFlowLayout` | A wrapping layout for fixed-size items |
| Page columns | `macos/App/DesignSystem/Components/PageColumns/` `PVPageColumns`; `ConclusionDetailPage.swift` `ConclusionDetailLoaded` | Page sections fill the **primary** column only (`primary:`); no section passes a secondary column today |
| Recipes | `macos/App/DesignSystem/Recipes/` (`macos/App/DesignSystem/Recipes/OmnibarHitRow/PVOmnibarHitRow.swift` and others); `macos/App/DesignSystem/README.md` built table | `PV`-prefixed types, one folder each, one README row each (`scripts/check-guidance-names.py`) |
| Tests | `macos/ProvenenciaTests/PersonTileTests.swift`, `PersonFamilyContentTests.swift`, `PersonFamilyLayoutTests.swift`, `PersonDetailTests.swift` (`sectionQueryKeys`), `PlaceRegistryTests.swift` (`personDetailOnFamilyWarmsTheFamily`), `CatalogQueryRegistryTests.swift` | — |

**Precedent.** S13-12 ([#676](https://github.com/mendahu/provenencia/pull/676)) is the same shape: a Person page section that reads its own key through `ConclusionPageSection.queryKey`, a pure content model with folds, a section view with `ConclusionDetailPresentation` states, tiles that navigate with `navigation.go(to:)`. Copy its section, content and fold shapes (`PersonFamilySection`, `PersonFamilyContent`), and its tests' layout. Do not copy its two-column geometry (`PersonFamilyLayout`): the Timeline is a list.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A participant's tile opens `PersonsList.location(participant.person, orders:).withPageSection(.timeline)` with one `navigation.go(to:)` (user, 2026-11-02) | Working rule "A timeline tile opens the participant on Timeline": walking a section stays in it, as Family; one place decides a Person link's history entry and title | Vitals (S13-D1 PPF-5), the user chose Timeline; building `.personDetail(` by hand |
| D-2 | An event opens `EventsList.location(event, orders:)` with one `navigation.go(to:)` (R-28) | The Events list's location: same title and history entry | A hand-built `.eventDetail(` |
| D-3 | `ConclusionPageSection.queryKey` gains `.timeline → .personEvents(project:entityId:)`. Nothing else is wired: `PlaceRegistry` warms it and the switcher's refresh follows it, through the two readers S13-12 added | S13-12 D-8: one declaration per section key; S13-05's descriptor already invalidates and refreshes | A per-section branch in `PlaceRegistry` |
| D-4 | `PersonTile.swift` moves to `macos/App/DesignSystem/Recipes/PersonTile/PVPersonTile.swift` *(new)*; `PersonTile` → `PVPersonTile`, `PersonTileContent` → `PVPersonTileContent`; the button style and body stay private in the file; a row in `macos/App/DesignSystem/README.md`'s built table. No behavior or look changes. Family's three call sites and `PersonTileTests` take the new names | Working rule "One person tile, built once"; `add-ui-component`: the second caller promotes a snowflake to `Recipes/<Name>/` with the `PV` prefix (S13-12 D-5 named this move) | Leaving it in `Features/` with two callers; a `Components/` home (it reads `CatalogPersonSummary`, product data) |
| D-5 | Role labels. `PersonTimelineRole.label(_ value: CatalogConclusionValue) -> String?` *(new, in `PersonTimelineContent.swift`)*: for `.term(_, key, label, origin)` returns `PropertyTermDisplay.name(key: key, origin: origin, propertyKey: SeededPropertyKey.role, catalogLabel: label)`, nil when empty; other cases nil. The person's roles and each participant's roles are their non-nil labels in the read's order. A participant's spoken role (the tile's `role:`) is its labels joined with `L10n.DesignSystem.spokenJoin`; nil when none | The read gives terms with Go's labels (S13-05 D-5); the product names seeded roles the way the Evidence graph does | A Swift role enum (the vocabulary is open); reading `key` alone (a custom role has a label only) |
| D-6 | `PersonTimelineContent` *(new, pure, `Equatable`)* in `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)*: `init(events: [CatalogPersonEvent], expanded: Set<Fold> = [], locale: Locale = .autoupdatingCurrent, orders: NameFormatOrders = NameFormatOrders())`. `dated: [Row]` and `undated: [Row]` keep the read's order, split by `CatalogPersonEvent.dated`; Swift never sorts and never decides "undated" from dates (S13-05 D-8). `Row` holds `id` (event entity id), `title: ConclusionTitleSource` (`EventTitleDisplay.titleSource(event, locale:orders:)`), `line: EventSecondaryDisplay.Content`, `imageRelPath` (`event.pictureImageRelPath`), `roles: [String]` (D-5), `location` (D-2), `participants: [Participant]`, `participantsHidden: Int`, and `spokenLabel` (the title, the line's date and place, then the roles; final wording with the board, TL-12). `Participant` holds `id` ("\(eventId)-\(personId)"), `tile: PVPersonTileContent` (`role:` the spoken role, `isFocus: false`), `roles: [String]`, `location` (D-1). `isEmpty` is `events.isEmpty` | Go orders and flags; Swift places (working rule "Engine PRs run ahead of their slice"); the Events row's own builders, so a timeline row and an Events row cannot disagree | Sorting or grouping in Swift; a second title or date-line builder |
| D-7 | The picture slot is `CachedThumbnail(projectDir:relPath: row.imageRelPath, … fallbackMark: .subjectEvent, fallbackLabel: L10n.Conclusions.eventPictureThumb, colors: nil)`, at the size the board picks (TL-2). R-26 is updated in the planning commit: the slot shows the event's lead picture, else the Event mark (user, 2026-11-02) | Event pictures shipped (S14-02, #641); the summary carries the lead picture; the Events list's fallback | An empty placeholder (R-26's old wording) |
| D-8 | Folds (user, 2026-11-02; working rule "Timeline lists fold"): `PersonTimelineContent.eventFoldLimit = 50`, `participantFoldLimit = 15`; `enum Fold: Hashable { case dated, undated, participants(eventId: String) }`. `datedHidden`, `undatedHidden` and each row's `participantsHidden` count what is folded. `PersonTimelineSection` keeps `@State expanded: Set<PersonTimelineContent.Fold>`, reset `.onChange(of: entityId)`, as Family | The user's caps; Family's fold shape; nothing is hidden for good | A shared fold type with Family (their lists differ; the limit and the reset are the shared parts, and both are one line) |
| D-9 | One "Show N more" row: `PersonFamilyMoreRow` moves out of `PersonFamilySection.swift` to `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` *(new)* as `ConclusionShowMoreRow(count:expand:)`, same body. `L10n.Conclusions.familyShowMore(_:)` is renamed `showMore(_:)`, key `conclusions.showMore`, comment "Button under a folded list on a Conclusion page that shows the rest; argument is how many are hidden"; the String Catalog entry moves with it. Family calls it with no visible change | One row and one string for both sections (brief TL-7, inventory) | A private copy in the Timeline; keeping a Family-named key for a shared row |
| D-10 | `PersonTimelineSection` *(new)* in `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)* reads `@Environment(WorkspaceSession.self)` and `@Environment(WorkspaceNavigation.self)`, gets its handle with `session.queryHandle(key)` and starts it in `.task { let _: QueryHandle<[CatalogPersonEvent]> = session.query(key) }`, as `PersonFamilySection`. States through `ConclusionDetailPresentation<PersonTimelineContent>`: `.loading` → `ProgressView()` centred 96pt from the top; `.failed(message)` → `PVCallout(tone: .danger, title: L10n.Conclusions.timelineErrorTitle, message:)`, id `persons.detail.timeline.error`; `.content` with `isEmpty` → `PVEmptyState(glyph: ConclusionPageSection.timeline.glyph, …)`, id `persons.detail.timeline.empty`; else the list, id `persons.detail.timeline`. `PersonDetailSections` draws it for `.timeline`; `.sources` keeps the `pageSectionLater` empty state | S13-12 D-10, D-11: one read mapping and one wiring shape for every section | Threading the session through `PersonDetailSections` |
| D-11 | Participants lay out in `PVFlowLayout` at the tile's Family width rule, `floor((column − 20 − gutter) / 2)` with Family's gutter (48 Regular, 96 Wide), height `PVPersonTile.height`; the exact spacing, and whether a role caption sits above each tile, come from the board (TL-6) | TL-W2: one tile size on both sections | A new tile size; stretching tiles at Wide |

**Frame-dependent Decisions, filled when the board is agreed** (§ 12): the row form (TL-1), the picture size (TL-2), the Undated caption (TL-3), the role mark and the "Subject" wording (TL-4), the period markers (TL-5), the participant role placement and order (TL-6), the empty copy (TL-10), and the Wide option (TL-W1). If the board picks Wide option B, a Decision adds a secondary slot to `ConclusionDetailPage`'s page sections (today they fill the primary column only) and a Step for it; that is checked against **Size** before the plan is Planned.

**Rules this plan bends.** S13-D1 PPF-5 (archived brief: "Opening a Person from anywhere lands on Vitals") → D-1 opens participant tiles on Timeline → user, 2026-11-02 → the planning commit adds the working rule "A timeline tile opens the participant on Timeline" to the map. R-26 ("a placeholder thumbnail slot") → D-7 shows the lead picture → user, 2026-11-02 → the planning commit updates R-26 and person-detail's "Not this spike" event-depictions line. No other rule is bent.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVPersonTile`, `PVPersonTileContent` *(moved, renamed)* | recipe | `PersonTile` | It is the same tile, promoted (D-4) |
| `ConclusionShowMoreRow` *(moved)*; `L10n.Conclusions.showMore` *(renamed)* | snowflake, key | `PersonFamilyMoreRow`, `familyShowMore` | Moved so both sections share it (D-9) |
| `PersonTimelineRole` *(new)* | enum (namespace) | `PersonFamilyRole`, `PropertyTermDisplay` | Family's roles are a fixed set; the timeline's are open vocabulary terms, labelled through `PropertyTermDisplay` |
| `PersonTimelineContent` *(new)* | pure model | `PersonFamilyContent`, `EventsList` | A different read; reuses `EventTitleDisplay`, `EventSecondaryDisplay`, `EventsList.location`, `PersonsList.location` |
| `PersonTimelineSection` *(new)* | snowflake | `PersonFamilySection`, `ConclusionDetailGallerySection` | A list of events with tile grids |
| `.timeline` case in `ConclusionPageSection.queryKey` | key | `.family` | One more case (D-3) |
| `L10n.Conclusions.timeline*` *(new)* | keys | `L10n.Conclusions.family*` | New copy |

**Builds on.** `CatalogQueryKey.personEvents` and its descriptor (S13-05), `ConclusionPageSection.queryKey` and its two readers (S13-12), `ConclusionDetailPresentation`, `EventTitleDisplay`, `EventSecondaryDisplay` / `EventSecondaryLine`, `EventsList.location`, `PersonsList.location`, `PropertyTermDisplay`, `CachedThumbnail`, `PVCard`, `PVBadge`, `PVFlowLayout`, `PVEmptyState`, `PVCallout`, `pvMicroCaps`, `L10n.DesignSystem.spokenJoin`.
**Deletes.** `macos/App/Features/Conclusions/PersonTile.swift` (moved); `PersonFamilyMoreRow` (moved); `L10n.Conclusions.familyShowMore` and key `conclusions.family.showMore` (renamed); `.timeline` leaves the `pageSectionLater` empty case (the key stays for Sources).
**Next adds.** S13-16 adds `.sources → .personClaims` to `ConclusionPageSection.queryKey` and may reuse `ConclusionShowMoreRow`; an Event page with sections (R-1) can reuse the event row.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A participant tile opens a Person on Timeline | `PersonDetailView` / `ConclusionDetailPage`, keyed by entity id | `viewIdentity: { $0.entityId }` resets the page; `expanded` resets on `entityId` (D-8); a participant merged since the read gives `conclusiondetails.not_found`, the page's Callout |
| An event title opens an Event page | `EventDetailView` | The Events list's path; an event merged since gives the Event page's not-found Callout |
| Back to a Timeline entry | `PersonTimelineSection` | D-3 warms both keys; `.evict` while hidden, so Back reloads (spinner, then list) |
| Only undated events; no events | `PersonTimelineContent` | `dated` empty → the Undated group alone (AC-1.9); `[]` → EmptyState (AC-1.7) |

### Scale

At the brief's sizes (typical 8 – 25 events with 0 – 4 participants; large 120 events, an event with 41 participants): one timeline read per page view, warm with no graph query (S13-05 AC-2.1). Swift builds one row per event and one tile content per participant (linear); the first paint draws at most 50 rows and 15 tiles per row (D-8), each `CachedThumbnail` through `ThumbnailCache`. The key refreshes on any touched handle only while visible (`.evict` when hidden), as S13-05 decided. T-6 builds the content for 120 events and 41 participants.

### Edges

- **Assumptions.** The read's rows are in date order with undated last (S13-05 D-8); D-6 splits by `dated` and keeps order, so an out-of-order `dated` flag would put a dated event in the Undated group: T-4 holds the split. A participant appears once per event (S13-05 D-5 de-duplicates); `Participant.id` includes the event id, since one person can be on several events (T-4).
- **Neighbors.** A role with an empty label and no product name → no role shown for it (D-5); a participant with no roles → the tile with no spoken role; an event with roles empty → no role mark (data the app did not write; T-5).
- **Removed inputs.** An event or participant unpromoted or merged after the read: the next refresh drops them (the key follows every touched handle); a click before that lands on the destination's not-found Callout.
- **Real data.** None — no table or seed drives the section; role labels come from the vocabulary through `PropertyTermDisplay`.
- **Untrusted size.** Events and participants are not bounded in Go; the first paint is bounded by D-8's folds, and the unfolded list is linear.

### Schema

None — no catalog change.

### Go core

None — unless the board asks for another participant order (§ 12 question 6), which adds a sort in `persondetail.PersonEvents` and its test before STEP-1.

### FFI

None — `ListPersonEvents` (S13-05) carries everything the brief draws.

### macOS

The board is not agreed. This section lists the building blocks fixed now; the frames are described in words when the board lands.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Timeline section | snowflake | `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)* | D-10 states; `PersonTimelineContent(events:expanded:locale:orders: session.nameFormat)`; `projectDir` from `session.projectKey.projectDir`. `#Preview`s for the brief's frames 1, 3, 4 and 6 from a `#if DEBUG` preview data enum in the same file, as S13-12's Landed note (5) |
| Content | pure | `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)* | D-5, D-6, D-8 |
| Event row | snowflake (in the section) | `PersonTimelineSection.swift` | `PVCard` or rows on one surface (TL-1, board); the title through `ConclusionListRow.title(_, lineLimit: 2)` as a button to D-2; `EventSecondaryLine` from `row.line`; the role mark (TL-4, board) |
| Picture | `CachedThumbnail` | — | D-7 |
| Participant tiles | `PVPersonTile` in `PVFlowLayout` | — | D-11; id `persons.detail.timeline.tile.<Participant.id>` |
| Show more | `ConclusionShowMoreRow` | — | D-9; id `persons.detail.timeline.more` |
| Undated caption | micro caps or `PVSectionHeader` | — | TL-3 (board); `.accessibilityAddTraits(.isHeader)` |
| Empty / error / loading | `PVEmptyState`, `PVCallout`, `ProgressView` | — | D-10 |
| Refresh | the switcher row | `ConclusionDetailPage.swift` | D-3; no code change |

Accessibility: each row is `.accessibilityElement(children: .contain)` labelled with `row.spokenLabel`; the title is a button; tiles are buttons with their spoken label; decoration hidden (TL-12).

### Strings

L10n keys *(new)* under `L10n.Conclusions` (copy from the brief; final with the board): `timelineEmptyTitle` = "No events yet"; `timelineEmptyMessage` = "Events appear here when events this person took part in are promoted from an Evidence graph."; `timelineErrorTitle` = "Timeline could not be read"; `timelineUndated` = "Undated". The role mark's wording ("as %@" or a bare Badge) and any period-marker copy come with the board (§ 12 questions 3, 4). Renamed: `familyShowMore` → `showMore` (D-9). Each key is added in the step that first uses it (`add-localized-string`), with its entry in `macos/App/Resources/Localizable.xcstrings`.

## 7. Steps

**STEP-1 — The person tile becomes a recipe**
Files: `macos/App/Features/Conclusions/PersonTile.swift` → `macos/App/DesignSystem/Recipes/PersonTile/PVPersonTile.swift` *(new)*, `PersonFamilySection.swift`, `PersonFamilyContent.swift`, `PersonFamilyLayout.swift`, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/PersonTileTests.swift`. Skill: `add-ui-component`.
Do: D-4. No other change.
Check: the app builds; `PersonTileTests`, `PersonFamilyContentTests`, `PersonFamilyLayoutTests` pass unchanged but for names; `scripts/check-guidance-names.py` passes.

**STEP-2 — One "Show N more" row**
Files: `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` *(new)*, `PersonFamilySection.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-9.
Check: Family's fold row looks and reads as before; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — The Timeline section's key**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`, `macos/ProvenenciaTests/PlaceRegistryTests.swift`. Skill: `add-workspace-place`.
Do: D-3.
Check: T-1, T-2 pass.

**STEP-4 — The section's content**
Files: `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)*, `macos/ProvenenciaTests/PersonTimelineContentTests.swift` *(new)*, L10n files. Skill: `add-swift-test`, `add-localized-string`.
Do: D-1, D-2, D-5 – D-8.
Check: T-3 – T-6 pass.

**STEP-5 — The section view**
Files: `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)*, L10n files. Skill: `add-ui-component`.
Do: D-7, D-10, D-11 and the board's frame Decisions; `#Preview`s.
Check: the app builds; the previews render as § 6 describes each frame.

**STEP-6 — Wiring**
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`.
Do: `.timeline` in `PersonDetailSections` draws `PersonTimelineSection(entityId:)`.
Check: `scripts/check-pr.sh` passes; the manual check (§ 10).

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `design/S13-D4-timeline-section.md` → `design/archive/`, `design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys` (extended) | `.timeline.queryKey(project:entityId: "e1")` → `.personEvents(project:entityId: "e1")`; `.family` unchanged; `.vitals`, `.gallery`, `.sources` → nil | D-3 |
| T-2 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `personDetailOnTimelineWarmsTheTimeline` *(new)* | A Person location with `pageSection: .timeline` → `.conclusionDetail` and `.personEvents` | AC-4.2, D-3 |
| T-3 | Swift | `macos/ProvenenciaTests/PersonTimelineContentTests.swift` *(new)* `rowsFromEvents` | Johann's 13 events (brief § 5) → `dated` events 1 – 11 in input order, `undated` 12 and 13; event 4's title, line "4 Jun 1844" · "Bremen", roles ["Subject"], participants Maria, Friedrich with roles; event 13's line empty | AC-1.1, AC-1.3, AC-1.6, D-6 |
| T-4 | Swift | same file `splitKeepsOrderAndIds` | Input order kept (no sort) even when dates would sort otherwise; a `dated: false` row in the middle → in `undated`; one person on two events → two participants with distinct ids | D-6, Edges |
| T-5 | Swift | same file `rolesAndLocations` | Roles `witness`, `informant` → ["Witness", "Informant"]; a custom role with label "Godparent" → "Godparent"; an empty label on a custom key → no role; Anna as `mother` → tile spoken "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851"; a participant's location → `PersonsList.location(…).withPageSection(.timeline)`; a row's location → `EventsList.location(…)`; a row's `imageRelPath` from the summary | AC-1.4, AC-1.5, AC-2.1, AC-2.3, AC-2.4, D-1, D-2, D-5, D-7 |
| T-6 | Swift | same file `foldsAtFiftyAndFifteen` | 120 dated events → 50 rows, `datedHidden` 70; with `.dated` expanded → 120; 41 participants → 15, `participantsHidden` 26; with `.participants(eventId:)` expanded → 41; `[]` → `isEmpty` | AC-1.7, AC-3.1, AC-3.2, D-8, Scale |
| T-7 | Swift | `macos/ProvenenciaTests/PersonTileTests.swift` (renamed types) | Unchanged cases on `PVPersonTileContent` | D-4 |

AC-1.2 is S13-05's (`TestPersonEvents`). AC-1.8, AC-1.9, AC-2.2, AC-2.5, AC-3.3, AC-4.1 are covered by the manual check (§ 10). Frame-dependent tests (the role mark's wording, period markers, the Undated caption rule) are added when the board is agreed. Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `PersonTileTests`, `PersonFamilyContentTests`, `PersonFamilyLayoutTests` (STEP-1), `PlaceRegistryTests` (STEP-3).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `PersonTileContent`, `PersonTile`, `PersonFamilyMoreRow`, `familyShowMore` and `conclusions.family.showMore` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/` and `L10n.swift` with no hit left outside spike-13 plan history; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (done in the planning commit)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Bar 4 and the slice's Check. Use a project with the brief § 5 sample promoted (or the product owner's own test people).

1. (AC-1.1, AC-1.3, AC-1.5, AC-1.6) Open Johann, choose Timeline: events 1 – 11 earliest first, then "Undated" with Residence and Migration; event 4 reads as AC-1.3; event 1's picture and event 11's Event mark.
2. (AC-1.2) Luise's birth is not listed.
3. (AC-1.4) Event 8 shows Witness and Informant.
4. (AC-2.1, AC-2.2, AC-2.4, AC-2.5) Click Anna's tile on event 1: her Timeline opens, crumbs *Persons › Weber, Anna › Timeline*; Back returns. Tab from event 1's title through its tiles; VoiceOver on Anna's tile reads AC-2.4.
5. (AC-2.3) Click event 4's title: the marriage's Event page.
6. (AC-3.1, AC-3.2, AC-3.3) On the obituary's subject, the burial shows 15 tiles and "Show 26 more"; press it; open another person and Back: folded again.
7. (AC-4.1) With Johann's Timeline showing, edit Maria Schulz through her Evidence graph and return with Back: the change shows; "Refreshing" in the switcher row while it reloads.
8. (AC-1.7, AC-1.8, AC-1.9) A person with no events: the EmptyState; a Timeline entry from history for a person merged since: the danger Callout; a person with only undated events.
9. Resize across 1600pt on Timeline: section, scroll, focus and unfolded lists kept (TL-W3); the Wide arrangement the board chose. Dark mode as frame 11.
10. Family looks and behaves as before (tiles, "Show N more").

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The board picks Wide option B, which needs a secondary column for page sections that `ConclusionDetailPage` does not offer | Decided with the board (§ 12 question 7); checked against **Size**; may become a preparatory PR |
| The board puts the role on the tile, changing the tile Family uses | Brief TL-6 keeps the tile's height and content; a tile change is a Decision with the board, and Family is rechecked (manual check 10) |
| 50 rows of tiles render slowly | First paint is bounded by D-8; `CachedThumbnail` loads through the cache |

**Stop and report if:**

- `CatalogPersonEvent`, `CatalogEventParticipant`, `.personEvents` or `ConclusionPageSection.queryKey` differ from § 4.
- The board needs data `ListPersonEvents` does not return (an age, a participant order the read does not give): a Decision and a Go step, not Swift logic.
- Moving the tile changes how Family looks or reads (`PersonFamilyLayoutTests` or `PersonTileTests` expectations change beyond names).
- `PVFlowLayout` cannot place fixed-size tiles at D-11's width.

## 12. Open questions

Each waits on the S13-D4 board; the brief asks it. Recommendation in brackets.

1. **Row form (TL-1).** A Card per event, or rows divided by rules? [Card per event, as the tile is a card.]
2. **Picture size (TL-2).** The Events list row's size or larger? [The Events list's size.]
3. **Role mark and "Subject" (TL-4).** Badge or "as …" line; how "Subject" reads on the person's own events? [A subtle neutral Badge; "Subject" as worded.]
4. **Period markers (TL-5).** Year in the margin, year or decade captions, or a rail? [A year caption where the year changes.]
5. **Participant role placement (TL-6).** Above the tile or on it? [Above, so the tile stays one design.]
6. **Participant order (TL-6).** The read's order, or by role or title? [The read's order; another order adds a Go sort here.]
7. **Wide (TL-W1).** Primary-only, or a secondary column (decade jump list or life summary)? [The board recommends; option B needs a page-section secondary slot, checked against Size.]
8. **Empty copy (TL-10).** [As the brief.]

## 13. PR

**Title:** S13-13 — Timeline section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md`.
