-- Cross-resource FKs: Artifact, Citation, and property terms are resources.
-- SQLite cannot ALTER ON DELETE; rebuild those three tables.
-- Observation date/name values are exclusive (partial unique indexes).

PRAGMA foreign_keys=OFF;

CREATE TABLE artifacts__new (
	id          BLOB PRIMARY KEY,
	ref         TEXT UNIQUE NOT NULL,
	source_id   BLOB NOT NULL REFERENCES sources(id),
	file_id     BLOB REFERENCES files(id),
	label       TEXT NOT NULL DEFAULT '',
	description TEXT
) STRICT;

INSERT INTO artifacts__new (id, ref, source_id, file_id, label, description)
SELECT id, ref, source_id, file_id, label, description FROM artifacts;

DROP TABLE artifacts;
ALTER TABLE artifacts__new RENAME TO artifacts;

CREATE TABLE citations__new (
	id                        BLOB PRIMARY KEY,
	ref                       TEXT UNIQUE NOT NULL,
	artifact_id               BLOB NOT NULL REFERENCES artifacts(id),
	locator_json              TEXT NOT NULL,
	transcription             TEXT,
	description               TEXT,
	transcription_uncertain   INTEGER NOT NULL DEFAULT 0,
	transcription_note        TEXT,

	CHECK (transcription_uncertain IN (0, 1))
) STRICT;

INSERT INTO citations__new (
	id, ref, artifact_id, locator_json, transcription, description,
	transcription_uncertain, transcription_note
)
SELECT
	id, ref, artifact_id, locator_json, transcription, description,
	transcription_uncertain, transcription_note
FROM citations;

DROP TABLE citations;
ALTER TABLE citations__new RENAME TO citations;

CREATE TABLE property_terms__new (
	id           BLOB PRIMARY KEY,
	property_id  BLOB NOT NULL REFERENCES properties(id),
	key          TEXT NOT NULL,
	origin       TEXT NOT NULL,
	label        TEXT NOT NULL,
	description  TEXT,

	UNIQUE (property_id, key, origin)
) STRICT;

INSERT INTO property_terms__new (id, property_id, key, origin, label, description)
SELECT id, property_id, key, origin, label, description FROM property_terms;

DROP TABLE property_terms;
ALTER TABLE property_terms__new RENAME TO property_terms;

CREATE UNIQUE INDEX observations_value_date_id_uidx
	ON observations(value_date_id) WHERE value_date_id IS NOT NULL;
CREATE UNIQUE INDEX observations_value_name_id_uidx
	ON observations(value_name_id) WHERE value_name_id IS NOT NULL;

PRAGMA foreign_key_check;
PRAGMA foreign_keys=ON;
