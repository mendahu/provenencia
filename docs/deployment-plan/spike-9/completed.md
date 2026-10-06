# Spike 9 — Completed steps

Finished Spike 9 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S9-NN`, `S9-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| S9-01 | PR | Conclusion schema + stores |
| S9-02 | PR | Promote write v1 |
| S9-03 | PR | Source graph carries membership |
| S9-D8 | Design | Graph subject card: Promote + membership |
| S9-04 | PR | Graph card: Promote + membership |
| S9-05 | PR | Resolver core v1 |
| S9-06 | PR | Resolved-values cache |
| S9-07 | PR | Person header composer + list read |
| S9-07b | PR | Rename configuration: Metadata, Properties |
| S9-D1 | Design | Workspace sidebar: Source / Conclude / Configure |
| S9-08 | PR | Sidebar sections: Source, Conclude, Configure |
| S9-D2 | Design | Persons list |
| S9-09 | PR | Persons list |
| S9-10 | PR | Promote write + reads: existing target |
| S9-34a | PR | Handle search from cached values + kinds filter |
| S9-D9 | Design | Promote shell + choose target |
| S9-11 | PR | Promote shell + choose target |
| S9-D10 | Design | Promote claim fields + save |
| S9-12 | PR | Promote claim fields + save |
| S9-13a | PR | Land migrations 000037 / 000038 |
| S9-13 | PR | Reconciler pipeline + text / integer / term modules |
| S9-13b | PR | Name module |
| S9-14 | PR | Evidence + reasoning in the auto-reconciler cache |
| S9-15 | PR | Detail composer + detail read |
| S9-D5 | Design | Person detail (revised for reasoning) |
| S9-16 | PR | Person detail |
| S9-17 | PR | Pins + backfill engine |

## Steps

### S9-01 — Conclusion schema + stores

Shipped the Conclusion tables and the Go stores Promote writes through. No FFI, no UI.

**What shipped**

- Migration `000032`: `claim_confidence_grades`, `canonical_entities`, `identity_claims` (composite FKs on `subject_type_id`, `UNIQUE (subject_id, entity_id)`, one-accepted partial index, status CHECK), `identity_claim_evidence`, FK / membership indexes
- `claimconfidencegrades`: registry + `Install` at create (`low_confidence` / `moderate` / `high_confidence`); open does not heal
- `canonicalentities`: `Create` / `InsertTx` minting `{ref_prefix}-…` (PER-…), `Get` / `GetTx` / `GetByRef` / `ListByType`; audited `create_canonical_entity`
- `identityclaims`: `Create` / `InsertTx` (subject type copied from the subject; `ErrTypeMismatch` from the composite FK, `ErrAlreadyMember` for a second accepted claim), `AcceptedEntityForSubject` / `AcceptedMembers` (+ `Tx`); audited `create_identity_claim`
- Delete-impact register: composite-FK support in the honesty test (`FromCols`); Subject type → handles, merged-into, claim grade → handles, and pinned Observation → handle probes. `ViaSamenessEvidence` / `KindSamenessClaim` retired

**What stayed out**

- FFI, Swift, Subject-delete Impact naming the handle (**S9-02**)
- Evidence pin API and backfill (**S9-17**); pin L10n (**S9-18**)
- Resolved-values cache (**S9-06**)
- Reconciliation tables, `canonical_entity_notes`, merge behavior
- Grade backfill for projects created before S9-01

### S9-02 — Promote write v1

Shipped the first callable Conclusion write: mint a handle from a Subject and file an accepted Identity Claim, one transaction, over FFI and from Swift. Deletes now release their facets through one audited registry: they name the handles they affect, never refuse because of Conclusion, and audit everything they remove.

**What shipped**

- `core/database/promote`: `Save` mints a handle of the Subject's type plus an accepted claim (zero pins), audited as one `promote_subject` revision. Refuses members (`identityclaims.already_member`) and, in v1, non-primary kinds (`promote.unsupported_type`)
- FFI `METHOD_PROMOTE_SUBJECT` (`PromoteSubjectRequest` → `CanonicalEntity` + `IdentityClaim`); Swift `promoteSubject` on `GenealogyStore` / `GoStore` / `FakeStore` (`membershipBySubject` for S9-03 / S9-04)
- **Facet release** (`deleteimpact/facets.go`): every audited `CASCADE` facet is one registry entry. The entry defines its FK, its `Release` (delete plus audit), a `Remaining` check that fails the delete if the backstop would fire, and, when **Named**, the `Report.Cascades` probe drawn from the same predicate.
  - Domain deletes make one call, `ReleaseFacets`: Subject, Observation, Citation, Source and Source Field. Releases compose: claims take their pins; connection Observations take their notes, pins and owned values.
  - `Released.Handles` is the seam for S9-06 and S9-34.
  - `TestFacetReleaseHonesty` makes every CASCADE FK declare itself audited or silent.
- **Newly audited:** Source notes, metadata, credibility and layout; Citation notes; layout rows on Source Field delete. Connection-facet Observations are now audited as full rows. All of these were previously deleted silently by the schema.
- **Pins never block deletes.** Migration `000033` rebuilds `identity_claim_evidence` with `observation_id … ON DELETE CASCADE` as a backstop only. The S9-01 blocking evidence probe is gone.
- **Confirm lines.** A Subject delete names the handles it leaves ("It's removed from PER-…"). A pinned Observation delete names the evidence it leaves. Any other cascade gets a generic "This also changes {kind} {refs}" line, so none is dropped.
- L10n: error copy for the identity-claim, canonical-entity and promote codes; `canonical_entity` Impact kind noun; the three confirm lines
- Future UI: weak-claim review alert written into model §5.2 and [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md) (Spike 10)

**What stayed out**

- Existing targets and suggestions (**S9-10**); pins and backfill (**S9-17**)
- Resolved-values cache and upkeep on claim create / Subject delete (**S9-06**)
- Graph membership read (**S9-03**) and the card's Promote control (**S9-04**)
- Bridge filing (**S9-28**); the Swift `sameness_claim` leftovers (**S9-18**)

### S9-03 — Source graph carries membership

The Evidence graph now knows which Subjects are promoted and onto which handle. This is data only; the card UI is S9-04.

**What shipped**

- **Membership** = a read-only view of a Subject's **accepted Identity Claim** (model §6's "member"), never a row of its own. It carries the claim id so later claim management (Spike 10) can act on it from the card. Provisional / rejected claims are not memberships; when they reach cards they get their own claims read.
- `identityclaims.MembershipsBySource`: one indexed query (Source's Subjects → accepted claim → handle → type key). Unpromoted Subjects have no row.
- FFI `METHOD_LIST_SUBJECT_MEMBERSHIPS` (`SubjectMembership { subject_id, claim_id, CanonicalEntity entity, kind }`).
- Swift `listSubjectMemberships` on `GenealogyStore` / `GoStore` / `FakeStore`, loaded as the fourth parallel read of the `sourceGraph` key. `CatalogSubject` stays Interpretation-only.
- `SourceGraphRows.memberships` → `SourceGraphPlacedSubject.membership` / `SourceGraphPlacedBridge.membership` (`nil` = unpromoted). The membership survives position patches.
- `CatalogMutation.promotedSubject(sourceId:)` invalidates exactly that Source's graph key.

**What stayed out**

- The card's Promote control, membership row, and dispatching `.promotedSubject` (**S9-04**)
- The handle's resolved name on the card (**S9-09**)
- Busting other Conclusion keys on Promote (lists arrive in **S9-07**)

### S9-D8 — Design: graph subject card

**Board pick (Frame 14, earlier frames amended in place):**
- **One fixed 44pt footer** on every primary card, below Add property. It shows **Promote** until the Subject has an accepted Identity Claim, then the **membership row**.
  - Both states are the same height, so `contentHeight(for:)` adds one constant and promoting never moves an edge or a hit rect.
  - On cited cards the footer is the ruled stack's last row. On uncited cards it bleeds to the edges under a dashed rule.
- **Promote** is a kit secondary small Button with a leading ↗, on the kind wash, because it is still an action on this Subject.
- **The membership row is a Conclusion link:** paper fill, an accent `PER-` badge, an accent chevron, and no kind pigment or micro-caps label.
  - S9-04 shows *Open person page*; S9-09 puts the resolved name in the same slot.
  - Hover tints with the accent, press scales .985, and keyboard focus is an inset accent ring.
- **v1 Promote** is a kit irreversible Confirm: *Create a new Person from CPR-…?*, a *First member* key chip, *Create Person* / *Leave unpromoted*, with focus on cancel.
- **Bridges are unchanged:** no Promote, and no filed mark until S9-D12.

**Rev 1 (designer revision, shipped in S9-04).** This supersedes the footer styling above; the structure and behaviour are unchanged.
- **Size and band:** the footer is **36pt**. It sits on the **kind chip** under a **1px kind-line** rule.
- **Promote:** a kit **ghost** Button. Its inset is 8pt less than the card padding, so the label lines up with the content above.
- **Membership row:**
  - The ref is a **neutral outline Badge** with kind-ink text, and the link text and chevron are in kind ink.
  - Hover is the chip mixed with 10% kind ink; pressed is 18% plus the .985 scale. The keyboard focus spec is unchanged.
- **Why:** the secondary button and paper band lost contrast on the kind washes in dark mode, and the accent band merged into person cards, which share the accent hue.
- **Accepted deviation from the brief:** the membership row now takes the kind colour. It is told apart from Observation rows by shape: no micro-caps label, an outline mono ref, the name, and a disclosure chevron.

**Rev 2 (light-mode footer colour, shipped in S9-04).**
- **New alias** `subject-{kind}-band` (`PVColor.subject{Person,Event,Place}Band`, the `band` role on `EvidenceSubjectKindStyle`). The footer fill and its 10% / 18% hover and pressed mixes move from the chip to the band.
- **Light mode:** OKLab 300 at 45% into 100, so person `#B9CED8`, event `#F2CBB2`, place `#B7D8C9`.
- **Dark mode:** equals the kind chip, so dark is unchanged.
- **Shipped as specified, with a known contrast gap.** Kind-ink text on the band is 5.88:1 for Person, but **4.30:1** for Event and **4.44:1** for Place, just under the brief's own 4.5:1 check. The ghost Promote label passes everywhere (≥4.6:1). This is left for the designers to revisit; the options are a 900-step ink on the band (~7.5:1) or lighter bands.

Brief archived: [`design/archive/S9-D8-graph-subject-card.md`](design/archive/S9-D8-graph-subject-card.md).

### S9-04 — Graph card: Promote + membership

Primary graph cards can now be promoted from the canvas, and promoted cards show their handle.

**What shipped**

- **`EvidenceSubjectCard` footer:**
  - Two new targets, `promote` and `openHandle`, each the full footer rect.
  - The footer's height is one constant (`footerHeight`, 36pt after rev 1), shared by paint, hits and edges.
  - The uncited Add property hit is re-anchored above the footer.
  - Promote ignores the no-Artifact gate.
- **Kit:**
  - `.pvHostInteraction(hovered:pressed:)` lets paint-only hosts drive `PVHoverEffect`, so `PVButton` shows hover and press on the canvas.
  - `PVBadge(text:…, foreground:)` recolours the text only (kind ink on a neutral outline). The tone still owns the fill and the line.
  - `GraphCanvasPointerController.pressedCardAction` supplies the press.
  - Documented in the DesignSystem README under *Host-owned pointer*.
- **Measured card geometry:**
  - Before this, hit targets and edge anchors on both card kinds were worked out from row-height constants. They drifted 3–30pt from the paint (wrapped titles and values, taller rows), and the bottom-anchored footer made that visible.
  - Cards now tag their actionable pieces (`.evidenceCardHitRegion`) and report an `EvidenceCardLayout`. An `EvidenceCardLayoutStore` per graph feeds hit targets, edges and the connect rubber band, so each hit moves with the piece it belongs to. The constants are only the first-frame fallback.
  - `EvidenceCardLayoutTests` checks the measured layout against the drawn pixels, the footer, row bands and the bridge card.
- **Model:**
  - `beginPromote` builds a `PromoteRequest`.
  - `confirmPromote` mints via `promoteSubject`, then applies `.promotedSubject(sourceId:)` so the graph reloads with membership. Errors keep the sheet open with a Callout.
  - `openHandle` returns `nil` until a handle page exists.
- **VoiceOver:** a footer action named *Promote James Robins, CPR-…* or *Open person PER-…*.
- **Copy:** per-kind strings (person / event / place) for the tooltip, the Confirm title, message and action, the link text, and the VoiceOver open action.

**What stayed out**

- Routing *Open person page* (S9-16 / S9-24 / S9-27)
- The resolved name on the row (S9-09)
- The Promote flow (S9-11)
- A filed mark on bridges (S9-D12)
- A keyboard focus ring on the footer: the canvas has no per-control keyboard focus yet, and VoiceOver actions are the keyboard path

### S9-05 — Resolver core v1

The first piece of value resolution (R2): a handle's candidate values for one Property go in as a list and come out as ranked clusters. Pure Go; nothing calls it until the cache (S9-06).

**What shipped**

- `core/resolve`: `Resolve(valueType, candidates, concluded) (Result, error)`. No catalog access, no writes. Reuses `namevalues.Value` / `datevalues.Value` and the `properties.ValueType*` set.
- **Lists in, lists out.** Zero, one, or many candidates → zero, one, or many `Cluster`s (representative value, member Observation ids, support). `Clusters[0]` is rank 1.
- **State is derived, not stored.** `Result.State()` reads it off the shape: empty, `single` (one cluster, support 1), `merged` (one cluster, support > 1), `mixed` (several), `concluded` (the one flag the shape can't carry).
- **v1 clustering is exact equality:** names by normalized `form` (case, punctuation, whitespace ignored); text trimmed but otherwise exact (*York* ≠ *york*); integers, terms, subjects by value; dates on every structured field (`MAY 1985` ≠ `14 MAY 1985` yet).
- **Order:** support descending, then each cluster's lowest Observation id; independent of input order. The representative is the lowest-id member's value.
- **Concluded input** always takes rank 1: it absorbs the cluster it equals, or leads with support 0.
- Unknown value types and values missing their type's field are errors (`ErrUnknownValueType`, `ErrValueMismatch`).

**What stayed out**

- Name and date auto-reconcilers (**S9-13**, **S9-21**) — they replace the cluster key for those types
- Provenance ranking and negative polarity (**S9-14**) — callers pass positive candidates only
- Subject values mapped to handles (**S9-28**)
- The cache, its `state` column decision, and any FFI or Swift (**S9-06** onward)

### S9-06 — Resolved-values cache

The resolver's output now lives in one derived table, rewritten in the transaction of every write that can change it and rebuilt on open when stale. Nothing reads it yet; S9-07 composes the Persons list from it.

**What shipped**

- Migration `000034`: `conclusion_resolved_values` (R3 shape minus `state`; PK `(entity_id, property_id, rank)`; edge, sort, and date-window indexes) and `conclusion_resolved_meta` (cache version). Registered in delete Impact as skipped tables with silent `CASCADE` backstops.
- **No `state` column.** Readers derive single / merged / mixed from the rows, as `resolve.Result.State()` does.
- `core/database/resolvedvalues`:
  - `RecomputeTx(q, entityIDs)` / `RecomputeSubjectsTx(q, subjectIDs)`: delete and rewrite whole handles. Recompute is per handle, not per (handle, Property).
  - Batched loader: candidates, date values, name values, name parts — four queries per batch of up to 500 handles, tested constant.
  - `Rebuild`, `NeedsRebuild`, `StoredVersion`, `EnsureCatalog`; `CacheVersion = 1`. Bumping it is the dev rebuild.
- **Upkeep hooks:** `promote.Save`; `observations.InsertManyTx` (so `AddToCitation`, `citations.CreateWithObservations`, `connect.CreateCitedBridge`); `observations.Update` (old and new Subject); `observations.Delete` (its Subject's handle plus `Released.Handles`); `subjects.Delete` (`Released.Handles`).
- **Open:** `resolvedvalues.EnsureCatalog` beside `searchindex.EnsureCatalog` in `catalogsession` and both `onboarding` paths.
- `core/valuecodec`: the date and name proto converters moved out of the FFI handlers, plus Marshal / Unmarshal pairs. Dates and names are stored as `DateValueInput` / `NameValueInput` bytes (Q12). This is core's first import of `api/proto/engine`.
- `datevalues.LookupManyTx`, `namevalues.LookupManyTx` (querier variants; Catalog readers deadlock inside a write tx). `resolve.SortKey`: normalized form for names, case-folded text, order-preserving integers.
- **Tests:** hand-computed rows per value type; each hook; stale version rebuilds; loader query count; **rebuild equals upkeep**, two ways: `TestRebuildEqualsUpkeep_SeededSequences` (generated sequences from fixed seeds, 4 × 200 steps; deterministic, failures name seed and step) and `TestRebuildEqualsUpkeep_Scenarios` (named hand-written sequences; a failure a seed finds is shrunk into one). Removing any live hook fails the sequences.

**What stayed out**

- Provenance in the loader and its triggers (**S9-14**)
- `date_lo` / `date_hi` and a date `sort_key` (**S9-21**)
- Subject-valued Properties and `value_entity_id`, plus the inbound-end trigger (**S9-28**). Until then the Subject-delete hook cannot change rows (a Subject with Observations is refused), so the rebuild-equals-upkeep tests exercise it without being able to catch its absence.
- Narrowing upkeep to (handle, Property); timings (**S9-33**)
- Readers, FFI, Swift (**S9-07**)

### S9-07 — Person header composer + list read

The first reader of the resolved-values cache: Go composes every Person's row header in one query, the app loads it behind a Conclusion query key, and a stub Persons place is registered. Nothing new is reachable in the app until the sidebar row (S9-08).

**What shipped**

- `core/database/conclusionheaders.ListPersons`: unmerged Person handles with the rank-1 resolved name (NameValue) and the name cluster count, ordered by name `sort_key`, then ref. One query for the whole list (tested constant for 1 and 51 Persons). Structures only; mixed is derived (cluster count > 1).
- `canonicalentities.CountByTypeKey` (excludes merged handles).
- FFI `METHOD_LIST_PERSON_HEADERS` (`PersonHeader { entity, name, name_cluster_count }`); `GetWorkspaceNavCountsResponse.persons`.
- Swift: `CatalogPersonHeader` / `CatalogNameValue`; `listPersonHeaders` on `GenealogyStore` / `GoStore` / `FakeStore` (FakeStore clusters member name Observations by case-folded form).
- **Text in the app:** `NameValueDisplay.string(for:)` (form, else parts in order) and `PersonHeaderDisplay.title` (name → label → ref).
- `CatalogQueryKey.personsList` on `CatalogQueryRegistry.conclusionTriggers` — `savedCitation`, `deletedSubject`, `promotedSubject`, `deletedSource`, `mutatedSourceWorkspace` (credibility; over-busts on notes / artifacts / metadata, accepted).
- Place: `WorkspaceSection.persons` (history id `persons`), `PlaceID` / presentation `personsList`, `PersonsListView` (kit EmptyState), `PVSymbol.person`, L10n. Not in the sidebar.
- `CatalogCounts.persons` from `refreshAll()`.
- `macos-client-patterns.md`: Conclusion keys, the planned eviction exception, structures-vs-text.

**What stayed out**

- Sidebar row and re-publishing the count after Promote (**S9-08**); the designed list and the place's query key (**S9-09**)
- Eviction of non-visible detail keys and visible-place tracking (**S9-15**, with the first detail key)
- A certainty mutation and a dedicated credibility mutation (**S9-14**)

### S9-07b — Rename configuration: Metadata, Properties

Two configuration views took plain names, all the way down: **Source fields → Metadata** (a row is a *metadata field*) and **Subject fields → Properties**. Source types kept its name. No behavior changed.

**What shipped**

- **Storage:** migration `000035` renames `subject_type_fields` → `subject_type_properties` (indexes too; nothing had an FK into it) and rewrites the stored audit strings `source_field` → `metadata_field` and `delete_source_field` → `delete_metadata_field`.
- **Go:** package `sourcefields` → `metadatafields`; kinds `source_field` → `metadata_field` (search, search index, delete Impact) with search `ProjectionVersion` 5; section ids `source-fields` / `subject-fields` → `metadata` / `properties`; `apperr` `sourcefields.*` → `metadatafields.*`; `searchindex.ReprojectMetadataField`.
- **Proto / FFI:** `SubjectTypeField` → `SubjectTypeProperty` (and its List / Assign / Remove RPCs), `GetSubjectFieldsWorkspace` → `GetPropertiesWorkspace`, `SubjectTypeFieldsGroup` → `SubjectTypePropertiesGroup` (its `fields` → `properties`), nav counts `source_fields` → `metadata_fields`. Method and field numbers unchanged.
- **App:** `Features/Metadata`, `Features/Properties` and every type, case, query key (`propertiesWorkspace`), store method, count, L10n key and value, accessibility id, and preview that said the old names. Omnibar chip "Field" → "Metadata field".
- **Old ids still work:** `WorkspaceSection(id:)` maps `files`, `source-fields` and `subject-fields`; history decoding, engine locations (`GoStore.mapWorkspaceLocationFromProto`) and the sidebar all parse through it, so saved history keeps its entries without a load-failure reset.
- **Docs and skills** that describe the live code use the new names. Archives and earlier `completed.md` entries keep the names of their time.

**What stayed out**

- The sidebar layout and section titles (**S9-08**)
- Renaming Source types, or `source_metadata_fields` / `source_type_metadata_fields` (already "metadata")

### S9-D1 — Design: workspace sidebar

**Board:** Claude Design project *Main Application Layout* (`fb7b3e33-1683-4961-aec1-90e6e19f8214`), `Workspace Chrome.dc.html`, Part 2 (Frames 6–11 and Findings). Part 1 is the earlier S2-01 chrome.

- **Sections:** **Source** (Sources) and **Conclude** (Persons, Events, Places) top-aligned; **Configure** (Source types, Metadata, Properties) bottom-aligned above the session footer. Every destination is a top-level row; nothing discloses or collapses.
- **Titles** are `PVSidebarNav`'s own group heading (micro caps, `--text-faint`), localized data — no new component.
- **Research vs Configure**, in order of strength: flexible space (24pt floor), a hairline above Configure inset to the rows, and no counts on Configure rows. Research rows always carry a count; zero reads `0`.
- **Collapsed rail:** titles drop; a 24pt hairline stands in for each, so the rail shows the same three blocks.
- **Short window:** the column scrolls as one (W-5b); the space shrinks to its floor and Configure follows Conclude.
- **Narrate** is reserved after Conclude (Frame 11, not shipped) and takes its height from the space.
- **Icons:** Persons `person`, Events `calendar`, Places `mappin`.
- The board's destination table predates S9-07b and lists `source-fields` / `subject-fields`; the shipped ids are `metadata` / `properties`.

Brief archived: [`design/archive/S9-D1-sidebar.md`](design/archive/S9-D1-sidebar.md).

### S9-08 — Sidebar sections: Source, Conclude, Configure

The sidebar now reads as the research workflow: titled Source and Conclude sections at the top, Configure at the bottom.

**What shipped**

- `WorkspaceSidebarSections`: the sections as data (title, top/bottom placement, rows, counts). `WorkspaceSidebar` renders one `PVSidebarNav` per section in the existing scroll column, with `Spacer(minLength: 24)` and a hairline before Configure, and a 24pt hairline between every section on the collapsed rail. The Configure destinations are no longer children of Sources.
- **Conclude:** Persons, Events, Places, each with a count. `WorkspaceSection.events` / `.places`, their places and presentations, and `ConclusionStubView` (one EmptyState stub for all three; replaces S9-07's `PersonsListView`). `PVSymbol.mapPin`.
- **Counts:** nav counts gain `events` / `places` (`canonicalentities.CountByTypeKey`); `CatalogCounts` exposes them. A Promote recounts the sidebar (`EvidenceGraphModel.confirmPromote` → `CatalogCounts.refreshAll`), so a new handle shows at once.
- L10n: section titles Source / Conclude / Configure, Events / Places titles and stub copy. The unused "Source layer" eyebrow is gone.

**What stayed out**

- The Persons list (**S9-09**), Events and Places pages (**S9-23**, **S9-26**)
- Narrate (Narrative layer, later spike)

### S9-D2 — Design: Persons list

**Board:** Claude Design project *Persons List* (`5c00cb55-8261-4579-b0e6-bc571ab23f78`), `Persons List.dc.html`, frames 01–07.

- **Rows are kit `PVList`** with the feature snowflake `ConclusionListRow` supplying the slots. The anatomy is shared by Events (S9-D3) and Places (S9-D4):

  | Slot | Rule |
  | --- | --- |
  | thumbnail | Generic. 44pt tile, always reserved; the kind's `subject_*` mark. Never implies a photo. |
  | title | Generic. Resolved value → *italic* working label → mono ref. Truncates. |
  | secondary line | Per kind. Person: `b.` group · `d.` group (label · mono date · *italic* place, +N). An empty group is omitted; all empty ⇒ the line is omitted. Filled in S9-32. |
  | ref | Generic. Trailing mono ref, always shown — even when it is also the title. |
  | VoiceOver | One button; the app composes "James Robins, born 14 May 1817 in York, Upper Canada, died 2 January 1880 in Toronto, PER-7KD45". |

- **Frames:** typical list (S9-32 complete), row anatomy, the S9-09 ship state (no secondary line; the tile holds row height), narrow truncation (places give way first, then the title), empty (no count, no action — Promote lives on the graph card), first load (static skeleton, rows inert), refreshing (stale rows live, header meta "· refreshing").
- **No *mixed* marker in rows** — a deliberate deviation from the brief's PL-2, confirmed by the researcher: a mixed value shows its top-ranked value, unmarked, and disagreement is surfaced on the detail page. The S9-D3 / S9-D4 briefs and R5 were updated to match.

Brief archived: [`design/archive/S9-D2-persons-list.md`](design/archive/S9-D2-persons-list.md).

### S9-09 — Persons list

The Persons page now lists every Person, one row per handle, and each row opens that Person's page.

**What shipped**

- **Kit `PVList`** (`DesignSystem/Components/List/PVList.swift`), ported from the design system's React `PVList`: thumbnail slot, title, optional secondary line, trailing mono ref, chevron; rows are buttons (hover, pressed, focus ring), one focus stop with ↑/↓ (⌥ to the ends), Home/End, Page Up/Down, Return/Space; `PVListSkeleton` for first load. Key handling is pure (`PVListKeyboard`) and unit-tested.
- **`PersonsListView`** reads `.personsList` (now the place's query key): `PVSectionHeader` "Persons" with "N persons" / "· refreshing" meta; skeleton on first load; `PVEmptyState` explaining Promote when empty. `ConclusionListRow` gives the `subject_person` tile, the name → *italic* label → mono ref title (`PersonHeaderDisplay.titleSource`), and the trailing ref. No secondary line yet.
- **Person-detail stub place:** `WorkspaceLocation.entityId` (persisted; old history still decodes), `PlaceID` / presentation `personDetail`, breadcrumb `Persons › PER-…`. Rows and the Evidence graph card's membership row (`EvidenceGraphModel.openHandle`, persons only) open it.
- **Card membership row shows the resolved name** in the *Open person page* slot (same slot, no relayout). `identityclaims.MembershipsBySource` joins the rank-1 name from the resolved-values cache; `SubjectMembership.name` carries it.

**What stayed out**

- Life dates and places in rows (**S9-32**); the Person page itself (**S9-16**); Events and Places lists (**S9-23**, **S9-26**)
- Mixed markers in rows (descoped, see S9-D2)
- **Known limit:** the card's name comes with that Source's graph load. A name edit made on *another* Source reaches this card when its graph next reloads, because the graph key is per Source.

### S9-10 — Promote write + reads: existing target

Promote can now file a Subject onto an existing handle, carry a confidence grade and argument, and suggest which handles to join. Engine and store only; the flow UI is S9-11 / S9-12.

**What shipped**

- **Join:** `promote.Input.EntityID` (nil mints). A join loads the handle, refuses a missing or merged one (`promote.invalid`) and one of another type (`identityclaims.type_mismatch`), and writes the claim only — one `promote_subject` revision with one change. The handle's cache is recomputed as on a mint.
- **Claim fields:** `ConfidenceGradeID` (optional) and `Argument` reach the claim on both paths; an unknown grade is `identityclaims.invalid`.
- **Rebuild equals upkeep:** the seeded sequences join same-kind handles (and fail if a seed never joins); scenario *joining a second member merges its name, then disagrees*.
- **Matching** ([`docs/matching.md`](../../matching.md)), built as its own module for Promote, merge hints and later surfaces:
  - `core/match` (pure): per-kind **Profiles** of **Features** (Property, Comparer, Weight, Contradiction, plus MinScore), with additive scores and per-Feature **reasons**.
  - **Names are compared word by word, with part types as data:** each word has a role (family, given, nick, untyped) and a weight. Any word can pair with any word of the other name, discounted when roles differ, so names in different formats still connect. A surname conflict and a suffix conflict (Jr. vs Sr.) scale the score. Roles come from a part-type → role map that a name format profile can supply. A name with no typed parts is read from `form` as untyped words.
  - Word matching counts an adjacent-letter swap as one edit and allows one added or dropped letter in short names. Dashes separate words; this is a shared-normalizer change, so it comes with resolved-values `CacheVersion` 2.
  - Other comparers: text, term (with neutral terms), date (spans of years: points, ranges and BEF/AFT bounds, with tolerance and ABT widening) and integer.
  - Comparer settings are pointers set with `match.Set`, so zero is a real setting.
  - **One registry** (`core/match/registry.go`) holds every weight, score and limit, the built-in name patterns, and the default profiles. Comparers fall back to it, and guard tests keep it complete.
  - Culture-specific name logic lives in a `NamePattern` from a `NamePatterns` source. Only `western` is built in; a data-driven source replaces it later.
  - The adversarial review's remaining items are shelved in [`ideas/name-matching-enhancements.md`](../../ideas/name-matching-enhancements.md).
  - Default profiles for person (name, sex at birth), event (type, date, start / end) and place (toponym). They are tunable per call through `Profile.With`, or in `core/match/registry.go`.
  - `resolve.NormalizeForm` is exported so names compare exactly as the cache keys them.
- **`core/database/matching`:** `ForSubject` (Observations as the probe; never the Subject's own handle) and `ForEntity` (a handle's cached values as the probe; merge hints). Candidates are every unmerged same-type handle's cached values at every rank, with a constant query count. `loadCandidates` is the seam for blocking.
- **`promotetargets.Suggest`:** shapes `ForSubject` for the picker. Each suggestion has the entity, score, reasons, and a `PersonHeader` for Persons. Events and Places are suggested too, as the handle alone. Default limit 10. Non-primary kinds are `promote.unsupported_type`; `promote.PrimaryKind` is the shared check.
- `conclusionheaders.PersonsByIDs` and `canonicalentities.GetManyTx`: headers and handles for a set of ids, each one query.
- **FFI:** `PromoteSubjectRequest` gains `entity_id`, `confidence_grade_id` and `argument`; `IdentityClaim` gains `confidence_grade_id` and `argument`. New `METHOD_LIST_PROMOTE_TARGET_SUGGESTIONS` (`PromoteTargetSuggestion`: entity, score, `MatchReason`s, optional `person`) and `METHOD_LIST_CLAIM_CONFIDENCE_GRADES` (`ClaimConfidenceGrade`).
- **Swift:**
  - `promoteSubject(…, entityID:, confidenceGradeID:, argument:)`; a protocol extension keeps the mint-only call.
  - `listPromoteTargetSuggestions` returns `CatalogPromoteTargetSuggestion` / `CatalogMatchReason`.
  - `listClaimConfidenceGrades` and `CatalogClaimConfidenceGrade`; `CatalogIdentityClaim` carries the grade and the argument.
  - FakeStore mirrors the join, a simple name score, and the three grades. No query key: S9-11 decides whether suggestions are a key or a model fetch.

**What stayed out**

- Pins and backfill (**S9-17**); related-first suggestions during a walk (**S9-29**)
- Event and Place **headers** on suggestions (**S9-22**, **S9-25**). Event and Place matching itself ships.
- Signals that need edges: a Person's life dates and places, family context (**S9-28**, **S9-29**)
- Researcher-facing tuning, or profiles persisted per project
- Blocking candidates through the R8 search index (**S9-34**)
- The Promote flow, picker and claim fields (**S9-11**, **S9-12**)

### S9-34a — Handle search from cached values + kinds filter

Catalog search can now find Persons, Events and Places, on request. This was pulled forward from Slice 10 for the Promote picker (S9-11). The omnibar doesn't show these kinds yet; S9-35 designs their rows.

**What shipped**

- **Index:** `searchindex.ReprojectHandles` builds one document per unmerged handle from the resolved-values cache, at every rank:
  - **Person:** rank-1 name as the title; other name groupings as secondary text.
  - **Place:** rank-1 toponym as the title; other toponyms as secondary text.
  - **Event:** the event-type term label as the title; other types and date years as secondary text.
  - The working label, then the ref, are fallbacks. Values are data, not composed sentences.
  - Merged handles drop out, and `RebuildAll` covers handles.
  - `ProjectionVersion` 6.
- **Upkeep:** a single hook. `resolvedvalues.RecomputeTx` reprojects the handles it rewrites, in the same transaction, so every write that changes a handle's values (promote mint or join, Observation writes, Subject delete) updates search. On open, the cache is ensured before the search index.
- **Rebuild equals upkeep:** the cache's seeded sequences and scenarios now also compare handle search documents with a full rebuild.
- **Engine:** kinds `person` / `event` / `place`, with `DefaultInEverything: false`. `Query.Kinds` filters candidates in SQL at every retrieval step (ref, full-text, typo shortlist). Locations go to `persons` / `events` / `places` with `entityId`, and `Hit.MemberCount` is filled in one query per page.
- **Proto, FFI and Swift:**
  - `SearchCatalogRequest.kinds`, `SearchHit.member_count`, `WorkspaceLocation.entity_id`.
  - `searchCatalog(…, kinds:)` on the store; a protocol extension keeps the omnibar's call.
  - `CatalogSearchHit.memberCount`.
  - FakeStore searches handles by name, ref or label.

**What stayed out**

- Header-built documents ("Birth of James Robins", life dates and places) and header dependents (**S9-34**).
- Omnibar rows for handle kinds (**S9-35**).
- Event search is thin until the Event composer and `event_name` exist (**S9-20**, **S9-22**): it matches by type, year or ref.

### S9-D9 — Design: Promote shell + choose target

**Board:** Claude Design project *Promote flow* (`bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd`), `Promote flow.dc.html`, frames 01–06.

- **Shell decision (PT-7): a workspace place, not a sheet.** The walk (S9-D12) can run for many subjects, and compare (S9-D11) needs the page's width. A sheet would hold the graph hostage and cap the width at a dialog's. Promote follows the citation composer: it pushes onto history, the toolbar's Back returns to the graph, and the graph redraws with the subjects filed.
- **The shell on every step:**
  - the subject (ref, label, Source) in a header band in the kind's wash;
  - step progress, composed from text and one chevron icon (no stepper component);
  - Done;
  - the leave guard.

  Back and Done both ask first when the step holds an unsaved choice.
- **Steps:** mint is Choose a Person → Claim fields; join is Choose a Person → Compare → Claim fields. The row grows when Existing is chosen.
- **Choose target:**
  - kit radios **New Person** / **Existing Person**, each with a description;
  - Existing indents a kit ComboBox (person rows) and a **Suggested** group: section header with count and "Ranked by name, dates, place and event type";
  - candidates are the D2 list row (tile, name, years, italic place · members line, trailing ref) with a leading radio;
  - Next is disabled until a row or a search result is chosen;
  - a failed search keeps the suggestions visible;
  - no suggestions: a compact EmptyState in a dashed frame.
- **Leave guard (frame 05):**
  - kit Confirm, irreversible tone, cancel first;
  - "Leave Promote without filing James Robins?" / "Leave Promote" / "Keep promoting";
  - leaving with no choice, or right after a save, doesn't ask.
- **During a walk (frame 06):** a "Related to PER-…" group comes first, a "1 saved" footer badge appears, and the header changes to the next subject. Ships with S9-29 / S9-30.

Brief archived: [`design/archive/S9-D9-promote-target.md`](design/archive/S9-D9-promote-target.md).

### S9-11 — Promote shell + choose target

The graph card's Promote now opens the Promote place: choose a new or existing handle, then file the subject.

**What shipped**

- **Place:**
  - `SourceSurface.promote`, `PlaceID` / presentation `sourcePromote`, and `WorkspaceLocation.promote(…)`. The subject's kind travels as `subjectTypeKey`.
  - A registry spec at priority 120, keyed on `sourceGraph`, `sourcesList` and the new `promoteTargets(project:subjectId:)` (`listPromoteTargetSuggestions`, invalidated on `conclusionTriggers`).
  - A host arm with a frozen `PromoteEntry` (fail-closed).
  - Breadcrumbs `Sources › Evidence graph for {Source} › Promote {ref}`.
  - A subject that is gone or already promoted returns to the graph.
- **`Features/Promote`:**
  - `PromoteView`: header band, step row and footer.
  - `PromoteTargetStep`: radios, search, Suggested rows and the empty state.
  - **`PromoteFlow`, the state machine:** a pure value and the single source of truth.
    - It holds a queue of subjects (the walk), the current step and its plan (the steps the choice implies, as designed), the draft, and a phase: editing, saving (with any navigation held until the write lands), confirming leave, or finished.
    - `send(event)` returns the effects to run: save, refresh after save, navigate to the graph, resume or cancel navigation.
    - `requestLeave` answers the leave guard.
    - Advancing skips steps that aren't built yet (`PromoteStep.built`), so later PRs add a step by building it.
  - **`PromoteModel`:** sends events and runs effects (the write, `.promotedSubject`, counts refresh, navigation). It also holds the debounced search results, as view data rather than flow state.
- **Kit:**
  - `PVRadio` and `PVRadioMark`, ported from `PVRadio.jsx`.
  - `PVComboBox` is content-agnostic: rows (`row`) and the empty line (`empty`, given the typed query) are caller-built views. The kit keeps only the generic plain row, and gains **remote results** (`onQueryChange`). Promote's search row (`PromoteSearchRow`) lives in the feature.
  - `PVEmptyState(verbatimTitle:)`.
  - `PVSymbol.userSearch`.
- **Engine:** `PromoteTargetSuggestion.member_count`, for the "N members" line.
- **Graph:**
  - The card's Promote (pointer and VoiceOver) returns the Promote location.
  - The S9-04 confirm, its `promoteConfirm*` / `promoteCancel` / `promoteFirstMember` strings, and `EvidenceGraphModel.catalogCounts` are removed.
- **Copy:** `L10n.Promote`, with person / event / place variants.

**Deviations from the board (interim)**

- **Next files the claim straight away** (accepted, no confidence or argument), because the claim step (S9-12) and compare (S9-19) don't exist yet. The step row shows them as drawn. The footer hints say what happens now ("A new Person is filed when you press Next"; "James Robins will join PER-… and its 2 members") until those PRs restore the board's copy.
- The leave guard leaves out the walk sentence until S9-30.
- **Candidate rows:** no life years or place until S9-32. The line under the name shows only the member count.
- Events and Places use the same flow with their own copy. Their candidates show label or ref until their header composers exist (S9-22, S9-25).

**What stayed out**

- Claim fields (**S9-12**), compare (**S9-19**), the walk and related-first ordering (**S9-29**, **S9-30**).

### S9-D10 — Design: Promote claim fields + save

**Board:** frames D10-01 – D10-06, added to Claude Design *Promote flow* (`bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd`), `Promote flow.dc.html`, inside the S9-D9 shell.

- **The step (CF-1 – CF-3, CF-6):**
  - "Claim fields" with the step count;
  - a sunken summary card, "Will write · one Identity Claim": subject and CPR → the handle (New Person with its pending ref, or the chosen one), a pins badge, and a line about the handle;
  - **Status** (a real Select with one option, Accepted, so Provisional and Rejected arrive without a relayout) and **Confidence** ("Not stated" plus the grades), side by side;
  - **Argument**, drafted from Compare's confirmed matches on the join path, empty on a new handle.
- **Done discards the step (CF-4).** A claim is a researcher's statement, so it is never written on the way out. Done asks through the shell's leave guard; only Save writes. Right after a save, or after a failure that wrote nothing, Done leaves without asking.
- **Back (added during review).** A ghost "Back to {previous step}" leads the footer after the first step. It returns inside Promote with the choice and the fields kept, and never asks. The toolbar's Back still leaves Promote through the guard.
- **Saving (CF-5):** fields, Back and Done lock; the primary button shows the kit loading state and reads "Saving".
- **Failure (CF-5):** a race where another window filed the subject. A compact danger Callout names the subject and the handle and says nothing was written; Save is disabled because retrying can't succeed.
- **Saved:** a past-tense toast ("James Robins filed on PER-…"); the walk opens the next subject (S9-D12), or the flow returns to the graph.

Brief archived: [`design/archive/S9-D10-promote-claim-fields.md`](design/archive/S9-D10-promote-claim-fields.md).

### S9-12 — Promote claim fields + save

Next on the target step now opens Claim fields, and **Save & next** writes the Identity Claim with its confidence and argument.

**What shipped**

- **`PromoteFlow`:**
  - `.claim` is built. Mint runs Choose → Claim fields (2 of 2); join runs Choose → Claim fields (3 of 3), skipping Compare until S9-19 builds it.
  - **Navigation is derived in one place.** `controls` gives where Back goes, whether the primary button moves to a step or saves (`Advance.step` / `.save`), and what is enabled. Footer and model read it and never check the step. With Compare built, the join path stops on it with no footer or model edits; a test injects Compare to prove it.
  - **Edits belong to steps.** Inputs are `.edit(Edit)`, and each `Edit` names its step (choose / select target on Choose, confidence / argument on Claim fields); `send` ignores an edit anywhere else.
  - Back and forward keep the whole draft. Changing the target keeps the claim fields, since nothing is drafted from it yet.
  - `Draft.status` (`ClaimStatus`, Accepted only).
  - **Blocked phase:** a write refused for good (`identityclaims.already_member`). Nothing is unsaved, so leaving and Done don't ask. The graph's report of the handle fills the callout title instead of bouncing to the graph. A retryable failure stays on the step with Save live.
  - `saveSucceeded(entityRef:)` and the effect `announceFiled`.
- **Feature:**
  - `PromoteClaimStep` (the brief's snowflake).
  - The footer's ghost Back.
  - The primary button reads Next, Save & next, or Saving, from `controls`.
  - The filed toast goes through `WorkspaceSession.noticeToast`, so it shows on the graph after the place closes.
  - The new-handle ref prefix (PER, EVT, PLC) comes from the subject types in the cached Properties snapshot, not from code.
  - Promote names the subject by its first asserted naming Observation, else its label, else its ref (`SourceGraphPlacedSubject.displayName`): a Person's `name` form, a Place's `toponym`. Events use the label for now. The header, hints, leave guard and toast all use it.
- **Query:** `confidenceGradesList` (`listClaimConfidenceGrades`, session-fresh, never invalidated) is on the Promote place's keys, with `propertiesWorkspace`.
- **Kit:**
  - `PVButton` and `PVCallout` titles take `PVCopy`, so formatted strings work as labels.
  - `PVSymbol.arrowLeft` and `.pin`.
  - The vocabulary toast overlay hides an empty body.
- **Copy:**
  - Claim-fields strings in `L10n.Promote`.
  - The New hint is the board's: "A new Person goes straight to its claim fields".

**Deviations from the board**

- **No pins.** The badge always reads "No pins" (pins are S9-17 / S9-19).
- **The join argument isn't drafted** and its hint is "Optional"; the summary line has no Compare clause; the choose-target join hint keeps the interim "will join PER-… and its N members". All wait for S9-19.
- **No walk copy:** the toast has no "next in this walk" body, and the callout leaves out "earlier steps stay saved". Both wait for S9-30.
- **Confidence options show the vocabulary's own labels** ("Low confidence", …), not the board's Low / Moderate / High.

**What stayed out**

- Provisional and Rejected statuses (a `ClaimStatus` case each, plus engine support).
- Compare (**S9-19**), pins (**S9-17**), the walk (**S9-30**).

### S9-13a — Land migrations 000037 / 000038

The first step of the replanned slice 4 ([`conclusion-reconciliation.md`](../../conclusion-reconciliation.md)). PRs #255 and #256 were built to the first plan and closed, but the researcher's local projects had already run their migrations, so `main` refused to open them (a newer `user_version`). This lands both migrations byte-for-byte so those projects open again.

**What shipped**

- **Migration 000037** retypes `initial` name parts as `given`. An initial is the part it stands for, and the name reconciler compares parts only with the same type. The `initial` part type is gone from the Go registry and the Western name pattern, the Swift `NamePartType` enum, the `nameValue.part.type.initial` string, the tests and the seeded-vocabulary lists. Cherry-picked unchanged from #255.
- **Migration 000038** adds `conclusion_resolved_values.against` (default 0). Nothing writes it until S9-14. Copied unchanged from #256.
- No cache version change: resolution didn't change.

**Rules this sets**

- Migrations 000037 and 000038 are fixed. Later changes are new migrations (000039 on), never edits.
- Cache versions start at **5**. Projects may carry caches stamped 3 or 4 by the closed PRs' builds, and a new meaning must never reuse a stamp.

**What stayed out**

- Everything else from #255 / #256: the pipeline (**S9-13**), name module (**S9-13b**), evidence and reasoning (**S9-14**).

### S9-13 — Reconciler pipeline + text / integer / term modules

Every value type now goes through one reconciler pipeline ([`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) §5). Nothing the pipeline declines to display is dropped: every value keeps its cache row with the reason.

**What shipped**

- **`core/resolve` pipeline** (`pipeline.go`):
  - rank by provenance, then id;
  - admit (no evidence; provisional never displayed);
  - deny (a stronger negative eliminates the same value; equal or weaker only counts against);
  - group and fold per unit;
  - majority in distinct Sources (`MinMajoritySupport` 2 and more than half);
  - confidence (weak values drop when non-weak evidence survived);
  - survivors grouped into displayed values.
- **Result:** `Clusters` holds every value, displayed first, each with `Support` (distinct Sources), `Against` and `Reason`. `Candidates` holds one `Outcome` per candidate. `State()` reads displayed values only.
- **Candidate inputs:** `SourceID`, `Provenance` (`Weak`, `Stronger`), `Negative`, `Provisional`. An empty `SourceID` counts as its own Source.
- **Module interface** (`modules.go`): `split` into units, `fold`, `assemble`. The interface is ready for names' per-part-type units.
  - **Text:** trimmed, **case-insensitive**. This changes S9-05's rule: *York* = *york*.
  - **Integer.**
  - **Term:** `NeutralTermKeys` (`unknown`, `indeterminate`) are no evidence.
  - **Interim** name (normalized form), date (every field) and subject (id) modules keep S9-05's behaviour.
- **Cache:**
  - migration **000039** adds `conclusion_resolved_values.reason`;
  - every value is written, with `support`, `against` and `reason`;
  - the loader reads each term's key;
  - `CacheVersion = 5`.
- **Readers:** the Persons header's *+N* counts displayed names only. Search and matching read every value, so an outvoted name still finds its Person.
- **Tests:**
  - `TestPipeline`, 22 cases, mutation-checked: removing majority, confidence, deny, provisional or Source counting fails its group;
  - `TestPipelineOutcomes`, `TestPipelineConcluded`;
  - seeded invariants (input order, outcomes, negatives never members);
  - `TestModules`, `TestProvenance`;
  - the cache keeps an outvoted row;
  - stale versions 0, 2, 3 and 4 rebuild;
  - rebuild-equals-upkeep compares `reason` and `against`;
  - `+N` counts displayed names only;
  - search finds an outvoted name.

**What changed for researchers**

- Case-only differences in text values merge (*York* / *york*).
- With two of three records agreeing, the third value is no longer displayed or counted in *+N*. It keeps its row and stays searchable.

**What stayed out**

- The name module (**S9-13b**), date module (**S9-21**) and subject module (**S9-28**).
- Loading Sources, provenance, negatives and provisional members (**S9-14**). Until then every candidate is its own Source and of standard strength.
- The per-candidate reasoning table (**S9-14**, migration 000040).
- Multi-valued cardinality (**S9-36**).

### S9-13b — Name module

Names now reconcile by their structured parts on the shared pipeline ([`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) §7.2), into **one name per Person**. *J. Robins* and *James Robins* are one *James Robins*; *Jake Robins* and *James Robins* are one *Jake James Robins*.

**What shipped**

- **`nameModule`** (`core/resolve/names.go`):
  - **Split:** one unit per part type; each unit is the type's **words** in idx order: parts normalized (accents kept) and split on spaces and hyphens, so *Smith-Jones* = *Smith Jones* = *Smith* + *Jones*. The recorded parts are what's displayed. `form` is never read; a name with no parts is no evidence.
  - **Fold:** a unit folds into a fuller one when its parts map, in order, onto a subsequence, each equal or an initial of it (`[J]` → `[James]`, `[James]` → `[James, Kenneth]`).
  - **Majority only outvotes misspellings:** a minority value is outvoted only when it's a spelling variant of the winner (same number of parts, each similar by `SpellingSimilarity`). A different name is never outvoted; weak evidence still drops.
  - **One name, never mixed:** every displayed record contributes to a single NameValue. Each type keeps every surviving value, best supported first, skipping a recorded part whose words are all present already. Assembled in the type order of the member with the most types, or a member's own value when one carries exactly those parts.
- **Pipeline hooks** for modules: `oneValue` (every displayed candidate forms one value) and `outvotes` (what majority may remove). Assembly receives every surviving value per unit.
- **Spelling rule shared with matching:** `resolve.SpellingSimilarity` and `resolve.EditDistance` (moved from `core/match`), thresholds `SpellingFloor` 0.8 and `SpellingShortMin` 3, which the match registry now points at.
- `resolve.IsInitial` (moved from `core/match`). `CacheVersion = 8`.
- `namevaluestest.Western` builds given parts and a surname for fixtures. Fixtures in conclusion headers, matching, promote targets, search and FFI use it. Two matching scores move (9.3 → 9.1, 5.0 → 6.0) because both sides are now typed.
- **Tests:**
  - `TestReconcileNames` (55) and `TestReconcileNamesProvenance` (20), ported from the closed #255 / #256, with each row's reason pinned. 23 expectations changed when names became one structure and were re-reviewed case by case.
  - `TestReconcileNamesCombine` (7): nickname and given name, two given names, a minority different name kept, a married surname kept, a misspelling outvoted, a misspelling tie, weak still dropped.
  - `TestReconcileNamesWords` (6): hyphenated, spaced and separate given names; the best-ranked spelling displayed; a maiden surname folding into a hyphenated married one; the same words in another order; a misspelt word outvoted; a different second surname kept.
  - `TestSpellingSimilarity`, `TestReconcileNamesValue`, seeded invariants.
  - Cache: an initial expands; a parts-less name isn't cached; deleting the full given name drops back to *J. Robins*; a new given name joins the one name; stale versions 0–6 rebuild.
  - Mutation checks: no folding fails 19 cases, whole-name keys 34, no one-value 30, outvoting anything 10, outvoting nothing 9, whole parts instead of words 5.
- Benchmark: about 0.87 ms for 200 names (performance ledger).

**What changed for researchers**

- A Person shows one name. Different given names, nicknames and surnames from different records all appear in it: a married surname beside the birth one, a nickname recorded as a given name beside the given name.
- A misspelling in the minority (*Robbins* beside two *Robins*) is not shown; it stays cached and searchable.
- Hyphenated and spaced names match their separate parts: *Mary-Ann* = *Mary Ann*, and *Smith* or *Jones* folds into *Smith-Jones*. *O'Brien* vs *O Brien* still differs (spaced particles, in the name-matching ideas).
- A name with no parts isn't shown; that Person reads by its label until a member's name has parts.
- The form reads the parts in order (*Jake James Robins*); display styles will separate alternatives later.

**Deviations from the plan**

- **One name per Person, misspellings-only majority and word comparison** were decided while building it, after working through examples. The plan had kept competing given names as separate names (mixed).
- The code landed in fewer commits than planned so every commit's tests pass.

**What stayed out**

- Name display styles (natural / sorted) and name format profiles.
- Accent folding, nicknames dictionaries, phonetic matching ([`ideas/international-names.md`](../../ideas/international-names.md), [`ideas/name-matching-enhancements.md`](../../ideas/name-matching-enhancements.md)).
- Sources, provenance, negatives and provisional members loaded from the catalog (**S9-14**).

### S9-14 — Evidence + reasoning in the auto-reconciler cache

The auto-reconciler now weighs the real evidence behind each record, stores what it did with each one, and has one name everywhere.

**What shipped**

- **One name: "auto-reconciler"** (decided while planning). Glossary in [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md) §1.
  - `core/resolve` → `core/autoreconcile` (`Reconcile`, `ReconciledValue`, `Result.Values` / `Outcomes`, `Outcome.Value`).
  - `core/database/resolvedvalues` → `core/database/autoreconciler`.
  - Migration **000040** renames `conclusion_resolved_values` / `_meta` → `auto_reconciler_values` / `_meta` (indexes recreated) and adds `auto_reconciler_outcomes`.
  - `PersonHeader.name_cluster_count` → `name_value_count` (wire-compatible), Swift `nameValueCount`.
  - Comments, the string catalog and living docs follow.
  - A Reconciliation Claim is laid over the auto-reconciled value by composers, never written into these tables.
- **Evidence loaded,** in the same one query per batch: each Observation's Source, polarity, Source credibility, transcription certainty, and claim confidence and status. Provisional members are passed in (reasoning only); rejected members aren't.
- **Outcomes cached:** one `auto_reconciler_outcomes` row per Observation, with its reason, the rank of its value and the negative that denied it. Rebuild-equals-upkeep compares them.
- **Upkeep:**
  - `RecomputeSourceTx` from `sourcecredibility.Upsert` when the grade is set or changed;
  - `RecomputeCitationTx` from `citations.Update` when certainty flips;
  - provisional members count as members for upkeep.
- `CacheVersion = 9`.
- **Tests:**
  - `TestReconciledEvidence` (8): low trust, uncertain transcription, one Source one vote, a stronger negative, low claim confidence, a misspelling outvoted across three Sources, provisional and rejected members, outcome rows. Blanking each loaded field (Source, negative, provisional, provenance) fails it.
  - Three scenarios: lowering credibility, a negative removed, certainty toggled.
  - Seeded sequences gain credibility, certainty, negative and graded-promote steps across two Sources, checked right after each edit. Replacing either hook with a no-op fails both the scenarios and the sequences.
  - The cache fixture makes Sources on demand. Tests that meant independent evidence now cite it from separate Sources.

**What changed for researchers**

- Two records from the same Source are one vote, not two.
- A spelling from a low-trust Source, an uncertain transcription or a low-confidence claim drops when better evidence disagrees, and changing a Source's credibility or a Citation's certainty updates Persons and Places straight away.
- A stronger record saying "not this" removes the value it denies.

**What stayed out**

- Showing the outcomes (*Why*): **S9-15** / **S9-16**.
- A claim-confidence or claim-status edit path (none exists; Promote sets confidence and recomputes).
- Laying Reconciliation Claims over auto-reconciled values (claims are a later spike).

### S9-15 — Detail composer + detail read

One read now gives a handle's page everything the auto-reconciler knows: each field's values and state, the values it didn't show and why, and what it did with every record. Nothing draws it yet; S9-16 does.

**What shipped**

- **Go composer** `core/database/conclusiondetails.ForEntity`, generic over kind:
  - one field per Property bound to the handle's type, in binding order (subject-valued Properties skipped until S9-28);
  - the field's state, computed by the `Result.State` rule;
  - every `auto_reconciler_values` row, with support, `against` and reason;
  - every `auto_reconciler_outcomes` row, with the record's own value, its Subject, Citation and Source, and the evidence the auto-reconciler weighed (credibility and confidence keys, and `autoreconcile.Provenance` relative to the default grades);
  - five queries per call whatever the field count (`TestForEntityQueryCountIsConstant`);
  - unknown, malformed or merged handles are `conclusiondetails.not_found`.
- **FFI** `GetConclusionDetail` (method 90). A `ConclusionValue` oneof carries text, integer, term, date or name. `runRPC` tests, plus a wire round-trip of every value kind.
- **Swift:**
  - `GenealogyStore.getConclusionDetail` → `CatalogConclusionDetail` (GoStore, and a FakeStore that composes a Person's name field from its members);
  - `CatalogQueryKey.conclusionDetail(project:, entityId:)`, staled by every Conclusion trigger;
  - the Person detail place loads it (the page is still the stub);
  - an `L10n.Errors` string for `not_found`.
- **`ReconciledValueDisplay`** words a field: state line (*1 Source*, *merged · 3 Sources*, *mixed*, *Nothing recorded*), *+N*, each value through the name and date displays, and each outcome (*folded into James Robins*, *weak · low-trust Source*, *denied by OBS-…*, …). New strings live in `L10n.Conclusions`.
- **Eviction, the rule-2 exception deferred from S9-07:**
  - `WorkspaceSession.visibleKeys` is the keys of the last `apply(location:)`;
  - a registry key tagged `evictWhenHidden` that isn't visible is evicted on a trigger (handle dropped, load cancelled);
  - `conclusionDetail` is tagged; the visible page and list keys revalidate as before;
  - [`macos-client-patterns.md`](../../macos-client-patterns.md) records the rule.
- **Tests:**
  - Go: the composer (merged, folded, outvoted, weak with its provenance, denied and against, `no_evidence`, empty field, binding order, not-found, query count) and the FFI table.
  - Swift: `ReconciledValueDisplayTests` (every state, plurals, *+N*, every outcome, each weak cause); the registry (the key on every trigger, only detail keys evict, FakeStore load and not-found); the Person place's key; the session (hidden page evicted, visible page and list revalidated, revisit reloads). Disabling eviction fails the session test.

**What changed for researchers**

- Nothing visible yet. Opening a Person now loads its detail in the background, and editing evidence no longer reloads every Person page visited earlier.

**What stayed out**

- The Person page itself: **S9-16**.
- Event and Place detail places: **S9-22** / **S9-25**.
- Laying Reconciliation Claims over values (claims are a later spike); the `concluded` state is worded but never sent yet.

### S9-D5 — Design: Person detail (revised for reasoning)

**Board:** Claude Design project *Person Detail* (`c8660cdb-2c61-4c30-bc39-b2411036e6c0`), `Person Detail.dc.html`, frames 1a–1i (rev. 2026-10-05).

- **Each field is one row:** label · lead value · state · support in Sources · disclosures. The state is always text (a *Merged* / *Mixed* / *Concluded* badge, or the Source count alone for a single value), never colour alone. A negative record that eliminated nothing still shows: *1 record disagrees*.
- **Mixed leads with the first surviving value;** the rest disclose (*1 other value*). There is no "choose" control: that is reconciliation, later.
- **Why lists every record considered**, as Read as · Source · Outcome. Each outcome is one mark and one phrase, readable without colour:

  | Outcome | Mark | Phrase |
  | --- | --- | --- |
  | kept | check | *kept* |
  | folded | git-merge | *folded into James Robins* |
  | outvoted | scale | *outvoted (2 of 3 Sources)* |
  | weak | signal-low | *weak · low-trust Source* (or uncertain transcription, low-confidence claim) |
  | denied | ban | *denied by Court deposition, 1862* |
  | against | circle-minus | *disagrees · did not eliminate* (a negative that denied something reads *kept*) |
  | no usable value | minus | *no usable value* |

- **Room for later:** the Why's empty last column is reserved for "conclude this value", and *Concluded* uses the same row columns, so neither needs a relayout (PD-8).
- **Header:** 80pt thumbnail slot with the person mark; title name → label → mono ref (the ref isn't repeated when it is the title); b. / d. shorthand saying *date unknown · place unknown* until the life Events exist; ref and member count at the right.
- **Ship state (frame 1g):** name filled, the four life rows stated empty, so the page keeps its shape when S9-32 fills them.
- **Member list (frame 1i)** is a stretch slot: deferred, see S9-16.

Brief archived: [`design/archive/S9-D5-person-detail.md`](design/archive/S9-D5-person-detail.md).

### S9-16 — Person detail

Opening a Person now shows their page: every field with its value, how the evidence agrees or disagrees, and, under *Why*, each record behind it and what happened to it.

**What shipped**

- **The vote behind *outvoted*:**
  - `autoreconcile.Outcome.Vote` records the winning value's Sources of every Source that voted on that unit (for a name, the part it lost on);
  - migration **000041** stores it (`vote_support`, `vote_total` on `auto_reconciler_outcomes`), and cache version **10** rebuilds on open.
- **The detail read:**
  - fields come from the handle's own cache rows (the Properties its records speak to), not from every Property its kind could have, so an unrecorded Property is never read;
  - each outcome carries its Artifact and vote;
  - the detail carries the accepted member count (one more query, still constant).
- **The page** (`Features/Conclusions/`):
  - `PersonDetailView` replaces the stub on `PlaceID.personDetail`;
  - the header follows the board;
  - under *Details*, Name and the four life rows always show, stated empty when nothing is recorded (the life rows until S9-32). Every other field (sex at birth, a custom birth weight) shows only once a record speaks to it, in binding order;
  - each field is a `ReconciledValueRow`, its Why a `ReconciliationReasoningView` of `ReconciliationOutcome`s;
  - a Why record's Source opens that Citation in the composer with its Observation in focus; Back returns to the Person.
  - `PersonDetailContent` and `ReconciledValueDisplay` hold all the wording and are unit-tested; the views only lay it out.
- **Kit:**
  - `PVDisclosureButton`: a `PVButton` whose parent owns the expanded state, the chevron and the spoken state. The caller places what it discloses.
  - `pvExpandedState(_:)` is its VoiceOver half, now also used by `PVSelect` and by the Source page's artifact rows, which expanded silently before. The strings moved to `designSystem.disclosure.*`.
  - `PVSymbol` gains the outcome and state marks.
- **Design-system review:** the disclosure was the one piece worth promoting. Sunken panels compose `PVCard`, the caps labels `pvMicroCaps()`, and the Source link `PVButton(.link)`. The row, the Why and the state badges stay snowflakes shared from `Features/Conclusions/` (one feature, three kinds).
- **Tests:**
  - Go:
    - the vote (pipeline, cache with rebuild-equals-upkeep, detail, FFI);
    - writing no vote fails the cache test.
  - Swift:
    - every board phrase, mark and spoken label (frames 1c–1h);
    - row order and empty text;
    - the header fallbacks;
    - the composer location parses as an edit of that Citation;
    - the disclosure's chevron, toggle and spoken state;
    - every new SF Symbol resolves on macOS 14.

**What changed for researchers**

- A Person's page shows each value with its state and Source count. *Why* shows every record behind it, including the spellings that lost (*outvoted (2 of 3 Sources)*), the weak ones and the denied ones. A click on a Source opens that record.
- VoiceOver reads each field's state in words, and says *expanded* or *collapsed* on disclosures, including a Source page's artifact rows.

**What stayed out**

- Life dates and places (rows and header): **S9-32**.
- The member list (frame 1i): needs a per-handle members read.
- Concluding a value (the reserved Why column, the *Concluded* badge): Reconciliation Claims, a later spike.
- Event and Place pages: **S9-24** / **S9-27**, reusing the row and the Why.

### S9-17 — Pins + backfill engine

Promote can now record which records a join was confirmed against, on both claims, with every pin audited. This is the engine only. It's salvaged from the first S9-17 / S9-18 (#265 / #266, closed unmerged) after the replan to Promote alignment, and nothing calls it from the app yet.

**What shipped**

- **`autoreconcile.Compatible`:** the pipeline's *same value* or *fold* for one pair of values. Every unit both carry must agree (*J. Robins* ~ *James Robins*; *Robins* / *Robbins* differ). A unit only one carries doesn't count against them; no evidence or no shared unit is never compatible. Promote alignment (S9-41) counts agreements with it.
- **Pins:**
  - `identityclaims.PinTx` writes one pin and audits it as `identity_claim_evidence` `create` under the claim's id, the mirror of the audited release. A pin the claim already has is skipped.
  - `PinnedObservations` reads a claim's pins.
- **Pairs and backfill in `promote.Save`:**
  - `Input.Pairs` takes confirmed pairs, on a join only;
  - each pair is checked in the transaction: the incoming Observation is the Subject's, the member's is on the same Property, and its Subject is an accepted member of the target;
  - both Observations are pinned on the new claim and on the member's claim, in the same `promote_subject` revision;
  - the member's `argument` is untouched;
  - `Result.Pins` counts the new claim's pins.

  Go only: there's no FFI or Swift path yet.
- **Pinned deletes, end to end:**
  - the delete-impact release fixture now gets its pins from Promote;
  - new tests: a pinned Observation's confirm names its handle once though two claims pin it;
  - the delete takes the pin off both claims and keeps their other pins;
  - each claim's audit replays as Promote's creates, then the delete's removal;
  - a member emptied of its records takes its claim's backfilled pins with it;
  - the cache equals a rebuild after each.
- **Retired:** the Swift `sameness_claim` delete-impact noun, its two catalog keys, and the preview built on it.

**What changed for researchers**

- Nothing visible yet.

**What stayed out**

- Pins one hop through a bridge, and the batch write: **S9-43**.
- Any UI, FFI or Swift pairs API, and FakeStore pins: **S9-43 / S9-44**.
- The per-Subject comparison read and the compare step from #265 were dropped; alignment (S9-41) and the evidence sheet (S9-44) replace them.
