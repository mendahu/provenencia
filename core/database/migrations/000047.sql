-- A neutral term is no evidence either way: an "unknown" sex at birth never
-- wins or outvotes a recorded one, and never agrees or conflicts in matching.
-- The flag is a property of the term, like `directed` (000044) and
-- `inverse_key` (000046), so reconciliation and matching read it instead of
-- naming keys. Install sets it on new catalogs.

ALTER TABLE property_terms ADD COLUMN neutral INTEGER NOT NULL DEFAULT 0
	CHECK (neutral IN (0, 1));
