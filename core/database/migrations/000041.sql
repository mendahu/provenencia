-- S9-16: an outvoted Observation keeps the vote that beat it: the winning
-- value's Sources (vote_support) of every Source that voted on that unit
-- (vote_total). NULL for every other outcome. The cache is rebuilt on open
-- (cache version 10), so existing rows need no backfill.

ALTER TABLE auto_reconciler_outcomes ADD COLUMN vote_support INTEGER;
ALTER TABLE auto_reconciler_outcomes ADD COLUMN vote_total INTEGER;
