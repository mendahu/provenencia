-- S10-39: a Person's preferred portrait is a display pointer to one likeness
-- Observation. Not evidence. Deleting the Observation clears it.
ALTER TABLE canonical_entities ADD COLUMN preferred_likeness_observation_id BLOB
	REFERENCES observations(id) ON DELETE SET NULL;
CREATE INDEX canonical_entities_preferred_likeness_observation_id_idx
	ON canonical_entities(preferred_likeness_observation_id);
