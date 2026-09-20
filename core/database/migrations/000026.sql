CREATE TABLE citations (
	id                        BLOB PRIMARY KEY,
	ref                       TEXT UNIQUE NOT NULL,
	artifact_id               BLOB NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
	locator_json              TEXT NOT NULL,
	transcription             TEXT,
	description               TEXT,
	transcription_uncertain   INTEGER NOT NULL DEFAULT 0,
	transcription_note        TEXT,

	CHECK (transcription_uncertain IN (0, 1))
) STRICT;

CREATE TABLE citation_notes (
	id              BLOB PRIMARY KEY,
	citation_id     BLOB NOT NULL REFERENCES citations(id) ON DELETE CASCADE,
	body            TEXT NOT NULL
) STRICT;

CREATE TABLE observations (
	id               BLOB PRIMARY KEY,
	ref              TEXT UNIQUE NOT NULL,
	citation_id      BLOB NOT NULL REFERENCES citations(id),
	subject_id       BLOB NOT NULL REFERENCES subjects(id),
	property_id      BLOB NOT NULL REFERENCES properties(id),
	polarity         TEXT NOT NULL DEFAULT 'positive',

	value_text       TEXT,
	value_integer    INTEGER,
	value_date_id    BLOB REFERENCES date_values(id),
	value_name_id    BLOB REFERENCES name_values(id),
	value_subject_id BLOB REFERENCES subjects(id),
	value_term_id    BLOB REFERENCES property_terms(id),

	CHECK (polarity IN ('positive', 'negative'))
) STRICT;

CREATE TABLE observation_notes (
	id              BLOB PRIMARY KEY,
	observation_id  BLOB NOT NULL REFERENCES observations(id) ON DELETE CASCADE,
	body            TEXT NOT NULL
) STRICT;
