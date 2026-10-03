-- Resolved-values cache (Spike 9 R3): the resolver's ranked clusters for every
-- Property on every canonical handle. Derived and rebuildable; nothing
-- references it. Maintained in each write's transaction by resolvedvalues and
-- rebuilt on open when conclusion_resolved_meta.cache_version is stale.
--
-- No state column: single / merged / mixed is read off the rows (rank 2
-- exists → mixed; else rank 1 support). Dates and names are stored whole as
-- protobuf DateValueInput / NameValueInput; SQL orders by sort_key and
-- filters dates by date_lo / date_hi (filled from S9-21). value_entity_id is
-- filled from S9-28. Value columns carry no FKs: a rebuild is the repair.

CREATE TABLE conclusion_resolved_values (
	entity_id       BLOB NOT NULL REFERENCES canonical_entities(id) ON DELETE CASCADE,
	property_id     BLOB NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
	rank            INTEGER NOT NULL CHECK (rank >= 1),
	value_text      TEXT,
	value_integer   INTEGER,
	value_term_id   BLOB,
	value_entity_id BLOB,
	value_date      BLOB,
	value_name      BLOB,
	date_lo         INTEGER,
	date_hi         INTEGER,
	sort_key        TEXT,
	support         INTEGER NOT NULL,

	PRIMARY KEY (entity_id, property_id, rank)
) STRICT;

CREATE INDEX conclusion_resolved_values_edge_idx ON conclusion_resolved_values(property_id, value_entity_id);
CREATE INDEX conclusion_resolved_values_sort_idx ON conclusion_resolved_values(property_id, sort_key);
CREATE INDEX conclusion_resolved_values_date_idx ON conclusion_resolved_values(property_id, date_lo, date_hi);

CREATE TABLE conclusion_resolved_meta (
	id            INTEGER PRIMARY KEY CHECK (id = 1),
	cache_version INTEGER NOT NULL
) STRICT;

INSERT INTO conclusion_resolved_meta (id, cache_version) VALUES (1, 0);
