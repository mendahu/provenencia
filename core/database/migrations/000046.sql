-- A bridge Subject the researcher switched off on the Promote page: they
-- don't accept that relationship from this Source. Filing skips it on every
-- later Done and claim until they switch it back on. Primary Subjects keep 0.
ALTER TABLE subjects ADD COLUMN filing_declined INTEGER NOT NULL DEFAULT 0 CHECK (filing_declined IN (0, 1));
