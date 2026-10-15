-- canonical_entity audit scopes: when work on a Person, Event, or Place last
-- happened (S10-21). audit.Record stores them from effects.HandleScopes.
-- This backfill mirrors those paths for older history, best effort: a
-- change's own fields first, then live rows, then other recorded changes of
-- the same row. Subjects map to handles through today's accepted and
-- provisional claims; membership at the time of the write is not recorded.

-- claim id -> handle id, from the live claim and every recorded entity_id.
CREATE TEMP TABLE claim_handle AS
SELECT id AS claim_id, entity_id FROM identity_claims
UNION
SELECT entity_id, unhex(replace(json_extract(changes_json, '$.entity_id.old'), '-', ''))
	FROM audit_changes WHERE entity_type = 'identity_claim'
UNION
SELECT entity_id, unhex(replace(json_extract(changes_json, '$.entity_id.new'), '-', ''))
	FROM audit_changes WHERE entity_type = 'identity_claim';

-- observation id -> subject id, from the live row and every recorded subject_id.
CREATE TEMP TABLE observation_subject AS
SELECT id AS observation_id, subject_id FROM observations
UNION
SELECT entity_id, unhex(replace(json_extract(changes_json, '$.subject_id.old'), '-', ''))
	FROM audit_changes WHERE entity_type = 'observation'
UNION
SELECT entity_id, unhex(replace(json_extract(changes_json, '$.subject_id.new'), '-', ''))
	FROM audit_changes WHERE entity_type = 'observation';

-- revision -> subject touched by evidence on it.
CREATE TEMP TABLE change_subject AS
SELECT audit_transaction_id AS tx_id,
	unhex(replace(json_extract(changes_json, '$.subject_id.old'), '-', '')) AS subject_id
	FROM audit_changes
	WHERE entity_type = 'observation' AND json_type(changes_json, '$.subject_id') IS NOT NULL
UNION
SELECT audit_transaction_id, unhex(replace(json_extract(changes_json, '$.subject_id.new'), '-', ''))
	FROM audit_changes
	WHERE entity_type = 'observation' AND json_type(changes_json, '$.subject_id') IS NOT NULL
UNION
SELECT c.audit_transaction_id, os.subject_id FROM audit_changes c
	JOIN observation_subject os ON os.observation_id = c.entity_id
	WHERE c.entity_type = 'observation' AND json_type(c.changes_json, '$.subject_id') IS NULL
UNION
SELECT c.audit_transaction_id, os.subject_id FROM audit_changes c
	JOIN observation_subject os ON os.observation_id =
		unhex(replace(json_extract(c.changes_json, '$.observation_id.old'), '-', ''))
	WHERE c.entity_type = 'observation_note'
UNION
SELECT c.audit_transaction_id, os.subject_id FROM audit_changes c
	JOIN observation_subject os ON os.observation_id =
		unhex(replace(json_extract(c.changes_json, '$.observation_id.new'), '-', ''))
	WHERE c.entity_type = 'observation_note'
UNION
SELECT c.audit_transaction_id, os.subject_id FROM audit_changes c
	JOIN observation_notes n ON n.id = c.entity_id
	JOIN observation_subject os ON os.observation_id = n.observation_id
	WHERE c.entity_type = 'observation_note'
UNION
SELECT c.audit_transaction_id, o.subject_id FROM audit_changes c
	JOIN observations o ON o.value_date_id = c.entity_id
	WHERE c.entity_type = 'date_value'
UNION
SELECT c.audit_transaction_id, o.subject_id FROM audit_changes c
	JOIN observations o ON o.value_name_id = c.entity_id
	WHERE c.entity_type = 'name_value';

INSERT OR IGNORE INTO audit_transaction_scopes (audit_transaction_id, scope_type, scope_id)
SELECT tx_id, 'canonical_entity', entity_id FROM (
	SELECT audit_transaction_id AS tx_id, entity_id FROM audit_changes
		WHERE entity_type = 'canonical_entity'
	UNION
	SELECT audit_transaction_id, unhex(replace(json_extract(changes_json, '$.entity_id.old'), '-', ''))
		FROM audit_changes
		WHERE entity_type = 'identity_claim' AND json_type(changes_json, '$.entity_id') IS NOT NULL
	UNION
	SELECT audit_transaction_id, unhex(replace(json_extract(changes_json, '$.entity_id.new'), '-', ''))
		FROM audit_changes
		WHERE entity_type = 'identity_claim' AND json_type(changes_json, '$.entity_id') IS NOT NULL
	UNION
	SELECT c.audit_transaction_id, ch.entity_id FROM audit_changes c
		JOIN claim_handle ch ON ch.claim_id = c.entity_id
		WHERE (c.entity_type = 'identity_claim' AND json_type(c.changes_json, '$.entity_id') IS NULL)
			OR c.entity_type = 'identity_claim_evidence'
	UNION
	SELECT cs.tx_id, ic.entity_id FROM change_subject cs
		JOIN identity_claims ic ON ic.subject_id = cs.subject_id
		WHERE ic.status IN ('accepted', 'provisional')
)
WHERE entity_id IS NOT NULL AND length(entity_id) = 16;

DROP TABLE change_subject;
DROP TABLE observation_subject;
DROP TABLE claim_handle;
