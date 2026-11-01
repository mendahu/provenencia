# Ship log — S14-15 Xcode synchronized folders

**Plan:** [`S14-15-xcode-synchronized-folders.md`](S14-15-xcode-synchronized-folders.md) · **Branch:** `s14-15-xcode-synchronized-folders` · **PR:** #668
**State:** ready

## Round 0 — implement
Head `382e3a00`. Gate PASS (orchestrator run against `60a7e122`). Report deviations: AC-3.3 plant 3 prints its two lines sorted per D-7 (AC lists them the other way); project-file commit gated before STEP-6–8, STEP-7 and STEP-8 in one commit; baseline from `60a7e122` (descendant of planned `067be77`, Current state unchanged); Xcode UI manual check not run (left to the user); CI box open until ready; conversion script kept in scratchpad (D-1).

## Round 1 — review
Reviewers: CON (Sonnet), MNT. Skipped: UX (the only `macos/App/**` file is `DesignSystem/README.md`, contributor docs; no strings, copy, or error codes); SEC+PRF (no listed security or performance path; the new guidance check walks the repo's own folders, not data the app did not write).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major (reviewer: blocker) | `membershipExceptions` read by a second list parser beside `_id_list` | Fix now → F-1 | verified by reading `check-guidance-names.py:150-152`: splits on `"\n"`, so a one-line list is one entry. Re-rated to major: Xcode writes one entry per line today, so the gate does not fail on the real tree; still a defect in code this PR adds |
| MNT-4 | CON (routed from MNT "Outside my lens") | minor | Exceptions from every set on a root are merged, whatever the set's `target` | Fix now → F-2 | verified by reading `:149-152`: `target` is never read. No root is in two targets today, so no wrong result yet; a minor in a file this PR adds |
| MNT-2 | MNT | minor | Script summary line does not mention the synchronized-folder check | Fix now → F-3 (summary line); split into its own script → Follow-up | docstring line 1 unchanged in the diff |

Nits (logged): `CON-1 · .agents/skills/add-swift-test/SKILL.md:6 · rewrap the long description line` · `CON-2 · Xcode UI manual check (plan §10) is the user's, CI on format 77 pending` · `MNT-3 · scripts/check-guidance-names.py:144 · the parent-path lookup climbs one group level only; name it or comment it`

### Fix plan 1

**F-1 — One pbxproj list reader** (from MNT-1)
Files: `scripts/check-guidance-names.py`.
Failing test first: none (the script has no test file); prove it with a plant instead. Temporarily rewrite the `Features` exception set in a scratch copy of the project as one line, `membershipExceptions = (GraphCanvas/README.md, );`, run the check against it (or a copy of the repo), and show it reports a false problem before the fix and OK after. Restore the project file; do not commit the plant.
Do: add `_list(body, key)` returning the stripped, unquoted entries of `key = ( … );` (`re.S`, split on commas, drop `/* … */` comments and empty entries). Make `_id_list` filter `_list` for 24-hex IDs. Replace the inline `membershipExceptions` regex-and-split with `_list(sets[sid], "membershipExceptions")`.
Check: `python3 scripts/check-guidance-names.py` prints OK on the real tree; the AC-3.3 plants still fail with their D-7 lines; `scripts/check-pr.sh` PASS.

**F-2 — Count only exception sets for a target that builds the root** (from MNT-4)
Files: `scripts/check-guidance-names.py`.
Do: read the `PBXNativeTarget` objects; for each root, the targets whose `fileSystemSynchronizedGroups` list it. Take `membershipExceptions` only from sets whose `_field(set, "target")` is one of those targets. Update the docstring paragraph to say so.
Check: a plant that changes the `Features` set's `target` to the test target `AA00000000000000000000B6` (scratch copy, not committed) makes the check report `macos/App/Features/GraphCanvas/README.md: not Swift and not in a membership exception …`; the real tree prints OK; `scripts/check-pr.sh` PASS.

**F-3 — Name the new check in the script's summary line** (from MNT-2)
Files: `scripts/check-guidance-names.py`.
Do: extend the docstring's first line so it also names the Xcode synchronized-folder check (non-Swift files outside a membership exception, and the reverse).
Check: `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `7166ef6d`. Gate PASS (orchestrator run). F-1, F-2, F-3 done in one commit. Deviations: the single-entry F-1 plant did not reproduce on the old code (it stripped the trailing comma); a two-entry one-line list did, and passes after the fix. `_field` now tolerates a trailing `/* name */` comment, which F-2 needs to read `target`.
F-2 (MNT-4, minor): **Resolved**, by reading the diff: `builders` maps each root to the targets whose `fileSystemSynchronizedGroups` list it; a set counts only when its `target` is one of them.
F-3 (MNT-2, minor): **Resolved**: the docstring's first line names the synchronized-folder check.

## Round 2 — review
Reviewers: MNT (had the Fix-now major MNT-1). CON had no Fix-now blocker or major.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) second list parser | Resolved | `_list` is the only list reader; `_id_list` filters it; probe of one-line, multi-line and comma-in-comment lists returns the right entries; real tree OK |

No new blockers. Nits (logged): `MNT-5 · scripts/check-guidance-names.py:29 · re-wrap the docstring line F-2 lengthened`

## Result
Gate PASS · CI pending at ready · open: none · follow-ups: MNT-2 (split the disk-vs-declared-list checks out of `check-guidance-names.py` if a third non-guidance check arrives) · to user: — · CON miss: none
