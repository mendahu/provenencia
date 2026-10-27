-- The navigable record a hit opens inside: an Evidence graph's Source (S12-15).
-- Display stubs like 000020's, not FTS-indexed; empty for kinds that open on
-- their own id. ProjectionVersion 9 fills them on open.

ALTER TABLE catalog_search_docs ADD COLUMN display_parent_id TEXT NOT NULL DEFAULT '';
ALTER TABLE catalog_search_docs ADD COLUMN display_parent_ref TEXT NOT NULL DEFAULT '';
