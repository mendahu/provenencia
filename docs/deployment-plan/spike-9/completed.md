# Spike 9 — Completed steps

Finished Spike 9 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S9-NN`, `S9-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| S9-01 | PR | Conclusion schema + stores |
| S9-02 | PR | Promote write v1 |

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

Shipped the first callable Conclusion write: mint a handle from a Subject and file an accepted Identity Claim, one transaction, over FFI and from Swift. Subject and Observation deletes now name the handles they affect, never refuse because of Conclusion, and audit what they remove.

**What shipped**

- `core/database/promote`: `Save` mints a handle of the Subject's type plus an accepted claim (zero pins), audited as one `promote_subject` revision. Refuses members (`identityclaims.already_member`) and, in v1, non-primary kinds (`promote.unsupported_type`)
- FFI `METHOD_PROMOTE_SUBJECT` (`PromoteSubjectRequest` → `CanonicalEntity` + `IdentityClaim`); Swift `promoteSubject` on `GenealogyStore` / `GoStore` / `FakeStore` (`membershipBySubject` for S9-03 / S9-04)
- Delete Impact `cascades`: a non-blocking list (`cascadeEdges`) through Go, proto and Swift. A promoted Subject reports `identity_claims.subject_id` → its handle; the DeleteImpact confirm appends "It will no longer belong to PER-…. That record stays."
- L10n: error copy for the identity-claim, canonical-entity and promote codes; `canonical_entity` Impact kind noun
- **Pins never block deletes.** Migration `000033` rebuilds `identity_claim_evidence` with `observation_id … ON DELETE CASCADE` (backstop only). `observations.Delete` removes the Observation's pins and `subjects.Delete` removes its claims and the pins on its Observations **explicitly**, audited in the same revision (`identityclaims.ReleaseObservationPinsTx` / `ReleaseSubjectTx`). A pinned Observation's confirm says it leaves the evidence for PER-…; the S9-01 blocking evidence probe is gone
- Future UI: weak-claim review alert written into model §5.2 and [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md) (Spike 10)

**What stayed out**

- Existing targets and suggestions (**S9-10**); pins and backfill (**S9-17**)
- Resolved-values cache and upkeep on claim create / Subject delete (**S9-06**)
- Graph membership read (**S9-03**) and the card's Promote control (**S9-04**)
- Bridge filing (**S9-28**); the Swift `sameness_claim` leftovers (**S9-18**)
