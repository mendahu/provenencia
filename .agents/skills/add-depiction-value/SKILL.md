---
name: add-depiction-value
description: >-
  Uses or extends Provenencia depiction values (pure type in core/depictionvalue, queries in core/database/depictionvalues):
  a region of a Citation's page plus an image File, the gallery and the preferred portrait or thumbnail.
  Use when saving, reading, redrawing or reconciling a depiction, value_depiction_id,
  depiction_values, likeness, gallery, Readable, FromJoin, StageDepictions, the gallery or the portrait.
---

# Add or use a depiction value

A depiction is a region of a Citation's page and the image File Go stored from it. The pure type and its read rules live in **`core/depictionvalue`**; the queries live in **`core/database/depictionvalues`**.

The macOS client draws a depiction with the rectangle tool alone (`CitationComposerModel.viewerRegionTools`), and the saved image is the region's bounding box; Go accepts any valid region, so a polygon saved earlier opens, shows and saves as it is.

## Package

```
core/depictionvalue/                       # pure: no core/database import
  depictionvalue.go                        # Value, ErrInvalid, Validate, Readable, ReadableImage
  validate_test.go
  readable_test.go
  depictionvaluetest/depictionvaluetest.go # Locator, Region (fixture consts)
core/database/depictionvalues/             # queries
  depictionvalues.go                       # InsertTx, LookupTx, FromJoin
  depictionvalues_test.go
```

Matching layers and the reconciler import `core/depictionvalue`, never `core/database/depictionvalues`. `TestLayerPackages` in `core/matchregistry/guard_test.go` keeps both pure packages pure.

## Model

An owned value row: a region plus an image File, replaced never rewritten (a new region always comes with a new image). The image is not an Artifact and not a thumbnail; Go re-encodes its own PNG (`ingest.StageImage`, `core/raster`). See [`docs/interpretation-layer-data-model.md`](../../../docs/interpretation-layer-data-model.md) and the [Spike 10 decisions](../../../docs/deployment-plan/archive/spike-10/README.md).

## Write path

- Staging and request rules are `add-ffi-handler`'s ([Files from the client](../add-ffi-handler/SKILL.md)); the depiction call is `ingest.StageDepictions` through `withStagedDepictions`, which returns `ingest.Staged` stages (`InsertTx` inside the write, `Settle` after it).
- `observations.Input.Depiction` (`observations.DepictionInput`) carries the value. `resolveDepiction` checks `locator.ValidateDepictionRegion`, inserts the File, then `depictionvalues.InsertTx`, which trims and calls `depictionvalue.Validate`.
- A redraw releases the old row (`releaseReplacedValue` in `core/database/observations/update.go`). Release follows the `catalogmodel` owned-outbound FKs (`observations.value_depiction_id`, `depiction_values.image_file_id`).
- Audit follows [`add-audit-scope`](../add-audit-scope/SKILL.md).

## Read path

Every reader that returns a depiction to the app goes through `depictionvalue.Readable`, usually through `depictionvalues.FromJoin` and the three joins its doc names. `LookupTx` does not (audit records what is stored). Returning paths follows `add-ffi-handler`.

## Reconcile, gallery and portrait

- Depictions group by image File id (`core/autoreconcile/modules.go`).
- The gallery is `gallery.ForEntities`, admitted by `autoreconcile.Reason.InGallery` and filtered on `vocabkey.GalleryKeys()` (`likeness`, `event_picture`, `place_picture`); the rule is in [`docs/conclusion-layer-data-model.md`](../../../docs/conclusion-layer-data-model.md) §6 and [`docs/conclusion-reconciliation.md`](../../../docs/conclusion-reconciliation.md). Do not restate it here.
- Each gallery tile's image is the Observation's own `depiction_values.image_file_id`, not the reconciler cache's `value_image_file_id`.
- `gallery.Lead` is the one lead rule. The preferred portrait or thumbnail is `canonical_entities.preferred_likeness_observation_id`, written by `canonicalentities.SetPreferredLikeness`.
- Tests save depictions with `observationstest.SaveDepiction` (or `SaveLikeness` / `SaveNegativeLikeness` for the Person key).

## Tests

- Pure tests use `depictionvaluetest.Locator` and `depictionvaluetest.Region`; check with `go test ./core/depictionvalue/...`.
- Database tests use `observationstest.SaveDepiction`, `SaveLikeness`, `Crop`, `LikenessRegion`, `ImageOf`, `FileArtifact`, `SaveCitation`; run `CGO_ENABLED=1 go test -tags fts5 ./core/database/depictionvalues/ ./core/database/gallery/`.

## Do not

- Store the app's bytes.
- Read a stored depiction without `Readable`.
- Add a second depiction struct.
- Add a likeness field to the Person page (the gallery is the display).
- Import `observationstest` from a pure package.

The request and staging rules are `add-ffi-handler`'s.
