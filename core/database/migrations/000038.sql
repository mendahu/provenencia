-- S9-14: how many negative Observations count against each cached value, so
-- a detail page can say a record disagrees without recomputing.

ALTER TABLE conclusion_resolved_values ADD COLUMN against INTEGER NOT NULL DEFAULT 0;
