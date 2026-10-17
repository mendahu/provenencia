# Review — S10-99 Close Spike 10

**Verdict:** round 1: Fits with changes — Ready with changes (all applied)
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None. The user answered the two product calls before review (2026-10-17): the dogfood bar is checked, and the `likeness` Property on existing projects is handled outside the repo.

## Writer's own calls
- No Spike 10 file is kept as a live contract; the three design notes are deleted, not moved to `ideas/archive/` (D-1, D-2).
- The plan, its review log, and its ship log are deleted in the branch's last commit; no Landed note or `completed.md` paragraph (D-5, D-6). Written into `archive-docs` and `implement-pr` so later closes inherit it (S-1, S-2).
- The two landed Spike 11 plans that name `docs/deployment-plan/spike-10/**` in an `rg --glob` stay (D-7).
- The archive README text is fixed in the plan (D-3).

## Round 1

### X-1 · blocker · STEP-5's gate fails on `plan-docs` after STEP-4
The plan links `../deployment-plan.md`, which STEP-4 deletes; `check-plan-docs.py` runs on this changed plan and reports a broken link.
**Answer:** Fixed — D-10: the gate runs at STEP-3 (before deletion) and at STEP-6 (plan gone, `plan-docs` has nothing to check); STEP-5 no longer runs it.

### X-2 · major · T-1 and T-6 do not run as written (`\|` table escape)
**Answer:** Fixed — §8 commands are in a fenced `sh` block with plain `|`. Checked: T-1 prints the 10 expected pre-edit lines.

### X-3 · major · No branch or commit sequence
**Answer:** Fixed — §7: branch from `main`, four named commits, STEP-6's subject.

### X-4 · minor · STEP-6 `git rm` fails if a log was never created; who writes the last commit
**Answer:** Fixed — STEP-6 removes whatever remains; subjects for the `ship-pr` and no-`ship-pr` cases.

### X-5 · minor · STEP-2 check does not prove R-4 to R-8
**Answer:** Fixed — STEP-2 check adds T-6, T-7, and per-file counts.

### X-6 · minor · R-9 indent
**Answer:** Fixed — OLD-A / OLD-B lines, no indent.

### X-7 · nit · T-1 misses `./spike-10/` links
**Answer:** Fixed — T-1 pattern covers `(./spike-10` and `(spike-10`.

### M-1 · major · R-1 puts a shipped feature under matching.md "Not yet handled"
**Answer:** Fixed — R-1 drops Abbreviations from that list (lines 93–94 are its contract) and the archive link.

### M-2 · major · README copies registry values and links into Spike 11's folder
**Answer:** Fixed — the 0.95 / 0.98 values and the frequency sentence are gone; `matching.md` is in Live models.

### M-3 · major · The close-PR pattern lives only in this plan
**Answer:** Fixed — S-1 (`archive-docs` checklist line) and S-2 (`implement-pr` step 8 sentence) are in Scope and STEP-3. D-10's off-Mac exception is dropped; `ship-pr`'s Mac rule stands.

### M-4 · major · R-4 leaves a stale status line
**Answer:** Fixed — R-4 replaces the whole Status line, like R-5.

### M-5 · minor · Refused lists parked items
**Answer:** Fixed — Refused keeps only what will not be done; parked items are once, under Leftovers.

### M-6 · minor · "Matcher leftovers" row points at a note with no leftovers
**Answer:** Fixed — rows now name the real homes: Spike 11, `matching-test-suite.md`, `identity-claim-review.md`.

### M-7 · minor · Leftover labels enumerate items and miss some
**Answer:** Fixed — rows are labelled by home ("Depiction leftovers", "Dogfood notes still open").

### M-8 · minor · Present-tense "Spike 10" pointers
**Answer:** Fixed — R-11 to R-13, as the Spike 9 close did; T-7 proves it.

### M-9 · nit · Live models misses `structured-name-model.md`
**Answer:** Fixed.
