# S13-13 — Timeline section

**Status:** Blocked — on S13-33
**Spike / slice:** Spike 13 · Slice 4 — Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D4](../design/S13-D4-timeline-section.md), board drawn 2026-11-02 (turn 12) and agreed with the user's answers of 2026-11-02 (compact cards, participant order by role, full width at Wide)
**Board:** Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Timeline.dc.html*, turn 12 "S13-D4 · Timeline — the event list and the lifeline (update 2026-11-02)", frames 12a – 12y. The implementer may not be able to open it: **Design → macOS** describes every frame in words and gives every measure
**Depends on:** S13-05 (landed [#680](https://github.com/mendahu/provenencia/pull/680)), S13-12 (landed [#676](https://github.com/mendahu/provenencia/pull/676)), S13-33 ([plan](S13-33-timeline-positions.md): each event's span, the birth and death markers, participants ordered by role then name)
**Blocks:** —
**Branch:** `s13-13-timeline-section`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go or proto change (S13-33 has them)
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-05 and S13-12 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 3. Timeline](../person-detail.md#3-timeline) (R-2, R-3, R-23 – R-28, R-24a); brief [S13-D4](../design/S13-D4-timeline-section.md); the S13-05, S13-12 and S13-33 plans ([`S13-05-person-timeline-read.md`](S13-05-person-timeline-read.md), [`S13-12-family-section.md`](S13-12-family-section.md), [`S13-33-timeline-positions.md`](S13-33-timeline-positions.md)) and their **Landed** notes; working rules: "Engine PRs run ahead of their slice" (data, not layout), "Each section reads its own query", "Timeline is direct participation, any role", "One person tile, two sizes, in the feature", "Colors reuse `PersonGenderTreatment`", "A timeline tile opens the participant on Timeline", "Timeline lists fold", "The timeline is a lifeline", "The Timeline shows each event's picture", "Sections are places"
**Planned:** 2026-11-02 against `main` @ `51dccd4`; refreshed the same day after the board (turn 12), the descoping of S13-32 and the S13-33 plan. STEP-0 re-checks S13-33's landed names
**Review:** [`S13-13-timeline-section.review.md`](S13-13-timeline-section.review.md) — round 1 Fits with changes / Ready with changes on the draft; round 2 pending on this refresh

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

A researcher opens a Person's Timeline section and reads the person's life on a lifeline drawn to scale from birth to death beside a list of events, earliest first. Each dated event points to its place on the line, so the stretches with no records show as empty line. Each row shows the event's picture, title, date and place, and the person's role; the other people who took part follow as compact person cards, ordered by role. Birth and death always have a marker, a placeholder when unrecorded; undated events sit just before the death marker. A click on an event opens its Event page; a click on a card opens that person's Timeline. It fills the empty place S13-10 left (S13-D1 PPF-10) with the data S13-05 and S13-33 read.

**Why now.** Slice 4's Check is bar 4 ("A life in order"); S13-05 and S13-33 give the read, S13-12 has the tile, the board is drawn, and this is the last PR of the slice.

**What stays.** The header, the switcher, Vitals, Family (its look; its host code and fold row move, D-9, D-12), Gallery and the Sources empty place; the regular person tile; `ListPersonEvents` and its order (S13-05, S13-33); the Events list and Event page; every other link into a Person page; every other page on Page columns (D-17 adds an opt-in mode).

## 2. User stories

**US-1.** As a researcher, when I open a Person's Timeline, I see every event they took part in, in date order, so that I read their life from birth to death.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann (brief § 5 sample, events 1 – 13; on the board his death is 18 Sep 1880 and his burial 22 Sep 1880) | Events 1 – 11 in that order; 12 (Residence at Altona) and 13 (Migration) under "Undated", just before the death marker (frame 12a; R-23, R-24; D-6, D-16) |
| AC-1.2 | Luise's birth names no father | It is not listed (R-23; S13-05 filters it) |
| AC-1.3 | Event 4 | Title "Marriage of Johann Weber and Maria Schulz", "4 Jun 1844 · Bremen", role Badge "Subject" (outline), cards Maria Schulz (Subject) and Friedrich Weber (Witness) (R-25, R-27; D-5, D-6) |
| AC-1.4 | Event 8: Johann is Witness and Informant | Two neutral Badges, "Witness" then "Informant" (D-5) |
| AC-1.5 | Event 1 has a lead picture; event 11 has none | Event 1 shows its picture at 72pt; event 11 shows the Event mark on the sunken tone (R-26; D-7) |
| AC-1.6 | Event 13: no date, no place | The date and place line is not drawn |
| AC-1.7 | A person with no promoted events | The EmptyState of TL-10, no line (frame 12m) |
| AC-1.8 | The read fails with `conclusiondetails.not_found` | A danger Callout in the section titled "Timeline could not be read" with the mapped message; the header and switcher still show (frame 12o) |
| AC-1.9 | Only undated events (Ida, frame 12l) | The list reads "Birth unknown", "Undated" with its rows, "Death unknown"; the line is the two placeholders joined by a dashed stretch (D-15, D-16) |

**US-2.** As a researcher, I can move from an event to its page and from a participant to their life.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | On Johann's Timeline, click Anna Weber's card on event 1 | Anna's Person page opens on Timeline; breadcrumbs *Persons › Weber, Anna › Timeline*; Back lists "Johann Weber · Timeline" (frame 12r; D-1) |
| AC-2.2 | Then Back | Johann's Timeline again |
| AC-2.3 | Click event 4's title | The Event page of that marriage opens, titled as the Events list titles it (D-2) |
| AC-2.4 | VoiceOver on Anna's card under event 1 | "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851", a button: the full life lines, not the card's years (D-4, D-5) |
| AC-2.5 | Tab from event 1's title | Focus moves to its cards in order, then any "Show N more", then event 2's title; Return or Space activates (frame 12q; D-10) |
| AC-2.6 | Hover Anna's compact card | The tooltip gives her name and full *b.* and *d.* lines (D-4) |

**US-3.** As a researcher, long lives and crowded events stay readable without hiding anything.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | An event with 41 other participants | 15 cards, then "Show 26 more"; pressing it shows all 41 (frames 12g, 12h; D-8). Evidence: T-6, and the `#Preview` "Large event" (D-13) |
| AC-3.2 | A person with 120 dated events | 50 events, then "Show 70 more"; the line still shows all 120 points; a click on a folded event's point unfolds the list and scrolls to its row (frame 12j; D-8, D-15) |
| AC-3.3 | Unfold a list, then open another person and come Back | The list is folded again (D-8, as Family) |

**US-4.** As a researcher, the section stays current while I work.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Johann's Timeline is showing; a write touches Maria Schulz's handle | The timeline reloads; the list and line stay drawn and the switcher row shows "Refreshing" until it lands (frame 12p; D-3) |
| AC-4.2 | Opening a Person straight on Timeline (from a card, or Back) | The timeline read starts with the page, not after the detail read (D-3) |

**US-5.** As a researcher, I see the shape of the life at a glance: where its records cluster and where they fall silent (R-24a).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | Johann: birth 12 Mar 1817, death 18 Sep 1880 | A vertical line in an 80pt column beside the list, from the birth cap at the top to the death cap, then on to the burial four days later; each dated event has a point; nothing marks 1861 – 1880 but empty line (frame 12a; D-15) |
| AC-5.2 | Event 4 on 4 Jun 1844 and event 5 on 2 Feb 1845 | Their points are 0.67 years apart at the line's scale, which is the list's height over the life's span (about 43pt a year for Johann at Regular) (D-15) |
| AC-5.3 | A day; "1835"; "1844 – 1849"; "bef. 1900" | A dot; a bar over 1835; a bar from 1844 to 1850; a bar ending at 1901 with an open (fading) start (D-15) |
| AC-5.4 | Johann without his birth and death events (frame 12w) | A dashed "?" cap and the muted row "Birth unknown" 40pt before the baptism, the stretch dashed and not to scale; "Death unknown" after the 1861 marriage and before the burial; no year shown (D-16) |
| AC-5.5 | A funeral in 1880, a census in 1875, no death | "Death unknown" sits after the census and before the funeral; the funeral sits at its own date (D-16) |
| AC-5.6 | A child born 1845, died 1848, three events (frame 12x) | The same line height spread over three years: year labels with quarter ticks (D-15) |
| AC-5.7 | Click a point on the line | The list scrolls to that event's row (D-15) |
| AC-5.8 | Scroll the list | Each row in view has a short curve from its point on the line to the dot at the row's edge, in the accent; rows out of view have none; an undated row has a dashed dot and no curve (frames 12a, 12b; D-15) |

**US-6.** As a researcher on a wide window, I see more people per event (TL-W1).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-6.1 | Johann's Timeline at Wide (pane 2296) | The section fills the pane between the page gutters; the line keeps its 80pt column; titles and date lines keep their Regular measure; each event's cards fill the row at a 232pt minimum, seven or eight a row (frame 12d; D-11, D-17) |
| AC-6.2 | The header and switcher at Wide | Unchanged: they keep the 1180pt group (D-17) |
| AC-6.3 | Resize across 1600pt | Section, scroll position, focus and unfolded lists kept (TL-W3) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Shared pieces: the kit's Return modifier, a term value's name, Page columns' full-width primary (D-10, D-5, D-17) | STEP-1, AC-2.5, AC-6.1, AC-6.2 |
| One "Show N more" row and one host for a section's own read; Family moves onto both, unchanged (D-9, D-12) | STEP-2 |
| The person tile's compact size (TL-6; D-4) | STEP-3, AC-2.4, AC-2.6 |
| The Timeline section's query key (D-3) | STEP-4, AC-4.1, AC-4.2 |
| The section's content and lifeline geometry (D-1, D-2, D-5 – D-8, D-15, D-16) | STEP-5, AC-1.1 – AC-1.6, AC-1.9, AC-2.1, AC-2.3, AC-3.1, AC-3.2, AC-5.1 – AC-5.6 |
| The section view, its line and its states (TL-1 – TL-13, TL-W1 – TL-W3; D-11, D-14, D-15) | STEP-6, STEP-7, AC-1.7, AC-1.8, AC-5.7, AC-5.8, AC-6.1 |

| Out | Home |
| --- | --- |
| Positions, markers and participant order in Go | S13-33 |
| Zoom, pan, a draggable canvas or a swimlane | not planned (brief § 2.1) |
| Gap labels; an estimated year for an unknown birth or death; marks on events outside the life | not planned (user, 2026-11-02) |
| Age at each event | not planned this spike (brief § 2.1) |
| Filtering or searching the timeline | not planned |
| The provenance control on a row, role or card | not planned this spike (R-13b) |
| Editing events or participations from the section | not planned this spike |
| Full width for Vitals, Family or Gallery | not planned; D-17's mode is opt-in per section (Gallery may adopt it later, S13-D5) |
| Renaming the tile's "this person" keys (`familyThisPerson`, `familyThisPersonSpoken`) | not planned: the Timeline has no focus card |
| `ReconciledValueDisplay.string(for:)`'s `.term` case (label else key, no product names), an older copy of the term-name rule | logged in [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) by STEP-8 |
| `EventSubject` folding into {role, `PersonSummary`} | logged by S13-05 in [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Timeline section | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` | `.timeline, .sources` draw one `PVEmptyState` with `L10n.Conclusions.pageSectionLater`. `.family` draws `PersonFamilySection(entityId:)` |
| Section key | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection.queryKey(project:entityId:)` | `.family → .personFamily`; every other section nil. Read by `PlaceRegistry` Spec `.personDetail` (warms it) and `ConclusionDetailLoaded.sectionRefreshing` (`ConclusionDetailPage.swift`) |
| Timeline key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.personEvents(project:entityId:)`; `CatalogQueryRegistry.swift` its descriptor | `.sessionFresh`, `.evict` when hidden, refreshes on any touched handle. Value today `[CatalogPersonEvent]`; S13-33 makes it `CatalogPersonTimeline` |
| Timeline data | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonEvent`, `CatalogEventParticipant`; after S13-33 also `CatalogPersonTimeline { events; birth; death }`, `CatalogPersonEvent.span: CatalogTimelineSpan?` (`from`, `to` on the year line, `openStart`, `openEnd`; nil exactly when `dated` is false), `CatalogLifeMarker { eventId: String?; besideEventId: String? }` | Rows in date order (an open-start date at its known end), undated last; participants ordered by role rank, then name (S13-33 D-10); roles are `.term` values with labels filled in Go |
| Person tile | `macos/App/Features/Conclusions/PersonTile.swift` `PersonTile`, `PersonTileContent`, private `PersonTileButtonStyle`, `PersonTileButtonBody` | One size: 92pt, a 70pt likeness, a two-line title, *b.* and *d.* lines. `PersonTileContent.init(summary:role:isFocus:locale:orders:)` speaks the role and never draws it. Return handled by hand (`.focused` + `.onKeyPress(.return)`). Callers: `PersonFamilySection.swift`, `PersonFamilyContent.swift`, `PersonFamilyLayout.swift` (`PersonTile.height`) |
| Life lines | `macos/App/Features/Conclusions/PersonLifeDisplay.swift` `lifeLines(_:locale:missing:)` | Builds *b.* and *d.* lines from `CatalogPersonSummary`; no years-only form |
| Show more | `macos/App/Features/Conclusions/PersonFamilySection.swift` private `PersonFamilyMoreRow`; `L10n.Conclusions.familyShowMore(_:)` (key `conclusions.family.showMore`) | A ghost `PVButton` with hand-written Return handling; Family only |
| Folds | `macos/App/Features/Conclusions/PersonFamilyContent.swift` `foldLimit`, `Fold`; `PersonFamilySection` `@State expanded`, reset `.onChange(of: entityId)` | The pattern to copy |
| Section host | `macos/App/Features/Conclusions/PersonFamilySection.swift` `PersonFamilySection`, private `PersonFamilyLoading`, `PersonFamilyLoaded` | Handle lookup, `.task`, 96pt spinner, danger Callout, EmptyState and ids around Family's diagram |
| Read states | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPresentation` | Generic over the read's value (S13-12 D-10) |
| Page columns | `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` `PVPageColumns`, `PVPageColumnsMetrics.columnWidths(pane:widthClass:hasSecondary:)`; `ConclusionDetailPage.swift` `ConclusionDetailLoaded`; `macos/ProvenenciaTests/PVPageColumnsTests.swift` | The primary column is `min(widthContentMax, pane − 2 · gutterPage)` at every width: 1180pt at Wide. Sections fill the primary only |
| Event title, line, location, spoken | `macos/App/Features/Conclusions/EventTitleDisplay.swift` `titleSource(_:locale:orders:)`; `EventsListView.swift` `EventsList.location(_:orders:)`, `EventsList.accessibilityLabel(_:orders:)`, `EventSecondaryDisplay.content(_:locale:)`, `EventSecondaryLine` | The Events row's builders |
| Role labels | `macos/App/Platform/PropertyTermDisplay.swift` `name(key:origin:propertyKey:catalogLabel:)`, `name(term:propertyKey:)`; `SeededPropertyKey.role`; `macos/ProvenenciaTests/PropertyTermDisplayTests.swift` | Product name, else the trimmed catalog label, else the key. No overload for a `CatalogConclusionValue` |
| Person link | `macos/App/Features/Conclusions/PersonsListView.swift` `PersonsList.location(_:orders:)`; `WorkspaceLocation.withPageSection(_:)` | As S13-12 D-1 |
| Thumbnail | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail`; mark `PVMarkKey.subjectEvent`; `L10n.Conclusions.eventPictureThumb` | The Events list's thumb and fallback (44pt there) |
| Divider | `macos/App/DesignSystem/Components/Divider/` `PVDivider` | The kit rule between rows |
| Preview data | `macos/App/Features/Conclusions/PersonFamilyPreviewData.swift` `PersonFamilyPreviewData` (`#if DEBUG`), `summary(_:_:born:)` | Shared by Family's previews and tests |
| README | `macos/App/DesignSystem/README.md` built table; `scripts/check-guidance-names.py` | One row per `Components/<Name>/` folder |
| Tests | `macos/ProvenenciaTests/PersonTileTests.swift`, `PersonFamilyContentTests.swift`, `PersonFamilyLayoutTests.swift`, `PersonDetailTests.swift` (`sectionQueryKeys`), `PlaceRegistryTests.swift`, `CatalogQueryRegistryTests.swift` | — |

**Precedent.** S13-12 ([#676](https://github.com/mendahu/provenencia/pull/676)) is the same shape: a section that reads its own key through `ConclusionPageSection.queryKey`, a pure content model with folds, a pure layout tested on its geometry (`PersonFamilyLayout`), a section view on `ConclusionDetailPresentation`, tiles that navigate with `navigation.go(to:)`, and preview data shared with the tests. Copy those shapes; Family's host moves into D-12 rather than being copied.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A participant's card opens `PersonsList.location(participant.person, orders:).withPageSection(.timeline)` with one `navigation.go(to:)` (user, 2026-11-02) | Working rule "A timeline tile opens the participant on Timeline" | Vitals (S13-D1 PPF-5); building `.personDetail(` by hand |
| D-2 | An event opens `EventsList.location(event, orders:)` with one `navigation.go(to:)` (R-28) | The Events list's location: same title and history entry | A hand-built `.eventDetail(` |
| D-3 | `ConclusionPageSection.queryKey` gains `.timeline → .personEvents(project:entityId:)`; `PlaceRegistry` warms it and the switcher's refresh follows it through S13-12's two readers | S13-12 D-8: one declaration per section key | A per-section branch in `PlaceRegistry` |
| D-4 | The person tile gains a size: `enum PersonTile.Size { case regular, compact }` *(new)*, `PersonTile(content:projectDir:size: = .regular, open:)`. **Regular** is today's tile, unchanged. **Compact** (frame 12f): 52pt tall, `PVCard(cornerRadius: PVRadius.sm, padding: 6)`, a 40pt `CachedThumbnail` with the same gender colors, the title on one line (`ConclusionListRow.title(_, lineLimit: 1)`), and one line of life years in mono caption (`PVFont.mono(size: PVTypeScale.caption)`, `PVColor.textMuted`): `PersonLifeDisplay.years(_ summary:) -> String?` *(new)* gives "1790 – 1851", "b. 1820", "d. 1851" or nil, each year the date's `startYear` (a date with no year is left out). The full *b.* and *d.* lines go in `.help` (the title, then the lines, joined by newlines) and the spoken label is unchanged (`PersonTileContent.spokenLabel`). Heights and the likeness size are `PersonTile.Size` properties (`height`, `likeness`), `nonisolated`; `PersonTile.height` becomes `PersonTile.Size.regular.height` at its one caller (`PersonFamilyLayout`). A compact card is never the focus | Working rule "One person tile, two sizes, in the feature" (user, 2026-11-02: the cards are smaller on purpose); one type, one content, one interaction | A second card type; the kit shell (S13-32, descoped); dropping the tooltip (the years lose the day and place) |
| D-5 | Role labels: `PropertyTermDisplay.name(_ value: CatalogConclusionValue, propertyKey: String) -> String?` *(new)* in `macos/App/Platform/PropertyTermDisplay.swift`: for `.term(_, key, label, origin)` it returns `name(key: key, origin: origin, propertyKey: propertyKey, catalogLabel: label)`; nil for any other case. Each role is `name(value, propertyKey: SeededPropertyKey.role)`. The person's roles show as Badges in the read's order: `PVBadge` neutral for each, the `subject` key in the outline style (frame 12f). A participant's role caption (micro caps above the card) and spoken role are its names joined with `L10n.DesignSystem.spokenJoin` | One rule turns a term into a name (round 1 M-2); the board's Badges | A timeline-only role type with its own empty-label rule |
| D-6 | `PersonTimelineContent` *(new, pure, `Equatable`)* in `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)*: `init(timeline: CatalogPersonTimeline, expanded: Set<Fold> = [], locale:, orders:)`. `dated: [Row]` and `undated: [Row]` keep `timeline.events`' order, split by `dated`; Swift never sorts. `Row`: `id` (event entity id), `title` (`EventTitleDisplay.titleSource`), `line: EventSecondaryDisplay.Content`, `imageRelPath`, `roles: [RoleBadge]` (`text`, `isSubject`), `location` (D-2), `span: CatalogTimelineSpan?`, `participants: [Participant]`, `participantsHidden`, `spokenLabel` = `spokenJoin(EventsList.accessibilityLabel(event, orders:, includesRef: false), L10n.Conclusions.timelineAsRole(roles))` ("Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject"). `EventsList.accessibilityLabel` gains `includesRef: Bool = true` *(new parameter)*: with false it uses the title and date only (`L10n.Conclusions.lineJoin` when both), then the place and "+N places" as today. `Participant`: `id` ("\(eventId)-\(personId)"), `tile: PersonTileContent` (role: the spoken role), `caption: String?`, `location` (D-1). `birth`, `death: Marker` (`eventId: String?`, `besideEventId: String?`, `isPlaceholder`), copied from the timeline. `isEmpty` is `timeline.events.isEmpty` | Go orders and places; Swift lays out; the Events row's own builders, spoken label included | Sorting in Swift; a second spoken label (round 1 M-3) |
| D-7 | The picture is `CachedThumbnail(projectDir:relPath: row.imageRelPath, size: 72, cornerRadius: PVRadius.sm, fallbackMark: .subjectEvent, fallbackLabel: L10n.Conclusions.eventPictureThumb, colors: nil)` (frame 12f) | The board's 72pt on a row that leads with the event; the Events list's mark and label | The Events list's 44pt (the board weighed it and chose 72) |
| D-8 | Folds (working rule "Timeline lists fold"): `eventFoldLimit = 50`, `participantFoldLimit = 15`; `enum Fold: Hashable { case dated, undated, participants(eventId: String) }`; `datedHidden`, `undatedHidden`, each row's `participantsHidden`. The line still gets every dated event's span (D-15). `expanded` lives in the section as `@State`, reset on a new entity id | The user's caps; Family's fold shape | A shared fold type with Family |
| D-9 | One "Show N more" row: `PersonFamilyMoreRow` moves to `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` *(new)* as `ConclusionShowMoreRow(count:expand:)`, on the Return modifier (D-10). `L10n.Conclusions.familyShowMore(_:)` → `showMore(_:)`, key `conclusions.showMore`, comment "Button under a folded list on a Conclusion page that shows the rest; argument is how many are hidden"; the String Catalog entry moves with it. Family calls it, unchanged | One row and one string for both sections | A private copy |
| D-10 | Return activation is one kit modifier: `func pvActivatesOnReturn(_ action: @escaping () -> Void) -> some View` *(new)* in `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` *(new)*: a view modifier that owns a `@FocusState`, applies `.focused`, and handles `.onKeyPress(.return)` only while focused, as `PersonTile` does today. Space stays the `Button` default. `PersonTile`, `ConclusionShowMoreRow` and the event title use it; a README row `Components/ReturnActivation/PVReturnActivation.swift` in the built table | Three callers in this PR (round 1 M-4); content-agnostic, so a component | A third hand copy |
| D-11 | Each event's cards lay out in `LazyVGrid(columns: [GridItem(.adaptive(minimum: PersonTimelineContent.cardMinimumWidth), spacing: 8)], alignment: .leading, spacing: 8)`, `cardMinimumWidth = 232` *(new)*, indented under the title (the grid starts at the title's leading edge), each card `PersonTile.Size.compact.height` tall, the role caption above each card in `.pvMicroCaps()` `PVColor.textMuted` (frame 12f) | The board's 232pt minimum: three or four a row at Regular, seven or eight at Wide | Family's column formula |
| D-12 | One host for a page section's own read: `ConclusionPageSectionHost<Value, Content: Equatable, Body: View>` *(new)* in `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` *(new)*: the key, an identifier root, the glyph, the error title, the empty title and message, `make: (Value) -> Content`, `isEmpty`, and a body builder. It reads the session from the environment, gets the handle with `session.queryHandle(key)`, starts it in `.task`, and maps `ConclusionDetailPresentation`: `.loading` → `ProgressView()` 96pt from the top; `.failed` → `PVCallout(tone: .danger, title:, message:)`, id `<root>.error`; empty → `PVEmptyState`, id `<root>.empty`; else the body, id `<root>`. Family moves onto it with its ids, copy and look unchanged; the Timeline is its second caller | Only the content differs (round 1 M-6); S13-16 is the third | Copying Family's host |
| D-13 | Fixtures: `macos/App/Features/Conclusions/PersonTimelinePreviewData.swift` *(new)* (`#if DEBUG`, internal) `enum PersonTimelinePreviewData` builds `CatalogPersonTimeline` values for the board's frames: `johann` (13 events, death 18 Sep 1880, burial 22 Sep 1880, pictures on events 1, 2, 4, 5, 9, 10), `unknownEnds` (johann without events 1 and 10, both markers placeholders), `ernst` (1845 – 1848, three events, baptism "Dec 1845"), `largeEvent` (a burial with 41 participants), `lise` (120 events over 98 years, residences as ranges, a burial "bef. 1900"), `ida` (two undated events), `ludwig` (none). People from `PersonFamilyPreviewData.summary`. The section's `#Preview`s and the tests use it | One fixture for previews and tests (S13-12 Landed note (6)) | A test-only copy |
| D-14 | Accessibility ids under `persons.detail.timeline`: the list; `.line`; a row `.row.<eventId>`; its title `.title.<eventId>`; a card `.tile.<Participant.id>`; `.undated`; `.birth`, `.death`; fold rows `.more.dated`, `.more.undated`, `.more.<eventId>`; D-12's `.error` and `.empty` | Unique ids | One shared `.more` id |
| D-15 | The lifeline (frames 12a – 12x). Geometry is a pure function, `PersonTimelineLifeline` *(new)* in `macos/App/Features/Conclusions/PersonTimelineLifeline.swift` *(new)*: `init(content:, rowFrames: [String: CGRect], listHeight: CGFloat)`. **Orientation and size:** vertical, in an 80pt column at the list's leading edge, scrolling with it; it spans the list's measured height (from the first row's top to the last row's bottom). **Scale:** the placed range runs from the earliest `span.from` to the latest `span.to` among the dated events (the burial past death included); points per year = (height − placeholder stubs) ÷ years; a placeholder takes a fixed 40pt dashed stub outside the scale. A placed range under one year is drawn as one year. **Ticks:** the smallest step in 1, 2, 5, 10, 20, 50 years whose ticks are at least 34pt apart, labelled with the year in mono micro; quarter ticks between years when a year is at least 136pt. **Marks:** a day (`to − from` ≤ 1/365) is a dot; a span a bar from `from` to `to`; `openStart` or `openEnd` fades that end; birth and death are caps at the line's ends (a recorded marker at its event's point). **Tethers:** a row whose frame is in the viewport gets a short curve from its point to a dot at the row's leading edge, in `PVColor.accent`; the others none; undated rows get a dashed dot and no curve. **Clicks:** a point scrolls the list to its row (`ScrollViewReader` on `.row.<eventId>`); a folded row's point first adds `.dated` to `expanded`. The line, ticks, points and curves are `.accessibilityHidden(true)`. Swift does no date arithmetic: positions are S13-33's year-line numbers | Working rule "The timeline is a lifeline"; the board's TL-5; a pure, tested function as `PersonFamilyLayout` | Fixed points per year (a short life is a dot); `GraphCanvas`; date maths in Swift |
| D-16 | Markers in the list (TL-3, TL-5a): a recorded birth or death is its event's row. A placeholder is a muted italic row, "Birth unknown" or "Death unknown", with a dashed "?" cap on the line, placed in the list before the row `besideEventId` names (birth) or after it (death), at the top or bottom when there is none. The undated rows sit **just before the death marker** (before the death row, or the death placeholder), under "Undated" (`.isHeader`) and the hint "No date, so no place on the line" (frames 12a, 12l) | The board's TL-3 and TL-5a; positions are Go's | Undated at the list's end (R-24 changed); an estimated year |
| D-17 | Full width at Wide (user, 2026-11-02): `PVPageColumns` gains `primaryFill: Bool = false` *(new)*. When true and the width class is `.wide`, the primary column is `pane − 2 · gutterPage` and there is no secondary; the header band keeps the 1180pt group. `PVPageColumnsMetrics.columnWidths` gains the same parameter. `ConclusionPageSection.fillsWidth: Bool` *(new)*, true for `.timeline` only; `ConclusionDetailLoaded` passes `primaryFill: pageSection?.fillsWidth ?? false`. Inside the Timeline, titles and date lines keep the Regular measure (`.frame(maxWidth: PVSpacing.widthContentMax − 80, alignment: .leading)`); the card grid fills | The board's TL-W1 (frame 12d) and the user's answer; one opt-in mode in the component that owns widths | A Timeline-only frame override; capping at 1180pt (the user chose full width) |

**Rules this plan bends.** S13-D1 PPF-5 → D-1 opens cards on Timeline → user, 2026-11-02 → working rule "A timeline tile opens the participant on Timeline" (planning commit). R-24 and R-26 → D-16, D-7 → user, 2026-11-02 → `person-detail.md` updated (planning commit). "One person tile, built once" (one size) → D-4's compact size → user, 2026-11-02 ("I decided to make them smaller") → working rule "One person tile, two sizes, in the feature" (planning commit). S13-D1 PPF-W3 (a full-width Page columns mode "does not exist", a note for later) → D-17 adds it, opt-in → user, 2026-11-02 → `docs/macos-client-patterns.md` § 9 gains the mode (STEP-8). No other rule is bent.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `pvActivatesOnReturn` *(new)* | component | `PersonTile`'s and `PersonFamilyMoreRow`'s hand-written Return | Extracted on the third caller (D-10) |
| `PropertyTermDisplay.name(_:propertyKey:)` *(new overload)* | func | `name(term:propertyKey:)` | Same rule, one more input type (D-5) |
| `PVPageColumns(primaryFill:)` *(new parameter)*; `ConclusionPageSection.fillsWidth` *(new)* | component, property | `PVPageColumnsMetrics.columnWidths` | Extended (D-17) |
| `PersonTile.Size` *(new)*; `PersonLifeDisplay.years` *(new)* | enum, func | `PersonTile`, `PersonLifeDisplay.lifeLines` | Extended (D-4) |
| `ConclusionShowMoreRow` *(moved)*; `L10n.Conclusions.showMore` *(renamed)* | snowflake, key | `PersonFamilyMoreRow` | Shared (D-9) |
| `ConclusionPageSectionHost` *(new)* | snowflake | `PersonFamilySection`'s host | Extracted (D-12) |
| `EventsList.accessibilityLabel(_:orders:includesRef:)` *(new parameter)* | func | the same | Extended (D-6) |
| `PersonTimelineContent`, `PersonTimelineLifeline` *(new)* | pure model, pure layout | `PersonFamilyContent`, `PersonFamilyLayout` | A different read and a time scale |
| `PersonTimelineSection` *(new)* | snowflake | `PersonFamilySection` | The list and the line, on D-12's host |
| `PersonTimelinePreviewData` *(new)* | fixture | `PersonFamilyPreviewData` | Timeline events (D-13) |
| `L10n.Conclusions.timeline*`, `lifeYears*` *(new)* | keys | `L10n.Conclusions.family*` | New copy |

**Builds on.** `CatalogQueryKey.personEvents` (S13-05) with S13-33's value type, `ConclusionPageSection.queryKey` and its readers (S13-12), `ConclusionDetailPresentation`, `PersonTile`, `PersonTileContent`, `PropertyTermDisplay`, `EventTitleDisplay`, `EventSecondaryDisplay` / `EventSecondaryLine`, `EventsList`, `PersonsList.location`, `CachedThumbnail`, `PVBadge`, `PVCard`, `PVDivider`, `PVEmptyState`, `PVCallout`, `PVPageColumns`, `pvMicroCaps`, `L10n.DesignSystem.spokenJoin`.
**Deletes.** `PersonFamilyMoreRow` (moved); `familyShowMore` and key `conclusions.family.showMore` (renamed); `PersonFamilyLoading` and `PersonFamilySection`'s host code (moved into D-12); the hand-written Return handling in `PersonTile` and the fold row (D-10); `PersonTile.height` (now `PersonTile.Size.regular.height`); `.timeline` leaves the `pageSectionLater` case.
**Next adds.** S13-16 adds `.sources → .personClaims` and is the host's third caller; Gallery (S13-D5) can opt into `fillsWidth`; an Event page with sections (R-1) can reuse the event row.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A card opens a Person on Timeline | `PersonDetailView` / `ConclusionDetailPage` | `viewIdentity` resets the page; `expanded` resets on `entityId`; a merged participant gives the page's not-found Callout |
| An event title opens an Event page | `EventDetailView` | The Events list's path |
| Back to a Timeline entry | `PersonTimelineSection` | D-3 warms both keys; `.evict` while hidden, so Back reloads |
| Only undated; no events; one dated event | `PersonTimelineContent`, `PersonTimelineLifeline` | Placeholders only (AC-1.9); EmptyState (AC-1.7); a range under a year drawn as a year (T-9) |
| A Person page at Wide on Timeline, then on Vitals | `PVPageColumns` | `primaryFill` follows the open section; Vitals is capped again (T-10) |
| An older engine with no spans or markers | `PersonTimelineContent` | No points, both markers placeholders at the ends; the list still reads (S13-33 AC-3.2) |

### Scale

At the brief's sizes (8 – 25 events typical; 120 events and a 41-person event large): one read per view, warm with no query (S13-05). Swift builds one row per event and one tile content per participant; the first paint draws at most 50 rows and 15 cards each (D-8), cards in a `LazyVGrid`. `PersonTimelineLifeline` is linear in the events and recomputes on a change of `rowFrames`, `listHeight` or the viewport, not per frame: row frames come from one `onGeometryChange` per row and tethers are only for rows in the viewport. T-11 lays out 120 events.

### Edges

- **Assumptions.** The read's rows are in date order with undated last (S13-05, S13-33); D-6 splits by `dated` and keeps order (T-6). A dated row has a span and an undated row none (S13-33 AC-1.7); D-15 reads `span` only. The line spans the list's height, so unfolding a list changes the scale; T-9 holds the recomputation.
- **Neighbors.** A role that is not a term → no Badge; an empty catalog label → its key (D-5). A placeholder with no `besideEventId` → at the list's top or bottom. A burial past death → the line runs on past the death cap (AC-5.1). A birth event with no date → a recorded marker with `besideEventId` set: drawn as a placeholder cap beside that event, with its own row (S13-33 AC-2.5).
- **Removed inputs.** An event or participant unpromoted or merged after the read: the next refresh drops it (the key follows every touched handle).
- **Real data.** None — no table drives the section; S13-33 owns the role rank and the after-death table.
- **Untrusted size.** Events and participants are not bounded in Go; the first paint is bounded by D-8, and the line's points are linear.

### Schema

None — no catalog change.

### Go core

None — S13-33.

### FFI

None — S13-33.

### macOS

**Frames in words** (turn 12; every frame is the Person page pane with the switcher on Timeline; header and switcher as S13-D1):

- **12a / 12d Full timeline, Regular and Wide.** An 80pt line column at the left; the list beside it. Each row, divided from the next by the kit rule: the 72pt picture, then the title in the display face at heading size as a link-styled button with the kit focus ring, the date in mono at body size · the place in italic, and the role Badges at the row's trailing edge (Subject in the outline Badge). Below, indented to the title, the role captions and compact cards in a grid. The line runs from the birth cap (1817) to the burial (22 Sep 1880), four days past the death cap, at about 43pt a year; 1861 – 1880 is empty. "Undated" with its hint and two rows sits just before the death row. 12d: the list fills the pane; seven or eight cards a row; titles keep their measure.
- **12w Unknown ends.** "Birth unknown" (muted italic) with a dashed "?" cap 40pt before the baptism; "Death unknown" after the 1861 marriage and before the burial.
- **12x Short life.** Ernst, 1845 – 1848: the same line height over three years; year labels with quarter ticks; the baptism "Dec 1845" a month bar.
- **12b Scrolled.** The end of Johann's list: undated, death, burial; the rows in view light their points.
- **12f Row sheet.** Picture and Event mark; one and two role Badges; no, one and five people; a two-line title; a cut place; no date; no place; a ref title in mono; hover and focus on a title and a card. Compact card: role caption above, a 40pt likeness with the gender tint, the name on one line, the years in mono.
- **12g / 12h / 12k Large event.** 15 cards and "Show 26 more", ordered widow, children, officiant, mourners and pallbearers A–Z; unfolded; at Wide.
- **12j Long life.** Lise, 98 years: 10-year ticks, a point or bar for each of 120 events, the folded ones too; residences as range bars; her burial "bef. 1900" an open end past her death.
- **12l Only undated.** Birth unknown, Undated with two rows, Death unknown; the line is the two placeholder caps joined by a dashed stretch.
- **12m No events.** The EmptyState; no line. **12n / 12o / 12p** First load spinner; the error Callout; refreshing with the list and line drawn.
- **12q Keyboard.** Focus ring on Anna Weber's card; the Tab order and VoiceOver notes of AC-2.5 and TL-12.
- **12r After clicking Anna.** Crumbs *Persons › Weber, Anna › Timeline*, the Back menu open on "Johann Weber · Timeline"; her line 1790 – 1851.
- **12s / 12y / 12u / 12v / 12t Dark.** 12a, 12w, 12m, 12d and 12f in dark.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Section host | snowflake | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` *(new)* | D-12 |
| Timeline section | snowflake | `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)* | The line column and the list in an `HStack(alignment: .top, spacing: 0)` inside a `ScrollViewReader`; `#Preview`s for 12a, 12w, 12x, 12f, 12g, 12j, 12l from D-13 |
| Content, lifeline | pure | `PersonTimelineContent.swift`, `PersonTimelineLifeline.swift` *(new)* | D-6, D-15, D-16 |
| Line drawing | `Path` | `PersonTimelineSection.swift` | `PVColor.borderStrong` for the line and ticks, `PVColor.accent` for points in view and tethers, `PVColor.textMuted` for tick labels and placeholders, `StrokeStyle(dash: [3, 3])` for placeholder stretches and undated dots |
| Event row | snowflake | `PersonTimelineSection.swift` | D-7, D-10 (title), `EventSecondaryLine`, `PVBadge` (D-5); rows divided by `PVDivider` |
| Cards | `PersonTile(size: .compact)` | — | D-4, D-11 |
| Show more | `ConclusionShowMoreRow` | — | D-9 |
| Undated heading, placeholders | text | `PersonTimelineSection.swift` | D-16 |
| Width | `PVPageColumns(primaryFill:)` | `ConclusionDetailPage.swift` | D-17 |

Ids: D-14. Accessibility: each row `.accessibilityElement(children: .contain)` labelled `row.spokenLabel`; the title a button; cards buttons with the tile's spoken label; markers read in their place ("Born 12 Mar 1817, Bremen", "Death unknown"); the line hidden.

### Strings

L10n keys *(new)* under `L10n.Conclusions`: `timelineEmptyTitle` = "No events yet"; `timelineEmptyMessage` = "Events appear here when events this person took part in are promoted from an Evidence graph."; `timelineErrorTitle` = "Timeline could not be read"; `timelineUndated` = "Undated"; `timelineUndatedHint` = "No date, so no place on the line"; `timelineBirthUnknown` = "Birth unknown"; `timelineDeathUnknown` = "Death unknown"; `timelineAsRole(_ roles: String)` = "as %@"; `lifeYearsRange(_ born: String, _ died: String)` = "%1$@ – %2$@", `lifeYearsBorn(_ year: String)` = "b. %@", `lifeYearsDied(_ year: String)` = "d. %@" (the compact card's years). Renamed: `familyShowMore` → `showMore` (D-9). Each key is added in the step that first uses it, with its `Localizable.xcstrings` entry.

## 7. Steps

**STEP-0 — Check S13-33's names**
Files: this plan.
Do: read S13-33's **Landed** note; if any name in § 4's Timeline data row differs, update § 4, D-6, D-15 and D-16 before code.
Check: the names match `macos/App/Platform/GenealogyStore.swift`.

**STEP-1 — Shared pieces**
Files: `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` *(new)*, `macos/App/DesignSystem/README.md`, `macos/App/Platform/PropertyTermDisplay.swift`, `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/App/Features/Workspace/WorkspaceLocation.swift` (`fillsWidth`, all false), `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Features/Conclusions/PersonTile.swift` (Return), `macos/ProvenenciaTests/PropertyTermDisplayTests.swift`, `PVPageColumnsTests.swift`. Skill: `add-ui-component`.
Do: D-10 (`PersonTile` moves onto it), D-5's overload, D-17 with no section opting in.
Check: T-1, T-10 pass; Family's tiles still open on Return; every page looks as before.

**STEP-2 — One fold row and one section host**
Files: `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` *(new)*, `ConclusionPageSectionHost.swift` *(new)*, `PersonFamilySection.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`, `add-workspace-place`.
Do: D-9, D-12 with Family as the only caller.
Check: `PersonFamilyContentTests`, `PersonFamilyLayoutTests` pass; Family's states keep their ids and look (its previews); `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — The compact tile**
Files: `macos/App/Features/Conclusions/PersonTile.swift`, `PersonLifeDisplay.swift`, `PersonFamilyLayout.swift`, `macos/ProvenenciaTests/PersonTileTests.swift`, L10n files.
Do: D-4.
Check: T-2 passes; `PersonTileTests` and Family's previews unchanged at the regular size.

**STEP-4 — The Timeline section's key**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift` (`sectionQueryKeys`: `.timeline` leaves the nil list), `PlaceRegistryTests.swift`.
Do: D-3.
Check: T-3, T-4 pass.

**STEP-5 — Content, lifeline geometry and fixtures**
Files: `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)*, `PersonTimelineLifeline.swift` *(new)*, `PersonTimelinePreviewData.swift` *(new)*, `EventsListView.swift` (`includesRef:`), `macos/ProvenenciaTests/PersonTimelineContentTests.swift` *(new)*, `PersonTimelineLifelineTests.swift` *(new)*, L10n files. Skill: `add-swift-test`.
Do: D-1, D-2, D-5 – D-8, D-13, D-15's geometry, D-16's placement.
Check: T-5 – T-9, T-11 pass; `PersonsListTests` and the Events list tests pass unchanged.

**STEP-6 — The section view**
Files: `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)*, L10n files. Skill: `add-ui-component`.
Do: D-7, D-11, D-12's second caller, D-14, D-15's drawing, tethers and clicks, D-16; the previews.
Check: the app builds; each preview reads as § 6's frame.

**STEP-7 — Wiring**
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`, `WorkspaceLocation.swift` (`fillsWidth` true for `.timeline`).
Do: `.timeline` draws `PersonTimelineSection(entityId:)`.
Check: `scripts/check-pr.sh` passes; the manual check (§ 10).

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `design/S13-D4-timeline-section.md` → `design/archive/`, `design/README.md`, `docs/macos-client-patterns.md` (§ 9: the opt-in full-width primary; the sectioned page: a section on `ConclusionPageSectionHost`), `docs/ideas/structural-debt.md` (the `ReconciledValueDisplay.string(for:)` row).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PropertyTermDisplayTests.swift` `nameOfAValue` *(new)* | `.term` witness → "Witness"; custom "Godparent" → "Godparent"; empty label, key `godparent` → "godparent"; `.text` → nil | D-5 |
| T-2 | Swift | `macos/ProvenenciaTests/PersonTileTests.swift` `compactYears` *(new)* | Anna 1790 – 1851 → "1790 – 1851"; birth only → "b. 1820"; death only → "d. 1851"; neither → nil; `Size.compact.height` 52, likeness 40; regular unchanged | AC-2.4, AC-2.6, D-4 |
| T-3 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys` (extended) | `.timeline` → `.personEvents`; `.timeline.fillsWidth` true, every other section false | D-3, D-17 |
| T-4 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `personDetailOnTimelineWarmsTheTimeline` *(new)* | A Person location on `.timeline` → `.conclusionDetail` and `.personEvents` | AC-4.2, D-3 |
| T-5 | Swift | `macos/ProvenenciaTests/PersonTimelineContentTests.swift` *(new)* `rowsFromTimeline` | `johann` → `dated` 1 – 11 and the burial, `undated` 12 and 13; event 4's title, line, Badge Subject (outline), cards Maria and Friedrich with captions; event 1's picture set, 11's empty; event 13's line empty | AC-1.1, AC-1.3, AC-1.5, AC-1.6, D-6, D-7 |
| T-6 | Swift | same file `foldsAndOrder` | `lise` → 50 dated rows, `datedHidden` 70, expanded → 120; `largeEvent` → 15, `participantsHidden` 26, expanded → 41; input order kept, no sort; one person on two events → distinct ids | AC-3.1, AC-3.2, D-8 |
| T-7 | Swift | same file `rolesLocationsAndSpoken` | Event 8 → Badges Witness, Informant; Anna's card spoken "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851"; card location `…withPageSection(.timeline)`; row location `EventsList.location`; event 1's `spokenLabel` "Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject" with no ref; `EventsList.accessibilityLabel(_:orders:)` unchanged | AC-1.4, AC-2.1, AC-2.3, AC-2.4, D-1, D-2, D-5, D-6 |
| T-8 | Swift | `macos/ProvenenciaTests/PersonTimelineLifelineTests.swift` *(new)* `pointsToScale` | `johann`, height 2730 → birth at 0, the burial at the end; events 4 and 5 0.67 years apart; a day a dot, "1835" a bar, a range a bar, "bef. 1900" an open start; ticks every 2 years at about 43pt a year; `lise` → 10-year ticks | AC-5.1, AC-5.2, AC-5.3, D-15 |
| T-9 | Swift | same file `placeholdersAndShortLives` | `unknownEnds` → a 40pt stub before the baptism and one between the marriage and the burial, outside the scale; a funeral and no death → the death placeholder before the funeral; `ida` → two caps joined; `ernst` → year labels and quarter ticks; a range under a year → drawn as a year; a taller list (unfolded) → a new scale | AC-1.9, AC-5.4, AC-5.5, AC-5.6, D-15, D-16 |
| T-10 | Swift | `macos/ProvenenciaTests/PVPageColumnsTests.swift` (case added) | Pane 2296 `.wide` `primaryFill` → primary 2296 − 2 · gutterPage, no secondary; `.regular` → unchanged; `primaryFill` false → 1180 | AC-6.1, AC-6.2, D-17 |
| T-11 | Swift | `PersonTimelineLifelineTests.swift` `oneHundredTwentyEvents` | 120 spans and 120 row frames → 120 points in row order, built in one pass | AC-3.2, D-15 |

AC-1.2 and the participant order are S13-33's tests. AC-1.7, AC-1.8, AC-2.2, AC-2.5, AC-3.3, AC-4.1, AC-5.7, AC-5.8 and AC-6.3 are covered by the manual check (§ 10). Most likely to break: `PersonTileTests`, `PersonFamilyContentTests`, `PersonFamilyLayoutTests` (STEP-1 – STEP-3), `PlaceRegistryTests` (STEP-4), the Events list tests (STEP-5).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `PersonFamilyMoreRow`, `PersonFamilyLoading`, `familyShowMore`, `conclusions.family.showMore` and `PersonTile.height` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/` and `L10n.swift` with no hit left outside spike-13 plan history; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory and frames 12a – 12y; no new kit primitive but `pvActivatesOnReturn` and `primaryFill`
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] docs/macos-client-patterns.md: the full-width mode and the section host (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Bar 4 and the slice's Check. Use a project with the brief § 5 sample promoted (or the product owner's own test people).

1. (AC-1.1, AC-1.3, AC-1.5, AC-1.6, AC-5.1, AC-5.2, AC-5.3, AC-5.8) Open Johann, choose Timeline: frame 12a as § 6 reads it. Scroll: the tethers follow the rows in view.
2. (AC-1.2, AC-1.4) Luise's birth is not listed. Event 8 shows Witness and Informant.
3. (AC-2.1, AC-2.2, AC-2.4, AC-2.5, AC-2.6) Hover Anna's card: the tooltip with her full lines. Click it: her Timeline, crumbs *Persons › Weber, Anna › Timeline*; Back returns. Tab from event 1's title through its cards; Return opens; VoiceOver reads AC-2.4.
4. (AC-2.3, AC-5.7) Click event 4's title: the marriage's Event page. Back; click a point on the line: the list scrolls to its row.
5. (AC-3.1, AC-3.2, AC-3.3) The "Large event" and "Long life" previews fold and unfold; a click on a folded point in "Long life" unfolds and scrolls. In the app, unfold a list, open another person, come Back: folded.
6. (AC-4.1) With Johann's Timeline showing, edit Maria Schulz through her Evidence graph and come Back: the change shows; "Refreshing" in the switcher row meanwhile.
7. (AC-1.7, AC-1.8, AC-1.9, AC-5.4, AC-5.5, AC-5.6) A person with no events; a history entry for a person merged since; the `ida`, `unknownEnds` and `ernst` previews (frames 12l, 12w, 12x).
8. (AC-6.1, AC-6.2, AC-6.3) Widen past 1600pt on Timeline: the list fills the pane, the header stays at its measure; switch to Vitals: capped again; resize back: section, scroll, focus and unfolded lists kept. Dark mode: frames 12s – 12v.
9. Family looks and behaves as before (tiles, Return, "Show N more", loading, error and empty states).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S13-33 lands with other names | STEP-0 |
| Measuring every row's frame and drawing tethers per scroll is slow at 50 rows | Row frames update only on layout change; tethers only for rows in the viewport; T-11; the "Long life" preview scrolls smoothly in the manual check |
| The line's scale follows the list's height, so unfolding a list rescales it | Intended (the board's TL-5); T-9 |
| `primaryFill` changes another page | Off by default; only `.timeline` sets it (T-3, T-10) |

**Stop and report if:**

- S13-33 has not landed, or `CatalogPersonTimeline`, `CatalogTimelineSpan` or `CatalogLifeMarker` differ from § 4 in a way STEP-0 cannot map.
- `PVPageColumns` cannot widen the primary without widening its header band (D-17).
- The compact tile changes the regular tile's look or tests (STEP-3).
- A frame needs data the read does not return (an age, an estimated year): a follow-up, not Swift logic.

## 12. Open questions

None. The board's open items were answered by the user on 2026-11-02 (compact cards, participant order by role then name in Go, full width at Wide); the 72pt picture is kept as drawn (D-7).

## 13. PR

**Title:** S13-13 — Timeline section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md`.
