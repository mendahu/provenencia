# Deployment plan — Spike 13

The Person page becomes the place a researcher goes to understand one canonical person: a fixed header over five sections (Vitals, Family, Timeline, Gallery, Sources), a reusable provenance sheet beside every value in Vitals, and controls to remove or review a person's identity claims. The requirements are [`person-detail.md`](person-detail.md) (R-1 … R-40, R-D1 … R-D3); this map links to them instead of restating them. It pulls in the person-page half of [`ideas/kinship-view.md`](../../ideas/kinship-view.md) (siblings from a shared parent) and the first-class Source date, [`source-provenance-date.md`](source-provenance-date.md) (moved here from ideas, 2026-10-26), which the Sources section shows. One behavior, one PR.

**Slice 0 comes first** (user, 2026-10-29): it removes Identity Claim Observation pins, Promote's evidence sheet and the engine behind them, as decided in [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5 (docs PR #623). It is precursor work: the Person page's claims sections build on Identity Claims without pins.

**Design, then build, one view at a time** ([R-D1 – R-D3](person-detail.md#design-then-build-one-view-at-a-time)). Each slice draws its brief, agrees the board, then implements it before the next brief starts (user, 2026-10-26). Chrome scope entries are provisional until their board is agreed. Engine PRs may land ahead of their slice.

Map review: [`deployment-plan.review.md`](deployment-plan.review.md).

**Spikes 11 and 12 are closed; Spike 14 runs in parallel.** Spike 14's S14-16, S14-20 and S14-22 (moved from Spike 12) change Promote code that slice 0 removes and wait on it (working rule "Spike 14's Promote PRs build on the removal"). PRs that touch code Spike 11 is still changing name the Spike 11 PR in **Depends on**. The overlaps are in [person-detail.md § Cross-spike dependencies](person-detail.md#cross-spike-dependencies). A PR plan written before its Spike 11 dependency lands stops at `Blocked — on S11-NN`.

## Status

**Open.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** from any Person page a researcher can read the person's facts with their provenance, see and walk their nuclear family and their life's events, browse their likenesses, see and remove the claims that make them up, and see how the person's records connect.

## Goal (dogfood bar)

**By spike close**, each of these is true on the product owner's own test people:

1. **One page, five sections.** The header (name, born/died line, portrait, gender badge) stays put while the user moves between Vitals, Family, Timeline, Gallery and Sources in one action from anywhere on the page.
2. **Every Vitals value explains itself.** Clicking the provenance control beside the birth date opens a sheet listing every record of that date (from the birth event), the auto-reconciler's reasoning, and links that open the Source, the Citation and the exact Observation. A disagreeing value still shows as mixed in the row.
3. **The family is there.** A person with two marriages shows both families, each spouse with their children; their parents; and their siblings, including siblings known only through a shared parent. Every tile shows a likeness, name, birth and death date and place, in the person's gender/sex color, and opens that relative's page. Nobody unpromoted appears.
4. **A life in order.** The Timeline lists every event the person took part in, earliest first, each with its name, date, place, the person's role, a depiction placeholder, and the other canonical participants with their roles; undated events sit in an Undated group at the end. Every event and participant opens its page.
5. **Likenesses.** The Gallery shows every likeness; one opens a viewer that steps through them; from the gallery or viewer a likeness can be set as the portrait or opened at its Citation.
6. **Sources.** The Sources section lists every identity claim, grouped under its Source (title, type, date) and ordered by Source title, with each claim's Subject card; a census contributing two claims from two graphs shows one group. Each Source opens its page.
7. **Claims.** Removing a claim un-promotes the Subject; removing the last one warns that the person will be deleted, and proceeding lands on the Persons list. (A claim's review became bar 10 on 2026-10-29.)
8. **Matching is unchanged.** Promote proposals are the same as before the spike: a shared parent is not sibling evidence ([R-18a](person-detail.md#2-family)).
9. **Promote only files.** On a Source of 100 new people, Promote asks for a target per row and nothing else: no evidence sheet, no pins, no confidence or argument, and no strong, medium or weak band (the score still ranks the suggestions). Done files the claims and connections, an existing project opens with no `identity_claim_evidence` table, and Identity Claims carry no confidence grade or argument. Every filed row is Accepted. (Added 2026-10-29 for slice 0; the claim fields 2026-10-29; the per-row status descoped 2026-10-30 with S13-30.)
10. **How the records connect.** The Person page shows the person's Evidence graphs and the links between them: a link opens what the two records agree and conflict on, a record that links to no other stands apart, and two groups of records that do not connect show as two groups. (Added 2026-10-29 for slice 7; replaces the claim review.)

**Scale** (slice 0, 2026-10-29): Spike 12's — a Source of 100 new people on one Promote page, the 41-Subject obituary, a project whose every accepted claim has pins. Removal only takes work away: Done writes one claim per filed row and its bridges, and no evidence read runs before it. **Slice 7:** a person with 5 – 20 Evidence graphs is typical and 50 is the budget (1,225 pair comparisons per read).

## Where the notes went

| Requirement, idea or note | PR | Brief |
| --- | --- | --- |
| Drop Identity Claim Observation pins and the evidence sheet ([`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5, docs PR #623; user, 2026-10-29) | S13-23, S13-24, S13-25 | S13-D11 |
| Drop Identity Claim confidence and argument (user, 2026-10-29); status on the Promote row, Accepted or Provisional (user, 2026-10-29; descoped 2026-10-30) | S13-23 – S13-25 (S13-30 descoped) | S13-D11 |
| R-1 – R-7 layout, fixed header, section switching | S13-09, S13-10 | S13-D1 |
| R-8 – R-15 provenance sheet, Vitals | S13-06, S13-11 | S13-D12, S13-D2 |
| R-16 – R-22 Family; [`ideas/kinship-view.md`](../../ideas/kinship-view.md) "Family trees and person pages", siblings from shared parents | S13-02, S13-03, S11-51, S13-04, S13-12 | S13-D3 |
| R-23 – R-28 Timeline | S13-05, S13-13 | S13-D4 |
| R-29 – R-32 Gallery and viewer | S13-14, S13-15 | S13-D5 |
| R-33 – R-37 Sources | S13-01, S13-16 | S13-D6 |
| R-35 Source date; [`source-provenance-date.md`](source-provenance-date.md) (was `ideas/`) | S13-20, S13-21, S13-22 | S13-D8, S13-D9 |
| R-38, R-39 remove a claim, last claim deletes the person | S13-08, S13-17 | S13-D6 |
| R-40 claim review, replaced 2026-10-29 by how a person's records connect ([`evidence-graph-comparison.md`](evidence-graph-comparison.md)) | S13-26, S13-27, S13-28 (S13-07, S13-18 descoped) | S13-D10 (S13-D7 descoped) |

**Back in the ideas folder** — [`ideas/kinship-view.md`](../../ideas/kinship-view.md) keeps kinship beyond the nuclear family, conflicts between shapes, a project's own patterns, citing an implied fact on a tree, and sibling evidence for matching (R-18a). [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md) keeps weak-claim alerts, the review queue, re-pinning and claim editing. The rest of [person-detail.md § Not this spike](person-detail.md#not-this-spike).

## Working rules

These are the calls for this spike. PR plans treat them as binding. The product rules are the requirements in [`person-detail.md`](person-detail.md); the rules below are the architecture and process calls that sit on top of them.

### Process

- **Design, then build, one brief at a time.** A slice writes its brief just before its chrome PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), from [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)), agrees the board, and implements it; the next slice's brief starts after (user, 2026-10-26; supersedes "nine briefs, one approval"). Briefs for slices that do not touch each other's views (the Source date briefs S13-D8 and S13-D9) may run alongside.
- **Redesigns adopt Page columns** (user, 2026-10-29). Spike 14's Wide layout ([`macos-client-patterns.md` § 9](../../macos-client-patterns.md#9-window-widths-regular-and-wide), brief [S14-D1](../spike-14/design/archive/S14-D1-wide-layout.md)) has no migration PR of its own: each page moves onto Page columns when it is redesigned here. Every brief draws its view at Regular and Wide ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md#wide-layout-binding)). S13-D1 composes the Person frame from Page columns (the S14-D1 board was agreed 2026-10-29: pull the fresh kit), and S13-10 builds `PVPageColumns` from that board and moves `ConclusionDetailPage.page` onto it, so the Event and Place pages share it. S13-D8 (Source page) and S13-D9 (Sources list) draw their Wide frames too; whether the Source page moves onto Page columns in S13-21 is S13-D8's call. Pages not redesigned in this spike stay on [`ideas/wide-layout.md`](../../ideas/wide-layout.md).
- **S13-D1 draws the whole frame.** The first board draws the header, the switcher and all five sections as named places, the four later ones as placeholders, so later boards fill a frame instead of moving it. A later board that needs the landed frame or tile changed adds a follow-up PR to the map; landed PRs are not reopened (R-D2).
- **Engine PRs run ahead of their slice.** Each engine PR is listed in the slice whose chrome needs it, but it may be planned and landed any time its dependencies allow (R-D3). Each read returns **data, not layout**: grouping into families or Sources happens in Go, placement on the page happens in Swift. If a board asks for data a landed read lacks, a follow-up PR adds it.
- **Each PR updates the docs and skills it makes false.** S13-03: [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) (page-only derivations), [`matching-model.md`](../../matching-model.md) only if a knob is added, and [`ideas/kinship-view.md`](../../ideas/kinship-view.md) (siblings shipped for the page). S13-08: [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5 (Unpromote of the last member), [`catalog-deletes.md`](../../catalog-deletes.md). S13-01: [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md) (per-handle invalidation spec). S13-06: the `LifeFacts` and `ReconciledValueRowModel.derived` comments ("its page owns the Why"). S13-09: [`macos-client-patterns.md`](../../macos-client-patterns.md) (sectioned detail page).

### Slice 0 — Removing Observation pins

- **Pins leave in three PRs, the client first** (user, 2026-10-29). S13-23 stops the Mac app asking for or sending pins, confidence and argument, so those wire fields fall idle; S13-24 deletes the engine that offers and writes pins and reserves its wire; S13-25 drops the table. Each step leaves a working app, and no step keeps a shim for the one before. Binds S13-23 – S13-25.
- **Promote's removal has a brief** (user, 2026-10-29; replaces "No brief for the removal"). Removing chrome changes how the page looks, so it is gated like any other change: brief S13-D11 removes the evidence sheet frames and lays the row out without the assessment column (its band, Contested mark, reason and chevron), and decides what the freed width becomes. The target menu, connection subrows, row notes (conflicting decisions, possible duplicates) and the Updated caption keep their look. The caption shows a settling move, else the via neighbor, else nothing (user, 2026-10-29). Binds S13-23.
- **Promote files accepted claims and nothing else** (user, 2026-10-29; status descoped 2026-10-30; [conclusion §5.3](../../conclusion-layer-data-model.md#53-promote)). Done sends no pin ids, `default_pins` false, no confidence, no argument and no status: every filed row is an accepted Identity Claim, as before the spike. Neither Provisional nor Rejected is offered at Promote. A per-row Accepted or Provisional status (S13-30) was planned on 2026-10-29 and descoped on 2026-10-30: Promote writes no provisional claims, and nothing this spike reads them; the schema keeps the status, and the feature is parked in [`ideas/provisional-identity-claims.md`](../../ideas/provisional-identity-claims.md). Binds S13-23, S13-24.
- **Identity Claims carry no confidence and no argument** (user, 2026-10-29). An Identity Claim is membership only: subject, entity, status. S13-23 stops the app reading or sending either field; S13-24 reserves `confidence_grade_id` and `argument` on every wire message that has them (`IdentityClaim`, `PromoteSubjectRequest`, `ApplyPromoteGraphAlignmentRow`), deletes `ListClaimConfidenceGrades`, and drops the auto-reconciler's claim-confidence tiebreak (`autoreconcile` `Provenance.ClaimConfidence`; `autoreconciler.CacheVersion` bumps); S13-25's migration drops `identity_claims.confidence_grade_id`, `identity_claims.argument`, the `claim_confidence_grades` table and its seed (`claimconfidencegrades`, its `Install`, and the onboarding readiness check that reads it), in the same format step as the pin table. Where the reasoning that records belong together lives (between Evidence graphs, or on the entity) is not decided; nothing replaces the argument now, and `canonical_entities.argument` stays. Reconciliation Claim confidence is a separate question, since Reconciliation Claims are not built. Binds S13-23 – S13-25.
- **The score ranks, unseen** (user, 2026-10-29). Graph alignment's scoring keeps choosing and ordering Promote's suggestions (the shortlist gate, the accept bar, settling), but the page shows no band, Contested mark or reason. No PR in Spikes 13 – 14 types, extends or explains the band further. Replacing the scoring waits until the comparison of slice 7 has proven itself; Promote could then rank with it ([`ideas/promote-ranking.md`](../../ideas/promote-ranking.md)). Binds S13-23, S13-24 and Spike 14's S14-22.
- **Audit history keeps its pin rows.** The migration drops the table without recording a removal per pin. `audit_changes` and `audit_transaction_scopes` rows already written for `identity_claim_evidence` stay untouched. Nothing reads audit rows back by entity type (only `sources/notes.go` reads `source_note`), so no retired-type handling is built; the effects registry entry leaves with the `catalogmodel` table, since its init refuses an entry for an unknown table. Binds S13-25.
- **Spike 14's Promote PRs build on the removal.** S14-20 changes `promotealign`, and S14-16 rewrites `ApplyPromoteGraphAlignment`'s handler and Swift mapping and deletes `promoteSubject`; both land after S13-24. S14-22 changes the Promote rows and lands after S13-23. (All three moved from Spike 12 on 2026-10-29, as S12-44, S12-32 and S12-48.) S14-15 (synchronized folders) and S13-23 both touch `project.pbxproj`: whichever lands second rebases, and S13-23 removes the `PromoteEvidence*` file entries only if S14-15 has not landed. Spike 14's S14-02 is re-planned after S13-23, which deletes the Promote files it was to rename. Binds S13-23, S13-24.
- **Claim review becomes how the records connect** (user, 2026-10-29). S13-07 is descoped: there are no stored pins for it to read back. S13-18 and brief S13-D7 are descoped: the Person page shows the person's Evidence graphs and the links between them instead (slice 7, [`evidence-graph-comparison.md`](evidence-graph-comparison.md)).
- **Each removal PR updates what it makes false.** Each PR's Definition of done `rg`s the names it deletes across `.agents/skills`, `.cursor/rules` and `docs/`.
  - S13-23: [conclusion §5.3 and §5.4](../../conclusion-layer-data-model.md#53-promote) already state status at Promote (planning commit); the Swift-side sentences of [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) (the Swift enums and `PromoteEvidenceDisplay`, which `scripts/check-guidance-names.py` fails on once they are gone; the rest stays S13-24's); [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §3 and §8 (the page's draft) and line 7's "Until the code catches up"; [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) (binding order's "sheet is to be removed"); [`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md) (its example key `promote.sheet.subtitle.*`); [`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md) line 80 (share phrases, `recordedDirect`, the sheet's caption); the ideas [`term-management.md`](../../ideas/term-management.md), [`international-names.md`](../../ideas/international-names.md) and [`binding-order-editing.md`](../../ideas/binding-order-editing.md), which describe the sheet as current.
  - S13-24: [`research-judgment-model.md`](../../research-judgment-model.md)'s Claim confidence sections (Identity Claims carry none; say Reconciliation Claim confidence is open, not decided); [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) §4's input table (no claim confidence); [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) in full (the description's package list, the claim evidence lines, the five Swift enums, the `core/claimevidence` bullet, which now says the per-Property router moved to `subjectcompare`, and its test command); [`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md) line 58 (`claimevidence`'s `seedOrder`); `.cursor/rules/match-rules.mdc` (its `core/claimevidence/**` glob); the path lists in [`ship-pr`](../../../.agents/skills/ship-pr/SKILL.md) and [`review-pr/lenses/security.md`](../../../.agents/skills/review-pr/lenses/security.md); [`matching-model.md`](../../matching-model.md) (the `evidence` section and its knob rows); [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §4 and §9; [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) (the Promote package row); the claim evidence mentions in the archived [`comparison-layers.md`](../archive/spike-11/comparison-layers.md).
  - S13-25: [conclusion](../../conclusion-layer-data-model.md) §4 (the confidence and `argument` paragraphs), §5 and §5.3 ("status and argument" becomes "status"), §10 invariant 19, §12's claim argument examples; [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §5.5; [`person-detail.md`](person-detail.md) R-40; [`catalog-deletes.md`](../../catalog-deletes.md); [`audit-revision-history.md`](../../audit-revision-history.md) (the handle scope's "evidence pins"); [`add-catalog-delete`](../../../.agents/skills/add-catalog-delete/SKILL.md); [`add-audit-scope`](../../../.agents/skills/add-audit-scope/SKILL.md); `.cursor/rules/catalog-audit.mdc` ("claims, pins, promote"); [conclusion §5](../../conclusion-layer-data-model.md#5-membership-evidence)'s "planned work" sentence and §10 invariant 20 in the past tense.

### Slice 7 — How a person's records connect

- **Links between Evidence graphs are computed, never stored** (user, 2026-10-29). A node is an Evidence graph holding an accepted member of the person; a link compares the person's two members in two graphs. Nothing is added to the catalog: no researcher-chosen links and no research log. Each Identity Claim keeps its status (it carries no argument, working rule "Identity Claims carry no confidence and no argument"). Binds S13-26 – S13-28.
- **Slice 7 captures requirements, not an implementation** (user, 2026-10-29). Scoring the connection between two Evidence graphs and visualizing the connections are two concerns, specified in [`evidence-graph-comparison.md`](evidence-graph-comparison.md) Requirements 1 and 2 and built separately. The scoring must be pure, layered and driven by the match registry like matching, start from Subject compare, and read none of graph alignment's weights, band or accept bar; how it is built and expressed is open question 11. The slice runs last, after the Promote cleanup, the pin removal and the Person page, and its PRs stay `Blocked` until the design is settled; their scopes are re-cut then. S13-24 still moves `claimevidence`'s per-Property router into `subjectcompare` rather than deleting it, so the starting point survives. Binds S13-26 – S13-29.
- **Each PR updates what it makes false.** S13-26: [conclusion §5.1](../../conclusion-layer-data-model.md#51-two-lenses) and §13 item 4 (the comparison is designed), [`matching-model.md`](../../matching-model.md) and [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) (the new scoring layers, their registry sections and knob rows), [`research-judgment-model.md`](../../research-judgment-model.md) §5 if strength reads credibility. Binds S13-26.

### Engine

- **Each section reads its own query.** The Person page is several reads, each with its own RPC and `CatalogQueryKey`, so a section loads and refreshes on its own: today's `GetConclusionDetail` (header, Vitals, Gallery) plus one each for claims (S13-01), family (S13-04) and timeline (S13-05). `GetConclusionDetail` does not grow the other sections ([`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md)). The three new reads are composed in one package, `core/database/persondetail` (beside `conclusiondetails`); handlers stay thin ([`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md)). Binds S13-01, S13-04, S13-05.
- **One invalidation rule for person-section keys.** Today `CatalogQueryRegistry.invalidations(for:project:)` hand-writes the `touchedHandles` → `.conclusionDetail` fan-out. S13-01 turns that into a descriptor entry with its own `CatalogQueryFollows` selector (per-handle keys refreshed for every touched handle) and declares its key with it; S13-04 and S13-05 add one entry each. A relative's or co-participant's change touches their id, not this person's, so the family, timeline and record-links keys also refresh on any promote or unpromote effect (`.allCached` for that kind) unless the plan finds a keyed effect. Binds S13-01, S13-04, S13-05, S13-27.
- **One person header for every tile.** Family tiles and Timeline participants carry `PersonHeader` (`conclusionheaders.PersonsByIDs`), which already holds name, birth and death `LifeFacts` and likeness, and whose package doc already names "later tree nodes". S13-02 adds the gender and sex terms to it; no new summary type. `PersonsByIDs` returns rows in title order, so S13-04 and S13-05 read through `conclusionheaders.ByIDs` (S11-40), which finds a header by id, instead of re-keying. Binds S13-02, S13-04, S13-05.
- **Canonical only, filtered in Go** (R-2). Reads return canonical entities with at least one accepted member; unpromoted Subjects never cross the wire. One `graphcache` `Node` accepted-members helper (S13-01) replaces the private copies `promotealign.acceptedClaims` and `conclusiondetails.acceptedMembers`, and every reader uses it. Binds S13-01, S13-04, S13-05.
- **Siblings are page-only** (R-18a). The shared-parent sibling derivation lives in `core/kinship` but is **not** in the pattern set `canonlinks.Reader.For` gives matching; `promotealign` and S11-39's goldens see no change, and a test proves it. The Family read asks for siblings explicitly. S13-03 updates [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md) (which today says a sibling pattern "changes `Derive`") with where page-only derivations go. Binds S13-03, S13-04.
- **One person's links, direct and implied, from one function.** S11-51 (moved to Spike 11 to ship right after S11-39, user 2026-10-26) moves S11-39's `linkSet` merge into `kinship.LinkSet` and adds `canonlinks.Reader.Relationships`, which returns a person's recorded and implied relationship links merged once. The Family read calls it. Binds S13-04.
- **Family groups in Go.** One family per spouse (with the children who have both as parents), one per distinct parent set, and children with one known parent in their own group. Binds S13-04.
- **Timeline is direct participation, any role** (R-23). S13-05 adds an unfiltered participation hop to `core/hops` (beside `EventsOfSubject`, which is `subject` role only) and returns, per event, `EventHeader` (title, date, folded places), the person's role term, and the other participants as {role, `PersonHeader`}, ordered by `conclusionheaders.sortEvents` (undated last). Swift titles events with `EventTitleDisplay` and groups the undated ones (R-24). Binds S13-05.
- **A derived row carries its owning entity's records.** `ConclusionDetail` already carries `ReconcilerOutcomeDetail` records per field. S13-06 adds, for each derived row, the records of the field on the entity that owns it (every surviving birth or death event's date, the lead first; each of those events' Places' names, not their parent chain or the location link; user, 2026-10-30), keyed by that entity on `ConclusionDetail`, not on `LifeFacts` (which lists and tiles also carry). Event's derived place row can take the same path later. No new provenance RPC. Binds S13-06, S13-11.
- **Claims are read per person from the graph cache, grouped by Source in Go.** S13-01 reads the person's accepted members from `graphcache` (`Member` has Subject, Source and graph ids; `EvidenceGraph` has the Source title), adds Source type through `sourcetypes`, and returns groups ordered by Source title (R-36). Each claim carries the existing `CatalogSubject` and Observation wire messages the Evidence graph's Subject card already renders. The Source date joins the groups in S13-20. Binds S13-01, S13-20.
- **The Source date is one Source-level attribute, not a metadata key** ([settled](source-provenance-date.md#settled-do-not-reopen)). S13-20 adds it to `sources` with its migration ([`add-catalog-migration`](../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../.agents/skills/add-catalog-model/SKILL.md)); (user, 2026-10-26): it means the **original publication date**, when the Source was created; it is stored as a structured date (`date_value_id` into the shared `date_values`, ABT / BEF like every other date); it is optional, and existing Sources start empty, filled by hand. Metadata date fields stay text and are never copied into it. Binds S13-20, S13-21, S13-22.
- **The Sources list shows and sorts by the Source date** (user, 2026-10-26). Undated Sources sort last. Binds S13-22.
- **One Unpromote, deletes through the registry.** The Person page's Remove calls the same `UnpromoteSubject` as the Evidence graph (R-38). Removing a canonical entity's last accepted member deletes the entity **wherever** Unpromote runs, the Evidence graph included (user, 2026-10-26; R-39). Both confirms warn first. After a delete the Person page lands on the Persons list; the Evidence graph stays where it is. S13-08 deletes the canonical entity through `core/database/deleteimpact` ([`add-catalog-delete`](../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-audit-scope`](../../../.agents/skills/add-audit-scope/SKILL.md)), and the confirm reads an Impact report through an unpromote preflight, as `GetDeleteImpact` and `GetEvidenceGraphUnlinkImpact` do, not a one-off flag. Binds S13-08, S13-17.

### Client

- **Sections are places.** The section is part of the Person page's `WorkspaceLocation`; switching sections pushes history, and Back / Forward and restore return to the section ([`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md); user, 2026-10-26). Binds S13-10.
- **The provenance sheet replaces the Why on the Person page only.** Event and Place keep the Why accordion until they adopt sections; the accordion is deleted then, not this spike (user, 2026-10-26). The shared row takes the sheet or the accordion by caller; this is known debt owned by the Event and Place redesign. Binds S13-11.

- **One sectioned Conclusion page, not a Person fork.** `ConclusionDetailPage` stays the one page for Person, Event and Place. S13-09 gives it a sections slot; Event and Place pass a single section and look unchanged (R-1), and the Person page later passes five. When Event and Place adopt sections, they fill the slot. Binds S13-09, S13-10.
- **One person tile, built once.** The tile is built in S13-12 (Family) and reused by S13-13 (Timeline), placed per [`add-ui-component`](../../../.agents/skills/add-ui-component/SKILL.md) as a recipe once it has the second caller. Briefs S13-D3 and S13-D4 draw the same tile. Binds S13-12, S13-13.
- **Colors reuse `PersonGenderTreatment`** (R-3), fed from `PersonHeader`'s terms (S13-02). Binds S13-12, S13-13.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S13-23 | Promote stops asking for pins | 0 | Swift, plus the Swift-parity rows in `api/ffi/swift_vocabulary_test.go`: the evidence sheet and everything only it uses (board, cards, display, wire enums, share wording, grade picker read, store methods, batch-row pin and claim fields) go; the assessment column (band, Contested, reason) goes; Done sends no pins, confidence or argument, and every row files Accepted; the app reads and sends no claim confidence or argument anywhere; every Promote string says promote, never file | S13-D11 | — | [S13-23](prs/S13-23-promote-stops-asking-for-pins.md) | Landed #633 |
| S13-24 | The pin engine goes | 0 | Go: `GetPromoteClaimEvidence` and `ListClaimConfidenceGrades` and their wire, Apply's pin and claim-field inputs, `confidence_grade_id` and `argument` on every message, and the proposal row's display-only fields (reserved), `promotealign` evidence and pins, `core/claimevidence`, the registry's `evidence` section, Promote's pin writes and backfill, the reconciler's claim-confidence tiebreak, and the helpers only they call; proto regenerated | — | S13-23 | [S13-24](prs/S13-24-the-pin-engine-goes.md) | Planned |
| S13-25 | The pin table goes | 0 | One migration drops `identity_claim_evidence`, `identity_claims.confidence_grade_id` and `.argument`, and `claim_confidence_grades` with its seed; catalog model, `identityclaims` pins, the delete facet and `ReleaseClaim`'s pin step, the effects entry; audit history keeps its old rows | — | S13-24 | — | Scoped |
| S13-30 | Provisional claims at Promote | 0 | Promote files every row Accepted; no status control and no provisional claims (working rule "Promote files accepted claims and nothing else") | — | — | — | Descoped (2026-10-30) → [idea](../../ideas/provisional-identity-claims.md) |
| S13-09 | Sections slot on the Conclusion page | 1 | Preparatory: `ConclusionDetailPage` gains a sections slot; Person, Event and Place pass one section; no visible change | — | — | [S13-09](prs/S13-09-sections-slot.md) | Landed #624 |
| S13-10 | Header and sections | 1 | The fixed header, the section switcher, and Vitals as the first section (today's rows), on `PVPageColumns` (built here from the S14-D1 board) | [S13-D1](design/archive/S13-D1-person-page-frame.md) | S13-09, S14-D1 (board agreed 2026-10-29), S14-24 (landed #628) | [S13-10](prs/S13-10-header-and-sections.md) | Landed #634 |
| S13-06 | Derived rows carry their records | 2 | `ConclusionDetail` carries, for each derived row, the reconciler records of the owning entity's field (every surviving birth and death event's date; their Places' names), and the four life rows show them in the Why | S13-D12 | S13-D12 | [S13-06](prs/S13-06-derived-rows-carry-their-records.md) | Planned |
| S13-11 | Provenance sheet | 2 | The provenance control and sheet replace the Why accordion on every Vitals row, with links to Source, Citation and Observation | S13-D2 | S13-06, S13-10 | — | Scoped |
| S13-01 | Person claims read | 3 | `ListPersonClaims` RPC from the graph cache: a person's accepted claims grouped by Source (title, type), ordered by title, each with its Subject card data, graph and claim id; one accepted-members helper; per-handle invalidation spec | — | — | — | Scoped |
| S13-02 | Gender in the person header | 3 | `PersonHeader` gains the gender and sex terms; `PersonGenderTreatment` reads only the header | — | S11-30 | — | Scoped |
| S13-03 | Siblings from a shared parent | 3 | A page-only sibling derivation in `core/kinship`, outside matching's pattern set, with a test that matching is unchanged | — | S11-39 | — | Scoped |
| S13-04 | Person family read | 3 | `GetPersonFamily` RPC: parents, siblings, spouses, children from S11-51's links, grouped into families, as `PersonHeader`s | — | S13-01, S13-02, S13-03, S11-51 | — | Scoped |
| S13-12 | Family section | 3 | The family diagram with person tiles, lines and several families | S13-D3 | S13-04, S13-10 | — | Scoped |
| S13-19 | One person's links | 3 | Moved to Spike 11 as S11-51 so it ships with S11-39 | — | — | — | Descoped (2026-10-26) → S11-51 |
| S13-05 | Person timeline read | 4 | An any-role participation hop; `ListPersonEvents` RPC: `EventHeader`s sorted by `sortEvents`, with the person's role and the other participants as {role, `PersonHeader`} | — | S13-01, S13-02 | — | Scoped |
| S13-13 | Timeline section | 4 | The chronological event list with roles, depiction placeholder, participants as person tiles, Undated group | S13-D4 | S13-05, S13-12 | — | Scoped |
| S13-14 | Gallery section | 5 | The gallery as a section, with set as portrait and open Citation | S13-D5 | S13-10 | — | Scoped |
| S13-15 | Likeness viewer | 5 | Expand a likeness and step through the gallery | S13-D5 | S13-14 | — | Scoped |
| S13-08 | Unpromoting the last member | 6 | `UnpromoteSubject` on the last accepted member deletes the person through `deleteimpact`, wherever Unpromote runs; an unpromote Impact preflight for both confirms | — | S13-25 | — | Scoped |
| S13-16 | Sources section | 6 | Claims grouped by Source (title, type, date) with Subject cards, each Source opening its page | S13-D6 | S13-01, S13-10, S13-20 | — | Scoped |
| S13-17 | Remove a claim | 6 | Remove on a claim, the last-claim warning, and landing on the Persons list | S13-D6 | S13-08, S13-16 | — | Scoped |
| S13-20 | Source date engine | 6 | An optional publication date on `sources` as a `date_values` reference (migration, model, read and write RPCs); `ListPersonClaims` groups carry it | — | S13-01 | — | Scoped |
| S13-07 | Stored claim read | 7 | Identity Claims have no stored pins to read back (conclusion §5) | — | — | — | Descoped (2026-10-29) → open question 9 |
| S13-18 | Claim review | 7 | Replaced by slice 7's view of how the records connect | — | — | — | Descoped (2026-10-29) → S13-26 – S13-28 |
| S13-26 | Scoring two Evidence graphs | 7 | Requirement 1: a pure, registry-driven function that scores the connection between two Evidence graphs from the Subjects they share, explained in GPS terms; shape decided with open question 11 | — | S13-24 | — | Blocked — on open question 11 |
| S13-29 | One loader for a Subject's comparable values | 7 | Preparatory: `promotealign`'s per-graph layer loader (values with lifespans, bridges, metas) moves into a shared `core/database` package with one memo key; Promote and `persondetail` read it; Go only | — | S13-24 | — | Blocked — on open question 11 |
| S13-27 | Person record links read | 7 | `GetPersonRecordLinks` RPC in `persondetail`: the person's Evidence graphs (Source title, type) and a link for each pair, from S13-26; its key joins the per-person invalidation spec | — | S13-01, S13-26, S13-29 | — | Blocked — on open question 11 |
| S13-28 | How the records connect | 7 | The view of the person's Evidence graphs and their links on the Person page, with what each link agrees and conflicts on | S13-D10 | S13-16, S13-27 | — | Blocked — on open question 11 |
| S13-21 | Source date on the Source page | 8 | Set and show the Source date where the board puts it | S13-D8 | S13-20 | — | Scoped |
| S13-22 | Source date on the Sources list | 8 | Each Sources list row shows the date; the list sorts by it, undated last | S13-D9 | S13-20 | — | Scoped |
| S13-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

Each brief is written just before the PR it gates, its board agreed, and that PR implemented before the next brief (working rule "Design, then build, one brief at a time").

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S13-D11 | Promote page without the evidence sheet and the assessment column | **S13-23** (S13-30 descoped) | Slice 0; an enhancement of the Promote flow board that removes frames. Decides what the freed column width becomes ([brief](design/archive/S13-D11-promote-without-assessment.md)) |
| [S13-D1](design/archive/S13-D1-person-page-frame.md) | Person page: header, section switching, Vitals | **S13-10** | Board agreed 2026-10-29. Explores stacked-with-shortcuts versus tabs (R-7). Draws the whole frame, all five sections as places (working rule "S13-D1 draws the whole frame"), at Regular and Wide on Page columns, after the S14-D1 board (working rule "Redesigns adopt Page columns") |
| [S13-D12](design/S13-D12-life-rows-why.md) | Person page: the Why on the four life rows | **S13-06** | Added 2026-10-30 (user: the life rows show their records now); board agreed 2026-10-30, option (a): a group line per owner. Interim: S13-D2 replaces the Why on the Person page |
| S13-D2 | Provenance sheet | **S13-11** | Reusable beyond Person (R-13); leaves room for Reconciliation Claims (R-10); keeps S13-D12's group line per owner for records from several events or Places |
| S13-D3 | Family section | **S13-12** | Draws the shared person tile; two marriages and conflicting parents |
| S13-D4 | Timeline section | **S13-13** | Reuses S13-D3's tile; depiction placeholder; Undated group |
| S13-D5 | Gallery section and likeness viewer | **S13-14**, **S13-15** | Viewer options: expand, slideshow (R-30) |
| S13-D6 | Sources section | **S13-16**, **S13-17** | Source groups, claim cards, Remove and the last-claim warning |
| S13-D7 | Claim review sheet | — | **Descoped (2026-10-29)** with S13-18 |
| S13-D10 | Person page: how the records connect | **S13-28** | The person's Evidence graphs as nodes and their links, a link's agreements and conflicts, a record that stands apart, two groups that do not connect. Where it sits (inside Sources or beside it) is the board's call |
| S13-D8 | Source page: the Source date | **S13-21** | Where the date lives so it reads as the Source's own date, not another metadata row ([UI](source-provenance-date.md#what-this-idea-still-owns)). Draws the Source page at Wide; decides whether it moves onto Page columns with its pinned header |
| S13-D9 | Sources list: the Source date | **S13-22** | The date on each row and the sort; beside S12-D5's graph buttons. Draws the list at Wide |

## Slices

**✎ = design brief.** **Check** = what you can verify in the app when the slice lands.

```text
SLICE 0 — Removing Observation pins   (precursor, first)
  ✎ S13-D11 ──▶ S13-23  Promote stops asking for pins
  S13-24  The pin engine goes
  S13-25  The pin table goes
  Check: bar 9

SLICE 1 — Page frame and Vitals
  S13-09  Sections slot on the Conclusion page   (preparatory, any time)
  ✎ S13-D1 ──▶ S13-10  Header and sections
  Check: bar 1 — the header and switcher, Vitals filled, the other four as places

SLICE 2 — Provenance sheet
  ✎ S13-D12 ──▶ S13-06  Derived rows carry their records
  ✎ S13-D2 ──▶ S13-11  Provenance sheet
  Check: bar 2

SLICE 3 — Family
  S13-01  Person claims read   (its accepted-members helper and invalidation spec are needed here)
  S13-02  Gender in the person header   (after S11-30)
  S13-03  Siblings from a shared parent   (after S11-39)
          S11-51 One person's links   (Spike 11, after S11-39)
  S13-04  Person family read
  ✎ S13-D3 ──▶ S13-12  Family section
  Check: bar 3 and 8

SLICE 4 — Timeline
  S13-05  Person timeline read
  ✎ S13-D4 ──▶ S13-13  Timeline section
  Check: bar 4

SLICE 5 — Gallery
  ✎ S13-D5 ──▶ S13-14  Gallery section
               S13-15  Likeness viewer
  Check: bar 5

SLICE 6 — Sources
  S13-20  Source date engine
  S13-08  Unpromoting the last member
  ✎ S13-D6 ──▶ S13-16  Sources section
               S13-17  Remove a claim
  Check: bar 6 and the removal half of 7

SLICE 7 — How a person's records connect   (last; blocked on open question 11)
  S13-26  Scoring two Evidence graphs   (after S13-24 and open question 11)
  S13-29  One loader for a Subject's comparable values   (preparatory)
  S13-27  Person record links read
  ✎ S13-D10 ──▶ S13-28  How the records connect
  Check: bar 10

SLICE 8 — Source date on the Source pages   (any time after S13-20)
  ✎ S13-D8 ──▶ S13-21  Source date on the Source page
  ✎ S13-D9 ──▶ S13-22  Source date on the Sources list
  Check: a Source's date set on its page shows on its Sources group and its
         Sources list row, and the list sorts by it, undated last

S13-99 — Dogfood close / docs
```

**Order and parallel work:** slice 0 runs first: S13-23, S13-24 and S13-25 in order; S13-09 and the other engine PRs may run alongside it, except S13-08, which waits on S13-25 because both change `deleteimpact.ReleaseClaim`. The design-and-build slices run in order, 1 to 7, one brief at a time; slice 7 waits on open question 11. Slice 8 touches other pages and may run alongside any of them once S13-20 lands. Engine PRs can be planned and landed ahead of their slice: S13-01 and S13-09 have no unmet dependencies, S13-06 is Planned (its board S13-D12 was agreed 2026-10-30), and S13-08 waits on S13-25; S13-02 waits on S11-30, S13-03 and S11-51 on S11-39.

### Dependencies at a glance

```text
S13-23 ──▶ S13-24 ──▶ S13-25 ──▶ S13-08
S13-23 ──▶ S14-02, S14-22    S13-24 ──▶ S14-16, S14-20      (Spike 14)
S11-30 ──▶ S13-02 ──┬──────────────▶ S13-04 ──▶ S13-12 ──▶ S13-13
S11-39 ──▶ S13-03 ──┤                  ▲                     ▲
S11-39 ──▶ S11-51 ──┘                  │                     │
S13-01 ──┬─────────────────────────────┘                     │
         ├──▶ S13-05 ◀── S13-02 ─────────────────────────────┘
         └──▶ S13-16
S13-09 ──▶ S13-10 ──┬──▶ S13-11 ◀── S13-06
                    ├──▶ S13-12
                    ├──▶ S13-14 ──▶ S13-15
                    └──▶ S13-16 ──┬──▶ S13-17 ◀── S13-08
                                  └──▶ S13-28 ◀── S13-27 ◀── S13-26, S13-29, S13-01
S13-24 ──▶ S13-26, S13-29
each brief S13-Dn ──▶ the PR(s) it gates; slice n's chrome ──▶ slice n+1's brief (1 … 7)
S13-01 ──▶ S13-20 ──┬──▶ S13-21
                    ├──▶ S13-22
                    └──▶ S13-16
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Chrome scopes are provisional until their board is agreed.

### Slice 0 — Removing Observation pins

#### S13-23 — Promote stops asking for pins

**Ships.** The Mac app's half of the removal; after it, nothing in Swift outside `Generated/` names a field or Method that S13-24 removes, so S13-24's regeneration needs no Swift edit.
- **The sheet and what only it uses:** `PromoteEvidenceSheet`, `PromoteEvidenceBoard`, `PromoteEvidenceCards`, `PromoteEvidenceDisplay`; the sheet state, pin choices and per-row status, confidence and argument in `PromoteFlow` and `PromoteModel`; the wire enums `EvidenceOutcome`, `EvidencePropertyCause`, `KinshipPattern`, `ValueRelation`, `ValueCause` in `ConclusionVocabulary.swift` and `L10n.Promote.outcome(…)`; `CatalogEvidenceTerm` and the `CatalogClaimEvidence*` types.
- **The share wording:** `ConnectionSentenceRegistry`'s `KindEntry.share`, `ShareInput`, `share(_:similar:)`, `seededShare`, `namedShare` and its initializer from `CatalogClaimEvidenceRelationship`; `L10n.EvidenceGraph.share*` and the `shareHeader*` keys; `L10n.PropertyTerm.Wording`'s `share`, `shareSimilar`, `shareOtherEnd`, `shareOtherEndSimilar` and `recordedDirect`; the share cases in `ConnectionSentenceRegistryTests`.
- **The grade picker's read:** `PromoteModel.grades` and `confidenceKey`, `PromoteView`'s query, the Promote place's warm of `.confidenceGradesList` in `PlaceRegistry`, `CatalogQueryKey.confidenceGradesList` and its `CatalogQueryRegistry` entry, and `listClaimConfidenceGrades` in `GenealogyStore`, `GoStore` and `FakeStore`.
- **The store and wire mapping:** `getPromoteClaimEvidence` in `GenealogyStore`, `GoStore` (with every evidence mapper), `FakeStore` and the `OnboardingModelTests` conformer; `CoreMethod.getPromoteClaimEvidence`; `CatalogPromoteBatchRow`'s `confidenceGradeID`, `argument`, `defaultPins`, `pinIDs` and `CatalogPromoteBatchWritten.pins`, and their assignments in `GoStore.applyPromoteGraphAlignment`. Done sends none of them, which Go reads as no pins (`ResolvePins` skips a row with neither).
- **The column:** `PromoteView`'s assessment column goes, with `PromoteReadoutBadges` and the band and Contested display. `PromoteFlow.draftTarget` still reads `assessment` and the `taken` reason to decide what a row starts on until S13-24 moves that rule into Go. The Updated caption shows the move reason, else the via neighbor (`via_neighbor_subject_id`), else nothing (user, 2026-10-29), so the row's own and the alternatives' `reason` stop being read. The target menu, connection subrows, row notes and `PromoteUpdatedCaption` stay; the caption's comments drop the sheet as a caller.
- **Strings and tests:** the `promote.sheet.*` keys and the keys the column alone read (`promote.assessment.*`, `assessment.chosenNone`, `row.contested`, `row.assessmentContested`, `row.evidenceFor`, `page.column.assessment`, and the `reason.*` keys off the caption path) from `Localizable.xcstrings`; the assessment and contested cases in `PromoteFlowTests` and `PromoteModelTests` (the draft cases stay until S13-24); the evidence tests; the matching rows and the "every term a sentence shares has its share phrases" subtest in `api/ffi/swift_vocabulary_test.go`, the only Go edit.
- **The claim fields:** `CatalogIdentityClaim.confidenceGradeID` and `argument`, and `promoteSubject`'s `confidenceGradeID` and `argument` parameters with their wire assignments, so S13-24 can reserve those fields with no Swift edit.
- **Promote, never file** (board S13-D11, user 2026-10-29): every user-facing string that names the process says promote — the Promote page's copy and the Evidence graph's Unpromote confirm; the row list widens to 1600 pt at Wide (frame 27W).
- **The project file:** the `PromoteEvidence*` entries in `project.pbxproj` if S14-15 has not landed.

**Not this PR.** The Go evidence engine and wire (S13-24); the table and the claim columns (S13-25); any change to proposals or to the band.
**Check.** Bar 9, app half: on a Source with people, no Promote row opens a sheet and Done files every row; `identity_claim_evidence` gains no rows for that Done; `rg 'pinIds|defaultPins|getPromoteClaimEvidence|ClaimEvidence|confidenceGradesList' macos/App macos/ProvenenciaTests` finds hits only in `Generated/`; Swift tests and `go test ./api/ffi/...` pass. Fan-out at Scale: no evidence read per row, one Apply.
**Context.** Working rules "Pins leave in three PRs", "Promote's removal has a brief", "Promote files accepted claims and nothing else", "Identity Claims carry no confidence and no argument", "Each removal PR updates what it makes false"; [conclusion §5.3](../../conclusion-layer-data-model.md#53-promote); what S11-20 and S11-25 built; `add-localized-string`, `add-swift-test`. **Spike 11 leftovers (2026-10-29):** `CatalogPromoteBatchRow.target` becomes a typed value (reusing `PromoteTargetCode.wire`) while this PR edits the batch row, and `PromoteFlow.item` stops rebuilding a conclusion header the store split into three optionals.

#### S13-24 — The pin engine goes

**Ships.**
- **The RPCs:** `GetPromoteClaimEvidence` and `ListClaimConfidenceGrades` (handlers, dispatch entries, Methods reserved in `engine.proto`, messages deleted). Neither comes back: claims carry no confidence (working rule "Identity Claims carry no confidence and no argument"); S13-25 drops the grades table.
- **The claim fields on the wire:** `confidence_grade_id` and `argument` are reserved on `IdentityClaim` and `PromoteSubjectRequest`; `identityclaims` stops reading and writing the two columns (`CreateInput`, `Claim`, its SQL; moved here from S13-25, 2026-10-30) and `promote` stops taking them; the handlers stop mapping them.
- **The reconciler:** the claim-confidence tiebreak goes (`autoreconcile` `Provenance.ClaimConfidence`, its use in `Weak` and the tiebreak, and its load in `core/database/autoreconciler`); `autoreconciler.CacheVersion` is bumped so every catalog rebuilds its cache on open. `conclusiondetails` and `ReconcilerOutcomeDetail` stop carrying claim confidence (`claim_confidence_key`, `claim_confidence_offset` reserved), and Vitals' weak cause "low-confidence claim" goes, with no brief (user, 2026-10-30). The provenance columns both readers select (Source credibility offset, transcription certainty) get one owner, `autoreconciler.ProvenanceColumns` with its joins and scan helper, read by `conclusiondetails` too (user, 2026-10-30).
- **Apply's inputs:** on `ApplyPromoteGraphAlignmentRow`, `pin_ids`, `default_pins`, `confidence_grade_id` and `argument` are reserved and the result's `pins` count goes; `core/database/promote` loses `BatchRow.ConfidenceGradeID`, `Argument`, their validation, `Written.Pins`, its pin sets and the backfill onto members' claims, so Done files accepted claims and bridges only.
- **The evidence engine:** `promotealign` loses `evidence.go` (with `acceptedClaims`, which has no other caller), `pins.go` (`PinOffers`, `ResolvePins`) and `values.go`. `core/claimevidence`'s per-Property router moves into `subjectcompare` as `Rules.Meta` beside `Rules.Role`, which already gives the rest; no compare over every Property is added and `lifespan` stays unexported (user, 2026-10-30); the rest of `core/claimevidence` (`Build`, relationships, pins, `Rules.Checked`, `ByEye`, `Defining`) goes; the registry's `evidence` section goes with its knob rows and guard entries; and helpers whose only caller was the evidence read go: `kinship.Recording.Shown`, `graphcache.Vocabulary.BindingOrder`. `canonlinks.Reader.Relationships` stays: S13-04 calls it (working rule "One person's links, direct and implied, from one function").
- **Tests:** the evidence budgets in `promotealign/perf_test.go`, `TestEvidenceWarmIssuesNoQuery` and the evidence calls in `TestMemoFollowsCommit` and `TestMemoIsNotMutated`, the `membersMemoKey` snapshot in `export_test.go`, `matchregistry/evidence_test.go`, `promote/pins_test.go`, and the pin parts of `batch_test.go`, `bridges_test.go` and `api/ffi/handlers/promote_batch_test.go`.
- **What a row starts on is Go's** (audit 2 B-26, moved here from Spike 12): the proposal's `target` says whether a row starts on a record, a weak draft on its best match, or unset, and Swift's `PromoteFlow.draftTarget` rule goes with its draft tests.
- **The display-only wire goes:** `PromoteGraphAlignmentRow`'s `score`, `assessment`, `contested`, `reason`, `reason_property_key`, `reason_property_origin` and `PromoteGraphAlignmentAlternative`'s `score`, `assessment`, `reason` and `reason_property_*` are reserved; `graphalign` drops `Flags.Contested` and the row-level comparisons it was built from unless `move` reads them; the alternative's `via_neighbor_subject_id` and the row's `via_bridge_type` and `via_role` are reserved too (audit 2 B-37's wire half); Swift's `PromoteReasonCode` becomes `PromoteMoveReason` with the move codes, and the `swift_vocabulary_test.go` "promote reasons" row shrinks to them.
- The proto is regenerated; Swift's generated file changes, and the only other Swift edit is the draft rule above.

Go, plus deleting Swift's draft rule.
**Not this PR.** The table, `identityclaims` pins, the claim columns and the grades table (S13-25). Apply keeps filing accepted; it gains no `status` (S13-30 descoped).
**Check.** `go test ./...`; a Promote Apply test files claims and bridges and calls no pin writer; an auto-reconciler test with two members that differ only in claim confidence no longer breaks the tie on it; `TestLargeSourceBudgets` without its evidence subtests; `TestKnobInventoryMatchesRegistry` with the evidence rows gone; `rg 'claimevidence|ResolvePins|PinOffers|acceptedClaims' core api` finds nothing.
**Context.** Working rules "Pins leave in three PRs", "Each removal PR updates what it makes false"; what S11-19, S11-25 and S11-40 built; `add-ffi-handler` (reserving Methods and fields), `add-match-rule`. Spike 14's S14-16 and S14-20 wait on this PR.

#### S13-25 — The pin table goes

**Ships.**
- **Schema:** one migration drops `identity_claim_evidence`, `identity_claims.confidence_grade_id` (with its FK) and `identity_claims.argument`, and `claim_confidence_grades` ([`add-catalog-migration`](../../../.agents/skills/add-catalog-migration/SKILL.md)); `catalogmodel` loses the two tables, the column's FK and the pin table's two FKs ([`add-catalog-model`](../../../.agents/skills/add-catalog-model/SKILL.md)).
- **The grades vocabulary:** `core/database/claimconfidencegrades` (registry, `Install`, tests) goes, with its create-time install and the readiness check in `core/onboarding/ready.go` ([`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md)); `identityclaims` already reads and writes neither column (S13-24).
- **Go:** `identityclaims/pins.go` goes; `deleteimpact` loses the pin facet, its probes and `ReleaseClaim`'s pins-first step ([`add-catalog-delete`](../../../.agents/skills/add-catalog-delete/SKILL.md)); the effects registry loses the entry and its audit scope ([`add-audit-scope`](../../../.agents/skills/add-audit-scope/SKILL.md)).
- **Swift:** `PVDeleteImpact`'s `identity_claim_evidence.observation_id` case in `cascadeLines`, `L10n.DeleteImpact.leavesEvidence` and its key, and the `pinnedObservation` preview fixture; `FakeStore`'s `pinsByObservation` and the pin cascade in `observationDeleteImpact`; the pin cases in `PromoteModelTests` and `PVDeleteImpactTests` (`pinnedObservationConfirmNamesTheEvidenceItLeaves`).

**Not this PR.** A Reconciliation Claim exhibit ([conclusion §13](../../conclusion-layer-data-model.md#13-open-schema-questions) item 5); `canonical_entities.argument`; where the reasoning that records belong together lives (not decided).
**Check.** Bar 9, catalog half: a project made before the PR, with pins and with claims that carry a grade and an argument, opens on the new format with no pin table, no `claim_confidence_grades` and no claim grade or argument columns, its claims and their statuses intact, and its `audit_changes` rows of type `identity_claim_evidence` survive the migration; deleting an Observation that was pinned names no handles in its Impact; `TestPragmaHonesty` and the facet-release honesty tests pass; `rg identity_claim_evidence core api macos .agents .cursor` finds only migrations and history docs.
**Context.** Working rules "Audit history keeps its pin rows", "Each removal PR updates what it makes false"; [`catalog-deletes.md`](../../catalog-deletes.md) item 9; S13-08 waits on this PR (both change `ReleaseClaim`).

#### S13-30 — Provisional claims at Promote

**Descoped (2026-10-30).** Promote files every row Accepted, as it does today (working rule "Promote files accepted claims and nothing else"). Nothing depended on it: Apply gains no `status`, the proposal preselects no provisional record, and S14-16 no longer waits on it. Board S13-D11's frames 30, 30b, 30W and 31 (PS-1 – PS-6) are not built. The schema keeps the provisional status; the scope, model rules and design are parked in [`ideas/provisional-identity-claims.md`](../../ideas/provisional-identity-claims.md).

### Slice 1 — Page frame and Vitals

#### S13-09 — Sections slot on the Conclusion page

**Ships.** A preparatory refactor: `ConclusionDetailPage` gains a `pageSections` slot below the header; a convenience initializer supplies today's Details rows, likenesses and Place relationship rows as the default fill so Person, Event and Place call sites stay unchanged; no visible change.
**Not this PR.** Any visible change; the switcher (S13-10); a custom `pageSections` at a call site (S13-10 for Person).
**Check.** The Person, Event and Place pages look the same; existing detail page tests pass.
**Context.** Working rule "One sectioned Conclusion page, not a Person fork"; R-1; `ConclusionDetailPage.swift`.

#### S13-10 — Header and sections

**Ships.** The fixed, display-only header and the section switcher from S13-D1, with Vitals (today's rows) as the first section and the other four as the places the board draws. The section is part of the `WorkspaceLocation`: switching pushes history. Until their PRs land, Gallery shows today's likeness grid unchanged and Family, Timeline and Sources show an empty state; the header thumb keeps its "Use earliest likeness" menu until the S13-D5 board (user, 2026-10-29).
**Reuse first (from S13-09 review MNT-1).** `ConclusionDetailDefaultPageSections` is one block: Details, likenesses and Place relationship sections, with `ConclusionDetailLikenessSection` private. Before building Person's fill, split it into named internal parts (a Details view taking `content` and `refreshing`, and the likeness section), and rebuild the default fill from them. Person's Vitals then reuses the Details view instead of copying the header, spinner and rows.
**Page columns (user, 2026-10-29).** Build `PVPageColumns` from the agreed S14-D1 board (WL-3 – WL-10; its tokens `widthSecondaryMax` and the check of `widthWideBreakpoint` against the CSS) and move `ConclusionDetailPage.page` onto it; the Person frame fills its slots as S13-D1 draws them at Regular and Wide. The Event and Place pages take the shared scaffold with no other change. Build the header slot and pinned mode only if S13-D1 uses them. Read the page's class from `@Environment(\.pvWidthClass)` (S14-24).
**One segmented control (user, 2026-10-29).** The section switcher is the kit's segmented `ChipGroup`. Build it once as `PVSegmentedControl` and move the date-kind, Observation-kind and Property value-type pickers onto it; `PVChip` sizes from `PVControlSize` and takes a glyph.
**Check.** Bar 1: the frame with Vitals; the other four sections are places that later slices fill. At 1366 and 2560pt wide the Person page matches the board's Regular and Wide frames; the Event and Place pages look as before at Regular.
**Context.** R-4 – R-7, R-14, R-15; brief S13-D1; [S14-D1](../spike-14/design/archive/S14-D1-wide-layout.md); working rule "Redesigns adopt Page columns".

### Slice 2 — Provenance sheet

#### S13-06 — Derived rows carry their records

**Ships.** `ConclusionDetail` carries, keyed by owning entity, the `ReconcilerOutcomeDetail` records of each derived row's field (every surviving birth or death event's date, and those events' Places' names), so `ReconciledValueRowModel.derived` rows have records like other fields and show today's Why (brief S13-D12; user, 2026-10-30). Not on `LifeFacts`. The "its page owns the Why" comments are updated.
**Not this PR.** The sheet (S13-11).
**Check.** A handler test: a person with two birth records on one birth event returns both on the birth-date row.
**Context.** Working rule "A derived row carries its owning entity's records"; R-13a, R-14.

#### S13-11 — Provenance sheet

**Ships.** The provenance control beside every Vitals value and the sheet from S13-D2: every record, the reconciler's reasoning, and links to the Source, Citation and Observation. It replaces the Why accordion on the Person page only; Event and Place keep it.
**Not this PR.** Reconciliation Claims (R-10); the control outside Vitals (R-13b).
**Check.** Bar 2.
**Context.** R-8 – R-13b; working rule "A derived row carries its owning entity's records"; `ReconciliationReasoningView`; working rule "The provenance sheet replaces the Why on the Person page only".

### Slice 3 — Family and Timeline

### Slice 3 — Family

#### S13-01 — Person claims read

**Ships.** `ListPersonClaims(personId)` in `core/database/persondetail`, read from `graphcache`: the person's accepted identity claims grouped by Source, groups ordered by Source title, each group with the Source's id, ref, title and type name, and each claim with its claim id, Evidence graph id and name, and the `CatalogSubject` and Observations the Subject card renders. Also: one `graphcache` accepted-members helper with `conclusiondetails.acceptedMembers` moved onto it (`promotealign.acceptedClaims` is deleted by S13-24, its only caller being the evidence read), and the per-handle invalidation descriptor entry (its own `CatalogQueryFollows`) with `.conclusionDetail` and this key declared through it.
**Not this PR.** The section (S13-16); the Source date (S13-20).
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

**Descoped (2026-10-26).** Moved to Spike 11 as [S11-51](../archive/spike-11/README.md), right after S11-39, which is building the merge it lifts. S13-04 depends on it.

#### S13-04 — Person family read

**Ships.** `GetPersonFamily(personId)` in `core/database/persondetail`, from S11-51's links plus S13-03's siblings: parent sets, spouses with their shared children, children with one known parent, and siblings, each as a `PersonHeader` (re-keyed by id) with the relationship kind and whether it was recorded directly or implied. Canonical people only. Its key joins the S13-01 invalidation spec.
**Not this PR.** The diagram (S13-12); grandparents and beyond.
**Check.** Handler tests: two marriages give two families; two conflicting parent sets give two parent groups; a relationship recorded directly and implied by a marriage appears once.
**Context.** Working rules "One person's links, direct and implied, from one function" (S11-51), "Family groups in Go", "Canonical only"; R-17 – R-19.

#### S13-12 — Family section

**Ships.** The Family section from S13-D3: person tiles, lines, several families, each tile opening its person.
**Check.** Bar 3.
**Context.** R-16 – R-22; working rules "One person tile, built once", "Colors reuse `PersonGenderTreatment`".

### Slice 4 — Timeline

#### S13-05 — Person timeline read

**Ships.** An any-role participation hop in `core/hops`, and `ListPersonEvents(personId)` in `core/database/persondetail`: every canonical event the person took part in directly, as `EventHeader`s ordered by `conclusionheaders.sortEvents`, each with the person's role term and the other canonical participants as {role, `PersonHeader`}. Its key joins the S13-01 invalidation spec.
**Not this PR.** The section (S13-13); event depictions (R-26).
**Check.** Handler tests: a witness role appears; a child's birth the person is not linked to does not; undated events come last and flagged.
**Context.** Working rule "Timeline is direct participation, any role"; R-23 – R-28; `hops.EventsOfSubject`, `EventHeader.subjects` (whether `EventSubject` folds into {role, `PersonHeader`} later is logged in the plan).

#### S13-13 — Timeline section

**Ships.** The Timeline section from S13-D4: events in order, roles, depiction placeholder, participants as the S13-12 tile, Undated group.
**Check.** Bar 4.
**Context.** R-23 – R-28.

### Slice 4 — Gallery

### Slice 5 — Gallery

#### S13-14 — Gallery section

**Ships.** The gallery as a section from S13-D5, keeping set as portrait and opening the Citation. Built from the kind-neutral gallery slot on `ConclusionDetailBody` that [S14-02](../spike-14/deployment-plan.md#s14-02--pictures-in-the-app) adds, so the old and new galleries can coexist and retiring the old one does not drop the Event and Place galleries. Not a dependency: if S13-14 lands first, S14-02 moves it onto the slot.
**Check.** Bar 5 without the viewer.
**Context.** R-29, R-31, R-32; `ConclusionPortraitModel`, `SetPreferredLikeness`.

#### S13-15 — Likeness viewer

**Ships.** The viewer from S13-D5: expand a likeness and step through the gallery, with set as portrait and open Citation.
**Check.** Bar 5.
**Context.** R-30.

### Slice 5 — Sources and claims

### Slice 6 — Sources

#### S13-20 — Source date engine

**Ships.** An optional publication date on `sources`, a `date_value_id` into `date_values` ([`add-date-value`](../../../.agents/skills/add-date-value/SKILL.md)): migration, catalog model, read on the Source, the Sources list rows and `ListPersonClaims` groups, and the write RPC. Existing Sources start empty.
**Not this PR.** The Source page control (S13-21); the Sources list row and sort (S13-22).
**Check.** Migration and handler tests; a Source with a date returns it on its Sources group.
**Context.** Working rule "The Source date is one Source-level attribute"; [`source-provenance-date.md`](source-provenance-date.md); [`source-layer-data-model.md`](../../source-layer-data-model.md) §1.2, §5 (updated by this PR).

#### S13-08 — Unpromoting the last member

**Ships.** `UnpromoteSubject` on a canonical entity's last accepted member deletes the entity through `core/database/deleteimpact` (Impact in the same transaction, probes, audit scope), and an unpromote Impact preflight the confirm words, as `GetDeleteImpact` does.
**Not this PR.** The Person page control (S13-17).
**Check.** Promote tests for the last member; the Evidence graph's unpromote confirm words the new outcome.
**Context.** Working rule "One Unpromote, deletes through the registry"; R-38, R-39; `promote.Unpromote`, `deleteimpact.ReleaseClaim`; `add-catalog-delete`, `add-audit-scope`; `EvidenceGraphModel.beginUnpromote` / `confirmUnpromote` (the confirm copy "the handle stays" changes).

### Slice 2 — Page frame and Vitals

#### S13-16 — Sources section

**Ships.** The Sources section from S13-D6, read from S13-01: Source groups with Subject cards, each Source opening its page.
**Check.** Bar 6.
**Context.** R-33 – R-37.

#### S13-17 — Remove a claim

**Ships.** Remove on each claim through `UnpromoteSubject`, the last-claim warning, and navigation to the Persons list when the person is gone.
**Check.** Bar 7, removal half.
**Context.** R-38, R-39; working rule "One Unpromote"; [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md) for leaving a page whose entity is gone.

### Slice 7 — Claim review

#### S13-07 — Stored claim read

**Descoped (2026-10-29).** It read a claim's stored pins back through `promotealign.Evidence`; S13-24 and S13-25 remove both. What a claim review reads instead is open question 9.

#### S13-18 — Claim review

**Descoped (2026-10-29).** Replaced by slice 7: the Person page shows the person's Evidence graphs and the links between them (working rule "Claim review becomes how the records connect").

### Slice 7 — How a person's records connect

#### S13-26 — Scoring two Evidence graphs

**Blocked — on open question 11.** Ships [Requirement 1](evidence-graph-comparison.md#requirement-1--score-the-connection-between-two-evidence-graphs): a function that takes two Evidence graphs and scores their connection from the Subjects they share on canonical entities (primary and relationship entities), the strength of each shared pair, conflicts and contested values, and derived structure such as a spouse relationship against a shared marriage event, every result explaining itself. Pure, layered and driven by the match registry; Subject compare and the per-Property router S13-24 moved into `subjectcompare` are the starting point. The scope is re-cut, and may split by layer, when the design settles.
**Not this PR.** The read (S13-27); the view (S13-28); any change to Promote.
**Context.** [`evidence-graph-comparison.md`](evidence-graph-comparison.md); working rule "Slice 7 captures requirements, not an implementation"; [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md).

#### S13-29 — One loader for a Subject's comparable values

**Ships.** A preparatory move: `promotealign`'s per-Evidence-graph layer loader (`loadLayer`, `loadLayerMemo`, `loadLayerBridges`, `layerValues`, `lifeEvent`, `decode`, `loadMetas`), which builds each primary Subject's values through `subjectprofile.ForSubject` with lifespans and its bridges as `kinship.Edge`s, moves into a shared `core/database` package with one memo key. Promote reads it from there; no behavior change. Go only.
**Not this PR.** The record-links read (S13-27); profile memos (parked in [`ideas/promote-ranking.md`](../../ideas/promote-ranking.md)).
**Check.** `promotealign` tests and `TestLargeSourceBudgets` unchanged; `PROVENENCIA_GRAPH_VERIFY=1` passes.
**Context.** Working rule "A new comparison, not the band"; `core/database/promotealign/layer.go`; `use-catalog-session`. 

#### S13-27 — Person record links read

**Ships.** `GetPersonRecordLinks(personId)` in `core/database/persondetail`: the person's Evidence graphs that hold an accepted member (graph id and name, Source id, title and type), and for each pair a link from S13-26 with its agreements, conflicts and strength. Members and graphs come from the `graphcache` accepted-members helper (S13-01); each member's values and bridges from the shared loader (S13-29); neighbors' handles from the canonical graph. Its key joins the per-person invalidation spec, refreshed on the person's touched handles and, like the family and timeline keys, on any promote or unpromote effect.
**Not this PR.** The view (S13-28); Event and Place.
**Check.** A handler test: three graphs where two share a matching birth and the third shares only a nickname give one strong link and two weak ones; a graph whose member conflicts on everything stands apart. A registry test: promoting the father in another graph refreshes the son's record-links key. Fan-out at Scale: 50 graphs give 1,225 pair comparisons, within a budget subtest.
**Context.** [`evidence-graph-comparison.md`](evidence-graph-comparison.md) § Scale; working rules "Each section reads its own query", "One invalidation rule for person-section keys", "Canonical only".

#### S13-28 — How the records connect

**Ships.** [Requirement 2](evidence-graph-comparison.md#requirement-2--visualize-the-connections), from S13-D10 on the Person page: the person's Evidence graphs and their connections, a connection opening what explains it, a record that stands apart, and groups that do not connect. Each graph opens its Evidence graph. The form (a node web, a chord diagram, or another) is the board's, and the view reads only S13-27's output.
**Not this PR.** Editing claims or links; Event and Place pages.
**Check.** Bar 10.
**Context.** Brief S13-D10; [`evidence-graph-comparison.md`](evidence-graph-comparison.md); S13-16's Sources section, which the board places it in or beside; [`add-ui-component`](../../../.agents/skills/add-ui-component/SKILL.md) (a graph view may reuse the `GraphCanvas` host: [`ideas/graph-canvas-second-host.md`](../../ideas/graph-canvas-second-host.md)).

### Slice 8 — Source date on the Source pages

#### S13-21 — Source date on the Source page

**Ships.** Set, change and clear the Source date on the Source page as S13-D8 draws it, with the app's structured date entry.
**Check.** Setting a census's date on its page shows it on the Person page's Sources group.
**Context.** Brief S13-D8; S13-20.

#### S13-22 — Source date on the Sources list

**Ships.** Each Sources list row shows its Source date as S13-D9 draws it, and the list gains a sort by Source date, undated last.
**Check.** Sorting the Sources list by date puts an 1851 census before an 1871 census, with undated Sources at the end.
**Context.** Working rule "The Sources list shows and sorts by the Source date"; brief S13-D9; S12-D5's row layout.

## Open questions

Questions 1–8 were answered 2026-10-26 and are now working rules or requirements. New ones go here; a PR that depends on one stays `Scoped` and `plan-pr` asks it again.

9. ~~**What does a claim review show now that pins are gone?**~~ Answered 2026-10-29: the Person page shows how the person's records connect (slice 7); S13-18 and S13-D7 are descoped.
10. ~~**Which Properties a comparison counts.**~~ Superseded 2026-10-29 by question 11: slice 7 captures requirements first.
11. **How the connection between two Evidence graphs is scored** ([Requirement 1](evidence-graph-comparison.md#requirement-1--score-the-connection-between-two-evidence-graphs) and its Open list): how shared Subjects, each pair's strength, conflicts and derived structure combine; how the result is expressed (named levels, an explained number, or other); what joins two graphs into one group; whether Source quality and independence weigh in. The user is thinking it through (2026-10-29). Blocks S13-26 – S13-29.
12. **The visualization's form** (Requirement 2): a node web, a chord diagram, or another. Settled by brief S13-D10, after question 11.
