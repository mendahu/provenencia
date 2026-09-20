-- Source metadata field data_type = url (text-shaped; values stay in value_text).
-- SQLite cannot ALTER CHECK; recreate source_metadata_fields with the expanded set.

PRAGMA foreign_keys=OFF;

CREATE TABLE source_metadata_fields__new (
	id          BLOB PRIMARY KEY,
	key         TEXT NOT NULL,
	origin      TEXT NOT NULL,
	label       TEXT NOT NULL,
	data_type   TEXT NOT NULL DEFAULT 'text',
	description TEXT,

	UNIQUE (key, origin),
	CHECK (data_type IN ('text', 'date', 'url'))
) STRICT;

INSERT INTO source_metadata_fields__new (id, key, origin, label, data_type, description)
SELECT id, key, origin, label, data_type, description FROM source_metadata_fields;

DROP TABLE source_metadata_fields;
ALTER TABLE source_metadata_fields__new RENAME TO source_metadata_fields;

PRAGMA foreign_keys=ON;
