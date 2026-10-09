# rowchange

One row-level mutation: entity type, id, action (`create`, `update`, `delete`), and each field's old and new value.

`audit.Record` stores a `Change` on `audit_changes`. [`effects`](../effects) reads the same value when it walks a path. [`deleteimpact`](../deleteimpact) builds these changes when it releases facet rows. The type lives here so those packages do not import each other.

`FullRow` is a create: every key is present, old is nil, new is the value. `DeletedRow` is a delete: old is the value, new is nil. An update records a field only when it changed. A column left off the diff is unchanged.

This package does not know table names, foreign keys, or what a change affects. The map of the schema packages is in [`catalogmodel`](../catalogmodel/README.md).
