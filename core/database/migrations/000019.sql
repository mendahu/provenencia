-- Trigram FTS for omnibar typo shortlist (S3-11). Same external content as
-- catalog_search_fts; Go projectors keep both indexes in sync on writes.

CREATE VIRTUAL TABLE catalog_search_fts_trigram USING fts5(
	title,
	ref,
	secondary,
	body,
	content='catalog_search_docs',
	content_rowid='rowid',
	tokenize='trigram case_sensitive 0 remove_diacritics 1'
);
