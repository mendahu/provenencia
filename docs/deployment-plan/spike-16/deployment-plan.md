# Deployment plan — Spike 16

Spike 16 is an odds-and-ends spike ([odd and even spikes](../README.md#odd-and-even-spikes)): small fixes, UX refinements and follow-ups, in batches, each a slice with its own Check. Batch 1 takes the high and medium findings of the [Person page audit](person-page-audit.md) (2026-11-08), an area audit of Spike 13's Person page with all five [`review-pr`](../../../.agents/skills/review-pr/SKILL.md) lenses. Findings that edit the same code or the same concern share a PR (user, 2026-11-08), unless the map review found a reason to keep them apart ([review](deployment-plan.review.md)); the rest are parked in idea files ([Where the notes went](#where-the-notes-went)).

**A brief covers one view.** A PR that adds, changes or removes chrome waits on a brief for each view it touches. Briefs and PR plans are written just before the PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)). A bugfix or a behavior change with no new chrome is not gated.

## Status

**Open.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** a run of small, independent improvements, each batch landing on its own.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project. A batch adds its outcomes here when it is cut.

**Batch 1 — Person page audit fixes** (user, 2026-11-08; [audit](person-page-audit.md)):

1. **A burial-only life reads in order.** A person known only from a parish burial reads Birth unknown, the undated events, Death unknown, then the burial, in the list and on the lifeline. A person with a dated baptism and burial and no death has Death unknown just before the burial, on the list and on the lifeline, so the life between them is drawn before the death.
2. **A crafted project cannot stall the Person page.** A project holding an event with 10,000 places, a person with 20,000 children, a person with 50,000 parents, and 20,000 of one person's records in one Evidence graph opens, and each Person section draws in under a second.
3. **A long list unfolds a page at a time.** A person with 2,000 events, one of them with 500 participants, opens on Timeline without a stall; each Show more, in any section, reveals at most a page of 50 and moves focus to the first item it reveals; every event still has its point on the lifeline. Hovering a graph on the record links diagram whose record line is folded lights up that Source's Show more, and a date edit from another window rescales the lifeline to the new range.
4. **Provenance shows today's words.** Rename a Source, or fix a Citation's label or page: the Vitals provenance sheet, the Gallery captions and the likeness viewer's record row show the new words, as the Sources section does.
5. **A failed reload says so.** When a section's reload fails, its rows stay and a notice with Try again sits above them; a first load that fails offers Try again too; VoiceOver hears both.
6. **The diagram says when it is partial.** A person whose records were too many to compare, or too many to draw, shows a note in the diagram column instead of drawing every link thin or dropping graphs silently.
7. **Keyboard and VoiceOver keep their place.** ⌘1–⌘5 announces the section and leaves focus on the selected segment; removing a record that keeps the person is announced; Show more in Family, Timeline and Sources moves focus to the first item it reveals (bar 3).
8. **One word for a person's records.** The Person page's header counts "7 records" (the Event and Place headers too), the Remove confirm ends "…, and the person stays", and a person deleted from another window leaves with "This person no longer exists".

**Scale.** Real data, which must stay as fast as today: a person in 50 Evidence graphs (Spike 13's budget) with one or two records per graph, 200 events with 10 participants each (12 shown per event at Wide), 20 children in four parent sets, one to three places per event. Crafted data, which only has to stay bounded: each axis in bar 2 and 3, at least 20 times past its cap.

## Where the notes went

| Audit item | PR | Brief |
| --- | --- | --- |
| P-1 A missing death sits after a burial | S16-02 | — |
| P-2 Folding an event's places is quadratic | S16-08 | — |
| P-3 The Family read is quadratic in a crafted wide family | S16-01 | — |
| P-4 The claims read rescans its graph per record | S16-01 | — |
| P-5 A failed refresh is silent | S16-04 | S16-D1 |
| P-6 Vitals and Gallery keep old Source words | S16-09 | — |
| P-7 The Timeline builds every event | S16-07 (pages), S16-02 (build what is shown) | — |
| P-10 The diagram draws uncounted links as weak | S16-01 (engine), S16-05 (note) | S16-D1 |
| P-11 Switching sections is silent | S16-06 | — |
| P-12 A Remove that keeps the person is silent | S16-06 | — |
| P-13 Section plumbing | S16-03 | — |
| P-14 Fold focus copied per section | S16-07 | — |
| P-16 The leave-on-delete decision is not tested (map review K-8) | S16-06 | — |
| P-15 Hovering a node whose record line is folded highlights nothing | S16-07 | — |
| P-17 "handle", "members" and "This record no longer exists" on the Person page | S16-04 (the not-found line), S16-06 (the count and the confirm) | S16-D1 (the not-found line) |
| P-20 The Timeline keeps an old scale after a date edit | S16-02 | — |
| P-21 `ListPersonClaims` answers for a non-Person handle | S16-01 | — |
| P-22 (part) The record links wire's unread fields and missing `source_ref` | S16-01 | — |
| P-24 `edgeKey` joins ids with no separator | S16-01 | — |
| P-26 (part) `ConclusionDetailPage.swift` holds 15 types | S16-03 | — |
| `CatalogQueryStalePolicy` is unused ([`structural-debt.md`](../../ideas/structural-debt.md), S14-11 plan) | S16-09 | — |

The last eight rows were parked in idea files and folded into the PRs that already touch their files (user, 2026-11-08).

**Back in the ideas folder** (user, 2026-11-08) —
- [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md): P-8 (participants sent once per event), P-9 (record links memo and a household-sized budget), P-23 (siblings sent two or three times), and the parts of P-7 a render cap leaves (memoising section content, the Sources diagram's rebuild on focus, lazy cards after an unfold).
- [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md): P-18, P-19, and the Conclusion lists' silent refresh failure (map review K-2).
- [`ideas/structural-debt.md`](../../ideas/structural-debt.md): P-22 (the rest), P-25 (A-41), P-26 (the rest), P-27.

## Working rules

These are the calls for this spike. PR plans treat them as binding.

### Batch 1 — Person page audit fixes

- **(Batch 1) Crafted fan-out is capped, not engineered** (user, 2026-11-08). A cost that only a crafted or badly broken project reaches gets a fixed cap checked before the work, not a faster algorithm. Real data never reaches a cap. Past a cap a read keeps the first items in the order it already reads them and drops the rest without saying so; the record links read is the exception (rule "The record links read is bounded and says so"). The caps:
  - an event's places, 20, applied in one helper that the four reads of an event's places in `conclusionsummaries/compose.go` (`eventPlaceGraph`, `lifeOwners`, `composeEvents`, `eventPlaces`) all use, so no read walks the rest (S16-08);
  - the Family read's relatives per role (parents, children, spouses, siblings), 100, in one capped kin per person that both of `Families`' loops use (the relatives it loads, and `Around` and `FamilyUnits`) (S16-01);
  - implied siblings only through a person with at most `kinship.Rules.MaxPivotEnd` parents, the bound `parentUnits` already uses to stop pairing (S16-01);
  - a person's records per Evidence graph in the claims read, `recordlinks` `Rules.MaxEntityMembers` (10, user 2026-11-07: "malformed") (S16-01).

  A cap matching does not tune is a package constant beside the read it bounds, with a `// decided: user, 2026-11-08` comment, as `graphcache` `maxWalkEdges` and `PersonRecordLinksContent.maxArcs` are; an existing registry value is reused, not copied. Each cap has a timing test at least 20 times past it (bar 2). For crafted data only, this supersedes "nothing is hidden for good" and "Go still returns every member" in Spike 13's fold rules (`spike-13/deployment-plan.md` "Timeline lists fold"; `PersonFamilyContent.swift`). Binds S16-01, S16-08.
- **(Batch 1) The record links read is bounded and says so.** `recordlinks.Pairs` reports whether it counted (it already knows: past `MaxReadWork`, or no pairs when the lines alone pass it). `persondetail.RecordLinks` sends at most `maxRecordLinkLines` lines (75,000, the number the diagram draws today), strong lines first, as a package constant that carries `maxArcs`'s `// decided:` provenance, not a match-registry knob, and the response says when it did not count or dropped lines. Swift's `PersonRecordLinksContent.maxArcs` and its strong-first pass go in the same PR, since Go now owns that bound. The 1,000-node cap stays in Swift, where it bounds the circle, and the diagram's note covers it too: one note says the drawing is partial, whichever side cut it. The note is about the read, not a line, so Spike 13's rule that the view shows no data about a line holds. Binds S16-01, S16-05.
- **(Batch 1) A marker's position is one value.** A life marker that has no span of its own sits beside one dated row or before one, never both: one `oneof` on the wire (beside or before) and one Swift enum, in place of an empty `beside_event_id` that means both "nothing dated" and "nothing qualifies". A placeholder death sits just before the earliest dated after-death row (burial, funeral) whenever one is dated, else beside the last dated row; a placeholder birth mirrors it (just after the last dated before-birth row, else before the first dated row) (user, 2026-11-08: "just before burial"). This replaces Spike 13's "just after the latest dated event that cannot come after death", which drew the life between a baptism and a burial after the death; [R-24a](../spike-13/person-detail.md#3-timeline) and Spike 13's rule "The timeline is a lifeline" say so. Swift places what Go names, and the lifeline puts a "before" stub at that row's start (Spike 13 rule "Positions are data"). Binds S16-02.
- **(Batch 1) One paged fold that moves focus** (user, 2026-11-08: a render cap or load more, kept simple; map review K-1). `ConclusionFold`, already "the one fold rule of a Conclusion page's lists", gains a page: each Show more reveals at most 50 more, in every section, and one fold state type (which folds, how many pages each has revealed) replaces the per-section `Set`s and Sources' `rowsExpanded`. `firstRevealed` names the first item of the page just revealed, and `ConclusionShowMoreRow` claims focus for it through `FocusClaim`; the private `Task` hops go. One focus keeper in `macos/App/Platform/`, beside `FocusClaim`, holds focus across a width change for Gallery and Timeline ([`macos-client-patterns.md`](../../macos-client-patterns.md) Wide rule 5). The folds' first limits stay as they are (50 events, 5 / 12 participants; user, 2026-11-03). Binds S16-07, S16-02.
- **(Batch 1) The Timeline builds what it shows.** `PersonTimelineContent` builds rows and person cards only for what the folds show; a folded row keeps its id, span and spoken label for the lifeline and the markers. The lifeline draws every point and makes a click target only for a shown row. Binds S16-02.
- **(Batch 1) One open section, one list of reads.** `WorkspaceLocation.openPageSection` resolves the open page section once (`pageSection ?? offered.first`). `ConclusionPageSection` declares its reads as one list, each tagged own, companion or supporting, in place of `queryKey`, `companionQueryKey` and `supportingQueryKeys`. The place registry, the page's refreshing state, Try again (S16-04) and the section hosts all read that list. The section host's identifier root comes from the page identifier. [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md) gains the page-section sub-place and [`macos-client-patterns.md`](../../macos-client-patterns.md) its sectioned-page lines. Binds S16-03, S16-04, S16-06.
- **(Batch 1) A refresh failure is read beside its value** (map review K-2). `QueryHandle` gains one reading that is loading, failed, or loaded with an optional refresh failure, and a failure revision that bumps on each failure so a notice can re-announce; `read` is unchanged for its other callers. `ConclusionDetailPresentation` is built from it, and the page and each section host show the kit's `PVRetryCallout` (it announces and has Try again; precedent S14-D7) above the stale content, in its compact form in a column, and on a first-load failure. The page frame is shared, so Event and Place pages gain it too. The Conclusion lists' presentation adopts it later, with its own brief ([idea](../../ideas/person-page-followups.md)); `EvidenceGraphModel`'s and `PromoteModel`'s attempt counters may follow. [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md) ("Stale-while-revalidate") and the `QueryHandle` row of [`macos-client-patterns.md`](../../macos-client-patterns.md) teach it. Binds S16-04.
- **(Batch 1) A cached detail follows the Sources its records cite.** `.conclusionDetail` refreshes when a write names a Source that one of its field records, derived-row owners or pictures cites (`applyHiddenPolicy(followsValue: .source, …)`), so a Source rename and a Citation or Artifact label edit reach Vitals, the Gallery and the viewer, on Event and Place pages as well. Fan-out: `sourceScopes` is the audit "last activity" scope, which 16 effect entries set, so any write inside a cited Source revalidates a visible detail (one read, `GetConclusionDetail`); a hidden detail is evicted, as today. [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md)'s slot list gains `.source`, and [`macos-client-patterns.md`](../../macos-client-patterns.md) "Conclusion keys" says what the detail key follows. Binds S16-09.
- **(Batch 1) Every outcome on the page is spoken.** No Toast is a visual choice (S13-17 RC-6), not a silence: a section switch announces the section, and a Remove that keeps the person announces it, with spoken-only strings. The leave-or-stay decision after a Remove moves out of `PersonSourcesSection.confirm` into a function a unit test reaches, which returns leave, or stay with the announcement. Binds S16-06.
- **(Batch 1) A person's records are records** (user, 2026-11-08). The page uses one word for the Identity Claims behind a canonical entity. The header's count reads "7 records" on every Conclusion page (one string, `conclusions.memberCount`). The Unpromote confirm ends "…, and the person stays", worded per kind like `confirmMessageDeletes`, the Evidence graph's confirm included. A Person page whose person was deleted or merged elsewhere says "This person no longer exists" in its leave Toast and its section notices; the kind supplies that line, so Event and Place keep the shared one. Binds S16-04, S16-06.
- **(Batch 1) Refused: memoising and re-shaping for speed** (user, 2026-11-08). P-8 and P-9 are descoped, and P-7 gets a render cap and pages, not a memo. Their notes are in [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md); measure before pulling any in.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S16-01 | Cap the Person page reads against crafted projects | 1a | Go: the Family read's relatives per role and implied siblings, and a person's records per graph in the claims read, capped before the work; the record links read sends at most 75,000 lines and says when it did not count or dropped lines, and Swift's own arc cap goes; one Person check for every read, an unambiguous `edgeKey`, and `source_ref` on the record links wire in place of two unread fields | — | — | — | Scoped |
| S16-08 | Cap the places an event folds | 1a | Go: one helper for an event's places, deduped and capped at 20, used by the four reads in `conclusionsummaries` | — | S14-40 (Spike 14) | — | Scoped |
| S16-07 | One paged fold that moves focus | 1a | Preparatory, Swift: `ConclusionFold` reveals a page of 50 at a time in every section, one fold state type, Show more moves focus to the first revealed item (Family gains it), one focus keeper across width changes; hovering a graph whose record line is folded lights up its Show more | — | — | — | Scoped |
| S16-02 | The Timeline places a missing death and builds what it shows | 1a | Go and Swift: a placeholder life marker sits beside or before one row, so a missing death precedes a burial; the Timeline builds rows and cards only for what the folds show and makes click targets only for shown rows; a date edit rescales the lifeline | — | S16-07 | — | Scoped |
| S16-09 | Conclusion pages follow the Sources their records cite | 1b | Swift: `.conclusionDetail` refreshes when a write names a Source its records cite, so Vitals, the Gallery and the viewer show today's Source, Citation and Artifact words; the unused `CatalogQueryStalePolicy` goes | — | — | — | Scoped |
| S16-03 | One open section and one list of section reads | 1b | Preparatory, Swift only: `WorkspaceLocation.openPageSection` and one read list per `ConclusionPageSection`, read by the place registry, the page, the Person fill and the section hosts; `add-workspace-location` teaches page sections; no visible change; `ConclusionDetailPage.swift` split into the files its types belong to | — | — | — | Scoped |
| S16-04 | A failed reload keeps the rows and says so | 1b | A `QueryHandle` reading with a refresh failure beside the value; the page and each section show an announced Try again notice above stale rows and on a first-load failure; a deleted person's page says the person no longer exists | S16-D1 | S16-03, S16-D1 | — | Scoped |
| S16-05 | The record links diagram says when it is partial | 1b | Swift: the diagram column shows a note when the read did not count facts, Go dropped lines, or the circle dropped graphs | S16-D1 | S16-01, S16-D1 | — | Scoped |
| S16-06 | Section switches and removals are spoken | 1c | Swift: ⌘1–⌘5 announces the section and focus lands on the selected segment (a `PVSegmentedControl` fix); a Remove that keeps the person is announced, from a leave-or-stay decision a unit test reaches; the header counts records and the Remove confirm says the person stays | — | S16-03 | — | Scoped |
| S16-99 | Dogfood close / docs | — | Bar pass; the ship logs' follow-ups triaged; the audit's [Close the loop](person-page-audit.md#close-the-loop) classes added to the lenses' Known traps; [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S16-D1 | Person page: notices for a failed reload (the header band, a full-width section, the Sources diagram column) and for a record links diagram that is partial | **S16-04**, **S16-05** | An enhancement of the Spike 13 Person page boards (S13-D1, S13-D10). Reuses `PVRetryCallout` as the Promote page does (S14-D7). Draws Regular and Wide. The Conclusion page frame is shared, so the board also says the Event and Place pages take the same notice. Copy: the notice titles, Try again, the diagram's partial note |

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify in the app when the slice lands.

```text
BATCH 1 — Person page audit fixes

SLICE 1a — Bounded reads, folds and the Timeline
  S16-01  Cap the Person page reads against crafted projects
  S16-08  Cap the places an event folds                      (after S14-40)
  S16-07  One paged fold that moves focus                   (preparatory)
  S16-02  The Timeline places a missing death and builds what it shows
  Check: bar 1, 2 and 3.

SLICE 1b — Current data and honest notices
  S16-09  Conclusion pages follow the Sources their records cite
  S16-03  One open section and one list of section reads    (preparatory)
  ✎ S16-D1 ──▶ S16-04  A failed reload keeps the rows and says so
  ✎ S16-D1 ──▶ S16-05  The record links diagram says when it is partial
  Check: bar 4, 5 and 6.

SLICE 1c — Keyboard and VoiceOver
  S16-06  Section switches and removals are spoken
  Check: bar 7 (with S16-07's focus moves).

S16-99 — Dogfood close / docs
```

S16-01, S16-07, S16-09 and S16-03 can start together; S16-08 waits for Spike 14's S14-40. Shared files, so whichever lands second rebases:
- S16-01 and S16-02 both change `api/proto/engine.proto` and its `GoStore` mapping (and regenerate the protobuf).
- S16-03, S16-04 and S16-06 all edit `ConclusionDetailPage.swift`; S16-03 and S16-07 both edit `PersonSourcesSection.swift` and `PersonTimelineSection.swift`.
- S16-04 and S16-05 both change the Sources diagram column (`PersonSourcesSection` `diagram`, `PersonRecordLinksColumn`): S16-04 lands first, and S16-05 places its note beside S16-04's notice.

### Dependencies at a glance

```text
S14-40 ──▶ S16-08
S16-01 ──▶ S16-05
S16-07 ──▶ S16-02
S16-03 ──▶ S16-04
S16-03 ──▶ S16-06
S16-D1 ──▶ S16-04
S16-D1 ──▶ S16-05
S16-09
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Each names its audit items; the evidence, the failing tests and the line numbers are in [the audit](person-page-audit.md).

### Slice 1a — Bounded reads, folds and the Timeline

#### S16-01 — Cap the Person page reads against crafted projects

**Ships.** The Person-read caps of rule "Crafted fan-out is capped, not engineered", each checked before the work it bounds, with the audit's seed test as a timing test at least 20 times past the cap: one capped kin per person in `Families`, used by both of its loops, and no implied siblings through a person past `MaxPivotEnd` parents (P-3); a person's records per graph in `memberGraphs` (P-4). `recordlinks.Pairs` reports whether it counted; `persondetail.RecordLinks` sends at most `maxRecordLinkLines` lines, strong first, and the response says when it did not count or dropped lines; `GoStore` decodes it, and Swift's `maxArcs` and strong-first pass go (P-10). Three small items ride along in the same files: every `persondetail` read checks the Person through `personNode`, `Claims` and `Families` included (P-21); `canonlinks` `edgeKey` prefixes each part with its length, beside the sibling cap in `Reader` (P-24); `RecordLinksGraph` gains `source_ref` and loses the unread `source_id` and `source_type_id`, so the Sources section no longer hands `graphSourceRefs` to `PersonRecordLinksContent`, and the "name, else ref" and "title, else ref" fallbacks are written once (P-22, part).
**Not this PR.** Faster algorithms (refused); an event's places (S16-08); the note in the diagram (S16-05); the record links memo ([idea](../../ideas/person-page-performance.md)); the claims read owning every graph label (the rest of P-22, [`structural-debt.md`](../../ideas/structural-debt.md)).
**Check.** Bar 2 for the family, sibling and claims axes; a handler test where a read past `MaxReadWork` says it did not count; the real-data tests (`TestPersonFamilyQueryCount`, the record links budgets at 50 graphs) unchanged. The `persondetail` table tests gain a non-Person handle for every read; an `edgeKey` test gives two different links whose joined ids collide today two keys.
**Context.** Audit P-3, P-4, P-10, P-21, P-22, P-24; rules "Crafted fan-out is capped, not engineered", "The record links read is bounded and says so"; `core/database/persondetail/family.go`, `claims.go`, `recordlinks.go`, `core/kinship/siblings.go`, `core/database/canonlinks/reader.go`, `core/recordlinks/recordlinks.go`, `api/ffi/handlers/person_detail.go`, `macos/App/Platform/GoStore.swift`, `macos/App/Features/Conclusions/PersonRecordLinksContent.swift`, `PersonSourcesContent.swift`, `PersonSourcesSection.swift`; [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md) for the field. Lines this makes false, updated here: [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) line 33 (what keeps `Siblings` linear); the comment on `kinship.Siblings` ("no MaxPivotEnd bound"); [`matching-model.md`](../../matching-model.md) rows for `MaxPivotEnd` and `MaxEntityMembers` (their new page readers) and its record links paragraph (the read now says when it did not count); the `PersonFamilyContent.swift` comment "Go still returns every member".

#### S16-08 — Cap the places an event folds

**Ships.** One helper reads an event's places, drops duplicate ids and keeps at most 20 in link order, and the four reads in `conclusionsummaries/compose.go` (`eventPlaceGraph`, `lifeOwners`, `composeEvents`, `eventPlaces`) use it, so `FoldLocations` and the place graph never see more (P-2). A timing test folds an event with 10,000 places under a second.
**Not this PR.** A linear `FoldLocations` (refused); the Person reads (S16-01).
**Check.** Bar 2 for the places axis, including a project whose search projection is rebuilt on open; the summary tests unchanged.
**Context.** Audit P-2; rule "Crafted fan-out is capped, not engineered"; `core/database/conclusionsummaries/compose.go`, `placechains.go`. **Depends on S14-40** (Spike 14, Planned), which adds input scopes to these reads and routes them through one scope `follow`; its map names this PR.

#### S16-07 — One paged fold that moves focus

**Ships.** `ConclusionFold` gains a page of 50 and one fold state type that Family, Timeline and Sources use in place of their `Set`s and Sources' `rowsExpanded`; `firstRevealed` names the first item of the page just revealed and `ConclusionShowMoreRow` claims focus for it through `FocusClaim` (Family gains it; the `Task` hops in the Timeline and Sources go) (P-14, P-7's pages). One focus keeper in `macos/App/Platform/` holds focus across a width change for Gallery and Timeline. When the record links diagram highlights a graph whose record line is folded, that Source's Show more takes the highlight (P-15; the content already maps folded claims to graphs in `claimGraphs`).
**Not this PR.** Building only shown rows and the lifeline's click targets (S16-02); Back restoring a section's folds (P-19, [idea](../../ideas/person-page-followups.md)).
**Check.** Bar 3's paging and bar 7's fold focus: content tests name the revealed item after each page for every section, Family included; a test unfolds 2,000 hidden items in pages of 50. A content test highlights a folded line's Show more.
**Context.** Audit P-7, P-14, P-15; rule "One paged fold that moves focus"; `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift`, `PersonFamilySection.swift`, `PersonFamilyContent.swift`, `PersonTimelineSection.swift`, `PersonTimelineContent.swift`, `PersonSourcesSection.swift`, `PersonSourcesContent.swift`, `ConclusionGallerySection.swift`, `macos/App/Platform/FocusClaim.swift`; [`macos-client-patterns.md`](../../macos-client-patterns.md) "One sectioned detail page" (teaches the fold) and Wide rule 5 (cites the keeper).

#### S16-02 — The Timeline places a missing death and builds what it shows

**Ships.** A life marker's position becomes one value, beside or before a row, on the wire (`oneof`), in `CatalogLifeMarker` and in Go's `LifeMarker`; `lifeMarkers` places a placeholder death just before the earliest dated after-death row whenever one is dated (the mirror for birth), `PersonTimelineContent` inserts the marker there and `PersonTimelineLifeline` puts its stub at that row's start (P-1; user, 2026-11-08). `PersonTimelineContent` builds rows and person cards only for what S16-07's folds show, and the lifeline makes click targets only for shown rows (P-7). The lifeline keeps a chosen scale only while its year range is unchanged as well as its ids, so a date edit from another window rescales it (P-20).
**Not this PR.** Memoising the content, participants once per read, lazy cards ([idea](../../ideas/person-page-performance.md)).
**Check.** Bar 1 and 3; P-1's failing test and Go `TestLifeMarkers` cases pass ("only a burial dated", and "a baptism and a burial": the death sits before the burial, and its stub at the burial's start, not the baptism's end); `PersonTimelineLifelineTests` asserts click targets only for shown rows; a content test builds 2,000 events × 500 participants and touches only shown rows and cards. A placement test moves one event's year under the same ids and gets a new scale.
**Context.** Audit P-1, P-7, P-20; rules "A marker's position is one value", "The Timeline builds what it shows"; Spike 13 rules "The timeline is a lifeline" and "Timeline lists fold", [R-24a](../spike-13/person-detail.md#3-timeline); `core/database/persondetail/lifeline.go`, `timeline.go`, `api/proto/engine.proto`, `api/ffi/handlers/person_detail.go`, `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `macos/App/Features/Conclusions/PersonTimelineContent.swift`, `PersonTimelineLifeline.swift`, `PersonTimelineSection.swift`.

### Slice 1b — Current data and honest notices

#### S16-09 — Conclusion pages follow the Sources their records cite

**Ships.** The `.conclusionDetail` descriptor follows `.source` for the Sources its field records, derived-row owners and pictures cite (P-6). The unused `CatalogQueryStalePolicy` goes while the cache is open here (parked from the S14-11 plan: no descriptor sets `.refetchOnRevisit`): the enum (`CatalogQueryStalePolicy.swift`), the `stalePolicy:` argument of every descriptor arm, `CatalogQueryRegistry.stalePolicy(for:)` and the five test assertions that read it.
**Not this PR.** Event and Place pages' other keys; a narrower Source effect for renames only (the audit scope is what the cache can follow).
**Check.** Bar 4; P-6's `WorkspaceSessionTests` case passes and `personClaimsFollowHandlesAndGraphs`' "the detail page follows only the handle" assertion is updated; manual: a Source rename, then Back to a Person whose birth cites it. `rg CatalogQueryStalePolicy macos` finds nothing.
**Context.** Audit P-6; rule "A cached detail follows the Sources its records cite" (its fan-out); `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `Session/CatalogQueryStalePolicy.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `CatalogQueryRegistryTests.swift`; [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md) (slot list), [`macos-client-patterns.md`](../../macos-client-patterns.md) "Conclusion keys".

#### S16-03 — One open section and one list of section reads

**Ships.** `WorkspaceLocation.openPageSection`; one read list per `ConclusionPageSection` replacing its three slot functions, read by `PlaceRegistry`, the page's refreshing state, the Person fill and both section hosts; the host's identifier root from the page identifier; [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md) and [`macos-client-patterns.md`](../../macos-client-patterns.md) updated (P-13). No visible change. First, `ConclusionDetailPage.swift` (15 types in 736 lines) is split by moving types beside their kin: the gallery model (`ConclusionDetailPicture`, `ConclusionDetailGallery`) to `ConclusionGallerySection.swift`, `ConclusionDetailRowsSection` beside `ReconciledValueRow`, and the Place relationship model and section beside `PlaceRelationshipRow`, so S16-04 and S16-06 edit a smaller file (P-26, part).
**Not this PR.** Event and Place sections; a per-case spec table for label and glyph (the next section is one list entry either way).
**Check.** Unit tests: a first section that reads its own key is in the place's query keys and stays visible across a write; the page, the place and the fill agree on the open section for every location; the Person page's identifiers are unchanged.
**Context.** Audit P-13, P-26; rule "One open section, one list of reads"; `macos/App/Features/Workspace/WorkspaceLocation.swift`, `Session/PlaceRegistry.swift`, `WorkspaceSession.swift` (visible keys), `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `ConclusionPageSectionHost.swift`, `PersonDetailView.swift`, `PersonSourcesSection.swift`; `PlaceRegistryTests`, `WorkspaceNavigationTests`.

#### S16-04 — A failed reload keeps the rows and says so

**Ships.** `QueryHandle`'s reading with a refresh failure beside the value and a failure revision; `ConclusionDetailPresentation` built from it; the page and each section host draw `PVRetryCallout` above stale content (compact in the diagram column) and on a first-load failure, as S16-D1 draws them; Try again re-queries the open section's reads (S16-03); the skill and the doc teach the reading (P-5). A Person page whose person was deleted or merged elsewhere says "This person no longer exists" in its leave Toast and its section notices; the kind supplies the line, so Event and Place keep the shared one (P-17).
**Not this PR.** `QueryHandle.read` for its other callers; the Conclusion lists ([idea](../../ideas/person-page-followups.md)); `EvidenceGraphModel`'s and `PromoteModel`'s attempt counters; Back restoring a section's place (P-19).
**Check.** Bar 5, and bar 8's not-found line; P-5's FakeStore test passes; a test that a second failure bumps the failure revision.
**Context.** Audit P-5, P-17; rules "A refresh failure is read beside its value", "One open section, one list of reads", "A person's records are records"; brief S16-D1; S14-20 for the retry pattern; `macos/App/Features/Workspace/Session/QueryHandle.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `ConclusionPageSectionHost.swift`, `macos/App/DesignSystem/Components/Callout/PVRetryCallout.swift`; [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md) "Stale-while-revalidate", [`macos-client-patterns.md`](../../macos-client-patterns.md) (the `QueryHandle` row).

#### S16-05 — The record links diagram says when it is partial

**Ships.** `PersonRecordLinksContent` reads S16-01's field and its own 1,000-node cut into one partial state, and `PersonRecordLinksColumn` shows S16-D1's note, without uncounted arcs drawn as weak (P-10).
**Not this PR.** Any data about a line (Spike 13 rule "The view is a circular link diagram"); hover on a folded record line (P-15, S16-07).
**Check.** Bar 6; a content test for each of "did not count", "lines dropped" and "graphs dropped".
**Context.** Audit P-10; rule "The record links read is bounded and says so"; brief S16-D1; lands after S16-04 (same column); `macos/App/Features/Conclusions/PersonRecordLinksContent.swift`, `PersonRecordLinksColumn.swift`.

### Slice 1c — Keyboard and VoiceOver

#### S16-06 — Section switches and removals are spoken

**Ships.** `PVSegmentedControl`'s `focus:` handle lands on the selected segment rather than the first enabled one (the component and its doc comment); the switcher passes a handle, a section switch announces the section and claims the selected segment when the old focus is gone (P-11). The leave-or-stay decision after a Remove moves into a function on `UnpromoteFlow` or a static helper that returns leave, or stay with the announcement, and the Sources section posts it; both branches are unit-tested (P-12, P-16). The wording of rule "A person's records are records": the header's count reads records on every Conclusion page, and the Unpromote confirm ends "…, and the person stays", per kind (P-17).
**Not this PR.** Fold focus (S16-07); Back restoring scroll, folds and focus (P-19, [idea](../../ideas/person-page-followups.md)).
**Check.** Bar 7, and bar 8's count and confirm; unit tests for both leave-or-stay branches, replacing `UnpromoteFlowTests.confirmReportsTheEffects`' inline copy; manual VoiceOver steps for ⌘3 from a Family tile.
**Context.** Audit P-11, P-12, P-16, P-17; rules "Every outcome on the page is spoken", "A person's records are records", "One open section, one list of reads"; `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `PersonSourcesSection.swift`, `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift`, `macos/ProvenenciaTests/UnpromoteFlowTests.swift`; [`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md).
