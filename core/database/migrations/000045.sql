-- S9-38: place_relationship vocabulary (type, from/to, place_relationship_type,
-- part_of / succeeded_by, period bindings) is create-time subjectvocab.Install
-- only (UUIDv7). Do not SQL-seed here — pre-production catalogs are refreshed
-- by re-running Install on the command line. This step remains so catalogs
-- already at user_version ≥ 45 stay openable.
SELECT 1;
