-- Hot foreign-key indexes. SQLite does not auto-index FK columns; these
-- cover ListBySource / HasAnyForSource, SourceIDsForFile, notes-by-source,
-- subjects-by-source, and ReprojectSourcesForType (plus ON DELETE CASCADE
-- lookups on artifacts.source_id / source_notes.source_id).

CREATE INDEX artifacts_source_id_idx ON artifacts(source_id);
CREATE INDEX artifacts_file_id_idx ON artifacts(file_id);
CREATE INDEX source_notes_source_id_idx ON source_notes(source_id);
CREATE INDEX subjects_source_id_idx ON subjects(source_id);
CREATE INDEX sources_source_type_id_idx ON sources(source_type_id);
