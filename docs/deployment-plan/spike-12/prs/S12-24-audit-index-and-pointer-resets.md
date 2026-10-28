# S12-24 — Audit history: indexed note authors, audited pointer resets

**Status:** Planned
**Spike / slice:** Spike 12 · Slice 5 — Audit fixes: High
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: an index and two audited releases; no view change)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s12-24-audit-index-and-pointer-resets`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-23 did; the catalog format number moves to 57 on its own ([`docs/versioning.md`](../../../versioning.md#project-format-version))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-migrations.mdc`, `.cursor/rules/catalog-model.mdc`, `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/catalog-audit.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [app-health audit](../app-health-audit.md) A-4, A-22; [map review](../deployment-plan.review.md) K-17, K-13; working rules "Each PR updates the docs and skills it makes false" (slices 5–7 line names this PR for `catalog-deletes.md` and `add-catalog-delete`); [`audit-revision-history.md`](../../../audit-revision-history.md); [`catalog-deletes.md`](../../../catalog-deletes.md); [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) preferred-likeness paragraph; Spike 14 S14-01 reuses the same preferred pointer for Event/Place thumbnails
**Planned:** 2026-10-28 against `main` @ `0f8a2165`
**Review:** [`S12-24-audit-index-and-pointer-resets.review.md`](S12-24-audit-index-and-pointer-resets.review.md) — Fits / Ready, 2 rounds; every blocker Fixed

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

Opening a Source page with notes no longer scans the whole audit history for each note's author (audit A-4). Deleting a cover Artifact or a portrait Observation records the pointer clear in the same revision as the delete, so history and replay agree with what `SetCover` and `SetPreferredLikeness` already audit (audit A-22).

**Why now.** Bar outcome 12 ("Graph work stays fast on a split census" — Source page does not scan history) and the Medium conformance hole on cover/portrait resets. Slice 5. Map review K-17 put both in one PR.

**What stays.** Note SQL shape (two correlated subqueries), the `sources_clear_cover_on_primary_null` trigger, SQLite `ON DELETE SET NULL` on both FKs as a backstop, delete confirm UI (no new Cascades), and every other facet release. `FK.Audited` stays CASCADE-only (`add-catalog-model`).

## 2. User stories

**US-1.** As the engine, when a Source's notes load create attribution from `audit_changes`, each subquery seeks by `entity_id` instead of scanning the table, so a Source page stays fast on a large history.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `EXPLAIN QUERY PLAN` on `sqlListNotes` and on `sqlGetNoteAttributed` with a 16-byte bind, on a catalog that includes migration `000057` | No plan detail equals or starts with `SCAN c` |

**US-2.** As a researcher, when I delete the Artifact that is a Source's cover, the delete revision records the Source's cover clear the same way `SetCover` would, so history shows when the cover went.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Source with cover_mode `artifact` and `primary_artifact_id` = Artifact A; `artifacts.Delete` of A succeeds | The Source row has `cover_mode = type_icon` and `primary_artifact_id` NULL |
| AC-2.2 | AC-2.1's `delete_artifact` revision | Contains a `source` / `update` change whose `Fields` equal those from `SetCover(…, CoverModeTypeIcon, nil)` on the same prior row, ahead of the `artifact` / `delete` change |
| AC-2.3 | Artifact B that is not any Source's cover; `artifacts.Delete` of B | No `source` change in that revision |

**US-3.** As a researcher, when I delete the Observation that is a Person's preferred portrait, the delete revision records the handle's pointer clear the same way `SetPreferredLikeness` would.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Handle H with `preferred_likeness_observation_id` = Observation O; `observations.Delete` of O succeeds | H's pointer is NULL |
| AC-3.2 | AC-3.1's `delete_observation` revision | Contains a `canonical_entity` / `update` change whose `Fields` equal those from `SetPreferredLikeness(…, nil)` on the same prior row, ahead of the `observation` / `delete` change (pins and notes may sit between them as today) |
| AC-3.3 | Observation P that is no handle's preferred portrait; `observations.Delete` of P | No `canonical_entity` change in that revision |

**US-4.** As a contributor, when I add a `BucketOptional` `SET NULL` FK, `TestFacetReleaseHonesty` fails until it has a `facetRelease`, so the next pointer cannot skip audit.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Both live `BucketOptional` FKs (`sources.primary_artifact_id`, `canonical_entities.preferred_likeness_observation_id`) | Each stays not `Audited`; each has exactly one `facetRelease` whose `Via` matches the FK key |
| AC-4.2 | `TestFacetReleaseHonesty` | Passes; removing either release fails the test |

## 3. Scope

| In | Covered by |
| --- | --- |
| Migration `000057.sql`: `CREATE INDEX audit_changes_entity_id_idx ON audit_changes(entity_id)` (A-4) | STEP-1, AC-1.1 |
| `dbtest` query-plan test on the note attribution queries (proves the index; no `user_version` / `sqlite_master` assert) | STEP-1, AC-1.1, T-1 |
| `optionalPointer` *(new)* beside `rowFacet`; two call sites for cover and preferred likeness that return update changes (A-22; K-17) | STEP-2, AC-2.*, AC-3.*, T-2, T-3 |
| Parity: delete-release `Fields` equal `SetCover` / `SetPreferredLikeness` clear `Fields` on the same prior row | STEP-2, AC-2.2, AC-3.2, T-2, T-3 |
| `TestFacetReleaseHonesty` extended to require a release for every `BucketOptional` `SET NULL` FK whose `To` is deletable (`Audited` stays CASCADE-only) | STEP-2, AC-4.*, T-4 |
| `artifacts.Delete` comment no longer says cover is SQLite's job | STEP-2 |
| Docs and skills that still tie facet release to CASCADE only, plus the preferred-likeness sentence | STEP-3 |

| Out | Home |
| --- | --- |
| Folding the two note attribution subqueries into one | not planned (audit A-4 called it optional; the index alone meets the bar) |
| `GetSourceWorkspace`'s per-Artifact lookups (audit L-11) | [audit, Not scheduled](../app-health-audit.md#not-scheduled); map scope entry |
| Naming cover or portrait clears in `Report.Cascades` (`Named: true`) | not planned (map: update changes only; confirm UI unchanged) |
| Dropping `sources_clear_cover_on_primary_null` or the SQLite `SET NULL` clauses | not planned (backstops for non-official writes) |
| Setting `Audited: true` on `BucketOptional` FKs | not planned (`Audited` stays CASCADE-only per `add-catalog-model`) |
| Event/Place thumbnail parity on the same pointer | S14-01 (map: "the release covers it whatever the kind") |
| A new index on `(entity_type, entity_id)` | not planned (audit measured `entity_id` alone) |
| A per-migration `user_version` or `sqlite_master` index-presence test | not planned (`add-catalog-migration` forbids it; T-1's plan failure proves the index) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Note attribution SQL | `core/database/sources/notes.go` `sqlListNotes`, `sqlGetNoteAttributed` | Two correlated subqueries each: `audit_changes c` filtered by `entity_type = 'source_note' AND entity_id = n.id AND action = 'create'`. `ListNotes` feeds `GetSourceWorkspace`; `GetNote` uses the same SQL for the single-note RPC |
| Audit indexes | `core/database/migrations/000031.sql` | Only `audit_changes_transaction_id_idx` on `audit_changes`. Highest migration on `main` is `000056.sql` (format 56) |
| Query-plan helper | `core/database/dbtest/dbtest.go` `QueryPlan` | `EXPLAIN QUERY PLAN`; returns detail strings. Precedent: `evidencegraphs.TestListQueryPlan`, `identityclaims.TestBridgeHeadQueryPlans` (plan only; no `user_version` assert) |
| Optional FKs | `core/database/catalogmodel/model.go` L140, L191; `FK.Audited` comment L71–74 | `sources.primary_artifact_id` → `artifacts` `SET NULL` `BucketOptional`; `canonical_entities.preferred_likeness_observation_id` → `observations` `SET NULL` `BucketOptional`. Neither `Audited`. `Audited` is documented for CASCADE facets only |
| Cover clear trigger | `core/database/migrations/000015.sql` `sources_clear_cover_on_primary_null` | After `primary_artifact_id` becomes NULL while `cover_mode = 'artifact'`, sets `cover_mode = 'type_icon'`. Still present |
| SetCover audit | `core/database/sources/sources.go` `SetCover`, `CoverModeTypeIcon` | Updates `cover_mode` + `primary_artifact_id`; returns one `source` / `update` with fields `cover_mode` always and `primary_artifact_id` when it changes (`database.UUIDJSON`) |
| SetPreferredLikeness audit | `core/database/canonicalentities/canonicalentities.go` `SetPreferredLikeness` | Updates the pointer; returns one `canonical_entity` / `update` with field `preferred_likeness_observation_id` (`database.UUIDJSON`) |
| Artifact delete | `core/database/artifacts/artifacts.go` `Delete`; `artifacts_test.go` ~L597–634 | `ReleaseFacets` → `SnapshotOwned` → `DELETE` → `ReleaseSnapshot`. Comment: "Cover SET NULL is SQLite's job." Cover erase asserts the Source row only, not the revision |
| Observation delete | `core/database/observations/write.go` `Delete`; `core/database/likenesses/likenesses_test.go` subtest "deleting the Observation clears the pointer" (~L373) | Same release pattern; no preferred-pointer release. Likeness subtest sets the pointer, deletes, asserts NULL; discards changes |
| Facet registry | `core/database/deleteimpact/facets.go` `buildFacetReleases`, `rowFacet`, `facetRelease` | Releases are CASCADE row deletes (and custom claim/pin/connection releases). `Remaining` required for CASCADE. No `BucketOptional` entry. Production imports: `database`, `catalogmodel`, `files`, `rowchange`, `apperr` — not `sources` or `canonicalentities` |
| Honesty test | `core/database/deleteimpact/pragma_test.go` `TestFacetReleaseHonesty` | Requires Audited ↔ release for CASCADE only; errors on `Audited` when not CASCADE; errors on a release whose FK is silent CASCADE |
| Docs / skills | `docs/catalog-deletes.md` § Facet release and policy §1; `.agents/skills/add-catalog-delete/SKILL.md`; `.agents/skills/add-catalog-model/SKILL.md` step 3 and 5; `.agents/skills/add-catalog-migration/SKILL.md` optional-pointer sentence; `facets.go` header / `Via` / `Remaining` comments; `conclusion-layer-data-model.md` preferred-likeness paragraph | Facet release tied to audited CASCADE only. Conclusion: "Deleting the Observation clears it" with no audit sentence |

**Precedent.** `f323fb84` introduced `facetRelease` + `rowFacet` + `TestFacetReleaseHonesty`. Extend with `optionalPointer` beside `rowFacet` (same declared shape: table, columns, derived SQL). For the query-plan test, copy `evidencegraphs.TestListQueryPlan` only. For parity, drive both the domain clear write and the delete in one test and compare `Fields`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Migration `000057.sql` is exactly `CREATE INDEX audit_changes_entity_id_idx ON audit_changes(entity_id);` (plus a one-line comment naming A-4 / S12-24). No other DDL | Map and audit A-4; next free step after `000056` | `(entity_type, entity_id)` composite — audit measured `entity_id` alone |
| D-2 | `TestNotesAttributionQueryPlan` *(new)* in `core/database/sources/notes_test.go` *(new)* (package `sources`) creates a catalog, runs `dbtest.QueryPlan` on `sqlListNotes` and `sqlGetNoteAttributed` with a 16-byte id bind, and fails when any detail equals `SCAN c` or has prefix `SCAN c `. No `user_version` or `sqlite_master` assert | Proves the index the way `TestListQueryPlan` does; `add-catalog-migration` forbids migration-presence tests; a hard-coded `57` breaks on the next step | Asserting format 57 / index name in `sqlite_master` — migration test |
| D-3 | Leave both `BucketOptional` FKs with `Audited: false`. `FK.Audited` and `add-catalog-model` stay CASCADE-only | The bucket already means "optional pointer that official Delete must release"; a always-true `Audited` on optional is a second reading of the flag | `Audited: true` on optional — bends `add-catalog-model` step 3 |
| D-4 | Add `optionalPointer(parent Kind, table, fkCol, entityType, entityIDCol string, sideEffectCols ...string) facetRelease` *(new)* beside `rowFacet`. It derives `Via` (`table.fkCol`), `Remaining` (`SELECT COUNT(*) FROM table WHERE fkCol = ?`), and Release that: (1) `SELECT entityIDCol, fkCol, sideEffectCols… FROM table WHERE fkCol = ? ORDER BY entityIDCol`; (2) for each row `UPDATE table SET fkCol = NULL WHERE entityIDCol = ?` (only the FK — so `sources_clear_cover_on_primary_null` still owns cover_mode); (3) re-`SELECT fkCol, sideEffectCols…` for that id; (4) emit one `ActionUpdate` whose `Fields` always include `fkCol` (old `database.UUIDJSON` → nil) and every `sideEffectCol` (old value → re-read value; text as string, UUID as `UUIDJSON`), matching `SetCover`'s habit of always naming `cover_mode`. `Named` stays false. Call sites: (1) immediately after the `KindArtifact` `evidence_graph_artifacts` `rowFacet`: `optionalPointer(KindArtifact, "sources", "primary_artifact_id", "source", "id", "cover_mode")`; (2) immediately after the `KindObservation` `observation_notes` `rowFacet`: `optionalPointer(KindObservation, "canonical_entities", "preferred_likeness_observation_id", "canonical_entity", "id")` with no side-effect cols | One declared shape; the trigger stays the single owner of "pointer cleared → type_icon"; Remaining cannot be forgotten; no import of `sources` / `canonicalentities`; no third hard-coded `"type_icon"` | `pointerReset{col, new: "type_icon"}` — third copy of the cover rule beside the trigger and `SetCover`. Two hand-written Release funcs. Calling domain writers — import cycle |
| D-5 | T-2 and T-3 each include a parity case in the same catalog: two Sources (cover) or two handles (portrait) with identical prior pointer state; on the first, capture `Fields` from `SetCover(…, type_icon, nil)` / `SetPreferredLikeness(…, nil)`; on the second, run the delete and require the release change's `Fields` deep-equal those maps | Ties the helper's diffs to the domain writers without importing them into `deleteimpact`; two rows keep the clear write from changing the delete target | Documenting "match SetCover" with no test — copies drift. Serial clear-then-re-set on one row — harder to read |
| D-6 | `TestFacetReleaseHonesty` gains a `BucketOptional` branch only: every such FK must be `SET NULL`; when `fk.To` is deletable it must have exactly one release; a release on such an FK needs non-empty `Remaining` and non-nil `Release`. Keep "Audited but not CASCADE" as an error (optional stays not Audited). Do not change the existing `!Audited && has` path — non-CASCADE FKs already `continue` before it, and connection-facet releases already sit on non-CASCADE FKs | Map: extend the same honesty test, no second list; `Audited` stays CASCADE-only | Requiring `Audited` on optional — D-3. Rewriting the CASCADE `!Audited && has` branch — unnecessary |
| D-7 | Keep the `000015` trigger and both SQLite `SET NULL` clauses. Official Delete clears through the release first; Remaining proves zero pointers before parent `DELETE` | Backstop for raw SQL | Dropping the trigger in this migration |
| D-8 | Domain `Delete` call sites stay as they are. Only the `artifacts.Delete` comment changes to say cover is released and audited through `deleteimpact` before the parent delete | Callers already prepend released changes | Hand-clearing cover inside `artifacts.Delete` |
| D-9 | Extend the existing cover-erase path in `artifacts_test.go` (~L597–634) with `t.Run` cases for AC-2.1–AC-2.3 and D-5 parity: call `Delete` through `writes.Call` (same pattern as the "releases its graph links" subtest ~L649) and assert the returned changes / revision. Extend `likenesses_test.go` "deleting the Observation clears the pointer" (~L373) the same way for AC-3.* and D-5. No new top-level test names | One fixture per delete; `go-tests.mdc` one family; `runDelete` already uses `writes.Run` but drops changes | New top-level tests that re-seed the same setup |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `000057.sql` *(new)* | migration | `000031.sql` audit indexes | New index only; do not edit shipped `000031` |
| `optionalPointer` *(new)* | func | `rowFacet` in `facets.go` | `rowFacet` `DELETE`s child rows; this `UPDATE`s only the FK, re-reads side-effect columns (trigger-owned), and emits an update change |
| Two `optionalPointer(…)` call sites | registry entries | existing `rowFacet(…)` lines in `buildFacetReleases` | New vias; same list |
| `TestNotesAttributionQueryPlan` *(new)* | test | `evidencegraphs.TestListQueryPlan` | Different SQL; same helper |

**Builds on.** `facetRelease` registry + honesty (`add-catalog-delete`); `rowFacet`'s declared-columns pattern; `dbtest.QueryPlan`; `SetCover` / `SetPreferredLikeness` as the parity oracles.
**Deletes.** The mechanism "official Artifact/Observation delete leaves cover/portrait clear to SQLite SET NULL without an audit change" — replaced by the two `optionalPointer` releases. The `artifacts.Delete` comment line "Cover SET NULL is SQLite's job."
**Next adds.** When a third `BucketOptional` FK's parent becomes deletable (narrative `primary_entity_id` in `docs/narrative-layer-data-model.md` points at `canonical_entities`, which has no delete path yet), it is one `optionalPointer(…)` line and fails honesty until added. S14-01 reuses `preferred_likeness_observation_id` with no new release.

### Reach

None — no new path into existing code; the same deletes gain audit changes and the same note reads gain an index. Bridge `eraseObservation` already goes through `releaseFacets`, so the portrait release fires there too.

### Scale

At the map's Scale (200,000 audit rows; a Source page with notes): each note attribution subquery seeks `audit_changes_entity_id_idx` instead of `SCAN c` (audit: 140.7 ms → ~0.1 ms for five notes at 200k). Cover/portrait release: one `SELECT` + one `UPDATE` per pointing row (0 or 1 in normal use) inside the existing delete transaction; `sources_primary_artifact_id_idx` and `canonical_entities_preferred_likeness_observation_id_idx` already cover the `Remaining` seeks. No new invalidation keys (Source / handle effects follow the update change through `effects.Resolve`).

### Schema

Migration `core/database/migrations/000057.sql` *(new)*:

```sql
-- A-4 / S12-24: note attribution seeks audit_changes by entity_id.
CREATE INDEX audit_changes_entity_id_idx ON audit_changes(entity_id);
```

No `catalogmodel` table or FK field change (`Audited` stays false on both optional FKs). Delete class unchanged. Audit: update changes on existing entity types `source` / `canonical_entity`.

### Go core

| Package | Change |
| --- | --- |
| `core/database/migrations` | `000057.sql` *(new)* |
| `core/database/deleteimpact/facets.go` | `optionalPointer` *(new)*; two call sites in `buildFacetReleases`; comments on header / `Via` / `Remaining` so SET NULL optional releases are named |
| `core/database/deleteimpact/pragma_test.go` | `TestFacetReleaseHonesty` optional branch (D-6) |
| `core/database/artifacts/artifacts.go` | `Delete` comment only |
| `core/database/artifacts/artifacts_test.go` | extend cover-erase path (D-9) |
| `core/database/likenesses/likenesses_test.go` | extend preferred-delete subtest (D-9) |
| `core/database/sources/notes_test.go` *(new)* | `TestNotesAttributionQueryPlan` |

Sketch — `optionalPointer` signature only:

```go
func optionalPointer(parent catalogmodel.Kind, table, fkCol, entityType, entityIDCol string, sideEffectCols ...string) facetRelease
```

Errors: none new — incomplete release still `deleteimpact.release_incomplete`.

### FFI

None — existing Delete RPCs record whatever changes domain `Delete` returns.

### macOS

None — no chrome, no store method, no L10n.

### Strings

None — no user-facing copy. Sweep covers the deleted comment phrase and CASCADE-only facet-release wording in the docs/skills this PR updates.

## 7. Steps

**STEP-1 — Index + query-plan test**
Files: `core/database/migrations/000057.sql` *(new)*, `core/database/sources/notes_test.go` *(new)*. Skill: `add-catalog-migration`.
Do: add the migration (D-1). Add `TestNotesAttributionQueryPlan` (D-2).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sources/ -run TestNotesAttributionQueryPlan` passes. Temporarily commenting out the index in a local edit makes T-1 fail with `SCAN c` (do not commit that).

**STEP-2 — optionalPointer releases + honesty + tests**
Files: `core/database/deleteimpact/facets.go`, `core/database/deleteimpact/pragma_test.go`, `core/database/artifacts/artifacts.go` (comment), `core/database/artifacts/artifacts_test.go`, `core/database/likenesses/likenesses_test.go`. Skills: `add-catalog-delete`, `add-catalog-model` (read for Audited/CASCADE rule; no model FK edit), `add-audit-scope` (confirm no new entity type).
Do: implement `optionalPointer` and the two call sites (D-4). Extend honesty (D-6). Fix the `Delete` comment (D-8). Extend the two existing tests (D-9) including parity (D-5).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/deleteimpact/ ./core/database/artifacts/ ./core/database/likenesses/ ./core/database/observations/` passes; removing either `optionalPointer` call site fails `TestFacetReleaseHonesty`.

**STEP-3 — Docs**
Files: `docs/catalog-deletes.md` (policy §1 optional back-pointers; policy item 7 and checklist item 5 that today say only CASCADE needs classification — extend to `BucketOptional` `SET NULL`; § Facet release — audited CASCADE **and** optional pointers via `optionalPointer`; list the two vias; Remaining still required so SET NULL never fires on an official Delete), `.agents/skills/add-catalog-delete/SKILL.md` (same), `.agents/skills/add-catalog-model/SKILL.md` steps 3, 5, and the honesty paragraph at L51 (a deletable parent's `BucketOptional` `SET NULL` FK needs one facet release; `Audited: true` remains CASCADE-only), `.agents/skills/add-catalog-migration/SKILL.md` (optional back-pointer sentence notes the release lives in `deleteimpact`), `core/database/catalogmodel/model.go` `FK.Audited` comment (CASCADE-only, unchanged meaning), `core/database/deleteimpact/facets.go` comments (STEP-2), `docs/conclusion-layer-data-model.md` (after "Deleting the Observation clears it": official `observations.Delete` releases and audits the clear through `deleteimpact`, same fields as `SetPreferredLikeness`), `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status).
Do: per Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-24-audit-index-and-pointer-resets.md` clean; sweep below passes.

Commits: one per behavior. Suggested subjects: "Seek note authors in audit history by entity id." then "Audit cover and portrait clears when their targets are deleted." then docs. Each subject is one full imperative sentence ending in a period.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sources/notes_test.go` `TestNotesAttributionQueryPlan` *(new)* | Plan of `sqlListNotes` / `sqlGetNoteAttributed` → no detail `SCAN c` | AC-1.1, D-2 |
| T-2 | Go | `core/database/artifacts/artifacts_test.go` cover-erase path extended *(existing + `t.Run`)* | Cover delete → row cleared; release `Fields` deep-equal `SetCover` clear; non-cover delete → no `source` change | AC-2.1, AC-2.2, AC-2.3, D-5, D-9 |
| T-3 | Go | `core/database/likenesses/likenesses_test.go` "deleting the Observation clears the pointer" extended | Preferred delete → pointer NULL; release `Fields` deep-equal `SetPreferredLikeness` clear; non-preferred delete → no `canonical_entity` change | AC-3.1, AC-3.2, AC-3.3, D-5, D-9 |
| T-4 | Go | `core/database/deleteimpact/pragma_test.go` `TestFacetReleaseHonesty` (extended) | Both optional FKs have a release; removing a release fails; `Audited` remains false on them | AC-4.1, AC-4.2, D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. Suites most likely to break: `./core/database/deleteimpact/`, `./core/database/artifacts/`, `./core/database/likenesses/`, `./core/database/` (schema hash / migrate).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `rg` for "Cover SET NULL is SQLite's job" across `.agents/`, `.cursor/rules/`, `docs/*.md`, `core/` with no hit left outside this plan's Current state / Landed note; `rg` for facet-release wording that still says CASCADE-only in `docs/catalog-deletes.md`, `.agents/skills/add-catalog-delete/SKILL.md`, `.agents/skills/add-catalog-model/SKILL.md` step 5, and `facets.go` comments — each updated; `scripts/check-guidance-names.py` passes
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — N/A
- [ ] Domain doc updated: catalog-deletes.md (policy §1, item 7, checklist item 5, Facet release); conclusion-layer-data-model.md; add-catalog-delete; add-catalog-model steps 3/5 and L51 honesty line; add-catalog-migration optional-pointer sentence
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source that has at least one note. Open its Source page. Notes still show author and date; the page opens without a multi-second stall on a dogfood-sized history (bar outcome 12, second half).
2. Set an Artifact as the Source cover. Delete that Artifact (confirm allowed). The Source shows its type icon. (Engine tests own the revision assert; this checks the visible clear.)
3. On a Person with a chosen portrait, delete that likeness Observation (when delete is allowed). The Person leads with the earliest remaining likeness. (Engine tests own the revision assert.)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Query plan still shows `SCAN c` on some SQLite builds | T-1 fails; stop and report the full `EXPLAIN QUERY PLAN` output |
| `optionalPointer` field encoding drifts from `SetCover` / `SetPreferredLikeness` | D-5 parity fails |
| Cover FK clear fires `sources_clear_cover_on_primary_null` | Intended (D-4); re-read captures `cover_mode` for the audit Fields |

**Stop and report if:**

- `main` already has `000057.sql`, or highest migration is not `000056`.
- A third `BucketOptional` FK appears on `main` that this plan does not name — add its `optionalPointer` call site in a plan refresh before coding.
- `TestFacetReleaseHonesty` cannot express the optional branch without breaking a CASCADE case.
- Parity cannot compare `Fields` because `SetCover` / `SetPreferredLikeness` emit a different key set than `optionalPointer` for the same prior row — stop; do not weaken the parity assert by filtering keys.
- A likeness Observation delete is refused for a reason this plan does not cover in the only fixture path T-3 can use — report the gate; do not weaken delete policy.

## 12. Open questions

None.

## 13. PR

**Title:** S12-24 — Index note authors in audit history and audit cover and portrait resets
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-24-audit-index-and-pointer-resets.md`.
