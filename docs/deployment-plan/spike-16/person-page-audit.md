# Person page audit — 2026-11-08

The findings behind Batch 1 of the [map](deployment-plan.md). An area audit of the Spike 13 Person page (header, section switcher, Vitals, Family, Timeline, Gallery, Sources and the record links diagram) at `main` @ 78f2681, with [`review-pr`](../../../.agents/skills/review-pr/SKILL.md) ("Area audit"), all five lenses.

**How to use this doc.** A scope entry names its items (`P-n`). `plan-pr` reads those items here for where, evidence and fix, and re-checks each against the code first: line numbers are from 78f2681 and move. Nothing was run on the audit machine (no Go tests, no Xcode), so every finding comes from reading the code, every timing is an estimate, and each item names the test that should fail today. The plan's first step is to add that test and see it fail.

**What held up.** Sections are places (Back, Forward and restore return to the section; an old section key decodes to Vitals); family links open Family, timeline participants open Timeline, other Person links open Vitals; a gone handle leaves through `.gone`; nothing unpromoted crosses the wire (R-2); family grouping handles two marriages, conflicting parent sets, a co-parent who is not a spouse, self-links and cycles; FFI runs off the main thread; likeness images decode through `BoundedImageDecode`; Swift arithmetic on Go's spans and counts is bounded (`datevalue.MinYear` … `MaxYear`); no Go panic path was found (`ffi.Call` recovers); no SwiftUI layout cycle in scope.

## Decisions on this audit

User, 2026-11-08. They are the map's Batch 1 working rules.

- **P-1 and P-5 – P-7 and P-10 – P-14 proceed.**
- **P-2 – P-4 are capped, not engineered.** They need a crafted or badly broken project; a simple cap checked before the work is enough.
- **P-7 is a render cap.** The Timeline builds what its folds show and every fold unfolds a page at a time; no memo of the section's content.
- **The map review re-cut four items** (2026-11-08, [review](deployment-plan.review.md)): P-2 has its own PR after Spike 14's S14-40; P-6 has its own PR, ungated; P-14 and P-7's pages share one fold PR; P-16 is scheduled with P-12.
- **P-8 and P-9 are descoped** to [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md).
- **Eight parked items ride with the PRs that touch their files** (user, later on 2026-11-08): P-15, P-17 (with the wording the audit proposed), P-20, P-21, P-24, part of P-22 and part of P-26, and the unused `CatalogQueryStalePolicy` from [`ideas/structural-debt.md`](../../ideas/structural-debt.md) (S16-09).
- **The PRs were combined once more to cut review cycles** (user, later on 2026-11-08): P-2 joins the other caps in S16-01, P-10's note joins the notices in S16-04, and part of P-22 moves to S16-12, which changes the same message.
- **P-18, P-19, P-23, P-25, P-27 and the rest of P-22 and P-26 stay parked** in idea files (table below).

## Scheduled

| Item | Title | Priority | Lens | PR |
| --- | --- | --- | --- | --- |
| P-1 | A missing death sits after a burial | High | CON | S16-02 |
| P-2 | Folding an event's places is quadratic, on project open too | High | SEC | S16-01 |
| P-3 | The Family read is quadratic in a crafted wide family | High | SEC | S16-01 |
| P-4 | The claims read rescans its graph for each record | High | SEC | S16-01 |
| P-5 | A failed refresh is silent and the old rows stay | Medium | UX | S16-04 |
| P-6 | Vitals and Gallery keep old Source, Citation and Artifact words | Medium | CON | S16-09 |
| P-7 | The Timeline builds every event, not the 50 it shows | Medium | PRF, SEC | S16-07, S16-02 |
| P-10 | The diagram silently draws uncounted links as weak | Medium | SEC, UX | S16-01, S16-04 |
| P-11 | Switching sections is silent and drops focus | Medium | UX | S16-06 |
| P-12 | A Remove that keeps the person is not announced | Medium | UX | S16-06 |
| P-13 | Section plumbing: seven switches, three ways to find the open section | Medium | MNT | S16-03 |
| P-14 | Fold focus is copied per section, and Family drops it | Medium | MNT, UX | S16-07 |
| P-16 | The leave-on-delete decision is not tested | Low | CON | S16-06 |
| P-15 | Hovering a node whose record line is folded highlights nothing | Low | CON | S16-07 |
| P-17 | "handle", "members" and "This record no longer exists" on the Person page | Low | UX | S16-04, S16-06 |
| P-20 | The Timeline keeps an old scale after a date edit | Low | PRF | S16-02 |
| P-21 | `ListPersonClaims` answers for a non-Person handle | Low | MNT, CON | S16-01 |
| P-22 (part) | The record links wire's unread fields and missing `source_ref` | Low | MNT | S16-12 |
| P-24 | `edgeKey` joins ids with no separator | Low | SEC | S16-16 |
| P-26 (part) | `ConclusionDetailPage.swift` holds 15 types | Low | MNT | S16-03 |

## Descoped and parked

| Item | Title | Home |
| --- | --- | --- |
| P-8 | Timeline participants' summaries are sent once per event | [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md) |
| P-9 | The record links read recomputes every pair, with no memo | [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md) |
| P-18 | Birth and death event counts have no plural; the catalog check skips them | [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md) |
| P-19 | Back resets a section's scroll, folds and focus | [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md) |
| P-22 (the rest) | The claims read could own every graph label the record links diagram shows | [`ideas/structural-debt.md`](../../ideas/structural-debt.md) |
| P-23 | The family response sends each sibling two or three times | [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md) |
| P-25 | A fourth hover-and-focus-ring button body (A-41 grew) | [`ideas/structural-debt.md`](../../ideas/structural-debt.md) (A-41) |
| P-26 (the rest) | Layer tidy-ups in `Features/Conclusions` | [`ideas/structural-debt.md`](../../ideas/structural-debt.md) |
| P-27 | Per-key copies of registry tests and FakeStore reads | [`ideas/structural-debt.md`](../../ideas/structural-debt.md) |

## Items

### P-1 · A missing death sits after a burial

**Where:** `core/database/persondetail/lifeline.go` `lifeMarkers` (49–66); `lifeline_test.go` case "only a burial dated" (31); `macos/App/Features/Conclusions/PersonTimelineContent.swift` (164–171); `PersonTimelineLifeline.swift` insertions (118–126); `api/proto/engine.proto` `LifeMarker` (1569–1583).

**Problem.** `beside_event_id` is empty for two different reasons: nothing is dated, or every dated row is a burial or funeral (`lifeline.afterDeath`). Swift reads both as "nothing dated" and puts the death marker last. A person known only from a burial reads Birth unknown → Burial 1850 → the undated events → Death unknown, and the lifeline draws the death stub at the burial's end. A recorded death with no date does the same. The birth side has the same fault the day `lifeline.beforeBirth` gains a type. This breaks R-24a and the Spike 13 rule "a recorded funeral with no death puts the death placeholder before the funeral".

**Evidence.** The Go case pins `deathBeside: nil` for "only a burial dated"; Swift then inserts at `dated.count`. Failing test for `macos/ProvenenciaTests/PersonTimelineContentTests.swift`:

```swift
@Test func deathNeverFollowsAnAfterDeathEvent() {
    var timeline = Data.ida                      // i1, i2 undated
    timeline.events.insert(Data.event("b1", type: "burial", when: .year(1850)), at: 0)
    timeline.birth = CatalogLifeMarker(besideEventId: "b1")
    timeline.death = CatalogLifeMarker()         // Go today: no row qualifies → empty
    #expect(content(timeline).items.map(\.id) == ["birth", "undated", "i1", "i2", "death", "b1"])
}
```

Today: `["birth", "b1", "undated", "i1", "i2", "death"]`.

**Fix.** Go says where the marker goes: a placeholder death sits just before the earliest dated after-death row whenever one is dated, else beside the last dated row (the mirror for birth). Swift inserts the marker before that row, and the lifeline puts the stub at that row's `from`. Add the Go cases and the Swift test. The user chose "just before the burial" on 2026-11-08 for the neighbour too: with a dated baptism and burial, R-24a as written put the death right after the baptism and drew the life between them after the death.

**Origin.** S13-33 (#687) returned one empty value for both reasons; S13-13 (#691) D-5 read it as "last". The S13-13 plan's neighbours asked the question and the Decision answered it for one reason only.

### P-2 · Folding an event's places is quadratic, on project open too

**Where:** `core/database/conclusionsummaries/placechains.go` `FoldLocations` (233–258), `isAncestor` (204–225); `compose.go` `eventPlaces` (1105–1135), which removes duplicate ids and nothing else. Since S14-40 landed (#746), an event's places are read in `placesOfEvents` (used by `eventPlaceGraph`), `leadLife`, `lifeOwners` and `composeEvents`, and `eventPlaces` takes its caller's links; `summaryChainDepth` moved to `placechains.go`. Callers: every event summary, every person's birth and death row, the Timeline (`EventsByIDs`), the Events list and Event pages. Open path: `core/catalogsession/session.go` `ReconcileOnOpen` (54–68) → `searchindex.EnsureCatalog` (`core/database/searchindex/searchindex.go:638`) → `NeedsRebuild` (277) → `reprojectAllHandles` (`handles.go:50`) → `RowsForSearch`.

**Problem.** `FoldLocations` checks every ordered pair of an event's places, each with a fresh upward search that allocates. Nothing bounds how many places an event names. An event with 5,000–10,000 location links means 25–100 million searches: minutes. A crafted catalog that also fails `NeedsRebuild` (an old projection version) runs it on every open. `summaryChainDepth` bounds how far a chain climbs, not how many places an event names.

**Evidence.** Seed test for `placechains_internal_test.go`: 5,000 unlinked places through `FoldLocations` must finish under a second.

**Fix (capped, per the decision).** `placesOfEvents` drops duplicates and keeps at most a fixed number in link order, and the other reads of an event's places (`leadLife`, `lifeOwners`, `composeEvents`) move onto it (rule "Crafted fan-out is capped, not engineered"). A real event names one to three.

**Origin.** Older than Spike 13 (`e951c0f7`, 2026-10-07); Spike 13 reaches it from the Timeline (S13-05) and the header.

### P-3 · The Family read is quadratic in a crafted wide family

**Where:** `core/database/persondetail/family.go` `Families` (49–80): every relative the kin names goes to `LoadRelationships`; `core/database/canonlinks/reader.go` `For` per relative (446–452), whose `around` and `forLinks` (639–656) walk the person's whole link list again for each relative; `core/kinship/siblings.go` makes one implied link per parent × child and has no bound on parents (its comment at 26); `core/kinship/families.go` scans a sibling's shapes once per parent (231–240) and `partnerUnits` checks children with `slices.ContainsFunc` (315).

**Problem.** With C children, each child's `For` reaches the person and rebuilds a slice and map the size of the person's links: C × links. A sibling who shares P parents carries P shapes, merged and placed in P² per sibling. A crafted person with 20,000 children, or one sibling with 50,000 shared parents, takes tens of seconds. Family is restored at launch (S13-16 SEC-2), and the catalog session is held, so every pane waits. `TestPersonFamilyQueryCount` counts queries, not work; queries stay fixed.

**Evidence.** Seed test: `persondetail.PersonFamily` on 1,000 and 4,000 children; the larger must not take more than 8× the smaller. A `kinship` test with 30,000 shared parents through `Siblings` → `KinOf` → `FamilyUnits` under a second.

**Fix (capped).** Keep at most a fixed number of relatives per role from the person's kin, in link order, in one capped kin that both of `Families`' loops use (the relatives it loads, then `Around` and `FamilyUnits`), so no relative is read that was not loaded; derive no implied siblings through a person with more than `kinship.Rules.MaxPivotEnd` parents (the bound `parentUnits` already uses to stop pairing).

**Origin.** `forLinks`' walk predates Spike 13 (S11-49); S13-04 (#671) runs it once per relative and added `partnerUnits`; S13-03 (#666) added one shape per parent.

### P-4 · The claims read rescans its graph for each record

**Where:** `core/database/persondetail/claims.go` `claims` (75–81: every Observation of the graph per member) and `memberGraphs` (117–150: the graph's Members and Subjects per member); `api/ffi/handlers/person_detail.go` (42–50) with `source_read.go` (79–83), which rebuild a label map of every Subject once per claim. `persondetail/recordlinks.go` (32) reaches the same `claims`.

**Problem.** M members of the person in one graph of S Subjects and O Observations cost M × (2S + O) compares and M × S map inserts. A crafted catalog promoting 20,000 Subjects of one graph onto one Person runs for minutes, in Sources and in the record links read. A real person has one or two records per graph.

**Evidence.** Seed test with `claims_test.go`'s fixture: one graph of 5,000 persons, all promoted onto one handle; `claims` and `ListPersonClaims` under a second.

**Fix (capped).** Read at most `RecordLinks.MaxEntityMembers` (10; user, 2026-11-07: "more than ten members in a graph is malformed") of the person's members per graph, in member order.

**Origin.** S13-01 (#655).

### P-5 · A failed refresh is silent and the old rows stay

**Where:** `macos/App/Features/Workspace/Session/QueryHandle.swift` `read` (53–57: a value wins, even a stale one); `WorkspaceSession.swift` (458: a failed reload keeps the value); `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` (162–229) and `ConclusionDetailPage.swift` (13–30, 478–480), which switch on `read` and never read `handle.failure`. The error Callout does not announce and says "Please try again" with nothing to press.

**Problem.** When a reload fails behind a value, the section keeps the old rows and says nothing. After a Remove whose reload fails, the removed line stays; a second Remove is a silent no-op (`core/database/promote/unpromote.go:18–35`). Affects Family, Timeline, Sources, the record links column and the header, on the Event and Place pages too.

**Evidence.** `SourcesModel.swift` (147, 203) and `EvidenceGraphModel.swift` (173) read `failure` beside a value; `PVRetryCallout` announces and has a retry (`EvidenceGraphView.swift:296`, `PromoteView.swift:89`). Test with FakeStore: load `.personClaims`, make the next `listPersonClaims` throw, apply effects naming the handle; the host still draws the old rows and no `persons.detail.sources.error`.

**Fix.** The page's presentation carries a failure beside its value; the page and the section host show `PVRetryCallout` above the stale content (compact in a column), and on a first-load failure too; Try again re-queries the section's reads (S16-03).

**Origin.** The `read` rule is S9-24's; S13-13 (#691) copied it into the section host and every later host inherited it.

### P-6 · Vitals and Gallery keep old Source, Citation and Artifact words

**Where:** `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `.conclusionDetail` (556–571): it follows its own id in `touchedHandles` and property labels, nothing else. Records carry `SourceTitle`, the Artifact and Citation labels and the page (`core/database/conclusiondetails/conclusiondetails.go` 95, 445).

**Problem.** A Source rename reports the Source and its graphs but no handle (`api/ffi/handlers/source_write_effects_test.go:135–160`); a Citation label or page edit and an Artifact label edit name no handle either. The cached detail is never refreshed, so the Vitals provenance sheet, Gallery captions and the viewer's record row keep the old words, while the Sources section on the same page shows the new title.

**Evidence.** `WorkspaceSessionTests.personClaimsFollowHandlesAndGraphs` asserts the detail page follows only the handle. Failing test for `WorkspaceSessionTests.swift`: a detail whose picture record cites `src1`; `session.revalidatedKeys(for: CatalogWriteEffects(sourceScopes: ["src1"]))` should contain the detail key and is empty today.

**Fix.** Add `.applyHiddenPolicy(followsValue: .source, ids:)` to the `.conclusionDetail` descriptor, returning the `sourceID`s of the detail's field records, derived-row owners and pictures. All three edits move their Source's scope.

**Origin.** `sourceTitle` predates Spike 13; S13-11 (#650) added the Citation parts with no new follow, and S13-14 / S13-15 put the citation line on every tile and in the viewer.

### P-7 · The Timeline builds every event, not the 50 it shows

**Where:** `macos/App/Features/Conclusions/PersonTimelineContent.swift` `init` (126–129: a `Row` with its cards for every event) and the folds (190–212, applied after); `ConclusionFold.shown(…, expanded:)` reveals every hidden item at once; `PersonTimelineLifeline.swift` (158–176: a mark for every dated row, folded ones included) and `PersonTimelineSection.swift` (392–401: a `Button` with a tooltip per mark).

**Problem.** Participant cards are folded per row (5 at Regular, 12 at Wide), but rows are built for every event: at 200 events and Wide that is 150 rows and up to 1,800 cards formatted on the main thread and never drawn, on every pass of the section host. One "Show N more" builds every hidden row or card at once, and the lifeline holds a button per dated event. A crafted person with 20,000 events, or an event with 10,000 participants, stalls the main thread, and the Timeline is restored at launch.

**Evidence.** `PersonTimelineLifelineTests.oneHundredTwentyEvents` asserts a folded line has 120 marks. Budget test: build `PersonTimelineContent` from 2,000 events × 50 participants at Wide; the build touches only the shown rows and cards.

**Fix (render cap, per the decision).** `ConclusionFold` reveals a page at a time for every section (S16-07). The Timeline builds rows and cards only for what the folds show; folded rows keep their id, span and spoken label for the lifeline and the markers. The lifeline makes a click target only for a shown row's mark (S16-02). Not here: memoising the content per revision, and the Sources diagram's rebuild on focus ([`ideas/person-page-performance.md`](../../ideas/person-page-performance.md)).

**Origin.** S13-13 (#691).

### P-10 · The diagram silently draws uncounted links as weak

**Where:** `core/recordlinks/recordlinks.go` `Pairs` (175–192): past `MaxReadWork` every pair is listed with 0 facts; past the line limit no pair is returned. `core/matchregistry/recordlinks.go` `MaxReadWork` (1,100,000). `api/ffi/handlers/person_detail.go` (207–234). `macos/App/Features/Conclusions/PersonRecordLinksContent.swift` (16–50): the 75,000-arc and 1,000-node caps and the strong-first order are Swift's.

**Problem.** A read that stops counting comes back looking like real data: every arc thin, or no arcs ("these records stand apart"). A researcher can conclude the records are weakly connected when nothing was compared. Go also builds, marshals and sends up to about 1.1M lines that the app then cuts to 75,000.

**Evidence.** Census-sized households of 12 people with implied sibling links are estimated at 52–78% of `MaxReadWork` (performance lens), so a large real person comes near the edge. Handler test: a read past `MaxReadWork` returns lines with 0 facts and nothing that says so.

**Fix.** `Pairs` reports whether it counted; `persondetail.RecordLinks` caps the lines it sends at the number the diagram draws, strong lines first, says when it did not count or dropped lines, and Swift's own arc cap goes (S16-01). The diagram column shows one note when the read did not count, Go dropped lines or the circle dropped graphs past its 1,000-node cap, instead of drawing uncounted arcs as weak (S16-04, brief S16-D1).

**Origin.** S13-27 (#743) and S13-28 (#744).

### P-11 · Switching sections is silent and drops focus

**Where:** `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (529: `select` only navigates; 595–611: the switcher has no focus handle); `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift` (no Tab stop without a handle).

**Problem.** ⌘1–⌘5 replaces the section view and destroys whatever had focus. VoiceOver hears nothing; keyboard focus falls back to the window's first key view.

**Evidence.** No `pvAnnounce` or `AccessibilityAnnouncement` in the page or the switcher. Manual: VoiceOver and Full Keyboard Access on; on Family, Tab to a tile, press ⌘3; nothing is spoken and the next Tab leaves the section.

**Fix.** Announce the chosen section's label on change (not on first appear); give the switcher a focus handle and claim the selected segment when the old focus is gone. `PVSegmentedControl` puts its handle on the first enabled segment today, so the component changes to the selected one.

**Origin.** S13-10 (#634).

### P-12 · A Remove that keeps the person is not announced

**Where:** `macos/App/Features/Conclusions/PersonSourcesSection.swift` `confirm` (75–88): the announcement is on the delete branch only.

**Problem.** On the common path the confirm closes and focus moves to the next line's Unpromote; VoiceOver reads that button, which sounds as if nothing happened.

**Evidence.** S13-17 fixed "deleting a person is silent" and kept RC-6's no-Toast answer (S13-17 ship log, in git with #729); the keep-person path was not raised. Manual: VoiceOver on, Unpromote one of two records.

**Fix.** A spoken-only string ("{name} unpromoted.") posted on the branch where the person stays. No Toast. The leave-or-stay decision moves out of the view into a function a unit test reaches, which also closes P-16.

**Origin.** S13-17 (#729).

### P-13 · Section plumbing: seven switches, three ways to find the open section

**Where:** `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection` (29–99: `offered`, `queryKey`, `supportingQueryKeys`, `companionQueryKey`, `fillsWidth`, `label`, `glyph`); `Session/PlaceRegistry.swift` (192–201: reads the raw `pageSection`, nil for the first section); `ConclusionDetailPage.swift` (483–499: `?? offered[0]`; the refreshing state lists two of the three reads); `PersonDetailView.swift` (34: `?? .vitals`); `ConclusionPageSectionHost.swift` (158–160: id root `persons.detail.` hard-coded); `PersonSourcesSection.swift` (202: passes the companion key again).

**Problem.** Each consumer re-lists a section's reads and resolves the open section its own way. A first section that reads its own key is missing from the place's `queryKeys`, so `visibleKeys` (`WorkspaceSession.swift:222`) misses it and a write evicts it while on screen (`.evict`, line 204). Harmless while Vitals is first and reads nothing; it breaks when Event or Place adopt sections. The skill the "Sections are places" rule cites, [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md), never mentions page sections; [`macos-client-patterns.md`](../../macos-client-patterns.md) line 103 still says only Family reads its own key.

**Fix.** `WorkspaceLocation.openPageSection` (`pageSection ?? offered.first`), used by the place registry, the page and the Person fill; one `reads(project:entityId:)` list per section, each read tagged own, companion or supporting, iterated by the place, the refreshing state and the hosts; the host's id root from the page identifier; the skill and the doc updated.

**Origin.** S13-12 (#676), S13-16 (#722) and S13-28 (#744) each added a slot.

### P-14 · Fold focus is copied per section, and Family drops it

**Where:** `PersonTimelineSection.swift` (57–69, 116–127, 154–172), `ConclusionGallerySection.swift` (20–29, 51–69, 108–136), `PersonSourcesSection.swift` (247 and 287: a `Task` hop; 270 and 280: `FocusClaim`), `PersonFamilySection.swift` (34, 93), `macos/App/Platform/FocusClaim.swift`, `ConclusionShowMoreRow.swift`.

**Problem.** "Show N more moves focus to the first revealed item" and "focus survives a width change" are solved privately per section, with two mechanisms in one file. Family has none: when its Show more row disappears, keyboard focus is lost. Sources keeps its folds as a `Set<String>` plus a `Bool`, beside the other sections' fold enums.

**Fix.** `ConclusionFold`'s `firstRevealed` names the first item of the page just revealed and `ConclusionShowMoreRow` claims it through `FocusClaim`; Family, Timeline and Sources use it with one fold state type. One focus keeper in `Platform/`, beside `FocusClaim`, holds focus across a width change for Gallery and Timeline.

**Origin.** S13-13 (#691) wrote the Timeline copy; S13-16 (#722) copied it (its MNT-5 was fixed by matching); S13-17 (#729) added `FocusClaim` beside it.

### P-16 · The leave-on-delete decision is not tested

**Where:** `macos/App/Features/Conclusions/PersonSourcesSection.swift` `confirm` (75–88); `macos/ProvenenciaTests/UnpromoteFlowTests.swift` `confirmReportsTheEffects` (228–232).

**Problem.** R-39's "land on the Persons list" decision (leave when the write deleted this person, else stay) is inline in a View, and the test rewrites it in its own `apply:` closure. Breaking `PersonSourcesSection.confirm`, say by always calling `session.apply(effects)`, fails no test; S13-17's AC-2.4 and AC-2.5 rest on the manual check. P-12 adds an announcement to the same branch.

**Fix.** Move the decision into a function on `UnpromoteFlow` or a static helper that returns leave, or stay with the announcement text, and test both branches against a `WorkspaceNavigation`. Scheduled with P-12 (map review K-8).

**Origin.** S13-17 (#729).

### P-15 · Hovering a node whose record line is folded highlights nothing

**Where:** `macos/App/Features/Conclusions/PersonSourcesSection.swift` (356–365: only drawn lines take the highlight); `PersonSourcesContent.swift` (13–17: lines fold past five).

**Problem.** The diagram and the Sources rows highlight each other (Spike 13 rule "The view is a circular link diagram"), but a Source with more than five of the person's graphs folds its sixth line on, so hovering or focusing that graph's node shows nothing in the list. A parish register filed one graph per entry reaches it.

**Fix.** When the highlighted graph's line is folded, that row's Show more takes the highlight; the content already maps folded claims to graphs (`claimGraphs`). With S16-07, which rewrites the Sources folds.

**Origin.** S13-28 (#744) on S13-16 (#722).

### P-17 · "handle", "members" and "This record no longer exists" on the Person page

**Where:** `unpromote.confirmMessage` (`macos/App/Platform/L10n.swift` about 8387); `conclusions.memberCount` (about 10181), read by `PersonDetailContent`, `EventDetailContent` and `PlaceDetailContent`; `error.conclusiondetails.not_found` (about 10638), shown by `ConclusionDetailPage.swift` (552) and `ConclusionPageSectionHost.swift` (130).

**Problem.** The confirm ends "…, and the handle stays", the only user-facing "handle"; the header counts "members" while Sources and the confirms say records; a deleted person's page says "This record no longer exists", where "record" names the person on a page where it names a claim.

**Fix (user, 2026-11-08).** The count reads records on every Conclusion page; the confirm ends "…, and the person stays", per kind like `confirmMessageDeletes` (S16-06); a Person page's not-found line is "This person no longer exists", supplied by the kind (S16-04). Rule "A person's records are records".

**Origin.** The count is S9's; "records" came with S13-08 and S13-16; the Person page's not-found Toast with S13-17.

### P-20 · The Timeline keeps an old scale after a date edit

**Where:** `macos/App/Features/Conclusions/PersonTimelineSection.swift` `PersonTimelineList.rebuild` (139).

**Problem.** The chosen scale is kept while the item ids and the viewport height stay the same. A date edit from another window keeps the ids, so the old points per year apply to the new range: a short life drawn dense, then an event moved decades away, gives a line many times its usual height.

**Fix.** Key the kept scale on the line's year range too. With S16-02.

**Origin.** S13-13 (#691).

### P-21 · `ListPersonClaims` answers for a non-Person handle

**Where:** `core/database/persondetail/claims.go` `Claims` (41–47) checks only `conclusiondetails.Handle`; `personNode` (52–65), "the one place that decides whether an id is a Person page's", gates `RecordLinks` and `PersonEvents`; `family.go` re-derives the check.

**Fix.** Every `persondetail` read goes through `personNode`; the proto comments say one sentence; each table test gains a non-Person handle. With S16-01.

**Origin.** S13-01 (#655), S13-04 (#671), S13-27 (#743).

### P-22 (part) · The record links wire's unread fields and missing `source_ref`

**Where:** `api/proto/engine.proto` `RecordLinksGraph` (1643–1650); `core/database/persondetail/recordlinks.go` (185–198); `macos/App/Features/Conclusions/PersonRecordLinksContent.swift` (20–26, 55–75); `PersonSourcesContent.swift` (58–64, 85–87, 130); `PersonSourcesSection.swift` (207–212).

**Problem.** The wire repeats the claims read's graph rows but not `source_ref`, so the Sources section hands `graphSourceRefs` to the diagram's adapter, and the "name, else ref" and "title, else ref" fallbacks are written twice. `source_id` and `source_type_id` are decoded and never read.

**Fix.** Drop the unread `source_type_id` and the handoff, and write each fallback once. S16-12 found `source_ref` unneeded: a Source's title is required, so the diagram labels with the title, and `source_id` stays for the app's `.source` follow (user, 2026-11-12). With S16-12, which changes the same message (moved from S16-01 when the PRs were combined). The rest (the claims read owning every graph label) stays in [`ideas/structural-debt.md`](../../ideas/structural-debt.md).

**Origin.** S13-27 (#743), S13-28 (#744).

### P-24 · `edgeKey` joins ids with no separator

**Where:** `core/database/canonlinks/reader.go` `edgeKey` (623), used at 606–612 to merge the sibling edges `kinship.Siblings` reads (616).

**Problem.** The key concatenates the two ends and the associations with no separator, and ids are not length-checked when loaded, so a crafted catalog can give two links one key and hide a sibling or a link's source. Logged in the S13-03 ship log (SEC-2) and tracked nowhere until this audit.

**Fix.** Prefix each part with its length (or skip ids that are not 16 bytes). With S16-16, which rewrites how `canonlinks` composes kinship answers (moved from S16-01 when Batch 2 was cut).

**Origin.** S13-03 (#666).

### P-26 (part) · `ConclusionDetailPage.swift` holds 15 types

**Where:** `macos/App/Features/Conclusions/ConclusionDetailPage.swift`: the Place relationship model (33–56), the gallery model with `LeadCopy` and `ViewerCopy` (90–240), the Vitals rows and their provenance sheet presenter (414–462), the Place relationship section (625–662).

**Problem.** The file grew from 484 to 736 lines in Spike 13; S16-03, S16-04 and S16-06 all edit it.

**Fix.** Move the gallery model to `ConclusionGallerySection.swift`, `ConclusionDetailRowsSection` beside `ReconciledValueRow`, and the Place relationship model and section beside `PlaceRelationshipRow`, in S16-03 before its other steps. The other tidy-ups stay in [`ideas/structural-debt.md`](../../ideas/structural-debt.md).

**Origin.** S13-09, S13-10, S13-11, S13-14 and S13-15.

## Close the loop

Classes the audit traced to a review gate, for the lenses' **Known traps** (not yet added):

- **A nil with two meanings** (P-1, conformance): list each reason a returned empty can be empty and check the consumer's fallback for each.
- **A value that grew, a key that did not follow** (P-6, conformance): for each field a PR adds to a cached value, name the write that changes it and check its effects reach the key.
- **`handle.read` hides a refresh failure** (P-5, UX): a view that switches on `read` never sees a failure beside a value.
- **A shortcut that swaps content** (P-11, UX): name where focus lands and what VoiceOver hears.
- **No Toast is a visual choice** (P-12, UX): every write outcome is still spoken.
- **Few queries is not little work** (P-3, security): a per-item step that walks a shared node's whole list costs items × that node's size.
- **A fold bounds rows, not the work beside them** (P-7, security and performance): count every view and value built per item, not only the rows drawn.
- **A slot per read role** (P-13, maintainability): one read list and one resolved open section.
