-- Catalog metadata is filing text only: drop date_value_id and data_type=date.
-- Flatten any structured DateValue into value_text (phrase, else year).
-- SQLite cannot ALTER CHECK or DROP a column in place; recreate both tables.

PRAGMA foreign_keys=OFF;

UPDATE source_metadata
SET value_text = (
	SELECT COALESCE(
		NULLIF(TRIM(source_metadata.value_text), ''),
		NULLIF(TRIM(d.phrase), ''),
		CAST(d.start_year AS TEXT),
		''
	)
	FROM date_values d
	WHERE d.id = source_metadata.date_value_id
)
WHERE date_value_id IS NOT NULL
	AND (value_text IS NULL OR TRIM(value_text) = '');

CREATE TEMP TABLE metadata_date_ids AS
SELECT date_value_id AS id FROM source_metadata WHERE date_value_id IS NOT NULL;

CREATE TABLE source_metadata__new (
	id         BLOB PRIMARY KEY,
	source_id  BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
	field_id   BLOB NOT NULL REFERENCES source_metadata_fields(id),
	value_text TEXT,
	UNIQUE (source_id, field_id)
) STRICT;

INSERT INTO source_metadata__new (id, source_id, field_id, value_text)
SELECT id, source_id, field_id, value_text FROM source_metadata;

DROP TABLE source_metadata;
ALTER TABLE source_metadata__new RENAME TO source_metadata;

CREATE TABLE source_metadata_fields__new (
	id          BLOB PRIMARY KEY,
	key         TEXT NOT NULL,
	origin      TEXT NOT NULL,
	label       TEXT NOT NULL,
	data_type   TEXT NOT NULL DEFAULT 'text',
	description TEXT,

	UNIQUE (key, origin),
	CHECK (data_type IN ('text', 'url'))
) STRICT;

INSERT INTO source_metadata_fields__new (id, key, origin, label, data_type, description)
SELECT id, key, origin, label,
	CASE WHEN data_type = 'date' THEN 'text' ELSE data_type END,
	description
FROM source_metadata_fields;

DROP TABLE source_metadata_fields;
ALTER TABLE source_metadata_fields__new RENAME TO source_metadata_fields;

DELETE FROM date_values
WHERE id IN (SELECT id FROM metadata_date_ids)
	AND id NOT IN (
		SELECT value_date_id FROM observations WHERE value_date_id IS NOT NULL
	);

DROP TABLE metadata_date_ids;

PRAGMA foreign_keys=ON;
