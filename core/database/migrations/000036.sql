-- Audit scopes: which aggregates a revision touched, so "when did anything
-- under X last change" is one indexed MAX(revision). audit.Record derives the
-- rows from each change through per-entity resolvers (audit/scopes.go).
-- scope_id is polymorphic (like audit_changes.entity_id), so it has no FK.
-- Only scope_type 'source' is written today.

CREATE TABLE audit_transaction_scopes (
	audit_transaction_id BLOB NOT NULL REFERENCES audit_transactions(id),
	scope_type           TEXT NOT NULL,
	scope_id             BLOB NOT NULL,

	PRIMARY KEY (scope_type, scope_id, audit_transaction_id)
) STRICT;

CREATE INDEX audit_transaction_scopes_transaction_id_idx
	ON audit_transaction_scopes(audit_transaction_id);

-- Best-effort backfill mirroring the resolvers. Parent ids come from the
-- change's own fields (old or new) or from live rows; a delete whose whole
-- parent chain is gone resolves to nothing and that history stays unscoped.
-- UUIDs in changes_json are dashed strings; unhex() of a non-UUID is NULL.

CREATE TEMP TABLE scope_fields AS
SELECT
	c.audit_transaction_id AS tx_id,
	c.entity_type          AS entity_type,
	c.entity_id            AS entity_id,
	unhex(replace(json_extract(c.changes_json, '$.' || k.key || '.old'), '-', '')) AS old_id,
	unhex(replace(json_extract(c.changes_json, '$.' || k.key || '.new'), '-', '')) AS new_id
FROM audit_changes c
JOIN (
	SELECT 'source_note' AS entity_type, 'source_id' AS key
	UNION ALL SELECT 'source_metadata', 'source_id'
	UNION ALL SELECT 'source_metadata_layout', 'source_id'
	UNION ALL SELECT 'source_credibility_assessment', 'source_id'
	UNION ALL SELECT 'subject', 'source_id'
	UNION ALL SELECT 'artifact', 'source_id'
	UNION ALL SELECT 'citation', 'artifact_id'
	UNION ALL SELECT 'citation_note', 'citation_id'
	UNION ALL SELECT 'observation', 'citation_id'
	UNION ALL SELECT 'observation_note', 'observation_id'
) k ON k.entity_type = c.entity_type;

CREATE TEMP TABLE scope_parents AS
SELECT tx_id, entity_type, old_id AS parent_id FROM scope_fields WHERE old_id IS NOT NULL
UNION
SELECT tx_id, entity_type, new_id FROM scope_fields WHERE new_id IS NOT NULL
UNION
SELECT c.audit_transaction_id, c.entity_type, n.source_id
	FROM audit_changes c JOIN source_notes n ON n.id = c.entity_id
	WHERE c.entity_type = 'source_note'
UNION
SELECT c.audit_transaction_id, c.entity_type, m.source_id
	FROM audit_changes c JOIN source_metadata m ON m.id = c.entity_id
	WHERE c.entity_type = 'source_metadata'
UNION
SELECT c.audit_transaction_id, c.entity_type, a.source_id
	FROM audit_changes c JOIN source_credibility_assessments a ON a.id = c.entity_id
	WHERE c.entity_type = 'source_credibility_assessment'
UNION
SELECT c.audit_transaction_id, c.entity_type, s.source_id
	FROM audit_changes c JOIN subjects s ON s.id = c.entity_id
	WHERE c.entity_type = 'subject'
UNION
SELECT c.audit_transaction_id, c.entity_type, a.source_id
	FROM audit_changes c JOIN artifacts a ON a.id = c.entity_id
	WHERE c.entity_type = 'artifact'
UNION
SELECT c.audit_transaction_id, c.entity_type, ci.artifact_id
	FROM audit_changes c JOIN citations ci ON ci.id = c.entity_id
	WHERE c.entity_type = 'citation'
UNION
SELECT c.audit_transaction_id, c.entity_type, n.citation_id
	FROM audit_changes c JOIN citation_notes n ON n.id = c.entity_id
	WHERE c.entity_type = 'citation_note'
UNION
SELECT c.audit_transaction_id, c.entity_type, o.citation_id
	FROM audit_changes c JOIN observations o ON o.id = c.entity_id
	WHERE c.entity_type = 'observation'
UNION
SELECT c.audit_transaction_id, c.entity_type, n.observation_id
	FROM audit_changes c JOIN observation_notes n ON n.id = c.entity_id
	WHERE c.entity_type = 'observation_note';

INSERT OR IGNORE INTO audit_transaction_scopes (audit_transaction_id, scope_type, scope_id)
SELECT tx_id, 'source', source_id FROM (
	SELECT c.audit_transaction_id AS tx_id, c.entity_id AS source_id
		FROM audit_changes c WHERE c.entity_type = 'source'
	UNION
	SELECT p.tx_id, p.parent_id FROM scope_parents p
		WHERE p.entity_type IN ('source_note', 'source_metadata', 'source_metadata_layout',
			'source_credibility_assessment', 'subject', 'artifact')
	UNION
	SELECT p.tx_id, a.source_id FROM scope_parents p
		JOIN artifacts a ON a.id = p.parent_id
		WHERE p.entity_type = 'citation'
	UNION
	SELECT p.tx_id, a.source_id FROM scope_parents p
		JOIN citations ci ON ci.id = p.parent_id
		JOIN artifacts a ON a.id = ci.artifact_id
		WHERE p.entity_type IN ('citation_note', 'observation')
	UNION
	SELECT p.tx_id, a.source_id FROM scope_parents p
		JOIN observations o ON o.id = p.parent_id
		JOIN citations ci ON ci.id = o.citation_id
		JOIN artifacts a ON a.id = ci.artifact_id
		WHERE p.entity_type = 'observation_note'
	UNION
	SELECT c.audit_transaction_id, a.source_id FROM audit_changes c
		JOIN observations o ON o.value_date_id = c.entity_id
		JOIN citations ci ON ci.id = o.citation_id
		JOIN artifacts a ON a.id = ci.artifact_id
		WHERE c.entity_type = 'date_value'
	UNION
	SELECT c.audit_transaction_id, a.source_id FROM audit_changes c
		JOIN observations o ON o.value_name_id = c.entity_id
		JOIN citations ci ON ci.id = o.citation_id
		JOIN artifacts a ON a.id = ci.artifact_id
		WHERE c.entity_type = 'name_value'
	UNION
	SELECT c.audit_transaction_id, a.source_id FROM audit_changes c
		JOIN artifacts a ON a.file_id = c.entity_id
		WHERE c.entity_type = 'file'
)
WHERE length(source_id) = 16;

DROP TABLE scope_parents;
DROP TABLE scope_fields;
