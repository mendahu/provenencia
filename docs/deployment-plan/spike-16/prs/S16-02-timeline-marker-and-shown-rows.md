# S16-02 — The Timeline places a missing death and builds what it shows

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 1a — Bounded reads, folds and the Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the placeholder marker, its dashed stub and the points are drawn as board S13-D4 draws them; only the row a placeholder sits beside changes, a bugfix back to R-24a; user, 2026-11-09)
**Board:** —
**Depends on:** S16-07 (landed [#761](https://github.com/mendahu/provenencia/pull/761))
**Blocks:** S16-19
**Branch:** `s16-02-timeline-marker-and-shown-rows`
**Version bump:** patch — a fix: a missing death no longer sits after a burial, a long Timeline builds only what it shows, and a date edit rescales the line; the wire change ships in the same bundle as both of its ends ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md) (the `LifeMarker` message), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) (the handler test), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (the lifeline's click targets, a Timeline snowflake)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [audit](../person-page-audit.md) P-1, P-7, P-20; working rules "(Batch 1) A marker's position is one value", "(Batch 1) The Timeline builds what it shows", "(Batch 1) One paged fold that moves focus"; [S16-07 plan](S16-07-paged-fold-focus.md) (its Landed note and D-10); brief [S16-D2](../design/archive/S16-D2-person-page-folds.md) FH-5; Spike 13 [decisions](../../archive/spike-13/README.md#decisions) (the Timeline is a lifeline; its rules and R-24a are in git history)
**Planned:** 2026-11-09 against `main` @ `462699f`
**Review:** [`S16-02-timeline-marker-and-shown-rows.review.md`](S16-02-timeline-marker-and-shown-rows.review.md) — Fits with changes — Not ready → fixed, 1 round; no open blocker; To user answered (M-1, X-5)

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

A person with no recorded death and a dated burial or funeral reads "Death unknown" just before that burial, in the list and on the lifeline, so the life between a baptism and a burial is drawn before the death (P-1; user, 2026-11-08: "just before burial"). A long Timeline builds rows and person cards only for what its folds show; folded points stay on the lifeline and share one click target that opens the fold through the clicked point's page (P-7; user, 2026-11-09). A date edit from another window rescales the lifeline instead of keeping the old points per year (P-20).

**Why now.** Bar 1 and bar 3 of the spike. S16-07 landed the paged fold this PR builds against, and S16-19 rewrites `core/database/persondetail/timeline.go` after this PR.

**What stays.** The fold limits and pages (S16-07), the row and card layout, the placeholder's text ("Birth unknown", "Death unknown"), the stub's look and length (40pt), every event's point drawn on the line, a shown row's point keeping its own click target and hover name, FH-5's behavior (a folded point opens the dated fold through its page and scrolls to it), the placement of rows against the line.

## 2. User stories

**US-1.** As a researcher, when a person has no dated death but a dated burial or funeral, I see "Death unknown" just before the burial, in the list and on the lifeline, so that the life reads in order.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Only a burial dated (`b1`, 1850), two undated events `i1`, `i2`, no birth or death recorded | List: Birth unknown, Undated, `i1`, `i2`, Death unknown, `b1`. Go: birth before `b1`, death before `b1` |
| AC-1.2 | A dated baptism (1820) and a dated burial (1850), no death recorded | Death unknown sits after the baptism and before the burial in the list; its stub on the line ends at the burial's start, with the death cap at the stub's top, not after the baptism's end |
| AC-1.3 | Two dated burials (1850, 1852) | Death sits before the 1850 one |
| AC-1.4 | A recorded death event with no date, and a dated funeral | The recorded death row sits just before the funeral, as a marker with a stub |
| AC-1.5 | No burial or funeral dated; a census (1851) is the last dated row | Death sits after the census, its stub after the census's end (today's behavior) |
| AC-1.6 | Nothing dated | Birth first, death last, the line two caps and one dashed stretch (today's behavior) |
| AC-1.7 | A type listed as before-birth is dated before the first other row | A placeholder birth sits just after the last dated before-birth row (the mirror; no shipped type is before-birth today) |

**US-2.** As a researcher with a person who has thousands of events, I open the Timeline without a stall, so that a crafted or huge project does not freeze the page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | 2,000 dated events, each with 500 participants, at Wide, folded | The content holds 50 event rows and at most 12 cards on each; no `Row` and no card exists for the other 1,950 events; `allSpans` holds 2,000 spans |
| AC-2.2 | Same, the lifeline | 2,000 points drawn; one click target per shown item with a point (the 50 shown rows and the birth and death caps) plus one shared target for the folded points |
| AC-2.3 | Click on the shared target nearest the 180th event's point | The dated fold opens through that event's page (rows 1–200 shown) and the list scrolls to row 180 (FH-5, unchanged) |
| AC-2.4 | Hover a shown point | Its row's spoken label shows as the tooltip; a folded point has no tooltip |

**US-3.** As a researcher, when another window changes an event's date, the lifeline takes a new scale for the new range, so that a moved event does not stretch the line many times its usual height.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The same items and viewport, one event's year moved from 1850 to 1900 | The kept scale does not apply; a new one is chosen |
| AC-3.2 | The same items, viewport and years | The kept scale applies |

## 3. Scope

| In | Covered by |
| --- | --- |
| Go: a life marker names one place, after or before one dated row; a placeholder death sits before the earliest dated after-death row, else after the last dated row; the birth mirror (P-1) | STEP-1, AC-1.1 – AC-1.7 |
| Wire: `LifeMarker` carries a `oneof place` (`after_event_id`, `before_event_id`); `beside_event_id` goes | STEP-2 |
| Swift: `CatalogLifeMarker.Place`; the content inserts a marker after or before its row; the lifeline's stub follows the place (P-1) | STEP-3, STEP-4, AC-1.1, AC-1.2 |
| The content builds `Row`s and cards only for shown items; folded rows keep id and span (P-7) | STEP-5, AC-2.1 |
| The lifeline: one click target per shown point, one shared target for the folded points (P-7; user, 2026-11-09) | STEP-6, AC-2.2 – AC-2.4 |
| The kept scale follows the line's year range too (P-20) | STEP-7, AC-3.1, AC-3.2 |
| The working rule "The Timeline builds what it shows" updated to the user's answer | STEP-8 |

| Out | Home |
| --- | --- |
| Memoising the content, participants once per read, lazy cards after an unfold | [person page performance idea](../../../ideas/person-page-performance.md) (S16-27 memoises section content) |
| Folding the dated rows after the death marker (a crafted person with thousands of their own dated burials builds every one) | [person page performance idea](../../../ideas/person-page-performance.md) (user, 2026-11-09) |
| Adding a before-birth event type to `lifeline.beforeBirth` | not planned |
| Event and participant fields moving onto summary rows | S16-19 |
| Back restoring a section's folds (P-19) | [person page follow-ups idea](../../../ideas/person-page-followups.md) |
| A Go cap on a person's events | not planned (rule "Crafted fan-out is capped, not engineered" names no events cap; the fold bounds what is built) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Marker rule | `core/database/persondetail/lifeline.go` `var lifeline` (`afterDeath`: burial, funeral; `beforeBirth`: empty), `type LifeMarker { Event, Beside []byte }`, `lifeMarkers(events, types, birth, death)` | Birth beside the first dated row not before-birth; death beside the last dated row not after-death; `Beside` nil both for "nothing dated" and "every dated row is after-death" (P-1) |
| Marker tests | `core/database/persondetail/lifeline_test.go` `TestLifeMarkers` | Pins "only a burial dated" → death `nil`; "a funeral after a census" → death beside the census |
| Read tests | `core/database/persondetail/timeline_test.go` `TestPersonEventsLifeline`: a `markers{birthEvent, birthBeside, deathEvent, deathBeside}` struct, a `check` helper reading `tl.Birth.Beside` / `tl.Death.Beside`, and twelve subtests (AC-2.1 – AC-2.9, "a researcher's burial term is not the product's", "lead birth is an open-start date", "provisional earlier birth and accepted later birth"); AC-2.9 reads `tl.Birth.Beside` directly | Pins today's beside rule end to end; AC-2.3 pins death beside the census before a funeral (P-1) |
| Timeline read | `core/database/persondetail/timeline.go` `PersonEvents`, `Timeline{Events, Birth, Death}` | Calls `lifeMarkers(rows, typeOf, birth, death)` last; rows in date order, undated last |
| Wire | `api/proto/engine.proto` `ListPersonEventsResponse` (comment lines 1564–1574), `LifeMarker { event_id = 1; beside_event_id = 2; }` | One string for both positions |
| Handler | `api/ffi/handlers/person_detail.go` `lifeMarkerProto` | Maps `Beside` to `BesideEventId` |
| Handler test | `api/ffi/handlers/person_detail_test.go` `TestListPersonEvents` | Asserts the death placeholder is beside the census (`cs`, 1851), with the person a witness at a burial (`u`, 1890); its subtest with person `lee` (a dated birth `born`, an undated residence) asserts birth `GetBesideEventId() == ""` and death `GetBesideEventId() == born` |
| Swift model | `macos/App/Platform/GenealogyStore.swift` `CatalogLifeMarker { eventId, besideEventId }` | |
| Swift mapping | `macos/App/Platform/GoStore.swift` `listPersonEvents`, `mapLifeMarker` | Empty string → nil |
| Content | `macos/App/Features/Conclusions/PersonTimelineContent.swift` `init(timeline:folds:locale:orders:widthClass:)` | Builds a `Row` (title, line, roles, location, spoken label, cards) for every event (the `rows` map at the top of `init`), and `revealed[.participants]` for every event (the `for event in timeline.events` loop); inserts a placeholder birth before its `besideEventId`, a death after it, else first and last (steps (2), (3)); then folds lists of built `Item`s (`before`, `after`, `counted`, `hiddenIDs`, the `compactMap` that turns the hidden run into one `.more`); `Span` carries `spokenLabel`; `pointLabels` names every span, folded or not. `Marker(kind:row:spokenLabel:)` is built by the local `marker(_:_:)` |
| Lifeline | `macos/App/Features/Conclusions/PersonTimelineLifeline.swift` `init(content:pointsPerYear:)`, `Insertion`, `map(_:)`, `nextRowID(after:in:folded:)`, `previousRowID(before:in:folded:)` | A birth stub at the next row's `from` (years `>= at` shift), a death stub at the previous row's `to` (years `> at` shift); the stub's neighbor is found by walking items. Birth cap at the stub's top, death cap at its bottom |
| Line view | `macos/App/Features/Conclusions/PersonTimelineSection.swift` `PersonTimelineLine` (`ForEach(lifeline.marks, id: \.id)`, a `Button` with `.help(labels[mark.id] ?? "")` per mark) | A button per point, folded ones included (P-7) |
| Kept scale | same file, `PersonTimelineList.ChosenScale { ids, viewportHeight, pointsPerYear }`, `rebuild()` | Kept while ids and viewport height match (P-20) |
| Placement | `macos/App/Features/Conclusions/PersonTimelinePlacement.swift` `init(content:heights:viewportHeight:pointsPerYear:)` | Takes a fixed scale or chooses one |
| Fixtures | `macos/App/Features/Conclusions/PersonTimelinePreviewData.swift` `unknownEnds` (`besideEventId: "2"` / `"9"`); `PersonTimelineSection.swift` preview "Row sheet" (`besideEventId: "4"` / `"11"`). `johann`, `ernst`, `largeEvent`, `lise`, `ida` use `eventId` or empty markers only | The two named fixtures are written in `besideEventId` |
| Tests | `macos/ProvenenciaTests/PersonTimelineContentTests.swift` (`itemsInOrder`, `markerInsideHiddenRunStaysShown`, `foldedMarkersStayAfterPagedRun`, `pointsAreNamedByTheirRow` and others), `PersonTimelineLifelineTests.swift` (`placeholdersAndShortLives`, its funeral case with `besideEventId: "c"` for both markers, the recorded-birth case `CatalogLifeMarker(eventId: "1", besideEventId: "2")`, `oneHundredTwentyEvents`), `PersonTimelinePlacementTests.swift`, `CatalogQueryRegistryTests.swift` (lines 1586–1598, a `CatalogLifeMarker(eventId:besideEventId:)` round trip and the defaults) | `rg besideEventId macos` lists every site STEP-3 rewrites |

**Precedent.** S16-01 ([#756](https://github.com/mendahu/provenencia/pull/756)) changed a Person page read's wire message, regenerated with `./scripts/generate-proto.sh`, and mapped it in `GoStore`: copy its shape for the field change. A wire `oneof` is mapped in `GoStore.mapConclusionSummary` (`switch summary.kind { case .person(let p): … case nil: nil }`): copy that switch for `m.place`. S16-07 ([#761](https://github.com/mendahu/provenencia/pull/761)) D-10 added `pages(toShow:)` and `select`: keep both as they are.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Go: `type MarkerSide int` *(new)* with `SideAfter MarkerSide = iota + 1` and `SideBefore` (the zero value is no side); `type MarkerPlace struct { Row []byte; Side MarkerSide }` *(new)*; `LifeMarker` becomes `{ Event []byte; Place *MarkerPlace }`, and `Beside` goes. `Place` is nil when the marker's event has a span, or when no row is dated | Rule "A marker's position is one value"; nil now means one thing per case, and "nothing qualifies" cannot happen | Two `[]byte` fields (both could be set) |
| D-2 | The rule, in `lifelineTypes.markers` *(new; was `lifeMarkers`)*: a birth with no span sits **after** the last dated row whose type is before-birth when there is one, else **before** the first dated row; a death with no span sits **before** the first dated row (in the read's order, which is date order) that is after-death when there is one, else **after** the last dated row. A row is after-death (or before-birth) when its type and one of the person's roles on it make an entry of the list: `{burial, subject}`, `{funeral, subject}` (a role matches when `Term.ProductKey()` is the role's key). A burial the person only witnessed is an ordinary row | Working rule; user, 2026-11-08 ("just before burial") and 2026-11-09 (only the person's own burial or funeral: a witness was alive at it) | Spike 13's "just after the latest event that cannot come after death" (P-1) |
| D-3 | `var lifeline` becomes a value of `type lifelineTypes struct{ afterDeath, beforeBirth []subjectcompare.EventRole }` *(new)*, `afterDeath: {{EventBurial, RoleSubject}, {EventFuneral, RoleSubject}}`, with the method `markers(events, types, birth, death) (Birth, Death LifeMarker)`; a private `matches(list, row PersonEvent, t vocabkey.EventType) bool` *(new)* reads `row.Roles`. `PersonEvents` calls `lifeline.markers(…)`. The test builds its own `lifelineTypes` for the before-birth mirror. The list stays separate from `matchregistry` `Lifespans[Person].Ends` (which says where a life ends for matching: death or burial, no funeral) but shares its type | The mirror is testable without mutating a package variable under parallel tests; `subjectcompare.EventRole` is the codebase's one event-and-role shape (review M-1) | Testing the mirror by editing `lifeline` in the test |
| D-4 | Wire: `message LifeMarker { string event_id = 1; reserved 2; reserved "beside_event_id"; oneof place { string after_event_id = 3; string before_event_id = 4; } }`. The `ListPersonEventsResponse` comment states D-2. `lifeMarkerProto` sets one of the two from `Place` | Rule: one `oneof`; field 2 is reserved, not reused, because its old meaning (beside, before for a birth) differs from "after" | Reusing number 2 |
| D-5 | Swift: `CatalogLifeMarker { var eventId: String?; var place: Place? }` with `enum Place: Sendable, Equatable { case after(eventId: String), before(eventId: String) }` *(new)*. `GoStore.mapLifeMarker` switches on `m.place`: an empty id or `nil` maps to `nil` | Rule: one Swift enum | Two optionals |
| D-6 | Content insertion: a marker whose row has no span is placed in the dated items at `.before(id)` → the index of `id`, `.after(id)` → that index + 1, nil → birth at 0, death at the end. Birth is placed first. A death placed `.after(id)` steps past a birth marker that sits right after `id`, so the birth never follows the death. An `id` not among the dated items counts as nil. The content resolves the place once: `Marker.place` (D-7) holds the place as resolved, nil when its id was not a dated row, and the lifeline reads it as given | Bar 1; "Swift places what Go names" | Swift deciding the place from types |
| D-7 | `PersonTimelineContent.Marker` gains `var place: CatalogLifeMarker.Place? = nil` (the resolved place of a marker with a stub, D-6; nil for a marker whose row has a span, and for a place whose id is not a dated row); the default keeps `Marker(kind:row:spokenLabel:)` calls compiling | The lifeline reads the place instead of walking items | Re-deriving the neighbor in the lifeline |
| D-8 | Lifeline stubs: an `Insertion` is `{ item, at, isBefore }`: `.before(id)` → `at` = that row's `from`, `isBefore` true; `.after(id)` → that row's `to`, false; nil → birth at `range.from` (before), death at `range.to` (after); the lifeline reads `Marker.place` and resolves nothing itself. The spans come from `content.allSpans`, which holds folded rows. The k-th stub in item order spans `[base(at) + 40k, base(at) + 40k + 40]` with `base(year) = pad + (year − range.from) × pointsPerYear`; its cap sits at the stub's top when `isBefore`, else at its bottom. `nextRowID(after:in:folded:)` and `previousRowID(before:in:folded:)` go | One rule for both kinds, keyed on the place; reproduces today's geometry for a birth before and a death after (T-6) | Keying the cap on the kind (a death before a burial would cap at the burial's point) |
| D-9 | Lifeline points shift by the stubs before them **in item order**: a dot or bar of an item after k stubs is drawn at `base(year) + 40k` for both ends. The `.more(.dated)` item stands for every folded point at its place. Ticks shift by year: a `before` stub moves years `>= at`, an `after` stub years `> at` | The list and the line agree on order even when a baptism and a burial share a year (an infant's death; §6 Edges); ticks have no item | Year comparison for points (a same-year baptism would be drawn below the death stub) |
| D-10 | Building: the content lays out ids first, then builds. A private `enum Slot: Equatable { case event(String), marker(Marker.Kind, rowID: String?), undatedHeader, more(Fold, count: Int) }` *(new)* stands for an `Item` before it is built and owns the one id rule: `Slot.id` is the event id for `.event`, `rowID ?? "birth"/"death"` for `.marker`, "undated", and "more.dated" / "more.undated" (today's `Item.id` strings); `Slot.isEvent` is true for `.event` and for a `.marker` with a `rowID` (today's `eventRow != nil`). `Item` gains a private `slot: Slot` computed property (`.row(r)` → `.event(r.id)`, `.marker(m)` → `.marker(m.kind, rowID: m.row?.id)`, the other two as they are), and `Item.id` becomes `slot.id`, so the fold ids and the item ids have one source. A marker's `rowID` is the recorded event's id when it is one of the events, else nil. Pass 1 builds `allSpans` (id and span only) and a `[String: CatalogPersonEvent]` by id. Pass 2 runs today's steps (1)–(4) and the folds unchanged in meaning on `[Slot]`: `dated`, `before`, `after`, `counted` (the slots with `isEvent`), `datedFoldIDs`, `hiddenIDs`, `deathAt`, the `compactMap` to one `.more(.dated, …)`, the undated fold. A recorded marker counts as dated only when its event has a span, as today. Pass 3 maps each slot of the final list to an `Item`: `.event(id)` → `.row(Self.row(events[id]!, …))`; `.marker(kind)` → a `Marker` whose `row` is `Self.row(…)` of its recorded event (nil for a placeholder) and whose `place` is D-7's; and sets `revealed[.participants(eventId:)]` for each built row. `Self.row` stays `static` and is called once per shown event. `revealed[.dated]` and `revealed[.undated]` are the slot ids `firstRevealed` returns, as today. `isShown(_:)` and `row(holdingCard:)` keep reading `items`. Folded event rows and their participants are never read past their id and span | Rule "The Timeline builds what it shows"; P-7 | Building every row and dropping the hidden ones (today) |
| D-11 | `PersonTimelineContent.Span` drops `spokenLabel`. `pointLabels` goes; `Item.spokenLabel: String?` *(new, beside `eventRow`)* is a row's or a marker's spoken label, nil for the Undated heading and a Show more. D-12's targets carry their label from it, so a folded point has no name (user, 2026-11-09) and `PersonTimelineLine` loses its `labels` parameter | A folded row builds nothing past id and span | Building a spoken label for 2,000 folded rows |
| D-12 | Click targets: `PersonTimelineLifeline` gains `struct Target: Equatable { var mark: Mark; var label: String }` *(new)* and `targets: [Target]` *(new)*, one per item in `content.items` that has a mark, in item order, labeled with `Item.spokenLabel`, and `foldedHit: ClosedRange<CGFloat>?` *(new)*, the folded marks' lowest `hit.lowerBound` through their highest `hit.upperBound`, nil when nothing is folded, where `Mark.hit: ClosedRange<CGFloat>` *(new, beside `pointY`)* is `y − 6 … y + 6` for a dot or cap and `from − 6 … max(to + 6, from + 6)` for a bar (today's button extents); and `foldedMark(nearest y: CGFloat) -> String?` *(new)*, the folded mark whose `pointY` is closest to `y` in the line's space (the earlier in item order on a tie), nil when nothing is folded. `PersonTimelineLine` changes its `ForEach(lifeline.marks, …)` to `ForEach(lifeline.targets, id: \.mark.id)` with the same `Button`, framed `PersonTimelineDrawing.hit` × `hit.upperBound − hit.lowerBound` and offset to `y: target.mark.hit.lowerBound` (from `Mark.hit`; `PersonTimelineLine.top(_:)` and `hitHeight(_:)` go), and `.help(target.label)`, and adds, before that `ForEach` in the `ZStack` (so the buttons sit on top), `if let hit = lifeline.foldedHit { Color.clear.frame(width: PersonTimelineDrawing.hit, height: hit.upperBound - hit.lowerBound).contentShape(Rectangle()).onTapGesture(coordinateSpace: .local) { location in if let id = lifeline.foldedMark(nearest: hit.lowerBound + location.y) { select(id) } }.offset(x: lineX - PersonTimelineDrawing.hit / 2, y: hit.lowerBound).accessibilityIdentifier("persons.detail.timeline.folded") }`, with no `.help` | User, 2026-11-09: one shared target keeps FH-5; a click target per shown row (rule) | A button per folded point (P-7); no target for folded points (drops FH-5) |
| D-13 | Kept scale: `PersonTimelinePlacement.KeptScale` *(new)* `{ ids: [String]; years: ClosedRange<Double>?; viewportHeight: CGFloat; pointsPerYear: CGFloat }` with `init(content:viewportHeight:pointsPerYear:)` (years from `PersonTimelineLifeline.range(content)`) and `func pointsPerYear(for content:, viewportHeight:) -> CGFloat?`, non-nil only when ids, years and viewport height all match. `PersonTimelineList.ChosenScale` goes. `PersonTimelineList` holds `@State private var keptScale: PersonTimelinePlacement.KeptScale?`; `rebuild()` reads `let fixed = keptScale?.pointsPerYear(for: content, viewportHeight: viewportSize.height)` and, after placing, sets `keptScale` to `KeptScale(content: content, viewportHeight: viewportSize.height, pointsPerYear: next.pointsPerYear)` when every item is measured and the items are not empty, else nil (today's rule), writing only when the value changes | P-20, in pure code a test reaches | Keying on the content value (any reload would rescale) |
| D-14 | The handler test's expectation changes: with the person a witness at a dated burial `u` (1890), that burial is an ordinary row, so the death placeholder sits after it (`after_event_id` = `u`), not beside the census | D-2: only the person's own burial counts; the witness was alive in 1890 | Counting any role (the death would be drawn before a burial the person attended) |
| D-15 | `TestPersonEventsLifeline`'s `markers` struct becomes `{ birthEvent []byte; birth *persondetail.MarkerPlace; deathEvent []byte; death *persondetail.MarkerPlace }` with helpers `before(id)` and `after(id)` in the test, and `check` compares `Event` bytes and the place (nil, or side and row). Expectations under D-2: AC-2.1 both nil (recorded with spans); AC-2.2 birth before C, death after M; AC-2.3 birth before C, death **before F** (was beside C); AC-2.4 both nil; AC-2.5 birth event B, before C, death after C; AC-2.6 birth nil, death after B2; "a researcher's burial term" birth before C, death event D, after U (a researcher's term has no product key, so U is not after-death; was beside U); "lead birth" birth nil, death after Later; "provisional" both nil; AC-2.7 birth before B, death after B; AC-2.8 birth before D, death nil; AC-2.9 birth before C | Every subtest gets an expectation; only AC-2.3 changes its meaning | Leaving the rewrite to the implementer |

**Rules this plan bends.**

- The working rule "(Batch 1) A marker's position is one value" says "beside or before" and "the earliest dated after-death row (burial, funeral)" → the code says after or before ("beside" meant before for a birth and after for a death), and only the person's own burial or funeral counts → user, 2026-11-09: subject only (review M-1); the after/before wording is review M-3 → STEP-8 rewrites the rule.
- The working rule "(Batch 1) The Timeline builds what it shows" says a folded row keeps its spoken label and the lifeline makes a click target only for a shown row → a folded row keeps only its id and span, and folded points share one click target, because S16-D2's FH-5 (agreed 2026-11-09, after the rule) makes folded points clickable → user, 2026-11-09: one shared target for the folded points (D-11, D-12) → STEP-8 rewrites the rule in `deployment-plan.md`.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `MarkerSide`, `MarkerPlace` *(new)* | type | `persondetail.LifeMarker.Beside` | Replaces it (D-1); searched `Side`, `Before`, `After` in `core/database/persondetail` |
| `lifelineTypes` *(new)* with `markers`, `matches` | type + method | `var lifeline` + `lifeMarkers`; `subjectcompare.EventRole`, `matchregistry` `Lifespans[Person].Ends` | The same rule, made a value with the existing event-and-role shape; `Ends` answers another question (D-3) |
| `Item.spokenLabel` *(new)* | property | `pointLabels` | Replaces the stored index (D-11) |
| `LifeMarker.place` oneof, `after_event_id`, `before_event_id` *(new)* | proto | `beside_event_id` | Replaces it (D-4) |
| `CatalogLifeMarker.Place` *(new)* | type | `CatalogLifeMarker.besideEventId` | Replaces it (D-5) |
| `PersonTimelineContent.Marker.place` *(new)* | property | `nextRowID` / `previousRowID` in the lifeline | Replaces the walk (D-7, D-8) |
| `Slot` *(new, private)* | type | `PersonTimelineContent.Item` | An unbuilt item; `Item.id` reads `Slot.id`, so one id rule (D-10) |
| `PersonTimelineLifeline.Target`, `targets`, `foldedHit`, `foldedMark(nearest:)`, `Mark.hit` *(new)* | type / property / func | `PersonTimelineLine.top(_:)`, `hitHeight(_:)` (moved into `Mark.hit`), | `PersonTimelineLifeline.marks`, `mark(_:)` | `marks` stays the drawing list; these are the click list (D-12) |
| `PersonTimelinePlacement.KeptScale` *(new)* | type | `PersonTimelineList.ChosenScale` | Moves it to the pure placement file and adds the years (D-13) |

**Builds on.** S16-07's `ConclusionFold` and `ConclusionFoldState`, `pages(toShow:)` and `select` (D-10 there); the existing `CatalogLifeMarker` model and `GoStore` mapping; the `oneof` mapping shape in `GoStore.mapConclusionSummary`.
**Deletes.** `persondetail.LifeMarker.Beside`; `lifeMarkers` (becomes `lifelineTypes.markers`); proto `beside_event_id` (reserved); `CatalogLifeMarker.besideEventId`; `PersonTimelineLifeline.nextRowID(after:in:folded:)` and `previousRowID(before:in:folded:)`; `PersonTimelineContent.Span.spokenLabel`; `PersonTimelineContent.pointLabels`; `PersonTimelineLine`'s `labels` parameter, `top(_:)` and `hitHeight(_:)`; `PersonTimelineList.ChosenScale`; the per-event `revealed[.participants]` loop over all events.
**Next adds.** S16-19 reads participants from summary rows inside the same `PersonEvents` and keeps `lifeline.markers` as is; S16-27 keys derived values on the content, which now costs the shown rows only; a before-birth type is one `EventRole` entry in `lifelineTypes.beforeBirth`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A death marker before a row (a new state) | `PersonTimelineContent.init`, `PersonTimelineLifeline.init`, `PersonTimelinePlacement` | Content: D-6 (T-4); lifeline: D-8 (T-6); placement anchors a placeholder at half its height, unchanged |
| A place naming an id Swift does not have | `PersonTimelineContent.init` D-6 | Counts as nil: birth first, death last (T-4) |
| A place naming a folded row (a death after the last dated row past the 50th) | D-8 | `at` comes from `allSpans`, which holds folded rows (T-6) |
| A click on the folded region with no folded mark near | `foldedMark(nearest:)` | Always returns one when `foldedHit` exists; `select` opens through that page (T-7) |

### Scale

At the map's Scale (crafted: 2,000 events, 500 participants on each; real: 200 events, 10 participants each): pass 1 and pass 2 read only ids and spans, O(events); pass 3 builds at most 50 dated rows + 50 undated rows + 2 markers, each with at most 12 cards and one `firstRevealed` per row. `allSpans` holds every dated span (2,000 values, no strings built). The lifeline draws 2,000 marks in one `Canvas`, as today, and holds 50 `Button`s plus one region (was 2,000). `foldedMark(nearest:)` is one linear pass over the folded marks per click. Real data: 200 events fold to 50 rows; each row built as today, so a real person's page builds a quarter of today's rows. `KeptScale.pointsPerYear(for:)` compares the ids array, as `ChosenScale` does today, plus two doubles.

### Edges

- **Assumptions.** D-2 assumes `PersonEvent.Roles` holds the person's roles on the row, which `PersonEvents` fills before the markers (S16-19 keeps that). D-6 assumes the read's order is date order, so the first dated after-death row is the earliest; `PersonEvents` orders by span (test `TestListPersonEvents` `wantOrder`). D-9 assumes item order is line order for dated items, which D-6 keeps by inserting markers among the dated rows. The fold assumes the hidden rows are a suffix of the counted rows (S16-07 T-4); a death `.before` an after-death row puts that row after the death marker, outside the counted rows, so it is never folded.
- **Neighbors.** A dated baptism and burial (death before the burial, T-1, T-4); only a burial (birth and death both before it, T-1, T-4: `["birth", "undated", "i1", "i2", "death", "b1"]`); a census then a funeral (death before the funeral, T-1); no after-death row (death after the last row, T-1, today's behavior); a baptism and a burial in the same year (both points below the birth stub, the baptism above the death stub, the burial below it: T-6); a recorded death with no date and a dated funeral (T-1, T-4); birth and death both `.after` the same row (birth then death, D-6: T-4).
- **Removed inputs.** A person a witness at another person's burial: that burial is an ordinary row, so the placeholder death sits after it (D-14; today it sat after the census, before the burial). A person both subject and witness on one burial: subject counts (T-1). An after-death row that is undated is not dated and is ignored, as today. A date edit that moves the anchor row: Go re-sends the place with the read; the kept scale resets (D-13).
- **Real data.** `lifeline.afterDeath` is burial and funeral; `beforeBirth` is empty, so a real birth is always `.before` the first dated row or nil. The seeded event types in `core/database/subjectvocab/registry.go` include baptism; it is not before-birth, so a baptism dated before a recorded birth's year is ordered by date and the birth sits before it as today.
- **Untrusted size.** Events per person have no Go cap (Out); the fold bounds what Swift builds to 50 dated + 50 undated rows per page, and cards per row to 5 / 12 per page (S16-07). The lifeline's marks grow with events in one `Canvas`; click targets do not.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/persondetail` (`lifeline.go`) | `type MarkerSide int` *(new)*, `SideAfter`, `SideBefore`; `type MarkerPlace struct { Row []byte; Side MarkerSide }` *(new)*; `LifeMarker { Event []byte; Place *MarkerPlace }`; `type lifelineTypes struct{ afterDeath, beforeBirth []subjectcompare.EventRole }` *(new)*, `var lifeline = lifelineTypes{afterDeath: …}`, `matches` *(new)*, `func (l lifelineTypes) markers(events []PersonEvent, types func([]byte) vocabkey.EventType, birth, death []byte) (Birth, Death LifeMarker)` (D-2) |
| `core/database/persondetail` (`timeline.go`) | `PersonEvents` calls `lifeline.markers(rows, typeOf, birth, death)`; its doc comment says "placed after or before a dated event" |

Errors: none new.

### FFI

`api/proto/engine.proto`: `LifeMarker` per D-4; the `ListPersonEventsResponse` comment restates D-2 ("a placeholder death sits before the earliest dated burial or funeral of the person's own (role subject), else after the last dated event; a placeholder birth sits after the last dated event that may precede birth, else before the first dated event; place is unset when the marker has its own span or nothing is dated"). Handler `api/ffi/handlers/person_detail.go` `lifeMarkerProto` sets `Place` as `&engine.LifeMarker_AfterEventId{…}` or `&engine.LifeMarker_BeforeEventId{…}`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Marker model | `CatalogLifeMarker`, `CatalogLifeMarker.Place` | `macos/App/Platform/GenealogyStore.swift` | D-5 |
| Mapping | `GoStore.mapLifeMarker` | `macos/App/Platform/GoStore.swift` | D-5 |
| Content | `PersonTimelineContent` (snowflake model) | `macos/App/Features/Conclusions/PersonTimelineContent.swift` | D-6, D-7, D-10, D-11 |
| Lifeline geometry | `PersonTimelineLifeline` | `macos/App/Features/Conclusions/PersonTimelineLifeline.swift` | D-8, D-9, D-12 |
| Line view | `PersonTimelineLine` (snowflake) | `PersonTimelineSection.swift` | D-12: buttons over `targets`, one region over `foldedHit` |
| Kept scale | `PersonTimelinePlacement.KeptScale` | `PersonTimelinePlacement.swift` | D-13 |

```swift
struct CatalogLifeMarker: Sendable, Equatable {
    enum Place: Sendable, Equatable { case after(eventId: String), before(eventId: String) }
    var eventId: String? = nil
    var place: Place? = nil
}

extension PersonTimelineLifeline {
    struct Target: Equatable { var mark: Mark; var label: String }
    var targets: [Target] { get }
    var foldedHit: ClosedRange<CGFloat>? { get }
    func foldedMark(nearest y: CGFloat) -> String?
}

extension PersonTimelinePlacement {
    struct KeptScale: Equatable {
        init(content: PersonTimelineContent, viewportHeight: CGFloat, pointsPerYear: CGFloat)
        func pointsPerYear(for content: PersonTimelineContent, viewportHeight: CGFloat) -> CGFloat?
    }
}
```

States: an empty timeline is unchanged; nothing dated draws the two caps (AC-1.6). Accessibility: the line stays hidden from VoiceOver; the rows carry the meaning; the folded region is not focusable.

### Strings

None — no key added or changed; "Birth unknown" and "Death unknown" are unchanged.

## 7. Steps

**STEP-1 — Go places a marker after or before one row**
Files: `core/database/persondetail/lifeline.go`, `lifeline_test.go`, `timeline.go`, `timeline_test.go`.
Do: D-1, D-2, D-3, D-15. First add T-1's cases "only a burial dated" and "a baptism and a burial" against today's code and see them fail.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/...` passes with T-1 and T-9; `rg '\.Beside' core` finds nothing.

**STEP-2 — The wire carries one place**
Files: `api/proto/engine.proto`, `api/ffi/handlers/person_detail.go`, `api/ffi/handlers/person_detail_test.go`, the generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`, `use-catalog-session`.
Do: D-4, D-14. Run `./scripts/generate-proto.sh` (needs `protoc`, `protoc-gen-go` and `protoc-gen-swift`, as on the Mac the repo's README sets up).
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes with T-2; `rg beside_event_id api core` finds only the `reserved` line.

**STEP-3 — Swift reads the place**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `macos/App/Features/Conclusions/PersonTimelinePreviewData.swift`, `PersonTimelineSection.swift` (the "Row sheet" preview), `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `PersonTimelineContentTests.swift`, `PersonTimelineLifelineTests.swift`. Skill: `add-swift-test`.
Do: D-5. Rewrite every `besideEventId:` as the place Go sends under D-2 for that fixture's data:

| Site | Birth | Death |
| --- | --- | --- |
| `PersonTimelinePreviewData.unknownEnds` | `.before(eventId: "2")` | `.before(eventId: "11")` (its burial) |
| `PersonTimelineSection.swift` preview "Row sheet" | `.before(eventId: "4")` | `.before(eventId: "11")` |
| `PersonTimelineLifelineTests.placeholdersAndShortLives`, the funeral case | `.before(eventId: "c")` | `.before(eventId: "f")` |
| the same test, the recorded-birth case `eventId: "1"` | `eventId: "1", place: .before(eventId: "2")` | — |
| `PersonTimelineContentTests.itemsInOrder`, the recorded birth | `eventId: "1", place: .before(eventId: "2")` | — |
| `PersonTimelineContentTests.markerInsideHiddenRunStaysShown` | `.before(eventId: "d55")` (a Swift placement case Go does not produce; kept) | unchanged |
| `PersonTimelineContentTests.foldedMarkersStayAfterPagedRun` | `.before(eventId: "d110")` (kept, as above) | unchanged |
| `CatalogQueryRegistryTests` round trip | `CatalogLifeMarker(eventId: nil, place: .before(eventId: "ev1"))` | `CatalogLifeMarker(eventId: "ev1", place: nil)`; the defaults compare with `CatalogLifeMarker(eventId: nil, place: nil)` |

Check: the app builds; `rg besideEventId macos` finds nothing.

**STEP-4 — The content and the lifeline place the marker**
Files: `PersonTimelineContent.swift`, `PersonTimelineLifeline.swift`, `PersonTimelineContentTests.swift`, `PersonTimelineLifelineTests.swift`.
Do: D-6, D-7, D-8, D-9. Today's assertions in `placeholdersAndShortLives` for the funeral case (items `["birth", "c", "death", "f"]`, death above `f`) hold unchanged. First add the audit's test (T-4's first case) and see it fail on the old insertion.
Check: T-4, T-6 pass; `placeholdersAndShortLives` passes with its death assertions updated to D-8 (§8 T-6).

**STEP-5 — The content builds only what it shows**
Files: `PersonTimelineContent.swift`, `PersonTimelineContentTests.swift`.
Do: D-10, D-11. Update `pointsAreNamedByTheirRow` to D-11 (allowed under the Stop conditions).
Check: T-3, T-5 pass; every existing `PersonTimelineContentTests` case passes with unchanged items.

**STEP-6 — The lifeline makes a target per shown point and one for the folded**
Files: `PersonTimelineLifeline.swift`, `PersonTimelineSection.swift` (`PersonTimelineLine`), `PersonTimelineLifelineTests.swift`. Skill: `add-ui-component`.
Do: D-12. `PersonTimelineList` passes `select` unchanged.
Check: T-7 passes; manual check 3.

**STEP-7 — A date edit rescales the line**
Files: `PersonTimelinePlacement.swift`, `PersonTimelineSection.swift` (`rebuild()`), `macos/ProvenenciaTests/PersonTimelinePlacementTests.swift`.
Do: D-13.
Check: T-8 passes; `rg ChosenScale macos` finds nothing.

**STEP-8 — Docs**
Files: the S16-02 map row's Scope cell and the `#### S16-02` scope entry's **Ships** and **Check** in `docs/deployment-plan/spike-16/deployment-plan.md` (say after or before a row, the person's own burial or funeral, `lifelineTypes.markers`, a click target per shown row and one shared target for the folded points, and the years in the kept scale), `docs/ideas/person-page-performance.md` (one paragraph under "An unfolded Timeline lays out every card at once (P-7)": the dated rows after the death marker never fold; a crafted person with thousands of their own dated burials builds every one; user, 2026-11-09), `docs/deployment-plan/spike-16/deployment-plan.md` (the working rule "(Batch 1) A marker's position is one value": its first sentence becomes "A life marker that has no span of its own sits after one dated row or before one, never both: one `oneof` on the wire (after or before) and one Swift enum, in place of an empty `beside_event_id`.", its second "A placeholder death sits just before the earliest dated burial or funeral of the person's own (role subject) whenever one is dated, else after the last dated row; a placeholder birth mirrors it (just after the last dated before-birth row, else before the first dated row) (user, 2026-11-08: \"just before burial\"; 2026-11-09: a burial the person only witnessed is an ordinary row).", the rest stays; the working rule "(Batch 1) The Timeline builds what it shows": "a folded row keeps its id and span for the lifeline and the markers; the lifeline draws every point, makes a click target for each shown row and one shared target for the folded points, which opens the dated fold through the nearest folded point's page (FH-5; user, 2026-11-09)"; the map row status), `docs/deployment-plan/spike-16/completed.md`, this plan (status, Landed note).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-02-timeline-marker-and-shown-rows.md` reports no errors; links resolve; `rg -n 'beside' docs/deployment-plan/spike-16/deployment-plan.md api/proto/engine.proto core/database/persondetail` finds no marker wording.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Place a missing death just before a dated burial on the Timeline.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/persondetail/lifeline_test.go` `TestLifeMarkers` | The `typeOf` map gains 6 → `"baptism"` and 7 → `vocabkey.EventBurial`; the `row` helper gives the person the subject role, and a case can give another. Table: "a witness at a burial" (census 2, burial 4 as witness) → death after 4; "subject and witness on one burial" → death before 4; "only a burial dated" (4) → birth before 4, death before 4; "a baptism and a burial" (6, 4) → birth before 6, death before 4; "two burials" (4, 7) → death before 4; "a funeral after a census" → death before 3; "no burial or funeral dated" (census, marriage) → death after the marriage; "no dated row" → both nil; "a recorded birth with no span" → birth `Event` 1, before 2; "a recorded birth with a span" → birth nil place; "a recorded death with no span and a dated funeral" → death `Event` set, before the funeral; "a recorded death with a span" → nil place; with `lifelineTypes{beforeBirth: {"census"}}`: census, marriage dated → birth after the census | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, D-2, D-3 |
| T-2 | Go | `api/ffi/handlers/person_detail_test.go` `TestListPersonEvents` | Birth `event_id` = b1 and no place (`GetPlace() == nil`); death no `event_id`, `GetAfterEventId()` = u (the burial the person witnessed). The `lee` subtest: birth `event_id` = born, no place; death no `event_id`, `GetAfterEventId()` = born | D-4, D-14 |
| T-9 | Go | `core/database/persondetail/timeline_test.go` `TestPersonEventsLifeline` | Every subtest with D-15's expectation | AC-1.1, AC-1.5, AC-1.6, D-2, D-15 |
| T-3 | Swift | `PersonTimelineContentTests.swift` `buildsOnlyShownRowsAndCards` *(new)* | 2,000 dated events × 500 participants at Wide, folded: 50 rows in `items`, each with 12 cards; `allSpans.count == 2000`; `items.compactMap(\.spokenLabel).count == 52` (50 rows, the two placeholders); one page: 100 rows. Every event is a `census` (no after-death type), so all 2,000 are counted by the dated fold | AC-2.1, D-10 |
| T-4 | Swift | `PersonTimelineContentTests.swift` `deathNeverFollowsAnAfterDeathEvent` *(new)* | Audit P-1's case: `b1` burial 1850, `i1`, `i2` undated, birth `.before("b1")`, death `.before("b1")` → `["birth", "undated", "i1", "i2", "death", "b1"]`; baptism + burial, death `.before(burial)` → baptism, death, burial; a recorded death with no span `.before(funeral)` → its row as the marker before the funeral; birth and death both `.after("x")` → x, birth, death; death `.before("unknown")` → last | AC-1.1, AC-1.2, AC-1.4, D-6 |
| T-5 | Swift | `PersonTimelineContentTests.swift` `pointsAreNamedByTheirRow` (updated) | Every shown row's `Item.spokenLabel` is its row's spoken label; a lifeline's target label is its item's; `lise` folded: no target for `l60`; the `birth` placeholder's label is "Birth unknown" | AC-2.4, D-11 |
| T-6 | Swift | `PersonTimelineLifelineTests.swift` `placeholdersAndShortLives` (updated) and `aDeathBeforeABurial` *(new)* | `unknownEnds` (death `.before("11")`): birth cap at 18, point 2 at 58 (unchanged); death stub `[base(11.from) + 40, + 80]`, cap at its top, burial 11 at `base(11.from) + 80`. New: baptism 1820, burial 1850, both placeholders → death cap above the burial point by 40pt and below the baptism's point; a baptism and burial both in 1850 → baptism above the death cap, burial below the stub; birth and death both `.before` the one burial → birth stub is the 0th (`[base(at), base(at) + 40]`), death stub the 1st (`[base(at) + 40, base(at) + 80]`), burial at `base(at) + 80`; a death `.after` a folded row → stub at that row's `to`; a death `.before("unknown")` → the content's marker has `place == nil` and the stub sits at the range's end (review M-2) | AC-1.2, AC-1.5, D-8, D-9 |
| T-7 | Swift | `PersonTimelineLifelineTests.swift` `oneHundredTwentyEvents` (extended) | `lise` folded: `marks.count == 120`, `targets.count` = shown items with a point, `foldedHit` covers the folded marks; `foldedMark(nearest:)` at a folded mark's `pointY` → that mark; above the region → the first folded mark; unfolded: `foldedHit == nil`; a dot's `hit` is `y − 6 … y + 6`, a bar's `from − 6 … to + 6` | AC-2.2, AC-2.3, D-12 |
| T-8 | Swift | `PersonTimelinePlacementTests.swift` `aDateEditRescales` *(new)* | `KeptScale` from `johann` at 900: same content and height → its scale; event `"11"` (the burial, the last dated row) moved 50 years later under the same ids → nil; another viewport height → nil; another id list → nil | AC-3.1, AC-3.2, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonTimelineContentTests`, `PersonTimelineLifelineTests` (`placeholdersAndShortLives`), `CatalogQueryRegistryTests` (the timeline round trip), `TestListPersonEvents`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none added)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md`, `docs/deployment-plan/spike-16/` and `L10n.swift` with no hit left (`beside_event_id`, `besideEventId`, `lifeMarkers`, `pointLabels`, `ChosenScale`, `nextRowID`, `previousRowID`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (STEP-8)
- [ ] UI matches board S13-D4 and S16-D2; no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together): 0.4.4 → 0.4.5, or the next patch on main
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (none: not gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: none (the marker rule lives in the proto comment and the Go doc comment)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a project with: a person with a dated baptism and a dated burial and no death; a person known only from a burial with two undated events; a person with more than 100 dated events (the large fixture).

1. Open the baptism-and-burial person on Timeline: "Death unknown" sits between the baptism and the burial in the list; on the line, the dashed stub ends at the burial's point with the death cap at its top (AC-1.2).
2. Open the burial-only person: Birth unknown, Undated, the two events, Death unknown, the burial (AC-1.1).
3. Open the person with more than 100 events: 50 rows and "Show 50 more"; hover a shown point: its name shows; click below the 50th row's point among the folded points: the list opens through that page and scrolls to the row (AC-2.3, AC-2.4).
4. In a second window, move one of the first person's events 50 years later: the first window's lifeline takes a new scale and stays a usual height (AC-3.1).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `onTapGesture(coordinateSpace:)` does not fire on a clear view | `.contentShape(Rectangle())` before it, as the point buttons do; manual check 3 |
| Item-order point shifts move a point the placement anchors to | Placement reads `mark.pointY`, which D-9 sets; T-6 and the placement tests |
| The fixture rewrite changes a test's expected order | STEP-3 maps every fixture to what Go sends now; only `unknownEnds`' death moves, and its list order is unchanged |

**Stop and report if:**

- `PersonEvents` does not return rows in date order (D-2's "first after-death row" is then not the earliest).
- A test outside §8 must change its expected output for a reason other than the `besideEventId` → `place` rewrite, D-11's labels, or D-14.
- `PersonTimelinePlacement` needs a folded row's spoken label or a `Row` for a folded event.
- The shared folded region swallows clicks meant for a shown point's button (the buttons must stay on top).
- The S13-D4 board draws the death stub after the burial for any frame.

## 12. Open questions

None.

## 13. PR

**Title:** S16-02 — The Timeline places a missing death and builds what it shows
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-02-timeline-marker-and-shown-rows.md`.
