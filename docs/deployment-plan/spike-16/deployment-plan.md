# Deployment plan — Spike 16

Spike 16 is an odds-and-ends spike ([odd and even spikes](../README.md#odd-and-even-spikes)): small fixes, UX refinements and follow-ups, in batches, each a slice with its own Check. Batch 1 takes the high and medium findings of the [Person page audit](person-page-audit.md) (2026-11-08), an area audit of Spike 13's Person page with all five [`review-pr`](../../../.agents/skills/review-pr/SKILL.md) lenses. Findings that edit the same code or the same concern share a PR (user, 2026-11-08); the rest are parked in idea files ([Where the notes went](#where-the-notes-went)).

**A brief covers one view.** A PR that adds, changes or removes chrome waits on a brief for each view it touches. Briefs and PR plans are written just before the PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)). A bugfix or a behavior change with no new chrome is not gated.

## Status

**Open.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** a run of small, independent improvements, each batch landing on its own.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project. A batch adds its outcomes here when it is cut.

**Batch 1 — Person page audit fixes** (user, 2026-11-08; [audit](person-page-audit.md)):

1. **A burial-only life reads in order.** A person known only from a parish burial reads Birth unknown, the undated events, Death unknown, then the burial, in the list and on the lifeline.
2. **A crafted project cannot stall the Person page.** A project holding an event with 10,000 places, a person with 20,000 children, a person with 50,000 parents, and 20,000 of one person's records in one Evidence graph opens, and each Person section draws in under a second.
3. **A long Timeline unfolds a page at a time.** A person with 2,000 events, one of them with 500 participants, opens on Timeline without a stall; each Show more reveals at most a page of 50; every event still has its point on the lifeline.
4. **Provenance shows today's words.** Rename a Source, or fix a Citation's label or page: the Vitals provenance sheet, the Gallery captions and the likeness viewer's record row show the new words, as the Sources section does.
5. **A failed reload says so.** When a section's reload fails, its rows stay and a notice with Try again sits above them; a first load that fails offers Try again too; VoiceOver hears both.
6. **The diagram says when it did not count.** A person whose records were too many to compare shows a note in the diagram column instead of drawing every link thin.
7. **Keyboard and VoiceOver keep their place.** ⌘1–⌘5 announces the section and leaves focus on the switcher; removing a record that keeps the person is announced; Show more in Family, Timeline and Sources moves focus to the first item it reveals.

**Scale.** Real data, which must stay as fast as today: a person in 50 Evidence graphs (Spike 13's budget) with one or two records per graph, 200 events with 10 participants each (12 shown per event at Wide), 20 children in four parent sets, one to three places per event. Crafted data, which only has to stay bounded: each axis in bar 2 and 3, at least 20 times past its cap.

## Where the notes went

| Audit item | PR | Brief |
| --- | --- | --- |
| P-1 A missing death sits after a burial | S16-02 | — |
| P-2 Folding an event's places is quadratic | S16-01 | — |
| P-3 The Family read is quadratic in a crafted wide family | S16-01 | — |
| P-4 The claims read rescans its graph per record | S16-01 | — |
| P-5 A failed refresh is silent | S16-04 | S16-D1 |
| P-6 Vitals and Gallery keep old Source words | S16-04 | — |
| P-7 The Timeline builds every event | S16-02 | — |
| P-10 The diagram draws uncounted links as weak | S16-01 (engine), S16-05 (note) | S16-D1 |
| P-11 Switching sections is silent | S16-06 | — |
| P-12 A Remove that keeps the person is silent | S16-06 | — |
| P-13 Section plumbing | S16-03 | — |
| P-14 Fold focus copied per section | S16-06 | — |

**Back in the ideas folder** (user, 2026-11-08) —
- [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md): P-8 (participants sent once per event), P-9 (record links memo and a household-sized budget), P-23 (siblings sent two or three times), and the parts of P-7 a render cap leaves (memoising section content, the Sources diagram's rebuild on focus, lazy cards after an unfold).
- [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md): P-15, P-17, P-18, P-19, P-20.
- [`ideas/structural-debt.md`](../../ideas/structural-debt.md): P-16, P-21, P-22, P-25 (A-41), P-26, P-27.
- [`ideas/share-packages.md`](../../ideas/share-packages.md) § Harden before sharing: P-24.

## Working rules

These are the calls for this spike. PR plans treat them as binding.

### Batch 1 — Person page audit fixes

- **(Batch 1) Crafted fan-out is capped, not engineered** (user, 2026-11-08). A cost that only a crafted or badly broken project reaches gets a fixed cap checked before the work, not a faster algorithm. Real data never reaches a cap. Past a cap a read keeps the first items in the order it already reads them and drops the rest without saying so; the record links read is the exception (rule "The record links read is bounded and says so"). The caps: an event's places before they fold, 20 (`conclusionsummaries`); the Family read's relatives per role (parents, children, spouses, siblings) before their links load, 100 (`persondetail`); implied siblings only through a person with at most `kinship.Rules.MaxPivotEnd` parents, the bound `parentUnits` already uses to stop pairing; a person's records per Evidence graph in the claims read, `RecordLinks.MaxEntityMembers` (10, user 2026-11-07: "malformed"). A cap matching does not tune is a package constant beside the read it bounds, with a `// decided: user, 2026-11-08` comment, as `summaryChainDepth` and `deleteimpact.listedCap` are; an existing registry value is reused, not copied. Each cap has a timing test at least 20 times past it (bar 2). Binds S16-01.
- **(Batch 1) The record links read is bounded and says so.** Go sends at most `RecordLinks.MaxLines` lines (75,000, the number the diagram draws today), strong lines first, and the response says when the read did not count facts (past `MaxReadWork`) or dropped lines. The app's `PersonRecordLinksContent.maxArcs` and its strong-first pass go, since Go owns that bound; the 1,000-node cap stays in Swift, where it bounds the circle. The diagram shows a note instead of drawing uncounted arcs as weak; the note is about the read, not a line, so Spike 13's rule that the view shows no data about a line holds. Binds S16-01, S16-05.
- **(Batch 1) A marker's empty position has one meaning.** When no dated row qualifies for a placeholder's `beside_event_id` but dated rows exist, Go also names the row the placeholder sits before (a death before the first after-death row; the mirror for birth), so an empty field means only "nothing dated". Swift places what Go names (Spike 13 rule "Positions are data"). Binds S16-02.
- **(Batch 1) A Timeline fold bounds its work** (user, 2026-11-08: a render cap or load more, kept simple). Swift builds rows and person cards only for what the folds show; a folded row keeps its id, span and spoken label for the lifeline and the markers. Each "Show N more" reveals at most one page of 50, events and participants alike, so a crafted list unfolds page by page and a real event's participants unfold in one click. The folds stay at 50 events and 5 / 12 participants (user, 2026-11-03). The lifeline draws every point and makes a click target only for a shown row. Binds S16-02.
- **(Batch 1) One open section, one list of reads.** `WorkspaceLocation.openPageSection` resolves the open page section once (`pageSection ?? offered.first`). `ConclusionPageSection` declares its reads as one list, each tagged own, companion or supporting, in place of `queryKey`, `companionQueryKey` and `supportingQueryKeys`. The place registry, the page's refreshing state, Try again (S16-04) and the section hosts all read that list. The section host's identifier root comes from the page identifier. [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md) gains the page-section sub-place and [`macos-client-patterns.md`](../../macos-client-patterns.md) its sectioned-page lines. Binds S16-03, S16-04, S16-06.
- **(Batch 1) A failed reload keeps the rows and says so.** The Conclusion page's presentation carries a failure beside a value; the page and each section host show the kit's `PVRetryCallout` (it announces and has Try again; precedent S14-D7) above the stale content, in its compact form in a column, and on a first-load failure. `QueryHandle.read` is unchanged for its other callers. The page frame is shared, so Event and Place pages gain it too. Binds S16-04.
- **(Batch 1) A cached detail follows the Sources its records cite.** `.conclusionDetail` refreshes when a write names a Source that one of its field records, derived-row owners or pictures cites (`applyHiddenPolicy(followsValue: .source, …)`), so a Source rename and a Citation or Artifact label edit reach Vitals, the Gallery and the viewer, on Event and Place pages as well. Binds S16-04.
- **(Batch 1) Folds move focus through the Show-more row.** `ConclusionShowMoreRow` takes the item it reveals and claims focus for it through `FocusClaim` after expanding; Family, Timeline and Sources use it, and the private `Task` hops go. One focus keeper holds focus across a width change for Gallery and Timeline. Sources' folds become an enum like the other sections'. Binds S16-06.
- **(Batch 1) Every outcome on the page is spoken.** No Toast is a visual choice (S13-17 RC-6), not a silence: a section switch announces the section, and a Remove that keeps the person announces it, with spoken-only strings. Binds S16-06.
- **(Batch 1) Refused: memoising and re-shaping for speed** (user, 2026-11-08). P-8 and P-9 are descoped, and P-7 gets a render cap, not a memo. Their notes are in [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md); measure before pulling any in.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S16-01 | Cap the Person page reads against crafted projects | 1a | Go: caps checked before the work on an event's places, the Family read's relatives and implied siblings, and a person's records per graph in the claims read; the record links read sends at most 75,000 lines and says when it did not count or dropped lines (one wire field, not shown yet) | — | S14-40 (Spike 14) | — | Scoped |
| S16-02 | The Timeline places a missing death and builds what it shows | 1a | Go: a placeholder life marker names the row it sits before when none qualifies to sit beside; Swift places the marker and its lifeline stub there, builds rows and cards only for what the folds show, unfolds a page of 50 at a time, and makes click targets only for shown rows | — | — | — | Scoped |
| S16-03 | One open section and one list of section reads | 1b | Preparatory, Swift only: `WorkspaceLocation.openPageSection` and one read list per `ConclusionPageSection`, read by the place registry, the page, the Person fill and the section hosts; `add-workspace-location` teaches page sections; no visible change | — | — | — | Scoped |
| S16-04 | The Person page keeps its data current or says it could not | 1b | A failed reload keeps the rows under an announced Try again notice, also on a first-load failure, on the page and in each section; `.conclusionDetail` follows the Sources its records cite | S16-D1 | S16-03, S16-D1 | — | Scoped |
| S16-05 | The record links diagram says when it did not count | 1b | Swift: the diagram column shows a note when the read did not count facts or dropped lines, instead of drawing uncounted arcs as weak | S16-D1 | S16-01, S16-D1 | — | Scoped |
| S16-06 | Keyboard focus and VoiceOver on the Person page | 1c | Swift: the Show-more row moves focus to what it reveals (Family gains it), one focus keeper across width changes, Sources' folds as an enum; ⌘1–⌘5 announces the section and keeps focus on the switcher; a Remove that keeps the person is announced | — | S16-03 | — | Scoped |
| S16-99 | Dogfood close / docs | — | Bar pass; the ship logs' follow-ups triaged; the audit's [Close the loop](person-page-audit.md#close-the-loop) classes added to the lenses' Known traps; [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S16-D1 | Person page: notices for a failed reload (the header band, a full-width section, the Sources diagram column) and for a diagram that did not count | **S16-04**, **S16-05** | An enhancement of the Spike 13 Person page boards (S13-D1, S13-D10). Reuses `PVRetryCallout` as the Promote page does (S14-D7). Draws Regular and Wide. The Conclusion page frame is shared, so the board also says the Event and Place pages take the same notice. Copy: the notice titles, Try again, the diagram note |

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify in the app when the slice lands.

```text
BATCH 1 — Person page audit fixes

SLICE 1a — Bounded reads and the Timeline
  S16-01  Cap the Person page reads against crafted projects      (after S14-40)
  S16-02  The Timeline places a missing death and builds what it shows
  Check: bar 1, 2 and 3.

SLICE 1b — Current data and honest notices
  S16-03  One open section and one list of section reads          (preparatory)
  ✎ S16-D1 ──▶ S16-04  The Person page keeps its data current or says it could not
  ✎ S16-D1 ──▶ S16-05  The record links diagram says when it did not count
  Check: bar 4, 5 and 6.

SLICE 1c — Keyboard and VoiceOver
  S16-06  Keyboard focus and VoiceOver on the Person page
  Check: bar 7.

S16-99 — Dogfood close / docs
```

S16-01, S16-02 and S16-03 can run in parallel. S16-01 and S16-02 both add a field to `api/proto/engine.proto` and its `GoStore` mapping, and S16-03, S16-04 and S16-06 all edit `ConclusionDetailPage.swift`: whichever lands second rebases (and regenerates the protobuf).

### Dependencies at a glance

```text
S14-40 ──▶ S16-01 ──▶ S16-05
S16-D1 ──▶ S16-04
S16-D1 ──▶ S16-05
S16-03 ──▶ S16-04
S16-03 ──▶ S16-06
S16-02
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Each names its audit items; the evidence, the failing tests and the line numbers are in [the audit](person-page-audit.md).

### Slice 1a — Bounded reads and the Timeline

#### S16-01 — Cap the Person page reads against crafted projects

**Ships.** The caps of rule "Crafted fan-out is capped, not engineered", each checked before the work it bounds, with the audit's seed test as a timing test at least 20 times past the cap: an event's places in `eventPlaces` before `FoldLocations` (P-2); the Family read's relatives per role before `LoadRelationships`, and no implied siblings through a person past `MaxPivotEnd` parents (P-3); a person's records per graph in `memberGraphs` (P-4). The record links read sends at most `RecordLinks.MaxLines` lines, strong first, and its response says when it did not count or dropped lines (P-10); `GoStore` decodes the field and the app does not show it yet.
**Not this PR.** Faster algorithms for any of these (refused); the note in the diagram (S16-05); the record links memo ([idea](../../ideas/person-page-performance.md)); `edgeKey` ([`share-packages.md`](../../ideas/share-packages.md)).
**Check.** Bar 2; a handler test where a read past `MaxReadWork` says it did not count; the real-data tests (`TestPersonFamilyQueryCount`, the record links budgets at 50 graphs) unchanged.
**Context.** Audit P-2, P-3, P-4, P-10; rules above; `core/database/conclusionsummaries/compose.go` (`eventPlaces`), `placechains.go`, `core/database/persondetail/family.go`, `claims.go`, `recordlinks.go`, `core/kinship/siblings.go`, `core/recordlinks/recordlinks.go`, `core/matchregistry/recordlinks.go` (a new knob needs its inventory row, [`matching-model.md`](../../matching-model.md) and [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md)), `api/ffi/handlers/person_detail.go`. **Depends on S14-40** (Spike 14, Planned): it adds input scopes to `eventPlaces` and `placechains.go`, so this PR rebases on it.

#### S16-02 — The Timeline places a missing death and builds what it shows

**Ships.** `lifeMarkers` names the row a placeholder sits before when no dated row qualifies to sit beside, on the wire and in `CatalogLifeMarker`; `PersonTimelineContent` inserts the marker before that row and `PersonTimelineLifeline` puts its stub at that row's start (P-1). `PersonTimelineContent` builds rows and person cards only for what the folds show, each Show more reveals a page of 50, and the lifeline makes click targets only for shown rows (P-7). Both edit `PersonTimelineContent.init`, so they ship together.
**Not this PR.** Memoising the content, participants once per read, lazy cards ([idea](../../ideas/person-page-performance.md)); keeping the scale after a date edit (P-20, [idea](../../ideas/person-page-followups.md)).
**Check.** Bar 1 and 3; P-1's failing test and a Go `TestLifeMarkers` case pass; `PersonTimelineLifelineTests` asserts click targets only for shown rows; a content test builds 2,000 events × 500 participants and touches only shown rows and cards.
**Context.** Audit P-1, P-7; rules "A marker's empty position has one meaning", "A Timeline fold bounds its work"; Spike 13 rules "The timeline is a lifeline" and "Timeline lists fold", [R-24a](../spike-13/person-detail.md#3-timeline); `core/database/persondetail/lifeline.go`, `timeline.go`, `api/ffi/handlers/person_detail.go`, `macos/App/Features/Conclusions/PersonTimelineContent.swift`, `PersonTimelineLifeline.swift`, `PersonTimelineSection.swift`, `ConclusionShowMoreRow.swift`.

### Slice 1b — Current data and honest notices

#### S16-03 — One open section and one list of section reads

**Ships.** `WorkspaceLocation.openPageSection`; one read list per `ConclusionPageSection` replacing its three slot functions, read by `PlaceRegistry`, the page's refreshing state, the Person fill and both section hosts; the host's identifier root from the page identifier; [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md) and [`macos-client-patterns.md`](../../macos-client-patterns.md) updated (P-13). No visible change.
**Not this PR.** Event and Place sections; a per-case spec table for label and glyph (the next section is one list entry either way).
**Check.** Unit tests: a first section that reads its own key is in the place's query keys and stays visible across a write; the page, the place and the fill agree on the open section for every location; the Person page's identifiers are unchanged.
**Context.** Audit P-13; rule "One open section, one list of reads"; `macos/App/Features/Workspace/WorkspaceLocation.swift`, `Session/PlaceRegistry.swift`, `WorkspaceSession.swift` (visible keys), `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `ConclusionPageSectionHost.swift`, `PersonDetailView.swift`, `PersonSourcesSection.swift`; `PlaceRegistryTests`, `WorkspaceNavigationTests`.

#### S16-04 — The Person page keeps its data current or says it could not

**Ships.** The Conclusion page's presentation carries a failure beside its value; the page and each section host draw `PVRetryCallout` above stale content (compact in the diagram column) and on a first-load failure, as S16-D1 draws them; Try again re-queries the open section's reads (P-5). `.conclusionDetail` follows the Sources its records cite (P-6).
**Not this PR.** `QueryHandle.read` for its other callers; Back restoring a section's place (P-19, [idea](../../ideas/person-page-followups.md)).
**Check.** Bar 4 and 5; P-5's FakeStore test and P-6's `WorkspaceSessionTests` case pass; manual: a Source rename, then Back to a Person whose birth cites it.
**Context.** Audit P-5, P-6; rules "A failed reload keeps the rows and says so", "A cached detail follows the Sources its records cite"; brief S16-D1; S14-20 for the retry pattern; `ConclusionDetailPage.swift`, `ConclusionPageSectionHost.swift`, `Session/CatalogQueryRegistry.swift`, `macos/App/DesignSystem/Components/Callout/PVRetryCallout.swift`; [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md); [`macos-client-patterns.md`](../../macos-client-patterns.md) "Conclusion keys" says the detail key follows only `touchedHandles` and labels, and this PR updates it.

#### S16-05 — The record links diagram says when it did not count

**Ships.** `PersonRecordLinksContent` reads S16-01's field and `PersonRecordLinksColumn` shows S16-D1's note, without uncounted arcs drawn as weak (P-10).
**Not this PR.** Any data about a line (Spike 13 rule "The view is a circular link diagram"); hover on a folded record line (P-15, [idea](../../ideas/person-page-followups.md)).
**Check.** Bar 6; a content test for each of "did not count" and "dropped lines".
**Context.** Audit P-10; rule "The record links read is bounded and says so"; brief S16-D1; `macos/App/Features/Conclusions/PersonRecordLinksContent.swift`, `PersonRecordLinksColumn.swift`, `CircularLinkDiagram/`.

### Slice 1c — Keyboard and VoiceOver

#### S16-06 — Keyboard focus and VoiceOver on the Person page

**Ships.** `ConclusionShowMoreRow` claims focus for the item it reveals through `FocusClaim`, and Family, Timeline and Sources use it (Family gains it); one focus keeper across width changes for Gallery and Timeline; Sources' folds as an enum (P-14). The switcher takes a focus handle; a section switch announces the section and claims the selected segment when the old focus is gone (P-11). A Remove that keeps the person posts a spoken-only line (P-12).
**Not this PR.** Back restoring scroll, folds and focus (P-19); Remove wording (P-17) ([idea](../../ideas/person-page-followups.md)).
**Check.** Bar 7; content tests name the revealed item for each fold, Family included; manual VoiceOver steps for ⌘3 from a Family tile and for a Remove of one of two records.
**Context.** Audit P-11, P-12, P-14; rules "Folds move focus through the Show-more row", "Every outcome on the page is spoken", "One open section, one list of reads"; `macos/App/Platform/FocusClaim.swift`, `ConclusionShowMoreRow.swift`, `ConclusionDetailPage.swift`, `PersonFamilySection.swift`, `PersonTimelineSection.swift`, `PersonSourcesSection.swift`, `PersonSourcesContent.swift`, `ConclusionGallerySection.swift`, `DesignSystem/Components/Chip/PVSegmentedControl.swift`.
