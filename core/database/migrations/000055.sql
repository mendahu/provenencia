-- S12-03: Evidence graphs. A Source holds one or more graphs; each Subject
-- belongs to one graph of its Source (the composite key keeps it on its home
-- Source). Artifacts and graphs are linked many-to-many. Every existing Source
-- gets a Default graph whose id is the Source's id; its GPH- ref is minted in
-- Go at open (evidencegraphs.EnsureRefs). Every Subject moves into it and every
-- Artifact is linked to it.

CREATE TABLE evidence_graphs (
	id          BLOB PRIMARY KEY,
	ref         TEXT UNIQUE,
	source_id   BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
	name        TEXT NOT NULL,
	description TEXT,

	UNIQUE (id, source_id)
) STRICT;

CREATE INDEX evidence_graphs_source_id_idx ON evidence_graphs(source_id);

CREATE TABLE evidence_graph_artifacts (
	graph_id    BLOB NOT NULL REFERENCES evidence_graphs(id) ON DELETE CASCADE,
	artifact_id BLOB NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,

	PRIMARY KEY (graph_id, artifact_id)
) STRICT;

CREATE INDEX evidence_graph_artifacts_artifact_id_idx ON evidence_graph_artifacts(artifact_id);

INSERT INTO evidence_graphs (id, source_id, name)
SELECT id, id, 'Default' FROM sources;

INSERT INTO evidence_graph_artifacts (graph_id, artifact_id)
SELECT source_id, id FROM artifacts;

-- migrate: foreign_keys off

CREATE TABLE subjects__new (
	id              BLOB PRIMARY KEY,
	ref             TEXT UNIQUE NOT NULL,
	source_id       BLOB NOT NULL REFERENCES sources(id),
	graph_id        BLOB NOT NULL,
	subject_type_id BLOB NOT NULL REFERENCES subject_types(id),
	label           TEXT,
	description     TEXT,

	UNIQUE (id, subject_type_id),
	FOREIGN KEY (graph_id, source_id) REFERENCES evidence_graphs(id, source_id)
) STRICT;

INSERT INTO subjects__new (id, ref, source_id, graph_id, subject_type_id, label, description)
SELECT id, ref, source_id, source_id, subject_type_id, label, description FROM subjects;

DROP TABLE subjects;
ALTER TABLE subjects__new RENAME TO subjects;

CREATE INDEX subjects_source_id_idx ON subjects(source_id);
CREATE INDEX subjects_graph_id_idx ON subjects(graph_id);
CREATE INDEX subjects_subject_type_id_idx ON subjects(subject_type_id);
