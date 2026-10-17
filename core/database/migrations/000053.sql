-- S10-52: a depiction value reconciles by its image File. The auto-reconciler
-- cache stores that File id with each depiction value. No foreign key: the
-- cache is derived (BucketSkip), and a cached id must never keep an image File
-- alive. Upkeep rewrites the row when the Observation changes.

ALTER TABLE auto_reconciler_values ADD COLUMN value_image_file_id BLOB;
