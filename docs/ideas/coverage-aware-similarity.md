# Similarity that weighs how much was compared

**Status:** idea only — not roadmapped. Found in S11-16 (STEP-3 stop, the user's call of 2026-10-19).

## What happens

Graph alignment gates candidates on similarity ≥ `alignment.MinSimilarity` (0.4). Similarity counts only the properties both sides have. A handle that agrees on one property and has nothing else to compare scores 1, so it becomes a candidate. Its evidence is under the weak band, so the row has no target, with the handle as an unbanded alternative. Nothing is preselected.

In Promote the real case is a surname-only record against a full name. The shortlist's name word keys match, so every surname-only record with that surname reaches `Candidates` and ties at 1, up to `alignment.TieCap` (20). The row then shows up to `alignment.Alternatives` (3) entries with nothing behind them but the surname. Raising `MinSimilarity` does not help, because the tie is at 1.

The same rule makes an undated event against an undated event of the same type a candidate in `graphalign`. Promote does not reach this today: `shortlist.Keys` gives an event with no date no keys. It would appear if a later shortlist change keyed undated events.

## Suggested shape

- Similarity, or the candidate gate, also weighs coverage: how many counted properties beyond the kind-defining one (event type, a bare surname) were compared.
- The coverage minimum is a registry knob in `alignment`, with an estimate and an inventory row.

## Open questions

- Which properties count as kind-defining, per kind.
- Whether dogfood shows the noise at all. Check first how often Promote rows show unbanded surname-only alternatives.
