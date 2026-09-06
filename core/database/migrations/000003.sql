CREATE TABLE audit_transactions (
	id              BLOB PRIMARY KEY,
	revision        INTEGER NOT NULL UNIQUE,
	user_id         BLOB REFERENCES users(id),
	action_type     TEXT,
	description     TEXT,
	created_at      TEXT NOT NULL
) STRICT;

CREATE TABLE audit_changes (
	id                      BLOB PRIMARY KEY,
	audit_transaction_id    BLOB NOT NULL
		REFERENCES audit_transactions(id),
	entity_type             TEXT NOT NULL,
	entity_id               BLOB NOT NULL,
	action                  TEXT NOT NULL,
	changes_json            TEXT NOT NULL,
	CHECK (action IN ('create', 'update', 'delete'))
) STRICT;
