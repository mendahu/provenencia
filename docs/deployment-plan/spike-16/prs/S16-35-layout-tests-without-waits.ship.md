# Ship log — S16-35 Layout tests compute their frames instead of waiting on a window

**Plan:** [`S16-35-layout-tests-without-waits.md`](S16-35-layout-tests-without-waits.md) · **Branch:** `plan/s16-35-layout-tests-without-waits` · **PR:** [#775](https://github.com/mendahu/provenencia/pull/775)
**State:** review 1

## Before round 0

- The user asked to ship S16-35 on the planning PR (#775) rather than a new branch (2026-11-10). The plan's **Branch** line now names `plan/s16-35-layout-tests-without-waits`; no Decision, Scope or AC changed.
- Branch base: `main` @ `db5885df` (the plan's planning base); `origin/main` has not moved.

## Round 0 — implement
Head `d7bbf6cb` (code head `68b4c935`). Gate PASS (orchestrator run against `db5885df`, load 7.4 / 11.9 / 10.9). Report deviations: T-5 is one test looping over two notices; four code commits rather than one per STEP (STEP-1 to STEP-3 share a file); the before runs used the branch head (plan docs only on top of `db5885df`); manual check 3 not run (no UI).

D-8 run tables (implementer, one `xcodebuild` at a time; load = 1/5/15-min before each run):

| When | Run | Result | Duration | Load | Failing tests |
| --- | --- | --- | --- | --- | --- |
| Before | 1 | pass | 153 s (with build) | 4.10 3.87 4.92 | — |
| Before | 2 | FAIL | 39 s | 20.92 14.55 9.33 | `SourcePageOverviewLayoutTests` ×3, `PVPageColumnsTests.wideHostPlacesTheSecondaryBeside` |
| Before | 3 | pass | 28 s | 14.60 13.73 9.27 | — |
| Before | 4 | pass | 27 s | 11.03 12.97 9.15 | — |
| Before | 5 | pass | 27 s | 14.14 13.47 9.44 | — |
| After | 1 | pass | 28 s | 17.38 17.20 12.06 | — |
| After | 2 | pass | 25 s | 13.09 16.22 11.85 | — |
| After | 3 | pass | 24 s | 10.54 15.36 11.67 | — |
| After | 4 | pass | 25 s | 9.00 14.61 11.50 | — |
| After | 5 | pass | 25 s | 8.53 14.05 11.39 | — |

## Round 1 — review
Reviewers: CON, MNT, UX (`macos/App/**`). Skipped: SEC+PRF — no catalog data, store, list, cache or model code; the diff moves layout arithmetic into static functions that do the same work per layout pass, and deletes tests.
