-- Omnibar display stubs on projected search docs (icon + existing cover thumb).
-- Not FTS-indexed; carried on catalog_search_docs for hit hydration without a
-- second catalog round-trip. ProjectionVersion bump heals old projects.

ALTER TABLE catalog_search_docs ADD COLUMN display_icon_key TEXT NOT NULL DEFAULT '';
ALTER TABLE catalog_search_docs ADD COLUMN display_thumbnail_rel_path TEXT NOT NULL DEFAULT '';
