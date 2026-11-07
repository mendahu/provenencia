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

## Re-cut (2026-10-27)

- **K-2:** the user kept two keys (data model first, the UI adapts); the run cost is recorded in the working rule.
- **K-3:** the user keeps the old and new galleries side by side, so no dependency. S13-14's scope now says it builds from S14-02's slot.
- **K-5:** taken, inside S14-01 (Go, proto, wire mapping) and S14-02 (Swift view layer), not as a separate PR. Working rule "The gallery layer is named gallery, not likeness".
- **Consolidation:** S14-04 – S14-07 folded into S14-01 and S14-02 (working rule "Batch 1 is cut by layer"), IDs kept as Descoped.

## Batch 8 map review (2026-11-07)

One fresh maintainability reviewer on Batch 8 and its design, as first cut (four PRs). The user then asked for two PRs; the answers apply to the two-PR cut. **Verdict:** Fits with changes. All findings Fixed.

### K-1 · blocker · S14-42 deleted `Dependents` while `subjects.Delete` still called it until S14-43
**Answer:** Fixed by the two-PR cut: S14-43 is merged into S14-42. If S14-42's plan runs past about 10 steps, Subject delete splits back out after it and takes the `Dependents` delete with it (slice note). S14-42 now says its plan places handle reprojection and the index write so `autoreconciler.Rebuild` and `graphcachetest` still rebuild documents.

### K-2 · blocker · A row's reads of its own handle skip `Node`
`gallery.ForEntities` (lead picture) and `searchindex` `sqlHandleHeads` / `sqlHandleValues` read with raw SQL; a preferred-picture write moves no value row.
**Answer:** Fixed — new working rule "A row is its own input": every handle in `Set.Handles` is rewritten and dropped whether or not its values moved, and the index holds only inputs from other handles. Design § A row is its own input. S14-42 adds a preferred-picture acceptance test.

### K-3 · blocker · Index table: CASCADE would bring S14-38 back; two writers
**Answer:** Fixed — working rule "One input index, owned by `searchindex`": no foreign key; `ReprojectHandles` also writes association ends; `ClearAll` and `RebuildAll` cover it; `ProjectionVersion` bumped. Design § Storage.

### K-4 · major · `SetDisplayDeps` / `forgetDisplays` / `SummaryDepsWalked` become dead
**Answer:** Fixed — S14-42 deletes them; displays drop from `SummaryDeps` only; the test kit passes `SummaryDeps` or commits through `writes.Run`; `detachAssociation` and `mergedNeighbors` stay for node reloads and memo roots.

### K-5 · major · "read" already names something in `effects`
**Answer:** Fixed — the noun is "input" (input index, input key, input scope); the design file is renamed `summary-inputs.md`.

### K-6 · major · Rule and skill updates
**Answer:** Fixed — S14-40 adds the composer rule to `.cursor/rules/catalog-queries.mdc` and a line to `add-searchable-kind`; S14-42 rewrites the `canonicalgraph.Walk` line and the autoreconciler, searchindex and `subjects.Delete` docs.

### K-7 · major · Put the scope in `graphcache`
**Answer:** Fixed — working rule "Inputs are attributed per row …" binds a `graphcache.Graph` input scope keyed by a dependent id.

### K-8 · major · Key table, oracle and invalidation disagree
**Answer:** Fixed — (a) the link key covers every link of the bridge at that end, so a filtered hop sees a role or type change, and "moved" is in the table; (b) `Follow`'s internal `Node` records nothing more; (c) a row records the place-chain closure from its places with each part_of association's values; (d) S14-42 adds a `graphcachetest` merge test, and the design notes no production merge exists.

### K-9 · minor · Agreement with `Dependents` will not be identical
**Answer:** Fixed — S14-40's check is "subset, and the difference is exactly the detail-page hops"; the design names them and step 5 covers them.

### K-10 · minor · `HandlesObservingSubject`; bar 34 wording
**Answer:** Fixed — S14-42's plan decides between a `subjects` registry path and the released facets, and deletes `HandlesObservingSubject` if unused; bar 34 reads "Subject delete recomputes nothing itself".

### K-11 · minor · S14-42 size; name its test moves
**Answer:** Fixed — the tests are listed in S14-42's scope entry; the split-back rule covers size.

### K-12 · nit · Renaming a term leaves handle documents stale
**Answer:** Logged — `docs/ideas/structural-debt.md` row "Spike 14 map review K-12".
