-- S14-10: a Citation belongs to one Evidence graph. citations.graph_id is
-- NOT NULL and never changes. Backfill: the graph of the Subject of the
-- Citation's earliest Observation; else the first graph linked to its Artifact;
-- else the first graph of the Artifact's Source (first = the graph list's
-- order). Then every (graph, Artifact) pair a Citation holds is linked, so a
-- stored Citation's graph is always linked to its Artifact.
-- migrate: foreign_keys off

CREATE TABLE citations__new (
	id                        BLOB PRIMARY KEY,
	ref                       TEXT UNIQUE NOT NULL,
	artifact_id               BLOB NOT NULL REFERENCES artifacts(id),
	graph_id                  BLOB NOT NULL REFERENCES evidence_graphs(id) ON DELETE CASCADE,
	locator_json              TEXT NOT NULL,
	transcription             TEXT,
	label                     TEXT,
	transcription_uncertain   INTEGER NOT NULL DEFAULT 0,
	transcription_note        TEXT,

	CHECK (transcription_uncertain IN (0, 1))
) STRICT;

INSERT INTO citations__new (
	id, ref, artifact_id, graph_id, locator_json, transcription, label,
	transcription_uncertain, transcription_note
)
SELECT
	c.id, c.ref, c.artifact_id,
	COALESCE(
		(SELECT s.graph_id FROM observations o
			JOIN subjects s ON s.id = o.subject_id
			WHERE o.citation_id = c.id ORDER BY o.id LIMIT 1),
		(SELECT g.id FROM evidence_graph_artifacts l
			JOIN evidence_graphs g ON g.id = l.graph_id
			WHERE l.artifact_id = c.artifact_id
			ORDER BY g.name COLLATE NOCASE, g.ref, g.id LIMIT 1),
		(SELECT g.id FROM artifacts a
			JOIN evidence_graphs g ON g.source_id = a.source_id
			WHERE a.id = c.artifact_id
			ORDER BY g.name COLLATE NOCASE, g.ref, g.id LIMIT 1)
	),
	c.locator_json, c.transcription, c.label,
	c.transcription_uncertain, c.transcription_note
FROM citations c;

DROP TABLE citations;
ALTER TABLE citations__new RENAME TO citations;

CREATE INDEX citations_artifact_id_idx ON citations(artifact_id);
CREATE INDEX citations_graph_artifact_idx ON citations(graph_id, artifact_id);

INSERT OR IGNORE INTO evidence_graph_artifacts (graph_id, artifact_id)
SELECT DISTINCT graph_id, artifact_id FROM citations;
