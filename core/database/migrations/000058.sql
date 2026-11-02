-- S13-25: Identity Claims are membership only. Drops the Observation pins,
-- the claim confidence grade and argument columns, and the grades table.
-- Audit rows already written for pins and for the two columns are kept.
-- migrate: foreign_keys off

DROP TABLE identity_claim_evidence;

CREATE TABLE identity_claims__new (
	id              BLOB PRIMARY KEY,
	subject_id      BLOB NOT NULL,
	entity_id       BLOB NOT NULL,
	subject_type_id BLOB NOT NULL,
	status          TEXT NOT NULL,

	CHECK (status IN ('provisional', 'accepted', 'rejected')),
	FOREIGN KEY (subject_id, subject_type_id)
		REFERENCES subjects (id, subject_type_id)
		ON DELETE CASCADE,
	FOREIGN KEY (entity_id, subject_type_id)
		REFERENCES canonical_entities (id, subject_type_id)
		ON DELETE CASCADE,
	UNIQUE (subject_id, entity_id)
) STRICT;
INSERT INTO identity_claims__new (id, subject_id, entity_id, subject_type_id, status)
SELECT id, subject_id, entity_id, subject_type_id, status FROM identity_claims;
DROP TABLE identity_claims;
ALTER TABLE identity_claims__new RENAME TO identity_claims;

CREATE UNIQUE INDEX identity_claims_one_accepted_per_subject
	ON identity_claims (subject_id)
	WHERE status = 'accepted';
CREATE INDEX identity_claims_entity_id_idx ON identity_claims(entity_id, status);

DROP TABLE claim_confidence_grades;
