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
- **Layer 1 compares two values** of one type and returns a normalized result (relation + similarity) plus structured reasoning.
- **Layer 2 compares one property** across both sides' value sets, with their context (negatives, cardinality, reconciler standing, neutral terms).
- **Frequency weighting is not a comparison.** It moves above layer 2.
