-- Pins no longer block Observation deletes. Deleting a pinned Observation
-- weakens the claims that pinned it (§5.2 review), it does not refuse.
-- Go deletes pins explicitly so each removal is audited; this CASCADE is
-- only the backstop. SQLite cannot ALTER ON DELETE; rebuild the table.

PRAGMA foreign_keys=OFF;

CREATE TABLE identity_claim_evidence__new (
	identity_claim_id BLOB NOT NULL REFERENCES identity_claims(id) ON DELETE CASCADE,
	observation_id    BLOB NOT NULL REFERENCES observations(id) ON DELETE CASCADE,

	PRIMARY KEY (identity_claim_id, observation_id)
) STRICT;

INSERT INTO identity_claim_evidence__new (identity_claim_id, observation_id)
SELECT identity_claim_id, observation_id FROM identity_claim_evidence;

DROP TABLE identity_claim_evidence;
ALTER TABLE identity_claim_evidence__new RENAME TO identity_claim_evidence;

CREATE INDEX identity_claim_evidence_observation_id_idx ON identity_claim_evidence(observation_id);

PRAGMA foreign_key_check;
PRAGMA foreign_keys=ON;
