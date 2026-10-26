# Deployment plan — Spike 13

The Person page becomes the place a researcher goes to understand one canonical person: a fixed header over five sections (Vitals, Family, Timeline, Gallery, Sources), a reusable provenance sheet beside every value in Vitals, and controls to remove or review a person's identity claims. The requirements are [`person-detail.md`](person-detail.md) (R-1 … R-40, R-D1 … R-D3); this map links to them instead of restating them. It pulls in the person-page half of [`ideas/kinship-view.md`](../../ideas/kinship-view.md) (siblings from a shared parent). One behavior, one PR.

**Design first, then re-cut** ([R-D1, R-D2](person-detail.md#design-first-then-re-cut)). Unlike earlier spikes, every brief is written up front and the boards are approved **together**, as one page. Until then every chrome PR is `Scoped` and its scope entry is provisional; the map is re-cut when the boards are agreed. Engine PRs whose shape the boards cannot change run alongside the design ([R-D3](person-detail.md#design-first-then-re-cut)).

Map review: [`deployment-plan.review.md`](deployment-plan.review.md).

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
| R-16 – R-22 Family; [`ideas/kinship-view.md`](../../ideas/kinship-view.md) "Family trees and person pages", siblings from shared parents | S13-02, S13-03, S13-19, S13-04, S13-12 | S13-D3 |
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
- **Engine PRs that run ahead are the ones the boards cannot reshape.** S13-01 … S13-08 and S13-19 return data every plausible design needs (R-D3). Each read returns **data, not layout**: grouping into families or Sources happens in Go, placement on the page happens in Swift. If the boards ask for data a landed read lacks, a follow-up PR adds it.
- **Each PR updates the docs and skills it makes false.** S13-19: [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) (one link merge). S13-03: `add-match-rule` (page-only derivations), [`matching-model.md`](../../matching-model.md) only if a knob is added, and [`ideas/kinship-view.md`](../../ideas/kinship-view.md) (siblings shipped for the page). S13-08: [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5 (Unpromote of the last member), [`catalog-deletes.md`](../../catalog-deletes.md). S13-01: [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md) (per-handle invalidation spec). S13-06: the `LifeFacts` and `ReconciledValueRowModel.derived` comments ("its page owns the Why"). S13-09: [`macos-client-patterns.md`](../../macos-client-patterns.md) (sectioned detail page).

### Engine

- **Each section reads its own query.** The Person page is several reads, each with its own RPC and `CatalogQueryKey`, so a section loads and refreshes on its own: today's `GetConclusionDetail` (header, Vitals, Gallery) plus one each for claims (S13-01), family (S13-04) and timeline (S13-05). `GetConclusionDetail` does not grow the other sections ([`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md)). The three new reads are composed in one package, `core/database/persondetail` (beside `conclusiondetails`); handlers stay thin ([`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md)). Binds S13-01, S13-04, S13-05.
- **One invalidation rule for person-section keys.** Today `CatalogQueryRegistry.invalidations(for:project:)` hand-writes the `touchedHandles` → `.conclusionDetail` fan-out. S13-01 turns that into a registry `Spec` declaration (per-handle keys refreshed for every touched handle) and declares its key with it; S13-04 and S13-05 add one entry each. A relative's or co-participant's change touches their id, not this person's, so the family and timeline keys also refresh on any promote or unpromote effect (`.allCached` for that kind) unless the plan finds a keyed effect. Binds S13-01, S13-04, S13-05.
- **One person header for every tile.** Family tiles and Timeline participants carry `PersonHeader` (`conclusionheaders.PersonsByIDs`), which already holds name, birth and death `LifeFacts` and likeness, and whose package doc already names "later tree nodes". S13-02 adds the gender and sex terms to it; no new summary type. `PersonsByIDs` returns rows in title order, so S13-04 and S13-05 re-key by id. Binds S13-02, S13-04, S13-05.
- **Canonical only, filtered in Go** (R-2). Reads return canonical entities with at least one accepted member; unpromoted Subjects never cross the wire. One `graphcache` `Node` accepted-members helper (S13-01) replaces the private copies `promotealign.acceptedClaims` and `conclusiondetails.acceptedMembers`, and every reader uses it. Binds S13-01, S13-04, S13-05.
- **Siblings are page-only** (R-18a). The shared-parent sibling derivation lives in `core/kinship` but is **not** in the pattern set `canonlinks.Reader.For` gives matching; `promotealign` and S11-39's goldens see no change, and a test proves it. The Family read asks for siblings explicitly. S13-03 updates [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) (which today says a sibling pattern "changes `Derive`") with where page-only derivations go. Binds S13-03, S13-04.
- **One person's links, direct and implied, from one function.** `promotealign` merges direct and implied links through its private `edgeSeenKey` in two places (`canon.go` `loadCanon`, `layer.go`), and S11-39 adds more. S13-19 (preparatory) gives `canonlinks` one export that returns a person's links, direct (through `kinship.ParentStep`) and implied, deduped by an exported merge key, moves `promotealign` onto it, and updates `add-match-rule`'s "callers merge … through their own merge" line. The Family read calls it. Binds S13-19, S13-04.
- **Family groups in Go.** One family per spouse (with the children who have both as parents), one per distinct parent set, and children with one known parent in their own group. Binds S13-04.
- **Timeline is direct participation, any role** (R-23). S13-05 adds an unfiltered participation hop to `core/hops` (beside `EventsOfSubject`, which is `subject` role only) and returns, per event, `EventHeader` (title, date, folded places), the person's role term, and the other participants as {role, `PersonHeader`}, ordered by `conclusionheaders.sortEvents` (undated last). Swift titles events with `EventTitleDisplay` and groups the undated ones (R-24). Binds S13-05.
- **A derived row carries its owning entity's records.** `ConclusionDetail` already carries `ReconcilerOutcomeDetail` records per field. S13-06 adds, for each derived row, the records of the field on the entity that owns it (the lead birth or death event's date; its Place), keyed by that entity on `ConclusionDetail`, not on `LifeFacts` (which lists and tiles also carry). Event's derived place row can take the same path later. No new provenance RPC. Binds S13-06, S13-11.
- **Claims are read per person from the graph cache, grouped by Source in Go.** S13-01 reads the person's accepted members from `graphcache` (`Member` has Subject, Source and graph ids; `EvidenceGraph` has the Source title), adds Source type through `sourcetypes`, and returns groups ordered by Source title (R-36). Each claim carries the existing `CatalogSubject` and Observation wire messages the Evidence graph's Subject card already renders. **No Source date** until the open question is settled. Binds S13-01.
- **Claim review reuses the claim evidence query.** S13-07 adds a stored-claim mode to `promotealign.Evidence`: the claim's own Subject is left out of the members, lines are checked from the stored pins, and status, grade and argument come from `identityclaims`. `GetIdentityClaim` calls it; there is no second evidence builder. S11-40 also changes `Evidence`; whichever lands second rebases. Binds S13-07, S13-18.
- **One Unpromote, deletes through the registry.** The Person page's Remove calls the same `UnpromoteSubject` as the Evidence graph (R-38). If the last member's removal deletes the person (open question 1), S13-08 deletes the canonical entity through `core/database/deleteimpact` ([`add-catalog-delete`](../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-audit-scope`](../../../.agents/skills/add-audit-scope/SKILL.md)), and the confirm reads an Impact report through an unpromote preflight, as `GetDeleteImpact` and `GetEvidenceGraphUnlinkImpact` do, not a one-off flag. Binds S13-08, S13-17.

### Client

- **One sectioned Conclusion page, not a Person fork.** `ConclusionDetailPage` stays the one page for Person, Event and Place. S13-09 gives it a sections slot; Event and Place pass a single section and look unchanged (R-1), and the Person page later passes five. When Event and Place adopt sections, they fill the slot. Binds S13-09, S13-10.
- **One person tile, built once.** The tile is built in S13-12 (Family) and reused by S13-13 (Timeline), placed per [`add-ui-component`](../../../.agents/skills/add-ui-component/SKILL.md) as a recipe once it has the second caller. Briefs S13-D3 and S13-D4 draw the same tile. Binds S13-12, S13-13.
- **Colors reuse `PersonGenderTreatment`** (R-3), fed from `PersonHeader`'s terms (S13-02). Binds S13-12, S13-13.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S13-01 | Person claims read | 1 | `ListPersonClaims` RPC from the graph cache: a person's accepted claims grouped by Source (title, type), ordered by title, each with its Subject card data, graph and claim id; one accepted-members helper; per-handle invalidation spec | — | — | — | Scoped |
| S13-02 | Gender in the person header | 1 | `PersonHeader` gains the gender and sex terms; `PersonGenderTreatment` reads only the header | — | S11-30 | — | Scoped |
| S13-03 | Siblings from a shared parent | 1 | A page-only sibling derivation in `core/kinship`, outside matching's pattern set, with a test that matching is unchanged | — | S11-39 | — | Scoped |
| S13-04 | Person family read | 1 | `GetPersonFamily` RPC: parents, siblings, spouses, children from S13-19's links, grouped into families, as `PersonHeader`s | — | S13-01, S13-02, S13-03, S13-19 | — | Scoped |
| S13-05 | Person timeline read | 1 | An any-role participation hop; `ListPersonEvents` RPC: `EventHeader`s sorted by `sortEvents`, with the person's role and the other participants as {role, `PersonHeader`} | — | S13-01, S13-02 | — | Scoped |
| S13-06 | Derived rows carry their records | 1 | `ConclusionDetail` carries, for each derived row, the reconciler records of the owning entity's field (birth and death event date and place) | — | — | — | Scoped |
| S13-07 | Stored claim read | 1 | A stored-claim mode on `promotealign.Evidence`; `GetIdentityClaim` RPC returns its sections with pins checked, status, grade and argument | — | — | — | Scoped |
| S13-08 | Unpromoting the last member | 1 | `UnpromoteSubject` on the last accepted member deletes the person through `deleteimpact` (open question 1); an unpromote Impact preflight for the confirm | — | — | — | Blocked — on open question 1 |
| S13-19 | One person's links | 1 | Preparatory: `canonlinks` exports a person's direct and implied links deduped by one exported merge key; `promotealign` moves onto it | — | S11-39 | — | Scoped |
| S13-09 | Sections slot on the Conclusion page | 2 | Preparatory: `ConclusionDetailPage` gains a sections slot; Person, Event and Place pass one section; no visible change | — | — | — | Scoped |
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
  S13-02  Gender in the person header
  S13-03  Siblings from a shared parent
  S13-19  One person's links (preparatory)
  S13-04  Person family read
  S13-05  Person timeline read
  S13-06  Derived rows carry their records
  S13-07  Stored claim read
  S13-08  Unpromoting the last member
  Check: go tests; matching goldens unchanged; no visible change in the app
         except S13-08's Evidence graph unpromote behavior

SLICE 2 — Page frame and Vitals
  S13-09  Sections slot on the Conclusion page
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

**In parallel:** S13-01, S13-06 and S13-07 can start at once; S13-02 waits on S11-30, and S13-03 and S13-19 on S11-39; S13-04 and S13-05 follow S13-01 (its accepted-members helper and invalidation spec) and S13-02; S13-08 waits on open question 1. S13-09 can land during the design. After the design gate, slices 3, 4 and 5 can run in parallel once S13-10 lands.

### Dependencies at a glance

```text
S11-30 ──▶ S13-02 ──┬──────────────▶ S13-04 ──▶ S13-12 ──▶ S13-13
S11-39 ──▶ S13-03 ──┤                  ▲                     ▲
S11-39 ──▶ S13-19 ──┘                  │                     │
S13-01 ──┬─────────────────────────────┘                     │
         ├──▶ S13-05 ◀── S13-02 ─────────────────────────────┘
         └──▶ S13-16
S13-09 ──▶ S13-10 ──┬──▶ S13-11 ◀── S13-06
                    ├──▶ S13-12
                    ├──▶ S13-14 ──▶ S13-15
                    └──▶ S13-16 ──┬──▶ S13-17 ◀── S13-08 ◀── open question 1
                                  └──▶ S13-18 ◀── S13-07
design gate (S13-D1 … S13-D7) ──▶ S13-10 … S13-18
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Chrome scopes (slice 2 onward) are provisional until the design gate.

### Slice 1 — Engine reads

#### S13-01 — Person claims read

**Ships.** `ListPersonClaims(personId)` in `core/database/persondetail`, read from `graphcache`: the person's accepted identity claims grouped by Source, groups ordered by Source title, each group with the Source's id, ref, title and type name, and each claim with its claim id, Evidence graph id and name, and the `CatalogSubject` and Observations the Subject card renders. Also: one `graphcache` accepted-members helper with `promotealign.acceptedClaims` and `conclusiondetails.acceptedMembers` moved onto it, and the per-handle invalidation `Spec` with `.conclusionDetail` and this key declared through it.
**Not this PR.** The section (S13-16); a Source date (open question 2).
**Check.** A handler test: a Source with two graphs each contributing a claim returns one group with two claims; registry tests: a promote into the person refreshes both keys.
**Context.** Working rules "Each section reads its own query", "One invalidation rule for person-section keys", "Canonical only", "Claims are read per person from the graph cache"; R-33 – R-37.

#### S13-02 — Gender in the person header

**Ships.** `PersonHeader` (Go and proto) gains the gender and sex terms. `PersonGenderTreatment` reads only the header's terms; the Person page header moves onto it and the read from detail fields is deleted.
**Not this PR.** Any tile chrome (S13-12).
**Check.** `PersonsByIDs` tests cover the terms; Swift treatment tests read a header; the Person page badge and thumb colors are unchanged.
**Context.** Working rule "One person header for every tile"; R-3, R-21; S11-30's gender ↔ sex crosswalk lands first so the terms are final.

#### S13-03 — Siblings from a shared parent

**Ships.** A sibling derivation in `core/kinship`: two canonical people who share a parent (direct or implied) are siblings, with the parent and links behind it. It is reachable only from an explicit option, not from the pattern set `canonlinks.Reader.For` gives matching.
**Not this PR.** Half-sibling and step relationships as distinct kinds; sibling evidence in matching (R-18a).
**Check.** `go test ./core/kinship/...`; a test that `promotealign`'s inputs are unchanged with the derivation present; S11-39's goldens unchanged.
**Context.** Working rule "Siblings are page-only"; R-18, R-18a; [`ideas/kinship-view.md`](../../ideas/kinship-view.md); S11-39 edits the same package and goes first; updates `add-match-rule`.

#### S13-19 — One person's links

**Ships.** A preparatory refactor: `canonlinks` exports one function returning a person's links, direct Relationship links (direction through `kinship.ParentStep`) and implied ones, deduped by one exported merge key (today `promotealign`'s private `edgeSeenKey`). `promotealign`'s `loadCanon` and layer loader move onto the key; `add-match-rule`'s "own merge" line is rewritten.
**Not this PR.** Siblings (S13-03); any RPC.
**Check.** Matching goldens and `canonlinks` verify mode unchanged; a test that a relationship recorded directly and implied by a marriage comes back once.
**Context.** Working rule "One person's links, direct and implied, from one function"; S11-39 adds implied links to the same merge and goes first.

#### S13-04 — Person family read

**Ships.** `GetPersonFamily(personId)` in `core/database/persondetail`, from S13-19's links plus S13-03's siblings: parent sets, spouses with their shared children, children with one known parent, and siblings, each as a `PersonHeader` (re-keyed by id) with the relationship kind and whether it was recorded directly or implied. Canonical people only. Its key joins the S13-01 invalidation spec.
**Not this PR.** The diagram (S13-12); grandparents and beyond.
**Check.** Handler tests: two marriages give two families; two conflicting parent sets give two parent groups; a relationship recorded directly and implied by a marriage appears once.
**Context.** Working rules "One person's links, direct and implied, from one function", "Family groups in Go", "Canonical only"; R-17 – R-19.

#### S13-05 — Person timeline read

**Ships.** An any-role participation hop in `core/hops`, and `ListPersonEvents(personId)` in `core/database/persondetail`: every canonical event the person took part in directly, as `EventHeader`s ordered by `conclusionheaders.sortEvents`, each with the person's role term and the other canonical participants as {role, `PersonHeader`}. Its key joins the S13-01 invalidation spec.
**Not this PR.** The section (S13-13); event depictions (R-26).
**Check.** Handler tests: a witness role appears; a child's birth the person is not linked to does not; undated events come last and flagged.
**Context.** Working rule "Timeline is direct participation, any role"; R-23 – R-28; `hops.EventsOfSubject`, `EventHeader.subjects` (whether `EventSubject` folds into {role, `PersonHeader`} later is logged in the plan).

#### S13-06 — Derived rows carry their records

**Ships.** `ConclusionDetail` carries, keyed by owning entity, the `ReconcilerOutcomeDetail` records of each derived row's field (the lead birth or death event's date and Place), so `ReconciledValueRowModel.derived` rows have records like other fields. Not on `LifeFacts`. The "its page owns the Why" comments are updated.
**Not this PR.** The sheet (S13-11).
**Check.** A handler test: a person with two birth records on one birth event returns both on the birth-date row.
**Context.** Working rule "A derived row carries its owning entity's records"; R-13a, R-14.

#### S13-07 — Stored claim read

**Ships.** A stored-claim mode on `promotealign.Evidence` (the claim's own Subject left out of the members, lines checked from the stored pins) and `GetIdentityClaim(claimId)`, which returns those sections with the claim's status, grade and argument.
**Not this PR.** Editing pins or argument (R-40); the sheet (S13-18).
**Check.** A handler test round-trips a claim promoted with pins and an argument: the pinned lines come back checked.
**Context.** Working rule "Claim review reuses the claim evidence query"; `identityclaims.PinnedObservations`; S11-19, S11-25; S11-40 edits `Evidence` too.

#### S13-08 — Unpromoting the last member

**Ships.** `UnpromoteSubject` on a canonical entity's last accepted member deletes the entity through `core/database/deleteimpact` (Impact in the same transaction, probes, audit scope), and an unpromote Impact preflight the confirm words, as `GetDeleteImpact` does.
**Not this PR.** The Person page control (S13-17).
**Check.** Promote tests for the last member; the Evidence graph's unpromote confirm words the new outcome.
**Context.** Working rule "One Unpromote, deletes through the registry"; R-38, R-39; `promote.Unpromote`, `deleteimpact.ReleaseClaim`; `add-catalog-delete`, `add-audit-scope`. Blocked on open question 1.

### Slice 2 — Page frame and Vitals

#### S13-09 — Sections slot on the Conclusion page

**Ships.** A preparatory refactor: `ConclusionDetailPage` (and `ConclusionDetailBody`) gains a sections slot below the header; Person, Event and Place each pass today's content as one section.
**Not this PR.** Any visible change; the switcher (S13-10).
**Check.** The Person, Event and Place pages look the same; existing detail page tests pass.
**Context.** Working rule "One sectioned Conclusion page, not a Person fork"; R-1; `ConclusionDetailPage.swift`.

#### S13-10 — Header and sections

**Ships.** The fixed, display-only header and the section switcher from S13-D1, with Vitals (today's rows) as the first section and the other four present as the board draws them empty.
**Check.** Bar 1.
**Context.** R-4 – R-7, R-14, R-15; brief S13-D1.

#### S13-11 — Provenance sheet

**Ships.** The provenance control beside every Vitals value and the sheet from S13-D2: every record, the reconciler's reasoning, and links to the Source, Citation and Observation. It replaces the Why accordion on the Person page.
**Not this PR.** Reconciliation Claims (R-10); the control outside Vitals (R-13b).
**Check.** Bar 2.
**Context.** R-8 – R-13b; working rule "A derived row carries its owning entity's records"; `ReconciliationReasoningView`; open question 3.

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
2. **Which Source date?** Sources have no date column; dates live in typed metadata fields (`record_date`, `issue_date`, …) as text, and [`ideas/source-provenance-date.md`](../../ideas/source-provenance-date.md) already settles that a metadata key is not a first-class date. Options: drop the date from the group (R-35), or pull that idea into the spike as its own PR. S13-01 ships without a date either way.
3. **Provenance sheet on Event and Place pages.** The Why accordion is a shared row. Should S13-11 replace it on Event and Place too, or keep the accordion there until those pages are redesigned (R-1)? Recommendation: replace it everywhere and delete the accordion, so there is one provenance mechanism; keeping both is debt. Gates S13-11.
4. **Remembering the section.** Should Back / Forward and restore return to the section the user was on (the section is part of the `WorkspaceLocation`), or always open on Vitals? Gates S13-10.
