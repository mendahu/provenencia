-- S10-32: value_type = depiction. The value is a depiction_values row: a
-- region of the Citation's page and the image File Go stored from it.
-- likeness is seeded by subjectvocab.Install at create time, not here.
-- migrate: foreign_keys off

CREATE TABLE properties__new (
	id          BLOB PRIMARY KEY,
	key         TEXT NOT NULL,
	origin      TEXT NOT NULL,
	label       TEXT NOT NULL,
	description TEXT,
	value_type  TEXT NOT NULL,
	cardinality TEXT NOT NULL DEFAULT 'single'
		CHECK (cardinality IN ('single', 'multiple')),

	UNIQUE (key, origin),
	CHECK (value_type IN (
		'text', 'integer', 'date', 'name', 'subject', 'term', 'depiction'
	))
) STRICT;

INSERT INTO properties__new (id, key, origin, label, description, value_type, cardinality)
SELECT id, key, origin, label, description, value_type, cardinality FROM properties;

DROP TABLE properties;
ALTER TABLE properties__new RENAME TO properties;


CREATE TABLE depiction_values (
	id            BLOB PRIMARY KEY,
	region_json   TEXT NOT NULL,
	image_file_id BLOB NOT NULL REFERENCES files(id)
) STRICT;

CREATE INDEX depiction_values_image_file_id_idx ON depiction_values(image_file_id);

ALTER TABLE observations ADD COLUMN value_depiction_id BLOB REFERENCES depiction_values(id);

CREATE UNIQUE INDEX observations_value_depiction_id_uidx
	ON observations(value_depiction_id) WHERE value_depiction_id IS NOT NULL;
