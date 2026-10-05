-- S9-13: every value stays in the cache; reason says whether it is displayed
-- ('kept') or why not (outvoted, weak, denied, provisional). Nothing the
-- reconciler declines to display is dropped.

ALTER TABLE conclusion_resolved_values ADD COLUMN reason TEXT NOT NULL DEFAULT 'kept';
