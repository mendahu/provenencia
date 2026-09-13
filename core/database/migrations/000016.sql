-- Drop first-file auto-pin / backfill from the original 000015 rollout.
-- Source cover is type_icon until the researcher explicitly pins a raster Artifact.
UPDATE sources
SET
	cover_mode = 'type_icon',
	primary_artifact_id = NULL
WHERE cover_mode = 'artifact'
	OR primary_artifact_id IS NOT NULL;
