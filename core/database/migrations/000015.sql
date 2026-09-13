-- Source cover: pinned Artifact with a raster thumbnail, or type icon.
-- Cover is never auto-pinned; researchers set it explicitly.
ALTER TABLE sources ADD COLUMN cover_mode TEXT NOT NULL DEFAULT 'type_icon';
ALTER TABLE sources ADD COLUMN primary_artifact_id BLOB REFERENCES artifacts(id) ON DELETE SET NULL;

-- If the primary Artifact row is deleted, clear the pin and return to type icon.
CREATE TRIGGER sources_clear_cover_on_primary_null
AFTER UPDATE OF primary_artifact_id ON sources
FOR EACH ROW
WHEN OLD.primary_artifact_id IS NOT NULL
	AND NEW.primary_artifact_id IS NULL
	AND NEW.cover_mode = 'artifact'
BEGIN
	UPDATE sources SET cover_mode = 'type_icon' WHERE id = NEW.id;
END;
