-- S9-36: per-Property cardinality. single aims for one value; multiple keeps
-- every distinct surviving value. toponym is the seeded multiple Property
-- (Montréal and Montreal are both names of one Place).

ALTER TABLE properties ADD COLUMN cardinality TEXT NOT NULL DEFAULT 'single'
	CHECK (cardinality IN ('single', 'multiple'));

UPDATE properties SET cardinality = 'multiple'
WHERE key = 'toponym' AND origin = 'provenencia';
