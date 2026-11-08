# Review — S14-99 Close Spike 14 as a decisions note

**Verdict:** Fits with changes — Ready with changes (every round-2 finding Fixed; no blocker open)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None.

## Writer's own calls

- About 15 small ship-log follow-ups are dropped with a reason (copy and code nits, a 2 ms reload, a decode already moved off the session), listed in the note's Leftovers.
- Retarget rule (D-7): a link to a brief becomes the plain brief id; a link to the map or a plan goes to the decisions note; code-span mentions in dated Spike 13 logs and landed plans are left as they are.
- `check-plan-docs.py --links` is added as a manual tool, not a `check-pr.sh` gate (D-9): the base has about 110 broken links.
- The note leaves out the four Spike 12 decisions Spike 14 built (one line points at Spike 12 instead).
- `archive-docs` is rewritten in place (Keep section, keep checklist line, note template now Decisions / Refused / Leftovers pointing at Spike 12's note), and `ship-pr` step 2 gains a post-delete link rerun (D-10). These change two skills to match how Spikes 10 – 12 were closed.
- The base-versus-branch comparison lives as one command block in `archive-docs`, not as a `--base` script mode (D-9).

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

## Round 2

### X-10 · blocker · Two S14 ids in the moved doc are not in the edit table (lines 24, 31)
**Answer:** Fixed — §6 Moved design doc gains rows 24 and 31; §4 line list corrected.

### X-11 · blocker · STEP-2's note link check fails before the doc is moved
**Answer:** Fixed — the move is STEP-2 and the note STEP-3; STEP-2's check expects only the note link missing, STEP-3 checks both files.

### X-12 · major · Weak or unrunnable checks
**Answer:** Fixed — T-10 greps one literal per change; T-1 uses a space, not `\t`; T-2 runs `rg --hidden`; T-5 greps a Key column given per Re-home row; T-3 runs unexcluded (it lints changed plans only); PR body cites T-1 – T-11.

### X-13 · minor · Choices left open
**Answer:** Fixed — effects README: "input index" becomes the link text, existing parenthesis kept; the three ideas intros are given as whole sentences; D-6 links "Spike 16 — Odds and ends" to `spike-16/`; moved doc line 3 and every skill change are final Markdown in fences; change 2 names its place (end of line 48).

### X-14 · minor · Current state counts
**Answer:** Fixed — §4 Inbound links names the 16 Retarget files plus the two in Ideas made stale.

### M-7 · blocker · STEP-10 builds a last-commit mechanism beside ship-pr step 2
**Answer:** Fixed — STEP-10 removed. D-11: `ship-pr` "Hand off" step 2 makes the last commit and removes the folder with this plan; change 7 adds the link rerun there. The implementer's last step is STEP-9's sweep with `spike-14/` excluded.

### M-8 · blocker · archive-docs additions contradict its Keep rule and template
**Answer:** Fixed — D-10: changes 1, 4 and 5 replace the keep checklist line, the Keep section (two cases: live contract, design reasoning a domain doc links) and the note template (Decisions, Refused, Leftovers; Match Spike 12); change 8 updates `archive/README.md:5`. No rule is bent: the skill is changed to match four closes of practice.

### M-9 · blocker · Moved doc keeps S14 ids (lines 24, 31)
**Answer:** Fixed — same as X-10, with the reviewer's wording.

### M-10 · blocker · The comparison is still a script each close rewrites
**Answer:** Fixed — option (b): the exact command block is change 6 in `archive-docs` "Retarget"; T-1 and `ship-pr` step 2 run it. It uses `pwd -P` so macOS's `/private/var` temp path strips correctly. A `--base REF` mode is an Out row.

### M-11 · blocker · T-10 cannot verify the skill changes
**Answer:** Fixed — T-10 greps one literal per change, each expected once.

Base fact corrected from round 2: the base has 90 broken links, not about 110 (D-9, Out rows).

## Refresh — after the Spike 13 close (2026-11-08)

The user chose to land S13-99 (#753) first. #753 deleted the Spike 13 folder, rewrote `archive-docs` § Retarget (a whole-folder grep and two link forms) and moved Spike 13 to Completed. The plan was refreshed against `s13-99-dogfood-close` @ `9f150572`, then reviewed again by fresh maintainability and executability reviewers. It rides #753 to `main`.

**Refresh edits.** The nine Retarget rows into Spike 13 files and the D-7 "Not edited" line were dropped. AC-2.3, which linted two deleted S13 plans, was dropped, and AC-2.4 became AC-2.3. §4 Inbound links, Index, Lint (41 broken links at the base, 13 outside `spike-14/`) and Close skill were re-measured. D-7 now follows `archive-docs` § Retarget's two forms. Index text, AC-1.3 and T-6 put Spike 14 above Spike 13 in Completed. Manual check 4 no longer opens a deleted plan. `ship-pr` "Hand off" became § 8 "Finish". Change 4 and change 8 name Spike 13's kept doc.

**Verdicts:** Maintainability: Fits with changes. Executability: Ready with changes. No blocker, so there is no second refresh round.

### M-1 · major · The audit retarget form is missing from `archive-docs`
**Answer:** Fixed. Change 9 adds the audit form to § Retarget, and T-10 greps `An app-health audit becomes plain text`.

### M-2 / X-6 · major · A 40-line floor that Spike 12 (37 lines) and Spike 13 (32) miss
**Answer:** Fixed. Change 5 now reads "~30–80 lines" and "Match the headings of" Spike 12. AC-1.4 and T-7 read 30 – 80.

### M-3 / X-5 · major · The comparison block makes a git worktree every close
**Answer:** Fixed. Change 6 uses `git archive "$base" | tar -x -C "$tmp"` and cleans up with `rm -rf "$tmp" "$tmp.base" "$tmp.branch"`. The untracked `scripts/__pycache__/` came from the writer's own probe and was removed, so the plan needs no line about it.

### M-4 / X-3 · minor · T-10 cites AC-2.4
**Answer:** Fixed. It now cites AC-2.3.

### M-5 · minor · D-7 leaves out the `#decisions` anchor
**Answer:** Fixed. D-7 says "the note's `#decisions`". The depictions row targets `…/archive/spike-14/README.md#decisions`, and manual check 4 matches.

### X-1 · major · #753's `structural-debt.md:37` row, which STEP-1 half builds, is not decided
**Answer:** Fixed. §6 Ideas made stale gains a row 37 rewrite: `--links` exists, but anchors and a `docs/**` gate are still open. §4 Stale ideas and T-11 cover it.

### X-2 · major · plan-docs fails after STEP-8 if this plan is in the branch diff
**Answer:** Fixed. D-12 states that this plan reaches `main` with #753, before the branch is cut, and the implementer does not edit it, so the plan-docs step never lints it.

### X-4 · minor · Change 7's insertion point
**Answer:** Fixed. The sentence is appended at the end of that line, after the `archive-docs` link's closing parenthesis and full stop, with a space before it.
