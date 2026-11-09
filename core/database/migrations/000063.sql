-- S15-02: a Source's own date (original publication, or when the record was made) and a not-public flag.
ALTER TABLE sources ADD COLUMN date_value_id BLOB REFERENCES date_values(id);
ALTER TABLE sources ADD COLUMN is_public INTEGER NOT NULL DEFAULT 1 CHECK (is_public IN (0, 1));
CREATE UNIQUE INDEX sources_date_value_id_uidx ON sources(date_value_id) WHERE date_value_id IS NOT NULL;
