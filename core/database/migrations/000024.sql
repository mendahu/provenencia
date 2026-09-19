-- Property terms + value_type = term.
-- Seeded term rows come from subjectvocab.Install at create time (not SQL INSERTs).
-- SQLite cannot ALTER CHECK; recreate properties with the expanded value_type set.

PRAGMA foreign_keys=OFF;

CREATE TABLE properties__new (
	id          BLOB PRIMARY KEY,
	key         TEXT NOT NULL,
	origin      TEXT NOT NULL,
	label       TEXT NOT NULL,
	description TEXT,
	value_type  TEXT NOT NULL,

	UNIQUE (key, origin),
	CHECK (value_type IN (
		'text',
		'integer',
		'date',
		'name',
		'subject',
		'term'
	))
) STRICT;

INSERT INTO properties__new (id, key, origin, label, description, value_type)
SELECT id, key, origin, label, description, value_type FROM properties;

DROP TABLE properties;
ALTER TABLE properties__new RENAME TO properties;

CREATE TABLE property_terms (
	id           BLOB PRIMARY KEY,
	property_id  BLOB NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
	key          TEXT NOT NULL,
	origin       TEXT NOT NULL,
	label        TEXT NOT NULL,
	description  TEXT,

	UNIQUE (property_id, key, origin)
) STRICT;

PRAGMA foreign_keys=ON;
