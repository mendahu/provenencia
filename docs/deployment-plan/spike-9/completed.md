# Spike 9 — Completed steps

Finished Spike 9 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S9-NN`, `S9-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| S9-01 | PR | Conclusion schema + stores |

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
