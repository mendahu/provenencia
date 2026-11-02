# S13-13 — Timeline section

**Status:** Blocked — on S13-32, S13-33 and board S13-D4
**Spike / slice:** Spike 13 · Slice 4 — Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D4](../design/S13-D4-timeline-section.md), board not agreed yet
**Board:** — (not drawn yet). When it is agreed, the brief's **Board** line names the Claude Design turn and frames; this plan then describes every frame in words under **Design → macOS**, so an implementer who cannot open the board can still build it
**Depends on:** S13-05 (landed [#680](https://github.com/mendahu/provenencia/pull/680)), S13-12 (landed [#676](https://github.com/mendahu/provenencia/pull/676)), S13-32 (scoped: the tile as a kit shell with person content at the call site; not planned yet), S13-33 (scoped: each event's place on the line and the birth and death markers; not planned yet), S13-D4 (board)
**Blocks:** —
**Branch:** `s13-13-timeline-section`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go or proto change unless the board asks for another participant order (§ 12, question 6)
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-05 and S13-12 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 3. Timeline](../person-detail.md#3-timeline) (R-2, R-3, R-23 – R-28, R-24a); brief [S13-D4](../design/S13-D4-timeline-section.md); the S13-05 plan [`S13-05-person-timeline-read.md`](S13-05-person-timeline-read.md) and the S13-12 plan [`S13-12-family-section.md`](S13-12-family-section.md), with their **Landed** notes; working rules: "Engine PRs run ahead of their slice" (data, not layout), "Each section reads its own query", "Timeline is direct participation, any role", "One person tile, built once; the shell in the kit, the content at the call site", "Colors reuse `PersonGenderTreatment`", "A timeline tile opens the participant on Timeline", "Timeline lists fold", "The timeline is a lifeline", "Sections are places"
**Planned:** 2026-11-02 against `main` @ `51dccd4` (draft). Refreshed when S13-32 and S13-33 land (the tile's names and slots replace § 4's Person tile row and D-4; the position and marker fields replace D-15's provisional names) and when the board is agreed (§ 12)
**Review:** [`S13-13-timeline-section.review.md`](S13-13-timeline-section.review.md) — round 1 Fits with changes / Ready with changes; M-1 answered by the user (a preparatory PR, S13-32), every other finding Fixed or Declined; round 2 runs after the board

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

A researcher opens a Person's Timeline section and reads the person's life on a lifeline drawn to scale from birth to death, where gaps in the record show as empty line, beside a list of events, earliest first, each pointing to its place: each event's picture, title, date and place, the person's role in it, and the other people who took part as the same person tiles as on Family. Birth and death always have a marker, a placeholder when unrecorded; undated events sit between them. A click on an event opens its Event page, and a click on a participant opens that participant's Timeline. It fills the empty place S13-10 left (S13-D1 PPF-10) with the data S13-05 and S13-33 read, reuses the person tile as S13-32 leaves it (a kit shell filled with person content at the call site), and gives Family and the Timeline one host for a section's own read.

**Why now.** Slice 4's Check is bar 4 ("A life in order"); S13-05 has the read, S13-12 has the tile, S13-32 decouples it, and this is the last PR of the slice.

**What stays.** The header, the switcher, Vitals, Family (its look; its host code moves, D-12), Gallery and the Sources empty place; the person tile; `ListPersonEvents`, its order, its roles and its filtering (S13-05); the Events list and Event page; every other link into a Person page.

## 2. User stories

**US-1.** As a researcher, when I open a Person's Timeline, I see every event they took part in, in date order, so that I read their life from birth to death.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann (brief § 5 sample, events 1 – 13) | Events 1 – 11 in that order; 12 (Residence at Altona) and 13 (Migration) under "Undated" between the birth and death markers, in the order the read returns them, where the board places them (R-23, R-24; D-6, D-16) |
| AC-1.2 | Luise's birth names no father | It is not listed (R-23; S13-05 filters it) |
| AC-1.3 | Event 4 | Title "Marriage of Johann Weber and Maria Schulz" (`EventTitleDisplay`), "4 Jun 1844 · Bremen" (`EventSecondaryDisplay`), Johann's role "Subject", tiles Maria Schulz and Friedrich Weber (R-25, R-27; D-5, D-6) |
| AC-1.4 | Event 8: Johann is Witness and Informant | Both roles shown, in the read's order, joined as the board words it (D-5; § 12 question 3) |
| AC-1.5 | Event 1 has a lead picture; event 11 has none | Event 1 shows its picture; event 11 shows the Event mark on the sunken tone (R-26 as updated 2026-11-02; D-7) |
| AC-1.6 | Event 13: no date, no place | The date and place line is not drawn |
| AC-1.7 | A person with no promoted events (`[]`) | The EmptyState of TL-10 |
| AC-1.8 | The read fails with `conclusiondetails.not_found` | A danger Callout in the section titled "Timeline could not be read" with the mapped message; the header and switcher still show (TL-11) |
| AC-1.9 | Only undated events | Both placeholders ("Birth unknown", "Death unknown") with the undated events between them; no point on the line (D-15, D-16) |

**US-2.** As a researcher, I can move from an event to its page and from a participant to their life.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | On Johann's Timeline, click Anna Weber's tile on event 1 | Anna's Person page opens on Timeline; breadcrumbs *Persons › Weber, Anna › Timeline* (`PersonsList.location` gives the sorted title, D-1; the brief's TL-8 uses the same form) |
| AC-2.2 | Then Back | Johann's Timeline again |
| AC-2.3 | Click event 4's title | The Event page of that marriage opens, titled as the Events list titles it (D-2) |
| AC-2.4 | VoiceOver on Anna's tile under event 1 | "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851", a button (the tile's spoken label with the role term's label, D-5) |
| AC-2.5 | Tab from event 1's title | Focus moves to its tiles in order, then to event 2's title (TL-12) |

**US-3.** As a researcher, long lives and crowded events stay readable without hiding anything.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | An event with 41 other participants | 15 tiles, then "Show 26 more"; pressing it shows all 41 (D-8). Evidence: T-6, and the `#Preview` "Large event" (D-13) |
| AC-3.2 | A person with 120 dated events | 50 events, then "Show 70 more"; pressing it shows all 120; the Undated group folds the same way on its own (D-8). Evidence: T-6, and the `#Preview` "Long life" (D-13) |
| AC-3.3 | Unfold a list, then open another person and come Back | The list is folded again (D-8, as Family) |

**US-5.** As a researcher, I see the shape of the life at a glance: where its records cluster and where they fall silent (R-24a; user, 2026-11-02).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | Johann: birth 12 Mar 1817, death 1880 | The line runs from the birth marker to the death marker; each of events 1 – 11 points to its place; the stretch from 1861 to 1880 has no point (D-15) |
| AC-5.2 | Event 4 on 4 Jun 1844 and event 5 on 2 Feb 1845 | Their points are as far apart on the line as eight months are at the line's scale; the 1850 – 1859 gap is about twelve times longer (D-15) |
| AC-5.3 | A range "1844 – 1849"; "bef. 1880" | A span on the line; a point with an open start (D-15) |
| AC-5.4 | Johann without event 1 (no birth recorded) | A "Birth unknown" placeholder just before the baptism, joined by a dashed stretch not to scale; no year shown (D-16) |
| AC-5.5 | A funeral in 1880, a census in 1875, no death | The "Death unknown" placeholder sits after the census and before the funeral; the funeral sits at its own date (D-16) |
| AC-5.6 | The burial, dated after death | It sits at its own date, past the death marker, and the line runs on to it (D-15) |

**US-4.** As a researcher, the section stays current while I work.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Johann's Timeline is showing; a write touches Maria Schulz's handle | The timeline reloads; the list stays drawn and the switcher row shows "Refreshing" until it lands (D-3; the descriptor S13-05 added) |
| AC-4.2 | Opening a Person straight on Timeline (from a tile, or Back) | The timeline read starts with the page, not after the detail read (D-3) |

## 3. Scope

| In | Covered by |
| --- | --- |
| One "Show N more" row for Family and Timeline (TL-7, D-9) | STEP-1 |
| The lifeline: scale, event points, birth and death markers and placeholders (TL-5, TL-5a; D-15, D-16) | STEP-5, AC-5.1 – AC-5.6 |
| One host for a page section's own read; Family moves onto it, unchanged (D-12) | STEP-2 |
| The Timeline section's query key: warm, read, refresh (D-3) | STEP-3, AC-4.1, AC-4.2 |
| The section's content model: rows, roles, participants, locations, folds, accessibility (TL-1 – TL-9, TL-12; D-1, D-2, D-5 – D-8) | STEP-4, AC-1.1 – AC-1.6, AC-2.1, AC-2.3, AC-2.4, AC-3.1 – AC-3.3 |
| The section view and its states (TL-1 – TL-13, TL-W1 – TL-W3; D-10, D-11) | STEP-5, STEP-6, AC-1.7 – AC-1.9 |
| R-26 now shows the event's lead picture (user, 2026-11-02) | planning commit; STEP-4 (D-7) |

| Out | Home |
| --- | --- |
| Zoom, pan, a draggable canvas or a swimlane | not planned (brief § 2.1) |
| Gap labels; an estimated year for an unknown birth or death; marks on events outside the life | not planned (user, 2026-11-02) |
| Positions and markers in Go | S13-33 |
| Age at each event | not planned this spike (brief § 2.1) |
| Filtering or searching the timeline | not planned |
| Events reached only through a relative or a place | not planned (R-23) |
| The provenance control on a row, role or tile | not planned this spike (R-13b) |
| Editing events or participations from the section | not planned this spike |
| Splitting the tile into a kit shell and person content; `PropertyTermDisplay.name(_:propertyKey:)`; the kit's Return modifier; renaming the tile's "this person" keys | S13-32 |
| `ReconciledValueDisplay.string(for:)`'s `.term` case (label else key, no product names), an older copy of the term-name rule | logged in [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) by this PR's STEP-7 |
| A participant order other than the read's (by role or title) | a Go step added to this plan when the board is agreed, only if the board asks (§ 12 question 6); otherwise not planned |
| `EventSubject` folding into {role, `PersonSummary`} | logged by S13-05 in [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Timeline section | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` | `.timeline, .sources` draw one `PVEmptyState` with `L10n.Conclusions.pageSectionLater`. `.family` draws `PersonFamilySection(entityId:)` |
| Section key | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection.queryKey(project:entityId:)` | `.family → .personFamily`; every other section nil. Read by `PlaceRegistry` Spec `.personDetail` (warms it) and `ConclusionDetailLoaded.sectionRefreshing` (`ConclusionDetailPage.swift`) |
| Timeline key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.personEvents(project:entityId:)`; `CatalogQueryRegistry.swift` its descriptor | `.sessionFresh`, `.evict` when hidden, `.allCached(.personEvents)`, refreshes on any touched handle; loads `store.listPersonEvents`. No view queries it |
| Timeline data | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonEvent` (`event: CatalogEventSummary`, `roles: [CatalogConclusionValue]`, `participants: [CatalogEventParticipant]`, `dated`), `CatalogEventParticipant` (`person: CatalogPersonSummary`, `roles`) | Rows in Go's date order, undated last and flagged; roles are `.term(id:key:label:origin:)` with labels filled in Go; participants in participation-ref order, never the person. S13-33 (not landed) adds each dated event's place (first and last day, open ends) and the response's birth and death markers (a recorded event, or a placeholder with the event it sits before or after); refresh this row then |
| Person tile | Today: `macos/App/Features/Conclusions/PersonTile.swift` `PersonTile`, `PersonTileContent`. After S13-32 (not landed; refresh this row then): a tile shell in `macos/App/DesignSystem/` with slots and a declared minimum width, and the person content built in `Features/Conclusions` from a `CatalogPersonSummary` and an optional spoken role | `PersonTileContent.init(summary:role:isFocus:locale:orders:)` speaks the role and never draws it; height 92. Callers: `PersonFamilySection.swift`, `PersonFamilyContent.swift`, `PersonFamilyLayout.swift`; tests `macos/ProvenenciaTests/PersonTileTests.swift` |
| Show more | `macos/App/Features/Conclusions/PersonFamilySection.swift` private `PersonFamilyMoreRow`; `macos/App/Platform/L10n.swift` `L10n.Conclusions.familyShowMore(_:)` (key `conclusions.family.showMore`, "Show %lld more") | A ghost `PVButton` (`.pv(.ghost, size: .sm)`) with Return handling; Family only |
| Folds | `macos/App/Features/Conclusions/PersonFamilyContent.swift` `foldLimit` (30), `Fold`, `expanded`; `PersonFamilySection` keeps `@State expanded`, reset `.onChange(of: entityId)` | The pattern to copy for the Timeline's folds |
| Read states | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPresentation` | Generic `init<Value>(_ read: QueryRead<Value>, isFetching:content:)` (S13-12 D-10) |
| Event title, line, location, spoken | `macos/App/Features/Conclusions/EventTitleDisplay.swift` `titleSource(_:locale:orders:)`; `EventsListView.swift` `EventsList.location(_:orders:)`, `EventsList.accessibilityLabel(_:orders:)` (title, date, ref through `L10n.Workspace.eventRowAccessibility`, then place and `L10n.Conclusions.morePlaces`), `EventSecondaryDisplay.content(_:locale:)`, `EventSecondaryLine` | The Events row's title, spoken label, date · place · "+N", and the Event page location |
| Role labels | `macos/App/Platform/PropertyTermDisplay.swift` `PropertyTermDisplay.name(key:origin:propertyKey:catalogLabel:)`, `name(term:propertyKey:)`; `SeededPropertyKey.role`; `L10n.PropertyTerm.resource` role cases ("Subject", "Father", "Mother", "Child", "Witness", "Informant", "Pallbearer") | The product name for a seeded role, else the trimmed catalog label, else the key. S13-32 adds `name(_ value: CatalogConclusionValue, propertyKey:) -> String?` beside them |
| Person link | `macos/App/Features/Conclusions/PersonsListView.swift` `PersonsList.location(_:orders:)`; `WorkspaceLocation.withPageSection(_:)` | As S13-12 D-1 |
| Thumbnail | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail`; mark `PVMarkKey.subjectEvent` | The Events list's thumb and fallback |
| Tile grid | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailGallerySection` | A `LazyVGrid` with `.adaptive(minimum:maximum:)` columns of fixed-size likeness tiles: the precedent for the participant grid |
| Section host | `macos/App/Features/Conclusions/PersonFamilySection.swift` `PersonFamilySection`, private `PersonFamilyLoading`, `PersonFamilyLoaded` | The handle lookup, the `.task` that starts the query, the 96pt spinner, the danger Callout, the EmptyState and the ids, around Family's own diagram |
| Preview data | `macos/App/Features/Conclusions/PersonFamilyPreviewData.swift` `PersonFamilyPreviewData` (`#if DEBUG`), `summary(_:_:born:)` | Shared by Family's `#Preview`s and `PersonFamilyContentTests` / `PersonFamilyLayoutTests` (`@testable`) |
| Page columns | `macos/App/DesignSystem/Components/PageColumns/` `PVPageColumns`; `ConclusionDetailPage.swift` `ConclusionDetailLoaded` | Page sections fill the **primary** column only (`primary:`); no section passes a secondary column today |
| Tests | `macos/ProvenenciaTests/PersonTileTests.swift`, `PersonFamilyContentTests.swift`, `PersonFamilyLayoutTests.swift`, `PersonDetailTests.swift` (`sectionQueryKeys`), `PlaceRegistryTests.swift` (`personDetailOnFamilyWarmsTheFamily`), `CatalogQueryRegistryTests.swift` | — |

**Precedent.** S13-12 ([#676](https://github.com/mendahu/provenencia/pull/676)) is the same shape: a Person page section that reads its own key through `ConclusionPageSection.queryKey`, a pure content model with folds, a section view with `ConclusionDetailPresentation` states, tiles that navigate with `navigation.go(to:)`. Copy its content and fold shapes (`PersonFamilyContent`), its shared preview data (`PersonFamilyPreviewData`), and its tests' layout; its host moves into D-12's shared host rather than being copied. Do not copy its two-column geometry (`PersonFamilyLayout`): the Timeline is a list.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A participant's tile opens `PersonsList.location(participant.person, orders:).withPageSection(.timeline)` with one `navigation.go(to:)` (user, 2026-11-02) | Working rule "A timeline tile opens the participant on Timeline": walking a section stays in it, as Family; one place decides a Person link's history entry and title | Vitals (S13-D1 PPF-5), the user chose Timeline; building `.personDetail(` by hand |
| D-2 | An event opens `EventsList.location(event, orders:)` with one `navigation.go(to:)` (R-28) | The Events list's location: same title and history entry | A hand-built `.eventDetail(` |
| D-3 | `ConclusionPageSection.queryKey` gains `.timeline → .personEvents(project:entityId:)`. Nothing else is wired: `PlaceRegistry` warms it and the switcher's refresh follows it, through the two readers S13-12 added | S13-12 D-8: one declaration per section key; S13-05's descriptor already invalidates and refreshes | A per-section branch in `PlaceRegistry` |
| D-4 | The participant tile is the tile S13-32 lands: its kit shell filled by the person content builder in `Features/Conclusions` with `summary: participant.person`, `role:` the spoken role (D-5), not the focus. The Timeline adds no tile code and no tile variant. If S13-32 lands with other names, this plan is refreshed before STEP-1 | Working rule "One person tile, built once; the shell in the kit, the content at the call site" (user, 2026-11-02) | Moving today's `PersonTile` into `DesignSystem/Recipes/` (it reads six Conclusions display types: review M-1); a Timeline copy of the tile |
| D-5 | Role labels through one shared resolver: each role is `PropertyTermDisplay.name(value, propertyKey: SeededPropertyKey.role)` (S13-32), nil for a value that is not a term; a term with an empty label shows its key, as every term name does. The person's roles and each participant's roles are the non-nil names in the read's order. A participant's spoken role (the tile's `role:`) is its names joined with `L10n.DesignSystem.spokenJoin`; nil when none. No timeline-only role type | One rule turns a term into a name (review M-2); the read gives terms with Go's labels (S13-05 D-5) | `PersonTimelineRole` with its own empty-label rule; a Swift role enum (the vocabulary is open) |
| D-6 | `PersonTimelineContent` *(new, pure, `Equatable`)* in `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)*: `init(events: [CatalogPersonEvent], expanded: Set<Fold> = [], locale: Locale = .autoupdatingCurrent, orders: NameFormatOrders = NameFormatOrders())`. `dated: [Row]` and `undated: [Row]` keep the read's order, split by `CatalogPersonEvent.dated`; the view places `undated` between the markers (D-16); Swift never sorts and never decides "undated" from dates (S13-05 D-8). `Row` holds `id` (event entity id), `title: ConclusionTitleSource` (`EventTitleDisplay.titleSource(event, locale:orders:)`), `line: EventSecondaryDisplay.Content`, `imageRelPath` (`event.pictureImageRelPath`), `roles: [String]` (D-5), `location` (D-2), `participants: [Participant]`, `participantsHidden: Int`, and `spokenLabel` = `spokenJoin(EventsList.accessibilityLabel(event, orders:, includesRef: false), rolesSpoken)`, where `rolesSpoken` is the roles joined with `spokenJoin` and its wording ("as Subject" or the bare role) comes with the board (TL-4, TL-12). `EventsList.accessibilityLabel` gains `includesRef: Bool = true` *(new parameter)*: with false it uses the title and date only (`L10n.Conclusions.lineJoin` when both, else the one present), then the place and "+N places" as today; every existing caller is unchanged. `Participant` holds `id` ("\(eventId)-\(personId)"), `tile` (the S13-32 person content), `roles: [String]`, `location` (D-1). `isEmpty` is `events.isEmpty` | Go orders and flags; Swift places (working rule "Engine PRs run ahead of their slice"); the Events row's own builders, spoken label included, so a timeline row and an Events row cannot disagree (review M-3) | Sorting or grouping in Swift; a second title, date-line or spoken-label builder |
| D-7 | The picture slot is `CachedThumbnail(projectDir:relPath: row.imageRelPath, … fallbackMark: .subjectEvent, fallbackLabel: L10n.Conclusions.eventPictureThumb, colors: nil)`, at the size the board picks (TL-2). R-26 is updated in the planning commit: the slot shows the event's lead picture, else the Event mark (user, 2026-11-02) | Event pictures shipped (S14-02, #641); the summary carries the lead picture; the Events list's fallback | An empty placeholder (R-26's old wording) |
| D-8 | Folds (user, 2026-11-02; working rule "Timeline lists fold"): `PersonTimelineContent.eventFoldLimit = 50`, `participantFoldLimit = 15`; `enum Fold: Hashable { case dated, undated, participants(eventId: String) }`. `datedHidden`, `undatedHidden` and each row's `participantsHidden` count what is folded. The section keeps `expanded` as `@State`, reset on a new entity id, through D-12's host | The user's caps; Family's fold shape; nothing is hidden for good | A shared fold type with Family (their lists differ) |
| D-9 | One "Show N more" row: `PersonFamilyMoreRow` moves out of `PersonFamilySection.swift` to `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` *(new)* as `ConclusionShowMoreRow(count:expand:)`; its Return handling is the kit's Return modifier S13-32 adds. `L10n.Conclusions.familyShowMore(_:)` is renamed `showMore(_:)`, key `conclusions.showMore`, comment "Button under a folded list on a Conclusion page that shows the rest; argument is how many are hidden"; the String Catalog entry moves with it. Family calls it with no visible change | One row and one string for both sections (brief TL-7) | A private copy in the Timeline; keeping a Family-named key for a shared row |
| D-10 | The event title is a `Button` to D-2 whose label is `ConclusionListRow.title(row.title, lineLimit: 2)`, styled as the kit's link button style with the kit focus ring (the board confirms the style, TL-1), with the kit's Return modifier (S13-32). No hand-written `.onKeyPress(.return)` in the section | One owner for Return activation (review M-4) | A third hand copy of `focused` + `onKeyPress` |
| D-11 | Participants lay out in a `LazyVGrid(columns: [GridItem(.adaptive(minimum: tileMinimumWidth), spacing: 12)], spacing: 12)` over the row's width, where `tileMinimumWidth` is the minimum width S13-32 declares on the tile shell beside its height; tiles keep the shell's height. Family keeps its own layout. Spacing and the role caption's placement are confirmed by the board (TL-6) | The tile's size is stated once, on the tile (review M-5, executability X-2); the Gallery's adaptive grid is the precedent; Wide adds columns at the same minimum (TL-W2) | Family's `PersonFamilyLayout` column formula (a rule of the Family diagram, and it narrows tiles at Wide); `PVFlowLayout` (it does not share spare width across a row) |
| D-12 | One host for a page section's own read: `ConclusionPageSectionHost<Value, Content: Equatable, Body: View>` *(new)* in `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` *(new)* takes the `CatalogQueryKey`, the identifier root (`persons.detail.family`, `persons.detail.timeline`), the glyph, the error title, the empty title and message, a pure `make: (Value) -> Content`, `isEmpty: (Content) -> Bool`, and a `@ViewBuilder` body over the content. It reads `@Environment(WorkspaceSession.self)`, gets the handle with `session.queryHandle(key)`, starts it in `.task { let _: QueryHandle<Value> = session.query(key) }`, and maps `ConclusionDetailPresentation<Content>`: `.loading` (or no handle) → `ProgressView()` centred 96pt from the top; `.failed(message)` → `PVCallout(tone: .danger, title:, message:)`, id `<root>.error`; `.content` with `isEmpty` → `PVEmptyState(glyph:title:message:)`, id `<root>.empty`; else the body, id `<root>`. `PersonFamilySection` moves onto it with its ids, copy and look unchanged; `PersonTimelineSection` *(new)* in `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)* is its second caller, with `L10n.Conclusions.timeline*` copy and the glyph `ConclusionPageSection.timeline.glyph`. Each section keeps its own `@State expanded`, reset `.onChange(of: entityId)` | Only the content differs between sections; S13-16 is the named third caller (review M-6) | Copying `PersonFamilySection`'s host (three private copies by S13-16) |
| D-13 | Fixtures: `macos/App/Features/Conclusions/PersonTimelinePreviewData.swift` *(new)* (`#if DEBUG`, internal) `enum PersonTimelinePreviewData` builds `[CatalogPersonEvent]` for the brief's frames: `johann` (the 13 events of brief § 5, with events 1, 2, 4, 5, 9 and 10 holding a `pictureImageRelPath`), `largeEvent` (one burial with 41 participants), `longLife` (120 dated events and 2 undated), `onlyUndated` (2 undated). People come from `PersonFamilyPreviewData.summary`; events from one `event(_ id:, title rule:, date:, place:, picture:)` helper in the file. `PersonTimelineSection`'s `#Preview`s and `PersonTimelineContentTests` (`@testable`) both use it | S13-12's Landed note (6): one fixture shared by previews and tests (review M-7, X-5) | A private preview enum plus a test-only copy |
| D-14 | Accessibility ids under `persons.detail.timeline`: the list `persons.detail.timeline`; a row `.row.<eventId>`; its title `.title.<eventId>`; a tile `.tile.<Participant.id>`; the Undated caption `.undated`; fold rows `.more.dated`, `.more.undated`, `.more.<eventId>`; plus D-12's `.error` and `.empty` | Unique ids for UI tests and the manual check (X-7) | One shared `.more` id |

| D-15 | The lifeline's geometry is a pure function, `PersonTimelineLifeline` *(new)* in `macos/App/Features/Conclusions/PersonTimelineLifeline.swift` *(new)*: `init(content: PersonTimelineContent, length: CGFloat)` maps S13-33's day positions linearly onto `length` between the first and last placed thing (the birth marker or the earliest point, the death marker or the latest point, a burial past death included), and gives each dated row's point (`at: CGFloat`, `spanTo: CGFloat?`, `openStart`, `openEnd`), the markers' positions, and the year ticks. Swift does no date arithmetic: it reads S13-33's day numbers only. The orientation, the scale rule (fit or points per year) and how a row points to its place are the board's (TL-5) and become this function's parameters | Working rule "The timeline is a lifeline": positions are data, Swift only scales them; a pure function, tested, as `PersonFamilyLayout` | Computing positions from `CatalogDateValueInput` in Swift (a second copy of `valuecompare.DateEnds`); `GraphCanvas` |
| D-16 | Markers: the birth and death markers come from S13-33. A recorded marker is drawn at its event's point; a placeholder is drawn a fixed distance (the board's) before or after the event S13-33 names, joined by a dashed stretch that `PersonTimelineLifeline` leaves out of the scale; with no bounding event, the placeholders sit at the two ends of the line. The undated rows sit between the markers in the list, where the board places them (TL-3). No marker or event is ever hidden or flagged | User, 2026-11-02: every timeline has a spot for birth and death; no invented dates; events sit at their dates | An estimated year (matching's window pads 110 years); a warning on an event outside the life |

**Frame-dependent Decisions, filled when the board is agreed** (§ 12): the row form (TL-1), the picture size (TL-2), where the undated rows sit between the markers (TL-3), the lifeline's orientation, scale rule and row-to-point link (TL-5), the placeholder's distance and look (TL-5a), the role mark and the "Subject" wording (TL-4), the role caption's spacing and the participant order (TL-6), the empty copy (TL-10), the roles' spoken wording (TL-12), and the Wide option (TL-W1). If the board picks Wide option B, a Decision adds a secondary slot to `ConclusionDetailPage`'s page sections (today they fill the primary column only) and a Step for it; that is checked against **Size** before the plan is Planned.

**Rules this plan bends.** S13-D1 PPF-5 (archived brief: "Opening a Person from anywhere lands on Vitals") → D-1 opens participant tiles on Timeline → user, 2026-11-02 → the planning commit adds the working rule "A timeline tile opens the participant on Timeline" to the map. R-26 ("a placeholder thumbnail slot") → D-7 shows the lead picture → user, 2026-11-02 → the planning commit updates R-26 and person-detail's "Not this spike" event-depictions line. The working rule "One person tile, built once" said the tile becomes a recipe with its second caller → that would put Conclusions display types in the design system (review M-1, against `docs/design-system-layers.md`) → user, 2026-11-02: the design system holds content-agnostic shells and treatments, value resolvers live at the repository level, content at the call site; a preparatory PR (S13-32) → the planning commit rewrites the rule and adds S13-32 to the map. No other rule is bent.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionShowMoreRow` *(moved)*; `L10n.Conclusions.showMore` *(renamed)* | snowflake, key | `PersonFamilyMoreRow`, `familyShowMore` | Moved so both sections share it (D-9) |
| `ConclusionPageSectionHost` *(new)* | snowflake (generic) | `PersonFamilySection`'s host, `ConclusionDetailPresentation` | Extracts Family's host so the Timeline and S13-16 add content only (D-12) |
| `EventsList.accessibilityLabel(_:orders:includesRef:)` *(new parameter)* | func | `EventsList.accessibilityLabel(_:orders:)` | Extended, not copied (D-6) |
| `PersonTimelineContent` *(new)* | pure model | `PersonFamilyContent`, `EventsList` | A different read; reuses `EventTitleDisplay`, `EventSecondaryDisplay`, `EventsList.location`, `EventsList.accessibilityLabel`, `PersonsList.location`, `PropertyTermDisplay` |
| `PersonTimelineSection` *(new)* | snowflake | `PersonFamilySection`, `ConclusionDetailGallerySection` | A list of events with tile grids, on D-12's host |
| `PersonTimelineLifeline` *(new)* | pure layout | `PersonFamilyLayout`, `PVPageColumns` | A scale from day numbers to points; neither knows time (D-15) |
| `PersonTimelinePreviewData` *(new)* | fixture | `PersonFamilyPreviewData` | Timeline events; its people come from `PersonFamilyPreviewData.summary` (D-13) |
| `.timeline` case in `ConclusionPageSection.queryKey` | key | `.family` | One more case (D-3) |
| `L10n.Conclusions.timeline*` *(new)* | keys | `L10n.Conclusions.family*` | New copy |

`Participant` (D-6) and Family's `PersonFamilyContent.Tile` both pair a tile with a location; the Timeline adds `roles`. They stay two types: S13-32 decides the person content type both hold, and a shared "tile with a destination" is reconsidered then (review M-9).

**Builds on.** `CatalogQueryKey.personEvents` and its descriptor (S13-05), `ConclusionPageSection.queryKey` and its two readers (S13-12), `ConclusionDetailPresentation`, the S13-32 tile shell, person content, Return modifier and `PropertyTermDisplay.name(_:propertyKey:)`, `EventTitleDisplay`, `EventSecondaryDisplay` / `EventSecondaryLine`, `EventsList.location`, `EventsList.accessibilityLabel`, `PersonsList.location`, `CachedThumbnail`, `PVCard`, `PVBadge`, `PVEmptyState`, `PVCallout`, `pvMicroCaps`, `L10n.DesignSystem.spokenJoin`.
**Deletes.** `PersonFamilyMoreRow` (moved); `L10n.Conclusions.familyShowMore` and key `conclusions.family.showMore` (renamed); `PersonFamilyLoading`, `PersonFamilyLoaded`'s state mapping and `PersonFamilySection`'s handle and `.task` code (moved into D-12's host); `.timeline` leaves the `pageSectionLater` empty case (the key stays for Sources).
**Next adds.** S13-16 adds `.sources → .personClaims` to `ConclusionPageSection.queryKey` and is the host's third caller (copy, root id, content); it may reuse `ConclusionShowMoreRow`. An Event page with sections (R-1) can reuse the event row.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A participant tile opens a Person on Timeline | `PersonDetailView` / `ConclusionDetailPage`, keyed by entity id | `viewIdentity: { $0.entityId }` resets the page; `expanded` resets on `entityId` (D-8); a participant merged since the read gives `conclusiondetails.not_found`, the page's Callout |
| An event title opens an Event page | `EventDetailView` | The Events list's path; an event merged since gives the Event page's not-found Callout |
| Back to a Timeline entry | `PersonTimelineSection` | D-3 warms both keys; `.evict` while hidden, so Back reloads (spinner, then list) |
| Only undated events; no events | `PersonTimelineContent`, `PersonTimelineLifeline` | `dated` empty → both placeholders at the line's ends with the undated rows between (AC-1.9); `[]` → EmptyState (AC-1.7) |
| One dated event only; birth and death on one day | `PersonTimelineLifeline` | A zero-length scale: the line has a minimum length and the point sits mid-line (T-9) |

### Scale

At the brief's sizes (typical 8 – 25 events with 0 – 4 participants; large 120 events, an event with 41 participants): one timeline read per page view, warm with no graph query (S13-05 AC-2.1). Swift builds one row per event and one tile content per participant (linear); the first paint draws at most 50 rows and 15 tiles per row (D-8), each `CachedThumbnail` through `ThumbnailCache`. The key refreshes on any touched handle only while visible (`.evict` when hidden), as S13-05 decided. T-6 builds the content for 120 events and 41 participants.

### Edges

- **Assumptions.** The read's rows are in date order with undated last (S13-05 D-8); S13-33's positions are in the same order for dated rows (a span's first day), so points never cross rows' order except where spans overlap (T-8); D-6 splits by `dated` and keeps order, so an out-of-order `dated` flag would put a dated event in the Undated group: T-4 holds the split. A participant appears once per event (S13-05 D-5 de-duplicates); `Participant.id` includes the event id, since one person can be on several events (T-4).
- **Neighbors.** A role with an empty label and no product name → its key, as every term name (D-5); a value that is not a term → no role; a participant with no roles → the tile with no spoken role; an event with roles empty → no role mark (data the app did not write; T-5).
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
| Section host | snowflake | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` *(new)* | D-12 |
| Timeline section | snowflake | `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)* | On the host; `PersonTimelineContent(events:expanded:locale:orders: session.nameFormat)`; `projectDir` from `session.projectKey.projectDir`. `#Preview`s "Full timeline", "Row sheet", "Large event", "Long life", "Only undated" from `PersonTimelinePreviewData` (D-13) |
| Content | pure | `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)* | D-5, D-6, D-8 |
| Event row | snowflake (in the section) | `PersonTimelineSection.swift` | `PVCard` or rows on one surface (TL-1, board); the title (D-10); `EventSecondaryLine` from `row.line`; the role mark (TL-4, board) |
| Picture | `CachedThumbnail` | — | D-7 |
| Lifeline | `Path` in kit tokens | `macos/App/Features/Conclusions/PersonTimelineLifeline.swift` *(new)* (geometry) and its drawing in `PersonTimelineSection.swift` | D-15, D-16; line, ticks and connectors `.accessibilityHidden(true)`; markers are list items VoiceOver reads (TL-12) |
| Participant tiles | S13-32's tile shell with person content, in an adaptive `LazyVGrid` | — | D-4, D-11 |
| Show more | `ConclusionShowMoreRow` | — | D-9 |
| Undated caption | micro caps or `PVSectionHeader` | — | TL-3 (board); `.accessibilityAddTraits(.isHeader)` |
| Refresh | the switcher row | `ConclusionDetailPage.swift` | D-3; no code change |

Ids: D-14. Accessibility: each row is `.accessibilityElement(children: .contain)` labelled with `row.spokenLabel`; the title is a button; tiles are buttons with their spoken label; decoration hidden (TL-12).

### Strings

L10n keys *(new)* under `L10n.Conclusions` (copy from the brief; final with the board): `timelineEmptyTitle` = "No events yet"; `timelineEmptyMessage` = "Events appear here when events this person took part in are promoted from an Evidence graph."; `timelineErrorTitle` = "Timeline could not be read"; `timelineUndated` = "Undated"; `timelineBirthUnknown` = "Birth unknown"; `timelineDeathUnknown` = "Death unknown". The role mark's wording ("as %@" or a bare Badge) and any period-marker copy come with the board (§ 12 questions 3, 4). Renamed: `familyShowMore` → `showMore` (D-9). Each key is added in the step that first uses it (`add-localized-string`), with its entry in `macos/App/Resources/Localizable.xcstrings`.

## 7. Steps

**STEP-0 — Refresh against S13-32**
Files: this plan.
Do: when S13-32 and S13-33 have landed, read their **Landed** notes; replace D-15 and D-16's provisional field names with S13-33's; replace § 4's Person tile row, D-4, D-5's resolver name and D-11's minimum-width name with the landed names. When the board is agreed, fill § 12's Decisions and the frames in words, and run `review-plan` round 2.
Check: § 12 Open questions reads "None."; status `Planned`.

**STEP-1 — One "Show N more" row**
Files: `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` *(new)*, `PersonFamilySection.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-9.
Check: Family's fold row looks and reads as before; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-2 — One host for a section's read**
Files: `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` *(new)*, `PersonFamilySection.swift`. Skill: `add-workspace-place`.
Do: D-12 with Family as its only caller.
Check: the app builds; `PersonFamilyContentTests`, `PersonFamilyLayoutTests` pass; Family's loading, error, empty and diagram states keep their ids and look (its `#Preview`s).

**STEP-3 — The Timeline section's key**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift` (`sectionQueryKeys`: `.timeline` leaves the nil list), `macos/ProvenenciaTests/PlaceRegistryTests.swift`. Skill: `add-workspace-place`.
Do: D-3.
Check: T-1, T-2 pass.

**STEP-4 — The section's content and fixtures**
Files: `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)*, `PersonTimelineLifeline.swift` *(new)*, `PersonTimelinePreviewData.swift` *(new)*, `EventsListView.swift` (`includesRef:`), `macos/ProvenenciaTests/PersonTimelineContentTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-1, D-2, D-5 – D-8, D-13, D-15's geometry.
Check: T-3 – T-9 pass; `PersonsListTests` and the Events list tests pass unchanged.

**STEP-5 — The section view**
Files: `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)*, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-7, D-10, D-11, D-12's second caller, D-14, D-16 and the lifeline's drawing (D-15), and the board's frame Decisions; the `#Preview`s.
Check: the app builds; the previews render as § 6 describes each frame.

**STEP-6 — Wiring**
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`.
Do: `.timeline` in `PersonDetailSections` draws `PersonTimelineSection(entityId:)`.
Check: `scripts/check-pr.sh` passes; the manual check (§ 10).

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `design/S13-D4-timeline-section.md` → `design/archive/`, `design/README.md`, `docs/macos-client-patterns.md` (the sectioned detail page: a section that reads its own key builds on `ConclusionPageSectionHost`), `docs/ideas/structural-debt.md` (a row, next free A-number: `ReconciledValueDisplay.string(for:)`'s `.term` case is an older copy of the term-name rule).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys` (extended) | `.timeline.queryKey(project:entityId: "e1")` → `.personEvents(project:entityId: "e1")`; `.family` unchanged; `.vitals`, `.gallery`, `.sources` → nil | D-3 |
| T-2 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `personDetailOnTimelineWarmsTheTimeline` *(new)* | A Person location with `pageSection: .timeline` → `.conclusionDetail` and `.personEvents` | AC-4.2, D-3 |
| T-3 | Swift | `macos/ProvenenciaTests/PersonTimelineContentTests.swift` *(new)* `rowsFromEvents` | `PersonTimelinePreviewData.johann` → `dated` events 1 – 11 in input order, `undated` 12 and 13; event 4's title, line "4 Jun 1844" · "Bremen", roles ["Subject"], participants Maria, Friedrich with roles; event 13's line empty; event 1's `imageRelPath` set, event 11's empty | AC-1.1, AC-1.3, AC-1.5, AC-1.6, D-6, D-7 |
| T-4 | Swift | same file `splitKeepsOrderAndIds` | Input order kept (no sort) even when dates would sort otherwise; a `dated: false` row in the middle → in `undated`; one person on two events → two participants with distinct ids | D-6, Edges |
| T-5 | Swift | same file `rolesAndLocations` | Roles `witness`, `informant` → ["Witness", "Informant"]; a custom role labelled "Godparent" → "Godparent"; a custom role with an empty label and key `godparent` → "godparent"; a `.text` value → no role; Anna as `mother` → tile spoken "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851"; a participant's location → `PersonsList.location(…).withPageSection(.timeline)`; a row's location → `EventsList.location(…)` | AC-1.4, AC-2.1, AC-2.3, AC-2.4, D-1, D-2, D-5 |
| T-6 | Swift | same file `foldsAtFiftyAndFifteen` | `longLife` → 50 dated rows, `datedHidden` 70; with `.dated` expanded → 120; `largeEvent` → 15 participants, `participantsHidden` 26; with `.participants(eventId:)` expanded → 41; `[]` → `isEmpty` | AC-1.7, AC-3.1, AC-3.2, D-8, Scale |
| T-7 | Swift | same file `spokenLabelReusesTheEventsRow` | Event 1's `spokenLabel` begins with `EventsList.accessibilityLabel(event, orders:, includesRef: false)` and holds no ref; `EventsList.accessibilityLabel(event, orders:)` unchanged (with the ref) | D-6 |
| T-8 | Swift | `macos/ProvenenciaTests/PersonTimelineLifelineTests.swift` *(new)* `pointsToScale` | Johann's events with S13-33 positions, length 1000 → birth at 0, the burial (past death) at 1000; event 4 and event 5 eight months apart; the 1850 – 1859 gap about twelve times that; a range → `spanTo` set; a BEF date → `openStart`; points in row order | AC-5.1, AC-5.2, AC-5.3, AC-5.6, D-15 |
| T-9 | Swift | same file `placeholdersAndDegenerateLives` | No birth recorded → the birth placeholder before the baptism, outside the scale; a funeral and no death → the death placeholder before the funeral; only undated → both placeholders at the ends; one dated event → a minimum-length line, the point mid-line | AC-1.9, AC-5.4, AC-5.5, D-16 |

AC-1.2 is S13-05's (`TestPersonEvents`). AC-1.8, AC-1.9, AC-2.2, AC-2.5, AC-3.3, AC-4.1 are covered by the manual check (§ 10). Frame-dependent tests (the role mark's and spoken roles' wording, period markers, the Undated caption rule) are added when the board is agreed. Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `PersonFamilyContentTests`, `PersonFamilyLayoutTests` (STEP-2), `PlaceRegistryTests` (STEP-3), `PersonsListTests` and the Events list tests (STEP-4).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `PersonFamilyMoreRow`, `PersonFamilyLoading`, `familyShowMore` and `conclusions.family.showMore` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/` and `L10n.swift` with no hit left outside spike-13 plan history; `scripts/check-guidance-names.py` passes (check-pr runs it)
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

1. (AC-1.1, AC-1.3, AC-1.5, AC-1.6, AC-5.1 – AC-5.6) Open Johann, choose Timeline: the line from birth to death with the 1861 – 1880 stretch empty and the burial past death; events 1 – 11 earliest first, each pointing to its place; "Undated" with Residence and Migration between the markers; remove his birth record and see "Birth unknown" before the baptism; event 4 reads as AC-1.3; event 1's picture and event 11's Event mark.
2. (AC-1.2) Luise's birth is not listed.
3. (AC-1.4) Event 8 shows Witness and Informant.
4. (AC-2.1, AC-2.2, AC-2.4, AC-2.5) Click Anna's tile on event 1: her Timeline opens, crumbs *Persons › Weber, Anna › Timeline*; Back returns. Tab from event 1's title through its tiles; VoiceOver on Anna's tile reads AC-2.4.
5. (AC-2.3) Click event 4's title: the marriage's Event page.
6. (AC-3.1, AC-3.2, AC-3.3) In the "Large event" and "Long life" previews, the folds read "Show 26 more" and "Show 70 more" and unfold on press. In the app, on any person with a folded list, unfold it, open another person and Back: folded again.
7. (AC-4.1) With Johann's Timeline showing, edit Maria Schulz through her Evidence graph and return with Back: the change shows; "Refreshing" in the switcher row while it reloads.
8. (AC-1.7, AC-1.8, AC-1.9) A person with no events: the EmptyState; a Timeline entry from history for a person merged since: the danger Callout; a person with only undated events.
9. Resize across 1600pt on Timeline: section, scroll, focus and unfolded lists kept (TL-W3); the Wide arrangement the board chose. Dark mode as frame 11.
10. Family looks and behaves as before (tiles, "Show N more", loading, error and empty states).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S13-32 lands with names or slots other than its scope entry says | STEP-0 refreshes § 4, D-4, D-5 and D-11 before any code |
| The board picks Wide option B, which needs a secondary column for page sections that `ConclusionDetailPage` does not offer | Decided with the board (§ 12 question 7); checked against **Size**; may become a preparatory PR |
| Moving Family onto the host changes its states | STEP-2's check: Family's ids and previews unchanged |
| 50 rows of tiles render slowly | First paint is bounded by D-8; `LazyVGrid` and `CachedThumbnail` load lazily |

**Stop and report if:**

- `CatalogPersonEvent`, `CatalogEventParticipant`, `.personEvents` or `ConclusionPageSection.queryKey` differ from § 4.
- S13-32 has not landed, or its tile shell has no declared minimum width or no Return modifier (D-10, D-11).
- The board needs data `ListPersonEvents` does not return (an age, a participant order the read does not give): a Decision and a Go step, not Swift logic.
- Family's look, ids or tests change beyond names in STEP-1 or STEP-2.
- `EventsList.accessibilityLabel`'s existing callers would need a change for D-6.

## 12. Open questions

Questions 1 – 9 wait on the S13-D4 board, which the brief asks; question 10 waits on S13-32 and S13-33. Recommendation in brackets.

1. **Row form (TL-1).** A Card per event, or rows divided by rules? [Card per event, as the tile is a card.]
2. **Picture size (TL-2).** The Events list row's size or larger? [The Events list's size.]
3. **Role mark and "Subject" (TL-4).** Badge or "as …" line; how "Subject" reads on the person's own events? [A subtle neutral Badge; "Subject" as worded.]
4. **The lifeline (TL-5, TL-5a).** Orientation, scale rule (fit the section or points per year), how a row points to its place, where undated rows sit between the markers, the placeholder's distance and look. [A vertical rail beside the list, points per year with a minimum and maximum, a connector from each row.]
5. **Participant role caption (TL-6).** Its spacing above the tile, and the grid's columns at Regular and Wide. [As the Family captions.]
6. **Participant order (TL-6).** The read's order, or by role or title? [The read's order; another order adds a Go sort here.]
7. **Wide (TL-W1).** Primary-only, or a secondary column (decade jump list or life summary)? [The board recommends; option B needs a page-section secondary slot, checked against Size.]
8. **Empty copy (TL-10).** [As the brief.]
9. **Spoken roles (TL-12).** "as Subject" or the bare role after the event. ["as Subject", as the brief's sample.]
10. **S13-32's and S13-33's names.** The tile shell, its minimum width, the Return modifier and the term resolver as landed (STEP-0).

## 13. PR

**Title:** S13-13 — Timeline section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md`.
