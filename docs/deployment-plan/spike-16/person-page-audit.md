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
- **P-15 and P-17 – P-27 are parked** in idea files (table below).

## Scheduled

| Item | Title | Priority | Lens | PR |
| --- | --- | --- | --- | --- |
| P-1 | A missing death sits after a burial | High | CON | S16-02 |
| P-2 | Folding an event's places is quadratic, on project open too | High | SEC | S16-08 |
| P-3 | The Family read is quadratic in a crafted wide family | High | SEC | S16-01 |
| P-4 | The claims read rescans its graph for each record | High | SEC | S16-01 |
| P-5 | A failed refresh is silent and the old rows stay | Medium | UX | S16-04 |
| P-6 | Vitals and Gallery keep old Source, Citation and Artifact words | Medium | CON | S16-09 |
| P-7 | The Timeline builds every event, not the 50 it shows | Medium | PRF, SEC | S16-07, S16-02 |
| P-10 | The diagram silently draws uncounted links as weak | Medium | SEC, UX | S16-01, S16-05 |
| P-11 | Switching sections is silent and drops focus | Medium | UX | S16-06 |
| P-12 | A Remove that keeps the person is not announced | Medium | UX | S16-06 |
| P-13 | Section plumbing: seven switches, three ways to find the open section | Medium | MNT | S16-03 |
| P-14 | Fold focus is copied per section, and Family drops it | Medium | MNT, UX | S16-07 |
| P-16 | The leave-on-delete decision is not tested | Low | CON | S16-06 |

## Descoped and parked

| Item | Title | Home |
| --- | --- | --- |
| P-8 | Timeline participants' summaries are sent once per event | [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md) |
| P-9 | The record links read recomputes every pair, with no memo | [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md) |
| P-15 | Hovering a node whose record line is folded highlights nothing | [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md) |
| P-17 | "handle", "members" and "This record no longer exists" on the Person page | [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md) |
| P-18 | Birth and death event counts have no plural; the catalog check skips them | [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md) |
| P-19 | Back resets a section's scroll, folds and focus | [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md) |
| P-20 | The Timeline keeps an old scale after a date edit | [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md) |
| P-21 | `ListPersonClaims` answers for a non-Person handle | [`ideas/structural-debt.md`](../../ideas/structural-debt.md) |
| P-22 | The record links wire repeats the claims read's graph rows | [`ideas/structural-debt.md`](../../ideas/structural-debt.md) |
| P-23 | The family response sends each sibling two or three times | [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md) |
| P-24 | `edgeKey` joins ids with no separator | [`ideas/share-packages.md`](../../ideas/share-packages.md) |
| P-25 | A fourth hover-and-focus-ring button body (A-41 grew) | [`ideas/structural-debt.md`](../../ideas/structural-debt.md) (A-41) |
| P-26 | File and layer tidy-ups in `Features/Conclusions` | [`ideas/structural-debt.md`](../../ideas/structural-debt.md) |
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

**Fix.** Go says where the marker goes when no row qualifies for `Beside`: the first dated after-death row the death sits before (the mirror for birth). Swift inserts the marker before that row, and the lifeline puts the stub at that row's `from`. Add the Go case and the Swift test.

**Origin.** S13-33 (#687) returned one empty value for both reasons; S13-13 (#691) D-5 read it as "last". The S13-13 plan's neighbours asked the question and the Decision answered it for one reason only.

### P-2 · Folding an event's places is quadratic, on project open too

**Where:** `core/database/conclusionsummaries/placechains.go` `FoldLocations` (233–258), `isAncestor` (204–225); `compose.go` `eventPlaces` (1105–1135), which removes duplicate ids and nothing else. Callers: every event summary, every person's birth and death row, the Timeline (`EventsByIDs`), the Events list and Event pages. Open path: `core/catalogsession/session.go` `ReconcileOnOpen` (54–68) → `searchindex.EnsureCatalog` (`core/database/searchindex/searchindex.go:638`) → `NeedsRebuild` (277) → `reprojectAllHandles` (`handles.go:50`) → `RowsForSearch`.

**Problem.** `FoldLocations` checks every ordered pair of an event's places, each with a fresh upward search that allocates. Nothing bounds how many places an event names. An event with 5,000–10,000 location links means 25–100 million searches: minutes. A crafted catalog that also fails `NeedsRebuild` (an old projection version) runs it on every open. `summaryChainDepth` bounds how far a chain climbs, not how many places an event names.

**Evidence.** Seed test for `placechains_internal_test.go`: 5,000 unlinked places through `FoldLocations` must finish under a second.

**Fix (capped, per the decision).** One helper reads an event's places, drops duplicates and keeps at most a fixed number in link order, and the four reads of an event's places in `compose.go` (`eventPlaceGraph` 797, `lifeOwners` 818, `composeEvents` 876, `eventPlaces` 1108) use it (rule "Crafted fan-out is capped, not engineered"). A real event names one to three.

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

**Fix.** `Pairs` reports whether it counted; `persondetail.RecordLinks` caps the lines it sends at the number the diagram draws, strong lines first, says when it did not count or dropped lines, and Swift's own arc cap goes (S16-01). The diagram column shows one note when the read did not count, Go dropped lines or the circle dropped graphs past its 1,000-node cap, instead of drawing uncounted arcs as weak (S16-05, brief S16-D1).

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

**Evidence.** S13-17 fixed "deleting a person is silent" and kept RC-6's no-Toast answer ([ship log](../spike-13/prs/S13-17-remove-a-claim.ship.md)); the keep-person path was not raised. Manual: VoiceOver on, Unpromote one of two records.

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
