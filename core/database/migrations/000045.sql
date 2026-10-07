-- S9-38: place relationships and periods. Term category (hierarchical /
-- temporal) for place_relationship_type. Existing catalogs get the type,
-- properties, terms, and bindings; new catalogs are empty here and Install
-- mints UUIDv7s.

ALTER TABLE property_terms ADD COLUMN category TEXT
	CHECK (category IS NULL OR category IN ('hierarchical', 'temporal'));

-- Subject type place_relationship (only when provenencia types already exist).
INSERT INTO subject_types (id, key, origin, label, description, ref_prefix, candidate_ref_prefix)
SELECT randomblob(16), 'place_relationship', 'provenencia', 'Place relationship',
	'Association between two places: part-of or succession.',
	'PLR', 'CLR'
WHERE EXISTS (
	SELECT 1 FROM subject_types WHERE key = 'place' AND origin = 'provenencia'
)
AND NOT EXISTS (
	SELECT 1 FROM subject_types WHERE key = 'place_relationship' AND origin = 'provenencia'
);

INSERT INTO properties (id, key, origin, label, description, value_type, cardinality)
SELECT randomblob(16), 'from', 'provenencia', 'From',
	'Place relationship: the part, or the predecessor',
	'subject', 'single'
WHERE EXISTS (
	SELECT 1 FROM subject_types WHERE key = 'place' AND origin = 'provenencia'
)
AND NOT EXISTS (
	SELECT 1 FROM properties WHERE key = 'from' AND origin = 'provenencia'
);

INSERT INTO properties (id, key, origin, label, description, value_type, cardinality)
SELECT randomblob(16), 'to', 'provenencia', 'To',
	'Place relationship: the whole, or the successor',
	'subject', 'single'
WHERE EXISTS (
	SELECT 1 FROM subject_types WHERE key = 'place' AND origin = 'provenencia'
)
AND NOT EXISTS (
	SELECT 1 FROM properties WHERE key = 'to' AND origin = 'provenencia'
);

INSERT INTO properties (id, key, origin, label, description, value_type, cardinality)
SELECT randomblob(16), 'place_relationship_type', 'provenencia', 'Place relationship type',
	'Part of or succeeded by. Product-locked vocabulary.',
	'term', 'single'
WHERE EXISTS (
	SELECT 1 FROM subject_types WHERE key = 'place' AND origin = 'provenencia'
)
AND NOT EXISTS (
	SELECT 1 FROM properties WHERE key = 'place_relationship_type' AND origin = 'provenencia'
);

-- Period on place (reuse start_date / end_date Properties).
INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order)
SELECT st.id, p.id, 1
FROM subject_types st
JOIN properties p ON p.key = 'start_date' AND p.origin = 'provenencia'
WHERE st.key = 'place' AND st.origin = 'provenencia'
AND NOT EXISTS (
	SELECT 1 FROM subject_type_properties stp
	WHERE stp.subject_type_id = st.id AND stp.property_id = p.id
);

INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order)
SELECT st.id, p.id, 2
FROM subject_types st
JOIN properties p ON p.key = 'end_date' AND p.origin = 'provenencia'
WHERE st.key = 'place' AND st.origin = 'provenencia'
AND NOT EXISTS (
	SELECT 1 FROM subject_type_properties stp
	WHERE stp.subject_type_id = st.id AND stp.property_id = p.id
);

-- Bridge bindings: from, to, type, membership span.
INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order)
SELECT st.id, p.id, 0
FROM subject_types st
JOIN properties p ON p.key = 'from' AND p.origin = 'provenencia'
WHERE st.key = 'place_relationship' AND st.origin = 'provenencia'
AND NOT EXISTS (
	SELECT 1 FROM subject_type_properties stp
	WHERE stp.subject_type_id = st.id AND stp.property_id = p.id
);

INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order)
SELECT st.id, p.id, 1
FROM subject_types st
JOIN properties p ON p.key = 'to' AND p.origin = 'provenencia'
WHERE st.key = 'place_relationship' AND st.origin = 'provenencia'
AND NOT EXISTS (
	SELECT 1 FROM subject_type_properties stp
	WHERE stp.subject_type_id = st.id AND stp.property_id = p.id
);

INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order)
SELECT st.id, p.id, 2
FROM subject_types st
JOIN properties p ON p.key = 'place_relationship_type' AND p.origin = 'provenencia'
WHERE st.key = 'place_relationship' AND st.origin = 'provenencia'
AND NOT EXISTS (
	SELECT 1 FROM subject_type_properties stp
	WHERE stp.subject_type_id = st.id AND stp.property_id = p.id
);

INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order)
SELECT st.id, p.id, 3
FROM subject_types st
JOIN properties p ON p.key = 'start_date' AND p.origin = 'provenencia'
WHERE st.key = 'place_relationship' AND st.origin = 'provenencia'
AND NOT EXISTS (
	SELECT 1 FROM subject_type_properties stp
	WHERE stp.subject_type_id = st.id AND stp.property_id = p.id
);

INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order)
SELECT st.id, p.id, 4
FROM subject_types st
JOIN properties p ON p.key = 'end_date' AND p.origin = 'provenencia'
WHERE st.key = 'place_relationship' AND st.origin = 'provenencia'
AND NOT EXISTS (
	SELECT 1 FROM subject_type_properties stp
	WHERE stp.subject_type_id = st.id AND stp.property_id = p.id
);

INSERT INTO property_terms (id, property_id, key, origin, label, description, directed, category)
SELECT randomblob(16), p.id, 'part_of', 'provenencia', 'Part of',
	'From is part of to. Builds display chains.',
	1, 'hierarchical'
FROM properties p
WHERE p.key = 'place_relationship_type' AND p.origin = 'provenencia'
AND NOT EXISTS (
	SELECT 1 FROM property_terms t
	WHERE t.property_id = p.id AND t.key = 'part_of' AND t.origin = 'provenencia'
);

INSERT INTO property_terms (id, property_id, key, origin, label, description, directed, category)
SELECT randomblob(16), p.id, 'succeeded_by', 'provenencia', 'Succeeded by',
	'From was succeeded by to. Lineage; never a chain.',
	1, 'temporal'
FROM properties p
WHERE p.key = 'place_relationship_type' AND p.origin = 'provenencia'
AND NOT EXISTS (
	SELECT 1 FROM property_terms t
	WHERE t.property_id = p.id AND t.key = 'succeeded_by' AND t.origin = 'provenencia'
);
