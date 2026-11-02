# S13-13 — Timeline section

**Status:** Blocked — on S13-33, S13-35
**Spike / slice:** Spike 13 · Slice 4 — Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D4](../design/S13-D4-timeline-section.md), board drawn 2026-11-02 (turn 12) and agreed with the user's answers of 2026-11-02 (compact cards, participant order by role, full width at Wide)
**Board:** Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Timeline.dc.html*, turn 12 "S13-D4 · Timeline — the event list and the lifeline (update 2026-11-02)", frames 12a – 12y. The implementer may not be able to open it: **Design → macOS** describes every frame in words and gives every measure
**Depends on:** S13-05 (landed [#680](https://github.com/mendahu/provenencia/pull/680)), S13-12 (landed [#676](https://github.com/mendahu/provenencia/pull/676)), S13-33 ([plan](S13-33-timeline-positions.md), Planned: each event's span, the birth and death markers, participants ordered by role then name), S13-34 ([plan](S13-34-display-pieces-move-down.md): title styling and the date · place line as kit components, value formatting in `Platform/`, one term-name rule, year precision), S13-35 ([plan](S13-35-person-card-joins-the-design-system.md): the card shell, Return activation, the focus-ring style, the person card in regular and compact sizes)
**Blocks:** —
**Branch:** `s13-13-timeline-section`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go or proto change (S13-33 has them)
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-05 and S13-12 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md) (STEP-7: its Paste block's Wide rule 2)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 3. Timeline](../person-detail.md#3-timeline) (R-2, R-3, R-23 – R-28, R-24a); brief [S13-D4](../design/S13-D4-timeline-section.md); the S13-05, S13-12, S13-33, S13-34 and S13-35 plans ([`S13-05-person-timeline-read.md`](S13-05-person-timeline-read.md), [`S13-12-family-section.md`](S13-12-family-section.md), [`S13-33-timeline-positions.md`](S13-33-timeline-positions.md), [`S13-34-display-pieces-move-down.md`](S13-34-display-pieces-move-down.md), [`S13-35-person-card-joins-the-design-system.md`](S13-35-person-card-joins-the-design-system.md)) and their **Landed** notes; working rules: "Engine PRs run ahead of their slice" (data, not layout), "Each section reads its own query", "Timeline is direct participation, any role", "The person card lives in the design system; display is decoupled from content", "Colors reuse `PersonGenderTreatment`", "A timeline tile opens the participant on Timeline", "Timeline lists fold", "The timeline is a lifeline", "The Timeline shows each event's picture", "Sections are places"
**Planned:** 2026-11-02 against `main` @ `51dccd4`; refreshed the same day after the board (turn 12), the descoping of S13-32, the S13-33 plan, review round 2, and the re-cut of the card work into S13-34 and S13-35 (base `04a2c99`, the prep plans' base). STEP-0 re-checks S13-33's, S13-34's and S13-35's landed names
**Review:** [`S13-13-timeline-section.review.md`](S13-13-timeline-section.review.md) — round 1 Fits with changes / Ready with changes; round 2 Fits with changes / Not ready, every finding Fixed or answered by the user; scope change after round 2 (the card pieces moved to S13-34 and S13-35), not yet re-reviewed

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

A researcher opens a Person's Timeline section and reads the person's life on a lifeline drawn to scale from birth to death beside a list of events, earliest first. Each dated event points to its place on the line, so the stretches with no records show as empty line. Each row shows the event's picture, title, date and place, and the person's role; the other people who took part follow as compact person cards, ordered by role. Birth and death always have a marker, a placeholder when unrecorded; undated events sit just before the death marker. A click on an event opens its Event page; a click on a card opens that person's Timeline. It fills the empty place S13-10 left (S13-D1 PPF-10) with the data S13-05 and S13-33 read, on the person card and card shell S13-35 puts in the design system.

**Why now.** Slice 4's Check is bar 4 ("A life in order"); S13-05 and S13-33 give the read, S13-34 and S13-35 give the card and the display pieces, the board is drawn, and this is the last PR of the slice.

**What changes beyond the Timeline.** Family's look is unchanged; its host and fold row move into shared pieces (D-8, D-10). Page columns gains an opt-in full-width mode and publishes its visible rect (D-13, D-15). The term-name change in Vitals and the card's move into the design system are S13-34's and S13-35's.

**What stays.** The header, the switcher, Vitals, Gallery and the Sources empty place; the regular person card (S13-35); `ListPersonEvents` and its order (S13-05, S13-33); the Events list and Event page; every other link into a Person page; every other page on Page columns.

## 2. User stories

**US-1.** As a researcher, when I open a Person's Timeline, I see every event they took part in, in date order, so that I read their life from birth to death.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann (brief § 5 sample, events 1 – 13; death 18 Sep 1880, burial 22 Sep 1880) | Items in order: event 1 (his birth, the birth marker), events 2 – 9, "Undated" with 12 and 13, event 10 (the death marker), event 11 (the burial, past death) (frame 12a; R-23, R-24; D-5) |
| AC-1.2 | Luise's birth names no father | It is not listed (R-23; S13-05 filters it) |
| AC-1.3 | Event 4 | Title "Marriage of Johann Weber and Maria Schulz", "4 Jun 1844 · Bremen", Badge "Subject" (outline), cards Maria Schulz (caption "Subject") and Friedrich Weber ("Witness") (R-25, R-27; D-4, D-5) |
| AC-1.4 | Event 8: Johann is Witness and Informant | Two neutral Badges, "Witness" then "Informant" (D-4) |
| AC-1.5 | Event 1 has a lead picture; event 11 has none | Event 1 shows its picture at 72pt; event 11 shows the Event mark on the sunken tone (R-26; D-6) |
| AC-1.6 | Event 13: no date, no place | The date and place line is not drawn |
| AC-1.7 | A person with no promoted events | The EmptyState of TL-10, no line (frame 12m) |
| AC-1.8 | The read fails with `conclusiondetails.not_found` | A danger Callout in the section titled "Timeline could not be read" with the mapped message; the header and switcher still show (frame 12o) |
| AC-1.9 | Only undated events (Ida, frame 12l) | Items: "Birth unknown", "Undated" with its two rows, "Death unknown"; the line is the two placeholder caps joined by a dashed stretch (D-5, D-13) |
| AC-1.10 | An event whose roles are all empty | No Badges; its spoken label has no "as …" clause (D-4, D-5) |

**US-2.** As a researcher, I can move from an event to its page and from a participant to their life.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | On Johann's Timeline, click Anna Weber's card on event 1 | Anna's Person page opens on Timeline; breadcrumbs *Persons › Weber, Anna › Timeline*; Back lists "Johann Weber · Timeline" (frame 12r; D-1) |
| AC-2.2 | Then Back | Johann's Timeline again |
| AC-2.3 | Click event 4's title | The Event page of that marriage opens, titled as the Events list titles it (D-2) |
| AC-2.4 | VoiceOver on Anna's card under event 1 | "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851", a button: the card's spoken label with the role (`PVPersonCardContent`, S13-35; D-5) |
| AC-2.5 | Tab from event 1's title | Focus moves to its cards in order, then any "Show N more", then event 2's title; Return or Space activates (frame 12q; D-16, S13-35's `pvActivatesOnReturn`) |
| AC-2.6 | Hover Anna's compact card | The tooltip reads "Anna Weber", "b. 3 May 1790 · Bremen", "d. 1851" on three lines (S13-35's `helpText`, tested there; manual here) |

**US-3.** As a researcher, long lives and crowded events stay readable without hiding anything.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | An event with 41 other participants | 15 cards, then "Show 26 more"; pressing it shows all 41 (frames 12g, 12h; D-7). Evidence: T-4, and the `#Preview` "Large event" (D-11) |
| AC-3.2 | Lise: 120 dated events, the death the 118th | The first 50 rows, "Show 67 more", then "Undated", the death row and the two rows after it; the line still shows all 120 points; a click on a folded event's point unfolds the list and scrolls to its row (frame 12j; D-5, D-7, D-13) |
| AC-3.3 | Unfold a list, then open another person and come Back | The list is folded again (D-7, as Family) |

**US-4.** As a researcher, the section stays current while I work.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Johann's Timeline is showing; a write touches Maria Schulz's handle | The timeline reloads; the list and line stay drawn and the switcher row shows "Refreshing" until it lands (frame 12p; D-3) |
| AC-4.2 | Opening a Person straight on Timeline (from a card, or Back) | The timeline read starts with the page, not after the detail read (D-3) |

**US-5.** As a researcher, I see the shape of the life at a glance: where its records cluster and where they fall silent (R-24a).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | Johann: birth 12 Mar 1817, death 18 Sep 1880 | A vertical line in an 80pt column beside the list, from the birth cap at the top to the death cap, then on to the burial four days later; each dated event has a point; nothing marks 1861 – 1880 but empty line (frame 12a; D-13) |
| AC-5.2 | Event 4 on 4 Jun 1844 and event 5 on 2 Feb 1845 | Their points are 0.67 years apart at the line's scale (D-13) |
| AC-5.3 | A day; "1835"; "1844 – 1849"; "bef. 1900" | A dot; a bar over 1835; a bar from 1844 to 1850; a bar over 1900 with an open (fading) start (D-13) |
| AC-5.4 | Johann without his birth and death events (frame 12w) | A dashed "?" cap and the muted item "Birth unknown" 40pt before the baptism, the stretch dashed and not to scale; "Death unknown" after the 1861 marriage and before the burial; no year shown (D-5, D-13) |
| AC-5.5 | A funeral in 1880, a census in 1875, no death | "Death unknown" sits after the census and before the funeral; the funeral sits at its own date (D-5) |
| AC-5.6 | Ernst, 1845 – 1848, three events (frame 12x) | The line spans the same list height over three years: year labels with quarter ticks (D-13) |
| AC-5.7 | Click a point or a cap on the line | The list scrolls the item to the top of the visible area; a second click on the same point scrolls again (D-13) |
| AC-5.8 | Scroll the list | Each row in the visible area has a short curve from its point to the dot at the row's edge, in the accent; rows out of view have none; an undated row has a dashed dot and no curve (frames 12a, 12b; D-13) |

**US-6.** As a researcher on a wide window, I see more people per event (TL-W1).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-6.1 | Johann's Timeline at Wide (pane 2296) | The section fills the pane between the page gutters; the line keeps its 80pt column; titles and date lines keep their Regular measure; each event's cards fill the row at a 232pt minimum, eight a row (frame 12d; D-9, D-15) |
| AC-6.2 | The header and switcher at Wide | Unchanged: they keep the 1180pt band (D-15) |
| AC-6.3 | Resize across 1600pt | Section, scroll position, focus and unfolded lists kept (TL-W3) |

The former US-7 (one term name in Vitals, on the badge and on the Timeline) and its two criteria moved to S13-34 (its AC-1.1 and AC-1.2).

## 3. Scope

| In | Covered by |
| --- | --- |
| Page columns' full-width primary, its band and its visible rect (D-13, D-15) | STEP-1, AC-6.1, AC-6.2 |
| One "Show N more" row and one host for a section's own read; Family moves onto both, unchanged (D-8, D-10) | STEP-2 |
| The Timeline section's query key and its full-width flag (D-3, D-15) | STEP-3, AC-4.1, AC-4.2 |
| The section's items, lifeline geometry and fixtures (D-1, D-2, D-4 – D-7, D-11, D-13) | STEP-4, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.9, AC-1.10, AC-2.1, AC-2.3, AC-2.4, AC-3.1, AC-3.2, AC-5.1 – AC-5.6 |
| The section view: event rows on the card shell, compact cards, the line, its states (TL-1 – TL-13, TL-W1 – TL-W3; D-6, D-9, D-12, D-13, D-14, D-16) | STEP-5, STEP-6, AC-1.7, AC-1.8, AC-2.5, AC-2.6, AC-5.7, AC-5.8, AC-6.1 |

| Out | Home |
| --- | --- |
| Positions, markers and participant order in Go | S13-33 |
| Title styling, the date · place line, value formatting in `Platform/`, the term-name rule, year precision | S13-34 |
| The card shell, Return activation, the focus-ring style, the gender treatment and the person card in regular and compact sizes | S13-35 |
| Zoom, pan, a draggable canvas or a swimlane | not planned (brief § 2.1) |
| Gap labels; an estimated year for an unknown birth or death; marks on events outside the life | not planned (user, 2026-11-02) |
| Age at each event | not planned this spike (brief § 2.1) |
| Filtering or searching the timeline | not planned |
| The provenance control on a row, role or card | not planned this spike (R-13b) |
| Editing events or participations from the section | not planned this spike |
| Full width for Vitals, Family or Gallery | not planned; D-15's mode is opt-in per section (Gallery may adopt it later, S13-D5) |
| An event card recipe (`PVEventCard`) | not planned: the event row has one caller; the evaluation's follow-up 6 when a second caller comes |
| `EventSubject` folding into {role, `PersonSummary`} | logged by S13-05 in [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Timeline section | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` | `.timeline, .sources` draw one `PVEmptyState` with `L10n.Conclusions.pageSectionLater`. `.family` draws `PersonFamilySection(entityId:)` |
| Section key | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection` (`String` raw values `vitals`, `family`, `timeline`, …), `queryKey(project:entityId:)`, `glyph` | `.family → .personFamily`; every other section nil. Read by `PlaceRegistry` Spec `.personDetail` and `ConclusionDetailLoaded.sectionRefreshing` (`ConclusionDetailPage.swift`) |
| Timeline key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.personEvents(project:entityId:)`; `CatalogQueryRegistry.swift` its descriptor | `.sessionFresh`, `.evict` when hidden, refreshes on any touched handle. Value today `[CatalogPersonEvent]`; S13-33 makes it `CatalogPersonTimeline` |
| Timeline data (after S13-33) | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonTimeline { events; birth; death }`, `CatalogPersonEvent.span: CatalogTimelineSpan?` (`from`, `to` on the year line, `openStart`, `openEnd`) with a computed `dated` (`span != nil`), `CatalogLifeMarker { eventId: String?; besideEventId: String? }` | Rows in date order (an open-start date at its known end), undated last; a recorded marker is exactly the header's birth or death when that event is a row, else a placeholder; `besideEventId` set whenever the marker has no span (S13-33 D-5, D-6); participants ordered by role rank, then name (S13-33 D-10); roles are `.term` values with labels filled in Go |
| Card and shell (after S13-35) | `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` *(new)* (S13-35) `PVPersonCard(content:projectDir:size: PVTileCardSize, open:)`, `PVPersonCardContent(summary:role:isFocus:locale:orders:)` (`spokenLabel`, `years`, `helpText`); `macos/App/DesignSystem/Components/TileCard/PVTileCard.swift` *(new)* (S13-35) `PVTileCard(size:isRaised:accessibilityLabel:help:action:thumbnail:title:lines:trailing:)`, `PVTileCardSize` (`.regular`, `.compact`; `height`, `thumbnail`, `titleLineLimit`, `linesLineLimit`, `minimumWidth`, `padding`); `macos/App/DesignSystem/Components/ReturnActivation/` *(new)* (S13-35) `pvActivatesOnReturn(_:)`, `pvActivatesOnReturn(when:_:)`; `macos/App/DesignSystem/Components/FocusRingButton/` *(new)* (S13-35) `PVFocusRingButtonStyle(cornerRadius:)` | Not landed; today the card is `macos/App/Features/Conclusions/PersonTile.swift` `PersonTile`, one size. STEP-0 checks the names |
| Display pieces (after S13-34) | `macos/App/DesignSystem/Components/TitleText/` *(new)* (S13-34) `PVTitleSource`, `PVTitleText(_:style:extraCount:lineLimit:emphasized:)`; `macos/App/DesignSystem/Components/DatePlaceLine/` *(new)* (S13-34) `PVDatePlaceLine(leading:date:place:accessibilityLabel:missing:dateSize:extraCount:)` (the Events list's line is a call to it after S13-34); `macos/App/Platform/PropertyTermDisplay.swift` `name(_ value: CatalogConclusionValue, propertyKey:) -> String?`; `Platform/` `PersonTitleDisplay`, `PersonLifeDisplay`, `DateRowDisplay` | Not landed; today `ConclusionTitleSource`, `ConclusionListRow.title`, `ConclusionDatePlaceLine` in `Features/Conclusions`, and no `CatalogConclusionValue` overload. STEP-0 checks the names |
| Show more | `macos/App/Features/Conclusions/PersonFamilySection.swift` private `PersonFamilyMoreRow` (on `pvActivatesOnReturn` after S13-35); `L10n.Conclusions.familyShowMore(_:)` (key `conclusions.family.showMore`) | A ghost `PVButton` (`.pv(.ghost, size: .sm)`); Family only |
| Folds | `macos/App/Features/Conclusions/PersonFamilyContent.swift` `foldLimit`, `Fold`; `PersonFamilySection` `@State expanded`, reset `.onChange(of: entityId)` | The pattern to copy |
| Section host | `macos/App/Features/Conclusions/PersonFamilySection.swift` `PersonFamilySection` (builds `.personFamily(project:entityId:)` by hand), private `PersonFamilyLoading`, `PersonFamilyLoaded` | Handle lookup, `.task`, 96pt spinner, danger Callout, EmptyState and ids around Family's diagram |
| Read states | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPresentation` | Generic over the read's value |
| Page columns | `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` `PVPageColumns` (three initializers: the main one, `where Secondary == EmptyView`, `where Header == EmptyView, Secondary == EmptyView`), `PVPageColumnsMetrics.columnWidths(pane:widthClass:hasSecondary:)` and `columnWidths(proposal:…)` returning one `group`, private `PVPageColumnsLayout` and `PVPageColumnsBandLayout` (both read `group`), private `@State offset` and `viewportHeight` measured in `PVPageColumnsMetrics.scrollSpace`; `macos/ProvenenciaTests/PVPageColumnsTests.swift` | The primary is `min(widthContentMax, pane − 2 · gutterPage)` at every width: 1180pt at Wide. `ConclusionDetailLoaded` uses the `Secondary == EmptyView` initializer |
| Scroll to a row | `macos/App/DesignSystem/Components/ScrollRowIntoView/PVScrollRowIntoView.swift` `pvScrollsRowIntoView(_:anchor:animated:…)`; `macos/App/DesignSystem/README.md` ("do not add another `ScrollViewReader`") | Scrolls when its row value changes |
| Event title, line, location, spoken | `macos/App/Features/Conclusions/EventTitleDisplay.swift` `titleSource(_:locale:orders:)`; `EventsListView.swift` `EventsList.location(_:orders:)`, `EventsList.accessibilityLabel(_:orders:)` (title, date, ref through `L10n.Workspace.eventRowAccessibility`, then place, `morePlaces`), `EventSecondaryDisplay.content(_:locale:)`, `EventSecondaryLine` (date in mono caption); tested in `macos/ProvenenciaTests/PersonsListTests.swift` | The Events row's builders |
| Fonts | `macos/App/DesignSystem/Tokens/PVTypography.swift` `PVFont.display(size:weight:)`, `PVTypeScale.h3` (18), `.body` (16), `.caption` (13) | — |
| Seeded keys | `macos/App/Platform/ConclusionVocabulary.swift` `SeededRoleTerm.subject`, `SeededPropertyKey.role`; `CatalogOrigin.isProduct` | — |
| Person link | `macos/App/Features/Conclusions/PersonsListView.swift` `PersonsList.location(_:orders:)`; `WorkspaceLocation.withPageSection(_:)` | As S13-12 D-1 |
| Thumbnail | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail`; mark `PVMarkKey.subjectEvent`; `L10n.Conclusions.eventPictureThumb` | The Events list's thumb and fallback |
| Divider | `macos/App/DesignSystem/Components/Divider/` `PVDivider` | The kit rule between rows |
| Preview data | `macos/App/Features/Conclusions/PersonFamilyPreviewData.swift` `PersonFamilyPreviewData` (`#if DEBUG`), `summary(_:_:born:)` | Shared by Family's previews and tests |
| README | `macos/App/DesignSystem/README.md` built table; `scripts/check-guidance-names.py` | One row per `Components/<Name>/` folder |
| Wide rule 2 | `docs/macos-client-patterns.md` § 9 rule 2; `.agents/skills/add-design-brief/SKILL.md` Paste block rule 2 ("change both together") | "Prose, forms and the primary column keep their Regular maximum widths" |
| Tests | `macos/ProvenenciaTests/PersonFamilyContentTests.swift`, `PersonFamilyLayoutTests.swift`, `PersonDetailTests.swift` (`sectionQueryKeys`), `PlaceRegistryTests.swift`, `CatalogQueryRegistryTests.swift`, `PVPageColumnsTests.swift`, `PersonsListTests.swift` | — |

**Precedent.** S13-12 ([#676](https://github.com/mendahu/provenencia/pull/676)) is the same shape: a section that reads its own key through `ConclusionPageSection.queryKey`, a pure content model with folds, a pure layout tested on its geometry (`PersonFamilyLayout`), a section view on `ConclusionDetailPresentation`, tiles that navigate with `navigation.go(to:)`, and preview data shared with the tests. Copy those shapes; Family's host moves into D-10 rather than being copied.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A participant's card opens `PersonsList.location(participant.person, orders:).withPageSection(.timeline)` with one `navigation.go(to:)` (user, 2026-11-02) | Working rule "A timeline tile opens the participant on Timeline" | Vitals (S13-D1 PPF-5); building `.personDetail(` by hand |
| D-2 | An event opens `EventsList.location(event, orders:)` with one `navigation.go(to:)` (R-28) | The Events list's location: same title and history entry | A hand-built `.eventDetail(` |
| D-3 | `ConclusionPageSection.queryKey` gains `.timeline → .personEvents(project:entityId:)`; `PlaceRegistry` warms it and the switcher's refresh follows it through S13-12's two readers | S13-12 D-8: one declaration per section key | A per-section branch in `PlaceRegistry` |
| D-4 | Role badges and captions. Each role is S13-34's `PropertyTermDisplay.name(value, propertyKey: SeededPropertyKey.role)`; nil roles are dropped. The person's roles show as Badges in the read's order: `PVBadge` neutral; outline when `isSubject` = `CatalogOrigin.isProduct(origin) && key == SeededRoleTerm.subject` (frame 12f). A participant's role caption and spoken role are its names joined with `L10n.DesignSystem.spokenJoin` ("Witness, Informant"); nil when it has none (no caption, no spoken role) | One rule turns a term into a name (S13-34); round 2 M-22 | A timeline-only role type; a bare key compare for Subject |
| D-5 | `PersonTimelineContent` *(new, pure, `Equatable`)* in `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)*: `init(timeline: CatalogPersonTimeline, expanded: Set<Fold> = [], locale:, orders:)` builds **`items: [Item]`**, the one display order the list and the line both read (round 2 M-12, X-15). `enum Item: Identifiable { case marker(Marker), row(Row), undatedHeader, more(Fold, count: Int) }`; ids: a row's and a recorded marker's are the event id, a placeholder's `"birth"` / `"death"`, the header `"undated"`, a more row `"more.dated"`, `"more.undated"`, `"more.<eventId>"`. `Marker` holds `kind` (`.birth`, `.death`), `row: Row?` (the recorded event's row; nil for a placeholder) and `spokenLabel` (a recorded marker reads its row's label, round 2 X-18; a placeholder its text). **Order:** (1) the dated rows in the read's order; (2) the birth marker: a recorded birth with a span is its own row in place; otherwise (a placeholder, or a recorded birth with no span) the marker goes immediately before the row `besideEventId` names, or first when nil; (3) the death marker: a recorded death with a span is its row in place; otherwise it goes immediately after the row `besideEventId` names, or last when nil; (4) the undated header and the undated rows (a recorded but undated birth or death leaves this list: it is its marker) go immediately before the death marker. **Folds** (D-7) apply to the dated rows before the death marker (a recorded birth's row counts among them; a birth placeholder is never folded): past the 50th, the rest of that stretch is replaced by one `more(.dated, count:)` item placed where the hidden rows were; the death marker, the undated group and every row after the death marker always show. The undated rows fold at 50 behind `more(.undated, count:)`. `Row`: `id`, `title: PVTitleSource` (`EventTitleDisplay.titleSource`), `line: EventSecondaryDisplay.Content`, `imageRelPath`, `roles: [RoleBadge]` (`text`, `isSubject`), `location` (D-2), `span: CatalogTimelineSpan?`, `participants: [Participant]`, `participantsHidden`, `spokenLabel` = `EventsList.spokenLine(event, orders:)` *(new)*, then, when the roles are not empty, `L10n.Conclusions.timelineAsRole(spokenJoin(roles))` joined with `spokenJoin` ("Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject"). `EventsList.spokenLine(_:orders:) -> String` *(new)* is `spokenJoin` of the title, the date, the place and `morePlaces` when present, no ref (round 2 M-19); `EventsList.accessibilityLabel` is unchanged. `Participant`: `id` ("\(eventId)-\(personId)"), `card: PVPersonCardContent` (S13-35; `role:` the spoken role, `isFocus: false`), `caption: String?`, `location` (D-1). `allSpans: [(id, span)]` gives every dated row's span, folded or not, for the line. `isEmpty` is `timeline.events.isEmpty` | One owner of the order; the folded-death and undated-death cases decided once and tested (T-3, T-4) | Two consumers deriving the order from markers; a Bool on `accessibilityLabel` |
| D-6 | The picture is `CachedThumbnail(projectDir:relPath: row.imageRelPath, size: 72, cornerRadius: PVRadius.sm, fallbackMark: .subjectEvent, fallbackLabel: L10n.Conclusions.eventPictureThumb, colors: nil)`, 72pt (frame 12f) | The board's 72pt (the event row is a Timeline snowflake; the shell has no row size, user 2026-11-02) | The Events list's 44pt |
| D-7 | Folds (working rule "Timeline lists fold"): `eventFoldLimit = 50`, `participantFoldLimit = 15`; `enum Fold: Hashable { case dated, undated, participants(eventId: String) }`. `expanded` lives in the section as `@State`, reset `.onChange(of: entityId)` | The user's caps; Family's fold shape | A shared fold type with Family |
| D-8 | One "Show N more" row: `PersonFamilyMoreRow` moves to `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` *(new)* as `ConclusionShowMoreRow(count:expand:)`, the same `.pv(.ghost, size: .sm)` button on S13-35's `pvActivatesOnReturn`. `L10n.Conclusions.familyShowMore(_:)` → `showMore(_:)`, key `conclusions.showMore`, comment "Button under a folded list on a Conclusion page that shows the rest; argument is how many are hidden"; the String Catalog entry moves with it | One row and one string for both sections | A private copy |
| D-9 | Each event's cards are `PVPersonCard(content: participant.card, projectDir:, size: PVTileCardSize.compact, open: { navigation.go(to: participant.location) })` (S13-35), laid out in `LazyVGrid(columns: [GridItem(.adaptive(minimum: PVTileCardSize.compact.minimumWidth), spacing: 8)], alignment: .leading, spacing: 8)`, starting at the title's leading edge, each `PVTileCardSize.compact.height` tall; above each card its caption in `.pvMicroCaps()` `PVColor.textMuted` (none when nil) (frame 12f) | The card's size lives on the card (S13-35; round 2 M-21); three or four a row at Regular, eight at Wide | Family's column formula; a timeline copy of the card |
| D-10 | One host for a page section's own read: `ConclusionPageSectionHost<Value, Content: Equatable, Body: View>` *(new)* in `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` *(new)*: `init(section: ConclusionPageSection, entityId: String, errorTitle: LocalizedStringResource, emptyTitle: LocalizedStringResource, emptyMessage: String, make: @escaping (Value) -> Content, isEmpty: @escaping (Content) -> Bool, @ViewBuilder body: @escaping (Content) -> Body)`. The key is `section.queryKey(project: session.projectKey, entityId:)` (a `precondition` that it is not nil), the glyph `section.glyph`, the id root `"persons.detail.\(section.rawValue)"` (round 2 M-17). It reads `@Environment(WorkspaceSession.self)`, gets the handle with `session.queryHandle(key) as QueryHandle<Value>?`, starts it in `.task { let _: QueryHandle<Value> = session.query(key) }`, and maps `ConclusionDetailPresentation<Content>`: `.loading` (or no handle) → `ProgressView()` 96pt from the top; `.failed` → `PVCallout(tone: .danger, title: errorTitle, message:)`, id `<root>.error`; `isEmpty` → `PVEmptyState`, id `<root>.empty`; else `body(content)`, id `<root>`. Family moves onto it (its hand-built key goes) with ids, copy and look unchanged; the Timeline is its second caller | Only the content differs (round 1 M-6); S13-16 is the third | Copying Family's host; passing the key and glyph separately |
| D-11 | Fixtures: `macos/App/Features/Conclusions/PersonTimelinePreviewData.swift` *(new)* (`#if DEBUG`, internal) `enum PersonTimelinePreviewData`, people from `PersonFamilyPreviewData.summary`, events from one `event(_ id:, type:, title:, date:, place:, picture:, roles:, participants:)` helper: `johann` (brief § 5, events 1 – 13, death 18 Sep 1880, burial 22 Sep 1880, pictures on 1, 2, 4, 5, 9, 10; spans as S13-33 gives them); `unknownEnds` (johann without 1 and 10; birth placeholder beside event 2, death placeholder beside event 9); `ernst` (birth 3 Apr 1845, baptism "Dec 1845", death 14 Feb 1848); `largeEvent` (one burial with 41 participants: 4 Child, 1 Witness, 1 Informant, 6 Pallbearer, 29 with the researcher's role "Mourner", in S13-33's order); `lise` (born 1801, died 1899: 120 dated events, the death the 118th, a residence range every few years, a burial "bef. 1900" and an undated two; the writer's choice of the rest, tests check only the named properties); `ida` (two undated events); `ludwig` (none). The section's `#Preview`s and the tests use it | One fixture for previews and tests (S13-12 Landed note (6)) | A test-only copy |
| D-12 | Accessibility ids under `persons.detail.timeline`: the list; `.line`; an item `.item.<Item.id>`; a title `.title.<eventId>`; a card `.tile.<Participant.id>`; D-10's `.error` and `.empty` | Unique ids from the item ids | One shared `.more` id |
| D-13 | The lifeline (frames 12a – 12x). **Geometry** is a pure function, `PersonTimelineLifeline` *(new)* in `macos/App/Features/Conclusions/PersonTimelineLifeline.swift` *(new)*: `init(content:, itemFrames: [String: CGRect], visible: CGRect)`, all in the section's own coordinate space (below). The line spans from the first item's top to the last item's bottom. **Scale:** the placed range is the earliest `span.from` to the latest `span.to` of `content.allSpans`; placeholder markers take a fixed 40pt dashed stub each, outside the scale; points per year = (line length − stubs) ÷ years; a range under one year is drawn as one year. **Ticks:** the smallest step in 1, 2, 5, 10, 20, 50 years with ticks at least 34pt apart, labelled with the year in mono micro; if even the 50-year step gives more than 200 ticks, the step grows by powers of ten (100, 200, 500, 1000, …) until it gives at most 200 (the engine bounds years to `datevalue.MinYear…MaxYear`, S13-33, so a placed range is at most 19999 years (-9999.0 to 10000.0)); quarter ticks when a year is at least 136pt. **Marks:** a span of at most 1/365 year is a dot; else a bar; `openStart` / `openEnd` fades that end; birth and death are caps (a recorded marker at its row's point; a placeholder a dashed "?" cap on its stub). **Tethers:** for each row item whose frame intersects `visible`, a curve from its point to a dot at the row's leading edge in `PVColor.accent`; undated rows get a dashed dot and no curve. **Coordinates** (round 2 M-13, X-13): `PVPageColumns` publishes its visible rect as `EnvironmentValues.pvPageVisibleRect: CGRect` *(new)* in `PVPageColumnsMetrics.scrollSpace` (`CGRect(x: 0, y: offset, width: proposed width, height: viewportHeight)`, from the state it already measures; `.zero` outside a page). The section measures its own frame in `scrollSpace` once and each item's frame in its own named space `"persons.detail.timeline"` with `onGeometryChange`; `visible` is the published rect shifted by the section's origin. **Clicks** (X-14): a point or cap sets `@State scrollTarget: String?` to its item id, and the list applies `.pvScrollsRowIntoView(scrollTarget, anchor: .top)` inside `PVPageColumns`' scroll view; a click first sets `scrollTarget = nil`, then the id on the next main-actor turn (`Task { @MainActor in scrollTarget = id }`), so a repeat click scrolls again; a folded row's point first inserts `.dated` into `expanded`, then sets the target on the next turn, after the row exists. The line, ticks, points and tethers are `.accessibilityHidden(true)`. Swift does no date arithmetic | Working rule "The timeline is a lifeline"; one scroll-to-row mechanism and one owner of the page's scroll geometry | A second `ScrollViewReader`; measuring the scroll offset again in the section; fixed points per year |
| D-14 | Placeholder items read "Birth unknown" / "Death unknown" in muted italic body text, ids `.item.birth` / `.item.death`; the undated header is "Undated" (`.isHeader`) with the hint "No date, so no place on the line" in caption muted (frames 12a, 12l) | The board's TL-3 and TL-5a | An estimated year |
| D-15 | Full width at Wide (user, 2026-11-02). `PVPageColumnsMetrics.columnWidths(pane:widthClass:hasSecondary:primaryFill:)` *(new parameter, default false)* returns `(primary, secondary, group, band)` *(new member)*: unchanged when `primaryFill` is false (`band == group`); when true, `widthClass == .wide` and no secondary, `primary = group = pane − 2 · gutterPage` and `band = min(widthContentMax, primary)`; `primaryFill` with a secondary is ignored. The `proposal:` overload passes it through; with no finite proposal it is ignored. `PVPageColumnsLayout` centres on `group`; `PVPageColumnsBandLayout` sizes the header from `band` (round 2 M-16, X-21). All three `PVPageColumns` initializers gain `primaryFill: Bool = false`. `ConclusionPageSection.fillsWidth: Bool` *(new)*, true for `.timeline` only; `ConclusionDetailLoaded` passes `primaryFill: pageSection?.fillsWidth ?? false`. Inside the Timeline, titles and date lines keep `.frame(maxWidth: PVSpacing.widthContentMax − 80, alignment: .leading)`; the card grids fill | The board's TL-W1 and the user's answer; one owner of every width | A Timeline-only frame override; one `group` for both the columns and the band |
| D-16 | The event row is a Timeline snowflake (user, 2026-11-02: no row size on the shell), private `PersonTimelineEventRow` *(new)* in `PersonTimelineSection.swift`: an `HStack(alignment: .top, spacing: 16)` of D-6's picture, a `VStack(alignment: .leading, spacing: 4)` of the title button and `PVDatePlaceLine(date:place:accessibilityLabel:missing: .omitted, dateSize: PVTypeScale.body, extraCount: row.line.extraPlaces)` (S13-34), and the role Badges at the trailing edge; no card surface; the row is `.accessibilityElement(children: .contain)` labelled `row.spokenLabel`, so the title button stays reachable. The title button (round 2 M-18, X-20) is `Button { PVTitleText(row.title, lineLimit: 2).font(PVFont.display(size: PVTypeScale.h3)).foregroundStyle(PVColor.textPrimary) } action: { navigation.go(to: row.location) }` with `.buttonStyle(PVFocusRingButtonStyle())` and `.pvActivatesOnReturn { navigation.go(to: row.location) }` (both S13-35); only the title is the hit area. The cards (D-9) sit below, indented to the title. A future event card with a surface uses `PVTileCard`'s sizes and becomes a recipe on its second caller | One caller, so a snowflake; the kit's date · place line, focus-ring style and Return modifier; the kit's `.link` variant forces its own font | A row size on the shell (cut, user 2026-11-02); `.pv(.link)`; an event card recipe for one caller; `EventSecondaryLine(dateSize:)` (it is a call to `PVDatePlaceLine` after S13-34) |

**Rules this plan bends.** S13-D1 PPF-5 → D-1 opens cards on Timeline → user, 2026-11-02 → working rule "A timeline tile opens the participant on Timeline" (planning commit). R-24 and R-26 → D-5, D-6 → user, 2026-11-02 → `person-detail.md` (planning commit). S13-D1 PPF-W3 and Wide rule 2 ("the primary column keep[s] [its] Regular maximum widths", `docs/macos-client-patterns.md` § 9 and the `add-design-brief` Paste block) → D-15's opt-in full width → user, 2026-11-02 (full width) → STEP-7 rewords rule 2 in both places together: "Prose, forms and the primary column keep their Regular maximum widths, unless the section opts into Page columns' full-width mode; its text still keeps its measure. Wide adds columns; it does not stretch text or fields." The compact card (a second card size) and the term-name change in Vitals are no longer bent here: S13-35 and S13-34 carry them under the working rule "The person card lives in the design system; display is decoupled from content". No other rule is bent.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVPageColumns(primaryFill:)`, `columnWidths(…primaryFill:)`'s `band`, `pvPageVisibleRect` *(new)*; `ConclusionPageSection.fillsWidth` *(new)* | component, env, property | `PVPageColumnsMetrics.columnWidths`, its private scroll state | Extended (D-13, D-15) |
| `ConclusionShowMoreRow` *(moved)*; `L10n.Conclusions.showMore` *(renamed)* | snowflake, key | `PersonFamilyMoreRow` | Shared (D-8) |
| `ConclusionPageSectionHost` *(new)* | snowflake | `PersonFamilySection`'s host | Extracted (D-10) |
| `EventsList.spokenLine` *(new)*; `PersonTimelineEventRow` *(new, private)* | func, snowflake | `EventsList.accessibilityLabel`, `PVList`'s row slots | Split (D-5); one caller, so a snowflake (D-16) |
| `PersonTimelineContent`, `PersonTimelineLifeline` *(new)* | pure model, pure layout | `PersonFamilyContent`, `PersonFamilyLayout` | A different read and a time scale |
| `PersonTimelineSection` *(new)* | snowflake | `PersonFamilySection` | The list and the line, on D-10's host, the event rows a private snowflake (D-16) |
| `PersonTimelinePreviewData` *(new)* | fixture | `PersonFamilyPreviewData` | Timeline events (D-11) |
| `L10n.Conclusions.timeline*` *(new)* | keys | `L10n.Conclusions.family*` | New copy |

`Participant` (D-5) and Family's `PersonFamilyContent.Tile` both pair card content with a location; the Timeline adds `caption`. They stay two types: a shared struct would carry an optional field one caller never sets (round 1 M-9, round 2 M-20).

**Builds on.** `CatalogQueryKey.personEvents` with S13-33's value type; S13-35's `PVPersonCard`, `PVPersonCardContent`, `PVTileCard`, `PVTileCardSize`, `pvActivatesOnReturn`, `PVFocusRingButtonStyle`; S13-34's `PVTitleSource`, `PVTitleText`, `PVDatePlaceLine`, `PropertyTermDisplay.name(_:propertyKey:)`; `ConclusionPageSection.queryKey` and `glyph`, `ConclusionDetailPresentation`, `EventTitleDisplay`, `EventSecondaryDisplay`, `EventsList`, `PersonsList.location`, `CachedThumbnail`, `PVBadge`, `PVDivider`, `PVEmptyState`, `PVCallout`, `PVPageColumns`, `pvScrollsRowIntoView`, `pvMicroCaps`, `L10n.DesignSystem.spokenJoin`.
**Deletes.** `PersonFamilyMoreRow` (moved); `familyShowMore` and key `conclusions.family.showMore` (renamed); `PersonFamilyLoading`, `PersonFamilySection`'s host code and hand-built key (D-10); `.timeline` leaves the `pageSectionLater` case.
**Next adds.** S13-16 passes `.sources` to the host and adds `.sources → .personClaims`; Gallery (S13-D5) can opt into `fillsWidth`; any section that needs "in view" reads `pvPageVisibleRect`; a family tree places `PVPersonCard` from its own feature; an Event page with sections (R-1), or a second event-card caller, can turn the event row into a `PVEventCard` recipe on the same shell, with `EventsList.spokenLine`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A card opens a Person on Timeline | `PersonDetailView` / `ConclusionDetailPage` | `viewIdentity` resets the page; `expanded` resets on `entityId`; a merged participant gives the page's not-found Callout |
| An event title opens an Event page | `EventDetailView` | The Events list's path |
| Back to a Timeline entry | `PersonTimelineSection` | D-3 warms both keys; `.evict` while hidden, so Back reloads |
| Only undated; no events; one dated event; a recorded but undated birth or death | `PersonTimelineContent`, `PersonTimelineLifeline` | D-5's order rules (T-3, T-7); EmptyState (AC-1.7) |
| A Person page at Wide on Timeline, then on Vitals | `PVPageColumns` | `primaryFill` follows the open section; Vitals is capped again (T-8) |

### Scale

At the brief's sizes (8 – 25 events typical; 120 events and a 41-person event large): one read per view, warm with no query (S13-05). Swift builds the items once per value and fold change; the first paint draws at most 50 dated rows plus the markers, the undated group and the post-death rows, and 15 cards each, in a `LazyVGrid`. `PersonTimelineLifeline` is linear in the spans; it recomputes when an item frame, the fold set or the visible rect changes. The visible rect changes per scroll step, so the tether pass is linear in the items shown (at most about 60) and computes no scale (the scale depends only on spans and frames, cached in `@State` and recomputed on their change). T-9 lays out 120 spans.

### Edges

- **Assumptions.** The read's rows are in date order with undated last, and `span` is set exactly on dated rows (S13-33); D-5 keeps that order and only inserts markers and the undated group (T-3, T-4). The line spans the list's height, so unfolding a list changes the scale (T-7).
- **Neighbors.** A role that is not a term, or whose name is empty → no Badge; all roles empty → no "as" clause (AC-1.10). A placeholder with no `besideEventId` → first (birth) or last (death). A burial past death → the line runs past the death cap (AC-5.1). A recorded birth with no span → its row sits before `besideEventId`'s row as the birth marker (T-3). The death past the fold → it still shows, after "Show N more" and the undated group (AC-3.2). A researcher's term keyed `subject` → a neutral Badge, not the outline (T-5).
- **Removed inputs.** An event or participant unpromoted or merged after the read: the next refresh drops it.
- **Real data.** None — no table drives the section; S13-33 owns the role rank and the after-death table.
- **Untrusted size.** Events and participants are not bounded in Go; the first paint is bounded by D-7; the line is linear.

### Schema

None — no catalog change.

### Go core

None — S13-33.

### FFI

None — S13-33.

### macOS

**Frames in words** (turn 12; every frame is the Person page pane with the switcher on Timeline; header and switcher as S13-D1):

- **12a / 12d Full timeline, Regular and Wide.** An 80pt line column at the left; the items beside it. Each row, divided from the next by `PVDivider`, is the event row (D-16): the 72pt picture, then the title (display face, 18pt, two lines) as a button with the focus ring, below it the date in mono at body size · the place in italic, and the role Badges at the row's trailing edge (Subject in the outline Badge). Below, indented to the title, the role captions and compact person cards in a grid. The line runs from the birth cap (1817) to the burial (22 Sep 1880), four days past the death cap; 1861 – 1880 is empty. "Undated" with its hint and two rows sits just before the death row. 12d: the section fills the pane; eight cards a row; titles keep their measure.
- **12w Unknown ends.** "Birth unknown" (muted italic) with a dashed "?" cap 40pt before the baptism; "Death unknown" after the 1861 marriage and before the burial.
- **12x Short life.** Ernst, 1845 – 1848: year labels with quarter ticks; the baptism "Dec 1845" a month bar.
- **12b Scrolled.** The end of Johann's list: undated, death, burial; the rows in view light their points.
- **12f Row sheet.** Picture and Event mark; one and two role Badges; no, one and five people; a two-line title; a cut place; no date; no place; a ref title in mono; hover and focus on a title and a card. Compact card (S13-35): caption above, a 40pt likeness with the gender tint, the name on one line, the life span in years in mono.
- **12g / 12h / 12k Large event.** 15 cards and "Show 26 more", in the read's role order; unfolded; at Wide.
- **12j Long life.** Lise, 98 years: 10-year ticks, a point or bar for each of 120 events, the folded ones too; residences as range bars; her burial "bef. 1900" an open start past her death.
- **12l Only undated.** Birth unknown, Undated with two rows, Death unknown; the line is the two caps joined by a dashed stretch.
- **12m No events.** The EmptyState; no line. **12n / 12o / 12p** First load spinner; the error Callout; refreshing with the list and line drawn.
- **12q Keyboard.** Focus ring on Anna Weber's card; the Tab order and VoiceOver of AC-2.5 and TL-12.
- **12r After clicking Anna.** Crumbs *Persons › Weber, Anna › Timeline*, the Back menu open on "Johann Weber · Timeline".
- **12s / 12y / 12u / 12v / 12t Dark.** 12a, 12w, 12m, 12d and 12f in dark.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Section host | snowflake | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` *(new)* | D-10 |
| Timeline section | snowflake | `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)* | An `HStack(alignment: .top, spacing: 0)` of the 80pt line column and a `LazyVStack` of `content.items`, each `.id(item.id)`; `#Preview`s "Full timeline", "Unknown ends", "Short life", "Row sheet", "Large event", "Long life", "Only undated" from D-11 |
| Content, lifeline | pure | `PersonTimelineContent.swift`, `PersonTimelineLifeline.swift` *(new)* | D-5, D-13 |
| Line drawing | `Path` | `PersonTimelineSection.swift` | `PVColor.borderStrong` line and ticks, `PVColor.accent` points in view and tethers, `PVColor.textMuted` labels and placeholders, `StrokeStyle(dash: [3, 3])` for stubs and undated dots |
| Event row | `PersonTimelineEventRow` *(new, private)* with `PVDatePlaceLine` (S13-34) | `PersonTimelineSection.swift` | D-6, D-16, `PVBadge` (D-4) |
| Cards | `PVPersonCard(size: PVTileCardSize.compact)` (S13-35) | — | D-9 |
| Show more | `ConclusionShowMoreRow` | — | D-8 |
| Width, visible rect | `PVPageColumns(primaryFill:)`, `pvPageVisibleRect` | `ConclusionDetailPage.swift` | D-13, D-15 |

Ids: D-12. Accessibility: each row `.accessibilityElement(children: .contain)` labelled `row.spokenLabel`; the title a button; cards buttons with the card's spoken label; markers read their `spokenLabel`; the undated header `.isHeader`; the line hidden.

### Strings

L10n keys *(new)* under `L10n.Conclusions`: `timelineEmptyTitle` = "No events yet"; `timelineEmptyMessage` = "Events appear here when events this person took part in are promoted from an Evidence graph."; `timelineErrorTitle` = "Timeline could not be read"; `timelineUndated` = "Undated"; `timelineUndatedHint` = "No date, so no place on the line"; `timelineBirthUnknown` = "Birth unknown"; `timelineDeathUnknown` = "Death unknown"; `timelineAsRole(_ roles: String)` = "as %@". Renamed: `familyShowMore` → `showMore` (D-8). The compact card's copy and the term names are S13-35's and S13-34's. Each key is added in the step that first uses it, with its `Localizable.xcstrings` entry.

## 7. Steps

**STEP-0 — Check the prep PRs' names**
Files: this plan.
Do: read the **Landed** notes of S13-33, S13-34 and S13-35; if any name in § 4's Timeline data, Card and shell, or Display pieces rows differs, update § 4 and the Decisions that cite it (D-4, D-5, D-6, D-9, D-13, D-16) before code.
Check: the names match `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/PropertyTermDisplay.swift` and the folders S13-34 and S13-35 add to `macos/App/DesignSystem/` (`Recipes/PersonCard`, `Components/TileCard`, `Components/TitleText`).

**STEP-1 — Page columns: full width and the visible rect**
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/App/Features/Workspace/WorkspaceLocation.swift` (`fillsWidth`, all false), `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/ProvenenciaTests/PVPageColumnsTests.swift`. Skill: `add-ui-component`.
Do: D-15 with no section opting in; D-13's `pvPageVisibleRect`.
Check: T-8 passes; every page looks as before.

**STEP-2 — One fold row and one section host**
Files: `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` *(new)*, `ConclusionPageSectionHost.swift` *(new)*, `PersonFamilySection.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`, `add-workspace-place`.
Do: D-8, D-10 with Family as the only caller.
Check: `PersonFamilyContentTests`, `PersonFamilyLayoutTests` pass; Family's states keep their ids and look; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — The Timeline section's key and width**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift` (`sectionQueryKeys`: `.timeline` leaves the nil list), `PlaceRegistryTests.swift`.
Do: D-3; `fillsWidth` true for `.timeline` (round 2 X-16: until STEP-6 the Timeline's empty state is centred in the wider column, which reads the same).
Check: T-1, T-2 pass.

**STEP-4 — Items, lifeline geometry and fixtures**
Files: `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)*, `PersonTimelineLifeline.swift` *(new)*, `PersonTimelinePreviewData.swift` *(new)*, `EventsListView.swift` (`spokenLine`), `macos/ProvenenciaTests/PersonTimelineContentTests.swift` *(new)*, `PersonTimelineLifelineTests.swift` *(new)*, L10n files. Skill: `add-swift-test`.
Do: D-1, D-2, D-4 – D-7, D-11, D-13's geometry.
Check: T-3 – T-7, T-9 pass; `PersonsListTests` pass unchanged.

**STEP-5 — The section view**
Files: `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)*, L10n files. Skill: `add-ui-component`.
Do: D-6, D-9, D-10's second caller, D-12, D-13's drawing, tethers and clicks, D-14, D-16; the previews.
Check: the app builds; each preview reads as § 6's frame; in "Full timeline", a click on a point scrolls to its row and a second click scrolls again.

**STEP-6 — Wiring**
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`.
Do: `.timeline` draws `PersonTimelineSection(entityId:)`.
Check: `scripts/check-pr.sh` passes; the manual check (§ 10).

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `design/S13-D4-timeline-section.md` → `design/archive/`, `design/README.md`, `docs/macos-client-patterns.md` (§ 9 rule 2 reworded as **Rules this plan bends** says, and the opt-in full-width mode; the sectioned page: a section on `ConclusionPageSectionHost`), `.agents/skills/add-design-brief/SKILL.md` (the Paste block's rule 2, the same words, in the same commit). Skill: `add-design-brief`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md` reports no errors; § 9 and the Paste block's rule 2 read the same.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys` (extended) | `.timeline` → `.personEvents`; `.timeline.fillsWidth` true, every other false | D-3, D-15 |
| T-2 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `personDetailOnTimelineWarmsTheTimeline` *(new)* | A Person location on `.timeline` → `.conclusionDetail` and `.personEvents` | AC-4.2, D-3 |
| T-3 | Swift | `macos/ProvenenciaTests/PersonTimelineContentTests.swift` *(new)* `itemsInOrder` | `johann` → item ids 1, 2 … 9, `undated`, 12, 13, 10, 11; event 4's title, line, Badge Subject (outline), cards Maria and Friedrich with captions; event 1's picture set, 11's empty; 13's line empty. `unknownEnds` → `birth`, 2 … 9, `undated`, 12, 13, `death`, 11. `ida` → `birth`, `undated`, its two rows, `death`. A recorded birth with no span, beside event 2 → that row first, as the birth marker, not under Undated | AC-1.1, AC-1.3, AC-1.5, AC-1.6, AC-1.9, AC-5.4, D-5, D-6 |
| T-4 | Swift | same file `foldsAndOrder` | `lise` → 50 dated rows, `more.dated` (67), `undated` and its two rows, the death row, the two rows after it; expanded → every row; `largeEvent` → 15, `participantsHidden` 26, expanded → 41; input order kept, no sort; one person on two events → distinct ids | AC-3.1, AC-3.2, D-5, D-7 |
| T-5 | Swift | same file `rolesLocationsAndSpoken` | Event 8 → Badges Witness, Informant; a researcher's term keyed `subject` → neutral, not outline; all roles empty → no Badges and `spokenLabel` with no "as"; Anna's card content spoken "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851"; card location `…withPageSection(.timeline)`; row location `EventsList.location`; event 1's `spokenLabel` "Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject"; the birth marker's `spokenLabel` equals its row's; `EventsList.spokenLine` has no ref; `EventsList.accessibilityLabel` unchanged | AC-1.4, AC-1.10, AC-2.1, AC-2.3, AC-2.4, D-1, D-2, D-4, D-5 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonTimelineLifelineTests.swift` *(new)* `pointsToScale` | `johann` with item frames summing to a 2730pt line → birth at 0, the burial at 2730; 63.5 years → about 43pt a year → 1-year ticks; events 4 and 5 0.67 years apart (about 29pt); a day a dot, "1835" a bar, a range a bar, "bef. 1900" an open start; `lise` with a 4000pt line (98 years, about 41pt a year) → 1-year ticks; with a 1000pt line → 5-year ticks | AC-5.1, AC-5.2, AC-5.3, D-13 |
| T-7 | Swift | same file `placeholdersAndShortLives` | `unknownEnds` → two 40pt stubs outside the scale; a funeral and no death → the death stub before the funeral; `ida` → two caps joined; `ernst` with a 1200pt line → year labels and quarter ticks; a range under a year → drawn as a year; a taller line → a new scale; tethers only for items intersecting `visible` | AC-1.9, AC-5.4, AC-5.5, AC-5.6, AC-5.8, D-13 |
| T-8 | Swift | `macos/ProvenenciaTests/PVPageColumnsTests.swift` (cases added) | Pane 2296 `.wide` `primaryFill` → primary and group 2296 − 2 · gutterPage, `band` 1180; with a secondary → unchanged; `.regular` → unchanged; `primaryFill` false → primary 1180, `band == group` | AC-6.1, AC-6.2, D-15 |
| T-9 | Swift | `PersonTimelineLifelineTests.swift` `oneHundredTwentyEvents` | 120 spans and 120 item frames → 120 points in item order in one pass | AC-3.2, D-13 |
| T-10 | Swift | same file `tickCountIsCapped` *(new)* | A placed range of 9999 years (spans −9999 and 9999 on a 4000pt line) → at most 200 ticks, with a step above 50 years | D-13 |

Moved: the old T-1 (`PropertyTermDisplay.name(_:propertyKey:)`) and T-12 (one term rule in Vitals and the badge) to S13-34; the old T-2 (compact years, `helpText`, card sizes) to S13-35. AC-1.2 and the participant order are S13-33's tests. AC-1.7, AC-1.8, AC-2.2, AC-2.5, AC-2.6, AC-3.3, AC-4.1, AC-5.7 and AC-6.3 are covered by the manual check (§ 10); the fold reset on a new entity id (AC-3.3) is the section's `.onChange`, as Family's. Most likely to break: `PersonFamilyContentTests`, `PersonFamilyLayoutTests` (STEP-2), `PlaceRegistryTests` (STEP-3), `PersonsListTests` (STEP-4).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `PersonFamilyMoreRow`, `PersonFamilyLoading`, `familyShowMore`, `conclusions.family.showMore`, `One person tile, built once` and `S13-32` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/` and `L10n.swift` with no hit left outside spike-13 plan and review history and the map's S13-32 row; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (rule 2 in § 9 and the add-design-brief Paste block in the same commit)
- [ ] UI matches the brief inventory and frames 12a – 12y; no new kit primitive but `primaryFill` and `pvPageVisibleRect`
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] docs/macos-client-patterns.md: rule 2, the full-width mode and the section host (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Bar 4 and the slice's Check. Use a project with the brief § 5 sample promoted (or the product owner's own test people).

1. (AC-1.1, AC-1.3, AC-1.5, AC-1.6, AC-5.1, AC-5.2, AC-5.3, AC-5.8) Open Johann, choose Timeline: frame 12a as § 6 reads it. Scroll: the tethers follow the rows in view.
2. (AC-1.2, AC-1.4, AC-1.10) Luise's birth is not listed. Event 8 shows Witness and Informant.
3. (AC-2.1, AC-2.2, AC-2.4, AC-2.5, AC-2.6) Hover Anna's card: the three-line tooltip. Click it: her Timeline, crumbs *Persons › Weber, Anna › Timeline*; Back returns. Tab from event 1's title through its cards; Return opens; VoiceOver reads AC-2.4.
4. (AC-2.3, AC-5.7) Click event 4's title: the marriage's Event page. Back; click a point on the line, then the same point again: the list scrolls to its row both times; click the death cap: the death row scrolls to the top.
5. (AC-3.1, AC-3.2, AC-3.3) The "Large event" and "Long life" previews fold and unfold; a click on a folded point in "Long life" unfolds and scrolls. In the app, unfold a list, open another person, come Back: folded.
6. (AC-4.1) With Johann's Timeline showing, edit Maria Schulz through her Evidence graph and come Back: the change shows; "Refreshing" in the switcher row meanwhile.
7. (AC-1.7, AC-1.8, AC-1.9, AC-5.4, AC-5.5, AC-5.6) A person with no events; a history entry for a person merged since; the "Only undated", "Unknown ends" and "Short life" previews (frames 12l, 12w, 12x).
8. (AC-6.1, AC-6.2, AC-6.3) Widen past 1600pt on Timeline: the section fills the pane, the header stays at its measure; switch to Vitals: capped again; resize back: section, scroll, focus and unfolded lists kept. Dark mode: frames 12s – 12v.
9. Family looks and behaves as before (cards, Return, focus ring, "Show N more", loading, error and empty states).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A prep PR lands with other names | STEP-0 |
| The tether pass on every scroll step is slow at 60 items | The scale is cached; the pass is linear and reads frames only; T-9; the "Long life" preview scrolls smoothly |
| The line's scale follows the list's height, so unfolding a list rescales it | Intended (the board's TL-5); T-7 |
| `primaryFill` changes another page | Off by default; only `.timeline` sets it (T-1, T-8) |

**Stop and report if:**

- S13-33, S13-34 or S13-35 has not landed, or `CatalogPersonTimeline`, `CatalogTimelineSpan`, `CatalogLifeMarker`, `PVPersonCard`, `PVTileCardSize`, `PVDatePlaceLine`, `pvActivatesOnReturn`, `PVFocusRingButtonStyle`, `PVTitleText` or `PropertyTermDisplay.name(_:propertyKey:)` differ from § 4 in a way STEP-0 cannot map.
- The event row's title button is not reachable by Tab or VoiceOver inside the row's `.contain` group (D-16).
- `PVPageColumns` cannot publish its visible rect from the state it already measures, or cannot size its band apart from its columns (D-13, D-15).
- `pvScrollsRowIntoView` cannot scroll the page's own `ScrollView` from inside a section (D-13).
- A frame needs data the read does not return (an age, an estimated year): a follow-up, not Swift logic.

## 12. Open questions

None. The board's open items were answered by the user on 2026-11-02 (compact cards, participant order by role then name in Go, full width at Wide), as were the two round-2 questions (one term-name rule, now S13-34; the life span reused at year precision, now S13-34 and S13-35); the 72pt picture is kept as drawn (D-6).

## 13. PR

**Title:** S13-13 — Timeline section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md`.
