# writes

The transaction around a catalog write that has moved onto it.

`Run` begins one transaction. The catalog pool has one connection, and the transaction holds it, so the closure and everything it calls must use that transaction rather than `Catalog.DB`. The closure writes the rows and returns them as `[]rowchange.Change`. On a non-empty list, `Run` records the audit revision (source scopes from `effects.Sources`), asks `effects.Resolve` what those changes touch, recomputes handles and reprojects search documents when those sets are non-empty, commits, runs `AfterCommit`, then notifies listeners. A closure error rolls back and does not notify. An empty change list rolls back when `Op.Unaudited` is false.

`Op.Unaudited` commits with no revision, no effects, and no listeners, then runs `AfterCommit`. Thumbnail inserts use it. `file_derivatives` stays `{None: true}`, and the Source search refresh stays in `AfterCommit`.

`Run` is generic. The closure's value is the created or updated row (or nothing, for a delete). `Result` is the revision and the resolved effects.

A second `Run` on the same catalog, including from `AfterCommit` or a listener, returns `database.ErrWriteReentry`. `Run` does not take the catalog session lock.

Commit listeners implement `OnCommit` and `Drop`. If any `OnCommit` fails, `Run` calls `Drop` on every listener. The write is already committed. Nothing registers a listener yet. `Drop` is the hook a later cache will use to discard everything.

Property terms and the Source layer are the writes on `Run`: sources, notes, metadata, source types, metadata fields, and artifacts. The FFI handlers call `Run`. Those functions take `*database.Tx` and do not begin, commit, or record audit. `ingest.File` and `derivatives.Ensure` call `Run` themselves. A checksum conflict starts a second `Run` after the first returns. `artifacts.Delete` unlinks released objects from `AfterCommit`. Source-type and metadata-field `Upsert` stays the un-audited seed path. Evidence-layer and conclusion-layer writes still begin their own transaction.

The map of the schema packages is in [`catalogmodel`](../database/catalogmodel/README.md).
