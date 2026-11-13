# Promote matching: ranking, performance and quality

**Status:** parked 2026-10-29 when Spike 12 closed; "Name index on commit" and "Narrower canon walk" moved to Spike 16 Batch 2 on 2026-11-08 ([S16-20, S16-22](../deployment-plan/spike-16/cache-architecture.md)); consolidated 2026-11-04 from this file plus `name-index-on-commit.md`, `narrower-canon-walk.md` and `coverage-aware-similarity.md`. Promote's suggestions are still ranked by graph alignment's score, unseen since Spike 13's S13-23 ([`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) §5.3). This note collects the matching work that waits until that ranking is rebuilt, likely on the Evidence graph scoring of [Spike 13's requirements](../deployment-plan/archive/spike-13/evidence-graph-comparison.md) once it has proven itself.

Audit items below are from the second app-health audit (2026-10-28, at `3cc609c`; deleted when Spike 14 closed, in git history). Budgets were to be measured at Spike 12's Scale: 37,000 canonical records, 100 member Sources on one record, 100 new people on one Source, 6,000 events on one place.

## The direction

Rank a Source Subject's candidate records with the same scoring the handle pages use to show how records connect, instead of a second number tuned only for proposing. Until then, graph alignment keeps ranking and nothing extends it.

## Inventory (audit leftovers)

| Item | Was | Problem |
| --- | --- | --- |
| B-2 | S12-41 | Setting a Source person to New or Skip, or another row taking a record, can make a one-cardinality conflict (their birth) vanish from the neighbors' scores. |
| B-11 | S12-51 | Each proposal compares would-be-new rows against each other more than once (2.29 s per decision at 200 new people): the alike pairs within one Evidence graph could be kept on the graph's memo. Its other half, the whole-canon walk, went to Spike 16 S16-22. |
| B-19, B-20, B-22 | S12-46 | Names compare differently depending on how they are written: initials and aliases on folded tokens, NFC and compatibility forms, format characters. Also mirrored in Swift's `TextMatch`. |

## Memoise a person's composed kinship

From S16-16 (user, 2026-11-13). S16-16 removed the stored per-person implied links and siblings answers: a person's implied links and siblings are now composed from the pivot memos on every read. On a warm proposal (the Promote page opening again, and every re-proposal after a decision) that composition runs for every person the canon walk reaches. Measured on an M1 under load (S16-16 ship log has both tables): `BenchmarkLargeSourcePropose` warm cost 1.56–1.74× the base at 4,000 persons and 1.68–2.27× at 16,000, run to run. The proposal after a handle commit is unchanged within noise (best of 3 and 5, alternating base and head, at 4,000 persons: 625 ms on the base, 630 ms at head). The user accepted the warm cost as the new normal.

**Pick it up when** researchers say the pause after a Promote decision is too long on a large tree. Measure first: S16-22 reads each handle's links from the pivots instead of walking the component, which may shrink the walked set enough to settle it.

**Shape.** Memoise `canonlinks.Reader.For` per person in a keyed registry entry whose Apply drops only the people whose pivots changed and the people one parent step from them, without a catch-all drop on any handle commit (the working rule "Every cache names the signal that updates it, and updates exactly"). Measure memory at 16,000 against `docs/catalog-caches.md` first (rule "No stored answer without a measurement").

## Coverage-aware similarity

Was its own idea (2026-10-19, from S11-16 STEP-3). Quality of the candidate gate, not a rebuild of the ranker — still waits with the matching work above.

### What happens

Graph alignment gates candidates on similarity ≥ `alignment.MinSimilarity` (0.4). Similarity counts only the properties both sides have. A handle that agrees on one property and has nothing else to compare scores 1, so it becomes a candidate. Its evidence is under the weak band, so the row has no target, with the handle as an unbanded alternative. Nothing is preselected.

In Promote the real case is a surname-only record against a full name. The shortlist's name word keys match, so every surname-only record with that surname reaches `Candidates` and ties at 1, up to `alignment.TieCap` (20). The row then shows up to `alignment.Alternatives` (3) entries with nothing behind them but the surname. Raising `MinSimilarity` does not help, because the tie is at 1.

The same rule makes an undated event against an undated event of the same type a candidate in `graphalign`. Promote does not reach this today: `shortlist.Keys` gives an event with no date no keys. It would appear if a later shortlist change keyed undated events.

### Shape

- Similarity, or the candidate gate, also weighs coverage: how many counted properties beyond the kind-defining one (event type, a bare surname) were compared.
- The coverage minimum is a registry knob in `alignment`, with an estimate and an inventory row.

### Open questions

- Which properties count as kind-defining, per kind.
- Whether dogfood shows the noise at all. Check first how often Promote rows show unbanded surname-only alternatives.

## Related, kept separate

- [`matching-test-suite.md`](matching-test-suite.md) — scenario goldens, invariants, labeler; tooling for changing the algorithm with confidence, not a Promote performance fix. Also holds "remember researcher decisions."
- [`promote-graph-alignment.md`](../promote-graph-alignment.md) — what proves a join; the live model.

## A value limit under every property compare

From the S13-26 ship loop (#736), the user's idea, 2026-11-07.

### Problem

`core/recordlinks` stops a crafted catalog with `recordlinks.MaxUnitValues` and `MaxValueBytes`, but those limits sit in that one package. Subject compare runs the same values × values loop through `propertycompare.Compare`, on names with no limit on how many values a Subject holds or how long each one is. Promote keeps that exposure. S13-26's probes measured the record-links read, not Promote: 20 names of 12 words on each side took 1.54 s for one member pair.

### Direction

Move the limit down into `propertycompare.Compare`, so every caller gets it. Over the limit, the property reads as unknown, the way text compare answers `CauseTooLong`. Do not compare only the first few values: which values come first would decide the result, and the matching value may be the one left out, so Promote could see a false conflict.

Stopping early when values match does not help: matches are already the cheap path, and the slow input is many similar values that never match.

### Before picking it up

- The move changes Promote's code path, so it needs a Promote regression check. The Subject compare and graph alignment suites are where to start.
- Measure Promote at the map's Scale first: the numbers above are from the record-links read.
- `recordlinks` would then read the shared limit rather than its own knobs, and the two inventory rows would move.
