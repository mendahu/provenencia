# S13-13 — Timeline section

**Status:** Landed in #691 (2026-11-02)
**Spike / slice:** Spike 13 · Slice 4 — Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D4](../design/archive/S13-D4-timeline-section.md), board drawn 2026-11-02 (turn 12) and agreed with the user's answers of 2026-11-02 (compact cards, participant order by role, full width at Wide)
**Board:** Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Timeline.dc.html*, turn 12 "S13-D4 · Timeline — the event list and the lifeline (update 2026-11-02)", frames 12a – 12y. The implementer may not be able to open it: **Design → macOS** describes every frame in words and gives every measure
**Depends on:** S13-05 (landed [#680](https://github.com/mendahu/provenencia/pull/680)), S13-12 (landed [#676](https://github.com/mendahu/provenencia/pull/676)), S13-33 (landed [#687](https://github.com/mendahu/provenencia/pull/687): each event's span, the birth and death markers, participants ordered by role then name), S13-34 (landed [#689](https://github.com/mendahu/provenencia/pull/689): title styling and the date · place line as kit components, value formatting in `Platform/`, one term-name rule, year precision), S13-35 (landed [#690](https://github.com/mendahu/provenencia/pull/690): the card shell, Return activation, the focus-ring style, the person card in regular and compact sizes)
**Blocks:** —
**Branch:** `s13-13-timeline-section`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go or proto change (S13-33 has them)
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-05 and S13-12 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md) (STEP-8: its Paste block's Wide rule 2)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 3. Timeline](../person-detail.md#3-timeline) (R-2, R-3, R-23 – R-28, R-24a); brief [S13-D4](../design/archive/S13-D4-timeline-section.md); the S13-05, S13-12, S13-33, S13-34 and S13-35 plans ([`S13-05-person-timeline-read.md`](S13-05-person-timeline-read.md), [`S13-12-family-section.md`](S13-12-family-section.md), [`S13-33-timeline-positions.md`](S13-33-timeline-positions.md), [`S13-34-display-pieces-move-down.md`](S13-34-display-pieces-move-down.md), [`S13-35-person-card-joins-the-design-system.md`](S13-35-person-card-joins-the-design-system.md)) and their **Landed** notes; working rules: "Engine PRs run ahead of their slice" (data, not layout), "Each section reads its own query", "Timeline is direct participation, any role", "The person card lives in the design system; display is decoupled from content", "Colors reuse `PVGenderTreatment`", "A timeline tile opens the participant on Timeline", "Timeline lists fold", "The timeline is a lifeline", "The Timeline shows each event's picture", "Sections are places"
**Planned:** 2026-11-02 against `main` @ `51dccd4`; refreshed the same day after the board (turn 12), the descoping of S13-32, the S13-33 plan, review round 2, and the re-cut of the card work into S13-34 and S13-35 (base `04a2c99`, the prep plans' base); refreshed again after S13-33, S13-34 and S13-35 landed, against `main` @ `92f87db`: every § 4 name checked in the code and against the three Landed notes
**Landed:** the code follows the plan; four notes. (1) `PersonFamilyContent.Tile` was replaced by `PersonCardLink` throughout, with no typealias. (2) The line draws in one `Canvas` with a `Button` per point (not focusable); the click-to-scroll, tether look and Wide behaviour were checked only by a hosted-view snapshot, not in the running app. (3) Review round 1: the row's picture and role Badges are hidden from VoiceOver (the row's label names the event and the roles), "Show N more" moves focus to the first item it adds (`PersonTimelineContent.firstRevealed(by:)`), a request token (`whenRequested:`) alone scrolls (a change of `row` does not), and each point's click button has the row's spoken label as its tooltip. (4) The line draws in two `Canvas` layers: `PersonTimelineLine` (static) and `PersonTimelineTethers` (the only reader of the scroll); the preview fixture's year spans still add 1 to a year, and its day and month spans are literals.
**Review:** [`S13-13-timeline-section.review.md`](S13-13-timeline-section.review.md) — round 1 Fits with changes / Ready with changes; round 2 Fits with changes / Not ready, every finding Fixed or answered by the user; round 3 (after the scope change and the refresh) Fits with changes / Not ready, round 4 Fits with changes / Ready with changes; every finding Fixed or Declined with evidence; none open

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
| The section view: event rows on the card shell, compact cards, the line, its states (TL-1 – TL-13, TL-W1 – TL-W3; D-6, D-9, D-12, D-13, D-14, D-16) | STEP-5, STEP-6, STEP-7, AC-1.7, AC-1.8, AC-2.5, AC-2.6, AC-5.7, AC-5.8, AC-6.1 |

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
| The wide gutter (`PVSpacing.gutterPageWide`) for the full-width section | not planned: D-15 keeps `gutterPage`, as every page does today |
| An event card recipe (`PVEventCard`) | not planned: the event row has one caller; the evaluation's follow-up 6 when a second caller comes |
| `EventSubject` folding into {role, `PersonSummary`} | logged by S13-05 in [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Timeline section | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` | `.timeline, .sources` draw one `PVEmptyState` with `L10n.Conclusions.pageSectionLater`. `.family` draws `PersonFamilySection(entityId:)` |
| Section key | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection` (`String` raw values `vitals`, `family`, `timeline`, …), `queryKey(project:entityId:)`, `glyph` | `.family → .personFamily`; every other section nil. Read by `PlaceRegistry` Spec `.personDetail` and `ConclusionDetailLoaded.sectionRefreshing` (`ConclusionDetailPage.swift`) |
| Timeline key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.personEvents(project:entityId:)`; `CatalogQueryRegistry.swift` its descriptor | `.sessionFresh`, `.evict` when hidden, refreshes on any touched handle. Value `CatalogPersonTimeline` (S13-33) |
| Timeline data (S13-33) | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonTimeline { events; birth; death }`, `CatalogPersonEvent` (`event: CatalogEventSummary`, `roles: [CatalogConclusionValue]`, `participants: [CatalogEventParticipant]` with `person` and `roles`, `span: CatalogTimelineSpan?` with `from`, `to` on the year line, `openStart`, `openEnd`) with a computed `dated` (`span != nil`), `CatalogLifeMarker { eventId: String?; besideEventId: String? }` | Rows in date order (an open-start date at its known end), undated last; a recorded marker is exactly the header's birth or death when that event is a row, else a placeholder; `besideEventId` set whenever the marker has no span (S13-33 D-5, D-6); participants ordered by role rank, then name (S13-33 D-10); roles are `.term` values with labels filled in Go; the header's birth and death are canonical events of the product's own types only (S13-33 Landed note (5)); years are held to `datevalue.MinYear…MaxYear` (Landed note (6)) |
| Card and shell (S13-35) | `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` `PVPersonCard(content:projectDir:size:open:)` (`size: PVTileCardSize = .regular`), `PVPersonCardContent(summary:role:isFocus:locale:orders:)` (`spokenLabel`, `years`, `helpText`; `role` spoken first: "Mother, Anna Weber, …"); `macos/App/DesignSystem/Components/TileCard/PVTileCard.swift` `PVTileCard(size:isRaised:accessibilityLabel:help:action:thumbnail:title:lines:trailing:)`, `PVTileCardSize` (`.regular`, `.compact`; `height` 92 / 52, `thumbnail` 70 / 40, `titleLineLimit`, `linesLineLimit`, `minimumWidth` 120 / 232, `padding`); `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` `pvActivatesOnReturn(_:)`, `pvActivatesOnReturn(when:_:)`; `macos/App/DesignSystem/Components/FocusRingButton/PVFocusRingButtonStyle.swift` `PVFocusRingButtonStyle` | Landed as planned; the compact card's tooltip is `helpText` at `.compact` (S13-35 Landed notes (1) – (5)). Family draws `PVPersonCard`; `PersonTile` is gone |
| Display pieces (S13-34) | `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` `enum PVTitleSource` (`.name`, `.label`, `.ref`; `text`, `spokenText`), `PVTitleText(_:style:extraCount:lineLimit:emphasized:)` (`style: .row` takes the caller's font for a name or label; a ref draws in mono); `macos/App/DesignSystem/Components/DatePlaceLine/PVDatePlaceLine.swift` `PVDatePlaceLine(leading:date:place:accessibilityLabel:missing:dateSize:dateTone:extraCount:)` (date and place `String?`; `text(date:place:)`); `macos/App/Platform/PropertyTermDisplay.swift` `name(_ value: CatalogConclusionValue, propertyKey:) -> String?` (nil for a non-term or an empty name); `macos/App/Platform/` `PersonTitleDisplay`, `PersonLifeDisplay`, `DateRowDisplay` | Landed; `PVDatePlaceLine` gained `dateTone` (S13-34 Landed note (7)), and `EventSecondaryLine` stays as the Events row's wrapper (note (5)) |
| Show more | `macos/App/Features/Conclusions/PersonFamilySection.swift` private `PersonFamilyMoreRow` (on `pvActivatesOnReturn`); `L10n.Conclusions.familyShowMore(_:)` (key `conclusions.family.showMore`) | A ghost `PVButton` (`.pv(.ghost, size: .sm)`); Family only |
| Folds | `macos/App/Features/Conclusions/PersonFamilyContent.swift` `foldLimit`, `Fold`; `PersonFamilySection` `@State expanded`, reset `.onChange(of: entityId)` | The pattern to copy |
| Section host | `macos/App/Features/Conclusions/PersonFamilySection.swift` `PersonFamilySection` (builds `.personFamily(project:entityId:)` by hand), private `PersonFamilyLoading`, `PersonFamilyLoaded` | Handle lookup, `.task`, 96pt spinner, danger Callout, EmptyState and ids around Family's diagram |
| Read states | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPresentation` | Generic over the read's value |
| Page columns | `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` (README row "PageColumns" in `macos/App/DesignSystem/README.md`), `PVPageColumns` (three initializers: the main one, `where Secondary == EmptyView`, `where Header == EmptyView, Secondary == EmptyView`), `PVPageColumnsMetrics.columnWidths(pane:widthClass:hasSecondary:)` and `columnWidths(proposal:…)` returning one `group`, private `PVPageColumnsLayout` and `PVPageColumnsBandLayout` (both read `group`), private `@State offset` (the top sentinel's `-minY` in `PVPageColumnsMetrics.scrollSpace`, the space on the `ScrollView`) and `viewportHeight`, both written only behind `guard measures` (true for `headerPinned && hasHeader`, which Person pages are); `macos/ProvenenciaTests/PVPageColumnsTests.swift` | The primary is `min(widthContentMax, pane − 2 · gutterPage)` at every width: 1180pt at Wide. `ConclusionDetailLoaded` uses the `Secondary == EmptyView` initializer |
| Scroll to a row | `macos/App/DesignSystem/Components/ScrollRowIntoView/PVScrollRowIntoView.swift` `pvScrollsRowIntoView(_:anchor:animated:scrollsOnAppear:startsAtTopWhenRowsChange:restartsAt:whenChanged:)`; `macos/App/DesignSystem/README.md` ("do not add another `ScrollViewReader`") | Scrolls when its row value changes |
| Event title, line, location, spoken | `macos/App/Features/Conclusions/EventTitleDisplay.swift` `titleSource(_:locale:orders:)`; `EventsListView.swift` `EventsList.location(_:orders:)`, `EventsList.accessibilityLabel(_:orders:)` (title, date, ref through `L10n.Workspace.eventRowAccessibility`, then place, `morePlaces`), `EventSecondaryDisplay.content(_:locale:)` (a `Content` of `String`s, empty when missing) and `line(_:)`; `EventSecondaryLine(content:)`, a thin wrapper on `PVDatePlaceLine` that draws nothing when neither half is recorded, maps an empty half to nil, mutes the date (`dateTone: PVColor.textMuted`) at caption size and is hidden from VoiceOver; tested in `macos/ProvenenciaTests/PersonsListTests.swift` | The Events row's builders |
| Fonts | `macos/App/DesignSystem/Tokens/PVTypography.swift` `PVFont.display(size:weight:)`, `PVTypeScale.h3` (18), `.body` (16), `.caption` (13) | — |
| Seeded keys | `macos/App/Platform/ConclusionVocabulary.swift` `SeededRoleTerm.subject`, `SeededPropertyKey.role`; `CatalogOrigin.isProduct` | — |
| Person link | `macos/App/Features/Conclusions/PersonsListView.swift` `PersonsList.location(_:orders:)`; `WorkspaceLocation.withPageSection(_:)` | As S13-12 D-1 |
| Thumbnail | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail`; mark `PVMarkKey.subjectEvent`; `L10n.Conclusions.eventPictureThumb` | The Events list's thumb and fallback |
| Divider | `macos/App/DesignSystem/Components/Divider/` `PVDivider` | The kit rule between rows |
| Family's tile | `macos/App/Features/Conclusions/PersonFamilyContent.swift` `PersonFamilyContent.Tile { id; content: PVPersonCardContent; location: WorkspaceLocation? }`, read by `PersonFamilySection.swift` and `PersonFamilyContentTests.swift` | Family only |
| Product term key | `macos/App/Platform/GenealogyStore.swift` `CatalogPropertyTerm.productKey`; `macos/App/DesignSystem/Recipes/GenderTreatment/PVGenderTreatment.swift` `treatment(_:basis:)` (`CatalogOrigin.isProduct(origin) ? key : ""` by hand) | No `CatalogConclusionValue` form |
| Title sizes | `PVTitleText.Style` `.row`, `.display(compact:)`; `displayText` ignores `lineLimit`; the header calls `.display(compact:)` in `ConclusionDetailPage.swift` | — |
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
| D-4 | Role badges and captions. Each role is S13-34's `PropertyTermDisplay.name(value, propertyKey: SeededPropertyKey.role)`; nil roles are dropped. The person's roles show as Badges in the read's order: `PVBadge` neutral; drawn `PVBadge(text:tone: .neutral, subtle: isSubject)` (the outline is `subtle`), where `isSubject` = `value.productKey == SeededRoleTerm.subject`. `CatalogConclusionValue.productKey: String?` *(new)* in `macos/App/Platform/GenealogyStore.swift`, beside `CatalogPropertyTerm.productKey`: a `.term`'s key when `CatalogOrigin.isProduct(origin)`, else nil, and nil for any other value; `PVGenderTreatment.treatment(_:basis:)` moves onto it (its `CatalogOrigin.isProduct(origin) ? key : ""` becomes `value.productKey ?? ""`) (frame 12f). A participant's role is its names joined with `L10n.DesignSystem.spokenJoin` ("Witness, Informant"), nil when it has none; it is the card content's `role` (spoken first by `PVPersonCardContent`) and the caption drawn above the card, one field for both (round 3 M-31) | One rule turns a term into a name (S13-34); one rule for "a product term keyed like this" on conclusion values (round 3 M-34); round 2 M-22 | A timeline-only role type; a bare key compare for Subject; an inline `isProduct` check; a separate caption field |
| D-5 | `PersonTimelineContent` *(new, pure, `Equatable`)* in `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)*: `init(timeline: CatalogPersonTimeline, expanded: Set<Fold> = [], locale:, orders:)` builds **`items: [Item]`**, the one display order the list and the line both read (round 2 M-12, X-15). `enum Item: Identifiable { case marker(Marker), row(Row), undatedHeader, more(Fold, count: Int) }`, where a `more` item's fold is `.dated` or `.undated` only (a row's participant fold is drawn inside the row, D-9); ids: a row's and a recorded marker's are the event id, a placeholder's `"birth"` / `"death"`, the header `"undated"`, a more item `"more.dated"` / `"more.undated"`. `Marker` holds `kind` (`.birth`, `.death`), `row: Row?` (the recorded event's row; nil for a placeholder) and `spokenLabel` (a recorded marker reads its row's label, round 2 X-18; a placeholder its text). **Order:** (1) the dated rows in the read's order; (2) the birth marker: a recorded birth with a span is its own row in place; otherwise (a placeholder, or a recorded birth with no span) the marker goes immediately before the row `besideEventId` names, or first when nil; (3) the death marker: a recorded death with a span is its row in place; otherwise it goes immediately after the row `besideEventId` names, or last when nil; (4) the undated header and the undated rows (a recorded but undated birth or death leaves this list: it is its marker) go immediately before the death marker. Placement runs on the full row list before folding; a `besideEventId` that names no row in the read is treated as nil. With no dated row, the order is the birth marker, the undated group, the death marker. **Folds** (D-7) apply to the dated rows before the death marker (a recorded birth's row counts among them; a birth placeholder is never folded): past the 50th, the rest of that stretch is replaced by one `more(.dated, count:)` item placed where the hidden rows were (a death marker placed beside a hidden row follows that item); the death marker, the undated group and every row after the death marker always show. The undated rows fold at 50 behind `more(.undated, count:)`. `Row`: `id`, `title: PVTitleSource` (`EventTitleDisplay.titleSource`), `line: EventSecondaryDisplay.Content`, `imageRelPath`, `roles: [RoleBadge]` (`text`, `isSubject`), `location` (D-2), `span: CatalogTimelineSpan?`, `participants: [PersonCardLink]` (the shown ones), `participantsHidden: Int`, `spokenLabel` = `EventsList.spokenLine(event, orders:)` *(new)*, (with the content's `locale`) then, when the roles are not empty, `L10n.Conclusions.timelineAsRole(spokenJoin(roles))` joined with `spokenJoin` ("Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject"). `EventsList.spokenLine(_:orders:locale:) -> String` *(new; `locale: Locale = .autoupdatingCurrent`, passed to `titleSource` and `EventSecondaryDisplay.content`, round 4 M-42)* is `spokenJoin` of the title, then the date when not empty, then `EventsList.placeTail(_:)` *(new, private)*: the place when not empty and `morePlaces` when `extraPlaces > 0`; no ref (round 2 M-19). `accessibilityLabel(_:orders:locale:)` gains the same defaulted `locale` (existing callers unchanged), keeps its head (`eventRowAccessibility` with the ref) and reads its place and "+N places" from the same `placeTail`, so its output is unchanged (round 3 M-36). Empty parts are never joined (round 3 X-38). A participant is a `PersonCardLink` *(new)*: `id` ("\(eventId)-\(personId)"), `content: PVPersonCardContent` (S13-35; `role:` D-4's role, `isFocus: false`), `location` (D-1). `PersonCardLink { id: String; content: PVPersonCardContent; location: WorkspaceLocation? }` lives in `macos/App/Features/Conclusions/PersonCardLink.swift` *(new)*; Family's `PersonFamilyContent.Tile` (same three fields, `location` nil for the focus) moves onto it in STEP-2 and is deleted (round 3 M-31). `allSpans: [Span]` with `struct Span: Equatable { id: String; span: CatalogTimelineSpan }` gives every dated row's span, folded or not, for the line (a struct, so `Equatable` is synthesised, round 3 X-42). `isEmpty` is `timeline.events.isEmpty` | One owner of the order; the folded-death and undated-death cases decided once and tested (T-3, T-4) | Two consumers deriving the order from markers; a Bool on `accessibilityLabel` |
| D-6 | The picture is `CachedThumbnail(projectDir:relPath: row.imageRelPath, size: 72, cornerRadius: PVRadius.sm, fallbackMark: .subjectEvent, fallbackLabel: L10n.Conclusions.eventPictureThumb, colors: nil)`, 72pt (frame 12f) | The board's 72pt (the event row is a Timeline snowflake; the shell has no row size, user 2026-11-02) | The Events list's 44pt |
| D-7 | Folds (working rule "Timeline lists fold"): `eventFoldLimit = 50`, `participantFoldLimit = 15`; `enum Fold: Hashable { case dated, undated, participants(eventId: String) }`. `expanded` lives in the section as `@State`, reset `.onChange(of: entityId)` | The user's caps; Family's fold shape | A shared fold type with Family |
| D-8 | One "Show N more" row: `PersonFamilyMoreRow` moves to `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` *(new)* as `ConclusionShowMoreRow(count:expand:)`, the same `.pv(.ghost, size: .sm)` button on S13-35's `pvActivatesOnReturn`. `L10n.Conclusions.familyShowMore(_:)` → `showMore(_:)`, key `conclusions.showMore`, comment "Button under a folded list on a Conclusion page that shows the rest; argument is how many are hidden"; the String Catalog entry moves with it | One row and one string for both sections | A private copy |
| D-9 | Each event's cards are `PVPersonCard(content: link.content, projectDir:, size: PVTileCardSize.compact, open: { if let location = link.location { navigation.go(to: location) } })` (S13-35), laid out in `LazyVGrid(columns: [GridItem(.adaptive(minimum: PVTileCardSize.compact.minimumWidth), spacing: 8, alignment: .topLeading)], alignment: .leading, spacing: 8)`, starting at the title's leading edge and filling the row's full width, each `PVTileCardSize.compact.height` tall; each cell is a `VStack(alignment: .leading, spacing: 4)` of the caption, `link.content.role` in `.pvMicroCaps()` `PVColor.textMuted`, `.accessibilityHidden(true)` as Family's captions (the card speaks the role) (none when nil, no height reserved), and the card (frame 12f). When `participantsHidden > 0`, a `ConclusionShowMoreRow(count:expand:)` (D-8) follows the grid inside the row, id `persons.detail.timeline.more.<eventId>`, and inserts `.participants(eventId:)` into `expanded` | The card's size lives on the card (S13-35; round 2 M-21); three or four a row at Regular, eight at Wide | Family's column formula; a timeline copy of the card |
| D-10 | One host for a page section's own read: `ConclusionPageSectionHost<Value, Content: Equatable, Body: View>` *(new)* in `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` *(new)*: `init(section: ConclusionPageSection, entityId: String, errorTitle: LocalizedStringResource, emptyTitle: LocalizedStringResource, emptyMessage: String, make: @escaping (Value) -> Content, isEmpty: @escaping (Content) -> Bool, @ViewBuilder body: @escaping (Content) -> Body)`. The key is `section.queryKey(project: session.projectKey, entityId:)` (a `precondition` that it is not nil), the glyph `section.glyph`, the id root `"persons.detail.\(section.rawValue)"` (round 2 M-17). It reads `@Environment(WorkspaceSession.self)`, gets the handle with `session.queryHandle(key) as QueryHandle<Value>?`, starts it in `.task { let _: QueryHandle<Value> = session.query(key) }`, and maps `ConclusionDetailPresentation<Content>`: `.loading` (or no handle) → `ProgressView()` 96pt from the top; `.failed` → `PVCallout(tone: .danger, title: errorTitle, message:)`, id `<root>.error`; `isEmpty` → `PVEmptyState`, id `<root>.empty`; else `body(content)`, id `<root>`. Family moves onto it (its hand-built key goes) with ids, copy and look unchanged; the Timeline is its second caller | Only the content differs (round 1 M-6); S13-16 is the third | Copying Family's host; passing the key and glyph separately |
| D-11 | Fixtures: `macos/App/Features/Conclusions/PersonTimelinePreviewData.swift` *(new)* (`#if DEBUG`, internal) `enum PersonTimelinePreviewData`, people from `PersonFamilyPreviewData.summary`, except Anna Weber, built by hand with birth 3 May 1790 at Bremen (a `CatalogChainedPlace`) and death 1851, as `PersonDetailTests.lifeLinesOmitMissingHalves` builds a summary (round 4 X-53), events from one `event(_ id:, type:, title:, date:, place:, picture:, roles:, participants:)` helper: `johann` (brief § 5, events 1 – 13, death 18 Sep 1880, burial 22 Sep 1880, pictures on 1, 2, 4, 5, 9, 10; spans as S13-33 gives them); `unknownEnds` (johann without 1 and 10; birth placeholder beside event 2, death placeholder beside event 9); `ernst` (birth 3 Apr 1845, baptism "Dec 1845", death 14 Feb 1848); `largeEvent` (one burial with 41 participants: 4 Child, 1 Witness, 1 Informant, 6 Pallbearer, 29 with the researcher's role "Mourner", in S13-33's order); `lise` (born 1801, died 1899: 120 dated events, the death the 118th, a residence range every few years, a burial "bef. 1900" and an undated two; the writer's choice of the rest, tests check only the named properties); `ida` (two undated events); `ludwig` (none). The section's `#Preview`s and the tests use it | One fixture for previews and tests (S13-12 Landed note (6)) | A test-only copy |
| D-12 | Accessibility ids under `persons.detail.timeline`: the list; `.line`; an item `.item.<Item.id>`; a title `.title.<eventId>`; a card `.tile.<PersonCardLink.id>`; a row's participant fold `.more.<eventId>`; D-10's `.error` and `.empty` | Unique ids from the item ids | One shared `.more` id |
| D-13 | The lifeline (frames 12a – 12x). **Geometry** is pure, `struct PersonTimelineLifeline: Equatable` *(new)* in `macos/App/Features/Conclusions/PersonTimelineLifeline.swift` *(new)*, `static let columnWidth: CGFloat = 80` (the section and D-15 read it, round 3 M-37). `init(content: PersonTimelineContent, itemFrames: [String: CGRect])`, frames keyed by `Item.id` in the section's own coordinate space, gives: `length: CGFloat` (the first item's `minY` to the last item's `maxY`); `marks: [Mark]` in item order, `struct Mark: Equatable { id: String; shape: Shape }` with `enum Shape: Equatable { case dot(y: CGFloat), bar(from: CGFloat, to: CGFloat, openStart: Bool, openEnd: Bool), cap(y: CGFloat, isPlaceholder: Bool) }`; `stubs: [ClosedRange<CGFloat>]` (dashed stretches); `ticks: [Tick]`, `struct Tick: Equatable { y: CGFloat; year: Int?; isQuarter: Bool }` (`year` nil on a quarter tick). Tethers are a method, `tethers(itemFrames:visible:) -> [Tether]`, `struct Tether: Equatable { id: String; from: CGPoint; to: CGPoint; isDashed: Bool }`, so a scroll step computes no scale. **Scale:** the placed range is the earliest `span.from` to the latest `span.to` of `content.allSpans`; a range under one year becomes `[from, from + 1)`; placeholder markers take a fixed 40pt stub each, outside the scale; points per year = (length − stubs) ÷ years. With no span (only undated, AC-1.9) there is no scale and no tick: the two caps sit at the first and last items and one dashed stub joins them. **Ticks:** at the years divisible by the step inside the range, the step the smallest in 1, 2, 5, 10, 20, 50 with ticks at least 34pt apart; if even the 50-year step gives more than 200 ticks, the step grows through 100, 200, 500, 1000, 2000, 5000, 10000 until it gives at most 200 (the engine bounds years to `datevalue.MinYear…MaxYear`, S13-33, so a placed range is at most 19999 years (-9999.0 to 10000.0)); quarter ticks when a year is at least 136pt; a year label is `String(year)` ("1820", "-44"). **Mapping** (round 4 X-51): a year's y is `(year − range.from) × pointsPerYear` plus 40pt for each stub inserted at or before it. A birth stub is inserted at the `from` of the row its marker precedes (the range start when it precedes none); a death stub at the `to` of its `besideEventId` row (the range end when nil). A folded row's span still counts, and a bar that crosses an insertion point is drawn from its mapped `from` to its mapped `to`, not split. **Marks:** a span with `to − from <= 1.0 / 365.0 + 1e-9` is a dot (S13-33 divides every year by 365, leap years too; round 4 X-52); else a bar; `openStart` / `openEnd` fades that end; birth and death are caps: a recorded marker with a span at its row's point; a placeholder a dashed "?" cap (`isPlaceholder: true`) on its stub; a recorded marker with no span a filled cap (`isPlaceholder: false`) on a stub of its own, placed as a placeholder's, clickable, with an undated row's dashed dot and no curve (round 4 X-50). **Tethers:** for each row item whose frame intersects `visible`, from its point on the line to a dot at the row's leading edge at the row's `minY + 12`; an undated row gets a dashed dot at its edge and no curve. **Measures** (the writer's, from the board's look; round 3 X-36): in the 80pt column the line is at x = 56, 1.5pt `PVColor.borderStrong`; a year tick 8pt and a quarter tick 4pt left of the line, 1pt; labels `PVFont.mono(size: PVTypeScale.micro)` `PVColor.textMuted`, trailing-aligned 4pt left of the ticks; a dot 8pt; a bar 6pt wide with round ends, an open end fading to clear over its last 16pt; a cap a 12pt circle (recorded: filled `borderStrong`; placeholder: dashed outline with "?" in micro mono); a tether a cubic curve with both control points at the column's half-way x, 1pt `PVColor.accent`, ending in a 6pt dot at x = 80; points of rows in view in `PVColor.accent`, others `borderStrong`; dashes `[3, 3]`. **Coordinates** (round 3 X-30, M-33; round 4 M-40): `PVPageColumns` publishes `EnvironmentValues.pvPageViewportSize: CGSize` *(new)*, the scroll view's size, from a `@State viewportSize` *(new)* that the existing viewport `onGeometryChange` writes on both the pinned and the unpinned path (`viewportHeight`, `offset` and `refreshScrolled` stay behind `measures`, unchanged); outside a `PVPageColumns` the value is `.zero`, and it changes on resize, not on scroll. `PVPageColumnsMetrics.scrollSpace` is the viewport's own space (it sits on the `ScrollView`), so the section's frame in it, measured with `onGeometryChange` on every change, already says where the viewport is: `visible = CGRect(x: -frame.minX, y: -frame.minY, width: size.width, height: size.height)`, with no second offset. Each item's frame is measured in the section's named space `"persons.detail.timeline"`. A zero-height `visible` (no viewport size) draws no tether. **Clicks:** a point or cap sets `@State scrollTarget: String?` to its item id and bumps `@State scrollRequest: Int`; the list applies `.pvScrollsRowIntoView(scrollTarget, anchor: .top, whenRequested: scrollRequest)` inside `PVPageColumns`' scroll view. `pvScrollsRowIntoView` gains `whenRequested request: AnyHashable? = nil` *(new)*: when the token changes it scrolls to `row` again on the next main-actor turn, after a row inserted in the same update exists; no existing caller passes it (round 3 M-32). A folded row's point inserts `.dated` into `expanded` and sets the target and token in the same update. The line, ticks, points and tethers are `.accessibilityHidden(true)`. Swift does no date arithmetic | Working rule "The timeline is a lifeline"; one scroll-to-row mechanism, which owns its own timing, and one measurement of the scroll position, the section's own frame | A second `ScrollViewReader`; a `Task` hop in the feature; publishing `PVPageColumns`' offset beside the section's frame (two measurements of one fact, round 4 M-40); fixed points per year |
| D-14 | Placeholder items read "Birth unknown" / "Death unknown" in muted italic body text, ids `.item.birth` / `.item.death`; the undated header is "Undated" (`.isHeader`) with the hint "No date, so no place on the line" in caption muted (frames 12a, 12l) | The board's TL-3 and TL-5a | An estimated year |
| D-15 | Full width at Wide (user, 2026-11-02). `PVPageColumnsMetrics.columnWidths(pane:widthClass:hasSecondary:primaryFill:)` *(new parameter, default false)* returns `(primary, secondary, group, band)` *(new member)*: unchanged when `primaryFill` is false (`band == group`); when true, `widthClass == .wide` and no secondary, `primary = group = pane − 2 · gutterPage` and `band = min(widthContentMax, primary)`; `primaryFill` with a secondary is ignored. The `proposal:` overload passes it through; with no finite proposal it is ignored. `PVPageColumnsLayout` centres on `group`; `PVPageColumnsBandLayout` sizes the header from `band` (round 2 M-16, X-21). All three `PVPageColumns` initializers gain `primaryFill: Bool = false`. `ConclusionPageSection.fillsWidth: Bool` *(new)*, true for `.timeline` only; `ConclusionDetailLoaded` passes `primaryFill: pageSection?.fillsWidth ?? false` in its `.content` case only (loading and failed stay capped, round 3 X-42). The gutter stays `PVSpacing.gutterPage`, as `columnWidths` uses today; `gutterPageWide` is not used (round 3 X-42). Inside the Timeline, each row's top line (picture, title and date line, Badges) keeps `.frame(maxWidth: PVSpacing.widthContentMax − PersonTimelineLifeline.columnWidth, alignment: .leading)`, so the Badges sit at the measure's trailing edge (round 3 X-35); the card grids fill | The board's TL-W1 and the user's answer; one owner of every width | A Timeline-only frame override; one `group` for both the columns and the band |
| D-16 | The event row is a Timeline snowflake (user, 2026-11-02: no row size on the shell), private `PersonTimelineEventRow` *(new)* in `PersonTimelineSection.swift`: an `HStack(alignment: .top, spacing: 16)` of D-6's picture, a `VStack(alignment: .leading, spacing: 4)` of the title button and `EventSecondaryLine(content: row.line, dateSize: PVTypeScale.body)`, and the role Badges at the trailing edge; no card surface; the row is `.accessibilityElement(children: .contain)` labelled `row.spokenLabel`, so the title button stays reachable. The title button (round 2 M-18, X-20) is `Button { navigation.go(to: row.location) } label: { PVTitleText(row.title, style: .display(compact: true), lineLimit: 2) }` (round 4 X-53): the header's compact display title (h3 display face, a label italic, a ref mono at h4) at two lines (round 3 M-30). `PVTitleText.init`'s existing `lineLimit: Int = 1` becomes `lineLimit: Int? = nil`: `.row` reads `lineLimit ?? 1`, and `displayText` applies `.lineLimit(lineLimit).truncationMode(.tail)` only when it is set; `Style` is unchanged, and the header, `ConclusionListPage` and `PVPersonCard` behave as today (round 4 M-41). Its `.isHeader` trait stays: each event title is a heading VoiceOver can jump between with `.buttonStyle(PVFocusRingButtonStyle())` and `.pvActivatesOnReturn { navigation.go(to: row.location) }` (both S13-35); only the title is the hit area. `EventSecondaryLine` gains `var dateSize: CGFloat = PVTypeScale.caption` *(new)*, passed to its `PVDatePlaceLine`; the Events row passes nothing and is unchanged. The cards (D-9) sit below, indented to the title. A future event card with a surface uses `PVTileCard`'s sizes and becomes a recipe on its second caller | One caller, so a snowflake; the Events row's own date · place line (brief inventory: "unchanged"), which already owns the no-line guard and the empty-to-nil mapping (S13-34 Landed note (5)), so AC-1.6 needs no second copy of either; the kit's focus-ring style and Return modifier; one title component at both sizes (round 3 M-30); the kit's `.link` variant forces its own font | A row size on the shell (cut, user 2026-11-02); `.pv(.link)`; `PVTitleText` `.row` with an outer font (a ref keeps its own small mono); an event card recipe for one caller; calling `PVDatePlaceLine` directly (it would repeat `EventSecondaryLine`'s guard and mapping) |

**Rules this plan bends.** S13-D1 PPF-5 → D-1 opens cards on Timeline → user, 2026-11-02 → working rule "A timeline tile opens the participant on Timeline" (planning commit). R-24 and R-26 → D-5, D-6 → user, 2026-11-02 → `person-detail.md` (planning commit). S13-D1 PPF-W3 and Wide rule 2 ("the primary column keep[s] [its] Regular maximum widths", `docs/macos-client-patterns.md` § 9 and the `add-design-brief` Paste block) → D-15's opt-in full width → user, 2026-11-02 (full width) → STEP-8 rewords rule 2 in both places together: "Prose, forms and the primary column keep their Regular maximum widths, unless the section opts into Page columns' full-width mode; its text still keeps its measure. Wide adds columns; it does not stretch text or fields." The compact card (a second card size) and the term-name change in Vitals are no longer bent here: S13-35 and S13-34 carry them under the working rule "The person card lives in the design system; display is decoupled from content". No other rule is bent.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVPageColumns(primaryFill:)`, `columnWidths(…primaryFill:)`'s `band`, `pvPageViewportSize` *(new)*; `ConclusionPageSection.fillsWidth` *(new)* | component, env, property | `PVPageColumnsMetrics.columnWidths`, its private scroll state | Extended (D-13, D-15) |
| `ConclusionShowMoreRow` *(moved)*; `L10n.Conclusions.showMore` *(renamed)* | snowflake, key | `PersonFamilyMoreRow` | Shared (D-8) |
| `PersonCardLink` *(moved from `PersonFamilyContent.Tile`)* | type | `PersonFamilyContent.Tile` | Shared by Family and the Timeline (D-5) |
| `CatalogConclusionValue.productKey` *(new)* | property | `CatalogPropertyTerm.productKey` | Its twin on values (D-4) |
| `PVTitleText(lineLimit:)` made optional; `pvScrollsRowIntoView(whenRequested:)` *(new)* | parameter | the same components | Extended (D-13, D-16) |
| `ConclusionPageSectionHost` *(new)* | snowflake | `PersonFamilySection`'s host | Extracted (D-10) |
| `EventsList.spokenLine` *(new)*; `PersonTimelineEventRow` *(new, private)*; `EventSecondaryLine.dateSize` *(new)* | func, snowflake, property | `EventsList.accessibilityLabel`, `PVList`'s row slots, `EventSecondaryLine` | Split (D-5); one caller, so a snowflake (D-16); the Events line extended, not copied (D-16) |
| `PersonTimelineContent`, `PersonTimelineLifeline` *(new)* | pure model, pure layout | `PersonFamilyContent`, `PersonFamilyLayout` | A different read and a time scale |
| `PersonTimelineSection` *(new)* | snowflake | `PersonFamilySection` | The list and the line, on D-10's host, the event rows a private snowflake (D-16) |
| `PersonTimelinePreviewData` *(new)* | fixture | `PersonFamilyPreviewData` | Timeline events (D-11) |
| `L10n.Conclusions.timeline*` *(new)* | keys | `L10n.Conclusions.family*` | New copy |

Family and the Timeline share `PersonCardLink` (D-5): with the role on the card content (D-4), the Timeline's former `Participant` had no field Family's `Tile` lacks (round 3 M-31, superseding round 1 M-9 and round 2 M-20).

**Builds on.** `CatalogQueryKey.personEvents` with S13-33's value type; S13-35's `PVPersonCard`, `PVPersonCardContent`, `PVTileCard`, `PVTileCardSize`, `pvActivatesOnReturn`, `PVFocusRingButtonStyle`; S13-34's `PVTitleSource`, `PVTitleText`, `PropertyTermDisplay.name(_:propertyKey:)`; `EventSecondaryLine` on S13-34's `PVDatePlaceLine`; `ConclusionPageSection.queryKey` and `glyph`, `ConclusionDetailPresentation`, `EventTitleDisplay`, `EventSecondaryDisplay`, `EventsList`, `PersonsList.location`, `CachedThumbnail`, `PVBadge`, `PVDivider`, `PVEmptyState`, `PVCallout`, `PVPageColumns`, `pvScrollsRowIntoView`, `pvMicroCaps`, `L10n.DesignSystem.spokenJoin`.
**Deletes.** `PersonFamilyMoreRow` (moved); `PersonFamilyContent.Tile` (moved to `PersonCardLink`); `PVGenderTreatment`'s hand-written product check; `familyShowMore` and key `conclusions.family.showMore` (renamed); `PersonFamilyLoading`, `PersonFamilySection`'s host code and hand-built key (D-10); `.timeline` leaves the `pageSectionLater` case.
**Next adds.** S13-16 passes `.sources` to the host and adds `.sources → .personClaims`; Gallery (S13-D5) can opt into `fillsWidth`; any section that needs "in view" reads `pvPageViewportSize`; a family tree places `PVPersonCard` from its own feature; an Event page with sections (R-1), or a second event-card caller, can turn the event row into a `PVEventCard` recipe on the same shell, with `EventsList.spokenLine`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A card opens a Person on Timeline | `PersonDetailView` / `ConclusionDetailPage` | `viewIdentity` resets the page; `expanded` resets on `entityId`; a merged participant gives the page's not-found Callout |
| An event title opens an Event page | `EventDetailView` | The Events list's path |
| Back to a Timeline entry | `PersonTimelineSection` | D-3 warms both keys; `.evict` while hidden, so Back reloads |
| Only undated; no events; one dated event; a recorded but undated birth or death | `PersonTimelineContent`, `PersonTimelineLifeline` | D-5's order rules (T-3, T-7); EmptyState (AC-1.7) |
| A Person page at Wide on Timeline, then on Vitals | `PVPageColumns` | `primaryFill` follows the open section; Vitals is capped again (T-8) |

### Scale

At the brief's sizes (8 – 25 events typical; 120 events and a 41-person event large): one read per view, warm with no query (S13-05). Swift builds the items once per value and fold change; the first paint draws at most 50 dated rows plus the markers, the undated group and the post-death rows, and 15 cards each, in a `LazyVGrid`; the item stack is a plain `VStack` (D-13 needs every frame), so unfolding Lise lays out 120 rows at once. `PersonTimelineLifeline` is linear in the spans; it recomputes when an item frame, the fold set or the visible rect changes. The visible rect changes per scroll step, so the tether pass is linear in the items shown (at most about 60) and computes no scale (the scale depends only on spans and frames, cached in `@State` and recomputed on their change). T-9 lays out 120 spans.

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
| Timeline section | snowflake | `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)* | An `HStack(alignment: .top, spacing: 0)` of the `PersonTimelineLifeline.columnWidth` line column and a `VStack(alignment: .leading, spacing: 0)` of `content.items`, each `.id(item.id)`: a plain stack, because the line's length and scale need every item's frame (round 3 X-31); D-7's folds bound it to about 60 items and 15 cards each until unfolded, 120 rows unfolded at the brief's largest; `#Preview`s "Full timeline", "Unknown ends", "Short life", "Row sheet", "Large event", "Long life", "Only undated" from D-11 |
| Content, lifeline | pure | `PersonTimelineContent.swift`, `PersonTimelineLifeline.swift` *(new)* | D-5, D-13 |
| Line drawing | `Path` | `PersonTimelineSection.swift` | `PVColor.borderStrong` line and ticks, `PVColor.accent` points in view and tethers, `PVColor.textMuted` labels and placeholders, `StrokeStyle(dash: [3, 3])` for stubs and undated dots |
| Event row | `PersonTimelineEventRow` *(new, private)* with `EventSecondaryLine(dateSize:)` | `PersonTimelineSection.swift` | D-6, D-16, `PVBadge` (D-4) |
| Cards | `PVPersonCard(size: PVTileCardSize.compact)` (S13-35) | — | D-9 |
| Show more | `ConclusionShowMoreRow` | — | D-8 |
| Width, visible rect | `PVPageColumns(primaryFill:)`, `pvPageViewportSize` | `ConclusionDetailPage.swift` | D-13, D-15 |

Ids: D-12. Accessibility: each row `.accessibilityElement(children: .contain)` labelled `row.spokenLabel`; the title a button; cards buttons with the card's spoken label; markers read their `spokenLabel`; the undated header `.isHeader`; the line hidden.

### Strings

L10n keys *(new)* under `L10n.Conclusions`: `timelineEmptyTitle` = "No events yet"; `timelineEmptyMessage` = "Events appear here when events this person took part in are promoted from an Evidence graph."; `timelineErrorTitle` = "Timeline could not be read"; `timelineUndated` = "Undated"; `timelineUndatedHint` = "No date, so no place on the line"; `timelineBirthUnknown` = "Birth unknown"; `timelineDeathUnknown` = "Death unknown"; `timelineAsRole(_ roles: String)` = "as %@". Renamed: `familyShowMore` → `showMore` (D-8). The compact card's copy and the term names are S13-35's and S13-34's. Each key is added in the step that first uses it, with its `Localizable.xcstrings` entry: `showMore` in STEP-2; `timelineAsRole`, `timelineBirthUnknown`, `timelineDeathUnknown` and `timelineUndated` in STEP-4 (the spoken labels); the rest in STEP-5.

## 7. Steps

**STEP-0 — Check the base**
Files: this plan.
Do: S13-33, S13-34 and S13-35 landed before this refresh, and every § 4 name was checked at `92f87db`. Run `git diff --stat 92f87db main -- macos/App/Platform/GenealogyStore.swift macos/App/Platform/PropertyTermDisplay.swift macos/App/DesignSystem/Recipes/PersonCard macos/App/DesignSystem/Components/TileCard macos/App/DesignSystem/Components/TitleText macos/App/DesignSystem/Components/DatePlaceLine macos/App/DesignSystem/Components/PageColumns macos/App/Features/Conclusions/EventsListView.swift macos/App/Features/Conclusions/PersonFamilySection.swift macos/App/Features/Workspace/WorkspaceLocation.swift`; for each file it lists, reread its § 4 row.
Check: the diff is empty, or each listed file's § 4 row is still true; otherwise stop and report the rows (Implementer block, item 2).

**STEP-1 — Page columns: full width and the visible rect**
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/App/DesignSystem/README.md` (row "PageColumns": `primaryFill`, `band`, `pvPageViewportSize`), `macos/App/Features/Workspace/WorkspaceLocation.swift` (`fillsWidth`, all false), `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/ProvenenciaTests/PVPageColumnsTests.swift`. Skill: `add-ui-component`.
Do: D-15 with no section opting in; D-13's `pvPageViewportSize`.
Check: T-8 passes; `PVPageColumnsTests` pass unchanged otherwise; the app builds.

**STEP-2 — One fold row and one section host**
Files: `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` *(new)*, `ConclusionPageSectionHost.swift` *(new)*, `PersonCardLink.swift` *(new)*, `PersonFamilyContent.swift`, `PersonFamilySection.swift`, `macos/ProvenenciaTests/PersonFamilyContentTests.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`, `add-workspace-place`.
Do: D-8, D-10 with Family as the only caller (the host's `@ViewBuilder` closure is stored as `content`, not `body`, round 3 X-42); `PersonFamilyContent.Tile` → `PersonCardLink` (D-5).
Check: `PersonFamilyContentTests`, `PersonFamilyLayoutTests` pass; Family's states keep their ids and look; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — The Timeline section's key and width**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift` (`sectionQueryKeys`: `.timeline` leaves the nil list), `PlaceRegistryTests.swift`.
Do: D-3; `fillsWidth` true for `.timeline` (round 2 X-16: until STEP-7 the Timeline's empty state is centred in the wider column, which reads the same).
Check: T-1, T-2 pass.

**STEP-4 — Items, lifeline geometry and fixtures**
Files: `macos/App/Features/Conclusions/PersonTimelineContent.swift` *(new)*, `PersonTimelineLifeline.swift` *(new)*, `PersonTimelinePreviewData.swift` *(new)*, `EventsListView.swift` (`spokenLine`, `placeTail`), `macos/App/Platform/GenealogyStore.swift` (`CatalogConclusionValue.productKey`), `macos/App/DesignSystem/Recipes/GenderTreatment/PVGenderTreatment.swift`, `macos/App/Platform/L10n.swift` and `macos/App/Resources/Localizable.xcstrings` (§ 6 Strings' STEP-4 keys), `macos/ProvenenciaTests/PersonTimelineContentTests.swift` *(new)*, `PersonTimelineLifelineTests.swift` *(new)*. Skill: `add-swift-test`, `add-localized-string`.
Do: D-1, D-2, D-4 – D-7, D-11, D-13's geometry.
Check: T-3 – T-7, T-9, T-10 pass; `PersonsListTests` and `PVGenderTreatmentTests` pass unchanged.

**STEP-5 — The section and its event rows**
Files: `macos/App/Features/Conclusions/PersonTimelineSection.swift` *(new)*, `EventsListView.swift` (`EventSecondaryLine.dateSize`), `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` (`lineLimit: Int?`), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-6, D-9, D-10's second caller, D-12, D-14, D-16; the items stack and the empty 80pt column; the previews.
Check: the app builds; T-11 passes; `PersonDetailTests` pass unchanged (the header's title).

**STEP-6 — The lifeline: drawing, tethers and clicks**
Files: `PersonTimelineSection.swift`, `macos/App/DesignSystem/Components/ScrollRowIntoView/PVScrollRowIntoView.swift` (`whenRequested`), `macos/App/DesignSystem/README.md` (row "ScrollRowIntoView"). Skill: `add-ui-component`.
Do: D-13's frames, drawing, tethers and clicks.
Check: the app builds; every existing `pvScrollsRowIntoView` caller's tests pass unchanged; in the "Full timeline" preview a click on a point scrolls to its row and a second click scrolls again.

**STEP-7 — Wiring**
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`.
Do: `.timeline` draws `PersonTimelineSection(entityId:)`.
Check: `scripts/check-pr.sh` passes; the manual check (§ 10).

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `design/archive/S13-D4-timeline-section.md` → `design/archive/`, `design/README.md`, `docs/macos-client-patterns.md` (§ 9 rule 2 reworded as **Rules this plan bends** says, and the opt-in full-width mode; the "One sectioned detail page" bullet: a section that reads its own query sits on `ConclusionPageSectionHost`), `.agents/skills/add-design-brief/SKILL.md` (the Paste block's rule 2, the same words, in the same commit). Skill: `add-design-brief`.
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
| T-5 | Swift | same file `rolesLocationsAndSpoken` | Event 8 → Badges Witness, Informant; a researcher's term keyed `subject` → neutral, not outline; all roles empty → no Badges and `spokenLabel` with no "as"; event 13 (no date, no place) → `spokenLabel` is its title alone, no empty part; the content and `spokenLine` built with `Locale(identifier: "en_US")`, so the expected strings do not depend on the machine; `CatalogConclusionValue.productKey` → the key for a product term, nil for a researcher's term and for a non-term; a participant's caption is its card content's `role`; Anna's card content spoken "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851"; card location `…withPageSection(.timeline)`; row location `EventsList.location`; event 1's `spokenLabel` "Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject"; the birth marker's `spokenLabel` equals its row's; `EventsList.spokenLine` has no ref; `EventsList.accessibilityLabel` unchanged | AC-1.4, AC-1.10, AC-2.1, AC-2.3, AC-2.4, D-1, D-2, D-4, D-5 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonTimelineLifelineTests.swift` *(new)* `pointsToScale` | `johann` with item frames summing to a 2730pt line → `length` 2730, the birth cap at 0, the burial's mark at 2730; 63.5 years → about 43pt a year → 1-year ticks; events 4 and 5 0.67 years apart (about 29pt); a day a dot, "1835" a bar, a range a bar, "bef. 1900" an open start; `lise` with a 4000pt line (98 years, about 41pt a year) → 1-year ticks; with a 1000pt line → 5-year ticks | AC-5.1, AC-5.2, AC-5.3, D-13 |
| T-7 | Swift | same file `placeholdersAndShortLives` | `unknownEnds` → two 40pt stubs outside the scale; a funeral and no death → the death stub before the funeral; `ida` → two caps joined; `ernst` with a 1200pt line → year labels and quarter ticks; a range under a year → drawn as a year; `unknownEnds` → the death stub inserted at event 9's `to`, the burial's mark 40pt lower than the scale alone gives; a recorded birth with no span → a filled cap on its own stub with no curve; a spec-exact one-day span → a dot; a taller line → a new scale; tethers only for items intersecting `visible` | AC-1.9, AC-5.4, AC-5.5, AC-5.6, AC-5.8, D-13 |
| T-8 | Swift | `macos/ProvenenciaTests/PVPageColumnsTests.swift` (cases added) | Pane 2296 `.wide` `primaryFill` → primary and group 2296 − 2 · gutterPage, `band` 1180; with a secondary → unchanged; `.regular` → unchanged; `primaryFill` false → primary 1180, `band == group`; `EnvironmentValues().pvPageViewportSize` → `.zero` | AC-6.1, AC-6.2, D-13, D-15 |
| T-9 | Swift | `PersonTimelineLifelineTests.swift` `oneHundredTwentyEvents` | 120 spans and 120 item frames → 120 points in item order in one pass | AC-3.2, D-13 |
| T-10 | Swift | same file `tickCountIsCapped` *(new)* | A placed range of 9999 years (spans −9999 and 9999 on a 4000pt line) → at most 200 ticks, with a step above 50 years; labels "-9000" style for negative years; `ida` (no span) → no tick, two caps, one stub | D-13 |
| T-11 | Swift | `macos/ProvenenciaTests/PersonTitleDisplayTests.swift` (case added) | `PVTitleText(.name("x")).lineLimit` is nil and a `.row` title draws one line (`lineLimit ?? 1`); `PVTitleText(.name("x"), style: .display(compact: true), lineLimit: 2).lineLimit` is 2 | D-16 |

Moved: the old T-1 (`PropertyTermDisplay.name(_:propertyKey:)`) and T-12 (one term rule in Vitals and the badge) to S13-34; the old T-2 (compact years, `helpText`, card sizes) to S13-35. AC-1.2 and the participant order are S13-33's tests. AC-1.7, AC-1.8, AC-2.2, AC-2.5, AC-2.6, AC-3.3, AC-4.1, AC-5.7 and AC-6.3 are covered by the manual check (§ 10); the fold reset on a new entity id (AC-3.3) is the section's `.onChange`, as Family's. Most likely to break: `PersonFamilyContentTests`, `PersonFamilyLayoutTests` (STEP-2), `PlaceRegistryTests` (STEP-3), `PersonsListTests`, `PVGenderTreatmentTests` (STEP-4).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `PersonFamilyMoreRow`, `PersonFamilyLoading`, `PersonFamilyContent.Tile`, `familyShowMore`, `conclusions.family.showMore`, `PersonTile`, `PersonTile.Size` and `One person tile, two sizes` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `macos/` with no hit left outside spike-13 plan and review history; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (rule 2 in § 9 and the add-design-brief Paste block in the same commit)
- [ ] UI matches § 6 (Design → macOS) and the brief's frames 12a – 12y; no new kit primitive but `primaryFill`, `pvPageViewportSize`, `whenRequested` and an optional `PVTitleText` `lineLimit`
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] docs/macos-client-patterns.md: rule 2, the full-width mode and the section host (STEP-8)
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
| `main` moves past `92f87db` under a § 4 file | STEP-0 |
| The tether pass on every scroll step is slow at 60 items | The scale is cached; the pass is linear and reads frames only; T-9; the "Long life" preview scrolls smoothly |
| The line's scale follows the list's height, so unfolding a list rescales it | Intended (the board's TL-5); T-7 |
| `primaryFill` changes another page | Off by default; only `.timeline` sets it (T-1, T-8) |

**Stop and report if:**

- STEP-0 finds that `CatalogPersonTimeline`, `CatalogTimelineSpan`, `CatalogLifeMarker`, `PVPersonCard`, `PVTileCardSize`, `PVDatePlaceLine`, `pvActivatesOnReturn`, `PVFocusRingButtonStyle`, `PVTitleText` or `PropertyTermDisplay.name(_:propertyKey:)` differs from § 4.
- The event row's title button is not reachable by Tab or VoiceOver inside the row's `.contain` group (D-16).
- `PVPageColumns` cannot publish its visible rect from the state it already measures, or cannot size its band apart from its columns (D-13, D-15).
- `pvScrollsRowIntoView` cannot scroll the page's own `ScrollView` from inside a section (D-13).
- A frame needs data the read does not return (an age, an estimated year): a follow-up, not Swift logic.

## 12. Open questions

None. The board's open items were answered by the user on 2026-11-02 (compact cards, participant order by role then name in Go, full width at Wide), as were the two round-2 questions (one term-name rule, now S13-34; the life span reused at year precision, now S13-34 and S13-35); the 72pt picture is kept as drawn (D-6).

## 13. PR

**Title:** S13-13 — Timeline section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md`.
