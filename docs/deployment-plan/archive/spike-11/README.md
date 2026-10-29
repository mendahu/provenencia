# Spike 11 — Promote matching refinement

**Done** (closed 2026-10-29). Matching became one stack of pure, registry-driven layers that Promote, the reconciler and search borrow from, and the old `core/match` code and its duplicates were deleted as each layer landed. Live models: [`matching-model.md`](../../../matching-model.md) (the layers, glossary and knob inventory), [`matching.md`](../../../matching.md), [`promote-graph-alignment.md`](../../../promote-graph-alignment.md), [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md), [`structured-name-model.md`](../../../structured-name-model.md). Design reasoning: [`comparison-layers.md`](comparison-layers.md).

## Decisions

- **Matching is six pure layers over one registry.** Value, property, Subject, entity and segment compare, then graph alignment. A layer reads only the layers under it and its own section of `core/matchregistry`; pure means no build tags and no `core/database` import. One guard holds the registry to the knob inventory, and one skill, [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), teaches every layer.
- **Value compare returns a relation, a similarity and a detail**, and every value type has a normalized `Key` that agrees with it. The reconciler, search scoring, the term duplicate check and Swift's `TextMatch` share that key.
- **Entity compare is against the reconciled conclusion.** The profile is strict: displayed values positive, denied values negative, nothing else.
- **Bands come from evidence.** Similarity only picks candidates and gives the first sort. A conflict shows a Contested mark and never caps the band. (Since 2026-10-29 the band is deprecated as a measure of membership: [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §5.)
- **Frequency weighting is gone** for names, event types and toponyms.
- **Lifespan is for people only:** a window and a span from every dated event a person takes part in. A span outside the other side's window is a large conflict; agreement weighs almost nothing.
- **Neighbor credit travels one hop**, and the walk is best-first. Candidates come from a shortlist index, warmed behind the open spinner so the first proposal pays only for its Source.
- **Proposals are liberal; the researcher decides.** The candidate gate is 0.4, so a one-name Subject reaches a record with more names, and a row below the accept bar comes back with no target and its candidates as alternatives. A name prefix is an initial generalised (`Contains` at a prefix credit).
- **Kinship is matched in any shape through implied links.** Two subjects of one marriage imply `spouse`; a parent and the subject of a birth imply `parent`; a parent of a parent implies `grandparent`. Product terms only, no new weight, counted once across sides, and a pivot with more than ten participants implies nothing. The subject is an event's principal. The canon's implied links live in the graph cache, and one function returns a person's direct and implied links.
- **The reconciler keeps its own decisions on top of value compare.** Values group on `Same` and fold on `Contains`; `Equivalent` never groups; a nickname folds into the formal name its product seed gives it.
- **Gender meets sex at birth through a registry crosswalk**, so `man` agrees with `male`.
- **A record Promote did not offer is a decided row.** The target menu searches every record of the kind through the one kind-scoped catalog search; no new RPC.
- **Connections show as subrows of the rows they join** and file by their ends; there is no per-connection switch.
- **The loaded Evidence graphs are bounded** by an LRU in the graph cache.
- **Superseded 2026-10-29:** the claim evidence sheet, its pins and the per-row evidence query this spike built are being removed by Spike 13's slice 0 ([`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §5).

## Refused

Name frequency weighting. A `spouse` participation role. Damped multi-hop credit, two-hop lifespan rules and a coarse misspelling shortlist key (start simple; revisit with evidence).

## Leftovers

| Item | Home |
| --- | --- |
| Quote nicknames and add former surnames (S11-29, planned, never built) | Spike 14, S14-09 |
| The claim evidence sheet, pins and evidence query; a related person's lines and the evidence button's hover (S11-52, S11-34, descoped) | Removed by [Spike 13 slice 0](../../spike-13/deployment-plan.md); membership comparison in [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §5.1 |
| The rest of the matching test suite; matcher calibration | [`matching-test-suite.md`](../../../ideas/matching-test-suite.md) |
| Merge hints and weak-claim review views | [`identity-claim-review.md`](../../../ideas/identity-claim-review.md) |
| Term affinity, Daitch–Mokotoff, the place gazetteer | [`daitch-mokotoff.md`](../../../ideas/daitch-mokotoff.md), [`place-gazetteer-service.md`](../../../ideas/place-gazetteer-service.md) |
| Coverage-aware similarity; a narrower canon walk; the name index kept on commit | [`coverage-aware-similarity.md`](../../../ideas/coverage-aware-similarity.md), Spike 12 S12-51, S12-43 |
| Ship-log follow-ups: stale graph-cache link after a merge (verify mode), one keyed warm slot, display rows on memo slots | Spike 12 S12-43 |
| Ship-log follow-ups: a term key shared across origins; kinship shapes bounded per link; nickname search; guard lists read from Go; focus after the Source type editor | Spike 12 S12-42, S12-49, S12-47, S12-48, S12-30 |
| Ship-log follow-ups: a typed Promote batch target; `PromoteFlow.item`'s header | Spike 13 S13-23 |
| Ship-log follow-ups: duplicate-term shortcut; hand-kept seeded term lists | [`term-management.md`](../../../ideas/term-management.md) items 5–6 |
| Ship-log follow-ups: impossible dates saved; upward card shadows; P3 colors | [`docs/dogfood/ux.md`](../../../dogfood/ux.md) |
| Dropped: hand-listed key lists and value-type literals, test-side date builder copies, a duplicated `period` loop, per-call allocations with no measured cost, an untested guard failure branch, a negative integer tolerance (code-only registry), an unwrapped profile error that fires only on a broken loader | Small and guarded by tests; not worth a PR (2026-10-29) |
