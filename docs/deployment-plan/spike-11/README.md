# Spike 11 — Promote Matching Refinement

## Status

**Scoping (2026-10-16).** Requirements are being written; there is no PR map yet. Spike 10 is still open.

This spike is entirely about the Promote flow's matching: making the algorithm graspable, cleanly layered and tunable, then building the test suite and calibration on top.

**Goal: one matching module the rest of the app borrows from.** Matching becomes powerful, robust and reusable, with one well-defined home for every comparison. Retiring the duplicate comparison, normalization and dedupe code around the app is part of this spike, not a follow-up: each PR that moves a layer also moves that layer's other callers onto it and deletes what it replaces. The inventory is [Reuse across the app](comparison-layers.md#reuse-across-the-app). The starting idea is [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md).

## Documents

| Doc | Role |
| --- | --- |
| [**Comparison layers**](comparison-layers.md) | Requirements for layers 1–6 (value, property, Subject, entity and segment comparison, graph alignment), the lifespan check, the rules every layer follows, and the reuse and retirement inventory |

## Spike-wide principle: sparse data first

Three generations back, records thin out. Often all there is is a name and a birth date. The algorithm has to work well on a small set of key information: get the **core properties** and the **core relationships** right, and everything else reinforces them.

- **Core:** a person's name and sex; an event's type and date; a place's name; and the links that carry most identity: subject of a birth, parent / child, spouse, place part of place.
- **Note:** a person's birth date is not a person property. It lives on the birth event, through a link. So the most common sparse case, "same name, same birth date", spans Subject comparison (the name) *and* graph alignment (one low-fan-out link to a matching birth). Both halves must be solid on their own.
- **Weights:** core properties and links carry most of the weight. Secondary properties get small agree weights. They raise a match, but a match never depends on them, and they never outweigh a core conflict.
- **Effort:** the core comparers (names, dates) and the single-link case get the most tests, the most labeler time and the first tuning.
- **Tests:** the test suite has a **sparse tier**: goldens and simulated cases with only a name and a birth date (or less), reported as their own scoreboard row. Sparsity is a simulator noise setting: drop properties. Invariant: adding an agreeing secondary property never lowers a score, and never turns a core conflict into a match.

## Plan notes (for the PR map)

- **First PR: make the algorithm graspable.** The one-page model with a diagram, the glossary, and the knob inventory, pulled together from the comparison-layers doc. Docs first; it lands before any code moves.
- **The test suite comes after the model is re-implemented** (scenario files, invariants, simulator, scoreboard, labeler, fitting, dogfood export from [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md)). Re-implementing will teach us things that change it. Each layer's PR still carries its own unit tests and invariants.
- **Registry starting numbers** are settled with the PR map.

## Decided so far

- **Matching is a stack of pure layers.** Each takes only what it needs; a rule that needs more moves up a layer.
- **Packages:** value compare (`core/valuecompare`) and property compare (`core/propertycompare`).
- **Layer 1 compares two values** of one type and returns a normalized result (relation + similarity) plus structured reasoning.
- **Layer 2 compares one property** between two Subjects, with only what a Subject can say about itself: negatives, neutral terms, cardinality, missing values. No support or standing.
- **Frequency weighting is descoped entirely,** for names, event types and toponyms. Property weights are plain registry numbers. Generic properties (event type) get a low weight; places get their strength from relationships.
- **The entity profile is strict:** only the reconciled conclusion. Real alternates (maiden names) belong in the conclusion. The profile builder is its own package, `core/subjectprofile`.
- **Layer 3 is Subject comparison:** one Subject against one Subject. It combines property comparisons and holds the ones that span properties.
- **Layer 4 is entity comparison against the reconciled conclusion,** not each member. The conclusion becomes a Subject-shaped profile (concluded values positive, denied values negative), then layer 3 runs. Support weighting is outsourced to the reconciler.
- **`Contains` scores below 1, and `Near` has a minimum similarity** per type, in layer 1.
- **Comparison layers make no workflow decisions.** Promote, the reconciler and others keep their own decision logic and compose the layers.
- **Layer 3 compares the properties both Subjects have,** driven by a per-type registry. It returns two axes: **similarity** (0–1, are they alike) and **evidence** (points, how strong). No knockout flag: a decisive conflict is a large registry penalty.
- **Similarity weighs a conflict by its penalty,** so a decisive conflict pulls similarity down as well as evidence.
- **Dedupe is in scope.** Everything in the reuse inventory (other consumers, small widening changes, duplicates to retire, the lighter-touch items and the bugs found in passing) is Spike 11 work, split across PRs as each layer lands.
- **Layer 5 is segment comparison** (primary – bridge – primary, one hop, pure) and **layer 6 is graph alignment** (the walk and every decision). Bridge cardinality is declared in the registry; a conflict through a one-cardinality bridge subtracts; segment evidence includes the neighbor's match.
- **Lifespan, for people only:** the profile builder derives a person's lifespan from direct neighbors' dates by registry rules, and Subject comparison treats an impossible overlap as a conflict.
- **Layer 6 procedure:** shortlist by registry keys → score and keep the top K → one queue for every candidate (no separate fallback; the answer doesn't depend on where the walk starts) → segment comparison only around accepted pairs → settle only dirty pairs. Memoizing across re-runs is the plan's last step, built only if a performance test shows the need.
