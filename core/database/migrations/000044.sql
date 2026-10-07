-- S9-28: whether a kinship term keeps direction. Spouse, sibling, and cousin
-- stay symmetric (a spouse link matches either way). The other seeded kinship
-- terms are directed (a parent link matches only that way). The flag is a
-- property of the term, not a list in the filer. Place-relationship terms
-- (S9-38) use the same column.

ALTER TABLE property_terms ADD COLUMN directed INTEGER NOT NULL DEFAULT 0
	CHECK (directed IN (0, 1));

UPDATE property_terms SET directed = 1
WHERE origin = 'provenencia'
	AND key NOT IN ('spouse', 'sibling', 'cousin')
	AND property_id = (
		SELECT id FROM properties
		WHERE key = 'relationship_type' AND origin = 'provenencia'
	);
