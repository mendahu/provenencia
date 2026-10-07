-- S9-20: event_name vocabulary is create-time subjectvocab.Install only
-- (UUIDv7). Do not SQL-seed Properties here — pre-production catalogs are
-- refreshed by re-running Install on the command line. This step remains so
-- catalogs already at user_version ≥ 42 stay openable.
SELECT 1;
