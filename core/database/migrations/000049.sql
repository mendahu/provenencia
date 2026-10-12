-- Abbreviation and nickname tables. Rows are create-time Install, not SQL
-- seeds. Existing catalogs get the empty tables only.

CREATE TABLE name_alias_tables (
	id     BLOB PRIMARY KEY,
	key    TEXT NOT NULL,
	origin TEXT NOT NULL,
	label  TEXT NOT NULL,
	kind   TEXT NOT NULL,
	UNIQUE (key, origin),
	CHECK (kind IN ('abbreviation', 'nickname'))
) STRICT;

CREATE TABLE name_alias_members (
	table_id BLOB NOT NULL REFERENCES name_alias_tables(id) ON DELETE CASCADE,
	set_idx  INTEGER NOT NULL,
	token    TEXT NOT NULL,
	PRIMARY KEY (table_id, set_idx, token),
	CHECK (set_idx >= 0),
	CHECK (token <> '')
) STRICT;

CREATE TABLE name_alias_roles (
	table_id BLOB NOT NULL REFERENCES name_alias_tables(id) ON DELETE CASCADE,
	role     TEXT NOT NULL,
	PRIMARY KEY (table_id, role),
	CHECK (role IN ('family', 'given', 'nick', 'untyped', 'generation', 'ignored'))
) STRICT;
