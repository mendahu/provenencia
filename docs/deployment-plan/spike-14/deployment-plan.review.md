# Review — Spike 14 PR map (batch 1)

**Verdict:** Fits with changes
**Rounds:** 1. **Reviewers:** maintainability (fresh agent, map review)

## To user

- **K-2 — One `picture` key for Event and Place?** `start_date`/`end_date` already bind one Property to several kinds, and two keys break the composer's run rule (`applyObservationSubject` drops a Property the next Subject does not offer). The user chose two keys for separate descriptions. Recommendation: the user's call; one key is smaller.
- **K-3 — Order against Spike 13's gallery.** S13-14 and S13-15 rework the same `ConclusionDetailLikenessSection`. Recommendation: S14-02 lands the kind-neutral gallery slot first, and S13-14 builds on it (a line in each map).
- **K-5 — Rename the gallery layer.** After S14-01, "likeness" names both the Person key and every gallery type (`likenesses`, `LikenessTile`, `LikenessDetail`, `CatalogLikeness`, …). Recommendation: a preparatory mechanical rename to "gallery" (wire field numbers kept), ahead of S14-01; if declined, log it in the ideas file.

## Findings

### K-1 · major · One gallery slot on the body
**Answer:** Fixed — working rule "The gallery is one slot on the detail body"; S14-02 Ships; design § What changes.

### K-2 · major · One `picture` key instead of two
**Answer:** To user.

### K-3 · major · Spike 13 and Spike 14 both rework the gallery section
**Answer:** To user.

### K-4 · major · The Promote evidence sheet calls pictures likenesses
**Answer:** Fixed — new S14-04 (its own view, so its own PR); bar outcome 5; design § What already works (Promote) and § Stays later corrected.

### K-5 · major · "Likeness" names two concepts
**Answer:** To user. K-11 folds in.

### K-6 · major · Parameterize the test helper
**Answer:** Fixed — S14-01 Ships: `SaveDepiction(…, key, …)`, `SaveLikeness` wraps it; `add-depiction-value` Tests line in the docs list.

### K-7 · major · Reuse the bounding geometry
**Answer:** Fixed — S14-03 Ships: `ArtifactRegionDraft.boundingRectangle` from `boundingRect(of:)` and `rectanglePoints`; tool list as composer state.

### K-8 · minor · Portrait enforced only by the caller
**Answer:** Fixed — working rule "No preferred picture": the Go pointer stays generic and the app limits it, written into `conclusion-layer-data-model.md` by S14-01; no tile marked as portrait without one.

### K-9 · minor · Guard the gallery-key list
**Answer:** Fixed — S14-01 Ships: a `subjectvocab` test that every gallery key is a seeded `depiction` Property.

### K-10 · minor · Docs and strings not listed
**Answer:** Fixed — design § Domain docs adds `conclusion-reconciliation.md` and the skill's description and Do-not line; S14-02 Ships the translator comments and `depictionPerson`.

### K-11 · nit · Promote and thumbnail names
**Answer:** Logged; folds into K-5.

## After round 1

The user asked for thumbnail parity with a Person's portrait (2026-10-27). The working rule "No preferred picture for an Event or Place" is replaced by "An Event or Place chooses its thumbnail as a Person chooses its portrait", on the one existing mechanism (K-8's generic Go pointer), and S14-05 – S14-07 are added. K-8's "app limits it" answer no longer applies: all three kinds have the portrait.
