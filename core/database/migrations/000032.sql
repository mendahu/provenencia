-- Conclusion layer (Spike 9): canonical handles, Identity Claims, and their
-- Observation pins. Claim confidence grades are seeded by
-- claimconfidencegrades.Install at create time (not SQL INSERTs).
-- Reconciliation Claims and canonical_entity_notes are later.

CREATE TABLE claim_confidence_grades (
	id         BLOB PRIMARY KEY,
	key        TEXT NOT NULL,
	origin     TEXT NOT NULL,
	label      TEXT NOT NULL,
	sort_order INTEGER NOT NULL,
	UNIQUE (key, origin)
) STRICT;

CREATE TABLE canonical_entities (
	id              BLOB PRIMARY KEY,
	subject_type_id BLOB NOT NULL REFERENCES subject_types(id),
	ref             TEXT UNIQUE NOT NULL,
	argument        TEXT,
	label           TEXT,
	merged_into_id  BLOB REFERENCES canonical_entities(id),

	UNIQUE (id, subject_type_id)
) STRICT;

-- subject_type_id is copied from the subject so the composite FKs reject a
-- subject claimed onto a handle of another type.
CREATE TABLE identity_claims (
	id                  BLOB PRIMARY KEY,
	subject_id          BLOB NOT NULL,
	entity_id           BLOB NOT NULL,
	subject_type_id     BLOB NOT NULL,
	status              TEXT NOT NULL,
	confidence_grade_id BLOB REFERENCES claim_confidence_grades(id),
	argument            TEXT,

	CHECK (status IN ('provisional', 'accepted', 'rejected')),
	FOREIGN KEY (subject_id, subject_type_id)
		REFERENCES subjects (id, subject_type_id)
		ON DELETE CASCADE,
	FOREIGN KEY (entity_id, subject_type_id)
		REFERENCES canonical_entities (id, subject_type_id)
		ON DELETE CASCADE,
	UNIQUE (subject_id, entity_id)
) STRICT;

CREATE UNIQUE INDEX identity_claims_one_accepted_per_subject
	ON identity_claims (subject_id)
	WHERE status = 'accepted';

CREATE TABLE identity_claim_evidence (
	identity_claim_id BLOB NOT NULL REFERENCES identity_claims(id) ON DELETE CASCADE,
	observation_id    BLOB NOT NULL REFERENCES observations(id),

	PRIMARY KEY (identity_claim_id, observation_id)
) STRICT;

CREATE INDEX canonical_entities_subject_type_id_idx ON canonical_entities(subject_type_id);
CREATE INDEX canonical_entities_merged_into_id_idx ON canonical_entities(merged_into_id);
CREATE INDEX identity_claims_entity_id_idx ON identity_claims(entity_id, status);
CREATE INDEX identity_claims_confidence_grade_id_idx ON identity_claims(confidence_grade_id);
CREATE INDEX identity_claim_evidence_observation_id_idx ON identity_claim_evidence(observation_id);
