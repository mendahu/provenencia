# deleteimpact

Official resource and vocab deletes. `Impact` reports whether a row may be erased and what an erase would take with it. `ReleaseFacets` removes audited facet rows before the parent `DELETE` and returns the [`rowchange`](../rowchange) values the revision stores. Owned outbound children are released from the `BucketOwnedOutbound` foreign keys in `catalogmodel`: a row child is deleted by its `PK`, and a File is deleted when no declared foreign key into `files` points at it (its own thumbnail links do not count). Every exported release takes the write's `*database.Tx` and schedules `files.PurgeAfterCommit` for the Files it released; callers never see the objects. Connection facets are released on that same delete. `ReleaseClaim` is the one claim removal (pins first, audited). Subject delete and Unpromote both call it.

The table and foreign-key list is [`catalogmodel`](../catalogmodel). This package adds the delete policy the schema does not store: which kinds have a delete lookup, which foreign keys have a list probe, which cascades official deletes release and audit, and which origins are locked.

It does not decide what a surviving write touches. That registry is [`effects`](../effects). The two do not import each other.

Policy: [`docs/catalog-deletes.md`](../../../docs/catalog-deletes.md). Adding a delete: [`.agents/skills/add-catalog-delete`](../../../.agents/skills/add-catalog-delete/SKILL.md). The map of the schema packages is in [`catalogmodel`](../catalogmodel/README.md).
