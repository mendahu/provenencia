-- S14-21: the words of a search document that SQLite's tokenizer indexes
-- differently from rescoring's key (letters remove_diacritics cannot fold, and
-- words it splits at punctuation). Written by searchindex.Upsert for every
-- kind; ProjectionVersion 10 fills it on open. An FTS5 table cannot add a
-- column, so catalog_search_fts is rebuilt empty here and RebuildAll refills it.

ALTER TABLE catalog_search_docs ADD COLUMN folded TEXT NOT NULL DEFAULT '';

DROP TABLE catalog_search_fts;
CREATE VIRTUAL TABLE catalog_search_fts USING fts5(
	title,
	ref,
	secondary,
	body,
	folded,
	content='catalog_search_docs',
	content_rowid='rowid',
	tokenize='unicode61 remove_diacritics 2',
	prefix='2 3'
);
