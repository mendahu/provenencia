# Spike 14 — Odds and ends

**Done** (closed 2026-11-08). The first odds-and-ends spike: eight batches of small features, audit fixes and dogfood UX, each landing on its own, ending with summary rows that record their inputs so a write finds what it moved by lookup. It also built Spike 12's one-graph Citation, seeded FakeStore and Go-owned open allowlist ([Spike 12](../spike-12/README.md)). Live models: [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md), [`seeded-vocabulary.md`](../../../seeded-vocabulary.md), [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md), [`catalog-deletes.md`](../../../catalog-deletes.md), [`macos-client-patterns.md`](../../../macos-client-patterns.md), [`structured-name-model.md`](../../../structured-name-model.md), [`matching-model.md`](../../../matching-model.md). Design reasoning: [`summary-inputs.md`](summary-inputs.md).

## Decisions

- **Even spikes are odds and ends.** Small features, refinements and fixes land in batches, each a slice with its own Check; large features go to odd spikes.
- **Events and Places have pictures.** One locked depiction Property per kind, `event_picture` and `place_picture`, both labelled Picture; `likeness` stays the Person key. New seeds reach new catalogs only.
- **The gallery layer is named gallery.** "Likeness" stays only for the Person key, its copy and the portrait; `canonical_entities.preferred_likeness_observation_id` keeps its name.
- **One gallery for every kind.** The gallery reads a list of gallery keys in `core/vocabkey`; one preferred pointer, one lead rule and one gallery slot on `ConclusionDetailBody` serve Person, Event and Place. An Event or Place chooses its thumbnail as a Person chooses its portrait.
- **A depiction is drawn as a rectangle.** A Citation's own region keeps all seven tools; an old polygon depiction loads and saves as it is.
- **Oversized JPEGs get a reduced decode of our own.** Over the 64-megapixel budget, a pure-Go decoder reads only DC coefficients at 1/8 scale, bounded and fuzzed, so a Windows build reuses it. Other formats over the budget stay refused.
- **A hidden graph or Source page goes stale; it does not reload.**
- **File work happens before the catalog session.** One staged-object type in `ingest` streams, hashes, sniffs and decodes outside `catalogsession.Do`.
- **Kit components announce themselves.** A danger confirm's default is Cancel; a dirty form asks before a link inside it navigates away.
- **Every audited write returns its effects**, held by a test in `api/ffi`. Sidebar counts are an effects list, and every list reaches its keys through its descriptor.
- **Proposals are liberal; the researcher decides.** Other-link support is 0.75, so a thin name plus one lined-up link clears Weak.
- **A nickname folds into its set's formal name and reappears as a nick part:** *James “Jim” Robins*. Former surnames show in parentheses, italic on the titles.
- **One width breakpoint: 1600pt of pane.** Pages adopt Page columns in their own redesigns.
- **Boolean is a first-class value type.** Single-only, a Yes / No segmented control with nothing chosen at first, no Negative checkbox.
- **Smaller calls.** Conflict brackets honour cardinality; Tab reaches a following text area through `PVTabPair`; place relationship rows read the inverse label at the far end; creating a bridge between two handles files it; the Persons sort ignores honorific prefixes.
- **The structural fix comes before the patch.** A bug that needs a patch at each call site waits for the mechanism to be replaced.
- **Summary rows record their inputs; invalidation is a lookup.** An input is a handle's values, one bridge's links at one end, or an association's values; a row is its own input; isolated composition is the oracle. One input index, owned by `searchindex`, with no foreign key; every write reports what it moved through `summary_handles`. The hand-written inverse (`Dependents`, `AssociationEndpoints`, `linkEnds`, `SetDisplayDeps`) is gone.

## Refused

A transparent-polygon depiction crop. A per-kind thumbnail column or RPC. A Property named `depiction`. A separate Page columns PR (built with Spike 13's Person page). Promote picture wording (the evidence sheet went in Spike 13).

## Leftovers

| Item | Home |
| --- | --- |
| A loading caption while the viewer opens; a drawing hint that says "tools above" for one tool; EXIF orientation; a larger cover derivative; pictures as matching evidence for an Event or Place | [`depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) |
| A reloading sign on a revisited canvas | [`evidence-graph-followups.md`](../../../ideas/evidence-graph-followups.md) |
| A failed sidebar recount is silent | [`shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) |
| A failed bind after a Property create | [`properties-and-terms.md`](../../../ideas/properties-and-terms.md) |
| "Soby-Kirke" misses "Søby Kirke" | [`international-names.md`](../../../ideas/international-names.md) |
| A name part length bound on Validate and Insert | [`share-packages.md`](../../../ideas/share-packages.md) "Harden before sharing" |
| A field relabel reaching Source pages beside its effects; a new-spelling join reprojecting the city; a two-term filed-link set and its repeated link key | [`structural-debt.md`](../../../ideas/structural-debt.md) |
| The two app-health audits' unscheduled items | Dropped with the audits (user, 2026-11-08); in git history |
| Polygon vertex add, keyboard nudge, VoiceOver per vertex | Dropped (user, 2026-11-05) |
| Dropped: a thumbnail choice reloading three lists; a term create reloading Properties (2 ms); a broken oversized JPEG re-decoded per response (off the session since S14-12); copy nits ("A smaller copy of the file will open", a generic "Discard changes?", a race with another window's unlink); a test for the localizable script; a split of `check-guidance-names.py`; code nits in `evidencegraphs`, `FakeStore` and `L10n.Promote` | Not worth a PR; no measured cost (2026-11-08) |
