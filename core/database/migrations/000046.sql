-- A directed kinship term can name its inverse: "Mary parent of John" and
-- "John child of Mary" are one relationship read from opposite ends. Matching
-- normalizes a term and its inverse to one key with the ends swapped. The
-- inverse is a property of the term, like `directed` (000044); Install sets
-- it on new catalogs, and the seeded pairs are marked here for older ones.

ALTER TABLE property_terms ADD COLUMN inverse_key TEXT;

UPDATE property_terms SET inverse_key = CASE key
		WHEN 'parent' THEN 'child'
		WHEN 'child' THEN 'parent'
		WHEN 'grandparent' THEN 'grandchild'
		WHEN 'grandchild' THEN 'grandparent'
		WHEN 'pibling' THEN 'nibling'
		WHEN 'nibling' THEN 'pibling'
		WHEN 'guardian' THEN 'ward'
		WHEN 'ward' THEN 'guardian'
	END
WHERE origin = 'provenencia'
	AND key IN ('parent', 'child', 'grandparent', 'grandchild', 'pibling', 'nibling', 'guardian', 'ward')
	AND property_id = (
		SELECT id FROM properties
		WHERE key = 'relationship_type' AND origin = 'provenencia'
	);
