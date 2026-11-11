-- S15-06: a Source type suggests fields per level (the Source, an Artifact,
-- an Evidence graph); the level is part of the key, so a field may be
-- suggested at more than one level. Existing rows become 'source'.
-- migrate: foreign_keys off

CREATE TABLE source_type_metadata_fields__new (
	source_type_id  BLOB NOT NULL REFERENCES source_types(id) ON DELETE CASCADE,
	field_id        BLOB NOT NULL REFERENCES source_metadata_fields(id) ON DELETE CASCADE,
	level           TEXT NOT NULL CHECK (level IN ('source', 'artifact', 'graph')),
	sort_order      INTEGER,

	PRIMARY KEY (source_type_id, field_id, level)
) STRICT;

INSERT INTO source_type_metadata_fields__new (source_type_id, field_id, level, sort_order)
SELECT source_type_id, field_id, 'source', sort_order FROM source_type_metadata_fields;

DROP TABLE source_type_metadata_fields;
ALTER TABLE source_type_metadata_fields__new RENAME TO source_type_metadata_fields;

CREATE INDEX source_type_metadata_fields_field_id_idx ON source_type_metadata_fields(field_id);
