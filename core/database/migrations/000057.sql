-- A-4 / S12-24: note attribution seeks audit_changes by entity_id.
CREATE INDEX audit_changes_entity_id_idx ON audit_changes(entity_id);
