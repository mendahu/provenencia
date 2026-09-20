CREATE TABLE name_values (
	id   BLOB PRIMARY KEY,
	form TEXT NOT NULL
) STRICT;

CREATE TABLE name_value_parts (
	id            BLOB PRIMARY KEY,
	name_value_id BLOB NOT NULL REFERENCES name_values(id) ON DELETE CASCADE,
	idx           INTEGER NOT NULL,
	value         TEXT NOT NULL,
	type          TEXT,

	UNIQUE (name_value_id, idx),
	CHECK (idx >= 0)
) STRICT;
