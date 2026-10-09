# writes

The transaction around a catalog write that has moved onto it.

`Run` begins one transaction. The catalog pool has one connection, and the transaction holds it, so the closure and everything it calls must use that transaction rather than `Catalog.DB`. The closure writes the rows and returns them as `[]rowchange.Change`. On a non-empty list, `Run` records the audit revision (source scopes from `effects.Sources`), asks `effects.Resolve` what those changes touch, recomputes handles and reprojects search documents when those sets are non-empty, commits, runs `AfterCommit`, then notifies listeners. A closure error or an empty change list rolls back and does not notify.

`Run` is generic. The closure's value is the created or updated row (or nothing, for a delete). `Result` is the revision and the resolved effects.

A second `Run` on the same catalog, including from `AfterCommit` or a listener, returns `database.ErrWriteReentry`. `Run` does not take the catalog session lock.

Commit listeners implement `OnCommit` and `Drop`. If any `OnCommit` fails, `Run` calls `Drop` on every listener. The write is already committed. Nothing registers a listener yet. `Drop` is the hook a later cache will use to discard everything.

Property-term create, update, and delete are the writes on `Run`. The FFI handlers call it. `Create`, `Update`, and `Delete` take `*database.Tx` and do not begin, commit, or record audit. Every other write still begins its own transaction.

The map of the schema packages is in [`catalogmodel`](../database/catalogmodel/README.md).
