# Ship log — S13-17 Remove a claim

**Plan:** [`S13-17-remove-a-claim.md`](S13-17-remove-a-claim.md) · **Branch:** `s13-17-remove-a-claim` · **PR:** #729
**State:** review 1

## Round 0 — implement
Head `69e06cdf`. Gate PASS (orchestrator run). Report: `UnpromoteFlowTests` and T-7 seed through `FakeStore` and let the session load; `unpromoteLastRecordNamesTheDeletedHandle` lists the Subject's own handle id; `PersonSourcesList` keeps `unpromotingClaimID` (D-10 did not say where the removed claim's id comes from); the Source page keeps its `model.sync` `.onChange`; the strings landed in the STEP-2/3 commit. The manual check was not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. SEC+PRF runs for `macos/App/Features/**` models and session code.
