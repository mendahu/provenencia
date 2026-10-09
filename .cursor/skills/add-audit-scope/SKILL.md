---
name: add-audit-scope
description: >-
  Keeps Provenencia audit scopes (audit_transaction_scopes) correct. Use when adding a
  new audited table or EntityType passed to audit.Record, writing a delete change, adding
  a scope resolver in core/database/audit/scopes.go, reading "last activity" / the Sources
  list "Updated" sort / updated_revision, or adding a new scope type (e.g. canonical entity).
---

# Audit scopes

Every revision is tagged with the aggregates it worked on in `audit_transaction_scopes`
(`scope_type`, `scope_id`, `audit_transaction_id`). "When did anything under X last change"
is then one indexed `MAX(revision)`, which survives deletes. Design:
[`docs/audit-revision-history.md`](../../../docs/audit-revision-history.md) § 4.1.

Scopes are **derived inside `audit.Record`**, never passed by callers. Each entity type has a
resolver in [`core/database/audit/scopes.go`](../../../core/database/audit/scopes.go). Today
the only scope type is `source`: the Source row and everything under it (notes, metadata,
layout, credibility, artifacts and their files, subjects, citations, observations, their notes
and values).

## New audited entity type (the common case)

A new `EntityType` string passed to `audit.Record` (or a new `rowFacet` in
`deleteimpact/facets.go`) needs a `resolvers` entry:

| Row shape | Resolver |
| --- | --- |
| Carries `source_id` | `directSource("<table>")` |
| Hangs off an artifact / citation / observation | `r.via(ch, "<parent>_id", r.sourceOfX, …)` — reuse the `sourceOfArtifact` / `sourceOfCitation` / `sourceOfObservation` chain |
| Referenced by another row (values, files) | `r.sources(<query>, ch.EntityID)` over the live referrers |
| Vocabulary or conclusion layer | `noScope` — a deliberate decision, not a default |

`Record` rejects an entity type without a resolver (`ErrInvalid`), and
`TestResolversCoverEveryEntityType` parses `core/` so a missing one fails CI before runtime.

## Write changes so they resolve

- **Create / update** changes may carry only the changed fields. Resolvers fall back to the
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
2. Make the relevant resolvers return it (one change may yield several scopes). Revisit each
   `noScope` entry on purpose.
3. Read it with the query above.
4. Backfill existing history with `INSERT OR IGNORE … SELECT` in the next migration
   ([`add-catalog-migration`](../add-catalog-migration/SKILL.md); `000036.sql` is the
   pattern). Per that skill, **no migration test** — test the resolvers.

## Tests (table-driven)

- Resolver behavior (ghost parents, moved rows, no-scope cases):
  `core/database/audit/scopes_internal_test.go`.
- End to end — real write path, then "only its owner's revision moved":
  `core/database/audit/scopes_test.go`.

## Do not

- Tag scopes by hand at call sites or add an "extra scopes" field to `audit.Revision`
- Give `scope_id` an FK (it is polymorphic, like `audit_changes.entity_id`)
- Scope conclusion-layer work (identity claims, evidence pins, canonical entities, promote) to a
  Source — decided: it does not move a Source's "Updated" revision
- Drop a parent id from a delete change to "keep the diff small"
