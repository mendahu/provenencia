# Review — S14-99 Close Spike 14 as a decisions note

**Verdict:** round 1: Fits with changes — Not ready; round 2 pending
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None.

## Writer's own calls

- About 15 small ship-log follow-ups are dropped with a reason (copy and code nits, a 2 ms reload, a decode already moved off the session), listed in the note's Leftovers.
- Retarget rule (D-7): a link to a brief becomes the plain brief id; a link to the map or a plan goes to the decisions note; code-span mentions in dated Spike 13 logs and landed plans are left as they are.
- `check-plan-docs.py --links` is added as a manual tool, not a `check-pr.sh` gate (D-9): the base has about 110 broken links.
- The note leaves out the four Spike 12 decisions Spike 14 built (one line points at Spike 12 instead).

## Round 1

### X-1 · blocker · Retarget table misses inbound references
S13-D1:14 and S13-D4:19 link deleted briefs; code-span hits in S13-10:169, S13-12 ship log, S13-25 review log, S13-31 plan; file count wrong.
**Answer:** Fixed — §6 Retarget gains S13-D1:14 and S13-D4:19 (already broken at the base, so the planning diff missed them). D-7 decides code spans in dated logs and landed plans are not edited, and T-2 matches Markdown link targets only. Counts corrected in §4 (15 files).

### X-2 · blocker · STEP-7 sweep and AC-1.2 cannot pass before the last commit
**Answer:** Fixed — STEP-9 runs T-1 and T-2 with `docs/deployment-plan/spike-14/` excluded; STEP-10 reruns T-1, T-2, T-4 unexcluded after deleting the plan. AC-1.1, AC-1.2, AC-2.1, AC-2.2 are checked after STEP-10.

### X-3 · major · T-1 leaves every choice to the implementer
**Answer:** Fixed — with M-3: `check-plan-docs.py --links` reuses `check_links` (fenced code skipped as for plans); §6 Lint script gives the exact base-worktree comparison, `file<TAB>target` without line numbers.

### X-4 · major · check-pr.sh cannot pass off macOS
**Answer:** Fixed — D-12: PASS on macOS, or every step but xcode-test and swift-warnings plus CI's macOS job green.

### X-5 · major · STEP-3 leaves wording and placement open
**Answer:** Fixed — §6 Leftovers gives the literal text, form (bullet, section, row, append) and place of each item, including the `share-packages.md` row's Item and Was cells.

### X-6 · minor · STEP-5 placement and wording
**Answer:** Fixed — §6 Index text: Current keeps Spike 13 and 15 as they are, Spike 16 line after Spike 15; literal Completed row.

### X-7 · minor · AC-3.2 / T-8 not checkable
**Answer:** Fixed — §6 "Drop phrases" lists the eleven phrases T-8 greps.

### X-8 · minor · Current state rows not true
**Answer:** Fixed — §4: 35 plans; briefs S14-D1 – S14-D9 plus one review log; 15 inbound files; audits also named in two Spike 13 logs.

### X-9 · minor · Smaller choices
**Answer:** Fixed — the note's date is the day the file is written (D-2); STEP-3's edits are an exact table; the review and ship logs are deleted in STEP-10 after they exist. The branch `s14-99-close-spike-14` is the implementer's; the planning branch is separate.

### M-1 · major · The kept summary-inputs.md has no reader once the folder goes
**Answer:** Fixed — §6 Retarget links it from `core/database/effects/README.md:15`, `docs/catalog-deletes.md:46` and `add-searchable-kind` line 31; the `structural-debt.md:35` source cell links it (Ideas made stale). Line 3 takes Spike 12's "those docs and the code win" wording. `.cursor/rules/catalog-queries.mdc` is left out (Out row: its paired skill gets the link).

### M-2 · major · The close pattern exists only in a plan that deletes itself
**Answer:** Fixed — STEP-7 and §6 "archive-docs additions" (audits and `ux/` in the delete list, the S#-99 plan carrying the note and triage, the link-check command, kept files linked from their teaching doc). AC-1.4 now matches the skill's 40 – 80 lines.

### M-3 · major · T-1 re-writes a link checker the repo has
**Answer:** Fixed — STEP-1 adds `--links` to `check-plan-docs.py` (D-9). Not a `check-pr.sh` gate, with evidence: about 110 broken links at the base.

### M-4 · major · Stale ideas lines
**Answer:** Fixed — §6 "Ideas made stale": rows 18 and 25 of `structural-debt.md` deleted, rows 35 and 36 present tense, `provisional-identity-claims.md:34,51` and `depictions-and-likenesses.md:3` reworded, three `docs/ideas/README.md` summaries extended.

### M-5 · minor · The moved doc cites deleted plan ids
**Answer:** Fixed — §6 Moved design doc removes every S14, T- and D- id and keeps the numbers.

### M-6 · minor · The note repeats Spike 12's decisions
**Answer:** Fixed — three decisions and one Refused item removed; one line in the note's opening points at Spike 12.
