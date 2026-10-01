# Catalog deletes

Authoritative contract for official **resource** and **vocab** deletes. The table-by-table register lives in [`core/database/deleteimpact`](../core/database/deleteimpact). Spike 8 shipped the first screens: [`archive/spike-8`](deployment-plan/archive/spike-8/).

Skills: [`add-catalog-delete`](../.cursor/skills/add-catalog-delete/SKILL.md), [`add-catalog-migration`](../.cursor/skills/add-catalog-migration/SKILL.md).

## Policy

1. **Buckets are named on the FK.** Resource → resource is `NO ACTION`. Facet → parent is `ON DELETE CASCADE`. Owned outbound (parent column → child) stays `NO ACTION`; official `Delete` releases the child (`always` or `ifUnused`). `ifUnused` for `files` also unlinks `objects/` after commit when the `files` row is actually dropped. Optional back-pointers are `SET NULL`. **Connection facets** are a predicate split on `observations.subject_id` (below) — not SQLite `CASCADE`.
2. **`Impact` explains the schema; it does not replace it.** A successful erase must be a transaction the FK graph would accept **after** official release of owned outbound and connection facets. Owned outbound cannot be looked up after the parent row is gone: take `SnapshotOwned` (and `CollectFileObjects` for files) **before** `DELETE`, then `ReleaseSnapshot`. `ReleaseOwned` is only valid while the parent still exists. Connection facets on a bridge subject are released **before** the parent `DELETE`.
3. **Erase iff the target exists, extra gates pass, and inbound resources are empty.** Extra gates: `not_found`, `edge_locked`, `infra`, `origin_locked`.
4. **No general cross-resource cascade.** Citation ↛ Observations. The one named exception is connection facets on a bridge subject.
5. **Every official `Delete` calls `deleteimpact.Impact` in the same tx.** No private `sqlInUse`. Preview is `GetDeleteImpact`; the writer re-runs `Impact`. Writers with no UI still cut over when the registry lands.
6. **Edge-lock is a gate on Observation writes.** `observations.Delete` / `Update` of an edge row stay `edge_locked`. The only writer that may remove an edge row is official `subjects.Delete` releasing connection facets.
7. **The map stays honest.** Every live FK is in the register. Every catalog table (except `sqlite_%` / FTS shadows) is registered. CI fails if `PRAGMA foreign_key_list` shows an unregistered or mis-tagged FK, a resource edge without a list probe, an owned-outbound column missing from the parent’s release list, or an FK child column with no covering index. Resource/vocab kinds with an `Exists` check have a projector `Section`. `Impact.allowed` (inbound) iff a raw resource `DELETE` would succeed with FKs on, **or** the only remaining pointers are registered `connectionFacet` predicates. Every `CASCADE` FK is classified **audited** or **silent**; an audited one has exactly one facet release (below), a silent one has none.
8. **Composite FKs are keyed by their leading column.** `PRAGMA foreign_key_list` returns one row per column; the register names the FK by its `seq = 0` column and lists the whole tuple in `FromCols` (`identity_claims (subject_id, subject_type_id) → subjects`). Only the leading column needs a covering index.
9. **Conclusion rows.** `identity_claims` is a facet of both its Subject and its handle; `identity_claim_evidence` (pins) is a facet of both its claim and its Observation. All four FKs are **audited** `CASCADE`s. Nothing in Conclusion blocks a delete: a Subject leaves its handles, a pinned Observation leaves the claims that pinned it (they stay, weaker — model §5.2). Impact names the affected handles as `cascades`. Handles and claim confidence grades have no delete path yet.

### Connection facets (bridge subjects only)

`observations.subject_id` is one SQLite FK and two logical edges:

| Predicate | Bucket | Blocks Impact? |
| --- | --- | --- |
| Ordinary / extra Observations on this subject | Resource inbound | Yes |
| Edge-locked rows on this subject, plus each Connect rule’s `Disambiguation` property (from `core/connectrules.All()`, matching the property’s origin) | `connectionFacet` | No — official `subjects.Delete` releases them first |
| `observations.value_subject_id` (this subject is an endpoint) | Resource inbound | Yes (G2) |

Endpoints and the Citation stay. Extra Add-property rows on the bridge still block. Connection facets are released through the same facet-release registry (below), so each released Observation's own notes, pins, and owned values go with it, audited.

### Facet release (audited facets)

Research rows that a delete takes with it are removed **explicitly** and audited, never left to SQLite. One `facetRelease` entry in `core/database/deleteimpact/facets.go` defines each relationship:

| Field | Role |
| --- | --- |
| `Parent`, `Via` | The parent kind and the FK covered (`child_table.fk_col`). The honesty test checks it against the register. |
| `Release` | Deletes the rows and returns their audit changes, plus any handles whose membership or evidence changed. |
| `Remaining` | A count of rows still pointing at the parent. It must be zero after `Release`, or the delete fails with `deleteimpact.release_incomplete`. So the `CASCADE` backstop can never fire silently. |
| `Named` + `Count` / `List` / `Child` | Named releases also appear in `Report.Cascades`, from the same predicate `Release` deletes. |

- **One call per delete.** Domain `Delete` calls `deleteimpact.ReleaseFacets(tx, kind, id)` once, after `Refuse` and before the parent `DELETE`. Then it prepends `Released.Changes` to its own audit change. It never hand-calls a per-table helper.
- **Releases compose.** A released row that is itself a parent (a claim, a connection Observation) has its own facets released first, through the same function.
- **The registry owns ordering.** Callers don't sequence anything.
- **`Released.Handles` is the seam for derived data.** Resolved-value upkeep and search reprojection read it instead of re-querying.

| Audited (released) | Silent (`CASCADE` only) |
| --- | --- |
| `source_notes`, `source_metadata`, `source_credibility_assessments`, `source_metadata_layout` (both FKs), `citation_notes`, `observation_notes`, `identity_claims` (both FKs; `entity_id` has no delete path yet), `identity_claim_evidence` (both FKs) | `subject_positions` (unaudited layout), `source_type_metadata_fields` / `subject_type_fields` (vocab joins, never audited), `name_value_parts` (part of its name value) |

A silent `CASCADE` is only for rows that are never audited on create either. If a row has an audit entity type, its delete is audited.

### Origin locks

| Origin | Source types / source fields | Properties / terms |
| --- | --- | --- |
| `plugin:…` | Never (`origin_locked`) | Never |
| `provenencia` (seeded) | Unused → erase | Locked |
| `user` | Unused → erase | Unused and no term rows → erase |

Plugin rows are a future plugin manager, not researcher trash.

### Terms are resources

`property_terms.property_id` is `NO ACTION`. A property with any term row is blocked until those terms are deleted. There is no Terms page — that notice is honest. Type-bindings (`subject_type_fields`) CASCADE and do not block.

## Impact report

```text
Impact
  allowed          bool
  gate             enum     // ok | inbound | not_found | edge_locked | infra | origin_locked
  groups[]                  // inbound only; empty when gate ≠ inbound
    via, kind, total, listed[] { id, ref, title, location }
  cascades[]                // non-blocking; same shape; rows that go with the target
```

- **Cascades never gate.** They are the **Named** facet releases: rows the delete removes that the researcher should hear about. They're filled whenever the target exists and no extra gate fired, and `Refuse` ignores them. Today there are two:
  - a Subject's Identity Claims, of any status (`identity_claims.subject_id` → the handles it leaves)
  - a pinned Observation's pins (`identity_claim_evidence.observation_id` → the handles whose claims lose that evidence)
- **The confirm speaks every cascade.** It appends one sentence per group: specific copy for known vias, and a generic "This also changes {kind} {refs}" otherwise. An unknown via is never dropped, matching the rule for blocking groups.

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
5. Classify each new `CASCADE` FK as audited or silent. An audited one needs a `facetRelease` entry: use `rowFacet` for plain rows, or a custom `Release` when the row has facets of its own.
6. Domain `Delete`: load → extra gates → Impact → `Refuse` → `ReleaseFacets` → `SnapshotOwned` if owned outbound → `DELETE` parent → `ReleaseSnapshot` → audit (released changes first) / FTS.
7. No UI yet: still register + cut over any existing `Delete`.
8. Tests: empty → erase; inbound → listed + `total`; `GetDeleteImpact` missing → `not_found`; domain `Delete` missing → `ErrInvalid`; extra gates; pragma; SQLite-agrees (with connection-facet carve-out); audit + searchindex.

## UI

Official trash uses the DeleteImpact recipe: confirm if `allowed`, notice if not. Facet writes stay on `.pvConfirm`. `usedBy` must match the Impact totals that actually block. Swift screens share **`DeleteImpactFlow`**: confirm errors live in the sheet; `*.in_use` refetches Impact so the sheet becomes a notice; confirm deletes `target.id`; listed-row activate dismisses before navigating; subject erase applies `.deletedSubject(sourceId:)`.

## FFI

`GetDeleteImpact` is the only rich payload. `Delete*` errors stay coded and generic. Do not attach `DeleteImpact` to `Error`.
