-- Catalog omnibar FTS5 projection (navigable roots + rolled Source body text).
-- External-content FTS: Go projectors keep docs and fts in sync on writes;
-- search.EnsureIndex rebuilds when projection_version lags.

CREATE TABLE catalog_search_docs (
	rowid INTEGER PRIMARY KEY,
	kind TEXT NOT NULL,
	entity_id TEXT NOT NULL,
	display_ref TEXT NOT NULL DEFAULT '',
	display_title TEXT NOT NULL,
	display_subtitle TEXT NOT NULL DEFAULT '',
	title TEXT NOT NULL DEFAULT '',
	ref TEXT NOT NULL DEFAULT '',
	secondary TEXT NOT NULL DEFAULT '',
	body TEXT NOT NULL DEFAULT '',
	UNIQUE (kind, entity_id)
) STRICT;

CREATE VIRTUAL TABLE catalog_search_fts USING fts5(
	title,
	ref,
	secondary,
	body,
	content='catalog_search_docs',
	content_rowid='rowid',
	tokenize='unicode61 remove_diacritics 2',
	prefix='2 3'
);

CREATE TABLE catalog_search_meta (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	projection_version INTEGER NOT NULL
) STRICT;

INSERT INTO catalog_search_meta (id, projection_version) VALUES (1, 0);
