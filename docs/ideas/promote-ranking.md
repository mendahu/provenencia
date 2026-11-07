# Promote matching: ranking, performance and quality

**Status:** parked 2026-10-29 when Spike 12 closed; consolidated 2026-11-04 from this file plus `name-index-on-commit.md`, `narrower-canon-walk.md` and `coverage-aware-similarity.md`. Promote's suggestions are still ranked by graph alignment's score, unseen since Spike 13's S13-23 ([`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) §5.3). This note collects the matching work that waits until that ranking is rebuilt, likely on the Evidence graph scoring of [Spike 13's requirements](../deployment-plan/spike-13/evidence-graph-comparison.md) once it has proven itself.

Audit items below are from the second app-health audit, now at [`deployment-plan/spike-14/app-health-audit-2.md`](../deployment-plan/spike-14/app-health-audit-2.md). Budgets were to be measured at Spike 12's Scale: 37,000 canonical records, 100 member Sources on one record, 100 new people on one Source, 6,000 events on one place.

## The direction

Rank a Source Subject's candidate records with the same scoring the handle pages use to show how records connect, instead of a second number tuned only for proposing. Until then, graph alignment keeps ranking and nothing extends it.

## Inventory (audit leftovers)

| Item | Was | Problem |
| --- | --- | --- |
| B-2 | S12-41 | Setting a Source person to New or Skip, or another row taking a record, can make a one-cardinality conflict (their birth) vanish from the neighbors' scores. |
| B-8, B-9 | S12-43 (performance half) | A handle commit throws away the name index and every Source's ranked candidates, and profiles are rebuilt; see [Name index on commit](#name-index-on-commit). Also: the graph cache's three warm slots (alias index, name format, match index) should be one keyed slot, and its summary display rows should be memo slots. |
| B-8, B-11 | S12-51 | Each proposal walks the whole connected canon and compares would-be-new rows against each other more than once; see [Narrower canon walk](#narrower-canon-walk). |
| B-19, B-20, B-22 | S12-46 | Names compare differently depending on how they are written: initials and aliases on folded tokens, NFC and compatibility forms, format characters. Also mirrored in Swift's `TextMatch`. |

## Name index on commit

Was its own idea (2026-10-26); scheduled in Spike 12 (S12-43 / S12-51) then parked with this note. From the S11-39 ship loop (#567) and the slice 9 re-cut.

### Problem

Promote's name index (`promotealign`'s shortlist, built over every person, event and place) and every Source's ranked candidates are thrown away by any handle commit (`graphcache.OnCommit` clears the match index; the candidates memo is keyed on its generation). The next proposal rebuilds the index (about 70 ms on an 18,000-handle catalog) and ranks every Subject again (about 200 ms). On that catalog a proposal after a handle commit took about 0.54 s locally and 1.75 s on CI against S11-21's 1.5 s budget.

S11-50 moves the first build behind the open spinner. It does not help after a commit.

A filed connect (S14-33) is now a handle commit too. At 18,233 handles the first same-graph proposal after one is about 0.78 s, against 0.57 s after an unfiled connect; at 36,977 handles about 1.35 s, against 0.95 s. Skipping the drops for bridges is not safe: a participation changes a Person's profile, and a relationship changes `canonlinks`. A location and a place_relationship could skip both. The user accepted the cost on 2026-11-06.

A filed location (S14-33) also loads the place's whole link list twice in one write (`graphcache.loadIDs`, from `OnCommit`'s reload and `composeEvents`): about 72 ms at 6,000 links, 306 ms at 24,000. A possible fix is to patch the loaded node with the one new link, which removes both loads. The user accepted the cost on 2026-11-06.

S11-46 (#602) widened the rebuild: prefix keys file about 46% more index keys. On S11-21's large fixture a proposal after a handle commit went from about 424 to 512 ms. On a 20,000-person catalog with short shared given-name stems ("Jo", "Al", "Ed", "Ann"), a proposal after a handle commit took about 1.94 s with prefix keys and 1.82 s without, both over the 1.5 s budget. The user kept prefix keys uncapped (2026-10-28); this is the fix for that cost.

### Shape

- The commit updates the index in place: remove the committed handles' old keys, add their new ones, keep the generation stable for handles it did not touch.
- The candidates memo re-ranks only the Subjects whose shortlist keys meet a changed handle's old or new keys, and keeps the rest.
- Graph verify mode compares the kept index with a fresh build, as it does for nodes.

### Open questions

- Whether the shortlist index (`core/shortlist`) can remove a handle cheaply, or needs per-handle key lists.
- How a merge (two handles become one) reads: remove both, add the survivor. S14-19's `OnCommit` `merged` list is the input that shape would read.

## Narrower canon walk

Was its own idea (2026-10-26); scheduled in Spike 12 (S12-43 / S12-51) then parked with this note. From the S11-39 CI fix round, where the user raised the warm budget from 250 ms to 350 ms.

### Problem

Every proposal walks the whole connected component of the canon from its seeds. On the large fixture that is about 18,000 handles and 24,000 links, plus about 6,000 implied links, so per-link work grows with the tree, not with the Source. The warm proposal on a 4,000-person connected tree costs about 115 ms locally and 297 ms on CI (S11-39).

### Shape

- Walk only as far as alignment can use: the hops a layer link can correspond to from the seeds and their placed neighbors.
- Keep the walk deterministic.
- Measure against `TestLargeSourceBudgets`.

### Memoize the canon (2026-11-03)

A profile of warm proposals alone (`TestLargeSourceBudgets/warm`, an allocs profile before and after ten proposals) shows each one allocating about 122 MB in about 700,000 allocations with no query: `promotealign.loadCanon` is 69 % of it (about 85 MB a proposal), `kinship.(*LinkSet).Add` 26 %, `graphalign.newState` 18 %. A one-to-two-collection run over a 170 MB heap is what flaked the 350 ms budget on CI (388 ms, best of three). The layer and the candidates are already memoized on the Evidence graph (`evidencelayer.Load`, `loadCandidatesMemo`); the canon is the one large piece rebuilt every time. A canon memo beside them, keyed by the fixed anchors and dropped with the handle graph, would remove most of the churn; a narrower walk would shrink what is memoized. The test now collects before each timed run and takes the best of five, so CI holds while this waits.

### Open questions

- How far a grandparent or place chain needs to reach.
- Whether the 250 ms warm figure can return.

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
