-- Interpretation Property vocabulary and Subject-type bindings.
-- Seeded rows come from subjectvocab.Install at create time (not SQL INSERTs).
-- Locked bindings / capabilities live in the compiled registry, not columns here.

CREATE TABLE properties (
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
		'subject'
	))
) STRICT;

CREATE TABLE subject_type_fields (
	subject_type_id BLOB NOT NULL REFERENCES subject_types(id) ON DELETE CASCADE,
	property_id     BLOB NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
	sort_order      INTEGER NOT NULL,

	PRIMARY KEY (subject_type_id, property_id)
) STRICT;
