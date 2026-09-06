CREATE TABLE date_values (
	id              BLOB PRIMARY KEY,
	kind            TEXT NOT NULL,
	qualifier       TEXT,
	calendar        TEXT,

	start_year      INTEGER,
	start_month     INTEGER,
	start_day       INTEGER,
	start_hour      INTEGER,
	start_minute    INTEGER,
	start_second    INTEGER,
	start_millisecond INTEGER,

	end_year        INTEGER,
	end_month       INTEGER,
	end_day         INTEGER,
	end_hour        INTEGER,
	end_minute      INTEGER,
	end_second      INTEGER,
	end_millisecond INTEGER,

	phrase          TEXT,

	CHECK (start_month IS NULL OR start_month BETWEEN 1 AND 12),
	CHECK (end_month   IS NULL OR end_month BETWEEN 1 AND 12),
	CHECK (start_day   IS NULL OR start_day BETWEEN 1 AND 31),
	CHECK (end_day     IS NULL OR end_day BETWEEN 1 AND 31),
	CHECK (start_hour IS NULL OR start_hour BETWEEN 0 AND 23),
	CHECK (end_hour   IS NULL OR end_hour BETWEEN 0 AND 23),
	CHECK (start_minute IS NULL OR start_minute BETWEEN 0 AND 59),
	CHECK (end_minute   IS NULL OR end_minute BETWEEN 0 AND 59),
	CHECK (start_second IS NULL OR start_second BETWEEN 0 AND 59),
	CHECK (end_second   IS NULL OR end_second BETWEEN 0 AND 59),
	CHECK (start_millisecond IS NULL OR start_millisecond BETWEEN 0 AND 999),
	CHECK (end_millisecond   IS NULL OR end_millisecond BETWEEN 0 AND 999)
) STRICT;
