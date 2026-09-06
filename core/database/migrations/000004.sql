CREATE TABLE date_values (
	id              BLOB PRIMARY KEY,
	kind            TEXT NOT NULL,
	qualifier       TEXT,
	calendar        TEXT,

	start_year      INTEGER,
	start_month     INTEGER,
	start_day       INTEGER,

	end_year        INTEGER,
	end_month       INTEGER,
	end_day         INTEGER,

	phrase          TEXT,

	CHECK (start_month IS NULL OR start_month BETWEEN 1 AND 12),
	CHECK (end_month   IS NULL OR end_month BETWEEN 1 AND 12),
	CHECK (start_day   IS NULL OR start_day BETWEEN 1 AND 31),
	CHECK (end_day     IS NULL OR end_day BETWEEN 1 AND 31)
) STRICT;
