-- Require a human-readable Source title (list / omnibar headline).
-- Backfill any NULL or blank titles before enforcing NOT NULL.
-- SQLite cannot ALTER COLUMN; recreate the table with FKs off.

-- migrate: foreign_keys off

UPDATE sources
SET title = 'Untitled'
WHERE title IS NULL OR TRIM(title) = '';

CREATE TABLE sources__new (
	id              BLOB PRIMARY KEY,
	ref             TEXT UNIQUE NOT NULL,
	source_type_id  BLOB NOT NULL REFERENCES source_types(id),
	title           TEXT NOT NULL,
	description     TEXT
) STRICT;

INSERT INTO sources__new (id, ref, source_type_id, title, description)
SELECT id, ref, source_type_id, title, description FROM sources;

DROP TABLE sources;
ALTER TABLE sources__new RENAME TO sources;
