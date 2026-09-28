-- Restore artifact FK indexes dropped when 000030 rebuilt artifacts.
-- Add indexes Impact COUNT/list probes and parent DELETE scans need.

CREATE INDEX artifacts_source_id_idx ON artifacts(source_id);
CREATE INDEX artifacts_file_id_idx ON artifacts(file_id);
CREATE INDEX citations_artifact_id_idx ON citations(artifact_id);
CREATE INDEX citation_notes_citation_id_idx ON citation_notes(citation_id);
CREATE INDEX observations_citation_id_idx ON observations(citation_id);
CREATE INDEX observations_property_id_idx ON observations(property_id);
CREATE INDEX observations_value_subject_id_idx ON observations(value_subject_id);
CREATE INDEX observations_value_term_id_idx ON observations(value_term_id);
CREATE INDEX observation_notes_observation_id_idx ON observation_notes(observation_id);
CREATE INDEX sources_primary_artifact_id_idx ON sources(primary_artifact_id);
CREATE INDEX source_metadata_field_id_idx ON source_metadata(field_id);
CREATE INDEX source_credibility_assessments_grade_id_idx ON source_credibility_assessments(credibility_grade_id);
CREATE INDEX source_metadata_layout_source_id_idx ON source_metadata_layout(source_id);
CREATE INDEX source_metadata_layout_field_id_idx ON source_metadata_layout(field_id);
CREATE INDEX source_type_metadata_fields_type_id_idx ON source_type_metadata_fields(source_type_id);
CREATE INDEX source_type_metadata_fields_field_id_idx ON source_type_metadata_fields(field_id);
CREATE INDEX file_derivatives_derived_file_id_idx ON file_derivatives(derived_file_id);
CREATE INDEX subjects_subject_type_id_idx ON subjects(subject_type_id);
CREATE INDEX subject_positions_subject_id_idx ON subject_positions(subject_id);
CREATE INDEX subject_type_fields_type_id_idx ON subject_type_fields(subject_type_id);
CREATE INDEX subject_type_fields_property_id_idx ON subject_type_fields(property_id);
CREATE INDEX project_updated_by_idx ON project(updated_by);
CREATE INDEX audit_transactions_user_id_idx ON audit_transactions(user_id);
CREATE INDEX audit_changes_transaction_id_idx ON audit_changes(audit_transaction_id);
