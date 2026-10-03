-- S9-07b: the Subject fields view is now Properties and Source fields is now
-- Metadata, all the way down. The binding table takes the new name, and the
-- two audit strings already stored for metadata-field deletes are rewritten
-- so history reads in today's vocabulary. Nothing has an FK into
-- subject_type_fields; SQLite carries its own outbound FKs across the rename.

ALTER TABLE subject_type_fields RENAME TO subject_type_properties;

DROP INDEX subject_type_fields_type_id_idx;
DROP INDEX subject_type_fields_property_id_idx;
CREATE INDEX subject_type_properties_type_id_idx ON subject_type_properties(subject_type_id);
CREATE INDEX subject_type_properties_property_id_idx ON subject_type_properties(property_id);

UPDATE audit_changes SET entity_type = 'metadata_field' WHERE entity_type = 'source_field';
UPDATE audit_transactions SET action_type = 'delete_metadata_field' WHERE action_type = 'delete_source_field';
