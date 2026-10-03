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
