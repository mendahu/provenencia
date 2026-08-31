-- Project label/bookkeeping and users.ref.
-- ref is required by the app; NULL rows are backfilled with USR-… on Open/Create.

ALTER TABLE users ADD COLUMN ref TEXT;

CREATE UNIQUE INDEX users_ref_uidx ON users(ref) WHERE ref IS NOT NULL;

CREATE TABLE project (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	label TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	updated_by BLOB NOT NULL REFERENCES users(id)
) STRICT;
