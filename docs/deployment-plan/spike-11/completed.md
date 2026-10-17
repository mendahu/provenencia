# Spike 11 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S11-01 — One-page matching model

**Landed 2026-10-17** in [#432](https://github.com/mendahu/provenencia/pull/432). The matching algorithm now has one page to read. [`matching-model.md`](../../matching-model.md) draws the six layers in one diagram, defines the terms (relation, similarity, evidence, segment, shortlist, candidate, band, contested), walks one Promote row from value compare to band with real numbers, and lists every registry knob with its section, starting value or estimate, and what it changes. The model also fixes lifespan under `subject` (`subject.person.lifespan.*`) and keys the sex rows on `sex_at_birth`, which covers `gender` through the pair. `matching.md` and `promote-graph-alignment.md` carry banners saying which sections Spike 11 replaces and that they describe the running code until then. Docs only; no code changed.
