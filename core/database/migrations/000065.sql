-- S15-08: Artifacts and Evidence graphs hold metadata values and layout, with
-- the same fields as a Source (map § 6.6).

CREATE TABLE artifact_metadata (
	id          BLOB PRIMARY KEY,
	artifact_id BLOB NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
	field_id    BLOB NOT NULL REFERENCES source_metadata_fields(id),
	value_text  TEXT,
	UNIQUE (artifact_id, field_id)
) STRICT;

CREATE TABLE artifact_metadata_layout (
	artifact_id BLOB NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
	field_id    BLOB NOT NULL REFERENCES source_metadata_fields(id) ON DELETE CASCADE,
	sort_order  INTEGER NOT NULL,
	dismissed   INTEGER NOT NULL DEFAULT 0 CHECK (dismissed IN (0, 1)),
	PRIMARY KEY (artifact_id, field_id)
) STRICT;

CREATE TABLE evidence_graph_metadata (
	id          BLOB PRIMARY KEY,
	graph_id    BLOB NOT NULL REFERENCES evidence_graphs(id) ON DELETE CASCADE,
	field_id    BLOB NOT NULL REFERENCES source_metadata_fields(id),
	value_text  TEXT,
	UNIQUE (graph_id, field_id)
) STRICT;

CREATE TABLE evidence_graph_metadata_layout (
	graph_id    BLOB NOT NULL REFERENCES evidence_graphs(id) ON DELETE CASCADE,
	field_id    BLOB NOT NULL REFERENCES source_metadata_fields(id) ON DELETE CASCADE,
	sort_order  INTEGER NOT NULL,
	dismissed   INTEGER NOT NULL DEFAULT 0 CHECK (dismissed IN (0, 1)),
	PRIMARY KEY (graph_id, field_id)
) STRICT;

CREATE INDEX artifact_metadata_field_id_idx ON artifact_metadata(field_id);
CREATE INDEX artifact_metadata_layout_field_id_idx ON artifact_metadata_layout(field_id);
CREATE INDEX evidence_graph_metadata_field_id_idx ON evidence_graph_metadata(field_id);
CREATE INDEX evidence_graph_metadata_layout_field_id_idx ON evidence_graph_metadata_layout(field_id);
