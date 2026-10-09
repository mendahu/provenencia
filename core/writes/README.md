# writes

The transaction around a catalog write that has moved onto it.

`Run` begins one transaction. The catalog pool has one connection, and the transaction holds it, so the closure and everything it calls must use that transaction rather than `Catalog.DB`. The closure writes the rows and returns them as `[]rowchange.Change`. On a non-empty list, `Run` records the audit revision (source scopes from `effects.Sources`), asks `effects.Resolve` what those changes touch, recomputes handles and reprojects search documents when those sets are non-empty, commits, runs `AfterCommit`, then notifies listeners. A closure error rolls back and does not notify. An empty change list rolls back.

A change list whose every row names a table with effect `None` commits with no revision, no effects, and no listeners, then runs `AfterCommit`. `subject_position` and `file_derivative` are those rows. Thumbnail inserts return the derivative link and leave the derived `files` row off the list, so the batch takes this path; the Source search refresh stays in `AfterCommit`. A mixed batch is audited. `audit.Record` rejects an entity with no scope resolver, so a position or a derivative written inside an audited transaction stays off that change list.

`Run` is generic. The closure's value is the created or updated row (or nothing, for a delete). `Result` is the revision and the resolved effects. `Call` is `Run` with the revision discarded, for a test that only needs the written value.

A second `Run` on the same catalog, including from `AfterCommit` or a listener, returns `database.ErrWriteReentry`. `Run` does not take the catalog session lock.

Commit listeners implement `OnCommit` and `Drop`. If any `OnCommit` fails, `Run` calls `Drop` on every listener. The write is already committed. Nothing registers a listener yet. `Drop` is the hook a later cache will use to discard everything.

Property terms, the Source layer, and the evidence layer are the writes on `Run`: sources, notes, metadata, source types, metadata fields, artifacts, citations, observations, subjects, connect, and positions. The FFI handlers call `Run`. Those functions take `*database.Tx` and do not begin, commit, or record audit. `ingest.File` and `derivatives.Ensure` call `Run` themselves. A checksum conflict starts a second `Run` after the first returns. `artifacts.Delete` unlinks released objects from `AfterCommit`. Source-type and metadata-field `Upsert` stays the un-audited seed path. Name and date inserts stay on the observation's transaction. Subject delete still calls `RecomputeTx` for linked handles, header dependents, and the handles the released rows named, because that walk is gone by the time `Run` resolves effects and the header documents are written in the same call. Conclusion-layer writes still begin their own transaction.

The map of the schema packages is in [`catalogmodel`](../database/catalogmodel/README.md).
