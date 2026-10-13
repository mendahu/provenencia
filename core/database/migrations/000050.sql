-- Citation description becomes a short label. Existing text moves with the column.

ALTER TABLE citations RENAME COLUMN description TO label;
