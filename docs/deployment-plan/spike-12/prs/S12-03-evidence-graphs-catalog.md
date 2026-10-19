# S12-03 — Evidence graphs in the catalog

**Status:** Landed in [#498](https://github.com/mendahu/provenencia/pull/498) (2026-10-19)
**Landed:** No difference from the plan.
**Spike / slice:** Spike 12 · Slice 2 — Graphs in the catalog
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: schema, one Go package and the create paths; no FFI, no Swift)
**Board:** —
**Depends on:** —
**Blocks:** S12-04, S12-06, S12-07
**Branch:** `s12-03-evidence-graphs-catalog`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 and S12-02 did; the catalog format number moves to 55 on its own ([`docs/versioning.md`](../../../versioning.md#project-format-version))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-ref`](../../../../.agents/skills/add-catalog-ref/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [data model](../source-structure.md#data-model), [rules](../source-structure.md#rules), [migration](../source-structure.md#migration); working rules: "Graph data model" (all), "One bridge, one function, deleted in S12-08", "`core/database/evidencegraphs` owns graphs and links from S12-03", "Every Source has at least one graph", "Refusals are of two kinds", "Artifact ↔ graph links", "The graph list has its own effect", "A re-keyed symbol is renamed in the PR that re-keys it", "Each PR updates the docs and skills it makes false" (S12-03 line)
**Planned:** 2026-10-19 against `main` @ `231f7ed`
**Review:** [`S12-03-evidence-graphs-catalog.review.md`](S12-03-evidence-graphs-catalog.review.md) — Fits with changes / Ready with changes, 2 rounds

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

The catalog learns Evidence graphs. Every Source holds at least one graph, named Default, with its own `GPH-` ref; every Subject belongs to one graph of its Source; Artifacts and graphs are linked many-to-many. Existing projects migrate in place: each Source's Default graph holds its Subjects and is linked to all its Artifacts, so nothing a researcher sees changes. This is the foundation slice 2 and slice 3 build on ([design](../source-structure.md#multiple-evidence-graphs)), from the "Dense Evidence graphs (census-scale)" dogfood note.

**Why now.** It is the first graph PR; S12-04, S12-06 and S12-07 all start from these tables and from `core/database/evidencegraphs`.

**What stays.** Every RPC, every screen, the canvas, Promote and Copy Source behave as before. Creating a second graph is not possible yet (S12-04 adds `Create`; S12-11 / S12-12 expose it), so every Source has exactly one graph and every Subject and Artifact reaches it through `evidencegraphs.SoleGraphTx` or the "every graph" default.

## 2. User stories

**US-1.** As a researcher, when I open an existing project after this update, every Source has one Default graph holding its Subjects and linked to its Artifacts, and the project looks the same.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A v54 catalog with Source S (two Artifacts A1, A2, one Subject P with a grid position) is migrated | `evidence_graphs` has one row: `id = S.id`, `source_id = S.id`, `name = 'Default'`, `ref` NULL; `evidence_graph_artifacts` has (S.id, A1) and (S.id, A2); `subjects.graph_id` of P is `S.id`; P's `subject_positions` row is unchanged |
| AC-1.2 | That catalog is opened through `catalogsession.Do` or `onboarding.OpenCatalog` | The Default graph's `ref` is a valid `GPH-XXXXX` (`ref.Valid`); a second open leaves it unchanged |
| AC-1.3 | Any catalog after migration | `catalogmodel.TestPragmaHonesty`, the `deleteimpact` honesty tests and the `effects` completeness tests pass |

**US-2.** As the engine, every new Source starts with a Default graph, and new Subjects and Artifacts join it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `sources.Create` "Census 1881" | One `evidence_graphs` row for it: a new UUIDv7 id that is **not** the Source id, `name = 'Default'`, a valid `GPH-` ref, no description. The changes are `[source create, evidence_graph create]` |
| AC-2.2 | `subjects.Create` on that Source | The Subject's `graph_id` is the Default graph's id; `Subject.GraphID` returns it; the `subject` create change carries `graph_id` |
| AC-2.3 | `artifacts.Create` on that Source with `GraphIDs` empty | One link (Default, new Artifact); changes `[artifact create, evidence_graph_artifact create]` |
| AC-2.4 | `artifacts.Create` with `GraphIDs` = [the Default graph id] | Same result as AC-2.3 |
| AC-2.5 | `artifacts.Create` with `GraphIDs` naming a graph of another Source, an unknown id, a malformed id, or the same id twice | `evidencegraphs.ErrInvalid` (`evidencegraphs.invalid`); after `writes.Run` returns the error, no Artifact and no link exist |
| AC-2.6 | `sourcecopy.Copy` of a Source with two Artifacts, both chosen | The copy has its own Default graph (not the original's), linked to both copied Artifacts; the original's graph and links are unchanged |
| AC-2.7 | `connect.CreateCitedBridge` on a Source | The bridge Subject's `graph_id` is that Source's Default graph |

**US-3.** As the engine, Artifact links follow the link rules.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `evidencegraphs.Link(G, A)` where A is on G's Source and not linked | One `evidence_graph_artifact` create change; the row exists |
| AC-3.2 | `Link(G, A)` when already linked | No changes, no error |
| AC-3.3 | `Link(G, A)` where A is on another Source; or G or A unknown | `evidencegraphs.ErrInvalid` (`evidencegraphs.invalid`) |
| AC-3.4 | `Unlink(G, A1)` where G is linked to A1 and A2 and no Citation on A1 has an Observation on a Subject of G | One `evidence_graph_artifact` delete change; the row is gone |
| AC-3.5 | `Unlink(G, A1)` where Citation `CIT-AAAAA` on A1 has an Observation on Subject P of G | `evidencegraphs.ErrArtifactInUse` (`evidencegraphs.artifact_in_use`, user kind, no params); the link stays; `CitationsUsingArtifactTx(G, A1)` returns exactly that Citation's id |
| AC-3.6 | `Unlink(G, A1)` where A1 is G's only link and no Citation uses it | `evidencegraphs.ErrLastArtifact` (`evidencegraphs.last_artifact`, user kind); the link stays |
| AC-3.7 | `Unlink(G, A)` when not linked | No changes, no error |
| AC-3.8 | `artifacts.Delete(A1)` with no Citations | A1 and its link rows are gone; the changes are `[evidence_graph_artifact delete …, artifact delete]` |
| AC-3.9 | `sources.Delete(S)` with no Artifacts and no Subjects | S and its Default graph are gone; the changes are `[…, evidence_graph delete, source delete]` and the revision is scoped to S |

**US-4.** As the engine, graph and link writes are audited under their Source and flag the graph list.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Any write above that creates or deletes an `evidence_graph` or `evidence_graph_artifact` row | `effects.Resolve` returns `Set.EvidenceGraphs = true` and the Source in `SourceScopes`; `Set.Sources` does not gain the Source from those rows |
| AC-4.2 | A Subject create, a Citation create, a Source title edit | `Set.EvidenceGraphs` is false |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ref.PrefixEvidenceGraph = "GPH"`, reserved | STEP-1, T-1 |
| Error codes `evidencegraphs.invalid`, `evidencegraphs.artifact_in_use`, `evidencegraphs.last_artifact` | STEP-1, AC-3.3, AC-3.5, AC-3.6 |
| Migration `000055.sql`: `evidence_graphs`, `evidence_graph_artifacts`, `subjects.graph_id` with the composite FK, Default backfill | STEP-2, AC-1.1 |
| `catalogmodel` tables, FKs and `KindEvidenceGraph` | STEP-2, AC-1.3 |
| `deleteimpact`: the `subjects.graph_id` probe; facet releases for `evidence_graphs.source_id` and `evidence_graph_artifacts.artifact_id` | STEP-2, AC-1.3, AC-3.8, AC-3.9 |
| `effects`: two entries, `Effect.EvidenceGraphs` and `Set.EvidenceGraphs` | STEP-2, AC-4.1, AC-4.2 |
| `core/database/evidencegraphs`: `Graph`, `DefaultName`, errors, `CreateDefault`, `SoleGraphTx` | STEP-2, AC-2.1 |
| `sources.Create` mints Default; `subjects.InsertTx` takes its graph from `SoleGraphTx`; `subjects.Subject.GraphID` | STEP-2, AC-2.1, AC-2.2, AC-2.7 |
| `subjectpositions_test.go` moves to an external test package and creates its Subject with `subjects.Create` | STEP-2 |
| `evidencegraphs` reads `SubjectGraphTx`, `IsLinkedTx`, `CitationsUsingArtifactTx`; writes `Link`, `Unlink`, `LinkArtifact` | STEP-3, AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.6, AC-3.7 |
| `artifacts.CreateInput.GraphIDs`; `artifacts.Create` links through `LinkArtifact`; `artifacts.Delete` releases links | STEP-4, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-3.8 |
| `evidencegraphs.EnsureRefs`; `catalogsession.ReconcileOnOpen`, the one open-time repair list both open paths call | STEP-5, AC-1.2 |
| Audit and effects tests for the new entity types | STEP-6, AC-4.1, AC-4.2 |
| Docs and skills this PR makes false | STEP-8 |

| Out | Home |
| --- | --- |
| `evidencegraphs.List`, `Create` (researcher graphs), `Update`, `Delete`; the `evidence_graph_artifacts.graph_id` facet release; `KindEvidenceGraph` in `deleteimpact`'s deletable set and its projector | S12-04 |
| FFI, `GenealogyStore`, `CreateArtifactRequest.graph_ids`, the wire's `EvidenceGraphs` flag, L10n for the new codes, and the preflight read that lists the Citations behind an `artifact_in_use` refusal (over `CitationsUsingArtifactTx`, shaped for the DeleteImpact recipe) | S12-05 |
| Wording of the unlink refusals | S12-12, S12-13 (briefs S12-D3, S12-D4) |
| `requireSameGraph`; `connect` taking the bridge's graph from its endpoints; `canonicalgraph` joining on graph | S12-06 |
| `graphcache`, loaders, positions, progress keyed by graph; `Effect.Source` for graph rows | S12-07 |
| A `graph_id` on `subjects.CreateInput`; deleting `SoleGraphTx` | S12-08 |
| Promote per graph | S12-09 |
| Graphs in search (`Effect.Search`) | S12-15 |
| A `graph` audit scope type | not planned (working rule "Audit scope stays `source`") |
| One ref-mint-and-insert helper and one `uuidJSON` for every table package | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md#one-ref-mint-loop) (logged during planning) |
| One set of catalog-wide list flags on `Effect` / `Set` in place of `Vocabulary`, `PropertyUsage`, `EvidenceGraphs` bools | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md#catalog-wide-list-flags) (logged during planning) |
| Moving `insertSource` / `insertArtifact` in `core/database/deleteimpact/release_test.go` and the Source / Artifact inserts in `core/derivatives/thumbnail_test.go` onto `sources.Create` / `artifacts.Create` | not planned (D-14: import cycles; the map's working rule is narrowed) |
| Updating `core/database/subjects/schema_test.go`'s column list | not planned (it checks presence only and passes; [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md) says not to grow per-migration table tests) |
| A Default graph id on `sources.Source` | not planned (S12-04's `List` is the graph read) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Migrations | `core/database/migrations/000054.sql`; `core/database/migrate.go` `migrate` | Highest step is 54. `migrate` runs each file with one `db.Exec`, outside a transaction, so `PRAGMA foreign_keys=OFF` works inside a file |
| Table rebuild precedent | `core/database/migrations/000030.sql` | Rebuilds `artifacts`, `citations` and `property_terms` (tables with children) with `PRAGMA foreign_keys=OFF`; `CREATE TABLE x__new`; `INSERT … SELECT`; `DROP TABLE x`; `ALTER TABLE x__new RENAME TO x`; then `PRAGMA foreign_key_check`; `PRAGMA foreign_keys=ON`. `000052.sql` rebuilds `properties` the same way without `foreign_key_check`. `subjects` has never been rebuilt |
| Live `subjects` | created by `000021.sql`, indexes from later steps | `id`, `ref TEXT UNIQUE NOT NULL`, `source_id … REFERENCES sources(id)`, `subject_type_id … REFERENCES subject_types(id)`, `label`, `description`, `UNIQUE (id, subject_type_id)`; indexes `subjects_source_id_idx`, `subjects_subject_type_id_idx`. Children: `observations.subject_id`, `observations.value_subject_id`, `subject_positions.subject_id`, `identity_claims (subject_id, subject_type_id)` (`000032.sql`) |
| Catalog model | `core/database/catalogmodel/model.go` `Tables`, `FKs`, `Kind…` | No graph tables. Composite-FK precedent: `identity_claims` row with `FromCols`. `KindIdentityClaim` / `KindCanonicalEntity` are kinds that are not deletable |
| Delete probes | `core/database/deleteimpact/register.go` `inboundEdgeList`, `deletableKinds`; `probes.go` `fkEdge` | One `fkEdge` per resource FK; `TestResourceAndOwnedHonesty` (`pragma_test.go`) fails a resource FK with no probe |
| Facet releases | `core/database/deleteimpact/facets.go` `buildFacetReleases`, `rowFacet`, `colJSON` | `rowFacet(parent, table, fkCol, entityType, entityIDCol, cols…)` selects by `rowid`, deletes, returns `DeletedRow` changes. `TestFacetReleaseHonesty` requires one release per audited CASCADE whose parent kind is deletable |
| Delete refusals | [`docs/catalog-deletes.md`](../../../catalog-deletes.md) § FFI and § Impact report | A refusal is a generic code; the list of blocking rows is a separate preflight read (`GetDeleteImpact`), never refs in `apperr` params. Every `WithParams` caller today passes one fixed-position arg |
| Effects | `core/database/effects/registry.go` `registry`, `Set`, `Effect`, `Effect.validate`; `resolve.go` `Resolve` | One entry per non-skip table; `Vocabulary` is the catalog-wide flag precedent (`if eff.Vocabulary { set.Vocabulary = true }`). `validate` rejects a `None` entry that sets `Vocabulary` |
| Source create / delete | `core/database/sources/sources.go` `Create`, `Delete` | `Create` inserts the row, returns one `source` change. `Delete`: `Impact` → `Refuse` → `ReleaseFacets(KindSource)` → `DELETE` → released changes then the `source` delete change |
| Artifact create / delete | `core/database/artifacts/artifacts.go` `CreateInput`, `Create`, `Delete` | `CreateInput{SourceID, FileID, Label, Description}`; `Create` returns one `artifact` change. `Delete`: `Impact` → `Refuse` → `SnapshotOwned` → `DELETE` → `ReleaseSnapshot`; it does not call `ReleaseFacets` (artifacts have no facet yet) |
| Write rollback | `core/writes/run.go` `Run` | A write function that returns an error rolls the whole transaction back; nothing it inserted survives |
| Subject create | `core/database/subjects/subjects.go` `InsertTx`, `Create`, `Subject`, `getTx`, `scanSubject`, `sqlInsert`, `sqlGet`, `sqlGetByRef`, `Delete` | `InsertTx(tx *sql.Tx, in CreateInput)` is the only insert; `Create` and `connect.CreateCitedBridge` (`core/database/connect/connect.go`) call it. Its mint loop uses a package const `maxRefRetries = 8` and `database.IsUniqueConflict`. The create change is `FullRow` of every column; `Delete` returns a delete change with `source_id` |
| Copy Source | `core/database/sourcecopy/sourcecopy.go` `CreateFromDraft` | Calls `sources.Create`, then `artifacts.Create` per chosen Artifact. No change needed: it gets a Default graph and links through those two |
| Refs | `core/ref/ref.go` `Prefix…`, `reservedPrefixes`; `core/database/users/users.go` `EnsureRefs` | Five reserved prefixes. `users.EnsureRefs` selects rows with no ref and sets a minted one, unaudited, with no retry |
| Open paths | `core/catalogsession/session.go` `openResearcher`; `core/onboarding/ready.go` `OpenCatalog`, `createCatalog` | `openResearcher` and `OpenCatalog` repeat the same repairs in the same order: `users.EnsureRefs`, `project.EnsureUUID`, `autoreconciler.EnsureCatalog`, `searchindex.EnsureCatalog`, each closing the catalog on error; `openResearcher` then runs `namealiases.Warm`. `onboarding` imports `catalogsession`, not the reverse. `OpenCatalog` has no non-test caller. `createCatalog` runs `users.EnsureRefs` and `project.EnsureUUID`, then six `Install` calls, then `autoreconciler.EnsureCatalog` and `searchindex.EnsureCatalog`: a third copy. None of the six installers reads `users` or `project` |
| Errors | `core/apperr/apperr.go` codes | Package codes are `<package>.invalid`, `<package>.in_use`; params are ordered L10n format args |
| Direct-insert fixtures | `core/database/subjectpositions/subjectpositions_test.go` `mustPerson` (internal package, inserts a `subjects` row with ref `"CPR-"+label`; no test reads that ref); `core/database/deleteimpact/release_test.go` `insertSource`, `insertArtifact` (internal); `core/derivatives/thumbnail_test.go` (internal, Source + Artifact) | `subjects` imports `subjectpositions`, `sources` imports `deleteimpact` and `core/derivatives`, so none of these internal test packages can import the create functions |
| Tests that create Subjects | `core/database/effects/{conclusion_handles,evidence_handles,handle_paths}_test.go`, `core/database/deleteimpact/{conclusion_release,conclusion,impact,pins_e2e}_test.go` and most packages under `core/database` | They call `subjects.Create`, so they fail the moment `subjects.graph_id` is `NOT NULL` and nothing fills it |
| Citation fixture precedent | `core/database/effects/evidence_handles_test.go` `TestEvidenceHandleSets` | `citations.CreateWithObservations` with a person Subject and the `name` property (`namevalue.Value{Form: "Ada"}`) |
| Tests | `core/database/sources/sources_test.go` `TestSources`, `TestSourceDelete`; `core/database/artifacts/artifacts_test.go` `TestArtifacts`, `TestArtifactDelete`; `core/database/subjects/subjects_test.go` `TestSubjects`; `core/database/sourcecopy/sourcecopy_test.go` `TestCopy`; `core/database/connect/connect_test.go` `TestCreateCitedBridge`; `core/database/audit/scopes_test.go` `TestSourceScopes`; `core/database/effects/registry_completeness_test.go` `TestSourceEvidenceIsNotActivity` (package `effects`, internal); `core/database/effects/evidence_handles_test.go` (package `effects_test`); `core/ref/ref_test.go` `TestMint`, `TestValidatePrefix`; `core/onboarding/ready_test.go`; `core/catalogsession/session_test.go` `TestDo` | Table-driven; each builds a temp catalog with `database.Create` |

**Precedent.** No single earlier change has this shape. Copy these parts:

- The `subjects` rebuild: `000030.sql` (tables with children rebuilt with FKs off, `foreign_key_check`), with the indexes recreated by hand because a rebuild drops them ([`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md) step 3).
- `EnsureRefs`: `users.EnsureRefs`.
- The composite FK registration: `identity_claims` in `catalogmodel.FKs` (`Column` = leading column, `FromCols` = the tuple).
- Facet releases: the `source_metadata_layout` `rowFacet` (a row audited under a non-`id` column) and `sources.Delete` (`ReleaseFacets` before the `DELETE`, released changes first).
- A link row audited under its parent's id: `identity_claim_evidence` pins (`pinRelease`, `EntityID` = claim id).

The SQL in §6 Schema was run during planning against a v54 catalog holding a Source, two Artifacts, a Subject and its position: the backfill and positions were right, `foreign_key_check` was clean, and a Subject pointing at another Source's graph was refused.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Migration `000055.sql` is exactly the SQL in §6 Schema: the two tables and their indexes first, the backfill `INSERT`s, then the `subjects` rebuild with `graph_id BLOB NOT NULL` after `source_id` and `FOREIGN KEY (graph_id, source_id) REFERENCES evidence_graphs(id, source_id)` (no `ON DELETE`, so `NO ACTION`), its two indexes recreated plus `subjects_graph_id_idx`. Backfilled Default graphs have `id = source_id`, `name = 'Default'`, `ref` NULL, `description` NULL | Working rules "Tables", "Refs are minted in Go", "The migration backfills in SQL". SQLite cannot add a `NOT NULL` column with a foreign key through `ALTER TABLE`, so `subjects` is rebuilt | `ALTER TABLE subjects ADD COLUMN graph_id` — must be nullable with a NULL default; a trigger to fill it — hidden logic |
| D-2 | A Default graph made by `sources.Create` (and every later graph) gets a fresh UUIDv7 id. Only backfilled graphs share their Source's id | Distinct ids make a Source-for-graph mix-up fail in tests before S12-08 moves ids onto the wire | `id = source_id` for every Default — hides those mix-ups in every new-project test |
| D-3 | `catalogmodel`: `KindEvidenceGraph Kind = "evidence_graph"` *(new)* beside the conclusion kinds with the comment "registered for its inbound probe; delete path in S12-04". Tables: `{Name: "evidence_graphs", PK: "id", Kind: KindEvidenceGraph, Bucket: BucketFacet}`, `{Name: "evidence_graph_artifacts", Bucket: BucketFacet}`. FKs: `evidence_graphs.source_id → sources CASCADE, BucketFacet, Audited`; `evidence_graph_artifacts.graph_id → evidence_graphs CASCADE, BucketFacet, Audited`; `evidence_graph_artifacts.artifact_id → artifacts CASCADE, BucketFacet, Audited`; `subjects.graph_id` with `FromCols: {"graph_id", "source_id"} → evidence_graphs NO ACTION, BucketResource` | Working rules "A graph is part of its Source's record" and "A link row is part of both its graph and its Artifact". Graph and link rows are audited, so their cascades are audited facets ([`catalog-deletes.md`](../../../catalog-deletes.md): "A silent CASCADE is only for rows that are never audited on create either"). A Subject blocks its graph's delete, so that FK is a resource edge | `BucketResource` for the `evidence_graphs` table — nothing blocks on it except through the probe; a silent link cascade — links are audited |
| D-4 | `deleteimpact`: add `fkEdge(catalogmodel.KindEvidenceGraph, "subjects.graph_id", catalogmodel.KindSubject, "subjects", "graph_id", "ref")` to `inboundEdgeList` after the two `KindSource` edges. Add two `rowFacet` entries to `buildFacetReleases`: in the Source group, `rowFacet(catalogmodel.KindSource, "evidence_graphs", "source_id", "evidence_graph", "id", col{"id", colUUID}, col{"ref", colText}, col{"source_id", colUUID}, col{"name", colText}, col{"description", colText})`; in the Interpretation group, `rowFacet(catalogmodel.KindArtifact, "evidence_graph_artifacts", "artifact_id", "evidence_graph_artifact", "graph_id", col{"graph_id", colUUID}, col{"artifact_id", colUUID})`. `KindEvidenceGraph` is **not** added to `deletableKinds` | The honesty tests demand the probe for every resource FK and a release for each audited CASCADE whose parent is deletable (Source, Artifact). The `graph_id` release waits for the graph delete (S12-04). A Source is deletable only with no Artifacts and no Subjects, so its graphs have no links and no Subjects by then | Leaving the Source's graph to the SQLite CASCADE — the delete would not be audited |
| D-5 | `effects`: `Effect.EvidenceGraphs bool` *(new)* and `Set.EvidenceGraphs bool` *(new)*, each with a one-line comment ("the graph list changed: a graph or link row was written"). `Resolve` sets `set.EvidenceGraphs = true` when an entry has it, beside `Vocabulary`. `validate` rejects a `None` entry with `EvidenceGraphs`. Entries: `"evidence_graphs": {Entity: "evidence_graph", SourceScope: field("source_id"), EvidenceGraphs: true}` and `"evidence_graph_artifacts": {Entity: "evidence_graph_artifact", SourceScope: up("graph_id", "source_id"), EvidenceGraphs: true}`. No `Source`, `CitationSource`, `Handles` or `Search` | Working rules "The graph list has its own effect" and "Audit scope stays `source`". No read loads these tables until S12-05 / S12-07, so they drop no Source graph (`Effect.Source`) | A `graph` scope type; `Source: field("source_id")` — would drop the Source's cached graph on every link write for nothing; one list-flag set for all catalog-wide lists — logged (Out) |
| D-6 | Audit change shapes. Graph create: `EntityType "evidence_graph"`, `EntityID` = graph id, `rowchange.FullRow` of `"id"`, `"ref"`, `"source_id"`, `"name"`, `"description"` (`nil` when empty, so JSON null). Link create / delete: `EntityType "evidence_graph_artifact"`, `EntityID` = graph id, `FullRow` / `DeletedRow` of `"graph_id"`, `"artifact_id"`. UUIDs as strings through a package-local `uuidJSON`, as `subjects` has | Matches the `rowFacet` delete changes D-4 produces, so create and delete of one row read alike in history; the pin precedent audits a link under its parent's id | `EntityID` = artifact id — a graph's history is the more useful read |
| D-7 | `core/database/evidencegraphs` *(new)* is one file `evidencegraphs.go` (plus `evidencegraphs_test.go` in package `evidencegraphs_test`). It imports only `core/apperr`, `core/database`, `core/database/rowchange`, `core/ref`, `github.com/google/uuid` and the standard library. API: `Graph{ID []byte; Ref string; SourceID []byte; Name, Description string}`; `const DefaultName = "Default"`; `ErrInvalid`, `ErrArtifactInUse`, `ErrLastArtifact`; `CreateDefault`, `SoleGraphTx`, `SubjectGraphTx`, `IsLinkedTx`, `CitationsUsingArtifactTx`, `Link`, `Unlink`, `LinkArtifact`, `EnsureRefs` (signatures in §6 Go core). All SQL on `evidence_graphs` and `evidence_graph_artifacts` outside migrations, `deleteimpact` and tests lives in this package | Working rule "`core/database/evidencegraphs` owns graphs and links from S12-03". A small import set lets `sources`, `artifacts`, `subjects` and later `observations` import it with no cycle | Putting the Default insert in `sources` — S12-04's `Create` would copy it |
| D-8 | `CreateDefault(tx *database.Tx, sourceID []byte) (Graph, []rowchange.Change, error)` inserts through an unexported `insert(tx *sql.Tx, sourceID []byte, name, description string) (Graph, rowchange.Change, error)` whose mint loop is a copy of `subjects.InsertTx`'s: a package const `maxRefRetries = 8`, `ref.Mint(ref.PrefixEvidenceGraph)`, retry while `database.IsUniqueConflict(err)`, `ErrInvalid` after the last attempt. `CreateDefault` does not check the user id; its only caller is `sources.Create`, which already did. `sources.Create` calls it right after its own insert and returns `[source change, graph change]` | One insert for S12-04's `Create` to reuse. The loop's 7th copy is logged (Out) rather than refactored across eight packages here | An exported `Insert` — no second caller yet |
| D-9 | `SoleGraphTx(tx *sql.Tx, sourceID []byte) ([]byte, error)` reads through an unexported `graphIDsTx(tx *sql.Tx, sourceID []byte) ([][]byte, error)` (`SELECT id FROM evidence_graphs WHERE source_id = ? ORDER BY id`), the one per-Source graph-id read, which `LinkArtifact` also uses. One row → its id. No row → `ErrInvalid` (unknown Source, as `subjects.requireSource` answers). Two rows → `fmt.Errorf("evidencegraphs: source %x has more than one graph", sourceID)` (internal; not reachable before S12-11). Its doc comment says: "The one bridge between S12-03 and S12-08. S12-08 deletes it." | Working rule "One bridge, one function" | Returning the first graph — would hide a second graph silently |
| D-10 | `subjects.InsertTx` calls `evidencegraphs.SoleGraphTx(tx, in.SourceID)` after `requireSource` and writes the id to the new `graph_id` column. `sqlInsert`, `sqlGet`, `sqlGetByRef` and `scanSubject` gain `graph_id`; `Subject` gains `GraphID []byte` after `SourceID`; the create `FullRow` and the delete change gain `"graph_id"`. `CreateInput` does not change | Working rule: "Subject creation uses `SoleGraphTx`". `InsertTx` is the one insert, so `Create` and `connect` both get it. The column belongs on the row struct | Calling `SoleGraphTx` in `Create` and `connect` separately — two call sites to delete in S12-08 |
| D-11 | `artifacts.CreateInput` gains `GraphIDs [][]byte` *(new)* with the comment "empty = every graph of the Source". After its insert, `artifacts.Create` calls `evidencegraphs.LinkArtifact(tx, userID, in.SourceID, newID, in.GraphIDs)` and appends its changes after the `artifact` change, returning its error unchanged. `artifacts` holds no graph SQL. `LinkArtifact(tx *database.Tx, userID, sourceID, artifactID []byte, graphIDs [][]byte) ([]rowchange.Change, error)`: empty `graphIDs` → every graph of `sourceID` (`graphIDsTx`; a Source with no graph, which only direct-SQL test fixtures make, gets no link and no error); otherwise a 15-byte or repeated id → `ErrInvalid`; then `Link` per graph in the given order, which refuses an unknown graph or one of another Source. A refusal rolls back the Artifact insert with the rest of the write (`writes.Run`) | Working rule "Artifact ↔ graph links": Artifact create goes through `evidencegraphs`. One home for "a graph belongs to this Artifact's Source" (`Link`); S12-05's `graph_ids` and S12-13 reuse the same entry. "Empty", not "nil": a proto3 `repeated` field decodes empty as nil, so S12-05 cannot tell them apart | Validating `GraphIDs` in `artifacts` with its own graph query — a third per-Source graph read in another package; `nil` vs `[]` meaning different things |
| D-12 | `Link(tx *database.Tx, userID, graphID, artifactID []byte) ([]rowchange.Change, error)`: `tx` nil, an id that is not 16 bytes, or a `database.RequireUserID` failure → `ErrInvalid`; then one read `SELECT g.source_id, a.source_id FROM evidence_graphs g, artifacts a WHERE g.id = ? AND a.id = ?` — no row → `ErrInvalid`, the two differ → `ErrInvalid`; already linked (`IsLinkedTx`) → `nil, nil`; else insert and return one create change. `Unlink(tx *database.Tx, userID, graphID, artifactID []byte) ([]rowchange.Change, error)`: the same checks; not linked (`IsLinkedTx`) → `nil, nil`; then, in this order, `CitationsUsingArtifactTx` non-empty → `ErrArtifactInUse`; the graph has exactly one link → `ErrLastArtifact`; else delete and return one delete change. Neither error carries params | "Unchanged → no changes" is the `artifacts.Update` rule. Citations first because that refusal tells the researcher what to fix. A refusal is a plain code and the blocking list is a separate read, as for every delete ([`catalog-deletes.md`](../../../catalog-deletes.md) § FFI); S12-05 adds that read | Refs in `apperr` params — a variable-length param list no other code uses, and a second "blocked by these rows" shape beside `deleteimpact.Group`; erroring on a repeated link — the checkbox UI (S12-13) sends the full state |
| D-13 | `CitationsUsingArtifactTx(tx *sql.Tx, graphID, artifactID []byte) ([][]byte, error)` returns the ids of the distinct Citations on that Artifact with at least one Observation whose Subject is in that graph, ordered by `ref COLLATE NOCASE`: `SELECT DISTINCT c.id, c.ref FROM citations c JOIN observations o ON o.citation_id = c.id JOIN subjects s ON s.id = o.subject_id WHERE c.artifact_id = ? AND s.graph_id = ? ORDER BY c.ref COLLATE NOCASE` (the ref is selected for the order and dropped). Bad input → `nil, ErrInvalid` | Working rule: "through Observations → Subjects". A Citation with no Observation belongs to no graph and does not block. Ids only: S12-05's preflight projects them for the DeleteImpact recipe, so no row type is added here | A `CitationRef` type — a near-copy of `deleteimpact`'s listed row; every Citation on the Artifact — would block an unlink for another graph's Citations |
| D-14 | Fixtures: `subjectpositions_test.go` becomes `package subjectpositions_test` (it uses no unexported name), calls `subjectpositions.Set` / `Get` qualified, and `mustPerson` keeps its `(t, c, sourceID, label)` signature and creates its Subject with `subjects.Create(tx, userID, subjects.CreateInput{SourceID: sourceID, SubjectTypeID: person.ID, Label: label}, nil)` inside `writes.Run`, using the enclosing test's `userID`; its ref is minted. `core/database/deleteimpact/release_test.go` and `core/derivatives/thumbnail_test.go` keep their direct `sources` / `artifacts` inserts: they insert no Subject, the schema still accepts them, and their packages cannot import `sources` or `artifacts`. The map's working rule "The migration backfills in SQL" was narrowed to say so with this plan | The rule's goal is fixtures that match the schema; a Source with no graph is harmless where nothing creates a Subject | Moving those two files to external packages — both use unexported helpers |
| D-15 | `EnsureRefs(c *database.Catalog) error` has the shape of `users.EnsureRefs`: select ids with `ref IS NULL`, then per row `ref.Mint(ref.PrefixEvidenceGraph)` and `UPDATE evidence_graphs SET ref = ? WHERE id = ?`, unaudited, no retry (as `users`). `catalogsession.ReconcileOnOpen(c *database.Catalog) error` *(new)* runs the open-time repairs once, in this order: `users.EnsureRefs`, `evidencegraphs.EnsureRefs`, `project.EnsureUUID`, `autoreconciler.EnsureCatalog`, `searchindex.EnsureCatalog`, returning the first error and not closing `c`. `openResearcher` calls it after `database.Open` (closing `c` on error, as today) and then `namealiases.Warm`; `onboarding.OpenCatalog` calls it after `database.Open` (closing on error). `createCatalog` becomes `database.Create`, the six `Install` calls in today's order, then `ReconcileOnOpen` (closing on error) | Working rule "Refs are minted in Go, never in SQL": refs are minted at open, beside `users.EnsureRefs`. One list of open repairs means the next one is one line, not two copies to keep in step. `onboarding` already imports `catalogsession`; the comment that `Do` uses "the same open policy as `OpenCatalog`" becomes true by construction | Adding the line to both copies — a third repair duplicated; auditing the backfill — a format repair is not research |
| D-16 | `artifacts.Delete` calls `deleteimpact.ReleaseFacets(tx, catalogmodel.KindArtifact, id)` after `Refuse` and before `SnapshotOwned`, and returns the released changes before its own delete change, as `sources.Delete` does | Links are audited facets (D-3); an unreleased link would go by the silent CASCADE backstop, and `ReleaseFacets` is the one call | Deleting links by hand in `artifacts` |
| D-17 | Error codes *(new)* in `core/apperr/apperr.go`: `CodeEvidenceGraphsInvalid = "evidencegraphs.invalid"`, `CodeEvidenceGraphsArtifactInUse = "evidencegraphs.artifact_in_use"`, `CodeEvidenceGraphsLastArtifact = "evidencegraphs.last_artifact"`, all `KindUser`. No `L10n.Errors` entry in this PR | Working rule "Refusals are of two kinds": domain invariants with their own user-kind code. Nothing reaches the app until S12-05, which maps them | `KindConflict` like `*.in_use` — the working rule says user kind |
| D-18 | `ref.PrefixEvidenceGraph = "GPH"` *(new)* in the reserved const block and `reservedPrefixes` | Working rule "Ref prefix `GPH`" | — |
| D-19 | No FFI, proto, Swift or `VERSION` change | Out lines S12-05 and later | — |
| D-20 | Test reads: tests read graph and link rows with SQL through `c.DB()` after the write commits (`SELECT id, ref, name, COALESCE(description, '') FROM evidence_graphs WHERE source_id = ?`, `SELECT artifact_id FROM evidence_graph_artifacts WHERE graph_id = ?`). A test that needs a graph id inside a write takes it from `SoleGraphTx(tx.Tx, sourceID)` in that write's function. Each test file that needs it declares its own small `defaultGraphID(t, c, sourceID) []byte` helper over the first query; no shared test package is added. `SubjectGraphTx` and `IsLinkedTx` answer bad input with a zero value and `ErrInvalid` | `List` is S12-04; tests use the schema directly, as `TestSourceDelete` and `TestSubjectPositions` do | A test-only export from `evidencegraphs` |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/evidencegraphs` *(new)* | package | `core/database/users` (ref + `EnsureRefs`), `core/database/subjectpositions` (small table package) | A new table needs its own package ([`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md)); the working rule names it |
| `evidencegraphs.Graph` *(new)* | type | `sources.Source`, `subjects.Subject` | One row struct per table; S12-04 extends it with linked Artifact ids |
| `evidencegraphs.EnsureRefs` *(new)* | func | `users.EnsureRefs` | Same shape, other table and prefix; a shared helper would need a table parameter for two callers |
| `evidencegraphs.SoleGraphTx` *(new)* | func | — | The bridge the working rule names; deleted in S12-08 |
| `evidencegraphs.SubjectGraphTx`, `IsLinkedTx` *(new)* | func | `subjects.Subject.GraphID` (this PR) | Membership reads the working rule puts in `evidencegraphs`, so `observations` (S12-06) can read them without importing `subjects` (which imports `autoreconciler`, `deleteimpact`, `conclusionheaders`) |
| `evidencegraphs.CitationsUsingArtifactTx` *(new)* | func | `deleteimpact` `fkEdge(KindArtifact, "citations.artifact_id", …)` | That probe counts every Citation on the Artifact; this one is per graph. Returns ids, no new row type |
| `evidencegraphs.LinkArtifact` *(new)* | func | `evidencegraphs.Link` (this PR) | The one entry for "link a new Artifact to these graphs, or all of them"; `artifacts.Create` now, S12-05's `graph_ids` later |
| `catalogsession.ReconcileOnOpen` *(new)* | func | `openResearcher`, `onboarding.OpenCatalog`, `onboarding.createCatalog` bodies | Replaces their three copies of the repair list |
| `Effect.EvidenceGraphs`, `Set.EvidenceGraphs` *(new)* | field | `Effect.Vocabulary`, `Set.Vocabulary` | Same catalog-wide flag shape for another list; the working rule names it. A shared list-flag set is logged (Out) |
| `artifacts.CreateInput.GraphIDs` *(new)* | field | — | The working rule names it |
| `subjects.Subject.GraphID` *(new)* | field | `Subject.SourceID` | The row's new column |
| `KindEvidenceGraph` *(new)* | const | `KindIdentityClaim` (a kind that is not deletable) | Needed as the probe's parent |
| `ref.PrefixEvidenceGraph` *(new)* | const | `ref.PrefixObservation` | — |
| Three `CodeEvidenceGraphs…` *(new)* | const | `CodeArtifactsInvalid`, `CodeSourcesInUse` | One code per refusal; the working rule wants distinct user-kind codes, and [`catalog-deletes.md`](../../../catalog-deletes.md) forbids a private `sqlInUse` |
| `evidencegraphs_test.go` *(new)* | test | `subjects/subjects_test.go` table style | External package so it can use `sources.Create` / `artifacts.Create` |
| `TestEvidenceGraphsResolve` *(new)* | test | `evidence_handles_test.go` (package `effects_test`) | `Resolve` over real rows needs `sources` / `artifacts`, which import `effects`; the internal test file cannot |

**Builds on.** `catalogmodel` + `deleteimpact` registries (probe, `rowFacet`), the `effects` registry and `Resolve`, `core/ref`, `writes.Run` rollback, the `users.EnsureRefs` open hook.
**Deletes.** The repair sequence copied in `catalogsession.openResearcher`, `onboarding.OpenCatalog` and `onboarding.createCatalog` (all three call `ReconcileOnOpen`). `SoleGraphTx` is born to die in S12-08.
**Next adds.** S12-04 adds `List` (absorbing `LinkArtifact`'s every-graph read), `Create` (through `insert` and `Link`), `Update`, `Delete`, the `evidence_graph_artifacts.graph_id` `rowFacet` with parent `KindEvidenceGraph`, and puts `KindEvidenceGraph` in `deletableKinds` with a projector — each a registry entry. S12-05 puts `Set.EvidenceGraphs` on the wire in `api/ffi/handlers/write_effects.go`, passes `graph_ids` to `LinkArtifact`, and adds the unlink preflight over `CitationsUsingArtifactTx`. S12-06 calls `SubjectGraphTx` / `IsLinkedTx`. The next open-time repair is one line in `ReconcileOnOpen`.

### Schema

Migration `core/database/migrations/000055.sql` *(new)*:

```sql
-- S12-03: Evidence graphs. A Source holds one or more graphs; each Subject
-- belongs to one graph of its Source (the composite key keeps it on its home
-- Source). Artifacts and graphs are linked many-to-many. Every existing Source
-- gets a Default graph whose id is the Source's id; its GPH- ref is minted in
-- Go at open (evidencegraphs.EnsureRefs). Every Subject moves into it and every
-- Artifact is linked to it.

CREATE TABLE evidence_graphs (
	id          BLOB PRIMARY KEY,
	ref         TEXT UNIQUE,
	source_id   BLOB NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
	name        TEXT NOT NULL,
	description TEXT,

	UNIQUE (id, source_id)
) STRICT;

CREATE INDEX evidence_graphs_source_id_idx ON evidence_graphs(source_id);

CREATE TABLE evidence_graph_artifacts (
	graph_id    BLOB NOT NULL REFERENCES evidence_graphs(id) ON DELETE CASCADE,
	artifact_id BLOB NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,

	PRIMARY KEY (graph_id, artifact_id)
) STRICT;

CREATE INDEX evidence_graph_artifacts_artifact_id_idx ON evidence_graph_artifacts(artifact_id);

INSERT INTO evidence_graphs (id, source_id, name)
SELECT id, id, 'Default' FROM sources;

INSERT INTO evidence_graph_artifacts (graph_id, artifact_id)
SELECT source_id, id FROM artifacts;

PRAGMA foreign_keys=OFF;

CREATE TABLE subjects__new (
	id              BLOB PRIMARY KEY,
	ref             TEXT UNIQUE NOT NULL,
	source_id       BLOB NOT NULL REFERENCES sources(id),
	graph_id        BLOB NOT NULL,
	subject_type_id BLOB NOT NULL REFERENCES subject_types(id),
	label           TEXT,
	description     TEXT,

	UNIQUE (id, subject_type_id),
	FOREIGN KEY (graph_id, source_id) REFERENCES evidence_graphs(id, source_id)
) STRICT;

INSERT INTO subjects__new (id, ref, source_id, graph_id, subject_type_id, label, description)
SELECT id, ref, source_id, source_id, subject_type_id, label, description FROM subjects;

DROP TABLE subjects;
ALTER TABLE subjects__new RENAME TO subjects;

CREATE INDEX subjects_source_id_idx ON subjects(source_id);
CREATE INDEX subjects_graph_id_idx ON subjects(graph_id);
CREATE INDEX subjects_subject_type_id_idx ON subjects(subject_type_id);

PRAGMA foreign_key_check;
PRAGMA foreign_keys=ON;
```

Registered in `catalogmodel` per D-3. Delete class: graph = audited facet of its Source; link = audited facet of its graph and of its Artifact; `subjects → evidence_graphs` = resource (D-4). Audit: `source` scope through `source_id` / `graph_id → source_id` (D-5). The schema hash is derived at init; nothing to update ([`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md#schema-hash)).

### Go core

| Package | Change |
| --- | --- |
| `core/ref` | `PrefixEvidenceGraph = "GPH"` *(new)*, reserved |
| `core/apperr` | three codes (D-17) |
| `core/database/catalogmodel` | `KindEvidenceGraph`, two `Tables` rows, four `FKs` rows (D-3) |
| `core/database/deleteimpact` | one probe, two `rowFacet` entries (D-4) |
| `core/database/effects` | `Effect.EvidenceGraphs`, `Set.EvidenceGraphs`, two entries, `Resolve`, `validate` (D-5) |
| `core/database/evidencegraphs` *(new)* | below |
| `core/database/sources` | `Create` calls `evidencegraphs.CreateDefault(tx, idBytes)` and appends its change (D-8) |
| `core/database/subjects` | `InsertTx` calls `SoleGraphTx`; `graph_id` in SQL, scan, struct, changes (D-10) |
| `core/database/artifacts` | `CreateInput.GraphIDs`; `Create` calls `LinkArtifact`; `Delete` releases facets (D-11, D-16) |
| `core/catalogsession` | `ReconcileOnOpen` *(new)*; `openResearcher` calls it (D-15) |
| `core/onboarding` | `OpenCatalog` and `createCatalog` call `catalogsession.ReconcileOnOpen` (D-15) |

`core/database/evidencegraphs/evidencegraphs.go` *(new)*:

```go
// Package evidencegraphs owns Evidence graphs and their Artifact links.
package evidencegraphs

const DefaultName = "Default"

var (
	ErrInvalid       = apperr.New(apperr.CodeEvidenceGraphsInvalid, apperr.KindUser)
	ErrArtifactInUse = apperr.New(apperr.CodeEvidenceGraphsArtifactInUse, apperr.KindUser)
	ErrLastArtifact  = apperr.New(apperr.CodeEvidenceGraphsLastArtifact, apperr.KindUser)
)

type Graph struct {
	ID          []byte
	Ref         string
	SourceID    []byte
	Name        string
	Description string
}

func CreateDefault(tx *database.Tx, sourceID []byte) (Graph, []rowchange.Change, error)
func SoleGraphTx(tx *sql.Tx, sourceID []byte) ([]byte, error)
func SubjectGraphTx(tx *sql.Tx, subjectID []byte) ([]byte, error)                 // unknown Subject → ErrInvalid
func IsLinkedTx(tx *sql.Tx, graphID, artifactID []byte) (bool, error)
func CitationsUsingArtifactTx(tx *sql.Tx, graphID, artifactID []byte) ([][]byte, error)
func Link(tx *database.Tx, userID, graphID, artifactID []byte) ([]rowchange.Change, error)
func Unlink(tx *database.Tx, userID, graphID, artifactID []byte) ([]rowchange.Change, error)
func LinkArtifact(tx *database.Tx, userID, sourceID, artifactID []byte, graphIDs [][]byte) ([]rowchange.Change, error)
func EnsureRefs(c *database.Catalog) error
```

Each exported read checks `tx != nil` and 16-byte ids and returns its zero value and `ErrInvalid` otherwise. SQL lives as unexported consts in the same file.

Errors: `evidencegraphs.invalid` → bad input, unknown Source / graph / Artifact / Subject, an Artifact on another Source, a bad `GraphIDs`. `evidencegraphs.artifact_in_use` → `Unlink` while the graph's Citations use the Artifact. `evidencegraphs.last_artifact` → `Unlink` of the graph's only link. None carries params.

### FFI

None — S12-05.

### macOS

None — no Swift change; the app reads nothing new.

### Strings

None — the three codes get `L10n.Errors` entries in S12-05.

## 7. Steps

In every step, a test that fails **only** because a Source create now also returns an `evidence_graph` change, an Artifact create an `evidence_graph_artifact` change, a Source or Artifact delete a released graph or link change, or a revision now holds more `audit_changes` rows, is updated in that step to the new count or list. Any other failure is a Stop condition.

**STEP-1 — Ref prefix and error codes**
Files: `core/ref/ref.go`, `core/ref/ref_test.go`, `core/apperr/apperr.go`. Skill: `add-catalog-ref`.
Do: D-18 and D-17. Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/ref/ ./core/apperr/` passes. Commit.

**STEP-2 — Schema, registries, Default graph, Subjects join it**
Files: `core/database/migrations/000055.sql` *(new)*, `core/database/catalogmodel/model.go`, `core/database/deleteimpact/register.go`, `core/database/deleteimpact/facets.go`, `core/database/effects/registry.go`, `core/database/effects/resolve.go`, `core/database/effects/registry_completeness_test.go`, `core/database/evidencegraphs/evidencegraphs.go` *(new)*, `core/database/evidencegraphs/evidencegraphs_test.go` *(new)*, `core/database/sources/sources.go`, `core/database/subjects/subjects.go`, `core/database/subjectpositions/subjectpositions_test.go`, `core/database/sources/sources_test.go`, `core/database/subjects/subjects_test.go`, `core/database/connect/connect_test.go`, and any `api/...` or `core/...` test that fails only under the §7 rule. Skills: `add-catalog-migration`, `add-catalog-model`, `add-catalog-delete`, `add-audit-scope`, `add-catalog-query`.
Do: §6 Schema verbatim; D-3, D-4, D-5. The package with `Graph`, `DefaultName`, the three errors, `insert`, `CreateDefault`, `SoleGraphTx` (D-7, D-8, D-9). D-10 in `subjects`. `sources.Create` per D-8. D-14 for `subjectpositions_test.go`. Add T-2, T-3, T-4, T-5, T-6, T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` and `go vet -tags fts5 ./api/... ./core/...` pass. Commit: "Give every Source an Evidence graph that holds its Subjects."

**STEP-3 — Membership reads and links**
Files: `core/database/evidencegraphs/evidencegraphs.go` *(new)*, `core/database/evidencegraphs/evidencegraphs_test.go` *(new)* (both created in STEP-2; this step adds to them).
Do: `SubjectGraphTx`, `IsLinkedTx`, `CitationsUsingArtifactTx`, `Link`, `Unlink`, `LinkArtifact` (D-11, D-12, D-13, D-20). Add T-8, T-9, T-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs` passes. Commit.

**STEP-4 — Artifacts link on create and release on delete**
Files: `core/database/artifacts/artifacts.go`, `core/database/artifacts/artifacts_test.go`, `core/database/sourcecopy/sourcecopy_test.go`.
Do: D-11 (the `artifacts` half) and D-16. Add T-11, T-12, T-13.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes. Commit.

**STEP-5 — One open-time repair list that mints missing graph refs**
Files: `core/database/evidencegraphs/evidencegraphs.go` *(new)*, `core/database/evidencegraphs/evidencegraphs_test.go` *(new)* (both created in STEP-2; this step adds to them), `core/catalogsession/session.go`, `core/catalogsession/session_test.go`, `core/onboarding/ready.go`, `core/onboarding/ready_test.go`. Skills: `add-catalog-ref`, `use-catalog-session`.
Do: D-15. Add T-14, T-15, T-16.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs ./core/onboarding ./core/catalogsession` passes. Commit.

**STEP-6 — Audit scopes and the graph-list flag**
Files: `core/database/audit/scopes_test.go`, `core/database/effects/registry_completeness_test.go`, `core/database/effects/evidence_graphs_test.go` *(new)*. Skill: `add-audit-scope`.
Do: T-17, T-18, T-19.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/audit ./core/database/effects` passes. Commit.

**STEP-7 — Whole suite**
Files: none, unless the run finds a count or list assertion the earlier steps missed (fixed under the rule above).
Do: run the whole Go suite and `scripts/check-pr.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` and `go vet -tags fts5 ./api/... ./core/...` pass; `scripts/check-pr.sh` prints `check-pr: PASS` on macOS.

**STEP-8 — Docs and skills**
Files:
- `docs/source-layer-data-model.md`: a new section after § 7 `artifacts` for `evidence_graphs` and `evidence_graph_artifacts`, with the rules from [Rules](../source-structure.md#rules) that the schema and `evidencegraphs` hold today; both tables in § 10 Current schema and § 2 Relationship overview.
- `docs/interpretation-layer-data-model.md` § 4.2: `graph_id`, the composite key, "a Subject belongs to one graph and never moves".
- `docs/catalog-refs.md`: the `GPH` row and the `EnsureRefs` note.
- `docs/data-model-source-interpretation-conclusion.md` § 2: `GPH` in the prefix list and the reserved list.
- `docs/catalog-deletes.md`: the three audited cascades in the Audited column; the `subjects.graph_id` resource edge.
- `core/database/effects/README.md`: `Set.EvidenceGraphs`.
- `.agents/skills/add-catalog-ref/SKILL.md`: the prefix table.
- `.agents/skills/use-catalog-session/SKILL.md`: `Do` opens through `ReconcileOnOpen`, the same repairs as `OpenCatalog`.
- `.agents/skills/add-seeded-vocabulary/SKILL.md` "Where create vs open run": the `OpenCatalog` and `Do` lines name `ReconcileOnOpen`.
- `.agents/skills/add-catalog-model/SKILL.md` step 6: the `Effect` field list matches `effects.Effect` (`Source`, `CitationSource`, `SourceScope`, `Handles`, `HandleScope`, `Search`, `Vocabulary`, `PropertyUsage`, `EvidenceGraphs`, `Structure`).
- `docs/deployment-plan/spike-12/deployment-plan.md`: map row status (the working-rule edits for D-11, D-12 and D-14 were made with this plan).
- `docs/deployment-plan/spike-12/completed.md`; this plan (status, Landed note).

Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-03-evidence-graphs-catalog.md` reports no errors; links resolve. Commit.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits; every commit passes its step's check.

## 8. Tests

All Go, table-driven (`.cursor/rules/go-tests.mdc`). New `evidencegraphs` tests are in `package evidencegraphs_test` and build fixtures with `sources.Create`, `artifacts.Create`, `subjects.Create` inside `writes.Run`; graph and link rows are read per D-20. A Citation that uses an Artifact is made the way `TestEvidenceHandleSets` does: `citations.CreateWithObservations` with a person Subject, the `name` property and `namevalue.Value{Form: "Ada"}` (`subjectvocab.Install` first), with `observationstest.LikenessLocator` as the locator.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/ref/ref_test.go` `TestValidatePrefix`, `TestMint` | `"GPH"` → `ErrReservedPrefix`; `TestMint` row `PrefixEvidenceGraph` → `ref.Valid` | D-18 |
| T-2 | Go | `core/database/effects/registry_completeness_test.go` `TestSourceEvidenceIsNotActivity` | rows `{"evidence_graphs", false, true, false}`, `{"evidence_graph_artifacts", false, true, false}` | D-5 |
| T-3 | Go | `core/database/evidencegraphs/evidencegraphs_test.go` `TestCreateDefault` *(new)* | `sources.Create` "Census 1881" → one graph: name `Default`, `ref.Valid(ref)` with prefix `GPH`, id ≠ Source id, empty description; changes `[source, evidence_graph]` with the graph change's `source_id` = the Source | AC-2.1, D-2, D-6 |
| T-4 | Go | same file `TestSoleGraphTx` *(new)* | Source from `sources.Create` → its Default id; unknown 16-byte id → `ErrInvalid`; 15-byte id → `ErrInvalid`; a second row `INSERT INTO evidence_graphs (id, source_id, name) VALUES (<new UUIDv7>, <source id>, 'Second')` → an error that is not `ErrInvalid` | D-9 |
| T-5 | Go | `core/database/subjects/subjects_test.go` `TestSubjects` *(new case "create joins the Source's graph")* | `subjects.Create` → `GraphID` = `defaultGraphID` of the Source; `subjects.Get` returns the same `GraphID`; the create change has `graph_id`; the `subjects.Delete` change has `graph_id` | AC-2.2, D-10 |
| T-6 | Go | `core/database/connect/connect_test.go` `TestCreateCitedBridge` *(new assertion in the happy case)* | the bridge Subject's `graph_id` = the Source's Default graph | AC-2.7 |
| T-7 | Go | `core/database/sources/sources_test.go` `TestSourceDelete` *(new case "releases the Default graph")* | Source with no Artifacts and no Subjects → `evidence_graphs` has no row for it; changes contain one `evidence_graph` delete before the `source` delete | AC-3.9, D-4 |
| T-8 | Go | `evidencegraphs_test.go` `TestLink` *(new)* | same-Source Artifact (created with `GraphIDs` empty, then its link row deleted by SQL) → one create change and the row; repeat → no changes; Artifact of another Source → `ErrInvalid`; unknown graph → `ErrInvalid`; unknown Artifact → `ErrInvalid`; empty user id → `ErrInvalid`. `LinkArtifact` with a repeated id or a 15-byte id → `ErrInvalid` | AC-3.1, AC-3.2, AC-3.3, D-11, D-12 |
| T-9 | Go | `evidencegraphs_test.go` `TestUnlink` *(new)* | Default linked to A1, A2, no Citations: `Unlink(A1)` → one delete change, row gone; then `Unlink(A2)` → `ErrLastArtifact`; `Unlink(A1)` again → no changes; with a Citation on A2 used by a Subject of the graph and A1 re-linked: `Unlink(A2)` → `ErrArtifactInUse` with no params, link kept | AC-3.4, AC-3.5, AC-3.6, AC-3.7, D-12 |
| T-10 | Go | `evidencegraphs_test.go` `TestMembership` *(new)* | `SubjectGraphTx` of a Subject → its graph; unknown Subject → `ErrInvalid`; `IsLinkedTx` true for a linked pair, false after `Unlink`; `CitationsUsingArtifactTx` → exactly the Citation with an Observation, not a second Citation on the same Artifact with no Observation | AC-3.5, D-13 |
| T-11 | Go | `core/database/artifacts/artifacts_test.go` `TestArtifacts` *(new cases)* | `GraphIDs` empty → link to Default, changes `[artifact, evidence_graph_artifact]`; `GraphIDs` = [Default] → same; another Source's graph id, an unknown id, a 15-byte id, a repeated id → `evidencegraphs.ErrInvalid` and the `artifacts` count unchanged after `writes.Run` returns | AC-2.3, AC-2.4, AC-2.5, D-11 |
| T-12 | Go | `artifacts_test.go` `TestArtifactDelete` *(new case "releases its graph links")* | linked Artifact with no Citations → deleted; no `evidence_graph_artifacts` row for it; changes `[evidence_graph_artifact delete, artifact delete]` | AC-3.8, D-16 |
| T-13 | Go | `core/database/sourcecopy/sourcecopy_test.go` `TestCopy` *(new assertions in the every-part case)* | the copy's Default graph id ≠ the original's; its links are exactly the two copied Artifacts; the original's links unchanged | AC-2.6 |
| T-14 | Go | `evidencegraphs_test.go` `TestEnsureRefs` *(new)* | Two Sources; `UPDATE evidence_graphs SET ref = NULL` on one → after `EnsureRefs` it has a valid `GPH-` ref; the other's ref unchanged; a second call changes nothing | AC-1.2, D-15 |
| T-15 | Go | `core/onboarding/ready_test.go` `TestOpenCatalogMintsGraphRefs` *(new)* | `createCatalog` in a temp dir, a user (`users.Upsert`), a seeded type (`sourcetypes.Lookup` of a `sourcevocab` key) and `sources.Create` in `writes.Run`; `UPDATE evidence_graphs SET ref = NULL`; `Close`; then `OpenCatalog` → the ref is valid | AC-1.2, D-15 |
| T-16 | Go | `core/catalogsession/session_test.go` `TestDo` *(new case "open mints missing graph refs")* | the same fixture in its own temp project dir, built with `database.Create`, `sourcevocab.Install`, a user and `sources.Create`, ref set NULL, closed; then `catalogsession.Do` on that dir → inside `fn` the ref is valid; `catalogsession.Close` of that dir at the end | AC-1.2, D-15 |
| T-17 | Go | `core/database/audit/scopes_test.go` `TestSourceScopes` *(new case "graph link and unlink")* | inside the case, add a second Artifact to A with `runArtifactCreate` (as the "artifact create and update" case does), then `evidencegraphs.Unlink` and `Link` of it on A's graph (graph id from `SoleGraphTx` inside each write) each bump A only; `newScopeSeed` is unchanged | AC-4.1, D-5 |
| T-18 | Go | `core/database/effects/registry_completeness_test.go` `TestEvidenceGraphsFlag` *(new)* | the entries with `EvidenceGraphs` are exactly `evidence_graphs` and `evidence_graph_artifacts`; a `None` entry with `EvidenceGraphs` fails `validate` | D-5 |
| T-19 | Go | `core/database/effects/evidence_graphs_test.go` *(new)* `TestEvidenceGraphsResolve` (package `effects_test`) | on a real catalog with a Source and an Artifact, inside a transaction: `Resolve` over the Source's `evidence_graph` create change and the Artifact's link create change → `EvidenceGraphs` true, the Source in `SourceScopes`, `Sources` without it from those rows; over a `subjects` create change, a `citations` create change and a `sources` title update → `EvidenceGraphs` false | AC-4.1, AC-4.2 |

AC-1.1 and AC-1.3 are covered by the existing suite (`TestPragmaHonesty`, the `deleteimpact` honesty tests, `verifySchema` on every Create/Open) and the manual check below. No migration test ([`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md) step 4).

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/audit`, `core/database/sourcecopy`, `api/ffi/handlers` tests that read the latest audit change.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each (AC-1.1 and AC-1.2's migration half are evidenced by the macOS manual check; an implementer without macOS reports them as not run and the ship-pr owner runs §10)
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain docs updated: source-layer-data-model.md, interpretation-layer-data-model.md, catalog-refs.md, data-model-source-interpretation-conclusion.md § 2, catalog-deletes.md, effects/README.md
- [ ] Skills updated: add-catalog-ref, use-catalog-session, add-seeded-vocabulary, add-catalog-model
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On the macOS dev machine (it has the app and the `sqlite3` CLI):

1. On `main`, copy a dogfood project that has Sources with Subjects, Citations and Artifacts. Open the copy and note one Source's Subject count and canvas layout.
2. Build this branch and open the same copy. The Source's canvas shows the same Subjects in the same places; the Sources list, Source page, Citation composer and Promote work as before.
3. Quit, then run `sqlite3 <copy>/provenencia.sqlite "SELECT ref, name, hex(id) = hex(source_id) FROM evidence_graphs LIMIT 3; SELECT COUNT(*) FROM subjects WHERE graph_id IS NOT source_id; SELECT (SELECT COUNT(*) FROM artifacts) - (SELECT COUNT(*) FROM evidence_graph_artifacts);"` → refs `GPH-…`, name `Default`, `1`; then `0`; then `0`.
4. Reopen, add a Source and a page scan, then Copy that Source with the scan. Both open as before. The same `sqlite3` read shows each new Source with one `Default` graph whose id is not its Source id, linked to its Artifact.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The `subjects` rebuild drops a child's FK or index | §6 recreates the indexes; `TestPragmaHonesty` checks every FK and covering index; the SQL was run against a v54 catalog with children (§4 Precedent) |
| Tests change because Source and Artifact writes return more changes | The rule at the top of §7 allows only count and list updates for exactly those changes |
| A Spike 11 PR lands a new Subject insert path before this PR merges | `InsertTx` is the only insert today; a new direct `INSERT INTO subjects` fails with `NOT NULL constraint failed: subjects.graph_id`, which is a Stop condition |
| A researcher project already has a Subject type with prefix `GPH` | Reserving only blocks new types; refs are unique per table. Dogfood projects use the seeded types only (`subjectvocab`) |

**Stop and report if:**

- `subjects.InsertTx` is no longer the only Go insert into `subjects`, or `sources.Create` / `artifacts.Create` no longer the only non-test inserts into `sources` / `artifacts`.
- Migration `000055.sql` already exists on `main`.
- `TestPragmaHonesty`, `TestResourceAndOwnedHonesty` or `TestFacetReleaseHonesty` cannot pass with the registrations in D-3 and D-4.
- An import cycle appears when `sources`, `subjects` or `artifacts` import `evidencegraphs`, or when `onboarding` calls `catalogsession.ReconcileOnOpen` (D-7 or D-15 would be wrong).
- A test fails for any reason other than the extra graph or link changes the §7 rule names.
- `writes.Run` rejects the new entity types after STEP-2 (`audit.Record` unknown entity).
- `deleteimpact.ReleaseFacets` for `KindArtifact` fails with `release_incomplete`.

## 12. Open questions

None.

## 13. PR

**Title:** S12-03 — Evidence graphs in the catalog
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-03-evidence-graphs-catalog.md`.
