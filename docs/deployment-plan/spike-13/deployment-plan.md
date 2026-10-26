# Deployment plan — Spike 13

The Person page becomes the place a researcher goes to understand one canonical person: a fixed header over five sections (Vitals, Family, Timeline, Gallery, Sources), a reusable provenance sheet beside every value in Vitals, and controls to remove or review a person's identity claims. The requirements are [`person-detail.md`](person-detail.md) (R-1 … R-40, R-D1 … R-D3); this map links to them instead of restating them. It pulls in the person-page half of [`ideas/kinship-view.md`](../../ideas/kinship-view.md) (siblings from a shared parent). One behavior, one PR.

**Design first, then re-cut** ([R-D1, R-D2](person-detail.md#design-first-then-re-cut)). Unlike earlier spikes, every brief is written up front and the boards are approved **together**, as one page. Until then every chrome PR is `Scoped` and its scope entry is provisional; the map is re-cut when the boards are agreed. Engine PRs whose shape the boards cannot change run alongside the design ([R-D3](person-detail.md#design-first-then-re-cut)).

**Spike 11 runs in parallel; Spike 12 is expected to be closed.** PRs that touch code Spike 11 is still changing name the Spike 11 PR in **Depends on**. The overlaps are in [person-detail.md § Cross-spike dependencies](person-detail.md#cross-spike-dependencies). A PR plan written before its Spike 11 dependency lands stops at `Blocked — on S11-NN`.

## Status

**Open — design phase.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** from any Person page a researcher can read the person's facts with their provenance, see and walk their nuclear family and their life's events, browse their likenesses, and see, review and remove the claims that make them up.

## Goal (dogfood bar)

**By spike close**, each of these is true on the product owner's own test people:

1. **One page, five sections.** The header (name, born/died line, portrait, gender badge) stays put while the user moves between Vitals, Family, Timeline, Gallery and Sources in one action from anywhere on the page.
2. **Every Vitals value explains itself.** Clicking the provenance control beside the birth date opens a sheet listing every record of that date (from the birth event), the auto-reconciler's reasoning, and links that open the Source, the Citation and the exact Observation. A disagreeing value still shows as mixed in the row.
3. **The family is there.** A person with two marriages shows both families, each spouse with their children; their parents; and their siblings, including siblings known only through a shared parent. Every tile shows a likeness, name, birth and death date and place, in the person's gender/sex color, and opens that relative's page. Nobody unpromoted appears.
4. **A life in order.** The Timeline lists every event the person took part in, earliest first, each with its name, date, place, the person's role, a depiction placeholder, and the other canonical participants with their roles; undated events sit in an Undated group at the end. Every event and participant opens its page.
5. **Likenesses.** The Gallery shows every likeness; one opens a viewer that steps through them; from the gallery or viewer a likeness can be set as the portrait or opened at its Citation.
6. **Sources.** The Sources section lists every identity claim, grouped under its Source (title, type, date) and ordered by Source title, with each claim's Subject card; a census contributing two claims from two graphs shows one group. Each Source opens its page.
7. **Claims.** A claim's review shows its pinned evidence and argument, read only. Removing a claim un-promotes the Subject; removing the last one warns that the person will be deleted, and proceeding lands on the Persons list.
8. **Matching is unchanged.** Promote proposals and the evidence sheet are the same as before the spike: a shared parent is not sibling evidence ([R-18a](person-detail.md#2-family)).

## Where the notes went

| Requirement, idea or note | PR | Brief |
| --- | --- | --- |
| R-1 – R-7 layout, fixed header, section switching | S13-09, S13-10 | S13-D1 |
| R-8 – R-15 provenance sheet, Vitals | S13-06, S13-11 | S13-D2 |
| R-16 – R-22 Family; [`ideas/kinship-view.md`](../../ideas/kinship-view.md) "Family trees and person pages", siblings from shared parents | S13-02, S13-03, S13-04, S13-12 | S13-D3 |
| R-23 – R-28 Timeline | S13-05, S13-13 | S13-D4 |
| R-29 – R-32 Gallery and viewer | S13-14, S13-15 | S13-D5 |
| R-33 – R-37 Sources | S13-01, S13-16 | S13-D6 |
| R-38, R-39 remove a claim, last claim deletes the person | S13-08, S13-17 | S13-D6 |
| R-40 claim review | S13-07, S13-18 | S13-D7 |

**Back in the ideas folder** — [`ideas/kinship-view.md`](../../ideas/kinship-view.md) keeps kinship beyond the nuclear family, conflicts between shapes, a project's own patterns, citing an implied fact on a tree, and sibling evidence for matching (R-18a). [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md) keeps weak-claim alerts, the review queue, re-pinning and claim editing. The rest of [person-detail.md § Not this spike](person-detail.md#not-this-spike).

## Working rules

These are the calls for this spike. PR plans treat them as binding. The product rules are the requirements in [`person-detail.md`](person-detail.md); the rules below are the architecture and process calls that sit on top of them.

### Process

- **Seven briefs, one approval.** S13-D1 … S13-D7 are written together at the start of the spike, drawn as one board set, and approved together (R-D1). No chrome PR is planned before that approval. On approval the map is re-cut: scope entries for S13-09 onward are rewritten, PRs may split, merge or be descoped, and engine PRs already landed are amended by follow-ups, not reopened (R-D2).
- **Engine PRs that run ahead are the ones the boards cannot reshape.** S13-01 … S13-08 return data every plausible design needs (R-D3). Each read returns **data, not layout**: grouping into families or Sources happens in Go, placement on the page happens in Swift. If the boards ask for data a landed read lacks, a follow-up PR adds it.
- **Each PR updates the docs and skills it makes false.** S13-03: [`matching-model.md`](../../matching-model.md) only if a knob is added, and [`ideas/kinship-view.md`](../../ideas/kinship-view.md) (siblings shipped for the page). S13-08: [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5 (Unpromote of the last member). S13-09: [`macos-client-patterns.md`](../../macos-client-patterns.md) if the detail-page pattern changes.

### Engine

- **Each section reads its own query.** The Person page is several reads, each with its own RPC and `CatalogQueryKey`, so a section loads and refreshes on its own: today's `GetConclusionDetail` (header, Vitals, Gallery) plus one each for claims (S13-01), family (S13-04) and timeline (S13-05). `GetConclusionDetail` does not grow the other sections ([`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md)). Through [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md). Binds S13-01, S13-04, S13-05.
- **One person summary for every tile.** Family tiles and Timeline participants carry the same summary: entity, name, birth and death life facts, likeness path, and the gender and sex terms `PersonGenderTreatment` reads. It is built by `conclusionheaders.PersonsByIDs`, the code that builds the Person header, never a second copy (S13-02). Binds S13-02, S13-04, S13-05.
- **Canonical only, filtered in Go** (R-2). Reads return canonical entities with at least one accepted member; unpromoted Subjects never cross the wire. Binds S13-01, S13-04, S13-05.
- **Siblings are page-only** (R-18a). The shared-parent sibling derivation lives beside the existing patterns in `core/kinship` but is **not** in the pattern set `canonlinks.Reader.For` gives matching; `promotealign` and S11-39's goldens see no change, and a test proves it. The Family read asks for siblings explicitly. Binds S13-03, S13-04.
- **Family reads direct and implied links as one fact.** Parents, spouses and children come from direct Relationship links (`graphcache.Node.Links` with `vocabkey.BridgeRelationship`, direction through `kinship.ParentStep`) merged with `canonlinks.Reader.For`'s implied links, keyed so a relationship recorded both ways appears once. Families are grouped in Go: one family per spouse (with the children who have both as parents), one per distinct parent set, and children with one known parent in their own group. Binds S13-04.
- **Timeline is direct participation, any role** (R-23). The read walks the person's event links in every role, not `hops.EventsOfSubject`'s `subject` role only, and returns each event's title, date, place, the person's role term and the other canonical participants with theirs. Sorting by date happens in Go; undated events come back flagged, grouped by Swift (R-24). Binds S13-05.
- **The provenance sheet reads the records the page already has.** `ConclusionDetail` already carries `ReconcilerOutcomeDetail` records per field; S13-06 adds them for the derived life rows (birth and death date and place, read from the lead birth or death event's fields) so every Vitals row has records. No new provenance RPC. Binds S13-06, S13-11.
- **Claims are read per person, grouped by Source in Go.** S13-01 reads accepted members through `identityclaims.AcceptedMembers`, joins each member's Evidence graph (`evidencegraphs`) and Source (`sources`, `sourcetypes`), and returns groups ordered by Source title (R-36) with the claims inside. The Source **date** is the provisional open question below. Binds S13-01.
- **Claim review reuses the claim evidence types.** S13-07's stored-claim read returns pins (`identityclaims.PinnedObservations`), argument and confidence in `claimevidence`'s `Line` / `Section` shapes, so the review sheet can reuse the Promote evidence sheet's line rendering. When S11-40 adds implied-link relationship lines, the review picks them up through the same types. Binds S13-07, S13-18.
- **One Unpromote.** The Person page's Remove calls the same `UnpromoteSubject` as the Evidence graph (R-38). S13-08 changes what that one path does on the last accepted member (open question below); there is no Person-page-only removal. Binds S13-08, S13-17.

### Client

- **The Person page leaves the shared detail page.** Event and Place keep `ConclusionDetailPage` unchanged (R-1). S13-09 moves the Person page onto its own page type that reuses the header, rows and likeness section as components, with no visible change, before any new chrome. Binds S13-09, S13-10.
- **One person tile, built once.** The tile is built in S13-12 (Family) and reused by S13-13 (Timeline), placed per [`add-ui-component`](../../../.agents/skills/add-ui-component/SKILL.md) as a recipe once it has the second caller. Briefs S13-D3 and S13-D4 draw the same tile. Binds S13-12, S13-13.
- **Colors reuse `PersonGenderTreatment`** (R-3), fed from the person summary. Binds S13-12, S13-13.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S13-01 | Person claims read | 1 | `ListPersonClaims` RPC: a person's accepted claims grouped by Source (title, type, date), ordered by title, each with its Subject, graph and claim id | — | — | — | Scoped |
| S13-02 | Person summary | 1 | One person summary (name, life facts, likeness, gender and sex terms) from `conclusionheaders.PersonsByIDs`, on the wire for tiles | — | S11-30 | — | Scoped |
| S13-03 | Siblings from a shared parent | 1 | A page-only sibling derivation in `core/kinship`, outside matching's pattern set, with a test that matching is unchanged | — | S11-39 | — | Scoped |
| S13-04 | Person family read | 1 | `GetPersonFamily` RPC: parents, siblings, spouses, children from direct and implied links, grouped into families, as person summaries | — | S13-02, S13-03 | — | Scoped |
| S13-05 | Person timeline read | 1 | `ListPersonEvents` RPC: every event the person took part in, any role, sorted by date, with role, place and the other participants as person summaries | — | S13-02 | — | Scoped |
| S13-06 | Life rows carry their records | 1 | `ConclusionDetail`'s birth and death date and place rows carry the lead event's reconciler records | — | — | — | Scoped |
| S13-07 | Stored claim read | 1 | `GetIdentityClaim` RPC: an accepted claim's pins, argument and confidence in `claimevidence` shapes | — | — | — | Scoped |
| S13-08 | Unpromoting the last member | 1 | `UnpromoteSubject` on a person's last accepted member (open question) with an impact flag the confirm can word | — | — | — | Scoped |
| S13-09 | Person page on its own page | 2 | Preparatory: the Person page moves off `ConclusionDetailPage` onto its own page type; no visible change | — | — | — | Scoped |
| S13-10 | Header and sections | 2 | The fixed header, the section switcher, and Vitals as the first section (today's rows) | S13-D1 | S13-09, design gate | — | Scoped |
| S13-11 | Provenance sheet | 2 | The provenance control and sheet replace the Why accordion on every Vitals row, with links to Source, Citation and Observation | S13-D2 | S13-06, S13-10, design gate | — | Scoped |
| S13-12 | Family section | 3 | The family diagram with person tiles, lines and several families | S13-D3 | S13-04, S13-10, design gate | — | Scoped |
| S13-13 | Timeline section | 3 | The chronological event list with roles, depiction placeholder, participants as person tiles, Undated group | S13-D4 | S13-05, S13-12, design gate | — | Scoped |
| S13-14 | Gallery section | 4 | The gallery as a section, with set as portrait and open Citation | S13-D5 | S13-10, design gate | — | Scoped |
| S13-15 | Likeness viewer | 4 | Expand a likeness and step through the gallery | S13-D5 | S13-14 | — | Scoped |
| S13-16 | Sources section | 5 | Claims grouped by Source with Subject cards, each Source opening its page | S13-D6 | S13-01, S13-10, design gate | — | Scoped |
| S13-17 | Remove a claim | 5 | Remove on a claim, the last-claim warning, and landing on the Persons list | S13-D6 | S13-08, S13-16 | — | Scoped |
| S13-18 | Claim review | 5 | A read-only sheet with a claim's pinned evidence and argument | S13-D7 | S13-07, S13-16, design gate | — | Scoped |
| S13-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

All seven briefs are written up front and approved together (working rule "Seven briefs, one approval"). The **design gate** is that approval; it is a step in the map, not a PR.

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S13-D1 | Person page: header, section switching, Vitals | **S13-10** | Explores stacked-with-shortcuts versus tabs (R-7). Every other board sits inside this frame |
| S13-D2 | Provenance sheet | **S13-11** | Reusable beyond Person (R-13); leaves room for Reconciliation Claims (R-10) |
| S13-D3 | Family section | **S13-12** | Draws the shared person tile; two marriages and conflicting parents |
| S13-D4 | Timeline section | **S13-13** | Reuses S13-D3's tile; depiction placeholder; Undated group |
| S13-D5 | Gallery section and likeness viewer | **S13-14**, **S13-15** | Viewer options: expand, slideshow (R-30) |
| S13-D6 | Sources section | **S13-16**, **S13-17** | Source groups, claim cards, Remove and the last-claim warning |
| S13-D7 | Claim review sheet | **S13-18** | Read only; reuses the Promote evidence sheet's lines (S11-D2) |

## Slices

**✎ = design brief.** **Check** = what you can verify in the app when the slice lands.

```text
SLICE 0 — Design (runs alongside slice 1)
  ✎ S13-D1 … S13-D7, drawn and approved together
  Check: the product owner approves the whole page; the map is re-cut

SLICE 1 — Engine reads (runs alongside the design)
  S13-01  Person claims read
  S13-02  Person summary
  S13-03  Siblings from a shared parent
  S13-04  Person family read
  S13-05  Person timeline read
  S13-06  Life rows carry their records
  S13-07  Stored claim read
  S13-08  Unpromoting the last member
  Check: go tests; matching goldens unchanged; no visible change in the app
         except S13-08's Evidence graph unpromote behavior

SLICE 2 — Page frame and Vitals
  S13-09  Person page on its own page
  ✎ S13-D1 ──▶ S13-10  Header and sections
  ✎ S13-D2 ──▶ S13-11  Provenance sheet
  Check: bar 1 and 2

SLICE 3 — Family and Timeline
  ✎ S13-D3 ──▶ S13-12  Family section
  ✎ S13-D4 ──▶ S13-13  Timeline section
  Check: bar 3, 4 and 8

SLICE 4 — Gallery
  ✎ S13-D5 ──▶ S13-14  Gallery section
               S13-15  Likeness viewer
  Check: bar 5

SLICE 5 — Sources and claims
  ✎ S13-D6 ──▶ S13-16  Sources section
               S13-17  Remove a claim
  ✎ S13-D7 ──▶ S13-18  Claim review
  Check: bar 6 and 7

S13-99 — Dogfood close / docs
```

**In parallel:** slice 1's PRs are independent of each other except S13-04 (after S13-02 and S13-03) and S13-05 (after S13-02); S13-01, S13-06, S13-07 and S13-08 can start at once. S13-09 can land during the design. After the design gate, slices 3, 4 and 5 can run in parallel once S13-10 lands.

### Dependencies at a glance

```text
S11-30 ──▶ S13-02 ──┬──▶ S13-04 ──▶ S13-12 ──▶ S13-13
S11-39 ──▶ S13-03 ──┘                ▲           ▲
           S13-02 ──▶ S13-05 ────────┼───────────┘
S13-09 ──▶ S13-10 ──┬──▶ S13-11 ◀── S13-06
                    ├──▶ S13-12
                    ├──▶ S13-14 ──▶ S13-15
                    └──▶ S13-16 ──┬──▶ S13-17 ◀── S13-08
                                  └──▶ S13-18 ◀── S13-07
S13-01 ──▶ S13-16
design gate (S13-D1 … S13-D7) ──▶ S13-10 … S13-18
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Chrome scopes (slice 2 onward) are provisional until the design gate.

### Slice 1 — Engine reads

#### S13-01 — Person claims read

**Ships.** `ListPersonClaims(personId)`: the person's accepted identity claims grouped by Source, groups ordered by Source title, each group with the Source's id, ref, title, type name and date, and each claim with its claim id, Subject (ref and card fields), Evidence graph id and name. Its own query key, refreshed by the effects that promote or unpromote into the person.
**Not this PR.** The section (S13-16); the date if the open question is not settled (ships without it, added after).
**Check.** A handler test: a Source with two graphs each contributing a claim returns one group with two claims.
**Context.** Working rules "Each section reads its own query", "Claims are read per person"; R-33 – R-37; `identityclaims.AcceptedMembers`, `evidencegraphs.List`.

#### S13-02 — Person summary

**Ships.** A `PersonSummary` on the wire (entity, name, birth and death `LifeFacts`, likeness path, gender and sex terms) built by `conclusionheaders.PersonsByIDs`; `PersonHeader` reuses it or embeds it so there is one shape. Swift's `PersonGenderTreatment` can read from it without full fields.
**Not this PR.** Any tile chrome (S13-12).
**Check.** `PersonsByIDs` tests cover the gender and sex terms; Swift treatment tests read a summary.
**Context.** Working rule "One person summary for every tile"; R-3, R-21; S11-30's gender ↔ sex crosswalk must land first so the terms are final.

#### S13-03 — Siblings from a shared parent

**Ships.** A sibling derivation in `core/kinship`: two canonical people who share a parent (direct or implied) are siblings, with the parent and links behind it. It is reachable only from an explicit option, not from the pattern set `canonlinks.Reader.For` gives matching.
**Not this PR.** Half-sibling and step relationships as distinct kinds; sibling evidence in matching (R-18a).
**Check.** `go test ./core/kinship/...`; a test that `promotealign`'s inputs are unchanged with the derivation present; S11-39's goldens unchanged.
**Context.** Working rule "Siblings are page-only"; R-18, R-18a; [`ideas/kinship-view.md`](../../ideas/kinship-view.md); S11-39 edits the same package and goes first.

#### S13-04 — Person family read

**Ships.** `GetPersonFamily(personId)`: parent sets, spouses with their shared children, children with one known parent, and siblings, each as a person summary and each with the relationship kind and whether it was recorded directly or implied. Canonical people only.
**Not this PR.** The diagram (S13-12); grandparents and beyond.
**Check.** Handler tests: two marriages give two families; two conflicting parent sets give two parent groups; a relationship recorded directly and implied by a marriage appears once.
**Context.** Working rules "Family reads direct and implied links as one fact", "Canonical only"; R-17 – R-19.

#### S13-05 — Person timeline read

**Ships.** `ListPersonEvents(personId)`: every canonical event the person took part in directly, in any role, sorted by date, each with title, date, place, the person's role, and the other canonical participants with their roles as person summaries; undated events flagged.
**Not this PR.** The section (S13-13); event depictions (R-26).
**Check.** Handler tests: a witness role appears; a child's birth the person is not linked to does not; undated events come last and flagged.
**Context.** Working rule "Timeline is direct participation, any role"; R-23 – R-28; `graphcache.Link.TermKey`, `conclusionheaders` event titles.

#### S13-06 — Life rows carry their records

**Ships.** `ConclusionDetail` gives the birth and death date and place rows the lead birth or death event's `ReconcilerOutcomeDetail` records, so `ReconciledValueRowModel.derived` rows have records like other fields.
**Not this PR.** The sheet (S13-11).
**Check.** A handler test: a person with two birth records on one birth event returns both on the birth-date row.
**Context.** Working rule "The provenance sheet reads the records the page already has"; R-13a, R-14.

#### S13-07 — Stored claim read

**Ships.** `GetIdentityClaim(claimId)`: the claim's status, confidence, argument and pinned Observations, in `claimevidence` `Section` / `Line` shapes.
**Not this PR.** Editing pins or argument (R-40); the sheet (S13-18).
**Check.** A handler test round-trips a claim promoted with pins and an argument.
**Context.** Working rule "Claim review reuses the claim evidence types"; `identityclaims.PinnedObservations`; S11-19, S11-25.

#### S13-08 — Unpromoting the last member

**Ships.** `UnpromoteSubject` on a canonical entity's last accepted member does what the open question below decides, and its preflight tells the confirm that this is the last member so the copy can warn.
**Not this PR.** The Person page control (S13-17).
**Check.** Promote tests for the last member; the Evidence graph's unpromote confirm words the new outcome.
**Context.** Working rule "One Unpromote"; R-38, R-39; `promote.Unpromote`, `deleteimpact.ReleaseClaim`. **Blocked on a product answer** before `plan-pr`.

### Slice 2 — Page frame and Vitals

#### S13-09 — Person page on its own page

**Ships.** A preparatory refactor: `PersonDetailView` stops configuring `ConclusionDetailPage` and gets its own page type composed of the header, rows and likeness section, now shared components. Event and Place are untouched.
**Not this PR.** Any visible change.
**Check.** The Person, Event and Place pages look the same; existing Person page tests pass.
**Context.** Working rule "The Person page leaves the shared detail page"; R-1; `ConclusionDetailPage.swift`.

#### S13-10 — Header and sections

**Ships.** The fixed, display-only header and the section switcher from S13-D1, with Vitals (today's rows) as the first section and the other four present as the board draws them empty.
**Check.** Bar 1.
**Context.** R-4 – R-7, R-14, R-15; brief S13-D1.

#### S13-11 — Provenance sheet

**Ships.** The provenance control beside every Vitals value and the sheet from S13-D2: every record, the reconciler's reasoning, and links to the Source, Citation and Observation. It replaces the Why accordion on the Person page.
**Not this PR.** Reconciliation Claims (R-10); the control outside Vitals (R-13b).
**Check.** Bar 2.
**Context.** R-8 – R-13b; working rule "The provenance sheet reads the records the page already has"; `ReconciliationReasoningView`.

### Slice 3 — Family and Timeline

#### S13-12 — Family section

**Ships.** The Family section from S13-D3: person tiles, lines, several families, each tile opening its person.
**Check.** Bar 3.
**Context.** R-16 – R-22; working rules "One person tile, built once", "Colors reuse `PersonGenderTreatment`".

#### S13-13 — Timeline section

**Ships.** The Timeline section from S13-D4: events in order, roles, depiction placeholder, participants as the S13-12 tile, Undated group.
**Check.** Bar 4.
**Context.** R-23 – R-28.

### Slice 4 — Gallery

#### S13-14 — Gallery section

**Ships.** The gallery as a section from S13-D5, keeping set as portrait and opening the Citation.
**Check.** Bar 5 without the viewer.
**Context.** R-29, R-31, R-32; `ConclusionPortraitModel`, `SetPreferredLikeness`.

#### S13-15 — Likeness viewer

**Ships.** The viewer from S13-D5: expand a likeness and step through the gallery, with set as portrait and open Citation.
**Check.** Bar 5.
**Context.** R-30.

### Slice 5 — Sources and claims

#### S13-16 — Sources section

**Ships.** The Sources section from S13-D6, read from S13-01: Source groups with Subject cards, each Source opening its page.
**Check.** Bar 6.
**Context.** R-33 – R-37.

#### S13-17 — Remove a claim

**Ships.** Remove on each claim through `UnpromoteSubject`, the last-claim warning, and navigation to the Persons list when the person is gone.
**Check.** Bar 7, removal half.
**Context.** R-38, R-39; working rule "One Unpromote"; [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md) for leaving a page whose entity is gone.

#### S13-18 — Claim review

**Ships.** The read-only claim review sheet from S13-D7, read from S13-07.
**Not this PR.** Editing (R-40).
**Check.** Bar 7, review half.
**Context.** R-40; working rule "Claim review reuses the claim evidence types"; `PromoteEvidenceSheet`.

## Open questions

Not decided; a PR that depends on one stays `Scoped` and `plan-pr` asks it again.

1. **Last member, everywhere or Person page only?** Today unpromoting the last member keeps the canonical person with no members ("the handle stays"). R-39 says removing the last claim deletes the person. With one Unpromote (working rule), that would change the Evidence graph's behavior too. Gates S13-08.
2. **Which Source date?** Sources have no date column; dates live in typed metadata fields (`record_date`, `issue_date`, …) as text. Which field, and what when a Source has none? Gates the date in S13-01.
3. **Provenance sheet on Event and Place pages.** The Why accordion is a shared row. Should S13-11 replace it on Event and Place too (one mechanism), or keep the accordion there until those pages are redesigned (R-1)? Gates S13-11.
4. **Remembering the section.** Should Back / Forward and restore return to the section the user was on (the section is part of the `WorkspaceLocation`), or always open on Vitals? Gates S13-10.
