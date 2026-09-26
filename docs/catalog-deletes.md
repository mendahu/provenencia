# Catalog deletes

Authoritative contract for official **resource** and **vocab** deletes. Spike 8 implements this (`S8-09`…`S8-19`). The table-by-table register lives in [`core/database/deleteimpact`](../core/database/deleteimpact) (policy list still in [`deployment-plan/spike-8/deployment-plan.md`](deployment-plan/spike-8/deployment-plan.md) §S8-09.3 until the spike archives).

Skills: [`add-catalog-delete`](../.cursor/skills/add-catalog-delete/SKILL.md), [`add-catalog-migration`](../.cursor/skills/add-catalog-migration/SKILL.md).

## Policy

1. **Buckets are named on the FK.** Resource → resource is `NO ACTION`. Facet → parent is `ON DELETE CASCADE`. Owned outbound (parent column → child) stays `NO ACTION`; official `Delete` releases the child (`always` or `ifUnused`). Optional back-pointers are `SET NULL`. **Connection facets** are a predicate split on `observations.subject_id` (below) — not SQLite `CASCADE`.
2. **`Impact` explains the schema; it does not replace it.** A successful erase must be a transaction the FK graph would accept **after** official release of owned outbound and connection facets.
3. **Erase iff the target exists, extra gates pass, and inbound resources are empty.** Extra gates: `not_found`, `edge_locked`, `infra`, `origin_locked`.
4. **No general cross-resource cascade.** Citation ↛ Observations. The one named exception is connection facets on a bridge subject.
5. **Every official `Delete` calls `deleteimpact.Impact` in the same tx.** No private `sqlInUse`. Preview is `GetDeleteImpact`; the writer re-runs `Impact`. Writers with no UI still cut over when the registry lands.
6. **Edge-lock is a gate on Observation writes.** `observations.Delete` / `Update` of an edge row stay `edge_locked`. The only writer that may remove an edge row is official `subjects.Delete` releasing connection facets.
7. **The map stays honest.** Every live FK is in the register. CI fails if `PRAGMA foreign_key_list` shows an unregistered or mis-tagged FK, a resource edge without a list probe, or an owned-outbound column missing from the parent’s release list. `Impact.allowed` (inbound) iff a raw resource `DELETE` would succeed with FKs on, **or** the only remaining pointers are registered `connectionFacet` predicates.

### Connection facets (bridge subjects only)

`observations.subject_id` is one SQLite FK and two logical edges:

| Predicate | Bucket | Blocks Impact? |
| --- | --- | --- |
| Ordinary / extra Observations on this subject | Resource inbound | Yes |
| Edge-locked rows on this subject, plus the Connect rule’s disambiguation row (`role` / `relationship_type`) | `connectionFacet` | No — official `subjects.Delete` releases them first |
| `observations.value_subject_id` (this subject is an endpoint) | Resource inbound | Yes (G2) |

Endpoints and the Citation stay. Extra Add-property rows on the bridge still block.

### Origin locks

| Origin | Source types / source fields | Properties / terms |
| --- | --- | --- |
| `plugin:…` | Never (`origin_locked`) | Never |
| `provenencia` (seeded) | Unused → erase | Locked |
| `user` | Unused → erase | Unused and no term rows → erase |

Plugin rows are a future plugin manager, not researcher trash.

### Terms are resources

`property_terms.property_id` is `NO ACTION`. A property with any term row is blocked until those terms are deleted. There is no Terms page this spike — that notice is honest. Type-bindings (`subject_type_fields`) CASCADE and do not block.

## Impact report

```text
Impact
  allowed          bool
  gate             enum     // ok | inbound | not_found | edge_locked | infra | origin_locked
  groups[]                  // inbound only; empty when gate ≠ inbound
    via, kind, total, listed[] { id, ref, title, location }
```

- Missing id → `not_found`, never allowed.
- Infra / skip tables → `infra`.
- `via` / `kind` are machine keys. **DeleteImpact recipe** L10n writes headings and overflow. Unknown `via` still lists refs.
- Title and `WorkspaceLocation` come from **per-kind projectors** registered by the domain package.
- The **report proto is only on `GetDeleteImpact`** (preflight fetch). `Delete` re-runs `Impact` in-tx as a safety net; refuse is a generic `*.in_use` / extra-gate code — not the report on `Error`, not refs in `apperr` params. The UI must not call `Delete` when preflight is blocked.

## New table (same PR)

1. Classify every new FK (migration skill).
2. Register the table + probes + owned-outbound / connection-facet release in `core/database/deleteimpact`.
3. Register title + location projectors.
4. Add proto `kind` / `via` values. Recipe L10n gets a heading; unknown keys must still render.
5. Domain `Delete`: load → extra gates → Impact → release connection facets / owned outbound as registered → `DELETE` parent → audit / FTS.
6. No UI yet: still register + cut over any existing `Delete`.
7. Tests: empty → erase; inbound → listed + `total`; missing → `not_found`; extra gates; pragma; SQLite-agrees (with connection-facet carve-out); audit + searchindex.

## UI

Official trash uses the DeleteImpact recipe: confirm if `allowed`, notice if not. Facet writes stay on `.pvConfirm`. `usedBy` must match the Impact totals that actually block.

## FFI

`GetDeleteImpact` is the only rich payload. `Delete*` errors stay coded and generic. Do not attach `DeleteImpact` to `Error`.
