---
name: add-catalog-delete
description: >-
  Adds or rewires a Provenencia official catalog Delete so it goes through
  core/database/deleteimpact (Impact report, owned-outbound release, no private
  sqlInUse). Use when adding Delete for a resource or vocab table, GetDeleteImpact,
  citations.Delete / subjects.Delete / sourcetypes.Delete, ErrInUse cutover,
  inbound list probes, or when a new table needs a delete command.
---

# Add a catalog Delete

Official **resource** and **vocab** deletes go through **`core/database/deleteimpact`**.
Contract: [`docs/catalog-deletes.md`](../../../docs/catalog-deletes.md).
Register: [`core/database/deleteimpact`](../../../core/database/deleteimpact).
Do **not** add a private `SELECT 1 FROM … LIMIT 1` / `sqlInUse` next to `DELETE`.

Schema first: [`add-catalog-migration`](../add-catalog-migration/SKILL.md).
Query package: [`add-catalog-query`](../add-catalog-query/SKILL.md).
Locations: [`add-workspace-location`](../add-workspace-location/SKILL.md).
Rule: [`.cursor/rules/catalog-deletes.mdc`](../../rules/catalog-deletes.mdc).

## What counts as official Delete

| Yes — must call `deleteimpact` | No — leave as a field/facet write |
| --- | --- |
| `sources` / `artifacts` / `citations` / `observations` / `subjects` | Note row delete, `ClearSourceMetadata` |
| Vocab: `sourcetypes`, `sourcefields`, `properties`, `propertyterms`, later subject types | `sourcevocab` / `subjectvocab` join-row remove |
| Later Claim / Narrative resources | `searchindex.Delete` (FTS doc only) |
| Existing `Delete` with no UI this spike | Owned-outbound child SQL that the register already walks |

Facet CASCADE is SQLite’s job. Official `Delete` still **registers** those FKs as `facet` so the pragma test stays honest.

## Checklist (same PR as the table or the writer)

```
- [ ] Table classified (resource / vocab / facet / pool / ownedOutbound)
- [ ] Every live FK tagged; resource inbound edges have count + list probes
- [ ] Title + location projectors registered on the blocker kind (domain package)
- [ ] Proto kind / via keys; recipe L10n heading (unknown via still renders)
- [ ] Domain `Delete`: load → extra gates → Impact in the same tx → connection facets if any → `SnapshotOwned` if owned outbound → DELETE parent → `ReleaseSnapshot` → audit / FTS
- [ ] Extra gates: domain `Delete` missing row → `ErrInvalid`; `GetDeleteImpact` missing → `not_found`; `edge_locked` / infra / `origin_locked` via Impact + `Refuse`
- [ ] Bridge `subjects.Delete` releases connection facets (loop `connectrules.All()` endpoints + `Disambiguation` matching property origin; do not hard-code `role` / `relationship_type`) before the parent; other `observations.Delete` stays `edge_locked`
- [ ] FFI: GetDeleteImpact returns the report; Delete* refuse is a generic in_use / extra-gate code (no report on Error, no apperr ref params)
- [ ] Audit + searchindex stay in the domain package
- [ ] FFI: GetDeleteImpact works for this kind (UI or not)
- [ ] Swift (when the screen ships): `DeleteImpactFlow` + recipe; confirm uses `target.id`;
        `*.in_use` refetch; `.deletedSubject(sourceId:)` for subject erase
- [ ] Tests: empty → erase; inbound → listed + total; GetDeleteImpact missing → not_found;
        domain Delete missing → ErrInvalid; extra gates; pragma includes new FKs;
        Impact.allowed iff SQLite resource DELETE would succeed; audit + FTS on the writer
```

## Domain `Delete`

```go
func Delete(c *database.Catalog, userID, id []byte) error {
    // load row — missing → ErrInvalid (do not run Impact as empty)
    // begin tx
    // extra gates (edge_locked, infra, …) before Impact when they are not already Impact gates
    report, err := deleteimpact.Impact(tx, deleteimpact.KindCitation, id)
    if err != nil { return err }
    if err := deleteimpact.Refuse(report, deleteimpact.Codes{
        InUse: ErrInUse, NotFound: ErrInvalid, // OriginLocked / EdgeLocked when the kind has those gates
    }); err != nil { return err }
    // connection facets (bridge subjects only): released, err := deleteimpact.ReleaseConnectionFacets
    // owned outbound: snap, err := deleteimpact.SnapshotOwned(tx, kind, id) — missing parent is ErrInvalid
    // DELETE parent
    // deleteimpact.ReleaseSnapshot(tx, snap)
    // audit.Record (include ReleasedFacet rows), searchindex.Delete / Reproject
}
```

`ReleaseOwned` snapshots then releases **while the parent row still exists** (`SnapshotOwned` errors if the parent is gone). After `DELETE` it cannot see owned columns — use `SnapshotOwned` → `DELETE` → `ReleaseSnapshot` instead. Writers with no owned outbound skip snapshot/release. Map Impact gates with `deleteimpact.Refuse` so origin/edge/inbound share one switch; do not re-check origin before Impact.

- Preview and write share one **function**. UI calls `GetDeleteImpact` **before** confirm; `Delete` re-runs Impact in-tx. The proto is only on the fetch.
- `GetDeleteImpact` on a missing id returns gate `not_found`. Domain `Delete` on a missing id returns `ErrInvalid`.
- Keep domain sentinels (`citations.ErrInUse`) for `errors.Is`. Do not put refs in `apperr` params or on `Error`.
- Do not `DELETE FROM observations WHERE citation_id = ?` as a cascade. Policy is refuse.
- Writers with **no UI** still cut over in the registry PR (`propertyterms.Delete`). Do not leave `sqlInUse`.

## New table

1. Classify the FKs in the migration skill (`NO ACTION` / `CASCADE` / `SET NULL` / owned outbound).
2. Add register rows + probes + projectors in `core/database/deleteimpact` **in the same PR**.
3. If researchers can delete the row, add `Delete` in the domain package using the pattern above.
4. If this spike has no UI, still register + cut over any existing `Delete`.

## UI

Resource delete chrome uses the **DeleteImpact** recipe ([`macos/App/DesignSystem/Recipes/DeleteImpact/`](../../../macos/App/DesignSystem/Recipes/DeleteImpact/)): confirm when `allowed`, notice when blocked. Recipe owns `via` / `kind` / overflow L10n. Do not hand-build a yes/no sheet that ignores the report. Official `Delete` for a screen lands **in that screen’s PR**, except no-UI writers (registry PR). `usedBy` is not the only gate and must match Impact totals that block.

Shared Swift state is **`DeleteImpactFlow`** (next to the recipe). Models call `ask` / `confirm` / `cancel` and keep only post-delete side effects.

- Confirm errors render **in the confirm sheet** (`deleteImpact.confirmError`). Pane callouts stay for fetch failures when `request == nil`.
- On `*.in_use`, the flow refetches Impact so the same sheet becomes a notice instead of leaving a confirm that can no longer succeed.
- Confirm deletes `request.target.id` (catalog id), never a live selection that may have moved (`activeCitationID`, selected row).
- Listed-row `activate` dismisses first, then navigates on the next run loop (`then:` is injectable in tests).
- Subject erase applies `.deletedSubject(sourceId:)` so Observation `usedBy` and citation lists reload; source erase also busts `sourceWorkspace` / `sourceGraph`; type erase busts `typeSuggestions`.

## Do not

- Raw `DELETE FROM <resource>` in FFI or Swift
- Per-screen count SQL that duplicates a list probe
- `ON DELETE CASCADE` from Artifact→Source or Citation→Artifact
- Treat `searchindex.Delete` as catalog policy
- Treat inbound-empty as allowed when the row is missing or extra-gated
