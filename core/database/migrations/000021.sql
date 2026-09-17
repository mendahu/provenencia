-- Interpretation subject vocabulary, Subjects, and Source-scoped graph layout.
-- subject_types / subjects are research domain rows (audit via audit_* tables).
-- subject_positions is unaudited UI layout state: arranging bubbles is not a
-- research assertion (contrast source_metadata_layout, which is audited).

CREATE TABLE subject_types (
	id                   BLOB PRIMARY KEY,
	key                  TEXT NOT NULL,
	origin               TEXT NOT NULL,
	label                TEXT NOT NULL,
	description          TEXT,
	ref_prefix           TEXT NOT NULL UNIQUE,
	candidate_ref_prefix TEXT NOT NULL UNIQUE,

	UNIQUE (key, origin)
) STRICT;

CREATE TABLE subjects (
	id              BLOB PRIMARY KEY,
	ref             TEXT UNIQUE NOT NULL,
	source_id       BLOB NOT NULL REFERENCES sources(id),
	subject_type_id BLOB NOT NULL REFERENCES subject_types(id),
	label           TEXT,
	description     TEXT,

	UNIQUE (id, subject_type_id)
) STRICT;

CREATE TABLE subject_positions (
	source_id  BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
	subject_id BLOB NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
	grid_x     INTEGER NOT NULL,
	grid_y     INTEGER NOT NULL,

	PRIMARY KEY (source_id, subject_id)
) STRICT;
