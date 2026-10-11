---
name: add-audit-scope
description: >-
  Keeps Provenencia audit scopes (audit_transaction_scopes) correct. Use when adding a
  new audited table or EntityType passed to audit.Record, writing a delete change,
  adding an effect entry whose Source path feeds audit scopes, reading "last activity" /
  the Sources list "Updated" sort / updated_revision, or adding a new scope type
  (e.g. canonical entity).
---

# Audit scopes

Every revision is tagged with the aggregates it worked on in `audit_transaction_scopes`
(`scope_type`, `scope_id`, `audit_transaction_id`). "When did anything under X last change"
is then one indexed `MAX(revision)`, which survives deletes. Design:
[`docs/audit-revision-history.md`](../../../docs/audit-revision-history.md) § 4.1.

Scopes are **derived inside `audit.Record`**, never passed by callers. `Record` stores
source scopes from [`effects.Sources`](../../../core/database/effects/resolve.go). Today
the only scope type is `source`: the Source row and everything under it (notes, metadata,
layout, credibility, artifacts and their files, subjects, citations, observations, their notes
and values).

## New audited entity type (the common case)

A new `EntityType` string passed to `audit.Record` (or a new `rowFacet` in
`deleteimpact/facets.go`) needs a non-`None` effect in
[`core/database/effects`](../../../core/database/effects/registry.go). `Record` rejects an
unknown type and a `None` row (`ErrInvalid`). `TestResolversCoverEveryEntityType` parses
`core/` so a missing effect fails CI before runtime.

An entity whose effect is `None` (`subject_position`, `file_derivative`) is not recorded.
An all-`None` list commits without a revision. In a mixed list, `Run` drops the `None`
rows before `Record`, so they are stored with the revision and are not recorded.

Source scopes come from the effect's `Source` path, not a hand resolver.

| Row shape | Effect |
| --- | --- |
| Carries `source_id` | `Source: field("source_id")` |
| Hangs off an artifact / citation / observation | `Source: up(...)` along that chain |
| Referenced by another row (values, files) | `Source: chain(inbound(...), ...)` |
| Vocabulary or conclusion layer | no `Source` path — a deliberate decision, not a default |

## Write changes so they resolve

- **Create / update** changes may carry only the changed fields. The path falls back to the
  live row by `entity_id`.
- **Delete** changes must carry the parent FK (`source_id`, `artifact_id`, `citation_id`,
  `observation_id`, …) in `rowchange.DeletedRow(...)`. The row is gone by `Record` time, so the
  fields are the only link. Children released in the same revision (notes before their
  citation) reach a deleted parent through that parent's delete change (the ghost map).
- A `rowFacet(...)` must list its FK column in `cols` for the same reason.
- A row that moves between parents (update with `old` ≠ `new` parent id) scopes to both.

## Read "last activity under X"

```sql
SELECT COALESCE(MAX(t.revision), 0)
FROM audit_transaction_scopes sc
JOIN audit_transactions t ON t.id = sc.audit_transaction_id
WHERE sc.scope_type = 'source' AND sc.scope_id = ?
```

See `sqlList` / `sqlLatestRevision` in `core/database/sources/sources.go`. Do **not** join
`audit_changes` on `entity_type = …` for this — it misses children and deletes.

## New scope type

1. Add a const next to `ScopeSource` in `scopes.go`.
2. Teach `Record` to store it from the effect set (one change may yield several scopes).
   Revisit each effect that has no `Source` path on purpose.
3. Read it with the query above.
4. Backfill existing history with `INSERT OR IGNORE … SELECT` in the next migration
   ([`add-catalog-migration`](../add-catalog-migration/SKILL.md); `000036.sql` is the
   pattern). Per that skill, **no migration test** — test the effect path.

## Tests (table-driven)

- Stored source scopes match `effects.Sources`:
  `core/database/audit/scopes_effects_test.go`.
- End to end — real write path, then "only its owner's revision moved":
  `core/database/audit/scopes_test.go`.

## Do not

- Tag scopes by hand at call sites or add an "extra scopes" field to `audit.Revision`
- Give `scope_id` an FK (it is polymorphic, like `audit_changes.entity_id`)
- Scope conclusion-layer work (identity claims, evidence pins, canonical entities, promote) to a
  Source — decided: it does not move a Source's "Updated" revision
- Drop a parent id from a delete change to "keep the diff small"
- Add a hand-written resolver map in `scopes.go`. A recorded entity is a non-`None` effect.
