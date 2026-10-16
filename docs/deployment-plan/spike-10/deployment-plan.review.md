# Map review — Spike 10, Slice 8 (Likeness) re-cut

**Verdict:** Fits with changes (all applied)
**Rounds:** 1. **Reviewer:** maintainability (opus), map questions from `review-plan`. **Date:** 2026-10-16.

Scope: the Slice 8 re-cut after the user moved a depiction's value from "its own Citation" to an owned row holding the region and an extracted image.

## To user

- K-1, answered 2026-10-16: the value points straight at its image File. `file_derivatives` stays disposable thumbnails only.

## Writer's own calls

- S10-51 (file release and purge) and S10-53 (image staging) are two preparatory PRs, not one.
- S10-54 is a separate FFI PR; S10-34 does not carry the proto change.
- Images over 2048 on the long edge are scaled down, not refused.
- A Citation on a fileless Artifact cannot hold a depiction.
- The image `files` row is audited (`create_file`), like an import; thumbnails stay unaudited.

## Round 1

### K-1 · blocker · `file_derivatives` would hold two different kinds of file
Derivatives are "reproducible and disposable" (`source-layer-data-model.md` §1.5, §8, invariant 18), and a depiction image is neither.
**Answer:** To user, then Fixed. The user chose `depiction_values.image_file_id` → `files`, with lineage through region → Citation → Artifact → File. The working rule, the S10-32 scope, the gotcha, and the design note's Settled table now say so. S10-51 no longer rebuilds `file_derivatives`.

### K-2 · blocker · File release is hand-coded and cannot count a column that does not exist yet
**Answer:** Fixed. S10-51 counts every `catalogmodel` FK into `files` and walks `ownedReleases` `Then` generically. The working rule "A File is released only when nothing points at it…" records this, and the add-catalog-delete skill is in "Docs to update".

### K-3 · blocker · Only Artifact delete removes files from disk
**Answer:** Fixed. S10-51 has `ReleaseSnapshot` report released objects and adds one shared after-commit purge; Artifact delete moves onto it. A staged image discards its own object on failure (S10-53).

### K-4 · blocker · Image work cannot run inside the write
**Answer:** Fixed. S10-53 adds the two-phase `StageImage` / insert / `Discard` API, and S10-32 adds one core entry point that every handler calls (S10-54).

### K-5 · major · Neither the ingest checks nor the raster decode can be reused as is
**Answer:** Fixed. S10-53 extracts the open-and-check step from `ingest.File` and one bounded decode from `raster.EncodeJPEG`, shared with a new PNG encoder. The contradiction is settled: an image over 2048 is scaled down.

### K-6 · major · S10-32 is not one plan of about 10 steps
**Answer:** Fixed. The FFI moves to S10-54, so S10-32 is engine only and keeps the Swift label the parity test needs. S10-34 and S10-36 depend on S10-54.

### K-7 · major · Code that assumes one derivative per (File, type)
**Answer:** Not needed. Under K-1, `file_derivatives` keeps one thumbnail per File, so `Lookup`, `LookupMany`, and `Ensure` stay correct. The gotcha says a depiction image has no `file_derivatives` row.

### K-8 · minor · The depiction region needs one shape and one validator
**Answer:** Fixed. The region is a locator document, checked by one named `core/locator` rule. Go derives the source File from Citation → Artifact, and a fileless Artifact is refused.

### K-9 · minor · Where S10-39's pointer lives, and one definition of "earliest"
**Answer:** Fixed. S10-39 names `canonical_entities.preferred_likeness_observation_id` with `ON DELETE SET NULL`, read through S10-35's admit rule. S10-35 defines "earliest" once, in the shared Go read.

### K-10 · minor · Skills and docs the slice makes false
**Answer:** Fixed. "Docs to update" has rows for add-catalog-delete (S10-51), the `core/derivatives` package doc (S10-53), and add-audit-scope and add-seeded-vocabulary (S10-32). The design note no longer claims the image is a derivative, so `source-layer-data-model.md` §8 stays true.

### K-11 · minor · S10-52 inherits its cleanup from a plan marked "do not implement"
**Answer:** Fixed. The S10-52 scope names the `knownValueType`, `Carries`, and `hasValue` cleanup directly.
