# Spike 9 — Completed steps

Finished Spike 9 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S9-NN`, `S9-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| S9-01 | PR | Conclusion schema + stores |
| S9-02 | PR | Promote write v1 |
| S9-03 | PR | Source graph carries membership |

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

- `identityclaims.MembershipsBySource`: one indexed query (Source's Subjects → accepted claim → handle → type key). Unpromoted Subjects have no row.
- FFI `METHOD_LIST_SUBJECT_MEMBERSHIPS` (`SubjectMembership { subject_id, CanonicalEntity entity, kind }`).
- Swift `listSubjectMemberships` on `GenealogyStore` / `GoStore` / `FakeStore`, loaded as the fourth parallel read of the `sourceGraph` key. `CatalogSubject` stays Interpretation-only.
- `SourceGraphRows.memberships` → `SourceGraphPlacedSubject.membership` / `SourceGraphPlacedBridge.membership` (`nil` = unpromoted). The membership survives position patches.
- `CatalogMutation.promotedSubject(sourceId:)` invalidates exactly that Source's graph key.

**What stayed out**

- The card's Promote control, membership row, and dispatching `.promotedSubject` (**S9-04**)
- The handle's resolved name on the card (**S9-09**)
- Busting other Conclusion keys on Promote (lists arrive in **S9-07**)
