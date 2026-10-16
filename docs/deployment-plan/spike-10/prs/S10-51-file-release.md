# S10-51 — File release from the declared model

**Status:** Draft
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-51--file-release-from-the-declared-model)
**Brief:** — (no new chrome: an engine change to how deletes release Files)
**Board:** —
**Depends on:** —
**Blocks:** S10-32
**Branch:** `s10-51-file-release`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/catalog-model.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [`docs/catalog-deletes.md`](../../../catalog-deletes.md), [`docs/source-layer-data-model.md`](../../../source-layer-data-model.md) §6–§8, [`core/database/deleteimpact/README.md`](../../../../core/database/deleteimpact/README.md), working rules "A File is released only when nothing points at it, and one purge removes its object" and "Deletes follow the evidence chain", map review [`../deployment-plan.review.md`](../deployment-plan.review.md) K-2, K-3
**Planned:** 2026-10-16 against `main` @ `98819f86`
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

A stored File is deleted only when nothing in the catalog still points at it, and one shared step removes its bytes from disk. Today the release counts only Artifacts, so deleting one of two Artifacts whose Files share a deduplicated thumbnail deletes that thumbnail out from under the other. Owned-row releases are hand-coded per child table, and only Artifact delete removes objects. After this PR, a new owned File pointer (S10-32's depiction image) is one `catalogmodel` FK and one `ownedReleases` entry, with no `deleteimpact` code.

**Why now.** Slice 8 stores each likeness image as a File owned by its depiction value (S10-32). Without this PR, S10-32 would add a third hand-written release case, a second copy of the object purge, and a pointer the release does not count, which loses data.

**What stays.** Which deletes are allowed and what Impact reports. Thumbnails stay disposable and unaudited. Artifact delete still removes its unused File, its thumbnail, and their objects. Date and name values still release with their Observation. No schema change, no FFI change, no Swift change.

## 2. User stories

**US-1.** As a researcher, when I delete an Artifact, a thumbnail another Artifact shares stays, so the other Artifact keeps its picture.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Files `A` and `B` (different checksums), each with a `thumbnail` link to the same derived File `D`; Artifact 1 on `A`, Artifact 2 on `B`. Delete Artifact 1 | `A` and the `A → D` link are gone. `D`, the `B → D` link, and `D`'s object on disk stay. |
| AC-1.2 | Then delete Artifact 2 | `B`, `B → D`, and `D` are gone, and both objects are removed after commit. |
| AC-1.3 | Two Artifacts on one File `A` with thumbnail `T`; delete one | Nothing about `A` or `T` changes. |
| AC-1.4 | One Artifact on File `A` with thumbnail `T`; delete it | `A`, `T`, and the link are gone; both objects are removed after commit. A rolled-back delete removes no object. |

**US-2.** As a contributor, a new table that owns a File is released by registration alone.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A test registers an FK into `files` in `catalogmodel.FKs` only | `CountFilePointers` counts rows through it, with no change to `deleteimpact`. |
| AC-2.2 | A test registers an owned row child with a nested owned File (`Parent` row → child row → `files`) in `ownedReleases` | Deleting the parent releases the child row and then its File when unused, and reports the File's object. |
| AC-2.3 | An Observation with a date value and a name value is deleted | Both value rows are released, as today. |

## 3. Scope

| In | Covered by |
| --- | --- |
| `files.Object`, `files.PurgeUnreferenced`, `files.PurgeAfterCommit`; `artifacts.purgeReleasedObjects` moves there | STEP-1, AC-1.4 |
| `CountFilePointers` counts every `catalogmodel` FK into `files` except `file_derivatives.source_file_id` | STEP-2, AC-1.1, AC-2.1 |
| The File release removes only the released File's own derivative links; a shared derived File stays | STEP-3, AC-1.1, AC-1.2, AC-1.3 |
| `ReleaseSnapshot` returns the objects it released; `Released` carries them; `CollectFileObjects` is deleted | STEP-3, AC-1.4 |
| Owned release is generic: snapshot every registered column, release row children and their `Then` entries, release `files` children through the File release | STEP-4, AC-2.2, AC-2.3 |
| Artifact delete uses the returned objects and the shared purge | STEP-5, AC-1.4 |
| Docs and the add-catalog-delete skill | STEP-6 |

| Out | Home |
| --- | --- |
| The depiction value, its FKs and registration | S10-32 |
| Purging objects from Observation, Subject, and Source deletes (none releases a File today) | S10-32, when depictions make them release Files |
| Staging a client image | S10-53 |
| Auditing File or thumbnail rows on release | Not planned; releases stay unaudited as today |
| Orphan `files` rows left when an ingest is never attached | Not planned |
| A sweep of objects with no `files` row | Not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Pointer count | `core/database/deleteimpact/release.go` `CountFilePointers` | `SELECT COUNT(*) FROM artifacts WHERE file_id = ?`. No other pointer counts. |
| File release | `core/database/deleteimpact/release.go` `releaseFileIfUnused`, `releaseFileDerivatives`, `listDerivedFileIDs` | When the count is 0: list derived File ids, `DELETE FROM file_derivatives WHERE source_file_id = ? OR derived_file_id = ?`, delete the File, then recurse into each derived File. The recursion's delete of `derived_file_id = D` links removes another source's link to a shared `D`, and the Artifact-only count lets `D` be deleted while that source still uses it. |
| Snapshot | `core/database/deleteimpact/release.go` `SnapshotOwned` | Loops `ownedFor(kind)` and reads a column only when `rel.Child` is `files`, `date_values`, or `name_values`. |
| Release dispatch | `core/database/deleteimpact/release.go` `releaseFromSnapshot`, `deleteValueRow` | `switch rel.Child`: `date_values`/`name_values` delete the row; `files` → `releaseFileIfUnused`; `file_derivatives` and anything else do nothing. `Then` entries are never executed. |
| Object collection | `core/database/deleteimpact/release.go` `CollectFileObjects`, `collectFileObjectsFromID`, `lookupFileObject`, `FileObject` | Reads only `snap.cols["artifacts.file_id"]`, then walks every derived File, before the release. |
| Registration | `core/database/deleteimpact/register.go` `ownedRelease` (`Parent`, `Column`, `Child`, `Mode`, `Then`), `ownedReleases`, `ownedFor`, `ownedColumns` | Three entries: `observations.value_date_id` → `date_values` (always), `observations.value_name_id` → `name_values` (always), `artifacts.file_id` → `files` (ifUnused) with `Then` naming `file_derivatives.source_file_id` and `file_derivatives.derived_file_id` (always). `ownedColumns` walks `Then` for honesty only. |
| Honesty | `core/database/deleteimpact/pragma_test.go` `TestResourceAndOwnedHonesty` | Every `BucketOwnedOutbound` FK must appear in `ownedColumns()`. |
| Model | `core/database/catalogmodel/model.go` `FK` (`From`, `Column`, `To`, …), `FKs` | FKs into `files`: `artifacts.file_id`, `file_derivatives.source_file_id`, `file_derivatives.derived_file_id`. `files` is `{PK: "id", Kind: KindFile, Bucket: BucketPool}`; `date_values`, `name_values`, `file_derivatives` have no PK registered. |
| Released facets | `core/database/deleteimpact/facets.go` `Released{Changes}`, `merge`, `eraseObservation` | `eraseObservation` calls `SnapshotOwned` / `ReleaseSnapshot` for an Observation and keeps no objects. |
| Purge | `core/database/artifacts/artifacts.go` `Delete`, `purgeReleasedObjects` | `Delete` snapshots, calls `CollectFileObjects`, deletes, `ReleaseSnapshot`, then `tx.AfterCommit(func() { purgeReleasedObjects(c, objects) })`. The purge skips a checksum `files.LookupByChecksum` still finds and removes `c.Dir()/objects/…` otherwise. |
| Observation delete | `core/database/observations/write.go` `Delete` (around line 100) | `SnapshotOwned` → DELETE → `ReleaseSnapshot`. |
| Files package | `core/database/files/files.go` `StorageRelPath`, `LookupByChecksum` | Imports `database`, `sqlbatch`, `objectpath`; nothing that imports `deleteimpact`. |
| Tests | `core/database/deleteimpact/release_test.go` `TestCountFilePointersIfUnused`, `TestCollectFileObjects`, `TestSnapshotOwnedRequiresParent`, helpers `testCatalogInternal`, `insertFile`, `insertArtifact`, `newFileID`; `core/database/artifacts/artifacts_test.go` `TestArtifactDelete`, `TestArtifactDeletePurgesDisk` | Cover the Artifact-only count, object collection, and the disk purge. |

**Precedent.** `30d4b18b` ("Unlink objects/ when the last Artifact pointer drops a file") added the snapshot-collect-purge sequence. Keep its order (collect before the rows go, unlink only after commit, skip a checksum still in `files`); move it, do not change it. `e80e08e6` moved the table model into `catalogmodel`; this PR reads that model the same way `inboundEdges` does.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `files.Object{ChecksumSHA256, MediaType string}` *(new)* replaces `deleteimpact.FileObject`. `files.PurgeUnreferenced(c *database.Catalog, objs []Object)` *(new)* is today's `purgeReleasedObjects` body, moved. `files.PurgeAfterCommit(tx *database.Tx, c *database.Catalog, objs []Object)` *(new)* registers `PurgeUnreferenced` with `tx.AfterCommit` when `len(objs) > 0`. `deleteimpact` imports `files` for the type. | One purge for every owner of a File. `files` already owns the object path and checksum lookup. | Keeping the purge in `artifacts` — S10-32 would copy it. A purge inside `deleteimpact` — it takes `*sql.Tx` and has no catalog directory. |
| D-2 | `CountFilePointers(tx, fileID)` sums `SELECT COUNT(*) FROM <From> WHERE <Column> = ?` over every `catalogmodel.FKs` entry with `To == "files"`, except `From == "file_derivatives" && Column == "source_file_id"`. It reads `catalogmodel.FKs` on each call; no cache. | A missed pointer loses data; the declared model makes forgetting one impossible. A File's own thumbnail link (`source_file_id`) is owned by that File, not a pointer to it. Reading on each call keeps test registrations visible; the loop is over about 60 entries. | A hand-kept list in `deleteimpact` — the bug this PR fixes. Caching at init — tests could not register an FK. |
| D-3 | The File release, when the count is 0: list the File's own derived ids (`source_file_id = ?`), report the File's object, `DELETE FROM file_derivatives WHERE source_file_id = ?`, delete the File, then release each derived File through the same count. It no longer deletes links by `derived_file_id`. | A File with any `derived_file_id` link pointing at it has a count above 0, so it is never released while such a link exists. Deleting only its own links keeps another source's link to a shared derived File. | Keeping `OR derived_file_id = ?` — it removes the other source's link (AC-1.1). |
| D-4 | `ReleaseSnapshot(tx, snap) ([]files.Object, error)` returns the object of every `files` row it deleted, collected just before each delete. `ReleaseOwned` returns the same. `Released` gains `Objects []files.Object`, and `merge` appends it. `eraseObservation` puts its `ReleaseSnapshot` objects on its `Released`. `CollectFileObjects`, `collectFileObjectsFromID`, and `FileObject` are deleted. | The release knows exactly which rows it deleted; collecting beforehand reports Files that end up kept (AC-1.3). | Keeping `CollectFileObjects` — a second walk that can disagree with the release. |
| D-5 | Owned release is generic. `SnapshotOwned` reads every `ownedFor(kind)` column with `lookupParentColumn` (no `switch` on the child). Release, per entry with a non-empty id: if `Child == "files"`, run the File release (D-3), and ignore its `Then` (those name the derivative links the File release removes). Otherwise the child is a row: read each `Then` entry's column from the child row (`lookupParentColumn(tx, then.Column, childID)`), `DELETE FROM <Child> WHERE id = ?`, then release each `Then` the same way. | S10-32's `observations.value_depiction_id` → `depiction_values` → `depiction_values.image_file_id` → `files` is then one registration. Row children keyed by `id` match `date_values` and `name_values`. | A per-table case for each new child — the copy the map review (K-2) found. |
| D-6 | A registration test pins the shapes D-5 relies on: every `ownedReleases` entry (and nested `Then`) with `Child == "files"` has `Mode == "ifUnused"`; every other entry has `Mode == "always"`; and every `Then` column's table is its parent entry's `Child` (or `file_derivatives` under a `files` child). | The generic walk trusts these; a wrong registration should fail CI, not delete data. | No test — the walk would silently do something else. |
| D-7 | Artifact delete: snapshot → DELETE → `objs, err := deleteimpact.ReleaseSnapshot(...)` → `files.PurgeAfterCommit(tx, c, objs)`. `purgeReleasedObjects` is deleted. | Same behavior, one purge path. | — |
| D-8 | Callers that release nothing with a File today (`observations.Delete`, `eraseObservation`'s callers) take the new return value and do not purge. S10-32 adds their purge when depictions own Files. | Out: no Observation-owned File exists yet. | Wiring purges now — untestable without a File owner. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `files.Object` *(new)* | type | `deleteimpact.FileObject` | Moved, so `files` can purge it; the old type is deleted. |
| `files.PurgeUnreferenced` *(new)* | func | `artifacts.purgeReleasedObjects` | Moved unchanged; the private copy is deleted. |
| `files.PurgeAfterCommit` *(new)* | func | the `tx.AfterCommit(func() { purgeReleasedObjects(c, objects) })` line in `artifacts.Delete` | One call site shape for every owner. |
| `Released.Objects` *(new)* | field | `Released.Changes` | Same carrier, merged the same way. |
| Generic release in `releaseFromSnapshot` | refactor | the `switch rel.Child` | Replaces it. |
| `TestOwnedReleaseShapes` *(new)* | test | `TestResourceAndOwnedHonesty` | Pins D-6 beside the existing honesty test. |

**Builds on.** `catalogmodel.FKs` (as `inboundEdges` does), the `ownedReleases` registry, the 30d4b18b purge order.
**Deletes.** `deleteimpact.FileObject`, `CollectFileObjects`, `collectFileObjectsFromID`, `TestCollectFileObjects`, the `switch` in `SnapshotOwned` and `releaseFromSnapshot`, `releaseFileDerivatives`' `OR derived_file_id = ?`, `artifacts.purgeReleasedObjects`.
**Next adds.** S10-32 registers `observations.value_depiction_id` → `depiction_values` (always) with `Then` `depiction_values.image_file_id` → `files` (ifUnused), adds the FKs to `catalogmodel`, and calls `files.PurgeAfterCommit` in the Observation, Subject, and Source delete paths that can now release a File.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/files` | `Object`, `PurgeUnreferenced`, `PurgeAfterCommit` per D-1, in a new file `core/database/files/purge.go` *(new)*. |
| `core/database/deleteimpact` | `CountFilePointers` per D-2; `releaseFileIfUnused` per D-3 returning `[]files.Object`; `SnapshotOwned`, `ReleaseSnapshot`, `ReleaseOwned`, `releaseFromSnapshot` per D-4 and D-5; `Released.Objects`; `eraseObservation` keeps the objects. |
| `core/database/artifacts` | `Delete` per D-7. |
| `core/database/observations` | `Delete` takes `_` for the returned objects with a one-line comment pointing at S10-32 (D-8). |

```go
// core/database/files/purge.go (sketch)
type Object struct{ ChecksumSHA256, MediaType string }
func PurgeUnreferenced(c *database.Catalog, objs []Object)
func PurgeAfterCommit(tx *database.Tx, c *database.Catalog, objs []Object)

// core/database/deleteimpact/release.go (sketch)
func CountFilePointers(tx *sql.Tx, fileID []byte) (int, error)
func ReleaseSnapshot(tx *sql.Tx, snap OwnedSnapshot) ([]files.Object, error)
func ReleaseOwned(tx *sql.Tx, kind catalogmodel.Kind, id []byte) ([]files.Object, error)
```

Errors: unchanged.

### FFI

None — no RPC changes.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — One object purge in `files`**
Files: `core/database/files/purge.go` *(new)*, `core/database/files/files_test.go`, `core/database/artifacts/artifacts.go`.
Do: D-1. Move the `purgeReleasedObjects` body into `files.PurgeUnreferenced`; add `PurgeAfterCommit`. In `artifacts.Delete`, convert the collected `deleteimpact.FileObject`s to `files.Object` for this step only, and call `files.PurgeAfterCommit`. Delete `purgeReleasedObjects`. Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/files/... ./core/database/artifacts/...` passes, including `TestArtifactDeletePurgesDisk`.

**STEP-2 — Count every declared File pointer**
Files: `core/database/deleteimpact/release.go`, `core/database/deleteimpact/release_test.go`.
Do: D-2. Add T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/deleteimpact/...` passes, including `TestCountFilePointersIfUnused`.

**STEP-3 — Release only a File's own links, and report what was released**
Files: `core/database/deleteimpact/release.go`, `core/database/deleteimpact/facets.go`, `core/database/deleteimpact/release_test.go`, `core/database/artifacts/artifacts.go`, `core/database/observations/write.go`.
Do: D-3 and D-4. Replace `FileObject` with `files.Object` everywhere; delete `CollectFileObjects` and `TestCollectFileObjects`. `artifacts.Delete` now takes the objects from `ReleaseSnapshot` (D-7) and drops the STEP-1 conversion. `observations.Delete` per D-8. Add T-3 and T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**STEP-4 — Generic owned release**
Files: `core/database/deleteimpact/release.go`, `core/database/deleteimpact/release_test.go`, `core/database/deleteimpact/pragma_test.go`.
Do: D-5 and D-6. Add T-5, T-6, and T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes, including `TestResourceAndOwnedHonesty` and `catalogmodel.TestPragmaHonesty`.

**STEP-5 — Artifact delete end to end**
Files: `core/database/artifacts/artifacts_test.go`.
Do: Add T-8 (AC-1.1, AC-1.2 through `artifacts.Delete` and `writes.Run`).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/artifacts/...` passes.

**STEP-6 — Docs**
Files: `core/database/deleteimpact/README.md`, `.agents/skills/add-catalog-delete/SKILL.md`, `docs/catalog-deletes.md`, `docs/source-layer-data-model.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do:
- `core/database/deleteimpact/README.md` and `docs/catalog-deletes.md`: a File is released when no `catalogmodel` FK into `files` points at it (a File's own thumbnail links do not count); owned row children and their `Then` entries release generically; `ReleaseSnapshot` returns released objects, and the owner calls `files.PurgeAfterCommit`.
- `add-catalog-delete` skill: under "New table", a table that owns a File registers its FK in `catalogmodel` and one `ownedReleases` entry (row child `always`, `files` child `ifUnused` under `Then`), adds no `deleteimpact` code, and its delete path calls `files.PurgeAfterCommit` with the objects `ReleaseSnapshot` returns.
- `source-layer-data-model.md` §8 (`file_derivatives`): releasing a File removes only its own derivative links; a derived File is deleted when nothing points at it.
- `deployment-plan.md`: the S10-51 **Plan:** line and SLICE 8 line read "— done, #N"; `completed.md`: a `## S10-51 — File release from the declared model` section at the top in the S10-50 shape.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-51-file-release.md` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/files/files_test.go` `TestPurgeUnreferenced` *(new)*, table-driven | an object whose checksum has no `files` row → removed; one whose checksum still has a row → kept; an empty list → no error. `PurgeAfterCommit` on a rolled-back `writes.Run` → nothing removed | AC-1.4, D-1 |
| T-2 | Go | `core/database/deleteimpact/release_test.go` `TestCountFilePointersIfUnused` (extend) | a File that is `derived_file_id` of one link → count 1; the same File as `source_file_id` only → count 0; with a test FK appended to `catalogmodel.FKs` (restored in `t.Cleanup`) on a temp table created in the test → counted | AC-2.1, D-2 |
| T-3 | Go | `release_test.go` `TestReleaseKeepsSharedDerived` *(new)* | Files `A`, `B`, derived `D` with `A → D` and `B → D` thumbnail links; `ReleaseSnapshot` for an Artifact on `A` → `A` and `A → D` gone, `D` and `B → D` stay; returned objects are `A` only | AC-1.1, D-3, D-4 |
| T-4 | Go | `release_test.go` `TestReleaseReportsReleasedObjects` *(new)*, table-driven | lone Artifact on `A` with thumbnail `T` → objects `A` and `T`; two Artifacts on `A`, release one → no objects; fileless Artifact → no objects | AC-1.3, AC-1.4, D-4 |
| T-5 | Go | `release_test.go` `TestGenericOwnedRelease` *(new)* | in the test catalog: `ALTER TABLE observations ADD COLUMN test_value_id BLOB`, `CREATE TABLE test_values (id BLOB PRIMARY KEY, file_id BLOB)`; append to `ownedReleases` (restored in `t.Cleanup`) `{Parent: KindObservation, Column: "observations.test_value_id", Child: "test_values", Mode: "always", Then: {Column: "test_values.file_id", Child: "files", Mode: "ifUnused"}}` and the FK `test_values.file_id → files` to `catalogmodel.FKs`; delete an Observation that owns one → the `test_values` row and its File are gone and the File's object is returned; a second Observation sharing that File through its own row keeps it until both go | AC-2.2, D-5 |
| T-6 | Go | `release_test.go` `TestObservationValuesStillRelease` *(new)* | an Observation with a date value and one with a name value, deleted through `ReleaseSnapshot` → `date_values` and `name_values` rows gone, no objects | AC-2.3, D-5 |
| T-7 | Go | `core/database/deleteimpact/pragma_test.go` `TestOwnedReleaseShapes` *(new)* | walks `ownedReleases` per D-6 | D-6 |
| T-8 | Go | `core/database/artifacts/artifacts_test.go` `TestArtifactDeleteKeepsSharedThumbnail` *(new)* | two Artifacts on Files `A`, `B` whose `thumbnail` links share derived `D` (inserted with `filederivatives.Insert`); `artifacts.Delete` the first through `writes.Run` → `D`, `B → D`, and `D`'s object stay; delete the second → all gone and both objects removed | AC-1.1, AC-1.2, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestArtifactDeletePurgesDisk`, `TestResourceAndOwnedHonesty`, `catalogmodel.TestPragmaHonesty` (T-2 and T-5 must restore every registry they touch), the Subject delete tests that reach `eraseObservation`.

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
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: catalog-deletes.md, source-layer-data-model.md §8, deleteimpact README, add-catalog-delete skill
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project. Add two Sources, each with an Artifact made from the same JPEG saved under two names with one byte changed at the end of the file (so the Files differ but their thumbnails are the same pixels). Both cards show a thumbnail.
2. Delete the first Artifact. The second Artifact still shows its thumbnail after reopening the Source.
3. Delete the second Artifact. Its thumbnail is gone, and the project's objects folder has no file for either scan.

An implementer without a Mac reports this check as not run; T-8 covers the same behavior.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A test registration leaks into another test | T-2 and T-5 restore `catalogmodel.FKs` and `ownedReleases` in `t.Cleanup` and do not call `t.Parallel`. |
| The generic walk deletes a row the old switch ignored | Only `file_derivatives` was ignored, and it appears only under a `files` child's `Then`, which D-5 skips. T-7 pins it. |
| S10-44 (planned) also edits `core/database/deleteimpact/facets.go` | It changes `releaseSubjectClaims`; this PR changes `Released` and `eraseObservation`. Rebase on whichever lands first. |

**Stop and report if:**

- A `files` row is deleted anywhere other than the File release (`grep -rn "DELETE FROM files" core`).
- A `catalogmodel.FKs` entry into `files` exists that is not one of the three in §4, at the base you start from.
- `TestArtifactDeletePurgesDisk` needs its expectations changed to pass.
- T-5 cannot add a column to `observations` in the test catalog without breaking the schema-hash check on that catalog; report what failed instead of weakening the check.

## 12. Open questions

None.

## 13. PR

**Title:** S10-51 — File release from the declared model
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-51-file-release.md`.
