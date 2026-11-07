-- S14-36: value_type = boolean. A yes/no fact: observations.value_boolean
-- holds 0 or 1, and the auto-reconciler cache keeps it beside value_integer
-- under the same check.
-- No seed Property is boolean; researchers create them (from S14-37).
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
		'text', 'integer', 'date', 'name', 'subject', 'term', 'depiction', 'boolean'
	)),
	CHECK (value_type <> 'boolean' OR cardinality = 'single')
) STRICT;

INSERT INTO properties__new (id, key, origin, label, description, value_type, cardinality)
SELECT id, key, origin, label, description, value_type, cardinality FROM properties;

DROP TABLE properties;
ALTER TABLE properties__new RENAME TO properties;

ALTER TABLE observations ADD COLUMN value_boolean INTEGER CHECK (value_boolean IN (0, 1));
ALTER TABLE auto_reconciler_values ADD COLUMN value_boolean INTEGER CHECK (value_boolean IN (0, 1));
