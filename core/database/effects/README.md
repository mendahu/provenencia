# effects

What a change to one table touches after the row is written: the Sources it belongs to, the handles whose reconciled values must be rewritten, the search documents to reproject, and whether vocabulary labels or structure changed.

One `Effect` per table outside [`catalogmodel`](../catalogmodel)'s skip bucket. An empty field means that job is unaffected. `None` is an explicit empty entry (`users`, layout rows, grade tables).

Paths walk edges declared in `catalogmodel.FKs`: `up`, `across`, `inbound`, `from`, `field`, `self`, `chain`, `union`, `onField`, `onStatus`, and `sql` for a walk those cannot say. Package init panics when a path names a missing edge. A column on the diff contributes its old and new ids. A column absent from the diff is read from the live row. A delete's old fields are used only when the row is already gone.

`Sources` and `Handles` take `[]rowchange.Change`. `audit.Record` still resolves source scopes itself. Handle recomputes are still placed by hand. This package does not import `audit` or [`deleteimpact`](../deleteimpact). The write orchestrator is what will call the registry.

`registry_completeness_test.go` checks that every non-skip table has an entry and that a cascade into a table with effects is audited. `handle_paths_test.go` checks membership, trigger columns, claim status, and date values. Source parity runs beside `audit.TestSourceScopes`.

The map of the schema packages is in [`catalogmodel`](../catalogmodel/README.md).
