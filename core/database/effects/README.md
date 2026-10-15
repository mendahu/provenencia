# effects

What a change to one table touches after the row is written: the Sources it belongs to, the handles whose reconciled values must be rewritten, the search documents to reproject, and whether vocabulary labels or structure changed.

One `Effect` per table outside [`catalogmodel`](../catalogmodel)'s skip bucket. An empty field means that job is unaffected. `None` is an explicit empty entry (`users`, layout rows, grade tables).

Paths walk edges declared in `catalogmodel.FKs`: `up`, `across`, `inbound`, `from`, `field`, `self`, `chain`, `union`, `onField`, `onStatus`, and `sql` for a walk those cannot say. Package init panics when a path names a missing edge. A column on the diff contributes its old and new ids. A column absent from the diff is read from the live row. A delete's old fields are used only when the row is already gone.

`Handles` and `Resolve` take `[]rowchange.Change`. `Resolve` is the one walk of a batch: it returns `Set.Sources` and `Set.HandleScopes` (the audit scopes), `Set.Handles` (the recompute set), search documents, and the vocabulary and structure flags. `Effect.HandleScope` is the handles whose last activity a change moves; `Handles` is wider. [`writes.Run`](../../writes) calls `Resolve` once, hands the `Set` to `audit.Record`, then recomputes handles and reprojects search when those sets are non-empty. Each scope path is walked once per write.

`Scopes` (`scopes.go`) declares every audit scope type once: its name and which ids of a `Set` belong to it (`source` from `Sources`, `canonical_entity` from `HandleScopes`). `audit.Record` stores one `audit_transaction_scopes` row per id of each `Scopes` entry, and the display cache ([`graphcache`](../graphcache)) keeps each type's last activity current from the same ids. A new scope type is one `Scopes` entry. `Commit` (`commit.go`) is a landed write: its revision, the time that revision was recorded, and its `Set`. It lives here because the display cache accepts it and `core/database`, which imports the cache, hands it to listeners. This package does not import `audit` or [`deleteimpact`](../deleteimpact).

`registry_completeness_test.go` checks that every non-skip table has an entry and that a cascade into a table with effects is audited. `handle_paths_test.go` checks membership, trigger columns, claim status, and date values. Source parity runs beside `audit.TestSourceScopes`.

The map of the schema packages is in [`catalogmodel`](../catalogmodel/README.md).
