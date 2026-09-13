-- Durable project identity for install-local chrome (navigation history, etc.).
-- Mint on Create; heal NULL on Open. Never rewrite after mint.

ALTER TABLE project ADD COLUMN uuid BLOB;

CREATE UNIQUE INDEX project_uuid_uidx ON project(uuid) WHERE uuid IS NOT NULL;
