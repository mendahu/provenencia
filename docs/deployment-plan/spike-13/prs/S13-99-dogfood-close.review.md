# Review — S13-99 Dogfood close and docs

**Verdict:** pending round 2
**Rounds:** 1 so far. **Reviewers:** maintainability (inherit), executability (sonnet)

## To user

- M-1: Event and Place leftovers. Answered 2026-11-08: the pages are correct as built and are not a leftover; nothing is recorded for them (D-4).
- M-2: name the dropped follow-ups. Answered 2026-11-08: "Name them, drop flakes" (D-3).
- M-7 and M-5: answered 2026-11-08: fix line 332, and log the all-docs link check as a `structural-debt.md` row (STEP-4).

## Writer's own calls

- Keep `evidence-graph-comparison.md` in the archive; delete `person-detail.md` and `provenance-sheet.md` (D-2).
- A deleted plan, brief or ship log becomes plain text with the PR number; a deleted map or feature doc becomes a link to the archive README's Decisions (D-5).
- Record-link scoring questions are homed in `conclusion-layer-data-model.md` §13 item 4 and `promote-ranking.md`, the existing homes (D-4).
- The decisions note's floor is 30 lines, not 40; the Spike 12 note is 37 (X-1).

## Round 1

### X-1 · blocker · The decisions note is 32 lines; AC-1.1, T-3 and STEP-1 want 40 – 80
**Answer:** Fixed — the bound is 30 – 80 in AC-1.1, T-3 and STEP-1. The `archive-docs` target is "~40–80" and the Spike 12 note is 37.

### X-2 · blocker · STEP-6 removes a review log and a ship log that do not exist yet
**Answer:** Fixed — STEP-6 says the review log exists from planning, `ship-pr` writes the ship log, and a missing ship log is skipped.

### X-3 · major · `check-pr.sh` cannot pass without Xcode
**Answer:** Declined in round 1 on wrong evidence; reopened as X-10 in round 2 and fixed there.

### X-4 · major · L-3 and L-4 point at a README that may not state the rule
**Answer:** Fixed — the exact replacements are in §4, and they link `README.md#decisions`. The Decisions bullet "Family is grouped in Go and shaped as a family tree's first step" states the engine rule. Stop condition 4 now covers only links outside L-1 – L-13.

### X-5 · major · Replacement wording leaves choices (L-11 – L-13, kept doc line 30)
**Answer:** Fixed — §4 gives an exact old → new string pair for each of L-1 – L-13, and §6 **Kept doc** gives line 30's full before and after.

### X-6 · minor · Sweep commands disagree
**Answer:** Fixed — one sweep (T-2), with targets judged where they resolve; the DoD reuses it and adds the code and skills `rg`.

### X-7 · minor · Prose ids of deleted files
**Answer:** Fixed — an Out row keeps them as written (`archive-docs` § Retarget).

### X-8 · minor · Decisions note facts are only partly checked
**Answer:** Fixed — STEP-1 checks more names, `deleteimpact` and the dropped pin table. The note says "the match registry's `RecordLinks.StrongAt`".

### X-9 · major · README line 11 also says Spike 13
**Answer:** Fixed — STEP-5, T-5 and AC-2.3 expect line 11 to stay.

### M-1 · blocker · Open leftovers with no live home
**Answer:** To user → answered (2026-11-08) — the Event and Place pages are not a leftover. The scoring questions point at conclusion §13 item 4, and Promote ranking at `promote-ranking.md`. D-4 is rewritten. No rule bends: `archive-docs` homes open leftovers, and the user says these are not open.

### M-2 · major · Dropped follow-ups not named
**Answer:** To user → answered (2026-11-08) — the Leftovers row names every dropped item, flaky suites included. D-3 is rewritten.

### M-3 · minor · D-5 misstates the precedent
**Answer:** Fixed — D-5 is two cases now, and it says it departs from Spike 11. L-8 and L-9 gain `#decisions`.

### M-4 · minor · R-10's room for Reconciliation Claims would be lost
**Answer:** Fixed — the provenance-sheet bullet in the decisions note carries it, with a link to conclusion §13.

### M-5 · minor · A third hand-rolled link sweep
**Answer:** Fixed (user, 2026-11-08) — logged as a `structural-debt.md` row (§6 **Debt row**, STEP-4). Building it is Out.

### M-6 · nit · Brief ids in the decisions note
**Answer:** Fixed — "board S13-D6 rev 2" is gone from the note and D-1.

### M-7 · nit · Line 332 of the conclusion doc
**Answer:** Fixed (user, 2026-11-08) — STEP-4 and AC-2.4.

## Round 2

### X-10 · blocker · `check-pr.sh` fails off macOS by design, so the X-3 answer was false
**Answer:** Fixed — the finding is right (`scripts/check-pr.sh` lines 123 – 127 add `xcode-test (needs macOS)` to the failures). STEP-6 and the DoD now want PASS on a Mac; off macOS the implementer reports the FAIL line with every other step passing, and the gate runs on a Mac before the PR leaves draft.

### X-11 · major · L-1's old string starts with `(`, not `[`
**Answer:** Fixed — L-1's pair starts `[Spike 13's design](`.

### X-12 · minor · `check-plan-docs.py` has two errors before STEP-3
**Answer:** Fixed — the stop condition applies after STEP-3 and names the two errors it deletes.

### X-13 · minor · The kept-doc matching-model link has no line number
**Answer:** Fixed — line 19.

### X-14 · minor · No commit is named when there is no ship log
**Answer:** Fixed — the other two files go in commit 2 all the same.
