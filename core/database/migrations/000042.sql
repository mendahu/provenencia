-- S9-20: seed event_name (text) bound to event for catalogs created before
-- this Property was in subjectvocab.Install. New catalogs are empty here;
-- Install writes the row with a UUIDv7. Property ids are opaque 16-byte keys.

-- Shift bindings that already sit at sort_order >= 1 so the name can follow
-- event_type. Runs only when the Property is not present yet.
UPDATE subject_type_properties
SET sort_order = sort_order + 1
WHERE subject_type_id IN (
	SELECT id FROM subject_types WHERE key = 'event' AND origin = 'provenencia'
)
AND sort_order >= 1
AND NOT EXISTS (
	SELECT 1 FROM properties WHERE key = 'event_name' AND origin = 'provenencia'
);

-- Fixed UUIDv7 (not randomblob); FFI parseID requires version 7.
INSERT INTO properties (id, key, origin, label, description, value_type)
SELECT X'01a11826b4887b4e93d4e39ada6fa988', 'event_name', 'provenencia', 'Event name',
	'Recorded name of a historical event. Not a personal NameValue.',
	'text'
WHERE EXISTS (
	SELECT 1 FROM subject_types WHERE key = 'event' AND origin = 'provenencia'
)
AND NOT EXISTS (
	SELECT 1 FROM properties WHERE key = 'event_name' AND origin = 'provenencia'
);

INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order)
SELECT st.id, p.id, 1
FROM subject_types st
JOIN properties p ON p.key = 'event_name' AND p.origin = 'provenencia'
WHERE st.key = 'event' AND st.origin = 'provenencia'
AND NOT EXISTS (
	SELECT 1 FROM subject_type_properties stp
	WHERE stp.subject_type_id = st.id AND stp.property_id = p.id
);
