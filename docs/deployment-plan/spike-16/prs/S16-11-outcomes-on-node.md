# S16-11 — Outcome records live on the canonical node

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 2b — Entity records
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine refactor; every page draws the same values, records and pictures)
**Board:** —
**Depends on:** S14-42 (Spike 14, landed [#748](https://github.com/mendahu/provenencia/pull/748)); S16-10 (landed [#789](https://github.com/mendahu/provenencia/pull/789), which this plan's `docs/catalog-caches.md` row and `BenchmarkCacheSizes` build on)
**Blocks:** S16-30 (composes a provenance record from the node's outcome row); S15-28 (`FieldRecordCounts` from the node's outcome rows)
**Branch:** `s16-11-outcomes-on-node`
**Version bump:** none — no change a researcher can see, no wire and no catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-depiction-value`](../../../../.agents/skills/add-depiction-value/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) (for the handler test's `catalogsession.Do`)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [design § Entity records](../cache-architecture.md#entity-records); [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §2.7 and §6 (the gallery); [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §7 (depictions) and §10; [`docs/catalog-caches.md`](../../../catalog-caches.md); working rules: "(Batch 2) The point of this batch is simpler code", "(Batch 2) Cache entities and results per key; compose answers", "(Batch 2) Every cache names the signal that updates it, and updates exactly", "(Batch 2) Apply states its fan-out", "(Batch 2) A write's one-shot graph fills only what it reads", "(Batch 2) Orders are defined and results do not move", "(Batch 2) Memory is budgeted", "(Batch 2) Measure growth"; goal bars 11 and 15
**Planned:** 2026-11-11 against `main` @ `cba6e26`
**Review:** — (not reviewed yet)

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

A Conclusion page's "why" (each record's outcome), its gallery, the lists' lead pictures and the portrait choice all read the auto-reconciler's outcome rows. Today three SQL reads fetch them again on every page open and every summary batch, beside a node that the graph cache already keeps exact. After this PR the canonical node holds its outcome rows and its chosen likeness, and every reader composes from it: one gallery rule over the node, no reconciler or gallery SQL on a warm page open. What remains of the detail read's SQL is the record shells (Citation, Artifact, Source words) keyed by Observation, which S16-30 composes from cached records.

**Why now.** Slice 2b's first PR. S16-30 and S15-28 compose provenance and record counts from the node's outcome rows; the map ([design § Entity records](../cache-architecture.md#entity-records)) puts outcomes on the node first.

**What stays.** Every value, outcome, picture, lead and portrait the app shows; the wire (`GetConclusionDetail`, `SetPreferredLikeness`); the record shells' SQL (S16-30); `PersonLifeOwners` (S16-19); the catalog schema.

## 2. User stories

**US-1.** As the engine, when a canonical node loads or a commit reloads it, the node holds every outcome row of the handle and its chosen likeness, so readers compose from memory and Verify checks both.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A person handle with a `sex_at_birth` record (`male`) is loaded | `Node.Outcomes[sex_at_birth id]` is one row: that Observation id, `Reason` `kept`, `Value` 0 (value rank 1), no `DeniedBy`, zero `Vote` |
| AC-1.2 | A second `sex_at_birth` record (`female`) is filed on the handle and the commit names it | The reloaded node's `sex_at_birth` rows are both Observations in ascending id order, each with the reason and value index the stored rows hold |
| AC-1.3 | The handle's portrait is set to Observation B, then cleared | `Node.PreferredLikeness` is B's id after the first commit and nil after the second |
| AC-1.4 | `Verify` runs after `auto_reconciler_outcomes.reason` of a loaded node's row is changed by hand | `Verify` returns an error naming the node's outcomes |
| AC-1.5 | `Verify` runs after `canonical_entities.preferred_likeness_observation_id` of a loaded node is changed by hand | `Verify` returns an error |

**US-2.** As a researcher, when I open a Person, Event or Place page, choose a portrait, or read a list, I see the same values, records, gallery, lead pictures and portrait as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Every existing case of `TestForEntity*`, `TestFieldsOf`, `TestDetail*`, `TestLinkRecordReadsKeepQueriesConstant`, the `PersonDerivedRows` tests and the handler tests `TestGetConclusionDetail*` | Each passes with its assertions unchanged (the two test edits D-12 names change setup only) |
| AC-2.2 | Every gallery case of today's `TestForEntities` and `TestForEntitiesPreferred` (earliest first, redraw keeps its place, negatives, provisional, weak, place and event pictures, pointer leads, fall-backs, delete clears the pointer) read through `gallery.Of` | Same pictures, order, `Preferred` and `Leads` as today |
| AC-2.3 | `SetPreferredLikeness` with another Person's likeness, a negative of the same picture, an unknown Observation, an unknown entity, or a nil user | `canonicalentities.ErrInvalid`, the pointer unchanged (today's `TestSetPreferredLikeness` "refused") |
| AC-2.4 | A Person with two likenesses is opened once through `GetConclusionDetail`, then opened again | The second open reads neither `auto_reconciler_outcomes` nor `depiction_values` (SQLite authorizer) and issues no graph query |

**US-3.** As a contributor, I can see what outcomes on the node cost at the batch's Scale.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `BenchmarkCacheSizes` at 4,000 and 16,000 persons, before and after this PR, on the same machine | The ship log records `nodes-B` and `nodes-ms` both times; `nodes-B` grows by under 6 MB at 16,000 persons |
| AC-3.2 | `TestLargeSourceBudgets` before and after | Every budget passes; the ship log records "warm at open" and the four city-write figures both times |

## 3. Scope

| In | Covered by |
| --- | --- |
| `Node` holds its outcome rows (`Outcomes`) and chosen likeness (`PreferredLikeness`), loaded in one more batched query per load chunk and reloaded by `Set.Handles` | STEP-2, AC-1.1 – AC-1.3 |
| Verify checks both | STEP-2, AC-1.4, AC-1.5 |
| One gallery rule over the node: `gallery.Of`, `gallery.PropertyIDs`; `gallery.ForEntities` and its SQL go; a tile's image is the value's image (user, 2026-11-11) | STEP-3, AC-2.2 |
| The detail read's fields and outcomes from the node: `sqlPropertyIDs` and the reconciler half of `sqlOutcomes` go; the record shells stay one SQL read keyed by Observation | STEP-4, AC-2.1 |
| The summaries' lead picture from the node (`leadPictures`) | STEP-5, AC-2.1 |
| `SetPreferredLikeness` validates and reads the previous pointer through the node on `graphcache.Over(tx)`; `sqlPreferred` goes | STEP-6, AC-2.3 |
| A warm open reads no reconciler or gallery SQL, held by a test | STEP-7, AC-2.4 |
| The large fixture carries outcome rows; nodes' memory and time at 4,000 and 16,000 measured before and after | STEP-1, STEP-8, AC-3.1, AC-3.2 |
| Docs: the gallery image rule, §2.7, §10, the cache table, the skill, the package comment, the map and design | STEP-9 |

| Out | Home |
| --- | --- |
| The record shells (Citation, Artifact, Source words), `toponymsOf`, the detail's date and name lookups | S16-30 |
| Source records, the records tier | S16-12, S16-13 |
| `PersonLifeOwners` and `PersonDerivedRows`' second pass | S16-19, S16-30 |
| `FieldRecordCounts` and the records read by (owner, Property) | S15-28 |
| Any wire or Swift change | not planned |
| A memory budget gate on nodes, or shrinking today's 58 MB of nodes toward the 40 MB budget | not planned in this PR: reported to the user (S16-10 ship log "Go caches ≈118 MB"); the map's rule "Memory is budgeted" decides where it goes |
| Subject id on the node's outcome row | S16-30 adds it if its composition needs it (D-2) |

## 4. Current state

Verified at `cba6e26`.

| Area | Where | Today |
| --- | --- | --- |
| Node | `core/database/graphcache/graph.go` `Node`, `Value`, `Member` | `Node` has `Values` (property id → ranks), `Links`, `Members`; no outcomes, no likeness pointer |
| Node load | `core/database/graphcache/load.go` `loadChunk`, `entityRows`, `StoredValues`, `loadMembers` | Per chunk of ≤ `sqlbatch.Size` (500) ids: entity rows, values, members, links. `loadMembers` is the shape to copy: one query, `IN (placeholders)`, map by entity id |
| Commit reload | `core/database/graphcache/commit.go` `OnCommit` | `Set.Handles` ids are reloaded through `loadIDs(…, true)` → `loadChunk`, so anything `loadChunk` loads is exact after a commit that names the handle |
| Verify | `core/database/graphcache/verify.go` `verifyIDs`, `sameMembers` | Identity read back from `entityRows`; values against the registered truth recompute; members read back with `loadMembers` and compared by `sameMembers` |
| Outcome table | `core/database/migrations/000040.sql`, `000041.sql` | `auto_reconciler_outcomes(entity_id, property_id, observation_id, reason, value_rank, denied_by, vote_support, vote_total)`, PRIMARY KEY `(entity_id, property_id, observation_id)`; `observation_id` has no foreign key |
| Outcome writer | `core/database/autoreconciler/autoreconciler.go` `recomputeBatch` | Writes one row per `autoreconcile.Outcome`: `value_rank = o.Value + 1` or NULL when `o.Value < 0`; `denied_by` NULL when empty; votes NULL when the vote is zero. Values and outcomes of a handle are rewritten together in the same transaction |
| Outcome type | `core/autoreconcile/autoreconcile.go` `Outcome`, `Vote`, `Reason.InGallery`, `Reason.IntoValue` | `Outcome{ObservationID, Reason, Value (index, -1 none), DeniedBy, Vote}`: exactly the stored row's fields |
| Gallery | `core/database/gallery/gallery.go` `ForEntities`, `sqlGallery`, `galleryPrefixArgs`, `ByObservation`, `Lead`, `Picture` | `ForEntities` joins values, entities, properties, outcomes, observations, `depiction_values`, files; keeps `InGallery` outcomes; image = the Observation's own `dv.image_file_id` through `depictionvalue.ReadableImage`; groups by (entity, rank); sorts each picture's Observations ascending and pictures by first Observation then rank; marks `Preferred` and `Leads` |
| Gallery callers | `core/database/conclusiondetails/conclusiondetails.go:192` `ForEntity`; `core/database/conclusionsummaries/compose.go:644` `leadPictures`; `core/database/canonicalentities/canonicalentities.go:275` `SetPreferredLikeness` | The three production callers; nothing else calls `ForEntities` |
| Value image | `core/database/autoreconciler/autoreconciler.go` (candidate load joins `depiction_values` and `files`, `depictionvalues.FromJoin` → `depictionvalue.Readable`); `core/valuecompare/depiction.go` | `value_image_file_id` is already limited to readable images: it is written from a candidate read through `Readable`. Depictions are the same value only when their image File ids are equal, so every record counted into a value rank has that rank's image. No production code updates `files.media_type` (`rg "UPDATE files" core` finds only `original_filename` and tests) |
| Detail read | `core/database/conclusiondetails/conclusiondetails.go` `ForEntity`, `fieldsOf`, `loadProperties`, `attachValues`, `loadOutcomes`, `sqlPropertyIDs`, `sqlOutcomes`, `toponymsOf` | `loadProperties` lists a handle's fields from `SELECT DISTINCT entity_id, property_id FROM auto_reconciler_outcomes`; `loadOutcomes` runs `sqlOutcomes` (outcomes joined to observations, subjects, citations, artifacts, sources, terms and credibility) by entity id, then date, name and toponym lookups; a gallery lead's outcome is kept apart for the tiles |
| Derived rows | `conclusiondetails.go` `PersonDerivedRows` | Loads owner nodes, then `fieldsOf(…, subjectValued=true)`: it gains the node path with no change of its own |
| Summaries | `core/database/conclusionsummaries/compose.go` `leadPictures`, the person path (`need [][]byte`, `byID`) near line 611 and the event path (`need []*graphcache.Node`) near line 998 | `leadPictures(g, ids)` runs `gallery.ForEntities(g, ids)` and keeps `gallery.Lead`'s image |
| Likeness write | `core/database/canonicalentities/canonicalentities.go` `SetPreferredLikeness`, `sqlPreferred`, `sqlSetPreferred` | Reads the previous pointer with `sqlPreferred` (no row → `ErrInvalid`), validates the Observation with `gallery.ForEntities(tx.Tx, …)` + `gallery.ByObservation`, writes with `sqlSetPreferred` |
| Pointer effects | `core/database/effects/registry.go` `"canonical_entities"` | `Handles: self()`: a pointer write reloads the node. An Observation delete clears the pointer by `ON DELETE SET NULL` in the same write that recomputes the handle |
| Handler | `api/ffi/handlers/conclusion_detail.go` `GetConclusionDetail`, `attachSummary`, `attachPictures` | `ForEntity`, then the summary (`PersonsByIDs`, `PersonDerivedRows`), then `files.RelPaths` |
| SELECT counter | `core/database/databasetest/selects.go` `Selects`, `CountSelects` | A SQLite authorizer counts `SQLITE_SELECT`; it does not record which tables a statement reads |
| Large fixture | `core/database/promotealign/bench_test.go` `bulkCanon`, `BenchmarkCacheSizes`; `perf_test.go` `newLargeFixture`, `TestLargeSourceBudgets` | `bulkCanon` writes entities and `auto_reconciler_values` only: no `auto_reconciler_outcomes` row, so the bulk nodes hold no outcomes. `BenchmarkCacheSizes` reports `nodes-B` for loading every person, event and place |
| Cache table | `docs/catalog-caches.md` "Typed `Graph` fields" | The `nodes` row: 45.3 / 58.2 MB at 4,000 / 16,000 persons |
| Tests | `core/database/gallery/gallery_test.go` `TestForEntities`, `TestForEntitiesPreferred`, `TestSetPreferredLikeness`, `fixture.gallery`; `core/database/autoreconciler/autoreconciler_test.go` `galleryOf` (line 686); `core/database/conclusiondetails/conclusiondetails_test.go` (`TestForEntityLikenessNeedsAPNG` line 376, `TestForEntityQueryCountIsConstant` line 497); `placerecords_test.go` `TestLinkRecordReadsKeepQueriesConstant`; `core/database/graphcache/graph_test.go`; `api/ffi/handlers/conclusion_detail_test.go` | Hold the gallery, detail, likeness and node behavior this PR keeps |

**Precedent.** `Member` and `loadMembers` in `core/database/graphcache/load.go`, with `sameMembers` in `verify.go`: a per-node list loaded in one batched query beside `StoredValues`, kept exact by the `Set.Handles` reload and read back by Verify. Copy that shape for outcomes. Do not copy the truth recompute that values use (D-4).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `Node` gains `Outcomes map[string][]autoreconcile.Outcome` (property id → that Property's rows in ascending Observation id) and `PreferredLikeness []byte` (nil for none). An outcome row reuses `autoreconcile.Outcome`: `Value` is `value_rank - 1`, or -1 when `value_rank` is NULL; `DeniedBy` nil when NULL; `Vote` zero when the vote columns are NULL. Every id is cloned (`cloneID`) | `autoreconcile.Outcome` is exactly what `recomputeBatch` stored; reusing it adds no type and gives readers `Reason.InGallery` and `IntoValue` directly. `Values` is keyed the same way | A new `graphcache.Outcome` with a 1-based `ValueRank` — a second type for the same row; `conclusiondetails.Outcome` — it carries record words that are S16-30's |
| D-2 | The outcome rows load from `auto_reconciler_outcomes` alone: `SELECT entity_id, property_id, observation_id, reason, value_rank, denied_by, vote_support, vote_total FROM auto_reconciler_outcomes WHERE entity_id IN (placeholders) ORDER BY entity_id, property_id, observation_id`, one query per load chunk, in a new `(g *Graph) loadOutcomes(ids)` beside `loadMembers`, called from `loadChunk` after `loadMembers`. The table name comes from `catalogmodel` like `valueTable` (`g.outcomeTable = mustTableName("auto_reconciler_outcomes")`). No Subject id and no join to `observations`. This narrows the map's "(Observation, Subject, reason, value rank, denied by, vote)": the plan's commit updates the map's scope entry and the design | Nothing in this PR reads an outcome's Subject; the detail's `SubjectID` still comes from the record shell (D-6). One table read in primary-key order keeps every node load, at open and on a write's one-shot graph, as light as it can be, and lets the large fixture carry outcome rows (D-11). S16-30 maps an Observation to its Subject through the records tier, or adds the column when it plans | Joining `observations` for `subject_id` — a join on every node load for a field no reader uses |
| D-3 | `entityRows` selects `e.preferred_likeness_observation_id` too; `entityRow` gains `likeness []byte`; `loadChunk` copies it to `Node.PreferredLikeness` (cloned, nil when NULL) | The pointer lives on the entity row the load already reads; `Handles: self()` reloads it | A separate query — the row is already read |
| D-4 | Verify reads both back: `verifyIDs` compares `PreferredLikeness` with the entity row inside the identity check, and compares `Outcomes` with a fresh `g.loadOutcomes(ids)` through `sameOutcomes` *(new)*: same Property set, and per Property the same rows in order, each equal in Observation id, Reason, Value, DeniedBy and Vote. The error reads `graphcache: %x outcomes: …` | Values and outcomes of a handle are rewritten by the same `recomputeBatch`, so the values' truth recompute already catches a handle whose recompute was skipped; what Verify must add is that the node equals the committed rows, as `sameMembers` does | Recomputing outcomes in `truthRows` — changes `TruthFunc`'s signature and the autoreconciler to check what the values' truth check already covers |
| D-5 | `gallery.ForEntities`, `sqlGallery` and `galleryPrefixArgs` are deleted. `gallery.PropertyIDs(v graphcache.Vocabulary) [][]byte` *(new)* returns the ids of `vocabkey.GalleryKeys()` at origin `provenencia` that exist in `v`, in `GalleryKeys` order. `gallery.Of(n *graphcache.Node, props [][]byte) []Picture` *(new)*: nil for a nil node. For each gallery Property id, each of the node's values of that Property whose `ImageFileID` is non-empty is a picture when at least one outcome row of that Property has `Value + 1 == value.Rank` and `Reason.InGallery()`; its `ObservationIDs` are those rows' Observation ids, ascending; its `ImageFileID` is the value's (cloned); its `Rank` the value's. Pictures sort as today (first Observation id, then rank, stable); `Preferred` and the lead come from `n.PreferredLikeness` through `ByObservation` as today. `Picture`, `ByObservation` and `Lead` are unchanged | One gallery rule over the node, so the detail page, the summaries and the likeness check agree by construction. The value's image equals every counted record's image (§4 "Value image"); the user chose the value's image on 2026-11-11 | Keeping each record's own `depiction_values.image_file_id` — needs a per-row image on the node or a SQL read, so the gallery SQL would stay (user declined, 2026-11-11) |
| D-6 | `conclusiondetails`: `loadProperties` takes each node's field Properties from `n.Outcomes`' keys instead of `sqlPropertyIDs` (deleted), with the same filters (depiction never; subject only when `subjectValued`). Before its type-order sort it orders a node's Properties by id bytes (`bytes.Compare`), then `sort.SliceStable` by type order as today, so ties keep a fixed order (the primary-key order SQL gave). `loadOutcomes` is replaced by `loadRecords` *(new)*: it collects the Observation ids of every listed field's outcome rows plus the gallery leads, reads the record shells with `sqlRecords` *(new)* through `sqlbatch.Query` keyed by Observation id, and builds each `Outcome` from the node row (`ObservationID`, `Reason`, `ValueRank = Value + 1` or 0, `DeniedBy`, `Vote`) and the record (every other field, the date, name and toponym lookups unchanged). An outcome row whose record is missing is dropped, as today's inner join drops it. `sqlRecords` is `sqlOutcomes` without `auto_reconciler_outcomes`: `FROM observations o JOIN subjects s … JOIN sources src LEFT JOIN property_terms t … ProvenanceJoins WHERE o.id IN (`, selecting `o.id` first in place of the outcome columns. The querier stays `c.DB()` | The node holds the reconciler half; the record half is S16-30's | Moving the record shells to the graph querier — no reader needs it; S16-30 deletes them |
| D-7 | `ForEntity` reads the gallery as `gallery.Of(n, gallery.PropertyIDs(v))` with `v` from `c.Graph().Vocabulary()`. Leads and tiles are built as today from those pictures | One rule | — |
| D-8 | `leadPictures(g *graphcache.Graph, nodes []*graphcache.Node) (map[string][]byte, error)`: one `g.Vocabulary()` and one `gallery.PropertyIDs` per call, then `gallery.Lead(gallery.Of(n, props))` per node. The event path passes `need`; the person path passes `byID[string(id)]` for each id of `need`, in `need` order | The lead now reads the node, which is already in each row's inputs, so the input index needs nothing new (rule "A cache a composer reads hands its inputs to the row") | Calling `PropertyIDs` per node — a vocabulary scan per row in `RowsForSearch` |
| D-9 | `SetPreferredLikeness`: after `RequireUserID`, `n, err := graphcache.Over(tx.Tx).Node(entityID)`; a nil node is `ErrInvalid`; `prev := n.PreferredLikeness`; the Observation is valid when `gallery.ByObservation(gallery.Of(n, gallery.PropertyIDs(v)), observationID)` finds it, `v` from the same `Over` graph's `Vocabulary()`. `sqlPreferred` is deleted. A merged handle behaves as today (it has a row, so it is not refused for being merged) | The map: the write's check composes through the node on `graphcache.Over(tx)`, so the one gallery rule covers it. The write's one-shot graph loads only this node and the vocabulary (rule "A write's one-shot graph fills only what it reads") | Keeping `sqlPreferred` beside the node — two reads of one row |
| D-10 | `databasetest.Selects` gains `Read(table string) bool` *(new)*: whether any statement since `CountSelects` or the last `Reset` read the table. The authorizer also records `arg1` of each `sqlite3.SQLITE_READ` in a set guarded by a `sync.Mutex`; `Reset` clears it. AC-2.4's test (T-9) uses it | The one helper that already watches every statement on the catalog's one connection; the table set is the narrowest proof that the reconciler and gallery reads are gone | A test-only export in `conclusiondetails` — misses the summary and the likeness reads; the graph's counter — `ForEntity`'s SQL does not go through the graph |
| D-11 | `bulkCanon` ends, before its commit, with one statement giving every bulk value its record's outcome: `INSERT INTO auto_reconciler_outcomes (entity_id, property_id, observation_id, reason, value_rank) SELECT v.entity_id, v.property_id, randomblob(16), v.reason, v.rank FROM auto_reconciler_values v WHERE NOT EXISTS (SELECT 1 FROM auto_reconciler_outcomes o WHERE o.entity_id = v.entity_id)` (handles a real reconcile wrote keep theirs). `BenchmarkCacheSizes` also reports the nodes fill's wall time as `nodes-ms`. No new budget: this PR claims no speed (rule "Measure growth" binds a speed claim); the existing budgets in `TestLargeSourceBudgets` gate "warm at open" and the city writes | Without outcome rows the large fixture's nodes cannot show what outcomes cost, and "warm at open" would measure nothing new. One outcome per value is the floor of a real catalog (every value has a record) | Writing real Observations in `bulkCanon` — subjects, citations and artifacts for 37,000 handles, a fixture rewrite for a measurement |
| D-12 | Test edits, each in setup only: (a) `TestForEntityLikenessNeedsAPNG` runs `autoreconciler.Rebuild(db)` and `f.C.Graph().Drop()` after its `UPDATE files`, since the value, not the file row, is now the reading and no production write changes a media type; (b) the gallery subtest "tile image is the depiction row, not the reconciler cache" is deleted: `autoreconciler_test.go` "Verify sees a stale image" already holds that a hand-edited `value_image_file_id` is caught; (c) "a stored pointer at a negative falls back to the earliest" drops the graph after its raw `UPDATE`; (d) `fixture.gallery` and `galleryOf` read `gallery.Of(node, gallery.PropertyIDs(vocab))` through `f.c.Graph()`; `TestForEntities` and `TestForEntitiesPreferred` are renamed `TestOf` and `TestOfPreferred`, and "no ids" becomes "nil node" (`gallery.Of(nil, props)` is nil) | Each edit follows a Decision; no assertion changes | Keeping a SQL fallback so the tests pass untouched — a second gallery path |
| D-13 | No registry entry. Outcomes and the likeness pointer are node fields, in the `nodes` row of `docs/catalog-caches.md` ("Typed `Graph` fields") | They are part of the node, which `Set.Handles` reloads; an entry would be a second home for node data | A keyed `outcomes` entry — a second mechanism for what the node reload already does exactly |

**Rules this plan bends.** The gallery image rule: [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §6 ("Each tile's image is the Observation's own `depiction_values.image_file_id`, not the reconciler cache's `value_image_file_id`"), [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §7 ("The gallery shows each record's own `depiction_values.image_file_id`, not the reconciler cache copy") and [`add-depiction-value`](../../../../.agents/skills/add-depiction-value/SKILL.md) ("Reconcile, gallery and portrait"). D-5 reads the value's image instead → in a consistent catalog the two are the same File (§4 "Value image"), and the node path removes the gallery SQL → the user chose the value's image (2026-11-11) → STEP-9 rewrites the three lines.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `Node.Outcomes`, `Node.PreferredLikeness` *(new)* | fields | `Node.Members`, `Node.Values` | Extends `Node`, the registry's typed graph; reuses `autoreconcile.Outcome` (D-1) |
| `(g *Graph) loadOutcomes` *(new)*, `outcomeTable` *(new)* | func, field | `loadMembers`, `valueTable` | Same shape; a new table read |
| `sameOutcomes` *(new)* | func | `sameMembers` | Same shape for another field |
| `gallery.Of`, `gallery.PropertyIDs` *(new)* | funcs | `gallery.ForEntities` | Replace it (deleted) |
| `conclusiondetails.loadRecords`, `sqlRecords` *(new)* | func, const | `loadOutcomes`, `sqlOutcomes` | Replace them (deleted); the reconciler half moves to the node |
| `databasetest.Selects.Read` *(new)* | test helper | `Selects.N` | Extends the one statement watcher |
| `nodes-ms` metric *(new)* | benchmark metric | `nodes-B` | Same fill, its time |

Searched: `rg "auto_reconciler_outcomes" core api` (readers: `gallery`, `conclusiondetails`, `autoreconciler`), `rg "ForEntities" core api`, `rg "preferred_likeness" core`, `rg "CountSelects|SetQueryCounterForTest" core api`.

**Builds on.** The node and its `Set.Handles` reload (`core/database/graphcache/load.go`, `core/database/graphcache/commit.go`); Verify's read-back (`sameMembers`); `autoreconcile.Outcome`; `databasetest.CountSelects`; `BenchmarkCacheSizes` and `docs/catalog-caches.md` (S16-10).
**Deletes.** `gallery.ForEntities`, `sqlGallery`, `galleryPrefixArgs` (and `gallery`'s imports of `database`, `sqlbatch`, `depictionvalue` if unused after); `conclusiondetails.sqlPropertyIDs`, `sqlOutcomes`, `loadOutcomes` (replaced by `sqlRecords`, `loadRecords`); `canonicalentities.sqlPreferred`; the gallery subtest "tile image is the depiction row, not the reconciler cache"; the three doc lines on the record's own image; the package comment's "The Why stays a SQL read each time the page opens" and "(gallery.ForEntities)".
**Next adds.** S16-30 composes a provenance record from `Node.Outcomes` plus the records tier and Source records, deleting `sqlRecords`; S15-28 counts records per field from `Node.Outcomes` with no query.

### Reach

None — no new path into existing code, no new state reachable. The node may now carry an outcome row whose Observation no longer exists only if `auto_reconciler_outcomes` is damaged (no foreign key); the detail drops it (no record, D-6) and the gallery lists its id, which `ForEntity` drops (no record) and `SetPreferredLikeness`' write would refuse by the pointer's foreign key. The reconciler rewrites a handle's outcomes with every Observation delete that recomputes it, so a real catalog never holds one.

### Scale

At the map's Scale (large fixture at 4,000 and 16,000 persons; 18,000 and 37,000 handles; a person in 80 Evidence graphs; a Source with 10,000 Observations):

- **Node load.** One more query per chunk of 500 ids, on the primary key `(entity_id, …)`, returning the chunk's outcome rows. Open loads every person, event and place (`WarmOnOpen`); the fixture's one outcome per value (D-11) makes that 37,000 handles' rows at 16,000. Est. 120 B per row; measured as `nodes-B` (AC-3.1).
- **Apply's work per commit.** No new Apply: a commit that names handles reloads those nodes as today, now with their outcome rows. Per reloaded handle the work is its record count: a person in 80 graphs with two records per graph and ten Properties is about 1,600 rows; a handle with 10,000 records (bar 2's crafted Source) is 10,000 rows, read in primary-key order, the same rows the detail read fetched on each open before.
- **Write's one-shot graph.** `SetPreferredLikeness` loads one node and the vocabulary on `graphcache.Over(tx)`; summary composition on a write loads only the nodes it reads, now with outcomes. The city writes' budgets hold it (AC-3.2).
- **Detail read.** The record shells read by Observation id in chunks of 500: one query for a person with up to 500 records, four at 2,000. Today's read was one query by entity id; S16-30 removes it.
- **Memory.** Bounded by the catalog's outcome rows of loaded nodes. The map's 40 MB budget is already exceeded by today's caches (S16-10 ship log); this PR measures and reports, it does not gate (Out).

### Edges

- **Assumptions.** Readers rely on (1) each Property's outcome rows in ascending Observation id — the load's `ORDER BY` holds it, and `gallery.Of` takes a picture's Observation ids in that order (T-4 "earliest first, and a redraw keeps its place"); (2) every record of a value rank shows that rank's image — depictions are one value only on equal File ids (`core/valuecompare/depiction.go`), held by T-4's existing cases; (3) the field list's order — fixed by the id pre-sort (D-6), held by `TestFieldsOf` and `TestForEntity`.
- **Neighbors.** A `weak` record of a picture is in the gallery (InGallery) and a `denied`, `provisional`, `against` or `no_evidence` record is not — T-4's negative, provisional and weak cases; a depiction value whose image the loader refused (empty `ImageFileID`) is no picture — `TestForEntityLikenessNeedsAPNG` (D-12a) and the autoreconciler's "an unreadable image is no evidence".
- **Removed inputs.** An Observation deleted: the write recomputes the handle and the pointer is set NULL, `Handles` reloads the node — T-4 "deleting the Observation clears the pointer". A record unpromoted: the node reloads without it, the pointer row stays and the gallery falls back — "unpromoted likeness falls back and keeps the pointer row". A merged handle: `ForEntity` refuses it as today (`Handle`).
- **Real data.** None — no table, seed or configuration drives a rule here; `GalleryKeys` is unchanged.
- **Untrusted size.** Outcome rows are written by the app's reconciler, one per member Observation of a handle; their count is the handle's record count, bounded as the detail read was before (Scale). No new list from data the app did not write.

### Schema

None — no catalog change. `auto_reconciler_outcomes` is read with its existing primary key.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` | `graph.go`: `Node.Outcomes map[string][]autoreconcile.Outcome`, `Node.PreferredLikeness []byte` *(new)*, documented; `Graph.outcomeTable` *(new)* set in `New`. `load.go`: `entityRow.likeness`, `entityRows` selects the pointer; `loadOutcomes(ids [][]byte) (map[string]map[string][]autoreconcile.Outcome, error)` *(new)*; `loadChunk` fills both fields (an empty `Outcomes` is a non-nil empty map, as `Values`). `verify.go`: pointer in the identity check, `sameOutcomes` *(new)* (D-4) |
| `core/database/gallery` | `func PropertyIDs(v graphcache.Vocabulary) [][]byte` *(new)*; `func Of(n *graphcache.Node, props [][]byte) []Picture` *(new)*; `ForEntities` and its SQL deleted; package comment says the gallery is composed from the node |
| `core/database/conclusiondetails` | D-6, D-7; package comment: outcomes come from the node; the record shells are one SQL read keyed by Observation until S16-30; pictures from `gallery.Of` |
| `core/database/conclusionsummaries` | `leadPictures` per D-8 and its two callers |
| `core/database/canonicalentities` | `SetPreferredLikeness` per D-9 |
| `core/database/databasetest` | `Selects.Read` per D-10 |

Errors: none new. `SetPreferredLikeness` keeps `ErrInvalid` for an unknown entity and an Observation outside the gallery.

```go
// Outcomes are the auto-reconciler's outcome rows by property id, each
// Property's rows in ascending Observation id: what it did with each record
// of an accepted or provisional member. Value is the value rank minus one.
Outcomes map[string][]autoreconcile.Outcome
// PreferredLikeness is canonical_entities.preferred_likeness_observation_id;
// nil for none.
PreferredLikeness []byte
```

### FFI

None — no wire change; `GetConclusionDetail` and `SetPreferredLikeness` keep their requests and responses.

### macOS

None — no Swift change.

### Strings

None — no copy changes; no action changes what it covers.

## 7. Steps

**STEP-1 — Outcome rows in the large fixture, and today's figures**
Files: `core/database/promotealign/bench_test.go`. Skill: —.
Do: D-11 — the `INSERT … SELECT` at the end of `bulkCanon`'s transaction, and the `nodes-ms` metric (time the same fill `measure("nodes", …)` runs; report milliseconds with `b.ReportMetric(ms, "nodes-ms")`). Then, before any other change, record today's figures in the ship log: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run '^$' -bench BenchmarkCacheSizes -benchtime 1x` (`nodes-B`, `nodes-ms` at both sizes) and `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -v` ("warm at open", "locate an event at the city", "unpromote an event at the city", "delete a location at the city", "second part_of from the city").
Check: both commands pass; `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/` passes.

**STEP-2 — The node holds its outcomes and chosen likeness**
Files: `core/database/graphcache/graph.go`, `load.go`, `verify.go`, `graph_test.go`. Skill: `add-catalog-query`.
Do: D-1 – D-4. Write T-1 – T-3 first.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/...` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/conclusiondetails/ ./core/database/gallery/` passes.

**STEP-3 — One gallery rule over the node**
Files: `core/database/gallery/gallery.go`, `gallery_test.go`, `core/database/autoreconciler/autoreconciler_test.go` (`galleryOf` only). Skill: `add-depiction-value`.
Do: D-5; the test edits D-12 (b), (c), (d). Callers still compile against `ForEntities` until STEP-4 – STEP-6: keep `ForEntities` in this step and delete it in STEP-6, so each step builds.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/gallery/ ./core/database/autoreconciler/` passes.

**STEP-4 — The detail read takes fields, outcomes and pictures from the node**
Files: `core/database/conclusiondetails/conclusiondetails.go`, `export_test.go` (only if a helper names a deleted symbol), `conclusiondetails_test.go` (D-12a only).
Do: D-6, D-7; package comment.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusiondetails/ ./api/ffi/handlers/` passes; `rg -n "sqlPropertyIDs|sqlOutcomes|auto_reconciler_outcomes" core/database/conclusiondetails` finds nothing.

**STEP-5 — The summaries' lead picture from the node**
Files: `core/database/conclusionsummaries/compose.go`.
Do: D-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/... ./core/database/searchindex/...` passes.

**STEP-6 — The likeness write checks through the node; `ForEntities` goes**
Files: `core/database/canonicalentities/canonicalentities.go`, `core/database/gallery/gallery.go`.
Do: D-9; delete `ForEntities`, `sqlGallery`, `galleryPrefixArgs` and imports left unused.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonicalentities/... ./core/database/gallery/ ./api/ffi/handlers/` passes; `rg -n "ForEntities|sqlPreferred" core api` finds nothing.

**STEP-7 — A warm open reads no reconciler or gallery SQL**
Files: `core/database/databasetest/selects.go`, `selects_test.go`, `api/ffi/handlers/conclusion_detail_test.go`. Skill: `use-catalog-session`.
Do: D-10; T-8, T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/databasetest/ ./api/ffi/handlers/ -run 'TestCountSelects|TestGetConclusionDetail'` passes.

**STEP-8 — Measure after**
Files: the ship log.
Do: rerun STEP-1's two commands; record `nodes-B`, `nodes-ms` and the six figures after, beside STEP-1's.
Check: `TestLargeSourceBudgets` passes; `nodes-B` at 16,000 persons is under 6 MB above STEP-1's (else Stop condition).

**STEP-9 — Docs**
Files: `docs/conclusion-layer-data-model.md` (§2.7: a canonical node in the graph cache carries the handle's outcome rows and chosen likeness, reloaded with it; §6: the tile image line → "Each tile's image is the value's image File (`value_image_file_id`), which every record counted into that value shows"), `docs/conclusion-reconciliation.md` (§7 depiction paragraph: same change; §10: the graph cache's node holds the outcomes and reloads with the handle), `.agents/skills/add-depiction-value/SKILL.md` ("The gallery is `gallery.Of` over the node"; the image line as §6), `docs/catalog-caches.md` (the `nodes` row: holds outcome rows and the chosen likeness; STEP-8's sizes), `docs/deployment-plan/spike-16/cache-architecture.md` and `deployment-plan.md` S16-11 scope entry (the outcome row without Subject, D-2), `completed.md`, the map row, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-11-outcomes-on-node.md` and `scripts/check-guidance-names.py` pass; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphcache/graph_test.go` `TestNodeOutcomes` *(new)* | `graphcachetest.Canon`: `Entity("person")`, `Term(p, "sex_at_birth", "male")`, `Commit(p)` → one row, `kept`, `Value` 0, nil `DeniedBy`, zero `Vote`; a second `Term(p, "sex_at_birth", "female")` and `Commit(p)` → two rows in ascending Observation id, each equal to the stored row read with SQL; `Verify` passes | AC-1.1, AC-1.2, D-1, D-2 |
| T-2 | Go | `graph_test.go` `TestNodePreferredLikeness` *(new)* | Raw `UPDATE canonical_entities SET preferred_likeness_observation_id = X`, `OnCommit` with `Handles: {id}` → `PreferredLikeness` is X; set NULL and commit → nil | AC-1.3, D-3 |
| T-3 | Go | `graph_test.go` `TestVerifyChecksOutcomesAndLikeness` *(new)*, table-driven | Loaded node, then by raw SQL with no commit: (a) `UPDATE auto_reconciler_outcomes SET reason = 'outvoted'` → `Verify` errors with "outcomes"; (b) set the pointer → `Verify` errors | AC-1.4, AC-1.5, D-4 |
| T-4 | Go | `core/database/gallery/gallery_test.go` `TestOf`, `TestOfPreferred` (renamed), `TestSetPreferredLikeness` | Every existing case through `gallery.Of` with today's assertions; "nil node" → nil; the subtest deleted by D-12b gone | AC-2.2, AC-2.3, D-5, D-9, D-12 |
| T-5 | Go | `core/database/autoreconciler/autoreconciler_test.go` (the likeness subtests using `galleryOf`) | Unchanged assertions through `gallery.Of` | AC-2.2 |
| T-6 | Go | `core/database/conclusiondetails/*_test.go` | Every existing test; `TestForEntityLikenessNeedsAPNG` with D-12a's setup | AC-2.1, D-6, D-7 |
| T-7 | Go | `core/database/conclusionsummaries/...`, `api/ffi/handlers/conclusion_detail_test.go` | Existing summary and handler tests (`TestGetConclusionDetailListsPictures`, `…EventPictureLead`, `TestSetPreferredLikeness`) unchanged | AC-2.1, D-8 |
| T-8 | Go | `core/database/databasetest/selects_test.go` `TestCountSelects` (extended) | A `SELECT` on `users` → `Read("users")` true, `Read("files")` false; `Reset` → `Read("users")` false | D-10 |
| T-9 | Go | `api/ffi/handlers/conclusion_detail_test.go` `TestGetConclusionDetailWarmReadsNoReconcilerOrGallerySQL` *(new)* | `likenessFixture`, two `saveLikeness`, `promoteInDir`; one `GetConclusionDetail`; then in `catalogsession.Do` `sel := databasetest.CountSelects(t, c)` and the graph's `SetQueryCounterForTest`; a second `GetConclusionDetail` → `sel.Read("auto_reconciler_outcomes")` and `sel.Read("depiction_values")` false, graph queries 0, two pictures returned | AC-2.4 |
| T-10 | Go | `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes`; `perf_test.go` `TestLargeSourceBudgets` | `nodes-B`, `nodes-ms` reported; budgets pass | AC-3.1, AC-3.2, D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/gallery`, `core/database/conclusiondetails`, `core/database/conclusionsummaries` (lead pictures and the rebuild-equals-upkeep tests), `core/database/searchindex` (summary rows and their inputs), `core/database/promotealign` (budgets), `api/ffi/handlers`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/... passes
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`gallery.ForEntities`, `sqlPropertyIDs`, `sqlOutcomes`, `sqlPreferred`, "not the reconciler cache"); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (STEP-9)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump: none
- [ ] Ship log: STEP-1 and STEP-8 figures side by side (nodes-B, nodes-ms at 4,000 and 16,000; warm at open; four city writes)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated: none)
- [ ] Domain docs updated: conclusion-layer-data-model.md §2.7 and §6, conclusion-reconciliation.md §7 and §10, catalog-caches.md, add-depiction-value
- [ ] PR description: what shipped, AC evidence, deviations, the measurements
```

## 10. Manual check

1. Open a project with a Person who has two likenesses from two Sources and a name from each. Open the Person on Vitals: the same values and provenance sheet as before this PR (Sources, Citation labels, pages, denied and outvoted rows).
2. Open the Gallery: the same pictures in the same order, the lead first; choose the second picture as portrait: the header and the People list show it; clear it: the earliest leads again.
3. Delete the Observation of the chosen portrait from its Source: the Person's portrait falls back to the other picture.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A reader relied on the gallery SQL's own image read | §4 "Value image" and D-5; T-4 and T-5 hold every gallery case |
| Field order moves where two Properties tie | D-6's id pre-sort; `TestFieldsOf` |
| Open or a city write slows as every node load carries outcomes | STEP-1 / STEP-8 figures; the existing budgets gate |
| Memory over the estimate | AC-3.1's 6 MB line; Stop condition |

**Stop and report if:**

- `nodes-B` at 16,000 persons grows by 6 MB or more between STEP-1 and STEP-8: report both figures and the outcome row count (`SELECT COUNT(*) FROM auto_reconciler_outcomes` on the fixture).
- A `TestLargeSourceBudgets` budget that passes at STEP-1 fails at STEP-8.
- An existing gallery, detail, summary or likeness assertion fails and only a change to the assertion (not D-12's setup edits) would make it pass.
- T-9 finds `auto_reconciler_outcomes` or `depiction_values` read on the warm open by code this plan does not name.
- A production writer of `files.media_type` exists at the base (`rg "UPDATE files" core` beyond `original_filename`).
- `auto_reconciler_outcomes` has gained a column, a foreign key on `observation_id`, or a second writer besides `recomputeBatch` and `Rebuild`.

## 12. Open questions

None.

## 13. PR

**Title:** S16-11 — Outcome records live on the canonical node
**Body:** objective in one sentence; what shipped against each US; AC evidence; the STEP-1 / STEP-8 figures; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-11-outcomes-on-node.md`.
