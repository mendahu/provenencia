# Spike 11 — Promote Matching Refinement

## Status

**Scoping (2026-10-16).** Requirements are being written; there is no PR map yet. Spike 10 is still open.

This spike is entirely about the Promote flow's matching: making the algorithm graspable, cleanly layered and tunable, then building the test suite and calibration on top. The starting idea is [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md).

## Documents

| Doc | Role |
| --- | --- |
| [**Comparison layers**](comparison-layers.md) | Requirements for layers 1 (value comparison) and 2 (property comparison), and the rules every layer follows |

## Decided so far

- **Matching is a stack of pure layers.** Each takes only what it needs; a rule that needs more moves up a layer.
- **Packages:** value compare (`core/valuecompare`) and property compare (`core/propertycompare`).
- **Layer 1 compares two values** of one type and returns a normalized result (relation + similarity) plus structured reasoning.
- **Layer 2 compares one property** between two Subjects, with only what a Subject can say about itself: negatives, neutral terms, cardinality, missing values. No support or standing.
- **Frequency weighting is descoped entirely,** for names, event types and toponyms. Property weights are plain registry numbers. Generic properties (event type) get a low weight; places get their strength from relationships.
- **The entity profile is strict:** only the reconciled conclusion. Real alternates (maiden names) belong in the conclusion. The profile builder is its own package, subject profile builder.
- **Layer 3 is Subject comparison:** one Subject against one Subject. It combines property comparisons and holds the ones that span properties.
- **Layer 4 is entity comparison against the reconciled conclusion,** not each member. The conclusion becomes a Subject-shaped profile (concluded values positive, denied values negative), then layer 3 runs. Support weighting is outsourced to the reconciler. Graph alignment is layer 5.
- **`Contains` scores below 1, and `Near` has a minimum similarity** per type, in layer 1.
- **Comparison layers make no workflow decisions.** Promote, the reconciler and others keep their own decision logic and compose the layers.
- **Layer 3 (draft):** compares the properties both Subjects have, driven by a per-type registry; mainly a knockout layer. Proposed result: verdict + agreement (0–1) + evidence (points).
