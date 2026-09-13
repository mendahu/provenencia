-- Source cover: pinned Artifact or type icon (no durable auto mode).
-- First file-bearing Artifact is auto-pinned as the same artifact cover state.
ALTER TABLE sources ADD COLUMN cover_mode TEXT NOT NULL DEFAULT 'type_icon';
ALTER TABLE sources ADD COLUMN primary_artifact_id BLOB REFERENCES artifacts(id) ON DELETE SET NULL;

-- Existing catalogs: pin the first file-bearing Artifact (by ref) when any exist.
UPDATE sources
SET
	cover_mode = 'artifact',
	primary_artifact_id = (
		SELECT a.id
		FROM artifacts a
		WHERE a.source_id = sources.id
			AND a.file_id IS NOT NULL
		ORDER BY a.ref COLLATE NOCASE
		LIMIT 1
	)
WHERE EXISTS (
	SELECT 1
	FROM artifacts a
	WHERE a.source_id = sources.id
		AND a.file_id IS NOT NULL
);

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
