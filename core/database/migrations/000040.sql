-- S9-14: the cache is named for what fills it, the auto-reconciler. A
-- Reconciliation Claim's value is never mirrored here; screens lay a claim
-- over the auto-reconciled value. SQLite can't rename an index, so the value
-- indexes are recreated under the new name.
--
-- auto_reconciler_outcomes: one row per Observation the auto-reconciler
-- considered, with what it did (kept, folded, outvoted, weak, denied,
-- provisional, no_evidence, against) and the rank of the value it went into.
-- Derived like the values: rewritten with its handle, rebuilt on open.

ALTER TABLE conclusion_resolved_values RENAME TO auto_reconciler_values;
ALTER TABLE conclusion_resolved_meta RENAME TO auto_reconciler_meta;

DROP INDEX conclusion_resolved_values_edge_idx;
DROP INDEX conclusion_resolved_values_sort_idx;
DROP INDEX conclusion_resolved_values_date_idx;
CREATE INDEX auto_reconciler_values_edge_idx ON auto_reconciler_values(property_id, value_entity_id);
CREATE INDEX auto_reconciler_values_sort_idx ON auto_reconciler_values(property_id, sort_key);
CREATE INDEX auto_reconciler_values_date_idx ON auto_reconciler_values(property_id, date_lo, date_hi);

CREATE TABLE auto_reconciler_outcomes (
	entity_id      BLOB NOT NULL REFERENCES canonical_entities(id) ON DELETE CASCADE,
	property_id    BLOB NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
	observation_id BLOB NOT NULL,
	reason         TEXT NOT NULL,
	value_rank     INTEGER,
	denied_by      BLOB,

	PRIMARY KEY (entity_id, property_id, observation_id)
) STRICT;

CREATE INDEX auto_reconciler_outcomes_property_idx ON auto_reconciler_outcomes(property_id);
