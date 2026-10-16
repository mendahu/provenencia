# S10-51 — File release from the declared model

**Status:** Draft
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-51--file-release-from-the-declared-model)
**Brief:** — (no new chrome: an engine change to how deletes release Files)
**Board:** —
**Depends on:** —
**Blocks:** S10-53, S10-32
**Branch:** `s10-51-file-release`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/catalog-model.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [`docs/catalog-deletes.md`](../../../catalog-deletes.md), [`docs/source-layer-data-model.md`](../../../source-layer-data-model.md) §6–§8, [`core/database/deleteimpact/README.md`](../../../../core/database/deleteimpact/README.md), working rules "A File is released only when nothing points at it, and one purge removes its object" and "Deletes follow the evidence chain", map review [`../deployment-plan.review.md`](../deployment-plan.review.md) K-2, K-3
**Planned:** 2026-10-16 against `main` @ `98819f86`
**Review:** [`S10-51-file-release.review.md`](S10-51-file-release.review.md) — round 1 answered; round 2 pending

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

A stored File is deleted only when nothing in the catalog still points at it, and every delete removes the bytes it freed from disk. Today the release counts only Artifacts, so deleting one of two Artifacts whose Files share a deduplicated thumbnail deletes that thumbnail out from under the other. Owned releases are a hand-kept list beside the catalog model, each child table is a hand-written case, and only Artifact delete removes objects. After this PR, everything that owns a File or a value row is read from `catalogmodel`: S10-32's depiction image is two FK declarations, its region edit is one call, and every delete path already purges.

**Why now.** Slice 8 stores each likeness image as a File owned by its depiction value (S10-32). Without this PR, S10-32 would add a third hand-written release case, a second copy of the purge, signature changes on four delete functions, and a pointer the release does not count, which loses data.

**What stays.** Which deletes are allowed and what Impact reports. Thumbnails stay disposable and unaudited. Artifact delete still removes its unused File, its thumbnail, and their objects. Date and name values still release with their Observation, and the update path keeps its own date and name release. No schema change, no FFI change, no Swift change.

## 2. User stories

**US-1.** As a researcher, when I delete an Artifact, a thumbnail another Artifact shares stays, so the other Artifact keeps its picture.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Files `A` and `B` (different checksums), each with a `thumbnail` link to the same derived File `D`; Artifact 1 on `A`, Artifact 2 on `B`. Delete Artifact 1 | `A` and the `A → D` link are gone. `D`, the `B → D` link, and `D`'s object on disk stay. |
| AC-1.2 | Then delete Artifact 2 | `B`, `B → D`, and `D` are gone, and both objects are removed after commit. |
| AC-1.3 | Two Artifacts on one File `A` with thumbnail `T`; delete one | Nothing about `A` or `T` changes, and no object is reported. |
| AC-1.4 | One Artifact on File `A` with thumbnail `T`; delete it | `A`, `T`, and the link are gone; the release reports `A`'s object then `T`'s; both objects are removed after commit. A rolled-back delete removes no object. |

**US-2.** As a contributor, a table that owns a File or a value row is released from its `catalogmodel` declaration alone.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A pointer set built from a test FK slice that includes `test_values.file_id → files` | `CountFilePointers` counts rows through it. |
| AC-2.2 | An owned plan built from test FKs `test_parent.value_id → test_values` and `test_values.file_id → files` (both `BucketOwnedOutbound`), on temp tables | Releasing a `test_parent` row deletes its `test_values` row, then its File when unused, and reports the File's object. A second parent sharing that File keeps it until both go. |
| AC-2.3 | The same test plan; replace a parent's `value_id` with a new row and call `ReleaseReplaced` for the old one | The old `test_values` row and its unused File go; the new ones stay. |
| AC-2.4 | An Observation with a date value and a name value is deleted through `observations.Delete` | Both value rows are released, as today. |
| AC-2.5 | Every delete that releases owned rows or facets (`artifacts`, `observations`, `subjects`, `sources`, `citations`, `metadatafields`) | Calls `files.PurgeAfterCommit` with what it released; with nothing released, nothing is scheduled. |

## 3. Scope

| In | Covered by |
| --- | --- |
| `database.Tx` knows its catalog; `files.Object`, `files.PurgeUnreferenced`, `files.PurgeAfterCommit`; `artifacts.purgeReleasedObjects` moves there | STEP-1, AC-1.4 |
| `CountFilePointers` counts every declared FK into `files` except `file_derivatives.source_file_id` | STEP-2, AC-1.1, AC-2.1 |
| The File release removes only the released File's own derivative links, and returns what it released | STEP-3, AC-1.1, AC-1.2, AC-1.3, AC-1.4 |
| Owned release derived from `catalogmodel` owned-outbound FKs; `ownedReleases` deleted; `ReleaseReplaced` for the update path | STEP-4, AC-2.2, AC-2.3, AC-2.4 |
| Every releasing delete purges | STEP-5, AC-2.5 |
| Artifact delete end to end | STEP-6, AC-1.1, AC-1.2 |
| Docs and the add-catalog-delete skill | STEP-7 |

| Out | Home |
| --- | --- |
| The depiction value, its FKs, and its update call | S10-32 |
| Moving `core/database/observations/update.go` `releaseDateValue` / `releaseNameValue` onto `ReleaseReplaced` | Not planned; they also emit `date_value` / `name_value` audit changes the shared release does not |
| `Impact(KindFile)` inbound edges from the declared pointer set | The next PR that adds a File delete; no File delete exists today |
| Staging a client image | S10-53 |
| Auditing File or thumbnail rows on release | Not planned; releases stay unaudited as today |
| Orphan `files` rows left when an ingest is never attached | Not planned |
| A sweep of objects with no `files` row | Not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Pointer count | `core/database/deleteimpact/release.go` `CountFilePointers` | `SELECT COUNT(*) FROM artifacts WHERE file_id = ?`. No other pointer counts. |
| File release | `core/database/deleteimpact/release.go` `releaseFileIfUnused`, `releaseFileDerivatives`, `listDerivedFileIDs` | When the count is 0: list derived File ids, `DELETE FROM file_derivatives WHERE source_file_id = ? OR derived_file_id = ?`, delete the File, then recurse into each derived File. The recursion's delete of `derived_file_id = D` links removes another source's link to a shared `D`, and the Artifact-only count lets `D` be deleted while that source still uses it. |
| Snapshot | `core/database/deleteimpact/release.go` `SnapshotOwned`, `lookupParentColumn`, `splitColumn` | Loops `ownedFor(kind)` and reads a column only when `rel.Child` is `files`, `date_values`, or `name_values`. |
| Release dispatch | `core/database/deleteimpact/release.go` `releaseFromSnapshot`, `deleteValueRow` | `switch rel.Child`: `date_values`/`name_values` delete the row; `files` → `releaseFileIfUnused`; anything else does nothing. `Then` entries are never executed. |
| Object collection | `core/database/deleteimpact/release.go` `CollectFileObjects`, `collectFileObjectsFromID`, `lookupFileObject`, `FileObject` | Reads only `snap.cols["artifacts.file_id"]`, walks every derived File before the release, and de-duplicates by checksum. |
| Registration | `core/database/deleteimpact/register.go` `ownedRelease`, `ownedReleases`, `ownedFor`, `ownedColumns` | A hand-kept list of three entries mirroring the `BucketOwnedOutbound` FKs below, plus `Then` naming both `file_derivatives` columns under `artifacts.file_id`. `ownedColumns` exists only for the honesty test. |
| Derived registry precedent | `core/database/deleteimpact/register.go` `inboundEdgeList` (`sync.OnceValue`) | Production code already derives a registry once from the model. |
| Honesty | `core/database/deleteimpact/pragma_test.go` `TestResourceAndOwnedHonesty` | Its `BucketOwnedOutbound` branch requires every owned FK in `ownedColumns()`. |
| Model | `core/database/catalogmodel/model.go` `Table{Name, PK, Kind, Bucket}`, `FK{From, Column, FromCols, To, OnDelete, Bucket, Audited}`, `Tables`, `FKs` | Owned-outbound FKs: `artifacts.file_id → files`, `file_derivatives.source_file_id → files`, `file_derivatives.derived_file_id → files`, `observations.value_date_id → date_values`, `observations.value_name_id → name_values`. `date_values`, `name_values`, `file_derivatives` are `BucketOwnedOutbound` with no `PK`. `artifacts` and `observations` carry their `Kind`. |
| PK default | `core/database/effects/path.go` `pkOf` | Returns a table's `PK`, else `"id"`, so registering `PK: "id"` changes nothing there. |
| Tx | `core/database/tx.go` `Tx{*sql.Tx, after}`, `BeginWrite`, `AfterCommit` | Built only by `BeginWrite` in production; it does not hold its catalog. Two effects tests build `&database.Tx{Tx: sqlTx}` directly. |
| Released facets | `core/database/deleteimpact/facets.go` `Released{Changes}`, `merge`, `ReleaseFacets`, `eraseObservation` | `eraseObservation` snapshots and releases an Observation and keeps no objects. |
| Purge | `core/database/artifacts/artifacts.go` `Delete(tx, c, userID, id)`, `purgeReleasedObjects` | Snapshot, `CollectFileObjects`, DELETE, `ReleaseSnapshot`, then `tx.AfterCommit(func() { purgeReleasedObjects(c, objects) })`. The purge skips a checksum `files.LookupByChecksum` still finds and removes `c.Dir()/objects/…` otherwise. |
| Releasing deletes | `core/database/observations/write.go` `Delete`; `core/database/subjects/subjects.go` `Delete`; `core/database/sources/sources.go` `Delete`; `core/database/citations/citations.go` `Delete`; `core/database/metadatafields/metadatafields.go` `Delete` | All take `(tx *database.Tx, userID, id)` and call `ReleaseFacets`; `observations.Delete` also calls `SnapshotOwned` / `ReleaseSnapshot`. None purges. |
| Update path | `core/database/observations/update.go` `releaseDateValue`, `releaseNameValue` | Count references and delete the old value row, with an audit change. Out of this PR. |
| Tests | `core/database/deleteimpact/release_test.go` `TestCountFilePointersIfUnused` (calls `releaseFileIfUnused` directly near line 84), `TestCollectFileObjects`, `TestSnapshotOwnedRequiresParent` (calls `ReleaseOwned` near line 215), helpers `testCatalogInternal`, `insertSourceType`, `insertSource`, `insertFile`, `insertArtifact`, `newFileID`; `core/database/deleteimpact/impact_test.go` `TestSnapshotOwnedThenRelease` (calls `ReleaseSnapshot` near lines 629 and 654); `core/database/observations/observations_test.go` `TestObservations` row "delete drops owned date and name values"; `core/database/artifacts/artifacts_test.go` `TestArtifactDelete`, `TestArtifactDeletePurgesDisk` | — |

**Precedent.** `30d4b18b` ("Unlink objects/ when the last Artifact pointer drops a file") added the snapshot-collect-purge sequence. Keep its order (know the objects before the rows go, unlink only after commit, skip a checksum still in `files`); move it, do not change it. `inboundEdgeList` is the shape for a registry derived once from `catalogmodel`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `database.Tx` gains an unexported `catalog *Catalog`, set by `BeginWrite`, and `func (t *Tx) Catalog() *Catalog` *(new)*. `files.Object{ChecksumSHA256, MediaType string}` *(new)* replaces `deleteimpact.FileObject`. `files.PurgeUnreferenced(c *database.Catalog, objs []Object)` *(new)* is today's `purgeReleasedObjects` body, moved, with a doc line: "Object bytes are removed here and by `ingest.File`'s rollback of a failed import; nowhere else (S10-53's `StagedImage.Discard` calls this function)." `files.PurgeAfterCommit(tx *database.Tx, objs []Object)` *(new)*: nothing when `len(objs) == 0`; otherwise `tx.AfterCommit(func() { PurgeUnreferenced(tx.Catalog(), objs) })`. A non-empty list on a `Tx` with no catalog panics with "files: purge needs the write's catalog". `PurgeUnreferenced` does nothing for an object whose checksum appears twice after the first removal. | Any delete can purge without a new parameter, so S10-32 changes no delete signature. `files` already owns the object path and checksum lookup. Duplicate objects in the list are harmless. | Passing `*database.Catalog` into five more `Delete` signatures — wider change, and every FFI handler. A purge in `deleteimpact` — it takes `*sql.Tx` and has no directory. |
| D-2 | `CountFilePointers(tx, fileID)` sums `SELECT COUNT(*) FROM <From> WHERE <Column> = ?` over `filePointers()`: every `catalogmodel.FKs` entry with `To == "files"` except `From == "file_derivatives" && Column == "source_file_id"`. `filePointers` is `sync.OnceValue` over `filePointersFrom(catalogmodel.FKs)`; an internal `countFilePointers(tx, fks []catalogmodel.FK, fileID)` takes the slice so tests pass their own. | A missed pointer loses data. A File's own thumbnail link (`source_file_id`) belongs to that File, not to something pointing at it. Derived once, like `inboundEdgeList`, with a seam for tests instead of mutating the global. | A hand-kept list — the bug this PR fixes. Mutating `catalogmodel.FKs` in tests — process-wide state beside `TestPragmaHonesty`. |
| D-3 | The File release (`releaseFile(tx, plan, fileID) ([]files.Object, error)`), when the count is 0: list the File's own derived ids (`source_file_id = ?`); report the File's object; `DELETE FROM file_derivatives WHERE source_file_id = ?`; delete the File; then release each derived File the same way and append what they report. The released File's object comes before its derived Files' objects. It never deletes links by `derived_file_id`. | A File any `derived_file_id` link points at has a count above 0, so it is never released while that link exists. Deleting only its own links keeps another source's link to a shared derived File. | Keeping `OR derived_file_id = ?` — it removes the other source's link (AC-1.1). |
| D-4 | The owned plan is derived from `catalogmodel`: an *owned edge* is an FK with `Bucket == BucketOwnedOutbound`. For a parent table, its edges are the owned FKs whose `From` is that table. An edge whose `To` is `files` releases through D-3. Any other `To` is a *row child*: its own edges are the owned FKs whose `From` is the child table, read from the child row before it is deleted, and released after it. `file_derivatives` is never a `To`, so its two FKs are reached only through D-3. `ownedPlan()` is `sync.OnceValue` over `ownedPlanFrom(catalogmodel.FKs, catalogmodel.Tables)`; internal functions take the plan so tests pass their own. `ownedRelease`, `ownedReleases`, `ownedFor`, and `ownedColumns` are deleted. | `catalogmodel` already declares every owned edge; the second list only restated it. S10-32 becomes two FK lines. | Keeping `ownedReleases` and walking its `Then` — a registry whose `Mode` is derivable and whose `file_derivatives` entries the walk ignores (review M-3, M-6). |
| D-5 | A row child is deleted with `DELETE FROM <table> WHERE <PK> = ?` using its `catalogmodel.Tables` `PK`. `date_values` and `name_values` gain `PK: "id"` in `catalogmodel`. A row child table with no `PK` is a plan error caught by T-7. A snapshot column that is NULL, a child row that is already gone, and a child column that is NULL are each skipped. | Keys come from the declared model, as `tableByKind` does. Today's behavior already skips empty columns. | Hard-coding `id` — a second place keys are spelled. |
| D-6 | Public API: `SnapshotOwned(tx, kind, id) (OwnedSnapshot, error)` (kind → its table through `tableByKind`, as today), `ReleaseSnapshot(tx, snap) ([]files.Object, error)`, `ReleaseOwned(tx, kind, id) ([]files.Object, error)`, and `ReleaseReplaced(tx *sql.Tx, column string, oldID []byte) ([]files.Object, error)` *(new)*. `ReleaseReplaced` finds the owned edge named `column` (`"table.column"`), and releases `oldID` as that edge's child (row child with its own edges, or a File through D-3). Call it after the parent row already holds the new value; an empty `oldID` is a no-op; an unknown column is `ErrInvalid`. | The update path (S10-32's region edit replaces `depiction_values.image_file_id`) needs the same release as delete; the private File release is otherwise unreachable from `observations` (review M-1). | A depiction-specific release in S10-32 — a third copy. |
| D-7 | `Released` gains `Objects []files.Object`, and `merge` appends it. `eraseObservation` puts its `ReleaseSnapshot` objects on its `Released`. `CollectFileObjects`, `collectFileObjectsFromID`, `lookupFileObject` (if unused after D-3), and `FileObject` are deleted. | One walk reports what it released; a pre-walk can report Files that end up kept (AC-1.3). | Keeping `CollectFileObjects` — a second walk that can disagree. |
| D-8 | Every delete that calls `ReleaseFacets` or `ReleaseSnapshot` calls `files.PurgeAfterCommit(tx, objs)` with everything it released: `artifacts.Delete`, `observations.Delete`, `subjects.Delete`, `sources.Delete`, `citations.Delete`, `metadatafields.Delete`. `artifacts.Delete` keeps its `c` parameter (no signature change in this PR). `artifacts.purgeReleasedObjects` is deleted. | When S10-32 makes an Observation own a File, every path that can delete it already purges. Today the lists are empty except for Artifacts, so behavior does not change. | Wiring only Artifacts now (the first draft's D-8) — S10-32 would have to find every path. |
| D-9 | Pinned by T-7: every owned edge's `From` is a table with a `Kind`, a row child reached by an owned edge, or `file_derivatives`; every row child has a `PK`; no owned edge cycles. `TestResourceAndOwnedHonesty` drops its owned branch (the plan is the model). | The derived walk trusts these; a wrong declaration should fail CI, not delete data. | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `Tx.Catalog()` *(new)* | method | `Tx.AfterCommit` | Same struct; `BeginWrite` already has the catalog. |
| `files.Object`, `files.PurgeUnreferenced`, `files.PurgeAfterCommit` in `core/database/files/purge.go` *(new)* | type / func | `deleteimpact.FileObject`, `artifacts.purgeReleasedObjects` | Moved; the old copies are deleted. |
| `filePointers`, `countFilePointers` *(new)* | func | `inboundEdgeList` | Same derive-once pattern. |
| `ownedPlan`, `ownedPlanFrom` *(new)* | func | `ownedReleases` | Replaces the hand-kept list with the model. |
| `ReleaseReplaced` *(new)* | func | `releaseFromSnapshot` | The same per-edge release, exposed for updates. |
| `Released.Objects` *(new)* | field | `Released.Changes` | Same carrier, merged the same way. |
| `TestOwnedPlanShapes` *(new)* | test | `TestResourceAndOwnedHonesty` | Pins D-9. |

**Builds on.** `catalogmodel.FKs` and `Tables` (as `inboundEdgeList` and `tableByKind` do), the 30d4b18b purge order, `Tx.AfterCommit`.
**Deletes.** `ownedRelease`, `ownedReleases`, `ownedFor`, `ownedColumns`, the owned branch of `TestResourceAndOwnedHonesty`, `deleteimpact.FileObject`, `CollectFileObjects`, `collectFileObjectsFromID`, `TestCollectFileObjects`, the `switch` in `SnapshotOwned` and `releaseFromSnapshot`, `deleteValueRow`, the `OR derived_file_id = ?` delete, `artifacts.purgeReleasedObjects`.
**Next adds.** S10-32 declares `observations.value_depiction_id → depiction_values` and `depiction_values.image_file_id → files` as `BucketOwnedOutbound` with `depiction_values` `PK: "id"`, and calls `ReleaseReplaced(tx, "depiction_values.image_file_id", oldFileID)` plus `files.PurgeAfterCommit` on a region edit. No other `deleteimpact` or delete-path code.

### Schema

None — no catalog change. `catalogmodel` gains `PK: "id"` on `date_values` and `name_values` (model only).

### Go core

| Package | Change |
| --- | --- |
| `core/database` | `Tx.catalog`, `Tx.Catalog()`; `BeginWrite` sets it (D-1). |
| `core/database/files` | `purge.go` *(new)* per D-1. |
| `core/database/catalogmodel` | `PK: "id"` on `date_values`, `name_values` (D-5). |
| `core/database/deleteimpact` | D-2 to D-7 and D-9 in `release.go`, `register.go`, `facets.go`, `pragma_test.go`. |
| `core/database/artifacts`, `observations`, `subjects`, `sources`, `citations`, `metadatafields` | `files.PurgeAfterCommit` per D-8. |

```go
// core/database/files/purge.go (sketch)
type Object struct{ ChecksumSHA256, MediaType string }
func PurgeUnreferenced(c *database.Catalog, objs []Object)
func PurgeAfterCommit(tx *database.Tx, objs []Object)

// core/database/deleteimpact/release.go (sketch)
func CountFilePointers(tx *sql.Tx, fileID []byte) (int, error)
func SnapshotOwned(tx *sql.Tx, kind catalogmodel.Kind, id []byte) (OwnedSnapshot, error)
func ReleaseSnapshot(tx *sql.Tx, snap OwnedSnapshot) ([]files.Object, error)
func ReleaseOwned(tx *sql.Tx, kind catalogmodel.Kind, id []byte) ([]files.Object, error)
func ReleaseReplaced(tx *sql.Tx, column string, oldID []byte) ([]files.Object, error)
```

Errors: unchanged, plus `ErrInvalid` from `ReleaseReplaced` on an unknown column.

### FFI

None — no RPC changes.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — The write knows its catalog, and one purge lives in `files`**
Files: `core/database/tx.go`, `core/database/files/purge.go` *(new)*, `core/database/files/files_test.go`, `core/database/artifacts/artifacts.go`.
Do: D-1. In `artifacts.Delete`, convert the collected `deleteimpact.FileObject`s to `files.Object` for this step only and call `files.PurgeAfterCommit(tx, objs)`; delete `purgeReleasedObjects`. Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/ ./core/database/files/... ./core/database/artifacts/...` passes, including `TestArtifactDeletePurgesDisk`.

**STEP-2 — Count every declared File pointer**
Files: `core/database/deleteimpact/release.go`, `core/database/deleteimpact/release_test.go`.
Do: D-2. Add T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/deleteimpact/...` passes, including `TestCountFilePointersIfUnused`.

**STEP-3 — Release only a File's own links, and report what was released**
Files: `core/database/deleteimpact/release.go`, `core/database/deleteimpact/facets.go`, `core/database/deleteimpact/release_test.go`, `core/database/deleteimpact/impact_test.go`, `core/database/artifacts/artifacts.go`, `core/database/observations/write.go`.
Do: D-3 and D-7. Replace `FileObject` with `files.Object` everywhere. Update every existing caller of `ReleaseSnapshot`, `ReleaseOwned`, and `releaseFileIfUnused` to take the new return (`TestCountFilePointersIfUnused`, `TestSnapshotOwnedRequiresParent`, `TestSnapshotOwnedThenRelease`, `eraseObservation`, `observations.Delete`). `artifacts.Delete` now takes the objects from `ReleaseSnapshot` and drops the STEP-1 conversion. Delete `CollectFileObjects` and `TestCollectFileObjects`. Add T-3 and T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**STEP-4 — Owned release from the model, and `ReleaseReplaced`**
Files: `core/database/catalogmodel/model.go`, `core/database/deleteimpact/release.go`, `core/database/deleteimpact/register.go`, `core/database/deleteimpact/release_test.go`, `core/database/deleteimpact/pragma_test.go`.
Do: D-4, D-5, D-6, D-9. Add T-5 and T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes, including `catalogmodel.TestPragmaHonesty`, `TestResourceAndOwnedHonesty`, and `TestObservations` (T-6).

**STEP-5 — Every releasing delete purges**
Files: `core/database/observations/write.go`, `core/database/subjects/subjects.go`, `core/database/sources/sources.go`, `core/database/citations/citations.go`, `core/database/metadatafields/metadatafields.go`, `core/database/observations/observations_test.go`.
Do: D-8. Add T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**STEP-6 — Artifact delete end to end**
Files: `core/database/artifacts/artifacts_test.go`.
Do: Add T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/artifacts/...` passes.

**STEP-7 — Docs**
Files: `core/database/deleteimpact/README.md`, `.agents/skills/add-catalog-delete/SKILL.md`, `.agents/skills/add-catalog-model/SKILL.md`, `docs/catalog-deletes.md`, `docs/source-layer-data-model.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do:
- `core/database/deleteimpact/README.md` and `docs/catalog-deletes.md`: owned releases come from `catalogmodel` owned-outbound FKs (row children by their `PK`, Files through the File release); a File is released when no declared FK into `files` points at it (its own thumbnail links do not count); `ReleaseSnapshot` and `ReleaseReplaced` return released objects; every releasing delete calls `files.PurgeAfterCommit`.
- `add-catalog-delete` skill, "New table": a table that owns a File or a value row declares its FKs as `BucketOwnedOutbound` (and a `PK` on a row child) in `catalogmodel`, adds no `deleteimpact` code, calls `files.PurgeAfterCommit` with what its delete releases, and calls `ReleaseReplaced` when an update swaps an owned child. Remove every mention of `ownedReleases`.
- `add-catalog-model` skill: an owned-outbound row table registers `PK`.
- `source-layer-data-model.md` §8: releasing a File removes only its own derivative links; a derived File is deleted when nothing points at it.
- `deployment-plan.md`: in the working rule that begins "A File is released only when nothing points at it", replace "walks `ownedReleases` `Then` generically" with "derives owned releases from `catalogmodel`", and "one `catalogmodel` FK and one `ownedReleases` entry" with "two `catalogmodel` FK declarations"; the S10-51 **Plan:** line's "— Draft" becomes "— Landed #N", and " — done, #N" is appended to its SLICE 8 line (as `S10-50  … — done, #413`). `completed.md`: a `## S10-51 — File release from the declared model` section above the current first section, in the shape of the `S10-50` section.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-51-file-release.md` passes; `grep -rn "ownedReleases" core docs .agents` returns nothing outside this plan and the review logs.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

Test fixtures for T-2 and T-5 use temp tables created in the test catalog (`CREATE TABLE test_parent (id BLOB PRIMARY KEY, value_id BLOB)`, `CREATE TABLE test_values (id BLOB PRIMARY KEY, file_id BLOB)`) and an explicit FK slice; they never modify `catalogmodel.FKs`, `catalogmodel.Tables`, or a real table. Test FK literals: `{From: "test_parent", Column: "value_id", To: "test_values", OnDelete: "NO ACTION", Bucket: catalogmodel.BucketOwnedOutbound}` and `{From: "test_values", Column: "file_id", To: "files", OnDelete: "NO ACTION", Bucket: catalogmodel.BucketOwnedOutbound}`, with table `{Name: "test_values", PK: "id", Bucket: catalogmodel.BucketOwnedOutbound}`.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/files/files_test.go` `TestPurgeUnreferenced` *(new)*, table-driven | write object bytes with `os.WriteFile` at `StorageRelPath`: no `files` row → removed; a row with that checksum → kept; empty list → no error; the same object twice → removed once, no error. Then `writes.Run` whose function calls `PurgeAfterCommit(tx, objs)` and returns an error → the object still exists | AC-1.4, D-1 |
| T-2 | Go | `core/database/deleteimpact/release_test.go` `TestCountFilePointersIfUnused` (extend) | a File that is `derived_file_id` of one link → count 1; the same File as `source_file_id` only → 0; `countFilePointers` with the test FK slice and a `test_values` row naming the File → 1 | AC-2.1, D-2 |
| T-3 | Go | `release_test.go` `TestReleaseKeepsSharedDerived` *(new)* | Files `A`, `B`, derived `D` with `A → D` and `B → D` thumbnail links, an Artifact on `A`: `ReleaseOwned(KindArtifact)` → `A` and `A → D` gone, `D` and `B → D` stay; returned objects are `[A]` | AC-1.1, D-3, D-7 |
| T-4 | Go | `release_test.go` `TestReleaseReportsReleasedObjects` *(new)*, table-driven | lone Artifact on `A` with thumbnail `T` → `[A, T]` in that order; two Artifacts on `A`, release one → `[]`; fileless Artifact → `[]` | AC-1.3, AC-1.4, D-3, D-7 |
| T-5 | Go | `release_test.go` `TestOwnedPlanRelease` *(new)*, table-driven over a plan from `ownedPlanFrom(testFKs, testTables)` | release a `test_parent` row → its `test_values` row and File gone, object returned; two parents sharing one File through their own rows → File kept after the first, released after the second; `ReleaseReplaced`-equivalent internal call after pointing a parent at a new `test_values` row → old row and File gone, new ones stay; NULL `value_id` → no-op | AC-2.2, AC-2.3, D-4, D-5, D-6 |
| T-6 | Go | `core/database/observations/observations_test.go` `TestObservations` row "delete drops owned date and name values" | existing, unedited, passes | AC-2.4, D-4 |
| T-7 | Go | `core/database/deleteimpact/pragma_test.go` `TestOwnedPlanShapes` *(new)* | the production `ownedPlan()` satisfies D-9; a test plan with a row child and no `PK` fails the check | D-5, D-9 |
| T-8 | Go | `core/database/observations/observations_test.go` row "delete schedules no purge with nothing released" *(new)* | `observations.Delete` of a text Observation through `writes.Run` → commits; no object removed; a sentinel object written under `objects/` stays | AC-2.5, D-8 |
| T-9 | Go | `core/database/artifacts/artifacts_test.go` `TestArtifactDeleteKeepsSharedThumbnail` *(new)* | two Artifacts on Files `A`, `B` whose `thumbnail` links share derived `D` (inserted with `filederivatives.Insert`); `artifacts.Delete` the first through `writes.Run` → `D`, `B → D`, and `D`'s object stay; delete the second → all gone and both objects removed | AC-1.1, AC-1.2, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestArtifactDeletePurgesDisk`, `TestSnapshotOwnedThenRelease`, `catalogmodel.TestPragmaHonesty` (after D-5), the Subject delete tests that reach `eraseObservation`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); off macOS it is reported as failing on xcode-test only, with every other check passing
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none in this PR)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI in this PR)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and working-rule wording; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: catalog-deletes.md, source-layer-data-model.md §8, deleteimpact README, add-catalog-delete and add-catalog-model skills
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project. Add two Sources, each with an Artifact made from the same JPEG saved twice with one byte changed at the very end of the second copy (so the Files differ but their thumbnails are the same pixels). Both Artifact cards show a thumbnail.
2. Delete the first Artifact. Reopen the second Source: its Artifact still shows its thumbnail.
3. Delete the second Artifact. Its thumbnail is gone, and the project's objects folder holds no file for either scan.

An implementer without a Mac reports this check as not run; T-9 covers the same behavior.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The derived plan reaches an FK the old list skipped | Today the owned FKs and the old list are the same five (`TestResourceAndOwnedHonesty`); T-7 pins the shapes. |
| `PK: "id"` on `date_values` and `name_values` changes another reader | `effects.pkOf` already defaults to `"id"`; `tableByKind` skips tables with no `Kind`. |
| S10-44 (planned) also edits `core/database/deleteimpact/facets.go` | It changes `releaseSubjectClaims`; this PR changes `Released` and `eraseObservation`. Rebase on whichever lands first. |

**Stop and report if:**

- A `files` row is deleted anywhere other than the File release (`grep -rn "DELETE FROM files" core`).
- `catalogmodel.FKs` has a `BucketOwnedOutbound` FK other than the five in §4 at the base you start from.
- `catalogmodel.TestPragmaHonesty` refuses `PK` on a table with no `Kind`.
- `TestArtifactDeletePurgesDisk` needs its expectations changed to pass.
- Any `database.Tx` built outside `BeginWrite` reaches `files.PurgeAfterCommit` with a non-empty list.

## 12. Open questions

None.

## 13. PR

**Title:** S10-51 — File release from the declared model
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-51-file-release.md`.
