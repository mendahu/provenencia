-- Name format profiles (structured-name-model §4). Rows are create-time
-- Install, not SQL seeds. Existing catalogs get the empty tables only.

CREATE TABLE name_format_profiles (
	id          BLOB PRIMARY KEY,
	key         TEXT NOT NULL,
	origin      TEXT NOT NULL,
	label       TEXT NOT NULL,
	description TEXT,
	UNIQUE (key, origin)
) STRICT;

CREATE TABLE name_format_profile_parts (
	profile_id         BLOB NOT NULL REFERENCES name_format_profiles(id) ON DELETE CASCADE,
	style              TEXT NOT NULL,
	idx                INTEGER NOT NULL,
	part_type          TEXT NOT NULL,
	trailing_separator TEXT NOT NULL DEFAULT '',
	PRIMARY KEY (profile_id, style, idx),
	UNIQUE (profile_id, style, part_type),
	CHECK (idx >= 0),
	CHECK (style IN ('natural', 'sorted'))
) STRICT;

CREATE TABLE name_format_profile_roles (
	profile_id BLOB NOT NULL REFERENCES name_format_profiles(id) ON DELETE CASCADE,
	part_type  TEXT NOT NULL,
	role       TEXT NOT NULL,
	PRIMARY KEY (profile_id, part_type)
) STRICT;

CREATE TABLE name_format_profile_weights (
	profile_id BLOB NOT NULL REFERENCES name_format_profiles(id) ON DELETE CASCADE,
	slot       TEXT NOT NULL,
	weight     REAL NOT NULL,
	PRIMARY KEY (profile_id, slot)
) STRICT;

CREATE TABLE project_settings (
	key        TEXT PRIMARY KEY,
	value_text TEXT
) STRICT;
