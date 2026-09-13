-- Source types carry a curated evidence-icon key (design-system type_* set).
ALTER TABLE source_types ADD COLUMN icon_key TEXT NOT NULL DEFAULT 'type_evidence';
