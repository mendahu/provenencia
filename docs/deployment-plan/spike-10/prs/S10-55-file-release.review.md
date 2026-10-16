# Review — S10-55 File release from the declared model

**Verdict:** Fits with changes — Ready with changes (all applied).
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None.

## Writer's own calls

- D-1: `database.Tx` carries its catalog, so any delete can purge without a new parameter.
- D-4: the separate `ownedReleases` list is deleted. Owned releases come from `catalogmodel`'s owned-outbound FKs.
- D-6: `ReleaseReplaced` is the update path's slot. The existing date and name update releases stay where they are, because they also emit audit changes.
- D-8: every delete that releases purges, including deletes that release nothing with a File today.
- Out: File Impact edges from the declared pointer set wait for the first File delete.

## Round 1

### M-1 · blocker · S10-32's region edit has no release slot
**Answer:** Fixed. D-6 adds `ReleaseReplaced(tx, column, oldID)`, and T-5 covers it. Moving `releaseDateValue` and `releaseNameValue` onto it is declined: they emit `date_value` and `name_value` audit changes the shared release does not. It is listed in Out.

### M-2 · blocker · The purge is wired by hand per caller
**Answer:** Fixed. D-1 gives `Tx.Catalog()` and `files.PurgeAfterCommit(tx, objs)`. D-8 wires all six releasing deletes in this PR, and T-8 covers it.

### M-3 · major · `ownedReleases` restates `catalogmodel`
**Answer:** Fixed. D-4 derives the owned plan from the `BucketOwnedOutbound` FKs and deletes `ownedRelease`, `ownedReleases`, `ownedFor`, `ownedColumns`, and the honesty test's owned branch. D-9 and T-7 pin the shapes. STEP-7 rewrites the working rule's wording.

### M-4 · minor · Row-child key hard-coded as `id`
**Answer:** Fixed. D-5 reads `PK` from `catalogmodel`; `date_values` and `name_values` gain `PK: "id"`. `effects.pkOf` already defaults to `id`, so nothing else changes.

### M-5 · minor · Tests mutate process-wide registries and the live schema
**Answer:** Fixed. D-2 and D-4 take an explicit FK slice behind `sync.OnceValue`, and the tests use temp tables with no `ALTER TABLE` on a real table.

### M-6 · minor · `file_derivatives` `Then` entries the walk ignores
**Answer:** Fixed by D-4: `file_derivatives` is never a `To`, so it is reached only through the File release.

### M-7 · minor · `Impact(KindFile)` keeps its own pointer list
**Answer:** Logged. The Out row sends it to the first PR that adds a File delete. No File delete exists today.

### M-8 · nit · Where object removal lives
**Answer:** Fixed. D-1 names the removers in the `purge.go` doc line.

### X-1 · blocker · STEP-3 misses callers it breaks
**Answer:** Fixed. STEP-3 lists `impact_test.go` and names every caller to update.

### X-2 · blocker · No Observation fixture for T-5 and T-6
**Answer:** Fixed. T-5 uses temp tables with an explicit plan. T-6 is the existing `TestObservations` row "delete drops owned date and name values".

### X-3 · major · The test FK literal is unspecified
**Answer:** Fixed. §8 gives the literals.

### X-4 · major · T-1's rollback case is under-specified
**Answer:** Fixed. T-1 says how the object is written and how the rollback is forced.

### X-5 · minor · Dedupe and object order
**Answer:** Fixed. D-1 says a duplicate object is harmless, and D-3 fixes the order (the File first, then its derived Files).

### X-6 · minor · D-5 edge cases
**Answer:** Fixed. D-5 skips a NULL column, a missing child row, and a NULL child column, and T-7 rejects cycles.

### X-7 · minor · Status
**Answer:** Fixed by this log. The plan stays Draft until the review rounds end.

### X-8 · nit · Checks that cannot run here
**Answer:** Fixed. STEP-7's check now has a `grep`. The manual check already says what to report without a Mac.

## Round 2

### X-9 · blocker · T-1 cannot compile once `deleteimpact` imports `files`
**Answer:** Fixed. T-1 moves to `core/database/files/purge_test.go`, which is `package files_test`.

### X-10 · blocker · `releaseFile` cannot count from a test FK slice
**Answer:** Fixed. D-2 defines `ownedPlan{pointers, edges, pks}`. `releaseFile` and `countFilePointers` take the plan.

### X-11 · blocker · STEP-3 needs a type STEP-4 creates
**Answer:** Fixed. STEP-2 adds `ownedPlan` with `pointers` only, STEP-3 passes it, and STEP-4 fills `edges` and `pks`.

### X-12 · major · T-5's entry point is unnamed
**Answer:** Fixed. D-4 names `releaseOwnedRow(tx, p, table, id)`. The test tables now include `test_parent`, and `lookupParentColumn` reads the table's `PK`.

### X-13 · major · STEP-7's grep fails
**Answer:** Fixed. STEP-7 names each map row and doc line, and widens the grep (see M-9).

### X-14 · minor · Small choices
**Answer:** Fixed: T-1's `MkdirAll`; T-8 uses `toponymProp` with `writes.Call`; `observations.Delete` discards its return until STEP-5.

AC-2.5's evidence is partial in this PR. T-8 shows only that an empty purge is harmless; S10-32's T-8 and T-9 are the first tests that release a File through observations and subjects. The S10-44 import note is a Risks row.

### M-9 · major · Skill and rule lines this PR makes false
**Answer:** Fixed. STEP-7 adds `.cursor/rules/catalog-migrations.mdc` and `catalog-deletes.mdc`, the add-catalog-model and add-catalog-delete lines, and `docs/catalog-deletes.md` lines 9, 10, 15, and 96. The grep is widened to `release list`, `release entry`, `ownedReleases`, and `CollectFileObjects`.

### M-10 · minor · The precedent row overstates `inboundEdgeList`
**Answer:** Fixed. §4 names `inboundEdgeList` as the `sync.OnceValue` shape and `tableByKind` as the read-from-the-model precedent.
